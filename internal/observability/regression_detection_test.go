package observability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewRegressionDetectionSystem(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{
		DetectionInterval:       5 * time.Minute,
		BaselineWindow:          24 * time.Hour,
		DetectionWindow:         1 * time.Hour,
		MinDataPoints:           30,
		MaxHistoricalDataPoints: 10000,
		ConfidenceLevel:         0.95,
		PValueThreshold:         0.05,
		EnableRegressionAlerts:  true,
		AutoBaselineUpdate:      true,
		BaselineUpdateInterval:  24 * time.Hour,
		MaxConcurrentDetections: 10,
		DetectionTimeout:        30 * time.Second,
	}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	assert.NotNil(t, rds)
	assert.Equal(t, performanceMonitor, rds.performanceMonitor)
	assert.Equal(t, predictiveAnalytics, rds.predictiveAnalytics)
	assert.Equal(t, alertingSystem, rds.alertingSystem)
	assert.Equal(t, config, rds.config)
	assert.NotNil(t, rds.detectors)
	assert.NotNil(t, rds.baselines)
	assert.NotNil(t, rds.historicalData)
	assert.NotNil(t, rds.regressionHistory)
}

func TestRegressionDetectionSystem_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{
		DetectionInterval:      100 * time.Millisecond, // Short interval for testing
		BaselineUpdateInterval: 200 * time.Millisecond, // Short interval for testing
	}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the system
	err := rds.Start(ctx)
	assert.NoError(t, err)

	// Wait a bit for goroutines to start
	time.Sleep(50 * time.Millisecond)

	// Stop the system
	err = rds.Stop()
	assert.NoError(t, err)
}

func TestRegressionDetectionSystem_CollectPerformanceData(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	// Mock performance metrics
	metrics := &PerformanceMetrics{
		AverageResponseTime:  250 * time.Millisecond,
		SuccessRate:          0.98,
		RequestsPerSecond:    1000.0,
		ErrorRate:            0.02,
		CPUUsage:             75.0,
		MemoryUsage:          80.0,
		DiskUsage:            85.0,
		NetworkIO:            100.0,
		ActiveUsers:          100,
		DataProcessingVolume: 1000000,
	}

	// Mock the GetMetrics method
	rds.performanceMonitor = &PerformanceMonitor{}
	// In a real test, you would mock the GetMetrics method

	// Collect data
	rds.collectPerformanceData()

	// Check that data was collected
	rds.mu.RLock()
	assert.Equal(t, 1, len(rds.historicalData))
	rds.mu.RUnlock()
}

func TestRegressionDetectionSystem_CalculateBaseline(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	// Create test data
	data := make([]*PerformanceDataPoint, 50)
	now := time.Now().UTC()
	for i := 0; i < 50; i++ {
		data[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(200+i) * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
			DiskUsage:    85.0,
			NetworkIO:    100.0,
			ActiveUsers:  100,
			DataVolume:   1000000,
		}
	}

	// Calculate baseline
	baseline := rds.calculateBaseline("response_time", data)

	assert.NotNil(t, baseline)
	assert.Equal(t, "response_time", baseline.Metric)
	assert.True(t, baseline.IsActive)
	assert.Equal(t, 50, baseline.SampleSize)
	assert.Greater(t, baseline.Mean, 0.0)
	assert.Greater(t, baseline.StdDev, 0.0)
	assert.Equal(t, data[0].Timestamp, baseline.SampleStart)
	assert.Equal(t, data[49].Timestamp, baseline.SampleEnd)
}

func TestRegressionDetectionSystem_CalculateStatistics(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	// Test data
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}

	// Test mean calculation
	mean := rds.calculateMean(values)
	assert.Equal(t, 3.0, mean)

	// Test standard deviation calculation
	stdDev := rds.calculateStdDev(values, mean)
	assert.InDelta(t, 1.58, stdDev, 0.01)

	// Test min calculation
	min := rds.calculateMin(values)
	assert.Equal(t, 1.0, min)

	// Test max calculation
	max := rds.calculateMax(values)
	assert.Equal(t, 5.0, max)

	// Test percentile calculation
	p95 := rds.calculatePercentile(values, 95)
	assert.Equal(t, 5.0, p95) // Should be the last value for 95th percentile
}

