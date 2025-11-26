package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassificationCodes_GetTopMCC(t *testing.T) {
	tests := []struct {
		name     string
		codes    ClassificationCodes
		limit    int
		expected int
	}{
		{
			name: "get top 3 from 5 codes",
			codes: ClassificationCodes{
				MCC: []MCCCode{
					{Code: "5734", Description: "Computer Software Stores", Confidence: 0.90},
					{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85},
					{Code: "5045", Description: "Computers, Computer Peripheral Equipment", Confidence: 0.80},
					{Code: "4814", Description: "Telecommunications Equipment", Confidence: 0.75},
					{Code: "7379", Description: "Computer Related Services", Confidence: 0.70},
				},
			},
			limit:    3,
			expected: 3,
		},
		{
			name: "get top 3 from 2 codes",
			codes: ClassificationCodes{
				MCC: []MCCCode{
					{Code: "5734", Description: "Computer Software Stores", Confidence: 0.90},
					{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85},
				},
			},
			limit:    3,
			expected: 2,
		},
		{
			name: "empty codes",
			codes: ClassificationCodes{
				MCC: []MCCCode{},
			},
			limit:    3,
			expected: 0,
		},
		{
			name: "default limit when 0",
			codes: ClassificationCodes{
				MCC: []MCCCode{
					{Code: "5734", Description: "Computer Software Stores", Confidence: 0.90},
					{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85},
					{Code: "5045", Description: "Computers, Computer Peripheral Equipment", Confidence: 0.80},
				},
			},
			limit:    0,
			expected: 3, // Default limit is 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.codes.GetTopMCC(tt.limit)
			assert.Len(t, result, tt.expected)

			// Verify sorting by confidence (descending)
			for i := 1; i < len(result); i++ {
				assert.GreaterOrEqual(t, result[i-1].Confidence, result[i].Confidence,
					"Codes should be sorted by confidence in descending order")
			}

			// Verify top code has highest confidence
			if len(result) > 0 && len(tt.codes.MCC) > 0 {
				assert.Equal(t, 0.90, result[0].Confidence, "Top code should have highest confidence")
			}
		})
	}
}

func TestClassificationCodes_GetTopSIC(t *testing.T) {
	tests := []struct {
		name     string
		codes    ClassificationCodes
		limit    int
		expected int
	}{
		{
			name: "get top 3 from 5 codes",
			codes: ClassificationCodes{
				SIC: []SICCode{
					{Code: "7372", Description: "Prepackaged Software", Confidence: 0.90},
					{Code: "7371", Description: "Computer Programming Services", Confidence: 0.85},
					{Code: "7373", Description: "Computer Integrated Systems Design", Confidence: 0.80},
					{Code: "7374", Description: "Computer Processing and Data Preparation", Confidence: 0.75},
					{Code: "7375", Description: "Information Retrieval Services", Confidence: 0.70},
				},
			},
			limit:    3,
			expected: 3,
		},
		{
			name: "empty codes",
			codes: ClassificationCodes{
				SIC: []SICCode{},
			},
			limit:    3,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.codes.GetTopSIC(tt.limit)
			assert.Len(t, result, tt.expected)

			// Verify sorting by confidence (descending)
			for i := 1; i < len(result); i++ {
				assert.GreaterOrEqual(t, result[i-1].Confidence, result[i].Confidence,
					"Codes should be sorted by confidence in descending order")
			}
		})
	}
}

func TestClassificationCodes_GetTopNAICS(t *testing.T) {
	tests := []struct {
		name     string
		codes    ClassificationCodes
		limit    int
		expected int
	}{
		{
			name: "get top 3 from 5 codes",
			codes: ClassificationCodes{
				NAICS: []NAICSCode{
					{Code: "541511", Description: "Custom Computer Programming Services", Confidence: 0.90},
					{Code: "541512", Description: "Computer Systems Design Services", Confidence: 0.85},
					{Code: "334111", Description: "Electronic Computer Manufacturing", Confidence: 0.80},
					{Code: "518210", Description: "Data Processing, Hosting, and Related Services", Confidence: 0.75},
					{Code: "541519", Description: "Other Computer Related Services", Confidence: 0.70},
				},
			},
			limit:    3,
			expected: 3,
		},
		{
			name: "empty codes",
			codes: ClassificationCodes{
				NAICS: []NAICSCode{},
			},
			limit:    3,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.codes.GetTopNAICS(tt.limit)
			assert.Len(t, result, tt.expected)

			// Verify sorting by confidence (descending)
			for i := 1; i < len(result); i++ {
				assert.GreaterOrEqual(t, result[i-1].Confidence, result[i].Confidence,
					"Codes should be sorted by confidence in descending order")
			}
		})
	}
}

