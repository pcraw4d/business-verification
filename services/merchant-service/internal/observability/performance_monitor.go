package observability

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// PerformanceMonitor handles performance monitoring
type PerformanceMonitor struct {
	logger    *Logger
	metrics   map[string]*PerformanceMetric
	mu        sync.RWMutex
	config    *PerformanceConfig
	exporters []PerformanceExporter
}

// PerformanceMetric represents a performance metric
type PerformanceMetric struct {
	Name        string
	Value       float64
	Unit        string
	Timestamp   time.Time
	Labels      map[string]string
	Percentiles map[string]float64
}

// PerformanceConfig holds configuration for performance monitoring
type PerformanceConfig struct {
	Enabled              bool
	CollectionInterval   time.Duration
	TrackHTTPRequests    bool
	TrackDatabaseQueries bool
	TrackExternalAPIs    bool
	TrackMemoryUsage     bool
	TrackCPUUsage        bool
	TrackGoroutines      bool
	TrackGC              bool
	Percentiles          []float64
}

// PerformanceExporter interface for exporting performance data
type PerformanceExporter interface {
	Export(metrics []*PerformanceMetric) error
	Name() string
}

// PrometheusPerformanceExporter exports performance data to Prometheus
type PrometheusPerformanceExporter struct {
	logger *Logger
}

// NewPrometheusPerformanceExporter creates a new Prometheus performance exporter
func NewPrometheusPerformanceExporter(logger *Logger) *PrometheusPerformanceExporter {
	return &PrometheusPerformanceExporter{
		logger: logger,
	}
}

// Export exports performance data to Prometheus
func (ppe *PrometheusPerformanceExporter) Export(metrics []*PerformanceMetric) error {
	// In a real implementation, this would export to Prometheus
	ppe.logger.Debug("Exporting performance data to Prometheus", map[string]interface{}{
		"metric_count": len(metrics),
	})
	return nil
}

// Name returns the exporter name
func (ppe *PrometheusPerformanceExporter) Name() string {
	return "prometheus"
}

// LogPerformanceExporter exports performance data to logs
type LogPerformanceExporter struct {
	logger *Logger
}

// NewLogPerformanceExporter creates a new log performance exporter
func NewLogPerformanceExporter(logger *Logger) *LogPerformanceExporter {
	return &LogPerformanceExporter{
		logger: logger,
	}
}

// Export exports performance data to logs
func (lpe *LogPerformanceExporter) Export(metrics []*PerformanceMetric) error {
	lpe.logger.Info("Performance metrics", map[string]interface{}{
		"metric_count": len(metrics),
		"metrics":      metrics,
	})
	return nil
}

// Name returns the exporter name
func (lpe *LogPerformanceExporter) Name() string {
	return "log"
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger *Logger, config *PerformanceConfig) *PerformanceMonitor {
	return &PerformanceMonitor{
		logger:    logger,
		metrics:   make(map[string]*PerformanceMetric),
		exporters: make([]PerformanceExporter, 0),
		config:    config,
	}
}

// RecordMetric records a performance metric
func (pm *PerformanceMonitor) RecordMetric(name string, value float64, unit string, labels map[string]string) {
	if !pm.config.Enabled {
		return
	}

	pm.mu.Lock()
	defer pm.mu.Unlock()

	metricKey := pm.generateMetricKey(name, labels)

	metric := &PerformanceMetric{
		Name:        name,
		Value:       value,
		Unit:        unit,
		Timestamp:   time.Now(),
		Labels:      labels,
		Percentiles: make(map[string]float64),
	}

	pm.metrics[metricKey] = metric

	pm.logger.Debug("Performance metric recorded", map[string]interface{}{
		"name":       name,
		"value":      value,
		"unit":       unit,
		"labels":     labels,
		"metric_key": metricKey,
	})
}

