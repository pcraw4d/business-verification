package external

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfidenceScorer(t *testing.T) {
	scorer := NewConfidenceScorer()

	assert.NotNil(t, scorer)
	assert.NotNil(t, scorer.config.FieldWeights)
	assert.Equal(t, "weighted_average", scorer.config.ScoringAlgorithm)
	assert.Equal(t, 0.8, scorer.config.ConfidenceThresholds.HighThreshold)
	assert.Equal(t, 0.6, scorer.config.ConfidenceThresholds.MediumThreshold)
	assert.Equal(t, 0.4, scorer.config.ConfidenceThresholds.LowThreshold)
}

func TestNewConfidenceScorerWithConfig(t *testing.T) {
	config := ConfidenceScorerConfig{
		FieldWeights: map[string]float64{
			"test_field": 0.5,
		},
		ConfidenceThresholds: ConfidenceThresholds{
			HighThreshold:   0.9,
			MediumThreshold: 0.7,
			LowThreshold:    0.5,
		},
		ScoringAlgorithm: "custom",
	}

	scorer := NewConfidenceScorerWithConfig(config)

	assert.NotNil(t, scorer)
	assert.Equal(t, config, scorer.config)
}

func TestConfidenceScorer_CalculateConfidenceScore(t *testing.T) {
	scorer := NewConfidenceScorer()

	tests := []struct {
		name             string
		result           *VerificationResult
		expectedLevel    ConfidenceLevel
		expectedMinScore float64
	}{
		{
			name: "high confidence verification",
			result: &VerificationResult{
				ID:     "test-123",
				Status: "PASSED",
				FieldResults: map[string]FieldResult{
					"business_name": {
						Score:      0.9,
						Confidence: 0.95,
						Matched:    true,
					},
					"phone_numbers": {
						Score:      0.85,
						Confidence: 0.9,
						Matched:    true,
					},
					"addresses": {
						Score:      0.8,
						Confidence: 0.85,
						Matched:    true,
					},
				},
			},
			expectedLevel:    ConfidenceLevelHigh,
			expectedMinScore: 0.8,
		},
		{
			name: "medium confidence verification",
			result: &VerificationResult{
				ID:     "test-456",
				Status: "PARTIAL",
				FieldResults: map[string]FieldResult{
					"business_name": {
						Score:      0.65,
						Confidence: 0.7,
						Matched:    true,
					},
					"phone_numbers": {
						Score:      0.55,
						Confidence: 0.6,
						Matched:    false,
					},
				},
			},
			expectedLevel:    ConfidenceLevelMedium,
			expectedMinScore: 0.6,
		},
		{
			name: "low confidence verification",
			result: &VerificationResult{
				ID:     "test-789",
				Status: "FAILED",
				FieldResults: map[string]FieldResult{
					"business_name": {
						Score:      0.3,
						Confidence: 0.4,
						Matched:    false,
					},
				},
			},
			expectedLevel:    ConfidenceLevelLow,
			expectedMinScore: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, err := scorer.CalculateConfidenceScore(context.Background(), tt.result)

			require.NoError(t, err)
			assert.NotNil(t, score)
			assert.Equal(t, tt.expectedLevel, score.ConfidenceLevel)
			assert.GreaterOrEqual(t, score.OverallScore, tt.expectedMinScore)
			assert.LessOrEqual(t, score.OverallScore, 1.0)
			assert.NotNil(t, score.FieldScores)
			assert.NotNil(t, score.WeightedScores)
			assert.NotNil(t, score.ScoreBreakdown)
		})
	}
}

func TestConfidenceScorer_CalculateConfidenceScore_NilResult(t *testing.T) {
	scorer := NewConfidenceScorer()

	score, err := scorer.CalculateConfidenceScore(context.Background(), nil)

	assert.Error(t, err)
	assert.Nil(t, score)
	assert.Contains(t, err.Error(), "verification result cannot be nil")
}

func TestConfidenceScorer_calculateFieldScore(t *testing.T) {
	scorer := NewConfidenceScorer()

	tests := []struct {
		name             string
		fieldResult      FieldResult
		expectedMinScore float64
		expectedMaxScore float64
	}{
		{
			name: "high score with high confidence and match",
			fieldResult: FieldResult{
				Score:      0.9,
				Confidence: 0.95,
				Matched:    true,
			},
			expectedMinScore: 0.9,
			expectedMaxScore: 1.0,
		},
		{
			name: "medium score with medium confidence and no match",
			fieldResult: FieldResult{
				Score:      0.6,
				Confidence: 0.7,
				Matched:    false,
			},
			expectedMinScore: 0.6,
			expectedMaxScore: 0.8,
		},
		{
			name: "low score with low confidence",
			fieldResult: FieldResult{
				Score:      0.3,
				Confidence: 0.4,
				Matched:    false,
			},
			expectedMinScore: 0.3,
			expectedMaxScore: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateFieldScore(tt.fieldResult)

			assert.GreaterOrEqual(t, score, tt.expectedMinScore)
			assert.LessOrEqual(t, score, tt.expectedMaxScore)
			assert.GreaterOrEqual(t, score, 0.0)
			assert.LessOrEqual(t, score, 1.0)
		})
	}
}

