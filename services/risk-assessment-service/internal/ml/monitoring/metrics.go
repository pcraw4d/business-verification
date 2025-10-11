package monitoring

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MetricsCollector collects and tracks model performance metrics
type MetricsCollector struct {
	mu             sync.RWMutex
	modelMetrics   map[string]*ModelMetrics
	overallMetrics *OverallMetrics
	logger         *zap.Logger
	startTime      time.Time
	requestCount   int64
	errorCount     int64
	lastResetTime  time.Time
}

// ModelMetrics tracks metrics for a specific model
type ModelMetrics struct {
	ModelType           string           `json:"model_type"`
	RequestCount        int64            `json:"request_count"`
	ErrorCount          int64            `json:"error_count"`
	TotalLatency        time.Duration    `json:"total_latency"`
	MinLatency          time.Duration    `json:"min_latency"`
	MaxLatency          time.Duration    `json:"max_latency"`
	LatencyP50          time.Duration    `json:"latency_p50"`
	LatencyP95          time.Duration    `json:"latency_p95"`
	LatencyP99          time.Duration    `json:"latency_p99"`
	MemoryUsage         int64            `json:"memory_usage"`
	Accuracy            float64          `json:"accuracy"`
	LastUpdated         time.Time        `json:"last_updated"`
	HorizonDistribution map[int]int64    `json:"horizon_distribution"`
	ErrorTypes          map[string]int64 `json:"error_types"`
	latencyHistory      []time.Duration
}

// OverallMetrics tracks overall system metrics
type OverallMetrics struct {
	TotalRequests       int64            `json:"total_requests"`
	TotalErrors         int64            `json:"total_errors"`
	AverageLatency      time.Duration    `json:"average_latency"`
	TotalMemoryUsage    int64            `json:"total_memory_usage"`
	Uptime              time.Duration    `json:"uptime"`
	Throughput          float64          `json:"throughput"` // requests per minute
	ErrorRate           float64          `json:"error_rate"`
	LastUpdated         time.Time        `json:"last_updated"`
	ModelDistribution   map[string]int64 `json:"model_distribution"`
	HorizonDistribution map[int]int64    `json:"horizon_distribution"`
}

// PerformanceSnapshot represents a point-in-time snapshot of metrics
type PerformanceSnapshot struct {
	Timestamp      time.Time                `json:"timestamp"`
	ModelMetrics   map[string]*ModelMetrics `json:"model_metrics"`
	OverallMetrics *OverallMetrics          `json:"overall_metrics"`
	HealthStatus   string                   `json:"health_status"`
	Alerts         []Alert                  `json:"alerts,omitempty"`
}

// Alert represents a performance alert
type Alert struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
	ModelType   string    `json:"model_type,omitempty"`
	Threshold   float64   `json:"threshold,omitempty"`
	ActualValue float64   `json:"actual_value,omitempty"`
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(logger *zap.Logger) *MetricsCollector {
	return &MetricsCollector{
		modelMetrics: make(map[string]*ModelMetrics),
		overallMetrics: &OverallMetrics{
			ModelDistribution:   make(map[string]int64),
			HorizonDistribution: make(map[int]int64),
		},
		logger:        logger,
		startTime:     time.Now(),
		lastResetTime: time.Now(),
	}
}

