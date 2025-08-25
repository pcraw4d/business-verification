package middleware

import (
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SecurityHeadersConfig holds configuration for security headers
type SecurityHeadersConfig struct {
	// Content Security Policy
	CSPEnabled    bool   `json:"csp_enabled" yaml:"csp_enabled"`
	CSPDirectives string `json:"csp_directives" yaml:"csp_directives"`

	// HTTP Strict Transport Security
	HSTSEnabled           bool          `json:"hsts_enabled" yaml:"hsts_enabled"`
	HSTSMaxAge            time.Duration `json:"hsts_max_age" yaml:"hsts_max_age"`
	HSTSIncludeSubdomains bool          `json:"hsts_include_subdomains" yaml:"hsts_include_subdomains"`
	HSTSPreload           bool          `json:"hsts_preload" yaml:"hsts_preload"`

	// Frame Options
	FrameOptions string `json:"frame_options" yaml:"frame_options"`

	// Content Type Options
	ContentTypeOptions string `json:"content_type_options" yaml:"content_type_options"`

	// XSS Protection
	XSSProtection string `json:"xss_protection" yaml:"xss_protection"`

	// Referrer Policy
	ReferrerPolicy string `json:"referrer_policy" yaml:"referrer_policy"`

	// Permissions Policy
	PermissionsPolicyEnabled bool   `json:"permissions_policy_enabled" yaml:"permissions_policy_enabled"`
	PermissionsPolicy        string `json:"permissions_policy" yaml:"permissions_policy"`

	// Server Information
	ServerName string `json:"server_name" yaml:"server_name"`

	// Additional Headers
	AdditionalHeaders map[string]string `json:"additional_headers" yaml:"additional_headers"`

	// Exclude Paths
	ExcludePaths []string `json:"exclude_paths" yaml:"exclude_paths"`
}

// SecurityHeadersMiddleware provides comprehensive security headers
type SecurityHeadersMiddleware struct {
	config *SecurityHeadersConfig
	logger *zap.Logger
}

// NewSecurityHeadersMiddleware creates a new security headers middleware
func NewSecurityHeadersMiddleware(config *SecurityHeadersConfig, logger *zap.Logger) *SecurityHeadersMiddleware {
	if config == nil {
		config = &SecurityHeadersConfig{
			CSPEnabled:               true,
			CSPDirectives:            "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; font-src 'self' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com data:; script-src 'self' 'unsafe-inline'; img-src 'self' data: https:;",
			HSTSEnabled:              true,
			HSTSMaxAge:               31536000 * time.Second, // 1 year
			HSTSIncludeSubdomains:    true,
			HSTSPreload:              false,
			FrameOptions:             "DENY",
			ContentTypeOptions:       "nosniff",
			XSSProtection:            "1; mode=block",
			ReferrerPolicy:           "strict-origin-when-cross-origin",
			PermissionsPolicyEnabled: true,
			PermissionsPolicy:        "geolocation=(), microphone=(), camera=()",
			ServerName:               "KYB-Tool",
			AdditionalHeaders:        make(map[string]string),
			ExcludePaths:             []string{},
		}
	}

	return &SecurityHeadersMiddleware{
		config: config,
		logger: logger,
	}
}

// Middleware applies security headers to HTTP responses
func (m *SecurityHeadersMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path should be excluded
		if m.shouldExcludePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Apply security headers
		m.applySecurityHeaders(w, r)

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// shouldExcludePath checks if the path should be excluded from security headers
func (m *SecurityHeadersMiddleware) shouldExcludePath(path string) bool {
	for _, excludePath := range m.config.ExcludePaths {
		if strings.HasPrefix(path, excludePath) {
			return true
		}
	}
	return false
}

// applySecurityHeaders applies all configured security headers
func (m *SecurityHeadersMiddleware) applySecurityHeaders(w http.ResponseWriter, r *http.Request) {
	// Content Security Policy
	if m.config.CSPEnabled && m.config.CSPDirectives != "" {
		w.Header().Set("Content-Security-Policy", m.config.CSPDirectives)
	}

	// HTTP Strict Transport Security
	if m.config.HSTSEnabled {
		hstsValue := "max-age=" + string(rune(m.config.HSTSMaxAge.Seconds()))
		if m.config.HSTSIncludeSubdomains {
			hstsValue += "; includeSubDomains"
		}
		if m.config.HSTSPreload {
			hstsValue += "; preload"
		}
		w.Header().Set("Strict-Transport-Security", hstsValue)
	}

	// X-Frame-Options
	if m.config.FrameOptions != "" {
		w.Header().Set("X-Frame-Options", m.config.FrameOptions)
	}

	// X-Content-Type-Options
	if m.config.ContentTypeOptions != "" {
		w.Header().Set("X-Content-Type-Options", m.config.ContentTypeOptions)
	}

	// X-XSS-Protection
	if m.config.XSSProtection != "" {
		w.Header().Set("X-XSS-Protection", m.config.XSSProtection)
	}

	// Referrer-Policy
	if m.config.ReferrerPolicy != "" {
		w.Header().Set("Referrer-Policy", m.config.ReferrerPolicy)
	}

	// Permissions Policy
	if m.config.PermissionsPolicyEnabled && m.config.PermissionsPolicy != "" {
		w.Header().Set("Permissions-Policy", m.config.PermissionsPolicy)
	}

	// Server Information
	if m.config.ServerName != "" {
		w.Header().Set("Server", m.config.ServerName)
	}

	// Additional Headers
	for key, value := range m.config.AdditionalHeaders {
		w.Header().Set(key, value)
	}

	// Log security headers application
	m.logger.Debug("Security headers applied",
		zap.String("path", r.URL.Path),
		zap.String("method", r.Method),
		zap.String("user_agent", r.UserAgent()),
		zap.String("remote_addr", r.RemoteAddr))
}

// UpdateConfig updates the security headers configuration
func (m *SecurityHeadersMiddleware) UpdateConfig(config *SecurityHeadersConfig) {
	if config != nil {
		m.config = config
		m.logger.Info("Security headers configuration updated")
	}
}

// GetConfig returns the current security headers configuration
func (m *SecurityHeadersMiddleware) GetConfig() *SecurityHeadersConfig {
	return m.config
}

// AddExcludePath adds a path to the exclude list
func (m *SecurityHeadersMiddleware) AddExcludePath(path string) {
	m.config.ExcludePaths = append(m.config.ExcludePaths, path)
	m.logger.Info("Security headers exclude path added", zap.String("path", path))
}

// RemoveExcludePath removes a path from the exclude list
func (m *SecurityHeadersMiddleware) RemoveExcludePath(path string) {
	for i, excludePath := range m.config.ExcludePaths {
		if excludePath == path {
			m.config.ExcludePaths = append(m.config.ExcludePaths[:i], m.config.ExcludePaths[i+1:]...)
			m.logger.Info("Security headers exclude path removed", zap.String("path", path))
			break
		}
	}
}

// AddAdditionalHeader adds an additional security header
func (m *SecurityHeadersMiddleware) AddAdditionalHeader(key, value string) {
	if m.config.AdditionalHeaders == nil {
		m.config.AdditionalHeaders = make(map[string]string)
	}
	m.config.AdditionalHeaders[key] = value
	m.logger.Info("Additional security header added", zap.String("key", key), zap.String("value", value))
}

// RemoveAdditionalHeader removes an additional security header
func (m *SecurityHeadersMiddleware) RemoveAdditionalHeader(key string) {
	if m.config.AdditionalHeaders != nil {
		delete(m.config.AdditionalHeaders, key)
		m.logger.Info("Additional security header removed", zap.String("key", key))
	}
}

// Predefined security header configurations
var (
	// StrictSecurityConfig provides maximum security
	StrictSecurityConfig = &SecurityHeadersConfig{
		CSPEnabled:               true,
		CSPDirectives:            "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self'; frame-ancestors 'none';",
		HSTSEnabled:              true,
		HSTSMaxAge:               31536000 * time.Second, // 1 year
		HSTSIncludeSubdomains:    true,
		HSTSPreload:              true,
		FrameOptions:             "DENY",
		ContentTypeOptions:       "nosniff",
		XSSProtection:            "1; mode=block",
		ReferrerPolicy:           "no-referrer",
		PermissionsPolicyEnabled: true,
		PermissionsPolicy:        "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()",
		ServerName:               "KYB-Tool",
		AdditionalHeaders: map[string]string{
			"X-Download-Options":                "noopen",
			"X-Permitted-Cross-Domain-Policies": "none",
		},
	}

	// BalancedSecurityConfig provides balanced security and functionality
	BalancedSecurityConfig = &SecurityHeadersConfig{
		CSPEnabled:               true,
		CSPDirectives:            "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; font-src 'self' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com data:; script-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' https:;",
		HSTSEnabled:              true,
		HSTSMaxAge:               31536000 * time.Second, // 1 year
		HSTSIncludeSubdomains:    true,
		HSTSPreload:              false,
		FrameOptions:             "SAMEORIGIN",
		ContentTypeOptions:       "nosniff",
		XSSProtection:            "1; mode=block",
		ReferrerPolicy:           "strict-origin-when-cross-origin",
		PermissionsPolicyEnabled: true,
		PermissionsPolicy:        "geolocation=(), microphone=(), camera=()",
		ServerName:               "KYB-Tool",
		AdditionalHeaders: map[string]string{
			"X-Download-Options": "noopen",
		},
	}

	// DevelopmentSecurityConfig provides security suitable for development
	DevelopmentSecurityConfig = &SecurityHeadersConfig{
		CSPEnabled:               false, // Disabled for development flexibility
		CSPDirectives:            "",
		HSTSEnabled:              false, // Disabled for development
		HSTSMaxAge:               0,
		HSTSIncludeSubdomains:    false,
		HSTSPreload:              false,
		FrameOptions:             "SAMEORIGIN",
		ContentTypeOptions:       "nosniff",
		XSSProtection:            "1; mode=block",
		ReferrerPolicy:           "no-referrer-when-downgrade",
		PermissionsPolicyEnabled: false,
		PermissionsPolicy:        "",
		ServerName:               "KYB-Tool-Dev",
		AdditionalHeaders:        make(map[string]string),
	}
)

// NewStrictSecurityHeadersMiddleware creates middleware with strict security configuration
func NewStrictSecurityHeadersMiddleware(logger *zap.Logger) *SecurityHeadersMiddleware {
	return NewSecurityHeadersMiddleware(StrictSecurityConfig, logger)
}

// NewBalancedSecurityHeadersMiddleware creates middleware with balanced security configuration
func NewBalancedSecurityHeadersMiddleware(logger *zap.Logger) *SecurityHeadersMiddleware {
	return NewSecurityHeadersMiddleware(BalancedSecurityConfig, logger)
}

// NewDevelopmentSecurityHeadersMiddleware creates middleware with development security configuration
func NewDevelopmentSecurityHeadersMiddleware(logger *zap.Logger) *SecurityHeadersMiddleware {
	return NewSecurityHeadersMiddleware(DevelopmentSecurityConfig, logger)
}
