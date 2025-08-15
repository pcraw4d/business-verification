package webanalysis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// MultiSourceSearchService provides multi-source search integration
type MultiSourceSearchService struct {
	googleSearchClient *GoogleSearchClient
	bingSearchClient   *BingSearchClient
	resultFilter       *MultiSourceSearchResultFilter
	relevanceScorer    *SearchRelevanceScorer
	fallbackManager    *SearchFallbackManager
	cache              *SearchResultCache
	config             SearchIntegrationConfig
}

// SearchIntegrationConfig holds configuration for search integration
type SearchIntegrationConfig struct {
	GoogleAPIKey         string        `json:"google_api_key"`
	GoogleSearchEngineID string        `json:"google_search_engine_id"`
	BingAPIKey           string        `json:"bing_api_key"`
	BingEndpoint         string        `json:"bing_endpoint"`
	MaxResults           int           `json:"max_results"`
	Timeout              time.Duration `json:"timeout"`
	EnableCaching        bool          `json:"enable_caching"`
	CacheTTL             time.Duration `json:"cache_ttl"`
	RetryAttempts        int           `json:"retry_attempts"`
	RetryDelay           time.Duration `json:"retry_delay"`
}

// MultiSourceSearchResult represents a search result from any provider
type MultiSourceSearchResult struct {
	Title              string            `json:"title"`
	URL                string            `json:"url"`
	Snippet            string            `json:"snippet"`
	Provider           string            `json:"provider"`
	Rank               int               `json:"rank"`
	RelevanceScore     float64           `json:"relevance_score"`
	QualityScore       float64           `json:"quality_score"`
	IndustryIndicators []string          `json:"industry_indicators"`
	Metadata           map[string]string `json:"metadata"`
	RetrievedAt        time.Time         `json:"retrieved_at"`
}

// MultiSourceSearchResponse represents the response from search integration
type MultiSourceSearchResponse struct {
	Results      []*MultiSourceSearchResult `json:"results"`
	TotalResults int                        `json:"total_results"`
	Provider     string                     `json:"provider"`
	Query        string                     `json:"query"`
	SearchTime   time.Duration              `json:"search_time"`
	FallbackUsed bool                       `json:"fallback_used"`
	Metadata     map[string]interface{}     `json:"metadata"`
	RetrievedAt  time.Time                  `json:"retrieved_at"`
}

// NewMultiSourceSearchService creates a new multi-source search service
func NewMultiSourceSearchService(config SearchIntegrationConfig) *MultiSourceSearchService {
	return &MultiSourceSearchService{
		googleSearchClient: NewGoogleSearchClient(config.GoogleAPIKey, config.GoogleSearchEngineID),
		bingSearchClient:   NewBingSearchClient(config.BingAPIKey, config.BingEndpoint),
		resultFilter:       NewMultiSourceSearchResultFilter(),
		relevanceScorer:    NewSearchRelevanceScorer(),
		fallbackManager:    NewSearchFallbackManager(),
		cache:              NewSearchResultCache(config.CacheTTL),
		config:             config,
	}
}

// Search performs multi-source search with fallback mechanisms
func (mss *MultiSourceSearchService) Search(ctx context.Context, query string, business string) (*MultiSourceSearchResponse, error) {
	startTime := time.Now()

	// Check cache first
	if mss.config.EnableCaching {
		if cached := mss.cache.Get(query); cached != nil {
			return cached, nil
		}
	}

	// Try Google Search first
	response, err := mss.searchWithGoogle(ctx, query, business)
	if err != nil {
		// Try Bing Search as fallback
		response, err = mss.searchWithBing(ctx, query, business)
		if err != nil {
			// Use fallback manager for additional fallback strategies
			response, err = mss.fallbackManager.ExecuteFallbackSearch(ctx, query, business)
			if err != nil {
				return nil, fmt.Errorf("all search providers failed: %w", err)
			}
		}
		response.FallbackUsed = true
	}

	// Filter and score results
	response.Results = mss.resultFilter.FilterResults(response.Results, business)
	response.Results = mss.relevanceScorer.ScoreResults(response.Results, business)

	// Cache results
	if mss.config.EnableCaching {
		mss.cache.Set(query, response)
	}

	response.SearchTime = time.Since(startTime)
	response.Metadata = map[string]interface{}{
		"query":    query,
		"business": business,
		"cached":   false,
		"filtered": len(response.Results),
	}

	return response, nil
}

