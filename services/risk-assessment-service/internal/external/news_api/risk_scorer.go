package news_api

import (
	"math"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AdverseMediaRiskScorer provides automated risk scoring for adverse media articles
type AdverseMediaRiskScorer struct {
	logger *zap.Logger
	config *AdverseMediaMonitorConfig
}

// NewAdverseMediaRiskScorer creates a new adverse media risk scorer
func NewAdverseMediaRiskScorer(config *AdverseMediaMonitorConfig, logger *zap.Logger) *AdverseMediaRiskScorer {
	return &AdverseMediaRiskScorer{
		logger: logger,
		config: config,
	}
}

// ScoreArticle scores an adverse media article for risk
func (rs *AdverseMediaRiskScorer) ScoreArticle(article AdverseMediaArticle) AdverseMediaArticle {
	rs.logger.Debug("Scoring adverse media article",
		zap.String("article_id", article.ArticleID),
		zap.String("title", article.Title))

	// Calculate risk score based on multiple factors
	riskScore := rs.calculateRiskScore(article)

	// Determine risk level based on score
	riskLevel := rs.determineRiskLevel(riskScore)

	// Determine severity based on risk level and content
	severity := rs.determineSeverity(riskLevel, article)

	// Categorize the article
	category, subcategory := rs.categorizeArticle(article)

	// Analyze sentiment
	sentiment, sentimentScore := rs.analyzeSentiment(article)

	// Update article with calculated values
	article.RiskScore = riskScore
	article.RiskLevel = riskLevel
	article.Severity = severity
	article.Category = category
	article.Subcategory = subcategory
	article.Sentiment = sentiment
	article.SentimentScore = sentimentScore
	article.LastUpdated = time.Now()

	rs.logger.Debug("Article scored",
		zap.String("article_id", article.ArticleID),
		zap.Float64("risk_score", riskScore),
		zap.String("risk_level", riskLevel),
		zap.String("severity", severity))

	return article
}

// calculateRiskScore calculates the overall risk score for an article
func (rs *AdverseMediaRiskScorer) calculateRiskScore(article AdverseMediaArticle) float64 {
	var totalScore float64
	var weightSum float64

	// Factor 1: Title analysis (30% weight)
	titleScore := rs.analyzeTitle(article.Title)
	totalScore += titleScore * 0.3
	weightSum += 0.3

	// Factor 2: Content analysis (40% weight)
	contentScore := rs.analyzeContent(article.Content)
	totalScore += contentScore * 0.4
	weightSum += 0.4

	// Factor 3: Source credibility (10% weight)
	sourceScore := rs.analyzeSource(article.Source)
	totalScore += sourceScore * 0.1
	weightSum += 0.1

	// Factor 4: Recency (10% weight)
	recencyScore := rs.analyzeRecency(article.PublishedDate)
	totalScore += recencyScore * 0.1
	weightSum += 0.1

	// Factor 5: Entity confidence (10% weight)
	confidenceScore := article.EntityConfidence
	totalScore += confidenceScore * 0.1
	weightSum += 0.1

	// Normalize the score
	if weightSum > 0 {
		return totalScore / weightSum
	}
	return 0.0
}

// analyzeTitle analyzes the article title for risk indicators
func (rs *AdverseMediaRiskScorer) analyzeTitle(title string) float64 {
	title = strings.ToLower(title)

	// High-risk keywords
	highRiskKeywords := []string{
		"fraud", "corruption", "bribery", "embezzlement", "money laundering",
		"criminal", "arrest", "indictment", "conviction", "guilty",
		"regulatory violation", "enforcement action", "cease and desist",
		"investigation", "probe", "scandal", "misconduct", "illegal",
	}

	// Medium-risk keywords
	mediumRiskKeywords := []string{
		"fine", "penalty", "sanction", "warning", "violation",
		"compliance", "regulatory", "audit", "review", "concern",
		"allegation", "accusation", "claim", "lawsuit", "litigation",
	}

	// Low-risk keywords
	lowRiskKeywords := []string{
		"announcement", "update", "news", "report", "statement",
		"response", "clarification", "explanation", "comment",
	}

	// Check for high-risk keywords
	for _, keyword := range highRiskKeywords {
		if strings.Contains(title, keyword) {
			return 0.9 // High risk
		}
	}

	// Check for medium-risk keywords
	for _, keyword := range mediumRiskKeywords {
		if strings.Contains(title, keyword) {
			return 0.6 // Medium risk
		}
	}

	// Check for low-risk keywords
	for _, keyword := range lowRiskKeywords {
		if strings.Contains(title, keyword) {
			return 0.3 // Low risk
		}
	}

	// Default risk if no keywords found
	return 0.5
}

// analyzeContent analyzes the article content for risk indicators
func (rs *AdverseMediaRiskScorer) analyzeContent(content string) float64 {
	content = strings.ToLower(content)

	// High-risk phrases
	highRiskPhrases := []string{
		"under investigation", "criminal charges", "regulatory action",
		"enforcement proceeding", "cease and desist order", "civil penalty",
		"criminal conviction", "guilty plea", "sentenced to",
		"money laundering", "securities fraud", "insider trading",
		"accounting fraud", "tax evasion", "bribery scheme",
	}

	// Medium-risk phrases
	mediumRiskPhrases := []string{
		"regulatory review", "compliance issue", "audit findings",
		"internal investigation", "whistleblower", "allegations",
		"legal proceedings", "court case", "settlement",
		"regulatory fine", "penalty imposed", "warning issued",
	}

	// Count occurrences of high-risk phrases
	highRiskCount := 0
	for _, phrase := range highRiskPhrases {
		if strings.Contains(content, phrase) {
			highRiskCount++
		}
	}

	// Count occurrences of medium-risk phrases
	mediumRiskCount := 0
	for _, phrase := range mediumRiskPhrases {
		if strings.Contains(content, phrase) {
			mediumRiskCount++
		}
	}

	// Calculate score based on phrase counts
	if highRiskCount > 0 {
		return math.Min(0.9, 0.5+float64(highRiskCount)*0.2) // 0.7-0.9 for high-risk phrases
	}

	if mediumRiskCount > 0 {
		return math.Min(0.6, 0.3+float64(mediumRiskCount)*0.1) // 0.4-0.6 for medium-risk phrases
	}

	// Default risk if no risk phrases found
	return 0.3
}

// analyzeSource analyzes the credibility and risk profile of the news source
func (rs *AdverseMediaRiskScorer) analyzeSource(source string) float64 {
	// High-credibility sources (lower risk multiplier)
	highCredibilitySources := []string{
		"reuters", "ap news", "bbc news", "financial times",
		"wall street journal", "bloomberg", "new york times",
		"washington post", "guardian", "telegraph",
	}

	// Medium-credibility sources
	mediumCredibilitySources := []string{
		"cnn", "forbes", "business insider", "cnbc", "marketwatch",
		"yahoo finance", "msnbc", "abc news", "cbs news", "nbc news",
	}

	// Low-credibility sources (higher risk multiplier)
	lowCredibilitySources := []string{
		"blog", "rumor", "unverified", "anonymous", "social media",
	}

	source = strings.ToLower(source)

	// Check for high-credibility sources
	for _, credibleSource := range highCredibilitySources {
		if strings.Contains(source, credibleSource) {
			return 0.2 // Low risk multiplier for credible sources
		}
	}

	// Check for medium-credibility sources
	for _, mediumSource := range mediumCredibilitySources {
		if strings.Contains(source, mediumSource) {
			return 0.4 // Medium risk multiplier
		}
	}

	// Check for low-credibility sources
	for _, lowSource := range lowCredibilitySources {
		if strings.Contains(source, lowSource) {
			return 0.8 // High risk multiplier for low-credibility sources
		}
	}

	// Default risk multiplier for unknown sources
	return 0.5
}

// analyzeRecency analyzes how recent the article is (newer = higher risk)
func (rs *AdverseMediaRiskScorer) analyzeRecency(publishedDate time.Time) float64 {
	now := time.Now()
	age := now.Sub(publishedDate)

	// Articles published within the last 24 hours have higher risk
	if age <= 24*time.Hour {
		return 0.9
	}

	// Articles published within the last week have medium-high risk
	if age <= 7*24*time.Hour {
		return 0.7
	}

	// Articles published within the last month have medium risk
	if age <= 30*24*time.Hour {
		return 0.5
	}

	// Older articles have lower risk
	return 0.3
}

// determineRiskLevel determines the risk level based on the calculated score
func (rs *AdverseMediaRiskScorer) determineRiskLevel(riskScore float64) string {
	if riskScore >= rs.config.HighRiskThreshold {
		return "high"
	} else if riskScore >= rs.config.MediumRiskThreshold {
		return "medium"
	} else {
		return "low"
	}
}

// determineSeverity determines the severity based on risk level and content
func (rs *AdverseMediaRiskScorer) determineSeverity(riskLevel string, article AdverseMediaArticle) string {
	// Base severity on risk level
	switch riskLevel {
	case "high":
		// Check for critical indicators
		if rs.hasCriticalIndicators(article) {
			return "critical"
		}
		return "severe"
	case "medium":
		return "moderate"
	default:
		return "minor"
	}
}

// hasCriticalIndicators checks for critical risk indicators
func (rs *AdverseMediaRiskScorer) hasCriticalIndicators(article AdverseMediaArticle) bool {
	content := strings.ToLower(article.Title + " " + article.Content)

	criticalIndicators := []string{
		"criminal conviction", "guilty plea", "sentenced to prison",
		"regulatory shutdown", "cease and desist", "criminal charges",
		"money laundering", "securities fraud", "insider trading",
		"accounting fraud", "tax evasion", "bribery",
	}

	for _, indicator := range criticalIndicators {
		if strings.Contains(content, indicator) {
			return true
		}
	}

	return false
}

// categorizeArticle categorizes the article based on content analysis
func (rs *AdverseMediaRiskScorer) categorizeArticle(article AdverseMediaArticle) (string, string) {
	content := strings.ToLower(article.Title + " " + article.Content)

	// Financial crime category
	financialCrimeKeywords := []string{"fraud", "embezzlement", "money laundering", "securities fraud", "insider trading", "accounting fraud"}
	for _, keyword := range financialCrimeKeywords {
		if strings.Contains(content, keyword) {
			return "financial_crime", rs.getFinancialCrimeSubcategory(content)
		}
	}

	// Corruption category
	corruptionKeywords := []string{"corruption", "bribery", "kickback", "payoff", "graft"}
	for _, keyword := range corruptionKeywords {
		if strings.Contains(content, keyword) {
			return "corruption", rs.getCorruptionSubcategory(content)
		}
	}

	// Regulatory violation category
	regulatoryKeywords := []string{"regulatory", "violation", "compliance", "enforcement", "fine", "penalty"}
	for _, keyword := range regulatoryKeywords {
		if strings.Contains(content, keyword) {
			return "regulatory_violation", rs.getRegulatorySubcategory(content)
		}
	}

	// Litigation category
	litigationKeywords := []string{"lawsuit", "litigation", "court", "legal", "settlement", "judgment"}
	for _, keyword := range litigationKeywords {
		if strings.Contains(content, keyword) {
			return "litigation", rs.getLitigationSubcategory(content)
		}
	}

	// Investigation category
	investigationKeywords := []string{"investigation", "probe", "audit", "review", "inquiry"}
	for _, keyword := range investigationKeywords {
		if strings.Contains(content, keyword) {
			return "investigation", rs.getInvestigationSubcategory(content)
		}
	}

	// Default category
	return "reputational_damage", "general"
}

// getFinancialCrimeSubcategory returns the specific financial crime subcategory
func (rs *AdverseMediaRiskScorer) getFinancialCrimeSubcategory(content string) string {
	if strings.Contains(content, "securities") || strings.Contains(content, "insider trading") {
		return "securities_fraud"
	}
	if strings.Contains(content, "money laundering") {
		return "money_laundering"
	}
	if strings.Contains(content, "accounting") {
		return "accounting_fraud"
	}
	if strings.Contains(content, "tax") {
		return "tax_evasion"
	}
	return "general_fraud"
}

// getCorruptionSubcategory returns the specific corruption subcategory
func (rs *AdverseMediaRiskScorer) getCorruptionSubcategory(content string) string {
	if strings.Contains(content, "bribery") {
		return "bribery"
	}
	if strings.Contains(content, "kickback") {
		return "kickback"
	}
	return "general_corruption"
}

// getRegulatorySubcategory returns the specific regulatory subcategory
func (rs *AdverseMediaRiskScorer) getRegulatorySubcategory(content string) string {
	if strings.Contains(content, "fine") || strings.Contains(content, "penalty") {
		return "regulatory_fine"
	}
	if strings.Contains(content, "cease and desist") {
		return "cease_and_desist"
	}
	return "regulatory_violation"
}

// getLitigationSubcategory returns the specific litigation subcategory
func (rs *AdverseMediaRiskScorer) getLitigationSubcategory(content string) string {
	if strings.Contains(content, "criminal") {
		return "criminal_charges"
	}
	if strings.Contains(content, "civil") {
		return "civil_litigation"
	}
	return "general_litigation"
}

// getInvestigationSubcategory returns the specific investigation subcategory
func (rs *AdverseMediaRiskScorer) getInvestigationSubcategory(content string) string {
	if strings.Contains(content, "internal") {
		return "internal_investigation"
	}
	if strings.Contains(content, "regulatory") {
		return "regulatory_investigation"
	}
	return "general_investigation"
}

// analyzeSentiment analyzes the sentiment of the article
func (rs *AdverseMediaRiskScorer) analyzeSentiment(article AdverseMediaArticle) (string, float64) {
	content := strings.ToLower(article.Title + " " + article.Content)

	// Negative sentiment indicators
	negativeWords := []string{
		"fraud", "corruption", "violation", "illegal", "criminal",
		"guilty", "conviction", "penalty", "fine", "investigation",
		"scandal", "misconduct", "allegation", "accusation", "lawsuit",
		"problem", "issue", "concern", "warning", "threat",
	}

	// Positive sentiment indicators
	positiveWords := []string{
		"cooperation", "compliance", "resolution", "settlement",
		"improvement", "reform", "corrective", "remedial", "good",
		"success", "achievement", "progress", "positive", "beneficial",
	}

	// Count negative words
	negativeCount := 0
	for _, word := range negativeWords {
		if strings.Contains(content, word) {
			negativeCount++
		}
	}

	// Count positive words
	positiveCount := 0
	for _, word := range positiveWords {
		if strings.Contains(content, word) {
			positiveCount++
		}
	}

	// Calculate sentiment score (-1.0 to 1.0)
	totalWords := negativeCount + positiveCount
	if totalWords == 0 {
		return "neutral", 0.0
	}

	sentimentScore := float64(positiveCount-negativeCount) / float64(totalWords)

	// Determine sentiment category
	if sentimentScore > 0.2 {
		return "positive", sentimentScore
	} else if sentimentScore < -0.2 {
		return "negative", sentimentScore
	} else {
		return "neutral", sentimentScore
	}
}

// GetRiskScoreBreakdown provides a detailed breakdown of the risk score calculation
func (rs *AdverseMediaRiskScorer) GetRiskScoreBreakdown(article AdverseMediaArticle) map[string]interface{} {
	breakdown := make(map[string]interface{})

	breakdown["title_score"] = rs.analyzeTitle(article.Title)
	breakdown["content_score"] = rs.analyzeContent(article.Content)
	breakdown["source_score"] = rs.analyzeSource(article.Source)
	breakdown["recency_score"] = rs.analyzeRecency(article.PublishedDate)
	breakdown["confidence_score"] = article.EntityConfidence

	// Calculate weighted total
	totalScore := breakdown["title_score"].(float64)*0.3 +
		breakdown["content_score"].(float64)*0.4 +
		breakdown["source_score"].(float64)*0.1 +
		breakdown["recency_score"].(float64)*0.1 +
		breakdown["confidence_score"].(float64)*0.1

	breakdown["total_score"] = totalScore
	breakdown["risk_level"] = rs.determineRiskLevel(totalScore)
	breakdown["severity"] = rs.determineSeverity(rs.determineRiskLevel(totalScore), article)

	return breakdown
}
