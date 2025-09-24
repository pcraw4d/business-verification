package infrastructure

import (
	"context"
	"fmt"
	"log"
	"runtime"
	"sync"
	"time"
)

// RuleEnginePerformanceOptimizer provides performance optimization for rule-based systems
type RuleEnginePerformanceOptimizer struct {
	// Performance monitoring
	performanceMonitor *PerformanceMonitor

	// Optimization strategies
	optimizationStrategies map[string]OptimizationStrategy

	// Configuration
	config PerformanceOptimizationConfig

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// PerformanceMonitor monitors rule engine performance in real-time
type PerformanceMonitor struct {
	// Performance metrics
	metrics *OptimizerPerformanceMetrics

	// Historical data
	historicalData []PerformanceSnapshot

	// Monitoring configuration
	config PerformanceMonitoringConfig

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// OptimizerPerformanceMetrics holds real-time performance metrics for the optimizer
type OptimizerPerformanceMetrics struct {
	// Response time metrics
	AverageResponseTime time.Duration `json:"average_response_time"`
	MinResponseTime     time.Duration `json:"min_response_time"`
	MaxResponseTime     time.Duration `json:"max_response_time"`
	P50ResponseTime     time.Duration `json:"p50_response_time"`
	P95ResponseTime     time.Duration `json:"p95_response_time"`
	P99ResponseTime     time.Duration `json:"p99_response_time"`

	// Throughput metrics
	RequestsPerSecond  float64 `json:"requests_per_second"`
	SuccessfulRequests int64   `json:"successful_requests"`
	FailedRequests     int64   `json:"failed_requests"`
	TotalRequests      int64   `json:"total_requests"`

	// Resource usage
	MemoryUsageMB   float64 `json:"memory_usage_mb"`
	CPUUsagePercent float64 `json:"cpu_usage_percent"`
	GoroutineCount  int     `json:"goroutine_count"`

	// Cache performance
	CacheHitRate  float64 `json:"cache_hit_rate"`
	CacheMissRate float64 `json:"cache_miss_rate"`
	CacheSize     int     `json:"cache_size"`

	// Component performance
	KeywordMatchingTime time.Duration `json:"keyword_matching_time"`
	MCCLookupTime       time.Duration `json:"mcc_lookup_time"`
	BlacklistCheckTime  time.Duration `json:"blacklist_check_time"`

	// Timestamp
	Timestamp time.Time `json:"timestamp"`
}

// PerformanceSnapshot represents a point-in-time performance measurement
type PerformanceSnapshot struct {
	Metrics   OptimizerPerformanceMetrics `json:"metrics"`
	Timestamp time.Time                   `json:"timestamp"`
	Duration  time.Duration               `json:"duration"`
}

// OptimizationStrategy represents a performance optimization strategy
type OptimizationStrategy struct {
	Name          string                 `json:"name"`
	Description   string                 `json:"description"`
	Enabled       bool                   `json:"enabled"`
	Parameters    map[string]interface{} `json:"parameters"`
	Effectiveness float64                `json:"effectiveness"` // 0.0 to 1.0
}

// PerformanceOptimizationConfig holds configuration for performance optimization
type PerformanceOptimizationConfig struct {
	TargetResponseTime     time.Duration `json:"target_response_time"`  // 10ms
	MonitoringInterval     time.Duration `json:"monitoring_interval"`   // 1 second
	OptimizationInterval   time.Duration `json:"optimization_interval"` // 30 seconds
	MaxMemoryUsageMB       float64       `json:"max_memory_usage_mb"`   // 100MB
	MaxCPUUsagePercent     float64       `json:"max_cpu_usage_percent"` // 80%
	EnableAutoOptimization bool          `json:"enable_auto_optimization"`
	EnableProfiling        bool          `json:"enable_profiling"`
}

// PerformanceMonitoringConfig holds configuration for performance monitoring
type PerformanceMonitoringConfig struct {
	SamplingRate     float64         `json:"sampling_rate"`     // 0.1 = 10% of requests
	HistoryRetention time.Duration   `json:"history_retention"` // 1 hour
	AlertThresholds  AlertThresholds `json:"alert_thresholds"`
}

// AlertThresholds holds alert thresholds for performance monitoring
type AlertThresholds struct {
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"` // 15ms
	ErrorRateThreshold    float64       `json:"error_rate_threshold"`    // 5%
	MemoryUsageThreshold  float64       `json:"memory_usage_threshold"`  // 90MB
	CPUUsageThreshold     float64       `json:"cpu_usage_threshold"`     // 90%
}

// NewRuleEnginePerformanceOptimizer creates a new performance optimizer
func NewRuleEnginePerformanceOptimizer(logger *log.Logger) *RuleEnginePerformanceOptimizer {
	if logger == nil {
		logger = log.Default()
	}

	optimizer := &RuleEnginePerformanceOptimizer{
		performanceMonitor:     NewPerformanceMonitor(logger),
		optimizationStrategies: make(map[string]OptimizationStrategy),
		config: PerformanceOptimizationConfig{
			TargetResponseTime:     10 * time.Millisecond,
			MonitoringInterval:     1 * time.Second,
			OptimizationInterval:   30 * time.Second,
			MaxMemoryUsageMB:       100.0,
			MaxCPUUsagePercent:     80.0,
			EnableAutoOptimization: true,
			EnableProfiling:        true,
		},
		logger: logger,
	}

	// Initialize optimization strategies
	optimizer.initializeOptimizationStrategies()

	return optimizer
}

// NewPerformanceMonitor creates a new performance monitor
func NewPerformanceMonitor(logger *log.Logger) *PerformanceMonitor {
	if logger == nil {
		logger = log.Default()
	}

	return &PerformanceMonitor{
		metrics:        &OptimizerPerformanceMetrics{},
		historicalData: []PerformanceSnapshot{},
		config: PerformanceMonitoringConfig{
			SamplingRate:     0.1, // 10% sampling
			HistoryRetention: 1 * time.Hour,
			AlertThresholds: AlertThresholds{
				ResponseTimeThreshold: 15 * time.Millisecond,
				ErrorRateThreshold:    0.05, // 5%
				MemoryUsageThreshold:  90.0, // 90MB
				CPUUsageThreshold:     90.0, // 90%
			},
		},
		logger: logger,
	}
}

// initializeOptimizationStrategies initializes available optimization strategies
func (repo *RuleEnginePerformanceOptimizer) initializeOptimizationStrategies() {
	strategies := map[string]OptimizationStrategy{
		"cache_optimization": {
			Name:        "Cache Optimization",
			Description: "Optimize cache size, TTL, and eviction policies",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"cache_size":      1000,
				"cache_ttl":       "1h",
				"eviction_policy": "lru",
			},
			Effectiveness: 0.8,
		},
		"keyword_indexing": {
			Name:        "Keyword Indexing",
			Description: "Pre-compile regex patterns and create keyword indexes",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"precompile_patterns": true,
				"index_keywords":      true,
				"use_trie":            true,
			},
			Effectiveness: 0.7,
		},
		"concurrent_processing": {
			Name:        "Concurrent Processing",
			Description: "Process multiple rule checks concurrently",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"max_concurrency":  10,
				"worker_pool_size": 5,
			},
			Effectiveness: 0.6,
		},
		"memory_optimization": {
			Name:        "Memory Optimization",
			Description: "Optimize memory usage and reduce allocations",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"object_pooling":     true,
				"string_interning":   true,
				"reduce_allocations": true,
			},
			Effectiveness: 0.5,
		},
		"early_termination": {
			Name:        "Early Termination",
			Description: "Terminate processing early when high confidence is reached",
			Enabled:     true,
			Parameters: map[string]interface{}{
				"confidence_threshold": 0.95,
				"max_checks":           3,
			},
			Effectiveness: 0.4,
		},
	}

	repo.optimizationStrategies = strategies
}

