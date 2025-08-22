package classification

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTop3SelectionEngine(t *testing.T) {
	// Arrange & Act
	engine := NewTop3SelectionEngine(nil, nil)

	// Assert
	assert.NotNil(t, engine)
	assert.Equal(t, 0.15, engine.minConfidenceThreshold)
	assert.Equal(t, 0.3, engine.maxConfidenceGap)
	assert.Equal(t, 0.1, engine.diversityPenalty)
	assert.Equal(t, 0.15, engine.consistencyBonus)
	assert.Equal(t, 0.2, engine.industryCoverageWeight)
	assert.Equal(t, 0.8, engine.confidenceDecayFactor)
}

func TestSelectTop3Classifications_EmptyList(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
	classifications := []IndustryClassification{}

	// Act
	result := engine.SelectTop3Classifications(context.Background(), classifications)

	// Assert
	assert.NotNil(t, result)
	assert.Empty(t, result.AllClassifications)
	assert.Nil(t, result.SecondaryIndustry)
	assert.Nil(t, result.TertiaryIndustry)
	assert.Equal(t, "enhanced_top3_selection", result.SelectionMethod)
}

func TestSelectTop3Classifications_SingleClassification(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
	classifications := []IndustryClassification{
		{
			IndustryCode:         "511210",
			IndustryName:         "Software Publishers",
			ConfidenceScore:      0.8,
			ClassificationMethod: "keyword_match",
		},
	}

	// Act
	result := engine.SelectTop3Classifications(context.Background(), classifications)

	// Assert
	assert.NotNil(t, result)
	assert.Len(t, result.AllClassifications, 1)
	assert.Equal(t, "511210", result.PrimaryIndustry.IndustryCode)
	assert.Nil(t, result.SecondaryIndustry)
	assert.Nil(t, result.TertiaryIndustry)
	assert.Equal(t, "511210", result.PrimaryIndustry.IndustryCode)
}

func TestSelectTop3Classifications_ThreeClassifications(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
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
			ConfidenceScore:      0.8,
			ClassificationMethod: "description_match",
		},
		{
			IndustryCode:         "541512",
			IndustryName:         "Computer Systems Design Services",
			ConfidenceScore:      0.7,
			ClassificationMethod: "fuzzy_match",
		},
	}

	// Act
	result := engine.SelectTop3Classifications(context.Background(), classifications)

	// Assert
	assert.NotNil(t, result)
	assert.Len(t, result.AllClassifications, 3)
	assert.Equal(t, "511210", result.PrimaryIndustry.IndustryCode)
	assert.NotNil(t, result.SecondaryIndustry)
	assert.Equal(t, "541511", result.SecondaryIndustry.IndustryCode)
	assert.NotNil(t, result.TertiaryIndustry)
	assert.Equal(t, "541512", result.TertiaryIndustry.IndustryCode)
}

func TestSelectTop3Classifications_MoreThanThree(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
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
			ConfidenceScore:      0.8,
			ClassificationMethod: "description_match",
		},
		{
			IndustryCode:         "541512",
			IndustryName:         "Computer Systems Design Services",
			ConfidenceScore:      0.7,
			ClassificationMethod: "fuzzy_match",
		},
		{
			IndustryCode:         "541519",
			IndustryName:         "Other Computer Related Services",
			ConfidenceScore:      0.6,
			ClassificationMethod: "business_type",
		},
		{
			IndustryCode:         "541611",
			IndustryName:         "Administrative Management and General Management Consulting Services",
			ConfidenceScore:      0.5,
			ClassificationMethod: "industry_hint",
		},
	}

	// Act
	result := engine.SelectTop3Classifications(context.Background(), classifications)

	// Assert
	assert.NotNil(t, result)
	assert.Len(t, result.AllClassifications, 2) // Confidence gap validation may filter out some
	assert.Equal(t, "511210", result.PrimaryIndustry.IndustryCode)
	assert.Equal(t, "541511", result.SecondaryIndustry.IndustryCode)
	assert.Nil(t, result.TertiaryIndustry) // May be filtered out due to confidence gap
}

