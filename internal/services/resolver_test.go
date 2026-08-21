package services

import (
	"fmt"
	"testing"
	"time"

	"github.com/blkhole-sh/blkhole/internal/cache"
	"github.com/blkhole-sh/blkhole/internal/model"
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

func (m *mockStatsCache) Increment(deviceHash string) {
	m.incrementedHashes = append(m.incrementedHashes, deviceHash)
}
func (m *mockStatsCache) IncrementBlocked(deviceHash string) {
	m.blockedHashes = append(m.blockedHashes, deviceHash)
}
func (m *mockStatsCache) IncrementAt(h string, _ time.Time, _ int)                     {}
func (m *mockStatsCache) IncrementBlockedAt(h string, _ time.Time, _ int)              {}
func (m *mockStatsCache) GetCounts(h, r string) []model.StatCount                      { return nil }
func (m *mockStatsCache) GetBlockedCounts(h, r string) []model.StatCount               { return nil }
func (m *mockStatsCache) GetUserCounts(hs []string, r string) []model.StatCount        { return nil }
func (m *mockStatsCache) GetUserBlockedCounts(hs []string, r string) []model.StatCount { return nil }
func (m *mockStatsCache) GetUserSecondCounts(hs []string) map[int64]int                { return nil }
func (m *mockStatsCache) GetUserBlockedSecondCounts(hs []string) map[int64]int         { return nil }
func (m *mockStatsCache) Start()                                                       {}
func (m *mockStatsCache) Cleanup()                                                     {}

// newTestDeviceCache returns a device cache populated with the given hashes
func newTestDeviceCache(hashes ...string) cache.DeviceCache {
	dc := cache.NewDeviceCache()
	devices := make([]*model.Device, len(hashes))
	for i, h := range hashes {
		devices[i] = &model.Device{ID: i + 1, Hash: h}
	}
	dc.LoadDevices(devices)
	return dc
}

func newTestDNSMsg(domain string) *dns.Msg {
	msg := new(dns.Msg)
	msg.SetQuestion(dns.Fqdn(domain), dns.TypeA)
	return msg
}

func TestResolverUpdatesUpstreamDNS(t *testing.T) {
	r := NewResolver(&mockContentBlocker{}, &mockStatsCache{}, newTestDeviceCache(), "9.9.9.9:53", nil)
	r.SetUpstreamDNS("1.1.1.1:53")
	if got := r.UpstreamDNS(); got != "1.1.1.1:53" {
		t.Fatalf("UpstreamDNS() = %q, want 1.1.1.1:53", got)
	}
}

func TestValidateUpstreamDNS(t *testing.T) {
	if _, err := ValidateUpstreamDNS("1.1.1.1:53"); err != nil {
		t.Fatalf("valid upstream DNS rejected: %v", err)
	}
	if _, err := ValidateUpstreamDNS("example.com:53"); err == nil {
		t.Fatal("hostname upstream DNS accepted")
	}
}

func TestResolver_BlockedDomain_ReturnsNXDOMAIN(t *testing.T) {
	blocker := &mockContentBlocker{
		isBlockedFunc: func(domain, _ string) (bool, error) {
			return domain == "blocked.com", nil
		},
	}
	stats := &mockStatsCache{}
	r := NewResolver(blocker, stats, newTestDeviceCache("device-hash"), "127.0.0.1:5353", nil)

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
	r := NewResolver(blocker, stats, newTestDeviceCache("dev-hash"), "127.0.0.1:5353", nil)

	r.Resolve(newTestDNSMsg("blocked.com"), "dev-hash")

	if len(stats.blockedHashes) == 0 || stats.blockedHashes[0] != "dev-hash" {
		t.Errorf("IncrementBlocked not called with correct device hash, got %v", stats.blockedHashes)
	}
}

func TestResolver_EmptyDeviceHash_SkipsStatIncrement(t *testing.T) {
	blocker := &mockContentBlocker{}
	stats := &mockStatsCache{}
	r := NewResolver(blocker, stats, newTestDeviceCache(), "127.0.0.1:5353", nil)

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
	r := NewResolver(blocker, stats, newTestDeviceCache("my-device"), "127.0.0.1:5353", nil)

	r.Resolve(newTestDNSMsg("example.com"), "my-device")

	if len(stats.incrementedHashes) == 0 || stats.incrementedHashes[0] != "my-device" {
		t.Errorf("Increment not called with correct device hash, got %v", stats.incrementedHashes)
	}
}

func TestResolver_UnknownDevice_ReturnsREFUSED(t *testing.T) {
	blocker := &mockContentBlocker{}
	stats := &mockStatsCache{}
	r := NewResolver(blocker, stats, newTestDeviceCache(), "127.0.0.1:5353", nil)

	resp, err := r.Resolve(newTestDNSMsg("example.com"), "unknown-device")

	if err != nil {
		t.Fatalf("Resolve() unexpected error: %v", err)
	}
	if resp.Rcode != dns.RcodeRefused {
		t.Errorf("expected REFUSED (%d), got %d", dns.RcodeRefused, resp.Rcode)
	}
	if len(stats.incrementedHashes) != 0 {
		t.Errorf("Increment should not be called for unknown device, got %v", stats.incrementedHashes)
	}
}

func TestResolver_ContentBlockerError_DoesNotBlock(t *testing.T) {
	blocker := &mockContentBlocker{
		isBlockedFunc: func(domain, _ string) (bool, error) {
			return false, fmt.Errorf("invalid domain format: %s", domain)
		},
	}
	stats := &mockStatsCache{}
	r := NewResolver(blocker, stats, newTestDeviceCache("dev"), "127.0.0.1:5353", nil)

	resp, _ := r.Resolve(newTestDNSMsg("_dmarc.example.com"), "dev")

	if resp.Rcode == dns.RcodeNameError {
		t.Errorf("content blocker error must not result in NXDOMAIN")
	}
	if len(stats.blockedHashes) != 0 {
		t.Errorf("IncrementBlocked should not be called on blocker error, got %v", stats.blockedHashes)
	}
}

func TestResolver_UpstreamFailure_ReturnsSERVFAIL(t *testing.T) {
	blocker := &mockContentBlocker{
		isBlockedFunc: func(_, _ string) (bool, error) { return false, nil },
	}
	stats := &mockStatsCache{}
	r := NewResolver(blocker, stats, newTestDeviceCache(), "127.0.0.1:5353", nil) // unreachable upstream

	msg := newTestDNSMsg("example.com")
	resp, err := r.Resolve(msg, "")

	if err == nil {
		t.Error("expected error for upstream failure, got nil")
	}
	if resp.Rcode != dns.RcodeServerFailure {
		t.Errorf("expected SERVFAIL (%d), got %d", dns.RcodeServerFailure, resp.Rcode)
	}
}
