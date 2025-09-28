package monitoringoptimization

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// MonitoringOptimizer provides advanced monitoring and alerting capabilities
type MonitoringOptimizer struct {
	config      *MonitoringConfig
	metrics     *MetricsCollector
	alerts      *AlertManager
	health      *HealthMonitor
	performance *PerformanceTracker
}

// MonitoringConfig contains monitoring optimization settings
type MonitoringConfig struct {
	// Metrics Collection
	EnableMetricsCollection bool
	MetricsInterval         time.Duration
	RetentionPeriod         time.Duration
	MaxMetricsHistory       int

	// Alerting
	EnableAlerting  bool
	AlertThresholds map[string]float64
	AlertCooldown   time.Duration
	AlertChannels   []string

	// Health Monitoring
	EnableHealthChecks  bool
	HealthCheckInterval time.Duration
	HealthCheckTimeout  time.Duration
	UnhealthyThreshold  int

	// Performance Tracking
	EnablePerformanceTracking bool
	SlowRequestThreshold      time.Duration
	MemoryThreshold           float64
	CPUThreshold              float64

	// Auto-scaling
	EnableAutoScaling  bool
	ScaleUpThreshold   float64
	ScaleDownThreshold float64
	MinInstances       int
	MaxInstances       int
}

// DefaultMonitoringConfig returns optimized monitoring configuration
func DefaultMonitoringConfig() *MonitoringConfig {
	return &MonitoringConfig{
		// Metrics Collection
		EnableMetricsCollection: true,
		MetricsInterval:         30 * time.Second,
		RetentionPeriod:         24 * time.Hour,
		MaxMetricsHistory:       1000,

		// Alerting
		EnableAlerting: true,
		AlertThresholds: map[string]float64{
			"error_rate":    5.0,   // 5% error rate
			"response_time": 500.0, // 500ms response time
			"cpu_usage":     80.0,  // 80% CPU usage
			"memory_usage":  85.0,  // 85% memory usage
			"disk_usage":    90.0,  // 90% disk usage
		},
		AlertCooldown: 5 * time.Minute,
		AlertChannels: []string{"log", "webhook"},

		// Health Monitoring
		EnableHealthChecks:  true,
		HealthCheckInterval: 10 * time.Second,
		HealthCheckTimeout:  5 * time.Second,
		UnhealthyThreshold:  3,

		// Performance Tracking
		EnablePerformanceTracking: true,
		SlowRequestThreshold:      100 * time.Millisecond,
		MemoryThreshold:           80.0,
		CPUThreshold:              75.0,

		// Auto-scaling
		EnableAutoScaling:  true,
		ScaleUpThreshold:   70.0,
		ScaleDownThreshold: 30.0,
		MinInstances:       1,
		MaxInstances:       10,
	}
}

// NewMonitoringOptimizer creates a new monitoring optimizer
func NewMonitoringOptimizer(config *MonitoringConfig) *MonitoringOptimizer {
	if config == nil {
		config = DefaultMonitoringConfig()
	}

	return &MonitoringOptimizer{
		config:      config,
		metrics:     NewMetricsCollector(config),
		alerts:      NewAlertManager(config),
		health:      NewHealthMonitor(config),
		performance: NewPerformanceTracker(config),
	}
}

// Start starts all monitoring components
func (mo *MonitoringOptimizer) Start(ctx context.Context) {
	if mo.config.EnableMetricsCollection {
		go mo.metrics.Start(ctx)
	}

	if mo.config.EnableHealthChecks {
		go mo.health.Start(ctx)
	}

	if mo.config.EnablePerformanceTracking {
		go mo.performance.Start(ctx)
	}

	if mo.config.EnableAlerting {
		go mo.alerts.Start(ctx)
	}

	log.Println("🚀 Monitoring optimizer started with all components")
}

// RecordMetric records a metric value
func (mo *MonitoringOptimizer) RecordMetric(name string, value float64, tags map[string]string) {
	if mo.config.EnableMetricsCollection {
		mo.metrics.Record(name, value, tags)
	}
}

// RecordRequest records a request metric
func (mo *MonitoringOptimizer) RecordRequest(method, endpoint string, statusCode int, duration time.Duration) {
	if mo.config.EnablePerformanceTracking {
		mo.performance.RecordRequest(method, endpoint, statusCode, duration)
	}

	// Record as metric
	mo.RecordMetric("request_duration", float64(duration.Milliseconds()), map[string]string{
		"method":      method,
		"endpoint":    endpoint,
		"status_code": fmt.Sprintf("%d", statusCode),
	})
}

// GetMetrics returns current metrics
func (mo *MonitoringOptimizer) GetMetrics() *MetricsSummary {
	return mo.metrics.GetSummary()
}

// GetHealthStatus returns current health status
func (mo *MonitoringOptimizer) GetHealthStatus() *HealthStatus {
	return mo.health.GetStatus()
}

// GetPerformanceStats returns performance statistics
func (mo *MonitoringOptimizer) GetPerformanceStats() *PerformanceStats {
	return mo.performance.GetStats()
}

