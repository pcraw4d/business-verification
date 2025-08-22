package risk_assessment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ExternalAPIRateLimiter provides advanced rate limiting for external API calls
type ExternalAPIRateLimiter struct {
	config *ExternalRateLimitConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// Per-API rate limits
	apiLimits map[string]*ExternalAPILimit

	// Global rate limiting
	globalLimits *GlobalRateLimit

	// Monitoring and alerting
	monitor *RateLimitMonitor

	// Fallback strategies
	fallback *RateLimitFallback

	// Optimization and caching
	optimizer *RateLimitOptimizer
}

// ExternalRateLimitConfig contains configuration for external API rate limiting
type ExternalRateLimitConfig struct {
	// Global settings
	GlobalRequestsPerMinute int           `json:"global_requests_per_minute"`
	GlobalRequestsPerHour   int           `json:"global_requests_per_hour"`
	GlobalRequestsPerDay    int           `json:"global_requests_per_day"`
	DefaultTimeout          time.Duration `json:"default_timeout"`

	// Per-API configurations
	APIConfigs map[string]*APIConfig `json:"api_configs"`

	// Monitoring settings
	MonitorConfig *MonitorConfig `json:"monitor_config"`

	// Fallback settings
	FallbackConfig *FallbackConfig `json:"fallback_config"`

	// Optimization settings
	OptimizationConfig *OptimizationConfig `json:"optimization_config"`
}

// APIConfig contains configuration for a specific API
type APIConfig struct {
	APIEndpoint       string        `json:"api_endpoint"`
	RequestsPerMinute int           `json:"requests_per_minute"`
	RequestsPerHour   int           `json:"requests_per_hour"`
	RequestsPerDay    int           `json:"requests_per_day"`
	Timeout           time.Duration `json:"timeout"`
	Priority          int           `json:"priority"` // Higher number = higher priority
	RetryAttempts     int           `json:"retry_attempts"`
	BackoffStrategy   string        `json:"backoff_strategy"` // linear, exponential, jitter
	QuotaExceeded     bool          `json:"quota_exceeded"`
	Enabled           bool          `json:"enabled"`
}

// ExternalAPILimit contains rate limit information for an external API
type ExternalAPILimit struct {
	Config *APIConfig

	// Current usage
	CurrentRequestsPerMinute int
	CurrentRequestsPerHour   int
	CurrentRequestsPerDay    int

	// Timestamps
	LastMinuteReset  time.Time
	LastHourReset    time.Time
	LastDayReset     time.Time

	// Status
	QuotaExceeded bool
	RetryAfter    time.Time
	LastError     error
	LastSuccess   time.Time

	// Statistics
	TotalRequests    int64
	SuccessfulRequests int64
	FailedRequests   int64
	AverageResponseTime time.Duration
}

// GlobalRateLimit contains global rate limiting information
type GlobalRateLimit struct {
	CurrentRequestsPerMinute int
	CurrentRequestsPerHour   int
	CurrentRequestsPerDay    int

	LastMinuteReset time.Time
	LastHourReset   time.Time
	LastDayReset    time.Time

	QuotaExceeded bool
	RetryAfter    time.Time
}

// ExternalRateLimitResult contains the result of an external API rate limit check
type ExternalRateLimitResult struct {
	Allowed           bool
	APIEndpoint       string
	RemainingRequests int
	ResetTime         time.Time
	RetryAfter        time.Time
	QuotaExceeded     bool
	WaitTime          time.Duration
	Priority          int
	FallbackAvailable bool
	CacheHit          bool
}

// MonitorConfig contains monitoring configuration
type MonitorConfig struct {
	Enabled              bool          `json:"enabled"`
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval"`
	AlertThreshold       float64       `json:"alert_threshold"`
	AlertCooldown        time.Duration `json:"alert_cooldown"`
	
	// Alert thresholds
	QuotaExceededThreshold    float64 `json:"quota_exceeded_threshold"`
	HighUsageThreshold        float64 `json:"high_usage_threshold"`
	LowSuccessRateThreshold   float64 `json:"low_success_rate_threshold"`
	HighLatencyThreshold      time.Duration `json:"high_latency_threshold"`
	
	// Metrics retention
	MetricsRetentionDays int `json:"metrics_retention_days"`
	
	// Alert retention
	AlertRetentionDays int `json:"alert_retention_days"`
}

// FallbackConfig contains fallback strategy configuration
type FallbackConfig struct {
	Enabled           bool     `json:"enabled"`
	FallbackAPIs      []string `json:"fallback_apis"`
	CacheFallback     bool     `json:"cache_fallback"`
	RetryWithBackoff  bool     `json:"retry_with_backoff"`
	MaxRetryAttempts  int      `json:"max_retry_attempts"`
}

// OptimizationConfig contains optimization configuration
type OptimizationConfig struct {
	Enabled           bool          `json:"enabled"`
	CacheEnabled      bool          `json:"cache_enabled"`
	CacheTTL          time.Duration `json:"cache_ttl"`
	RequestBatching   bool          `json:"request_batching"`
	BatchSize         int           `json:"batch_size"`
	BatchTimeout      time.Duration `json:"batch_timeout"`
}

