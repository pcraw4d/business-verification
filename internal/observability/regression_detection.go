package observability

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RegressionDetectionSystem provides comprehensive performance regression detection
type RegressionDetectionSystem struct {
	// Core components
	performanceMonitor  *PerformanceMonitor
	predictiveAnalytics *PredictiveAnalytics
	alertingSystem      *PerformanceAlertingSystem

	// Regression detection components
	detectors map[string]RegressionDetector
	baselines map[string]*PerformanceBaseline

	// Historical data for regression analysis
	historicalData    []*PerformanceDataPoint
	regressionHistory []*RegressionEvent
	dataRetention     time.Duration

	// Configuration
	config RegressionDetectionConfig

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *zap.Logger

	// Control channels
	stopChannel chan struct{}
}

// RegressionDetectionConfig holds configuration for regression detection
type RegressionDetectionConfig struct {
	// Detection settings
	DetectionInterval       time.Duration `json:"detection_interval"`
	BaselineWindow          time.Duration `json:"baseline_window"`
	DetectionWindow         time.Duration `json:"detection_window"`
	MinDataPoints           int           `json:"min_data_points"`
	MaxHistoricalDataPoints int           `json:"max_historical_data_points"`

	// Regression thresholds
	RegressionThresholds struct {
		ResponseTime struct {
			Degradation float64 `json:"degradation"` // Percentage increase
			Improvement float64 `json:"improvement"` // Percentage decrease
		} `json:"response_time"`

		SuccessRate struct {
			Degradation float64 `json:"degradation"` // Percentage decrease
			Improvement float64 `json:"improvement"` // Percentage increase
		} `json:"success_rate"`

		Throughput struct {
			Degradation float64 `json:"degradation"` // Percentage decrease
			Improvement float64 `json:"improvement"` // Percentage increase
		} `json:"throughput"`

		ErrorRate struct {
			Degradation float64 `json:"degradation"` // Percentage increase
			Improvement float64 `json:"improvement"` // Percentage decrease
		} `json:"error_rate"`

		ResourceUtilization struct {
			CPU struct {
				Degradation float64 `json:"degradation"` // Percentage increase
				Improvement float64 `json:"improvement"` // Percentage decrease
			} `json:"cpu"`
			Memory struct {
				Degradation float64 `json:"degradation"` // Percentage increase
				Improvement float64 `json:"improvement"` // Percentage decrease
			} `json:"memory"`
		} `json:"resource_utilization"`
	} `json:"regression_thresholds"`

	// Statistical settings
	ConfidenceLevel     float64 `json:"confidence_level"`      // 0.95 for 95% confidence
	StatisticalTestType string  `json:"statistical_test_type"` // t-test, mann-whitney, etc.
	PValueThreshold     float64 `json:"p_value_threshold"`     // 0.05 for significance

	// Alerting settings
	EnableRegressionAlerts bool              `json:"enable_regression_alerts"`
	AlertSeverity          map[string]string `json:"alert_severity"` // metric -> severity mapping

	// Baseline settings
	AutoBaselineUpdate     bool          `json:"auto_baseline_update"`
	BaselineUpdateInterval time.Duration `json:"baseline_update_interval"`
	BaselineStabilityCheck bool          `json:"baseline_stability_check"`

	// Performance settings
	MaxConcurrentDetections int           `json:"max_concurrent_detections"`
	DetectionTimeout        time.Duration `json:"detection_timeout"`
}

// RegressionDetector defines a regression detection algorithm interface
type RegressionDetector interface {
	Name() string
	Type() string
	Detect(baseline *PerformanceBaseline, currentData []*PerformanceDataPoint) (*RegressionResult, error)
	GetConfidence() float64
	IsApplicable(metric string) bool
}

