package middleware

import (
	"net/http"
	"net/url"
	"strings"
)

// BrowserExtensionPreflight must run before the application's global CORS
// middleware, which intentionally rejects non-dashboard origins.
func BrowserExtensionPreflight(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions && strings.HasPrefix(r.URL.Path, "/api/browser/v1/") {
			BrowserExtensionCORS(next).ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// BrowserExtensionCORS allows only browser-extension origins on the public
// pairing and rules endpoints. It never enables credentialed requests.
func BrowserExtensionCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && IsBrowserExtensionOrigin(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, If-None-Match")
			w.Header().Set("Access-Control-Expose-Headers", "ETag")
			w.Header().Add("Vary", "Origin")
		}
		if r.Method == http.MethodOptions {
			if origin == "" || !IsBrowserExtensionOrigin(origin) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// IsBrowserExtensionOrigin validates origins emitted by supported extension APIs.
func IsBrowserExtensionOrigin(origin string) bool {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Host == "" || parsed.Path != "" {
		return false
	}
	switch parsed.Scheme {
	case "chrome-extension", "moz-extension", "safari-web-extension":
		return true
	default:
		return false
	}
}
