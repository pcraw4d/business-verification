package industry_codes

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewVotingOptimizer(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	optimizer := NewVotingOptimizer(nil, logger)
	assert.NotNil(t, optimizer)
	assert.NotNil(t, optimizer.config)
	assert.Equal(t, true, optimizer.config.EnableAutoOptimization)
	assert.Equal(t, 1*time.Hour, optimizer.config.OptimizationInterval)
	assert.Equal(t, 100, optimizer.config.MinSamplesForOptimization)

	// Test with custom config
	customConfig := &VotingOptimizationConfig{
		EnableAutoOptimization:       false,
		OptimizationInterval:         2 * time.Hour,
		MinSamplesForOptimization:    200,
		MaxOptimizationsPerDay:       12,
		OptimizationTimeout:          10 * time.Minute,
		MinAccuracyImprovement:       0.03,
		MinConfidenceImprovement:     0.02,
		MaxPerformanceRegression:     0.03,
		MinVotingScoreImprovement:    0.02,
		EnableStrategyOptimization:   false,
		EnableWeightOptimization:     false,
		EnableThresholdOptimization:  false,
		EnableOutlierOptimization:    false,
		EnableAdaptiveLearning:       false,
		LearningRate:                 0.05,
		AdaptationThreshold:          0.03,
		PerformanceDecayFactor:       0.9,
		EnableOptimizationValidation: false,
		ValidationWindow:             15 * time.Minute,
		RollbackThreshold:            0.05,
		MaxRollbackAttempts:          5,
	}

	optimizer = NewVotingOptimizer(customConfig, logger)
	assert.NotNil(t, optimizer)
	assert.Equal(t, customConfig, optimizer.config)
	assert.Equal(t, false, optimizer.config.EnableAutoOptimization)
	assert.Equal(t, 2*time.Hour, optimizer.config.OptimizationInterval)
	assert.Equal(t, 200, optimizer.config.MinSamplesForOptimization)
}

func TestVotingOptimizer_SetVotingComponents(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Create mock components
	votingEngine := &VotingEngine{
		config: &VotingConfig{
			Strategy:               VotingStrategyWeightedAverage,
			MinVoters:              2,
			RequiredAgreement:      0.5,
			ConfidenceWeight:       0.4,
			ConsistencyWeight:      0.3,
			DiversityWeight:        0.3,
			EnableTieBreaking:      true,
			EnableOutlierFiltering: true,
			OutlierThreshold:       2.0,
		},
		logger: logger,
	}

	votingValidator := NewVotingValidator(nil, logger)
	confidenceCalculator := NewConfidenceCalculator(nil, logger)

	// Set components
	optimizer.SetVotingComponents(votingEngine, votingValidator, confidenceCalculator)

	assert.Equal(t, votingEngine, optimizer.votingEngine)
	assert.Equal(t, votingValidator, optimizer.votingValidator)
	assert.Equal(t, confidenceCalculator, optimizer.confidenceCalculator)
}

func TestVotingOptimizer_RecordVotingPerformance(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Create test metrics
	metrics := &VotingPerformanceMetrics{
		OverallAccuracy:       0.85,
		Top1Accuracy:          0.80,
		Top3Accuracy:          0.90,
		Top5Accuracy:          0.95,
		AverageConfidence:     0.75,
		ConfidenceVariance:    0.15,
		HighConfidenceRate:    0.70,
		VotingScore:           0.80,
		AgreementScore:        0.75,
		ConsistencyScore:      0.80,
		DiversityScore:        0.85,
		AverageProcessingTime: 100 * time.Millisecond,
		Throughput:            100.0,
		ErrorRate:             0.05,
		StrategyPerformance: map[string]*StrategyPerformanceMetrics{
			"keyword": {
				StrategyName:         "keyword",
				AverageAccuracy:      0.80,
				AverageConfidence:    0.75,
				PerformanceScore:     0.95,
				TotalClassifications: 100,
				SuccessfulMatches:    95,
			},
			"ml": {
				StrategyName:         "ml",
				AverageAccuracy:      0.85,
				AverageConfidence:    0.80,
				PerformanceScore:     0.90,
				TotalClassifications: 100,
				SuccessfulMatches:    90,
			},
		},
		Timestamp: time.Now(),
	}

	// Record performance
	optimizer.RecordVotingPerformance(metrics)

	// Verify metrics were recorded
	history := optimizer.GetPerformanceHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, metrics.OverallAccuracy, history[0].OverallAccuracy)
	assert.Equal(t, metrics.AverageConfidence, history[0].AverageConfidence)
	assert.Equal(t, metrics.VotingScore, history[0].VotingScore)
}

