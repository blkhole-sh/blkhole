package servers

import (
	"context"
	"crypto/tls"
	"log"
	"strings"

	"github.com/blkhole-sh/blkhole/internal/services"
	"github.com/miekg/dns"
)

// deviceHashFromSNI extracts the device hash from the TLS server name.
// A connection to <hash>.<domain> identifies the device; the bare domain
// carries no device.
func deviceHashFromSNI(w dns.ResponseWriter, domain string) string {
	cs, ok := w.(dns.ConnectionStater)
	if !ok {
		return ""
	}
	state := cs.ConnectionState()
	if state == nil {
		return ""
	}
	host := state.ServerName
	if host != domain && strings.HasSuffix(host, "."+domain) {
		return strings.TrimSuffix(host, "."+domain)
	}
	return ""
}

func StartDoT(ctx context.Context, resolver services.Resolver, domain string, tlsConfig *tls.Config) error {
	server := &dns.Server{
		Addr:      ":853",
		Net:       "tcp-tls",
		TLSConfig: tlsConfig,
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, msg *dns.Msg) {
			// Get device hash from the connection's TLS SNI
			deviceHash := deviceHashFromSNI(w, domain)

			// Resolve DNS query
			resp, err := resolver.Resolve(msg, deviceHash)
			if err != nil {
				log.Printf("error resolving dns query: %v", err)
			}

			// If no response is available, reply with SERVFAIL
			if resp == nil {
				resp = new(dns.Msg)
				resp.SetRcode(msg, dns.RcodeServerFailure)
			}

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
