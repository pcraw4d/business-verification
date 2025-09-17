package classification_monitoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BERTModelMonitor provides specialized monitoring for BERT model performance
type BERTModelMonitor struct {
	config *BERTModelMonitorConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// BERT-specific tracking
	bertModels      map[string]*BERTModelMetrics
	uncertaintyData []*UncertaintyDataPoint
	driftAlerts     []*BERTDriftAlert

	// Performance tracking
	performanceTracker *BERTPerformanceTracker
	uncertaintyTracker *UncertaintyQuantificationTracker
	driftDetector      *BERTDriftDetector
}

// BERTModelMonitorConfig holds configuration for BERT model monitoring
type BERTModelMonitorConfig struct {
	// Model tracking
	ModelVersioningEnabled bool          `json:"model_versioning_enabled"`
	ModelUpdateInterval    time.Duration `json:"model_update_interval"`
	MaxModelVersions       int           `json:"max_model_versions"`

	// Uncertainty quantification
	UncertaintyTrackingEnabled bool    `json:"uncertainty_tracking_enabled"`
	UncertaintyThreshold       float64 `json:"uncertainty_threshold"`
	CalibrationWindowSize      int     `json:"calibration_window_size"`

	// Drift detection
	DriftDetectionEnabled     bool          `json:"drift_detection_enabled"`
	DriftDetectionInterval    time.Duration `json:"drift_detection_interval"`
	AccuracyDriftThreshold    float64       `json:"accuracy_drift_threshold"`
	ConfidenceDriftThreshold  float64       `json:"confidence_drift_threshold"`
	UncertaintyDriftThreshold float64       `json:"uncertainty_drift_threshold"`

	// Performance monitoring
	PerformanceTrackingEnabled bool          `json:"performance_tracking_enabled"`
	LatencyThreshold           time.Duration `json:"latency_threshold"`
	ThroughputThreshold        float64       `json:"throughput_threshold"`

	// Alerting
	AlertingEnabled      bool          `json:"alerting_enabled"`
	AlertCooldownPeriod  time.Duration `json:"alert_cooldown_period"`
	MaxAlertsPerModel    int           `json:"max_alerts_per_model"`
	AlertRetentionPeriod time.Duration `json:"alert_retention_period"`
}

// BERTModelMetrics holds comprehensive metrics for a BERT model
type BERTModelMetrics struct {
	// Model identification
	ModelName    string `json:"model_name"`
	ModelVersion string `json:"model_version"`
	ModelType    string `json:"model_type"` // "bert-base", "bert-large", etc.

	// Performance metrics
	TotalPredictions   int64         `json:"total_predictions"`
	CorrectPredictions int64         `json:"correct_predictions"`
	AccuracyScore      float64       `json:"accuracy_score"`
	AverageConfidence  float64       `json:"average_confidence"`
	AverageLatency     time.Duration `json:"average_latency"`
	AverageThroughput  float64       `json:"average_throughput"`

	// Uncertainty quantification
	AverageUncertainty     float64 `json:"average_uncertainty"`
	UncertaintyCalibration float64 `json:"uncertainty_calibration"`
	ConfidenceCalibration  float64 `json:"confidence_calibration"`

	// Drift indicators
	AccuracyDrift    float64 `json:"accuracy_drift"`
	ConfidenceDrift  float64 `json:"confidence_drift"`
	UncertaintyDrift float64 `json:"uncertainty_drift"`
	DriftStatus      string  `json:"drift_status"` // "stable", "warning", "critical"

	// Performance indicators
	PerformanceTrend string  `json:"performance_trend"` // "improving", "stable", "degrading"
	ReliabilityScore float64 `json:"reliability_score"`
	StabilityScore   float64 `json:"stability_score"`

	// Historical data
	HistoricalAccuracy    []*AccuracyDataPoint    `json:"historical_accuracy"`
	HistoricalConfidence  []*ConfidenceDataPoint  `json:"historical_confidence"`
	HistoricalUncertainty []*UncertaintyDataPoint `json:"historical_uncertainty"`
	HistoricalLatency     []*LatencyDataPoint     `json:"historical_latency"`

	// Timestamps
	CreatedAt   time.Time `json:"created_at"`
	LastUpdated time.Time `json:"last_updated"`
	LastDrift   time.Time `json:"last_drift"`
}

