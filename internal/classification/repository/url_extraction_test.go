package repository

import (
	"log"
	"net/url"
	"os"
	"strings"
	"testing"

	"kyb-platform/internal/classification/word_segmentation"
)

func TestExtractKeywordsFromURLEnhanced(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := &SupabaseKeywordRepository{
		logger:    logger,
		segmenter: word_segmentation.NewSegmenter(),
	}

	tests := []struct {
		name     string
		url      string
		wantMin  int
		wantKeys []string
	}{
		{
			name:     "wine retailer domain",
			url:      "https://www.thegreenegrape.com",
			wantMin:  3,
			wantKeys: []string{"green", "grape", "wine", "retail"},
		},
		{
			name:     "technology company",
			url:      "https://techsolutions.io",
			wantMin:  2,
			wantKeys: []string{"tech", "technology", "software"},
		},
		{
			name:     "shop TLD",
			url:      "https://example.shop",
			wantMin:  2,
			wantKeys: []string{"retail", "ecommerce", "store", "shop"},
		},
		{
			name:     "wine TLD",
			url:      "https://vineyard.wine",
			wantMin:  3,
			wantKeys: []string{"wine", "beverage", "alcohol", "winery"},
		},
		{
			name:     "invalid URL",
			url:      "not-a-valid-url",
			wantMin:  0,
			wantKeys: []string{},
		},
		{
			name:     "URL without scheme",
			url:      "thegreenegrape.com",
			wantMin:  2,
			wantKeys: []string{"green", "grape"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := repo.extractKeywordsFromURLEnhanced(tt.url)

			if len(keywords) < tt.wantMin {
				t.Errorf("extractKeywordsFromURLEnhanced() returned %d keywords, want at least %d", len(keywords), tt.wantMin)
			}

			// Check for expected keywords
			keywordMap := make(map[string]bool)
			for _, kw := range keywords {
				keywordMap[strings.ToLower(kw.Keyword)] = true
			}

			foundCount := 0
			for _, wantKey := range tt.wantKeys {
				if keywordMap[strings.ToLower(wantKey)] {
					foundCount++
				}
			}

			// At least one expected keyword should be found
			if foundCount == 0 && len(tt.wantKeys) > 0 {
				t.Errorf("extractKeywordsFromURLEnhanced() found none of expected keywords: %v. Got: %v", tt.wantKeys, keywords)
			}

			// Verify all keywords have context
			for _, kw := range keywords {
				if kw.Context != "website_url" {
					t.Errorf("extractKeywordsFromURLEnhanced() keyword %s has wrong context: %s", kw.Keyword, kw.Context)
				}
				if strings.TrimSpace(kw.Keyword) == "" {
					t.Errorf("extractKeywordsFromURLEnhanced() returned empty keyword")
				}
			}
		})
	}
}

func TestSplitDomainName(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := &SupabaseKeywordRepository{
		logger: logger,
	}

	tests := []struct {
		name     string
		domain   string
		wantMin  int
		wantKeys []string
	}{
		{
			name:     "compound domain",
			domain:   "thegreenegrape.com",
			wantMin:  1, // Note: compound words without separators aren't split (would require word segmentation library)
			wantKeys: []string{"thegreenegrape"}, // Returns whole word as-is
		},
		{
			name:     "hyphenated domain",
			domain:   "wine-shop.com",
			wantMin:  2,
			wantKeys: []string{"wine", "shop"},
		},
		{
			name:     "camelCase domain",
			domain:   "WineShop.com",
			wantMin:  2,
			wantKeys: []string{"wine", "shop"},
		},
		{
			name:     "empty domain",
			domain:   "",
			wantMin:  0,
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := repo.splitDomainName(tt.domain)

			if len(parts) < tt.wantMin {
				t.Errorf("splitDomainName() returned %d parts, want at least %d. Got: %v", len(parts), tt.wantMin, parts)
			}

			// Check for expected parts
			partsMap := make(map[string]bool)
			for _, part := range parts {
				partsMap[strings.ToLower(part)] = true
			}

			for _, wantKey := range tt.wantKeys {
				if !partsMap[strings.ToLower(wantKey)] {
					t.Errorf("splitDomainName() missing expected part: %s. Got: %v", wantKey, parts)
				}
			}
		})
	}
}

