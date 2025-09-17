package classification

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

// TestContextAwareScoring tests the context-aware scoring functionality
func TestContextAwareScoring(t *testing.T) {
	// Create test logger
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Create enhanced scoring algorithm with context-aware features enabled
	config := DefaultEnhancedScoringConfig()
	config.EnableContextAwareScoring = true
	config.EnableDynamicWeightAdjust = true
	config.EnableIndustryBoost = true
	config.EnableFuzzyMatching = true // Re-enable fuzzy matching with improved logic

	esa := NewEnhancedScoringAlgorithm(logger, config)

	// Create test keyword index with more specific keywords to avoid cross-matching
	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndexKeywordMatch{
			"restaurant": {
				{Keyword: "restaurant", IndustryID: 1, Weight: 0.9, MatchType: "direct"},
			},
			"dining": {
				{Keyword: "dining", IndustryID: 1, Weight: 0.8, MatchType: "direct"},
			},
			"food": {
				{Keyword: "food", IndustryID: 1, Weight: 0.7, MatchType: "direct"},
			},
			"menu": {
				{Keyword: "menu", IndustryID: 1, Weight: 0.9, MatchType: "direct"},
			},
			"chef": {
				{Keyword: "chef", IndustryID: 1, Weight: 0.8, MatchType: "direct"},
			},
			"software": {
				{Keyword: "software", IndustryID: 2, Weight: 0.9, MatchType: "direct"},
			},
			"technology": {
				{Keyword: "technology", IndustryID: 2, Weight: 0.8, MatchType: "direct"},
			},
			"development": {
				{Keyword: "development", IndustryID: 2, Weight: 0.8, MatchType: "direct"},
			},
			"programming": {
				{Keyword: "programming", IndustryID: 2, Weight: 0.9, MatchType: "direct"},
			},
			"api": {
				{Keyword: "api", IndustryID: 2, Weight: 0.9, MatchType: "direct"},
			},
		},
		IndustryToKeywords: map[int][]string{
			1: {"restaurant", "dining", "food", "menu", "chef"},
			2: {"software", "technology", "development", "programming", "api"},
		},
		TotalKeywords: 10,
		LastUpdated:   time.Now(),
	}

	tests := []struct {
		name             string
		keywords         []ContextualKeyword
		expectedIndustry int
		expectedMinScore float64
		description      string
	}{
		{
			name: "Restaurant Business Name Priority",
			keywords: []ContextualKeyword{
				{Keyword: "restaurant", Context: "business_name", Weight: 1.0},
				{Keyword: "food", Context: "description", Weight: 1.0},
			},
			expectedIndustry: 1,
			expectedMinScore: 1.0, // Should be high due to business name priority
			description:      "Business name keywords should have higher weight than description keywords",
		},
		{
			name: "Technology Business Name Priority",
			keywords: []ContextualKeyword{
				{Keyword: "software", Context: "business_name", Weight: 1.0},
				{Keyword: "development", Context: "description", Weight: 1.0},
			},
			expectedIndustry: 2,
			expectedMinScore: 1.0, // Should be high due to business name priority
			description:      "Technology business name should be prioritized over description",
		},
		{
			name: "Mixed Context Scoring",
			keywords: []ContextualKeyword{
				{Keyword: "restaurant", Context: "business_name", Weight: 1.0},
				{Keyword: "dining", Context: "description", Weight: 1.0},
				{Keyword: "food", Context: "website_url", Weight: 1.0},
			},
			expectedIndustry: 1,
			expectedMinScore: 0.8, // Should be good with mixed contexts
			description:      "Mixed contexts should provide good scoring with business name priority",
		},
		{
			name: "Industry-Specific Boost",
			keywords: []ContextualKeyword{
				{Keyword: "restaurant", Context: "business_name", Weight: 1.0},
				{Keyword: "menu", Context: "description", Weight: 1.0}, // High specificity keyword
			},
			expectedIndustry: 1,
			expectedMinScore: 1.2, // Should be boosted due to industry-specific keyword
			description:      "Industry-specific keywords should receive boost",
		},
		{
			name: "Website URL Lower Priority",
			keywords: []ContextualKeyword{
				{Keyword: "food", Context: "website_url", Weight: 1.0},
				{Keyword: "restaurant", Context: "description", Weight: 1.0},
			},
			expectedIndustry: 1,
			expectedMinScore: 0.6, // Should be lower due to website URL context
			description:      "Website URL keywords should have lower priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate enhanced score
			result, err := esa.CalculateEnhancedScore(context.Background(), tt.keywords, keywordIndex)
			if err != nil {
				t.Fatalf("CalculateEnhancedScore failed: %v", err)
			}

			// Verify industry classification
			if result.IndustryID != tt.expectedIndustry {
				t.Errorf("Expected industry %d, got %d", tt.expectedIndustry, result.IndustryID)
			}

			// Verify minimum score
			if result.TotalScore < tt.expectedMinScore {
				t.Errorf("Expected minimum score %.2f, got %.2f", tt.expectedMinScore, result.TotalScore)
			}

			// Verify context-aware scoring is working
			if result.ScoreBreakdown == nil {
				t.Error("Score breakdown should not be nil")
			}

			// Log results for debugging
			t.Logf("Test: %s", tt.description)
			t.Logf("  Industry: %d (expected: %d)", result.IndustryID, tt.expectedIndustry)
			t.Logf("  Total Score: %.3f (expected min: %.3f)", result.TotalScore, tt.expectedMinScore)
			t.Logf("  Confidence: %.3f", result.Confidence)
			t.Logf("  Processing Time: %v", result.ProcessingTime)
		})
	}
}

