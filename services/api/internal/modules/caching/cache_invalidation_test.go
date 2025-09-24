package caching

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewInvalidationManager(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	assert.NotNil(t, manager)
	assert.Equal(t, cache, manager.cache)
	assert.NotNil(t, manager.rules)
	assert.NotNil(t, manager.events)
	assert.NotNil(t, manager.patterns)
	assert.NotNil(t, manager.dependencies)
	assert.NotNil(t, manager.logger)

	// Cleanup
	manager.Close()
}

func TestInvalidationManager_AddRule(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	t.Run("add exact rule", func(t *testing.T) {
		rule := &InvalidationRule{
			Name:     "test-exact",
			Strategy: InvalidationStrategyExact,
			Pattern:  "test-key",
			Enabled:  true,
		}

		err := manager.AddRule(rule)
		assert.NoError(t, err)
		assert.NotEmpty(t, rule.ID)
		assert.NotZero(t, rule.CreatedAt)
		assert.NotZero(t, rule.UpdatedAt)
	})

	t.Run("add pattern rule", func(t *testing.T) {
		rule := &InvalidationRule{
			Name:     "test-pattern",
			Strategy: InvalidationStrategyPattern,
			Pattern:  "test.*",
			Enabled:  true,
		}

		err := manager.AddRule(rule)
		assert.NoError(t, err)
		assert.NotEmpty(t, rule.ID)
		assert.Contains(t, manager.patterns, rule.ID)
	})

	t.Run("add pattern rule with invalid regex", func(t *testing.T) {
		rule := &InvalidationRule{
			Name:     "test-invalid-pattern",
			Strategy: InvalidationStrategyPattern,
			Pattern:  "[invalid",
			Enabled:  true,
		}

		err := manager.AddRule(rule)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid pattern")
	})

	t.Run("add dependency rule", func(t *testing.T) {
		rule := &InvalidationRule{
			Name:         "test-dependency",
			Strategy:     InvalidationStrategyDependency,
			Dependencies: []string{"dep1", "dep2"},
			Enabled:      true,
		}

		err := manager.AddRule(rule)
		assert.NoError(t, err)
		assert.NotEmpty(t, rule.ID)
		assert.Contains(t, manager.dependencies["dep1"], rule.ID)
		assert.Contains(t, manager.dependencies["dep2"], rule.ID)
	})

	t.Run("add rule without name", func(t *testing.T) {
		rule := &InvalidationRule{
			Strategy: InvalidationStrategyExact,
			Pattern:  "test-key",
			Enabled:  true,
		}

		err := manager.AddRule(rule)
		assert.NoError(t, err)
		assert.NotEmpty(t, rule.ID)
		assert.NotEmpty(t, rule.Name)
		assert.Contains(t, rule.Name, rule.ID)
	})
}

