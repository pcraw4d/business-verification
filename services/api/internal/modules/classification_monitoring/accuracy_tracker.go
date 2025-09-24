package classification_monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AccuracyTracker tracks and monitors classification accuracy in real-time
type AccuracyTracker struct {
	config     *AccuracyConfig
	logger     *zap.Logger
	mu         sync.RWMutex
	startTime  time.Time
	metrics    map[string]*AccuracyMetrics
	historical []*HistoricalAccuracyPoint
	alerts     []*AccuracyAlert
	collectors []MetricCollector
}

// AccuracyConfig holds configuration for accuracy tracking
type AccuracyConfig struct {
	EnableRealTimeTracking     bool          `json:"enable_real_time_tracking"`
	EnableMisclassificationLog bool          `json:"enable_misclassification_log"`
	EnableTrendAnalysis        bool          `json:"enable_trend_analysis"`
	TargetAccuracy             float64       `json:"target_accuracy"`             // 0.90 (90%)
	CriticalAccuracyThreshold  float64       `json:"critical_accuracy_threshold"` // 0.85 (85%)
	AlertCooldownPeriod        time.Duration `json:"alert_cooldown_period"`
	MetricsRetentionPeriod     time.Duration `json:"metrics_retention_period"`
	SampleWindowSize           int           `json:"sample_window_size"`
	TrendWindowSize            int           `json:"trend_window_size"`
	EnableDimensionalAnalysis  bool          `json:"enable_dimensional_analysis"`
}

// AccuracyMetrics represents accuracy metrics for a specific dimension
type AccuracyMetrics struct {
	DimensionName          string                     `json:"dimension_name"`
	DimensionValue         string                     `json:"dimension_value"`
	TotalClassifications   int                        `json:"total_classifications"`
	CorrectClassifications int                        `json:"correct_classifications"`
	AccuracyScore          float64                    `json:"accuracy_score"`
	ConfidenceScore        float64                    `json:"confidence_score"`
	LastUpdated            time.Time                  `json:"last_updated"`
	WindowedAccuracy       []float64                  `json:"windowed_accuracy"`
	Misclassifications     []*MisclassificationRecord `json:"misclassifications"`
	TrendIndicator         string                     `json:"trend_indicator"` // improving, stable, declining
}

// MisclassificationRecord represents a misclassification incident
type MisclassificationRecord struct {
	ID                     string                 `json:"id"`
	Timestamp              time.Time              `json:"timestamp"`
	BusinessName           string                 `json:"business_name"`
	ExpectedClassification string                 `json:"expected_classification"`
	ActualClassification   string                 `json:"actual_classification"`
	ConfidenceScore        float64                `json:"confidence_score"`
	ClassificationMethod   string                 `json:"classification_method"`
	InputData              map[string]interface{} `json:"input_data"`
	ErrorType              string                 `json:"error_type"`
	Severity               string                 `json:"severity"`
	RootCause              string                 `json:"root_cause"`
	ActionRequired         bool                   `json:"action_required"`
}

// HistoricalAccuracyPoint represents a point in accuracy history
type HistoricalAccuracyPoint struct {
	Timestamp           time.Time              `json:"timestamp"`
	OverallAccuracy     float64                `json:"overall_accuracy"`
	SampleSize          int                    `json:"sample_size"`
	DimensionAccuracies map[string]float64     `json:"dimension_accuracies"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// AccuracyAlert represents an accuracy-related alert
type AccuracyAlert struct {
	ID             string     `json:"id"`
	Type           string     `json:"type"`     // threshold, trend, dimensional
	Severity       string     `json:"severity"` // low, medium, high, critical
	DimensionName  string     `json:"dimension_name"`
	DimensionValue string     `json:"dimension_value"`
	CurrentValue   float64    `json:"current_value"`
	ThresholdValue float64    `json:"threshold_value"`
	Message        string     `json:"message"`
	Timestamp      time.Time  `json:"timestamp"`
	Resolved       bool       `json:"resolved"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	Actions        []string   `json:"actions"`
}