// RateLimitMonitor provides comprehensive monitoring and alerting for rate limits
type RateLimitMonitor struct {
	config *MonitorConfig
	logger *zap.Logger
	mu     sync.RWMutex

	// Metrics storage
	metrics map[string]*RateLimitMetrics

	// Alert management
	alerts map[string]*RateLimitAlert
	alertHistory []*RateLimitAlert

	// Background monitoring
	stopChan chan struct{}
	monitoringActive bool

	// Alert handlers
	alertHandlers []AlertHandler
}

// RateLimitMetrics contains detailed metrics for rate limiting
type RateLimitMetrics struct {
	APIEndpoint        string
	TotalChecks        int64
	AllowedRequests    int64
	BlockedRequests    int64
	AverageWaitTime    time.Duration
	LastAlertTime      time.Time
	LastCheckTime      time.Time
	LastSuccessTime    time.Time
	LastFailureTime    time.Time

	// Rate limit specific metrics
	RateLimitHits      int64
	QuotaExceededCount int64
	AverageResponseTime time.Duration
	SuccessRate        float64
	ErrorRate          float64

	// Time-based metrics
	MinuteMetrics *TimeWindowMetrics
	HourMetrics   *TimeWindowMetrics
	DayMetrics    *TimeWindowMetrics
}

// TimeWindowMetrics contains metrics for a specific time window
type TimeWindowMetrics struct {
	WindowStart    time.Time
	WindowEnd      time.Time
	RequestCount   int64
	SuccessCount   int64
	FailureCount   int64
	AverageLatency time.Duration
	PeakLatency    time.Duration
	MinLatency     time.Duration
}

// RateLimitAlert represents a rate limit alert
type RateLimitAlert struct {
	ID          string
	APIEndpoint string
	AlertType   AlertType
	Severity    AlertSeverity
	Message     string
	Timestamp   time.Time
	Acknowledged bool
	Resolved    bool
	ResolvedAt  *time.Time
	Metadata    map[string]interface{}
}

// AlertType represents the type of rate limit alert
type AlertType string

const (
	AlertTypeQuotaExceeded    AlertType = "quota_exceeded"
	AlertTypeHighUsage        AlertType = "high_usage"
	AlertTypeLowSuccessRate   AlertType = "low_success_rate"
	AlertTypeHighLatency      AlertType = "high_latency"
	AlertTypeGlobalLimitHit   AlertType = "global_limit_hit"
	AlertTypeFallbackUsed     AlertType = "fallback_used"
	AlertTypeCacheMiss        AlertType = "cache_miss"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertHandler is a function that handles rate limit alerts
type AlertHandler func(alert *RateLimitAlert) error

// NewRateLimitMonitor creates a new rate limit monitor
func NewRateLimitMonitor(config *MonitorConfig, logger *zap.Logger) *RateLimitMonitor {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = DefaultMonitorConfig()
	}

	monitor := &RateLimitMonitor{
		config:       config,
		logger:       logger,
		metrics:      make(map[string]*RateLimitMetrics),
		alerts:       make(map[string]*RateLimitAlert),
		alertHistory: make([]*RateLimitAlert, 0),
		stopChan:     make(chan struct{}),
		alertHandlers: make([]AlertHandler, 0),
	}

	// Start background monitoring if enabled
	if config.Enabled {
		monitor.StartMonitoring()
	}

	return monitor
}

// DefaultMonitorConfig returns default monitoring configuration
func DefaultMonitorConfig() *MonitorConfig {
	return &MonitorConfig{
		Enabled:              true,
		MetricsCollectionInterval: 30 * time.Second,
		AlertThreshold:       0.8,
		AlertCooldown:        5 * time.Minute,
		QuotaExceededThreshold: 0.1,  // 10% quota exceeded
		HighUsageThreshold:   0.8,    // 80% usage
		LowSuccessRateThreshold: 0.9, // 90% success rate
		HighLatencyThreshold: 5 * time.Second,
		MetricsRetentionDays: 30,
		AlertRetentionDays:   90,
	}
}

// StartMonitoring starts background monitoring
func (rlm *RateLimitMonitor) StartMonitoring() {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	if rlm.monitoringActive {
		return
	}

	rlm.monitoringActive = true
	go rlm.monitoringLoop()
}

// StopMonitoring stops background monitoring
func (rlm *RateLimitMonitor) StopMonitoring() {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	if !rlm.monitoringActive {
		return
	}

	rlm.monitoringActive = false
	close(rlm.stopChan)
}

// AddAlertHandler adds an alert handler
func (rlm *RateLimitMonitor) AddAlertHandler(handler AlertHandler) {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	rlm.alertHandlers = append(rlm.alertHandlers, handler)
}

