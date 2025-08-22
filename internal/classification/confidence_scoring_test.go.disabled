package classification

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfidenceScoringEngine(t *testing.T) {
	// Arrange & Act
	engine := NewConfidenceScoringEngine(nil, nil)

	// Assert
	assert.NotNil(t, engine)
	assert.Equal(t, 0.25, engine.baseConfidenceWeight)
	assert.Equal(t, 0.20, engine.keywordMatchWeight)
	assert.Equal(t, 0.15, engine.descriptionMatchWeight)
	assert.Equal(t, 0.10, engine.businessTypeWeight)
	assert.Equal(t, 0.10, engine.industryHintWeight)
	assert.Equal(t, 0.05, engine.fuzzyMatchWeight)
	assert.Equal(t, 0.05, engine.consistencyWeight)
	assert.Equal(t, 0.03, engine.diversityWeight)
	assert.Equal(t, 0.03, engine.popularityWeight)
	assert.Equal(t, 0.02, engine.recencyWeight)
	assert.Equal(t, 0.02, engine.validationWeight)
	assert.NotNil(t, engine.industryPopularity)
	assert.NotNil(t, engine.industryValidationRates)
	assert.NotNil(t, engine.keywordReliability)
}

func TestCalculateConfidenceScore_Basic(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.8,
		ClassificationMethod: "keyword_match",
		Keywords:             []string{"software", "publishing"},
	}
	request := &ClassificationRequest{
		BusinessName: "Test Software Company",
		Keywords:     "software,technology",
		Description:  "Software development company",
		BusinessType: "Corporation",
		Industry:     "Technology",
	}
	allClassifications := []IndustryClassification{classification}

	// Act
	result := engine.CalculateConfidenceScore(context.Background(), classification, request, allClassifications)

	// Assert
	assert.NotNil(t, result)
	assert.Greater(t, result.OverallScore, 0.0)
	assert.LessOrEqual(t, result.OverallScore, 1.0)
	assert.Equal(t, "enhanced_confidence_scoring", result.ScoringMethod)
	assert.NotEmpty(t, result.ConfidenceLevel)
	assert.NotEmpty(t, result.ReliabilityFactors)
}

func TestCalculateConfidenceScore_EmptyRequest(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode:         "511210",
		IndustryName:         "Software Publishers",
		ConfidenceScore:      0.8,
		ClassificationMethod: "keyword_match",
	}
	request := &ClassificationRequest{
		BusinessName: "Test Company",
	}
	allClassifications := []IndustryClassification{classification}

	// Act
	result := engine.CalculateConfidenceScore(context.Background(), classification, request, allClassifications)

	// Assert
	assert.NotNil(t, result)
	assert.Greater(t, result.OverallScore, 0.0)
	assert.Contains(t, result.UncertaintyFactors, "no_keywords_provided")
	assert.Contains(t, result.UncertaintyFactors, "no_description_provided")
}

func TestCalculateBaseConfidence(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		ConfidenceScore:      0.8,
		ClassificationMethod: "keyword_match",
	}

	// Act
	result := engine.calculateBaseConfidence(classification)

	// Assert
	assert.InDelta(t, 0.88, result, 0.001) // 0.8 * 1.1 = 0.88
}

func TestCalculateBaseConfidence_FuzzyMatch(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		ConfidenceScore:      0.8,
		ClassificationMethod: "fuzzy_match",
	}

	// Act
	result := engine.calculateBaseConfidence(classification)

	// Assert
	assert.InDelta(t, 0.72, result, 0.001) // 0.8 * 0.9 = 0.72
}

func TestCalculateKeywordScore(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		Keywords: []string{"software", "technology", "development"},
	}
	request := &ClassificationRequest{
		Keywords: "software,technology,programming",
	}

	// Act
	result := engine.calculateKeywordScore(classification, request)

	// Assert
	assert.Greater(t, result, 0.0)
	assert.LessOrEqual(t, result, 1.0)
}

func TestCalculateKeywordScore_NoKeywords(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		Keywords: []string{"software", "technology"},
	}
	request := &ClassificationRequest{
		Keywords: "",
	}

	// Act
	result := engine.calculateKeywordScore(classification, request)

	// Assert
	assert.Equal(t, 0.0, result)
}

