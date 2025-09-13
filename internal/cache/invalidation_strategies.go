package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// InvalidationStrategy defines different cache invalidation strategies
type InvalidationStrategy interface {
	Invalidate(ctx context.Context, cache Cache, key string) error
	GetName() string
	GetDescription() string
}

// TimeBasedInvalidationStrategy invalidates cache entries based on time
type TimeBasedInvalidationStrategy struct {
	TTL    time.Duration
	logger *zap.Logger
}

// NewTimeBasedInvalidationStrategy creates a new time-based invalidation strategy
func NewTimeBasedInvalidationStrategy(ttl time.Duration, logger *zap.Logger) *TimeBasedInvalidationStrategy {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &TimeBasedInvalidationStrategy{
		TTL:    ttl,
		logger: logger,
	}
}

// Invalidate implements the InvalidationStrategy interface
func (tbis *TimeBasedInvalidationStrategy) Invalidate(ctx context.Context, cache Cache, key string) error {
	// Set TTL for the key
	err := cache.SetTTL(ctx, key, tbis.TTL)
	if err != nil {
		tbis.logger.Error("Failed to set TTL for cache key",
			zap.String("key", key),
			zap.Duration("ttl", tbis.TTL),
			zap.Error(err))
		return fmt.Errorf("failed to set TTL: %w", err)
	}

	tbis.logger.Debug("Set TTL for cache key",
		zap.String("key", key),
		zap.Duration("ttl", tbis.TTL))

	return nil
}

// GetName returns the strategy name
func (tbis *TimeBasedInvalidationStrategy) GetName() string {
	return "time_based"
}

// GetDescription returns the strategy description
func (tbis *TimeBasedInvalidationStrategy) GetDescription() string {
	return fmt.Sprintf("Invalidates cache entries after %v", tbis.TTL)
}

// EventBasedInvalidationStrategy invalidates cache entries based on events
type EventBasedInvalidationStrategy struct {
	eventHandlers map[string][]func(context.Context, Cache, string) error
	mu            sync.RWMutex
	logger        *zap.Logger
}

// NewEventBasedInvalidationStrategy creates a new event-based invalidation strategy
func NewEventBasedInvalidationStrategy(logger *zap.Logger) *EventBasedInvalidationStrategy {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &EventBasedInvalidationStrategy{
		eventHandlers: make(map[string][]func(context.Context, Cache, string) error),
		logger:        logger,
	}
}

// RegisterEventHandler registers an event handler for a specific event type
func (ebis *EventBasedInvalidationStrategy) RegisterEventHandler(eventType string, handler func(context.Context, Cache, string) error) {
	ebis.mu.Lock()
	defer ebis.mu.Unlock()

	ebis.eventHandlers[eventType] = append(ebis.eventHandlers[eventType], handler)

	ebis.logger.Debug("Registered event handler",
		zap.String("event_type", eventType),
		zap.Int("handler_count", len(ebis.eventHandlers[eventType])))
}

// TriggerEvent triggers an invalidation event
func (ebis *EventBasedInvalidationStrategy) TriggerEvent(ctx context.Context, cache Cache, eventType string, key string) error {
	ebis.mu.RLock()
	handlers := ebis.eventHandlers[eventType]
	ebis.mu.RUnlock()

	if len(handlers) == 0 {
		ebis.logger.Debug("No handlers registered for event type",
			zap.String("event_type", eventType))
		return nil
	}

	for i, handler := range handlers {
		err := handler(ctx, cache, key)
		if err != nil {
			ebis.logger.Error("Event handler failed",
				zap.String("event_type", eventType),
				zap.Int("handler_index", i),
				zap.String("key", key),
				zap.Error(err))
			return fmt.Errorf("event handler %d failed: %w", i, err)
		}
	}

	ebis.logger.Debug("Triggered event handlers",
		zap.String("event_type", eventType),
		zap.String("key", key),
		zap.Int("handler_count", len(handlers)))

	return nil
}

