package repository

import (
	"context"
	"log"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

// TestExtractBusinessKeywords tests the extractBusinessKeywords function
func TestExtractBusinessKeywords(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	tests := []struct {
		name     string
		content  string
		wantMin  int
		wantKeys []string
	}{
		{
			name:     "wine retailer content",
			content:  "We are a wine shop and retail store selling fine wines, spirits, and beverages. Visit our wine bar for tastings.",
			wantMin:  5,
			wantKeys: []string{"wine", "shop", "retail", "store", "beverage", "spirits"},
		},
		{
			name:     "technology company content",
			content:  "Software development company specializing in cloud computing, AI, and digital transformation services.",
			wantMin:  5,
			wantKeys: []string{"software", "technology", "cloud", "development", "digital"},
		},
		{
			name:     "e-commerce content",
			content:  "Online store and e-commerce platform for retail sales and digital commerce.",
			wantMin:  4,
			wantKeys: []string{"online", "store", "ecommerce", "retail", "commerce"},
		},
		{
			name:     "healthcare content",
			content:  "Medical services and healthcare provider offering patient care and treatment.",
			wantMin:  4,
			wantKeys: []string{"medical", "healthcare", "health", "care", "treatment"},
		},
		{
			name:     "finance content",
			content:  "Financial services and banking solutions for investment and wealth management.",
			wantMin:  4,
			wantKeys: []string{"financial", "banking", "investment", "finance", "wealth"},
		},
		{
			name:     "empty content",
			content:  "",
			wantMin:  0,
			wantKeys: []string{},
		},
		{
			name:     "content with no business keywords",
			content:  "This is just some random text without any business-related terms.",
			wantMin:  0,
			wantKeys: []string{},
		},
		{
			name:     "mixed content with phrases",
			content:  "Wine shop offering wine tasting and retail sales. Our wine store has a wine bar.",
			wantMin:  5,
			wantKeys: []string{"wine", "shop", "retail", "store", "bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := repo.extractBusinessKeywords(tt.content)

			if len(keywords) < tt.wantMin {
				t.Errorf("extractBusinessKeywords() returned %d keywords, want at least %d. Got: %v", len(keywords), tt.wantMin, keywords)
			}

			// Check for expected keywords
			keywordMap := make(map[string]bool)
			for _, kw := range keywords {
				keywordMap[strings.ToLower(kw)] = true
			}

			foundCount := 0
			for _, wantKey := range tt.wantKeys {
				if keywordMap[strings.ToLower(wantKey)] {
					foundCount++
				}
			}

			// At least some expected keywords should be found
			if foundCount == 0 && len(tt.wantKeys) > 0 {
				t.Errorf("extractBusinessKeywords() found none of expected keywords: %v. Got: %v", tt.wantKeys, keywords)
			}

			// Verify no empty keywords
			for _, kw := range keywords {
				if strings.TrimSpace(kw) == "" {
					t.Errorf("extractBusinessKeywords() returned empty keyword")
				}
			}
		})
	}
}

// TestExtractTextFromHTML tests the extractTextFromHTML function with Phase 9.1 optimizations
func TestExtractTextFromHTML_Repository(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	tests := []struct {
		name     string
		html     string
		wantText string
		wantMin  int // Minimum text length expected
	}{
		{
			name:     "simple HTML",
			html:     "<html><body><p>Hello World</p></body></html>",
			wantText: "Hello World",
			wantMin:  10,
		},
		{
			name:     "with scripts and styles",
			html:     "<html><head><style>body { color: red; }</style><script>alert('test');</script></head><body><p>Content</p></body></html>",
			wantText: "Content",
			wantMin:  7,
		},
		{
			name:     "with multiple paragraphs",
			html:     "<html><body><p>First paragraph</p><p>Second paragraph</p></body></html>",
			wantText: "First paragraph Second paragraph",
			wantMin:  30,
		},
		{
			name:     "with headings",
			html:     "<html><body><h1>Title</h1><h2>Subtitle</h2><p>Content</p></body></html>",
			wantText: "Title Subtitle Content",
			wantMin:  20,
		},
		{
			name:     "empty HTML",
			html:     "<html><body></body></html>",
			wantText: "",
			wantMin:  0,
		},
		{
			name:     "large content (should be limited to 50KB)",
			html:     "<html><body><p>" + strings.Repeat("A", 60000) + "</p></body></html>",
			wantText: strings.Repeat("A", 50000),
			wantMin:  50000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text := repo.extractTextFromHTML(tt.html)

			if len(text) < tt.wantMin {
				t.Errorf("extractTextFromHTML() returned text of length %d, want at least %d", len(text), tt.wantMin)
			}

			if tt.wantText != "" && !strings.Contains(text, tt.wantText) {
				t.Errorf("extractTextFromHTML() = %q, want to contain %q", text, tt.wantText)
			}

			// Verify no HTML tags remain
			if strings.Contains(text, "<") || strings.Contains(text, ">") {
				t.Errorf("extractTextFromHTML() still contains HTML tags: %q", text)
			}
		})
	}
}

