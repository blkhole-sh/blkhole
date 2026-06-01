package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeviceHashFromHost(t *testing.T) {
	domain := "example.com"
	handler := DeviceHashFromHost(domain)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hash, _ := r.Context().Value(DeviceKey).(string)
		w.Header().Set("X-Device-Hash", hash)
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		host           string
		expectedStatus int
		expectedHash   string
	}{
		{
			name:           "Exact domain match sets empty hash",
			host:           "example.com",
			expectedStatus: http.StatusOK,
			expectedHash:   "",
		},
		{
			name:           "Subdomain extracts hash",
			host:           "abc123.example.com",
			expectedStatus: http.StatusOK,
			expectedHash:   "abc123",
		},
		{
			name:           "Host with port strips port before matching",
			host:           "abc123.example.com:8080",
			expectedStatus: http.StatusOK,
			expectedHash:   "abc123",
		},
		{
			name:           "Exact domain with port returns empty hash",
			host:           "example.com:443",
			expectedStatus: http.StatusOK,
			expectedHash:   "",
		},
		{
			name:           "Unrelated host returns 403",
			host:           "other.com",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Partial suffix mismatch returns 403",
			host:           "notexample.com",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = tt.host
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if tt.expectedStatus == http.StatusOK {
				if got := rr.Header().Get("X-Device-Hash"); got != tt.expectedHash {
					t.Errorf("expected device hash %q, got %q", tt.expectedHash, got)
				}
			}
		})
	}
}
