package caching

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// CacheMetricType represents the type of cache metric
type CacheMetricType string

const (
	CacheMetricTypeHitRate        CacheMetricType = "hit_rate"
	CacheMetricTypeMissRate       CacheMetricType = "miss_rate"
	CacheMetricTypeEvictionRate   CacheMetricType = "eviction_rate"
	CacheMetricTypeExpirationRate CacheMetricType = "expiration_rate"
	CacheMetricTypeSize           CacheMetricType = "size"
	CacheMetricTypeEntryCount     CacheMetricType = "entry_count"
	CacheMetricTypeAccessTime     CacheMetricType = "access_time"
	CacheMetricTypeMemoryUsage    CacheMetricType = "memory_usage"
	CacheMetricTypeThroughput     CacheMetricType = "throughput"
	CacheMetricTypeLatency        CacheMetricType = "latency"
)

// CacheMetric represents a single cache metric
type CacheMetric struct {
	Type      CacheMetricType
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}

// CachePerformanceSnapshot represents a snapshot of cache performance
type CachePerformanceSnapshot struct {
	Timestamp         time.Time
	HitRate           float64
	MissRate          float64
	EvictionRate      float64
	ExpirationRate    float64
	TotalSize         int64
	EntryCount        int64
	AverageAccessTime time.Duration
	MemoryUsage       int64
	Throughput        float64
	AverageLatency    time.Duration
	ShardCount        int
	ActiveShards      int
}

// CacheBottleneck represents a detected performance bottleneck
type CacheBottleneck struct {
	ID              string
	Type            string
	Severity        string
	Description     string
	Metric          CacheMetricType
	Value           float64
	Threshold       float64
	Timestamp       time.Time
	Recommendations []string
}

// CacheAlert represents a cache performance alert
type CacheAlert struct {
	ID           string
	Type         string
	Severity     string
	Message      string
	Metric       CacheMetricType
	Value        float64
	Threshold    float64
	Timestamp    time.Time
	Acknowledged bool
}

// CachePerformanceReport represents a comprehensive performance report
type CachePerformanceReport struct {
	GeneratedAt     time.Time
	Period          time.Duration
	Summary         CachePerformanceSummary
	Metrics         []CacheMetric
	Bottlenecks     []CacheBottleneck
	Alerts          []CacheAlert
	Recommendations []string
	Trends          map[CacheMetricType]CacheMetricTrend
}

// CachePerformanceSummary represents a summary of cache performance
type CachePerformanceSummary struct {
	OverallScore     float64
	Status           string
	TotalOperations  int64
	TotalHits        int64
	TotalMisses      int64
	TotalEvictions   int64
	TotalExpirations int64
	AverageHitRate   float64
	AverageLatency   time.Duration
	PeakMemoryUsage  int64
	PeakThroughput   float64
}

// CacheMetricTrend represents a trend analysis for a metric
type CacheMetricTrend struct {
	Metric     CacheMetricType
	Direction  string // "increasing", "decreasing", "stable"
	Slope      float64
	Confidence float64
	Prediction float64
}

// CacheMonitorConfig represents the configuration for cache monitoring
type CacheMonitorConfig struct {
	Enabled              bool
	CollectionInterval   time.Duration
	RetentionPeriod      time.Duration
	MaxMetrics           int
	AlertThresholds      map[CacheMetricType]float64
	BottleneckThresholds map[CacheMetricType]float64
	TrendAnalysis        bool
	PredictionWindow     time.Duration
	Logger               *zap.Logger
}

// CacheMonitor manages cache performance monitoring
type CacheMonitor struct {
	cache              *IntelligentCache
	config             CacheMonitorConfig
	metrics            []CacheMetric
	bottlenecks        []CacheBottleneck
	alerts             []CacheAlert
	snapshots          []CachePerformanceSnapshot
	mu                 sync.RWMutex
	ctx                context.Context
	cancel             context.CancelFunc
	lastSnapshot       *CachePerformanceSnapshot
	alertHandlers      []AlertHandler
	bottleneckHandlers []BottleneckHandler
}

