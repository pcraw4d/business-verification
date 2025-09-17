package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestTask2_3_1_PhraseMatching tests the phrase matching functionality
func TestTask2_3_1_PhraseMatching(t *testing.T) {
	// Create a test repository instance
	repo := &SupabaseKeywordRepository{}

	t.Run("extractKeywordsAndPhrases", func(t *testing.T) {
		text := "fast food restaurant serving burgers and fries"
		keywords := repo.extractKeywordsAndPhrases(text)

		// Should extract individual words
		assert.Contains(t, keywords, "fast")
		assert.Contains(t, keywords, "food")
		assert.Contains(t, keywords, "restaurant")
		assert.Contains(t, keywords, "serving")
		assert.Contains(t, keywords, "burgers")
		assert.Contains(t, keywords, "fries")

		// Should extract 2-word phrases
		assert.Contains(t, keywords, "fast food")
		assert.Contains(t, keywords, "food restaurant")
		assert.Contains(t, keywords, "restaurant serving")
		assert.Contains(t, keywords, "serving burgers")
		assert.Contains(t, keywords, "burgers and")

		// Should extract 3-word phrases
		assert.Contains(t, keywords, "fast food restaurant")
		assert.Contains(t, keywords, "food restaurant serving")
		assert.Contains(t, keywords, "restaurant serving burgers")
	})

	t.Run("extractPhrases", func(t *testing.T) {
		text := "fine dining italian restaurant"

		// Test 2-word phrases
		phrases2 := repo.extractPhrases(text, 2)
		expected2 := []string{"fine dining", "dining italian", "italian restaurant"}
		assert.ElementsMatch(t, expected2, phrases2)

		// Test 3-word phrases
		phrases3 := repo.extractPhrases(text, 3)
		expected3 := []string{"fine dining italian", "dining italian restaurant"}
		assert.ElementsMatch(t, expected3, phrases3)
	})

	t.Run("isValidPhrase", func(t *testing.T) {
		// Valid phrases
		assert.True(t, repo.isValidPhrase("fast food"))
		assert.True(t, repo.isValidPhrase("fine dining"))
		assert.True(t, repo.isValidPhrase("italian restaurant"))
		assert.True(t, repo.isValidPhrase("software development"))

		// Invalid phrases (too short)
		assert.False(t, repo.isValidPhrase("a"))
		assert.False(t, repo.isValidPhrase("ab"))

		// Invalid phrases (only common words)
		assert.False(t, repo.isValidPhrase("the and"))
		assert.False(t, repo.isValidPhrase("in the"))
	})

	t.Run("isCommonWord", func(t *testing.T) {
		// Common words
		assert.True(t, repo.isCommonWord("the"))
		assert.True(t, repo.isCommonWord("and"))
		assert.True(t, repo.isCommonWord("in"))
		assert.True(t, repo.isCommonWord("www"))
		assert.True(t, repo.isCommonWord("com"))

		// Business-relevant words
		assert.False(t, repo.isCommonWord("restaurant"))
		assert.False(t, repo.isCommonWord("software"))
		assert.False(t, repo.isCommonWord("medical"))
		assert.False(t, repo.isCommonWord("legal"))
	})

	t.Run("hasPhraseOverlap", func(t *testing.T) {
		// Phrases with overlap
		assert.True(t, repo.hasPhraseOverlap("fast food", "food service"))
		assert.True(t, repo.hasPhraseOverlap("italian restaurant", "restaurant business"))
		assert.True(t, repo.hasPhraseOverlap("software development", "development company"))

		// Phrases without meaningful overlap
		assert.False(t, repo.hasPhraseOverlap("fast food", "legal services"))
		assert.False(t, repo.hasPhraseOverlap("medical practice", "software company"))

		// Phrases with only common word overlap
		assert.False(t, repo.hasPhraseOverlap("the restaurant", "the company"))
	})
}

// TestTask2_3_1_PhraseMatchingIntegration tests phrase matching integration
func TestTask2_3_1_PhraseMatchingIntegration(t *testing.T) {
	repo := &SupabaseKeywordRepository{}

	t.Run("extractKeywords with phrase extraction", func(t *testing.T) {
		keywords := repo.extractKeywords("Mario's Italian Bistro", "Fine dining Italian restaurant serving authentic pasta and wine", "")

		// Should extract individual words
		assert.Contains(t, keywords, "mario's")
		assert.Contains(t, keywords, "italian")
		assert.Contains(t, keywords, "bistro")
		assert.Contains(t, keywords, "fine")
		assert.Contains(t, keywords, "dining")
		assert.Contains(t, keywords, "restaurant")
		assert.Contains(t, keywords, "serving")
		assert.Contains(t, keywords, "authentic")
		assert.Contains(t, keywords, "pasta")
		assert.Contains(t, keywords, "wine")

		// Should extract 2-word phrases
		assert.Contains(t, keywords, "fine dining")
		assert.Contains(t, keywords, "italian restaurant")
		assert.Contains(t, keywords, "authentic pasta")

		// Should extract 3-word phrases
		assert.Contains(t, keywords, "fine dining italian")
		assert.Contains(t, keywords, "italian restaurant serving")
	})

	t.Run("phrase matching with business classification", func(t *testing.T) {
		// Test with restaurant business
		keywords := repo.extractKeywords("McDonald's", "Fast food restaurant chain serving burgers and fries", "")

		// Should extract phrase "fast food"
		assert.Contains(t, keywords, "fast food")
		assert.Contains(t, keywords, "restaurant chain")
		assert.Contains(t, keywords, "burgers and")
	})

	t.Run("phrase matching with technology business", func(t *testing.T) {
		// Test with technology business
		keywords := repo.extractKeywords("TechCorp Solutions", "Software development company specializing in enterprise applications", "")

		// Should extract phrases
		assert.Contains(t, keywords, "software development")
		assert.Contains(t, keywords, "development company")
		assert.Contains(t, keywords, "enterprise applications")
	})
}

