package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/external"
)

// ContinuousImprovementHandler handles continuous improvement API requests
type ContinuousImprovementHandler struct {
	manager *external.ContinuousImprovementManager
	logger  *zap.Logger
}

// NewContinuousImprovementHandler creates a new continuous improvement handler
func NewContinuousImprovementHandler(manager *external.ContinuousImprovementManager, logger *zap.Logger) *ContinuousImprovementHandler {
	return &ContinuousImprovementHandler{
		manager: manager,
		logger:  logger,
	}
}

// RegisterRoutes registers the continuous improvement routes
func (h *ContinuousImprovementHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/v1/continuous-improvement/analyze", h.AnalyzeAndRecommend).Methods("POST")
	router.HandleFunc("/api/v1/continuous-improvement/apply", h.ApplyImprovement).Methods("POST")
	router.HandleFunc("/api/v1/continuous-improvement/evaluate/{strategyID}", h.EvaluateStrategy).Methods("GET")
	router.HandleFunc("/api/v1/continuous-improvement/rollback/{strategyID}", h.RollbackStrategy).Methods("POST")
	router.HandleFunc("/api/v1/continuous-improvement/strategies", h.GetActiveStrategies).Methods("GET")
	router.HandleFunc("/api/v1/continuous-improvement/history", h.GetImprovementHistory).Methods("GET")
	router.HandleFunc("/api/v1/continuous-improvement/config", h.GetConfig).Methods("GET")
	router.HandleFunc("/api/v1/continuous-improvement/config", h.UpdateConfig).Methods("PUT")
}

// AnalyzeAndRecommendRequest represents a request to analyze and generate recommendations
type AnalyzeAndRecommendRequest struct {
	ForceAnalysis bool `json:"force_analysis"`
}

// AnalyzeAndRecommendResponse represents the response from analyze and recommend
type AnalyzeAndRecommendResponse struct {
	Success         bool                                  `json:"success"`
	Recommendations []*external.ImprovementRecommendation `json:"recommendations"`
	TotalCount      int                                   `json:"total_count"`
	AnalysisTime    time.Time                             `json:"analysis_time"`
	Message         string                                `json:"message,omitempty"`
}

// ApplyImprovementRequest represents a request to apply an improvement
type ApplyImprovementRequest struct {
	RecommendationID string                              `json:"recommendation_id"`
	Recommendation   *external.ImprovementRecommendation `json:"recommendation"`
}

// ApplyImprovementResponse represents the response from applying an improvement
type ApplyImprovementResponse struct {
	Success  bool                          `json:"success"`
	Strategy *external.ImprovementStrategy `json:"strategy"`
	Message  string                        `json:"message,omitempty"`
}

// EvaluateStrategyResponse represents the response from evaluating a strategy
type EvaluateStrategyResponse struct {
	Success    bool                         `json:"success"`
	Evaluation *external.StrategyEvaluation `json:"evaluation"`
	Message    string                       `json:"message,omitempty"`
}

// RollbackStrategyRequest represents a request to rollback a strategy
type RollbackStrategyRequest struct {
	Reason string `json:"reason"`
}

// RollbackStrategyResponse represents the response from rolling back a strategy
type RollbackStrategyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// GetActiveStrategiesResponse represents the response from getting active strategies
type GetActiveStrategiesResponse struct {
	Success    bool                            `json:"success"`
	Strategies []*external.ImprovementStrategy `json:"strategies"`
	TotalCount int                             `json:"total_count"`
	Message    string                          `json:"message,omitempty"`
}

// GetImprovementHistoryResponse represents the response from getting improvement history
type GetImprovementHistoryResponse struct {
	Success    bool                           `json:"success"`
	History    []*external.ImprovementHistory `json:"history"`
	TotalCount int                            `json:"total_count"`
	Message    string                         `json:"message,omitempty"`
}

// GetContinuousImprovementConfigResponse represents the response from getting configuration
type GetContinuousImprovementConfigResponse struct {
	Success bool                                  `json:"success"`
	Config  *external.ContinuousImprovementConfig `json:"config"`
	Message string                                `json:"message,omitempty"`
}

// UpdateContinuousImprovementConfigRequest represents a request to update configuration
type UpdateContinuousImprovementConfigRequest struct {
	Config *external.ContinuousImprovementConfig `json:"config"`
}