func TestVotingOptimizer_AnalyzeOptimizationOpportunities(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Create test metrics with optimization opportunities
	metrics := &VotingPerformanceMetrics{
		OverallAccuracy:    0.65, // Below threshold
		AverageConfidence:  0.60, // Below threshold
		VotingScore:        0.55, // Below threshold
		AgreementScore:     0.40, // Below threshold
		ConsistencyScore:   0.50, // Below threshold
		ConfidenceVariance: 0.45, // Above threshold for outlier filtering enhancement
		StrategyPerformance: map[string]*StrategyPerformanceMetrics{
			"keyword": {
				StrategyName:      "keyword",
				AverageAccuracy:   0.60, // Below threshold
				AverageConfidence: 0.65,
				PerformanceScore:  0.65,
			},
			"ml": {
				StrategyName:      "ml",
				AverageAccuracy:   0.70, // Below threshold
				AverageConfidence: 0.75,
				PerformanceScore:  0.75,
			},
		},
		Timestamp: time.Now(),
	}

	// Analyze opportunities
	opportunities, err := optimizer.analyzeOptimizationOpportunities(metrics)
	require.NoError(t, err)
	assert.NotEmpty(t, opportunities)

	// Verify different types of opportunities were identified
	opportunityTypes := make(map[string]bool)
	for _, opp := range opportunities {
		opportunityTypes[opp.Type] = true
	}

	// Should have strategy improvement opportunities
	assert.True(t, opportunityTypes["strategy_improvement"])

	// Should have weight optimization opportunities
	assert.True(t, opportunityTypes["confidence_weight_adjustment"])
	assert.True(t, opportunityTypes["consistency_weight_adjustment"])

	// Should have threshold optimization opportunities
	assert.True(t, opportunityTypes["agreement_threshold_adjustment"])
	assert.True(t, opportunityTypes["outlier_threshold_adjustment"])

	// Should have outlier optimization opportunities
	assert.True(t, opportunityTypes["outlier_filtering_enhancement"])

	// Log actual opportunity types for debugging
	t.Logf("Actual opportunity types: %v", opportunityTypes)
}

func TestVotingOptimizer_AnalyzeStrategyOptimization(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Create test metrics with underperforming strategies
	metrics := &VotingPerformanceMetrics{
		StrategyPerformance: map[string]*StrategyPerformanceMetrics{
			"keyword": {
				StrategyName:      "keyword",
				AverageAccuracy:   0.60, // Below threshold
				AverageConfidence: 0.65,
				PerformanceScore:  0.65,
			},
			"ml": {
				StrategyName:      "ml",
				AverageAccuracy:   0.85, // Above threshold
				AverageConfidence: 0.80,
				PerformanceScore:  0.85,
			},
			"website": {
				StrategyName:      "website",
				AverageAccuracy:   0.55, // Below threshold
				AverageConfidence: 0.60,
				PerformanceScore:  0.55,
			},
		},
	}

	// Set voting engine for weight rebalancing analysis
	votingEngine := &VotingEngine{
		config: &VotingConfig{
			Strategy:               VotingStrategyWeightedAverage,
			MinVoters:              2,
			RequiredAgreement:      0.5,
			ConfidenceWeight:       0.4,
			ConsistencyWeight:      0.3,
			DiversityWeight:        0.3,
			EnableTieBreaking:      true,
			EnableOutlierFiltering: true,
			OutlierThreshold:       2.0,
		},
		logger: logger,
	}
	optimizer.SetVotingComponents(votingEngine, nil, nil)

	// Analyze strategy optimization opportunities
	opportunities := optimizer.analyzeStrategyOptimization(metrics)

	// Should have opportunities for underperforming strategies
	assert.Len(t, opportunities, 3) // 2 strategy improvements + 1 weight rebalancing

	// Verify strategy improvement opportunities
	strategyImprovements := 0
	for _, opp := range opportunities {
		if opp.Type == "strategy_improvement" {
			strategyImprovements++
			assert.Contains(t, opp.Description, "Improve performance of strategy")
			assert.Equal(t, 0.1, opp.PotentialImpact)
			assert.Equal(t, 0.8, opp.Confidence)
			assert.Equal(t, "medium", opp.Effort)
			assert.Equal(t, 2, opp.Priority)
		}
	}
	assert.Equal(t, 2, strategyImprovements) // keyword and website strategies

	// Verify weight rebalancing opportunity
	weightRebalancing := 0
	for _, opp := range opportunities {
		if opp.Type == "weight_rebalancing" {
			weightRebalancing++
			assert.Equal(t, "Rebalance strategy weights based on performance", opp.Description)
			assert.Equal(t, 0.05, opp.PotentialImpact)
			assert.Equal(t, 0.7, opp.Confidence)
			assert.Equal(t, "low", opp.Effort)
			assert.Equal(t, 3, opp.Priority)
		}
	}
	assert.Equal(t, 1, weightRebalancing)
}

