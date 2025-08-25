package caching

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewCacheOptimizer(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	config := OptimizationConfig{
		Enabled: true,
		Logger:  zap.NewNop(),
	}

	optimizer := NewCacheOptimizer(cache, monitor, config)
	assert.NotNil(t, optimizer)
	assert.Equal(t, cache, optimizer.cache)
	assert.Equal(t, monitor, optimizer.monitor)
	assert.Equal(t, 1*time.Hour, optimizer.config.OptimizationInterval)
	assert.Equal(t, 0.05, optimizer.config.MinImprovement)
	assert.Equal(t, "medium", optimizer.config.MaxRiskLevel)
	assert.True(t, optimizer.config.Enabled)

	optimizer.Close()
}

func TestCacheOptimizer_GenerateOptimizationPlan(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	optimizer := NewCacheOptimizer(cache, monitor, OptimizationConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer optimizer.Close()

	t.Run("generate plan with performance issues", func(t *testing.T) {
		snapshot := &CachePerformanceSnapshot{
			Timestamp:      time.Now(),
			HitRate:        0.6, // Low hit rate
			MissRate:       0.4,
			EvictionRate:   0.2,        // High eviction rate
			TotalSize:      512 * 1024, // Small size
			EntryCount:     100,
			AverageLatency: 20 * time.Millisecond, // High latency
			MemoryUsage:    2 << 30,               // High memory usage
			Throughput:     1000,
			ShardCount:     2, // Low shard count
		}

		monitor.mu.Lock()
		monitor.lastSnapshot = snapshot
		monitor.mu.Unlock()

		plan, err := optimizer.GenerateOptimizationPlan()
		assert.NoError(t, err)
		assert.NotNil(t, plan)
		assert.NotEmpty(t, plan.ID)
		assert.NotEmpty(t, plan.Name)
		assert.NotEmpty(t, plan.Description)
		assert.NotEmpty(t, plan.Actions)
		assert.Equal(t, "pending", plan.Status)
		assert.Greater(t, plan.EstimatedROI, 0.0)
		assert.NotEmpty(t, plan.RiskLevel)
	})

	t.Run("generate plan with good performance", func(t *testing.T) {
		snapshot := &CachePerformanceSnapshot{
			Timestamp:      time.Now(),
			HitRate:        0.95, // High hit rate
			MissRate:       0.05,
			EvictionRate:   0.01,    // Low eviction rate
			TotalSize:      2 << 30, // Large size
			EntryCount:     1000,
			AverageLatency: 2 * time.Millisecond, // Low latency
			MemoryUsage:    512 * 1024,           // Low memory usage
			Throughput:     5000,
			ShardCount:     8, // High shard count
		}

		monitor.mu.Lock()
		monitor.lastSnapshot = snapshot
		monitor.mu.Unlock()

		plan, err := optimizer.GenerateOptimizationPlan()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no optimization actions identified")
		assert.Nil(t, plan)
	})
}

func TestCacheOptimizer_ExecuteOptimizationPlan(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	optimizer := NewCacheOptimizer(cache, monitor, OptimizationConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer optimizer.Close()

	t.Run("execute valid optimization plan", func(t *testing.T) {
		plan := &OptimizationPlan{
			ID:          "test_plan_1",
			Name:        "Test Plan",
			Description: "Test optimization plan",
			Actions: []OptimizationAction{
				{
					ID:          "action_1",
					Strategy:    OptimizationStrategySizeAdjustment,
					Description: "Increase cache size",
					Parameters: map[string]interface{}{
						"new_size": int64(2 * 1024 * 1024),
					},
					Priority:      1,
					Impact:        "high",
					Risk:          "low",
					EstimatedGain: 0.15,
					EstimatedCost: 0.05,
					ROI:           3.0,
				},
			},
			EstimatedTotalGain: 0.15,
			EstimatedTotalCost: 0.05,
			EstimatedROI:       3.0,
			RiskLevel:          "low",
			ExecutionTime:      10 * time.Second,
			CreatedAt:          time.Now(),
			Status:             "pending",
		}

		optimizer.mu.Lock()
		optimizer.plans = append(optimizer.plans, *plan)
		optimizer.mu.Unlock()

		result, err := optimizer.ExecuteOptimizationPlan("test_plan_1")
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, "test_plan_1", result.ActionID)
		assert.NotZero(t, result.Duration)
		assert.NotZero(t, result.Timestamp)
	})

	t.Run("execute non-existent plan", func(t *testing.T) {
		result, err := optimizer.ExecuteOptimizationPlan("non_existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		assert.Nil(t, result)
	})
}

func TestCacheOptimizer_GetMethods(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	optimizer := NewCacheOptimizer(cache, monitor, OptimizationConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer optimizer.Close()

	t.Run("get optimization plans", func(t *testing.T) {
		plan1 := OptimizationPlan{ID: "plan1", Name: "Test Plan 1"}
		plan2 := OptimizationPlan{ID: "plan2", Name: "Test Plan 2"}

		optimizer.mu.Lock()
		optimizer.plans = append(optimizer.plans, plan1, plan2)
		optimizer.mu.Unlock()

		plans := optimizer.GetOptimizationPlans()
		assert.Len(t, plans, 2)
		assert.Equal(t, "plan1", plans[0].ID)
		assert.Equal(t, "plan2", plans[1].ID)
	})

	t.Run("get optimization results", func(t *testing.T) {
		result1 := OptimizationResult{ActionID: "action1", Strategy: OptimizationStrategySizeAdjustment}
		result2 := OptimizationResult{ActionID: "action2", Strategy: OptimizationStrategyTTLOptimization}

		optimizer.mu.Lock()
		optimizer.results = append(optimizer.results, result1, result2)
		optimizer.mu.Unlock()

		results := optimizer.GetOptimizationResults()
		assert.Len(t, results, 2)
		assert.Equal(t, "action1", results[0].ActionID)
		assert.Equal(t, "action2", results[1].ActionID)
	})
}

func TestCacheOptimizer_Close(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	optimizer := NewCacheOptimizer(cache, monitor, OptimizationConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer optimizer.Close()

	t.Run("close optimizer", func(t *testing.T) {
		err := optimizer.Close()
		assert.NoError(t, err)
	})
}

// Benchmark tests
func BenchmarkCacheOptimizer_GenerateOptimizationPlan(b *testing.B) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(b, err)
	defer cache.Close()

	monitor := NewCacheMonitor(cache, CacheMonitorConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer monitor.Close()

	optimizer := NewCacheOptimizer(cache, monitor, OptimizationConfig{
		Enabled: false,
		Logger:  zap.NewNop(),
	})
	defer optimizer.Close()

	snapshot := &CachePerformanceSnapshot{
		Timestamp:      time.Now(),
		HitRate:        0.6,
		MemoryUsage:    2 << 30,
		AverageLatency: 20 * time.Millisecond,
		EvictionRate:   0.2,
		TotalSize:      512 * 1024,
		ShardCount:     2,
	}

	monitor.mu.Lock()
	monitor.lastSnapshot = snapshot
	monitor.mu.Unlock()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			optimizer.GenerateOptimizationPlan()
		}
	})
}
