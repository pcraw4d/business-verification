package observability

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PredictiveAnalytics provides advanced predictive performance analytics
type PredictiveAnalytics struct {
	// Core components
	performanceMonitor *PerformanceMonitor
	alertingSystem     *PerformanceAlertingSystem

	// Prediction models
	models map[string]PredictionModel

	// Historical data management
	historicalData     []*PerformanceDataPoint
	historicalDataSize int
	dataRetention      time.Duration

	// Prediction settings
	config PredictiveAnalyticsConfig

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *zap.Logger

	// Control channels
	stopChannel chan struct{}
}

// PredictiveAnalyticsConfig holds configuration for predictive analytics
type PredictiveAnalyticsConfig struct {
	// Data collection settings
	DataCollectionInterval  time.Duration `json:"data_collection_interval"`
	DataRetentionPeriod     time.Duration `json:"data_retention_period"`
	MaxHistoricalDataPoints int           `json:"max_historical_data_points"`

	// Prediction settings
	PredictionHorizons []time.Duration `json:"prediction_horizons"` // 5min, 15min, 1hour, 6hours, 24hours
	PredictionInterval time.Duration   `json:"prediction_interval"`
	ConfidenceLevels   []float64       `json:"confidence_levels"` // 0.8, 0.9, 0.95, 0.99

	// Model settings
	ModelTypes      []string      `json:"model_types"` // linear, exponential, arima, lstm, ensemble
	AutoRetrain     bool          `json:"auto_retrain"`
	RetrainInterval time.Duration `json:"retrain_interval"`

	// Feature engineering
	FeatureWindow     time.Duration `json:"feature_window"`
	SeasonalityPeriod time.Duration `json:"seasonality_period"`
	TrendWindow       time.Duration `json:"trend_window"`

	// Alerting settings
	EnablePredictiveAlerts bool               `json:"enable_predictive_alerts"`
	AlertThresholds        map[string]float64 `json:"alert_thresholds"`

	// Performance settings
	MaxConcurrentPredictions int           `json:"max_concurrent_predictions"`
	PredictionTimeout        time.Duration `json:"prediction_timeout"`
}

// PredictionModel defines a prediction model interface
type PredictionModel interface {
	Name() string
	Type() string
	Train(data []*PerformanceDataPoint) error
	Predict(features map[string]float64, horizon time.Duration) (*PredictionResult, error)
	GetAccuracy() float64
	GetLastTraining() time.Time
	IsTrained() bool
}

// PerformanceDataPoint represents a single performance data point
type PerformanceDataPoint struct {
	Timestamp time.Time `json:"timestamp"`

	// Core metrics
	ResponseTime time.Duration `json:"response_time"`
	SuccessRate  float64       `json:"success_rate"`
	Throughput   float64       `json:"throughput"`
	ErrorRate    float64       `json:"error_rate"`

	// Resource metrics
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage   float64 `json:"disk_usage"`
	NetworkIO   float64 `json:"network_io"`

	// Business metrics
	ActiveUsers int64 `json:"active_users"`
	DataVolume  int64 `json:"data_volume"`

	// Derived features
	Features map[string]float64 `json:"features,omitempty"`
}

// PredictionResult represents a prediction result
type PredictionResult struct {
	ID                string        `json:"id"`
	Metric            string        `json:"metric"`
	PredictedValue    float64       `json:"predicted_value"`
	Confidence        float64       `json:"confidence"`
	PredictionHorizon time.Duration `json:"prediction_horizon"`
	ModelUsed         string        `json:"model_used"`
	Timestamp         time.Time     `json:"timestamp"`

	// Uncertainty
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	StdDev     float64 `json:"std_dev"`

	// Trend analysis
	Trend         string  `json:"trend"`          // improving, stable, degrading
	TrendStrength float64 `json:"trend_strength"` // 0-1
	Seasonality   bool    `json:"seasonality"`

	// Contributing factors
	Factors []PredictionFactor `json:"factors"`

	// Model metadata
	ModelAccuracy float64   `json:"model_accuracy"`
	LastTraining  time.Time `json:"last_training"`
}

// PredictionFactor represents a factor contributing to the prediction
type PredictionFactor struct {
	Name        string  `json:"name"`
	Impact      float64 `json:"impact"` // -1 to 1, negative means negative impact
	Confidence  float64 `json:"confidence"`
	Description string  `json:"description"`
}

