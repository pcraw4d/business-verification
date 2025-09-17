package classification_monitoring

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SecurityProcessingTimeMonitor provides specialized monitoring for security validation processing time
type SecurityProcessingTimeMonitor struct {
	config *SecurityProcessingTimeConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// Processing time tracking
	processingTimes    map[string][]*SecurityProcessingTimeData
	performanceMetrics map[string]*SecurityPerformanceMetrics
	processingAlerts   []*SecurityProcessingAlert

	// Analysis components
	performanceAnalyzer *SecurityPerformanceAnalyzer
	thresholdMonitor    *SecurityThresholdMonitor
	trendAnalyzer       *SecurityTrendAnalyzer
}

// SecurityProcessingTimeConfig holds configuration for security processing time monitoring
type SecurityProcessingTimeConfig struct {
	// Data collection
	DataCollectionEnabled    bool          `json:"data_collection_enabled"`
	MaxDataPointsPerMethod   int           `json:"max_data_points_per_method"`
	CollectionInterval       time.Duration `json:"collection_interval"`

	// Performance thresholds
	SlowProcessingThreshold    time.Duration `json:"slow_processing_threshold"`
	CriticalProcessingThreshold time.Duration `json:"critical_processing_threshold"`
	AverageProcessingThreshold  time.Duration `json:"average_processing_threshold"`

	// Trend analysis
	TrendAnalysisEnabled     bool          `json:"trend_analysis_enabled"`
	TrendWindowSize          int           `json:"trend_window_size"`
	MinSamplesForTrend       int           `json:"min_samples_for_trend"`
	PerformanceDegradationThreshold float64 `json:"performance_degradation_threshold"`

	// Alerting
	AlertingEnabled      bool          `json:"alerting_enabled"`
	AlertCooldownPeriod  time.Duration `json:"alert_cooldown_period"`
	MaxAlertsPerMethod   int           `json:"max_alerts_per_method"`
	AlertRetentionPeriod time.Duration `json:"alert_retention_period"`

	// Monitoring intervals
	MonitoringInterval    time.Duration `json:"monitoring_interval"`
	AnalysisInterval      time.Duration `json:"analysis_interval"`
	CleanupInterval       time.Duration `json:"cleanup_interval"`
}

// SecurityProcessingTimeData represents a security processing time measurement
type SecurityProcessingTimeData struct {
	Timestamp       time.Time `json:"timestamp"`
	MethodName      string    `json:"method_name"`
	ValidationType  string    `json:"validation_type"`
	ProcessingTime  time.Duration `json:"processing_time"`
	Success         bool      `json:"success"`
	ErrorType       string    `json:"error_type,omitempty"`
	ResourceUsage   float64   `json:"resource_usage,omitempty"`
	ConcurrentLoad  int       `json:"concurrent_load"`
	SampleSize      int       `json:"sample_size"`
}

// SecurityPerformanceMetrics represents performance metrics for a security method
type SecurityPerformanceMetrics struct {
	MethodName              string        `json:"method_name"`
	TotalExecutions         int64         `json:"total_executions"`
	SuccessfulExecutions    int64         `json:"successful_executions"`
	FailedExecutions        int64         `json:"failed_executions"`
	AverageProcessingTime   time.Duration `json:"average_processing_time"`
	MinProcessingTime       time.Duration `json:"min_processing_time"`
	MaxProcessingTime       time.Duration `json:"max_processing_time"`
	P95ProcessingTime       time.Duration `json:"p95_processing_time"`
	P99ProcessingTime       time.Duration `json:"p99_processing_time"`
	SuccessRate             float64       `json:"success_rate"`
	ErrorRate               float64       `json:"error_rate"`
	PerformanceTrend        string        `json:"performance_trend"` // "improving", "stable", "degrading"
	StabilityScore          float64       `json:"stability_score"`
	LastUpdated             time.Time     `json:"last_updated"`
}

