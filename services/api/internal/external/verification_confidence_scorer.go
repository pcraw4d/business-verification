package external

import (
	"context"
	"fmt"
	"math"
	"time"
)

// ConfidenceLevel represents the confidence level categorization
type ConfidenceLevel string

const (
	ConfidenceLevelHigh   ConfidenceLevel = "high"
	ConfidenceLevelMedium ConfidenceLevel = "medium"
	ConfidenceLevelLow    ConfidenceLevel = "low"
)

// ConfidenceScore represents a detailed confidence score for verification
type ConfidenceScore struct {
	OverallScore    float64            `json:"overall_score"`
	ConfidenceLevel ConfidenceLevel    `json:"confidence_level"`
	FieldScores     map[string]float64 `json:"field_scores"`
	WeightedScores  map[string]float64 `json:"weighted_scores"`
	CalibrationData CalibrationData    `json:"calibration_data"`
	GeneratedAt     time.Time          `json:"generated_at"`
	ScoreBreakdown  ScoreBreakdown     `json:"score_breakdown"`
}

// CalibrationData contains data used for score calibration and validation
type CalibrationData struct {
	SampleSize         int       `json:"sample_size"`
	AverageScore       float64   `json:"average_score"`
	StandardDeviation  float64   `json:"standard_deviation"`
	ConfidenceInterval float64   `json:"confidence_interval"`
	LastCalibrated     time.Time `json:"last_calibrated"`
}

// ScoreBreakdown provides detailed breakdown of how the score was calculated
type ScoreBreakdown struct {
	BaseScores        map[string]float64 `json:"base_scores"`
	FieldWeights      map[string]float64 `json:"field_weights"`
	NormalizedScores  map[string]float64 `json:"normalized_scores"`
	ConfidenceFactors map[string]float64 `json:"confidence_factors"`
	PenaltyFactors    map[string]float64 `json:"penalty_factors"`
	BonusFactors      map[string]float64 `json:"bonus_factors"`
}

// ConfidenceScorerConfig defines the configuration for confidence scoring
type ConfidenceScorerConfig struct {
	FieldWeights         map[string]float64   `json:"field_weights"`
	ConfidenceThresholds ConfidenceThresholds `json:"confidence_thresholds"`
	CalibrationSettings  CalibrationSettings  `json:"calibration_settings"`
	ScoringAlgorithm     string               `json:"scoring_algorithm"`
}

// ConfidenceThresholds define the thresholds for confidence level categorization
type ConfidenceThresholds struct {
	HighThreshold   float64 `json:"high_threshold"`
	MediumThreshold float64 `json:"medium_threshold"`
	LowThreshold    float64 `json:"low_threshold"`
}

// CalibrationSettings define how confidence scores should be calibrated
type CalibrationSettings struct {
	MinSampleSize     int           `json:"min_sample_size"`
	CalibrationPeriod time.Duration `json:"calibration_period"`
	OutlierThreshold  float64       `json:"outlier_threshold"`
	ConfidenceLevel   float64       `json:"confidence_level"`
}

// ConfidenceScorer provides confidence scoring functionality
type ConfidenceScorer struct {
	config          ConfidenceScorerConfig
	calibrationData map[string]CalibrationData
}

// NewConfidenceScorer creates a new confidence scorer with default configuration
func NewConfidenceScorer() *ConfidenceScorer {
	return &ConfidenceScorer{
		config: ConfidenceScorerConfig{
			FieldWeights: map[string]float64{
				"business_name":   0.25,
				"phone_numbers":   0.20,
				"email_addresses": 0.15,
				"addresses":       0.20,
				"website_urls":    0.10,
				"industries":      0.10,
			},
			ConfidenceThresholds: ConfidenceThresholds{
				HighThreshold:   0.8,
				MediumThreshold: 0.6,
				LowThreshold:    0.4,
			},
			CalibrationSettings: CalibrationSettings{
				MinSampleSize:     100,
				CalibrationPeriod: 24 * time.Hour,
				OutlierThreshold:  2.0,
				ConfidenceLevel:   0.95,
			},
			ScoringAlgorithm: "weighted_average",
		},
		calibrationData: make(map[string]CalibrationData),
	}
}

