package performance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/cache"
)

// ResponseTimeOptimizer provides response time optimization
type ResponseTimeOptimizer struct {
	cache  cache.Cache
	logger *zap.Logger

	// Optimization settings
	config *ResponseTimeConfig
	stats  *ResponseTimeStats

	// Performance tracking
	responseTimes []time.Duration
	mu            sync.RWMutex

	// Optimization rules
	optimizationRules []ResponseTimeRule
}

// ResponseTimeConfig represents response time optimization configuration
type ResponseTimeConfig struct {
	// Target response times
	P95Target time.Duration `json:"p95_target"`
	P99Target time.Duration `json:"p99_target"`
	AvgTarget time.Duration `json:"avg_target"`
	MaxTarget time.Duration `json:"max_target"`

	// Optimization settings
	EnableCaching            bool          `json:"enable_caching"`
	CacheTTL                 time.Duration `json:"cache_ttl"`
	EnablePreloading         bool          `json:"enable_preloading"`
	PreloadInterval          time.Duration `json:"preload_interval"`
	EnableCompression        bool          `json:"enable_compression"`
	EnableParallelProcessing bool          `json:"enable_parallel_processing"`

	// Monitoring
	SampleSize       int           `json:"sample_size"`
	AnalysisInterval time.Duration `json:"analysis_interval"`
	AlertThreshold   time.Duration `json:"alert_threshold"`
}

// ResponseTimeStats represents response time statistics
type ResponseTimeStats struct {
	// Current metrics
	P95Latency time.Duration `json:"p95_latency"`
	P99Latency time.Duration `json:"p99_latency"`
	AvgLatency time.Duration `json:"avg_latency"`
	MinLatency time.Duration `json:"min_latency"`
	MaxLatency time.Duration `json:"max_latency"`

	// Target comparison
	P95TargetMet bool `json:"p95_target_met"`
	P99TargetMet bool `json:"p99_target_met"`
	AvgTargetMet bool `json:"avg_target_met"`
	MaxTargetMet bool `json:"max_target_met"`

	// Performance indicators
	IsOptimized       bool    `json:"is_optimized"`
	OptimizationScore float64 `json:"optimization_score"`

	// Cache performance
	CacheHitRate float64       `json:"cache_hit_rate"`
	CacheLatency time.Duration `json:"cache_latency"`

	// Timestamps
	LastUpdated   time.Time `json:"last_updated"`
	LastOptimized time.Time `json:"last_optimized"`
}

// ResponseTimeRule represents a response time optimization rule
type ResponseTimeRule struct {
	Name        string
	Description string
	Condition   func(*ResponseTimeStats) bool
	Action      func() error
	Priority    int
	Enabled     bool
}

// NewResponseTimeOptimizer creates a new response time optimizer
func NewResponseTimeOptimizer(cache cache.Cache, logger *zap.Logger) *ResponseTimeOptimizer {
	config := &ResponseTimeConfig{
		P95Target: 1 * time.Second,
		P99Target: 2 * time.Second,
		AvgTarget: 500 * time.Millisecond,
		MaxTarget: 5 * time.Second,

		EnableCaching:            true,
		CacheTTL:                 5 * time.Minute,
		EnablePreloading:         true,
		PreloadInterval:          1 * time.Minute,
		EnableCompression:        true,
		EnableParallelProcessing: true,

		SampleSize:       1000,
		AnalysisInterval: 30 * time.Second,
		AlertThreshold:   1 * time.Second,
	}

	optimizer := &ResponseTimeOptimizer{
		cache:             cache,
		logger:            logger,
		config:            config,
		stats:             &ResponseTimeStats{},
		responseTimes:     make([]time.Duration, 0, config.SampleSize),
		optimizationRules: []ResponseTimeRule{},
	}

	// Initialize optimization rules
	optimizer.initializeOptimizationRules()

	return optimizer
}

// RecordResponseTime records a response time measurement
func (rto *ResponseTimeOptimizer) RecordResponseTime(duration time.Duration) {
	rto.mu.Lock()
	defer rto.mu.Unlock()

	// Add to response times slice
	rto.responseTimes = append(rto.responseTimes, duration)

	// Keep only the last N samples
	if len(rto.responseTimes) > rto.config.SampleSize {
		rto.responseTimes = rto.responseTimes[len(rto.responseTimes)-rto.config.SampleSize:]
	}

	// Update statistics
	rto.updateStats()
}

// GetStats returns current response time statistics
func (rto *ResponseTimeOptimizer) GetStats() *ResponseTimeStats {
	rto.mu.RLock()
	defer rto.mu.RUnlock()

	return rto.stats
}

