package cache

import (
	"testing"

	"github.com/lemon3studio/blkhole/internal/model"
)

func TestDomainCache_LookupDomainID_Hit(t *testing.T) {
	dc := NewDomainCache()
	dc.LoadDomains([]*model.Domain{
		{ID: 1, Name: "example.com"},
		{ID: 2, Name: "ads.tracker.net"},
	})

	tests := []struct {
		domain string
		wantID int
		wantOK bool
	}{
		{"example.com", 1, true},
		{"ads.tracker.net", 2, true},
		{"unknown.org", 0, false},
	}

	for _, tc := range tests {
		t.Run(tc.domain, func(t *testing.T) {
			id, ok := dc.LookupDomainID(tc.domain)
			if ok != tc.wantOK {
				t.Errorf("LookupDomainID(%q) ok = %v; want %v", tc.domain, ok, tc.wantOK)
			}
			if id != tc.wantID {
				t.Errorf("LookupDomainID(%q) id = %d; want %d", tc.domain, id, tc.wantID)
			}
		})
	}
}

func TestDomainCache_LookupDomainID_Miss(t *testing.T) {
	dc := NewDomainCache()
	id, ok := dc.LookupDomainID("missing.com")
	if ok {
		t.Errorf("expected cache miss for empty cache, got ok=true with id=%d", id)
	}
}

func TestDomainCache_LookupDomainID_SubdomainMatchesParent(t *testing.T) {
	dc := NewDomainCache()
	dc.LoadDomains([]*model.Domain{{ID: 10, Name: "evil.com"}})

	id, ok := dc.LookupDomainID("sub.evil.com")
	if !ok {
		t.Errorf("LookupDomainID(sub.evil.com) expected match via parent domain evil.com")
	}
	if id != 10 {
		t.Errorf("LookupDomainID(sub.evil.com) id = %d; want 10", id)
	}
}

func TestDomainCache_LookupDomainID_ExactMatchPreferred(t *testing.T) {
	dc := NewDomainCache()
	dc.LoadDomains([]*model.Domain{
		{ID: 1, Name: "example.com"},
		{ID: 2, Name: "sub.example.com"},
	})

	id, ok := dc.LookupDomainID("sub.example.com")
	if !ok {
		t.Errorf("LookupDomainID(sub.example.com) expected ok=true")
	}
	if id != 2 {
		t.Errorf("LookupDomainID(sub.example.com) id = %d; want 2 (exact match)", id)
	}
}

func TestDomainCache_GetRules(t *testing.T) {
	dc := NewDomainCache()
	dc.LoadRules([]*model.Rule{
		{ID: 100, DomainID: 1},
		{ID: 101, DomainID: 1},
		{ID: 200, DomainID: 2},
	})

	r1 := dc.GetRules(1)
	if len(r1) != 2 {
		t.Fatalf("GetRules(1) returned %d entries; want 2", len(r1))
	}

	r2 := dc.GetRules(2)
	if len(r2) != 1 || r2[0] != 200 {
		t.Errorf("GetRules(2) = %v; want [200]", r2)
	}

	r3 := dc.GetRules(999)
	if len(r3) != 0 {
		t.Errorf("GetRules(999) = %v; want empty", r3)
	}
}

func TestDomainCache_GetRules_EmptyCache(t *testing.T) {
	dc := NewDomainCache()
	if rules := dc.GetRules(1); len(rules) != 0 {
		t.Errorf("GetRules on empty cache = %v; want empty", rules)
	}
}

func TestReverseDomain(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"example.com", "com.example"},
		{"sub.example.com", "com.example.sub"},
		{"a.b.c.d", "d.c.b.a"},
		{"single", "single"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := reverseDomain(tc.input)
			if got != tc.want {
				t.Errorf("reverseDomain(%q) = %q; want %q", tc.input, got, tc.want)
			}
		})
	}
}
