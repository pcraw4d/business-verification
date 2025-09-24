package error_monitoring

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrorRateMonitor provides comprehensive error rate monitoring for verification processes
type ErrorRateMonitor struct {
	config          *ErrorMonitoringConfig
	logger          *zap.Logger
	mu              sync.RWMutex
	processStats    map[string]*ProcessErrorStats
	globalStats     *GlobalErrorStats
	alertManager    AlertManager
	metricCollector MetricCollector
}

// ErrorMonitoringConfig contains configuration for error monitoring
type ErrorMonitoringConfig struct {
	MaxErrorRate           float64                            `json:"max_error_rate"`          // Target: <5%
	CriticalErrorRate      float64                            `json:"critical_error_rate"`     // 10%
	WarningErrorRate       float64                            `json:"warning_error_rate"`      // 7%
	MonitoringWindow       time.Duration                      `json:"monitoring_window"`       // 15 minutes
	AlertCooldownPeriod    time.Duration                      `json:"alert_cooldown_period"`   // 5 minutes
	MetricRetentionPeriod  time.Duration                      `json:"metric_retention_period"` // 24 hours
	EnableRealTimeAlerts   bool                               `json:"enable_real_time_alerts"`
	EnableTrendAnalysis    bool                               `json:"enable_trend_analysis"`
	EnablePredictiveAlerts bool                               `json:"enable_predictive_alerts"`
	ProcessMonitoring      map[string]ProcessMonitoringConfig `json:"process_monitoring"`
}

// ProcessMonitoringConfig contains monitoring configuration for specific processes
type ProcessMonitoringConfig struct {
	Enabled           bool    `json:"enabled"`
	MaxErrorRate      float64 `json:"max_error_rate"`
	SampleSize        int     `json:"sample_size"`
	AlertThreshold    int     `json:"alert_threshold"`
	CriticalThreshold int     `json:"critical_threshold"`
}

// ProcessErrorStats contains error statistics for a specific verification process
type ProcessErrorStats struct {
	ProcessName        string                    `json:"process_name"`
	TotalRequests      int64                     `json:"total_requests"`
	TotalErrors        int64                     `json:"total_errors"`
	ErrorRate          float64                   `json:"error_rate"`
	LastErrorTime      time.Time                 `json:"last_error_time"`
	LastSuccessTime    time.Time                 `json:"last_success_time"`
	ErrorsByType       map[string]int64          `json:"errors_by_type"`
	ErrorsByCategory   map[string]int64          `json:"errors_by_category"`
	RecentErrors       []ErrorEntry              `json:"recent_errors"`
	ErrorTrend         string                    `json:"error_trend"`
	PerformanceMetrics ProcessPerformanceMetrics `json:"performance_metrics"`
	WindowedStats      []WindowedErrorStats      `json:"windowed_stats"`
	AlertStatus        AlertStatus               `json:"alert_status"`
	LastUpdated        time.Time                 `json:"last_updated"`
}

// GlobalErrorStats contains overall system error statistics
type GlobalErrorStats struct {
	OverallErrorRate   float64            `json:"overall_error_rate"`
	TotalRequests      int64              `json:"total_requests"`
	TotalErrors        int64              `json:"total_errors"`
	ErrorRateByProcess map[string]float64 `json:"error_rate_by_process"`
	ErrorTrend         string             `json:"error_trend"`
	HealthStatus       string             `json:"health_status"`
	LastUpdated        time.Time          `json:"last_updated"`
	TopErrorCategories map[string]int64   `json:"top_error_categories"`
	CriticalProcesses  []string           `json:"critical_processes"`
}

// ErrorEntry represents a single error occurrence
type ErrorEntry struct {
	Timestamp     time.Time              `json:"timestamp"`
	ProcessName   string                 `json:"process_name"`
	ErrorType     string                 `json:"error_type"`
	ErrorCategory string                 `json:"error_category"`
	ErrorMessage  string                 `json:"error_message"`
	Severity      string                 `json:"severity"`
	Context       map[string]interface{} `json:"context"`
	RequestID     string                 `json:"request_id"`
	UserID        string                 `json:"user_id"`
	Duration      time.Duration          `json:"duration"`
	RetryCount    int                    `json:"retry_count"`
}

// ProcessPerformanceMetrics contains performance metrics for a process
type ProcessPerformanceMetrics struct {
	AverageResponseTime time.Duration `json:"average_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`
	Throughput          float64       `json:"throughput"`
	SuccessRate         float64       `json:"success_rate"`
	RetryRate           float64       `json:"retry_rate"`
}

