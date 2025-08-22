package enrichment

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewConfidenceScorer(t *testing.T) {
	tests := []struct {
		name   string
		config *ConfidenceConfig
		logger *zap.Logger
	}{
		{
			name:   "with nil inputs",
			config: nil,
			logger: nil,
		},
		{
			name: "with custom config",
			config: &ConfidenceConfig{
				DataQualityWeight:       0.3,
				ConsistencyWeight:       0.25,
				ValidationWeight:        0.2,
				EvidenceWeight:          0.15,
				FreshnessWeight:         0.05,
				SourceReliabilityWeight: 0.05,
				MinConfidenceThreshold:  0.4,
				MaxConfidenceThreshold:  0.95,
				HighConfidenceThreshold: 0.85,
			},
			logger: zap.NewNop(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer := NewConfidenceScorer(tt.config, tt.logger)
			assert.NotNil(t, scorer)
			assert.NotNil(t, scorer.config)
			assert.NotNil(t, scorer.logger)
			assert.NotNil(t, scorer.tracer)
		})
	}
}

func TestConfidenceScorer_CalculateConfidence(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name                  string
		result                *CompanySizeResult
		expectedMinConfidence float64
		expectedMaxConfidence float64
		expectedLevel         string
	}{
		{
			name: "high quality result",
			result: &CompanySizeResult{
				CompanySize:     "sme",
				ConfidenceScore: 0.8,
				EmployeeAnalysis: &EmployeeCountResult{
					EmployeeCount:    100,
					ConfidenceScore:  0.9,
					ExtractionMethod: "direct_mention",
					IsValidated:      true,
					ExtractedAt:      time.Now(),
				},
				RevenueAnalysis: &RevenueResult{
					RevenueAmount:    5000000,
					ConfidenceScore:  0.85,
					ExtractionMethod: "direct_mention",
					IsValidated:      true,
					ExtractedAt:      time.Now(),
				},
				ConsistencyScore: 1.0,
				Evidence:         []string{"Employee count: 100", "Revenue: $5M"},
				IsValidated:      true,
				ValidationStatus: ValidationStatus{IsValid: true},
				ClassifiedAt:     time.Now(),
				SourceURL:        "https://company.com",
			},
			expectedMinConfidence: 0.6,
			expectedMaxConfidence: 1.0,
			expectedLevel:         "medium",
		},
		{
			name: "medium quality result",
			result: &CompanySizeResult{
				CompanySize:     "startup",
				ConfidenceScore: 0.6,
				EmployeeAnalysis: &EmployeeCountResult{
					EmployeeCount:    25,
					ConfidenceScore:  0.7,
					ExtractionMethod: "linkedin_style",
					IsValidated:      false,
					ExtractedAt:      time.Now().Add(-7 * 24 * time.Hour),
				},
				ConsistencyScore: 0.8,
				Evidence:         []string{"Employee count: 25"},
				IsValidated:      false,
				ValidationStatus: ValidationStatus{IsValid: false},
				ClassifiedAt:     time.Now().Add(-7 * 24 * time.Hour),
				SourceURL:        "https://startup.com",
			},
			expectedMinConfidence: 0.4,
			expectedMaxConfidence: 0.8,
			expectedLevel:         "medium",
		},
		{
			name: "low quality result",
			result: &CompanySizeResult{
				CompanySize:      "unknown",
				ConfidenceScore:  0.3,
				ConsistencyScore: 0.4,
				Evidence:         []string{},
				IsValidated:      false,
				ValidationStatus: ValidationStatus{IsValid: false, ValidationErrors: []string{"No data"}},
				ClassifiedAt:     time.Now().Add(-90 * 24 * time.Hour),
				SourceURL:        "http://unknown.com",
			},
			expectedMinConfidence: 0.2,
			expectedMaxConfidence: 0.6,
			expectedLevel:         "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidenceScore, err := scorer.CalculateConfidence(context.Background(), tt.result)
			require.NoError(t, err)
			assert.NotNil(t, confidenceScore)

			// Check overall confidence
			assert.True(t, confidenceScore.OverallConfidence >= tt.expectedMinConfidence &&
				confidenceScore.OverallConfidence <= tt.expectedMaxConfidence,
				"Expected confidence between %f and %f, got %f",
				tt.expectedMinConfidence, tt.expectedMaxConfidence, confidenceScore.OverallConfidence)

			// Check confidence level
			assert.Equal(t, tt.expectedLevel, confidenceScore.ConfidenceLevel)

			// Check component scores
			assert.True(t, confidenceScore.DataQualityScore >= 0.0 && confidenceScore.DataQualityScore <= 1.0)
			assert.True(t, confidenceScore.ConsistencyScore >= 0.0 && confidenceScore.ConsistencyScore <= 1.0)
			assert.True(t, confidenceScore.ValidationScore >= 0.0 && confidenceScore.ValidationScore <= 1.0)
			assert.True(t, confidenceScore.EvidenceScore >= 0.0 && confidenceScore.EvidenceScore <= 1.0)
			assert.True(t, confidenceScore.FreshnessScore >= 0.0 && confidenceScore.FreshnessScore <= 1.0)
			assert.True(t, confidenceScore.SourceReliabilityScore >= 0.0 && confidenceScore.SourceReliabilityScore <= 1.0)

			// Check uncertainty quantification
			assert.True(t, confidenceScore.UncertaintyLevel >= 0.0 && confidenceScore.UncertaintyLevel <= 1.0)
			assert.InDelta(t, 1.0, confidenceScore.OverallConfidence+confidenceScore.UncertaintyLevel, 0.01)

			// Check confidence interval
			assert.True(t, confidenceScore.ConfidenceInterval.LowerBound >= 0.0)
			assert.True(t, confidenceScore.ConfidenceInterval.UpperBound <= 1.0)
			assert.True(t, confidenceScore.ConfidenceInterval.LowerBound <= confidenceScore.OverallConfidence)
			assert.True(t, confidenceScore.ConfidenceInterval.UpperBound >= confidenceScore.OverallConfidence)

			// Check component breakdown
			assert.NotEmpty(t, confidenceScore.ComponentBreakdown)
			assert.NotEmpty(t, confidenceScore.Factors)
			assert.NotEmpty(t, confidenceScore.Recommendations)

			// Check metadata
			assert.False(t, confidenceScore.CalculatedAt.IsZero())
		})
	}
}