// SecurityProcessingAlert represents an alert about security processing time issues
type SecurityProcessingAlert struct {
	ID              string    `json:"id"`
	AlertType       string    `json:"alert_type"` // "slow_processing", "critical_processing", "performance_degradation", "high_error_rate"
	Severity        string    `json:"severity"`   // "warning", "critical"
	MethodName      string    `json:"method_name"`
	Message         string    `json:"message"`
	Details         string    `json:"details"`
	Metrics         map[string]interface{} `json:"metrics"`
	Timestamp       time.Time `json:"timestamp"`
	Acknowledged    bool      `json:"acknowledged"`
	AcknowledgedAt  time.Time `json:"acknowledged_at,omitempty"`
	Resolved        bool      `json:"resolved"`
	ResolvedAt      time.Time `json:"resolved_at,omitempty"`
}

// SecurityPerformanceAnalyzer analyzes security processing performance
type SecurityPerformanceAnalyzer struct {
	config *SecurityProcessingTimeConfig
	logger *zap.Logger
}

// SecurityThresholdMonitor monitors security processing thresholds
type SecurityThresholdMonitor struct {
	config *SecurityProcessingTimeConfig
	logger *zap.Logger
}

// SecurityTrendAnalyzer analyzes security processing trends
type SecurityTrendAnalyzer struct {
	config *SecurityProcessingTimeConfig
	logger *zap.Logger
}