// RecordRateLimitCheck records a rate limit check
func (rlm *RateLimitMonitor) RecordRateLimitCheck(apiEndpoint string, result *ExternalRateLimitResult) {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	metrics, exists := rlm.metrics[apiEndpoint]
	if !exists {
		metrics = &RateLimitMetrics{
			APIEndpoint: apiEndpoint,
			MinuteMetrics: &TimeWindowMetrics{},
			HourMetrics:   &TimeWindowMetrics{},
			DayMetrics:    &TimeWindowMetrics{},
		}
		rlm.metrics[apiEndpoint] = metrics
	}

	now := time.Now()
	metrics.LastCheckTime = now
	metrics.TotalChecks++

	if result.Allowed {
		metrics.AllowedRequests++
		metrics.LastSuccessTime = now
	} else {
		metrics.BlockedRequests++
		metrics.LastFailureTime = now
		metrics.RateLimitHits++
	}

	if result.QuotaExceeded {
		metrics.QuotaExceededCount++
	}

	// Update average wait time
	if result.WaitTime > 0 {
		if metrics.AverageWaitTime == 0 {
			metrics.AverageWaitTime = result.WaitTime
		} else {
			metrics.AverageWaitTime = (metrics.AverageWaitTime + result.WaitTime) / 2
		}
	}

	// Update success rate
	if metrics.TotalChecks > 0 {
		metrics.SuccessRate = float64(metrics.AllowedRequests) / float64(metrics.TotalChecks)
		metrics.ErrorRate = 1.0 - metrics.SuccessRate
	}

	// Update time window metrics
	rlm.updateTimeWindowMetrics(metrics, now, result)

	// Check for alerts
	rlm.checkForAlerts(apiEndpoint, metrics, result)
}

// RecordAPICall records an API call
func (rlm *RateLimitMonitor) RecordAPICall(apiEndpoint string, success bool, responseTime time.Duration, err error) {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	metrics, exists := rlm.metrics[apiEndpoint]
	if !exists {
		metrics = &RateLimitMetrics{
			APIEndpoint: apiEndpoint,
			MinuteMetrics: &TimeWindowMetrics{},
			HourMetrics:   &TimeWindowMetrics{},
			DayMetrics:    &TimeWindowMetrics{},
		}
		rlm.metrics[apiEndpoint] = metrics
	}

	now := time.Now()
	if success {
		metrics.LastSuccessTime = now
	} else {
		metrics.LastFailureTime = now
	}

	// Update average response time
	if metrics.AverageResponseTime == 0 {
		metrics.AverageResponseTime = responseTime
	} else {
		metrics.AverageResponseTime = (metrics.AverageResponseTime + responseTime) / 2
	}

	// Update time window metrics for API calls
	rlm.updateTimeWindowMetricsForAPICall(metrics, now, success, responseTime)

	rlm.logger.Debug("Recording API call",
		zap.String("api_endpoint", apiEndpoint),
		zap.Bool("success", success),
		zap.Duration("response_time", responseTime),
		zap.Error(err))
}

// GetMetrics returns metrics for an API endpoint
func (rlm *RateLimitMonitor) GetMetrics(apiEndpoint string) *RateLimitMetrics {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	if metrics, exists := rlm.metrics[apiEndpoint]; exists {
		return metrics
	}
	return nil
}

// GetAllMetrics returns all metrics
func (rlm *RateLimitMonitor) GetAllMetrics() map[string]*RateLimitMetrics {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	result := make(map[string]*RateLimitMetrics)
	for k, v := range rlm.metrics {
		result[k] = v
	}
	return result
}

// GetAlerts returns current alerts
func (rlm *RateLimitMonitor) GetAlerts() []*RateLimitAlert {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	alerts := make([]*RateLimitAlert, 0, len(rlm.alerts))
	for _, alert := range rlm.alerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// GetAlertHistory returns alert history
func (rlm *RateLimitMonitor) GetAlertHistory() []*RateLimitAlert {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	history := make([]*RateLimitAlert, len(rlm.alertHistory))
	copy(history, rlm.alertHistory)
	return history
}

// AcknowledgeAlert acknowledges an alert
func (rlm *RateLimitMonitor) AcknowledgeAlert(alertID string) error {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	if alert, exists := rlm.alerts[alertID]; exists {
		alert.Acknowledged = true
		return nil
	}
	return fmt.Errorf("alert not found: %s", alertID)
}

// ResolveAlert resolves an alert
func (rlm *RateLimitMonitor) ResolveAlert(alertID string) error {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	if alert, exists := rlm.alerts[alertID]; exists {
		alert.Resolved = true
		now := time.Now()
		alert.ResolvedAt = &now
		return nil
	}
	return fmt.Errorf("alert not found: %s", alertID)
}

// ClearOldMetrics clears old metrics based on retention policy
func (rlm *RateLimitMonitor) ClearOldMetrics() {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -rlm.config.MetricsRetentionDays)
	
	for apiEndpoint, metrics := range rlm.metrics {
		if metrics.LastCheckTime.Before(cutoff) {
			delete(rlm.metrics, apiEndpoint)
		}
	}
}

// ClearOldAlerts clears old alerts based on retention policy
func (rlm *RateLimitMonitor) ClearOldAlerts() {
	rlm.mu.Lock()
	defer rlm.mu.Unlock()

	cutoff := time.Now().AddDate(0, 0, -rlm.config.AlertRetentionDays)
	
	// Clear from current alerts
	for alertID, alert := range rlm.alerts {
		if alert.Timestamp.Before(cutoff) {
			delete(rlm.alerts, alertID)
		}
	}

	// Clear from history
	newHistory := make([]*RateLimitAlert, 0)
	for _, alert := range rlm.alertHistory {
		if alert.Timestamp.After(cutoff) {
			newHistory = append(newHistory, alert)
		}
	}
	rlm.alertHistory = newHistory
}

