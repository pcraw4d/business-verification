package observability

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewPredictiveAnalytics(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{
		DataCollectionInterval:  30 * time.Second,
		DataRetentionPeriod:     24 * time.Hour,
		MaxHistoricalDataPoints: 1000,
		PredictionHorizons:      []time.Duration{5 * time.Minute, 15 * time.Minute, 1 * time.Hour},
		PredictionInterval:      5 * time.Minute,
		ConfidenceLevels:        []float64{0.8, 0.9, 0.95},
		ModelTypes:              []string{"linear", "exponential", "arima"},
		AutoRetrain:             true,
		RetrainInterval:         24 * time.Hour,
		FeatureWindow:           1 * time.Hour,
		TrendWindow:             24 * time.Hour,
		EnablePredictiveAlerts:  true,
		AlertThresholds: map[string]float64{
			"response_time": 1000.0,
			"success_rate":  0.95,
		},
		MaxConcurrentPredictions: 10,
		PredictionTimeout:        30 * time.Second,
	}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	assert.NotNil(t, pa)
	assert.Equal(t, performanceMonitor, pa.performanceMonitor)
	assert.Equal(t, alertingSystem, pa.alertingSystem)
	assert.Equal(t, config, pa.config)
	assert.NotNil(t, pa.models)
	assert.NotNil(t, pa.historicalData)
	assert.Equal(t, 0, pa.historicalDataSize)
}

func TestPredictiveAnalytics_StartStop(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{
		DataCollectionInterval: 100 * time.Millisecond, // Short interval for testing
		PredictionInterval:     200 * time.Millisecond, // Short interval for testing
	}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start the system
	err := pa.Start(ctx)
	assert.NoError(t, err)

	// Wait a bit for goroutines to start
	time.Sleep(50 * time.Millisecond)

	// Stop the system
	err = pa.Stop()
	assert.NoError(t, err)
}

func TestPredictiveAnalytics_CollectPerformanceData(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

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
	pa.performanceMonitor = &PerformanceMonitor{}
	// In a real test, you would mock the GetMetrics method

	// Collect data
	pa.collectPerformanceData()

	// Check that data was collected
	pa.mu.RLock()
	assert.Equal(t, 1, pa.historicalDataSize)
	assert.Len(t, pa.historicalData, 1)
	pa.mu.RUnlock()
}

func TestPredictiveAnalytics_CalculateFeatures(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	// Add some historical data
	dataPoint := &PerformanceDataPoint{
		Timestamp:    time.Now().UTC(),
		ResponseTime: 250 * time.Millisecond,
		SuccessRate:  0.98,
		Throughput:   1000.0,
		ErrorRate:    0.02,
		CPUUsage:     75.0,
		MemoryUsage:  80.0,
		DiskUsage:    85.0,
		NetworkIO:    100.0,
		ActiveUsers:  100,
		DataVolume:   1000000,
		Features:     make(map[string]float64),
	}

	pa.mu.Lock()
	pa.historicalData = append(pa.historicalData, dataPoint)
	pa.historicalDataSize++
	pa.mu.Unlock()

	// Calculate features
	pa.calculateFeatures(dataPoint)

	// Check that features were calculated
	assert.NotEmpty(t, dataPoint.Features)
	assert.Contains(t, dataPoint.Features, "hour_of_day")
	assert.Contains(t, dataPoint.Features, "hour_sin")
	assert.Contains(t, dataPoint.Features, "hour_cos")
	assert.Contains(t, dataPoint.Features, "day_of_week")
	assert.Contains(t, dataPoint.Features, "day_sin")
	assert.Contains(t, dataPoint.Features, "day_cos")
}

func TestPredictiveAnalytics_CalculateMovingAverages(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	// Add historical data
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(200+i*10) * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
			DiskUsage:    85.0,
			NetworkIO:    100.0,
			ActiveUsers:  100,
			DataVolume:   1000000,
			Features:     make(map[string]float64),
		}
		pa.mu.Lock()
		pa.historicalData = append(pa.historicalData, dataPoint)
		pa.historicalDataSize++
		pa.mu.Unlock()
	}

	// Create a new data point
	newDataPoint := &PerformanceDataPoint{
		Timestamp:    now.Add(10 * time.Minute),
		ResponseTime: 300 * time.Millisecond,
		SuccessRate:  0.98,
		Throughput:   1000.0,
		ErrorRate:    0.02,
		CPUUsage:     75.0,
		MemoryUsage:  80.0,
		DiskUsage:    85.0,
		NetworkIO:    100.0,
		ActiveUsers:  100,
		DataVolume:   1000000,
		Features:     make(map[string]float64),
	}

	// Calculate moving averages
	pa.calculateMovingAverages(newDataPoint)

	// Check that moving averages were calculated
	assert.Contains(t, newDataPoint.Features, "response_time_ma_5min")
	assert.Contains(t, newDataPoint.Features, "response_time_ma_1hour")
}

