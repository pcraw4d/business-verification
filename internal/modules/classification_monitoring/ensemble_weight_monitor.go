package classification_monitoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// EnsembleWeightMonitor provides comprehensive monitoring for ensemble weight distribution
type EnsembleWeightMonitor struct {
	config *EnsembleWeightMonitorConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// Weight tracking
	weightHistory  map[string][]*WeightDataPoint
	currentWeights map[string]float64
	weightChanges  []*WeightChangeEvent
	weightAlerts   []*WeightDistributionAlert

	// Performance tracking
	performanceTracker *EnsemblePerformanceTracker
	weightAnalyzer     *WeightDistributionAnalyzer
	stabilityMonitor   *WeightStabilityMonitor
}

// EnsembleWeightMonitorConfig holds configuration for ensemble weight monitoring
type EnsembleWeightMonitorConfig struct {
	// Weight tracking
	WeightTrackingEnabled bool          `json:"weight_tracking_enabled"`
	WeightHistorySize     int           `json:"weight_history_size"`
	WeightChangeThreshold float64       `json:"weight_change_threshold"`
	WeightStabilityWindow time.Duration `json:"weight_stability_window"`

	// Distribution analysis
	DistributionAnalysisEnabled bool    `json:"distribution_analysis_enabled"`
	ImbalanceThreshold          float64 `json:"imbalance_threshold"`
	ConcentrationThreshold      float64 `json:"concentration_threshold"`

	// Performance correlation
	PerformanceCorrelationEnabled bool `json:"performance_correlation_enabled"`
	CorrelationWindowSize         int  `json:"correlation_window_size"`
	MinSamplesForCorrelation      int  `json:"min_samples_for_correlation"`

	// Alerting
	AlertingEnabled      bool          `json:"alerting_enabled"`
	AlertCooldownPeriod  time.Duration `json:"alert_cooldown_period"`
	MaxAlertsPerMethod   int           `json:"max_alerts_per_method"`
	AlertRetentionPeriod time.Duration `json:"alert_retention_period"`

	// Monitoring intervals
	MonitoringInterval     time.Duration `json:"monitoring_interval"`
	AnalysisInterval       time.Duration `json:"analysis_interval"`
	StabilityCheckInterval time.Duration `json:"stability_check_interval"`
}

// WeightDataPoint represents a weight measurement at a specific time
type WeightDataPoint struct {
	Timestamp   time.Time `json:"timestamp"`
	MethodName  string    `json:"method_name"`
	Weight      float64   `json:"weight"`
	Reason      string    `json:"reason"`
	Performance float64   `json:"performance,omitempty"`
	Accuracy    float64   `json:"accuracy,omitempty"`
	Latency     float64   `json:"latency,omitempty"`
	SampleSize  int       `json:"sample_size"`
}

// WeightChangeEvent represents a significant weight change
type WeightChangeEvent struct {
	ID                string    `json:"id"`
	MethodName        string    `json:"method_name"`
	PreviousWeight    float64   `json:"previous_weight"`
	NewWeight         float64   `json:"new_weight"`
	ChangeAmount      float64   `json:"change_amount"`
	ChangePercent     float64   `json:"change_percent"`
	Reason            string    `json:"reason"`
	Timestamp         time.Time `json:"timestamp"`
	PerformanceImpact float64   `json:"performance_impact,omitempty"`
}

// WeightDistributionAlert represents an alert about weight distribution issues
type WeightDistributionAlert struct {
	ID             string    `json:"id"`
	AlertType      string    `json:"alert_type"` // "imbalance", "concentration", "instability", "performance_correlation"
	Severity       string    `json:"severity"`   // "warning", "critical"
	MethodName     string    `json:"method_name,omitempty"`
	Message        string    `json:"message"`
	Details        string    `json:"details"`
	Timestamp      time.Time `json:"timestamp"`
	Acknowledged   bool      `json:"acknowledged"`
	AcknowledgedAt time.Time `json:"acknowledged_at,omitempty"`
	Resolved       bool      `json:"resolved"`
	ResolvedAt     time.Time `json:"resolved_at,omitempty"`
}

// EnsemblePerformanceTracker tracks performance metrics for ensemble methods
type EnsemblePerformanceTracker struct {
	config *EnsembleWeightMonitorConfig
	logger *zap.Logger
}

