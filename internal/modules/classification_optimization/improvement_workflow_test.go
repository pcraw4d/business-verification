package classification_optimization

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewImprovementWorkflow(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	assert.NotNil(t, workflow)
	assert.NotNil(t, workflow.config)
	assert.True(t, workflow.config.AutoImprovementEnabled)
	assert.Equal(t, 24*time.Hour, workflow.config.ImprovementInterval)
	assert.Equal(t, 0.85, workflow.config.AccuracyThreshold)
	assert.Equal(t, 0.8, workflow.config.ConfidenceThreshold)
	assert.Equal(t, 10, workflow.config.MaxIterations)
	assert.Equal(t, 0.01, workflow.config.ConvergenceThreshold)
	assert.True(t, workflow.config.EnableABTesting)
	assert.Equal(t, 0.2, workflow.config.TestSplitRatio)
}

func TestImprovementWorkflow_SetDependencies(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	algorithmRegistry := NewAlgorithmRegistry(logger)
	performanceTracker := NewPerformanceTracker(logger)
	accuracyValidator := NewAccuracyValidator(nil, logger)

	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	assert.NotNil(t, workflow.algorithmRegistry)
	assert.NotNil(t, workflow.performanceTracker)
	assert.NotNil(t, workflow.accuracyValidator)
}

func TestImprovementWorkflow_StartContinuousImprovement_Success(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	// Set up dependencies
	algorithmRegistry := NewAlgorithmRegistry(logger)
	performanceTracker := NewPerformanceTracker(logger)
	accuracyValidator := NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Register a test algorithm
	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithm)

	// Start continuous improvement
	execution, err := workflow.StartContinuousImprovement(context.Background(), "test-algorithm")

	assert.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, "test-algorithm", execution.AlgorithmID)
	assert.Equal(t, WorkflowStatusCompleted, execution.Status)
	assert.Equal(t, WorkflowTypeContinuousImprovement, execution.Type)
	assert.NotNil(t, execution.BaselineMetrics)
	assert.NotNil(t, execution.FinalMetrics)
	assert.GreaterOrEqual(t, len(execution.Iterations), 0)
}

func TestImprovementWorkflow_StartContinuousImprovement_AlgorithmNotFound(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	// Set up dependencies
	algorithmRegistry := NewAlgorithmRegistry(logger)
	performanceTracker := NewPerformanceTracker(logger)
	accuracyValidator := NewAccuracyValidator(nil, logger)

	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Start continuous improvement with non-existent algorithm
	execution, err := workflow.StartContinuousImprovement(context.Background(), "non-existent")

	assert.Error(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, WorkflowStatusFailed, execution.Status)
	assert.Contains(t, execution.Error, "algorithm not found")
}

func TestImprovementWorkflow_StartABTesting_Success(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	// Set up dependencies
	algorithmRegistry := NewAlgorithmRegistry(logger)
	performanceTracker := NewPerformanceTracker(logger)
	accuracyValidator := NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Register test algorithms
	algorithmA := &ClassificationAlgorithm{
		ID:                  "algorithm-a",
		Name:                "Algorithm A",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmB := &ClassificationAlgorithm{
		ID:                  "algorithm-b",
		Name:                "Algorithm B",
		Category:            "test-category",
		ConfidenceThreshold: 0.8,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithmA)
	algorithmRegistry.RegisterAlgorithm(algorithmB)

	// Create test cases - need more to account for splitting (0.2 ratio means we need at least 500 total)
	testCases := make([]*TestCase, 500)
	for i := 0; i < 500; i++ {
		testCases[i] = &TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	// Start A/B testing
	execution, err := workflow.StartABTesting(context.Background(), "algorithm-a", "algorithm-b", testCases)

	assert.NoError(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, "algorithm-a", execution.AlgorithmID)
	assert.Equal(t, WorkflowStatusCompleted, execution.Status)
	assert.Equal(t, WorkflowTypeABTesting, execution.Type)
	assert.Equal(t, 1, len(execution.Iterations))
	assert.NotNil(t, execution.FinalMetrics)
}

func TestImprovementWorkflow_StartABTesting_AlgorithmNotFound(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	// Set up dependencies
	algorithmRegistry := NewAlgorithmRegistry(logger)
	performanceTracker := NewPerformanceTracker(logger)
	accuracyValidator := NewAccuracyValidator(nil, logger)

	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	// Create test cases
	testCases := make([]*TestCase, 10)
	for i := 0; i < 10; i++ {
		testCases[i] = &TestCase{
			ID:             fmt.Sprintf("test-%d", i),
			Input:          map[string]interface{}{"name": fmt.Sprintf("Company %d", i)},
			ExpectedOutput: "technology",
			TestCaseType:   "standard",
			Difficulty:     "easy",
		}
	}

	// Start A/B testing with non-existent algorithms
	execution, err := workflow.StartABTesting(context.Background(), "algorithm-a", "algorithm-b", testCases)

	assert.Error(t, err)
	assert.NotNil(t, execution)
	assert.Equal(t, WorkflowStatusFailed, execution.Status)
	assert.Contains(t, execution.Error, "algorithms not found")
}

func TestImprovementWorkflow_GetWorkflowHistory(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	// Initially empty
	history := workflow.GetWorkflowHistory()
	assert.Empty(t, history)

	// Set up dependencies and run a workflow
	algorithmRegistry := NewAlgorithmRegistry(logger)
	performanceTracker := NewPerformanceTracker(logger)
	accuracyValidator := NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithm)

	_, err := workflow.StartContinuousImprovement(context.Background(), "test-algorithm")
	assert.NoError(t, err)

	// Check history
	history = workflow.GetWorkflowHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, "test-algorithm", history[0].AlgorithmID)
}

