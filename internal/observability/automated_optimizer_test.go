package observability

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewAutomatedOptimizer(t *testing.T) {
	config := AutomatedOptimizerConfig{
		OptimizationCheckInterval:    60 * time.Second,
		PerformanceHistoryWindow:     48 * time.Hour,
		OptimizationCooldown:         10 * time.Minute,
		ResponseTimeThreshold:        1 * time.Second,
		SuccessRateThreshold:         0.90,
		ResourceUtilizationThreshold: 0.85,
		ThroughputThreshold:          50.0,
		EnableAutoScaling:            true,
		EnableCacheOptimization:      true,
		EnableDatabaseOptimization:   true,
		EnableConnectionPooling:      true,
		EnableLoadBalancing:          true,
		MaxOptimizationsPerHour:      5,
		OptimizationImpactLimit:      0.30,
		RollbackThreshold:            -0.15,
		EnableMachineLearning:        true,
		LearningRate:                 0.02,
		HistoricalDataPoints:         2000,
	}

	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, config, logger)

	if optimizer == nil {
		t.Fatal("Expected optimizer to be created, got nil")
	}

	if optimizer.config.OptimizationCheckInterval != 60*time.Second {
		t.Errorf("Expected optimization check interval to be 60 seconds, got %v", optimizer.config.OptimizationCheckInterval)
	}

	if optimizer.config.ResponseTimeThreshold != 1*time.Second {
		t.Errorf("Expected response time threshold to be 1 second, got %v", optimizer.config.ResponseTimeThreshold)
	}

	if optimizer.config.SuccessRateThreshold != 0.90 {
		t.Errorf("Expected success rate threshold to be 0.90, got %f", optimizer.config.SuccessRateThreshold)
	}

	if optimizer.config.MaxOptimizationsPerHour != 5 {
		t.Errorf("Expected max optimizations per hour to be 5, got %d", optimizer.config.MaxOptimizationsPerHour)
	}

	if !optimizer.config.EnableAutoScaling {
		t.Error("Expected auto scaling to be enabled")
	}

	if !optimizer.config.EnableMachineLearning {
		t.Error("Expected machine learning to be enabled")
	}
}

func TestNewAutomatedOptimizerDefaults(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	if optimizer.config.OptimizationCheckInterval != 30*time.Second {
		t.Errorf("Expected default optimization check interval to be 30 seconds, got %v", optimizer.config.OptimizationCheckInterval)
	}

	if optimizer.config.PerformanceHistoryWindow != 24*time.Hour {
		t.Errorf("Expected default performance history window to be 24 hours, got %v", optimizer.config.PerformanceHistoryWindow)
	}

	if optimizer.config.OptimizationCooldown != 5*time.Minute {
		t.Errorf("Expected default optimization cooldown to be 5 minutes, got %v", optimizer.config.OptimizationCooldown)
	}

	if optimizer.config.ResponseTimeThreshold != 500*time.Millisecond {
		t.Errorf("Expected default response time threshold to be 500ms, got %v", optimizer.config.ResponseTimeThreshold)
	}

	if optimizer.config.SuccessRateThreshold != 0.95 {
		t.Errorf("Expected default success rate threshold to be 0.95, got %f", optimizer.config.SuccessRateThreshold)
	}

	if optimizer.config.ResourceUtilizationThreshold != 0.80 {
		t.Errorf("Expected default resource utilization threshold to be 0.80, got %f", optimizer.config.ResourceUtilizationThreshold)
	}

	if optimizer.config.ThroughputThreshold != 100.0 {
		t.Errorf("Expected default throughput threshold to be 100.0, got %f", optimizer.config.ThroughputThreshold)
	}

	if optimizer.config.MaxOptimizationsPerHour != 10 {
		t.Errorf("Expected default max optimizations per hour to be 10, got %d", optimizer.config.MaxOptimizationsPerHour)
	}

	if optimizer.config.OptimizationImpactLimit != 0.20 {
		t.Errorf("Expected default optimization impact limit to be 0.20, got %f", optimizer.config.OptimizationImpactLimit)
	}

	if optimizer.config.RollbackThreshold != -0.10 {
		t.Errorf("Expected default rollback threshold to be -0.10, got %f", optimizer.config.RollbackThreshold)
	}

	if optimizer.config.LearningRate != 0.01 {
		t.Errorf("Expected default learning rate to be 0.01, got %f", optimizer.config.LearningRate)
	}

	if optimizer.config.HistoricalDataPoints != 1000 {
		t.Errorf("Expected default historical data points to be 1000, got %d", optimizer.config.HistoricalDataPoints)
	}
}

