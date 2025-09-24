package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MetricsCollector collects and aggregates cache metrics
type MetricsCollector struct {
	caches []Cache
	config *MetricsConfig
	logger *zap.Logger

	// Metrics storage
	metrics     *AggregatedMetrics
	metricsLock sync.RWMutex

	// Collection control
	stopChan chan struct{}
	running  bool
	mu       sync.RWMutex

	// Historical data
	historicalData []*MetricsSnapshot
	historyLock    sync.RWMutex
}

// MetricsConfig holds configuration for metrics collection
type MetricsConfig struct {
	CollectionInterval    time.Duration `json:"collection_interval"`
	HistoryRetention      time.Duration `json:"history_retention"`
	MaxHistoryEntries     int           `json:"max_history_entries"`
	EnableDetailedMetrics bool          `json:"enable_detailed_metrics"`
}

// AggregatedMetrics holds aggregated metrics from all caches
type AggregatedMetrics struct {
	Timestamp           time.Time     `json:"timestamp"`
	TotalHits           int64         `json:"total_hits"`
	TotalMisses         int64         `json:"total_misses"`
	TotalErrors         int64         `json:"total_errors"`
	TotalSize           int64         `json:"total_size"`
	TotalMemoryUsage    int64         `json:"total_memory_usage"`
	AverageHitRate      float64       `json:"average_hit_rate"`
	AverageErrorRate    float64       `json:"average_error_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`

	// Per-cache metrics
	CacheMetrics map[string]*DetailedCacheMetrics `json:"cache_metrics"`

	// Performance metrics
	PerformanceMetrics *PerformanceMetrics `json:"performance_metrics"`
}

