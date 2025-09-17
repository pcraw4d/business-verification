package feedback

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// MLRecommendationEngine generates recommendations for ML model improvements
type MLRecommendationEngine struct {
	config *MLAnalysisConfig
	logger *zap.Logger
}

// NewMLRecommendationEngine creates a new ML recommendation engine
func NewMLRecommendationEngine(config *MLAnalysisConfig, logger *zap.Logger) *MLRecommendationEngine {
	return &MLRecommendationEngine{
		config: config,
		logger: logger,
	}
}

// GenerateModelRecommendations generates recommendations for ML model improvements
func (mre *MLRecommendationEngine) GenerateModelRecommendations(ctx context.Context, analysis *MLAnalysisResult) ([]*MLModelRecommendation, error) {
	var recommendations []*MLRecommendation

	// Generate recommendations based on different analysis components
	patternRecs := mre.generatePatternBasedRecommendations(analysis.FeedbackPatterns)
	recommendations = append(recommendations, patternRecs...)

	misclassificationRecs := mre.generateMisclassificationBasedRecommendations(analysis.Misclassifications)
	recommendations = append(recommendations, misclassificationRecs...)

	driftRecs := mre.generateDriftBasedRecommendations(analysis.MLModelDrift)
	recommendations = append(recommendations, driftRecs...)

	disagreementRecs := mre.generateDisagreementBasedRecommendations(analysis.EnsembleDisagreements)
	recommendations = append(recommendations, disagreementRecs...)

	performanceRecs := mre.generatePerformanceBasedRecommendations(analysis.ModelPerformanceTrends)
	recommendations = append(recommendations, performanceRecs...)

	securityRecs := mre.generateSecurityBasedRecommendations(analysis.SecurityAnalysis)
	recommendations = append(recommendations, securityRecs...)

	// Convert to MLModelRecommendation format
	var mlRecommendations []*MLModelRecommendation
	for _, rec := range recommendations {
		mlRec := mre.convertToMLModelRecommendation(rec, analysis)
		if mlRec != nil {
			mlRecommendations = append(mlRecommendations, mlRec)
		}
	}

	// Sort by priority and expected improvement
	sort.Slice(mlRecommendations, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		if priorityOrder[mlRecommendations[i].Priority] == priorityOrder[mlRecommendations[j].Priority] {
			return mlRecommendations[i].ExpectedImprovement > mlRecommendations[j].ExpectedImprovement
		}
		return priorityOrder[mlRecommendations[i].Priority] > priorityOrder[mlRecommendations[j].Priority]
	})

	return mlRecommendations, nil
}

// generatePatternBasedRecommendations generates recommendations based on feedback patterns
func (mre *MLRecommendationEngine) generatePatternBasedRecommendations(patterns []*FeedbackPattern) []*MLRecommendation {
	var recommendations []*MLRecommendation

	for _, pattern := range patterns {
		rec := mre.analyzePatternForRecommendations(pattern)
		if rec != nil {
			recommendations = append(recommendations, rec)
		}
	}

	return recommendations
}

