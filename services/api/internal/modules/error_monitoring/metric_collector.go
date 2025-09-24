package error_monitoring

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DefaultMetricCollector provides a default implementation of MetricCollector
type DefaultMetricCollector struct {
	logger  *zap.Logger
	mu      sync.RWMutex
	metrics map[string][]MetricPoint
	config  MetricCollectorConfig
}

// MetricCollectorConfig contains configuration for the metric collector
type MetricCollectorConfig struct {
	MaxMetricsPerProcess int           `json:"max_metrics_per_process"`
	RetentionPeriod      time.Duration `json:"retention_period"`
	AggregationInterval  time.Duration `json:"aggregation_interval"`
	EnablePersistence    bool          `json:"enable_persistence"`
	StoragePath          string        `json:"storage_path"`
}

// NewDefaultMetricCollector creates a new default metric collector
func NewDefaultMetricCollector(config MetricCollectorConfig, logger *zap.Logger) *DefaultMetricCollector {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config.MaxMetricsPerProcess == 0 {
		config.MaxMetricsPerProcess = 10000
	}

	if config.RetentionPeriod == 0 {
		config.RetentionPeriod = 24 * time.Hour
	}

	if config.AggregationInterval == 0 {
		config.AggregationInterval = 1 * time.Minute
	}

	return &DefaultMetricCollector{
		logger:  logger,
		metrics: make(map[string][]MetricPoint),
		config:  config,
	}
}

// RecordErrorRate records an error rate metric
func (dmc *DefaultMetricCollector) RecordErrorRate(processName string, errorRate float64, timestamp time.Time) error {
	dmc.mu.Lock()
	defer dmc.mu.Unlock()

	metricPoint := MetricPoint{
		Timestamp: timestamp,
		Value:     errorRate,
		Labels: map[string]string{
			"process_name": processName,
			"metric_type":  "error_rate",
		},
		Context: map[string]interface{}{
			"process_name": processName,
			"error_rate":   errorRate,
		},
	}

	key := fmt.Sprintf("%s_error_rate", processName)
	dmc.addMetricPoint(key, metricPoint)

	dmc.logger.Debug("Error rate metric recorded",
		zap.String("process", processName),
		zap.Float64("error_rate", errorRate),
		zap.Time("timestamp", timestamp))

	return nil
}

// RecordError records an error occurrence metric
func (dmc *DefaultMetricCollector) RecordError(errorEntry ErrorEntry) error {
	dmc.mu.Lock()
	defer dmc.mu.Unlock()

	metricPoint := MetricPoint{
		Timestamp: errorEntry.Timestamp,
		Value:     1.0, // Error count
		Labels: map[string]string{
			"process_name":   errorEntry.ProcessName,
			"error_type":     errorEntry.ErrorType,
			"error_category": errorEntry.ErrorCategory,
			"severity":       errorEntry.Severity,
			"metric_type":    "error_count",
		},
		Context: map[string]interface{}{
			"process_name":   errorEntry.ProcessName,
			"error_type":     errorEntry.ErrorType,
			"error_category": errorEntry.ErrorCategory,
			"error_message":  errorEntry.ErrorMessage,
			"severity":       errorEntry.Severity,
			"request_id":     errorEntry.RequestID,
			"user_id":        errorEntry.UserID,
			"duration":       errorEntry.Duration,
			"retry_count":    errorEntry.RetryCount,
		},
	}

	key := fmt.Sprintf("%s_errors", errorEntry.ProcessName)
	dmc.addMetricPoint(key, metricPoint)

	// Also record by error type
	typeKey := fmt.Sprintf("%s_errors_%s", errorEntry.ProcessName, errorEntry.ErrorType)
	dmc.addMetricPoint(typeKey, metricPoint)

	// Record by error category
	categoryKey := fmt.Sprintf("%s_errors_%s", errorEntry.ProcessName, errorEntry.ErrorCategory)
	dmc.addMetricPoint(categoryKey, metricPoint)

	dmc.logger.Debug("Error metric recorded",
		zap.String("process", errorEntry.ProcessName),
		zap.String("error_type", errorEntry.ErrorType),
		zap.String("error_category", errorEntry.ErrorCategory),
		zap.Time("timestamp", errorEntry.Timestamp))

	return nil
}

// RecordSuccess records a successful operation metric
func (dmc *DefaultMetricCollector) RecordSuccess(processName string, duration time.Duration, timestamp time.Time) error {
	dmc.mu.Lock()
	defer dmc.mu.Unlock()

	// Record success count
	successMetric := MetricPoint{
		Timestamp: timestamp,
		Value:     1.0,
		Labels: map[string]string{
			"process_name": processName,
			"metric_type":  "success_count",
		},
		Context: map[string]interface{}{
			"process_name": processName,
			"duration":     duration,
		},
	}

	successKey := fmt.Sprintf("%s_success", processName)
	dmc.addMetricPoint(successKey, successMetric)

	// Record response time
	responseTimeMetric := MetricPoint{
		Timestamp: timestamp,
		Value:     float64(duration.Milliseconds()),
		Labels: map[string]string{
			"process_name": processName,
			"metric_type":  "response_time",
		},
		Context: map[string]interface{}{
			"process_name":   processName,
			"duration_ms":    duration.Milliseconds(),
			"duration_human": duration.String(),
		},
	}

	responseTimeKey := fmt.Sprintf("%s_response_time", processName)
	dmc.addMetricPoint(responseTimeKey, responseTimeMetric)

	dmc.logger.Debug("Success metric recorded",
		zap.String("process", processName),
		zap.Duration("duration", duration),
		zap.Time("timestamp", timestamp))

	return nil
}