func TestImprovementWorkflow_GetActiveWorkflows(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	// Initially empty
	active := workflow.GetActiveWorkflows()
	assert.Empty(t, active)

	// Set up dependencies
	algorithmRegistry := NewAlgorithmRegistry(logger)
	performanceTracker := NewPerformanceTracker(logger)
	accuracyValidator := NewAccuracyValidator(nil, logger)
	accuracyValidator.SetAlgorithmRegistry(algorithmRegistry)

	workflow.SetDependencies(algorithmRegistry, performanceTracker, accuracyValidator, nil)

	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	algorithmRegistry.RegisterAlgorithm(algorithm)

	// Run workflow (should complete immediately in test)
	_, err := workflow.StartContinuousImprovement(context.Background(), "test-algorithm")
	assert.NoError(t, err)

	// Should be empty after completion
	active = workflow.GetActiveWorkflows()
	assert.Empty(t, active)
}

func TestImprovementWorkflow_StopWorkflow(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	// Try to stop non-existent workflow
	err := workflow.StopWorkflow("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "workflow not found")
}

func TestImprovementWorkflow_establishBaseline(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}

	baseline, err := workflow.establishBaseline(context.Background(), algorithm)

	assert.NoError(t, err)
	assert.NotNil(t, baseline)
	assert.Equal(t, 0.75, baseline.Accuracy)
	assert.Equal(t, 0.70, baseline.F1Score)
	assert.Equal(t, 0.80, baseline.AverageConfidence)
}

func TestImprovementWorkflow_analyzeOptimizationOpportunities(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}

	// Test with low accuracy
	metrics := &ValidationMetrics{
		Accuracy:          0.70, // Below threshold
		F1Score:           0.65,
		AverageConfidence: 0.85,
	}

	opportunities, err := workflow.analyzeOptimizationOpportunities(algorithm, metrics)

	assert.NoError(t, err)
	assert.Len(t, opportunities, 1)
	assert.Equal(t, OptimizationTypeFeatures, opportunities[0].Type)
	assert.Equal(t, "high", opportunities[0].Priority)

	// Test with low confidence
	metrics = &ValidationMetrics{
		Accuracy:          0.90,
		F1Score:           0.85,
		AverageConfidence: 0.70, // Below threshold
	}

	opportunities, err = workflow.analyzeOptimizationOpportunities(algorithm, metrics)

	assert.NoError(t, err)
	assert.Len(t, opportunities, 1)
	assert.Equal(t, OptimizationTypeThreshold, opportunities[0].Type)
	assert.Equal(t, "medium", opportunities[0].Priority)
}

func TestImprovementWorkflow_applyOptimizations(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}

	opportunities := []*OptimizationOpportunity{
		{
			ID:       "opp_1",
			Type:     OptimizationTypeThreshold,
			Category: "test-category",
			Priority: "high",
			Actions:  []string{"adjust threshold"},
		},
		{
			ID:       "opp_2",
			Type:     OptimizationTypeFeatures,
			Category: "test-category",
			Priority: "medium",
			Actions:  []string{"optimize features"},
		},
	}

	changes, err := workflow.applyOptimizations(algorithm, opportunities)

	assert.NoError(t, err)
	assert.Len(t, changes, 2)
	assert.Equal(t, "confidence_threshold", changes[0].Parameter)
	assert.Equal(t, "features", changes[1].Parameter)
}

func TestImprovementWorkflow_applyOptimization(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}

	opportunity := &OptimizationOpportunity{
		ID:       "opp_1",
		Type:     OptimizationTypeThreshold,
		Category: "test-category",
		Priority: "high",
		Actions:  []string{"adjust threshold"},
	}

	change, err := workflow.applyOptimization(algorithm, opportunity)

	assert.NoError(t, err)
	assert.NotNil(t, change)
	assert.Equal(t, "confidence_threshold", change.Parameter)
	assert.Equal(t, "threshold_optimization", change.ChangeType)
	assert.Equal(t, 0.8, change.Confidence)

	// Test unsupported optimization type
	opportunity.Type = "unsupported"
	change, err = workflow.applyOptimization(algorithm, opportunity)

	assert.Error(t, err)
	assert.Nil(t, change)
	assert.Contains(t, err.Error(), "unsupported optimization type")
}

