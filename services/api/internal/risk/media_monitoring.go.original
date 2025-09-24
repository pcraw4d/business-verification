package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// MediaProvider represents a media monitoring provider
type MediaProvider interface {
	GetNewsArticles(ctx context.Context, businessID string, query NewsQuery) (*NewsResult, error)
	GetSocialMediaMentions(ctx context.Context, businessID string, query SocialMediaQuery) (*SocialMediaResult, error)
	GetMediaSentiment(ctx context.Context, businessID string) (*SentimentResult, error)
	GetReputationScore(ctx context.Context, businessID string) (*ReputationScore, error)
	GetMediaAlerts(ctx context.Context, businessID string) (*MediaAlerts, error)
	GetProviderName() string
	IsAvailable() bool
}

// NewsQuery represents a news search query
type NewsQuery struct {
	BusinessName    string    `json:"business_name"`
	Keywords        []string  `json:"keywords,omitempty"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Sources         []string  `json:"sources,omitempty"`
	Language        string    `json:"language,omitempty"`
	MaxResults      int       `json:"max_results,omitempty"`
	IncludeNegative bool      `json:"include_negative"`
	IncludePositive bool      `json:"include_positive"`
	IncludeNeutral  bool      `json:"include_neutral"`
}

// NewsResult represents news monitoring results
type NewsResult struct {
	BusinessID       string                 `json:"business_id"`
	Provider         string                 `json:"provider"`
	LastUpdated      time.Time              `json:"last_updated"`
	TotalArticles    int                    `json:"total_articles"`
	PositiveCount    int                    `json:"positive_count"`
	NegativeCount    int                    `json:"negative_count"`
	NeutralCount     int                    `json:"neutral_count"`
	Articles         []NewsArticle          `json:"articles,omitempty"`
	RiskLevel        RiskLevel              `json:"risk_level"`
	OverallSentiment float64                `json:"overall_sentiment"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// NewsArticle represents a news article
type NewsArticle struct {
	ArticleID      string    `json:"article_id"`
	Title          string    `json:"title"`
	Content        string    `json:"content"`
	URL            string    `json:"url"`
	Source         string    `json:"source"`
	Author         string    `json:"author,omitempty"`
	PublishedDate  time.Time `json:"published_date"`
	Sentiment      string    `json:"sentiment"` // "positive", "negative", "neutral"
	SentimentScore float64   `json:"sentiment_score"`
	RelevanceScore float64   `json:"relevance_score"`
	RiskLevel      RiskLevel `json:"risk_level"`
	Keywords       []string  `json:"keywords,omitempty"`
	Summary        string    `json:"summary,omitempty"`
	Language       string    `json:"language"`
	Country        string    `json:"country,omitempty"`
}

// SocialMediaQuery represents a social media search query
type SocialMediaQuery struct {
	BusinessName    string    `json:"business_name"`
	Platforms       []string  `json:"platforms"` // "twitter", "facebook", "linkedin", "instagram"
	Keywords        []string  `json:"keywords,omitempty"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	MaxResults      int       `json:"max_results,omitempty"`
	IncludeNegative bool      `json:"include_negative"`
	IncludePositive bool      `json:"include_positive"`
	IncludeNeutral  bool      `json:"include_neutral"`
}

// SocialMediaResult represents social media monitoring results
type SocialMediaResult struct {
	BusinessID       string                 `json:"business_id"`
	Provider         string                 `json:"provider"`
	LastUpdated      time.Time              `json:"last_updated"`
	TotalMentions    int                    `json:"total_mentions"`
	PositiveCount    int                    `json:"positive_count"`
	NegativeCount    int                    `json:"negative_count"`
	NeutralCount     int                    `json:"neutral_count"`
	Mentions         []SocialMediaMention   `json:"mentions,omitempty"`
	RiskLevel        RiskLevel              `json:"risk_level"`
	OverallSentiment float64                `json:"overall_sentiment"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// SocialMediaMention represents a social media mention
type SocialMediaMention struct {
	MentionID      string    `json:"mention_id"`
	Platform       string    `json:"platform"`
	Username       string    `json:"username"`
	Content        string    `json:"content"`
	URL            string    `json:"url"`
	PublishedDate  time.Time `json:"published_date"`
	Sentiment      string    `json:"sentiment"`
	SentimentScore float64   `json:"sentiment_score"`
	RelevanceScore float64   `json:"relevance_score"`
	RiskLevel      RiskLevel `json:"risk_level"`
	Engagement     int       `json:"engagement"` // likes, shares, comments
	Followers      int       `json:"followers,omitempty"`
	Language       string    `json:"language"`
	Country        string    `json:"country,omitempty"`
}

// SentimentResult represents sentiment analysis results
type SentimentResult struct {
	BusinessID    string                 `json:"business_id"`
	Provider      string                 `json:"provider"`
	LastUpdated   time.Time              `json:"last_updated"`
	OverallScore  float64                `json:"overall_score"`
	PositiveScore float64                `json:"positive_score"`
	NegativeScore float64                `json:"negative_score"`
	NeutralScore  float64                `json:"neutral_score"`
	Confidence    float64                `json:"confidence"`
	RiskLevel     RiskLevel              `json:"risk_level"`
	Trend         string                 `json:"trend"` // "improving", "stable", "declining"
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ReputationScore represents a reputation score
type ReputationScore struct {
	BusinessID   string                 `json:"business_id"`
	Provider     string                 `json:"provider"`
	LastUpdated  time.Time              `json:"last_updated"`
	OverallScore float64                `json:"overall_score"`
	NewsScore    float64                `json:"news_score"`
	SocialScore  float64                `json:"social_score"`
	ReviewScore  float64                `json:"review_score"`
	RiskLevel    RiskLevel              `json:"risk_level"`
	Trend        string                 `json:"trend"`
	Confidence   float64                `json:"confidence"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// MediaAlerts represents media monitoring alerts
type MediaAlerts struct {
	BusinessID     string                 `json:"business_id"`
	Provider       string                 `json:"provider"`
	LastUpdated    time.Time              `json:"last_updated"`
	TotalAlerts    int                    `json:"total_alerts"`
	HighPriority   int                    `json:"high_priority"`
	MediumPriority int                    `json:"medium_priority"`
	LowPriority    int                    `json:"low_priority"`
	Alerts         []MediaAlert           `json:"alerts,omitempty"`
	RiskLevel      RiskLevel              `json:"risk_level"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// MediaAlert represents a media monitoring alert
type MediaAlert struct {
	AlertID      string     `json:"alert_id"`
	AlertType    string     `json:"alert_type"` // "negative_news", "social_media_crisis", "reputation_drop"
	Priority     string     `json:"priority"`   // "high", "medium", "low"
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Source       string     `json:"source"`
	URL          string     `json:"url,omitempty"`
	CreatedDate  time.Time  `json:"created_date"`
	RiskLevel    RiskLevel  `json:"risk_level"`
	Resolved     bool       `json:"resolved"`
	ResolvedDate *time.Time `json:"resolved_date,omitempty"`
	ResolvedBy   string     `json:"resolved_by,omitempty"`
	Notes        string     `json:"notes,omitempty"`
}

// RealMediaProvider represents a real media monitoring provider with API integration
type RealMediaProvider struct {
	name          string
	apiKey        string
	baseURL       string
	timeout       time.Duration
	retryAttempts int
	available     bool
	logger        *observability.Logger
	httpClient    *http.Client
}

// NewRealMediaProvider creates a new real media monitoring provider
func NewRealMediaProvider(name, apiKey, baseURL string, logger *observability.Logger) *RealMediaProvider {
	return &RealMediaProvider{
		name:          name,
		apiKey:        apiKey,
		baseURL:       baseURL,
		timeout:       30 * time.Second,
		retryAttempts: 3,
		available:     true,
		logger:        logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetNewsArticles implements MediaProvider interface for real providers
func (p *RealMediaProvider) GetNewsArticles(ctx context.Context, businessID string, query NewsQuery) (*NewsResult, error) {
	requestID := ctx.Value("request_id").(string)

	p.logger.Info("Requesting news articles from real provider",
		"request_id", requestID,
		"business_id", businessID,
		"provider", p.name,
	)

	url := fmt.Sprintf("%s/news/search", p.baseURL)

	// Create request body
	requestBody := map[string]interface{}{
		"business_name": query.BusinessName,
		"keywords":      query.Keywords,
		"start_date":    query.StartDate.Format(time.RFC3339),
		"end_date":      query.EndDate.Format(time.RFC3339),
		"sources":       query.Sources,
		"language":      query.Language,
		"max_results":   query.MaxResults,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < p.retryAttempts; attempt++ {
		resp, err = p.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if attempt < p.retryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		p.logger.Error("Failed to get news articles from real provider",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.logger.Error("Real provider returned error status for news articles",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"status_code", resp.StatusCode,
		)
		return nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	var newsResult NewsResult
	if err := json.NewDecoder(resp.Body).Decode(&newsResult); err != nil {
		p.logger.Error("Failed to decode news articles response",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	p.logger.Info("Successfully retrieved news articles from real provider",
		"request_id", requestID,
		"business_id", businessID,
		"provider", p.name,
		"total_articles", newsResult.TotalArticles,
	)

	return &newsResult, nil
}

// GetSocialMediaMentions implements MediaProvider interface for real providers
func (p *RealMediaProvider) GetSocialMediaMentions(ctx context.Context, businessID string, query SocialMediaQuery) (*SocialMediaResult, error) {
	requestID := ctx.Value("request_id").(string)

	p.logger.Info("Requesting social media mentions from real provider",
		"request_id", requestID,
		"business_id", businessID,
		"provider", p.name,
	)

	url := fmt.Sprintf("%s/social/search", p.baseURL)

	// Create request body
	requestBody := map[string]interface{}{
		"business_name": query.BusinessName,
		"platforms":     query.Platforms,
		"keywords":      query.Keywords,
		"start_date":    query.StartDate.Format(time.RFC3339),
		"end_date":      query.EndDate.Format(time.RFC3339),
		"max_results":   query.MaxResults,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < p.retryAttempts; attempt++ {
		resp, err = p.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if attempt < p.retryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		p.logger.Error("Failed to get social media mentions from real provider",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.logger.Error("Real provider returned error status for social media mentions",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"status_code", resp.StatusCode,
		)
		return nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	var socialMediaResult SocialMediaResult
	if err := json.NewDecoder(resp.Body).Decode(&socialMediaResult); err != nil {
		p.logger.Error("Failed to decode social media mentions response",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	p.logger.Info("Successfully retrieved social media mentions from real provider",
		"request_id", requestID,
		"business_id", businessID,
		"provider", p.name,
		"total_mentions", socialMediaResult.TotalMentions,
	)

	return &socialMediaResult, nil
}

// GetMediaSentiment implements MediaProvider interface for real providers
func (p *RealMediaProvider) GetMediaSentiment(ctx context.Context, businessID string) (*SentimentResult, error) {
	requestID := ctx.Value("request_id").(string)

	p.logger.Info("Requesting media sentiment from real provider",
		"request_id", requestID,
		"business_id", businessID,
		"provider", p.name,
	)

	url := fmt.Sprintf("%s/sentiment/%s", p.baseURL, businessID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < p.retryAttempts; attempt++ {
		resp, err = p.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if attempt < p.retryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		p.logger.Error("Failed to get media sentiment from real provider",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.logger.Error("Real provider returned error status for media sentiment",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"status_code", resp.StatusCode,
		)
		return nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	var sentimentResult SentimentResult
	if err := json.NewDecoder(resp.Body).Decode(&sentimentResult); err != nil {
		p.logger.Error("Failed to decode media sentiment response",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	p.logger.Info("Successfully retrieved media sentiment from real provider",
		"request_id", requestID,
		"business_id", businessID,
		"provider", p.name,
		"overall_score", sentimentResult.OverallScore,
	)

	return &sentimentResult, nil
}

// GetReputationScore implements MediaProvider interface for real providers
func (p *RealMediaProvider) GetReputationScore(ctx context.Context, businessID string) (*ReputationScore, error) {
	requestID := ctx.Value("request_id").(string)

	p.logger.Info("Requesting reputation score from real provider",
		"request_id", requestID,
		"business_id", businessID,
		"provider", p.name,
	)

	url := fmt.Sprintf("%s/reputation/%s", p.baseURL, businessID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < p.retryAttempts; attempt++ {
		resp, err = p.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if attempt < p.retryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		p.logger.Error("Failed to get reputation score from real provider",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.logger.Error("Real provider returned error status for reputation score",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"status_code", resp.StatusCode,
		)
		return nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	var reputationScore ReputationScore
	if err := json.NewDecoder(resp.Body).Decode(&reputationScore); err != nil {
		p.logger.Error("Failed to decode reputation score response",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	p.logger.Info("Successfully retrieved reputation score from real provider",
		"request_id", requestID,
		"business_id", businessID,
		"provider", p.name,
		"overall_score", reputationScore.OverallScore,
	)

	return &reputationScore, nil
}

// GetMediaAlerts implements MediaProvider interface for real providers
func (p *RealMediaProvider) GetMediaAlerts(ctx context.Context, businessID string) (*MediaAlerts, error) {
	requestID := ctx.Value("request_id").(string)

	p.logger.Info("Requesting media alerts from real provider",
		"request_id", requestID,
		"business_id", businessID,
		"provider", p.name,
	)

	url := fmt.Sprintf("%s/alerts/%s", p.baseURL, businessID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	for attempt := 0; attempt < p.retryAttempts; attempt++ {
		resp, err = p.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}

		if attempt < p.retryAttempts-1 {
			time.Sleep(time.Duration(attempt+1) * time.Second)
		}
	}

	if err != nil {
		p.logger.Error("Failed to get media alerts from real provider",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.logger.Error("Real provider returned error status for media alerts",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"status_code", resp.StatusCode,
		)
		return nil, fmt.Errorf("provider returned status %d", resp.StatusCode)
	}

	var mediaAlerts MediaAlerts
	if err := json.NewDecoder(resp.Body).Decode(&mediaAlerts); err != nil {
		p.logger.Error("Failed to decode media alerts response",
			"request_id", requestID,
			"business_id", businessID,
			"provider", p.name,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	p.logger.Info("Successfully retrieved media alerts from real provider",
		"request_id", requestID,
		"business_id", businessID,
		"provider", p.name,
		"total_alerts", mediaAlerts.TotalAlerts,
	)

	return &mediaAlerts, nil
}

// GetProviderName implements MediaProvider interface for real providers
func (p *RealMediaProvider) GetProviderName() string {
	return p.name
}

// IsAvailable implements MediaProvider interface for real providers
func (p *RealMediaProvider) IsAvailable() bool {
	return p.available
}

// SetAvailable sets the availability status of the provider
func (p *RealMediaProvider) SetAvailable(available bool) {
	p.available = available
}

// NewsProvider represents a news monitoring provider
type NewsProvider struct {
	*RealMediaProvider
}

// NewNewsProvider creates a new news monitoring provider
func NewNewsProvider(apiKey, baseURL string, logger *observability.Logger) *NewsProvider {
	return &NewsProvider{
		RealMediaProvider: NewRealMediaProvider("news_provider", apiKey, baseURL, logger),
	}
}

// SocialMediaProvider represents a social media monitoring provider
type SocialMediaProvider struct {
	*RealMediaProvider
}

// NewSocialMediaProvider creates a new social media monitoring provider
func NewSocialMediaProvider(apiKey, baseURL string, logger *observability.Logger) *SocialMediaProvider {
	return &SocialMediaProvider{
		RealMediaProvider: NewRealMediaProvider("social_media_provider", apiKey, baseURL, logger),
	}
}

// SentimentProvider represents a sentiment analysis provider
type SentimentProvider struct {
	*RealMediaProvider
}

// NewSentimentProvider creates a new sentiment analysis provider
func NewSentimentProvider(apiKey, baseURL string, logger *observability.Logger) *SentimentProvider {
	return &SentimentProvider{
		RealMediaProvider: NewRealMediaProvider("sentiment_provider", apiKey, baseURL, logger),
	}
}

// ReputationProvider represents a reputation monitoring provider
type ReputationProvider struct {
	*RealMediaProvider
}

// NewReputationProvider creates a new reputation monitoring provider
func NewReputationProvider(apiKey, baseURL string, logger *observability.Logger) *ReputationProvider {
	return &ReputationProvider{
		RealMediaProvider: NewRealMediaProvider("reputation_provider", apiKey, baseURL, logger),
	}
}

// AlertProvider represents a media alert provider
type AlertProvider struct {
	*RealMediaProvider
}

// NewAlertProvider creates a new media alert provider
func NewAlertProvider(apiKey, baseURL string, logger *observability.Logger) *AlertProvider {
	return &AlertProvider{
		RealMediaProvider: NewRealMediaProvider("alert_provider", apiKey, baseURL, logger),
	}
}

// MediaProviderManager manages multiple media monitoring providers
type MediaProviderManager struct {
	logger            *observability.Logger
	providers         map[string]MediaProvider
	primaryProvider   string
	fallbackProviders []string
	timeout           time.Duration
	retryAttempts     int
}

// NewMediaProviderManager creates a new media provider manager
func NewMediaProviderManager(logger *observability.Logger) *MediaProviderManager {
	return &MediaProviderManager{
		logger:            logger,
		providers:         make(map[string]MediaProvider),
		primaryProvider:   "media_provider",
		fallbackProviders: []string{"backup_media_provider"},
		timeout:           30 * time.Second,
		retryAttempts:     3,
	}
}

// RegisterProvider registers a media monitoring provider
func (m *MediaProviderManager) RegisterProvider(name string, provider MediaProvider) {
	m.providers[name] = provider
	m.logger.Info("Media provider registered",
		"provider_name", name,
		"available", provider.IsAvailable(),
	)
}

// GetNewsArticles retrieves news articles from available providers
func (m *MediaProviderManager) GetNewsArticles(ctx context.Context, businessID string, query NewsQuery) (*NewsResult, error) {
	requestID := ctx.Value("request_id").(string)

	m.logger.Info("Retrieving news articles",
		"request_id", requestID,
		"business_id", businessID,
		"business_name", query.BusinessName,
	)

	// Try primary provider first
	if provider, exists := m.providers[m.primaryProvider]; exists && provider.IsAvailable() {
		result, err := provider.GetNewsArticles(ctx, businessID, query)
		if err == nil {
			m.logger.Info("Retrieved news articles from primary provider",
				"request_id", requestID,
				"business_id", businessID,
				"provider", m.primaryProvider,
				"total_articles", result.TotalArticles,
			)
			return result, nil
		}
		m.logger.Warn("Primary provider failed, trying fallback providers",
			"request_id", requestID,
			"business_id", businessID,
			"provider", m.primaryProvider,
			"error", err.Error(),
		)
	}

	// Try fallback providers
	for _, providerName := range m.fallbackProviders {
		if provider, exists := m.providers[providerName]; exists && provider.IsAvailable() {
			result, err := provider.GetNewsArticles(ctx, businessID, query)
			if err == nil {
				m.logger.Info("Retrieved news articles from fallback provider",
					"request_id", requestID,
					"business_id", businessID,
					"provider", providerName,
					"total_articles", result.TotalArticles,
				)
				return result, nil
			}
			m.logger.Warn("Fallback provider failed",
				"request_id", requestID,
				"business_id", businessID,
				"provider", providerName,
				"error", err.Error(),
			)
		}
	}

	// If no providers available, return mock data
	m.logger.Warn("No media providers available, returning mock data",
		"request_id", requestID,
		"business_id", businessID,
	)
	return m.generateMockNewsResult(businessID, query), nil
}

// GetSocialMediaMentions retrieves social media mentions from available providers
func (m *MediaProviderManager) GetSocialMediaMentions(ctx context.Context, businessID string, query SocialMediaQuery) (*SocialMediaResult, error) {
	requestID := ctx.Value("request_id").(string)

	m.logger.Info("Retrieving social media mentions",
		"request_id", requestID,
		"business_id", businessID,
		"business_name", query.BusinessName,
	)

	// Try primary provider first
	if provider, exists := m.providers[m.primaryProvider]; exists && provider.IsAvailable() {
		result, err := provider.GetSocialMediaMentions(ctx, businessID, query)
		if err == nil {
			m.logger.Info("Retrieved social media mentions from primary provider",
				"request_id", requestID,
				"business_id", businessID,
				"provider", m.primaryProvider,
				"total_mentions", result.TotalMentions,
			)
			return result, nil
		}
	}

	// Try fallback providers
	for _, providerName := range m.fallbackProviders {
		if provider, exists := m.providers[providerName]; exists && provider.IsAvailable() {
			result, err := provider.GetSocialMediaMentions(ctx, businessID, query)
			if err == nil {
				m.logger.Info("Retrieved social media mentions from fallback provider",
					"request_id", requestID,
					"business_id", businessID,
					"provider", providerName,
					"total_mentions", result.TotalMentions,
				)
				return result, nil
			}
		}
	}

	// Return mock social media data
	m.logger.Warn("No media providers available, returning mock data",
		"request_id", requestID,
		"business_id", businessID,
	)
	return m.generateMockSocialMediaResult(businessID, query), nil
}

// GetMediaSentiment retrieves media sentiment from available providers
func (m *MediaProviderManager) GetMediaSentiment(ctx context.Context, businessID string) (*SentimentResult, error) {
	requestID := ctx.Value("request_id").(string)

	m.logger.Info("Retrieving media sentiment",
		"request_id", requestID,
		"business_id", businessID,
	)

	// Try primary provider first
	if provider, exists := m.providers[m.primaryProvider]; exists && provider.IsAvailable() {
		result, err := provider.GetMediaSentiment(ctx, businessID)
		if err == nil {
			m.logger.Info("Retrieved media sentiment from primary provider",
				"request_id", requestID,
				"business_id", businessID,
				"provider", m.primaryProvider,
				"overall_score", result.OverallScore,
			)
			return result, nil
		}
	}

	// Try fallback providers
	for _, providerName := range m.fallbackProviders {
		if provider, exists := m.providers[providerName]; exists && provider.IsAvailable() {
			result, err := provider.GetMediaSentiment(ctx, businessID)
			if err == nil {
				m.logger.Info("Retrieved media sentiment from fallback provider",
					"request_id", requestID,
					"business_id", businessID,
					"provider", providerName,
					"overall_score", result.OverallScore,
				)
				return result, nil
			}
		}
	}

	// Return mock sentiment data
	m.logger.Warn("No media providers available, returning mock data",
		"request_id", requestID,
		"business_id", businessID,
	)
	return m.generateMockSentimentResult(businessID), nil
}

// GetReputationScore retrieves reputation score from available providers
func (m *MediaProviderManager) GetReputationScore(ctx context.Context, businessID string) (*ReputationScore, error) {
	requestID := ctx.Value("request_id").(string)

	m.logger.Info("Retrieving reputation score",
		"request_id", requestID,
		"business_id", businessID,
	)

	// Try primary provider first
	if provider, exists := m.providers[m.primaryProvider]; exists && provider.IsAvailable() {
		result, err := provider.GetReputationScore(ctx, businessID)
		if err == nil {
			m.logger.Info("Retrieved reputation score from primary provider",
				"request_id", requestID,
				"business_id", businessID,
				"provider", m.primaryProvider,
				"overall_score", result.OverallScore,
			)
			return result, nil
		}
	}

	// Try fallback providers
	for _, providerName := range m.fallbackProviders {
		if provider, exists := m.providers[providerName]; exists && provider.IsAvailable() {
			result, err := provider.GetReputationScore(ctx, businessID)
			if err == nil {
				m.logger.Info("Retrieved reputation score from fallback provider",
					"request_id", requestID,
					"business_id", businessID,
					"provider", providerName,
					"overall_score", result.OverallScore,
				)
				return result, nil
			}
		}
	}

	// Return mock reputation score
	m.logger.Warn("No media providers available, returning mock data",
		"request_id", requestID,
		"business_id", businessID,
	)
	return m.generateMockReputationScore(businessID), nil
}

// GetMediaAlerts retrieves media alerts from available providers
func (m *MediaProviderManager) GetMediaAlerts(ctx context.Context, businessID string) (*MediaAlerts, error) {
	requestID := ctx.Value("request_id").(string)

	m.logger.Info("Retrieving media alerts",
		"request_id", requestID,
		"business_id", businessID,
	)

	// Try primary provider first
	if provider, exists := m.providers[m.primaryProvider]; exists && provider.IsAvailable() {
		result, err := provider.GetMediaAlerts(ctx, businessID)
		if err == nil {
			m.logger.Info("Retrieved media alerts from primary provider",
				"request_id", requestID,
				"business_id", businessID,
				"provider", m.primaryProvider,
				"total_alerts", result.TotalAlerts,
			)
			return result, nil
		}
	}

	// Try fallback providers
	for _, providerName := range m.fallbackProviders {
		if provider, exists := m.providers[providerName]; exists && provider.IsAvailable() {
			result, err := provider.GetMediaAlerts(ctx, businessID)
			if err == nil {
				m.logger.Info("Retrieved media alerts from fallback provider",
					"request_id", requestID,
					"business_id", businessID,
					"provider", providerName,
					"total_alerts", result.TotalAlerts,
				)
				return result, nil
			}
		}
	}

	// Return mock media alerts
	m.logger.Warn("No media providers available, returning mock data",
		"request_id", requestID,
		"business_id", businessID,
	)
	return m.generateMockMediaAlerts(businessID), nil
}

// Mock data generation functions
func (m *MediaProviderManager) generateMockNewsResult(businessID string, query NewsQuery) *NewsResult {
	return &NewsResult{
		BusinessID:       businessID,
		Provider:         "mock_media_provider",
		LastUpdated:      time.Now(),
		TotalArticles:    15,
		PositiveCount:    8,
		NegativeCount:    2,
		NeutralCount:     5,
		RiskLevel:        RiskLevelLow,
		OverallSentiment: 0.6,
		Articles: []NewsArticle{
			{
				ArticleID:      "article_1",
				Title:          "Company Achieves Record Growth",
				Content:        "The company has achieved record growth in the last quarter...",
				URL:            "https://example.com/article1",
				Source:         "Business News",
				Author:         "John Smith",
				PublishedDate:  time.Now().AddDate(0, 0, -5),
				Sentiment:      "positive",
				SentimentScore: 0.8,
				RelevanceScore: 0.9,
				RiskLevel:      RiskLevelLow,
				Language:       "en",
				Country:        "US",
			},
		},
		Metadata: map[string]interface{}{
			"data_quality": "mock",
			"confidence":   0.85,
		},
	}
}

func (m *MediaProviderManager) generateMockSocialMediaResult(businessID string, query SocialMediaQuery) *SocialMediaResult {
	return &SocialMediaResult{
		BusinessID:       businessID,
		Provider:         "mock_media_provider",
		LastUpdated:      time.Now(),
		TotalMentions:    25,
		PositiveCount:    15,
		NegativeCount:    3,
		NeutralCount:     7,
		RiskLevel:        RiskLevelLow,
		OverallSentiment: 0.7,
		Mentions: []SocialMediaMention{
			{
				MentionID:      "mention_1",
				Platform:       "twitter",
				Username:       "@user1",
				Content:        "Great experience with this company!",
				URL:            "https://twitter.com/user1/status/123",
				PublishedDate:  time.Now().AddDate(0, 0, -2),
				Sentiment:      "positive",
				SentimentScore: 0.9,
				RelevanceScore: 0.8,
				RiskLevel:      RiskLevelLow,
				Engagement:     15,
				Followers:      1000,
				Language:       "en",
				Country:        "US",
			},
		},
		Metadata: map[string]interface{}{
			"data_quality": "mock",
			"confidence":   0.80,
		},
	}
}

func (m *MediaProviderManager) generateMockSentimentResult(businessID string) *SentimentResult {
	return &SentimentResult{
		BusinessID:    businessID,
		Provider:      "mock_media_provider",
		LastUpdated:   time.Now(),
		OverallScore:  0.7,
		PositiveScore: 0.6,
		NegativeScore: 0.2,
		NeutralScore:  0.2,
		Confidence:    0.85,
		RiskLevel:     RiskLevelLow,
		Trend:         "stable",
		Metadata: map[string]interface{}{
			"data_quality": "mock",
			"confidence":   0.85,
		},
	}
}

func (m *MediaProviderManager) generateMockReputationScore(businessID string) *ReputationScore {
	return &ReputationScore{
		BusinessID:   businessID,
		Provider:     "mock_media_provider",
		LastUpdated:  time.Now(),
		OverallScore: 75.0,
		NewsScore:    80.0,
		SocialScore:  70.0,
		ReviewScore:  75.0,
		RiskLevel:    RiskLevelLow,
		Trend:        "stable",
		Confidence:   0.85,
		Metadata: map[string]interface{}{
			"data_quality": "mock",
			"confidence":   0.85,
		},
	}
}

func (m *MediaProviderManager) generateMockMediaAlerts(businessID string) *MediaAlerts {
	return &MediaAlerts{
		BusinessID:     businessID,
		Provider:       "mock_media_provider",
		LastUpdated:    time.Now(),
		TotalAlerts:    2,
		HighPriority:   0,
		MediumPriority: 1,
		LowPriority:    1,
		RiskLevel:      RiskLevelLow,
		Alerts: []MediaAlert{
			{
				AlertID:     "alert_1",
				AlertType:   "negative_news",
				Priority:    "medium",
				Title:       "Minor Customer Complaint",
				Description: "A customer posted a complaint on social media",
				Source:      "Twitter",
				CreatedDate: time.Now().AddDate(0, 0, -1),
				RiskLevel:   RiskLevelLow,
				Resolved:    false,
			},
		},
		Metadata: map[string]interface{}{
			"data_quality": "mock",
			"confidence":   0.85,
		},
	}
}
