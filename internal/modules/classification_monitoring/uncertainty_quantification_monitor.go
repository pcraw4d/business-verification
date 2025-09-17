package classification_monitoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// UncertaintyQuantificationMonitor provides comprehensive monitoring for uncertainty quantification accuracy
type UncertaintyQuantificationMonitor struct {
	config *UncertaintyQuantificationConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// Uncertainty tracking
	uncertaintyData   []*UncertaintyDataPoint
	calibrationData   []*CalibrationDataPoint
	reliabilityData   []*ReliabilityDataPoint
	uncertaintyAlerts []*UncertaintyAlert

	// Analysis components
	calibrationAnalyzer *UncertaintyCalibrationAnalyzer
	reliabilityTracker  *UncertaintyReliabilityTracker
	accuracyValidator   *UncertaintyAccuracyValidator
}

// UncertaintyQuantificationConfig holds configuration for uncertainty quantification monitoring
type UncertaintyQuantificationConfig struct {
	// Data collection
	DataCollectionEnabled bool          `json:"data_collection_enabled"`
	MaxDataPoints         int           `json:"max_data_points"`
	CollectionInterval    time.Duration `json:"collection_interval"`

	// Calibration analysis
	CalibrationAnalysisEnabled bool    `json:"calibration_analysis_enabled"`
	CalibrationBins            int     `json:"calibration_bins"`
	CalibrationThreshold       float64 `json:"calibration_threshold"`
	MinSamplesPerBin           int     `json:"min_samples_per_bin"`

	// Reliability tracking
	ReliabilityTrackingEnabled bool    `json:"reliability_tracking_enabled"`
	ReliabilityWindowSize      int     `json:"reliability_window_size"`
	ReliabilityThreshold       float64 `json:"reliability_threshold"`

	// Accuracy validation
	AccuracyValidationEnabled bool    `json:"accuracy_validation_enabled"`
	AccuracyThreshold         float64 `json:"accuracy_threshold"`
	ValidationWindowSize      int     `json:"validation_window_size"`

	// Alerting
	AlertingEnabled      bool          `json:"alerting_enabled"`
	AlertCooldownPeriod  time.Duration `json:"alert_cooldown_period"`
	MaxAlertsPerType     int           `json:"max_alerts_per_type"`
	AlertRetentionPeriod time.Duration `json:"alert_retention_period"`

	// Monitoring intervals
	MonitoringInterval time.Duration `json:"monitoring_interval"`
	AnalysisInterval   time.Duration `json:"analysis_interval"`
	ValidationInterval time.Duration `json:"validation_interval"`
}

// CalibrationDataPoint represents a calibration measurement
type CalibrationDataPoint struct {
	Timestamp         time.Time `json:"timestamp"`
	ConfidenceBin     int       `json:"confidence_bin"`
	ConfidenceRange   string    `json:"confidence_range"`
	PredictedAccuracy float64   `json:"predicted_accuracy"`
	ActualAccuracy    float64   `json:"actual_accuracy"`
	CalibrationError  float64   `json:"calibration_error"`
	SampleSize        int       `json:"sample_size"`
}

// ReliabilityDataPoint represents a reliability measurement
type ReliabilityDataPoint struct {
	Timestamp          time.Time `json:"timestamp"`
	UncertaintyScore   float64   `json:"uncertainty_score"`
	ConfidenceScore    float64   `json:"confidence_score"`
	PredictionAccuracy float64   `json:"prediction_accuracy"`
	ReliabilityScore   float64   `json:"reliability_score"`
	CalibrationScore   float64   `json:"calibration_score"`
	SampleSize         int       `json:"sample_size"`
}