// analyzePatternForRecommendations analyzes a pattern and generates recommendations
func (mre *MLRecommendationEngine) analyzePatternForRecommendations(pattern *FeedbackPattern) *MLRecommendation {
	switch pattern.PatternType {
	case "confidence_low":
		return &MLRecommendation{
			Type:                "confidence_calibration",
			Priority:            "high",
			Description:         "Improve confidence calibration for low-confidence predictions",
			ExpectedImprovement: 0.15,
			ImplementationCost:  "medium",
			Timeline:            "2-4 weeks",
			Evidence:            []string{fmt.Sprintf("Low confidence pattern detected: %.2f%% of predictions below 0.5", float64(pattern.OccurrenceCount)/float64(pattern.OccurrenceCount)*100)},
		}

	case "confidence_calibration":
		return &MLRecommendation{
			Type:                "confidence_calibration",
			Priority:            "high",
			Description:         "Implement confidence calibration techniques",
			ExpectedImprovement: 0.20,
			ImplementationCost:  "medium",
			Timeline:            "3-5 weeks",
			Evidence:            []string{fmt.Sprintf("Confidence calibration issue: average confidence %.2f", pattern.Metadata["average_confidence"])},
		}

	case "high_corrections":
		return &MLRecommendation{
			Type:                "model_retraining",
			Priority:            "high",
			Description:         "Retrain models with corrected data",
			ExpectedImprovement: 0.25,
			ImplementationCost:  "high",
			Timeline:            "4-6 weeks",
			Evidence:            []string{fmt.Sprintf("High correction rate: %.2f%%", pattern.Metadata["correction_rate"])},
		}

	case "low_accuracy":
		return &MLRecommendation{
			Type:                "model_improvement",
			Priority:            "critical",
			Description:         "Comprehensive model improvement and retraining",
			ExpectedImprovement: 0.30,
			ImplementationCost:  "high",
			Timeline:            "6-8 weeks",
			Evidence:            []string{fmt.Sprintf("Low accuracy rate: %.2f%%", pattern.Metadata["accuracy_rate"])},
		}

	case "common_error":
		return &MLRecommendation{
			Type:                "error_specific_improvement",
			Priority:            "medium",
			Description:         fmt.Sprintf("Address specific error type: %s", pattern.Metadata["error_type"]),
			ExpectedImprovement: 0.10,
			ImplementationCost:  "low",
			Timeline:            "1-2 weeks",
			Evidence:            []string{fmt.Sprintf("Common error type: %s with %d occurrences", pattern.Metadata["error_type"], pattern.Metadata["error_count"])},
		}

	case "performance_gap":
		return &MLRecommendation{
			Type:                "ensemble_optimization",
			Priority:            "medium",
			Description:         "Optimize ensemble weights to address performance gaps",
			ExpectedImprovement: 0.12,
			ImplementationCost:  "low",
			Timeline:            "1-2 weeks",
			Evidence:            []string{fmt.Sprintf("Performance gap between %s and %s: %.2f", pattern.Metadata["method1"], pattern.Metadata["method2"], pattern.Metadata["accuracy_gap"])},
		}

	case "ensemble_disagreement":
		return &MLRecommendation{
			Type:                "ensemble_consensus",
			Priority:            "medium",
			Description:         "Implement better ensemble consensus mechanisms",
			ExpectedImprovement: 0.08,
			ImplementationCost:  "medium",
			Timeline:            "2-3 weeks",
			Evidence:            []string{fmt.Sprintf("High ensemble disagreement: %d feedback items", pattern.Metadata["disagreement_count"])},
		}

	case "temporal_peak":
		return &MLRecommendation{
			Type:                "performance_optimization",
			Priority:            "low",
			Description:         "Optimize performance during peak usage hours",
			ExpectedImprovement: 0.05,
			ImplementationCost:  "low",
			Timeline:            "1 week",
			Evidence:            []string{fmt.Sprintf("Peak activity during hours: %v", pattern.Metadata["peak_hours"])},
		}

	case "daily_trend":
		return &MLRecommendation{
			Type:                "trend_monitoring",
			Priority:            "low",
			Description:         "Implement trend monitoring and adaptive adjustments",
			ExpectedImprovement: 0.03,
			ImplementationCost:  "low",
			Timeline:            "1 week",
			Evidence:            []string{fmt.Sprintf("Daily trend: %s", pattern.Metadata["trend"])},
		}
	}

	return nil
}

// generateMisclassificationBasedRecommendations generates recommendations based on misclassifications
func (mre *MLRecommendationEngine) generateMisclassificationBasedRecommendations(misclassifications []*ModelMisclassification) []*MLRecommendation {
	var recommendations []*MLRecommendation

	for _, misclassification := range misclassifications {
		rec := mre.analyzeMisclassificationForRecommendations(misclassification)
		if rec != nil {
			recommendations = append(recommendations, rec)
		}
	}

	return recommendations
}

