package performance

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Optimizer provides comprehensive performance optimization
type Optimizer struct {
	logger          *zap.Logger
	profiler        *Profiler
	dbOptimizer     *DBOptimizer
	cacheOptimizer  *CacheOptimizer
	responseMonitor *ResponseMonitor
	config          *OptimizerConfig
	mu              sync.RWMutex
}

// OptimizerConfig contains optimizer configuration
type OptimizerConfig struct {
	EnableProfiling          bool          `json:"enable_profiling"`
	EnableDBOptimization     bool          `json:"enable_db_optimization"`
	EnableCaching            bool          `json:"enable_caching"`
	EnableResponseMonitoring bool          `json:"enable_response_monitoring"`
	PerformanceThreshold     time.Duration `json:"performance_threshold"`
	OptimizationInterval     time.Duration `json:"optimization_interval"`
	EnableAutoOptimization   bool          `json:"enable_auto_optimization"`
	TargetP95                time.Duration `json:"target_p95"`
	TargetP99                time.Duration `json:"target_p99"`
	TargetThroughput         int           `json:"target_throughput"`
}

// OptimizationReport contains optimization results
type OptimizationReport struct {
	Timestamp            time.Time                    `json:"timestamp"`
	OverallScore         float64                      `json:"overall_score"`
	PerformanceScore     float64                      `json:"performance_score"`
	DatabaseScore        float64                      `json:"database_score"`
	CacheScore           float64                      `json:"cache_score"`
	ResponseTimeScore    float64                      `json:"response_time_score"`
	Recommendations      []OptimizationRecommendation `json:"recommendations"`
	AppliedOptimizations []string                     `json:"applied_optimizations"`
	Metrics              map[string]interface{}       `json:"metrics"`
}

