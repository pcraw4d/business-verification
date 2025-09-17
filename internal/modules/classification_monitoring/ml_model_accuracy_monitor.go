package classification_monitoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MLModelAccuracyMonitor tracks ML model performance and detects drift
type MLModelAccuracyMonitor struct {
	config *MLModelMonitorConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// Model tracking
	models      map[string]*MLModelDetailedMetrics
	driftAlerts []*DriftAlert

	// Trend analysis
	trendAnalyzer      *MLTrendAnalyzer
	driftDetector      *DriftDetector
	performanceTracker *MLPerformanceTracker
}

// MLModelMonitorConfig configuration for ML model monitoring
type MLModelMonitorConfig struct {
	// Drift detection thresholds
	AccuracyDriftThreshold   float64
	ConfidenceDriftThreshold float64
	LatencyDriftThreshold    time.Duration

	// Trend analysis settings
	TrendWindowSize       int
	MinDataPointsForTrend int
	SignificanceLevel     float64

	// Alert settings
	AlertCooldownPeriod time.Duration
	MaxAlertsPerModel   int

	// Performance tracking
	PerformanceWindowSize int
	BaselineWindowSize    int
}

// LatencyDataPoint represents a latency measurement at a specific time
type LatencyDataPoint struct {
	Timestamp time.Time
	Value     time.Duration
}

// MLModelDetailedMetrics extends the existing MLModelMetrics with additional tracking
type MLModelDetailedMetrics struct {
	// Base metrics (reuse existing MLModelMetrics)
	*MLModelMetrics

	// Additional detailed tracking
	AverageLatency time.Duration

	// Trend data
	HistoricalAccuracy   []*AccuracyDataPoint
	HistoricalConfidence []*ConfidenceDataPoint
	HistoricalLatency    []*LatencyDataPoint

	// Drift indicators
	AccuracyDrift   float64
	ConfidenceDrift float64
	LatencyDrift    float64
	DriftStatus     string // "stable", "warning", "critical"

	// Performance indicators
	PerformanceTrend string // "improving", "stable", "degrading"
	ReliabilityScore float64
	StabilityScore   float64
}

// DriftAlert represents a drift detection alert
type DriftAlert struct {
	ID              string
	ModelName       string
	AlertType       string // "accuracy_drift", "confidence_drift", "latency_drift"
	Severity        string // "warning", "critical"
	Message         string
	CurrentValue    float64
	BaselineValue   float64
	DriftPercentage float64
	Timestamp       time.Time
	Resolved        bool
}

// MLTrendAnalyzer analyzes trends in ML model performance
type MLTrendAnalyzer struct {
	config *MLModelMonitorConfig
	logger *zap.Logger
}

// DriftDetector detects performance drift in ML models
type DriftDetector struct {
	config *MLModelMonitorConfig
	logger *zap.Logger
}

// MLPerformanceTracker tracks ML model performance over time
type MLPerformanceTracker struct {
	config *MLModelMonitorConfig
	logger *zap.Logger
}

// DefaultMLModelMonitorConfig returns default configuration
func DefaultMLModelMonitorConfig() *MLModelMonitorConfig {
	return &MLModelMonitorConfig{
		AccuracyDriftThreshold:   0.05, // 5% accuracy drop
		ConfidenceDriftThreshold: 0.10, // 10% confidence drop
		LatencyDriftThreshold:    2 * time.Second,
		TrendWindowSize:          100,
		MinDataPointsForTrend:    20,
		SignificanceLevel:        0.05,
		AlertCooldownPeriod:      30 * time.Minute,
		MaxAlertsPerModel:        10,
		PerformanceWindowSize:    50,
		BaselineWindowSize:       200,
	}
}