func TestConfidenceScorer_CalculateDataQualityScore(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *CompanySizeResult
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "high quality employee and revenue data",
			result: &CompanySizeResult{
				EmployeeAnalysis: &EmployeeCountResult{
					ExtractionMethod: "direct_mention",
					ConfidenceScore:  0.9,
					IsValidated:      true,
				},
				RevenueAnalysis: &RevenueResult{
					ExtractionMethod: "direct_mention",
					ConfidenceScore:  0.85,
					IsValidated:      true,
				},
				DataQualityScore: 0.9,
			},
			expectedMin: 0.8,
			expectedMax: 1.0,
		},
		{
			name: "medium quality data",
			result: &CompanySizeResult{
				EmployeeAnalysis: &EmployeeCountResult{
					ExtractionMethod: "linkedin_style",
					ConfidenceScore:  0.7,
					IsValidated:      false,
				},
				DataQualityScore: 0.6,
			},
			expectedMin: 0.5,
			expectedMax: 0.8,
		},
		{
			name: "low quality data",
			result: &CompanySizeResult{
				EmployeeAnalysis: &EmployeeCountResult{
					ExtractionMethod: "size_keyword",
					ConfidenceScore:  0.4,
					IsValidated:      false,
				},
				DataQualityScore: 0.3,
			},
			expectedMin: 0.2,
			expectedMax: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateDataQualityScore(tt.result)
			assert.True(t, score >= tt.expectedMin && score <= tt.expectedMax,
				"Expected score between %f and %f, got %f", tt.expectedMin, tt.expectedMax, score)
		})
	}
}

func TestConfidenceScorer_CalculateConsistencyScore(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *CompanySizeResult
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "perfect consistency",
			result: &CompanySizeResult{
				ConsistencyScore:       1.0,
				EmployeeAnalysis:       &EmployeeCountResult{EmployeeCount: 100},
				RevenueAnalysis:        &RevenueResult{RevenueAmount: 5000000},
				EmployeeClassification: "sme",
				RevenueClassification:  "sme",
				Evidence:               []string{"Employee count: 100", "Revenue: $5M"},
			},
			expectedMin: 0.6,
			expectedMax: 1.0,
		},
		{
			name: "inconsistent classifications",
			result: &CompanySizeResult{
				ConsistencyScore:       0.7,
				EmployeeAnalysis:       &EmployeeCountResult{EmployeeCount: 25},
				RevenueAnalysis:        &RevenueResult{RevenueAmount: 5000000},
				EmployeeClassification: "startup",
				RevenueClassification:  "sme",
				Evidence:               []string{"Employee count: 25", "Revenue: $5M"},
			},
			expectedMin: 0.4,
			expectedMax: 0.7,
		},
		{
			name: "no consistency data",
			result: &CompanySizeResult{
				ConsistencyScore: 0.0,
				Evidence:         []string{},
			},
			expectedMin: 0.0,
			expectedMax: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateConsistencyScore(tt.result)
			assert.True(t, score >= tt.expectedMin && score <= tt.expectedMax,
				"Expected score between %f and %f, got %f", tt.expectedMin, tt.expectedMax, score)
		})
	}
}

