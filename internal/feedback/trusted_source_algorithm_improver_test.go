package feedback

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestTrustedSourceAlgorithmImprover_NewTrustedSourceAlgorithmImprover(t *testing.T) {
	config := &TrustedSourceConfig{
		ImprovementInterval:        1 * time.Hour,
		MinFeedbackCount:           10,
		ImprovementThreshold:       0.7,
		MaxConsecutiveImprovements: 5,
		EnableAutoImprovement:      true,
		ValidationTimeout:          500 * time.Millisecond,
		ConfidenceThreshold:        0.8,
		TrustScoreWeight:           0.4,
		ReliabilityWeight:          0.3,
		AccuracyWeight:             0.3,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	assert.NotNil(t, improver)
	assert.Equal(t, config, improver.config)
	assert.Equal(t, logger, improver.logger)
	assert.Equal(t, mockRepo, improver.feedbackRepository)
	assert.Equal(t, mockAnalyzer, improver.securityAnalyzer)
	assert.Equal(t, mockDetector, improver.patternDetector)
	assert.NotNil(t, improver.improvementHistory)
	assert.NotNil(t, improver.performanceMetrics)
	assert.Equal(t, 0.7, improver.improvementThreshold)
	assert.Equal(t, 5, improver.maxConsecutiveImprovements)
}

func TestTrustedSourceAlgorithmImprover_StartImprovementProcess(t *testing.T) {
	config := &TrustedSourceConfig{
		ImprovementInterval:        100 * time.Millisecond,
		MinFeedbackCount:           1,
		ImprovementThreshold:       0.5,
		MaxConsecutiveImprovements: 3,
		EnableAutoImprovement:      true,
		ValidationTimeout:          100 * time.Millisecond,
		ConfidenceThreshold:        0.6,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := improver.StartImprovementProcess(ctx)
	assert.NoError(t, err)

	// Wait for the process to run
	time.Sleep(150 * time.Millisecond)
}

func TestTrustedSourceAlgorithmImprover_ShouldImprove(t *testing.T) {
	tests := []struct {
		name                    string
		config                  *TrustedSourceConfig
		consecutiveImprovements int
		lastImprovementTime     time.Time
		performanceMetrics      *TrustedSourceMetrics
		expected                bool
	}{
		{
			name: "auto improvement disabled",
			config: &TrustedSourceConfig{
				EnableAutoImprovement: false,
			},
			expected: false,
		},
		{
			name: "max consecutive improvements reached",
			config: &TrustedSourceConfig{
				EnableAutoImprovement:      true,
				MaxConsecutiveImprovements: 3,
			},
			consecutiveImprovements: 3,
			expected:                false,
		},
		{
			name: "insufficient time since last improvement",
			config: &TrustedSourceConfig{
				EnableAutoImprovement: true,
				ImprovementInterval:   1 * time.Hour,
			},
			lastImprovementTime: time.Now().Add(-30 * time.Minute),
			expected:            false,
		},
		{
			name: "low trust score - should improve",
			config: &TrustedSourceConfig{
				EnableAutoImprovement:      true,
				ImprovementInterval:        1 * time.Hour,
				ValidationTimeout:          500 * time.Millisecond,
				ConfidenceThreshold:        0.8,
				MaxConsecutiveImprovements: 5,
			},
			consecutiveImprovements: 0, // Not at max
			lastImprovementTime:     time.Now().Add(-2 * time.Hour),
			performanceMetrics: &TrustedSourceMetrics{
				AverageTrustScore:        0.6, // Below 0.7 threshold
				AverageReliabilityScore:  0.8,
				AverageAccuracyScore:     0.8,
				FalsePositiveRate:        0.05,
				FalseNegativeRate:        0.05,
				SourceTrustViolationRate: 0.02,
				AverageValidationTime:    200 * time.Millisecond,
			},
			expected: true,
		},
		{
			name: "low reliability score - should improve",
			config: &TrustedSourceConfig{
				EnableAutoImprovement:      true,
				ImprovementInterval:        1 * time.Hour,
				ValidationTimeout:          500 * time.Millisecond,
				ConfidenceThreshold:        0.8,
				MaxConsecutiveImprovements: 5,
			},
			consecutiveImprovements: 0, // Not at max
			lastImprovementTime:     time.Now().Add(-2 * time.Hour),
			performanceMetrics: &TrustedSourceMetrics{
				AverageTrustScore:        0.8,
				AverageReliabilityScore:  0.6, // Below 0.7 threshold
				AverageAccuracyScore:     0.8,
				FalsePositiveRate:        0.05,
				FalseNegativeRate:        0.05,
				SourceTrustViolationRate: 0.02,
				AverageValidationTime:    200 * time.Millisecond,
			},
			expected: true,
		},
		{
			name: "low accuracy score - should improve",
			config: &TrustedSourceConfig{
				EnableAutoImprovement:      true,
				ImprovementInterval:        1 * time.Hour,
				ValidationTimeout:          500 * time.Millisecond,
				ConfidenceThreshold:        0.8,
				MaxConsecutiveImprovements: 5,
			},
			consecutiveImprovements: 0, // Not at max
			lastImprovementTime:     time.Now().Add(-2 * time.Hour),
			performanceMetrics: &TrustedSourceMetrics{
				AverageTrustScore:        0.8,
				AverageReliabilityScore:  0.8,
				AverageAccuracyScore:     0.6, // Below 0.7 threshold
				FalsePositiveRate:        0.05,
				FalseNegativeRate:        0.05,
				SourceTrustViolationRate: 0.02,
				AverageValidationTime:    200 * time.Millisecond,
			},
			expected: true,
		},
		{
			name: "high false positive rate - should improve",
			config: &TrustedSourceConfig{
				EnableAutoImprovement:      true,
				ImprovementInterval:        1 * time.Hour,
				ValidationTimeout:          500 * time.Millisecond,
				ConfidenceThreshold:        0.8,
				MaxConsecutiveImprovements: 5,
			},
			consecutiveImprovements: 0, // Not at max
			lastImprovementTime:     time.Now().Add(-2 * time.Hour),
			performanceMetrics: &TrustedSourceMetrics{
				AverageTrustScore:        0.8,
				AverageReliabilityScore:  0.8,
				AverageAccuracyScore:     0.8,
				FalsePositiveRate:        0.15, // Above 0.1 threshold
				FalseNegativeRate:        0.05,
				SourceTrustViolationRate: 0.02,
				AverageValidationTime:    200 * time.Millisecond,
			},
			expected: true,
		},
		{
			name: "high source trust violation rate - should improve",
			config: &TrustedSourceConfig{
				EnableAutoImprovement:      true,
				ImprovementInterval:        1 * time.Hour,
				ValidationTimeout:          500 * time.Millisecond,
				ConfidenceThreshold:        0.8,
				MaxConsecutiveImprovements: 5,
			},
			consecutiveImprovements: 0, // Not at max
			lastImprovementTime:     time.Now().Add(-2 * time.Hour),
			performanceMetrics: &TrustedSourceMetrics{
				AverageTrustScore:        0.8,
				AverageReliabilityScore:  0.8,
				AverageAccuracyScore:     0.8,
				FalsePositiveRate:        0.05,
				FalseNegativeRate:        0.05,
				SourceTrustViolationRate: 0.08, // Above 0.05 threshold
				AverageValidationTime:    200 * time.Millisecond,
			},
			expected: true,
		},
		{
			name: "no improvement opportunities",
			config: &TrustedSourceConfig{
				EnableAutoImprovement: true,
				ImprovementInterval:   1 * time.Hour,
				ValidationTimeout:     500 * time.Millisecond,
				ConfidenceThreshold:   0.8,
			},
			lastImprovementTime: time.Now().Add(-2 * time.Hour),
			performanceMetrics: &TrustedSourceMetrics{
				AverageTrustScore:        0.8,
				AverageReliabilityScore:  0.8,
				AverageAccuracyScore:     0.8,
				FalsePositiveRate:        0.05,
				FalseNegativeRate:        0.05,
				SourceTrustViolationRate: 0.02,
				AverageValidationTime:    200 * time.Millisecond,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			mockRepo := &MockFeedbackRepository{}
			mockAnalyzer := &SecurityFeedbackAnalyzer{}
			mockDetector := &SecurityPatternDetector{}

			improver := NewTrustedSourceAlgorithmImprover(
				tt.config,
				logger,
				mockRepo,
				mockAnalyzer,
				mockDetector,
			)

			improver.consecutiveImprovements = tt.consecutiveImprovements
			improver.lastImprovementTime = tt.lastImprovementTime
			if tt.performanceMetrics != nil {
				improver.performanceMetrics = tt.performanceMetrics
			}

			result := improver.shouldImprove()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrustedSourceAlgorithmImprover_HasImprovementOpportunities(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *TrustedSourceMetrics
		config   *TrustedSourceConfig
		expected bool
	}{
		{
			name: "low trust score",
			metrics: &TrustedSourceMetrics{
				AverageTrustScore: 0.6, // Below 0.7 threshold
			},
			config: &TrustedSourceConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "low reliability score",
			metrics: &TrustedSourceMetrics{
				AverageReliabilityScore: 0.6, // Below 0.7 threshold
			},
			config: &TrustedSourceConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "low accuracy score",
			metrics: &TrustedSourceMetrics{
				AverageAccuracyScore: 0.6, // Below 0.7 threshold
			},
			config: &TrustedSourceConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "high false positive rate",
			metrics: &TrustedSourceMetrics{
				FalsePositiveRate: 0.15, // Above 0.1 threshold
			},
			config: &TrustedSourceConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "high source trust violation rate",
			metrics: &TrustedSourceMetrics{
				SourceTrustViolationRate: 0.08, // Above 0.05 threshold
			},
			config: &TrustedSourceConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "slow validation time",
			metrics: &TrustedSourceMetrics{
				AverageValidationTime: 600 * time.Millisecond, // Above 500ms threshold
			},
			config: &TrustedSourceConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "no improvement opportunities",
			metrics: &TrustedSourceMetrics{
				AverageTrustScore:        0.8,
				AverageReliabilityScore:  0.8,
				AverageAccuracyScore:     0.8,
				FalsePositiveRate:        0.05,
				FalseNegativeRate:        0.05,
				SourceTrustViolationRate: 0.02,
				AverageValidationTime:    200 * time.Millisecond,
			},
			config: &TrustedSourceConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			mockRepo := &MockFeedbackRepository{}
			mockAnalyzer := &SecurityFeedbackAnalyzer{}
			mockDetector := &SecurityPatternDetector{}

			improver := NewTrustedSourceAlgorithmImprover(
				tt.config,
				logger,
				mockRepo,
				mockAnalyzer,
				mockDetector,
			)

			improver.performanceMetrics = tt.metrics

			result := improver.hasImprovementOpportunities()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTrustedSourceAlgorithmImprover_GeneratePatternBasedImprovement(t *testing.T) {
	config := &TrustedSourceConfig{
		ConfidenceThreshold: 0.7,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	pattern := &SecurityPattern{
		PatternID:          "pattern_123",
		PatternType:        "untrusted_source_pattern",
		Description:        "Frequent untrusted source issues",
		Confidence:         0.85,
		Severity:           "high",
		AffectedComponents: []string{"source_validation", "trust_scoring"},
	}

	improvement := improver.generatePatternBasedImprovement(pattern)

	assert.NotNil(t, improvement)
	assert.Equal(t, "pattern_based", improvement.ImprovementType)
	assert.Contains(t, improvement.ImprovementID, "pattern_improvement_pattern_123")
	assert.Contains(t, improvement.Description, "untrusted_source_pattern")
	assert.Equal(t, 0.85, improvement.Confidence)
	assert.Equal(t, "pattern_123", improvement.Changes["pattern_id"])
	assert.Equal(t, "untrusted_source_pattern", improvement.Changes["pattern_type"])
	assert.Equal(t, "high", improvement.Changes["severity"])
	assert.Equal(t, "enhance_source_validation", improvement.Changes["action"])
}

func TestTrustedSourceAlgorithmImprover_GeneratePerformanceBasedImprovements(t *testing.T) {
	config := &TrustedSourceConfig{
		ValidationTimeout:   500 * time.Millisecond,
		ConfidenceThreshold: 0.8,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Test with low trust score
	improver.performanceMetrics = &TrustedSourceMetrics{
		AverageTrustScore:        0.6, // Below 0.7 threshold
		AverageReliabilityScore:  0.8,
		AverageAccuracyScore:     0.8,
		FalsePositiveRate:        0.05,
		FalseNegativeRate:        0.05,
		SourceTrustViolationRate: 0.02,
		AverageValidationTime:    200 * time.Millisecond,
	}

	improvements := improver.generatePerformanceBasedImprovements()

	assert.Len(t, improvements, 1)
	assert.Equal(t, "performance_based", improvements[0].ImprovementType)
	assert.Contains(t, improvements[0].Description, "trust score")
	assert.Equal(t, "average_trust_score", improvements[0].Changes["metric"])
	assert.Equal(t, 0.6, improvements[0].Changes["current_value"])
	assert.Equal(t, 0.8, improvements[0].Changes["target_value"])
}

func TestTrustedSourceAlgorithmImprover_GenerateSourceBasedImprovements(t *testing.T) {
	config := &TrustedSourceConfig{
		ConfidenceThreshold: 0.8,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Set up source metrics with low trust score
	improver.performanceMetrics = &TrustedSourceMetrics{
		SourceMetrics: map[string]*SourceMetrics{
			"source1": {
				SourceName:        "source1",
				TrustScore:        0.5, // Below 0.6 threshold
				ReliabilityScore:  0.8,
				AccuracyScore:     0.8,
				SuccessRate:       0.9,
				FalsePositiveRate: 0.2, // Above 0.15 threshold
			},
			"source2": {
				SourceName:        "source2",
				TrustScore:        0.8, // Above threshold
				ReliabilityScore:  0.8,
				AccuracyScore:     0.8,
				SuccessRate:       0.9,
				FalsePositiveRate: 0.1, // Below threshold
			},
		},
	}

	improvements := improver.generateSourceBasedImprovements()

	// Should generate improvements for source1 (low trust score and high FP rate)
	assert.Len(t, improvements, 2) // One for trust score, one for false positive rate

	// Check trust score improvement
	trustImprovement := improvements[0]
	assert.Equal(t, "source_based", trustImprovement.ImprovementType)
	assert.Contains(t, trustImprovement.Description, "source1")
	assert.Contains(t, trustImprovement.Description, "trust score")
	assert.Equal(t, "source1", trustImprovement.Changes["source_name"])

	// Check false positive rate improvement
	fpImprovement := improvements[1]
	assert.Equal(t, "source_based", fpImprovement.ImprovementType)
	assert.Contains(t, fpImprovement.Description, "source1")
	assert.Contains(t, fpImprovement.Description, "false positive rate")
	assert.Equal(t, "source1", fpImprovement.Changes["source_name"])
}

func TestTrustedSourceAlgorithmImprover_ValidateImprovement(t *testing.T) {
	tests := []struct {
		name          string
		improvement   *TrustedSourceImprovement
		config        *TrustedSourceConfig
		expectedError bool
		errorContains string
	}{
		{
			name: "valid improvement",
			improvement: &TrustedSourceImprovement{
				Confidence: 0.9, // Above threshold
			},
			config: &TrustedSourceConfig{
				ConfidenceThreshold: 0.8,
			},
			expectedError: false,
		},
		{
			name: "low confidence",
			improvement: &TrustedSourceImprovement{
				Confidence: 0.6, // Below threshold
			},
			config: &TrustedSourceConfig{
				ConfidenceThreshold: 0.8,
			},
			expectedError: true,
			errorContains: "confidence too low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			mockRepo := &MockFeedbackRepository{}
			mockAnalyzer := &SecurityFeedbackAnalyzer{}
			mockDetector := &SecurityPatternDetector{}

			improver := NewTrustedSourceAlgorithmImprover(
				tt.config,
				logger,
				mockRepo,
				mockAnalyzer,
				mockDetector,
			)

			err := improver.validateImprovement(context.Background(), tt.improvement)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTrustedSourceAlgorithmImprover_RecordImprovement(t *testing.T) {
	config := &TrustedSourceConfig{
		MaxConsecutiveImprovements: 3,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	improvement := &TrustedSourceImprovement{
		ImprovementID: "test_improvement_1",
		AppliedAt:     time.Now(),
	}

	// Record first improvement
	improver.recordImprovement(improvement)

	assert.Len(t, improver.improvementHistory, 1)
	assert.Equal(t, 1, improver.consecutiveImprovements)
	assert.False(t, improver.lastImprovementTime.IsZero())

	// Record second improvement
	improvement2 := &TrustedSourceImprovement{
		ImprovementID: "test_improvement_2",
		AppliedAt:     time.Now(),
	}

	improver.recordImprovement(improvement2)

	assert.Len(t, improver.improvementHistory, 2)
	assert.Equal(t, 2, improver.consecutiveImprovements)
}

func TestTrustedSourceAlgorithmImprover_GetImprovementHistory(t *testing.T) {
	config := &TrustedSourceConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Add some improvements
	improvement1 := &TrustedSourceImprovement{ImprovementID: "improvement_1"}
	improvement2 := &TrustedSourceImprovement{ImprovementID: "improvement_2"}

	improver.improvementHistory = []*TrustedSourceImprovement{improvement1, improvement2}

	history := improver.GetImprovementHistory()

	assert.Len(t, history, 2)
	assert.Equal(t, "improvement_1", history[0].ImprovementID)
	assert.Equal(t, "improvement_2", history[1].ImprovementID)

	// Ensure it's a copy (modifying the returned slice shouldn't affect the original)
	history[0] = &TrustedSourceImprovement{ImprovementID: "modified"}
	assert.Equal(t, "improvement_1", improver.improvementHistory[0].ImprovementID)
}

func TestTrustedSourceAlgorithmImprover_GetPerformanceMetrics(t *testing.T) {
	config := &TrustedSourceConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Set up some metrics
	improver.performanceMetrics = &TrustedSourceMetrics{
		TotalValidations:        100,
		SuccessfulValidations:   95,
		AverageTrustScore:       0.9,
		AverageReliabilityScore: 0.85,
		AverageAccuracyScore:    0.88,
	}

	metrics := improver.GetPerformanceMetrics()

	assert.Equal(t, int64(100), metrics.TotalValidations)
	assert.Equal(t, int64(95), metrics.SuccessfulValidations)
	assert.Equal(t, 0.9, metrics.AverageTrustScore)
	assert.Equal(t, 0.85, metrics.AverageReliabilityScore)
	assert.Equal(t, 0.88, metrics.AverageAccuracyScore)

	// Ensure it's a copy (modifying the returned metrics shouldn't affect the original)
	metrics.TotalValidations = 200
	assert.Equal(t, int64(100), improver.performanceMetrics.TotalValidations)
}

func TestTrustedSourceAlgorithmImprover_ValidateImprovementAfterApplied(t *testing.T) {
	config := &TrustedSourceConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Add an improvement to history
	improvement := &TrustedSourceImprovement{
		ImprovementID: "test_improvement",
		AppliedAt:     time.Now(),
	}
	improver.improvementHistory = []*TrustedSourceImprovement{improvement}

	validation, err := improver.ValidateImprovement(context.Background(), "test_improvement")

	require.NoError(t, err)
	assert.NotNil(t, validation)
	assert.Equal(t, "comprehensive_testing", validation.ValidationMethod)
	assert.Equal(t, 100, validation.TestCases)
	assert.Equal(t, 95, validation.PassedCases)
	assert.Equal(t, 0.95, validation.TrustScoreAccuracy)
	assert.Equal(t, 0.92, validation.ReliabilityAccuracy)
	assert.Equal(t, 0.94, validation.OverallAccuracy)
	assert.Equal(t, 0.9, validation.Performance)
	assert.Equal(t, 0.98, validation.Stability)
	assert.NotNil(t, improvement.ValidatedAt)
	assert.NotNil(t, improvement.ValidationResults)
}

func TestTrustedSourceAlgorithmImprover_ValidateImprovementNotFound(t *testing.T) {
	config := &TrustedSourceConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	validation, err := improver.ValidateImprovement(context.Background(), "nonexistent_improvement")

	assert.Error(t, err)
	assert.Nil(t, validation)
	assert.Contains(t, err.Error(), "improvement not found")
}

func TestTrustedSourceAlgorithmImprover_RollbackImprovement(t *testing.T) {
	config := &TrustedSourceConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Add an improvement to history
	improvement := &TrustedSourceImprovement{
		ImprovementID: "test_improvement",
		AppliedAt:     time.Now(),
	}
	improver.improvementHistory = []*TrustedSourceImprovement{improvement}

	err := improver.RollbackImprovement(context.Background(), "test_improvement", "performance degradation")

	require.NoError(t, err)
	assert.True(t, improvement.RollbackRequired)
	assert.Equal(t, "performance degradation", improvement.RollbackReason)
}

func TestTrustedSourceAlgorithmImprover_RollbackImprovementNotFound(t *testing.T) {
	config := &TrustedSourceConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	err := improver.RollbackImprovement(context.Background(), "nonexistent_improvement", "test reason")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "improvement not found")
}

func TestTrustedSourceAlgorithmImprover_GetImprovementStatus(t *testing.T) {
	config := &TrustedSourceConfig{
		EnableAutoImprovement:      true,
		MaxConsecutiveImprovements: 5,
		ImprovementThreshold:       0.7,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Set up some state
	improver.consecutiveImprovements = 2
	improver.lastImprovementTime = time.Now().Add(-1 * time.Hour)
	improver.improvementHistory = []*TrustedSourceImprovement{
		{ImprovementID: "improvement_1"},
		{ImprovementID: "improvement_2"},
	}

	status := improver.GetImprovementStatus()

	assert.NotNil(t, status)
	assert.Equal(t, true, status["is_running"])
	assert.Equal(t, 2, status["consecutive_improvements"])
	assert.Equal(t, 5, status["max_consecutive_improvements"])
	assert.Equal(t, 2, status["total_improvements"])
	assert.Equal(t, 0.7, status["improvement_threshold"])
	assert.NotNil(t, status["performance_metrics"])
	assert.NotNil(t, status["last_improvement_time"])
}

func TestTrustedSourceAlgorithmImprover_ConcurrentOperations(t *testing.T) {
	config := &TrustedSourceConfig{
		EnableAutoImprovement: true,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Test concurrent access to improvement history
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			improvement := &TrustedSourceImprovement{
				ImprovementID: fmt.Sprintf("improvement_%d", id),
				AppliedAt:     time.Now(),
			}
			improver.recordImprovement(improvement)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all improvements were recorded
	history := improver.GetImprovementHistory()
	assert.Len(t, history, 10)
	assert.Equal(t, 10, improver.consecutiveImprovements)
}

func TestTrustedSourceAlgorithmImprover_EdgeCases(t *testing.T) {
	config := &TrustedSourceConfig{
		EnableAutoImprovement: true,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewTrustedSourceAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Test with nil performance metrics
	improver.performanceMetrics = nil
	hasOpportunities := improver.hasImprovementOpportunities()
	assert.False(t, hasOpportunities)

	// Test with empty improvement history
	history := improver.GetImprovementHistory()
	assert.NotNil(t, history)
	assert.Len(t, history, 0)

	// Test with nil performance metrics in GetPerformanceMetrics
	improver.performanceMetrics = nil
	metrics := improver.GetPerformanceMetrics()
	assert.NotNil(t, metrics)
}
