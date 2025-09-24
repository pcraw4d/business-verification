package caching

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"go.uber.org/zap"
)

// InvalidationStrategy represents the type of cache invalidation
type InvalidationStrategy string

const (
	InvalidationStrategyExact      InvalidationStrategy = "exact"      // Invalidate exact key
	InvalidationStrategyPattern    InvalidationStrategy = "pattern"    // Invalidate by pattern
	InvalidationStrategyPrefix     InvalidationStrategy = "prefix"     // Invalidate by prefix
	InvalidationStrategySuffix     InvalidationStrategy = "suffix"     // Invalidate by suffix
	InvalidationStrategyTag        InvalidationStrategy = "tag"        // Invalidate by tag
	InvalidationStrategyDependency InvalidationStrategy = "dependency" // Invalidate by dependency
	InvalidationStrategyTime       InvalidationStrategy = "time"       // Invalidate by time
	InvalidationStrategySize       InvalidationStrategy = "size"       // Invalidate by size
	InvalidationStrategyPriority   InvalidationStrategy = "priority"   // Invalidate by priority
	InvalidationStrategyAll        InvalidationStrategy = "all"        // Invalidate all entries
)

// InvalidationRule represents a cache invalidation rule
type InvalidationRule struct {
	ID           string
	Name         string
	Strategy     InvalidationStrategy
	Pattern      string
	Tags         []string
	Dependencies []string
	Conditions   InvalidationConditions
	Priority     int
	Enabled      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
	mu           sync.RWMutex
}

// InvalidationConditions represents conditions for invalidation
type InvalidationConditions struct {
	MaxAge      time.Duration
	MaxSize     int64
	MaxEntries  int64
	MinPriority int
	MaxPriority int
	TimeOfDay   *TimeOfDayCondition
	DayOfWeek   []time.Weekday
	AccessCount int64
	LastAccess  time.Duration
	HitRate     float64
	MissRate    float64
}

// TimeOfDayCondition represents time-based conditions
type TimeOfDayCondition struct {
	Start time.Time
	End   time.Time
}

// InvalidationEvent represents a cache invalidation event
type InvalidationEvent struct {
	ID        string
	RuleID    string
	Strategy  InvalidationStrategy
	Keys      []string
	Count     int64
	Reason    string
	Timestamp time.Time
	Duration  time.Duration
	Error     error
}

// InvalidationManager manages cache invalidation
type InvalidationManager struct {
	cache        *IntelligentCache
	rules        map[string]*InvalidationRule
	events       []*InvalidationEvent
	patterns     map[string]*regexp.Regexp
	dependencies map[string][]string
	logger       *zap.Logger
	mu           sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
}

// InvalidationResult represents the result of an invalidation operation
type InvalidationResult struct {
	RuleID          string
	Strategy        InvalidationStrategy
	KeysInvalidated int64
	KeysMatched     []string
	Duration        time.Duration
	Error           error
}

// InvalidationStats represents invalidation statistics
type InvalidationStats struct {
	TotalRules       int64
	ActiveRules      int64
	TotalEvents      int64
	TotalInvalidated int64
	LastInvalidation time.Time
	AverageDuration  time.Duration
	ErrorCount       int64
	SuccessCount     int64
}

// NewInvalidationManager creates a new invalidation manager
func NewInvalidationManager(cache *IntelligentCache, logger *zap.Logger) *InvalidationManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &InvalidationManager{
		cache:        cache,
		rules:        make(map[string]*InvalidationRule),
		events:       make([]*InvalidationEvent, 0),
		patterns:     make(map[string]*regexp.Regexp),
		dependencies: make(map[string][]string),
		logger:       logger,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start background invalidation worker
	go manager.invalidationWorker()

	return manager
}

