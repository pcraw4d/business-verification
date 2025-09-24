package enrichment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewGeographicValidator(t *testing.T) {
	tests := []struct {
		name   string
		config *GeographicValidatorConfig
		expect func(*testing.T, *GeographicValidator)
	}{
		{
			name:   "with nil config",
			config: nil,
			expect: func(t *testing.T, gv *GeographicValidator) {
				assert.NotNil(t, gv)
				assert.NotNil(t, gv.config)
				assert.NotEmpty(t, gv.config.CountryMappings)
				assert.NotEmpty(t, gv.config.ValidCountries)
			},
		},
		{
			name: "with custom config",
			config: &GeographicValidatorConfig{
				CountryMappings: map[string]string{"test": "Test Country"},
				ValidCountries:  map[string]bool{"Test Country": true},
			},
			expect: func(t *testing.T, gv *GeographicValidator) {
				assert.NotNil(t, gv)
				assert.Equal(t, "Test Country", gv.config.CountryMappings["test"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewGeographicValidator(zap.NewNop(), tt.config)
			tt.expect(t, validator)
		})
	}
}

func TestGeographicValidator_ValidateLocation(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name     string
		location Location
		expect   func(*testing.T, *StandardizedLocation, error)
	}{
		{
			name: "valid US location",
			location: Location{
				Address:    "123 Main Street",
				City:       "new york",
				State:      "ny",
				Country:    "United States",
				PostalCode: "10001",
				Phone:      "+1-555-123-4567",
			},
			expect: func(t *testing.T, result *StandardizedLocation, err error) {
				require.NoError(t, err)
				assert.Equal(t, "United States", result.Country)
				assert.Equal(t, "US", result.CountryCode)
				assert.Equal(t, "New York", result.City)
				assert.Equal(t, "New York", result.State)
				assert.Equal(t, "North America", result.Region)
				assert.True(t, result.IsValid)
				assert.Greater(t, result.ConfidenceScore, 0.8)
			},
		},
		{
			name: "location with country mapping",
			location: Location{
				Address: "456 Oak Avenue",
				City:    "london",
				Country: "uk", // Should map to "United Kingdom"
			},
			expect: func(t *testing.T, result *StandardizedLocation, err error) {
				require.NoError(t, err)
				assert.Equal(t, "United Kingdom", result.Country)
				assert.Equal(t, "GB", result.CountryCode)
				assert.Equal(t, "London", result.City)
				assert.Equal(t, "Europe", result.Region)
			},
		},
		{
			name: "location with invalid country",
			location: Location{
				Address: "789 Invalid Street",
				Country: "Invalid Country",
			},
			expect: func(t *testing.T, result *StandardizedLocation, err error) {
				require.NoError(t, err)
				assert.False(t, result.IsValid)
				assert.GreaterOrEqual(t, len(result.ValidationErrors), 1)

				// Check for country validation error
				foundCountryError := false
				for _, validationErr := range result.ValidationErrors {
					if validationErr.Field == "country" {
						foundCountryError = true
						assert.Equal(t, "error", validationErr.Severity)
						break
					}
				}
				assert.True(t, foundCountryError, "Should have country validation error")
			},
		},
		{
			name: "location with invalid postal code",
			location: Location{
				Address:    "123 Test Street",
				Country:    "United States",
				PostalCode: "invalid",
			},
			expect: func(t *testing.T, result *StandardizedLocation, err error) {
				require.NoError(t, err)
				assert.Equal(t, "United States", result.Country)

				// Should have postal code warning
				foundPostalError := false
				for _, validationErr := range result.ValidationErrors {
					if validationErr.Field == "postal_code" {
						foundPostalError = true
						assert.Equal(t, "warning", validationErr.Severity)
						break
					}
				}
				assert.True(t, foundPostalError, "Should have postal code validation warning")
			},
		},
		{
			name:     "empty location",
			location: Location{},
			expect: func(t *testing.T, result *StandardizedLocation, err error) {
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.LessOrEqual(t, result.ConfidenceScore, 0.5)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateLocation(context.Background(), tt.location)
			tt.expect(t, result, err)
		})
	}
}

func TestGeographicValidator_ValidateGeography_LocationResult(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	locationResult := &LocationResult{
		Locations: []Location{
			{
				Address: "123 Main St",
				City:    "New York",
				State:   "NY",
				Country: "United States",
			},
			{
				Address: "456 Oak Ave",
				City:    "London",
				Country: "United Kingdom",
			},
		},
		Countries: []string{"United States", "United Kingdom"},
	}

	result, err := validator.ValidateGeography(context.Background(), locationResult)
	require.NoError(t, err)
	require.NotNil(t, result)

	geography := result.StandardizedGeography
	assert.Len(t, geography.Locations, 2)
	assert.Contains(t, geography.Countries, "United States")
	assert.Contains(t, geography.Countries, "United Kingdom")
	assert.Contains(t, geography.CountryCodes, "US")
	assert.Contains(t, geography.CountryCodes, "GB")
	assert.Contains(t, geography.Regions, "North America")
	assert.Contains(t, geography.Regions, "Europe")
	assert.Greater(t, geography.ConfidenceScore, 0.0)
	assert.Greater(t, result.ProcessingTime, time.Duration(0))
}

func TestGeographicValidator_ValidateGeography_InternationalPresenceResult(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	internationalResult := &InternationalPresenceResult{
		Countries: []string{"United States", "Canada", "germany"},
		Regions:   []string{"North America", "Europe", "Invalid Region"},
	}

	result, err := validator.ValidateGeography(context.Background(), internationalResult)
	require.NoError(t, err)
	require.NotNil(t, result)

	geography := result.StandardizedGeography
	assert.Contains(t, geography.Countries, "United States")
	assert.Contains(t, geography.Countries, "Canada")
	assert.Contains(t, geography.Countries, "Germany")
	assert.Contains(t, geography.Regions, "North America")
	assert.Contains(t, geography.Regions, "Europe")

	// Should have validation error for invalid region
	assert.GreaterOrEqual(t, len(geography.ValidationErrors), 1)

	foundRegionError := false
	for _, validationErr := range geography.ValidationErrors {
		if validationErr.Field == "region" && validationErr.Value == "Invalid Region" {
			foundRegionError = true
			break
		}
	}
	assert.True(t, foundRegionError, "Should have validation error for invalid region")
}

func TestGeographicValidator_ValidateGeography_MarketCoverageResult(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	marketResult := &MarketCoverageResult{
		ServiceAreas: []ServiceArea{
			{
				Name:        "California Service Area",
				Description: "Services throughout California",
				Countries:   []string{"United States"},
				States:      []string{"California"},
			},
			{
				Name:        "European Service Area",
				Description: "Services in Europe",
				Countries:   []string{"Germany", "France"},
			},
		},
	}

	result, err := validator.ValidateGeography(context.Background(), marketResult)
	require.NoError(t, err)
	require.NotNil(t, result)

	geography := result.StandardizedGeography
	assert.Contains(t, geography.Countries, "United States")
	assert.Contains(t, geography.Countries, "Germany")
	assert.Contains(t, geography.Countries, "France")
	assert.Len(t, geography.ServiceAreas, 2)
}

func TestGeographicValidator_ValidateCountry(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name         string
		country      string
		expectedName string
		expectedCode string
		expectError  bool
	}{
		{
			name:         "valid country",
			country:      "United States",
			expectedName: "United States",
			expectedCode: "US",
			expectError:  false,
		},
		{
			name:         "country mapping",
			country:      "usa",
			expectedName: "United States",
			expectedCode: "US",
			expectError:  false,
		},
		{
			name:        "invalid country",
			country:     "Invalid Country",
			expectError: true,
		},
		{
			name:        "empty country",
			country:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, code, err := validator.validateCountry(tt.country)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, name)
				assert.Equal(t, tt.expectedCode, code)
			}
		})
	}
}