// RecordHTTPRequest records HTTP request performance metrics
func (pm *PerformanceMonitor) RecordHTTPRequest(method, endpoint string, statusCode int, duration time.Duration, labels map[string]string) {
	if !pm.config.TrackHTTPRequests {
		return
	}

	requestLabels := map[string]string{
		"method":      method,
		"endpoint":    endpoint,
		"status_code": fmt.Sprintf("%d", statusCode),
	}

	// Merge with additional labels
	for k, v := range labels {
		requestLabels[k] = v
	}

	pm.RecordMetric("http_request_duration_seconds", duration.Seconds(), "seconds", requestLabels)
	pm.RecordMetric("http_requests_total", 1, "count", requestLabels)
}

// RecordDatabaseQuery records database query performance metrics
func (pm *PerformanceMonitor) RecordDatabaseQuery(queryType, table string, duration time.Duration, success bool, labels map[string]string) {
	if !pm.config.TrackDatabaseQueries {
		return
	}

	queryLabels := map[string]string{
		"query_type": queryType,
		"table":      table,
		"success":    fmt.Sprintf("%t", success),
	}

	// Merge with additional labels
	for k, v := range labels {
		queryLabels[k] = v
	}

	pm.RecordMetric("database_query_duration_seconds", duration.Seconds(), "seconds", queryLabels)
	pm.RecordMetric("database_queries_total", 1, "count", queryLabels)
}

// RecordExternalAPI records external API performance metrics
func (pm *PerformanceMonitor) RecordExternalAPI(provider, endpoint string, duration time.Duration, statusCode int, labels map[string]string) {
	if !pm.config.TrackExternalAPIs {
		return
	}

	apiLabels := map[string]string{
		"provider":    provider,
		"endpoint":    endpoint,
		"status_code": fmt.Sprintf("%d", statusCode),
	}

	// Merge with additional labels
	for k, v := range labels {
		apiLabels[k] = v
	}

	pm.RecordMetric("external_api_duration_seconds", duration.Seconds(), "seconds", apiLabels)
	pm.RecordMetric("external_api_calls_total", 1, "count", apiLabels)
}

// RecordBusinessOperation records business operation performance metrics
func (pm *PerformanceMonitor) RecordBusinessOperation(operation string, duration time.Duration, success bool, labels map[string]string) {
	operationLabels := map[string]string{
		"operation": operation,
		"success":   fmt.Sprintf("%t", success),
	}

	// Merge with additional labels
	for k, v := range labels {
		operationLabels[k] = v
	}

	pm.RecordMetric("business_operation_duration_seconds", duration.Seconds(), "seconds", operationLabels)
	pm.RecordMetric("business_operations_total", 1, "count", operationLabels)
}

// RecordMerchantOperation records merchant operation performance metrics
func (pm *PerformanceMonitor) RecordMerchantOperation(operation, merchantID string, duration time.Duration, success bool, labels map[string]string) {
	operationLabels := map[string]string{
		"operation":   operation,
		"merchant_id": merchantID,
		"success":     fmt.Sprintf("%t", success),
	}

	// Merge with additional labels
	for k, v := range labels {
		operationLabels[k] = v
	}

	pm.RecordMetric("merchant_operation_duration_seconds", duration.Seconds(), "seconds", operationLabels)
	pm.RecordMetric("merchant_operations_total", 1, "count", operationLabels)
}

