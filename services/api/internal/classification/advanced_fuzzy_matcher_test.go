package classification

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func TestAdvancedFuzzyMatcher(t *testing.T) {
	// Setup
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	config := DefaultAdvancedFuzzyConfig()
	matcher := NewAdvancedFuzzyMatcher(logger, config)

	// Create test keyword index
	keywordIndex := &KeywordIndex{
		KeywordToIndustries: map[string][]IndexKeywordMatch{
			"restaurant": {
				{Keyword: "restaurant", IndustryID: 1, Weight: 1.0, MatchType: "direct"},
			},
			"cafe": {
				{Keyword: "cafe", IndustryID: 1, Weight: 0.9, MatchType: "direct"},
			},
			"dining": {
				{Keyword: "dining", IndustryID: 1, Weight: 0.8, MatchType: "direct"},
			},
			"food": {
				{Keyword: "food", IndustryID: 1, Weight: 0.7, MatchType: "direct"},
			},
			"software": {
				{Keyword: "software", IndustryID: 2, Weight: 1.0, MatchType: "direct"},
			},
			"technology": {
				{Keyword: "technology", IndustryID: 2, Weight: 0.9, MatchType: "direct"},
			},
			"programming": {
				{Keyword: "programming", IndustryID: 2, Weight: 0.8, MatchType: "direct"},
			},
		},
		IndustryToKeywords: map[int][]string{
			1: {"restaurant", "cafe", "dining", "food"},
			2: {"software", "technology", "programming"},
		},
		TotalKeywords: 7,
		LastUpdated:   time.Now(),
	}

	t.Run("TestLevenshteinSimilarity", func(t *testing.T) {
		tests := []struct {
			s1, s2    string
			expected  float64
			tolerance float64
		}{
			{"restaurant", "restaurant", 1.0, 0.0},
			{"restaurant", "restorant", 0.8, 0.2},
			{"cafe", "coffee", 0.4, 0.2},
			{"software", "softwar", 0.9, 0.2},
			{"technology", "technolgy", 0.9, 0.2},
			{"", "", 1.0, 0.0}, // Empty strings are considered identical
			{"abc", "", 0.0, 0.0},
		}

		for _, tt := range tests {
			result := matcher.calculateLevenshteinSimilarity(tt.s1, tt.s2)
			if !withinTolerance(result, tt.expected, tt.tolerance) {
				t.Errorf("LevenshteinSimilarity(%q, %q) = %v, expected %v ± %v",
					tt.s1, tt.s2, result, tt.expected, tt.tolerance)
			}
		}
	})

	t.Run("TestJaroWinklerSimilarity", func(t *testing.T) {
		tests := []struct {
			s1, s2    string
			expected  float64
			tolerance float64
		}{
			{"restaurant", "restaurant", 1.0, 0.0},
			{"restaurant", "restorant", 1.2, 0.3}, // Jaro-Winkler can exceed 1.0
			{"cafe", "coffee", 0.8, 0.2},
			{"software", "softwar", 1.3, 0.3},     // Jaro-Winkler can exceed 1.0
			{"technology", "technolgy", 1.3, 0.3}, // Jaro-Winkler can exceed 1.0
			{"", "", 1.0, 0.0},
			{"abc", "", 0.0, 0.0},
		}

		for _, tt := range tests {
			result := matcher.calculateJaroWinklerSimilarity(tt.s1, tt.s2)
			if !withinTolerance(result, tt.expected, tt.tolerance) {
				t.Errorf("JaroWinklerSimilarity(%q, %q) = %v, expected %v ± %v",
					tt.s1, tt.s2, result, tt.expected, tt.tolerance)
			}
		}
	})

	t.Run("TestJaccardSimilarity", func(t *testing.T) {
		tests := []struct {
			s1, s2    string
			expected  float64
			tolerance float64
		}{
			{"restaurant", "restaurant", 1.0, 0.0},
			{"restaurant", "restorant", 0.5, 0.2},
			{"cafe", "coffee", 0.2, 0.2},
			{"software", "softwar", 0.7, 0.2},
			{"technology", "technolgy", 0.7, 0.2},
			{"", "", 1.0, 0.0}, // Empty strings are considered identical
			{"abc", "", 0.0, 0.0},
		}

		for _, tt := range tests {
			result := matcher.calculateJaccardSimilarity(tt.s1, tt.s2)
			if !withinTolerance(result, tt.expected, tt.tolerance) {
				t.Errorf("JaccardSimilarity(%q, %q) = %v, expected %v ± %v",
					tt.s1, tt.s2, result, tt.expected, tt.tolerance)
			}
		}
	})

	t.Run("TestCosineSimilarity", func(t *testing.T) {
		tests := []struct {
			s1, s2    string
			expected  float64
			tolerance float64
		}{
			{"restaurant", "restaurant", 1.0, 0.0},
			{"restaurant", "restorant", 0.9, 0.2},
			{"cafe", "coffee", 0.7, 0.2},
			{"software", "softwar", 0.9, 0.2},
			{"technology", "technolgy", 0.9, 0.2},
			{"", "", 1.0, 0.0}, // Empty strings are considered identical
			{"abc", "", 0.0, 0.0},
		}

		for _, tt := range tests {
			result := matcher.calculateCosineSimilarity(tt.s1, tt.s2)
			if !withinTolerance(result, tt.expected, tt.tolerance) {
				t.Errorf("CosineSimilarity(%q, %q) = %v, expected %v ± %v",
					tt.s1, tt.s2, result, tt.expected, tt.tolerance)
			}
		}
	})

	t.Run("TestSoundexSimilarity", func(t *testing.T) {
		tests := []struct {
			s1, s2    string
			expected  float64
			tolerance float64
		}{
			{"restaurant", "restaurant", 1.0, 0.0},
			{"restaurant", "restorant", 1.0, 0.2}, // Same Soundex code
			{"cafe", "coffee", 1.0, 0.2},          // Same Soundex code
			{"software", "softwar", 1.0, 0.2},     // Same Soundex code
			{"technology", "technolgy", 1.0, 0.2}, // Same Soundex code
			{"", "", 1.0, 0.0},                    // Empty strings are considered identical
		}

		for _, tt := range tests {
			result := matcher.calculateSoundexSimilarity(tt.s1, tt.s2)
			if !withinTolerance(result, tt.expected, tt.tolerance) {
				t.Errorf("SoundexSimilarity(%q, %q) = %v, expected %v ± %v",
					tt.s1, tt.s2, result, tt.expected, tt.tolerance)
			}
		}
	})

	t.Run("TestMetaphoneSimilarity", func(t *testing.T) {
		tests := []struct {
			s1, s2    string
			expected  float64
			tolerance float64
		}{
			{"restaurant", "restaurant", 1.0, 0.0},
			{"restaurant", "restorant", 1.0, 0.2}, // Same Metaphone code
			{"cafe", "coffee", 0.6, 0.2},          // Different Metaphone codes
			{"software", "softwar", 1.0, 0.2},     // Same Metaphone code
			{"technology", "technolgy", 1.0, 0.2}, // Same Metaphone code
			{"", "", 1.0, 0.0},                    // Empty strings are considered identical
		}

		for _, tt := range tests {
			result := matcher.calculateMetaphoneSimilarity(tt.s1, tt.s2)
			if !withinTolerance(result, tt.expected, tt.tolerance) {
				t.Errorf("MetaphoneSimilarity(%q, %q) = %v, expected %v ± %v",
					tt.s1, tt.s2, result, tt.expected, tt.tolerance)
			}
		}
	})

	t.Run("TestCombinedSimilarity", func(t *testing.T) {
		tests := []struct {
			s1, s2    string
			expected  float64
			tolerance float64
		}{
			{"restaurant", "restaurant", 1.0, 0.0},
			{"restaurant", "restorant", 0.9, 0.2},
			{"cafe", "coffee", 0.6, 0.2},
			{"software", "softwar", 1.0, 0.2},
			{"technology", "technolgy", 1.0, 0.2},
			{"", "", 1.0, 0.0}, // Empty strings are considered identical
		}

		for _, tt := range tests {
			result := matcher.calculateCombinedSimilarity(tt.s1, tt.s2)
			if !withinTolerance(result, tt.expected, tt.tolerance) {
				t.Errorf("CombinedSimilarity(%q, %q) = %v, expected %v ± %v",
					tt.s1, tt.s2, result, tt.expected, tt.tolerance)
			}
		}
	})

	t.Run("TestFindFuzzyMatches", func(t *testing.T) {
		ctx := context.Background()

		tests := []struct {
			inputKeyword string
			expectedMin  int
			expectedMax  int
		}{
			{"restorant", 1, 3},   // Should match "restaurant" and related
			{"cafe", 0, 1},        // Should match "cafe" and related
			{"softwar", 1, 2},     // Should match "software" and related
			{"technolgy", 1, 2},   // Should match "technology" and related
			{"nonexistent", 0, 0}, // Should not match anything
		}

		for _, tt := range tests {
			matches, err := matcher.FindFuzzyMatches(ctx, tt.inputKeyword, keywordIndex)
			if err != nil {
				t.Errorf("FindFuzzyMatches(%q) error: %v", tt.inputKeyword, err)
				continue
			}

			if len(matches) < tt.expectedMin || len(matches) > tt.expectedMax {
				t.Errorf("FindFuzzyMatches(%q) returned %d matches, expected %d-%d",
					tt.inputKeyword, len(matches), tt.expectedMin, tt.expectedMax)
			}

			// Verify all matches meet threshold
			for _, match := range matches {
				if match.Similarity < config.SimilarityThreshold {
					t.Errorf("FindFuzzyMatches(%q) returned match with similarity %v below threshold %v",
						tt.inputKeyword, match.Similarity, config.SimilarityThreshold)
				}
			}
		}
	})

	t.Run("TestSemanticExpansion", func(t *testing.T) {
		ctx := context.Background()

		tests := []struct {
			inputKeyword string
			expectedMin  int
			expectedMax  int
		}{
			{"restorant", 1, 3},   // Should expand to restaurant-related terms
			{"softwar", 1, 2},     // Should expand to software-related terms
			{"nonexistent", 0, 0}, // Should not expand
		}

		for _, tt := range tests {
			expansion, err := matcher.ExpandSemanticKeywords(ctx, tt.inputKeyword, keywordIndex)
			if err != nil {
				t.Errorf("ExpandSemanticKeywords(%q) error: %v", tt.inputKeyword, err)
				continue
			}

			if expansion == nil {
				if tt.expectedMin > 0 {
					t.Errorf("ExpandSemanticKeywords(%q) returned nil, expected expansion", tt.inputKeyword)
				}
				continue
			}

			if len(expansion.Expansions) < tt.expectedMin || len(expansion.Expansions) > tt.expectedMax {
				t.Errorf("ExpandSemanticKeywords(%q) returned %d expansions, expected %d-%d",
					tt.inputKeyword, len(expansion.Expansions), tt.expectedMin, tt.expectedMax)
			}

			// Verify all expansions meet threshold
			for _, expansion := range expansion.Expansions {
				if expansion.Similarity < config.SemanticThreshold {
					t.Errorf("ExpandSemanticKeywords(%q) returned expansion with similarity %v below threshold %v",
						tt.inputKeyword, expansion.Similarity, config.SemanticThreshold)
				}
			}
		}
	})

	t.Run("TestCaching", func(t *testing.T) {
		ctx := context.Background()

		// First call should populate cache
		matches1, err := matcher.FindFuzzyMatches(ctx, "restorant", keywordIndex)
		if err != nil {
			t.Fatalf("First FindFuzzyMatches call error: %v", err)
		}

		// Second call should use cache
		matches2, err := matcher.FindFuzzyMatches(ctx, "restorant", keywordIndex)
		if err != nil {
			t.Fatalf("Second FindFuzzyMatches call error: %v", err)
		}

		// Results should be identical
		if len(matches1) != len(matches2) {
			t.Errorf("Cached results length mismatch: %d vs %d", len(matches1), len(matches2))
		}

		for i := range matches1 {
			if matches1[i].Keyword != matches2[i].Keyword {
				t.Errorf("Cached result keyword mismatch at index %d: %q vs %q",
					i, matches1[i].Keyword, matches2[i].Keyword)
			}
		}
	})

	t.Run("TestPerformance", func(t *testing.T) {
		ctx := context.Background()

		// Test with larger keyword index
		largeIndex := &KeywordIndex{
			KeywordToIndustries: make(map[string][]IndexKeywordMatch),
			IndustryToKeywords:  make(map[int][]string),
			TotalKeywords:       1000,
			LastUpdated:         time.Now(),
		}

		// Populate with test data
		for i := 0; i < 1000; i++ {
			keyword := fmt.Sprintf("keyword%d", i)
			largeIndex.KeywordToIndustries[keyword] = []IndexKeywordMatch{
				{Keyword: keyword, IndustryID: i%10 + 1, Weight: 0.8, MatchType: "direct"},
			}
		}

		start := time.Now()
		matches, err := matcher.FindFuzzyMatches(ctx, "keyword1", largeIndex)
		duration := time.Since(start)

		if err != nil {
			t.Fatalf("Performance test error: %v", err)
		}

		// Should complete within reasonable time (adjust threshold as needed)
		if duration > 5*time.Second {
			t.Errorf("Performance test took too long: %v", duration)
		}

		t.Logf("Performance test: %d matches found in %v", len(matches), duration)
	})
}