// NewConfidenceScorerWithConfig creates a new confidence scorer with custom configuration
func NewConfidenceScorerWithConfig(config ConfidenceScorerConfig) *ConfidenceScorer {
	return &ConfidenceScorer{
		config:          config,
		calibrationData: make(map[string]CalibrationData),
	}
}

// CalculateConfidenceScore calculates a comprehensive confidence score for verification results
func (cs *ConfidenceScorer) CalculateConfidenceScore(ctx context.Context, result *VerificationResult) (*ConfidenceScore, error) {
	if result == nil {
		return nil, fmt.Errorf("verification result cannot be nil")
	}

	// Calculate base scores for each field
	fieldScores := cs.calculateFieldScores(result)

	// Apply field weights
	weightedScores := cs.applyFieldWeights(fieldScores)

	// Calculate confidence factors
	confidenceFactors := cs.calculateConfidenceFactors(result)

	// Apply penalties and bonuses
	penaltyFactors := cs.calculatePenaltyFactors(result)
	bonusFactors := cs.calculateBonusFactors(result)

	// Calculate normalized scores
	normalizedScores := cs.normalizeScores(fieldScores, confidenceFactors, penaltyFactors, bonusFactors)

	// Calculate overall score
	overallScore := cs.calculateOverallScore(normalizedScores, weightedScores)

	// Determine confidence level
	confidenceLevel := cs.determineConfidenceLevel(overallScore)

	// Get calibration data
	calibrationData := cs.getCalibrationData(string(result.Status))

	// Create score breakdown
	scoreBreakdown := ScoreBreakdown{
		BaseScores:        fieldScores,
		FieldWeights:      cs.config.FieldWeights,
		NormalizedScores:  normalizedScores,
		ConfidenceFactors: confidenceFactors,
		PenaltyFactors:    penaltyFactors,
		BonusFactors:      bonusFactors,
	}

	return &ConfidenceScore{
		OverallScore:    overallScore,
		ConfidenceLevel: confidenceLevel,
		FieldScores:     fieldScores,
		WeightedScores:  weightedScores,
		CalibrationData: calibrationData,
		GeneratedAt:     time.Now(),
		ScoreBreakdown:  scoreBreakdown,
	}, nil
}

// calculateFieldScores calculates base scores for each field in the verification result
func (cs *ConfidenceScorer) calculateFieldScores(result *VerificationResult) map[string]float64 {
	scores := make(map[string]float64)

	// Calculate scores based on field results
	for fieldName, fieldResult := range result.FieldResults {
		score := cs.calculateFieldScore(fieldResult)
		scores[fieldName] = score
	}

	return scores
}

// calculateFieldScore calculates the score for a single field
func (cs *ConfidenceScorer) calculateFieldScore(fieldResult FieldResult) float64 {
	// Base score from the field result
	baseScore := fieldResult.Score

	// Adjust based on confidence level
	confidenceAdjustment := fieldResult.Confidence * 0.2

	// Adjust based on match status
	matchAdjustment := 0.0
	if fieldResult.Matched {
		matchAdjustment = 0.1
	}

	// Calculate final score
	finalScore := baseScore + confidenceAdjustment + matchAdjustment

	// Ensure score is within 0-1 range
	return math.Max(0.0, math.Min(1.0, finalScore))
}

// applyFieldWeights applies field weights to the scores
func (cs *ConfidenceScorer) applyFieldWeights(fieldScores map[string]float64) map[string]float64 {
	weightedScores := make(map[string]float64)

	for fieldName, score := range fieldScores {
		weight := cs.config.FieldWeights[fieldName]
		if weight == 0 {
			weight = 0.1 // Default weight for unknown fields
		}
		weightedScores[fieldName] = score * weight
	}

	return weightedScores
}

