package website_verification

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// SuccessMonitor implements verification success monitoring
type SuccessMonitor struct {
	// Configuration
	config *SuccessMonitorConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Success rate tracking
	successTracker *SuccessRateTracker
	trackerMux     sync.RWMutex

	// Failure analysis
	failureAnalyzer *FailureAnalyzer
	analyzerMux     sync.RWMutex

	// Success rate alerts
	alertManager *SuccessAlertManager
	alertMux     sync.RWMutex

	// Performance metrics
	performanceTracker *VerificationPerformanceTracker
	performanceMux     sync.RWMutex

	// Historical tracking
	historicalTracker *HistoricalSuccessTracker
	historicalMux     sync.RWMutex
}

// SuccessMonitorConfig configuration for success monitoring
type SuccessMonitorConfig struct {
	// Success rate tracking settings
	SuccessRateTrackingEnabled bool
	TrackingWindow             time.Duration
	MinSampleSize              int
	SuccessThreshold           float64

	// Failure analysis settings
	FailureAnalysisEnabled  bool
	AnalysisWindow          time.Duration
	MaxFailurePatterns      int
	FailurePatternThreshold float64

	// Alert settings
	AlertingEnabled bool
	AlertThreshold  float64
	AlertCooldown   time.Duration
	AlertChannels   []string

	// Performance tracking settings
	PerformanceTrackingEnabled bool
	PerformanceWindow          time.Duration
	PerformanceMetrics         []string

	// Historical tracking settings
	HistoricalTrackingEnabled bool
	HistoricalRetentionPeriod time.Duration
	HistoricalGranularity     time.Duration
	HistoricalMaxDataPoints   int
}

// SuccessRateTracker tracks verification success rates
type SuccessRateTracker struct {
	enabled          bool
	window           time.Duration
	minSampleSize    int
	successThreshold float64
	successCount     int64
	failureCount     int64
	totalCount       int64
	windowStart      time.Time
	windowMux        sync.RWMutex
	successHistory   []SuccessEvent
	historyMux       sync.RWMutex
}

// SuccessEvent represents a verification success event
type SuccessEvent struct {
	Timestamp  time.Time
	Domain     string
	Success    bool
	Method     string
	Duration   time.Duration
	Confidence float64
	Error      string
}

// FailureAnalyzer analyzes verification failures
type FailureAnalyzer struct {
	enabled          bool
	window           time.Duration
	maxPatterns      int
	patternThreshold float64
	failurePatterns  map[string]*FailurePattern
	patternMux       sync.RWMutex
	failureHistory   []FailureEvent
	historyMux       sync.RWMutex
}

// FailurePattern represents a failure pattern
type FailurePattern struct {
	Pattern         string
	Count           int
	LastSeen        time.Time
	FirstSeen       time.Time
	FailureRate     float64
	CommonErrors    []string
	AffectedDomains []string
}

// FailureEvent represents a verification failure event
type FailureEvent struct {
	Timestamp  time.Time
	Domain     string
	Method     string
	Error      string
	ErrorType  string
	Duration   time.Duration
	RetryCount int
}

// SuccessAlertManager manages success rate alerts
type SuccessAlertManager struct {
	enabled      bool
	threshold    float64
	cooldown     time.Duration
	channels     []string
	lastAlert    time.Time
	alertMux     sync.RWMutex
	alertHistory []AlertEvent
	historyMux   sync.RWMutex
}

// AlertEvent represents an alert event
type AlertEvent struct {
	Timestamp   time.Time
	Type        string
	Message     string
	Severity    string
	SuccessRate float64
	Threshold   float64
	Channels    []string
}

// VerificationPerformanceTracker tracks verification performance metrics
type VerificationPerformanceTracker struct {
	enabled         bool
	window          time.Duration
	metrics         []string
	performanceData map[string]*PerformanceMetric
	performanceMux  sync.RWMutex
	metricHistory   []PerformanceEvent
	historyMux      sync.RWMutex
}

