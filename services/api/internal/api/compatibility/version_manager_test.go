package compatibility

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewVersionManager(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	vm := NewVersionManager(logger, nil)
	require.NotNil(t, vm)
	assert.Equal(t, "v3", vm.currentVersion)
	assert.Equal(t, "v3", vm.defaultVersion)
	assert.Equal(t, "v1", vm.minSupportedVersion)

	// Test with custom config
	config := &VersionConfig{
		CurrentVersion:       "v2",
		DefaultVersion:       "v2",
		MinSupportedVersion:  "v1",
		DeprecationPeriod:    12 * 30 * 24 * time.Hour, // 12 months
		EnableAutoVersioning: true,
		EnableDeprecation:    true,
		StrictVersioning:     true,
	}

	vm2 := NewVersionManager(logger, config)
	require.NotNil(t, vm2)
	assert.Equal(t, "v2", vm2.currentVersion)
	assert.Equal(t, "v2", vm2.defaultVersion)
	assert.Equal(t, "v1", vm2.minSupportedVersion)
}

func TestVersionManager_GetVersion(t *testing.T) {
	logger := zap.NewNop()
	vm := NewVersionManager(logger, nil)

	tests := []struct {
		name        string
		version     string
		expectError bool
	}{
		{
			name:        "get v1 version",
			version:     "v1",
			expectError: false,
		},
		{
			name:        "get v2 version",
			version:     "v2",
			expectError: false,
		},
		{
			name:        "get v3 version",
			version:     "v3",
			expectError: false,
		},
		{
			name:        "get non-existent version",
			version:     "v99",
			expectError: true,
		},
		{
			name:        "empty version",
			version:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := vm.GetVersion(context.Background(), tt.version)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, version)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, version)
				assert.Equal(t, tt.version, version.Version)
			}
		})
	}
}

func TestVersionManager_ListVersions(t *testing.T) {
	logger := zap.NewNop()
	vm := NewVersionManager(logger, nil)

	versions, err := vm.ListVersions(context.Background())
	require.NoError(t, err)
	require.NotNil(t, versions)

	// Should have at least 3 versions (v1, v2, v3)
	assert.GreaterOrEqual(t, len(versions), 3)

	// Check that all expected versions are present
	versionMap := make(map[string]bool)
	for _, v := range versions {
		versionMap[v.Version] = true
	}

	assert.True(t, versionMap["v1"])
	assert.True(t, versionMap["v2"])
	assert.True(t, versionMap["v3"])
}

func TestVersionManager_IsVersionSupported(t *testing.T) {
	logger := zap.NewNop()
	vm := NewVersionManager(logger, nil)

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "v1 is supported",
			version:  "v1",
			expected: true,
		},
		{
			name:     "v2 is supported",
			version:  "v2",
			expected: true,
		},
		{
			name:     "v3 is supported",
			version:  "v3",
			expected: true,
		},
		{
			name:     "v99 is not supported",
			version:  "v99",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			supported := vm.IsVersionSupported(context.Background(), tt.version)
			assert.Equal(t, tt.expected, supported)
		})
	}
}

func TestVersionManager_IsVersionDeprecated(t *testing.T) {
	logger := zap.NewNop()
	vm := NewVersionManager(logger, nil)

	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{
			name:     "v1 is deprecated",
			version:  "v1",
			expected: true,
		},
		{
			name:     "v2 is not deprecated",
			version:  "v2",
			expected: false,
		},
		{
			name:     "v3 is not deprecated",
			version:  "v3",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deprecated := vm.IsVersionDeprecated(context.Background(), tt.version)
			assert.Equal(t, tt.expected, deprecated)
		})
	}
}

func TestVersionManager_GetCompatibility(t *testing.T) {
	logger := zap.NewNop()
	vm := NewVersionManager(logger, nil)

	tests := []struct {
		name          string
		sourceVersion string
		targetVersion string
		expectError   bool
		expectedLevel string
	}{
		{
			name:          "v1 to v2 compatibility",
			sourceVersion: "v1",
			targetVersion: "v2",
			expectError:   false,
			expectedLevel: "partial",
		},
		{
			name:          "v2 to v3 compatibility",
			sourceVersion: "v2",
			targetVersion: "v3",
			expectError:   false,
			expectedLevel: "partial",
		},
		{
			name:          "v1 to v3 compatibility",
			sourceVersion: "v1",
			targetVersion: "v3",
			expectError:   false,
			expectedLevel: "partial",
		},
		{
			name:          "non-existent compatibility",
			sourceVersion: "v1",
			targetVersion: "v99",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compatibility, err := vm.GetCompatibility(context.Background(), tt.sourceVersion, tt.targetVersion)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, compatibility)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, compatibility)
				assert.Equal(t, tt.sourceVersion, compatibility.SourceVersion)
				assert.Equal(t, tt.targetVersion, compatibility.TargetVersion)
				assert.Equal(t, tt.expectedLevel, compatibility.CompatibilityLevel)
				assert.True(t, compatibility.IsCompatible)
			}
		})
	}
}