// NewSecurityProcessingTimeMonitor creates a new security processing time monitor
func NewSecurityProcessingTimeMonitor(config *SecurityProcessingTimeConfig, logger *zap.Logger) *SecurityProcessingTimeMonitor {
	if config == nil {
		config = DefaultSecurityProcessingTimeConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &SecurityProcessingTimeMonitor{
		config:              config,
		logger:              logger,
		processingTimes:     make(map[string][]*SecurityProcessingTimeData),
		performanceMetrics:  make(map[string]*SecurityPerformanceMetrics),
		processingAlerts:    make([]*SecurityProcessingAlert, 0),
		performanceAnalyzer: &SecurityPerformanceAnalyzer{
			config: config,
			logger: logger,
		},
		thresholdMonitor: &SecurityThresholdMonitor{
			config: config,
			logger: logger,
		},
		trendAnalyzer: &SecurityTrendAnalyzer{
			config: config,
			logger: logger,
		},
	}
}

// DefaultSecurityProcessingTimeConfig returns default configuration
func DefaultSecurityProcessingTimeConfig() *SecurityProcessingTimeConfig {
	return &SecurityProcessingTimeConfig{
		DataCollectionEnabled:     true,
		MaxDataPointsPerMethod:   1000,
		CollectionInterval:       1 * time.Minute,
		SlowProcessingThreshold:   500 * time.Millisecond,
		CriticalProcessingThreshold: 2 * time.Second,
		AverageProcessingThreshold: 200 * time.Millisecond,
		TrendAnalysisEnabled:      true,
		TrendWindowSize:           100,
		MinSamplesForTrend:        20,
		PerformanceDegradationThreshold: 0.2, // 20% degradation threshold
		AlertingEnabled:           true,
		AlertCooldownPeriod:       15 * time.Minute,
		MaxAlertsPerMethod:        20,
		AlertRetentionPeriod:      24 * time.Hour,
		MonitoringInterval:        1 * time.Minute,
		AnalysisInterval:          5 * time.Minute,
		CleanupInterval:           1 * time.Hour,
	}
}

// TrackSecurityProcessingTime tracks security validation processing time
func (sptm *SecurityProcessingTimeMonitor) TrackSecurityProcessingTime(
	ctx context.Context,
	methodName string,
	validationType string,
	processingTime time.Duration,
	success bool,
	errorType string,
	resourceUsage float64,
	concurrentLoad int,
) error {
	sptm.mu.Lock()
	defer sptm.mu.Unlock()

	if !sptm.config.DataCollectionEnabled {
		return nil
	}

	// Create processing time data point
	dataPoint := &SecurityProcessingTimeData{
		Timestamp:      time.Now(),
		MethodName:     methodName,
		ValidationType: validationType,
		ProcessingTime: processingTime,
		Success:        success,
		ErrorType:      errorType,
		ResourceUsage:  resourceUsage,
		ConcurrentLoad: concurrentLoad,
		SampleSize:     1,
	}

	// Add to processing times
	sptm.processingTimes[methodName] = append(sptm.processingTimes[methodName], dataPoint)

	// Maintain data size
	if len(sptm.processingTimes[methodName]) > sptm.config.MaxDataPointsPerMethod {
		sptm.processingTimes[methodName] = sptm.processingTimes[methodName][1:]
	}

	// Update performance metrics
	sptm.updatePerformanceMetrics(methodName, dataPoint)

	// Check thresholds
	sptm.checkProcessingThresholds(methodName, dataPoint)

	// Analyze trends
	if sptm.config.TrendAnalysisEnabled {
		sptm.analyzeProcessingTrends(methodName)
	}

	sptm.logger.Debug("Security processing time tracked",
		zap.String("method_name", methodName),
		zap.String("validation_type", validationType),
		zap.Duration("processing_time", processingTime),
		zap.Bool("success", success))

	return nil
}

// updatePerformanceMetrics updates performance metrics for a method
func (sptm *SecurityProcessingTimeMonitor) updatePerformanceMetrics(methodName string, dataPoint *SecurityProcessingTimeData) {
	metrics, exists := sptm.performanceMetrics[methodName]
	if !exists {
		metrics = &SecurityPerformanceMetrics{
			MethodName:            methodName,
			TotalExecutions:       0,
			SuccessfulExecutions:  0,
			FailedExecutions:      0,
			AverageProcessingTime: 0,
			MinProcessingTime:     dataPoint.ProcessingTime,
			MaxProcessingTime:     dataPoint.ProcessingTime,
			P95ProcessingTime:     0,
			P99ProcessingTime:     0,
			SuccessRate:           0.0,
			ErrorRate:             0.0,
			PerformanceTrend:      "stable",
			StabilityScore:        0.0,
			LastUpdated:           time.Now(),
		}
		sptm.performanceMetrics[methodName] = metrics
	}

	// Update execution counts
	metrics.TotalExecutions++
	if dataPoint.Success {
		metrics.SuccessfulExecutions++
	} else {
		metrics.FailedExecutions++
	}

	// Update processing time metrics
	metrics.AverageProcessingTime = (metrics.AverageProcessingTime*time.Duration(metrics.TotalExecutions-1) + dataPoint.ProcessingTime) / time.Duration(metrics.TotalExecutions)

	if dataPoint.ProcessingTime < metrics.MinProcessingTime {
		metrics.MinProcessingTime = dataPoint.ProcessingTime
	}
	if dataPoint.ProcessingTime > metrics.MaxProcessingTime {
		metrics.MaxProcessingTime = dataPoint.ProcessingTime
	}

	// Update success/error rates
	metrics.SuccessRate = float64(metrics.SuccessfulExecutions) / float64(metrics.TotalExecutions)
	metrics.ErrorRate = float64(metrics.FailedExecutions) / float64(metrics.TotalExecutions)

	// Calculate percentiles
	sptm.calculatePercentiles(metrics)

	// Update stability score
	metrics.StabilityScore = sptm.calculateStabilityScore(methodName)

	metrics.LastUpdated = time.Now()
}

// calculatePercentiles calculates P95 and P99 processing times
func (sptm *SecurityProcessingTimeMonitor) calculatePercentiles(metrics *SecurityPerformanceMetrics) {
	processingTimes := sptm.processingTimes[metrics.MethodName]
	if len(processingTimes) < 20 {
		return // Need sufficient data for percentiles
	}

	// Sort processing times
	times := make([]time.Duration, len(processingTimes))
	for i, data := range processingTimes {
		times[i] = data.ProcessingTime
	}

	// Simple sorting (in production, use a more efficient algorithm)
	for i := 0; i < len(times)-1; i++ {
		for j := i + 1; j < len(times); j++ {
			if times[i] > times[j] {
				times[i], times[j] = times[j], times[i]
			}
		}
	}

	// Calculate percentiles
	p95Index := int(float64(len(times)) * 0.95)
	p99Index := int(float64(len(times)) * 0.99)

	if p95Index < len(times) {
		metrics.P95ProcessingTime = times[p95Index]
	}
	if p99Index < len(times) {
		metrics.P99ProcessingTime = times[p99Index]
	}
}

// calculateStabilityScore calculates stability score for a method
func (sptm *SecurityProcessingTimeMonitor) calculateStabilityScore(methodName string) float64 {
	processingTimes := sptm.processingTimes[methodName]
	if len(processingTimes) < 10 {
		return 0.0
	}

	// Calculate coefficient of variation
	times := make([]float64, len(processingTimes))
	for i, data := range processingTimes {
		times[i] = float64(data.ProcessingTime.Nanoseconds())
	}

	mean := sptm.calculateMean(times)
	if mean == 0 {
		return 0.0
	}

	variance := sptm.calculateVariance(times, mean)
	stdDev := math.Sqrt(variance)
	coefficientOfVariation := stdDev / mean

	// Stability score (lower coefficient of variation = higher stability)
	return 1.0 - math.Min(coefficientOfVariation, 1.0)
}

// calculateMean calculates mean of a slice of float64 values
func (sptm *SecurityProcessingTimeMonitor) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}

	return sum / float64(len(values))
}

