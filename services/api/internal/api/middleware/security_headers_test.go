package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewSecurityHeadersMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		config   *SecurityHeadersConfig
		expected *SecurityHeadersConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			expected: &SecurityHeadersConfig{
				CSPEnabled:               true,
				CSPDirectives:            "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; font-src 'self' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com data:; script-src 'self' 'unsafe-inline'; img-src 'self' data: https:;",
				HSTSEnabled:              true,
				HSTSMaxAge:               31536000 * time.Second,
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
			},
		},
		{
			name: "custom config",
			config: &SecurityHeadersConfig{
				CSPEnabled:   false,
				FrameOptions: "SAMEORIGIN",
				ServerName:   "Custom-Server",
				ExcludePaths: []string{"/health"},
			},
			expected: &SecurityHeadersConfig{
				CSPEnabled:   false,
				FrameOptions: "SAMEORIGIN",
				ServerName:   "Custom-Server",
				ExcludePaths: []string{"/health"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewSecurityHeadersMiddleware(tt.config, logger)

			assert.NotNil(t, middleware)
			assert.Equal(t, logger, middleware.logger)

			if tt.config == nil {
				// Check default values
				config := middleware.GetConfig()
				assert.Equal(t, tt.expected.CSPEnabled, config.CSPEnabled)
				assert.Equal(t, tt.expected.FrameOptions, config.FrameOptions)
				assert.Equal(t, tt.expected.ServerName, config.ServerName)
			} else {
				// Check custom values
				config := middleware.GetConfig()
				assert.Equal(t, tt.expected.CSPEnabled, config.CSPEnabled)
				assert.Equal(t, tt.expected.FrameOptions, config.FrameOptions)
				assert.Equal(t, tt.expected.ServerName, config.ServerName)
				assert.Equal(t, tt.expected.ExcludePaths, config.ExcludePaths)
			}
		})
	}
}

func TestSecurityHeadersMiddleware_Middleware(t *testing.T) {
	tests := []struct {
		name            string
		config          *SecurityHeadersConfig
		path            string
		expectedHeaders map[string]string
		shouldExclude   bool
	}{
		{
			name: "default headers applied",
			config: &SecurityHeadersConfig{
				CSPEnabled:               true,
				CSPDirectives:            "default-src 'self'",
				HSTSEnabled:              true,
				HSTSMaxAge:               31536000 * time.Second,
				HSTSIncludeSubdomains:    true,
				FrameOptions:             "DENY",
				ContentTypeOptions:       "nosniff",
				XSSProtection:            "1; mode=block",
				ReferrerPolicy:           "strict-origin-when-cross-origin",
				PermissionsPolicyEnabled: true,
				PermissionsPolicy:        "geolocation=()",
				ServerName:               "KYB-Tool",
				AdditionalHeaders: map[string]string{
					"X-Custom-Header": "test-value",
				},
			},
			path: "/api/test",
			expectedHeaders: map[string]string{
				"Content-Security-Policy":   "default-src 'self'",
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains",
				"X-Frame-Options":           "DENY",
				"X-Content-Type-Options":    "nosniff",
				"X-XSS-Protection":          "1; mode=block",
				"Referrer-Policy":           "strict-origin-when-cross-origin",
				"Permissions-Policy":        "geolocation=()",
				"Server":                    "KYB-Tool",
				"X-Custom-Header":           "test-value",
			},
			shouldExclude: false,
		},
		{
			name: "excluded path",
			config: &SecurityHeadersConfig{
				ExcludePaths: []string{"/health"},
			},
			path:            "/health",
			expectedHeaders: map[string]string{},
			shouldExclude:   true,
		},
		{
			name: "excluded path prefix",
			config: &SecurityHeadersConfig{
				ExcludePaths: []string{"/api/"},
			},
			path:            "/api/v1/users",
			expectedHeaders: map[string]string{},
			shouldExclude:   true,
		},
		{
			name: "HSTS with preload",
			config: &SecurityHeadersConfig{
				HSTSEnabled:           true,
				HSTSMaxAge:            31536000 * time.Second,
				HSTSIncludeSubdomains: true,
				HSTSPreload:           true,
			},
			path: "/test",
			expectedHeaders: map[string]string{
				"Strict-Transport-Security": "max-age=31536000; includeSubDomains; preload",
			},
			shouldExclude: false,
		},
		{
			name: "disabled CSP",
			config: &SecurityHeadersConfig{
				CSPEnabled:    false,
				CSPDirectives: "default-src 'self'",
			},
			path:            "/test",
			expectedHeaders: map[string]string{},
			shouldExclude:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewSecurityHeadersMiddleware(tt.config, logger)

			req := httptest.NewRequest("GET", tt.path, nil)
			rec := httptest.NewRecorder()

			// Create a simple handler
			handlerCalled := false
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			// Apply middleware
			middleware.Middleware(handler).ServeHTTP(rec, req)

			// Check if handler was called
			assert.True(t, handlerCalled)

			// Check headers
			for key, expectedValue := range tt.expectedHeaders {
				actualValue := rec.Header().Get(key)
				assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", key)
			}

			// Check that excluded paths don't have security headers
			if tt.shouldExclude {
				for key := range tt.expectedHeaders {
					actualValue := rec.Header().Get(key)
					assert.Empty(t, actualValue, "Header %s should not be set for excluded path", key)
				}
			}
		})
	}
}

