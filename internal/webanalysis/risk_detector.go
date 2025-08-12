package webanalysis

import (
	"context"
	"fmt"
	"strings"
)

// RiskDetector handles risk analysis and detection
type RiskDetector struct {
	riskPatterns              map[string][]RiskPattern
	illegalKeywords           []string
	suspiciousProducts        []string
	moneyLaunderingIndicators []string
	config                    RiskDetectorConfig
}

// RiskPattern represents a risk detection pattern
type RiskPattern struct {
	Pattern     string
	Category    string
	Severity    string
	Confidence  float64
	Description string
}

// RiskDetectorConfig holds configuration for risk detection
type RiskDetectorConfig struct {
	MinConfidence                    float64
	EnableIllegalActivityDetection   bool
	EnableSuspiciousProductDetection bool
	EnableMoneyLaunderingDetection   bool
}

// NewRiskDetector creates a new risk detector
func NewRiskDetector() *RiskDetector {
	rd := &RiskDetector{
		riskPatterns: make(map[string][]RiskPattern),
		config: RiskDetectorConfig{
			MinConfidence:                    0.5,
			EnableIllegalActivityDetection:   true,
			EnableSuspiciousProductDetection: true,
			EnableMoneyLaunderingDetection:   true,
		},
	}

	// Initialize risk patterns
	rd.initializeRiskPatterns()

	return rd
}

// AnalyzeContent performs risk analysis on content
func (rd *RiskDetector) AnalyzeContent(ctx context.Context, content string, businessName string) (*RiskAssessment, error) {
	normalizedContent := strings.ToLower(content)

	var riskFactors []RiskFactor
	var riskIndicators []RiskIndicator

	// Detect illegal activities
	if rd.config.EnableIllegalActivityDetection {
		illegalRisks := rd.detectIllegalActivities(normalizedContent)
		riskFactors = append(riskFactors, illegalRisks...)
	}

	// Detect suspicious products
	if rd.config.EnableSuspiciousProductDetection {
		suspiciousRisks := rd.detectSuspiciousProducts(normalizedContent)
		riskFactors = append(riskFactors, suspiciousRisks...)
	}

	// Detect money laundering indicators
	if rd.config.EnableMoneyLaunderingDetection {
		mlRisks := rd.detectMoneyLaundering(normalizedContent)
		riskFactors = append(riskFactors, mlRisks...)
	}

	// Create risk indicators
	for _, factor := range riskFactors {
		indicator := RiskIndicator{
			Type:        factor.Category,
			Description: factor.Description,
			Confidence:  factor.Confidence,
			Source:      "content_analysis",
		}
		riskIndicators = append(riskIndicators, indicator)
	}

	// Calculate overall risk score
	riskScore := rd.calculateOverallRiskScore(riskFactors)

	// Determine overall risk level
	overallRisk := rd.determineOverallRisk(riskScore)

	// Generate recommendations
	recommendations := rd.generateRecommendations(riskFactors)

	return &RiskAssessment{
		OverallRisk:     overallRisk,
		RiskScore:       riskScore,
		RiskFactors:     riskFactors,
		RiskIndicators:  riskIndicators,
		Recommendations: recommendations,
	}, nil
}

// detectIllegalActivities detects illegal activities in content
func (rd *RiskDetector) detectIllegalActivities(content string) []RiskFactor {
	var risks []RiskFactor

	// Check for illegal activity keywords
	for _, keyword := range rd.illegalKeywords {
		if strings.Contains(content, keyword) {
			risk := RiskFactor{
				Category:    "Illegal Activity",
				Description: fmt.Sprintf("Potential illegal activity detected: %s", keyword),
				Severity:    "High",
				Confidence:  0.7,
				Evidence:    fmt.Sprintf("Found keyword: %s", keyword),
			}
			risks = append(risks, risk)
		}
	}

	// Check for risk patterns
	if patterns, exists := rd.riskPatterns["illegal_activity"]; exists {
		for _, pattern := range patterns {
			if strings.Contains(content, pattern.Pattern) {
				risk := RiskFactor{
					Category:    "Illegal Activity",
					Description: pattern.Description,
					Severity:    pattern.Severity,
					Confidence:  pattern.Confidence,
					Evidence:    fmt.Sprintf("Matched pattern: %s", pattern.Pattern),
				}
				risks = append(risks, risk)
			}
		}
	}

	return risks
}

// detectSuspiciousProducts detects suspicious products in content
func (rd *RiskDetector) detectSuspiciousProducts(content string) []RiskFactor {
	var risks []RiskFactor

	// Check for suspicious product keywords
	for _, product := range rd.suspiciousProducts {
		if strings.Contains(content, product) {
			risk := RiskFactor{
				Category:    "Suspicious Products",
				Description: fmt.Sprintf("Suspicious product detected: %s", product),
				Severity:    "Medium",
				Confidence:  0.6,
				Evidence:    fmt.Sprintf("Found product: %s", product),
			}
			risks = append(risks, risk)
		}
	}

	// Check for risk patterns
	if patterns, exists := rd.riskPatterns["suspicious_products"]; exists {
		for _, pattern := range patterns {
			if strings.Contains(content, pattern.Pattern) {
				risk := RiskFactor{
					Category:    "Suspicious Products",
					Description: pattern.Description,
					Severity:    pattern.Severity,
					Confidence:  pattern.Confidence,
					Evidence:    fmt.Sprintf("Matched pattern: %s", pattern.Pattern),
				}
				risks = append(risks, risk)
			}
		}
	}

	return risks
}

