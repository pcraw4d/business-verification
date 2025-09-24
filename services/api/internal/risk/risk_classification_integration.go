package risk

import (
	"context"
	"fmt"
	"time"

	"kyb-platform/internal/classification"
	"kyb-platform/internal/shared"

	"go.uber.org/zap"
)

// RiskClassificationIntegration provides integration between risk detection and classification
type RiskClassificationIntegration struct {
	riskDetectionService  *RiskDetectionService
	multiMethodClassifier *classification.MultiMethodClassifier
	logger                *zap.Logger
	config                *RiskClassificationConfig
}

// RiskClassificationConfig contains configuration for risk-classification integration
type RiskClassificationConfig struct {
	EnableRiskDetection        bool    `json:"enable_risk_detection"`
	EnableWebsiteAnalysis      bool    `json:"enable_website_analysis"`
	EnableContentAnalysis      bool    `json:"enable_content_analysis"`
	EnablePatternDetection     bool    `json:"enable_pattern_detection"`
	RiskWeightInClassification float64 `json:"risk_weight_in_classification"`
	MinRiskThreshold           float64 `json:"min_risk_threshold"`
	HighRiskThreshold          float64 `json:"high_risk_threshold"`
	CriticalRiskThreshold      float64 `json:"critical_risk_threshold"`
}

// DefaultRiskClassificationConfig returns default configuration
func DefaultRiskClassificationConfig() *RiskClassificationConfig {
	return &RiskClassificationConfig{
		EnableRiskDetection:        true,
		EnableWebsiteAnalysis:      true,
		EnableContentAnalysis:      true,
		EnablePatternDetection:     true,
		RiskWeightInClassification: 0.3, // 30% weight for risk in classification
		MinRiskThreshold:           0.3,
		HighRiskThreshold:          0.7,
		CriticalRiskThreshold:      0.9,
	}
}

// EnhancedClassificationResult represents the result of classification with risk assessment
type EnhancedClassificationResult struct {
	Classification             *shared.IndustryClassification `json:"classification"`
	RiskAssessment             *EnhancedRiskDetectionResult   `json:"risk_assessment"`
	CombinedScore              float64                        `json:"combined_score"`
	RiskAdjustedClassification *shared.IndustryClassification `json:"risk_adjusted_classification"`
	Recommendations            []RiskRecommendation           `json:"recommendations"`
	Alerts                     []RiskAlert                    `json:"alerts"`
	ProcessingTime             time.Duration                  `json:"processing_time"`
	Metadata                   map[string]interface{}         `json:"metadata"`
}

// NewRiskClassificationIntegration creates a new risk-classification integration service
func NewRiskClassificationIntegration(
	riskDetectionService *RiskDetectionService,
	multiMethodClassifier *classification.MultiMethodClassifier,
	logger *zap.Logger,
	config *RiskClassificationConfig,
) *RiskClassificationIntegration {
	if config == nil {
		config = DefaultRiskClassificationConfig()
	}

	return &RiskClassificationIntegration{
		riskDetectionService:  riskDetectionService,
		multiMethodClassifier: multiMethodClassifier,
		logger:                logger,
		config:                config,
	}
}

