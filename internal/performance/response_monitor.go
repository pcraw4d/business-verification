package performance

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ResponseMonitor monitors API response times and performance
type ResponseMonitor struct {
	logger    *zap.Logger
	profiler  *Profiler
	config    *ResponseMonitorConfig
	stats     *ResponseStats
	alerts    []Alert
	mu        sync.RWMutex
	alertChan chan Alert
	stopChan  chan struct{}
}

// ResponseMonitorConfig contains response monitoring configuration
type ResponseMonitorConfig struct {
	P95Threshold     time.Duration `json:"p95_threshold"`
	P99Threshold     time.Duration `json:"p99_threshold"`
	AverageThreshold time.Duration `json:"average_threshold"`
	MaxResponseTime  time.Duration `json:"max_response_time"`
	AlertWindow      time.Duration `json:"alert_window"`
	SampleRate       float64       `json:"sample_rate"`
	EnableAlerts     bool          `json:"enable_alerts"`
	EnableMetrics    bool          `json:"enable_metrics"`
	EnableLogging    bool          `json:"enable_logging"`
}

// ResponseStats contains response time statistics
type ResponseStats struct {
	TotalRequests      int64           `json:"total_requests"`
	SuccessfulRequests int64           `json:"successful_requests"`
	FailedRequests     int64           `json:"failed_requests"`
	AverageTime        time.Duration   `json:"average_time"`
	P50Time            time.Duration   `json:"p50_time"`
	P95Time            time.Duration   `json:"p95_time"`
	P99Time            time.Duration   `json:"p99_time"`
	MaxTime            time.Duration   `json:"max_time"`
	MinTime            time.Duration   `json:"min_time"`
	TotalTime          time.Duration   `json:"total_time"`
	SlowRequests       int64           `json:"slow_requests"`
	LastUpdated        time.Time       `json:"last_updated"`
	ResponseTimes      []time.Duration `json:"-"` // For percentile calculations
}