func TestSelectTop3Classifications_LowConfidenceFiltering(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
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
			ConfidenceScore:      0.1, // Below threshold
			ClassificationMethod: "description_match",
		},
		{
			IndustryCode:         "541512",
			IndustryName:         "Computer Systems Design Services",
			ConfidenceScore:      0.8,
			ClassificationMethod: "fuzzy_match",
		},
	}

	// Act
	result := engine.SelectTop3Classifications(context.Background(), classifications)

	// Assert
	assert.NotNil(t, result)
	assert.Len(t, result.AllClassifications, 2) // Should filter out low confidence
	assert.Equal(t, "511210", result.PrimaryIndustry.IndustryCode)
	assert.Equal(t, "541512", result.SecondaryIndustry.IndustryCode)
	assert.Nil(t, result.TertiaryIndustry)
}

func TestApplyConfidenceThreshold(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
	classifications := []IndustryClassification{
		{ConfidenceScore: 0.9},
		{ConfidenceScore: 0.1}, // Below threshold
		{ConfidenceScore: 0.8},
		{ConfidenceScore: 0.05}, // Below threshold
		{ConfidenceScore: 0.7},
	}

	// Act
	result := engine.applyConfidenceThreshold(classifications)

	// Assert
	assert.Len(t, result, 3)
	assert.Equal(t, 0.9, result[0].ConfidenceScore)
	assert.Equal(t, 0.8, result[1].ConfidenceScore)
	assert.Equal(t, 0.7, result[2].ConfidenceScore)
}

func TestCalculateEnhancedScores(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
	classifications := []IndustryClassification{
		{
			IndustryCode:         "511210",
			ConfidenceScore:      0.9,
			ClassificationMethod: "keyword_match",
		},
		{
			IndustryCode:         "541511",
			ConfidenceScore:      0.8,
			ClassificationMethod: "description_match",
		},
		{
			IndustryCode:         "541512",
			ConfidenceScore:      0.7,
			ClassificationMethod: "fuzzy_match",
		},
	}

	// Act
	result := engine.calculateEnhancedScores(classifications)

	// Assert
	assert.Len(t, result, 3)
	// Should be sorted by enhanced confidence score
	assert.GreaterOrEqual(t, result[0].ConfidenceScore, result[1].ConfidenceScore)
	assert.GreaterOrEqual(t, result[1].ConfidenceScore, result[2].ConfidenceScore)
}

func TestCalculateMethodBonus(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)

	// Act & Assert
	assert.Equal(t, 0.05, engine.calculateMethodBonus(IndustryClassification{ClassificationMethod: "keyword_match"}))
	assert.Equal(t, 0.03, engine.calculateMethodBonus(IndustryClassification{ClassificationMethod: "description_match"}))
	assert.Equal(t, 0.02, engine.calculateMethodBonus(IndustryClassification{ClassificationMethod: "business_type"}))
	assert.Equal(t, 0.01, engine.calculateMethodBonus(IndustryClassification{ClassificationMethod: "industry_hint"}))
	assert.Equal(t, 0.0, engine.calculateMethodBonus(IndustryClassification{ClassificationMethod: "fuzzy_match"}))
	assert.Equal(t, 0.0, engine.calculateMethodBonus(IndustryClassification{ClassificationMethod: "unknown"}))
}

func TestApplyDiversityRules(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
	classifications := []IndustryClassification{
		{IndustryCode: "511210", ConfidenceScore: 0.9}, // Software
		{IndustryCode: "511211", ConfidenceScore: 0.8}, // Software (same category)
		{IndustryCode: "541511", ConfidenceScore: 0.7}, // Programming
		{IndustryCode: "541512", ConfidenceScore: 0.6}, // Design
		{IndustryCode: "541611", ConfidenceScore: 0.5}, // Consulting
	}

	// Act
	result := engine.applyDiversityRules(classifications)

	// Assert
	assert.Len(t, result, 2) // Should filter out similar industries (511211 is same category as 511210)
	assert.Equal(t, "511210", result[0].IndustryCode)
	assert.Equal(t, "541511", result[1].IndustryCode)
}