func TestConfidenceScorer_CalculateValidationScore(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *CompanySizeResult
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "fully validated",
			result: &CompanySizeResult{
				IsValidated:      true,
				EmployeeAnalysis: &EmployeeCountResult{IsValidated: true},
				RevenueAnalysis:  &RevenueResult{IsValidated: true},
				ValidationStatus: ValidationStatus{IsValid: true},
			},
			expectedMin: 0.8,
			expectedMax: 1.0,
		},
		{
			name: "partially validated",
			result: &CompanySizeResult{
				IsValidated:      false,
				EmployeeAnalysis: &EmployeeCountResult{IsValidated: true},
				RevenueAnalysis:  &RevenueResult{IsValidated: false},
				ValidationStatus: ValidationStatus{IsValid: false, ValidationErrors: []string{}},
			},
			expectedMin: 0.5,
			expectedMax: 0.8,
		},
		{
			name: "not validated",
			result: &CompanySizeResult{
				IsValidated:      false,
				ValidationStatus: ValidationStatus{IsValid: false, ValidationErrors: []string{"Error"}},
			},
			expectedMin: 0.2,
			expectedMax: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateValidationScore(tt.result)
			assert.True(t, score >= tt.expectedMin && score <= tt.expectedMax,
				"Expected score between %f and %f, got %f", tt.expectedMin, tt.expectedMax, score)
		})
	}
}

func TestConfidenceScorer_CalculateEvidenceScore(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *CompanySizeResult
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "multiple high quality evidence",
			result: &CompanySizeResult{
				Evidence: []string{
					"Direct employee count: 100",
					"Direct revenue mention: $5M",
					"Financial indicator: profitable",
				},
			},
			expectedMin: 0.6,
			expectedMax: 1.0,
		},
		{
			name: "single evidence",
			result: &CompanySizeResult{
				Evidence: []string{"Employee count: 50"},
			},
			expectedMin: 0.2,
			expectedMax: 0.5,
		},
		{
			name: "no evidence",
			result: &CompanySizeResult{
				Evidence: []string{},
			},
			expectedMin: 0.0,
			expectedMax: 0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateEvidenceScore(tt.result)
			assert.True(t, score >= tt.expectedMin && score <= tt.expectedMax,
				"Expected score between %f and %f, got %f", tt.expectedMin, tt.expectedMax, score)
		})
	}
}

func TestConfidenceScorer_CalculateFreshnessScore(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *CompanySizeResult
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "very recent data",
			result: &CompanySizeResult{
				ClassifiedAt:     time.Now(),
				EmployeeAnalysis: &EmployeeCountResult{ExtractedAt: time.Now()},
				RevenueAnalysis:  &RevenueResult{ExtractedAt: time.Now()},
			},
			expectedMin: 0.9,
			expectedMax: 1.0,
		},
		{
			name: "recent data",
			result: &CompanySizeResult{
				ClassifiedAt:     time.Now().Add(-3 * 24 * time.Hour),
				EmployeeAnalysis: &EmployeeCountResult{ExtractedAt: time.Now().Add(-2 * 24 * time.Hour)},
			},
			expectedMin: 0.7,
			expectedMax: 0.9,
		},
		{
			name: "old data",
			result: &CompanySizeResult{
				ClassifiedAt:     time.Now().Add(-60 * 24 * time.Hour),
				EmployeeAnalysis: &EmployeeCountResult{ExtractedAt: time.Now().Add(-90 * 24 * time.Hour)},
			},
			expectedMin: 0.3,
			expectedMax: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateFreshnessScore(tt.result)
			assert.True(t, score >= tt.expectedMin && score <= tt.expectedMax,
				"Expected score between %f and %f, got %f", tt.expectedMin, tt.expectedMax, score)
		})
	}
}

