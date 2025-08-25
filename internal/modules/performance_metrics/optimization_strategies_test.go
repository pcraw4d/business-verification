package performance_metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewOptimizationStrategies(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	config := DefaultOptimizationConfig()

	strategies := NewOptimizationStrategies(logger, detector, config)

	assert.NotNil(t, strategies)
	assert.Equal(t, logger, strategies.logger)
	assert.Equal(t, detector, strategies.detector)
	assert.Equal(t, config, strategies.config)
	assert.NotNil(t, strategies.strategies)
	assert.NotNil(t, strategies.plans)
}

func TestDefaultOptimizationConfig(t *testing.T) {
	config := DefaultOptimizationConfig()

	assert.True(t, config.EnableOptimization)
	assert.Equal(t, 10*time.Minute, config.AnalysisInterval)
	assert.Equal(t, 7*24*time.Hour, config.RetentionPeriod)
	assert.Equal(t, 50, config.MaxStrategies)
	assert.Equal(t, 1.5, config.ROIThreshold)
	assert.Equal(t, "medium", config.RiskTolerance)
	assert.False(t, config.AutoApply)
}

func TestOptimizationStrategies_GenerateOptimizationPlan_NoBottlenecks(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	ctx := context.Background()
	plan, err := strategies.GenerateOptimizationPlan(ctx)

	require.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "No bottlenecks detected - no optimization strategies needed", plan.Summary)
	assert.Equal(t, 0, len(plan.Strategies))
	assert.Equal(t, 0.0, plan.TotalExpectedImpact)
	assert.Equal(t, 0.0, plan.TotalROI)
}

func TestOptimizationStrategies_GenerateOptimizationPlan_WithBottlenecks(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	ctx := context.Background()

	// Create some bottlenecks manually
	bottleneck1 := &Bottleneck{
		ID:        "test1",
		Type:      BottleneckTypeAlgorithm,
		Severity:  BottleneckSeverityHigh,
		Operation: "classification",
		Metrics: map[string]float64{
			"avg_response_time": 3000.0,
		},
		DetectedAt: time.Now(),
	}
	bottleneck2 := &Bottleneck{
		ID:       "test2",
		Type:     BottleneckTypeCPU,
		Severity: BottleneckSeverityMedium,
		Metrics: map[string]float64{
			"avg_cpu_usage": 85.0,
		},
		DetectedAt: time.Now(),
	}

	detector.bottlenecks["test1"] = bottleneck1
	detector.bottlenecks["test2"] = bottleneck2

	plan, err := strategies.GenerateOptimizationPlan(ctx)

	require.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Greater(t, len(plan.Strategies), 0)
	assert.Greater(t, plan.TotalExpectedImpact, 0.0)
	assert.Greater(t, plan.TotalROI, 0.0)
	assert.NotEmpty(t, plan.Summary)
	assert.NotEmpty(t, plan.SuccessMetrics)
	assert.NotEmpty(t, plan.RiskAssessment)
}

func TestOptimizationStrategies_GenerateAlgorithmStrategies(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	bottleneck := &Bottleneck{
		ID:        "algo_bottleneck",
		Type:      BottleneckTypeAlgorithm,
		Severity:  BottleneckSeverityHigh,
		Operation: "classification",
		Metrics: map[string]float64{
			"avg_response_time": 2500.0,
		},
	}

	algorithmStrategies := strategies.generateAlgorithmStrategies(bottleneck)

	assert.Equal(t, 2, len(algorithmStrategies)) // Algorithm optimization + caching

	// Check algorithm optimization strategy
	var algoStrategy *OptimizationStrategy
	for _, s := range algorithmStrategies {
		if s.Type == OptimizationTypeAlgorithm {
			algoStrategy = s
			break
		}
	}

	require.NotNil(t, algoStrategy)
	assert.Equal(t, OptimizationTypeAlgorithm, algoStrategy.Type)
	assert.Equal(t, OptimizationPriorityHigh, algoStrategy.Priority)
	assert.Contains(t, algoStrategy.Name, "Optimize Classification Algorithm")
	assert.Equal(t, 0.4, algoStrategy.ExpectedImpact)
	assert.Equal(t, 0.85, algoStrategy.Confidence)
	assert.Equal(t, "medium", algoStrategy.Effort)
	assert.Equal(t, "low", algoStrategy.Risk)
	assert.Equal(t, 2.5, algoStrategy.ROI)
	assert.NotEmpty(t, algoStrategy.Implementation)
	assert.NotEmpty(t, algoStrategy.Prerequisites)
	assert.Equal(t, "proposed", algoStrategy.Status)

	// Check caching strategy
	var cacheStrategy *OptimizationStrategy
	for _, s := range algorithmStrategies {
		if s.Type == OptimizationTypeCaching {
			cacheStrategy = s
			break
		}
	}

	require.NotNil(t, cacheStrategy)
	assert.Equal(t, OptimizationTypeCaching, cacheStrategy.Type)
	assert.Contains(t, cacheStrategy.Name, "Implement Intelligent Caching")
	assert.Equal(t, 0.3, cacheStrategy.ExpectedImpact)
	assert.Equal(t, 0.90, cacheStrategy.Confidence)
	assert.Equal(t, "low", cacheStrategy.Effort)
	assert.Equal(t, "low", cacheStrategy.Risk)
	assert.Equal(t, 3.0, cacheStrategy.ROI)
}