// calculateConfidenceFactors calculates confidence factors based on verification context
func (cs *ConfidenceScorer) calculateConfidenceFactors(result *VerificationResult) map[string]float64 {
	factors := make(map[string]float64)

	// Data completeness factor
	completenessFactor := cs.calculateCompletenessFactor(result)
	factors["data_completeness"] = completenessFactor

	// Data consistency factor
	consistencyFactor := cs.calculateConsistencyFactor(result)
	factors["data_consistency"] = consistencyFactor

	// Source reliability factor
	reliabilityFactor := cs.calculateReliabilityFactor(result)
	factors["source_reliability"] = reliabilityFactor

	return factors
}

// calculateCompletenessFactor calculates how complete the verification data is
func (cs *ConfidenceScorer) calculateCompletenessFactor(result *VerificationResult) float64 {
	totalFields := len(result.FieldResults)
	if totalFields == 0 {
		return 0.0
	}

	populatedFields := 0
	for _, fieldResult := range result.FieldResults {
		if fieldResult.Score > 0 {
			populatedFields++
		}
	}

	return float64(populatedFields) / float64(totalFields)
}

// calculateConsistencyFactor calculates how consistent the verification data is
func (cs *ConfidenceScorer) calculateConsistencyFactor(result *VerificationResult) float64 {
	if len(result.FieldResults) < 2 {
		return 1.0 // Single field is always consistent
	}

	var scores []float64
	for _, fieldResult := range result.FieldResults {
		scores = append(scores, fieldResult.Score)
	}

	// Calculate standard deviation
	mean := 0.0
	for _, score := range scores {
		mean += score
	}
	mean /= float64(len(scores))

	variance := 0.0
	for _, score := range scores {
		variance += math.Pow(score-mean, 2)
	}
	variance /= float64(len(scores))

	stdDev := math.Sqrt(variance)

	// Convert to consistency factor (lower std dev = higher consistency)
	consistencyFactor := 1.0 - (stdDev * 2) // Scale factor
	return math.Max(0.0, math.Min(1.0, consistencyFactor))
}

// calculateReliabilityFactor calculates the reliability of the verification source
func (cs *ConfidenceScorer) calculateReliabilityFactor(result *VerificationResult) float64 {
	// This would typically be based on the source of the verification data
	// For now, we'll use a default high reliability factor
	return 0.9
}

// calculatePenaltyFactors calculates penalty factors that reduce confidence
func (cs *ConfidenceScorer) calculatePenaltyFactors(result *VerificationResult) map[string]float64 {
	penalties := make(map[string]float64)

	// Penalty for failed status
	if result.Status == "FAILED" {
		penalties["status_failed"] = 0.3
	}

	// Penalty for missing critical fields
	missingCriticalPenalty := cs.calculateMissingCriticalPenalty(result)
	penalties["missing_critical"] = missingCriticalPenalty

	// Penalty for low confidence fields
	lowConfidencePenalty := cs.calculateLowConfidencePenalty(result)
	penalties["low_confidence"] = lowConfidencePenalty

	return penalties
}

// calculateMissingCriticalPenalty calculates penalty for missing critical fields
func (cs *ConfidenceScorer) calculateMissingCriticalPenalty(result *VerificationResult) float64 {
	criticalFields := []string{"business_name", "phone_numbers", "addresses"}
	missingCount := 0

	for _, fieldName := range criticalFields {
		if fieldResult, exists := result.FieldResults[fieldName]; !exists || fieldResult.Score == 0 {
			missingCount++
		}
	}

	return float64(missingCount) * 0.1
}

// calculateLowConfidencePenalty calculates penalty for fields with low confidence
func (cs *ConfidenceScorer) calculateLowConfidencePenalty(result *VerificationResult) float64 {
	lowConfidenceCount := 0

	for _, fieldResult := range result.FieldResults {
		if fieldResult.Confidence < 0.5 {
			lowConfidenceCount++
		}
	}

	return float64(lowConfidenceCount) * 0.05
}

