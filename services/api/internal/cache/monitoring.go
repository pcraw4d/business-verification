package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheMonitoringService provides monitoring and metrics for cache operations
type CacheMonitoringService struct {
	caches []Cache
	config *MonitoringConfig
	logger *zap.Logger

	// Metrics collection
	metrics     *CacheMonitoringMetrics
	metricsLock sync.RWMutex

	// Monitoring control
	stopChan chan struct{}
	running  bool
	mu       sync.RWMutex
}

// MonitoringConfig holds configuration for cache monitoring
type MonitoringConfig struct {
	CollectionInterval time.Duration   `json:"collection_interval"` // How often to collect metrics
	AlertThresholds    AlertThresholds `json:"alert_thresholds"`    // Alert thresholds
	EnableAlerts       bool            `json:"enable_alerts"`       // Whether to enable alerts
}

// AlertThresholds defines thresholds for cache alerts
type AlertThresholds struct {
	HitRateLow      float64 `json:"hit_rate_low"`      // Alert if hit rate below this
	MemoryUsageHigh int64   `json:"memory_usage_high"` // Alert if memory usage above this
	SizeHigh        int64   `json:"size_high"`         // Alert if cache size above this
	ErrorRateHigh   float64 `json:"error_rate_high"`   // Alert if error rate above this
}

// CacheMonitoringMetrics holds aggregated cache metrics
type CacheMonitoringMetrics struct {
	TotalHits        int64     `json:"total_hits"`
	TotalMisses      int64     `json:"total_misses"`
	TotalErrors      int64     `json:"total_errors"`
	TotalSize        int64     `json:"total_size"`
	TotalMemoryUsage int64     `json:"total_memory_usage"`
	AverageHitRate   float64   `json:"average_hit_rate"`
	AverageErrorRate float64   `json:"average_error_rate"`
	LastUpdated      time.Time `json:"last_updated"`

	// Per-cache metrics
	CacheMetrics map[string]*CacheStats `json:"cache_metrics"`
}

// NewCacheMonitoringService creates a new cache monitoring service
func NewCacheMonitoringService(caches []Cache, config *MonitoringConfig, logger *zap.Logger) *CacheMonitoringService {
	if config == nil {
		config = &MonitoringConfig{
			CollectionInterval: 30 * time.Second,
			AlertThresholds: AlertThresholds{
				HitRateLow:      0.7,               // Alert if hit rate below 70%
				MemoryUsageHigh: 100 * 1024 * 1024, // Alert if memory usage above 100MB
				SizeHigh:        10000,             // Alert if cache size above 10k entries
				ErrorRateHigh:   0.05,              // Alert if error rate above 5%
			},
			EnableAlerts: true,
		}
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	return &CacheMonitoringService{
		caches: caches,
		config: config,
		logger: logger,
		metrics: &CacheMonitoringMetrics{
			CacheMetrics: make(map[string]*CacheStats),
		},
		stopChan: make(chan struct{}),
	}
}

// Start begins monitoring cache metrics
func (cms *CacheMonitoringService) Start() error {
	cms.mu.Lock()
	defer cms.mu.Unlock()

	if cms.running {
		return fmt.Errorf("monitor is already running")
	}

	cms.running = true
	cms.stopChan = make(chan struct{})

	cms.logger.Info("Starting cache monitor",
		zap.Duration("collection_interval", cms.config.CollectionInterval),
		zap.Bool("alerts_enabled", cms.config.EnableAlerts))

	// Start metrics collection goroutine
	go cms.collectMetrics()

	return nil
}

// Stop stops monitoring cache metrics
func (cms *CacheMonitoringService) Stop() error {
	cms.mu.Lock()
	defer cms.mu.Unlock()

	if !cms.running {
		return fmt.Errorf("monitor is not running")
	}

	cms.running = false
	close(cms.stopChan)

	cms.logger.Info("Stopped cache monitor")
	return nil
}

// GetMetrics returns current cache metrics
func (cms *CacheMonitoringService) GetMetrics() *CacheMonitoringMetrics {
	cms.metricsLock.RLock()
	defer cms.metricsLock.RUnlock()

	// Return a copy of the metrics
	metrics := *cms.metrics
	metrics.CacheMetrics = make(map[string]*CacheStats)
	for k, v := range cms.metrics.CacheMetrics {
		stats := *v
		metrics.CacheMetrics[k] = &stats
	}

	return &metrics
}

// GetHealthStatus returns the health status of all caches
func (cms *CacheMonitoringService) GetHealthStatus(ctx context.Context) map[string]bool {
	health := make(map[string]bool)

	for i, cache := range cms.caches {
		cacheName := fmt.Sprintf("cache_%d", i)

		// Try to get stats to check if cache is healthy
		_, err := cache.GetStats(ctx)
		health[cacheName] = err == nil

		// For Redis cache, also check connection
		if redisCache, ok := cache.(*RedisCacheImpl); ok {
			err := redisCache.HealthCheck(ctx)
			health[cacheName] = err == nil
		}
	}

	return health
}

// CheckAlerts checks for alert conditions and returns any alerts
func (cms *CacheMonitoringService) CheckAlerts() []Alert {
	cms.metricsLock.RLock()
	defer cms.metricsLock.RUnlock()

	var alerts []Alert

	// Check hit rate
	if cms.metrics.AverageHitRate < cms.config.AlertThresholds.HitRateLow {
		alerts = append(alerts, Alert{
			Type:         "low_hit_rate",
			Severity:     "warning",
			Message:      fmt.Sprintf("Cache hit rate is low: %.2f%%", cms.metrics.AverageHitRate*100),
			Threshold:    cms.config.AlertThresholds.HitRateLow,
			CurrentValue: cms.metrics.AverageHitRate,
			Timestamp:    time.Now(),
		})
	}

	// Check memory usage
	if cms.metrics.TotalMemoryUsage > cms.config.AlertThresholds.MemoryUsageHigh {
		alerts = append(alerts, Alert{
			Type:         "high_memory_usage",
			Severity:     "warning",
			Message:      fmt.Sprintf("Cache memory usage is high: %d bytes", cms.metrics.TotalMemoryUsage),
			Threshold:    float64(cms.config.AlertThresholds.MemoryUsageHigh),
			CurrentValue: float64(cms.metrics.TotalMemoryUsage),
			Timestamp:    time.Now(),
		})
	}

	// Check cache size
	if cms.metrics.TotalSize > cms.config.AlertThresholds.SizeHigh {
		alerts = append(alerts, Alert{
			Type:         "high_cache_size",
			Severity:     "info",
			Message:      fmt.Sprintf("Cache size is high: %d entries", cms.metrics.TotalSize),
			Threshold:    float64(cms.config.AlertThresholds.SizeHigh),
			CurrentValue: float64(cms.metrics.TotalSize),
			Timestamp:    time.Now(),
		})
	}

	// Check error rate
	if cms.metrics.AverageErrorRate > cms.config.AlertThresholds.ErrorRateHigh {
		alerts = append(alerts, Alert{
			Type:         "high_error_rate",
			Severity:     "error",
			Message:      fmt.Sprintf("Cache error rate is high: %.2f%%", cms.metrics.AverageErrorRate*100),
			Threshold:    cms.config.AlertThresholds.ErrorRateHigh,
			CurrentValue: cms.metrics.AverageErrorRate,
			Timestamp:    time.Now(),
		})
	}

	return alerts
}

// collectMetrics collects metrics from all caches
func (cms *CacheMonitoringService) collectMetrics() {
	ticker := time.NewTicker(cms.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cms.updateMetrics()
		case <-cms.stopChan:
			return
		}
	}
}

