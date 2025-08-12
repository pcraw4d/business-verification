package webanalysis

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

func TestNewIntelligentPageDiscovery(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	if ipd == nil {
		t.Fatal("Expected non-nil IntelligentPageDiscovery instance")
	}

	if ipd.scraper != scraper {
		t.Error("Expected scraper to be set correctly")
	}

	if len(ipd.pagePatterns) == 0 {
		t.Error("Expected page patterns to be initialized")
	}

	if len(ipd.keywordPatterns) == 0 {
		t.Error("Expected keyword patterns to be initialized")
	}

	if len(ipd.relevanceWeights) == 0 {
		t.Error("Expected relevance weights to be initialized")
	}
}

func TestDeterminePageType(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	tests := []struct {
		name     string
		url      string
		content  *ScrapedContent
		expected PageType
	}{
		{
			name: "About page by URL",
			url:  "https://example.com/about",
			content: &ScrapedContent{
				Title: "Some Title",
				Text:  "Some content",
			},
			expected: PageTypeAbout,
		},
		{
			name: "Mission page by URL",
			url:  "https://example.com/mission",
			content: &ScrapedContent{
				Title: "Some Title",
				Text:  "Some content",
			},
			expected: PageTypeMission,
		},
		{
			name: "Products page by URL",
			url:  "https://example.com/products",
			content: &ScrapedContent{
				Title: "Some Title",
				Text:  "Some content",
			},
			expected: PageTypeProducts,
		},
		{
			name: "About page by content",
			url:  "https://example.com/some-page",
			content: &ScrapedContent{
				Title: "About Us",
				Text:  "This is about our story and history",
			},
			expected: PageTypeAbout,
		},
		{
			name: "Unknown page",
			url:  "https://example.com/random",
			content: &ScrapedContent{
				Title: "Random Page",
				Text:  "Random content",
			},
			expected: PageTypeUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ipd.determinePageType(tt.url, tt.content)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestCalculateRelevanceScore(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	tests := []struct {
		name     string
		job      *PageDiscoveryJob
		content  *ScrapedContent
		pageType PageType
		expected float64
	}{
		{
			name: "High relevance about page",
			job: &PageDiscoveryJob{
				URL:      "https://example.com/about",
				Business: "Test Company",
				Depth:    0,
			},
			content: &ScrapedContent{
				Title: "About Test Company",
				Text:  "Test Company is a leading business in the industry. We provide excellent services and products.",
			},
			pageType: PageTypeAbout,
			expected: 0.8, // Should be high due to page type match and business name match
		},
		{
			name: "Low relevance random page",
			job: &PageDiscoveryJob{
				URL:      "https://example.com/random",
				Business: "Test Company",
				Depth:    2,
			},
			content: &ScrapedContent{
				Title: "Random Page",
				Text:  "Random content with no business information",
			},
			pageType: PageTypeUnknown,
			expected: 0.0, // Should be low due to no matches and depth penalty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ipd.calculateRelevanceScore(tt.job, tt.content, tt.pageType)
			if result < tt.expected-0.1 || result > tt.expected+0.1 {
				t.Errorf("Expected relevance score around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculateKeywordDensity(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	tests := []struct {
		name     string
		content  *ScrapedContent
		business string
		expected float64
	}{
		{
			name: "High keyword density",
			content: &ScrapedContent{
				Title: "Test Company Services",
				Text:  "Test Company provides excellent business services and products. Our company serves many clients in the industry.",
			},
			business: "Test Company",
			expected: 0.7, // Should be high due to multiple business keywords
		},
		{
			name: "Low keyword density",
			content: &ScrapedContent{
				Title: "Random Page",
				Text:  "This is just random content with no business information.",
			},
			business: "Test Company",
			expected: 0.09, // Should be low due to no business keywords
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ipd.calculateKeywordDensity(tt.content, tt.business)
			if result < tt.expected-0.05 || result > tt.expected+0.05 {
				t.Errorf("Expected keyword density around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculateBusinessNameMatch(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	tests := []struct {
		name     string
		content  *ScrapedContent
		business string
		expected float64
	}{
		{
			name: "Exact match",
			content: &ScrapedContent{
				Title: "About Test Company",
				Text:  "Test Company is a leading business.",
			},
			business: "Test Company",
			expected: 1.0,
		},
		{
			name: "Partial match",
			content: &ScrapedContent{
				Title: "About Our Company",
				Text:  "Test is a leading Company in the industry.",
			},
			business: "Test Company",
			expected: 1.0, // Both words match
		},
		{
			name: "No match",
			content: &ScrapedContent{
				Title: "Random Page",
				Text:  "No business information here.",
			},
			business: "Test Company",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ipd.calculateBusinessNameMatch(tt.content, tt.business)
			if result != tt.expected {
				t.Errorf("Expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculateContentQuality(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	tests := []struct {
		name     string
		content  *ScrapedContent
		expected float64
	}{
		{
			name: "High quality content",
			content: &ScrapedContent{
				Title: "About Our Company",
				Text:  "This is a comprehensive page about our company with substantial content that provides valuable information.",
				HTML:  "<h1>About Us</h1><p>This is a comprehensive page about our company.</p>",
			},
			expected: 0.9, // Should be high due to good title, substantial text, and HTML structure
		},
		{
			name: "Low quality content",
			content: &ScrapedContent{
				Title: "Page",
				Text:  "Short content",
				HTML:  "<div>menu menu menu menu menu</div>",
			},
			expected: 0.0, // Should be low due to short content and navigation-heavy HTML
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ipd.calculateContentQuality(tt.content)
			if result < tt.expected-0.2 || result > tt.expected+0.2 {
				t.Errorf("Expected content quality around %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestCalculatePriority(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	tests := []struct {
		name     string
		result   *PageDiscoveryResult
		expected int
	}{
		{
			name: "High priority about page",
			result: &PageDiscoveryResult{
				PageType:       PageTypeAbout,
				RelevanceScore: 0.9,
				Priority:       100,
				Depth:          0,
			},
			expected: 180, // 100 + 50 (about page) + 30 (high relevance)
		},
		{
			name: "Medium priority services page",
			result: &PageDiscoveryResult{
				PageType:       PageTypeServices,
				RelevanceScore: 0.7,
				Priority:       100,
				Depth:          1,
			},
			expected: 155, // 100 + 40 (services page) + 20 (medium relevance) - 5 (depth penalty)
		},
		{
			name: "Low priority unknown page",
			result: &PageDiscoveryResult{
				PageType:       PageTypeUnknown,
				RelevanceScore: 0.3,
				Priority:       100,
				Depth:          2,
			},
			expected: 90, // 100 + 10 (low relevance) - 20 (depth penalty)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ipd.calculatePriority(tt.result)
			if result != tt.expected {
				t.Errorf("Expected priority %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestShouldExcludeURL(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "PDF file",
			url:      "https://example.com/document.pdf",
			expected: true,
		},
		{
			name:     "Image file",
			url:      "https://example.com/image.jpg",
			expected: true,
		},
		{
			name:     "Admin page",
			url:      "https://example.com/admin/",
			expected: true,
		},
		{
			name:     "Login page",
			url:      "https://example.com/login/",
			expected: true,
		},
		{
			name:     "Valid about page",
			url:      "https://example.com/about",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ipd.shouldExcludeURL(tt.url)
			if result != tt.expected {
				t.Errorf("Expected %t for URL %s, got %t", tt.expected, tt.url, result)
			}
		})
	}
}

func TestShouldIncludeURL(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "About page",
			url:      "https://example.com/about",
			expected: true,
		},
		{
			name:     "Mission page",
			url:      "https://example.com/mission",
			expected: true,
		},
		{
			name:     "Products page",
			url:      "https://example.com/products",
			expected: true,
		},
		{
			name:     "Random page",
			url:      "https://example.com/random",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ipd.shouldIncludeURL(tt.url)
			if result != tt.expected {
				t.Errorf("Expected %t for URL %s, got %t", tt.expected, tt.url, result)
			}
		})
	}
}

func TestExtractInternalLinks(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	htmlContent := `
		<html>
			<body>
				<a href="/about">About Us</a>
				<a href="/products">Products</a>
				<a href="https://external.com">External Link</a>
				<a href="/contact">Contact</a>
			</body>
		</html>
	`

	baseURL := "https://example.com"
	links := ipd.extractInternalLinks(baseURL, htmlContent)

	expectedLinks := []string{
		"https://example.com/about",
		"https://example.com/products",
		"https://example.com/contact",
	}

	if len(links) != len(expectedLinks) {
		t.Errorf("Expected %d links, got %d", len(expectedLinks), len(links))
	}

	for _, expectedLink := range expectedLinks {
		found := false
		for _, link := range links {
			if link == expectedLink {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected link %s not found", expectedLink)
		}
	}
}

func TestGetDiscoveryStats(t *testing.T) {
	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	// Add some mock results
	ipd.mu.Lock()
	ipd.results["https://example.com/about"] = &PageDiscoveryResult{
		URL:            "https://example.com/about",
		RelevanceScore: 0.9,
		PageType:       PageTypeAbout,
	}
	ipd.results["https://example.com/products"] = &PageDiscoveryResult{
		URL:            "https://example.com/products",
		RelevanceScore: 0.7,
		PageType:       PageTypeProducts,
	}
	ipd.mu.Unlock()

	stats := ipd.GetDiscoveryStats()

	if stats["total_pages_discovered"] != 2 {
		t.Errorf("Expected 2 pages discovered, got %v", stats["total_pages_discovered"])
	}

	pageTypes := stats["page_types"].(map[PageType]int)
	if pageTypes[PageTypeAbout] != 1 {
		t.Errorf("Expected 1 about page, got %d", pageTypes[PageTypeAbout])
	}
	if pageTypes[PageTypeProducts] != 1 {
		t.Errorf("Expected 1 products page, got %d", pageTypes[PageTypeProducts])
	}

	avgRelevance := stats["average_relevance"].(float64)
	expectedAvg := (0.9 + 0.7) / 2.0
	if avgRelevance != expectedAvg {
		t.Errorf("Expected average relevance %f, got %f", expectedAvg, avgRelevance)
	}
}

func TestDiscoverPagesIntegration(t *testing.T) {
	// This is a mock test that would require a real scraper
	// In a real implementation, you would use a mock scraper
	t.Skip("Integration test requires mock scraper implementation")

	scraper := &WebScraper{}
	ipd := NewIntelligentPageDiscovery(scraper)

	ctx := context.Background()
	results, err := ipd.DiscoverPages(ctx, "https://example.com", "Test Company")

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(results) == 0 {
		t.Error("Expected at least one discovered page")
	}

	// Check that results are sorted by priority
	for i := 1; i < len(results); i++ {
		if results[i-1].Priority < results[i].Priority {
			t.Errorf("Results not sorted by priority: %d < %d", results[i-1].Priority, results[i].Priority)
		}
	}
}

func TestPageDiscoveryResultJSON(t *testing.T) {
	result := &PageDiscoveryResult{
		URL:               "https://example.com/about",
		RelevanceScore:    0.85,
		PageType:          PageTypeAbout,
		ContentIndicators: []string{"company_information", "mission_statement"},
		BusinessKeywords:  []string{"test", "company"},
		Priority:          150,
		Depth:             0,
		DiscoveredAt:      time.Now(),
		Metadata: map[string]string{
			"domain": "example.com",
		},
	}

	// Test that the struct can be marshaled to JSON
	_, err := json.Marshal(result)
	if err != nil {
		t.Errorf("Failed to marshal PageDiscoveryResult to JSON: %v", err)
	}
}