// RecordInference records an inference request
func (mc *MetricsCollector) RecordInference(modelType string, latency time.Duration, horizonMonths int, err error) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Update overall metrics
	mc.requestCount++
	mc.overallMetrics.TotalRequests++
	mc.overallMetrics.ModelDistribution[modelType]++
	mc.overallMetrics.HorizonDistribution[horizonMonths]++

	if err != nil {
		mc.errorCount++
		mc.overallMetrics.TotalErrors++
	}

	// Update model-specific metrics
	modelMetrics, exists := mc.modelMetrics[modelType]
	if !exists {
		modelMetrics = &ModelMetrics{
			ModelType:           modelType,
			HorizonDistribution: make(map[int]int64),
			ErrorTypes:          make(map[string]int64),
			latencyHistory:      make([]time.Duration, 0, 1000),
		}
		mc.modelMetrics[modelType] = modelMetrics
	}

	// Update model metrics
	modelMetrics.RequestCount++
	modelMetrics.HorizonDistribution[horizonMonths]++
	modelMetrics.LastUpdated = time.Now()

	if err != nil {
		modelMetrics.ErrorCount++
		errorType := "unknown"
		if err != nil {
			errorType = err.Error()
		}
		modelMetrics.ErrorTypes[errorType]++
	} else {
		// Update latency metrics
		modelMetrics.TotalLatency += latency
		modelMetrics.latencyHistory = append(modelMetrics.latencyHistory, latency)

		// Keep only last 1000 latency measurements
		if len(modelMetrics.latencyHistory) > 1000 {
			modelMetrics.latencyHistory = modelMetrics.latencyHistory[1:]
		}

		// Update min/max latency
		if modelMetrics.MinLatency == 0 || latency < modelMetrics.MinLatency {
			modelMetrics.MinLatency = latency
		}
		if latency > modelMetrics.MaxLatency {
			modelMetrics.MaxLatency = latency
		}

		// Calculate percentiles
		modelMetrics.LatencyP50 = mc.calculatePercentile(modelMetrics.latencyHistory, 0.5)
		modelMetrics.LatencyP95 = mc.calculatePercentile(modelMetrics.latencyHistory, 0.95)
		modelMetrics.LatencyP99 = mc.calculatePercentile(modelMetrics.latencyHistory, 0.99)
	}

	// Update overall metrics
	mc.updateOverallMetrics()
}

// RecordMemoryUsage records memory usage for a model
func (mc *MetricsCollector) RecordMemoryUsage(modelType string, memoryUsage int64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	modelMetrics, exists := mc.modelMetrics[modelType]
	if !exists {
		modelMetrics = &ModelMetrics{
			ModelType:           modelType,
			HorizonDistribution: make(map[int]int64),
			ErrorTypes:          make(map[string]int64),
			latencyHistory:      make([]time.Duration, 0, 1000),
		}
		mc.modelMetrics[modelType] = modelMetrics
	}

	modelMetrics.MemoryUsage = memoryUsage
	modelMetrics.LastUpdated = time.Now()

	// Update overall memory usage
	mc.updateOverallMetrics()
}

// RecordAccuracy records accuracy for a model
func (mc *MetricsCollector) RecordAccuracy(modelType string, accuracy float64) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	modelMetrics, exists := mc.modelMetrics[modelType]
	if !exists {
		modelMetrics = &ModelMetrics{
			ModelType:           modelType,
			HorizonDistribution: make(map[int]int64),
			ErrorTypes:          make(map[string]int64),
			latencyHistory:      make([]time.Duration, 0, 1000),
		}
		mc.modelMetrics[modelType] = modelMetrics
	}

	modelMetrics.Accuracy = accuracy
	modelMetrics.LastUpdated = time.Now()
}

// GetSnapshot returns a current snapshot of all metrics
func (mc *MetricsCollector) GetSnapshot() *PerformanceSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Create a deep copy of metrics
	snapshot := &PerformanceSnapshot{
		Timestamp:      time.Now(),
		ModelMetrics:   make(map[string]*ModelMetrics),
		OverallMetrics: &OverallMetrics{},
		HealthStatus:   mc.getHealthStatus(),
		Alerts:         mc.generateAlerts(),
	}

	// Copy model metrics
	for modelType, metrics := range mc.modelMetrics {
		snapshot.ModelMetrics[modelType] = &ModelMetrics{
			ModelType:           metrics.ModelType,
			RequestCount:        metrics.RequestCount,
			ErrorCount:          metrics.ErrorCount,
			TotalLatency:        metrics.TotalLatency,
			MinLatency:          metrics.MinLatency,
			MaxLatency:          metrics.MaxLatency,
			LatencyP50:          metrics.LatencyP50,
			LatencyP95:          metrics.LatencyP95,
			LatencyP99:          metrics.LatencyP99,
			MemoryUsage:         metrics.MemoryUsage,
			Accuracy:            metrics.Accuracy,
			LastUpdated:         metrics.LastUpdated,
			HorizonDistribution: make(map[int]int64),
			ErrorTypes:          make(map[string]int64),
		}

		// Copy maps
		for k, v := range metrics.HorizonDistribution {
			snapshot.ModelMetrics[modelType].HorizonDistribution[k] = v
		}
		for k, v := range metrics.ErrorTypes {
			snapshot.ModelMetrics[modelType].ErrorTypes[k] = v
		}
	}

	// Copy overall metrics
	*snapshot.OverallMetrics = *mc.overallMetrics
	snapshot.OverallMetrics.ModelDistribution = make(map[string]int64)
	snapshot.OverallMetrics.HorizonDistribution = make(map[int]int64)

	for k, v := range mc.overallMetrics.ModelDistribution {
		snapshot.OverallMetrics.ModelDistribution[k] = v
	}
	for k, v := range mc.overallMetrics.HorizonDistribution {
		snapshot.OverallMetrics.HorizonDistribution[k] = v
	}

	return snapshot
}