// AddRule adds an invalidation rule
func (im *InvalidationManager) AddRule(rule *InvalidationRule) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	if rule.ID == "" {
		rule.ID = generateInvalidationRuleID()
	}

	if rule.Name == "" {
		rule.Name = fmt.Sprintf("rule_%s", rule.ID)
	}

	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()

	// Compile pattern if needed
	if rule.Strategy == InvalidationStrategyPattern && rule.Pattern != "" {
		compiled, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern '%s': %w", rule.Pattern, err)
		}
		im.patterns[rule.ID] = compiled
	}

	// Build dependency map
	if rule.Strategy == InvalidationStrategyDependency {
		for _, dep := range rule.Dependencies {
			im.dependencies[dep] = append(im.dependencies[dep], rule.ID)
		}
	}

	im.rules[rule.ID] = rule
	im.logger.Info("Added invalidation rule",
		zap.String("rule_id", rule.ID),
		zap.String("name", rule.Name),
		zap.String("strategy", string(rule.Strategy)),
	)

	return nil
}

// RemoveRule removes an invalidation rule
func (im *InvalidationManager) RemoveRule(ruleID string) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	rule, exists := im.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	// Clean up patterns
	if rule.Strategy == InvalidationStrategyPattern {
		delete(im.patterns, ruleID)
	}

	// Clean up dependencies
	if rule.Strategy == InvalidationStrategyDependency {
		for _, dep := range rule.Dependencies {
			deps := im.dependencies[dep]
			for i, id := range deps {
				if id == ruleID {
					im.dependencies[dep] = append(deps[:i], deps[i+1:]...)
					break
				}
			}
		}
	}

	delete(im.rules, ruleID)
	im.logger.Info("Removed invalidation rule", zap.String("rule_id", ruleID))

	return nil
}

// UpdateRule updates an existing invalidation rule
func (im *InvalidationManager) UpdateRule(ruleID string, updates *InvalidationRule) error {
	im.mu.Lock()
	defer im.mu.Unlock()

	rule, exists := im.rules[ruleID]
	if !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	// Update fields
	if updates.Name != "" {
		rule.Name = updates.Name
	}
	if updates.Strategy != "" {
		rule.Strategy = updates.Strategy
	}
	if updates.Pattern != "" {
		rule.Pattern = updates.Pattern
		// Recompile pattern
		if rule.Strategy == InvalidationStrategyPattern {
			compiled, err := regexp.Compile(rule.Pattern)
			if err != nil {
				return fmt.Errorf("invalid pattern '%s': %w", rule.Pattern, err)
			}
			im.patterns[ruleID] = compiled
		}
	}
	if len(updates.Tags) > 0 {
		rule.Tags = updates.Tags
	}
	if len(updates.Dependencies) > 0 {
		rule.Dependencies = updates.Dependencies
	}
	// Check if conditions were provided (check individual fields)
	if updates.Conditions.MaxAge != 0 || updates.Conditions.MaxSize != 0 ||
		updates.Conditions.MaxEntries != 0 || updates.Conditions.MinPriority != 0 ||
		updates.Conditions.MaxPriority != 0 || updates.Conditions.TimeOfDay != nil ||
		len(updates.Conditions.DayOfWeek) > 0 || updates.Conditions.AccessCount != 0 ||
		updates.Conditions.LastAccess != 0 || updates.Conditions.HitRate != 0 ||
		updates.Conditions.MissRate != 0 {
		rule.Conditions = updates.Conditions
	}
	if updates.Priority != 0 {
		rule.Priority = updates.Priority
	}
	rule.Enabled = updates.Enabled
	rule.UpdatedAt = time.Now()

	im.logger.Info("Updated invalidation rule", zap.String("rule_id", ruleID))
	return nil
}

// GetRule retrieves an invalidation rule
func (im *InvalidationManager) GetRule(ruleID string) (*InvalidationRule, error) {
	im.mu.RLock()
	defer im.mu.RUnlock()

	rule, exists := im.rules[ruleID]
	if !exists {
		return nil, fmt.Errorf("rule %s not found", ruleID)
	}

	return rule, nil
}

// ListRules lists all invalidation rules
func (im *InvalidationManager) ListRules() []*InvalidationRule {
	im.mu.RLock()
	defer im.mu.RUnlock()

	rules := make([]*InvalidationRule, 0, len(im.rules))
	for _, rule := range im.rules {
		rules = append(rules, rule)
	}

	return rules
}

