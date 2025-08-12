package webanalysis

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewWebSearchIntegration(t *testing.T) {
	wsi := NewWebSearchIntegration()

	if wsi == nil {
		t.Fatal("Expected non-nil WebSearchIntegration")
	}

	if wsi.resultAnalyzer == nil {
		t.Error("Expected non-nil WebSearchResultAnalyzer")
	}

	if wsi.queryOptimizer == nil {
		t.Error("Expected non-nil WebSearchQueryOptimizer")
	}

	if wsi.rankingEngine == nil {
		t.Error("Expected non-nil WebSearchRankingEngine")
	}

	if wsi.extractionEngine == nil {
		t.Error("Expected non-nil WebBusinessExtractionEngine")
	}
}

func TestWebSearchIntegrationConfiguration(t *testing.T) {
	wsi := NewWebSearchIntegration()

	// Test default configuration
	if !wsi.config.EnableMultiSourceSearch {
		t.Error("Expected multi-source search to be enabled by default")
	}

	if !wsi.config.EnableResultAnalysis {
		t.Error("Expected result analysis to be enabled by default")
	}

	if !wsi.config.EnableQueryOptimization {
		t.Error("Expected query optimization to be enabled by default")
	}

	if !wsi.config.EnableResultRanking {
		t.Error("Expected result ranking to be enabled by default")
	}

	if !wsi.config.EnableBusinessExtraction {
		t.Error("Expected business extraction to be enabled by default")
	}

	if wsi.config.MaxResultsPerEngine != 10 {
		t.Errorf("Expected max results per engine to be 10, got: %d", wsi.config.MaxResultsPerEngine)
	}

	if wsi.config.SearchTimeout != time.Second*30 {
		t.Errorf("Expected search timeout to be 30 seconds, got: %v", wsi.config.SearchTimeout)
	}

	if wsi.config.RetryAttempts != 3 {
		t.Errorf("Expected retry attempts to be 3, got: %d", wsi.config.RetryAttempts)
	}

	if wsi.config.RateLimitDelay != time.Millisecond*500 {
		t.Errorf("Expected rate limit delay to be 500ms, got: %v", wsi.config.RateLimitDelay)
	}
}

func TestNewGoogleWebSearchEngine(t *testing.T) {
	apiKey := "test-api-key"
	cx := "test-cx"

	gwse := NewGoogleWebSearchEngine(apiKey, cx)

	if gwse == nil {
		t.Fatal("Expected non-nil GoogleWebSearchEngine")
	}

	if gwse.apiKey != apiKey {
		t.Errorf("Expected API key %s, got: %s", apiKey, gwse.apiKey)
	}

	if gwse.cx != cx {
		t.Errorf("Expected CX %s, got: %s", cx, gwse.cx)
	}

	if gwse.rateLimit != time.Millisecond*1000 {
		t.Errorf("Expected rate limit 1 second, got: %v", gwse.rateLimit)
	}

	if gwse.httpClient == nil {
		t.Error("Expected non-nil HTTP client")
	}
}

func TestNewBingWebSearchEngine(t *testing.T) {
	apiKey := "test-api-key"

	bwse := NewBingWebSearchEngine(apiKey)

	if bwse == nil {
		t.Fatal("Expected non-nil BingWebSearchEngine")
	}

	if bwse.apiKey != apiKey {
		t.Errorf("Expected API key %s, got: %s", apiKey, bwse.apiKey)
	}

	if bwse.rateLimit != time.Millisecond*1000 {
		t.Errorf("Expected rate limit 1 second, got: %v", bwse.rateLimit)
	}

	if bwse.httpClient == nil {
		t.Error("Expected non-nil HTTP client")
	}
}

func TestNewDuckDuckGoWebSearchEngine(t *testing.T) {
	ddgwse := NewDuckDuckGoWebSearchEngine()

	if ddgwse == nil {
		t.Fatal("Expected non-nil DuckDuckGoWebSearchEngine")
	}

	if ddgwse.rateLimit != time.Millisecond*500 {
		t.Errorf("Expected rate limit 500ms, got: %v", ddgwse.rateLimit)
	}

	if ddgwse.httpClient == nil {
		t.Error("Expected non-nil HTTP client")
	}
}