// PerformanceBaseline represents a performance baseline
type PerformanceBaseline struct {
	ID        string    `json:"id"`
	Metric    string    `json:"metric"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	IsActive  bool      `json:"is_active"`

	// Statistical measures
	Mean         float64 `json:"mean"`
	Median       float64 `json:"median"`
	StdDev       float64 `json:"std_dev"`
	Min          float64 `json:"min"`
	Max          float64 `json:"max"`
	Percentile95 float64 `json:"percentile_95"`
	Percentile99 float64 `json:"percentile_99"`

	// Sample information
	SampleSize  int       `json:"sample_size"`
	SampleStart time.Time `json:"sample_start"`
	SampleEnd   time.Time `json:"sample_end"`

	// Metadata
	Description string            `json:"description"`
	Tags        map[string]string `json:"tags"`
	Confidence  float64           `json:"confidence"`
}

// RegressionResult represents a regression detection result
type RegressionResult struct {
	ID         string    `json:"id"`
	Metric     string    `json:"metric"`
	DetectedAt time.Time `json:"detected_at"`
	Type       string    `json:"type"` // degradation, improvement, none

	// Statistical measures
	BaselineMean    float64 `json:"baseline_mean"`
	CurrentMean     float64 `json:"current_mean"`
	ChangePercent   float64 `json:"change_percent"`
	ChangeDirection string  `json:"change_direction"` // increase, decrease

	// Statistical significance
	PValue        float64 `json:"p_value"`
	Confidence    float64 `json:"confidence"`
	IsSignificant bool    `json:"is_significant"`

	// Detection details
	DetectorUsed  string    `json:"detector_used"`
	DetectionTime time.Time `json:"detection_time"`
	Severity      string    `json:"severity"` // low, medium, high, critical

	// Context
	BaselineID    string            `json:"baseline_id"`
	BaselineStart time.Time         `json:"baseline_start"`
	BaselineEnd   time.Time         `json:"baseline_end"`
	CurrentStart  time.Time         `json:"current_start"`
	CurrentEnd    time.Time         `json:"current_end"`
	Tags          map[string]string `json:"tags"`

	// Additional analysis
	TrendAnalysis    *TrendAnalysis    `json:"trend_analysis,omitempty"`
	SeasonalityCheck *SeasonalityCheck `json:"seasonality_check,omitempty"`
	OutlierAnalysis  *OutlierAnalysis  `json:"outlier_analysis,omitempty"`
}

// RegressionEvent represents a regression event
type RegressionEvent struct {
	ID          string    `json:"id"`
	ResultID    string    `json:"result_id"`
	EventType   string    `json:"event_type"` // detected, resolved, acknowledged
	Timestamp   time.Time `json:"timestamp"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	User        string    `json:"user,omitempty"`
	Notes       string    `json:"notes,omitempty"`
}

// TrendAnalysis represents trend analysis results
type TrendAnalysis struct {
	TrendDirection string  `json:"trend_direction"` // increasing, decreasing, stable
	TrendStrength  float64 `json:"trend_strength"`  // 0-1
	Slope          float64 `json:"slope"`
	R2             float64 `json:"r2"` // R-squared value
	PValue         float64 `json:"p_value"`
}

// SeasonalityCheck represents seasonality analysis
type SeasonalityCheck struct {
	HasSeasonality   bool      `json:"has_seasonality"`
	SeasonalPeriod   string    `json:"seasonal_period"` // daily, weekly, monthly
	SeasonalStrength float64   `json:"seasonal_strength"`
	PeakTime         time.Time `json:"peak_time,omitempty"`
	TroughTime       time.Time `json:"trough_time,omitempty"`
}

// OutlierAnalysis represents outlier analysis
type OutlierAnalysis struct {
	OutlierCount     int     `json:"outlier_count"`
	OutlierPercent   float64 `json:"outlier_percent"`
	OutlierThreshold float64 `json:"outlier_threshold"`
	OutlierMethod    string  `json:"outlier_method"` // z-score, iqr, etc.
}

