package metadata

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// MetadataVersion represents a metadata version
type MetadataVersion struct {
	Version       string                 `json:"version"`
	SchemaVersion string                 `json:"schema_version"`
	CreatedAt     time.Time              `json:"created_at"`
	DeprecatedAt  *time.Time             `json:"deprecated_at,omitempty"`
	RemovedAt     *time.Time             `json:"removed_at,omitempty"`
	Changes       []VersionChange        `json:"changes"`
	Compatibility CompatibilityInfo      `json:"compatibility"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// VersionChange represents a change in metadata version
type VersionChange struct {
	ChangeID      string `json:"change_id"`
	ChangeType    string `json:"change_type"` // "added", "modified", "deprecated", "removed"
	Field         string `json:"field"`
	Description   string `json:"description"`
	Breaking      bool   `json:"breaking"`
	MigrationPath string `json:"migration_path,omitempty"`
}

// CompatibilityInfo represents compatibility information
type CompatibilityInfo struct {
	BackwardCompatible bool     `json:"backward_compatible"`
	ForwardCompatible  bool     `json:"forward_compatible"`
	SupportedVersions  []string `json:"supported_versions"`
	DeprecatedVersions []string `json:"deprecated_versions"`
	RemovedVersions    []string `json:"removed_versions"`
}

// MetadataVersionManager provides functionality for managing metadata versions
type MetadataVersionManager struct {
	logger *zap.Logger
	config *VersioningConfig

	// Version registry
	versions map[string]*MetadataVersion
}

// VersioningConfig contains configuration for metadata versioning
type VersioningConfig struct {
	// Versioning settings
	CurrentVersion      string `json:"current_version"`
	DefaultVersion      string `json:"default_version"`
	MinSupportedVersion string `json:"min_supported_version"`

	// Evolution settings
	EnableAutoVersioning bool          `json:"enable_auto_versioning"`
	EnableDeprecation    bool          `json:"enable_deprecation"`
	DeprecationPeriod    time.Duration `json:"deprecation_period"`

	// Migration settings
	EnableAutoMigration bool          `json:"enable_auto_migration"`
	MigrationTimeout    time.Duration `json:"migration_timeout"`

	// Validation settings
	StrictVersioning bool `json:"strict_versioning"`
	ValidateOnLoad   bool `json:"validate_on_load"`
}

// NewMetadataVersionManager creates a new metadata version manager
func NewMetadataVersionManager(logger *zap.Logger, config *VersioningConfig) *MetadataVersionManager {
	if config == nil {
		config = getDefaultVersioningConfig()
	}

	vm := &MetadataVersionManager{
		logger:   logger,
		config:   config,
		versions: make(map[string]*MetadataVersion),
	}

	// Initialize with default versions
	vm.initializeDefaultVersions()

	return vm
}

// RegisterVersion registers a new metadata version
func (vm *MetadataVersionManager) RegisterVersion(ctx context.Context, version *MetadataVersion) error {
	if version == nil {
		return fmt.Errorf("version cannot be nil")
	}

	if version.Version == "" {
		return fmt.Errorf("version identifier cannot be empty")
	}

	// Check if version already exists
	if _, exists := vm.versions[version.Version]; exists {
		return fmt.Errorf("version %s already registered", version.Version)
	}

	// Set creation timestamp if not provided
	if version.CreatedAt.IsZero() {
		version.CreatedAt = time.Now()
	}

	// Initialize metadata if not provided
	if version.Metadata == nil {
		version.Metadata = make(map[string]interface{})
	}

	vm.versions[version.Version] = version

	vm.logger.Info("Metadata version registered",
		zap.String("version", version.Version),
		zap.String("schema_version", version.SchemaVersion))

	return nil
}

// GetVersion retrieves a metadata version
func (vm *MetadataVersionManager) GetVersion(ctx context.Context, version string) (*MetadataVersion, error) {
	ver, exists := vm.versions[version]
	if !exists {
		return nil, fmt.Errorf("version %s not found", version)
	}

	return ver, nil
}

// ListVersions returns all registered versions
func (vm *MetadataVersionManager) ListVersions(ctx context.Context) ([]*MetadataVersion, error) {
	versions := make([]*MetadataVersion, 0, len(vm.versions))
	for _, version := range vm.versions {
		versions = append(versions, version)
	}

	return versions, nil
}

// DeprecateVersion marks a version as deprecated
func (vm *MetadataVersionManager) DeprecateVersion(ctx context.Context, version string, reason string) error {
	ver, exists := vm.versions[version]
	if !exists {
		return fmt.Errorf("version %s not found", version)
	}

	if !vm.config.EnableDeprecation {
		return fmt.Errorf("deprecation is not enabled")
	}

	now := time.Now()
	ver.DeprecatedAt = &now

	// Add deprecation change
	ver.Changes = append(ver.Changes, VersionChange{
		ChangeID:    fmt.Sprintf("deprecate_%s", version),
		ChangeType:  "deprecated",
		Field:       "version",
		Description: fmt.Sprintf("Version %s deprecated: %s", version, reason),
		Breaking:    false,
	})

	// Update compatibility info
	ver.Compatibility.DeprecatedVersions = append(ver.Compatibility.DeprecatedVersions, version)

	vm.logger.Info("Metadata version deprecated",
		zap.String("version", version),
		zap.String("reason", reason))

	return nil
}

// RemoveVersion marks a version as removed
func (vm *MetadataVersionManager) RemoveVersion(ctx context.Context, version string, reason string) error {
	ver, exists := vm.versions[version]
	if !exists {
		return fmt.Errorf("version %s not found", version)
	}

	// Check if version is deprecated first
	if ver.DeprecatedAt == nil {
		return fmt.Errorf("version %s must be deprecated before removal", version)
	}

	// Check deprecation period
	if time.Since(*ver.DeprecatedAt) < vm.config.DeprecationPeriod {
		return fmt.Errorf("version %s cannot be removed before deprecation period ends", version)
	}

	now := time.Now()
	ver.RemovedAt = &now

	// Add removal change
	ver.Changes = append(ver.Changes, VersionChange{
		ChangeID:    fmt.Sprintf("remove_%s", version),
		ChangeType:  "removed",
		Field:       "version",
		Description: fmt.Sprintf("Version %s removed: %s", version, reason),
		Breaking:    true,
	})

	// Update compatibility info
	ver.Compatibility.RemovedVersions = append(ver.Compatibility.RemovedVersions, version)

	vm.logger.Info("Metadata version removed",
		zap.String("version", version),
		zap.String("reason", reason))

	return nil
}

// CheckCompatibility checks compatibility between versions
func (vm *MetadataVersionManager) CheckCompatibility(ctx context.Context, fromVersion, toVersion string) (*CompatibilityResult, error) {
	fromVer, exists := vm.versions[fromVersion]
	if !exists {
		return nil, fmt.Errorf("source version %s not found", fromVersion)
	}

	toVer, exists := vm.versions[toVersion]
	if !exists {
		return nil, fmt.Errorf("target version %s not found", toVersion)
	}

	result := &CompatibilityResult{
		FromVersion:   fromVersion,
		ToVersion:     toVersion,
		Compatible:    true,
		Issues:        []CompatibilityIssue{},
		Warnings:      []CompatibilityWarning{},
		MigrationPath: []string{},
	}

	// Check if target version is deprecated
	if toVer.DeprecatedAt != nil {
		result.Warnings = append(result.Warnings, CompatibilityWarning{
			Type:           "deprecated_version",
			Message:        fmt.Sprintf("Target version %s is deprecated", toVersion),
			Severity:       "medium",
			Recommendation: fmt.Sprintf("Consider upgrading to a supported version"),
		})
	}

	// Check if target version is removed
	if toVer.RemovedAt != nil {
		result.Compatible = false
		result.Issues = append(result.Issues, CompatibilityIssue{
			Type:     "removed_version",
			Message:  fmt.Sprintf("Target version %s has been removed", toVersion),
			Severity: "high",
		})
	}

	// Check if source version is removed
	if fromVer.RemovedAt != nil {
		result.Compatible = false
		result.Issues = append(result.Issues, CompatibilityIssue{
			Type:     "removed_version",
			Message:  fmt.Sprintf("Source version %s has been removed", fromVersion),
			Severity: "high",
		})
	}

	// Check breaking changes
	breakingChanges := vm.findBreakingChanges(fromVersion, toVersion)
	for _, change := range breakingChanges {
		result.Compatible = false
		result.Issues = append(result.Issues, CompatibilityIssue{
			Type:     "breaking_change",
			Message:  fmt.Sprintf("Breaking change: %s", change.Description),
			Severity: "high",
			Field:    change.Field,
		})
	}

	// Generate migration path if compatible
	if result.Compatible {
		result.MigrationPath = vm.generateMigrationPath(fromVersion, toVersion)
	}

	return result, nil
}

// MigrateMetadata migrates metadata from one version to another
func (vm *MetadataVersionManager) MigrateMetadata(ctx context.Context, metadata *ResponseMetadata, targetVersion string) (*ResponseMetadata, error) {
	if metadata == nil {
		return nil, fmt.Errorf("metadata cannot be nil")
	}

	// Check compatibility
	compatibility, err := vm.CheckCompatibility(ctx, metadata.APIVersion, targetVersion)
	if err != nil {
		return nil, fmt.Errorf("compatibility check failed: %w", err)
	}

	if !compatibility.Compatible {
		return nil, fmt.Errorf("incompatible versions: %s -> %s", metadata.APIVersion, targetVersion)
	}

	// Create a copy of the metadata
	migratedMetadata := vm.deepCopyMetadata(metadata)

	// Apply migrations along the path
	for _, version := range compatibility.MigrationPath {
		migratedMetadata, err = vm.applyMigration(migratedMetadata, version)
		if err != nil {
			return nil, fmt.Errorf("migration to %s failed: %w", version, err)
		}
	}

	// Update version
	migratedMetadata.APIVersion = targetVersion

	vm.logger.Info("Metadata migration completed",
		zap.String("from_version", metadata.APIVersion),
		zap.String("to_version", targetVersion))

	return migratedMetadata, nil
}

// GetCurrentVersion returns the current metadata version
func (vm *MetadataVersionManager) GetCurrentVersion(ctx context.Context) string {
	return vm.config.CurrentVersion
}

// GetDefaultVersion returns the default metadata version
func (vm *MetadataVersionManager) GetDefaultVersion(ctx context.Context) string {
	return vm.config.DefaultVersion
}

// GetSupportedVersions returns all supported versions
func (vm *MetadataVersionManager) GetSupportedVersions(ctx context.Context) ([]string, error) {
	var supported []string

	for version, ver := range vm.versions {
		// Skip deprecated and removed versions
		if ver.DeprecatedAt == nil && ver.RemovedAt == nil {
			supported = append(supported, version)
		}
	}

	return supported, nil
}

// Helper types and methods

// CompatibilityResult represents the result of a compatibility check
type CompatibilityResult struct {
	FromVersion   string                 `json:"from_version"`
	ToVersion     string                 `json:"to_version"`
	Compatible    bool                   `json:"compatible"`
	Issues        []CompatibilityIssue   `json:"issues"`
	Warnings      []CompatibilityWarning `json:"warnings"`
	MigrationPath []string               `json:"migration_path"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// CompatibilityIssue represents a compatibility issue
type CompatibilityIssue struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Field    string `json:"field,omitempty"`
}