// GetMetrics retrieves metrics for a process within a time range
func (dmc *DefaultMetricCollector) GetMetrics(processName string, start, end time.Time) ([]MetricPoint, error) {
	dmc.mu.RLock()
	defer dmc.mu.RUnlock()

	var result []MetricPoint

	// Get all metric keys for the process
	for key, metrics := range dmc.metrics {
		if len(processName) == 0 || containsProcess(key, processName) {
			for _, metric := range metrics {
				if metric.Timestamp.After(start) && metric.Timestamp.Before(end) {
					result = append(result, metric)
				}
			}
		}
	}

	dmc.logger.Debug("Metrics retrieved",
		zap.String("process", processName),
		zap.Time("start", start),
		zap.Time("end", end),
		zap.Int("count", len(result)))

	return result, nil
}

// addMetricPoint adds a metric point and manages retention
func (dmc *DefaultMetricCollector) addMetricPoint(key string, point MetricPoint) {
	if _, exists := dmc.metrics[key]; !exists {
		dmc.metrics[key] = make([]MetricPoint, 0)
	}

	dmc.metrics[key] = append(dmc.metrics[key], point)

	// Enforce retention policy
	dmc.enforceRetention(key)
}

// enforceRetention enforces metric retention policies
func (dmc *DefaultMetricCollector) enforceRetention(key string) {
	if metrics, exists := dmc.metrics[key]; exists {
		cutoff := time.Now().Add(-dmc.config.RetentionPeriod)

		// Remove old metrics
		var filteredMetrics []MetricPoint
		for _, metric := range metrics {
			if metric.Timestamp.After(cutoff) {
				filteredMetrics = append(filteredMetrics, metric)
			}
		}

		// Enforce max metrics per process
		if len(filteredMetrics) > dmc.config.MaxMetricsPerProcess {
			// Keep the most recent metrics
			start := len(filteredMetrics) - dmc.config.MaxMetricsPerProcess
			filteredMetrics = filteredMetrics[start:]
		}

		dmc.metrics[key] = filteredMetrics
	}
}

// GetMetricStats returns statistics about stored metrics
func (dmc *DefaultMetricCollector) GetMetricStats() map[string]interface{} {
	dmc.mu.RLock()
	defer dmc.mu.RUnlock()

	stats := make(map[string]interface{})

	totalMetrics := 0
	metricsPerProcess := make(map[string]int)
	metricTypes := make(map[string]int)

	for key, metrics := range dmc.metrics {
		totalMetrics += len(metrics)

		// Extract process name from key
		if processName := extractProcessName(key); processName != "" {
			metricsPerProcess[processName] += len(metrics)
		}

		// Count metric types
		for _, metric := range metrics {
			if metricType, exists := metric.Labels["metric_type"]; exists {
				metricTypes[metricType]++
			}
		}
	}

	stats["total_metrics"] = totalMetrics
	stats["total_metric_keys"] = len(dmc.metrics)
	stats["metrics_per_process"] = metricsPerProcess
	stats["metric_types"] = metricTypes
	stats["max_metrics_per_process"] = dmc.config.MaxMetricsPerProcess
	stats["retention_period"] = dmc.config.RetentionPeriod.String()

	return stats
}

// GetAggregatedMetrics returns aggregated metrics for a time period
func (dmc *DefaultMetricCollector) GetAggregatedMetrics(processName string, start, end time.Time, aggregationType string) (map[string]float64, error) {
	metrics, err := dmc.GetMetrics(processName, start, end)
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64)
	metricGroups := make(map[string][]float64)

	// Group metrics by type
	for _, metric := range metrics {
		if metricType, exists := metric.Labels["metric_type"]; exists {
			metricGroups[metricType] = append(metricGroups[metricType], metric.Value)
		}
	}

	// Aggregate by type
	for metricType, values := range metricGroups {
		switch aggregationType {
		case "avg", "average":
			result[metricType] = calculateAverage(values)
		case "sum", "total":
			result[metricType] = calculateSum(values)
		case "min", "minimum":
			result[metricType] = calculateMin(values)
		case "max", "maximum":
			result[metricType] = calculateMax(values)
		case "count":
			result[metricType] = float64(len(values))
		case "p95":
			result[metricType] = calculatePercentile(values, 0.95)
		case "p99":
			result[metricType] = calculatePercentile(values, 0.99)
		default:
			result[metricType] = calculateAverage(values)
		}
	}

	return result, nil
}

// CleanupOldMetrics removes old metrics beyond retention period
func (dmc *DefaultMetricCollector) CleanupOldMetrics() {
	dmc.mu.Lock()
	defer dmc.mu.Unlock()

	for key := range dmc.metrics {
		dmc.enforceRetention(key)
	}

	dmc.logger.Info("Old metrics cleaned up")
}

// Helper functions

func containsProcess(key, processName string) bool {
	return len(processName) == 0 || key[:len(processName)] == processName
}

func extractProcessName(key string) string {
	// Simple extraction - in a real implementation, this would be more sophisticated
	parts := make([]string, 0)
	current := ""
	for _, char := range key {
		if char == '_' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}

	if len(parts) > 0 {
		return parts[0]
	}
	return ""
}

func calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func calculateSum(values []float64) float64 {
	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum
}

func calculateMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	min := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
	}
	return min
}

func calculateMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	max := values[0]
	for _, value := range values[1:] {
		if value > max {
			max = value
		}
	}
	return max
}

func calculatePercentile(values []float64, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Simple percentile calculation - in production, use a proper sorting algorithm
	// This is a simplified implementation
	if percentile >= 1.0 {
		return calculateMax(values)
	}
	if percentile <= 0.0 {
		return calculateMin(values)
	}

	// For simplicity, just return average for percentile calculations
	// In a real implementation, sort the values and calculate proper percentile
	return calculateAverage(values)
}