// calculateBonusFactors calculates bonus factors that increase confidence
func (cs *ConfidenceScorer) calculateBonusFactors(result *VerificationResult) map[string]float64 {
	bonuses := make(map[string]float64)

	// Bonus for passed status
	if result.Status == "PASSED" {
		bonuses["status_passed"] = 0.1
	}

	// Bonus for high confidence fields
	highConfidenceBonus := cs.calculateHighConfidenceBonus(result)
	bonuses["high_confidence"] = highConfidenceBonus

	// Bonus for comprehensive verification
	comprehensiveBonus := cs.calculateComprehensiveBonus(result)
	bonuses["comprehensive"] = comprehensiveBonus

	return bonuses
}

// calculateHighConfidenceBonus calculates bonus for fields with high confidence
func (cs *ConfidenceScorer) calculateHighConfidenceBonus(result *VerificationResult) float64 {
	highConfidenceCount := 0

	for _, fieldResult := range result.FieldResults {
		if fieldResult.Confidence >= 0.8 {
			highConfidenceCount++
		}
	}

	return float64(highConfidenceCount) * 0.02
}

// calculateComprehensiveBonus calculates bonus for comprehensive verification
func (cs *ConfidenceScorer) calculateComprehensiveBonus(result *VerificationResult) float64 {
	fieldCount := len(result.FieldResults)
	if fieldCount >= 5 {
		return 0.05
	} else if fieldCount >= 3 {
		return 0.03
	}
	return 0.0
}

// normalizeScores normalizes scores with confidence factors, penalties, and bonuses
func (cs *ConfidenceScorer) normalizeScores(fieldScores map[string]float64, confidenceFactors, penaltyFactors, bonusFactors map[string]float64) map[string]float64 {
	normalizedScores := make(map[string]float64)

	for fieldName, score := range fieldScores {
		normalizedScore := score

		// Apply confidence factors
		for _, factor := range confidenceFactors {
			normalizedScore *= (1.0 + factor*0.1)
		}

		// Apply penalties
		for _, penalty := range penaltyFactors {
			normalizedScore *= (1.0 - penalty)
		}

		// Apply bonuses
		for _, bonus := range bonusFactors {
			normalizedScore *= (1.0 + bonus)
		}

		// Ensure score is within 0-1 range
		normalizedScores[fieldName] = math.Max(0.0, math.Min(1.0, normalizedScore))
	}

	return normalizedScores
}

