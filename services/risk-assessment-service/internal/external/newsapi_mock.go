package external

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

// NewsAPIMock provides a mock implementation of NewsAPI for risk assessment
type NewsAPIMock struct {
	logger *zap.Logger
}

// NewNewsAPIMock creates a new NewsAPIMock
func NewNewsAPIMock(logger *zap.Logger) *NewsAPIMock {
	return &NewsAPIMock{
		logger: logger,
	}
}

// NewsArticle represents a news article
type NewsArticle struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	URL         string    `json:"url"`
	PublishedAt time.Time `json:"published_at"`
	Source      string    `json:"source"`
	Sentiment   string    `json:"sentiment"` // positive, negative, neutral
	Relevance   float64   `json:"relevance"` // 0.0 to 1.0
}

// NewsResponse represents the response from NewsAPI
type NewsResponse struct {
	Articles []NewsArticle `json:"articles"`
	Total    int           `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// SearchNews searches for news articles related to a business
func (n *NewsAPIMock) SearchNews(ctx context.Context, businessName string, industry string, limit int) (*NewsResponse, error) {
	n.logger.Info("Searching news for business",
		zap.String("business_name", businessName),
		zap.String("industry", industry),
		zap.Int("limit", limit))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(100)+50) * time.Millisecond)

	// Generate mock news articles based on business characteristics
	articles := n.generateMockArticles(businessName, industry, limit)

	response := &NewsResponse{
		Articles: articles,
		Total:    len(articles),
		Page:     1,
		PageSize: limit,
	}

	n.logger.Info("News search completed",
		zap.String("business_name", businessName),
		zap.Int("articles_found", len(articles)))

	return response, nil
}

// generateMockArticles generates mock news articles
func (n *NewsAPIMock) generateMockArticles(businessName, industry string, limit int) []NewsArticle {
	articles := make([]NewsArticle, 0, limit)

	// Generate articles based on industry and business characteristics
	baseArticles := n.getBaseArticles(industry)

	for i := 0; i < limit && i < len(baseArticles); i++ {
		article := baseArticles[i]

		// Customize article for specific business
		article.Title = fmt.Sprintf("%s: %s", businessName, article.Title)
		article.Description = fmt.Sprintf("Recent developments regarding %s in the %s industry. %s",
			businessName, industry, article.Description)
		article.URL = fmt.Sprintf("https://news.example.com/article/%d", i+1)
		article.PublishedAt = time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour)
		article.Relevance = n.calculateRelevance(businessName, industry, article)

		articles = append(articles, article)
	}

	return articles
}

// getBaseArticles returns base articles for different industries
func (n *NewsAPIMock) getBaseArticles(industry string) []NewsArticle {
	baseArticles := map[string][]NewsArticle{
		"finance": {
			{
				Title:       "Regulatory Compliance Update",
				Description: "New regulations affecting financial services companies.",
				Source:      "Financial Times",
				Sentiment:   "neutral",
			},
			{
				Title:       "Market Performance Analysis",
				Description: "Analysis of recent market trends and their impact on financial institutions.",
				Source:      "Reuters",
				Sentiment:   "positive",
			},
			{
				Title:       "Cybersecurity Concerns",
				Description: "Rising cybersecurity threats in the financial sector.",
				Source:      "TechCrunch",
				Sentiment:   "negative",
			},
		},
		"healthcare": {
			{
				Title:       "Healthcare Innovation",
				Description: "New technologies transforming healthcare delivery.",
				Source:      "Healthcare Weekly",
				Sentiment:   "positive",
			},
			{
				Title:       "Regulatory Changes",
				Description: "Updates to healthcare regulations and compliance requirements.",
				Source:      "Health Policy Journal",
				Sentiment:   "neutral",
			},
			{
				Title:       "Patient Safety Focus",
				Description: "Industry focus on improving patient safety standards.",
				Source:      "Medical News Today",
				Sentiment:   "positive",
			},
		},
		"technology": {
			{
				Title:       "AI and Machine Learning",
				Description: "Advancements in AI and ML technologies.",
				Source:      "TechCrunch",
				Sentiment:   "positive",
			},
			{
				Title:       "Data Privacy Regulations",
				Description: "New data privacy laws affecting tech companies.",
				Source:      "Wired",
				Sentiment:   "neutral",
			},
			{
				Title:       "Cybersecurity Threats",
				Description: "Increasing cybersecurity challenges for technology companies.",
				Source:      "Security Weekly",
				Sentiment:   "negative",
			},
		},
		"retail": {
			{
				Title:       "E-commerce Growth",
				Description: "Continued growth in online retail and e-commerce.",
				Source:      "Retail Dive",
				Sentiment:   "positive",
			},
			{
				Title:       "Supply Chain Challenges",
				Description: "Ongoing supply chain disruptions affecting retail.",
				Source:      "Supply Chain News",
				Sentiment:   "negative",
			},
			{
				Title:       "Consumer Behavior Changes",
				Description: "Shifts in consumer shopping patterns and preferences.",
				Source:      "Consumer Reports",
				Sentiment:   "neutral",
			},
		},
		"manufacturing": {
			{
				Title:       "Industrial Automation",
				Description: "Advances in manufacturing automation and robotics.",
				Source:      "Manufacturing Today",
				Sentiment:   "positive",
			},
			{
				Title:       "Supply Chain Resilience",
				Description: "Building more resilient supply chains in manufacturing.",
				Source:      "Industry Week",
				Sentiment:   "positive",
			},
			{
				Title:       "Environmental Regulations",
				Description: "New environmental regulations affecting manufacturing.",
				Source:      "Environmental News",
				Sentiment:   "neutral",
			},
		},
	}

	if articles, exists := baseArticles[industry]; exists {
		return articles
	}

	// Default articles for unknown industries
	return []NewsArticle{
		{
			Title:       "Industry Update",
			Description: "General industry news and developments.",
			Source:      "Business News",
			Sentiment:   "neutral",
		},
		{
			Title:       "Market Trends",
			Description: "Current market trends and their impact on businesses.",
			Source:      "Market Watch",
			Sentiment:   "positive",
		},
	}
}

// calculateRelevance calculates the relevance of an article to a business
func (n *NewsAPIMock) calculateRelevance(businessName, industry string, article NewsArticle) float64 {
	relevance := 0.5 // Base relevance

	// Higher relevance for industry-specific articles
	if article.Source != "Business News" && article.Source != "Market Watch" {
		relevance += 0.2
	}

	// Adjust based on sentiment
	switch article.Sentiment {
	case "positive":
		relevance += 0.1
	case "negative":
		relevance += 0.15 // Negative news is often more relevant for risk assessment
	case "neutral":
		relevance += 0.05
	}

	// Add some randomness
	relevance += rand.Float64() * 0.1

	// Ensure relevance is within bounds
	if relevance > 1.0 {
		relevance = 1.0
	}
	if relevance < 0.0 {
		relevance = 0.0
	}

	return relevance
}

// GetSentimentAnalysis analyzes the sentiment of news articles
func (n *NewsAPIMock) GetSentimentAnalysis(ctx context.Context, articles []NewsArticle) (map[string]float64, error) {
	n.logger.Info("Analyzing sentiment of news articles", zap.Int("article_count", len(articles)))

	sentimentScores := map[string]float64{
		"positive": 0.0,
		"negative": 0.0,
		"neutral":  0.0,
	}

	if len(articles) == 0 {
		return sentimentScores, nil
	}

	// Count sentiment distribution
	for _, article := range articles {
		sentimentScores[article.Sentiment]++
	}

	// Convert to percentages
	total := float64(len(articles))
	for sentiment := range sentimentScores {
		sentimentScores[sentiment] = sentimentScores[sentiment] / total
	}

	n.logger.Info("Sentiment analysis completed",
		zap.Float64("positive", sentimentScores["positive"]),
		zap.Float64("negative", sentimentScores["negative"]),
		zap.Float64("neutral", sentimentScores["neutral"]))

	return sentimentScores, nil
}
