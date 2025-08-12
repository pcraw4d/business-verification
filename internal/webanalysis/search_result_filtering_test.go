package webanalysis

import (
	"strings"
	"testing"
)

func TestNewSearchResultFilter(t *testing.T) {
	filter := NewSearchResultFilter()

	if filter == nil {
		t.Fatal("Expected filter to be created, got nil")
	}

	if len(filter.filters) == 0 {
		t.Error("Expected filters to be initialized")
	}

	if len(filter.rankers) == 0 {
		t.Error("Expected rankers to be initialized")
	}

	config := filter.GetConfig()
	if config.MinRelevanceScore != 0.3 {
		t.Errorf("Expected MinRelevanceScore 0.3, got %f", config.MinRelevanceScore)
	}

	if config.MaxResultsToReturn != 20 {
		t.Errorf("Expected MaxResultsToReturn 20, got %d", config.MaxResultsToReturn)
	}

	if !config.EnableSpamFiltering {
		t.Error("Expected EnableSpamFiltering true, got false")
	}
}

func TestSearchResultFilter_FilterAndRank(t *testing.T) {
	filter := NewSearchResultFilter()

	// Reset duplicate filter state for this test
	for _, f := range filter.filters {
		if df, ok := f.(*DuplicateFilter); ok {
			df.seenURLs = make(map[string]bool)
		}
	}

	// Create test results
	results := []WebSearchResult{
		{
			Title:          "Test Business - Official Website",
			URL:            "https://testbusiness.com",
			Description:    "Official website of Test Business, a leading technology company",
			RelevanceScore: 0.9,
			Source:         "google",
		},
		{
			Title:          "Spam Result - Click Here to Buy Now",
			URL:            "https://spam.com/buy-now",
			Description:    "Limited time offer! Click here to buy now! Make money fast!",
			RelevanceScore: 0.8,
			Source:         "google",
		},
		{
			Title:          "Test Business - About Us",
			URL:            "https://testbusiness.com/about",
			Description:    "Learn more about Test Business and our mission to innovate",
			RelevanceScore: 0.7,
			Source:         "google",
		},
		{
			Title:          "Low Quality Result",
			URL:            "http://lowquality.com",
			Description:    "This is a low quality result with insufficient content and poor relevance score",
			RelevanceScore: 0.2,
			Source:         "google",
		},
	}

	query := "test business"
	filteredResults := filter.FilterAndRank(results, query)

	// Should have filtered out spam and low quality results

	// Should have filtered out spam and low quality results
	if len(filteredResults) != 2 {
		t.Errorf("Expected 2 filtered results, got %d", len(filteredResults))
	}

	// First result should be the official website (highest score)
	firstResult := filteredResults[0]
	if firstResult.Result.Title != "Test Business - Official Website" {
		t.Errorf("Expected first result to be 'Test Business - Official Website', got %s", firstResult.Result.Title)
	}

	// Check that spam was filtered out
	for _, result := range filteredResults {
		if strings.Contains(result.Result.Title, "Spam Result") {
			t.Error("Expected spam result to be filtered out")
		}
	}

	// Check that low quality result was filtered out
	for _, result := range filteredResults {
		if strings.Contains(result.Result.Title, "Low Quality Result") {
			t.Error("Expected low quality result to be filtered out")
		}
	}
}

func TestRelevanceFilter(t *testing.T) {
	filter := &RelevanceFilter{
		config: FilterConfig{MinRelevanceScore: 0.5},
	}

	highRelevanceResult := &WebSearchResult{
		Title:          "High Relevance",
		RelevanceScore: 0.8,
	}

	lowRelevanceResult := &WebSearchResult{
		Title:          "Low Relevance",
		RelevanceScore: 0.3,
	}

	if !filter.Filter(highRelevanceResult) {
		t.Error("Expected high relevance result to pass filter")
	}

	if filter.Filter(lowRelevanceResult) {
		t.Error("Expected low relevance result to be filtered out")
	}
}

func TestSpamFilter(t *testing.T) {
	filter := &SpamFilter{
		config: FilterConfig{
			EnableSpamFiltering: true,
			SpamKeywords:        []string{"click here", "buy now", "make money"},
		},
	}

	legitimateResult := &WebSearchResult{
		Title:       "Legitimate Business",
		Description: "This is a legitimate business description",
	}

	spamResult := &WebSearchResult{
		Title:       "Spam Result",
		Description: "Click here to buy now! Make money fast!",
	}

	if !filter.Filter(legitimateResult) {
		t.Error("Expected legitimate result to pass spam filter")
	}

	if filter.Filter(spamResult) {
		t.Error("Expected spam result to be filtered out")
	}
}