// UpdateContinuousImprovementConfigResponse represents the response from updating configuration
type UpdateContinuousImprovementConfigResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// AnalyzeAndRecommend handles the analyze and recommend endpoint
func (h *ContinuousImprovementHandler) AnalyzeAndRecommend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AnalyzeAndRecommendRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(AnalyzeAndRecommendResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Generate recommendations
	recommendations, err := h.manager.AnalyzeAndRecommend(ctx)
	if err != nil {
		h.logger.Error("Failed to analyze and recommend", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(AnalyzeAndRecommendResponse{
			Success: false,
			Message: fmt.Sprintf("Analysis failed: %v", err),
		})
		return
	}

	response := AnalyzeAndRecommendResponse{
		Success:         true,
		Recommendations: recommendations,
		TotalCount:      len(recommendations),
		AnalysisTime:    time.Now(),
	}

	if len(recommendations) == 0 {
		response.Message = "No improvement recommendations found"
	} else {
		response.Message = fmt.Sprintf("Generated %d improvement recommendations", len(recommendations))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ApplyImprovement handles the apply improvement endpoint
func (h *ContinuousImprovementHandler) ApplyImprovement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ApplyImprovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ApplyImprovementResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate request
	if req.Recommendation == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ApplyImprovementResponse{
			Success: false,
			Message: "Recommendation is required",
		})
		return
	}

	// Apply the improvement
	strategy, err := h.manager.ApplyImprovement(ctx, req.Recommendation)
	if err != nil {
		h.logger.Error("Failed to apply improvement", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ApplyImprovementResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to apply improvement: %v", err),
		})
		return
	}

	response := ApplyImprovementResponse{
		Success:  true,
		Strategy: strategy,
		Message:  fmt.Sprintf("Successfully applied improvement strategy: %s", strategy.Name),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// EvaluateStrategy handles the evaluate strategy endpoint
func (h *ContinuousImprovementHandler) EvaluateStrategy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	strategyID := vars["strategyID"]

	if strategyID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(EvaluateStrategyResponse{
			Success: false,
			Message: "Strategy ID is required",
		})
		return
	}

	// Evaluate the strategy
	evaluation, err := h.manager.EvaluateStrategy(ctx, strategyID)
	if err != nil {
		h.logger.Error("Failed to evaluate strategy", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(EvaluateStrategyResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to evaluate strategy: %v", err),
		})
		return
	}

	response := EvaluateStrategyResponse{
		Success:    true,
		Evaluation: evaluation,
		Message:    fmt.Sprintf("Strategy evaluation completed. Improvement: %.2f%%", evaluation.Improvement*100),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RollbackStrategy handles the rollback strategy endpoint
func (h *ContinuousImprovementHandler) RollbackStrategy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	vars := mux.Vars(r)
	strategyID := vars["strategyID"]

	if strategyID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RollbackStrategyResponse{
			Success: false,
			Message: "Strategy ID is required",
		})
		return
	}

	var req RollbackStrategyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(RollbackStrategyResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Rollback the strategy
	err := h.manager.RollbackStrategy(ctx, strategyID, req.Reason)
	if err != nil {
		h.logger.Error("Failed to rollback strategy", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(RollbackStrategyResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to rollback strategy: %v", err),
		})
		return
	}

	response := RollbackStrategyResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully rolled back strategy: %s", strategyID),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetActiveStrategies handles the get active strategies endpoint
func (h *ContinuousImprovementHandler) GetActiveStrategies(w http.ResponseWriter, r *http.Request) {
	// Get active strategies
	strategies := h.manager.GetActiveStrategies()

	response := GetActiveStrategiesResponse{
		Success:    true,
		Strategies: strategies,
		TotalCount: len(strategies),
		Message:    fmt.Sprintf("Found %d active improvement strategies", len(strategies)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetImprovementHistory handles the get improvement history endpoint
func (h *ContinuousImprovementHandler) GetImprovementHistory(w http.ResponseWriter, r *http.Request) {
	// Get improvement history
	history := h.manager.GetImprovementHistory()

	response := GetImprovementHistoryResponse{
		Success:    true,
		History:    history,
		TotalCount: len(history),
		Message:    fmt.Sprintf("Found %d improvement history entries", len(history)),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetConfig handles the get configuration endpoint
func (h *ContinuousImprovementHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	// Get current configuration
	config := h.manager.GetConfig()

	response := GetContinuousImprovementConfigResponse{
		Success: true,
		Config:  config,
		Message: "Configuration retrieved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateConfig handles the update configuration endpoint
func (h *ContinuousImprovementHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateContinuousImprovementConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdateContinuousImprovementConfigResponse{
			Success: false,
			Message: "Invalid request format",
		})
		return
	}

	// Validate request
	if req.Config == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(UpdateContinuousImprovementConfigResponse{
			Success: false,
			Message: "Configuration is required",
		})
		return
	}

	// Update configuration
	err := h.manager.UpdateConfig(req.Config)
	if err != nil {
		h.logger.Error("Failed to update configuration", zap.Error(err))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(UpdateContinuousImprovementConfigResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update configuration: %v", err),
		})
		return
	}

	response := UpdateContinuousImprovementConfigResponse{
		Success: true,
		Message: "Configuration updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
