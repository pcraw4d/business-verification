package performance_metrics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewBottleneckDetector(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	config := DefaultBottleneckDetectorConfig()

	detector := NewBottleneckDetector(logger, metrics, config)

	assert.NotNil(t, detector)
	assert.Equal(t, logger, detector.logger)
	assert.Equal(t, metrics, detector.metrics)
	assert.Equal(t, config.Thresholds, detector.thresholds)
	assert.Equal(t, config, detector.config)
	assert.NotNil(t, detector.bottlenecks)
}

func TestDefaultBottleneckThresholds(t *testing.T) {
	thresholds := DefaultBottleneckThresholds()

	assert.Equal(t, 5000.0, thresholds.ResponseTimeCritical)
	assert.Equal(t, 2000.0, thresholds.ResponseTimeHigh)
	assert.Equal(t, 1000.0, thresholds.ResponseTimeMedium)
	assert.Equal(t, 10.0, thresholds.ErrorRateCritical)
	assert.Equal(t, 5.0, thresholds.ErrorRateHigh)
	assert.Equal(t, 2.0, thresholds.ErrorRateMedium)
	assert.Equal(t, 10.0, thresholds.ThroughputLow)
	assert.Equal(t, 80.0, thresholds.CPUHigh)
	assert.Equal(t, 85.0, thresholds.MemoryHigh)
	assert.Equal(t, 20.0, thresholds.CacheMissHigh)
}

func TestDefaultBottleneckDetectorConfig(t *testing.T) {
	config := DefaultBottleneckDetectorConfig()

	assert.True(t, config.EnableDetection)
	assert.Equal(t, 5*time.Minute, config.AnalysisInterval)
	assert.Equal(t, 24*time.Hour, config.RetentionPeriod)
	assert.NotNil(t, config.Thresholds)
	assert.True(t, config.EnableAutoAnalysis)
	assert.Equal(t, 100, config.MaxBottlenecks)
}

func TestBottleneckDetector_AnalyzeBottlenecks_NoMetrics(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	ctx := context.Background()
	analysis, err := detector.AnalyzeBottlenecks(ctx)

	require.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, "No significant performance bottlenecks detected", analysis.Summary)
	assert.Equal(t, 0, len(analysis.Bottlenecks))
	assert.Equal(t, 0, analysis.CriticalCount)
	assert.Equal(t, 0, analysis.HighCount)
	assert.Equal(t, 0, analysis.MediumCount)
	assert.Equal(t, 0, analysis.LowCount)
}

func TestBottleneckDetector_DetectResponseTimeBottlenecks(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	ctx := context.Background()

	// Record response time metrics
	metrics.RecordResponseTime(ctx, "classification", 6000*time.Millisecond, nil) // Critical
	metrics.RecordResponseTime(ctx, "classification", 3000*time.Millisecond, nil) // High
	metrics.RecordResponseTime(ctx, "validation", 1500*time.Millisecond, nil)     // Medium

	bottlenecks := detector.detectResponseTimeBottlenecks(ctx)

	assert.Equal(t, 2, len(bottlenecks)) // classification and validation

	// Check classification bottleneck (should be high due to average of 6000ms and 3000ms = 4500ms)
	var classificationBottleneck *Bottleneck
	for _, b := range bottlenecks {
		if b.Operation == "classification" {
			classificationBottleneck = b
			break
		}
	}

	require.NotNil(t, classificationBottleneck)
	assert.Equal(t, BottleneckSeverityHigh, classificationBottleneck.Severity)
	assert.Equal(t, BottleneckTypeAlgorithm, classificationBottleneck.Type)
	assert.Contains(t, classificationBottleneck.Name, "Slow Response Time")
	assert.Contains(t, classificationBottleneck.Description, "High response time")
	assert.Greater(t, classificationBottleneck.Impact, 0.0)
	assert.Equal(t, 0.85, classificationBottleneck.Confidence)
	assert.Equal(t, "active", classificationBottleneck.Status)
	assert.NotEmpty(t, classificationBottleneck.RootCause)
	assert.NotEmpty(t, classificationBottleneck.Recommendations)
}