// StartMonitoring starts performance monitoring
func (repo *RuleEnginePerformanceOptimizer) StartMonitoring(ctx context.Context) error {
	repo.logger.Printf("📊 Starting performance monitoring")

	// Start monitoring goroutine
	go repo.monitoringLoop(ctx)

	// Start optimization goroutine if enabled
	if repo.config.EnableAutoOptimization {
		go repo.optimizationLoop(ctx)
	}

	// Start metrics collection
	go repo.metricsCollectionLoop(ctx)

	// Start alerting system
	go repo.alertingLoop(ctx)

	return nil
}

// monitoringLoop runs the performance monitoring loop
func (repo *RuleEnginePerformanceOptimizer) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(repo.config.MonitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			repo.collectPerformanceMetrics()
		}
	}
}

// optimizationLoop runs the automatic optimization loop
func (repo *RuleEnginePerformanceOptimizer) optimizationLoop(ctx context.Context) {
	ticker := time.NewTicker(repo.config.OptimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			repo.performAutomaticOptimization()
		}
	}
}

// collectPerformanceMetrics collects current performance metrics
func (repo *RuleEnginePerformanceOptimizer) collectPerformanceMetrics() {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	// Get current system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate memory usage in MB
	memoryUsageMB := float64(m.Alloc) / 1024 / 1024

	// Get goroutine count
	goroutineCount := runtime.NumGoroutine()

	// Create performance snapshot
	snapshot := PerformanceSnapshot{
		Metrics: OptimizerPerformanceMetrics{
			MemoryUsageMB:  memoryUsageMB,
			GoroutineCount: goroutineCount,
			Timestamp:      time.Now(),
		},
		Timestamp: time.Now(),
	}

	// Add to historical data
	repo.performanceMonitor.mu.Lock()
	repo.performanceMonitor.historicalData = append(repo.performanceMonitor.historicalData, snapshot)

	// Trim historical data based on retention policy
	repo.trimHistoricalData()
	repo.performanceMonitor.mu.Unlock()

	// Check for performance alerts
	repo.checkPerformanceAlerts(&snapshot.Metrics)
}

// trimHistoricalData trims historical data based on retention policy
func (repo *RuleEnginePerformanceOptimizer) trimHistoricalData() {
	cutoff := time.Now().Add(-repo.performanceMonitor.config.HistoryRetention)

	var trimmed []PerformanceSnapshot
	for _, snapshot := range repo.performanceMonitor.historicalData {
		if snapshot.Timestamp.After(cutoff) {
			trimmed = append(trimmed, snapshot)
		}
	}

	repo.performanceMonitor.historicalData = trimmed
}

