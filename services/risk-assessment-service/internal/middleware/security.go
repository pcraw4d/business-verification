package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SecurityConfig holds security middleware configuration
type SecurityConfig struct {
	AllowedOrigins           []string      `json:"allowed_origins"`
	AllowedMethods           []string      `json:"allowed_methods"`
	AllowedHeaders           []string      `json:"allowed_headers"`
	ExposedHeaders           []string      `json:"exposed_headers"`
	AllowCredentials         bool          `json:"allow_credentials"`
	MaxAge                   int           `json:"max_age"`
	RequestSizeLimit         int64         `json:"request_size_limit"`
	RequestTimeout           time.Duration `json:"request_timeout"`
	EnableHSTS               bool          `json:"enable_hsts"`
	HSTSMaxAge               int           `json:"hsts_max_age"`
	EnableCSP                bool          `json:"enable_csp"`
	CSPDirectives            string        `json:"csp_directives"`
	EnableXSSProtection      bool          `json:"enable_xss_protection"`
	EnableFrameOptions       bool          `json:"enable_frame_options"`
	FrameOptionsValue        string        `json:"frame_options_value"`
	EnableContentTypeOptions bool          `json:"enable_content_type_options"`
	EnableReferrerPolicy     bool          `json:"enable_referrer_policy"`
	ReferrerPolicyValue      string        `json:"referrer_policy_value"`
}

// SecurityMiddleware handles security headers and CORS
type SecurityMiddleware struct {
	config SecurityConfig
	logger *zap.Logger
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config SecurityConfig, logger *zap.Logger) *SecurityMiddleware {
	return &SecurityMiddleware{
		config: config,
		logger: logger,
	}
}

// SecurityMiddleware creates a middleware function for security headers and CORS
func (sm *SecurityMiddleware) SecurityMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set security headers
			sm.setSecurityHeaders(w, r)

			// Handle CORS
			sm.handleCORS(w, r)

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Set request size limit
			if sm.config.RequestSizeLimit > 0 {
				r.Body = http.MaxBytesReader(w, r.Body, sm.config.RequestSizeLimit)
			}

			// Set request timeout
			if sm.config.RequestTimeout > 0 {
				http.TimeoutHandler(next, sm.config.RequestTimeout, "Request timeout").ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// setSecurityHeaders sets various security headers
func (sm *SecurityMiddleware) setSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	// HSTS (HTTP Strict Transport Security)
	if sm.config.EnableHSTS {
		maxAge := sm.config.HSTSMaxAge
		if maxAge == 0 {
			maxAge = 31536000 // 1 year
		}
		w.Header().Set("Strict-Transport-Security",
			fmt.Sprintf("max-age=%d; includeSubDomains; preload", maxAge))
	}

	// Content Security Policy
	if sm.config.EnableCSP {
		csp := sm.config.CSPDirectives
		if csp == "" {
			csp = "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';"
		}
		w.Header().Set("Content-Security-Policy", csp)
	}

	// X-XSS-Protection
	if sm.config.EnableXSSProtection {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
	}

	// X-Frame-Options
	if sm.config.EnableFrameOptions {
		value := sm.config.FrameOptionsValue
		if value == "" {
			value = "DENY"
		}
		w.Header().Set("X-Frame-Options", value)
	}

	// X-Content-Type-Options
	if sm.config.EnableContentTypeOptions {
		w.Header().Set("X-Content-Type-Options", "nosniff")
	}

	// Referrer-Policy
	if sm.config.EnableReferrerPolicy {
		value := sm.config.ReferrerPolicyValue
		if value == "" {
			value = "strict-origin-when-cross-origin"
		}
		w.Header().Set("Referrer-Policy", value)
	}

	// Additional security headers
	w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
	w.Header().Set("X-Download-Options", "noopen")
	w.Header().Set("X-DNS-Prefetch-Control", "off")

	// Remove server information
	w.Header().Set("Server", "")

	// Cache control for sensitive endpoints
	if sm.isSensitiveEndpoint(r.URL.Path) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
	}
}