// Invalidate implements the InvalidationStrategy interface
func (ebis *EventBasedInvalidationStrategy) Invalidate(ctx context.Context, cache Cache, key string) error {
	// This strategy doesn't directly invalidate, it triggers events
	// The actual invalidation is handled by registered event handlers
	return ebis.TriggerEvent(ctx, cache, "invalidate", key)
}

// GetName returns the strategy name
func (ebis *EventBasedInvalidationStrategy) GetName() string {
	return "event_based"
}

// GetDescription returns the strategy description
func (ebis *EventBasedInvalidationStrategy) GetDescription() string {
	return "Invalidates cache entries based on events"
}

// PatternBasedInvalidationStrategy invalidates cache entries based on patterns
type PatternBasedInvalidationStrategy struct {
	patterns map[string]time.Duration
	mu       sync.RWMutex
	logger   *zap.Logger
}

// NewPatternBasedInvalidationStrategy creates a new pattern-based invalidation strategy
func NewPatternBasedInvalidationStrategy(logger *zap.Logger) *PatternBasedInvalidationStrategy {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &PatternBasedInvalidationStrategy{
		patterns: make(map[string]time.Duration),
		logger:   logger,
	}
}

// AddPattern adds a pattern with its associated TTL
func (pbis *PatternBasedInvalidationStrategy) AddPattern(pattern string, ttl time.Duration) {
	pbis.mu.Lock()
	defer pbis.mu.Unlock()

	pbis.patterns[pattern] = ttl

	pbis.logger.Debug("Added invalidation pattern",
		zap.String("pattern", pattern),
		zap.Duration("ttl", ttl))
}

// RemovePattern removes a pattern
func (pbis *PatternBasedInvalidationStrategy) RemovePattern(pattern string) {
	pbis.mu.Lock()
	defer pbis.mu.Unlock()

	delete(pbis.patterns, pattern)

	pbis.logger.Debug("Removed invalidation pattern",
		zap.String("pattern", pattern))
}

// Invalidate implements the InvalidationStrategy interface
func (pbis *PatternBasedInvalidationStrategy) Invalidate(ctx context.Context, cache Cache, key string) error {
	pbis.mu.RLock()
	patterns := make(map[string]time.Duration)
	for pattern, ttl := range pbis.patterns {
		patterns[pattern] = ttl
	}
	pbis.mu.RUnlock()

	// Check if key matches any pattern
	for pattern, ttl := range patterns {
		if pbis.matchesPattern(key, pattern) {
			err := cache.SetTTL(ctx, key, ttl)
			if err != nil {
				pbis.logger.Error("Failed to set TTL for pattern-matched key",
					zap.String("key", key),
					zap.String("pattern", pattern),
					zap.Duration("ttl", ttl),
					zap.Error(err))
				return fmt.Errorf("failed to set TTL for pattern %s: %w", pattern, err)
			}

			pbis.logger.Debug("Applied pattern-based TTL",
				zap.String("key", key),
				zap.String("pattern", pattern),
				zap.Duration("ttl", ttl))
		}
	}

	return nil
}

// matchesPattern checks if a key matches a pattern (simple string matching for now)
func (pbis *PatternBasedInvalidationStrategy) matchesPattern(key, pattern string) bool {
	// Simple pattern matching - can be enhanced with regex
	if pattern == "*" {
		return true
	}

	if pattern == key {
		return true
	}

	// Check if pattern is a prefix
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}

	return false
}

// GetName returns the strategy name
func (pbis *PatternBasedInvalidationStrategy) GetName() string {
	return "pattern_based"
}

// GetDescription returns the strategy description
func (pbis *PatternBasedInvalidationStrategy) GetDescription() string {
	return "Invalidates cache entries based on key patterns"
}

// LRUBasedInvalidationStrategy invalidates cache entries based on LRU (Least Recently Used)
type LRUBasedInvalidationStrategy struct {
	maxSize int
	logger  *zap.Logger
}