// analyzeMisclassificationForRecommendations analyzes a misclassification and generates recommendations
func (mre *MLRecommendationEngine) analyzeMisclassificationForRecommendations(misclassification *ModelMisclassification) *MLRecommendation {
	// Determine priority based on frequency and confidence
	priority := "low"
	if misclassification.Frequency > 50 || misclassification.Confidence > 0.8 {
		priority = "high"
	} else if misclassification.Frequency > 20 || misclassification.Confidence > 0.6 {
		priority = "medium"
	}

	// Determine expected improvement based on misclassification type
	expectedImprovement := 0.10
	switch misclassification.MisclassificationType {
	case "industry_misclassification":
		expectedImprovement = 0.20
	case "confidence_miscalibration":
		expectedImprovement = 0.15
	case "security_validation_failure":
		expectedImprovement = 0.25
	case "ml_model_failure":
		expectedImprovement = 0.18
	case "ensemble_disagreement":
		expectedImprovement = 0.12
	}

	// Determine implementation cost and timeline
	implementationCost := "medium"
	timeline := "2-4 weeks"
	if misclassification.MisclassificationType == "ml_model_failure" {
		implementationCost = "high"
		timeline = "4-6 weeks"
	} else if misclassification.MisclassificationType == "security_validation_failure" {
		implementationCost = "high"
		timeline = "3-5 weeks"
	}

	// Generate evidence
	evidence := []string{
		fmt.Sprintf("Misclassification type: %s", misclassification.MisclassificationType),
		fmt.Sprintf("Frequency: %d occurrences", misclassification.Frequency),
		fmt.Sprintf("Confidence: %.2f", misclassification.Confidence),
	}
	if len(misclassification.AffectedIndustries) > 0 {
		evidence = append(evidence, fmt.Sprintf("Affected industries: %s", strings.Join(misclassification.AffectedIndustries, ", ")))
	}

	return &MLRecommendation{
		Type:                "misclassification_specific",
		Priority:            priority,
		Description:         fmt.Sprintf("Address %s misclassifications in %s model", misclassification.MisclassificationType, misclassification.ModelType),
		ExpectedImprovement: expectedImprovement,
		ImplementationCost:  implementationCost,
		Timeline:            timeline,
		Evidence:            evidence,
	}
}

// generateDriftBasedRecommendations generates recommendations based on model drift
func (mre *MLRecommendationEngine) generateDriftBasedRecommendations(driftAnalysis *MLModelDriftAnalysis) []*MLRecommendation {
	var recommendations []*MLRecommendation

	if driftAnalysis == nil || !driftAnalysis.DriftDetected {
		return recommendations
	}

	// Determine priority based on drift score
	priority := "low"
	if driftAnalysis.DriftScore > 0.3 {
		priority = "high"
	} else if driftAnalysis.DriftScore > 0.2 {
		priority = "medium"
	}

	// Determine expected improvement
	expectedImprovement := driftAnalysis.DriftScore * 0.8 // Assume 80% of drift can be addressed

	// Determine implementation cost and timeline
	implementationCost := "high"
	timeline := "4-6 weeks"

	evidence := []string{
		fmt.Sprintf("Model drift detected: %.2f", driftAnalysis.DriftScore),
		fmt.Sprintf("Drift type: %s", driftAnalysis.DriftType),
	}
	if len(driftAnalysis.AffectedFeatures) > 0 {
		evidence = append(evidence, fmt.Sprintf("Affected features: %s", strings.Join(driftAnalysis.AffectedFeatures, ", ")))
	}

	recommendation := &MLRecommendation{
		Type:                "model_drift_correction",
		Priority:            priority,
		Description:         "Address model drift through retraining and data updates",
		ExpectedImprovement: expectedImprovement,
		ImplementationCost:  implementationCost,
		Timeline:            timeline,
		Evidence:            evidence,
	}

	recommendations = append(recommendations, recommendation)

	// Add specific recommendations based on drift type
	switch driftAnalysis.DriftType {
	case "concept_drift":
		recommendations = append(recommendations, &MLRecommendation{
			Type:                "concept_drift_handling",
			Priority:            "medium",
			Description:         "Implement concept drift detection and adaptation mechanisms",
			ExpectedImprovement: 0.10,
			ImplementationCost:  "medium",
			Timeline:            "2-3 weeks",
			Evidence:            []string{"Concept drift detected in model behavior"},
		})
	case "data_drift":
		recommendations = append(recommendations, &MLRecommendation{
			Type:                "data_drift_handling",
			Priority:            "medium",
			Description:         "Update training data to address data distribution changes",
			ExpectedImprovement: 0.12,
			ImplementationCost:  "medium",
			Timeline:            "3-4 weeks",
			Evidence:            []string{"Data drift detected in input distribution"},
		})
	}

	return recommendations
}