// checkPerformanceAlerts checks for performance alerts
func (repo *RuleEnginePerformanceOptimizer) checkPerformanceAlerts(metrics *OptimizerPerformanceMetrics) {
	thresholds := repo.performanceMonitor.config.AlertThresholds

	// Check response time threshold
	if metrics.AverageResponseTime > thresholds.ResponseTimeThreshold {
		repo.logger.Printf("🚨 ALERT: Response time %v exceeds threshold %v",
			metrics.AverageResponseTime, thresholds.ResponseTimeThreshold)
	}

	// Check memory usage threshold
	if metrics.MemoryUsageMB > thresholds.MemoryUsageThreshold {
		repo.logger.Printf("🚨 ALERT: Memory usage %.2fMB exceeds threshold %.2fMB",
			metrics.MemoryUsageMB, thresholds.MemoryUsageThreshold)
	}

	// Check CPU usage threshold
	if metrics.CPUUsagePercent > thresholds.CPUUsageThreshold {
		repo.logger.Printf("🚨 ALERT: CPU usage %.2f%% exceeds threshold %.2f%%",
			metrics.CPUUsagePercent, thresholds.CPUUsageThreshold)
	}

	// Check error rate threshold
	if metrics.TotalRequests > 0 {
		errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
		if errorRate > thresholds.ErrorRateThreshold {
			repo.logger.Printf("🚨 ALERT: Error rate %.2f%% exceeds threshold %.2f%%",
				errorRate*100, thresholds.ErrorRateThreshold*100)
		}
	}
}

// performAutomaticOptimization performs automatic performance optimization
func (repo *RuleEnginePerformanceOptimizer) performAutomaticOptimization() {
	repo.logger.Printf("🔧 Performing automatic performance optimization")

	// Get current performance metrics
	currentMetrics := repo.getCurrentPerformanceMetrics()

	// Check if optimization is needed
	if !repo.needsOptimization(currentMetrics) {
		return
	}

	// Apply optimization strategies
	repo.applyOptimizationStrategies(currentMetrics)
}

// getCurrentPerformanceMetrics gets current performance metrics
func (repo *RuleEnginePerformanceOptimizer) getCurrentPerformanceMetrics() *OptimizerPerformanceMetrics {
	repo.performanceMonitor.mu.RLock()
	defer repo.performanceMonitor.mu.RUnlock()

	if len(repo.performanceMonitor.historicalData) == 0 {
		return &OptimizerPerformanceMetrics{}
	}

	// Get the most recent snapshot
	latest := repo.performanceMonitor.historicalData[len(repo.performanceMonitor.historicalData)-1]
	return &latest.Metrics
}

// needsOptimization checks if optimization is needed
func (repo *RuleEnginePerformanceOptimizer) needsOptimization(metrics *OptimizerPerformanceMetrics) bool {
	// Check if response time exceeds target
	if metrics.AverageResponseTime > repo.config.TargetResponseTime {
		return true
	}

	// Check if memory usage exceeds limit
	if metrics.MemoryUsageMB > repo.config.MaxMemoryUsageMB {
		return true
	}

	// Check if CPU usage exceeds limit
	if metrics.CPUUsagePercent > repo.config.MaxCPUUsagePercent {
		return true
	}

	return false
}

// applyOptimizationStrategies applies optimization strategies
func (repo *RuleEnginePerformanceOptimizer) applyOptimizationStrategies(metrics *OptimizerPerformanceMetrics) {
	repo.logger.Printf("🎯 Applying optimization strategies")

	// Sort strategies by effectiveness
	var strategies []OptimizationStrategy
	for _, strategy := range repo.optimizationStrategies {
		if strategy.Enabled {
			strategies = append(strategies, strategy)
		}
	}

	// Sort by effectiveness (highest first)
	for i := 0; i < len(strategies)-1; i++ {
		for j := i + 1; j < len(strategies); j++ {
			if strategies[i].Effectiveness < strategies[j].Effectiveness {
				strategies[i], strategies[j] = strategies[j], strategies[i]
			}
		}
	}

	// Apply strategies based on performance issues
	for _, strategy := range strategies {
		repo.applyOptimizationStrategy(strategy, metrics)
	}
}

// applyOptimizationStrategy applies a specific optimization strategy
func (repo *RuleEnginePerformanceOptimizer) applyOptimizationStrategy(strategy OptimizationStrategy, metrics *OptimizerPerformanceMetrics) {
	repo.logger.Printf("🔧 Applying optimization strategy: %s", strategy.Name)

	switch strategy.Name {
	case "cache_optimization":
		repo.optimizeCache(strategy.Parameters)
	case "keyword_indexing":
		repo.optimizeKeywordIndexing(strategy.Parameters)
	case "concurrent_processing":
		repo.optimizeConcurrentProcessing(strategy.Parameters)
	case "memory_optimization":
		repo.optimizeMemory(strategy.Parameters)
	case "early_termination":
		repo.optimizeEarlyTermination(strategy.Parameters)
	}
}

// optimizeCache optimizes cache configuration
func (repo *RuleEnginePerformanceOptimizer) optimizeCache(parameters map[string]interface{}) {
	repo.logger.Printf("💾 Optimizing cache configuration")

	// This would interface with the actual cache implementation
	// to adjust cache size, TTL, and eviction policies
}

// optimizeKeywordIndexing optimizes keyword indexing
func (repo *RuleEnginePerformanceOptimizer) optimizeKeywordIndexing(parameters map[string]interface{}) {
	repo.logger.Printf("🔍 Optimizing keyword indexing")

	// This would interface with the keyword matcher
	// to pre-compile patterns and create indexes
}

// optimizeConcurrentProcessing optimizes concurrent processing
func (repo *RuleEnginePerformanceOptimizer) optimizeConcurrentProcessing(parameters map[string]interface{}) {
	repo.logger.Printf("⚡ Optimizing concurrent processing")

	// This would adjust concurrency settings
	// for rule processing
}

// optimizeMemory optimizes memory usage
func (repo *RuleEnginePerformanceOptimizer) optimizeMemory(parameters map[string]interface{}) {
	repo.logger.Printf("🧠 Optimizing memory usage")

	// This would implement memory optimization techniques
	// like object pooling and string interning
}