// Background monitoring loop
func (rlm *RateLimitMonitor) monitoringLoop() {
	ticker := time.NewTicker(rlm.config.MetricsCollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rlm.stopChan:
			return
		case <-ticker.C:
			rlm.performMonitoringTasks()
		}
	}
}

// Perform monitoring tasks
func (rlm *RateLimitMonitor) performMonitoringTasks() {
	// Clear old data
	rlm.ClearOldMetrics()
	rlm.ClearOldAlerts()

	// Check for global alerts
	rlm.checkGlobalAlerts()

	// Log monitoring summary
	rlm.logMonitoringSummary()
}

// Check for alerts based on metrics
func (rlm *RateLimitMonitor) checkForAlerts(apiEndpoint string, metrics *RateLimitMetrics, result *ExternalRateLimitResult) {
	now := time.Now()

	// Check if we're in cooldown period
	if !metrics.LastAlertTime.IsZero() && now.Sub(metrics.LastAlertTime) < rlm.config.AlertCooldown {
		return
	}

	// Check quota exceeded threshold
	if float64(metrics.QuotaExceededCount)/float64(metrics.TotalChecks) > rlm.config.QuotaExceededThreshold {
		rlm.createAlert(apiEndpoint, AlertTypeQuotaExceeded, AlertSeverityWarning,
			fmt.Sprintf("Quota exceeded rate is %.2f%%", 
				float64(metrics.QuotaExceededCount)/float64(metrics.TotalChecks)*100))
	}

	// Check high usage threshold
	if float64(metrics.BlockedRequests)/float64(metrics.TotalChecks) > rlm.config.HighUsageThreshold {
		rlm.createAlert(apiEndpoint, AlertTypeHighUsage, AlertSeverityWarning,
			fmt.Sprintf("High usage detected: %.2f%% requests blocked", 
				float64(metrics.BlockedRequests)/float64(metrics.TotalChecks)*100))
	}

	// Check low success rate
	if metrics.SuccessRate < rlm.config.LowSuccessRateThreshold {
		rlm.createAlert(apiEndpoint, AlertTypeLowSuccessRate, AlertSeverityCritical,
			fmt.Sprintf("Low success rate: %.2f%%", metrics.SuccessRate*100))
	}

	// Check high latency
	if metrics.AverageResponseTime > rlm.config.HighLatencyThreshold {
		rlm.createAlert(apiEndpoint, AlertTypeHighLatency, AlertSeverityWarning,
			fmt.Sprintf("High latency detected: %v average response time", metrics.AverageResponseTime))
	}

	// Check for fallback usage
	if result.FallbackAvailable && !result.Allowed {
		rlm.createAlert(apiEndpoint, AlertTypeFallbackUsed, AlertSeverityInfo,
			"Fallback API used due to rate limiting")
	}

	// Check for cache misses
	if result.CacheHit == false && !result.Allowed {
		rlm.createAlert(apiEndpoint, AlertTypeCacheMiss, AlertSeverityInfo,
			"Cache miss occurred during rate limiting")
	}
}

// Create an alert
func (rlm *RateLimitMonitor) createAlert(apiEndpoint string, alertType AlertType, severity AlertSeverity, message string) {
	alert := &RateLimitAlert{
		ID:          fmt.Sprintf("%s-%s-%d", apiEndpoint, alertType, time.Now().Unix()),
		APIEndpoint: apiEndpoint,
		AlertType:   alertType,
		Severity:    severity,
		Message:     message,
		Timestamp:   time.Now(),
		Metadata:    make(map[string]interface{}),
	}

	rlm.mu.Lock()
	rlm.alerts[alert.ID] = alert
	rlm.alertHistory = append(rlm.alertHistory, alert)
	rlm.mu.Unlock()

	// Update last alert time for the API
	if metrics, exists := rlm.metrics[apiEndpoint]; exists {
		metrics.LastAlertTime = time.Now()
	}

	// Notify alert handlers
	rlm.notifyAlertHandlers(alert)

	rlm.logger.Warn("Rate limit alert created",
		zap.String("alert_id", alert.ID),
		zap.String("api_endpoint", apiEndpoint),
		zap.String("alert_type", string(alertType)),
		zap.String("severity", string(severity)),
		zap.String("message", message))
}

// Notify alert handlers
func (rlm *RateLimitMonitor) notifyAlertHandlers(alert *RateLimitAlert) {
	rlm.mu.RLock()
	handlers := make([]AlertHandler, len(rlm.alertHandlers))
	copy(handlers, rlm.alertHandlers)
	rlm.mu.RUnlock()

	for _, handler := range handlers {
		go func(h AlertHandler, a *RateLimitAlert) {
			if err := h(a); err != nil {
				rlm.logger.Error("Alert handler failed",
					zap.String("alert_id", a.ID),
					zap.Error(err))
			}
		}(handler, alert)
	}
}

