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

			// CRITICAL: Remove any existing CORS headers to prevent duplicates
			// This handles cases where Railway or other proxies may have set headers
			// Use Del() to remove all values, then Set() to set exactly one value
			w.Header().Del("Access-Control-Allow-Origin")
			w.Header().Del("Access-Control-Allow-Methods")
			w.Header().Del("Access-Control-Allow-Headers")
			w.Header().Del("Access-Control-Allow-Credentials")
			w.Header().Del("Access-Control-Max-Age")
			
			// Set the header exactly once based on configuration
			var originToSet string
			
			// CRITICAL: With AllowCredentials=true, we CANNOT use "*" - must use specific origin
			// Browsers reject "*" when credentials are allowed
			if cfg.AllowCredentials {
				// With credentials, always use the requesting origin (if present)
				// or the first allowed origin if no origin header
				if origin != "" {
					// Use the requesting origin (browsers require this with credentials)
					originToSet = origin
				} else if len(cfg.AllowedOrigins) > 0 && cfg.AllowedOrigins[0] != "*" {
					// No origin header, use first non-wildcard allowed origin
					originToSet = cfg.AllowedOrigins[0]
				}
				// If origin is empty and only "*" is allowed, we can't set it (browser will reject)
			} else {
				// Without credentials, can use "*" if configured
				if len(cfg.AllowedOrigins) == 1 && cfg.AllowedOrigins[0] == "*" {
					originToSet = "*"
				} else if isOriginAllowed(origin, cfg.AllowedOrigins) {
					originToSet = origin
				} else if origin != "" {
					// Default to allowing the requesting origin
					originToSet = origin
				}
			}
			
			// Set the header exactly once using Set() which overwrites any existing value
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