// TestGetCachedRegex tests the regex caching functionality (Phase 9.1)
func TestGetCachedRegex(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	pattern := `\b(wine|shop)\b`

	// First call should compile and cache
	regex1 := repo.getCachedRegex(pattern)
	if regex1 == nil {
		t.Fatal("getCachedRegex() returned nil")
	}

	// Second call should return cached version
	regex2 := repo.getCachedRegex(pattern)
	if regex2 == nil {
		t.Fatal("getCachedRegex() returned nil on second call")
	}

	// Should be the same instance (pointer comparison)
	if regex1 != regex2 {
		t.Error("getCachedRegex() did not return cached regex on second call")
	}

	// Test that it works correctly
	testString := "wine shop"
	if !regex1.MatchString(testString) {
		t.Error("Cached regex does not match expected string")
	}

	// Test with different pattern (should compile new regex)
	pattern2 := `\b(retail|store)\b`
	regex3 := repo.getCachedRegex(pattern2)
	if regex3 == nil {
		t.Fatal("getCachedRegex() returned nil for different pattern")
	}
	if regex3 == regex1 {
		t.Error("getCachedRegex() returned same regex for different pattern")
	}

	// Test that both patterns are cached
	regex1Again := repo.getCachedRegex(pattern)
	regex3Again := repo.getCachedRegex(pattern2)
	if regex1Again != regex1 {
		t.Error("getCachedRegex() did not return cached regex for first pattern")
	}
	if regex3Again != regex3 {
		t.Error("getCachedRegex() did not return cached regex for second pattern")
	}
}

// TestGetCachedRegex_Concurrent tests concurrent access to regex cache (Phase 9.1)
func TestGetCachedRegex_Concurrent(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	patterns := []string{
		`\b(wine|shop)\b`,
		`\b(retail|store)\b`,
		`\b(technology|software)\b`,
		`\b(healthcare|medical)\b`,
	}

	// Test concurrent access
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()
			pattern := patterns[id%len(patterns)]
			regex := repo.getCachedRegex(pattern)
			if regex == nil {
				t.Errorf("Goroutine %d: getCachedRegex() returned nil", id)
			}
			// Verify it works
			if !regex.MatchString("test") {
				// This is fine, just checking it doesn't panic
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all patterns are cached
	for _, pattern := range patterns {
		regex1 := repo.getCachedRegex(pattern)
		regex2 := repo.getCachedRegex(pattern)
		if regex1 != regex2 {
			t.Errorf("Pattern %s: getCachedRegex() did not return cached regex", pattern)
		}
	}
}

// TestGetCachedDNSResolution tests DNS caching functionality (Phase 9.2)
func TestGetCachedDNSResolution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping DNS test in short mode")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resolver := &net.Resolver{
		PreferGo: true,
	}

	// Test DNS lookup with caching
	host := "google.com"
	
	// First call should perform DNS lookup
	ips1, err := repo.getCachedDNSResolution(host, resolver, ctx)
	if err != nil {
		t.Fatalf("getCachedDNSResolution() failed: %v", err)
	}
	if len(ips1) == 0 {
		t.Fatal("getCachedDNSResolution() returned no IPs")
	}

	// Second call should use cache
	ips2, err := repo.getCachedDNSResolution(host, resolver, ctx)
	if err != nil {
		t.Fatalf("getCachedDNSResolution() failed on second call: %v", err)
	}

	// Should return same IPs
	if len(ips1) != len(ips2) {
		t.Errorf("getCachedDNSResolution() returned different number of IPs: %d vs %d", len(ips1), len(ips2))
	}

	// Verify IPs match
	for i, ip1 := range ips1 {
		if i < len(ips2) && ip1.IP.String() != ips2[i].IP.String() {
			t.Errorf("getCachedDNSResolution() returned different IP at index %d: %s vs %s", i, ip1.IP.String(), ips2[i].IP.String())
		}
	}
}

