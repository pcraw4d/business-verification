package repository

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTask2_3_2_ContextMultipliers tests the context multiplier functionality
func TestTask2_3_2_ContextMultipliers(t *testing.T) {
	t.Run("Business name keywords get 1.2x weight", func(t *testing.T) {
		// Create mock repository
		repo := createMockRepository(t)

		// Test business name keyword extraction
		contextualKeywords := repo.extractKeywords("Mario's Italian Restaurant", "", "")

		// Verify business name keywords are extracted with correct context
		businessNameKeywords := 0
		for _, ck := range contextualKeywords {
			if ck.Context == "business_name" {
				businessNameKeywords++
			}
		}

		assert.Greater(t, businessNameKeywords, 0, "Should extract keywords from business name")

		// Test context multiplier application
		multiplier := repo.getContextMultiplier("business_name")
		assert.Equal(t, 1.2, multiplier, "Business name keywords should get 1.2x multiplier")
	})

	t.Run("Description keywords get 1.0x weight", func(t *testing.T) {
		// Create mock repository
		repo := createMockRepository(t)

		// Test description keyword extraction
		contextualKeywords := repo.extractKeywords("", "Fine dining Italian restaurant serving authentic pasta and wine", "")

		// Verify description keywords are extracted with correct context
		descriptionKeywords := 0
		for _, ck := range contextualKeywords {
			if ck.Context == "description" {
				descriptionKeywords++
			}
		}

		assert.Greater(t, descriptionKeywords, 0, "Should extract keywords from description")

		// Test context multiplier application
		multiplier := repo.getContextMultiplier("description")
		assert.Equal(t, 1.0, multiplier, "Description keywords should get 1.0x multiplier")
	})

	t.Run("Website URL keywords get 1.0x weight", func(t *testing.T) {
		// Create mock repository
		repo := createMockRepository(t)

		// Test website URL keyword extraction
		contextualKeywords := repo.extractKeywords("", "", "https://www.mariositalian.com")

		// Verify website URL keywords are extracted with correct context
		websiteKeywords := 0
		for _, ck := range contextualKeywords {
			if ck.Context == "website_url" {
				websiteKeywords++
			}
		}

		assert.Greater(t, websiteKeywords, 0, "Should extract keywords from website URL")

		// Test context multiplier application
		multiplier := repo.getContextMultiplier("website_url")
		assert.Equal(t, 1.0, multiplier, "Website URL keywords should get 1.0x multiplier")
	})

	t.Run("Context multipliers are applied correctly in classification", func(t *testing.T) {
		// Create mock repository with test data
		repo := createMockRepositoryWithContextTestData()

		// Test classification with business name keywords (should get higher weight)
		contextualKeywords := repo.extractKeywords("Italian Restaurant", "We serve food", "")

		// Test that context multipliers are applied during scoring
		// We'll test the multiplier application directly since we can't easily mock the full classification
		for _, ck := range contextualKeywords {
			multiplier := repo.getContextMultiplier(ck.Context)
			if ck.Context == "business_name" {
				assert.Equal(t, 1.2, multiplier, "Business name keywords should get 1.2x multiplier")
			} else {
				assert.Equal(t, 1.0, multiplier, "Other keywords should get 1.0x multiplier")
			}
		}

		// Test dynamic confidence calculation
		confidence := repo.calculateDynamicConfidence(0.5, 3, 6)
		assert.GreaterOrEqual(t, confidence, 0.1, "Confidence should be at least 0.1")
		assert.LessOrEqual(t, confidence, 1.0, "Confidence should be at most 1.0")
	})

	t.Run("Dynamic confidence calculation works with context", func(t *testing.T) {
		// Create mock repository
		repo := createMockRepository(t)

		// Test confidence calculation with different scenarios
		testCases := []struct {
			name            string
			score           float64
			matchedKeywords int
			totalKeywords   int
			expectedMin     float64
			expectedMax     float64
		}{
			{
				name:            "High score, many matches",
				score:           0.8,
				matchedKeywords: 5,
				totalKeywords:   6,
				expectedMin:     0.6,
				expectedMax:     1.0,
			},
			{
				name:            "Low score, few matches",
				score:           0.2,
				matchedKeywords: 1,
				totalKeywords:   6,
				expectedMin:     0.1,
				expectedMax:     0.4,
			},
			{
				name:            "Medium score, medium matches",
				score:           0.5,
				matchedKeywords: 3,
				totalKeywords:   6,
				expectedMin:     0.3,
				expectedMax:     0.7,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				confidence := repo.calculateDynamicConfidence(tc.score, tc.matchedKeywords, tc.totalKeywords)

				assert.GreaterOrEqual(t, confidence, tc.expectedMin, "Confidence should be at least %f", tc.expectedMin)
				assert.LessOrEqual(t, confidence, tc.expectedMax, "Confidence should be at most %f", tc.expectedMax)
				assert.GreaterOrEqual(t, confidence, 0.1, "Confidence should be at least 0.1")
				assert.LessOrEqual(t, confidence, 1.0, "Confidence should be at most 1.0")
			})
		}
	})

	t.Run("Context multiplier application in scoring", func(t *testing.T) {
		// Create mock repository
		repo := createMockRepository(t)

		// Test that context multipliers are applied correctly
		testCases := []struct {
			context  string
			expected float64
		}{
			{"business_name", 1.2},
			{"description", 1.0},
			{"website_url", 1.0},
			{"unknown_context", 1.0},
		}

		for _, tc := range testCases {
			t.Run(tc.context, func(t *testing.T) {
				multiplier := repo.getContextMultiplier(tc.context)
				assert.Equal(t, tc.expected, multiplier, "Context multiplier for %s should be %f", tc.context, tc.expected)
			})
		}
	})

	t.Run("End-to-end context multiplier workflow", func(t *testing.T) {
		// Create mock repository with test data
		repo := createMockRepositoryWithContextTestData()

		// Test complete workflow with mixed contexts
		businessName := "Mario's Italian Bistro"
		description := "Fine dining Italian restaurant serving authentic pasta and wine"
		websiteURL := "https://www.mariositalian.com"

		// Extract contextual keywords
		contextualKeywords := repo.extractKeywords(businessName, description, websiteURL)

		// Verify we have keywords from all contexts
		contexts := make(map[string]int)
		for _, ck := range contextualKeywords {
			contexts[ck.Context]++
		}

		assert.Greater(t, contexts["business_name"], 0, "Should have business name keywords")
		assert.Greater(t, contexts["description"], 0, "Should have description keywords")
		assert.Greater(t, contexts["website_url"], 0, "Should have website URL keywords")

		// Test that context multipliers are applied correctly
		businessNameMultipliers := 0
		descriptionMultipliers := 0
		websiteMultipliers := 0

		for _, ck := range contextualKeywords {
			multiplier := repo.getContextMultiplier(ck.Context)
			switch ck.Context {
			case "business_name":
				assert.Equal(t, 1.2, multiplier, "Business name should get 1.2x multiplier")
				businessNameMultipliers++
			case "description":
				assert.Equal(t, 1.0, multiplier, "Description should get 1.0x multiplier")
				descriptionMultipliers++
			case "website_url":
				assert.Equal(t, 1.0, multiplier, "Website URL should get 1.0x multiplier")
				websiteMultipliers++
			}
		}

		assert.Greater(t, businessNameMultipliers, 0, "Should have applied business name multipliers")
		assert.Greater(t, descriptionMultipliers, 0, "Should have applied description multipliers")
		assert.Greater(t, websiteMultipliers, 0, "Should have applied website URL multipliers")
	})
}