// MetricCollector interface for collecting accuracy metrics
type MetricCollector interface {
	CollectMetrics(ctx context.Context) ([]*AccuracyMetrics, error)
	GetDimensions() []string
}

// ClassificationResult represents a classification result for accuracy tracking
type ClassificationResult struct {
	ID                     string                 `json:"id"`
	BusinessName           string                 `json:"business_name"`
	ActualClassification   string                 `json:"actual_classification"`
	ExpectedClassification *string                `json:"expected_classification,omitempty"`
	ConfidenceScore        float64                `json:"confidence_score"`
	ClassificationMethod   string                 `json:"classification_method"`
	Timestamp              time.Time              `json:"timestamp"`
	Metadata               map[string]interface{} `json:"metadata"`
	IsCorrect              *bool                  `json:"is_correct,omitempty"`
}

// NewAccuracyTracker creates a new accuracy tracker
func NewAccuracyTracker(config *AccuracyConfig, logger *zap.Logger) *AccuracyTracker {
	if config == nil {
		config = DefaultAccuracyConfig()
	}

	return &AccuracyTracker{
		config:     config,
		logger:     logger,
		startTime:  time.Now(),
		metrics:    make(map[string]*AccuracyMetrics),
		historical: make([]*HistoricalAccuracyPoint, 0),
		alerts:     make([]*AccuracyAlert, 0),
		collectors: make([]MetricCollector, 0),
	}
}

// DefaultAccuracyConfig returns default configuration
func DefaultAccuracyConfig() *AccuracyConfig {
	return &AccuracyConfig{
		EnableRealTimeTracking:     true,
		EnableMisclassificationLog: true,
		EnableTrendAnalysis:        true,
		TargetAccuracy:             0.90,
		CriticalAccuracyThreshold:  0.85,
		AlertCooldownPeriod:        15 * time.Minute,
		MetricsRetentionPeriod:     30 * 24 * time.Hour, // 30 days
		SampleWindowSize:           100,
		TrendWindowSize:            10,
		EnableDimensionalAnalysis:  true,
	}
}

// TrackClassification tracks a new classification result
func (at *AccuracyTracker) TrackClassification(ctx context.Context, result *ClassificationResult) error {
	at.mu.Lock()
	defer at.mu.Unlock()

	// Update overall metrics
	if err := at.updateOverallMetrics(result); err != nil {
		return fmt.Errorf("failed to update overall metrics: %w", err)
	}

	// Update dimensional metrics if enabled
	if at.config.EnableDimensionalAnalysis {
		if err := at.updateDimensionalMetrics(result); err != nil {
			at.logger.Error("Failed to update dimensional metrics", zap.Error(err))
		}
	}

	// Check for misclassification
	if result.IsCorrect != nil && !*result.IsCorrect {
		if err := at.logMisclassification(result); err != nil {
			at.logger.Error("Failed to log misclassification", zap.Error(err))
		}
	}

	// Check alert conditions
	if err := at.checkAlertConditions(); err != nil {
		at.logger.Error("Failed to check alert conditions", zap.Error(err))
	}

	return nil
}

// updateOverallMetrics updates overall accuracy metrics
func (at *AccuracyTracker) updateOverallMetrics(result *ClassificationResult) error {
	overall, exists := at.metrics["overall"]
	if !exists {
		overall = &AccuracyMetrics{
			DimensionName:      "overall",
			DimensionValue:     "all",
			WindowedAccuracy:   make([]float64, 0, at.config.SampleWindowSize),
			Misclassifications: make([]*MisclassificationRecord, 0),
		}
		at.metrics["overall"] = overall
	}

	overall.TotalClassifications++
	overall.LastUpdated = time.Now()

	if result.IsCorrect != nil {
		if *result.IsCorrect {
			overall.CorrectClassifications++
		}
		overall.AccuracyScore = float64(overall.CorrectClassifications) / float64(overall.TotalClassifications)
	}

	// Update windowed accuracy
	if result.IsCorrect != nil {
		accuracy := 0.0
		if *result.IsCorrect {
			accuracy = 1.0
		}

		overall.WindowedAccuracy = append(overall.WindowedAccuracy, accuracy)
		if len(overall.WindowedAccuracy) > at.config.SampleWindowSize {
			overall.WindowedAccuracy = overall.WindowedAccuracy[1:]
		}
	}

	// Update trend indicator
	overall.TrendIndicator = at.calculateTrendIndicator(overall.WindowedAccuracy)

	return nil
}

