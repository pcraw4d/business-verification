package caching

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewCacheMonitor(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	t.Run("create monitor with default config", func(t *testing.T) {
		config := CacheMonitorConfig{
			Enabled: true,
			Logger:  zap.NewNop(),
		}

		monitor := NewCacheMonitor(cache, config)
		assert.NotNil(t, monitor)
		assert.Equal(t, cache, monitor.cache)
		assert.Equal(t, 30*time.Second, monitor.config.CollectionInterval)
		assert.Equal(t, 24*time.Hour, monitor.config.RetentionPeriod)
		assert.Equal(t, 10000, monitor.config.MaxMetrics)
		assert.Equal(t, 1*time.Hour, monitor.config.PredictionWindow)
		assert.True(t, monitor.config.Enabled)

		monitor.Close()
	})

	t.Run("create monitor with custom config", func(t *testing.T) {
		config := CacheMonitorConfig{
			Enabled:            true,
			CollectionInterval: 10 * time.Second,
			RetentionPeriod:    12 * time.Hour,
			MaxMetrics:         5000,
			PredictionWindow:   30 * time.Minute,
			Logger:             zap.NewNop(),
		}

		monitor := NewCacheMonitor(cache, config)
		assert.NotNil(t, monitor)
		assert.Equal(t, 10*time.Second, monitor.config.CollectionInterval)
		assert.Equal(t, 12*time.Hour, monitor.config.RetentionPeriod)
		assert.Equal(t, 5000, monitor.config.MaxMetrics)
		assert.Equal(t, 30*time.Minute, monitor.config.PredictionWindow)

		monitor.Close()
	})

	t.Run("create monitor with nil logger", func(t *testing.T) {
		config := CacheMonitorConfig{
			Enabled: true,
		}

		monitor := NewCacheMonitor(cache, config)
		assert.NotNil(t, monitor)
		assert.NotNil(t, monitor.config.Logger)

		monitor.Close()
	})
}

func TestCacheMonitor_RecordMetric(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false, // Disable background worker for testing
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("record hit rate metric", func(t *testing.T) {
		monitor.RecordMetric(CacheMetricTypeHitRate, 0.85, map[string]string{"shard": "0"})

		metrics := monitor.GetMetrics(CacheMetricTypeHitRate, time.Now().Add(-time.Minute), time.Now().Add(time.Minute))
		assert.Len(t, metrics, 1)
		assert.Equal(t, CacheMetricTypeHitRate, metrics[0].Type)
		assert.Equal(t, 0.85, metrics[0].Value)
		assert.Equal(t, "0", metrics[0].Labels["shard"])
	})

	t.Run("record multiple metrics", func(t *testing.T) {
		monitor.RecordMetric(CacheMetricTypeMissRate, 0.15, nil)
		monitor.RecordMetric(CacheMetricTypeLatency, 5.2, nil)
		monitor.RecordMetric(CacheMetricTypeSize, 1024.0, nil)

		hitRateMetrics := monitor.GetMetrics(CacheMetricTypeHitRate, time.Now().Add(-time.Minute), time.Now().Add(time.Minute))
		missRateMetrics := monitor.GetMetrics(CacheMetricTypeMissRate, time.Now().Add(-time.Minute), time.Now().Add(time.Minute))
		latencyMetrics := monitor.GetMetrics(CacheMetricTypeLatency, time.Now().Add(-time.Minute), time.Now().Add(time.Minute))

		assert.Len(t, hitRateMetrics, 1)
		assert.Len(t, missRateMetrics, 1)
		assert.Len(t, latencyMetrics, 1)
		assert.Equal(t, 0.15, missRateMetrics[0].Value)
		assert.Equal(t, 5.2, latencyMetrics[0].Value)
	})
}

