package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewCORSMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		config   *CORSConfig
		expected *CORSConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			expected: &CORSConfig{
				AllowedOrigins:   []string{"http://localhost:3000", "https://localhost:3000"},
				AllowAllOrigins:  false,
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
				ExposedHeaders:   []string{"X-Total-Count", "X-Page-Count"},
				AllowCredentials: true,
				MaxAge:           86400 * time.Second,
				Debug:            false,
				PathRules:        []CORSPathRule{},
			},
		},
		{
			name: "custom config",
			config: &CORSConfig{
				AllowedOrigins:   []string{"https://example.com"},
				AllowAllOrigins:  false,
				AllowedMethods:   []string{"GET", "POST"},
				AllowCredentials: false,
				MaxAge:           3600 * time.Second,
				Debug:            true,
			},
			expected: &CORSConfig{
				AllowedOrigins:   []string{"https://example.com"},
				AllowAllOrigins:  false,
				AllowedMethods:   []string{"GET", "POST"},
				AllowCredentials: false,
				MaxAge:           3600 * time.Second,
				Debug:            true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewCORSMiddleware(tt.config, logger)

			assert.NotNil(t, middleware)
			assert.Equal(t, tt.expected.AllowedOrigins, middleware.config.AllowedOrigins)
			assert.Equal(t, tt.expected.AllowAllOrigins, middleware.config.AllowAllOrigins)
			assert.Equal(t, tt.expected.AllowedMethods, middleware.config.AllowedMethods)
			assert.Equal(t, tt.expected.AllowCredentials, middleware.config.AllowCredentials)
			assert.Equal(t, tt.expected.MaxAge, middleware.config.MaxAge)
			assert.Equal(t, tt.expected.Debug, middleware.config.Debug)
		})
	}
}

func TestCORSMiddleware_Middleware(t *testing.T) {
	logger := zap.NewNop()
	config := &CORSConfig{
		AllowedOrigins:   []string{"https://example.com", "http://localhost:3000"},
		AllowAllOrigins:  false,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposedHeaders:   []string{"X-Total-Count", "X-RateLimit-Limit"},
		AllowCredentials: true,
		MaxAge:           3600 * time.Second,
		Debug:            false,
	}
	middleware := NewCORSMiddleware(config, logger)

	tests := []struct {
		name            string
		method          string
		origin          string
		expectedStatus  int
		expectedHeaders map[string]string
		shouldHaveCORS  bool
	}{
		{
			name:           "allowed origin preflight",
			method:         "OPTIONS",
			origin:         "https://example.com",
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://example.com",
				"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Origin, Content-Type, Accept, Authorization",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "3600",
			},
			shouldHaveCORS: true,
		},
		{
			name:           "allowed origin actual request",
			method:         "GET",
			origin:         "https://example.com",
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://example.com",
				"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Origin, Content-Type, Accept, Authorization",
				"Access-Control-Expose-Headers":    "X-Total-Count, X-RateLimit-Limit",
				"Access-Control-Allow-Credentials": "true",
			},
			shouldHaveCORS: true,
		},
		{
			name:            "disallowed origin preflight",
			method:          "OPTIONS",
			origin:          "https://malicious.com",
			expectedStatus:  http.StatusForbidden,
			expectedHeaders: map[string]string{},
			shouldHaveCORS:  false,
		},
		{
			name:            "disallowed origin actual request",
			method:          "GET",
			origin:          "https://malicious.com",
			expectedStatus:  http.StatusOK,
			expectedHeaders: map[string]string{},
			shouldHaveCORS:  false,
		},
		{
			name:            "no origin header",
			method:          "GET",
			origin:          "",
			expectedStatus:  http.StatusOK,
			expectedHeaders: map[string]string{},
			shouldHaveCORS:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/test", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.shouldHaveCORS {
				for key, expectedValue := range tt.expectedHeaders {
					actualValue := recorder.Header().Get(key)
					assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", key)
				}
			} else {
				// Should not have CORS headers
				corsHeaders := []string{
					"Access-Control-Allow-Origin",
					"Access-Control-Allow-Methods",
					"Access-Control-Allow-Headers",
					"Access-Control-Expose-Headers",
					"Access-Control-Allow-Credentials",
					"Access-Control-Max-Age",
				}
				for _, header := range corsHeaders {
					value := recorder.Header().Get(header)
					assert.Empty(t, value, "Should not have CORS header %s", header)
				}
			}
		})
	}
}

