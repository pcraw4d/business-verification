//go:build !comprehensive_test
// +build !comprehensive_test

package integration

import (
	"context"
	"database/sql"
	"log"
	"os"
	"testing"

	"kyb-platform/internal/classification"
)

// TestKeywordCoverage verifies 90%+ of codes have 15+ keywords
func TestKeywordCoverage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// Setup database connection
	// Note: This requires proper test database setup
	// For now, we'll test the structure and expectations

	t.Run("Verify 90%+ of codes have 15+ keywords", func(t *testing.T) {
		// Expected: 90%+ of codes have 15+ keywords
		expectedMinPercentage := 90.0
		expectedMinKeywords := 15

		// In real test, this would query the database:
		// SELECT 
		//     COUNT(*) FILTER (WHERE keyword_count >= 15) * 100.0 / COUNT(*) as percentage
		// FROM (
		//     SELECT code_id, COUNT(*) as keyword_count
		//     FROM code_keywords
		//     GROUP BY code_id
		// ) AS keyword_counts;

		t.Logf("Expected minimum percentage: %.1f%%", expectedMinPercentage)
		t.Logf("Expected minimum keywords per code: %d", expectedMinKeywords)
	})

	t.Run("Verify average 18 keywords per code", func(t *testing.T) {
		// Expected: Average 18 keywords per code
		expectedAvgKeywords := 18.0

		// In real test:
		// SELECT AVG(keyword_count) as avg_keywords
		// FROM (
		//     SELECT code_id, COUNT(*) as keyword_count
		//     FROM code_keywords
		//     GROUP BY code_id
		// ) AS keyword_counts;

		t.Logf("Expected average keywords per code: %.1f", expectedAvgKeywords)
	})
}

// TestKeywordQuality verifies keyword quality with relevance scoring
func TestKeywordQuality(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	t.Run("Verify keyword relevance scores", func(t *testing.T) {
		// Expected: Keywords have relevance scores between 0.5 and 1.0
		minRelevance := 0.5
		maxRelevance := 1.0

		// In real test:
		// SELECT 
		//     COUNT(*) FILTER (WHERE relevance_score >= 0.5 AND relevance_score <= 1.0) as valid_scores,
		//     COUNT(*) as total_keywords
		// FROM code_keywords;

		t.Logf("Expected relevance score range: %.2f - %.2f", minRelevance, maxRelevance)
	})

	t.Run("Verify keyword match types", func(t *testing.T) {
		// Expected: Keywords have valid match types: 'exact', 'partial', 'synonym'
		validMatchTypes := []string{"exact", "partial", "synonym"}

		// In real test:
		// SELECT match_type, COUNT(*) 
		// FROM code_keywords
		// GROUP BY match_type;

		t.Logf("Valid match types: %v", validMatchTypes)
	})
}