// InvalidateByKey invalidates a specific key
func (im *InvalidationManager) InvalidateByKey(key string) *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		Strategy: InvalidationStrategyExact,
	}

	// Check if key exists and delete it
	deleted := im.cache.Delete(key)
	if deleted {
		result.KeysInvalidated = 1
		result.KeysMatched = []string{key}
	}

	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// InvalidateByPattern invalidates keys matching a pattern
func (im *InvalidationManager) InvalidateByPattern(pattern string) *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		Strategy: InvalidationStrategyPattern,
	}

	compiled, err := regexp.Compile(pattern)
	if err != nil {
		result.Error = fmt.Errorf("invalid pattern '%s': %w", pattern, err)
		result.Duration = time.Since(start)
		im.recordEvent(result, err)
		return result
	}

	matchedKeys := im.findMatchingKeys(compiled)
	result.KeysInvalidated = int64(len(matchedKeys))
	result.KeysMatched = matchedKeys

	// Delete matched keys
	for _, key := range matchedKeys {
		im.cache.Delete(key)
	}

	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// InvalidateByPrefix invalidates keys with a specific prefix
func (im *InvalidationManager) InvalidateByPrefix(prefix string) *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		Strategy: InvalidationStrategyPrefix,
	}

	matchedKeys := im.findKeysByPrefix(prefix)
	result.KeysInvalidated = int64(len(matchedKeys))
	result.KeysMatched = matchedKeys

	// Delete matched keys
	for _, key := range matchedKeys {
		im.cache.Delete(key)
	}

	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// InvalidateBySuffix invalidates keys with a specific suffix
func (im *InvalidationManager) InvalidateBySuffix(suffix string) *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		Strategy: InvalidationStrategySuffix,
	}

	matchedKeys := im.findKeysBySuffix(suffix)
	result.KeysInvalidated = int64(len(matchedKeys))
	result.KeysMatched = matchedKeys

	// Delete matched keys
	for _, key := range matchedKeys {
		im.cache.Delete(key)
	}

	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// InvalidateByTag invalidates entries with specific tags
func (im *InvalidationManager) InvalidateByTag(tags ...string) *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		Strategy: InvalidationStrategyTag,
	}

	matchedKeys := im.findKeysByTags(tags)
	result.KeysInvalidated = int64(len(matchedKeys))
	result.KeysMatched = matchedKeys

	// Delete matched keys
	for _, key := range matchedKeys {
		im.cache.Delete(key)
	}

	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// InvalidateByDependency invalidates entries based on dependencies
func (im *InvalidationManager) InvalidateByDependency(dependency string) *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		Strategy: InvalidationStrategyDependency,
	}

	// Find dependent rules
	im.mu.RLock()
	dependentRules := im.dependencies[dependency]
	im.mu.RUnlock()

	// Execute dependent rules
	var totalInvalidated int64
	var allMatchedKeys []string

	for _, ruleID := range dependentRules {
		rule, err := im.GetRule(ruleID)
		if err != nil {
			continue
		}

		ruleResult := im.executeRule(rule)
		totalInvalidated += ruleResult.KeysInvalidated
		allMatchedKeys = append(allMatchedKeys, ruleResult.KeysMatched...)
	}

	result.KeysInvalidated = totalInvalidated
	result.KeysMatched = allMatchedKeys
	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// InvalidateByTime invalidates entries based on time conditions
func (im *InvalidationManager) InvalidateByTime(conditions InvalidationConditions) *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		Strategy: InvalidationStrategyTime,
	}

	matchedKeys := im.findKeysByTimeConditions(conditions)
	result.KeysInvalidated = int64(len(matchedKeys))
	result.KeysMatched = matchedKeys

	// Delete matched keys
	for _, key := range matchedKeys {
		im.cache.Delete(key)
	}

	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// InvalidateBySize invalidates entries based on size conditions