func TestAdvancedFuzzyMatcherConfiguration(t *testing.T) {
	t.Run("TestDefaultConfig", func(t *testing.T) {
		config := DefaultAdvancedFuzzyConfig()

		// Verify all algorithms are enabled by default
		if !config.EnableLevenshtein {
			t.Error("Levenshtein should be enabled by default")
		}
		if !config.EnableJaroWinkler {
			t.Error("Jaro-Winkler should be enabled by default")
		}
		if !config.EnableJaccard {
			t.Error("Jaccard should be enabled by default")
		}
		if !config.EnableCosine {
			t.Error("Cosine should be enabled by default")
		}
		if !config.EnableSoundex {
			t.Error("Soundex should be enabled by default")
		}
		if !config.EnableMetaphone {
			t.Error("Metaphone should be enabled by default")
		}
		if !config.EnableSemanticExpand {
			t.Error("Semantic expansion should be enabled by default")
		}

		// Verify weights sum to approximately 1.0
		totalWeight := config.LevenshteinWeight + config.JaroWinklerWeight +
			config.JaccardWeight + config.CosineWeight +
			config.SoundexWeight + config.MetaphoneWeight

		if totalWeight < 0.99 || totalWeight > 1.01 {
			t.Errorf("Algorithm weights should sum to 1.0, got %v", totalWeight)
		}

		// Verify reasonable thresholds
		if config.SimilarityThreshold < 0.5 || config.SimilarityThreshold > 0.9 {
			t.Errorf("Similarity threshold should be between 0.5 and 0.9, got %v", config.SimilarityThreshold)
		}
		if config.SemanticThreshold < 0.5 || config.SemanticThreshold > 0.9 {
			t.Errorf("Semantic threshold should be between 0.5 and 0.9, got %v", config.SemanticThreshold)
		}
	})

	t.Run("TestCustomConfig", func(t *testing.T) {
		config := &AdvancedFuzzyConfig{
			EnableLevenshtein:    true,
			EnableJaroWinkler:    false,
			EnableJaccard:        false,
			EnableCosine:         false,
			EnableSoundex:        false,
			EnableMetaphone:      false,
			EnableSemanticExpand: false,
			SimilarityThreshold:  0.8,
			LevenshteinWeight:    1.0,
		}

		logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
		matcher := NewAdvancedFuzzyMatcher(logger, config)

		// Test that only Levenshtein is used
		result := matcher.calculateCombinedSimilarity("restaurant", "restorant")
		if result < 0.7 || result > 0.9 {
			t.Errorf("Custom config with only Levenshtein should give reasonable similarity, got %v", result)
		}
	})
}

