package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func TestBrowserExtensionCORSPreflightThroughRouter(t *testing.T) {
	router := chi.NewRouter()
	router.With(BrowserExtensionCORS).Options("/api/browser/v1/pair", func(http.ResponseWriter, *http.Request) {})

	request := httptest.NewRequest(http.MethodOptions, "/api/browser/v1/pair", nil)
	request.Header.Set("Origin", "chrome-extension://abcdefghijklmnop")
	request.Header.Set("Access-Control-Request-Method", http.MethodPost)
	request.Header.Set("Access-Control-Request-Headers", "content-type")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "chrome-extension://abcdefghijklmnop" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
}

func TestBrowserExtensionPreflightRunsBeforeGlobalCORS(t *testing.T) {
	router := chi.NewRouter()
	router.Use(BrowserExtensionPreflight)
	router.Use(cors.Handler(cors.Options{AllowedOrigins: []string{"https://dashboard.example.com"}}))
	router.Options("/api/browser/v1/rules", func(http.ResponseWriter, *http.Request) {})

	request := httptest.NewRequest(http.MethodOptions, "/api/browser/v1/rules", nil)
	request.Header.Set("Origin", "moz-extension://12345678-abcd")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", "authorization, if-none-match")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "moz-extension://12345678-abcd" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); !strings.Contains(got, "If-None-Match") {
		t.Fatalf("Access-Control-Allow-Headers = %q", got)
	}
}

func TestBrowserExtensionCORSRejectsWebOrigin(t *testing.T) {
	handler := BrowserExtensionCORS(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	request := httptest.NewRequest(http.MethodOptions, "/api/browser/v1/pair", nil)
	request.Header.Set("Origin", "https://example.com")
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", response.Code)
	}
}
