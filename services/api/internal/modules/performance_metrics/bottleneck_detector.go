package performance_metrics

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BottleneckType represents the type of performance bottleneck
type BottleneckType string

const (
	BottleneckTypeCPU        BottleneckType = "cpu"
	BottleneckTypeMemory     BottleneckType = "memory"
	BottleneckTypeNetwork    BottleneckType = "network"
	BottleneckTypeDatabase   BottleneckType = "database"
	BottleneckTypeCache      BottleneckType = "cache"
	BottleneckTypeAlgorithm  BottleneckType = "algorithm"
	BottleneckTypeExternal   BottleneckType = "external"
	BottleneckTypeConcurrent BottleneckType = "concurrent"
	BottleneckTypeResource   BottleneckType = "resource"
)

// BottleneckSeverity represents the severity level of a bottleneck
type BottleneckSeverity string

const (
	BottleneckSeverityCritical BottleneckSeverity = "critical"
	BottleneckSeverityHigh     BottleneckSeverity = "high"
	BottleneckSeverityMedium   BottleneckSeverity = "medium"
	BottleneckSeverityLow      BottleneckSeverity = "low"
	BottleneckSeverityInfo     BottleneckSeverity = "info"
)

// Bottleneck represents a detected performance bottleneck
type Bottleneck struct {
	ID              string             `json:"id"`
	Type            BottleneckType     `json:"type"`
	Severity        BottleneckSeverity `json:"severity"`
	Name            string             `json:"name"`
	Description     string             `json:"description"`
	Operation       string             `json:"operation"`
	DetectedAt      time.Time          `json:"detected_at"`
	Duration        time.Duration      `json:"duration"`
	Impact          float64            `json:"impact"`
	Confidence      float64            `json:"confidence"`
	Metrics         map[string]float64 `json:"metrics"`
	Labels          map[string]string  `json:"labels"`
	RootCause       string             `json:"root_cause"`
	Recommendations []string           `json:"recommendations"`
	Status          string             `json:"status"`
	ResolvedAt      *time.Time         `json:"resolved_at,omitempty"`
}

// BottleneckAnalysis represents the result of bottleneck analysis
type BottleneckAnalysis struct {
	AnalysisID      string        `json:"analysis_id"`
	Timestamp       time.Time     `json:"timestamp"`
	Duration        time.Duration `json:"duration"`
	Bottlenecks     []*Bottleneck `json:"bottlenecks"`
	Summary         string        `json:"summary"`
	CriticalCount   int           `json:"critical_count"`
	HighCount       int           `json:"high_count"`
	MediumCount     int           `json:"medium_count"`
	LowCount        int           `json:"low_count"`
	TotalImpact     float64       `json:"total_impact"`
	Recommendations []string      `json:"recommendations"`
}

// BottleneckThresholds defines thresholds for bottleneck detection
type BottleneckThresholds struct {
	ResponseTimeCritical float64 `json:"response_time_critical"` // ms
	ResponseTimeHigh     float64 `json:"response_time_high"`     // ms
	ResponseTimeMedium   float64 `json:"response_time_medium"`   // ms
	ErrorRateCritical    float64 `json:"error_rate_critical"`    // percentage
	ErrorRateHigh        float64 `json:"error_rate_high"`        // percentage
	ErrorRateMedium      float64 `json:"error_rate_medium"`      // percentage
	ThroughputLow        float64 `json:"throughput_low"`         // requests per second
	CPUHigh              float64 `json:"cpu_high"`               // percentage
	MemoryHigh           float64 `json:"memory_high"`            // percentage
	CacheMissHigh        float64 `json:"cache_miss_high"`        // percentage
}

