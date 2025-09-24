package external

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewContinuousImprovementManager(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	monitor := NewVerificationSuccessMonitor(nil, logger)
	manager := NewContinuousImprovementManager(nil, monitor, logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.config)
	assert.Equal(t, true, manager.config.EnableAutoImprovement)
	assert.Equal(t, 1*time.Hour, manager.config.ImprovementInterval)
	assert.Equal(t, 0.7, manager.config.ConfidenceThreshold)

	// Test with custom config
	customConfig := &ContinuousImprovementConfig{
		EnableAutoImprovement:    false,
		ImprovementInterval:      30 * time.Minute,
		ConfidenceThreshold:      0.8,
		MinDataPointsForAnalysis: 50,
	}

	manager2 := NewContinuousImprovementManager(customConfig, monitor, logger)
	assert.NotNil(t, manager2)
	assert.Equal(t, false, manager2.config.EnableAutoImprovement)
	assert.Equal(t, 30*time.Minute, manager2.config.ImprovementInterval)
	assert.Equal(t, 0.8, manager2.config.ConfidenceThreshold)
	assert.Equal(t, 50, manager2.config.MinDataPointsForAnalysis)
}

func TestDefaultContinuousImprovementConfig(t *testing.T) {
	config := DefaultContinuousImprovementConfig()

	assert.True(t, config.EnableAutoImprovement)
	assert.True(t, config.EnableStrategyOptimization)
	assert.True(t, config.EnableThresholdAdjustment)
	assert.True(t, config.EnableRetryOptimization)
	assert.Equal(t, 1*time.Hour, config.ImprovementInterval)
	assert.Equal(t, 100, config.MinDataPointsForAnalysis)
	assert.Equal(t, 1000, config.MaxImprovementHistory)
	assert.Equal(t, 0.7, config.ConfidenceThreshold)
	assert.Equal(t, -0.05, config.RollbackThreshold)
}

func TestAnalyzeAndRecommend(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)
	manager := NewContinuousImprovementManager(nil, monitor, logger)

	// Test with insufficient data points
	recommendations, err := manager.AnalyzeAndRecommend(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data points")
	assert.Nil(t, recommendations)

	// Add sufficient data points
	for i := 0; i < 150; i++ {
		dataPoint := DataPoint{
			URL:          fmt.Sprintf("https://example%d.com", i),
			Success:      i%3 != 0, // 2/3 success rate
			ResponseTime: 2 * time.Second,
			ErrorType:    "timeout",
			StrategyUsed: "user_agent_rotation",
		}
		monitor.RecordAttempt(context.Background(), dataPoint)
	}

	// Now test with sufficient data
	recommendations, err = manager.AnalyzeAndRecommend(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, recommendations)

	// Should have recommendations for timeout errors and strategy optimization
	assert.Greater(t, len(recommendations), 0)

	// Check that recommendations are sorted by impact * confidence
	for i := 0; i < len(recommendations)-1; i++ {
		scoreI := recommendations[i].Impact * recommendations[i].Confidence
		scoreJ := recommendations[i+1].Impact * recommendations[i+1].Confidence
		assert.GreaterOrEqual(t, scoreI, scoreJ)
	}
}

func TestApplyImprovement(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)
	manager := NewContinuousImprovementManager(nil, monitor, logger)

	recommendation := &ImprovementRecommendation{
		ID:          "test_rec_1",
		Type:        "strategy",
		Priority:    "high",
		Description: "Optimize user agent rotation strategy",
		Impact:      0.05,
		Confidence:  0.8,
		Parameters: map[string]interface{}{
			"strategy_name": "user_agent_rotation",
			"action":        "optimize",
		},
		Reasoning: "High failure rate in user agent rotation strategy",
		CreatedAt: time.Now(),
	}

	strategy, err := manager.ApplyImprovement(context.Background(), recommendation)
	assert.NoError(t, err)
	assert.NotNil(t, strategy)

	assert.Equal(t, "active", strategy.Status)
	assert.Equal(t, recommendation.Type, strategy.Type)
	assert.Equal(t, recommendation.Impact, strategy.Impact)
	assert.Equal(t, recommendation.Confidence, strategy.Confidence)
	assert.NotNil(t, strategy.ActivatedAt)

	// Test invalid improvement type
	invalidRecommendation := &ImprovementRecommendation{
		ID:   "test_rec_2",
		Type: "invalid_type",
	}

	strategy, err = manager.ApplyImprovement(context.Background(), invalidRecommendation)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown improvement type")
	assert.Nil(t, strategy)
}