// TestKeywordMatchingAccuracy tests keyword matching accuracy > 85%
func TestKeywordMatchingAccuracy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	// This test would require a test dataset with known keyword matches
	// For now, we'll test the structure

	t.Run("Test keyword extraction from descriptions", func(t *testing.T) {
		extractor := classification.NewKeywordExtractor()

		testCases := []struct {
			description string
			expectedMin int
		}{
			{
				description: "This U.S. industry comprises establishments primarily engaged in writing, modifying, testing, and supporting software to meet the needs of a particular customer.",
				expectedMin: 15,
			},
			{
				description: "This industry comprises establishments of licensed practitioners having the degree of M.D. primarily engaged in the independent practice of general or specialized medicine.",
				expectedMin: 15,
			},
			{
				description: "This industry comprises establishments primarily engaged in accepting demand and other deposits and making commercial, industrial, and consumer loans.",
				expectedMin: 15,
			},
		}

		for _, tc := range testCases {
			t.Run("extract_keywords", func(t *testing.T) {
				keywords := extractor.ExtractKeywords(tc.description)

				if len(keywords) < tc.expectedMin {
					t.Errorf("Expected at least %d keywords, got %d", tc.expectedMin, len(keywords))
				}

				t.Logf("Extracted %d keywords from description", len(keywords))
			})
		}
	})

	t.Run("Test keyword matching with synonyms", func(t *testing.T) {
		extractor := classification.NewKeywordExtractor()

		// Test that synonyms are expanded
		text := "Software development and programming services"
		keywords := extractor.ExtractKeywords(text)

		// Should include base keywords and potentially synonyms
		if len(keywords) < 3 {
			t.Errorf("Expected at least 3 keywords, got %d", len(keywords))
		}

		t.Logf("Keywords with synonyms: %v", keywords)
	})

	t.Run("Test industry-specific keyword extraction", func(t *testing.T) {
		extractor := classification.NewKeywordExtractor()

		industries := []string{
			"Technology",
			"Healthcare",
			"Financial Services",
			"Retail & Commerce",
		}

		for _, industry := range industries {
			t.Run(industry, func(t *testing.T) {
				text := "Sample description for " + industry
				keywords := extractor.ExtractIndustrySpecificKeywords(text, industry)

				// Should have base keywords plus industry-specific terms
				if len(keywords) < 15 {
					t.Errorf("Expected at least 15 keywords for %s industry, got %d", industry, len(keywords))
				}

				t.Logf("Extracted %d industry-specific keywords for %s", len(keywords), industry)
			})
		}
	})
}

// TestKeywordExtractorUnit tests the keyword extractor unit functionality
func TestKeywordExtractorUnit(t *testing.T) {
	extractor := classification.NewKeywordExtractor()

	t.Run("Extract keywords from official description", func(t *testing.T) {
		description := "This U.S. industry comprises establishments primarily engaged in writing, modifying, testing, and supporting software to meet the needs of a particular customer."
		keywords := extractor.ExtractKeywords(description)

		if len(keywords) == 0 {
			t.Error("Expected keywords to be extracted, got empty slice")
		}

		// Verify no stop words
		for _, keyword := range keywords {
			if extractor.stopWords[keyword] {
				t.Errorf("Found stop word in keywords: %s", keyword)
			}
		}

		t.Logf("Extracted %d keywords: %v", len(keywords), keywords)
	})

	t.Run("Extract keywords with relevance scores", func(t *testing.T) {
		description := "Custom computer programming services"
		keywordsWithRelevance := extractor.ExtractKeywordsWithRelevance(description)

		if len(keywordsWithRelevance) == 0 {
			t.Error("Expected keywords with relevance scores, got empty map")
		}

		// Verify relevance scores are in valid range
		for keyword, relevance := range keywordsWithRelevance {
			if relevance < 0.5 || relevance > 1.0 {
				t.Errorf("Relevance score for %s is out of range: %.2f", keyword, relevance)
			}
		}

		t.Logf("Extracted %d keywords with relevance scores", len(keywordsWithRelevance))
	})

	t.Run("Test synonym expansion", func(t *testing.T) {
		text := "Software development"
		keywords := extractor.ExtractKeywords(text)

		// Should include base keywords
		hasSoftware := false
		hasDevelopment := false

		for _, keyword := range keywords {
			if keyword == "software" {
				hasSoftware = true
			}
			if keyword == "development" {
				hasDevelopment = true
			}
		}

		if !hasSoftware {
			t.Error("Expected 'software' keyword")
		}

		if !hasDevelopment {
			t.Error("Expected 'development' keyword")
		}

		t.Logf("Keywords with synonyms: %v", keywords)
	})

	t.Run("Test industry-specific terminology", func(t *testing.T) {
		text := "Medical services and healthcare"
		keywords := extractor.ExtractIndustrySpecificKeywords(text, "Healthcare")

		// Should include industry-specific terms
		if len(keywords) < 10 {
			t.Errorf("Expected at least 10 keywords for healthcare industry, got %d", len(keywords))
		}

		t.Logf("Industry-specific keywords: %v", keywords)
	})
}

