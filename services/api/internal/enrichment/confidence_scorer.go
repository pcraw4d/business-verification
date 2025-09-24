package enrichment

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ConfidenceScorer provides advanced confidence scoring for size indicators
type ConfidenceScorer struct {
	config *ConfidenceConfig
	logger *zap.Logger
	tracer trace.Tracer
}

// ConfidenceConfig contains configuration for confidence scoring
type ConfidenceConfig struct {
	// Scoring weights
	DataQualityWeight       float64 `json:"data_quality_weight"`       // Weight for data quality factors
	ConsistencyWeight       float64 `json:"consistency_weight"`        // Weight for data consistency
	ValidationWeight        float64 `json:"validation_weight"`         // Weight for validation status
	EvidenceWeight          float64 `json:"evidence_weight"`           // Weight for evidence quality
	FreshnessWeight         float64 `json:"freshness_weight"`          // Weight for data freshness
	SourceReliabilityWeight float64 `json:"source_reliability_weight"` // Weight for source reliability

	// Thresholds
	MinConfidenceThreshold  float64 `json:"min_confidence_threshold"`
	MaxConfidenceThreshold  float64 `json:"max_confidence_threshold"`
	HighConfidenceThreshold float64 `json:"high_confidence_threshold"`

	// Data quality settings
	EnableDataQualityScoring bool `json:"enable_data_quality_scoring"`
	EnableConsistencyScoring bool `json:"enable_consistency_scoring"`
	EnableValidationScoring  bool `json:"enable_validation_scoring"`
	EnableEvidenceScoring    bool `json:"enable_evidence_scoring"`
	EnableFreshnessScoring   bool `json:"enable_freshness_scoring"`
	EnableSourceScoring      bool `json:"enable_source_scoring"`

	// Calibration settings
	EnableCalibration        bool    `json:"enable_calibration"`
	CalibrationFactor        float64 `json:"calibration_factor"`
	HistoricalAccuracyWeight float64 `json:"historical_accuracy_weight"`

	// Advanced features
	EnableUncertaintyQuantification bool `json:"enable_uncertainty_quantification"`
	EnableConfidenceIntervals       bool `json:"enable_confidence_intervals"`
	EnableAnomalyDetection          bool `json:"enable_anomaly_detection"`
}

// ConfidenceScore contains detailed confidence scoring information
type ConfidenceScore struct {
	// Overall confidence
	OverallConfidence float64 `json:"overall_confidence"` // 0.0-1.0
	ConfidenceLevel   string  `json:"confidence_level"`   // low, medium, high, very_high

	// Component scores
	DataQualityScore       float64 `json:"data_quality_score"`
	ConsistencyScore       float64 `json:"consistency_score"`
	ValidationScore        float64 `json:"validation_score"`
	EvidenceScore          float64 `json:"evidence_score"`
	FreshnessScore         float64 `json:"freshness_score"`
	SourceReliabilityScore float64 `json:"source_reliability_score"`

	// Uncertainty quantification
	UncertaintyLevel   float64            `json:"uncertainty_level"` // 0.0-1.0 (inverse of confidence)
	ConfidenceInterval ConfidenceInterval `json:"confidence_interval"`
	AnomalyScore       float64            `json:"anomaly_score"` // 0.0-1.0 (higher = more anomalous)

	// Detailed breakdown
	ComponentBreakdown map[string]ComponentScore `json:"component_breakdown"`
	Factors            []ConfidenceFactor        `json:"factors"`
	Recommendations    []string                  `json:"recommendations"`

	// Metadata
	CalculatedAt       time.Time `json:"calculated_at"`
	CalibrationApplied bool      `json:"calibration_applied"`
	CalibrationFactor  float64   `json:"calibration_factor"`
	HistoricalAccuracy float64   `json:"historical_accuracy"`
}