func TestImprovementWorkflow_optimizeThresholds(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.8,
		IsActive:            true,
	}

	opportunity := &OptimizationOpportunity{
		ID:       "opp_1",
		Type:     OptimizationTypeThreshold,
		Category: "test-category",
		Priority: "high",
		Actions:  []string{"adjust threshold"},
	}

	change, err := workflow.optimizeThresholds(algorithm, opportunity)

	assert.NoError(t, err)
	assert.NotNil(t, change)
	assert.Equal(t, "confidence_threshold", change.Parameter)
	assert.Equal(t, "0.800", change.OldValue)
	assert.Equal(t, "0.720", change.NewValue) // 0.8 * 0.9
	assert.Equal(t, "threshold_optimization", change.ChangeType)
	assert.Equal(t, "improved recall", change.Impact)
	assert.Equal(t, 0.8, change.Confidence)
}

func TestImprovementWorkflow_optimizeWeights(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}

	opportunity := &OptimizationOpportunity{
		ID:       "opp_1",
		Type:     OptimizationTypeWeights,
		Category: "test-category",
		Priority: "high",
		Actions:  []string{"optimize weights"},
	}

	change, err := workflow.optimizeWeights(algorithm, opportunity)

	assert.NoError(t, err)
	assert.NotNil(t, change)
	assert.Equal(t, "weights", change.Parameter)
	assert.Equal(t, "default_weights", change.OldValue)
	assert.Equal(t, "optimized_weights", change.NewValue)
	assert.Equal(t, "weight_optimization", change.ChangeType)
	assert.Equal(t, "improved accuracy", change.Impact)
	assert.Equal(t, 0.7, change.Confidence)
}

func TestImprovementWorkflow_optimizeFeatures(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}

	opportunity := &OptimizationOpportunity{
		ID:       "opp_1",
		Type:     OptimizationTypeFeatures,
		Category: "test-category",
		Priority: "high",
		Actions:  []string{"optimize features"},
	}

	change, err := workflow.optimizeFeatures(algorithm, opportunity)

	assert.NoError(t, err)
	assert.NotNil(t, change)
	assert.Equal(t, "features", change.Parameter)
	assert.Equal(t, "basic_features", change.OldValue)
	assert.Equal(t, "enhanced_features", change.NewValue)
	assert.Equal(t, "feature_optimization", change.ChangeType)
	assert.Equal(t, "improved feature extraction", change.Impact)
	assert.Equal(t, 0.6, change.Confidence)
}

func TestImprovementWorkflow_calculateImprovementScore(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	baseline := &ValidationMetrics{
		Accuracy:          0.75,
		F1Score:           0.70,
		AverageConfidence: 0.80,
	}

	final := &ValidationMetrics{
		Accuracy:          0.85,
		F1Score:           0.80,
		AverageConfidence: 0.90,
	}

	score := workflow.calculateImprovementScore(baseline, final)

	// Expected: (0.1 * 0.5) + (0.1 * 0.3) + (0.1 * 0.2) = 0.1
	assert.Equal(t, 0.1, score)

	// Test with nil metrics
	score = workflow.calculateImprovementScore(nil, final)
	assert.Equal(t, 0.0, score)

	score = workflow.calculateImprovementScore(baseline, nil)
	assert.Equal(t, 0.0, score)
}

func TestImprovementWorkflow_generateWorkflowRecommendations(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	// Test with significant improvement
	execution := &WorkflowExecution{
		ID:               "test-execution",
		AlgorithmID:      "test-algorithm",
		ImprovementScore: 0.1, // Significant improvement
		Iterations:       make([]*WorkflowIteration, 5),
	}

	recommendations := workflow.generateWorkflowRecommendations(execution)

	assert.Len(t, recommendations, 1)
	assert.Equal(t, "success", recommendations[0].Type)
	assert.Equal(t, "high", recommendations[0].Priority)
	assert.Contains(t, recommendations[0].Description, "Significant improvement")

	// Test with performance degradation
	execution.ImprovementScore = -0.1
	recommendations = workflow.generateWorkflowRecommendations(execution)

	assert.Len(t, recommendations, 1)
	assert.Equal(t, "warning", recommendations[0].Type)
	assert.Equal(t, "high", recommendations[0].Priority)
	assert.Contains(t, recommendations[0].Description, "Performance degradation")

	// Test with maximum iterations
	execution.ImprovementScore = 0.0
	execution.Iterations = make([]*WorkflowIteration, 10) // Max iterations
	recommendations = workflow.generateWorkflowRecommendations(execution)

	assert.Len(t, recommendations, 1)
	assert.Equal(t, "info", recommendations[0].Type)
	assert.Equal(t, "medium", recommendations[0].Priority)
	assert.Contains(t, recommendations[0].Description, "Maximum iterations")
}

func TestImprovementWorkflow_generateTestCases(t *testing.T) {
	logger := zap.NewNop()
	workflow := NewImprovementWorkflow(nil, logger)

	testCases := workflow.generateTestCases()

	assert.Len(t, testCases, 100)
	for i, testCase := range testCases {
		assert.Equal(t, fmt.Sprintf("test_%d", i), testCase.ID)
		assert.Equal(t, "technology", testCase.ExpectedOutput)
		assert.Equal(t, "standard", testCase.TestCaseType)
		assert.Equal(t, "easy", testCase.Difficulty)
		assert.Contains(t, testCase.Input, "name")
	}
}
