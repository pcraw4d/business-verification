package middleware

import (
	"net/http"
	"strings"
	"time"
)

// SecurityHeaders adds security headers to HTTP responses and browser caching headers
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set security headers
		// HSTS (only for HTTPS)
		if r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// X-Frame-Options
		w.Header().Set("X-Frame-Options", "DENY")

		// X-Content-Type-Options
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Remove server information
		w.Header().Set("Server", "")

		// Additional security headers
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("X-Download-Options", "noopen")
		w.Header().Set("X-DNS-Prefetch-Control", "off")

		// Cache control for sensitive endpoints
		if isSensitiveEndpoint(r.URL.Path) {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		} else if r.Method == http.MethodGet {
			// Browser caching headers for non-sensitive GET requests
			// Health check and metrics can be cached briefly
			if strings.HasPrefix(r.URL.Path, "/health") || strings.HasPrefix(r.URL.Path, "/metrics") {
				w.Header().Set("Cache-Control", "public, max-age=60") // 1 minute cache
				w.Header().Set("Expires", time.Now().Add(60*time.Second).UTC().Format(http.TimeFormat))
			} else if strings.HasPrefix(r.URL.Path, "/api/v1/classify") {
				// Classification responses - short cache for GET requests (if any)
				w.Header().Set("Cache-Control", "public, max-age=30") // 30 second cache
			}
		}

		next.ServeHTTP(w, r)
	})
}

// isSensitiveEndpoint checks if an endpoint should have no-cache headers
func isSensitiveEndpoint(path string) bool {
	sensitivePaths := []string{
		"/api/v1/auth",
		"/api/v1/merchants",
		"/api/v1/risk",
	}
	for _, sensitivePath := range sensitivePaths {
		if len(path) >= len(sensitivePath) && path[:len(sensitivePath)] == sensitivePath {
			return true
		}
	}
	return false
}