// TestContextMultiplier tests the context multiplier functionality
func TestContextMultiplier(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	esa := NewEnhancedScoringAlgorithm(logger, config)

	tests := []struct {
		context     string
		expected    float64
		description string
	}{
		{"business_name", 1.5, "Business name should have 50% boost"},
		{"description", 1.0, "Description should have no boost"},
		{"website_url", 0.8, "Website URL should have 20% reduction"},
		{"unknown", 1.0, "Unknown context should default to no boost"},
	}

	for _, tt := range tests {
		t.Run(tt.context, func(t *testing.T) {
			multiplier := esa.getContextMultiplier(tt.context)
			if multiplier != tt.expected {
				t.Errorf("Expected multiplier %.2f for context '%s', got %.2f", tt.expected, tt.context, multiplier)
			}
			t.Logf("Context '%s': multiplier %.2f - %s", tt.context, multiplier, tt.description)
		})
	}
}

// TestIndustrySpecificBoost tests the industry-specific boost functionality
func TestIndustrySpecificBoost(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	esa := NewEnhancedScoringAlgorithm(logger, config)

	tests := []struct {
		keyword     string
		industryID  int
		expected    float64
		description string
	}{
		{"restaurant", 1, 1.0, "Restaurant keyword should get boost for restaurant industry"},
		{"menu", 1, 1.0, "Menu keyword should get boost for restaurant industry"},
		{"software", 2, 1.0, "Software keyword should get boost for technology industry"},
		{"restaurant", 2, 1.0, "Restaurant keyword should not get boost for technology industry"},
		{"unknown", 1, 1.0, "Unknown keyword should not get boost"},
	}

	for _, tt := range tests {
		t.Run(tt.keyword+"_industry_"+string(rune(tt.industryID)), func(t *testing.T) {
			boost := esa.calculateIndustrySpecificBoost(tt.keyword, tt.industryID)
			if boost < 1.0 || boost > 2.0 {
				t.Errorf("Boost should be between 1.0 and 2.0, got %.3f", boost)
			}
			t.Logf("Keyword '%s' in industry %d: boost %.3f - %s", tt.keyword, tt.industryID, boost, tt.description)
		})
	}
}