// calculateVariance calculates variance of a slice of float64 values
func (sptm *SecurityProcessingTimeMonitor) calculateVariance(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	sumSquaredDiffs := 0.0
	for _, value := range values {
		diff := value - mean
		sumSquaredDiffs += diff * diff
	}

	return sumSquaredDiffs / float64(len(values)-1)
}

// checkProcessingThresholds checks processing time thresholds
func (sptm *SecurityProcessingTimeMonitor) checkProcessingThresholds(methodName string, dataPoint *SecurityProcessingTimeData) {
	// Check for critical processing time
	if dataPoint.ProcessingTime > sptm.config.CriticalProcessingThreshold {
		sptm.createProcessingAlert("critical_processing", "critical", methodName,
			fmt.Sprintf("Critical processing time exceeded: %v (threshold: %v)",
				dataPoint.ProcessingTime, sptm.config.CriticalProcessingThreshold),
			fmt.Sprintf("Method: %s, Validation: %s, Success: %v",
				methodName, dataPoint.ValidationType, dataPoint.Success),
			map[string]interface{}{
				"processing_time": dataPoint.ProcessingTime.String(),
				"threshold":       sptm.config.CriticalProcessingThreshold.String(),
				"validation_type": dataPoint.ValidationType,
				"success":         dataPoint.Success,
			})
	}

	// Check for slow processing time
	if dataPoint.ProcessingTime > sptm.config.SlowProcessingThreshold {
		sptm.createProcessingAlert("slow_processing", "warning", methodName,
			fmt.Sprintf("Slow processing time detected: %v (threshold: %v)",
				dataPoint.ProcessingTime, sptm.config.SlowProcessingThreshold),
			fmt.Sprintf("Method: %s, Validation: %s, Success: %v",
				methodName, dataPoint.ValidationType, dataPoint.Success),
			map[string]interface{}{
				"processing_time": dataPoint.ProcessingTime.String(),
				"threshold":       sptm.config.SlowProcessingThreshold.String(),
				"validation_type": dataPoint.ValidationType,
				"success":         dataPoint.Success,
			})
	}

	// Check for high error rate
	metrics := sptm.performanceMetrics[methodName]
	if metrics != nil && metrics.TotalExecutions >= 50 { // Need sufficient data
		if metrics.ErrorRate > 0.1 { // 10% error rate threshold
			sptm.createProcessingAlert("high_error_rate", "warning", methodName,
				fmt.Sprintf("High error rate detected: %.1f%% (threshold: 10%%)",
					metrics.ErrorRate*100),
				fmt.Sprintf("Method: %s, Total executions: %d, Failed: %d",
					methodName, metrics.TotalExecutions, metrics.FailedExecutions),
				map[string]interface{}{
					"error_rate":        metrics.ErrorRate,
					"total_executions":  metrics.TotalExecutions,
					"failed_executions": metrics.FailedExecutions,
				})
		}
	}
}