func TestAutomatedOptimizerStartStop(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	ctx := context.Background()

	// Start the optimizer
	err := optimizer.Start(ctx)
	if err != nil {
		t.Errorf("Expected no error when starting optimizer, got %v", err)
	}

	// Wait a bit for goroutines to start
	time.Sleep(100 * time.Millisecond)

	// Stop the optimizer
	optimizer.Stop()

	// Wait a bit for goroutines to stop
	time.Sleep(100 * time.Millisecond)
}

func TestGetOptimizationState(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	// Get optimization state
	state := optimizer.GetOptimizationState()

	if state == nil {
		t.Fatal("Expected optimization state to be returned, got nil")
	}

	if state.Status != "active" {
		t.Errorf("Expected status to be 'active', got %s", state.Status)
	}

	if state.OptimizationsToday != 0 {
		t.Errorf("Expected optimizations today to be 0, got %d", state.OptimizationsToday)
	}

	if state.ActiveOptimizations == nil {
		t.Error("Expected active optimizations to be initialized")
	}

	if state.OptimizationHistory == nil {
		t.Error("Expected optimization history to be initialized")
	}

	if state.LearningData == nil {
		t.Error("Expected learning data to be initialized")
	}
}

func TestGetPerformanceHistory(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	// Get performance history (should be empty initially)
	history := optimizer.GetPerformanceHistory(10)

	if history == nil {
		t.Fatal("Expected performance history to be returned, got nil")
	}

	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d items", len(history))
	}

	// Test with limit
	history = optimizer.GetPerformanceHistory(5)
	if len(history) != 0 {
		t.Errorf("Expected empty history with limit, got %d items", len(history))
	}
}

func TestRegisterStrategy(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	// Create a custom strategy
	customStrategy := &CustomOptimizationStrategy{}

	// Register the strategy
	optimizer.RegisterStrategy(customStrategy)

	// Verify it was registered by checking if we can force an optimization
	err := optimizer.ForceOptimization("custom_strategy", map[string]interface{}{})
	if err != nil {
		t.Errorf("Expected no error when forcing optimization with registered strategy, got %v", err)
	}
}

func TestForceOptimization(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	// Test forcing optimization with existing strategy
	err := optimizer.ForceOptimization("cache_optimization", map[string]interface{}{
		"cache_size": "increase_50%",
	})
	if err != nil {
		t.Errorf("Expected no error when forcing cache optimization, got %v", err)
	}

	// Test forcing optimization with non-existent strategy
	err = optimizer.ForceOptimization("non_existent_strategy", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error when forcing optimization with non-existent strategy")
	}
}

func TestRollbackOptimization(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	// Test rolling back non-existent optimization
	err := optimizer.RollbackOptimization("non_existent_optimization")
	if err == nil {
		t.Error("Expected error when rolling back non-existent optimization")
	}

	// Test rolling back existing optimization (should fail since no optimizations are active)
	err = optimizer.RollbackOptimization("opt_1234567890")
	if err == nil {
		t.Error("Expected error when rolling back optimization that doesn't exist")
	}
}

