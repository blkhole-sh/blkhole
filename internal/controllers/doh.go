package controllers

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"

	"github.com/lemon3studio/blkhole/internal/cache"
	"github.com/lemon3studio/blkhole/internal/middleware"
	"github.com/lemon3studio/blkhole/internal/services"

	"github.com/miekg/dns"
)

// DoHController defines the interface for DNS operations
type DoHController interface {
	DNSQuery(http.ResponseWriter, *http.Request)
}

// doHController implements the DnsController interface
type dohController struct {
	resolver    services.Resolver
	upstreamDNS string
	domain      string
	statsCache  cache.StatsCache
}

// NewDoHController creates a new DnsController instance
func NewDoHController(resolver services.Resolver, upstreamDNS string, domain string, statsCache cache.StatsCache) DoHController {
	return &dohController{
		resolver:    resolver,
		upstreamDNS: upstreamDNS,
		domain:      domain,
		statsCache:  statsCache,
	}
}

// DNSQuery handles GET and POST requests to process DNS queries
func (dc *dohController) DNSQuery(w http.ResponseWriter, r *http.Request) {
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

	// Get device hash from context
	deviceHash, _ := r.Context().Value(middleware.DeviceKey).(string)

	// Unpack the DNS message to analyze it
	msg := new(dns.Msg)
	err = msg.Unpack(dnsMsg)
	if err != nil {
		log.Printf("failed to unpack dns query: %v", err)
		http.Error(w, "Failed to unpack DNS query", http.StatusBadRequest)
		return
	}

	// Create the response message
	resp, err := dc.resolver.Resolve(msg, deviceHash)
	if err != nil {
		log.Printf("failed to resolve dns query: %v", err)
		http.Error(w, "Failed to resolve DNS query", http.StatusInternalServerError)
		return
	}

	// Pack and send the DNS response
	respBytes, err := resp.Pack()
	if err != nil {
		log.Printf("failed to pack dns response: %v", err)
		http.Error(w, "Failed to pack DNS response", http.StatusInternalServerError)
		return
	}

	// Respond with the DNS message
	w.Header().Set("Content-Type", "application/dns-message")
	w.Write(respBytes)
}