func TestGoogleWebSearchEngineInterface(t *testing.T) {
	gwse := NewGoogleWebSearchEngine("test-key", "test-cx")

	// Test GetName
	name := gwse.GetName()
	if name != "Google" {
		t.Errorf("Expected name 'Google', got: %s", name)
	}

	// Test GetRateLimit
	rateLimit := gwse.GetRateLimit()
	if rateLimit != time.Millisecond*1000 {
		t.Errorf("Expected rate limit 1 second, got: %v", rateLimit)
	}
}

func TestBingWebSearchEngineInterface(t *testing.T) {
	bwse := NewBingWebSearchEngine("test-key")

	// Test GetName
	name := bwse.GetName()
	if name != "Bing" {
		t.Errorf("Expected name 'Bing', got: %s", name)
	}

	// Test GetRateLimit
	rateLimit := bwse.GetRateLimit()
	if rateLimit != time.Millisecond*1000 {
		t.Errorf("Expected rate limit 1 second, got: %v", rateLimit)
	}
}

func TestDuckDuckGoWebSearchEngineInterface(t *testing.T) {
	ddgwse := NewDuckDuckGoWebSearchEngine()

	// Test GetName
	name := ddgwse.GetName()
	if name != "DuckDuckGo" {
		t.Errorf("Expected name 'DuckDuckGo', got: %s", name)
	}

	// Test GetRateLimit
	rateLimit := ddgwse.GetRateLimit()
	if rateLimit != time.Millisecond*500 {
		t.Errorf("Expected rate limit 500ms, got: %v", rateLimit)
	}
}

func TestNewWebSearchResultAnalyzer(t *testing.T) {
	wsra := NewWebSearchResultAnalyzer()

	if wsra == nil {
		t.Fatal("Expected non-nil WebSearchResultAnalyzer")
	}

	if len(wsra.filters) == 0 {
		t.Error("Expected non-empty filters")
	}

	if len(wsra.analyzers) == 0 {
		t.Error("Expected non-empty analyzers")
	}
}

func TestWebSearchResultAnalyzerConfiguration(t *testing.T) {
	wsra := NewWebSearchResultAnalyzer()

	// Test default configuration
	if !wsra.config.EnableContentAnalysis {
		t.Error("Expected content analysis to be enabled by default")
	}

	if !wsra.config.EnableSpamDetection {
		t.Error("Expected spam detection to be enabled by default")
	}

	if !wsra.config.EnableDuplicateDetection {
		t.Error("Expected duplicate detection to be enabled by default")
	}

	if wsra.config.MinRelevanceScore != 0.3 {
		t.Errorf("Expected min relevance score to be 0.3, got: %f", wsra.config.MinRelevanceScore)
	}

	if wsra.config.MaxResultsToAnalyze != 50 {
		t.Errorf("Expected max results to analyze to be 50, got: %d", wsra.config.MaxResultsToAnalyze)
	}
}

func TestNewWebSearchQueryOptimizer(t *testing.T) {
	wsqo := NewWebSearchQueryOptimizer()

	if wsqo == nil {
		t.Fatal("Expected non-nil WebSearchQueryOptimizer")
	}

	if len(wsqo.optimizers) == 0 {
		t.Error("Expected non-empty optimizers")
	}
}

func TestWebSearchQueryOptimizerConfiguration(t *testing.T) {
	wsqo := NewWebSearchQueryOptimizer()

	// Test default configuration
	if !wsqo.config.EnableQueryExpansion {
		t.Error("Expected query expansion to be enabled by default")
	}

	if !wsqo.config.EnableQueryRefinement {
		t.Error("Expected query refinement to be enabled by default")
	}

	if !wsqo.config.EnableSynonymExpansion {
		t.Error("Expected synonym expansion to be enabled by default")
	}

	if wsqo.config.MaxQueryLength != 100 {
		t.Errorf("Expected max query length to be 100, got: %d", wsqo.config.MaxQueryLength)
	}

	if wsqo.config.MinQueryLength != 3 {
		t.Errorf("Expected min query length to be 3, got: %d", wsqo.config.MinQueryLength)
	}
}