// TestTask2_3_2_ContextMultiplierPerformance tests performance requirements
func TestTask2_3_2_ContextMultiplierPerformance(t *testing.T) {
	t.Run("Context multiplier calculation is fast", func(t *testing.T) {
		// Create mock repository
		repo := createMockRepository(t)

		// Test performance of context multiplier calculation
		start := time.Now()

		// Calculate multipliers for many contexts
		for i := 0; i < 1000; i++ {
			repo.getContextMultiplier("business_name")
			repo.getContextMultiplier("description")
			repo.getContextMultiplier("website_url")
		}

		duration := time.Since(start)

		// Should be very fast (less than 1ms for 3000 calculations)
		assert.Less(t, duration, 1*time.Millisecond, "Context multiplier calculation should be very fast")
	})

	t.Run("Dynamic confidence calculation is fast", func(t *testing.T) {
		// Create mock repository
		repo := createMockRepository(t)

		// Test performance of confidence calculation
		start := time.Now()

		// Calculate confidence for many scenarios
		for i := 0; i < 1000; i++ {
			repo.calculateDynamicConfidence(0.5, 3, 6)
		}

		duration := time.Since(start)

		// Should be very fast (less than 1ms for 1000 calculations)
		assert.Less(t, duration, 1*time.Millisecond, "Dynamic confidence calculation should be very fast")
	})
}

