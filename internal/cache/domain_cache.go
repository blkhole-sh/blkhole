// Package cache provides in-memory caching for DNS blocking lookups
package cache

import (
	"strings"
	"sync"

	"github.com/armon/go-radix"
	"github.com/lemon3studio/blkhole/internal/model"
)

// DomainCache provides fast domain-to-rule lookups using a radix tree
type DomainCache interface {
	LoadDomains(domains []*model.Domain)
	LoadRules(rules []*model.Rule)
	LookupDomainID(domain string) (int, bool)
	GetRules(domainID int) []int
}

// domainCache implements the DomainCache interface
type domainCache struct {
	mu           sync.RWMutex
	domainTree   *radix.Tree   // Reversed domain → domain ID
	domainToRule map[int][]int // Domain ID → Rule IDs
}

// NewDomainCache creates a new domain cache instance
func NewDomainCache() DomainCache {
	return &domainCache{
		domainTree:   radix.New(),
		domainToRule: make(map[int][]int),
	}
}

// reverseDomain converts "example.com" to "com.example" for radix tree storage
func reverseDomain(domain string) string {
	// Split domain by dots
	parts := strings.Split(domain, ".")

	// Reverse array in-place
	for i := 0; i < len(parts)/2; i++ {
		parts[i], parts[len(parts)-1-i] = parts[len(parts)-1-i], parts[i]
	}

	// Join reversed parts
	return strings.Join(parts, ".")
}

// LoadDomains populates the radix tree with domain data
func (dc *domainCache) LoadDomains(domains []*model.Domain) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	for _, d := range domains {
		reversedDomain := reverseDomain(d.Name)
		dc.domainTree.Insert(reversedDomain, d.ID)
	}
}

// LoadRules populates the domain-to-rule mapping
func (dc *domainCache) LoadRules(rules []*model.Rule) {
	dc.mu.Lock()
	defer dc.mu.Unlock()

	bildCount := 0
	for _, r := range rules {
		dc.domainToRule[r.DomainID] = append(dc.domainToRule[r.DomainID], r.ID)
		// Debug bild.de related rules
		if r.DomainID == 279661 || r.DomainID == 279660 || (r.DomainID >= 12270 && r.DomainID <= 12280) {
			bildCount++
		}
	}
}

// LookupDomainID finds the most specific domain ID for a given domain
func (dc *domainCache) LookupDomainID(domain string) (int, bool) {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	// Reverse domain for radix tree lookup
	reversedDomain := reverseDomain(domain)
	parts := strings.Split(reversedDomain, ".")

	// Check from full domain to parent domains
	for i := len(parts); i > 0; i-- {
		checkDomain := strings.Join(parts[:i], ".")
		if val, ok := dc.domainTree.Get(checkDomain); ok {
			if id, ok := val.(int); ok {
				return id, true
			}
		}
	}

	return 0, false
}

// GetRules returns all rule IDs for a given domain ID
func (dc *domainCache) GetRules(domainID int) []int {
	dc.mu.RLock()
	defer dc.mu.RUnlock()

	return dc.domainToRule[domainID]
}
