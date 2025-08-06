// Package middleware provides authentication middleware for HTTP requests.
package middleware

import (
	"net/http"

	"github.com/go-chi/jwtauth/v5"
)

// CookieAuthenticator extracts JWT tokens from HttpOnly cookies and authenticates them
func CookieAuthenticator(ja *jwtauth.JWTAuth) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from HttpOnly cookie
			if cookie, err := r.Cookie("access_token"); err == nil && cookie.Value != "" {
				// Set the token in Authorization header format that jwtauth expects
				r.Header.Set("Authorization", "Bearer "+cookie.Value)
			}

			// Use the standard jwtauth verifier and authenticator
			jwtauth.Verifier(ja)(jwtauth.Authenticator(ja)(next)).ServeHTTP(w, r)
		})
	}
}