func TestOptimizationStrategies_GenerateCPUStrategies(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	bottleneck := &Bottleneck{
		ID:       "cpu_bottleneck",
		Type:     BottleneckTypeCPU,
		Severity: BottleneckSeverityCritical,
		Metrics: map[string]float64{
			"avg_cpu_usage": 90.0,
		},
	}

	cpuStrategies := strategies.generateCPUStrategies(bottleneck)

	assert.Equal(t, 2, len(cpuStrategies)) // Concurrency + code optimization

	// Check concurrency strategy
	var concurrencyStrategy *OptimizationStrategy
	for _, s := range cpuStrategies {
		if s.Type == OptimizationTypeConcurrency {
			concurrencyStrategy = s
			break
		}
	}

	require.NotNil(t, concurrencyStrategy)
	assert.Equal(t, OptimizationTypeConcurrency, concurrencyStrategy.Type)
	assert.Equal(t, OptimizationPriorityCritical, concurrencyStrategy.Priority)
	assert.Contains(t, concurrencyStrategy.Name, "Implement Parallel Processing")
	assert.Equal(t, 0.35, concurrencyStrategy.ExpectedImpact)
	assert.Equal(t, 0.80, concurrencyStrategy.Confidence)
	assert.Equal(t, "medium", concurrencyStrategy.Effort)
	assert.Equal(t, "medium", concurrencyStrategy.Risk)
	assert.Equal(t, 2.2, concurrencyStrategy.ROI)

	// Check code optimization strategy
	var codeStrategy *OptimizationStrategy
	for _, s := range cpuStrategies {
		if s.Type == OptimizationTypeCode {
			codeStrategy = s
			break
		}
	}

	require.NotNil(t, codeStrategy)
	assert.Equal(t, OptimizationTypeCode, codeStrategy.Type)
	assert.Contains(t, codeStrategy.Name, "Optimize CPU-Intensive Code")
	assert.Equal(t, 0.25, codeStrategy.ExpectedImpact)
	assert.Equal(t, 0.75, codeStrategy.Confidence)
	assert.Equal(t, "high", codeStrategy.Effort)
	assert.Equal(t, "low", codeStrategy.Risk)
	assert.Equal(t, 1.8, codeStrategy.ROI)
}

func TestOptimizationStrategies_GenerateMemoryStrategies(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	bottleneck := &Bottleneck{
		ID:       "memory_bottleneck",
		Type:     BottleneckTypeMemory,
		Severity: BottleneckSeverityHigh,
		Metrics: map[string]float64{
			"avg_memory_usage": 88.0,
		},
	}

	memoryStrategies := strategies.generateMemoryStrategies(bottleneck)

	assert.Equal(t, 1, len(memoryStrategies))

	strategy := memoryStrategies[0]
	assert.Equal(t, OptimizationTypeResource, strategy.Type)
	assert.Equal(t, OptimizationPriorityHigh, strategy.Priority)
	assert.Contains(t, strategy.Name, "Optimize Memory Usage")
	assert.Equal(t, 0.3, strategy.ExpectedImpact)
	assert.Equal(t, 0.85, strategy.Confidence)
	assert.Equal(t, "medium", strategy.Effort)
	assert.Equal(t, "low", strategy.Risk)
	assert.Equal(t, 2.0, strategy.ROI)
	assert.NotEmpty(t, strategy.Implementation)
	assert.NotEmpty(t, strategy.Prerequisites)
}