// searchWithGoogle performs search using Google Custom Search API
func (mss *MultiSourceSearchService) searchWithGoogle(ctx context.Context, query string, business string) (*MultiSourceSearchResponse, error) {
	results, err := mss.googleSearchClient.Search(ctx, query, mss.config.MaxResults)
	if err != nil {
		return nil, fmt.Errorf("Google search failed: %w", err)
	}

	// Convert Google results to MultiSourceSearchResult format
	var searchResults []*MultiSourceSearchResult
	for i, result := range results {
		searchResult := &MultiSourceSearchResult{
			Title:       result.Title,
			URL:         result.Link,
			Snippet:     result.Snippet,
			Provider:    "google",
			Rank:        i + 1,
			RetrievedAt: time.Now(),
		}
		searchResults = append(searchResults, searchResult)
	}

	return &MultiSourceSearchResponse{
		Results:      searchResults,
		TotalResults: len(searchResults),
		Provider:     "google",
		Query:        query,
	}, nil
}

// searchWithBing performs search using Bing Search API
func (mss *MultiSourceSearchService) searchWithBing(ctx context.Context, query string, business string) (*MultiSourceSearchResponse, error) {
	results, err := mss.bingSearchClient.Search(ctx, query, mss.config.MaxResults)
	if err != nil {
		return nil, fmt.Errorf("Bing search failed: %w", err)
	}

	// Convert Bing results to MultiSourceSearchResult format
	var searchResults []*MultiSourceSearchResult
	for i, result := range results {
		searchResult := &MultiSourceSearchResult{
			Title:       result.Name,
			URL:         result.URL,
			Snippet:     result.Snippet,
			Provider:    "bing",
			Rank:        i + 1,
			RetrievedAt: time.Now(),
		}
		searchResults = append(searchResults, searchResult)
	}

	return &MultiSourceSearchResponse{
		Results:      searchResults,
		TotalResults: len(searchResults),
		Provider:     "bing",
		Query:        query,
	}, nil
}

// GoogleSearchClient handles Google Custom Search API integration
type GoogleSearchClient struct {
	apiKey         string
	searchEngineID string
	httpClient     *http.Client
}

// GoogleSearchResult represents a Google search result
type GoogleSearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// GoogleSearchResponse represents Google search API response
type GoogleSearchResponse struct {
	Items []GoogleSearchResult `json:"items"`
}

// NewGoogleSearchClient creates a new Google search client
func NewGoogleSearchClient(apiKey, searchEngineID string) *GoogleSearchClient {
	return &GoogleSearchClient{
		apiKey:         apiKey,
		searchEngineID: searchEngineID,
		httpClient:     &http.Client{Timeout: 30 * time.Second},
	}
}

// Search performs search using Google Custom Search API
func (gsc *GoogleSearchClient) Search(ctx context.Context, query string, maxResults int) ([]GoogleSearchResult, error) {
	// Build search URL
	searchURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Set("key", gsc.apiKey)
	params.Set("cx", gsc.searchEngineID)
	params.Set("q", query)
	params.Set("num", fmt.Sprintf("%d", maxResults))

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := gsc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Google API error: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var searchResponse GoogleSearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return searchResponse.Items, nil
}

// BingSearchClient handles Bing Search API integration
type BingSearchClient struct {
	apiKey     string
	endpoint   string
	httpClient *http.Client
}

// BingSearchResult represents a Bing search result
type BingSearchResult struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// MultiSourceBingSearchResponse represents Bing search API response
type MultiSourceBingSearchResponse struct {
	WebPages struct {
		Value []BingSearchResult `json:"value"`
	} `json:"webPages"`
}

