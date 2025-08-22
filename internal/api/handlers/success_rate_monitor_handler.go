package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/modules/success_monitoring"
)

// SuccessRateMonitorHandler handles HTTP requests for success rate monitoring
type SuccessRateMonitorHandler struct {
	monitor *success_monitoring.SuccessRateMonitor
	logger  *zap.Logger
}

// BusinessProcessingAttemptRequest represents a request to record a business processing attempt
type BusinessProcessingAttemptRequest struct {
	ProcessName     string                 `json:"process_name"`
	InputType       string                 `json:"input_type"`
	Success         bool                   `json:"success"`
	ResponseTime    time.Duration          `json:"response_time"`
	StatusCode      int                    `json:"status_code"`
	ErrorType       string                 `json:"error_type,omitempty"`
	ErrorMessage    string                 `json:"error_message,omitempty"`
	ProcessingStage string                 `json:"processing_stage,omitempty"`
	InputSize       int                    `json:"input_size,omitempty"`
	OutputSize      int                    `json:"output_size,omitempty"`
	ConfidenceScore float64                `json:"confidence_score,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// BusinessProcessingAttemptResponse represents the response for recording an attempt
type BusinessProcessingAttemptResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message"`
	CurrentRate   float64 `json:"current_success_rate"`
	TotalAttempts int64   `json:"total_attempts"`
}

// BusinessProcessMetricsResponse represents the response for getting metrics
type BusinessProcessMetricsResponse struct {
	Success    bool                               `json:"success"`
	Metrics    *success_monitoring.ProcessMetrics `json:"metrics"`
	TargetRate float64                            `json:"target_success_rate"`
	IsAchieved bool                               `json:"target_achieved"`
	Message    string                             `json:"message,omitempty"`
}

// GetAllMetricsResponse represents the response for getting all process metrics
type GetAllMetricsResponse struct {
	Success        bool                                          `json:"success"`
	Metrics        map[string]*success_monitoring.ProcessMetrics `json:"metrics"`
	OverallMetrics *success_monitoring.OverallMetrics            `json:"overall_metrics"`
	TargetRate     float64                                       `json:"target_success_rate"`
	IsAchieved     bool                                          `json:"target_achieved"`
}

// BusinessFailureAnalysisResponse represents the response for failure analysis
type BusinessFailureAnalysisResponse struct {
	Success  bool                                `json:"success"`
	Analysis *success_monitoring.FailureAnalysis `json:"analysis"`
	Message  string                              `json:"message,omitempty"`
}

// BusinessTrendAnalysisResponse represents the response for trend analysis
type BusinessTrendAnalysisResponse struct {
	Success  bool                              `json:"success"`
	Analysis *success_monitoring.TrendAnalysis `json:"analysis"`
	Message  string                            `json:"message,omitempty"`
}

// GetAlertsResponse represents the response for getting alerts
type GetAlertsResponse struct {
	Success bool                              `json:"success"`
	Alerts  []success_monitoring.SuccessAlert `json:"alerts"`
	Count   int                               `json:"count"`
}

// ResolveAlertRequest represents a request to resolve an alert
type ResolveAlertRequest struct {
	AlertID string `json:"alert_id"`
}

// ResolveAlertResponse represents the response for resolving an alert
type ResolveAlertResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetReportResponse represents the response for getting success rate report
type GetReportResponse struct {
	Success bool                                  `json:"success"`
	Report  *success_monitoring.SuccessRateReport `json:"report"`
}

// NewSuccessRateMonitorHandler creates a new success rate monitor handler
func NewSuccessRateMonitorHandler(monitor *success_monitoring.SuccessRateMonitor, logger *zap.Logger) *SuccessRateMonitorHandler {
	return &SuccessRateMonitorHandler{
		monitor: monitor,
		logger:  logger,
	}
}

// RecordAttempt handles POST requests to record a processing attempt
func (h *SuccessRateMonitorHandler) RecordAttempt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BusinessProcessingAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ProcessName == "" {
		http.Error(w, "process_name is required", http.StatusBadRequest)
		return
	}

	if req.ResponseTime == 0 {
		http.Error(w, "response_time is required", http.StatusBadRequest)
		return
	}

	// Create data point
	dataPoint := success_monitoring.ProcessingDataPoint{
		Timestamp:       time.Now(),
		ProcessName:     req.ProcessName,
		InputType:       req.InputType,
		Success:         req.Success,
		ResponseTime:    req.ResponseTime,
		StatusCode:      req.StatusCode,
		ErrorType:       req.ErrorType,
		ErrorMessage:    req.ErrorMessage,
		ProcessingStage: req.ProcessingStage,
		InputSize:       req.InputSize,
		OutputSize:      req.OutputSize,
		ConfidenceScore: req.ConfidenceScore,
		Metadata:        req.Metadata,
	}

	// Record the attempt
	if err := h.monitor.RecordProcessingAttempt(r.Context(), dataPoint); err != nil {
		h.logger.Error("Failed to record processing attempt", zap.Error(err))
		http.Error(w, "Failed to record attempt", http.StatusInternalServerError)
		return
	}

	// Get current metrics for response
	metrics := h.monitor.GetProcessMetrics(req.ProcessName)
	if metrics == nil {
		http.Error(w, "Failed to retrieve metrics", http.StatusInternalServerError)
		return
	}

	response := BusinessProcessingAttemptResponse{
		Success:       true,
		Message:       "Processing attempt recorded successfully",
		CurrentRate:   metrics.SuccessRate,
		TotalAttempts: metrics.TotalAttempts,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Processing attempt recorded",
		zap.String("process", req.ProcessName),
		zap.String("input_type", req.InputType),
		zap.Bool("success", req.Success),
		zap.Duration("response_time", req.ResponseTime),
		zap.Float64("current_success_rate", metrics.SuccessRate))
}