func TestConfidenceScorer_CalculateSourceReliabilityScore(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name        string
		result      *CompanySizeResult
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "reliable sources",
			result: &CompanySizeResult{
				SourceURL:        "https://linkedin.com/company/example",
				EmployeeAnalysis: &EmployeeCountResult{SourceURL: "https://crunchbase.com/company/example"},
				RevenueAnalysis:  &RevenueResult{SourceURL: "https://company.com"},
			},
			expectedMin: 0.7,
			expectedMax: 0.9,
		},
		{
			name: "mixed sources",
			result: &CompanySizeResult{
				SourceURL:        "https://company.com",
				EmployeeAnalysis: &EmployeeCountResult{SourceURL: "http://unknown.com"},
			},
			expectedMin: 0.4,
			expectedMax: 0.7,
		},
		{
			name: "unreliable sources",
			result: &CompanySizeResult{
				SourceURL:        "http://unknown.com",
				EmployeeAnalysis: &EmployeeCountResult{SourceURL: "http://suspicious.com"},
			},
			expectedMin: 0.3,
			expectedMax: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateSourceReliabilityScore(tt.result)
			assert.True(t, score >= tt.expectedMin && score <= tt.expectedMax,
				"Expected score between %f and %f, got %f", tt.expectedMin, tt.expectedMax, score)
		})
	}
}

func TestConfidenceScorer_DetermineConfidenceLevel(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name          string
		confidence    float64
		expectedLevel string
	}{
		{
			name:          "very high confidence",
			confidence:    0.95,
			expectedLevel: "very_high",
		},
		{
			name:          "high confidence",
			confidence:    0.85,
			expectedLevel: "high",
		},
		{
			name:          "medium confidence",
			confidence:    0.7,
			expectedLevel: "medium",
		},
		{
			name:          "low confidence",
			confidence:    0.4,
			expectedLevel: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := scorer.determineConfidenceLevel(tt.confidence)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

func TestConfidenceScorer_CalculateConfidenceInterval(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name            string
		confidenceScore *ConfidenceScore
		expectedLevel   float64
	}{
		{
			name: "high confidence",
			confidenceScore: &ConfidenceScore{
				OverallConfidence: 0.9,
				UncertaintyLevel:  0.1,
			},
			expectedLevel: 0.95,
		},
		{
			name: "medium confidence",
			confidenceScore: &ConfidenceScore{
				OverallConfidence: 0.6,
				UncertaintyLevel:  0.4,
			},
			expectedLevel: 0.95,
		},
		{
			name: "low confidence",
			confidenceScore: &ConfidenceScore{
				OverallConfidence: 0.3,
				UncertaintyLevel:  0.7,
			},
			expectedLevel: 0.95,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interval := scorer.calculateConfidenceInterval(tt.confidenceScore)
			assert.Equal(t, tt.expectedLevel, interval.Level)
			assert.True(t, interval.LowerBound >= 0.0)
			assert.True(t, interval.UpperBound <= 1.0)
			assert.True(t, interval.LowerBound <= tt.confidenceScore.OverallConfidence)
			assert.True(t, interval.UpperBound >= tt.confidenceScore.OverallConfidence)
		})
	}
}

func TestConfidenceScorer_CalculateAnomalyScore(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name            string
		result          *CompanySizeResult
		confidenceScore *ConfidenceScore
		expectedMin     float64
		expectedMax     float64
	}{
		{
			name: "suspiciously high confidence with little evidence",
			result: &CompanySizeResult{
				Evidence: []string{"Single evidence"},
			},
			confidenceScore: &ConfidenceScore{OverallConfidence: 0.98},
			expectedMin:     0.2,
			expectedMax:     0.4,
		},
		{
			name: "unusual employee-to-revenue ratio",
			result: &CompanySizeResult{
				EmployeeAnalysis: &EmployeeCountResult{EmployeeCount: 10},
				RevenueAnalysis:  &RevenueResult{RevenueAmount: 20000000}, // $20M for 10 employees
			},
			confidenceScore: &ConfidenceScore{OverallConfidence: 0.7},
			expectedMin:     0.3,
			expectedMax:     0.5,
		},
		{
			name: "inconsistent classifications",
			result: &CompanySizeResult{
				EmployeeClassification: "startup",
				RevenueClassification:  "enterprise",
			},
			confidenceScore: &ConfidenceScore{OverallConfidence: 0.6},
			expectedMin:     0.1,
			expectedMax:     0.3,
		},
		{
			name: "normal data",
			result: &CompanySizeResult{
				EmployeeAnalysis:       &EmployeeCountResult{EmployeeCount: 100},
				RevenueAnalysis:        &RevenueResult{RevenueAmount: 5000000},
				EmployeeClassification: "sme",
				RevenueClassification:  "sme",
				Evidence:               []string{"Employee count: 100", "Revenue: $5M"},
			},
			confidenceScore: &ConfidenceScore{OverallConfidence: 0.8},
			expectedMin:     0.0,
			expectedMax:     0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := scorer.calculateAnomalyScore(tt.result, tt.confidenceScore)
			assert.True(t, score >= tt.expectedMin && score <= tt.expectedMax,
				"Expected score between %f and %f, got %f", tt.expectedMin, tt.expectedMax, score)
		})
	}
}

