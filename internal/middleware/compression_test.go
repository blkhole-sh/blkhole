package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestCompressionMiddleware(t *testing.T) {
	mockFS := fstest.MapFS{
		"app.js":       &fstest.MapFile{Data: []byte("uncompressed js")},
		"app.js.br":    &fstest.MapFile{Data: []byte("brotli js")},
		"app.js.gz":    &fstest.MapFile{Data: []byte("gzip js")},
		"style.css":    &fstest.MapFile{Data: []byte("uncompressed css")},
		"style.css.gz": &fstest.MapFile{Data: []byte("gzip css")},
		"image.png":    &fstest.MapFile{Data: []byte("uncompressed png")},
		"index.html":   &fstest.MapFile{Data: []byte("html content")},
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Next-Handler", "called")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("next handler content"))
	})

	middleware := CompressionMiddleware(mockFS)(nextHandler)

	tests := []struct {
		name                string
		path                string
		acceptEncoding      string
		expectedStatus      int
		expectedBody        string
		expectedEncoding    string
		expectedContentType string
		expectNextHandler   bool
	}{
		{
			name:                "Brotli preferred when both supported and available",
			path:                "/app.js",
			acceptEncoding:      "gzip, deflate, br",
			expectedStatus:      http.StatusOK,
			expectedBody:        "brotli js",
			expectedEncoding:    "br",
			expectedContentType: "application/javascript",
			expectNextHandler:   false,
		},
		{
			name:                "Gzip fallback when Brotli not supported",
			path:                "/app.js",
			acceptEncoding:      "gzip, deflate",
			expectedStatus:      http.StatusOK,
			expectedBody:        "gzip js",
			expectedEncoding:    "gzip",
			expectedContentType: "application/javascript",
			expectNextHandler:   false,
		},
		{
			name:                "Gzip used when Brotli file not available",
			path:                "/style.css",
			acceptEncoding:      "gzip, deflate, br",
			expectedStatus:      http.StatusOK,
			expectedBody:        "gzip css",
			expectedEncoding:    "gzip",
			expectedContentType: "text/css",
			expectNextHandler:   false,
		},
		{
			name:                "No compression if neither supported",
			path:                "/app.js",
			acceptEncoding:      "identity",
			expectedStatus:      http.StatusOK,
			expectedBody:        "next handler content",
			expectedEncoding:    "",
			expectedContentType: "",
			expectNextHandler:   true,
		},
		{
			name:                "No compression if compressed files don't exist",
			path:                "/image.png",
			acceptEncoding:      "gzip, deflate, br",
			expectedStatus:      http.StatusOK,
			expectedBody:        "next handler content",
			expectedEncoding:    "",
			expectedContentType: "",
			expectNextHandler:   true,
		},
		{
			name:                "No compression for non-static assets",
			path:                "/index.html",
			acceptEncoding:      "gzip, deflate, br",
			expectedStatus:      http.StatusOK,
			expectedBody:        "next handler content",
			expectedEncoding:    "",
			expectedContentType: "",
			expectNextHandler:   true,
		},
		{
			name:                "No compression for root path (defaults to index.html)",
			path:                "/",
			acceptEncoding:      "gzip, deflate, br",
			expectedStatus:      http.StatusOK,
			expectedBody:        "next handler content",
			expectedEncoding:    "",
			expectedContentType: "",
			expectNextHandler:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			rr := httptest.NewRecorder()
			middleware.ServeHTTP(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response body: %v", err)
			}

			if resp.StatusCode != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, resp.StatusCode)
			}

			if string(body) != tt.expectedBody {
				t.Errorf("Expected body %q, got %q", tt.expectedBody, string(body))
			}

			if gotEncoding := resp.Header.Get("Content-Encoding"); gotEncoding != tt.expectedEncoding {
				t.Errorf("Expected Content-Encoding %q, got %q", tt.expectedEncoding, gotEncoding)
			}

			if tt.expectedEncoding != "" {
				if gotVary := resp.Header.Get("Vary"); gotVary != "Accept-Encoding" {
					t.Errorf("Expected Vary 'Accept-Encoding', got %q", gotVary)
				}
				if gotContentType := resp.Header.Get("Content-Type"); gotContentType != tt.expectedContentType {
					t.Errorf("Expected Content-Type %q, got %q", tt.expectedContentType, gotContentType)
				}
			}

			nextHandlerCalled := resp.Header.Get("X-Next-Handler") == "called"
			if nextHandlerCalled != tt.expectNextHandler {
				t.Errorf("Expected next handler called: %v, got: %v", tt.expectNextHandler, nextHandlerCalled)
			}
		})
	}
}

func TestSetContentType(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{".css", "text/css"},
		{".js", "application/javascript"},
		{".svg", "image/svg+xml"},
		{".ico", "image/x-icon"},
		{".png", "image/png"},
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".webp", "image/webp"},
		{".html", "text/html; charset=utf-8"},
		{"unknown", "text/html; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			setContentType(rr, tt.path)

			if got := rr.Header().Get("Content-Type"); got != tt.expected {
				t.Errorf("setContentType(%q) = %q, want %q", tt.path, got, tt.expected)
			}
		})
	}
}