// GetProcessMetrics handles GET requests to retrieve metrics for a specific process
func (h *SuccessRateMonitorHandler) GetProcessMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract process name from query parameters
	processName := r.URL.Query().Get("process_name")
	if processName == "" {
		http.Error(w, "process_name query parameter is required", http.StatusBadRequest)
		return
	}

	// Get metrics
	metrics := h.monitor.GetProcessMetrics(processName)
	if metrics == nil {
		response := BusinessProcessMetricsResponse{
			Success:    false,
			Message:    fmt.Sprintf("Process %s not found", processName),
			TargetRate: 0.95, // Default target
			IsAchieved: false,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Check if target is achieved
	targetRate := 0.95 // Default target
	isAchieved := metrics.SuccessRate >= targetRate

	response := BusinessProcessMetricsResponse{
		Success:    true,
		Metrics:    metrics,
		TargetRate: targetRate,
		IsAchieved: isAchieved,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAllProcessMetrics handles GET requests to retrieve metrics for all processes
func (h *SuccessRateMonitorHandler) GetAllProcessMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all metrics
	allMetrics := h.monitor.GetAllProcessMetrics()

	// Calculate overall metrics
	var totalAttempts, totalSuccesses, totalFailures int64
	var totalResponseTime time.Duration

	for _, metrics := range allMetrics {
		totalAttempts += metrics.TotalAttempts
		totalSuccesses += metrics.SuccessfulAttempts
		totalFailures += metrics.FailedAttempts
		totalResponseTime += metrics.AverageResponseTime * time.Duration(metrics.TotalAttempts)
	}

	overallMetrics := &success_monitoring.OverallMetrics{
		TotalAttempts:      totalAttempts,
		SuccessfulAttempts: totalSuccesses,
		FailedAttempts:     totalFailures,
	}

	if totalAttempts > 0 {
		overallMetrics.SuccessRate = float64(totalSuccesses) / float64(totalAttempts)
		overallMetrics.AverageResponseTime = totalResponseTime / time.Duration(totalAttempts)
	}

	// Check if target is achieved
	targetRate := 0.95 // Default target
	isAchieved := overallMetrics.SuccessRate >= targetRate

	response := GetAllMetricsResponse{
		Success:        true,
		Metrics:        allMetrics,
		OverallMetrics: overallMetrics,
		TargetRate:     targetRate,
		IsAchieved:     isAchieved,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetFailureAnalysis handles GET requests to perform failure analysis
func (h *SuccessRateMonitorHandler) GetFailureAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract process name from query parameters
	processName := r.URL.Query().Get("process_name")
	if processName == "" {
		http.Error(w, "process_name query parameter is required", http.StatusBadRequest)
		return
	}

	// Perform failure analysis
	analysis, err := h.monitor.AnalyzeFailures(r.Context(), processName)
	if err != nil {
		response := BusinessFailureAnalysisResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to analyze failures: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := BusinessFailureAnalysisResponse{
		Success:  true,
		Analysis: analysis,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTrendAnalysis handles GET requests to perform trend analysis
func (h *SuccessRateMonitorHandler) GetTrendAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract process name from query parameters
	processName := r.URL.Query().Get("process_name")
	if processName == "" {
		http.Error(w, "process_name query parameter is required", http.StatusBadRequest)
		return
	}

	// Perform trend analysis
	analysis, err := h.monitor.AnalyzeTrends(r.Context(), processName)
	if err != nil {
		response := BusinessTrendAnalysisResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to analyze trends: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := BusinessTrendAnalysisResponse{
		Success:  true,
		Analysis: analysis,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetAlerts handles GET requests to retrieve current alerts
func (h *SuccessRateMonitorHandler) GetAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get alerts
	alerts := h.monitor.GetAlerts()

	response := GetAlertsResponse{
		Success: true,
		Alerts:  alerts,
		Count:   len(alerts),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ResolveAlert handles POST requests to resolve an alert
func (h *SuccessRateMonitorHandler) ResolveAlert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ResolveAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AlertID == "" {
		http.Error(w, "alert_id is required", http.StatusBadRequest)
		return
	}

	// Resolve the alert
	if err := h.monitor.ResolveAlert(req.AlertID); err != nil {
		response := ResolveAlertResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to resolve alert: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := ResolveAlertResponse{
		Success: true,
		Message: "Alert resolved successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Alert resolved", zap.String("alert_id", req.AlertID))
}

// GetSuccessRateReport handles GET requests to generate a comprehensive success rate report
func (h *SuccessRateMonitorHandler) GetSuccessRateReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate report
	report, err := h.monitor.GetSuccessRateReport(r.Context())
	if err != nil {
		h.logger.Error("Failed to generate success rate report", zap.Error(err))
		http.Error(w, "Failed to generate report", http.StatusInternalServerError)
		return
	}

	response := GetReportResponse{
		Success: true,
		Report:  report,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetMetrics handles GET requests to retrieve metrics (alias for GetProcessMetrics for backward compatibility)
func (h *SuccessRateMonitorHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	h.GetProcessMetrics(w, r)
}
