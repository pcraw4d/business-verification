package performance_metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewPerformanceMetricsService(t *testing.T) {
	logger := zap.NewNop()

	// Test with default config
	service := NewPerformanceMetricsService(logger, nil)
	assert.NotNil(t, service)
	assert.NotNil(t, service.config)
	assert.True(t, service.config.EnableMetricsCollection)
	assert.Equal(t, 24*time.Hour, service.config.MetricsRetentionPeriod)
	assert.Equal(t, 1*time.Minute, service.config.CollectionInterval)
	assert.Equal(t, 1000, service.config.MaxMetricsPerType)

	// Test with custom config
	customConfig := &MetricsConfig{
		EnableMetricsCollection: false,
		MetricsRetentionPeriod:  12 * time.Hour,
		CollectionInterval:      30 * time.Second,
		MaxMetricsPerType:       500,
	}

	service = NewPerformanceMetricsService(logger, customConfig)
	assert.NotNil(t, service)
	assert.Equal(t, customConfig, service.config)
}

func TestPerformanceMetricsService_RecordCounter(t *testing.T) {
	logger := zap.NewNop()
	service := NewPerformanceMetricsService(logger, nil)
	ctx := context.Background()

	// Test recording counter metric
	err := service.RecordCounter(ctx, "test_counter", 42.0, map[string]string{"label1": "value1"})
	require.NoError(t, err)

	// Verify metric was recorded
	metrics := service.GetMetrics(MetricTypeCounter)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "test_counter", metrics[0].Name)
	assert.Equal(t, 42.0, metrics[0].Value)
	assert.Equal(t, "value1", metrics[0].Labels["label1"])
	assert.NotZero(t, metrics[0].Timestamp)
}

func TestPerformanceMetricsService_RecordGauge(t *testing.T) {
	logger := zap.NewNop()
	service := NewPerformanceMetricsService(logger, nil)
	ctx := context.Background()

	// Test recording gauge metric
	err := service.RecordGauge(ctx, "test_gauge", 75.5, "percentage", map[string]string{"label1": "value1"})
	require.NoError(t, err)

	// Verify metric was recorded
	metrics := service.GetMetrics(MetricTypeGauge)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "test_gauge", metrics[0].Name)
	assert.Equal(t, 75.5, metrics[0].Value)
	assert.Equal(t, "percentage", metrics[0].Unit)
	assert.Equal(t, "value1", metrics[0].Labels["label1"])
	assert.NotZero(t, metrics[0].Timestamp)
}

func TestPerformanceMetricsService_RecordResponseTime(t *testing.T) {
	logger := zap.NewNop()
	service := NewPerformanceMetricsService(logger, nil)
	ctx := context.Background()

	// Test recording response time
	duration := 150 * time.Millisecond
	err := service.RecordResponseTime(ctx, "api_call", duration, map[string]string{"endpoint": "/users"})
	require.NoError(t, err)

	// Verify metric was recorded
	metrics := service.GetMetrics(MetricTypeGauge)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "response_time", metrics[0].Name)
	assert.Equal(t, 150.0, metrics[0].Value)
	assert.Equal(t, "ms", metrics[0].Unit)
	assert.Equal(t, "api_call", metrics[0].Labels["operation"])
	assert.Equal(t, "/users", metrics[0].Labels["endpoint"])
}

func TestPerformanceMetricsService_RecordThroughput(t *testing.T) {
	logger := zap.NewNop()
	service := NewPerformanceMetricsService(logger, nil)
	ctx := context.Background()

	// Test recording throughput
	err := service.RecordThroughput(ctx, "data_processing", 1000, map[string]string{"batch": "large"})
	require.NoError(t, err)

	// Verify metric was recorded
	metrics := service.GetMetrics(MetricTypeCounter)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "throughput", metrics[0].Name)
	assert.Equal(t, 1000.0, metrics[0].Value)
	assert.Equal(t, "data_processing", metrics[0].Labels["operation"])
	assert.Equal(t, "large", metrics[0].Labels["batch"])
}

