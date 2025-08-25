package industry_codes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestMetadataManager_Initialize(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	mm := NewMetadataManager(db, logger)

	ctx := context.Background()
	err := mm.Initialize(ctx)
	require.NoError(t, err)

	// Verify tables were created by checking if we can query them
	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM code_descriptions").Scan(&count)
	require.NoError(t, err)

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM code_metadata").Scan(&count)
	require.NoError(t, err)

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM code_relationships").Scan(&count)
	require.NoError(t, err)

	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM code_crosswalks").Scan(&count)
	require.NoError(t, err)
}

func TestMetadataManager_SaveAndGetCodeDescription(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	mm := NewMetadataManager(db, logger)

	ctx := context.Background()
	err := mm.Initialize(ctx)
	require.NoError(t, err)

	// Create a test code first
	icdb := NewIndustryCodeDatabase(db, logger)
	err = icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-code-1",
		Code:        "5411",
		Type:        CodeTypeSIC,
		Description: "Legal Services",
		Category:    "Professional Services",
		Confidence:  0.9,
	}
	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Test saving and retrieving code description
	desc := &CodeDescription{
		ID:               "desc-1",
		CodeID:           "test-code-1",
		ShortDescription: "Legal services and law firms",
		LongDescription:  "This category includes establishments primarily engaged in providing legal services such as legal advice and representation in civil and criminal cases, and other legal services.",
		Examples:         []string{"Law firms", "Legal consulting", "Attorney services"},
		Exclusions:       []string{"Court reporting services", "Process serving"},
		Notes:            "Updated for 2024 classification standards",
		Source:           "SIC Manual 2024",
		Version:          "2024.1",
	}

	err = mm.SaveCodeDescription(ctx, desc)
	require.NoError(t, err)

	// Retrieve the description
	retrieved, err := mm.GetCodeDescription(ctx, "test-code-1", "2024.1")
	require.NoError(t, err)
	assert.Equal(t, desc.ID, retrieved.ID)
	assert.Equal(t, desc.CodeID, retrieved.CodeID)
	assert.Equal(t, desc.ShortDescription, retrieved.ShortDescription)
	assert.Equal(t, desc.LongDescription, retrieved.LongDescription)
	assert.Equal(t, desc.Examples, retrieved.Examples)
	assert.Equal(t, desc.Exclusions, retrieved.Exclusions)
	assert.Equal(t, desc.Notes, retrieved.Notes)
	assert.Equal(t, desc.Source, retrieved.Source)
	assert.Equal(t, desc.Version, retrieved.Version)
}

func TestMetadataManager_GetLatestCodeDescription(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	mm := NewMetadataManager(db, logger)

	ctx := context.Background()
	err := mm.Initialize(ctx)
	require.NoError(t, err)

	// Create a test code first
	icdb := NewIndustryCodeDatabase(db, logger)
	err = icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-code-2",
		Code:        "5412",
		Type:        CodeTypeSIC,
		Description: "Accounting Services",
		Category:    "Professional Services",
		Confidence:  0.9,
	}
	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Save multiple versions
	desc1 := &CodeDescription{
		ID:               "desc-2-v1",
		CodeID:           "test-code-2",
		ShortDescription: "Accounting services v1",
		Source:           "SIC Manual 2023",
		Version:          "2023.1",
	}
	err = mm.SaveCodeDescription(ctx, desc1)
	require.NoError(t, err)

	desc2 := &CodeDescription{
		ID:               "desc-2-v2",
		CodeID:           "test-code-2",
		ShortDescription: "Accounting services v2",
		Source:           "SIC Manual 2024",
		Version:          "2024.1",
	}
	err = mm.SaveCodeDescription(ctx, desc2)
	require.NoError(t, err)

	// Get latest description
	latest, err := mm.GetLatestCodeDescription(ctx, "test-code-2")
	require.NoError(t, err)
	assert.Equal(t, "2024.1", latest.Version)
	assert.Equal(t, "Accounting services v2", latest.ShortDescription)
}

