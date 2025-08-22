package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

// MemoryOptimizationDashboardHandler provides HTTP handlers for memory optimization
type MemoryOptimizationDashboardHandler struct {
	memorySystem *observability.MemoryOptimizationSystem
	logger       *zap.Logger
}

// NewMemoryOptimizationDashboardHandler creates a new memory optimization dashboard handler
func NewMemoryOptimizationDashboardHandler(
	memorySystem *observability.MemoryOptimizationSystem,
	logger *zap.Logger,
) *MemoryOptimizationDashboardHandler {
	return &MemoryOptimizationDashboardHandler{
		memorySystem: memorySystem,
		logger:       logger,
	}
}

// GetCurrentMetrics returns current memory metrics
func (h *MemoryOptimizationDashboardHandler) GetCurrentMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	metrics, err := h.memorySystem.GetCurrentMetrics(ctx)
	if err != nil {
		h.logger.Error("failed to get current metrics", zap.Error(err))
		http.Error(w, "Failed to get current metrics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    metrics,
	})
}

// GetMetricsHistory returns memory metrics history
func (h *MemoryOptimizationDashboardHandler) GetMetricsHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse duration parameter
	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "1h" // Default to 1 hour
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		http.Error(w, "Invalid duration parameter", http.StatusBadRequest)
		return
	}

	metrics, err := h.memorySystem.GetMetricsHistory(ctx, duration)
	if err != nil {
		h.logger.Error("failed to get metrics history", zap.Error(err))
		http.Error(w, "Failed to get metrics history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    metrics,
		"count":   len(metrics),
	})
}

// TakeHeapProfile takes a heap profile
func (h *MemoryOptimizationDashboardHandler) TakeHeapProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profile, err := h.memorySystem.TakeHeapProfile(ctx)
	if err != nil {
		h.logger.Error("failed to take heap profile", zap.Error(err))
		http.Error(w, "Failed to take heap profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    profile,
	})
}

// TakeGoroutineProfile takes a goroutine profile
func (h *MemoryOptimizationDashboardHandler) TakeGoroutineProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	profile, err := h.memorySystem.TakeGoroutineProfile(ctx)
	if err != nil {
		h.logger.Error("failed to take goroutine profile", zap.Error(err))
		http.Error(w, "Failed to take goroutine profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    profile,
	})
}

// DetectLeaks detects memory leaks
func (h *MemoryOptimizationDashboardHandler) DetectLeaks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	detection, err := h.memorySystem.DetectLeaks(ctx)
	if err != nil {
		h.logger.Error("failed to detect leaks", zap.Error(err))
		http.Error(w, "Failed to detect leaks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    detection,
	})
}

// OptimizeMemory performs memory optimization
func (h *MemoryOptimizationDashboardHandler) OptimizeMemory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	optimization, err := h.memorySystem.OptimizeMemory(ctx)
	if err != nil {
		h.logger.Error("failed to optimize memory", zap.Error(err))
		http.Error(w, "Failed to optimize memory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    optimization,
	})
}

// ForceGC forces garbage collection
func (h *MemoryOptimizationDashboardHandler) ForceGC(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.memorySystem.ForceGC(ctx)
	if err != nil {
		h.logger.Error("failed to force GC", zap.Error(err))
		http.Error(w, "Failed to force GC", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Garbage collection completed successfully",
	})
}

// AnalyzeMemory performs comprehensive memory analysis
func (h *MemoryOptimizationDashboardHandler) AnalyzeMemory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse duration parameter
	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "24h" // Default to 24 hours
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		http.Error(w, "Invalid duration parameter", http.StatusBadRequest)
		return
	}

	analysis, err := h.memorySystem.AnalyzeMemory(ctx, duration)
	if err != nil {
		h.logger.Error("failed to analyze memory", zap.Error(err))
		http.Error(w, "Failed to analyze memory", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    analysis,
	})
}

