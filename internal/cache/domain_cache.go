// Package cache provides in-memory caching for DNS blocking lookups
package cache

import (
	"strings"
	"sync"
	"sync/atomic"

	"github.com/armon/go-radix"
	"github.com/blkhole-sh/blkhole/internal/model"
)

// DomainCache provides fast domain-to-rule lookups using a radix tree
type DomainCache interface {
	Load(domains []*model.Domain, rules []*model.Rule)
	LoadDomains(domains []*model.Domain)
	LoadRules(rules []*model.Rule)
	LookupDomainID(domain string) (int, bool)
	GetRules(domainID int) []int
}

type domainSnapshot struct {
	domainTree   *radix.Tree   // Reversed domain → domain ID
	domainToRule map[int][]int // Domain ID → Rule IDs
}

// domainCache implements the DomainCache interface
type domainCache struct {
	writeMu  sync.Mutex
	snapshot atomic.Pointer[domainSnapshot]
}

// NewDomainCache creates a new domain cache instance
func NewDomainCache() DomainCache {
	dc := &domainCache{}
	dc.snapshot.Store(newDomainSnapshot())
	return dc
}

func newDomainSnapshot() *domainSnapshot {
	return &domainSnapshot{
		domainTree:   radix.New(),
		domainToRule: make(map[int][]int),
	}
}

func buildDomainTree(domains []*model.Domain) *radix.Tree {
	tree := radix.New()
	for _, d := range domains {
		reversedDomain := reverseDomain(d.Name)
		tree.Insert(reversedDomain, d.ID)
	}
	return tree
}

func buildDomainRules(rules []*model.Rule) map[int][]int {
	domainToRule := make(map[int][]int)
	for _, r := range rules {
		domainToRule[r.DomainID] = append(domainToRule[r.DomainID], r.ID)
	}
	return domainToRule
}

func (dc *domainCache) current() *domainSnapshot {
	if snapshot := dc.snapshot.Load(); snapshot != nil {
		return snapshot
	}
	return newDomainSnapshot()
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

// Load publishes a complete domain cache snapshot.
func (dc *domainCache) Load(domains []*model.Domain, rules []*model.Rule) {
	dc.writeMu.Lock()
	defer dc.writeMu.Unlock()

	dc.snapshot.Store(&domainSnapshot{
		domainTree:   buildDomainTree(domains),
		domainToRule: buildDomainRules(rules),
	})
}

// LoadDomains populates the radix tree with domain data
func (dc *domainCache) LoadDomains(domains []*model.Domain) {
	dc.writeMu.Lock()
	defer dc.writeMu.Unlock()

	current := dc.current()
	dc.snapshot.Store(&domainSnapshot{
		domainTree:   buildDomainTree(domains),
		domainToRule: current.domainToRule,
	})
}

// LoadRules populates the domain-to-rule mapping
func (dc *domainCache) LoadRules(rules []*model.Rule) {
	dc.writeMu.Lock()
	defer dc.writeMu.Unlock()

	current := dc.current()
	dc.snapshot.Store(&domainSnapshot{
		domainTree:   current.domainTree,
		domainToRule: buildDomainRules(rules),
	})
}

// LookupDomainID finds the most specific domain ID for a given domain
func (dc *domainCache) LookupDomainID(domain string) (int, bool) {
	snapshot := dc.current()

	// Reverse domain for radix tree lookup
	reversedDomain := reverseDomain(domain)
	parts := strings.Split(reversedDomain, ".")

	// Check from full domain to parent domains
	for i := len(parts); i > 0; i-- {
		checkDomain := strings.Join(parts[:i], ".")
		if val, ok := snapshot.domainTree.Get(checkDomain); ok {
			if id, ok := val.(int); ok {
				return id, true
			}
		}
	}

	return 0, false
}

// GetRules returns all rule IDs for a given domain ID
func (dc *domainCache) GetRules(domainID int) []int {
	return append([]int(nil), dc.current().domainToRule[domainID]...)
}