// DetailedCacheMetrics holds metrics for a single cache
type DetailedCacheMetrics struct {
	CacheName           string        `json:"cache_name"`
	HitCount            int64         `json:"hit_count"`
	MissCount           int64         `json:"miss_count"`
	ErrorCount          int64         `json:"error_count"`
	Size                int64         `json:"size"`
	MemoryUsage         int64         `json:"memory_usage"`
	HitRate             float64       `json:"hit_rate"`
	ErrorRate           float64       `json:"error_rate"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// PerformanceMetrics holds performance-related metrics
type PerformanceMetrics struct {
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	P95Latency         time.Duration `json:"p95_latency"`
	P99Latency         time.Duration `json:"p99_latency"`
	Throughput         float64       `json:"throughput"` // requests per second
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	Timestamp time.Time          `json:"timestamp"`
	Metrics   *AggregatedMetrics `json:"metrics"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(caches []Cache, config *MetricsConfig, logger *zap.Logger) *MetricsCollector {
	if config == nil {
		config = &MetricsConfig{
			CollectionInterval:    30 * time.Second,
			HistoryRetention:      24 * time.Hour,
			MaxHistoryEntries:     1000,
			EnableDetailedMetrics: true,
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &MetricsCollector{
		caches:         caches,
		config:         config,
		logger:         logger,
		metrics:        &AggregatedMetrics{},
		historicalData: make([]*MetricsSnapshot, 0),
		stopChan:       make(chan struct{}),
	}
}

// Start begins metrics collection
func (mc *MetricsCollector) Start() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if mc.running {
		return fmt.Errorf("metrics collector is already running")
	}

	mc.running = true
	mc.stopChan = make(chan struct{})

	mc.logger.Info("Starting metrics collector",
		zap.Duration("collection_interval", mc.config.CollectionInterval),
		zap.Duration("history_retention", mc.config.HistoryRetention))

	// Start collection goroutine
	go mc.collectMetrics()

	// Start cleanup goroutine
	go mc.cleanupHistory()

	return nil
}

// Stop stops metrics collection
func (mc *MetricsCollector) Stop() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.running {
		return fmt.Errorf("metrics collector is not running")
	}

	mc.running = false
	close(mc.stopChan)

	mc.logger.Info("Stopped metrics collector")
	return nil
}

// GetCurrentMetrics returns the current aggregated metrics
func (mc *MetricsCollector) GetCurrentMetrics() *AggregatedMetrics {
	mc.metricsLock.RLock()
	defer mc.metricsLock.RUnlock()

	// Return a copy of the metrics
	metrics := *mc.metrics
	metrics.CacheMetrics = make(map[string]*DetailedCacheMetrics)
	for k, v := range mc.metrics.CacheMetrics {
		cacheMetrics := *v
		metrics.CacheMetrics[k] = &cacheMetrics
	}

	if mc.metrics.PerformanceMetrics != nil {
		perfMetrics := *mc.metrics.PerformanceMetrics
		metrics.PerformanceMetrics = &perfMetrics
	}

	return &metrics
}

// GetHistoricalMetrics returns historical metrics within the specified time range
func (mc *MetricsCollector) GetHistoricalMetrics(startTime, endTime time.Time) []*MetricsSnapshot {
	mc.historyLock.RLock()
	defer mc.historyLock.RUnlock()

	var result []*MetricsSnapshot
	for _, snapshot := range mc.historicalData {
		if snapshot.Timestamp.After(startTime) && snapshot.Timestamp.Before(endTime) {
			result = append(result, snapshot)
		}
	}

	return result
}

// GetMetricsTrend returns the trend of a specific metric over time
func (mc *MetricsCollector) GetMetricsTrend(metricName string, duration time.Duration) ([]float64, error) {
	mc.historyLock.RLock()
	defer mc.historyLock.RUnlock()

	cutoff := time.Now().Add(-duration)
	var values []float64

	for _, snapshot := range mc.historicalData {
		if snapshot.Timestamp.After(cutoff) {
			value, err := mc.extractMetricValue(snapshot.Metrics, metricName)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
	}

	return values, nil
}

// collectMetrics collects metrics from all caches
func (mc *MetricsCollector) collectMetrics() {
	ticker := time.NewTicker(mc.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.updateMetrics()
		case <-mc.stopChan:
			return
		}
	}
}

// updateMetrics updates the aggregated metrics
func (mc *MetricsCollector) updateMetrics() {
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var totalHits, totalMisses, totalErrors, totalSize, totalMemoryUsage int64
	cacheMetrics := make(map[string]*DetailedCacheMetrics)

	// Collect metrics from each cache
	for i, cache := range mc.caches {
		cacheName := fmt.Sprintf("cache_%d", i)

		// Get cache stats
		stats, err := cache.GetStats(ctx)
		if err != nil {
			mc.logger.Error("Failed to get cache stats",
				zap.String("cache", cacheName),
				zap.Error(err))
			totalErrors++
			continue
		}

		// Calculate cache-specific metrics
		memoryUsage := cache.GetMemoryUsage()
		hitRate := 0.0
		errorRate := 0.0

		totalRequests := stats.HitCount + stats.MissCount
		if totalRequests > 0 {
			hitRate = float64(stats.HitCount) / float64(totalRequests)
		}

		if totalRequests+totalErrors > 0 {
			errorRate = float64(totalErrors) / float64(totalRequests+totalErrors)
		}

		// Store cache metrics
		cacheMetrics[cacheName] = &DetailedCacheMetrics{
			CacheName:           cacheName,
			HitCount:            stats.HitCount,
			MissCount:           stats.MissCount,
			ErrorCount:          totalErrors,
			Size:                stats.Size,
			MemoryUsage:         memoryUsage,
			HitRate:             hitRate,
			ErrorRate:           errorRate,
			AverageResponseTime: 0, // Would need to track this separately
			LastUpdated:         time.Now(),
		}

		// Aggregate metrics
		totalHits += stats.HitCount
		totalMisses += stats.MissCount
		totalSize += stats.Size
		totalMemoryUsage += memoryUsage
	}

	// Calculate aggregated rates
	totalRequests := totalHits + totalMisses
	hitRate := 0.0
	errorRate := 0.0

	if totalRequests > 0 {
		hitRate = float64(totalHits) / float64(totalRequests)
	}

	if totalRequests+totalErrors > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests+totalErrors)
	}

	// Calculate performance metrics
	collectionDuration := time.Since(start)
	throughput := 0.0
	if collectionDuration > 0 {
		throughput = float64(totalRequests) / collectionDuration.Seconds()
	}

	performanceMetrics := &PerformanceMetrics{
		TotalRequests:      totalRequests,
		SuccessfulRequests: totalHits,
		FailedRequests:     totalMisses + totalErrors,
		AverageLatency:     collectionDuration,
		P95Latency:         0, // Would need to track latency percentiles
		P99Latency:         0, // Would need to track latency percentiles
		Throughput:         throughput,
	}

	// Update aggregated metrics
	mc.metricsLock.Lock()
	mc.metrics = &AggregatedMetrics{
		Timestamp:           time.Now(),
		TotalHits:           totalHits,
		TotalMisses:         totalMisses,
		TotalErrors:         totalErrors,
		TotalSize:           totalSize,
		TotalMemoryUsage:    totalMemoryUsage,
		AverageHitRate:      hitRate,
		AverageErrorRate:    errorRate,
		AverageResponseTime: collectionDuration,
		CacheMetrics:        cacheMetrics,
		PerformanceMetrics:  performanceMetrics,
	}
	mc.metricsLock.Unlock()

	// Store historical data
	mc.storeHistoricalData()

	mc.logger.Debug("Updated cache metrics",
		zap.Int64("total_hits", totalHits),
		zap.Int64("total_misses", totalMisses),
		zap.Float64("hit_rate", hitRate),
		zap.Int64("total_size", totalSize),
		zap.Duration("collection_duration", collectionDuration))
}

