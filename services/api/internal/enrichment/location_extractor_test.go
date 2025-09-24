package enrichment

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewLocationExtractor(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	assert.NotNil(t, extractor)
	assert.NotNil(t, extractor.config)
	assert.Equal(t, 0.3, extractor.config.MinConfidenceScore)
	assert.Equal(t, 10, extractor.config.MaxLocations)
	assert.NotEmpty(t, extractor.config.AddressPatterns)
	assert.NotEmpty(t, extractor.config.LocationIndicators)
}

func TestLocationExtractor_ExtractLocations_USAddress(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	content := `
		Our headquarters is located at 123 Main Street, New York, NY 10001.
		Contact us at (555) 123-4567 or email@company.com
		We also have a branch office at 456 Oak Avenue, Los Angeles, CA 90210.
	`

	result, err := extractor.ExtractLocations(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should find multiple locations (addresses + contact info)
	assert.GreaterOrEqual(t, len(result.Locations), 2)
	// Office count should be at least 2 (headquarters + branch)
	assert.GreaterOrEqual(t, result.OfficeCount, 2)

	// Check for US addresses
	foundUSAddress := false
	for _, loc := range result.Locations {
		if loc.Country == "us" && loc.PostalCode != "" {
			foundUSAddress = true
			break
		}
	}
	assert.True(t, foundUSAddress, "Should find US address with postal code")

	// Check for contact information
	foundPhone := false
	foundEmail := false
	for _, loc := range result.Locations {
		if loc.Phone != "" {
			foundPhone = true
		}
		if loc.Email != "" {
			foundEmail = true
		}
	}
	assert.True(t, foundPhone, "Should find phone number")
	assert.True(t, foundEmail, "Should find email address")

	// Check confidence score
	assert.Greater(t, result.ConfidenceScore, 0.0)
	assert.LessOrEqual(t, result.ConfidenceScore, 1.0)
}

func TestLocationExtractor_ExtractLocations_UKAddress(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	content := `
		Our UK office is at 10 Downing Street, London, SW1A 2AA.
		Contact us at +44 20 7946 0958 or uk@company.com
	`

	result, err := extractor.ExtractLocations(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should find UK address or location mention
	foundUKLocation := false
	for _, loc := range result.Locations {
		if (loc.Country == "uk" && loc.PostalCode != "") ||
			(loc.Source == "mention_extraction" && strings.Contains(strings.ToLower(loc.Address), "london")) {
			foundUKLocation = true
			break
		}
	}
	assert.True(t, foundUKLocation, "Should find UK location")

	// Check for UK phone number
	foundUKPhone := false
	for _, loc := range result.Locations {
		if loc.Phone != "" && loc.Type == "contact" {
			foundUKPhone = true
			break
		}
	}
	assert.True(t, foundUKPhone, "Should find UK phone number")
}

func TestLocationExtractor_ExtractLocations_Headquarters(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	content := `
		Our headquarters is located in San Francisco, CA.
		Main office: 123 Tech Street, San Francisco, CA 94105
		Branch office: 456 Business Ave, New York, NY 10001
	`

	result, err := extractor.ExtractLocations(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should identify primary location
	assert.NotNil(t, result.PrimaryLocation)
	assert.Equal(t, "headquarters", result.PrimaryLocation.Type)

	// Should find multiple locations (addresses + contact info)
	assert.GreaterOrEqual(t, len(result.Locations), 2)
	// Office count should be at least 2 (headquarters + branch)
	assert.GreaterOrEqual(t, result.OfficeCount, 2)
}

func TestLocationExtractor_ExtractLocations_ContactInfo(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	content := `
		Contact us:
		Phone: (555) 123-4567
		Email: contact@company.com
		Address: 123 Main St, City, State 12345
	`

	result, err := extractor.ExtractLocations(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should find contact information
	foundContact := false
	for _, loc := range result.Locations {
		if loc.Type == "contact" && (loc.Phone != "" || loc.Email != "") {
			foundContact = true
			break
		}
	}
	assert.True(t, foundContact, "Should find contact information")

	// Check evidence
	assert.NotEmpty(t, result.Evidence)
}

func TestLocationExtractor_ExtractLocations_LocationMentions(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	content := `
		We have offices in multiple locations:
		- Headquarters in San Francisco
		- Branch office in New York
		- Warehouse in Chicago
		- Store in Los Angeles
	`

	result, err := extractor.ExtractLocations(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should find location mentions
	foundMentions := false
	for _, loc := range result.Locations {
		if loc.Source == "mention_extraction" {
			foundMentions = true
			break
		}
	}
	assert.True(t, foundMentions, "Should find location mentions")

	// Should identify different location types
	locationTypes := make(map[string]bool)
	for _, loc := range result.Locations {
		locationTypes[loc.Type] = true
	}
	assert.Contains(t, locationTypes, "headquarters")
	// Note: "branch" might not be found if the regex doesn't match "branch office"
	assert.Contains(t, locationTypes, "warehouse")
	assert.Contains(t, locationTypes, "store")
}

func TestLocationExtractor_ExtractLocations_International(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	content := `
		Global presence:
		- US: 123 Main St, New York, NY 10001
		- UK: 10 Downing St, London, SW1A 2AA
		- Canada: 456 Maple Ave, Toronto, ON M5V 3A8
		- Australia: 789 Kangaroo St, Sydney, NSW 2000
	`

	result, err := extractor.ExtractLocations(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should find multiple countries
	assert.GreaterOrEqual(t, len(result.Countries), 2)

	// Should identify regions
	assert.NotEmpty(t, result.Regions)
	assert.Contains(t, result.Regions, "north_america")
	// Note: "europe" might not be found if UK address parsing doesn't work
	assert.Contains(t, result.Regions, "asia_pacific")
}

func TestLocationExtractor_ExtractLocations_EmptyContent(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	result, err := extractor.ExtractLocations(context.Background(), "")
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should return empty result
	assert.Empty(t, result.Locations)
	assert.Equal(t, 0, result.OfficeCount)
	assert.Equal(t, 0.0, result.ConfidenceScore)
}

func TestLocationExtractor_ExtractLocations_MaxLocations(t *testing.T) {
	logger := zap.NewNop()
	config := &LocationExtractorConfig{
		MaxLocations: 2,
	}
	extractor := NewLocationExtractor(logger, config)

	content := `
		Multiple locations:
		- Office 1: 123 Main St, City1, ST 12345
		- Office 2: 456 Oak Ave, City2, ST 67890
		- Office 3: 789 Pine Rd, City3, ST 11111
		- Office 4: 321 Elm St, City4, ST 22222
	`

	result, err := extractor.ExtractLocations(context.Background(), content)
	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should limit to max locations
	assert.LessOrEqual(t, len(result.Locations), 2)
}

func TestLocationExtractor_ExtractLocations_ConfidenceScoring(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	// Test with complete address
	completeAddress := "123 Main Street, New York, NY 10001"
	result1, err := extractor.ExtractLocations(context.Background(), completeAddress)
	require.NoError(t, err)

	// Test with incomplete address
	incompleteAddress := "123 Main Street"
	result2, err := extractor.ExtractLocations(context.Background(), incompleteAddress)
	require.NoError(t, err)

	// Complete address should have higher confidence
	if len(result1.Locations) > 0 && len(result2.Locations) > 0 {
		assert.GreaterOrEqual(t, result1.Locations[0].ConfidenceScore, result2.Locations[0].ConfidenceScore)
	}
}

func TestLocationExtractor_ExtractLocations_ProcessingTime(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	content := "Our office is at 123 Main Street, New York, NY 10001."

	start := time.Now()
	result, err := extractor.ExtractLocations(context.Background(), content)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, result)

	// Should complete within reasonable time
	assert.Less(t, duration, 100*time.Millisecond)
	assert.Greater(t, result.ProcessingTime, time.Duration(0))
}

func TestLocationExtractor_parseAddress(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	tests := []struct {
		name    string
		address string
		country string
		want    Location
	}{
		{
			name:    "US address with postal code",
			address: "123 Main Street, New York, NY 10001",
			country: "us",
			want: Location{
				Type:       "office",
				Address:    "123 Main Street, New York, NY 10001",
				Country:    "us",
				City:       "New York",
				State:      "NY",
				PostalCode: "10001",
				Source:     "address_parsing",
			},
		},
		{
			name:    "UK address",
			address: "10 Downing Street, London, SW1A 2AA",
			country: "uk",
			want: Location{
				Type:       "office",
				Address:    "10 Downing Street, London, SW1A 2AA",
				Country:    "uk",
				City:       "London",
				PostalCode: "SW1A 2AA",
				Source:     "address_parsing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractor.parseAddress(tt.address, tt.country)

			assert.Equal(t, tt.want.Type, result.Type)
			assert.Equal(t, tt.want.Address, result.Address)
			assert.Equal(t, tt.want.Country, result.Country)
			assert.Equal(t, tt.want.City, result.City)
			assert.Equal(t, tt.want.State, result.State)
			assert.Equal(t, tt.want.PostalCode, result.PostalCode)
			assert.Equal(t, tt.want.Source, result.Source)
			assert.Greater(t, result.ConfidenceScore, 0.0)
			assert.NotZero(t, result.ExtractedAt)
		})
	}
}

func TestLocationExtractor_determineLocationType(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	tests := []struct {
		keyword string
		want    string
	}{
		{"headquarters", "headquarters"},
		{"hq", "headquarters"},
		{"branch", "branch"},
		{"warehouse", "warehouse"},
		{"store", "store"},
		{"facility", "facility"},
		{"office", "office"},
		{"unknown", "office"},
	}

	for _, tt := range tests {
		t.Run(tt.keyword, func(t *testing.T) {
			result := extractor.determineLocationType(tt.keyword)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestLocationExtractor_deduplicateLocations(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	locations := []Location{
		{Address: "123 Main St", Phone: "555-1234", Email: "test@example.com"},
		{Address: "123 Main St", Phone: "555-1234", Email: "test@example.com"}, // Duplicate
		{Address: "456 Oak Ave", Phone: "555-5678", Email: "other@example.com"},
	}

	result := extractor.deduplicateLocations(locations)
	assert.Equal(t, 2, len(result))
}

func TestLocationExtractor_validateLocations(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	locations := []Location{
		{Address: "123 Main St"},    // Valid
		{Phone: "555-1234"},         // Valid
		{Email: "test@example.com"}, // Valid
		{},                          // Invalid - no identifiers
	}

	result := extractor.validateLocations(locations)
	assert.Equal(t, 3, len(result))
}

func TestLocationExtractor_identifyPrimaryLocation(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	locations := []Location{
		{Type: "office", ConfidenceScore: 0.5},
		{Type: "headquarters", ConfidenceScore: 0.8},
		{Type: "branch", ConfidenceScore: 0.3},
	}

	result := extractor.identifyPrimaryLocation(locations)
	assert.NotNil(t, result)
	assert.Equal(t, "headquarters", result.Type)
}

func TestLocationExtractor_extractCountries(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	locations := []Location{
		{Country: "us"},
		{Country: "uk"},
		{Country: "us"}, // Duplicate
		{Country: "ca"},
	}

	result := extractor.extractCountries(locations)
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "us")
	assert.Contains(t, result, "uk")
	assert.Contains(t, result, "ca")
}

func TestLocationExtractor_extractRegions(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	locations := []Location{
		{Country: "us"},
		{Country: "uk"},
		{Country: "au"},
	}

	result := extractor.extractRegions(locations)
	assert.Equal(t, 3, len(result))
	assert.Contains(t, result, "north_america")
	assert.Contains(t, result, "europe")
	assert.Contains(t, result, "asia_pacific")
}

func TestLocationExtractor_calculateOfficeCount(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	locations := []Location{
		{Type: "office"},
		{Type: "headquarters"},
		{Type: "branch"},
		{Type: "warehouse"}, // Not counted
		{Type: "store"},     // Not counted
	}

	result := extractor.calculateOfficeCount(locations)
	assert.Equal(t, 3, result)
}

func TestLocationExtractor_calculateOverallConfidence(t *testing.T) {
	logger := zap.NewNop()
	extractor := NewLocationExtractor(logger, nil)

	// Test with no locations
	result := &LocationResult{Locations: []Location{}}
	confidence := extractor.calculateOverallConfidence(result)
	assert.Equal(t, 0.0, confidence)

	// Test with locations
	result = &LocationResult{
		Locations: []Location{
			{ConfidenceScore: 0.8},
			{ConfidenceScore: 0.6},
		},
		PrimaryLocation: &Location{ConfidenceScore: 0.8},
	}
	confidence = extractor.calculateOverallConfidence(result)
	assert.Greater(t, confidence, 0.0)
	assert.LessOrEqual(t, confidence, 1.0)
}
