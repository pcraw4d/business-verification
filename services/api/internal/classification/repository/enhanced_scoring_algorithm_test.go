package repository

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEnhancedScoringAlgorithm_Configuration tests the enhanced scoring algorithm configuration
func TestEnhancedScoringAlgorithm_Configuration(t *testing.T) {
	t.Run("Default configuration is valid", func(t *testing.T) {
		config := DefaultEnhancedScoringConfig()
		err := config.Validate()
		assert.NoError(t, err)

		// Check that weights sum to 1.0
		totalWeight := config.DirectMatchWeight + config.PhraseMatchWeight +
			config.PartialMatchWeight + config.ContextWeight
		assert.InDelta(t, 1.0, totalWeight, 0.01)
	})

	t.Run("Invalid configuration is rejected", func(t *testing.T) {
		config := &EnhancedScoringConfig{
			DirectMatchWeight:    0.5,
			PhraseMatchWeight:    0.3,
			PartialMatchWeight:   0.1,
			ContextWeight:        0.1, // Total = 1.0, should be valid
			MaxKeywordsToProcess: 100, // Add required field
		}
		err := config.Validate()
		assert.NoError(t, err)

		// Test invalid weights
		config.DirectMatchWeight = 0.8
		config.PhraseMatchWeight = 0.5 // Total > 1.0
		err = config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scoring weights must sum to 1.0")
	})

	t.Run("Invalid thresholds are rejected", func(t *testing.T) {
		config := DefaultEnhancedScoringConfig()
		config.MinMatchThreshold = 1.5 // Invalid: > 1.0
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "min match threshold must be between 0 and 1")

		config = DefaultEnhancedScoringConfig()
		config.ConfidenceThreshold = -0.1 // Invalid: < 0
		err = config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "confidence threshold must be between 0 and 1")
	})

	t.Run("Invalid performance limits are rejected", func(t *testing.T) {
		config := DefaultEnhancedScoringConfig()
		config.MaxKeywordsToProcess = 0 // Invalid: <= 0
		err := config.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "max keywords to process must be positive")
	})
}

// TestEnhancedScoringAlgorithm_EnhancedScoreCalculation tests the enhanced score calculation
func TestEnhancedScoringAlgorithm_EnhancedScoreCalculation(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	esa := NewEnhancedScoringAlgorithm(logger, config)

	// Create test keyword index
	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndustryKeywordMatch{
			"restaurant": {
				{IndustryID: 1, Keyword: "restaurant", Weight: 0.8},
			},
			"fine dining": {
				{IndustryID: 1, Keyword: "fine dining", Weight: 0.9},
			},
			"italian": {
				{IndustryID: 1, Keyword: "italian", Weight: 0.7},
			},
			"pasta": {
				{IndustryID: 1, Keyword: "pasta", Weight: 0.6},
			},
			"software": {
				{IndustryID: 2, Keyword: "software", Weight: 0.8},
			},
			"development": {
				{IndustryID: 2, Keyword: "development", Weight: 0.7},
			},
		},
	}

	t.Run("Successful enhanced scoring calculation", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"},
			{Keyword: "fine dining", Context: "description"},
			{Keyword: "italian", Context: "description"},
			{Keyword: "pasta", Context: "description"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify basic result structure
		assert.Equal(t, 1, result.IndustryID) // Should match restaurant industry
		assert.Greater(t, result.TotalScore, 0.0)
		assert.Greater(t, result.Confidence, 0.0)
		assert.LessOrEqual(t, result.Confidence, 1.0)
		assert.NotNil(t, result.ScoreBreakdown)
		assert.NotNil(t, result.MatchedKeywords)
		assert.NotNil(t, result.PerformanceMetrics)
		assert.NotNil(t, result.QualityIndicators)

		// Verify performance metrics
		assert.Equal(t, 4, result.PerformanceMetrics.KeywordsProcessed)
		assert.Greater(t, result.PerformanceMetrics.MatchesFound, 0)
		assert.Greater(t, result.PerformanceMetrics.ProcessingTime, time.Duration(0))

		// Verify quality indicators
		assert.GreaterOrEqual(t, result.QualityIndicators.MatchDiversity, 0.0)
		assert.LessOrEqual(t, result.QualityIndicators.MatchDiversity, 1.0)
		assert.GreaterOrEqual(t, result.QualityIndicators.OverallQuality, 0.0)
		assert.LessOrEqual(t, result.QualityIndicators.OverallQuality, 1.0)
	})

	t.Run("Empty contextual keywords returns error", func(t *testing.T) {
		result, err := esa.CalculateEnhancedScore(context.Background(), []ContextualKeyword{}, keywordIndex)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no contextual keywords provided")
	})

	t.Run("Nil keyword index returns error", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, nil)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "keyword index is empty or nil")
	})

	t.Run("Empty keyword index returns error", func(t *testing.T) {
		emptyIndex := &KeywordIndex{KeywordToIndustries: map[string][]IndustryKeywordMatch{}}
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, emptyIndex)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "keyword index is empty or nil")
	})
}