// Alert represents a performance alert
type Alert struct {
	Type      string        `json:"type"`
	Severity  string        `json:"severity"`
	Message   string        `json:"message"`
	Value     interface{}   `json:"value"`
	Threshold interface{}   `json:"threshold"`
	Timestamp time.Time     `json:"timestamp"`
	Endpoint  string        `json:"endpoint,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
}

// EndpointStats contains statistics for a specific endpoint
type EndpointStats struct {
	Endpoint           string          `json:"endpoint"`
	Method             string          `json:"method"`
	TotalRequests      int64           `json:"total_requests"`
	SuccessfulRequests int64           `json:"successful_requests"`
	FailedRequests     int64           `json:"failed_requests"`
	AverageTime        time.Duration   `json:"average_time"`
	P95Time            time.Duration   `json:"p95_time"`
	P99Time            time.Duration   `json:"p99_time"`
	MaxTime            time.Duration   `json:"max_time"`
	MinTime            time.Duration   `json:"min_time"`
	SlowRequests       int64           `json:"slow_requests"`
	LastUpdated        time.Time       `json:"last_updated"`
	ResponseTimes      []time.Duration `json:"-"`
}

// NewResponseMonitor creates a new response monitor
func NewResponseMonitor(logger *zap.Logger, profiler *Profiler, config *ResponseMonitorConfig) *ResponseMonitor {
	monitor := &ResponseMonitor{
		logger:    logger,
		profiler:  profiler,
		config:    config,
		stats:     &ResponseStats{},
		alerts:    make([]Alert, 0),
		alertChan: make(chan Alert, 100),
		stopChan:  make(chan struct{}),
	}

	// Start monitoring routines
	go monitor.processAlerts()
	go monitor.monitor()

	return monitor
}

// RecordResponse records a response time measurement
func (rm *ResponseMonitor) RecordResponse(endpoint, method string, duration time.Duration, success bool) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Update overall statistics
	rm.stats.TotalRequests++
	if success {
		rm.stats.SuccessfulRequests++
	} else {
		rm.stats.FailedRequests++
	}

	rm.stats.TotalTime += duration
	rm.stats.AverageTime = rm.stats.TotalTime / time.Duration(rm.stats.TotalRequests)

	// Update min/max
	if rm.stats.TotalRequests == 1 {
		rm.stats.MinTime = duration
		rm.stats.MaxTime = duration
	} else {
		if duration < rm.stats.MinTime {
			rm.stats.MinTime = duration
		}
		if duration > rm.stats.MaxTime {
			rm.stats.MaxTime = duration
		}
	}

	// Add to response times for percentile calculations
	rm.stats.ResponseTimes = append(rm.stats.ResponseTimes, duration)

	// Keep only last 1000 measurements for memory efficiency
	if len(rm.stats.ResponseTimes) > 1000 {
		rm.stats.ResponseTimes = rm.stats.ResponseTimes[len(rm.stats.ResponseTimes)-1000:]
	}

	// Calculate percentiles
	if len(rm.stats.ResponseTimes) >= 10 {
		rm.stats.P50Time = rm.calculatePercentile(rm.stats.ResponseTimes, 0.50)
		rm.stats.P95Time = rm.calculatePercentile(rm.stats.ResponseTimes, 0.95)
		rm.stats.P99Time = rm.calculatePercentile(rm.stats.ResponseTimes, 0.99)
	}

	// Check for slow requests
	if duration > rm.config.P95Threshold {
		rm.stats.SlowRequests++
	}

	rm.stats.LastUpdated = time.Now()

	// Record in profiler
	if rm.profiler != nil {
		rm.profiler.RecordMetric("response_time", duration)
		rm.profiler.RecordMetric(fmt.Sprintf("endpoint_%s_%s", method, endpoint), duration)
	}

	// Check for alerts
	rm.checkAlerts(endpoint, method, duration, success)

	// Log slow requests
	if rm.config.EnableLogging && duration > rm.config.P95Threshold {
		rm.logger.Warn("Slow request detected",
			zap.String("endpoint", endpoint),
			zap.String("method", method),
			zap.Duration("duration", duration),
			zap.Duration("threshold", rm.config.P95Threshold),
			zap.Bool("success", success))
	}
}

// calculatePercentile calculates the nth percentile of response times
func (rm *ResponseMonitor) calculatePercentile(times []time.Duration, percentile float64) time.Duration {
	if len(times) == 0 {
		return 0
	}

	// Sort times
	sorted := make([]time.Duration, len(times))
	copy(sorted, times)

	// Simple bubble sort for small arrays
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

// checkAlerts checks for performance alerts
func (rm *ResponseMonitor) checkAlerts(endpoint, method string, duration time.Duration, success bool) {
	if !rm.config.EnableAlerts {
		return
	}

	// Check P95 threshold
	if duration > rm.config.P95Threshold {
		alert := Alert{
			Type:      "slow_response",
			Severity:  "warning",
			Message:   fmt.Sprintf("Response time exceeded P95 threshold"),
			Value:     duration,
			Threshold: rm.config.P95Threshold,
			Timestamp: time.Now(),
			Endpoint:  endpoint,
			Duration:  duration,
		}
		rm.sendAlert(alert)
	}

	// Check P99 threshold
	if duration > rm.config.P99Threshold {
		alert := Alert{
			Type:      "very_slow_response",
			Severity:  "critical",
			Message:   fmt.Sprintf("Response time exceeded P99 threshold"),
			Value:     duration,
			Threshold: rm.config.P99Threshold,
			Timestamp: time.Now(),
			Endpoint:  endpoint,
			Duration:  duration,
		}
		rm.sendAlert(alert)
	}

	// Check max response time
	if duration > rm.config.MaxResponseTime {
		alert := Alert{
			Type:      "max_response_time_exceeded",
			Severity:  "critical",
			Message:   fmt.Sprintf("Response time exceeded maximum allowed time"),
			Value:     duration,
			Threshold: rm.config.MaxResponseTime,
			Timestamp: time.Now(),
			Endpoint:  endpoint,
			Duration:  duration,
		}
		rm.sendAlert(alert)
	}

	// Check average response time
	if rm.stats.TotalRequests > 100 && rm.stats.AverageTime > rm.config.AverageThreshold {
		alert := Alert{
			Type:      "high_average_response_time",
			Severity:  "warning",
			Message:   fmt.Sprintf("Average response time is high"),
			Value:     rm.stats.AverageTime,
			Threshold: rm.config.AverageThreshold,
			Timestamp: time.Now(),
		}
		rm.sendAlert(alert)
	}
}

// sendAlert sends an alert
func (rm *ResponseMonitor) sendAlert(alert Alert) {
	select {
	case rm.alertChan <- alert:
		// Alert sent successfully
	default:
		// Alert channel is full, log the alert
		rm.logger.Warn("Alert channel full, dropping alert",
			zap.String("type", alert.Type),
			zap.String("severity", alert.Severity))
	}
}

// processAlerts processes incoming alerts
func (rm *ResponseMonitor) processAlerts() {
	for {
		select {
		case alert := <-rm.alertChan:
			rm.mu.Lock()
			rm.alerts = append(rm.alerts, alert)

			// Keep only last 1000 alerts
			if len(rm.alerts) > 1000 {
				rm.alerts = rm.alerts[len(rm.alerts)-1000:]
			}
			rm.mu.Unlock()

			// Log alert
			rm.logger.Warn("Performance alert triggered",
				zap.String("type", alert.Type),
				zap.String("severity", alert.Severity),
				zap.String("message", alert.Message),
				zap.Any("value", alert.Value),
				zap.Any("threshold", alert.Threshold),
				zap.String("endpoint", alert.Endpoint),
				zap.Duration("duration", alert.Duration))

		case <-rm.stopChan:
			return
		}
	}
}

// monitor performs periodic monitoring
func (rm *ResponseMonitor) monitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rm.logPerformanceSummary()
		case <-rm.stopChan:
			return
		}
	}
}

// logPerformanceSummary logs a performance summary
func (rm *ResponseMonitor) logPerformanceSummary() {
	rm.mu.RLock()
	stats := *rm.stats
	rm.mu.RUnlock()

	if stats.TotalRequests == 0 {
		return
	}

	rm.logger.Info("Performance summary",
		zap.Int64("total_requests", stats.TotalRequests),
		zap.Int64("successful_requests", stats.SuccessfulRequests),
		zap.Int64("failed_requests", stats.FailedRequests),
		zap.Duration("average_time", stats.AverageTime),
		zap.Duration("p95_time", stats.P95Time),
		zap.Duration("p99_time", stats.P99Time),
		zap.Duration("max_time", stats.MaxTime),
		zap.Int64("slow_requests", stats.SlowRequests))

	// Check if we're meeting our performance targets
	if stats.P95Time > rm.config.P95Threshold {
		rm.logger.Warn("P95 response time target not met",
			zap.Duration("p95_time", stats.P95Time),
			zap.Duration("target", rm.config.P95Threshold))
	}
}

// GetStats returns current response statistics
func (rm *ResponseMonitor) GetStats() *ResponseStats {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a copy
	stats := *rm.stats
	return &stats
}

// GetAlerts returns recent alerts
func (rm *ResponseMonitor) GetAlerts(limit int) []Alert {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if limit <= 0 || limit > len(rm.alerts) {
		limit = len(rm.alerts)
	}

	// Return most recent alerts
	start := len(rm.alerts) - limit
	if start < 0 {
		start = 0
	}

	alerts := make([]Alert, limit)
	copy(alerts, rm.alerts[start:])
	return alerts
}

// ClearAlerts clears all alerts
func (rm *ResponseMonitor) ClearAlerts() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.alerts = make([]Alert, 0)
	rm.logger.Info("Alerts cleared")
}

// Reset resets all statistics
func (rm *ResponseMonitor) Reset() {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.stats = &ResponseStats{}
	rm.alerts = make([]Alert, 0)
	rm.logger.Info("Response monitor reset")
}

// Stop stops the response monitor
func (rm *ResponseMonitor) Stop() {
	close(rm.stopChan)
	rm.logger.Info("Response monitor stopped")
}

// IsHealthy checks if the system is healthy based on response times
func (rm *ResponseMonitor) IsHealthy() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// System is healthy if:
	// 1. P95 response time is below threshold
	// 2. P99 response time is below threshold
	// 3. Average response time is below threshold
	// 4. We have enough data points

	if rm.stats.TotalRequests < 10 {
		return true // Not enough data yet
	}

	return rm.stats.P95Time <= rm.config.P95Threshold &&
		rm.stats.P99Time <= rm.config.P99Threshold &&
		rm.stats.AverageTime <= rm.config.AverageThreshold
}

// GetHealthScore returns a health score from 0-100
func (rm *ResponseMonitor) GetHealthScore() float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.stats.TotalRequests < 10 {
		return 100.0 // Not enough data yet
	}

	score := 100.0

	// Deduct points for exceeding thresholds
	if rm.stats.P95Time > rm.config.P95Threshold {
		excess := float64(rm.stats.P95Time-rm.config.P95Threshold) / float64(rm.config.P95Threshold)
		score -= excess * 20 // Up to 20 points deduction
	}

	if rm.stats.P99Time > rm.config.P99Threshold {
		excess := float64(rm.stats.P99Time-rm.config.P99Threshold) / float64(rm.config.P99Threshold)
		score -= excess * 30 // Up to 30 points deduction
	}

	if rm.stats.AverageTime > rm.config.AverageThreshold {
		excess := float64(rm.stats.AverageTime-rm.config.AverageThreshold) / float64(rm.config.AverageThreshold)
		score -= excess * 25 // Up to 25 points deduction
	}

	// Deduct points for high error rate
	if rm.stats.TotalRequests > 0 {
		errorRate := float64(rm.stats.FailedRequests) / float64(rm.stats.TotalRequests)
		score -= errorRate * 25 // Up to 25 points deduction
	}

	if score < 0 {
		score = 0
	}

	return score
}

// GetPerformanceReport generates a performance report
func (rm *ResponseMonitor) GetPerformanceReport() string {
	stats := rm.GetStats()
	alerts := rm.GetAlerts(10)
	healthScore := rm.GetHealthScore()

	report := fmt.Sprintf("=== RESPONSE TIME PERFORMANCE REPORT ===\n")
	report += fmt.Sprintf("Generated: %s\n", stats.LastUpdated.Format(time.RFC3339))
	report += fmt.Sprintf("Health Score: %.1f/100\n", healthScore)
	report += fmt.Sprintf("Total Requests: %d\n", stats.TotalRequests)
	report += fmt.Sprintf("Successful Requests: %d\n", stats.SuccessfulRequests)
	report += fmt.Sprintf("Failed Requests: %d\n", stats.FailedRequests)
	report += fmt.Sprintf("Average Response Time: %v\n", stats.AverageTime)
	report += fmt.Sprintf("P50 Response Time: %v\n", stats.P50Time)
	report += fmt.Sprintf("P95 Response Time: %v\n", stats.P95Time)
	report += fmt.Sprintf("P99 Response Time: %v\n", stats.P99Time)
	report += fmt.Sprintf("Min Response Time: %v\n", stats.MinTime)
	report += fmt.Sprintf("Max Response Time: %v\n", stats.MaxTime)
	report += fmt.Sprintf("Slow Requests: %d\n", stats.SlowRequests)

	if stats.TotalRequests > 0 {
		errorRate := float64(stats.FailedRequests) / float64(stats.TotalRequests) * 100
		report += fmt.Sprintf("Error Rate: %.2f%%\n", errorRate)
	}

	report += fmt.Sprintf("\n=== PERFORMANCE TARGETS ===\n")
	report += fmt.Sprintf("P95 Target: %v (Current: %v) - %s\n",
		rm.config.P95Threshold, stats.P95Time,
		func() string {
			if stats.P95Time <= rm.config.P95Threshold {
				return "✅ PASS"
			}
			return "❌ FAIL"
		}())
	report += fmt.Sprintf("P99 Target: %v (Current: %v) - %s\n",
		rm.config.P99Threshold, stats.P99Time,
		func() string {
			if stats.P99Time <= rm.config.P99Threshold {
				return "✅ PASS"
			}
			return "❌ FAIL"
		}())
	report += fmt.Sprintf("Average Target: %v (Current: %v) - %s\n",
		rm.config.AverageThreshold, stats.AverageTime,
		func() string {
			if stats.AverageTime <= rm.config.AverageThreshold {
				return "✅ PASS"
			}
			return "❌ FAIL"
		}())

	report += fmt.Sprintf("\n=== RECENT ALERTS ===\n")
	if len(alerts) == 0 {
		report += "No recent alerts\n"
	} else {
		for i, alert := range alerts {
			report += fmt.Sprintf("%d. [%s] %s - %s\n",
				i+1, alert.Severity, alert.Type, alert.Message)
			report += fmt.Sprintf("   Value: %v, Threshold: %v, Time: %s\n",
				alert.Value, alert.Threshold, alert.Timestamp.Format(time.RFC3339))
		}
	}

	return report
}

// DefaultResponseMonitorConfig returns a default response monitor configuration
func DefaultResponseMonitorConfig() *ResponseMonitorConfig {
	return &ResponseMonitorConfig{
		P95Threshold:     1 * time.Second,
		P99Threshold:     2 * time.Second,
		AverageThreshold: 500 * time.Millisecond,
		MaxResponseTime:  5 * time.Second,
		AlertWindow:      5 * time.Minute,
		SampleRate:       1.0,
		EnableAlerts:     true,
		EnableMetrics:    true,
		EnableLogging:    true,
	}
}
