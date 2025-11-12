package feedback

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// EnsembleWeightOptimizer analyzes and optimizes ensemble weights based on feedback
type EnsembleWeightOptimizer struct {
	config *MLAnalysisConfig
	logger *zap.Logger
}

// NewEnsembleWeightOptimizer creates a new ensemble weight optimizer
func NewEnsembleWeightOptimizer(config *MLAnalysisConfig, logger *zap.Logger) *EnsembleWeightOptimizer {
	return &EnsembleWeightOptimizer{
		config: config,
		logger: logger,
	}
}

// AnalyzeWeightOptimization analyzes opportunities for ensemble weight optimization
func (ewo *EnsembleWeightOptimizer) AnalyzeWeightOptimization(ctx context.Context, feedback []*UserFeedback) ([]*WeightRecommendation, error) {
	if len(feedback) < ewo.config.MinFeedbackThreshold {
		return []*WeightRecommendation{}, nil
	}

	// Group feedback by method
	// Stub: UserFeedback doesn't have ClassificationMethod field
	// TODO: Refactor to use ClassificationClassificationUserFeedback
	methodFeedback := make(map[ClassificationMethod][]*UserFeedback)
	for _, fb := range feedback {
		// Default to ensemble method since we don't have ClassificationMethod
		methodFeedback[MethodEnsemble] = append(methodFeedback[MethodEnsemble], fb)
	}

	var recommendations []*WeightRecommendation

	// Analyze each method's performance
	methodPerformance := make(map[ClassificationMethod]*MethodPerformanceMetrics)
	for method, methodFb := range methodFeedback {
		if len(methodFb) >= ewo.config.MinFeedbackThreshold {
			perf := ewo.calculateMethodPerformance(methodFb)
			methodPerformance[method] = perf
		}
	}

	// Generate weight recommendations based on performance
	weightRecs := ewo.generateWeightRecommendations(methodPerformance)
	recommendations = append(recommendations, weightRecs...)

	// Analyze ensemble disagreement patterns
	disagreementRecs := ewo.analyzeEnsembleDisagreements(feedback, methodPerformance)
	recommendations = append(recommendations, disagreementRecs...)

	// Analyze temporal performance patterns
	temporalRecs := ewo.analyzeTemporalPerformance(feedback, methodPerformance)
	recommendations = append(recommendations, temporalRecs...)

	// Analyze confidence-based performance
	confidenceRecs := ewo.analyzeConfidenceBasedPerformance(feedback, methodPerformance)
	recommendations = append(recommendations, confidenceRecs...)

	// Sort recommendations by expected impact
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].ExpectedImpact > recommendations[j].ExpectedImpact
	})

	return recommendations, nil
}

// calculateMethodPerformance calculates performance metrics for a method
func (ewo *EnsembleWeightOptimizer) calculateMethodPerformance(feedback []*UserFeedback) *MethodPerformanceMetrics {
	metrics := &MethodPerformanceMetrics{
		TotalFeedback:  len(feedback),
		Accuracy:       0.0,
		Confidence:     0.0,
		ErrorRate:      0.0,
		ProcessingTime: 0.0,
		SecurityScore:  0.0,
		Consistency:    0.0,
		Reliability:    0.0,
	}

	if len(feedback) == 0 {
		return metrics
	}

	var totalAccuracy, totalConfidence, totalProcessingTime float64
	var accuracyCount, confidenceCount, processingTimeCount, errorCount, securityCount int
	var consistencyScores []float64

	// Stub: UserFeedback doesn't have FeedbackType, FeedbackValue, ConfidenceScore, or ProcessingTimeMs fields
	// TODO: Refactor to use ClassificationClassificationUserFeedback
	for _, fb := range feedback {
		// Use ClassificationAccuracy from UserFeedback if available
		if fb.ClassificationAccuracy > 0 {
			totalAccuracy += fb.ClassificationAccuracy
			accuracyCount++
			consistencyScores = append(consistencyScores, fb.ClassificationAccuracy)
		}
		// Confidence, processing time, error rate, and security counts would need to be extracted from Metadata
		// For now, skip these calculations
	}

	// Calculate final metrics
	if accuracyCount > 0 {
		metrics.Accuracy = totalAccuracy / float64(accuracyCount)
	}
	if confidenceCount > 0 {
		metrics.Confidence = totalConfidence / float64(confidenceCount)
	}
	if processingTimeCount > 0 {
		metrics.ProcessingTime = totalProcessingTime / float64(processingTimeCount)
	}
	metrics.ErrorRate = float64(errorCount) / float64(len(feedback))
	if securityCount > 0 {
		metrics.SecurityScore = totalAccuracy / float64(securityCount)
	}

	// Calculate consistency (inverse of variance)
	if len(consistencyScores) > 1 {
		metrics.Consistency = ewo.calculateConsistency(consistencyScores)
	} else {
		metrics.Consistency = 1.0 // Perfect consistency if only one score
	}

	// Calculate reliability (combination of accuracy, consistency, and security)
	metrics.Reliability = (metrics.Accuracy + metrics.Consistency + metrics.SecurityScore) / 3.0

	return metrics
}