// calculateOverallScore calculates the overall confidence score
func (cs *ConfidenceScorer) calculateOverallScore(normalizedScores, weightedScores map[string]float64) float64 {
	if len(weightedScores) == 0 {
		return 0.0
	}

	totalWeightedScore := 0.0
	totalWeight := 0.0

	for fieldName, weightedScore := range weightedScores {
		totalWeightedScore += weightedScore
		totalWeight += cs.config.FieldWeights[fieldName]
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalWeightedScore / totalWeight
}

// determineConfidenceLevel determines the confidence level based on the overall score
func (cs *ConfidenceScorer) determineConfidenceLevel(score float64) ConfidenceLevel {
	if score >= cs.config.ConfidenceThresholds.HighThreshold {
		return ConfidenceLevelHigh
	} else if score >= cs.config.ConfidenceThresholds.MediumThreshold {
		return ConfidenceLevelMedium
	} else if score >= cs.config.ConfidenceThresholds.LowThreshold {
		return ConfidenceLevelLow
	}
	return ConfidenceLevelLow
}

// getCalibrationData retrieves calibration data for the given status
func (cs *ConfidenceScorer) getCalibrationData(status string) CalibrationData {
	if data, exists := cs.calibrationData[status]; exists {
		return data
	}

	// Return default calibration data
	return CalibrationData{
		SampleSize:         0,
		AverageScore:       0.5,
		StandardDeviation:  0.2,
		ConfidenceInterval: 0.1,
		LastCalibrated:     time.Now(),
	}
}

// UpdateCalibrationData updates calibration data with new verification results
func (cs *ConfidenceScorer) UpdateCalibrationData(ctx context.Context, status string, scores []float64) error {
	if len(scores) == 0 {
		return fmt.Errorf("no scores provided for calibration")
	}

	// Calculate statistics
	mean := 0.0
	for _, score := range scores {
		mean += score
	}
	mean /= float64(len(scores))

	variance := 0.0
	for _, score := range scores {
		variance += math.Pow(score-mean, 2)
	}
	variance /= float64(len(scores))

	stdDev := math.Sqrt(variance)

	// Calculate confidence interval (95% confidence level)
	confidenceInterval := 1.96 * stdDev / math.Sqrt(float64(len(scores)))

	// Update calibration data
	cs.calibrationData[status] = CalibrationData{
		SampleSize:         len(scores),
		AverageScore:       mean,
		StandardDeviation:  stdDev,
		ConfidenceInterval: confidenceInterval,
		LastCalibrated:     time.Now(),
	}

	return nil
}

// ValidateConfidenceScore validates if a confidence score is reasonable
func (cs *ConfidenceScorer) ValidateConfidenceScore(score *ConfidenceScore) error {
	if score == nil {
		return fmt.Errorf("confidence score cannot be nil")
	}

	// Validate overall score range
	if score.OverallScore < 0 || score.OverallScore > 1 {
		return fmt.Errorf("overall score must be between 0 and 1, got %f", score.OverallScore)
	}

	// Validate field scores
	for fieldName, fieldScore := range score.FieldScores {
		if fieldScore < 0 || fieldScore > 1 {
			return fmt.Errorf("field score for %s must be between 0 and 1, got %f", fieldName, fieldScore)
		}
	}

	// Validate confidence level
	if score.ConfidenceLevel != ConfidenceLevelHigh &&
		score.ConfidenceLevel != ConfidenceLevelMedium &&
		score.ConfidenceLevel != ConfidenceLevelLow {
		return fmt.Errorf("invalid confidence level: %s", score.ConfidenceLevel)
	}

	return nil
}

// GetConfig returns the current configuration
func (cs *ConfidenceScorer) GetConfig() ConfidenceScorerConfig {
	return cs.config
}

// UpdateConfig updates the confidence scorer configuration
func (cs *ConfidenceScorer) UpdateConfig(config ConfidenceScorerConfig) error {
	// Validate configuration
	if err := cs.validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cs.config = config
	return nil
}

// validateConfig validates the confidence scorer configuration
func (cs *ConfidenceScorer) validateConfig(config ConfidenceScorerConfig) error {
	// Validate field weights
	totalWeight := 0.0
	for fieldName, weight := range config.FieldWeights {
		if weight < 0 {
			return fmt.Errorf("field weight for %s cannot be negative", fieldName)
		}
		totalWeight += weight
	}

	if totalWeight == 0 {
		return fmt.Errorf("at least one field must have a positive weight")
	}

	// Validate confidence thresholds
	if config.ConfidenceThresholds.HighThreshold <= config.ConfidenceThresholds.MediumThreshold {
		return fmt.Errorf("high threshold must be greater than medium threshold")
	}

	if config.ConfidenceThresholds.MediumThreshold <= config.ConfidenceThresholds.LowThreshold {
		return fmt.Errorf("medium threshold must be greater than low threshold")
	}

	if config.ConfidenceThresholds.HighThreshold > 1.0 ||
		config.ConfidenceThresholds.MediumThreshold > 1.0 ||
		config.ConfidenceThresholds.LowThreshold > 1.0 {
		return fmt.Errorf("confidence thresholds must be between 0 and 1")
	}

	return nil
}

// GetCalibrationData returns calibration data for all statuses
func (cs *ConfidenceScorer) GetCalibrationData() map[string]CalibrationData {
	return cs.calibrationData
}

// GetStatistics returns statistics about the confidence scorer
func (cs *ConfidenceScorer) GetStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Configuration statistics
	stats["field_count"] = len(cs.config.FieldWeights)
	stats["algorithm"] = cs.config.ScoringAlgorithm

	// Calibration statistics
	stats["calibration_statuses"] = len(cs.calibrationData)

	// Threshold statistics
	stats["high_threshold"] = cs.config.ConfidenceThresholds.HighThreshold
	stats["medium_threshold"] = cs.config.ConfidenceThresholds.MediumThreshold
	stats["low_threshold"] = cs.config.ConfidenceThresholds.LowThreshold

	return stats
}
