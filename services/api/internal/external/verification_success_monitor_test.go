package external

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewVerificationSuccessMonitor(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	monitor := NewVerificationSuccessMonitor(nil, logger)
	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.config)
	assert.Equal(t, 0.90, monitor.config.TargetSuccessRate)
	assert.Equal(t, 0.85, monitor.config.AlertThreshold)

	// Test with custom config
	customConfig := &SuccessMonitorConfig{
		TargetSuccessRate: 0.95,
		AlertThreshold:    0.90,
	}
	monitor2 := NewVerificationSuccessMonitor(customConfig, logger)
	assert.NotNil(t, monitor2)
	assert.Equal(t, 0.95, monitor2.config.TargetSuccessRate)
	assert.Equal(t, 0.90, monitor2.config.AlertThreshold)
}

func TestDefaultSuccessMonitorConfig(t *testing.T) {
	config := DefaultSuccessMonitorConfig()

	assert.True(t, config.EnableRealTimeMonitoring)
	assert.True(t, config.EnableFailureAnalysis)
	assert.True(t, config.EnableTrendAnalysis)
	assert.True(t, config.EnableAlerting)
	assert.Equal(t, 0.90, config.TargetSuccessRate)
	assert.Equal(t, 0.85, config.AlertThreshold)
	assert.Equal(t, 30*24*time.Hour, config.MetricsRetentionPeriod)
	assert.Equal(t, 1*time.Hour, config.AnalysisWindow)
	assert.Equal(t, 24*time.Hour, config.TrendWindow)
	assert.Equal(t, 100, config.MinDataPoints)
	assert.Equal(t, 10000, config.MaxDataPoints)
}

func TestRecordAttempt(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)

	// Record successful attempt
	dataPoint := DataPoint{
		URL:          "https://example.com",
		Success:      true,
		ResponseTime: 2 * time.Second,
		StatusCode:   200,
	}

	err := monitor.RecordAttempt(context.Background(), dataPoint)
	assert.NoError(t, err)

	metrics := monitor.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalAttempts)
	assert.Equal(t, int64(1), metrics.SuccessfulAttempts)
	assert.Equal(t, int64(0), metrics.FailedAttempts)
	assert.Equal(t, 1.0, metrics.SuccessRate)
	assert.Equal(t, 2*time.Second, metrics.AverageResponseTime)

	// Record failed attempt
	failedDataPoint := DataPoint{
		URL:          "https://example2.com",
		Success:      false,
		ResponseTime: 5 * time.Second,
		StatusCode:   500,
		ErrorType:    "timeout",
		ErrorMessage: "request timeout",
	}

	err = monitor.RecordAttempt(context.Background(), failedDataPoint)
	assert.NoError(t, err)

	metrics = monitor.GetMetrics()
	assert.Equal(t, int64(2), metrics.TotalAttempts)
	assert.Equal(t, int64(1), metrics.SuccessfulAttempts)
	assert.Equal(t, int64(1), metrics.FailedAttempts)
	assert.Equal(t, 0.5, metrics.SuccessRate)
	assert.Equal(t, 3*time.Second+500*time.Millisecond, metrics.AverageResponseTime) // (2+5)/2 = 3.5
}

func TestGetMetrics(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)

	// Add some test data
	for i := 0; i < 5; i++ {
		success := i < 4 // 4 successful, 1 failed
		dataPoint := DataPoint{
			URL:          fmt.Sprintf("https://example%d.com", i),
			Success:      success,
			ResponseTime: time.Duration(i+1) * time.Second,
			StatusCode:   200,
		}
		err := monitor.RecordAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	metrics := monitor.GetMetrics()
	assert.Equal(t, int64(5), metrics.TotalAttempts)
	assert.Equal(t, int64(4), metrics.SuccessfulAttempts)
	assert.Equal(t, int64(1), metrics.FailedAttempts)
	assert.Equal(t, 0.8, metrics.SuccessRate)
	assert.Equal(t, 3*time.Second, metrics.AverageResponseTime) // (1+2+3+4+5)/5 = 3
	assert.Len(t, metrics.DataPoints, 5)
}