func TestConfidenceScorer_ApplyCalibration(t *testing.T) {
	config := &ConfidenceConfig{
		CalibrationFactor:      1.2,
		MinConfidenceThreshold: 0.3,
		MaxConfidenceThreshold: 1.0,
	}
	scorer := NewConfidenceScorer(config, zap.NewNop())

	tests := []struct {
		name               string
		confidenceScore    *ConfidenceScore
		expectedConfidence float64
	}{
		{
			name: "calibration applied",
			confidenceScore: &ConfidenceScore{
				OverallConfidence: 0.8,
				ConfidenceLevel:   "high",
			},
			expectedConfidence: 0.96, // 0.8 * 1.2 = 0.96, but capped at 1.0
		},
		{
			name: "calibration below minimum",
			confidenceScore: &ConfidenceScore{
				OverallConfidence: 0.2,
				ConfidenceLevel:   "low",
			},
			expectedConfidence: 0.3, // 0.2 * 1.2 = 0.24, but minimum is 0.3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scorer.applyCalibration(tt.confidenceScore)
			assert.Equal(t, tt.expectedConfidence, tt.confidenceScore.OverallConfidence)
			assert.True(t, tt.confidenceScore.CalibrationApplied)
			assert.Equal(t, config.CalibrationFactor, tt.confidenceScore.CalibrationFactor)
		})
	}
}

func TestConfidenceScorer_GenerateRecommendations(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	tests := []struct {
		name            string
		confidenceScore *ConfidenceScore
		result          *CompanySizeResult
		expectedCount   int
	}{
		{
			name: "low confidence with multiple issues",
			confidenceScore: &ConfidenceScore{
				OverallConfidence: 0.3,
				DataQualityScore:  0.4,
				ConsistencyScore:  0.5,
				EvidenceScore:     0.3,
				FreshnessScore:    0.4,
				AnomalyScore:      0.6,
			},
			result: &CompanySizeResult{
				Evidence: []string{"Single evidence"},
			},
			expectedCount: 12, // Multiple recommendations expected
		},
		{
			name: "high confidence with few issues",
			confidenceScore: &ConfidenceScore{
				OverallConfidence: 0.9,
				DataQualityScore:  0.8,
				ConsistencyScore:  0.9,
				EvidenceScore:     0.8,
				FreshnessScore:    0.9,
				AnomalyScore:      0.1,
			},
			result: &CompanySizeResult{
				Evidence: []string{"Evidence 1", "Evidence 2", "Evidence 3"},
			},
			expectedCount: 0, // No recommendations for high confidence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := scorer.generateRecommendations(tt.confidenceScore, tt.result)
			assert.Len(t, recommendations, tt.expectedCount)

			// Check that recommendations are meaningful
			for _, rec := range recommendations {
				assert.NotEmpty(t, rec)
				assert.True(t, len(rec) > 10) // Recommendations should be substantial
			}
		})
	}
}