// calculateConsistency calculates consistency based on confidence score variance
func (ewo *EnsembleWeightOptimizer) calculateConsistency(scores []float64) float64 {
	if len(scores) < 2 {
		return 1.0
	}

	// Calculate mean
	var sum float64
	for _, score := range scores {
		sum += score
	}
	mean := sum / float64(len(scores))

	// Calculate variance
	var variance float64
	for _, score := range scores {
		variance += (score - mean) * (score - mean)
	}
	variance /= float64(len(scores))

	// Convert variance to consistency (lower variance = higher consistency)
	// Use exponential decay to map variance to [0, 1]
	consistency := 1.0 / (1.0 + variance*10) // Scale factor of 10 for reasonable mapping

	return consistency
}

// generateWeightRecommendations generates weight recommendations based on method performance
func (ewo *EnsembleWeightOptimizer) generateWeightRecommendations(methodPerformance map[ClassificationMethod]*MethodPerformanceMetrics) []*WeightRecommendation {
	var recommendations []*WeightRecommendation

	// Calculate current weights (assuming equal weights for now)
	currentWeights := make(map[ClassificationMethod]float64)
	totalMethods := float64(len(methodPerformance))
	if totalMethods > 0 {
		equalWeight := 1.0 / totalMethods
		for method := range methodPerformance {
			currentWeights[method] = equalWeight
		}
	}

	// Calculate optimal weights based on performance
	optimalWeights := ewo.calculateOptimalWeights(methodPerformance)

	// Generate recommendations for weight changes
	for method, currentWeight := range currentWeights {
		optimalWeight := optimalWeights[method]
		weightChange := optimalWeight - currentWeight

		// Only recommend changes above threshold
		if abs(weightChange) >= ewo.config.WeightAdjustmentStep {
			// Limit weight change to max allowed
			if abs(weightChange) > ewo.config.MaxWeightChange {
				if weightChange > 0 {
					weightChange = ewo.config.MaxWeightChange
				} else {
					weightChange = -ewo.config.MaxWeightChange
				}
			}

			// Calculate expected impact
			expectedImpact := ewo.calculateExpectedImpact(method, methodPerformance[method], weightChange)

			// Determine confidence in recommendation
			confidence := ewo.calculateRecommendationConfidence(method, methodPerformance[method])

			// Generate reasoning
			reasoning := ewo.generateWeightReasoning(method, methodPerformance[method], weightChange)

			recommendation := &WeightRecommendation{
				RecommendationID:   fmt.Sprintf("weight_rec_%s_%d", method, time.Now().UnixNano()),
				Method:             method,
				CurrentWeight:      currentWeight,
				RecommendedWeight:  currentWeight + weightChange,
				WeightChange:       weightChange,
				Reasoning:          reasoning,
				ExpectedImpact:     expectedImpact,
				Confidence:         confidence,
				ImplementationDate: time.Now().Add(24 * time.Hour), // Implement after 24 hours
			}

			recommendations = append(recommendations, recommendation)
		}
	}

	return recommendations
}

// calculateOptimalWeights calculates optimal weights based on method performance
func (ewo *EnsembleWeightOptimizer) calculateOptimalWeights(methodPerformance map[ClassificationMethod]*MethodPerformanceMetrics) map[ClassificationMethod]float64 {
	optimalWeights := make(map[ClassificationMethod]float64)

	// Calculate performance scores for each method
	performanceScores := make(map[ClassificationMethod]float64)
	var totalScore float64

	for method, metrics := range methodPerformance {
		// Calculate composite performance score
		score := ewo.calculateCompositePerformanceScore(metrics)
		performanceScores[method] = score
		totalScore += score
	}

	// Normalize to weights (ensure they sum to 1.0)
	if totalScore > 0 {
		for method, score := range performanceScores {
			optimalWeights[method] = score / totalScore
		}
	} else {
		// Fallback to equal weights if no performance data
		equalWeight := 1.0 / float64(len(methodPerformance))
		for method := range methodPerformance {
			optimalWeights[method] = equalWeight
		}
	}

	return optimalWeights
}

