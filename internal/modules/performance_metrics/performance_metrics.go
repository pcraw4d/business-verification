package performance_metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MetricType represents the type of metric being collected
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric represents a single performance metric
type Metric struct {
	Name        string            `json:"name"`
	Type        MetricType        `json:"type"`
	Value       float64           `json:"value"`
	Unit        string            `json:"unit,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	Description string            `json:"description,omitempty"`
}

// MetricsConfig holds configuration for the performance metrics service
type MetricsConfig struct {
	EnableMetricsCollection bool          `json:"enable_metrics_collection"`
	MetricsRetentionPeriod  time.Duration `json:"metrics_retention_period"`
	CollectionInterval      time.Duration `json:"collection_interval"`
	MaxMetricsPerType       int           `json:"max_metrics_per_type"`
}

// DefaultMetricsConfig returns a default configuration
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		EnableMetricsCollection: true,
		MetricsRetentionPeriod:  24 * time.Hour,
		CollectionInterval:      1 * time.Minute,
		MaxMetricsPerType:       1000,
	}
}

// PerformanceMetricsService handles collection and management of performance metrics
type PerformanceMetricsService struct {
	logger  *zap.Logger
	metrics map[string]*Metric
	mutex   sync.RWMutex
	config  *MetricsConfig
}

// NewPerformanceMetricsService creates a new performance metrics service
func NewPerformanceMetricsService(logger *zap.Logger, config *MetricsConfig) *PerformanceMetricsService {
	if config == nil {
		config = DefaultMetricsConfig()
	}

	service := &PerformanceMetricsService{
		logger:  logger,
		metrics: make(map[string]*Metric),
		config:  config,
	}

	if config.EnableMetricsCollection {
		go service.startCleanupRoutine()
	}

	return service
}

// RecordCounter records a counter metric
func (p *PerformanceMetricsService) RecordCounter(ctx context.Context, name string, value float64, labels map[string]string) error {
	if !p.config.EnableMetricsCollection {
		return nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	metric := &Metric{
		Name:      name,
		Type:      MetricTypeCounter,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	key := p.generateMetricKey(name, labels)
	p.metrics[key] = metric

	p.logger.Debug("Recorded counter metric",
		zap.String("name", name),
		zap.Float64("value", value),
		zap.Any("labels", labels))

	return nil
}

// RecordGauge records a gauge metric
func (p *PerformanceMetricsService) RecordGauge(ctx context.Context, name string, value float64, unit string, labels map[string]string) error {
	if !p.config.EnableMetricsCollection {
		return nil
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	metric := &Metric{
		Name:      name,
		Type:      MetricTypeGauge,
		Value:     value,
		Unit:      unit,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	key := p.generateMetricKey(name, labels)
	p.metrics[key] = metric

	p.logger.Debug("Recorded gauge metric",
		zap.String("name", name),
		zap.Float64("value", value),
		zap.String("unit", unit),
		zap.Any("labels", labels))

	return nil
}

// RecordResponseTime records a response time metric
func (p *PerformanceMetricsService) RecordResponseTime(ctx context.Context, operation string, duration time.Duration, labels map[string]string) error {
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["operation"] = operation

	return p.RecordGauge(ctx, "response_time", float64(duration.Milliseconds()), "ms", labels)
}

// RecordThroughput records a throughput metric
func (p *PerformanceMetricsService) RecordThroughput(ctx context.Context, operation string, count int64, labels map[string]string) error {
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["operation"] = operation

	return p.RecordCounter(ctx, "throughput", float64(count), labels)
}

// RecordErrorRate records an error rate metric
func (p *PerformanceMetricsService) RecordErrorRate(ctx context.Context, operation string, rate float64, labels map[string]string) error {
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["operation"] = operation

	return p.RecordGauge(ctx, "error_rate", rate, "percentage", labels)
}

// GetMetrics retrieves all metrics of a specific type
func (p *PerformanceMetricsService) GetMetrics(metricType MetricType) []*Metric {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var metrics []*Metric
	for _, metric := range p.metrics {
		if metric.Type == metricType {
			metrics = append(metrics, metric)
		}
	}
	return metrics
}

// GetMetricsByName retrieves metrics by name
func (p *PerformanceMetricsService) GetMetricsByName(name string) []*Metric {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	var metrics []*Metric
	for _, metric := range p.metrics {
		if metric.Name == name {
			metrics = append(metrics, metric)
		}
	}
	return metrics
}

// GetMetricsStats returns statistics about the collected metrics
func (p *PerformanceMetricsService) GetMetricsStats() map[string]interface{} {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	stats := make(map[string]interface{})

	typeCounts := make(map[MetricType]int)
	for _, metric := range p.metrics {
		typeCounts[metric.Type]++
	}
	stats["metrics_by_type"] = typeCounts
	stats["total_metrics"] = len(p.metrics)

	return stats
}

// ClearMetrics clears all metrics
func (p *PerformanceMetricsService) ClearMetrics() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.metrics = make(map[string]*Metric)
	p.logger.Info("Cleared all performance metrics")
}

// generateMetricKey generates a unique key for a metric
func (p *PerformanceMetricsService) generateMetricKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}

	key := name
	for k, v := range labels {
		key += fmt.Sprintf("_%s_%s", k, v)
	}
	return key
}

// startCleanupRoutine starts the background cleanup routine
func (p *PerformanceMetricsService) startCleanupRoutine() {
	ticker := time.NewTicker(p.config.CollectionInterval)
	defer ticker.Stop()

	for range ticker.C {
		p.cleanupOldMetrics()
	}
}

// cleanupOldMetrics removes metrics older than the retention period
func (p *PerformanceMetricsService) cleanupOldMetrics() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	cutoff := time.Now().Add(-p.config.MetricsRetentionPeriod)

	for key, metric := range p.metrics {
		if metric.Timestamp.Before(cutoff) {
			delete(p.metrics, key)
		}
	}

	p.logger.Debug("Cleaned up old metrics",
		zap.Time("cutoff", cutoff),
		zap.Int("remaining_metrics", len(p.metrics)))
}