func TestConfidenceScorer_applyFieldWeights(t *testing.T) {
	scorer := NewConfidenceScorer()

	fieldScores := map[string]float64{
		"business_name": 0.8,
		"phone_numbers": 0.7,
		"unknown_field": 0.6,
	}

	weightedScores := scorer.applyFieldWeights(fieldScores)

	assert.InDelta(t, 0.8*0.25, weightedScores["business_name"], 0.001)
	assert.InDelta(t, 0.7*0.20, weightedScores["phone_numbers"], 0.001)
	assert.InDelta(t, 0.6*0.1, weightedScores["unknown_field"], 0.001) // Default weight
}

func TestConfidenceScorer_calculateConfidenceFactors(t *testing.T) {
	scorer := NewConfidenceScorer()

	result := &VerificationResult{
		FieldResults: map[string]FieldResult{
			"business_name": {Score: 0.8, Confidence: 0.9},
			"phone_numbers": {Score: 0.7, Confidence: 0.8},
			"addresses":     {Score: 0.0, Confidence: 0.0}, // Missing field
		},
	}

	factors := scorer.calculateConfidenceFactors(result)

	assert.NotNil(t, factors)
	assert.Contains(t, factors, "data_completeness")
	assert.Contains(t, factors, "data_consistency")
	assert.Contains(t, factors, "source_reliability")

	// Data completeness should be 2/3 = 0.67 (2 populated fields out of 3)
	assert.InDelta(t, 0.67, factors["data_completeness"], 0.01)
}

func TestConfidenceScorer_calculateCompletenessFactor(t *testing.T) {
	scorer := NewConfidenceScorer()

	tests := []struct {
		name           string
		result         *VerificationResult
		expectedFactor float64
	}{
		{
			name: "all fields populated",
			result: &VerificationResult{
				FieldResults: map[string]FieldResult{
					"field1": {Score: 0.8},
					"field2": {Score: 0.7},
				},
			},
			expectedFactor: 1.0,
		},
		{
			name: "half fields populated",
			result: &VerificationResult{
				FieldResults: map[string]FieldResult{
					"field1": {Score: 0.8},
					"field2": {Score: 0.0},
				},
			},
			expectedFactor: 0.5,
		},
		{
			name: "no fields populated",
			result: &VerificationResult{
				FieldResults: map[string]FieldResult{
					"field1": {Score: 0.0},
					"field2": {Score: 0.0},
				},
			},
			expectedFactor: 0.0,
		},
		{
			name: "empty result",
			result: &VerificationResult{
				FieldResults: map[string]FieldResult{},
			},
			expectedFactor: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factor := scorer.calculateCompletenessFactor(tt.result)
			assert.Equal(t, tt.expectedFactor, factor)
		})
	}
}

func TestConfidenceScorer_calculateConsistencyFactor(t *testing.T) {
	scorer := NewConfidenceScorer()

	tests := []struct {
		name              string
		result            *VerificationResult
		expectedMinFactor float64
	}{
		{
			name: "consistent scores",
			result: &VerificationResult{
				FieldResults: map[string]FieldResult{
					"field1": {Score: 0.8},
					"field2": {Score: 0.8},
					"field3": {Score: 0.8},
				},
			},
			expectedMinFactor: 0.9, // High consistency
		},
		{
			name: "inconsistent scores",
			result: &VerificationResult{
				FieldResults: map[string]FieldResult{
					"field1": {Score: 0.9},
					"field2": {Score: 0.1},
					"field3": {Score: 0.8},
				},
			},
			expectedMinFactor: 0.0, // Lower consistency
		},
		{
			name: "single field",
			result: &VerificationResult{
				FieldResults: map[string]FieldResult{
					"field1": {Score: 0.8},
				},
			},
			expectedMinFactor: 1.0, // Single field is always consistent
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			factor := scorer.calculateConsistencyFactor(tt.result)
			assert.GreaterOrEqual(t, factor, tt.expectedMinFactor)
			assert.LessOrEqual(t, factor, 1.0)
		})
	}
}

