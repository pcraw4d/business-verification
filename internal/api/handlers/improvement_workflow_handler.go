package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/modules/classification_optimization"
)

// ImprovementWorkflowHandler handles HTTP requests for improvement workflow operations
type ImprovementWorkflowHandler struct {
	workflow *classification_optimization.ImprovementWorkflow
	logger   *zap.Logger
}

// NewImprovementWorkflowHandler creates a new improvement workflow handler
func NewImprovementWorkflowHandler(workflow *classification_optimization.ImprovementWorkflow, logger *zap.Logger) *ImprovementWorkflowHandler {
	return &ImprovementWorkflowHandler{
		workflow: workflow,
		logger:   logger,
	}
}

// StartContinuousImprovementRequest represents a request to start continuous improvement
type StartContinuousImprovementRequest struct {
	AlgorithmID string `json:"algorithm_id"`
}

// StartABTestingRequest represents a request to start A/B testing
type StartABTestingRequest struct {
	AlgorithmA string                                  `json:"algorithm_a"`
	AlgorithmB string                                  `json:"algorithm_b"`
	TestCases  []*classification_optimization.TestCase `json:"test_cases"`
}

// WorkflowConfigRequest represents a request to update workflow configuration
type WorkflowConfigRequest struct {
	AutoImprovementEnabled bool    `json:"auto_improvement_enabled"`
	AccuracyThreshold      float64 `json:"accuracy_threshold"`
	ConfidenceThreshold    float64 `json:"confidence_threshold"`
	MaxIterations          int     `json:"max_iterations"`
	EnableABTesting        bool    `json:"enable_ab_testing"`
	TestSplitRatio         float64 `json:"test_split_ratio"`
}

