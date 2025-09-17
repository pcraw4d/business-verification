package industry_codes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupTestConfidenceFilter(t *testing.T) (*ConfidenceFilter, func()) {
	db, cleanup := setupTestDatabase(t)

	database := NewIndustryCodeDatabase(db, zaptest.NewLogger(t))
	metadataManager := NewMetadataManager(db, zaptest.NewLogger(t))
	confidenceScorer := NewConfidenceScorer(database, metadataManager, zaptest.NewLogger(t))
	confidenceFilter := NewConfidenceFilter(confidenceScorer, zaptest.NewLogger(t))

	return confidenceFilter, cleanup
}

func TestConfidenceFilter_FilterByConfidence(t *testing.T) {
	confidenceFilter, cleanup := setupTestConfidenceFilter(t)
	defer cleanup()

	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	tests := []struct {
		name          string
		threshold     *ConfidenceThreshold
		expectedCount int
		expectedError bool
	}{
		{
			name:          "default threshold filtering",
			threshold:     nil,
			expectedCount: 1, // Only the highest confidence result passes
			expectedError: false,
		},
		{
			name: "high threshold filtering",
			threshold: &ConfidenceThreshold{
				GlobalMinConfidence: 0.7,
			},
			expectedCount: 0, // No results meet the high threshold
			expectedError: false,
		},
		{
			name: "low threshold filtering",
			threshold: &ConfidenceThreshold{
				GlobalMinConfidence: 0.1,
			},
			expectedCount: 4, // Should keep all results
			expectedError: false,
		},
		{
			name: "type-specific threshold filtering",
			threshold: &ConfidenceThreshold{
				GlobalMinConfidence: 0.3,
				TypeSpecificThresholds: map[string]float64{
					"sic": 0.5, // Higher threshold for SIC codes
				},
			},
			expectedCount: 1, // Only the NAICS result should pass
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filteringResult, filteredResults, err := confidenceFilter.FilterByConfidence(context.Background(), results, request, tt.threshold)

			if tt.expectedError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, filteringResult)
			assert.NotNil(t, filteredResults)
			assert.Equal(t, tt.expectedCount, len(filteredResults))
			assert.Equal(t, tt.expectedCount, filteringResult.FilteredCount)
			assert.Equal(t, len(results)-tt.expectedCount, filteringResult.RejectedCount)
			assert.Equal(t, len(results), filteringResult.OriginalCount)

			// Verify all filtered results meet the threshold
			if tt.threshold != nil {
				for _, result := range filteredResults {
					assert.GreaterOrEqual(t, result.Confidence, tt.threshold.GlobalMinConfidence)
				}
			}
		})
	}
}

func TestConfidenceFilter_AdaptiveThresholds(t *testing.T) {
	confidenceFilter, cleanup := setupTestConfidenceFilter(t)
	defer cleanup()

	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	threshold := &ConfidenceThreshold{
		GlobalMinConfidence: 0.3,
		AdaptiveThresholds: &AdaptiveThreshold{
			Enabled:           true,
			BaseThreshold:     0.3,
			QualityMultiplier: 0.1,
			VolumeMultiplier:  0.05,
			MaxThreshold:      0.8,
			MinThreshold:      0.1,
		},
	}

	filteringResult, filteredResults, err := confidenceFilter.FilterByConfidence(context.Background(), results, request, threshold)

	require.NoError(t, err)
	assert.NotNil(t, filteringResult)
	assert.NotNil(t, filteredResults)
	assert.NotEmpty(t, filteringResult.AdaptiveAdjustments)

	// Verify adaptive adjustments were applied
	for _, adjustment := range filteringResult.AdaptiveAdjustments {
		assert.NotEqual(t, adjustment.OriginalThreshold, adjustment.AdjustedThreshold)
		assert.NotZero(t, adjustment.AdjustmentFactor)
		assert.NotEmpty(t, adjustment.Reason)
	}
}

func TestConfidenceFilter_QualityBasedThresholds(t *testing.T) {
	confidenceFilter, cleanup := setupTestConfidenceFilter(t)
	defer cleanup()

	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	threshold := &ConfidenceThreshold{
		GlobalMinConfidence: 0.3,
		QualityBasedThresholds: &QualityThreshold{
			Enabled: true,
			QualityThresholds: map[string]float64{
				"high":   0.2,
				"medium": 0.4,
				"low":    0.6,
			},
		},
	}

	filteringResult, filteredResults, err := confidenceFilter.FilterByConfidence(context.Background(), results, request, threshold)

	require.NoError(t, err)
	assert.NotNil(t, filteringResult)
	assert.NotNil(t, filteredResults)
	assert.NotNil(t, filteringResult.QualityAnalysis)

	// Verify quality analysis
	qualityAnalysis := filteringResult.QualityAnalysis
	assert.GreaterOrEqual(t, qualityAnalysis.HighQualityCount, 0)
	assert.GreaterOrEqual(t, qualityAnalysis.MediumQualityCount, 0)
	assert.GreaterOrEqual(t, qualityAnalysis.LowQualityCount, 0)
	assert.Greater(t, qualityAnalysis.AverageQuality, 0.0)
	assert.Greater(t, qualityAnalysis.QualityScore, 0.0)
}

