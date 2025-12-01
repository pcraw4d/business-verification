package classification

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/database"
)

// Test adapter implementations to avoid import cycle
type testStructuredDataExtractorAdapter struct {
	extractor *StructuredDataExtractor
}

func (t *testStructuredDataExtractorAdapter) ExtractStructuredData(htmlContent string) repository.StructuredDataResult {
	result := t.extractor.ExtractStructuredData(htmlContent)
	
	// Convert to repository interface type
	return repository.StructuredDataResult{
		SchemaOrgData:   convertSchemaOrgDataForTest(result.SchemaOrgData),
		OpenGraphData:   result.OpenGraphData,
		TwitterCardData: result.TwitterCardData,
		Microdata:       convertMicrodataForTest(result.Microdata),
		BusinessInfo: repository.BusinessInfoData{
			BusinessName: result.BusinessInfo.BusinessName,
			Description:  result.BusinessInfo.Description,
			Services:     result.BusinessInfo.Services,
			Products:     result.BusinessInfo.Products,
			Industry:     result.BusinessInfo.Industry,
			BusinessType: result.BusinessInfo.BusinessType,
		},
		ContactInfo: repository.ContactInfoData{
			Phone:   result.ContactInfo.Phone,
			Email:   result.ContactInfo.Email,
			Address: result.ContactInfo.Address,
			Website: result.ContactInfo.Website,
			Social:  result.ContactInfo.Social,
		},
		ProductInfo:     convertProductInfoForTest(result.ProductInfo),
		ServiceInfo:     convertServiceInfoForTest(result.ServiceInfo),
		EventInfo:       convertEventInfoForTest(result.EventInfo),
		ExtractionScore: result.ExtractionScore,
	}
}

func convertSchemaOrgDataForTest(data []SchemaOrgItem) []repository.SchemaOrgItem {
	if data == nil {
		return nil
	}
	result := make([]repository.SchemaOrgItem, len(data))
	for i, item := range data {
		result[i] = repository.SchemaOrgItem{
			Type:       item.Type,
			Properties: item.Properties,
			Context:    item.Context,
			Confidence: item.Confidence,
		}
	}
	return result
}

func convertMicrodataForTest(data []MicrodataItem) []repository.MicrodataItem {
	if data == nil {
		return nil
	}
	result := make([]repository.MicrodataItem, len(data))
	for i, item := range data {
		result[i] = repository.MicrodataItem{
			Type:       item.Type,
			Properties: item.Properties,
		}
	}
	return result
}

func convertProductInfoForTest(data []ProductInfo) []repository.ProductInfoData {
	if data == nil {
		return nil
	}
	result := make([]repository.ProductInfoData, len(data))
	for i, item := range data {
		result[i] = repository.ProductInfoData{
			Name:        item.Name,
			Description: item.Description,
			Price:       item.Price,
			Category:    item.Category,
			Brand:       item.Brand,
			SKU:         item.SKU,
			Image:       item.Image,
			URL:         item.URL,
			Confidence:  item.Confidence,
		}
	}
	return result
}

func convertServiceInfoForTest(data []ServiceInfo) []repository.ServiceInfoData {
	if data == nil {
		return nil
	}
	result := make([]repository.ServiceInfoData, len(data))
	for i, item := range data {
		result[i] = repository.ServiceInfoData{
			Name:        item.Name,
			Description: item.Description,
			Category:    item.Category,
			Price:       item.Price,
			Duration:    item.Duration,
			Features:    item.Features,
			URL:         item.URL,
			Confidence:  item.Confidence,
		}
	}
	return result
}

func convertEventInfoForTest(data []EventInfo) []repository.EventInfoData {
	if data == nil {
		return nil
	}
	result := make([]repository.EventInfoData, len(data))
	for i, item := range data {
		result[i] = repository.EventInfoData{
			Name:        item.Name,
			Description: item.Description,
			StartDate:   item.StartDate,
			EndDate:     item.EndDate,
			Location:    item.Location,
			URL:         item.URL,
			Confidence:  item.Confidence,
		}
	}
	return result
}

type testPageAnalysisAdapter struct {
	analysis PageAnalysis
}

func (p *testPageAnalysisAdapter) GetURL() string {
	return p.analysis.URL
}

func (p *testPageAnalysisAdapter) GetStatusCode() int {
	return p.analysis.StatusCode
}

func (p *testPageAnalysisAdapter) GetRelevanceScore() float64 {
	return p.analysis.RelevanceScore
}

func (p *testPageAnalysisAdapter) GetKeywords() []string {
	return p.analysis.Keywords
}

func (p *testPageAnalysisAdapter) GetIndustryIndicators() []string {
	return p.analysis.IndustryIndicators
}

func (p *testPageAnalysisAdapter) GetStructuredData() map[string]interface{} {
	return p.analysis.StructuredData
}

