package servers

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"time"
)

func StartHTTP(ctx context.Context, handler http.Handler, domain string, port string, tlsConfig *tls.Config) error {
	// Determine if we should use TLS
	// Use TLS only when:
	// 1. TLS config is provided AND
	// 2. Port is empty (will default to 443) or explicitly 443
	// This allows fly.io to use port 8080 without TLS (reverse proxy handles it)
	useTLS := false
	if tlsConfig != nil && (port == "" || port == "443") {
		useTLS = true
		if port == "" {
			port = "443"
		}
	} else if port == "" {
		port = "80" // Default to HTTP
	}

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

	if useTLS {
		server.TLSConfig = tlsConfig
		log.Printf("starting https server on :%s (domain: %s)\n", port, domain)
		return server.ListenAndServeTLS("", "")
	}

	// Plain HTTP (local mode or behind a reverse proxy like fly.io)
	log.Printf("starting http server on :%s\n", port)
	return server.ListenAndServe()
}
