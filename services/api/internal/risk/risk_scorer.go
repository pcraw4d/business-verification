package risk

import (
	"go.uber.org/zap"
)

// RiskScorer provides risk scoring capabilities for detected keywords and patterns
type RiskScorer struct {
	logger *zap.Logger
	config *RiskDetectionConfig
}

// NewRiskScorer creates a new risk scorer
func NewRiskScorer(logger *zap.Logger, config *RiskDetectionConfig) *RiskScorer {
	return &RiskScorer{
		logger: logger,
		config: config,
	}
}

// CalculateRiskScore calculates the overall risk score from detected keywords
func (rs *RiskScorer) CalculateRiskScore(keywords []DetectedRiskKeyword) (float64, RiskLevel) {
	if len(keywords) == 0 {
		return 0.0, RiskLevelMinimal
	}

	var totalScore float64
	var totalWeight float64

	// Weight different risk categories
	categoryWeights := map[string]float64{
		"illegal":    1.0,  // Highest weight for illegal activities
		"sanctions":  0.95, // Very high weight for sanctions violations
		"prohibited": 0.8,  // High weight for prohibited activities
		"tbml":       0.7,  // High weight for trade-based money laundering
		"fraud":      0.6,  // Medium-high weight for fraud
		"high_risk":  0.5,  // Medium weight for high-risk activities
	}

	// Weight different severity levels
	severityWeights := map[string]float64{
		"critical": 1.0,
		"high":     0.8,
		"medium":   0.6,
		"low":      0.4,
	}

	// Calculate weighted score
	for _, keyword := range keywords {
		categoryWeight := categoryWeights[keyword.Category]
		if categoryWeight == 0 {
			categoryWeight = 0.3 // Default weight for unknown categories
		}

		severityWeight := severityWeights[keyword.Severity]
		if severityWeight == 0 {
			severityWeight = 0.5 // Default weight for unknown severity
		}

		// Calculate individual keyword score
		keywordScore := keyword.Confidence * categoryWeight * severityWeight
		keywordWeight := categoryWeight * severityWeight

		totalScore += keywordScore
		totalWeight += keywordWeight
	}

	// Normalize score
	var overallScore float64
	if totalWeight > 0 {
		overallScore = totalScore / totalWeight
	}

	// Apply amplification for multiple detections
	if len(keywords) > 1 {
		amplificationFactor := 1.0 + (float64(len(keywords)-1) * 0.1)
		overallScore *= amplificationFactor
		if overallScore > 1.0 {
			overallScore = 1.0
		}
	}

	// Determine risk level
	riskLevel := rs.determineRiskLevel(overallScore)

	return overallScore, riskLevel
}

// CalculatePatternRiskScore calculates risk score from detected patterns
func (rs *RiskScorer) CalculatePatternRiskScore(patterns []DetectedPattern) (float64, RiskLevel) {
	if len(patterns) == 0 {
		return 0.0, RiskLevelMinimal
	}

	var totalScore float64
	var totalWeight float64

	// Weight different pattern types
	patternTypeWeights := map[string]float64{
		"money_laundering":    0.9,
		"fraud_pattern":       0.8,
		"shell_company":       0.7,
		"sanctions_evasion":   0.95,
		"terrorist_financing": 1.0,
		"drug_trafficking":    1.0,
		"weapons_trafficking": 1.0,
		"human_trafficking":   1.0,
		"cybercrime":          0.6,
		"suspicious_activity": 0.5,
	}

	// Calculate weighted score
	for _, pattern := range patterns {
		patternWeight := patternTypeWeights[pattern.PatternType]
		if patternWeight == 0 {
			patternWeight = 0.4 // Default weight for unknown pattern types
		}

		patternScore := pattern.Confidence * patternWeight
		totalScore += patternScore
		totalWeight += patternWeight
	}

	// Normalize score
	var overallScore float64
	if totalWeight > 0 {
		overallScore = totalScore / totalWeight
	}

	// Apply amplification for multiple patterns
	if len(patterns) > 1 {
		amplificationFactor := 1.0 + (float64(len(patterns)-1) * 0.15)
		overallScore *= amplificationFactor
		if overallScore > 1.0 {
			overallScore = 1.0
		}
	}

	// Determine risk level
	riskLevel := rs.determineRiskLevel(overallScore)

	return overallScore, riskLevel
}

// determineRiskLevel determines the risk level based on score
func (rs *RiskScorer) determineRiskLevel(score float64) RiskLevel {
	switch {
	case score >= rs.config.CriticalRiskThreshold:
		return RiskLevelCritical
	case score >= rs.config.HighRiskThreshold:
		return RiskLevelHigh
	case score >= 0.5:
		return RiskLevelMedium
	case score >= 0.2:
		return RiskLevelLow
	default:
		return RiskLevelMinimal
	}
}

