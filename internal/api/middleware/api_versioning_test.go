package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"kyb-platform/internal/api/compatibility"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewAPIVersioningMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		config         *APIVersioningConfig
		versionManager *compatibility.VersionManager
		logger         *zap.Logger
		expectPanic    bool
	}{
		{
			name:           "valid configuration",
			config:         GetDefaultAPIVersioningConfig(),
			versionManager: compatibility.NewVersionManager(zap.NewNop(), nil),
			logger:         zap.NewNop(),
			expectPanic:    false,
		},
		{
			name:           "nil config uses default",
			config:         nil,
			versionManager: compatibility.NewVersionManager(zap.NewNop(), nil),
			logger:         zap.NewNop(),
			expectPanic:    false,
		},
		{
			name:           "nil version manager panics",
			config:         GetDefaultAPIVersioningConfig(),
			versionManager: nil,
			logger:         zap.NewNop(),
			expectPanic:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				assert.Panics(t, func() {
					NewAPIVersioningMiddleware(tt.config, tt.versionManager, tt.logger)
				})
			} else {
				middleware := NewAPIVersioningMiddleware(tt.config, tt.versionManager, tt.logger)
				assert.NotNil(t, middleware)
				assert.NotNil(t, middleware.config)
				assert.NotNil(t, middleware.versionManager)
				assert.NotNil(t, middleware.logger)
				assert.NotNil(t, middleware.versionRegex)
			}
		})
	}
}

func TestAPIVersioningMiddleware_Middleware(t *testing.T) {
	versionManager := compatibility.NewVersionManager(zap.NewNop(), nil)

	tests := []struct {
		name            string
		config          *APIVersioningConfig
		requestPath     string
		requestHeaders  map[string]string
		expectedStatus  int
		expectedHeaders map[string]string
		checkContext    bool
	}{
		{
			name:           "valid URL versioning",
			config:         GetDefaultAPIVersioningConfig(),
			requestPath:    "/v3/test",
			requestHeaders: map[string]string{},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"X-API-Version":           "v3",
				"X-API-Version-Requested": "v3",
			},
			checkContext: true,
		},
		{
			name:           "valid header versioning",
			config:         GetDefaultAPIVersioningConfig(),
			requestPath:    "/test",
			requestHeaders: map[string]string{"X-API-Version": "v2"},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"X-API-Version":           "v2",
				"X-API-Version-Requested": "v2",
			},
			checkContext: true,
		},
		{
			name:           "valid query versioning",
			config:         GetDefaultAPIVersioningConfig(),
			requestPath:    "/test?version=v2",
			requestHeaders: map[string]string{},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"X-API-Version":           "v2",
				"X-API-Version-Requested": "v2",
			},
			checkContext: true,
		},
		{
			name:           "valid accept header versioning",
			config:         GetDefaultAPIVersioningConfig(),
			requestPath:    "/test",
			requestHeaders: map[string]string{"Accept": "application/vnd.kyb-platform.v2+json"},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"X-API-Version":           "v2",
				"X-API-Version-Requested": "v2",
			},
			checkContext: true,
		},
		{
			name:           "unsupported version with error response",
			config:         GetDefaultAPIVersioningConfig(),
			requestPath:    "/v99/test",
			requestHeaders: map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectedHeaders: map[string]string{
				"Content-Type": "application/json",
			},
			checkContext: false,
		},
		{
			name: "unsupported version with 404",
			config: func() *APIVersioningConfig {
				config := GetDefaultAPIVersioningConfig()
				config.ReturnVersionErrors = false
				return config
			}(),
			requestPath:     "/v99/test",
			requestHeaders:  map[string]string{},
			expectedStatus:  http.StatusNotFound,
			expectedHeaders: map[string]string{},
			checkContext:    false,
		},
		{
			name:           "deprecated version with warnings",
			config:         GetDefaultAPIVersioningConfig(),
			requestPath:    "/v1/test",
			requestHeaders: map[string]string{},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"X-API-Version":           "v1",
				"X-API-Version-Requested": "v1",
				"X-API-Deprecated":        "true",
			},
			checkContext: true,
		},
		{
			name:           "client version validation",
			config:         GetDefaultAPIVersioningConfig(),
			requestPath:    "/v3/test",
			requestHeaders: map[string]string{"X-Client-Version": "3.0.0"},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"X-API-Version":    "v3",
				"X-Client-Version": "3.0.0",
			},
			checkContext: true,
		},
		{
			name:           "invalid client version",
			config:         GetDefaultAPIVersioningConfig(),
			requestPath:    "/v3/test",
			requestHeaders: map[string]string{"X-Client-Version": "1.0.0"},
			expectedStatus: http.StatusOK,
			expectedHeaders: map[string]string{
				"X-API-Version":            "v3",
				"X-Client-Version":         "1.0.0",
				"X-Client-Version-Warning": "Client version may not be fully compatible",
			},
			checkContext: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create middleware
			middleware := NewAPIVersioningMiddleware(tt.config, versionManager, zap.NewNop())

			// Create test handler
			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true

				// Check context if needed
				if tt.checkContext {
					versionInfo := GetVersionInfo(r.Context())
					assert.NotNil(t, versionInfo)
					assert.NotEmpty(t, versionInfo.ResolvedVersion)

					apiVersion := GetAPIVersion(r.Context())
					assert.NotEmpty(t, apiVersion)
					assert.Equal(t, versionInfo.ResolvedVersion, apiVersion)
				}

				w.WriteHeader(http.StatusOK)
				w.Write([]byte("success"))
			})

			// Create request
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			for key, value := range tt.requestHeaders {
				req.Header.Set(key, value)
			}

			// Create response recorder
			w := httptest.NewRecorder()

			// Apply middleware
			middleware.Middleware(testHandler).ServeHTTP(w, req)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			for key, expectedValue := range tt.expectedHeaders {
				actualValue := w.Header().Get(key)
				assert.Equal(t, expectedValue, actualValue, "Header %s mismatch", key)
			}

			if tt.expectedStatus == http.StatusOK {
				assert.True(t, handlerCalled, "Handler should have been called")
			} else {
				assert.False(t, handlerCalled, "Handler should not have been called")
			}
		})
	}
}

