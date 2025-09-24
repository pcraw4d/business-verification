// Package observability provides enhanced performance monitoring capabilities for the KYB platform.
// This module extends the unified monitoring system with comprehensive performance tracking.
package observability

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

// EnhancedPerformanceMonitor extends the UnifiedPerformanceMonitor with additional performance tracking.
type EnhancedPerformanceMonitor struct {
	*UnifiedPerformanceMonitor
	config *EnhancedMonitoringConfig
	logger *zap.Logger

	// Internal state for enhanced monitoring
	mu                sync.RWMutex
	lastMetricsUpdate time.Time
	metricsCache      map[string]float64
	alertThresholds   map[string]EnhancedAlertThreshold
}

// EnhancedMonitoringConfig contains configuration for enhanced performance monitoring.
type EnhancedMonitoringConfig struct {
	// Collection settings
	CollectionInterval time.Duration `yaml:"collection_interval"`
	CacheTimeout       time.Duration `yaml:"cache_timeout"`
	BatchSize          int           `yaml:"batch_size"`

	// Database monitoring
	DatabaseMonitoring struct {
		Enabled           bool          `yaml:"enabled"`
		QueryTimeout      time.Duration `yaml:"query_timeout"`
		ConnectionTimeout time.Duration `yaml:"connection_timeout"`
		MaxConnections    int           `yaml:"max_connections"`
	} `yaml:"database_monitoring"`

	// Application monitoring
	ApplicationMonitoring struct {
		Enabled            bool          `yaml:"enabled"`
		MemoryThreshold    float64       `yaml:"memory_threshold"`
		CPUThreshold       float64       `yaml:"cpu_threshold"`
		GoroutineThreshold int           `yaml:"goroutine_threshold"`
		GCThreshold        time.Duration `yaml:"gc_threshold"`
	} `yaml:"application_monitoring"`

	// Business metrics monitoring
	BusinessMetricsMonitoring struct {
		Enabled                   bool    `yaml:"enabled"`
		ClassificationAccuracyMin float64 `yaml:"classification_accuracy_min"`
		RiskDetectionLatencyMax   float64 `yaml:"risk_detection_latency_max"`
		APIResponseTimeMax        float64 `yaml:"api_response_time_max"`
		ErrorRateMax              float64 `yaml:"error_rate_max"`
	} `yaml:"business_metrics_monitoring"`

	// Alerting configuration
	Alerting struct {
		Enabled              bool                              `yaml:"enabled"`
		CooldownPeriod       time.Duration                     `yaml:"cooldown_period"`
		NotificationChannels []string                          `yaml:"notification_channels"`
		Thresholds           map[string]EnhancedAlertThreshold `yaml:"thresholds"`
	} `yaml:"alerting"`
}

// EnhancedAlertThreshold defines enhanced alerting thresholds for performance metrics.
type EnhancedAlertThreshold struct {
	MetricName  string        `yaml:"metric_name"`
	Condition   string        `yaml:"condition"` // 'gt', 'lt', 'eq', 'ne'
	Value       float64       `yaml:"value"`
	Severity    string        `yaml:"severity"` // 'info', 'warning', 'critical'
	Duration    time.Duration `yaml:"duration"`
	Cooldown    time.Duration `yaml:"cooldown"`
	Enabled     bool          `yaml:"enabled"`
	Description string        `yaml:"description"`
}

// EnhancedPerformanceMetrics represents collected enhanced performance metrics.
type EnhancedPerformanceMetrics struct {
	Timestamp time.Time `json:"timestamp"`

	// Database metrics
	Database struct {
		ActiveConnections   int           `json:"active_connections"`
		IdleConnections     int           `json:"idle_connections"`
		MaxConnections      int           `json:"max_connections"`
		QueryDuration       time.Duration `json:"query_duration"`
		QueryCount          int64         `json:"query_count"`
		ErrorCount          int64         `json:"error_count"`
		ConnectionPoolUsage float64       `json:"connection_pool_usage"`
	} `json:"database"`

	// Application metrics
	Application struct {
		MemoryUsage     float64       `json:"memory_usage"`
		MemoryAllocated uint64        `json:"memory_allocated"`
		MemorySystem    uint64        `json:"memory_system"`
		CPUUsage        float64       `json:"cpu_usage"`
		GoroutineCount  int           `json:"goroutine_count"`
		GCDuration      time.Duration `json:"gc_duration"`
		NumGC           int64         `json:"num_gc"`
	} `json:"application"`

	// Business metrics
	Business struct {
		ClassificationAccuracy float64 `json:"classification_accuracy"`
		RiskDetectionLatency   float64 `json:"risk_detection_latency"`
		APIResponseTime        float64 `json:"api_response_time"`
		ErrorRate              float64 `json:"error_rate"`
		RequestRate            float64 `json:"request_rate"`
		ActiveUsers            int     `json:"active_users"`
	} `json:"business"`
}