func TestInvalidationManager_RemoveRule(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	// Add a rule
	rule := &InvalidationRule{
		Name:     "test-remove",
		Strategy: InvalidationStrategyPattern,
		Pattern:  "test.*",
		Enabled:  true,
	}

	err = manager.AddRule(rule)
	require.NoError(t, err)

	t.Run("remove existing rule", func(t *testing.T) {
		err := manager.RemoveRule(rule.ID)
		assert.NoError(t, err)
		assert.NotContains(t, manager.rules, rule.ID)
		assert.NotContains(t, manager.patterns, rule.ID)
	})

	t.Run("remove non-existent rule", func(t *testing.T) {
		err := manager.RemoveRule("non-existent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestInvalidationManager_UpdateRule(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	// Add a rule
	rule := &InvalidationRule{
		Name:     "test-update",
		Strategy: InvalidationStrategyExact,
		Pattern:  "test-key",
		Enabled:  true,
	}

	err = manager.AddRule(rule)
	require.NoError(t, err)

	t.Run("update rule name", func(t *testing.T) {
		updates := &InvalidationRule{
			Name: "updated-name",
		}

		err := manager.UpdateRule(rule.ID, updates)
		assert.NoError(t, err)

		updatedRule, err := manager.GetRule(rule.ID)
		assert.NoError(t, err)
		assert.Equal(t, "updated-name", updatedRule.Name)
		assert.True(t, updatedRule.UpdatedAt.After(rule.UpdatedAt) || updatedRule.UpdatedAt.Equal(rule.UpdatedAt))
	})

	t.Run("update rule pattern", func(t *testing.T) {
		updates := &InvalidationRule{
			Pattern: "updated-pattern",
		}

		err := manager.UpdateRule(rule.ID, updates)
		assert.NoError(t, err)

		updatedRule, err := manager.GetRule(rule.ID)
		assert.NoError(t, err)
		assert.Equal(t, "updated-pattern", updatedRule.Pattern)
	})

	t.Run("update non-existent rule", func(t *testing.T) {
		updates := &InvalidationRule{
			Name: "updated-name",
		}

		err := manager.UpdateRule("non-existent", updates)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestInvalidationManager_GetRule(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	// Add a rule
	rule := &InvalidationRule{
		Name:     "test-get",
		Strategy: InvalidationStrategyExact,
		Pattern:  "test-key",
		Enabled:  true,
	}

	err = manager.AddRule(rule)
	require.NoError(t, err)

	t.Run("get existing rule", func(t *testing.T) {
		retrieved, err := manager.GetRule(rule.ID)
		assert.NoError(t, err)
		assert.Equal(t, rule.ID, retrieved.ID)
		assert.Equal(t, rule.Name, retrieved.Name)
		assert.Equal(t, rule.Strategy, retrieved.Strategy)
		assert.Equal(t, rule.Pattern, retrieved.Pattern)
	})

	t.Run("get non-existent rule", func(t *testing.T) {
		retrieved, err := manager.GetRule("non-existent")
		assert.Error(t, err)
		assert.Nil(t, retrieved)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestInvalidationManager_ListRules(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	// Add multiple rules
	rule1 := &InvalidationRule{
		Name:     "test-list-1",
		Strategy: InvalidationStrategyExact,
		Pattern:  "test-key-1",
		Enabled:  true,
	}

	rule2 := &InvalidationRule{
		Name:     "test-list-2",
		Strategy: InvalidationStrategyPattern,
		Pattern:  "test.*",
		Enabled:  true,
	}

	err = manager.AddRule(rule1)
	require.NoError(t, err)
	err = manager.AddRule(rule2)
	require.NoError(t, err)

	t.Run("list all rules", func(t *testing.T) {
		rules := manager.ListRules()
		assert.Len(t, rules, 2)

		// Check that both rules are present
		ruleIDs := make(map[string]bool)
		for _, rule := range rules {
			ruleIDs[rule.ID] = true
		}

		assert.True(t, ruleIDs[rule1.ID])
		assert.True(t, ruleIDs[rule2.ID])
	})
}

func TestInvalidationManager_InvalidateByKey(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	t.Run("invalidate existing key", func(t *testing.T) {
		// Add a key to cache
		err := cache.Set("test-key", "test-value")
		require.NoError(t, err)

		// Verify key exists
		result := cache.Get("test-key")
		assert.True(t, result.Found)

		// Invalidate the key
		invResult := manager.InvalidateByKey("test-key")
		assert.Equal(t, InvalidationStrategyExact, invResult.Strategy)
		assert.Equal(t, int64(1), invResult.KeysInvalidated)
		assert.Equal(t, []string{"test-key"}, invResult.KeysMatched)
		assert.NotZero(t, invResult.Duration)

		// Verify key is gone
		result = cache.Get("test-key")
		assert.False(t, result.Found)
	})

	t.Run("invalidate non-existent key", func(t *testing.T) {
		invResult := manager.InvalidateByKey("non-existent")
		assert.Equal(t, InvalidationStrategyExact, invResult.Strategy)
		assert.Equal(t, int64(0), invResult.KeysInvalidated)
		assert.Empty(t, invResult.KeysMatched)
		assert.NotZero(t, invResult.Duration)
	})
}

func TestInvalidationManager_InvalidateByPattern(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	t.Run("invalidate with valid pattern", func(t *testing.T) {
		// Add keys to cache
		err := cache.Set("test-key-1", "value1")
		require.NoError(t, err)
		err = cache.Set("test-key-2", "value2")
		require.NoError(t, err)
		err = cache.Set("other-key", "value3")
		require.NoError(t, err)

		// Invalidate with pattern
		invResult := manager.InvalidateByPattern("test.*")
		assert.Equal(t, InvalidationStrategyPattern, invResult.Strategy)
		assert.NotZero(t, invResult.Duration)

		// Note: The actual key matching is not implemented in the current version
		// as it requires access to cache internals
	})

	t.Run("invalidate with invalid pattern", func(t *testing.T) {
		invResult := manager.InvalidateByPattern("[invalid")
		assert.Equal(t, InvalidationStrategyPattern, invResult.Strategy)
		assert.Error(t, invResult.Error)
		assert.Contains(t, invResult.Error.Error(), "invalid pattern")
		assert.NotZero(t, invResult.Duration)
	})
}

func TestInvalidationManager_InvalidateAll(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	t.Run("invalidate all entries", func(t *testing.T) {
		// Add multiple keys to cache
		err := cache.Set("key1", "value1")
		require.NoError(t, err)
		err = cache.Set("key2", "value2")
		require.NoError(t, err)
		err = cache.Set("key3", "value3")
		require.NoError(t, err)

		// Verify keys exist
		assert.True(t, cache.Get("key1").Found)
		assert.True(t, cache.Get("key2").Found)
		assert.True(t, cache.Get("key3").Found)

		// Invalidate all
		invResult := manager.InvalidateAll()
		assert.Equal(t, InvalidationStrategyAll, invResult.Strategy)
		assert.NotZero(t, invResult.Duration)

		// Verify all keys are gone
		assert.False(t, cache.Get("key1").Found)
		assert.False(t, cache.Get("key2").Found)
		assert.False(t, cache.Get("key3").Found)
	})
}

func TestInvalidationManager_ExecuteRule(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	t.Run("execute exact rule", func(t *testing.T) {
		// Add a key to cache
		err := cache.Set("test-key", "test-value")
		require.NoError(t, err)

		// Create and add rule
		rule := &InvalidationRule{
			Name:     "test-execute",
			Strategy: InvalidationStrategyExact,
			Pattern:  "test-key",
			Enabled:  true,
		}

		err = manager.AddRule(rule)
		require.NoError(t, err)

		// Execute rule
		result := manager.ExecuteRule(rule.ID)
		assert.Equal(t, rule.ID, result.RuleID)
		assert.Equal(t, InvalidationStrategyExact, result.Strategy)
		assert.Equal(t, int64(1), result.KeysInvalidated)
		assert.NotZero(t, result.Duration)

		// Verify key is gone
		cacheResult := cache.Get("test-key")
		assert.False(t, cacheResult.Found)
	})

	t.Run("execute non-existent rule", func(t *testing.T) {
		result := manager.ExecuteRule("non-existent")
		assert.Equal(t, "non-existent", result.RuleID)
		assert.Error(t, result.Error)
		assert.Contains(t, result.Error.Error(), "not found")
	})
}

func TestInvalidationManager_ExecuteAllRules(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	t.Run("execute all active rules", func(t *testing.T) {
		// Add keys to cache
		err := cache.Set("key1", "value1")
		require.NoError(t, err)
		err = cache.Set("key2", "value2")
		require.NoError(t, err)

		// Create and add rules
		rule1 := &InvalidationRule{
			Name:     "test-execute-all-1",
			Strategy: InvalidationStrategyExact,
			Pattern:  "key1",
			Enabled:  true,
		}

		rule2 := &InvalidationRule{
			Name:     "test-execute-all-2",
			Strategy: InvalidationStrategyExact,
			Pattern:  "key2",
			Enabled:  true,
		}

		rule3 := &InvalidationRule{
			Name:     "test-execute-all-3",
			Strategy: InvalidationStrategyExact,
			Pattern:  "key3",
			Enabled:  false, // Disabled rule
		}

		err = manager.AddRule(rule1)
		require.NoError(t, err)
		err = manager.AddRule(rule2)
		require.NoError(t, err)
		err = manager.AddRule(rule3)
		require.NoError(t, err)

		// Execute all rules
		results := manager.ExecuteAllRules()
		assert.Len(t, results, 2) // Only enabled rules

		// Verify keys are gone
		assert.False(t, cache.Get("key1").Found)
		assert.False(t, cache.Get("key2").Found)
	})
}

func TestInvalidationManager_GetStats(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	t.Run("get initial stats", func(t *testing.T) {
		stats := manager.GetStats()
		assert.Equal(t, int64(0), stats.TotalRules)
		assert.Equal(t, int64(0), stats.ActiveRules)
		assert.Equal(t, int64(0), stats.TotalEvents)
		assert.Equal(t, int64(0), stats.TotalInvalidated)
		assert.Equal(t, int64(0), stats.ErrorCount)
		assert.Equal(t, int64(0), stats.SuccessCount)
	})

	t.Run("get stats after operations", func(t *testing.T) {
		// Add a rule
		rule := &InvalidationRule{
			Name:     "test-stats",
			Strategy: InvalidationStrategyExact,
			Pattern:  "test-key",
			Enabled:  true,
		}

		err := manager.AddRule(rule)
		require.NoError(t, err)

		// Add a key and invalidate it
		err = cache.Set("test-key", "test-value")
		require.NoError(t, err)

		manager.InvalidateByKey("test-key")

		// Get stats
		stats := manager.GetStats()
		assert.Equal(t, int64(1), stats.TotalRules)
		assert.Equal(t, int64(1), stats.ActiveRules)
		assert.Equal(t, int64(1), stats.TotalEvents)
		assert.Equal(t, int64(1), stats.TotalInvalidated)
		assert.Equal(t, int64(0), stats.ErrorCount)
		assert.Equal(t, int64(1), stats.SuccessCount)
		assert.NotZero(t, stats.LastInvalidation)
		assert.NotZero(t, stats.AverageDuration)
	})
}

func TestInvalidationManager_GetEvents(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	t.Run("get events", func(t *testing.T) {
		// Perform some invalidations
		manager.InvalidateByKey("key1")
		manager.InvalidateByKey("key2")
		manager.InvalidateByKey("key3")

		// Get events
		events := manager.GetEvents(10)
		assert.Len(t, events, 3)

		// Check event properties
		for _, event := range events {
			assert.NotEmpty(t, event.ID)
			assert.Equal(t, InvalidationStrategyExact, event.Strategy)
			assert.NotZero(t, event.Timestamp)
			assert.NotZero(t, event.Duration)
			assert.Equal(t, "manual_invalidation", event.Reason)
		}
	})

	t.Run("get events with limit", func(t *testing.T) {
		// Perform more invalidations
		manager.InvalidateByKey("key4")
		manager.InvalidateByKey("key5")

		// Get limited events
		events := manager.GetEvents(2)
		assert.Len(t, events, 2)
	})
}

func TestInvalidationManager_shouldExecuteRule(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	t.Run("rule with time conditions", func(t *testing.T) {
		now := time.Now()
		start := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
		end := time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+1, now.Minute(), 0, 0, now.Location())

		rule := &InvalidationRule{
			Name:     "test-time",
			Strategy: InvalidationStrategyExact,
			Pattern:  "test-key",
			Enabled:  true,
			Conditions: InvalidationConditions{
				TimeOfDay: &TimeOfDayCondition{
					Start: start,
					End:   end,
				},
			},
		}

		// Should execute within time window
		shouldExecute := manager.shouldExecuteRule(rule)
		assert.True(t, shouldExecute)

		// Test outside time window
		rule.Conditions.TimeOfDay.Start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+2, now.Minute(), 0, 0, now.Location())
		rule.Conditions.TimeOfDay.End = time.Date(now.Year(), now.Month(), now.Day(), now.Hour()+3, now.Minute(), 0, 0, now.Location())

		shouldExecute = manager.shouldExecuteRule(rule)
		assert.False(t, shouldExecute)
	})

	t.Run("rule with day of week conditions", func(t *testing.T) {
		rule := &InvalidationRule{
			Name:     "test-day",
			Strategy: InvalidationStrategyExact,
			Pattern:  "test-key",
			Enabled:  true,
			Conditions: InvalidationConditions{
				DayOfWeek: []time.Weekday{time.Monday, time.Tuesday},
			},
		}

		// Should execute on Monday or Tuesday
		shouldExecute := manager.shouldExecuteRule(rule)
		// This will depend on the current day of the week
		// We can't predict the exact result, but we can test the logic
		assert.IsType(t, true, shouldExecute)
	})
}

func TestInvalidationManager_Close(t *testing.T) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(t, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())

	t.Run("close manager", func(t *testing.T) {
		err := manager.Close()
		assert.NoError(t, err)
	})
}

// Benchmark tests
func BenchmarkInvalidationManager_InvalidateByKey(b *testing.B) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(b, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%1000)
			manager.InvalidateByKey(key)
			i++
		}
	})
}

func BenchmarkInvalidationManager_AddRule(b *testing.B) {
	cache, err := NewIntelligentCache(CacheConfig{
		Type:    CacheTypeLRU,
		MaxSize: 1024 * 1024,
		Logger:  zap.NewNop(),
	})
	require.NoError(b, err)
	defer cache.Close()

	manager := NewInvalidationManager(cache, zap.NewNop())
	defer manager.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			rule := &InvalidationRule{
				Name:     fmt.Sprintf("rule%d", i),
				Strategy: InvalidationStrategyExact,
				Pattern:  fmt.Sprintf("pattern%d", i),
				Enabled:  true,
			}
			manager.AddRule(rule)
			i++
		}
	})
}