// AlertHandler is a function that handles cache alerts
type AlertHandler func(alert *CacheAlert)

// BottleneckHandler is a function that handles cache bottlenecks
type BottleneckHandler func(bottleneck *CacheBottleneck)

// NewCacheMonitor creates a new cache monitor
func NewCacheMonitor(cache *IntelligentCache, config CacheMonitorConfig) *CacheMonitor {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	if config.CollectionInterval == 0 {
		config.CollectionInterval = 30 * time.Second
	}

	if config.RetentionPeriod == 0 {
		config.RetentionPeriod = 24 * time.Hour
	}

	if config.MaxMetrics == 0 {
		config.MaxMetrics = 10000
	}

	if config.PredictionWindow == 0 {
		config.PredictionWindow = 1 * time.Hour
	}

	ctx, cancel := context.WithCancel(context.Background())

	monitor := &CacheMonitor{
		cache:              cache,
		config:             config,
		metrics:            make([]CacheMetric, 0),
		bottlenecks:        make([]CacheBottleneck, 0),
		alerts:             make([]CacheAlert, 0),
		snapshots:          make([]CachePerformanceSnapshot, 0),
		ctx:                ctx,
		cancel:             cancel,
		alertHandlers:      make([]AlertHandler, 0),
		bottleneckHandlers: make([]BottleneckHandler, 0),
	}

	if config.Enabled {
		go monitor.monitoringWorker()
	}

	return monitor
}

// AddAlertHandler adds an alert handler
func (cm *CacheMonitor) AddAlertHandler(handler AlertHandler) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.alertHandlers = append(cm.alertHandlers, handler)
}

// AddBottleneckHandler adds a bottleneck handler
func (cm *CacheMonitor) AddBottleneckHandler(handler BottleneckHandler) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.bottleneckHandlers = append(cm.bottleneckHandlers, handler)
}

// RecordMetric records a cache metric
func (cm *CacheMonitor) RecordMetric(metricType CacheMetricType, value float64, labels map[string]string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	metric := CacheMetric{
		Type:      metricType,
		Value:     value,
		Timestamp: time.Now(),
		Labels:    labels,
	}

	cm.metrics = append(cm.metrics, metric)

	// Keep metrics within retention period
	cm.cleanupOldMetrics()

	// Check for alerts
	cm.checkAlerts(metric)

	// Check for bottlenecks
	cm.checkBottlenecks(metric)
}

// GetMetrics retrieves metrics for a given time range
func (cm *CacheMonitor) GetMetrics(metricType CacheMetricType, start, end time.Time) []CacheMetric {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var filtered []CacheMetric
	for _, metric := range cm.metrics {
		if metric.Type == metricType && metric.Timestamp.After(start) && metric.Timestamp.Before(end) {
			filtered = append(filtered, metric)
		}
	}

	return filtered
}

// GetSnapshots retrieves performance snapshots for a given time range
func (cm *CacheMonitor) GetSnapshots(start, end time.Time) []CachePerformanceSnapshot {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var filtered []CachePerformanceSnapshot
	for _, snapshot := range cm.snapshots {
		if snapshot.Timestamp.After(start) && snapshot.Timestamp.Before(end) {
			filtered = append(filtered, snapshot)
		}
	}

	return filtered
}

// GetBottlenecks retrieves detected bottlenecks
func (cm *CacheMonitor) GetBottlenecks() []CacheBottleneck {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	bottlenecks := make([]CacheBottleneck, len(cm.bottlenecks))
	copy(bottlenecks, cm.bottlenecks)
	return bottlenecks
}

