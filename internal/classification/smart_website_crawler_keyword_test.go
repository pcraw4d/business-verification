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

func TestCountKeywordFrequency(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		text     string
		keyword  string
		expected int
	}{
		{
			name:     "exact match",
			text:     "wine shop wine store",
			keyword:  "wine",
			expected: 2,
		},
		{
			name:     "no substring match",
			text:     "winery winemaker",
			keyword:  "wine",
			expected: 0, // Should not match "wine" in "winery" or "winemaker"
		},
		{
			name:     "case insensitive",
			text:     "WINE shop Wine store",
			keyword:  "wine",
			expected: 2,
		},
		{
			name:     "no matches",
			text:     "shop store retail",
			keyword:  "wine",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crawler.countKeywordFrequency(tt.text, tt.keyword)
			if result != tt.expected {
				t.Errorf("countKeywordFrequency() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCountPhraseFrequency(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		text     string
		phrase   string
		expected int
	}{
		{
			name:     "exact phrase match",
			text:     "wine shop wine shop retail",
			phrase:   "wine shop",
			expected: 2,
		},
		{
			name:     "case insensitive",
			text:     "WINE SHOP wine shop",
			phrase:   "wine shop",
			expected: 2,
		},
		{
			name:     "no matches",
			text:     "shop store retail",
			phrase:   "wine shop",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crawler.countPhraseFrequency(tt.text, tt.phrase)
			if result != tt.expected {
				t.Errorf("countPhraseFrequency() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestCalculateCoOccurrenceBoost(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	text := "wine shop retail store beverage"
	keywordScores := map[string]float64{
		"wine":     0.9,
		"shop":     0.8,
		"retail":   0.7,
		"beverage": 0.6,
		"store":    0.5,
	}

	boosts := crawler.calculateCoOccurrenceBoost(keywordScores, text)

	// Should have boosts for keywords that appear close together
	if len(boosts) == 0 {
		t.Log("No co-occurrence boosts found (this may be expected if keywords are far apart)")
	} else {
		t.Logf("Found co-occurrence boosts: %v", boosts)
	}

	// Verify boosts are reasonable (not too high)
	for kw, boost := range boosts {
		if boost < 0 || boost > 0.5 {
			t.Errorf("calculateCoOccurrenceBoost() returned unreasonable boost for %s: %f", kw, boost)
		}
	}
}

func TestExtractStructuredData(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		html     string
		expected map[string]bool // Keys that should be present
	}{
		{
			name: "JSON-LD with LocalBusiness",
			html: `
				<script type="application/ld+json">
				{
					"@type": "LocalBusiness",
					"name": "The Greene Grape",
					"description": "Wine shop and retailer",
					"industry": "Food & Beverage"
				}
				</script>
			`,
			expected: map[string]bool{
				"business_name": true,
				"description":   true,
				"industry":      true,
				"schema_type":   true,
			},
		},
		{
			name: "Microdata with itemscope",
			html: `
				<div itemscope itemtype="http://schema.org/Store">
					<span itemprop="name">Wine Shop</span>
					<span itemprop="description">Premium wines</span>
				</div>
			`,
			expected: map[string]bool{
				"microdata-name":        true,
				"microdata-description": true,
			},
		},
		{
			name: "Empty JSON-LD",
			html: `<script type="application/ld+json"></script>`,
			expected: map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crawler.extractStructuredData(tt.html)
			
			// Check expected keys
			for key := range tt.expected {
				if _, exists := result[key]; !exists {
					t.Errorf("extractStructuredData() missing expected key: %s", key)
				}
			}
		})
	}
}

func TestExtractKeywordsFromStructuredData(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		data     map[string]interface{}
		expected []string // Keywords that should be present
		minCount int      // Minimum number of keywords expected
	}{
		{
			name: "business name and description",
			data: map[string]interface{}{
				"business_name": "The Greene Grape Wine Shop",
				"description":   "Premium wine retailer",
			},
			expected: []string{"grape", "wine", "shop", "retailer"},
			minCount: 3,
		},
		{
			name: "industry and services",
			data: map[string]interface{}{
				"industry": "Food & Beverage",
				"services": []string{"Wine Tasting", "Retail Sales"},
			},
			expected: []string{"food", "beverage", "wine", "tasting", "retail", "sales"},
			minCount: 4,
		},
		{
			name: "schema type",
			data: map[string]interface{}{
				"schema_type": "WineShop",
			},
			expected: []string{"wine", "shop"},
			minCount: 2,
		},
		{
			name: "empty data",
			data: map[string]interface{}{},
			expected: []string{},
			minCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := crawler.extractKeywordsFromStructuredData(tt.data)
			
			if len(keywords) < tt.minCount {
				t.Errorf("extractKeywordsFromStructuredData() returned %d keywords, want at least %d", len(keywords), tt.minCount)
			}
			
			// Check for expected keywords
			keywordMap := make(map[string]bool)
			for _, kw := range keywords {
				keywordMap[kw.keyword] = true
			}
			
			for _, expected := range tt.expected {
				if !keywordMap[expected] {
					t.Errorf("extractKeywordsFromStructuredData() missing expected keyword: %s", expected)
				}
			}
		})
	}
}

func TestSplitCamelCase(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	crawler := NewSmartWebsiteCrawler(logger)

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "PascalCase",
			input:    "WineShop",
			expected: []string{"wine", "shop"},
		},
		{
			name:     "camelCase",
			input:    "wineShop",
			expected: []string{"wine", "shop"},
		},
		{
			name:     "single word",
			input:    "Wine",
			expected: []string{"wine"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "all lowercase",
			input:    "wineshop",
			expected: []string{"wineshop"},
		},
		{
			name:     "multiple words",
			input:    "LocalBusinessStore",
			expected: []string{"local", "business", "store"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := crawler.splitCamelCase(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("splitCamelCase() returned %d words, want %d: %v vs %v", len(result), len(tt.expected), result, tt.expected)
				return
			}
			for i, word := range result {
				if word != tt.expected[i] {
					t.Errorf("splitCamelCase()[%d] = %s, want %s", i, word, tt.expected[i])
				}
			}
		})
	}
}