func TestAPIVersioningMiddleware_ExtractVersionFromURL(t *testing.T) {
	versionManager := compatibility.NewVersionManager(zap.NewNop(), nil)
	middleware := NewAPIVersioningMiddleware(GetDefaultAPIVersioningConfig(), versionManager, zap.NewNop())

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"valid version v1", "/v1/test", "v1"},
		{"valid version v2", "/v2/api/users", "v2"},
		{"valid version v3", "/v3/", "v3"},
		{"no version", "/test", ""},
		{"invalid version format", "/version1/test", ""},
		{"empty path", "", ""},
		{"root path", "/", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.extractVersionFromURL(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAPIVersioningMiddleware_ExtractVersionFromAccept(t *testing.T) {
	versionManager := compatibility.NewVersionManager(zap.NewNop(), nil)
	middleware := NewAPIVersioningMiddleware(GetDefaultAPIVersioningConfig(), versionManager, zap.NewNop())

	tests := []struct {
		name     string
		accept   string
		expected string
	}{
		{"valid version v2", "application/vnd.kyb-platform.v2+json", "v2"},
		{"valid version v3", "application/vnd.kyb-platform.v3+json", "v3"},
		{"multiple accept types", "application/json, application/vnd.kyb-platform.v2+json", "v2"},
		{"with quality value", "application/vnd.kyb-platform.v2+json;q=0.9", "v2"},
		{"no version", "application/json", ""},
		{"empty accept", "", ""},
		{"invalid format", "application/vnd.kyb-platform.version2+json", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.extractVersionFromAccept(tt.accept)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAPIVersioningMiddleware_ResolveVersion(t *testing.T) {
	versionManager := compatibility.NewVersionManager(zap.NewNop(), nil)

	tests := []struct {
		name             string
		config           *APIVersioningConfig
		requestedVersion string
		expectedVersion  string
		expectError      bool
	}{
		{
			name:             "supported version",
			config:           GetDefaultAPIVersioningConfig(),
			requestedVersion: "v3",
			expectedVersion:  "v3",
			expectError:      false,
		},
		{
			name:             "unsupported version with fallback",
			config:           GetDefaultAPIVersioningConfig(),
			requestedVersion: "v99",
			expectedVersion:  "v3", // Should fallback to current version
			expectError:      false,
		},
		{
			name: "unsupported version strict mode",
			config: func() *APIVersioningConfig {
				config := GetStrictAPIVersioningConfig()
				return config
			}(),
			requestedVersion: "v99",
			expectedVersion:  "",
			expectError:      true,
		},
		{
			name:             "deprecated version",
			config:           GetDefaultAPIVersioningConfig(),
			requestedVersion: "v1",
			expectedVersion:  "v1",
			expectError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewAPIVersioningMiddleware(tt.config, versionManager, zap.NewNop())

			result, err := middleware.resolveVersion(tt.requestedVersion)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedVersion, result)
			}
		})
	}
}

func TestAPIVersioningMiddleware_RemoveVersionFromPath(t *testing.T) {
	versionManager := compatibility.NewVersionManager(zap.NewNop(), nil)
	middleware := NewAPIVersioningMiddleware(GetDefaultAPIVersioningConfig(), versionManager, zap.NewNop())

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"version with path", "/v1/test", "/test"},
		{"version with nested path", "/v2/api/users", "/api/users"},
		{"version only", "/v3", "/v3"}, // No change if no path after version
		{"no version", "/test", "/test"},
		{"empty path", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := middleware.removeVersionFromPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAPIVersioningMiddleware_HandleVersionError(t *testing.T) {
	versionManager := compatibility.NewVersionManager(zap.NewNop(), nil)

	tests := []struct {
		name           string
		config         *APIVersioningConfig
		error          error
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "version error with error response",
			config:         GetDefaultAPIVersioningConfig(),
			error:          &VersionError{Type: "unsupported_version", Message: "Version not supported"},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  true,
		},
		{
			name: "version error with 404",
			config: func() *APIVersioningConfig {
				config := GetDefaultAPIVersioningConfig()
				config.ReturnVersionErrors = false
				return config
			}(),
			error:          &VersionError{Type: "unsupported_version", Message: "Version not supported"},
			expectedStatus: http.StatusNotFound,
			checkResponse:  false,
		},
		{
			name:           "generic error",
			config:         GetDefaultAPIVersioningConfig(),
			error:          assert.AnError,
			expectedStatus: http.StatusBadRequest,
			checkResponse:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create observer for logging
			core, obs := observer.New(zap.InfoLevel)
			logger := zap.New(core)

			middleware := NewAPIVersioningMiddleware(tt.config, versionManager, logger)

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			middleware.handleVersionError(w, req, tt.error)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse {
				assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response["success"].(bool))
				assert.NotNil(t, response["error"])
			}

			// Check logging
			if tt.config.LogVersionFailures {
				assert.Greater(t, obs.Len(), 0)
			}
		})
	}
}

