package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestCompressionMiddleware(t *testing.T) {
	// Create an in-memory filesystem with test assets
	mockFS := fstest.MapFS{
		"app.js":       {Data: []byte("uncompressed app.js")},
		"app.js.br":    {Data: []byte("brotli compressed app.js")},
		"app.js.gz":    {Data: []byte("gzip compressed app.js")},
		"style.css":    {Data: []byte("uncompressed style.css")},
		"style.css.gz": {Data: []byte("gzip compressed style.css")}, // Note: no .br for style.css
		"image.png":    {Data: []byte("uncompressed image.png")}, // Note: no compressed versions
		"index.html":   {Data: []byte("uncompressed index.html")},
	}

	// Create a dummy next handler that always returns "fallthrough"
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Next-Called", "true")
		w.Write([]byte("fallthrough"))
	})

	// Initialize the middleware
	middleware := CompressionMiddleware(mockFS)
	handler := middleware(nextHandler)

	tests := []struct {
		name                 string
		path                 string
		acceptEncoding       string
		expectedStatus       int
		expectedBody         string
		expectedEncoding     string
		expectedContentType  string
		expectedVary         string
		expectNextCalled     bool
	}{
		{
			name:                 "Brotli Compression - Static Asset",
			path:                 "/app.js",
			acceptEncoding:       "gzip, deflate, br",
			expectedStatus:       http.StatusOK,
			expectedBody:         "brotli compressed app.js",
			expectedEncoding:     "br",
			expectedContentType:  "application/javascript",
			expectedVary:         "Accept-Encoding",
			expectNextCalled:     false,
		},
		{
			name:                 "Gzip Compression - Static Asset",
			path:                 "/app.js",
			acceptEncoding:       "gzip, deflate",
			expectedStatus:       http.StatusOK,
			expectedBody:         "gzip compressed app.js",
			expectedEncoding:     "gzip",
			expectedContentType:  "application/javascript",
			expectedVary:         "Accept-Encoding",
			expectNextCalled:     false,
		},
		{
			name:                 "Fallback to Gzip if Brotli file missing",
			path:                 "/style.css",
			acceptEncoding:       "gzip, deflate, br",
			expectedStatus:       http.StatusOK,
			expectedBody:         "gzip compressed style.css",
			expectedEncoding:     "gzip",
			expectedContentType:  "text/css",
			expectedVary:         "Accept-Encoding",
			expectNextCalled:     false,
		},
		{
			name:                 "No Accept-Encoding - Fallthrough",
			path:                 "/app.js",
			acceptEncoding:       "",
			expectedStatus:       http.StatusOK,
			expectedBody:         "fallthrough",
			expectedEncoding:     "",
			expectedContentType:  "text/plain; charset=utf-8",
			expectedVary:         "",
			expectNextCalled:     true,
		},
		{
			name:                 "Not a Static Asset - Fallthrough",
			path:                 "/index.html",
			acceptEncoding:       "gzip, deflate, br",
			expectedStatus:       http.StatusOK,
			expectedBody:         "fallthrough",
			expectedEncoding:     "",
			expectedContentType:  "text/plain; charset=utf-8",
			expectedVary:         "",
			expectNextCalled:     true,
		},
		{
			name:                 "Missing Compressed Files - Fallthrough",
			path:                 "/image.png",
			acceptEncoding:       "gzip, deflate, br",
			expectedStatus:       http.StatusOK,
			expectedBody:         "fallthrough",
			expectedEncoding:     "",
			expectedContentType:  "text/plain; charset=utf-8",
			expectedVary:         "",
			expectNextCalled:     true,
		},
		{
			name:                 "Root Path resolves to index.html - Fallthrough",
			path:                 "/",
			acceptEncoding:       "gzip, deflate, br",
			expectedStatus:       http.StatusOK,
			expectedBody:         "fallthrough",
			expectedEncoding:     "",
			expectedContentType:  "text/plain; charset=utf-8",
			expectedVary:         "",
			expectNextCalled:     true,
		},
		{
			name:                 "File Not Found - Fallthrough",
			path:                 "/notfound.js",
			acceptEncoding:       "gzip, deflate, br",
			expectedStatus:       http.StatusOK,
			expectedBody:         "fallthrough",
			expectedEncoding:     "",
			expectedContentType:  "text/plain; charset=utf-8",
			expectedVary:         "",
			expectNextCalled:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}

			if encoding := rr.Header().Get("Content-Encoding"); encoding != tt.expectedEncoding {
				t.Errorf("expected Content-Encoding %q, got %q", tt.expectedEncoding, encoding)
			}

			if contentType := rr.Header().Get("Content-Type"); contentType != tt.expectedContentType {
				t.Errorf("expected Content-Type %q, got %q", tt.expectedContentType, contentType)
			}

			if vary := rr.Header().Get("Vary"); vary != tt.expectedVary {
				t.Errorf("expected Vary %q, got %q", tt.expectedVary, vary)
			}

			nextCalled := rr.Header().Get("X-Next-Called") == "true"
			if nextCalled != tt.expectNextCalled {
				t.Errorf("expected nextHandler called: %v, got: %v", tt.expectNextCalled, nextCalled)
			}
		})
	}
}

func TestSetContentType(t *testing.T) {
	tests := []struct {
		path         string
		expectedType string
	}{
		{"style.css", "text/css"},
		{"app.js", "application/javascript"},
		{"icon.svg", "image/svg+xml"},
		{"favicon.ico", "image/x-icon"},
		{"image.png", "image/png"},
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"image.webp", "image/webp"},
		{"unknown.txt", "text/html; charset=utf-8"},
		{"noext", "text/html; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			setContentType(rr, tt.path)
			if ct := rr.Header().Get("Content-Type"); ct != tt.expectedType {
				t.Errorf("setContentType(%q) = %q, want %q", tt.path, ct, tt.expectedType)
			}
		})
	}
}

func TestIsStaticAsset(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"app.js", true},
		{"style.css", true},
		{"icon.svg", true},
		{"favicon.ico", true},
		{"image.png", true},
		{"photo.jpg", true},
		{"photo.jpeg", true},
		{"image.webp", true},
		{"index.html", false},
		{"data.json", false},
		{"unknown.txt", false},
		{"noext", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isStaticAsset(tt.path); got != tt.expected {
				t.Errorf("isStaticAsset(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