// handleCORS handles Cross-Origin Resource Sharing
func (sm *SecurityMiddleware) handleCORS(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	if sm.isOriginAllowed(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else if len(sm.config.AllowedOrigins) == 0 {
		// If no specific origins configured, allow all (for development)
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	// Set CORS headers
	if len(sm.config.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(sm.config.AllowedMethods, ", "))
	} else {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	}

	if len(sm.config.AllowedHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(sm.config.AllowedHeaders, ", "))
	} else {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With, X-API-Key, X-User-ID, X-Tenant-ID, X-Service-Token, X-Service-ID")
	}

	if len(sm.config.ExposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(sm.config.ExposedHeaders, ", "))
	} else {
		w.Header().Set("Access-Control-Expose-Headers", "X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset, X-RateLimit-Tier")
	}

	if sm.config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	if sm.config.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", sm.config.MaxAge))
	}
}

// isOriginAllowed checks if an origin is in the allowed list
func (sm *SecurityMiddleware) isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range sm.config.AllowedOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			return true
		}

		// Support for wildcard subdomains
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := strings.TrimPrefix(allowedOrigin, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}

// isSensitiveEndpoint checks if an endpoint contains sensitive data
func (sm *SecurityMiddleware) isSensitiveEndpoint(path string) bool {
	sensitivePaths := []string{
		"/api/v1/assess",
		"/api/v1/predict",
		"/api/v1/batch",
		"/api/v1/webhooks",
		"/api/v1/reports",
		"/api/v1/dashboards",
	}

	for _, sensitivePath := range sensitivePaths {
		if strings.HasPrefix(path, sensitivePath) {
			return true
		}
	}

	return false
}

// DefaultSecurityConfig returns a default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		AllowedOrigins: []string{
			"https://*.railway.app",
			"https://*.vercel.app",
			"https://*.netlify.app",
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
		},
		AllowedMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
		},
		AllowedHeaders: []string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"X-API-Key",
			"X-User-ID",
			"X-Tenant-ID",
			"X-Service-Token",
			"X-Service-ID",
			"X-Request-ID",
			"X-Correlation-ID",
		},
		ExposedHeaders: []string{
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"X-RateLimit-Tier",
			"X-Request-ID",
			"X-Correlation-ID",
		},
		AllowCredentials:         true,
		MaxAge:                   86400,            // 24 hours
		RequestSizeLimit:         10 * 1024 * 1024, // 10MB
		RequestTimeout:           30 * time.Second,
		EnableHSTS:               true,
		HSTSMaxAge:               31536000, // 1 year
		EnableCSP:                true,
		CSPDirectives:            "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'; frame-ancestors 'none';",
		EnableXSSProtection:      true,
		EnableFrameOptions:       true,
		FrameOptionsValue:        "DENY",
		EnableContentTypeOptions: true,
		EnableReferrerPolicy:     true,
		ReferrerPolicyValue:      "strict-origin-when-cross-origin",
	}
}

// DevelopmentSecurityConfig returns a more permissive security configuration for development
func DevelopmentSecurityConfig() SecurityConfig {
	config := DefaultSecurityConfig()
	config.AllowedOrigins = []string{"*"}
	config.EnableHSTS = false
	config.EnableCSP = false
	config.RequestSizeLimit = 50 * 1024 * 1024 // 50MB for development
	return config
}

// ProductionSecurityConfig returns a strict security configuration for production
func ProductionSecurityConfig() SecurityConfig {
	config := DefaultSecurityConfig()
	config.AllowedOrigins = []string{
		"https://yourdomain.com",
		"https://app.yourdomain.com",
		"https://admin.yourdomain.com",
	}
	config.EnableHSTS = true
	config.EnableCSP = true
	config.RequestSizeLimit = 5 * 1024 * 1024 // 5MB for production
	config.RequestTimeout = 15 * time.Second
	return config
}