func TestCacheMonitor_GetMetrics(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("get metrics for time range", func(t *testing.T) {
		// Record metrics at different times
		now := time.Now()
		monitor.RecordMetric(CacheMetricTypeHitRate, 0.8, nil)
		time.Sleep(10 * time.Millisecond)
		monitor.RecordMetric(CacheMetricTypeHitRate, 0.9, nil)
		time.Sleep(10 * time.Millisecond)
		monitor.RecordMetric(CacheMetricTypeHitRate, 0.7, nil)

		// Get metrics for a specific range
		start := now.Add(-time.Minute)
		end := now.Add(time.Minute)
		metrics := monitor.GetMetrics(CacheMetricTypeHitRate, start, end)

		assert.Len(t, metrics, 3)
		assert.Equal(t, 0.8, metrics[0].Value)
		assert.Equal(t, 0.9, metrics[1].Value)
		assert.Equal(t, 0.7, metrics[2].Value)
	})

	t.Run("get metrics for empty range", func(t *testing.T) {
		start := time.Now().Add(-time.Hour)
		end := time.Now().Add(-30 * time.Minute)
		metrics := monitor.GetMetrics(CacheMetricTypeHitRate, start, end)

		assert.Len(t, metrics, 0)
	})
}

func TestCacheMonitor_GetSnapshots(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("get snapshots for time range", func(t *testing.T) {
		// Create snapshots manually
		snapshot1 := monitor.createSnapshot()
		time.Sleep(10 * time.Millisecond)
		snapshot2 := monitor.createSnapshot()

		monitor.mu.Lock()
		monitor.snapshots = append(monitor.snapshots, *snapshot1, *snapshot2)
		monitor.mu.Unlock()

		start := time.Now().Add(-time.Minute)
		end := time.Now().Add(time.Minute)
		snapshots := monitor.GetSnapshots(start, end)

		assert.Len(t, snapshots, 2)
		assert.True(t, snapshots[0].Timestamp.Before(snapshots[1].Timestamp))
	})
}

func TestCacheMonitor_GetCurrentSnapshot(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("get current snapshot", func(t *testing.T) {
		snapshot := monitor.GetCurrentSnapshot()
		assert.NotNil(t, snapshot)
		assert.NotZero(t, snapshot.Timestamp)
		assert.GreaterOrEqual(t, snapshot.HitRate, 0.0)
		assert.LessOrEqual(t, snapshot.HitRate, 1.0)
		assert.GreaterOrEqual(t, snapshot.MissRate, 0.0)
		assert.LessOrEqual(t, snapshot.MissRate, 1.0)
		assert.GreaterOrEqual(t, snapshot.TotalSize, int64(0))
		assert.GreaterOrEqual(t, snapshot.EntryCount, int64(0))
		assert.GreaterOrEqual(t, snapshot.ShardCount, 1)
	})
}