func TestCORSMiddleware_AllowAllOrigins(t *testing.T) {
	logger := zap.NewNop()
	config := &CORSConfig{
		AllowedOrigins:   []string{},
		AllowAllOrigins:  true,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: false, // Cannot be true with AllowAllOrigins
		MaxAge:           3600 * time.Second,
		Debug:            false,
	}
	middleware := NewCORSMiddleware(config, logger)

	tests := []struct {
		name           string
		origin         string
		expectedOrigin string
	}{
		{
			name:           "any origin allowed",
			origin:         "https://example.com",
			expectedOrigin: "*",
		},
		{
			name:           "another origin allowed",
			origin:         "https://another.com",
			expectedOrigin: "*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("OPTIONS", "/api/test", nil)
			req.Header.Set("Origin", tt.origin)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			assert.Equal(t, http.StatusNoContent, recorder.Code)
			assert.Equal(t, tt.expectedOrigin, recorder.Header().Get("Access-Control-Allow-Origin"))
		})
	}
}

func TestCORSMiddleware_PathRules(t *testing.T) {
	logger := zap.NewNop()
	config := &CORSConfig{
		AllowedOrigins:   []string{"https://example.com"},
		AllowAllOrigins:  false,
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: true,
		MaxAge:           3600 * time.Second,
		Debug:            false,
		PathRules: []CORSPathRule{
			{
				Path:             "/api/public",
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "OPTIONS"},
				AllowCredentials: false,
				MaxAge:           300 * time.Second,
			},
			{
				Path:             "/api/admin",
				AllowedOrigins:   []string{"https://admin.example.com"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
				AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization", "X-Admin-Key"},
				ExposedHeaders:   []string{"X-Admin-Data"},
				AllowCredentials: true,
				MaxAge:           1800 * time.Second,
			},
		},
	}
	middleware := NewCORSMiddleware(config, logger)

	tests := []struct {
		name            string
		path            string
		origin          string
		expectedStatus  int
		expectedHeaders map[string]string
	}{
		{
			name:           "public path with any origin",
			path:           "/api/public/test",
			origin:         "https://any.com",
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "*",
				"Access-Control-Allow-Methods": "GET, OPTIONS",
				"Access-Control-Max-Age":       "300",
			},
		},
		{
			name:           "admin path with admin origin",
			path:           "/api/admin/users",
			origin:         "https://admin.example.com",
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":      "https://admin.example.com",
				"Access-Control-Allow-Methods":     "GET, POST, PUT, DELETE, OPTIONS",
				"Access-Control-Allow-Headers":     "Origin, Content-Type, Authorization, X-Admin-Key",
				"Access-Control-Expose-Headers":    "X-Admin-Data",
				"Access-Control-Allow-Credentials": "true",
				"Access-Control-Max-Age":           "1800",
			},
		},
		{
			name:            "admin path with non-admin origin",
			path:            "/api/admin/users",
			origin:          "https://example.com",
			expectedStatus:  http.StatusForbidden,
			expectedHeaders: map[string]string{},
		},
		{
			name:           "default path with allowed origin",
			path:           "/api/default/test",
			origin:         "https://example.com",
			expectedStatus: http.StatusNoContent,
			expectedHeaders: map[string]string{
				"Access-Control-Allow-Origin":  "https://example.com",
				"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
				"Access-Control-Max-Age":       "3600",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("OPTIONS", tt.path, nil)
			req.Header.Set("Origin", tt.origin)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			for key, expectedValue := range tt.expectedHeaders {
				actualValue := recorder.Header().Get(key)
				assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", key)
			}
		})
	}
}