// updateDimensionalMetrics updates metrics for various dimensions
func (at *AccuracyTracker) updateDimensionalMetrics(result *ClassificationResult) error {
	dimensions := []struct {
		name  string
		value string
	}{
		{"method", result.ClassificationMethod},
		{"confidence_range", at.getConfidenceRange(result.ConfidenceScore)},
		{"time_of_day", at.getTimeOfDayCategory(result.Timestamp)},
	}

	for _, dim := range dimensions {
		key := fmt.Sprintf("%s:%s", dim.name, dim.value)

		metrics, exists := at.metrics[key]
		if !exists {
			metrics = &AccuracyMetrics{
				DimensionName:      dim.name,
				DimensionValue:     dim.value,
				WindowedAccuracy:   make([]float64, 0, at.config.SampleWindowSize),
				Misclassifications: make([]*MisclassificationRecord, 0),
			}
			at.metrics[key] = metrics
		}

		metrics.TotalClassifications++
		metrics.LastUpdated = time.Now()

		if result.IsCorrect != nil {
			if *result.IsCorrect {
				metrics.CorrectClassifications++
			}
			metrics.AccuracyScore = float64(metrics.CorrectClassifications) / float64(metrics.TotalClassifications)

			// Update windowed accuracy
			accuracy := 0.0
			if *result.IsCorrect {
				accuracy = 1.0
			}

			metrics.WindowedAccuracy = append(metrics.WindowedAccuracy, accuracy)
			if len(metrics.WindowedAccuracy) > at.config.SampleWindowSize {
				metrics.WindowedAccuracy = metrics.WindowedAccuracy[1:]
			}
		}

		// Update trend indicator
		metrics.TrendIndicator = at.calculateTrendIndicator(metrics.WindowedAccuracy)
	}

	return nil
}

// logMisclassification logs a misclassification incident
func (at *AccuracyTracker) logMisclassification(result *ClassificationResult) error {
	if !at.config.EnableMisclassificationLog {
		return nil
	}

	record := &MisclassificationRecord{
		ID:                     fmt.Sprintf("misclass_%d", time.Now().UnixNano()),
		Timestamp:              result.Timestamp,
		BusinessName:           result.BusinessName,
		ActualClassification:   result.ActualClassification,
		ExpectedClassification: *result.ExpectedClassification,
		ConfidenceScore:        result.ConfidenceScore,
		ClassificationMethod:   result.ClassificationMethod,
		InputData:              result.Metadata,
		ErrorType:              at.classifyErrorType(result),
		Severity:               at.calculateMisclassificationSeverity(result),
		ActionRequired:         at.requiresAction(result),
	}

	// Add to overall metrics
	if overall, exists := at.metrics["overall"]; exists {
		overall.Misclassifications = append(overall.Misclassifications, record)

		// Keep only recent misclassifications
		if len(overall.Misclassifications) > 1000 {
			overall.Misclassifications = overall.Misclassifications[500:]
		}
	}

	at.logger.Warn("Misclassification detected",
		zap.String("id", record.ID),
		zap.String("business", record.BusinessName),
		zap.String("expected", record.ExpectedClassification),
		zap.String("actual", record.ActualClassification),
		zap.Float64("confidence", record.ConfidenceScore),
		zap.String("method", record.ClassificationMethod),
		zap.String("severity", record.Severity))

	return nil
}

