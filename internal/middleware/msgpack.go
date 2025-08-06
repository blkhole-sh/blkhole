// Package middleware provides msgpack content-type middleware for HTTP requests.
package middleware

import (
	"net/http"
)

// MsgPackMiddleware sets the content-type to application/msgpack
func MsgPackMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/msgpack")
		next.ServeHTTP(w, r)
	})
}