package multi_site_aggregation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewSiteConsistencyValidator(t *testing.T) {
	logger := zap.NewNop()
	dataStore := NewInMemoryDataStore()
	locationStore := NewInMemoryLocationStore()

	// Test with default config
	validator := NewSiteConsistencyValidator(dataStore, locationStore, logger, nil)
	assert.NotNil(t, validator)
	assert.Equal(t, dataStore, validator.dataStore)
	assert.Equal(t, locationStore, validator.locationStore)
	assert.Equal(t, logger, validator.logger)
	assert.NotNil(t, validator.config)

	// Test with custom config
	customConfig := &ConsistencyValidationConfig{
		MinConsistencyScore: 0.8,
		ValidationTimeout:   60 * time.Second,
	}
	validator = NewSiteConsistencyValidator(dataStore, locationStore, logger, customConfig)
	assert.Equal(t, customConfig, validator.config)
}

func TestDefaultConsistencyValidationConfig(t *testing.T) {
	config := DefaultConsistencyValidationConfig()
	assert.NotNil(t, config)
	assert.Equal(t, 0.7, config.MinConsistencyScore)
	assert.Equal(t, 0.9, config.HighConsistencyThreshold)
	assert.Equal(t, 0.7, config.MediumConsistencyThreshold)
	assert.Equal(t, 0.5, config.LowConsistencyThreshold)
	assert.Equal(t, 5, config.MaxFieldVariations)
	assert.Equal(t, 30*time.Second, config.ValidationTimeout)
	assert.True(t, config.EnableTrendAnalysis)
	assert.True(t, config.EnableRecommendations)
}