// analyzeProcessingTrends analyzes processing time trends
func (sptm *SecurityProcessingTimeMonitor) analyzeProcessingTrends(methodName string) {
	processingTimes := sptm.processingTimes[methodName]
	if len(processingTimes) < sptm.config.MinSamplesForTrend {
		return
	}

	// Get recent data for trend analysis
	recentData := processingTimes
	if len(processingTimes) > sptm.config.TrendWindowSize {
		recentData = processingTimes[len(processingTimes)-sptm.config.TrendWindowSize:]
	}

	// Calculate trend
	trend := sptm.calculateProcessingTrend(recentData)
	metrics := sptm.performanceMetrics[methodName]
	if metrics != nil {
		metrics.PerformanceTrend = trend
	}

	// Check for performance degradation
	if trend == "degrading" {
		// Calculate degradation percentage
		degradation := sptm.calculatePerformanceDegradation(recentData)
		if degradation > sptm.config.PerformanceDegradationThreshold {
			sptm.createProcessingAlert("performance_degradation", "warning", methodName,
				fmt.Sprintf("Performance degradation detected: %.1f%% (threshold: %.1f%%)",
					degradation*100, sptm.config.PerformanceDegradationThreshold*100),
				fmt.Sprintf("Method: %s, Recent trend shows declining performance",
					methodName),
				map[string]interface{}{
					"degradation_percentage": degradation,
					"threshold":              sptm.config.PerformanceDegradationThreshold,
					"trend":                  trend,
				})
		}
	}
}

// calculateProcessingTrend calculates processing time trend
func (sptm *SecurityProcessingTimeMonitor) calculateProcessingTrend(data []*SecurityProcessingTimeData) string {
	if len(data) < 10 {
		return "stable"
	}

	// Split data into two halves
	midPoint := len(data) / 2
	firstHalf := data[:midPoint]
	secondHalf := data[midPoint:]

	// Calculate average processing times
	firstAvg := sptm.calculateAverageProcessingTime(firstHalf)
	secondAvg := sptm.calculateAverageProcessingTime(secondHalf)

	// Calculate trend
	changePercent := float64(secondAvg-firstAvg) / float64(firstAvg)

	if changePercent > 0.1 { // 10% increase
		return "degrading"
	} else if changePercent < -0.1 { // 10% decrease
		return "improving"
	} else {
		return "stable"
	}
}

// calculateAverageProcessingTime calculates average processing time
func (sptm *SecurityProcessingTimeMonitor) calculateAverageProcessingTime(data []*SecurityProcessingTimeData) time.Duration {
	if len(data) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, point := range data {
		total += point.ProcessingTime
	}

	return total / time.Duration(len(data))
}

// calculatePerformanceDegradation calculates performance degradation percentage
func (sptm *SecurityProcessingTimeMonitor) calculatePerformanceDegradation(data []*SecurityProcessingTimeData) float64 {
	if len(data) < 20 {
		return 0.0
	}

	// Split into quarters
	quarterSize := len(data) / 4
	firstQuarter := data[:quarterSize]
	lastQuarter := data[len(data)-quarterSize:]

	firstAvg := sptm.calculateAverageProcessingTime(firstQuarter)
	lastAvg := sptm.calculateAverageProcessingTime(lastQuarter)

	if firstAvg == 0 {
		return 0.0
	}

	return (lastAvg - firstAvg).Seconds() / firstAvg.Seconds()
}

// createProcessingAlert creates a security processing alert
func (sptm *SecurityProcessingTimeMonitor) createProcessingAlert(alertType, severity, methodName, message, details string, metrics map[string]interface{}) {
	if !sptm.config.AlertingEnabled {
		return
	}

	// Check cooldown period
	lastAlert := sptm.getLastAlertForMethod(methodName, alertType)
	if lastAlert != nil && time.Since(lastAlert.Timestamp) < sptm.config.AlertCooldownPeriod {
		return
	}

	// Check alert limits
	alertCount := sptm.getAlertCountForMethod(methodName)
	if alertCount >= sptm.config.MaxAlertsPerMethod {
		return
	}

	alert := &SecurityProcessingAlert{
		ID:           fmt.Sprintf("security_processing_alert_%s_%s_%d", alertType, methodName, time.Now().Unix()),
		AlertType:    alertType,
		Severity:     severity,
		MethodName:   methodName,
		Message:      message,
		Details:      details,
		Metrics:      metrics,
		Timestamp:    time.Now(),
		Acknowledged: false,
		Resolved:     false,
	}

	sptm.processingAlerts = append(sptm.processingAlerts, alert)

	// Clean up old alerts
	sptm.cleanupOldAlerts()

	sptm.logger.Warn("Security processing alert created",
		zap.String("alert_type", alertType),
		zap.String("severity", severity),
		zap.String("method_name", methodName),
		zap.String("message", message))
}