// ClassifyWithRiskAssessment performs classification with integrated risk assessment
func (rci *RiskClassificationIntegration) ClassifyWithRiskAssessment(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*EnhancedClassificationResult, error) {
	startTime := time.Now()
	requestID := generateRequestID()

	rci.logger.Info("Starting enhanced classification with risk assessment",
		zap.String("request_id", requestID),
		zap.String("business_name", businessName))

	// Initialize result
	result := &EnhancedClassificationResult{
		Metadata: make(map[string]interface{}),
	}

	// Step 1: Perform standard classification
	classificationResult, err := rci.multiMethodClassifier.ClassifyWithMultipleMethods(
		ctx, businessName, description, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	result.Classification = classificationResult.Classifications[0] // Get the best classification

	// Step 2: Perform risk assessment if enabled
	if rci.config.EnableRiskDetection {
		riskRequest := &RiskDetectionRequest{
			BusinessName:            businessName,
			BusinessDescription:     description,
			WebsiteURL:              websiteURL,
			IncludeWebsiteAnalysis:  rci.config.EnableWebsiteAnalysis,
			IncludeContentAnalysis:  rci.config.EnableContentAnalysis,
			IncludePatternDetection: rci.config.EnablePatternDetection,
		}

		// Add classification codes to risk request if available
		if result.Classification.Metadata != nil {
			if mccCode, ok := result.Classification.Metadata["mcc_code"].(string); ok {
				riskRequest.MCCCode = mccCode
			}
			if naicsCode, ok := result.Classification.Metadata["naics_code"].(string); ok {
				riskRequest.NAICSCode = naicsCode
			}
			if sicCode, ok := result.Classification.Metadata["sic_code"].(string); ok {
				riskRequest.SICCode = sicCode
			}
		}

		riskAssessment, err := rci.riskDetectionService.DetectRisk(ctx, riskRequest)
		if err != nil {
			rci.logger.Warn("Risk assessment failed", zap.Error(err))
			// Continue without risk assessment
		} else {
			result.RiskAssessment = riskAssessment
			result.Recommendations = riskAssessment.Recommendations
			result.Alerts = riskAssessment.Alerts
		}
	}

	// Step 3: Calculate combined score
	result.CombinedScore = rci.calculateCombinedScore(result.Classification, result.RiskAssessment)

	// Step 4: Generate risk-adjusted classification
	result.RiskAdjustedClassification = rci.generateRiskAdjustedClassification(
		result.Classification, result.RiskAssessment)

	// Step 5: Set processing time
	result.ProcessingTime = time.Since(startTime)

	// Step 6: Add metadata
	result.Metadata["request_id"] = requestID
	result.Metadata["classification_method"] = result.Classification.ClassificationMethod
	result.Metadata["risk_detection_enabled"] = rci.config.EnableRiskDetection
	if result.RiskAssessment != nil {
		result.Metadata["risk_score"] = result.RiskAssessment.OverallRiskScore
		result.Metadata["risk_level"] = string(result.RiskAssessment.OverallRiskLevel)
		result.Metadata["detected_keywords"] = len(result.RiskAssessment.DetectedKeywords)
	}

	rci.logger.Info("Enhanced classification completed",
		zap.String("request_id", requestID),
		zap.Float64("classification_confidence", result.Classification.ConfidenceScore),
		zap.Float64("combined_score", result.CombinedScore),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// calculateCombinedScore calculates a combined score from classification and risk assessment
func (rci *RiskClassificationIntegration) calculateCombinedScore(
	classification *shared.IndustryClassification,
	riskAssessment *EnhancedRiskDetectionResult,
) float64 {
	// Base classification confidence
	classificationScore := classification.ConfidenceScore

	// If no risk assessment, return classification score
	if riskAssessment == nil {
		return classificationScore
	}

	// Risk-adjusted score
	riskScore := riskAssessment.OverallRiskScore
	riskWeight := rci.config.RiskWeightInClassification

	// Calculate combined score
	// Higher risk reduces the overall confidence
	riskAdjustment := 1.0 - (riskScore * riskWeight)
	combinedScore := classificationScore * riskAdjustment

	// Ensure score is within bounds
	if combinedScore < 0.0 {
		combinedScore = 0.0
	}
	if combinedScore > 1.0 {
		combinedScore = 1.0
	}

	return combinedScore
}

// generateRiskAdjustedClassification generates a risk-adjusted classification result
func (rci *RiskClassificationIntegration) generateRiskAdjustedClassification(
	originalClassification *shared.IndustryClassification,
	riskAssessment *EnhancedRiskDetectionResult,
) *shared.IndustryClassification {
	// Start with a copy of the original classification
	adjustedClassification := &shared.IndustryClassification{
		IndustryCode:         originalClassification.IndustryCode,
		IndustryName:         originalClassification.IndustryName,
		ConfidenceScore:      rci.calculateCombinedScore(originalClassification, riskAssessment),
		ClassificationMethod: originalClassification.ClassificationMethod + "_risk_adjusted",
		Description:          originalClassification.Description,
		Evidence:             originalClassification.Evidence,
		ProcessingTime:       originalClassification.ProcessingTime,
		Metadata:             make(map[string]interface{}),
	}

	// Copy original metadata
	for k, v := range originalClassification.Metadata {
		adjustedClassification.Metadata[k] = v
	}

	// Add risk-related metadata
	if riskAssessment != nil {
		adjustedClassification.Metadata["risk_score"] = riskAssessment.OverallRiskScore
		adjustedClassification.Metadata["risk_level"] = string(riskAssessment.OverallRiskLevel)
		adjustedClassification.Metadata["risk_keywords_count"] = len(riskAssessment.DetectedKeywords)
		adjustedClassification.Metadata["risk_categories"] = rci.getRiskCategories(riskAssessment)
		adjustedClassification.Metadata["risk_adjustment_factor"] = 1.0 - (riskAssessment.OverallRiskScore * rci.config.RiskWeightInClassification)
	}

	// Adjust description based on risk level
	if riskAssessment != nil {
		riskDescription := rci.generateRiskAdjustedDescription(originalClassification, riskAssessment)
		adjustedClassification.Description = riskDescription
	}

	return adjustedClassification
}

// generateRiskAdjustedDescription generates a description that includes risk information
func (rci *RiskClassificationIntegration) generateRiskAdjustedDescription(
	originalClassification *shared.IndustryClassification,
	riskAssessment *EnhancedRiskDetectionResult,
) string {
	baseDescription := originalClassification.Description

	if riskAssessment == nil {
		return baseDescription
	}

	riskLevel := riskAssessment.OverallRiskLevel
	riskScore := riskAssessment.OverallRiskScore

	var riskNote string
	switch riskLevel {
	case RiskLevelCritical:
		riskNote = fmt.Sprintf(" CRITICAL RISK DETECTED (Score: %.2f) - Immediate review required.", riskScore)
	case RiskLevelHigh:
		riskNote = fmt.Sprintf(" HIGH RISK DETECTED (Score: %.2f) - Enhanced monitoring required.", riskScore)
	case RiskLevelMedium:
		riskNote = fmt.Sprintf(" MEDIUM RISK DETECTED (Score: %.2f) - Standard monitoring.", riskScore)
	case RiskLevelLow:
		riskNote = fmt.Sprintf(" LOW RISK DETECTED (Score: %.2f) - Minimal monitoring.", riskScore)
	default:
		riskNote = fmt.Sprintf(" MINIMAL RISK DETECTED (Score: %.2f) - Standard processing.", riskScore)
	}

	return baseDescription + riskNote
}

// getRiskCategories extracts risk categories from risk assessment
func (rci *RiskClassificationIntegration) getRiskCategories(riskAssessment *EnhancedRiskDetectionResult) []string {
	if riskAssessment == nil {
		return []string{}
	}

	var categories []string
	for category := range riskAssessment.RiskCategories {
		categories = append(categories, category)
	}

	return categories
}

// GetRiskAdjustedRecommendations generates recommendations based on combined classification and risk
func (rci *RiskClassificationIntegration) GetRiskAdjustedRecommendations(
	result *EnhancedClassificationResult,
) []RiskRecommendation {
	var recommendations []RiskRecommendation

	// Add existing risk recommendations
	recommendations = append(recommendations, result.Recommendations...)

	// Add classification-based recommendations
	if result.Classification.ConfidenceScore < 0.7 {
		recommendations = append(recommendations, RiskRecommendation{
			ID:          generateRequestID(),
			RiskFactor:  "classification_confidence",
			Title:       "Low Classification Confidence",
			Description: "Classification confidence is below recommended threshold. Additional verification may be required.",
			Priority:    RiskLevelMedium,
			Action:      "Manual review of classification",
			Impact:      "Medium - Additional verification required",
			Timeline:    "Within 48 hours",
			CreatedAt:   time.Now(),
		})
	}

	// Add combined score recommendations
	if result.CombinedScore < 0.5 {
		recommendations = append(recommendations, RiskRecommendation{
			ID:          generateRequestID(),
			RiskFactor:  "combined_score",
			Title:       "Low Combined Score",
			Description: "Combined classification and risk score is below threshold. Comprehensive review recommended.",
			Priority:    RiskLevelHigh,
			Action:      "Comprehensive business review",
			Impact:      "High - Business approval may be delayed",
			Timeline:    "Within 24 hours",
			CreatedAt:   time.Now(),
		})
	}

	return recommendations
}

// ValidateRiskClassificationResult validates the result of risk-classification integration
func (rci *RiskClassificationIntegration) ValidateRiskClassificationResult(
	result *EnhancedClassificationResult,
) error {
	// Validate classification
	if result.Classification == nil {
		return fmt.Errorf("classification result is nil")
	}

	if result.Classification.ConfidenceScore < 0.0 || result.Classification.ConfidenceScore > 1.0 {
		return fmt.Errorf("invalid classification confidence score: %f", result.Classification.ConfidenceScore)
	}

	// Validate combined score
	if result.CombinedScore < 0.0 || result.CombinedScore > 1.0 {
		return fmt.Errorf("invalid combined score: %f", result.CombinedScore)
	}

	// Validate risk assessment if present
	if result.RiskAssessment != nil {
		if result.RiskAssessment.OverallRiskScore < 0.0 || result.RiskAssessment.OverallRiskScore > 1.0 {
			return fmt.Errorf("invalid risk score: %f", result.RiskAssessment.OverallRiskScore)
		}
	}

	return nil
}