func TestVotingOptimizer_AnalyzeWeightOptimization(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Test with low confidence
	metrics := &VotingPerformanceMetrics{
		AverageConfidence: 0.60, // Below threshold
		ConsistencyScore:  0.50, // Below threshold
	}

	opportunities := optimizer.analyzeWeightOptimization(metrics)
	assert.Len(t, opportunities, 2)

	// Verify confidence weight adjustment opportunity
	confidenceOpp := opportunities[0]
	assert.Equal(t, "confidence_weight_adjustment", confidenceOpp.Type)
	assert.Equal(t, "Adjust confidence weight to improve overall confidence", confidenceOpp.Description)
	assert.Equal(t, 0.03, confidenceOpp.PotentialImpact)
	assert.Equal(t, 0.6, confidenceOpp.Confidence)
	assert.Equal(t, "low", confidenceOpp.Effort)
	assert.Equal(t, 4, confidenceOpp.Priority)

	// Verify consistency weight adjustment opportunity
	consistencyOpp := opportunities[1]
	assert.Equal(t, "consistency_weight_adjustment", consistencyOpp.Type)
	assert.Equal(t, "Adjust consistency weight to improve result consistency", consistencyOpp.Description)
	assert.Equal(t, 0.04, consistencyOpp.PotentialImpact)
	assert.Equal(t, 0.7, consistencyOpp.Confidence)
	assert.Equal(t, "low", consistencyOpp.Effort)
	assert.Equal(t, 3, consistencyOpp.Priority)
}

func TestVotingOptimizer_AnalyzeThresholdOptimization(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Test with low agreement and high variance
	metrics := &VotingPerformanceMetrics{
		AgreementScore:     0.40, // Below threshold
		ConfidenceVariance: 0.35, // Above threshold
	}

	opportunities := optimizer.analyzeThresholdOptimization(metrics)
	assert.Len(t, opportunities, 2)

	// Verify agreement threshold adjustment opportunity
	agreementOpp := opportunities[0]
	assert.Equal(t, "agreement_threshold_adjustment", agreementOpp.Type)
	assert.Equal(t, "Adjust required agreement threshold to improve consensus", agreementOpp.Description)
	assert.Equal(t, 0.06, agreementOpp.PotentialImpact)
	assert.Equal(t, 0.8, agreementOpp.Confidence)
	assert.Equal(t, "low", agreementOpp.Effort)
	assert.Equal(t, 2, agreementOpp.Priority)

	// Verify outlier threshold adjustment opportunity
	outlierOpp := opportunities[1]
	assert.Equal(t, "outlier_threshold_adjustment", outlierOpp.Type)
	assert.Equal(t, "Adjust outlier threshold to reduce variance", outlierOpp.Description)
	assert.Equal(t, 0.04, outlierOpp.PotentialImpact)
	assert.Equal(t, 0.7, outlierOpp.Confidence)
	assert.Equal(t, "low", outlierOpp.Effort)
	assert.Equal(t, 3, outlierOpp.Priority)
}

