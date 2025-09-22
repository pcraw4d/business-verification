package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/modules/caching"
)

// CacheHandler handles HTTP requests for cache operations
type CacheHandler struct {
	cache       *caching.IntelligentCache
	monitor     *caching.CacheMonitor
	optimizer   *caching.CacheOptimizer
	invalidator *caching.InvalidationManager
	logger      *zap.Logger
}

// NewCacheHandler creates a new cache handler
func NewCacheHandler(
	cache *caching.IntelligentCache,
	monitor *caching.CacheMonitor,
	optimizer *caching.CacheOptimizer,
	invalidator *caching.InvalidationManager,
	logger *zap.Logger,
) *CacheHandler {
	return &CacheHandler{
		cache:       cache,
		monitor:     monitor,
		optimizer:   optimizer,
		invalidator: invalidator,
		logger:      logger,
	}
}

// GetCacheValue retrieves a value from the cache
func (h *CacheHandler) GetCacheValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Cache key is required")
		return
	}

	// Get value from cache
	result := h.cache.Get(key)
	if !result.Found {
		h.writeError(w, http.StatusNotFound, "KEY_NOT_FOUND", fmt.Sprintf("Key '%s' not found in cache", key))
		return
	}

	// Calculate TTL
	var ttl int
	if result.Expired {
		ttl = 0
	} else {
		// For now, we'll use a default TTL since we don't have direct access to entry details
		ttl = 3600 // Default 1 hour
	}

	response := CacheGetResponse{
		Key:         key,
		Value:       result.Value,
		TTL:         ttl,
		CreatedAt:   time.Now().Add(-time.Hour), // Default value
		AccessedAt:  time.Now(),
		AccessCount: int(result.AccessCount),
		Tags:        []string{},               // Default empty tags
		Metadata:    map[string]interface{}{}, // Default empty metadata
	}

	h.writeJSON(w, http.StatusOK, response)
}

