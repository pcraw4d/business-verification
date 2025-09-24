package compatibility

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// APIVersion represents an API version
type APIVersion struct {
	Version            string     `json:"version"`
	ReleaseDate        time.Time  `json:"release_date"`
	DeprecatedAt       *time.Time `json:"deprecated_at,omitempty"`
	RemovedAt          *time.Time `json:"removed_at,omitempty"`
	MinClientVersion   string     `json:"min_client_version,omitempty"`
	MaxClientVersion   string     `json:"max_client_version,omitempty"`
	BreakingChanges    []string   `json:"breaking_changes,omitempty"`
	Features           []string   `json:"features,omitempty"`
	DeprecationMessage string     `json:"deprecation_message,omitempty"`
	MigrationGuide     string     `json:"migration_guide,omitempty"`
}

// VersionCompatibility represents compatibility between versions
type VersionCompatibility struct {
	SourceVersion      string    `json:"source_version"`
	TargetVersion      string    `json:"target_version"`
	IsCompatible       bool      `json:"is_compatible"`
	CompatibilityLevel string    `json:"compatibility_level"` // "full", "partial", "none"
	BreakingChanges    []string  `json:"breaking_changes,omitempty"`
	MigrationRequired  bool      `json:"migration_required"`
	MigrationSteps     []string  `json:"migration_steps,omitempty"`
	LastChecked        time.Time `json:"last_checked"`
}

// VersionManager manages API versioning and compatibility
type VersionManager struct {
	logger              *zap.Logger
	versions            map[string]*APIVersion
	compatibility       map[string]*VersionCompatibility
	currentVersion      string
	defaultVersion      string
	minSupportedVersion string
	deprecationPeriod   time.Duration
	enableDeprecation   bool
}

// VersionConfig holds configuration for version management
type VersionConfig struct {
	CurrentVersion       string        `json:"current_version"`
	DefaultVersion       string        `json:"default_version"`
	MinSupportedVersion  string        `json:"min_supported_version"`
	DeprecationPeriod    time.Duration `json:"deprecation_period"`
	EnableAutoVersioning bool          `json:"enable_auto_versioning"`
	EnableDeprecation    bool          `json:"enable_deprecation"`
	StrictVersioning     bool          `json:"strict_versioning"`
}

// NewVersionManager creates a new version manager
func NewVersionManager(logger *zap.Logger, config *VersionConfig) *VersionManager {
	if config == nil {
		config = &VersionConfig{
			CurrentVersion:       "v3",
			DefaultVersion:       "v3",
			MinSupportedVersion:  "v1",
			DeprecationPeriod:    6 * 30 * 24 * time.Hour, // 6 months
			EnableAutoVersioning: true,
			EnableDeprecation:    true,
			StrictVersioning:     false,
		}
	}

	vm := &VersionManager{
		logger:              logger,
		versions:            make(map[string]*APIVersion),
		compatibility:       make(map[string]*VersionCompatibility),
		currentVersion:      config.CurrentVersion,
		defaultVersion:      config.DefaultVersion,
		minSupportedVersion: config.MinSupportedVersion,
		deprecationPeriod:   config.DeprecationPeriod,
		enableDeprecation:   config.EnableDeprecation,
	}

	// Initialize default versions
	vm.initializeDefaultVersions()

	return vm
}

// initializeDefaultVersions sets up the default API versions
func (vm *VersionManager) initializeDefaultVersions() {
	now := time.Now()

	// Version 1 (Legacy)
	vm.versions["v1"] = &APIVersion{
		Version:          "v1",
		ReleaseDate:      now.AddDate(0, -12, 0), // 1 year ago
		DeprecatedAt:     &now,
		MinClientVersion: "1.0.0",
		MaxClientVersion: "1.9.9",
		BreakingChanges: []string{
			"Limited industry code support",
			"Basic confidence scoring",
			"No metadata support",
		},
		Features: []string{
			"Basic business classification",
			"Single industry code",
			"Simple confidence score",
		},
		DeprecationMessage: "API v1 is deprecated. Please migrate to v2 or v3 for enhanced features.",
		MigrationGuide:     "https://docs.kyb-platform.com/migration/v1-to-v2",
	}

	// Version 2 (Enhanced)
	vm.versions["v2"] = &APIVersion{
		Version:          "v2",
		ReleaseDate:      now.AddDate(0, -6, 0), // 6 months ago
		MinClientVersion: "2.0.0",
		MaxClientVersion: "2.9.9",
		BreakingChanges: []string{
			"Enhanced response format",
			"Multiple industry codes",
			"Geographic region support",
		},
		Features: []string{
			"Enhanced business classification",
			"Multiple industry codes (MCC, SIC, NAICS)",
			"Geographic region classification",
			"Enhanced confidence scoring",
			"Basic metadata support",
		},
	}

	// Version 3 (Current)
	vm.versions["v3"] = &APIVersion{
		Version:          "v3",
		ReleaseDate:      now,
		MinClientVersion: "3.0.0",
		Features: []string{
			"Advanced business classification",
			"Intelligent routing system",
			"Comprehensive metadata",
			"Enhanced confidence scoring",
			"Data source tracking",
			"Quality assessment",
			"Traceability",
			"Compliance tracking",
		},
	}

	// Initialize compatibility matrix
	vm.initializeCompatibilityMatrix()
}

