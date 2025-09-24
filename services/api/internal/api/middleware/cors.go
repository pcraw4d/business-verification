package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CORSConfig holds configuration for CORS policy
type CORSConfig struct {
	// Allowed Origins
	AllowedOrigins  []string `json:"allowed_origins" yaml:"allowed_origins"`
	AllowAllOrigins bool     `json:"allow_all_origins" yaml:"allow_all_origins"`

	// Allowed Methods
	AllowedMethods []string `json:"allowed_methods" yaml:"allowed_methods"`

	// Allowed Headers
	AllowedHeaders []string `json:"allowed_headers" yaml:"allowed_headers"`

	// Exposed Headers
	ExposedHeaders []string `json:"exposed_headers" yaml:"exposed_headers"`

	// Credentials
	AllowCredentials bool `json:"allow_credentials" yaml:"allow_credentials"`

	// Preflight Cache
	MaxAge time.Duration `json:"max_age" yaml:"max_age"`

	// Debug Mode
	Debug bool `json:"debug" yaml:"debug"`

	// Path-based CORS rules
	PathRules []CORSPathRule `json:"path_rules" yaml:"path_rules"`
}

// CORSPathRule defines CORS rules for specific paths
type CORSPathRule struct {
	Path             string        `json:"path" yaml:"path"`
	AllowedOrigins   []string      `json:"allowed_origins" yaml:"allowed_origins"`
	AllowedMethods   []string      `json:"allowed_methods" yaml:"allowed_methods"`
	AllowedHeaders   []string      `json:"allowed_headers" yaml:"allowed_headers"`
	ExposedHeaders   []string      `json:"exposed_headers" yaml:"exposed_headers"`
	AllowCredentials bool          `json:"allow_credentials" yaml:"allow_credentials"`
	MaxAge           time.Duration `json:"max_age" yaml:"max_age"`
}

// CORSMiddleware provides CORS policy enforcement
type CORSMiddleware struct {
	config *CORSConfig
	logger *zap.Logger
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(config *CORSConfig, logger *zap.Logger) *CORSMiddleware {
	if config == nil {
		config = &CORSConfig{
			AllowedOrigins:   []string{"http://localhost:3000", "https://localhost:3000"},
			AllowAllOrigins:  false,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
			ExposedHeaders:   []string{"X-Total-Count", "X-Page-Count"},
			AllowCredentials: true,
			MaxAge:           86400 * time.Second, // 24 hours
			Debug:            false,
			PathRules:        []CORSPathRule{},
		}
	}

	return &CORSMiddleware{
		config: config,
		logger: logger,
	}
}

// Middleware applies CORS policy to HTTP responses
func (m *CORSMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Find applicable path rule
		pathRule := m.findPathRule(r.URL.Path)

		// Get CORS configuration for this request
		corsConfig := m.getCORSConfig(pathRule)

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			m.handlePreflight(w, r, corsConfig)
			return
		}

		// Handle actual requests
		m.handleActualRequest(w, r, corsConfig)

		next.ServeHTTP(w, r)
	})
}

// findPathRule finds the applicable CORS path rule for the given path
func (m *CORSMiddleware) findPathRule(path string) *CORSPathRule {
	for _, rule := range m.config.PathRules {
		if strings.HasPrefix(path, rule.Path) {
			return &rule
		}
	}
	return nil
}

// getCORSConfig returns the CORS configuration for the request
func (m *CORSMiddleware) getCORSConfig(pathRule *CORSPathRule) *CORSConfig {
	if pathRule == nil {
		return m.config
	}

	// Merge path rule with global config
	config := &CORSConfig{
		AllowedOrigins:   m.config.AllowedOrigins,
		AllowAllOrigins:  m.config.AllowAllOrigins,
		AllowedMethods:   m.config.AllowedMethods,
		AllowedHeaders:   m.config.AllowedHeaders,
		ExposedHeaders:   m.config.ExposedHeaders,
		AllowCredentials: m.config.AllowCredentials,
		MaxAge:           m.config.MaxAge,
		Debug:            m.config.Debug,
	}

	// Override with path-specific settings
	if len(pathRule.AllowedOrigins) > 0 {
		config.AllowedOrigins = pathRule.AllowedOrigins
		// Check if path rule allows all origins
		for _, origin := range pathRule.AllowedOrigins {
			if origin == "*" {
				config.AllowAllOrigins = true
				break
			}
		}
	}
	if len(pathRule.AllowedMethods) > 0 {
		config.AllowedMethods = pathRule.AllowedMethods
	}
	if len(pathRule.AllowedHeaders) > 0 {
		config.AllowedHeaders = pathRule.AllowedHeaders
	}
	if len(pathRule.ExposedHeaders) > 0 {
		config.ExposedHeaders = pathRule.ExposedHeaders
	}
	if pathRule.MaxAge > 0 {
		config.MaxAge = pathRule.MaxAge
	}

	return config
}

