package servers

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/acme/autocert"
)

func StartHTTP(ctx context.Context, handler http.Handler, domain string, port string, tlsConfig *tls.Config) error {
	server := &http.Server{
		Addr:    ":" + port,
		Handler: handler,
	}

	// Graceful shutdown
	go func() {
		<-ctx.Done()
		log.Println("shutting down http server...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(shutdownCtx)
	}()

	// Plain HTTP (local or behind a reverse proxy like Fly.io)
	if port != "" && tlsConfig == nil {
		log.Printf("starting http server on :%s\n", port)
		return server.ListenAndServe()
	}

	// TLS via provided TLSConfig (Caddy or oldschool)
	if tlsConfig != nil && domain != "localhost" {
		server.TLSConfig = tlsConfig
		log.Printf("starting https server on %s:%s\n", domain, port)
		return server.ListenAndServeTLS("", "")
	}

	// TLS via Autocert
	if tlsConfig == nil && domain != "local" {
		configDir, _ := os.UserConfigDir()
		m := &autocert.Manager{
			Cache:      autocert.DirCache(filepath.Join(configDir, "blkhole", "certs")),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(domain),
		}

		// ACME HTTP Challenge on :80
		go func() {
			log.Println("starting acme http server on :80")
			if err := http.ListenAndServe(":80", m.HTTPHandler(nil)); err != nil {
				fmt.Errorf("acme server error: %v", err)
			}
		}()

		server.TLSConfig = m.TLSConfig()
		log.Printf("starting https server on :%s using Autocert for domain %s\n", port, domain)
		return server.ListenAndServeTLS("", "")
	}

	return fmt.Errorf("invalid server configuration\n")
}
