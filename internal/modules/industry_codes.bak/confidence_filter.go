package industry_codes

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// ConfidenceThreshold defines confidence thresholds for filtering
type ConfidenceThreshold struct {
	GlobalMinConfidence    float64            `json:"global_min_confidence"`
	TypeSpecificThresholds map[string]float64 `json:"type_specific_thresholds"`
	AdaptiveThresholds     *AdaptiveThreshold `json:"adaptive_thresholds"`
	QualityBasedThresholds *QualityThreshold  `json:"quality_based_thresholds"`
	ValidationRules        []ThresholdRule    `json:"validation_rules"`
}

// AdaptiveThreshold provides adaptive confidence thresholds
type AdaptiveThreshold struct {
	Enabled           bool    `json:"enabled"`
	BaseThreshold     float64 `json:"base_threshold"`
	QualityMultiplier float64 `json:"quality_multiplier"`
	VolumeMultiplier  float64 `json:"volume_multiplier"`
	MaxThreshold      float64 `json:"max_threshold"`
	MinThreshold      float64 `json:"min_threshold"`
}

// QualityThreshold provides quality-based confidence thresholds
type QualityThreshold struct {
	Enabled           bool               `json:"enabled"`
	HighQualityMin    float64            `json:"high_quality_min"`
	MediumQualityMin  float64            `json:"medium_quality_min"`
	LowQualityMin     float64            `json:"low_quality_min"`
	QualityThresholds map[string]float64 `json:"quality_thresholds"`
}

