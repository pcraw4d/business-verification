package security

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecurityHeadersManager(t *testing.T) {
	tests := []struct {
		name   string
		config *SecurityHeadersConfig
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &SecurityHeadersConfig{
				CORSEnabled:        true,
				AllowedOrigins:     []string{"https://example.com"},
				AllowedMethods:     []string{"GET", "POST"},
				AllowedHeaders:     []string{"Content-Type"},
				ContentTypeOptions: "nosniff",
				FrameOptions:       "DENY",
				XSSProtection:      "1; mode=block",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shm := NewSecurityHeadersManager(tt.config)
			assert.NotNil(t, shm)
			assert.NotNil(t, shm.config)
		})
	}
}

func TestSecurityHeadersManager_SecurityHeadersMiddleware(t *testing.T) {
	shm := NewSecurityHeadersManager(nil)
	middleware := shm.SecurityHeadersMiddleware()

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Wrap with middleware
	wrappedHandler := middleware(handler)

	tests := []struct {
		name           string
		method         string
		origin         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "GET request",
			method:         "GET",
			origin:         "",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "OPTIONS request (CORS preflight)",
			method:         "OPTIONS",
			origin:         "https://example.com",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "POST request with origin",
			method:         "POST",
			origin:         "https://example.com",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkHeaders {
				// Check security headers
				assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
				assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
				assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
				assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
				assert.Equal(t, "geolocation=(), microphone=(), camera=()", w.Header().Get("Permissions-Policy"))
				assert.Equal(t, "RiskAssessment/1.0", w.Header().Get("Server"))
				assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
				assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
				assert.Equal(t, "0", w.Header().Get("Expires"))

				// Check that dangerous headers are removed
				assert.Empty(t, w.Header().Get("X-Powered-By"))
				assert.Empty(t, w.Header().Get("X-AspNet-Version"))
				assert.Empty(t, w.Header().Get("X-AspNetMvc-Version"))
			}
		})
	}
}

func TestSecurityHeadersManager_SetCORSHeaders(t *testing.T) {
	tests := []struct {
		name           string
		config         *SecurityHeadersConfig
		origin         string
		expectedOrigin string
		expectError    bool
	}{
		{
			name: "wildcard origin allowed",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"*"},
			},
			origin:         "https://example.com",
			expectedOrigin: "https://example.com",
			expectError:    false,
		},
		{
			name: "specific origin allowed",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"https://example.com", "https://test.com"},
			},
			origin:         "https://example.com",
			expectedOrigin: "https://example.com",
			expectError:    false,
		},
		{
			name: "origin not allowed",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"https://example.com"},
			},
			origin:         "https://malicious.com",
			expectedOrigin: "",
			expectError:    false,
		},
		{
			name: "CORS disabled",
			config: &SecurityHeadersConfig{
				CORSEnabled: false,
			},
			origin:         "https://example.com",
			expectedOrigin: "",
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shm := NewSecurityHeadersManager(tt.config)
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			w := httptest.NewRecorder()
			shm.setCORSHeaders(w, req)

			if tt.expectedOrigin != "" {
				assert.Equal(t, tt.expectedOrigin, w.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestSecurityHeadersManager_HandleCORSPreflight(t *testing.T) {
	tests := []struct {
		name           string
		config         *SecurityHeadersConfig
		origin         string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "allowed origin",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"https://example.com"},
			},
			origin:         "https://example.com",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name: "disallowed origin",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"https://example.com"},
			},
			origin:         "https://malicious.com",
			expectedStatus: http.StatusForbidden,
			expectError:    true,
		},
		{
			name: "wildcard origin",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"*"},
			},
			origin:         "https://any-origin.com",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shm := NewSecurityHeadersManager(tt.config)
			req := httptest.NewRequest("OPTIONS", "/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			w := httptest.NewRecorder()
			shm.handleCORSPreflight(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectError {
				// Check CORS headers are set
				assert.Equal(t, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
				assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Methods"))
				assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Headers"))
			}
		})
	}
}