func TestSplitCamelCase(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := &SupabaseKeywordRepository{
		logger: logger,
	}

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "camelCase",
			input: "WineShop",
			want:  []string{"wine", "shop"},
		},
		{
			name:  "single word",
			input: "wine",
			want:  []string{"wine"},
		},
		{
			name:  "empty string",
			input: "",
			want:  []string{},
		},
		{
			name:  "all lowercase",
			input: "wineshop",
			want:  []string{"wineshop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := repo.splitCamelCase(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitCamelCase() returned %d words, want %d. Got: %v, Want: %v", len(got), len(tt.want), got, tt.want)
			}
		})
	}
}

func TestExtractTLDHints(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := &SupabaseKeywordRepository{
		logger: logger,
	}

	tests := []struct {
		name     string
		url      string
		wantKeys []string
	}{
		{
			name:     "shop TLD",
			url:      "https://example.shop",
			wantKeys: []string{"retail", "ecommerce", "store", "shop"},
		},
		{
			name:     "wine TLD",
			url:      "https://vineyard.wine",
			wantKeys: []string{"wine", "beverage", "alcohol", "winery"},
		},
		{
			name:     "tech TLD",
			url:      "https://company.tech",
			wantKeys: []string{"technology", "tech", "software"},
		},
		{
			name:     "com TLD (no hints)",
			url:      "https://example.com",
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := parseURL(tt.url)
			if err != nil {
				t.Fatalf("Failed to parse URL: %v", err)
			}

			hints := repo.extractTLDHints(parsedURL)
			hintsMap := make(map[string]bool)
			for _, hint := range hints {
				hintsMap[strings.ToLower(hint)] = true
			}

			for _, wantKey := range tt.wantKeys {
				if !hintsMap[strings.ToLower(wantKey)] {
					t.Errorf("extractTLDHints() missing expected hint: %s. Got: %v", wantKey, hints)
				}
			}
		})
	}
}

func TestInferIndustryFromDomain(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := &SupabaseKeywordRepository{
		logger: logger,
	}

	tests := []struct {
		name     string
		domain   string
		wantKeys []string
	}{
		{
			name:     "wine domain",
			domain:   "wineshop.com",
			wantKeys: []string{"wine", "beverage", "alcohol", "retail"},
		},
		{
			name:     "grape domain",
			domain:   "thegreenegrape.com",
			wantKeys: []string{"wine", "grape", "beverage", "retail"},
		},
		{
			name:     "tech domain",
			domain:   "techsolutions.com",
			wantKeys: []string{"technology", "tech", "software"},
		},
		{
			name:     "no industry match",
			domain:   "example.com",
			wantKeys: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := repo.inferIndustryFromDomain(tt.domain)
			keywordMap := make(map[string]bool)
			for _, kw := range keywords {
				keywordMap[strings.ToLower(kw)] = true
			}

			for _, wantKey := range tt.wantKeys {
				if !keywordMap[strings.ToLower(wantKey)] {
					t.Errorf("inferIndustryFromDomain() missing expected keyword: %s. Got: %v", wantKey, keywords)
				}
			}
		})
	}
}

// Helper function to parse URL
func parseURL(urlStr string) (*url.URL, error) {
	parsed, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

func extractHost(urlStr string) string {
	// Simple extraction for testing
	if strings.HasPrefix(urlStr, "https://") {
		urlStr = strings.TrimPrefix(urlStr, "https://")
	}
	if strings.HasPrefix(urlStr, "http://") {
		urlStr = strings.TrimPrefix(urlStr, "http://")
	}
	parts := strings.Split(urlStr, "/")
	return parts[0]
}