// CollectSystemMetrics collects system-level performance metrics
func (pm *PerformanceMonitor) CollectSystemMetrics() {
	if !pm.config.Enabled {
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Memory metrics
	if pm.config.TrackMemoryUsage {
		pm.RecordMetric("system_memory_alloc_bytes", float64(m.Alloc), "bytes", map[string]string{"type": "alloc"})
		pm.RecordMetric("system_memory_total_alloc_bytes", float64(m.TotalAlloc), "bytes", map[string]string{"type": "total_alloc"})
		pm.RecordMetric("system_memory_sys_bytes", float64(m.Sys), "bytes", map[string]string{"type": "sys"})
		pm.RecordMetric("system_memory_heap_alloc_bytes", float64(m.HeapAlloc), "bytes", map[string]string{"type": "heap_alloc"})
		pm.RecordMetric("system_memory_heap_sys_bytes", float64(m.HeapSys), "bytes", map[string]string{"type": "heap_sys"})
		pm.RecordMetric("system_memory_heap_idle_bytes", float64(m.HeapIdle), "bytes", map[string]string{"type": "heap_idle"})
		pm.RecordMetric("system_memory_heap_inuse_bytes", float64(m.HeapInuse), "bytes", map[string]string{"type": "heap_inuse"})
	}

	// GC metrics
	if pm.config.TrackGC {
		pm.RecordMetric("system_gc_runs_total", float64(m.NumGC), "count", map[string]string{})
		pm.RecordMetric("system_gc_pause_ns", float64(m.PauseNs[(m.NumGC+255)%256]), "nanoseconds", map[string]string{})
	}

	// Goroutine metrics
	if pm.config.TrackGoroutines {
		pm.RecordMetric("system_goroutines", float64(runtime.NumGoroutine()), "count", map[string]string{})
	}

	// CPU metrics (simplified)
	if pm.config.TrackCPUUsage {
		pm.RecordMetric("system_cpu_usage_percent", 0.0, "percent", map[string]string{}) // Would need actual CPU monitoring
	}

	pm.logger.Debug("System performance metrics collected", map[string]interface{}{
		"memory_alloc": m.Alloc,
		"goroutines":   runtime.NumGoroutine(),
		"gc_runs":      m.NumGC,
	})
}

// GetMetrics returns all performance metrics
func (pm *PerformanceMonitor) GetMetrics() map[string]*PerformanceMetric {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := make(map[string]*PerformanceMetric)
	for k, v := range pm.metrics {
		metrics[k] = &PerformanceMetric{
			Name:        v.Name,
			Value:       v.Value,
			Unit:        v.Unit,
			Timestamp:   v.Timestamp,
			Labels:      v.Labels,
			Percentiles: v.Percentiles,
		}
	}
	return metrics
}

// GetMetricsByName returns metrics filtered by name
func (pm *PerformanceMonitor) GetMetricsByName(name string) map[string]*PerformanceMetric {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	filtered := make(map[string]*PerformanceMetric)
	for k, v := range pm.metrics {
		if v.Name == name {
			filtered[k] = &PerformanceMetric{
				Name:        v.Name,
				Value:       v.Value,
				Unit:        v.Unit,
				Timestamp:   v.Timestamp,
				Labels:      v.Labels,
				Percentiles: v.Percentiles,
			}
		}
	}
	return filtered
}

// GetMetricsByTimeRange returns metrics within a time range
func (pm *PerformanceMonitor) GetMetricsByTimeRange(start, end time.Time) map[string]*PerformanceMetric {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	filtered := make(map[string]*PerformanceMetric)
	for k, v := range pm.metrics {
		if v.Timestamp.After(start) && v.Timestamp.Before(end) {
			filtered[k] = &PerformanceMetric{
				Name:        v.Name,
				Value:       v.Value,
				Unit:        v.Unit,
				Timestamp:   v.Timestamp,
				Labels:      v.Labels,
				Percentiles: v.Percentiles,
			}
		}
	}
	return filtered
}

// GetSummary returns performance metrics summary
func (pm *PerformanceMonitor) GetSummary() map[string]interface{} {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	summary := map[string]interface{}{
		"total_metrics":  len(pm.metrics),
		"by_name":        make(map[string]int),
		"by_unit":        make(map[string]int),
		"recent_metrics": make([]*PerformanceMetric, 0),
	}

	now := time.Now()
	recentThreshold := now.Add(-1 * time.Hour)

	for _, metric := range pm.metrics {
		summary["by_name"].(map[string]int)[metric.Name]++
		summary["by_unit"].(map[string]int)[metric.Unit]++

		// Add recent metrics (last hour)
		if metric.Timestamp.After(recentThreshold) {
			summary["recent_metrics"] = append(summary["recent_metrics"].([]*PerformanceMetric), &PerformanceMetric{
				Name:        metric.Name,
				Value:       metric.Value,
				Unit:        metric.Unit,
				Timestamp:   metric.Timestamp,
				Labels:      metric.Labels,
				Percentiles: metric.Percentiles,
			})
		}
	}

	return summary
}

// AddExporter adds a performance exporter
func (pm *PerformanceMonitor) AddExporter(exporter PerformanceExporter) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.exporters = append(pm.exporters, exporter)
	pm.logger.Info("Performance exporter added", map[string]interface{}{
		"exporter": exporter.Name(),
	})
}