// NewMLModelAccuracyMonitor creates a new ML model accuracy monitor
func NewMLModelAccuracyMonitor(config *MLModelMonitorConfig, logger *zap.Logger) *MLModelAccuracyMonitor {
	if config == nil {
		config = DefaultMLModelMonitorConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &MLModelAccuracyMonitor{
		config:      config,
		logger:      logger,
		models:      make(map[string]*MLModelDetailedMetrics),
		driftAlerts: make([]*DriftAlert, 0),
		trendAnalyzer: &MLTrendAnalyzer{
			config: config,
			logger: logger,
		},
		driftDetector: &DriftDetector{
			config: config,
			logger: logger,
		},
		performanceTracker: &MLPerformanceTracker{
			config: config,
			logger: logger,
		},
	}
}

// TrackModelPrediction tracks a prediction result for an ML model
func (monitor *MLModelAccuracyMonitor) TrackModelPrediction(ctx context.Context, result *ClassificationResult) error {
	monitor.mu.Lock()
	defer monitor.mu.Unlock()

	// Extract model version from metadata or use default
	modelVersion := "v1.0" // Default version
	if version, exists := result.Metadata["model_version"]; exists {
		if v, ok := version.(string); ok {
			modelVersion = v
		}
	}

	modelKey := fmt.Sprintf("%s_%s", result.ClassificationMethod, modelVersion)

	// Get or create model metrics
	metrics, exists := monitor.models[modelKey]
	if !exists {
		metrics = monitor.createModelMetrics(result.ClassificationMethod, modelVersion)
		monitor.models[modelKey] = metrics
	}

	// Update metrics
	monitor.updateModelMetrics(metrics, result)

	// Analyze trends
	monitor.analyzeModelTrends(metrics)

	// Detect drift
	monitor.detectModelDrift(metrics)

	// Update performance tracking
	monitor.updatePerformanceTracking(metrics)

	return nil
}

// GetModelMetrics returns metrics for a specific model
func (monitor *MLModelAccuracyMonitor) GetModelMetrics(modelName, modelVersion string) *MLModelDetailedMetrics {
	monitor.mu.RLock()
	defer monitor.mu.RUnlock()

	modelKey := fmt.Sprintf("%s_%s", modelName, modelVersion)
	metrics, exists := monitor.models[modelKey]
	if !exists {
		return nil
	}

	return monitor.copyModelMetrics(metrics)
}

// GetAllModelMetrics returns metrics for all tracked models
func (monitor *MLModelAccuracyMonitor) GetAllModelMetrics() map[string]*MLModelDetailedMetrics {
	monitor.mu.RLock()
	defer monitor.mu.RUnlock()

	result := make(map[string]*MLModelDetailedMetrics)
	for key, metrics := range monitor.models {
		result[key] = monitor.copyModelMetrics(metrics)
	}

	return result
}

// GetDriftAlerts returns all drift alerts
func (monitor *MLModelAccuracyMonitor) GetDriftAlerts() []*DriftAlert {
	monitor.mu.RLock()
	defer monitor.mu.RUnlock()

	alerts := make([]*DriftAlert, len(monitor.driftAlerts))
	copy(alerts, monitor.driftAlerts)

	return alerts
}

// GetActiveDriftAlerts returns unresolved drift alerts
func (monitor *MLModelAccuracyMonitor) GetActiveDriftAlerts() []*DriftAlert {
	monitor.mu.RLock()
	defer monitor.mu.RUnlock()

	var activeAlerts []*DriftAlert
	for _, alert := range monitor.driftAlerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// createModelMetrics creates new model metrics
func (monitor *MLModelAccuracyMonitor) createModelMetrics(modelName, modelVersion string) *MLModelDetailedMetrics {
	baseMetrics := &MLModelMetrics{
		ModelName:        modelName,
		ModelVersion:     modelVersion,
		LastUpdated:      time.Now(),
		WindowedAccuracy: make([]float64, 0),
	}

	return &MLModelDetailedMetrics{
		MLModelMetrics:       baseMetrics,
		HistoricalAccuracy:   make([]*AccuracyDataPoint, 0),
		HistoricalConfidence: make([]*ConfidenceDataPoint, 0),
		HistoricalLatency:    make([]*LatencyDataPoint, 0),
		DriftStatus:          "stable",
		PerformanceTrend:     "stable",
	}
}

// updateModelMetrics updates model metrics with new prediction result
func (monitor *MLModelAccuracyMonitor) updateModelMetrics(metrics *MLModelDetailedMetrics, result *ClassificationResult) {
	metrics.LastUpdated = time.Now()
	metrics.TotalPredictions++

	// Update accuracy
	if result.IsCorrect != nil && *result.IsCorrect {
		metrics.CorrectPredictions++
	}
	metrics.AccuracyScore = float64(metrics.CorrectPredictions) / float64(metrics.TotalPredictions)

	// Update confidence
	metrics.AverageConfidence = (metrics.AverageConfidence*float64(metrics.TotalPredictions-1) + result.ConfidenceScore) / float64(metrics.TotalPredictions)

	// Update latency (approximate)
	predictionLatency := time.Since(result.Timestamp)
	metrics.AverageLatency = (metrics.AverageLatency*time.Duration(metrics.TotalPredictions-1) + predictionLatency) / time.Duration(metrics.TotalPredictions)

	// Add to historical data
	accuracyPoint := &AccuracyDataPoint{
		Timestamp:  result.Timestamp,
		Accuracy:   metrics.AccuracyScore,
		SampleSize: 1,
	}
	metrics.HistoricalAccuracy = append(metrics.HistoricalAccuracy, accuracyPoint)

	confidencePoint := &ConfidenceDataPoint{
		Timestamp:  result.Timestamp,
		Confidence: result.ConfidenceScore,
		SampleSize: 1,
	}
	metrics.HistoricalConfidence = append(metrics.HistoricalConfidence, confidencePoint)

	latencyPoint := &LatencyDataPoint{
		Timestamp: result.Timestamp,
		Value:     predictionLatency,
	}
	metrics.HistoricalLatency = append(metrics.HistoricalLatency, latencyPoint)

	// Maintain window size
	if len(metrics.HistoricalAccuracy) > monitor.config.TrendWindowSize {
		metrics.HistoricalAccuracy = metrics.HistoricalAccuracy[1:]
	}
	if len(metrics.HistoricalConfidence) > monitor.config.TrendWindowSize {
		metrics.HistoricalConfidence = metrics.HistoricalConfidence[1:]
	}
	if len(metrics.HistoricalLatency) > monitor.config.TrendWindowSize {
		metrics.HistoricalLatency = metrics.HistoricalLatency[1:]
	}
}

// analyzeModelTrends analyzes trends in model performance
func (monitor *MLModelAccuracyMonitor) analyzeModelTrends(metrics *MLModelDetailedMetrics) {
	if len(metrics.HistoricalAccuracy) < monitor.config.MinDataPointsForTrend {
		return
	}

	// Analyze accuracy trend
	accuracyTrend := monitor.trendAnalyzer.AnalyzeTrend(metrics.HistoricalAccuracy)

	// Analyze confidence trend
	confidenceTrend := monitor.trendAnalyzer.AnalyzeConfidenceTrend(metrics.HistoricalConfidence)

	// Analyze latency trend
	latencyTrend := monitor.trendAnalyzer.AnalyzeLatencyTrend(metrics.HistoricalLatency)

	// Determine overall performance trend
	monitor.determinePerformanceTrend(metrics, accuracyTrend, confidenceTrend, latencyTrend)
}

// detectModelDrift detects performance drift in the model
func (monitor *MLModelAccuracyMonitor) detectModelDrift(metrics *MLModelDetailedMetrics) {
	if len(metrics.HistoricalAccuracy) < monitor.config.BaselineWindowSize {
		return
	}

	// Detect accuracy drift
	accuracyDrift := monitor.driftDetector.DetectAccuracyDrift(metrics.HistoricalAccuracy)
	metrics.AccuracyDrift = accuracyDrift

	// Detect confidence drift
	confidenceDrift := monitor.driftDetector.DetectConfidenceDrift(metrics.HistoricalConfidence)
	metrics.ConfidenceDrift = confidenceDrift

	// Detect latency drift
	latencyDrift := monitor.driftDetector.DetectLatencyDrift(metrics.HistoricalLatency)
	metrics.LatencyDrift = latencyDrift

	// Determine drift status
	monitor.determineDriftStatus(metrics)

	// Create alerts if necessary
	monitor.createDriftAlerts(metrics)
}

// updatePerformanceTracking updates performance tracking metrics
func (monitor *MLModelAccuracyMonitor) updatePerformanceTracking(metrics *MLModelDetailedMetrics) {
	// Calculate reliability score
	metrics.ReliabilityScore = monitor.performanceTracker.CalculateReliabilityScore(metrics)

	// Calculate stability score
	metrics.StabilityScore = monitor.performanceTracker.CalculateStabilityScore(metrics)
}

// determinePerformanceTrend determines the overall performance trend
func (monitor *MLModelAccuracyMonitor) determinePerformanceTrend(metrics *MLModelDetailedMetrics, accuracyTrend, confidenceTrend, latencyTrend string) {
	improvingCount := 0
	degradingCount := 0

	if accuracyTrend == "improving" {
		improvingCount++
	} else if accuracyTrend == "degrading" {
		degradingCount++
	}

	if confidenceTrend == "improving" {
		improvingCount++
	} else if confidenceTrend == "degrading" {
		degradingCount++
	}

	if latencyTrend == "improving" {
		improvingCount++
	} else if latencyTrend == "degrading" {
		degradingCount++
	}

	if improvingCount > degradingCount {
		metrics.PerformanceTrend = "improving"
	} else if degradingCount > improvingCount {
		metrics.PerformanceTrend = "degrading"
	} else {
		metrics.PerformanceTrend = "stable"
	}
}

// determineDriftStatus determines the drift status based on drift values
func (monitor *MLModelAccuracyMonitor) determineDriftStatus(metrics *MLModelDetailedMetrics) {
	criticalDrift := 0
	warningDrift := 0

	// Check accuracy drift
	if math.Abs(metrics.AccuracyDrift) > monitor.config.AccuracyDriftThreshold*2 {
		criticalDrift++
	} else if math.Abs(metrics.AccuracyDrift) > monitor.config.AccuracyDriftThreshold {
		warningDrift++
	}

	// Check confidence drift
	if math.Abs(metrics.ConfidenceDrift) > monitor.config.ConfidenceDriftThreshold*2 {
		criticalDrift++
	} else if math.Abs(metrics.ConfidenceDrift) > monitor.config.ConfidenceDriftThreshold {
		warningDrift++
	}

	// Check latency drift
	if math.Abs(metrics.LatencyDrift) > float64(monitor.config.LatencyDriftThreshold.Milliseconds())*2 {
		criticalDrift++
	} else if math.Abs(metrics.LatencyDrift) > float64(monitor.config.LatencyDriftThreshold.Milliseconds()) {
		warningDrift++
	}

	if criticalDrift > 0 {
		metrics.DriftStatus = "critical"
	} else if warningDrift > 0 {
		metrics.DriftStatus = "warning"
	} else {
		metrics.DriftStatus = "stable"
	}
}

// createDriftAlerts creates drift alerts if necessary
func (monitor *MLModelAccuracyMonitor) createDriftAlerts(metrics *MLModelDetailedMetrics) {
	modelKey := fmt.Sprintf("%s_%s", metrics.ModelName, metrics.ModelVersion)

	// Check if we should create alerts based on cooldown and limits
	recentAlerts := monitor.getRecentAlertsForModel(modelKey)
	if len(recentAlerts) >= monitor.config.MaxAlertsPerModel {
		return
	}

	// Create accuracy drift alert
	if math.Abs(metrics.AccuracyDrift) > monitor.config.AccuracyDriftThreshold {
		alert := &DriftAlert{
			ID:              fmt.Sprintf("accuracy_drift_%s_%d", modelKey, time.Now().Unix()),
			ModelName:       modelKey,
			AlertType:       "accuracy_drift",
			Severity:        monitor.getDriftSeverity(metrics.AccuracyDrift, monitor.config.AccuracyDriftThreshold),
			Message:         fmt.Sprintf("Accuracy drift detected: %.2f%%", metrics.AccuracyDrift*100),
			CurrentValue:    metrics.AccuracyScore,
			BaselineValue:   metrics.AccuracyScore - metrics.AccuracyDrift,
			DriftPercentage: metrics.AccuracyDrift * 100,
			Timestamp:       time.Now(),
			Resolved:        false,
		}
		monitor.driftAlerts = append(monitor.driftAlerts, alert)
	}

	// Create confidence drift alert
	if math.Abs(metrics.ConfidenceDrift) > monitor.config.ConfidenceDriftThreshold {
		alert := &DriftAlert{
			ID:              fmt.Sprintf("confidence_drift_%s_%d", modelKey, time.Now().Unix()),
			ModelName:       modelKey,
			AlertType:       "confidence_drift",
			Severity:        monitor.getDriftSeverity(metrics.ConfidenceDrift, monitor.config.ConfidenceDriftThreshold),
			Message:         fmt.Sprintf("Confidence drift detected: %.2f%%", metrics.ConfidenceDrift*100),
			CurrentValue:    metrics.AverageConfidence,
			BaselineValue:   metrics.AverageConfidence - metrics.ConfidenceDrift,
			DriftPercentage: metrics.ConfidenceDrift * 100,
			Timestamp:       time.Now(),
			Resolved:        false,
		}
		monitor.driftAlerts = append(monitor.driftAlerts, alert)
	}

	// Create latency drift alert
	if math.Abs(metrics.LatencyDrift) > float64(monitor.config.LatencyDriftThreshold.Milliseconds()) {
		alert := &DriftAlert{
			ID:              fmt.Sprintf("latency_drift_%s_%d", modelKey, time.Now().Unix()),
			ModelName:       modelKey,
			AlertType:       "latency_drift",
			Severity:        monitor.getDriftSeverity(metrics.LatencyDrift, float64(monitor.config.LatencyDriftThreshold.Milliseconds())),
			Message:         fmt.Sprintf("Latency drift detected: %.2f ms", metrics.LatencyDrift),
			CurrentValue:    float64(metrics.AverageLatency.Milliseconds()),
			BaselineValue:   float64(metrics.AverageLatency.Milliseconds()) - metrics.LatencyDrift,
			DriftPercentage: (metrics.LatencyDrift / float64(metrics.AverageLatency.Milliseconds())) * 100,
			Timestamp:       time.Now(),
			Resolved:        false,
		}
		monitor.driftAlerts = append(monitor.driftAlerts, alert)
	}
}

// getRecentAlertsForModel gets recent alerts for a specific model
func (monitor *MLModelAccuracyMonitor) getRecentAlertsForModel(modelKey string) []*DriftAlert {
	var recentAlerts []*DriftAlert
	cutoff := time.Now().Add(-monitor.config.AlertCooldownPeriod)

	for _, alert := range monitor.driftAlerts {
		if alert.ModelName == modelKey && alert.Timestamp.After(cutoff) {
			recentAlerts = append(recentAlerts, alert)
		}
	}

	return recentAlerts
}

// getDriftSeverity determines alert severity based on drift magnitude
func (monitor *MLModelAccuracyMonitor) getDriftSeverity(drift, threshold float64) string {
	if math.Abs(drift) > threshold*2 {
		return "critical"
	}
	return "warning"
}

// copyModelMetrics creates a copy of model metrics
func (monitor *MLModelAccuracyMonitor) copyModelMetrics(metrics *MLModelDetailedMetrics) *MLModelDetailedMetrics {
	baseCopy := &MLModelMetrics{
		ModelName:          metrics.ModelName,
		ModelVersion:       metrics.ModelVersion,
		LastUpdated:        metrics.LastUpdated,
		TotalPredictions:   metrics.TotalPredictions,
		CorrectPredictions: metrics.CorrectPredictions,
		AccuracyScore:      metrics.AccuracyScore,
		AverageConfidence:  metrics.AverageConfidence,
		ModelDriftScore:    metrics.ModelDriftScore,
		LastRetrained:      metrics.LastRetrained,
		WindowedAccuracy:   make([]float64, len(metrics.WindowedAccuracy)),
		TrendIndicator:     metrics.TrendIndicator,
		UncertaintyScore:   metrics.UncertaintyScore,
	}

	// Copy windowed accuracy
	copy(baseCopy.WindowedAccuracy, metrics.WindowedAccuracy)

	detailedCopy := &MLModelDetailedMetrics{
		MLModelMetrics:       baseCopy,
		AverageLatency:       metrics.AverageLatency,
		AccuracyDrift:        metrics.AccuracyDrift,
		ConfidenceDrift:      metrics.ConfidenceDrift,
		LatencyDrift:         metrics.LatencyDrift,
		DriftStatus:          metrics.DriftStatus,
		PerformanceTrend:     metrics.PerformanceTrend,
		ReliabilityScore:     metrics.ReliabilityScore,
		StabilityScore:       metrics.StabilityScore,
		HistoricalAccuracy:   make([]*AccuracyDataPoint, len(metrics.HistoricalAccuracy)),
		HistoricalConfidence: make([]*ConfidenceDataPoint, len(metrics.HistoricalConfidence)),
		HistoricalLatency:    make([]*LatencyDataPoint, len(metrics.HistoricalLatency)),
	}

	// Copy historical data
	for i, point := range metrics.HistoricalAccuracy {
		detailedCopy.HistoricalAccuracy[i] = &AccuracyDataPoint{
			Timestamp:  point.Timestamp,
			Accuracy:   point.Accuracy,
			SampleSize: point.SampleSize,
		}
	}

	for i, point := range metrics.HistoricalConfidence {
		detailedCopy.HistoricalConfidence[i] = &ConfidenceDataPoint{
			Timestamp:  point.Timestamp,
			Confidence: point.Confidence,
			SampleSize: point.SampleSize,
		}
	}

	for i, point := range metrics.HistoricalLatency {
		detailedCopy.HistoricalLatency[i] = &LatencyDataPoint{
			Timestamp: point.Timestamp,
			Value:     point.Value,
		}
	}

	return detailedCopy
}

// MLTrendAnalyzer methods

// AnalyzeTrend analyzes trend in accuracy data
func (ta *MLTrendAnalyzer) AnalyzeTrend(data []*AccuracyDataPoint) string {
	if len(data) < ta.config.MinDataPointsForTrend {
		return "insufficient_data"
	}

	// Simple linear regression to determine trend
	slope := ta.calculateSlope(data)

	if slope > 0.01 {
		return "improving"
	} else if slope < -0.01 {
		return "degrading"
	}
	return "stable"
}

// AnalyzeConfidenceTrend analyzes trend in confidence data
func (ta *MLTrendAnalyzer) AnalyzeConfidenceTrend(data []*ConfidenceDataPoint) string {
	if len(data) < ta.config.MinDataPointsForTrend {
		return "insufficient_data"
	}

	slope := ta.calculateConfidenceSlope(data)

	if slope > 0.01 {
		return "improving"
	} else if slope < -0.01 {
		return "degrading"
	}
	return "stable"
}

// AnalyzeLatencyTrend analyzes trend in latency data
func (ta *MLTrendAnalyzer) AnalyzeLatencyTrend(data []*LatencyDataPoint) string {
	if len(data) < ta.config.MinDataPointsForTrend {
		return "insufficient_data"
	}

	slope := ta.calculateLatencySlope(data)

	if slope < -10 { // Decreasing latency is improving
		return "improving"
	} else if slope > 10 { // Increasing latency is degrading
		return "degrading"
	}
	return "stable"
}

// calculateSlope calculates the slope of accuracy trend
func (ta *MLTrendAnalyzer) calculateSlope(data []*AccuracyDataPoint) float64 {
	if len(data) < 2 {
		return 0.0
	}

	n := float64(len(data))
	var sumX, sumY, sumXY, sumXX float64

	for i, point := range data {
		x := float64(i)
		y := point.Accuracy
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	return slope
}

// calculateConfidenceSlope calculates the slope of confidence trend
func (ta *MLTrendAnalyzer) calculateConfidenceSlope(data []*ConfidenceDataPoint) float64 {
	if len(data) < 2 {
		return 0.0
	}

	n := float64(len(data))
	var sumX, sumY, sumXY, sumXX float64

	for i, point := range data {
		x := float64(i)
		y := point.Confidence
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	return slope
}

// calculateLatencySlope calculates the slope of latency trend
func (ta *MLTrendAnalyzer) calculateLatencySlope(data []*LatencyDataPoint) float64 {
	if len(data) < 2 {
		return 0.0
	}

	n := float64(len(data))
	var sumX, sumY, sumXY, sumXX float64

	for i, point := range data {
		x := float64(i)
		y := float64(point.Value.Milliseconds())
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	slope := (n*sumXY - sumX*sumY) / (n*sumXX - sumX*sumX)
	return slope
}

// DriftDetector methods

// DetectAccuracyDrift detects accuracy drift using statistical methods
func (dd *DriftDetector) DetectAccuracyDrift(data []*AccuracyDataPoint) float64 {
	if len(data) < dd.config.BaselineWindowSize {
		return 0.0
	}

	// Split data into baseline and recent windows
	baselineSize := dd.config.BaselineWindowSize / 2
	recentSize := dd.config.BaselineWindowSize / 2

	if len(data) < baselineSize+recentSize {
		return 0.0
	}

	baseline := data[:baselineSize]
	recent := data[len(data)-recentSize:]

	// Calculate means
	baselineMean := dd.calculateMean(baseline)
	recentMean := dd.calculateMean(recent)

	// Calculate drift as percentage change
	drift := (recentMean - baselineMean) / baselineMean
	return drift
}

// DetectConfidenceDrift detects confidence drift
func (dd *DriftDetector) DetectConfidenceDrift(data []*ConfidenceDataPoint) float64 {
	if len(data) < dd.config.BaselineWindowSize {
		return 0.0
	}

	baselineSize := dd.config.BaselineWindowSize / 2
	recentSize := dd.config.BaselineWindowSize / 2

	if len(data) < baselineSize+recentSize {
		return 0.0
	}

	baseline := data[:baselineSize]
	recent := data[len(data)-recentSize:]

	baselineMean := dd.calculateConfidenceMean(baseline)
	recentMean := dd.calculateConfidenceMean(recent)

	drift := (recentMean - baselineMean) / baselineMean
	return drift
}

// DetectLatencyDrift detects latency drift
func (dd *DriftDetector) DetectLatencyDrift(data []*LatencyDataPoint) float64 {
	if len(data) < dd.config.BaselineWindowSize {
		return 0.0
	}

	baselineSize := dd.config.BaselineWindowSize / 2
	recentSize := dd.config.BaselineWindowSize / 2

	if len(data) < baselineSize+recentSize {
		return 0.0
	}

	baseline := data[:baselineSize]
	recent := data[len(data)-recentSize:]

	baselineMean := dd.calculateLatencyMean(baseline)
	recentMean := dd.calculateLatencyMean(recent)

	drift := (recentMean - baselineMean) / baselineMean
	return drift
}

// calculateMean calculates mean of accuracy data points
func (dd *DriftDetector) calculateMean(data []*AccuracyDataPoint) float64 {
	if len(data) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, point := range data {
		sum += point.Accuracy
	}

	return sum / float64(len(data))
}

// calculateConfidenceMean calculates mean of confidence data points
func (dd *DriftDetector) calculateConfidenceMean(data []*ConfidenceDataPoint) float64 {
	if len(data) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, point := range data {
		sum += point.Confidence
	}

	return sum / float64(len(data))
}

// calculateLatencyMean calculates mean of latency data points
func (dd *DriftDetector) calculateLatencyMean(data []*LatencyDataPoint) float64 {
	if len(data) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, point := range data {
		sum += float64(point.Value.Milliseconds())
	}

	return sum / float64(len(data))
}

// MLPerformanceTracker methods

// CalculateReliabilityScore calculates reliability score based on consistency
func (pt *MLPerformanceTracker) CalculateReliabilityScore(metrics *MLModelDetailedMetrics) float64 {
	if len(metrics.HistoricalAccuracy) < 10 {
		return 0.0
	}

	// Calculate coefficient of variation (lower is more reliable)
	mean := pt.calculateAccuracyMean(metrics.HistoricalAccuracy)
	stdDev := pt.calculateAccuracyStdDev(metrics.HistoricalAccuracy, mean)

	if mean == 0 {
		return 0.0
	}

	cv := stdDev / mean
	// Convert to reliability score (0-1, higher is better)
	reliability := math.Max(0, 1.0-cv)
	return reliability
}

// CalculateStabilityScore calculates stability score based on variance
func (pt *MLPerformanceTracker) CalculateStabilityScore(metrics *MLModelDetailedMetrics) float64 {
	if len(metrics.HistoricalAccuracy) < 10 {
		return 0.0
	}

	// Calculate stability based on accuracy variance
	mean := pt.calculateAccuracyMean(metrics.HistoricalAccuracy)
	variance := pt.calculateAccuracyVariance(metrics.HistoricalAccuracy, mean)

	// Convert variance to stability score (0-1, higher is better)
	stability := math.Max(0, 1.0-variance)
	return stability
}

// calculateAccuracyMean calculates mean of accuracy data
func (pt *MLPerformanceTracker) calculateAccuracyMean(data []*AccuracyDataPoint) float64 {
	if len(data) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, point := range data {
		sum += point.Accuracy
	}

	return sum / float64(len(data))
}

// calculateAccuracyStdDev calculates standard deviation of accuracy data
func (pt *MLPerformanceTracker) calculateAccuracyStdDev(data []*AccuracyDataPoint, mean float64) float64 {
	if len(data) < 2 {
		return 0.0
	}

	sumSquaredDiff := 0.0
	for _, point := range data {
		diff := point.Accuracy - mean
		sumSquaredDiff += diff * diff
	}

	variance := sumSquaredDiff / float64(len(data)-1)
	return math.Sqrt(variance)
}

// calculateAccuracyVariance calculates variance of accuracy data
func (pt *MLPerformanceTracker) calculateAccuracyVariance(data []*AccuracyDataPoint, mean float64) float64 {
	if len(data) < 2 {
		return 0.0
	}

	sumSquaredDiff := 0.0
	for _, point := range data {
		diff := point.Accuracy - mean
		sumSquaredDiff += diff * diff
	}

	return sumSquaredDiff / float64(len(data)-1)
}
