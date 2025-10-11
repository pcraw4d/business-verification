package performance

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Profiler provides performance profiling capabilities
type Profiler struct {
	logger    *zap.Logger
	metrics   map[string]*Metric
	mu        sync.RWMutex
	enabled   bool
	threshold time.Duration
}

// Metric tracks performance metrics for a specific operation
type Metric struct {
	Name        string          `json:"name"`
	Count       int64           `json:"count"`
	TotalTime   time.Duration   `json:"total_time"`
	MinTime     time.Duration   `json:"min_time"`
	MaxTime     time.Duration   `json:"max_time"`
	AvgTime     time.Duration   `json:"avg_time"`
	P95Time     time.Duration   `json:"p95_time"`
	P99Time     time.Duration   `json:"p99_time"`
	LastTime    time.Duration   `json:"last_time"`
	LastUpdated time.Time       `json:"last_updated"`
	Times       []time.Duration `json:"-"` // For percentile calculations
	mu          sync.RWMutex
}

// PerformanceStats provides overall performance statistics
type PerformanceStats struct {
	TotalOperations     int64              `json:"total_operations"`
	TotalTime           time.Duration      `json:"total_time"`
	AverageResponseTime time.Duration      `json:"average_response_time"`
	P95ResponseTime     time.Duration      `json:"p95_response_time"`
	P99ResponseTime     time.Duration      `json:"p99_response_time"`
	SlowOperations      int64              `json:"slow_operations"`
	Metrics             map[string]*Metric `json:"metrics"`
	MemoryStats         MemoryStats        `json:"memory_stats"`
	GoroutineCount      int                `json:"goroutine_count"`
	Timestamp           time.Time          `json:"timestamp"`
}

// MemoryStats provides memory usage statistics
type MemoryStats struct {
	Alloc         uint64  `json:"alloc"`
	TotalAlloc    uint64  `json:"total_alloc"`
	Sys           uint64  `json:"sys"`
	Lookups       uint64  `json:"lookups"`
	Mallocs       uint64  `json:"mallocs"`
	Frees         uint64  `json:"frees"`
	HeapAlloc     uint64  `json:"heap_alloc"`
	HeapSys       uint64  `json:"heap_sys"`
	HeapIdle      uint64  `json:"heap_idle"`
	HeapInuse     uint64  `json:"heap_inuse"`
	HeapReleased  uint64  `json:"heap_released"`
	HeapObjects   uint64  `json:"heap_objects"`
	StackInuse    uint64  `json:"stack_inuse"`
	StackSys      uint64  `json:"stack_sys"`
	MSpanInuse    uint64  `json:"mspan_inuse"`
	MSpanSys      uint64  `json:"mspan_sys"`
	MCacheInuse   uint64  `json:"mcache_inuse"`
	MCacheSys     uint64  `json:"mcache_sys"`
	BuckHashSys   uint64  `json:"buck_hash_sys"`
	GCSys         uint64  `json:"gc_sys"`
	OtherSys      uint64  `json:"other_sys"`
	NextGC        uint64  `json:"next_gc"`
	LastGC        uint64  `json:"last_gc"`
	PauseTotalNs  uint64  `json:"pause_total_ns"`
	NumGC         uint32  `json:"num_gc"`
	NumForcedGC   uint32  `json:"num_forced_gc"`
	GCCPUFraction float64 `json:"gc_cpu_fraction"`
}

// NewProfiler creates a new performance profiler
func NewProfiler(logger *zap.Logger, threshold time.Duration) *Profiler {
	return &Profiler{
		logger:    logger,
		metrics:   make(map[string]*Metric),
		enabled:   true,
		threshold: threshold,
	}
}

// StartTimer starts timing an operation
func (p *Profiler) StartTimer(name string) func() {
	if !p.enabled {
		return func() {}
	}

	start := time.Now()
	return func() {
		duration := time.Since(start)
		p.RecordMetric(name, duration)
	}
}