// WindowedErrorStats contains error statistics for a specific time window
type WindowedErrorStats struct {
	WindowStart    time.Time     `json:"window_start"`
	WindowEnd      time.Time     `json:"window_end"`
	Requests       int64         `json:"requests"`
	Errors         int64         `json:"errors"`
	ErrorRate      float64       `json:"error_rate"`
	AverageLatency time.Duration `json:"average_latency"`
}

// AlertStatus represents the current alert status for a process
type AlertStatus struct {
	Level                 string    `json:"level"` // none, warning, critical
	Active                bool      `json:"active"`
	LastTriggered         time.Time `json:"last_triggered"`
	LastCleared           time.Time `json:"last_cleared"`
	ConsecutiveViolations int       `json:"consecutive_violations"`
	Message               string    `json:"message"`
}

// ErrorRateReport contains comprehensive error rate analysis
type ErrorRateReport struct {
	ReportTimestamp time.Time                     `json:"report_timestamp"`
	ReportPeriod    string                        `json:"report_period"`
	GlobalStats     *GlobalErrorStats             `json:"global_stats"`
	ProcessStats    map[string]*ProcessErrorStats `json:"process_stats"`
	Trends          *ErrorRateTrendAnalysis       `json:"trends"`
	Recommendations []string                      `json:"recommendations"`
	Alerts          []Alert                       `json:"alerts"`
	Compliance      ComplianceStatus              `json:"compliance"`
}

// ErrorRateTrendAnalysis contains trend analysis results
type ErrorRateTrendAnalysis struct {
	GlobalTrend        string             `json:"global_trend"`
	ProcessTrends      map[string]string  `json:"process_trends"`
	PredictedErrorRate float64            `json:"predicted_error_rate"`
	TrendConfidence    float64            `json:"trend_confidence"`
	Seasonality        map[string]float64 `json:"seasonality"`
}

// ComplianceStatus represents compliance with error rate targets
type ComplianceStatus struct {
	IsCompliant        bool      `json:"is_compliant"`
	ComplianceScore    float64   `json:"compliance_score"`
	ViolatingProcesses []string  `json:"violating_processes"`
	DaysInCompliance   int       `json:"days_in_compliance"`
	LastViolation      time.Time `json:"last_violation"`
}

// Alert represents an error rate alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Level       string                 `json:"level"`
	ProcessName string                 `json:"process_name"`
	Message     string                 `json:"message"`
	Timestamp   time.Time              `json:"timestamp"`
	Context     map[string]interface{} `json:"context"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at"`
}

// AlertManager interface for handling alerts
type AlertManager interface {
	SendAlert(ctx context.Context, alert Alert) error
	ClearAlert(ctx context.Context, alertID string) error
	GetActiveAlerts(ctx context.Context) ([]Alert, error)
}

// MetricCollector interface for collecting and storing metrics
type MetricCollector interface {
	RecordErrorRate(processName string, errorRate float64, timestamp time.Time) error
	RecordError(errorEntry ErrorEntry) error
	RecordSuccess(processName string, duration time.Duration, timestamp time.Time) error
	GetMetrics(processName string, start, end time.Time) ([]MetricPoint, error)
}

// MetricPoint represents a single metric data point
type MetricPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Labels    map[string]string      `json:"labels"`
	Context   map[string]interface{} `json:"context"`
}

// NewErrorRateMonitor creates a new error rate monitor
func NewErrorRateMonitor(config *ErrorMonitoringConfig, logger *zap.Logger, alertManager AlertManager, metricCollector MetricCollector) *ErrorRateMonitor {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = getDefaultConfig()
	}

	return &ErrorRateMonitor{
		config:       config,
		logger:       logger,
		processStats: make(map[string]*ProcessErrorStats),
		globalStats: &GlobalErrorStats{
			ErrorRateByProcess: make(map[string]float64),
			TopErrorCategories: make(map[string]int64),
			CriticalProcesses:  make([]string, 0),
		},
		alertManager:    alertManager,
		metricCollector: metricCollector,
	}
}

