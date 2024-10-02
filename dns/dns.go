package dns

import (
	"fmt"
	"log"
	"os"

	"github.com/miekg/dns"
)

type dnsHandler struct{}

func resolve(domain string, qtype uint16, upstream string) ([]dns.RR, error) {
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(domain), qtype)
	m.RecursionDesired = true

	c := new(dns.Client)
	in, _, err := c.Exchange(m, upstream)
	if err != nil {
		return nil, err
	}

	return in.Answer, nil
}

func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	// Fetch the upstream DNS server address from the environment
	upstreamDnsServer := os.Getenv("UPSTREAM_DNS_SERVER")

	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, question := range r.Question {
		log.Printf(":53 received query: %s\n", question.Name)
		answers, err := resolve(question.Name, question.Qtype, upstreamDnsServer)
		if err != nil {
			fmt.Println(err)
		}

		msg.Answer = append(msg.Answer, answers...)
	}

	w.WriteMsg(msg)
}

func ListenAndServe() {
	handler := new(dnsHandler)
	server := &dns.Server{
		Addr:      ":53",
		Net:       "udp",
		Handler:   handler,
		UDPSize:   65535,
		ReusePort: true,
	}

	log.Println("Starting DNS server on port 53")
	err := server.ListenAndServe()
	if err != nil {
		log.Printf("Failed to start server: %s\n", err.Error())
	}
}