// TestApplyRateLimit tests rate limiting functionality (Phase 9.3)
func TestApplyRateLimit(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	domain := "example.com"

	// First call should not delay (no previous request)
	start := time.Now()
	repo.applyRateLimit(domain)
	elapsed1 := time.Since(start)
	// Allow some tolerance for system scheduling
	if elapsed1 > 200*time.Millisecond {
		t.Errorf("applyRateLimit() delayed on first call: %v (expected < 200ms)", elapsed1)
	}

	// Second call immediately after should delay
	start = time.Now()
	repo.applyRateLimit(domain)
	elapsed2 := time.Since(start)
	// Should wait at least ~1 second (with jitter), but allow some tolerance
	if elapsed2 < 800*time.Millisecond {
		t.Errorf("applyRateLimit() did not delay on second call: %v (expected >= 800ms)", elapsed2)
	}
	// Should not wait too long (max 1.2s base + 20% jitter = ~1.44s, allow up to 2s for safety)
	if elapsed2 > 2*time.Second {
		t.Errorf("applyRateLimit() delayed too long on second call: %v (expected <= 2s)", elapsed2)
	}

	// Third call after delay should not delay (enough time has passed)
	time.Sleep(1200 * time.Millisecond) // Wait slightly more than minDelay
	start = time.Now()
	repo.applyRateLimit(domain)
	elapsed3 := time.Since(start)
	// Allow some tolerance for system scheduling
	if elapsed3 > 200*time.Millisecond {
		t.Errorf("applyRateLimit() delayed on third call after wait: %v (expected < 200ms)", elapsed3)
	}
}

// TestExtractKeywordsFallbackChain tests the fallback chain (Phase 5)
func TestExtractKeywordsFallbackChain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	tests := []struct {
		name          string
		businessName  string
		websiteURL    string
		wantMin       int
		expectedLevel string // Expected fallback level used
	}{
		{
			name:          "business name only (with brand match)",
			businessName:  "Marriott Hotel",
			websiteURL:    "",
			wantMin:       1, // Brand matches in MCC 3000-3831 return keywords
			expectedLevel: "business_name",
		},
		{
			name:          "invalid URL (should fallback to business name)",
			businessName:  "Test Business",
			websiteURL:    "not-a-valid-url",
			wantMin:       2,
			expectedLevel: "business_name",
		},
		{
			name:          "URL without scheme (should add https)",
			businessName:  "",
			websiteURL:    "example.com",
			wantMin:       1,
			expectedLevel: "url_only",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			keywords := repo.extractKeywords(ctx, tt.businessName, tt.websiteURL)

			if len(keywords) < tt.wantMin {
				t.Errorf("extractKeywords() returned %d keywords, want at least %d. Got: %v", len(keywords), tt.wantMin, keywords)
			}

			// Verify all keywords have context
			for _, kw := range keywords {
				if kw.Context == "" {
					t.Errorf("extractKeywords() returned keyword without context: %s", kw.Keyword)
				}
				if strings.TrimSpace(kw.Keyword) == "" {
					t.Errorf("extractKeywords() returned empty keyword")
				}
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
		})
	}
}

// TestExtractBusinessKeywords_EdgeCases tests edge cases for extractBusinessKeywords
func TestExtractBusinessKeywords_EdgeCases(t *testing.T) {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	repo := NewSupabaseKeywordRepositoryWithInterface(&MockSupabaseClient{}, logger)

	tests := []struct {
		name    string
		content string
		wantMin int
		wantMax int // Maximum expected keywords (to catch excessive extraction)
	}{
		{
			name:    "very long content",
			content: strings.Repeat("wine shop retail store beverage ", 100),
			wantMin: 4,
			wantMax: 20, // Should not extract duplicates excessively
		},
		{
			name:    "content with special characters",
			content: "Wine & Spirits! Retail@Store #Beverage$",
			wantMin: 3,
			wantMax: 10,
		},
		{
			name:    "content with numbers",
			content: "Store #123, Wine Shop 456, Retail 789",
			wantMin: 3,
			wantMax: 10,
		},
		{
			name:    "mixed case content",
			content: "WINE SHOP Retail Store BEVERAGE wine shop",
			wantMin: 3,
			wantMax: 10,
		},
		{
			name:    "content with newlines and tabs",
			content: "Wine\nShop\tRetail\nStore",
			wantMin: 3,
			wantMax: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := repo.extractBusinessKeywords(tt.content)

			if len(keywords) < tt.wantMin {
				t.Errorf("extractBusinessKeywords() returned %d keywords, want at least %d", len(keywords), tt.wantMin)
			}

			if len(keywords) > tt.wantMax {
				t.Errorf("extractBusinessKeywords() returned %d keywords, want at most %d (possible duplicate extraction)", len(keywords), tt.wantMax)
			}

			// Verify no empty keywords
			for _, kw := range keywords {
				if strings.TrimSpace(kw) == "" {
					t.Errorf("extractBusinessKeywords() returned empty keyword")
				}
			}

			// Verify no duplicate keywords (case-insensitive)
			seen := make(map[string]bool)
			for _, kw := range keywords {
				key := strings.ToLower(kw)
				if seen[key] {
					t.Errorf("extractBusinessKeywords() returned duplicate keyword: %s", kw)
				}
				seen[key] = true
			}
		})
	}
}