func TestPredictiveAnalytics_CalculateTrends(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	// Add historical data with increasing trend
	now := time.Now().UTC()
	for i := 0; i < 15; i++ {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(200+i*5) * time.Millisecond, // Increasing trend
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
			DiskUsage:    85.0,
			NetworkIO:    100.0,
			ActiveUsers:  100,
			DataVolume:   1000000,
			Features:     make(map[string]float64),
		}
		pa.mu.Lock()
		pa.historicalData = append(pa.historicalData, dataPoint)
		pa.historicalDataSize++
		pa.mu.Unlock()
	}

	// Create a new data point
	newDataPoint := &PerformanceDataPoint{
		Timestamp:    now.Add(15 * time.Minute),
		ResponseTime: 275 * time.Millisecond,
		SuccessRate:  0.98,
		Throughput:   1000.0,
		ErrorRate:    0.02,
		CPUUsage:     75.0,
		MemoryUsage:  80.0,
		DiskUsage:    85.0,
		NetworkIO:    100.0,
		ActiveUsers:  100,
		DataVolume:   1000000,
		Features:     make(map[string]float64),
	}

	// Calculate trends
	pa.calculateTrends(newDataPoint)

	// Check that trends were calculated
	assert.Contains(t, newDataPoint.Features, "response_time_trend")
	assert.Contains(t, newDataPoint.Features, "response_time_trend_strength")

	// Trend should be positive (increasing)
	trend := newDataPoint.Features["response_time_trend"]
	assert.Greater(t, trend, 0.0)
}

func TestPredictiveAnalytics_CalculateSeasonality(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	// Create a data point at a specific time
	dataPoint := &PerformanceDataPoint{
		Timestamp:    time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC), // Monday 2:30 PM
		ResponseTime: 250 * time.Millisecond,
		SuccessRate:  0.98,
		Throughput:   1000.0,
		ErrorRate:    0.02,
		CPUUsage:     75.0,
		MemoryUsage:  80.0,
		DiskUsage:    85.0,
		NetworkIO:    100.0,
		ActiveUsers:  100,
		DataVolume:   1000000,
		Features:     make(map[string]float64),
	}

	// Calculate seasonality
	pa.calculateSeasonality(dataPoint)

	// Check that seasonality features were calculated
	assert.Contains(t, dataPoint.Features, "hour_of_day")
	assert.Contains(t, dataPoint.Features, "hour_sin")
	assert.Contains(t, dataPoint.Features, "hour_cos")
	assert.Contains(t, dataPoint.Features, "day_of_week")
	assert.Contains(t, dataPoint.Features, "day_sin")
	assert.Contains(t, dataPoint.Features, "day_cos")

	// Check specific values
	assert.Equal(t, 14.0, dataPoint.Features["hour_of_day"])
	assert.Equal(t, 1.0, dataPoint.Features["day_of_week"]) // Monday
}

func TestPredictiveAnalytics_CalculateVolatility(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	// Add historical data with varying response times
	now := time.Now().UTC()
	responseTimes := []int{200, 250, 180, 300, 220, 280, 190, 310, 240, 260, 210, 290, 230, 270, 200, 320, 250, 280, 190, 300}

	for i, rt := range responseTimes {
		dataPoint := &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(rt) * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   1000.0,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
			DiskUsage:    85.0,
			NetworkIO:    100.0,
			ActiveUsers:  100,
			DataVolume:   1000000,
			Features:     make(map[string]float64),
		}
		pa.mu.Lock()
		pa.historicalData = append(pa.historicalData, dataPoint)
		pa.historicalDataSize++
		pa.mu.Unlock()
	}

	// Create a new data point
	newDataPoint := &PerformanceDataPoint{
		Timestamp:    now.Add(20 * time.Minute),
		ResponseTime: 250 * time.Millisecond,
		SuccessRate:  0.98,
		Throughput:   1000.0,
		ErrorRate:    0.02,
		CPUUsage:     75.0,
		MemoryUsage:  80.0,
		DiskUsage:    85.0,
		NetworkIO:    100.0,
		ActiveUsers:  100,
		DataVolume:   1000000,
		Features:     make(map[string]float64),
	}

	// Calculate volatility
	pa.calculateVolatility(newDataPoint)

	// Check that volatility features were calculated
	assert.Contains(t, newDataPoint.Features, "response_time_volatility")
	assert.Contains(t, newDataPoint.Features, "response_time_cv")

	// Volatility should be positive
	volatility := newDataPoint.Features["response_time_volatility"]
	assert.Greater(t, volatility, 0.0)
}