func TestConfidenceScorer_Integration(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	// Create a realistic company size result
	result := &CompanySizeResult{
		CompanySize:     "sme",
		ConfidenceScore: 0.75,
		EmployeeAnalysis: &EmployeeCountResult{
			EmployeeCount:    150,
			ConfidenceScore:  0.85,
			ExtractionMethod: "direct_mention",
			IsValidated:      true,
			ExtractedAt:      time.Now().Add(-2 * 24 * time.Hour),
			SourceURL:        "https://linkedin.com/company/example",
		},
		RevenueAnalysis: &RevenueResult{
			RevenueAmount:    8000000,
			ConfidenceScore:  0.8,
			ExtractionMethod: "direct_mention",
			IsValidated:      true,
			ExtractedAt:      time.Now().Add(-1 * 24 * time.Hour),
			SourceURL:        "https://company.com",
		},
		ConsistencyScore: 1.0,
		Evidence: []string{
			"Direct employee count: 150",
			"Direct revenue mention: $8M",
			"Financial indicator: profitable",
		},
		IsValidated:      true,
		ValidationStatus: ValidationStatus{IsValid: true},
		ClassifiedAt:     time.Now(),
		SourceURL:        "https://company.com",
		DataQualityScore: 0.85,
	}

	confidenceScore, err := scorer.CalculateConfidence(context.Background(), result)
	require.NoError(t, err)
	assert.NotNil(t, confidenceScore)

	// Verify comprehensive scoring
	assert.True(t, confidenceScore.OverallConfidence > 0.7, "Expected high confidence")
	assert.Equal(t, "high", confidenceScore.ConfidenceLevel)
	assert.True(t, confidenceScore.DataQualityScore > 0.7)
	assert.True(t, confidenceScore.ConsistencyScore > 0.8)
	assert.True(t, confidenceScore.ValidationScore > 0.8)
	assert.True(t, confidenceScore.EvidenceScore > 0.6)
	assert.True(t, confidenceScore.FreshnessScore > 0.8)
	assert.True(t, confidenceScore.SourceReliabilityScore > 0.6)

	// Verify uncertainty quantification
	assert.True(t, confidenceScore.UncertaintyLevel < 0.3)
	assert.InDelta(t, 1.0, confidenceScore.OverallConfidence+confidenceScore.UncertaintyLevel, 0.01)

	// Verify confidence interval
	assert.True(t, confidenceScore.ConfidenceInterval.LowerBound > 0.6)
	assert.True(t, confidenceScore.ConfidenceInterval.UpperBound < 1.0)
	assert.Equal(t, 0.95, confidenceScore.ConfidenceInterval.Level)

	// Verify anomaly detection
	assert.True(t, confidenceScore.AnomalyScore < 0.3, "Expected low anomaly score for normal data")

	// Verify component breakdown
	assert.Len(t, confidenceScore.ComponentBreakdown, 6) // All components should be present
	assert.NotEmpty(t, confidenceScore.Factors)
	// Recommendations may be empty for high confidence results

	// Verify metadata
	assert.False(t, confidenceScore.CalculatedAt.IsZero())
	assert.False(t, confidenceScore.CalibrationApplied) // No calibration by default
}

func TestConfidenceScorer_Performance(t *testing.T) {
	scorer := NewConfidenceScorer(nil, zap.NewNop())

	// Create a complex result for performance testing
	result := &CompanySizeResult{
		CompanySize:     "enterprise",
		ConfidenceScore: 0.9,
		EmployeeAnalysis: &EmployeeCountResult{
			EmployeeCount:    1000,
			ConfidenceScore:  0.95,
			ExtractionMethod: "direct_mention",
			IsValidated:      true,
			ExtractedAt:      time.Now(),
			SourceURL:        "https://linkedin.com/company/enterprise",
		},
		RevenueAnalysis: &RevenueResult{
			RevenueAmount:    100000000,
			ConfidenceScore:  0.9,
			ExtractionMethod: "direct_mention",
			IsValidated:      true,
			ExtractedAt:      time.Now(),
			SourceURL:        "https://crunchbase.com/company/enterprise",
		},
		ConsistencyScore: 1.0,
		Evidence: []string{
			"Direct employee count: 1000",
			"Direct revenue mention: $100M",
			"Financial indicator: profitable",
			"Company size: enterprise",
			"Industry leader",
		},
		IsValidated:      true,
		ValidationStatus: ValidationStatus{IsValid: true},
		ClassifiedAt:     time.Now(),
		SourceURL:        "https://enterprise.com",
		DataQualityScore: 0.95,
	}

	start := time.Now()
	confidenceScore, err := scorer.CalculateConfidence(context.Background(), result)
	duration := time.Since(start)

	require.NoError(t, err)
	assert.NotNil(t, confidenceScore)

	// Should complete within 50ms
	assert.True(t, duration < 50*time.Millisecond, "Confidence calculation took too long: %v", duration)

	// Should still produce high-quality results
	assert.True(t, confidenceScore.OverallConfidence > 0.8)
	assert.Equal(t, "high", confidenceScore.ConfidenceLevel)
	assert.NotEmpty(t, confidenceScore.ComponentBreakdown)
	assert.NotEmpty(t, confidenceScore.Factors)
}