// Update time window metrics
func (rlm *RateLimitMonitor) updateTimeWindowMetrics(metrics *RateLimitMetrics, now time.Time, result *ExternalRateLimitResult) {
	// Update minute metrics
	if metrics.MinuteMetrics.WindowStart.IsZero() || now.Sub(metrics.MinuteMetrics.WindowStart) >= time.Minute {
		metrics.MinuteMetrics = &TimeWindowMetrics{
			WindowStart: now.Truncate(time.Minute),
			WindowEnd:   now.Truncate(time.Minute).Add(time.Minute),
		}
	}
	metrics.MinuteMetrics.RequestCount++

	// Update hour metrics
	if metrics.HourMetrics.WindowStart.IsZero() || now.Sub(metrics.HourMetrics.WindowStart) >= time.Hour {
		metrics.HourMetrics = &TimeWindowMetrics{
			WindowStart: now.Truncate(time.Hour),
			WindowEnd:   now.Truncate(time.Hour).Add(time.Hour),
		}
	}
	metrics.HourMetrics.RequestCount++

	// Update day metrics
	if metrics.DayMetrics.WindowStart.IsZero() || now.Sub(metrics.DayMetrics.WindowStart) >= 24*time.Hour {
		metrics.DayMetrics = &TimeWindowMetrics{
			WindowStart: now.Truncate(24 * time.Hour),
			WindowEnd:   now.Truncate(24 * time.Hour).Add(24 * time.Hour),
		}
	}
	metrics.DayMetrics.RequestCount++
}

// Update time window metrics for API calls
func (rlm *RateLimitMonitor) updateTimeWindowMetricsForAPICall(metrics *RateLimitMetrics, now time.Time, success bool, responseTime time.Duration) {
	// Update minute metrics
	if metrics.MinuteMetrics != nil {
		if success {
			metrics.MinuteMetrics.SuccessCount++
		} else {
			metrics.MinuteMetrics.FailureCount++
		}
		
		// Update latency metrics
		if metrics.MinuteMetrics.AverageLatency == 0 {
			metrics.MinuteMetrics.AverageLatency = responseTime
			metrics.MinuteMetrics.PeakLatency = responseTime
			metrics.MinuteMetrics.MinLatency = responseTime
		} else {
			metrics.MinuteMetrics.AverageLatency = (metrics.MinuteMetrics.AverageLatency + responseTime) / 2
			if responseTime > metrics.MinuteMetrics.PeakLatency {
				metrics.MinuteMetrics.PeakLatency = responseTime
			}
			if responseTime < metrics.MinuteMetrics.MinLatency {
				metrics.MinuteMetrics.MinLatency = responseTime
			}
		}
	}

	// Update hour metrics
	if metrics.HourMetrics != nil {
		if success {
			metrics.HourMetrics.SuccessCount++
		} else {
			metrics.HourMetrics.FailureCount++
		}
		
		// Update latency metrics
		if metrics.HourMetrics.AverageLatency == 0 {
			metrics.HourMetrics.AverageLatency = responseTime
			metrics.HourMetrics.PeakLatency = responseTime
			metrics.HourMetrics.MinLatency = responseTime
		} else {
			metrics.HourMetrics.AverageLatency = (metrics.HourMetrics.AverageLatency + responseTime) / 2
			if responseTime > metrics.HourMetrics.PeakLatency {
				metrics.HourMetrics.PeakLatency = responseTime
			}
			if responseTime < metrics.HourMetrics.MinLatency {
				metrics.HourMetrics.MinLatency = responseTime
			}
		}
	}

	// Update day metrics
	if metrics.DayMetrics != nil {
		if success {
			metrics.DayMetrics.SuccessCount++
		} else {
			metrics.DayMetrics.FailureCount++
		}
		
		// Update latency metrics
		if metrics.DayMetrics.AverageLatency == 0 {
			metrics.DayMetrics.AverageLatency = responseTime
			metrics.DayMetrics.PeakLatency = responseTime
			metrics.DayMetrics.MinLatency = responseTime
		} else {
			metrics.DayMetrics.AverageLatency = (metrics.DayMetrics.AverageLatency + responseTime) / 2
			if responseTime > metrics.DayMetrics.PeakLatency {
				metrics.DayMetrics.PeakLatency = responseTime
			}
			if responseTime < metrics.DayMetrics.MinLatency {
				metrics.DayMetrics.MinLatency = responseTime
			}
		}
	}
}

// Check for global alerts
func (rlm *RateLimitMonitor) checkGlobalAlerts() {
	// This would check for system-wide issues
	// Implementation depends on global metrics availability
}

// Log monitoring summary
func (rlm *RateLimitMonitor) logMonitoringSummary() {
	rlm.mu.RLock()
	defer rlm.mu.RUnlock()

	activeAlerts := 0
	for _, alert := range rlm.alerts {
		if !alert.Resolved {
			activeAlerts++
		}
	}

	rlm.logger.Info("Rate limit monitoring summary",
		zap.Int("total_apis", len(rlm.metrics)),
		zap.Int("active_alerts", activeAlerts),
		zap.Int("total_alerts", len(rlm.alertHistory)))
}

// RateLimitFallback provides fallback strategies for rate-limited APIs
type RateLimitFallback struct {
	config *FallbackConfig
	logger *zap.Logger
	mu     sync.RWMutex
	fallbacks map[string][]string
}