func TestSecurityHeadersManager_IsOriginAllowed(t *testing.T) {
	tests := []struct {
		name     string
		config   *SecurityHeadersConfig
		origin   string
		expected bool
	}{
		{
			name: "wildcard origin",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"*"},
			},
			origin:   "https://example.com",
			expected: true,
		},
		{
			name: "specific origin allowed",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"https://example.com", "https://test.com"},
			},
			origin:   "https://example.com",
			expected: true,
		},
		{
			name: "specific origin not allowed",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"https://example.com"},
			},
			origin:   "https://malicious.com",
			expected: false,
		},
		{
			name: "CORS disabled",
			config: &SecurityHeadersConfig{
				CORSEnabled: false,
			},
			origin:   "https://example.com",
			expected: false,
		},
		{
			name: "empty origin",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"https://example.com"},
			},
			origin:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shm := NewSecurityHeadersManager(tt.config)
			result := shm.isOriginAllowed(tt.origin)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSecurityHeadersManager_UpdateConfig(t *testing.T) {
	shm := NewSecurityHeadersManager(nil)
	originalConfig := shm.GetConfig()

	newConfig := &SecurityHeadersConfig{
		CORSEnabled:    false,
		AllowedOrigins: []string{"https://new-origin.com"},
		FrameOptions:   "SAMEORIGIN",
	}

	shm.UpdateConfig(newConfig)
	updatedConfig := shm.GetConfig()

	assert.Equal(t, newConfig.CORSEnabled, updatedConfig.CORSEnabled)
	assert.Equal(t, newConfig.AllowedOrigins, updatedConfig.AllowedOrigins)
	assert.Equal(t, newConfig.FrameOptions, updatedConfig.FrameOptions)
	assert.NotEqual(t, originalConfig.CORSEnabled, updatedConfig.CORSEnabled)
}

func TestSecurityHeadersManager_ValidateConfig(t *testing.T) {
	tests := []struct {
		name          string
		config        *SecurityHeadersConfig
		expectedError bool
		errorContains string
	}{
		{
			name: "valid config",
			config: &SecurityHeadersConfig{
				CORSEnabled:        true,
				AllowedOrigins:     []string{"https://example.com"},
				AllowedMethods:     []string{"GET", "POST"},
				ContentTypeOptions: "nosniff",
				FrameOptions:       "DENY",
			},
			expectedError: false,
		},
		{
			name: "CORS enabled but no origins",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{},
				AllowedMethods: []string{"GET", "POST"},
			},
			expectedError: true,
			errorContains: "no allowed origins specified",
		},
		{
			name: "CORS enabled but no methods",
			config: &SecurityHeadersConfig{
				CORSEnabled:    true,
				AllowedOrigins: []string{"https://example.com"},
				AllowedMethods: []string{},
			},
			expectedError: true,
			errorContains: "no allowed methods specified",
		},
		{
			name: "credentials with wildcard origin",
			config: &SecurityHeadersConfig{
				CORSEnabled:      true,
				AllowedOrigins:   []string{"*"},
				AllowCredentials: true,
			},
			expectedError: true,
			errorContains: "credentials cannot be allowed with wildcard origin",
		},
		{
			name: "invalid content type options",
			config: &SecurityHeadersConfig{
				ContentTypeOptions: "invalid",
			},
			expectedError: true,
			errorContains: "should be 'nosniff'",
		},
		{
			name: "invalid frame options",
			config: &SecurityHeadersConfig{
				FrameOptions: "invalid",
			},
			expectedError: true,
			errorContains: "should be 'DENY' or 'SAMEORIGIN'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shm := NewSecurityHeadersManager(tt.config)
			errors := shm.ValidateConfig()

			if tt.expectedError {
				assert.NotEmpty(t, errors)
				if tt.errorContains != "" {
					found := false
					for _, err := range errors {
						if err == tt.errorContains {
							found = true
							break
						}
					}
					assert.True(t, found, "Expected error message not found: %s", tt.errorContains)
				}
			} else {
				assert.Empty(t, errors)
			}
		})
	}
}

func TestCreateStrictConfig(t *testing.T) {
	config := CreateStrictConfig()

	assert.False(t, config.CORSEnabled)
	assert.Empty(t, config.AllowedOrigins)
	assert.Equal(t, []string{"GET", "POST"}, config.AllowedMethods)
	assert.Equal(t, []string{"Content-Type", "Authorization"}, config.AllowedHeaders)
	assert.Empty(t, config.ExposedHeaders)
	assert.False(t, config.AllowCredentials)
	assert.Equal(t, 0, config.MaxAge)
	assert.Equal(t, "nosniff", config.ContentTypeOptions)
	assert.Equal(t, "DENY", config.FrameOptions)
	assert.Equal(t, "1; mode=block", config.XSSProtection)
	assert.Equal(t, "no-referrer", config.ReferrerPolicy)
	assert.Equal(t, "geolocation=(), microphone=(), camera=(), payment=(), usb=()", config.PermissionsPolicy)
	assert.Equal(t, "max-age=31536000; includeSubDomains; preload", config.StrictTransportSecurity)
	assert.Equal(t, "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'", config.ContentSecurityPolicy)
	assert.Empty(t, config.ServerHeader)
	assert.Equal(t, "no-cache, no-store, must-revalidate, private", config.CacheControl)
	assert.Equal(t, "no-cache", config.Pragma)
	assert.Equal(t, "0", config.Expires)
}