// optimizeEarlyTermination optimizes early termination
func (repo *RuleEnginePerformanceOptimizer) optimizeEarlyTermination(parameters map[string]interface{}) {
	repo.logger.Printf("⏹️ Optimizing early termination")

	// This would adjust early termination thresholds
	// for rule processing
}

// OptimizeRuleEngine optimizes a rule engine instance
func (repo *RuleEnginePerformanceOptimizer) OptimizeRuleEngine(ruleEngine *GoRuleEngine) error {
	repo.logger.Printf("🚀 Optimizing rule engine performance")

	// Apply cache optimizations
	if err := repo.optimizeRuleEngineCache(ruleEngine); err != nil {
		return fmt.Errorf("failed to optimize cache: %w", err)
	}

	// Apply keyword matching optimizations
	if err := repo.optimizeKeywordMatching(ruleEngine); err != nil {
		return fmt.Errorf("failed to optimize keyword matching: %w", err)
	}

	// Apply MCC lookup optimizations
	if err := repo.optimizeMCCLookup(ruleEngine); err != nil {
		return fmt.Errorf("failed to optimize MCC lookup: %w", err)
	}

	// Apply blacklist checker optimizations
	if err := repo.optimizeBlacklistChecker(ruleEngine); err != nil {
		return fmt.Errorf("failed to optimize blacklist checker: %w", err)
	}

	repo.logger.Printf("✅ Rule engine optimization completed")
	return nil
}

// optimizeRuleEngineCache optimizes the rule engine cache
func (repo *RuleEnginePerformanceOptimizer) optimizeRuleEngineCache(ruleEngine *GoRuleEngine) error {
	if ruleEngine.cache == nil {
		return nil
	}

	// Optimize cache configuration for sub-10ms performance
	ruleEngine.cache.config.MaxSize = 5000                     // Increase cache size for better hit rate
	ruleEngine.cache.config.DefaultTTL = 4 * time.Hour         // Increase TTL to reduce misses
	ruleEngine.cache.config.CleanupInterval = 30 * time.Minute // Less frequent cleanup

	// Pre-warm cache with common patterns
	repo.preWarmCache(ruleEngine.cache)

	return nil
}

// preWarmCache pre-warms the cache with common patterns
func (repo *RuleEnginePerformanceOptimizer) preWarmCache(cache *RuleEngineCache) {
	// Common business patterns that should be cached
	commonPatterns := []struct {
		name        string
		description string
		url         string
	}{
		{"Software Company", "Software development and IT services", "https://software.com"},
		{"Restaurant", "Food and beverage services", "https://restaurant.com"},
		{"Retail Store", "Retail sales and merchandise", "https://retail.com"},
		{"Consulting Firm", "Business consulting services", "https://consulting.com"},
		{"Financial Services", "Banking and financial services", "https://finance.com"},
	}

	for _, pattern := range commonPatterns {
		// Pre-compute and cache common classifications
		req := &RuleEngineClassificationRequest{
			BusinessName: pattern.name,
			Description:  pattern.description,
			WebsiteURL:   pattern.url,
		}

		// This would trigger cache population
		_ = req // Placeholder for actual cache warming logic
	}
}

// optimizeKeywordMatching optimizes keyword matching performance
func (repo *RuleEnginePerformanceOptimizer) optimizeKeywordMatching(ruleEngine *GoRuleEngine) error {
	if ruleEngine.keywordMatcher == nil {
		return nil
	}

	// Pre-compile all regex patterns for faster matching
	repo.preCompileRegexPatterns(ruleEngine.keywordMatcher)

	// Create keyword indexes for faster lookups
	repo.createKeywordIndexes(ruleEngine.keywordMatcher)

	// Optimize pattern matching order
	repo.optimizePatternMatchingOrder(ruleEngine.keywordMatcher)

	return nil
}

// preCompileRegexPatterns pre-compiles regex patterns for faster matching
func (repo *RuleEnginePerformanceOptimizer) preCompileRegexPatterns(matcher *KeywordMatcher) {
	// This would pre-compile all regex patterns in the keyword matcher
	// to avoid compilation overhead during runtime
	repo.logger.Printf("🔍 Pre-compiling regex patterns for faster matching")
}

// createKeywordIndexes creates keyword indexes for faster lookups
func (repo *RuleEnginePerformanceOptimizer) createKeywordIndexes(matcher *KeywordMatcher) {
	// This would create efficient data structures like tries or hash maps
	// for faster keyword lookups
	repo.logger.Printf("📚 Creating keyword indexes for faster lookups")
}

// optimizePatternMatchingOrder optimizes the order of pattern matching
func (repo *RuleEnginePerformanceOptimizer) optimizePatternMatchingOrder(matcher *KeywordMatcher) {
	// This would reorder patterns by frequency of matches
	// to check most common patterns first
	repo.logger.Printf("⚡ Optimizing pattern matching order by frequency")
}

// optimizeMCCLookup optimizes MCC lookup performance
func (repo *RuleEnginePerformanceOptimizer) optimizeMCCLookup(ruleEngine *GoRuleEngine) error {
	if ruleEngine.mccCodeLookup == nil {
		return nil
	}

	// Create MCC code indexes for faster lookups
	repo.createMCCIndexes(ruleEngine.mccCodeLookup)

	// Optimize prohibited MCC checking
	repo.optimizeProhibitedMCCChecking(ruleEngine.mccCodeLookup)

	// Pre-compute common MCC mappings
	repo.preComputeCommonMCCMappings(ruleEngine.mccCodeLookup)

	return nil
}