func TestAnalyzeFailures(t *testing.T) {
	logger := zap.NewNop()
	config := &SuccessMonitorConfig{
		EnableFailureAnalysis: true,
		AnalysisWindow:        1 * time.Hour,
	}
	monitor := NewVerificationSuccessMonitor(config, logger)

	// Add test data with failures
	testData := []DataPoint{
		{URL: "https://example1.com", Success: false, ErrorType: "timeout", StrategyUsed: "user_agent_rotation"},
		{URL: "https://example1.com", Success: false, ErrorType: "timeout", StrategyUsed: "user_agent_rotation"},
		{URL: "https://example2.com", Success: false, ErrorType: "blocked", StrategyUsed: "proxy_rotation"},
		{URL: "https://example3.com", Success: true, StrategyUsed: "direct"},
		{URL: "https://example4.com", Success: false, ErrorType: "timeout", StrategyUsed: "user_agent_rotation"},
	}

	for _, dp := range testData {
		dp.Timestamp = time.Now() // Ensure recent timestamps
		err := monitor.RecordAttempt(context.Background(), dp)
		assert.NoError(t, err)
	}

	analysis, err := monitor.AnalyzeFailures(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, int64(4), analysis.TotalFailures)
	assert.Equal(t, 0.8, analysis.FailureRate) // 4 failures out of 5 attempts

	// Check error type analysis
	assert.Equal(t, int64(3), analysis.CommonErrorTypes["timeout"])
	assert.Equal(t, int64(1), analysis.CommonErrorTypes["blocked"])

	// Check URL analysis
	assert.Equal(t, int64(2), analysis.ProblematicURLs["https://example1.com"])
	assert.Equal(t, int64(1), analysis.ProblematicURLs["https://example2.com"])
	assert.Equal(t, int64(1), analysis.ProblematicURLs["https://example4.com"])

	// Check strategy analysis
	assert.Equal(t, int64(3), analysis.StrategyFailures["user_agent_rotation"])
	assert.Equal(t, int64(1), analysis.StrategyFailures["proxy_rotation"])

	// Check recommendations
	assert.Len(t, analysis.Recommendations, 2) // timeout error and problematic URL
}

func TestAnalyzeTrends(t *testing.T) {
	logger := zap.NewNop()
	config := &SuccessMonitorConfig{
		EnableTrendAnalysis: true,
		TrendWindow:         24 * time.Hour,
		MinDataPoints:       5,
	}
	monitor := NewVerificationSuccessMonitor(config, logger)

	// Add test data with a clear trend (improving success rate)
	now := time.Now()
	testData := []DataPoint{
		{Timestamp: now.Add(-5 * time.Hour), Success: false, ResponseTime: 5 * time.Second},
		{Timestamp: now.Add(-4 * time.Hour), Success: false, ResponseTime: 4 * time.Second},
		{Timestamp: now.Add(-3 * time.Hour), Success: true, ResponseTime: 3 * time.Second},
		{Timestamp: now.Add(-2 * time.Hour), Success: true, ResponseTime: 2 * time.Second},
		{Timestamp: now.Add(-1 * time.Hour), Success: true, ResponseTime: 1 * time.Second},
	}

	for _, dp := range testData {
		err := monitor.RecordAttempt(context.Background(), dp)
		assert.NoError(t, err)
	}

	trends, err := monitor.AnalyzeTrends(context.Background())
	assert.NoError(t, err)

	assert.Equal(t, 24*time.Hour, trends.Period)
	assert.Greater(t, trends.SuccessRateTrend, 0.0) // Should be improving
	assert.Len(t, trends.Seasonality, 0)            // No seasonality with small dataset
	assert.Len(t, trends.Predictions, 6)            // 6 hour predictions
}

