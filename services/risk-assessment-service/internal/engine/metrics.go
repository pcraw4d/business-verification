package engine

import (
	"fmt"
	"sync"
	"time"
)

// Metrics collects and tracks performance metrics
type Metrics struct {
	mu sync.RWMutex

	// Request metrics
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	TotalDuration      time.Duration `json:"total_duration"`
	MinDuration        time.Duration `json:"min_duration"`
	MaxDuration        time.Duration `json:"max_duration"`
	AvgDuration        time.Duration `json:"avg_duration"`

	// Cache metrics
	CacheHits   int64 `json:"cache_hits"`
	CacheMisses int64 `json:"cache_misses"`

	// Batch metrics
	TotalBatches    int64         `json:"total_batches"`
	TotalBatchItems int64         `json:"total_batch_items"`
	BatchDuration   time.Duration `json:"batch_duration"`

	// Error metrics
	ErrorCounts map[string]int64 `json:"error_counts"`

	// Performance metrics
	RequestsPerSecond float64   `json:"requests_per_second"`
	LastResetTime     time.Time `json:"last_reset_time"`
}

// NewMetrics creates a new metrics collector
func NewMetrics() *Metrics {
	return &Metrics{
		MinDuration:   time.Hour, // Initialize with a large value
		ErrorCounts:   make(map[string]int64),
		LastResetTime: time.Now(),
	}
}

// RecordRequest records a request metric
func (m *Metrics) RecordRequest(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.SuccessfulRequests++
	m.TotalDuration += duration

	// Update min/max duration
	if duration < m.MinDuration {
		m.MinDuration = duration
	}
	if duration > m.MaxDuration {
		m.MaxDuration = duration
	}

	// Calculate average duration
	if m.TotalRequests > 0 {
		m.AvgDuration = m.TotalDuration / time.Duration(m.TotalRequests)
	}

	// Calculate requests per second
	elapsed := time.Since(m.LastResetTime)
	if elapsed > 0 {
		m.RequestsPerSecond = float64(m.TotalRequests) / elapsed.Seconds()
	}
}

// RecordError records an error metric
func (m *Metrics) RecordError() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.FailedRequests++
}

// RecordErrorWithType records an error with a specific type
func (m *Metrics) RecordErrorWithType(errorType string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests++
	m.FailedRequests++
	m.ErrorCounts[errorType]++
}

// RecordCacheHit records a cache hit
func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheHits++
}

// RecordCacheMiss records a cache miss
func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.CacheMisses++
}

// RecordBatchRequest records a batch request metric
func (m *Metrics) RecordBatchRequest(duration time.Duration, itemCount int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalBatches++
	m.TotalBatchItems += int64(itemCount)
	m.BatchDuration += duration
}

// GetStats returns current metrics statistics
func (m *Metrics) GetStats() MetricsStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Calculate cache hit rate
	var cacheHitRate float64
	totalCacheRequests := m.CacheHits + m.CacheMisses
	if totalCacheRequests > 0 {
		cacheHitRate = float64(m.CacheHits) / float64(totalCacheRequests)
	}

	// Calculate success rate
	var successRate float64
	if m.TotalRequests > 0 {
		successRate = float64(m.SuccessfulRequests) / float64(m.TotalRequests)
	}

	// Calculate average batch size
	var avgBatchSize float64
	if m.TotalBatches > 0 {
		avgBatchSize = float64(m.TotalBatchItems) / float64(m.TotalBatches)
	}

	// Calculate average batch duration
	var avgBatchDuration time.Duration
	if m.TotalBatches > 0 {
		avgBatchDuration = m.BatchDuration / time.Duration(m.TotalBatches)
	}

	return MetricsStats{
		TotalRequests:      m.TotalRequests,
		SuccessfulRequests: m.SuccessfulRequests,
		FailedRequests:     m.FailedRequests,
		SuccessRate:        successRate,
		TotalDuration:      m.TotalDuration,
		MinDuration:        m.MinDuration,
		MaxDuration:        m.MaxDuration,
		AvgDuration:        m.AvgDuration,
		CacheHits:          m.CacheHits,
		CacheMisses:        m.CacheMisses,
		CacheHitRate:       cacheHitRate,
		TotalBatches:       m.TotalBatches,
		TotalBatchItems:    m.TotalBatchItems,
		AvgBatchSize:       avgBatchSize,
		AvgBatchDuration:   avgBatchDuration,
		RequestsPerSecond:  m.RequestsPerSecond,
		ErrorCounts:        m.copyErrorCounts(),
		LastResetTime:      m.LastResetTime,
	}
}

// copyErrorCounts creates a copy of the error counts map
func (m *Metrics) copyErrorCounts() map[string]int64 {
	copy := make(map[string]int64)
	for k, v := range m.ErrorCounts {
		copy[k] = v
	}
	return copy
}

// Reset resets all metrics
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests = 0
	m.SuccessfulRequests = 0
	m.FailedRequests = 0
	m.TotalDuration = 0
	m.MinDuration = time.Hour
	m.MaxDuration = 0
	m.AvgDuration = 0
	m.CacheHits = 0
	m.CacheMisses = 0
	m.TotalBatches = 0
	m.TotalBatchItems = 0
	m.BatchDuration = 0
	m.RequestsPerSecond = 0
	m.ErrorCounts = make(map[string]int64)
	m.LastResetTime = time.Now()
}