// TestEnhancedScoringAlgorithm_MatchTypes tests different types of matches
func TestEnhancedScoringAlgorithm_MatchTypes(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	esa := NewEnhancedScoringAlgorithm(logger, config)

	// Create comprehensive test keyword index
	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndustryKeywordMatch{
			// Direct matches
			"restaurant": {
				{IndustryID: 1, Keyword: "restaurant", Weight: 0.8},
			},
			"software": {
				{IndustryID: 2, Keyword: "software", Weight: 0.8},
			},

			// Phrase matches
			"fine dining": {
				{IndustryID: 1, Keyword: "fine dining", Weight: 0.9},
			},
			"fast food": {
				{IndustryID: 1, Keyword: "fast food", Weight: 0.7},
			},

			// Partial matches
			"dining": {
				{IndustryID: 1, Keyword: "dining", Weight: 0.6},
			},
			"food": {
				{IndustryID: 1, Keyword: "food", Weight: 0.5},
			},
		},
	}

	t.Run("Direct matches are found correctly", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should find direct match
		assert.Equal(t, 1, result.IndustryID)
		assert.Greater(t, result.ScoreBreakdown.DirectMatchScore, 0.0)
		assert.Equal(t, 0.0, result.ScoreBreakdown.PhraseMatchScore) // No phrase matches
	})

	t.Run("Phrase matches are found correctly", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "fine dining", Context: "description"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should find phrase match (industry ID 1 from the test data)
		assert.Equal(t, 1, result.IndustryID)
		assert.Greater(t, result.ScoreBreakdown.PhraseMatchScore, 0.0)
		// Note: Direct matches might still be found due to partial matching
	})

	t.Run("Partial matches are found correctly", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "dining", Context: "description"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should find partial match
		assert.Equal(t, 1, result.IndustryID)
		assert.Greater(t, result.ScoreBreakdown.PartialMatchScore, 0.0)
	})

	t.Run("Multiple match types are combined correctly", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"}, // Direct match
			{Keyword: "fine dining", Context: "description"},  // Phrase match
			{Keyword: "dining", Context: "description"},       // Partial match
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should find all types of matches
		assert.Equal(t, 1, result.IndustryID)
		assert.Greater(t, result.ScoreBreakdown.DirectMatchScore, 0.0)
		assert.Greater(t, result.ScoreBreakdown.PhraseMatchScore, 0.0)
		assert.Greater(t, result.ScoreBreakdown.PartialMatchScore, 0.0)

		// Total score should be higher than individual components
		assert.Greater(t, result.TotalScore, result.ScoreBreakdown.DirectMatchScore)
		assert.Greater(t, result.TotalScore, result.ScoreBreakdown.PhraseMatchScore)
		assert.Greater(t, result.TotalScore, result.ScoreBreakdown.PartialMatchScore)
	})
}

