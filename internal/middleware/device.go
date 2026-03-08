package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const DeviceKey contextKey = "device"

// DeviceHashFromHost extracts a device hash from the subdomain.
//
// If the host matches the domain exactly → deviceHash is empty
// If the host is a subdomain of the domain → deviceHash = subdomain
// Otherwise → returns HTTP 403 Forbidden
func DeviceHashFromHost(domain string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := strings.Split(r.Host, ":")[0]
			deviceHash := ""
			if host != domain && strings.HasSuffix(host, "."+domain) {
				deviceHash = strings.TrimSuffix(host, "."+domain)
			} else if host != domain {
				http.Error(w, "invalid host", http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), DeviceKey, deviceHash)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
