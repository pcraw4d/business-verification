package api

import (
	"net/http"
	"strings"
)

// APIVersion represents an API version
type APIVersion struct {
	Version string
	Path    string
}

// VersionManager manages API versioning
type VersionManager struct {
	versions       map[string]APIVersion
	defaultVersion string
}

// NewVersionManager creates a new version manager
func NewVersionManager() *VersionManager {
	return &VersionManager{
		versions:       make(map[string]APIVersion),
		defaultVersion: "v1",
	}
}

// AddVersion adds a new API version
func (vm *VersionManager) AddVersion(version string, path string) {
	vm.versions[version] = APIVersion{
		Version: version,
		Path:    path,
	}
}

// SetDefaultVersion sets the default API version
func (vm *VersionManager) SetDefaultVersion(version string) {
	vm.defaultVersion = version
}

// GetVersionFromRequest extracts API version from request
func (vm *VersionManager) GetVersionFromRequest(r *http.Request) string {
	// Check URL path for version
	path := r.URL.Path
	if strings.HasPrefix(path, "/v") {
		parts := strings.Split(path, "/")
		if len(parts) > 1 {
			version := parts[1]
			if _, exists := vm.versions[version]; exists {
				return version
			}
		}
	}

	// Check Accept header for version
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "version=") {
		// Extract version from Accept header
		// Example: application/json; version=v2
		parts := strings.Split(accept, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if strings.HasPrefix(part, "version=") {
				version := strings.TrimPrefix(part, "version=")
				if _, exists := vm.versions[version]; exists {
					return version
				}
			}
		}
	}

	// Check custom header for version
	version := r.Header.Get("X-API-Version")
	if version != "" {
		if _, exists := vm.versions[version]; exists {
			return version
		}
	}

	// Return default version
	return vm.defaultVersion
}

// GetVersionPath returns the path for a specific version
func (vm *VersionManager) GetVersionPath(version string) string {
	if apiVersion, exists := vm.versions[version]; exists {
		return apiVersion.Path
	}
	return vm.versions[vm.defaultVersion].Path
}

// IsVersionSupported checks if a version is supported
func (vm *VersionManager) IsVersionSupported(version string) bool {
	_, exists := vm.versions[version]
	return exists
}

// GetSupportedVersions returns a list of supported versions
func (vm *VersionManager) GetSupportedVersions() []string {
	versions := make([]string, 0, len(vm.versions))
	for version := range vm.versions {
		versions = append(versions, version)
	}
	return versions
}

// VersionMiddleware adds version information to response headers
func (vm *VersionManager) VersionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		version := vm.GetVersionFromRequest(r)

		// Add version headers
		w.Header().Set("X-API-Version", version)
		w.Header().Set("X-Supported-Versions", strings.Join(vm.GetSupportedVersions(), ", "))

		next.ServeHTTP(w, r)
	})
}
