package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miekg/dns"
)

type mockSettingsRepo struct {
	value string
}

func (m *mockSettingsRepo) GetUpstreamDNS(fallback string) (string, error) {
	if m.value == "" {
		return fallback, nil
	}
	return m.value, nil
}

func (m *mockSettingsRepo) UpdateUpstreamDNS(value string) error {
	m.value = value
	return nil
}

type mockMutableResolver struct {
	upstreamDNS string
}

func (m *mockMutableResolver) Resolve(*dns.Msg, string) (*dns.Msg, error) {
	return nil, nil
}

func (m *mockMutableResolver) SetUpstreamDNS(value string) {
	m.upstreamDNS = value
}

func (m *mockMutableResolver) UpstreamDNS() string {
	return m.upstreamDNS
}

func TestSettingsControllerUpdateSettings(t *testing.T) {
	repo := &mockSettingsRepo{}
	resolver := &mockMutableResolver{upstreamDNS: "9.9.9.9:53"}
	controller := NewSettingsController(repo, resolver, mockAuth(1))
	req := httptest.NewRequest(http.MethodPatch, "/api/settings", strings.NewReader(`{"upstreamDns":"1.1.1.1:53"}`))
	rr := httptest.NewRecorder()

	controller.UpdateSettings(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", rr.Code, rr.Body.String())
	}
	if repo.value != "1.1.1.1:53" || resolver.UpstreamDNS() != "1.1.1.1:53" {
		t.Fatalf("setting was not persisted and applied: repo=%q resolver=%q", repo.value, resolver.UpstreamDNS())
	}
}

func TestSettingsControllerRejectsInvalidUpstreamDNS(t *testing.T) {
	repo := &mockSettingsRepo{}
	resolver := &mockMutableResolver{upstreamDNS: "9.9.9.9:53"}
	controller := NewSettingsController(repo, resolver, mockAuth(1))
	req := httptest.NewRequest(http.MethodPatch, "/api/settings", strings.NewReader(`{"upstreamDns":"example.com"}`))
	rr := httptest.NewRecorder()

	controller.UpdateSettings(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rr.Code)
	}
	if resolver.UpstreamDNS() != "9.9.9.9:53" {
		t.Fatalf("invalid setting was applied: %q", resolver.UpstreamDNS())
	}
}
