package external

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// NewsAPIClient provides integration with NewsAPI for adverse media monitoring
type NewsAPIClient struct {
	*Client
}

// NewsAPIResponse represents the response from NewsAPI
type NewsAPIResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

// Article represents a news article
type Article struct {
	Source      Source    `json:"source"`
	Author      string    `json:"author"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	URLToImage  string    `json:"urlToImage"`
	PublishedAt time.Time `json:"publishedAt"`
	Content     string    `json:"content"`
}

// Source represents a news source
type Source struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AdverseMediaResult represents the result of adverse media monitoring
type AdverseMediaResult struct {
	BusinessName    string    `json:"business_name"`
	TotalArticles   int       `json:"total_articles"`
	AdverseArticles []Article `json:"adverse_articles"`
	RiskScore       float64   `json:"risk_score"`
	LastChecked     time.Time `json:"last_checked"`
}

// NewNewsAPIClient creates a new NewsAPI client
func NewNewsAPIClient(apiKey string, logger *zap.Logger) *NewsAPIClient {
	config := Config{
		BaseURL:    "https://newsapi.org/v2",
		APIKey:     apiKey,
		Timeout:    10 * time.Second,
		MaxRetries: 3,
	}

	return &NewsAPIClient{
		Client: NewClient(config, logger),
	}
}

// SearchAdverseMedia searches for adverse media mentions of a business
func (c *NewsAPIClient) SearchAdverseMedia(ctx context.Context, businessName string) (*AdverseMediaResult, error) {
	c.logger.Info("Searching for adverse media",
		zap.String("business_name", businessName))

	// Search for negative keywords related to the business
	negativeKeywords := []string{
		"fraud", "scam", "illegal", "criminal", "arrest", "lawsuit",
		"bankruptcy", "insolvency", "shutdown", "closure", "violation",
		"fine", "penalty", "investigation", "raid", "seizure",
	}

	var allArticles []Article
	totalResults := 0

	// Search for each negative keyword combined with business name
	for _, keyword := range negativeKeywords {
		query := fmt.Sprintf("%s AND %s", businessName, keyword)

		params := map[string]string{
			"q":        query,
			"sortBy":   "publishedAt",
			"pageSize": "10",
			"language": "en",
		}

		resp, err := c.Get(ctx, "/everything", params)
		if err != nil {
			c.logger.Warn("Failed to search for adverse media",
				zap.String("keyword", keyword),
				zap.Error(err))
			continue
		}
		defer resp.Body.Close()

		var newsResponse NewsAPIResponse
		if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
			c.logger.Warn("Failed to decode news response",
				zap.String("keyword", keyword),
				zap.Error(err))
			continue
		}

		if newsResponse.Status == "ok" {
			allArticles = append(allArticles, newsResponse.Articles...)
			totalResults += newsResponse.TotalResults
		}
	}

	// Calculate risk score based on number of adverse articles
	riskScore := c.calculateAdverseMediaRiskScore(len(allArticles), totalResults)

	result := &AdverseMediaResult{
		BusinessName:    businessName,
		TotalArticles:   totalResults,
		AdverseArticles: allArticles,
		RiskScore:       riskScore,
		LastChecked:     time.Now(),
	}

	c.logger.Info("Adverse media search completed",
		zap.String("business_name", businessName),
		zap.Int("total_articles", totalResults),
		zap.Int("adverse_articles", len(allArticles)),
		zap.Float64("risk_score", riskScore))

	return result, nil
}

// SearchRecentNews searches for recent news about a business
func (c *NewsAPIClient) SearchRecentNews(ctx context.Context, businessName string, days int) ([]Article, error) {
	c.logger.Info("Searching for recent news",
		zap.String("business_name", businessName),
		zap.Int("days", days))

	// Calculate date range
	toDate := time.Now()
	fromDate := toDate.AddDate(0, 0, -days)

	params := map[string]string{
		"q":        businessName,
		"from":     fromDate.Format("2006-01-02"),
		"to":       toDate.Format("2006-01-02"),
		"sortBy":   "publishedAt",
		"pageSize": "20",
		"language": "en",
	}

	resp, err := c.Get(ctx, "/everything", params)
	if err != nil {
		return nil, fmt.Errorf("failed to search recent news: %w", err)
	}
	defer resp.Body.Close()

	var newsResponse NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode news response: %w", err)
	}

	if newsResponse.Status != "ok" {
		return nil, fmt.Errorf("news API returned error status: %s", newsResponse.Status)
	}

	c.logger.Info("Recent news search completed",
		zap.String("business_name", businessName),
		zap.Int("articles_found", len(newsResponse.Articles)))

	return newsResponse.Articles, nil
}

// GetTopHeadlines gets top headlines for a specific country or category
func (c *NewsAPIClient) GetTopHeadlines(ctx context.Context, country, category string) ([]Article, error) {
	c.logger.Info("Getting top headlines",
		zap.String("country", country),
		zap.String("category", category))

	params := map[string]string{
		"pageSize": "20",
		"language": "en",
	}

	if country != "" {
		params["country"] = country
	}
	if category != "" {
		params["category"] = category
	}

	resp, err := c.Get(ctx, "/top-headlines", params)
	if err != nil {
		return nil, fmt.Errorf("failed to get top headlines: %w", err)
	}
	defer resp.Body.Close()

	var newsResponse NewsAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&newsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode headlines response: %w", err)
	}

	if newsResponse.Status != "ok" {
		return nil, fmt.Errorf("news API returned error status: %s", newsResponse.Status)
	}

	c.logger.Info("Top headlines retrieved",
		zap.Int("articles_found", len(newsResponse.Articles)))

	return newsResponse.Articles, nil
}

// calculateAdverseMediaRiskScore calculates risk score based on adverse media findings
func (c *NewsAPIClient) calculateAdverseMediaRiskScore(adverseCount, totalCount int) float64 {
	if totalCount == 0 {
		return 0.0
	}

	// Base risk score from adverse articles ratio
	adverseRatio := float64(adverseCount) / float64(totalCount)

	// Scale the risk score (0.0 to 1.0)
	riskScore := adverseRatio * 0.8 // Cap at 0.8 for adverse media alone

	// Add penalty for high volume of adverse articles
	if adverseCount > 10 {
		riskScore += 0.2 // Additional 0.2 for high volume
	} else if adverseCount > 5 {
		riskScore += 0.1 // Additional 0.1 for medium volume
	}

	// Ensure score doesn't exceed 1.0
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	return riskScore
}

// IsHealthy checks if the NewsAPI service is healthy
func (c *NewsAPIClient) IsHealthy(ctx context.Context) error {
	// Try to get top headlines as a health check
	_, err := c.GetTopHeadlines(ctx, "us", "business")
	if err != nil {
		return fmt.Errorf("NewsAPI health check failed: %w", err)
	}
	return nil
}