// UncertaintyAlert represents an alert about uncertainty quantification issues
type UncertaintyAlert struct {
	ID             string             `json:"id"`
	AlertType      string             `json:"alert_type"` // "poor_calibration", "low_reliability", "accuracy_mismatch", "uncertainty_drift"
	Severity       string             `json:"severity"`   // "warning", "critical"
	Message        string             `json:"message"`
	Details        string             `json:"details"`
	Metrics        map[string]float64 `json:"metrics"`
	Timestamp      time.Time          `json:"timestamp"`
	Acknowledged   bool               `json:"acknowledged"`
	AcknowledgedAt time.Time          `json:"acknowledged_at,omitempty"`
	Resolved       bool               `json:"resolved"`
	ResolvedAt     time.Time          `json:"resolved_at,omitempty"`
}

// UncertaintyCalibrationAnalyzer analyzes uncertainty calibration
type UncertaintyCalibrationAnalyzer struct {
	config *UncertaintyQuantificationConfig
	logger *zap.Logger
}

// UncertaintyReliabilityTracker tracks uncertainty reliability
type UncertaintyReliabilityTracker struct {
	config *UncertaintyQuantificationConfig
	logger *zap.Logger
}

// UncertaintyAccuracyValidator validates uncertainty accuracy
type UncertaintyAccuracyValidator struct {
	config *UncertaintyQuantificationConfig
	logger *zap.Logger
}

