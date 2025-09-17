package classification_monitoring

import (
	"sync"
	"time"
)

// PerformanceMonitor monitors performance metrics for accuracy tracking
type PerformanceMonitor struct {
	config          *AdvancedAccuracyConfig
	mu              sync.RWMutex
	processingTimes []time.Duration
	lastUpdate      time.Time
	totalCalls      int64
	errorCount      int64
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(config *AdvancedAccuracyConfig, logger interface{}) *PerformanceMonitor {
	return &PerformanceMonitor{
		config:          config,
		processingTimes: make([]time.Duration, 0),
		lastUpdate:      time.Now(),
	}
}

// RecordProcessingTime records a processing time
func (pm *PerformanceMonitor) RecordProcessingTime(duration time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.processingTimes = append(pm.processingTimes, duration)
	pm.totalCalls++
	pm.lastUpdate = time.Now()

	// Keep only recent processing times (last 1000)
	if len(pm.processingTimes) > 1000 {
		pm.processingTimes = pm.processingTimes[500:]
	}
}

// GetAverageLatency returns the average processing latency
func (pm *PerformanceMonitor) GetAverageLatency() time.Duration {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.processingTimes) == 0 {
		return 0
	}

	total := time.Duration(0)
	for _, duration := range pm.processingTimes {
		total += duration
	}

	return total / time.Duration(len(pm.processingTimes))
}

// GetPerformanceSnapshot returns a performance snapshot
func (pm *PerformanceMonitor) GetPerformanceSnapshot() *PerformanceSnapshot {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return &PerformanceSnapshot{
		AverageLatency: pm.GetAverageLatency(),
		Throughput:     pm.calculateThroughput(),
		ErrorRate:      pm.calculateErrorRate(),
		CPUUsage:       0.0, // Would be populated from system metrics
		MemoryUsage:    0.0, // Would be populated from system metrics
	}
}

// calculateThroughput calculates current throughput
func (pm *PerformanceMonitor) calculateThroughput() float64 {
	// Calculate requests per second based on recent activity
	recentCalls := 0

	// Count calls in the last minute
	// This is a simplified calculation - in reality, you'd track timestamps
	if pm.totalCalls > 0 {
		recentCalls = int(pm.totalCalls / 60) // Rough estimate
	}

	return float64(recentCalls) / 60.0 // per second
}

// calculateErrorRate calculates the error rate
func (pm *PerformanceMonitor) calculateErrorRate() float64 {
	if pm.totalCalls == 0 {
		return 0.0
	}

	return float64(pm.errorCount) / float64(pm.totalCalls)
}
