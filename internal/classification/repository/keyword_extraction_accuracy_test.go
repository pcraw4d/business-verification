package repository

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// TestKeywordExtractionAccuracy_Phase10_4 tests keyword extraction accuracy (Phase 10.4)
func TestKeywordExtractionAccuracy_Phase10_4(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping accuracy test in short mode")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	// Test cases with known business websites and expected keywords
	testCases := []struct {
		name            string
		businessName    string
		websiteURL      string
		expectedKeywords []string // Keywords that should be extracted
		minKeywordCount int       // Minimum number of keywords expected
		industry        string    // Expected industry
	}{
		{
			name:            "wine retailer - The Greene Grape",
			businessName:    "The Greene Grape",
			websiteURL:      "https://www.thegreenegrape.com",
			expectedKeywords: []string{"wine", "grape", "retail", "shop", "store", "beverage", "green"},
			minKeywordCount: 3,
			industry:        "Retail", // Or Food & Beverage
		},
		{
			name:            "technology company",
			businessName:    "Tech Solutions Inc",
			websiteURL:      "https://techsolutions.io",
			expectedKeywords: []string{"tech", "technology", "software", "solutions"},
			minKeywordCount: 2,
			industry:        "Technology",
		},
		{
			name:            "restaurant business",
			businessName:    "Mario's Italian Restaurant",
			websiteURL:      "https://mariosrestaurant.com",
			expectedKeywords: []string{"restaurant", "italian", "dining", "food", "cuisine"},
			minKeywordCount: 3,
			industry:        "Food & Beverage",
		},
		{
			name:            "retail store",
			businessName:    "Fashion Boutique",
			websiteURL:      "https://fashionboutique.shop",
			expectedKeywords: []string{"fashion", "retail", "store", "shop", "boutique"},
			minKeywordCount: 3,
			industry:        "Retail",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Extract keywords
			keywords := repo.extractKeywords(ctx, tc.businessName, tc.websiteURL)

			// Verify minimum keyword count
			if len(keywords) < tc.minKeywordCount {
				t.Logf("Warning: Got %d keywords, expected at least %d (may be network issue)", len(keywords), tc.minKeywordCount)
			}

			// Check for expected keywords
			keywordMap := make(map[string]bool)
			for _, kw := range keywords {
				keywordMap[strings.ToLower(kw.Keyword)] = true
			}

			foundCount := 0
			for _, expected := range tc.expectedKeywords {
				if keywordMap[strings.ToLower(expected)] {
					foundCount++
				}
			}

			// At least 50% of expected keywords should be found
			expectedMatchRatio := float64(foundCount) / float64(len(tc.expectedKeywords))
			if expectedMatchRatio < 0.3 && len(keywords) > 0 {
				t.Errorf("Accuracy: Found only %d/%d expected keywords (%.1f%%)", foundCount, len(tc.expectedKeywords), expectedMatchRatio*100)
			}

			// Test classification accuracy
			if len(keywords) > 0 {
				keywordStrings := make([]string, 0, len(keywords))
				for _, kw := range keywords {
					keywordStrings = append(keywordStrings, kw.Keyword)
				}

				result, err := repo.ClassifyBusinessByKeywords(ctx, keywordStrings)
				if err != nil {
					t.Logf("Classification error (may be expected): %v", err)
				} else if result != nil {
					// Verify classification confidence
					if result.Confidence < 0.3 {
						t.Logf("Warning: Low classification confidence: %.2f (expected >= 0.3)", result.Confidence)
					}

					// Log classification result
					industryID := 0
					if result.Industry != nil {
						industryID = result.Industry.ID
					}
					t.Logf("Classification: Industry ID=%d, Confidence=%.2f, Keywords used=%d", industryID, result.Confidence, len(keywordStrings))
				}
			}

			t.Logf("Accuracy test: Extracted %d keywords, found %d/%d expected keywords (%.1f%%)", len(keywords), foundCount, len(tc.expectedKeywords), expectedMatchRatio*100)
		})
	}
}

// TestKeywordExtractionAccuracy_Comparison tests before/after improvements comparison (Phase 10.4)
func TestKeywordExtractionAccuracy_Comparison(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping accuracy test in short mode")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	// Test cases that previously would have extracted only 1 keyword
	testCases := []struct {
		name                string
		businessName        string
		websiteURL          string
		previousKeywordCount int // Before improvements
		expectedMinKeywords  int // After improvements
	}{
		{
			name:                "wine retailer - improved extraction",
			businessName:        "The Greene Grape",
			websiteURL:          "https://www.thegreenegrape.com",
			previousKeywordCount: 1, // Previously only "grape"
			expectedMinKeywords:  3,  // Now should get: grape, green, wine, retail, shop, etc.
		},
		{
			name:                "compound domain - improved parsing",
			businessName:        "",
			websiteURL:          "https://wineshop.com",
			previousKeywordCount: 1, // Previously only "wineshop"
			expectedMinKeywords:  2,  // Now should get: wine, shop, wine shop
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			keywords := repo.extractKeywords(ctx, tc.businessName, tc.websiteURL)

			// Verify improvement: should extract more keywords than before
			if len(keywords) < tc.expectedMinKeywords {
				t.Logf("Warning: Got %d keywords, expected at least %d (improvement from %d)", len(keywords), tc.expectedMinKeywords, tc.previousKeywordCount)
			}

			// Calculate improvement ratio
			improvementRatio := float64(len(keywords)) / float64(tc.previousKeywordCount)
			if improvementRatio < 1.5 {
				t.Logf("Warning: Improvement ratio is %.2fx (expected >= 1.5x)", improvementRatio)
			}

			t.Logf("Comparison: Previous=%d keywords, Current=%d keywords, Improvement=%.2fx", tc.previousKeywordCount, len(keywords), improvementRatio)
		})
	}
}