// PredictionAlert represents a predictive alert
type PredictionAlert struct {
	ID             string        `json:"id"`
	Metric         string        `json:"metric"`
	AlertType      string        `json:"alert_type"` // threshold_breach, trend_change, anomaly
	Severity       string        `json:"severity"`
	Message        string        `json:"message"`
	PredictedValue float64       `json:"predicted_value"`
	Threshold      float64       `json:"threshold"`
	Horizon        time.Duration `json:"horizon"`
	Confidence     float64       `json:"confidence"`
	Timestamp      time.Time     `json:"timestamp"`
}

// NewPredictiveAnalytics creates a new predictive analytics system
func NewPredictiveAnalytics(
	performanceMonitor *PerformanceMonitor,
	alertingSystem *PerformanceAlertingSystem,
	config PredictiveAnalyticsConfig,
	logger *zap.Logger,
) *PredictiveAnalytics {
	// Set default values
	if config.DataCollectionInterval == 0 {
		config.DataCollectionInterval = 30 * time.Second
	}
	if config.DataRetentionPeriod == 0 {
		config.DataRetentionPeriod = 30 * 24 * time.Hour // 30 days
	}
	if config.MaxHistoricalDataPoints == 0 {
		config.MaxHistoricalDataPoints = 10000
	}
	if config.PredictionInterval == 0 {
		config.PredictionInterval = 5 * time.Minute
	}
	if config.FeatureWindow == 0 {
		config.FeatureWindow = 1 * time.Hour
	}
	if config.TrendWindow == 0 {
		config.TrendWindow = 24 * time.Hour
	}
	if config.RetrainInterval == 0 {
		config.RetrainInterval = 24 * time.Hour
	}
	if config.PredictionTimeout == 0 {
		config.PredictionTimeout = 30 * time.Second
	}

	pa := &PredictiveAnalytics{
		performanceMonitor: performanceMonitor,
		alertingSystem:     alertingSystem,
		models:             make(map[string]PredictionModel),
		historicalData:     make([]*PerformanceDataPoint, 0),
		config:             config,
		logger:             logger,
		stopChannel:        make(chan struct{}),
	}

	// Initialize prediction models
	pa.initializeModels()

	return pa
}

// Start starts the predictive analytics system
func (pa *PredictiveAnalytics) Start(ctx context.Context) error {
	pa.logger.Info("Starting predictive analytics system")

	// Start data collection
	go pa.collectData(ctx)

	// Start predictions
	go pa.runPredictions(ctx)

	// Start model retraining
	if pa.config.AutoRetrain {
		go pa.retrainModels(ctx)
	}

	// Start predictive alerts
	if pa.config.EnablePredictiveAlerts {
		go pa.monitorPredictiveAlerts(ctx)
	}

	pa.logger.Info("Predictive analytics system started")
	return nil
}

// Stop stops the predictive analytics system
func (pa *PredictiveAnalytics) Stop() error {
	pa.logger.Info("Stopping predictive analytics system")

	close(pa.stopChannel)

	pa.logger.Info("Predictive analytics system stopped")
	return nil
}

// collectData collects performance data for analysis
func (pa *PredictiveAnalytics) collectData(ctx context.Context) {
	ticker := time.NewTicker(pa.config.DataCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pa.stopChannel:
			return
		case <-ticker.C:
			pa.collectPerformanceData()
		}
	}
}