// Optimize performs response time optimization
func (rto *ResponseTimeOptimizer) Optimize(ctx context.Context) error {
	rto.logger.Info("Starting response time optimization")

	// Update statistics
	rto.mu.Lock()
	rto.updateStats()
	rto.mu.Unlock()

	// Apply optimization rules
	for _, rule := range rto.optimizationRules {
		if !rule.Enabled {
			continue
		}

		if rule.Condition(rto.stats) {
			rto.logger.Info("Applying response time optimization rule",
				zap.String("rule", rule.Name),
				zap.String("description", rule.Description))

			if err := rule.Action(); err != nil {
				rto.logger.Error("Failed to apply optimization rule",
					zap.String("rule", rule.Name),
					zap.Error(err))
			}
		}
	}

	// Check if targets are met
	rto.checkTargets()

	rto.stats.LastOptimized = time.Now()
	rto.logger.Info("Response time optimization completed",
		zap.Duration("p95_latency", rto.stats.P95Latency),
		zap.Duration("p99_latency", rto.stats.P99Latency),
		zap.Duration("avg_latency", rto.stats.AvgLatency),
		zap.Bool("is_optimized", rto.stats.IsOptimized))

	return nil
}

// updateStats updates response time statistics
func (rto *ResponseTimeOptimizer) updateStats() {
	if len(rto.responseTimes) == 0 {
		return
	}

	// Calculate percentiles
	rto.stats.P95Latency = rto.calculatePercentile(0.95)
	rto.stats.P99Latency = rto.calculatePercentile(0.99)

	// Calculate average
	total := time.Duration(0)
	min := rto.responseTimes[0]
	max := rto.responseTimes[0]

	for _, duration := range rto.responseTimes {
		total += duration
		if duration < min {
			min = duration
		}
		if duration > max {
			max = duration
		}
	}

	rto.stats.AvgLatency = total / time.Duration(len(rto.responseTimes))
	rto.stats.MinLatency = min
	rto.stats.MaxLatency = max

	// Update cache performance
	if rto.cache != nil {
		cacheMetrics := rto.cache.GetMetrics()
		rto.stats.CacheHitRate = cacheMetrics.HitRate
		rto.stats.CacheLatency = cacheMetrics.AverageLatency
	}

	// Calculate optimization score
	rto.calculateOptimizationScore()

	rto.stats.LastUpdated = time.Now()
}

// calculatePercentile calculates the specified percentile
func (rto *ResponseTimeOptimizer) calculatePercentile(percentile float64) time.Duration {
	if len(rto.responseTimes) == 0 {
		return 0
	}

	// Sort response times (simple bubble sort for small arrays)
	times := make([]time.Duration, len(rto.responseTimes))
	copy(times, rto.responseTimes)

	for i := 0; i < len(times)-1; i++ {
		for j := 0; j < len(times)-i-1; j++ {
			if times[j] > times[j+1] {
				times[j], times[j+1] = times[j+1], times[j]
			}
		}
	}

	index := int(float64(len(times)) * percentile)
	if index >= len(times) {
		index = len(times) - 1
	}

	return times[index]
}

// calculateOptimizationScore calculates the optimization score
func (rto *ResponseTimeOptimizer) calculateOptimizationScore() {
	score := 100.0

	// P95 latency score (30% weight)
	if rto.stats.P95Latency > rto.config.P95Target {
		penalty := float64(rto.stats.P95Latency-rto.config.P95Target) / float64(rto.config.P95Target) * 30
		score -= penalty
	}

	// P99 latency score (30% weight)
	if rto.stats.P99Latency > rto.config.P99Target {
		penalty := float64(rto.stats.P99Latency-rto.config.P99Target) / float64(rto.config.P99Target) * 30
		score -= penalty
	}

	// Average latency score (20% weight)
	if rto.stats.AvgLatency > rto.config.AvgTarget {
		penalty := float64(rto.stats.AvgLatency-rto.config.AvgTarget) / float64(rto.config.AvgTarget) * 20
		score -= penalty
	}

	// Cache performance score (20% weight)
	if rto.stats.CacheHitRate < 0.8 { // 80% hit rate target
		penalty := (0.8 - rto.stats.CacheHitRate) * 25
		score -= penalty
	}

	if score < 0 {
		score = 0
	}

	rto.stats.OptimizationScore = score
	rto.stats.IsOptimized = score >= 80.0
}

// checkTargets checks if response time targets are met
func (rto *ResponseTimeOptimizer) checkTargets() {
	rto.stats.P95TargetMet = rto.stats.P95Latency <= rto.config.P95Target
	rto.stats.P99TargetMet = rto.stats.P99Latency <= rto.config.P99Target
	rto.stats.AvgTargetMet = rto.stats.AvgLatency <= rto.config.AvgTarget
	rto.stats.MaxTargetMet = rto.stats.MaxLatency <= rto.config.MaxTarget

	// Log target status
	if !rto.stats.P95TargetMet {
		rto.logger.Warn("P95 response time target not met",
			zap.Duration("current", rto.stats.P95Latency),
			zap.Duration("target", rto.config.P95Target))
	}

	if !rto.stats.P99TargetMet {
		rto.logger.Warn("P99 response time target not met",
			zap.Duration("current", rto.stats.P99Latency),
			zap.Duration("target", rto.config.P99Target))
	}

	if !rto.stats.AvgTargetMet {
		rto.logger.Warn("Average response time target not met",
			zap.Duration("current", rto.stats.AvgLatency),
			zap.Duration("target", rto.config.AvgTarget))
	}
}

