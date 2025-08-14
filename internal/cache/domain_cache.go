// Package cache provides in-memory caching for DNS blocking lookups
package cache

import (
	"strings"

	"github.com/armon/go-radix"
	"github.com/lemon3studio/leo/internal/model"
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
	for _, d := range domains {
		reversedDomain := reverseDomain(d.Name)
		dc.domainTree.Insert(reversedDomain, d.ID)
	}
}

// LoadRules populates the domain-to-rule mapping
func (dc *domainCache) LoadRules(rules []*model.Rule) {
	bildCount := 0
	for _, r := range rules {
		dc.domainToRule[r.DomainID] = append(dc.domainToRule[r.DomainID], r.ID)
		// Debug bild.de related rules
		if r.DomainID == 279661 || (r.DomainID >= 12270 && r.DomainID <= 12280) {
			fmt.Printf("DEBUG: Rule %d for domain ID %d\n", r.ID, r.DomainID)
			bildCount++
		}
	}
	fmt.Printf("DEBUG: Loaded rules for %d bild-related domains\n", bildCount)
}

// LookupDomainID finds the most specific domain ID for a given domain using longest-prefix match
func (dc *domainCache) LookupDomainID(domain string) (int, bool) {
	// Convert domain to reversed format for radix tree lookup (e.g., "api.example.com" → "com.example.api")
	reversedDomain := reverseDomain(domain)

	var domainID int
	var matchedKey string
	walkCalled := false

	fmt.Printf("DEBUG: Starting domain lookup for %s -> reversed %s\n", domain, reversedDomain)

	// Walk the radix tree to find the longest matching prefix (enables hierarchical domain blocking)
	dc.domainTree.WalkPath(reversedDomain, func(key string, value any) bool {
		walkCalled = true
		fmt.Printf("DEBUG: WalkPath callback - key: %s, value: %v\n", key, value)
		// Extract domain ID from the matched node
		if id, ok := value.(int); ok {
			domainID = id
			matchedKey = key
			fmt.Printf("DEBUG: Found match - key: %s, domain ID: %d\n", key, id)
		}
		// Return false to get the longest match (most specific domain)
		return false
	})

	if !walkCalled {
		fmt.Printf("DEBUG: WalkPath was never called - no matches in tree\n")
	}

	fmt.Printf("DEBUG: Final result - domain ID: %d, found: %v\n", domainID, domainID != 0)

	// Return the domain ID and whether a match was found
	return domainID, domainID != 0
}

// GetRules returns all rule IDs for a given domain ID
func (dc *domainCache) GetRules(domainID int) []int {
	return dc.domainToRule[domainID]
}
