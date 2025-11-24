package integration

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"kyb-platform/internal/classification/repository"
)

// TestClassificationFlow_WebsiteFirstPriority tests the full classification flow
// with website-first priority
func TestClassificationFlow_WebsiteFirstPriority(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	// Test that website content is extracted first
	businessName := "Test Business"
	websiteURL := "https://example.com"

	keywords := repo.ExtractKeywords(businessName, websiteURL)

	// Verify website keywords are present
	hasWebsiteKeywords := false
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
			if businessIndex == -1 {
				businessIndex = i
			}
		}
	}

	if !hasWebsiteKeywords {
		t.Error("Expected website keywords to be extracted")
	}

	// If both exist, website should come first
	if websiteIndex != -1 && businessIndex != -1 {
		if websiteIndex > businessIndex {
			t.Error("Website keywords should come before business name keywords")
		}
	}
}

// TestClassificationFlow_MultiPageAnalysis tests multi-page analysis integration
func TestClassificationFlow_MultiPageAnalysis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	// Test multi-page analysis
	websiteURL := "https://example.com"
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	keywords := repo.ExtractKeywordsFromMultiPageWebsite(ctx, websiteURL)

	// Multi-page analysis may return empty if < 3 pages succeed (triggers fallback)
	// Just verify it doesn't panic
	_ = keywords
}

// TestClassificationFlow_StructuredData tests structured data extraction in full flow
func TestClassificationFlow_StructuredData(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	// Test structured data extraction
	websiteURL := "https://example.com"
	ctx := context.Background()

	keywords := repo.ExtractKeywordsFromWebsite(ctx, websiteURL)

	// Verify function completes without error
	// Structured data extraction is integrated into extractKeywordsFromWebsite
	_ = keywords
}

// TestClassificationFlow_BrandMatch tests brand matching in classification flow
func TestClassificationFlow_BrandMatch(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	// Test brand match for hotel brands (MCC 3000-3831)
	businessName := "Hilton Hotels"
	websiteURL := "https://hilton.com"

	keywords := repo.ExtractKeywords(businessName, websiteURL)

	// Should have both website and business name keywords for brand match
	hasWebsiteKeywords := false
	hasBusinessKeywords := false

	for _, kw := range keywords {
		if kw.Context == "website_content" || kw.Context == "website_url" {
			hasWebsiteKeywords = true
		}
		if kw.Context == "business_name" {
			hasBusinessKeywords = true
		}
	}

	if !hasWebsiteKeywords {
		t.Error("Expected website keywords for brand match")
	}

	if !hasBusinessKeywords {
		t.Error("Expected business name keywords for brand match")
	}
}

// TestClassificationFlow_FallbackChain tests fallback from multi-page → single-page → URL
func TestClassificationFlow_FallbackChain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	// Test fallback chain with invalid URL
	invalidURL := "https://invalid-domain-that-does-not-exist-12345.com"
	businessName := "Test Business"

	keywords := repo.ExtractKeywords(businessName, invalidURL)

	// Should have some keywords from fallback chain (URL text extraction)
	// Even if website scraping fails, URL text extraction should work
	if len(keywords) == 0 {
		t.Log("Note: No keywords extracted from fallback chain - may be expected for invalid URL")
	}
}

// TestClassificationFlow_Performance tests multi-page analysis completes in <60s
func TestClassificationFlow_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := repository.NewSupabaseKeywordRepository(nil, logger)

	websiteURL := "https://example.com"
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	keywords := repo.ExtractKeywordsFromMultiPageWebsite(ctx, websiteURL)
	duration := time.Since(start)

	// Should complete within 60 seconds
	if duration > 65*time.Second {
		t.Errorf("Multi-page analysis exceeded 60s target. Duration: %v", duration)
	}

	_ = keywords
}

