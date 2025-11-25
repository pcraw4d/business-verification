package classification

import (
	"log"
	"os"
	"strings"
	"testing"
)

func TestExtractPageKeywords(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		content  string
		pageType string
		wantMin  int // Minimum number of keywords expected
		wantKeys []string // Keywords that should be present
	}{
		{
			name: "wine retailer page",
			content: `
				<html>
					<head>
						<title>The Greene Grape - Wine Shop & Retail Store</title>
						<meta name="description" content="Premium wine retailer offering fine wines, spirits, and beverages">
					</head>
					<body>
						<h1>Welcome to The Greene Grape</h1>
						<p>We are a premier wine shop offering a curated selection of wines, spirits, and beverages. Visit our retail store for wine tastings and expert recommendations.</p>
						<h2>Our Wine Selection</h2>
						<p>Browse our extensive collection of wines from vineyards around the world.</p>
					</body>
				</html>
			`,
			pageType: "homepage",
			wantMin:  5,
			wantKeys: []string{"wine", "retail", "shop", "beverage", "store"},
		},
		{
			name: "technology company page",
			content: `
				<html>
					<head>
						<title>Tech Solutions - Software Development</title>
						<meta name="description" content="Leading software development company specializing in cloud solutions and digital transformation">
					</head>
					<body>
						<h1>Technology Solutions</h1>
						<p>We provide software development, cloud computing, and digital transformation services.</p>
					</body>
				</html>
			`,
			pageType: "services",
			wantMin:  5,
			wantKeys: []string{"technology", "software", "cloud", "digital"},
		},
		{
			name:     "empty content",
			content:  "<html><body></body></html>",
			pageType: "homepage",
			wantMin:  0,
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := crawler.extractPageKeywords(tt.content, tt.pageType)

			if len(keywords) < tt.wantMin {
				t.Errorf("extractPageKeywords() returned %d keywords, want at least %d", len(keywords), tt.wantMin)
			}

			// Check for expected keywords
			keywordMap := make(map[string]bool)
			for _, kw := range keywords {
				keywordMap[strings.ToLower(kw)] = true
			}

			for _, wantKey := range tt.wantKeys {
				if !keywordMap[strings.ToLower(wantKey)] {
					t.Errorf("extractPageKeywords() missing expected keyword: %s. Got: %v", wantKey, keywords)
				}
			}

			// Verify no empty keywords
			for _, kw := range keywords {
				if strings.TrimSpace(kw) == "" {
					t.Errorf("extractPageKeywords() returned empty keyword")
				}
			}
		})
	}
}

func TestExtractTextFromHTML(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		html     string
		wantText string
	}{
		{
			name:     "simple HTML",
			html:     "<html><body><p>Hello World</p></body></html>",
			wantText: "Hello World",
		},
		{
			name:     "with scripts and styles",
			html:     "<html><head><style>body { color: red; }</style><script>alert('test');</script></head><body><p>Content</p></body></html>",
			wantText: "Content",
		},
		{
			name:     "with HTML entities",
			html:     "<html><body><p>Hello&nbsp;World &amp; More</p></body></html>",
			wantText: "Hello World & More",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := crawler.extractTextFromHTML(tt.html)
			if !strings.Contains(text, tt.wantText) {
				t.Errorf("extractTextFromHTML() = %q, want to contain %q", text, tt.wantText)
			}
		})
	}
}

func TestExtractStructuredKeywords(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		content  string
		wantKeys []string
	}{
		{
			name: "with title and meta",
			content: `
				<html>
					<head>
						<title>Wine Shop - Premium Wines</title>
						<meta name="description" content="Best wine retailer with fine selection">
					</head>
					<body>
						<h1>Welcome</h1>
						<h2>Our Products</h2>
					</body>
				</html>
			`,
			wantKeys: []string{"wine", "shop", "premium", "wines"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := crawler.extractStructuredKeywords(tt.content)
			keywordMap := make(map[string]bool)
			for _, kw := range keywords {
				keywordMap[strings.ToLower(kw)] = true
			}

			for _, wantKey := range tt.wantKeys {
				if !keywordMap[strings.ToLower(wantKey)] {
					t.Errorf("extractStructuredKeywords() missing expected keyword: %s. Got: %v", wantKey, keywords)
				}
			}
		})
	}
}

func TestExtractIndustryIndicators(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		content  string
		wantKeys []string
	}{
		{
			name:    "wine industry",
			content: "We sell wine, grapes, and beverages at our winery and retail store.",
			wantKeys: []string{"food_beverage:wine", "food_beverage:grape", "food_beverage:beverage", "retail:retail"},
		},
		{
			name:    "technology industry",
			content: "Software development and cloud computing solutions.",
			wantKeys: []string{"technology:software", "technology:cloud"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indicators := crawler.extractIndustryIndicators(tt.content)
			indicatorMap := make(map[string]bool)
			for _, ind := range indicators {
				indicatorMap[strings.ToLower(ind)] = true
			}

			for _, wantKey := range tt.wantKeys {
				if !indicatorMap[strings.ToLower(wantKey)] {
					t.Errorf("extractIndustryIndicators() missing expected indicator: %s. Got: %v", wantKey, indicators)
				}
			}
		})
	}
}

func TestLimitToTopKeywords(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	scoredKeywords := []keywordScore{
		{keyword: "wine", score: 0.9},
		{keyword: "retail", score: 0.8},
		{keyword: "shop", score: 0.7},
		{keyword: "beverage", score: 0.6},
		{keyword: "store", score: 0.5},
	}

	limited := crawler.limitToTopKeywords(scoredKeywords, 3)
	if len(limited) != 3 {
		t.Errorf("limitToTopKeywords() returned %d keywords, want 3", len(limited))
	}

	if limited[0] != "wine" || limited[1] != "retail" || limited[2] != "shop" {
		t.Errorf("limitToTopKeywords() returned wrong keywords: %v", limited)
	}
}