func TestConfidenceScorer_calculatePenaltyFactors(t *testing.T) {
	scorer := NewConfidenceScorer()

	tests := []struct {
		name              string
		result            *VerificationResult
		expectedPenalties []string
	}{
		{
			name: "failed status",
			result: &VerificationResult{
				Status: "FAILED",
				FieldResults: map[string]FieldResult{
					"business_name": {Score: 0.8},
				},
			},
			expectedPenalties: []string{"status_failed"},
		},
		{
			name: "missing critical fields",
			result: &VerificationResult{
				Status: "PASSED",
				FieldResults: map[string]FieldResult{
					"business_name": {Score: 0.0}, // Missing
					"phone_numbers": {Score: 0.0}, // Missing
					"addresses":     {Score: 0.8},
				},
			},
			expectedPenalties: []string{"missing_critical"},
		},
		{
			name: "low confidence fields",
			result: &VerificationResult{
				Status: "PASSED",
				FieldResults: map[string]FieldResult{
					"business_name": {Score: 0.8, Confidence: 0.3}, // Low confidence
					"phone_numbers": {Score: 0.7, Confidence: 0.4}, // Low confidence
				},
			},
			expectedPenalties: []string{"low_confidence"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			penalties := scorer.calculatePenaltyFactors(tt.result)

			for _, expectedPenalty := range tt.expectedPenalties {
				assert.Contains(t, penalties, expectedPenalty)
			}
		})
	}
}

func TestConfidenceScorer_calculateBonusFactors(t *testing.T) {
	scorer := NewConfidenceScorer()

	tests := []struct {
		name            string
		result          *VerificationResult
		expectedBonuses []string
	}{
		{
			name: "passed status",
			result: &VerificationResult{
				Status: "PASSED",
				FieldResults: map[string]FieldResult{
					"business_name": {Score: 0.8},
				},
			},
			expectedBonuses: []string{"status_passed"},
		},
		{
			name: "high confidence fields",
			result: &VerificationResult{
				Status: "PASSED",
				FieldResults: map[string]FieldResult{
					"business_name": {Score: 0.8, Confidence: 0.9},  // High confidence
					"phone_numbers": {Score: 0.7, Confidence: 0.85}, // High confidence
				},
			},
			expectedBonuses: []string{"high_confidence"},
		},
		{
			name: "comprehensive verification",
			result: &VerificationResult{
				Status: "PASSED",
				FieldResults: map[string]FieldResult{
					"field1": {Score: 0.8},
					"field2": {Score: 0.7},
					"field3": {Score: 0.6},
					"field4": {Score: 0.5},
					"field5": {Score: 0.4},
				},
			},
			expectedBonuses: []string{"comprehensive"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonuses := scorer.calculateBonusFactors(tt.result)

			for _, expectedBonus := range tt.expectedBonuses {
				assert.Contains(t, bonuses, expectedBonus)
			}
		})
	}
}

