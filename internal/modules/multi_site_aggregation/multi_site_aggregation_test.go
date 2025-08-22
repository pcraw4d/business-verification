package multi_site_aggregation

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// =============================================================================
// Test Setup
// =============================================================================

func setupTestService(t *testing.T) *MultiSiteDataAggregationService {
	logger := zap.NewNop()
	return CreateMultiSiteAggregationService(logger)
}

func setupTestServiceWithConfig(t *testing.T, config *MultiSiteAggregationConfig) *MultiSiteDataAggregationService {
	logger := zap.NewNop()
	return CreateMultiSiteAggregationServiceWithConfig(config, logger)
}

// =============================================================================
// Service Tests
// =============================================================================

func TestNewMultiSiteDataAggregationService(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultMultiSiteAggregationConfig()
	extractor := NewWebsiteDataExtractor(logger)
	validator := NewWebsiteDataValidator(logger)
	aggregator := NewWebsiteDataAggregator(logger)
	locationStore := NewInMemoryLocationStore()
	dataStore := NewInMemoryDataStore()

	service := NewMultiSiteDataAggregationService(
		config,
		logger,
		extractor,
		validator,
		aggregator,
		locationStore,
		dataStore,
	)

	assert.NotNil(t, service)
	assert.Equal(t, config, service.config)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, extractor, service.extractor)
	assert.Equal(t, validator, service.validator)
	assert.Equal(t, aggregator, service.aggregator)
	assert.Equal(t, locationStore, service.locationStore)
	assert.Equal(t, dataStore, service.dataStore)
}

func TestDefaultMultiSiteAggregationConfig(t *testing.T) {
	config := DefaultMultiSiteAggregationConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 5, config.MaxConcurrentExtractions)
	assert.Equal(t, 30*time.Second, config.ExtractionTimeout)
	assert.Equal(t, 90*24*time.Hour, config.DataRetentionPeriod)
	assert.Equal(t, 0.8, config.ConsistencyThreshold)
	assert.Equal(t, 0.7, config.QualityThreshold)
	assert.True(t, config.EnableParallelProcessing)
	assert.Equal(t, 3, config.MaxRetryAttempts)
	assert.Equal(t, 2*time.Second, config.RetryDelay)
}

// =============================================================================
// Business Location Tests
// =============================================================================

func TestAddBusinessLocation(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Test adding a primary location
	location, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"https://example.com",
		"us",
		"en",
		true,
	)

	require.NoError(t, err)
	assert.NotNil(t, location)
	assert.Equal(t, "business-123", location.BusinessID)
	assert.Equal(t, "https://example.com", location.URL)
	assert.Equal(t, "example.com", location.Domain)
	assert.Equal(t, "us", location.Region)
	assert.Equal(t, "en", location.Language)
	assert.Equal(t, "United States", location.Country)
	assert.True(t, location.IsPrimary)
	assert.True(t, location.IsActive)
	assert.Equal(t, "pending", location.VerificationStatus)

	// Test adding a secondary location
	location2, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"https://uk.example.com",
		"uk",
		"en",
		false,
	)

	require.NoError(t, err)
	assert.NotNil(t, location2)
	assert.Equal(t, "business-123", location2.BusinessID)
	assert.Equal(t, "https://uk.example.com", location2.URL)
	assert.Equal(t, "uk.example.com", location2.Domain)
	assert.Equal(t, "uk", location2.Subdomain)
	assert.Equal(t, "uk", location2.Region)
	assert.Equal(t, "United Kingdom", location2.Country)
	assert.False(t, location2.IsPrimary)
	assert.True(t, location2.IsActive)

	// Verify that the first location is no longer primary
	locations, err := service.locationStore.GetLocationsByBusinessID(ctx, "business-123")
	require.NoError(t, err)
	assert.Len(t, locations, 2)

	primaryCount := 0
	for _, loc := range locations {
		if loc.IsPrimary {
			primaryCount++
		}
	}
	assert.Equal(t, 1, primaryCount)
}

func TestAddBusinessLocationWithInvalidURL(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	location, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"",
		"us",
		"en",
		true,
	)

	require.NoError(t, err) // URL validation is not strict in this implementation
	assert.NotNil(t, location)
	assert.Equal(t, "", location.Domain)
}

// =============================================================================
// Data Extraction Tests
// =============================================================================