// calculateCompositePerformanceScore calculates a composite performance score for a method
func (ewo *EnsembleWeightOptimizer) calculateCompositePerformanceScore(metrics *MethodPerformanceMetrics) float64 {
	// Weighted combination of different performance metrics
	// Accuracy: 40%, Reliability: 30%, Consistency: 20%, Security: 10%

	accuracyScore := metrics.Accuracy * 0.4
	reliabilityScore := metrics.Reliability * 0.3
	consistencyScore := metrics.Consistency * 0.2
	securityScore := metrics.SecurityScore * 0.1

	// Penalize high error rates
	errorPenalty := metrics.ErrorRate * 0.2

	// Penalize slow processing times (normalize to 0-1 scale)
	processingPenalty := 0.0
	if metrics.ProcessingTime > 1000 { // More than 1 second
		processingPenalty = (metrics.ProcessingTime - 1000) / 10000 * 0.1 // Max 10% penalty
	}

	compositeScore := accuracyScore + reliabilityScore + consistencyScore + securityScore - errorPenalty - processingPenalty

	// Ensure score is between 0 and 1
	if compositeScore < 0 {
		compositeScore = 0
	} else if compositeScore > 1 {
		compositeScore = 1
	}

	return compositeScore
}

// calculateExpectedImpact calculates the expected impact of a weight change
func (ewo *EnsembleWeightOptimizer) calculateExpectedImpact(method ClassificationMethod, metrics *MethodPerformanceMetrics, weightChange float64) float64 {
	// Expected impact is based on the method's performance and the magnitude of weight change
	baseImpact := abs(weightChange) * metrics.Reliability

	// Scale impact based on method performance
	if metrics.Accuracy > 0.8 {
		baseImpact *= 1.2 // High-performing methods have higher impact
	} else if metrics.Accuracy < 0.5 {
		baseImpact *= 0.8 // Low-performing methods have lower impact
	}

	// Scale impact based on consistency
	if metrics.Consistency > 0.8 {
		baseImpact *= 1.1 // Consistent methods have higher impact
	} else if metrics.Consistency < 0.5 {
		baseImpact *= 0.9 // Inconsistent methods have lower impact
	}

	return baseImpact
}

// calculateRecommendationConfidence calculates confidence in the weight recommendation
func (ewo *EnsembleWeightOptimizer) calculateRecommendationConfidence(method ClassificationMethod, metrics *MethodPerformanceMetrics) float64 {
	// Base confidence on the amount of feedback and consistency
	baseConfidence := 0.5

	// Increase confidence with more feedback
	if metrics.TotalFeedback > 100 {
		baseConfidence += 0.3
	} else if metrics.TotalFeedback > 50 {
		baseConfidence += 0.2
	} else if metrics.TotalFeedback > 20 {
		baseConfidence += 0.1
	}

	// Increase confidence with higher consistency
	baseConfidence += metrics.Consistency * 0.2

	// Ensure confidence is between 0 and 1
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	}

	return baseConfidence
}

// generateWeightReasoning generates human-readable reasoning for weight recommendations
func (ewo *EnsembleWeightOptimizer) generateWeightReasoning(method ClassificationMethod, metrics *MethodPerformanceMetrics, weightChange float64) string {
	var reasoning string

	if weightChange > 0 {
		reasoning = fmt.Sprintf("Increase weight for %s method due to ", method)
	} else {
		reasoning = fmt.Sprintf("Decrease weight for %s method due to ", method)
	}

	// Add specific reasons based on performance metrics
	var reasons []string

	if metrics.Accuracy > 0.8 {
		reasons = append(reasons, "high accuracy")
	} else if metrics.Accuracy < 0.5 {
		reasons = append(reasons, "low accuracy")
	}

	if metrics.Consistency > 0.8 {
		reasons = append(reasons, "high consistency")
	} else if metrics.Consistency < 0.5 {
		reasons = append(reasons, "low consistency")
	}

	if metrics.SecurityScore > 0.9 {
		reasons = append(reasons, "excellent security")
	} else if metrics.SecurityScore < 0.7 {
		reasons = append(reasons, "security concerns")
	}

	if metrics.ErrorRate > 0.2 {
		reasons = append(reasons, "high error rate")
	}

	if metrics.ProcessingTime > 2000 {
		reasons = append(reasons, "slow processing")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "performance analysis")
	}

	reasoning += strings.Join(reasons, ", ")
	reasoning += fmt.Sprintf(" (accuracy: %.2f, consistency: %.2f, reliability: %.2f)",
		metrics.Accuracy, metrics.Consistency, metrics.Reliability)

	return reasoning
}

