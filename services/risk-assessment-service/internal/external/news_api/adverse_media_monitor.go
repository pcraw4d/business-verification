package news_api

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// AdverseMediaMonitor provides automated adverse media monitoring
type AdverseMediaMonitor struct {
	logger *zap.Logger
	config *AdverseMediaMonitorConfig
	scorer *AdverseMediaRiskScorer
}

// AdverseMediaMonitorConfig holds configuration for adverse media monitoring
type AdverseMediaMonitorConfig struct {
	APIKey    string        `json:"api_key"`
	BaseURL   string        `json:"base_url"`
	Timeout   time.Duration `json:"timeout"`
	RateLimit int           `json:"rate_limit_per_minute"`
	Enabled   bool          `json:"enabled"`
	// Monitoring configuration
	ScanInterval       time.Duration `json:"scan_interval"`         // How often to scan for new articles
	RetentionDays      int           `json:"retention_days"`        // How long to keep historical data
	MaxArticlesPerScan int           `json:"max_articles_per_scan"` // Maximum articles to process per scan
	// Risk scoring configuration
	HighRiskThreshold   float64 `json:"high_risk_threshold"`   // 0.0-1.0
	MediumRiskThreshold float64 `json:"medium_risk_threshold"` // 0.0-1.0
	// Alerting configuration
	EnableAlerts    bool     `json:"enable_alerts"`
	AlertEmail      string   `json:"alert_email"`
	AlertWebhookURL string   `json:"alert_webhook_url"`
	AlertKeywords   []string `json:"alert_keywords"`
}