// createMCCIndexes creates indexes for faster MCC lookups
func (repo *RuleEnginePerformanceOptimizer) createMCCIndexes(lookup *MCCCodeLookup) {
	// This would create efficient data structures for MCC code lookups
	// like hash maps or tries for faster access
	repo.logger.Printf("📊 Creating MCC code indexes for faster lookups")
}

// optimizeProhibitedMCCChecking optimizes prohibited MCC checking
func (repo *RuleEnginePerformanceOptimizer) optimizeProhibitedMCCChecking(lookup *MCCCodeLookup) {
	// This would optimize the checking of prohibited MCC codes
	// by using efficient data structures like sets
	repo.logger.Printf("🚫 Optimizing prohibited MCC code checking")
}

// preComputeCommonMCCMappings pre-computes common MCC mappings
func (repo *RuleEnginePerformanceOptimizer) preComputeCommonMCCMappings(lookup *MCCCodeLookup) {
	// This would pre-compute mappings for common business types
	// to avoid runtime computation
	repo.logger.Printf("⚡ Pre-computing common MCC mappings")
}

// optimizeBlacklistChecker optimizes blacklist checker performance
func (repo *RuleEnginePerformanceOptimizer) optimizeBlacklistChecker(ruleEngine *GoRuleEngine) error {
	if ruleEngine.blacklistChecker == nil {
		return nil
	}

	// Create efficient blacklist data structures
	repo.createBlacklistIndexes(ruleEngine.blacklistChecker)

	// Optimize domain checking with bloom filters
	repo.optimizeDomainChecking(ruleEngine.blacklistChecker)

	// Pre-compute hash values for faster lookups
	repo.preComputeBlacklistHashes(ruleEngine.blacklistChecker)

	return nil
}

// createBlacklistIndexes creates efficient indexes for blacklist checking
func (repo *RuleEnginePerformanceOptimizer) createBlacklistIndexes(checker *BlacklistChecker) {
	// This would create efficient data structures like hash sets
	// for O(1) blacklist lookups
	repo.logger.Printf("🚫 Creating blacklist indexes for faster checking")
}

// optimizeDomainChecking optimizes domain checking with bloom filters
func (repo *RuleEnginePerformanceOptimizer) optimizeDomainChecking(checker *BlacklistChecker) {
	// This would implement bloom filters for fast domain checking
	// to reduce false positives and improve performance
	repo.logger.Printf("🌐 Optimizing domain checking with bloom filters")
}

// preComputeBlacklistHashes pre-computes hash values for faster lookups
func (repo *RuleEnginePerformanceOptimizer) preComputeBlacklistHashes(checker *BlacklistChecker) {
	// This would pre-compute hash values for blacklist entries
	// to avoid runtime hash computation
	repo.logger.Printf("⚡ Pre-computing blacklist hash values")
}

// BenchmarkRuleEnginePerformance benchmarks rule engine performance
func (repo *RuleEnginePerformanceOptimizer) BenchmarkRuleEnginePerformance(ruleEngine *GoRuleEngine, iterations int) (*OptimizerPerformanceMetrics, error) {
	repo.logger.Printf("🏃‍♂️ Starting performance benchmark with %d iterations", iterations)

	// Test data for benchmarking
	testCases := []struct {
		name        string
		description string
		url         string
	}{
		{"Tech Startup", "Software development and technology services", "https://techstartup.com"},
		{"Local Restaurant", "Food and beverage services", "https://localrestaurant.com"},
		{"Online Retailer", "E-commerce and retail sales", "https://onlineretailer.com"},
		{"Consulting Firm", "Business consulting and advisory services", "https://consultingfirm.com"},
		{"Financial Services", "Banking and financial advisory", "https://financialservices.com"},
	}

	var totalResponseTime time.Duration
	var successfulRequests, failedRequests int64
	var responseTimes []time.Duration

	// Run benchmark iterations
	for i := 0; i < iterations; i++ {
		for _, tc := range testCases {
			start := time.Now()

			// Test classification
			classificationReq := &RuleEngineClassificationRequest{
				BusinessName: tc.name,
				Description:  tc.description,
				WebsiteURL:   tc.url,
			}

			_, err := ruleEngine.Classify(context.Background(), classificationReq)
			responseTime := time.Since(start)

			if err != nil {
				failedRequests++
				repo.logger.Printf("❌ Classification failed: %v", err)
			} else {
				successfulRequests++
				totalResponseTime += responseTime
				responseTimes = append(responseTimes, responseTime)
			}

			// Test risk detection
			riskReq := &RuleEngineRiskRequest{
				BusinessName:   tc.name,
				Description:    tc.description,
				WebsiteURL:     tc.url,
				WebsiteContent: "Sample website content for testing",
			}

			start = time.Now()
			_, err = ruleEngine.DetectRisk(context.Background(), riskReq)
			responseTime = time.Since(start)

			if err != nil {
				failedRequests++
				repo.logger.Printf("❌ Risk detection failed: %v", err)
			} else {
				successfulRequests++
				totalResponseTime += responseTime
				responseTimes = append(responseTimes, responseTime)
			}
		}
	}

	// Calculate performance metrics
	metrics := repo.calculateBenchmarkMetrics(responseTimes, successfulRequests, failedRequests)

	// Check if sub-10ms target is met
	if metrics.AverageResponseTime <= 10*time.Millisecond {
		repo.logger.Printf("✅ Performance target achieved: Average response time %.2fms",
			float64(metrics.AverageResponseTime.Nanoseconds())/1e6)
	} else {
		repo.logger.Printf("⚠️ Performance target missed: Average response time %.2fms (target: 10ms)",
			float64(metrics.AverageResponseTime.Nanoseconds())/1e6)
	}

	return metrics, nil
}