// GetAlerts retrieves active alerts
func (cm *CacheMonitor) GetAlerts() []CacheAlert {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	var activeAlerts []CacheAlert
	for _, alert := range cm.alerts {
		if !alert.Acknowledged {
			activeAlerts = append(activeAlerts, alert)
		}
	}

	return activeAlerts
}

// AcknowledgeAlert acknowledges an alert
func (cm *CacheMonitor) AcknowledgeAlert(alertID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for i, alert := range cm.alerts {
		if alert.ID == alertID {
			cm.alerts[i].Acknowledged = true
			return nil
		}
	}

	return fmt.Errorf("alert %s not found", alertID)
}

// GenerateReport generates a comprehensive performance report
func (cm *CacheMonitor) GenerateReport(period time.Duration) *CachePerformanceReport {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	end := time.Now()
	start := end.Add(-period)

	// Get metrics for the period
	metrics := cm.getMetricsForPeriod(start, end)
	snapshots := cm.getSnapshotsForPeriod(start, end)
	bottlenecks := cm.getBottlenecksForPeriod(start, end)
	alerts := cm.getAlertsForPeriod(start, end)

	// Calculate summary
	summary := cm.calculateSummary(metrics, snapshots)

	// Analyze trends
	trends := make(map[CacheMetricType]CacheMetricTrend)
	if cm.config.TrendAnalysis {
		trends = cm.analyzeTrends(metrics)
	}

	// Generate recommendations
	recommendations := cm.generateRecommendations(summary, bottlenecks, trends)

	return &CachePerformanceReport{
		GeneratedAt:     time.Now(),
		Period:          period,
		Summary:         summary,
		Metrics:         metrics,
		Bottlenecks:     bottlenecks,
		Alerts:          alerts,
		Recommendations: recommendations,
		Trends:          trends,
	}
}

// GetCurrentSnapshot gets the current performance snapshot
func (cm *CacheMonitor) GetCurrentSnapshot() *CachePerformanceSnapshot {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.lastSnapshot == nil {
		return cm.createSnapshot()
	}

	return cm.lastSnapshot
}

// Close closes the cache monitor
func (cm *CacheMonitor) Close() error {
	cm.cancel()
	return nil
}

// monitoringWorker runs the background monitoring worker
func (cm *CacheMonitor) monitoringWorker() {
	ticker := time.NewTicker(cm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.collectMetrics()
		}
	}
}

// collectMetrics collects current cache metrics
func (cm *CacheMonitor) collectMetrics() {
	// Get cache statistics
	stats := cm.cache.GetStats()
	analytics := cm.cache.GetAnalytics()

	// Calculate metrics
	hitRate := stats.HitRate
	missRate := stats.MissRate
	evictionRate := float64(stats.Evictions) / float64(stats.Hits+stats.Misses)
	expirationRate := float64(stats.Expirations) / float64(stats.Hits+stats.Misses)

	// Record metrics
	cm.RecordMetric(CacheMetricTypeHitRate, hitRate, nil)
	cm.RecordMetric(CacheMetricTypeMissRate, missRate, nil)
	cm.RecordMetric(CacheMetricTypeEvictionRate, evictionRate, nil)
	cm.RecordMetric(CacheMetricTypeExpirationRate, expirationRate, nil)
	cm.RecordMetric(CacheMetricTypeSize, float64(stats.TotalSize), nil)
	cm.RecordMetric(CacheMetricTypeEntryCount, float64(stats.EntryCount), nil)
	cm.RecordMetric(CacheMetricTypeAccessTime, analytics.AverageAccessTime.Seconds(), nil)

	// Create snapshot
	snapshot := cm.createSnapshot()
	cm.mu.Lock()
	cm.snapshots = append(cm.snapshots, *snapshot)
	cm.lastSnapshot = snapshot
	cm.mu.Unlock()

	// Cleanup old snapshots
	cm.cleanupOldSnapshots()
}