// ConfidenceInterval represents confidence interval bounds
type ConfidenceInterval struct {
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Level      float64 `json:"level"` // 0.95 for 95% confidence interval
}

// ComponentScore represents individual component scoring
type ComponentScore struct {
	Score         float64  `json:"score"`
	Weight        float64  `json:"weight"`
	WeightedScore float64  `json:"weighted_score"`
	Factors       []string `json:"factors"`
	Issues        []string `json:"issues"`
}

// ConfidenceFactor represents individual confidence factors
type ConfidenceFactor struct {
	Factor      string  `json:"factor"`
	Impact      float64 `json:"impact"` // -1.0 to 1.0
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
}

// EnhancedDataQualityMetrics extends DataQualityMetrics with confidence-specific metrics
type EnhancedDataQualityMetrics struct {
	DataQualityMetrics

	// Confidence-specific metrics
	SourceReliability float64 `json:"source_reliability"` // 0.0-1.0
	DataFreshness     float64 `json:"data_freshness"`     // 0.0-1.0
	ValidationStatus  float64 `json:"validation_status"`  // 0.0-1.0
	EvidenceQuality   float64 `json:"evidence_quality"`   // 0.0-1.0
}

// NewConfidenceScorer creates a new confidence scorer with default configuration
func NewConfidenceScorer(config *ConfidenceConfig, logger *zap.Logger) *ConfidenceScorer {
	if config == nil {
		config = &ConfidenceConfig{
			// Default weights
			DataQualityWeight:       0.25,
			ConsistencyWeight:       0.20,
			ValidationWeight:        0.15,
			EvidenceWeight:          0.20,
			FreshnessWeight:         0.10,
			SourceReliabilityWeight: 0.10,

			// Default thresholds
			MinConfidenceThreshold:  0.3,
			MaxConfidenceThreshold:  1.0,
			HighConfidenceThreshold: 0.8,

			// Enable all scoring components
			EnableDataQualityScoring: true,
			EnableConsistencyScoring: true,
			EnableValidationScoring:  true,
			EnableEvidenceScoring:    true,
			EnableFreshnessScoring:   true,
			EnableSourceScoring:      true,

			// Calibration settings
			EnableCalibration:        true,
			CalibrationFactor:        1.0,
			HistoricalAccuracyWeight: 0.1,

			// Advanced features
			EnableUncertaintyQuantification: true,
			EnableConfidenceIntervals:       true,
			EnableAnomalyDetection:          true,
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &ConfidenceScorer{
		config: config,
		logger: logger,
		tracer: otel.Tracer("confidence-scorer"),
	}
}

// CalculateConfidence calculates comprehensive confidence score for size indicators
func (cs *ConfidenceScorer) CalculateConfidence(ctx context.Context, result *CompanySizeResult) (*ConfidenceScore, error) {
	ctx, span := cs.tracer.Start(ctx, "confidence_scorer.calculate",
		trace.WithAttributes(
			attribute.String("company_size", result.CompanySize),
			attribute.Float64("base_confidence", result.ConfidenceScore),
		))
	defer span.End()

	cs.logger.Info("Starting confidence calculation",
		zap.String("company_size", result.CompanySize),
		zap.Float64("base_confidence", result.ConfidenceScore))

	confidenceScore := &ConfidenceScore{
		CalculatedAt:       time.Now(),
		ComponentBreakdown: make(map[string]ComponentScore),
		Factors:            []ConfidenceFactor{},
		Recommendations:    []string{},
	}

	// Calculate component scores
	if cs.config.EnableDataQualityScoring {
		confidenceScore.DataQualityScore = cs.calculateDataQualityScore(result)
		confidenceScore.ComponentBreakdown["data_quality"] = ComponentScore{
			Score:         confidenceScore.DataQualityScore,
			Weight:        cs.config.DataQualityWeight,
			WeightedScore: confidenceScore.DataQualityScore * cs.config.DataQualityWeight,
			Factors:       cs.getDataQualityFactors(result),
		}
	}

	if cs.config.EnableConsistencyScoring {
		confidenceScore.ConsistencyScore = cs.calculateConsistencyScore(result)
		confidenceScore.ComponentBreakdown["consistency"] = ComponentScore{
			Score:         confidenceScore.ConsistencyScore,
			Weight:        cs.config.ConsistencyWeight,
			WeightedScore: confidenceScore.ConsistencyScore * cs.config.ConsistencyWeight,
			Factors:       cs.getConsistencyFactors(result),
		}
	}

	if cs.config.EnableValidationScoring {
		confidenceScore.ValidationScore = cs.calculateValidationScore(result)
		confidenceScore.ComponentBreakdown["validation"] = ComponentScore{
			Score:         confidenceScore.ValidationScore,
			Weight:        cs.config.ValidationWeight,
			WeightedScore: confidenceScore.ValidationScore * cs.config.ValidationWeight,
			Factors:       cs.getValidationFactors(result),
		}
	}

	if cs.config.EnableEvidenceScoring {
		confidenceScore.EvidenceScore = cs.calculateEvidenceScore(result)
		confidenceScore.ComponentBreakdown["evidence"] = ComponentScore{
			Score:         confidenceScore.EvidenceScore,
			Weight:        cs.config.EvidenceWeight,
			WeightedScore: confidenceScore.EvidenceScore * cs.config.EvidenceWeight,
			Factors:       cs.getEvidenceFactors(result),
		}
	}

	if cs.config.EnableFreshnessScoring {
		confidenceScore.FreshnessScore = cs.calculateFreshnessScore(result)
		confidenceScore.ComponentBreakdown["freshness"] = ComponentScore{
			Score:         confidenceScore.FreshnessScore,
			Weight:        cs.config.FreshnessWeight,
			WeightedScore: confidenceScore.FreshnessScore * cs.config.FreshnessWeight,
			Factors:       cs.getFreshnessFactors(result),
		}
	}

	if cs.config.EnableSourceScoring {
		confidenceScore.SourceReliabilityScore = cs.calculateSourceReliabilityScore(result)
		confidenceScore.ComponentBreakdown["source_reliability"] = ComponentScore{
			Score:         confidenceScore.SourceReliabilityScore,
			Weight:        cs.config.SourceReliabilityWeight,
			WeightedScore: confidenceScore.SourceReliabilityScore * cs.config.SourceReliabilityWeight,
			Factors:       cs.getSourceReliabilityFactors(result),
		}
	}

	// Calculate overall confidence
	confidenceScore.OverallConfidence = cs.calculateOverallConfidence(confidenceScore)
	confidenceScore.ConfidenceLevel = cs.determineConfidenceLevel(confidenceScore.OverallConfidence)

	// Calculate uncertainty quantification
	if cs.config.EnableUncertaintyQuantification {
		confidenceScore.UncertaintyLevel = 1.0 - confidenceScore.OverallConfidence
	}

	// Calculate confidence intervals
	if cs.config.EnableConfidenceIntervals {
		confidenceScore.ConfidenceInterval = cs.calculateConfidenceInterval(confidenceScore)
	}

	// Detect anomalies
	if cs.config.EnableAnomalyDetection {
		confidenceScore.AnomalyScore = cs.calculateAnomalyScore(result, confidenceScore)
	}

	// Apply calibration if enabled
	if cs.config.EnableCalibration {
		cs.applyCalibration(confidenceScore)
	}

	// Generate recommendations
	confidenceScore.Recommendations = cs.generateRecommendations(confidenceScore, result)

	// Aggregate all factors
	confidenceScore.Factors = cs.aggregateFactors(confidenceScore)

	cs.logger.Info("Confidence calculation completed",
		zap.Float64("overall_confidence", confidenceScore.OverallConfidence),
		zap.String("confidence_level", confidenceScore.ConfidenceLevel),
		zap.Int("factor_count", len(confidenceScore.Factors)))

	return confidenceScore, nil
}

// calculateDataQualityScore calculates data quality component score
func (cs *ConfidenceScorer) calculateDataQualityScore(result *CompanySizeResult) float64 {
	score := 0.0
	factorCount := 0

	// Employee data quality
	if result.EmployeeAnalysis != nil {
		employeeQuality := cs.assessEmployeeDataQuality(result.EmployeeAnalysis)
		score += employeeQuality
		factorCount++
	}

	// Revenue data quality
	if result.RevenueAnalysis != nil {
		revenueQuality := cs.assessRevenueDataQuality(result.RevenueAnalysis)
		score += revenueQuality
		factorCount++
	}

	// Overall data quality metrics
	if result.DataQualityScore > 0 {
		score += result.DataQualityScore
		factorCount++
	}

	if factorCount == 0 {
		return 0.0
	}

	return score / float64(factorCount)
}

// calculateConsistencyScore calculates consistency component score
func (cs *ConfidenceScorer) calculateConsistencyScore(result *CompanySizeResult) float64 {
	// Base consistency from the result
	score := result.ConsistencyScore

	// Additional consistency factors
	hasEmployeeData := result.EmployeeAnalysis != nil && result.EmployeeAnalysis.EmployeeCount > 0
	hasRevenueData := result.RevenueAnalysis != nil && result.RevenueAnalysis.RevenueAmount > 0

	// Consistency bonus for having both data types
	if hasEmployeeData && hasRevenueData {
		score += 0.1
	}

	// Consistency penalty for conflicting classifications
	if hasEmployeeData && hasRevenueData {
		if result.EmployeeClassification != result.RevenueClassification {
			score -= 0.2
		}
	}

	// Evidence consistency
	if len(result.Evidence) > 1 {
		evidenceConsistency := cs.assessEvidenceConsistency(result.Evidence)
		score = (score + evidenceConsistency) / 2.0
	}

	return math.Max(0.0, math.Min(1.0, score))
}

// calculateValidationScore calculates validation component score
func (cs *ConfidenceScorer) calculateValidationScore(result *CompanySizeResult) float64 {
	score := 0.0
	factorCount := 0

	// Validation status
	if result.IsValidated {
		score += 0.8
		factorCount++
	}

	// Employee validation
	if result.EmployeeAnalysis != nil && result.EmployeeAnalysis.IsValidated {
		score += 0.9
		factorCount++
	}

	// Revenue validation
	if result.RevenueAnalysis != nil && result.RevenueAnalysis.IsValidated {
		score += 0.9
		factorCount++
	}

	// Validation status quality
	if result.ValidationStatus.IsValid {
		score += 1.0
		factorCount++
	} else if len(result.ValidationStatus.ValidationErrors) == 0 {
		score += 0.6
		factorCount++
	} else {
		score += 0.2
		factorCount++
	}

	if factorCount == 0 {
		return 0.5 // Default moderate validation score
	}

	return score / float64(factorCount)
}

// calculateEvidenceScore calculates evidence component score
func (cs *ConfidenceScorer) calculateEvidenceScore(result *CompanySizeResult) float64 {
	score := 0.0

	// Evidence quantity
	evidenceCount := len(result.Evidence)
	if evidenceCount == 0 {
		return 0.1
	} else if evidenceCount == 1 {
		score += 0.3
	} else if evidenceCount <= 3 {
		score += 0.6
	} else if evidenceCount <= 5 {
		score += 0.8
	} else {
		score += 0.9
	}

	// Evidence quality
	evidenceQuality := cs.assessEvidenceQuality(result.Evidence)
	score = (score + evidenceQuality) / 2.0

	// Evidence diversity
	evidenceDiversity := cs.assessEvidenceDiversity(result.Evidence)
	score = (score + evidenceDiversity) / 2.0

	return score
}

// calculateFreshnessScore calculates data freshness component score
func (cs *ConfidenceScorer) calculateFreshnessScore(result *CompanySizeResult) float64 {
	score := 0.5 // Default moderate freshness

	// Check extraction timestamps
	if !result.ClassifiedAt.IsZero() {
		age := time.Since(result.ClassifiedAt)
		if age < 24*time.Hour {
			score = 1.0
		} else if age < 7*24*time.Hour {
			score = 0.9
		} else if age < 30*24*time.Hour {
			score = 0.7
		} else if age < 90*24*time.Hour {
			score = 0.5
		} else {
			score = 0.3
		}
	}

	// Check employee data freshness
	if result.EmployeeAnalysis != nil && !result.EmployeeAnalysis.ExtractedAt.IsZero() {
		employeeAge := time.Since(result.EmployeeAnalysis.ExtractedAt)
		employeeFreshness := cs.calculateTimeBasedScore(employeeAge)
		score = (score + employeeFreshness) / 2.0
	}

	// Check revenue data freshness
	if result.RevenueAnalysis != nil && !result.RevenueAnalysis.ExtractedAt.IsZero() {
		revenueAge := time.Since(result.RevenueAnalysis.ExtractedAt)
		revenueFreshness := cs.calculateTimeBasedScore(revenueAge)
		score = (score + revenueFreshness) / 2.0
	}

	return score
}

// calculateSourceReliabilityScore calculates source reliability component score
func (cs *ConfidenceScorer) calculateSourceReliabilityScore(result *CompanySizeResult) float64 {
	score := 0.5 // Default moderate reliability

	// Source URL reliability
	if result.SourceURL != "" {
		urlReliability := cs.assessURLReliability(result.SourceURL)
		score = (score + urlReliability) / 2.0
	}

	// Employee source reliability
	if result.EmployeeAnalysis != nil && result.EmployeeAnalysis.SourceURL != "" {
		employeeSourceReliability := cs.assessURLReliability(result.EmployeeAnalysis.SourceURL)
		score = (score + employeeSourceReliability) / 2.0
	}

	// Revenue source reliability
	if result.RevenueAnalysis != nil && result.RevenueAnalysis.SourceURL != "" {
		revenueSourceReliability := cs.assessURLReliability(result.RevenueAnalysis.SourceURL)
		score = (score + revenueSourceReliability) / 2.0
	}

	return score
}

// calculateOverallConfidence calculates the overall confidence score
func (cs *ConfidenceScorer) calculateOverallConfidence(confidenceScore *ConfidenceScore) float64 {
	totalWeightedScore := 0.0
	totalWeight := 0.0

	// Sum weighted scores from all components
	for _, component := range confidenceScore.ComponentBreakdown {
		totalWeightedScore += component.WeightedScore
		totalWeight += component.Weight
	}

	if totalWeight == 0 {
		return 0.5 // Default moderate confidence
	}

	overallConfidence := totalWeightedScore / totalWeight

	// Apply bounds
	overallConfidence = math.Max(cs.config.MinConfidenceThreshold,
		math.Min(cs.config.MaxConfidenceThreshold, overallConfidence))

	return overallConfidence
}

// determineConfidenceLevel determines the confidence level category
func (cs *ConfidenceScorer) determineConfidenceLevel(confidence float64) string {
	if confidence >= 0.9 {
		return "very_high"
	} else if confidence >= 0.8 {
		return "high"
	} else if confidence >= 0.6 {
		return "medium"
	} else {
		return "low"
	}
}

// calculateConfidenceInterval calculates confidence interval bounds
func (cs *ConfidenceScorer) calculateConfidenceInterval(confidenceScore *ConfidenceScore) ConfidenceInterval {
	// Calculate standard error based on uncertainty
	standardError := confidenceScore.UncertaintyLevel * 0.5

	// 95% confidence interval (1.96 * standard error)
	margin := 1.96 * standardError

	lowerBound := math.Max(0.0, confidenceScore.OverallConfidence-margin)
	upperBound := math.Min(1.0, confidenceScore.OverallConfidence+margin)

	return ConfidenceInterval{
		LowerBound: lowerBound,
		UpperBound: upperBound,
		Level:      0.95,
	}
}

// calculateAnomalyScore calculates anomaly detection score
func (cs *ConfidenceScorer) calculateAnomalyScore(result *CompanySizeResult, confidenceScore *ConfidenceScore) float64 {
	anomalyScore := 0.0

	// Check for unusual confidence patterns
	if confidenceScore.OverallConfidence > 0.95 && len(result.Evidence) < 2 {
		anomalyScore += 0.3 // Suspiciously high confidence with little evidence
	}

	if confidenceScore.OverallConfidence < 0.3 && len(result.Evidence) > 5 {
		anomalyScore += 0.2 // Suspiciously low confidence with much evidence
	}

	// Check for data inconsistencies
	if result.EmployeeAnalysis != nil && result.RevenueAnalysis != nil {
		employeeCount := result.EmployeeAnalysis.EmployeeCount
		revenueAmount := result.RevenueAnalysis.RevenueAmount

		// Unusual employee-to-revenue ratios
		if employeeCount > 0 && revenueAmount > 0 {
			revenuePerEmployee := float64(revenueAmount) / float64(employeeCount)
			if revenuePerEmployee > 1000000 { // $1M per employee
				anomalyScore += 0.4
			} else if revenuePerEmployee < 10000 { // $10K per employee
				anomalyScore += 0.3
			}
		}
	}

	// Check for classification inconsistencies
	if result.EmployeeClassification != result.RevenueClassification {
		anomalyScore += 0.2
	}

	return math.Min(1.0, anomalyScore)
}

// applyCalibration applies calibration to the confidence score
func (cs *ConfidenceScorer) applyCalibration(confidenceScore *ConfidenceScore) {
	if cs.config.CalibrationFactor != 1.0 {
		originalConfidence := confidenceScore.OverallConfidence
		calibratedConfidence := originalConfidence * cs.config.CalibrationFactor

		// Ensure calibrated confidence stays within bounds
		calibratedConfidence = math.Max(cs.config.MinConfidenceThreshold,
			math.Min(cs.config.MaxConfidenceThreshold, calibratedConfidence))

		confidenceScore.OverallConfidence = calibratedConfidence
		confidenceScore.ConfidenceLevel = cs.determineConfidenceLevel(calibratedConfidence)
		confidenceScore.CalibrationApplied = true
		confidenceScore.CalibrationFactor = cs.config.CalibrationFactor
	}
}

// generateRecommendations generates improvement recommendations
func (cs *ConfidenceScorer) generateRecommendations(confidenceScore *ConfidenceScore, result *CompanySizeResult) []string {
	recommendations := []string{}

	// Low confidence recommendations
	if confidenceScore.OverallConfidence < 0.5 {
		recommendations = append(recommendations, "Consider collecting additional data sources")
		recommendations = append(recommendations, "Validate extracted information with external sources")
	}

	// Data quality recommendations
	if confidenceScore.DataQualityScore < 0.6 {
		recommendations = append(recommendations, "Improve data quality through better extraction methods")
		recommendations = append(recommendations, "Implement data validation checks")
	}

	// Consistency recommendations
	if confidenceScore.ConsistencyScore < 0.7 {
		recommendations = append(recommendations, "Resolve inconsistencies between employee and revenue data")
		recommendations = append(recommendations, "Cross-validate information across multiple sources")
	}

	// Evidence recommendations
	if confidenceScore.EvidenceScore < 0.5 {
		recommendations = append(recommendations, "Collect more evidence to support classification")
		recommendations = append(recommendations, "Diversify evidence sources")
	}

	// Freshness recommendations
	if confidenceScore.FreshnessScore < 0.6 {
		recommendations = append(recommendations, "Update data to ensure freshness")
		recommendations = append(recommendations, "Implement regular data refresh cycles")
	}

	// Anomaly recommendations
	if confidenceScore.AnomalyScore > 0.5 {
		recommendations = append(recommendations, "Investigate potential data anomalies")
		recommendations = append(recommendations, "Verify unusual patterns with additional sources")
	}

	return recommendations
}

// Helper methods for component scoring

func (cs *ConfidenceScorer) assessEmployeeDataQuality(analysis *EmployeeCountResult) float64 {
	score := 0.5

	// Extraction method quality
	switch analysis.ExtractionMethod {
	case "direct_mention":
		score += 0.3
	case "linkedin_style":
		score += 0.2
	case "size_keyword":
		score += 0.1
	}

	// Confidence score
	score += analysis.ConfidenceScore * 0.2

	// Validation status
	if analysis.IsValidated {
		score += 0.1
	}

	return math.Min(1.0, score)
}

func (cs *ConfidenceScorer) assessRevenueDataQuality(analysis *RevenueResult) float64 {
	score := 0.5

	// Extraction method quality
	switch analysis.ExtractionMethod {
	case "direct_mention":
		score += 0.3
	case "revenue_range":
		score += 0.2
	case "financial_indicator":
		score += 0.1
	}

	// Confidence score
	score += analysis.ConfidenceScore * 0.2

	// Validation status
	if analysis.IsValidated {
		score += 0.1
	}

	return math.Min(1.0, score)
}

func (cs *ConfidenceScorer) assessEvidenceConsistency(evidence []string) float64 {
	if len(evidence) < 2 {
		return 0.5
	}

	// Simple consistency check - count unique evidence types
	evidenceTypes := make(map[string]int)
	for _, e := range evidence {
		// Extract evidence type from evidence string
		if strings.Contains(e, "employee") {
			evidenceTypes["employee"]++
		} else if strings.Contains(e, "revenue") {
			evidenceTypes["revenue"]++
		} else if strings.Contains(e, "financial") {
			evidenceTypes["financial"]++
		}
	}

	// More diverse evidence types = higher consistency
	diversity := float64(len(evidenceTypes)) / float64(len(evidence))
	return math.Min(1.0, diversity+0.3)
}

func (cs *ConfidenceScorer) assessEvidenceQuality(evidence []string) float64 {
	if len(evidence) == 0 {
		return 0.0
	}

	score := 0.0
	for _, e := range evidence {
		// Assess individual evidence quality
		if strings.Contains(e, "direct") {
			score += 0.9
		} else if strings.Contains(e, "mention") {
			score += 0.7
		} else if strings.Contains(e, "indicator") {
			score += 0.5
		} else {
			score += 0.3
		}
	}

	return score / float64(len(evidence))
}

func (cs *ConfidenceScorer) assessEvidenceDiversity(evidence []string) float64 {
	if len(evidence) < 2 {
		return 0.3
	}

	// Count unique evidence patterns
	patterns := make(map[string]bool)
	for _, e := range evidence {
		// Extract pattern from evidence
		if strings.Contains(e, "employee") {
			patterns["employee"] = true
		} else if strings.Contains(e, "revenue") {
			patterns["revenue"] = true
		} else if strings.Contains(e, "financial") {
			patterns["financial"] = true
		} else {
			patterns["other"] = true
		}
	}

	return float64(len(patterns)) / 4.0 // Normalize to 0-1 range
}

func (cs *ConfidenceScorer) calculateTimeBasedScore(age time.Duration) float64 {
	if age < 24*time.Hour {
		return 1.0
	} else if age < 7*24*time.Hour {
		return 0.9
	} else if age < 30*24*time.Hour {
		return 0.7
	} else if age < 90*24*time.Hour {
		return 0.5
	} else {
		return 0.3
	}
}

func (cs *ConfidenceScorer) assessURLReliability(url string) float64 {
	// Simple URL reliability assessment
	if strings.Contains(url, "linkedin.com") {
		return 0.9
	} else if strings.Contains(url, "crunchbase.com") {
		return 0.8
	} else if strings.Contains(url, "company.com") || strings.Contains(url, "corp.com") {
		return 0.7
	} else if strings.Contains(url, "https://") {
		return 0.6
	} else {
		return 0.4
	}
}

// Factor generation methods

func (cs *ConfidenceScorer) getDataQualityFactors(result *CompanySizeResult) []string {
	factors := []string{}

	if result.EmployeeAnalysis != nil {
		factors = append(factors, "employee_data_quality")
	}
	if result.RevenueAnalysis != nil {
		factors = append(factors, "revenue_data_quality")
	}
	if result.DataQualityScore > 0.7 {
		factors = append(factors, "high_overall_quality")
	}

	return factors
}

func (cs *ConfidenceScorer) getConsistencyFactors(result *CompanySizeResult) []string {
	factors := []string{}

	if result.ConsistencyScore > 0.8 {
		factors = append(factors, "high_consistency")
	} else if result.ConsistencyScore < 0.5 {
		factors = append(factors, "low_consistency")
	}

	if result.EmployeeAnalysis != nil && result.RevenueAnalysis != nil {
		factors = append(factors, "dual_data_sources")
	}

	return factors
}

func (cs *ConfidenceScorer) getValidationFactors(result *CompanySizeResult) []string {
	factors := []string{}

	if result.IsValidated {
		factors = append(factors, "result_validated")
	}
	if result.EmployeeAnalysis != nil && result.EmployeeAnalysis.IsValidated {
		factors = append(factors, "employee_validated")
	}
	if result.RevenueAnalysis != nil && result.RevenueAnalysis.IsValidated {
		factors = append(factors, "revenue_validated")
	}

	return factors
}

func (cs *ConfidenceScorer) getEvidenceFactors(result *CompanySizeResult) []string {
	factors := []string{}

	evidenceCount := len(result.Evidence)
	if evidenceCount > 3 {
		factors = append(factors, "multiple_evidence")
	} else if evidenceCount > 0 {
		factors = append(factors, "some_evidence")
	} else {
		factors = append(factors, "no_evidence")
	}

	return factors
}

func (cs *ConfidenceScorer) getFreshnessFactors(result *CompanySizeResult) []string {
	factors := []string{}

	if !result.ClassifiedAt.IsZero() {
		age := time.Since(result.ClassifiedAt)
		if age < 24*time.Hour {
			factors = append(factors, "very_recent")
		} else if age < 7*24*time.Hour {
			factors = append(factors, "recent")
		} else {
			factors = append(factors, "older_data")
		}
	}

	return factors
}

func (cs *ConfidenceScorer) getSourceReliabilityFactors(result *CompanySizeResult) []string {
	factors := []string{}

	if result.SourceURL != "" {
		if strings.Contains(result.SourceURL, "linkedin.com") {
			factors = append(factors, "reliable_source")
		} else if strings.Contains(result.SourceURL, "https://") {
			factors = append(factors, "secure_source")
		}
	}

	return factors
}

func (cs *ConfidenceScorer) aggregateFactors(confidenceScore *ConfidenceScore) []ConfidenceFactor {
	factors := []ConfidenceFactor{}

	// Aggregate factors from all components
	for componentName, component := range confidenceScore.ComponentBreakdown {
		for _, factor := range component.Factors {
			factors = append(factors, ConfidenceFactor{
				Factor:      factor,
				Impact:      component.Score - 0.5, // -0.5 to 0.5 range
				Weight:      component.Weight,
				Description: fmt.Sprintf("%s factor: %s", componentName, factor),
				Category:    componentName,
			})
		}
	}

	return factors
}