// DefaultBottleneckThresholds returns default threshold values
func DefaultBottleneckThresholds() *BottleneckThresholds {
	return &BottleneckThresholds{
		ResponseTimeCritical: 5000, // 5 seconds
		ResponseTimeHigh:     2000, // 2 seconds
		ResponseTimeMedium:   1000, // 1 second
		ErrorRateCritical:    10.0, // 10%
		ErrorRateHigh:        5.0,  // 5%
		ErrorRateMedium:      2.0,  // 2%
		ThroughputLow:        10.0, // 10 req/s
		CPUHigh:              80.0, // 80%
		MemoryHigh:           85.0, // 85%
		CacheMissHigh:        20.0, // 20%
	}
}

// BottleneckDetector handles detection and analysis of performance bottlenecks
type BottleneckDetector struct {
	logger      *zap.Logger
	metrics     *PerformanceMetricsService
	thresholds  *BottleneckThresholds
	bottlenecks map[string]*Bottleneck
	mutex       sync.RWMutex
	config      *BottleneckDetectorConfig
}

// BottleneckDetectorConfig holds configuration for the bottleneck detector
type BottleneckDetectorConfig struct {
	EnableDetection    bool                  `json:"enable_detection"`
	AnalysisInterval   time.Duration         `json:"analysis_interval"`
	RetentionPeriod    time.Duration         `json:"retention_period"`
	Thresholds         *BottleneckThresholds `json:"thresholds"`
	EnableAutoAnalysis bool                  `json:"enable_auto_analysis"`
	MaxBottlenecks     int                   `json:"max_bottlenecks"`
}

// DefaultBottleneckDetectorConfig returns default configuration
func DefaultBottleneckDetectorConfig() *BottleneckDetectorConfig {
	return &BottleneckDetectorConfig{
		EnableDetection:    true,
		AnalysisInterval:   5 * time.Minute,
		RetentionPeriod:    24 * time.Hour,
		Thresholds:         DefaultBottleneckThresholds(),
		EnableAutoAnalysis: true,
		MaxBottlenecks:     100,
	}
}

// NewBottleneckDetector creates a new bottleneck detector
func NewBottleneckDetector(logger *zap.Logger, metrics *PerformanceMetricsService, config *BottleneckDetectorConfig) *BottleneckDetector {
	if config == nil {
		config = DefaultBottleneckDetectorConfig()
	}

	detector := &BottleneckDetector{
		logger:      logger,
		metrics:     metrics,
		thresholds:  config.Thresholds,
		bottlenecks: make(map[string]*Bottleneck),
		config:      config,
	}

	if config.EnableAutoAnalysis {
		go detector.startAutoAnalysis()
	}

	return detector
}

// AnalyzeBottlenecks performs comprehensive bottleneck analysis
func (b *BottleneckDetector) AnalyzeBottlenecks(ctx context.Context) (*BottleneckAnalysis, error) {
	start := time.Now()
	b.logger.Info("Starting bottleneck analysis")

	analysis := &BottleneckAnalysis{
		AnalysisID:  fmt.Sprintf("analysis_%d", time.Now().Unix()),
		Timestamp:   time.Now(),
		Bottlenecks: []*Bottleneck{},
	}

	// Analyze different types of bottlenecks
	bottlenecks := []*Bottleneck{}

	// Response time bottlenecks
	responseTimeBottlenecks := b.detectResponseTimeBottlenecks(ctx)
	bottlenecks = append(bottlenecks, responseTimeBottlenecks...)

	// Error rate bottlenecks
	errorRateBottlenecks := b.detectErrorRateBottlenecks(ctx)
	bottlenecks = append(bottlenecks, errorRateBottlenecks...)

	// Throughput bottlenecks
	throughputBottlenecks := b.detectThroughputBottlenecks(ctx)
	bottlenecks = append(bottlenecks, throughputBottlenecks...)

	// Resource bottlenecks
	resourceBottlenecks := b.detectResourceBottlenecks(ctx)
	bottlenecks = append(bottlenecks, resourceBottlenecks...)

	// Algorithm bottlenecks
	algorithmBottlenecks := b.detectAlgorithmBottlenecks(ctx)
	bottlenecks = append(bottlenecks, algorithmBottlenecks...)

	// Sort bottlenecks by severity and impact
	sort.Slice(bottlenecks, func(i, j int) bool {
		if b.getSeverityWeight(bottlenecks[i].Severity) != b.getSeverityWeight(bottlenecks[j].Severity) {
			return b.getSeverityWeight(bottlenecks[i].Severity) > b.getSeverityWeight(bottlenecks[j].Severity)
		}
		return bottlenecks[i].Impact > bottlenecks[j].Impact
	})

	analysis.Bottlenecks = bottlenecks
	analysis.Duration = time.Since(start)

	// Calculate summary statistics
	b.calculateAnalysisSummary(analysis)

	// Generate recommendations
	analysis.Recommendations = b.generateRecommendations(analysis)

	// Store detected bottlenecks
	b.storeBottlenecks(bottlenecks)

	b.logger.Info("Completed bottleneck analysis",
		zap.Int("total_bottlenecks", len(bottlenecks)),
		zap.Int("critical_count", analysis.CriticalCount),
		zap.Duration("analysis_duration", analysis.Duration))

	return analysis, nil
}