// CompatibilityWarning represents a compatibility warning
type CompatibilityWarning struct {
	Type           string `json:"type"`
	Message        string `json:"message"`
	Severity       string `json:"severity"`
	Recommendation string `json:"recommendation,omitempty"`
}

func (vm *MetadataVersionManager) initializeDefaultVersions() {
	// Version 1.0 - Initial version
	v1 := &MetadataVersion{
		Version:       "1.0",
		SchemaVersion: "1.0.0",
		CreatedAt:     time.Now(),
		Changes:       []VersionChange{},
		Compatibility: CompatibilityInfo{
			BackwardCompatible: true,
			ForwardCompatible:  false,
			SupportedVersions:  []string{"1.0"},
			DeprecatedVersions: []string{},
			RemovedVersions:    []string{},
		},
		Metadata: make(map[string]interface{}),
	}

	// Version 2.0 - Enhanced metadata
	v2 := &MetadataVersion{
		Version:       "2.0",
		SchemaVersion: "2.0.0",
		CreatedAt:     time.Now(),
		Changes: []VersionChange{
			{
				ChangeID:    "add_confidence_metadata",
				ChangeType:  "added",
				Field:       "confidence",
				Description: "Added detailed confidence metadata",
				Breaking:    false,
			},
			{
				ChangeID:    "add_quality_metadata",
				ChangeType:  "added",
				Field:       "quality",
				Description: "Added quality assessment metadata",
				Breaking:    false,
			},
		},
		Compatibility: CompatibilityInfo{
			BackwardCompatible: true,
			ForwardCompatible:  false,
			SupportedVersions:  []string{"1.0", "2.0"},
			DeprecatedVersions: []string{},
			RemovedVersions:    []string{},
		},
		Metadata: make(map[string]interface{}),
	}

	// Version 3.0 - Current version with comprehensive metadata
	v3 := &MetadataVersion{
		Version:       "3.0",
		SchemaVersion: "3.0.0",
		CreatedAt:     time.Now(),
		Changes: []VersionChange{
			{
				ChangeID:    "add_traceability_metadata",
				ChangeType:  "added",
				Field:       "traceability",
				Description: "Added comprehensive traceability metadata",
				Breaking:    false,
			},
			{
				ChangeID:    "add_compliance_metadata",
				ChangeType:  "added",
				Field:       "compliance",
				Description: "Added compliance assessment metadata",
				Breaking:    false,
			},
			{
				ChangeID:    "enhance_data_source_metadata",
				ChangeType:  "modified",
				Field:       "data_sources",
				Description: "Enhanced data source metadata with quality and performance metrics",
				Breaking:    false,
			},
		},
		Compatibility: CompatibilityInfo{
			BackwardCompatible: true,
			ForwardCompatible:  false,
			SupportedVersions:  []string{"1.0", "2.0", "3.0"},
			DeprecatedVersions: []string{},
			RemovedVersions:    []string{},
		},
		Metadata: make(map[string]interface{}),
	}

	vm.versions["1.0"] = v1
	vm.versions["2.0"] = v2
	vm.versions["3.0"] = v3
}

