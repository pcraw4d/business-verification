package observability

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MetricsCollector handles metrics collection and aggregation
type MetricsCollector struct {
	logger    *Logger
	metrics   map[string]*Metric
	mu        sync.RWMutex
	exporters []MetricsExporter
}

// Metric represents a collected metric
type Metric struct {
	Name        string
	Value       float64
	Type        MetricType
	Labels      map[string]string
	Timestamp   time.Time
	Description string
}

// MetricType represents the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// MetricsExporter interface for exporting metrics
type MetricsExporter interface {
	Export(metrics []*Metric) error
	Name() string
}

// PrometheusExporter exports metrics to Prometheus
type PrometheusExporter struct {
	logger *Logger
}

// NewPrometheusExporter creates a new Prometheus exporter
func NewPrometheusExporter(logger *Logger) *PrometheusExporter {
	return &PrometheusExporter{
		logger: logger,
	}
}

// Export exports metrics to Prometheus
func (pe *PrometheusExporter) Export(metrics []*Metric) error {
	// In a real implementation, this would export to Prometheus
	pe.logger.Debug("Exporting metrics to Prometheus", map[string]interface{}{
		"metric_count": len(metrics),
	})
	return nil
}

// Name returns the exporter name
func (pe *PrometheusExporter) Name() string {
	return "prometheus"
}

// ConsoleExporter exports metrics to console (for debugging)
type ConsoleExporter struct {
	logger *Logger
}

// NewConsoleExporter creates a new console exporter
func NewConsoleExporter(logger *Logger) *ConsoleExporter {
	return &ConsoleExporter{
		logger: logger,
	}
}

// Export exports metrics to console
func (ce *ConsoleExporter) Export(metrics []*Metric) error {
	ce.logger.Info("Exporting metrics to console", map[string]interface{}{
		"metric_count": len(metrics),
	})

	for _, metric := range metrics {
		ce.logger.Info("Metric", map[string]interface{}{
			"name":        metric.Name,
			"value":       metric.Value,
			"type":        metric.Type,
			"labels":      metric.Labels,
			"timestamp":   metric.Timestamp,
			"description": metric.Description,
		})
	}
	return nil
}

// Name returns the exporter name
func (ce *ConsoleExporter) Name() string {
	return "console"
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *Logger) *MetricsCollector {
	return &MetricsCollector{
		logger:    logger,
		metrics:   make(map[string]*Metric),
		exporters: make([]MetricsExporter, 0),
	}
}

// RecordMetric records a metric
func (mc *MetricsCollector) RecordMetric(name string, value float64, metricType MetricType, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metricKey := mc.generateMetricKey(name, labels)

	metric := &Metric{
		Name:        name,
		Value:       value,
		Type:        metricType,
		Labels:      labels,
		Timestamp:   time.Now(),
		Description: mc.getMetricDescription(name),
	}

	mc.metrics[metricKey] = metric

	mc.logger.Debug("Metric recorded", map[string]interface{}{
		"name":       name,
		"value":      value,
		"type":       metricType,
		"labels":     labels,
		"metric_key": metricKey,
	})
}

// IncrementCounter increments a counter metric
func (mc *MetricsCollector) IncrementCounter(name string, labels map[string]string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	metricKey := mc.generateMetricKey(name, labels)

	if existing, exists := mc.metrics[metricKey]; exists && existing.Type == MetricTypeCounter {
		existing.Value++
		existing.Timestamp = time.Now()
	} else {
		mc.metrics[metricKey] = &Metric{
			Name:        name,
			Value:       1,
			Type:        MetricTypeCounter,
			Labels:      labels,
			Timestamp:   time.Now(),
			Description: mc.getMetricDescription(name),
		}
	}
}

// SetGauge sets a gauge metric
func (mc *MetricsCollector) SetGauge(name string, value float64, labels map[string]string) {
	mc.RecordMetric(name, value, MetricTypeGauge, labels)
}

// RecordHistogram records a histogram metric
func (mc *MetricsCollector) RecordHistogram(name string, value float64, labels map[string]string) {
	mc.RecordMetric(name, value, MetricTypeHistogram, labels)
}

// GetMetrics returns all metrics
func (mc *MetricsCollector) GetMetrics() map[string]*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := make(map[string]*Metric)
	for k, v := range mc.metrics {
		metrics[k] = &Metric{
			Name:        v.Name,
			Value:       v.Value,
			Type:        v.Type,
			Labels:      v.Labels,
			Timestamp:   v.Timestamp,
			Description: v.Description,
		}
	}
	return metrics
}

// GetMetric returns a specific metric
func (mc *MetricsCollector) GetMetric(name string, labels map[string]string) (*Metric, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metricKey := mc.generateMetricKey(name, labels)
	metric, exists := mc.metrics[metricKey]
	if !exists {
		return nil, false
	}

	// Return a copy
	return &Metric{
		Name:        metric.Name,
		Value:       metric.Value,
		Type:        metric.Type,
		Labels:      metric.Labels,
		Timestamp:   metric.Timestamp,
		Description: metric.Description,
	}, true
}