func TestVotingOptimizer_AnalyzeOutlierOptimization(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Test with high variance
	metrics := &VotingPerformanceMetrics{
		ConfidenceVariance: 0.45, // Above threshold
	}

	opportunities := optimizer.analyzeOutlierOptimization(metrics)
	assert.Len(t, opportunities, 1)

	// Verify outlier filtering enhancement opportunity
	outlierOpp := opportunities[0]
	assert.Equal(t, "outlier_filtering_enhancement", outlierOpp.Type)
	assert.Equal(t, "Enhance outlier filtering to reduce confidence variance", outlierOpp.Description)
	assert.Equal(t, 0.05, outlierOpp.PotentialImpact)
	assert.Equal(t, 0.8, outlierOpp.Confidence)
	assert.Equal(t, "medium", outlierOpp.Effort)
	assert.Equal(t, 2, outlierOpp.Priority)
}

func TestVotingOptimizer_ApplyOptimizations(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Set up voting engine
	votingEngine := &VotingEngine{
		config: &VotingConfig{
			Strategy:               VotingStrategyWeightedAverage,
			MinVoters:              2,
			RequiredAgreement:      0.5,
			ConfidenceWeight:       0.4,
			ConsistencyWeight:      0.3,
			DiversityWeight:        0.3,
			EnableTieBreaking:      true,
			EnableOutlierFiltering: false,
			OutlierThreshold:       2.0,
		},
		logger: logger,
	}
	optimizer.SetVotingComponents(votingEngine, nil, nil)

	// Create test opportunities
	opportunities := []*OptimizationOpportunity{
		{
			Type:            "confidence_weight_adjustment",
			Description:     "Test confidence weight adjustment",
			PotentialImpact: 0.03,
			Confidence:      0.6,
			Effort:          "low",
			Priority:        4,
			Parameters: map[string]interface{}{
				"current_confidence": 0.60,
				"target_confidence":  0.75,
			},
		},
		{
			Type:            "consistency_weight_adjustment",
			Description:     "Test consistency weight adjustment",
			PotentialImpact: 0.04,
			Confidence:      0.7,
			Effort:          "low",
			Priority:        3,
			Parameters: map[string]interface{}{
				"current_consistency": 0.50,
				"target_consistency":  0.70,
			},
		},
		{
			Type:            "agreement_threshold_adjustment",
			Description:     "Test agreement threshold adjustment",
			PotentialImpact: 0.06,
			Confidence:      0.8,
			Effort:          "low",
			Priority:        2,
			Parameters: map[string]interface{}{
				"current_agreement": 0.40,
				"target_agreement":  0.60,
			},
		},
		{
			Type:            "outlier_threshold_adjustment",
			Description:     "Test outlier threshold adjustment",
			PotentialImpact: 0.04,
			Confidence:      0.7,
			Effort:          "low",
			Priority:        3,
			Parameters: map[string]interface{}{
				"current_variance": 0.35,
				"target_variance":  0.20,
			},
		},
		{
			Type:            "outlier_filtering_enhancement",
			Description:     "Test outlier filtering enhancement",
			PotentialImpact: 0.05,
			Confidence:      0.8,
			Effort:          "medium",
			Priority:        2,
			Parameters: map[string]interface{}{
				"current_variance": 0.45,
				"target_variance":  0.25,
				"filtering_method": "adaptive_zscore",
			},
		},
	}

	// Apply optimizations
	ctx := context.Background()
	changes, err := optimizer.applyOptimizations(ctx, opportunities)
	require.NoError(t, err)
	assert.Len(t, changes, 5)

	// Verify confidence weight adjustment
	confidenceChange := changes[0]
	assert.Equal(t, "confidence_weight", confidenceChange.Parameter)
	assert.Equal(t, 0.4, confidenceChange.OldValue)
	assert.InDelta(t, 0.44, confidenceChange.NewValue, 0.001) // 0.4 * 1.1
	assert.Equal(t, "threshold_adjustment", confidenceChange.ChangeType)
	assert.Equal(t, "low", confidenceChange.Impact)
	assert.Equal(t, 0.6, confidenceChange.Confidence)

	// Verify consistency weight adjustment
	consistencyChange := changes[1]
	assert.Equal(t, "consistency_weight", consistencyChange.Parameter)
	assert.Equal(t, 0.3, consistencyChange.OldValue)
	assert.InDelta(t, 0.345, consistencyChange.NewValue, 0.001) // 0.3 * 1.15
	assert.Equal(t, "threshold_adjustment", consistencyChange.ChangeType)
	assert.Equal(t, "low", consistencyChange.Impact)
	assert.Equal(t, 0.7, consistencyChange.Confidence)

	// Verify agreement threshold adjustment
	agreementChange := changes[2]
	assert.Equal(t, "required_agreement", agreementChange.Parameter)
	assert.Equal(t, 0.5, agreementChange.OldValue)
	assert.Equal(t, 0.45, agreementChange.NewValue) // 0.5 * 0.9
	assert.Equal(t, "threshold_adjustment", agreementChange.ChangeType)
	assert.Equal(t, "medium", agreementChange.Impact)
	assert.Equal(t, 0.8, agreementChange.Confidence)

	// Verify outlier threshold adjustment
	outlierThresholdChange := changes[3]
	assert.Equal(t, "outlier_threshold", outlierThresholdChange.Parameter)
	assert.Equal(t, 2.0, outlierThresholdChange.OldValue)
	assert.Equal(t, 1.6, outlierThresholdChange.NewValue) // 2.0 * 0.8
	assert.Equal(t, "threshold_adjustment", outlierThresholdChange.ChangeType)
	assert.Equal(t, "medium", outlierThresholdChange.Impact)
	assert.Equal(t, 0.7, outlierThresholdChange.Confidence)

	// Verify outlier filtering enhancement
	outlierFilteringChange := changes[4]
	assert.Equal(t, "enhanced_outlier_filtering", outlierFilteringChange.Parameter)
	assert.Equal(t, false, outlierFilteringChange.OldValue)
	assert.Equal(t, true, outlierFilteringChange.NewValue)
	assert.Equal(t, "feature_enablement", outlierFilteringChange.ChangeType)
	assert.Equal(t, "medium", outlierFilteringChange.Impact)
	assert.Equal(t, 0.8, outlierFilteringChange.Confidence)

	// Verify voting engine config was updated
	assert.InDelta(t, 0.44, votingEngine.config.ConfidenceWeight, 0.001)
	assert.InDelta(t, 0.345, votingEngine.config.ConsistencyWeight, 0.001)
	assert.Equal(t, 0.45, votingEngine.config.RequiredAgreement)
	assert.Equal(t, 1.6, votingEngine.config.OutlierThreshold)
	assert.Equal(t, true, votingEngine.config.EnableOutlierFiltering)
}