func TestClassificationCodes_CalculateCodeDistribution(t *testing.T) {
	tests := []struct {
		name           string
		codes          ClassificationCodes
		expectedTotal  int
		expectedMCC    int
		expectedSIC    int
		expectedNAICS  int
		checkTopCodes  bool
	}{
		{
			name: "full distribution",
			codes: ClassificationCodes{
				MCC: []MCCCode{
					{Code: "5734", Description: "Computer Software Stores", Confidence: 0.90},
					{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85},
					{Code: "5045", Description: "Computers, Computer Peripheral Equipment", Confidence: 0.80},
					{Code: "4814", Description: "Telecommunications Equipment", Confidence: 0.75},
				},
				SIC: []SICCode{
					{Code: "7372", Description: "Prepackaged Software", Confidence: 0.90},
					{Code: "7371", Description: "Computer Programming Services", Confidence: 0.85},
					{Code: "7373", Description: "Computer Integrated Systems Design", Confidence: 0.80},
				},
				NAICS: []NAICSCode{
					{Code: "541511", Description: "Custom Computer Programming Services", Confidence: 0.90},
					{Code: "541512", Description: "Computer Systems Design Services", Confidence: 0.85},
				},
			},
			expectedTotal: 9,
			expectedMCC:   4,
			expectedSIC:   3,
			expectedNAICS: 2,
			checkTopCodes: true,
		},
		{
			name: "empty codes",
			codes: ClassificationCodes{
				MCC:   []MCCCode{},
				SIC:   []SICCode{},
				NAICS: []NAICSCode{},
			},
			expectedTotal: 0,
			expectedMCC:   0,
			expectedSIC:   0,
			expectedNAICS: 0,
			checkTopCodes: false,
		},
		{
			name: "only MCC codes",
			codes: ClassificationCodes{
				MCC: []MCCCode{
					{Code: "5734", Description: "Computer Software Stores", Confidence: 0.90},
					{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85},
				},
				SIC:   []SICCode{},
				NAICS: []NAICSCode{},
			},
			expectedTotal: 2,
			expectedMCC:   2,
			expectedSIC:   0,
			expectedNAICS: 0,
			checkTopCodes: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := tt.codes.CalculateCodeDistribution()

			require.NotNil(t, dist)
			assert.Equal(t, tt.expectedTotal, dist.Total)
			assert.Equal(t, tt.expectedMCC, dist.MCC.Count)
			assert.Equal(t, tt.expectedSIC, dist.SIC.Count)
			assert.Equal(t, tt.expectedNAICS, dist.NAICS.Count)

			if tt.checkTopCodes {
				// Verify top codes are sorted by confidence
				if len(tt.codes.MCC) > 0 {
					assert.LessOrEqual(t, len(dist.MCC.TopCodes), 3, "Should have at most 3 top MCC codes")
					if len(dist.MCC.TopCodes) > 1 {
						assert.GreaterOrEqual(t, dist.MCC.TopCodes[0].Confidence, dist.MCC.TopCodes[1].Confidence,
							"Top codes should be sorted by confidence")
					}
					// Verify average confidence calculation
					if dist.MCC.Count > 0 {
						assert.Greater(t, dist.MCC.AverageConfidence, 0.0)
						assert.LessOrEqual(t, dist.MCC.AverageConfidence, 1.0)
					}
				}

				if len(tt.codes.SIC) > 0 {
					assert.LessOrEqual(t, len(dist.SIC.TopCodes), 3, "Should have at most 3 top SIC codes")
					if len(dist.SIC.TopCodes) > 1 {
						assert.GreaterOrEqual(t, dist.SIC.TopCodes[0].Confidence, dist.SIC.TopCodes[1].Confidence,
							"Top codes should be sorted by confidence")
					}
					if dist.SIC.Count > 0 {
						assert.Greater(t, dist.SIC.AverageConfidence, 0.0)
						assert.LessOrEqual(t, dist.SIC.AverageConfidence, 1.0)
					}
				}

				if len(tt.codes.NAICS) > 0 {
					assert.LessOrEqual(t, len(dist.NAICS.TopCodes), 3, "Should have at most 3 top NAICS codes")
					if len(dist.NAICS.TopCodes) > 1 {
						assert.GreaterOrEqual(t, dist.NAICS.TopCodes[0].Confidence, dist.NAICS.TopCodes[1].Confidence,
							"Top codes should be sorted by confidence")
					}
					if dist.NAICS.Count > 0 {
						assert.Greater(t, dist.NAICS.AverageConfidence, 0.0)
						assert.LessOrEqual(t, dist.NAICS.AverageConfidence, 1.0)
					}
				}
			}
		})
	}
}

func TestClassificationCodes_CalculateCodeDistribution_AverageConfidence(t *testing.T) {
	codes := ClassificationCodes{
		MCC: []MCCCode{
			{Code: "5734", Description: "Computer Software Stores", Confidence: 0.90},
			{Code: "7372", Description: "Prepackaged Software", Confidence: 0.85},
			{Code: "5045", Description: "Computers, Computer Peripheral Equipment", Confidence: 0.80},
		},
	}

	dist := codes.CalculateCodeDistribution()

	// Calculate expected average
	expectedAvg := (0.90 + 0.85 + 0.80) / 3.0
	assert.InDelta(t, expectedAvg, dist.MCC.AverageConfidence, 0.01, "Average confidence should be calculated correctly")
}

func TestClassificationCodes_CalculateCodeDistribution_TopCodesLimit(t *testing.T) {
	// Create codes with more than 3 items
	codes := ClassificationCodes{
		MCC: []MCCCode{
			{Code: "5734", Description: "Code 1", Confidence: 0.90},
			{Code: "7372", Description: "Code 2", Confidence: 0.85},
			{Code: "5045", Description: "Code 3", Confidence: 0.80},
			{Code: "4814", Description: "Code 4", Confidence: 0.75},
			{Code: "7379", Description: "Code 5", Confidence: 0.70},
		},
	}

	dist := codes.CalculateCodeDistribution()

	// Should only have top 3 codes
	assert.LessOrEqual(t, len(dist.MCC.TopCodes), 3, "Should limit to top 3 codes")
	assert.Equal(t, 3, len(dist.MCC.TopCodes), "Should have exactly 3 top codes")

	// Verify they are the highest confidence codes
	assert.Equal(t, 0.90, dist.MCC.TopCodes[0].Confidence, "First code should have highest confidence")
	assert.Equal(t, 0.85, dist.MCC.TopCodes[1].Confidence, "Second code should have second highest confidence")
	assert.Equal(t, 0.80, dist.MCC.TopCodes[2].Confidence, "Third code should have third highest confidence")
}