// WeightDistributionAnalyzer analyzes weight distribution patterns
type WeightDistributionAnalyzer struct {
	config *EnsembleWeightMonitorConfig
	logger *zap.Logger
}

// WeightStabilityMonitor monitors weight stability over time
type WeightStabilityMonitor struct {
	config *EnsembleWeightMonitorConfig
	logger *zap.Logger
}

// NewEnsembleWeightMonitor creates a new ensemble weight monitor
func NewEnsembleWeightMonitor(config *EnsembleWeightMonitorConfig, logger *zap.Logger) *EnsembleWeightMonitor {
	if config == nil {
		config = DefaultEnsembleWeightMonitorConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &EnsembleWeightMonitor{
		config:         config,
		logger:         logger,
		weightHistory:  make(map[string][]*WeightDataPoint),
		currentWeights: make(map[string]float64),
		weightChanges:  make([]*WeightChangeEvent, 0),
		weightAlerts:   make([]*WeightDistributionAlert, 0),
		performanceTracker: &EnsemblePerformanceTracker{
			config: config,
			logger: logger,
		},
		weightAnalyzer: &WeightDistributionAnalyzer{
			config: config,
			logger: logger,
		},
		stabilityMonitor: &WeightStabilityMonitor{
			config: config,
			logger: logger,
		},
	}
}

// DefaultEnsembleWeightMonitorConfig returns default configuration
func DefaultEnsembleWeightMonitorConfig() *EnsembleWeightMonitorConfig {
	return &EnsembleWeightMonitorConfig{
		WeightTrackingEnabled:         true,
		WeightHistorySize:             1000,
		WeightChangeThreshold:         0.1, // 10% change threshold
		WeightStabilityWindow:         1 * time.Hour,
		DistributionAnalysisEnabled:   true,
		ImbalanceThreshold:            0.7, // 70% max weight for single method
		ConcentrationThreshold:        0.8, // 80% concentration threshold
		PerformanceCorrelationEnabled: true,
		CorrelationWindowSize:         100,
		MinSamplesForCorrelation:      50,
		AlertingEnabled:               true,
		AlertCooldownPeriod:           15 * time.Minute,
		MaxAlertsPerMethod:            20,
		AlertRetentionPeriod:          24 * time.Hour,
		MonitoringInterval:            1 * time.Minute,
		AnalysisInterval:              5 * time.Minute,
		StabilityCheckInterval:        10 * time.Minute,
	}
}

// TrackWeightChange tracks a weight change for a method
func (ewm *EnsembleWeightMonitor) TrackWeightChange(
	ctx context.Context,
	methodName string,
	previousWeight, newWeight float64,
	reason string,
	performance float64,
) error {
	ewm.mu.Lock()
	defer ewm.mu.Unlock()

	if !ewm.config.WeightTrackingEnabled {
		return nil
	}

	// Calculate change metrics
	changeAmount := newWeight - previousWeight
	changePercent := 0.0
	if previousWeight > 0 {
		changePercent = (changeAmount / previousWeight) * 100
	}

	// Create weight data point
	weightPoint := &WeightDataPoint{
		Timestamp:   time.Now(),
		MethodName:  methodName,
		Weight:      newWeight,
		Reason:      reason,
		Performance: performance,
		SampleSize:  1,
	}

	// Add to weight history
	ewm.weightHistory[methodName] = append(ewm.weightHistory[methodName], weightPoint)

	// Maintain history size
	if len(ewm.weightHistory[methodName]) > ewm.config.WeightHistorySize {
		ewm.weightHistory[methodName] = ewm.weightHistory[methodName][1:]
	}

	// Update current weights
	ewm.currentWeights[methodName] = newWeight

	// Check for significant changes
	if math.Abs(changePercent) >= ewm.config.WeightChangeThreshold*100 {
		ewm.recordWeightChange(methodName, previousWeight, newWeight, changeAmount, changePercent, reason, performance)
	}

	// Analyze weight distribution
	if ewm.config.DistributionAnalysisEnabled {
		ewm.analyzeWeightDistribution()
	}

	// Check weight stability
	if ewm.config.WeightTrackingEnabled {
		ewm.checkWeightStability(methodName)
	}

	// Analyze performance correlation
	if ewm.config.PerformanceCorrelationEnabled {
		ewm.analyzePerformanceCorrelation(methodName)
	}

	ewm.logger.Debug("Weight change tracked",
		zap.String("method_name", methodName),
		zap.Float64("previous_weight", previousWeight),
		zap.Float64("new_weight", newWeight),
		zap.Float64("change_percent", changePercent),
		zap.String("reason", reason))

	return nil
}

// recordWeightChange records a significant weight change event
func (ewm *EnsembleWeightMonitor) recordWeightChange(
	methodName string,
	previousWeight, newWeight, changeAmount, changePercent float64,
	reason string,
	performance float64,
) {
	changeEvent := &WeightChangeEvent{
		ID:                fmt.Sprintf("weight_change_%s_%d", methodName, time.Now().Unix()),
		MethodName:        methodName,
		PreviousWeight:    previousWeight,
		NewWeight:         newWeight,
		ChangeAmount:      changeAmount,
		ChangePercent:     changePercent,
		Reason:            reason,
		Timestamp:         time.Now(),
		PerformanceImpact: performance,
	}

	ewm.weightChanges = append(ewm.weightChanges, changeEvent)

	// Maintain change history size
	if len(ewm.weightChanges) > ewm.config.WeightHistorySize {
		ewm.weightChanges = ewm.weightChanges[1:]
	}

	ewm.logger.Info("Significant weight change recorded",
		zap.String("method_name", methodName),
		zap.Float64("change_percent", changePercent),
		zap.String("reason", reason))
}

// analyzeWeightDistribution analyzes the current weight distribution
func (ewm *EnsembleWeightMonitor) analyzeWeightDistribution() {
	if len(ewm.currentWeights) == 0 {
		return
	}

	// Calculate distribution metrics
	totalWeight := 0.0
	maxWeight := 0.0
	maxMethod := ""
	weightSum := 0.0

	for method, weight := range ewm.currentWeights {
		totalWeight += weight
		weightSum += weight * weight // For concentration calculation
		if weight > maxWeight {
			maxWeight = weight
			maxMethod = method
		}
	}

	// Check for imbalance (single method dominating)
	if maxWeight/totalWeight > ewm.config.ImbalanceThreshold {
		ewm.createWeightAlert("imbalance", "warning", maxMethod,
			fmt.Sprintf("Method %s has %.1f%% of total weight (threshold: %.1f%%)",
				maxMethod, (maxWeight/totalWeight)*100, ewm.config.ImbalanceThreshold*100),
			fmt.Sprintf("Current weight: %.3f, Total weight: %.3f", maxWeight, totalWeight))
	}

	// Check for concentration (weights too concentrated)
	concentration := weightSum / (totalWeight * totalWeight)
	if concentration > ewm.config.ConcentrationThreshold {
		ewm.createWeightAlert("concentration", "warning", "",
			fmt.Sprintf("Weight concentration is %.1f%% (threshold: %.1f%%)",
				concentration*100, ewm.config.ConcentrationThreshold*100),
			fmt.Sprintf("Concentration index: %.3f", concentration))
	}

	// Check for zero weights (methods not being used)
	zeroWeightMethods := make([]string, 0)
	for method, weight := range ewm.currentWeights {
		if weight <= 0.001 { // Near zero
			zeroWeightMethods = append(zeroWeightMethods, method)
		}
	}

	if len(zeroWeightMethods) > 0 {
		ewm.createWeightAlert("zero_weights", "warning", "",
			fmt.Sprintf("%d methods have zero or near-zero weights", len(zeroWeightMethods)),
			fmt.Sprintf("Methods: %v", zeroWeightMethods))
	}
}

// checkWeightStability checks weight stability for a method
func (ewm *EnsembleWeightMonitor) checkWeightStability(methodName string) {
	history, exists := ewm.weightHistory[methodName]
	if !exists || len(history) < 10 {
		return
	}

	// Check recent weight stability
	recentWindow := ewm.config.WeightStabilityWindow
	cutoff := time.Now().Add(-recentWindow)
	recentWeights := make([]float64, 0)

	for _, point := range history {
		if point.Timestamp.After(cutoff) {
			recentWeights = append(recentWeights, point.Weight)
		}
	}

	if len(recentWeights) < 5 {
		return
	}

	// Calculate weight variance
	variance := ewm.calculateVariance(recentWeights)
	meanWeight := ewm.calculateMean(recentWeights)
	coefficientOfVariation := 0.0
	if meanWeight > 0 {
		coefficientOfVariation = math.Sqrt(variance) / meanWeight
	}

	// Check for instability (high coefficient of variation)
	if coefficientOfVariation > 0.2 { // 20% coefficient of variation
		ewm.createWeightAlert("instability", "warning", methodName,
			fmt.Sprintf("Method %s shows weight instability (CV: %.1f%%)",
				methodName, coefficientOfVariation*100),
			fmt.Sprintf("Mean weight: %.3f, Std dev: %.3f", meanWeight, math.Sqrt(variance)))
	}
}

// analyzePerformanceCorrelation analyzes correlation between weight changes and performance
func (ewm *EnsembleWeightMonitor) analyzePerformanceCorrelation(methodName string) {
	history, exists := ewm.weightHistory[methodName]
	if !exists || len(history) < ewm.config.MinSamplesForCorrelation {
		return
	}

	// Get recent data for correlation analysis
	recentData := history
	if len(history) > ewm.config.CorrelationWindowSize {
		recentData = history[len(history)-ewm.config.CorrelationWindowSize:]
	}

	// Extract weights and performance values
	weights := make([]float64, len(recentData))
	performances := make([]float64, len(recentData))

	for i, point := range recentData {
		weights[i] = point.Weight
		performances[i] = point.Performance
	}

	// Calculate correlation coefficient
	correlation := ewm.calculateCorrelation(weights, performances)

	// Check for negative correlation (weight increases but performance decreases)
	if correlation < -0.5 { // Strong negative correlation
		ewm.createWeightAlert("negative_correlation", "critical", methodName,
			fmt.Sprintf("Method %s shows negative weight-performance correlation (%.2f)",
				methodName, correlation),
			fmt.Sprintf("Weight and performance are inversely related"))
	}

	// Check for weak correlation (weight changes don't affect performance)
	if math.Abs(correlation) < 0.1 { // Very weak correlation
		ewm.createWeightAlert("weak_correlation", "warning", methodName,
			fmt.Sprintf("Method %s shows weak weight-performance correlation (%.2f)",
				methodName, correlation),
			fmt.Sprintf("Weight changes may not be improving performance"))
	}
}

// createWeightAlert creates a weight distribution alert
func (ewm *EnsembleWeightMonitor) createWeightAlert(alertType, severity, methodName, message, details string) {
	if !ewm.config.AlertingEnabled {
		return
	}

	// Check cooldown period
	if methodName != "" {
		lastAlert := ewm.getLastAlertForMethod(methodName, alertType)
		if lastAlert != nil && time.Since(lastAlert.Timestamp) < ewm.config.AlertCooldownPeriod {
			return
		}
	}

	// Check alert limits
	alertCount := ewm.getAlertCountForMethod(methodName)
	if alertCount >= ewm.config.MaxAlertsPerMethod {
		return
	}

	alert := &WeightDistributionAlert{
		ID:           fmt.Sprintf("weight_alert_%s_%d", alertType, time.Now().Unix()),
		AlertType:    alertType,
		Severity:     severity,
		MethodName:   methodName,
		Message:      message,
		Details:      details,
		Timestamp:    time.Now(),
		Acknowledged: false,
		Resolved:     false,
	}

	ewm.weightAlerts = append(ewm.weightAlerts, alert)

	// Clean up old alerts
	ewm.cleanupOldAlerts()

	ewm.logger.Warn("Weight distribution alert created",
		zap.String("alert_type", alertType),
		zap.String("severity", severity),
		zap.String("method_name", methodName),
		zap.String("message", message))
}

// getLastAlertForMethod gets the last alert for a specific method and type
func (ewm *EnsembleWeightMonitor) getLastAlertForMethod(methodName, alertType string) *WeightDistributionAlert {
	for i := len(ewm.weightAlerts) - 1; i >= 0; i-- {
		alert := ewm.weightAlerts[i]
		if alert.MethodName == methodName && alert.AlertType == alertType {
			return alert
		}
	}
	return nil
}

// getAlertCountForMethod gets the alert count for a specific method
func (ewm *EnsembleWeightMonitor) getAlertCountForMethod(methodName string) int {
	count := 0
	for _, alert := range ewm.weightAlerts {
		if alert.MethodName == methodName {
			count++
		}
	}
	return count
}

// cleanupOldAlerts removes old alerts beyond retention period
func (ewm *EnsembleWeightMonitor) cleanupOldAlerts() {
	cutoff := time.Now().Add(-ewm.config.AlertRetentionPeriod)
	var validAlerts []*WeightDistributionAlert

	for _, alert := range ewm.weightAlerts {
		if alert.Timestamp.After(cutoff) {
			validAlerts = append(validAlerts, alert)
		}
	}

	ewm.weightAlerts = validAlerts
}

// calculateVariance calculates variance of a slice of float64 values
func (ewm *EnsembleWeightMonitor) calculateVariance(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	mean := ewm.calculateMean(values)
	sumSquaredDiffs := 0.0

	for _, value := range values {
		diff := value - mean
		sumSquaredDiffs += diff * diff
	}

	return sumSquaredDiffs / float64(len(values)-1)
}

// calculateMean calculates mean of a slice of float64 values
func (ewm *EnsembleWeightMonitor) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}