// createSnapshot creates a performance snapshot
func (cm *CacheMonitor) createSnapshot() *CachePerformanceSnapshot {
	stats := cm.cache.GetStats()
	analytics := cm.cache.GetAnalytics()

	return &CachePerformanceSnapshot{
		Timestamp:         time.Now(),
		HitRate:           stats.HitRate,
		MissRate:          stats.MissRate,
		EvictionRate:      float64(stats.Evictions) / float64(stats.Hits+stats.Misses),
		ExpirationRate:    float64(stats.Expirations) / float64(stats.Hits+stats.Misses),
		TotalSize:         stats.TotalSize,
		EntryCount:        stats.EntryCount,
		AverageAccessTime: analytics.AverageAccessTime,
		MemoryUsage:       stats.TotalSize,
		Throughput:        float64(stats.Hits+stats.Misses) / time.Since(time.Time{}).Seconds(),
		AverageLatency:    analytics.AverageAccessTime,
		ShardCount:        len(cm.cache.shards),
		ActiveShards:      len(cm.cache.shards), // All shards are active in current implementation
	}
}

// checkAlerts checks if a metric triggers an alert
func (cm *CacheMonitor) checkAlerts(metric CacheMetric) {
	threshold, exists := cm.config.AlertThresholds[metric.Type]
	if !exists {
		return
	}

	var shouldAlert bool
	switch metric.Type {
	case CacheMetricTypeHitRate:
		shouldAlert = metric.Value < threshold
	case CacheMetricTypeMissRate:
		shouldAlert = metric.Value > threshold
	case CacheMetricTypeEvictionRate:
		shouldAlert = metric.Value > threshold
	case CacheMetricTypeExpirationRate:
		shouldAlert = metric.Value > threshold
	case CacheMetricTypeLatency:
		shouldAlert = metric.Value > threshold
	default:
		shouldAlert = metric.Value > threshold
	}

	if shouldAlert {
		alert := &CacheAlert{
			ID:        generateAlertID(),
			Type:      string(metric.Type),
			Severity:  cm.determineSeverity(metric.Value, threshold),
			Message:   cm.generateAlertMessage(metric),
			Metric:    metric.Type,
			Value:     metric.Value,
			Threshold: threshold,
			Timestamp: metric.Timestamp,
		}

		cm.alerts = append(cm.alerts, *alert)

		// Notify alert handlers
		for _, handler := range cm.alertHandlers {
			handler(alert)
		}
	}
}

// checkBottlenecks checks if a metric indicates a bottleneck
func (cm *CacheMonitor) checkBottlenecks(metric CacheMetric) {
	threshold, exists := cm.config.BottleneckThresholds[metric.Type]
	if !exists {
		return
	}

	var isBottleneck bool
	switch metric.Type {
	case CacheMetricTypeHitRate:
		isBottleneck = metric.Value < threshold
	case CacheMetricTypeMissRate:
		isBottleneck = metric.Value > threshold
	case CacheMetricTypeEvictionRate:
		isBottleneck = metric.Value > threshold
	case CacheMetricTypeLatency:
		isBottleneck = metric.Value > threshold
	default:
		isBottleneck = metric.Value > threshold
	}

	if isBottleneck {
		bottleneck := &CacheBottleneck{
			ID:              generateBottleneckID(),
			Type:            string(metric.Type),
			Severity:        cm.determineSeverity(metric.Value, threshold),
			Description:     cm.generateBottleneckDescription(metric),
			Metric:          metric.Type,
			Value:           metric.Value,
			Threshold:       threshold,
			Timestamp:       metric.Timestamp,
			Recommendations: cm.generateBottleneckRecommendations(metric),
		}

		cm.bottlenecks = append(cm.bottlenecks, *bottleneck)

		// Notify bottleneck handlers
		for _, handler := range cm.bottleneckHandlers {
			handler(bottleneck)
		}
	}
}