func TestPredictiveAnalytics_CalculateRatios(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	dataPoint := &PerformanceDataPoint{
		Timestamp:    time.Now().UTC(),
		ResponseTime: 250 * time.Millisecond,
		SuccessRate:  0.98,
		Throughput:   1000.0,
		ErrorRate:    0.02,
		CPUUsage:     75.0,
		MemoryUsage:  80.0,
		DiskUsage:    85.0,
		NetworkIO:    100.0,
		ActiveUsers:  100,
		DataVolume:   1000000,
		Features:     make(map[string]float64),
	}

	// Calculate ratios
	pa.calculateRatios(dataPoint)

	// Check that ratios were calculated
	assert.Contains(t, dataPoint.Features, "cpu_memory_ratio")
	assert.Contains(t, dataPoint.Features, "error_success_ratio")
	assert.Contains(t, dataPoint.Features, "throughput_per_user")
	assert.Contains(t, dataPoint.Features, "data_per_request")

	// Check specific values
	assert.Equal(t, 75.0/80.0, dataPoint.Features["cpu_memory_ratio"])
	assert.Equal(t, 0.02/0.98, dataPoint.Features["error_success_ratio"])
	assert.Equal(t, 1000.0/100.0, dataPoint.Features["throughput_per_user"])
	assert.Equal(t, 1000000.0/1000.0, dataPoint.Features["data_per_request"])
}