func TestConfidenceScorer_determineConfidenceLevel(t *testing.T) {
	scorer := NewConfidenceScorer()

	tests := []struct {
		name          string
		score         float64
		expectedLevel ConfidenceLevel
	}{
		{name: "high confidence", score: 0.9, expectedLevel: ConfidenceLevelHigh},
		{name: "high threshold boundary", score: 0.8, expectedLevel: ConfidenceLevelHigh},
		{name: "medium confidence", score: 0.7, expectedLevel: ConfidenceLevelMedium},
		{name: "medium threshold boundary", score: 0.6, expectedLevel: ConfidenceLevelMedium},
		{name: "low confidence", score: 0.5, expectedLevel: ConfidenceLevelLow},
		{name: "low threshold boundary", score: 0.4, expectedLevel: ConfidenceLevelLow},
		{name: "very low confidence", score: 0.3, expectedLevel: ConfidenceLevelLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := scorer.determineConfidenceLevel(tt.score)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

func TestConfidenceScorer_UpdateCalibrationData(t *testing.T) {
	scorer := NewConfidenceScorer()

	scores := []float64{0.8, 0.7, 0.9, 0.6, 0.85}

	err := scorer.UpdateCalibrationData(context.Background(), "PASSED", scores)

	assert.NoError(t, err)

	calibrationData := scorer.GetCalibrationData()
	assert.Contains(t, calibrationData, "PASSED")

	data := calibrationData["PASSED"]
	assert.Equal(t, 5, data.SampleSize)
	assert.InDelta(t, 0.77, data.AverageScore, 0.01) // (0.8+0.7+0.9+0.6+0.85)/5
	assert.Greater(t, data.StandardDeviation, 0.0)
	assert.Greater(t, data.ConfidenceInterval, 0.0)
}

func TestConfidenceScorer_UpdateCalibrationData_EmptyScores(t *testing.T) {
	scorer := NewConfidenceScorer()

	err := scorer.UpdateCalibrationData(context.Background(), "PASSED", []float64{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no scores provided for calibration")
}

func TestConfidenceScorer_ValidateConfidenceScore(t *testing.T) {
	scorer := NewConfidenceScorer()

	tests := []struct {
		name          string
		score         *ConfidenceScore
		expectedError string
	}{
		{
			name: "valid score",
			score: &ConfidenceScore{
				OverallScore:    0.8,
				ConfidenceLevel: ConfidenceLevelHigh,
				FieldScores: map[string]float64{
					"field1": 0.8,
					"field2": 0.7,
				},
			},
			expectedError: "",
		},
		{
			name:          "nil score",
			score:         nil,
			expectedError: "confidence score cannot be nil",
		},
		{
			name: "invalid overall score",
			score: &ConfidenceScore{
				OverallScore:    1.5,
				ConfidenceLevel: ConfidenceLevelHigh,
				FieldScores:     map[string]float64{},
			},
			expectedError: "overall score must be between 0 and 1",
		},
		{
			name: "invalid field score",
			score: &ConfidenceScore{
				OverallScore:    0.8,
				ConfidenceLevel: ConfidenceLevelHigh,
				FieldScores: map[string]float64{
					"field1": 1.5,
				},
			},
			expectedError: "field score for field1 must be between 0 and 1",
		},
		{
			name: "invalid confidence level",
			score: &ConfidenceScore{
				OverallScore:    0.8,
				ConfidenceLevel: "invalid",
				FieldScores:     map[string]float64{},
			},
			expectedError: "invalid confidence level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := scorer.ValidateConfidenceScore(tt.score)

			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}

func TestConfidenceScorer_UpdateConfig(t *testing.T) {
	scorer := NewConfidenceScorer()

	// Valid config
	config := ConfidenceScorerConfig{
		FieldWeights: map[string]float64{
			"test_field": 0.5,
		},
		ConfidenceThresholds: ConfidenceThresholds{
			HighThreshold:   0.9,
			MediumThreshold: 0.7,
			LowThreshold:    0.5,
		},
		ScoringAlgorithm: "custom",
	}

	err := scorer.UpdateConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, config, scorer.config)

	// Invalid config - negative weight
	invalidConfig := config
	invalidConfig.FieldWeights["test_field"] = -0.1

	err = scorer.UpdateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "field weight for test_field cannot be negative")

	// Invalid config - invalid thresholds
	invalidThresholdConfig := ConfidenceScorerConfig{
		FieldWeights: map[string]float64{
			"test_field": 0.5,
		},
		ConfidenceThresholds: ConfidenceThresholds{
			HighThreshold:   0.5,
			MediumThreshold: 0.7,
			LowThreshold:    0.3,
		},
		ScoringAlgorithm: "custom",
	}

	err = scorer.UpdateConfig(invalidThresholdConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "high threshold must be greater than medium threshold")
}

func TestConfidenceScorer_GetStatistics(t *testing.T) {
	scorer := NewConfidenceScorer()

	stats := scorer.GetStatistics()

	assert.NotNil(t, stats)
	assert.Contains(t, stats, "field_count")
	assert.Contains(t, stats, "algorithm")
	assert.Contains(t, stats, "calibration_statuses")
	assert.Contains(t, stats, "high_threshold")
	assert.Contains(t, stats, "medium_threshold")
	assert.Contains(t, stats, "low_threshold")

	assert.Equal(t, 6, stats["field_count"]) // Default field weights count
	assert.Equal(t, "weighted_average", stats["algorithm"])
	assert.Equal(t, 0, stats["calibration_statuses"])
}

func TestConfidenceScorer_GetConfig(t *testing.T) {
	scorer := NewConfidenceScorer()

	config := scorer.GetConfig()

	assert.NotNil(t, config)
	assert.NotNil(t, config.FieldWeights)
	assert.Equal(t, "weighted_average", config.ScoringAlgorithm)
	assert.Equal(t, 0.8, config.ConfidenceThresholds.HighThreshold)
}

func TestConfidenceScorer_GetCalibrationData(t *testing.T) {
	scorer := NewConfidenceScorer()

	// Initially empty
	calibrationData := scorer.GetCalibrationData()
	assert.Empty(t, calibrationData)

	// Add some calibration data
	scores := []float64{0.8, 0.7, 0.9}
	err := scorer.UpdateCalibrationData(context.Background(), "PASSED", scores)
	assert.NoError(t, err)

	calibrationData = scorer.GetCalibrationData()
	assert.Contains(t, calibrationData, "PASSED")
	assert.Equal(t, 3, calibrationData["PASSED"].SampleSize)
}