func TestCacheMonitor_Alerts(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		AlertThresholds: map[CacheMetricType]float64{
			CacheMetricTypeHitRate:  0.8, // Alert if hit rate < 80%
			CacheMetricTypeMissRate: 0.2, // Alert if miss rate > 20%
		},
		Logger: zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("trigger hit rate alert", func(t *testing.T) {
		monitor.RecordMetric(CacheMetricTypeHitRate, 0.75, nil) // Below threshold

		alerts := monitor.GetAlerts()
		assert.Len(t, alerts, 1)
		assert.Equal(t, string(CacheMetricTypeHitRate), alerts[0].Type)
		assert.Equal(t, "low", alerts[0].Severity)
		assert.Contains(t, alerts[0].Message, "75.00%")
		assert.False(t, alerts[0].Acknowledged)
	})

	t.Run("trigger miss rate alert", func(t *testing.T) {
		monitor.RecordMetric(CacheMetricTypeMissRate, 0.25, nil) // Above threshold

		alerts := monitor.GetAlerts()
		assert.Len(t, alerts, 2) // Previous alert + new alert

		// Find the miss rate alert
		var missRateAlert *CacheAlert
		for i := range alerts {
			if alerts[i].Type == string(CacheMetricTypeMissRate) {
				missRateAlert = &alerts[i]
				break
			}
		}

		assert.NotNil(t, missRateAlert)
		assert.Equal(t, "medium", missRateAlert.Severity)
		assert.Contains(t, missRateAlert.Message, "25.00%")
	})

	t.Run("acknowledge alert", func(t *testing.T) {
		alerts := monitor.GetAlerts()
		require.Len(t, alerts, 2)

		alertID := alerts[0].ID
		err := monitor.AcknowledgeAlert(alertID)
		assert.NoError(t, err)

		// Check that alert is acknowledged
		monitor.mu.RLock()
		var acknowledged bool
		for _, alert := range monitor.alerts {
			if alert.ID == alertID {
				acknowledged = alert.Acknowledged
				break
			}
		}
		monitor.mu.RUnlock()
		assert.True(t, acknowledged)

		// Active alerts should be reduced
		activeAlerts := monitor.GetAlerts()
		assert.Len(t, activeAlerts, 1)
	})

	t.Run("acknowledge non-existent alert", func(t *testing.T) {
		err := monitor.AcknowledgeAlert("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestCacheMonitor_Bottlenecks(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		BottleneckThresholds: map[CacheMetricType]float64{
			CacheMetricTypeHitRate: 0.7,  // Bottleneck if hit rate < 70%
			CacheMetricTypeLatency: 10.0, // Bottleneck if latency > 10ms
		},
		Logger: zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("detect hit rate bottleneck", func(t *testing.T) {
		monitor.RecordMetric(CacheMetricTypeHitRate, 0.65, nil) // Below threshold

		bottlenecks := monitor.GetBottlenecks()
		assert.Len(t, bottlenecks, 1)
		assert.Equal(t, string(CacheMetricTypeHitRate), bottlenecks[0].Type)
		assert.Equal(t, "low", bottlenecks[0].Severity)
		assert.Contains(t, bottlenecks[0].Description, "65.00%")
		assert.NotEmpty(t, bottlenecks[0].Recommendations)
	})

	t.Run("detect latency bottleneck", func(t *testing.T) {
		monitor.RecordMetric(CacheMetricTypeLatency, 15.0, nil) // Above threshold

		bottlenecks := monitor.GetBottlenecks()
		assert.Len(t, bottlenecks, 2) // Previous bottleneck + new bottleneck

		// Find the latency bottleneck
		var latencyBottleneck *CacheBottleneck
		for i := range bottlenecks {
			if bottlenecks[i].Type == string(CacheMetricTypeLatency) {
				latencyBottleneck = &bottlenecks[i]
				break
			}
		}

		assert.NotNil(t, latencyBottleneck)
		assert.Equal(t, "medium", latencyBottleneck.Severity)
		assert.Contains(t, latencyBottleneck.Description, "15.00ms")
		assert.NotEmpty(t, latencyBottleneck.Recommendations)
	})
}

func TestCacheMonitor_AlertHandlers(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		AlertThresholds: map[CacheMetricType]float64{
			CacheMetricTypeHitRate: 0.8,
		},
		Logger: zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("alert handler is called", func(t *testing.T) {
		var receivedAlert *CacheAlert
		handler := func(alert *CacheAlert) {
			receivedAlert = alert
		}

		monitor.AddAlertHandler(handler)
		monitor.RecordMetric(CacheMetricTypeHitRate, 0.75, nil)

		assert.NotNil(t, receivedAlert)
		assert.Equal(t, string(CacheMetricTypeHitRate), receivedAlert.Type)
		assert.Equal(t, "low", receivedAlert.Severity)
	})
}

func TestCacheMonitor_BottleneckHandlers(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		BottleneckThresholds: map[CacheMetricType]float64{
			CacheMetricTypeHitRate: 0.7,
		},
		Logger: zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("bottleneck handler is called", func(t *testing.T) {
		var receivedBottleneck *CacheBottleneck
		handler := func(bottleneck *CacheBottleneck) {
			receivedBottleneck = bottleneck
		}

		monitor.AddBottleneckHandler(handler)
		monitor.RecordMetric(CacheMetricTypeHitRate, 0.65, nil)

		assert.NotNil(t, receivedBottleneck)
		assert.Equal(t, string(CacheMetricTypeHitRate), receivedBottleneck.Type)
		assert.Equal(t, "low", receivedBottleneck.Severity)
		assert.NotEmpty(t, receivedBottleneck.Recommendations)
	})
}

func TestCacheMonitor_GenerateReport(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled:       false,
		TrendAnalysis: true,
		Logger:        zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("generate report with data", func(t *testing.T) {
		// Add some test data for trend analysis
		for i := 0; i < 5; i++ {
			monitor.RecordMetric(CacheMetricTypeHitRate, 0.8+float64(i)*0.02, nil)
			monitor.RecordMetric(CacheMetricTypeMissRate, 0.2-float64(i)*0.02, nil)
			monitor.RecordMetric(CacheMetricTypeLatency, 5.0+float64(i)*0.5, nil)
		}

		// Create a snapshot
		snapshot := monitor.createSnapshot()
		monitor.mu.Lock()
		monitor.snapshots = append(monitor.snapshots, *snapshot)
		monitor.mu.Unlock()

		// Generate report
		report := monitor.GenerateReport(1 * time.Hour)

		assert.NotNil(t, report)
		assert.Equal(t, 1*time.Hour, report.Period)
		assert.NotZero(t, report.GeneratedAt)
		assert.Len(t, report.Metrics, 15) // 5 iterations * 3 metrics
		assert.NotNil(t, report.Summary)
		assert.NotEmpty(t, report.Recommendations)
		assert.NotEmpty(t, report.Trends)
	})

	t.Run("generate report with no data", func(t *testing.T) {
		// Create a new monitor with no data
		emptyMonitor := NewCacheMonitor(cache, CacheMonitorConfig{
			Enabled:       false,
			TrendAnalysis: true,
			Logger:        zap.NewNop(),
		})
		defer emptyMonitor.Close()

		report := emptyMonitor.GenerateReport(1 * time.Hour)

		assert.NotNil(t, report)
		assert.Equal(t, 1*time.Hour, report.Period)
		assert.Len(t, report.Metrics, 0)
		assert.Empty(t, report.Bottlenecks)
		assert.Empty(t, report.Alerts)
	})
}

func TestCacheMonitor_Cleanup(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled:         false,
		RetentionPeriod: 100 * time.Millisecond,
		MaxMetrics:      5,
		Logger:          zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("cleanup old metrics", func(t *testing.T) {
		// Add metrics
		for i := 0; i < 10; i++ {
			monitor.RecordMetric(CacheMetricTypeHitRate, float64(i)/10.0, nil)
		}

		// Wait for cleanup
		time.Sleep(150 * time.Millisecond)

		// Trigger cleanup
		monitor.RecordMetric(CacheMetricTypeHitRate, 0.5, nil)

		metrics := monitor.GetMetrics(CacheMetricTypeHitRate, time.Now().Add(-time.Minute), time.Now().Add(time.Minute))
		assert.Len(t, metrics, 1) // Only the latest metric should remain
		assert.Equal(t, 0.5, metrics[0].Value)
	})

	t.Run("limit max metrics", func(t *testing.T) {
		// Add more metrics than max
		for i := 0; i < 10; i++ {
			monitor.RecordMetric(CacheMetricTypeMissRate, float64(i)/10.0, nil)
		}

		metrics := monitor.GetMetrics(CacheMetricTypeMissRate, time.Now().Add(-time.Minute), time.Now().Add(time.Minute))
		assert.Len(t, metrics, 5) // Should be limited to MaxMetrics
	})
}

func TestCacheMonitor_SeverityCalculation(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("severity levels", func(t *testing.T) {
		// Test different severity levels
		testCases := []struct {
			value     float64
			threshold float64
			expected  string
		}{
			{0.5, 1.0, "low"},      // ratio = 0.5
			{1.3, 1.0, "medium"},   // ratio = 1.3
			{1.6, 1.0, "high"},     // ratio = 1.6
			{2.5, 1.0, "critical"}, // ratio = 2.5
		}

		for _, tc := range testCases {
			severity := monitor.determineSeverity(tc.value, tc.threshold)
			assert.Equal(t, tc.expected, severity, "value: %f, threshold: %f", tc.value, tc.threshold)
		}
	})
}

func TestCacheMonitor_TrendAnalysis(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled:       false,
		TrendAnalysis: true,
		Logger:        zap.NewNop(),
	})
	defer monitor.Close()

	t.Run("calculate trend for increasing values", func(t *testing.T) {
		metrics := []CacheMetric{
			{Type: CacheMetricTypeHitRate, Value: 0.5, Timestamp: time.Now()},
			{Type: CacheMetricTypeHitRate, Value: 0.6, Timestamp: time.Now()},
			{Type: CacheMetricTypeHitRate, Value: 0.7, Timestamp: time.Now()},
		}

		trend := monitor.calculateTrend(metrics)
		assert.Equal(t, CacheMetricTypeHitRate, trend.Metric)
		assert.Equal(t, "increasing", trend.Direction)
		assert.Greater(t, trend.Slope, 0.0)
		assert.Greater(t, trend.Confidence, 0.0)
	})

	t.Run("calculate trend for decreasing values", func(t *testing.T) {
		metrics := []CacheMetric{
			{Type: CacheMetricTypeHitRate, Value: 0.7, Timestamp: time.Now()},
			{Type: CacheMetricTypeHitRate, Value: 0.6, Timestamp: time.Now()},
			{Type: CacheMetricTypeHitRate, Value: 0.5, Timestamp: time.Now()},
		}

		trend := monitor.calculateTrend(metrics)
		assert.Equal(t, CacheMetricTypeHitRate, trend.Metric)
		assert.Equal(t, "decreasing", trend.Direction)
		assert.Less(t, trend.Slope, 0.0)
		assert.Greater(t, trend.Confidence, 0.0)
	})

	t.Run("calculate trend for stable values", func(t *testing.T) {
		metrics := []CacheMetric{
			{Type: CacheMetricTypeHitRate, Value: 0.5, Timestamp: time.Now()},
			{Type: CacheMetricTypeHitRate, Value: 0.51, Timestamp: time.Now()},
			{Type: CacheMetricTypeHitRate, Value: 0.49, Timestamp: time.Now()},
		}

		trend := monitor.calculateTrend(metrics)
		assert.Equal(t, CacheMetricTypeHitRate, trend.Metric)
		assert.Equal(t, "stable", trend.Direction)
		assert.Less(t, abs(trend.Slope), 0.01)
	})

	t.Run("calculate trend with insufficient data", func(t *testing.T) {
		metrics := []CacheMetric{
			{Type: CacheMetricTypeHitRate, Value: 0.5, Timestamp: time.Now()},
		}

		trend := monitor.calculateTrend(metrics)
		assert.Equal(t, CacheMetricType(""), trend.Metric)
		assert.Equal(t, "", trend.Direction)
		assert.Equal(t, 0.0, trend.Slope)
	})
}

func TestCacheMonitor_Close(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})

	t.Run("close monitor", func(t *testing.T) {
		err := monitor.Close()
		assert.NoError(t, err)
	})
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Benchmark tests
func BenchmarkCacheMonitor_RecordMetric(b *testing.B) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(b, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			monitor.RecordMetric(CacheMetricTypeHitRate, float64(i%100)/100.0, nil)
			i++
		}
	})
}

func BenchmarkCacheMonitor_GetMetrics(b *testing.B) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(b, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	// Pre-populate with metrics
	for i := 0; i < 1000; i++ {
		monitor.RecordMetric(CacheMetricTypeHitRate, float64(i%100)/100.0, nil)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			start := time.Now().Add(-time.Hour)
			end := time.Now().Add(time.Hour)
			monitor.GetMetrics(CacheMetricTypeHitRate, start, end)
		}
	})
}
