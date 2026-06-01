// Package middleware provides JSON content-type middleware for HTTP requests.
package middleware

import (
	"net/http"
)

// JSONMiddleware sets the content-type to application/json
func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