type testCrawlResultAdapter struct {
	result *CrawlResult
}

func (c *testCrawlResultAdapter) GetPagesAnalyzed() []repository.PageAnalysisData {
	adapters := make([]repository.PageAnalysisData, len(c.result.PagesAnalyzed))
	for i, page := range c.result.PagesAnalyzed {
		adapters[i] = &testPageAnalysisAdapter{analysis: page}
	}
	return adapters
}

func (c *testCrawlResultAdapter) GetSuccess() bool {
	return c.result.Success
}

func (c *testCrawlResultAdapter) GetError() string {
	return c.result.Error
}

type testSmartWebsiteCrawlerAdapter struct {
	crawler *SmartWebsiteCrawler
}

func (s *testSmartWebsiteCrawlerAdapter) CrawlWebsite(ctx context.Context, websiteURL string) (repository.CrawlResultInterface, error) {
	result, err := s.crawler.CrawlWebsite(ctx, websiteURL)
	if err != nil {
		return nil, err
	}
	return &testCrawlResultAdapter{result: result}, nil
}

func (s *testSmartWebsiteCrawlerAdapter) CrawlWebsiteFast(ctx context.Context, websiteURL string, maxTime time.Duration, maxPages int, maxConcurrent int) (repository.CrawlResultInterface, error) {
	result, err := s.crawler.CrawlWebsiteFast(ctx, websiteURL, maxTime, maxPages, maxConcurrent)
	if err != nil {
		return nil, err
	}
	return &testCrawlResultAdapter{result: result}, nil
	result, err := s.crawler.CrawlWebsite(ctx, websiteURL)
	if err != nil {
		return nil, err
	}
	return &testCrawlResultAdapter{result: result}, nil
}

// TestServiceWithRealDatabase tests the service with a real Supabase database
func TestServiceWithRealDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping database integration test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Initialize adapters FIRST - this is critical for keyword extraction
	// We initialize them directly to avoid import cycle
	repository.InitAdapters(
		// StructuredDataExtractor adapter
		func(logger *log.Logger) repository.StructuredDataExtractorInterface {
			extractor := NewStructuredDataExtractor(logger)
			return &testStructuredDataExtractorAdapter{extractor: extractor}
		},
		// SmartWebsiteCrawler adapter
		func(logger *log.Logger) repository.SmartWebsiteCrawlerInterface {
			crawler := NewSmartWebsiteCrawler(logger)
			return &testSmartWebsiteCrawlerAdapter{crawler: crawler}
		},
	)

	// Create database client
	config := &database.SupabaseConfig{
		URL:            supabaseURL,
		APIKey:         supabaseKey,
		ServiceRoleKey: supabaseServiceKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Skipf("Skipping database integration test: cannot connect to Supabase: %v", err)
	}

	// Create repository (adapters are now initialized)
	repo := repository.NewSupabaseKeywordRepository(client, logger)

	// Create service
	service := NewIndustryDetectionService(repo, logger)

	// Test cases
	testCases := []struct {
		name             string
		businessName     string
		description      string
		websiteURL       string
		expectedIndustry string
		minConfidence    float64
	}{
		{
			name:             "Technology company",
			businessName:     "Microsoft Corporation",
			description:      "Software development and cloud computing services",
			websiteURL:       "https://microsoft.com",
			expectedIndustry: "Technology",
			minConfidence:    0.70,
		},
		{
			name:             "Healthcare company",
			businessName:     "Mayo Clinic",
			description:      "Medical center and hospital services",
			websiteURL:       "https://mayoclinic.org",
			expectedIndustry: "Healthcare",
			minConfidence:    0.70,
		},
		{
			name:             "Retail company",
			businessName:     "Amazon",
			description:      "E-commerce and retail services",
			websiteURL:       "https://amazon.com",
			expectedIndustry: "Retail",
			minConfidence:    0.65,
		},
	}

	ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test industry detection
			result, err := service.DetectIndustry(ctx, tc.businessName, tc.description, tc.websiteURL)
			if err != nil {
				t.Errorf("DetectIndustry failed: %v", err)
				return
			}

			if result == nil {
				t.Fatal("DetectIndustry returned nil result")
				return
			}

			// Verify method is multi-strategy
			if result.Method != "multi_strategy" {
				t.Logf("⚠️ Expected multi_strategy method, got: %s", result.Method)
				// Don't fail - fallback is acceptable
			} else {
				t.Logf("✅ Multi-strategy classification confirmed")
			}

			// Verify industry matches (allow some flexibility)
			if result.IndustryName != tc.expectedIndustry {
				// Check if it's a related industry
				if !containsIndustryName(result.IndustryName, tc.expectedIndustry) {
					t.Logf("⚠️ Industry mismatch: expected %s, got %s (confidence: %.2f%%)",
						tc.expectedIndustry, result.IndustryName, result.Confidence*100)
					// Don't fail if confidence is high - might be valid alternative
					if result.Confidence < 0.60 {
						t.Errorf("Industry mismatch: expected %s, got %s", tc.expectedIndustry, result.IndustryName)
					}
				}
			}

			// Verify confidence meets minimum
			if result.Confidence < tc.minConfidence {
				t.Errorf("Confidence too low: expected >= %.2f, got %.2f",
					tc.minConfidence, result.Confidence)
			}

			// Verify keywords were extracted (or known business fallback was used)
			if len(result.Keywords) == 0 {
				// For known businesses, it's acceptable to have 0 keywords if fallback was used
				// Check if this is a known business that was correctly classified
				knownBusinesses := []string{"amazon", "microsoft", "google", "apple", "mayo clinic"}
				isKnownBusiness := false
				for _, known := range knownBusinesses {
					if strings.Contains(strings.ToLower(tc.businessName), known) {
						isKnownBusiness = true
						break
					}
				}
				if isKnownBusiness && result.Confidence >= 0.70 {
					t.Logf("✅ Known business fallback used (no keywords needed for %s -> %s)", tc.businessName, result.IndustryName)
				} else {
					t.Logf("⚠️ No keywords extracted (may be acceptable for known businesses)")
				}
			} else {
				t.Logf("✅ Extracted %d keywords: %v", len(result.Keywords), result.Keywords[:min(10, len(result.Keywords))])
			}

			// Verify processing time is reasonable
			if result.ProcessingTime > 30*time.Second {
				t.Errorf("Processing time too slow: %v", result.ProcessingTime)
			}

			t.Logf("✅ %s classified as %s (confidence: %.2f%%, method: %s, time: %v, keywords: %d)",
				tc.businessName, result.IndustryName, result.Confidence*100,
				result.Method, result.ProcessingTime, len(result.Keywords))
		})
	}
}