// calculateCorrelation calculates Pearson correlation coefficient
func (ewm *EnsembleWeightMonitor) calculateCorrelation(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0.0
	}

	n := len(x)
	meanX := ewm.calculateMean(x)
	meanY := ewm.calculateMean(y)

	numerator := 0.0
	sumXSquared := 0.0
	sumYSquared := 0.0

	for i := 0; i < n; i++ {
		dx := x[i] - meanX
		dy := y[i] - meanY
		numerator += dx * dy
		sumXSquared += dx * dx
		sumYSquared += dy * dy
	}

	denominator := math.Sqrt(sumXSquared * sumYSquared)
	if denominator == 0 {
		return 0.0
	}

	return numerator / denominator
}

// GetCurrentWeights returns current weight distribution
func (ewm *EnsembleWeightMonitor) GetCurrentWeights() map[string]float64 {
	ewm.mu.RLock()
	defer ewm.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]float64)
	for method, weight := range ewm.currentWeights {
		result[method] = weight
	}

	return result
}

// GetWeightHistory returns weight history for a method
func (ewm *EnsembleWeightMonitor) GetWeightHistory(methodName string) ([]*WeightDataPoint, error) {
	ewm.mu.RLock()
	defer ewm.mu.RUnlock()

	history, exists := ewm.weightHistory[methodName]
	if !exists {
		return nil, fmt.Errorf("weight history not found for method: %s", methodName)
	}

	// Return a copy to avoid race conditions
	result := make([]*WeightDataPoint, len(history))
	copy(result, history)

	return result, nil
}

