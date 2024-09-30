package view

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/miekg/dns"
)

var blockedDomains = []string{
	"reddit.com",
	"google.com",
	"startmunich.de",
}

// Check if the domain is in the blocked list
func isBlockedDomain(domain string) bool {
	// Normalize the domain to lowercase and remove trailing dot if present
	domain = strings.ToLower(strings.TrimSuffix(domain, "."))
	for _, blocked := range blockedDomains {
		if strings.HasSuffix(domain, blocked) {
			return true
		}
	}
	return false
}

func DnsQueryHandler(w http.ResponseWriter, r *http.Request) {
	var dnsMsg []byte
	var err error

	// Fetch the upstream DNS server address from the environment
	upstreamDnsServer := os.Getenv("UPSTREAM_DNS_SERVER")
	if upstreamDnsServer == "" {
		http.Error(w, "UPSTREAM_DNS_SERVER environment variable not set", http.StatusInternalServerError)
		return
	}

	// Handle DNS queries sent via GET or POST
	switch r.Method {
	case "GET":
		// GET request: decode the Base64 DNS message
		queryParams := r.URL.Query()
		dnsMsgB64 := queryParams.Get("dns")
		dnsMsg, err = base64.RawURLEncoding.DecodeString(dnsMsgB64)
		if err != nil {
			http.Error(w, "Failed to decode base64 DNS query", http.StatusBadRequest)
			return
		}

	case "POST":
		// POST request: read the DNS message directly from the body
		dnsMsg, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read DNS query from body", http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "Only GET and POST methods are supported", http.StatusMethodNotAllowed)
		return
	}

	// Unpack the DNS message to analyze it
	msg := new(dns.Msg)
	err = msg.Unpack(dnsMsg)
	if err != nil {
		http.Error(w, "Failed to unpack DNS query", http.StatusBadRequest)
		return
	}

	// Check if the domain is in the blocked list
	for _, question := range msg.Question {
		domain := strings.TrimSuffix(question.Name, ".")
		fmt.Println("Requested domain:", domain)

		// If the domain is blocked, return NXDomain (Non-Existent Domain)
		if isBlockedDomain(domain) {
			fmt.Printf("Domain %s is blocked. Returning NXDOMAIN\n", domain)

			// Create a DNS response with RcodeNameError (NXDOMAIN)
			response := new(dns.Msg)
			response.SetRcode(msg, dns.RcodeNameError) // NXDOMAIN response

			// Pack and send the DNS response
			responseBytes, err := response.Pack()
			if err != nil {
				http.Error(w, "Failed to pack NXDOMAIN response", http.StatusInternalServerError)
				return
			}

			// Respond with the NXDOMAIN DNS message
			w.Header().Set("Content-Type", "application/dns-message")
			w.Write(responseBytes)
			return
		}
	}

	// If the domain is not blocked, forward the DNS query to the upstream server
	client := new(dns.Client)
	response, _, err := client.Exchange(msg, upstreamDnsServer)
	if err != nil {
		log.Printf("Failed to forward DNS query to upstream server: %v", err)

		// Create a DNS response with RcodeServerFailure (SERVFAIL)
		fallbackResponse := new(dns.Msg)
		fallbackResponse.SetRcode(msg, dns.RcodeServerFailure)

		// Pack and send the SERVFAIL response
		responseBytes, err := fallbackResponse.Pack()
		if err != nil {
			http.Error(w, "Failed to pack SERVFAIL response", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/dns-message")
		w.Write(responseBytes)
		return
	}

	// If the query was successful, return the response from the upstream server
	responseBytes, err := response.Pack()
	if err != nil {
		http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
		return
	}

	// Respond with the DNS response from the upstream server
	w.Header().Set("Content-Type", "application/dns-message")
	w.Write(responseBytes)
}