func TestCalculateDescriptionScore(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		Keywords: []string{"software", "development", "technology"},
	}
	request := &ClassificationRequest{
		Description: "Software development company specializing in web applications and mobile technology",
	}

	// Act
	result := engine.calculateDescriptionScore(classification, request)

	// Assert
	assert.Greater(t, result, 0.0)
	assert.LessOrEqual(t, result, 1.0)
}

func TestCalculateDescriptionScore_NoDescription(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		Keywords: []string{"software", "technology"},
	}
	request := &ClassificationRequest{
		Description: "",
	}

	// Act
	result := engine.calculateDescriptionScore(classification, request)

	// Assert
	assert.Equal(t, 0.0, result)
}

func TestCalculateBusinessTypeScore(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{}
	request := &ClassificationRequest{
		BusinessType: "Corporation",
	}

	// Act
	result := engine.calculateBusinessTypeScore(classification, request)

	// Assert
	assert.Equal(t, 0.8, result)
}

func TestCalculateBusinessTypeScore_UnknownType(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{}
	request := &ClassificationRequest{
		BusinessType: "UnknownType",
	}

	// Act
	result := engine.calculateBusinessTypeScore(classification, request)

	// Assert
	assert.Equal(t, 0.5, result)
}

func TestCalculateIndustryHintScore(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryName: "Software Publishers",
		Keywords:     []string{"software", "technology"},
	}
	request := &ClassificationRequest{
		Industry: "Technology",
	}

	// Act
	result := engine.calculateIndustryHintScore(classification, request)

	// Assert
	assert.Greater(t, result, 0.0)
	assert.LessOrEqual(t, result, 1.0)
}

func TestCalculateIndustryHintScore_NoHint(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryName: "Software Publishers",
	}
	request := &ClassificationRequest{
		Industry: "",
	}

	// Act
	result := engine.calculateIndustryHintScore(classification, request)

	// Assert
	assert.Equal(t, 0.0, result)
}

func TestCalculateFuzzyMatchScore(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryName: "Software Publishers",
	}
	request := &ClassificationRequest{
		BusinessName: "Software Company",
	}

	// Act
	result := engine.calculateFuzzyMatchScore(classification, request)

	// Assert
	assert.Greater(t, result, 0.0)
	assert.LessOrEqual(t, result, 1.0)
}

func TestCalculateFuzzyMatchScore_NoBusinessName(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryName: "Software Publishers",
	}
	request := &ClassificationRequest{
		BusinessName: "",
	}

	// Act
	result := engine.calculateFuzzyMatchScore(classification, request)

	// Assert
	assert.Equal(t, 0.0, result)
}

func TestConfidenceScoringEngine_CalculateConsistencyScore(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode: "511210",
		Keywords:     []string{"software", "technology"},
	}
	allClassifications := []IndustryClassification{
		classification,
		{
			IndustryCode: "511211",
			Keywords:     []string{"software", "publishing"},
		},
		{
			IndustryCode: "541511",
			Keywords:     []string{"programming", "computer"},
		},
	}

	// Act
	result := engine.calculateConsistencyScore(classification, allClassifications)

	// Assert
	assert.Greater(t, result, 0.0)
	assert.LessOrEqual(t, result, 1.0)
}

func TestCalculateConsistencyScore_SingleClassification(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode: "511210",
	}
	allClassifications := []IndustryClassification{classification}

	// Act
	result := engine.calculateConsistencyScore(classification, allClassifications)

	// Assert
	assert.Equal(t, 1.0, result)
}

func TestConfidenceScoringEngine_CalculateDiversityScore(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode: "511210",
	}
	allClassifications := []IndustryClassification{
		classification,
		{IndustryCode: "541511"},
		{IndustryCode: "541512"},
	}

	// Act
	result := engine.calculateDiversityScore(classification, allClassifications)

	// Assert
	assert.Greater(t, result, 0.0)
	assert.LessOrEqual(t, result, 1.0)
}

