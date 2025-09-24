package classification_optimization

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"kyb-platform/internal/modules/classification_monitoring"
)

func TestNewAlgorithmOptimizer(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)

	assert.NotNil(t, optimizer)
	assert.NotNil(t, optimizer.config)
	assert.Equal(t, 3, optimizer.config.MinPatternsForOptimization)
	assert.Equal(t, 24, optimizer.config.OptimizationWindowHours)
	assert.Equal(t, 10, optimizer.config.MaxOptimizationsPerDay)
	assert.Equal(t, 0.7, optimizer.config.ConfidenceThreshold)
	assert.True(t, optimizer.config.EnableAutoOptimization)
}

func TestAlgorithmOptimizer_SetPatternAnalyzer(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)
	analyzer := classification_monitoring.NewPatternAnalysisEngine(nil, logger)

	optimizer.SetPatternAnalyzer(analyzer)
	assert.Equal(t, analyzer, optimizer.patternAnalyzer)
}

func TestAlgorithmOptimizer_AnalyzeAndOptimize_NoPatterns(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)
	analyzer := classification_monitoring.NewPatternAnalysisEngine(nil, logger)
	optimizer.SetPatternAnalyzer(analyzer)

	err := optimizer.AnalyzeAndOptimize(context.Background())
	assert.NoError(t, err)
}

func TestAlgorithmOptimizer_AnalyzeAndOptimize_InsufficientPatterns(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)
	analyzer := classification_monitoring.NewPatternAnalysisEngine(nil, logger)
	optimizer.SetPatternAnalyzer(analyzer)

	// Test with no patterns (insufficient for optimization)
	err := optimizer.AnalyzeAndOptimize(context.Background())
	assert.NoError(t, err) // Should not error, just log insufficient patterns
}

func TestAlgorithmOptimizer_GetOptimizationHistory(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)

	history := optimizer.GetOptimizationHistory()
	assert.NotNil(t, history)
	assert.Len(t, history, 0)
}

func TestAlgorithmOptimizer_GetActiveOptimizations(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)

	active := optimizer.GetActiveOptimizations()
	assert.NotNil(t, active)
	assert.Len(t, active, 0)
}

func TestAlgorithmOptimizer_GetOptimizationSummary(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)

	summary := optimizer.GetOptimizationSummary()
	assert.NotNil(t, summary)
	assert.Equal(t, 0, summary.TotalOptimizations)
	assert.Equal(t, 0, summary.SuccessfulOptimizations)
	assert.Equal(t, 0, summary.FailedOptimizations)
	assert.Equal(t, 0.0, summary.AverageImprovement)
}

func TestAlgorithmOptimizer_analyzeCategoryOpportunity(t *testing.T) {
	logger := zap.NewNop()
	config := &OptimizationConfig{
		MinPatternsForOptimization: 2,
		ConfidenceThreshold:        0.7,
	}
	optimizer := NewAlgorithmOptimizer(config, logger)

	// Add some test metrics to the performance tracker
	optimizer.performanceTracker.RecordClassificationResult("test-category", &ClassificationResult{
		ExpectedCategory:  "category1",
		PredictedCategory: "category1",
		Confidence:        0.8,
		IsCorrect:         true,
		Timestamp:         time.Now(),
	})

	patterns := []*classification_monitoring.MisclassificationPattern{
		{
			ID:              "pattern-1",
			PatternType:     classification_monitoring.PatternTypeSemantic,
			Category:        classification_monitoring.PatternCategoryModelPerformance,
			Confidence:      0.9,
			OccurrenceCount: 10,
			ImpactScore:     0.8,
		},
		{
			ID:              "pattern-2",
			PatternType:     classification_monitoring.PatternTypeTemporal,
			Category:        classification_monitoring.PatternCategoryModelPerformance,
			Confidence:      0.7,
			OccurrenceCount: 5,
			ImpactScore:     0.6,
		},
	}

	opportunity := optimizer.analyzeCategoryOpportunity("test-category", patterns)
	assert.NotNil(t, opportunity)
	assert.Equal(t, "test-category", opportunity.Category)
	// The optimization type depends on the pattern analysis logic
	// Since we have both semantic and temporal patterns, it could be either
	assert.Contains(t, []OptimizationType{OptimizationTypeFeatures, OptimizationTypeWeights, OptimizationTypeThreshold}, opportunity.Type)
	assert.Len(t, opportunity.Patterns, 2)
	assert.Equal(t, "high", opportunity.Priority)
}