func TestCORSMiddleware_OriginMatching(t *testing.T) {
	logger := zap.NewNop()
	config := &CORSConfig{
		AllowedOrigins: []string{
			"https://example.com",
			"https://*.example.com",
			"http://localhost:3000",
			"*",
		},
		AllowAllOrigins:  false,
		AllowedMethods:   []string{"GET", "OPTIONS"},
		AllowedHeaders:   []string{"Origin"},
		AllowCredentials: true,
		MaxAge:           3600 * time.Second,
		Debug:            false,
	}
	middleware := NewCORSMiddleware(config, logger)

	tests := []struct {
		name           string
		origin         string
		expectedStatus int
		shouldAllow    bool
	}{
		{
			name:           "exact match",
			origin:         "https://example.com",
			expectedStatus: http.StatusNoContent,
			shouldAllow:    true,
		},
		{
			name:           "subdomain wildcard match",
			origin:         "https://api.example.com",
			expectedStatus: http.StatusNoContent,
			shouldAllow:    true,
		},
		{
			name:           "another subdomain wildcard match",
			origin:         "https://admin.example.com",
			expectedStatus: http.StatusNoContent,
			shouldAllow:    true,
		},
		{
			name:           "localhost match",
			origin:         "http://localhost:3000",
			expectedStatus: http.StatusNoContent,
			shouldAllow:    true,
		},
		{
			name:           "wildcard match",
			origin:         "https://any-domain.com",
			expectedStatus: http.StatusNoContent,
			shouldAllow:    true,
		},
		{
			name:           "no match",
			origin:         "https://malicious.com",
			expectedStatus: http.StatusForbidden,
			shouldAllow:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("OPTIONS", "/api/test", nil)
			req.Header.Set("Origin", tt.origin)

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.shouldAllow {
				assert.NotEmpty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
			} else {
				assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
			}
		})
	}
}

func TestCORSMiddleware_MatchesOrigin(t *testing.T) {
	logger := zap.NewNop()
	middleware := NewCORSMiddleware(nil, logger)

	tests := []struct {
		name    string
		origin  string
		pattern string
		matches bool
	}{
		{
			name:    "exact match",
			origin:  "https://example.com",
			pattern: "https://example.com",
			matches: true,
		},
		{
			name:    "wildcard subdomain match",
			origin:  "https://api.example.com",
			pattern: "https://*.example.com",
			matches: true,
		},
		{
			name:    "wildcard subdomain no match",
			origin:  "https://api.sub.example.com",
			pattern: "https://*.example.com",
			matches: false,
		},
		{
			name:    "wildcard match",
			origin:  "https://any-domain.com",
			pattern: "*",
			matches: true,
		},
		{
			name:    "no match",
			origin:  "https://example.com",
			pattern: "https://other.com",
			matches: false,
		},
		{
			name:    "empty origin",
			origin:  "",
			pattern: "*",
			matches: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.matchesOrigin(tt.origin, tt.pattern)
			assert.Equal(t, tt.matches, result)
		})
	}
}

func TestGetDefaultCORSConfig(t *testing.T) {
	config := GetDefaultCORSConfig()

	assert.NotNil(t, config)
	assert.False(t, config.AllowAllOrigins)
	assert.True(t, config.AllowCredentials)
	assert.Equal(t, 86400*time.Second, config.MaxAge)
	assert.False(t, config.Debug)

	// Check default origins
	expectedOrigins := []string{
		"http://localhost:3000",
		"https://localhost:3000",
		"http://localhost:8080",
		"https://localhost:8080",
	}
	assert.Equal(t, expectedOrigins, config.AllowedOrigins)

	// Check default methods
	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"}
	assert.Equal(t, expectedMethods, config.AllowedMethods)
}

func TestGetStrictCORSConfig(t *testing.T) {
	config := GetStrictCORSConfig()

	assert.NotNil(t, config)
	assert.False(t, config.AllowAllOrigins)
	assert.True(t, config.AllowCredentials)
	assert.Equal(t, 3600*time.Second, config.MaxAge)
	assert.False(t, config.Debug)
	assert.Empty(t, config.AllowedOrigins) // Must be explicitly set

	// Check strict methods
	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	assert.Equal(t, expectedMethods, config.AllowedMethods)
}

func TestGetDevelopmentCORSConfig(t *testing.T) {
	config := GetDevelopmentCORSConfig()

	assert.NotNil(t, config)
	assert.True(t, config.AllowAllOrigins)
	assert.True(t, config.AllowCredentials)
	assert.Equal(t, 300*time.Second, config.MaxAge)
	assert.True(t, config.Debug)

	// Check development origins
	assert.Equal(t, []string{"*"}, config.AllowedOrigins)

	// Check development methods
	expectedMethods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"}
	assert.Equal(t, expectedMethods, config.AllowedMethods)

	// Check wildcard headers
	assert.Equal(t, []string{"*"}, config.AllowedHeaders)
	assert.Equal(t, []string{"*"}, config.ExposedHeaders)
}
