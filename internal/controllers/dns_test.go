package controllers

import (
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lemon3studio/blkhole/internal/model"
	"github.com/miekg/dns"
)

type MockContentBlocker struct{}

func (m *MockContentBlocker) Init() error { return nil }
func (m *MockContentBlocker) IsBlocked(domain string, deviceHash string) (bool, error) {
	return false, nil
}

type MockStatsCache struct{}

func (m *MockStatsCache) Increment(deviceHash string) {}
func (m *MockStatsCache) IncrementBlocked(deviceHash string) {}
func (m *MockStatsCache) IncrementAt(deviceHash string, timestamp time.Time, count int) {}
func (m *MockStatsCache) IncrementBlockedAt(deviceHash string, timestamp time.Time, count int) {}
func (m *MockStatsCache) GetCounts(deviceHash string, timeRange string) []model.StatCount { return nil }
func (m *MockStatsCache) GetBlockedCounts(deviceHash string, timeRange string) []model.StatCount { return nil }
func (m *MockStatsCache) GetUserCounts(deviceHashes []string, timeRange string) []model.StatCount { return nil }
func (m *MockStatsCache) GetUserBlockedCounts(deviceHashes []string, timeRange string) []model.StatCount { return nil }
func (m *MockStatsCache) Start() {}
func (m *MockStatsCache) Cleanup() {}


func TestDNSQuery(t *testing.T) {
	// Start a mock DNS server
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen packet: %v", err)
	}
	defer pc.Close()

	server := &dns.Server{
		PacketConn: pc,
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
			m := new(dns.Msg)
			m.SetReply(r)
			w.WriteMsg(m)
		}),
	}
	go server.ActivateAndServe()

	// Wait a bit for server to start
	time.Sleep(100 * time.Millisecond)

	upstreamDNS := pc.LocalAddr().String()

	dc := NewDNSController(&MockContentBlocker{}, upstreamDNS, "example.com", &MockStatsCache{})

	// Create a DNS query
	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
	msg, _ := m.Pack()
	b64Msg := base64.RawURLEncoding.EncodeToString(msg)

	req := httptest.NewRequest("GET", "/dns?dns="+b64Msg, nil)
	w := httptest.NewRecorder()

	dc.DNSQuery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status OK, got %v", w.Code)
	}

	// Parse response body
	respMsg := new(dns.Msg)
	if err := respMsg.Unpack(w.Body.Bytes()); err != nil {
		t.Errorf("failed to unpack response: %v", err)
	}

	if respMsg.Rcode != dns.RcodeSuccess {
		t.Errorf("expected Rcode Success, got %v", respMsg.Rcode)
	}
}

func BenchmarkDNSQuery(b *testing.B) {
	// Start a mock DNS server
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		b.Fatalf("failed to listen packet: %v", err)
	}
	defer pc.Close()

	server := &dns.Server{
		PacketConn: pc,
		Handler: dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
			m := new(dns.Msg)
			m.SetReply(r)
			// Simulate delay
			time.Sleep(1 * time.Millisecond) // Minimal delay to simulate network RTT
			w.WriteMsg(m)
		}),
	}
	go server.ActivateAndServe()

    // Wait a bit for server to start
    time.Sleep(100 * time.Millisecond)

    upstreamDNS := pc.LocalAddr().String()

	dc := NewDNSController(&MockContentBlocker{}, upstreamDNS, "example.com", &MockStatsCache{})

    // Create a DNS query
    m := new(dns.Msg)
    m.SetQuestion(dns.Fqdn("example.com"), dns.TypeA)
    msg, _ := m.Pack()
    b64Msg := base64.RawURLEncoding.EncodeToString(msg)

    // Pre-create request to avoid overhead in loop, assuming handler doesn't modify it destructively
    req := httptest.NewRequest("GET", "/dns?dns="+b64Msg, nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			dc.DNSQuery(w, req)
		}
	})
}