// checkAlertConditions checks for conditions that should trigger alerts
func (at *AccuracyTracker) checkAlertConditions() error {
	now := time.Now()

	// Check overall accuracy threshold
	if overall, exists := at.metrics["overall"]; exists && overall.TotalClassifications >= 10 {
		if overall.AccuracyScore < at.config.CriticalAccuracyThreshold {
			alert := &AccuracyAlert{
				ID:             fmt.Sprintf("alert_%d", now.UnixNano()),
				Type:           "threshold",
				Severity:       "critical",
				DimensionName:  "overall",
				DimensionValue: "all",
				CurrentValue:   overall.AccuracyScore,
				ThresholdValue: at.config.CriticalAccuracyThreshold,
				Message:        fmt.Sprintf("Overall accuracy %.2f%% is below critical threshold %.2f%%", overall.AccuracyScore*100, at.config.CriticalAccuracyThreshold*100),
				Timestamp:      now,
				Actions:        []string{"investigate_recent_changes", "check_data_quality", "review_classification_logic"},
			}
			at.addAlert(alert)
		} else if overall.AccuracyScore < at.config.TargetAccuracy {
			alert := &AccuracyAlert{
				ID:             fmt.Sprintf("alert_%d", now.UnixNano()),
				Type:           "threshold",
				Severity:       "medium",
				DimensionName:  "overall",
				DimensionValue: "all",
				CurrentValue:   overall.AccuracyScore,
				ThresholdValue: at.config.TargetAccuracy,
				Message:        fmt.Sprintf("Overall accuracy %.2f%% is below target %.2f%%", overall.AccuracyScore*100, at.config.TargetAccuracy*100),
				Timestamp:      now,
				Actions:        []string{"monitor_trend", "analyze_patterns", "consider_optimization"},
			}
			at.addAlert(alert)
		}
	}

	// Check dimensional accuracy
	for key, metrics := range at.metrics {
		if key == "overall" || metrics.TotalClassifications < 5 {
			continue
		}

		if metrics.AccuracyScore < at.config.CriticalAccuracyThreshold {
			alert := &AccuracyAlert{
				ID:             fmt.Sprintf("alert_%d_%s", now.UnixNano(), key),
				Type:           "dimensional",
				Severity:       "high",
				DimensionName:  metrics.DimensionName,
				DimensionValue: metrics.DimensionValue,
				CurrentValue:   metrics.AccuracyScore,
				ThresholdValue: at.config.CriticalAccuracyThreshold,
				Message:        fmt.Sprintf("Accuracy for %s:%s is %.2f%% (below critical threshold)", metrics.DimensionName, metrics.DimensionValue, metrics.AccuracyScore*100),
				Timestamp:      now,
				Actions:        []string{"investigate_dimension_specific_issues", "check_training_data", "review_feature_engineering"},
			}
			at.addAlert(alert)
		}
	}

	// Check trend-based alerts
	if at.config.EnableTrendAnalysis {
		at.checkTrendAlerts()
	}

	return nil
}

// checkTrendAlerts checks for trend-based alert conditions
func (at *AccuracyTracker) checkTrendAlerts() {
	for key, metrics := range at.metrics {
		if len(metrics.WindowedAccuracy) < at.config.TrendWindowSize {
			continue
		}

		trend := at.calculateTrend(metrics.WindowedAccuracy)
		if trend < -0.05 { // Declining trend of more than 5%
			alert := &AccuracyAlert{
				ID:             fmt.Sprintf("trend_alert_%d_%s", time.Now().UnixNano(), key),
				Type:           "trend",
				Severity:       "medium",
				DimensionName:  metrics.DimensionName,
				DimensionValue: metrics.DimensionValue,
				CurrentValue:   metrics.AccuracyScore,
				ThresholdValue: 0.0,
				Message:        fmt.Sprintf("Declining accuracy trend detected for %s:%s (%.2f%% decrease)", metrics.DimensionName, metrics.DimensionValue, trend*100),
				Timestamp:      time.Now(),
				Actions:        []string{"analyze_recent_changes", "check_data_drift", "review_model_performance"},
			}
			at.addAlert(alert)
		}
	}
}