func TestLinearModel_Train(t *testing.T) {
	model := NewLinearModel("response_time")

	// Create training data
	data := make([]*PerformanceDataPoint, 15)
	now := time.Now().UTC()
	for i := 0; i < 15; i++ {
		data[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(200+i*10) * time.Millisecond, // Linear trend
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

	// Train the model
	err := model.Train(data)
	assert.NoError(t, err)
	assert.True(t, model.IsTrained())
	assert.Greater(t, model.GetAccuracy(), 0.0)
}

func TestLinearModel_Predict(t *testing.T) {
	model := NewLinearModel("response_time")

	// Train the model first
	data := make([]*PerformanceDataPoint, 15)
	now := time.Now().UTC()
	for i := 0; i < 15; i++ {
		data[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(200+i*10) * time.Millisecond,
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

	err := model.Train(data)
	assert.NoError(t, err)

	// Make prediction
	features := map[string]float64{
		"current_response_time": 350.0,
		"response_time_trend":   10.0,
	}

	prediction, err := model.Predict(features, 5*time.Minute)
	assert.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, "response_time", prediction.Metric)
	assert.Greater(t, prediction.PredictedValue, 0.0)
	assert.Equal(t, 0.8, prediction.Confidence)
	assert.Equal(t, 5*time.Minute, prediction.PredictionHorizon)
}

func TestExponentialModel_Train(t *testing.T) {
	model := NewExponentialModel("success_rate")

	// Create training data
	data := make([]*PerformanceDataPoint, 10)
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		data[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: 250 * time.Millisecond,
			SuccessRate:  0.98 - float64(i)*0.001, // Slight decreasing trend
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

	// Train the model
	err := model.Train(data)
	assert.NoError(t, err)
	assert.True(t, model.IsTrained())
	assert.Greater(t, model.GetAccuracy(), 0.0)
}

func TestExponentialModel_Predict(t *testing.T) {
	model := NewExponentialModel("success_rate")

	// Train the model first
	data := make([]*PerformanceDataPoint, 10)
	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		data[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: 250 * time.Millisecond,
			SuccessRate:  0.98 - float64(i)*0.001,
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

	err := model.Train(data)
	assert.NoError(t, err)

	// Make prediction
	features := map[string]float64{
		"current_success_rate": 0.97,
		"current_error_rate":   0.03,
	}

	prediction, err := model.Predict(features, 15*time.Minute)
	assert.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, "success_rate", prediction.Metric)
	assert.Greater(t, prediction.PredictedValue, 0.0)
	assert.LessOrEqual(t, prediction.PredictedValue, 1.0)
	assert.Equal(t, 0.75, prediction.Confidence)
}

func TestARIMAModel_Train(t *testing.T) {
	model := NewARIMAModel("throughput")

	// Create training data
	data := make([]*PerformanceDataPoint, 25)
	now := time.Now().UTC()
	for i := 0; i < 25; i++ {
		data[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: 250 * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   1000.0 + float64(i)*10, // Increasing trend
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
			DiskUsage:    85.0,
			NetworkIO:    100.0,
			ActiveUsers:  100,
			DataVolume:   1000000,
		}
	}

	// Train the model
	err := model.Train(data)
	assert.NoError(t, err)
	assert.True(t, model.IsTrained())
	assert.Greater(t, model.GetAccuracy(), 0.0)
}

func TestARIMAModel_Predict(t *testing.T) {
	model := NewARIMAModel("throughput")

	// Train the model first
	data := make([]*PerformanceDataPoint, 25)
	now := time.Now().UTC()
	for i := 0; i < 25; i++ {
		data[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: 250 * time.Millisecond,
			SuccessRate:  0.98,
			Throughput:   1000.0 + float64(i)*10,
			ErrorRate:    0.02,
			CPUUsage:     75.0,
			MemoryUsage:  80.0,
			DiskUsage:    85.0,
			NetworkIO:    100.0,
			ActiveUsers:  100,
			DataVolume:   1000000,
		}
	}

	err := model.Train(data)
	assert.NoError(t, err)

	// Make prediction
	features := map[string]float64{
		"current_throughput":       1250.0,
		"response_time_trend":      5.0,
		"response_time_volatility": 50.0,
	}

	prediction, err := model.Predict(features, 1*time.Hour)
	assert.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, "throughput", prediction.Metric)
	assert.Greater(t, prediction.PredictedValue, 0.0)
	assert.Equal(t, 0.85, prediction.Confidence)
}

func TestEnsembleModel_Train(t *testing.T) {
	model := NewEnsembleModel("response_time")

	// Create training data
	data := make([]*PerformanceDataPoint, 25)
	now := time.Now().UTC()
	for i := 0; i < 25; i++ {
		data[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(200+i*8) * time.Millisecond,
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

	// Train the model
	err := model.Train(data)
	assert.NoError(t, err)
	assert.True(t, model.IsTrained())
	assert.Greater(t, model.GetAccuracy(), 0.0)
}

func TestEnsembleModel_Predict(t *testing.T) {
	model := NewEnsembleModel("response_time")

	// Train the model first
	data := make([]*PerformanceDataPoint, 25)
	now := time.Now().UTC()
	for i := 0; i < 25; i++ {
		data[i] = &PerformanceDataPoint{
			Timestamp:    now.Add(time.Duration(i) * time.Minute),
			ResponseTime: time.Duration(200+i*8) * time.Millisecond,
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

	err := model.Train(data)
	assert.NoError(t, err)

	// Make prediction
	features := map[string]float64{
		"current_response_time": 400.0,
		"response_time_trend":   8.0,
	}

	prediction, err := model.Predict(features, 30*time.Minute)
	assert.NoError(t, err)
	assert.NotNil(t, prediction)
	assert.Equal(t, "response_time", prediction.Metric)
	assert.Greater(t, prediction.PredictedValue, 0.0)
	assert.Greater(t, prediction.Confidence, 0.0)
	assert.LessOrEqual(t, prediction.Confidence, 1.0)
}

func TestPredictiveAnalytics_GetPredictions(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	predictions := pa.GetPredictions()
	assert.NotNil(t, predictions)
	assert.IsType(t, []*PredictionResult{}, predictions)
}

func TestPredictiveAnalytics_GetPredictionAccuracy(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{
		ModelTypes: []string{"linear", "exponential"},
	}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	accuracies := pa.GetPredictionAccuracy()
	assert.NotNil(t, accuracies)
	assert.IsType(t, map[string]float64{}, accuracies)
}

func TestPredictiveAnalytics_GetHistoricalData(t *testing.T) {
	logger := zap.NewNop()
	config := PredictiveAnalyticsConfig{}

	performanceMonitor := &PerformanceMonitor{}
	alertingSystem := &PerformanceAlertingSystem{}

	pa := NewPredictiveAnalytics(performanceMonitor, alertingSystem, config, logger)

	// Add some historical data
	dataPoint := &PerformanceDataPoint{
		Timestamp:    time.Now().UTC(),
		ResponseTime: 250 * time.Millisecond,
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

	pa.mu.Lock()
	pa.historicalData = append(pa.historicalData, dataPoint)
	pa.historicalDataSize++
	pa.mu.Unlock()

	historicalData := pa.GetHistoricalData()
	assert.NotNil(t, historicalData)
	assert.Len(t, historicalData, 1)
	assert.Equal(t, dataPoint.Timestamp, historicalData[0].Timestamp)
}