func TestBottleneckDetector_DetectErrorRateBottlenecks(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	ctx := context.Background()

	// Record error rate metrics
	metrics.RecordErrorRate(ctx, "classification", 15.0, nil) // Critical
	metrics.RecordErrorRate(ctx, "validation", 7.0, nil)      // High
	metrics.RecordErrorRate(ctx, "enrichment", 3.0, nil)      // Medium

	bottlenecks := detector.detectErrorRateBottlenecks(ctx)

	assert.Equal(t, 3, len(bottlenecks))

	// Check classification bottleneck (should be critical due to 15% error rate)
	var classificationBottleneck *Bottleneck
	for _, b := range bottlenecks {
		if b.Operation == "classification" {
			classificationBottleneck = b
			break
		}
	}

	require.NotNil(t, classificationBottleneck)
	assert.Equal(t, BottleneckSeverityCritical, classificationBottleneck.Severity)
	assert.Equal(t, BottleneckTypeAlgorithm, classificationBottleneck.Type)
	assert.Contains(t, classificationBottleneck.Name, "High Error Rate")
	assert.Contains(t, classificationBottleneck.Description, "Critical error rate")
	assert.Equal(t, 0.15, classificationBottleneck.Impact) // 15% normalized
	assert.Equal(t, 0.90, classificationBottleneck.Confidence)
	assert.Equal(t, "active", classificationBottleneck.Status)
}

func TestBottleneckDetector_DetectThroughputBottlenecks(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	ctx := context.Background()

	// Record throughput metrics
	metrics.RecordThroughput(ctx, "classification", 5, nil) // Low (below 10 req/s threshold)
	metrics.RecordThroughput(ctx, "validation", 15, nil)    // Normal

	bottlenecks := detector.detectThroughputBottlenecks(ctx)

	assert.Equal(t, 1, len(bottlenecks)) // Only classification should be detected

	bottleneck := bottlenecks[0]
	assert.Equal(t, "classification", bottleneck.Operation)
	assert.Equal(t, BottleneckSeverityMedium, bottleneck.Severity)
	assert.Equal(t, BottleneckTypeResource, bottleneck.Type)
	assert.Contains(t, bottleneck.Name, "Low Throughput")
	assert.Contains(t, bottleneck.Description, "Low throughput")
	assert.Greater(t, bottleneck.Impact, 0.0)
	assert.Equal(t, 0.80, bottleneck.Confidence)
}

func TestBottleneckDetector_DetectResourceBottlenecks(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	ctx := context.Background()

	// Record resource metrics
	metrics.RecordGauge(ctx, "cpu_usage", 85.0, "percentage", nil)    // High
	metrics.RecordGauge(ctx, "memory_usage", 90.0, "percentage", nil) // High

	bottlenecks := detector.detectResourceBottlenecks(ctx)

	assert.Equal(t, 2, len(bottlenecks)) // Both CPU and memory should be detected

	// Check CPU bottleneck
	var cpuBottleneck *Bottleneck
	for _, b := range bottlenecks {
		if b.Type == BottleneckTypeCPU {
			cpuBottleneck = b
			break
		}
	}

	require.NotNil(t, cpuBottleneck)
	assert.Equal(t, BottleneckSeverityHigh, cpuBottleneck.Severity)
	assert.Equal(t, BottleneckTypeCPU, cpuBottleneck.Type)
	assert.Contains(t, cpuBottleneck.Name, "High CPU Usage")
	assert.Equal(t, 0.85, cpuBottleneck.Impact) // 85% normalized
	assert.Equal(t, 0.85, cpuBottleneck.Confidence)

	// Check memory bottleneck
	var memoryBottleneck *Bottleneck
	for _, b := range bottlenecks {
		if b.Type == BottleneckTypeMemory {
			memoryBottleneck = b
			break
		}
	}

	require.NotNil(t, memoryBottleneck)
	assert.Equal(t, BottleneckSeverityHigh, memoryBottleneck.Severity)
	assert.Equal(t, BottleneckTypeMemory, memoryBottleneck.Type)
	assert.Contains(t, memoryBottleneck.Name, "High Memory Usage")
	assert.Equal(t, 0.90, memoryBottleneck.Impact) // 90% normalized
}

func TestBottleneckDetector_DetectAlgorithmBottlenecks(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	ctx := context.Background()

	// Record classification time metrics
	metrics.RecordGauge(ctx, "classification_time", 1500.0, "ms", nil) // Slow

	bottlenecks := detector.detectAlgorithmBottlenecks(ctx)

	assert.Equal(t, 1, len(bottlenecks))

	bottleneck := bottlenecks[0]
	assert.Equal(t, BottleneckSeverityMedium, bottleneck.Severity)
	assert.Equal(t, BottleneckTypeAlgorithm, bottleneck.Type)
	assert.Contains(t, bottleneck.Name, "Slow Classification Algorithm")
	assert.Contains(t, bottleneck.Description, "Classification taking too long")
	assert.Equal(t, "classification", bottleneck.Operation)
	assert.Equal(t, 0.30, bottleneck.Impact) // 1500ms / 5000ms
	assert.Equal(t, 0.80, bottleneck.Confidence)
}