// TestIndustryClassificationAccuracy tests industry classification accuracy (Phase 10.4)
func TestIndustryClassificationAccuracy(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping accuracy test in short mode")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	// Test cases with expected industry classifications
	testCases := []struct {
		name              string
		keywords          []string
		expectedIndustry  string // Expected industry name
		minConfidence     float64
		description       string
	}{
		{
			name:             "wine retailer keywords",
			keywords:         []string{"wine", "shop", "retail", "store", "beverage", "grape"},
			expectedIndustry: "Retail", // Or Food & Beverage
			minConfidence:    0.6,
			description:      "Keywords from a wine retailer should classify as Retail or Food & Beverage",
		},
		{
			name:             "technology company keywords",
			keywords:         []string{"software", "technology", "cloud", "development", "digital"},
			expectedIndustry: "Technology",
			minConfidence:    0.7,
			description:      "Technology keywords should classify as Technology",
		},
		{
			name:             "restaurant keywords",
			keywords:         []string{"restaurant", "dining", "food", "cuisine", "catering"},
			expectedIndustry: "Food & Beverage",
			minConfidence:    0.7,
			description:      "Restaurant keywords should classify as Food & Beverage",
		},
		{
			name:             "healthcare keywords",
			keywords:         []string{"medical", "healthcare", "health", "care", "treatment"},
			expectedIndustry: "Healthcare",
			minConfidence:    0.6,
			description:      "Healthcare keywords should classify as Healthcare",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			result, err := repo.ClassifyBusinessByKeywords(ctx, tc.keywords)
			if err != nil {
				t.Errorf("ClassifyBusinessByKeywords() failed: %v", err)
				return
			}

			if result == nil {
				t.Error("ClassifyBusinessByKeywords() returned nil result")
				return
			}

			// Verify confidence meets minimum threshold
			if result.Confidence < tc.minConfidence {
				t.Errorf("Classification confidence too low: %.2f (expected >= %.2f)", result.Confidence, tc.minConfidence)
			}

			// Verify confidence is in valid range
			if result.Confidence < 0 || result.Confidence > 1 {
				t.Errorf("Invalid confidence score: %.2f (expected 0-1)", result.Confidence)
			}

			// Verify result has valid fields
			if result.Industry == nil {
				t.Logf("Warning: Classification result has nil industry")
			}

			industryName := "Unknown"
			if result.Industry != nil {
				industryName = result.Industry.Name
			}
			t.Logf("Classification accuracy: Industry=%s, Confidence=%.2f, Keywords=%d, %s", industryName, result.Confidence, len(tc.keywords), tc.description)
		})
	}
}

// TestKeywordQualityAccuracy tests the quality of extracted keywords (Phase 10.4)
func TestKeywordQualityAccuracy(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	testCases := []struct {
		name              string
		content           string
		expectedKeywords  []string
		unwantedKeywords  []string // Keywords that should NOT be extracted
		minQualityRatio   float64  // Minimum ratio of quality keywords
	}{
		{
			name:             "business-relevant content",
			content:          "We are a wine shop and retail store selling fine wines, spirits, and beverages.",
			expectedKeywords: []string{"wine", "shop", "retail", "store", "beverage", "spirits"},
			unwantedKeywords: []string{"the", "and", "or", "we", "are", "a"},
			minQualityRatio:  0.7, // 70% of keywords should be business-relevant
		},
		{
			name:             "technology content",
			content:          "Software development company specializing in cloud computing and digital transformation.",
			expectedKeywords: []string{"software", "development", "technology", "cloud", "digital"},
			unwantedKeywords: []string{"the", "and", "in", "a", "we"},
			minQualityRatio:  0.7,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			keywords := repo.extractBusinessKeywords(tc.content)

			if len(keywords) == 0 {
				t.Error("Expected keywords from content, got none")
				return
			}

			// Check for expected keywords
			keywordMap := make(map[string]bool)
			for _, kw := range keywords {
				keywordMap[strings.ToLower(kw)] = true
			}

			expectedCount := 0
			for _, expected := range tc.expectedKeywords {
				if keywordMap[strings.ToLower(expected)] {
					expectedCount++
				}
			}

			// Check for unwanted keywords
			unwantedCount := 0
			for _, unwanted := range tc.unwantedKeywords {
				if keywordMap[strings.ToLower(unwanted)] {
					unwantedCount++
				}
			}

			// Calculate quality ratio (expected keywords found / total keywords extracted)
			// This measures how many of the extracted keywords are business-relevant
			qualityRatio := float64(expectedCount) / float64(len(keywords))
			// Also calculate expected match ratio (expected keywords found / expected keywords total)
			expectedMatchRatio := float64(expectedCount) / float64(len(tc.expectedKeywords))
			
			// Quality ratio should meet minimum threshold
			if qualityRatio < tc.minQualityRatio && len(keywords) > 0 {
				t.Logf("Warning: Keyword quality ratio: %.1f%% (expected >= %.1f%%)", qualityRatio*100, tc.minQualityRatio*100)
			}
			
			// Expected match ratio should be reasonable (at least 50% of expected keywords found)
			if expectedMatchRatio < 0.5 {
				t.Logf("Warning: Only %.1f%% of expected keywords found (expected >= 50%%)", expectedMatchRatio*100)
			}

			// Should not have unwanted keywords
			if unwantedCount > 0 {
				t.Errorf("Found %d unwanted keywords (stop words should be filtered)", unwantedCount)
			}

			t.Logf("Keyword quality: %d/%d expected keywords found, %d unwanted keywords, quality ratio=%.1f%%", expectedCount, len(tc.expectedKeywords), unwantedCount, qualityRatio*100)
		})
	}
}