// NewEnhancedPerformanceMonitor creates a new enhanced performance monitor instance.
func NewEnhancedPerformanceMonitor(
	baseMonitor *UnifiedPerformanceMonitor,
	config *EnhancedMonitoringConfig,
	logger *zap.Logger,
) *EnhancedPerformanceMonitor {
	epm := &EnhancedPerformanceMonitor{
		UnifiedPerformanceMonitor: baseMonitor,
		config:                    config,
		logger:                    logger,
		metricsCache:              make(map[string]float64),
		alertThresholds:           make(map[string]EnhancedAlertThreshold),
	}

	// Load alert thresholds
	epm.loadAlertThresholds()

	return epm
}

// loadAlertThresholds loads alert thresholds from configuration.
func (epm *EnhancedPerformanceMonitor) loadAlertThresholds() {
	for name, threshold := range epm.config.Alerting.Thresholds {
		epm.alertThresholds[name] = threshold
	}
}

// StartEnhancedMonitoring begins the enhanced performance monitoring process.
func (epm *EnhancedPerformanceMonitor) StartEnhancedMonitoring(ctx context.Context) error {
	epm.logger.Info("Starting enhanced performance monitor")

	// Start enhanced metrics collection
	go epm.collectEnhancedMetricsLoop(ctx)

	// Start enhanced alert checking
	if epm.config.Alerting.Enabled {
		go epm.enhancedAlertCheckingLoop(ctx)
	}

	return nil
}

// StopEnhancedMonitoring stops the enhanced performance monitoring process.
func (epm *EnhancedPerformanceMonitor) StopEnhancedMonitoring() error {
	epm.logger.Info("Stopping enhanced performance monitor")
	return nil
}

// collectEnhancedMetricsLoop runs the enhanced metrics collection loop.
func (epm *EnhancedPerformanceMonitor) collectEnhancedMetricsLoop(ctx context.Context) {
	ticker := time.NewTicker(epm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := epm.collectEnhancedMetrics(ctx); err != nil {
				epm.logger.Error("Failed to collect enhanced metrics", zap.Error(err))
			}
		}
	}
}

// collectEnhancedMetrics collects all enhanced performance metrics.
func (epm *EnhancedPerformanceMonitor) collectEnhancedMetrics(ctx context.Context) error {
	epm.mu.Lock()
	defer epm.mu.Unlock()

	metrics := &EnhancedPerformanceMetrics{
		Timestamp: time.Now(),
	}

	// Collect database metrics
	if epm.config.DatabaseMonitoring.Enabled {
		if err := epm.collectEnhancedDatabaseMetrics(ctx, metrics); err != nil {
			epm.logger.Error("Failed to collect enhanced database metrics", zap.Error(err))
		}
	}

	// Collect application metrics
	if epm.config.ApplicationMonitoring.Enabled {
		epm.collectEnhancedApplicationMetrics(metrics)
	}

	// Collect business metrics
	if epm.config.BusinessMetricsMonitoring.Enabled {
		if err := epm.collectEnhancedBusinessMetrics(ctx, metrics); err != nil {
			epm.logger.Error("Failed to collect enhanced business metrics", zap.Error(err))
		}
	}

	// Cache metrics
	epm.cacheEnhancedMetrics(metrics)

	epm.lastMetricsUpdate = time.Now()

	return nil
}

// collectEnhancedDatabaseMetrics collects enhanced database performance metrics.
func (epm *EnhancedPerformanceMonitor) collectEnhancedDatabaseMetrics(ctx context.Context, metrics *EnhancedPerformanceMetrics) error {
	// Get database connection stats
	stats := epm.db.Stats()
	metrics.Database.ActiveConnections = stats.OpenConnections
	metrics.Database.IdleConnections = stats.Idle
	metrics.Database.MaxConnections = stats.MaxOpenConnections
	// Note: QueriesTotal and ErrorsTotal are not available in sql.DBStats
	// These would need to be tracked separately in a real implementation
	metrics.Database.QueryCount = 0
	metrics.Database.ErrorCount = 0

	// Calculate connection pool usage
	if stats.MaxOpenConnections > 0 {
		metrics.Database.ConnectionPoolUsage = float64(stats.OpenConnections) / float64(stats.MaxOpenConnections)
	}

	// Test query performance
	start := time.Now()
	if err := epm.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}
	metrics.Database.QueryDuration = time.Since(start)

	return nil
}

