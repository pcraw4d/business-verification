package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ScalingMetrics collects and provides metrics for auto-scaling decisions
type ScalingMetrics struct {
	logger *zap.Logger
	mu     sync.RWMutex

	// Request metrics
	requestCount int64
	requestRate  float64
	errorCount   int64
	errorRate    float64

	// Latency metrics
	avgLatency time.Duration
	p95Latency time.Duration
	p99Latency time.Duration

	// Resource metrics
	cpuUsage    float64
	memoryUsage float64

	// Queue metrics
	queueDepth          int64
	queueProcessingRate float64

	// Custom metrics
	customMetrics map[string]float64

	// Timestamps
	lastUpdate    time.Time
	lastScaleUp   time.Time
	lastScaleDown time.Time
}

// ScalingMetricsData represents the current scaling metrics
type ScalingMetricsData struct {
	RequestCount        int64              `json:"request_count"`
	RequestRate         float64            `json:"request_rate"`
	ErrorCount          int64              `json:"error_count"`
	ErrorRate           float64            `json:"error_rate"`
	AvgLatency          time.Duration      `json:"avg_latency"`
	P95Latency          time.Duration      `json:"p95_latency"`
	P99Latency          time.Duration      `json:"p99_latency"`
	CPUUsage            float64            `json:"cpu_usage"`
	MemoryUsage         float64            `json:"memory_usage"`
	QueueDepth          int64              `json:"queue_depth"`
	QueueProcessingRate float64            `json:"queue_processing_rate"`
	CustomMetrics       map[string]float64 `json:"custom_metrics"`
	LastUpdate          time.Time          `json:"last_update"`
	LastScaleUp         time.Time          `json:"last_scale_up"`
	LastScaleDown       time.Time          `json:"last_scale_down"`
}

// ScalingRecommendation represents a scaling recommendation
type ScalingRecommendation struct {
	Action          string    `json:"action"` // "scale_up", "scale_down", "no_action"
	Reason          string    `json:"reason"`
	Priority        int       `json:"priority"` // 1-10, higher is more urgent
	TargetReplicas  int       `json:"target_replicas"`
	CurrentReplicas int       `json:"current_replicas"`
	Confidence      float64   `json:"confidence"` // 0-1
	EstimatedImpact string    `json:"estimated_impact"`
	Timestamp       time.Time `json:"timestamp"`
}

// NewScalingMetrics creates a new scaling metrics collector
func NewScalingMetrics(logger *zap.Logger) *ScalingMetrics {
	return &ScalingMetrics{
		logger:        logger,
		customMetrics: make(map[string]float64),
		lastUpdate:    time.Now(),
	}
}

// UpdateRequestMetrics updates request-related metrics
func (sm *ScalingMetrics) UpdateRequestMetrics(count int64, errors int64, duration time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.requestCount += count
	sm.errorCount += errors
	sm.avgLatency = duration
	sm.lastUpdate = time.Now()

	// Calculate rates (simplified)
	sm.requestRate = float64(sm.requestCount) / time.Since(sm.lastUpdate).Seconds()
	sm.errorRate = float64(sm.errorCount) / float64(sm.requestCount)
}

// UpdateLatencyMetrics updates latency-related metrics
func (sm *ScalingMetrics) UpdateLatencyMetrics(avg, p95, p99 time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.avgLatency = avg
	sm.p95Latency = p95
	sm.p99Latency = p99
	sm.lastUpdate = time.Now()
}

// UpdateResourceMetrics updates resource usage metrics
func (sm *ScalingMetrics) UpdateResourceMetrics(cpu, memory float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.cpuUsage = cpu
	sm.memoryUsage = memory
	sm.lastUpdate = time.Now()
}

// UpdateQueueMetrics updates queue-related metrics
func (sm *ScalingMetrics) UpdateQueueMetrics(depth int64, processingRate float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.queueDepth = depth
	sm.queueProcessingRate = processingRate
	sm.lastUpdate = time.Now()
}

// UpdateCustomMetric updates a custom metric
func (sm *ScalingMetrics) UpdateCustomMetric(name string, value float64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.customMetrics[name] = value
	sm.lastUpdate = time.Now()
}

// GetMetrics returns the current scaling metrics
func (sm *ScalingMetrics) GetMetrics() *ScalingMetricsData {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create a deep copy
	customMetrics := make(map[string]float64)
	for k, v := range sm.customMetrics {
		customMetrics[k] = v
	}

	return &ScalingMetricsData{
		RequestCount:        sm.requestCount,
		RequestRate:         sm.requestRate,
		ErrorCount:          sm.errorCount,
		ErrorRate:           sm.errorRate,
		AvgLatency:          sm.avgLatency,
		P95Latency:          sm.p95Latency,
		P99Latency:          sm.p99Latency,
		CPUUsage:            sm.cpuUsage,
		MemoryUsage:         sm.memoryUsage,
		QueueDepth:          sm.queueDepth,
		QueueProcessingRate: sm.queueProcessingRate,
		CustomMetrics:       customMetrics,
		LastUpdate:          sm.lastUpdate,
		LastScaleUp:         sm.lastScaleUp,
		LastScaleDown:       sm.lastScaleDown,
	}
}