// detectResponseTimeBottlenecks detects response time related bottlenecks
func (b *BottleneckDetector) detectResponseTimeBottlenecks(ctx context.Context) []*Bottleneck {
	var bottlenecks []*Bottleneck

	responseTimeMetrics := b.metrics.GetMetricsByName("response_time")
	if len(responseTimeMetrics) == 0 {
		return bottlenecks
	}

	// Group metrics by operation
	operationMetrics := make(map[string][]*Metric)
	for _, metric := range responseTimeMetrics {
		operation := metric.Labels["operation"]
		operationMetrics[operation] = append(operationMetrics[operation], metric)
	}

	for operation, metrics := range operationMetrics {
		// Calculate average response time
		var totalTime float64
		for _, metric := range metrics {
			totalTime += metric.Value
		}
		avgResponseTime := totalTime / float64(len(metrics))

		// Determine severity based on thresholds
		var severity BottleneckSeverity
		var description string

		if avgResponseTime >= b.thresholds.ResponseTimeCritical {
			severity = BottleneckSeverityCritical
			description = fmt.Sprintf("Critical response time for %s: %.2fms", operation, avgResponseTime)
		} else if avgResponseTime >= b.thresholds.ResponseTimeHigh {
			severity = BottleneckSeverityHigh
			description = fmt.Sprintf("High response time for %s: %.2fms", operation, avgResponseTime)
		} else if avgResponseTime >= b.thresholds.ResponseTimeMedium {
			severity = BottleneckSeverityMedium
			description = fmt.Sprintf("Medium response time for %s: %.2fms", operation, avgResponseTime)
		} else {
			continue // No bottleneck detected
		}

		bottleneck := &Bottleneck{
			ID:          fmt.Sprintf("response_time_%s_%d", operation, time.Now().Unix()),
			Type:        BottleneckTypeAlgorithm,
			Severity:    severity,
			Name:        fmt.Sprintf("Slow Response Time - %s", operation),
			Description: description,
			Operation:   operation,
			DetectedAt:  time.Now(),
			Impact:      avgResponseTime / b.thresholds.ResponseTimeCritical,
			Confidence:  0.85,
			Metrics: map[string]float64{
				"avg_response_time": avgResponseTime,
				"sample_count":      float64(len(metrics)),
			},
			Labels: map[string]string{
				"operation": operation,
				"metric":    "response_time",
			},
			RootCause: "High response time may indicate inefficient algorithms, database queries, or external service calls",
			Recommendations: []string{
				"Optimize algorithm efficiency",
				"Review database query performance",
				"Implement caching strategies",
				"Consider async processing for long-running operations",
			},
			Status: "active",
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	return bottlenecks
}

// detectErrorRateBottlenecks detects error rate related bottlenecks
func (b *BottleneckDetector) detectErrorRateBottlenecks(ctx context.Context) []*Bottleneck {
	var bottlenecks []*Bottleneck

	errorRateMetrics := b.metrics.GetMetricsByName("error_rate")
	if len(errorRateMetrics) == 0 {
		return bottlenecks
	}

	// Group metrics by operation
	operationMetrics := make(map[string][]*Metric)
	for _, metric := range errorRateMetrics {
		operation := metric.Labels["operation"]
		operationMetrics[operation] = append(operationMetrics[operation], metric)
	}

	for operation, metrics := range operationMetrics {
		// Calculate average error rate
		var totalErrorRate float64
		for _, metric := range metrics {
			totalErrorRate += metric.Value
		}
		avgErrorRate := totalErrorRate / float64(len(metrics))

		// Determine severity based on thresholds
		var severity BottleneckSeverity
		var description string

		if avgErrorRate >= b.thresholds.ErrorRateCritical {
			severity = BottleneckSeverityCritical
			description = fmt.Sprintf("Critical error rate for %s: %.2f%%", operation, avgErrorRate)
		} else if avgErrorRate >= b.thresholds.ErrorRateHigh {
			severity = BottleneckSeverityHigh
			description = fmt.Sprintf("High error rate for %s: %.2f%%", operation, avgErrorRate)
		} else if avgErrorRate >= b.thresholds.ErrorRateMedium {
			severity = BottleneckSeverityMedium
			description = fmt.Sprintf("Medium error rate for %s: %.2f%%", operation, avgErrorRate)
		} else {
			continue // No bottleneck detected
		}

		bottleneck := &Bottleneck{
			ID:          fmt.Sprintf("error_rate_%s_%d", operation, time.Now().Unix()),
			Type:        BottleneckTypeAlgorithm,
			Severity:    severity,
			Name:        fmt.Sprintf("High Error Rate - %s", operation),
			Description: description,
			Operation:   operation,
			DetectedAt:  time.Now(),
			Impact:      avgErrorRate / 100.0, // Normalize to 0-1 range
			Confidence:  0.90,
			Metrics: map[string]float64{
				"avg_error_rate": avgErrorRate,
				"sample_count":   float64(len(metrics)),
			},
			Labels: map[string]string{
				"operation": operation,
				"metric":    "error_rate",
			},
			RootCause: "High error rate may indicate bugs, external service issues, or resource constraints",
			Recommendations: []string{
				"Review error logs and fix bugs",
				"Check external service health",
				"Implement better error handling",
				"Add circuit breakers for external calls",
			},
			Status: "active",
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	return bottlenecks
}

// detectThroughputBottlenecks detects throughput related bottlenecks
func (b *BottleneckDetector) detectThroughputBottlenecks(ctx context.Context) []*Bottleneck {
	var bottlenecks []*Bottleneck

	throughputMetrics := b.metrics.GetMetricsByName("throughput")
	if len(throughputMetrics) == 0 {
		return bottlenecks
	}

	// Group metrics by operation
	operationMetrics := make(map[string][]*Metric)
	for _, metric := range throughputMetrics {
		operation := metric.Labels["operation"]
		operationMetrics[operation] = append(operationMetrics[operation], metric)
	}

	for operation, metrics := range operationMetrics {
		// Calculate average throughput
		var totalThroughput float64
		for _, metric := range metrics {
			totalThroughput += metric.Value
		}
		avgThroughput := totalThroughput / float64(len(metrics))

		// Check if throughput is below threshold
		if avgThroughput >= b.thresholds.ThroughputLow {
			continue // No bottleneck detected
		}

		bottleneck := &Bottleneck{
			ID:          fmt.Sprintf("throughput_%s_%d", operation, time.Now().Unix()),
			Type:        BottleneckTypeResource,
			Severity:    BottleneckSeverityMedium,
			Name:        fmt.Sprintf("Low Throughput - %s", operation),
			Description: fmt.Sprintf("Low throughput for %s: %.2f req/s", operation, avgThroughput),
			Operation:   operation,
			DetectedAt:  time.Now(),
			Impact:      1.0 - (avgThroughput / b.thresholds.ThroughputLow),
			Confidence:  0.80,
			Metrics: map[string]float64{
				"avg_throughput": avgThroughput,
				"sample_count":   float64(len(metrics)),
			},
			Labels: map[string]string{
				"operation": operation,
				"metric":    "throughput",
			},
			RootCause: "Low throughput may indicate resource constraints, inefficient processing, or bottlenecks",
			Recommendations: []string{
				"Scale up resources (CPU, memory)",
				"Optimize processing algorithms",
				"Implement parallel processing",
				"Review resource allocation",
			},
			Status: "active",
		}

		bottlenecks = append(bottlenecks, bottleneck)
	}

	return bottlenecks
}

// detectResourceBottlenecks detects resource-related bottlenecks
func (b *BottleneckDetector) detectResourceBottlenecks(ctx context.Context) []*Bottleneck {
	var bottlenecks []*Bottleneck

	// This is a mock implementation - in a real system, you would collect
	// actual resource metrics from the system
	cpuMetrics := b.metrics.GetMetricsByName("cpu_usage")
	memoryMetrics := b.metrics.GetMetricsByName("memory_usage")

	// CPU bottleneck detection
	if len(cpuMetrics) > 0 {
		var totalCPU float64
		for _, metric := range cpuMetrics {
			totalCPU += metric.Value
		}
		avgCPU := totalCPU / float64(len(cpuMetrics))

		if avgCPU >= b.thresholds.CPUHigh {
			bottleneck := &Bottleneck{
				ID:          fmt.Sprintf("cpu_high_%d", time.Now().Unix()),
				Type:        BottleneckTypeCPU,
				Severity:    BottleneckSeverityHigh,
				Name:        "High CPU Usage",
				Description: fmt.Sprintf("High CPU usage: %.2f%%", avgCPU),
				DetectedAt:  time.Now(),
				Impact:      avgCPU / 100.0,
				Confidence:  0.85,
				Metrics: map[string]float64{
					"avg_cpu_usage": avgCPU,
				},
				RootCause: "High CPU usage may indicate inefficient algorithms or insufficient resources",
				Recommendations: []string{
					"Optimize CPU-intensive algorithms",
					"Scale up CPU resources",
					"Implement caching to reduce computation",
					"Consider async processing",
				},
				Status: "active",
			}
			bottlenecks = append(bottlenecks, bottleneck)
		}
	}

	// Memory bottleneck detection
	if len(memoryMetrics) > 0 {
		var totalMemory float64
		for _, metric := range memoryMetrics {
			totalMemory += metric.Value
		}
		avgMemory := totalMemory / float64(len(memoryMetrics))

		if avgMemory >= b.thresholds.MemoryHigh {
			bottleneck := &Bottleneck{
				ID:          fmt.Sprintf("memory_high_%d", time.Now().Unix()),
				Type:        BottleneckTypeMemory,
				Severity:    BottleneckSeverityHigh,
				Name:        "High Memory Usage",
				Description: fmt.Sprintf("High memory usage: %.2f%%", avgMemory),
				DetectedAt:  time.Now(),
				Impact:      avgMemory / 100.0,
				Confidence:  0.85,
				Metrics: map[string]float64{
					"avg_memory_usage": avgMemory,
				},
				RootCause: "High memory usage may indicate memory leaks or insufficient memory allocation",
				Recommendations: []string{
					"Check for memory leaks",
					"Scale up memory resources",
					"Optimize memory usage in algorithms",
					"Implement memory pooling",
				},
				Status: "active",
			}
			bottlenecks = append(bottlenecks, bottleneck)
		}
	}

	return bottlenecks
}

// detectAlgorithmBottlenecks detects algorithm-specific bottlenecks
func (b *BottleneckDetector) detectAlgorithmBottlenecks(ctx context.Context) []*Bottleneck {
	var bottlenecks []*Bottleneck

	// Analyze classification performance
	classificationMetrics := b.metrics.GetMetricsByName("classification_time")
	if len(classificationMetrics) > 0 {
		var totalTime float64
		for _, metric := range classificationMetrics {
			totalTime += metric.Value
		}
		avgTime := totalTime / float64(len(classificationMetrics))

		// Check if classification is taking too long
		if avgTime > 1000 { // More than 1 second
			bottleneck := &Bottleneck{
				ID:          fmt.Sprintf("classification_slow_%d", time.Now().Unix()),
				Type:        BottleneckTypeAlgorithm,
				Severity:    BottleneckSeverityMedium,
				Name:        "Slow Classification Algorithm",
				Description: fmt.Sprintf("Classification taking too long: %.2fms", avgTime),
				Operation:   "classification",
				DetectedAt:  time.Now(),
				Impact:      avgTime / 5000.0, // Normalize to 0-1 range
				Confidence:  0.80,
				Metrics: map[string]float64{
					"avg_classification_time": avgTime,
				},
				RootCause: "Classification algorithm may be inefficient or processing too much data",
				Recommendations: []string{
					"Optimize classification algorithms",
					"Implement early termination for simple cases",
					"Add caching for repeated classifications",
					"Consider parallel processing for complex cases",
				},
				Status: "active",
			}
			bottlenecks = append(bottlenecks, bottleneck)
		}
	}

	return bottlenecks
}

// calculateAnalysisSummary calculates summary statistics for the analysis
func (b *BottleneckDetector) calculateAnalysisSummary(analysis *BottleneckAnalysis) {
	analysis.CriticalCount = 0
	analysis.HighCount = 0
	analysis.MediumCount = 0
	analysis.LowCount = 0
	analysis.TotalImpact = 0

	for _, bottleneck := range analysis.Bottlenecks {
		analysis.TotalImpact += bottleneck.Impact

		switch bottleneck.Severity {
		case BottleneckSeverityCritical:
			analysis.CriticalCount++
		case BottleneckSeverityHigh:
			analysis.HighCount++
		case BottleneckSeverityMedium:
			analysis.MediumCount++
		case BottleneckSeverityLow:
			analysis.LowCount++
		}
	}

	// Generate summary
	if analysis.CriticalCount > 0 {
		analysis.Summary = fmt.Sprintf("Critical performance issues detected: %d critical, %d high, %d medium bottlenecks",
			analysis.CriticalCount, analysis.HighCount, analysis.MediumCount)
	} else if analysis.HighCount > 0 {
		analysis.Summary = fmt.Sprintf("Performance issues detected: %d high, %d medium bottlenecks",
			analysis.HighCount, analysis.MediumCount)
	} else if analysis.MediumCount > 0 {
		analysis.Summary = fmt.Sprintf("Minor performance issues detected: %d medium bottlenecks", analysis.MediumCount)
	} else {
		analysis.Summary = "No significant performance bottlenecks detected"
	}
}

// generateRecommendations generates actionable recommendations based on analysis
func (b *BottleneckDetector) generateRecommendations(analysis *BottleneckAnalysis) []string {
	var recommendations []string

	// Critical bottlenecks
	if analysis.CriticalCount > 0 {
		recommendations = append(recommendations,
			"Immediate action required: Address critical bottlenecks first",
			"Consider emergency scaling or resource allocation",
			"Review system architecture for fundamental issues")
	}

	// High severity bottlenecks
	if analysis.HighCount > 0 {
		recommendations = append(recommendations,
			"Prioritize high-severity bottlenecks for next sprint",
			"Implement monitoring alerts for these bottlenecks",
			"Consider performance optimization sprints")
	}

	// General recommendations
	if len(analysis.Bottlenecks) > 0 {
		recommendations = append(recommendations,
			"Implement continuous performance monitoring",
			"Set up automated performance testing",
			"Create performance budgets for key operations",
			"Establish performance review processes")
	}

	return recommendations
}

// storeBottlenecks stores detected bottlenecks for tracking
func (b *BottleneckDetector) storeBottlenecks(bottlenecks []*Bottleneck) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	for _, bottleneck := range bottlenecks {
		b.bottlenecks[bottleneck.ID] = bottleneck
	}

	// Clean up old bottlenecks
	b.cleanupOldBottlenecks()
}

// cleanupOldBottlenecks removes old bottlenecks based on retention period
func (b *BottleneckDetector) cleanupOldBottlenecks() {
	cutoff := time.Now().Add(-b.config.RetentionPeriod)

	for id, bottleneck := range b.bottlenecks {
		if bottleneck.DetectedAt.Before(cutoff) {
			delete(b.bottlenecks, id)
		}
	}
}

// GetBottlenecks retrieves all stored bottlenecks
func (b *BottleneckDetector) GetBottlenecks() []*Bottleneck {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	var bottlenecks []*Bottleneck
	for _, bottleneck := range b.bottlenecks {
		bottlenecks = append(bottlenecks, bottleneck)
	}

	// Sort by detection time (newest first)
	sort.Slice(bottlenecks, func(i, j int) bool {
		return bottlenecks[i].DetectedAt.After(bottlenecks[j].DetectedAt)
	})

	return bottlenecks
}

// GetBottlenecksBySeverity retrieves bottlenecks by severity
func (b *BottleneckDetector) GetBottlenecksBySeverity(severity BottleneckSeverity) []*Bottleneck {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	var bottlenecks []*Bottleneck
	for _, bottleneck := range b.bottlenecks {
		if bottleneck.Severity == severity {
			bottlenecks = append(bottlenecks, bottleneck)
		}
	}

	return bottlenecks
}

// GetBottlenecksByType retrieves bottlenecks by type
func (b *BottleneckDetector) GetBottlenecksByType(bottleneckType BottleneckType) []*Bottleneck {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	var bottlenecks []*Bottleneck
	for _, bottleneck := range b.bottlenecks {
		if bottleneck.Type == bottleneckType {
			bottlenecks = append(bottlenecks, bottleneck)
		}
	}

	return bottlenecks
}

// ResolveBottleneck marks a bottleneck as resolved
func (b *BottleneckDetector) ResolveBottleneck(bottleneckID string) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	bottleneck, exists := b.bottlenecks[bottleneckID]
	if !exists {
		return fmt.Errorf("bottleneck not found: %s", bottleneckID)
	}

	now := time.Now()
	bottleneck.Status = "resolved"
	bottleneck.ResolvedAt = &now

	b.logger.Info("Bottleneck resolved",
		zap.String("bottleneck_id", bottleneckID),
		zap.String("name", bottleneck.Name))

	return nil
}

// getSeverityWeight returns a numeric weight for severity comparison
func (b *BottleneckDetector) getSeverityWeight(severity BottleneckSeverity) int {
	switch severity {
	case BottleneckSeverityCritical:
		return 4
	case BottleneckSeverityHigh:
		return 3
	case BottleneckSeverityMedium:
		return 2
	case BottleneckSeverityLow:
		return 1
	case BottleneckSeverityInfo:
		return 0
	default:
		return 0
	}
}

// startAutoAnalysis starts the automatic analysis routine
func (b *BottleneckDetector) startAutoAnalysis() {
	ticker := time.NewTicker(b.config.AnalysisInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx := context.Background()
		analysis, err := b.AnalyzeBottlenecks(ctx)
		if err != nil {
			b.logger.Error("Auto analysis failed", zap.Error(err))
			continue
		}

		if len(analysis.Bottlenecks) > 0 {
			b.logger.Info("Auto analysis completed",
				zap.Int("bottlenecks_found", len(analysis.Bottlenecks)),
				zap.Int("critical_count", analysis.CriticalCount),
				zap.Int("high_count", analysis.HighCount))
		}
	}
}