// GetWeightChanges returns recent weight change events
func (ewm *EnsembleWeightMonitor) GetWeightChanges() []*WeightChangeEvent {
	ewm.mu.RLock()
	defer ewm.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]*WeightChangeEvent, len(ewm.weightChanges))
	copy(result, ewm.weightChanges)

	return result
}

// GetWeightAlerts returns all weight distribution alerts
func (ewm *EnsembleWeightMonitor) GetWeightAlerts() []*WeightDistributionAlert {
	ewm.mu.RLock()
	defer ewm.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]*WeightDistributionAlert, len(ewm.weightAlerts))
	copy(result, ewm.weightAlerts)

	return result
}

// GetWeightDistributionMetrics returns comprehensive weight distribution metrics
func (ewm *EnsembleWeightMonitor) GetWeightDistributionMetrics() *WeightDistributionMetrics {
	ewm.mu.RLock()
	defer ewm.mu.RUnlock()

	metrics := &WeightDistributionMetrics{
		Timestamp:          time.Now(),
		TotalMethods:       len(ewm.currentWeights),
		CurrentWeights:     make(map[string]float64),
		WeightStatistics:   make(map[string]*WeightStatistics),
		DistributionHealth: "healthy",
		AlertsCount:        len(ewm.weightAlerts),
		ChangesCount:       len(ewm.weightChanges),
	}

	// Copy current weights
	for method, weight := range ewm.currentWeights {
		metrics.CurrentWeights[method] = weight
	}

	// Calculate statistics for each method
	for method, history := range ewm.weightHistory {
		if len(history) > 0 {
			stats := ewm.calculateWeightStatistics(history)
			metrics.WeightStatistics[method] = stats
		}
	}

	// Determine distribution health
	metrics.DistributionHealth = ewm.determineDistributionHealth()

	return metrics
}