func (im *InvalidationManager) InvalidateBySize(conditions InvalidationConditions) *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		Strategy: InvalidationStrategySize,
	}

	matchedKeys := im.findKeysBySizeConditions(conditions)
	result.KeysInvalidated = int64(len(matchedKeys))
	result.KeysMatched = matchedKeys

	// Delete matched keys
	for _, key := range matchedKeys {
		im.cache.Delete(key)
	}

	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// InvalidateByPriority invalidates entries based on priority
func (im *InvalidationManager) InvalidateByPriority(conditions InvalidationConditions) *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		Strategy: InvalidationStrategyPriority,
	}

	matchedKeys := im.findKeysByPriorityConditions(conditions)
	result.KeysInvalidated = int64(len(matchedKeys))
	result.KeysMatched = matchedKeys

	// Delete matched keys
	for _, key := range matchedKeys {
		im.cache.Delete(key)
	}

	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// InvalidateAll invalidates all cache entries
func (im *InvalidationManager) InvalidateAll() *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		Strategy: InvalidationStrategyAll,
	}

	// Get all keys before clearing
	allKeys := im.getAllKeys()
	result.KeysInvalidated = int64(len(allKeys))
	result.KeysMatched = allKeys

	// Clear the entire cache
	im.cache.Clear()

	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// ExecuteRule executes a specific invalidation rule
func (im *InvalidationManager) ExecuteRule(ruleID string) *InvalidationResult {
	rule, err := im.GetRule(ruleID)
	if err != nil {
		return &InvalidationResult{
			RuleID: ruleID,
			Error:  err,
		}
	}

	return im.executeRule(rule)
}

// ExecuteAllRules executes all active invalidation rules
func (im *InvalidationManager) ExecuteAllRules() []*InvalidationResult {
	im.mu.RLock()
	rules := make([]*InvalidationRule, 0, len(im.rules))
	for _, rule := range im.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	im.mu.RUnlock()

	results := make([]*InvalidationResult, 0, len(rules))
	for _, rule := range rules {
		result := im.executeRule(rule)
		results = append(results, result)
	}

	return results
}

// GetStats returns invalidation statistics
func (im *InvalidationManager) GetStats() *InvalidationStats {
	im.mu.RLock()
	defer im.mu.RUnlock()

	stats := &InvalidationStats{
		TotalRules:  int64(len(im.rules)),
		TotalEvents: int64(len(im.events)),
	}

	// Count active rules
	for _, rule := range im.rules {
		if rule.Enabled {
			stats.ActiveRules++
		}
	}

	// Calculate event statistics
	if len(im.events) > 0 {
		stats.LastInvalidation = im.events[len(im.events)-1].Timestamp

		var totalDuration time.Duration
		for _, event := range im.events {
			stats.TotalInvalidated += event.Count
			totalDuration += event.Duration

			if event.Error != nil {
				stats.ErrorCount++
			} else {
				stats.SuccessCount++
			}
		}

		if len(im.events) > 0 {
			stats.AverageDuration = totalDuration / time.Duration(len(im.events))
		}
	}

	return stats
}

// GetEvents returns invalidation events
func (im *InvalidationManager) GetEvents(limit int) []*InvalidationEvent {
	im.mu.RLock()
	defer im.mu.RUnlock()

	if limit <= 0 || limit > len(im.events) {
		limit = len(im.events)
	}

	events := make([]*InvalidationEvent, limit)
	copy(events, im.events[len(im.events)-limit:])

	return events
}

// Close closes the invalidation manager
func (im *InvalidationManager) Close() error {
	im.cancel()
	return nil
}

