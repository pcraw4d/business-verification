package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"kyb-platform/services/api-gateway/internal/config"
)

// CORS middleware handles Cross-Origin Resource Sharing
func CORS(cfg config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			origin := r.Header.Get("Origin")

			// Debug logging
			fmt.Printf("CORS: Request from origin: %s, Method: %s, Path: %s\n", origin, r.Method, r.URL.Path)
			fmt.Printf("CORS: Allowed origins: %v\n", cfg.AllowedOrigins)

			// Always set CORS headers (override any Railway settings)
			// CRITICAL: Only set the header ONCE to avoid duplicates
			// Check if header is already set (by Railway) and remove it first
			if existingOrigin := w.Header().Get("Access-Control-Allow-Origin"); existingOrigin != "" {
				w.Header().Del("Access-Control-Allow-Origin")
				fmt.Printf("CORS: Removed existing Access-Control-Allow-Origin header: %s\n", existingOrigin)
			}
			
			// Set the header exactly once based on configuration
			var originToSet string
			if cfg.AllowCredentials && origin != "" {
				// With credentials, must use specific origin, not "*"
				if isOriginAllowed(origin, cfg.AllowedOrigins) || len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
					originToSet = origin
				} else {
					originToSet = origin // Default to requesting origin
				}
			} else if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
				// Without credentials, can use "*"
				originToSet = "*"
			} else if isOriginAllowed(origin, cfg.AllowedOrigins) {
				originToSet = origin
			} else if origin != "" {
				// Default to allowing the requesting origin
				originToSet = origin
			} else {
				// No origin header, use wildcard if allowed
				if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
					originToSet = "*"
				}
			}
			
			// Set the header exactly once
			if originToSet != "" {
				w.Header().Set("Access-Control-Allow-Origin", originToSet)
				fmt.Printf("CORS: Set Access-Control-Allow-Origin to: %s\n", originToSet)
			}

			w.Header().Set("Access-Control-Allow-Methods", joinStrings(cfg.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", joinStrings(cfg.AllowedHeaders, ", "))

			if cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if cfg.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(cfg.MaxAge))
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				fmt.Printf("CORS: Handling preflight request\n")
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if an origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}

	return false
}

// joinStrings joins a slice of strings with a separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}

	return result
}