func TestCalculateDiversityScore_SingleClassification(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode: "511210",
	}
	allClassifications := []IndustryClassification{classification}

	// Act
	result := engine.calculateDiversityScore(classification, allClassifications)

	// Assert
	assert.Equal(t, 1.0, result)
}

func TestCalculatePopularityScore(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	engine.SetIndustryData(
		map[string]float64{"511210": 0.8},
		nil,
		nil,
	)
	classification := IndustryClassification{
		IndustryCode: "511210",
	}

	// Act
	result := engine.calculatePopularityScore(classification)

	// Assert
	assert.Equal(t, 0.8, result)
}

func TestCalculatePopularityScore_UnknownIndustry(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode: "999999",
	}

	// Act
	result := engine.calculatePopularityScore(classification)

	// Assert
	assert.Equal(t, 0.5, result)
}

func TestCalculateRecencyScore(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode: "511210",
	}

	// Act
	result := engine.calculateRecencyScore(classification)

	// Assert
	assert.Equal(t, 0.8, result)
}

func TestCalculateValidationScore(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	engine.SetIndustryData(
		nil,
		map[string]float64{"511210": 0.9},
		nil,
	)
	classification := IndustryClassification{
		IndustryCode: "511210",
	}

	// Act
	result := engine.calculateValidationScore(classification)

	// Assert
	assert.Equal(t, 0.9, result)
}

func TestCalculateValidationScore_UnknownIndustry(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		IndustryCode: "999999",
	}

	// Act
	result := engine.calculateValidationScore(classification)

	// Assert
	assert.Equal(t, 0.7, result)
}

func TestDetermineConfidenceLevel(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)

	// Act & Assert
	assert.Equal(t, "very_high", engine.determineConfidenceLevel(0.95))
	assert.Equal(t, "high", engine.determineConfidenceLevel(0.85))
	assert.Equal(t, "medium_high", engine.determineConfidenceLevel(0.75))
	assert.Equal(t, "medium", engine.determineConfidenceLevel(0.65))
	assert.Equal(t, "medium_low", engine.determineConfidenceLevel(0.55))
	assert.Equal(t, "low", engine.determineConfidenceLevel(0.35))
	assert.Equal(t, "very_low", engine.determineConfidenceLevel(0.15))
}

func TestIdentifyReliabilityFactors(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		ConfidenceScore:      0.9,
		ClassificationMethod: "keyword_match",
		Keywords:             []string{"software", "technology"},
	}
	request := &ClassificationRequest{
		Keywords:     "software,technology",
		Description:  "Software development company with extensive experience in web applications",
		BusinessType: "Corporation",
		Industry:     "Technology",
	}

	// Act
	result := engine.identifyReliabilityFactors(classification, request)

	// Assert
	assert.Contains(t, result, "strong_keyword_matches")
	assert.Contains(t, result, "detailed_description")
	assert.Contains(t, result, "specific_business_type")
	assert.Contains(t, result, "industry_hint_provided")
	assert.Contains(t, result, "high_base_confidence")
	assert.Contains(t, result, "reliable_classification_method")
}

func TestIdentifyUncertaintyFactors(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	classification := IndustryClassification{
		ConfidenceScore:      0.3,
		ClassificationMethod: "fuzzy_match",
	}
	request := &ClassificationRequest{
		BusinessName: "Test Company",
		// Missing keywords, description, and business type
	}

	// Act
	result := engine.identifyUncertaintyFactors(classification, request)

	// Assert
	assert.Contains(t, result, "no_keywords_provided")
	assert.Contains(t, result, "no_description_provided")
	assert.Contains(t, result, "low_base_confidence")
	assert.Contains(t, result, "fuzzy_match_method")
	assert.Contains(t, result, "no_business_type_specified")
}

func TestCalculateStringSimilarity(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)

	// Act & Assert
	assert.Equal(t, 1.0, engine.calculateStringSimilarity("software company", "software company"))
	assert.Greater(t, engine.calculateStringSimilarity("software development", "software company"), 0.0)
	assert.Equal(t, 0.0, engine.calculateStringSimilarity("software", "manufacturing"))
}