// Reset resets all metrics
func (mc *MetricsCollector) Reset() {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.modelMetrics = make(map[string]*ModelMetrics)
	mc.overallMetrics = &OverallMetrics{
		ModelDistribution:   make(map[string]int64),
		HorizonDistribution: make(map[int]int64),
	}
	mc.requestCount = 0
	mc.errorCount = 0
	mc.startTime = time.Now()
	mc.lastResetTime = time.Now()

	mc.logger.Info("Metrics reset")
}

// GetModelMetrics returns metrics for a specific model
func (mc *MetricsCollector) GetModelMetrics(modelType string) (*ModelMetrics, bool) {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	metrics, exists := mc.modelMetrics[modelType]
	if !exists {
		return nil, false
	}

	// Return a copy
	return &ModelMetrics{
		ModelType:           metrics.ModelType,
		RequestCount:        metrics.RequestCount,
		ErrorCount:          metrics.ErrorCount,
		TotalLatency:        metrics.TotalLatency,
		MinLatency:          metrics.MinLatency,
		MaxLatency:          metrics.MaxLatency,
		LatencyP50:          metrics.LatencyP50,
		LatencyP95:          metrics.LatencyP95,
		LatencyP99:          metrics.LatencyP99,
		MemoryUsage:         metrics.MemoryUsage,
		Accuracy:            metrics.Accuracy,
		LastUpdated:         metrics.LastUpdated,
		HorizonDistribution: metrics.HorizonDistribution,
		ErrorTypes:          metrics.ErrorTypes,
	}, true
}

// GetOverallMetrics returns overall system metrics
func (mc *MetricsCollector) GetOverallMetrics() *OverallMetrics {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	// Return a copy
	return &OverallMetrics{
		TotalRequests:       mc.overallMetrics.TotalRequests,
		TotalErrors:         mc.overallMetrics.TotalErrors,
		AverageLatency:      mc.overallMetrics.AverageLatency,
		TotalMemoryUsage:    mc.overallMetrics.TotalMemoryUsage,
		Uptime:              mc.overallMetrics.Uptime,
		Throughput:          mc.overallMetrics.Throughput,
		ErrorRate:           mc.overallMetrics.ErrorRate,
		LastUpdated:         mc.overallMetrics.LastUpdated,
		ModelDistribution:   mc.overallMetrics.ModelDistribution,
		HorizonDistribution: mc.overallMetrics.HorizonDistribution,
	}
}

// updateOverallMetrics updates overall system metrics
func (mc *MetricsCollector) updateOverallMetrics() {
	now := time.Now()
	uptime := now.Sub(mc.startTime)

	mc.overallMetrics.Uptime = uptime
	mc.overallMetrics.LastUpdated = now

	// Calculate average latency
	if mc.requestCount > 0 {
		var totalLatency time.Duration
		for _, metrics := range mc.modelMetrics {
			totalLatency += metrics.TotalLatency
		}
		mc.overallMetrics.AverageLatency = totalLatency / time.Duration(mc.requestCount)
	}

	// Calculate throughput (requests per minute)
	if uptime > 0 {
		mc.overallMetrics.Throughput = float64(mc.requestCount) / uptime.Minutes()
	}

	// Calculate error rate
	if mc.requestCount > 0 {
		mc.overallMetrics.ErrorRate = float64(mc.errorCount) / float64(mc.requestCount)
	}

	// Calculate total memory usage
	var totalMemory int64
	for _, metrics := range mc.modelMetrics {
		totalMemory += metrics.MemoryUsage
	}
	mc.overallMetrics.TotalMemoryUsage = totalMemory
}