func TestAPIVersioningMiddleware_GetSupportedVersions(t *testing.T) {
	versionManager := compatibility.NewVersionManager(zap.NewNop(), nil)
	middleware := NewAPIVersioningMiddleware(GetDefaultAPIVersioningConfig(), versionManager, zap.NewNop())

	versions := middleware.getSupportedVersions()

	// Should include v1, v2, v3 (v1 is deprecated but still supported)
	assert.Contains(t, versions, "v1")
	assert.Contains(t, versions, "v2")
	assert.Contains(t, versions, "v3")
	assert.Len(t, versions, 3)
}

func TestAPIVersioningMiddleware_ContextHelpers(t *testing.T) {
	// Test GetVersionInfo
	versionInfo := &VersionInfo{
		RequestedVersion: "v2",
		ResolvedVersion:  "v2",
		IsDeprecated:     false,
	}

	ctx := context.WithValue(context.Background(), "version_info", versionInfo)

	retrieved := GetVersionInfo(ctx)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "v2", retrieved.RequestedVersion)
	assert.Equal(t, "v2", retrieved.ResolvedVersion)

	// Test GetAPIVersion
	ctx = context.WithValue(context.Background(), "api_version", "v3")

	version := GetAPIVersion(ctx)
	assert.Equal(t, "v3", version)

	// Test with nil context
	assert.Nil(t, GetVersionInfo(context.Background()))
	assert.Empty(t, GetAPIVersion(context.Background()))
}