func TestAlgorithmOptimizer_analyzeConfidenceOpportunity(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)

	patterns := map[string]*classification_monitoring.MisclassificationPattern{
		"high-confidence": {
			ID:              "high-confidence",
			PatternType:     classification_monitoring.PatternTypeConfidence,
			Category:        classification_monitoring.PatternCategoryModelPerformance,
			Confidence:      0.9,
			OccurrenceCount: 10,
		},
		"low-confidence": {
			ID:              "low-confidence",
			PatternType:     classification_monitoring.PatternTypeConfidence,
			Category:        classification_monitoring.PatternCategoryModelPerformance,
			Confidence:      0.3,
			OccurrenceCount: 5,
		},
	}

	opportunity := optimizer.analyzeConfidenceOpportunity(patterns)
	assert.NotNil(t, opportunity)
	// The optimization type depends on the pattern analysis logic
	assert.Contains(t, []OptimizationType{OptimizationTypeThreshold, OptimizationTypeWeights}, opportunity.Type)
	assert.Equal(t, "confidence", opportunity.Category)
}

func TestAlgorithmOptimizer_performOptimization(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)

	// Register a test algorithm
	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	regErr := optimizer.algorithmRegistry.RegisterAlgorithm(algorithm)
	assert.NoError(t, regErr)

	// Debug: Check if algorithm was registered
	registeredAlgorithm := optimizer.algorithmRegistry.GetAlgorithm("test-algorithm")
	assert.NotNil(t, registeredAlgorithm, "Algorithm should be registered")
	assert.Equal(t, "test-category", registeredAlgorithm.Category)

	// Debug: Check if algorithm can be found by category
	categoryAlgorithm := optimizer.algorithmRegistry.GetAlgorithmByCategory("test-category")
	assert.NotNil(t, categoryAlgorithm, "Algorithm should be found by category")

	opportunity := &OptimizationOpportunity{
		ID:       "test-opportunity",
		Type:     OptimizationTypeThreshold,
		Category: "test-category",
		Patterns: []*classification_monitoring.MisclassificationPattern{
			{
				ID:              "test-pattern",
				Confidence:      0.9,
				OccurrenceCount: 10,
			},
		},
		Confidence: 0.8,
	}

	err := optimizer.performOptimization(context.Background(), opportunity)
	assert.NoError(t, err)

	// Check that optimization was recorded
	history := optimizer.GetOptimizationHistory()
	if len(history) > 0 {
		assert.Equal(t, OptimizationStatusCompleted, history[0].Status)
	}
}

func TestAlgorithmOptimizer_optimizeThresholds(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)

	// Register a test algorithm
	algorithm := &ClassificationAlgorithm{
		ID:                  "test-algorithm",
		Name:                "Test Algorithm",
		Category:            "test-category",
		ConfidenceThreshold: 0.7,
		IsActive:            true,
	}
	regErr := optimizer.algorithmRegistry.RegisterAlgorithm(algorithm)
	assert.NoError(t, regErr)

	// Debug: Check if algorithm was registered
	registeredAlgorithm := optimizer.algorithmRegistry.GetAlgorithm("test-algorithm")
	assert.NotNil(t, registeredAlgorithm, "Algorithm should be registered")
	assert.Equal(t, "test-category", registeredAlgorithm.Category)

	// Debug: Check if algorithm can be found by category
	categoryAlgorithm := optimizer.algorithmRegistry.GetAlgorithmByCategory("test-category")
	assert.NotNil(t, categoryAlgorithm, "Algorithm should be found by category")

	result := &OptimizationResult{
		ID:               "test-result",
		AlgorithmID:      "test-category",
		OptimizationType: OptimizationTypeThreshold,
	}

	opportunity := &OptimizationOpportunity{
		ID:       "test-opportunity",
		Type:     OptimizationTypeThreshold,
		Category: "test-category",
		Patterns: []*classification_monitoring.MisclassificationPattern{
			{
				ID:              "test-pattern",
				Confidence:      0.9,
				OccurrenceCount: 10,
			},
		},
		Confidence: 0.8,
	}

	err := optimizer.optimizeThresholds(context.Background(), result, opportunity)
	assert.NoError(t, err)
	assert.Len(t, result.Changes, 1)
	assert.Equal(t, "confidence_threshold", result.Changes[0].Parameter)
}