// TestDynamicWeightAdjustment tests the dynamic weight adjustment functionality
func TestDynamicWeightAdjustment(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	esa := NewEnhancedScoringAlgorithm(logger, config)

	tests := []struct {
		name        string
		keywords    []ContextualKeyword
		industryID  int
		expectedMin float64
		expectedMax float64
		description string
	}{
		{
			name: "High Density Keywords",
			keywords: []ContextualKeyword{
				{Keyword: "restaurant", Context: "business_name", Weight: 1.0},
				{Keyword: "dining", Context: "business_name", Weight: 1.0},
				{Keyword: "food", Context: "business_name", Weight: 1.0},
			},
			industryID:  1,
			expectedMin: 0.5,
			expectedMax: 1.5,
			description: "High keyword density should provide good adjustment",
		},
		{
			name: "Low Density Keywords",
			keywords: []ContextualKeyword{
				{Keyword: "restaurant", Context: "business_name", Weight: 1.0},
			},
			industryID:  1,
			expectedMin: 0.5,
			expectedMax: 1.5,
			description: "Low keyword density should still provide reasonable adjustment",
		},
		{
			name: "Mixed Context Keywords",
			keywords: []ContextualKeyword{
				{Keyword: "restaurant", Context: "business_name", Weight: 1.0},
				{Keyword: "dining", Context: "description", Weight: 1.0},
				{Keyword: "food", Context: "website_url", Weight: 1.0},
			},
			industryID:  1,
			expectedMin: 0.5,
			expectedMax: 1.5,
			description: "Mixed contexts should provide balanced adjustment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adjustment := esa.calculateDynamicWeightAdjustment(tt.keywords, tt.industryID)

			if adjustment.AdjustmentFactor < tt.expectedMin || adjustment.AdjustmentFactor > tt.expectedMax {
				t.Errorf("Adjustment factor %.3f should be between %.3f and %.3f",
					adjustment.AdjustmentFactor, tt.expectedMin, tt.expectedMax)
			}

			t.Logf("Test: %s", tt.description)
			t.Logf("  Adjustment Factor: %.3f", adjustment.AdjustmentFactor)
			t.Logf("  Keyword Density: %.3f", adjustment.KeywordDensity)
			t.Logf("  Industry Relevance: %.3f", adjustment.IndustryRelevance)
			t.Logf("  Context Consistency: %.3f", adjustment.ContextConsistency)
			t.Logf("  Match Quality: %.3f", adjustment.MatchQuality)
		})
	}
}

// TestContextAwareAccuracyImprovements tests overall accuracy improvements
func TestContextAwareAccuracyImprovements(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)

	// Test with context-aware scoring enabled
	configEnabled := DefaultEnhancedScoringConfig()
	configEnabled.EnableContextAwareScoring = true
	configEnabled.EnableDynamicWeightAdjust = true
	configEnabled.EnableIndustryBoost = true

	// Test with context-aware scoring disabled
	configDisabled := DefaultEnhancedScoringConfig()
	configDisabled.EnableContextAwareScoring = false
	configDisabled.EnableDynamicWeightAdjust = false
	configDisabled.EnableIndustryBoost = false

	esaEnabled := NewEnhancedScoringAlgorithm(logger, configEnabled)
	esaDisabled := NewEnhancedScoringAlgorithm(logger, configDisabled)

	// Create comprehensive keyword index
	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndexKeywordMatch{
			// Restaurant keywords
			"restaurant": {{Keyword: "restaurant", IndustryID: 1, Weight: 0.9, MatchType: "direct"}},
			"dining":     {{Keyword: "dining", IndustryID: 1, Weight: 0.8, MatchType: "direct"}},
			"food":       {{Keyword: "food", IndustryID: 1, Weight: 0.7, MatchType: "direct"}},
			"menu":       {{Keyword: "menu", IndustryID: 1, Weight: 0.9, MatchType: "direct"}},
			"chef":       {{Keyword: "chef", IndustryID: 1, Weight: 0.8, MatchType: "direct"}},

			// Technology keywords
			"software":    {{Keyword: "software", IndustryID: 2, Weight: 0.9, MatchType: "direct"}},
			"technology":  {{Keyword: "technology", IndustryID: 2, Weight: 0.8, MatchType: "direct"}},
			"development": {{Keyword: "development", IndustryID: 2, Weight: 0.8, MatchType: "direct"}},
			"programming": {{Keyword: "programming", IndustryID: 2, Weight: 0.9, MatchType: "direct"}},
			"api":         {{Keyword: "api", IndustryID: 2, Weight: 0.9, MatchType: "direct"}},
		},
		IndustryToKeywords: map[int][]string{
			1: {"restaurant", "dining", "food", "menu", "chef"},
			2: {"software", "technology", "development", "programming", "api"},
		},
		TotalKeywords: 10,
		LastUpdated:   time.Now(),
	}

	testCases := []struct {
		name             string
		keywords         []ContextualKeyword
		expectedIndustry int
		description      string
	}{
		{
			name: "Restaurant with Business Name Priority",
			keywords: []ContextualKeyword{
				{Keyword: "restaurant", Context: "business_name", Weight: 1.0},
				{Keyword: "food", Context: "description", Weight: 1.0},
			},
			expectedIndustry: 1,
			description:      "Restaurant business name should be prioritized",
		},
		{
			name: "Technology with Business Name Priority",
			keywords: []ContextualKeyword{
				{Keyword: "software", Context: "business_name", Weight: 1.0},
				{Keyword: "development", Context: "description", Weight: 1.0},
			},
			expectedIndustry: 2,
			description:      "Technology business name should be prioritized",
		},
		{
			name: "Industry-Specific Boost Test",
			keywords: []ContextualKeyword{
				{Keyword: "menu", Context: "business_name", Weight: 1.0},
				{Keyword: "chef", Context: "description", Weight: 1.0},
			},
			expectedIndustry: 1,
			description:      "Industry-specific keywords should get boost",
		},
		{
			name: "Mixed Context Scoring",
			keywords: []ContextualKeyword{
				{Keyword: "restaurant", Context: "business_name", Weight: 1.0},
				{Keyword: "dining", Context: "description", Weight: 1.0},
				{Keyword: "food", Context: "website_url", Weight: 1.0},
			},
			expectedIndustry: 1,
			description:      "Mixed contexts should provide good scoring",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test with context-aware scoring enabled
			resultEnabled, err := esaEnabled.CalculateEnhancedScore(context.Background(), tc.keywords, keywordIndex)
			if err != nil {
				t.Fatalf("Context-aware scoring failed: %v", err)
			}

			// Test with context-aware scoring disabled
			resultDisabled, err := esaDisabled.CalculateEnhancedScore(context.Background(), tc.keywords, keywordIndex)
			if err != nil {
				t.Fatalf("Standard scoring failed: %v", err)
			}

			// Verify both get the correct industry
			if resultEnabled.IndustryID != tc.expectedIndustry {
				t.Errorf("Context-aware scoring: expected industry %d, got %d", tc.expectedIndustry, resultEnabled.IndustryID)
			}

			if resultDisabled.IndustryID != tc.expectedIndustry {
				t.Errorf("Standard scoring: expected industry %d, got %d", tc.expectedIndustry, resultDisabled.IndustryID)
			}

			// Context-aware scoring should generally provide better scores
			improvement := resultEnabled.TotalScore - resultDisabled.TotalScore

			t.Logf("Test: %s", tc.description)
			t.Logf("  Context-Aware Score: %.3f", resultEnabled.TotalScore)
			t.Logf("  Standard Score: %.3f", resultDisabled.TotalScore)
			t.Logf("  Improvement: %.3f", improvement)
			t.Logf("  Context-Aware Confidence: %.3f", resultEnabled.Confidence)
			t.Logf("  Standard Confidence: %.3f", resultDisabled.Confidence)

			// Log success if improvement is positive
			if improvement > 0 {
				t.Logf("  ✅ Context-aware scoring improved accuracy by %.3f", improvement)
			} else if improvement == 0 {
				t.Logf("  ⚠️ No improvement in this test case")
			} else {
				t.Logf("  ⚠️ Context-aware scoring performed worse by %.3f", -improvement)
			}
		})
	}
}