// PerformanceMetric represents a performance metric
type PerformanceMetric struct {
	Name         string
	Value        float64
	Unit         string
	LastUpdated  time.Time
	MinValue     float64
	MaxValue     float64
	AverageValue float64
	SampleCount  int
}

// PerformanceEvent represents a performance event
type PerformanceEvent struct {
	Timestamp  time.Time
	MetricName string
	Value      float64
	Unit       string
	Domain     string
	Method     string
}

// HistoricalSuccessTracker tracks historical success rates
type HistoricalSuccessTracker struct {
	enabled          bool
	retentionPeriod  time.Duration
	granularity      time.Duration
	maxDataPoints    int
	historicalData   map[string]*HistoricalDataPoint
	historicalMux    sync.RWMutex
	dataPointHistory []HistoricalDataPoint
	historyMux       sync.RWMutex
}

// HistoricalDataPoint represents a historical data point
type HistoricalDataPoint struct {
	Timestamp       time.Time
	SuccessRate     float64
	TotalCount      int
	SuccessCount    int
	FailureCount    int
	AverageDuration time.Duration
	Methods         map[string]MethodStats
}

// MethodStats represents statistics for a verification method
type MethodStats struct {
	SuccessCount    int
	FailureCount    int
	AverageDuration time.Duration
	SuccessRate     float64
}

// NewSuccessMonitor creates a new success monitor
func NewSuccessMonitor(config *SuccessMonitorConfig, logger *observability.Logger, tracer trace.Tracer) *SuccessMonitor {
	if config == nil {
		config = &SuccessMonitorConfig{
			SuccessRateTrackingEnabled: true,
			TrackingWindow:             1 * time.Hour,
			MinSampleSize:              10,
			SuccessThreshold:           0.9,
			FailureAnalysisEnabled:     true,
			AnalysisWindow:             24 * time.Hour,
			MaxFailurePatterns:         50,
			FailurePatternThreshold:    0.1,
			AlertingEnabled:            true,
			AlertThreshold:             0.8,
			AlertCooldown:              30 * time.Minute,
			AlertChannels:              []string{"email", "slack", "pagerduty"},
			PerformanceTrackingEnabled: true,
			PerformanceWindow:          1 * time.Hour,
			PerformanceMetrics: []string{
				"verification_duration",
				"success_rate",
				"failure_rate",
				"retry_count",
				"cache_hit_rate",
			},
			HistoricalTrackingEnabled: true,
			HistoricalRetentionPeriod: 30 * 24 * time.Hour, // 30 days
			HistoricalGranularity:     1 * time.Hour,
			HistoricalMaxDataPoints:   720, // 30 days * 24 hours
		}
	}

	sm := &SuccessMonitor{
		config: config,
		logger: logger,
		tracer: tracer,
	}

	// Initialize components
	sm.successTracker = &SuccessRateTracker{
		enabled:          config.SuccessRateTrackingEnabled,
		window:           config.TrackingWindow,
		minSampleSize:    config.MinSampleSize,
		successThreshold: config.SuccessThreshold,
		windowStart:      time.Now(),
		successHistory:   make([]SuccessEvent, 0),
	}

	sm.failureAnalyzer = &FailureAnalyzer{
		enabled:          config.FailureAnalysisEnabled,
		window:           config.AnalysisWindow,
		maxPatterns:      config.MaxFailurePatterns,
		patternThreshold: config.FailurePatternThreshold,
		failurePatterns:  make(map[string]*FailurePattern),
		failureHistory:   make([]FailureEvent, 0),
	}

	sm.alertManager = &SuccessAlertManager{
		enabled:      config.AlertingEnabled,
		threshold:    config.AlertThreshold,
		cooldown:     config.AlertCooldown,
		channels:     config.AlertChannels,
		alertHistory: make([]AlertEvent, 0),
	}

	sm.performanceTracker = &VerificationPerformanceTracker{
		enabled:         config.PerformanceTrackingEnabled,
		window:          config.PerformanceWindow,
		metrics:         config.PerformanceMetrics,
		performanceData: make(map[string]*PerformanceMetric),
		metricHistory:   make([]PerformanceEvent, 0),
	}

	sm.historicalTracker = &HistoricalSuccessTracker{
		enabled:          config.HistoricalTrackingEnabled,
		retentionPeriod:  config.HistoricalRetentionPeriod,
		granularity:      config.HistoricalGranularity,
		maxDataPoints:    config.HistoricalMaxDataPoints,
		historicalData:   make(map[string]*HistoricalDataPoint),
		dataPointHistory: make([]HistoricalDataPoint, 0),
	}

	// Start background workers
	sm.startBackgroundWorkers()

	return sm
}