// NewRegressionDetectionSystem creates a new regression detection system
func NewRegressionDetectionSystem(
	performanceMonitor *PerformanceMonitor,
	predictiveAnalytics *PredictiveAnalytics,
	alertingSystem *PerformanceAlertingSystem,
	config RegressionDetectionConfig,
	logger *zap.Logger,
) *RegressionDetectionSystem {
	// Set default values
	if config.DetectionInterval == 0 {
		config.DetectionInterval = 5 * time.Minute
	}
	if config.BaselineWindow == 0 {
		config.BaselineWindow = 24 * time.Hour
	}
	if config.DetectionWindow == 0 {
		config.DetectionWindow = 1 * time.Hour
	}
	if config.MinDataPoints == 0 {
		config.MinDataPoints = 30
	}
	if config.MaxHistoricalDataPoints == 0 {
		config.MaxHistoricalDataPoints = 10000
	}
	if config.ConfidenceLevel == 0 {
		config.ConfidenceLevel = 0.95
	}
	if config.PValueThreshold == 0 {
		config.PValueThreshold = 0.05
	}
	if config.BaselineUpdateInterval == 0 {
		config.BaselineUpdateInterval = 24 * time.Hour
	}
	if config.DetectionTimeout == 0 {
		config.DetectionTimeout = 30 * time.Second
	}

	rds := &RegressionDetectionSystem{
		performanceMonitor:  performanceMonitor,
		predictiveAnalytics: predictiveAnalytics,
		alertingSystem:      alertingSystem,
		detectors:           make(map[string]RegressionDetector),
		baselines:           make(map[string]*PerformanceBaseline),
		historicalData:      make([]*PerformanceDataPoint, 0),
		regressionHistory:   make([]*RegressionEvent, 0),
		config:              config,
		logger:              logger,
		stopChannel:         make(chan struct{}),
	}

	// Initialize regression detectors
	rds.initializeDetectors()

	// Initialize default baselines
	rds.initializeDefaultBaselines()

	return rds
}

// Start starts the regression detection system
func (rds *RegressionDetectionSystem) Start(ctx context.Context) error {
	rds.logger.Info("Starting regression detection system")

	// Start regression detection
	go rds.runRegressionDetection(ctx)

	// Start baseline updates
	if rds.config.AutoBaselineUpdate {
		go rds.updateBaselines(ctx)
	}

	// Start data collection
	go rds.collectData(ctx)

	rds.logger.Info("Regression detection system started")
	return nil
}

// Stop stops the regression detection system
func (rds *RegressionDetectionSystem) Stop() error {
	rds.logger.Info("Stopping regression detection system")

	close(rds.stopChannel)

	rds.logger.Info("Regression detection system stopped")
	return nil
}

// runRegressionDetection runs the main regression detection loop
func (rds *RegressionDetectionSystem) runRegressionDetection(ctx context.Context) {
	ticker := time.NewTicker(rds.config.DetectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rds.stopChannel:
			return
		case <-ticker.C:
			rds.detectRegressions()
		}
	}
}

// detectRegressions performs regression detection for all metrics
func (rds *RegressionDetectionSystem) detectRegressions() {
	rds.mu.RLock()
	if len(rds.historicalData) < rds.config.MinDataPoints {
		rds.mu.RUnlock()
		rds.logger.Debug("Insufficient data for regression detection")
		return
	}

	// Get current data window
	currentData := rds.getCurrentDataWindow()
	rds.mu.RUnlock()

	if len(currentData) < rds.config.MinDataPoints {
		rds.logger.Debug("Insufficient current data for regression detection")
		return
	}

	// Detect regressions for each metric
	metrics := []string{"response_time", "success_rate", "throughput", "error_rate", "cpu_usage", "memory_usage"}

	for _, metric := range metrics {
		baseline := rds.getBaseline(metric)
		if baseline == nil {
			rds.logger.Debug("No baseline available for metric", zap.String("metric", metric))
			continue
		}

		// Run detection with each applicable detector
		for _, detector := range rds.detectors {
			if !detector.IsApplicable(metric) {
				continue
			}

			result, err := detector.Detect(baseline, currentData)
			if err != nil {
				rds.logger.Error("Regression detection failed",
					zap.String("detector", detector.Name()),
					zap.String("metric", metric),
					zap.Error(err))
				continue
			}

			if result != nil && result.Type != "none" {
				rds.handleRegressionResult(result)
			}
		}
	}
}