// collectEnhancedApplicationMetrics collects enhanced application performance metrics.
func (epm *EnhancedPerformanceMonitor) collectEnhancedApplicationMetrics(metrics *EnhancedPerformanceMetrics) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Memory metrics
	metrics.Application.MemoryUsage = float64(m.Alloc)
	metrics.Application.MemoryAllocated = m.Alloc
	metrics.Application.MemorySystem = m.Sys

	// CPU and goroutine metrics
	metrics.Application.GoroutineCount = runtime.NumGoroutine()

	// GC metrics
	metrics.Application.GCDuration = time.Duration(m.PauseTotalNs)
	metrics.Application.NumGC = int64(m.NumGC)

	// Calculate CPU usage (simplified)
	metrics.Application.CPUUsage = epm.calculateEnhancedCPUUsage()
}

// collectEnhancedBusinessMetrics collects enhanced business-specific performance metrics.
func (epm *EnhancedPerformanceMonitor) collectEnhancedBusinessMetrics(ctx context.Context, metrics *EnhancedPerformanceMetrics) error {
	// Get classification accuracy from database
	if err := epm.getEnhancedClassificationAccuracy(ctx, &metrics.Business.ClassificationAccuracy); err != nil {
		epm.logger.Warn("Failed to get enhanced classification accuracy", zap.Error(err))
	}

	// Get risk detection latency from database
	if err := epm.getEnhancedRiskDetectionLatency(ctx, &metrics.Business.RiskDetectionLatency); err != nil {
		epm.logger.Warn("Failed to get enhanced risk detection latency", zap.Error(err))
	}

	// Get API response time from database
	if err := epm.getEnhancedAPIResponseTime(ctx, &metrics.Business.APIResponseTime); err != nil {
		epm.logger.Warn("Failed to get enhanced API response time", zap.Error(err))
	}

	// Get error rate from database
	if err := epm.getEnhancedErrorRate(ctx, &metrics.Business.ErrorRate); err != nil {
		epm.logger.Warn("Failed to get enhanced error rate", zap.Error(err))
	}

	// Get request rate from database
	if err := epm.getEnhancedRequestRate(ctx, &metrics.Business.RequestRate); err != nil {
		epm.logger.Warn("Failed to get enhanced request rate", zap.Error(err))
	}

	// Get active users from database
	if err := epm.getEnhancedActiveUsers(ctx, &metrics.Business.ActiveUsers); err != nil {
		epm.logger.Warn("Failed to get enhanced active users", zap.Error(err))
	}

	return nil
}

// getEnhancedClassificationAccuracy retrieves enhanced classification accuracy from database.
func (epm *EnhancedPerformanceMonitor) getEnhancedClassificationAccuracy(ctx context.Context, accuracy *float64) error {
	query := `
		SELECT AVG(confidence_score) * 100
		FROM classification_results 
		WHERE created_at >= NOW() - INTERVAL '1 hour'
		AND confidence_score IS NOT NULL
	`
	return epm.db.QueryRowContext(ctx, query).Scan(accuracy)
}

// getEnhancedRiskDetectionLatency retrieves enhanced risk detection latency from database.
func (epm *EnhancedPerformanceMonitor) getEnhancedRiskDetectionLatency(ctx context.Context, latency *float64) error {
	query := `
		SELECT AVG(EXTRACT(EPOCH FROM (updated_at - created_at)))
		FROM business_risk_assessments 
		WHERE created_at >= NOW() - INTERVAL '1 hour'
		AND updated_at IS NOT NULL
	`
	return epm.db.QueryRowContext(ctx, query).Scan(latency)
}

// getEnhancedAPIResponseTime retrieves enhanced API response time from database.
func (epm *EnhancedPerformanceMonitor) getEnhancedAPIResponseTime(ctx context.Context, responseTime *float64) error {
	query := `
		SELECT AVG(response_time_ms) / 1000.0
		FROM api_metrics 
		WHERE created_at >= NOW() - INTERVAL '1 hour'
		AND response_time_ms IS NOT NULL
	`
	return epm.db.QueryRowContext(ctx, query).Scan(responseTime)
}

// getEnhancedErrorRate retrieves enhanced error rate from database.
func (epm *EnhancedPerformanceMonitor) getEnhancedErrorRate(ctx context.Context, errorRate *float64) error {
	query := `
		SELECT 
			(COUNT(CASE WHEN status_code >= 400 THEN 1 END) * 100.0 / COUNT(*))
		FROM api_metrics 
		WHERE created_at >= NOW() - INTERVAL '1 hour'
	`
	return epm.db.QueryRowContext(ctx, query).Scan(errorRate)
}

// getEnhancedRequestRate retrieves enhanced request rate from database.
func (epm *EnhancedPerformanceMonitor) getEnhancedRequestRate(ctx context.Context, requestRate *float64) error {
	query := `
		SELECT COUNT(*) / 3600.0
		FROM api_metrics 
		WHERE created_at >= NOW() - INTERVAL '1 hour'
	`
	return epm.db.QueryRowContext(ctx, query).Scan(requestRate)
}

