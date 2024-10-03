package controllers

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"server/services"
	"strings"

	"github.com/miekg/dns"
)

type DnsController struct {
	ContentBlocker *services.ContentBlocker
}

func (dc *DnsController) DnsQuery(w http.ResponseWriter, r *http.Request) {
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

	// Create the response message
	response := new(dns.Msg)
	response.SetReply(msg)

	// Process each question in the DNS query
	for _, question := range msg.Question {
		domain := strings.TrimSuffix(question.Name, ".")
		fmt.Println("Requested domain:", domain)

		// Check if the domain is blocked
		if blocked, _ := dc.ContentBlocker.IsBlocked(domain); blocked {
			fmt.Printf("Domain %s is blocked. Returning NXDOMAIN for this domain.\n", domain)

			// Set the NXDOMAIN response code
			response.Rcode = dns.RcodeNameError
			break // Stop processing further questions if one domain is blocked
		}
	}

	// If no domain was blocked, forward the DNS query to the upstream server
	if response.Rcode == dns.RcodeSuccess {
		client := new(dns.Client)
		res, _, err := client.Exchange(msg, "127.0.0.1:53")
		if err != nil {
			log.Printf("Failed to forward DNS query to upstream server: %v", err)

			// Set SERVFAIL in case of an upstream failure
			response.SetRcode(msg, dns.RcodeServerFailure)
		} else {
			// Append the answer from the upstream server
			response.Answer = append(response.Answer, res.Answer...)
		}
	}

	// Pack and send the DNS response
	responseBytes, err := response.Pack()
	if err != nil {
		http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
		return
	}

	// Respond with the DNS message
	w.Header().Set("Content-Type", "application/dns-message")
	w.Write(responseBytes)
}