func TestSecurityHeadersMiddleware_shouldExcludePath(t *testing.T) {
	tests := []struct {
		name         string
		excludePaths []string
		path         string
		expected     bool
	}{
		{
			name:         "no exclude paths",
			excludePaths: []string{},
			path:         "/api/test",
			expected:     false,
		},
		{
			name:         "exact match",
			excludePaths: []string{"/health"},
			path:         "/health",
			expected:     true,
		},
		{
			name:         "prefix match",
			excludePaths: []string{"/api/"},
			path:         "/api/v1/users",
			expected:     true,
		},
		{
			name:         "no match",
			excludePaths: []string{"/api/"},
			path:         "/health",
			expected:     false,
		},
		{
			name:         "multiple exclude paths",
			excludePaths: []string{"/health", "/metrics"},
			path:         "/metrics",
			expected:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			config := &SecurityHeadersConfig{
				ExcludePaths: tt.excludePaths,
			}
			middleware := NewSecurityHeadersMiddleware(config, logger)

			result := middleware.shouldExcludePath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityHeadersMiddleware_ConfigurationMethods(t *testing.T) {
	logger := zap.NewNop()
	middleware := NewSecurityHeadersMiddleware(nil, logger)

	t.Run("UpdateConfig", func(t *testing.T) {
		newConfig := &SecurityHeadersConfig{
			ServerName:   "Updated-Server",
			FrameOptions: "SAMEORIGIN",
		}

		middleware.UpdateConfig(newConfig)
		config := middleware.GetConfig()

		assert.Equal(t, "Updated-Server", config.ServerName)
		assert.Equal(t, "SAMEORIGIN", config.FrameOptions)
	})

	t.Run("AddExcludePath", func(t *testing.T) {
		middleware.AddExcludePath("/new-exclude")

		config := middleware.GetConfig()
		assert.Contains(t, config.ExcludePaths, "/new-exclude")
	})

	t.Run("RemoveExcludePath", func(t *testing.T) {
		// Add a path first
		middleware.AddExcludePath("/to-remove")

		// Remove it
		middleware.RemoveExcludePath("/to-remove")

		config := middleware.GetConfig()
		assert.NotContains(t, config.ExcludePaths, "/to-remove")
	})

	t.Run("AddAdditionalHeader", func(t *testing.T) {
		middleware.AddAdditionalHeader("X-Test-Header", "test-value")

		config := middleware.GetConfig()
		assert.Equal(t, "test-value", config.AdditionalHeaders["X-Test-Header"])
	})

	t.Run("RemoveAdditionalHeader", func(t *testing.T) {
		// Add a header first
		middleware.AddAdditionalHeader("X-To-Remove", "value")

		// Remove it
		middleware.RemoveAdditionalHeader("X-To-Remove")

		config := middleware.GetConfig()
		_, exists := config.AdditionalHeaders["X-To-Remove"]
		assert.False(t, exists)
	})
}

func TestPredefinedSecurityConfigs(t *testing.T) {
	logger := zap.NewNop()

	t.Run("StrictSecurityConfig", func(t *testing.T) {
		middleware := NewStrictSecurityHeadersMiddleware(logger)
		config := middleware.GetConfig()

		assert.True(t, config.CSPEnabled)
		assert.True(t, config.HSTSEnabled)
		assert.True(t, config.HSTSPreload)
		assert.Equal(t, "DENY", config.FrameOptions)
		assert.Equal(t, "no-referrer", config.ReferrerPolicy)
		assert.Contains(t, config.AdditionalHeaders, "X-Download-Options")
		assert.Contains(t, config.AdditionalHeaders, "X-Permitted-Cross-Domain-Policies")
	})

	t.Run("BalancedSecurityConfig", func(t *testing.T) {
		middleware := NewBalancedSecurityHeadersMiddleware(logger)
		config := middleware.GetConfig()

		assert.True(t, config.CSPEnabled)
		assert.True(t, config.HSTSEnabled)
		assert.False(t, config.HSTSPreload)
		assert.Equal(t, "SAMEORIGIN", config.FrameOptions)
		assert.Equal(t, "strict-origin-when-cross-origin", config.ReferrerPolicy)
		assert.Contains(t, config.AdditionalHeaders, "X-Download-Options")
	})

	t.Run("DevelopmentSecurityConfig", func(t *testing.T) {
		middleware := NewDevelopmentSecurityHeadersMiddleware(logger)
		config := middleware.GetConfig()

		assert.False(t, config.CSPEnabled)
		assert.False(t, config.HSTSEnabled)
		assert.Equal(t, "SAMEORIGIN", config.FrameOptions)
		assert.Equal(t, "no-referrer-when-downgrade", config.ReferrerPolicy)
		assert.Equal(t, "KYB-Tool-Dev", config.ServerName)
	})
}

func TestSecurityHeadersMiddleware_Integration(t *testing.T) {
	logger := zap.NewNop()

	// Test with strict security config
	middleware := NewStrictSecurityHeadersMiddleware(logger)

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	middleware.Middleware(handler).ServeHTTP(rec, req)

	// Verify strict security headers are applied
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "test response", rec.Body.String())

	// Check for strict security headers
	headers := rec.Header()
	assert.NotEmpty(t, headers.Get("Content-Security-Policy"))
	assert.NotEmpty(t, headers.Get("Strict-Transport-Security"))
	assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
	assert.Equal(t, "no-referrer", headers.Get("Referrer-Policy"))
	assert.NotEmpty(t, headers.Get("Permissions-Policy"))
	assert.Equal(t, "KYB-Tool", headers.Get("Server"))
	assert.Equal(t, "noopen", headers.Get("X-Download-Options"))
	assert.Equal(t, "none", headers.Get("X-Permitted-Cross-Domain-Policies"))
}

func TestSecurityHeadersMiddleware_Performance(t *testing.T) {
	logger := zap.NewNop()
	middleware := NewSecurityHeadersMiddleware(nil, logger)

	req := httptest.NewRequest("GET", "/api/test", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Benchmark the middleware
	b := testing.B{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		middleware.Middleware(handler).ServeHTTP(rec, req)
	}
}

func TestSecurityHeadersMiddleware_EdgeCases(t *testing.T) {
	logger := zap.NewNop()

	t.Run("empty path", func(t *testing.T) {
		middleware := NewSecurityHeadersMiddleware(nil, logger)
		result := middleware.shouldExcludePath("")
		assert.False(t, result)
	})

	t.Run("nil additional headers", func(t *testing.T) {
		config := &SecurityHeadersConfig{
			AdditionalHeaders: nil,
		}
		middleware := NewSecurityHeadersMiddleware(config, logger)

		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware.Middleware(handler).ServeHTTP(rec, req)

		// Should not panic and should still apply other headers
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.NotEmpty(t, rec.Header().Get("X-Content-Type-Options"))
	})

	t.Run("HSTS disabled", func(t *testing.T) {
		config := &SecurityHeadersConfig{
			HSTSEnabled: false,
		}
		middleware := NewSecurityHeadersMiddleware(config, logger)

		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware.Middleware(handler).ServeHTTP(rec, req)

		// HSTS header should not be set
		assert.Empty(t, rec.Header().Get("Strict-Transport-Security"))
	})

	t.Run("CSP disabled", func(t *testing.T) {
		config := &SecurityHeadersConfig{
			CSPEnabled:    false,
			CSPDirectives: "default-src 'self'",
		}
		middleware := NewSecurityHeadersMiddleware(config, logger)

		req := httptest.NewRequest("GET", "/test", nil)
		rec := httptest.NewRecorder()

		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware.Middleware(handler).ServeHTTP(rec, req)

		// CSP header should not be set
		assert.Empty(t, rec.Header().Get("Content-Security-Policy"))
	})
}