func TestMetadataManager_SaveAndGetCodeMetadata(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	mm := NewMetadataManager(db, logger)

	ctx := context.Background()
	err := mm.Initialize(ctx)
	require.NoError(t, err)

	// Create a test code first
	icdb := NewIndustryCodeDatabase(db, logger)
	err = icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-code-3",
		Code:        "5413",
		Type:        CodeTypeSIC,
		Description: "Architectural Services",
		Category:    "Professional Services",
		Confidence:  0.9,
	}
	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Test saving and retrieving code metadata
	metadata := &CodeMetadata{
		ID:               "meta-1",
		CodeID:           "test-code-3",
		Version:          "2024.1",
		EffectiveDate:    time.Now(),
		Source:           "SIC Manual 2024",
		SourceURL:        "https://www.census.gov/sic/",
		UpdateFrequency:  "Annual",
		DataQuality:      "High",
		Confidence:       0.95,
		UsageCount:       150,
		Popularity:       "High",
		Tags:             []string{"professional", "services", "architecture"},
		CustomFields:     map[string]string{"industry_group": "54", "subsector": "541"},
		ValidationStatus: "validated",
		ValidationNotes:  "Manually reviewed and approved",
	}

	err = mm.SaveCodeMetadata(ctx, metadata)
	require.NoError(t, err)

	// Retrieve the metadata
	retrieved, err := mm.GetCodeMetadata(ctx, "test-code-3", "2024.1")
	require.NoError(t, err)
	assert.Equal(t, metadata.ID, retrieved.ID)
	assert.Equal(t, metadata.CodeID, retrieved.CodeID)
	assert.Equal(t, metadata.Version, retrieved.Version)
	assert.Equal(t, metadata.Source, retrieved.Source)
	assert.Equal(t, metadata.SourceURL, retrieved.SourceURL)
	assert.Equal(t, metadata.UpdateFrequency, retrieved.UpdateFrequency)
	assert.Equal(t, metadata.DataQuality, retrieved.DataQuality)
	assert.Equal(t, metadata.Confidence, retrieved.Confidence)
	assert.Equal(t, metadata.UsageCount, retrieved.UsageCount)
	assert.Equal(t, metadata.Popularity, retrieved.Popularity)
	assert.Equal(t, metadata.Tags, retrieved.Tags)
	assert.Equal(t, metadata.CustomFields, retrieved.CustomFields)
	assert.Equal(t, metadata.ValidationStatus, retrieved.ValidationStatus)
	assert.Equal(t, metadata.ValidationNotes, retrieved.ValidationNotes)
}

func TestMetadataManager_SaveAndGetCodeRelationships(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	mm := NewMetadataManager(db, logger)

	ctx := context.Background()
	err := mm.Initialize(ctx)
	require.NoError(t, err)

	// Create test codes first
	icdb := NewIndustryCodeDatabase(db, logger)
	err = icdb.Initialize(ctx)
	require.NoError(t, err)

	codes := []*IndustryCode{
		{
			ID:          "parent-code",
			Code:        "54",
			Type:        CodeTypeSIC,
			Description: "Professional, Scientific, and Technical Services",
			Category:    "Professional Services",
			Confidence:  0.9,
		},
		{
			ID:          "child-code",
			Code:        "541",
			Type:        CodeTypeSIC,
			Description: "Professional, Scientific, and Technical Services",
			Category:    "Professional Services",
			Confidence:  0.9,
		},
	}

	for _, code := range codes {
		err = icdb.InsertCode(ctx, code)
		require.NoError(t, err)
	}

	// Test saving and retrieving code relationships
	relationship := &CodeRelationship{
		ID:               "rel-1",
		SourceCodeID:     "parent-code",
		TargetCodeID:     "child-code",
		RelationshipType: RelationshipTypeParentChild,
		Confidence:       0.95,
		Notes:            "Parent-child relationship in SIC hierarchy",
	}

	err = mm.SaveCodeRelationship(ctx, relationship)
	require.NoError(t, err)

	// Retrieve relationships
	relationships, err := mm.GetCodeRelationships(ctx, "parent-code", "")
	require.NoError(t, err)
	assert.Len(t, relationships, 1)
	assert.Equal(t, relationship.ID, relationships[0].ID)
	assert.Equal(t, relationship.SourceCodeID, relationships[0].SourceCodeID)
	assert.Equal(t, relationship.TargetCodeID, relationships[0].TargetCodeID)
	assert.Equal(t, relationship.RelationshipType, relationships[0].RelationshipType)
	assert.Equal(t, relationship.Confidence, relationships[0].Confidence)
	assert.Equal(t, relationship.Notes, relationships[0].Notes)

	// Test filtering by relationship type
	parentChildRels, err := mm.GetCodeRelationships(ctx, "parent-code", RelationshipTypeParentChild)
	require.NoError(t, err)
	assert.Len(t, parentChildRels, 1)

	relatedRels, err := mm.GetCodeRelationships(ctx, "parent-code", RelationshipTypeRelated)
	require.NoError(t, err)
	assert.Len(t, relatedRels, 0)
}