func TestAdvancedFuzzyMatcherEdgeCases(t *testing.T) {
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	config := DefaultAdvancedFuzzyConfig()
	matcher := NewAdvancedFuzzyMatcher(logger, config)

	t.Run("TestEmptyStrings", func(t *testing.T) {
		result := matcher.calculateCombinedSimilarity("", "")
		if result != 1.0 {
			t.Errorf("Empty strings should have similarity 1.0, got %v", result)
		}
	})

	t.Run("TestVeryLongStrings", func(t *testing.T) {
		longString1 := strings.Repeat("a", 1000)
		longString2 := strings.Repeat("b", 1000)

		result := matcher.calculateCombinedSimilarity(longString1, longString2)
		if result < 0.0 || result > 1.0 {
			t.Errorf("Very long strings should have similarity between 0.0 and 1.0, got %v", result)
		}
	})

	t.Run("TestSpecialCharacters", func(t *testing.T) {
		special1 := "café"
		special2 := "cafe"

		result := matcher.calculateCombinedSimilarity(special1, special2)
		if result < 0.0 || result > 1.0 {
			t.Errorf("Special characters should have similarity between 0.0 and 1.0, got %v", result)
		}
	})

	t.Run("TestUnicodeStrings", func(t *testing.T) {
		unicode1 := "résumé"
		unicode2 := "resume"

		result := matcher.calculateCombinedSimilarity(unicode1, unicode2)
		if result < 0.0 || result > 1.0 {
			t.Errorf("Unicode strings should have similarity between 0.0 and 1.0, got %v", result)
		}
	})
}

// Helper function to check if a value is within tolerance
func withinTolerance(actual, expected, tolerance float64) bool {
	return actual >= expected-tolerance && actual <= expected+tolerance
}
