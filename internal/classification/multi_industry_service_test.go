package classification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMultiIndustryService_CalculateOverallConfidence(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{}

	testCases := []struct {
		name            string
		classifications []IndustryClassification
		expected        float64
	}{
		{
			name:            "Single classification",
			classifications: []IndustryClassification{{ConfidenceScore: 0.8}},
			expected:        0.8,
		},
		{
			name: "Multiple classifications",
			classifications: []IndustryClassification{
				{ConfidenceScore: 0.9},
				{ConfidenceScore: 0.7},
				{ConfidenceScore: 0.5},
			},
			expected: 0.773, // Weighted average: (0.9*1 + 0.7*0.5 + 0.5*0.33) / (1 + 0.5 + 0.33)
		},
		{
			name:            "Empty classifications",
			classifications: []IndustryClassification{},
			expected:        0.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := service.calculateOverallConfidence(tc.classifications)

			// Assert
			assert.InDelta(t, tc.expected, result, 0.01)
		})
	}
}

func TestMultiIndustryService_RankAndFilterClassifications(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{
		minConfidenceThreshold: 0.1,
	}

	classifications := []IndustryClassification{
		{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.9,
			ClassificationMethod: "keyword_match",
		},
		{
			IndustryCode:         "541511",
			IndustryName:         "Custom Computer Programming Services",
			ConfidenceScore:      0.7,
			ClassificationMethod: "description_match",
		},
		{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.8,
			ClassificationMethod: "fuzzy_match",
		},
		{
			IndustryCode:         "541512",
			IndustryName:         "Computer Systems Design Services",
			ConfidenceScore:      0.05, // Below threshold
			ClassificationMethod: "business_type",
		},
	}

	// Act
	result := service.rankAndFilterClassifications(classifications)

	// Assert
	assert.Len(t, result, 2)                          // Should filter out low confidence and duplicates
	assert.Equal(t, "511210", result[0].IndustryCode) // Highest confidence first
	assert.Equal(t, "541511", result[1].IndustryCode)
	assert.Equal(t, 0.9, result[0].ConfidenceScore)
	assert.Equal(t, 0.7, result[1].ConfidenceScore)
}

func TestMultiIndustryService_SelectTopClassifications(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{
		maxClassifications: 3,
	}

	classifications := []IndustryClassification{
		{IndustryCode: "1", ConfidenceScore: 0.9},
		{IndustryCode: "2", ConfidenceScore: 0.8},
		{IndustryCode: "3", ConfidenceScore: 0.7},
		{IndustryCode: "4", ConfidenceScore: 0.6},
		{IndustryCode: "5", ConfidenceScore: 0.5},
	}

	// Act
	result := service.selectTopClassifications(classifications)

	// Assert
	assert.Len(t, result, 3)
	assert.Equal(t, "1", result[0].IndustryCode)
	assert.Equal(t, "2", result[1].IndustryCode)
	assert.Equal(t, "3", result[2].IndustryCode)
}

func TestMultiIndustryService_CalculateValidationScore(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{}

	result := &MultiIndustryClassification{
		PrimaryIndustry: IndustryClassification{
			ConfidenceScore: 0.9,
		},
		Classifications: []IndustryClassification{
			{ConfidenceScore: 0.9, ClassificationMethod: "keyword_match"},
			{ConfidenceScore: 0.7, ClassificationMethod: "description_match"},
		},
		OverallConfidence: 0.8,
	}

	// Act
	score := service.calculateValidationScore(result)

	// Assert
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)
}

func TestMultiIndustryService_CalculateClassificationConsistency(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{}

	testCases := []struct {
		name            string
		classifications []IndustryClassification
		expected        float64
	}{
		{
			name:            "Single classification",
			classifications: []IndustryClassification{{IndustryCode: "1"}},
			expected:        1.0,
		},
		{
			name: "Multiple classifications",
			classifications: []IndustryClassification{
				{IndustryCode: "1"},
				{IndustryCode: "2"},
			},
			expected: 0.0, // Not related by default implementation
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := service.calculateClassificationConsistency(tc.classifications)

			// Assert
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMultiIndustryService_CalculateMethodDiversity(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{}

	testCases := []struct {
		name            string
		classifications []IndustryClassification
		expected        float64
	}{
		{
			name:            "Single classification",
			classifications: []IndustryClassification{{ClassificationMethod: "method1"}},
			expected:        1.0,
		},
		{
			name: "Multiple methods",
			classifications: []IndustryClassification{
				{ClassificationMethod: "method1"},
				{ClassificationMethod: "method2"},
				{ClassificationMethod: "method3"},
			},
			expected: 1.0, // All different methods
		},
		{
			name: "Duplicate methods",
			classifications: []IndustryClassification{
				{ClassificationMethod: "method1"},
				{ClassificationMethod: "method1"},
				{ClassificationMethod: "method2"},
			},
			expected: 0.67, // 2 unique methods / 3 total
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			result := service.calculateMethodDiversity(tc.classifications)

			// Assert
			assert.InDelta(t, tc.expected, result, 0.01)
		})
	}
}

func TestMultiIndustryService_GenerateKeywordBasedClassifications(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{}

	req := &ClassificationRequest{
		Keywords: "software,development,technology",
	}

	// Act
	result := service.generateKeywordBasedClassifications(req)

	// Assert
	// Since the helper methods are stubs, we expect empty results
	assert.Len(t, result, 0)
}

func TestMultiIndustryService_GenerateDescriptionBasedClassifications(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{}

	req := &ClassificationRequest{
		Description: "Software development company specializing in web applications",
	}

	// Act
	result := service.generateDescriptionBasedClassifications(req)

	// Assert
	// Since the helper methods are stubs, we expect empty results
	assert.Len(t, result, 0)
}

func TestMultiIndustryService_GenerateBusinessTypeClassifications(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{}

	req := &ClassificationRequest{
		BusinessType: "Corporation",
	}

	// Act
	result := service.generateBusinessTypeClassifications(req)

	// Assert
	// Since the helper methods are stubs, we expect empty results
	assert.Len(t, result, 0)
}

func TestMultiIndustryService_GenerateIndustryHintClassifications(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{}

	req := &ClassificationRequest{
		Industry: "Technology",
	}

	// Act
	result := service.generateIndustryHintClassifications(req)

	// Assert
	// Since the helper methods are stubs, we expect empty results
	assert.Len(t, result, 0)
}

func TestMultiIndustryService_GenerateFuzzyMatchClassifications(t *testing.T) {
	// Arrange
	service := &MultiIndustryService{}

	req := &ClassificationRequest{
		BusinessName: "Test Business",
	}

	// Act
	result := service.generateFuzzyMatchClassifications(req)

	// Assert
	// Since the helper methods are stubs, we expect empty results
	assert.Len(t, result, 0)
}

// Integration test helper
func TestMultiIndustryService_Integration(t *testing.T) {
	// This test would require a full integration setup with real data
	// For now, we'll skip it and focus on unit tests
	t.Skip("Integration test requires full setup with real industry data")
}