// generateDisagreementBasedRecommendations generates recommendations based on ensemble disagreements
func (mre *MLRecommendationEngine) generateDisagreementBasedRecommendations(disagreements []*EnsembleDisagreement) []*MLRecommendation {
	var recommendations []*MLRecommendation

	if len(disagreements) == 0 {
		return recommendations
	}

	// Calculate overall disagreement score
	var totalDisagreement float64
	for _, disagreement := range disagreements {
		totalDisagreement += disagreement.DisagreementScore
	}
	avgDisagreement := totalDisagreement / float64(len(disagreements))

	// Determine priority based on disagreement level
	priority := "low"
	if avgDisagreement > 0.5 {
		priority = "high"
	} else if avgDisagreement > 0.3 {
		priority = "medium"
	}

	// Determine expected improvement
	expectedImprovement := avgDisagreement * 0.6 // Assume 60% of disagreements can be resolved

	evidence := []string{
		fmt.Sprintf("Average disagreement score: %.2f", avgDisagreement),
		fmt.Sprintf("Number of disagreement types: %d", len(disagreements)),
	}

	// Add specific disagreement types to evidence
	disagreementTypes := make(map[string]int)
	for _, disagreement := range disagreements {
		disagreementTypes[disagreement.DisagreementType]++
	}
	for disagreementType, count := range disagreementTypes {
		evidence = append(evidence, fmt.Sprintf("%s disagreements: %d", disagreementType, count))
	}

	recommendation := &MLRecommendation{
		Type:                "ensemble_consensus_improvement",
		Priority:            priority,
		Description:         "Improve ensemble consensus mechanisms to reduce disagreements",
		ExpectedImprovement: expectedImprovement,
		ImplementationCost:  "medium",
		Timeline:            "2-4 weeks",
		Evidence:            evidence,
	}

	recommendations = append(recommendations, recommendation)

	// Add specific recommendations for different disagreement types
	for disagreementType, count := range disagreementTypes {
		if count > len(disagreements)/2 {
			recommendations = append(recommendations, &MLRecommendation{
				Type:                fmt.Sprintf("disagreement_%s_resolution", disagreementType),
				Priority:            "medium",
				Description:         fmt.Sprintf("Address %s disagreements in ensemble", disagreementType),
				ExpectedImprovement: 0.08,
				ImplementationCost:  "low",
				Timeline:            "1-2 weeks",
				Evidence:            []string{fmt.Sprintf("%s disagreements: %d occurrences", disagreementType, count)},
			})
		}
	}

	return recommendations
}

// generatePerformanceBasedRecommendations generates recommendations based on performance trends
func (mre *MLRecommendationEngine) generatePerformanceBasedRecommendations(trends []*ModelPerformanceTrend) []*MLRecommendation {
	var recommendations []*MLRecommendation

	for _, trend := range trends {
		rec := mre.analyzePerformanceTrendForRecommendations(trend)
		if rec != nil {
			recommendations = append(recommendations, rec)
		}
	}

	return recommendations
}

// analyzePerformanceTrendForRecommendations analyzes a performance trend and generates recommendations
func (mre *MLRecommendationEngine) analyzePerformanceTrendForRecommendations(trend *ModelPerformanceTrend) *MLRecommendation {
	// Determine priority based on trend direction and strength
	priority := "low"
	if trend.TrendDirection == "decreasing" && trend.TrendStrength > 0.7 {
		priority = "high"
	} else if trend.TrendDirection == "decreasing" && trend.TrendStrength > 0.4 {
		priority = "medium"
	}

	// Determine expected improvement
	expectedImprovement := 0.05
	if trend.TrendDirection == "decreasing" {
		expectedImprovement = trend.TrendStrength * 0.3 // Assume 30% of trend can be reversed
	}

	// Determine implementation cost and timeline
	implementationCost := "medium"
	timeline := "2-3 weeks"
	if trend.TrendType == "accuracy" && trend.TrendDirection == "decreasing" {
		implementationCost = "high"
		timeline = "4-6 weeks"
	}

	evidence := []string{
		fmt.Sprintf("Performance trend: %s %s", trend.TrendType, trend.TrendDirection),
		fmt.Sprintf("Trend strength: %.2f", trend.TrendStrength),
		fmt.Sprintf("Time window: %s", trend.TimeWindow),
	}

	if trend.Projection != nil {
		evidence = append(evidence, fmt.Sprintf("Projected %s: %.2f", trend.TrendType, trend.Projection.ProjectedAccuracy))
	}

	return &MLRecommendation{
		Type:                fmt.Sprintf("performance_trend_%s", trend.TrendType),
		Priority:            priority,
		Description:         fmt.Sprintf("Address %s %s trend in model performance", trend.TrendDirection, trend.TrendType),
		ExpectedImprovement: expectedImprovement,
		ImplementationCost:  implementationCost,
		Timeline:            timeline,
		Evidence:            evidence,
	}
}

