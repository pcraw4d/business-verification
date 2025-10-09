package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PerformanceMonitor tracks service performance metrics
type PerformanceMonitor struct {
	logger *zap.Logger
	mu     sync.RWMutex

	// Request metrics
	requestCount    int64
	requestDuration time.Duration
	errorCount      int64

	// Response time metrics
	responseTimes   []time.Duration
	maxResponseTime time.Duration
	minResponseTime time.Duration

	// Throughput metrics
	requestsPerSecond float64
	requestsPerMinute float64

	// Error rate metrics
	errorRate float64

	// System metrics
	memoryUsage    uint64
	cpuUsage       float64
	goroutineCount int

	// Target metrics
	targetRPS        float64       // Target requests per second
	targetLatency    time.Duration // Target response time
	targetErrorRate  float64       // Target error rate
	targetThroughput float64       // Target requests per minute

	// Alerts
	alerts []Alert
}

// Alert represents a performance alert
type Alert struct {
	Type         AlertType
	Message      string
	Timestamp    time.Time
	Severity     AlertSeverity
	Threshold    float64
	CurrentValue float64
}

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeHighLatency    AlertType = "high_latency"
	AlertTypeHighErrorRate  AlertType = "high_error_rate"
	AlertTypeLowThroughput  AlertType = "low_throughput"
	AlertTypeHighMemory     AlertType = "high_memory"
	AlertTypeHighCPU        AlertType = "high_cpu"
	AlertTypeHighGoroutines AlertType = "high_goroutines"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityCritical AlertSeverity = "critical"
)

// PerformanceStats represents current performance statistics
type PerformanceStats struct {
	RequestCount      int64           `json:"request_count"`
	RequestDuration   time.Duration   `json:"request_duration"`
	ErrorCount        int64           `json:"error_count"`
	ResponseTimes     []time.Duration `json:"response_times"`
	MaxResponseTime   time.Duration   `json:"max_response_time"`
	MinResponseTime   time.Duration   `json:"min_response_time"`
	RequestsPerSecond float64         `json:"requests_per_second"`
	RequestsPerMinute float64         `json:"requests_per_minute"`
	ErrorRate         float64         `json:"error_rate"`
	MemoryUsage       uint64          `json:"memory_usage"`
	CPUUsage          float64         `json:"cpu_usage"`
	GoroutineCount    int             `json:"goroutine_count"`
	Alerts            []Alert         `json:"alerts"`
	Timestamp         time.Time       `json:"timestamp"`
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger *zap.Logger) *PerformanceMonitor {
	return &PerformanceMonitor{
		logger:           logger,
		responseTimes:    make([]time.Duration, 0, 1000),
		targetRPS:        16.67, // 1000 requests per minute / 60 seconds
		targetLatency:    1 * time.Second,
		targetErrorRate:  0.01, // 1%
		targetThroughput: 1000, // 1000 requests per minute
		alerts:           make([]Alert, 0),
	}
}

// RecordRequest records a request and its duration
func (pm *PerformanceMonitor) RecordRequest(duration time.Duration, isError bool) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.requestCount++
	pm.requestDuration += duration

	if isError {
		pm.errorCount++
	}

	// Track response times (keep last 1000)
	pm.responseTimes = append(pm.responseTimes, duration)
	if len(pm.responseTimes) > 1000 {
		pm.responseTimes = pm.responseTimes[1:]
	}

	// Update min/max response times
	if pm.maxResponseTime == 0 || duration > pm.maxResponseTime {
		pm.maxResponseTime = duration
	}
	if pm.minResponseTime == 0 || duration < pm.minResponseTime {
		pm.minResponseTime = duration
	}

	// Calculate metrics
	pm.calculateMetrics()
	pm.checkAlerts()
}

// calculateMetrics calculates current performance metrics
func (pm *PerformanceMonitor) calculateMetrics() {
	if pm.requestCount == 0 {
		return
	}

	// Calculate error rate
	pm.errorRate = float64(pm.errorCount) / float64(pm.requestCount)

	// Calculate average response time
	if len(pm.responseTimes) > 0 {
		var totalDuration time.Duration
		for _, duration := range pm.responseTimes {
			totalDuration += duration
		}
		avgResponseTime := totalDuration / time.Duration(len(pm.responseTimes))

		// Calculate RPS based on average response time
		if avgResponseTime > 0 {
			pm.requestsPerSecond = float64(time.Second) / float64(avgResponseTime)
		}
	}

	// Calculate requests per minute
	pm.requestsPerMinute = pm.requestsPerSecond * 60

	// Update system metrics
	pm.updateSystemMetrics()
}

// updateSystemMetrics updates system resource usage
func (pm *PerformanceMonitor) updateSystemMetrics() {
	// In a real implementation, you would use runtime.MemStats and other system APIs
	// For now, we'll use mock values
	pm.memoryUsage = 1024 * 1024 * 100 // 100MB
	pm.cpuUsage = 25.5                 // 25.5%
	pm.goroutineCount = 50
}