// getDefaultConfig returns default configuration
func getDefaultConfig() *ErrorMonitoringConfig {
	return &ErrorMonitoringConfig{
		MaxErrorRate:           0.05, // 5%
		CriticalErrorRate:      0.10, // 10%
		WarningErrorRate:       0.07, // 7%
		MonitoringWindow:       15 * time.Minute,
		AlertCooldownPeriod:    5 * time.Minute,
		MetricRetentionPeriod:  24 * time.Hour,
		EnableRealTimeAlerts:   true,
		EnableTrendAnalysis:    true,
		EnablePredictiveAlerts: true,
		ProcessMonitoring: map[string]ProcessMonitoringConfig{
			"business_verification": {
				Enabled:           true,
				MaxErrorRate:      0.05,
				SampleSize:        100,
				AlertThreshold:    5,
				CriticalThreshold: 10,
			},
			"website_analysis": {
				Enabled:           true,
				MaxErrorRate:      0.05,
				SampleSize:        100,
				AlertThreshold:    5,
				CriticalThreshold: 10,
			},
			"risk_assessment": {
				Enabled:           true,
				MaxErrorRate:      0.05,
				SampleSize:        100,
				AlertThreshold:    5,
				CriticalThreshold: 10,
			},
			"data_discovery": {
				Enabled:           true,
				MaxErrorRate:      0.05,
				SampleSize:        100,
				AlertThreshold:    5,
				CriticalThreshold: 10,
			},
		},
	}
}

// RecordError records an error occurrence for a verification process
func (erm *ErrorRateMonitor) RecordError(ctx context.Context, processName string, errorType string, errorMessage string, severity string, context map[string]interface{}) error {
	erm.mu.Lock()
	defer erm.mu.Unlock()

	now := time.Now()

	// Get or create process stats
	stats := erm.getOrCreateProcessStats(processName)

	// Update process statistics
	stats.TotalRequests++
	stats.TotalErrors++
	stats.LastErrorTime = now
	stats.LastUpdated = now

	// Update error counts by type and category
	stats.ErrorsByType[errorType]++
	category := erm.categorizeError(errorType)
	stats.ErrorsByCategory[category]++

	// Calculate error rate
	stats.ErrorRate = float64(stats.TotalErrors) / float64(stats.TotalRequests)

	// Create error entry
	errorEntry := ErrorEntry{
		Timestamp:     now,
		ProcessName:   processName,
		ErrorType:     errorType,
		ErrorCategory: category,
		ErrorMessage:  errorMessage,
		Severity:      severity,
		Context:       context,
		RequestID:     erm.getRequestID(context),
		UserID:        erm.getUserID(context),
		Duration:      erm.getDuration(context),
		RetryCount:    erm.getRetryCount(context),
	}

	// Add to recent errors (keep last 100)
	stats.RecentErrors = append(stats.RecentErrors, errorEntry)
	if len(stats.RecentErrors) > 100 {
		stats.RecentErrors = stats.RecentErrors[1:]
	}

	// Update windowed statistics
	erm.updateWindowedStats(stats, now, false)

	// Update error trend
	stats.ErrorTrend = erm.calculateErrorTrend(stats)

	// Check alert conditions
	erm.checkAlertConditions(ctx, stats)

	// Update global statistics
	erm.updateGlobalStats()

	// Record metric
	if erm.metricCollector != nil {
		if err := erm.metricCollector.RecordError(errorEntry); err != nil {
			erm.logger.Error("Failed to record error metric", zap.Error(err))
		}
	}

	erm.logger.Warn("Error recorded",
		zap.String("process", processName),
		zap.String("error_type", errorType),
		zap.String("error_message", errorMessage),
		zap.String("severity", severity),
		zap.Float64("error_rate", stats.ErrorRate))

	return nil
}

// RecordSuccess records a successful verification process execution
func (erm *ErrorRateMonitor) RecordSuccess(ctx context.Context, processName string, duration time.Duration, context map[string]interface{}) error {
	erm.mu.Lock()
	defer erm.mu.Unlock()

	now := time.Now()

	// Get or create process stats
	stats := erm.getOrCreateProcessStats(processName)

	// Update process statistics
	stats.TotalRequests++
	stats.LastSuccessTime = now
	stats.LastUpdated = now

	// Calculate error rate
	stats.ErrorRate = float64(stats.TotalErrors) / float64(stats.TotalRequests)

	// Update windowed statistics
	erm.updateWindowedStats(stats, now, true)

	// Update performance metrics
	erm.updatePerformanceMetrics(stats, duration)

	// Update error trend
	stats.ErrorTrend = erm.calculateErrorTrend(stats)

	// Check if alerts should be cleared
	erm.checkAlertClearance(ctx, stats)

	// Update global statistics
	erm.updateGlobalStats()

	// Record metric
	if erm.metricCollector != nil {
		if err := erm.metricCollector.RecordSuccess(processName, duration, now); err != nil {
			erm.logger.Error("Failed to record success metric", zap.Error(err))
		}
	}

	return nil
}