func TestBottleneckDetector_CalculateAnalysisSummary(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	analysis := &BottleneckAnalysis{
		Bottlenecks: []*Bottleneck{
			{Severity: BottleneckSeverityCritical, Impact: 0.8},
			{Severity: BottleneckSeverityCritical, Impact: 0.9},
			{Severity: BottleneckSeverityHigh, Impact: 0.6},
			{Severity: BottleneckSeverityMedium, Impact: 0.4},
			{Severity: BottleneckSeverityLow, Impact: 0.2},
		},
	}

	detector.calculateAnalysisSummary(analysis)

	assert.Equal(t, 2, analysis.CriticalCount)
	assert.Equal(t, 1, analysis.HighCount)
	assert.Equal(t, 1, analysis.MediumCount)
	assert.Equal(t, 1, analysis.LowCount)
	assert.InDelta(t, 2.9, analysis.TotalImpact, 0.001) // 0.8 + 0.9 + 0.6 + 0.4 + 0.2
	assert.Contains(t, analysis.Summary, "Critical performance issues detected")
}

func TestBottleneckDetector_GenerateRecommendations(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	// Test with critical bottlenecks
	analysis := &BottleneckAnalysis{
		CriticalCount: 2,
		HighCount:     1,
		Bottlenecks:   []*Bottleneck{{}, {}, {}},
	}

	recommendations := detector.generateRecommendations(analysis)

	assert.Greater(t, len(recommendations), 0)
	assert.Contains(t, recommendations[0], "Immediate action required")
	assert.Contains(t, recommendations[1], "emergency scaling")
	assert.Contains(t, recommendations[2], "system architecture")

	// Test with no bottlenecks
	analysis = &BottleneckAnalysis{
		CriticalCount: 0,
		HighCount:     0,
		Bottlenecks:   []*Bottleneck{},
	}

	recommendations = detector.generateRecommendations(analysis)
	assert.Equal(t, 0, len(recommendations))
}

func TestBottleneckDetector_GetBottlenecks(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	// Add some bottlenecks manually
	bottleneck1 := &Bottleneck{
		ID:         "test1",
		Name:       "Test Bottleneck 1",
		DetectedAt: time.Now().Add(-1 * time.Hour),
	}
	bottleneck2 := &Bottleneck{
		ID:         "test2",
		Name:       "Test Bottleneck 2",
		DetectedAt: time.Now(),
	}

	detector.bottlenecks["test1"] = bottleneck1
	detector.bottlenecks["test2"] = bottleneck2

	bottlenecks := detector.GetBottlenecks()

	assert.Equal(t, 2, len(bottlenecks))
	// Should be sorted by detection time (newest first)
	assert.Equal(t, "test2", bottlenecks[0].ID)
	assert.Equal(t, "test1", bottlenecks[1].ID)
}

func TestBottleneckDetector_GetBottlenecksBySeverity(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	// Add bottlenecks with different severities
	bottleneck1 := &Bottleneck{
		ID:       "critical1",
		Severity: BottleneckSeverityCritical,
	}
	bottleneck2 := &Bottleneck{
		ID:       "high1",
		Severity: BottleneckSeverityHigh,
	}
	bottleneck3 := &Bottleneck{
		ID:       "critical2",
		Severity: BottleneckSeverityCritical,
	}

	detector.bottlenecks["critical1"] = bottleneck1
	detector.bottlenecks["high1"] = bottleneck2
	detector.bottlenecks["critical2"] = bottleneck3

	criticalBottlenecks := detector.GetBottlenecksBySeverity(BottleneckSeverityCritical)
	assert.Equal(t, 2, len(criticalBottlenecks))

	highBottlenecks := detector.GetBottlenecksBySeverity(BottleneckSeverityHigh)
	assert.Equal(t, 1, len(highBottlenecks))

	mediumBottlenecks := detector.GetBottlenecksBySeverity(BottleneckSeverityMedium)
	assert.Equal(t, 0, len(mediumBottlenecks))
}