// GetMetricsByType returns metrics filtered by type
func (mc *MetricsCollector) GetMetricsByType(metricType MetricType) map[string]*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	filtered := make(map[string]*Metric)
	for k, v := range mc.metrics {
		if v.Type == metricType {
			filtered[k] = &Metric{
				Name:        v.Name,
				Value:       v.Value,
				Type:        v.Type,
				Labels:      v.Labels,
				Timestamp:   v.Timestamp,
				Description: v.Description,
			}
		}
	}
	return filtered
}

// GetMetricsByName returns metrics filtered by name
func (mc *MetricsCollector) GetMetricsByName(name string) map[string]*Metric {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	filtered := make(map[string]*Metric)
	for k, v := range mc.metrics {
		if v.Name == name {
			filtered[k] = &Metric{
				Name:        v.Name,
				Value:       v.Value,
				Type:        v.Type,
				Labels:      v.Labels,
				Timestamp:   v.Timestamp,
				Description: v.Description,
			}
		}
	}
	return filtered
}

// ClearMetrics clears all metrics
func (mc *MetricsCollector) ClearMetrics() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.metrics = make(map[string]*Metric)
	mc.logger.Info("All metrics cleared", map[string]interface{}{})
}

// ClearMetricsByType clears metrics of a specific type
func (mc *MetricsCollector) ClearMetricsByType(metricType MetricType) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	count := 0
	for k, v := range mc.metrics {
		if v.Type == metricType {
			delete(mc.metrics, k)
			count++
		}
	}

	mc.logger.Info("Metrics cleared by type", map[string]interface{}{
		"type":  metricType,
		"count": count,
	})
}

// AddExporter adds a metrics exporter
func (mc *MetricsCollector) AddExporter(exporter MetricsExporter) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.exporters = append(mc.exporters, exporter)
	mc.logger.Info("Metrics exporter added", map[string]interface{}{
		"exporter": exporter.Name(),
	})
}

// ExportMetrics exports all metrics using registered exporters
func (mc *MetricsCollector) ExportMetrics() error {
	mc.mu.RLock()
	metrics := make([]*Metric, 0, len(mc.metrics))
	for _, metric := range mc.metrics {
		metrics = append(metrics, &Metric{
			Name:        metric.Name,
			Value:       metric.Value,
			Type:        metric.Type,
			Labels:      metric.Labels,
			Timestamp:   metric.Timestamp,
			Description: metric.Description,
		})
	}
	mc.mu.RUnlock()

	for _, exporter := range mc.exporters {
		if err := exporter.Export(metrics); err != nil {
			mc.logger.Error("Failed to export metrics", map[string]interface{}{
				"exporter": exporter.Name(),
				"error":    err.Error(),
			})
			return fmt.Errorf("failed to export metrics with %s: %w", exporter.Name(), err)
		}
	}

	mc.logger.Debug("Metrics exported successfully", map[string]interface{}{
		"metric_count": len(metrics),
		"exporters":    len(mc.exporters),
	})
	return nil
}

// GetMetricsSummary returns a summary of metrics
func (mc *MetricsCollector) GetMetricsSummary() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	summary := map[string]interface{}{
		"total_metrics": len(mc.metrics),
		"by_type":       make(map[MetricType]int),
		"by_name":       make(map[string]int),
	}

	// Count by type
	for _, metric := range mc.metrics {
		summary["by_type"].(map[MetricType]int)[metric.Type]++
		summary["by_name"].(map[string]int)[metric.Name]++
	}

	return summary
}

// generateMetricKey generates a unique key for a metric
func (mc *MetricsCollector) generateMetricKey(name string, labels map[string]string) string {
	key := name
	for k, v := range labels {
		key += fmt.Sprintf("_%s_%s", k, v)
	}
	return key
}

// getMetricDescription returns a description for a metric
func (mc *MetricsCollector) getMetricDescription(name string) string {
	descriptions := map[string]string{
		"http_requests_total":            "Total number of HTTP requests",
		"http_request_duration":          "HTTP request duration in seconds",
		"http_requests_in_flight":        "Number of HTTP requests currently in flight",
		"database_queries_total":         "Total number of database queries",
		"database_query_duration":        "Database query duration in seconds",
		"external_api_calls_total":       "Total number of external API calls",
		"external_api_duration":          "External API call duration in seconds",
		"system_memory_alloc_bytes":      "System memory allocated in bytes",
		"system_goroutines":              "Number of goroutines",
		"system_gc_runs_total":           "Total number of garbage collection runs",
		"business_classifications_total": "Total number of business classifications",
		"risk_assessments_total":         "Total number of risk assessments",
		"compliance_checks_total":        "Total number of compliance checks",
		"user_sessions_total":            "Total number of user sessions",
		"merchant_operations_total":      "Total number of merchant operations",
	}

	if desc, exists := descriptions[name]; exists {
		return desc
	}
	return fmt.Sprintf("Metric: %s", name)
}

// StartPeriodicExport starts periodic export of metrics
func (mc *MetricsCollector) StartPeriodicExport(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			mc.logger.Info("Periodic metrics export stopped", map[string]interface{}{})
			return
		case <-ticker.C:
			if err := mc.ExportMetrics(); err != nil {
				mc.logger.Error("Periodic metrics export failed", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}
}