func TestAreIndustriesTooSimilar(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)

	// Act & Assert
	assert.True(t, engine.areIndustriesTooSimilar("511210", "511211"))  // Same major category
	assert.True(t, engine.areIndustriesTooSimilar("541511", "541512"))  // Same major category
	assert.False(t, engine.areIndustriesTooSimilar("511210", "541511")) // Different major categories
	assert.False(t, engine.areIndustriesTooSimilar("51", "54"))         // Different major categories
}

func TestSelectFinalTop3(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
	classifications := []IndustryClassification{
		{IndustryCode: "1", ConfidenceScore: 0.9},
		{IndustryCode: "2", ConfidenceScore: 0.8},
		{IndustryCode: "3", ConfidenceScore: 0.7},
		{IndustryCode: "4", ConfidenceScore: 0.6},
		{IndustryCode: "5", ConfidenceScore: 0.5},
	}

	// Act
	result := engine.selectFinalTop3(classifications)

	// Assert
	assert.Len(t, result, 3)
	assert.Equal(t, "1", result[0].IndustryCode)
	assert.Equal(t, "2", result[1].IndustryCode)
	assert.Equal(t, "3", result[2].IndustryCode)
}

func TestValidateConfidenceGaps(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
	classifications := []IndustryClassification{
		{IndustryCode: "1", ConfidenceScore: 0.9},
		{IndustryCode: "2", ConfidenceScore: 0.7}, // Gap of 0.2 (acceptable)
		{IndustryCode: "3", ConfidenceScore: 0.3}, // Gap of 0.4 (too large)
		{IndustryCode: "4", ConfidenceScore: 0.2},
	}

	// Act
	result := engine.validateConfidenceGaps(classifications)

	// Assert
	assert.Len(t, result, 2) // Should stop at gap that's too large
	assert.Equal(t, "1", result[0].IndustryCode)
	assert.Equal(t, "2", result[1].IndustryCode)
}

func TestGetSecondaryIndustry(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)

	// Act & Assert
	assert.Nil(t, engine.getSecondaryIndustry([]IndustryClassification{}))
	assert.Nil(t, engine.getSecondaryIndustry([]IndustryClassification{{IndustryCode: "1"}}))

	secondary := engine.getSecondaryIndustry([]IndustryClassification{
		{IndustryCode: "1"},
		{IndustryCode: "2"},
	})
	assert.NotNil(t, secondary)
	assert.Equal(t, "2", secondary.IndustryCode)
}

func TestGetTertiaryIndustry(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)

	// Act & Assert
	assert.Nil(t, engine.getTertiaryIndustry([]IndustryClassification{}))
	assert.Nil(t, engine.getTertiaryIndustry([]IndustryClassification{{IndustryCode: "1"}}))
	assert.Nil(t, engine.getTertiaryIndustry([]IndustryClassification{{IndustryCode: "1"}, {IndustryCode: "2"}}))

	tertiary := engine.getTertiaryIndustry([]IndustryClassification{
		{IndustryCode: "1"},
		{IndustryCode: "2"},
		{IndustryCode: "3"},
	})
	assert.NotNil(t, tertiary)
	assert.Equal(t, "3", tertiary.IndustryCode)
}

func TestGetIndustryCode(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)

	// Act & Assert
	assert.Equal(t, "", engine.getIndustryCode(nil))
	assert.Equal(t, "511210", engine.getIndustryCode(&IndustryClassification{IndustryCode: "511210"}))
}