// analyzeEnsembleDisagreements analyzes ensemble disagreement patterns for weight optimization
func (ewo *EnsembleWeightOptimizer) analyzeEnsembleDisagreements(feedback []*UserFeedback, methodPerformance map[ClassificationMethod]*MethodPerformanceMetrics) []*WeightRecommendation {
	var recommendations []*WeightRecommendation

	// Look for feedback that indicates ensemble disagreements
	var disagreementFeedback []*UserFeedback
	for _, fb := range feedback {
		if strings.Contains(strings.ToLower(fb.Comments), "disagree") ||
			strings.Contains(strings.ToLower(fb.Comments), "conflict") ||
			strings.Contains(strings.ToLower(fb.Comments), "different") {
			disagreementFeedback = append(disagreementFeedback, fb)
		}
	}

	if len(disagreementFeedback) < ewo.config.MinFeedbackThreshold {
		return recommendations
	}

	// Analyze which methods are most often involved in disagreements
	// Stub: UserFeedback doesn't have ClassificationMethod field
	methodDisagreementCount := make(map[ClassificationMethod]int)
	for range disagreementFeedback {
		// All feedback assigned to MethodEnsemble
		methodDisagreementCount[MethodEnsemble]++
	}

	// Generate recommendations to reduce disagreements
	for method, count := range methodDisagreementCount {
		if count > len(disagreementFeedback)/3 { // Method involved in >33% of disagreements
			// Recommend reducing weight for methods with high disagreement involvement
			if _, exists := methodPerformance[method]; exists {
				recommendation := &WeightRecommendation{
					RecommendationID:   fmt.Sprintf("disagreement_rec_%s_%d", method, time.Now().UnixNano()),
					Method:             method,
					CurrentWeight:      1.0 / float64(len(methodPerformance)),         // Assume equal weights
					RecommendedWeight:  (1.0 / float64(len(methodPerformance))) * 0.8, // Reduce by 20%
					WeightChange:       -(1.0 / float64(len(methodPerformance))) * 0.2,
					Reasoning:          fmt.Sprintf("Reduce weight due to high disagreement involvement (%d/%d disagreements)", count, len(disagreementFeedback)),
					ExpectedImpact:     0.05,
					Confidence:         0.7,
					ImplementationDate: time.Now().Add(48 * time.Hour), // Implement after 48 hours
				}
				recommendations = append(recommendations, recommendation)
			}
		}
	}

	return recommendations
}

// analyzeTemporalPerformance analyzes temporal performance patterns for weight optimization
func (ewo *EnsembleWeightOptimizer) analyzeTemporalPerformance(feedback []*UserFeedback, methodPerformance map[ClassificationMethod]*MethodPerformanceMetrics) []*WeightRecommendation {
	var recommendations []*WeightRecommendation

	// Group feedback by time periods
	// Stub: UserFeedback doesn't have ClassificationMethod or CreatedAt fields
	hourlyFeedback := make(map[ClassificationMethod]map[int][]*UserFeedback)
	for _, fb := range feedback {
		hour := fb.SubmittedAt.Hour() // Use SubmittedAt instead of CreatedAt
		if hourlyFeedback[MethodEnsemble] == nil {
			hourlyFeedback[MethodEnsemble] = make(map[int][]*UserFeedback)
		}
		hourlyFeedback[MethodEnsemble][hour] = append(hourlyFeedback[MethodEnsemble][hour], fb)
	}

	// Analyze performance variations by time
	for method, hourData := range hourlyFeedback {
		// Calculate performance for each hour
		hourlyPerformance := make(map[int]*MethodPerformanceMetrics)
		for hour, hourFb := range hourData {
			if len(hourFb) >= 5 { // Minimum 5 feedback items per hour
				hourlyPerformance[hour] = ewo.calculateMethodPerformance(hourFb)
			}
		}

		// Find hours with significantly different performance
		basePerformance := methodPerformance[method]
		if basePerformance == nil {
			continue
		}

		for hour, hourPerf := range hourlyPerformance {
			// Check for significant performance differences
			accuracyDiff := hourPerf.Accuracy - basePerformance.Accuracy
			if abs(accuracyDiff) > 0.2 { // 20% difference
				// Generate time-based weight recommendation
				weightAdjustment := accuracyDiff * 0.1 // 10% of accuracy difference

				recommendation := &WeightRecommendation{
					RecommendationID:   fmt.Sprintf("temporal_rec_%s_hour_%d_%d", method, hour, time.Now().UnixNano()),
					Method:             method,
					CurrentWeight:      1.0 / float64(len(methodPerformance)),
					RecommendedWeight:  (1.0 / float64(len(methodPerformance))) + weightAdjustment,
					WeightChange:       weightAdjustment,
					Reasoning:          fmt.Sprintf("Adjust weight for hour %d based on performance difference (%.2f vs %.2f accuracy)", hour, hourPerf.Accuracy, basePerformance.Accuracy),
					ExpectedImpact:     abs(weightAdjustment) * 0.5,
					Confidence:         0.6,
					ImplementationDate: time.Now().Add(72 * time.Hour), // Implement after 72 hours
				}
				recommendations = append(recommendations, recommendation)
			}
		}
	}

	return recommendations
}