func TestVotingOptimizer_CalculateImprovement(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Create before and after metrics
	before := &VotingPerformanceMetrics{
		OverallAccuracy:       0.75,
		AverageConfidence:     0.70,
		VotingScore:           0.75,
		AverageProcessingTime: 100 * time.Millisecond,
	}

	after := &VotingPerformanceMetrics{
		OverallAccuracy:       0.80,                  // +0.05
		AverageConfidence:     0.75,                  // +0.05
		VotingScore:           0.80,                  // +0.05
		AverageProcessingTime: 80 * time.Millisecond, // -20ms
	}

	// Calculate improvement
	improvement := optimizer.calculateImprovement(before, after)

	// Verify improvement calculations
	assert.InDelta(t, 0.05, improvement.AccuracyImprovement, 0.001)
	assert.InDelta(t, 0.05, improvement.ConfidenceImprovement, 0.001)
	assert.InDelta(t, 0.05, improvement.VotingScoreImprovement, 0.001)
	assert.InDelta(t, 0.2, improvement.ProcessingTimeImprovement, 0.001) // 20ms/100ms = 0.2

	// Verify overall improvement (weighted average)
	expectedOverall := 0.05*0.4 + 0.05*0.3 + 0.05*0.2 + 0.2*0.1
	assert.InDelta(t, expectedOverall, improvement.OverallImprovement, 0.001)

	// Verify significance (should be significant since improvement > 0.02)
	assert.True(t, improvement.IsSignificant)
}

func TestVotingOptimizer_CalculateOptimizationConfidence(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Create test improvement and changes
	improvement := &VotingImprovementMetrics{
		OverallImprovement: 0.05, // Significant improvement
		IsSignificant:      true,
	}

	changes := []*VotingOptimizationChange{
		{
			Parameter:  "confidence_weight",
			ChangeType: "threshold_adjustment",
			Impact:     "low",
			Confidence: 0.6,
		},
		{
			Parameter:  "consistency_weight",
			ChangeType: "threshold_adjustment",
			Impact:     "low",
			Confidence: 0.7,
		},
		{
			Parameter:  "required_agreement",
			ChangeType: "threshold_adjustment",
			Impact:     "high",
			Confidence: 0.8,
		},
	}

	// Calculate confidence
	confidence := optimizer.calculateOptimizationConfidence(improvement, changes)

	// Base confidence: 0.5
	// Significant improvement: +0.2
	// Number of changes: +0.15 (3 * 0.05, capped at 0.2)
	// High impact changes: +0.1 (1 * 0.1)
	// Expected: 0.5 + 0.2 + 0.15 + 0.1 = 0.95
	expectedConfidence := 0.5 + 0.2 + 0.15 + 0.1
	assert.Equal(t, expectedConfidence, confidence)
}

