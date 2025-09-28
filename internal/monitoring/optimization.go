package monitoring

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PerformanceMonitor provides comprehensive performance monitoring
type PerformanceMonitor struct {
	metrics   *PerformanceMetrics
	alerting  *AlertManager
	profiling *Profiler
	analytics *AnalyticsEngine
	config    *MonitoringConfig
	mu        sync.RWMutex
}

// MonitoringConfig contains monitoring configuration
type MonitoringConfig struct {
	EnableMetrics   bool            `yaml:"enable_metrics"`
	EnableAlerting  bool            `yaml:"enable_alerting"`
	EnableProfiling bool            `yaml:"enable_profiling"`
	EnableAnalytics bool            `yaml:"enable_analytics"`
	MetricsInterval time.Duration   `yaml:"metrics_interval"`
	AlertThresholds AlertThresholds `yaml:"alert_thresholds"`
	RetentionPeriod time.Duration   `yaml:"retention_period"`
}

// AlertThresholds defines alerting thresholds
type AlertThresholds struct {
	ResponseTime    time.Duration `yaml:"response_time"`
	ErrorRate       float64       `yaml:"error_rate"`
	CacheHitRate    float64       `yaml:"cache_hit_rate"`
	MemoryUsage     float64       `yaml:"memory_usage"`
	CPUUsage        float64       `yaml:"cpu_usage"`
	DatabaseLatency time.Duration `yaml:"database_latency"`
}

// PerformanceMetrics contains Prometheus metrics
type PerformanceMetrics struct {
	ResponseTime      prometheus.HistogramVec
	RequestCount      prometheus.CounterVec
	ErrorRate         prometheus.CounterVec
	CacheHitRate      prometheus.GaugeVec
	DatabaseLatency   prometheus.HistogramVec
	MemoryUsage       prometheus.GaugeVec
	CPUUsage          prometheus.GaugeVec
	ActiveConnections prometheus.GaugeVec
	QueueSize         prometheus.GaugeVec
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(config *MonitoringConfig) *PerformanceMonitor {
	pm := &PerformanceMonitor{
		config: config,
	}

	if config.EnableMetrics {
		pm.metrics = pm.createMetrics()
	}

	if config.EnableAlerting {
		pm.alerting = NewAlertManager(config.AlertThresholds)
	}

	if config.EnableProfiling {
		pm.profiling = NewProfiler()
	}

	if config.EnableAnalytics {
		pm.analytics = NewAnalyticsEngine(config.RetentionPeriod)
	}

	return pm
}

// createMetrics creates Prometheus metrics
func (pm *PerformanceMonitor) createMetrics() *PerformanceMetrics {
	return &PerformanceMetrics{
		ResponseTime: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kyb_response_time_seconds",
				Help:    "Response time in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"endpoint", "method", "status"},
		),
		RequestCount: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyb_requests_total",
				Help: "Total number of requests",
			},
			[]string{"endpoint", "method", "status"},
		),
		ErrorRate: *promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyb_errors_total",
				Help: "Total number of errors",
			},
			[]string{"endpoint", "method", "error_type"},
		),
		CacheHitRate: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kyb_cache_hit_rate",
				Help: "Cache hit rate percentage",
			},
			[]string{"cache_type"},
		),
		DatabaseLatency: *promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kyb_database_latency_seconds",
				Help:    "Database query latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"operation", "table"},
		),
		MemoryUsage: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kyb_memory_usage_bytes",
				Help: "Memory usage in bytes",
			},
			[]string{"type"},
		),
		CPUUsage: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kyb_cpu_usage_percent",
				Help: "CPU usage percentage",
			},
			[]string{"type"},
		),
		ActiveConnections: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kyb_active_connections",
				Help: "Number of active connections",
			},
			[]string{"type"},
		),
		QueueSize: *promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kyb_queue_size",
				Help: "Queue size",
			},
			[]string{"queue_type"},
		),
	}
}