// StartContinuousImprovement starts a continuous improvement workflow
func (h *ImprovementWorkflowHandler) StartContinuousImprovement(w http.ResponseWriter, r *http.Request) {
	var req StartContinuousImprovementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AlgorithmID == "" {
		http.Error(w, "algorithm_id is required", http.StatusBadRequest)
		return
	}

	execution, err := h.workflow.StartContinuousImprovement(r.Context(), req.AlgorithmID)
	if err != nil {
		h.logger.Error("Failed to start continuous improvement",
			zap.String("algorithm_id", req.AlgorithmID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to start workflow: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(execution)
}

// StartABTesting starts an A/B testing workflow
func (h *ImprovementWorkflowHandler) StartABTesting(w http.ResponseWriter, r *http.Request) {
	var req StartABTestingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AlgorithmA == "" || req.AlgorithmB == "" {
		http.Error(w, "algorithm_a and algorithm_b are required", http.StatusBadRequest)
		return
	}

	if len(req.TestCases) == 0 {
		http.Error(w, "test_cases are required", http.StatusBadRequest)
		return
	}

	execution, err := h.workflow.StartABTesting(r.Context(), req.AlgorithmA, req.AlgorithmB, req.TestCases)
	if err != nil {
		h.logger.Error("Failed to start A/B testing",
			zap.String("algorithm_a", req.AlgorithmA),
			zap.String("algorithm_b", req.AlgorithmB),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to start A/B testing: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(execution)
}

// GetWorkflowHistory returns the history of workflow executions
func (h *ImprovementWorkflowHandler) GetWorkflowHistory(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	limit := 100 // Default limit
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	history := h.workflow.GetWorkflowHistory()

	// Apply limit
	if len(history) > limit {
		history = history[:limit]
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workflows": history,
		"total":     len(history),
		"limit":     limit,
	})
}

// GetActiveWorkflows returns currently active workflows
func (h *ImprovementWorkflowHandler) GetActiveWorkflows(w http.ResponseWriter, r *http.Request) {
	active := h.workflow.GetActiveWorkflows()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"active_workflows": active,
		"count":            len(active),
	})
}

// GetWorkflowExecution returns a specific workflow execution
func (h *ImprovementWorkflowHandler) GetWorkflowExecution(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["workflow_id"]

	if workflowID == "" {
		http.Error(w, "workflow_id is required", http.StatusBadRequest)
		return
	}

	// Get from history first
	history := h.workflow.GetWorkflowHistory()
	for _, execution := range history {
		if execution.ID == workflowID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(execution)
			return
		}
	}

	// Check active workflows
	active := h.workflow.GetActiveWorkflows()
	for _, execution := range active {
		if execution.ID == workflowID {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(execution)
			return
		}
	}

	http.Error(w, "Workflow not found", http.StatusNotFound)
}

// StopWorkflow stops an active workflow
func (h *ImprovementWorkflowHandler) StopWorkflow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workflowID := vars["workflow_id"]

	if workflowID == "" {
		http.Error(w, "workflow_id is required", http.StatusBadRequest)
		return
	}

	err := h.workflow.StopWorkflow(workflowID)
	if err != nil {
		h.logger.Error("Failed to stop workflow",
			zap.String("workflow_id", workflowID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to stop workflow: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message":     "Workflow stopped successfully",
		"workflow_id": workflowID,
	})
}

// GetWorkflowStatistics returns statistics about workflow executions
func (h *ImprovementWorkflowHandler) GetWorkflowStatistics(w http.ResponseWriter, r *http.Request) {
	history := h.workflow.GetWorkflowHistory()
	active := h.workflow.GetActiveWorkflows()

	// Calculate statistics
	var totalExecutions, completedExecutions, failedExecutions int
	var totalImprovementScore float64
	workflowTypes := make(map[string]int)
	statuses := make(map[string]int)

	for _, execution := range history {
		totalExecutions++
		workflowTypes[string(execution.Type)]++
		statuses[string(execution.Status)]++

		if execution.Status == classification_optimization.WorkflowStatusCompleted {
			completedExecutions++
			totalImprovementScore += execution.ImprovementScore
		} else if execution.Status == classification_optimization.WorkflowStatusFailed {
			failedExecutions++
		}
	}

	// Add active workflows
	for _, execution := range active {
		totalExecutions++
		workflowTypes[string(execution.Type)]++
		statuses[string(execution.Status)]++
	}

	successRate := 0.0
	if totalExecutions > 0 {
		successRate = float64(completedExecutions) / float64(totalExecutions) * 100
	}

	averageImprovement := 0.0
	if completedExecutions > 0 {
		averageImprovement = totalImprovementScore / float64(completedExecutions)
	}

	statistics := map[string]interface{}{
		"total_executions":     totalExecutions,
		"active_workflows":     len(active),
		"completed_executions": completedExecutions,
		"failed_executions":    failedExecutions,
		"success_rate":         successRate,
		"average_improvement":  averageImprovement,
		"workflow_types":       workflowTypes,
		"status_distribution":  statuses,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statistics)
}

// GetWorkflowRecommendations returns recommendations for workflow improvements
func (h *ImprovementWorkflowHandler) GetWorkflowRecommendations(w http.ResponseWriter, r *http.Request) {
	history := h.workflow.GetWorkflowHistory()

	// Analyze recent workflows for recommendations
	var recommendations []map[string]interface{}

	// Check for failed workflows
	failedCount := 0
	for _, execution := range history {
		if execution.Status == classification_optimization.WorkflowStatusFailed {
			failedCount++
		}
	}

	if failedCount > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "warning",
			"priority":    "high",
			"title":       "Failed Workflows Detected",
			"description": fmt.Sprintf("%d workflows have failed recently", failedCount),
			"action":      "Review failed workflow logs and adjust parameters",
		})
	}

	// Check for low improvement scores
	lowImprovementCount := 0
	for _, execution := range history {
		if execution.Status == classification_optimization.WorkflowStatusCompleted && execution.ImprovementScore < 0.01 {
			lowImprovementCount++
		}
	}

	if lowImprovementCount > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "info",
			"priority":    "medium",
			"title":       "Low Improvement Scores",
			"description": fmt.Sprintf("%d workflows showed minimal improvement", lowImprovementCount),
			"action":      "Consider adjusting optimization parameters or algorithm configuration",
		})
	}

	// Check for long-running workflows
	longRunningCount := 0
	for _, execution := range history {
		if execution.Status == classification_optimization.WorkflowStatusRunning {
			duration := time.Since(execution.StartTime)
			if duration > 30*time.Minute {
				longRunningCount++
			}
		}
	}

	if longRunningCount > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "warning",
			"priority":    "medium",
			"title":       "Long-Running Workflows",
			"description": fmt.Sprintf("%d workflows have been running for over 30 minutes", longRunningCount),
			"action":      "Monitor workflow performance and consider timeout adjustments",
		})
	}

	// Default recommendation if no issues found
	if len(recommendations) == 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "success",
			"priority":    "low",
			"title":       "System Operating Normally",
			"description": "All workflows are performing as expected",
			"action":      "Continue monitoring for optimal performance",
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"recommendations": recommendations,
		"total":           len(recommendations),
	})
}