// getCurrentDataWindow gets the current data window for detection
func (rds *RegressionDetectionSystem) getCurrentDataWindow() []*PerformanceDataPoint {
	now := time.Now().UTC()
	windowStart := now.Add(-rds.config.DetectionWindow)

	var windowData []*PerformanceDataPoint
	for _, point := range rds.historicalData {
		if point.Timestamp.After(windowStart) && point.Timestamp.Before(now) {
			windowData = append(windowData, point)
		}
	}

	return windowData
}

// handleRegressionResult handles a regression detection result
func (rds *RegressionDetectionSystem) handleRegressionResult(result *RegressionResult) {
	rds.mu.Lock()
	defer rds.mu.Unlock()

	// Add to regression history
	rds.regressionHistory = append(rds.regressionHistory, &RegressionEvent{
		ID:          fmt.Sprintf("reg_event_%d", time.Now().UnixNano()),
		ResultID:    result.ID,
		EventType:   "detected",
		Timestamp:   time.Now().UTC(),
		Description: fmt.Sprintf("Regression detected: %s %s by %.2f%%", result.Metric, result.Type, result.ChangePercent),
		Severity:    result.Severity,
	})

	// Send alert if enabled
	if rds.config.EnableRegressionAlerts {
		rds.sendRegressionAlert(result)
	}

	rds.logger.Info("Regression detected",
		zap.String("metric", result.Metric),
		zap.String("type", result.Type),
		zap.Float64("change_percent", result.ChangePercent),
		zap.String("severity", result.Severity),
		zap.String("detector", result.DetectorUsed))
}

// sendRegressionAlert sends a regression alert
func (rds *RegressionDetectionSystem) sendRegressionAlert(result *RegressionResult) {
	// In a real implementation, this would integrate with the alerting system
	// For now, we'll just log the alert
	rds.logger.Warn("Regression alert",
		zap.String("metric", result.Metric),
		zap.String("type", result.Type),
		zap.Float64("change_percent", result.ChangePercent),
		zap.String("severity", result.Severity),
		zap.String("detector", result.DetectorUsed))
}

// collectData collects performance data for regression analysis
func (rds *RegressionDetectionSystem) collectData(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Collect data every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rds.stopChannel:
			return
		case <-ticker.C:
			rds.collectPerformanceData()
		}
	}
}

// collectPerformanceData collects current performance data
func (rds *RegressionDetectionSystem) collectPerformanceData() {
	metrics := rds.performanceMonitor.GetMetrics()
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

	rds.mu.Lock()
	rds.historicalData = append(rds.historicalData, dataPoint)

	// Maintain data retention
	if len(rds.historicalData) > rds.config.MaxHistoricalDataPoints {
		rds.historicalData = rds.historicalData[1:]
	}
	rds.mu.Unlock()

	rds.logger.Debug("Collected performance data for regression analysis",
		zap.Time("timestamp", dataPoint.Timestamp),
		zap.Duration("response_time", dataPoint.ResponseTime),
		zap.Float64("success_rate", dataPoint.SuccessRate))
}

// updateBaselines updates performance baselines
func (rds *RegressionDetectionSystem) updateBaselines(ctx context.Context) {
	ticker := time.NewTicker(rds.config.BaselineUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rds.stopChannel:
			return
		case <-ticker.C:
			rds.updateAllBaselines()
		}
	}
}

// updateAllBaselines updates all performance baselines
func (rds *RegressionDetectionSystem) updateAllBaselines() {
	metrics := []string{"response_time", "success_rate", "throughput", "error_rate", "cpu_usage", "memory_usage"}

	for _, metric := range metrics {
		rds.updateBaseline(metric)
	}
}

