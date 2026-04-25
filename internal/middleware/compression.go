package middleware

import (
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

// CompressionMiddleware handles serving compressed static files
// serveCompressed tries to serve a compressed version of the file
func serveCompressed(w http.ResponseWriter, webFS fs.FS, path, encoding, ext string) bool {
	compressedPath := path + ext
	if content, err := fs.ReadFile(webFS, compressedPath); err == nil {
		w.Header().Set("Content-Encoding", encoding)
		w.Header().Set("Vary", "Accept-Encoding")
		setContentType(w, path)
		w.Write(content)
		return true
	}
	return false
}

func CompressionMiddleware(webFS fs.FS) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := strings.TrimPrefix(r.URL.Path, "/")
			if path == "" {
				path = "index.html"
			}

			// Only apply compression to static assets
			if !isStaticAsset(path) {
				next.ServeHTTP(w, r)
				return
			}

			acceptEncoding := r.Header.Get("Accept-Encoding")

			// Try brotli first
			if strings.Contains(acceptEncoding, "br") {
				if serveCompressed(w, webFS, path, "br", ".br") {
					return
				}
			}

			// Fallback to gzip
			if strings.Contains(acceptEncoding, "gzip") {
				if serveCompressed(w, webFS, path, "gzip", ".gz") {
					return
				}
			}

			// Continue to next handler for uncompressed files
			next.ServeHTTP(w, r)
		})
	}
}

func isStaticAsset(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".js" || ext == ".css" || ext == ".svg" || ext == ".ico" || ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".webp"
}

func setContentType(w http.ResponseWriter, path string) {
	switch filepath.Ext(path) {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".webp":
		w.Header().Set("Content-Type", "image/webp")
	default:
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
}