func TestVotingOptimizer_GetOptimizationHistory(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Add some test optimization results
	result1 := &VotingOptimizationResult{
		ID:               "opt_1",
		OptimizationType: OptimizationTypeStrategy,
		Status:           OptimizationStatusCompleted,
		StartTime:        time.Now().Add(-1 * time.Hour),
	}

	result2 := &VotingOptimizationResult{
		ID:               "opt_2",
		OptimizationType: OptimizationTypeWeights,
		Status:           OptimizationStatusCompleted,
		StartTime:        time.Now().Add(-30 * time.Minute),
	}

	optimizer.mu.Lock()
	optimizer.optimizationHistory = append(optimizer.optimizationHistory, result1, result2)
	optimizer.mu.Unlock()

	// Get history
	history := optimizer.GetOptimizationHistory()
	assert.Len(t, history, 2)
	assert.Equal(t, "opt_1", history[0].ID)
	assert.Equal(t, "opt_2", history[1].ID)
}

func TestVotingOptimizer_GetPerformanceHistory(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Add some test performance metrics
	metrics1 := &VotingPerformanceMetrics{
		OverallAccuracy: 0.75,
		Timestamp:       time.Now().Add(-1 * time.Hour),
	}

	metrics2 := &VotingPerformanceMetrics{
		OverallAccuracy: 0.80,
		Timestamp:       time.Now().Add(-30 * time.Minute),
	}

	optimizer.mu.Lock()
	optimizer.performanceHistory = append(optimizer.performanceHistory, metrics1, metrics2)
	optimizer.mu.Unlock()

	// Get history
	history := optimizer.GetPerformanceHistory()
	assert.Len(t, history, 2)
	assert.Equal(t, 0.75, history[0].OverallAccuracy)
	assert.Equal(t, 0.80, history[1].OverallAccuracy)
}

func TestVotingOptimizer_GetActiveOptimizations(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Add some test active optimizations
	result1 := &VotingOptimizationResult{
		ID:               "opt_1",
		OptimizationType: OptimizationTypeStrategy,
		Status:           OptimizationStatusRunning,
		StartTime:        time.Now(),
	}

	result2 := &VotingOptimizationResult{
		ID:               "opt_2",
		OptimizationType: OptimizationTypeWeights,
		Status:           OptimizationStatusRunning,
		StartTime:        time.Now(),
	}

	optimizer.mu.Lock()
	optimizer.activeOptimizations["opt_1"] = result1
	optimizer.activeOptimizations["opt_2"] = result2
	optimizer.mu.Unlock()

	// Get active optimizations
	active := optimizer.GetActiveOptimizations()
	assert.Len(t, active, 2)

	// Verify both optimizations are present
	optIDs := make(map[string]bool)
	for _, opt := range active {
		optIDs[opt.ID] = true
	}
	assert.True(t, optIDs["opt_1"])
	assert.True(t, optIDs["opt_2"])
}

func TestVotingOptimizer_CheckOptimizationNeeded(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Test with insufficient samples
	optimizer.checkOptimizationNeeded()
	// Should not trigger optimization due to insufficient samples

	// Add sufficient samples with poor performance
	for i := 0; i < 150; i++ {
		metrics := &VotingPerformanceMetrics{
			OverallAccuracy:   0.65, // Below threshold
			AverageConfidence: 0.60, // Below threshold
			VotingScore:       0.55, // Below threshold
			Timestamp:         time.Now(),
		}
		optimizer.RecordVotingPerformance(metrics)
	}

	// Should trigger optimization check
	// Note: This test doesn't actually run the optimization due to goroutine
	// but we can verify the metrics are recorded
	history := optimizer.GetPerformanceHistory()
	assert.Len(t, history, 150)
}