// getLastAlertForMethod gets the last alert for a specific method and type
func (sptm *SecurityProcessingTimeMonitor) getLastAlertForMethod(methodName, alertType string) *SecurityProcessingAlert {
	for i := len(sptm.processingAlerts) - 1; i >= 0; i-- {
		alert := sptm.processingAlerts[i]
		if alert.MethodName == methodName && alert.AlertType == alertType {
			return alert
		}
	}
	return nil
}

// getAlertCountForMethod gets the alert count for a specific method
func (sptm *SecurityProcessingTimeMonitor) getAlertCountForMethod(methodName string) int {
	count := 0
	for _, alert := range sptm.processingAlerts {
		if alert.MethodName == methodName {
			count++
		}
	}
	return count
}

// cleanupOldAlerts removes old alerts beyond retention period
func (sptm *SecurityProcessingTimeMonitor) cleanupOldAlerts() {
	cutoff := time.Now().Add(-sptm.config.AlertRetentionPeriod)
	var validAlerts []*SecurityProcessingAlert

	for _, alert := range sptm.processingAlerts {
		if alert.Timestamp.After(cutoff) {
			validAlerts = append(validAlerts, alert)
		}
	}

	sptm.processingAlerts = validAlerts
}

// GetSecurityProcessingMetrics returns comprehensive security processing metrics
func (sptm *SecurityProcessingTimeMonitor) GetSecurityProcessingMetrics() *SecurityProcessingMetrics {
	sptm.mu.RLock()
	defer sptm.mu.RUnlock()

	metrics := &SecurityProcessingMetrics{
		Timestamp:           time.Now(),
		TotalMethods:        len(sptm.performanceMetrics),
		PerformanceMetrics:  make(map[string]*SecurityPerformanceMetrics),
		OverallHealth:       "healthy",
		AlertsCount:         len(sptm.processingAlerts),
		CriticalAlertsCount: 0,
		WarningAlertsCount:  0,
	}

	// Copy performance metrics
	for method, perfMetrics := range sptm.performanceMetrics {
		metrics.PerformanceMetrics[method] = perfMetrics
	}

	// Count alerts by severity
	for _, alert := range sptm.processingAlerts {
		if !alert.Resolved {
			if alert.Severity == "critical" {
				metrics.CriticalAlertsCount++
			} else if alert.Severity == "warning" {
				metrics.WarningAlertsCount++
			}
		}
	}

	// Determine overall health
	metrics.OverallHealth = sptm.determineOverallHealth()

	return metrics
}

// SecurityProcessingMetrics represents comprehensive security processing metrics
type SecurityProcessingMetrics struct {
	Timestamp           time.Time                           `json:"timestamp"`
	TotalMethods        int                                 `json:"total_methods"`
	PerformanceMetrics  map[string]*SecurityPerformanceMetrics `json:"performance_metrics"`
	OverallHealth       string                              `json:"overall_health"`
	AlertsCount         int                                 `json:"alerts_count"`
	CriticalAlertsCount int                                 `json:"critical_alerts_count"`
	WarningAlertsCount  int                                 `json:"warning_alerts_count"`
}

// determineOverallHealth determines overall health status
func (sptm *SecurityProcessingTimeMonitor) determineOverallHealth() string {
	// Check for critical alerts
	for _, alert := range sptm.processingAlerts {
		if alert.Severity == "critical" && !alert.Resolved {
			return "critical"
		}
	}

	// Check for warning alerts
	for _, alert := range sptm.processingAlerts {
		if alert.Severity == "warning" && !alert.Resolved {
			return "warning"
		}
	}

	// Check for performance issues
	for _, metrics := range sptm.performanceMetrics {
		if metrics.PerformanceTrend == "degrading" {
			return "warning"
		}
		if metrics.ErrorRate > 0.2 { // 20% error rate
			return "warning"
		}
	}

	return "healthy"
}