func TestCalculateSelectionMetrics(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
	selected := []IndustryClassification{
		{IndustryCode: "511210", ConfidenceScore: 0.9, ClassificationMethod: "keyword_match"},
		{IndustryCode: "541511", ConfidenceScore: 0.8, ClassificationMethod: "description_match"},
		{IndustryCode: "541512", ConfidenceScore: 0.7, ClassificationMethod: "fuzzy_match"},
	}
	all := append(selected, IndustryClassification{IndustryCode: "999999", ConfidenceScore: 0.5})

	// Act
	metrics := engine.calculateSelectionMetrics(selected, all)

	// Assert
	assert.NotNil(t, metrics)
	assert.InDelta(t, 0.8, metrics.OverallConfidence, 0.01)
	assert.InDelta(t, 0.2, metrics.ConfidenceSpread, 0.01)
	assert.Equal(t, 2, len(metrics.ConfidenceGaps))
	assert.Equal(t, 3, len(metrics.MethodDistribution))
	assert.Equal(t, 2, len(metrics.IndustryCategories)) // 51 and 54 categories (541511 and 541512 are both in category 54)
	assert.Greater(t, metrics.SelectionQuality, 0.0)
}

func TestCalculateDiversityScore(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)

	// Act & Assert
	assert.Equal(t, 1.0, engine.calculateDiversityScore([]IndustryClassification{}))
	assert.Equal(t, 1.0, engine.calculateDiversityScore([]IndustryClassification{{IndustryCode: "511210"}}))

	// Same category
	assert.Equal(t, 0.5, engine.calculateDiversityScore([]IndustryClassification{
		{IndustryCode: "511210"},
		{IndustryCode: "511211"},
	}))

	// Different categories
	assert.Equal(t, 1.0, engine.calculateDiversityScore([]IndustryClassification{
		{IndustryCode: "511210"},
		{IndustryCode: "541511"},
	}))
}

func TestCalculateConsistencyScore(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)

	// Act & Assert
	assert.Equal(t, 1.0, engine.calculateConsistencyScore([]IndustryClassification{}))
	assert.Equal(t, 1.0, engine.calculateConsistencyScore([]IndustryClassification{{ConfidenceScore: 0.8}}))

	// High consistency
	highConsistency := engine.calculateConsistencyScore([]IndustryClassification{
		{ConfidenceScore: 0.9},
		{ConfidenceScore: 0.8},
	})
	assert.Greater(t, highConsistency, 0.8)

	// Low consistency
	lowConsistency := engine.calculateConsistencyScore([]IndustryClassification{
		{ConfidenceScore: 0.9},
		{ConfidenceScore: 0.3},
	})
	assert.Less(t, lowConsistency, 0.5)
}

func TestCalculateCoverageScore(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)
	selected := []IndustryClassification{
		{ConfidenceScore: 0.9},
		{ConfidenceScore: 0.8},
	}
	all := append(selected, IndustryClassification{ConfidenceScore: 0.5})

	// Act
	coverage := engine.calculateCoverageScore(selected, all)

	// Assert
	assert.InDelta(t, 0.77, coverage, 0.01) // (0.9 + 0.8) / (0.9 + 0.8 + 0.5)
}

func TestSetConfiguration(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)

	// Act
	engine.SetConfiguration(0.2, 0.4, 0.15, 0.2, 0.25, 0.75)

	// Assert
	assert.Equal(t, 0.2, engine.minConfidenceThreshold)
	assert.Equal(t, 0.4, engine.maxConfidenceGap)
	assert.Equal(t, 0.15, engine.diversityPenalty)
	assert.Equal(t, 0.2, engine.consistencyBonus)
	assert.Equal(t, 0.25, engine.industryCoverageWeight)
	assert.Equal(t, 0.75, engine.confidenceDecayFactor)
}

func TestGetConfiguration(t *testing.T) {
	// Arrange
	engine := NewTop3SelectionEngine(nil, nil)

	// Act
	config := engine.GetConfiguration()

	// Assert
	assert.Equal(t, 0.15, config["min_confidence_threshold"])
	assert.Equal(t, 0.3, config["max_confidence_gap"])
	assert.Equal(t, 0.1, config["diversity_penalty"])
	assert.Equal(t, 0.15, config["consistency_bonus"])
	assert.Equal(t, 0.2, config["coverage_weight"])
	assert.Equal(t, 0.8, config["decay_factor"])
}