// analyzeConfidenceBasedPerformance analyzes confidence-based performance for weight optimization
func (ewo *EnsembleWeightOptimizer) analyzeConfidenceBasedPerformance(feedback []*UserFeedback, methodPerformance map[ClassificationMethod]*MethodPerformanceMetrics) []*WeightRecommendation {
	var recommendations []*WeightRecommendation

	// Group feedback by confidence ranges
	confidenceRanges := map[string][]float64{
		"low":    {0.0, 0.4},
		"medium": {0.4, 0.7},
		"high":   {0.7, 1.0},
	}

	for method := range methodPerformance {
		// Get feedback for this method
		var methodFeedback []*UserFeedback
		// Stub: UserFeedback doesn't have ClassificationMethod field
		// For now, include all feedback (they're all assigned to MethodEnsemble)
		methodFeedback = feedback

		if len(methodFeedback) < ewo.config.MinFeedbackThreshold {
			continue
		}

		// Analyze performance by confidence range
		rangePerformance := make(map[string]*MethodPerformanceMetrics)
		for rangeName, rangeBounds := range confidenceRanges {
			var rangeFeedback []*UserFeedback
			for _, fb := range methodFeedback {
				// Use ClassificationAccuracy instead of ConfidenceScore
				if fb.ClassificationAccuracy >= rangeBounds[0] && fb.ClassificationAccuracy < rangeBounds[1] {
					rangeFeedback = append(rangeFeedback, fb)
				}
			}

			if len(rangeFeedback) >= 5 {
				rangePerformance[rangeName] = ewo.calculateMethodPerformance(rangeFeedback)
			}
		}

		// Check for confidence calibration issues
		if lowPerf, exists := rangePerformance["low"]; exists {
			if highPerf, exists := rangePerformance["high"]; exists {
				// Check if low confidence predictions are actually more accurate than high confidence
				if lowPerf.Accuracy > highPerf.Accuracy+0.1 {
					// Method is poorly calibrated - recommend weight reduction
					recommendation := &WeightRecommendation{
						RecommendationID:   fmt.Sprintf("confidence_calibration_rec_%s_%d", method, time.Now().UnixNano()),
						Method:             method,
						CurrentWeight:      1.0 / float64(len(methodPerformance)),
						RecommendedWeight:  (1.0 / float64(len(methodPerformance))) * 0.9, // Reduce by 10%
						WeightChange:       -(1.0 / float64(len(methodPerformance))) * 0.1,
						Reasoning:          fmt.Sprintf("Reduce weight due to poor confidence calibration (low confidence accuracy: %.2f, high confidence accuracy: %.2f)", lowPerf.Accuracy, highPerf.Accuracy),
						ExpectedImpact:     0.08,
						Confidence:         0.8,
						ImplementationDate: time.Now().Add(24 * time.Hour),
					}
					recommendations = append(recommendations, recommendation)
				}
			}
		}
	}

	return recommendations
}

// MethodPerformanceMetrics represents performance metrics for a classification method
type MethodPerformanceMetrics struct {
	TotalFeedback  int     `json:"total_feedback"`
	Accuracy       float64 `json:"accuracy"`
	Confidence     float64 `json:"confidence"`
	ErrorRate      float64 `json:"error_rate"`
	ProcessingTime float64 `json:"processing_time_ms"`
	SecurityScore  float64 `json:"security_score"`
	Consistency    float64 `json:"consistency"`
	Reliability    float64 `json:"reliability"`
}

// Helper function to calculate absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
