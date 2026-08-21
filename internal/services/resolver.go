package services

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/blkhole-sh/blkhole/internal/cache"
	"github.com/miekg/dns"
)

// Resolver contains the core dns resolving logic used in DoH and DoT
type Resolver interface {
	Resolve(*dns.Msg, string) (*dns.Msg, error)
}

// MutableResolver supports changing the upstream DNS server without restarting.
type MutableResolver interface {
	Resolver
	SetUpstreamDNS(string)
	UpstreamDNS() string
}

// resolver implements the Resolver interface
type resolver struct {
	contentBlocker ContentBlocker
	statsCache     cache.StatsCache
	deviceCache    cache.DeviceCache
	upstreamDNS    string
	upstreamMu     sync.RWMutex
	dnsClient      *dns.Client
	queryLog       *QueryLogBuffer
}

// NewResolver creates a new Resolver instance
func NewResolver(contentBlocker ContentBlocker, statsCache cache.StatsCache, deviceCache cache.DeviceCache, upstreamDNS string, queryLog *QueryLogBuffer) MutableResolver {
	return &resolver{
		contentBlocker: contentBlocker,
		statsCache:     statsCache,
		deviceCache:    deviceCache,
		upstreamDNS:    upstreamDNS,
		queryLog:       queryLog,
		dnsClient: &dns.Client{
			Timeout:        1 * time.Second,
			SingleInflight: true,
		},
	}
}

// ValidateUpstreamDNS validates an IP address and port used as an upstream DNS server.
func ValidateUpstreamDNS(addr string) (string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", fmt.Errorf("invalid upstream DNS format (expected host:port): %w", err)
	}
	if ip := net.ParseIP(host); ip == nil {
		return "", fmt.Errorf("invalid IP address: %s", host)
	}
	if _, err := net.LookupPort("tcp", port); err != nil {
		return "", fmt.Errorf("invalid port: %s", port)
	}
	return addr, nil
}

func (r *resolver) SetUpstreamDNS(value string) {
	r.upstreamMu.Lock()
	r.upstreamDNS = value
	r.upstreamMu.Unlock()
}

func (r *resolver) UpstreamDNS() string {
	r.upstreamMu.RLock()
	defer r.upstreamMu.RUnlock()
	return r.upstreamDNS
}

// Resolve proccesses a DNS message and returns the answer message
func (r *resolver) Resolve(msg *dns.Msg, deviceHash string) (*dns.Msg, error) {
	// Create the response message
	response := new(dns.Msg)
	response.SetReply(msg)

	// Refuse queries for unknown devices
	if deviceHash != "" {
		if _, ok := r.deviceCache.GetDeviceID(deviceHash); !ok {
			response.SetRcode(msg, dns.RcodeRefused)
			return response, nil
		}

		// Increment query count for this device
		r.statsCache.Increment(deviceHash)
	}

	// Check if any domain should be blocked
	blocked := false
	for _, question := range msg.Question {
		domain := strings.TrimSuffix(question.Name, ".")
		isBlocked, err := r.contentBlocker.IsBlocked(domain, deviceHash)
		if err != nil {
			// A failed check (e.g. unusual domain format) must not block the query
			log.Printf("content blocker check failed for %q: %v", domain, err)
			continue
		}
		if isBlocked {
			blocked = true
			break
		}
	}

	// If domain is blocked, return NXDOMAIN
	if blocked {
		if deviceHash != "" {
			r.statsCache.IncrementBlocked(deviceHash)
		}
		if r.queryLog != nil && deviceHash != "" && len(msg.Question) > 0 {
			domain := strings.TrimSuffix(msg.Question[0].Name, ".")
			r.queryLog.Enqueue(deviceHash, domain, true)
		}
		response.SetRcode(msg, dns.RcodeNameError)
		response.Answer = nil
		return response, nil
	}

	if r.queryLog != nil && deviceHash != "" && len(msg.Question) > 0 {
		domain := strings.TrimSuffix(msg.Question[0].Name, ".")
		r.queryLog.Enqueue(deviceHash, domain, false)
	}

	// Forward the DNS query to the upstream server
	res, _, err := r.dnsClient.Exchange(msg, r.UpstreamDNS())
	if err != nil {
		// Set SERVFAIL in case of an upstream failure
		response.SetRcode(msg, dns.RcodeServerFailure)
		return response, fmt.Errorf("failed to forward dns query to upstream server: %w", err)
	}

	// Append the answer from the upstream server
	response.Answer = append(response.Answer, res.Answer...)
	return response, nil
}
