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

// BingSearchService provides comprehensive Bing Search API integration
type BingSearchService struct {
	apiKey       string
	httpClient   *http.Client
	quotaManager *BingSearchQuotaManager
	config       BingCustomSearchConfig
	mu           sync.RWMutex
}

// BingCustomSearchConfig holds configuration for Bing Custom Search service
type BingCustomSearchConfig struct {
	MaxResultsPerQuery  int           `json:"max_results_per_query"`
	MaxQueriesPerDay    int           `json:"max_queries_per_day"`
	MaxQueriesPerSecond int           `json:"max_queries_per_second"`
	RequestTimeout      time.Duration `json:"request_timeout"`
	RetryAttempts       int           `json:"retry_attempts"`
	RetryDelay          time.Duration `json:"retry_delay"`
	EnableSafeSearch    bool          `json:"enable_safe_search"`
	EnableImageSearch   bool          `json:"enable_image_search"`
	EnableNewsSearch    bool          `json:"enable_news_search"`
	EnableVideoSearch   bool          `json:"enable_video_search"`
	EnableSpellCheck    bool          `json:"enable_spell_check"`
	EnableSuggestions   bool          `json:"enable_suggestions"`
	Market              string        `json:"market"`
	ResponseFilter      string        `json:"response_filter"`
	AnswerCount         int           `json:"answer_count"`
	Count               int           `json:"count"`
	Offset              int           `json:"offset"`
	Promote             string        `json:"promote"`
	ResponseFields      string        `json:"response_fields"`
	SafeSearch          string        `json:"safe_search"`
	SetLang             string        `json:"set_lang"`
	TextDecorations     bool          `json:"text_decorations"`
	TextFormat          string        `json:"text_format"`
}

// BingSearchQuotaManager manages API quotas and rate limiting
type BingSearchQuotaManager struct {
	dailyQuotaUsed    int
	lastQuotaReset    time.Time
	queriesThisSecond int
	lastSecondReset   time.Time
	mu                sync.RWMutex
}

// BingSearchRequest represents a search request
type BingSearchRequest struct {
	Query            string            `json:"query"`
	MaxResults       int               `json:"max_results"`
	Offset           int               `json:"offset"`
	Market           string            `json:"market"`
	ResponseFilter   string            `json:"response_filter"`
	AnswerCount      int               `json:"answer_count"`
	Count            int               `json:"count"`
	Promote          string            `json:"promote"`
	ResponseFields   string            `json:"response_fields"`
	SafeSearch       string            `json:"safe_search"`
	SetLang          string            `json:"set_lang"`
	TextDecorations  bool              `json:"text_decorations"`
	TextFormat       string            `json:"text_format"`
	AdditionalParams map[string]string `json:"additional_params"`
}

// BingSearchResponse represents the Bing search response
type BingSearchResponse struct {
	Type              string                 `json:"_type"`
	QueryContext      BingQueryContext       `json:"queryContext"`
	WebPages          BingWebPages           `json:"webPages"`
	Entities          BingEntities           `json:"entities"`
	Places            BingPlaces             `json:"places"`
	Videos            BingVideos             `json:"videos"`
	News              BingNews               `json:"news"`
	Images            BingImages             `json:"images"`
	SpellSuggestions  BingSpellSuggestions   `json:"spellSuggestions"`
	Computation       BingComputation        `json:"computation"`
	TimeZone          BingTimeZone           `json:"timeZone"`
	RankingResponse   BingRankingResponse    `json:"rankingResponse"`
	SearchSuggestions []BingSearchSuggestion `json:"searchSuggestions"`
}

// BingQueryContext contains query context information
type BingQueryContext struct {
	OriginalQuery           string `json:"originalQuery"`
	AlteredQuery            string `json:"alteredQuery"`
	AlterationOverrideQuery string `json:"alterationOverrideQuery"`
	AdultIntent             bool   `json:"adultIntent"`
	AskUserForLocation      bool   `json:"askUserForLocation"`
	IsTransactional         bool   `json:"isTransactional"`
}

