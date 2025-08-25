package industry_codes

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// ConfidenceCalculatorConfig defines configuration for confidence calculations
type ConfidenceCalculatorConfig struct {
	EnableAdaptiveWeighting    bool               `json:"enable_adaptive_weighting"`
	EnablePerformanceTracking  bool               `json:"enable_performance_tracking"`
	EnableCrossValidation      bool               `json:"enable_cross_validation"`
	BaseWeightAdjustmentFactor float64            `json:"base_weight_adjustment_factor"`
	PerformanceDecayFactor     float64            `json:"performance_decay_factor"`
	MinimumSampleSize          int                `json:"minimum_sample_size"`
	ConfidenceThresholds       map[string]float64 `json:"confidence_thresholds"`
}

// StrategyPerformanceMetrics tracks performance of a classification strategy
type StrategyPerformanceMetrics struct {
	StrategyName         string    `json:"strategy_name"`
	TotalClassifications int       `json:"total_classifications"`
	SuccessfulMatches    int       `json:"successful_matches"`
	AverageConfidence    float64   `json:"average_confidence"`
	AverageAccuracy      float64   `json:"average_accuracy"`
	PerformanceScore     float64   `json:"performance_score"`
	LastUpdated          time.Time `json:"last_updated"`
	RecentResults        []float64 `json:"recent_results"` // Sliding window for recent performance
}

// AdvancedConfidenceMetrics provides detailed confidence analysis
type AdvancedConfidenceMetrics struct {
	BaseConfidence        float64  `json:"base_confidence"`
	ConsistencyBonus      float64  `json:"consistency_bonus"`
	DiversityPenalty      float64  `json:"diversity_penalty"`
	PerformanceAdjustment float64  `json:"performance_adjustment"`
	AgreementBonus        float64  `json:"agreement_bonus"`
	FinalConfidence       float64  `json:"final_confidence"`
	ConfidenceLevel       string   `json:"confidence_level"`
	QualityIndicators     []string `json:"quality_indicators"`
}

// WeightingFactors contains various factors used in weighting calculations
type WeightingFactors struct {
	StrategyBaseWeight    float64 `json:"strategy_base_weight"`
	PerformanceMultiplier float64 `json:"performance_multiplier"`
	ConsistencyFactor     float64 `json:"consistency_factor"`
	RecencyFactor         float64 `json:"recency_factor"`
	AgreementFactor       float64 `json:"agreement_factor"`
	FinalWeight           float64 `json:"final_weight"`
}

// ConfidenceCalculator provides advanced confidence calculation capabilities
type ConfidenceCalculator struct {
	config             *ConfidenceCalculatorConfig
	logger             *zap.Logger
	performanceMetrics map[string]*StrategyPerformanceMetrics
}

// NewConfidenceCalculator creates a new confidence calculator
func NewConfidenceCalculator(config *ConfidenceCalculatorConfig, logger *zap.Logger) *ConfidenceCalculator {
	if config == nil {
		config = &ConfidenceCalculatorConfig{
			EnableAdaptiveWeighting:    true,
			EnablePerformanceTracking:  true,
			EnableCrossValidation:      false,
			BaseWeightAdjustmentFactor: 0.1,
			PerformanceDecayFactor:     0.95,
			MinimumSampleSize:          10,
			ConfidenceThresholds: map[string]float64{
				"very_high": 0.9,
				"high":      0.75,
				"medium":    0.6,
				"low":       0.4,
				"very_low":  0.2,
			},
		}
	}

	return &ConfidenceCalculator{
		config:             config,
		logger:             logger,
		performanceMetrics: make(map[string]*StrategyPerformanceMetrics),
	}
}