func TestNewWebSearchRankingEngine(t *testing.T) {
	wsre := NewWebSearchRankingEngine()

	if wsre == nil {
		t.Fatal("Expected non-nil WebSearchRankingEngine")
	}

	if len(wsre.rankingFactors) == 0 {
		t.Error("Expected non-empty ranking factors")
	}
}

func TestWebSearchRankingEngineConfiguration(t *testing.T) {
	wsre := NewWebSearchRankingEngine()

	// Test default configuration
	if !wsre.config.EnableMultiFactorRanking {
		t.Error("Expected multi-factor ranking to be enabled by default")
	}

	if wsre.config.TitleWeight != 0.3 {
		t.Errorf("Expected title weight to be 0.3, got: %f", wsre.config.TitleWeight)
	}

	if wsre.config.ContentWeight != 0.4 {
		t.Errorf("Expected content weight to be 0.4, got: %f", wsre.config.ContentWeight)
	}

	if wsre.config.URLWeight != 0.1 {
		t.Errorf("Expected URL weight to be 0.1, got: %f", wsre.config.URLWeight)
	}

	if wsre.config.FreshnessWeight != 0.1 {
		t.Errorf("Expected freshness weight to be 0.1, got: %f", wsre.config.FreshnessWeight)
	}

	if wsre.config.AuthorityWeight != 0.1 {
		t.Errorf("Expected authority weight to be 0.1, got: %f", wsre.config.AuthorityWeight)
	}
}

func TestNewWebBusinessExtractionEngine(t *testing.T) {
	wbee := NewWebBusinessExtractionEngine()

	if wbee == nil {
		t.Fatal("Expected non-nil WebBusinessExtractionEngine")
	}

	if len(wbee.extractors) == 0 {
		t.Error("Expected non-empty extractors")
	}
}

func TestWebBusinessExtractionEngineConfiguration(t *testing.T) {
	wbee := NewWebBusinessExtractionEngine()

	// Test default configuration
	if !wbee.config.EnableNameExtraction {
		t.Error("Expected name extraction to be enabled by default")
	}

	if !wbee.config.EnableAddressExtraction {
		t.Error("Expected address extraction to be enabled by default")
	}

	if !wbee.config.EnableContactExtraction {
		t.Error("Expected contact extraction to be enabled by default")
	}

	if !wbee.config.EnableIndustryExtraction {
		t.Error("Expected industry extraction to be enabled by default")
	}

	if !wbee.config.EnableSizeExtraction {
		t.Error("Expected size extraction to be enabled by default")
	}
}

func TestWebSearchQueryOptimization(t *testing.T) {
	wsqo := NewWebSearchQueryOptimizer()

	// Test query optimization
	originalQuery := "acme corp"
	optimizedQuery := wsqo.OptimizeQuery(originalQuery)

	if optimizedQuery == "" {
		t.Error("Expected non-empty optimized query")
	}

	// Test query length constraints
	longQuery := "this is a very long query that should be truncated to meet the maximum length requirements"
	optimizedLongQuery := wsqo.OptimizeQuery(longQuery)

	if len(optimizedLongQuery) > wsqo.config.MaxQueryLength {
		t.Errorf("Expected optimized query length to be <= %d, got: %d", wsqo.config.MaxQueryLength, len(optimizedLongQuery))
	}

	// Test short query fallback
	shortQuery := "ab"
	optimizedShortQuery := wsqo.OptimizeQuery(shortQuery)

	if optimizedShortQuery != shortQuery {
		t.Errorf("Expected short query to remain unchanged, got: %s", optimizedShortQuery)
	}
}