// BingWebPages contains web search results
type BingWebPages struct {
	Type                  string        `json:"_type"`
	WebSearchURL          string        `json:"webSearchUrl"`
	TotalEstimatedMatches int           `json:"totalEstimatedMatches"`
	Value                 []BingWebPage `json:"value"`
	SomeResultsRemoved    bool          `json:"someResultsRemoved"`
}

// BingWebPage represents a single web search result
type BingWebPage struct {
	ID               string          `json:"id"`
	Name             string          `json:"name"`
	URL              string          `json:"url"`
	DisplayURL       string          `json:"displayUrl"`
	Snippet          string          `json:"snippet"`
	DeepLinks        []BingDeepLink  `json:"deepLinks"`
	DateLastCrawled  string          `json:"dateLastCrawled"`
	CachedPageURL    string          `json:"cachedPageUrl"`
	Language         string          `json:"language"`
	IsFamilyFriendly bool            `json:"isFamilyFriendly"`
	IsNavigational   bool            `json:"isNavigational"`
	SearchTags       []BingSearchTag `json:"searchTags"`
}

// BingDeepLink represents a deep link
type BingDeepLink struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// BingSearchTag represents a search tag
type BingSearchTag struct {
	Name string `json:"name"`
}

// BingEntities contains entity information
type BingEntities struct {
	Type  string       `json:"_type"`
	Value []BingEntity `json:"value"`
}

// BingEntity represents an entity
type BingEntity struct {
	ID                     string                     `json:"id"`
	Name                   string                     `json:"name"`
	Description            string                     `json:"description"`
	URL                    string                     `json:"url"`
	Image                  BingImage                  `json:"image"`
	ContractualRules       []BingContractualRule      `json:"contractualRules"`
	WebSearchURL           string                     `json:"webSearchUrl"`
	EntityPresentationInfo BingEntityPresentationInfo `json:"entityPresentationInfo"`
}

// BingImage represents an image
type BingImage struct {
	Name         string         `json:"name"`
	ThumbnailURL string         `json:"thumbnailUrl"`
	Provider     []BingProvider `json:"provider"`
	HostPageURL  string         `json:"hostPageUrl"`
	Width        int            `json:"width"`
	Height       int            `json:"height"`
}

// BingProvider represents a provider
type BingProvider struct {
	Type string `json:"_type"`
	Name string `json:"name"`
}

// BingContractualRule represents a contractual rule
type BingContractualRule struct {
	Type                 string `json:"_type"`
	TargetPropertyName   string `json:"targetPropertyName"`
	MustBeCloseToContent bool   `json:"mustBeCloseToContent"`
}

// BingEntityPresentationInfo represents entity presentation info
type BingEntityPresentationInfo struct {
	EntityScenario        string   `json:"entityScenario"`
	EntityTypeHints       []string `json:"entityTypeHints"`
	EntityTypeDisplayHint string   `json:"entityTypeDisplayHint"`
}

// BingPlaces contains places information
type BingPlaces struct {
	Type  string      `json:"_type"`
	Value []BingPlace `json:"value"`
}

// BingPlace represents a place
type BingPlace struct {
	ID                     string                     `json:"id"`
	Name                   string                     `json:"name"`
	Description            string                     `json:"description"`
	URL                    string                     `json:"url"`
	EntityPresentationInfo BingEntityPresentationInfo `json:"entityPresentationInfo"`
}

// BingVideos contains video information
type BingVideos struct {
	Type  string      `json:"_type"`
	Value []BingVideo `json:"value"`
}

// BingVideo represents a video
type BingVideo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnailUrl"`
	Duration     string `json:"duration"`
	ViewCount    int    `json:"viewCount"`
}