// WeightDistributionMetrics represents comprehensive weight distribution metrics
type WeightDistributionMetrics struct {
	Timestamp          time.Time                    `json:"timestamp"`
	TotalMethods       int                          `json:"total_methods"`
	CurrentWeights     map[string]float64           `json:"current_weights"`
	WeightStatistics   map[string]*WeightStatistics `json:"weight_statistics"`
	DistributionHealth string                       `json:"distribution_health"`
	AlertsCount        int                          `json:"alerts_count"`
	ChangesCount       int                          `json:"changes_count"`
}

// WeightStatistics represents statistics for a method's weight history
type WeightStatistics struct {
	MethodName             string    `json:"method_name"`
	CurrentWeight          float64   `json:"current_weight"`
	MeanWeight             float64   `json:"mean_weight"`
	MinWeight              float64   `json:"min_weight"`
	MaxWeight              float64   `json:"max_weight"`
	WeightVariance         float64   `json:"weight_variance"`
	WeightStdDev           float64   `json:"weight_std_dev"`
	CoefficientOfVariation float64   `json:"coefficient_of_variation"`
	StabilityScore         float64   `json:"stability_score"`
	ChangeCount            int       `json:"change_count"`
	LastChangeTime         time.Time `json:"last_change_time"`
}

// calculateWeightStatistics calculates statistics for a method's weight history
func (ewm *EnsembleWeightMonitor) calculateWeightStatistics(history []*WeightDataPoint) *WeightStatistics {
	if len(history) == 0 {
		return &WeightStatistics{}
	}

	weights := make([]float64, len(history))
	for i, point := range history {
		weights[i] = point.Weight
	}

	mean := ewm.calculateMean(weights)
	variance := ewm.calculateVariance(weights)
	stdDev := math.Sqrt(variance)
	coefficientOfVariation := 0.0
	if mean > 0 {
		coefficientOfVariation = stdDev / mean
	}

	minWeight := weights[0]
	maxWeight := weights[0]
	for _, weight := range weights {
		if weight < minWeight {
			minWeight = weight
		}
		if weight > maxWeight {
			maxWeight = weight
		}
	}

	// Calculate stability score (lower coefficient of variation = higher stability)
	stabilityScore := 1.0 - math.Min(coefficientOfVariation, 1.0)

	return &WeightStatistics{
		MethodName:             history[0].MethodName,
		CurrentWeight:          weights[len(weights)-1],
		MeanWeight:             mean,
		MinWeight:              minWeight,
		MaxWeight:              maxWeight,
		WeightVariance:         variance,
		WeightStdDev:           stdDev,
		CoefficientOfVariation: coefficientOfVariation,
		StabilityScore:         stabilityScore,
		ChangeCount:            len(history),
		LastChangeTime:         history[len(history)-1].Timestamp,
	}
}