// ExportMetrics exports all metrics using registered exporters
func (pm *PerformanceMonitor) ExportMetrics() error {
	pm.mu.RLock()
	metrics := make([]*PerformanceMetric, 0, len(pm.metrics))
	for _, metric := range pm.metrics {
		metrics = append(metrics, &PerformanceMetric{
			Name:        metric.Name,
			Value:       metric.Value,
			Unit:        metric.Unit,
			Timestamp:   metric.Timestamp,
			Labels:      metric.Labels,
			Percentiles: metric.Percentiles,
		})
	}
	pm.mu.RUnlock()

	for _, exporter := range pm.exporters {
		if err := exporter.Export(metrics); err != nil {
			pm.logger.Error("Failed to export performance metrics", map[string]interface{}{
				"exporter": exporter.Name(),
				"error":    err.Error(),
			})
			return fmt.Errorf("failed to export performance metrics with %s: %w", exporter.Name(), err)
		}
	}

	pm.logger.Debug("Performance metrics exported successfully", map[string]interface{}{
		"metric_count": len(metrics),
		"exporters":    len(pm.exporters),
	})
	return nil
}

// StartPeriodicCollection starts periodic collection of performance metrics
func (pm *PerformanceMonitor) StartPeriodicCollection(ctx context.Context) {
	ticker := time.NewTicker(pm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			pm.logger.Info("Periodic performance collection stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			pm.CollectSystemMetrics()
			if err := pm.ExportMetrics(); err != nil {
				pm.logger.Error("Periodic performance export failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}
}

// ClearOldMetrics removes metrics older than the specified duration
func (pm *PerformanceMonitor) ClearOldMetrics(olderThan time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	threshold := now.Add(-olderThan)
	count := 0

	for id, metric := range pm.metrics {
		if metric.Timestamp.Before(threshold) {
			delete(pm.metrics, id)
			count++
		}
	}

	if count > 0 {
		pm.logger.Info("Cleared old performance metrics", map[string]interface{}{
			"count":      count,
			"older_than": olderThan.String(),
		})
	}
}

// generateMetricKey generates a unique key for a performance metric
func (pm *PerformanceMonitor) generateMetricKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += fmt.Sprintf("_%s_%s", k, v)
	}
	return key
}

// CalculatePercentiles calculates percentiles for a set of values
func (pm *PerformanceMonitor) CalculatePercentiles(values []float64, percentiles []float64) map[string]float64 {
	if len(values) == 0 {
		return make(map[string]float64)
	}

	// Sort values
	for i := 0; i < len(values)-1; i++ {
		for j := i + 1; j < len(values); j++ {
			if values[i] > values[j] {
				values[i], values[j] = values[j], values[i]
			}
		}
	}

	result := make(map[string]float64)
	for _, p := range percentiles {
		index := int(float64(len(values)-1) * p / 100.0)
		if index >= len(values) {
			index = len(values) - 1
		}
		result[fmt.Sprintf("p%.0f", p)] = values[index]
	}

	return result
}