// CalculateAdvancedStrategyConfidence calculates sophisticated confidence for a strategy's results
func (cc *ConfidenceCalculator) CalculateAdvancedStrategyConfidence(ctx context.Context, strategyName string, results []*ClassificationResult) (*AdvancedConfidenceMetrics, error) {
	if len(results) == 0 {
		return &AdvancedConfidenceMetrics{
			BaseConfidence:  0.0,
			FinalConfidence: 0.0,
			ConfidenceLevel: "none",
		}, nil
	}

	metrics := &AdvancedConfidenceMetrics{
		QualityIndicators: []string{},
	}

	// Calculate base confidence from result scores
	totalConfidence := 0.0
	confidenceValues := make([]float64, len(results))
	for i, result := range results {
		totalConfidence += result.Confidence
		confidenceValues[i] = result.Confidence
	}
	metrics.BaseConfidence = totalConfidence / float64(len(results))

	// Calculate consistency bonus based on confidence variance
	metrics.ConsistencyBonus = cc.calculateConsistencyBonus(confidenceValues)

	// Calculate diversity penalty for too many results
	metrics.DiversityPenalty = cc.calculateDiversityPenalty(results)

	// Apply performance adjustment if tracking is enabled
	if cc.config.EnablePerformanceTracking {
		metrics.PerformanceAdjustment = cc.calculatePerformanceAdjustment(strategyName)
	}

	// Calculate agreement bonus (requires external agreement info)
	metrics.AgreementBonus = 0.0 // Will be set externally when agreement is known

	// Calculate final confidence
	metrics.FinalConfidence = cc.combineFinalConfidence(metrics)
	metrics.ConfidenceLevel = cc.determineConfidenceLevel(metrics.FinalConfidence)
	metrics.QualityIndicators = cc.generateQualityIndicators(metrics, results)

	// Update performance metrics
	if cc.config.EnablePerformanceTracking {
		cc.updatePerformanceMetrics(strategyName, metrics.FinalConfidence)
	}

	cc.logger.Debug("Advanced strategy confidence calculated",
		zap.String("strategy", strategyName),
		zap.Float64("base_confidence", metrics.BaseConfidence),
		zap.Float64("final_confidence", metrics.FinalConfidence),
		zap.String("confidence_level", metrics.ConfidenceLevel),
		zap.Int("results_count", len(results)))

	return metrics, nil
}

// CalculateAdaptiveWeights calculates dynamic weights for strategies based on performance
func (cc *ConfidenceCalculator) CalculateAdaptiveWeights(ctx context.Context, strategyVotes []*StrategyVote) ([]*StrategyVote, error) {
	if !cc.config.EnableAdaptiveWeighting {
		return strategyVotes, nil
	}

	updatedVotes := make([]*StrategyVote, len(strategyVotes))
	copy(updatedVotes, strategyVotes)

	for i, vote := range updatedVotes {
		factors := cc.calculateWeightingFactors(vote.StrategyName, vote)
		vote.Weight = factors.FinalWeight

		// Store weighting factors in metadata
		if vote.Metadata == nil {
			vote.Metadata = make(map[string]interface{})
		}
		vote.Metadata["weighting_factors"] = factors

		cc.logger.Debug("Adaptive weight calculated",
			zap.String("strategy", vote.StrategyName),
			zap.Float64("original_weight", strategyVotes[i].Weight),
			zap.Float64("adaptive_weight", vote.Weight),
			zap.Float64("performance_multiplier", factors.PerformanceMultiplier))
	}

	return updatedVotes, nil
}

// CalculateEnhancedWeightedAverage calculates improved weighted average with advanced metrics
func (cc *ConfidenceCalculator) CalculateEnhancedWeightedAverage(aggregations map[string]*CodeVoteAggregation) ([]*ClassificationResult, error) {
	// Enhanced weighted scoring with multiple factors
	for _, aggregation := range aggregations {
		// Recalculate weighted score with advanced metrics
		enhancedScore := cc.calculateEnhancedAggregationScore(aggregation)
		aggregation.WeightedScore = enhancedScore

		// Calculate additional confidence metrics
		aggregation.AverageConfidence = cc.calculateAverageConfidence(aggregation)
		aggregation.ConfidenceVariance = cc.calculateConfidenceVariance(aggregation)
		aggregation.AgreementScore = cc.calculateInternalAgreement(aggregation)
	}

	// Sort by enhanced weighted score
	var sortedCodes []*CodeVoteAggregation
	for _, aggregation := range aggregations {
		sortedCodes = append(sortedCodes, aggregation)
	}

	// Advanced sorting with multiple criteria
	cc.sortByMultipleCriteria(sortedCodes)

	// Generate results with enhanced confidence calculation
	var results []*ClassificationResult
	for i, aggregation := range sortedCodes {
		result := cc.createEnhancedResult(aggregation, i)
		results = append(results, result)

		if i >= 10 { // Limit results
			break
		}
	}

	return results, nil
}