// generateSecurityBasedRecommendations generates recommendations based on security analysis
func (mre *MLRecommendationEngine) generateSecurityBasedRecommendations(securityAnalysis *SecurityFeedbackAnalysis) []*MLRecommendation {
	var recommendations []*MLRecommendation

	if securityAnalysis == nil {
		return recommendations
	}

	// Determine priority based on security score
	priority := "low"
	if securityAnalysis.OverallSecurityScore < 0.7 {
		priority = "high"
	} else if securityAnalysis.OverallSecurityScore < 0.8 {
		priority = "medium"
	}

	// Determine expected improvement
	expectedImprovement := (1.0 - securityAnalysis.OverallSecurityScore) * 0.8

	evidence := []string{
		fmt.Sprintf("Overall security score: %.2f", securityAnalysis.OverallSecurityScore),
		fmt.Sprintf("Security violations: %d", len(securityAnalysis.SecurityViolations)),
		fmt.Sprintf("Trusted source issues: %d", len(securityAnalysis.TrustedSourceIssues)),
		fmt.Sprintf("Website verification issues: %d", len(securityAnalysis.WebsiteVerificationIssues)),
	}

	recommendation := &MLRecommendation{
		Type:                "security_improvement",
		Priority:            priority,
		Description:         "Improve security validation and trusted data source handling",
		ExpectedImprovement: expectedImprovement,
		ImplementationCost:  "high",
		Timeline:            "3-5 weeks",
		Evidence:            evidence,
	}

	recommendations = append(recommendations, recommendation)

	// Add specific security recommendations
	if len(securityAnalysis.SecurityViolations) > 0 {
		recommendations = append(recommendations, &MLRecommendation{
			Type:                "security_violation_handling",
			Priority:            "high",
			Description:         "Implement better security violation detection and handling",
			ExpectedImprovement: 0.15,
			ImplementationCost:  "medium",
			Timeline:            "2-3 weeks",
			Evidence:            []string{fmt.Sprintf("Security violations detected: %d", len(securityAnalysis.SecurityViolations))},
		})
	}

	if len(securityAnalysis.TrustedSourceIssues) > 0 {
		recommendations = append(recommendations, &MLRecommendation{
			Type:                "trusted_source_improvement",
			Priority:            "medium",
			Description:         "Improve trusted data source validation and handling",
			ExpectedImprovement: 0.10,
			ImplementationCost:  "medium",
			Timeline:            "2-3 weeks",
			Evidence:            []string{fmt.Sprintf("Trusted source issues: %d", len(securityAnalysis.TrustedSourceIssues))},
		})
	}

	if len(securityAnalysis.WebsiteVerificationIssues) > 0 {
		recommendations = append(recommendations, &MLRecommendation{
			Type:                "website_verification_improvement",
			Priority:            "medium",
			Description:         "Enhance website verification processes",
			ExpectedImprovement: 0.08,
			ImplementationCost:  "low",
			Timeline:            "1-2 weeks",
			Evidence:            []string{fmt.Sprintf("Website verification issues: %d", len(securityAnalysis.WebsiteVerificationIssues))},
		})
	}

	return recommendations
}

// convertToMLModelRecommendation converts a generic recommendation to MLModelRecommendation
func (mre *MLRecommendationEngine) convertToMLModelRecommendation(rec *MLRecommendation, analysis *MLAnalysisResult) *MLModelRecommendation {
	if rec == nil {
		return nil
	}

	// Generate recommendation ID
	recommendationID := fmt.Sprintf("ml_rec_%d_%s", time.Now().UnixNano(), rec.Type)

	// Determine model ID based on recommendation type
	modelID := "ensemble" // Default to ensemble
	if strings.Contains(rec.Type, "ml_model") || strings.Contains(rec.Type, "bert") {
		modelID = "ml_model"
	} else if strings.Contains(rec.Type, "keyword") {
		modelID = "keyword_model"
	} else if strings.Contains(rec.Type, "similarity") {
		modelID = "similarity_model"
	}

	return &MLModelRecommendation{
		RecommendationID:    recommendationID,
		ModelID:             modelID,
		RecommendationType:  rec.Type,
		Priority:            rec.Priority,
		Description:         rec.Description,
		ExpectedImprovement: rec.ExpectedImprovement,
		ImplementationCost:  rec.ImplementationCost,
		Timeline:            rec.Timeline,
		Evidence:            rec.Evidence,
	}
}

// MLRecommendation represents a generic ML recommendation
type MLRecommendation struct {
	Type                string   `json:"type"`
	Priority            string   `json:"priority"`
	Description         string   `json:"description"`
	ExpectedImprovement float64  `json:"expected_improvement"`
	ImplementationCost  string   `json:"implementation_cost"`
	Timeline            string   `json:"timeline"`
	Evidence            []string `json:"evidence"`
}