// collectPerformanceData collects current performance data
func (pa *PredictiveAnalytics) collectPerformanceData() {
	metrics := pa.performanceMonitor.GetMetrics()
	if metrics == nil {
		return
	}

	dataPoint := &PerformanceDataPoint{
		Timestamp:    time.Now().UTC(),
		ResponseTime: metrics.AverageResponseTime,
		SuccessRate:  metrics.SuccessRate,
		Throughput:   metrics.RequestsPerSecond,
		ErrorRate:    metrics.ErrorRate,
		CPUUsage:     metrics.CPUUsage,
		MemoryUsage:  metrics.MemoryUsage,
		DiskUsage:    metrics.DiskUsage,
		NetworkIO:    metrics.NetworkIO,
		ActiveUsers:  int64(metrics.ActiveUsers),
		DataVolume:   metrics.DataProcessingVolume,
		Features:     make(map[string]float64),
	}

	// Calculate derived features
	pa.calculateFeatures(dataPoint)

	pa.mu.Lock()
	pa.historicalData = append(pa.historicalData, dataPoint)
	pa.historicalDataSize++

	// Maintain data retention
	if pa.historicalDataSize > pa.config.MaxHistoricalDataPoints {
		pa.historicalData = pa.historicalData[1:]
		pa.historicalDataSize--
	}

	// Remove old data
	cutoff := time.Now().UTC().Add(-pa.config.DataRetentionPeriod)
	for i, point := range pa.historicalData {
		if point.Timestamp.After(cutoff) {
			pa.historicalData = pa.historicalData[i:]
			pa.historicalDataSize = len(pa.historicalData)
			break
		}
	}
	pa.mu.Unlock()

	pa.logger.Debug("Collected performance data point",
		zap.Time("timestamp", dataPoint.Timestamp),
		zap.Duration("response_time", dataPoint.ResponseTime),
		zap.Float64("success_rate", dataPoint.SuccessRate))
}

// calculateFeatures calculates derived features for prediction
func (pa *PredictiveAnalytics) calculateFeatures(dataPoint *PerformanceDataPoint) {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	if len(pa.historicalData) == 0 {
		return
	}

	// Calculate moving averages
	pa.calculateMovingAverages(dataPoint)

	// Calculate trends
	pa.calculateTrends(dataPoint)

	// Calculate seasonality
	pa.calculateSeasonality(dataPoint)

	// Calculate volatility
	pa.calculateVolatility(dataPoint)

	// Calculate ratios and rates
	pa.calculateRatios(dataPoint)
}

// calculateMovingAverages calculates moving averages
func (pa *PredictiveAnalytics) calculateMovingAverages(dataPoint *PerformanceDataPoint) {
	// 5-minute moving average
	window5min := time.Now().UTC().Add(-5 * time.Minute)
	var sum5min float64
	count5min := 0

	for i := len(pa.historicalData) - 1; i >= 0; i-- {
		point := pa.historicalData[i]
		if point.Timestamp.Before(window5min) {
			break
		}
		sum5min += float64(point.ResponseTime.Milliseconds())
		count5min++
	}

	if count5min > 0 {
		dataPoint.Features["response_time_ma_5min"] = sum5min / float64(count5min)
	}

	// 1-hour moving average
	window1hour := time.Now().UTC().Add(-1 * time.Hour)
	var sum1hour float64
	count1hour := 0

	for i := len(pa.historicalData) - 1; i >= 0; i-- {
		point := pa.historicalData[i]
		if point.Timestamp.Before(window1hour) {
			break
		}
		sum1hour += float64(point.ResponseTime.Milliseconds())
		count1hour++
	}

	if count1hour > 0 {
		dataPoint.Features["response_time_ma_1hour"] = sum1hour / float64(count1hour)
	}
}

// calculateTrends calculates trend indicators
func (pa *PredictiveAnalytics) calculateTrends(dataPoint *PerformanceDataPoint) {
	if len(pa.historicalData) < 10 {
		return
	}

	// Calculate linear trend for response time
	recentPoints := pa.historicalData[len(pa.historicalData)-10:]
	var xSum, ySum, xySum, x2Sum float64

	for i, point := range recentPoints {
		x := float64(i)
		y := float64(point.ResponseTime.Milliseconds())
		xSum += x
		ySum += y
		xySum += x * y
		x2Sum += x * x
	}

	n := float64(len(recentPoints))
	slope := (n*xySum - xSum*ySum) / (n*x2Sum - xSum*xSum)

	dataPoint.Features["response_time_trend"] = slope
	dataPoint.Features["response_time_trend_strength"] = math.Abs(slope) / 100.0 // Normalize
}