// GetScalingRecommendation analyzes metrics and provides scaling recommendations
func (sm *ScalingMetrics) GetScalingRecommendation(currentReplicas int) *ScalingRecommendation {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Analyze metrics and determine scaling action
	action, reason, priority, confidence := sm.analyzeMetrics()

	// Calculate target replicas
	targetReplicas := sm.calculateTargetReplicas(currentReplicas, action)

	// Estimate impact
	impact := sm.estimateImpact(currentReplicas, targetReplicas, action)

	return &ScalingRecommendation{
		Action:          action,
		Reason:          reason,
		Priority:        priority,
		TargetReplicas:  targetReplicas,
		CurrentReplicas: currentReplicas,
		Confidence:      confidence,
		EstimatedImpact: impact,
		Timestamp:       time.Now(),
	}
}

// RecordScaleEvent records a scaling event
func (sm *ScalingMetrics) RecordScaleEvent(action string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if action == "scale_up" {
		sm.lastScaleUp = time.Now()
	} else if action == "scale_down" {
		sm.lastScaleDown = time.Now()
	}

	sm.logger.Info("Scaling event recorded",
		zap.String("action", action),
		zap.Time("timestamp", time.Now()))
}

// Helper methods

func (sm *ScalingMetrics) analyzeMetrics() (string, string, int, float64) {
	// Scale up conditions
	if sm.cpuUsage > 70 || sm.memoryUsage > 80 {
		return "scale_up", "High resource usage", 8, 0.9
	}

	if sm.requestRate > 1000 {
		return "scale_up", "High request rate", 7, 0.8
	}

	if sm.p95Latency > 1*time.Second {
		return "scale_up", "High latency", 6, 0.7
	}

	if sm.errorRate > 0.05 {
		return "scale_up", "High error rate", 5, 0.6
	}

	if sm.queueDepth > 100 {
		return "scale_up", "High queue depth", 6, 0.7
	}

	// Scale down conditions
	if sm.cpuUsage < 30 && sm.memoryUsage < 40 && sm.requestRate < 200 {
		return "scale_down", "Low resource usage and request rate", 4, 0.6
	}

	if sm.p95Latency < 100*time.Millisecond && sm.requestRate < 100 {
		return "scale_down", "Low latency and request rate", 3, 0.5
	}

	// No action
	return "no_action", "Metrics within normal range", 1, 0.5
}

func (sm *ScalingMetrics) calculateTargetReplicas(current int, action string) int {
	switch action {
	case "scale_up":
		// Calculate based on current load
		if sm.cpuUsage > 70 {
			return int(float64(current) * 1.5)
		}
		if sm.requestRate > 1000 {
			return int(float64(current) * 1.3)
		}
		return current + 1

	case "scale_down":
		// Calculate based on current load
		if sm.cpuUsage < 30 && sm.requestRate < 200 {
			return max(1, current-1)
		}
		return current

	default:
		return current
	}
}

func (sm *ScalingMetrics) estimateImpact(current, target int, action string) string {
	diff := target - current

	if diff == 0 {
		return "No impact"
	}

	if action == "scale_up" {
		return fmt.Sprintf("Increase capacity by %d replicas", diff)
	} else if action == "scale_down" {
		return fmt.Sprintf("Reduce capacity by %d replicas", -diff)
	}

	return "Unknown impact"
}

// Helper function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ScalingMetricsCollector collects metrics from various sources
type ScalingMetricsCollector struct {
	metrics *ScalingMetrics
	logger  *zap.Logger
}

// NewScalingMetricsCollector creates a new metrics collector
func NewScalingMetricsCollector(logger *zap.Logger) *ScalingMetricsCollector {
	return &ScalingMetricsCollector{
		metrics: NewScalingMetrics(logger),
		logger:  logger,
	}
}

// StartCollection starts collecting metrics from various sources
func (smc *ScalingMetricsCollector) StartCollection(ctx context.Context) {
	// Start goroutines to collect metrics from different sources
	go smc.collectRequestMetrics(ctx)
	go smc.collectResourceMetrics(ctx)
	go smc.collectQueueMetrics(ctx)
	go smc.collectCustomMetrics(ctx)
}

// collectRequestMetrics collects request-related metrics
func (smc *ScalingMetricsCollector) collectRequestMetrics(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// This would collect actual request metrics
			// For now, we'll use placeholder values
			smc.metrics.UpdateRequestMetrics(100, 5, 200*time.Millisecond)
		}
	}
}

// collectResourceMetrics collects resource usage metrics
func (smc *ScalingMetricsCollector) collectResourceMetrics(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// This would collect actual resource metrics
			// For now, we'll use placeholder values
			smc.metrics.UpdateResourceMetrics(45.0, 60.0)
		}
	}
}

// collectQueueMetrics collects queue-related metrics
func (smc *ScalingMetricsCollector) collectQueueMetrics(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// This would collect actual queue metrics
			// For now, we'll use placeholder values
			smc.metrics.UpdateQueueMetrics(50, 25.0)
		}
	}
}

// collectCustomMetrics collects custom metrics
func (smc *ScalingMetricsCollector) collectCustomMetrics(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// This would collect actual custom metrics
			// For now, we'll use placeholder values
			smc.metrics.UpdateCustomMetric("active_connections", 150.0)
			smc.metrics.UpdateCustomMetric("cache_hit_rate", 0.85)
		}
	}
}

// GetMetrics returns the current metrics
func (smc *ScalingMetricsCollector) GetMetrics() *ScalingMetricsData {
	return smc.metrics.GetMetrics()
}

// GetScalingRecommendation returns scaling recommendations
func (smc *ScalingMetricsCollector) GetScalingRecommendation(currentReplicas int) *ScalingRecommendation {
	return smc.metrics.GetScalingRecommendation(currentReplicas)
}