// GetErrorRateReport generates a comprehensive error rate report
func (erm *ErrorRateMonitor) GetErrorRateReport(ctx context.Context, period string) (*ErrorRateReport, error) {
	erm.mu.RLock()
	defer erm.mu.RUnlock()

	report := &ErrorRateReport{
		ReportTimestamp: time.Now(),
		ReportPeriod:    period,
		GlobalStats:     erm.globalStats,
		ProcessStats:    make(map[string]*ProcessErrorStats),
		Recommendations: make([]string, 0),
		Alerts:          make([]Alert, 0),
	}

	// Copy process stats
	for processName, stats := range erm.processStats {
		report.ProcessStats[processName] = stats
	}

	// Generate trend analysis
	if erm.config.EnableTrendAnalysis {
		report.Trends = erm.generateTrendAnalysis()
	}

	// Generate recommendations
	report.Recommendations = erm.generateRecommendations()

	// Get active alerts
	if erm.alertManager != nil {
		if alerts, err := erm.alertManager.GetActiveAlerts(ctx); err == nil {
			report.Alerts = alerts
		}
	}

	// Check compliance
	report.Compliance = erm.checkCompliance()

	return report, nil
}

// IsErrorRateCompliant checks if the system is compliant with error rate targets
func (erm *ErrorRateMonitor) IsErrorRateCompliant() bool {
	erm.mu.RLock()
	defer erm.mu.RUnlock()

	// Check global error rate
	if erm.globalStats.OverallErrorRate > erm.config.MaxErrorRate {
		return false
	}

	// Check individual process error rates
	for processName, stats := range erm.processStats {
		processConfig, exists := erm.config.ProcessMonitoring[processName]
		if !exists {
			processConfig = ProcessMonitoringConfig{MaxErrorRate: erm.config.MaxErrorRate}
		}

		if stats.ErrorRate > processConfig.MaxErrorRate {
			return false
		}
	}

	return true
}

// GetProcessErrorRate returns the current error rate for a specific process
func (erm *ErrorRateMonitor) GetProcessErrorRate(processName string) (float64, bool) {
	erm.mu.RLock()
	defer erm.mu.RUnlock()

	if stats, exists := erm.processStats[processName]; exists {
		return stats.ErrorRate, true
	}
	return 0.0, false
}

// GetGlobalErrorRate returns the current global error rate
func (erm *ErrorRateMonitor) GetGlobalErrorRate() float64 {
	erm.mu.RLock()
	defer erm.mu.RUnlock()

	return erm.globalStats.OverallErrorRate
}

// ResetProcessStats resets error statistics for a specific process
func (erm *ErrorRateMonitor) ResetProcessStats(processName string) {
	erm.mu.Lock()
	defer erm.mu.Unlock()

	if stats, exists := erm.processStats[processName]; exists {
		stats.TotalRequests = 0
		stats.TotalErrors = 0
		stats.ErrorRate = 0.0
		stats.ErrorsByType = make(map[string]int64)
		stats.ErrorsByCategory = make(map[string]int64)
		stats.RecentErrors = make([]ErrorEntry, 0)
		stats.ErrorTrend = "stable"
		stats.WindowedStats = make([]WindowedErrorStats, 0)
		stats.AlertStatus = AlertStatus{Level: "none", Active: false}
		stats.LastUpdated = time.Now()
	}
}

// getOrCreateProcessStats gets or creates process statistics
func (erm *ErrorRateMonitor) getOrCreateProcessStats(processName string) *ProcessErrorStats {
	stats, exists := erm.processStats[processName]
	if !exists {
		stats = &ProcessErrorStats{
			ProcessName:        processName,
			ErrorsByType:       make(map[string]int64),
			ErrorsByCategory:   make(map[string]int64),
			RecentErrors:       make([]ErrorEntry, 0),
			ErrorTrend:         "stable",
			WindowedStats:      make([]WindowedErrorStats, 0),
			AlertStatus:        AlertStatus{Level: "none", Active: false},
			PerformanceMetrics: ProcessPerformanceMetrics{},
			LastUpdated:        time.Now(),
		}
		erm.processStats[processName] = stats
	}
	return stats
}

// categorizeError categorizes an error based on its type
func (erm *ErrorRateMonitor) categorizeError(errorType string) string {
	categoryMap := map[string]string{
		"network_error":        "connectivity",
		"timeout_error":        "performance",
		"validation_error":     "data_quality",
		"authentication_error": "security",
		"authorization_error":  "security",
		"rate_limit_error":     "capacity",
		"server_error":         "system",
		"client_error":         "data_quality",
		"configuration_error":  "system",
		"dependency_error":     "external",
		"parsing_error":        "data_quality",
		"database_error":       "system",
		"api_error":            "external",
	}

	if category, exists := categoryMap[errorType]; exists {
		return category
	}
	return "unknown"
}