// NewBingSearchClient creates a new Bing search client
func NewBingSearchClient(apiKey, endpoint string) *BingSearchClient {
	if endpoint == "" {
		endpoint = "https://api.bing.microsoft.com/v7.0/search"
	}

	return &BingSearchClient{
		apiKey:     apiKey,
		endpoint:   endpoint,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// Search performs search using Bing Search API
func (bsc *BingSearchClient) Search(ctx context.Context, query string, maxResults int) ([]BingSearchResult, error) {
	// Build search URL
	searchURL := bsc.endpoint
	params := url.Values{}
	params.Set("q", query)
	params.Set("count", fmt.Sprintf("%d", maxResults))

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add Bing API key header
	req.Header.Set("Ocp-Apim-Subscription-Key", bsc.apiKey)

	// Execute request
	resp, err := bsc.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Bing API error: %s - %s", resp.Status, string(body))
	}

	// Parse response
	var searchResponse MultiSourceBingSearchResponse
	if err := json.Unmarshal(body, &searchResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return searchResponse.WebPages.Value, nil
}

// MultiSourceSearchResultFilter filters and validates search results
type MultiSourceSearchResultFilter struct {
	excludePatterns []string
	includePatterns []string
}

// NewMultiSourceSearchResultFilter creates a new search result filter
func NewMultiSourceSearchResultFilter() *MultiSourceSearchResultFilter {
	return &MultiSourceSearchResultFilter{
		excludePatterns: []string{
			"wikipedia.org",
			"youtube.com",
			"facebook.com",
			"twitter.com",
			"linkedin.com",
			"instagram.com",
		},
		includePatterns: []string{
			".com",
			".org",
			".net",
			".edu",
			".gov",
		},
	}
}

// FilterResults filters search results based on business relevance
func (srf *MultiSourceSearchResultFilter) FilterResults(results []*MultiSourceSearchResult, business string) []*MultiSourceSearchResult {
	var filteredResults []*MultiSourceSearchResult

	for _, result := range results {
		if srf.isRelevant(result, business) {
			filteredResults = append(filteredResults, result)
		}
	}

	return filteredResults
}

// isRelevant checks if a search result is relevant to the business
func (srf *MultiSourceSearchResultFilter) isRelevant(result *MultiSourceSearchResult, business string) bool {
	// Check exclude patterns
	for _, pattern := range srf.excludePatterns {
		if strings.Contains(strings.ToLower(result.URL), pattern) {
			return false
		}
	}

	// Check include patterns
	hasValidDomain := false
	for _, pattern := range srf.includePatterns {
		if strings.Contains(strings.ToLower(result.URL), pattern) {
			hasValidDomain = true
			break
		}
	}

	if !hasValidDomain {
		return false
	}

	// Check business name relevance
	businessLower := strings.ToLower(business)
	titleLower := strings.ToLower(result.Title)
	snippetLower := strings.ToLower(result.Snippet)

	// Check if business name appears in title or snippet
	if strings.Contains(titleLower, businessLower) || strings.Contains(snippetLower, businessLower) {
		return true
	}

	// Check for business-related keywords
	businessKeywords := []string{"company", "business", "corporate", "enterprise", "organization"}
	for _, keyword := range businessKeywords {
		if strings.Contains(titleLower, keyword) || strings.Contains(snippetLower, keyword) {
			return true
		}
	}

	return true // Default to including if no specific exclusions
}

// SearchRelevanceScorer scores search results for relevance
type SearchRelevanceScorer struct {
	keywordWeights map[string]float64
}

// NewSearchRelevanceScorer creates a new search relevance scorer
func NewSearchRelevanceScorer() *SearchRelevanceScorer {
	return &SearchRelevanceScorer{
		keywordWeights: map[string]float64{
			"company":      0.8,
			"business":     0.8,
			"corporate":    0.7,
			"enterprise":   0.7,
			"organization": 0.6,
			"services":     0.6,
			"products":     0.5,
			"about":        0.5,
			"contact":      0.4,
		},
	}
}

// ScoreResults scores search results for relevance
func (srs *SearchRelevanceScorer) ScoreResults(results []*MultiSourceSearchResult, business string) []*MultiSourceSearchResult {
	for _, result := range results {
		result.RelevanceScore = srs.calculateRelevanceScore(result, business)
		result.QualityScore = srs.calculateQualityScore(result)
	}

	return results
}

// calculateRelevanceScore calculates relevance score for a search result
func (srs *SearchRelevanceScorer) calculateRelevanceScore(result *MultiSourceSearchResult, business string) float64 {
	score := 0.0

	// Base score from rank
	score += float64(1.0 / float64(result.Rank))

	// Business name match
	businessLower := strings.ToLower(business)
	titleLower := strings.ToLower(result.Title)
	snippetLower := strings.ToLower(result.Snippet)

	if strings.Contains(titleLower, businessLower) {
		score += 0.4
	}
	if strings.Contains(snippetLower, businessLower) {
		score += 0.2
	}

	// Keyword relevance
	for keyword, weight := range srs.keywordWeights {
		if strings.Contains(titleLower, keyword) {
			score += weight * 0.1
		}
		if strings.Contains(snippetLower, keyword) {
			score += weight * 0.05
		}
	}

	// URL relevance
	urlLower := strings.ToLower(result.URL)
	if strings.Contains(urlLower, strings.ReplaceAll(businessLower, " ", "")) {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// calculateQualityScore calculates quality score for a search result
func (srs *SearchRelevanceScorer) calculateQualityScore(result *MultiSourceSearchResult) float64 {
	score := 0.0

	// Title quality
	if len(result.Title) > 10 && len(result.Title) < 100 {
		score += 0.3
	}

	// Snippet quality
	if len(result.Snippet) > 50 && len(result.Snippet) < 300 {
		score += 0.3
	}

	// URL quality
	if strings.HasPrefix(result.URL, "https://") {
		score += 0.2
	}

	// Provider quality
	if result.Provider == "google" {
		score += 0.1
	} else if result.Provider == "bing" {
		score += 0.05
	}

	return math.Min(score, 1.0)
}

// SearchFallbackManager manages fallback search strategies
type SearchFallbackManager struct {
	fallbackStrategies []FallbackSearchStrategy
}

// FallbackSearchStrategy defines a fallback search strategy
type FallbackSearchStrategy interface {
	Execute(ctx context.Context, query string, business string) (*MultiSourceSearchResponse, error)
	GetName() string
	GetPriority() int
}

// NewSearchFallbackManager creates a new search fallback manager
func NewSearchFallbackManager() *SearchFallbackManager {
	return &SearchFallbackManager{
		fallbackStrategies: []FallbackSearchStrategy{
			&BasicWebSearchStrategy{},
			&BusinessDirectorySearchStrategy{},
		},
	}
}

// ExecuteFallbackSearch executes fallback search strategies
func (sfm *SearchFallbackManager) ExecuteFallbackSearch(ctx context.Context, query string, business string) (*MultiSourceSearchResponse, error) {
	for _, strategy := range sfm.fallbackStrategies {
		response, err := strategy.Execute(ctx, query, business)
		if err == nil && len(response.Results) > 0 {
			return response, nil
		}
	}

	return nil, fmt.Errorf("all fallback strategies failed")
}

// BasicWebSearchStrategy implements basic web search fallback
type BasicWebSearchStrategy struct{}

// Execute implements basic web search
func (bwss *BasicWebSearchStrategy) Execute(ctx context.Context, query string, business string) (*MultiSourceSearchResponse, error) {
	// This would implement a basic web search using a different approach
	// For now, return a minimal response
	return &MultiSourceSearchResponse{
		Results:      []*MultiSourceSearchResult{},
		TotalResults: 0,
		Provider:     "basic_web_search",
		Query:        query,
	}, nil
}

// GetName returns strategy name
func (bwss *BasicWebSearchStrategy) GetName() string {
	return "BasicWebSearch"
}

// GetPriority returns strategy priority
func (bwss *BasicWebSearchStrategy) GetPriority() int {
	return 1
}

// BusinessDirectorySearchStrategy implements business directory search fallback
type BusinessDirectorySearchStrategy struct{}

// Execute implements business directory search
func (bdss *BusinessDirectorySearchStrategy) Execute(ctx context.Context, query string, business string) (*MultiSourceSearchResponse, error) {
	// This would implement business directory search
	// For now, return a minimal response
	return &MultiSourceSearchResponse{
		Results:      []*MultiSourceSearchResult{},
		TotalResults: 0,
		Provider:     "business_directory",
		Query:        query,
	}, nil
}

// GetName returns strategy name
func (bdss *BusinessDirectorySearchStrategy) GetName() string {
	return "BusinessDirectorySearch"
}

// GetPriority returns strategy priority
func (bdss *BusinessDirectorySearchStrategy) GetPriority() int {
	return 2
}

// SearchResultCache provides caching for search results
type SearchResultCache struct {
	cache map[string]*MultiSourceSearchResponse
	ttl   time.Duration
	mu    sync.RWMutex
}

// NewSearchResultCache creates a new search result cache
func NewSearchResultCache(ttl time.Duration) *SearchResultCache {
	cache := &SearchResultCache{
		cache: make(map[string]*MultiSourceSearchResponse),
		ttl:   ttl,
	}

	// Start cache cleanup goroutine
	go cache.cleanup()

	return cache
}

// Get retrieves a cached search response
func (src *SearchResultCache) Get(query string) *MultiSourceSearchResponse {
	src.mu.RLock()
	defer src.mu.RUnlock()

	if response, exists := src.cache[query]; exists {
		// Check if cache entry is still valid
		if time.Since(response.RetrievedAt) < src.ttl {
			return response
		}
		// Remove expired entry
		delete(src.cache, query)
	}

	return nil
}

// Set stores a search response in cache
func (src *SearchResultCache) Set(query string, response *MultiSourceSearchResponse) {
	src.mu.Lock()
	defer src.mu.Unlock()

	response.RetrievedAt = time.Now()
	src.cache[query] = response
}

// cleanup periodically cleans up expired cache entries
func (src *SearchResultCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		src.mu.Lock()
		now := time.Now()
		for query, response := range src.cache {
			if now.Sub(response.RetrievedAt) > src.ttl {
				delete(src.cache, query)
			}
		}
		src.mu.Unlock()
	}
}