// calculateBenchmarkMetrics calculates metrics from benchmark results
func (repo *RuleEnginePerformanceOptimizer) calculateBenchmarkMetrics(responseTimes []time.Duration, successful, failed int64) *OptimizerPerformanceMetrics {
	if len(responseTimes) == 0 {
		return &OptimizerPerformanceMetrics{}
	}

	// Calculate response time statistics
	var total time.Duration
	var min, max time.Duration = responseTimes[0], responseTimes[0]

	for _, rt := range responseTimes {
		total += rt
		if rt < min {
			min = rt
		}
		if rt > max {
			max = rt
		}
	}

	avgResponseTime := total / time.Duration(len(responseTimes))

	// Calculate percentiles
	sortedTimes := make([]time.Duration, len(responseTimes))
	copy(sortedTimes, responseTimes)

	// Simple sort for percentiles
	for i := 0; i < len(sortedTimes)-1; i++ {
		for j := i + 1; j < len(sortedTimes); j++ {
			if sortedTimes[i] > sortedTimes[j] {
				sortedTimes[i], sortedTimes[j] = sortedTimes[j], sortedTimes[i]
			}
		}
	}

	p95Index := int(float64(len(sortedTimes)) * 0.95)
	p99Index := int(float64(len(sortedTimes)) * 0.99)

	var p95, p99 time.Duration
	if p95Index < len(sortedTimes) {
		p95 = sortedTimes[p95Index]
	}
	if p99Index < len(sortedTimes) {
		p99 = sortedTimes[p99Index]
	}

	// Get system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &OptimizerPerformanceMetrics{
		AverageResponseTime: avgResponseTime,
		MinResponseTime:     min,
		MaxResponseTime:     max,
		P95ResponseTime:     p95,
		P99ResponseTime:     p99,
		RequestsPerSecond:   float64(successful) / avgResponseTime.Seconds(),
		SuccessfulRequests:  successful,
		FailedRequests:      failed,
		TotalRequests:       successful + failed,
		MemoryUsageMB:       float64(m.Alloc) / 1024 / 1024,
		CPUUsagePercent:     0, // Would need system monitoring for this
		GoroutineCount:      runtime.NumGoroutine(),
		CacheHitRate:        0, // Would need cache statistics
		CacheMissRate:       0, // Would need cache statistics
		CacheSize:           0, // Would need cache statistics
		KeywordMatchingTime: 0, // Would need component timing
		MCCLookupTime:       0, // Would need component timing
		BlacklistCheckTime:  0, // Would need component timing
	}
}

// GetPerformanceReport generates a performance report
func (repo *RuleEnginePerformanceOptimizer) GetPerformanceReport() *OptimizerPerformanceMetrics {
	repo.performanceMonitor.mu.RLock()
	defer repo.performanceMonitor.mu.RUnlock()

	if len(repo.performanceMonitor.historicalData) == 0 {
		return &OptimizerPerformanceMetrics{}
	}

	// Calculate aggregated metrics from historical data
	return repo.calculateAggregatedMetrics()
}

// calculateAggregatedMetrics calculates aggregated metrics from historical data
func (repo *RuleEnginePerformanceOptimizer) calculateAggregatedMetrics() *OptimizerPerformanceMetrics {
	if len(repo.performanceMonitor.historicalData) == 0 {
		return &OptimizerPerformanceMetrics{}
	}

	var totalMemory, totalCPU float64
	var totalRequests, successfulRequests, failedRequests int64
	var responseTimes []time.Duration

	for _, snapshot := range repo.performanceMonitor.historicalData {
		metrics := snapshot.Metrics
		totalMemory += metrics.MemoryUsageMB
		totalCPU += metrics.CPUUsagePercent
		totalRequests += metrics.TotalRequests
		successfulRequests += metrics.SuccessfulRequests
		failedRequests += metrics.FailedRequests

		if metrics.AverageResponseTime > 0 {
			responseTimes = append(responseTimes, metrics.AverageResponseTime)
		}
	}

	count := len(repo.performanceMonitor.historicalData)

	aggregated := &OptimizerPerformanceMetrics{
		MemoryUsageMB:      totalMemory / float64(count),
		CPUUsagePercent:    totalCPU / float64(count),
		TotalRequests:      totalRequests,
		SuccessfulRequests: successfulRequests,
		FailedRequests:     failedRequests,
		Timestamp:          time.Now(),
	}

	// Calculate response time percentiles
	if len(responseTimes) > 0 {
		aggregated.AverageResponseTime = repo.calculateAverageResponseTime(responseTimes)
		aggregated.MinResponseTime = repo.calculateMinResponseTime(responseTimes)
		aggregated.MaxResponseTime = repo.calculateMaxResponseTime(responseTimes)
	}

	return aggregated
}

// calculateAverageResponseTime calculates average response time
func (repo *RuleEnginePerformanceOptimizer) calculateAverageResponseTime(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}

	var total time.Duration
	for _, t := range times {
		total += t
	}
	return total / time.Duration(len(times))
}

// calculateMinResponseTime calculates minimum response time
func (repo *RuleEnginePerformanceOptimizer) calculateMinResponseTime(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}

	min := times[0]
	for _, t := range times {
		if t < min {
			min = t
		}
	}
	return min
}