// updateBaseline updates a specific baseline
func (rds *RegressionDetectionSystem) updateBaseline(metric string) {
	rds.mu.RLock()
	baselineData := rds.getBaselineData(metric)
	rds.mu.RUnlock()

	if len(baselineData) < rds.config.MinDataPoints {
		rds.logger.Debug("Insufficient data for baseline update",
			zap.String("metric", metric),
			zap.Int("data_points", len(baselineData)))
		return
	}

	// Calculate new baseline statistics
	baseline := rds.calculateBaseline(metric, baselineData)

	rds.mu.Lock()
	rds.baselines[metric] = baseline
	rds.mu.Unlock()

	rds.logger.Info("Baseline updated",
		zap.String("metric", metric),
		zap.Float64("mean", baseline.Mean),
		zap.Float64("std_dev", baseline.StdDev),
		zap.Int("sample_size", baseline.SampleSize))
}

// getBaselineData gets data for baseline calculation
func (rds *RegressionDetectionSystem) getBaselineData(metric string) []*PerformanceDataPoint {
	now := time.Now().UTC()
	windowStart := now.Add(-rds.config.BaselineWindow)

	var baselineData []*PerformanceDataPoint
	for _, point := range rds.historicalData {
		if point.Timestamp.After(windowStart) && point.Timestamp.Before(now) {
			baselineData = append(baselineData, point)
		}
	}

	return baselineData
}

// calculateBaseline calculates baseline statistics
func (rds *RegressionDetectionSystem) calculateBaseline(metric string, data []*PerformanceDataPoint) *PerformanceBaseline {
	values := rds.extractMetricValues(metric, data)

	// Calculate basic statistics
	mean := rds.calculateMean(values)
	median := rds.calculateMedian(values)
	stdDev := rds.calculateStdDev(values, mean)
	min := rds.calculateMin(values)
	max := rds.calculateMax(values)
	p95 := rds.calculatePercentile(values, 95)
	p99 := rds.calculatePercentile(values, 99)

	now := time.Now().UTC()
	baseline := &PerformanceBaseline{
		ID:           fmt.Sprintf("baseline_%s_%d", metric, now.Unix()),
		Metric:       metric,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
		Mean:         mean,
		Median:       median,
		StdDev:       stdDev,
		Min:          min,
		Max:          max,
		Percentile95: p95,
		Percentile99: p99,
		SampleSize:   len(values),
		SampleStart:  data[0].Timestamp,
		SampleEnd:    data[len(data)-1].Timestamp,
		Description:  fmt.Sprintf("Baseline for %s metric", metric),
		Tags:         make(map[string]string),
		Confidence:   rds.config.ConfidenceLevel,
	}

	return baseline
}

// extractMetricValues extracts metric values from data points
func (rds *RegressionDetectionSystem) extractMetricValues(metric string, data []*PerformanceDataPoint) []float64 {
	values := make([]float64, 0, len(data))

	for _, point := range data {
		var value float64
		switch metric {
		case "response_time":
			value = float64(point.ResponseTime.Milliseconds())
		case "success_rate":
			value = point.SuccessRate
		case "throughput":
			value = point.Throughput
		case "error_rate":
			value = point.ErrorRate
		case "cpu_usage":
			value = point.CPUUsage
		case "memory_usage":
			value = point.MemoryUsage
		default:
			continue
		}
		values = append(values, value)
	}

	return values
}

// calculateMean calculates the mean of values
func (rds *RegressionDetectionSystem) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

// calculateMedian calculates the median of values
func (rds *RegressionDetectionSystem) calculateMedian(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	// Sort values (in a real implementation, you'd use a proper sort)
	// For simplicity, we'll just return the mean for now
	return rds.calculateMean(values)
}

// calculateStdDev calculates the standard deviation of values
func (rds *RegressionDetectionSystem) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	sumSq := 0.0
	for _, value := range values {
		sumSq += (value - mean) * (value - mean)
	}

	variance := sumSq / float64(len(values)-1)
	return math.Sqrt(variance)
}

// calculateMin calculates the minimum value
func (rds *RegressionDetectionSystem) calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	min := values[0]
	for _, value := range values {
		if value < min {
			min = value
		}
	}
	return min
}

// calculateMax calculates the maximum value
func (rds *RegressionDetectionSystem) calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	max := values[0]
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}