// GetStrategyPerformanceMetrics returns performance metrics for a strategy
func (cc *ConfidenceCalculator) GetStrategyPerformanceMetrics(strategyName string) *StrategyPerformanceMetrics {
	return cc.performanceMetrics[strategyName]
}

// Internal helper methods

func (cc *ConfidenceCalculator) calculateConsistencyBonus(confidenceValues []float64) float64 {
	if len(confidenceValues) <= 1 {
		return 0.0
	}

	// Calculate coefficient of variation (std dev / mean)
	mean := 0.0
	for _, val := range confidenceValues {
		mean += val
	}
	mean /= float64(len(confidenceValues))

	variance := 0.0
	for _, val := range confidenceValues {
		variance += math.Pow(val-mean, 2)
	}
	variance /= float64(len(confidenceValues))
	stdDev := math.Sqrt(variance)

	if mean == 0 {
		return 0.0
	}

	coefficientOfVariation := stdDev / mean

	// Lower variation = higher consistency bonus (max 0.1)
	consistencyBonus := math.Max(0.0, (1.0-coefficientOfVariation)*0.1)
	return math.Min(0.1, consistencyBonus)
}

func (cc *ConfidenceCalculator) calculateDiversityPenalty(results []*ClassificationResult) float64 {
	// Penalize strategies with too many or too few results
	resultCount := len(results)

	if resultCount < 2 {
		return 0.05 // Small penalty for too few results
	}

	if resultCount > 10 {
		// Increasing penalty for too many results
		excess := float64(resultCount - 10)
		return math.Min(0.2, excess*0.02) // Max penalty of 0.2
	}

	return 0.0
}

func (cc *ConfidenceCalculator) calculatePerformanceAdjustment(strategyName string) float64 {
	metrics := cc.performanceMetrics[strategyName]
	if metrics == nil || metrics.TotalClassifications < cc.config.MinimumSampleSize {
		return 0.0 // No adjustment for new or insufficient data
	}

	// Performance adjustment based on historical success rate
	performanceScore := metrics.PerformanceScore

	// Convert performance score to adjustment (-0.1 to +0.1)
	adjustment := (performanceScore - 0.5) * cc.config.BaseWeightAdjustmentFactor
	return math.Max(-0.1, math.Min(0.1, adjustment))
}

func (cc *ConfidenceCalculator) combineFinalConfidence(metrics *AdvancedConfidenceMetrics) float64 {
	final := metrics.BaseConfidence +
		metrics.ConsistencyBonus -
		metrics.DiversityPenalty +
		metrics.PerformanceAdjustment +
		metrics.AgreementBonus

	return math.Max(0.0, math.Min(1.0, final))
}

func (cc *ConfidenceCalculator) determineConfidenceLevel(confidence float64) string {
	// Check thresholds in descending order
	if confidence >= cc.config.ConfidenceThresholds["very_high"] {
		return "very_high"
	}
	if confidence >= cc.config.ConfidenceThresholds["high"] {
		return "high"
	}
	if confidence >= cc.config.ConfidenceThresholds["medium"] {
		return "medium"
	}
	if confidence >= cc.config.ConfidenceThresholds["low"] {
		return "low"
	}
	return "very_low"
}

func (cc *ConfidenceCalculator) generateQualityIndicators(metrics *AdvancedConfidenceMetrics, results []*ClassificationResult) []string {
	indicators := []string{}

	if metrics.ConsistencyBonus > 0.05 {
		indicators = append(indicators, "high_consistency")
	}

	if metrics.DiversityPenalty > 0.1 {
		indicators = append(indicators, "low_focus")
	}

	if metrics.PerformanceAdjustment > 0.05 {
		indicators = append(indicators, "strong_historical_performance")
	} else if metrics.PerformanceAdjustment < -0.05 {
		indicators = append(indicators, "weak_historical_performance")
	}

	if len(results) >= 3 && len(results) <= 5 {
		indicators = append(indicators, "optimal_result_count")
	}

	return indicators
}