// RecordVerificationResult records a verification result
func (sm *SuccessMonitor) RecordVerificationResult(ctx context.Context, domain, method string, success bool, duration time.Duration, confidence float64, err error) {
	ctx, span := sm.tracer.Start(ctx, "SuccessMonitor.RecordVerificationResult")
	defer span.End()

	span.SetAttributes(
		attribute.String("domain", domain),
		attribute.String("method", method),
		attribute.Bool("success", success),
		attribute.Float64("confidence", confidence),
	)

	timestamp := time.Now()
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}

	// Record success event
	event := SuccessEvent{
		Timestamp:  timestamp,
		Domain:     domain,
		Success:    success,
		Method:     method,
		Duration:   duration,
		Confidence: confidence,
		Error:      errorMsg,
	}

	// Update success tracker
	if sm.config.SuccessRateTrackingEnabled {
		sm.successTracker.RecordEvent(event)
	}

	// Update failure analyzer if failed
	if !success && sm.config.FailureAnalysisEnabled {
		failureEvent := FailureEvent{
			Timestamp:  timestamp,
			Domain:     domain,
			Method:     method,
			Error:      errorMsg,
			ErrorType:  sm.classifyError(errorMsg),
			Duration:   duration,
			RetryCount: 0, // Would be set by caller if available
		}
		sm.failureAnalyzer.RecordFailure(failureEvent)
	}

	// Update performance tracker
	if sm.config.PerformanceTrackingEnabled {
		sm.performanceTracker.RecordMetric("verification_duration", float64(duration.Milliseconds()), "ms", domain, method)
		sm.performanceTracker.RecordMetric("success_rate", sm.calculateSuccessRate(), "percentage", domain, method)
	}

	// Update historical tracker
	if sm.config.HistoricalTrackingEnabled {
		sm.historicalTracker.RecordDataPoint(timestamp, success, duration, method)
	}

	// Check for alerts
	if sm.config.AlertingEnabled {
		sm.alertManager.checkAlerts()
	}

	sm.logger.Info("verification result recorded", map[string]interface{}{
		"domain":     domain,
		"method":     method,
		"success":    success,
		"duration":   duration,
		"confidence": confidence,
	})
}

// GetSuccessRate returns the current success rate
func (sm *SuccessMonitor) GetSuccessRate() float64 {
	if !sm.config.SuccessRateTrackingEnabled {
		return 0.0
	}

	return sm.successTracker.GetSuccessRate()
}

// GetFailureAnalysis returns failure analysis results
func (sm *SuccessMonitor) GetFailureAnalysis() map[string]*FailurePattern {
	if !sm.config.FailureAnalysisEnabled {
		return make(map[string]*FailurePattern)
	}

	return sm.failureAnalyzer.GetFailurePatterns()
}

// GetPerformanceMetrics returns performance metrics
func (sm *SuccessMonitor) GetPerformanceMetrics() map[string]*PerformanceMetric {
	if !sm.config.PerformanceTrackingEnabled {
		return make(map[string]*PerformanceMetric)
	}

	return sm.performanceTracker.GetMetrics()
}