// GetSecurityProcessingAlerts returns all security processing alerts
func (sptm *SecurityProcessingTimeMonitor) GetSecurityProcessingAlerts() []*SecurityProcessingAlert {
	sptm.mu.RLock()
	defer sptm.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make([]*SecurityProcessingAlert, len(sptm.processingAlerts))
	copy(result, sptm.processingAlerts)

	return result
}

// AcknowledgeSecurityProcessingAlert acknowledges a security processing alert
func (sptm *SecurityProcessingTimeMonitor) AcknowledgeSecurityProcessingAlert(alertID string) error {
	sptm.mu.Lock()
	defer sptm.mu.Unlock()

	for _, alert := range sptm.processingAlerts {
		if alert.ID == alertID {
			alert.Acknowledged = true
			alert.AcknowledgedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("security processing alert not found: %s", alertID)
}

// ResolveSecurityProcessingAlert resolves a security processing alert
func (sptm *SecurityProcessingTimeMonitor) ResolveSecurityProcessingAlert(alertID string) error {
	sptm.mu.Lock()
	defer sptm.mu.Unlock()

	for _, alert := range sptm.processingAlerts {
		if alert.ID == alertID {
			alert.Resolved = true
			alert.ResolvedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("security processing alert not found: %s", alertID)
}

// GetMethodPerformanceReport returns a detailed performance report for a method
func (sptm *SecurityProcessingTimeMonitor) GetMethodPerformanceReport(methodName string) (*SecurityMethodPerformanceReport, error) {
	sptm.mu.RLock()
	defer sptm.mu.RUnlock()

	metrics, exists := sptm.performanceMetrics[methodName]
	if !exists {
		return nil, fmt.Errorf("performance metrics not found for method: %s", methodName)
	}

	processingTimes := sptm.processingTimes[methodName]
	if len(processingTimes) == 0 {
		return nil, fmt.Errorf("no processing time data found for method: %s", methodName)
	}

	report := &SecurityMethodPerformanceReport{
		MethodName:           methodName,
		Timestamp:            time.Now(),
		PerformanceMetrics:   metrics,
		RecentProcessingTimes: make([]time.Duration, 0),
		Recommendations:      make([]string, 0),
	}

	// Get recent processing times (last 50)
	recentCount := 50
	if len(processingTimes) < recentCount {
		recentCount = len(processingTimes)
	}

	for i := len(processingTimes) - recentCount; i < len(processingTimes); i++ {
		report.RecentProcessingTimes = append(report.RecentProcessingTimes, processingTimes[i].ProcessingTime)
	}

	// Generate recommendations
	if metrics.PerformanceTrend == "degrading" {
		report.Recommendations = append(report.Recommendations,
			"Performance is degrading. Consider optimizing the method or investigating resource constraints.")
	}

	if metrics.ErrorRate > 0.1 {
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("High error rate detected (%.1f%%). Investigate error causes and improve error handling.",
				metrics.ErrorRate*100))
	}

	if metrics.AverageProcessingTime > sptm.config.SlowProcessingThreshold {
		report.Recommendations = append(report.Recommendations,
			fmt.Sprintf("Average processing time (%.2fms) exceeds threshold (%.2fms). Consider performance optimization.",
				float64(metrics.AverageProcessingTime.Nanoseconds())/1e6,
				float64(sptm.config.SlowProcessingThreshold.Nanoseconds())/1e6))
	}

	if metrics.StabilityScore < 0.7 {
		report.Recommendations = append(report.Recommendations,
			"Low stability score detected. Processing times are highly variable. Consider improving consistency.")
	}

	return report, nil
}

// SecurityMethodPerformanceReport represents a detailed performance report for a security method
type SecurityMethodPerformanceReport struct {
	MethodName            string          `json:"method_name"`
	Timestamp             time.Time       `json:"timestamp"`
	PerformanceMetrics    *SecurityPerformanceMetrics `json:"performance_metrics"`
	RecentProcessingTimes []time.Duration `json:"recent_processing_times"`
	Recommendations       []string        `json:"recommendations"`
}