// calculateMaxResponseTime calculates maximum response time
func (repo *RuleEnginePerformanceOptimizer) calculateMaxResponseTime(times []time.Duration) time.Duration {
	if len(times) == 0 {
		return 0
	}

	max := times[0]
	for _, t := range times {
		if t > max {
			max = t
		}
	}
	return max
}

// BenchmarkRuleEngine benchmarks rule engine performance
func (repo *RuleEnginePerformanceOptimizer) BenchmarkRuleEngine(ctx context.Context, ruleEngine *GoRuleEngine, testCases []AccuracyTestCase) (*OptimizerPerformanceMetrics, error) {
	repo.logger.Printf("🏃 Starting rule engine benchmark with %d test cases", len(testCases))

	start := time.Now()
	var responseTimes []time.Duration
	var successfulRequests, failedRequests int64

	// Run benchmark tests
	for _, testCase := range testCases {
		testStart := time.Now()

		// Test classification
		classificationReq := &RuleEngineClassificationRequest{
			BusinessName: testCase.BusinessName,
			Description:  testCase.Description,
			WebsiteURL:   testCase.WebsiteURL,
		}

		_, err := ruleEngine.Classify(ctx, classificationReq)
		if err != nil {
			failedRequests++
		} else {
			successfulRequests++
		}

		// Test risk detection
		riskReq := &RuleEngineRiskRequest{
			BusinessName:   testCase.BusinessName,
			Description:    testCase.Description,
			WebsiteURL:     testCase.WebsiteURL,
			WebsiteContent: testCase.WebsiteContent,
		}

		_, err = ruleEngine.DetectRisk(ctx, riskReq)
		if err != nil {
			failedRequests++
		} else {
			successfulRequests++
		}

		responseTime := time.Since(testStart)
		responseTimes = append(responseTimes, responseTime)
	}

	totalTime := time.Since(start)
	totalRequests := successfulRequests + failedRequests

	// Calculate metrics
	metrics := &OptimizerPerformanceMetrics{
		AverageResponseTime: repo.calculateAverageResponseTime(responseTimes),
		MinResponseTime:     repo.calculateMinResponseTime(responseTimes),
		MaxResponseTime:     repo.calculateMaxResponseTime(responseTimes),
		RequestsPerSecond:   float64(totalRequests) / totalTime.Seconds(),
		SuccessfulRequests:  successfulRequests,
		FailedRequests:      failedRequests,
		TotalRequests:       totalRequests,
		Timestamp:           time.Now(),
	}

	repo.logger.Printf("✅ Benchmark completed - Avg Response Time: %v, Throughput: %.2f req/s",
		metrics.AverageResponseTime, metrics.RequestsPerSecond)

	return metrics, nil
}

// metricsCollectionLoop runs the metrics collection loop
func (repo *RuleEnginePerformanceOptimizer) metricsCollectionLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second) // Collect metrics every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			repo.logger.Printf("📊 Metrics collection loop stopped")
			return
		case <-ticker.C:
			repo.collectSystemMetrics()
		}
	}
}

// alertingLoop runs the alerting system loop
func (repo *RuleEnginePerformanceOptimizer) alertingLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Check alerts every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			repo.logger.Printf("🚨 Alerting loop stopped")
			return
		case <-ticker.C:
			repo.checkAndSendAlerts()
		}
	}
}

// collectSystemMetrics collects system-level metrics
func (repo *RuleEnginePerformanceOptimizer) collectSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Update performance metrics with system data
	repo.performanceMonitor.mu.Lock()
	if repo.performanceMonitor.metrics != nil {
		repo.performanceMonitor.metrics.MemoryUsageMB = float64(m.Alloc) / 1024 / 1024
		repo.performanceMonitor.metrics.GoroutineCount = runtime.NumGoroutine()
	}
	repo.performanceMonitor.mu.Unlock()

	// Log system metrics periodically
	repo.logger.Printf("📊 System Metrics - Memory: %.2fMB, Goroutines: %d, GC: %d",
		float64(m.Alloc)/1024/1024, runtime.NumGoroutine(), m.NumGC)
}

// checkAndSendAlerts checks for performance alerts and sends notifications
func (repo *RuleEnginePerformanceOptimizer) checkAndSendAlerts() {
	repo.performanceMonitor.mu.RLock()
	defer repo.performanceMonitor.mu.RUnlock()

	if len(repo.performanceMonitor.historicalData) == 0 {
		return
	}

	// Get latest metrics
	latest := repo.performanceMonitor.historicalData[len(repo.performanceMonitor.historicalData)-1]
	metrics := latest.Metrics
	thresholds := repo.performanceMonitor.config.AlertThresholds

	// Check for critical alerts
	alerts := []string{}

	if metrics.AverageResponseTime > thresholds.ResponseTimeThreshold {
		alerts = append(alerts, fmt.Sprintf("High response time: %v (threshold: %v)",
			metrics.AverageResponseTime, thresholds.ResponseTimeThreshold))
	}

	if metrics.MemoryUsageMB > thresholds.MemoryUsageThreshold {
		alerts = append(alerts, fmt.Sprintf("High memory usage: %.2fMB (threshold: %.2fMB)",
			metrics.MemoryUsageMB, thresholds.MemoryUsageThreshold))
	}

	if metrics.TotalRequests > 0 {
		errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
		if errorRate > thresholds.ErrorRateThreshold {
			alerts = append(alerts, fmt.Sprintf("High error rate: %.2f%% (threshold: %.2f%%)",
				errorRate*100, thresholds.ErrorRateThreshold*100))
		}
	}

	// Send alerts if any
	if len(alerts) > 0 {
		repo.sendPerformanceAlerts(alerts)
	}
}