// executeRule executes a specific invalidation rule
func (im *InvalidationManager) executeRule(rule *InvalidationRule) *InvalidationResult {
	start := time.Now()

	result := &InvalidationResult{
		RuleID:   rule.ID,
		Strategy: rule.Strategy,
	}

	switch rule.Strategy {
	case InvalidationStrategyExact:
		if rule.Pattern != "" {
			deleted := im.cache.Delete(rule.Pattern)
			if deleted {
				result.KeysInvalidated = 1
				result.KeysMatched = []string{rule.Pattern}
			}
		}

	case InvalidationStrategyPattern:
		if rule.Pattern != "" {
			compiled := im.patterns[rule.ID]
			if compiled != nil {
				matchedKeys := im.findMatchingKeys(compiled)
				result.KeysInvalidated = int64(len(matchedKeys))
				result.KeysMatched = matchedKeys
				for _, key := range matchedKeys {
					im.cache.Delete(key)
				}
			}
		}

	case InvalidationStrategyPrefix:
		if rule.Pattern != "" {
			matchedKeys := im.findKeysByPrefix(rule.Pattern)
			result.KeysInvalidated = int64(len(matchedKeys))
			result.KeysMatched = matchedKeys
			for _, key := range matchedKeys {
				im.cache.Delete(key)
			}
		}

	case InvalidationStrategySuffix:
		if rule.Pattern != "" {
			matchedKeys := im.findKeysBySuffix(rule.Pattern)
			result.KeysInvalidated = int64(len(matchedKeys))
			result.KeysMatched = matchedKeys
			for _, key := range matchedKeys {
				im.cache.Delete(key)
			}
		}

	case InvalidationStrategyTag:
		if len(rule.Tags) > 0 {
			matchedKeys := im.findKeysByTags(rule.Tags)
			result.KeysInvalidated = int64(len(matchedKeys))
			result.KeysMatched = matchedKeys
			for _, key := range matchedKeys {
				im.cache.Delete(key)
			}
		}

	case InvalidationStrategyTime:
		matchedKeys := im.findKeysByTimeConditions(rule.Conditions)
		result.KeysInvalidated = int64(len(matchedKeys))
		result.KeysMatched = matchedKeys
		for _, key := range matchedKeys {
			im.cache.Delete(key)
		}

	case InvalidationStrategySize:
		matchedKeys := im.findKeysBySizeConditions(rule.Conditions)
		result.KeysInvalidated = int64(len(matchedKeys))
		result.KeysMatched = matchedKeys
		for _, key := range matchedKeys {
			im.cache.Delete(key)
		}

	case InvalidationStrategyPriority:
		matchedKeys := im.findKeysByPriorityConditions(rule.Conditions)
		result.KeysInvalidated = int64(len(matchedKeys))
		result.KeysMatched = matchedKeys
		for _, key := range matchedKeys {
			im.cache.Delete(key)
		}

	case InvalidationStrategyAll:
		allKeys := im.getAllKeys()
		result.KeysInvalidated = int64(len(allKeys))
		result.KeysMatched = allKeys
		im.cache.Clear()
	}

	result.Duration = time.Since(start)
	im.recordEvent(result, nil)

	return result
}

// findMatchingKeys finds keys matching a regex pattern
func (im *InvalidationManager) findMatchingKeys(pattern *regexp.Regexp) []string {
	var matchedKeys []string

	// This would need access to all keys in the cache
	// For now, we'll return an empty slice as this requires cache internals
	// In a real implementation, you'd iterate through all cache keys

	return matchedKeys
}

// findKeysByPrefix finds keys with a specific prefix
func (im *InvalidationManager) findKeysByPrefix(prefix string) []string {
	var matchedKeys []string

	// This would need access to all keys in the cache
	// For now, we'll return an empty slice as this requires cache internals

	return matchedKeys
}

// findKeysBySuffix finds keys with a specific suffix
func (im *InvalidationManager) findKeysBySuffix(suffix string) []string {
	var matchedKeys []string

	// This would need access to all keys in the cache
	// For now, we'll return an empty slice as this requires cache internals

	return matchedKeys
}

// findKeysByTags finds keys with specific tags
func (im *InvalidationManager) findKeysByTags(tags []string) []string {
	var matchedKeys []string

	// This would need access to all keys in the cache
	// For now, we'll return an empty slice as this requires cache internals

	return matchedKeys
}