func TestWebSearchResultRanking(t *testing.T) {
	wsre := NewWebSearchRankingEngine()

	// Create test results
	results := []WebSearchResult{
		{
			Title:          "Test Result 1",
			URL:            "https://example.com/1",
			Description:    "This is a test result",
			RelevanceScore: 0.5,
			Rank:           1,
			Source:         "test",
		},
		{
			Title:          "Test Result 2",
			URL:            "https://example.com/2",
			Description:    "This is another test result",
			RelevanceScore: 0.8,
			Rank:           2,
			Source:         "test",
		},
	}

	// Test ranking
	rankedResults := wsre.RankResults(results, "test query")

	if len(rankedResults) != len(results) {
		t.Errorf("Expected %d results, got: %d", len(results), len(rankedResults))
	}

	// Check that results are sorted by relevance score (descending)
	if rankedResults[0].RelevanceScore < rankedResults[1].RelevanceScore {
		t.Error("Expected results to be sorted by relevance score (descending)")
	}

	// Check that ranks are updated
	for i, result := range rankedResults {
		if result.Rank != i+1 {
			t.Errorf("Expected rank %d, got: %d", i+1, result.Rank)
		}
	}
}

func TestWebBusinessExtraction(t *testing.T) {
	wbee := NewWebBusinessExtractionEngine()

	// Create test results
	results := []WebSearchResult{
		{
			Title:       "Acme Corporation - Leading Technology Solutions",
			URL:         "https://acme.com",
			Description: "Acme Corporation is a technology company based in San Francisco, CA. Contact us at info@acme.com or call (555) 123-4567.",
			Source:      "test",
		},
		{
			Title:       "Tech Solutions Inc - Software Development",
			URL:         "https://techsolutions.com",
			Description: "Tech Solutions Inc provides software development services. Located at 123 Main St, New York, NY 10001.",
			Source:      "test",
		},
	}

	// Test business information extraction
	extractedInfo := wbee.ExtractBusinessInfo(results)

	if extractedInfo == nil {
		t.Fatal("Expected non-nil extracted info")
	}

	// Check that business names were extracted
	if businessNames, exists := extractedInfo["business_name"]; exists {
		if len(businessNames) == 0 {
			t.Error("Expected business names to be extracted")
		}
	}

	// Check that addresses were extracted
	if addresses, exists := extractedInfo["address"]; exists {
		if len(addresses) == 0 {
			t.Error("Expected addresses to be extracted")
		}
	}

	// Check that contact info was extracted
	if contacts, exists := extractedInfo["contact"]; exists {
		if len(contacts) == 0 {
			t.Error("Expected contact information to be extracted")
		}
	}

	// Check that industry info was extracted
	if industries, exists := extractedInfo["industry"]; exists {
		if len(industries) == 0 {
			t.Error("Expected industry information to be extracted")
		}
	}
}

func TestWebSearchResultFiltering(t *testing.T) {
	wsra := NewWebSearchResultAnalyzer()

	// Create test responses
	responses := []*WebSearchResponse{
		{
			EngineName: "test",
			Query:      "test query",
			Results: []WebSearchResult{
				{
					Title:          "Legitimate Business Result",
					URL:            "https://legitimate.com",
					Description:    "This is a legitimate business result with good content",
					RelevanceScore: 0.8,
					Source:         "test",
				},
				{
					Title:          "Spam Result - Click Here to Buy Now!",
					URL:            "https://spam.com",
					Description:    "Click here to buy now! Limited time offer! Act now!",
					RelevanceScore: 0.2,
					Source:         "test",
				},
			},
			TotalResults: 2,
		},
	}

	// Test result analysis
	analyzedResults := wsra.AnalyzeResults(responses)

	if len(analyzedResults) == 0 {
		t.Error("Expected at least one result after analysis")
	}

	// Check that spam results are filtered out
	for _, result := range analyzedResults {
		if strings.Contains(strings.ToLower(result.Title), "spam") {
			t.Error("Expected spam results to be filtered out")
		}
	}
}

func TestWebSearchIntegrationSearch(t *testing.T) {
	wsi := NewWebSearchIntegration()

	// Test search with empty engines (should not panic)
	ctx := context.Background()
	response, err := wsi.Search(ctx, "test query")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected non-nil response")
	}

	// Query optimizer expands the query, so we check that it contains the original query
	if !strings.Contains(response.Query, "test query") {
		t.Errorf("Expected query to contain 'test query', got: %s", response.Query)
	}

	if response.TotalResults != 0 {
		t.Errorf("Expected 0 results for empty engines, got: %d", response.TotalResults)
	}

	if response.SearchTime <= 0 {
		t.Error("Expected positive search time")
	}
}