// NewRateLimitFallback creates a new rate limit fallback
func NewRateLimitFallback(config *FallbackConfig, logger *zap.Logger) *RateLimitFallback {
	return &RateLimitFallback{
		config:   config,
		logger:   logger,
		fallbacks: make(map[string][]string),
	}
}

// HasFallback checks if a fallback is available for an API
func (rlf *RateLimitFallback) HasFallback(apiEndpoint string) bool {
	rlf.mu.RLock()
	defer rlf.mu.RUnlock()

	fallbacks, exists := rlf.fallbacks[apiEndpoint]
	return exists && len(fallbacks) > 0
}

// RateLimitOptimizer provides optimization and caching for rate limits
type RateLimitOptimizer struct {
	config *OptimizationConfig
	logger *zap.Logger
	mu     sync.RWMutex
	cache  map[string]*CachedResponse
}

// NewRateLimitOptimizer creates a new rate limit optimizer
func NewRateLimitOptimizer(config *OptimizationConfig, logger *zap.Logger) *RateLimitOptimizer {
	return &RateLimitOptimizer{
		config: config,
		logger: logger,
		cache:  make(map[string]*CachedResponse),
	}
}

// HasCachedResponse checks if a cached response is available
func (rlo *RateLimitOptimizer) HasCachedResponse(apiEndpoint string) bool {
	rlo.mu.RLock()
	defer rlo.mu.RUnlock()

	cached, exists := rlo.cache[apiEndpoint]
	if !exists {
		return false
	}

	// Check if cache is still valid
	return time.Since(cached.Timestamp) < cached.TTL
}

// CachedResponse contains a cached API response
type CachedResponse struct {
	Data      interface{}
	Timestamp time.Time
	TTL       time.Duration
}

// DefaultExternalRateLimitConfig returns default configuration
func DefaultExternalRateLimitConfig() *ExternalRateLimitConfig {
	return &ExternalRateLimitConfig{
		GlobalRequestsPerMinute: 100,
		GlobalRequestsPerHour:   5000,
		GlobalRequestsPerDay:    100000,
		DefaultTimeout:          30 * time.Second,
		APIConfigs: map[string]*APIConfig{
			"default": {
				APIEndpoint:       "default",
				RequestsPerMinute: 60,
				RequestsPerHour:   1000,
				RequestsPerDay:    10000,
				Timeout:           30 * time.Second,
				Priority:          1,
				RetryAttempts:     3,
				BackoffStrategy:   "exponential",
				Enabled:           true,
			},
		},
		MonitorConfig: &MonitorConfig{
			Enabled:              true,
			MetricsCollectionInterval: 30 * time.Second,
			AlertThreshold:       0.8,
			AlertCooldown:        5 * time.Minute,
			QuotaExceededThreshold: 0.1,
			HighUsageThreshold:   0.8,
			LowSuccessRateThreshold: 0.9,
			HighLatencyThreshold: 5 * time.Second,
			MetricsRetentionDays: 30,
			AlertRetentionDays:   90,
		},
		FallbackConfig: &FallbackConfig{
			Enabled:          true,
			FallbackAPIs:     []string{},
			CacheFallback:    true,
			RetryWithBackoff: true,
			MaxRetryAttempts: 3,
		},
		OptimizationConfig: &OptimizationConfig{
			Enabled:        true,
			CacheEnabled:   true,
			CacheTTL:       5 * time.Minute,
			RequestBatching: false,
			BatchSize:      10,
			BatchTimeout:   1 * time.Second,
		},
	}
}

// NewExternalAPIRateLimiter creates a new external API rate limiter
func NewExternalAPIRateLimiter(config *ExternalRateLimitConfig, logger *zap.Logger) *ExternalAPIRateLimiter {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = DefaultExternalRateLimitConfig()
	}

	limiter := &ExternalAPIRateLimiter{
		config:    config,
		logger:    logger,
		apiLimits: make(map[string]*ExternalAPILimit),
		globalLimits: &GlobalRateLimit{
			LastMinuteReset: time.Now(),
			LastHourReset:   time.Now(),
			LastDayReset:    time.Now(),
		},
	}

	// Initialize monitoring
	if config.MonitorConfig != nil && config.MonitorConfig.Enabled {
		limiter.monitor = NewRateLimitMonitor(config.MonitorConfig, logger)
	}

	// Initialize fallback strategies
	if config.FallbackConfig != nil && config.FallbackConfig.Enabled {
		limiter.fallback = NewRateLimitFallback(config.FallbackConfig, logger)
	}

	// Initialize optimization
	if config.OptimizationConfig != nil && config.OptimizationConfig.Enabled {
		limiter.optimizer = NewRateLimitOptimizer(config.OptimizationConfig, logger)
	}

	// Initialize API limits from config
	for apiEndpoint, apiConfig := range config.APIConfigs {
		limiter.apiLimits[apiEndpoint] = &ExternalAPILimit{
			Config:         apiConfig,
			LastMinuteReset: time.Now(),
			LastHourReset:   time.Now(),
			LastDayReset:    time.Now(),
		}
	}

	return limiter
}