func TestPerformanceMetricsService_RecordErrorRate(t *testing.T) {
	logger := zap.NewNop()
	service := NewPerformanceMetricsService(logger, nil)
	ctx := context.Background()

	// Test recording error rate
	err := service.RecordErrorRate(ctx, "payment_processing", 2.5, map[string]string{"gateway": "stripe"})
	require.NoError(t, err)

	// Verify metric was recorded
	metrics := service.GetMetrics(MetricTypeGauge)
	assert.Len(t, metrics, 1)
	assert.Equal(t, "error_rate", metrics[0].Name)
	assert.Equal(t, 2.5, metrics[0].Value)
	assert.Equal(t, "percentage", metrics[0].Unit)
	assert.Equal(t, "payment_processing", metrics[0].Labels["operation"])
	assert.Equal(t, "stripe", metrics[0].Labels["gateway"])
}

func TestPerformanceMetricsService_GetMetricsByName(t *testing.T) {
	logger := zap.NewNop()
	service := NewPerformanceMetricsService(logger, nil)
	ctx := context.Background()

	// Record multiple metrics with same name but different labels
	err := service.RecordCounter(ctx, "test_metric", 10.0, map[string]string{"env": "dev"})
	require.NoError(t, err)

	err = service.RecordCounter(ctx, "test_metric", 20.0, map[string]string{"env": "prod"})
	require.NoError(t, err)

	err = service.RecordGauge(ctx, "other_metric", 30.0, "", nil)
	require.NoError(t, err)

	// Get metrics by name
	metrics := service.GetMetricsByName("test_metric")
	assert.Len(t, metrics, 2)

	// Verify both metrics have the correct name
	for _, metric := range metrics {
		assert.Equal(t, "test_metric", metric.Name)
	}
}

func TestPerformanceMetricsService_GetMetricsStats(t *testing.T) {
	logger := zap.NewNop()
	service := NewPerformanceMetricsService(logger, nil)
	ctx := context.Background()

	// Record various metrics
	err := service.RecordCounter(ctx, "counter1", 10.0, nil)
	require.NoError(t, err)

	err = service.RecordCounter(ctx, "counter2", 20.0, nil)
	require.NoError(t, err)

	err = service.RecordGauge(ctx, "gauge1", 30.0, "", nil)
	require.NoError(t, err)

	// Get stats
	stats := service.GetMetricsStats()

	// Verify stats
	assert.Equal(t, 3, stats["total_metrics"])

	typeCounts := stats["metrics_by_type"].(map[MetricType]int)
	assert.Equal(t, 2, typeCounts[MetricTypeCounter])
	assert.Equal(t, 1, typeCounts[MetricTypeGauge])
}

func TestPerformanceMetricsService_ClearMetrics(t *testing.T) {
	logger := zap.NewNop()
	service := NewPerformanceMetricsService(logger, nil)
	ctx := context.Background()

	// Record some metrics
	err := service.RecordCounter(ctx, "test_counter", 10.0, nil)
	require.NoError(t, err)

	err = service.RecordGauge(ctx, "test_gauge", 20.0, "", nil)
	require.NoError(t, err)

	// Verify metrics exist
	assert.Len(t, service.GetMetrics(MetricTypeCounter), 1)
	assert.Len(t, service.GetMetrics(MetricTypeGauge), 1)

	// Clear metrics
	service.ClearMetrics()

	// Verify metrics are cleared
	assert.Len(t, service.GetMetrics(MetricTypeCounter), 0)
	assert.Len(t, service.GetMetrics(MetricTypeGauge), 0)
}

func TestPerformanceMetricsService_DisabledMetricsCollection(t *testing.T) {
	logger := zap.NewNop()
	config := &MetricsConfig{
		EnableMetricsCollection: false,
		MetricsRetentionPeriod:  24 * time.Hour,
		CollectionInterval:      1 * time.Minute,
		MaxMetricsPerType:       1000,
	}
	service := NewPerformanceMetricsService(logger, config)
	ctx := context.Background()

	// Try to record metrics
	err := service.RecordCounter(ctx, "test_counter", 10.0, nil)
	require.NoError(t, err) // Should not error, but should not record

	err = service.RecordGauge(ctx, "test_gauge", 20.0, "", nil)
	require.NoError(t, err) // Should not error, but should not record

	// Verify no metrics were recorded
	assert.Len(t, service.GetMetrics(MetricTypeCounter), 0)
	assert.Len(t, service.GetMetrics(MetricTypeGauge), 0)
}