// GetWorkflowMetrics returns detailed metrics for workflow analysis
func (h *ImprovementWorkflowHandler) GetWorkflowMetrics(w http.ResponseWriter, r *http.Request) {
	history := h.workflow.GetWorkflowHistory()

	// Calculate detailed metrics
	var totalDuration time.Duration
	var totalIterations int
	var accuracyImprovements []float64
	var f1ScoreImprovements []float64
	var confidenceImprovements []float64

	for _, execution := range history {
		if execution.Status == classification_optimization.WorkflowStatusCompleted && execution.EndTime != nil {
			duration := execution.EndTime.Sub(execution.StartTime)
			totalDuration += duration

			// Calculate improvements
			if execution.BaselineMetrics != nil && execution.FinalMetrics != nil {
				accuracyImprovements = append(accuracyImprovements,
					execution.FinalMetrics.Accuracy-execution.BaselineMetrics.Accuracy)
				f1ScoreImprovements = append(f1ScoreImprovements,
					execution.FinalMetrics.F1Score-execution.BaselineMetrics.F1Score)
				confidenceImprovements = append(confidenceImprovements,
					execution.FinalMetrics.AverageConfidence-execution.BaselineMetrics.AverageConfidence)
			}

			// Count iterations
			totalIterations += len(execution.Iterations)
		}
	}

	// Calculate averages
	averageDuration := 0.0
	averageIterations := 0.0
	averageAccuracyImprovement := 0.0
	averageF1Improvement := 0.0
	averageConfidenceImprovement := 0.0

	completedCount := 0
	for _, execution := range history {
		if execution.Status == classification_optimization.WorkflowStatusCompleted {
			completedCount++
		}
	}

	if completedCount > 0 {
		averageDuration = totalDuration.Seconds() / float64(completedCount)
		averageIterations = float64(totalIterations) / float64(completedCount)
	}

	if len(accuracyImprovements) > 0 {
		var sum float64
		for _, improvement := range accuracyImprovements {
			sum += improvement
		}
		averageAccuracyImprovement = sum / float64(len(accuracyImprovements))
	}

	if len(f1ScoreImprovements) > 0 {
		var sum float64
		for _, improvement := range f1ScoreImprovements {
			sum += improvement
		}
		averageF1Improvement = sum / float64(len(f1ScoreImprovements))
	}

	if len(confidenceImprovements) > 0 {
		var sum float64
		for _, improvement := range confidenceImprovements {
			sum += improvement
		}
		averageConfidenceImprovement = sum / float64(len(confidenceImprovements))
	}

	metrics := map[string]interface{}{
		"total_workflows":                len(history),
		"completed_workflows":            completedCount,
		"average_duration_seconds":       averageDuration,
		"average_iterations":             averageIterations,
		"average_accuracy_improvement":   averageAccuracyImprovement,
		"average_f1_improvement":         averageF1Improvement,
		"average_confidence_improvement": averageConfidenceImprovement,
		"total_iterations":               totalIterations,
		"total_duration_seconds":         totalDuration.Seconds(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}