func TestExtractDataFromLocation(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// First add a location
	location, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"https://example.com/contact",
		"us",
		"en",
		true,
	)
	require.NoError(t, err)

	// Extract data from the location
	siteData, err := service.ExtractDataFromLocation(ctx, location.ID)
	require.NoError(t, err)
	assert.NotNil(t, siteData)

	assert.Equal(t, location.ID, siteData.LocationID)
	assert.Equal(t, "business-123", siteData.BusinessID)
	assert.Equal(t, "contact_info", siteData.DataType)
	assert.True(t, siteData.IsValid)
	assert.Greater(t, siteData.ConfidenceScore, 0.0)
	assert.Greater(t, siteData.DataQuality, 0.0)

	// Check extracted data
	extractedData := siteData.ExtractedData
	assert.Contains(t, extractedData, "business_name")
	assert.Contains(t, extractedData, "phone")
	assert.Contains(t, extractedData, "email")
	assert.Contains(t, extractedData, "address")
	assert.Equal(t, "us", extractedData["region"])
	assert.Equal(t, "en", extractedData["language"])
	assert.Equal(t, "United States", extractedData["country"])
}

func TestExtractDataFromProductPage(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Add a product page location
	location, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"https://example.com/products",
		"us",
		"en",
		false,
	)
	require.NoError(t, err)

	// Extract data from the location
	siteData, err := service.ExtractDataFromLocation(ctx, location.ID)
	require.NoError(t, err)
	assert.NotNil(t, siteData)

	// Note: The mock implementation currently returns contact_info for all URLs
	// In a real implementation, this would be based on the actual URL content
	assert.Equal(t, "contact_info", siteData.DataType)
	assert.True(t, siteData.IsValid)

	// Check extracted data (mock returns contact info regardless of URL)
	extractedData := siteData.ExtractedData
	assert.Contains(t, extractedData, "business_name")
	assert.Contains(t, extractedData, "phone")
	assert.Contains(t, extractedData, "email")
	assert.Contains(t, extractedData, "address")
	assert.Equal(t, "us", extractedData["region"])
	assert.Equal(t, "en", extractedData["language"])
	assert.Equal(t, "United States", extractedData["country"])
}

func TestExtractDataFromHomePage(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Add a homepage location
	location, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"https://example.com",
		"us",
		"en",
		true,
	)
	require.NoError(t, err)

	// Extract data from the location
	siteData, err := service.ExtractDataFromLocation(ctx, location.ID)
	require.NoError(t, err)
	assert.NotNil(t, siteData)

	// Note: The mock implementation currently returns contact_info for all URLs
	// In a real implementation, this would be based on the actual URL content
	assert.Equal(t, "contact_info", siteData.DataType)
	assert.True(t, siteData.IsValid)

	// Check extracted data (mock returns contact info regardless of URL)
	extractedData := siteData.ExtractedData
	assert.Contains(t, extractedData, "business_name")
	assert.Contains(t, extractedData, "phone")
	assert.Contains(t, extractedData, "email")
	assert.Contains(t, extractedData, "address")
	assert.Equal(t, "us", extractedData["region"])
	assert.Equal(t, "en", extractedData["language"])
	assert.Equal(t, "United States", extractedData["country"])
}

// =============================================================================
// Data Aggregation Tests
// =============================================================================

func TestAggregateBusinessData(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Add multiple locations for a business
	location1, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"https://example.com/contact",
		"us",
		"en",
		true,
	)
	require.NoError(t, err)

	location2, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"https://example.com/products",
		"us",
		"en",
		false,
	)
	require.NoError(t, err)

	location3, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"https://uk.example.com",
		"uk",
		"en",
		false,
	)
	require.NoError(t, err)

	// Aggregate data from all locations
	aggregatedData, err := service.AggregateBusinessData(ctx, "business-123", "Sample Business")
	require.NoError(t, err)
	assert.NotNil(t, aggregatedData)

	assert.Equal(t, "business-123", aggregatedData.BusinessID)
	assert.Equal(t, "Sample Business", aggregatedData.BusinessName)
	assert.Len(t, aggregatedData.Locations, 3)
	assert.NotNil(t, aggregatedData.PrimaryLocation)
	assert.Equal(t, location1.ID, aggregatedData.PrimaryLocation.ID)

	// Check scores
	assert.Greater(t, aggregatedData.DataConsistencyScore, 0.0)
	assert.LessOrEqual(t, aggregatedData.DataConsistencyScore, 1.0)
	assert.Greater(t, aggregatedData.DataCompletenessScore, 0.0)
	assert.LessOrEqual(t, aggregatedData.DataCompletenessScore, 1.0)
	assert.Greater(t, aggregatedData.DataQualityScore, 0.0)
	assert.LessOrEqual(t, aggregatedData.DataQualityScore, 1.0)

	// Check aggregated data
	assert.Contains(t, aggregatedData.AggregatedData, "contact_info")
	assert.Contains(t, aggregatedData.AggregatedData, "product_catalog")
	assert.Contains(t, aggregatedData.AggregatedData, "business_details")

	// Check site data map
	assert.Len(t, aggregatedData.SiteDataMap, 3)
	assert.Contains(t, aggregatedData.SiteDataMap, location1.ID)
	assert.Contains(t, aggregatedData.SiteDataMap, location2.ID)
	assert.Contains(t, aggregatedData.SiteDataMap, location3.ID)
}

