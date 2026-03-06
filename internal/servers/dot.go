package servers

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/lemon3studio/blkhole/internal/middleware"
	"github.com/lemon3studio/blkhole/internal/services"
	"github.com/miekg/dns"
)

func StartDoT(ctx context.Context, resolver services.Resolver, tlsConfig *tls.Config) error {
	server := &dns.Server{
		Addr:      ":853",
		Net:       "tcp-tls",
		TLSConfig: tlsConfig,
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, msg *dns.Msg) {
			// Get device hash from context
			deviceHash, _ := ctx.Value(middleware.DeviceKey).(string)

			// Resolve DNS query
			resp, err := resolver.Resolve(msg, deviceHash)
			if err != nil {
				fmt.Errorf("error resolving dns query: %v", err)

				// If failed to resolve, return original message
				w.WriteMsg(msg)
				return
			}

			// If successfully resolved, return response
			w.WriteMsg(resp)
		}),
	}

	go func() {
		<-ctx.Done()
		log.Println("Shutting down dot server...")
		server.Shutdown()
	}()

	log.Println("starting dot server on :853")
	return server.ListenAndServe()
}