// GetOptimizations returns optimization history
func (h *MemoryOptimizationDashboardHandler) GetOptimizations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	optimizations, err := h.memorySystem.GetOptimizations(ctx)
	if err != nil {
		h.logger.Error("failed to get optimizations", zap.Error(err))
		http.Error(w, "Failed to get optimizations", http.StatusInternalServerError)
		return
	}

	// Apply limit
	if len(optimizations) > limit {
		optimizations = optimizations[len(optimizations)-limit:]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    optimizations,
		"count":   len(optimizations),
	})
}

// GetLeakDetections returns leak detection history
func (h *MemoryOptimizationDashboardHandler) GetLeakDetections(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	detections, err := h.memorySystem.GetLeakDetections(ctx)
	if err != nil {
		h.logger.Error("failed to get leak detections", zap.Error(err))
		http.Error(w, "Failed to get leak detections", http.StatusInternalServerError)
		return
	}

	// Apply limit
	if len(detections) > limit {
		detections = detections[len(detections)-limit:]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    detections,
		"count":   len(detections),
	})
}

// GetConfiguration returns the current configuration
func (h *MemoryOptimizationDashboardHandler) GetConfiguration(w http.ResponseWriter, r *http.Request) {
	config := h.memorySystem.GetConfiguration()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    config,
	})
}

// UpdateConfiguration updates the system configuration
func (h *MemoryOptimizationDashboardHandler) UpdateConfiguration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var config observability.MemoryOptimizationConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.memorySystem.UpdateConfiguration(&config); err != nil {
		h.logger.Error("failed to update configuration", zap.Error(err))
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Configuration updated successfully",
	})
}

// GetStatus returns the system status
func (h *MemoryOptimizationDashboardHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	status := h.memorySystem.GetStatus()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    status,
	})
}

// StartSystem starts the memory optimization system
func (h *MemoryOptimizationDashboardHandler) StartSystem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.memorySystem.Start(ctx); err != nil {
		h.logger.Error("failed to start memory optimization system", zap.Error(err))
		http.Error(w, "Failed to start memory optimization system", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Memory optimization system started successfully",
	})
}

// StopSystem stops the memory optimization system
func (h *MemoryOptimizationDashboardHandler) StopSystem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if err := h.memorySystem.Stop(ctx); err != nil {
		h.logger.Error("failed to stop memory optimization system", zap.Error(err))
		http.Error(w, "Failed to stop memory optimization system", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Memory optimization system stopped successfully",
	})
}

// GetSystemHealth returns system health information
func (h *MemoryOptimizationDashboardHandler) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	status := h.memorySystem.GetStatus()
	metrics, err := h.memorySystem.GetCurrentMetrics(ctx)
	if err != nil {
		h.logger.Error("failed to get current metrics for health check", zap.Error(err))
		http.Error(w, "Failed to get system health", http.StatusInternalServerError)
		return
	}

	health := map[string]interface{}{
		"status":          status,
		"current_metrics": metrics,
		"system_healthy":  true,
		"last_check":      time.Now(),
		"recommendations": []string{},
	}

	// Add health checks
	if metrics != nil {
		if metrics.HeapAllocPercent > 90 {
			health["system_healthy"] = false
			health["recommendations"] = append(health["recommendations"].([]string), "High memory usage detected")
		}

		if metrics.Goroutines > 1000 {
			health["system_healthy"] = false
			health["recommendations"] = append(health["recommendations"].([]string), "High goroutine count detected")
		}

		if metrics.GCCPUFraction > 0.2 {
			health["system_healthy"] = false
			health["recommendations"] = append(health["recommendations"].([]string), "High GC CPU fraction detected")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    health,
	})
}