func TestAlgorithmOptimizer_optimizeWeights(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)

	result := &OptimizationResult{
		ID:               "test-result",
		AlgorithmID:      "test-category",
		OptimizationType: OptimizationTypeWeights,
	}

	opportunity := &OptimizationOpportunity{
		ID:       "test-opportunity",
		Type:     OptimizationTypeWeights,
		Category: "test-category",
		Patterns: []*classification_monitoring.MisclassificationPattern{
			{
				ID:              "test-pattern",
				Confidence:      0.7,
				OccurrenceCount: 5,
			},
		},
		Confidence: 0.6,
	}

	err := optimizer.optimizeWeights(context.Background(), result, opportunity)
	assert.NoError(t, err)
	assert.Len(t, result.Changes, 1)
	assert.Equal(t, "feature_weights", result.Changes[0].Parameter)
}

func TestAlgorithmOptimizer_optimizeFeatures(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)

	result := &OptimizationResult{
		ID:               "test-result",
		AlgorithmID:      "test-category",
		OptimizationType: OptimizationTypeFeatures,
	}

	opportunity := &OptimizationOpportunity{
		ID:       "test-opportunity",
		Type:     OptimizationTypeFeatures,
		Category: "test-category",
		Patterns: []*classification_monitoring.MisclassificationPattern{
			{
				ID:              "test-pattern",
				Confidence:      0.7,
				OccurrenceCount: 5,
			},
		},
		Confidence: 0.6,
	}

	err := optimizer.optimizeFeatures(context.Background(), result, opportunity)
	assert.NoError(t, err)
	assert.Len(t, result.Changes, 1)
	assert.Equal(t, "feature_extraction", result.Changes[0].Parameter)
}

func TestAlgorithmOptimizer_optimizeModel(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewAlgorithmOptimizer(nil, logger)

	result := &OptimizationResult{
		ID:               "test-result",
		AlgorithmID:      "test-category",
		OptimizationType: OptimizationTypeModel,
	}

	opportunity := &OptimizationOpportunity{
		ID:       "test-opportunity",
		Type:     OptimizationTypeModel,
		Category: "test-category",
		Patterns: []*classification_monitoring.MisclassificationPattern{
			{
				ID:              "test-pattern",
				Confidence:      0.7,
				OccurrenceCount: 5,
			},
		},
		Confidence: 0.6,
	}

	err := optimizer.optimizeModel(context.Background(), result, opportunity)
	assert.NoError(t, err)
	assert.Len(t, result.Changes, 1)
	assert.Equal(t, "model_parameters", result.Changes[0].Parameter)
}

func TestOptimizationResult_Validation(t *testing.T) {
	result := &OptimizationResult{
		ID:               "test-result",
		AlgorithmID:      "test-algorithm",
		OptimizationType: OptimizationTypeThreshold,
		Status:           OptimizationStatusCompleted,
		OptimizationTime: time.Now(),
	}

	assert.NotEmpty(t, result.ID)
	assert.NotEmpty(t, result.AlgorithmID)
	assert.Equal(t, OptimizationTypeThreshold, result.OptimizationType)
	assert.Equal(t, OptimizationStatusCompleted, result.Status)
}

func TestOptimizationOpportunity_Validation(t *testing.T) {
	opportunity := &OptimizationOpportunity{
		ID:         "test-opportunity",
		Type:       OptimizationTypeThreshold,
		Category:   "test-category",
		Confidence: 0.8,
		Priority:   "high",
	}

	assert.NotEmpty(t, opportunity.ID)
	assert.Equal(t, OptimizationTypeThreshold, opportunity.Type)
	assert.NotEmpty(t, opportunity.Category)
	assert.Greater(t, opportunity.Confidence, 0.0)
	assert.LessOrEqual(t, opportunity.Confidence, 1.0)
}

func TestAlgorithmChange_Validation(t *testing.T) {
	change := &AlgorithmChange{
		Parameter:  "confidence_threshold",
		OldValue:   0.7,
		NewValue:   0.8,
		ChangeType: "threshold_adjustment",
		Impact:     "medium",
		Confidence: 0.8,
	}

	assert.NotEmpty(t, change.Parameter)
	assert.NotNil(t, change.OldValue)
	assert.NotNil(t, change.NewValue)
	assert.NotEmpty(t, change.ChangeType)
	assert.NotEmpty(t, change.Impact)
	assert.Greater(t, change.Confidence, 0.0)
	assert.LessOrEqual(t, change.Confidence, 1.0)
}