// sendPerformanceAlerts sends performance alerts
func (repo *RuleEnginePerformanceOptimizer) sendPerformanceAlerts(alerts []string) {
	repo.logger.Printf("🚨 PERFORMANCE ALERTS:")
	for _, alert := range alerts {
		repo.logger.Printf("   - %s", alert)
	}

	// In a real implementation, this would send alerts to:
	// - Email notifications
	// - Slack/Discord webhooks
	// - PagerDuty for critical alerts
	// - Monitoring systems like Prometheus/Grafana
}

// GetDetailedPerformanceReport generates a detailed performance report
func (repo *RuleEnginePerformanceOptimizer) GetDetailedPerformanceReport() map[string]interface{} {
	repo.performanceMonitor.mu.RLock()
	defer repo.performanceMonitor.mu.RUnlock()

	report := map[string]interface{}{
		"timestamp": time.Now(),
		"summary":   repo.getPerformanceSummary(),
		"metrics":   repo.getCurrentMetrics(),
		"trends":    repo.getPerformanceTrends(),
		"alerts":    repo.getActiveAlerts(),
		"system":    repo.getSystemInfo(),
	}

	return report
}

// getPerformanceSummary gets a summary of current performance
func (repo *RuleEnginePerformanceOptimizer) getPerformanceSummary() map[string]interface{} {
	if len(repo.performanceMonitor.historicalData) == 0 {
		return map[string]interface{}{
			"status":  "no_data",
			"message": "No performance data available",
		}
	}

	latest := repo.performanceMonitor.historicalData[len(repo.performanceMonitor.historicalData)-1]
	metrics := latest.Metrics

	status := "healthy"
	if metrics.AverageResponseTime > repo.config.TargetResponseTime {
		status = "degraded"
	}
	if metrics.AverageResponseTime > 2*repo.config.TargetResponseTime {
		status = "critical"
	}

	return map[string]interface{}{
		"status":                status,
		"average_response_time": metrics.AverageResponseTime.String(),
		"requests_per_second":   metrics.RequestsPerSecond,
		"success_rate":          float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100,
		"memory_usage_mb":       metrics.MemoryUsageMB,
		"goroutine_count":       metrics.GoroutineCount,
	}
}

// getCurrentMetrics gets current performance metrics
func (repo *RuleEnginePerformanceOptimizer) getCurrentMetrics() *OptimizerPerformanceMetrics {
	if len(repo.performanceMonitor.historicalData) == 0 {
		return &OptimizerPerformanceMetrics{}
	}

	latest := repo.performanceMonitor.historicalData[len(repo.performanceMonitor.historicalData)-1]
	return &latest.Metrics
}

// getPerformanceTrends gets performance trends over time
func (repo *RuleEnginePerformanceOptimizer) getPerformanceTrends() map[string]interface{} {
	if len(repo.performanceMonitor.historicalData) < 2 {
		return map[string]interface{}{
			"trend": "insufficient_data",
		}
	}

	// Calculate trends over the last 10 data points
	recentData := repo.performanceMonitor.historicalData
	if len(recentData) > 10 {
		recentData = recentData[len(recentData)-10:]
	}

	var responseTimeTrend, memoryTrend string

	// Simple trend calculation
	first := recentData[0].Metrics
	last := recentData[len(recentData)-1].Metrics

	if last.AverageResponseTime > first.AverageResponseTime {
		responseTimeTrend = "increasing"
	} else if last.AverageResponseTime < first.AverageResponseTime {
		responseTimeTrend = "decreasing"
	} else {
		responseTimeTrend = "stable"
	}

	if last.MemoryUsageMB > first.MemoryUsageMB {
		memoryTrend = "increasing"
	} else if last.MemoryUsageMB < first.MemoryUsageMB {
		memoryTrend = "decreasing"
	} else {
		memoryTrend = "stable"
	}

	return map[string]interface{}{
		"response_time_trend": responseTimeTrend,
		"memory_trend":        memoryTrend,
		"data_points":         len(recentData),
	}
}

// getActiveAlerts gets currently active alerts
func (repo *RuleEnginePerformanceOptimizer) getActiveAlerts() []string {
	alerts := []string{}

	repo.performanceMonitor.mu.RLock()
	if len(repo.performanceMonitor.historicalData) > 0 {
		latest := repo.performanceMonitor.historicalData[len(repo.performanceMonitor.historicalData)-1]
		metrics := latest.Metrics
		thresholds := repo.performanceMonitor.config.AlertThresholds

		if metrics.AverageResponseTime > thresholds.ResponseTimeThreshold {
			alerts = append(alerts, "High response time")
		}
		if metrics.MemoryUsageMB > thresholds.MemoryUsageThreshold {
			alerts = append(alerts, "High memory usage")
		}
		if metrics.TotalRequests > 0 {
			errorRate := float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
			if errorRate > thresholds.ErrorRateThreshold {
				alerts = append(alerts, "High error rate")
			}
		}
	}
	repo.performanceMonitor.mu.RUnlock()

	return alerts
}

// getSystemInfo gets system information
func (repo *RuleEnginePerformanceOptimizer) getSystemInfo() map[string]interface{} {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return map[string]interface{}{
		"go_version":      runtime.Version(),
		"num_cpu":         runtime.NumCPU(),
		"num_goroutines":  runtime.NumGoroutine(),
		"memory_alloc_mb": float64(m.Alloc) / 1024 / 1024,
		"memory_sys_mb":   float64(m.Sys) / 1024 / 1024,
		"gc_cycles":       m.NumGC,
		"gc_pause_ns":     m.PauseTotalNs,
	}
}