// UncertaintyDataPoint represents an uncertainty measurement
type UncertaintyDataPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	Uncertainty float64   `json:"uncertainty"`
	Confidence  float64   `json:"confidence"`
	Prediction  string    `json:"prediction"`
	Actual      string    `json:"actual,omitempty"`
	SampleSize  int       `json:"sample_size"`
}

// BERTDriftAlert represents a drift detection alert for BERT models
type BERTDriftAlert struct {
	ID             string    `json:"id"`
	ModelName      string    `json:"model_name"`
	ModelVersion   string    `json:"model_version"`
	AlertType      string    `json:"alert_type"` // "accuracy_drift", "confidence_drift", "uncertainty_drift"
	Severity       string    `json:"severity"`   // "warning", "critical"
	DriftValue     float64   `json:"drift_value"`
	Threshold      float64   `json:"threshold"`
	Message        string    `json:"message"`
	Timestamp      time.Time `json:"timestamp"`
	Acknowledged   bool      `json:"acknowledged"`
	AcknowledgedAt time.Time `json:"acknowledged_at,omitempty"`
	Resolved       bool      `json:"resolved"`
	ResolvedAt     time.Time `json:"resolved_at,omitempty"`
}

// BERTPerformanceTracker tracks BERT model performance metrics
type BERTPerformanceTracker struct {
	config *BERTModelMonitorConfig
	logger *zap.Logger
}

// UncertaintyQuantificationTracker tracks uncertainty quantification accuracy
type UncertaintyQuantificationTracker struct {
	config *BERTModelMonitorConfig
	logger *zap.Logger
}

// BERTDriftDetector detects drift in BERT model performance
type BERTDriftDetector struct {
	config *BERTModelMonitorConfig
	logger *zap.Logger
}