// updateMetrics updates the aggregated metrics
func (cms *CacheMonitoringService) updateMetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var totalHits, totalMisses, totalErrors, totalSize, totalMemoryUsage int64
	cacheMetrics := make(map[string]*CacheStats)

	for i, cache := range cms.caches {
		cacheName := fmt.Sprintf("cache_%d", i)

		stats, err := cache.GetStats(ctx)
		if err != nil {
			cms.logger.Error("Failed to get cache stats",
				zap.String("cache", cacheName),
				zap.Error(err))
			totalErrors++
			continue
		}

		// Aggregate metrics
		totalHits += stats.HitCount
		totalMisses += stats.MissCount
		totalSize += stats.Size
		totalMemoryUsage += cache.GetMemoryUsage()

		// Store per-cache metrics
		cacheMetrics[cacheName] = stats
	}

	// Calculate rates
	totalRequests := totalHits + totalMisses
	hitRate := 0.0
	errorRate := 0.0

	if totalRequests > 0 {
		hitRate = float64(totalHits) / float64(totalRequests)
	}

	if totalRequests+totalErrors > 0 {
		errorRate = float64(totalErrors) / float64(totalRequests+totalErrors)
	}

	// Update metrics
	cms.metricsLock.Lock()
	cms.metrics.TotalHits = totalHits
	cms.metrics.TotalMisses = totalMisses
	cms.metrics.TotalErrors = totalErrors
	cms.metrics.TotalSize = totalSize
	cms.metrics.TotalMemoryUsage = totalMemoryUsage
	cms.metrics.AverageHitRate = hitRate
	cms.metrics.AverageErrorRate = errorRate
	cms.metrics.LastUpdated = time.Now()
	cms.metrics.CacheMetrics = cacheMetrics
	cms.metricsLock.Unlock()

	// Check for alerts
	if cms.config.EnableAlerts {
		alerts := cms.CheckAlerts()
		for _, alert := range alerts {
			cms.logger.Warn("Cache alert triggered",
				zap.String("type", alert.Type),
				zap.String("severity", alert.Severity),
				zap.String("message", alert.Message))
		}
	}

	cms.logger.Debug("Updated cache metrics",
		zap.Int64("total_hits", totalHits),
		zap.Int64("total_misses", totalMisses),
		zap.Float64("hit_rate", hitRate),
		zap.Int64("total_size", totalSize))
}

// Alert represents a cache alert
type Alert struct {
	Type         string    `json:"type"`
	Severity     string    `json:"severity"` // info, warning, error
	Message      string    `json:"message"`
	Threshold    float64   `json:"threshold"`
	CurrentValue float64   `json:"current_value"`
	Timestamp    time.Time `json:"timestamp"`
}

// IsRunning returns whether the monitor is currently running
func (cms *CacheMonitoringService) IsRunning() bool {
	cms.mu.RLock()
	defer cms.mu.RUnlock()
	return cms.running
}

// GetConfig returns the monitoring configuration
func (cms *CacheMonitoringService) GetConfig() *MonitoringConfig {
	return cms.config
}

// UpdateConfig updates the monitoring configuration
func (cms *CacheMonitoringService) UpdateConfig(config *MonitoringConfig) {
	cms.mu.Lock()
	defer cms.mu.Unlock()

	cms.config = config
	cms.logger.Info("Updated cache monitor configuration",
		zap.Duration("collection_interval", config.CollectionInterval),
		zap.Bool("alerts_enabled", config.EnableAlerts))
}