// CheckRateLimit checks if a request is allowed based on rate limits
func (erl *ExternalAPIRateLimiter) CheckRateLimit(ctx context.Context, apiEndpoint string) (*ExternalRateLimitResult, error) {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	// Check global rate limits first
	if !erl.checkGlobalRateLimit() {
		return &ExternalRateLimitResult{
			Allowed:       false,
			APIEndpoint:   apiEndpoint,
			QuotaExceeded: true,
			RetryAfter:    erl.globalLimits.RetryAfter,
			WaitTime:      erl.globalLimits.RetryAfter.Sub(time.Now()),
		}, nil
	}

	// Get or create API limit
	apiLimit := erl.getOrCreateAPILimit(apiEndpoint)
	if apiLimit == nil {
		return nil, fmt.Errorf("API endpoint %s not configured", apiEndpoint)
	}

	// Reset counters if needed
	erl.resetAPILimitCounters(apiLimit)

	// Check if we're within limits
	result := &ExternalRateLimitResult{
		APIEndpoint:   apiEndpoint,
		Priority:      apiLimit.Config.Priority,
		ResetTime:     apiLimit.LastMinuteReset.Add(time.Minute),
	}

	// Check minute limit
	if apiLimit.CurrentRequestsPerMinute < apiLimit.Config.RequestsPerMinute {
		apiLimit.CurrentRequestsPerMinute++
		erl.globalLimits.CurrentRequestsPerMinute++
		result.Allowed = true
		result.RemainingRequests = apiLimit.Config.RequestsPerMinute - apiLimit.CurrentRequestsPerMinute
	} else {
		result.Allowed = false
		result.RemainingRequests = 0
		result.QuotaExceeded = true
		result.RetryAfter = apiLimit.LastMinuteReset.Add(time.Minute)
		result.WaitTime = result.RetryAfter.Sub(time.Now())
		apiLimit.QuotaExceeded = true
	}

	// Check for fallback availability
	if erl.fallback != nil {
		result.FallbackAvailable = erl.fallback.HasFallback(apiEndpoint)
	}

	// Check for cache availability
	if erl.optimizer != nil && erl.optimizer.config.CacheEnabled {
		result.CacheHit = erl.optimizer.HasCachedResponse(apiEndpoint)
	}

	// Record metrics
	if erl.monitor != nil {
		erl.monitor.RecordRateLimitCheck(apiEndpoint, result)
	}

	return result, nil
}

// WaitForRateLimit waits until rate limit allows the request
func (erl *ExternalAPIRateLimiter) WaitForRateLimit(ctx context.Context, apiEndpoint string) error {
	for {
		result, err := erl.CheckRateLimit(ctx, apiEndpoint)
		if err != nil {
			return err
		}

		if result.Allowed {
			return nil
		}

		// Check if we can use fallback
		if result.FallbackAvailable && erl.fallback != nil {
			erl.logger.Info("Using fallback API", zap.String("api_endpoint", apiEndpoint))
			return nil
		}

		// Check if we can use cache
		if result.CacheHit && erl.optimizer != nil {
			erl.logger.Info("Using cached response", zap.String("api_endpoint", apiEndpoint))
			return nil
		}

		// Wait for the rate limit to reset
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(result.WaitTime):
			continue
		}
	}
}

// RecordAPICall records the result of an API call
func (erl *ExternalAPIRateLimiter) RecordAPICall(apiEndpoint string, success bool, responseTime time.Duration, err error) {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	apiLimit, exists := erl.apiLimits[apiEndpoint]
	if !exists {
		return
	}

	apiLimit.TotalRequests++
	if success {
		apiLimit.SuccessfulRequests++
		apiLimit.LastSuccess = time.Now()
		apiLimit.LastError = nil
	} else {
		apiLimit.FailedRequests++
		apiLimit.LastError = err
	}

	// Update average response time
	if apiLimit.AverageResponseTime == 0 {
		apiLimit.AverageResponseTime = responseTime
	} else {
		apiLimit.AverageResponseTime = (apiLimit.AverageResponseTime + responseTime) / 2
	}

	// Record metrics
	if erl.monitor != nil {
		erl.monitor.RecordAPICall(apiEndpoint, success, responseTime, err)
	}
}

// GetRateLimitStatus gets the current rate limit status for an API
func (erl *ExternalAPIRateLimiter) GetRateLimitStatus(apiEndpoint string) *ExternalAPILimit {
	erl.mu.RLock()
	defer erl.mu.RUnlock()

	if apiLimit, exists := erl.apiLimits[apiEndpoint]; exists {
		return apiLimit
	}
	return nil
}

// GetGlobalRateLimitStatus gets the current global rate limit status
func (erl *ExternalAPIRateLimiter) GetGlobalRateLimitStatus() *GlobalRateLimit {
	erl.mu.RLock()
	defer erl.mu.RUnlock()

	return erl.globalLimits
}

// ResetRateLimit resets the rate limit for an API
func (erl *ExternalAPIRateLimiter) ResetRateLimit(apiEndpoint string) {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	if apiLimit, exists := erl.apiLimits[apiEndpoint]; exists {
		apiLimit.CurrentRequestsPerMinute = 0
		apiLimit.CurrentRequestsPerHour = 0
		apiLimit.CurrentRequestsPerDay = 0
		apiLimit.LastMinuteReset = time.Now()
		apiLimit.LastHourReset = time.Now()
		apiLimit.LastDayReset = time.Now()
		apiLimit.QuotaExceeded = false
	}
}

