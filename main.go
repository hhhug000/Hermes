package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	charmssh "github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	wishbubble "github.com/charmbracelet/wish/bubbletea"
	cryptossh "golang.org/x/crypto/ssh"
)

func main() {
	if err := InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	s, err := wish.NewServer(
		wish.WithAddress("0.0.0.0:2222"),
		wish.WithHostKeyPath("term_host_key"),
		wish.WithPublicKeyAuth(func(ctx charmssh.Context, key charmssh.PublicKey) bool {
			return true
		}),
		wish.WithMiddleware(
			wishbubble.Middleware(func(s charmssh.Session) (tea.Model, []tea.ProgramOption) {
				pubKey := s.PublicKey()
				var fingerprint string
				var username string

				if pubKey != nil {
					fingerprint = cryptossh.FingerprintSHA256(pubKey)
					var err error
					username, err = GetUsername(fingerprint)
					if err != nil {
						log.Printf("DB error looking up fingerprint: %v", err)
					}
				}

				m := initialModel(fingerprint, username)
				return m, []tea.ProgramOption{tea.WithAltScreen()}
			}),
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