func TestShouldOptimize(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	// Test with good performance metrics (should not optimize)
	goodMetrics := &PerformanceMetrics{
		AverageResponseTime: 200 * time.Millisecond,
		OverallSuccessRate:  0.98,
		CPUUsage:            0.60,
		RequestsPerSecond:   150.0,
		Timestamp:           time.Now(),
	}

	if optimizer.shouldOptimize(goodMetrics) {
		t.Error("Expected shouldOptimize to return false for good performance")
	}

	// Test with poor response time (should optimize)
	poorResponseTime := &PerformanceMetrics{
		AverageResponseTime: 1 * time.Second,
		OverallSuccessRate:  0.98,
		CPUUsage:            0.60,
		RequestsPerSecond:   150.0,
		Timestamp:           time.Now(),
	}

	if !optimizer.shouldOptimize(poorResponseTime) {
		t.Error("Expected shouldOptimize to return true for poor response time")
	}

	// Test with poor success rate (should optimize)
	poorSuccessRate := &PerformanceMetrics{
		AverageResponseTime: 200 * time.Millisecond,
		OverallSuccessRate:  0.85,
		CPUUsage:            0.60,
		RequestsPerSecond:   150.0,
		Timestamp:           time.Now(),
	}

	if !optimizer.shouldOptimize(poorSuccessRate) {
		t.Error("Expected shouldOptimize to return true for poor success rate")
	}

	// Test with high resource utilization (should optimize)
	highResourceUsage := &PerformanceMetrics{
		AverageResponseTime: 200 * time.Millisecond,
		OverallSuccessRate:  0.98,
		CPUUsage:            0.90,
		RequestsPerSecond:   150.0,
		Timestamp:           time.Now(),
	}

	if !optimizer.shouldOptimize(highResourceUsage) {
		t.Error("Expected shouldOptimize to return true for high resource usage")
	}

	// Test with low throughput (should optimize)
	lowThroughput := &PerformanceMetrics{
		AverageResponseTime: 200 * time.Millisecond,
		OverallSuccessRate:  0.98,
		CPUUsage:            0.60,
		RequestsPerSecond:   30.0,
		Timestamp:           time.Now(),
	}

	if !optimizer.shouldOptimize(lowThroughput) {
		t.Error("Expected shouldOptimize to return true for low throughput")
	}
}

func TestCanOptimize(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	// Initially should be able to optimize
	if !optimizer.canOptimize() {
		t.Error("Expected canOptimize to return true initially")
	}

	// Set optimizations today to max limit
	optimizer.optimizationState.OptimizationsToday = optimizer.config.MaxOptimizationsPerHour

	// Should not be able to optimize due to hourly limit
	if optimizer.canOptimize() {
		t.Error("Expected canOptimize to return false when hourly limit reached")
	}

	// Reset optimizations today
	optimizer.optimizationState.OptimizationsToday = 0

	// Set last optimization to recent time (within cooldown)
	optimizer.optimizationState.LastOptimization = time.Now().Add(-2 * time.Minute)

	// Should not be able to optimize due to cooldown
	if optimizer.canOptimize() {
		t.Error("Expected canOptimize to return false when in cooldown period")
	}
}

func TestFindApplicableStrategies(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	// Test with metrics that should trigger cache optimization
	cacheOptimizationMetrics := &PerformanceMetrics{
		CacheHitRate:        0.70,
		AverageResponseTime: 400 * time.Millisecond,
		Timestamp:           time.Now(),
	}

	strategies := optimizer.findApplicableStrategies(cacheOptimizationMetrics)

	if len(strategies) == 0 {
		t.Error("Expected to find applicable strategies for cache optimization metrics")
	}

	// Check if cache optimization strategy is included
	foundCacheStrategy := false
	for _, strategy := range strategies {
		if strategy.Name() == "cache_optimization" {
			foundCacheStrategy = true
			break
		}
	}

	if !foundCacheStrategy {
		t.Error("Expected to find cache optimization strategy")
	}

	// Test with metrics that should trigger database optimization
	databaseOptimizationMetrics := &PerformanceMetrics{
		DatabaseQueryEfficiency: 0.80,
		AverageResponseTime:     600 * time.Millisecond,
		Timestamp:               time.Now(),
	}

	strategies = optimizer.findApplicableStrategies(databaseOptimizationMetrics)

	foundDatabaseStrategy := false
	for _, strategy := range strategies {
		if strategy.Name() == "database_optimization" {
			foundDatabaseStrategy = true
			break
		}
	}

	if !foundDatabaseStrategy {
		t.Error("Expected to find database optimization strategy")
	}
}

func TestSelectBestStrategy(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	realTimeDashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, zap.NewNop())
	logger := zap.NewNop()

	optimizer := NewAutomatedOptimizer(successRateTracker, performanceMonitor, realTimeDashboard, AutomatedOptimizerConfig{}, logger)

	// Test with empty strategies list
	bestStrategy := optimizer.selectBestStrategy([]OptimizationStrategy{}, &PerformanceMetrics{})
	if bestStrategy != nil {
		t.Error("Expected no strategy to be selected from empty list")
	}

	// Test with single strategy
	singleStrategy := &CacheOptimizationStrategy{}
	strategies := []OptimizationStrategy{singleStrategy}
	metrics := &PerformanceMetrics{
		CacheHitRate: 0.70,
		Timestamp:    time.Now(),
	}

	bestStrategy = optimizer.selectBestStrategy(strategies, metrics)
	if bestStrategy == nil {
		t.Error("Expected strategy to be selected from single strategy list")
	}

	if bestStrategy.Name() != "cache_optimization" {
		t.Errorf("Expected cache optimization strategy, got %s", bestStrategy.Name())
	}

	// Test with multiple strategies (should select the one with highest impact)
	highImpactStrategy := &CustomOptimizationStrategy{impact: 0.50}
	lowImpactStrategy := &CustomOptimizationStrategy{impact: 0.20}
	strategies = []OptimizationStrategy{lowImpactStrategy, highImpactStrategy}

	bestStrategy = optimizer.selectBestStrategy(strategies, metrics)
	if bestStrategy == nil {
		t.Error("Expected strategy to be selected from multiple strategies")
	}

	if bestStrategy.GetExpectedImpact(metrics) != 0.50 {
		t.Errorf("Expected high impact strategy to be selected, got impact %f", bestStrategy.GetExpectedImpact(metrics))
	}
}