// BingNews contains news information
type BingNews struct {
	Type  string         `json:"_type"`
	Value []BingNewsItem `json:"value"`
}

// BingNewsItem represents a news item
type BingNewsItem struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	URL           string         `json:"url"`
	DatePublished string         `json:"datePublished"`
	Category      string         `json:"category"`
	Provider      []BingProvider `json:"provider"`
}

// BingImages contains image information
type BingImages struct {
	Type  string      `json:"_type"`
	Value []BingImage `json:"value"`
}

// BingSpellSuggestions contains spell suggestions
type BingSpellSuggestions struct {
	Type  string                `json:"_type"`
	Value []BingSpellSuggestion `json:"value"`
}

// BingSpellSuggestion represents a spell suggestion
type BingSpellSuggestion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// BingComputation represents computation information
type BingComputation struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Expression  string `json:"expression"`
	Value       string `json:"value"`
}

// BingTimeZone represents timezone information
type BingTimeZone struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	URL             string              `json:"url"`
	PrimaryCityTime BingPrimaryCityTime `json:"primaryCityTime"`
}

// BingPrimaryCityTime represents primary city time
type BingPrimaryCityTime struct {
	Location  string `json:"location"`
	Time      string `json:"time"`
	UTCOffset string `json:"utcOffset"`
}

// BingRankingResponse represents ranking response
type BingRankingResponse struct {
	Type     string       `json:"_type"`
	Mainline BingMainline `json:"mainline"`
	Sidebar  BingSidebar  `json:"sidebar"`
	Pole     BingPole     `json:"pole"`
}

// BingMainline represents mainline results
type BingMainline struct {
	Type  string            `json:"_type"`
	Items []BingRankingItem `json:"items"`
}

// BingSidebar represents sidebar results
type BingSidebar struct {
	Type  string            `json:"_type"`
	Items []BingRankingItem `json:"items"`
}

// BingPole represents pole results
type BingPole struct {
	Type  string            `json:"_type"`
	Items []BingRankingItem `json:"items"`
}

// BingRankingItem represents a ranking item
type BingRankingItem struct {
	Type        string           `json:"_type"`
	AnswerType  string           `json:"answerType"`
	ResultIndex int              `json:"resultIndex"`
	Value       BingRankingValue `json:"value"`
}

// BingRankingValue represents a ranking value
type BingRankingValue struct {
	ID string `json:"id"`
}

// BingSearchSuggestion represents a search suggestion
type BingSearchSuggestion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	DisplayText string `json:"displayText"`
	Query       string `json:"query"`
	SearchKind  string `json:"searchKind"`
}

// BingSearchError represents a Bing search error
type BingSearchError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// NewBingSearchService creates a new Bing Search service
func NewBingSearchService(apiKey string) *BingSearchService {
	config := BingCustomSearchConfig{
		MaxResultsPerQuery:  10,
		MaxQueriesPerDay:    3000, // Bing's free tier limit
		MaxQueriesPerSecond: 3,
		RequestTimeout:      time.Second * 30,
		RetryAttempts:       3,
		RetryDelay:          time.Second * 1,
		EnableSafeSearch:    true,
		EnableImageSearch:   false,
		EnableNewsSearch:    false,
		EnableVideoSearch:   false,
		EnableSpellCheck:    true,
		EnableSuggestions:   true,
		Market:              "en-US",
		ResponseFilter:      "Webpages",
		AnswerCount:         0,
		Count:               10,
		Offset:              0,
		Promote:             "",
		ResponseFields:      "",
		SafeSearch:          "Moderate",
		SetLang:             "en",
		TextDecorations:     false,
		TextFormat:          "Raw",
	}

	return &BingSearchService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
		quotaManager: NewBingSearchQuotaManager(),
		config:       config,
	}
}

