package middleware

import (
	"net/http"
)

// BlockedPage redirects blocked domain requests to /blocked
func BlockedPage(domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if host matches blocked.domain
			if r.Host == "blocked."+domain {
				// Redirect to the /blocked route
				http.Redirect(w, r, "/blocked", http.StatusFound)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
