package caching

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// OptimizationStrategy represents different cache optimization strategies
type OptimizationStrategy string

const (
	OptimizationStrategySizeAdjustment  OptimizationStrategy = "size_adjustment"
	OptimizationStrategyEvictionPolicy  OptimizationStrategy = "eviction_policy"
	OptimizationStrategyTTLOptimization OptimizationStrategy = "ttl_optimization"
	OptimizationStrategySharding        OptimizationStrategy = "sharding"
	OptimizationStrategyCompression     OptimizationStrategy = "compression"
)

// OptimizationAction represents a specific optimization action
type OptimizationAction struct {
	ID            string
	Strategy      OptimizationStrategy
	Description   string
	Parameters    map[string]interface{}
	Priority      int
	Impact        string
	Risk          string
	EstimatedGain float64
	EstimatedCost float64
	ROI           float64
}

// OptimizationResult represents the result of an optimization action
type OptimizationResult struct {
	ActionID        string
	Strategy        OptimizationStrategy
	Success         bool
	Error           error
	BeforeMetrics   map[string]float64
	AfterMetrics    map[string]float64
	Improvement     map[string]float64
	Duration        time.Duration
	Timestamp       time.Time
	Recommendations []string
}

// OptimizationPlan represents a comprehensive optimization plan
type OptimizationPlan struct {
	ID                 string
	Name               string
	Description        string
	Actions            []OptimizationAction
	Priority           int
	EstimatedTotalGain float64
	EstimatedTotalCost float64
	EstimatedROI       float64
	RiskLevel          string
	ExecutionTime      time.Duration
	CreatedAt          time.Time
	Status             string
}

// OptimizationConfig represents the configuration for cache optimization
type OptimizationConfig struct {
	Enabled              bool
	AutoOptimization     bool
	OptimizationInterval time.Duration
	MinImprovement       float64
	MaxRiskLevel         string
	Logger               *zap.Logger
}

// CacheOptimizer manages cache optimization strategies
type CacheOptimizer struct {
	cache   *IntelligentCache
	monitor *CacheMonitor
	config  OptimizationConfig
	mu      sync.RWMutex
	ctx     context.Context
	cancel  context.CancelFunc
	plans   []OptimizationPlan
	results []OptimizationResult
	lastRun time.Time
}

// NewCacheOptimizer creates a new cache optimizer
func NewCacheOptimizer(cache *IntelligentCache, monitor *CacheMonitor, config OptimizationConfig) *CacheOptimizer {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	if config.OptimizationInterval == 0 {
		config.OptimizationInterval = 1 * time.Hour
	}

	if config.MinImprovement == 0 {
		config.MinImprovement = 0.05
	}

	if config.MaxRiskLevel == "" {
		config.MaxRiskLevel = "medium"
	}

	ctx, cancel := context.WithCancel(context.Background())

	optimizer := &CacheOptimizer{
		cache:   cache,
		monitor: monitor,
		config:  config,
		ctx:     ctx,
		cancel:  cancel,
		plans:   make([]OptimizationPlan, 0),
		results: make([]OptimizationResult, 0),
	}

	if config.Enabled && config.AutoOptimization {
		go optimizer.optimizationWorker()
	}

	return optimizer
}

// GenerateOptimizationPlan generates an optimization plan based on current performance
func (co *CacheOptimizer) GenerateOptimizationPlan() (*OptimizationPlan, error) {
	co.mu.Lock()
	defer co.mu.Unlock()

	snapshot := co.monitor.GetCurrentSnapshot()
	if snapshot == nil {
		return nil, fmt.Errorf("no performance snapshot available")
	}

	actions := co.analyzePerformance(snapshot)
	if len(actions) == 0 {
		return nil, fmt.Errorf("no optimization actions identified")
	}

	totalGain := 0.0
	totalCost := 0.0
	maxRisk := "low"

	for _, action := range actions {
		totalGain += action.EstimatedGain
		totalCost += action.EstimatedCost
		if co.getRiskLevel(action.Risk) > co.getRiskLevel(maxRisk) {
			maxRisk = action.Risk
		}
	}

	roi := 0.0
	if totalCost > 0 {
		roi = totalGain / totalCost
	}

	plan := &OptimizationPlan{
		ID:                 generateOptimizationPlanID(),
		Name:               fmt.Sprintf("Optimization Plan %s", time.Now().Format("2006-01-02 15:04:05")),
		Description:        "Auto-generated optimization plan based on performance analysis",
		Actions:            actions,
		Priority:           co.calculatePlanPriority(actions),
		EstimatedTotalGain: totalGain,
		EstimatedTotalCost: totalCost,
		EstimatedROI:       roi,
		RiskLevel:          maxRisk,
		ExecutionTime:      co.estimateExecutionTime(actions),
		CreatedAt:          time.Now(),
		Status:             "pending",
	}

	co.plans = append(co.plans, *plan)
	return plan, nil
}