func TestRegressionDetectionSystem_ExtractMetricValues(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	// Create test data
	data := []*PerformanceDataPoint{
		{
			Timestamp:    time.Now().UTC(),
			ResponseTime: 250 * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
		},
		{
			Timestamp:    time.Now().UTC(),
			ResponseTime: 300 * time.Millisecond,
			SuccessRate:  0.97,
			Throughput:   950.0,
			ErrorRate:    0.03,
			CPUUsage:     80.0,
			MemoryUsage:  85.0,
		},
	}

	// Test response time extraction
	responseTimeValues := rds.extractMetricValues("response_time", data)
	assert.Len(t, responseTimeValues, 2)
	assert.Equal(t, 250.0, responseTimeValues[0])
	assert.Equal(t, 300.0, responseTimeValues[1])

	// Test success rate extraction
	successRateValues := rds.extractMetricValues("success_rate", data)
	assert.Len(t, successRateValues, 2)
	assert.Equal(t, 0.98, successRateValues[0])
	assert.Equal(t, 0.97, successRateValues[1])

	// Test throughput extraction
	throughputValues := rds.extractMetricValues("throughput", data)
	assert.Len(t, throughputValues, 2)
	assert.Equal(t, 1000.0, throughputValues[0])
	assert.Equal(t, 950.0, throughputValues[1])
}

func TestRegressionDetectionSystem_GetCurrentDataWindow(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{
		DetectionWindow: 1 * time.Hour,
	}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	// Add historical data
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: 250 * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
		}
		rds.mu.Lock()
		rds.historicalData = append(rds.historicalData, dataPoint)
		rds.mu.Unlock()
	}

	// Get current data window
	windowData := rds.getCurrentDataWindow()

	// Should return data points within the detection window
	assert.Len(t, windowData, 10) // All 10 points should be within 1 hour
}

func TestRegressionDetectionSystem_GetBaseline(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	// Test getting existing baseline
	baseline := rds.GetBaseline("response_time")
	assert.NotNil(t, baseline)
	assert.Equal(t, "response_time", baseline.Metric)
	assert.True(t, baseline.IsActive)

	// Test getting non-existent baseline
	baseline = rds.GetBaseline("non_existent_metric")
	assert.Nil(t, baseline)
}

func TestRegressionDetectionSystem_GetBaselines(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	baselines := rds.GetBaselines()

	// Should have default baselines for all metrics
	expectedMetrics := []string{"response_time", "success_rate", "throughput", "error_rate", "cpu_usage", "memory_usage"}
	for _, metric := range expectedMetrics {
		assert.Contains(t, baselines, metric)
		assert.NotNil(t, baselines[metric])
		assert.Equal(t, metric, baselines[metric].Metric)
	}
}

func TestRegressionDetectionSystem_GetRegressionHistory(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	// Initially no regression history
	history := rds.GetRegressionHistory()
	assert.Empty(t, history)

	// Add a regression event
	event := &RegressionEvent{
		EventType:   "detected",
		Description: "Test regression",
		Severity:    "medium",
	}
	rds.AddRegressionEvent(event)

	// Check that event was added
	history = rds.GetRegressionHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, "Test regression", history[0].Description)
	assert.Equal(t, "medium", history[0].Severity)
}

func TestStatisticalDetector_Detect(t *testing.T) {
	config := RegressionDetectionConfig{
		ConfidenceLevel: 0.95,
		PValueThreshold: 0.05,
		RegressionThresholds: struct {
			ResponseTime struct {
				Degradation float64 `json:"degradation"`
				Improvement float64 `json:"improvement"`
			} `json:"response_time"`
			SuccessRate struct {
				Degradation float64 `json:"degradation"`
				Improvement float64 `json:"improvement"`
			} `json:"success_rate"`
			Throughput struct {
				Degradation float64 `json:"degradation"`
				Improvement float64 `json:"improvement"`
			} `json:"throughput"`
			ErrorRate struct {
				Degradation float64 `json:"degradation"`
				Improvement float64 `json:"improvement"`
			} `json:"error_rate"`
			ResourceUtilization struct {
				CPU struct {
					Degradation float64 `json:"degradation"`
					Improvement float64 `json:"improvement"`
				} `json:"cpu"`
				Memory struct {
					Degradation float64 `json:"degradation"`
					Improvement float64 `json:"improvement"`
				} `json:"memory"`
			} `json:"resource_utilization"`
		}{
			ResponseTime: struct {
				Degradation float64 `json:"degradation"`
				Improvement float64 `json:"improvement"`
			}{
				Degradation: 10.0,
				Improvement: 5.0,
			},
		},
	}

	detector := NewStatisticalDetector(config)

	// Create baseline
	baseline := &PerformanceBaseline{
		ID:          "test_baseline",
		Metric:      "response_time",
		Mean:        250.0,
		StdDev:      50.0,
		SampleSize:  100,
		IsActive:    true,
		SampleStart: time.Now().UTC().Add(-24 * time.Hour),
		SampleEnd:   time.Now().UTC(),
	}

	// Create current data with degradation
	currentData := make([]*PerformanceDataPoint, 15)
	now := time.Now().UTC()
	for i := 0; i < 15; i++ {
		currentData[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(300+i*5) * time.Millisecond, // Degraded performance
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
		}
	}

	// Perform detection
	result, err := detector.Detect(baseline, currentData)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "response_time", result.Metric)
	assert.Equal(t, "statistical_detector", result.DetectorUsed)
	assert.Greater(t, result.ChangePercent, 0.0) // Should detect degradation
}

