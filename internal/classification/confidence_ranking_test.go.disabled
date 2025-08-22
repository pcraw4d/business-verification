package classification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfidenceRankingEngine(t *testing.T) {
	// Act
	engine := NewConfidenceRankingEngine()

	// Assert
	assert.NotNil(t, engine)
	assert.Equal(t, 0.4, engine.baseConfidenceWeight)
	assert.Equal(t, 0.2, engine.methodDiversityWeight)
	assert.Equal(t, 0.15, engine.consistencyWeight)
	assert.Equal(t, 0.15, engine.relevanceWeight)
	assert.Equal(t, 0.1, engine.industryPopularityWeight)
	assert.NotNil(t, engine.industryPopularity)
}

func TestRankClassifications_EmptyList(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{}

	// Act
	result := engine.RankClassifications(classifications)

	// Assert
	assert.Empty(t, result)
}

func TestRankClassifications_SingleClassification(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{
		{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.8,
			ClassificationMethod: "keyword_match",
		},
	}

	// Act
	result := engine.RankClassifications(classifications)

	// Assert
	assert.Len(t, result, 1)
	assert.Equal(t, "511210", result[0].IndustryCode)
	assert.GreaterOrEqual(t, result[0].ConfidenceScore, 0.7) // Should be enhanced but may not exceed original
}

func TestRankClassifications_MultipleClassifications(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{
		{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.7,
			ClassificationMethod: "keyword_match",
			Keywords:             []string{"software", "publishing"},
		},
		{
			IndustryCode:         "541511",
			IndustryName:         "Custom Computer Programming Services",
			ConfidenceScore:      0.9,
			ClassificationMethod: "description_match",
			Keywords:             []string{"programming", "computer"},
		},
		{
			IndustryCode:         "541512",
			IndustryName:         "Computer Systems Design Services",
			ConfidenceScore:      0.6,
			ClassificationMethod: "fuzzy_match",
		},
	}

	// Act
	result := engine.RankClassifications(classifications)

	// Assert
	assert.Len(t, result, 3)
	// Should be sorted by enhanced confidence score
	assert.GreaterOrEqual(t, result[0].ConfidenceScore, result[1].ConfidenceScore)
	assert.GreaterOrEqual(t, result[1].ConfidenceScore, result[2].ConfidenceScore)
}

func TestRankClassifications_RemoveDuplicates(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{
		{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.8,
			ClassificationMethod: "keyword_match",
		},
		{
			IndustryCode:         "511210", // Duplicate
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.9,
			ClassificationMethod: "description_match",
		},
		{
			IndustryCode:         "541511",
			IndustryName:         "Custom Computer Programming Services",
			ConfidenceScore:      0.7,
			ClassificationMethod: "fuzzy_match",
		},
	}

	// Act
	result := engine.RankClassifications(classifications)

	// Assert
	assert.Len(t, result, 2) // Should remove duplicate
	assert.Equal(t, "511210", result[0].IndustryCode)
	assert.Equal(t, "541511", result[1].IndustryCode)
}

func TestCalculateMethodDiversityBonus(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{
		{ClassificationMethod: "keyword_match"},
		{ClassificationMethod: "description_match"},
		{ClassificationMethod: "fuzzy_match"},
	}

	// Act
	bonus := engine.calculateMethodDiversityBonus(classifications, 0)

	// Assert
	assert.Greater(t, bonus, 0.0)
	assert.LessOrEqual(t, bonus, 1.0)
}

func TestCalculateMethodDiversityBonus_SingleMethod(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{
		{ClassificationMethod: "keyword_match"},
		{ClassificationMethod: "keyword_match"},
		{ClassificationMethod: "keyword_match"},
	}

	// Act
	bonus := engine.calculateMethodDiversityBonus(classifications, 0)

	// Assert
	assert.LessOrEqual(t, bonus, 0.2) // Low diversity bonus for single method
}