func TestEvaluateStrategy(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)
	manager := NewContinuousImprovementManager(nil, monitor, logger)

	// Test with non-existent strategy
	evaluation, err := manager.EvaluateStrategy(context.Background(), "non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "strategy not found")
	assert.Nil(t, evaluation)

	// Create and apply a strategy
	recommendation := &ImprovementRecommendation{
		ID:          "test_rec_3",
		Type:        "threshold",
		Description: "Adjust verification thresholds",
		Impact:      0.02,
		Confidence:  0.6,
		Parameters:  map[string]interface{}{},
		CreatedAt:   time.Now(),
	}

	strategy, err := manager.ApplyImprovement(context.Background(), recommendation)
	assert.NoError(t, err)

	// Add some data to see improvement
	for i := 0; i < 50; i++ {
		dataPoint := DataPoint{
			URL:     fmt.Sprintf("https://example%d.com", i),
			Success: true, // High success rate to show improvement
		}
		monitor.RecordAttempt(context.Background(), dataPoint)
	}

	// Evaluate the strategy
	evaluation, err = manager.EvaluateStrategy(context.Background(), strategy.ID)
	assert.NoError(t, err)
	assert.NotNil(t, evaluation)

	assert.Equal(t, strategy.ID, evaluation.StrategyID)
	assert.GreaterOrEqual(t, evaluation.SuccessRateAfter, evaluation.SuccessRateBefore)
	assert.True(t, evaluation.IsBeneficial)
}

func TestRollbackStrategy(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)
	manager := NewContinuousImprovementManager(nil, monitor, logger)

	// Test with non-existent strategy
	err := manager.RollbackStrategy(context.Background(), "non_existent", "test reason")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "strategy not found")

	// Create and apply a strategy
	recommendation := &ImprovementRecommendation{
		ID:          "test_rec_4",
		Type:        "retry",
		Description: "Optimize retry strategy",
		Impact:      0.04,
		Confidence:  0.75,
		Parameters:  map[string]interface{}{},
		CreatedAt:   time.Now(),
	}

	strategy, err := manager.ApplyImprovement(context.Background(), recommendation)
	assert.NoError(t, err)

	// Rollback the strategy
	err = manager.RollbackStrategy(context.Background(), strategy.ID, "Poor performance")
	assert.NoError(t, err)

	// Check that strategy is rolled back
	activeStrategies := manager.GetActiveStrategies()
	found := false
	for _, s := range activeStrategies {
		if s.ID == strategy.ID {
			found = true
			break
		}
	}
	assert.False(t, found, "Strategy should not be in active strategies after rollback")
}

func TestGetActiveStrategies(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)
	manager := NewContinuousImprovementManager(nil, monitor, logger)

	// Initially should be empty
	activeStrategies := manager.GetActiveStrategies()
	assert.Empty(t, activeStrategies)

	// Apply a strategy
	recommendation := &ImprovementRecommendation{
		ID:          "test_rec_5",
		Type:        "strategy",
		Description: "Test strategy",
		Impact:      0.03,
		Confidence:  0.7,
		Parameters:  map[string]interface{}{},
		CreatedAt:   time.Now(),
	}

	strategy, err := manager.ApplyImprovement(context.Background(), recommendation)
	assert.NoError(t, err)

	// Should now have one active strategy
	activeStrategies = manager.GetActiveStrategies()
	assert.Len(t, activeStrategies, 1)
	assert.Equal(t, strategy.ID, activeStrategies[0].ID)
}

func TestUpdateContinuousImprovementConfig(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)
	manager := NewContinuousImprovementManager(nil, monitor, logger)

	// Test valid config update
	newConfig := &ContinuousImprovementConfig{
		EnableAutoImprovement:    false,
		ConfidenceThreshold:      0.8,
		ImprovementInterval:      2 * time.Hour,
		MinDataPointsForAnalysis: 200,
	}

	err := manager.UpdateConfig(newConfig)
	assert.NoError(t, err)

	updatedConfig := manager.GetConfig()
	assert.Equal(t, false, updatedConfig.EnableAutoImprovement)
	assert.Equal(t, 0.8, updatedConfig.ConfidenceThreshold)
	assert.Equal(t, 2*time.Hour, updatedConfig.ImprovementInterval)
	assert.Equal(t, 200, updatedConfig.MinDataPointsForAnalysis)

	// Test invalid config (nil)
	err = manager.UpdateConfig(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")

	// Test invalid config (confidence threshold out of range)
	invalidConfig := &ContinuousImprovementConfig{
		ConfidenceThreshold: 1.5, // Invalid: > 1
	}

	err = manager.UpdateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "confidence threshold must be between 0 and 1")

	// Test invalid config (improvement interval too short)
	invalidConfig2 := &ContinuousImprovementConfig{
		ImprovementInterval: 30 * time.Second, // Invalid: < 1 minute
	}

	err = manager.UpdateConfig(invalidConfig2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "improvement interval must be at least 1 minute")
}