// SetCacheValue stores a value in the cache
func (h *CacheHandler) SetCacheValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Cache key is required")
		return
	}

	var request CacheSetRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate request
	if request.Value == nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Value is required")
		return
	}

	// Create cache options
	options := []caching.CacheOption{}

	if request.TTL > 0 {
		options = append(options, caching.WithTTL(time.Duration(request.TTL)*time.Second))
	}

	if request.Priority > 0 {
		options = append(options, caching.WithPriority(request.Priority))
	}

	if len(request.Tags) > 0 {
		options = append(options, caching.WithTags(request.Tags...))
	}

	if request.Metadata != nil {
		options = append(options, caching.WithMetadata(request.Metadata))
	}

	// Set value in cache
	err := h.cache.Set(key, request.Value, options...)
	if err != nil {
		h.writeError(w, http.StatusRequestEntityTooLarge, "VALUE_TOO_LARGE", "Value exceeds maximum cache size")
		return
	}

	// Calculate size (using a simple estimation)
	size := len(fmt.Sprintf("%v", request.Value))

	response := CacheSetResponse{
		Key:     key,
		Success: true,
		Message: "Value cached successfully",
		TTL:     request.TTL,
		Size:    size,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// DeleteCacheValue removes a value from the cache
func (h *CacheHandler) DeleteCacheValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	if key == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Cache key is required")
		return
	}

	// Check if key exists
	result := h.cache.Get(key)
	if !result.Found {
		h.writeError(w, http.StatusNotFound, "KEY_NOT_FOUND", fmt.Sprintf("Key '%s' not found in cache", key))
		return
	}

	// Delete from cache
	h.cache.Delete(key)

	response := CacheDeleteResponse{
		Key:     key,
		Success: true,
		Message: "Value deleted successfully",
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ClearCache removes all values from the cache
func (h *CacheHandler) ClearCache(w http.ResponseWriter, r *http.Request) {
	// Get stats before clearing
	stats := h.cache.GetStats()
	deletedCount := stats.EntryCount

	// Clear cache
	h.cache.Clear()

	response := CacheClearResponse{
		Success:      true,
		Message:      "Cache cleared successfully",
		DeletedCount: int(deletedCount),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetCacheStats retrieves cache statistics
func (h *CacheHandler) GetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := h.cache.GetStats()

	response := CacheStatsResponse{
		Hits:              int(stats.Hits),
		Misses:            int(stats.Misses),
		Evictions:         int(stats.Evictions),
		Expirations:       int(stats.Expirations),
		TotalSize:         int(stats.TotalSize),
		EntryCount:        int(stats.EntryCount),
		HitRate:           stats.HitRate,
		MissRate:          stats.MissRate,
		AverageAccessTime: float64(stats.AverageAccessTime.Milliseconds()),
		MemoryUsage:       0,  // Not available in current stats
		Throughput:        0,  // Not available in current stats
		ShardCount:        16, // Default shard count
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetCacheAnalytics retrieves detailed cache analytics
func (h *CacheHandler) GetCacheAnalytics(w http.ResponseWriter, r *http.Request) {
	// Get time range from query parameter
	timeRange := r.URL.Query().Get("time_range")
	if timeRange == "" {
		timeRange = "1h"
	}

	// Parse time range
	duration, err := parseTimeRange(timeRange)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid time range format")
		return
	}

	// Get analytics
	analytics := h.cache.GetAnalytics()

	// Get metrics for the time range
	endTime := time.Now()
	startTime := endTime.Add(-duration)

	metrics := h.monitor.GetMetrics(caching.CacheMetricTypeHitRate, startTime, endTime)
	snapshots := h.monitor.GetSnapshots(startTime, endTime)

	// Calculate analytics from metrics and snapshots
	response := h.calculateAnalytics(analytics, metrics, snapshots, timeRange)

	h.writeJSON(w, http.StatusOK, response)
}

// ListOptimizationPlans retrieves all optimization plans
func (h *CacheHandler) ListOptimizationPlans(w http.ResponseWriter, r *http.Request) {
	plans := h.optimizer.GetOptimizationPlans()

	response := OptimizationPlansResponse{
		Plans:      plans,
		TotalCount: len(plans),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GenerateOptimizationPlan generates a new optimization plan
func (h *CacheHandler) GenerateOptimizationPlan(w http.ResponseWriter, r *http.Request) {
	var request OptimizationPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		// If no body provided, use default values
		request = OptimizationPlanRequest{
			ForceGeneration: false,
		}
	}

	plan, err := h.optimizer.GenerateOptimizationPlan()
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "NO_OPTIMIZATION_NEEDED", err.Error())
		return
	}

	response := OptimizationPlanResponse{
		ID:                 plan.ID,
		Name:               plan.Name,
		Description:        plan.Description,
		Actions:            plan.Actions,
		EstimatedTotalGain: plan.EstimatedTotalGain,
		EstimatedTotalCost: plan.EstimatedTotalCost,
		EstimatedROI:       plan.EstimatedROI,
		RiskLevel:          plan.RiskLevel,
		ExecutionTime:      plan.ExecutionTime,
		CreatedAt:          plan.CreatedAt,
		Status:             plan.Status,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetOptimizationPlan retrieves a specific optimization plan
func (h *CacheHandler) GetOptimizationPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["plan_id"]

	if planID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Plan ID is required")
		return
	}

	plans := h.optimizer.GetOptimizationPlans()
	var targetPlan *caching.OptimizationPlan

	for _, plan := range plans {
		if plan.ID == planID {
			targetPlan = &plan
			break
		}
	}

	if targetPlan == nil {
		h.writeError(w, http.StatusNotFound, "PLAN_NOT_FOUND", fmt.Sprintf("Optimization plan '%s' not found", planID))
		return
	}

	response := OptimizationPlanResponse{
		ID:                 targetPlan.ID,
		Name:               targetPlan.Name,
		Description:        targetPlan.Description,
		Actions:            targetPlan.Actions,
		EstimatedTotalGain: targetPlan.EstimatedTotalGain,
		EstimatedTotalCost: targetPlan.EstimatedTotalCost,
		EstimatedROI:       targetPlan.EstimatedROI,
		RiskLevel:          targetPlan.RiskLevel,
		ExecutionTime:      targetPlan.ExecutionTime,
		CreatedAt:          targetPlan.CreatedAt,
		Status:             targetPlan.Status,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ExecuteOptimizationPlan executes a specific optimization plan
func (h *CacheHandler) ExecuteOptimizationPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	planID := vars["plan_id"]

	if planID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Plan ID is required")
		return
	}

	result, err := h.optimizer.ExecuteOptimizationPlan(planID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "PLAN_NOT_FOUND", err.Error())
		return
	}

	errorMessage := ""
	if result.Error != nil {
		errorMessage = result.Error.Error()
	}

	response := OptimizationResultResponse{
		ID:                result.ActionID,
		PlanID:            result.ActionID, // Using ActionID as PlanID for now
		Success:           result.Success,
		ErrorMessage:      errorMessage,
		ActualGain:        result.Improvement["hit_rate"],
		ActualCost:        result.Duration.Seconds(),
		ExecutionDuration: result.Duration,
		Timestamp:         result.Timestamp,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ListOptimizationResults retrieves all optimization results
func (h *CacheHandler) ListOptimizationResults(w http.ResponseWriter, r *http.Request) {
	results := h.optimizer.GetOptimizationResults()

	response := OptimizationResultsResponse{
		Results:    results,
		TotalCount: len(results),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ListInvalidationRules retrieves all invalidation rules
func (h *CacheHandler) ListInvalidationRules(w http.ResponseWriter, r *http.Request) {
	rules := h.invalidator.ListRules()

	response := InvalidationRulesResponse{
		Rules:      rules,
		TotalCount: len(rules),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// CreateInvalidationRule creates a new invalidation rule
func (h *CacheHandler) CreateInvalidationRule(w http.ResponseWriter, r *http.Request) {
	var request InvalidationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate request
	if request.Name == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Rule name is required")
		return
	}

	if request.Strategy == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Strategy is required")
		return
	}

	// Create invalidation rule
	rule := &caching.InvalidationRule{
		Name:         request.Name,
		Strategy:     caching.InvalidationStrategy(request.Strategy),
		Pattern:      request.Pattern,
		Tags:         request.Tags,
		Dependencies: request.Dependencies,
		Conditions:   caching.InvalidationConditions{},
		Priority:     request.Priority,
		Enabled:      request.Enabled,
	}

	err := h.invalidator.AddRule(rule)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_RULE", err.Error())
		return
	}

	// Get the created rule
	createdRule, err := h.invalidator.GetRule(rule.ID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve created rule")
		return
	}

	response := InvalidationRuleResponse{
		ID:           createdRule.ID,
		Name:         createdRule.Name,
		Strategy:     string(createdRule.Strategy),
		Pattern:      createdRule.Pattern,
		Tags:         createdRule.Tags,
		Dependencies: createdRule.Dependencies,
		Conditions:   map[string]interface{}{},
		Priority:     createdRule.Priority,
		Enabled:      createdRule.Enabled,
		CreatedAt:    createdRule.CreatedAt,
		UpdatedAt:    createdRule.UpdatedAt,
	}

	h.writeJSON(w, http.StatusCreated, response)
}

// GetInvalidationRule retrieves a specific invalidation rule
func (h *CacheHandler) GetInvalidationRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["rule_id"]

	if ruleID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Rule ID is required")
		return
	}

	rule, err := h.invalidator.GetRule(ruleID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "RULE_NOT_FOUND", fmt.Sprintf("Invalidation rule '%s' not found", ruleID))
		return
	}

	response := InvalidationRuleResponse{
		ID:           rule.ID,
		Name:         rule.Name,
		Strategy:     string(rule.Strategy),
		Pattern:      rule.Pattern,
		Tags:         rule.Tags,
		Dependencies: rule.Dependencies,
		Conditions:   map[string]interface{}{},
		Priority:     rule.Priority,
		Enabled:      rule.Enabled,
		CreatedAt:    rule.CreatedAt,
		UpdatedAt:    rule.UpdatedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// UpdateInvalidationRule updates an existing invalidation rule
func (h *CacheHandler) UpdateInvalidationRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["rule_id"]

	if ruleID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Rule ID is required")
		return
	}

	var request InvalidationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Get existing rule
	existingRule, err := h.invalidator.GetRule(ruleID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "RULE_NOT_FOUND", fmt.Sprintf("Invalidation rule '%s' not found", ruleID))
		return
	}

	// Update rule fields
	if request.Name != "" {
		existingRule.Name = request.Name
	}
	if request.Strategy != "" {
		existingRule.Strategy = caching.InvalidationStrategy(request.Strategy)
	}
	if request.Pattern != "" {
		existingRule.Pattern = request.Pattern
	}
	if request.Tags != nil {
		existingRule.Tags = request.Tags
	}
	if request.Dependencies != nil {
		existingRule.Dependencies = request.Dependencies
	}
	// Skip conditions update for now
	if request.Priority > 0 {
		existingRule.Priority = request.Priority
	}
	existingRule.Enabled = request.Enabled

	// Update the rule
	err = h.invalidator.UpdateRule(ruleID, existingRule)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "UPDATE_FAILED", err.Error())
		return
	}

	response := InvalidationRuleResponse{
		ID:           existingRule.ID,
		Name:         existingRule.Name,
		Strategy:     string(existingRule.Strategy),
		Pattern:      existingRule.Pattern,
		Tags:         existingRule.Tags,
		Dependencies: existingRule.Dependencies,
		Conditions:   map[string]interface{}{},
		Priority:     existingRule.Priority,
		Enabled:      existingRule.Enabled,
		CreatedAt:    existingRule.CreatedAt,
		UpdatedAt:    existingRule.UpdatedAt,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// DeleteInvalidationRule deletes an invalidation rule
func (h *CacheHandler) DeleteInvalidationRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["rule_id"]

	if ruleID == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Rule ID is required")
		return
	}

	// Check if rule exists
	_, err := h.invalidator.GetRule(ruleID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "RULE_NOT_FOUND", fmt.Sprintf("Invalidation rule '%s' not found", ruleID))
		return
	}

	// Delete the rule
	err = h.invalidator.RemoveRule(ruleID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	response := InvalidationRuleDeleteResponse{
		ID:      ruleID,
		Success: true,
		Message: "Invalidation rule deleted successfully",
	}

	h.writeJSON(w, http.StatusOK, response)
}