// OptimizationRecommendation provides optimization recommendations
type OptimizationRecommendation struct {
	Type        string   `json:"type"`
	Priority    string   `json:"priority"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Effort      string   `json:"effort"`
	Actions     []string `json:"actions"`
}

// NewOptimizer creates a new performance optimizer
func NewOptimizer(
	logger *zap.Logger,
	db *sql.DB,
	config *OptimizerConfig,
) *Optimizer {
	// Create profiler
	profiler := NewProfiler(logger, config.PerformanceThreshold)

	// Create database optimizer
	var dbOptimizer *DBOptimizer
	if config.EnableDBOptimization {
		dbConfig := DefaultDBConfig()
		dbOptimizer = NewDBOptimizer(logger, db, profiler, dbConfig)
	}

	// Create cache optimizer
	var cacheOptimizer *CacheOptimizer
	if config.EnableCaching {
		cacheConfig := DefaultCacheConfig()
		cacheOptimizer = NewCacheOptimizer(logger, profiler, cacheConfig)
	}

	// Create response monitor
	var responseMonitor *ResponseMonitor
	if config.EnableResponseMonitoring {
		responseConfig := DefaultResponseMonitorConfig()
		responseConfig.P95Threshold = config.TargetP95
		responseConfig.P99Threshold = config.TargetP99
		responseMonitor = NewResponseMonitor(logger, profiler, responseConfig)
	}

	optimizer := &Optimizer{
		logger:          logger,
		profiler:        profiler,
		dbOptimizer:     dbOptimizer,
		cacheOptimizer:  cacheOptimizer,
		responseMonitor: responseMonitor,
		config:          config,
	}

	// Start optimization routine
	if config.EnableAutoOptimization {
		go optimizer.optimize()
	}

	return optimizer
}

// GetProfiler returns the profiler instance
func (o *Optimizer) GetProfiler() *Profiler {
	return o.profiler
}

// GetDBOptimizer returns the database optimizer instance
func (o *Optimizer) GetDBOptimizer() *DBOptimizer {
	return o.dbOptimizer
}

// GetCacheOptimizer returns the cache optimizer instance
func (o *Optimizer) GetCacheOptimizer() *CacheOptimizer {
	return o.cacheOptimizer
}

// GetResponseMonitor returns the response monitor instance
func (o *Optimizer) GetResponseMonitor() *ResponseMonitor {
	return o.responseMonitor
}

// Optimize performs comprehensive performance optimization
func (o *Optimizer) Optimize() (*OptimizationReport, error) {
	o.logger.Info("Starting performance optimization")

	report := &OptimizationReport{
		Timestamp:            time.Now(),
		Recommendations:      make([]OptimizationRecommendation, 0),
		AppliedOptimizations: make([]string, 0),
		Metrics:              make(map[string]interface{}),
	}

	// Collect current metrics
	o.collectMetrics(report)

	// Analyze performance
	o.analyzePerformance(report)

	// Generate recommendations
	o.generateRecommendations(report)

	// Apply optimizations
	o.applyOptimizations(report)

	// Calculate overall score
	o.calculateOverallScore(report)

	o.logger.Info("Performance optimization completed",
		zap.Float64("overall_score", report.OverallScore),
		zap.Int("recommendations", len(report.Recommendations)),
		zap.Int("applied_optimizations", len(report.AppliedOptimizations)))

	return report, nil
}

// collectMetrics collects current performance metrics
func (o *Optimizer) collectMetrics(report *OptimizationReport) {
	// Profiler metrics
	if o.profiler != nil {
		report.Metrics["profiler"] = o.profiler.GetStats()
	}

	// Database metrics
	if o.dbOptimizer != nil {
		report.Metrics["database"] = o.dbOptimizer.GetStats()
	}

	// Cache metrics
	if o.cacheOptimizer != nil {
		report.Metrics["cache"] = o.cacheOptimizer.GetOverallStats()
	}

	// Response monitor metrics
	if o.responseMonitor != nil {
		report.Metrics["response_monitor"] = o.responseMonitor.GetStats()
	}
}

// analyzePerformance analyzes current performance
func (o *Optimizer) analyzePerformance(report *OptimizationReport) {
	// Analyze profiler data
	if profilerStats, ok := report.Metrics["profiler"].(*PerformanceStats); ok {
		report.PerformanceScore = o.calculatePerformanceScore(profilerStats)
	}

	// Analyze database performance
	if dbStats, ok := report.Metrics["database"].(*DBStats); ok {
		report.DatabaseScore = o.calculateDatabaseScore(dbStats)
	}

	// Analyze cache performance
	if cacheStats, ok := report.Metrics["cache"].(*CacheStats); ok {
		report.CacheScore = o.calculateCacheScore(cacheStats)
	}

	// Analyze response time performance
	if responseStats, ok := report.Metrics["response_monitor"].(*ResponseStats); ok {
		report.ResponseTimeScore = o.calculateResponseTimeScore(responseStats)
	}
}

// calculatePerformanceScore calculates performance score based on profiler stats
func (o *Optimizer) calculatePerformanceScore(stats *PerformanceStats) float64 {
	score := 100.0

	// Deduct points for slow operations
	if stats.P95ResponseTime > o.config.TargetP95 {
		excess := float64(stats.P95ResponseTime-o.config.TargetP95) / float64(o.config.TargetP95)
		score -= excess * 30
	}

	if stats.P99ResponseTime > o.config.TargetP99 {
		excess := float64(stats.P99ResponseTime-o.config.TargetP99) / float64(o.config.TargetP99)
		score -= excess * 40
	}

	// Deduct points for high memory usage
	if stats.MemoryStats.HeapAlloc > 500*1024*1024 { // 500MB
		excess := float64(stats.MemoryStats.HeapAlloc-500*1024*1024) / (500 * 1024 * 1024)
		score -= excess * 20
	}

	// Deduct points for high goroutine count
	if stats.GoroutineCount > 1000 {
		excess := float64(stats.GoroutineCount-1000) / 1000
		score -= excess * 10
	}

	if score < 0 {
		score = 0
	}

	return score
}

// calculateDatabaseScore calculates database performance score
func (o *Optimizer) calculateDatabaseScore(stats *DBStats) float64 {
	score := 100.0

	// Deduct points for connection pool issues
	if stats.WaitCount > 0 {
		score -= float64(stats.WaitCount) * 5
	}

	// Deduct points for slow queries
	if stats.TotalQueries > 0 {
		slowQueryRatio := float64(stats.SlowQueries) / float64(stats.TotalQueries)
		score -= slowQueryRatio * 50
	}

	// Deduct points for high average query time
	if stats.AverageQueryTime > 100*time.Millisecond {
		excess := float64(stats.AverageQueryTime-100*time.Millisecond) / float64(100*time.Millisecond)
		score -= excess * 30
	}

	if score < 0 {
		score = 0
	}

	return score
}

// calculateCacheScore calculates cache performance score
func (o *Optimizer) calculateCacheScore(stats *CacheStats) float64 {
	score := 100.0

	// Deduct points for low hit rate
	if stats.HitRate < 0.8 {
		score -= (0.8 - stats.HitRate) * 100
	}

	// Deduct points for high eviction rate
	if stats.Hits+stats.Misses > 0 {
		evictionRate := float64(stats.Evictions) / float64(stats.Hits+stats.Misses)
		score -= evictionRate * 50
	}

	if score < 0 {
		score = 0
	}

	return score
}

// calculateResponseTimeScore calculates response time performance score
func (o *Optimizer) calculateResponseTimeScore(stats *ResponseStats) float64 {
	score := 100.0

	// Deduct points for exceeding P95 threshold
	if stats.P95Time > o.config.TargetP95 {
		excess := float64(stats.P95Time-o.config.TargetP95) / float64(o.config.TargetP95)
		score -= excess * 40
	}

	// Deduct points for exceeding P99 threshold
	if stats.P99Time > o.config.TargetP99 {
		excess := float64(stats.P99Time-o.config.TargetP99) / float64(o.config.TargetP99)
		score -= excess * 50
	}

	// Deduct points for high error rate
	if stats.TotalRequests > 0 {
		errorRate := float64(stats.FailedRequests) / float64(stats.TotalRequests)
		score -= errorRate * 100
	}

	if score < 0 {
		score = 0
	}

	return score
}

// generateRecommendations generates optimization recommendations
func (o *Optimizer) generateRecommendations(report *OptimizationReport) {
	// Performance recommendations
	if report.PerformanceScore < 80 {
		report.Recommendations = append(report.Recommendations, OptimizationRecommendation{
			Type:        "performance",
			Priority:    "high",
			Title:       "Optimize Slow Operations",
			Description: "Some operations are exceeding performance thresholds",
			Impact:      "high",
			Effort:      "medium",
			Actions: []string{
				"Profile slow operations and identify bottlenecks",
				"Optimize database queries",
				"Implement caching for frequently accessed data",
				"Consider async processing for heavy operations",
			},
		})
	}

	// Database recommendations
	if report.DatabaseScore < 80 {
		report.Recommendations = append(report.Recommendations, OptimizationRecommendation{
			Type:        "database",
			Priority:    "high",
			Title:       "Optimize Database Performance",
			Description: "Database performance is below optimal levels",
			Impact:      "high",
			Effort:      "medium",
			Actions: []string{
				"Review and optimize slow queries",
				"Add missing database indexes",
				"Optimize connection pool settings",
				"Consider query result caching",
			},
		})
	}

	// Cache recommendations
	if report.CacheScore < 80 {
		report.Recommendations = append(report.Recommendations, OptimizationRecommendation{
			Type:        "cache",
			Priority:    "medium",
			Title:       "Improve Cache Performance",
			Description: "Cache hit rate is below optimal levels",
			Impact:      "medium",
			Effort:      "low",
			Actions: []string{
				"Review cache TTL settings",
				"Implement cache warming strategies",
				"Optimize cache key generation",
				"Consider cache preloading",
			},
		})
	}

	// Response time recommendations
	if report.ResponseTimeScore < 80 {
		report.Recommendations = append(report.Recommendations, OptimizationRecommendation{
			Type:        "response_time",
			Priority:    "high",
			Title:       "Improve Response Times",
			Description: "Response times are exceeding targets",
			Impact:      "high",
			Effort:      "high",
			Actions: []string{
				"Implement response time monitoring",
				"Optimize critical code paths",
				"Add request/response compression",
				"Consider CDN for static content",
			},
		})
	}
}

// applyOptimizations applies automatic optimizations
func (o *Optimizer) applyOptimizations(report *OptimizationReport) {
	// Apply database optimizations
	if o.dbOptimizer != nil {
		// Reset database statistics
		o.dbOptimizer.ResetStats()
		report.AppliedOptimizations = append(report.AppliedOptimizations, "database_stats_reset")
	}

	// Apply cache optimizations
	if o.cacheOptimizer != nil {
		// Clear expired cache entries
		report.AppliedOptimizations = append(report.AppliedOptimizations, "cache_cleanup")
	}

	// Apply profiler optimizations
	if o.profiler != nil {
		// Reset profiler statistics
		o.profiler.Reset()
		report.AppliedOptimizations = append(report.AppliedOptimizations, "profiler_reset")
	}
}

// calculateOverallScore calculates the overall optimization score
func (o *Optimizer) calculateOverallScore(report *OptimizationReport) {
	// Weighted average of all scores
	weights := map[string]float64{
		"performance":   0.3,
		"database":      0.25,
		"cache":         0.2,
		"response_time": 0.25,
	}

	score := 0.0
	totalWeight := 0.0

	if report.PerformanceScore > 0 {
		score += report.PerformanceScore * weights["performance"]
		totalWeight += weights["performance"]
	}

	if report.DatabaseScore > 0 {
		score += report.DatabaseScore * weights["database"]
		totalWeight += weights["database"]
	}

	if report.CacheScore > 0 {
		score += report.CacheScore * weights["cache"]
		totalWeight += weights["cache"]
	}

	if report.ResponseTimeScore > 0 {
		score += report.ResponseTimeScore * weights["response_time"]
		totalWeight += weights["response_time"]
	}

	if totalWeight > 0 {
		report.OverallScore = score / totalWeight
	} else {
		report.OverallScore = 0
	}
}

// optimize performs periodic optimization
func (o *Optimizer) optimize() {
	ticker := time.NewTicker(o.config.OptimizationInterval)
	defer ticker.Stop()

	for range ticker.C {
		report, err := o.Optimize()
		if err != nil {
			o.logger.Error("Optimization failed", zap.Error(err))
			continue
		}

		// Log optimization results
		o.logger.Info("Periodic optimization completed",
			zap.Float64("overall_score", report.OverallScore),
			zap.Float64("performance_score", report.PerformanceScore),
			zap.Float64("database_score", report.DatabaseScore),
			zap.Float64("cache_score", report.CacheScore),
			zap.Float64("response_time_score", report.ResponseTimeScore))

		// Alert if overall score is low
		if report.OverallScore < 70 {
			o.logger.Warn("Performance optimization score is low",
				zap.Float64("score", report.OverallScore),
				zap.Int("recommendations", len(report.Recommendations)))
		}
	}
}

// GetPerformanceReport generates a comprehensive performance report
func (o *Optimizer) GetPerformanceReport() string {
	report := "=== COMPREHENSIVE PERFORMANCE REPORT ===\n"
	report += fmt.Sprintf("Generated: %s\n", time.Now().Format(time.RFC3339))

	// Profiler report
	if o.profiler != nil {
		report += "\n" + o.profiler.GetPerformanceReport()
	}

	// Database report
	if o.dbOptimizer != nil {
		report += "\n" + o.dbOptimizer.GetPerformanceReport()
	}

	// Cache report
	if o.cacheOptimizer != nil {
		report += "\n" + o.cacheOptimizer.GetPerformanceReport()
	}

	// Response monitor report
	if o.responseMonitor != nil {
		report += "\n" + o.responseMonitor.GetPerformanceReport()
	}

	return report
}

// DefaultOptimizerConfig returns a default optimizer configuration
func DefaultOptimizerConfig() *OptimizerConfig {
	return &OptimizerConfig{
		EnableProfiling:          true,
		EnableDBOptimization:     true,
		EnableCaching:            true,
		EnableResponseMonitoring: true,
		PerformanceThreshold:     1 * time.Second,
		OptimizationInterval:     5 * time.Minute,
		EnableAutoOptimization:   true,
		TargetP95:                1 * time.Second,
		TargetP99:                2 * time.Second,
		TargetThroughput:         1000,
	}
}