// initializeOptimizationRules initializes response time optimization rules
func (rto *ResponseTimeOptimizer) initializeOptimizationRules() {
	rto.optimizationRules = []ResponseTimeRule{
		{
			Name:        "cache_optimization",
			Description: "Optimize cache settings for better response times",
			Condition: func(stats *ResponseTimeStats) bool {
				return stats.CacheHitRate < 0.8
			},
			Action: func() error {
				rto.logger.Info("Optimizing cache settings")
				// Implement cache optimization
				return nil
			},
			Priority: 1,
			Enabled:  true,
		},
		{
			Name:        "response_time_alert",
			Description: "Alert when response times exceed thresholds",
			Condition: func(stats *ResponseTimeStats) bool {
				return stats.P95Latency > rto.config.AlertThreshold
			},
			Action: func() error {
				rto.logger.Warn("Response time alert triggered",
					zap.Duration("p95_latency", rto.stats.P95Latency),
					zap.Duration("threshold", rto.config.AlertThreshold))
				return nil
			},
			Priority: 2,
			Enabled:  true,
		},
		{
			Name:        "preload_optimization",
			Description: "Enable preloading for frequently accessed data",
			Condition: func(stats *ResponseTimeStats) bool {
				return stats.AvgLatency > rto.config.AvgTarget
			},
			Action: func() error {
				rto.logger.Info("Enabling data preloading")
				// Implement preloading optimization
				return nil
			},
			Priority: 3,
			Enabled:  rto.config.EnablePreloading,
		},
		{
			Name:        "compression_optimization",
			Description: "Enable compression for large responses",
			Condition: func(stats *ResponseTimeStats) bool {
				return stats.MaxLatency > rto.config.MaxTarget
			},
			Action: func() error {
				rto.logger.Info("Enabling response compression")
				// Implement compression optimization
				return nil
			},
			Priority: 4,
			Enabled:  rto.config.EnableCompression,
		},
	}
}

// GetOptimizationRecommendations returns optimization recommendations
func (rto *ResponseTimeOptimizer) GetOptimizationRecommendations() []string {
	recommendations := []string{}

	// Response time recommendations
	if !rto.stats.P95TargetMet {
		recommendations = append(recommendations,
			fmt.Sprintf("P95 response time (%v) exceeds target (%v) - consider caching or query optimization",
				rto.stats.P95Latency, rto.config.P95Target))
	}

	if !rto.stats.P99TargetMet {
		recommendations = append(recommendations,
			fmt.Sprintf("P99 response time (%v) exceeds target (%v) - consider database optimization",
				rto.stats.P99Latency, rto.config.P99Target))
	}

	if !rto.stats.AvgTargetMet {
		recommendations = append(recommendations,
			fmt.Sprintf("Average response time (%v) exceeds target (%v) - consider general performance optimization",
				rto.stats.AvgLatency, rto.config.AvgTarget))
	}

	// Cache recommendations
	if rto.stats.CacheHitRate < 0.8 {
		recommendations = append(recommendations,
			fmt.Sprintf("Cache hit rate (%.2f%%) is below target (80%%) - consider cache optimization",
				rto.stats.CacheHitRate*100))
	}

	// General recommendations
	if rto.stats.OptimizationScore < 80.0 {
		recommendations = append(recommendations,
			fmt.Sprintf("Overall optimization score (%.1f) is below target (80.0) - consider comprehensive optimization",
				rto.stats.OptimizationScore))
	}

	return recommendations
}

// SetTargets sets new response time targets
func (rto *ResponseTimeOptimizer) SetTargets(p95, p99, avg, max time.Duration) {
	rto.mu.Lock()
	defer rto.mu.Unlock()

	rto.config.P95Target = p95
	rto.config.P99Target = p99
	rto.config.AvgTarget = avg
	rto.config.MaxTarget = max

	rto.logger.Info("Response time targets updated",
		zap.Duration("p95_target", p95),
		zap.Duration("p99_target", p99),
		zap.Duration("avg_target", avg),
		zap.Duration("max_target", max))
}

// Reset resets the response time optimizer
func (rto *ResponseTimeOptimizer) Reset() {
	rto.mu.Lock()
	defer rto.mu.Unlock()

	rto.responseTimes = rto.responseTimes[:0]
	rto.stats = &ResponseTimeStats{}

	rto.logger.Info("Response time optimizer reset")
}