func TestGenerateStrategyRecommendations(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)
	manager := NewContinuousImprovementManager(nil, monitor, logger)

	// Add test data with strategy failures
	for i := 0; i < 50; i++ {
		dataPoint := DataPoint{
			URL:          fmt.Sprintf("https://example%d.com", i),
			Success:      false,
			ErrorType:    "timeout",
			StrategyUsed: "user_agent_rotation",
		}
		monitor.RecordAttempt(context.Background(), dataPoint)
	}

	// Get failure analysis
	failureAnalysis, err := monitor.AnalyzeFailures(context.Background())
	assert.NoError(t, err)

	metrics := monitor.GetMetrics()

	// Generate strategy recommendations
	recommendations := manager.generateStrategyRecommendations(failureAnalysis, metrics)

	// Should have recommendations for strategy optimization
	assert.Greater(t, len(recommendations), 0)

	// Check that recommendations are for strategy type
	for _, rec := range recommendations {
		assert.Equal(t, "strategy", rec.Type)
		// Check for either "Optimize strategy" or "Add fallback strategy"
		assert.True(t, strings.Contains(rec.Description, "Optimize strategy") || strings.Contains(rec.Description, "Add fallback strategy"),
			"Description should contain either 'Optimize strategy' or 'Add fallback strategy', got: %s", rec.Description)
	}
}

func TestGenerateThresholdRecommendations(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)
	manager := NewContinuousImprovementManager(nil, monitor, logger)

	// Add test data with low success rate
	for i := 0; i < 50; i++ {
		dataPoint := DataPoint{
			URL:     fmt.Sprintf("https://example%d.com", i),
			Success: i%4 == 0, // 25% success rate (below 90% target)
		}
		monitor.RecordAttempt(context.Background(), dataPoint)
	}

	// Get failure analysis
	failureAnalysis, err := monitor.AnalyzeFailures(context.Background())
	assert.NoError(t, err)

	metrics := monitor.GetMetrics()

	// Generate threshold recommendations
	recommendations := manager.generateThresholdRecommendations(failureAnalysis, metrics)

	// Should have recommendations for threshold adjustment
	assert.Greater(t, len(recommendations), 0)

	// Check that recommendations are for threshold type
	for _, rec := range recommendations {
		assert.Equal(t, "threshold", rec.Type)
		assert.Contains(t, rec.Description, "Adjust verification thresholds")
	}
}

func TestGenerateRetryRecommendations(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)
	manager := NewContinuousImprovementManager(nil, monitor, logger)

	// Add test data with timeout errors
	for i := 0; i < 50; i++ {
		dataPoint := DataPoint{
			URL:       fmt.Sprintf("https://example%d.com", i),
			Success:   false,
			ErrorType: "timeout",
		}
		monitor.RecordAttempt(context.Background(), dataPoint)
	}

	// Get failure analysis
	failureAnalysis, err := monitor.AnalyzeFailures(context.Background())
	assert.NoError(t, err)

	metrics := monitor.GetMetrics()

	// Generate retry recommendations
	recommendations := manager.generateRetryRecommendations(failureAnalysis, metrics)

	// Should have recommendations for retry optimization
	assert.Greater(t, len(recommendations), 0)

	// Check that recommendations are for retry type
	for _, rec := range recommendations {
		assert.Equal(t, "retry", rec.Type)
		assert.Contains(t, rec.Description, "Optimize retry strategy")
	}
}

