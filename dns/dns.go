package dns

import (
	"fmt"
	"log"

	"github.com/miekg/dns"
)

type dnsHandler struct{}

// Resolve DNS question using given upstream server
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

// Parse DNS request and resolve each question
func (h *dnsHandler) ServeDNS(w dns.ResponseWriter, r *dns.Msg) {
	msg := new(dns.Msg)
	msg.SetReply(r)
	msg.Authoritative = true

	for _, question := range r.Question {
		log.Printf(":53 received query: %s\n", question.Name)
		answers, err := resolve(question.Name, question.Qtype, "1.1.1.2:53")
		if err != nil {
			fmt.Println(err)
		}

		msg.Answer = append(msg.Answer, answers...)
	}

	w.WriteMsg(msg)
}

// Listen on udp port 53 for DNS request, serve DNS
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
