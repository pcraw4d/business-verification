package success_monitoring

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewSuccessRateMonitor(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	monitor := NewSuccessRateMonitor(nil, logger)
	assert.NotNil(t, monitor)
	assert.Equal(t, 0.95, monitor.config.TargetSuccessRate)
	assert.Equal(t, 0.90, monitor.config.WarningThreshold)
	assert.Equal(t, 0.85, monitor.config.CriticalThreshold)
	assert.True(t, monitor.config.EnableRealTimeMonitoring)

	// Test with custom config
	customConfig := &SuccessMonitorConfig{
		TargetSuccessRate: 0.98,
		WarningThreshold:  0.95,
		CriticalThreshold: 0.90,
	}

	monitor2 := NewSuccessRateMonitor(customConfig, logger)
	assert.NotNil(t, monitor2)
	assert.Equal(t, 0.98, monitor2.config.TargetSuccessRate)
	assert.Equal(t, 0.95, monitor2.config.WarningThreshold)
	assert.Equal(t, 0.90, monitor2.config.CriticalThreshold)
}

func TestDefaultSuccessMonitorConfig(t *testing.T) {
	config := DefaultSuccessMonitorConfig()

	assert.Equal(t, 0.95, config.TargetSuccessRate)
	assert.Equal(t, 0.90, config.WarningThreshold)
	assert.Equal(t, 0.85, config.CriticalThreshold)
	assert.True(t, config.EnableRealTimeMonitoring)
	assert.True(t, config.EnableFailureAnalysis)
	assert.True(t, config.EnableTrendAnalysis)
	assert.True(t, config.EnableAlerting)
	assert.Equal(t, 30*24*time.Hour, config.MetricsRetentionPeriod)
	assert.Equal(t, 1*time.Hour, config.AnalysisWindow)
	assert.Equal(t, 24*time.Hour, config.TrendWindow)
	assert.Equal(t, 100, config.MinDataPoints)
	assert.Equal(t, 10000, config.MaxDataPoints)
	assert.Equal(t, 5*time.Minute, config.AlertCooldownPeriod)
}

func TestSuccessRateMonitor_RecordProcessingAttempt(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSuccessRateMonitor(nil, logger)

	// Test successful attempt
	dataPoint := ProcessingDataPoint{
		ProcessName:     "test_process",
		InputType:       "business_data",
		Success:         true,
		ResponseTime:    100 * time.Millisecond,
		StatusCode:      200,
		ProcessingStage: "validation",
		InputSize:       1024,
		OutputSize:      512,
		ConfidenceScore: 0.95,
		Metadata: map[string]interface{}{
			"user_id": "123",
		},
	}

	err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
	assert.NoError(t, err)

	// Verify metrics
	metrics := monitor.GetProcessMetrics("test_process")
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(1), metrics.TotalAttempts)
	assert.Equal(t, int64(1), metrics.SuccessfulAttempts)
	assert.Equal(t, int64(0), metrics.FailedAttempts)
	assert.Equal(t, 1.0, metrics.SuccessRate)
	assert.Equal(t, 100*time.Millisecond, metrics.AverageResponseTime)

	// Test failed attempt
	failedDataPoint := ProcessingDataPoint{
		ProcessName:     "test_process",
		InputType:       "business_data",
		Success:         false,
		ResponseTime:    50 * time.Millisecond,
		StatusCode:      400,
		ErrorType:       "validation_error",
		ErrorMessage:    "Invalid input data",
		ProcessingStage: "validation",
	}

	err = monitor.RecordProcessingAttempt(context.Background(), failedDataPoint)
	assert.NoError(t, err)

	// Verify updated metrics
	metrics = monitor.GetProcessMetrics("test_process")
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(2), metrics.TotalAttempts)
	assert.Equal(t, int64(1), metrics.SuccessfulAttempts)
	assert.Equal(t, int64(1), metrics.FailedAttempts)
	assert.Equal(t, 0.5, metrics.SuccessRate)
	assert.Equal(t, 75*time.Millisecond, metrics.AverageResponseTime)
	assert.Equal(t, 1, metrics.FailurePatterns["validation_error"])
}