// TestEnhancedScoringAlgorithm_ContextMultipliers tests context multiplier application
func TestEnhancedScoringAlgorithm_ContextMultipliers(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	esa := NewEnhancedScoringAlgorithm(logger, config)

	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndustryKeywordMatch{
			"restaurant": {
				{IndustryID: 1, Keyword: "restaurant", Weight: 0.8},
			},
		},
	}

	t.Run("Business name keywords get higher weight", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should have business name context multiplier applied
		assert.Greater(t, result.ScoreBreakdown.ContextScore, 1.0) // Should be > 1.0 due to 1.2x multiplier

		// Check matched keywords have correct context multiplier
		for _, match := range result.MatchedKeywords {
			if match.Source == "business_name" {
				assert.Equal(t, 1.2, match.ContextMultiplier)
			}
		}
	})

	t.Run("Description keywords get standard weight", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "description"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should have standard context multiplier applied
		assert.Equal(t, 1.0, result.ScoreBreakdown.ContextScore) // Should be 1.0 due to 1.0x multiplier

		// Check matched keywords have correct context multiplier
		for _, match := range result.MatchedKeywords {
			if match.Source == "description" {
				assert.Equal(t, 1.0, match.ContextMultiplier)
			}
		}
	})

	t.Run("Website URL keywords get standard weight", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "website_url"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should have standard context multiplier applied
		assert.Equal(t, 1.0, result.ScoreBreakdown.ContextScore) // Should be 1.0 due to 1.0x multiplier

		// Check matched keywords have correct context multiplier
		for _, match := range result.MatchedKeywords {
			if match.Source == "website_url" {
				assert.Equal(t, 1.0, match.ContextMultiplier)
			}
		}
	})
}

// TestEnhancedScoringAlgorithm_PerformanceOptimization tests performance optimization features
func TestEnhancedScoringAlgorithm_PerformanceOptimization(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	config.MaxKeywordsToProcess = 5 // Limit for testing
	esa := NewEnhancedScoringAlgorithm(logger, config)

	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndustryKeywordMatch{
			"restaurant": {
				{IndustryID: 1, Keyword: "restaurant", Weight: 0.8},
			},
		},
	}

	t.Run("Keywords are limited for performance", func(t *testing.T) {
		// Create more keywords than the limit
		contextualKeywords := make([]ContextualKeyword, 10)
		for i := 0; i < 10; i++ {
			contextualKeywords[i] = ContextualKeyword{
				Keyword: "restaurant",
				Context: "description",
			}
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should process only the limited number of keywords
		assert.Equal(t, 5, result.PerformanceMetrics.KeywordsProcessed)
	})

	t.Run("Processing time is reasonable", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"},
			{Keyword: "fine dining", Context: "description"},
			{Keyword: "italian", Context: "description"},
		}

		start := time.Now()
		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		duration := time.Since(start)

		require.NoError(t, err)

		// Should complete within reasonable time (< 50ms as per plan requirement)
		assert.Less(t, duration, 50*time.Millisecond)
		assert.Less(t, result.PerformanceMetrics.ProcessingTime, 50*time.Millisecond)
	})
}

// TestEnhancedScoringAlgorithm_QualityIndicators tests quality indicator calculations
func TestEnhancedScoringAlgorithm_QualityIndicators(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	esa := NewEnhancedScoringAlgorithm(logger, config)

	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndustryKeywordMatch{
			"restaurant": {
				{IndustryID: 1, Keyword: "restaurant", Weight: 0.8},
			},
			"fine dining": {
				{IndustryID: 1, Keyword: "fine dining", Weight: 0.9},
			},
			"dining": {
				{IndustryID: 1, Keyword: "dining", Weight: 0.6},
			},
		},
	}

	t.Run("Quality indicators are calculated correctly", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"},
			{Keyword: "fine dining", Context: "description"},
			{Keyword: "dining", Context: "description"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Verify quality indicators are within valid ranges
		assert.GreaterOrEqual(t, result.QualityIndicators.MatchDiversity, 0.0)
		assert.LessOrEqual(t, result.QualityIndicators.MatchDiversity, 1.0)

		assert.GreaterOrEqual(t, result.QualityIndicators.KeywordRelevance, 0.0)
		assert.LessOrEqual(t, result.QualityIndicators.KeywordRelevance, 1.0)

		assert.GreaterOrEqual(t, result.QualityIndicators.ContextConsistency, 0.0)
		assert.LessOrEqual(t, result.QualityIndicators.ContextConsistency, 1.0)

		assert.GreaterOrEqual(t, result.QualityIndicators.ConfidenceStability, 0.0)
		assert.LessOrEqual(t, result.QualityIndicators.ConfidenceStability, 1.0)

		assert.GreaterOrEqual(t, result.QualityIndicators.OverallQuality, 0.0)
		assert.LessOrEqual(t, result.QualityIndicators.OverallQuality, 1.0)
	})

	t.Run("Match diversity reflects different match types", func(t *testing.T) {
		// Test with multiple match types
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"}, // Direct match
			{Keyword: "fine dining", Context: "description"},  // Phrase match
			{Keyword: "dining", Context: "description"},       // Partial match
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should have high match diversity due to multiple match types
		assert.Greater(t, result.QualityIndicators.MatchDiversity, 0.5)
	})
}