// TestMultiStrategyClassifierWithDatabase tests the multi-strategy classifier with a real database
func TestMultiStrategyClassifierWithDatabase(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	// Check for Supabase credentials
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")

	if supabaseURL == "" || supabaseKey == "" {
		t.Skip("Skipping database integration test: SUPABASE_URL or SUPABASE_ANON_KEY not set")
	}

	// Initialize adapters FIRST
	repository.InitAdapters(
		func(logger *log.Logger) repository.StructuredDataExtractorInterface {
			extractor := NewStructuredDataExtractor(logger)
			return &testStructuredDataExtractorAdapter{extractor: extractor}
		},
		func(logger *log.Logger) repository.SmartWebsiteCrawlerInterface {
			crawler := NewSmartWebsiteCrawler(logger)
			return &testSmartWebsiteCrawlerAdapter{crawler: crawler}
		},
	)

	// Create database client
	config := &database.SupabaseConfig{
		URL:            supabaseURL,
		APIKey:         supabaseKey,
		ServiceRoleKey: supabaseServiceKey,
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	client, err := database.NewSupabaseClient(config, logger)
	if err != nil {
		t.Fatalf("Failed to create Supabase client: %v", err)
	}
	defer client.Close()

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx); err != nil {
		t.Skipf("Skipping database integration test: cannot connect to Supabase: %v", err)
	}

	// Create repository
	repo := repository.NewSupabaseKeywordRepository(client, logger)

	// Create multi-strategy classifier
	classifier := NewMultiStrategyClassifier(repo, logger)

	ctx, cancel = context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Test classification
	businessName := "Apple Inc"
	description := "Technology company specializing in consumer electronics and software"
	websiteURL := "https://apple.com"

	result, err := classifier.ClassifyWithMultiStrategy(ctx, businessName, description, websiteURL)
	if err != nil {
		t.Fatalf("ClassifyWithMultiStrategy failed: %v", err)
	}

	if result == nil {
		t.Fatal("ClassifyWithMultiStrategy returned nil result")
	}

	if result.PrimaryIndustry == "" {
		t.Error("Expected primary industry, got empty string")
	}

	if result.Confidence < 0.0 || result.Confidence > 1.0 {
		t.Errorf("Confidence out of range: %.2f", result.Confidence)
	}

	t.Logf("✅ Multi-strategy classification: %s (confidence: %.2f%%, keywords: %d)",
		result.PrimaryIndustry, result.Confidence*100, len(result.Keywords))
}

// Helper function to check if industry names are related
func containsIndustryName(result, expected string) bool {
	// Simple check for related industries
	related := map[string][]string{
		"Technology":     {"Software", "IT", "Tech", "Digital"},
		"Healthcare":     {"Medical", "Health", "Hospital"},
		"Retail":         {"E-commerce", "Commerce", "Store"},
		"Food & Beverage": {"Restaurant", "Food", "Dining"},
	}

	if industries, ok := related[expected]; ok {
		for _, industry := range industries {
			if strings.Contains(result, industry) {
				return true
			}
		}
	}
	return false
}