// findKeysByTimeConditions finds keys matching time conditions
func (im *InvalidationManager) findKeysByTimeConditions(conditions InvalidationConditions) []string {
	var matchedKeys []string

	// This would need access to all keys in the cache
	// For now, we'll return an empty slice as this requires cache internals

	return matchedKeys
}

// findKeysBySizeConditions finds keys matching size conditions
func (im *InvalidationManager) findKeysBySizeConditions(conditions InvalidationConditions) []string {
	var matchedKeys []string

	// This would need access to all keys in the cache
	// For now, we'll return an empty slice as this requires cache internals

	return matchedKeys
}

// findKeysByPriorityConditions finds keys matching priority conditions
func (im *InvalidationManager) findKeysByPriorityConditions(conditions InvalidationConditions) []string {
	var matchedKeys []string

	// This would need access to all keys in the cache
	// For now, we'll return an empty slice as this requires cache internals

	return matchedKeys
}

// getAllKeys gets all keys in the cache
func (im *InvalidationManager) getAllKeys() []string {
	var allKeys []string

	// This would need access to all keys in the cache
	// For now, we'll return an empty slice as this requires cache internals

	return allKeys
}

// recordEvent records an invalidation event
func (im *InvalidationManager) recordEvent(result *InvalidationResult, err error) {
	im.mu.Lock()
	defer im.mu.Unlock()

	event := &InvalidationEvent{
		ID:        generateInvalidationEventID(),
		RuleID:    result.RuleID,
		Strategy:  result.Strategy,
		Keys:      result.KeysMatched,
		Count:     result.KeysInvalidated,
		Reason:    "manual_invalidation",
		Timestamp: time.Now(),
		Duration:  result.Duration,
		Error:     err,
	}

	im.events = append(im.events, event)

	// Keep only last 1000 events
	if len(im.events) > 1000 {
		im.events = im.events[len(im.events)-1000:]
	}

	im.logger.Info("Cache invalidation event",
		zap.String("event_id", event.ID),
		zap.String("rule_id", event.RuleID),
		zap.String("strategy", string(event.Strategy)),
		zap.Int64("keys_invalidated", event.Count),
		zap.Duration("duration", event.Duration),
		zap.Error(event.Error),
	)
}

// invalidationWorker runs background invalidation tasks
func (im *InvalidationManager) invalidationWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-im.ctx.Done():
			return
		case <-ticker.C:
			im.executeScheduledInvalidations()
		}
	}
}

// executeScheduledInvalidations executes scheduled invalidation rules
func (im *InvalidationManager) executeScheduledInvalidations() {
	im.mu.RLock()
	rules := make([]*InvalidationRule, 0, len(im.rules))
	for _, rule := range im.rules {
		if rule.Enabled && im.shouldExecuteRule(rule) {
			rules = append(rules, rule)
		}
	}
	im.mu.RUnlock()

	for _, rule := range rules {
		im.executeRule(rule)
	}
}

// shouldExecuteRule determines if a rule should be executed
func (im *InvalidationManager) shouldExecuteRule(rule *InvalidationRule) bool {
	// Check time conditions
	if rule.Conditions.TimeOfDay != nil {
		now := time.Now()
		start := rule.Conditions.TimeOfDay.Start
		end := rule.Conditions.TimeOfDay.End

		currentTime := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
		startTime := time.Date(now.Year(), now.Month(), now.Day(), start.Hour(), start.Minute(), 0, 0, now.Location())
		endTime := time.Date(now.Year(), now.Month(), now.Day(), end.Hour(), end.Minute(), 0, 0, now.Location())

		if currentTime.Before(startTime) || currentTime.After(endTime) {
			return false
		}
	}

	// Check day of week conditions
	if len(rule.Conditions.DayOfWeek) > 0 {
		now := time.Now()
		found := false
		for _, day := range rule.Conditions.DayOfWeek {
			if now.Weekday() == day {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// generateInvalidationRuleID generates a unique rule ID
func generateInvalidationRuleID() string {
	return fmt.Sprintf("rule_%d", time.Now().UnixNano())
}

// generateInvalidationEventID generates a unique event ID
func generateInvalidationEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}