// NewBingSearchQuotaManager creates a new quota manager
func NewBingSearchQuotaManager() *BingSearchQuotaManager {
	return &BingSearchQuotaManager{
		lastQuotaReset:  time.Now(),
		lastSecondReset: time.Now(),
	}
}

// Search performs a Bing Search
func (bss *BingSearchService) Search(ctx context.Context, req *BingSearchRequest) (*BingSearchResponse, error) {
	// Check quota before making request
	if err := bss.quotaManager.CheckQuota(); err != nil {
		return nil, fmt.Errorf("quota exceeded: %w", err)
	}

	// Build search URL
	searchURL, err := bss.buildSearchURL(req)
	if err != nil {
		return nil, fmt.Errorf("failed to build search URL: %w", err)
	}

	// Make request with retries
	var response *BingSearchResponse
	var lastErr error

	for attempt := 0; attempt <= bss.config.RetryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(bss.config.RetryDelay)
		}

		response, lastErr = bss.makeSearchRequest(ctx, searchURL)
		if lastErr == nil {
			break
		}

		// Check if error is retryable
		if !bss.isRetryableError(lastErr) {
			break
		}
	}

	if lastErr != nil {
		return nil, fmt.Errorf("search failed after %d attempts: %w", bss.config.RetryAttempts+1, lastErr)
	}

	// Update quota usage
	bss.quotaManager.IncrementUsage()

	return response, nil
}

// buildSearchURL builds the Bing Search API URL
func (bss *BingSearchService) buildSearchURL(req *BingSearchRequest) (string, error) {
	baseURL := "https://api.bing.microsoft.com/v7.0/search"
	params := url.Values{}

	// Required parameters
	params.Set("q", req.Query)

	// Optional parameters
	if req.MaxResults > 0 {
		params.Set("count", strconv.Itoa(req.MaxResults))
	} else {
		params.Set("count", strconv.Itoa(bss.config.Count))
	}

	if req.Offset > 0 {
		params.Set("offset", strconv.Itoa(req.Offset))
	} else {
		params.Set("offset", strconv.Itoa(bss.config.Offset))
	}

	if req.Market != "" {
		params.Set("mkt", req.Market)
	} else {
		params.Set("mkt", bss.config.Market)
	}

	if req.ResponseFilter != "" {
		params.Set("responseFilter", req.ResponseFilter)
	} else {
		params.Set("responseFilter", bss.config.ResponseFilter)
	}

	if req.AnswerCount > 0 {
		params.Set("answerCount", strconv.Itoa(req.AnswerCount))
	} else {
		params.Set("answerCount", strconv.Itoa(bss.config.AnswerCount))
	}

	if req.Promote != "" {
		params.Set("promote", req.Promote)
	} else if bss.config.Promote != "" {
		params.Set("promote", bss.config.Promote)
	}

	if req.ResponseFields != "" {
		params.Set("responseFields", req.ResponseFields)
	} else if bss.config.ResponseFields != "" {
		params.Set("responseFields", bss.config.ResponseFields)
	}

	if req.SafeSearch != "" {
		params.Set("safeSearch", req.SafeSearch)
	} else {
		params.Set("safeSearch", bss.config.SafeSearch)
	}

	if req.SetLang != "" {
		params.Set("setLang", req.SetLang)
	} else {
		params.Set("setLang", bss.config.SetLang)
	}

	if req.TextFormat != "" {
		params.Set("textFormat", req.TextFormat)
	} else {
		params.Set("textFormat", bss.config.TextFormat)
	}

	// Add additional parameters
	for key, value := range req.AdditionalParams {
		params.Set(key, value)
	}

	searchURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())
	return searchURL, nil
}