// TrackRequest tracks a request's performance
func (pm *PerformanceMonitor) TrackRequest(endpoint string, method string, duration time.Duration, statusCode int) {
	if pm.metrics == nil {
		return
	}

	labels := prometheus.Labels{
		"endpoint": endpoint,
		"method":   method,
		"status":   fmt.Sprintf("%d", statusCode),
	}

	pm.metrics.ResponseTime.With(labels).Observe(duration.Seconds())
	pm.metrics.RequestCount.With(labels).Inc()

	if statusCode >= 400 {
		errorLabels := prometheus.Labels{
			"endpoint":   endpoint,
			"method":     method,
			"error_type": fmt.Sprintf("http_%d", statusCode),
		}
		pm.metrics.ErrorRate.With(errorLabels).Inc()
	}

	// Check for alerts
	if pm.alerting != nil {
		pm.alerting.CheckResponseTime(endpoint, method, duration)
		pm.alerting.CheckErrorRate(endpoint, method, statusCode >= 400)
	}
}

// TrackCacheHit tracks cache performance
func (pm *PerformanceMonitor) TrackCacheHit(cacheType string, hit bool) {
	if pm.metrics == nil {
		return
	}

	labels := prometheus.Labels{
		"cache_type": cacheType,
	}

	// Update cache hit rate
	pm.metrics.CacheHitRate.With(labels).Set(func() float64 {
		if hit {
			return 1.0
		}
		return 0.0
	}())

	// Check for alerts
	if pm.alerting != nil {
		pm.alerting.CheckCacheHitRate(cacheType, hit)
	}
}

// TrackDatabaseQuery tracks database query performance
func (pm *PerformanceMonitor) TrackDatabaseQuery(operation string, table string, duration time.Duration) {
	if pm.metrics == nil {
		return
	}

	labels := prometheus.Labels{
		"operation": operation,
		"table":     table,
	}

	pm.metrics.DatabaseLatency.With(labels).Observe(duration.Seconds())

	// Check for alerts
	if pm.alerting != nil {
		pm.alerting.CheckDatabaseLatency(operation, table, duration)
	}
}

// TrackSystemMetrics tracks system resource usage
func (pm *PerformanceMonitor) TrackSystemMetrics(memoryUsage, cpuUsage float64) {
	if pm.metrics == nil {
		return
	}

	pm.metrics.MemoryUsage.WithLabelValues("heap").Set(memoryUsage)
	pm.metrics.CPUUsage.WithLabelValues("process").Set(cpuUsage)

	// Check for alerts
	if pm.alerting != nil {
		pm.alerting.CheckMemoryUsage(memoryUsage)
		pm.alerting.CheckCPUUsage(cpuUsage)
	}
}

// TrackConnections tracks connection metrics
func (pm *PerformanceMonitor) TrackConnections(connectionType string, count int) {
	if pm.metrics == nil {
		return
	}

	pm.metrics.ActiveConnections.WithLabelValues(connectionType).Set(float64(count))
}

// TrackQueueSize tracks queue metrics
func (pm *PerformanceMonitor) TrackQueueSize(queueType string, size int) {
	if pm.metrics == nil {
		return
	}

	pm.metrics.QueueSize.WithLabelValues(queueType).Set(float64(size))
}

// AlertManager manages performance alerts
type AlertManager struct {
	thresholds AlertThresholds
	alerts     map[string]*Alert
	mu         sync.RWMutex
}

// Alert represents a performance alert
type Alert struct {
	Name          string
	Severity      string
	Condition     string
	Action        string
	LastTriggered time.Time
	Count         int
	Enabled       bool
}

// NewAlertManager creates a new alert manager
func NewAlertManager(thresholds AlertThresholds) *AlertManager {
	am := &AlertManager{
		thresholds: thresholds,
		alerts:     make(map[string]*Alert),
	}

	am.setupDefaultAlerts()
	return am
}

// setupDefaultAlerts sets up default performance alerts
func (am *AlertManager) setupDefaultAlerts() {
	alerts := []Alert{
		{
			Name:      "high_response_time",
			Severity:  "warning",
			Condition: "response_time > 500ms",
			Action:    "scale_up",
			Enabled:   true,
		},
		{
			Name:      "low_cache_hit_rate",
			Severity:  "warning",
			Condition: "cache_hit_rate < 80%",
			Action:    "investigate_cache",
			Enabled:   true,
		},
		{
			Name:      "high_error_rate",
			Severity:  "critical",
			Condition: "error_rate > 5%",
			Action:    "immediate_attention",
			Enabled:   true,
		},
		{
			Name:      "high_memory_usage",
			Severity:  "warning",
			Condition: "memory_usage > 80%",
			Action:    "scale_up",
			Enabled:   true,
		},
		{
			Name:      "high_cpu_usage",
			Severity:  "warning",
			Condition: "cpu_usage > 80%",
			Action:    "scale_up",
			Enabled:   true,
		},
		{
			Name:      "high_database_latency",
			Severity:  "warning",
			Condition: "database_latency > 100ms",
			Action:    "investigate_database",
			Enabled:   true,
		},
	}

	for _, alert := range alerts {
		am.alerts[alert.Name] = &alert
	}
}