func TestGeographicValidator_ValidatePostalCode(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name        string
		postalCode  string
		countryCode string
		expectValid bool
	}{
		{
			name:        "valid US ZIP",
			postalCode:  "12345",
			countryCode: "US",
			expectValid: true,
		},
		{
			name:        "valid US ZIP+4",
			postalCode:  "12345-6789",
			countryCode: "US",
			expectValid: true,
		},
		{
			name:        "invalid US ZIP",
			postalCode:  "invalid",
			countryCode: "US",
			expectValid: false,
		},
		{
			name:        "valid Canadian postal code",
			postalCode:  "K1A 0A6",
			countryCode: "CA",
			expectValid: true,
		},
		{
			name:        "empty postal code",
			postalCode:  "",
			countryCode: "US",
			expectValid: true, // Empty is considered valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _ := validator.validatePostalCode(tt.postalCode, tt.countryCode)
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

func TestGeographicValidator_ValidatePhoneNumber(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name        string
		phone       string
		countryCode string
		expectValid bool
	}{
		{
			name:        "valid US phone",
			phone:       "+15551234567",
			countryCode: "US",
			expectValid: true,
		},
		{
			name:        "US phone without country code",
			phone:       "5551234567",
			countryCode: "US",
			expectValid: true,
		},
		{
			name:        "formatted US phone",
			phone:       "(555) 123-4567",
			countryCode: "US",
			expectValid: true,
		},
		{
			name:        "empty phone",
			phone:       "",
			countryCode: "US",
			expectValid: true, // Empty is considered valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, _ := validator.validatePhoneNumber(tt.phone, tt.countryCode)
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

func TestGeographicValidator_ValidateCoordinates(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name      string
		lat       float64
		lng       float64
		expectErr bool
	}{
		{
			name:      "valid coordinates",
			lat:       40.7128,
			lng:       -74.0060,
			expectErr: false,
		},
		{
			name:      "invalid latitude too high",
			lat:       91.0,
			lng:       0.0,
			expectErr: true,
		},
		{
			name:      "invalid latitude too low",
			lat:       -91.0,
			lng:       0.0,
			expectErr: true,
		},
		{
			name:      "invalid longitude too high",
			lat:       0.0,
			lng:       181.0,
			expectErr: true,
		},
		{
			name:      "invalid longitude too low",
			lat:       0.0,
			lng:       -181.0,
			expectErr: true,
		},
		{
			name:      "edge case valid coordinates",
			lat:       90.0,
			lng:       180.0,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateCoordinates(tt.lat, tt.lng)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGeographicValidator_StandardizeAddress(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name     string
		address  string
		expected string
	}{
		{
			name:     "standardize street abbreviation",
			address:  "123 Main St",
			expected: "123 Main Street",
		},
		{
			name:     "standardize avenue abbreviation",
			address:  "456 Oak Ave.",
			expected: "456 Oak Avenue.",
		},
		{
			name:     "multiple abbreviations",
			address:  "789 Pine Rd. Apt 2B",
			expected: "789 Pine Road. Apt 2B",
		},
		{
			name:     "no changes needed",
			address:  "123 Main Street",
			expected: "123 Main Street",
		},
		{
			name:     "empty address",
			address:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.standardizeAddress(tt.address)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeographicValidator_StandardizeCity(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name     string
		city     string
		expected string
	}{
		{
			name:     "capitalize single word",
			city:     "london",
			expected: "London",
		},
		{
			name:     "capitalize multiple words",
			city:     "new york",
			expected: "New York",
		},
		{
			name:     "already capitalized",
			city:     "San Francisco",
			expected: "San Francisco",
		},
		{
			name:     "mixed case",
			city:     "lOs aNgElEs",
			expected: "Los Angeles",
		},
		{
			name:     "empty city",
			city:     "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.standardizeCity(tt.city)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeographicValidator_StandardizeState(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name        string
		state       string
		countryCode string
		expected    string
	}{
		{
			name:        "US state abbreviation",
			state:       "ny",
			countryCode: "US",
			expected:    "New York",
		},
		{
			name:        "US state full name",
			state:       "california",
			countryCode: "US",
			expected:    "California",
		},
		{
			name:        "non-US state",
			state:       "ontario",
			countryCode: "CA",
			expected:    "Ontario",
		},
		{
			name:        "empty state",
			state:       "",
			countryCode: "US",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.standardizeState(tt.state, tt.countryCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeographicValidator_StandardizePostalCode(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name        string
		postalCode  string
		countryCode string
		expected    string
	}{
		{
			name:        "US ZIP code",
			postalCode:  "12345",
			countryCode: "US",
			expected:    "12345",
		},
		{
			name:        "Canadian postal code",
			postalCode:  "k1a0a6",
			countryCode: "CA",
			expected:    "K1A 0A6",
		},
		{
			name:        "UK postal code",
			postalCode:  "sw1a 1aa",
			countryCode: "GB",
			expected:    "SW1A 1AA",
		},
		{
			name:        "empty postal code",
			postalCode:  "",
			countryCode: "US",
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.standardizePostalCode(tt.postalCode, tt.countryCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeographicValidator_GetRegionForCountry(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name     string
		country  string
		expected string
	}{
		{
			name:     "United States",
			country:  "United States",
			expected: "North America",
		},
		{
			name:     "Germany",
			country:  "Germany",
			expected: "Europe",
		},
		{
			name:     "Japan",
			country:  "Japan",
			expected: "Asia Pacific",
		},
		{
			name:     "invalid country",
			country:  "Invalid Country",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.getRegionForCountry(tt.country)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeographicValidator_SuggestCountry(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name     string
		country  string
		expected string
	}{
		{
			name:     "usa suggestion",
			country:  "usa",
			expected: "United States",
		},
		{
			name:     "britain suggestion",
			country:  "britain",
			expected: "United Kingdom",
		},
		{
			name:     "america suggestion",
			country:  "america",
			expected: "United States",
		},
		{
			name:     "no suggestion available",
			country:  "xyz",
			expected: "",
		},
		{
			name:     "empty country",
			country:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.suggestCountry(tt.country)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGeographicValidator_RemoveDuplicateStrings(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "remove duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "remove empty strings",
			input:    []string{"a", "", "b", ""},
			expected: []string{"a", "b"},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.removeDuplicateStrings(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestGeographicValidator_UnsupportedDataType(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	unsupportedData := "invalid data type"
	result, err := validator.ValidateGeography(context.Background(), unsupportedData)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported data type")
}

func TestGeographicValidator_ProcessingTime(t *testing.T) {
	validator := NewGeographicValidator(zap.NewNop(), nil)

	locationResult := &LocationResult{
		Locations: []Location{
			{Address: "123 Main St", Country: "United States"},
		},
	}

	result, err := validator.ValidateGeography(context.Background(), locationResult)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Greater(t, result.ProcessingTime, time.Duration(0))
	assert.Greater(t, result.StandardizedGeography.ProcessingTime, time.Duration(0))
}