// determineDistributionHealth determines overall distribution health
func (ewm *EnsembleWeightMonitor) determineDistributionHealth() string {
	// Check for critical alerts
	for _, alert := range ewm.weightAlerts {
		if alert.Severity == "critical" && !alert.Resolved {
			return "critical"
		}
	}

	// Check for warning alerts
	for _, alert := range ewm.weightAlerts {
		if alert.Severity == "warning" && !alert.Resolved {
			return "warning"
		}
	}

	// Check for weight imbalance
	if len(ewm.currentWeights) > 0 {
		totalWeight := 0.0
		maxWeight := 0.0

		for _, weight := range ewm.currentWeights {
			totalWeight += weight
			if weight > maxWeight {
				maxWeight = weight
			}
		}

		if totalWeight > 0 && maxWeight/totalWeight > ewm.config.ImbalanceThreshold {
			return "warning"
		}
	}

	return "healthy"
}

// AcknowledgeWeightAlert acknowledges a weight distribution alert
func (ewm *EnsembleWeightMonitor) AcknowledgeWeightAlert(alertID string) error {
	ewm.mu.Lock()
	defer ewm.mu.Unlock()

	for _, alert := range ewm.weightAlerts {
		if alert.ID == alertID {
			alert.Acknowledged = true
			alert.AcknowledgedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("weight alert not found: %s", alertID)
}

// ResolveWeightAlert resolves a weight distribution alert
func (ewm *EnsembleWeightMonitor) ResolveWeightAlert(alertID string) error {
	ewm.mu.Lock()
	defer ewm.mu.Unlock()

	for _, alert := range ewm.weightAlerts {
		if alert.ID == alertID {
			alert.Resolved = true
			alert.ResolvedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("weight alert not found: %s", alertID)
}