// MetricsStats holds comprehensive metrics statistics
type MetricsStats struct {
	TotalRequests      int64            `json:"total_requests"`
	SuccessfulRequests int64            `json:"successful_requests"`
	FailedRequests     int64            `json:"failed_requests"`
	SuccessRate        float64          `json:"success_rate"`
	TotalDuration      time.Duration    `json:"total_duration"`
	MinDuration        time.Duration    `json:"min_duration"`
	MaxDuration        time.Duration    `json:"max_duration"`
	AvgDuration        time.Duration    `json:"avg_duration"`
	CacheHits          int64            `json:"cache_hits"`
	CacheMisses        int64            `json:"cache_misses"`
	CacheHitRate       float64          `json:"cache_hit_rate"`
	TotalBatches       int64            `json:"total_batches"`
	TotalBatchItems    int64            `json:"total_batch_items"`
	AvgBatchSize       float64          `json:"avg_batch_size"`
	AvgBatchDuration   time.Duration    `json:"avg_batch_duration"`
	RequestsPerSecond  float64          `json:"requests_per_second"`
	ErrorCounts        map[string]int64 `json:"error_counts"`
	LastResetTime      time.Time        `json:"last_reset_time"`
}

// PerformanceThresholds defines performance thresholds
type PerformanceThresholds struct {
	MaxResponseTime      time.Duration `json:"max_response_time"`
	MinSuccessRate       float64       `json:"min_success_rate"`
	MinCacheHitRate      float64       `json:"min_cache_hit_rate"`
	MaxErrorRate         float64       `json:"max_error_rate"`
	MinRequestsPerSecond float64       `json:"min_requests_per_second"`
}

// DefaultPerformanceThresholds returns default performance thresholds
func DefaultPerformanceThresholds() *PerformanceThresholds {
	return &PerformanceThresholds{
		MaxResponseTime:      1 * time.Second, // Sub-1-second target
		MinSuccessRate:       0.95,            // 95% success rate
		MinCacheHitRate:      0.8,             // 80% cache hit rate
		MaxErrorRate:         0.05,            // 5% error rate
		MinRequestsPerSecond: 100,             // 100 requests per second
	}
}

// CheckPerformance checks if performance meets thresholds
func (m *Metrics) CheckPerformance(thresholds *PerformanceThresholds) PerformanceCheck {
	stats := m.GetStats()

	check := PerformanceCheck{
		Passed: true,
		Checks: make(map[string]bool),
		Issues: make([]string, 0),
	}

	// Check response time
	if stats.AvgDuration > thresholds.MaxResponseTime {
		check.Passed = false
		check.Checks["response_time"] = false
		check.Issues = append(check.Issues,
			fmt.Sprintf("Average response time %.2fms exceeds threshold %.2fms",
				float64(stats.AvgDuration.Nanoseconds())/1e6,
				float64(thresholds.MaxResponseTime.Nanoseconds())/1e6))
	} else {
		check.Checks["response_time"] = true
	}

	// Check success rate
	if stats.SuccessRate < thresholds.MinSuccessRate {
		check.Passed = false
		check.Checks["success_rate"] = false
		check.Issues = append(check.Issues,
			fmt.Sprintf("Success rate %.2f%% below threshold %.2f%%",
				stats.SuccessRate*100, thresholds.MinSuccessRate*100))
	} else {
		check.Checks["success_rate"] = true
	}

	// Check cache hit rate
	if stats.CacheHitRate < thresholds.MinCacheHitRate {
		check.Passed = false
		check.Checks["cache_hit_rate"] = false
		check.Issues = append(check.Issues,
			fmt.Sprintf("Cache hit rate %.2f%% below threshold %.2f%%",
				stats.CacheHitRate*100, thresholds.MinCacheHitRate*100))
	} else {
		check.Checks["cache_hit_rate"] = true
	}

	// Check error rate
	errorRate := 1 - stats.SuccessRate
	if errorRate > thresholds.MaxErrorRate {
		check.Passed = false
		check.Checks["error_rate"] = false
		check.Issues = append(check.Issues,
			fmt.Sprintf("Error rate %.2f%% exceeds threshold %.2f%%",
				errorRate*100, thresholds.MaxErrorRate*100))
	} else {
		check.Checks["error_rate"] = true
	}

	// Check requests per second
	if stats.RequestsPerSecond < thresholds.MinRequestsPerSecond {
		check.Passed = false
		check.Checks["requests_per_second"] = false
		check.Issues = append(check.Issues,
			fmt.Sprintf("Requests per second %.2f below threshold %.2f",
				stats.RequestsPerSecond, thresholds.MinRequestsPerSecond))
	} else {
		check.Checks["requests_per_second"] = true
	}

	return check
}

// PerformanceCheck holds the result of a performance check
type PerformanceCheck struct {
	Passed bool            `json:"passed"`
	Checks map[string]bool `json:"checks"`
	Issues []string        `json:"issues"`
}