func TestTrendDetector_Detect(t *testing.T) {
	config := RegressionDetectionConfig{
		RegressionThresholds: struct {
			ResponseTime struct {
				Degradation float64 `json:"degradation"`
				Improvement float64 `json:"improvement"`
			} `json:"response_time"`
			SuccessRate struct {
				Degradation float64 `json:"degradation"`
				Improvement float64 `json:"improvement"`
			} `json:"success_rate"`
			Throughput struct {
				Degradation float64 `json:"degradation"`
				Improvement float64 `json:"improvement"`
			} `json:"throughput"`
			ErrorRate struct {
				Degradation float64 `json:"degradation"`
				Improvement float64 `json:"improvement"`
			} `json:"error_rate"`
			ResourceUtilization struct {
				CPU struct {
					Degradation float64 `json:"degradation"`
					Improvement float64 `json:"improvement"`
				} `json:"cpu"`
				Memory struct {
					Degradation float64 `json:"degradation"`
					Improvement float64 `json:"improvement"`
				} `json:"memory"`
			} `json:"resource_utilization"`
		}{
			ResponseTime: struct {
				Degradation float64 `json:"degradation"`
				Improvement float64 `json:"improvement"`
			}{
				Degradation: 10.0,
				Improvement: 5.0,
			},
		},
	}

	detector := NewTrendDetector(config)

	// Create baseline
	baseline := &PerformanceBaseline{
		ID:          "test_baseline",
		Metric:      "response_time",
		Mean:        250.0,
		StdDev:      50.0,
		SampleSize:  100,
		IsActive:    true,
		SampleStart: time.Now().UTC().Add(-24 * time.Hour),
		SampleEnd:   time.Now().UTC(),
	}

	// Create current data with increasing trend
	currentData := make([]*PerformanceDataPoint, 25)
	now := time.Now().UTC()
	for i := 0; i < 25; i++ {
		currentData[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(250+i*2) * time.Millisecond, // Increasing trend
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
		}
	}

	// Perform detection
	result, err := detector.Detect(baseline, currentData)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "response_time", result.Metric)
	assert.Equal(t, "trend_detector", result.DetectorUsed)
	assert.NotNil(t, result.TrendAnalysis)
	assert.Equal(t, "increasing", result.TrendAnalysis.TrendDirection)
}

func TestThresholdDetector_Detect(t *testing.T) {
	config := RegressionDetectionConfig{}

	detector := NewThresholdDetector(config)

	// Create baseline
	baseline := &PerformanceBaseline{
		ID:           "test_baseline",
		Metric:       "response_time",
		Mean:         250.0,
		StdDev:       50.0,
		SampleSize:   100,
		IsActive:     true,
		SampleStart:  time.Now().UTC().Add(-24 * time.Hour),
		SampleEnd:    time.Now().UTC(),
		Percentile95: 350.0,
		Percentile99: 400.0,
	}

	// Create current data exceeding threshold
	currentData := make([]*PerformanceDataPoint, 10)
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		currentData[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(400) * time.Millisecond, // Exceeds threshold
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
		}
	}

	// Perform detection
	result, err := detector.Detect(baseline, currentData)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "response_time", result.Metric)
	assert.Equal(t, "threshold_detector", result.DetectorUsed)
	assert.Equal(t, "degradation", result.Type)
}