func TestAnalyzeTrendsInsufficientData(t *testing.T) {
	logger := zap.NewNop()
	config := &SuccessMonitorConfig{
		EnableTrendAnalysis: true,
		TrendWindow:         24 * time.Hour,
		MinDataPoints:       10, // Require more data points
	}
	monitor := NewVerificationSuccessMonitor(config, logger)

	// Add only 5 data points (less than required 10)
	for i := 0; i < 5; i++ {
		dataPoint := DataPoint{
			Timestamp: time.Now(),
			Success:   true,
		}
		err := monitor.RecordAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	trends, err := monitor.AnalyzeTrends(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient data points")
	assert.Nil(t, trends)
}

func TestGetSuccessRate(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)

	// Initially should be 0
	assert.Equal(t, 0.0, monitor.GetSuccessRate())

	// Add successful attempt
	dataPoint := DataPoint{Success: true}
	err := monitor.RecordAttempt(context.Background(), dataPoint)
	assert.NoError(t, err)
	assert.Equal(t, 1.0, monitor.GetSuccessRate())

	// Add failed attempt
	failedDataPoint := DataPoint{Success: false}
	err = monitor.RecordAttempt(context.Background(), failedDataPoint)
	assert.NoError(t, err)
	assert.Equal(t, 0.5, monitor.GetSuccessRate())
}

func TestIsTargetAchieved(t *testing.T) {
	logger := zap.NewNop()
	config := &SuccessMonitorConfig{
		TargetSuccessRate: 0.90, // 90%
	}
	monitor := NewVerificationSuccessMonitor(config, logger)

	// Initially should be false
	assert.False(t, monitor.IsTargetAchieved())

	// Add 9 successful attempts out of 10 (90%)
	for i := 0; i < 9; i++ {
		dataPoint := DataPoint{Success: true}
		err := monitor.RecordAttempt(context.Background(), dataPoint)
		assert.NoError(t, err)
	}

	// Add 1 failed attempt
	failedDataPoint := DataPoint{Success: false}
	err := monitor.RecordAttempt(context.Background(), failedDataPoint)
	assert.NoError(t, err)

	// Should now be at 90% and achieve target
	assert.True(t, monitor.IsTargetAchieved())
	assert.Equal(t, 0.9, monitor.GetSuccessRate())
}

func TestVerificationSuccessMonitorUpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)

	// Test valid config update
	newConfig := &SuccessMonitorConfig{
		TargetSuccessRate: 0.95,
		AlertThreshold:    0.90,
	}

	err := monitor.UpdateConfig(newConfig)
	assert.NoError(t, err)

	updatedConfig := monitor.GetConfig()
	assert.Equal(t, 0.95, updatedConfig.TargetSuccessRate)
	assert.Equal(t, 0.90, updatedConfig.AlertThreshold)

	// Test invalid config (nil)
	err = monitor.UpdateConfig(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config cannot be nil")

	// Test invalid config (target success rate out of range)
	invalidConfig := &SuccessMonitorConfig{
		TargetSuccessRate: 1.5, // Invalid: > 1
		AlertThreshold:    0.85,
	}

	err = monitor.UpdateConfig(invalidConfig)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "target success rate must be between 0 and 1")

	// Test invalid config (alert threshold >= target success rate)
	invalidConfig2 := &SuccessMonitorConfig{
		TargetSuccessRate: 0.90,
		AlertThreshold:    0.95, // Invalid: >= target
	}

	err = monitor.UpdateConfig(invalidConfig2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alert threshold must be less than target success rate")
}

func TestResetMetrics(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)

	// Add some data
	dataPoint := DataPoint{Success: true}
	err := monitor.RecordAttempt(context.Background(), dataPoint)
	assert.NoError(t, err)

	// Verify data exists
	metrics := monitor.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalAttempts)

	// Reset metrics
	monitor.ResetMetrics()

	// Verify data is reset
	metrics = monitor.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalAttempts)
	assert.Equal(t, int64(0), metrics.SuccessfulAttempts)
	assert.Equal(t, int64(0), metrics.FailedAttempts)
	assert.Equal(t, 0.0, metrics.SuccessRate)
	assert.Len(t, metrics.DataPoints, 0)
}