// getEnhancedActiveUsers retrieves enhanced active users from database.
func (epm *EnhancedPerformanceMonitor) getEnhancedActiveUsers(ctx context.Context, activeUsers *int) error {
	query := `
		SELECT COUNT(DISTINCT user_id)
		FROM api_metrics 
		WHERE created_at >= NOW() - INTERVAL '1 hour'
		AND user_id IS NOT NULL
	`
	return epm.db.QueryRowContext(ctx, query).Scan(activeUsers)
}

// calculateEnhancedCPUUsage calculates enhanced CPU usage percentage.
func (epm *EnhancedPerformanceMonitor) calculateEnhancedCPUUsage() float64 {
	// Simplified CPU usage calculation
	// In a real implementation, you would use more sophisticated methods
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Use GC frequency as a proxy for CPU usage
	if m.NumGC > 0 {
		return float64(m.NumGC) * 0.1 // Simplified calculation
	}
	return 0.0
}

// cacheEnhancedMetrics caches enhanced metrics for quick access.
func (epm *EnhancedPerformanceMonitor) cacheEnhancedMetrics(metrics *EnhancedPerformanceMetrics) {
	epm.metricsCache["database_connections"] = float64(metrics.Database.ActiveConnections)
	epm.metricsCache["memory_usage"] = metrics.Application.MemoryUsage
	epm.metricsCache["cpu_usage"] = metrics.Application.CPUUsage
	epm.metricsCache["goroutine_count"] = float64(metrics.Application.GoroutineCount)
	epm.metricsCache["classification_accuracy"] = metrics.Business.ClassificationAccuracy
	epm.metricsCache["risk_detection_latency"] = metrics.Business.RiskDetectionLatency
	epm.metricsCache["api_response_time"] = metrics.Business.APIResponseTime
	epm.metricsCache["error_rate"] = metrics.Business.ErrorRate
	epm.metricsCache["request_rate"] = metrics.Business.RequestRate
	epm.metricsCache["active_users"] = float64(metrics.Business.ActiveUsers)
}

// enhancedAlertCheckingLoop runs the enhanced alert checking loop.
func (epm *EnhancedPerformanceMonitor) enhancedAlertCheckingLoop(ctx context.Context) {
	ticker := time.NewTicker(epm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			epm.checkEnhancedAlerts(ctx)
		}
	}
}

// checkEnhancedAlerts checks for enhanced alert conditions.
func (epm *EnhancedPerformanceMonitor) checkEnhancedAlerts(ctx context.Context) {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	for alertName, threshold := range epm.alertThresholds {
		if !threshold.Enabled {
			continue
		}

		value, exists := epm.metricsCache[threshold.MetricName]
		if !exists {
			continue
		}

		if epm.evaluateEnhancedThreshold(value, threshold) {
			epm.triggerEnhancedAlert(ctx, alertName, threshold, value)
		}
	}
}

// evaluateEnhancedThreshold evaluates if a metric value meets the enhanced alert threshold.
func (epm *EnhancedPerformanceMonitor) evaluateEnhancedThreshold(value float64, threshold EnhancedAlertThreshold) bool {
	switch threshold.Condition {
	case "gt":
		return value > threshold.Value
	case "lt":
		return value < threshold.Value
	case "eq":
		return value == threshold.Value
	case "ne":
		return value != threshold.Value
	default:
		return false
	}
}

// triggerEnhancedAlert triggers an enhanced alert.
func (epm *EnhancedPerformanceMonitor) triggerEnhancedAlert(ctx context.Context, alertName string, threshold EnhancedAlertThreshold, value float64) {
	epm.logger.Warn("Enhanced performance alert triggered",
		zap.String("alert_name", alertName),
		zap.String("metric_name", threshold.MetricName),
		zap.Float64("current_value", value),
		zap.Float64("threshold", threshold.Value),
		zap.String("condition", threshold.Condition),
		zap.String("severity", threshold.Severity),
		zap.String("description", threshold.Description),
	)

	// TODO: Implement enhanced alert notification system
	// This would integrate with the existing alerting infrastructure
}

// GetEnhancedMetrics returns the current cached enhanced metrics.
func (epm *EnhancedPerformanceMonitor) GetEnhancedMetrics() map[string]float64 {
	epm.mu.RLock()
	defer epm.mu.RUnlock()

	// Return a copy of the metrics cache
	metrics := make(map[string]float64)
	for k, v := range epm.metricsCache {
		metrics[k] = v
	}
	return metrics
}

// GetEnhancedLastUpdateTime returns the time of the last enhanced metrics update.
func (epm *EnhancedPerformanceMonitor) GetEnhancedLastUpdateTime() time.Time {
	epm.mu.RLock()
	defer epm.mu.RUnlock()
	return epm.lastMetricsUpdate
}