// initializeCompatibilityMatrix sets up the compatibility matrix between versions
func (vm *VersionManager) initializeCompatibilityMatrix() {
	now := time.Now()

	// v1 to v2 compatibility
	vm.compatibility["v1-v2"] = &VersionCompatibility{
		SourceVersion:      "v1",
		TargetVersion:      "v2",
		IsCompatible:       true,
		CompatibilityLevel: "partial",
		BreakingChanges: []string{
			"Response format changed",
			"Additional fields in response",
		},
		MigrationRequired: true,
		MigrationSteps: []string{
			"Update response parsing to handle new fields",
			"Handle multiple industry codes instead of single",
			"Update confidence score interpretation",
		},
		LastChecked: now,
	}

	// v2 to v3 compatibility
	vm.compatibility["v2-v3"] = &VersionCompatibility{
		SourceVersion:      "v2",
		TargetVersion:      "v3",
		IsCompatible:       true,
		CompatibilityLevel: "partial",
		BreakingChanges: []string{
			"Enhanced metadata structure",
			"Additional confidence factors",
			"New traceability fields",
		},
		MigrationRequired: true,
		MigrationSteps: []string{
			"Update to handle enhanced metadata",
			"Parse new confidence factors",
			"Handle traceability information",
		},
		LastChecked: now,
	}

	// v1 to v3 compatibility
	vm.compatibility["v1-v3"] = &VersionCompatibility{
		SourceVersion:      "v1",
		TargetVersion:      "v3",
		IsCompatible:       true,
		CompatibilityLevel: "partial",
		BreakingChanges: []string{
			"Major response format changes",
			"Multiple industry codes",
			"Enhanced metadata",
			"New confidence scoring",
		},
		MigrationRequired: true,
		MigrationSteps: []string{
			"Major response format update required",
			"Update to handle multiple industry codes",
			"Implement metadata handling",
			"Update confidence score interpretation",
		},
		LastChecked: now,
	}
}

// GetVersion retrieves version information
func (vm *VersionManager) GetVersion(ctx context.Context, version string) (*APIVersion, error) {
	if version == "" {
		return nil, fmt.Errorf("version cannot be empty")
	}

	apiVersion, exists := vm.versions[version]
	if !exists {
		return nil, fmt.Errorf("version not found: %s", version)
	}

	return apiVersion, nil
}

// ListVersions returns all available versions
func (vm *VersionManager) ListVersions(ctx context.Context) ([]*APIVersion, error) {
	var versions []*APIVersion
	for _, version := range vm.versions {
		versions = append(versions, version)
	}
	return versions, nil
}

// GetCurrentVersion returns the current API version
func (vm *VersionManager) GetCurrentVersion(ctx context.Context) string {
	return vm.currentVersion
}

// GetDefaultVersion returns the default API version
func (vm *VersionManager) GetDefaultVersion(ctx context.Context) string {
	return vm.defaultVersion
}

// IsVersionSupported checks if a version is supported
func (vm *VersionManager) IsVersionSupported(ctx context.Context, version string) bool {
	apiVersion, err := vm.GetVersion(ctx, version)
	if err != nil {
		return false
	}

	// Check if version is removed
	if apiVersion.RemovedAt != nil {
		return false
	}

	// Check if version is deprecated and deprecation period has passed
	if apiVersion.DeprecatedAt != nil && vm.enableDeprecation {
		if time.Since(*apiVersion.DeprecatedAt) > vm.deprecationPeriod {
			return false
		}
	}

	return true
}

// IsVersionDeprecated checks if a version is deprecated
func (vm *VersionManager) IsVersionDeprecated(ctx context.Context, version string) bool {
	apiVersion, err := vm.GetVersion(ctx, version)
	if err != nil {
		return false
	}

	return apiVersion.DeprecatedAt != nil
}

// GetCompatibility checks compatibility between two versions
func (vm *VersionManager) GetCompatibility(ctx context.Context, sourceVersion, targetVersion string) (*VersionCompatibility, error) {
	if sourceVersion == "" || targetVersion == "" {
		return nil, fmt.Errorf("source and target versions cannot be empty")
	}

	key := fmt.Sprintf("%s-%s", sourceVersion, targetVersion)
	compatibility, exists := vm.compatibility[key]
	if !exists {
		return nil, fmt.Errorf("compatibility information not found for %s to %s", sourceVersion, targetVersion)
	}

	return compatibility, nil
}