// getRequestID extracts request ID from context
func (erm *ErrorRateMonitor) getRequestID(context map[string]interface{}) string {
	if requestID, exists := context["request_id"]; exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return ""
}

// getUserID extracts user ID from context
func (erm *ErrorRateMonitor) getUserID(context map[string]interface{}) string {
	if userID, exists := context["user_id"]; exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return ""
}

// getDuration extracts duration from context
func (erm *ErrorRateMonitor) getDuration(context map[string]interface{}) time.Duration {
	if duration, exists := context["duration"]; exists {
		if d, ok := duration.(time.Duration); ok {
			return d
		}
		if d, ok := duration.(int64); ok {
			return time.Duration(d)
		}
	}
	return 0
}

// getRetryCount extracts retry count from context
func (erm *ErrorRateMonitor) getRetryCount(context map[string]interface{}) int {
	if retryCount, exists := context["retry_count"]; exists {
		if count, ok := retryCount.(int); ok {
			return count
		}
	}
	return 0
}

// updateWindowedStats updates windowed statistics
func (erm *ErrorRateMonitor) updateWindowedStats(stats *ProcessErrorStats, timestamp time.Time, isSuccess bool) {
	windowDuration := erm.config.MonitoringWindow
	windowStart := timestamp.Truncate(windowDuration)
	windowEnd := windowStart.Add(windowDuration)

	// Find or create window
	var window *WindowedErrorStats
	for i := range stats.WindowedStats {
		if stats.WindowedStats[i].WindowStart == windowStart {
			window = &stats.WindowedStats[i]
			break
		}
	}

	if window == nil {
		newWindow := WindowedErrorStats{
			WindowStart: windowStart,
			WindowEnd:   windowEnd,
		}
		stats.WindowedStats = append(stats.WindowedStats, newWindow)
		window = &stats.WindowedStats[len(stats.WindowedStats)-1]
	}

	// Update window statistics
	window.Requests++
	if !isSuccess {
		window.Errors++
	}
	window.ErrorRate = float64(window.Errors) / float64(window.Requests)

	// Clean up old windows (keep last 24 hours)
	cutoff := timestamp.Add(-erm.config.MetricRetentionPeriod)
	var filteredWindows []WindowedErrorStats
	for _, w := range stats.WindowedStats {
		if w.WindowStart.After(cutoff) {
			filteredWindows = append(filteredWindows, w)
		}
	}
	stats.WindowedStats = filteredWindows
}

// updatePerformanceMetrics updates performance metrics
func (erm *ErrorRateMonitor) updatePerformanceMetrics(stats *ProcessErrorStats, duration time.Duration) {
	// Simple moving average for response time
	if stats.PerformanceMetrics.AverageResponseTime == 0 {
		stats.PerformanceMetrics.AverageResponseTime = duration
	} else {
		// Exponential moving average with alpha = 0.1
		alpha := 0.1
		current := float64(stats.PerformanceMetrics.AverageResponseTime)
		new := float64(duration)
		stats.PerformanceMetrics.AverageResponseTime = time.Duration(alpha*new + (1-alpha)*current)
	}

	// Update success rate
	stats.PerformanceMetrics.SuccessRate = 1.0 - stats.ErrorRate

	// Update throughput (requests per second)
	if len(stats.WindowedStats) > 0 {
		recentWindow := stats.WindowedStats[len(stats.WindowedStats)-1]
		windowDurationSeconds := erm.config.MonitoringWindow.Seconds()
		stats.PerformanceMetrics.Throughput = float64(recentWindow.Requests) / windowDurationSeconds
	}
}

// calculateErrorTrend calculates the error trend for a process
func (erm *ErrorRateMonitor) calculateErrorTrend(stats *ProcessErrorStats) string {
	if len(stats.WindowedStats) < 3 {
		return "stable"
	}

	// Analyze last 3 windows
	recentWindows := stats.WindowedStats[len(stats.WindowedStats)-3:]

	// Calculate trend
	increasing := 0
	decreasing := 0

	for i := 1; i < len(recentWindows); i++ {
		if recentWindows[i].ErrorRate > recentWindows[i-1].ErrorRate {
			increasing++
		} else if recentWindows[i].ErrorRate < recentWindows[i-1].ErrorRate {
			decreasing++
		}
	}

	if increasing > decreasing {
		return "increasing"
	} else if decreasing > increasing {
		return "decreasing"
	}
	return "stable"
}