func TestOptimizationStrategies_GenerateResourceStrategies(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	bottleneck := &Bottleneck{
		ID:        "resource_bottleneck",
		Type:      BottleneckTypeResource,
		Severity:  BottleneckSeverityMedium,
		Operation: "enrichment",
	}

	resourceStrategies := strategies.generateResourceStrategies(bottleneck)

	assert.Equal(t, 1, len(resourceStrategies))

	strategy := resourceStrategies[0]
	assert.Equal(t, OptimizationTypeResource, strategy.Type)
	assert.Equal(t, OptimizationPriorityMedium, strategy.Priority)
	assert.Contains(t, strategy.Name, "Scale Resources")
	assert.Equal(t, 0.5, strategy.ExpectedImpact)
	assert.Equal(t, 0.95, strategy.Confidence)
	assert.Equal(t, "low", strategy.Effort)
	assert.Equal(t, "low", strategy.Risk)
	assert.Equal(t, 1.5, strategy.ROI)
}

func TestOptimizationStrategies_CalculatePriority(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	assert.Equal(t, OptimizationPriorityCritical, strategies.calculatePriority(BottleneckSeverityCritical))
	assert.Equal(t, OptimizationPriorityHigh, strategies.calculatePriority(BottleneckSeverityHigh))
	assert.Equal(t, OptimizationPriorityMedium, strategies.calculatePriority(BottleneckSeverityMedium))
	assert.Equal(t, OptimizationPriorityLow, strategies.calculatePriority(BottleneckSeverityLow))
	assert.Equal(t, OptimizationPriorityMedium, strategies.calculatePriority(BottleneckSeverityInfo))
}

func TestOptimizationStrategies_SortStrategies(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	// Create strategies with different priorities and ROIs
	strategy1 := &OptimizationStrategy{
		Priority: OptimizationPriorityMedium,
		ROI:      2.0,
	}
	strategy2 := &OptimizationStrategy{
		Priority: OptimizationPriorityHigh,
		ROI:      1.5,
	}
	strategy3 := &OptimizationStrategy{
		Priority: OptimizationPriorityHigh,
		ROI:      2.5,
	}
	strategy4 := &OptimizationStrategy{
		Priority: OptimizationPriorityLow,
		ROI:      3.0,
	}

	strategyList := []*OptimizationStrategy{strategy1, strategy2, strategy3, strategy4}
	strategies.sortStrategies(strategyList)

	// Should be sorted by priority first, then by ROI
	// Expected order: strategy3 (high priority, high ROI), strategy2 (high priority, low ROI), strategy1 (medium priority), strategy4 (low priority)
	assert.Equal(t, strategy3, strategyList[0])
	assert.Equal(t, strategy2, strategyList[1])
	assert.Equal(t, strategy1, strategyList[2])
	assert.Equal(t, strategy4, strategyList[3])
}

func TestOptimizationStrategies_GetPriorityWeight(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	assert.Equal(t, 4, strategies.getPriorityWeight(OptimizationPriorityCritical))
	assert.Equal(t, 3, strategies.getPriorityWeight(OptimizationPriorityHigh))
	assert.Equal(t, 2, strategies.getPriorityWeight(OptimizationPriorityMedium))
	assert.Equal(t, 1, strategies.getPriorityWeight(OptimizationPriorityLow))
	assert.Equal(t, 2, strategies.getPriorityWeight("unknown"))
}

func TestOptimizationStrategies_CalculateTimeline(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	// Create strategies with different effort levels
	strategiesList := []*OptimizationStrategy{
		{Effort: "low"},
		{Effort: "medium"},
		{Effort: "high"},
		{Effort: "medium"},
	}

	timeline := strategies.calculateTimeline(strategiesList)

	// Expected: 1 + 3 + 7 + 3 = 14 days + 4*2 = 8 days buffer = 22 days
	expectedDays := 22
	assert.Equal(t, time.Duration(expectedDays)*24*time.Hour, timeline)
}

