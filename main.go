package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	charmssh "github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	cryptossh "golang.org/x/crypto/ssh"
)

func main() {
	s, err := wish.NewServer(
		wish.WithAddress("0.0.0.0:2222"),
		wish.WithHostKeyPath("term_host_key"),
		wish.WithMiddleware(
			func(next charmssh.Handler) charmssh.Handler {
				return func(s charmssh.Session) {
					pubKey := s.PublicKey()
					if pubKey == nil {
						fmt.Fprintln(s, "No public key, anonymous user")
						_ = s.Exit(0)
						return
					}

					fingerprint := cryptossh.FingerprintSHA256(pubKey)
					log.Printf("key found, fingerprint is %s", fingerprint)
					_ = s.Exit(0)
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
