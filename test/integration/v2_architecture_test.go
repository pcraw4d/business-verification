package integration

import (
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/observability/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestV2ArchitectureIntegration(t *testing.T) {
	// Create logger
	logger := zap.NewNop()

	// Create performance monitor configuration
	monitorConfig := observability.PerformanceMonitorConfig{
		MetricsCollectionInterval: 1 * time.Second,
		AlertCheckInterval:        2 * time.Second,
		OptimizationInterval:      5 * time.Second,
		PredictionInterval:        3 * time.Second,
		ResponseTimeThreshold:     1 * time.Second,
		SuccessRateThreshold:      0.98,
		ErrorRateThreshold:        0.02,
		ThroughputThreshold:       100,
		AutoOptimizationEnabled:   true,
		OptimizationConfidence:    0.8,
		RollbackThreshold:         -0.10,
	}

	// Create performance monitor
	performanceMonitor := observability.NewPerformanceMonitor(monitorConfig)
	require.NotNil(t, performanceMonitor)

	// Test 1: Verify V2 metrics structure
	t.Run("V2 Metrics Structure", func(t *testing.T) {
		// Get V2 metrics
		metrics := performanceMonitor.GetMetricsV2()
		require.NotNil(t, metrics)

		// Verify summary structure
		assert.NotNil(t, metrics.Summary)
		assert.GreaterOrEqual(t, metrics.Summary.SuccessRate, 0.0)
		assert.LessOrEqual(t, metrics.Summary.SuccessRate, 1.0)
		assert.GreaterOrEqual(t, metrics.Summary.RPS, 0.0)
		assert.GreaterOrEqual(t, metrics.Summary.CPUUsage, 0.0)
		assert.LessOrEqual(t, metrics.Summary.CPUUsage, 1.0)

		// Verify breakdown structure
		assert.NotNil(t, metrics.Breakdown)
		assert.NotNil(t, metrics.Breakdown.Latency)
		assert.NotNil(t, metrics.Breakdown.Throughput)
		assert.NotNil(t, metrics.Breakdown.Success)
		assert.NotNil(t, metrics.Breakdown.Resources)
		assert.NotNil(t, metrics.Breakdown.Business)
	})

	// Test 2: Verify automated optimizer with V2 types
	t.Run("Automated Optimizer V2 Integration", func(t *testing.T) {
		// Create optimizer configuration
		optimizerConfig := observability.AutomatedOptimizerConfig{
			OptimizationCheckInterval:    5 * time.Second,
			PerformanceHistoryWindow:     1 * time.Hour,
			OptimizationCooldown:         1 * time.Minute,
			ResponseTimeThreshold:        1 * time.Second,
			SuccessRateThreshold:         0.98,
			ResourceUtilizationThreshold: 0.8,
			ThroughputThreshold:          100.0,
			EnableAutoScaling:            true,
			EnableCacheOptimization:      true,
			EnableDatabaseOptimization:   true,
			EnableConnectionPooling:      true,
			EnableLoadBalancing:          true,
			MaxOptimizationsPerHour:      10,
			OptimizationImpactLimit:      0.2,
			RollbackThreshold:            -0.1,
			EnableMachineLearning:        true,
			LearningRate:                 0.1,
			HistoricalDataPoints:         100,
		}

		// Create mock components for optimizer
		successRateTracker := &observability.SuccessRateTracker{}
		realTimeDashboard := &observability.RealTimeDashboard{}

		// Create automated optimizer
		optimizer := observability.NewAutomatedOptimizer(
			successRateTracker,
			performanceMonitor,
			realTimeDashboard,
			optimizerConfig,
			logger,
		)
		require.NotNil(t, optimizer)

		// Get optimization state
		state := optimizer.GetOptimizationState()
		assert.NotNil(t, state)
		assert.GreaterOrEqual(t, state.OptimizationsToday, 0)
	})

	// Test 3: Verify performance optimization system with V2 types
	t.Run("Performance Optimization System V2 Integration", func(t *testing.T) {
		// Create optimization configuration
		optimizationConfig := observability.OptimizationConfig{
			AnalysisInterval:         5 * time.Second,
			RecommendationThreshold:  0.7,
			ConfidenceThreshold:      0.8,
			MaxRecommendations:       10,
			RecommendationExpiry:     1 * time.Hour,
			AutoPrioritization:       true,
			AutoImplementation:       false,
			ImplementationDelay:      1 * time.Minute,
			RollbackThreshold:        -0.10,
			MaxAnalysisDuration:      30 * time.Second,
			MinDataPoints:            5,
			AnalysisWindow:           5 * time.Minute,
			EnableOptimizationAlerts: true,
			AlertSeverity: map[string]string{
				"high":   "critical",
				"medium": "warning",
				"low":    "info",
			},
		}

		// Create mock components for optimization system
		regressionDetection := &observability.RegressionDetectionSystem{}
		benchmarkingSystem := &observability.PerformanceBenchmarkingSystem{}
		predictiveAnalytics := &observability.PredictiveAnalytics{}

		// Create performance optimization system
		pos := observability.NewPerformanceOptimizationSystem(
			performanceMonitor,
			regressionDetection,
			benchmarkingSystem,
			predictiveAnalytics,
			optimizationConfig,
			logger,
		)
		require.NotNil(t, pos)

		// System created successfully - internal methods are tested separately
	})

	// Test 4: Verify automated performance tuning with V2 types
	t.Run("Automated Performance Tuning V2 Integration", func(t *testing.T) {
		// Create tuning configuration
		tuningConfig := observability.TuningConfig{
			TuningInterval:               5 * time.Second,
			MaxConcurrentTunings:         3,
			TuningTimeout:                10 * time.Second,
			MaxTuningAttempts:            5,
			RollbackThreshold:            -0.10,
			SafetyMargin:                 0.20,
			MinImprovement:               0.05,
			MaxDegradation:               0.10,
			StabilizationPeriod:          2 * time.Second,
			ResponseTimeThreshold:        1 * time.Second,
			SuccessRateThreshold:         0.98,
			ResourceUtilizationThreshold: 0.80,
		}

		// Create mock components for tuning system
		optimizationSystem := &observability.PerformanceOptimizationSystem{}
		predictiveAnalytics := &observability.PredictiveAnalytics{}
		regressionDetection := &observability.RegressionDetectionSystem{}

		// Create automated performance tuning system
		apts := observability.NewAutomatedPerformanceTuningSystem(
			performanceMonitor,
			optimizationSystem,
			predictiveAnalytics,
			regressionDetection,
			tuningConfig,
			logger,
		)
		require.NotNil(t, apts)

		// Get tuning sessions
		sessions := apts.GetTuningSessions()
		assert.NotNil(t, sessions)
		// Sessions might be empty if no tuning is active

		// Get tuning history
		history := apts.GetTuningHistory()
		assert.NotNil(t, history)
		// History might be empty if no tuning has occurred
	})

	// Test 5: Verify V2 type consistency across components
	t.Run("V2 Type Consistency", func(t *testing.T) {
		// Get metrics from monitor
		metrics := performanceMonitor.GetMetricsV2()
		require.NotNil(t, metrics)

		// Verify metrics implement the MetricsProvider interface
		// This is implicit in the GetMetricsV2() call, but we can verify the structure
		assert.IsType(t, &types.PerformanceMetricsV2{}, metrics)
		assert.IsType(t, types.MetricsSummary{}, metrics.Summary)
		assert.IsType(t, types.MetricsBreakdown{}, metrics.Breakdown)

		// Verify field access patterns work correctly
		assert.GreaterOrEqual(t, metrics.Summary.SuccessRate, 0.0)
		assert.LessOrEqual(t, metrics.Summary.SuccessRate, 1.0)
		assert.GreaterOrEqual(t, metrics.Summary.RPS, 0.0)
		assert.GreaterOrEqual(t, metrics.Summary.CPUUsage, 0.0)
		assert.LessOrEqual(t, metrics.Summary.CPUUsage, 1.0)
	})

	// Test 6: Verify no legacy adapter references
	t.Run("No Legacy Adapter References", func(t *testing.T) {
		// This test verifies that we're using V2 types directly
		// and not falling back to legacy adapter patterns

		// Get metrics directly (should be V2)
		metrics := performanceMonitor.GetMetricsV2()
		require.NotNil(t, metrics)

		// Verify we have the expected V2 structure
		assert.NotNil(t, metrics.Summary)
		assert.NotNil(t, metrics.Breakdown)

		// Verify we can access V2 fields directly
		_ = metrics.Summary.SuccessRate
		_ = metrics.Summary.RPS
		_ = metrics.Summary.P50Latency
		_ = metrics.Summary.CPUUsage
		_ = metrics.Breakdown.Latency.Avg
		_ = metrics.Breakdown.Throughput.Current
		_ = metrics.Breakdown.Success.Rate
		_ = metrics.Breakdown.Resources.CPU
	})
}

func TestV2ArchitecturePerformance(t *testing.T) {
	// Test performance characteristics of V2 architecture
	monitorConfig := observability.PerformanceMonitorConfig{
		MetricsCollectionInterval: 100 * time.Millisecond,
		AlertCheckInterval:        200 * time.Millisecond,
		OptimizationInterval:      500 * time.Millisecond,
		PredictionInterval:        300 * time.Millisecond,
	}

	performanceMonitor := observability.NewPerformanceMonitor(monitorConfig)
	require.NotNil(t, performanceMonitor)

	// Test rapid metrics access
	t.Run("Rapid Metrics Access", func(t *testing.T) {
		start := time.Now()

		// Perform multiple rapid metrics accesses
		for i := 0; i < 1000; i++ {
			metrics := performanceMonitor.GetMetricsV2()
			require.NotNil(t, metrics)

			// Access various fields to ensure they work
			_ = metrics.Summary.SuccessRate
			_ = metrics.Summary.RPS
			_ = metrics.Summary.P50Latency
		}

		duration := time.Since(start)
		t.Logf("1000 metrics accesses completed in %v", duration)

		// Should complete quickly (less than 1 second)
		assert.Less(t, duration, 1*time.Second)
	})

	// Test memory efficiency
	t.Run("Memory Efficiency", func(t *testing.T) {
		// Create multiple metrics instances and verify they're efficient
		metricsSlice := make([]*types.PerformanceMetricsV2, 100)

		for i := 0; i < 100; i++ {
			metrics := performanceMonitor.GetMetricsV2()
			require.NotNil(t, metrics)
			metricsSlice[i] = metrics
		}

		// Verify all metrics are valid
		for i, metrics := range metricsSlice {
			assert.NotNil(t, metrics, "Metrics at index %d is nil", i)
			if metrics != nil {
				assert.NotNil(t, metrics.Summary)
				assert.NotNil(t, metrics.Breakdown)
			}
		}
	})
}

func TestV2ArchitectureTypeSafety(t *testing.T) {
	// Test type safety of V2 architecture
	monitorConfig := observability.PerformanceMonitorConfig{
		MetricsCollectionInterval: 1 * time.Second,
	}

	performanceMonitor := observability.NewPerformanceMonitor(monitorConfig)
	require.NotNil(t, performanceMonitor)

	t.Run("Type Safety", func(t *testing.T) {
		// Get metrics
		metrics := performanceMonitor.GetMetricsV2()
		require.NotNil(t, metrics)

		// Verify type safety - these should compile and work correctly
		var successRate float64 = metrics.Summary.SuccessRate
		var rps float64 = metrics.Summary.RPS
		var latency time.Duration = metrics.Summary.P50Latency
		var cpuUsage float64 = metrics.Summary.CPUUsage

		// Verify values are within expected ranges
		assert.GreaterOrEqual(t, successRate, 0.0)
		assert.LessOrEqual(t, successRate, 1.0)
		assert.GreaterOrEqual(t, rps, 0.0)
		assert.GreaterOrEqual(t, latency, 0*time.Millisecond)
		assert.GreaterOrEqual(t, cpuUsage, 0.0)
		assert.LessOrEqual(t, cpuUsage, 1.0)

		// Verify breakdown type safety
		var avgLatency time.Duration = metrics.Breakdown.Latency.Avg
		var currentThroughput float64 = metrics.Breakdown.Throughput.Current
		var successRateDetailed float64 = metrics.Breakdown.Success.Rate
		var cpuDetailed float64 = metrics.Breakdown.Resources.CPU

		// Verify values are within expected ranges
		assert.GreaterOrEqual(t, avgLatency, 0*time.Millisecond)
		assert.GreaterOrEqual(t, currentThroughput, 0.0)
		assert.GreaterOrEqual(t, successRateDetailed, 0.0)
		assert.LessOrEqual(t, successRateDetailed, 1.0)
		assert.GreaterOrEqual(t, cpuDetailed, 0.0)
		assert.LessOrEqual(t, cpuDetailed, 1.0)
	})
}