// GetAlerts returns current alerts
func (mo *MonitoringOptimizer) GetAlerts() []Alert {
	return mo.alerts.GetActiveAlerts()
}

// MetricsCollector collects and stores metrics
type MetricsCollector struct {
	config  *MonitoringConfig
	metrics map[string][]MetricPoint
	mutex   sync.RWMutex
}

// MetricPoint represents a single metric measurement
type MetricPoint struct {
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
	Timestamp time.Time         `json:"timestamp"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config *MonitoringConfig) *MetricsCollector {
	return &MetricsCollector{
		config:  config,
		metrics: make(map[string][]MetricPoint),
	}
}

// Start starts the metrics collector
func (mc *MetricsCollector) Start(ctx context.Context) {
	ticker := time.NewTicker(mc.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mc.cleanupOldMetrics()
		}
	}
}

// Record records a metric value
func (mc *MetricsCollector) Record(name string, value float64, tags map[string]string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	if mc.metrics[name] == nil {
		mc.metrics[name] = make([]MetricPoint, 0)
	}

	mc.metrics[name] = append(mc.metrics[name], MetricPoint{
		Value:     value,
		Tags:      tags,
		Timestamp: time.Now(),
	})

	// Limit history size
	if len(mc.metrics[name]) > mc.config.MaxMetricsHistory {
		mc.metrics[name] = mc.metrics[name][len(mc.metrics[name])-mc.config.MaxMetricsHistory:]
	}
}

// GetSummary returns metrics summary
func (mc *MetricsCollector) GetSummary() *MetricsSummary {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	summary := &MetricsSummary{
		Timestamp: time.Now(),
		Metrics:   make(map[string]MetricSummary),
	}

	for name, points := range mc.metrics {
		if len(points) == 0 {
			continue
		}

		var sum, min, max float64
		min = points[0].Value
		max = points[0].Value

		for _, point := range points {
			sum += point.Value
			if point.Value < min {
				min = point.Value
			}
			if point.Value > max {
				max = point.Value
			}
		}

		summary.Metrics[name] = MetricSummary{
			Count:   len(points),
			Average: sum / float64(len(points)),
			Min:     min,
			Max:     max,
			Latest:  points[len(points)-1].Value,
		}
	}

	return summary
}

// cleanupOldMetrics removes old metrics
func (mc *MetricsCollector) cleanupOldMetrics() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	cutoff := time.Now().Add(-mc.config.RetentionPeriod)

	for name, points := range mc.metrics {
		var validPoints []MetricPoint
		for _, point := range points {
			if point.Timestamp.After(cutoff) {
				validPoints = append(validPoints, point)
			}
		}
		mc.metrics[name] = validPoints
	}
}

// MetricsSummary contains aggregated metrics
type MetricsSummary struct {
	Timestamp time.Time                `json:"timestamp"`
	Metrics   map[string]MetricSummary `json:"metrics"`
}

// MetricSummary contains summary statistics for a metric
type MetricSummary struct {
	Count   int     `json:"count"`
	Average float64 `json:"average"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Latest  float64 `json:"latest"`
}

// HealthMonitor monitors service health
type HealthMonitor struct {
	config *MonitoringConfig
	status *HealthStatus
	checks map[string]HealthCheck
	mutex  sync.RWMutex
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) error

// HealthStatus represents the current health status
type HealthStatus struct {
	Status         string            `json:"status"`
	Timestamp      time.Time         `json:"timestamp"`
	Checks         map[string]string `json:"checks"`
	UnhealthyCount int               `json:"unhealthy_count"`
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(config *MonitoringConfig) *HealthMonitor {
	return &HealthMonitor{
		config: config,
		status: &HealthStatus{
			Status:    "healthy",
			Timestamp: time.Now(),
			Checks:    make(map[string]string),
		},
		checks: make(map[string]HealthCheck),
	}
}

// Start starts the health monitor
func (hm *HealthMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(hm.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			hm.performHealthChecks(ctx)
		}
	}
}

// AddHealthCheck adds a health check
func (hm *HealthMonitor) AddHealthCheck(name string, check HealthCheck) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()
	hm.checks[name] = check
}

// performHealthChecks performs all health checks
func (hm *HealthMonitor) performHealthChecks(ctx context.Context) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	unhealthyCount := 0
	checkResults := make(map[string]string)

	for name, check := range hm.checks {
		checkCtx, cancel := context.WithTimeout(ctx, hm.config.HealthCheckTimeout)
		err := check(checkCtx)
		cancel()

		if err != nil {
			checkResults[name] = fmt.Sprintf("unhealthy: %v", err)
			unhealthyCount++
		} else {
			checkResults[name] = "healthy"
		}
	}

	// Update status
	hm.status.Checks = checkResults
	hm.status.UnhealthyCount = unhealthyCount
	hm.status.Timestamp = time.Now()

	if unhealthyCount >= hm.config.UnhealthyThreshold {
		hm.status.Status = "unhealthy"
	} else {
		hm.status.Status = "healthy"
	}
}