func TestAPIVersioningMiddleware_ConfigurationPresets(t *testing.T) {
	tests := []struct {
		name   string
		config *APIVersioningConfig
	}{
		{"default config", GetDefaultAPIVersioningConfig()},
		{"strict config", GetStrictAPIVersioningConfig()},
		{"permissive config", GetPermissiveAPIVersioningConfig()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, tt.config)
			assert.NotEmpty(t, tt.config.VersionHeaderName)
			assert.NotEmpty(t, tt.config.QueryVersionParam)
			assert.NotEmpty(t, tt.config.AcceptVersionPrefix)
			assert.NotEmpty(t, tt.config.ClientVersionHeader)
		})
	}

	// Test specific differences
	strictConfig := GetStrictAPIVersioningConfig()
	permissiveConfig := GetPermissiveAPIVersioningConfig()

	// Strict should be more restrictive
	assert.True(t, strictConfig.StrictVersioning)
	assert.False(t, strictConfig.AllowVersionFallback)
	assert.False(t, strictConfig.EnableQueryVersioning)

	// Permissive should be more lenient
	assert.False(t, permissiveConfig.StrictVersioning)
	assert.True(t, permissiveConfig.AllowVersionFallback)
	assert.False(t, permissiveConfig.ReturnVersionErrors)
	assert.False(t, permissiveConfig.LogVersionFailures)
	assert.False(t, permissiveConfig.EnableDeprecationWarnings)
	assert.False(t, permissiveConfig.EnableClientValidation)
}

func TestAPIVersioningMiddleware_Integration(t *testing.T) {
	versionManager := compatibility.NewVersionManager(zap.NewNop(), nil)
	middleware := NewAPIVersioningMiddleware(GetDefaultAPIVersioningConfig(), versionManager, zap.NewNop())

	// Test handler that checks context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		versionInfo := GetVersionInfo(r.Context())
		apiVersion := GetAPIVersion(r.Context())

		response := map[string]interface{}{
			"version_info": versionInfo,
			"api_version":  apiVersion,
			"path":         r.URL.Path,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	// Test different version detection methods
	tests := []struct {
		name            string
		requestPath     string
		headers         map[string]string
		expectedVersion string
	}{
		{
			name:            "URL versioning",
			requestPath:     "/v2/test",
			headers:         map[string]string{},
			expectedVersion: "v2",
		},
		{
			name:            "Header versioning",
			requestPath:     "/test",
			headers:         map[string]string{"X-API-Version": "v3"},
			expectedVersion: "v3",
		},
		{
			name:            "Query versioning",
			requestPath:     "/test?version=v1",
			headers:         map[string]string{},
			expectedVersion: "v1",
		},
		{
			name:            "Accept header versioning",
			requestPath:     "/test",
			headers:         map[string]string{"Accept": "application/vnd.kyb-platform.v2+json"},
			expectedVersion: "v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			w := httptest.NewRecorder()

			middleware.Middleware(handler).ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedVersion, response["api_version"])

			// Check headers
			assert.Equal(t, tt.expectedVersion, w.Header().Get("X-API-Version"))
		})
	}
}

func TestVersionError_Error(t *testing.T) {
	versionError := &VersionError{
		Type:    "unsupported_version",
		Message: "Version v99 is not supported",
		Code:    "UNSUPPORTED_VERSION",
	}

	assert.Equal(t, "Version v99 is not supported", versionError.Error())
}