func TestVersionManager_NegotiateVersion(t *testing.T) {
	logger := zap.NewNop()
	vm := NewVersionManager(logger, nil)

	tests := []struct {
		name            string
		acceptHeader    string
		apiVersion      string
		expectedVersion string
	}{
		{
			name:            "Accept header v1",
			acceptHeader:    "application/vnd.kyb-platform.v1+json",
			expectedVersion: "v1",
		},
		{
			name:            "Accept header v2",
			acceptHeader:    "application/vnd.kyb-platform.v2+json",
			expectedVersion: "v2",
		},
		{
			name:            "Accept header v3",
			acceptHeader:    "application/vnd.kyb-platform.v3+json",
			expectedVersion: "v3",
		},
		{
			name:            "X-API-Version header v1",
			apiVersion:      "v1",
			expectedVersion: "v1",
		},
		{
			name:            "X-API-Version header v2",
			apiVersion:      "v2",
			expectedVersion: "v2",
		},
		{
			name:            "X-API-Version header v3",
			apiVersion:      "v3",
			expectedVersion: "v3",
		},
		{
			name:            "no headers - default to v3",
			expectedVersion: "v3",
		},
		{
			name:            "Accept header takes precedence",
			acceptHeader:    "application/vnd.kyb-platform.v1+json",
			apiVersion:      "v3",
			expectedVersion: "v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/classify", nil)

			if tt.acceptHeader != "" {
				req.Header.Set("Accept", tt.acceptHeader)
			}
			if tt.apiVersion != "" {
				req.Header.Set("X-API-Version", tt.apiVersion)
			}

			negotiatedVersion, err := vm.NegotiateVersion(context.Background(), req)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedVersion, negotiatedVersion)
		})
	}
}

func TestVersionManager_ValidateClientVersion(t *testing.T) {
	logger := zap.NewNop()
	vm := NewVersionManager(logger, nil)

	tests := []struct {
		name          string
		version       string
		clientVersion string
		expectError   bool
	}{
		{
			name:          "valid client version for v1",
			version:       "v1",
			clientVersion: "1.0.0",
			expectError:   false,
		},
		{
			name:          "valid client version for v2",
			version:       "v2",
			clientVersion: "2.0.0",
			expectError:   false,
		},
		{
			name:          "valid client version for v3",
			version:       "v3",
			clientVersion: "3.0.0",
			expectError:   false,
		},
		{
			name:          "client version too low",
			version:       "v1",
			clientVersion: "0.9.0",
			expectError:   true,
		},
		{
			name:          "client version too high",
			version:       "v1",
			clientVersion: "2.0.0",
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := vm.ValidateClientVersion(context.Background(), tt.version, tt.clientVersion)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestVersionManager_AddDeprecationHeaders(t *testing.T) {
	logger := zap.NewNop()
	vm := NewVersionManager(logger, nil)

	// Test with deprecated version (v1)
	w := httptest.NewRecorder()
	vm.AddDeprecationHeaders(context.Background(), w, "v1")

	assert.Equal(t, "true", w.Header().Get("X-API-Deprecated"))
	assert.NotEmpty(t, w.Header().Get("X-API-Deprecated-At"))
	assert.NotEmpty(t, w.Header().Get("X-API-Deprecation-Message"))
	assert.NotEmpty(t, w.Header().Get("X-API-Sunset-Date"))

	// Test with non-deprecated version (v3)
	w2 := httptest.NewRecorder()
	vm.AddDeprecationHeaders(context.Background(), w2, "v3")

	assert.Empty(t, w2.Header().Get("X-API-Deprecated"))
	assert.Empty(t, w2.Header().Get("X-API-Deprecated-At"))
	assert.Empty(t, w2.Header().Get("X-API-Deprecation-Message"))
	assert.Empty(t, w2.Header().Get("X-API-Sunset-Date"))
}

func TestVersionManager_GetMigrationPath(t *testing.T) {
	logger := zap.NewNop()
	vm := NewVersionManager(logger, nil)

	tests := []struct {
		name         string
		fromVersion  string
		toVersion    string
		expectError  bool
		expectedPath []string
	}{
		{
			name:         "v1 to v2 migration",
			fromVersion:  "v1",
			toVersion:    "v2",
			expectError:  false,
			expectedPath: []string{"v1", "v2"},
		},
		{
			name:         "v2 to v3 migration",
			fromVersion:  "v2",
			toVersion:    "v3",
			expectError:  false,
			expectedPath: []string{"v2", "v3"},
		},
		{
			name:         "v1 to v3 migration",
			fromVersion:  "v1",
			toVersion:    "v3",
			expectError:  false,
			expectedPath: []string{"v1", "v3"},
		},
		{
			name:         "same version",
			fromVersion:  "v3",
			toVersion:    "v3",
			expectError:  false,
			expectedPath: []string{},
		},
		{
			name:        "non-existent target version",
			fromVersion: "v1",
			toVersion:   "v99",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := vm.GetMigrationPath(context.Background(), tt.fromVersion, tt.toVersion)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, path)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPath, path)
			}
		})
	}
}

func TestVersionManager_CompareVersions(t *testing.T) {
	logger := zap.NewNop()
	vm := NewVersionManager(logger, nil)

	tests := []struct {
		name     string
		version1 string
		version2 string
		expected int
	}{
		{
			name:     "1.0.0 equals 1.0.0",
			version1: "1.0.0",
			version2: "1.0.0",
			expected: 0,
		},
		{
			name:     "1.0.0 less than 1.1.0",
			version1: "1.0.0",
			version2: "1.1.0",
			expected: -1,
		},
		{
			name:     "1.1.0 greater than 1.0.0",
			version1: "1.1.0",
			version2: "1.0.0",
			expected: 1,
		},
		{
			name:     "1.0.0 less than 2.0.0",
			version1: "1.0.0",
			version2: "2.0.0",
			expected: -1,
		},
		{
			name:     "2.0.0 greater than 1.9.9",
			version1: "2.0.0",
			version2: "1.9.9",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.compareVersions(tt.version1, tt.version2)
			assert.Equal(t, tt.expected, result)
		})
	}
}