// makeSearchRequest makes the actual HTTP request to Bing
func (bss *BingSearchService) makeSearchRequest(ctx context.Context, searchURL string) (*BingSearchResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Set("Ocp-Apim-Subscription-Key", bss.apiKey)
	req.Header.Set("User-Agent", "KYB-Platform/1.0")
	req.Header.Set("Accept", "application/json")

	resp, err := bss.httpClient.Do(req)
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
		// Try to parse as Bing error first
		var bingError struct {
			Error struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Status  string `json:"status"`
			} `json:"error"`
		}

		if err := json.Unmarshal(body, &bingError); err == nil && bingError.Error.Message != "" {
			return nil, fmt.Errorf("Bing API error: %s (code: %d)", bingError.Error.Message, bingError.Error.Code)
		}

		// Fallback to generic error
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response BingSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &response, nil
}

// isRetryableError checks if an error is retryable
func (bss *BingSearchService) isRetryableError(err error) bool {
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
		"429", // Too Many Requests
		"503", // Service Unavailable
	}

	for _, retryableErr := range retryableErrors {
		if strings.Contains(strings.ToLower(errStr), retryableErr) {
			return true
		}
	}

	return false
}

// CheckQuota checks if quota allows making a request
func (bsqm *BingSearchQuotaManager) CheckQuota() error {
	bsqm.mu.Lock()
	defer bsqm.mu.Unlock()

	now := time.Now()

	// Reset daily quota if it's a new day
	if now.Sub(bsqm.lastQuotaReset) >= 24*time.Hour {
		bsqm.dailyQuotaUsed = 0
		bsqm.lastQuotaReset = now
	}

	// Reset per-second quota if it's a new second
	if now.Sub(bsqm.lastSecondReset) >= time.Second {
		bsqm.queriesThisSecond = 0
		bsqm.lastSecondReset = now
	}

	// Check daily quota
	if bsqm.dailyQuotaUsed >= 3000 { // Bing's free tier limit
		return fmt.Errorf("daily quota exceeded")
	}

	// Check per-second quota
	if bsqm.queriesThisSecond >= 3 {
		return fmt.Errorf("rate limit exceeded")
	}

	return nil
}

// IncrementUsage increments the quota usage counters
func (bsqm *BingSearchQuotaManager) IncrementUsage() {
	bsqm.mu.Lock()
	defer bsqm.mu.Unlock()

	bsqm.dailyQuotaUsed++
	bsqm.queriesThisSecond++
}

// GetQuotaStatus returns current quota status
func (bsqm *BingSearchQuotaManager) GetQuotaStatus() map[string]interface{} {
	bsqm.mu.RLock()
	defer bsqm.mu.RUnlock()

	return map[string]interface{}{
		"daily_quota_used":      bsqm.dailyQuotaUsed,
		"daily_quota_limit":     3000,
		"queries_this_second":   bsqm.queriesThisSecond,
		"per_second_limit":      3,
		"last_quota_reset":      bsqm.lastQuotaReset,
		"last_second_reset":     bsqm.lastSecondReset,
		"daily_quota_remaining": 3000 - bsqm.dailyQuotaUsed,
	}
}

// GetConfig returns the current configuration
func (bss *BingSearchService) GetConfig() BingCustomSearchConfig {
	bss.mu.RLock()
	defer bss.mu.RUnlock()
	return bss.config
}

// UpdateConfig updates the configuration
func (bss *BingSearchService) UpdateConfig(config BingCustomSearchConfig) {
	bss.mu.Lock()
	defer bss.mu.Unlock()
	bss.config = config
	bss.httpClient.Timeout = config.RequestTimeout
}

// GetQuotaStatus returns quota status
func (bss *BingSearchService) GetQuotaStatus() map[string]interface{} {
	return bss.quotaManager.GetQuotaStatus()
}

// ValidateAPIKey validates the Bing API key
func (bss *BingSearchService) ValidateAPIKey(ctx context.Context) error {
	testReq := &BingSearchRequest{
		Query:      "test",
		MaxResults: 1,
	}

	_, err := bss.Search(ctx, testReq)
	if err != nil {
		return fmt.Errorf("API key validation failed: %w", err)
	}

	return nil
}