// calculateSeasonality calculates seasonality indicators
func (pa *PredictiveAnalytics) calculateSeasonality(dataPoint *PerformanceDataPoint) {
	// Calculate hour-of-day seasonality
	hour := float64(dataPoint.Timestamp.Hour())
	dataPoint.Features["hour_of_day"] = hour
	dataPoint.Features["hour_sin"] = math.Sin(2 * math.Pi * hour / 24)
	dataPoint.Features["hour_cos"] = math.Cos(2 * math.Pi * hour / 24)

	// Calculate day-of-week seasonality
	weekday := float64(dataPoint.Timestamp.Weekday())
	dataPoint.Features["day_of_week"] = weekday
	dataPoint.Features["day_sin"] = math.Sin(2 * math.Pi * weekday / 7)
	dataPoint.Features["day_cos"] = math.Cos(2 * math.Pi * weekday / 7)
}

// calculateVolatility calculates volatility indicators
func (pa *PredictiveAnalytics) calculateVolatility(dataPoint *PerformanceDataPoint) {
	if len(pa.historicalData) < 20 {
		return
	}

	// Calculate standard deviation of response time
	recentPoints := pa.historicalData[len(pa.historicalData)-20:]
	var sum, sumSq float64

	for _, point := range recentPoints {
		value := float64(point.ResponseTime.Milliseconds())
		sum += value
		sumSq += value * value
	}

	n := float64(len(recentPoints))
	mean := sum / n
	variance := (sumSq / n) - (mean * mean)
	stdDev := math.Sqrt(variance)

	dataPoint.Features["response_time_volatility"] = stdDev
	dataPoint.Features["response_time_cv"] = stdDev / mean // Coefficient of variation
}

// calculateRatios calculates ratio-based features
func (pa *PredictiveAnalytics) calculateRatios(dataPoint *PerformanceDataPoint) {
	// CPU to memory ratio
	if dataPoint.MemoryUsage > 0 {
		dataPoint.Features["cpu_memory_ratio"] = dataPoint.CPUUsage / dataPoint.MemoryUsage
	}

	// Error to success ratio
	if dataPoint.SuccessRate > 0 {
		dataPoint.Features["error_success_ratio"] = dataPoint.ErrorRate / dataPoint.SuccessRate
	}

	// Throughput per user
	if dataPoint.ActiveUsers > 0 {
		dataPoint.Features["throughput_per_user"] = dataPoint.Throughput / float64(dataPoint.ActiveUsers)
	}

	// Data volume per request
	if dataPoint.Throughput > 0 {
		dataPoint.Features["data_per_request"] = float64(dataPoint.DataVolume) / dataPoint.Throughput
	}
}

// runPredictions runs performance predictions
func (pa *PredictiveAnalytics) runPredictions(ctx context.Context) {
	ticker := time.NewTicker(pa.config.PredictionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pa.stopChannel:
			return
		case <-ticker.C:
			pa.generatePredictions()
		}
	}
}

// generatePredictions generates predictions for all configured metrics and horizons
func (pa *PredictiveAnalytics) generatePredictions() {
	pa.mu.RLock()
	if len(pa.historicalData) < 50 {
		pa.mu.RUnlock()
		pa.logger.Warn("Insufficient historical data for predictions")
		return
	}

	// Get latest data point for features
	latestPoint := pa.historicalData[len(pa.historicalData)-1]
	pa.mu.RUnlock()

	// Generate predictions for each metric and horizon
	metrics := []string{"response_time", "success_rate", "throughput", "error_rate", "cpu_usage", "memory_usage"}

	for _, metric := range metrics {
		for _, horizon := range pa.config.PredictionHorizons {
			prediction, err := pa.predictMetric(metric, latestPoint, horizon)
			if err != nil {
				pa.logger.Error("Failed to generate prediction",
					zap.String("metric", metric),
					zap.Duration("horizon", horizon),
					zap.Error(err))
				continue
			}

			// Store prediction
			pa.storePrediction(prediction)

			// Check for predictive alerts
			if pa.config.EnablePredictiveAlerts {
				pa.checkPredictiveAlerts(prediction)
			}
		}
	}
}

// predictMetric predicts a specific metric
func (pa *PredictiveAnalytics) predictMetric(metric string, latestPoint *PerformanceDataPoint, horizon time.Duration) (*PredictionResult, error) {
	// Select best model for this metric
	model := pa.selectBestModel(metric)
	if model == nil {
		return nil, fmt.Errorf("no suitable model found for metric %s", metric)
	}

	// Prepare features
	features := pa.prepareFeatures(latestPoint, metric)

	// Generate prediction
	prediction, err := model.Predict(features, horizon)
	if err != nil {
		return nil, fmt.Errorf("model prediction failed: %w", err)
	}

	// Enhance prediction with additional analysis
	pa.enhancePrediction(prediction, latestPoint, metric, horizon)

	return prediction, nil
}