func TestVotingOptimizer_OptimizeVotingAlgorithms(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Set up voting engine
	votingEngine := &VotingEngine{
		config: &VotingConfig{
			Strategy:               VotingStrategyWeightedAverage,
			MinVoters:              2,
			RequiredAgreement:      0.5,
			ConfidenceWeight:       0.4,
			ConsistencyWeight:      0.3,
			DiversityWeight:        0.3,
			EnableTieBreaking:      true,
			EnableOutlierFiltering: false,
			OutlierThreshold:       2.0,
		},
		logger: logger,
	}
	optimizer.SetVotingComponents(votingEngine, nil, nil)

	// Add performance metrics
	metrics := &VotingPerformanceMetrics{
		OverallAccuracy:    0.65, // Below threshold
		AverageConfidence:  0.60, // Below threshold
		VotingScore:        0.55, // Below threshold
		AgreementScore:     0.40, // Below threshold
		ConsistencyScore:   0.50, // Below threshold
		ConfidenceVariance: 0.35, // Above threshold
		StrategyPerformance: map[string]*StrategyPerformanceMetrics{
			"keyword": {
				StrategyName:      "keyword",
				AverageAccuracy:   0.60, // Below threshold
				AverageConfidence: 0.65,
				PerformanceScore:  0.65,
			},
		},
		Timestamp: time.Now(),
	}
	optimizer.RecordVotingPerformance(metrics)

	// Run optimization
	ctx := context.Background()
	result, err := optimizer.OptimizeVotingAlgorithms(ctx)
	require.NoError(t, err)

	// Verify optimization result
	assert.NotNil(t, result)
	assert.Equal(t, OptimizationTypeComprehensive, result.OptimizationType)
	assert.Equal(t, OptimizationStatusCompleted, result.Status)
	assert.NotNil(t, result.BeforeMetrics)
	assert.NotNil(t, result.AppliedChanges)
	assert.NotEmpty(t, result.AppliedChanges)

	// Verify optimization history was updated
	history := optimizer.GetOptimizationHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, result.ID, history[0].ID)
}

func TestVotingOptimizer_ApplyStrategyImprovement(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	opportunity := &OptimizationOpportunity{
		Type:            "strategy_improvement",
		Description:     "Improve performance of strategy: keyword",
		PotentialImpact: 0.1,
		Confidence:      0.8,
		Effort:          "medium",
		Priority:        2,
		Parameters: map[string]interface{}{
			"strategy_name":    "keyword",
			"current_accuracy": 0.60,
			"target_accuracy":  0.80,
		},
	}

	change, err := optimizer.applyStrategyImprovement(opportunity)
	require.NoError(t, err)

	assert.Equal(t, "strategy_weight_keyword", change.Parameter)
	assert.Equal(t, 1.0, change.OldValue)
	assert.Equal(t, 0.8, change.NewValue)
	assert.Equal(t, "weight_adjustment", change.ChangeType)
	assert.Equal(t, "medium", change.Impact)
	assert.Equal(t, 0.8, change.Confidence)
	assert.Contains(t, change.Reason, "Reduce weight for underperforming strategy: keyword")
}

func TestVotingOptimizer_ApplyWeightRebalancing(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	opportunity := &OptimizationOpportunity{
		Type:            "weight_rebalancing",
		Description:     "Rebalance strategy weights based on performance",
		PotentialImpact: 0.05,
		Confidence:      0.7,
		Effort:          "low",
		Priority:        3,
		Parameters: map[string]interface{}{
			"rebalancing_method": "performance_based",
		},
	}

	change, err := optimizer.applyWeightRebalancing(opportunity)
	require.NoError(t, err)

	assert.Equal(t, "strategy_weights", change.Parameter)
	assert.Equal(t, "uniform", change.OldValue)
	assert.Equal(t, "performance_based", change.NewValue)
	assert.Equal(t, "rebalancing", change.ChangeType)
	assert.Equal(t, "low", change.Impact)
	assert.Equal(t, 0.7, change.Confidence)
	assert.Equal(t, "Rebalance strategy weights based on performance metrics", change.Reason)
}