func TestDomainFilter(t *testing.T) {
	filter := &DomainFilter{
		config: FilterConfig{
			EnableDomainFiltering: true,
			BlockedDomains:        []string{"spam.com", "malware.com"},
		},
	}

	legitimateResult := &WebSearchResult{
		URL: "https://legitimate.com",
	}

	blockedResult := &WebSearchResult{
		URL: "https://spam.com/bad-content",
	}

	if !filter.Filter(legitimateResult) {
		t.Error("Expected legitimate result to pass domain filter")
	}

	if filter.Filter(blockedResult) {
		t.Error("Expected blocked domain result to be filtered out")
	}
}

func TestContentFilter(t *testing.T) {
	filter := &ContentFilter{
		config: FilterConfig{
			EnableContentFiltering: true,
			MinContentLength:       50,
			MaxContentLength:       1000,
		},
	}

	goodContentResult := &WebSearchResult{
		Title:       "Good Content Title",
		Description: "This is a good description with sufficient content length for filtering purposes",
	}

	shortContentResult := &WebSearchResult{
		Title:       "Short",
		Description: "Too short",
	}

	if !filter.Filter(goodContentResult) {
		t.Error("Expected good content result to pass content filter")
	}

	if filter.Filter(shortContentResult) {
		t.Error("Expected short content result to be filtered out")
	}
}

func TestDuplicateFilter(t *testing.T) {
	filter := &DuplicateFilter{}

	firstResult := &WebSearchResult{
		URL: "https://example.com/page1",
	}

	duplicateResult := &WebSearchResult{
		URL: "https://example.com/page1",
	}

	uniqueResult := &WebSearchResult{
		URL: "https://example.com/page2",
	}

	// First occurrence should pass
	if !filter.Filter(firstResult) {
		t.Error("Expected first result to pass duplicate filter")
	}

	// Duplicate should be filtered out
	if filter.Filter(duplicateResult) {
		t.Error("Expected duplicate result to be filtered out")
	}

	// Unique result should pass
	if !filter.Filter(uniqueResult) {
		t.Error("Expected unique result to pass duplicate filter")
	}
}

func TestLanguageFilter(t *testing.T) {
	filter := &LanguageFilter{
		config: FilterConfig{EnableLanguageFiltering: true},
	}

	englishResult := &WebSearchResult{
		Title:       "English Title",
		Description: "This is an English description with the word and in it",
	}

	nonEnglishResult := &WebSearchResult{
		Title:       "Non-English Title",
		Description: "Esta es una descripción en español sin palabras en inglés",
	}

	if !filter.Filter(englishResult) {
		t.Error("Expected English result to pass language filter")
	}

	if filter.Filter(nonEnglishResult) {
		t.Error("Expected non-English result to be filtered out")
	}
}

func TestQualityFilter(t *testing.T) {
	filter := &QualityFilter{
		config: FilterConfig{EnableQualityFiltering: true},
	}

	goodQualityResult := &WebSearchResult{
		Title: "Good Quality Title",
		URL:   "https://example.com",
	}

	badQualityResult := &WebSearchResult{
		Title: "",
		URL:   "http://example.com", // HTTP instead of HTTPS
	}

	if !filter.Filter(goodQualityResult) {
		t.Error("Expected good quality result to pass quality filter")
	}

	if filter.Filter(badQualityResult) {
		t.Error("Expected bad quality result to be filtered out")
	}
}

func TestRelevanceRanker(t *testing.T) {
	ranker := &RelevanceRanker{}

	result := &WebSearchResult{
		RelevanceScore: 0.8,
	}

	score := ranker.Rank(result, "test query")

	if score != 0.8 {
		t.Errorf("Expected relevance score 0.8, got %f", score)
	}
}

func TestDomainRanker(t *testing.T) {
	ranker := &DomainRanker{
		config: FilterConfig{
			PreferredDomains: []string{"wikipedia.org", "linkedin.com"},
		},
	}

	preferredResult := &WebSearchResult{
		URL: "https://wikipedia.org/article",
	}

	regularResult := &WebSearchResult{
		URL: "https://example.com/page",
	}

	suspiciousResult := &WebSearchResult{
		URL: "https://suspicious.tk/page",
	}

	preferredScore := ranker.Rank(preferredResult, "test query")
	regularScore := ranker.Rank(regularResult, "test query")
	suspiciousScore := ranker.Rank(suspiciousResult, "test query")

	if preferredScore <= regularScore {
		t.Error("Expected preferred domain to have higher score")
	}

	if suspiciousScore >= regularScore {
		t.Error("Expected suspicious domain to have lower score")
	}
}

