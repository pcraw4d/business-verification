package repository

import (
	"log"
	"os"
	"testing"
)

func TestIsHighConfidenceBrandMatch_KnownBrands(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	bm := NewBrandMatcher(logger)

	tests := []struct {
		name            string
		businessName    string
		expectedMatch   bool
		expectedBrand   string
		minConfidence   float64
	}{
		{
			name:          "exact match - Hilton",
			businessName:  "Hilton",
			expectedMatch: true,
			expectedBrand: "hilton",
			minConfidence: 0.85,
		},
		{
			name:          "exact match with suffix - Hilton Inc",
			businessName:  "Hilton Inc",
			expectedMatch: true,
			expectedBrand: "hilton",
			minConfidence: 0.85,
		},
		{
			name:          "exact match with LLC - Marriott LLC",
			businessName:  "Marriott LLC",
			expectedMatch: true,
			expectedBrand: "marriott",
			minConfidence: 0.85,
		},
		{
			name:          "partial match - Hilton Hotels",
			businessName:  "Hilton Hotels",
			expectedMatch: true,
			expectedBrand: "hilton",
			minConfidence: 0.85,
		},
		{
			name:          "Hyatt brand match",
			businessName:  "Hyatt Hotels",
			expectedMatch: true,
			expectedBrand: "hyatt",
			minConfidence: 0.85,
		},
		{
			name:          "IHG brand match",
			businessName:  "InterContinental Hotels",
			expectedMatch: true,
			expectedBrand: "intercontinental",
			minConfidence: 0.85,
		},
		{
			name:          "Holiday Inn brand match",
			businessName:  "Holiday Inn Express",
			expectedMatch: true,
			expectedBrand: "holiday inn",
			minConfidence: 0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isMatch, brandName, confidence := bm.IsHighConfidenceBrandMatch(tt.businessName)

			if isMatch != tt.expectedMatch {
				t.Errorf("IsHighConfidenceBrandMatch() isMatch = %v, want %v", isMatch, tt.expectedMatch)
			}

			if isMatch {
				if brandName == "" {
					t.Errorf("IsHighConfidenceBrandMatch() brandName should not be empty when match is true")
				}
				if confidence < tt.minConfidence {
					t.Errorf("IsHighConfidenceBrandMatch() confidence = %v, want >= %v", confidence, tt.minConfidence)
				}
			} else {
				if brandName != "" {
					t.Errorf("IsHighConfidenceBrandMatch() brandName should be empty when match is false, got %v", brandName)
				}
				if confidence != 0.0 {
					t.Errorf("IsHighConfidenceBrandMatch() confidence should be 0.0 when match is false, got %v", confidence)
				}
			}
		})
	}
}

func TestIsHighConfidenceBrandMatch_UnknownBrands(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	bm := NewBrandMatcher(logger)

	unknownBrands := []string{
		"Unknown Business Corp",
		"Acme Corporation",
		"Tech Startup Inc",
		"Local Restaurant LLC",
		"Retail Store",
	}

	for _, businessName := range unknownBrands {
		t.Run(businessName, func(t *testing.T) {
			isMatch, brandName, confidence := bm.IsHighConfidenceBrandMatch(businessName)

			if isMatch {
				t.Errorf("IsHighConfidenceBrandMatch() should return false for unknown brand '%s', got true", businessName)
			}
			if brandName != "" {
				t.Errorf("IsHighConfidenceBrandMatch() brandName should be empty for unknown brand, got '%s'", brandName)
			}
			if confidence != 0.0 {
				t.Errorf("IsHighConfidenceBrandMatch() confidence should be 0.0 for unknown brand, got %v", confidence)
			}
		})
	}
}

func TestIsHighConfidenceBrandMatch_Normalization(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	bm := NewBrandMatcher(logger)

	tests := []struct {
		name          string
		input         string
		shouldMatch   bool
	}{
		{
			name:        "remove Inc",
			input:       "Hilton Inc",
			shouldMatch: true,
		},
		{
			name:        "remove LLC",
			input:       "Marriott LLC",
			shouldMatch: true,
		},
		{
			name:        "remove Corp",
			input:       "Hyatt Corp",
			shouldMatch: true,
		},
		{
			name:        "remove Hotels",
			input:       "Hilton Hotels",
			shouldMatch: true,
		},
		{
			name:        "multiple suffixes",
			input:       "Hilton Hotels Inc",
			shouldMatch: true,
		},
		{
			name:        "lowercase conversion",
			input:       "HILTON",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isMatch, _, _ := bm.IsHighConfidenceBrandMatch(tt.input)
			if isMatch != tt.shouldMatch {
				t.Errorf("IsHighConfidenceBrandMatch() for '%s' = %v, want %v", tt.input, isMatch, tt.shouldMatch)
			}
		})
	}
}

func TestIsHighConfidenceBrandMatch_MCCRange(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	bm := NewBrandMatcher(logger)

	// Verify MCC range is 3000-3831
	mccRange := bm.GetMCCRangeForBrandMatch()
	if mccRange != "3000-3831" {
		t.Errorf("GetMCCRangeForBrandMatch() = %v, want '3000-3831'", mccRange)
	}

	// Verify brand matching only works for hotel brands (MCC 3000-3831)
	// This is implicit in the brand list - all brands are hotel brands
	hotelBrands := []string{"Hilton", "Marriott", "Hyatt", "Holiday Inn"}
	for _, brand := range hotelBrands {
		isMatch, _, _ := bm.IsHighConfidenceBrandMatch(brand)
		if !isMatch {
			t.Errorf("Hotel brand '%s' should match", brand)
		}
	}
}