func TestVotingOptimizer_ApplyOptimization_UnknownType(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	opportunity := &OptimizationOpportunity{
		Type: "unknown_type",
	}

	ctx := context.Background()
	_, err := optimizer.applyOptimization(ctx, opportunity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown optimization type")
}

func TestVotingOptimizer_ApplyOptimization_MissingVotingEngine(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	opportunity := &OptimizationOpportunity{
		Type: "confidence_weight_adjustment",
	}

	ctx := context.Background()
	_, err := optimizer.applyOptimization(ctx, opportunity)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "voting engine not available")
}

func TestVotingOptimizer_CalculateImprovement_ZeroProcessingTime(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Test with zero processing time to avoid division by zero
	before := &VotingPerformanceMetrics{
		OverallAccuracy:       0.75,
		AverageConfidence:     0.70,
		VotingScore:           0.75,
		AverageProcessingTime: 0, // Zero processing time
	}

	after := &VotingPerformanceMetrics{
		OverallAccuracy:       0.80,
		AverageConfidence:     0.75,
		VotingScore:           0.80,
		AverageProcessingTime: 50 * time.Millisecond,
	}

	// Should handle zero processing time gracefully
	improvement := optimizer.calculateImprovement(before, after)

	// Verify other improvements are calculated correctly
	assert.InDelta(t, 0.05, improvement.AccuracyImprovement, 0.001)
	assert.InDelta(t, 0.05, improvement.ConfidenceImprovement, 0.001)
	assert.InDelta(t, 0.05, improvement.VotingScoreImprovement, 0.001)

	// Processing time improvement should be 0 when before time is 0
	assert.InDelta(t, 0.0, improvement.ProcessingTimeImprovement, 0.001)
}

func TestVotingOptimizer_PerformanceHistoryLimit(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Add more than 1000 metrics to test history limit
	for i := 0; i < 1100; i++ {
		metrics := &VotingPerformanceMetrics{
			OverallAccuracy: float64(i) / 1000.0,
			Timestamp:       time.Now(),
		}
		optimizer.RecordVotingPerformance(metrics)
	}

	// Should keep only the last 1000 metrics
	history := optimizer.GetPerformanceHistory()
	assert.Len(t, history, 1000)

	// Should have the most recent metrics (highest accuracy values)
	// The last metric should be the highest value (1.099 from the loop)
	assert.InDelta(t, 1.099, history[999].OverallAccuracy, 0.001) // Last metric should be 1.099
}

func TestVotingOptimizer_OptimizationValidation(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Test validation engine
	validationEngine := optimizer.validationEngine
	assert.NotNil(t, validationEngine)

	// Create test optimization result
	result := &VotingOptimizationResult{
		ID:               "test_opt",
		OptimizationType: OptimizationTypeStrategy,
		Status:           OptimizationStatusCompleted,
		StartTime:        time.Now(),
	}

	// Validate optimization
	ctx := context.Background()
	validationResult, err := validationEngine.ValidateOptimization(ctx, result)
	require.NoError(t, err)

	assert.True(t, validationResult.IsValid)
	assert.Equal(t, 0.8, validationResult.ValidationScore)
	assert.Empty(t, validationResult.Issues)
	assert.Empty(t, validationResult.Warnings)
}

func TestVotingOptimizer_RollbackManager(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Test rollback manager
	rollbackManager := optimizer.rollbackManager
	assert.NotNil(t, rollbackManager)

	// Create test optimization result
	result := &VotingOptimizationResult{
		ID:               "test_opt",
		OptimizationType: OptimizationTypeStrategy,
		Status:           OptimizationStatusCompleted,
		StartTime:        time.Now(),
	}

	// Rollback optimization
	err := rollbackManager.RollbackOptimization(result)
	assert.NoError(t, err)
}

func TestVotingOptimizer_LearningModel(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Test learning model
	learningModel := optimizer.learningModel
	assert.NotNil(t, learningModel)
	assert.Equal(t, optimizer.config, learningModel.config)
	assert.Equal(t, logger, learningModel.logger)
}

func TestVotingOptimizer_AdaptationEngine(t *testing.T) {
	logger := zap.NewNop()
	optimizer := NewVotingOptimizer(nil, logger)

	// Test adaptation engine
	adaptationEngine := optimizer.adaptationEngine
	assert.NotNil(t, adaptationEngine)
	assert.Equal(t, optimizer.config, adaptationEngine.config)
	assert.Equal(t, logger, adaptationEngine.logger)
}