// CheckResponseTime checks if response time exceeds threshold
func (am *AlertManager) CheckResponseTime(endpoint, method string, duration time.Duration) {
	if duration > am.thresholds.ResponseTime {
		am.triggerAlert("high_response_time", fmt.Sprintf("Response time %v exceeds threshold for %s %s", duration, method, endpoint))
	}
}

// CheckErrorRate checks if error rate exceeds threshold
func (am *AlertManager) CheckErrorRate(endpoint, method string, isError bool) {
	// This is a simplified check - in production, you'd track error rates over time
	if isError {
		am.triggerAlert("high_error_rate", fmt.Sprintf("Error detected for %s %s", method, endpoint))
	}
}

// CheckCacheHitRate checks if cache hit rate is below threshold
func (am *AlertManager) CheckCacheHitRate(cacheType string, hit bool) {
	// This is a simplified check - in production, you'd track hit rates over time
	if !hit {
		am.triggerAlert("low_cache_hit_rate", fmt.Sprintf("Cache miss for %s", cacheType))
	}
}

// CheckMemoryUsage checks if memory usage exceeds threshold
func (am *AlertManager) CheckMemoryUsage(usage float64) {
	if usage > am.thresholds.MemoryUsage {
		am.triggerAlert("high_memory_usage", fmt.Sprintf("Memory usage %.2f%% exceeds threshold", usage))
	}
}

// CheckCPUUsage checks if CPU usage exceeds threshold
func (am *AlertManager) CheckCPUUsage(usage float64) {
	if usage > am.thresholds.CPUUsage {
		am.triggerAlert("high_cpu_usage", fmt.Sprintf("CPU usage %.2f%% exceeds threshold", usage))
	}
}

// CheckDatabaseLatency checks if database latency exceeds threshold
func (am *AlertManager) CheckDatabaseLatency(operation, table string, duration time.Duration) {
	if duration > am.thresholds.DatabaseLatency {
		am.triggerAlert("high_database_latency", fmt.Sprintf("Database latency %v exceeds threshold for %s on %s", duration, operation, table))
	}
}

// triggerAlert triggers an alert
func (am *AlertManager) triggerAlert(alertName, message string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	alert, exists := am.alerts[alertName]
	if !exists || !alert.Enabled {
		return
	}

	alert.Count++
	alert.LastTriggered = time.Now()

	log.Printf("ALERT [%s] %s: %s", alert.Severity, alertName, message)

	// In production, you would send this to your alerting system
	// (e.g., PagerDuty, Slack, email, etc.)
	am.sendAlert(alert, message)
}

// sendAlert sends an alert to the configured alerting system
func (am *AlertManager) sendAlert(alert *Alert, message string) {
	// This is where you would integrate with your alerting system
	// For now, we'll just log it
	log.Printf("Sending alert: %s - %s", alert.Name, message)
}

// GetAlerts returns all alerts
func (am *AlertManager) GetAlerts() map[string]*Alert {
	am.mu.RLock()
	defer am.mu.RUnlock()

	alerts := make(map[string]*Alert)
	for name, alert := range am.alerts {
		alerts[name] = &Alert{
			Name:          alert.Name,
			Severity:      alert.Severity,
			Condition:     alert.Condition,
			Action:        alert.Action,
			LastTriggered: alert.LastTriggered,
			Count:         alert.Count,
			Enabled:       alert.Enabled,
		}
	}

	return alerts
}

// Profiler provides performance profiling capabilities
type Profiler struct {
	enabled bool
	mu      sync.RWMutex
}

// NewProfiler creates a new profiler
func NewProfiler() *Profiler {
	return &Profiler{
		enabled: true,
	}
}

// StartProfiling starts performance profiling
func (p *Profiler) StartProfiling() error {
	if !p.enabled {
		return fmt.Errorf("profiling is disabled")
	}

	log.Println("Starting performance profiling...")

	// In production, you would start CPU, memory, and goroutine profiling
	// This is a simplified implementation

	return nil
}

