package webanalysis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// GoogleCustomSearchService provides comprehensive Google Custom Search API integration
type GoogleCustomSearchService struct {
	apiKey         string
	searchEngineID string
	httpClient     *http.Client
	quotaManager   *GoogleSearchQuotaManager
	config         GoogleCustomSearchConfig
	mu             sync.RWMutex
}

// GoogleCustomSearchConfig holds configuration for Google Custom Search service
type GoogleCustomSearchConfig struct {
	MaxResultsPerQuery    int           `json:"max_results_per_query"`
	MaxQueriesPerDay      int           `json:"max_queries_per_day"`
	MaxQueriesPerSecond   int           `json:"max_queries_per_second"`
	RequestTimeout        time.Duration `json:"request_timeout"`
	RetryAttempts         int           `json:"retry_attempts"`
	RetryDelay            time.Duration `json:"retry_delay"`
	EnableSafeSearch      bool          `json:"enable_safe_search"`
	EnableImageSearch     bool          `json:"enable_image_search"`
	EnableSiteRestriction bool          `json:"enable_site_restriction"`
	SiteRestriction       string        `json:"site_restriction"`
}

// GoogleSearchQuotaManager manages API quotas and rate limiting
type GoogleSearchQuotaManager struct {
	dailyQuotaUsed    int
	lastQuotaReset    time.Time
	queriesThisSecond int
	lastSecondReset   time.Time
	mu                sync.RWMutex
}

// GoogleCustomSearchRequest represents a search request
type GoogleCustomSearchRequest struct {
	Query            string            `json:"query"`
	MaxResults       int               `json:"max_results"`
	StartIndex       int               `json:"start_index"`
	Language         string            `json:"language"`
	Country          string            `json:"country"`
	SafeSearch       string            `json:"safe_search"`
	SiteRestriction  string            `json:"site_restriction"`
	DateRestrict     string            `json:"date_restrict"`
	FileType         string            `json:"file_type"`
	Rights           string            `json:"rights"`
	SearchType       string            `json:"search_type"`
	Sort             string            `json:"sort"`
	AdditionalParams map[string]string `json:"additional_params"`
}

// GoogleCustomSearchResponse represents the Google search response
type GoogleCustomSearchResponse struct {
	Items             []GoogleSearchItem      `json:"items"`
	SearchInformation GoogleSearchInformation `json:"searchInformation"`
	Queries           GoogleSearchQueries     `json:"queries"`
	Context           GoogleSearchContext     `json:"context"`
}

// GoogleSearchItem represents a single search result
type GoogleSearchItem struct {
	Title        string                 `json:"title"`
	Link         string                 `json:"link"`
	Snippet      string                 `json:"snippet"`
	DisplayLink  string                 `json:"displayLink"`
	FormattedURL string                 `json:"formattedUrl"`
	HTMLSnippet  string                 `json:"htmlSnippet"`
	Pagemap      map[string]interface{} `json:"pagemap"`
}

// GoogleSearchInformation contains search metadata
type GoogleSearchInformation struct {
	SearchTime            float64 `json:"searchTime"`
	FormattedSearchTime   string  `json:"formattedSearchTime"`
	TotalResults          string  `json:"totalResults"`
	FormattedTotalResults string  `json:"formattedTotalResults"`
}

// GoogleSearchQueries contains query information
type GoogleSearchQueries struct {
	Request  []GoogleSearchQuery `json:"request"`
	NextPage []GoogleSearchQuery `json:"nextPage"`
}

// GoogleSearchQuery represents a search query
type GoogleSearchQuery struct {
	Title          string `json:"title"`
	TotalResults   string `json:"totalResults"`
	SearchTerms    string `json:"searchTerms"`
	Count          int    `json:"count"`
	StartIndex     int    `json:"startIndex"`
	InputEncoding  string `json:"inputEncoding"`
	OutputEncoding string `json:"outputEncoding"`
	Safe           string `json:"safe"`
	CX             string `json:"cx"`
}