// ExecuteOptimizationPlan executes a specific optimization plan
func (co *CacheOptimizer) ExecuteOptimizationPlan(planID string) (*OptimizationResult, error) {
	co.mu.Lock()
	defer co.mu.Unlock()

	var plan *OptimizationPlan
	for i := range co.plans {
		if co.plans[i].ID == planID {
			plan = &co.plans[i]
			break
		}
	}

	if plan == nil {
		return nil, fmt.Errorf("optimization plan %s not found", planID)
	}

	if plan.Status != "pending" {
		return nil, fmt.Errorf("plan %s is not in pending status", planID)
	}

	plan.Status = "executing"
	var results []OptimizationResult
	startTime := time.Now()

	for _, action := range plan.Actions {
		result := co.executeAction(action)
		results = append(results, result)
		co.results = append(co.results, result)
	}

	plan.Status = "completed"
	if time.Since(startTime) > plan.ExecutionTime*2 {
		plan.Status = "failed"
	}

	overallResult := co.calculateOverallResult(results, plan)
	return overallResult, nil
}

// GetOptimizationPlans retrieves optimization plans
func (co *CacheOptimizer) GetOptimizationPlans() []OptimizationPlan {
	co.mu.RLock()
	defer co.mu.RUnlock()

	plans := make([]OptimizationPlan, len(co.plans))
	copy(plans, co.plans)
	return plans
}

// GetOptimizationResults retrieves optimization results
func (co *CacheOptimizer) GetOptimizationResults() []OptimizationResult {
	co.mu.RLock()
	defer co.mu.RUnlock()

	results := make([]OptimizationResult, len(co.results))
	copy(results, co.results)
	return results
}

// Close closes the cache optimizer
func (co *CacheOptimizer) Close() error {
	co.cancel()
	return nil
}

// optimizationWorker runs the background optimization worker
func (co *CacheOptimizer) optimizationWorker() {
	ticker := time.NewTicker(co.config.OptimizationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-co.ctx.Done():
			return
		case <-ticker.C:
			co.runAutoOptimization()
		}
	}
}

// runAutoOptimization runs automatic optimization
func (co *CacheOptimizer) runAutoOptimization() {
	plan, err := co.GenerateOptimizationPlan()
	if err != nil {
		co.config.Logger.Warn("Failed to generate optimization plan", zap.Error(err))
		return
	}

	if plan.EstimatedROI < 1.0 || plan.RiskLevel == "high" {
		co.config.Logger.Info("Skipping optimization - low ROI or high risk",
			zap.Float64("roi", plan.EstimatedROI),
			zap.String("risk", plan.RiskLevel))
		return
	}

	result, err := co.ExecuteOptimizationPlan(plan.ID)
	if err != nil {
		co.config.Logger.Error("Failed to execute optimization plan", zap.Error(err))
		return
	}

	co.config.Logger.Info("Auto-optimization completed",
		zap.String("plan_id", plan.ID),
		zap.Bool("success", result.Success),
		zap.Duration("duration", result.Duration))
}