// TestTask2_3_1_PhraseMatchingPerformance tests performance of phrase matching
func TestTask2_3_1_PhraseMatchingPerformance(t *testing.T) {
	repo := &SupabaseKeywordRepository{}

	t.Run("performance with large text", func(t *testing.T) {
		// Create a large text with many phrases
		largeText := "fast food restaurant chain serving burgers fries chicken sandwiches salads drinks desserts breakfast lunch dinner drive through takeout delivery catering corporate events birthday parties family gatherings quick service casual dining affordable prices quality food fresh ingredients local suppliers sustainable practices community involvement charitable donations employee training customer service satisfaction"

		keywords := repo.extractKeywordsAndPhrases(largeText)

		// Should extract many keywords and phrases
		assert.Greater(t, len(keywords), 50)

		// Should include specific phrases
		assert.Contains(t, keywords, "fast food")
		assert.Contains(t, keywords, "restaurant chain")
		assert.Contains(t, keywords, "drive through")
		assert.Contains(t, keywords, "customer service")
	})

	t.Run("performance with repeated phrases", func(t *testing.T) {
		// Text with repeated phrases
		repeatedText := "software development software development software development company"

		keywords := repo.extractKeywordsAndPhrases(repeatedText)

		// Should deduplicate phrases
		phraseCount := 0
		for _, keyword := range keywords {
			if keyword == "software development" {
				phraseCount++
			}
		}
		assert.Equal(t, 1, phraseCount) // Should appear only once
	})
}

// TestTask2_3_1_PhraseMatchingEdgeCases tests edge cases for phrase matching
func TestTask2_3_1_PhraseMatchingEdgeCases(t *testing.T) {
	repo := &SupabaseKeywordRepository{}

	t.Run("empty text", func(t *testing.T) {
		keywords := repo.extractKeywords("", "", "")
		assert.Empty(t, keywords)
	})

	t.Run("single word", func(t *testing.T) {
		keywords := repo.extractKeywords("restaurant", "", "")
		assert.Contains(t, keywords, "restaurant")
		assert.Len(t, keywords, 1)
	})

	t.Run("text with special characters", func(t *testing.T) {
		keywords := repo.extractKeywords("Café & Bistro", "French café serving coffee, pastries & light meals", "")

		// Should handle special characters
		assert.Contains(t, keywords, "café")
		assert.Contains(t, keywords, "bistro")
		assert.Contains(t, keywords, "french")
		assert.Contains(t, keywords, "coffee")
		assert.Contains(t, keywords, "pastries")
		assert.Contains(t, keywords, "light")
		assert.Contains(t, keywords, "meals")

		// Should extract phrases
		assert.Contains(t, keywords, "french café")
		assert.Contains(t, keywords, "light meals")
	})

	t.Run("text with numbers", func(t *testing.T) {
		keywords := repo.extractKeywords("24/7 Fitness", "24 hour fitness center with state-of-the-art equipment", "")

		// Should handle numbers
		assert.Contains(t, keywords, "24/7")
		assert.Contains(t, keywords, "fitness")
		assert.Contains(t, keywords, "center")
		assert.Contains(t, keywords, "equipment")

		// Should extract phrases
		assert.Contains(t, keywords, "fitness center")
	})

	t.Run("text with very long phrases", func(t *testing.T) {
		keywords := repo.extractKeywords("", "This is a very long business description that contains many words and should still work correctly with our phrase extraction algorithm", "")

		// Should extract 3-word phrases
		assert.Contains(t, keywords, "very long business")
		assert.Contains(t, keywords, "long business description")
		assert.Contains(t, keywords, "business description that")
	})
}

// TestTask2_3_1_PhraseMatchingBusinessScenarios tests real business scenarios
func TestTask2_3_1_PhraseMatchingBusinessScenarios(t *testing.T) {
	repo := &SupabaseKeywordRepository{}

	testCases := []struct {
		name            string
		businessName    string
		description     string
		expectedPhrases []string
	}{
		{
			name:            "Italian Restaurant",
			businessName:    "Mario's Italian Bistro",
			description:     "Fine dining Italian restaurant serving authentic pasta and wine",
			expectedPhrases: []string{"fine dining", "italian restaurant", "authentic pasta"},
		},
		{
			name:            "Fast Food Chain",
			businessName:    "McDonald's",
			description:     "Fast food restaurant chain serving burgers and fries",
			expectedPhrases: []string{"fast food", "restaurant chain", "burgers and"},
		},
		{
			name:            "Software Company",
			businessName:    "TechCorp Solutions",
			description:     "Software development company specializing in enterprise applications",
			expectedPhrases: []string{"software development", "development company", "enterprise applications"},
		},
		{
			name:            "Medical Practice",
			businessName:    "Downtown Medical Center",
			description:     "Full service medical practice providing primary care and specialty services",
			expectedPhrases: []string{"full service", "medical practice", "primary care", "specialty services"},
		},
		{
			name:            "Legal Firm",
			businessName:    "Smith & Associates",
			description:     "Law firm specializing in corporate law and business litigation",
			expectedPhrases: []string{"law firm", "corporate law", "business litigation"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keywords := repo.extractKeywords(tc.businessName, tc.description, "")

			// Check that expected phrases are extracted
			for _, expectedPhrase := range tc.expectedPhrases {
				assert.Contains(t, keywords, expectedPhrase,
					"Expected phrase '%s' not found in keywords for %s", expectedPhrase, tc.name)
			}
		})
	}
}
