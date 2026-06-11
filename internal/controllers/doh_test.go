package controllers

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blkhole-sh/blkhole/internal/cache"
	"github.com/blkhole-sh/blkhole/internal/middleware"
	"github.com/miekg/dns"
)

// MockResolver is a mock implementation of services.Resolver
type MockResolver struct {
	ResolveFunc func(*dns.Msg, string) (*dns.Msg, error)
}

func (m *MockResolver) Resolve(msg *dns.Msg, deviceHash string) (*dns.Msg, error) {
	if m.ResolveFunc != nil {
		return m.ResolveFunc(msg, deviceHash)
	}
	return nil, errors.New("ResolveFunc not implemented")
}

// MockStatsCache is a mock implementation of cache.StatsCache
type MockStatsCache struct {
	cache.StatsCache     // Embed the interface to satisfy the compiler
	IncrementFunc        func(deviceHash string)
	IncrementBlockedFunc func(deviceHash string)
}

func (m *MockStatsCache) Increment(deviceHash string) {
	if m.IncrementFunc != nil {
		m.IncrementFunc(deviceHash)
	}
}

func (m *MockStatsCache) IncrementBlocked(deviceHash string) {
	if m.IncrementBlockedFunc != nil {
		m.IncrementBlockedFunc(deviceHash)
	}
}

func createValidDNSMsg() []byte {
	m := new(dns.Msg)
	m.SetQuestion("example.com.", dns.TypeA)
	b, _ := m.Pack()
	return b
}

func TestDOHController_DNSQuery_GET(t *testing.T) {
	mockResolver := &MockResolver{
		ResolveFunc: func(msg *dns.Msg, deviceHash string) (*dns.Msg, error) {
			resp := new(dns.Msg)
			resp.SetReply(msg)
			resp.Answer = append(resp.Answer, &dns.A{
				Hdr: dns.RR_Header{Name: msg.Question[0].Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 300},
			})
			return resp, nil
		},
	}
	mockStatsCache := &MockStatsCache{}

	controller := NewDoHController(mockResolver, "8.8.8.8:53", "example.com", mockStatsCache)

	dnsMsg := createValidDNSMsg()
	encodedMsg := base64.RawURLEncoding.EncodeToString(dnsMsg)

	req := httptest.NewRequest(http.MethodGet, "/dns-query?dns="+encodedMsg, nil)
	w := httptest.NewRecorder()

	controller.DNSQuery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Header().Get("Content-Type") != "application/dns-message" {
		t.Errorf("Expected Content-Type application/dns-message, got %s", w.Header().Get("Content-Type"))
	}
}

func TestDOHController_DNSQuery_POST(t *testing.T) {
	mockResolver := &MockResolver{
		ResolveFunc: func(msg *dns.Msg, deviceHash string) (*dns.Msg, error) {
			resp := new(dns.Msg)
			resp.SetReply(msg)
			return resp, nil
		},
	}
	mockStatsCache := &MockStatsCache{}

	controller := NewDoHController(mockResolver, "8.8.8.8:53", "example.com", mockStatsCache)

	dnsMsg := createValidDNSMsg()

	req := httptest.NewRequest(http.MethodPost, "/dns-query", bytes.NewReader(dnsMsg))
	w := httptest.NewRecorder()

	controller.DNSQuery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestDOHController_DNSQuery_UnsupportedMethod(t *testing.T) {
	mockResolver := &MockResolver{}
	mockStatsCache := &MockStatsCache{}

	controller := NewDoHController(mockResolver, "8.8.8.8:53", "example.com", mockStatsCache)

	req := httptest.NewRequest(http.MethodPut, "/dns-query", nil)
	w := httptest.NewRecorder()

	controller.DNSQuery(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestDOHController_DNSQuery_GET_DecodeError(t *testing.T) {
	mockResolver := &MockResolver{}
	mockStatsCache := &MockStatsCache{}

	controller := NewDoHController(mockResolver, "8.8.8.8:53", "example.com", mockStatsCache)

	req := httptest.NewRequest(http.MethodGet, "/dns-query?dns=invalid_base64!", nil)
	w := httptest.NewRecorder()

	controller.DNSQuery(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestDOHController_DNSQuery_POST_ReadError(t *testing.T) {
	mockResolver := &MockResolver{}
	mockStatsCache := &MockStatsCache{}

	controller := NewDoHController(mockResolver, "8.8.8.8:53", "example.com", mockStatsCache)

	req := httptest.NewRequest(http.MethodPost, "/dns-query", &errorReader{})
	w := httptest.NewRecorder()

	controller.DNSQuery(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDOHController_DNSQuery_UnpackError(t *testing.T) {
	mockResolver := &MockResolver{}
	mockStatsCache := &MockStatsCache{}

	controller := NewDoHController(mockResolver, "8.8.8.8:53", "example.com", mockStatsCache)

	// Invalid DNS message
	req := httptest.NewRequest(http.MethodPost, "/dns-query", bytes.NewReader([]byte("not a dns msg")))
	w := httptest.NewRecorder()

	controller.DNSQuery(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestDOHController_DNSQuery_ResolveError(t *testing.T) {
	mockResolver := &MockResolver{
		ResolveFunc: func(msg *dns.Msg, deviceHash string) (*dns.Msg, error) {
			return nil, errors.New("resolve error")
		},
	}
	mockStatsCache := &MockStatsCache{}

	controller := NewDoHController(mockResolver, "8.8.8.8:53", "example.com", mockStatsCache)

	dnsMsg := createValidDNSMsg()

	req := httptest.NewRequest(http.MethodPost, "/dns-query", bytes.NewReader(dnsMsg))
	w := httptest.NewRecorder()

	controller.DNSQuery(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestDOHController_DNSQuery_DeviceHash(t *testing.T) {
	mockResolver := &MockResolver{
		ResolveFunc: func(msg *dns.Msg, deviceHash string) (*dns.Msg, error) {
			if deviceHash != "test-device-hash" {
				t.Errorf("Expected device hash 'test-device-hash', got '%s'", deviceHash)
			}
			resp := new(dns.Msg)
			resp.SetReply(msg)
			return resp, nil
		},
	}
	mockStatsCache := &MockStatsCache{}

	controller := NewDoHController(mockResolver, "8.8.8.8:53", "example.com", mockStatsCache)

	dnsMsg := createValidDNSMsg()

	req := httptest.NewRequest(http.MethodPost, "/dns-query", bytes.NewReader(dnsMsg))
	req = req.WithContext(context.WithValue(req.Context(), middleware.DeviceKey, "test-device-hash"))
	w := httptest.NewRecorder()

	controller.DNSQuery(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