// handlePreflight handles OPTIONS preflight requests
func (m *CORSMiddleware) handlePreflight(w http.ResponseWriter, r *http.Request, config *CORSConfig) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	if !m.isOriginAllowed(origin, config) {
		if m.config.Debug {
			m.logger.Warn("CORS preflight rejected - origin not allowed",
				zap.String("origin", origin),
				zap.String("path", r.URL.Path))
		}
		http.Error(w, "Origin not allowed", http.StatusForbidden)
		return
	}

	// Set CORS headers
	m.setCORSHeaders(w, r, config)

	// Set preflight cache
	if config.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", int64(config.MaxAge.Seconds())))
	}

	// Respond to preflight
	w.WriteHeader(http.StatusNoContent)
}

// handleActualRequest handles actual CORS requests
func (m *CORSMiddleware) handleActualRequest(w http.ResponseWriter, r *http.Request, config *CORSConfig) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	if !m.isOriginAllowed(origin, config) {
		if m.config.Debug {
			m.logger.Warn("CORS request rejected - origin not allowed",
				zap.String("origin", origin),
				zap.String("path", r.URL.Path))
		}
		return // Don't set CORS headers for disallowed origins
	}

	// Set CORS headers
	m.setCORSHeaders(w, r, config)
}

// isOriginAllowed checks if the origin is allowed
func (m *CORSMiddleware) isOriginAllowed(origin string, config *CORSConfig) bool {
	if config.AllowAllOrigins {
		return true
	}

	if origin == "" {
		return false
	}

	for _, allowedOrigin := range config.AllowedOrigins {
		if m.matchesOrigin(origin, allowedOrigin) {
			return true
		}
	}

	return false
}

// matchesOrigin checks if an origin matches a pattern
func (m *CORSMiddleware) matchesOrigin(origin, pattern string) bool {
	// Exact match
	if origin == pattern {
		return true
	}

	// Wildcard subdomain match (e.g., *.example.com)
	if strings.HasPrefix(pattern, "*.") {
		domain := pattern[2:] // Remove "*. "
		if strings.HasSuffix(origin, domain) && !strings.Contains(origin[:len(origin)-len(domain)], ".") {
			return true
		}
	}

	// Wildcard match
	if pattern == "*" {
		return true
	}

	return false
}

// setCORSHeaders sets the appropriate CORS headers
func (m *CORSMiddleware) setCORSHeaders(w http.ResponseWriter, r *http.Request, config *CORSConfig) {
	origin := r.Header.Get("Origin")

	// Set Access-Control-Allow-Origin
	if config.AllowAllOrigins {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	// Set Access-Control-Allow-Methods
	if len(config.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
	}

	// Set Access-Control-Allow-Headers
	if len(config.AllowedHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
	}

	// Set Access-Control-Expose-Headers
	if len(config.ExposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
	}

	// Set Access-Control-Allow-Credentials
	if config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}

// GetDefaultCORSConfig returns a default CORS configuration
func GetDefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",
			"https://localhost:3000",
			"http://localhost:8080",
			"https://localhost:8080",
		},
		AllowAllOrigins: false,
		AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-API-Key",
			"X-CSRF-Token",
		},
		ExposedHeaders: []string{
			"X-Total-Count",
			"X-Page-Count",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials: true,
		MaxAge:           86400 * time.Second, // 24 hours
		Debug:            false,
		PathRules:        []CORSPathRule{},
	}
}

// GetStrictCORSConfig returns a strict CORS configuration for production
func GetStrictCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:  []string{}, // Must be explicitly set
		AllowAllOrigins: false,
		AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-API-Key",
		},
		ExposedHeaders: []string{
			"X-Total-Count",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
		},
		AllowCredentials: true,
		MaxAge:           3600 * time.Second, // 1 hour
		Debug:            false,
		PathRules:        []CORSPathRule{},
	}
}

// GetDevelopmentCORSConfig returns a development-friendly CORS configuration
func GetDevelopmentCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins:  []string{"*"},
		AllowAllOrigins: true,
		AllowedMethods:  []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		AllowedHeaders: []string{
			"*",
		},
		ExposedHeaders: []string{
			"*",
		},
		AllowCredentials: true,
		MaxAge:           300 * time.Second, // 5 minutes
		Debug:            true,
		PathRules:        []CORSPathRule{},
	}
}