func (cc *ConfidenceCalculator) updatePerformanceMetrics(strategyName string, confidence float64) {
	metrics := cc.performanceMetrics[strategyName]
	if metrics == nil {
		metrics = &StrategyPerformanceMetrics{
			StrategyName:  strategyName,
			RecentResults: make([]float64, 0, 20), // Keep last 20 results
			LastUpdated:   time.Now(),
		}
		cc.performanceMetrics[strategyName] = metrics
	}

	metrics.TotalClassifications++
	if confidence > 0.5 {
		metrics.SuccessfulMatches++
	}

	// Update recent results (sliding window)
	metrics.RecentResults = append(metrics.RecentResults, confidence)
	if len(metrics.RecentResults) > 20 {
		metrics.RecentResults = metrics.RecentResults[1:]
	}

	// Recalculate averages
	cc.recalculatePerformanceMetrics(metrics)
}

func (cc *ConfidenceCalculator) recalculatePerformanceMetrics(metrics *StrategyPerformanceMetrics) {
	if len(metrics.RecentResults) == 0 {
		return
	}

	// Calculate average confidence from recent results
	total := 0.0
	for _, result := range metrics.RecentResults {
		total += result
	}
	metrics.AverageConfidence = total / float64(len(metrics.RecentResults))

	// Calculate accuracy as percentage of results above 0.5
	successCount := 0
	for _, result := range metrics.RecentResults {
		if result > 0.5 {
			successCount++
		}
	}
	metrics.AverageAccuracy = float64(successCount) / float64(len(metrics.RecentResults))

	// Calculate performance score (weighted combination)
	metrics.PerformanceScore = (metrics.AverageConfidence * 0.6) + (metrics.AverageAccuracy * 0.4)
	metrics.LastUpdated = time.Now()
}

func (cc *ConfidenceCalculator) calculateWeightingFactors(strategyName string, vote *StrategyVote) *WeightingFactors {
	factors := &WeightingFactors{
		StrategyBaseWeight:    vote.Weight, // Original weight
		PerformanceMultiplier: 1.0,
		ConsistencyFactor:     1.0,
		RecencyFactor:         1.0,
		AgreementFactor:       1.0,
	}

	// Performance-based adjustment
	if metrics := cc.performanceMetrics[strategyName]; metrics != nil && metrics.TotalClassifications >= cc.config.MinimumSampleSize {
		// Adjust based on performance score (0.8 to 1.2 multiplier)
		factors.PerformanceMultiplier = 0.8 + (metrics.PerformanceScore * 0.4)
	}

	// Consistency factor based on result confidence variance
	if len(vote.Results) > 1 {
		confidences := make([]float64, len(vote.Results))
		for i, result := range vote.Results {
			confidences[i] = result.Confidence
		}
		variance := cc.calculateVariance(confidences)
		factors.ConsistencyFactor = math.Max(0.8, 1.0-variance*0.5) // Lower variance = higher factor
	}

	// Recency factor (slightly favor recent votes)
	timeSinceVote := time.Since(vote.VoteTime)
	if timeSinceVote < time.Minute {
		factors.RecencyFactor = 1.05 // Small boost for very recent votes
	}

	// Calculate final weight
	factors.FinalWeight = factors.StrategyBaseWeight *
		factors.PerformanceMultiplier *
		factors.ConsistencyFactor *
		factors.RecencyFactor *
		factors.AgreementFactor

	// Ensure weight stays within reasonable bounds
	factors.FinalWeight = math.Max(0.1, math.Min(1.5, factors.FinalWeight))

	return factors
}

func (cc *ConfidenceCalculator) calculateEnhancedAggregationScore(aggregation *CodeVoteAggregation) float64 {
	if aggregation.TotalVotes == 0 {
		return 0.0
	}

	// Base weighted score
	baseScore := aggregation.WeightedScore / float64(aggregation.TotalVotes)

	// Bonus for multiple votes (consensus bonus)
	consensusBonus := 0.0
	if aggregation.TotalVotes > 1 {
		consensusBonus = math.Min(0.15, float64(aggregation.TotalVotes-1)*0.05)
	}

	// Variance penalty (prefer consistent scores)
	variancePenalty := math.Min(0.1, aggregation.ConfidenceVariance*0.2)

	enhancedScore := baseScore + consensusBonus - variancePenalty
	return math.Max(0.0, math.Min(1.0, enhancedScore))
}

