package cache

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CachePerformanceMonitor monitors cache performance metrics
type CachePerformanceMonitor struct {
	// Configuration
	config CacheConfig

	// Performance metrics
	metrics     *CachePerformanceMetrics
	metricsLock sync.RWMutex

	// Monitoring intervals
	monitoringInterval time.Duration

	// Thread safety
	mu sync.RWMutex

	// Control
	stopChannel chan struct{}

	// Logging
	logger *zap.Logger
}

// CachePerformanceMetrics holds cache performance metrics
type CachePerformanceMetrics struct {
	// Latency metrics
	AverageLatency time.Duration `json:"average_latency"`
	P50Latency     time.Duration `json:"p50_latency"`
	P95Latency     time.Duration `json:"p95_latency"`
	P99Latency     time.Duration `json:"p99_latency"`
	MaxLatency     time.Duration `json:"max_latency"`
	MinLatency     time.Duration `json:"min_latency"`

	// Throughput metrics
	RequestsPerSecond  float64 `json:"requests_per_second"`
	TotalRequests      int64   `json:"total_requests"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`

	// Hit rate metrics
	OverallHitRate float64 `json:"overall_hit_rate"`
	MemoryHitRate  float64 `json:"memory_hit_rate"`
	DiskHitRate    float64 `json:"disk_hit_rate"`
	RedisHitRate   float64 `json:"redis_hit_rate"`

	// Resource usage metrics
	MemoryUsage int64 `json:"memory_usage"`
	DiskUsage   int64 `json:"disk_usage"`
	RedisUsage  int64 `json:"redis_usage"`
	TotalUsage  int64 `json:"total_usage"`

	// Error metrics
	ErrorRate   float64 `json:"error_rate"`
	TotalErrors int64   `json:"total_errors"`

	// Cache efficiency metrics
	EvictionRate       float64 `json:"eviction_rate"`
	TotalEvictions     int64   `json:"total_evictions"`
	InvalidationRate   float64 `json:"invalidation_rate"`
	TotalInvalidations int64   `json:"total_invalidations"`

	// Timestamps
	LastUpdated time.Time `json:"last_updated"`
	StartTime   time.Time `json:"start_time"`
}

// NewCachePerformanceMonitor creates a new cache performance monitor
func NewCachePerformanceMonitor(config CacheConfig, logger *zap.Logger) *CachePerformanceMonitor {
	monitoringInterval := config.MetricsInterval
	if monitoringInterval <= 0 {
		monitoringInterval = 1 * time.Minute
	}

	cpm := &CachePerformanceMonitor{
		config:             config,
		metrics:            &CachePerformanceMetrics{StartTime: time.Now()},
		monitoringInterval: monitoringInterval,
		logger:             logger,
		stopChannel:        make(chan struct{}),
	}

	return cpm
}

// Start starts the performance monitoring
func (cpm *CachePerformanceMonitor) Start(ctx context.Context) error {
	cpm.logger.Info("Starting cache performance monitor")

	// Start monitoring goroutine
	go cpm.monitorPerformance(ctx)

	cpm.logger.Info("Cache performance monitor started")
	return nil
}

// Stop stops the performance monitoring
func (cpm *CachePerformanceMonitor) Stop() {
	cpm.logger.Info("Stopping cache performance monitor")
	close(cpm.stopChannel)
}

// RecordLatency records a latency measurement
func (cpm *CachePerformanceMonitor) RecordLatency(operation string, latency time.Duration) {
	cpm.metricsLock.Lock()
	defer cpm.metricsLock.Unlock()

	// Update latency metrics (simplified implementation)
	if cpm.metrics.MaxLatency < latency {
		cpm.metrics.MaxLatency = latency
	}
	if cpm.metrics.MinLatency == 0 || cpm.metrics.MinLatency > latency {
		cpm.metrics.MinLatency = latency
	}

	// Simplified average calculation
	totalRequests := cpm.metrics.TotalRequests
	if totalRequests > 0 {
		totalLatency := cpm.metrics.AverageLatency * time.Duration(totalRequests-1)
		cpm.metrics.AverageLatency = (totalLatency + latency) / time.Duration(totalRequests)
	} else {
		cpm.metrics.AverageLatency = latency
	}

	cpm.metrics.LastUpdated = time.Now()
}

// RecordRequest records a cache request
func (cpm *CachePerformanceMonitor) RecordRequest(success bool) {
	cpm.metricsLock.Lock()
	defer cpm.metricsLock.Unlock()

	cpm.metrics.TotalRequests++
	if success {
		cpm.metrics.SuccessfulRequests++
	} else {
		cpm.metrics.FailedRequests++
	}

	// Update error rate
	if cpm.metrics.TotalRequests > 0 {
		cpm.metrics.ErrorRate = float64(cpm.metrics.FailedRequests) / float64(cpm.metrics.TotalRequests)
	}

	cpm.metrics.LastUpdated = time.Now()
}

// RecordHitRate records hit rate for a cache layer
func (cpm *CachePerformanceMonitor) RecordHitRate(layer string, hitRate float64) {
	cpm.metricsLock.Lock()
	defer cpm.metricsLock.Unlock()

	switch layer {
	case "memory":
		cpm.metrics.MemoryHitRate = hitRate
	case "disk":
		cpm.metrics.DiskHitRate = hitRate
	case "redis":
		cpm.metrics.RedisHitRate = hitRate
	case "overall":
		cpm.metrics.OverallHitRate = hitRate
	}

	cpm.metrics.LastUpdated = time.Now()
}

