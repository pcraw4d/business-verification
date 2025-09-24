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

func TestSecurityValidationAlgorithmImprover_NewSecurityValidationAlgorithmImprover(t *testing.T) {
	config := &SecurityValidationConfig{
		ImprovementInterval:        1 * time.Hour,
		MinFeedbackCount:           10,
		ImprovementThreshold:       0.7,
		MaxConsecutiveImprovements: 5,
		EnableAutoImprovement:      true,
		ValidationTimeout:          500 * time.Millisecond,
		ConfidenceThreshold:        0.8,
		PatternWeight:              0.3,
		PerformanceWeight:          0.4,
		FeedbackWeight:             0.3,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
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

func TestSecurityValidationAlgorithmImprover_StartImprovementProcess(t *testing.T) {
	config := &SecurityValidationConfig{
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

	improver := NewSecurityValidationAlgorithmImprover(
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

func TestSecurityValidationAlgorithmImprover_ShouldImprove(t *testing.T) {
	tests := []struct {
		name                    string
		config                  *SecurityValidationConfig
		consecutiveImprovements int
		lastImprovementTime     time.Time
		performanceMetrics      *SecurityValidationMetrics
		expected                bool
	}{
		{
			name: "auto improvement disabled",
			config: &SecurityValidationConfig{
				EnableAutoImprovement: false,
			},
			expected: false,
		},
		{
			name: "max consecutive improvements reached",
			config: &SecurityValidationConfig{
				EnableAutoImprovement:      true,
				MaxConsecutiveImprovements: 3,
			},
			consecutiveImprovements: 3,
			expected:                false,
		},
		{
			name: "insufficient time since last improvement",
			config: &SecurityValidationConfig{
				EnableAutoImprovement: true,
				ImprovementInterval:   1 * time.Hour,
			},
			lastImprovementTime: time.Now().Add(-30 * time.Minute),
			expected:            false,
		},
		{
			name: "high false positive rate - should improve",
			config: &SecurityValidationConfig{
				EnableAutoImprovement:      true,
				ImprovementInterval:        1 * time.Hour,
				ValidationTimeout:          500 * time.Millisecond,
				ConfidenceThreshold:        0.8,
				MaxConsecutiveImprovements: 5,
			},
			consecutiveImprovements: 0, // Not at max
			lastImprovementTime:     time.Now().Add(-2 * time.Hour),
			performanceMetrics: &SecurityValidationMetrics{
				FalsePositiveRate:     0.15, // Above threshold
				FalseNegativeRate:     0.05,
				AverageConfidence:     0.9,
				SecurityViolationRate: 0.02,
				AverageValidationTime: 200 * time.Millisecond,
			},
			expected: true,
		},
		{
			name: "low confidence score - should improve",
			config: &SecurityValidationConfig{
				EnableAutoImprovement:      true,
				ImprovementInterval:        1 * time.Hour,
				ValidationTimeout:          500 * time.Millisecond,
				ConfidenceThreshold:        0.8,
				MaxConsecutiveImprovements: 5,
			},
			consecutiveImprovements: 0, // Not at max
			lastImprovementTime:     time.Now().Add(-2 * time.Hour),
			performanceMetrics: &SecurityValidationMetrics{
				FalsePositiveRate:     0.05,
				FalseNegativeRate:     0.05,
				AverageConfidence:     0.6, // Below threshold
				SecurityViolationRate: 0.02,
				AverageValidationTime: 200 * time.Millisecond,
			},
			expected: true,
		},
		{
			name: "no improvement opportunities",
			config: &SecurityValidationConfig{
				EnableAutoImprovement: true,
				ImprovementInterval:   1 * time.Hour,
				ValidationTimeout:     500 * time.Millisecond,
				ConfidenceThreshold:   0.8,
			},
			lastImprovementTime: time.Now().Add(-2 * time.Hour),
			performanceMetrics: &SecurityValidationMetrics{
				FalsePositiveRate:     0.05,
				FalseNegativeRate:     0.05,
				AverageConfidence:     0.9,
				SecurityViolationRate: 0.02,
				AverageValidationTime: 200 * time.Millisecond,
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

			improver := NewSecurityValidationAlgorithmImprover(
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

func TestSecurityValidationAlgorithmImprover_HasImprovementOpportunities(t *testing.T) {
	tests := []struct {
		name     string
		metrics  *SecurityValidationMetrics
		config   *SecurityValidationConfig
		expected bool
	}{
		{
			name: "high false positive rate",
			metrics: &SecurityValidationMetrics{
				FalsePositiveRate: 0.15, // Above 0.1 threshold
			},
			config: &SecurityValidationConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "high false negative rate",
			metrics: &SecurityValidationMetrics{
				FalseNegativeRate: 0.15, // Above 0.1 threshold
			},
			config: &SecurityValidationConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "low confidence score",
			metrics: &SecurityValidationMetrics{
				AverageConfidence: 0.6, // Below 0.8 threshold
			},
			config: &SecurityValidationConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "high security violation rate",
			metrics: &SecurityValidationMetrics{
				SecurityViolationRate: 0.08, // Above 0.05 threshold
			},
			config: &SecurityValidationConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "slow validation time",
			metrics: &SecurityValidationMetrics{
				AverageValidationTime: 600 * time.Millisecond, // Above 500ms threshold
			},
			config: &SecurityValidationConfig{
				ValidationTimeout:   500 * time.Millisecond,
				ConfidenceThreshold: 0.8,
			},
			expected: true,
		},
		{
			name: "no improvement opportunities",
			metrics: &SecurityValidationMetrics{
				FalsePositiveRate:     0.05,
				FalseNegativeRate:     0.05,
				AverageConfidence:     0.9,
				SecurityViolationRate: 0.02,
				AverageValidationTime: 200 * time.Millisecond,
			},
			config: &SecurityValidationConfig{
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

			improver := NewSecurityValidationAlgorithmImprover(
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

func TestSecurityValidationAlgorithmImprover_GeneratePatternBasedImprovement(t *testing.T) {
	config := &SecurityValidationConfig{
		ConfidenceThreshold: 0.7,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	pattern := &SecurityPattern{
		PatternID:          "pattern_123",
		PatternType:        "recurring_security_violation",
		Description:        "Frequent security validation failures",
		Confidence:         0.85,
		Severity:           "high",
		AffectedComponents: []string{"security_validation", "trust_scoring"},
	}

	improvement := improver.generatePatternBasedImprovement(pattern)

	assert.NotNil(t, improvement)
	assert.Equal(t, "pattern_based", improvement.ImprovementType)
	assert.Contains(t, improvement.ImprovementID, "pattern_improvement_pattern_123")
	assert.Contains(t, improvement.Description, "recurring_security_violation")
	assert.Equal(t, 0.85, improvement.Confidence)
	assert.Equal(t, "pattern_123", improvement.Changes["pattern_id"])
	assert.Equal(t, "recurring_security_violation", improvement.Changes["pattern_type"])
	assert.Equal(t, "high", improvement.Changes["severity"])
	assert.Equal(t, "enhance_validation_rules", improvement.Changes["action"])
}

func TestSecurityValidationAlgorithmImprover_GeneratePerformanceBasedImprovements(t *testing.T) {
	config := &SecurityValidationConfig{
		ValidationTimeout:   500 * time.Millisecond,
		ConfidenceThreshold: 0.8,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Test with high false positive rate
	improver.performanceMetrics = &SecurityValidationMetrics{
		FalsePositiveRate:     0.15, // Above threshold
		FalseNegativeRate:     0.05,
		AverageConfidence:     0.9,
		SecurityViolationRate: 0.02,
		AverageValidationTime: 200 * time.Millisecond,
	}

	improvements := improver.generatePerformanceBasedImprovements()

	assert.Len(t, improvements, 1)
	assert.Equal(t, "performance_based", improvements[0].ImprovementType)
	assert.Contains(t, improvements[0].Description, "false positive rate")
	assert.Equal(t, "false_positive_rate", improvements[0].Changes["metric"])
	assert.Equal(t, 0.15, improvements[0].Changes["current_value"])
	assert.Equal(t, 0.05, improvements[0].Changes["target_value"])
}

func TestSecurityValidationAlgorithmImprover_GenerateMethodBasedImprovements(t *testing.T) {
	config := &SecurityValidationConfig{
		ConfidenceThreshold: 0.8,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Set up method metrics with low success rate
	improver.performanceMetrics = &SecurityValidationMetrics{
		ValidationMethodMetrics: map[string]*ValidationMethodMetrics{
			"method1": {
				MethodName:        "method1",
				SuccessRate:       0.6, // Below 0.8 threshold
				FalsePositiveRate: 0.2, // Above 0.15 threshold
			},
			"method2": {
				MethodName:        "method2",
				SuccessRate:       0.9, // Above threshold
				FalsePositiveRate: 0.1, // Below threshold
			},
		},
	}

	improvements := improver.generateMethodBasedImprovements()

	// Should generate improvements for method1 (low success rate and high FP rate)
	assert.Len(t, improvements, 2) // One for success rate, one for false positive rate

	// Check success rate improvement
	successImprovement := improvements[0]
	assert.Equal(t, "method_based", successImprovement.ImprovementType)
	assert.Contains(t, successImprovement.Description, "method1")
	assert.Contains(t, successImprovement.Description, "success rate")
	assert.Equal(t, "method1", successImprovement.Changes["method_name"])

	// Check false positive rate improvement
	fpImprovement := improvements[1]
	assert.Equal(t, "method_based", fpImprovement.ImprovementType)
	assert.Contains(t, fpImprovement.Description, "method1")
	assert.Contains(t, fpImprovement.Description, "false positive rate")
	assert.Equal(t, "method1", fpImprovement.Changes["method_name"])
}

func TestSecurityValidationAlgorithmImprover_ValidateImprovement(t *testing.T) {
	tests := []struct {
		name          string
		improvement   *AlgorithmImprovement
		config        *SecurityValidationConfig
		expectedError bool
		errorContains string
	}{
		{
			name: "valid improvement",
			improvement: &AlgorithmImprovement{
				Confidence: 0.9, // Above threshold
			},
			config: &SecurityValidationConfig{
				ConfidenceThreshold: 0.8,
			},
			expectedError: false,
		},
		{
			name: "low confidence",
			improvement: &AlgorithmImprovement{
				Confidence: 0.6, // Below threshold
			},
			config: &SecurityValidationConfig{
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

			improver := NewSecurityValidationAlgorithmImprover(
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

func TestSecurityValidationAlgorithmImprover_RecordImprovement(t *testing.T) {
	config := &SecurityValidationConfig{
		MaxConsecutiveImprovements: 3,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	improvement := &AlgorithmImprovement{
		ImprovementID: "test_improvement_1",
		AppliedAt:     time.Now(),
	}

	// Record first improvement
	improver.recordImprovement(improvement)

	assert.Len(t, improver.improvementHistory, 1)
	assert.Equal(t, 1, improver.consecutiveImprovements)
	assert.False(t, improver.lastImprovementTime.IsZero())

	// Record second improvement
	improvement2 := &AlgorithmImprovement{
		ImprovementID: "test_improvement_2",
		AppliedAt:     time.Now(),
	}

	improver.recordImprovement(improvement2)

	assert.Len(t, improver.improvementHistory, 2)
	assert.Equal(t, 2, improver.consecutiveImprovements)
}

func TestSecurityValidationAlgorithmImprover_GetImprovementHistory(t *testing.T) {
	config := &SecurityValidationConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Add some improvements
	improvement1 := &AlgorithmImprovement{ImprovementID: "improvement_1"}
	improvement2 := &AlgorithmImprovement{ImprovementID: "improvement_2"}

	improver.improvementHistory = []*AlgorithmImprovement{improvement1, improvement2}

	history := improver.GetImprovementHistory()

	assert.Len(t, history, 2)
	assert.Equal(t, "improvement_1", history[0].ImprovementID)
	assert.Equal(t, "improvement_2", history[1].ImprovementID)

	// Ensure it's a copy (modifying the returned slice shouldn't affect the original)
	history[0] = &AlgorithmImprovement{ImprovementID: "modified"}
	assert.Equal(t, "improvement_1", improver.improvementHistory[0].ImprovementID)
}

func TestSecurityValidationAlgorithmImprover_GetPerformanceMetrics(t *testing.T) {
	config := &SecurityValidationConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Set up some metrics
	improver.performanceMetrics = &SecurityValidationMetrics{
		TotalValidations:      100,
		SuccessfulValidations: 95,
		AverageConfidence:     0.9,
	}

	metrics := improver.GetPerformanceMetrics()

	assert.Equal(t, int64(100), metrics.TotalValidations)
	assert.Equal(t, int64(95), metrics.SuccessfulValidations)
	assert.Equal(t, 0.9, metrics.AverageConfidence)

	// Ensure it's a copy (modifying the returned metrics shouldn't affect the original)
	metrics.TotalValidations = 200
	assert.Equal(t, int64(100), improver.performanceMetrics.TotalValidations)
}

func TestSecurityValidationAlgorithmImprover_ValidateImprovementAfterApplied(t *testing.T) {
	config := &SecurityValidationConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Add an improvement to history
	improvement := &AlgorithmImprovement{
		ImprovementID: "test_improvement",
		AppliedAt:     time.Now(),
	}
	improver.improvementHistory = []*AlgorithmImprovement{improvement}

	validation, err := improver.ValidateImprovement(context.Background(), "test_improvement")

	require.NoError(t, err)
	assert.NotNil(t, validation)
	assert.Equal(t, "comprehensive_testing", validation.ValidationMethod)
	assert.Equal(t, 100, validation.TestCases)
	assert.Equal(t, 95, validation.PassedCases)
	assert.Equal(t, 0.95, validation.Accuracy)
	assert.Equal(t, 0.9, validation.Performance)
	assert.Equal(t, 0.98, validation.Stability)
	assert.NotNil(t, improvement.ValidatedAt)
	assert.NotNil(t, improvement.ValidationResults)
}

func TestSecurityValidationAlgorithmImprover_ValidateImprovementNotFound(t *testing.T) {
	config := &SecurityValidationConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
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

func TestSecurityValidationAlgorithmImprover_RollbackImprovement(t *testing.T) {
	config := &SecurityValidationConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Add an improvement to history
	improvement := &AlgorithmImprovement{
		ImprovementID: "test_improvement",
		AppliedAt:     time.Now(),
	}
	improver.improvementHistory = []*AlgorithmImprovement{improvement}

	err := improver.RollbackImprovement(context.Background(), "test_improvement", "performance degradation")

	require.NoError(t, err)
	assert.True(t, improvement.RollbackRequired)
	assert.Equal(t, "performance degradation", improvement.RollbackReason)
}

func TestSecurityValidationAlgorithmImprover_RollbackImprovementNotFound(t *testing.T) {
	config := &SecurityValidationConfig{}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
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

func TestSecurityValidationAlgorithmImprover_GetImprovementStatus(t *testing.T) {
	config := &SecurityValidationConfig{
		EnableAutoImprovement:      true,
		MaxConsecutiveImprovements: 5,
		ImprovementThreshold:       0.7,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
		config,
		logger,
		mockRepo,
		mockAnalyzer,
		mockDetector,
	)

	// Set up some state
	improver.consecutiveImprovements = 2
	improver.lastImprovementTime = time.Now().Add(-1 * time.Hour)
	improver.improvementHistory = []*AlgorithmImprovement{
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

func TestSecurityValidationAlgorithmImprover_ConcurrentOperations(t *testing.T) {
	config := &SecurityValidationConfig{
		EnableAutoImprovement: true,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
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
			improvement := &AlgorithmImprovement{
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

func TestSecurityValidationAlgorithmImprover_EdgeCases(t *testing.T) {
	config := &SecurityValidationConfig{
		EnableAutoImprovement: true,
	}

	logger := zap.NewNop()
	mockRepo := &MockFeedbackRepository{}
	mockAnalyzer := &SecurityFeedbackAnalyzer{}
	mockDetector := &SecurityPatternDetector{}

	improver := NewSecurityValidationAlgorithmImprover(
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