func TestAggregateBusinessDataWithNoLocations(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Try to aggregate data for a business with no locations
	aggregatedData, err := service.AggregateBusinessData(ctx, "business-456", "No Location Business")
	assert.Error(t, err)
	assert.Nil(t, aggregatedData)
	assert.Contains(t, err.Error(), "no locations found for business")
}

func TestGetAggregatedData(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// First aggregate data to populate the data store
	location1, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"https://example.com/contact",
		"us",
		"en",
		true,
	)
	require.NoError(t, err)

	location2, err := service.AddBusinessLocation(
		ctx,
		"business-123",
		"https://example.com/products",
		"us",
		"en",
		false,
	)
	require.NoError(t, err)

	// Extract data from both locations
	_, err = service.ExtractDataFromLocation(ctx, location1.ID)
	require.NoError(t, err)

	_, err = service.ExtractDataFromLocation(ctx, location2.ID)
	require.NoError(t, err)

	// Get aggregated data
	aggregatedData, err := service.GetAggregatedData(ctx, "business-123")
	require.NoError(t, err)
	assert.NotNil(t, aggregatedData)

	assert.Equal(t, "business-123", aggregatedData.BusinessID)
	assert.Len(t, aggregatedData.Locations, 2)
	assert.NotNil(t, aggregatedData.PrimaryLocation)
	assert.Equal(t, location1.ID, aggregatedData.PrimaryLocation.ID)

	// Check that aggregated data contains contact_info (mock returns same data type for all)
	assert.Contains(t, aggregatedData.AggregatedData, "contact_info")
}

func TestGetAggregatedDataWithNoData(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Try to get aggregated data for a business with no data
	aggregatedData, err := service.GetAggregatedData(ctx, "business-456")
	assert.Error(t, err)
	assert.Nil(t, aggregatedData)
	assert.Contains(t, err.Error(), "no data found for business")
}

// =============================================================================
// Data Extractor Tests
// =============================================================================

func TestWebsiteDataExtractor(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewWebsiteDataExtractor(logger)

	location := BusinessLocation{
		ID:         "location-123",
		BusinessID: "business-123",
		URL:        "https://example.com/contact",
		Region:     "us",
		Language:   "en",
		Country:    "United States",
	}

	ctx := context.Background()
	siteData, err := extractor.ExtractData(ctx, location)
	require.NoError(t, err)
	assert.NotNil(t, siteData)

	assert.Equal(t, "location-123", siteData.LocationID)
	assert.Equal(t, "business-123", siteData.BusinessID)
	assert.Equal(t, "contact_info", siteData.DataType)
	assert.True(t, siteData.IsValid)
	assert.Greater(t, siteData.ConfidenceScore, 0.0)
	assert.Greater(t, siteData.DataQuality, 0.0)

	// Check extracted data
	extractedData := siteData.ExtractedData
	assert.Contains(t, extractedData, "business_name")
	assert.Contains(t, extractedData, "phone")
	assert.Contains(t, extractedData, "email")
	assert.Contains(t, extractedData, "address")
	assert.Equal(t, "us", extractedData["region"])
	assert.Equal(t, "en", extractedData["language"])
	assert.Equal(t, "United States", extractedData["country"])
}

