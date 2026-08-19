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
	LookupDomainIDs(domain string) []int
	GetRules(domainID int) []int
	RuleDomains(ruleIDs []int) []string
}

type domainSnapshot struct {
	domainTree   *radix.Tree    // Reversed domain → domain ID
	domainToRule map[int][]int  // Domain ID → Rule IDs
	domainNames  map[int]string // Domain ID → domain
	ruleDomains  map[int]string // Rule ID → domain
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
		domainNames:  make(map[int]string),
		ruleDomains:  make(map[int]string),
	}
}

func buildDomainNames(domains []*model.Domain) map[int]string {
	domainNames := make(map[int]string, len(domains))
	for _, domain := range domains {
		domainNames[domain.ID] = domain.Name
	}
	return domainNames
}

func buildRuleDomains(domainNames map[int]string, rules []*model.Rule) map[int]string {
	result := make(map[int]string)
	for _, rule := range rules {
		result[rule.ID] = domainNames[rule.DomainID]
	}
	return result
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

	domainNames := buildDomainNames(domains)
	dc.snapshot.Store(&domainSnapshot{
		domainTree:   buildDomainTree(domains),
		domainToRule: buildDomainRules(rules),
		domainNames:  domainNames,
		ruleDomains:  buildRuleDomains(domainNames, rules),
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
		domainNames:  buildDomainNames(domains),
		ruleDomains:  current.ruleDomains,
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
		domainNames:  current.domainNames,
		ruleDomains:  buildRuleDomains(current.domainNames, rules),
	})
}

// LookupDomainID finds the most specific domain ID for a given domain
func (dc *domainCache) LookupDomainID(domain string) (int, bool) {
	domainIDs := dc.LookupDomainIDs(domain)
	if len(domainIDs) == 0 {
		return 0, false
	}
	return domainIDs[0], true
}

// LookupDomainIDs finds known domain IDs from most specific to parent.
func (dc *domainCache) LookupDomainIDs(domain string) []int {
	snapshot := dc.current()

	// Reverse domain for radix tree lookup
	reversedDomain := reverseDomain(domain)
	parts := strings.Split(reversedDomain, ".")

	// Check from full domain to parent domains
	var domainIDs []int
	for i := len(parts); i > 0; i-- {
		checkDomain := strings.Join(parts[:i], ".")
		if val, ok := snapshot.domainTree.Get(checkDomain); ok {
			if id, ok := val.(int); ok {
				domainIDs = append(domainIDs, id)
			}
		}
	}
	return domainIDs
}

// GetRules returns all rule IDs for a given domain ID
func (dc *domainCache) GetRules(domainID int) []int {
	return append([]int(nil), dc.current().domainToRule[domainID]...)
}

// RuleDomains returns domains for rules, deduplicated by domain.
func (dc *domainCache) RuleDomains(ruleIDs []int) []string {
	snapshot := dc.current()
	seen := make(map[string]struct{})
	for _, ruleID := range ruleIDs {
		if domain := snapshot.ruleDomains[ruleID]; domain != "" {
			seen[domain] = struct{}{}
		}
	}

	result := make([]string, 0, len(seen))
	for domain := range seen {
		result = append(result, domain)
	}
	return result
}