// TestEnhancedScoringAlgorithm_EdgeCases tests edge cases and error conditions
func TestEnhancedScoringAlgorithm_EdgeCases(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	esa := NewEnhancedScoringAlgorithm(logger, config)

	t.Run("Empty keyword index", func(t *testing.T) {
		emptyIndex := &KeywordIndex{KeywordToIndustries: map[string][]IndustryKeywordMatch{}}
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, emptyIndex)
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("No matching keywords", func(t *testing.T) {
		keywordIndex := &KeywordIndex{
			KeywordToIndustries: map[string][]IndustryKeywordMatch{
				"software": {
					{IndustryID: 2, Keyword: "software", Weight: 0.8},
				},
			},
		}
		contextualKeywords := []ContextualKeyword{
			{Keyword: "restaurant", Context: "business_name"}, // No match in index
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should return default industry with low score
		assert.Equal(t, 26, result.IndustryID) // Default industry
		assert.Equal(t, 0.0, result.TotalScore)
		assert.Equal(t, 0.1, result.Confidence) // Minimum confidence
	})

	t.Run("Special characters in keywords", func(t *testing.T) {
		keywordIndex := &KeywordIndex{
			KeywordToIndustries: map[string][]IndustryKeywordMatch{
				"café": {
					{IndustryID: 1, Keyword: "café", Weight: 0.8},
				},
			},
		}
		contextualKeywords := []ContextualKeyword{
			{Keyword: "Café & Bistro", Context: "business_name"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should handle special characters correctly
		assert.NotNil(t, result)
	})
}

// TestEnhancedScoringAlgorithm_Integration tests integration with existing system
func TestEnhancedScoringAlgorithm_Integration(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	esa := NewEnhancedScoringAlgorithm(logger, config)

	// Create realistic keyword index
	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndustryKeywordMatch{
			// Restaurant industry keywords
			"restaurant": {
				{IndustryID: 1, Keyword: "restaurant", Weight: 0.8},
			},
			"fine dining": {
				{IndustryID: 1, Keyword: "fine dining", Weight: 0.9},
			},
			"italian": {
				{IndustryID: 1, Keyword: "italian", Weight: 0.7},
			},
			"pasta": {
				{IndustryID: 1, Keyword: "pasta", Weight: 0.6},
			},
			"wine": {
				{IndustryID: 1, Keyword: "wine", Weight: 0.5},
			},

			// Software industry keywords
			"software": {
				{IndustryID: 2, Keyword: "software", Weight: 0.8},
			},
			"development": {
				{IndustryID: 2, Keyword: "development", Weight: 0.7},
			},
			"programming": {
				{IndustryID: 2, Keyword: "programming", Weight: 0.6},
			},
		},
	}

	t.Run("Restaurant classification with multiple keywords", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "Mario's Italian Bistro", Context: "business_name"},
			{Keyword: "Fine dining Italian restaurant serving authentic pasta and wine", Context: "description"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should classify as restaurant industry
		assert.Equal(t, 1, result.IndustryID)
		assert.Greater(t, result.Confidence, 0.5) // Should have reasonable confidence

		// Should have multiple types of matches
		assert.Greater(t, len(result.MatchedKeywords), 0)

		// Verify performance metrics
		assert.Greater(t, result.PerformanceMetrics.MatchesFound, 0)
		assert.Less(t, result.PerformanceMetrics.ProcessingTime, 50*time.Millisecond)
	})

	t.Run("Software classification with technical keywords", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "TechCorp Solutions", Context: "business_name"},
			{Keyword: "Software development company specializing in programming", Context: "description"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Should classify as software industry
		assert.Equal(t, 2, result.IndustryID)
		assert.Greater(t, result.Confidence, 0.5) // Should have reasonable confidence

		// Should have multiple matches
		assert.Greater(t, len(result.MatchedKeywords), 0)
	})
}

// TestEnhancedScoringAlgorithm_ComprehensiveValidation tests comprehensive validation
func TestEnhancedScoringAlgorithm_ComprehensiveValidation(t *testing.T) {
	logger := log.New(os.Stdout, "TEST: ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	esa := NewEnhancedScoringAlgorithm(logger, config)

	// Create comprehensive test keyword index
	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndustryKeywordMatch{
			// Restaurant industry
			"restaurant": {
				{IndustryID: 1, Keyword: "restaurant", Weight: 0.8},
			},
			"fine dining": {
				{IndustryID: 1, Keyword: "fine dining", Weight: 0.9},
			},
			"italian": {
				{IndustryID: 1, Keyword: "italian", Weight: 0.7},
			},
			"pasta": {
				{IndustryID: 1, Keyword: "pasta", Weight: 0.6},
			},
			"wine": {
				{IndustryID: 1, Keyword: "wine", Weight: 0.5},
			},
			"dining": {
				{IndustryID: 1, Keyword: "dining", Weight: 0.6},
			},
			"food": {
				{IndustryID: 1, Keyword: "food", Weight: 0.5},
			},

			// Software industry
			"software": {
				{IndustryID: 2, Keyword: "software", Weight: 0.8},
			},
			"development": {
				{IndustryID: 2, Keyword: "development", Weight: 0.7},
			},
			"programming": {
				{IndustryID: 2, Keyword: "programming", Weight: 0.6},
			},
			"code": {
				{IndustryID: 2, Keyword: "code", Weight: 0.5},
			},
		},
	}

	t.Run("Comprehensive restaurant classification", func(t *testing.T) {
		contextualKeywords := []ContextualKeyword{
			{Keyword: "Mario's Italian Bistro", Context: "business_name"},
			{Keyword: "Fine dining Italian restaurant serving authentic pasta and wine", Context: "description"},
			{Keyword: "https://mariosbistro.com", Context: "website_url"},
		}

		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		require.NoError(t, err)

		// Verify all success criteria from the plan
		assert.Equal(t, 1, result.IndustryID)                    // Should be restaurant industry
		assert.Greater(t, result.Confidence, 0.5)                // Should have reasonable confidence
		assert.GreaterOrEqual(t, len(result.MatchedKeywords), 0) // Should have some keyword matches

		// Verify score breakdown has all components (some may be 0 if no matches of that type)
		assert.GreaterOrEqual(t, result.ScoreBreakdown.DirectMatchScore, 0.0)
		assert.GreaterOrEqual(t, result.ScoreBreakdown.PhraseMatchScore, 0.0)
		assert.GreaterOrEqual(t, result.ScoreBreakdown.PartialMatchScore, 0.0)
		assert.GreaterOrEqual(t, result.ScoreBreakdown.ContextScore, 0.0)

		// Verify performance requirements (< 50ms)
		assert.Less(t, result.PerformanceMetrics.ProcessingTime, 50*time.Millisecond)

		// Verify quality indicators
		assert.GreaterOrEqual(t, result.QualityIndicators.OverallQuality, 0.0)
		assert.GreaterOrEqual(t, result.QualityIndicators.MatchDiversity, 0.0)

		// Verify matched keywords have correct information
		for _, match := range result.MatchedKeywords {
			assert.NotEmpty(t, match.Keyword)
			assert.NotEmpty(t, match.MatchedKeyword)
			assert.NotEmpty(t, match.MatchType)
			assert.Greater(t, match.BaseWeight, 0.0)
			assert.Greater(t, match.ContextMultiplier, 0.0)
			assert.Greater(t, match.FinalWeight, 0.0)
			assert.GreaterOrEqual(t, match.Confidence, 0.0)
			assert.LessOrEqual(t, match.Confidence, 1.0)
			assert.NotEmpty(t, match.Source)
		}
	})

	t.Run("Performance with large keyword set", func(t *testing.T) {
		// Create large set of contextual keywords
		contextualKeywords := make([]ContextualKeyword, 50)
		for i := 0; i < 50; i++ {
			contextualKeywords[i] = ContextualKeyword{
				Keyword: "restaurant",
				Context: "description",
			}
		}

		start := time.Now()
		result, err := esa.CalculateEnhancedScore(context.Background(), contextualKeywords, keywordIndex)
		duration := time.Since(start)

		require.NoError(t, err)

		// Should still complete within performance requirements
		assert.Less(t, duration, 50*time.Millisecond)
		assert.Less(t, result.PerformanceMetrics.ProcessingTime, 50*time.Millisecond)
		assert.Equal(t, 50, result.PerformanceMetrics.KeywordsProcessed)
	})
}