// ResetGlobalRateLimit resets the global rate limit
func (erl *ExternalAPIRateLimiter) ResetGlobalRateLimit() {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	erl.globalLimits.CurrentRequestsPerMinute = 0
	erl.globalLimits.CurrentRequestsPerHour = 0
	erl.globalLimits.CurrentRequestsPerDay = 0
	erl.globalLimits.LastMinuteReset = time.Now()
	erl.globalLimits.LastHourReset = time.Now()
	erl.globalLimits.LastDayReset = time.Now()
	erl.globalLimits.QuotaExceeded = false
}

// AddAPIConfig adds or updates an API configuration
func (erl *ExternalAPIRateLimiter) AddAPIConfig(apiEndpoint string, config *APIConfig) {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	erl.config.APIConfigs[apiEndpoint] = config
	erl.apiLimits[apiEndpoint] = &ExternalAPILimit{
		Config:         config,
		LastMinuteReset: time.Now(),
		LastHourReset:   time.Now(),
		LastDayReset:    time.Now(),
	}
}

// RemoveAPIConfig removes an API configuration
func (erl *ExternalAPIRateLimiter) RemoveAPIConfig(apiEndpoint string) {
	erl.mu.Lock()
	defer erl.mu.Unlock()

	delete(erl.config.APIConfigs, apiEndpoint)
	delete(erl.apiLimits, apiEndpoint)
}

// GetAPIConfigs returns all API configurations
func (erl *ExternalAPIRateLimiter) GetAPIConfigs() map[string]*APIConfig {
	erl.mu.RLock()
	defer erl.mu.RUnlock()

	result := make(map[string]*APIConfig)
	for k, v := range erl.config.APIConfigs {
		result[k] = v
	}
	return result
}

// Helper methods

func (erl *ExternalAPIRateLimiter) checkGlobalRateLimit() bool {
	now := time.Now()

	// Reset minute counter
	if now.Sub(erl.globalLimits.LastMinuteReset) >= time.Minute {
		erl.globalLimits.CurrentRequestsPerMinute = 0
		erl.globalLimits.LastMinuteReset = now
	}

	// Reset hour counter
	if now.Sub(erl.globalLimits.LastHourReset) >= time.Hour {
		erl.globalLimits.CurrentRequestsPerHour = 0
		erl.globalLimits.LastHourReset = now
	}

	// Reset day counter
	if now.Sub(erl.globalLimits.LastDayReset) >= 24*time.Hour {
		erl.globalLimits.CurrentRequestsPerDay = 0
		erl.globalLimits.LastDayReset = now
	}

	// Check limits
	if erl.globalLimits.CurrentRequestsPerMinute >= erl.config.GlobalRequestsPerMinute ||
		erl.globalLimits.CurrentRequestsPerHour >= erl.config.GlobalRequestsPerHour ||
		erl.globalLimits.CurrentRequestsPerDay >= erl.config.GlobalRequestsPerDay {
		erl.globalLimits.QuotaExceeded = true
		erl.globalLimits.RetryAfter = erl.globalLimits.LastMinuteReset.Add(time.Minute)
		return false
	}

	erl.globalLimits.QuotaExceeded = false
	return true
}

func (erl *ExternalAPIRateLimiter) getOrCreateAPILimit(apiEndpoint string) *ExternalAPILimit {
	if apiLimit, exists := erl.apiLimits[apiEndpoint]; exists {
		return apiLimit
	}

	// Use default config if available
	if defaultConfig, exists := erl.config.APIConfigs["default"]; exists {
		config := &APIConfig{
			APIEndpoint:       apiEndpoint,
			RequestsPerMinute: defaultConfig.RequestsPerMinute,
			RequestsPerHour:   defaultConfig.RequestsPerHour,
			RequestsPerDay:    defaultConfig.RequestsPerDay,
			Timeout:           defaultConfig.Timeout,
			Priority:          defaultConfig.Priority,
			RetryAttempts:     defaultConfig.RetryAttempts,
			BackoffStrategy:   defaultConfig.BackoffStrategy,
			Enabled:           defaultConfig.Enabled,
		}

		apiLimit := &ExternalAPILimit{
			Config:         config,
			LastMinuteReset: time.Now(),
			LastHourReset:   time.Now(),
			LastDayReset:    time.Now(),
		}

		erl.apiLimits[apiEndpoint] = apiLimit
		return apiLimit
	}

	return nil
}

func (erl *ExternalAPIRateLimiter) resetAPILimitCounters(apiLimit *ExternalAPILimit) {
	now := time.Now()

	// Reset minute counter
	if now.Sub(apiLimit.LastMinuteReset) >= time.Minute {
		apiLimit.CurrentRequestsPerMinute = 0
		apiLimit.LastMinuteReset = now
	}

	// Reset hour counter
	if now.Sub(apiLimit.LastHourReset) >= time.Hour {
		apiLimit.CurrentRequestsPerHour = 0
		apiLimit.LastHourReset = now
	}

	// Reset day counter
	if now.Sub(apiLimit.LastDayReset) >= 24*time.Hour {
		apiLimit.CurrentRequestsPerDay = 0
		apiLimit.LastDayReset = now
	}
}