func TestMetadataManager_SaveAndGetCodeCrosswalks(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	mm := NewMetadataManager(db, logger)

	ctx := context.Background()
	err := mm.Initialize(ctx)
	require.NoError(t, err)

	// Test saving and retrieving code crosswalks
	crosswalk := &CodeCrosswalk{
		ID:          "cross-1",
		SourceCode:  "5411",
		SourceType:  CodeTypeSIC,
		TargetCode:  "541110",
		TargetType:  CodeTypeNAICS,
		MappingType: MappingTypeExact,
		Confidence:  0.98,
		Direction:   "bidirectional",
		Notes:       "Exact mapping between SIC and NAICS",
	}

	err = mm.SaveCodeCrosswalk(ctx, crosswalk)
	require.NoError(t, err)

	// Retrieve crosswalks
	crosswalks, err := mm.GetCodeCrosswalks(ctx, "5411", CodeTypeSIC, CodeTypeNAICS)
	require.NoError(t, err)
	assert.Len(t, crosswalks, 1)
	assert.Equal(t, crosswalk.ID, crosswalks[0].ID)
	assert.Equal(t, crosswalk.SourceCode, crosswalks[0].SourceCode)
	assert.Equal(t, crosswalk.SourceType, crosswalks[0].SourceType)
	assert.Equal(t, crosswalk.TargetCode, crosswalks[0].TargetCode)
	assert.Equal(t, crosswalk.TargetType, crosswalks[0].TargetType)
	assert.Equal(t, crosswalk.MappingType, crosswalks[0].MappingType)
	assert.Equal(t, crosswalk.Confidence, crosswalks[0].Confidence)
	assert.Equal(t, crosswalk.Direction, crosswalks[0].Direction)
	assert.Equal(t, crosswalk.Notes, crosswalks[0].Notes)
}

func TestMetadataManager_UpdateUsageCount(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	mm := NewMetadataManager(db, logger)

	ctx := context.Background()
	err := mm.Initialize(ctx)
	require.NoError(t, err)

	// Create a test code and metadata first
	icdb := NewIndustryCodeDatabase(db, logger)
	err = icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-code-4",
		Code:        "5414",
		Type:        CodeTypeSIC,
		Description: "Specialized Design Services",
		Category:    "Professional Services",
		Confidence:  0.9,
	}
	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	metadata := &CodeMetadata{
		ID:            "meta-2",
		CodeID:        "test-code-4",
		Version:       "2024.1",
		EffectiveDate: time.Now(),
		Source:        "SIC Manual 2024",
		UsageCount:    100,
	}
	err = mm.SaveCodeMetadata(ctx, metadata)
	require.NoError(t, err)

	// Update usage count
	err = mm.UpdateUsageCount(ctx, "test-code-4", 25)
	require.NoError(t, err)

	// Verify the update
	retrieved, err := mm.GetCodeMetadata(ctx, "test-code-4", "2024.1")
	require.NoError(t, err)
	assert.Equal(t, int64(125), retrieved.UsageCount)
}