// NewUncertaintyQuantificationMonitor creates a new uncertainty quantification monitor
func NewUncertaintyQuantificationMonitor(config *UncertaintyQuantificationConfig, logger *zap.Logger) *UncertaintyQuantificationMonitor {
	if config == nil {
		config = DefaultUncertaintyQuantificationConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &UncertaintyQuantificationMonitor{
		config:            config,
		logger:            logger,
		uncertaintyData:   make([]*UncertaintyDataPoint, 0),
		calibrationData:   make([]*CalibrationDataPoint, 0),
		reliabilityData:   make([]*ReliabilityDataPoint, 0),
		uncertaintyAlerts: make([]*UncertaintyAlert, 0),
		calibrationAnalyzer: &UncertaintyCalibrationAnalyzer{
			config: config,
			logger: logger,
		},
		reliabilityTracker: &UncertaintyReliabilityTracker{
			config: config,
			logger: logger,
		},
		accuracyValidator: &UncertaintyAccuracyValidator{
			config: config,
			logger: logger,
		},
	}
}

// DefaultUncertaintyQuantificationConfig returns default configuration
func DefaultUncertaintyQuantificationConfig() *UncertaintyQuantificationConfig {
	return &UncertaintyQuantificationConfig{
		DataCollectionEnabled:      true,
		MaxDataPoints:              10000,
		CollectionInterval:         1 * time.Minute,
		CalibrationAnalysisEnabled: true,
		CalibrationBins:            10,
		CalibrationThreshold:       0.1, // 10% calibration error threshold
		MinSamplesPerBin:           20,
		ReliabilityTrackingEnabled: true,
		ReliabilityWindowSize:      1000,
		ReliabilityThreshold:       0.7, // 70% reliability threshold
		AccuracyValidationEnabled:  true,
		AccuracyThreshold:          0.8, // 80% accuracy threshold
		ValidationWindowSize:       500,
		AlertingEnabled:            true,
		AlertCooldownPeriod:        15 * time.Minute,
		MaxAlertsPerType:           20,
		AlertRetentionPeriod:       24 * time.Hour,
		MonitoringInterval:         1 * time.Minute,
		AnalysisInterval:           5 * time.Minute,
		ValidationInterval:         10 * time.Minute,
	}
}

// TrackUncertaintyPrediction tracks an uncertainty prediction with accuracy validation
func (uqm *UncertaintyQuantificationMonitor) TrackUncertaintyPrediction(
	ctx context.Context,
	prediction string,
	confidence float64,
	uncertainty float64,
	actual string,
	isCorrect bool,
) error {
	uqm.mu.Lock()
	defer uqm.mu.Unlock()

	if !uqm.config.DataCollectionEnabled {
		return nil
	}

	// Create uncertainty data point
	uncertaintyPoint := &UncertaintyDataPoint{
		Timestamp:   time.Now(),
		Uncertainty: uncertainty,
		Confidence:  confidence,
		Prediction:  prediction,
		Actual:      actual,
		SampleSize:  1,
	}

	// Add to uncertainty data
	uqm.uncertaintyData = append(uqm.uncertaintyData, uncertaintyPoint)

	// Maintain data size
	if len(uqm.uncertaintyData) > uqm.config.MaxDataPoints {
		uqm.uncertaintyData = uqm.uncertaintyData[1:]
	}

	// Analyze calibration
	if uqm.config.CalibrationAnalysisEnabled {
		uqm.analyzeCalibration(uncertaintyPoint, isCorrect)
	}

	// Track reliability
	if uqm.config.ReliabilityTrackingEnabled {
		uqm.trackReliability(uncertaintyPoint, isCorrect)
	}

	// Validate accuracy
	if uqm.config.AccuracyValidationEnabled {
		uqm.validateAccuracy(uncertaintyPoint, isCorrect)
	}

	uqm.logger.Debug("Uncertainty prediction tracked",
		zap.String("prediction", prediction),
		zap.Float64("confidence", confidence),
		zap.Float64("uncertainty", uncertainty),
		zap.Bool("is_correct", isCorrect))

	return nil
}

// analyzeCalibration analyzes uncertainty calibration
func (uqm *UncertaintyQuantificationMonitor) analyzeCalibration(point *UncertaintyDataPoint, isCorrect bool) {
	// Determine confidence bin
	binSize := 1.0 / float64(uqm.config.CalibrationBins)
	binIndex := int(point.Confidence / binSize)
	if binIndex >= uqm.config.CalibrationBins {
		binIndex = uqm.config.CalibrationBins - 1
	}

	// Find or create calibration data for this bin
	var calibrationPoint *CalibrationDataPoint
	for _, cp := range uqm.calibrationData {
		if cp.ConfidenceBin == binIndex {
			calibrationPoint = cp
			break
		}
	}

	if calibrationPoint == nil {
		calibrationPoint = &CalibrationDataPoint{
			Timestamp:         time.Now(),
			ConfidenceBin:     binIndex,
			ConfidenceRange:   fmt.Sprintf("%.2f-%.2f", float64(binIndex)*binSize, float64(binIndex+1)*binSize),
			PredictedAccuracy: 0.0,
			ActualAccuracy:    0.0,
			CalibrationError:  0.0,
			SampleSize:        0,
		}
		uqm.calibrationData = append(uqm.calibrationData, calibrationPoint)
	}

	// Update calibration data
	calibrationPoint.SampleSize++

	// Update predicted accuracy (confidence should predict accuracy)
	calibrationPoint.PredictedAccuracy = (calibrationPoint.PredictedAccuracy*float64(calibrationPoint.SampleSize-1) + point.Confidence) / float64(calibrationPoint.SampleSize)

	// Update actual accuracy
	accuracy := 0.0
	if isCorrect {
		accuracy = 1.0
	}
	calibrationPoint.ActualAccuracy = (calibrationPoint.ActualAccuracy*float64(calibrationPoint.SampleSize-1) + accuracy) / float64(calibrationPoint.SampleSize)

	// Calculate calibration error
	calibrationPoint.CalibrationError = math.Abs(calibrationPoint.PredictedAccuracy - calibrationPoint.ActualAccuracy)

	// Check for poor calibration
	if calibrationPoint.SampleSize >= uqm.config.MinSamplesPerBin {
		if calibrationPoint.CalibrationError > uqm.config.CalibrationThreshold {
			uqm.createUncertaintyAlert("poor_calibration", "warning", "",
				fmt.Sprintf("Poor calibration in confidence bin %d (%.2f-%.2f): error %.3f",
					binIndex, float64(binIndex)*binSize, float64(binIndex+1)*binSize, calibrationPoint.CalibrationError),
				fmt.Sprintf("Predicted: %.3f, Actual: %.3f, Samples: %d",
					calibrationPoint.PredictedAccuracy, calibrationPoint.ActualAccuracy, calibrationPoint.SampleSize),
				map[string]float64{
					"confidence_bin":     float64(binIndex),
					"predicted_accuracy": calibrationPoint.PredictedAccuracy,
					"actual_accuracy":    calibrationPoint.ActualAccuracy,
					"calibration_error":  calibrationPoint.CalibrationError,
					"sample_size":        float64(calibrationPoint.SampleSize),
				})
		}
	}
}

// trackReliability tracks uncertainty reliability
func (uqm *UncertaintyQuantificationMonitor) trackReliability(point *UncertaintyDataPoint, isCorrect bool) {
	// Calculate reliability score (combination of calibration and accuracy)
	calibrationScore := 1.0 - math.Min(point.Confidence, 1.0-point.Confidence) // Higher confidence = higher calibration expectation
	accuracyScore := 0.0
	if isCorrect {
		accuracyScore = 1.0
	}

	// Reliability is the weighted combination of calibration and accuracy
	reliabilityScore := (calibrationScore * 0.6) + (accuracyScore * 0.4)

	// Create reliability data point
	reliabilityPoint := &ReliabilityDataPoint{
		Timestamp:          time.Now(),
		UncertaintyScore:   point.Uncertainty,
		ConfidenceScore:    point.Confidence,
		PredictionAccuracy: accuracyScore,
		ReliabilityScore:   reliabilityScore,
		CalibrationScore:   calibrationScore,
		SampleSize:         1,
	}

	// Add to reliability data
	uqm.reliabilityData = append(uqm.reliabilityData, reliabilityPoint)

	// Maintain data size
	if len(uqm.reliabilityData) > uqm.config.ReliabilityWindowSize {
		uqm.reliabilityData = uqm.reliabilityData[1:]
	}

	// Check for low reliability
	if len(uqm.reliabilityData) >= 100 { // Need sufficient data
		avgReliability := uqm.calculateAverageReliability()
		if avgReliability < uqm.config.ReliabilityThreshold {
			uqm.createUncertaintyAlert("low_reliability", "warning", "",
				fmt.Sprintf("Low uncertainty reliability: %.3f (threshold: %.3f)",
					avgReliability, uqm.config.ReliabilityThreshold),
				fmt.Sprintf("Recent reliability trend indicates poor uncertainty quantification"),
				map[string]float64{
					"average_reliability": avgReliability,
					"threshold":           uqm.config.ReliabilityThreshold,
					"sample_size":         float64(len(uqm.reliabilityData)),
				})
		}
	}
}

// validateAccuracy validates uncertainty accuracy
func (uqm *UncertaintyQuantificationMonitor) validateAccuracy(point *UncertaintyDataPoint, isCorrect bool) {
	// Calculate accuracy based on confidence threshold
	expectedAccuracy := 0.0
	if point.Confidence > 0.5 { // High confidence should mean high accuracy
		expectedAccuracy = point.Confidence
	} else { // Low confidence should mean low accuracy
		expectedAccuracy = point.Confidence
	}

	actualAccuracy := 0.0
	if isCorrect {
		actualAccuracy = 1.0
	}

	accuracyMismatch := math.Abs(expectedAccuracy - actualAccuracy)

	// Check for accuracy mismatch
	if accuracyMismatch > 0.3 { // 30% mismatch threshold
		uqm.createUncertaintyAlert("accuracy_mismatch", "critical", "",
			fmt.Sprintf("Accuracy mismatch detected: expected %.3f, actual %.3f",
				expectedAccuracy, actualAccuracy),
			fmt.Sprintf("Confidence: %.3f, Uncertainty: %.3f, Prediction: %s",
				point.Confidence, point.Uncertainty, point.Prediction),
			map[string]float64{
				"expected_accuracy": expectedAccuracy,
				"actual_accuracy":   actualAccuracy,
				"accuracy_mismatch": accuracyMismatch,
				"confidence":        point.Confidence,
				"uncertainty":       point.Uncertainty,
			})
	}
}

// calculateAverageReliability calculates average reliability over recent data
func (uqm *UncertaintyQuantificationMonitor) calculateAverageReliability() float64 {
	if len(uqm.reliabilityData) == 0 {
		return 0.0
	}

	// Use recent data (last 100 points or all if less)
	startIndex := 0
	if len(uqm.reliabilityData) > 100 {
		startIndex = len(uqm.reliabilityData) - 100
	}

	sum := 0.0
	count := 0
	for i := startIndex; i < len(uqm.reliabilityData); i++ {
		sum += uqm.reliabilityData[i].ReliabilityScore
		count++
	}

	if count == 0 {
		return 0.0
	}

	return sum / float64(count)
}

// createUncertaintyAlert creates an uncertainty alert
func (uqm *UncertaintyQuantificationMonitor) createUncertaintyAlert(alertType, severity, methodName, message, details string, metrics map[string]float64) {
	if !uqm.config.AlertingEnabled {
		return
	}

	// Check cooldown period
	lastAlert := uqm.getLastAlertOfType(alertType)
	if lastAlert != nil && time.Since(lastAlert.Timestamp) < uqm.config.AlertCooldownPeriod {
		return
	}

	// Check alert limits
	alertCount := uqm.getAlertCountOfType(alertType)
	if alertCount >= uqm.config.MaxAlertsPerType {
		return
	}

	alert := &UncertaintyAlert{
		ID:           fmt.Sprintf("uncertainty_alert_%s_%d", alertType, time.Now().Unix()),
		AlertType:    alertType,
		Severity:     severity,
		Message:      message,
		Details:      details,
		Metrics:      metrics,
		Timestamp:    time.Now(),
		Acknowledged: false,
		Resolved:     false,
	}

	uqm.uncertaintyAlerts = append(uqm.uncertaintyAlerts, alert)

	// Clean up old alerts
	uqm.cleanupOldAlerts()

	uqm.logger.Warn("Uncertainty alert created",
		zap.String("alert_type", alertType),
		zap.String("severity", severity),
		zap.String("message", message))
}

// getLastAlertOfType gets the last alert of a specific type
func (uqm *UncertaintyQuantificationMonitor) getLastAlertOfType(alertType string) *UncertaintyAlert {
	for i := len(uqm.uncertaintyAlerts) - 1; i >= 0; i-- {
		alert := uqm.uncertaintyAlerts[i]
		if alert.AlertType == alertType {
			return alert
		}
	}
	return nil
}

// getAlertCountOfType gets the alert count for a specific type
func (uqm *UncertaintyQuantificationMonitor) getAlertCountOfType(alertType string) int {
	count := 0
	for _, alert := range uqm.uncertaintyAlerts {
		if alert.AlertType == alertType {
			count++
		}
	}
	return count
}

// cleanupOldAlerts removes old alerts beyond retention period
func (uqm *UncertaintyQuantificationMonitor) cleanupOldAlerts() {
	cutoff := time.Now().Add(-uqm.config.AlertRetentionPeriod)
	var validAlerts []*UncertaintyAlert

	for _, alert := range uqm.uncertaintyAlerts {
		if alert.Timestamp.After(cutoff) {
			validAlerts = append(validAlerts, alert)
		}
	}

	uqm.uncertaintyAlerts = validAlerts
}

// GetUncertaintyMetrics returns comprehensive uncertainty quantification metrics
func (uqm *UncertaintyQuantificationMonitor) GetUncertaintyMetrics() *UncertaintyQuantificationMetrics {
	uqm.mu.RLock()
	defer uqm.mu.RUnlock()

	metrics := &UncertaintyQuantificationMetrics{
		Timestamp:          time.Now(),
		TotalPredictions:   len(uqm.uncertaintyData),
		CalibrationBins:    uqm.config.CalibrationBins,
		CalibrationData:    make([]*CalibrationDataPoint, len(uqm.calibrationData)),
		ReliabilityData:    make([]*ReliabilityDataPoint, len(uqm.reliabilityData)),
		OverallCalibration: 0.0,
		OverallReliability: 0.0,
		OverallAccuracy:    0.0,
		AlertsCount:        len(uqm.uncertaintyAlerts),
		HealthStatus:       "healthy",
	}

	// Copy calibration data
	copy(metrics.CalibrationData, uqm.calibrationData)

	// Copy reliability data
	copy(metrics.ReliabilityData, uqm.reliabilityData)

	// Calculate overall metrics
	metrics.OverallCalibration = uqm.calculateOverallCalibration()
	metrics.OverallReliability = uqm.calculateOverallReliability()
	metrics.OverallAccuracy = uqm.calculateOverallAccuracy()

	// Determine health status
	metrics.HealthStatus = uqm.determineHealthStatus()

	return metrics
}

// UncertaintyQuantificationMetrics represents comprehensive uncertainty quantification metrics
type UncertaintyQuantificationMetrics struct {
	Timestamp          time.Time               `json:"timestamp"`
	TotalPredictions   int                     `json:"total_predictions"`
	CalibrationBins    int                     `json:"calibration_bins"`
	CalibrationData    []*CalibrationDataPoint `json:"calibration_data"`
	ReliabilityData    []*ReliabilityDataPoint `json:"reliability_data"`
	OverallCalibration float64                 `json:"overall_calibration"`
	OverallReliability float64                 `json:"overall_reliability"`
	OverallAccuracy    float64                 `json:"overall_accuracy"`
	AlertsCount        int                     `json:"alerts_count"`
	HealthStatus       string                  `json:"health_status"`
}

// calculateOverallCalibration calculates overall calibration score
func (uqm *UncertaintyQuantificationMonitor) calculateOverallCalibration() float64 {
	if len(uqm.calibrationData) == 0 {
		return 0.0
	}

	totalError := 0.0
	totalSamples := 0

	for _, point := range uqm.calibrationData {
		if point.SampleSize >= uqm.config.MinSamplesPerBin {
			totalError += point.CalibrationError * float64(point.SampleSize)
			totalSamples += point.SampleSize
		}
	}

	if totalSamples == 0 {
		return 0.0
	}

	// Return calibration score (1 - average error)
	avgError := totalError / float64(totalSamples)
	return 1.0 - math.Min(avgError, 1.0)
}

// calculateOverallReliability calculates overall reliability score
func (uqm *UncertaintyQuantificationMonitor) calculateOverallReliability() float64 {
	return uqm.calculateAverageReliability()
}

// calculateOverallAccuracy calculates overall accuracy
func (uqm *UncertaintyQuantificationMonitor) calculateOverallAccuracy() float64 {
	if len(uqm.uncertaintyData) == 0 {
		return 0.0
	}

	correctCount := 0
	for _, point := range uqm.uncertaintyData {
		if point.Actual != "" && point.Prediction == point.Actual {
			correctCount++
		}
	}

	return float64(correctCount) / float64(len(uqm.uncertaintyData))
}

// determineHealthStatus determines overall health status
func (uqm *UncertaintyQuantificationMonitor) determineHealthStatus() string {
	// Check for critical alerts
	for _, alert := range uqm.uncertaintyAlerts {
		if alert.Severity == "critical" && !alert.Resolved {
			return "critical"
		}
	}

	// Check for warning alerts
	for _, alert := range uqm.uncertaintyAlerts {
		if alert.Severity == "warning" && !alert.Resolved {
			return "warning"
		}
	}

	// Check calibration threshold
	if uqm.calculateOverallCalibration() < 0.7 { // 70% calibration threshold
		return "warning"
	}

	// Check reliability threshold
	if uqm.calculateOverallReliability() < uqm.config.ReliabilityThreshold {
		return "warning"
	}

	return "healthy"
}

// GetUncertaintyAlerts returns all uncertainty alerts
func (uqm *UncertaintyQuantificationMonitor) GetUncertaintyAlerts() []*UncertaintyAlert {
	uqm.mu.RLock()
	defer uqm.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]*UncertaintyAlert, len(uqm.uncertaintyAlerts))
	copy(result, uqm.uncertaintyAlerts)

	return result
}