func (vm *MetadataVersionManager) findBreakingChanges(fromVersion, toVersion string) []VersionChange {
	var breakingChanges []VersionChange

	// This is a simplified implementation
	// In a real system, you would have a more sophisticated change tracking system

	fromVer := vm.versions[fromVersion]
	toVer := vm.versions[toVersion]

	if fromVer == nil || toVer == nil {
		return breakingChanges
	}

	// Check for breaking changes in the target version
	for _, change := range toVer.Changes {
		if change.Breaking {
			breakingChanges = append(breakingChanges, change)
		}
	}

	return breakingChanges
}

func (vm *MetadataVersionManager) generateMigrationPath(fromVersion, toVersion string) []string {
	// This is a simplified implementation
	// In a real system, you would have a more sophisticated path finding algorithm

	if fromVersion == toVersion {
		return []string{}
	}

	// Simple linear migration path
	return []string{toVersion}
}

func (vm *MetadataVersionManager) deepCopyMetadata(metadata *ResponseMetadata) *ResponseMetadata {
	if metadata == nil {
		return nil
	}

	copy := &ResponseMetadata{
		RequestID:      metadata.RequestID,
		ProcessingTime: metadata.ProcessingTime,
		Timestamp:      metadata.Timestamp,
		APIVersion:     metadata.APIVersion,
		DataSources:    make([]DataSourceMetadata, len(metadata.DataSources)),
		Metadata:       make(map[string]interface{}),
	}

	// Copy data sources
	for i, source := range metadata.DataSources {
		copy.DataSources[i] = source
	}

	// Copy other fields
	if metadata.Confidence != nil {
		copy.Confidence = metadata.Confidence
	}
	if metadata.Validation != nil {
		copy.Validation = metadata.Validation
	}
	if metadata.Quality != nil {
		copy.Quality = metadata.Quality
	}
	if metadata.Traceability != nil {
		copy.Traceability = metadata.Traceability
	}
	if metadata.Compliance != nil {
		copy.Compliance = metadata.Compliance
	}

	// Copy metadata
	for k, v := range metadata.Metadata {
		copy.Metadata[k] = v
	}

	return copy
}

func (vm *MetadataVersionManager) applyMigration(metadata *ResponseMetadata, targetVersion string) (*ResponseMetadata, error) {
	// This is a simplified implementation
	// In a real system, you would have specific migration logic for each version

	// For now, just return the metadata as-is
	// In practice, you would apply specific transformations based on the target version

	return metadata, nil
}

func getDefaultVersioningConfig() *VersioningConfig {
	return &VersioningConfig{
		CurrentVersion:       "3.0",
		DefaultVersion:       "3.0",
		MinSupportedVersion:  "1.0",
		EnableAutoVersioning: true,
		EnableDeprecation:    true,
		DeprecationPeriod:    6 * 30 * 24 * time.Hour, // 6 months
		EnableAutoMigration:  true,
		MigrationTimeout:     30 * time.Second,
		StrictVersioning:     false,
		ValidateOnLoad:       true,
	}
}