func TestMetadataManager_ValidateCodeMetadata(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	mm := NewMetadataManager(db, logger)

	ctx := context.Background()
	err := mm.Initialize(ctx)
	require.NoError(t, err)

	// Create a test code first
	icdb := NewIndustryCodeDatabase(db, logger)
	err = icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-code-5",
		Code:        "5415",
		Type:        CodeTypeSIC,
		Description: "Computer Systems Design",
		Category:    "Professional Services",
		Confidence:  0.9,
	}
	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Create complete metadata and description
	metadata := &CodeMetadata{
		ID:            "meta-3",
		CodeID:        "test-code-5",
		Version:       "latest",
		EffectiveDate: time.Now(),
		Source:        "SIC Manual 2024",
		DataQuality:   "High",
		Confidence:    0.95,
	}
	err = mm.SaveCodeMetadata(ctx, metadata)
	require.NoError(t, err)

	desc := &CodeDescription{
		ID:               "desc-3",
		CodeID:           "test-code-5",
		ShortDescription: "Computer systems design services",
		LongDescription:  "Establishments primarily engaged in providing computer systems design services.",
		Examples:         []string{"Software development", "System integration"},
		Source:           "SIC Manual 2024",
		Version:          "latest",
	}
	err = mm.SaveCodeDescription(ctx, desc)
	require.NoError(t, err)

	// Validate the metadata
	result, err := mm.ValidateCodeMetadata(ctx, "test-code-5")
	require.NoError(t, err)
	assert.Equal(t, "test-code-5", result.CodeID)
	assert.Greater(t, result.OverallScore, 0.8)
	assert.Equal(t, "valid", result.Status)
	assert.Empty(t, result.Issues)
	assert.NotEmpty(t, result.Recommendations) // Should have recommendation about relationships
}

func TestMetadataManager_GetMetadataStats(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	mm := NewMetadataManager(db, logger)

	ctx := context.Background()
	err := mm.Initialize(ctx)
	require.NoError(t, err)

	// Create some test data
	icdb := NewIndustryCodeDatabase(db, logger)
	err = icdb.Initialize(ctx)
	require.NoError(t, err)

	testCode := &IndustryCode{
		ID:          "test-code-6",
		Code:        "5416",
		Type:        CodeTypeSIC,
		Description: "Management Consulting",
		Category:    "Professional Services",
		Confidence:  0.9,
	}
	err = icdb.InsertCode(ctx, testCode)
	require.NoError(t, err)

	// Add some metadata
	metadata := &CodeMetadata{
		ID:            "meta-4",
		CodeID:        "test-code-6",
		Version:       "2024.1",
		EffectiveDate: time.Now(),
		Source:        "SIC Manual 2024",
		Confidence:    0.9,
	}
	err = mm.SaveCodeMetadata(ctx, metadata)
	require.NoError(t, err)

	// Get stats
	stats, err := mm.GetMetadataStats(ctx)
	require.NoError(t, err)

	assert.Contains(t, stats, "total_descriptions")
	assert.Contains(t, stats, "total_metadata")
	assert.Contains(t, stats, "total_relationships")
	assert.Contains(t, stats, "total_crosswalks")
	assert.Contains(t, stats, "average_confidence")

	assert.Equal(t, 0, stats["total_descriptions"])
	assert.Equal(t, 1, stats["total_metadata"])
	assert.Equal(t, 0, stats["total_relationships"])
	assert.Equal(t, 0, stats["total_crosswalks"])
	assert.Equal(t, 0.9, stats["average_confidence"])
}

func TestMetadataManager_ErrorHandling(t *testing.T) {
	db, cleanup := setupTestDatabase(t)
	defer cleanup()

	logger := zaptest.NewLogger(t)
	mm := NewMetadataManager(db, logger)

	ctx := context.Background()
	err := mm.Initialize(ctx)
	require.NoError(t, err)

	// Test getting non-existent description
	_, err = mm.GetCodeDescription(ctx, "non-existent", "2024.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code description not found")

	// Test getting non-existent metadata
	_, err = mm.GetCodeMetadata(ctx, "non-existent", "2024.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code metadata not found")

	// Test getting latest description for non-existent code
	_, err = mm.GetLatestCodeDescription(ctx, "non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "code description not found")

	// Test updating usage count for non-existent code
	err = mm.UpdateUsageCount(ctx, "non-existent", 1)
	assert.NoError(t, err) // This should not error, just log a warning
}