// selectBestModel selects the best model for a given metric
func (pa *PredictiveAnalytics) selectBestModel(metric string) PredictionModel {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	var bestModel PredictionModel
	bestAccuracy := 0.0

	for _, model := range pa.models {
		if !model.IsTrained() {
			continue
		}

		accuracy := model.GetAccuracy()
		if accuracy > bestAccuracy {
			bestAccuracy = accuracy
			bestModel = model
		}
	}

	return bestModel
}

// prepareFeatures prepares features for prediction
func (pa *PredictiveAnalytics) prepareFeatures(dataPoint *PerformanceDataPoint, metric string) map[string]float64 {
	features := make(map[string]float64)

	// Copy existing features
	for k, v := range dataPoint.Features {
		features[k] = v
	}

	// Add metric-specific features
	switch metric {
	case "response_time":
		features["current_response_time"] = float64(dataPoint.ResponseTime.Milliseconds())
		features["response_time_ma_5min"] = dataPoint.Features["response_time_ma_5min"]
		features["response_time_ma_1hour"] = dataPoint.Features["response_time_ma_1hour"]
		features["response_time_trend"] = dataPoint.Features["response_time_trend"]
		features["response_time_volatility"] = dataPoint.Features["response_time_volatility"]

	case "success_rate":
		features["current_success_rate"] = dataPoint.SuccessRate
		features["current_error_rate"] = dataPoint.ErrorRate
		features["error_success_ratio"] = dataPoint.Features["error_success_ratio"]

	case "throughput":
		features["current_throughput"] = dataPoint.Throughput
		features["active_users"] = float64(dataPoint.ActiveUsers)
		features["throughput_per_user"] = dataPoint.Features["throughput_per_user"]

	case "cpu_usage":
		features["current_cpu_usage"] = dataPoint.CPUUsage
		features["current_memory_usage"] = dataPoint.MemoryUsage
		features["cpu_memory_ratio"] = dataPoint.Features["cpu_memory_ratio"]

	case "memory_usage":
		features["current_memory_usage"] = dataPoint.MemoryUsage
		features["current_cpu_usage"] = dataPoint.CPUUsage
		features["cpu_memory_ratio"] = dataPoint.Features["cpu_memory_ratio"]
	}

	return features
}

// enhancePrediction enhances a prediction with additional analysis
func (pa *PredictiveAnalytics) enhancePrediction(prediction *PredictionResult, dataPoint *PerformanceDataPoint, metric string, horizon time.Duration) {
	// Calculate trend analysis
	pa.calculatePredictionTrend(prediction, dataPoint, metric)

	// Calculate seasonality
	pa.calculatePredictionSeasonality(prediction, dataPoint)

	// Calculate contributing factors
	pa.calculateContributingFactors(prediction, dataPoint, metric)

	// Calculate uncertainty bounds
	pa.calculateUncertaintyBounds(prediction, dataPoint, metric)
}

// calculatePredictionTrend calculates trend analysis for prediction
func (pa *PredictiveAnalytics) calculatePredictionTrend(prediction *PredictionResult, dataPoint *PerformanceDataPoint, metric string) {
	var currentValue float64

	switch metric {
	case "response_time":
		currentValue = float64(dataPoint.ResponseTime.Milliseconds())
	case "success_rate":
		currentValue = dataPoint.SuccessRate
	case "throughput":
		currentValue = dataPoint.Throughput
	case "error_rate":
		currentValue = dataPoint.ErrorRate
	case "cpu_usage":
		currentValue = dataPoint.CPUUsage
	case "memory_usage":
		currentValue = dataPoint.MemoryUsage
	default:
		return
	}

	// Calculate trend
	change := prediction.PredictedValue - currentValue
	changePercent := (change / currentValue) * 100

	if changePercent > 10 {
		prediction.Trend = "degrading"
		prediction.TrendStrength = math.Min(math.Abs(changePercent)/50.0, 1.0)
	} else if changePercent < -10 {
		prediction.Trend = "improving"
		prediction.TrendStrength = math.Min(math.Abs(changePercent)/50.0, 1.0)
	} else {
		prediction.Trend = "stable"
		prediction.TrendStrength = 0.1
	}
}

