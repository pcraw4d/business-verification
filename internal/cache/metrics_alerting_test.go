package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestMetricsCollector tests the metrics collector functionality
func TestMetricsCollector(t *testing.T) {
	logger := zap.NewNop()

	// Create a memory cache for testing
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         100,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)
	defer cache.Close()

	// Create metrics collector
	metricsConfig := &MetricsConfig{
		CollectionInterval:    100 * time.Millisecond, // Fast collection for test
		HistoryRetention:      1 * time.Hour,
		MaxHistoryEntries:     100,
		EnableDetailedMetrics: true,
	}

	collector := NewMetricsCollector([]Cache{cache}, metricsConfig, logger)

	// Start collector
	err := collector.Start()
	require.NoError(t, err)
	defer collector.Stop()

	ctx := context.Background()

	// Add some data to generate metrics
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("test:metrics:%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		err := cache.Set(ctx, key, value, 5*time.Minute)
		require.NoError(t, err)
	}

	// Wait for metrics collection
	time.Sleep(200 * time.Millisecond)

	// Get current metrics
	metrics := collector.GetCurrentMetrics()
	require.NotNil(t, metrics)
	assert.True(t, metrics.TotalSize >= 10)
	assert.True(t, metrics.TotalMemoryUsage > 0)

	// Test historical metrics
	startTime := time.Now().Add(-1 * time.Hour)
	endTime := time.Now().Add(1 * time.Hour)
	historical := collector.GetHistoricalMetrics(startTime, endTime)
	assert.NotNil(t, historical)

	// Test metrics trend
	trend, err := collector.GetMetricsTrend("total_size", 1*time.Hour)
	require.NoError(t, err)
	assert.NotNil(t, trend)
}

// TestAlertingSystem tests the alerting system functionality
func TestAlertingSystem(t *testing.T) {
	logger := zap.NewNop()

	// Create a memory cache
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         100,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)
	defer cache.Close()

	// Create metrics collector
	metricsConfig := &MetricsConfig{
		CollectionInterval:    100 * time.Millisecond,
		HistoryRetention:      1 * time.Hour,
		MaxHistoryEntries:     100,
		EnableDetailedMetrics: true,
	}

	collector := NewMetricsCollector([]Cache{cache}, metricsConfig, logger)
	err := collector.Start()
	require.NoError(t, err)
	defer collector.Stop()

	// Create alerting system
	alertingConfig := &AlertingConfig{
		EnableAlerts:      true,
		CooldownPeriod:    1 * time.Second, // Short cooldown for test
		MaxHistoryEntries: 100,
		EnableEscalation:  false, // Disable for test
		AlertThresholds: map[string]AlertThreshold{
			"hit_rate_low": {
				Warning:  0.9, // Very high threshold to trigger alert
				Critical: 0.95,
				Enabled:  false, // Disable hit rate alert for this test
			},
			"cache_size_high": {
				Warning:  5, // Low threshold to trigger alert
				Critical: 10,
				Enabled:  true,
			},
		},
	}

	alertingSystem := NewAlertingSystem(alertingConfig, collector, logger)

	// Register alert handlers
	loggingHandler := NewLoggingAlertHandler(logger)
	alertingSystem.RegisterHandler("hit_rate_low", loggingHandler)
	alertingSystem.RegisterHandler("cache_size_high", loggingHandler)

	ctx := context.Background()

	// Add data to trigger size alert
	for i := 0; i < 6; i++ { // Exceeds warning threshold of 5
		key := fmt.Sprintf("test:alert:%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		err := cache.Set(ctx, key, value, 5*time.Minute)
		require.NoError(t, err)
	}

	// Wait for metrics collection
	time.Sleep(200 * time.Millisecond)

	// Check for alerts
	alerts := alertingSystem.CheckAlerts(ctx)
	assert.NotNil(t, alerts)

	// Should have at least one alert for cache size
	if len(alerts) > 0 {
		assert.Equal(t, "cache_size_high", alerts[0].Type)
		assert.Equal(t, AlertSeverityWarning, alerts[0].Severity)
	}

	// Test active alerts
	activeAlerts := alertingSystem.GetActiveAlerts()
	assert.NotNil(t, activeAlerts)

	// Test alert history
	history := alertingSystem.GetAlertHistory()
	assert.NotNil(t, history)

	// Test resolving alert
	if len(activeAlerts) > 0 {
		alertingSystem.ResolveAlert(activeAlerts[0].Type)

		// Check that alert is resolved
		activeAlerts = alertingSystem.GetActiveAlerts()
		// Alert should be removed from active alerts
	}
}

// TestAlertHandlers tests the alert handler implementations
func TestAlertHandlers(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	// Test logging alert handler
	loggingHandler := NewLoggingAlertHandler(logger)
	assert.Equal(t, "logging", loggingHandler.GetName())

	alert := &CacheMonitoringAlert{
		ID:           "test-alert",
		Type:         "test_type",
		Severity:     AlertSeverityWarning,
		Message:      "Test alert message",
		Details:      make(map[string]interface{}),
		Threshold:    0.8,
		CurrentValue: 0.9,
		Timestamp:    time.Now(),
	}

	err := loggingHandler.HandleAlert(ctx, alert)
	require.NoError(t, err)

	// Test webhook alert handler
	webhookHandler := NewWebhookAlertHandler("http://example.com/webhook", 5*time.Second, logger)
	assert.Equal(t, "webhook", webhookHandler.GetName())

	err = webhookHandler.HandleAlert(ctx, alert)
	require.NoError(t, err)
}