// cleanupOldMetrics removes old metrics outside the retention period
func (cm *CacheMonitor) cleanupOldMetrics() {
	cutoff := time.Now().Add(-cm.config.RetentionPeriod)
	var valid []CacheMetric

	for _, metric := range cm.metrics {
		if metric.Timestamp.After(cutoff) {
			valid = append(valid, metric)
		}
	}

	// Limit the number of metrics
	if len(valid) > cm.config.MaxMetrics {
		valid = valid[len(valid)-cm.config.MaxMetrics:]
	}

	cm.metrics = valid
}

// cleanupOldSnapshots removes old snapshots outside the retention period
func (cm *CacheMonitor) cleanupOldSnapshots() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cutoff := time.Now().Add(-cm.config.RetentionPeriod)
	var valid []CachePerformanceSnapshot

	for _, snapshot := range cm.snapshots {
		if snapshot.Timestamp.After(cutoff) {
			valid = append(valid, snapshot)
		}
	}

	cm.snapshots = valid
}

// getMetricsForPeriod gets metrics for a specific time period
func (cm *CacheMonitor) getMetricsForPeriod(start, end time.Time) []CacheMetric {
	var filtered []CacheMetric
	for _, metric := range cm.metrics {
		if metric.Timestamp.After(start) && metric.Timestamp.Before(end) {
			filtered = append(filtered, metric)
		}
	}
	return filtered
}

// getSnapshotsForPeriod gets snapshots for a specific time period
func (cm *CacheMonitor) getSnapshotsForPeriod(start, end time.Time) []CachePerformanceSnapshot {
	var filtered []CachePerformanceSnapshot
	for _, snapshot := range cm.snapshots {
		if snapshot.Timestamp.After(start) && snapshot.Timestamp.Before(end) {
			filtered = append(filtered, snapshot)
		}
	}
	return filtered
}

// getBottlenecksForPeriod gets bottlenecks for a specific time period
func (cm *CacheMonitor) getBottlenecksForPeriod(start, end time.Time) []CacheBottleneck {
	var filtered []CacheBottleneck
	for _, bottleneck := range cm.bottlenecks {
		if bottleneck.Timestamp.After(start) && bottleneck.Timestamp.Before(end) {
			filtered = append(filtered, bottleneck)
		}
	}
	return filtered
}

// getAlertsForPeriod gets alerts for a specific time period
func (cm *CacheMonitor) getAlertsForPeriod(start, end time.Time) []CacheAlert {
	var filtered []CacheAlert
	for _, alert := range cm.alerts {
		if alert.Timestamp.After(start) && alert.Timestamp.Before(end) {
			filtered = append(filtered, alert)
		}
	}
	return filtered
}

// calculateSummary calculates performance summary
func (cm *CacheMonitor) calculateSummary(metrics []CacheMetric, snapshots []CachePerformanceSnapshot) CachePerformanceSummary {
	if len(metrics) == 0 {
		return CachePerformanceSummary{}
	}

	var totalOperations, totalHits, totalMisses, totalEvictions, totalExpirations int64
	var totalHitRate, totalLatency float64
	var peakMemoryUsage int64
	var peakThroughput float64

	// Calculate totals from snapshots
	for _, snapshot := range snapshots {
		totalOperations += int64(snapshot.Throughput * snapshot.Timestamp.Sub(snapshots[0].Timestamp).Seconds())
		if snapshot.MemoryUsage > peakMemoryUsage {
			peakMemoryUsage = snapshot.MemoryUsage
		}
		if snapshot.Throughput > peakThroughput {
			peakThroughput = snapshot.Throughput
		}
		totalHitRate += snapshot.HitRate
		totalLatency += snapshot.AverageLatency.Seconds()
	}

	// Calculate averages
	avgHitRate := totalHitRate / float64(len(snapshots))
	avgLatency := time.Duration(totalLatency/float64(len(snapshots))) * time.Second

	// Calculate overall score (0-100)
	overallScore := avgHitRate * 100

	// Determine status
	status := "excellent"
	if overallScore < 80 {
		status = "good"
	}
	if overallScore < 60 {
		status = "fair"
	}
	if overallScore < 40 {
		status = "poor"
	}

	return CachePerformanceSummary{
		OverallScore:     overallScore,
		Status:           status,
		TotalOperations:  totalOperations,
		TotalHits:        totalHits,
		TotalMisses:      totalMisses,
		TotalEvictions:   totalEvictions,
		TotalExpirations: totalExpirations,
		AverageHitRate:   avgHitRate,
		AverageLatency:   avgLatency,
		PeakMemoryUsage:  peakMemoryUsage,
		PeakThroughput:   peakThroughput,
	}
}