// RecordMetric records a performance metric
func (p *Profiler) RecordMetric(name string, duration time.Duration) {
	if !p.enabled {
		return
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	metric, exists := p.metrics[name]
	if !exists {
		metric = &Metric{
			Name:    name,
			MinTime: duration,
			MaxTime: duration,
			Times:   make([]time.Duration, 0, 1000), // Pre-allocate for efficiency
		}
		p.metrics[name] = metric
	}

	metric.mu.Lock()
	defer metric.mu.Unlock()

	// Update basic statistics
	metric.Count++
	metric.TotalTime += duration
	metric.LastTime = duration
	metric.LastUpdated = time.Now()

	// Update min/max
	if duration < metric.MinTime {
		metric.MinTime = duration
	}
	if duration > metric.MaxTime {
		metric.MaxTime = duration
	}

	// Update average
	metric.AvgTime = metric.TotalTime / time.Duration(metric.Count)

	// Add to times array for percentile calculations
	metric.Times = append(metric.Times, duration)

	// Keep only last 1000 measurements for memory efficiency
	if len(metric.Times) > 1000 {
		metric.Times = metric.Times[len(metric.Times)-1000:]
	}

	// Calculate percentiles
	if len(metric.Times) >= 10 {
		metric.P95Time = p.calculatePercentile(metric.Times, 0.95)
		metric.P99Time = p.calculatePercentile(metric.Times, 0.99)
	}

	// Log slow operations
	if duration > p.threshold {
		p.logger.Warn("Slow operation detected",
			zap.String("operation", name),
			zap.Duration("duration", duration),
			zap.Duration("threshold", p.threshold))
	}
}

// calculatePercentile calculates the nth percentile of a slice of durations
func (p *Profiler) calculatePercentile(times []time.Duration, percentile float64) time.Duration {
	if len(times) == 0 {
		return 0
	}

	// Sort times
	sorted := make([]time.Duration, len(times))
	copy(sorted, times)

	// Simple bubble sort for small arrays (optimized for our use case)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	// Calculate index
	index := int(float64(len(sorted)-1) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// GetStats returns current performance statistics
func (p *Profiler) GetStats() *PerformanceStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := &PerformanceStats{
		Metrics:        make(map[string]*Metric),
		MemoryStats:    p.getMemoryStats(),
		GoroutineCount: runtime.NumGoroutine(),
		Timestamp:      time.Now(),
	}

	// Copy metrics
	for name, metric := range p.metrics {
		metric.mu.RLock()
		stats.Metrics[name] = &Metric{
			Name:        metric.Name,
			Count:       metric.Count,
			TotalTime:   metric.TotalTime,
			MinTime:     metric.MinTime,
			MaxTime:     metric.MaxTime,
			AvgTime:     metric.AvgTime,
			P95Time:     metric.P95Time,
			P99Time:     metric.P99Time,
			LastTime:    metric.LastTime,
			LastUpdated: metric.LastUpdated,
		}
		metric.mu.RUnlock()

		stats.TotalOperations += metric.Count
		stats.TotalTime += metric.TotalTime
	}

	// Calculate overall statistics
	if stats.TotalOperations > 0 {
		stats.AverageResponseTime = stats.TotalTime / time.Duration(stats.TotalOperations)

		// Calculate overall P95 and P99
		allTimes := make([]time.Duration, 0)
		for _, metric := range p.metrics {
			metric.mu.RLock()
			allTimes = append(allTimes, metric.Times...)
			metric.mu.RUnlock()
		}

		if len(allTimes) > 0 {
			stats.P95ResponseTime = p.calculatePercentile(allTimes, 0.95)
			stats.P99ResponseTime = p.calculatePercentile(allTimes, 0.99)
		}
	}

	// Count slow operations
	for _, metric := range p.metrics {
		metric.mu.RLock()
		for _, duration := range metric.Times {
			if duration > p.threshold {
				stats.SlowOperations++
			}
		}
		metric.mu.RUnlock()
	}

	return stats
}

// getMemoryStats returns current memory statistics
func (p *Profiler) getMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		Lookups:       m.Lookups,
		Mallocs:       m.Mallocs,
		Frees:         m.Frees,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MSpanSys:      m.MSpanSys,
		MCacheInuse:   m.MCacheInuse,
		MCacheSys:     m.MCacheSys,
		BuckHashSys:   m.BuckHashSys,
		GCSys:         m.GCSys,
		OtherSys:      m.OtherSys,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		PauseTotalNs:  m.PauseTotalNs,
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
	}
}

// Reset clears all performance metrics
func (p *Profiler) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.metrics = make(map[string]*Metric)

	// Force garbage collection
	runtime.GC()

	p.logger.Info("Performance profiler reset")
}

// SetEnabled enables or disables profiling
func (p *Profiler) SetEnabled(enabled bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.enabled = enabled
	p.logger.Info("Performance profiler enabled", zap.Bool("enabled", enabled))
}

// SetThreshold sets the slow operation threshold
func (p *Profiler) SetThreshold(threshold time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.threshold = threshold
	p.logger.Info("Performance profiler threshold updated", zap.Duration("threshold", threshold))
}