// GetHistoricalData returns historical success rate data
func (sm *SuccessMonitor) GetHistoricalData(startTime, endTime time.Time) []HistoricalDataPoint {
	if !sm.config.HistoricalTrackingEnabled {
		return make([]HistoricalDataPoint, 0)
	}

	return sm.historicalTracker.GetHistoricalData(startTime, endTime)
}

// GetAlertHistory returns alert history
func (sm *SuccessMonitor) GetAlertHistory() []AlertEvent {
	if !sm.config.AlertingEnabled {
		return make([]AlertEvent, 0)
	}

	return sm.alertManager.GetAlertHistory()
}

// SuccessRateTracker methods

func (srt *SuccessRateTracker) RecordEvent(event SuccessEvent) {
	srt.windowMux.Lock()
	defer srt.windowMux.Unlock()

	// Check if we need to reset the window
	if time.Since(srt.windowStart) > srt.window {
		srt.resetWindow()
	}

	// Record the event
	if event.Success {
		srt.successCount++
	} else {
		srt.failureCount++
	}
	srt.totalCount++

	// Add to history
	srt.historyMux.Lock()
	srt.successHistory = append(srt.successHistory, event)
	srt.historyMux.Unlock()
}

func (srt *SuccessRateTracker) GetSuccessRate() float64 {
	srt.windowMux.RLock()
	defer srt.windowMux.RUnlock()

	if srt.totalCount < int64(srt.minSampleSize) {
		return 0.0
	}

	return float64(srt.successCount) / float64(srt.totalCount)
}

func (srt *SuccessRateTracker) resetWindow() {
	srt.successCount = 0
	srt.failureCount = 0
	srt.totalCount = 0
	srt.windowStart = time.Now()
}

// FailureAnalyzer methods

func (fa *FailureAnalyzer) RecordFailure(event FailureEvent) {
	fa.historyMux.Lock()
	fa.failureHistory = append(fa.failureHistory, event)
	fa.historyMux.Unlock()

	// Analyze failure pattern
	fa.analyzeFailurePattern(event)
}

func (fa *FailureAnalyzer) analyzeFailurePattern(event FailureEvent) {
	fa.patternMux.Lock()
	defer fa.patternMux.Unlock()

	// Create pattern key based on error type and method
	patternKey := fmt.Sprintf("%s:%s", event.ErrorType, event.Method)

	pattern, exists := fa.failurePatterns[patternKey]
	if !exists {
		pattern = &FailurePattern{
			Pattern:         patternKey,
			Count:           0,
			FirstSeen:       event.Timestamp,
			CommonErrors:    make([]string, 0),
			AffectedDomains: make([]string, 0),
		}
		fa.failurePatterns[patternKey] = pattern
	}

	pattern.Count++
	pattern.LastSeen = event.Timestamp

	// Add error message if not already present
	found := false
	for _, err := range pattern.CommonErrors {
		if err == event.Error {
			found = true
			break
		}
	}
	if !found {
		pattern.CommonErrors = append(pattern.CommonErrors, event.Error)
	}

	// Add domain if not already present
	found = false
	for _, domain := range pattern.AffectedDomains {
		if domain == event.Domain {
			found = true
			break
		}
	}
	if !found {
		pattern.AffectedDomains = append(pattern.AffectedDomains, event.Domain)
	}

	// Calculate failure rate
	totalEvents := len(fa.failureHistory)
	pattern.FailureRate = float64(pattern.Count) / float64(totalEvents)
}

func (fa *FailureAnalyzer) GetFailurePatterns() map[string]*FailurePattern {
	fa.patternMux.RLock()
	defer fa.patternMux.RUnlock()

	patterns := make(map[string]*FailurePattern)
	for key, pattern := range fa.failurePatterns {
		if pattern.FailureRate >= fa.patternThreshold {
			patterns[key] = pattern
		}
	}

	return patterns
}

// SuccessAlertManager methods