func TestContentRanker(t *testing.T) {
	ranker := &ContentRanker{}

	goodContentResult := &WebSearchResult{
		Title:       "Good Content - Structured Title",
		Description: "This is a good description with sufficient content length for ranking purposes",
	}

	shortContentResult := &WebSearchResult{
		Title:       "Short Title",
		Description: "Short description",
	}

	goodScore := ranker.Rank(goodContentResult, "test query")
	shortScore := ranker.Rank(shortContentResult, "test query")

	if goodScore <= shortScore {
		t.Error("Expected good content to have higher score")
	}
}

func TestAuthorityRanker(t *testing.T) {
	ranker := &AuthorityRanker{}

	authoritativeResult := &WebSearchResult{
		URL: "https://wikipedia.org/article",
	}

	regularResult := &WebSearchResult{
		URL: "https://example.com/page",
	}

	authScore := ranker.Rank(authoritativeResult, "test query")
	regularScore := ranker.Rank(regularResult, "test query")

	if authScore <= regularScore {
		t.Error("Expected authoritative domain to have higher score")
	}
}

func TestSearchResultFilter_UpdateConfig(t *testing.T) {
	filter := NewSearchResultFilter()

	newConfig := FilterConfig{
		MinRelevanceScore:        0.5,
		MaxResultsToReturn:       10,
		EnableSpamFiltering:      false,
		EnableDuplicateFiltering: false,
		EnableQualityFiltering:   false,
		EnableDomainFiltering:    false,
		EnableLanguageFiltering:  false,
		EnableDateFiltering:      true,
		EnableContentFiltering:   false,
		SpamKeywords:             []string{"custom", "spam", "keywords"},
		BlockedDomains:           []string{"custom", "blocked", "domains"},
		PreferredDomains:         []string{"custom", "preferred", "domains"},
		AllowedLanguages:         []string{"es", "fr"},
		MinContentLength:         100,
		MaxContentLength:         5000,
	}

	filter.UpdateConfig(newConfig)

	updatedConfig := filter.GetConfig()

	if updatedConfig.MinRelevanceScore != 0.5 {
		t.Errorf("Expected MinRelevanceScore 0.5, got %f", updatedConfig.MinRelevanceScore)
	}

	if updatedConfig.MaxResultsToReturn != 10 {
		t.Errorf("Expected MaxResultsToReturn 10, got %d", updatedConfig.MaxResultsToReturn)
	}

	if updatedConfig.EnableSpamFiltering {
		t.Error("Expected EnableSpamFiltering false, got true")
	}

	if len(updatedConfig.SpamKeywords) != 3 {
		t.Errorf("Expected 3 spam keywords, got %d", len(updatedConfig.SpamKeywords))
	}

	if len(updatedConfig.BlockedDomains) != 3 {
		t.Errorf("Expected 3 blocked domains, got %d", len(updatedConfig.BlockedDomains))
	}

	if len(updatedConfig.PreferredDomains) != 3 {
		t.Errorf("Expected 3 preferred domains, got %d", len(updatedConfig.PreferredDomains))
	}

	if len(updatedConfig.AllowedLanguages) != 2 {
		t.Errorf("Expected 2 allowed languages, got %d", len(updatedConfig.AllowedLanguages))
	}

	if updatedConfig.MinContentLength != 100 {
		t.Errorf("Expected MinContentLength 100, got %d", updatedConfig.MinContentLength)
	}

	if updatedConfig.MaxContentLength != 5000 {
		t.Errorf("Expected MaxContentLength 5000, got %d", updatedConfig.MaxContentLength)
	}
}

func TestSearchResultFilter_GetStats(t *testing.T) {
	filter := NewSearchResultFilter()

	stats := filter.GetStats()

	if stats["total_filters"] == nil {
		t.Error("Expected total_filters in stats")
	}

	if stats["total_rankers"] == nil {
		t.Error("Expected total_rankers in stats")
	}

	if stats["config"] == nil {
		t.Error("Expected config in stats")
	}

	totalFilters := stats["total_filters"].(int)
	if totalFilters != 7 {
		t.Errorf("Expected 7 filters, got %d", totalFilters)
	}

	totalRankers := stats["total_rankers"].(int)
	if totalRankers != 6 {
		t.Errorf("Expected 6 rankers, got %d", totalRankers)
	}
}