// ThresholdRule defines validation rules for confidence thresholds
type ThresholdRule struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "minimum", "maximum", "range", "conditional"
	Value       float64                `json:"value"`
	MinValue    float64                `json:"min_value,omitempty"`
	MaxValue    float64                `json:"max_value,omitempty"`
	Condition   string                 `json:"condition,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	Description string                 `json:"description"`
}

// FilteringResult represents the result of confidence filtering
type FilteringResult struct {
	OriginalCount       int                  `json:"original_count"`
	FilteredCount       int                  `json:"filtered_count"`
	RejectedCount       int                  `json:"rejected_count"`
	ThresholdUsed       float64              `json:"threshold_used"`
	FilteringMetrics    *FilteringMetrics    `json:"filtering_metrics"`
	RejectionReasons    map[string][]string  `json:"rejection_reasons"`
	QualityAnalysis     *QualityAnalysis     `json:"quality_analysis"`
	AdaptiveAdjustments []AdaptiveAdjustment `json:"adaptive_adjustments"`
}

// FilteringMetrics provides metrics about the filtering process
type FilteringMetrics struct {
	AverageConfidenceBefore float64        `json:"average_confidence_before"`
	AverageConfidenceAfter  float64        `json:"average_confidence_after"`
	ConfidenceDistribution  map[string]int `json:"confidence_distribution"`
	TypeDistribution        map[string]int `json:"type_distribution"`
	QualityDistribution     map[string]int `json:"quality_distribution"`
	FilteringTime           time.Duration  `json:"filtering_time"`
	ThresholdEffectiveness  float64        `json:"threshold_effectiveness"`
}

// QualityAnalysis provides analysis of result quality
type QualityAnalysis struct {
	HighQualityCount   int                `json:"high_quality_count"`
	MediumQualityCount int                `json:"medium_quality_count"`
	LowQualityCount    int                `json:"low_quality_count"`
	AverageQuality     float64            `json:"average_quality"`
	QualityScore       float64            `json:"quality_score"`
	QualityFactors     map[string]float64 `json:"quality_factors"`
}

// AdaptiveAdjustment represents an adaptive threshold adjustment
type AdaptiveAdjustment struct {
	OriginalThreshold float64 `json:"original_threshold"`
	AdjustedThreshold float64 `json:"adjusted_threshold"`
	AdjustmentFactor  float64 `json:"adjustment_factor"`
	Reason            string  `json:"reason"`
	AppliedTo         string  `json:"applied_to"`
}

// ConfidenceFilter provides advanced confidence filtering capabilities
type ConfidenceFilter struct {
	confidenceScorer *ConfidenceScorer
	logger           *zap.Logger
	defaultThreshold *ConfidenceThreshold
}

// NewConfidenceFilter creates a new confidence filter
func NewConfidenceFilter(confidenceScorer *ConfidenceScorer, logger *zap.Logger) *ConfidenceFilter {
	return &ConfidenceFilter{
		confidenceScorer: confidenceScorer,
		logger:           logger,
		defaultThreshold: &ConfidenceThreshold{
			GlobalMinConfidence: 0.3,
			TypeSpecificThresholds: map[string]float64{
				"mcc":   0.25,
				"sic":   0.35,
				"naics": 0.4,
			},
			AdaptiveThresholds: &AdaptiveThreshold{
				Enabled:           true,
				BaseThreshold:     0.3,
				QualityMultiplier: 0.1,
				VolumeMultiplier:  0.05,
				MaxThreshold:      0.8,
				MinThreshold:      0.1,
			},
			QualityBasedThresholds: &QualityThreshold{
				Enabled:          true,
				HighQualityMin:   0.2,
				MediumQualityMin: 0.4,
				LowQualityMin:    0.6,
				QualityThresholds: map[string]float64{
					"high":   0.2,
					"medium": 0.4,
					"low":    0.6,
				},
			},
			ValidationRules: []ThresholdRule{
				{
					Name:        "global_minimum",
					Type:        "minimum",
					Value:       0.1,
					Description: "Global minimum confidence threshold",
				},
				{
					Name:        "global_maximum",
					Type:        "maximum",
					Value:       1.0,
					Description: "Global maximum confidence threshold",
				},
			},
		},
	}
}

// FilterByConfidence filters results by confidence thresholds
func (cf *ConfidenceFilter) FilterByConfidence(ctx context.Context, results []*ClassificationResult, request *ClassificationRequest, threshold *ConfidenceThreshold) (*FilteringResult, []*ClassificationResult, error) {
	startTime := time.Now()

	if threshold == nil {
		threshold = cf.defaultThreshold
	}

	cf.logger.Info("Starting confidence filtering",
		zap.Int("total_results", len(results)),
		zap.Float64("global_min_confidence", threshold.GlobalMinConfidence))

	// Step 1: Calculate confidence scores for all results
	rankedResults, err := cf.calculateConfidenceScores(ctx, results, request)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to calculate confidence scores: %w", err)
	}

	// Step 2: Determine effective thresholds
	effectiveThresholds := cf.calculateEffectiveThresholds(rankedResults, threshold, request)

	// Step 3: Apply filtering
	filteredResults, rejectionReasons := cf.applyFiltering(rankedResults, effectiveThresholds)

	// Step 4: Calculate metrics
	filteringMetrics := cf.calculateFilteringMetrics(results, filteredResults, startTime)
	qualityAnalysis := cf.analyzeQuality(rankedResults)

	// Step 5: Create filtering result
	filteringResult := &FilteringResult{
		OriginalCount:       len(results),
		FilteredCount:       len(filteredResults),
		RejectedCount:       len(results) - len(filteredResults),
		ThresholdUsed:       effectiveThresholds.GlobalThreshold,
		FilteringMetrics:    filteringMetrics,
		RejectionReasons:    rejectionReasons,
		QualityAnalysis:     qualityAnalysis,
		AdaptiveAdjustments: effectiveThresholds.Adjustments,
	}

	cf.logger.Info("Confidence filtering completed",
		zap.Int("filtered_count", len(filteredResults)),
		zap.Int("rejected_count", filteringResult.RejectedCount),
		zap.Float64("threshold_used", effectiveThresholds.GlobalThreshold))

	return filteringResult, filteredResults, nil
}

// calculateConfidenceScores calculates confidence scores for all results
func (cf *ConfidenceFilter) calculateConfidenceScores(ctx context.Context, results []*ClassificationResult, request *ClassificationRequest) ([]*RankingResult, error) {
	rankedResults := make([]*RankingResult, 0, len(results))

	for _, result := range results {
		confidenceScore, err := cf.confidenceScorer.CalculateConfidence(ctx, result, request)
		if err != nil {
			cf.logger.Warn("Failed to calculate confidence score for result",
				zap.String("code", result.Code.Code),
				zap.Error(err))
			// Use the original confidence as fallback
			confidenceScore = &ConfidenceScore{
				OverallScore:     result.Confidence,
				ConfidenceLevel:  "unknown",
				ValidationStatus: "unchecked",
				Factors: &ConfidenceFactors{
					TextMatchScore:    result.Confidence,
					KeywordMatchScore: result.Confidence,
				},
				LastUpdated:  time.Now(),
				ScoreVersion: "fallback",
			}
		}

		rankedResult := &RankingResult{
			ClassificationResult: result,
			ConfidenceScore:      confidenceScore,
			RankingScore:         0.0,
			Rank:                 0,
			TypeRank:             0,
			RankingFactors: &RankingFactors{
				ConfidenceFactor: confidenceScore.OverallScore,
			},
			QualityIndicators: []string{},
		}

		rankedResults = append(rankedResults, rankedResult)
	}

	return rankedResults, nil
}

// EffectiveThresholds represents the effective thresholds after adjustments
type EffectiveThresholds struct {
	GlobalThreshold   float64
	TypeThresholds    map[string]float64
	QualityThresholds map[string]float64
	Adjustments       []AdaptiveAdjustment
}

// calculateEffectiveThresholds calculates effective thresholds with adaptive adjustments
func (cf *ConfidenceFilter) calculateEffectiveThresholds(results []*RankingResult, threshold *ConfidenceThreshold, request *ClassificationRequest) *EffectiveThresholds {
	effective := &EffectiveThresholds{
		GlobalThreshold:   threshold.GlobalMinConfidence,
		TypeThresholds:    make(map[string]float64),
		QualityThresholds: make(map[string]float64),
		Adjustments:       []AdaptiveAdjustment{},
	}

	// Copy type-specific thresholds
	for codeType, typeThreshold := range threshold.TypeSpecificThresholds {
		effective.TypeThresholds[codeType] = typeThreshold
	}

	// Copy quality-based thresholds
	if threshold.QualityBasedThresholds != nil && threshold.QualityBasedThresholds.Enabled {
		for quality, qualityThreshold := range threshold.QualityBasedThresholds.QualityThresholds {
			effective.QualityThresholds[quality] = qualityThreshold
		}
	}

	// Apply adaptive adjustments if enabled
	if threshold.AdaptiveThresholds != nil && threshold.AdaptiveThresholds.Enabled {
		cf.applyAdaptiveAdjustments(effective, results, threshold.AdaptiveThresholds, request)
	}

	// Validate thresholds against rules
	cf.validateThresholds(effective, threshold.ValidationRules)

	return effective
}

// applyAdaptiveAdjustments applies adaptive threshold adjustments
func (cf *ConfidenceFilter) applyAdaptiveAdjustments(effective *EffectiveThresholds, results []*RankingResult, adaptive *AdaptiveThreshold, request *ClassificationRequest) {
	// Calculate quality-based adjustment
	qualityAdjustment := cf.calculateQualityAdjustment(results, adaptive)
	if qualityAdjustment != 0 {
		effective.GlobalThreshold += qualityAdjustment
		effective.Adjustments = append(effective.Adjustments, AdaptiveAdjustment{
			OriginalThreshold: effective.GlobalThreshold - qualityAdjustment,
			AdjustedThreshold: effective.GlobalThreshold,
			AdjustmentFactor:  qualityAdjustment,
			Reason:            "quality-based adjustment",
			AppliedTo:         "global",
		})
	}

	// Calculate volume-based adjustment
	volumeAdjustment := cf.calculateVolumeAdjustment(len(results), adaptive)
	if volumeAdjustment != 0 {
		effective.GlobalThreshold += volumeAdjustment
		effective.Adjustments = append(effective.Adjustments, AdaptiveAdjustment{
			OriginalThreshold: effective.GlobalThreshold - volumeAdjustment,
			AdjustedThreshold: effective.GlobalThreshold,
			AdjustmentFactor:  volumeAdjustment,
			Reason:            "volume-based adjustment",
			AppliedTo:         "global",
		})
	}

	// Ensure thresholds stay within bounds
	effective.GlobalThreshold = math.Max(adaptive.MinThreshold, math.Min(adaptive.MaxThreshold, effective.GlobalThreshold))

	// Apply type-specific adjustments
	for codeType := range effective.TypeThresholds {
		typeAdjustment := cf.calculateTypeAdjustment(results, codeType, adaptive)
		if typeAdjustment != 0 {
			originalThreshold := effective.TypeThresholds[codeType]
			effective.TypeThresholds[codeType] += typeAdjustment
			effective.TypeThresholds[codeType] = math.Max(adaptive.MinThreshold, math.Min(adaptive.MaxThreshold, effective.TypeThresholds[codeType]))

			effective.Adjustments = append(effective.Adjustments, AdaptiveAdjustment{
				OriginalThreshold: originalThreshold,
				AdjustedThreshold: effective.TypeThresholds[codeType],
				AdjustmentFactor:  typeAdjustment,
				Reason:            "type-specific adjustment",
				AppliedTo:         codeType,
			})
		}
	}
}

// calculateQualityAdjustment calculates quality-based threshold adjustment
func (cf *ConfidenceFilter) calculateQualityAdjustment(results []*RankingResult, adaptive *AdaptiveThreshold) float64 {
	if len(results) == 0 {
		return 0
	}

	// Calculate average quality score
	totalQuality := 0.0
	for _, result := range results {
		if result.ConfidenceScore != nil && result.ConfidenceScore.Factors != nil {
			qualityScore := (result.ConfidenceScore.Factors.TextMatchScore +
				result.ConfidenceScore.Factors.KeywordMatchScore +
				result.ConfidenceScore.Factors.NameMatchScore) / 3.0
			totalQuality += qualityScore
		}
	}

	averageQuality := totalQuality / float64(len(results))

	// Adjust threshold based on quality
	// Higher quality results allow for lower thresholds
	qualityAdjustment := (0.5 - averageQuality) * adaptive.QualityMultiplier

	return qualityAdjustment
}

// calculateVolumeAdjustment calculates volume-based threshold adjustment
func (cf *ConfidenceFilter) calculateVolumeAdjustment(resultCount int, adaptive *AdaptiveThreshold) float64 {
	// More results allow for higher thresholds (more selective)
	volumeAdjustment := float64(resultCount-10) * adaptive.VolumeMultiplier / 100.0

	return volumeAdjustment
}

// calculateTypeAdjustment calculates type-specific threshold adjustment
func (cf *ConfidenceFilter) calculateTypeAdjustment(results []*RankingResult, codeType string, adaptive *AdaptiveThreshold) float64 {
	typeResults := 0
	typeQuality := 0.0

	for _, result := range results {
		if string(result.Code.Type) == codeType {
			typeResults++
			if result.ConfidenceScore != nil && result.ConfidenceScore.Factors != nil {
				qualityScore := (result.ConfidenceScore.Factors.TextMatchScore +
					result.ConfidenceScore.Factors.KeywordMatchScore +
					result.ConfidenceScore.Factors.NameMatchScore) / 3.0
				typeQuality += qualityScore
			}
		}
	}

	if typeResults == 0 {
		return 0
	}

	averageTypeQuality := typeQuality / float64(typeResults)

	// Adjust threshold based on type-specific quality
	typeAdjustment := (0.5 - averageTypeQuality) * adaptive.QualityMultiplier

	return typeAdjustment
}

// validateThresholds validates thresholds against rules
func (cf *ConfidenceFilter) validateThresholds(effective *EffectiveThresholds, rules []ThresholdRule) {
	for _, rule := range rules {
		switch rule.Type {
		case "minimum":
			if effective.GlobalThreshold < rule.Value {
				effective.GlobalThreshold = rule.Value
			}
		case "maximum":
			if effective.GlobalThreshold > rule.Value {
				effective.GlobalThreshold = rule.Value
			}
		case "range":
			if effective.GlobalThreshold < rule.MinValue {
				effective.GlobalThreshold = rule.MinValue
			}
			if effective.GlobalThreshold > rule.MaxValue {
				effective.GlobalThreshold = rule.MaxValue
			}
		}
	}
}

// applyFiltering applies confidence filtering to results
func (cf *ConfidenceFilter) applyFiltering(results []*RankingResult, thresholds *EffectiveThresholds) ([]*ClassificationResult, map[string][]string) {
	filteredResults := make([]*ClassificationResult, 0, len(results))
	rejectionReasons := make(map[string][]string)

	for _, result := range results {
		confidence := result.ConfidenceScore.OverallScore
		codeType := string(result.Code.Type)

		// Determine applicable threshold
		applicableThreshold := thresholds.GlobalThreshold
		if typeThreshold, exists := thresholds.TypeThresholds[codeType]; exists {
			applicableThreshold = typeThreshold
		}

		// Check quality-based threshold
		if qualityThreshold, exists := thresholds.QualityThresholds[result.ConfidenceScore.ConfidenceLevel]; exists {
			if qualityThreshold > applicableThreshold {
				applicableThreshold = qualityThreshold
			}
		}

		// Apply filtering
		if confidence >= applicableThreshold {
			filteredResults = append(filteredResults, result.ClassificationResult)
		} else {
			reason := fmt.Sprintf("confidence %.3f below threshold %.3f", confidence, applicableThreshold)
			rejectionReasons[result.Code.Code] = append(rejectionReasons[result.Code.Code], reason)
		}
	}

	return filteredResults, rejectionReasons
}

// calculateFilteringMetrics calculates metrics about the filtering process
func (cf *ConfidenceFilter) calculateFilteringMetrics(originalResults []*ClassificationResult, filteredResults []*ClassificationResult, startTime time.Time) *FilteringMetrics {
	metrics := &FilteringMetrics{
		ConfidenceDistribution: make(map[string]int),
		TypeDistribution:       make(map[string]int),
		QualityDistribution:    make(map[string]int),
		FilteringTime:          time.Since(startTime),
	}

	// Calculate average confidence before filtering
	totalConfidenceBefore := 0.0
	for _, result := range originalResults {
		totalConfidenceBefore += result.Confidence
	}
	if len(originalResults) > 0 {
		metrics.AverageConfidenceBefore = totalConfidenceBefore / float64(len(originalResults))
	}

	// Calculate average confidence after filtering
	totalConfidenceAfter := 0.0
	for _, result := range filteredResults {
		totalConfidenceAfter += result.Confidence
	}
	if len(filteredResults) > 0 {
		metrics.AverageConfidenceAfter = totalConfidenceAfter / float64(len(filteredResults))
	}

	// Calculate confidence distribution
	for _, result := range originalResults {
		confidenceLevel := cf.getConfidenceLevel(result.Confidence)
		metrics.ConfidenceDistribution[confidenceLevel]++
	}

	// Calculate type distribution
	for _, result := range filteredResults {
		codeType := string(result.Code.Type)
		metrics.TypeDistribution[codeType]++
	}

	// Calculate quality distribution
	for _, result := range filteredResults {
		qualityLevel := cf.getQualityLevel(result.Confidence)
		metrics.QualityDistribution[qualityLevel]++
	}

	// Calculate threshold effectiveness
	if len(originalResults) > 0 {
		metrics.ThresholdEffectiveness = float64(len(filteredResults)) / float64(len(originalResults))
	}

	return metrics
}

// analyzeQuality analyzes the quality of filtered results
func (cf *ConfidenceFilter) analyzeQuality(results []*RankingResult) *QualityAnalysis {
	analysis := &QualityAnalysis{
		QualityFactors: make(map[string]float64),
	}

	if len(results) == 0 {
		return analysis
	}

	totalQuality := 0.0
	for _, result := range results {
		confidence := result.ConfidenceScore.OverallScore
		totalQuality += confidence

		qualityLevel := cf.getQualityLevel(confidence)
		switch qualityLevel {
		case "high":
			analysis.HighQualityCount++
		case "medium":
			analysis.MediumQualityCount++
		case "low":
			analysis.LowQualityCount++
		}
	}

	analysis.AverageQuality = totalQuality / float64(len(results))
	analysis.QualityScore = analysis.AverageQuality

	// Calculate quality factors
	analysis.QualityFactors["high_quality_ratio"] = float64(analysis.HighQualityCount) / float64(len(results))
	analysis.QualityFactors["medium_quality_ratio"] = float64(analysis.MediumQualityCount) / float64(len(results))
	analysis.QualityFactors["low_quality_ratio"] = float64(analysis.LowQualityCount) / float64(len(results))

	return analysis
}

// getConfidenceLevel returns the confidence level for a given confidence score
func (cf *ConfidenceFilter) getConfidenceLevel(confidence float64) string {
	switch {
	case confidence >= 0.8:
		return "very_high"
	case confidence >= 0.6:
		return "high"
	case confidence >= 0.4:
		return "medium"
	case confidence >= 0.2:
		return "low"
	default:
		return "very_low"
	}
}

// getQualityLevel returns the quality level for a given confidence score
func (cf *ConfidenceFilter) getQualityLevel(confidence float64) string {
	switch {
	case confidence >= 0.7:
		return "high"
	case confidence >= 0.5:
		return "medium"
	default:
		return "low"
	}
}

// GetDefaultThreshold returns the default confidence threshold
func (cf *ConfidenceFilter) GetDefaultThreshold() *ConfidenceThreshold {
	return cf.defaultThreshold
}

// SetDefaultThreshold sets the default confidence threshold
func (cf *ConfidenceFilter) SetDefaultThreshold(threshold *ConfidenceThreshold) {
	cf.defaultThreshold = threshold
}