func TestImprovementRecommendationStruct(t *testing.T) {
	// Test that ImprovementRecommendation struct has all expected fields
	recommendation := &ImprovementRecommendation{
		ID:          "test_rec_6",
		Type:        "strategy",
		Priority:    "high",
		Description: "Test recommendation",
		Impact:      0.05,
		Confidence:  0.8,
		Parameters:  map[string]interface{}{"key": "value"},
		Reasoning:   "Test reasoning",
		CreatedAt:   time.Now(),
	}

	assert.Equal(t, "test_rec_6", recommendation.ID)
	assert.Equal(t, "strategy", recommendation.Type)
	assert.Equal(t, "high", recommendation.Priority)
	assert.Equal(t, "Test recommendation", recommendation.Description)
	assert.Equal(t, 0.05, recommendation.Impact)
	assert.Equal(t, 0.8, recommendation.Confidence)
	assert.Equal(t, "value", recommendation.Parameters["key"])
	assert.Equal(t, "Test reasoning", recommendation.Reasoning)
	assert.False(t, recommendation.CreatedAt.IsZero())
}

func TestImprovementStrategyStruct(t *testing.T) {
	// Test that ImprovementStrategy struct has all expected fields
	now := time.Now()
	strategy := &ImprovementStrategy{
		ID:          "test_strategy_1",
		Name:        "Test Strategy",
		Description: "Test strategy description",
		Type:        "strategy",
		Parameters:  map[string]interface{}{"param": "value"},
		Confidence:  0.8,
		Impact:      0.05,
		Status:      "active",
		CreatedAt:   now,
		ActivatedAt: &now,
		Metrics:     &StrategyMetrics{},
	}

	assert.Equal(t, "test_strategy_1", strategy.ID)
	assert.Equal(t, "Test Strategy", strategy.Name)
	assert.Equal(t, "Test strategy description", strategy.Description)
	assert.Equal(t, "strategy", strategy.Type)
	assert.Equal(t, "value", strategy.Parameters["param"])
	assert.Equal(t, 0.8, strategy.Confidence)
	assert.Equal(t, 0.05, strategy.Impact)
	assert.Equal(t, "active", strategy.Status)
	assert.Equal(t, now, strategy.CreatedAt)
	assert.Equal(t, &now, strategy.ActivatedAt)
	assert.NotNil(t, strategy.Metrics)
}

func TestStrategyMetricsStruct(t *testing.T) {
	// Test that StrategyMetrics struct has all expected fields
	now := time.Now()
	metrics := &StrategyMetrics{
		TotalAttempts:       100,
		SuccessfulAttempts:  80,
		FailedAttempts:      20,
		SuccessRate:         0.8,
		AverageResponseTime: 2 * time.Second,
		LastUpdated:         now,
	}

	assert.Equal(t, int64(100), metrics.TotalAttempts)
	assert.Equal(t, int64(80), metrics.SuccessfulAttempts)
	assert.Equal(t, int64(20), metrics.FailedAttempts)
	assert.Equal(t, 0.8, metrics.SuccessRate)
	assert.Equal(t, 2*time.Second, metrics.AverageResponseTime)
	assert.Equal(t, now, metrics.LastUpdated)
}

func TestStrategyEvaluationStruct(t *testing.T) {
	// Test that StrategyEvaluation struct has all expected fields
	now := time.Now()
	evaluation := &StrategyEvaluation{
		StrategyID:        "test_strategy_2",
		SuccessRateBefore: 0.7,
		SuccessRateAfter:  0.8,
		Improvement:       0.1,
		IsBeneficial:      true,
		ShouldRollback:    false,
		EvaluatedAt:       now,
	}

	assert.Equal(t, "test_strategy_2", evaluation.StrategyID)
	assert.Equal(t, 0.7, evaluation.SuccessRateBefore)
	assert.Equal(t, 0.8, evaluation.SuccessRateAfter)
	assert.Equal(t, 0.1, evaluation.Improvement)
	assert.True(t, evaluation.IsBeneficial)
	assert.False(t, evaluation.ShouldRollback)
	assert.Equal(t, now, evaluation.EvaluatedAt)
}

func TestHelperFunctions(t *testing.T) {
	// Test strategy ID generation
	strategyID1 := generateStrategyID()
	strategyID2 := generateStrategyID()

	assert.Contains(t, strategyID1, "strategy_")
	assert.Contains(t, strategyID2, "strategy_")
	assert.NotEqual(t, strategyID1, strategyID2)

	// Test recommendation ID generation
	recID1 := generateRecommendationID()
	recID2 := generateRecommendationID()

	assert.Contains(t, recID1, "rec_")
	assert.Contains(t, recID2, "rec_")
	assert.NotEqual(t, recID1, recID2)

	// Test that IDs have reasonable length
	assert.Greater(t, len(strategyID1), 10)
	assert.Greater(t, len(strategyID2), 10)
	assert.Greater(t, len(recID1), 5)
	assert.Greater(t, len(recID2), 5)
}
