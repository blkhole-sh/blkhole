package controllers

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/lemon3studio/leo/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/miekg/dns"
)

// DNSController defines the interface for DNS operations
type DNSController interface {
	DNSQuery(http.ResponseWriter, *http.Request)
}

// dnsController implements the DnsController interface
type dnsController struct {
	ContentBlocker services.ContentBlocker
}

// NewDNSController creates a new DnsController instance
func NewDNSController(contentBlocker services.ContentBlocker) DNSController {
	return &dnsController{ContentBlocker: contentBlocker}
}

// DNSQuery handles GET and POST requests to process DNS queries
func (dc *dnsController) DNSQuery(w http.ResponseWriter, r *http.Request) {
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
			log.Printf("failed to decode base64 dns query: %v", err)
			http.Error(w, "Failed to decode base64 DNS query", http.StatusBadRequest)
			return
		}

	case "POST":
		// POST request: read the DNS message directly from the body
		dnsMsg, err = io.ReadAll(r.Body)
		if err != nil {
			log.Printf("failed to read dns query from body: %v", err)
			http.Error(w, "Failed to read DNS query from body", http.StatusBadRequest)
			return
		}

	default:
		log.Printf("only get and post methods are supported")
		http.Error(w, "Only GET and POST methods are supported", http.StatusMethodNotAllowed)
		return
	}

	// Get device hash from url params
	deviceHash := chi.URLParam(r, "deviceHash")

	// Unpack the DNS message to analyze it
	msg := new(dns.Msg)
	err = msg.Unpack(dnsMsg)
	if err != nil {
		log.Printf("failed to unpack dns query: %v", err)
		http.Error(w, "Failed to unpack DNS query", http.StatusBadRequest)
		return
	}

	// Create the response message
	response := new(dns.Msg)
	response.SetReply(msg)

	// Process each question in the DNS query
	for _, question := range msg.Question {
		domain := strings.TrimSuffix(question.Name, ".")

		// Check if domain blocked
		blocked, err := dc.ContentBlocker.IsBlocked(domain, deviceHash)

		// If domain is blocked or err return NXDOMAIN
		if blocked || err != nil {

			// Set the NXDOMAIN response code
			response.Rcode = dns.RcodeNameError
			break // Stop processing further questions if one domain is blocked
		}
	}

	// If no domain was blocked, forward the DNS query to the upstream server
	if response.Rcode == dns.RcodeSuccess {
		client := new(dns.Client)
		res, _, err := client.Exchange(msg, "1.1.1.1:53")
		if err != nil {
			log.Printf("failed to forward dns query to upstream server: %v", err)

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
		log.Printf("failed to pack dns response: %v", err)
		http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
		return
	}

	// Respond with the DNS message
	w.Header().Set("Content-Type", "application/dns-message")
	w.Write(responseBytes)
}