func TestWebSearchResultAnalyzerFilters(t *testing.T) {
	// Test relevance filter
	highRelevanceResult := &WebSearchResult{
		Title:          "High Relevance",
		RelevanceScore: 0.8,
	}

	lowRelevanceResult := &WebSearchResult{
		Title:          "Low Relevance",
		RelevanceScore: 0.2,
	}

	// Test relevance filtering
	if !filterByWebRelevance(highRelevanceResult) {
		t.Error("Expected high relevance result to pass filter")
	}

	if filterByWebRelevance(lowRelevanceResult) {
		t.Error("Expected low relevance result to be filtered out")
	}

	// Test spam filter
	legitimateResult := &WebSearchResult{
		Title:       "Legitimate Business",
		Description: "This is a legitimate business description",
	}

	spamResult := &WebSearchResult{
		Title:       "Spam Result",
		Description: "Click here to buy now! Limited time offer!",
	}

	// Test spam filtering
	if !filterWebSpam(legitimateResult) {
		t.Error("Expected legitimate result to pass spam filter")
	}

	if filterWebSpam(spamResult) {
		t.Error("Expected spam result to be filtered out")
	}
}

func TestWebSearchQueryOptimizers(t *testing.T) {
	// Test query expansion
	originalQuery := "acme"
	expandedQuery := expandWebQuery(originalQuery)

	if !strings.Contains(expandedQuery, originalQuery) {
		t.Error("Expected expanded query to contain original query")
	}

	if !strings.Contains(expandedQuery, "company") || !strings.Contains(expandedQuery, "business") {
		t.Error("Expected expanded query to contain additional terms")
	}

	// Test query refinement
	unrefinedQuery := "  acme corp  "
	refinedQuery := refineWebQuery(unrefinedQuery)

	if refinedQuery != "acme corp" {
		t.Errorf("Expected refined query 'acme corp', got: '%s'", refinedQuery)
	}

	// Test synonym expansion
	synonymQuery := "acme corp"
	expandedSynonymQuery := expandWebSynonyms(synonymQuery)

	if !strings.Contains(expandedSynonymQuery, "corporation") {
		t.Error("Expected synonym expansion to replace 'corp' with 'corporation'")
	}
}

func TestWebSearchRankingFactors(t *testing.T) {
	// Test URL quality calculation
	httpsResult := &WebSearchResult{
		URL: "https://example.com",
	}

	httpResult := &WebSearchResult{
		URL: "http://example.com",
	}

	otherResult := &WebSearchResult{
		URL: "ftp://example.com",
	}

	httpsScore := calculateWebURLQuality(httpsResult)
	httpScore := calculateWebURLQuality(httpResult)
	otherScore := calculateWebURLQuality(otherResult)

	if httpsScore != 1.0 {
		t.Errorf("Expected HTTPS score 1.0, got: %f", httpsScore)
	}

	if httpScore != 0.8 {
		t.Errorf("Expected HTTP score 0.8, got: %f", httpScore)
	}

	if otherScore != 0.5 {
		t.Errorf("Expected other protocol score 0.5, got: %f", otherScore)
	}

	// Test authority calculation
	authorityResult := &WebSearchResult{
		URL: "https://wikipedia.org/article",
	}

	nonAuthorityResult := &WebSearchResult{
		URL: "https://unknown-site.com",
	}

	authorityScore := calculateWebAuthority(authorityResult)
	nonAuthorityScore := calculateWebAuthority(nonAuthorityResult)

	if authorityScore != 1.0 {
		t.Errorf("Expected authority score 1.0, got: %f", authorityScore)
	}

	if nonAuthorityScore != 0.5 {
		t.Errorf("Expected non-authority score 0.5, got: %f", nonAuthorityScore)
	}
}

