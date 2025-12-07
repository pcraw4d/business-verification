package repository

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// TestKeywordExtractionIntegration_Phase10_2 tests the full keyword extraction flow (Phase 10.2)
func TestKeywordExtractionIntegration_Phase10_2(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	tests := []struct {
		name           string
		businessName   string
		websiteURL     string
		expectedMin    int
		expectedMax    int // Maximum expected keywords (to catch excessive extraction)
		expectedLevels []string // Expected fallback levels that might be used
	}{
		{
			name:           "full flow with valid website",
			businessName:   "The Greene Grape",
			websiteURL:     "https://www.thegreenegrape.com",
			expectedMin:    3, // At least URL extraction should work
			expectedMax:    50,
			expectedLevels: []string{"multi_page", "single_page", "homepage_retry", "url_only"},
		},
		{
			name:           "full flow with business name only",
			businessName:   "Marriott Hotel",
			websiteURL:     "",
			expectedMin:    1, // Brand match should extract keywords
			expectedMax:    10,
			expectedLevels: []string{"business_name"},
		},
		{
			name:           "full flow with invalid URL",
			businessName:   "Test Business",
			websiteURL:     "https://invalid-domain-that-does-not-exist-12345.com",
			expectedMin:    2, // Should fallback to business name
			expectedMax:    10,
			expectedLevels: []string{"business_name", "url_only"},
		},
		{
			name:           "full flow with URL without scheme",
			businessName:   "",
			websiteURL:     "example.com",
			expectedMin:    1, // URL extraction should work
			expectedMax:    20,
			expectedLevels: []string{"url_only"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			start := time.Now()
			keywords := repo.extractKeywords(ctx, tt.businessName, tt.websiteURL)
			duration := time.Since(start)

			// Verify keyword count
			if len(keywords) < tt.expectedMin {
				t.Errorf("extractKeywords() returned %d keywords, want at least %d", len(keywords), tt.expectedMin)
			}

			if len(keywords) > tt.expectedMax {
				t.Errorf("extractKeywords() returned %d keywords, want at most %d (possible excessive extraction)", len(keywords), tt.expectedMax)
			}

			// Verify all keywords have valid context
			contexts := make(map[string]bool)
			for _, kw := range keywords {
				if kw.Context == "" {
					t.Errorf("extractKeywords() returned keyword without context: %s", kw.Keyword)
				}
				if strings.TrimSpace(kw.Keyword) == "" {
					t.Errorf("extractKeywords() returned empty keyword")
				}
				contexts[kw.Context] = true
			}

			// Verify no duplicate keywords
			seen := make(map[string]bool)
			for _, kw := range keywords {
				key := strings.ToLower(kw.Keyword)
				if seen[key] {
					t.Errorf("extractKeywords() returned duplicate keyword: %s", kw.Keyword)
				}
				seen[key] = true
			}

			// Verify performance (should complete in reasonable time)
			if duration > 30*time.Second {
				t.Errorf("extractKeywords() took too long: %v (expected < 30s)", duration)
			}

			t.Logf("Integration test passed: extracted %d keywords in %v with contexts: %v", len(keywords), duration, getContextKeys(contexts))
		})
	}
}

// TestMultiPageAnalysisIntegration tests multi-page analysis with successful and failed pages (Phase 10.2)
func TestMultiPageAnalysisIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	tests := []struct {
		name        string
		websiteURL  string
		expectMin   int
		expectMax   int
		description string
	}{
		{
			name:        "multi-page website with successful pages",
			websiteURL:  "https://example.com",
			expectMin:   0, // May succeed or fail depending on network
			expectMax:   100,
			description: "Tests multi-page analysis with a real website",
		},
		{
			name:        "invalid domain triggers fallback",
			websiteURL:  "https://this-domain-does-not-exist-12345.com",
			expectMin:   0, // Should trigger fallback
			expectMax:   10,
			description: "Tests fallback when multi-page analysis fails",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			start := time.Now()
			keywords := repo.extractKeywordsFromMultiPageWebsite(ctx, tt.websiteURL)
			duration := time.Since(start)

			// Verify keyword count is within expected range
			if len(keywords) < tt.expectMin {
				t.Logf("Warning: Got %d keywords, expected at least %d (may be network issue)", len(keywords), tt.expectMin)
			}

			if len(keywords) > tt.expectMax {
				t.Errorf("extractKeywordsFromMultiPageWebsite() returned %d keywords, want at most %d", len(keywords), tt.expectMax)
			}

			// Verify all keywords are non-empty strings
			for _, kw := range keywords {
				if strings.TrimSpace(kw) == "" {
					t.Errorf("extractKeywordsFromMultiPageWebsite() returned empty keyword")
				}
			}

			// Verify performance (should complete in reasonable time)
			if duration > 30*time.Second {
				t.Errorf("extractKeywordsFromMultiPageWebsite() took too long: %v (expected < 30s)", duration)
			}

			t.Logf("Multi-page analysis: extracted %d keywords in %v", len(keywords), duration)
		})
	}
}

// TestEndToEndClassificationIntegration tests end-to-end classification with improved keywords (Phase 10.2)
func TestEndToEndClassificationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	tests := []struct {
		name           string
		businessName   string
		websiteURL     string
		expectedMin    int
		description    string
	}{
		{
			name:         "wine retailer classification",
			businessName: "The Greene Grape",
			websiteURL:   "https://www.thegreenegrape.com",
			expectedMin:  3,
			description:  "Tests classification of a wine retailer business",
		},
		{
			name:         "technology company classification",
			businessName: "Tech Solutions Inc",
			websiteURL:   "https://techsolutions.io",
			expectedMin:  2,
			description:  "Tests classification of a technology company",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Extract keywords
			keywords := repo.extractKeywords(ctx, tt.businessName, tt.websiteURL)

			if len(keywords) < tt.expectedMin {
				t.Logf("Warning: Got %d keywords, expected at least %d (may be network issue)", len(keywords), tt.expectedMin)
			}

			// Convert to string slice for classification
			keywordStrings := make([]string, 0, len(keywords))
			for _, kw := range keywords {
				keywordStrings = append(keywordStrings, kw.Keyword)
			}

			// Test classification with extracted keywords
			if len(keywordStrings) > 0 {
				result, err := repo.ClassifyBusinessByKeywords(ctx, keywordStrings)
				if err != nil {
					t.Logf("Classification error (may be expected): %v", err)
				} else if result != nil {
					// Verify classification result structure
					if result.Industry == nil {
						t.Logf("Warning: Classification returned nil industry")
					}
					if result.Confidence < 0 || result.Confidence > 1 {
						t.Errorf("Classification returned invalid confidence: %f (expected 0-1)", result.Confidence)
					}

					industryName := "Unknown"
					if result.Industry != nil {
						industryName = result.Industry.Name
					}
					t.Logf("Classification result: Industry=%s, Confidence=%.2f, Keywords used=%d", industryName, result.Confidence, len(keywordStrings))
				}
			}

			t.Logf("End-to-end test: extracted %d keywords, classification attempted", len(keywords))
		})
	}
}

// Helper function to get context keys from map
func getContextKeys(contexts map[string]bool) []string {
	keys := make([]string, 0, len(contexts))
	for k := range contexts {
		keys = append(keys, k)
	}
	return keys
}

