package metadata

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMetadataManager_AddDataSource(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataManager(logger, nil)

	ctx := context.Background()

	t.Run("successful add", func(t *testing.T) {
		source := &DataSourceMetadata{
			SourceID:         "test_source_1",
			SourceName:       "Test Source 1",
			SourceType:       "api",
			ReliabilityScore: 0.95,
		}

		err := manager.AddDataSource(ctx, source)
		require.NoError(t, err)

		// Verify it was added
		retrieved, err := manager.GetDataSource(ctx, "test_source_1")
		require.NoError(t, err)
		assert.Equal(t, source.SourceID, retrieved.SourceID)
		assert.Equal(t, source.SourceName, retrieved.SourceName)
		assert.Equal(t, source.SourceType, retrieved.SourceType)
		assert.Equal(t, source.ReliabilityScore, retrieved.ReliabilityScore)
	})

	t.Run("nil source", func(t *testing.T) {
		err := manager.AddDataSource(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("empty source ID", func(t *testing.T) {
		source := &DataSourceMetadata{
			SourceName:       "Test Source",
			SourceType:       "api",
			ReliabilityScore: 0.95,
		}

		err := manager.AddDataSource(ctx, source)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("duplicate source ID", func(t *testing.T) {
		source := &DataSourceMetadata{
			SourceID:         "test_source_2",
			SourceName:       "Test Source 2",
			SourceType:       "api",
			ReliabilityScore: 0.95,
		}

		err := manager.AddDataSource(ctx, source)
		require.NoError(t, err)

		// Try to add again
		err = manager.AddDataSource(ctx, source)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
}

func TestMetadataManager_CalculateConfidence(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataManager(logger, nil)

	ctx := context.Background()

	t.Run("successful calculation", func(t *testing.T) {
		factors := []ConfidenceFactor{
			{
				FactorName:   "data_quality",
				FactorValue:  0.9,
				FactorWeight: 0.3,
				Description:  "High quality data",
				Impact:       "positive",
				Confidence:   0.9,
			},
			{
				FactorName:   "source_reliability",
				FactorValue:  0.8,
				FactorWeight: 0.2,
				Description:  "Reliable source",
				Impact:       "positive",
				Confidence:   0.8,
			},
			{
				FactorName:   "validation",
				FactorValue:  0.95,
				FactorWeight: 0.2,
				Description:  "Strong validation",
				Impact:       "positive",
				Confidence:   0.95,
			},
		}

		confidence, err := manager.CalculateConfidence(ctx, factors)
		require.NoError(t, err)
		assert.NotNil(t, confidence)
		assert.Greater(t, confidence.OverallConfidence, 0.0)
		assert.LessOrEqual(t, confidence.OverallConfidence, 1.0)
		assert.Equal(t, 3, len(confidence.Factors))
		assert.Equal(t, 3, len(confidence.ComponentScores))
	})

	t.Run("empty factors", func(t *testing.T) {
		_, err := manager.CalculateConfidence(ctx, []ConfidenceFactor{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one confidence factor is required")
	})

	t.Run("nil factors", func(t *testing.T) {
		_, err := manager.CalculateConfidence(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least one confidence factor is required")
	})
}

func TestMetadataManager_CreateResponseMetadata(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataManager(logger, nil)

	ctx := context.Background()

	t.Run("successful creation", func(t *testing.T) {
		metadata, err := manager.CreateResponseMetadata(ctx, "test_request_1")
		require.NoError(t, err)
		assert.NotNil(t, metadata)
		assert.Equal(t, "test_request_1", metadata.RequestID)
		assert.Equal(t, "v3", metadata.APIVersion)
		assert.False(t, metadata.Timestamp.IsZero())
		assert.NotNil(t, metadata.DataSources)
		assert.NotNil(t, metadata.Metadata)
	})

	t.Run("auto-generated request ID", func(t *testing.T) {
		metadata, err := manager.CreateResponseMetadata(ctx, "")
		require.NoError(t, err)
		assert.NotNil(t, metadata)
		assert.NotEmpty(t, metadata.RequestID)
		assert.True(t, len(metadata.RequestID) > 0)
	})
}

func TestMetadataValidator_ValidateMetadata(t *testing.T) {
	logger := zap.NewNop()
	validator := NewMetadataValidator(logger, nil)

	ctx := context.Background()

	t.Run("valid metadata", func(t *testing.T) {
		metadata := &ResponseMetadata{
			RequestID:      "test_request_1",
			Timestamp:      time.Now(),
			APIVersion:     "v3",
			ProcessingTime: 100 * time.Millisecond,
			DataSources:    []DataSourceMetadata{},
			Metadata:       make(map[string]interface{}),
		}

		result, err := validator.ValidateMetadata(ctx, metadata)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.Equal(t, 1.0, result.ValidationScore)
		assert.Empty(t, result.Errors)
	})

	t.Run("missing required fields", func(t *testing.T) {
		metadata := &ResponseMetadata{
			ProcessingTime: 100 * time.Millisecond,
			DataSources:    []DataSourceMetadata{},
			Metadata:       make(map[string]interface{}),
		}

		result, err := validator.ValidateMetadata(ctx, metadata)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsValid)
		assert.Less(t, result.ValidationScore, 1.0)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("invalid confidence range", func(t *testing.T) {
		metadata := &ResponseMetadata{
			RequestID:      "test_request_1",
			Timestamp:      time.Now(),
			APIVersion:     "v3",
			ProcessingTime: 100 * time.Millisecond,
			DataSources:    []DataSourceMetadata{},
			Confidence: &ConfidenceMetadata{
				OverallConfidence: 1.5, // Invalid: > 1.0
				ConfidenceLevel:   ConfidenceLevelHigh,
				ComponentScores:   make(map[string]float64),
				Factors:           []ConfidenceFactor{},
				CalculatedAt:      time.Now(),
				Metadata:          make(map[string]interface{}),
			},
			Metadata: make(map[string]interface{}),
		}

		result, err := validator.ValidateMetadata(ctx, metadata)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsValid)
		assert.Less(t, result.ValidationScore, 1.0)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("nil metadata", func(t *testing.T) {
		_, err := validator.ValidateMetadata(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestMetadataValidator_ValidateDataSource(t *testing.T) {
	logger := zap.NewNop()
	validator := NewMetadataValidator(logger, nil)

	ctx := context.Background()

	t.Run("valid data source", func(t *testing.T) {
		source := &DataSourceMetadata{
			SourceID:         "test_source_1",
			SourceName:       "Test Source 1",
			SourceType:       "api",
			ReliabilityScore: 0.95,
			LastUpdated:      time.Now(),
			Metadata:         make(map[string]interface{}),
		}

		result, err := validator.ValidateDataSource(ctx, source)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.Equal(t, 1.0, result.ValidationScore)
		assert.Empty(t, result.Errors)
	})

	t.Run("missing required fields", func(t *testing.T) {
		source := &DataSourceMetadata{
			ReliabilityScore: 0.95,
			Metadata:         make(map[string]interface{}),
		}

		result, err := validator.ValidateDataSource(ctx, source)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsValid)
		assert.Less(t, result.ValidationScore, 1.0)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("invalid reliability score", func(t *testing.T) {
		source := &DataSourceMetadata{
			SourceID:         "test_source_1",
			SourceName:       "Test Source 1",
			SourceType:       "api",
			ReliabilityScore: 1.5, // Invalid: > 1.0
			LastUpdated:      time.Now(),
			Metadata:         make(map[string]interface{}),
		}

		result, err := validator.ValidateDataSource(ctx, source)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsValid)
		assert.Less(t, result.ValidationScore, 1.0)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("nil data source", func(t *testing.T) {
		_, err := validator.ValidateDataSource(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestMetadataValidator_ValidateConfidence(t *testing.T) {
	logger := zap.NewNop()
	validator := NewMetadataValidator(logger, nil)

	ctx := context.Background()

	t.Run("valid confidence", func(t *testing.T) {
		confidence := &ConfidenceMetadata{
			OverallConfidence: 0.85,
			ConfidenceLevel:   ConfidenceLevelHigh,
			ComponentScores:   make(map[string]float64),
			Factors: []ConfidenceFactor{
				{
					FactorName:   "data_quality",
					FactorValue:  0.9,
					FactorWeight: 0.3,
					Description:  "High quality data",
					Impact:       "positive",
					Confidence:   0.9,
				},
			},
			CalculatedAt: time.Now(),
			Metadata:     make(map[string]interface{}),
		}

		result, err := validator.ValidateConfidence(ctx, confidence)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsValid)
		assert.Equal(t, 1.0, result.ValidationScore)
		assert.Empty(t, result.Errors)
	})

	t.Run("invalid overall confidence", func(t *testing.T) {
		confidence := &ConfidenceMetadata{
			OverallConfidence: 1.5, // Invalid: > 1.0
			ConfidenceLevel:   ConfidenceLevelHigh,
			ComponentScores:   make(map[string]float64),
			Factors:           []ConfidenceFactor{},
			CalculatedAt:      time.Now(),
			Metadata:          make(map[string]interface{}),
		}

		result, err := validator.ValidateConfidence(ctx, confidence)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsValid)
		assert.Less(t, result.ValidationScore, 1.0)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("invalid factor values", func(t *testing.T) {
		confidence := &ConfidenceMetadata{
			OverallConfidence: 0.85,
			ConfidenceLevel:   ConfidenceLevelHigh,
			ComponentScores:   make(map[string]float64),
			Factors: []ConfidenceFactor{
				{
					FactorName:   "data_quality",
					FactorValue:  1.5, // Invalid: > 1.0
					FactorWeight: 0.3,
					Description:  "High quality data",
					Impact:       "positive",
					Confidence:   0.9,
				},
			},
			CalculatedAt: time.Now(),
			Metadata:     make(map[string]interface{}),
		}

		result, err := validator.ValidateConfidence(ctx, confidence)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.IsValid)
		assert.Less(t, result.ValidationScore, 1.0)
		assert.NotEmpty(t, result.Errors)
	})

	t.Run("nil confidence", func(t *testing.T) {
		_, err := validator.ValidateConfidence(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestMetadataVersionManager_RegisterVersion(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataVersionManager(logger, nil)

	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		version := &MetadataVersion{
			Version:       "4.0",
			SchemaVersion: "4.0.0",
			CreatedAt:     time.Now(),
			Changes:       []VersionChange{},
			Compatibility: CompatibilityInfo{
				BackwardCompatible: true,
				ForwardCompatible:  false,
				SupportedVersions:  []string{"1.0", "2.0", "3.0", "4.0"},
				DeprecatedVersions: []string{},
				RemovedVersions:    []string{},
			},
			Metadata: make(map[string]interface{}),
		}

		err := manager.RegisterVersion(ctx, version)
		require.NoError(t, err)

		// Verify it was registered
		retrieved, err := manager.GetVersion(ctx, "4.0")
		require.NoError(t, err)
		assert.Equal(t, version.Version, retrieved.Version)
		assert.Equal(t, version.SchemaVersion, retrieved.SchemaVersion)
	})

	t.Run("nil version", func(t *testing.T) {
		err := manager.RegisterVersion(ctx, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})

	t.Run("empty version identifier", func(t *testing.T) {
		version := &MetadataVersion{
			SchemaVersion: "4.0.0",
			CreatedAt:     time.Now(),
			Changes:       []VersionChange{},
			Compatibility: CompatibilityInfo{},
			Metadata:      make(map[string]interface{}),
		}

		err := manager.RegisterVersion(ctx, version)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be empty")
	})

	t.Run("duplicate version", func(t *testing.T) {
		version := &MetadataVersion{
			Version:       "5.0",
			SchemaVersion: "5.0.0",
			CreatedAt:     time.Now(),
			Changes:       []VersionChange{},
			Compatibility: CompatibilityInfo{},
			Metadata:      make(map[string]interface{}),
		}

		err := manager.RegisterVersion(ctx, version)
		require.NoError(t, err)

		// Try to register again
		err = manager.RegisterVersion(ctx, version)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already registered")
	})
}

func TestMetadataVersionManager_CheckCompatibility(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataVersionManager(logger, nil)

	ctx := context.Background()

	t.Run("compatible versions", func(t *testing.T) {
		result, err := manager.CheckCompatibility(ctx, "1.0", "2.0")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Compatible)
		assert.Equal(t, "1.0", result.FromVersion)
		assert.Equal(t, "2.0", result.ToVersion)
		assert.Empty(t, result.Issues)
	})

	t.Run("same version", func(t *testing.T) {
		result, err := manager.CheckCompatibility(ctx, "1.0", "1.0")
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Compatible)
		assert.Empty(t, result.Issues)
	})

	t.Run("non-existent source version", func(t *testing.T) {
		_, err := manager.CheckCompatibility(ctx, "non_existent", "2.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("non-existent target version", func(t *testing.T) {
		_, err := manager.CheckCompatibility(ctx, "1.0", "non_existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestMetadataVersionManager_MigrateMetadata(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataVersionManager(logger, nil)

	ctx := context.Background()

	t.Run("successful migration", func(t *testing.T) {
		metadata := &ResponseMetadata{
			RequestID:      "test_request_1",
			Timestamp:      time.Now(),
			APIVersion:     "1.0",
			ProcessingTime: 100 * time.Millisecond,
			DataSources:    []DataSourceMetadata{},
			Metadata:       make(map[string]interface{}),
		}

		migrated, err := manager.MigrateMetadata(ctx, metadata, "2.0")
		require.NoError(t, err)
		assert.NotNil(t, migrated)
		assert.Equal(t, "2.0", migrated.APIVersion)
		assert.Equal(t, metadata.RequestID, migrated.RequestID)
		assert.Equal(t, metadata.Timestamp, migrated.Timestamp)
	})

	t.Run("incompatible versions", func(t *testing.T) {
		metadata := &ResponseMetadata{
			RequestID:      "test_request_1",
			Timestamp:      time.Now(),
			APIVersion:     "1.0",
			ProcessingTime: 100 * time.Millisecond,
			DataSources:    []DataSourceMetadata{},
			Metadata:       make(map[string]interface{}),
		}

		// Try to migrate to a non-existent version
		_, err := manager.MigrateMetadata(ctx, metadata, "non_existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("nil metadata", func(t *testing.T) {
		_, err := manager.MigrateMetadata(ctx, nil, "2.0")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be nil")
	})
}

func TestMetadataVersionManager_GetCurrentVersion(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataVersionManager(logger, nil)

	ctx := context.Background()

	version := manager.GetCurrentVersion(ctx)
	assert.Equal(t, "3.0", version)
}

func TestMetadataVersionManager_GetDefaultVersion(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataVersionManager(logger, nil)

	ctx := context.Background()

	version := manager.GetDefaultVersion(ctx)
	assert.Equal(t, "3.0", version)
}

func TestMetadataVersionManager_GetSupportedVersions(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataVersionManager(logger, nil)

	ctx := context.Background()

	versions, err := manager.GetSupportedVersions(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, versions)
	assert.Contains(t, versions, "1.0")
	assert.Contains(t, versions, "2.0")
	assert.Contains(t, versions, "3.0")
}

func TestMetadataVersionManager_DeprecateVersion(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataVersionManager(logger, nil)

	ctx := context.Background()

	t.Run("successful deprecation", func(t *testing.T) {
		err := manager.DeprecateVersion(ctx, "1.0", "Version 1.0 is deprecated")
		require.NoError(t, err)

		// Verify deprecation
		version, err := manager.GetVersion(ctx, "1.0")
		require.NoError(t, err)
		assert.NotNil(t, version.DeprecatedAt)
	})

	t.Run("non-existent version", func(t *testing.T) {
		err := manager.DeprecateVersion(ctx, "non_existent", "Test deprecation")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestMetadataVersionManager_RemoveVersion(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataVersionManager(logger, nil)

	ctx := context.Background()

	t.Run("successful removal", func(t *testing.T) {
		// Create a version manager with shorter deprecation period for testing
		testConfig := &VersioningConfig{
			CurrentVersion:       "3.0",
			DefaultVersion:       "3.0",
			MinSupportedVersion:  "1.0",
			EnableAutoVersioning: true,
			EnableDeprecation:    true,
			DeprecationPeriod:    1 * time.Millisecond, // Very short for testing
			EnableAutoMigration:  true,
			MigrationTimeout:     30 * time.Second,
			StrictVersioning:     false,
			ValidateOnLoad:       true,
		}
		testManager := NewMetadataVersionManager(logger, testConfig)

		// First deprecate the version
		err := testManager.DeprecateVersion(ctx, "2.0", "Version 2.0 is deprecated")
		require.NoError(t, err)

		// Wait a bit for deprecation period to pass
		time.Sleep(2 * time.Millisecond)

		// Then remove it
		err = testManager.RemoveVersion(ctx, "2.0", "Version 2.0 is removed")
		require.NoError(t, err)

		// Verify removal
		version, err := testManager.GetVersion(ctx, "2.0")
		require.NoError(t, err)
		assert.NotNil(t, version.RemovedAt)
	})

	t.Run("non-existent version", func(t *testing.T) {
		err := manager.RemoveVersion(ctx, "non_existent", "Test removal")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("version not deprecated", func(t *testing.T) {
		err := manager.RemoveVersion(ctx, "3.0", "Test removal")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be deprecated")
	})
}

func TestMetadataManager_Integration(t *testing.T) {
	logger := zap.NewNop()
	manager := NewMetadataManager(logger, nil)
	validator := NewMetadataValidator(logger, nil)
	versionManager := NewMetadataVersionManager(logger, nil)

	ctx := context.Background()

	t.Run("full metadata lifecycle", func(t *testing.T) {
		// 1. Create response metadata
		metadata, err := manager.CreateResponseMetadata(ctx, "test_request_1")
		require.NoError(t, err)
		assert.NotNil(t, metadata)

		// 2. Add data source
		source := &DataSourceMetadata{
			SourceID:         "test_source_1",
			SourceName:       "Test Source 1",
			SourceType:       "api",
			ReliabilityScore: 0.95,
			LastUpdated:      time.Now(),
			Metadata:         make(map[string]interface{}),
		}
		err = manager.AddDataSource(ctx, source)
		require.NoError(t, err)

		// 3. Calculate confidence
		factors := []ConfidenceFactor{
			{
				FactorName:   "data_quality",
				FactorValue:  0.9,
				FactorWeight: 0.3,
				Description:  "High quality data",
				Impact:       "positive",
				Confidence:   0.9,
			},
		}
		confidence, err := manager.CalculateConfidence(ctx, factors)
		require.NoError(t, err)
		assert.NotNil(t, confidence)

		// 4. Update metadata with data source and confidence
		metadata.DataSources = append(metadata.DataSources, *source)
		metadata.Confidence = confidence
		metadata.ProcessingTime = 150 * time.Millisecond

		err = manager.UpdateResponseMetadata(ctx, metadata)
		require.NoError(t, err)

		// 5. Validate metadata
		validationResult, err := validator.ValidateMetadata(ctx, metadata)
		require.NoError(t, err)
		assert.NotNil(t, validationResult)
		assert.True(t, validationResult.IsValid)

		// 6. Check version compatibility
		compatibility, err := versionManager.CheckCompatibility(ctx, "1.0", "3.0")
		require.NoError(t, err)
		assert.NotNil(t, compatibility)
		assert.True(t, compatibility.Compatible)

		// 7. Migrate metadata to newer version
		metadata.APIVersion = "1.0"
		migrated, err := versionManager.MigrateMetadata(ctx, metadata, "3.0")
		require.NoError(t, err)
		assert.NotNil(t, migrated)
		assert.Equal(t, "3.0", migrated.APIVersion)
	})

	t.Run("metadata quality assessment", func(t *testing.T) {
		// Create metadata with quality issues
		metadata := &ResponseMetadata{
			RequestID:      "test_request_2",
			Timestamp:      time.Now(),
			APIVersion:     "3.0",
			ProcessingTime: 5 * time.Second, // Slow processing
			DataSources: []DataSourceMetadata{
				{
					SourceID:         "low_quality_source",
					SourceName:       "Low Quality Source",
					SourceType:       "api",
					ReliabilityScore: 0.3, // Low reliability
					LastUpdated:      time.Now(),
					Metadata:         make(map[string]interface{}),
				},
			},
			Confidence: &ConfidenceMetadata{
				OverallConfidence: 0.4, // Low confidence
				ConfidenceLevel:   ConfidenceLevelLow,
				ComponentScores:   make(map[string]float64),
				Factors:           []ConfidenceFactor{},
				CalculatedAt:      time.Now(),
				Metadata:          make(map[string]interface{}),
			},
			Metadata: make(map[string]interface{}),
		}

		// Assess quality
		quality, err := manager.AssessQuality(ctx, metadata)
		require.NoError(t, err)
		assert.NotNil(t, quality)
		assert.Less(t, quality.OverallQuality, 1.0)
		assert.Equal(t, "low", quality.QualityLevel)
	})
}
