package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	charmssh "github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	wishbubble "github.com/charmbracelet/wish/bubbletea"
	"github.com/google/uuid"
	cryptossh "golang.org/x/crypto/ssh"
)

func main() {
	if err := InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	s, err := wish.NewServer(
		// all ips on port 2222
		wish.WithAddress("0.0.0.0:2222"),
		// use the same host key otherwise safety warnings are shown every restart
		wish.WithHostKeyPath("term_host_key"),
		// make sure to allow public key auth otherwise fingerprint never shows
		wish.WithPublicKeyAuth(func(ctx charmssh.Context, key charmssh.PublicKey) bool {
			return true
		}),
		// middleware for commands and tui
		wish.WithMiddleware(
			func(next charmssh.Handler) charmssh.Handler {
				return func(s charmssh.Session) {
					pubKey := s.PublicKey()
					if pubKey == nil {
						fmt.Fprintln(s, "Error: Public key required.")
						_ = s.Exit(1)
						return
					}

					fingerprint := cryptossh.FingerprintSHA256(pubKey)
					username, err := GetUsername(fingerprint)
					if err != nil || username == "" {
						fmt.Fprintln(s, "Error: You must register a username first. Run 'ssh localhost -p 2222' interactively.")
						_ = s.Exit(1)
						return
					}

					cmd := s.Command()

					// pipe in file uploads if command is send
					if len(cmd) >= 2 && cmd[0] == "send" {
						filename := cmd[1]
						fileID := uuid.New().String()
						storagePath := fmt.Sprintf("./storage/%s", fileID)

						dstFile, err := os.Create(storagePath)
						if err != nil {
							fmt.Fprintf(s, "File creation error  %v\r\n", err)
							_ = s.Exit(1)
							return
						}
						defer dstFile.Close()

						// stream bits to disk so it doesnt need to go to ram
						written, err := io.Copy(dstFile, s)
						if err != nil {
							fmt.Fprintf(s, "File streaming error %v\r\n", err)
							_ = s.Exit(1)
							return
						}

						// Save metadata to db
						if err := SaveFileRecord(fileID, filename, fingerprint, storagePath); err != nil {
							fmt.Fprintf(s, "Error saving file record: %v\r\n", err)
							_ = s.Exit(1)
							return
						}

						// grant access to uploader
						_ = GrantAccess(fileID, username)

						fmt.Fprintf(s, "Success! Received %d bytes for '%s'.\r\n", written, filename)
						_ = s.Exit(0)
						return
					}

					// launch tui if no command
					tuiHandler := wishbubble.Middleware(func(s charmssh.Session) (tea.Model, []tea.ProgramOption) {
						m := initialModel(fingerprint, username)
						return m, []tea.ProgramOption{tea.WithAltScreen()}
					})

					tuiHandler(next)(s)
				}
			},
		),
	)

	if err != nil {
		log.Fatalf("could not start server: %s", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Println("Hermes listening on port 2222")
		if err := s.ListenAndServe(); err != nil && err != charmssh.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Hermes shutting down...")
	_ = s.Close()
}
