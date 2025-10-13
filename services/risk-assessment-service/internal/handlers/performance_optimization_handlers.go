package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/performance"
)

// PerformanceOptimizationHandlers handles performance optimization endpoints
type PerformanceOptimizationHandlers struct {
	optimizationService   *performance.OptimizationService
	databaseOptimizer     *performance.DatabaseOptimizer
	responseTimeOptimizer *performance.ResponseTimeOptimizer
	logger                *zap.Logger
}

// NewPerformanceOptimizationHandlers creates new performance optimization handlers
func NewPerformanceOptimizationHandlers(
	optimizationService *performance.OptimizationService,
	databaseOptimizer *performance.DatabaseOptimizer,
	responseTimeOptimizer *performance.ResponseTimeOptimizer,
	logger *zap.Logger,
) *PerformanceOptimizationHandlers {
	return &PerformanceOptimizationHandlers{
		optimizationService:   optimizationService,
		databaseOptimizer:     databaseOptimizer,
		responseTimeOptimizer: responseTimeOptimizer,
		logger:                logger,
	}
}

// GetOptimizationStatus returns the current optimization status
func (h *PerformanceOptimizationHandlers) GetOptimizationStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting optimization status")

	// Get optimization metrics
	metrics := h.optimizationService.GetMetrics()

	// Get database stats
	dbStats := h.databaseOptimizer.GetStats()

	// Get response time stats
	responseTimeStats := h.responseTimeOptimizer.GetStats()

	// Create comprehensive status response
	status := map[string]interface{}{
		"overall": map[string]interface{}{
			"is_optimized":       metrics.IsOptimized,
			"optimization_score": metrics.OptimizationScore,
			"last_optimized":     metrics.LastOptimized,
			"last_updated":       metrics.LastUpdated,
		},
		"response_times": map[string]interface{}{
			"p95_latency":        responseTimeStats.P95Latency,
			"p99_latency":        responseTimeStats.P99Latency,
			"avg_latency":        responseTimeStats.AvgLatency,
			"min_latency":        responseTimeStats.MinLatency,
			"max_latency":        responseTimeStats.MaxLatency,
			"p95_target_met":     responseTimeStats.P95TargetMet,
			"p99_target_met":     responseTimeStats.P99TargetMet,
			"avg_target_met":     responseTimeStats.AvgTargetMet,
			"max_target_met":     responseTimeStats.MaxTargetMet,
			"is_optimized":       responseTimeStats.IsOptimized,
			"optimization_score": responseTimeStats.OptimizationScore,
		},
		"database": map[string]interface{}{
			"active_connections": dbStats.ActiveConnections,
			"idle_connections":   dbStats.IdleConnections,
			"total_connections":  dbStats.TotalConnections,
			"wait_count":         dbStats.WaitCount,
			"wait_duration":      dbStats.WaitDuration,
			"total_queries":      dbStats.TotalQueries,
			"slow_queries":       dbStats.SlowQueries,
			"average_query_time": dbStats.AverageQueryTime,
			"max_query_time":     dbStats.MaxQueryTime,
			"is_optimized":       dbStats.IsOptimized,
			"optimization_score": dbStats.OptimizationScore,
		},
		"cache": map[string]interface{}{
			"hit_rate":   metrics.CacheHitRate,
			"cache_size": metrics.CacheSize,
			"evictions":  metrics.CacheEvictions,
		},
		"system": map[string]interface{}{
			"memory_usage_mb":   metrics.MemoryUsageMB,
			"goroutine_count":   metrics.GoroutineCount,
			"cpu_usage_percent": metrics.CPUUsagePercent,
			"gc_percent":        metrics.GCPercent,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		h.logger.Error("Failed to encode optimization status", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// OptimizeNow triggers immediate optimization
func (h *PerformanceOptimizationHandlers) OptimizeNow(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Triggering immediate optimization")

	// Perform optimization
	if err := h.optimizationService.OptimizeNow(); err != nil {
		h.logger.Error("Failed to perform optimization", zap.Error(err))
		http.Error(w, "Optimization failed", http.StatusInternalServerError)
		return
	}

	// Get updated metrics
	metrics := h.optimizationService.GetMetrics()

	response := map[string]interface{}{
		"message":            "Optimization completed successfully",
		"optimization_score": metrics.OptimizationScore,
		"is_optimized":       metrics.IsOptimized,
		"last_optimized":     metrics.LastOptimized,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode optimization response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetDatabaseOptimization returns database optimization status
func (h *PerformanceOptimizationHandlers) GetDatabaseOptimization(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting database optimization status")

	// Get database stats
	stats := h.databaseOptimizer.GetStats()

	// Get recommendations
	recommendations := h.databaseOptimizer.GetOptimizationRecommendations()

	response := map[string]interface{}{
		"stats":           stats,
		"recommendations": recommendations,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode database optimization response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// OptimizeDatabase triggers database optimization
func (h *PerformanceOptimizationHandlers) OptimizeDatabase(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Triggering database optimization")

	// Perform database optimization
	if err := h.databaseOptimizer.Optimize(r.Context()); err != nil {
		h.logger.Error("Failed to perform database optimization", zap.Error(err))
		http.Error(w, "Database optimization failed", http.StatusInternalServerError)
		return
	}

	// Get updated stats
	stats := h.databaseOptimizer.GetStats()

	response := map[string]interface{}{
		"message":            "Database optimization completed successfully",
		"optimization_score": stats.OptimizationScore,
		"is_optimized":       stats.IsOptimized,
		"last_optimized":     stats.LastOptimized,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode database optimization response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetResponseTimeOptimization returns response time optimization status
func (h *PerformanceOptimizationHandlers) GetResponseTimeOptimization(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting response time optimization status")

	// Get response time stats
	stats := h.responseTimeOptimizer.GetStats()

	// Get recommendations
	recommendations := h.responseTimeOptimizer.GetOptimizationRecommendations()

	response := map[string]interface{}{
		"stats":           stats,
		"recommendations": recommendations,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response time optimization response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// OptimizeResponseTime triggers response time optimization
func (h *PerformanceOptimizationHandlers) OptimizeResponseTime(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Triggering response time optimization")

	// Perform response time optimization
	if err := h.responseTimeOptimizer.Optimize(r.Context()); err != nil {
		h.logger.Error("Failed to perform response time optimization", zap.Error(err))
		http.Error(w, "Response time optimization failed", http.StatusInternalServerError)
		return
	}

	// Get updated stats
	stats := h.responseTimeOptimizer.GetStats()

	response := map[string]interface{}{
		"message":            "Response time optimization completed successfully",
		"optimization_score": stats.OptimizationScore,
		"is_optimized":       stats.IsOptimized,
		"last_optimized":     stats.LastOptimized,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response time optimization response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// SetResponseTimeTargets sets new response time targets
func (h *PerformanceOptimizationHandlers) SetResponseTimeTargets(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Setting response time targets")

	// Parse request body
	var request struct {
		P95Target string `json:"p95_target"`
		P99Target string `json:"p99_target"`
		AvgTarget string `json:"avg_target"`
		MaxTarget string `json:"max_target"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse durations
	p95Target, err := time.ParseDuration(request.P95Target)
	if err != nil {
		http.Error(w, "Invalid P95 target duration", http.StatusBadRequest)
		return
	}

	p99Target, err := time.ParseDuration(request.P99Target)
	if err != nil {
		http.Error(w, "Invalid P99 target duration", http.StatusBadRequest)
		return
	}

	avgTarget, err := time.ParseDuration(request.AvgTarget)
	if err != nil {
		http.Error(w, "Invalid average target duration", http.StatusBadRequest)
		return
	}

	maxTarget, err := time.ParseDuration(request.MaxTarget)
	if err != nil {
		http.Error(w, "Invalid max target duration", http.StatusBadRequest)
		return
	}

	// Set targets
	h.responseTimeOptimizer.SetTargets(p95Target, p99Target, avgTarget, maxTarget)

	response := map[string]interface{}{
		"message":    "Response time targets updated successfully",
		"p95_target": p95Target,
		"p99_target": p99Target,
		"avg_target": avgTarget,
		"max_target": maxTarget,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetOptimizationRecommendations returns optimization recommendations
func (h *PerformanceOptimizationHandlers) GetOptimizationRecommendations(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting optimization recommendations")

	// Get recommendations from all optimizers
	dbRecommendations := h.databaseOptimizer.GetOptimizationRecommendations()
	responseTimeRecommendations := h.responseTimeOptimizer.GetOptimizationRecommendations()

	// Get current metrics
	metrics := h.optimizationService.GetMetrics()

	response := map[string]interface{}{
		"overall_score": metrics.OptimizationScore,
		"is_optimized":  metrics.IsOptimized,
		"recommendations": map[string]interface{}{
			"database":      dbRecommendations,
			"response_time": responseTimeRecommendations,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode recommendations response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// RecordResponseTime records a response time measurement
func (h *PerformanceOptimizationHandlers) RecordResponseTime(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Recording response time")

	// Parse request body
	var request struct {
		Duration string `json:"duration"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse duration
	duration, err := time.ParseDuration(request.Duration)
	if err != nil {
		http.Error(w, "Invalid duration format", http.StatusBadRequest)
		return
	}

	// Record response time
	h.responseTimeOptimizer.RecordResponseTime(duration)

	response := map[string]interface{}{
		"message":  "Response time recorded successfully",
		"duration": duration,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// GetPerformanceTargets returns current performance targets
func (h *PerformanceOptimizationHandlers) GetPerformanceTargets(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting performance targets")

	// Get current targets from response time optimizer
	stats := h.responseTimeOptimizer.GetStats()

	response := map[string]interface{}{
		"response_time_targets": map[string]interface{}{
			"p95_target": stats.P95Latency, // This should be from config, but we'll use current for now
			"p99_target": stats.P99Latency,
			"avg_target": stats.AvgLatency,
			"max_target": stats.MaxLatency,
		},
		"targets_met": map[string]interface{}{
			"p95_target_met": stats.P95TargetMet,
			"p99_target_met": stats.P99TargetMet,
			"avg_target_met": stats.AvgTargetMet,
			"max_target_met": stats.MaxTargetMet,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode performance targets response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ResetOptimization resets optimization data
func (h *PerformanceOptimizationHandlers) ResetOptimization(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Resetting optimization data")

	// Reset response time optimizer
	h.responseTimeOptimizer.Reset()

	response := map[string]interface{}{
		"message": "Optimization data reset successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode reset response", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