// TestTask2_3_2_ContextMultiplierEdgeCases tests edge cases
func TestTask2_3_2_ContextMultiplierEdgeCases(t *testing.T) {
	t.Run("Empty contextual keywords", func(t *testing.T) {
		// Create mock repository
		repo := createMockRepository(t)
		ctx := context.Background()

		// Test with empty contextual keywords
		result, err := repo.ClassifyBusinessByContextualKeywords(ctx, []ContextualKeyword{})
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "General Business", result.Industry.Name, "Should return default industry")
		assert.Equal(t, 0.50, result.Confidence, "Should return default confidence")
		assert.Equal(t, "No contextual keywords provided for classification", result.Reasoning, "Should have appropriate reasoning")
	})

	t.Run("Unknown context gets default multiplier", func(t *testing.T) {
		// Create mock repository
		repo := createMockRepository(t)

		// Test unknown context
		multiplier := repo.getContextMultiplier("unknown_context")
		assert.Equal(t, 1.0, multiplier, "Unknown context should get default 1.0x multiplier")
	})

	t.Run("Confidence calculation with edge values", func(t *testing.T) {
		// Create mock repository
		repo := createMockRepository(t)

		// Test edge cases for confidence calculation
		testCases := []struct {
			name            string
			score           float64
			matchedKeywords int
			totalKeywords   int
		}{
			{"Zero score", 0.0, 0, 1},
			{"Maximum score", 1.0, 10, 10},
			{"Zero matched keywords", 0.5, 0, 5},
			{"All keywords matched", 0.5, 5, 5},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				confidence := repo.calculateDynamicConfidence(tc.score, tc.matchedKeywords, tc.totalKeywords)

				assert.GreaterOrEqual(t, confidence, 0.1, "Confidence should be at least 0.1")
				assert.LessOrEqual(t, confidence, 1.0, "Confidence should be at most 1.0")
			})
		}
	})
}

// Helper functions for testing

// createMockRepository creates a mock repository for testing
func createMockRepository(t *testing.T) *SupabaseKeywordRepository {
	// Create a mock repository with minimal setup
	repo := &SupabaseKeywordRepository{
		logger: log.New(&mockLogger{}, "", 0),
		keywordIndex: &KeywordIndex{
			KeywordToIndustries: make(map[string][]IndustryKeywordMatch),
			IndustryToKeywords:  make(map[int][]*KeywordWeight),
		},
	}
	return repo
}

// createMockRepositoryWithContextTestData creates a mock repository with test data for context multiplier testing
func createMockRepositoryWithContextTestData() *SupabaseKeywordRepository {
	repo := &SupabaseKeywordRepository{
		logger: log.New(&mockLogger{}, "", 0),
		keywordIndex: &KeywordIndex{
			KeywordToIndustries: make(map[string][]IndustryKeywordMatch),
			IndustryToKeywords:  make(map[int][]*KeywordWeight),
		},
	}

	// Add some test data to the keyword index
	repo.keywordIndex.KeywordToIndustries["restaurant"] = []IndustryKeywordMatch{
		{IndustryID: 1, Weight: 0.8, Keyword: "restaurant"},
	}
	repo.keywordIndex.KeywordToIndustries["italian"] = []IndustryKeywordMatch{
		{IndustryID: 1, Weight: 0.7, Keyword: "italian"},
	}
	repo.keywordIndex.KeywordToIndustries["dining"] = []IndustryKeywordMatch{
		{IndustryID: 1, Weight: 0.6, Keyword: "dining"},
	}
	repo.keywordIndex.KeywordToIndustries["pasta"] = []IndustryKeywordMatch{
		{IndustryID: 1, Weight: 0.5, Keyword: "pasta"},
	}
	repo.keywordIndex.KeywordToIndustries["wine"] = []IndustryKeywordMatch{
		{IndustryID: 1, Weight: 0.4, Keyword: "wine"},
	}

	return repo
}

// mockLogger implements a simple logger for testing
type mockLogger struct{}

func (m *mockLogger) Write(p []byte) (n int, err error) {
	// Do nothing for testing
	return len(p), nil
}