// AdverseMediaArticle represents an adverse media article
type AdverseMediaArticle struct {
	ArticleID     string    `json:"article_id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	URL           string    `json:"url"`
	Source        string    `json:"source"`
	Author        string    `json:"author"`
	PublishedDate time.Time `json:"published_date"`
	ScrapedDate   time.Time `json:"scraped_date"`
	Language      string    `json:"language"`
	Country       string    `json:"country"`
	// Risk assessment
	RiskScore      float64  `json:"risk_score"`
	RiskLevel      string   `json:"risk_level"` // "low", "medium", "high", "critical"
	Severity       string   `json:"severity"`   // "minor", "moderate", "severe", "critical"
	Category       string   `json:"category"`   // "financial_crime", "corruption", "fraud", etc.
	Subcategory    string   `json:"subcategory"`
	Keywords       []string `json:"keywords"`
	Sentiment      string   `json:"sentiment"`       // "positive", "negative", "neutral"
	SentimentScore float64  `json:"sentiment_score"` // -1.0 to 1.0
	// Entity matching
	MatchedEntities  []string `json:"matched_entities"`
	EntityConfidence float64  `json:"entity_confidence"`
	// Metadata
	DataQuality string    `json:"data_quality"`
	LastUpdated time.Time `json:"last_updated"`
}

// AdverseMediaScanResult represents the result of an adverse media scan
type AdverseMediaScanResult struct {
	ScanID             string                `json:"scan_id"`
	QueryTerms         []string              `json:"query_terms"`
	ScanDate           time.Time             `json:"scan_date"`
	TotalArticles      int                   `json:"total_articles"`
	NewArticles        int                   `json:"new_articles"`
	HighRiskArticles   int                   `json:"high_risk_articles"`
	MediumRiskArticles int                   `json:"medium_risk_articles"`
	LowRiskArticles    int                   `json:"low_risk_articles"`
	Articles           []AdverseMediaArticle `json:"articles"`
	ScanTime           time.Duration         `json:"scan_time"`
	DataQuality        string                `json:"data_quality"`
	Sources            []string              `json:"sources"`
}

// AdverseMediaAlert represents an alert for high-risk adverse media
type AdverseMediaAlert struct {
	AlertID         string                `json:"alert_id"`
	EntityName      string                `json:"entity_name"`
	AlertType       string                `json:"alert_type"` // "new_high_risk", "escalation", "trending"
	Severity        string                `json:"severity"`
	RiskScore       float64               `json:"risk_score"`
	ArticleCount    int                   `json:"article_count"`
	Articles        []AdverseMediaArticle `json:"articles"`
	AlertDate       time.Time             `json:"alert_date"`
	IsResolved      bool                  `json:"is_resolved"`
	ResolvedDate    *time.Time            `json:"resolved_date,omitempty"`
	ResolvedBy      string                `json:"resolved_by,omitempty"`
	ResolutionNotes string                `json:"resolution_notes,omitempty"`
}

// AdverseMediaTrend represents trending adverse media data
type AdverseMediaTrend struct {
	EntityName       string    `json:"entity_name"`
	Timeframe        string    `json:"timeframe"` // "24h", "7d", "30d", "90d"
	TotalArticles    int       `json:"total_articles"`
	HighRiskCount    int       `json:"high_risk_count"`
	MediumRiskCount  int       `json:"medium_risk_count"`
	LowRiskCount     int       `json:"low_risk_count"`
	AverageRiskScore float64   `json:"average_risk_score"`
	TrendDirection   string    `json:"trend_direction"` // "increasing", "decreasing", "stable"
	TrendPercentage  float64   `json:"trend_percentage"`
	LastUpdated      time.Time `json:"last_updated"`
}

// NewAdverseMediaMonitor creates a new adverse media monitor
func NewAdverseMediaMonitor(config *AdverseMediaMonitorConfig, logger *zap.Logger) *AdverseMediaMonitor {
	scorer := NewAdverseMediaRiskScorer(config, logger)

	return &AdverseMediaMonitor{
		logger: logger,
		config: config,
		scorer: scorer,
	}
}

// StartMonitoring starts continuous adverse media monitoring
func (amm *AdverseMediaMonitor) StartMonitoring(ctx context.Context, entityNames []string) error {
	amm.logger.Info("Starting adverse media monitoring",
		zap.Int("entity_count", len(entityNames)),
		zap.Duration("scan_interval", amm.config.ScanInterval))

	ticker := time.NewTicker(amm.config.ScanInterval)
	defer ticker.Stop()

	// Initial scan
	if _, err := amm.performScan(ctx, entityNames); err != nil {
		amm.logger.Error("Initial adverse media scan failed", zap.Error(err))
	}

	// Continuous monitoring
	for {
		select {
		case <-ctx.Done():
			amm.logger.Info("Adverse media monitoring stopped")
			return ctx.Err()
		case <-ticker.C:
			if _, err := amm.performScan(ctx, entityNames); err != nil {
				amm.logger.Error("Adverse media scan failed", zap.Error(err))
			}
		}
	}
}

// PerformScan performs a single adverse media scan
func (amm *AdverseMediaMonitor) PerformScan(ctx context.Context, entityNames []string) (*AdverseMediaScanResult, error) {
	return amm.performScan(ctx, entityNames)
}

func (amm *AdverseMediaMonitor) performScan(ctx context.Context, entityNames []string) (*AdverseMediaScanResult, error) {
	startTime := time.Now()
	amm.logger.Info("Performing adverse media scan",
		zap.Int("entity_count", len(entityNames)))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(1000)+500) * time.Millisecond)

	// Generate mock adverse media articles
	articles := amm.generateAdverseMediaArticles(entityNames)

	// Score articles for risk
	for i := range articles {
		articles[i] = amm.scorer.ScoreArticle(articles[i])
	}

	// Categorize articles by risk level
	highRiskCount := 0
	mediumRiskCount := 0
	lowRiskCount := 0

	for _, article := range articles {
		switch article.RiskLevel {
		case "high", "critical":
			highRiskCount++
		case "medium":
			mediumRiskCount++
		case "low":
			lowRiskCount++
		}
	}

	// Generate sources
	sources := []string{"NewsAPI", "Google News", "Reuters", "AP News", "BBC News", "CNN", "Financial Times"}

	result := &AdverseMediaScanResult{
		ScanID:             amm.generateScanID(),
		QueryTerms:         entityNames,
		ScanDate:           time.Now(),
		TotalArticles:      len(articles),
		NewArticles:        len(articles), // All articles are "new" in mock
		HighRiskArticles:   highRiskCount,
		MediumRiskArticles: mediumRiskCount,
		LowRiskArticles:    lowRiskCount,
		Articles:           articles,
		ScanTime:           time.Since(startTime),
		DataQuality:        amm.generateDataQuality(),
		Sources:            sources,
	}

	// Check for alerts
	if amm.config.EnableAlerts {
		amm.checkForAlerts(ctx, result)
	}

	amm.logger.Info("Adverse media scan completed",
		zap.String("scan_id", result.ScanID),
		zap.Int("total_articles", result.TotalArticles),
		zap.Int("high_risk_articles", result.HighRiskArticles),
		zap.Duration("scan_time", result.ScanTime))

	return result, nil
}

// GetHistoricalData retrieves historical adverse media data
func (amm *AdverseMediaMonitor) GetHistoricalData(ctx context.Context, entityName string, days int) ([]AdverseMediaArticle, error) {
	amm.logger.Info("Retrieving historical adverse media data",
		zap.String("entity_name", entityName),
		zap.Int("days", days))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(500)+200) * time.Millisecond)

	// Generate mock historical data
	var articles []AdverseMediaArticle
	numArticles := rand.Intn(50) + 10 // 10-60 articles

	for i := 0; i < numArticles; i++ {
		article := AdverseMediaArticle{
			ArticleID:        fmt.Sprintf("HIST_%d", rand.Intn(100000)),
			Title:            amm.generateArticleTitle(entityName),
			Content:          amm.generateArticleContent(entityName),
			URL:              fmt.Sprintf("https://example.com/article/%d", rand.Intn(10000)),
			Source:           amm.generateSource(),
			Author:           amm.generateAuthor(),
			PublishedDate:    time.Now().Add(-time.Duration(rand.Intn(days*24)) * time.Hour),
			ScrapedDate:      time.Now().Add(-time.Duration(rand.Intn(days*24)) * time.Hour),
			Language:         "en",
			Country:          "US",
			RiskScore:        rand.Float64(),
			RiskLevel:        amm.generateRiskLevel(),
			Severity:         amm.generateSeverity(),
			Category:         amm.generateCategory(),
			Subcategory:      amm.generateSubcategory(),
			Keywords:         strings.Split(amm.generateKeywords(entityName), ", "),
			Sentiment:        amm.generateSentiment(),
			SentimentScore:   rand.Float64()*2 - 1, // -1.0 to 1.0
			MatchedEntities:  []string{entityName},
			EntityConfidence: rand.Float64()*0.3 + 0.7, // 0.7-1.0
			DataQuality:      amm.generateDataQuality(),
			LastUpdated:      time.Now(),
		}

		// Score the article
		article = amm.scorer.ScoreArticle(article)
		articles = append(articles, article)
	}

	amm.logger.Info("Historical adverse media data retrieved",
		zap.String("entity_name", entityName),
		zap.Int("article_count", len(articles)))

	return articles, nil
}

// GetTrendingData retrieves trending adverse media data
func (amm *AdverseMediaMonitor) GetTrendingData(ctx context.Context, entityName string, timeframe string) (*AdverseMediaTrend, error) {
	amm.logger.Info("Retrieving trending adverse media data",
		zap.String("entity_name", entityName),
		zap.String("timeframe", timeframe))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(300)+100) * time.Millisecond)

	// Generate mock trending data
	trend := &AdverseMediaTrend{
		EntityName:       entityName,
		Timeframe:        timeframe,
		TotalArticles:    rand.Intn(100) + 10,
		HighRiskCount:    rand.Intn(20) + 1,
		MediumRiskCount:  rand.Intn(30) + 5,
		LowRiskCount:     rand.Intn(50) + 10,
		AverageRiskScore: rand.Float64()*0.8 + 0.2, // 0.2-1.0
		TrendDirection:   amm.generateTrendDirection(),
		TrendPercentage:  rand.Float64()*100 - 50, // -50% to +50%
		LastUpdated:      time.Now(),
	}

	amm.logger.Info("Trending adverse media data retrieved",
		zap.String("entity_name", entityName),
		zap.String("timeframe", timeframe),
		zap.String("trend_direction", trend.TrendDirection))

	return trend, nil
}

// GetAlerts retrieves active adverse media alerts
func (amm *AdverseMediaMonitor) GetAlerts(ctx context.Context, entityName string) ([]AdverseMediaAlert, error) {
	amm.logger.Info("Retrieving adverse media alerts",
		zap.String("entity_name", entityName))

	// Simulate API delay
	time.Sleep(time.Duration(rand.Intn(200)+100) * time.Millisecond)

	// Generate mock alerts
	var alerts []AdverseMediaAlert
	numAlerts := rand.Intn(5) + 1 // 1-5 alerts

	for i := 0; i < numAlerts; i++ {
		alert := AdverseMediaAlert{
			AlertID:      fmt.Sprintf("ALERT_%d", rand.Intn(10000)),
			EntityName:   entityName,
			AlertType:    amm.generateAlertType(),
			Severity:     amm.generateSeverity(),
			RiskScore:    rand.Float64()*0.5 + 0.5, // 0.5-1.0
			ArticleCount: rand.Intn(10) + 1,
			Articles:     []AdverseMediaArticle{},                                    // Would be populated with actual articles
			AlertDate:    time.Now().Add(-time.Duration(rand.Intn(168)) * time.Hour), // Last week
			IsResolved:   rand.Float64() > 0.7,                                       // 30% resolved
		}

		if alert.IsResolved {
			resolvedDate := alert.AlertDate.Add(time.Duration(rand.Intn(72)) * time.Hour)
			alert.ResolvedDate = &resolvedDate
			alert.ResolvedBy = "system"
			alert.ResolutionNotes = "Automatically resolved after review"
		}

		alerts = append(alerts, alert)
	}

	amm.logger.Info("Adverse media alerts retrieved",
		zap.String("entity_name", entityName),
		zap.Int("alert_count", len(alerts)))

	return alerts, nil
}

// IsHealthy checks if the adverse media monitoring service is healthy
func (amm *AdverseMediaMonitor) IsHealthy(ctx context.Context) error {
	amm.logger.Info("Checking adverse media monitoring service health (mock)")

	// Simulate health check
	time.Sleep(50 * time.Millisecond)

	// Mock health check - always healthy
	return nil
}

// GenerateRiskFactors generates risk factors from adverse media data
func (amm *AdverseMediaMonitor) GenerateRiskFactors(scanResult *AdverseMediaScanResult) []models.RiskFactor {
	var riskFactors []models.RiskFactor
	now := time.Now()

	// Overall adverse media risk factor
	overallRiskScore := 0.1 // Base risk
	if scanResult.HighRiskArticles > 0 {
		overallRiskScore = 0.8 // High risk if high-risk articles found
	} else if scanResult.MediumRiskArticles > 0 {
		overallRiskScore = 0.5 // Medium risk if medium-risk articles found
	} else if scanResult.LowRiskArticles > 0 {
		overallRiskScore = 0.3 // Low risk if low-risk articles found
	}

	riskFactors = append(riskFactors, models.RiskFactor{
		Category:    models.RiskCategoryReputational,
		Subcategory: "adverse_media",
		Name:        "adverse_media_risk",
		Score:       overallRiskScore,
		Weight:      0.3,
		Description: "Risk associated with adverse media coverage",
		Source:      "adverse_media_monitor",
		Confidence:  0.85,
		Impact:      "Adverse media can impact business reputation and relationships",
		Mitigation:  "Monitor media coverage and assess reputational impact",
		LastUpdated: &now,
	})

	// Category-specific risk factors
	categoryRisks := make(map[string]int)
	for _, article := range scanResult.Articles {
		categoryRisks[article.Category]++
	}

	for category, count := range categoryRisks {
		if count > 0 {
			categoryRisk := 0.4 // Base category risk
			if count >= 5 {
				categoryRisk = 0.8 // High risk for many articles
			} else if count >= 2 {
				categoryRisk = 0.6 // Medium risk for multiple articles
			}

			riskFactors = append(riskFactors, models.RiskFactor{
				Category:    models.RiskCategoryReputational,
				Subcategory: category,
				Name:        fmt.Sprintf("%s_media_risk", category),
				Score:       categoryRisk,
				Weight:      0.2,
				Description: fmt.Sprintf("Risk associated with %s media coverage", category),
				Source:      "adverse_media_monitor",
				Confidence:  0.80,
				Impact:      fmt.Sprintf("%s coverage can impact business reputation", category),
				Mitigation:  "Monitor and address specific media coverage",
				LastUpdated: &now,
			})
		}
	}

	return riskFactors
}

// Helper methods

func (amm *AdverseMediaMonitor) checkForAlerts(ctx context.Context, scanResult *AdverseMediaScanResult) {
	// Check for high-risk articles that need alerts
	for _, article := range scanResult.Articles {
		if article.RiskLevel == "high" || article.RiskLevel == "critical" {
			amm.logger.Warn("High-risk adverse media article detected",
				zap.String("article_id", article.ArticleID),
				zap.String("title", article.Title),
				zap.Float64("risk_score", article.RiskScore),
				zap.String("risk_level", article.RiskLevel))

			// In a real implementation, this would send alerts via email/webhook
			if amm.config.AlertEmail != "" {
				amm.logger.Info("Alert would be sent to email",
					zap.String("email", amm.config.AlertEmail),
					zap.String("article_id", article.ArticleID))
			}

			if amm.config.AlertWebhookURL != "" {
				amm.logger.Info("Alert would be sent to webhook",
					zap.String("webhook_url", amm.config.AlertWebhookURL),
					zap.String("article_id", article.ArticleID))
			}
		}
	}
}

func (amm *AdverseMediaMonitor) generateAdverseMediaArticles(entityNames []string) []AdverseMediaArticle {
	var articles []AdverseMediaArticle

	// 80% of entities have no adverse media
	if rand.Float64() > 0.2 {
		return articles
	}

	// Generate 1-10 articles for entities with adverse media
	numArticles := rand.Intn(10) + 1
	for i := 0; i < numArticles; i++ {
		entityName := entityNames[rand.Intn(len(entityNames))]

		article := AdverseMediaArticle{
			ArticleID:        fmt.Sprintf("ADVERSE_%d", rand.Intn(100000)),
			Title:            amm.generateArticleTitle(entityName),
			Content:          amm.generateArticleContent(entityName),
			URL:              fmt.Sprintf("https://example.com/article/%d", rand.Intn(10000)),
			Source:           amm.generateSource(),
			Author:           amm.generateAuthor(),
			PublishedDate:    time.Now().Add(-time.Duration(rand.Intn(168)) * time.Hour), // Last week
			ScrapedDate:      time.Now(),
			Language:         "en",
			Country:          "US",
			RiskScore:        rand.Float64(),
			RiskLevel:        amm.generateRiskLevel(),
			Severity:         amm.generateSeverity(),
			Category:         amm.generateCategory(),
			Subcategory:      amm.generateSubcategory(),
			Keywords:         strings.Split(amm.generateKeywords(entityName), ", "),
			Sentiment:        amm.generateSentiment(),
			SentimentScore:   rand.Float64()*2 - 1, // -1.0 to 1.0
			MatchedEntities:  []string{entityName},
			EntityConfidence: rand.Float64()*0.3 + 0.7, // 0.7-1.0
			DataQuality:      amm.generateDataQuality(),
			LastUpdated:      time.Now(),
		}

		articles = append(articles, article)
	}

	return articles
}

func (amm *AdverseMediaMonitor) generateArticleTitle(entityName string) string {
	templates := []string{
		"%s under investigation for financial irregularities",
		"Regulatory action taken against %s",
		"%s faces legal challenges in court",
		"Whistleblower allegations against %s",
		"%s executives under scrutiny",
		"Compliance issues reported at %s",
		"%s involved in regulatory violation",
		"Legal proceedings initiated against %s",
		"%s under regulatory review",
		"Financial misconduct allegations at %s",
	}

	template := templates[rand.Intn(len(templates))]
	return fmt.Sprintf(template, entityName)
}

func (amm *AdverseMediaMonitor) generateArticleContent(entityName string) string {
	content := fmt.Sprintf("Recent developments have brought %s under scrutiny. ", entityName)
	content += "Regulatory authorities are investigating potential violations. "
	content += "The company has stated that it is cooperating fully with the investigation. "
	content += "Further details are expected to be released in the coming weeks."
	return content
}

func (amm *AdverseMediaMonitor) generateSource() string {
	sources := []string{
		"Reuters", "AP News", "BBC News", "CNN", "Financial Times",
		"Wall Street Journal", "Bloomberg", "Forbes", "Business Insider",
		"New York Times", "Washington Post", "Guardian", "Telegraph",
	}
	return sources[rand.Intn(len(sources))]
}

func (amm *AdverseMediaMonitor) generateAuthor() string {
	firstNames := []string{"John", "Jane", "Michael", "Sarah", "David", "Lisa", "Robert", "Emily"}
	lastNames := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis"}
	return fmt.Sprintf("%s %s", firstNames[rand.Intn(len(firstNames))], lastNames[rand.Intn(len(lastNames))])
}

func (amm *AdverseMediaMonitor) generateRiskLevel() string {
	levels := []string{"low", "medium", "high", "critical"}
	weights := []float64{0.4, 0.3, 0.2, 0.1} // 40% low, 30% medium, 20% high, 10% critical

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return levels[i]
		}
	}
	return "medium"
}

func (amm *AdverseMediaMonitor) generateSeverity() string {
	severities := []string{"minor", "moderate", "severe", "critical"}
	weights := []float64{0.3, 0.4, 0.2, 0.1} // 30% minor, 40% moderate, 20% severe, 10% critical

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return severities[i]
		}
	}
	return "moderate"
}

func (amm *AdverseMediaMonitor) generateCategory() string {
	categories := []string{
		"financial_crime", "corruption", "fraud", "money_laundering",
		"regulatory_violation", "enforcement_action", "litigation",
		"investigation", "compliance_failure", "reputational_damage",
	}
	return categories[rand.Intn(len(categories))]
}

func (amm *AdverseMediaMonitor) generateSubcategory() string {
	subcategories := []string{
		"securities_fraud", "tax_evasion", "bribery", "embezzlement",
		"insider_trading", "market_manipulation", "accounting_fraud",
		"regulatory_fine", "cease_and_desist", "criminal_charges",
	}
	return subcategories[rand.Intn(len(subcategories))]
}

func (amm *AdverseMediaMonitor) generateKeywords(entityName string) string {
	keywords := []string{
		"investigation", "regulatory", "violation", "fraud", "corruption",
		"compliance", "enforcement", "legal", "court", "fine", "penalty",
	}

	// Add entity name as keyword
	allKeywords := append([]string{entityName}, keywords...)

	// Return 3-6 random keywords
	numKeywords := rand.Intn(4) + 3
	selectedKeywords := make([]string, 0, numKeywords)

	for i := 0; i < numKeywords && i < len(allKeywords); i++ {
		selectedKeywords = append(selectedKeywords, allKeywords[i])
	}

	return strings.Join(selectedKeywords, ", ")
}

func (amm *AdverseMediaMonitor) generateSentiment() string {
	sentiments := []string{"positive", "negative", "neutral"}
	weights := []float64{0.1, 0.7, 0.2} // 10% positive, 70% negative, 20% neutral (adverse media)

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return sentiments[i]
		}
	}
	return "negative"
}

func (amm *AdverseMediaMonitor) generateAlertType() string {
	types := []string{"new_high_risk", "escalation", "trending", "regulatory_action"}
	return types[rand.Intn(len(types))]
}

func (amm *AdverseMediaMonitor) generateTrendDirection() string {
	directions := []string{"increasing", "decreasing", "stable"}
	weights := []float64{0.4, 0.2, 0.4} // 40% increasing, 20% decreasing, 40% stable

	randVal := rand.Float64()
	cumulative := 0.0
	for i, weight := range weights {
		cumulative += weight
		if randVal <= cumulative {
			return directions[i]
		}
	}
	return "stable"
}

func (amm *AdverseMediaMonitor) generateDataQuality() string {
	qualities := []string{"excellent", "good", "average"}
	return qualities[rand.Intn(len(qualities))]
}

func (amm *AdverseMediaMonitor) generateScanID() string {
	return fmt.Sprintf("SCAN_%d", time.Now().UnixNano())
}
