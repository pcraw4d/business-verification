package performance

import (
	"sync"
	"time"
)

// MetricsCollector collects performance metrics
type MetricsCollector struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64

	// Response time metrics
	TotalResponseTime time.Duration
	MinResponseTime   time.Duration
	MaxResponseTime   time.Duration

	// Error metrics
	ErrorCounts map[string]int64

	// Cache metrics
	CacheHits   int64
	CacheMisses int64
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		ErrorCounts:     make(map[string]int64),
		MinResponseTime: time.Hour, // Initialize with high value
	}
}

// RecordRequest records a request metric
func (mc *MetricsCollector) RecordRequest(success bool, responseTime time.Duration) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.TotalRequests++
	if success {
		mc.SuccessfulRequests++
	} else {
		mc.FailedRequests++
	}

	// Update response time metrics
	mc.TotalResponseTime += responseTime
	if responseTime < mc.MinResponseTime {
		mc.MinResponseTime = responseTime
	}
	if responseTime > mc.MaxResponseTime {
		mc.MaxResponseTime = responseTime
	}
}

// RecordError records an error metric
func (mc *MetricsCollector) RecordError(errorType string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.ErrorCounts[errorType]++
}

// RecordCacheHit records a cache hit
func (mc *MetricsCollector) RecordCacheHit() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.CacheHits++
}

// RecordCacheMiss records a cache miss
func (mc *MetricsCollector) RecordCacheMiss() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.CacheMisses++
}

// GetMetrics returns current metrics
func (mc *MetricsCollector) GetMetrics() map[string]interface{} {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	avgResponseTime := time.Duration(0)
	if mc.TotalRequests > 0 {
		avgResponseTime = mc.TotalResponseTime / time.Duration(mc.TotalRequests)
	}

	cacheHitRate := float64(0)
	if mc.CacheHits+mc.CacheMisses > 0 {
		cacheHitRate = float64(mc.CacheHits) / float64(mc.CacheHits+mc.CacheMisses) * 100
	}

	successRate := float64(0)
	if mc.TotalRequests > 0 {
		successRate = float64(mc.SuccessfulRequests) / float64(mc.TotalRequests) * 100
	}

	return map[string]interface{}{
		"requests": map[string]interface{}{
			"total":        mc.TotalRequests,
			"successful":   mc.SuccessfulRequests,
			"failed":       mc.FailedRequests,
			"success_rate": successRate,
		},
		"response_times": map[string]interface{}{
			"average": avgResponseTime.String(),
			"min":     mc.MinResponseTime.String(),
			"max":     mc.MaxResponseTime.String(),
		},
		"cache": map[string]interface{}{
			"hits":     mc.CacheHits,
			"misses":   mc.CacheMisses,
			"hit_rate": cacheHitRate,
		},
		"errors": mc.ErrorCounts,
	}
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.TotalRequests = 0
	mc.SuccessfulRequests = 0
	mc.FailedRequests = 0
	mc.TotalResponseTime = 0
	mc.MinResponseTime = time.Hour
	mc.MaxResponseTime = 0
	mc.ErrorCounts = make(map[string]int64)
	mc.CacheHits = 0
	mc.CacheMisses = 0
}