func (sam *SuccessAlertManager) checkAlerts() {
	sam.alertMux.Lock()
	defer sam.alertMux.Unlock()

	// Check if we're in cooldown period
	if time.Since(sam.lastAlert) < sam.cooldown {
		return
	}

	// Get current success rate (this would need to be passed in or retrieved)
	// For now, we'll assume it's available
	successRate := 0.85 // This would be retrieved from success tracker

	if successRate < sam.threshold {
		alert := AlertEvent{
			Timestamp:   time.Now(),
			Type:        "low_success_rate",
			Message:     fmt.Sprintf("Success rate %.2f%% is below threshold %.2f%%", successRate*100, sam.threshold*100),
			Severity:    "warning",
			SuccessRate: successRate,
			Threshold:   sam.threshold,
			Channels:    sam.channels,
		}

		sam.alertHistory = append(sam.alertHistory, alert)
		sam.lastAlert = time.Now()

		// Send alert (implementation would depend on alert channels)
		sam.sendAlert(alert)
	}
}

func (sam *SuccessAlertManager) sendAlert(alert AlertEvent) {
	// Implementation would send alerts to configured channels
	// For now, just log the alert
	// In production, this would integrate with email, Slack, PagerDuty, etc.
}

func (sam *SuccessAlertManager) GetAlertHistory() []AlertEvent {
	sam.historyMux.RLock()
	defer sam.historyMux.RUnlock()

	history := make([]AlertEvent, len(sam.alertHistory))
	copy(history, sam.alertHistory)

	return history
}

// VerificationPerformanceTracker methods

func (vpt *VerificationPerformanceTracker) RecordMetric(name string, value float64, unit string, domain, method string) {
	vpt.performanceMux.Lock()
	defer vpt.performanceMux.Unlock()

	metric, exists := vpt.performanceData[name]
	if !exists {
		metric = &PerformanceMetric{
			Name:         name,
			Value:        value,
			Unit:         unit,
			LastUpdated:  time.Now(),
			MinValue:     value,
			MaxValue:     value,
			AverageValue: value,
			SampleCount:  1,
		}
		vpt.performanceData[name] = metric
	} else {
		// Update metric
		metric.Value = value
		metric.LastUpdated = time.Now()
		metric.SampleCount++
		metric.AverageValue = (metric.AverageValue*float64(metric.SampleCount-1) + value) / float64(metric.SampleCount)

		if value < metric.MinValue {
			metric.MinValue = value
		}
		if value > metric.MaxValue {
			metric.MaxValue = value
		}
	}

	// Record performance event
	event := PerformanceEvent{
		Timestamp:  time.Now(),
		MetricName: name,
		Value:      value,
		Unit:       unit,
		Domain:     domain,
		Method:     method,
	}

	vpt.historyMux.Lock()
	vpt.metricHistory = append(vpt.metricHistory, event)
	vpt.historyMux.Unlock()
}

func (vpt *VerificationPerformanceTracker) GetMetrics() map[string]*PerformanceMetric {
	vpt.performanceMux.RLock()
	defer vpt.performanceMux.RUnlock()

	metrics := make(map[string]*PerformanceMetric)
	for key, metric := range vpt.performanceData {
		metrics[key] = metric
	}

	return metrics
}

// HistoricalSuccessTracker methods