// calculatePercentile calculates the percentile of latency values
func (mc *MetricsCollector) calculatePercentile(latencies []time.Duration, percentile float64) time.Duration {
	if len(latencies) == 0 {
		return 0
	}

	// Sort latencies
	sorted := make([]time.Duration, len(latencies))
	copy(sorted, latencies)

	// Simple bubble sort (could be optimized for large datasets)
	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if sorted[j] > sorted[j+1] {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	index := int(float64(len(sorted)) * percentile)
	if index >= len(sorted) {
		index = len(sorted) - 1
	}

	return sorted[index]
}

// getHealthStatus determines the overall health status
func (mc *MetricsCollector) getHealthStatus() string {
	// Check error rate
	if mc.overallMetrics.ErrorRate > 0.1 { // 10% error rate
		return "unhealthy"
	}

	// Check latency
	if mc.overallMetrics.AverageLatency > 200*time.Millisecond {
		return "degraded"
	}

	// Check memory usage
	if mc.overallMetrics.TotalMemoryUsage > 1.5*1024*1024*1024 { // 1.5GB
		return "degraded"
	}

	return "healthy"
}

// generateAlerts generates performance alerts
func (mc *MetricsCollector) generateAlerts() []Alert {
	var alerts []Alert
	now := time.Now()

	// Check error rate
	if mc.overallMetrics.ErrorRate > 0.05 { // 5% error rate
		alerts = append(alerts, Alert{
			Type:        "high_error_rate",
			Severity:    "warning",
			Message:     "High error rate detected",
			Timestamp:   now,
			Threshold:   0.05,
			ActualValue: mc.overallMetrics.ErrorRate,
		})
	}

	// Check latency
	if mc.overallMetrics.AverageLatency > 150*time.Millisecond {
		alerts = append(alerts, Alert{
			Type:        "high_latency",
			Severity:    "warning",
			Message:     "High average latency detected",
			Timestamp:   now,
			Threshold:   150,
			ActualValue: float64(mc.overallMetrics.AverageLatency.Milliseconds()),
		})
	}

	// Check memory usage
	if mc.overallMetrics.TotalMemoryUsage > 1*1024*1024*1024 { // 1GB
		alerts = append(alerts, Alert{
			Type:        "high_memory_usage",
			Severity:    "warning",
			Message:     "High memory usage detected",
			Timestamp:   now,
			Threshold:   1 * 1024 * 1024 * 1024,
			ActualValue: float64(mc.overallMetrics.TotalMemoryUsage),
		})
	}

	// Check individual model performance
	for modelType, metrics := range mc.modelMetrics {
		if metrics.LatencyP95 > 200*time.Millisecond {
			alerts = append(alerts, Alert{
				Type:        "model_high_latency",
				Severity:    "warning",
				Message:     "High P95 latency for model",
				Timestamp:   now,
				ModelType:   modelType,
				Threshold:   200,
				ActualValue: float64(metrics.LatencyP95.Milliseconds()),
			})
		}

		if metrics.ErrorCount > 0 && float64(metrics.ErrorCount)/float64(metrics.RequestCount) > 0.1 {
			alerts = append(alerts, Alert{
				Type:        "model_high_error_rate",
				Severity:    "error",
				Message:     "High error rate for model",
				Timestamp:   now,
				ModelType:   modelType,
				Threshold:   0.1,
				ActualValue: float64(metrics.ErrorCount) / float64(metrics.RequestCount),
			})
		}
	}

	return alerts
}

// StartPeriodicReporting starts periodic metrics reporting
func (mc *MetricsCollector) StartPeriodicReporting(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			snapshot := mc.GetSnapshot()
			mc.logger.Info("Performance metrics snapshot",
				zap.String("health_status", snapshot.HealthStatus),
				zap.Int64("total_requests", snapshot.OverallMetrics.TotalRequests),
				zap.Float64("error_rate", snapshot.OverallMetrics.ErrorRate),
				zap.Duration("average_latency", snapshot.OverallMetrics.AverageLatency),
				zap.Int64("total_memory_mb", snapshot.OverallMetrics.TotalMemoryUsage/1024/1024),
				zap.Int("alerts_count", len(snapshot.Alerts)),
			)

			// Log alerts
			for _, alert := range snapshot.Alerts {
				mc.logger.Warn("Performance alert",
					zap.String("type", alert.Type),
					zap.String("severity", alert.Severity),
					zap.String("message", alert.Message),
					zap.String("model_type", alert.ModelType),
					zap.Float64("threshold", alert.Threshold),
					zap.Float64("actual_value", alert.ActualValue),
				)
			}
		}
	}
}