func (cc *ConfidenceCalculator) calculateAverageConfidence(aggregation *CodeVoteAggregation) float64 {
	if len(aggregation.Votes) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	count := 0

	for _, vote := range aggregation.Votes {
		for _, result := range vote.Results {
			if result.Code.Code == aggregation.Code.Code && result.Code.Type == aggregation.Code.Type {
				totalConfidence += result.Confidence
				count++
			}
		}
	}

	if count == 0 {
		return 0.0
	}

	return totalConfidence / float64(count)
}

func (cc *ConfidenceCalculator) calculateConfidenceVariance(aggregation *CodeVoteAggregation) float64 {
	confidences := []float64{}

	for _, vote := range aggregation.Votes {
		for _, result := range vote.Results {
			if result.Code.Code == aggregation.Code.Code && result.Code.Type == aggregation.Code.Type {
				confidences = append(confidences, result.Confidence)
			}
		}
	}

	return cc.calculateVariance(confidences)
}

func (cc *ConfidenceCalculator) calculateInternalAgreement(aggregation *CodeVoteAggregation) float64 {
	if len(aggregation.Votes) <= 1 {
		return 1.0 // Perfect agreement with self
	}

	// Calculate agreement based on confidence consistency
	confidences := []float64{}
	for _, vote := range aggregation.Votes {
		for _, result := range vote.Results {
			if result.Code.Code == aggregation.Code.Code && result.Code.Type == aggregation.Code.Type {
				confidences = append(confidences, result.Confidence)
			}
		}
	}

	if len(confidences) <= 1 {
		return 1.0
	}

	variance := cc.calculateVariance(confidences)
	agreement := math.Max(0.0, 1.0-variance*2.0) // Lower variance = higher agreement
	return agreement
}

func (cc *ConfidenceCalculator) calculateVariance(values []float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}

	mean := 0.0
	for _, val := range values {
		mean += val
	}
	mean /= float64(len(values))

	variance := 0.0
	for _, val := range values {
		variance += math.Pow(val-mean, 2)
	}
	variance /= float64(len(values))

	return variance
}

func (cc *ConfidenceCalculator) sortByMultipleCriteria(aggregations []*CodeVoteAggregation) {
	// Multi-criteria sorting: weighted score, then agreement, then vote count
	for i := 0; i < len(aggregations); i++ {
		for j := i + 1; j < len(aggregations); j++ {
			shouldSwap := false

			// Primary: weighted score
			if aggregations[i].WeightedScore < aggregations[j].WeightedScore {
				shouldSwap = true
			} else if aggregations[i].WeightedScore == aggregations[j].WeightedScore {
				// Secondary: agreement score
				if aggregations[i].AgreementScore < aggregations[j].AgreementScore {
					shouldSwap = true
				} else if aggregations[i].AgreementScore == aggregations[j].AgreementScore {
					// Tertiary: vote count
					if aggregations[i].TotalVotes < aggregations[j].TotalVotes {
						shouldSwap = true
					}
				}
			}

			if shouldSwap {
				aggregations[i], aggregations[j] = aggregations[j], aggregations[i]
			}
		}
	}
}

func (cc *ConfidenceCalculator) createEnhancedResult(aggregation *CodeVoteAggregation, rank int) *ClassificationResult {
	finalConfidence := aggregation.WeightedScore

	// Additional confidence boost based on consensus and agreement
	consensusBoost := 0.0
	if aggregation.TotalVotes > 1 {
		consensusBoost = math.Min(0.1, float64(aggregation.TotalVotes-1)*0.02)
	}

	agreementBoost := aggregation.AgreementScore * 0.05

	finalConfidence = math.Min(1.0, finalConfidence+consensusBoost+agreementBoost)

	reasons := []string{
		fmt.Sprintf("Enhanced weighted average: %.3f", aggregation.WeightedScore),
		fmt.Sprintf("Consensus boost: %.3f", consensusBoost),
		fmt.Sprintf("Agreement: %.3f", aggregation.AgreementScore),
		fmt.Sprintf("Vote count: %d", aggregation.TotalVotes),
	}

	return &ClassificationResult{
		Code:       aggregation.Code,
		Confidence: finalConfidence,
		MatchType:  "enhanced_weighted_average",
		MatchedOn:  []string{"enhanced_consensus", "adaptive_weighting"},
		Reasons:    reasons,
		Weight:     1.0 - (float64(rank) * 0.05), // Slight decrease in weight by rank
	}
}