func TestOptimizationStrategies(t *testing.T) {
	// Test CacheOptimizationStrategy
	cacheStrategy := &CacheOptimizationStrategy{}

	if cacheStrategy.Name() != "cache_optimization" {
		t.Errorf("Expected cache optimization strategy name, got %s", cacheStrategy.Name())
	}

	// Test CanApply
	metrics := &PerformanceMetrics{
		CacheHitRate:        0.75,
		AverageResponseTime: 350 * time.Millisecond,
		Timestamp:           time.Now(),
	}

	if !cacheStrategy.CanApply(metrics) {
		t.Error("Expected cache strategy to be applicable")
	}

	// Test GetExpectedImpact
	impact := cacheStrategy.GetExpectedImpact(metrics)
	if impact <= 0 {
		t.Error("Expected positive impact from cache optimization")
	}

	// Test DatabaseOptimizationStrategy
	dbStrategy := &DatabaseOptimizationStrategy{}

	if dbStrategy.Name() != "database_optimization" {
		t.Errorf("Expected database optimization strategy name, got %s", dbStrategy.Name())
	}

	// Test CanApply
	metrics = &PerformanceMetrics{
		DatabaseQueryEfficiency: 0.80,
		AverageResponseTime:     600 * time.Millisecond,
		Timestamp:               time.Now(),
	}

	if !dbStrategy.CanApply(metrics) {
		t.Error("Expected database strategy to be applicable")
	}

	// Test GetExpectedImpact
	impact = dbStrategy.GetExpectedImpact(metrics)
	if impact <= 0 {
		t.Error("Expected positive impact from database optimization")
	}

	// Test ConnectionPoolOptimizationStrategy
	cpStrategy := &ConnectionPoolOptimizationStrategy{}

	if cpStrategy.Name() != "connection_pool_optimization" {
		t.Errorf("Expected connection pool optimization strategy name, got %s", cpStrategy.Name())
	}

	// Test CanApply
	metrics = &PerformanceMetrics{
		ConnectionPoolUtilization: 0.95,
		AverageResponseTime:       450 * time.Millisecond,
		Timestamp:                 time.Now(),
	}

	if !cpStrategy.CanApply(metrics) {
		t.Error("Expected connection pool strategy to be applicable")
	}

	// Test GetExpectedImpact
	impact = cpStrategy.GetExpectedImpact(metrics)
	if impact <= 0 {
		t.Error("Expected positive impact from connection pool optimization")
	}
}

// CustomOptimizationStrategy for testing
type CustomOptimizationStrategy struct {
	impact float64
}

func (cos *CustomOptimizationStrategy) Name() string {
	return "custom_strategy"
}

func (cos *CustomOptimizationStrategy) CanApply(metrics *PerformanceMetrics) bool {
	return true
}

func (cos *CustomOptimizationStrategy) Apply(ctx context.Context, metrics *PerformanceMetrics) (*OptimizationAction, error) {
	return &OptimizationAction{
		ID:             generateOptimizationID(),
		Strategy:       cos.Name(),
		Type:           "custom",
		Parameters:     map[string]interface{}{},
		ExpectedImpact: cos.GetExpectedImpact(metrics),
		Priority:       50,
		Timestamp:      time.Now(),
	}, nil
}

func (cos *CustomOptimizationStrategy) Rollback(ctx context.Context, action *OptimizationAction) error {
	return nil
}

func (cos *CustomOptimizationStrategy) GetExpectedImpact(metrics *PerformanceMetrics) float64 {
	return cos.impact
}