func TestAreIndustriesRelated(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	industry1 := IndustryClassification{
		IndustryCode: "511210",
		Keywords:     []string{"software", "technology"},
	}
	industry2 := IndustryClassification{
		IndustryCode: "511211",
		Keywords:     []string{"software", "publishing"},
	}

	// Act
	result := engine.areIndustriesRelated(industry1, industry2)

	// Assert
	assert.True(t, result) // Same major category (51)
}

func TestAreIndustriesRelated_DifferentCategories(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	industry1 := IndustryClassification{
		IndustryCode: "511210",
		Keywords:     []string{"software", "technology"},
	}
	industry2 := IndustryClassification{
		IndustryCode: "541511",
		Keywords:     []string{"programming", "computer"},
	}

	// Act
	result := engine.areIndustriesRelated(industry1, industry2)

	// Assert
	assert.False(t, result) // Different major categories
}

func TestConfidenceScoringEngine_CalculateKeywordOverlap(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	keywords1 := []string{"software", "technology", "development"}
	keywords2 := []string{"software", "programming", "computer"}

	// Act
	result := engine.calculateKeywordOverlap(keywords1, keywords2)

	// Assert
	assert.Equal(t, 0.2, result) // 1 common keyword out of 5 total unique keywords
}

func TestConfidenceScoringEngine_CalculateKeywordOverlap_NoOverlap(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	keywords1 := []string{"software", "technology"}
	keywords2 := []string{"manufacturing", "automotive"}

	// Act
	result := engine.calculateKeywordOverlap(keywords1, keywords2)

	// Assert
	assert.Equal(t, 0.0, result)
}

func TestGetKeywordReliabilityFactor(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	engine.SetIndustryData(
		nil,
		nil,
		map[string]float64{"511210": 0.9},
	)

	// Act & Assert
	assert.Equal(t, 0.9, engine.getKeywordReliabilityFactor("511210"))
	assert.Equal(t, 0.8, engine.getKeywordReliabilityFactor("999999")) // Default
}

func TestSetIndustryData(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	popularity := map[string]float64{"511210": 0.8}
	validationRates := map[string]float64{"511210": 0.9}
	keywordReliability := map[string]float64{"511210": 0.85}

	// Act
	engine.SetIndustryData(popularity, validationRates, keywordReliability)

	// Assert
	assert.Equal(t, popularity, engine.industryPopularity)
	assert.Equal(t, validationRates, engine.industryValidationRates)
	assert.Equal(t, keywordReliability, engine.keywordReliability)
}

func TestConfidenceScoringEngine_SetWeights(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)
	weights := map[string]float64{
		"base_confidence":   0.3,
		"keyword_match":     0.25,
		"description_match": 0.2,
		"business_type":     0.15,
		"industry_hint":     0.1,
	}

	// Act
	engine.SetWeights(weights)

	// Assert
	assert.Equal(t, 0.3, engine.baseConfidenceWeight)
	assert.Equal(t, 0.25, engine.keywordMatchWeight)
	assert.Equal(t, 0.2, engine.descriptionMatchWeight)
	assert.Equal(t, 0.15, engine.businessTypeWeight)
	assert.Equal(t, 0.1, engine.industryHintWeight)
}

func TestGetWeights(t *testing.T) {
	// Arrange
	engine := NewConfidenceScoringEngine(nil, nil)

	// Act
	weights := engine.GetWeights()

	// Assert
	assert.Equal(t, 0.25, weights["base_confidence"])
	assert.Equal(t, 0.20, weights["keyword_match"])
	assert.Equal(t, 0.15, weights["description_match"])
	assert.Equal(t, 0.10, weights["business_type"])
	assert.Equal(t, 0.10, weights["industry_hint"])
	assert.Equal(t, 0.05, weights["fuzzy_match"])
	assert.Equal(t, 0.05, weights["consistency"])
	assert.Equal(t, 0.03, weights["diversity"])
	assert.Equal(t, 0.03, weights["popularity"])
	assert.Equal(t, 0.02, weights["recency"])
	assert.Equal(t, 0.02, weights["validation"])
}