func (hst *HistoricalSuccessTracker) RecordDataPoint(timestamp time.Time, success bool, duration time.Duration, method string) {
	hst.historicalMux.Lock()
	defer hst.historicalMux.Unlock()

	// Create data point key based on granularity
	key := timestamp.Truncate(hst.granularity).Format(time.RFC3339)

	dataPoint, exists := hst.historicalData[key]
	if !exists {
		dataPoint = &HistoricalDataPoint{
			Timestamp: timestamp.Truncate(hst.granularity),
			Methods:   make(map[string]MethodStats),
		}
		hst.historicalData[key] = dataPoint
	}

	// Update overall statistics
	dataPoint.TotalCount++
	if success {
		dataPoint.SuccessCount++
	} else {
		dataPoint.FailureCount++
	}
	dataPoint.SuccessRate = float64(dataPoint.SuccessCount) / float64(dataPoint.TotalCount)

	// Update average duration
	totalDuration := dataPoint.AverageDuration*time.Duration(dataPoint.TotalCount-1) + duration
	dataPoint.AverageDuration = totalDuration / time.Duration(dataPoint.TotalCount)

	// Update method statistics
	methodStats, exists := dataPoint.Methods[method]
	if !exists {
		methodStats = MethodStats{}
	}

	methodStats.SuccessCount++
	if success {
		methodStats.SuccessCount++
	} else {
		methodStats.FailureCount++
	}
	methodStats.SuccessRate = float64(methodStats.SuccessCount) / float64(methodStats.SuccessCount+methodStats.FailureCount)

	totalMethodDuration := methodStats.AverageDuration*time.Duration(methodStats.SuccessCount+methodStats.FailureCount-1) + duration
	methodStats.AverageDuration = totalMethodDuration / time.Duration(methodStats.SuccessCount+methodStats.FailureCount)

	dataPoint.Methods[method] = methodStats

	// Add to history
	hst.dataPointHistory = append(hst.dataPointHistory, *dataPoint)

	// Clean up old data points
	hst.cleanupOldData()
}

func (hst *HistoricalSuccessTracker) GetHistoricalData(startTime, endTime time.Time) []HistoricalDataPoint {
	hst.historyMux.RLock()
	defer hst.historyMux.RUnlock()

	var filteredData []HistoricalDataPoint
	for _, dataPoint := range hst.dataPointHistory {
		if dataPoint.Timestamp.After(startTime) && dataPoint.Timestamp.Before(endTime) {
			filteredData = append(filteredData, dataPoint)
		}
	}

	return filteredData
}

func (hst *HistoricalSuccessTracker) cleanupOldData() {
	cutoffTime := time.Now().Add(-hst.retentionPeriod)

	// Remove old data points from map
	for key, dataPoint := range hst.historicalData {
		if dataPoint.Timestamp.Before(cutoffTime) {
			delete(hst.historicalData, key)
		}
	}

	// Remove old data points from history
	var filteredHistory []HistoricalDataPoint
	for _, dataPoint := range hst.dataPointHistory {
		if dataPoint.Timestamp.After(cutoffTime) {
			filteredHistory = append(filteredHistory, dataPoint)
		}
	}
	hst.dataPointHistory = filteredHistory

	// Limit history size
	if len(hst.dataPointHistory) > hst.maxDataPoints {
		hst.dataPointHistory = hst.dataPointHistory[len(hst.dataPointHistory)-hst.maxDataPoints:]
	}
}

// Helper methods

func (sm *SuccessMonitor) classifyError(errorMsg string) string {
	// Simple error classification based on error message patterns
	if errorMsg == "" {
		return "unknown"
	}

	errorMsg = strings.ToLower(errorMsg)
	switch {
	case strings.Contains(errorMsg, "timeout"):
		return "timeout"
	case strings.Contains(errorMsg, "connection"):
		return "connection"
	case strings.Contains(errorMsg, "dns"):
		return "dns"
	case strings.Contains(errorMsg, "captcha"):
		return "captcha"
	case strings.Contains(errorMsg, "blocked"):
		return "blocked"
	case strings.Contains(errorMsg, "rate limit"):
		return "rate_limit"
	default:
		return "other"
	}
}

func (sm *SuccessMonitor) calculateSuccessRate() float64 {
	if !sm.config.SuccessRateTrackingEnabled {
		return 0.0
	}
	return sm.successTracker.GetSuccessRate()
}

func (sm *SuccessMonitor) startBackgroundWorkers() {
	// Start background workers for cleanup and maintenance
	go sm.cleanupWorker()
	go sm.metricsWorker()
}