// CalculateCategoryRiskScore calculates risk score for a specific category
func (rs *RiskScorer) CalculateCategoryRiskScore(
	category string,
	keywords []DetectedRiskKeyword,
) (float64, RiskLevel) {
	if len(keywords) == 0 {
		return 0.0, RiskLevelMinimal
	}

	// Filter keywords by category
	var categoryKeywords []DetectedRiskKeyword
	for _, keyword := range keywords {
		if keyword.Category == category {
			categoryKeywords = append(categoryKeywords, keyword)
		}
	}

	if len(categoryKeywords) == 0 {
		return 0.0, RiskLevelMinimal
	}

	// Calculate score for this category
	score, level := rs.CalculateRiskScore(categoryKeywords)
	return score, level
}

// CalculateSeverityRiskScore calculates risk score for a specific severity level
func (rs *RiskScorer) CalculateSeverityRiskScore(
	severity string,
	keywords []DetectedRiskKeyword,
) (float64, RiskLevel) {
	if len(keywords) == 0 {
		return 0.0, RiskLevelMinimal
	}

	// Filter keywords by severity
	var severityKeywords []DetectedRiskKeyword
	for _, keyword := range keywords {
		if keyword.Severity == severity {
			severityKeywords = append(severityKeywords, keyword)
		}
	}

	if len(severityKeywords) == 0 {
		return 0.0, RiskLevelMinimal
	}

	// Calculate score for this severity
	score, level := rs.CalculateRiskScore(severityKeywords)
	return score, level
}

// CalculateSourceRiskScore calculates risk score for a specific source
func (rs *RiskScorer) CalculateSourceRiskScore(
	source string,
	keywords []DetectedRiskKeyword,
) (float64, RiskLevel) {
	if len(keywords) == 0 {
		return 0.0, RiskLevelMinimal
	}

	// Filter keywords by source
	var sourceKeywords []DetectedRiskKeyword
	for _, keyword := range keywords {
		if keyword.Source == source {
			sourceKeywords = append(sourceKeywords, keyword)
		}
	}

	if len(sourceKeywords) == 0 {
		return 0.0, RiskLevelMinimal
	}

	// Calculate score for this source
	score, level := rs.CalculateRiskScore(sourceKeywords)
	return score, level
}

// GetRiskScoreBreakdown provides a detailed breakdown of risk scoring
func (rs *RiskScorer) GetRiskScoreBreakdown(keywords []DetectedRiskKeyword) map[string]interface{} {
	breakdown := make(map[string]interface{})

	// Overall score
	overallScore, overallLevel := rs.CalculateRiskScore(keywords)
	breakdown["overall_score"] = overallScore
	breakdown["overall_level"] = string(overallLevel)

	// Category breakdown
	categoryScores := make(map[string]float64)
	categoryLevels := make(map[string]string)
	categories := []string{"illegal", "prohibited", "high_risk", "tbml", "sanctions", "fraud"}

	for _, category := range categories {
		score, level := rs.CalculateCategoryRiskScore(category, keywords)
		categoryScores[category] = score
		categoryLevels[category] = string(level)
	}
	breakdown["category_scores"] = categoryScores
	breakdown["category_levels"] = categoryLevels

	// Severity breakdown
	severityScores := make(map[string]float64)
	severityLevels := make(map[string]string)
	severities := []string{"critical", "high", "medium", "low"}

	for _, severity := range severities {
		score, level := rs.CalculateSeverityRiskScore(severity, keywords)
		severityScores[severity] = score
		severityLevels[severity] = string(level)
	}
	breakdown["severity_scores"] = severityScores
	breakdown["severity_levels"] = severityLevels

	// Source breakdown
	sourceScores := make(map[string]float64)
	sourceLevels := make(map[string]string)
	sources := []string{"business_name", "description", "website_content", "mcc_code", "naics_code", "sic_code", "pattern"}

	for _, source := range sources {
		score, level := rs.CalculateSourceRiskScore(source, keywords)
		sourceScores[source] = score
		sourceLevels[source] = string(level)
	}
	breakdown["source_scores"] = sourceScores
	breakdown["source_levels"] = sourceLevels

	// Statistics
	breakdown["total_keywords"] = len(keywords)
	breakdown["unique_categories"] = rs.getUniqueCategories(keywords)
	breakdown["unique_severities"] = rs.getUniqueSeverities(keywords)
	breakdown["unique_sources"] = rs.getUniqueSources(keywords)

	return breakdown
}

// Helper methods for breakdown

func (rs *RiskScorer) getUniqueCategories(keywords []DetectedRiskKeyword) []string {
	categoryMap := make(map[string]bool)
	for _, keyword := range keywords {
		categoryMap[keyword.Category] = true
	}

	var categories []string
	for category := range categoryMap {
		categories = append(categories, category)
	}
	return categories
}

func (rs *RiskScorer) getUniqueSeverities(keywords []DetectedRiskKeyword) []string {
	severityMap := make(map[string]bool)
	for _, keyword := range keywords {
		severityMap[keyword.Severity] = true
	}

	var severities []string
	for severity := range severityMap {
		severities = append(severities, severity)
	}
	return severities
}

func (rs *RiskScorer) getUniqueSources(keywords []DetectedRiskKeyword) []string {
	sourceMap := make(map[string]bool)
	for _, keyword := range keywords {
		sourceMap[keyword.Source] = true
	}

	var sources []string
	for source := range sourceMap {
		sources = append(sources, source)
	}
	return sources
}