// checkAlerts checks for performance alerts
func (pm *PerformanceMonitor) checkAlerts() {
	// Check latency alert
	if pm.maxResponseTime > pm.targetLatency {
		pm.addAlert(AlertTypeHighLatency, fmt.Sprintf("High latency detected: %v (target: %v)", pm.maxResponseTime, pm.targetLatency), AlertSeverityWarning, float64(pm.targetLatency), float64(pm.maxResponseTime))
	}

	// Check error rate alert
	if pm.errorRate > pm.targetErrorRate {
		pm.addAlert(AlertTypeHighErrorRate, fmt.Sprintf("High error rate detected: %.2f%% (target: %.2f%%)", pm.errorRate*100, pm.targetErrorRate*100), AlertSeverityCritical, pm.targetErrorRate, pm.errorRate)
	}

	// Check throughput alert
	if pm.requestsPerMinute < pm.targetThroughput {
		pm.addAlert(AlertTypeLowThroughput, fmt.Sprintf("Low throughput detected: %.2f req/min (target: %.2f req/min)", pm.requestsPerMinute, pm.targetThroughput), AlertSeverityWarning, pm.targetThroughput, pm.requestsPerMinute)
	}

	// Check memory alert
	if pm.memoryUsage > 1024*1024*500 { // 500MB
		pm.addAlert(AlertTypeHighMemory, fmt.Sprintf("High memory usage detected: %d MB", pm.memoryUsage/(1024*1024)), AlertSeverityWarning, 500, float64(pm.memoryUsage/(1024*1024)))
	}

	// Check CPU alert
	if pm.cpuUsage > 80.0 { // 80%
		pm.addAlert(AlertTypeHighCPU, fmt.Sprintf("High CPU usage detected: %.2f%%", pm.cpuUsage), AlertSeverityWarning, 80.0, pm.cpuUsage)
	}

	// Check goroutine alert
	if pm.goroutineCount > 1000 {
		pm.addAlert(AlertTypeHighGoroutines, fmt.Sprintf("High goroutine count detected: %d", pm.goroutineCount), AlertSeverityWarning, 1000, float64(pm.goroutineCount))
	}
}

// addAlert adds a new alert
func (pm *PerformanceMonitor) addAlert(alertType AlertType, message string, severity AlertSeverity, threshold, currentValue float64) {
	alert := Alert{
		Type:         alertType,
		Message:      message,
		Timestamp:    time.Now(),
		Severity:     severity,
		Threshold:    threshold,
		CurrentValue: currentValue,
	}

	pm.alerts = append(pm.alerts, alert)

	// Keep only last 100 alerts
	if len(pm.alerts) > 100 {
		pm.alerts = pm.alerts[1:]
	}

	// Log alert
	pm.logger.Warn("Performance alert triggered",
		zap.String("type", string(alertType)),
		zap.String("message", message),
		zap.String("severity", string(severity)),
		zap.Float64("threshold", threshold),
		zap.Float64("current_value", currentValue))
}

// GetStats returns current performance statistics
func (pm *PerformanceMonitor) GetStats() *PerformanceStats {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return &PerformanceStats{
		RequestCount:      pm.requestCount,
		RequestDuration:   pm.requestDuration,
		ErrorCount:        pm.errorCount,
		ResponseTimes:     pm.responseTimes,
		MaxResponseTime:   pm.maxResponseTime,
		MinResponseTime:   pm.minResponseTime,
		RequestsPerSecond: pm.requestsPerSecond,
		RequestsPerMinute: pm.requestsPerMinute,
		ErrorRate:         pm.errorRate,
		MemoryUsage:       pm.memoryUsage,
		CPUUsage:          pm.cpuUsage,
		GoroutineCount:    pm.goroutineCount,
		Alerts:            pm.alerts,
		Timestamp:         time.Now(),
	}
}

// GetAlerts returns current alerts
func (pm *PerformanceMonitor) GetAlerts() []Alert {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	return pm.alerts
}

// ClearAlerts clears all alerts
func (pm *PerformanceMonitor) ClearAlerts() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.alerts = make([]Alert, 0)
}

// Reset resets all metrics
func (pm *PerformanceMonitor) Reset() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.requestCount = 0
	pm.requestDuration = 0
	pm.errorCount = 0
	pm.responseTimes = make([]time.Duration, 0, 1000)
	pm.maxResponseTime = 0
	pm.minResponseTime = 0
	pm.requestsPerSecond = 0
	pm.requestsPerMinute = 0
	pm.errorRate = 0
	pm.alerts = make([]Alert, 0)
}

// SetTargets sets performance targets
func (pm *PerformanceMonitor) SetTargets(rps float64, latency time.Duration, errorRate float64, throughput float64) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.targetRPS = rps
	pm.targetLatency = latency
	pm.targetErrorRate = errorRate
	pm.targetThroughput = throughput
}

// IsHealthy checks if the service is performing within targets
func (pm *PerformanceMonitor) IsHealthy() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Check if any critical alerts exist
	for _, alert := range pm.alerts {
		if alert.Severity == AlertSeverityCritical {
			return false
		}
	}

	// Check if performance is within targets
	if pm.errorRate > pm.targetErrorRate {
		return false
	}

	if pm.maxResponseTime > pm.targetLatency {
		return false
	}

	if pm.requestsPerMinute < pm.targetThroughput*0.8 { // Allow 20% below target
		return false
	}

	return true
}

// StartMonitoring starts continuous monitoring
func (pm *PerformanceMonitor) StartMonitoring(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pm.calculateMetrics()
			pm.checkAlerts()
		}
	}
}