func TestSuccessRateMonitor_GetProcessMetrics(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSuccessRateMonitor(nil, logger)

	// Test non-existent process
	metrics := monitor.GetProcessMetrics("non_existent")
	assert.Nil(t, metrics)

	// Add some data
	dataPoint := ProcessingDataPoint{
		ProcessName:  "test_process",
		Success:      true,
		ResponseTime: 100 * time.Millisecond,
	}

	err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
	assert.NoError(t, err)

	// Get metrics
	metrics = monitor.GetProcessMetrics("test_process")
	assert.NotNil(t, metrics)
	assert.Equal(t, "test_process", metrics.ProcessName)
	assert.Equal(t, int64(1), metrics.TotalAttempts)
	assert.Equal(t, 1.0, metrics.SuccessRate)
}

func TestSuccessRateMonitor_GetAllProcessMetrics(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSuccessRateMonitor(nil, logger)

	// Initially should be empty
	allMetrics := monitor.GetAllProcessMetrics()
	assert.Empty(t, allMetrics)

	// Add data for multiple processes
	processes := []string{"process1", "process2", "process3"}

	for _, processName := range processes {
		dataPoint := ProcessingDataPoint{
			ProcessName:  processName,
			Success:      true,
			ResponseTime: 100 * time.Millisecond,
		}

		err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Get all metrics
	allMetrics = monitor.GetAllProcessMetrics()
	assert.Len(t, allMetrics, 3)

	for _, processName := range processes {
		metrics, exists := allMetrics[processName]
		assert.True(t, exists)
		assert.Equal(t, processName, metrics.ProcessName)
		assert.Equal(t, int64(1), metrics.TotalAttempts)
		assert.Equal(t, 1.0, metrics.SuccessRate)
	}
}

func TestSuccessRateMonitor_AnalyzeFailures(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultSuccessMonitorConfig()
	config.MinDataPoints = 5 // Lower for testing
	monitor := NewSuccessRateMonitor(config, logger)

	// Test with insufficient data
	analysis, err := monitor.AnalyzeFailures(context.Background(), "test_process")
	assert.Error(t, err)
	assert.Nil(t, analysis)
	assert.Contains(t, err.Error(), "process test_process not found")

	// Add sufficient data with failures
	for i := 0; i < 10; i++ {
		success := i < 7 // 70% success rate
		errorType := ""
		if !success {
			errorType = "validation_error"
		}

		dataPoint := ProcessingDataPoint{
			ProcessName:  "test_process",
			InputType:    "business_data",
			Success:      success,
			ResponseTime: 100 * time.Millisecond,
			StatusCode: func() int {
				if success {
					return 200
				} else {
					return 400
				}
			}(),
			ErrorType: errorType,
			ErrorMessage: func() string {
				if success {
					return ""
				} else {
					return "Invalid data"
				}
			}(),
			ProcessingStage: "validation",
		}

		err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Perform failure analysis
	analysis, err = monitor.AnalyzeFailures(context.Background(), "test_process")
	assert.NoError(t, err)
	assert.NotNil(t, analysis)

	assert.Equal(t, "test_process", analysis.ProcessName)
	assert.Equal(t, int64(3), analysis.TotalFailures)
	assert.Equal(t, 0.3, analysis.FailureRate)
	assert.Equal(t, 3, analysis.CommonErrorTypes["validation_error"])
	assert.Equal(t, 3, analysis.CommonErrorMessages["Invalid data"])
	assert.Equal(t, 3, analysis.ProblematicInputTypes["business_data"]) // Only failures count for problematic input types
	assert.Equal(t, 3, analysis.ProcessingStageFailures["validation"])
	assert.NotEmpty(t, analysis.Recommendations)
}

func TestSuccessRateMonitor_AnalyzeTrends(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultSuccessMonitorConfig()
	config.MinDataPoints = 5 // Lower for testing
	monitor := NewSuccessRateMonitor(config, logger)

	// Test with insufficient data
	analysis, err := monitor.AnalyzeTrends(context.Background(), "test_process")
	assert.Error(t, err)
	assert.Nil(t, analysis)
	assert.Contains(t, err.Error(), "process test_process not found")

	// Add data with improving trend (more recent successes)
	// Use timestamps within the analysis window (24 hours)
	baseTime := time.Now().Add(-12 * time.Hour) // Start 12 hours ago
	for i := 0; i < 10; i++ {
		success := i >= 5 // First 5 fail, last 5 succeed (improving trend)

		dataPoint := ProcessingDataPoint{
			ProcessName:  "test_process",
			InputType:    "business_data",
			Success:      success,
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    baseTime.Add(time.Duration(i) * time.Hour),
		}

		err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Perform trend analysis
	analysis, err = monitor.AnalyzeTrends(context.Background(), "test_process")
	assert.NoError(t, err)
	assert.NotNil(t, analysis)

	assert.Equal(t, "test_process", analysis.ProcessName)
	assert.True(t, analysis.SuccessRateTrend > 0) // Should be improving
	assert.NotEmpty(t, analysis.Predictions)
	assert.Greater(t, analysis.Confidence, 0.0)
}

func TestSuccessRateMonitor_GetAlerts(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultSuccessMonitorConfig()
	config.EnableAlerting = true
	monitor := NewSuccessRateMonitor(config, logger)

	// Initially no alerts
	alerts := monitor.GetAlerts()
	assert.Empty(t, alerts)

	// Add data that should trigger alerts (low success rate)
	for i := 0; i < 10; i++ {
		success := i < 2 // 20% success rate (below critical threshold)

		dataPoint := ProcessingDataPoint{
			ProcessName:  "test_process",
			Success:      success,
			ResponseTime: 100 * time.Millisecond,
		}

		err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Check for alerts
	alerts = monitor.GetAlerts()
	assert.NotEmpty(t, alerts)

	// Should have at least one critical alert
	hasCriticalAlert := false
	for _, alert := range alerts {
		if alert.AlertType == AlertTypeCritical {
			hasCriticalAlert = true
			assert.Equal(t, "test_process", alert.ProcessName)
			assert.False(t, alert.Resolved)
			// Check that the rate is below critical threshold (0.85)
			assert.Less(t, alert.CurrentRate, 0.85)
			assert.Equal(t, 0.95, alert.TargetRate)
		}
	}
	assert.True(t, hasCriticalAlert)
}

func TestSuccessRateMonitor_ResolveAlert(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultSuccessMonitorConfig()
	config.EnableAlerting = true
	monitor := NewSuccessRateMonitor(config, logger)

	// Create an alert by adding low success rate data
	for i := 0; i < 10; i++ {
		success := i < 2 // 20% success rate

		dataPoint := ProcessingDataPoint{
			ProcessName:  "test_process",
			Success:      success,
			ResponseTime: 100 * time.Millisecond,
		}

		err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Get alerts
	alerts := monitor.GetAlerts()
	assert.NotEmpty(t, alerts)

	// Resolve first alert
	alertID := alerts[0].ID
	err := monitor.ResolveAlert(alertID)
	assert.NoError(t, err)

	// Verify alert is resolved
	alerts = monitor.GetAlerts()
	for _, alert := range alerts {
		if alert.ID == alertID {
			assert.True(t, alert.Resolved)
			assert.NotNil(t, alert.ResolvedAt)
		}
	}

	// Test resolving non-existent alert
	err = monitor.ResolveAlert("non_existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestSuccessRateMonitor_GetSuccessRateReport(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSuccessRateMonitor(nil, logger)

	// Add data for multiple processes
	processes := []string{"process1", "process2"}

	for _, processName := range processes {
		for i := 0; i < 5; i++ {
			success := i < 4 // 80% success rate

			dataPoint := ProcessingDataPoint{
				ProcessName:  processName,
				Success:      success,
				ResponseTime: 100 * time.Millisecond,
			}

			err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
			assert.NoError(t, err)
		}
	}

	// Generate report for first process
	report, err := monitor.GetSuccessRateReport(context.Background(), processes[0])
	assert.NoError(t, err)
	assert.NotNil(t, report)

	// Verify report structure
	assert.NotZero(t, report.GeneratedAt)
	assert.Equal(t, processes[0], report.ProcessName)
	assert.Len(t, report.ProcessMetrics, 1)
	assert.NotNil(t, report.OverallMetrics)
	assert.NotNil(t, report.Recommendations)

	// Verify overall metrics
	assert.Equal(t, int64(10), report.OverallMetrics.TotalAttempts)
	assert.Equal(t, int64(8), report.OverallMetrics.TotalSuccesses)
	assert.Equal(t, int64(2), report.OverallMetrics.TotalFailures)
	assert.Equal(t, 0.8, report.OverallMetrics.AverageSuccessRate)

	// Verify process metrics
	metrics := report.ProcessMetrics[0]
	assert.Equal(t, int64(5), metrics.TotalAttempts)
	assert.Equal(t, int64(4), metrics.SuccessfulAttempts)
	assert.Equal(t, 0.8, metrics.SuccessRate)
}

func TestSuccessRateMonitor_HelperMethods(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewSuccessRateMonitor(nil, logger)

	// Test calculateSuccessRateFromDataPoints
	dataPoints := []ProcessingDataPoint{
		{Success: true},
		{Success: false},
		{Success: true},
		{Success: true},
		{Success: false},
	}

	successRate := monitor.calculateSuccessRateFromDataPoints(dataPoints)
	assert.Equal(t, 0.6, successRate) // 3/5 = 0.6

	// Test calculateFailureRateFromDataPoints
	failureRate := monitor.calculateFailureRateFromDataPoints(dataPoints)
	assert.Equal(t, 0.4, failureRate) // 2/5 = 0.4

	// Test calculateAverageResponseTime
	responseTimeDataPoints := []ProcessingDataPoint{
		{ResponseTime: 100 * time.Millisecond},
		{ResponseTime: 200 * time.Millisecond},
		{ResponseTime: 300 * time.Millisecond},
	}

	avgResponseTime := monitor.calculateAverageResponseTime(responseTimeDataPoints)
	assert.Equal(t, 200*time.Millisecond, avgResponseTime)

	// Test determineTrendDirection
	assert.Equal(t, TrendImproving, monitor.determineTrendDirection(0.05))
	assert.Equal(t, TrendDegrading, monitor.determineTrendDirection(-0.05))
	assert.Equal(t, TrendStable, monitor.determineTrendDirection(0.005))
}

func TestSuccessRateMonitor_DataPointCleanup(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultSuccessMonitorConfig()
	config.MetricsRetentionPeriod = 1 * time.Hour
	config.MaxDataPoints = 5
	monitor := NewSuccessRateMonitor(config, logger)

	// Add old data points
	oldTime := time.Now().Add(-2 * time.Hour)
	for i := 0; i < 3; i++ {
		dataPoint := ProcessingDataPoint{
			ProcessName:  "test_process",
			Success:      true,
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    oldTime.Add(time.Duration(i) * time.Minute),
		}

		err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Add recent data points
	for i := 0; i < 7; i++ {
		dataPoint := ProcessingDataPoint{
			ProcessName:  "test_process",
			Success:      true,
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    time.Now().Add(time.Duration(i) * time.Minute),
		}

		err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Verify cleanup (should keep only recent data points, max 5)
	metrics := monitor.GetProcessMetrics("test_process")
	assert.NotNil(t, metrics)
	assert.Len(t, metrics.DataPoints, 5)              // Max data points
	assert.Equal(t, int64(10), metrics.TotalAttempts) // Total attempts should still be 10
}

func TestSuccessRateMonitor_AlertCooldown(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultSuccessMonitorConfig()
	config.EnableAlerting = true
	config.AlertCooldownPeriod = 1 * time.Second
	monitor := NewSuccessRateMonitor(config, logger)

	// Add data to trigger alert
	for i := 0; i < 10; i++ {
		success := i < 2 // 20% success rate

		dataPoint := ProcessingDataPoint{
			ProcessName:  "test_process",
			Success:      success,
			ResponseTime: 100 * time.Millisecond,
		}

		err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Should have alerts
	alerts := monitor.GetAlerts()
	assert.NotEmpty(t, alerts)
	initialAlertCount := len(alerts)

	// Add more data immediately (should not trigger new alerts due to cooldown)
	for i := 0; i < 5; i++ {
		success := i < 1 // Still low success rate

		dataPoint := ProcessingDataPoint{
			ProcessName:  "test_process",
			Success:      success,
			ResponseTime: 100 * time.Millisecond,
		}

		err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Should have same number of alerts (cooldown in effect)
	alerts = monitor.GetAlerts()
	assert.Len(t, alerts, initialAlertCount)

	// Wait for cooldown to expire
	time.Sleep(2 * time.Second)

	// Add more data (should trigger new alerts)
	for i := 0; i < 5; i++ {
		success := i < 1

		dataPoint := ProcessingDataPoint{
			ProcessName:  "test_process",
			Success:      success,
			ResponseTime: 100 * time.Millisecond,
		}

		err := monitor.RecordProcessingAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Should have more alerts now
	alerts = monitor.GetAlerts()
	assert.Greater(t, len(alerts), initialAlertCount)
}

func TestSuccessRateMonitor_ConcurrentAccess(t *testing.T) {
	monitor := NewSuccessRateMonitor(DefaultSuccessMonitorConfig(), zap.NewNop())

	// Test concurrent access to the monitor
	var wg sync.WaitGroup
	numGoroutines := 10
	numOperations := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				processName := fmt.Sprintf("process_%d", id)
				dataPoint := ProcessingDataPoint{
					Timestamp:       time.Now(),
					ProcessName:     processName,
					Success:         true,
					ResponseTime:    100 * time.Millisecond,
					StatusCode:      200,
					ConfidenceScore: 0.8,
				}
				monitor.RecordProcessingAttempt(context.Background(), dataPoint)
			}
		}(i)
	}

	wg.Wait()

	// Verify all processes have metrics
	allMetrics := monitor.GetAllProcessMetrics()
	assert.Len(t, allMetrics, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		processName := fmt.Sprintf("process_%d", i)
		metrics, exists := allMetrics[processName]
		assert.True(t, exists)
		assert.Equal(t, int64(numOperations), metrics.TotalAttempts)
		assert.Equal(t, int64(numOperations), metrics.SuccessfulAttempts)
		assert.Equal(t, int64(0), metrics.FailedAttempts)
		assert.Equal(t, 1.0, metrics.SuccessRate)
	}
}

func TestSuccessRateMonitor_CreateSuccessRateOptimization(t *testing.T) {
	monitor := NewSuccessRateMonitor(DefaultSuccessMonitorConfig(), zap.NewNop())

	// Add some test data with failures to trigger optimization strategies
	processName := "test_process"

	// Add successful attempts
	for i := 0; i < 80; i++ {
		dataPoint := ProcessingDataPoint{
			Timestamp:       time.Now(),
			ProcessName:     processName,
			Success:         true,
			ResponseTime:    100 * time.Millisecond,
			StatusCode:      200,
			ConfidenceScore: 0.8,
		}
		monitor.RecordProcessingAttempt(context.Background(), dataPoint)
	}

	// Add failed attempts to create optimization opportunities
	for i := 0; i < 20; i++ {
		dataPoint := ProcessingDataPoint{
			Timestamp:       time.Now(),
			ProcessName:     processName,
			Success:         false,
			ResponseTime:    3 * time.Second,
			StatusCode:      500,
			ErrorType:       "timeout",
			ErrorMessage:    "Request timeout",
			ProcessingStage: "processing",
			InputSize:       1000,
			ConfidenceScore: 0.0,
		}
		monitor.RecordProcessingAttempt(context.Background(), dataPoint)
	}

	// Generate optimization
	optimization, err := monitor.CreateSuccessRateOptimization(context.Background(), processName)
	assert.NoError(t, err)
	assert.NotNil(t, optimization)

	// Verify optimization structure
	assert.Equal(t, processName, optimization.ProcessName)
	assert.Equal(t, 0.8, optimization.CurrentSuccessRate)        // 80/100 = 0.8
	assert.Equal(t, 0.95, optimization.TargetSuccessRate)        // From config
	assert.InDelta(t, 0.15, optimization.OptimizationGap, 0.001) // 0.95 - 0.8 = 0.15 (with tolerance for floating point)
	assert.True(t, optimization.ExpectedImprovement > 0)

	// Verify optimization strategies were generated
	assert.NotEmpty(t, optimization.OptimizationStrategies)

	// Verify performance tuning was generated
	assert.NotNil(t, optimization.PerformanceTuning)
	assert.NotNil(t, optimization.PerformanceTuning.ResponseTimeOptimization)
	assert.NotNil(t, optimization.PerformanceTuning.ThroughputOptimization)

	// Verify process improvements were generated
	assert.NotEmpty(t, optimization.ProcessImprovements)

	// Verify resource optimization was generated
	assert.NotNil(t, optimization.ResourceOptimization)
	assert.NotNil(t, optimization.ResourceOptimization.ScalingRecommendations)

	// Verify configuration optimization was generated
	assert.NotNil(t, optimization.ConfigurationOptimization)
	assert.NotNil(t, optimization.ConfigurationOptimization.ParameterOptimization)

	// Verify implementation plan was generated
	assert.NotNil(t, optimization.ImplementationPlan)
	assert.NotEmpty(t, optimization.ImplementationPlan.Phases)
	assert.NotNil(t, optimization.ImplementationPlan.RiskAssessment)
	assert.NotNil(t, optimization.ImplementationPlan.RollbackPlan)

	// Verify optimization status
	assert.NotNil(t, optimization.OptimizationStatus)

	// Test with non-existent process
	_, err = monitor.CreateSuccessRateOptimization(context.Background(), "non_existent_process")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no metrics found")
}

func TestSuccessRateMonitor_OptimizationStrategies(t *testing.T) {
	monitor := NewSuccessRateMonitor(DefaultSuccessMonitorConfig(), zap.NewNop())

	// Test optimization strategies generation
	metrics := &ProcessMetrics{
		ProcessName:         "test_process",
		TotalAttempts:       1001, // Greater than 1000 to trigger resource optimization
		SuccessfulAttempts:  800,
		FailedAttempts:      200,
		SuccessRate:         0.8,
		AverageResponseTime: 3 * time.Second, // High response time to trigger optimization
		FailurePatterns:     map[string]int{"timeout": 50, "validation_error": 30},
		LastUpdated:         time.Now(),
	}

	failureAnalysis := &FailureAnalysis{
		ProcessName:             "test_process",
		CommonErrorTypes:        map[string]int{"timeout": 50, "validation_error": 30},
		ProblematicInputTypes:   map[string]int{"large_input": 20},
		ProcessingStageFailures: map[string]int{"processing": 40},
		CorrelationAnalysis: CorrelationAnalysis{
			LoadCorrelation: 0.8, // High load correlation
		},
	}

	strategies := monitor.generateOptimizationStrategies(metrics, failureAnalysis)

	// Should generate multiple strategies
	assert.NotEmpty(t, strategies)

	// Check for specific strategies
	strategyNames := make(map[string]bool)
	for _, strategy := range strategies {
		strategyNames[strategy.Name] = true
	}

	// Debug: Print actual strategies generated
	t.Logf("Generated strategies: %v", strategyNames)

	// Should have error handling optimization (due to failure patterns)
	assert.True(t, strategyNames["Error Handling Optimization"], "Should have error handling optimization due to failure patterns")

	// Should have response time optimization (due to high response time > 2s)
	assert.True(t, strategyNames["Response Time Optimization"], "Should have response time optimization due to high response time")

	// Should have input validation optimization (due to validation errors in failure analysis)
	assert.True(t, strategyNames["Input Validation Optimization"], "Should have input validation optimization due to validation errors")

	// Should have resource optimization (due to high attempt count > 1000)
	assert.True(t, strategyNames["Resource Optimization"], "Should have resource optimization due to high attempt count")
}

func TestSuccessRateMonitor_PerformanceTuning(t *testing.T) {
	monitor := NewSuccessRateMonitor(DefaultSuccessMonitorConfig(), zap.NewNop())

	metrics := &ProcessMetrics{
		ProcessName:         "test_process",
		TotalAttempts:       1000,
		SuccessfulAttempts:  800,
		FailedAttempts:      200,
		SuccessRate:         0.8,
		AverageResponseTime: 2 * time.Second,
		LastUpdated:         time.Now(),
	}

	failureAnalysis := &FailureAnalysis{
		ProcessName:             "test_process",
		ProcessingStageFailures: map[string]int{"processing": 40},
		CorrelationAnalysis: CorrelationAnalysis{
			LoadCorrelation: 0.8,
		},
	}

	performanceTuning := monitor.generatePerformanceTuning(metrics, failureAnalysis)

	// Verify response time optimization
	assert.NotNil(t, performanceTuning.ResponseTimeOptimization)
	assert.Equal(t, 2*time.Second, performanceTuning.ResponseTimeOptimization.CurrentAverageResponseTime)
	assert.Equal(t, 1*time.Second, performanceTuning.ResponseTimeOptimization.TargetResponseTime)
	assert.NotEmpty(t, performanceTuning.ResponseTimeOptimization.Bottlenecks)
	assert.NotEmpty(t, performanceTuning.ResponseTimeOptimization.OptimizationStrategies)

	// Verify throughput optimization
	assert.NotNil(t, performanceTuning.ThroughputOptimization)
	assert.True(t, performanceTuning.ThroughputOptimization.CurrentThroughput > 0)
	assert.True(t, performanceTuning.ThroughputOptimization.TargetThroughput > performanceTuning.ThroughputOptimization.CurrentThroughput)

	// Verify resource optimization
	assert.NotNil(t, performanceTuning.ResourceOptimization)
	assert.NotNil(t, performanceTuning.ResourceOptimization.CPUOptimization)
	assert.NotNil(t, performanceTuning.ResourceOptimization.MemoryOptimization)
	assert.NotNil(t, performanceTuning.ResourceOptimization.NetworkOptimization)
	assert.NotNil(t, performanceTuning.ResourceOptimization.DiskOptimization)

	// Verify concurrency optimization
	assert.NotNil(t, performanceTuning.ConcurrencyOptimization)
	assert.True(t, performanceTuning.ConcurrencyOptimization.OptimalConcurrency > performanceTuning.ConcurrencyOptimization.CurrentConcurrency)

	// Verify cache optimization
	assert.NotNil(t, performanceTuning.CacheOptimization)
	assert.True(t, performanceTuning.CacheOptimization.TargetCacheHitRate > performanceTuning.CacheOptimization.CurrentCacheHitRate)
}

func TestSuccessRateMonitor_ImplementationPlan(t *testing.T) {
	monitor := NewSuccessRateMonitor(DefaultSuccessMonitorConfig(), zap.NewNop())

	strategies := []OptimizationStrategy{
		{
			ID:             "test_strategy",
			Name:           "Test Strategy",
			ExpectedImpact: 0.05,
		},
	}

	expectedImprovement := 0.15

	plan := monitor.generateImplementationPlan(strategies, expectedImprovement)

	// Verify plan structure
	assert.NotEmpty(t, plan.Phases)
	assert.Equal(t, 3, len(plan.Phases)) // Should have 3 phases

	// Verify phase details
	phase1 := plan.Phases[0]
	assert.Equal(t, 1, phase1.PhaseNumber)
	assert.Equal(t, "Quick Wins", phase1.Name)
	assert.Equal(t, "low", phase1.RiskLevel)

	phase2 := plan.Phases[1]
	assert.Equal(t, 2, phase2.PhaseNumber)
	assert.Equal(t, "Performance Optimization", phase2.Name)
	assert.Equal(t, "medium", phase2.RiskLevel)

	phase3 := plan.Phases[2]
	assert.Equal(t, 3, phase3.PhaseNumber)
	assert.Equal(t, "Advanced Optimizations", phase3.Name)
	assert.Equal(t, "high", phase3.RiskLevel)

	// Verify total duration calculation
	assert.True(t, plan.TotalDuration > 0)

	// Verify risk assessment
	assert.NotNil(t, plan.RiskAssessment)
	assert.Equal(t, "medium", plan.RiskAssessment.OverallRiskLevel)
	assert.NotEmpty(t, plan.RiskAssessment.RiskFactors)
	assert.NotEmpty(t, plan.RiskAssessment.MitigationStrategies)

	// Verify success criteria
	assert.NotEmpty(t, plan.SuccessCriteria)

	// Verify rollback plan
	assert.NotNil(t, plan.RollbackPlan)
	assert.NotEmpty(t, plan.RollbackPlan.RollbackTriggers)
	assert.NotEmpty(t, plan.RollbackPlan.RollbackSteps)
	assert.True(t, plan.RollbackPlan.RollbackDuration > 0)
}
