package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/blkhole-sh/blkhole/internal/model"
	"github.com/blkhole-sh/blkhole/internal/repos"
	"github.com/blkhole-sh/blkhole/internal/services"
)

type mockBrowserService struct {
	createPairing func(int) (*services.BrowserPairing, error)
	pair          func(string, string, string) (*services.BrowserPairResult, error)
	authenticate  func(string) (*model.BrowserClient, *model.Device, error)
	rules         func(string) ([]string, error)
}

func (m *mockBrowserService) CreatePairing(deviceID int) (*services.BrowserPairing, error) {
	return m.createPairing(deviceID)
}
func (m *mockBrowserService) Pair(token, name, browser string) (*services.BrowserPairResult, error) {
	if m.pair != nil {
		return m.pair(token, name, browser)
	}
	return nil, repos.ErrInvalidBrowserPairing
}

func TestBrowserControllerRateLimitsFailedPairingExchange(t *testing.T) {
	service := &mockBrowserService{pair: func(string, string, string) (*services.BrowserPairResult, error) {
		return nil, repos.ErrInvalidBrowserPairing
	}}
	controller := NewBrowserController(service, &MockDeviceRepo{}, mockAuth(1))
	for attempt := 1; attempt <= browserFailureLimit+1; attempt++ {
		request := httptest.NewRequest(http.MethodPost, "/api/browser/v1/pair", strings.NewReader(`{"pairingToken":"bad","clientName":"Test"}`))
		request.RemoteAddr = "192.0.2.1:1234"
		response := httptest.NewRecorder()
		controller.Pair(response, request)
		want := http.StatusUnauthorized
		if attempt > browserFailureLimit {
			want = http.StatusTooManyRequests
		}
		if response.Code != want {
			t.Fatalf("attempt %d status = %d, want %d", attempt, response.Code, want)
		}
	}
}
func (m *mockBrowserService) Authenticate(token string) (*model.BrowserClient, *model.Device, error) {
	return m.authenticate(token)
}
func (m *mockBrowserService) ListClients(int) ([]*model.BrowserClient, error) {
	return []*model.BrowserClient{}, nil
}
func (m *mockBrowserService) RevokeClient(int, int) error { return nil }
func (m *mockBrowserService) Rules(deviceHash string) ([]string, error) {
	return m.rules(deviceHash)
}

func TestBrowserControllerCreatePairingRejectsOtherUsersDevice(t *testing.T) {
	devices := &MockDeviceRepo{FindByIDFunc: func(id int) (*model.Device, error) {
		return &model.Device{ID: id, UserID: 2}, nil
	}}
	service := &mockBrowserService{createPairing: func(int) (*services.BrowserPairing, error) {
		t.Fatal("CreatePairing must not be called")
		return nil, nil
	}}
	controller := NewBrowserController(service, devices, mockAuth(1))
	request := withParam(httptest.NewRequest(http.MethodPost, "/api/devices/7/browser-pairings", nil), "id", "7")
	response := httptest.NewRecorder()

	controller.CreatePairing(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}

func TestBrowserControllerRulesETag(t *testing.T) {
	service := &mockBrowserService{
		authenticate: func(token string) (*model.BrowserClient, *model.Device, error) {
			if token != "secret" {
				t.Fatalf("token = %q, want secret", token)
			}
			return &model.BrowserClient{ID: 1}, &model.Device{ID: 2, Hash: "device-hash"}, nil
		},
		rules: func(deviceHash string) ([]string, error) {
			return []string{"a.example.com", "b.example.com"}, nil
		},
		createPairing: func(int) (*services.BrowserPairing, error) {
			return &services.BrowserPairing{ExpiresAt: time.Now()}, nil
		},
	}
	controller := NewBrowserController(service, &MockDeviceRepo{}, mockAuth(1))
	request := httptest.NewRequest(http.MethodGet, "/api/browser/v1/rules", nil)
	request.Header.Set("Authorization", "Bearer secret")
	response := httptest.NewRecorder()
	controller.Rules(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", response.Code)
	}
	etag := response.Header().Get("ETag")
	if etag == "" {
		t.Fatal("missing ETag")
	}

	request = httptest.NewRequest(http.MethodGet, "/api/browser/v1/rules", nil)
	request.Header.Set("Authorization", "Bearer secret")
	request.Header.Set("If-None-Match", etag)
	response = httptest.NewRecorder()
	controller.Rules(response, request)
	if response.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", response.Code)
	}
}