// Helper methods

func (at *AccuracyTracker) getConfidenceRange(confidence float64) string {
	switch {
	case confidence >= 0.9:
		return "high"
	case confidence >= 0.7:
		return "medium"
	case confidence >= 0.5:
		return "low"
	default:
		return "very_low"
	}
}

func (at *AccuracyTracker) getTimeOfDayCategory(timestamp time.Time) string {
	hour := timestamp.Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 18:
		return "afternoon"
	case hour >= 18 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}

func (at *AccuracyTracker) calculateTrendIndicator(values []float64) string {
	if len(values) < 5 {
		return "insufficient_data"
	}

	trend := at.calculateTrend(values)
	switch {
	case trend > 0.02:
		return "improving"
	case trend < -0.02:
		return "declining"
	default:
		return "stable"
	}
}

func (at *AccuracyTracker) calculateTrend(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}

	n := float64(len(values))
	sumX := n * (n - 1) / 2
	sumY := 0.0
	sumXY := 0.0
	sumX2 := (n - 1) * n * (2*n - 1) / 6

	for i, y := range values {
		x := float64(i)
		sumY += y
		sumXY += x * y
	}

	// Linear regression slope
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	return slope
}

func (at *AccuracyTracker) classifyErrorType(result *ClassificationResult) string {
	// Analyze the misclassification to determine type
	if result.ConfidenceScore < 0.5 {
		return "low_confidence"
	} else if result.ConfidenceScore > 0.8 {
		return "high_confidence_error"
	}
	return "medium_confidence_error"
}

func (at *AccuracyTracker) calculateMisclassificationSeverity(result *ClassificationResult) string {
	// Calculate severity based on confidence and business impact
	if result.ConfidenceScore > 0.9 {
		return "high" // High confidence but wrong
	} else if result.ConfidenceScore > 0.7 {
		return "medium"
	}
	return "low"
}

func (at *AccuracyTracker) requiresAction(result *ClassificationResult) bool {
	return result.ConfidenceScore > 0.8 // High confidence errors require action
}

func (at *AccuracyTracker) addAlert(alert *AccuracyAlert) {
	// Check for recent similar alerts to avoid spam
	for _, existing := range at.alerts {
		if existing.Type == alert.Type &&
			existing.DimensionName == alert.DimensionName &&
			existing.DimensionValue == alert.DimensionValue &&
			!existing.Resolved &&
			time.Since(existing.Timestamp) < at.config.AlertCooldownPeriod {
			return // Don't add duplicate alert
		}
	}

	at.alerts = append(at.alerts, alert)

	// Keep only recent alerts
	if len(at.alerts) > 1000 {
		at.alerts = at.alerts[500:]
	}

	at.logger.Warn("Accuracy alert triggered",
		zap.String("type", alert.Type),
		zap.String("severity", alert.Severity),
		zap.String("dimension", fmt.Sprintf("%s:%s", alert.DimensionName, alert.DimensionValue)),
		zap.Float64("current_value", alert.CurrentValue),
		zap.String("message", alert.Message))
}

// Public methods for accessing tracker data

// GetAccuracyMetrics returns current accuracy metrics
func (at *AccuracyTracker) GetAccuracyMetrics() map[string]*AccuracyMetrics {
	at.mu.RLock()
	defer at.mu.RUnlock()

	result := make(map[string]*AccuracyMetrics)
	for k, v := range at.metrics {
		result[k] = v
	}
	return result
}

// GetOverallAccuracy returns the overall accuracy score
func (at *AccuracyTracker) GetOverallAccuracy() float64 {
	at.mu.RLock()
	defer at.mu.RUnlock()

	if overall, exists := at.metrics["overall"]; exists {
		return overall.AccuracyScore
	}
	return 0.0
}