// GoogleSearchContext contains search context
type GoogleSearchContext struct {
	Title string `json:"title"`
}

// GoogleSearchError represents a Google search error
type GoogleSearchError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// NewGoogleCustomSearchService creates a new Google Custom Search service
func NewGoogleCustomSearchService(apiKey, searchEngineID string) *GoogleCustomSearchService {
	config := GoogleCustomSearchConfig{
		MaxResultsPerQuery:    10,
		MaxQueriesPerDay:      10000, // Google's free tier limit
		MaxQueriesPerSecond:   10,
		RequestTimeout:        time.Second * 30,
		RetryAttempts:         3,
		RetryDelay:            time.Second * 1,
		EnableSafeSearch:      true,
		EnableImageSearch:     false,
		EnableSiteRestriction: false,
	}

	return &GoogleCustomSearchService{
		apiKey:         apiKey,
		searchEngineID: searchEngineID,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
		quotaManager: NewGoogleSearchQuotaManager(),
		config:       config,
	}
}

// NewGoogleSearchQuotaManager creates a new quota manager
func NewGoogleSearchQuotaManager() *GoogleSearchQuotaManager {
	return &GoogleSearchQuotaManager{
		lastQuotaReset:  time.Now(),
		lastSecondReset: time.Now(),
	}
}

// Search performs a Google Custom Search
func (gcss *GoogleCustomSearchService) Search(ctx context.Context, req *GoogleCustomSearchRequest) (*GoogleCustomSearchResponse, error) {
	// Check quota before making request
	if err := gcss.quotaManager.CheckQuota(); err != nil {
		return nil, fmt.Errorf("quota exceeded: %w", err)
	}

	// Build search URL
	searchURL, err := gcss.buildSearchURL(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build search URL: %w", err)
	}

	// Make request with retries
	var response *GoogleCustomSearchResponse
	var lastErr error

	for attempt := 0; attempt <= gcss.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(gcss.config.RetryDelay)
		}

		response, lastErr = gcss.makeSearchRequest(ctx, searchURL)
		if lastErr == nil {
			break
		}

		// Check if error is retryable
		if !gcss.isRetryableError(lastErr) {
			break
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("search failed after %d attempts: %w", gcss.config.RetryAttempts+1, lastErr)
	}

	// Update quota usage
	gcss.quotaManager.IncrementUsage()

	return response, nil
}

// buildSearchURL builds the Google Custom Search API URL
func (gcss *GoogleCustomSearchService) buildSearchURL(req *GoogleCustomSearchRequest) (string, error) {
	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}

	// Required parameters
	params.Set("key", gcss.apiKey)
	params.Set("cx", gcss.searchEngineID)
	params.Set("q", req.Query)

	// Optional parameters
	if req.MaxResults > 0 {
		params.Set("num", strconv.Itoa(req.MaxResults))
	} else {
		params.Set("num", strconv.Itoa(gcss.config.MaxResultsPerQuery))
	}

	if req.StartIndex > 0 {
		params.Set("start", strconv.Itoa(req.StartIndex))
	}

	if req.Language != "" {
		params.Set("lr", req.Language)
	}

	if req.Country != "" {
		params.Set("cr", req.Country)
	}

	if gcss.config.EnableSafeSearch {
		params.Set("safe", "active")
	} else if req.SafeSearch != "" {
		params.Set("safe", req.SafeSearch)
	}

	if gcss.config.EnableSiteRestriction && gcss.config.SiteRestriction != "" {
		params.Set("siteSearch", gcss.config.SiteRestriction)
	} else if req.SiteRestriction != "" {
		params.Set("siteSearch", req.SiteRestriction)
	}

	if req.DateRestrict != "" {
		params.Set("dateRestrict", req.DateRestrict)
	}

	if req.FileType != "" {
		params.Set("fileType", req.FileType)
	}

	if req.Rights != "" {
		params.Set("rights", req.Rights)
	}

	if req.SearchType != "" {
		params.Set("searchType", req.SearchType)
	}

	if req.Sort != "" {
		params.Set("sort", req.Sort)
	}

	// Add additional parameters
	for key, value := range req.AdditionalParams {
		params.Set(key, value)
	}

	searchURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	return searchURL, nil
}

