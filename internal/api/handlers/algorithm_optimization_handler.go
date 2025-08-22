package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/modules/classification_optimization"
)

// AlgorithmOptimizationHandler handles algorithm optimization API requests
type AlgorithmOptimizationHandler struct {
	optimizer *classification_optimization.AlgorithmOptimizer
	logger    *zap.Logger
}

// NewAlgorithmOptimizationHandler creates a new algorithm optimization handler
func NewAlgorithmOptimizationHandler(optimizer *classification_optimization.AlgorithmOptimizer, logger *zap.Logger) *AlgorithmOptimizationHandler {
	return &AlgorithmOptimizationHandler{
		optimizer: optimizer,
		logger:    logger,
	}
}

// AnalyzeAndOptimizeHandler triggers analysis and optimization
func (h *AlgorithmOptimizationHandler) AnalyzeAndOptimizeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req struct {
		ForceOptimization bool `json:"force_optimization"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Perform analysis and optimization
	err := h.optimizer.AnalyzeAndOptimize(ctx)
	if err != nil {
		h.logger.Error("Failed to analyze and optimize", zap.Error(err))
		http.Error(w, "Optimization failed", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"status":  "success",
		"message": "Analysis and optimization completed",
		"time":    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetOptimizationHistoryHandler returns optimization history
func (h *AlgorithmOptimizationHandler) GetOptimizationHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Get optimization history
	history := h.optimizer.GetOptimizationHistory()
	if len(history) > limit {
		history = history[len(history)-limit:]
	}

	response := map[string]interface{}{
		"history": history,
		"total":   len(history),
		"limit":   limit,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetActiveOptimizationsHandler returns currently active optimizations
func (h *AlgorithmOptimizationHandler) GetActiveOptimizationsHandler(w http.ResponseWriter, r *http.Request) {
	active := h.optimizer.GetActiveOptimizations()

	response := map[string]interface{}{
		"active_optimizations": active,
		"count":                len(active),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetOptimizationSummaryHandler returns optimization summary
func (h *AlgorithmOptimizationHandler) GetOptimizationSummaryHandler(w http.ResponseWriter, r *http.Request) {
	summary := h.optimizer.GetOptimizationSummary()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(summary)
}

// GetOptimizationByIDHandler returns a specific optimization by ID
func (h *AlgorithmOptimizationHandler) GetOptimizationByIDHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	optimizationID := vars["id"]

	if optimizationID == "" {
		http.Error(w, "Optimization ID is required", http.StatusBadRequest)
		return
	}

	// Get optimization history and find the specific one
	history := h.optimizer.GetOptimizationHistory()
	var targetOptimization *classification_optimization.OptimizationResult

	for _, opt := range history {
		if opt.ID == optimizationID {
			targetOptimization = opt
			break
		}
	}

	if targetOptimization == nil {
		http.Error(w, "Optimization not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(targetOptimization)
}

// GetOptimizationsByTypeHandler returns optimizations by type
func (h *AlgorithmOptimizationHandler) GetOptimizationsByTypeHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	optimizationType := vars["type"]

	if optimizationType == "" {
		http.Error(w, "Optimization type is required", http.StatusBadRequest)
		return
	}

	// Get optimization history and filter by type
	history := h.optimizer.GetOptimizationHistory()
	var filteredOptimizations []*classification_optimization.OptimizationResult

	for _, opt := range history {
		if string(opt.OptimizationType) == optimizationType {
			filteredOptimizations = append(filteredOptimizations, opt)
		}
	}

	response := map[string]interface{}{
		"optimizations": filteredOptimizations,
		"type":          optimizationType,
		"count":         len(filteredOptimizations),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetOptimizationsByAlgorithmHandler returns optimizations by algorithm
func (h *AlgorithmOptimizationHandler) GetOptimizationsByAlgorithmHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	algorithmID := vars["algorithm_id"]

	if algorithmID == "" {
		http.Error(w, "Algorithm ID is required", http.StatusBadRequest)
		return
	}

	// Get optimization history and filter by algorithm
	history := h.optimizer.GetOptimizationHistory()
	var filteredOptimizations []*classification_optimization.OptimizationResult

	for _, opt := range history {
		if opt.AlgorithmID == algorithmID {
			filteredOptimizations = append(filteredOptimizations, opt)
		}
	}

	response := map[string]interface{}{
		"optimizations": filteredOptimizations,
		"algorithm_id":  algorithmID,
		"count":         len(filteredOptimizations),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CancelOptimizationHandler cancels an active optimization
func (h *AlgorithmOptimizationHandler) CancelOptimizationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	optimizationID := vars["id"]

	if optimizationID == "" {
		http.Error(w, "Optimization ID is required", http.StatusBadRequest)
		return
	}

	// Check if optimization is active
	active := h.optimizer.GetActiveOptimizations()
	if _, exists := active[optimizationID]; !exists {
		http.Error(w, "Optimization not found or not active", http.StatusNotFound)
		return
	}

	// In a real implementation, you would cancel the optimization
	// For now, we'll just return a success response
	response := map[string]interface{}{
		"status":  "success",
		"message": "Optimization cancellation requested",
		"id":      optimizationID,
		"time":    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RollbackOptimizationHandler rolls back an optimization
func (h *AlgorithmOptimizationHandler) RollbackOptimizationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	optimizationID := vars["id"]

	if optimizationID == "" {
		http.Error(w, "Optimization ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Reason string `json:"reason"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// In a real implementation, you would rollback the optimization
	// For now, we'll just return a success response
	response := map[string]interface{}{
		"status":  "success",
		"message": "Optimization rollback requested",
		"id":      optimizationID,
		"reason":  req.Reason,
		"time":    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetOptimizationRecommendationsHandler returns optimization recommendations
func (h *AlgorithmOptimizationHandler) GetOptimizationRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, this would analyze current patterns and generate recommendations
	// For now, we'll return a placeholder response
	recommendations := []map[string]interface{}{
		{
			"id":          "rec-001",
			"type":        "threshold",
			"priority":    "high",
			"title":       "Adjust confidence thresholds",
			"description": "High-confidence misclassifications detected",
			"impact":      "medium",
			"effort":      "low",
			"confidence":  0.8,
		},
		{
			"id":          "rec-002",
			"type":        "features",
			"priority":    "medium",
			"title":       "Enhance feature extraction",
			"description": "Semantic patterns suggest feature improvements",
			"impact":      "high",
			"effort":      "medium",
			"confidence":  0.6,
		},
	}

	response := map[string]interface{}{
		"recommendations": recommendations,
		"count":           len(recommendations),
		"generated_at":    time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