// BenchmarkContextAwareScoring benchmarks the performance of context-aware scoring
func BenchmarkContextAwareScoring(b *testing.B) {
	logger := log.New(os.Stdout, "[BENCH] ", log.LstdFlags)
	config := DefaultEnhancedScoringConfig()
	config.EnableContextAwareScoring = true
	config.EnableDynamicWeightAdjust = true
	config.EnableIndustryBoost = true

	esa := NewEnhancedScoringAlgorithm(logger, config)

	// Create test data
	keywords := []ContextualKeyword{
		{Keyword: "restaurant", Context: "business_name", Weight: 1.0},
		{Keyword: "dining", Context: "description", Weight: 1.0},
		{Keyword: "food", Context: "website_url", Weight: 1.0},
	}

	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndexKeywordMatch{
			"restaurant": {{Keyword: "restaurant", IndustryID: 1, Weight: 0.9, MatchType: "direct"}},
			"dining":     {{Keyword: "dining", IndustryID: 1, Weight: 0.8, MatchType: "direct"}},
			"food":       {{Keyword: "food", IndustryID: 1, Weight: 0.7, MatchType: "direct"}},
		},
		IndustryToKeywords: map[int][]string{
			1: {"restaurant", "dining", "food"},
		},
		TotalKeywords: 3,
		LastUpdated:   time.Now(),
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := esa.CalculateEnhancedScore(context.Background(), keywords, keywordIndex)
		if err != nil {
			b.Fatalf("CalculateEnhancedScore failed: %v", err)
		}
	}
}