// calculatePredictionSeasonality calculates seasonality for prediction
func (pa *PredictiveAnalytics) calculatePredictionSeasonality(prediction *PredictionResult, dataPoint *PerformanceDataPoint) {
	// Check for hour-of-day seasonality
	hour := dataPoint.Timestamp.Hour()
	if hour >= 9 && hour <= 17 {
		prediction.Seasonality = true
	} else {
		prediction.Seasonality = false
	}
}

// calculateContributingFactors calculates factors contributing to the prediction
func (pa *PredictiveAnalytics) calculateContributingFactors(prediction *PredictionResult, dataPoint *PerformanceDataPoint, metric string) {
	factors := make([]PredictionFactor, 0)

	switch metric {
	case "response_time":
		if dataPoint.Features["response_time_trend"] > 0 {
			factors = append(factors, PredictionFactor{
				Name:        "increasing_trend",
				Impact:      0.3,
				Confidence:  0.8,
				Description: "Response time shows an increasing trend",
			})
		}

		if dataPoint.CPUUsage > 80 {
			factors = append(factors, PredictionFactor{
				Name:        "high_cpu_usage",
				Impact:      0.4,
				Confidence:  0.9,
				Description: "High CPU usage may impact response times",
			})
		}

		if dataPoint.Features["response_time_volatility"] > 100 {
			factors = append(factors, PredictionFactor{
				Name:        "high_volatility",
				Impact:      0.2,
				Confidence:  0.7,
				Description: "High response time volatility indicates instability",
			})
		}

	case "success_rate":
		if dataPoint.ErrorRate > 0.05 {
			factors = append(factors, PredictionFactor{
				Name:        "high_error_rate",
				Impact:      -0.5,
				Confidence:  0.9,
				Description: "High error rate may indicate system issues",
			})
		}

		if dataPoint.MemoryUsage > 90 {
			factors = append(factors, PredictionFactor{
				Name:        "high_memory_usage",
				Impact:      -0.3,
				Confidence:  0.8,
				Description: "High memory usage may cause failures",
			})
		}
	}

	prediction.Factors = factors
}

// calculateUncertaintyBounds calculates uncertainty bounds for prediction
func (pa *PredictiveAnalytics) calculateUncertaintyBounds(prediction *PredictionResult, dataPoint *PerformanceDataPoint, metric string) {
	// Calculate standard deviation based on historical volatility
	var stdDev float64

	switch metric {
	case "response_time":
		stdDev = dataPoint.Features["response_time_volatility"]
	case "success_rate":
		stdDev = 0.05 // 5% standard deviation for success rate
	case "throughput":
		stdDev = dataPoint.Throughput * 0.1 // 10% of current throughput
	default:
		stdDev = prediction.PredictedValue * 0.1 // 10% of predicted value
	}

	// Calculate confidence intervals
	zScore := 1.96 // 95% confidence interval
	margin := zScore * stdDev

	prediction.LowerBound = prediction.PredictedValue - margin
	prediction.UpperBound = prediction.PredictedValue + margin
	prediction.StdDev = stdDev
}

// storePrediction stores a prediction result
func (pa *PredictiveAnalytics) storePrediction(prediction *PredictionResult) {
	// In a real implementation, this would store to a database
	// For now, we'll just log it
	pa.logger.Info("Generated prediction",
		zap.String("metric", prediction.Metric),
		zap.Float64("predicted_value", prediction.PredictedValue),
		zap.Float64("confidence", prediction.Confidence),
		zap.Duration("horizon", prediction.PredictionHorizon),
		zap.String("trend", prediction.Trend))
}