// GetActiveAlerts returns currently active alerts
func (at *AccuracyTracker) GetActiveAlerts() []*AccuracyAlert {
	at.mu.RLock()
	defer at.mu.RUnlock()

	activeAlerts := make([]*AccuracyAlert, 0)
	for _, alert := range at.alerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}
	return activeAlerts
}

// GetMisclassifications returns recent misclassifications
func (at *AccuracyTracker) GetMisclassifications(limit int) []*MisclassificationRecord {
	at.mu.RLock()
	defer at.mu.RUnlock()

	if overall, exists := at.metrics["overall"]; exists {
		misclassifications := overall.Misclassifications
		if len(misclassifications) <= limit {
			return misclassifications
		}
		return misclassifications[len(misclassifications)-limit:]
	}
	return []*MisclassificationRecord{}
}

// ResolveAlert marks an alert as resolved
func (at *AccuracyTracker) ResolveAlert(alertID string) error {
	at.mu.Lock()
	defer at.mu.Unlock()

	for _, alert := range at.alerts {
		if alert.ID == alertID && !alert.Resolved {
			alert.Resolved = true
			now := time.Now()
			alert.ResolvedAt = &now

			at.logger.Info("Alert resolved",
				zap.String("alert_id", alertID),
				zap.String("type", alert.Type),
				zap.String("dimension", fmt.Sprintf("%s:%s", alert.DimensionName, alert.DimensionValue)))

			return nil
		}
	}

	return fmt.Errorf("alert not found or already resolved: %s", alertID)
}

// AddMetricCollector adds a metric collector
func (at *AccuracyTracker) AddMetricCollector(collector MetricCollector) {
	at.mu.Lock()
	defer at.mu.Unlock()
	at.collectors = append(at.collectors, collector)
}

// CollectMetrics collects metrics from all registered collectors
func (at *AccuracyTracker) CollectMetrics(ctx context.Context) error {
	for _, collector := range at.collectors {
		metrics, err := collector.CollectMetrics(ctx)
		if err != nil {
			at.logger.Error("Failed to collect metrics", zap.Error(err))
			continue
		}

		at.mu.Lock()
		for _, metric := range metrics {
			key := fmt.Sprintf("%s:%s", metric.DimensionName, metric.DimensionValue)
			at.metrics[key] = metric
		}
		at.mu.Unlock()
	}
	return nil
}

// CreateSnapshot creates a historical snapshot of current metrics
func (at *AccuracyTracker) CreateSnapshot() *HistoricalAccuracyPoint {
	at.mu.RLock()
	defer at.mu.RUnlock()

	point := &HistoricalAccuracyPoint{
		Timestamp:           time.Now(),
		DimensionAccuracies: make(map[string]float64),
		Metadata:            make(map[string]interface{}),
	}

	if overall, exists := at.metrics["overall"]; exists {
		point.OverallAccuracy = overall.AccuracyScore
		point.SampleSize = overall.TotalClassifications
	}

	for key, metrics := range at.metrics {
		if key != "overall" {
			point.DimensionAccuracies[key] = metrics.AccuracyScore
		}
	}

	point.Metadata["total_dimensions"] = len(at.metrics) - 1 // Exclude overall
	point.Metadata["active_alerts"] = len(at.GetActiveAlerts())

	at.historical = append(at.historical, point)

	// Keep historical data within retention period
	cutoff := time.Now().Add(-at.config.MetricsRetentionPeriod)
	for i, hp := range at.historical {
		if hp.Timestamp.After(cutoff) {
			at.historical = at.historical[i:]
			break
		}
	}

	return point
}

// GetHistoricalData returns historical accuracy data
func (at *AccuracyTracker) GetHistoricalData(since time.Time) []*HistoricalAccuracyPoint {
	at.mu.RLock()
	defer at.mu.RUnlock()

	result := make([]*HistoricalAccuracyPoint, 0)
	for _, point := range at.historical {
		if point.Timestamp.After(since) {
			result = append(result, point)
		}
	}
	return result
}