// storeHistoricalData stores the current metrics as historical data
func (mc *MetricsCollector) storeHistoricalData() {
	mc.historyLock.Lock()
	defer mc.historyLock.Unlock()

	// Create snapshot
	snapshot := &MetricsSnapshot{
		Timestamp: time.Now(),
		Metrics:   mc.GetCurrentMetrics(),
	}

	// Add to historical data
	mc.historicalData = append(mc.historicalData, snapshot)

	// Trim if we exceed max entries
	if len(mc.historicalData) > mc.config.MaxHistoryEntries {
		mc.historicalData = mc.historicalData[1:]
	}
}

// cleanupHistory removes old historical data
func (mc *MetricsCollector) cleanupHistory() {
	ticker := time.NewTicker(1 * time.Hour) // Cleanup every hour
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			mc.cleanupOldHistory()
		case <-mc.stopChan:
			return
		}
	}
}

// cleanupOldHistory removes historical data older than the retention period
func (mc *MetricsCollector) cleanupOldHistory() {
	mc.historyLock.Lock()
	defer mc.historyLock.Unlock()

	cutoff := time.Now().Add(-mc.config.HistoryRetention)
	var newData []*MetricsSnapshot

	for _, snapshot := range mc.historicalData {
		if snapshot.Timestamp.After(cutoff) {
			newData = append(newData, snapshot)
		}
	}

	removedCount := len(mc.historicalData) - len(newData)
	mc.historicalData = newData

	if removedCount > 0 {
		mc.logger.Debug("Cleaned up old historical data",
			zap.Int("removed_entries", removedCount),
			zap.Int("remaining_entries", len(mc.historicalData)))
	}
}

// extractMetricValue extracts a specific metric value from aggregated metrics
func (mc *MetricsCollector) extractMetricValue(metrics *AggregatedMetrics, metricName string) (float64, error) {
	switch metricName {
	case "hit_rate":
		return metrics.AverageHitRate, nil
	case "error_rate":
		return metrics.AverageErrorRate, nil
	case "total_hits":
		return float64(metrics.TotalHits), nil
	case "total_misses":
		return float64(metrics.TotalMisses), nil
	case "total_size":
		return float64(metrics.TotalSize), nil
	case "memory_usage":
		return float64(metrics.TotalMemoryUsage), nil
	case "throughput":
		if metrics.PerformanceMetrics != nil {
			return metrics.PerformanceMetrics.Throughput, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unknown metric: %s", metricName)
	}
}

// IsRunning returns whether the metrics collector is currently running
func (mc *MetricsCollector) IsRunning() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.running
}

// GetConfig returns the metrics configuration
func (mc *MetricsCollector) GetConfig() *MetricsConfig {
	return mc.config
}

// UpdateConfig updates the metrics configuration
func (mc *MetricsCollector) UpdateConfig(config *MetricsConfig) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.config = config
	mc.logger.Info("Updated metrics collector configuration",
		zap.Duration("collection_interval", config.CollectionInterval),
		zap.Duration("history_retention", config.HistoryRetention))
}