// StopProfiling stops performance profiling
func (p *Profiler) StopProfiling() error {
	if !p.enabled {
		return fmt.Errorf("profiling is disabled")
	}

	log.Println("Stopping performance profiling...")

	// In production, you would stop profiling and save the results

	return nil
}

// AnalyticsEngine provides performance analytics
type AnalyticsEngine struct {
	retentionPeriod time.Duration
	data            map[string][]DataPoint
	mu              sync.RWMutex
}

// DataPoint represents a performance data point
type DataPoint struct {
	Timestamp time.Time
	Value     float64
	Labels    map[string]string
}

// NewAnalyticsEngine creates a new analytics engine
func NewAnalyticsEngine(retentionPeriod time.Duration) *AnalyticsEngine {
	ae := &AnalyticsEngine{
		retentionPeriod: retentionPeriod,
		data:            make(map[string][]DataPoint),
	}

	// Start cleanup routine
	go ae.startCleanup()

	return ae
}

// RecordDataPoint records a performance data point
func (ae *AnalyticsEngine) RecordDataPoint(metric string, value float64, labels map[string]string) {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	dataPoint := DataPoint{
		Timestamp: time.Now(),
		Value:     value,
		Labels:    labels,
	}

	ae.data[metric] = append(ae.data[metric], dataPoint)
}

// GetAnalytics returns analytics data for a metric
func (ae *AnalyticsEngine) GetAnalytics(metric string, startTime, endTime time.Time) ([]DataPoint, error) {
	ae.mu.RLock()
	defer ae.mu.RUnlock()

	data, exists := ae.data[metric]
	if !exists {
		return nil, fmt.Errorf("metric %s not found", metric)
	}

	var filteredData []DataPoint
	for _, point := range data {
		if point.Timestamp.After(startTime) && point.Timestamp.Before(endTime) {
			filteredData = append(filteredData, point)
		}
	}

	return filteredData, nil
}

// startCleanup starts the data cleanup routine
func (ae *AnalyticsEngine) startCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		ae.cleanupOldData()
	}
}

// cleanupOldData removes old data points
func (ae *AnalyticsEngine) cleanupOldData() {
	ae.mu.Lock()
	defer ae.mu.Unlock()

	cutoff := time.Now().Add(-ae.retentionPeriod)

	for metric, data := range ae.data {
		var filteredData []DataPoint
		for _, point := range data {
			if point.Timestamp.After(cutoff) {
				filteredData = append(filteredData, point)
			}
		}
		ae.data[metric] = filteredData
	}
}

// GetPerformanceReport generates a comprehensive performance report
func (pm *PerformanceMonitor) GetPerformanceReport() (*PerformanceReport, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	report := &PerformanceReport{
		Timestamp: time.Now(),
		Metrics:   make(map[string]interface{}),
		Alerts:    make(map[string]*Alert),
	}

	// Collect metrics
	if pm.metrics != nil {
		report.Metrics["response_time"] = pm.metrics.ResponseTime
		report.Metrics["request_count"] = pm.metrics.RequestCount
		report.Metrics["error_rate"] = pm.metrics.ErrorRate
		report.Metrics["cache_hit_rate"] = pm.metrics.CacheHitRate
		report.Metrics["database_latency"] = pm.metrics.DatabaseLatency
		report.Metrics["memory_usage"] = pm.metrics.MemoryUsage
		report.Metrics["cpu_usage"] = pm.metrics.CPUUsage
	}

	// Collect alerts
	if pm.alerting != nil {
		report.Alerts = pm.alerting.GetAlerts()
	}

	// Generate recommendations
	report.Recommendations = pm.generateRecommendations()

	return report, nil
}

// PerformanceReport contains performance analysis
type PerformanceReport struct {
	Timestamp       time.Time              `json:"timestamp"`
	Metrics         map[string]interface{} `json:"metrics"`
	Alerts          map[string]*Alert      `json:"alerts"`
	Recommendations []string               `json:"recommendations"`
}

// generateRecommendations generates performance recommendations
func (pm *PerformanceMonitor) generateRecommendations() []string {
	var recommendations []string

	// This would analyze the collected metrics and generate recommendations
	// For now, we'll provide some general recommendations

	recommendations = append(recommendations, "Monitor response times and scale up if consistently high")
	recommendations = append(recommendations, "Optimize cache hit rates for better performance")
	recommendations = append(recommendations, "Monitor database query performance and add indexes if needed")
	recommendations = append(recommendations, "Set up automated scaling based on resource usage")

	return recommendations
}