func TestWebBusinessExtractors(t *testing.T) {
	// Test business name extraction
	content := "Acme Corporation is a leading technology company. Tech Solutions Inc provides software services."

	businessNames := extractWebBusinessNames(content)

	if len(businessNames) == 0 {
		t.Error("Expected business names to be extracted")
	}

	// Test address extraction
	addressContent := "Our office is located at 123 Main Street, San Francisco, CA 94102."

	addresses := extractWebAddresses(addressContent)

	if len(addresses) == 0 {
		t.Error("Expected addresses to be extracted")
	}

	// Test contact extraction
	contactContent := "Contact us at info@acme.com or call (555) 123-4567 for more information."

	contacts := extractWebContactInfo(contactContent)

	if len(contacts) == 0 {
		t.Error("Expected contact information to be extracted")
	}

	// Test industry extraction
	industryContent := "We are a technology company specializing in software development and IT services."

	industries := extractWebIndustryInfo(industryContent)

	if len(industries) == 0 {
		t.Error("Expected industry information to be extracted")
	}
}

func TestWebSearchResultAnalyzerAnalyzers(t *testing.T) {
	// Test relevance analyzer
	result := &WebSearchResult{
		Title:          "Test Result",
		Description:    "This is a test result",
		RelevanceScore: 0.7,
	}

	relevanceScore := analyzeWebRelevance(result)

	if relevanceScore != 0.7 {
		t.Errorf("Expected relevance score 0.7, got: %f", relevanceScore)
	}

	// Test content quality analyzer
	shortContentResult := &WebSearchResult{
		Title:       "Short",
		Description: "Brief",
	}

	longContentResult := &WebSearchResult{
		Title:       "Long Title with Multiple Words",
		Description: "This is a much longer description with many more words and detailed information about the content quality and relevance to the search query.",
	}

	shortScore := analyzeWebContentQuality(shortContentResult)
	longScore := analyzeWebContentQuality(longContentResult)

	if shortScore >= longScore {
		t.Error("Expected longer content to have higher quality score")
	}
}

func TestRemoveWebDuplicates(t *testing.T) {
	// Test duplicate removal
	slice := []string{"apple", "banana", "apple", "cherry", "banana", "date"}

	result := removeWebDuplicates(slice)

	if len(result) != 4 {
		t.Errorf("Expected 4 unique items, got: %d", len(result))
	}

	// Check that all items are unique
	seen := make(map[string]bool)
	for _, item := range result {
		if seen[item] {
			t.Errorf("Duplicate item found: %s", item)
		}
		seen[item] = true
	}

	// Test with no duplicates
	noDuplicates := []string{"apple", "banana", "cherry"}
	result2 := removeWebDuplicates(noDuplicates)

	if len(result2) != 3 {
		t.Errorf("Expected 3 items, got: %d", len(result2))
	}
}