func TestAnomalyDetector_Detect(t *testing.T) {
	config := RegressionDetectionConfig{}

	detector := NewAnomalyDetector(config)

	// Create baseline
	baseline := &PerformanceBaseline{
		ID:          "test_baseline",
		Metric:      "response_time",
		Mean:        250.0,
		StdDev:      50.0,
		SampleSize:  100,
		IsActive:    true,
		SampleStart: time.Now().UTC().Add(-24 * time.Hour),
		SampleEnd:   time.Now().UTC(),
	}

	// Create current data with anomalies
	currentData := make([]*PerformanceDataPoint, 15)
	now := time.Now().UTC()
	for i := 0; i < 15; i++ {
		responseTime := 250.0
		if i%3 == 0 { // Every third point is an anomaly
			responseTime = 500.0 // Anomaly
		}
		currentData[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(responseTime) * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
		}
	}

	// Perform detection
	result, err := detector.Detect(baseline, currentData)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "response_time", result.Metric)
	assert.Equal(t, "anomaly_detector", result.DetectorUsed)
	assert.NotNil(t, result.OutlierAnalysis)
	assert.Greater(t, result.OutlierAnalysis.OutlierCount, 0)
}

func TestDetectorInterfaces(t *testing.T) {
	config := RegressionDetectionConfig{}

	// Test StatisticalDetector interface
	statDetector := NewStatisticalDetector(config)
	assert.Equal(t, "statistical_detector", statDetector.Name())
	assert.Equal(t, "statistical", statDetector.Type())
	assert.Equal(t, 0.95, statDetector.GetConfidence())
	assert.True(t, statDetector.IsApplicable("response_time"))

	// Test TrendDetector interface
	trendDetector := NewTrendDetector(config)
	assert.Equal(t, "trend_detector", trendDetector.Name())
	assert.Equal(t, "trend", trendDetector.Type())
	assert.Equal(t, 0.8, trendDetector.GetConfidence())
	assert.True(t, trendDetector.IsApplicable("success_rate"))

	// Test ThresholdDetector interface
	thresholdDetector := NewThresholdDetector(config)
	assert.Equal(t, "threshold_detector", thresholdDetector.Name())
	assert.Equal(t, "threshold", thresholdDetector.Type())
	assert.Equal(t, 0.9, thresholdDetector.GetConfidence())
	assert.True(t, thresholdDetector.IsApplicable("throughput"))

	// Test AnomalyDetector interface
	anomalyDetector := NewAnomalyDetector(config)
	assert.Equal(t, "anomaly_detector", anomalyDetector.Name())
	assert.Equal(t, "anomaly", anomalyDetector.Type())
	assert.Equal(t, 0.7, anomalyDetector.GetConfidence())
	assert.True(t, anomalyDetector.IsApplicable("error_rate"))
}

func TestRegressionDetectionSystem_UpdateBaseline(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{
		MinDataPoints: 10,
	}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	// Add historical data for baseline update
	now := time.Now().UTC()
	for i := 0; i < 20; i++ {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(250+i) * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
		}
		rds.mu.Lock()
		rds.historicalData = append(rds.historicalData, dataPoint)
		rds.mu.Unlock()
	}

	// Update baseline
	err := rds.UpdateBaseline("response_time")
	assert.NoError(t, err)

	// Check that baseline was updated
	baseline := rds.GetBaseline("response_time")
	assert.NotNil(t, baseline)
	assert.Greater(t, baseline.Mean, 0.0)
	assert.Greater(t, baseline.StdDev, 0.0)
}

func TestRegressionDetectionSystem_AddRegressionEvent(t *testing.T) {
	logger := zap.NewNop()
	config := RegressionDetectionConfig{}

	performanceMonitor := &PerformanceMonitor{}
	predictiveAnalytics := &PredictiveAnalytics{}
	alertingSystem := &PerformanceAlertingSystem{}

	rds := NewRegressionDetectionSystem(performanceMonitor, predictiveAnalytics, alertingSystem, config, logger)

	// Add regression event
	event := &RegressionEvent{
		ResultID:    "test_result",
		EventType:   "detected",
		Description: "Test regression event",
		Severity:    "high",
		User:        "test_user",
		Notes:       "Test notes",
	}

	err := rds.AddRegressionEvent(event)
	assert.NoError(t, err)

	// Check that event was added
	history := rds.GetRegressionHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, "test_result", history[0].ResultID)
	assert.Equal(t, "detected", history[0].EventType)
	assert.Equal(t, "Test regression event", history[0].Description)
	assert.Equal(t, "high", history[0].Severity)
	assert.Equal(t, "test_user", history[0].User)
	assert.Equal(t, "Test notes", history[0].Notes)
	assert.NotEmpty(t, history[0].ID)
	assert.NotZero(t, history[0].Timestamp)
}