func TestOptimizationStrategies_AssessOverallRisk(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	// Test with high risk
	strategiesList := []*OptimizationStrategy{
		{Risk: "high"},
		{Risk: "low"},
	}
	risk := strategies.assessOverallRisk(strategiesList)
	assert.Equal(t, "high", risk)

	// Test with medium risk (more than 2 medium risk strategies)
	strategiesList = []*OptimizationStrategy{
		{Risk: "medium"},
		{Risk: "medium"},
		{Risk: "medium"},
		{Risk: "low"},
	}
	risk = strategies.assessOverallRisk(strategiesList)
	assert.Equal(t, "medium", risk)

	// Test with low risk
	strategiesList = []*OptimizationStrategy{
		{Risk: "low"},
		{Risk: "low"},
	}
	risk = strategies.assessOverallRisk(strategiesList)
	assert.Equal(t, "low", risk)
}

func TestOptimizationStrategies_DefineSuccessMetrics(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	strategiesList := []*OptimizationStrategy{
		{Type: OptimizationTypeCaching},
		{Type: OptimizationTypeConcurrency},
		{Type: OptimizationTypeResource},
	}

	metricsList := strategies.defineSuccessMetrics(strategiesList)

	// Should include base metrics plus strategy-specific ones
	assert.Contains(t, metricsList, "Overall response time reduction")
	assert.Contains(t, metricsList, "Error rate reduction")
	assert.Contains(t, metricsList, "Throughput improvement")
	assert.Contains(t, metricsList, "Resource utilization optimization")
	assert.Contains(t, metricsList, "Cache hit rate improvement")
	assert.Contains(t, metricsList, "CPU utilization optimization")
	assert.Contains(t, metricsList, "Memory usage reduction")
}

func TestOptimizationStrategies_GeneratePlanSummary(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	// Test with empty strategies
	strategiesList := []*OptimizationStrategy{}
	summary := strategies.generatePlanSummary(strategiesList)
	assert.Equal(t, "No optimization strategies generated", summary)

	// Test with strategies
	strategiesList = []*OptimizationStrategy{
		{Priority: OptimizationPriorityCritical, ExpectedImpact: 0.3},
		{Priority: OptimizationPriorityHigh, ExpectedImpact: 0.2},
		{Priority: OptimizationPriorityMedium, ExpectedImpact: 0.1},
	}

	summary = strategies.generatePlanSummary(strategiesList)
	assert.Contains(t, summary, "Optimization plan with 3 strategies")
	assert.Contains(t, summary, "1 critical, 1 high priority")
	assert.Contains(t, summary, "60.0% improvement")
}

