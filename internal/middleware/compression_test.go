package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestCompressionMiddleware(t *testing.T) {
	// Create a mock file system
	mockFS := fstest.MapFS{
		"index.html":       {Data: []byte("<html><body>index</body></html>")},
		"app.js":           {Data: []byte("console.log('uncompressed');")},
		"app.js.br":        {Data: []byte("compressed_brotli_js")},
		"app.js.gz":        {Data: []byte("compressed_gzip_js")},
		"style.css":        {Data: []byte("body { color: red; }")},
		"style.css.gz":     {Data: []byte("compressed_gzip_css")},
		"image.svg":        {Data: []byte("<svg></svg>")},
		"image.svg.br":     {Data: []byte("compressed_brotli_svg")},
		"logo.png":         {Data: []byte("png_data")},
		"photo.jpg":        {Data: []byte("jpg_data")},
		"icon.ico":         {Data: []byte("ico_data")},
		"picture.webp":     {Data: []byte("webp_data")},
	}

	// Create a simple next handler that just writes "next_handler"
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("next_handler"))
	})

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
		expectNextHandler    bool
	}{
		{
			name:                "Static asset with Brotli support and .br file exists",
			path:                "/app.js",
			acceptEncoding:      "gzip, deflate, br",
			expectedStatus:      http.StatusOK,
			expectedBody:        "compressed_brotli_js",
			expectedEncoding:    "br",
			expectedContentType: "application/javascript",
			expectedVary:        "Accept-Encoding",
		},
		{
			name:                "Static asset with Gzip support and .gz file exists",
			path:                "/app.js",
			acceptEncoding:      "gzip, deflate",
			expectedStatus:      http.StatusOK,
			expectedBody:        "compressed_gzip_js",
			expectedEncoding:    "gzip",
			expectedContentType: "application/javascript",
			expectedVary:        "Accept-Encoding",
		},
		{
			name:                "Static asset with Brotli support but only .gz file exists",
			path:                "/style.css",
			acceptEncoding:      "gzip, deflate, br",
			expectedStatus:      http.StatusOK,
			expectedBody:        "compressed_gzip_css", // Fallback to gzip because style.css.br doesn't exist
			expectedEncoding:    "gzip",
			expectedContentType: "text/css",
			expectedVary:        "Accept-Encoding",
		},
		{
			name:                "Static asset with Brotli support and .br file exists (SVG)",
			path:                "/image.svg",
			acceptEncoding:      "br",
			expectedStatus:      http.StatusOK,
			expectedBody:        "compressed_brotli_svg",
			expectedEncoding:    "br",
			expectedContentType: "image/svg+xml",
			expectedVary:        "Accept-Encoding",
		},
		{
			name:              "Static asset but no Accept-Encoding header",
			path:              "/app.js",
			acceptEncoding:    "",
			expectedStatus:    http.StatusOK,
			expectedBody:      "next_handler",
			expectNextHandler: true,
		},
		{
			name:              "Static asset but no compressed files exist",
			path:              "/logo.png",
			acceptEncoding:    "gzip, br",
			expectedStatus:    http.StatusOK,
			expectedBody:      "next_handler",
			expectNextHandler: true,
		},
		{
			name:              "Non-static asset (HTML)",
			path:              "/index.html",
			acceptEncoding:    "gzip, br",
			expectedStatus:    http.StatusOK,
			expectedBody:      "next_handler",
			expectNextHandler: true,
		},
		{
			name:              "Empty path defaults to index.html (Non-static)",
			path:              "/",
			acceptEncoding:    "gzip, br",
			expectedStatus:    http.StatusOK,
			expectedBody:      "next_handler",
			expectNextHandler: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.acceptEncoding != "" {
				req.Header.Set("Accept-Encoding", tt.acceptEncoding)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			res := rr.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			body, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("failed to read response body: %v", err)
			}

			if string(body) != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, string(body))
			}

			if !tt.expectNextHandler {
				if encoding := res.Header.Get("Content-Encoding"); encoding != tt.expectedEncoding {
					t.Errorf("expected Content-Encoding %q, got %q", tt.expectedEncoding, encoding)
				}
				if contentType := res.Header.Get("Content-Type"); contentType != tt.expectedContentType {
					t.Errorf("expected Content-Type %q, got %q", tt.expectedContentType, contentType)
				}
				if vary := res.Header.Get("Vary"); vary != tt.expectedVary {
					t.Errorf("expected Vary %q, got %q", tt.expectedVary, vary)
				}
			} else {
				// If we expect the next handler to run, Content-Encoding and Vary from our middleware shouldn't be set
				if encoding := res.Header.Get("Content-Encoding"); encoding != "" {
					t.Errorf("expected empty Content-Encoding, got %q", encoding)
				}
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
		{"unknown.txt", "text/html; charset=utf-8"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			rr := httptest.NewRecorder()
			setContentType(rr, "file"+tt.path)

			if contentType := rr.Header().Get("Content-Type"); contentType != tt.expected {
				t.Errorf("setContentType(%q) = %q, want %q", tt.path, contentType, tt.expected)
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
		{"image.svg", true},
		{"favicon.ico", true},
		{"logo.png", true},
		{"photo.jpg", true},
		{"photo.jpeg", true},
		{"picture.webp", true},
		{"index.html", false},
		{"data.json", false},
		{"unknown.txt", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isStaticAsset(tt.path); got != tt.expected {
				t.Errorf("isStaticAsset(%q) = %v, want %v", tt.path, got, tt.expected)
			}
		})
	}
}