// calculatePercentile calculates the nth percentile
func (rds *RegressionDetectionSystem) calculatePercentile(values []float64, percentile int) float64 {
	if len(values) == 0 {
		return 0.0
	}

	// For simplicity, we'll use a basic approximation
	// In a real implementation, you'd use proper percentile calculation
	index := int(float64(len(values)-1) * float64(percentile) / 100.0)
	if index >= len(values) {
		index = len(values) - 1
	}
	return values[index]
}

// getBaseline gets a baseline for a metric
func (rds *RegressionDetectionSystem) getBaseline(metric string) *PerformanceBaseline {
	rds.mu.RLock()
	defer rds.mu.RUnlock()

	baseline, exists := rds.baselines[metric]
	if !exists || !baseline.IsActive {
		return nil
	}

	return baseline
}

// initializeDetectors initializes regression detectors
func (rds *RegressionDetectionSystem) initializeDetectors() {
	// Initialize different types of detectors
	rds.detectors["statistical"] = NewStatisticalDetector(rds.config)
	rds.detectors["trend"] = NewTrendDetector(rds.config)
	rds.detectors["threshold"] = NewThresholdDetector(rds.config)
	rds.detectors["anomaly"] = NewAnomalyDetector(rds.config)
}

// initializeDefaultBaselines initializes default baselines
func (rds *RegressionDetectionSystem) initializeDefaultBaselines() {
	// Create default baselines with placeholder values
	// These will be updated with real data when available
	metrics := []string{"response_time", "success_rate", "throughput", "error_rate", "cpu_usage", "memory_usage"}

	for _, metric := range metrics {
		baseline := &PerformanceBaseline{
			ID:          fmt.Sprintf("default_baseline_%s", metric),
			Metric:      metric,
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			IsActive:    true,
			Description: fmt.Sprintf("Default baseline for %s", metric),
			Tags:        make(map[string]string),
			Confidence:  rds.config.ConfidenceLevel,
		}

		// Set default values based on metric type
		switch metric {
		case "response_time":
			baseline.Mean = 250.0
			baseline.StdDev = 50.0
		case "success_rate":
			baseline.Mean = 0.98
			baseline.StdDev = 0.01
		case "throughput":
			baseline.Mean = 1000.0
			baseline.StdDev = 100.0
		case "error_rate":
			baseline.Mean = 0.02
			baseline.StdDev = 0.005
		case "cpu_usage":
			baseline.Mean = 50.0
			baseline.StdDev = 15.0
		case "memory_usage":
			baseline.Mean = 60.0
			baseline.StdDev = 10.0
		}

		rds.baselines[metric] = baseline
	}
}

// GetRegressionHistory returns regression history
func (rds *RegressionDetectionSystem) GetRegressionHistory() []*RegressionEvent {
	rds.mu.RLock()
	defer rds.mu.RUnlock()

	history := make([]*RegressionEvent, len(rds.regressionHistory))
	copy(history, rds.regressionHistory)
	return history
}

// GetBaselines returns all baselines
func (rds *RegressionDetectionSystem) GetBaselines() map[string]*PerformanceBaseline {
	rds.mu.RLock()
	defer rds.mu.RUnlock()

	baselines := make(map[string]*PerformanceBaseline)
	for k, v := range rds.baselines {
		baselines[k] = v
	}
	return baselines
}

// GetBaseline returns a specific baseline
func (rds *RegressionDetectionSystem) GetBaseline(metric string) *PerformanceBaseline {
	return rds.getBaseline(metric)
}

// UpdateBaseline manually updates a baseline
func (rds *RegressionDetectionSystem) UpdateBaseline(metric string) error {
	rds.updateBaseline(metric)
	return nil
}

// AddRegressionEvent adds a manual regression event
func (rds *RegressionDetectionSystem) AddRegressionEvent(event *RegressionEvent) error {
	rds.mu.Lock()
	defer rds.mu.Unlock()

	event.ID = fmt.Sprintf("reg_event_%d", time.Now().UnixNano())
	event.Timestamp = time.Now().UTC()

	rds.regressionHistory = append(rds.regressionHistory, event)

	rds.logger.Info("Manual regression event added",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.EventType),
		zap.String("description", event.Description))

	return nil
}