func TestBottleneckDetector_GetBottlenecksByType(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	// Add bottlenecks with different types
	bottleneck1 := &Bottleneck{
		ID:   "cpu1",
		Type: BottleneckTypeCPU,
	}
	bottleneck2 := &Bottleneck{
		ID:   "memory1",
		Type: BottleneckTypeMemory,
	}
	bottleneck3 := &Bottleneck{
		ID:   "cpu2",
		Type: BottleneckTypeCPU,
	}

	detector.bottlenecks["cpu1"] = bottleneck1
	detector.bottlenecks["memory1"] = bottleneck2
	detector.bottlenecks["cpu2"] = bottleneck3

	cpuBottlenecks := detector.GetBottlenecksByType(BottleneckTypeCPU)
	assert.Equal(t, 2, len(cpuBottlenecks))

	memoryBottlenecks := detector.GetBottlenecksByType(BottleneckTypeMemory)
	assert.Equal(t, 1, len(memoryBottlenecks))

	networkBottlenecks := detector.GetBottlenecksByType(BottleneckTypeNetwork)
	assert.Equal(t, 0, len(networkBottlenecks))
}

func TestBottleneckDetector_ResolveBottleneck(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	// Add a bottleneck
	bottleneck := &Bottleneck{
		ID:     "test1",
		Name:   "Test Bottleneck",
		Status: "active",
	}
	detector.bottlenecks["test1"] = bottleneck

	// Resolve the bottleneck
	err := detector.ResolveBottleneck("test1")
	require.NoError(t, err)

	assert.Equal(t, "resolved", bottleneck.Status)
	assert.NotNil(t, bottleneck.ResolvedAt)

	// Try to resolve non-existent bottleneck
	err = detector.ResolveBottleneck("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "bottleneck not found")
}

func TestBottleneckDetector_GetSeverityWeight(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	assert.Equal(t, 4, detector.getSeverityWeight(BottleneckSeverityCritical))
	assert.Equal(t, 3, detector.getSeverityWeight(BottleneckSeverityHigh))
	assert.Equal(t, 2, detector.getSeverityWeight(BottleneckSeverityMedium))
	assert.Equal(t, 1, detector.getSeverityWeight(BottleneckSeverityLow))
	assert.Equal(t, 0, detector.getSeverityWeight(BottleneckSeverityInfo))
	assert.Equal(t, 0, detector.getSeverityWeight("unknown"))
}

func TestBottleneckDetector_CleanupOldBottlenecks(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	config := DefaultBottleneckDetectorConfig()
	config.RetentionPeriod = 1 * time.Hour
	detector := NewBottleneckDetector(logger, metrics, config)

	// Add old and new bottlenecks
	oldBottleneck := &Bottleneck{
		ID:         "old",
		DetectedAt: time.Now().Add(-2 * time.Hour), // Older than retention period
	}
	newBottleneck := &Bottleneck{
		ID:         "new",
		DetectedAt: time.Now(), // Within retention period
	}

	detector.bottlenecks["old"] = oldBottleneck
	detector.bottlenecks["new"] = newBottleneck

	// Trigger cleanup
	detector.cleanupOldBottlenecks()

	// Old bottleneck should be removed, new one should remain
	assert.NotContains(t, detector.bottlenecks, "old")
	assert.Contains(t, detector.bottlenecks, "new")
}

func TestBottleneckDetector_IntegrationTest(t *testing.T) {
	logger := zap.NewNop()
	metrics := NewPerformanceMetricsService(logger, nil)
	detector := NewBottleneckDetector(logger, metrics, nil)

	ctx := context.Background()

	// Record various metrics to trigger different types of bottlenecks
	metrics.RecordResponseTime(ctx, "classification", 6000*time.Millisecond, nil)
	metrics.RecordErrorRate(ctx, "validation", 12.0, nil)
	metrics.RecordThroughput(ctx, "enrichment", 5, nil)
	metrics.RecordGauge(ctx, "cpu_usage", 85.0, "percentage", nil)
	metrics.RecordGauge(ctx, "classification_time", 2000.0, "ms", nil)

	// Perform analysis
	analysis, err := detector.AnalyzeBottlenecks(ctx)
	require.NoError(t, err)

	// Should detect multiple bottlenecks
	assert.Greater(t, len(analysis.Bottlenecks), 0)
	assert.Greater(t, analysis.TotalImpact, 0.0)
	assert.NotEmpty(t, analysis.Summary)
	assert.NotEmpty(t, analysis.Recommendations)

	// Check that bottlenecks are stored
	storedBottlenecks := detector.GetBottlenecks()
	assert.Equal(t, len(analysis.Bottlenecks), len(storedBottlenecks))

	// Verify severity distribution
	totalCount := analysis.CriticalCount + analysis.HighCount + analysis.MediumCount + analysis.LowCount
	assert.Equal(t, len(analysis.Bottlenecks), totalCount)
}