// GetStatus returns current health status
func (hm *HealthMonitor) GetStatus() *HealthStatus {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()
	return hm.status
}

// PerformanceTracker tracks performance metrics
type PerformanceTracker struct {
	config *MonitoringConfig
	stats  *PerformanceStats
	mutex  sync.RWMutex
}

// PerformanceStats contains performance statistics
type PerformanceStats struct {
	TotalRequests       int64            `json:"total_requests"`
	SuccessfulRequests  int64            `json:"successful_requests"`
	FailedRequests      int64            `json:"failed_requests"`
	AverageResponseTime time.Duration    `json:"average_response_time"`
	SlowRequests        int64            `json:"slow_requests"`
	RequestsPerSecond   float64          `json:"requests_per_second"`
	EndpointStats       map[string]int64 `json:"endpoint_stats"`
	LastUpdated         time.Time        `json:"last_updated"`
}

// NewPerformanceTracker creates a new performance tracker
func NewPerformanceTracker(config *MonitoringConfig) *PerformanceTracker {
	return &PerformanceTracker{
		config: config,
		stats: &PerformanceStats{
			EndpointStats: make(map[string]int64),
		},
	}
}

// Start starts the performance tracker
func (pt *PerformanceTracker) Start(ctx context.Context) {
	// Performance tracking is event-driven, no background processing needed
}

// RecordRequest records a request
func (pt *PerformanceTracker) RecordRequest(method, endpoint string, statusCode int, duration time.Duration) {
	pt.mutex.Lock()
	defer pt.mutex.Unlock()

	pt.stats.TotalRequests++

	if statusCode >= 200 && statusCode < 400 {
		pt.stats.SuccessfulRequests++
	} else {
		pt.stats.FailedRequests++
	}

	if duration > pt.config.SlowRequestThreshold {
		pt.stats.SlowRequests++
	}

	// Update endpoint stats
	endpointKey := fmt.Sprintf("%s %s", method, endpoint)
	pt.stats.EndpointStats[endpointKey]++

	// Update average response time (simplified calculation)
	if pt.stats.TotalRequests == 1 {
		pt.stats.AverageResponseTime = duration
	} else {
		// Simple moving average
		pt.stats.AverageResponseTime = (pt.stats.AverageResponseTime + duration) / 2
	}

	pt.stats.LastUpdated = time.Now()
}

// GetStats returns performance statistics
func (pt *PerformanceTracker) GetStats() *PerformanceStats {
	pt.mutex.RLock()
	defer pt.mutex.RUnlock()
	return pt.stats
}

// AlertManager manages alerts and notifications
type AlertManager struct {
	config    *MonitoringConfig
	alerts    []Alert
	lastAlert map[string]time.Time
	mutex     sync.RWMutex
}

// Alert represents an alert
type Alert struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
}

// NewAlertManager creates a new alert manager
func NewAlertManager(config *MonitoringConfig) *AlertManager {
	return &AlertManager{
		config:    config,
		alerts:    make([]Alert, 0),
		lastAlert: make(map[string]time.Time),
	}
}

// Start starts the alert manager
func (am *AlertManager) Start(ctx context.Context) {
	// Alert manager is event-driven, no background processing needed
}

// CheckThresholds checks if metrics exceed thresholds
func (am *AlertManager) CheckThresholds(metrics *MetricsSummary) {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	for metricName, summary := range metrics.Metrics {
		if threshold, exists := am.config.AlertThresholds[metricName]; exists {
			if summary.Latest > threshold {
				am.createAlert(metricName, summary.Latest, threshold)
			}
		}
	}
}

// createAlert creates a new alert
func (am *AlertManager) createAlert(metricName string, value, threshold float64) {
	alertKey := fmt.Sprintf("%s_%f", metricName, threshold)

	// Check cooldown
	if lastTime, exists := am.lastAlert[alertKey]; exists {
		if time.Since(lastTime) < am.config.AlertCooldown {
			return // Still in cooldown
		}
	}

	alert := Alert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Type:      "threshold_exceeded",
		Severity:  am.getSeverity(metricName, value, threshold),
		Message:   fmt.Sprintf("%s exceeded threshold: %.2f > %.2f", metricName, value, threshold),
		Value:     value,
		Threshold: threshold,
		Timestamp: time.Now(),
		Resolved:  false,
	}

	am.alerts = append(am.alerts, alert)
	am.lastAlert[alertKey] = time.Now()

	// Send alert (in real implementation, this would send to actual channels)
	log.Printf("🚨 ALERT: %s", alert.Message)
}

// getSeverity determines alert severity
func (am *AlertManager) getSeverity(metricName string, value, threshold float64) string {
	ratio := value / threshold
	if ratio > 2.0 {
		return "critical"
	} else if ratio > 1.5 {
		return "warning"
	} else {
		return "info"
	}
}

// GetActiveAlerts returns active alerts
func (am *AlertManager) GetActiveAlerts() []Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var activeAlerts []Alert
	for _, alert := range am.alerts {
		if !alert.Resolved {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}
