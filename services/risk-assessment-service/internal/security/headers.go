package security

import (
	"net/http"
	"strings"
	"time"
)

// SecurityHeadersManager manages security headers and CORS configuration
type SecurityHeadersManager struct {
	config *SecurityHeadersConfig
}

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	// CORS Configuration
	CORSEnabled      bool     `json:"cors_enabled"`
	AllowedOrigins   []string `json:"allowed_origins"`
	AllowedMethods   []string `json:"allowed_methods"`
	AllowedHeaders   []string `json:"allowed_headers"`
	ExposedHeaders   []string `json:"exposed_headers"`
	AllowCredentials bool     `json:"allow_credentials"`
	MaxAge           int      `json:"max_age"`

	// Security Headers
	ContentTypeOptions      string `json:"content_type_options"`
	FrameOptions            string `json:"frame_options"`
	XSSProtection           string `json:"xss_protection"`
	ReferrerPolicy          string `json:"referrer_policy"`
	PermissionsPolicy       string `json:"permissions_policy"`
	StrictTransportSecurity string `json:"strict_transport_security"`
	ContentSecurityPolicy   string `json:"content_security_policy"`

	// Additional Security
	ServerHeader string `json:"server_header"`
	CacheControl string `json:"cache_control"`
	Pragma       string `json:"pragma"`
	Expires      string `json:"expires"`
}

// NewSecurityHeadersManager creates a new security headers manager
func NewSecurityHeadersManager(config *SecurityHeadersConfig) *SecurityHeadersManager {
	if config == nil {
		config = &SecurityHeadersConfig{
			CORSEnabled:      true,
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
			ExposedHeaders:   []string{"X-Total-Count", "X-Page-Count"},
			AllowCredentials: false,
			MaxAge:           86400, // 24 hours

			ContentTypeOptions:      "nosniff",
			FrameOptions:            "DENY",
			XSSProtection:           "1; mode=block",
			ReferrerPolicy:          "strict-origin-when-cross-origin",
			PermissionsPolicy:       "geolocation=(), microphone=(), camera=()",
			StrictTransportSecurity: "max-age=31536000; includeSubDomains",
			ContentSecurityPolicy:   "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'",

			ServerHeader: "RiskAssessment/1.0",
			CacheControl: "no-cache, no-store, must-revalidate",
			Pragma:       "no-cache",
			Expires:      "0",
		}
	}

	return &SecurityHeadersManager{
		config: config,
	}
}

// SecurityHeadersMiddleware returns middleware for setting security headers
func (shm *SecurityHeadersManager) SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set security headers
			shm.setSecurityHeaders(w, r)

			// Handle CORS preflight requests
			if r.Method == "OPTIONS" && shm.config.CORSEnabled {
				shm.handleCORSPreflight(w, r)
				return
			}

			// Set CORS headers for actual requests
			if shm.config.CORSEnabled {
				shm.setCORSHeaders(w, r)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// setSecurityHeaders sets all security-related headers
func (shm *SecurityHeadersManager) setSecurityHeaders(w http.ResponseWriter, r *http.Request) {
	// Content Security Policy
	if shm.config.ContentSecurityPolicy != "" {
		w.Header().Set("Content-Security-Policy", shm.config.ContentSecurityPolicy)
	}

	// X-Content-Type-Options
	if shm.config.ContentTypeOptions != "" {
		w.Header().Set("X-Content-Type-Options", shm.config.ContentTypeOptions)
	}

	// X-Frame-Options
	if shm.config.FrameOptions != "" {
		w.Header().Set("X-Frame-Options", shm.config.FrameOptions)
	}

	// X-XSS-Protection
	if shm.config.XSSProtection != "" {
		w.Header().Set("X-XSS-Protection", shm.config.XSSProtection)
	}

	// Referrer-Policy
	if shm.config.ReferrerPolicy != "" {
		w.Header().Set("Referrer-Policy", shm.config.ReferrerPolicy)
	}

	// Permissions-Policy
	if shm.config.PermissionsPolicy != "" {
		w.Header().Set("Permissions-Policy", shm.config.PermissionsPolicy)
	}

	// Strict-Transport-Security (only for HTTPS)
	if r.TLS != nil && shm.config.StrictTransportSecurity != "" {
		w.Header().Set("Strict-Transport-Security", shm.config.StrictTransportSecurity)
	}

	// Server header
	if shm.config.ServerHeader != "" {
		w.Header().Set("Server", shm.config.ServerHeader)
	}

	// Cache control headers
	if shm.config.CacheControl != "" {
		w.Header().Set("Cache-Control", shm.config.CacheControl)
	}

	if shm.config.Pragma != "" {
		w.Header().Set("Pragma", shm.config.Pragma)
	}

	if shm.config.Expires != "" {
		w.Header().Set("Expires", shm.config.Expires)
	}

	// Additional security headers
	w.Header().Set("X-Powered-By", "")
	w.Header().Set("X-AspNet-Version", "")
	w.Header().Set("X-AspNetMvc-Version", "")
}

// setCORSHeaders sets CORS headers for cross-origin requests
func (shm *SecurityHeadersManager) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	if shm.isOriginAllowed(origin) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	// Set other CORS headers
	if len(shm.config.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(shm.config.AllowedMethods, ", "))
	}

	if len(shm.config.AllowedHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(shm.config.AllowedHeaders, ", "))
	}

	if len(shm.config.ExposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(shm.config.ExposedHeaders, ", "))
	}

	if shm.config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	if shm.config.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", string(rune(shm.config.MaxAge)))
	}
}

