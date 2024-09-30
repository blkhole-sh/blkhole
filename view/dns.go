package view

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/miekg/dns"
)

// List of blocked domains
var blockedDomains = []string{
	"reddit.com",
	"google.com",
	"startmunich.de",
}

// Function to check if a domain is in the blocked list
func isBlockedDomain(domain string) bool {
	// Remove trailing dot if it exists
	domain = strings.TrimSuffix(domain, ".")
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

	switch r.Method {
	case "GET":
		fmt.Println("Incoming DNS request via GET")

		// Handle GET request
		queryParams := r.URL.Query()
		dnsMsgB64 := queryParams.Get("dns")
		dnsMsg, err = base64.RawURLEncoding.DecodeString(dnsMsgB64)
		if err != nil {
			http.Error(w, "Failed to decode base64 DNS query", http.StatusBadRequest)
			return
		}

	case "POST":
		fmt.Println("Incoming DNS request via POST")
		// Handle POST request
		dnsMsg, err = io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read DNS query from body", http.StatusBadRequest)
			return
		}

	default:
		http.Error(w, "Only GET and POST methods are supported", http.StatusMethodNotAllowed)
		return
	}

	// Create a DNS message from the received query
	msg := new(dns.Msg)
	err = msg.Unpack(dnsMsg)
	if err != nil {
		http.Error(w, "Failed to unpack DNS query", http.StatusBadRequest)
		return
	}

	// Check if the domain is blocked
	for _, question := range msg.Question {
		domain := question.Name
		// Remove trailing dot from the domain if present
		domain = strings.TrimSuffix(domain, ".")
		fmt.Println("Requested domain:", domain)

		if isBlockedDomain(domain) {
			redirectUrl := os.Getenv("REDIRECT_URL")
			fmt.Printf("Domain %s is blocked. Redirecting to %s\n", domain, redirectUrl)

			// Respond with a DNS CNAME record to redirect to leo.lemon3.studio
			blockedURL := redirectUrl + "." // Redirect to the root page of your server
			response := new(dns.Msg)
			response.SetReply(msg)

			// Create a CNAME record pointing to leo.lemon3.studio
			rr, err := dns.NewRR(fmt.Sprintf("%s CNAME %s", domain, blockedURL))
			if err != nil {
				http.Error(w, "Failed to create CNAME record", http.StatusInternalServerError)
				return
			}

			// Add the CNAME record to the response
			response.Answer = append(response.Answer, rr)

			// Pack the DNS response
			responseBytes, err := response.Pack()
			if err != nil {
				http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
				return
			}

			// Respond with the DNS CNAME response
			w.Header().Set("Content-Type", "application/dns-message")
			w.Write(responseBytes)
			return
		}
	}

	// If the domain is not blocked, forward the DNS query to the upstream DNS server
	client := new(dns.Client)
	response, _, err := client.Exchange(msg, os.Getenv("UPSTREAM_DNS_SERVER"))
	if err != nil {
		http.Error(w, "Failed to forward DNS query", http.StatusInternalServerError)
		return
	}

	// Pack the DNS response
	responseBytes, err := response.Pack()
	if err != nil {
		http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
		return
	}

	// Respond with raw DNS message for both GET and POST requests
	w.Header().Set("Content-Type", "application/dns-message")
	w.Write(responseBytes)
}