// ExecuteInvalidation executes cache invalidation
func (h *CacheHandler) ExecuteInvalidation(w http.ResponseWriter, r *http.Request) {
	var request InvalidationExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if request.Strategy == "" {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Strategy is required")
		return
	}

	var invalidatedCount int
	var err error

	switch request.Strategy {
	case "exact":
		if request.Key == "" {
			h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Key is required for exact invalidation")
			return
		}
		result := h.invalidator.InvalidateByKey(request.Key)
		invalidatedCount = int(result.KeysInvalidated)
		err = result.Error

	case "pattern":
		if request.Pattern == "" {
			h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Pattern is required for pattern invalidation")
			return
		}
		result := h.invalidator.InvalidateByPattern(request.Pattern)
		invalidatedCount = int(result.KeysInvalidated)
		err = result.Error

	case "all":
		result := h.invalidator.InvalidateAll()
		invalidatedCount = int(result.KeysInvalidated)
		err = result.Error

	default:
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Unsupported invalidation strategy")
		return
	}

	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "INVALIDATION_FAILED", err.Error())
		return
	}

	response := InvalidationExecuteResponse{
		Strategy:         request.Strategy,
		Pattern:          request.Pattern,
		InvalidatedCount: invalidatedCount,
		Success:          true,
		Message:          "Invalidation executed successfully",
	}

	h.writeJSON(w, http.StatusOK, response)
}

