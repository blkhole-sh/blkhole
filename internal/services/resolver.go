package services

import (
	"log"
	"strings"
	"time"

	"github.com/lemon3studio/blkhole/internal/cache"
	"github.com/miekg/dns"
)

// Resolver contains the core dns resolving logic used in DoH and DoT
type Resolver interface {
	Resolve(*dns.Msg, string) (*dns.Msg, error)
}

// resolver implements the Resolver interface
type resolver struct {
	contentBlocker ContentBlocker
	statsCache     cache.StatsCache
	upstreamDNS    string
	dnsClient      *dns.Client
	queryLog       *QueryLogBuffer
}

// NewResolver creates a new Resolver instance
func NewResolver(contentBlocker ContentBlocker, statsCache cache.StatsCache, upstreamDNS string, queryLog *QueryLogBuffer) Resolver {
	return &resolver{
		contentBlocker: contentBlocker,
		statsCache:     statsCache,
		upstreamDNS:    upstreamDNS,
		queryLog:       queryLog,
		dnsClient: &dns.Client{
			Timeout:        1 * time.Second,
			SingleInflight: true,
		},
	}
}

// Resolve proccesses a DNS message and returns the answer message
func (r *resolver) Resolve(msg *dns.Msg, deviceHash string) (*dns.Msg, error) {
	// Increment query count for this device
	// TODO: Check if device exists in cache, otherwise return DNS error
	if deviceHash != "" {
		r.statsCache.Increment(deviceHash)
	}

	// Create the response message
	response := new(dns.Msg)
	response.SetReply(msg)

	// Check if any domain should be blocked
	blocked := false
	for _, question := range msg.Question {
		domain := strings.TrimSuffix(question.Name, ".")
		isBlocked, err := r.contentBlocker.IsBlocked(domain, deviceHash)
		if isBlocked || err != nil {
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
	res, _, err := r.dnsClient.Exchange(msg, r.upstreamDNS)
	if err != nil {
		log.Printf("failed to forward dns query to upstream server: %v", err)

		// Set SERVFAIL in case of an upstream failure
		response.SetRcode(msg, dns.RcodeServerFailure)
		return response, nil
	}

	// Append the answer from the upstream server
	response.Answer = append(response.Answer, res.Answer...)
	return response, nil
}