// analyzeTrends analyzes trends in metrics
func (cm *CacheMonitor) analyzeTrends(metrics []CacheMetric) map[CacheMetricType]CacheMetricTrend {
	trends := make(map[CacheMetricType]CacheMetricTrend)

	// Group metrics by type
	metricsByType := make(map[CacheMetricType][]CacheMetric)
	for _, metric := range metrics {
		metricsByType[metric.Type] = append(metricsByType[metric.Type], metric)
	}

	// Analyze trends for each metric type
	for metricType, typeMetrics := range metricsByType {
		if len(typeMetrics) < 2 {
			continue
		}

		trend := cm.calculateTrend(typeMetrics)
		trends[metricType] = trend
	}

	return trends
}

// calculateTrend calculates trend for a set of metrics
func (cm *CacheMonitor) calculateTrend(metrics []CacheMetric) CacheMetricTrend {
	if len(metrics) < 2 {
		return CacheMetricTrend{}
	}

	// Simple linear regression
	var sumX, sumY, sumXY, sumX2 float64
	for i, metric := range metrics {
		x := float64(i)
		y := metric.Value
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	n := float64(len(metrics))
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Determine direction
	direction := "stable"
	if slope > 0.01 {
		direction = "increasing"
	} else if slope < -0.01 {
		direction = "decreasing"
	}

	// Calculate confidence (simple correlation coefficient)
	meanX := sumX / n
	meanY := sumY / n
	var sumDiffXY, sumDiffX2, sumDiffY2 float64
	for i, metric := range metrics {
		x := float64(i)
		y := metric.Value
		diffX := x - meanX
		diffY := y - meanY
		sumDiffXY += diffX * diffY
		sumDiffX2 += diffX * diffX
		sumDiffY2 += diffY * diffY
	}

	confidence := 0.0
	if sumDiffX2 > 0 && sumDiffY2 > 0 {
		confidence = (sumDiffXY * sumDiffXY) / (sumDiffX2 * sumDiffY2)
	}

	// Simple prediction
	lastValue := metrics[len(metrics)-1].Value
	prediction := lastValue + slope*float64(cm.config.PredictionWindow.Seconds()/30) // Assuming 30s intervals

	return CacheMetricTrend{
		Metric:     metrics[0].Type,
		Direction:  direction,
		Slope:      slope,
		Confidence: confidence,
		Prediction: prediction,
	}
}

// generateRecommendations generates performance recommendations
func (cm *CacheMonitor) generateRecommendations(summary CachePerformanceSummary, bottlenecks []CacheBottleneck, trends map[CacheMetricType]CacheMetricTrend) []string {
	var recommendations []string

	// Hit rate recommendations
	if summary.AverageHitRate < 0.8 {
		recommendations = append(recommendations, "Consider increasing cache size or optimizing cache keys")
	}

	// Latency recommendations
	if summary.AverageLatency > 10*time.Millisecond {
		recommendations = append(recommendations, "Investigate high latency operations and consider caching optimization")
	}

	// Memory usage recommendations
	if summary.PeakMemoryUsage > 1<<30 { // 1GB
		recommendations = append(recommendations, "Consider implementing cache eviction policies or reducing cache size")
	}

	// Bottleneck-specific recommendations
	for _, bottleneck := range bottlenecks {
		recommendations = append(recommendations, bottleneck.Recommendations...)
	}

	// Trend-based recommendations
	for _, trend := range trends {
		if trend.Direction == "decreasing" && trend.Metric == CacheMetricTypeHitRate {
			recommendations = append(recommendations, "Hit rate is declining - review cache strategy and key patterns")
		}
	}

	return recommendations
}

// determineSeverity determines the severity level based on value and threshold
func (cm *CacheMonitor) determineSeverity(value, threshold float64) string {
	ratio := value / threshold
	if ratio > 2.0 {
		return "critical"
	} else if ratio > 1.5 {
		return "high"
	} else if ratio > 1.2 {
		return "medium"
	} else {
		return "low"
	}
}

// generateAlertMessage generates an alert message
func (cm *CacheMonitor) generateAlertMessage(metric CacheMetric) string {
	switch metric.Type {
	case CacheMetricTypeHitRate:
		return fmt.Sprintf("Cache hit rate is low: %.2f%%", metric.Value*100)
	case CacheMetricTypeMissRate:
		return fmt.Sprintf("Cache miss rate is high: %.2f%%", metric.Value*100)
	case CacheMetricTypeEvictionRate:
		return fmt.Sprintf("Cache eviction rate is high: %.2f%%", metric.Value*100)
	case CacheMetricTypeLatency:
		return fmt.Sprintf("Cache latency is high: %.2fms", metric.Value)
	default:
		return fmt.Sprintf("Cache metric %s is at %.2f", metric.Type, metric.Value)
	}
}

// generateBottleneckDescription generates a bottleneck description
func (cm *CacheMonitor) generateBottleneckDescription(metric CacheMetric) string {
	switch metric.Type {
	case CacheMetricTypeHitRate:
		return fmt.Sprintf("Low cache hit rate of %.2f%% indicates poor cache effectiveness", metric.Value*100)
	case CacheMetricTypeMissRate:
		return fmt.Sprintf("High cache miss rate of %.2f%% indicates frequent cache misses", metric.Value*100)
	case CacheMetricTypeEvictionRate:
		return fmt.Sprintf("High eviction rate of %.2f%% indicates aggressive cache eviction", metric.Value*100)
	case CacheMetricTypeLatency:
		return fmt.Sprintf("High latency of %.2fms indicates slow cache operations", metric.Value)
	default:
		return fmt.Sprintf("Performance bottleneck detected in %s metric", metric.Type)
	}
}

// generateBottleneckRecommendations generates recommendations for a bottleneck
func (cm *CacheMonitor) generateBottleneckRecommendations(metric CacheMetric) []string {
	var recommendations []string

	switch metric.Type {
	case CacheMetricTypeHitRate:
		recommendations = append(recommendations,
			"Increase cache size",
			"Review cache key patterns",
			"Implement cache warming strategies",
			"Optimize cache eviction policies")
	case CacheMetricTypeMissRate:
		recommendations = append(recommendations,
			"Analyze cache miss patterns",
			"Implement better cache key strategies",
			"Consider cache preloading",
			"Review cache invalidation strategies")
	case CacheMetricTypeEvictionRate:
		recommendations = append(recommendations,
			"Increase cache capacity",
			"Optimize eviction policies",
			"Review cache entry sizes",
			"Implement cache compression")
	case CacheMetricTypeLatency:
		recommendations = append(recommendations,
			"Profile cache operations",
			"Optimize cache data structures",
			"Consider cache sharding",
			"Review cache serialization")
	}

	return recommendations
}

// generateAlertID generates a unique alert ID
func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// generateBottleneckID generates a unique bottleneck ID
func generateBottleneckID() string {
	return fmt.Sprintf("bottleneck_%d", time.Now().UnixNano())
}