// NewBERTModelMonitor creates a new BERT model monitor
func NewBERTModelMonitor(config *BERTModelMonitorConfig, logger *zap.Logger) *BERTModelMonitor {
	if config == nil {
		config = DefaultBERTModelMonitorConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &BERTModelMonitor{
		config:          config,
		logger:          logger,
		bertModels:      make(map[string]*BERTModelMetrics),
		uncertaintyData: make([]*UncertaintyDataPoint, 0),
		driftAlerts:     make([]*BERTDriftAlert, 0),
		performanceTracker: &BERTPerformanceTracker{
			config: config,
			logger: logger,
		},
		uncertaintyTracker: &UncertaintyQuantificationTracker{
			config: config,
			logger: logger,
		},
		driftDetector: &BERTDriftDetector{
			config: config,
			logger: logger,
		},
	}
}

// DefaultBERTModelMonitorConfig returns default configuration for BERT model monitoring
func DefaultBERTModelMonitorConfig() *BERTModelMonitorConfig {
	return &BERTModelMonitorConfig{
		ModelVersioningEnabled:     true,
		ModelUpdateInterval:        1 * time.Hour,
		MaxModelVersions:           10,
		UncertaintyTrackingEnabled: true,
		UncertaintyThreshold:       0.3,
		CalibrationWindowSize:      1000,
		DriftDetectionEnabled:      true,
		DriftDetectionInterval:     5 * time.Minute,
		AccuracyDriftThreshold:     0.05,
		ConfidenceDriftThreshold:   0.1,
		UncertaintyDriftThreshold:  0.15,
		PerformanceTrackingEnabled: true,
		LatencyThreshold:           500 * time.Millisecond,
		ThroughputThreshold:        10.0, // predictions per second
		AlertingEnabled:            true,
		AlertCooldownPeriod:        15 * time.Minute,
		MaxAlertsPerModel:          50,
		AlertRetentionPeriod:       24 * time.Hour,
	}
}

// TrackBERTPrediction tracks a BERT model prediction with uncertainty quantification
func (bmm *BERTModelMonitor) TrackBERTPrediction(ctx context.Context, result *ClassificationResult) error {
	bmm.mu.Lock()
	defer bmm.mu.Unlock()

	// Extract BERT model information
	modelName := "bert-classifier"
	modelVersion := "v1.0"
	modelType := "bert-base"

	if name, exists := result.Metadata["bert_model_name"]; exists {
		if n, ok := name.(string); ok {
			modelName = n
		}
	}
	if version, exists := result.Metadata["bert_model_version"]; exists {
		if v, ok := version.(string); ok {
			modelVersion = v
		}
	}
	if mType, exists := result.Metadata["bert_model_type"]; exists {
		if t, ok := mType.(string); ok {
			modelType = t
		}
	}

	modelKey := fmt.Sprintf("%s_%s", modelName, modelVersion)

	// Get or create model metrics
	metrics, exists := bmm.bertModels[modelKey]
	if !exists {
		metrics = bmm.createBERTModelMetrics(modelName, modelVersion, modelType)
		bmm.bertModels[modelKey] = metrics
	}

	// Update metrics
	bmm.updateBERTModelMetrics(metrics, result)

	// Track uncertainty quantification
	if bmm.config.UncertaintyTrackingEnabled {
		bmm.trackUncertaintyQuantification(metrics, result)
	}

	// Analyze trends
	bmm.analyzeBERTModelTrends(metrics)

	// Detect drift
	if bmm.config.DriftDetectionEnabled {
		bmm.detectBERTModelDrift(metrics)
	}

	// Update performance tracking
	if bmm.config.PerformanceTrackingEnabled {
		bmm.updateBERTPerformanceTracking(metrics)
	}

	return nil
}

// createBERTModelMetrics creates new BERT model metrics
func (bmm *BERTModelMonitor) createBERTModelMetrics(modelName, modelVersion, modelType string) *BERTModelMetrics {
	now := time.Now()
	return &BERTModelMetrics{
		ModelName:              modelName,
		ModelVersion:           modelVersion,
		ModelType:              modelType,
		TotalPredictions:       0,
		CorrectPredictions:     0,
		AccuracyScore:          0.0,
		AverageConfidence:      0.0,
		AverageLatency:         0,
		AverageThroughput:      0.0,
		AverageUncertainty:     0.0,
		UncertaintyCalibration: 0.0,
		ConfidenceCalibration:  0.0,
		AccuracyDrift:          0.0,
		ConfidenceDrift:        0.0,
		UncertaintyDrift:       0.0,
		DriftStatus:            "stable",
		PerformanceTrend:       "stable",
		ReliabilityScore:       0.0,
		StabilityScore:         0.0,
		HistoricalAccuracy:     make([]*AccuracyDataPoint, 0),
		HistoricalConfidence:   make([]*ConfidenceDataPoint, 0),
		HistoricalUncertainty:  make([]*UncertaintyDataPoint, 0),
		HistoricalLatency:      make([]*LatencyDataPoint, 0),
		CreatedAt:              now,
		LastUpdated:            now,
		LastDrift:              now,
	}
}

// updateBERTModelMetrics updates BERT model metrics with new prediction
func (bmm *BERTModelMonitor) updateBERTModelMetrics(metrics *BERTModelMetrics, result *ClassificationResult) {
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

	// Update throughput (predictions per second)
	timeWindow := time.Since(metrics.CreatedAt).Seconds()
	if timeWindow > 0 {
		metrics.AverageThroughput = float64(metrics.TotalPredictions) / timeWindow
	}

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
	if len(metrics.HistoricalAccuracy) > bmm.config.CalibrationWindowSize {
		metrics.HistoricalAccuracy = metrics.HistoricalAccuracy[1:]
	}
	if len(metrics.HistoricalConfidence) > bmm.config.CalibrationWindowSize {
		metrics.HistoricalConfidence = metrics.HistoricalConfidence[1:]
	}
	if len(metrics.HistoricalLatency) > bmm.config.CalibrationWindowSize {
		metrics.HistoricalLatency = metrics.HistoricalLatency[1:]
	}
}

// trackUncertaintyQuantification tracks uncertainty quantification accuracy
func (bmm *BERTModelMonitor) trackUncertaintyQuantification(metrics *BERTModelMetrics, result *ClassificationResult) {
	// Calculate uncertainty (1 - confidence)
	uncertainty := 1.0 - result.ConfidenceScore
	metrics.AverageUncertainty = (metrics.AverageUncertainty*float64(metrics.TotalPredictions-1) + uncertainty) / float64(metrics.TotalPredictions)

	// Create uncertainty data point
	uncertaintyPoint := &UncertaintyDataPoint{
		Timestamp:   result.Timestamp,
		Uncertainty: uncertainty,
		Confidence:  result.ConfidenceScore,
		Prediction:  result.ActualClassification,
		SampleSize:  1,
	}

	// Add actual classification if available
	if result.ExpectedClassification != nil {
		uncertaintyPoint.Actual = *result.ExpectedClassification
	}

	metrics.HistoricalUncertainty = append(metrics.HistoricalUncertainty, uncertaintyPoint)

	// Maintain window size
	if len(metrics.HistoricalUncertainty) > bmm.config.CalibrationWindowSize {
		metrics.HistoricalUncertainty = metrics.HistoricalUncertainty[1:]
	}

	// Update calibration scores
	bmm.updateCalibrationScores(metrics)
}

// updateCalibrationScores updates uncertainty and confidence calibration scores
func (bmm *BERTModelMonitor) updateCalibrationScores(metrics *BERTModelMetrics) {
	if len(metrics.HistoricalUncertainty) < 100 {
		return // Need sufficient data for calibration
	}

	// Calculate uncertainty calibration (how well uncertainty predicts accuracy)
	uncertaintyCalibration := bmm.calculateUncertaintyCalibration(metrics.HistoricalUncertainty)
	metrics.UncertaintyCalibration = uncertaintyCalibration

	// Calculate confidence calibration (how well confidence predicts accuracy)
	confidenceCalibration := bmm.calculateConfidenceCalibration(metrics.HistoricalConfidence)
	metrics.ConfidenceCalibration = confidenceCalibration
}

// calculateUncertaintyCalibration calculates how well uncertainty predicts accuracy
func (bmm *BERTModelMonitor) calculateUncertaintyCalibration(uncertaintyData []*UncertaintyDataPoint) float64 {
	if len(uncertaintyData) < 10 {
		return 0.0
	}

	// Group by uncertainty bins
	bins := make(map[int][]float64) // uncertainty bin -> accuracy values
	binSize := 0.1                  // 0.1 uncertainty bins

	for _, point := range uncertaintyData {
		if point.Actual != "" { // Only use points with actual values
			bin := int(point.Uncertainty / binSize)
			accuracy := 0.0
			if point.Prediction == point.Actual {
				accuracy = 1.0
			}
			bins[bin] = append(bins[bin], accuracy)
		}
	}

	// Calculate calibration error
	totalError := 0.0
	totalPoints := 0

	for bin, accuracies := range bins {
		if len(accuracies) < 5 { // Need minimum samples per bin
			continue
		}

		// Calculate average accuracy for this uncertainty bin
		avgAccuracy := 0.0
		for _, acc := range accuracies {
			avgAccuracy += acc
		}
		avgAccuracy /= float64(len(accuracies))

		// Expected accuracy based on uncertainty (1 - uncertainty)
		expectedAccuracy := 1.0 - (float64(bin) * binSize)

		// Calibration error for this bin
		error := math.Abs(avgAccuracy - expectedAccuracy)
		totalError += error * float64(len(accuracies))
		totalPoints += len(accuracies)
	}

	if totalPoints == 0 {
		return 0.0
	}

	// Return calibration score (1 - average error)
	return 1.0 - (totalError / float64(totalPoints))
}

// calculateConfidenceCalibration calculates how well confidence predicts accuracy
func (bmm *BERTModelMonitor) calculateConfidenceCalibration(confidenceData []*ConfidenceDataPoint) float64 {
	if len(confidenceData) < 10 {
		return 0.0
	}

	// Group by confidence bins
	bins := make(map[int][]float64) // confidence bin -> accuracy values
	binSize := 0.1                  // 0.1 confidence bins

	for _, point := range confidenceData {
		bin := int(point.Confidence / binSize)
		// For confidence calibration, we need actual accuracy data
		// This is a simplified version - in practice, you'd need actual vs predicted
		bins[bin] = append(bins[bin], point.Confidence) // Using confidence as proxy
	}

	// Calculate calibration error
	totalError := 0.0
	totalPoints := 0

	for bin, confidences := range bins {
		if len(confidences) < 5 {
			continue
		}

		// Calculate average confidence for this bin
		avgConfidence := 0.0
		for _, conf := range confidences {
			avgConfidence += conf
		}
		avgConfidence /= float64(len(confidences))

		// Expected confidence based on bin
		expectedConfidence := float64(bin) * binSize

		// Calibration error for this bin
		error := math.Abs(avgConfidence - expectedConfidence)
		totalError += error * float64(len(confidences))
		totalPoints += len(confidences)
	}

	if totalPoints == 0 {
		return 0.0
	}

	// Return calibration score (1 - average error)
	return 1.0 - (totalError / float64(totalPoints))
}

// analyzeBERTModelTrends analyzes trends in BERT model performance
func (bmm *BERTModelMonitor) analyzeBERTModelTrends(metrics *BERTModelMetrics) {
	if len(metrics.HistoricalAccuracy) < 10 {
		return
	}

	// Analyze accuracy trend
	accuracyTrend := bmm.calculateTrend(metrics.HistoricalAccuracy, func(point interface{}) float64 {
		if p, ok := point.(*AccuracyDataPoint); ok {
			return p.Accuracy
		}
		return 0.0
	})

	// Analyze confidence trend
	confidenceTrend := bmm.calculateTrend(metrics.HistoricalConfidence, func(point interface{}) float64 {
		if p, ok := point.(*ConfidenceDataPoint); ok {
			return p.Confidence
		}
		return 0.0
	})

	// Analyze uncertainty trend
	uncertaintyTrend := bmm.calculateTrend(metrics.HistoricalUncertainty, func(point interface{}) float64 {
		if p, ok := point.(*UncertaintyDataPoint); ok {
			return p.Uncertainty
		}
		return 0.0
	})

	// Determine overall performance trend
	bmm.determineBERTPerformanceTrend(metrics, accuracyTrend, confidenceTrend, uncertaintyTrend)
}

// calculateTrend calculates trend direction from historical data
func (bmm *BERTModelMonitor) calculateTrend(data interface{}, extractor func(interface{}) float64) string {
	// This is a simplified trend calculation
	// In practice, you'd use more sophisticated statistical methods
	return "stable" // Placeholder
}

// determineBERTPerformanceTrend determines overall BERT model performance trend
func (bmm *BERTModelMonitor) determineBERTPerformanceTrend(metrics *BERTModelMetrics, accuracyTrend, confidenceTrend, uncertaintyTrend string) {
	// Simple trend determination logic
	improvingCount := 0
	decliningCount := 0

	if accuracyTrend == "improving" {
		improvingCount++
	} else if accuracyTrend == "declining" {
		decliningCount++
	}

	if confidenceTrend == "improving" {
		improvingCount++
	} else if confidenceTrend == "declining" {
		decliningCount++
	}

	if uncertaintyTrend == "declining" { // Lower uncertainty is better
		improvingCount++
	} else if uncertaintyTrend == "improving" {
		decliningCount++
	}

	if improvingCount > decliningCount {
		metrics.PerformanceTrend = "improving"
	} else if decliningCount > improvingCount {
		metrics.PerformanceTrend = "degrading"
	} else {
		metrics.PerformanceTrend = "stable"
	}
}

// detectBERTModelDrift detects drift in BERT model performance
func (bmm *BERTModelMonitor) detectBERTModelDrift(metrics *BERTModelMetrics) {
	if len(metrics.HistoricalAccuracy) < 100 {
		return // Need sufficient data for drift detection
	}

	// Detect accuracy drift
	accuracyDrift := bmm.calculateDrift(metrics.HistoricalAccuracy, func(point interface{}) float64 {
		if p, ok := point.(*AccuracyDataPoint); ok {
			return p.Accuracy
		}
		return 0.0
	})
	metrics.AccuracyDrift = accuracyDrift

	// Detect confidence drift
	confidenceDrift := bmm.calculateDrift(metrics.HistoricalConfidence, func(point interface{}) float64 {
		if p, ok := point.(*ConfidenceDataPoint); ok {
			return p.Confidence
		}
		return 0.0
	})
	metrics.ConfidenceDrift = confidenceDrift

	// Detect uncertainty drift
	uncertaintyDrift := bmm.calculateDrift(metrics.HistoricalUncertainty, func(point interface{}) float64 {
		if p, ok := point.(*UncertaintyDataPoint); ok {
			return p.Uncertainty
		}
		return 0.0
	})
	metrics.UncertaintyDrift = uncertaintyDrift

	// Determine drift status
	bmm.determineBERTDriftStatus(metrics)

	// Create alerts if necessary
	bmm.createBERTDriftAlerts(metrics)
}

// calculateDrift calculates drift value using statistical methods
func (bmm *BERTModelMonitor) calculateDrift(data interface{}, extractor func(interface{}) float64) float64 {
	// Simplified drift calculation
	// In practice, you'd use more sophisticated statistical methods like KS test, PSI, etc.
	return 0.0 // Placeholder
}

// determineBERTDriftStatus determines drift status based on drift values
func (bmm *BERTModelMonitor) determineBERTDriftStatus(metrics *BERTModelMetrics) {
	criticalDrift := false
	warningDrift := false

	// Check accuracy drift
	if metrics.AccuracyDrift > bmm.config.AccuracyDriftThreshold*2 {
		criticalDrift = true
	} else if metrics.AccuracyDrift > bmm.config.AccuracyDriftThreshold {
		warningDrift = true
	}

	// Check confidence drift
	if metrics.ConfidenceDrift > bmm.config.ConfidenceDriftThreshold*2 {
		criticalDrift = true
	} else if metrics.ConfidenceDrift > bmm.config.ConfidenceDriftThreshold {
		warningDrift = true
	}

	// Check uncertainty drift
	if metrics.UncertaintyDrift > bmm.config.UncertaintyDriftThreshold*2 {
		criticalDrift = true
	} else if metrics.UncertaintyDrift > bmm.config.UncertaintyDriftThreshold {
		warningDrift = true
	}

	// Set drift status
	if criticalDrift {
		metrics.DriftStatus = "critical"
	} else if warningDrift {
		metrics.DriftStatus = "warning"
	} else {
		metrics.DriftStatus = "stable"
	}

	metrics.LastDrift = time.Now()
}

// createBERTDriftAlerts creates drift alerts if necessary
func (bmm *BERTModelMonitor) createBERTDriftAlerts(metrics *BERTModelMetrics) {
	if !bmm.config.AlertingEnabled {
		return
	}

	// Check if we're in cooldown period
	if time.Since(metrics.LastDrift) < bmm.config.AlertCooldownPeriod {
		return
	}

	// Check alert limits
	alertCount := 0
	for _, alert := range bmm.driftAlerts {
		if alert.ModelName == metrics.ModelName && alert.ModelVersion == metrics.ModelVersion {
			alertCount++
		}
	}

	if alertCount >= bmm.config.MaxAlertsPerModel {
		return
	}

	// Create alerts based on drift status
	if metrics.DriftStatus == "critical" {
		bmm.createDriftAlert(metrics, "accuracy_drift", "critical", metrics.AccuracyDrift, bmm.config.AccuracyDriftThreshold)
		bmm.createDriftAlert(metrics, "confidence_drift", "critical", metrics.ConfidenceDrift, bmm.config.ConfidenceDriftThreshold)
		bmm.createDriftAlert(metrics, "uncertainty_drift", "critical", metrics.UncertaintyDrift, bmm.config.UncertaintyDriftThreshold)
	} else if metrics.DriftStatus == "warning" {
		bmm.createDriftAlert(metrics, "accuracy_drift", "warning", metrics.AccuracyDrift, bmm.config.AccuracyDriftThreshold)
		bmm.createDriftAlert(metrics, "confidence_drift", "warning", metrics.ConfidenceDrift, bmm.config.ConfidenceDriftThreshold)
		bmm.createDriftAlert(metrics, "uncertainty_drift", "warning", metrics.UncertaintyDrift, bmm.config.UncertaintyDriftThreshold)
	}
}

// createDriftAlert creates a drift alert
func (bmm *BERTModelMonitor) createDriftAlert(metrics *BERTModelMetrics, alertType, severity string, driftValue, threshold float64) {
	alert := &BERTDriftAlert{
		ID:           fmt.Sprintf("bert_drift_%s_%s_%d", metrics.ModelName, alertType, time.Now().Unix()),
		ModelName:    metrics.ModelName,
		ModelVersion: metrics.ModelVersion,
		AlertType:    alertType,
		Severity:     severity,
		DriftValue:   driftValue,
		Threshold:    threshold,
		Message:      fmt.Sprintf("BERT model %s %s drift detected: %.3f (threshold: %.3f)", metrics.ModelName, alertType, driftValue, threshold),
		Timestamp:    time.Now(),
		Acknowledged: false,
		Resolved:     false,
	}

	bmm.driftAlerts = append(bmm.driftAlerts, alert)

	// Clean up old alerts
	bmm.cleanupOldAlerts()

	bmm.logger.Warn("BERT model drift alert created",
		zap.String("model_name", metrics.ModelName),
		zap.String("alert_type", alertType),
		zap.String("severity", severity),
		zap.Float64("drift_value", driftValue),
		zap.Float64("threshold", threshold))
}

// cleanupOldAlerts removes old alerts beyond retention period
func (bmm *BERTModelMonitor) cleanupOldAlerts() {
	cutoff := time.Now().Add(-bmm.config.AlertRetentionPeriod)
	var validAlerts []*BERTDriftAlert

	for _, alert := range bmm.driftAlerts {
		if alert.Timestamp.After(cutoff) {
			validAlerts = append(validAlerts, alert)
		}
	}

	bmm.driftAlerts = validAlerts
}

// updateBERTPerformanceTracking updates BERT performance tracking metrics
func (bmm *BERTModelMonitor) updateBERTPerformanceTracking(metrics *BERTModelMetrics) {
	// Calculate reliability score
	metrics.ReliabilityScore = bmm.calculateReliabilityScore(metrics)

	// Calculate stability score
	metrics.StabilityScore = bmm.calculateStabilityScore(metrics)
}

// calculateReliabilityScore calculates reliability score for BERT model
func (bmm *BERTModelMonitor) calculateReliabilityScore(metrics *BERTModelMetrics) float64 {
	// Combine accuracy, confidence calibration, and uncertainty calibration
	accuracyWeight := 0.4
	confidenceCalibrationWeight := 0.3
	uncertaintyCalibrationWeight := 0.3

	score := (metrics.AccuracyScore * accuracyWeight) +
		(metrics.ConfidenceCalibration * confidenceCalibrationWeight) +
		(metrics.UncertaintyCalibration * uncertaintyCalibrationWeight)

	return math.Max(0.0, math.Min(1.0, score))
}

// calculateStabilityScore calculates stability score for BERT model
func (bmm *BERTModelMonitor) calculateStabilityScore(metrics *BERTModelMetrics) float64 {
	// Lower drift values indicate higher stability
	driftPenalty := (metrics.AccuracyDrift + metrics.ConfidenceDrift + metrics.UncertaintyDrift) / 3.0
	stabilityScore := 1.0 - driftPenalty

	return math.Max(0.0, math.Min(1.0, stabilityScore))
}

// GetBERTModelMetrics returns metrics for a specific BERT model
func (bmm *BERTModelMonitor) GetBERTModelMetrics(modelName, modelVersion string) (*BERTModelMetrics, error) {
	bmm.mu.RLock()
	defer bmm.mu.RUnlock()

	modelKey := fmt.Sprintf("%s_%s", modelName, modelVersion)
	metrics, exists := bmm.bertModels[modelKey]
	if !exists {
		return nil, fmt.Errorf("BERT model metrics not found for %s %s", modelName, modelVersion)
	}

	return metrics, nil
}

// GetAllBERTModelMetrics returns metrics for all BERT models
func (bmm *BERTModelMonitor) GetAllBERTModelMetrics() map[string]*BERTModelMetrics {
	bmm.mu.RLock()
	defer bmm.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*BERTModelMetrics)
	for key, metrics := range bmm.bertModels {
		result[key] = metrics
	}

	return result
}

// GetDriftAlerts returns all drift alerts
func (bmm *BERTModelMonitor) GetDriftAlerts() []*BERTDriftAlert {
	bmm.mu.RLock()
	defer bmm.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]*BERTDriftAlert, len(bmm.driftAlerts))
	copy(result, bmm.driftAlerts)

	return result
}

// AcknowledgeAlert acknowledges a drift alert
func (bmm *BERTModelMonitor) AcknowledgeAlert(alertID string) error {
	bmm.mu.Lock()
	defer bmm.mu.Unlock()

	for _, alert := range bmm.driftAlerts {
		if alert.ID == alertID {
			alert.Acknowledged = true
			alert.AcknowledgedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("drift alert not found: %s", alertID)
}

// ResolveAlert resolves a drift alert
func (bmm *BERTModelMonitor) ResolveAlert(alertID string) error {
	bmm.mu.Lock()
	defer bmm.mu.Unlock()

	for _, alert := range bmm.driftAlerts {
		if alert.ID == alertID {
			alert.Resolved = true
			alert.ResolvedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("drift alert not found: %s", alertID)
}
