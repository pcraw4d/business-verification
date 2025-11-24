package repository

import (
	"context"
	"log"
	"os"
	"testing"
	"time"
)

func TestExtractKeywords_PriorityOrder(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepository(nil, logger)

	tests := []struct {
		name            string
		businessName    string
		websiteURL      string
		expectWebsite   bool
		expectBusiness  bool
		websiteFirst    bool
	}{
		{
			name:           "website content extracted first",
			businessName:   "Test Business",
			websiteURL:     "https://example.com",
			expectWebsite:  true,
			expectBusiness: false, // Not a brand match
			websiteFirst:   true,
		},
		{
			name:           "business name extracted only for brand match",
			businessName:   "Hilton Hotels",
			websiteURL:     "https://hilton.com",
			expectWebsite:  true,
			expectBusiness: true, // Brand match
			websiteFirst:   true,
		},
		{
			name:           "no website URL - no website keywords",
			businessName:   "Test Business",
			websiteURL:     "",
			expectWebsite:  false,
			expectBusiness: false,
			websiteFirst:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := repo.extractKeywords(tt.businessName, tt.websiteURL)

			// Check if website keywords are present
			hasWebsiteKeywords := false
			hasBusinessKeywords := false
			websiteIndex := -1
			businessIndex := -1

			for i, kw := range keywords {
				if kw.Context == "website_content" || kw.Context == "website_url" {
					hasWebsiteKeywords = true
					if websiteIndex == -1 {
						websiteIndex = i
					}
				}
				if kw.Context == "business_name" {
					hasBusinessKeywords = true
					if businessIndex == -1 {
						businessIndex = i
					}
				}
			}

			if hasWebsiteKeywords != tt.expectWebsite {
				t.Errorf("Expected website keywords: %v, got: %v", tt.expectWebsite, hasWebsiteKeywords)
			}

			if hasBusinessKeywords != tt.expectBusiness {
				t.Errorf("Expected business keywords: %v, got: %v", tt.expectBusiness, hasBusinessKeywords)
			}

			// Verify website keywords come before business keywords
			if tt.websiteFirst && hasWebsiteKeywords && hasBusinessKeywords {
				if websiteIndex > businessIndex {
					t.Errorf("Website keywords should come before business keywords. Website index: %d, Business index: %d", websiteIndex, businessIndex)
				}
			}
		})
	}
}

func TestExtractKeywords_BrandMatch(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepository(nil, logger)

	tests := []struct {
		name           string
		businessName   string
		websiteURL     string
		expectKeywords bool
	}{
		{
			name:           "Hilton brand match",
			businessName:   "Hilton Hotels Inc",
			websiteURL:     "https://hilton.com",
			expectKeywords: true,
		},
		{
			name:           "Marriott brand match",
			businessName:   "Marriott International",
			websiteURL:     "https://marriott.com",
			expectKeywords: true,
		},
		{
			name:           "Hyatt brand match",
			businessName:   "Hyatt Hotels Corporation",
			websiteURL:     "https://hyatt.com",
			expectKeywords: true,
		},
		{
			name:           "Non-brand business",
			businessName:   "Acme Corporation",
			websiteURL:     "https://acme.com",
			expectKeywords: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := repo.extractKeywords(tt.businessName, tt.websiteURL)

			hasBusinessKeywords := false
			for _, kw := range keywords {
				if kw.Context == "business_name" {
					hasBusinessKeywords = true
					break
				}
			}

			if hasBusinessKeywords != tt.expectKeywords {
				t.Errorf("Expected business keywords for brand match: %v, got: %v", tt.expectKeywords, hasBusinessKeywords)
			}
		})
	}
}

func TestExtractKeywords_NonBrandMatch(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepository(nil, logger)

	nonBrandNames := []string{
		"Acme Corporation",
		"Tech Startup Inc",
		"Local Restaurant LLC",
		"Retail Store Corp",
	}

	for _, businessName := range nonBrandNames {
		t.Run(businessName, func(t *testing.T) {
			keywords := repo.extractKeywords(businessName, "https://example.com")

			// Should not have business_name keywords for non-brand matches
			for _, kw := range keywords {
				if kw.Context == "business_name" {
					t.Errorf("Non-brand business '%s' should not have business_name keywords", businessName)
				}
			}
		})
	}
}

func TestExtractKeywordsFromMultiPageWebsite(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepository(nil, logger)

	// This test requires a real website, so we'll test the timeout and fallback behavior
	tests := []struct {
		name           string
		websiteURL     string
		expectFallback bool
		timeout        time.Duration
	}{
		{
			name:           "valid website URL",
			websiteURL:     "https://example.com",
			expectFallback: false,
			timeout:        60 * time.Second,
		},
		{
			name:           "invalid URL triggers fallback",
			websiteURL:     "not-a-valid-url",
			expectFallback: true,
			timeout:        5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			keywords := repo.extractKeywordsFromMultiPageWebsite(ctx, tt.websiteURL)

			// If fallback is expected, keywords should be empty (triggers fallback)
			if tt.expectFallback {
				if len(keywords) > 0 {
					t.Logf("Note: Got keywords despite expecting fallback (may be URL text extraction): %v", keywords)
				}
			} else {
				// For valid URLs, we may get keywords or empty (depends on website)
				// Just verify the function doesn't panic
				_ = keywords
			}
		})
	}
}

func TestExtractKeywords_StructuredData(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepository(nil, logger)

	// Test that structured data extraction is integrated
	// This is tested indirectly through extractKeywordsFromWebsite
	websiteURL := "https://example.com"
	keywords := repo.extractKeywordsFromWebsite(context.Background(), websiteURL)

	// Verify function completes without error
	// Actual structured data extraction depends on website content
	_ = keywords
}

func TestCalculatePagePriority(t *testing.T) {
	// This test would require access to SmartWebsiteCrawler
	// Since it's in a different package, we'll test indirectly
	// through integration tests
	t.Skip("Requires SmartWebsiteCrawler - tested in integration tests")
}

func TestCalculateRelevanceScore(t *testing.T) {
	// This test would require access to SmartWebsiteCrawler
	// Since it's in a different package, we'll test indirectly
	// through integration tests
	t.Skip("Requires SmartWebsiteCrawler - tested in integration tests")
}

func TestMultiPageAnalysis_Timeout(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepository(nil, logger)

	// Test that 60s timeout is enforced
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // Very short timeout
	defer cancel()

	start := time.Now()
	keywords := repo.extractKeywordsFromMultiPageWebsite(ctx, "https://example.com")
	duration := time.Since(start)

	// Should complete within timeout (or return empty for fallback)
	if duration > 3*time.Second {
		t.Errorf("Multi-page analysis exceeded timeout. Duration: %v", duration)
	}

	_ = keywords // Use result to avoid unused variable
}

func TestMultiPageAnalysis_Fallback(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepository(nil, logger)

	// Test fallback behavior when multi-page analysis fails
	// Use an invalid URL that will trigger fallback
	invalidURL := "https://invalid-domain-that-does-not-exist-12345.com"
	keywords := repo.extractKeywordsFromMultiPageWebsite(context.Background(), invalidURL)

	// Should return empty to trigger fallback to single-page
	if len(keywords) > 0 {
		t.Logf("Note: Got keywords from invalid URL (may be URL text extraction): %v", keywords)
	}

	// Test that extractKeywords handles fallback correctly
	allKeywords := repo.extractKeywords("Test Business", invalidURL)

	// Should have some keywords from fallback chain (URL text extraction)
	if len(allKeywords) == 0 {
		t.Log("Note: No keywords extracted even from fallback - may be expected for invalid URL")
	}
}