// analyzePerformance analyzes current performance and generates optimization actions
func (co *CacheOptimizer) analyzePerformance(snapshot *CachePerformanceSnapshot) []OptimizationAction {
	var actions []OptimizationAction

	if snapshot.HitRate < 0.8 {
		actions = append(actions, co.generateHitRateOptimizations(snapshot)...)
	}

	if snapshot.MemoryUsage > 1<<30 {
		actions = append(actions, co.generateMemoryOptimizations(snapshot)...)
	}

	if snapshot.AverageLatency > 10*time.Millisecond {
		actions = append(actions, co.generateLatencyOptimizations(snapshot)...)
	}

	if snapshot.EvictionRate > 0.1 {
		actions = append(actions, co.generateEvictionOptimizations(snapshot)...)
	}

	return actions
}

// generateHitRateOptimizations generates optimizations for hit rate
func (co *CacheOptimizer) generateHitRateOptimizations(snapshot *CachePerformanceSnapshot) []OptimizationAction {
	var actions []OptimizationAction

	if snapshot.TotalSize < 1<<30 {
		actions = append(actions, OptimizationAction{
			ID:          generateOptimizationActionID(),
			Strategy:    OptimizationStrategySizeAdjustment,
			Description: "Increase cache size to improve hit rate",
			Parameters: map[string]interface{}{
				"new_size": snapshot.TotalSize * 2,
			},
			Priority:      1,
			Impact:        "high",
			Risk:          "low",
			EstimatedGain: 0.15,
			EstimatedCost: 0.05,
			ROI:           3.0,
		})
	}

	actions = append(actions, OptimizationAction{
		ID:          generateOptimizationActionID(),
		Strategy:    OptimizationStrategyTTLOptimization,
		Description: "Optimize TTL based on access patterns",
		Parameters: map[string]interface{}{
			"base_ttl":      time.Hour,
			"access_factor": 1.5,
		},
		Priority:      2,
		Impact:        "medium",
		Risk:          "low",
		EstimatedGain: 0.08,
		EstimatedCost: 0.02,
		ROI:           4.0,
	})

	return actions
}

// generateMemoryOptimizations generates optimizations for memory usage
func (co *CacheOptimizer) generateMemoryOptimizations(snapshot *CachePerformanceSnapshot) []OptimizationAction {
	var actions []OptimizationAction

	actions = append(actions, OptimizationAction{
		ID:          generateOptimizationActionID(),
		Strategy:    OptimizationStrategyCompression,
		Description: "Enable compression to reduce memory usage",
		Parameters: map[string]interface{}{
			"compression_level": 6,
			"min_size":          1024,
		},
		Priority:      1,
		Impact:        "high",
		Risk:          "low",
		EstimatedGain: 0.3,
		EstimatedCost: 0.1,
		ROI:           3.0,
	})

	return actions
}

// generateLatencyOptimizations generates optimizations for latency
func (co *CacheOptimizer) generateLatencyOptimizations(snapshot *CachePerformanceSnapshot) []OptimizationAction {
	var actions []OptimizationAction

	actions = append(actions, OptimizationAction{
		ID:          generateOptimizationActionID(),
		Strategy:    OptimizationStrategySharding,
		Description: "Increase shard count to reduce contention",
		Parameters: map[string]interface{}{
			"new_shard_count": snapshot.ShardCount * 2,
		},
		Priority:      1,
		Impact:        "high",
		Risk:          "medium",
		EstimatedGain: 0.25,
		EstimatedCost: 0.1,
		ROI:           2.5,
	})

	return actions
}

// generateEvictionOptimizations generates optimizations for eviction rate
func (co *CacheOptimizer) generateEvictionOptimizations(snapshot *CachePerformanceSnapshot) []OptimizationAction {
	var actions []OptimizationAction

	actions = append(actions, OptimizationAction{
		ID:          generateOptimizationActionID(),
		Strategy:    OptimizationStrategySizeAdjustment,
		Description: "Increase cache size to reduce evictions",
		Parameters: map[string]interface{}{
			"new_size": int64(float64(snapshot.TotalSize) * 1.5),
		},
		Priority:      1,
		Impact:        "high",
		Risk:          "low",
		EstimatedGain: 0.2,
		EstimatedCost: 0.1,
		ROI:           2.0,
	})

	return actions
}

