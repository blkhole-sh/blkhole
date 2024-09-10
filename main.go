package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/miekg/dns"
)

// Template for the blocked message
var tmpl = template.Must(template.New("blocked").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Blocked</title>
</head>
<body>
    <h1>This page has been blocked by Leo</h1>
</body>
</html>
`))

const (
	upstreamDNS = "8.8.8.8:53"   // Upstream DNS server (Google's public DNS)
	hardcodedIP = "172.217.1.14" // Hardcoded IP address for blocked domains
)

// List of blocked domains
var blockedDomains = []string{
	"reddit.com.",
	"google.com.",
	"startmunich.de.",
}

// Function to check if a domain is in the blocked list
func isBlockedDomain(domain string) bool {
	fmt.Printf("Checking if %s is blocked\n", domain)
	for _, blocked := range blockedDomains {
		fmt.Printf("%s has suffix %s? -> %t\n", domain, blocked, strings.HasSuffix(domain, blocked))
		if strings.HasSuffix(domain, blocked) {
			return true
		}
	}
	return false
}

// Function to handle DNS queries
func dnsQueryHandler(w http.ResponseWriter, r *http.Request) {
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
		fmt.Println("Requested domain:", domain)

		if isBlockedDomain(domain) {
			fmt.Printf("Domain %s is blocked. Responding with hardcoded IP: %s\n", domain, hardcodedIP)

			// Create a DNS response with the hardcoded IP
			response := new(dns.Msg)
			response.SetReply(msg)

			// Create an A record with the hardcoded IP address
			rr, err := dns.NewRR(fmt.Sprintf("%s A %s", domain, hardcodedIP))
			if err != nil {
				http.Error(w, "Failed to create A record", http.StatusInternalServerError)
				return
			}

			// Add the A record to the response
			response.Answer = append(response.Answer, rr)

			// Pack the DNS response
			responseBytes, err := response.Pack()
			if err != nil {
				http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
				return
			}

			// Respond with the DNS response
			w.Header().Set("Content-Type", "application/dns-message")
			w.Write(responseBytes)
			return
		}
	}

	// If the domain is not blocked, forward the DNS query to the upstream DNS server
	client := new(dns.Client)
	response, _, err := client.Exchange(msg, upstreamDNS)
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

// Handler that serves the blocked message
func blockedPageHandler(w http.ResponseWriter, r *http.Request) {
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Unable to render template", http.StatusInternalServerError)
	}
}

func main() {
	// Create a new router using chi
	r := chi.NewRouter()

	// Add some middleware for better logging and recovery
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Route specifically for /dns-query, handling both GET and POST requests
	r.Route("/dns-query", func(r chi.Router) {
		r.Get("/", dnsQueryHandler)
		r.Post("/", dnsQueryHandler)
	})

	// Catch-all route to match any path and show the blocked message
	r.NotFound(blockedPageHandler)
	r.MethodNotAllowed(blockedPageHandler)

	// Start the server on port :8080
	log.Println("Starting Leo on :8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