func (sm *SuccessMonitor) cleanupWorker() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.performCleanup()
		}
	}
}

func (sm *SuccessMonitor) metricsWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.updateMetrics()
		}
	}
}

func (sm *SuccessMonitor) performCleanup() {
	// Clean up old data
	if sm.config.HistoricalTrackingEnabled {
		// Historical tracker handles its own cleanup
	}

	// Clean up old failure patterns
	if sm.config.FailureAnalysisEnabled {
		sm.failureAnalyzer.cleanupOldPatterns()
	}

	// Clean up old alerts
	if sm.config.AlertingEnabled {
		sm.alertManager.cleanupOldAlerts()
	}
}

func (sm *SuccessMonitor) updateMetrics() {
	// Update performance metrics
	if sm.config.PerformanceTrackingEnabled {
		// Performance tracker updates metrics in real-time
	}

	// Update success rate metrics
	if sm.config.SuccessRateTrackingEnabled {
		successRate := sm.successTracker.GetSuccessRate()
		sm.performanceTracker.RecordMetric("current_success_rate", successRate*100, "percentage", "", "")
	}
}

// Additional cleanup methods for components

func (fa *FailureAnalyzer) cleanupOldPatterns() {
	fa.patternMux.Lock()
	defer fa.patternMux.Unlock()

	cutoffTime := time.Now().Add(-fa.window)
	for key, pattern := range fa.failurePatterns {
		if pattern.LastSeen.Before(cutoffTime) {
			delete(fa.failurePatterns, key)
		}
	}
}

func (sam *SuccessAlertManager) cleanupOldAlerts() {
	sam.historyMux.Lock()
	defer sam.historyMux.Unlock()

	cutoffTime := time.Now().Add(-24 * time.Hour) // Keep alerts for 24 hours
	var filteredAlerts []AlertEvent
	for _, alert := range sam.alertHistory {
		if alert.Timestamp.After(cutoffTime) {
			filteredAlerts = append(filteredAlerts, alert)
		}
	}
	sam.alertHistory = filteredAlerts
}

// GetMonitoringStatistics returns comprehensive monitoring statistics
func (sm *SuccessMonitor) GetMonitoringStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Success rate statistics
	if sm.config.SuccessRateTrackingEnabled {
		stats["success_rate"] = map[string]interface{}{
			"current_rate":  sm.successTracker.GetSuccessRate(),
			"total_count":   sm.successTracker.totalCount,
			"success_count": sm.successTracker.successCount,
			"failure_count": sm.successTracker.failureCount,
			"window_start":  sm.successTracker.windowStart,
		}
	}

	// Failure analysis statistics
	if sm.config.FailureAnalysisEnabled {
		patterns := sm.failureAnalyzer.GetFailurePatterns()
		stats["failure_analysis"] = map[string]interface{}{
			"total_patterns": len(patterns),
			"patterns":       patterns,
		}
	}

	// Performance statistics
	if sm.config.PerformanceTrackingEnabled {
		metrics := sm.performanceTracker.GetMetrics()
		stats["performance"] = map[string]interface{}{
			"metrics": metrics,
		}
	}

	// Historical statistics
	if sm.config.HistoricalTrackingEnabled {
		stats["historical"] = map[string]interface{}{
			"data_points":      len(sm.historicalTracker.dataPointHistory),
			"retention_period": sm.historicalTracker.retentionPeriod,
		}
	}

	// Alert statistics
	if sm.config.AlertingEnabled {
		alerts := sm.alertManager.GetAlertHistory()
		stats["alerts"] = map[string]interface{}{
			"total_alerts":  len(alerts),
			"recent_alerts": len(alerts), // Could filter for recent alerts
		}
	}

	return stats
}

// Shutdown shuts down the success monitor
func (sm *SuccessMonitor) Shutdown() {
	sm.logger.Info("success monitor shutting down", map[string]interface{}{})
}