// GetSystemMetrics returns comprehensive system metrics
func (h *MemoryOptimizationDashboardHandler) GetSystemMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get current metrics
	metrics, err := h.memorySystem.GetCurrentMetrics(ctx)
	if err != nil {
		h.logger.Error("failed to get current metrics", zap.Error(err))
		http.Error(w, "Failed to get system metrics", http.StatusInternalServerError)
		return
	}

	// Get recent optimizations
	optimizations, err := h.memorySystem.GetOptimizations(ctx)
	if err != nil {
		h.logger.Error("failed to get optimizations", zap.Error(err))
		// Continue without optimizations
	}

	// Get recent leak detections
	detections, err := h.memorySystem.GetLeakDetections(ctx)
	if err != nil {
		h.logger.Error("failed to get leak detections", zap.Error(err))
		// Continue without detections
	}

	// Calculate summary statistics
	var totalOptimizations int
	var successfulOptimizations int
	var totalSavings uint64

	for _, opt := range optimizations {
		totalOptimizations++
		if opt.Success {
			successfulOptimizations++
			totalSavings += opt.EstimatedSavings
		}
	}

	successRate := 0.0
	if totalOptimizations > 0 {
		successRate = float64(successfulOptimizations) / float64(totalOptimizations) * 100
	}

	systemMetrics := map[string]interface{}{
		"current_metrics": metrics,
		"optimization_summary": map[string]interface{}{
			"total_optimizations":      totalOptimizations,
			"successful_optimizations": successfulOptimizations,
			"success_rate":             successRate,
			"total_estimated_savings":  totalSavings,
		},
		"leak_detection_summary": map[string]interface{}{
			"total_detections": len(detections),
			"recent_detections": func() []*observability.MemoryLeakDetection {
				if len(detections) > 5 {
					return detections[len(detections)-5:]
				}
				return detections
			}(),
		},
		"system_status": h.memorySystem.GetStatus(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    systemMetrics,
	})
}

// GetOptimizationRecommendations returns optimization recommendations
func (h *MemoryOptimizationDashboardHandler) GetOptimizationRecommendations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Perform analysis to get recommendations
	analysis, err := h.memorySystem.AnalyzeMemory(ctx, 24*time.Hour)
	if err != nil {
		h.logger.Error("failed to analyze memory for recommendations", zap.Error(err))
		http.Error(w, "Failed to get optimization recommendations", http.StatusInternalServerError)
		return
	}

	recommendations := map[string]interface{}{
		"recommendations": analysis.Recommendations,
		"risk_assessment": analysis.RiskAssessment,
		"patterns":        analysis.Patterns,
		"trends":          analysis.Trends,
		"anomalies":       analysis.Anomalies,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    recommendations,
	})
}

// ExportMetrics exports metrics in various formats
func (h *MemoryOptimizationDashboardHandler) ExportMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse parameters
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "json"
	}

	durationStr := r.URL.Query().Get("duration")
	if durationStr == "" {
		durationStr = "24h"
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		http.Error(w, "Invalid duration parameter", http.StatusBadRequest)
		return
	}

	// Get metrics
	metrics, err := h.memorySystem.GetMetricsHistory(ctx, duration)
	if err != nil {
		h.logger.Error("failed to get metrics for export", zap.Error(err))
		http.Error(w, "Failed to export metrics", http.StatusInternalServerError)
		return
	}

	// Export based on format
	switch format {
	case "json":
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=memory_metrics_%s.json", time.Now().Format("2006-01-02")))
		json.NewEncoder(w).Encode(map[string]interface{}{
			"export_timestamp": time.Now(),
			"duration":         duration.String(),
			"metrics_count":    len(metrics),
			"metrics":          metrics,
		})
	case "csv":
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=memory_metrics_%s.csv", time.Now().Format("2006-01-02")))

		// Write CSV header
		fmt.Fprintf(w, "Timestamp,HeapAlloc,HeapSys,HeapInuse,HeapIdle,Goroutines,Threads,GCCPUFraction\n")

		// Write data
		for _, metric := range metrics {
			fmt.Fprintf(w, "%s,%d,%d,%d,%d,%d,%d,%.6f\n",
				metric.Timestamp.Format(time.RFC3339),
				metric.HeapAlloc,
				metric.HeapSys,
				metric.HeapInuse,
				metric.HeapIdle,
				metric.Goroutines,
				metric.Threads,
				metric.GCCPUFraction,
			)
		}
	default:
		http.Error(w, "Unsupported export format", http.StatusBadRequest)
	}
}