// detectMoneyLaundering detects money laundering indicators
func (rd *RiskDetector) detectMoneyLaundering(content string) []RiskFactor {
	var risks []RiskFactor

	// Check for money laundering indicators
	for _, indicator := range rd.moneyLaunderingIndicators {
		if strings.Contains(content, indicator) {
			risk := RiskFactor{
				Category:    "Money Laundering",
				Description: fmt.Sprintf("Money laundering indicator detected: %s", indicator),
				Severity:    "High",
				Confidence:  0.8,
				Evidence:    fmt.Sprintf("Found indicator: %s", indicator),
			}
			risks = append(risks, risk)
		}
	}

	// Check for risk patterns
	if patterns, exists := rd.riskPatterns["money_laundering"]; exists {
		for _, pattern := range patterns {
			if strings.Contains(content, pattern.Pattern) {
				risk := RiskFactor{
					Category:    "Money Laundering",
					Description: pattern.Description,
					Severity:    pattern.Severity,
					Confidence:  pattern.Confidence,
					Evidence:    fmt.Sprintf("Matched pattern: %s", pattern.Pattern),
				}
				risks = append(risks, risk)
			}
		}
	}

	return risks
}

// calculateOverallRiskScore calculates the overall risk score
func (rd *RiskDetector) calculateOverallRiskScore(riskFactors []RiskFactor) float64 {
	if len(riskFactors) == 0 {
		return 0.0
	}

	totalScore := 0.0
	totalWeight := 0.0

	for _, factor := range riskFactors {
		weight := rd.getSeverityWeight(factor.Severity)
		score := factor.Confidence * weight
		totalScore += score
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

// getSeverityWeight returns the weight for a severity level
func (rd *RiskDetector) getSeverityWeight(severity string) float64 {
	switch strings.ToLower(severity) {
	case "high":
		return 1.0
	case "medium":
		return 0.6
	case "low":
		return 0.3
	default:
		return 0.5
	}
}

// determineOverallRisk determines the overall risk level
func (rd *RiskDetector) determineOverallRisk(riskScore float64) string {
	if riskScore >= 0.8 {
		return "Critical"
	} else if riskScore >= 0.6 {
		return "High"
	} else if riskScore >= 0.4 {
		return "Medium"
	} else if riskScore >= 0.2 {
		return "Low"
	} else {
		return "Minimal"
	}
}

// generateRecommendations generates recommendations based on risk factors
func (rd *RiskDetector) generateRecommendations(riskFactors []RiskFactor) []string {
	var recommendations []string

	if len(riskFactors) == 0 {
		recommendations = append(recommendations, "No significant risks detected")
		return recommendations
	}

	// Add general recommendations
	recommendations = append(recommendations, "Conduct additional due diligence")
	recommendations = append(recommendations, "Review business activities and products")

	// Add specific recommendations based on risk factors
	for _, factor := range riskFactors {
		switch factor.Category {
		case "Illegal Activity":
			recommendations = append(recommendations, "Investigate potential illegal activities")
		case "Suspicious Products":
			recommendations = append(recommendations, "Review product offerings for compliance")
		case "Money Laundering":
			recommendations = append(recommendations, "Conduct enhanced due diligence for AML compliance")
		}
	}

	return recommendations
}

// initializeRiskPatterns initializes risk detection patterns
func (rd *RiskDetector) initializeRiskPatterns() {
	// Illegal activity patterns
	rd.riskPatterns["illegal_activity"] = []RiskPattern{
		{
			Pattern:     "counterfeit",
			Category:    "Illegal Activity",
			Severity:    "High",
			Confidence:  0.8,
			Description: "Potential counterfeit product activity",
		},
		{
			Pattern:     "illegal",
			Category:    "Illegal Activity",
			Severity:    "High",
			Confidence:  0.7,
			Description: "Illegal activity mentioned",
		},
		{
			Pattern:     "fraud",
			Category:    "Illegal Activity",
			Severity:    "High",
			Confidence:  0.9,
			Description: "Fraudulent activity detected",
		},
	}

	// Suspicious product patterns
	rd.riskPatterns["suspicious_products"] = []RiskPattern{
		{
			Pattern:     "weapon",
			Category:    "Suspicious Products",
			Severity:    "High",
			Confidence:  0.8,
			Description: "Weapon-related products detected",
		},
		{
			Pattern:     "drug",
			Category:    "Suspicious Products",
			Severity:    "High",
			Confidence:  0.9,
			Description: "Drug-related products detected",
		},
		{
			Pattern:     "pornography",
			Category:    "Suspicious Products",
			Severity:    "Medium",
			Confidence:  0.7,
			Description: "Adult content detected",
		},
	}

	// Money laundering patterns
	rd.riskPatterns["money_laundering"] = []RiskPattern{
		{
			Pattern:     "anonymous",
			Category:    "Money Laundering",
			Severity:    "Medium",
			Confidence:  0.6,
			Description: "Anonymous transaction indicators",
		},
		{
			Pattern:     "offshore",
			Category:    "Money Laundering",
			Severity:    "High",
			Confidence:  0.7,
			Description: "Offshore account indicators",
		},
		{
			Pattern:     "cash only",
			Category:    "Money Laundering",
			Severity:    "Medium",
			Confidence:  0.5,
			Description: "Cash-only transaction indicators",
		},
	}

	// Initialize keyword lists
	rd.illegalKeywords = []string{
		"counterfeit", "fake", "illegal", "fraud", "scam", "phishing", "hacking",
		"stolen", "theft", "robbery", "extortion", "bribery", "corruption",
	}

	rd.suspiciousProducts = []string{
		"weapon", "gun", "ammunition", "explosive", "drug", "narcotic",
		"pornography", "adult content", "stolen goods", "counterfeit",
	}

	rd.moneyLaunderingIndicators = []string{
		"anonymous", "offshore", "tax haven", "shell company", "cash only",
		"no questions asked", "discrete", "confidential", "untraceable",
	}
}