func TestWebSearchIntegrationConcurrent(t *testing.T) {
	wsi := NewWebSearchIntegration()

	// Test concurrent search operations
	ctx := context.Background()
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func() {
			response, err := wsi.Search(ctx, "concurrent test query")
			if err != nil {
				t.Errorf("Search failed: %v", err)
			} else if response == nil {
				t.Error("Expected non-nil response")
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

func BenchmarkWebSearchQueryOptimization(b *testing.B) {
	wsqo := NewWebSearchQueryOptimizer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wsqo.OptimizeQuery("benchmark test query")
	}
}

func BenchmarkWebSearchResultRanking(b *testing.B) {
	wsre := NewWebSearchRankingEngine()

	results := []WebSearchResult{
		{
			Title:          "Test Result 1",
			URL:            "https://example.com/1",
			Description:    "This is a test result",
			RelevanceScore: 0.5,
			Rank:           1,
			Source:         "test",
		},
		{
			Title:          "Test Result 2",
			URL:            "https://example.com/2",
			Description:    "This is another test result",
			RelevanceScore: 0.8,
			Rank:           2,
			Source:         "test",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wsre.RankResults(results, "benchmark query")
	}
}

func BenchmarkWebBusinessExtraction(b *testing.B) {
	wbee := NewWebBusinessExtractionEngine()

	results := []WebSearchResult{
		{
			Title:       "Acme Corporation - Leading Technology Solutions",
			URL:         "https://acme.com",
			Description: "Acme Corporation is a technology company based in San Francisco, CA. Contact us at info@acme.com or call (555) 123-4567.",
			Source:      "test",
		},
		{
			Title:       "Tech Solutions Inc - Software Development",
			URL:         "https://techsolutions.com",
			Description: "Tech Solutions Inc provides software development services. Located at 123 Main St, New York, NY 10001.",
			Source:      "test",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wbee.ExtractBusinessInfo(results)
	}
}

func TestWebSearchIntegration_QuotaManagement(t *testing.T) {
	integration := NewWebSearchIntegration()

	// Test initial quota status
	quotaStatus := integration.GetQuotaStatus()
	if quotaStatus == nil {
		t.Fatal("Expected quota status, got nil")
	}

	// Check that global quota info exists
	global, exists := quotaStatus["global"]
	if !exists {
		t.Error("Expected global quota info")
	}

	globalMap, ok := global.(map[string]interface{})
	if !ok {
		t.Error("Expected global quota info to be a map")
	}

	if globalMap["total_daily_quota_limit"] == nil {
		t.Error("Expected total daily quota limit")
	}

	// Check that engines exist
	engines, exists := quotaStatus["engines"]
	if !exists {
		t.Error("Expected engines quota info")
	}

	enginesMap, ok := engines.(map[string]interface{})
	if !ok {
		t.Error("Expected engines quota info to be a map")
	}

	// Check for default engines
	expectedEngines := []string{"google", "bing", "duckduckgo"}
	for _, engineName := range expectedEngines {
		if _, exists := enginesMap[engineName]; !exists {
			t.Errorf("Expected engine %s in quota status", engineName)
		}
	}
}

func TestWebSearchIntegration_GetAvailableEngines(t *testing.T) {
	integration := NewWebSearchIntegration()

	availableEngines := integration.GetAvailableEngines()
	if len(availableEngines) == 0 {
		t.Error("Expected available engines, got empty list")
	}

	// Check that all default engines are available initially
	expectedEngines := []string{"google", "bing", "duckduckgo"}
	for _, engineName := range expectedEngines {
		found := false
		for _, available := range availableEngines {
			if available == engineName {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected engine %s to be available", engineName)
		}
	}
}

func TestWebSearchIntegration_GetFallbackEngine(t *testing.T) {
	integration := NewWebSearchIntegration()

	// Test fallback for Google
	fallback := integration.GetFallbackEngine("google")
	if fallback == "" {
		t.Error("Expected fallback engine for Google")
	}

	// Test fallback for Bing
	fallback = integration.GetFallbackEngine("bing")
	if fallback == "" {
		t.Error("Expected fallback engine for Bing")
	}

	// Test fallback for DuckDuckGo (should be empty as it has no fallbacks)
	fallback = integration.GetFallbackEngine("duckduckgo")
	if fallback != "" {
		t.Error("Expected no fallback engine for DuckDuckGo")
	}

	// Test fallback for nonexistent engine
	fallback = integration.GetFallbackEngine("nonexistent")
	if fallback != "" {
		t.Error("Expected no fallback engine for nonexistent engine")
	}
}

func TestWebSearchIntegration_QuotaConfig(t *testing.T) {
	integration := NewWebSearchIntegration()

	// Test getting current config
	config := integration.GetQuotaConfig()
	if !config.EnableQuotaManagement {
		t.Error("Expected quota management to be enabled by default")
	}

	// Test updating config
	newConfig := QuotaManagerConfig{
		EnableQuotaManagement: true,
		EnableRateLimiting:    true,
		EnableQuotaTracking:   true,
		EnableQuotaAlerts:     true,
		AlertThreshold:        0.9, // 90%
		MaxConcurrentRequests: 20,
		RequestTimeout:        time.Second * 60,
		RetryDelay:            time.Second * 2,
		MaxRetries:            5,
	}

	integration.UpdateQuotaConfig(newConfig)

	// Verify config was updated
	updatedConfig := integration.GetQuotaConfig()
	if updatedConfig.AlertThreshold != 0.9 {
		t.Errorf("Expected AlertThreshold 0.9, got %f", updatedConfig.AlertThreshold)
	}

	if updatedConfig.MaxConcurrentRequests != 20 {
		t.Errorf("Expected MaxConcurrentRequests 20, got %d", updatedConfig.MaxConcurrentRequests)
	}
}

func TestWebSearchIntegration_ResetQuotas(t *testing.T) {
	integration := NewWebSearchIntegration()

	// Reset quotas
	integration.ResetQuotas()

	// Get updated quota status
	updatedStatus := integration.GetQuotaStatus()

	// Check that quotas were reset
	engines, exists := updatedStatus["engines"]
	if !exists {
		t.Fatal("Expected engines in quota status")
	}

	enginesMap, ok := engines.(map[string]interface{})
	if !ok {
		t.Fatal("Expected engines to be a map")
	}

	for engineName, engineData := range enginesMap {
		engineMap, ok := engineData.(map[string]interface{})
		if !ok {
			t.Errorf("Expected engine data for %s to be a map", engineName)
			continue
		}

		if engineMap["daily_quota_used"] != 0 {
			t.Errorf("Expected daily quota used to be 0 for %s, got %v", engineName, engineMap["daily_quota_used"])
		}

		if engineMap["hourly_quota_used"] != 0 {
			t.Errorf("Expected hourly quota used to be 0 for %s, got %v", engineName, engineMap["hourly_quota_used"])
		}

		if engineMap["minute_quota_used"] != 0 {
			t.Errorf("Expected minute quota used to be 0 for %s, got %v", engineName, engineMap["minute_quota_used"])
		}
	}
}

func TestWebSearchIntegration_AddRemoveSearchEngine(t *testing.T) {
	integration := NewWebSearchIntegration()

	// Create a mock search engine
	mockEngine := &MockWebSearchEngine{
		name:      "test-engine",
		rateLimit: time.Millisecond * 100,
	}

	// Create quota info for the new engine
	quotaInfo := &EngineQuotaInfo{
		EngineName:            "test-engine",
		DailyQuotaLimit:       1000,
		HourlyQuotaLimit:      100,
		MinuteQuotaLimit:      10,
		MaxConcurrentRequests: 5,
		IsEnabled:             true,
		Priority:              4,
		FallbackEngines:       []string{"google"},
	}

	// Add the search engine
	err := integration.AddSearchEngine("test-engine", mockEngine, quotaInfo)
	if err != nil {
		t.Fatalf("Failed to add search engine: %v", err)
	}

	// Verify it was added to available engines
	availableEngines := integration.GetAvailableEngines()
	found := false
	for _, engine := range availableEngines {
		if engine == "test-engine" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected test-engine to be in available engines")
	}

	// Verify it appears in quota status
	quotaStatus := integration.GetQuotaStatus()
	engines, exists := quotaStatus["engines"]
	if !exists {
		t.Fatal("Expected engines in quota status")
	}

	enginesMap, ok := engines.(map[string]interface{})
	if !ok {
		t.Fatal("Expected engines to be a map")
	}

	if _, exists := enginesMap["test-engine"]; !exists {
		t.Error("Expected test-engine in quota status")
	}

	// Remove the search engine
	err = integration.RemoveSearchEngine("test-engine")
	if err != nil {
		t.Fatalf("Failed to remove search engine: %v", err)
	}

	// Verify it was removed from available engines
	availableEngines = integration.GetAvailableEngines()
	found = false
	for _, engine := range availableEngines {
		if engine == "test-engine" {
			found = true
			break
		}
	}
	if found {
		t.Error("Expected test-engine to be removed from available engines")
	}
}

// MockWebSearchEngine is a mock implementation for testing
type MockWebSearchEngine struct {
	name      string
	rateLimit time.Duration
}

func (m *MockWebSearchEngine) Search(ctx context.Context, query string, maxResults int) (*WebSearchResponse, error) {
	return &WebSearchResponse{
		EngineName:   m.name,
		Query:        query,
		Results:      []WebSearchResult{},
		TotalResults: 0,
		SearchTime:   time.Millisecond * 50,
	}, nil
}

func (m *MockWebSearchEngine) GetName() string {
	return m.name
}

func (m *MockWebSearchEngine) GetRateLimit() time.Duration {
	return m.rateLimit
}