// makeSearchRequest makes the actual HTTP request to Google
func (gcss *GoogleCustomSearchService) makeSearchRequest(ctx context.Context, searchURL string) (*GoogleCustomSearchResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("User-Agent", "KYB-Platform/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := gcss.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		// Try to parse as Google error first
		var googleError struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &googleError); err == nil && googleError.Error.Message != "" {
			return nil, fmt.Errorf("Google API error: %s (code: %d)", googleError.Error.Message, googleError.Error.Code)
		}

		// Fallback to generic error
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response GoogleCustomSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// isRetryableError checks if an error is retryable
func (gcss *GoogleCustomSearchService) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Retry on rate limiting and temporary errors
	retryableErrors := []string{
		"rate limit",
		"quota exceeded",
		"temporary",
		"timeout",
		"connection refused",
		"network error",
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryableErr) {
			return true
		}
	}

	return false
}

// CheckQuota checks if quota allows making a request
func (qsm *GoogleSearchQuotaManager) CheckQuota() error {
	qsm.mu.Lock()
	defer qsm.mu.Unlock()

	now := time.Now()

	// Reset daily quota if it's a new day
	if now.Sub(qsm.lastQuotaReset) >= 24*time.Hour {
		qsm.dailyQuotaUsed = 0
		qsm.lastQuotaReset = now
	}

	// Reset per-second quota if it's a new second
	if now.Sub(qsm.lastSecondReset) >= time.Second {
		qsm.queriesThisSecond = 0
		qsm.lastSecondReset = now
	}

	// Check daily quota
	if qsm.dailyQuotaUsed >= 10000 { // Google's free tier limit
		return fmt.Errorf("daily quota exceeded")
	}

	// Check per-second quota
	if qsm.queriesThisSecond >= 10 {
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

// IncrementUsage increments the quota usage counters
func (qsm *GoogleSearchQuotaManager) IncrementUsage() {
	qsm.mu.Lock()
	defer qsm.mu.Unlock()

	qsm.dailyQuotaUsed++
	qsm.queriesThisSecond++
}

// GetQuotaStatus returns current quota status
func (qsm *GoogleSearchQuotaManager) GetQuotaStatus() map[string]interface{} {
	qsm.mu.RLock()
	defer qsm.mu.RUnlock()

	return map[string]interface{}{
		"daily_quota_used":      qsm.dailyQuotaUsed,
		"daily_quota_limit":     10000,
		"queries_this_second":   qsm.queriesThisSecond,
		"per_second_limit":      10,
		"last_quota_reset":      qsm.lastQuotaReset,
		"last_second_reset":     qsm.lastSecondReset,
		"daily_quota_remaining": 10000 - qsm.dailyQuotaUsed,
	}
}

// GetConfig returns the current configuration
func (gcss *GoogleCustomSearchService) GetConfig() GoogleCustomSearchConfig {
	gcss.mu.RLock()
	defer gcss.mu.RUnlock()
	return gcss.config
}

// UpdateConfig updates the configuration
func (gcss *GoogleCustomSearchService) UpdateConfig(config GoogleCustomSearchConfig) {
	gcss.mu.Lock()
	defer gcss.mu.Unlock()
	gcss.config = config
	gcss.httpClient.Timeout = config.RequestTimeout
}

// GetQuotaStatus returns quota status
func (gcss *GoogleCustomSearchService) GetQuotaStatus() map[string]interface{} {
	return gcss.quotaManager.GetQuotaStatus()
}

// ValidateAPIKey validates the Google API key
func (gcss *GoogleCustomSearchService) ValidateAPIKey(ctx context.Context) error {
	testReq := &GoogleCustomSearchRequest{
		Query:      "test",
		MaxResults: 1,
	}

	_, err := gcss.Search(ctx, testReq)
	if err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}

	return nil
}