func TestOptimizationStrategies_GetOptimizationPlans(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	// Add some plans manually
	plan1 := &OptimizationPlan{
		PlanID:    "plan1",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	plan2 := &OptimizationPlan{
		PlanID:    "plan2",
		CreatedAt: time.Now(),
	}

	strategies.plans["plan1"] = plan1
	strategies.plans["plan2"] = plan2

	plans := strategies.GetOptimizationPlans()

	assert.Equal(t, 2, len(plans))
	// Should be sorted by creation time (newest first)
	assert.Equal(t, "plan2", plans[0].PlanID)
	assert.Equal(t, "plan1", plans[1].PlanID)
}

func TestOptimizationStrategies_GetStrategies(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	// Add some strategies manually
	strategy1 := &OptimizationStrategy{
		ID:        "strategy1",
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}
	strategy2 := &OptimizationStrategy{
		ID:        "strategy2",
		CreatedAt: time.Now(),
	}

	strategies.strategies["strategy1"] = strategy1
	strategies.strategies["strategy2"] = strategy2

	strategyList := strategies.GetStrategies()

	assert.Equal(t, 2, len(strategyList))
	// Should be sorted by creation time (newest first)
	assert.Equal(t, "strategy2", strategyList[0].ID)
	assert.Equal(t, "strategy1", strategyList[1].ID)
}

func TestOptimizationStrategies_GetStrategiesByType(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	// Add strategies with different types
	strategy1 := &OptimizationStrategy{
		ID:   "cache1",
		Type: OptimizationTypeCaching,
	}
	strategy2 := &OptimizationStrategy{
		ID:   "algo1",
		Type: OptimizationTypeAlgorithm,
	}
	strategy3 := &OptimizationStrategy{
		ID:   "cache2",
		Type: OptimizationTypeCaching,
	}

	strategies.strategies["cache1"] = strategy1
	strategies.strategies["algo1"] = strategy2
	strategies.strategies["cache2"] = strategy3

	cacheStrategies := strategies.GetStrategiesByType(OptimizationTypeCaching)
	assert.Equal(t, 2, len(cacheStrategies))

	algoStrategies := strategies.GetStrategiesByType(OptimizationTypeAlgorithm)
	assert.Equal(t, 1, len(algoStrategies))

	resourceStrategies := strategies.GetStrategiesByType(OptimizationTypeResource)
	assert.Equal(t, 0, len(resourceStrategies))
}

func TestOptimizationStrategies_GetStrategiesByPriority(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	// Add strategies with different priorities
	strategy1 := &OptimizationStrategy{
		ID:       "critical1",
		Priority: OptimizationPriorityCritical,
	}
	strategy2 := &OptimizationStrategy{
		ID:       "high1",
		Priority: OptimizationPriorityHigh,
	}
	strategy3 := &OptimizationStrategy{
		ID:       "critical2",
		Priority: OptimizationPriorityCritical,
	}

	strategies.strategies["critical1"] = strategy1
	strategies.strategies["high1"] = strategy2
	strategies.strategies["critical2"] = strategy3

	criticalStrategies := strategies.GetStrategiesByPriority(OptimizationPriorityCritical)
	assert.Equal(t, 2, len(criticalStrategies))

	highStrategies := strategies.GetStrategiesByPriority(OptimizationPriorityHigh)
	assert.Equal(t, 1, len(highStrategies))

	mediumStrategies := strategies.GetStrategiesByPriority(OptimizationPriorityMedium)
	assert.Equal(t, 0, len(mediumStrategies))
}

func TestOptimizationStrategies_ApplyStrategy(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	// Add a strategy
	strategy := &OptimizationStrategy{
		ID:             "test_strategy",
		Name:           "Test Strategy",
		Status:         "proposed",
		ExpectedImpact: 0.3,
		Metrics: map[string]float64{
			"test_metric": 100.0,
		},
	}
	strategies.strategies["test_strategy"] = strategy

	// Apply the strategy
	result, err := strategies.ApplyStrategy("test_strategy")
	require.NoError(t, err)

	assert.NotNil(t, result)
	assert.Equal(t, "test_strategy", result.StrategyID)
	assert.True(t, result.Success)
	assert.Equal(t, 0.27, result.ActualImpact) // 90% of expected
	assert.Equal(t, 0.3, result.ExpectedImpact)
	assert.Equal(t, 2*time.Hour, result.Duration)
	assert.NotEmpty(t, result.Recommendations)

	// Check that strategy status was updated
	assert.Equal(t, "applied", strategy.Status)
	assert.NotNil(t, strategy.AppliedAt)
	assert.NotNil(t, strategy.Results)

	// Try to apply non-existent strategy
	_, err = strategies.ApplyStrategy("non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "strategy not found")

	// Try to apply already applied strategy
	_, err = strategies.ApplyStrategy("test_strategy")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "strategy is not in proposed status")
}

func TestOptimizationStrategies_IntegrationTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)
	strategies := NewOptimizationStrategies(logger, detector, nil)

	ctx := context.Background()

	// Record metrics to create bottlenecks
	metrics.RecordResponseTime(ctx, "classification", 4000*time.Millisecond, nil)
	metrics.RecordGauge(ctx, "cpu_usage", 88.0, "percentage", nil)
	metrics.RecordGauge(ctx, "memory_usage", 92.0, "percentage", nil)

	// Perform bottleneck analysis
	_, err := detector.AnalyzeBottlenecks(ctx)
	require.NoError(t, err)

	// Generate optimization plan
	plan, err := strategies.GenerateOptimizationPlan(ctx)
	require.NoError(t, err)

	// Verify plan was generated
	assert.NotNil(t, plan)
	assert.Greater(t, len(plan.Strategies), 0)
	assert.Greater(t, plan.TotalExpectedImpact, 0.0)
	assert.Greater(t, plan.TotalROI, 0.0)
	assert.NotEmpty(t, plan.Summary)
	assert.NotEmpty(t, plan.SuccessMetrics)

	// Verify strategies are stored
	storedStrategies := strategies.GetStrategies()
	assert.Equal(t, len(plan.Strategies), len(storedStrategies))

	// Verify plans are stored
	storedPlans := strategies.GetOptimizationPlans()
	assert.Equal(t, 1, len(storedPlans))
	assert.Equal(t, plan.PlanID, storedPlans[0].PlanID)

	// Test strategy application
	if len(storedStrategies) > 0 {
		strategyID := storedStrategies[0].ID
		result, err := strategies.ApplyStrategy(strategyID)
		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
	}
}