func TestCreatePermissiveConfig(t *testing.T) {
	config := CreatePermissiveConfig()

	assert.True(t, config.CORSEnabled)
	assert.Equal(t, []string{"*"}, config.AllowedOrigins)
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}, config.AllowedMethods)
	assert.Equal(t, []string{"*"}, config.AllowedHeaders)
	assert.Equal(t, []string{"*"}, config.ExposedHeaders)
	assert.False(t, config.AllowCredentials)
	assert.Equal(t, 86400, config.MaxAge)
	assert.Equal(t, "nosniff", config.ContentTypeOptions)
	assert.Equal(t, "SAMEORIGIN", config.FrameOptions)
	assert.Equal(t, "1; mode=block", config.XSSProtection)
	assert.Equal(t, "strict-origin-when-cross-origin", config.ReferrerPolicy)
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", config.PermissionsPolicy)
	assert.Equal(t, "max-age=31536000", config.StrictTransportSecurity)
	assert.Equal(t, "default-src 'self' 'unsafe-inline' 'unsafe-eval'", config.ContentSecurityPolicy)
	assert.Equal(t, "RiskAssessment/1.0", config.ServerHeader)
	assert.Equal(t, "public, max-age=3600", config.CacheControl)
	assert.Empty(t, config.Pragma)
	assert.NotEmpty(t, config.Expires)
}

func TestCreateDevelopmentConfig(t *testing.T) {
	config := CreateDevelopmentConfig()

	assert.True(t, config.CORSEnabled)
	assert.Contains(t, config.AllowedOrigins, "http://localhost:3000")
	assert.Contains(t, config.AllowedOrigins, "http://localhost:8080")
	assert.Contains(t, config.AllowedOrigins, "http://127.0.0.1:3000")
	assert.Equal(t, []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}, config.AllowedMethods)
	assert.Contains(t, config.AllowedHeaders, "Content-Type")
	assert.Contains(t, config.AllowedHeaders, "Authorization")
	assert.Contains(t, config.AllowedHeaders, "X-Requested-With")
	assert.Contains(t, config.ExposedHeaders, "X-Total-Count")
	assert.Contains(t, config.ExposedHeaders, "X-Page-Count")
	assert.Contains(t, config.ExposedHeaders, "X-Request-ID")
	assert.True(t, config.AllowCredentials)
	assert.Equal(t, 3600, config.MaxAge)
	assert.Equal(t, "nosniff", config.ContentTypeOptions)
	assert.Equal(t, "SAMEORIGIN", config.FrameOptions)
	assert.Equal(t, "1; mode=block", config.XSSProtection)
	assert.Equal(t, "strict-origin-when-cross-origin", config.ReferrerPolicy)
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", config.PermissionsPolicy)
	assert.Empty(t, config.StrictTransportSecurity)
	assert.Equal(t, "default-src 'self' 'unsafe-inline' 'unsafe-eval' data: blob:", config.ContentSecurityPolicy)
	assert.Equal(t, "RiskAssessment-Dev/1.0", config.ServerHeader)
	assert.Equal(t, "no-cache", config.CacheControl)
	assert.Equal(t, "no-cache", config.Pragma)
	assert.Equal(t, "0", config.Expires)
}

func TestSecurityHeadersManager_HTTPSHeaders(t *testing.T) {
	// Create a test server with TLS
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shm := NewSecurityHeadersManager(nil)
		shm.setSecurityHeaders(w, r)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Make a request to the HTTPS server
	client := server.Client()
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Check that HSTS header is set for HTTPS
	assert.Equal(t, "max-age=31536000; includeSubDomains", resp.Header.Get("Strict-Transport-Security"))
}

func TestSecurityHeadersManager_ConcurrentAccess(t *testing.T) {
	shm := NewSecurityHeadersManager(nil)
	middleware := shm.SecurityHeadersMiddleware()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrappedHandler := middleware(handler)

	// Test concurrent requests
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer func() { done <- true }()

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