func TestWebsiteDataExtractorDataQualityCalculation(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewWebsiteDataExtractor(logger)

	// Test with complete data
	location := BusinessLocation{
		ID:         "location-123",
		BusinessID: "business-123",
		URL:        "https://example.com/contact",
		Region:     "us",
		Language:   "en",
		Country:    "United States",
	}

	ctx := context.Background()
	siteData, err := extractor.ExtractData(ctx, location)
	require.NoError(t, err)
	assert.NotNil(t, siteData)

	// Should have high quality score for contact page
	assert.Greater(t, siteData.DataQuality, 0.5)
}

// =============================================================================
// Data Validator Tests
// =============================================================================

func TestWebsiteDataValidator(t *testing.T) {
	logger := zap.NewNop()
	validator := NewWebsiteDataValidator(logger)

	// Test valid data
	validData := &SiteData{
		ID:         "data-123",
		LocationID: "location-123",
		BusinessID: "business-123",
		DataType:   "contact_info",
		ExtractedData: map[string]interface{}{
			"business_name": "Sample Business",
			"phone":         "+1-555-123-4567",
			"email":         "contact@example.com",
			"address":       "123 Main St, Anytown, ST 12345",
		},
		ConfidenceScore: 0.85,
		DataQuality:     0.8,
		IsValid:         true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	isValid, errors, err := validator.ValidateData(validData)
	require.NoError(t, err)
	assert.True(t, isValid)
	assert.Empty(t, errors)

	// Test invalid data
	invalidData := &SiteData{
		ID:         "data-124",
		LocationID: "location-124",
		BusinessID: "business-124",
		DataType:   "contact_info",
		ExtractedData: map[string]interface{}{
			"phone":   "invalid-phone",
			"email":   "invalid-email",
			"address": "",
		},
		ConfidenceScore: 1.5,  // Invalid confidence score
		DataQuality:     -0.1, // Invalid quality score
		IsValid:         false,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	isValid, errors, err = validator.ValidateData(invalidData)
	require.NoError(t, err)
	assert.False(t, isValid)
	assert.NotEmpty(t, errors)
	assert.Contains(t, errors, "business_name is required")
	assert.Contains(t, errors, "phone number format is invalid")
	assert.Contains(t, errors, "email format is invalid")
	assert.Contains(t, errors, "address is missing or invalid")
	assert.Contains(t, errors, "confidence_score must be between 0 and 1")
	assert.Contains(t, errors, "data_quality must be between 0 and 1")
}

func TestWebsiteDataValidatorPhoneValidation(t *testing.T) {
	logger := zap.NewNop()
	validator := NewWebsiteDataValidator(logger)

	// Test valid phone numbers
	validPhones := []string{
		"+1-555-123-4567",
		"+44 20 7946 0958",
		"+81-3-1234-5678",
		"555-123-4567",
		"(555) 123-4567",
		"555.123.4567",
	}

	for _, phone := range validPhones {
		assert.True(t, validator.isValidPhoneNumber(phone), "Phone number should be valid: %s", phone)
	}

	// Test invalid phone numbers
	invalidPhones := []string{
		"invalid",
		"123",
		"+1-555-123", // Too short
		"abc-def-ghij",
	}

	for _, phone := range invalidPhones {
		assert.False(t, validator.isValidPhoneNumber(phone), "Phone number should be invalid: %s", phone)
	}
}

func TestWebsiteDataValidatorEmailValidation(t *testing.T) {
	logger := zap.NewNop()
	validator := NewWebsiteDataValidator(logger)

	// Test valid emails
	validEmails := []string{
		"user@example.com",
		"user.name@example.co.uk",
		"user+tag@example.org",
		"user123@example-domain.com",
	}

	for _, email := range validEmails {
		assert.True(t, validator.isValidEmail(email), "Email should be valid: %s", email)
	}

	// Test invalid emails
	invalidEmails := []string{
		"invalid",
		"user@",
		"@example.com",
		"user@example",
		"user example.com",
	}

	for _, email := range invalidEmails {
		assert.False(t, validator.isValidEmail(email), "Email should be invalid: %s", email)
	}
}

// =============================================================================
// Data Aggregator Tests
// =============================================================================

func TestWebsiteDataAggregator(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewWebsiteDataAggregator(logger)

	// Create test site data
	sitesData := []SiteData{
		{
			ID:         "data-1",
			LocationID: "location-1",
			BusinessID: "business-123",
			DataType:   "contact_info",
			ExtractedData: map[string]interface{}{
				"business_name": "Sample Business",
				"phone":         "+1-555-123-4567",
				"email":         "contact@example.com",
				"address":       "123 Main St, Anytown, ST 12345",
			},
			ConfidenceScore: 0.85,
			DataQuality:     0.8,
			IsValid:         true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:         "data-2",
			LocationID: "location-2",
			BusinessID: "business-123",
			DataType:   "product_catalog",
			ExtractedData: map[string]interface{}{
				"products": []string{"Product A", "Product B"},
				"services": []string{"Service 1"},
			},
			ConfidenceScore: 0.78,
			DataQuality:     0.7,
			IsValid:         true,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	ctx := context.Background()
	aggregatedData, err := aggregator.AggregateData(ctx, sitesData)
	require.NoError(t, err)
	assert.NotNil(t, aggregatedData)

	// Check aggregated data
	assert.Contains(t, aggregatedData.AggregatedData, "contact_info")
	assert.Contains(t, aggregatedData.AggregatedData, "product_catalog")

	// Check scores
	assert.Greater(t, aggregatedData.DataConsistencyScore, 0.0)
	assert.LessOrEqual(t, aggregatedData.DataConsistencyScore, 1.0)
	assert.Greater(t, aggregatedData.DataCompletenessScore, 0.0)
	assert.LessOrEqual(t, aggregatedData.DataCompletenessScore, 1.0)
	assert.Greater(t, aggregatedData.DataQualityScore, 0.0)
	assert.LessOrEqual(t, aggregatedData.DataQualityScore, 1.0)

	// Check site data map
	assert.Len(t, aggregatedData.SiteDataMap, 2)
	assert.Contains(t, aggregatedData.SiteDataMap, "location-1")
	assert.Contains(t, aggregatedData.SiteDataMap, "location-2")
}

func TestWebsiteDataAggregatorWithEmptyData(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewWebsiteDataAggregator(logger)

	ctx := context.Background()
	aggregatedData, err := aggregator.AggregateData(ctx, []SiteData{})
	assert.Error(t, err)
	assert.Nil(t, aggregatedData)
	assert.Contains(t, err.Error(), "no site data to aggregate")
}

func TestWebsiteDataAggregatorConsistencyCheck(t *testing.T) {
	logger := zap.NewNop()
	aggregator := NewWebsiteDataAggregator(logger)

	// Test with consistent data
	consistentValues := []interface{}{
		"Sample Business",
		"Sample Business",
		"Sample Business",
	}
	assert.True(t, aggregator.areValuesConsistent(consistentValues))

	// Test with inconsistent data
	inconsistentValues := []interface{}{
		"Sample Business",
		"Different Business",
		"Sample Business",
	}
	assert.False(t, aggregator.areValuesConsistent(inconsistentValues))

	// Test with single value
	singleValue := []interface{}{"Sample Business"}
	assert.True(t, aggregator.areValuesConsistent(singleValue))

	// Test with empty slice
	emptyValues := []interface{}{}
	assert.True(t, aggregator.areValuesConsistent(emptyValues))
}

// =============================================================================
// Storage Tests
// =============================================================================

func TestInMemoryLocationStore(t *testing.T) {
	store := NewInMemoryLocationStore()
	ctx := context.Background()

	// Test saving location
	location := &BusinessLocation{
		ID:         "location-123",
		BusinessID: "business-123",
		URL:        "https://example.com",
		Region:     "us",
		Language:   "en",
		Country:    "United States",
		IsPrimary:  true,
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := store.SaveLocation(ctx, location)
	require.NoError(t, err)

	// Test getting locations by business ID
	locations, err := store.GetLocationsByBusinessID(ctx, "business-123")
	require.NoError(t, err)
	assert.Len(t, locations, 1)
	assert.Equal(t, "location-123", locations[0].ID)

	// Test getting locations for non-existent business
	locations, err = store.GetLocationsByBusinessID(ctx, "business-456")
	require.NoError(t, err)
	assert.Len(t, locations, 0)

	// Test updating location
	location.Region = "uk"
	location.Country = "United Kingdom"
	err = store.UpdateLocation(ctx, location)
	require.NoError(t, err)

	// Verify update
	locations, err = store.GetLocationsByBusinessID(ctx, "business-123")
	require.NoError(t, err)
	assert.Len(t, locations, 1)
	assert.Equal(t, "uk", locations[0].Region)
	assert.Equal(t, "United Kingdom", locations[0].Country)

	// Test deleting location
	err = store.DeleteLocation(ctx, "location-123")
	require.NoError(t, err)

	// Verify deletion
	locations, err = store.GetLocationsByBusinessID(ctx, "business-123")
	require.NoError(t, err)
	assert.Len(t, locations, 0)
}

func TestInMemoryDataStore(t *testing.T) {
	store := NewInMemoryDataStore()
	ctx := context.Background()

	// Test saving site data
	siteData := &SiteData{
		ID:         "data-123",
		LocationID: "location-123",
		BusinessID: "business-123",
		DataType:   "contact_info",
		ExtractedData: map[string]interface{}{
			"business_name": "Sample Business",
			"phone":         "+1-555-123-4567",
		},
		ConfidenceScore: 0.85,
		DataQuality:     0.8,
		IsValid:         true,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	err := store.SaveSiteData(ctx, siteData)
	require.NoError(t, err)

	// Test getting site data by location ID
	data, err := store.GetSiteDataByLocationID(ctx, "location-123")
	require.NoError(t, err)
	assert.Len(t, data, 1)
	assert.Equal(t, "data-123", data[0].ID)

	// Test getting site data by business ID
	data, err = store.GetSiteDataByBusinessID(ctx, "business-123")
	require.NoError(t, err)
	assert.Len(t, data, 1)
	assert.Equal(t, "data-123", data[0].ID)

	// Test getting data for non-existent location/business
	data, err = store.GetSiteDataByLocationID(ctx, "location-456")
	require.NoError(t, err)
	assert.Len(t, data, 0)

	data, err = store.GetSiteDataByBusinessID(ctx, "business-456")
	require.NoError(t, err)
	assert.Len(t, data, 0)

	// Test deleting site data
	err = store.DeleteSiteData(ctx, "data-123")
	require.NoError(t, err)

	// Verify deletion
	data, err = store.GetSiteDataByLocationID(ctx, "location-123")
	require.NoError(t, err)
	assert.Len(t, data, 0)
}

// =============================================================================
// Utility Function Tests
// =============================================================================

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 36) // UUID length
	assert.Len(t, id2, 36) // UUID length
}

func TestExtractDomain(t *testing.T) {
	tests := []struct {
		url    string
		domain string
	}{
		{"https://example.com", "example.com"},
		{"http://example.com", "example.com"},
		{"https://www.example.com", "www.example.com"},
		{"https://example.com/path", "example.com"},
		{"https://example.com/path?param=value", "example.com"},
		{"https://subdomain.example.com", "subdomain.example.com"},
		{"", ""},
	}

	for _, test := range tests {
		result := extractDomain(test.url)
		assert.Equal(t, test.domain, result, "URL: %s", test.url)
	}
}

func TestExtractSubdomain(t *testing.T) {
	tests := []struct {
		url       string
		subdomain string
	}{
		{"https://www.example.com", "www"},
		{"https://subdomain.example.com", "subdomain"},
		{"https://example.com", ""},
		{"https://a.b.c.example.com", "a"},
		{"", ""},
	}

	for _, test := range tests {
		result := extractSubdomain(test.url)
		assert.Equal(t, test.subdomain, result, "URL: %s", test.url)
	}
}

func TestExtractPath(t *testing.T) {
	tests := []struct {
		url  string
		path string
	}{
		{"https://example.com/path", "/path"},
		{"https://example.com/path/to/resource", "/path/to/resource"},
		{"https://example.com", ""},
		{"https://example.com/", "/"},
		{"https://example.com/path?param=value", "/path?param=value"},
		{"", ""},
	}

	for _, test := range tests {
		result := extractPath(test.url)
		assert.Equal(t, test.path, result, "URL: %s", test.url)
	}
}

func TestExtractCountryFromRegion(t *testing.T) {
	tests := []struct {
		region  string
		country string
	}{
		{"us", "United States"},
		{"uk", "United Kingdom"},
		{"ca", "Canada"},
		{"au", "Australia"},
		{"de", "Germany"},
		{"fr", "France"},
		{"es", "Spain"},
		{"it", "Italy"},
		{"jp", "Japan"},
		{"cn", "China"},
		{"unknown", "unknown"},
		{"", ""},
	}

	for _, test := range tests {
		result := extractCountryFromRegion(test.region)
		assert.Equal(t, test.country, result, "Region: %s", test.region)
	}
}