func TestSiteConsistencyValidator_ValidateConsistency(t *testing.T) {
	logger := zap.NewNop()
	dataStore := NewInMemoryDataStore()
	locationStore := NewInMemoryLocationStore()
	validator := NewSiteConsistencyValidator(dataStore, locationStore, logger, nil)

	ctx := context.Background()
	businessID := "business-123"

	// Test with no locations
	_, err := validator.ValidateConsistency(ctx, businessID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no locations found")

	// Add locations
	location1 := BusinessLocation{
		ID:         "location-1",
		BusinessID: businessID,
		URL:        "https://example1.com",
		Domain:     "example1.com",
		Region:     "us",
		Language:   "en",
		Country:    "United States",
		IsPrimary:  true,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	location2 := BusinessLocation{
		ID:         "location-2",
		BusinessID: businessID,
		URL:        "https://example2.com",
		Domain:     "example2.com",
		Region:     "us",
		Language:   "en",
		Country:    "United States",
		IsPrimary:  false,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = locationStore.SaveLocation(ctx, &location1)
	require.NoError(t, err)
	err = locationStore.SaveLocation(ctx, &location2)
	require.NoError(t, err)

	// Add site data with consistent values
	siteData1 := SiteData{
		ID:         "site-1",
		LocationID: "location-1",
		BusinessID: businessID,
		DataType:   "contact_info",
		ExtractedData: map[string]interface{}{
			"business_name": "Sample Business",
			"phone":         "+1-555-123-4567",
			"email":         "contact@samplebusiness.com",
		},
		ConfidenceScore:  0.85,
		ExtractionMethod: "contact_page_scraping",
		LastExtracted:    time.Now(),
		DataQuality:      0.8,
		IsValid:          true,
		Metadata:         make(map[string]interface{}),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	siteData2 := SiteData{
		ID:         "site-2",
		LocationID: "location-2",
		BusinessID: businessID,
		DataType:   "contact_info",
		ExtractedData: map[string]interface{}{
			"business_name": "Sample Business",
			"phone":         "+1-555-123-4567",
			"email":         "contact@samplebusiness.com",
		},
		ConfidenceScore:  0.82,
		ExtractionMethod: "contact_page_scraping",
		LastExtracted:    time.Now(),
		DataQuality:      0.78,
		IsValid:          true,
		Metadata:         make(map[string]interface{}),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err = dataStore.SaveSiteData(ctx, &siteData1)
	require.NoError(t, err)
	err = dataStore.SaveSiteData(ctx, &siteData2)
	require.NoError(t, err)

	// Test consistency validation
	result, err := validator.ValidateConsistency(ctx, businessID)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, businessID, result.BusinessID)
	assert.Equal(t, 2, result.TotalSites)
	assert.Greater(t, result.OverallScore, 0.8) // Should be high consistency
	assert.Equal(t, ConsistencyLevelHigh, result.ConsistencyLevel)
	assert.Len(t, result.FieldResults, 3) // business_name, phone, email
	assert.Len(t, result.Issues, 0)       // No issues for consistent data
}

func TestSiteConsistencyValidator_ValidateFieldConsistency(t *testing.T) {
	logger := zap.NewNop()
	dataStore := NewInMemoryDataStore()
	locationStore := NewInMemoryLocationStore()
	validator := NewSiteConsistencyValidator(dataStore, locationStore, logger, nil)

	ctx := context.Background()
	businessID := "business-789"

	// Add site data
	siteData1 := SiteData{
		ID:         "site-5",
		LocationID: "location-5",
		BusinessID: businessID,
		DataType:   "contact_info",
		ExtractedData: map[string]interface{}{
			"business_name": "Sample Business",
			"phone":         "+1-555-123-4567",
		},
		ConfidenceScore:  0.85,
		ExtractionMethod: "contact_page_scraping",
		LastExtracted:    time.Now(),
		DataQuality:      0.8,
		IsValid:          true,
		Metadata:         make(map[string]interface{}),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	siteData2 := SiteData{
		ID:         "site-6",
		LocationID: "location-6",
		BusinessID: businessID,
		DataType:   "contact_info",
		ExtractedData: map[string]interface{}{
			"business_name": "Sample Business",
			"phone":         "+1-555-123-4567",
		},
		ConfidenceScore:  0.82,
		ExtractionMethod: "contact_page_scraping",
		LastExtracted:    time.Now(),
		DataQuality:      0.78,
		IsValid:          true,
		Metadata:         make(map[string]interface{}),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	err := dataStore.SaveSiteData(ctx, &siteData1)
	require.NoError(t, err)
	err = dataStore.SaveSiteData(ctx, &siteData2)
	require.NoError(t, err)

	// Test field consistency validation
	result, err := validator.ValidateFieldConsistency(ctx, businessID, "business_name")
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "business_name", result.FieldName)
	assert.Equal(t, 1.0, result.ConsistencyScore) // Perfect consistency
	assert.Equal(t, ConsistencyLevelHigh, result.ConsistencyLevel)
	assert.True(t, result.IsConsistent)
	assert.Len(t, result.Values, 2)
	assert.Len(t, result.Issues, 0)

	// Test with non-existent field
	_, err = validator.ValidateFieldConsistency(ctx, businessID, "non_existent_field")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found in any site data")
}

func TestSiteConsistencyValidator_ValidateDataIntegrity(t *testing.T) {
	logger := zap.NewNop()
	dataStore := NewInMemoryDataStore()
	locationStore := NewInMemoryLocationStore()
	validator := NewSiteConsistencyValidator(dataStore, locationStore, logger, nil)

	ctx := context.Background()

	// Test with valid aggregated data
	validAggregatedData := &AggregatedBusinessData{
		BusinessID: "business-integrity",
		AggregatedData: map[string]interface{}{
			"contact_info": map[string]interface{}{
				"business_name": "Integrity Business",
				"phone":         "+1-555-333-4444",
				"email":         "contact@integritybusiness.com",
			},
			"business_details": map[string]interface{}{
				"business_name": "Integrity Business",
				"industry":      "Technology",
			},
		},
		ConsistencyIssues:    []DataConsistencyIssue{},
		DataConsistencyScore: 0.95,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	result, err := validator.ValidateDataIntegrity(ctx, validAggregatedData)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.IsValid)
	assert.Greater(t, result.IntegrityScore, 0.9)
	assert.Len(t, result.Issues, 0)
	assert.Len(t, result.Warnings, 0)

	// Test with invalid aggregated data
	invalidAggregatedData := &AggregatedBusinessData{
		BusinessID: "", // Missing business ID
		AggregatedData: map[string]interface{}{
			"contact_info": map[string]interface{}{
				// Missing required fields
			},
		},
		ConsistencyIssues: []DataConsistencyIssue{
			{
				ID:          "issue-1",
				FieldName:   "business_name",
				DataType:    "contact_info",
				IssueType:   "conflict",
				Severity:    "high",
				Description: "Conflicting business names",
			},
		},
		DataConsistencyScore: 0.5,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	result, err = validator.ValidateDataIntegrity(ctx, invalidAggregatedData)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.IsValid)
	assert.Less(t, result.IntegrityScore, 0.5)
	assert.Greater(t, len(result.Issues), 0)
	assert.Greater(t, len(result.Warnings), 0)
}

func TestSiteConsistencyValidator_HelperMethods(t *testing.T) {
	logger := zap.NewNop()
	dataStore := NewInMemoryDataStore()
	locationStore := NewInMemoryLocationStore()
	validator := NewSiteConsistencyValidator(dataStore, locationStore, logger, nil)

	// Test valueToString
	assert.Equal(t, "test", validator.valueToString("test"))
	assert.Equal(t, "123", validator.valueToString(123))
	assert.Equal(t, "a,b,c", validator.valueToString([]string{"a", "b", "c"}))
	assert.Equal(t, "nil", validator.valueToString(nil))

	// Test determineConsistencyLevel
	assert.Equal(t, ConsistencyLevelHigh, validator.determineConsistencyLevel(0.95))
	assert.Equal(t, ConsistencyLevelMedium, validator.determineConsistencyLevel(0.75))
	assert.Equal(t, ConsistencyLevelLow, validator.determineConsistencyLevel(0.55))
	assert.Equal(t, ConsistencyLevelPoor, validator.determineConsistencyLevel(0.3))

	// Test determineIssueSeverity
	assert.Equal(t, "high", validator.determineIssueSeverity(0.2))
	assert.Equal(t, "medium", validator.determineIssueSeverity(0.4))
	assert.Equal(t, "low", validator.determineIssueSeverity(0.7))

	// Test removeDuplicateStrings
	strings := []string{"a", "b", "a", "c", "b"}
	result := validator.removeDuplicateStrings(strings)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "a")
	assert.Contains(t, result, "b")
	assert.Contains(t, result, "c")

	// Test getLocationURL
	url := validator.getLocationURL("location-123")
	assert.Contains(t, url, "location-123")
}

func TestSiteConsistencyValidator_CalculateFieldConsistencyScore(t *testing.T) {
	logger := zap.NewNop()
	dataStore := NewInMemoryDataStore()
	locationStore := NewInMemoryLocationStore()
	validator := NewSiteConsistencyValidator(dataStore, locationStore, logger, nil)

	// Test with empty values
	score := validator.calculateFieldConsistencyScore([]FieldValue{})
	assert.Equal(t, 0.0, score)

	// Test with single value
	singleValue := []FieldValue{
		{
			Value:      "test",
			SiteID:     "site-1",
			Confidence: 0.8,
		},
	}
	score = validator.calculateFieldConsistencyScore(singleValue)
	assert.Equal(t, 1.0, score)

	// Test with identical values
	identicalValues := []FieldValue{
		{
			Value:      "test",
			SiteID:     "site-1",
			Confidence: 0.8,
		},
		{
			Value:      "test",
			SiteID:     "site-2",
			Confidence: 0.9,
		},
	}
	score = validator.calculateFieldConsistencyScore(identicalValues)
	assert.Equal(t, 1.0, score)

	// Test with different values
	differentValues := []FieldValue{
		{
			Value:      "test1",
			SiteID:     "site-1",
			Confidence: 0.8,
		},
		{
			Value:      "test2",
			SiteID:     "site-2",
			Confidence: 0.9,
		},
	}
	score = validator.calculateFieldConsistencyScore(differentValues)
	assert.Less(t, score, 1.0)
	assert.Greater(t, score, 0.0)
}

func TestSiteConsistencyValidator_CalculateOverallConsistencyScore(t *testing.T) {
	logger := zap.NewNop()
	dataStore := NewInMemoryDataStore()
	locationStore := NewInMemoryLocationStore()
	validator := NewSiteConsistencyValidator(dataStore, locationStore, logger, nil)

	// Test with empty results
	score := validator.calculateOverallConsistencyScore([]FieldConsistencyResult{})
	assert.Equal(t, 0.0, score)

	// Test with single result
	singleResult := []FieldConsistencyResult{
		{
			FieldName:        "test",
			ConsistencyScore: 0.8,
			Values: []FieldValue{
				{Value: "test", SiteID: "site-1"},
			},
		},
	}
	score = validator.calculateOverallConsistencyScore(singleResult)
	assert.Equal(t, 0.8, score)

	// Test with multiple results
	multipleResults := []FieldConsistencyResult{
		{
			FieldName:        "field1",
			ConsistencyScore: 0.8,
			Values: []FieldValue{
				{Value: "test1", SiteID: "site-1"},
				{Value: "test1", SiteID: "site-2"},
			},
		},
		{
			FieldName:        "field2",
			ConsistencyScore: 0.6,
			Values: []FieldValue{
				{Value: "test2", SiteID: "site-1"},
			},
		},
	}
	score = validator.calculateOverallConsistencyScore(multipleResults)
	// Should be weighted average: (0.8*2 + 0.6*1) / 3 = 0.733...
	assert.Greater(t, score, 0.7)
	assert.Less(t, score, 0.8)
}

func TestSiteConsistencyValidator_GenerateFieldConsistencyIssues(t *testing.T) {
	logger := zap.NewNop()
	dataStore := NewInMemoryDataStore()
	locationStore := NewInMemoryLocationStore()
	validator := NewSiteConsistencyValidator(dataStore, locationStore, logger, nil)

	// Test with too many variations
	manyVariations := []FieldValue{
		{Value: "value1", SiteID: "site-1"},
		{Value: "value2", SiteID: "site-2"},
		{Value: "value3", SiteID: "site-3"},
		{Value: "value4", SiteID: "site-4"},
		{Value: "value5", SiteID: "site-5"},
		{Value: "value6", SiteID: "site-6"},
	}
	issues := validator.generateFieldConsistencyIssues("test_field", manyVariations)
	assert.Greater(t, len(issues), 0)
	assert.Equal(t, "too_many_variations", issues[0].IssueType)

	// Test with empty values
	emptyValues := []FieldValue{
		{Value: "", SiteID: "site-1"},
		{Value: nil, SiteID: "site-2"},
	}
	issues = validator.generateFieldConsistencyIssues("test_field", emptyValues)
	assert.Greater(t, len(issues), 0)
	assert.Equal(t, "empty_values", issues[0].IssueType)
}

func TestSiteConsistencyValidator_CalculateIntegrityScore(t *testing.T) {
	logger := zap.NewNop()
	dataStore := NewInMemoryDataStore()
	locationStore := NewInMemoryLocationStore()
	validator := NewSiteConsistencyValidator(dataStore, locationStore, logger, nil)

	// Test with no issues or warnings
	score := validator.calculateIntegrityScore([]string{}, []string{})
	assert.Equal(t, 1.0, score)

	// Test with warnings only
	score = validator.calculateIntegrityScore([]string{}, []string{"warning1", "warning2"})
	assert.Less(t, score, 1.0)
	assert.GreaterOrEqual(t, score, 0.5)

	// Test with issues only
	score = validator.calculateIntegrityScore([]string{"issue1", "issue2"}, []string{})
	assert.LessOrEqual(t, score, 0.5)
	// The score should be 0.0 for 2 issues with no warnings
	// Formula: 1.0 - (2*2 + 0*1) / (2*2) = 1.0 - 4/4 = 0.0

	// Test with both issues and warnings
	score = validator.calculateIntegrityScore([]string{"issue1"}, []string{"warning1"})
	assert.Less(t, score, 1.0)
	assert.Greater(t, score, 0.0)
}