// executeAction executes a specific optimization action
func (co *CacheOptimizer) executeAction(action OptimizationAction) OptimizationResult {
	startTime := time.Now()
	beforeMetrics := co.collectMetrics()

	var err error
	var success bool

	switch action.Strategy {
	case OptimizationStrategySizeAdjustment:
		success, err = co.executeSizeAdjustment(action)
	case OptimizationStrategyEvictionPolicy:
		success, err = co.executeEvictionPolicyChange(action)
	case OptimizationStrategyTTLOptimization:
		success, err = co.executeTTLOptimization(action)
	case OptimizationStrategySharding:
		success, err = co.executeShardingOptimization(action)
	case OptimizationStrategyCompression:
		success, err = co.executeCompressionOptimization(action)
	default:
		err = fmt.Errorf("unknown optimization strategy: %s", action.Strategy)
		success = false
	}

	afterMetrics := co.collectMetrics()
	improvement := co.calculateImprovement(beforeMetrics, afterMetrics)

	result := OptimizationResult{
		ActionID:        action.ID,
		Strategy:        action.Strategy,
		Success:         success,
		Error:           err,
		BeforeMetrics:   beforeMetrics,
		AfterMetrics:    afterMetrics,
		Improvement:     improvement,
		Duration:        time.Since(startTime),
		Timestamp:       time.Now(),
		Recommendations: co.generateRecommendations(action, improvement),
	}

	return result
}

// executeSizeAdjustment executes size adjustment optimization
func (co *CacheOptimizer) executeSizeAdjustment(action OptimizationAction) (bool, error) {
	newSize, ok := action.Parameters["new_size"].(int64)
	if !ok {
		return false, fmt.Errorf("invalid new_size parameter")
	}

	co.cache.config.MaxSize = newSize
	return true, nil
}

// executeEvictionPolicyChange executes eviction policy change
func (co *CacheOptimizer) executeEvictionPolicyChange(action OptimizationAction) (bool, error) {
	policy, ok := action.Parameters["policy"].(string)
	if !ok {
		return false, fmt.Errorf("invalid policy parameter")
	}

	switch policy {
	case "lru":
		co.cache.config.Type = CacheTypeLRU
	case "lfu":
		co.cache.config.Type = CacheTypeLFU
	case "fifo":
		co.cache.config.Type = CacheTypeFIFO
	default:
		return false, fmt.Errorf("unsupported policy: %s", policy)
	}

	return true, nil
}

// executeTTLOptimization executes TTL optimization
func (co *CacheOptimizer) executeTTLOptimization(action OptimizationAction) (bool, error) {
	baseTTL, ok := action.Parameters["base_ttl"].(time.Duration)
	if !ok {
		return false, fmt.Errorf("invalid base_ttl parameter")
	}

	co.cache.config.DefaultTTL = baseTTL
	return true, nil
}

// executeShardingOptimization executes sharding optimization
func (co *CacheOptimizer) executeShardingOptimization(action OptimizationAction) (bool, error) {
	newShardCount, ok := action.Parameters["new_shard_count"].(int)
	if !ok {
		return false, fmt.Errorf("invalid new_shard_count parameter")
	}

	co.cache.config.ShardCount = newShardCount
	return true, nil
}

// executeCompressionOptimization executes compression optimization
func (co *CacheOptimizer) executeCompressionOptimization(action OptimizationAction) (bool, error) {
	co.cache.config.Compression = true
	return true, nil
}

// collectMetrics collects current performance metrics
func (co *CacheOptimizer) collectMetrics() map[string]float64 {
	snapshot := co.monitor.GetCurrentSnapshot()
	if snapshot == nil {
		return make(map[string]float64)
	}

	return map[string]float64{
		"hit_rate":      snapshot.HitRate,
		"miss_rate":     snapshot.MissRate,
		"eviction_rate": snapshot.EvictionRate,
		"latency":       snapshot.AverageLatency.Seconds(),
		"memory_usage":  float64(snapshot.MemoryUsage),
		"throughput":    snapshot.Throughput,
	}
}