func TestCalculateConsistencyBonus(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{
		{
			IndustryCode: "511210",
			Keywords:     []string{"software", "technology"},
		},
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
	bonus := engine.calculateConsistencyBonus(classifications, 0)

	// Assert
	assert.Greater(t, bonus, 0.0)
	assert.LessOrEqual(t, bonus, 1.0)
}

func TestCalculateConsistencyBonus_SingleClassification(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{
		{IndustryCode: "511210"},
	}

	// Act
	bonus := engine.calculateConsistencyBonus(classifications, 0)

	// Assert
	assert.Equal(t, 1.0, bonus) // Perfect consistency for single classification
}

func TestCalculateRelevanceBonus(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classification := IndustryClassification{
		IndustryCode:         "511210",
		ConfidenceScore:      0.85,
		ClassificationMethod: "keyword_match",
		Keywords:             []string{"software", "technology"},
		Description:          "Software development company",
	}

	// Act
	bonus := engine.calculateRelevanceBonus(classification)

	// Assert
	assert.Greater(t, bonus, 0.0)
	assert.LessOrEqual(t, bonus, 1.0)
}

func TestCalculateRelevanceBonus_NoKeywords(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classification := IndustryClassification{
		IndustryCode:         "511210",
		ConfidenceScore:      0.5,
		ClassificationMethod: "fuzzy_match",
		Description:          "Software development company",
	}

	// Act
	bonus := engine.calculateRelevanceBonus(classification)

	// Assert
	assert.Greater(t, bonus, 0.0)
	assert.Less(t, bonus, 0.5) // Lower bonus without keywords
}

func TestCalculateIndustryPopularityBonus(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	engine.SetIndustryPopularity(map[string]float64{
		"511210": 0.8,
		"541511": 0.6,
	})

	// Act
	bonus1 := engine.calculateIndustryPopularityBonus("511210")
	bonus2 := engine.calculateIndustryPopularityBonus("541511")
	bonus3 := engine.calculateIndustryPopularityBonus("999999") // Unknown

	// Assert
	assert.Equal(t, 0.8, bonus1)
	assert.Equal(t, 0.6, bonus2)
	assert.Equal(t, 0.5, bonus3) // Default for unknown
}

func TestAreIndustriesRelated_SameMajorCategory(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	industry1 := IndustryClassification{IndustryCode: "511210"}
	industry2 := IndustryClassification{IndustryCode: "511211"}

	// Act
	related := engine.areIndustriesRelated(industry1, industry2)

	// Assert
	assert.True(t, related) // Same major category (51)
}

func TestAreIndustriesRelated_DifferentMajorCategory(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	industry1 := IndustryClassification{IndustryCode: "511210"}
	industry2 := IndustryClassification{IndustryCode: "541511"}

	// Act
	related := engine.areIndustriesRelated(industry1, industry2)

	// Assert
	assert.False(t, related) // Different major categories
}

func TestAreIndustriesRelated_KeywordOverlap(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	industry1 := IndustryClassification{
		IndustryCode: "511210",
		Keywords:     []string{"software", "technology", "development", "publishing"},
	}
	industry2 := IndustryClassification{
		IndustryCode: "541511",
		Keywords:     []string{"software", "programming", "computer", "development"},
	}

	// Act
	related := engine.areIndustriesRelated(industry1, industry2)

	// Assert
	assert.True(t, related) // Should be related due to keyword overlap
}

func TestCalculateKeywordOverlap(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	industry1 := IndustryClassification{
		Keywords: []string{"software", "technology", "development"},
	}
	industry2 := IndustryClassification{
		Keywords: []string{"software", "programming", "computer"},
	}

	// Act
	overlap := engine.calculateKeywordOverlap(industry1, industry2)

	// Assert
	assert.Equal(t, 0.2, overlap) // 1 common keyword out of 5 total unique keywords
}

func TestCalculateKeywordOverlap_NoOverlap(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	industry1 := IndustryClassification{
		Keywords: []string{"software", "technology"},
	}
	industry2 := IndustryClassification{
		Keywords: []string{"manufacturing", "automotive"},
	}

	// Act
	overlap := engine.calculateKeywordOverlap(industry1, industry2)

	// Assert
	assert.Equal(t, 0.0, overlap) // No overlap
}

func TestCalculateKeywordOverlap_EmptyKeywords(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	industry1 := IndustryClassification{Keywords: []string{}}
	industry2 := IndustryClassification{Keywords: []string{"software"}}

	// Act
	overlap := engine.calculateKeywordOverlap(industry1, industry2)

	// Assert
	assert.Equal(t, 0.0, overlap) // No overlap when one has no keywords
}

func TestIsSameMajorCategory(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()

	// Act & Assert
	assert.True(t, engine.isSameMajorCategory("511210", "511211"))  // Same NAICS major category
	assert.True(t, engine.isSameMajorCategory("541511", "541512"))  // Same NAICS major category
	assert.False(t, engine.isSameMajorCategory("511210", "541511")) // Different major categories
	assert.False(t, engine.isSameMajorCategory("51", "54"))         // Different major categories
}

func TestRemoveDuplicates(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{
		{IndustryCode: "511210", ConfidenceScore: 0.8},
		{IndustryCode: "511210", ConfidenceScore: 0.9}, // Duplicate
		{IndustryCode: "541511", ConfidenceScore: 0.7},
		{IndustryCode: "541511", ConfidenceScore: 0.6}, // Duplicate
	}

	// Act
	result := engine.removeDuplicates(classifications)

	// Assert
	assert.Len(t, result, 2)
	assert.Equal(t, "511210", result[0].IndustryCode)
	assert.Equal(t, "541511", result[1].IndustryCode)
}

func TestSetWeights(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()

	// Act
	engine.SetWeights(0.5, 0.3, 0.1, 0.05, 0.05)

	// Assert
	assert.Equal(t, 0.5, engine.baseConfidenceWeight)
	assert.Equal(t, 0.3, engine.methodDiversityWeight)
	assert.Equal(t, 0.1, engine.consistencyWeight)
	assert.Equal(t, 0.05, engine.relevanceWeight)
	assert.Equal(t, 0.05, engine.industryPopularityWeight)
}

func TestGetRankingMetrics(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{
		{
			IndustryCode:         "511210",
			ConfidenceScore:      0.8,
			ClassificationMethod: "keyword_match",
		},
		{
			IndustryCode:         "541511",
			ConfidenceScore:      0.9,
			ClassificationMethod: "description_match",
		},
		{
			IndustryCode:         "541512",
			ConfidenceScore:      0.7,
			ClassificationMethod: "fuzzy_match",
		},
	}

	// Act
	metrics := engine.GetRankingMetrics(classifications)

	// Assert
	assert.Equal(t, 3, metrics["total_classifications"])
	assert.Equal(t, 3, metrics["unique_methods"])
	assert.InDelta(t, 0.8, metrics["average_confidence"], 0.01)
	assert.Equal(t, "0.700-0.900", metrics["confidence_range"])
	assert.Equal(t, 0.7, metrics["min_confidence"])
	assert.Equal(t, 0.9, metrics["max_confidence"])
}

func TestGetRankingMetrics_Empty(t *testing.T) {
	// Arrange
	engine := NewConfidenceRankingEngine()
	classifications := []IndustryClassification{}

	// Act
	metrics := engine.GetRankingMetrics(classifications)

	// Assert
	assert.Equal(t, 0, metrics["total_classifications"])
	assert.Equal(t, 0, metrics["unique_methods"])
	assert.Equal(t, 0.0, metrics["average_confidence"])
	assert.Equal(t, "0.0-0.0", metrics["confidence_range"])
}