// AcknowledgeUncertaintyAlert acknowledges an uncertainty alert
func (uqm *UncertaintyQuantificationMonitor) AcknowledgeUncertaintyAlert(alertID string) error {
	uqm.mu.Lock()
	defer uqm.mu.Unlock()

	for _, alert := range uqm.uncertaintyAlerts {
		if alert.ID == alertID {
			alert.Acknowledged = true
			alert.AcknowledgedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("uncertainty alert not found: %s", alertID)
}

// ResolveUncertaintyAlert resolves an uncertainty alert
func (uqm *UncertaintyQuantificationMonitor) ResolveUncertaintyAlert(alertID string) error {
	uqm.mu.Lock()
	defer uqm.mu.Unlock()

	for _, alert := range uqm.uncertaintyAlerts {
		if alert.ID == alertID {
			alert.Resolved = true
			alert.ResolvedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("uncertainty alert not found: %s", alertID)
}

// GetCalibrationReport returns a detailed calibration report
func (uqm *UncertaintyQuantificationMonitor) GetCalibrationReport() *CalibrationReport {
	uqm.mu.RLock()
	defer uqm.mu.RUnlock()

	report := &CalibrationReport{
		Timestamp:          time.Now(),
		TotalBins:          uqm.config.CalibrationBins,
		AnalyzedBins:       0,
		OverallCalibration: uqm.calculateOverallCalibration(),
		BinDetails:         make([]*CalibrationBinDetail, 0),
		Recommendations:    make([]string, 0),
	}

	// Analyze each calibration bin
	for _, point := range uqm.calibrationData {
		if point.SampleSize >= uqm.config.MinSamplesPerBin {
			report.AnalyzedBins++

			binDetail := &CalibrationBinDetail{
				ConfidenceBin:     point.ConfidenceBin,
				ConfidenceRange:   point.ConfidenceRange,
				PredictedAccuracy: point.PredictedAccuracy,
				ActualAccuracy:    point.ActualAccuracy,
				CalibrationError:  point.CalibrationError,
				SampleSize:        point.SampleSize,
				Status:            "good",
			}

			// Determine bin status
			if point.CalibrationError > uqm.config.CalibrationThreshold {
				binDetail.Status = "poor"
				report.Recommendations = append(report.Recommendations,
					fmt.Sprintf("Improve calibration in confidence range %s (error: %.3f)",
						point.ConfidenceRange, point.CalibrationError))
			} else if point.CalibrationError > uqm.config.CalibrationThreshold*0.5 {
				binDetail.Status = "warning"
			}

			report.BinDetails = append(report.BinDetails, binDetail)
		}
	}

	// Add overall recommendations
	if report.OverallCalibration < 0.7 {
		report.Recommendations = append(report.Recommendations,
			"Overall calibration is poor. Consider retraining the model or adjusting confidence thresholds.")
	}

	return report
}

// CalibrationReport represents a detailed calibration report
type CalibrationReport struct {
	Timestamp          time.Time               `json:"timestamp"`
	TotalBins          int                     `json:"total_bins"`
	AnalyzedBins       int                     `json:"analyzed_bins"`
	OverallCalibration float64                 `json:"overall_calibration"`
	BinDetails         []*CalibrationBinDetail `json:"bin_details"`
	Recommendations    []string                `json:"recommendations"`
}

// CalibrationBinDetail represents details for a calibration bin
type CalibrationBinDetail struct {
	ConfidenceBin     int     `json:"confidence_bin"`
	ConfidenceRange   string  `json:"confidence_range"`
	PredictedAccuracy float64 `json:"predicted_accuracy"`
	ActualAccuracy    float64 `json:"actual_accuracy"`
	CalibrationError  float64 `json:"calibration_error"`
	SampleSize        int     `json:"sample_size"`
	Status            string  `json:"status"` // "good", "warning", "poor"
}