// RecordUsage records resource usage for a cache layer
func (cpm *CachePerformanceMonitor) RecordUsage(layer string, usage int64) {
	cpm.metricsLock.Lock()
	defer cpm.metricsLock.Unlock()

	switch layer {
	case "memory":
		cpm.metrics.MemoryUsage = usage
	case "disk":
		cpm.metrics.DiskUsage = usage
	case "redis":
		cpm.metrics.RedisUsage = usage
	}

	// Update total usage
	cpm.metrics.TotalUsage = cpm.metrics.MemoryUsage + cpm.metrics.DiskUsage + cpm.metrics.RedisUsage

	cpm.metrics.LastUpdated = time.Now()
}

// RecordEviction records a cache eviction
func (cpm *CachePerformanceMonitor) RecordEviction() {
	cpm.metricsLock.Lock()
	defer cpm.metricsLock.Unlock()

	cpm.metrics.TotalEvictions++

	// Update eviction rate
	if cpm.metrics.TotalRequests > 0 {
		cpm.metrics.EvictionRate = float64(cpm.metrics.TotalEvictions) / float64(cpm.metrics.TotalRequests)
	}

	cpm.metrics.LastUpdated = time.Now()
}

// RecordInvalidation records a cache invalidation
func (cpm *CachePerformanceMonitor) RecordInvalidation() {
	cpm.metricsLock.Lock()
	defer cpm.metricsLock.Unlock()

	cpm.metrics.TotalInvalidations++

	// Update invalidation rate
	if cpm.metrics.TotalRequests > 0 {
		cpm.metrics.InvalidationRate = float64(cpm.metrics.TotalInvalidations) / float64(cpm.metrics.TotalRequests)
	}

	cpm.metrics.LastUpdated = time.Now()
}

// RecordError records a cache error
func (cpm *CachePerformanceMonitor) RecordError() {
	cpm.metricsLock.Lock()
	defer cpm.metricsLock.Unlock()

	cpm.metrics.TotalErrors++

	// Update error rate
	if cpm.metrics.TotalRequests > 0 {
		cpm.metrics.ErrorRate = float64(cpm.metrics.TotalErrors) / float64(cpm.metrics.TotalRequests)
	}

	cpm.metrics.LastUpdated = time.Now()
}

// GetMetrics returns current performance metrics
func (cpm *CachePerformanceMonitor) GetMetrics() *CachePerformanceMetrics {
	cpm.metricsLock.RLock()
	defer cpm.metricsLock.RUnlock()

	metrics := *cpm.metrics
	return &metrics
}

// ResetMetrics resets all performance metrics
func (cpm *CachePerformanceMonitor) ResetMetrics() {
	cpm.metricsLock.Lock()
	defer cpm.metricsLock.Unlock()

	cpm.metrics = &CachePerformanceMetrics{
		StartTime:   time.Now(),
		LastUpdated: time.Now(),
	}
}

// GetPerformanceReport generates a comprehensive performance report
func (cpm *CachePerformanceMonitor) GetPerformanceReport() *CachePerformanceReport {
	cpm.metricsLock.RLock()
	defer cpm.metricsLock.RUnlock()

	report := &CachePerformanceReport{
		Metrics:     *cpm.metrics,
		GeneratedAt: time.Now(),
		Uptime:      time.Since(cpm.metrics.StartTime),
	}

	// Calculate requests per second
	if report.Uptime > 0 {
		report.Metrics.RequestsPerSecond = float64(cpm.metrics.TotalRequests) / report.Uptime.Seconds()
	}

	// Generate recommendations
	report.Recommendations = cpm.generateRecommendations()

	return report
}

// CachePerformanceReport holds a comprehensive performance report
type CachePerformanceReport struct {
	Metrics         CachePerformanceMetrics `json:"metrics"`
	GeneratedAt     time.Time               `json:"generated_at"`
	Uptime          time.Duration           `json:"uptime"`
	Recommendations []string                `json:"recommendations"`
}

// Helper methods

func (cpm *CachePerformanceMonitor) monitorPerformance(ctx context.Context) {
	ticker := time.NewTicker(cpm.monitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cpm.stopChannel:
			return
		case <-ticker.C:
			cpm.collectMetrics(ctx)
		}
	}
}

func (cpm *CachePerformanceMonitor) collectMetrics(ctx context.Context) {
	cpm.logger.Debug("Collecting cache performance metrics")

	// This is where you would collect metrics from cache layers
	// For now, we'll just log that metrics collection is happening

	cpm.logger.Debug("Cache performance metrics collected",
		zap.Duration("average_latency", cpm.metrics.AverageLatency),
		zap.Float64("hit_rate", cpm.metrics.OverallHitRate),
		zap.Int64("total_requests", cpm.metrics.TotalRequests),
		zap.Float64("error_rate", cpm.metrics.ErrorRate))
}

func (cpm *CachePerformanceMonitor) generateRecommendations() []string {
	var recommendations []string

	// Check hit rate
	if cpm.metrics.OverallHitRate < 0.8 {
		recommendations = append(recommendations, "Consider increasing cache size or optimizing cache keys")
	}

	// Check error rate
	if cpm.metrics.ErrorRate > 0.05 {
		recommendations = append(recommendations, "High error rate detected - investigate cache layer issues")
	}

	// Check latency
	if cpm.metrics.AverageLatency > 100*time.Millisecond {
		recommendations = append(recommendations, "High latency detected - consider optimizing cache operations")
	}

	// Check eviction rate
	if cpm.metrics.EvictionRate > 0.1 {
		recommendations = append(recommendations, "High eviction rate - consider increasing cache size")
	}

	// Check resource usage
	if cpm.metrics.TotalUsage > 1024*1024*1024 { // 1GB
		recommendations = append(recommendations, "High cache usage - consider implementing cache eviction policies")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Cache performance is within acceptable parameters")
	}

	return recommendations
}