// NegotiateVersion determines the appropriate version based on request headers
func (vm *VersionManager) NegotiateVersion(ctx context.Context, r *http.Request) (string, error) {
	// Check Accept header for version preference
	acceptHeader := r.Header.Get("Accept")
	if acceptHeader != "" {
		version := vm.extractVersionFromAccept(acceptHeader)
		if version != "" && vm.IsVersionSupported(ctx, version) {
			return version, nil
		}
	}

	// Check X-API-Version header
	apiVersionHeader := r.Header.Get("X-API-Version")
	if apiVersionHeader != "" {
		if vm.IsVersionSupported(ctx, apiVersionHeader) {
			return apiVersionHeader, nil
		}
	}

	// Check URL path for version
	path := r.URL.Path
	if strings.HasPrefix(path, "/v") {
		parts := strings.Split(path[1:], "/")
		if len(parts) > 0 {
			version := parts[0]
			if vm.IsVersionSupported(ctx, version) {
				return version, nil
			}
		}
	}

	// Return default version
	return vm.defaultVersion, nil
}

// extractVersionFromAccept extracts version from Accept header
func (vm *VersionManager) extractVersionFromAccept(acceptHeader string) string {
	// Parse Accept header like: application/vnd.kyb-platform.v2+json
	parts := strings.Split(acceptHeader, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "vnd.kyb-platform.v") {
			// Extract version number
			start := strings.Index(part, "vnd.kyb-platform.v")
			if start != -1 {
				versionStart := start + len("vnd.kyb-platform.v")
				versionEnd := strings.Index(part[versionStart:], "+")
				if versionEnd == -1 {
					versionEnd = strings.Index(part[versionStart:], ";")
				}
				if versionEnd == -1 {
					versionEnd = len(part[versionStart:])
				}
				return "v" + part[versionStart:versionStart+versionEnd]
			}
		}
	}
	return ""
}

// AddDeprecationHeaders adds deprecation headers to response
func (vm *VersionManager) AddDeprecationHeaders(ctx context.Context, w http.ResponseWriter, version string) {
	apiVersion, err := vm.GetVersion(ctx, version)
	if err != nil {
		return
	}

	if apiVersion.DeprecatedAt != nil {
		w.Header().Set("X-API-Deprecated", "true")
		w.Header().Set("X-API-Deprecated-At", apiVersion.DeprecatedAt.Format(time.RFC3339))
		w.Header().Set("X-API-Deprecation-Message", apiVersion.DeprecationMessage)

		if apiVersion.MigrationGuide != "" {
			w.Header().Set("X-API-Migration-Guide", apiVersion.MigrationGuide)
		}

		// Calculate sunset date
		sunsetDate := apiVersion.DeprecatedAt.Add(vm.deprecationPeriod)
		w.Header().Set("X-API-Sunset-Date", sunsetDate.Format(time.RFC3339))
	}
}

// ValidateClientVersion validates client version compatibility
func (vm *VersionManager) ValidateClientVersion(ctx context.Context, version string, clientVersion string) error {
	apiVersion, err := vm.GetVersion(ctx, version)
	if err != nil {
		return fmt.Errorf("invalid API version: %w", err)
	}

	// Check minimum client version
	if apiVersion.MinClientVersion != "" {
		if !vm.isVersionGreaterOrEqual(clientVersion, apiVersion.MinClientVersion) {
			return fmt.Errorf("client version %s is below minimum required version %s", clientVersion, apiVersion.MinClientVersion)
		}
	}

	// Check maximum client version
	if apiVersion.MaxClientVersion != "" {
		if !vm.isVersionLessOrEqual(clientVersion, apiVersion.MaxClientVersion) {
			return fmt.Errorf("client version %s is above maximum supported version %s", clientVersion, apiVersion.MaxClientVersion)
		}
	}

	return nil
}

// isVersionGreaterOrEqual compares version strings
func (vm *VersionManager) isVersionGreaterOrEqual(version1, version2 string) bool {
	return vm.compareVersions(version1, version2) >= 0
}

// isVersionLessOrEqual compares version strings
func (vm *VersionManager) isVersionLessOrEqual(version1, version2 string) bool {
	return vm.compareVersions(version1, version2) <= 0
}

// compareVersions compares two version strings
func (vm *VersionManager) compareVersions(version1, version2 string) int {
	parts1 := strings.Split(version1, ".")
	parts2 := strings.Split(version2, ".")

	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}

	for i := 0; i < maxLen; i++ {
		var num1, num2 int
		if i < len(parts1) {
			num1, _ = strconv.Atoi(parts1[i])
		}
		if i < len(parts2) {
			num2, _ = strconv.Atoi(parts2[i])
		}

		if num1 > num2 {
			return 1
		} else if num1 < num2 {
			return -1
		}
	}

	return 0
}

// GetMigrationPath returns the migration path between versions
func (vm *VersionManager) GetMigrationPath(ctx context.Context, fromVersion, toVersion string) ([]string, error) {
	if fromVersion == toVersion {
		return []string{}, nil
	}

	// Simple migration path: direct migration
	compatibility, err := vm.GetCompatibility(ctx, fromVersion, toVersion)
	if err != nil {
		return nil, err
	}

	if !compatibility.IsCompatible {
		return nil, fmt.Errorf("no migration path available from %s to %s", fromVersion, toVersion)
	}

	return []string{fromVersion, toVersion}, nil
}