// handleCORSPreflight handles CORS preflight requests
func (shm *SecurityHeadersManager) handleCORSPreflight(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	if !shm.isOriginAllowed(origin) {
		http.Error(w, "CORS: Origin not allowed", http.StatusForbidden)
		return
	}

	// Set CORS headers
	shm.setCORSHeaders(w, r)

	// Set status to 200 OK
	w.WriteHeader(http.StatusOK)
}

// isOriginAllowed checks if the origin is allowed
func (shm *SecurityHeadersManager) isOriginAllowed(origin string) bool {
	if !shm.config.CORSEnabled {
		return false
	}

	// Allow all origins if configured
	if len(shm.config.AllowedOrigins) == 1 && shm.config.AllowedOrigins[0] == "*" {
		return true
	}

	// Check if origin is in allowed list
	for _, allowedOrigin := range shm.config.AllowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	return false
}

// UpdateConfig updates the security headers configuration
func (shm *SecurityHeadersManager) UpdateConfig(config *SecurityHeadersConfig) {
	if config != nil {
		shm.config = config
	}
}

// GetConfig returns the current configuration
func (shm *SecurityHeadersManager) GetConfig() *SecurityHeadersConfig {
	return shm.config
}

// ValidateConfig validates the security headers configuration
func (shm *SecurityHeadersManager) ValidateConfig() []string {
	var errors []string

	// Validate CORS configuration
	if shm.config.CORSEnabled {
		if len(shm.config.AllowedOrigins) == 0 {
			errors = append(errors, "CORS enabled but no allowed origins specified")
		}

		if len(shm.config.AllowedMethods) == 0 {
			errors = append(errors, "CORS enabled but no allowed methods specified")
		}

		if shm.config.AllowCredentials && len(shm.config.AllowedOrigins) == 1 && shm.config.AllowedOrigins[0] == "*" {
			errors = append(errors, "CORS credentials cannot be allowed with wildcard origin")
		}
	}

	// Validate security headers
	if shm.config.ContentTypeOptions != "nosniff" && shm.config.ContentTypeOptions != "" {
		errors = append(errors, "X-Content-Type-Options should be 'nosniff'")
	}

	if shm.config.FrameOptions != "DENY" && shm.config.FrameOptions != "SAMEORIGIN" && shm.config.FrameOptions != "" {
		errors = append(errors, "X-Frame-Options should be 'DENY' or 'SAMEORIGIN'")
	}

	return errors
}

// CreateStrictConfig creates a strict security configuration
func CreateStrictConfig() *SecurityHeadersConfig {
	return &SecurityHeadersConfig{
		CORSEnabled:      false,
		AllowedOrigins:   []string{},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           0,

		ContentTypeOptions:      "nosniff",
		FrameOptions:            "DENY",
		XSSProtection:           "1; mode=block",
		ReferrerPolicy:          "no-referrer",
		PermissionsPolicy:       "geolocation=(), microphone=(), camera=(), payment=(), usb=()",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
		ContentSecurityPolicy:   "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'",

		ServerHeader: "",
		CacheControl: "no-cache, no-store, must-revalidate, private",
		Pragma:       "no-cache",
		Expires:      "0",
	}
}

// CreatePermissiveConfig creates a permissive security configuration
func CreatePermissiveConfig() *SecurityHeadersConfig {
	return &SecurityHeadersConfig{
		CORSEnabled:      true,
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{"*"},
		AllowCredentials: false,
		MaxAge:           86400,

		ContentTypeOptions:      "nosniff",
		FrameOptions:            "SAMEORIGIN",
		XSSProtection:           "1; mode=block",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "geolocation=(), microphone=(), camera=()",
		StrictTransportSecurity: "max-age=31536000",
		ContentSecurityPolicy:   "default-src 'self' 'unsafe-inline' 'unsafe-eval'",

		ServerHeader: "RiskAssessment/1.0",
		CacheControl: "public, max-age=3600",
		Pragma:       "",
		Expires:      time.Now().Add(time.Hour).Format(time.RFC1123),
	}
}

// CreateDevelopmentConfig creates a development-friendly security configuration
func CreateDevelopmentConfig() *SecurityHeadersConfig {
	return &SecurityHeadersConfig{
		CORSEnabled:      true,
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080", "http://127.0.0.1:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With", "Accept", "Origin"},
		ExposedHeaders:   []string{"X-Total-Count", "X-Page-Count", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           3600,

		ContentTypeOptions:      "nosniff",
		FrameOptions:            "SAMEORIGIN",
		XSSProtection:           "1; mode=block",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "geolocation=(), microphone=(), camera=()",
		StrictTransportSecurity: "",
		ContentSecurityPolicy:   "default-src 'self' 'unsafe-inline' 'unsafe-eval' data: blob:",

		ServerHeader: "RiskAssessment-Dev/1.0",
		CacheControl: "no-cache",
		Pragma:       "no-cache",
		Expires:      "0",
	}
}