// GetMetric returns a specific metric
func (p *Profiler) GetMetric(name string) *Metric {
	p.mu.RLock()
	defer p.mu.RUnlock()

	metric, exists := p.metrics[name]
	if !exists {
		return nil
	}

	metric.mu.RLock()
	defer metric.mu.RUnlock()

	// Return a copy
	return &Metric{
		Name:        metric.Name,
		Count:       metric.Count,
		TotalTime:   metric.TotalTime,
		MinTime:     metric.MinTime,
		MaxTime:     metric.MaxTime,
		AvgTime:     metric.AvgTime,
		P95Time:     metric.P95Time,
		P99Time:     metric.P99Time,
		LastTime:    metric.LastTime,
		LastUpdated: metric.LastUpdated,
	}
}

// ProfileFunc profiles a function execution
func (p *Profiler) ProfileFunc(name string, fn func() error) error {
	if !p.enabled {
		return fn()
	}

	timer := p.StartTimer(name)
	defer timer()

	return fn()
}

// ProfileFuncWithContext profiles a function execution with context
func (p *Profiler) ProfileFuncWithContext(ctx context.Context, name string, fn func(context.Context) error) error {
	if !p.enabled {
		return fn(ctx)
	}

	timer := p.StartTimer(name)
	defer timer()

	return fn(ctx)
}

// GetSlowOperations returns operations that exceed the threshold
func (p *Profiler) GetSlowOperations() map[string][]time.Duration {
	p.mu.RLock()
	defer p.mu.RUnlock()

	slowOps := make(map[string][]time.Duration)

	for name, metric := range p.metrics {
		metric.mu.RLock()
		slowTimes := make([]time.Duration, 0)
		for _, duration := range metric.Times {
			if duration > p.threshold {
				slowTimes = append(slowTimes, duration)
			}
		}
		metric.mu.RUnlock()

		if len(slowTimes) > 0 {
			slowOps[name] = slowTimes
		}
	}

	return slowOps
}

// GetPerformanceReport generates a comprehensive performance report
func (p *Profiler) GetPerformanceReport() string {
	stats := p.GetStats()

	report := fmt.Sprintf("=== PERFORMANCE REPORT ===\n")
	report += fmt.Sprintf("Generated: %s\n", stats.Timestamp.Format(time.RFC3339))
	report += fmt.Sprintf("Total Operations: %d\n", stats.TotalOperations)
	report += fmt.Sprintf("Total Time: %v\n", stats.TotalTime)
	report += fmt.Sprintf("Average Response Time: %v\n", stats.AverageResponseTime)
	report += fmt.Sprintf("P95 Response Time: %v\n", stats.P95ResponseTime)
	report += fmt.Sprintf("P99 Response Time: %v\n", stats.P99ResponseTime)
	report += fmt.Sprintf("Slow Operations: %d\n", stats.SlowOperations)
	report += fmt.Sprintf("Goroutine Count: %d\n", stats.GoroutineCount)

	report += fmt.Sprintf("\n=== MEMORY STATS ===\n")
	report += fmt.Sprintf("Heap Alloc: %d bytes (%.2f MB)\n", stats.MemoryStats.HeapAlloc, float64(stats.MemoryStats.HeapAlloc)/1024/1024)
	report += fmt.Sprintf("Heap Sys: %d bytes (%.2f MB)\n", stats.MemoryStats.HeapSys, float64(stats.MemoryStats.HeapSys)/1024/1024)
	report += fmt.Sprintf("Heap Objects: %d\n", stats.MemoryStats.HeapObjects)
	report += fmt.Sprintf("GC Count: %d\n", stats.MemoryStats.NumGC)
	report += fmt.Sprintf("GC CPU Fraction: %.4f\n", stats.MemoryStats.GCCPUFraction)

	report += fmt.Sprintf("\n=== OPERATION METRICS ===\n")
	for name, metric := range stats.Metrics {
		report += fmt.Sprintf("%s:\n", name)
		report += fmt.Sprintf("  Count: %d\n", metric.Count)
		report += fmt.Sprintf("  Avg Time: %v\n", metric.AvgTime)
		report += fmt.Sprintf("  Min Time: %v\n", metric.MinTime)
		report += fmt.Sprintf("  Max Time: %v\n", metric.MaxTime)
		report += fmt.Sprintf("  P95 Time: %v\n", metric.P95Time)
		report += fmt.Sprintf("  P99 Time: %v\n", metric.P99Time)
		report += fmt.Sprintf("  Last Time: %v\n", metric.LastTime)
		report += fmt.Sprintf("\n")
	}

	return report
}