// calculateImprovement calculates improvement between before and after metrics
func (co *CacheOptimizer) calculateImprovement(before, after map[string]float64) map[string]float64 {
	improvement := make(map[string]float64)

	for key, beforeValue := range before {
		if afterValue, exists := after[key]; exists && beforeValue > 0 {
			improvement[key] = (afterValue - beforeValue) / beforeValue
		}
	}

	return improvement
}

// calculateOverallResult calculates overall optimization result
func (co *CacheOptimizer) calculateOverallResult(results []OptimizationResult, plan *OptimizationPlan) *OptimizationResult {
	successCount := 0
	totalDuration := time.Duration(0)
	var lastError error

	for _, result := range results {
		if result.Success {
			successCount++
		} else {
			lastError = result.Error
		}
		totalDuration += result.Duration
	}

	success := successCount == len(results)
	if !success && lastError != nil {
		lastError = fmt.Errorf("some optimizations failed: %w", lastError)
	}

	return &OptimizationResult{
		ActionID:  plan.ID,
		Strategy:  OptimizationStrategy("plan"),
		Success:   success,
		Error:     lastError,
		Duration:  totalDuration,
		Timestamp: time.Now(),
	}
}

// calculatePlanPriority calculates the priority of an optimization plan
func (co *CacheOptimizer) calculatePlanPriority(actions []OptimizationAction) int {
	if len(actions) == 0 {
		return 0
	}

	totalPriority := 0
	for _, action := range actions {
		totalPriority += action.Priority
	}

	return totalPriority / len(actions)
}

// estimateExecutionTime estimates the execution time for a plan
func (co *CacheOptimizer) estimateExecutionTime(actions []OptimizationAction) time.Duration {
	totalTime := time.Duration(0)
	for _, action := range actions {
		switch action.Strategy {
		case OptimizationStrategySizeAdjustment:
			totalTime += 5 * time.Second
		case OptimizationStrategyEvictionPolicy:
			totalTime += 10 * time.Second
		case OptimizationStrategyTTLOptimization:
			totalTime += 3 * time.Second
		case OptimizationStrategySharding:
			totalTime += 30 * time.Second
		case OptimizationStrategyCompression:
			totalTime += 15 * time.Second
		default:
			totalTime += 10 * time.Second
		}
	}

	return totalTime
}

// getRiskLevel converts risk string to numeric level
func (co *CacheOptimizer) getRiskLevel(risk string) int {
	switch risk {
	case "low":
		return 1
	case "medium":
		return 2
	case "high":
		return 3
	default:
		return 1
	}
}

// generateRecommendations generates recommendations based on optimization results
func (co *CacheOptimizer) generateRecommendations(action OptimizationAction, improvement map[string]float64) []string {
	var recommendations []string

	if hitRateImprovement, exists := improvement["hit_rate"]; exists && hitRateImprovement > 0.05 {
		recommendations = append(recommendations, "Significant hit rate improvement achieved")
	}

	if latencyImprovement, exists := improvement["latency"]; exists && latencyImprovement < -0.1 {
		recommendations = append(recommendations, "Latency significantly reduced")
	}

	if memoryImprovement, exists := improvement["memory_usage"]; exists && memoryImprovement < -0.1 {
		recommendations = append(recommendations, "Memory usage significantly reduced")
	}

	switch action.Strategy {
	case OptimizationStrategySizeAdjustment:
		recommendations = append(recommendations, "Consider monitoring memory usage after size increase")
	case OptimizationStrategyEvictionPolicy:
		recommendations = append(recommendations, "Monitor hit rate changes with new eviction policy")
	case OptimizationStrategySharding:
		recommendations = append(recommendations, "Verify load distribution across shards")
	}

	return recommendations
}

// generateOptimizationPlanID generates a unique optimization plan ID
func generateOptimizationPlanID() string {
	return fmt.Sprintf("plan_%d", time.Now().UnixNano())
}

// generateOptimizationActionID generates a unique optimization action ID
func generateOptimizationActionID() string {
	return fmt.Sprintf("action_%d", time.Now().UnixNano())
}
