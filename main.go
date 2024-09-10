package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"

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

const upstreamDNS = "8.8.8.8:53" // Upstream DNS server (Google's public DNS)

// Function to handle DNS queries
func dnsQueryHandler(w http.ResponseWriter, r *http.Request) {
	var dnsMsg []byte
	var err error

	switch r.Method {
	case "GET":
		fmt.Print("Incoming DNS request via GET")

		// Handle GET request
		queryParams := r.URL.Query()
		dnsMsgB64 := queryParams.Get("dns")
		dnsMsg, err = base64.RawURLEncoding.DecodeString(dnsMsgB64)
		if err != nil {
			http.Error(w, "Failed to decode base64 DNS query", http.StatusBadRequest)
			return
		}

	case "POST":
		fmt.Print("Incoming DNS request via POST")
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

	// Forward the DNS query to the upstream DNS server
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

	// Encode response for GET request
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/dns-message")
		encodedResponse := base64.RawURLEncoding.EncodeToString(responseBytes)
		fmt.Fprint(w, encodedResponse)
	} else {
		// Respond with raw DNS message for POST request
		w.Header().Set("Content-Type", "application/dns-message")
		w.Write(responseBytes)
	}
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

	// Catch-all route to match any path and show the blocked message
	r.NotFound(blockedPageHandler)
	r.MethodNotAllowed(blockedPageHandler)

	// Handle all routes by redirecting them to the blocked handler
	r.Route("/", func(r chi.Router) {
		// Route specifically for /dns-query, handling both GET and POST requests
		r.Get("/dns-query", dnsQueryHandler)
		r.Post("/dns-query", dnsQueryHandler)

		// Catch-all route for any other GET request
		r.Get("/*", blockedPageHandler)
	})

	// Start the server on port :8080
	log.Println("Starting Leo on :8080")
	err := http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal(err)
	}
}