func TestConfidenceFilter_FilteringMetrics(t *testing.T) {
	confidenceFilter, cleanup := setupTestConfidenceFilter(t)
	defer cleanup()

	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	filteringResult, filteredResults, err := confidenceFilter.FilterByConfidence(context.Background(), results, request, nil)

	require.NoError(t, err)
	assert.NotNil(t, filteringResult)
	assert.NotNil(t, filteringResult.FilteringMetrics)

	metrics := filteringResult.FilteringMetrics
	assert.Equal(t, len(results), filteringResult.OriginalCount)
	assert.Equal(t, len(filteredResults), filteringResult.FilteredCount)
	assert.Greater(t, metrics.FilteringTime, time.Duration(0))
	assert.Greater(t, metrics.AverageConfidenceBefore, 0.0)
	assert.Greater(t, metrics.AverageConfidenceAfter, 0.0)
	assert.Greater(t, metrics.ThresholdEffectiveness, 0.0)
	assert.LessOrEqual(t, metrics.ThresholdEffectiveness, 1.0)

	// Verify distributions
	assert.NotEmpty(t, metrics.ConfidenceDistribution)
	assert.NotEmpty(t, metrics.TypeDistribution)
	assert.NotEmpty(t, metrics.QualityDistribution)
}

func TestConfidenceFilter_RejectionReasons(t *testing.T) {
	confidenceFilter, cleanup := setupTestConfidenceFilter(t)
	defer cleanup()

	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	// Use a high threshold to ensure some results are rejected
	threshold := &ConfidenceThreshold{
		GlobalMinConfidence: 0.8,
	}

	filteringResult, filteredResults, err := confidenceFilter.FilterByConfidence(context.Background(), results, request, threshold)

	require.NoError(t, err)
	assert.NotNil(t, filteringResult)
	assert.Len(t, filteredResults, 0) // No results should pass the high threshold

	// Verify rejection reasons
	assert.NotEmpty(t, filteringResult.RejectionReasons)
	for _, reasons := range filteringResult.RejectionReasons {
		assert.NotEmpty(t, reasons)
		for _, reason := range reasons {
			assert.Contains(t, reason, "confidence")
			assert.Contains(t, reason, "below threshold")
		}
	}
}

func TestConfidenceFilter_ValidationRules(t *testing.T) {
	confidenceFilter, cleanup := setupTestConfidenceFilter(t)
	defer cleanup()

	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	threshold := &ConfidenceThreshold{
		GlobalMinConfidence: 0.3,
		ValidationRules: []ThresholdRule{
			{
				Name:        "test_minimum",
				Type:        "minimum",
				Value:       0.2,
				Description: "Test minimum rule",
			},
			{
				Name:        "test_maximum",
				Type:        "maximum",
				Value:       0.9,
				Description: "Test maximum rule",
			},
		},
	}

	filteringResult, filteredResults, err := confidenceFilter.FilterByConfidence(context.Background(), results, request, threshold)

	require.NoError(t, err)
	assert.NotNil(t, filteringResult)
	assert.NotNil(t, filteredResults)
	assert.GreaterOrEqual(t, filteringResult.ThresholdUsed, 0.2)
	assert.LessOrEqual(t, filteringResult.ThresholdUsed, 0.9)
}

func TestConfidenceFilter_EmptyResults(t *testing.T) {
	confidenceFilter, cleanup := setupTestConfidenceFilter(t)
	defer cleanup()

	results := []*ClassificationResult{}
	request := createTestClassificationRequest()

	filteringResult, filteredResults, err := confidenceFilter.FilterByConfidence(context.Background(), results, request, nil)

	require.NoError(t, err)
	assert.NotNil(t, filteringResult)
	assert.Empty(t, filteredResults)
	assert.Equal(t, 0, filteringResult.OriginalCount)
	assert.Equal(t, 0, filteringResult.FilteredCount)
	assert.Equal(t, 0, filteringResult.RejectedCount)
}

func TestConfidenceFilter_GetSetDefaultThreshold(t *testing.T) {
	confidenceFilter, cleanup := setupTestConfidenceFilter(t)
	defer cleanup()

	// Get default threshold
	defaultThreshold := confidenceFilter.GetDefaultThreshold()
	assert.NotNil(t, defaultThreshold)
	assert.Equal(t, 0.3, defaultThreshold.GlobalMinConfidence)

	// Set new threshold
	newThreshold := &ConfidenceThreshold{
		GlobalMinConfidence: 0.5,
		TypeSpecificThresholds: map[string]float64{
			"mcc": 0.4,
		},
	}

	confidenceFilter.SetDefaultThreshold(newThreshold)

	// Verify threshold was updated
	updatedThreshold := confidenceFilter.GetDefaultThreshold()
	assert.Equal(t, 0.5, updatedThreshold.GlobalMinConfidence)
	assert.Equal(t, 0.4, updatedThreshold.TypeSpecificThresholds["mcc"])
}

func TestConfidenceFilter_Integration(t *testing.T) {
	confidenceFilter, cleanup := setupTestConfidenceFilter(t)
	defer cleanup()

	results := createTestClassificationResults()
	request := createTestClassificationRequest()

	// Test with comprehensive threshold configuration
	threshold := &ConfidenceThreshold{
		GlobalMinConfidence: 0.3,
		TypeSpecificThresholds: map[string]float64{
			"sic":   0.4,
			"naics": 0.3,
			"mcc":   0.2,
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
			Enabled: true,
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
		},
	}

	filteringResult, filteredResults, err := confidenceFilter.FilterByConfidence(context.Background(), results, request, threshold)

	require.NoError(t, err)
	assert.NotNil(t, filteringResult)
	assert.NotNil(t, filteredResults)
	assert.NotNil(t, filteringResult.FilteringMetrics)
	assert.NotNil(t, filteringResult.QualityAnalysis)
	assert.NotEmpty(t, filteringResult.AdaptiveAdjustments)

	// Verify all components are working together
	assert.Greater(t, filteringResult.FilteredCount, 0)
	assert.Greater(t, filteringResult.FilteringMetrics.AverageConfidenceAfter, 0.0)
	assert.Greater(t, filteringResult.QualityAnalysis.AverageQuality, 0.0)
}