// NewLRUBasedInvalidationStrategy creates a new LRU-based invalidation strategy
func NewLRUBasedInvalidationStrategy(maxSize int, logger *zap.Logger) *LRUBasedInvalidationStrategy {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &LRUBasedInvalidationStrategy{
		maxSize: maxSize,
		logger:  logger,
	}
}

// Invalidate implements the InvalidationStrategy interface
func (lruis *LRUBasedInvalidationStrategy) Invalidate(ctx context.Context, cache Cache, key string) error {
	// Get current cache size
	currentSize := cache.GetSize()

	if currentSize >= int64(lruis.maxSize) {
		// Cache is at capacity, need to evict LRU entries
		// This is a simplified implementation - in practice, you'd need
		// access to the cache's internal LRU tracking
		lruis.logger.Debug("Cache at capacity, LRU eviction needed",
			zap.Int64("current_size", currentSize),
			zap.Int("max_size", lruis.maxSize))

		// For now, we'll just log this - actual LRU eviction would need
		// to be implemented in the cache itself
		return nil
	}

	return nil
}

// GetName returns the strategy name
func (lruis *LRUBasedInvalidationStrategy) GetName() string {
	return "lru_based"
}

// GetDescription returns the strategy description
func (lruis *LRUBasedInvalidationStrategy) GetDescription() string {
	return fmt.Sprintf("Invalidates cache entries based on LRU when cache size exceeds %d", lruis.maxSize)
}

// CompositeInvalidationStrategy combines multiple invalidation strategies
type CompositeInvalidationStrategy struct {
	strategies []InvalidationStrategy
	logger     *zap.Logger
}

// NewCompositeInvalidationStrategy creates a new composite invalidation strategy
func NewCompositeInvalidationStrategy(strategies []InvalidationStrategy, logger *zap.Logger) *CompositeInvalidationStrategy {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &CompositeInvalidationStrategy{
		strategies: strategies,
		logger:     logger,
	}
}

// AddStrategy adds a strategy to the composite
func (cis *CompositeInvalidationStrategy) AddStrategy(strategy InvalidationStrategy) {
	cis.strategies = append(cis.strategies, strategy)

	cis.logger.Debug("Added strategy to composite",
		zap.String("strategy", strategy.GetName()),
		zap.Int("total_strategies", len(cis.strategies)))
}

// RemoveStrategy removes a strategy from the composite
func (cis *CompositeInvalidationStrategy) RemoveStrategy(strategyName string) {
	for i, strategy := range cis.strategies {
		if strategy.GetName() == strategyName {
			cis.strategies = append(cis.strategies[:i], cis.strategies[i+1:]...)
			cis.logger.Debug("Removed strategy from composite",
				zap.String("strategy", strategyName),
				zap.Int("remaining_strategies", len(cis.strategies)))
			break
		}
	}
}

// Invalidate implements the InvalidationStrategy interface
func (cis *CompositeInvalidationStrategy) Invalidate(ctx context.Context, cache Cache, key string) error {
	var lastErr error

	for _, strategy := range cis.strategies {
		err := strategy.Invalidate(ctx, cache, key)
		if err != nil {
			cis.logger.Error("Strategy failed",
				zap.String("strategy", strategy.GetName()),
				zap.String("key", key),
				zap.Error(err))
			lastErr = err
		}
	}

	return lastErr
}

// GetName returns the strategy name
func (cis *CompositeInvalidationStrategy) GetName() string {
	return "composite"
}

// GetDescription returns the strategy description
func (cis *CompositeInvalidationStrategy) GetDescription() string {
	return fmt.Sprintf("Combines %d invalidation strategies", len(cis.strategies))
}

// GetStrategies returns all strategies in the composite
func (cis *CompositeInvalidationStrategy) GetStrategies() []InvalidationStrategy {
	return cis.strategies
}