func TestDataPointStructFields(t *testing.T) {
	// Test that DataPoint struct has all expected fields
	dataPoint := DataPoint{
		Timestamp:     time.Now(),
		URL:           "https://example.com",
		Success:       true,
		ResponseTime:  2 * time.Second,
		StatusCode:    200,
		ErrorType:     "timeout",
		ErrorMessage:  "request timeout",
		StrategyUsed:  "user_agent_rotation",
		UserAgentUsed: "Mozilla/5.0",
		ProxyUsed:     &Proxy{Host: "proxy.example.com", Port: 8080},
		Metadata:      map[string]interface{}{"key": "value"},
	}

	assert.False(t, dataPoint.Timestamp.IsZero())
	assert.Equal(t, "https://example.com", dataPoint.URL)
	assert.True(t, dataPoint.Success)
	assert.Equal(t, 2*time.Second, dataPoint.ResponseTime)
	assert.Equal(t, 200, dataPoint.StatusCode)
	assert.Equal(t, "timeout", dataPoint.ErrorType)
	assert.Equal(t, "request timeout", dataPoint.ErrorMessage)
	assert.Equal(t, "user_agent_rotation", dataPoint.StrategyUsed)
	assert.Equal(t, "Mozilla/5.0", dataPoint.UserAgentUsed)
	assert.NotNil(t, dataPoint.ProxyUsed)
	assert.Equal(t, "proxy.example.com", dataPoint.ProxyUsed.Host)
	assert.Equal(t, 8080, dataPoint.ProxyUsed.Port)
	assert.NotNil(t, dataPoint.Metadata)
	assert.Equal(t, "value", dataPoint.Metadata["key"])
}

func TestFailureAnalysisStructFields(t *testing.T) {
	// Test that FailureAnalysis struct has all expected fields
	analysis := &FailureAnalysis{
		TotalFailures:     10,
		FailureRate:       0.2,
		CommonErrorTypes:  map[string]int64{"timeout": 5, "blocked": 3},
		ProblematicURLs:   map[string]int64{"https://example.com": 3},
		StrategyFailures:  map[string]int64{"user_agent_rotation": 4},
		TimeBasedPatterns: map[string]int64{"14": 2, "15": 3},
		Recommendations:   []FailureRecommendation{{Type: "strategy", Priority: "high"}},
		LastAnalyzed:      time.Now(),
	}

	assert.Equal(t, int64(10), analysis.TotalFailures)
	assert.Equal(t, 0.2, analysis.FailureRate)
	assert.Equal(t, int64(5), analysis.CommonErrorTypes["timeout"])
	assert.Equal(t, int64(3), analysis.ProblematicURLs["https://example.com"])
	assert.Equal(t, int64(4), analysis.StrategyFailures["user_agent_rotation"])
	assert.Equal(t, int64(2), analysis.TimeBasedPatterns["14"])
	assert.Len(t, analysis.Recommendations, 1)
	assert.False(t, analysis.LastAnalyzed.IsZero())
}

func TestTrendAnalysisStructFields(t *testing.T) {
	// Test that TrendAnalysis struct has all expected fields
	trends := &TrendAnalysis{
		Period:            24 * time.Hour,
		SuccessRateTrend:  0.05,
		VolumeTrend:       0.1,
		ResponseTimeTrend: -0.02,
		Seasonality:       map[string]float64{"14": 0.8, "15": 0.9},
		Predictions:       []Prediction{{Timestamp: time.Now(), SuccessRate: 0.85, Confidence: 0.7}},
		LastUpdated:       time.Now(),
	}

	assert.Equal(t, 24*time.Hour, trends.Period)
	assert.Equal(t, 0.05, trends.SuccessRateTrend)
	assert.Equal(t, 0.1, trends.VolumeTrend)
	assert.Equal(t, -0.02, trends.ResponseTimeTrend)
	assert.Equal(t, 0.8, trends.Seasonality["14"])
	assert.Len(t, trends.Predictions, 1)
	assert.False(t, trends.LastUpdated.IsZero())
}