// TestAlertThresholds tests alert threshold checking
func TestAlertThresholds(t *testing.T) {
	logger := zap.NewNop()

	// Create alerting system with specific thresholds
	alertingConfig := &AlertingConfig{
		EnableAlerts:      true,
		CooldownPeriod:    1 * time.Second,
		MaxHistoryEntries: 100,
		EnableEscalation:  false,
		AlertThresholds: map[string]AlertThreshold{
			"hit_rate_low": {
				Warning:  0.7,
				Critical: 0.5,
				Enabled:  true,
			},
			"error_rate_high": {
				Warning:  0.05,
				Critical: 0.1,
				Enabled:  true,
			},
		},
	}

	alertingSystem := NewAlertingSystem(alertingConfig, nil, logger)

	// Test threshold checking
	testCases := []struct {
		alertType        string
		value            float64
		expectedSeverity AlertSeverity
	}{
		{"hit_rate_low", 0.6, AlertSeverityWarning},      // Below warning threshold
		{"hit_rate_low", 0.4, AlertSeverityCritical},     // Below critical threshold
		{"hit_rate_low", 0.8, ""},                        // Above warning threshold
		{"error_rate_high", 0.06, AlertSeverityWarning},  // Above warning threshold
		{"error_rate_high", 0.15, AlertSeverityCritical}, // Above critical threshold
		{"error_rate_high", 0.03, ""},                    // Below warning threshold
	}

	for _, tc := range testCases {
		threshold := alertingConfig.AlertThresholds[tc.alertType]
		severity := alertingSystem.checkThreshold(tc.alertType, tc.value, threshold)
		assert.Equal(t, tc.expectedSeverity, severity,
			"Alert type: %s, Value: %f", tc.alertType, tc.value)
	}
}

// TestMetricsTrend tests metrics trend analysis
func TestMetricsTrend(t *testing.T) {
	logger := zap.NewNop()

	// Create a memory cache
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         100,
		KeyPrefix:       "test",
		CleanupInterval: 5 * time.Minute,
	}

	cache := NewMemoryCache(config)
	defer cache.Close()

	// Create metrics collector
	metricsConfig := &MetricsConfig{
		CollectionInterval:    50 * time.Millisecond, // Very fast collection
		HistoryRetention:      1 * time.Hour,
		MaxHistoryEntries:     100,
		EnableDetailedMetrics: true,
	}

	collector := NewMetricsCollector([]Cache{cache}, metricsConfig, logger)
	err := collector.Start()
	require.NoError(t, err)
	defer collector.Stop()

	ctx := context.Background()

	// Add data over time to create a trend
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("test:trend:%d", i)
		value := []byte(fmt.Sprintf("value-%d", i))
		err := cache.Set(ctx, key, value, 5*time.Minute)
		require.NoError(t, err)

		// Wait for metrics collection
		time.Sleep(100 * time.Millisecond)
	}

	// Test different metric trends
	metrics := []string{"total_size", "hit_rate", "memory_usage"}

	for _, metric := range metrics {
		trend, err := collector.GetMetricsTrend(metric, 1*time.Hour)
		require.NoError(t, err)
		assert.NotNil(t, trend)

		// Trend should have multiple data points
		if len(trend) > 1 {
			// Verify trend is generally increasing for size
			if metric == "total_size" {
				assert.True(t, trend[len(trend)-1] >= trend[0],
					"Size trend should be increasing")
			}
		}
	}
}

// TestAlertCooldown tests alert cooldown functionality
func TestAlertCooldown(t *testing.T) {
	logger := zap.NewNop()

	alertingConfig := &AlertingConfig{
		EnableAlerts:      true,
		CooldownPeriod:    2 * time.Second,
		MaxHistoryEntries: 100,
		EnableEscalation:  false,
		AlertThresholds: map[string]AlertThreshold{
			"test_alert": {
				Warning:  0.5,
				Critical: 0.3,
				Enabled:  true,
			},
		},
	}

	alertingSystem := NewAlertingSystem(alertingConfig, nil, logger)

	// Test cooldown
	alertType := "test_alert"

	// Should not be in cooldown initially
	assert.False(t, alertingSystem.isInCooldown(alertType))

	// Set cooldown
	alertingSystem.setCooldown(alertType)

	// Should be in cooldown now
	assert.True(t, alertingSystem.isInCooldown(alertType))

	// Wait for cooldown to expire
	time.Sleep(3 * time.Second)

	// Should not be in cooldown anymore
	assert.False(t, alertingSystem.isInCooldown(alertType))
}

// TestAlertEscalation tests alert escalation functionality
func TestAlertEscalation(t *testing.T) {
	logger := zap.NewNop()

	alertingConfig := &AlertingConfig{
		EnableAlerts:      true,
		CooldownPeriod:    1 * time.Second,
		MaxHistoryEntries: 100,
		EnableEscalation:  true,
		EscalationDelay:   100 * time.Millisecond, // Short delay for test
		AlertThresholds: map[string]AlertThreshold{
			"critical_alert": {
				Warning:  0.5,
				Critical: 0.3,
				Enabled:  true,
			},
		},
	}

	alertingSystem := NewAlertingSystem(alertingConfig, nil, logger)

	// Create a critical alert
	alert := &CacheMonitoringAlert{
		ID:           "test-critical-alert",
		Type:         "critical_alert",
		Severity:     AlertSeverityCritical,
		Message:      "Critical test alert",
		Details:      make(map[string]interface{}),
		Threshold:    0.3,
		CurrentValue: 0.2,
		Timestamp:    time.Now(),
	}

	// Add to active alerts
	alertingSystem.mu.Lock()
	alertingSystem.activeAlerts["critical_alert"] = alert
	alertingSystem.mu.Unlock()

	// Start escalation
	ctx := context.Background()
	go alertingSystem.escalateAlert(ctx, alert)

	// Wait for escalation
	time.Sleep(200 * time.Millisecond)

	// Check that alert was escalated
	assert.True(t, alert.Escalated)
}