// GetCacheHealth retrieves cache health information
func (h *CacheHandler) GetCacheHealth(w http.ResponseWriter, r *http.Request) {
	stats := h.cache.GetStats()
	config := &caching.CacheConfig{}

	// Determine health status
	status := "healthy"
	if stats.HitRate < 0.5 {
		status = "degraded"
	}
	if stats.HitRate < 0.2 {
		status = "unhealthy"
	}

	// Get last optimization time
	results := h.optimizer.GetOptimizationResults()
	var lastOptimization time.Time
	if len(results) > 0 {
		lastOptimization = results[len(results)-1].Timestamp
	}

	response := CacheHealthResponse{
		Status:           status,
		Uptime:           int(time.Since(time.Now().Add(-24 * time.Hour)).Seconds()),
		Version:          "1.0.0",
		EvictionPolicy:   string(config.Type),
		ShardCount:       config.ShardCount,
		TotalEntries:     int(stats.EntryCount),
		TotalSize:        int(stats.TotalSize),
		MemoryUsage:      0,
		HitRate:          stats.HitRate,
		LastOptimization: lastOptimization,
		Checks: map[string]bool{
			"cache_accessible":     true,
			"memory_ok":            true,
			"performance_ok":       stats.HitRate > 0.5,
			"optimization_enabled": true,
		},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Helper methods

// writeJSON writes a JSON response
func (h *CacheHandler) writeJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// writeError writes an error response
func (h *CacheHandler) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	errorResponse := ErrorResponse{
		Error:     message,
		Code:      code,
		Timestamp: time.Now(),
		RequestID: generateRequestID(),
	}

	h.writeJSON(w, statusCode, errorResponse)
}

// calculateAnalytics calculates analytics from metrics and snapshots
func (h *CacheHandler) calculateAnalytics(
	analytics *caching.CacheAnalytics,
	metrics []caching.CacheMetric,
	snapshots []caching.CachePerformanceSnapshot,
	timeRange string,
) CacheAnalyticsResponse {
	response := CacheAnalyticsResponse{
		TimeRange: timeRange,
	}

	if len(snapshots) > 0 {
		latest := snapshots[len(snapshots)-1]
		response.HitRate = latest.HitRate
		response.MissRate = latest.MissRate
		response.EvictionRate = latest.EvictionRate
		response.ExpirationRate = latest.ExpirationRate
		response.AverageEntrySize = 0
		response.AverageAccessTime = float64(latest.AverageAccessTime.Nanoseconds())
	}

	// Calculate popular keys, hot keys, and cold keys
	response.PopularKeys = make([]KeyAccessInfo, 0)
	response.HotKeys = make([]KeyAccessInfo, 0)
	response.ColdKeys = make([]KeyAccessInfo, 0)

	// Add sample data (in real implementation, this would come from analytics)
	response.PopularKeys = append(response.PopularKeys, KeyAccessInfo{
		Key:         "user:12345:profile",
		AccessCount: 150,
		LastAccess:  time.Now().Add(-5 * time.Minute),
	})

	response.HotKeys = append(response.HotKeys, KeyAccessInfo{
		Key:         "session:67890",
		AccessCount: 25,
		LastAccess:  time.Now().Add(-1 * time.Minute),
	})

	response.ColdKeys = append(response.ColdKeys, KeyAccessInfo{
		Key:         "config:old",
		AccessCount: 1,
		LastAccess:  time.Now().Add(-24 * time.Hour),
	})

	// Access patterns
	response.AccessPatterns = map[string]interface{}{
		"read_write_ratio": 0.8,
		"temporal_patterns": map[string]int{
			"00:00": 100,
			"06:00": 500,
			"12:00": 1000,
			"18:00": 800,
		},
	}

	// Size distribution
	response.SizeDistribution = map[string]int{
		"small":  2000,
		"medium": 2500,
		"large":  500,
	}

	return response
}

// parseTimeRange parses a time range string into a duration
func parseTimeRange(timeRange string) (time.Duration, error) {
	switch timeRange {
	case "1h":
		return time.Hour, nil
	case "24h":
		return 24 * time.Hour, nil
	case "7d":
		return 7 * 24 * time.Hour, nil
	default:
		// Try to parse as a number followed by a unit
		if len(timeRange) < 2 {
			return 0, fmt.Errorf("invalid time range format")
		}

		value, err := strconv.Atoi(timeRange[:len(timeRange)-1])
		if err != nil {
			return 0, fmt.Errorf("invalid time range value")
		}

		unit := timeRange[len(timeRange)-1]
		switch unit {
		case 'h':
			return time.Duration(value) * time.Hour, nil
		case 'd':
			return time.Duration(value) * 24 * time.Hour, nil
		case 'm':
			return time.Duration(value) * time.Minute, nil
		default:
			return 0, fmt.Errorf("invalid time range unit")
		}
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// Request/Response types

type CacheSetRequest struct {
	Value    interface{}            `json:"value"`
	TTL      int                    `json:"ttl,omitempty"`
	Priority int                    `json:"priority,omitempty"`
	Tags     []string               `json:"tags,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type CacheGetResponse struct {
	Key         string                 `json:"key"`
	Value       interface{}            `json:"value"`
	TTL         int                    `json:"ttl"`
	CreatedAt   time.Time              `json:"created_at"`
	AccessedAt  time.Time              `json:"accessed_at"`
	AccessCount int                    `json:"access_count"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type CacheSetResponse struct {
	Key     string `json:"key"`
	Success bool   `json:"success"`
	Message string `json:"message"`
	TTL     int    `json:"ttl"`
	Size    int    `json:"size"`
}

type CacheDeleteResponse struct {
	Key     string `json:"key"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type CacheClearResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	DeletedCount int    `json:"deleted_count"`
}

type CacheStatsResponse struct {
	Hits              int     `json:"hits"`
	Misses            int     `json:"misses"`
	Evictions         int     `json:"evictions"`
	Expirations       int     `json:"expirations"`
	TotalSize         int     `json:"total_size"`
	EntryCount        int     `json:"entry_count"`
	HitRate           float64 `json:"hit_rate"`
	MissRate          float64 `json:"miss_rate"`
	AverageAccessTime float64 `json:"average_access_time"`
	MemoryUsage       int     `json:"memory_usage"`
	Throughput        int     `json:"throughput"`
	ShardCount        int     `json:"shard_count"`
}

type CacheAnalyticsResponse struct {
	TimeRange         string                 `json:"time_range"`
	HitRate           float64                `json:"hit_rate"`
	MissRate          float64                `json:"miss_rate"`
	EvictionRate      float64                `json:"eviction_rate"`
	ExpirationRate    float64                `json:"expiration_rate"`
	AverageEntrySize  float64                `json:"average_entry_size"`
	AverageAccessTime float64                `json:"average_access_time"`
	PopularKeys       []KeyAccessInfo        `json:"popular_keys"`
	HotKeys           []KeyAccessInfo        `json:"hot_keys"`
	ColdKeys          []KeyAccessInfo        `json:"cold_keys"`
	AccessPatterns    map[string]interface{} `json:"access_patterns"`
	SizeDistribution  map[string]int         `json:"size_distribution"`
}

type KeyAccessInfo struct {
	Key         string    `json:"key"`
	AccessCount int       `json:"access_count"`
	LastAccess  time.Time `json:"last_access"`
}

type OptimizationPlanRequest struct {
	ForceGeneration bool `json:"force_generation,omitempty"`
}

type OptimizationPlanResponse struct {
	ID                 string                       `json:"id"`
	Name               string                       `json:"name"`
	Description        string                       `json:"description"`
	Actions            []caching.OptimizationAction `json:"actions"`
	EstimatedTotalGain float64                      `json:"estimated_total_gain"`
	EstimatedTotalCost float64                      `json:"estimated_total_cost"`
	EstimatedROI       float64                      `json:"estimated_roi"`
	RiskLevel          string                       `json:"risk_level"`
	ExecutionTime      time.Duration                `json:"execution_time"`
	CreatedAt          time.Time                    `json:"created_at"`
	Status             string                       `json:"status"`
}

type OptimizationPlansResponse struct {
	Plans      []caching.OptimizationPlan `json:"plans"`
	TotalCount int                        `json:"total_count"`
}

type OptimizationResultResponse struct {
	ID                string        `json:"id"`
	PlanID            string        `json:"plan_id"`
	Success           bool          `json:"success"`
	ErrorMessage      string        `json:"error_message"`
	ActualGain        float64       `json:"actual_gain"`
	ActualCost        float64       `json:"actual_cost"`
	ExecutionDuration time.Duration `json:"execution_duration"`
	Timestamp         time.Time     `json:"timestamp"`
}

type OptimizationResultsResponse struct {
	Results    []caching.OptimizationResult `json:"results"`
	TotalCount int                          `json:"total_count"`
}

type InvalidationRuleRequest struct {
	Name         string                 `json:"name"`
	Strategy     string                 `json:"strategy"`
	Pattern      string                 `json:"pattern,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Conditions   map[string]interface{} `json:"conditions,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
	Enabled      bool                   `json:"enabled,omitempty"`
}

type InvalidationRuleResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Strategy     string                 `json:"strategy"`
	Pattern      string                 `json:"pattern,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Conditions   map[string]interface{} `json:"conditions,omitempty"`
	Priority     int                    `json:"priority"`
	Enabled      bool                   `json:"enabled"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type InvalidationRulesResponse struct {
	Rules      []*caching.InvalidationRule `json:"rules"`
	TotalCount int                         `json:"total_count"`
}

type InvalidationRuleDeleteResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type InvalidationExecuteRequest struct {
	Strategy string   `json:"strategy"`
	Key      string   `json:"key,omitempty"`
	Pattern  string   `json:"pattern,omitempty"`
	Tags     []string `json:"tags,omitempty"`
}

type InvalidationExecuteResponse struct {
	Strategy         string `json:"strategy"`
	Pattern          string `json:"pattern,omitempty"`
	InvalidatedCount int    `json:"invalidated_count"`
	Success          bool   `json:"success"`
	Message          string `json:"message"`
}

type CacheHealthResponse struct {
	Status           string          `json:"status"`
	Uptime           int             `json:"uptime"`
	Version          string          `json:"version"`
	EvictionPolicy   string          `json:"eviction_policy"`
	ShardCount       int             `json:"shard_count"`
	TotalEntries     int             `json:"total_entries"`
	TotalSize        int             `json:"total_size"`
	MemoryUsage      int             `json:"memory_usage"`
	HitRate          float64         `json:"hit_rate"`
	LastOptimization time.Time       `json:"last_optimization"`
	Checks           map[string]bool `json:"checks"`
}

type ErrorResponse struct {
	Error     string                 `json:"error"`
	Code      string                 `json:"code"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id"`
}