// checkPredictiveAlerts checks for predictive alerts
func (pa *PredictiveAnalytics) checkPredictiveAlerts(prediction *PredictionResult) {
	threshold, exists := pa.config.AlertThresholds[prediction.Metric]
	if !exists {
		return
	}

	var shouldAlert bool
	var alertType string

	// Check for threshold breach
	if prediction.PredictedValue > threshold {
		shouldAlert = true
		alertType = "threshold_breach"
	}

	// Check for significant trend change
	if prediction.TrendStrength > 0.7 && prediction.Trend == "degrading" {
		shouldAlert = true
		alertType = "trend_change"
	}

	if shouldAlert {
		alert := &PredictionAlert{
			ID:             fmt.Sprintf("pred_alert_%d", time.Now().UnixNano()),
			Metric:         prediction.Metric,
			AlertType:      alertType,
			Severity:       "warning",
			Message:        fmt.Sprintf("Predicted %s will reach %.2f in %v", prediction.Metric, prediction.PredictedValue, prediction.PredictionHorizon),
			PredictedValue: prediction.PredictedValue,
			Threshold:      threshold,
			Horizon:        prediction.PredictionHorizon,
			Confidence:     prediction.Confidence,
			Timestamp:      time.Now().UTC(),
		}

		pa.logger.Warn("Predictive alert triggered",
			zap.String("metric", alert.Metric),
			zap.String("alert_type", alert.AlertType),
			zap.Float64("predicted_value", alert.PredictedValue),
			zap.Float64("threshold", alert.Threshold))
	}
}

// retrainModels retrains prediction models
func (pa *PredictiveAnalytics) retrainModels(ctx context.Context) {
	ticker := time.NewTicker(pa.config.RetrainInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-pa.stopChannel:
			return
		case <-ticker.C:
			pa.retrainAllModels()
		}
	}
}

// retrainAllModels retrains all prediction models
func (pa *PredictiveAnalytics) retrainAllModels() {
	pa.mu.RLock()
	if len(pa.historicalData) < 100 {
		pa.mu.RUnlock()
		pa.logger.Warn("Insufficient data for model retraining")
		return
	}

	trainingData := make([]*PerformanceDataPoint, len(pa.historicalData))
	copy(trainingData, pa.historicalData)
	pa.mu.RUnlock()

	for _, model := range pa.models {
		if err := model.Train(trainingData); err != nil {
			pa.logger.Error("Failed to retrain model",
				zap.String("model", model.Name()),
				zap.Error(err))
		} else {
			pa.logger.Info("Model retrained successfully",
				zap.String("model", model.Name()),
				zap.Float64("accuracy", model.GetAccuracy()))
		}
	}
}

// monitorPredictiveAlerts monitors for predictive alerts
func (pa *PredictiveAnalytics) monitorPredictiveAlerts(ctx context.Context) {
	// This would integrate with the alerting system
	// For now, it's a placeholder
}

// initializeModels initializes prediction models
func (pa *PredictiveAnalytics) initializeModels() {
	// Initialize different types of models
	for _, modelType := range pa.config.ModelTypes {
		switch modelType {
		case "linear":
			pa.models["linear_response_time"] = NewLinearModel("response_time")
			pa.models["linear_success_rate"] = NewLinearModel("success_rate")
			pa.models["linear_throughput"] = NewLinearModel("throughput")

		case "exponential":
			pa.models["exponential_response_time"] = NewExponentialModel("response_time")
			pa.models["exponential_success_rate"] = NewExponentialModel("success_rate")

		case "arima":
			pa.models["arima_response_time"] = NewARIMAModel("response_time")
			pa.models["arima_throughput"] = NewARIMAModel("throughput")

		case "ensemble":
			pa.models["ensemble_response_time"] = NewEnsembleModel("response_time")
			pa.models["ensemble_success_rate"] = NewEnsembleModel("success_rate")
		}
	}
}

// GetPredictions returns current predictions
func (pa *PredictiveAnalytics) GetPredictions() []*PredictionResult {
	// In a real implementation, this would return stored predictions
	// For now, return empty slice
	return []*PredictionResult{}
}

// GetPredictionAccuracy returns prediction accuracy metrics
func (pa *PredictiveAnalytics) GetPredictionAccuracy() map[string]float64 {
	accuracies := make(map[string]float64)

	pa.mu.RLock()
	for name, model := range pa.models {
		accuracies[name] = model.GetAccuracy()
	}
	pa.mu.RUnlock()

	return accuracies
}

// GetHistoricalData returns historical data points
func (pa *PredictiveAnalytics) GetHistoricalData() []*PerformanceDataPoint {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	data := make([]*PerformanceDataPoint, len(pa.historicalData))
	copy(data, pa.historicalData)
	return data
}