func TestPredictionStructFields(t *testing.T) {
	// Test that Prediction struct has all expected fields
	prediction := Prediction{
		Timestamp:   time.Now(),
		SuccessRate: 0.85,
		Confidence:  0.7,
	}

	assert.False(t, prediction.Timestamp.IsZero())
	assert.Equal(t, 0.85, prediction.SuccessRate)
	assert.Equal(t, 0.7, prediction.Confidence)
}

func TestFailureRecommendationStructFields(t *testing.T) {
	// Test that FailureRecommendation struct has all expected fields
	recommendation := FailureRecommendation{
		Type:        "strategy",
		Priority:    "high",
		Description: "Strategy has many failures",
		Impact:      0.05,
		Action:      "Review and optimize strategy",
	}

	assert.Equal(t, "strategy", recommendation.Type)
	assert.Equal(t, "high", recommendation.Priority)
	assert.Equal(t, "Strategy has many failures", recommendation.Description)
	assert.Equal(t, 0.05, recommendation.Impact)
	assert.Equal(t, "Review and optimize strategy", recommendation.Action)
}

func TestCalculateSuccessRate(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)

	// Test with empty data points
	emptyDataPoints := []DataPoint{}
	rate := monitor.calculateSuccessRate(emptyDataPoints)
	assert.Equal(t, 0.0, rate)

	// Test with all successful
	successDataPoints := []DataPoint{
		{Success: true},
		{Success: true},
		{Success: true},
	}
	rate = monitor.calculateSuccessRate(successDataPoints)
	assert.Equal(t, 1.0, rate)

	// Test with mixed results
	mixedDataPoints := []DataPoint{
		{Success: true},
		{Success: false},
		{Success: true},
		{Success: false},
	}
	rate = monitor.calculateSuccessRate(mixedDataPoints)
	assert.Equal(t, 0.5, rate)
}

func TestCalculateAverageResponseTime(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)

	// Test with empty data points
	emptyDataPoints := []DataPoint{}
	avgTime := monitor.calculateAverageResponseTime(emptyDataPoints)
	assert.Equal(t, time.Duration(0), avgTime)

	// Test with single data point
	singleDataPoint := []DataPoint{{ResponseTime: 2 * time.Second}}
	avgTime = monitor.calculateAverageResponseTime(singleDataPoint)
	assert.Equal(t, 2*time.Second, avgTime)

	// Test with multiple data points
	multipleDataPoints := []DataPoint{
		{ResponseTime: 1 * time.Second},
		{ResponseTime: 2 * time.Second},
		{ResponseTime: 3 * time.Second},
	}
	avgTime = monitor.calculateAverageResponseTime(multipleDataPoints)
	assert.Equal(t, 2*time.Second, avgTime) // (1+2+3)/3 = 2
}

func TestCalculateSeasonality(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewVerificationSuccessMonitor(nil, logger)

	// Test with data points at different hours
	now := time.Now()
	dataPoints := []DataPoint{
		{Timestamp: now.Add(-2 * time.Hour), Success: true},  // Hour 14
		{Timestamp: now.Add(-2 * time.Hour), Success: false}, // Hour 14
		{Timestamp: now.Add(-1 * time.Hour), Success: true},  // Hour 15
		{Timestamp: now.Add(-1 * time.Hour), Success: true},  // Hour 15
	}

	seasonality := monitor.calculateSeasonality(dataPoints)

	// Should have success rates for hours 14 and 15
	assert.Len(t, seasonality, 2)
	assert.Equal(t, 0.5, seasonality["14"]) // 1 success out of 2 attempts
	assert.Equal(t, 1.0, seasonality["15"]) // 2 successes out of 2 attempts
}
