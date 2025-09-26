package performance

import (
	"context"
	"sync"
	"time"
)

// AutoScaler manages automatic scaling based on metrics
type AutoScaler struct {
	mu sync.RWMutex

	// Configuration
	maxConcurrentRequests int
	currentRequests       int
	scaleUpThreshold      int
	scaleDownThreshold    int

	// Metrics
	requestQueue  []time.Time
	responseTimes []time.Duration
	errorCounts   map[string]int

	// Callbacks
	onScaleUp   func() error
	onScaleDown func() error
}

// NewAutoScaler creates a new auto scaler
func NewAutoScaler(maxConcurrent int, scaleUpThreshold, scaleDownThreshold int) *AutoScaler {
	return &AutoScaler{
		maxConcurrentRequests: maxConcurrent,
		scaleUpThreshold:      scaleUpThreshold,
		scaleDownThreshold:    scaleDownThreshold,
		requestQueue:          make([]time.Time, 0),
		responseTimes:         make([]time.Duration, 0),
		errorCounts:           make(map[string]int),
	}
}

// SetScaleCallbacks sets the callbacks for scaling operations
func (as *AutoScaler) SetScaleCallbacks(onScaleUp, onScaleDown func() error) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.onScaleUp = onScaleUp
	as.onScaleDown = onScaleDown
}

// RecordRequest records a new request
func (as *AutoScaler) RecordRequest() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.currentRequests++
	as.requestQueue = append(as.requestQueue, time.Now())

	// Clean old requests (older than 1 minute)
	cutoff := time.Now().Add(-1 * time.Minute)
	for i, reqTime := range as.requestQueue {
		if reqTime.After(cutoff) {
			as.requestQueue = as.requestQueue[i:]
			break
		}
	}
}

// RecordResponse records a response completion
func (as *AutoScaler) RecordResponse(duration time.Duration) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.currentRequests--
	if as.currentRequests < 0 {
		as.currentRequests = 0
	}

	as.responseTimes = append(as.responseTimes, duration)

	// Keep only last 100 response times
	if len(as.responseTimes) > 100 {
		as.responseTimes = as.responseTimes[1:]
	}
}

// RecordError records an error
func (as *AutoScaler) RecordError(errorType string) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.errorCounts[errorType]++
}

// CheckAndScale checks metrics and triggers scaling if needed
func (as *AutoScaler) CheckAndScale(ctx context.Context) error {
	as.mu.RLock()

	// Calculate current load
	requestsPerMinute := len(as.requestQueue)
	avgResponseTime := as.calculateAverageResponseTime()
	errorRate := as.calculateErrorRate()

	as.mu.RUnlock()

	// Scale up conditions
	if requestsPerMinute > as.scaleUpThreshold || avgResponseTime > 2*time.Second {
		return as.scaleUp(ctx)
	}

	// Scale down conditions
	if requestsPerMinute < as.scaleDownThreshold && avgResponseTime < 500*time.Millisecond && errorRate < 1.0 {
		return as.scaleDown(ctx)
	}

	return nil
}

// scaleUp triggers scale up operations
func (as *AutoScaler) scaleUp(ctx context.Context) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.onScaleUp != nil {
		return as.onScaleUp()
	}

	// Default scale up actions
	// In a real implementation, this might:
	// - Increase connection pool size
	// - Enable additional caching
	// - Trigger Railway scaling
	return nil
}

// scaleDown triggers scale down operations
func (as *AutoScaler) scaleDown(ctx context.Context) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.onScaleDown != nil {
		return as.onScaleDown()
	}

	// Default scale down actions
	// In a real implementation, this might:
	// - Reduce connection pool size
	// - Disable some caching layers
	// - Trigger Railway scaling
	return nil
}

// calculateAverageResponseTime calculates the average response time
func (as *AutoScaler) calculateAverageResponseTime() time.Duration {
	if len(as.responseTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, duration := range as.responseTimes {
		total += duration
	}

	return total / time.Duration(len(as.responseTimes))
}

// calculateErrorRate calculates the error rate percentage
func (as *AutoScaler) calculateErrorRate() float64 {
	totalErrors := 0
	for _, count := range as.errorCounts {
		totalErrors += count
	}

	if len(as.responseTimes) == 0 {
		return 0
	}

	return float64(totalErrors) / float64(len(as.responseTimes)) * 100
}

// GetMetrics returns current auto-scaler metrics
func (as *AutoScaler) GetMetrics() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()

	return map[string]interface{}{
		"current_requests":     as.currentRequests,
		"requests_per_minute":  len(as.requestQueue),
		"avg_response_time":    as.calculateAverageResponseTime().String(),
		"error_rate":           as.calculateErrorRate(),
		"scale_up_threshold":   as.scaleUpThreshold,
		"scale_down_threshold": as.scaleDownThreshold,
	}
}

// StartMonitoring starts the auto-scaling monitoring loop
func (as *AutoScaler) StartMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := as.CheckAndScale(ctx); err != nil {
				// Log error but continue monitoring
				continue
			}
		}
	}
}
