package services

import (
	"testing"
	"time"

	"github.com/lemon3studio/blkhole/internal/cache"
	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/miekg/dns"
)

// mockContentBlocker implements ContentBlocker for resolver tests
type mockContentBlocker struct {
	isBlockedFunc func(domain, deviceHash string) (bool, error)
}

func (m *mockContentBlocker) Init() error   { return nil }
func (m *mockContentBlocker) Reload() error { return nil }
func (m *mockContentBlocker) IsBlocked(domain, deviceHash string) (bool, error) {
	if m.isBlockedFunc != nil {
		return m.isBlockedFunc(domain, deviceHash)
	}
	return false, nil
}

// mockStatsCache implements cache.StatsCache for resolver tests
type mockStatsCache struct {
	incrementedHashes []string
	blockedHashes     []string
	cache.DeviceCache
}

func (m *mockStatsCache) Increment(deviceHash string)              { m.incrementedHashes = append(m.incrementedHashes, deviceHash) }
func (m *mockStatsCache) IncrementBlocked(deviceHash string)       { m.blockedHashes = append(m.blockedHashes, deviceHash) }
func (m *mockStatsCache) IncrementAt(h string, _ time.Time, _ int) {}
func (m *mockStatsCache) IncrementBlockedAt(h string, _ time.Time, _ int) {}
func (m *mockStatsCache) GetCounts(h, r string) []model.StatCount          { return nil }
func (m *mockStatsCache) GetBlockedCounts(h, r string) []model.StatCount   { return nil }
func (m *mockStatsCache) GetUserCounts(hs []string, r string) []model.StatCount { return nil }
func (m *mockStatsCache) GetUserBlockedCounts(hs []string, r string) []model.StatCount { return nil }
func (m *mockStatsCache) Start()                                            {}
func (m *mockStatsCache) Cleanup()                                          {}

func newTestDNSMsg(domain string) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	return msg
}

func TestResolver_BlockedDomain_ReturnsNXDOMAIN(t *testing.T) {
	blocker := &mockContentBlocker{
		isBlockedFunc: func(domain, _ string) (bool, error) {
			return domain == "blocked.com", nil
		},
	}
	stats := &mockStatsCache{}
	r := NewResolver(blocker, stats, "127.0.0.1:5353", nil)

	msg := newTestDNSMsg("blocked.com")
	resp, err := r.Resolve(msg, "device-hash")

	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if resp.Rcode != dns.RcodeNameError {
		t.Errorf("expected NXDOMAIN (%d), got %d", dns.RcodeNameError, resp.Rcode)
	}
	if len(resp.Answer) != 0 {
		t.Errorf("expected empty Answer for blocked domain, got %d records", len(resp.Answer))
	}
}

func TestResolver_BlockedDomain_IncrementsBlockedStat(t *testing.T) {
	blocker := &mockContentBlocker{
		isBlockedFunc: func(_, _ string) (bool, error) { return true, nil },
	}
	stats := &mockStatsCache{}
	r := NewResolver(blocker, stats, "127.0.0.1:5353", nil)

	r.Resolve(newTestDNSMsg("blocked.com"), "dev-hash")

	if len(stats.blockedHashes) == 0 || stats.blockedHashes[0] != "dev-hash" {
		t.Errorf("IncrementBlocked not called with correct device hash, got %v", stats.blockedHashes)
	}
}

func TestResolver_EmptyDeviceHash_SkipsStatIncrement(t *testing.T) {
	blocker := &mockContentBlocker{}
	stats := &mockStatsCache{}
	r := NewResolver(blocker, stats, "127.0.0.1:5353", nil)

	r.Resolve(newTestDNSMsg("example.com"), "")

	if len(stats.incrementedHashes) != 0 {
		t.Errorf("Increment should not be called for empty device hash, got %v", stats.incrementedHashes)
	}
}

func TestResolver_NonBlockedDomain_IncrementsQueryStat(t *testing.T) {
	blocker := &mockContentBlocker{
		isBlockedFunc: func(_, _ string) (bool, error) { return false, nil },
	}
	stats := &mockStatsCache{}
	r := NewResolver(blocker, stats, "127.0.0.1:5353", nil)

	r.Resolve(newTestDNSMsg("example.com"), "my-device")

	if len(stats.incrementedHashes) == 0 || stats.incrementedHashes[0] != "my-device" {
		t.Errorf("Increment not called with correct device hash, got %v", stats.incrementedHashes)
	}
}

func TestResolver_UpstreamFailure_ReturnsSERVFAIL(t *testing.T) {
	blocker := &mockContentBlocker{
		isBlockedFunc: func(_, _ string) (bool, error) { return false, nil },
	}
	stats := &mockStatsCache{}
	r := NewResolver(blocker, stats, "127.0.0.1:5353", nil) // unreachable upstream

	msg := newTestDNSMsg("example.com")
	resp, err := r.Resolve(msg, "")

	if err == nil {
		t.Error("expected error for upstream failure, got nil")
	}
	if resp.Rcode != dns.RcodeServerFailure {
		t.Errorf("expected SERVFAIL (%d), got %d", dns.RcodeServerFailure, resp.Rcode)
	}
}
