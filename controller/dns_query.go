package controller

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/miekg/dns"
)

const localhost = "127.0.0.1"

var blockedDomains = []string{
	"reddit.com",
	"startmunich.de",
	"youtube.com",
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

func DnsQueryController(w http.ResponseWriter, r *http.Request) {
	var dnsMsg []byte
	var err error

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
		log.Println("DoH: Requested domain:", domain)

		// If the domain is blocked, return a fixed IP address
		if isBlockedDomain(domain) {
			log.Printf("DoH: Domain %s is blocked. Returning fixed IP: %s\n", domain, localhost)

			// Create a DNS response with the fixed IP address
			response := new(dns.Msg)
			response.SetReply(msg)

			// Create an A record for the fixed IP address
			rr, err := dns.NewRR(fmt.Sprintf("%s A %s", question.Name, localhost))
			if err != nil {
				http.Error(w, "Failed to create A record", http.StatusInternalServerError)
				return
			}

			// Add the A record to the answer section of the response
			response.Answer = append(response.Answer, rr)

			// Pack and send the DNS response
			responseBytes, err := response.Pack()
			if err != nil {
				http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
				return
			}

			// Respond with the DNS message containing the fixed IP
			w.Header().Set("Content-Type", "application/dns-message")
			w.Write(responseBytes)
			return
		}
	}

	// If the domain is not blocked, forward the DNS query to the upstream server
	client := new(dns.Client)
	response, _, err := client.Exchange(msg, "127.0.0.1:53")
	if err != nil {
		log.Printf("DoH: Failed to forward DNS query to upstream server: %v", err)

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
