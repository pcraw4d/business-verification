package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/external"
)

// VerificationSuccessMonitorHandler handles HTTP requests for verification success monitoring
type VerificationSuccessMonitorHandler struct {
	monitor *external.VerificationSuccessMonitor
	logger  *zap.Logger
}

// RecordAttemptRequest represents a request to record a verification attempt
type RecordAttemptRequest struct {
	URL           string                 `json:"url"`
	Success       bool                   `json:"success"`
	ResponseTime  time.Duration          `json:"response_time"`
	StatusCode    int                    `json:"status_code"`
	ErrorType     string                 `json:"error_type,omitempty"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	StrategyUsed  string                 `json:"strategy_used,omitempty"`
	UserAgentUsed string                 `json:"user_agent_used,omitempty"`
	ProxyUsed     *external.Proxy        `json:"proxy_used,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// RecordAttemptResponse represents the response for recording an attempt
type RecordAttemptResponse struct {
	Success       bool    `json:"success"`
	Message       string  `json:"message"`
	CurrentRate   float64 `json:"current_success_rate"`
	TotalAttempts int64   `json:"total_attempts"`
}

// GetMetricsResponse represents the response for getting metrics
type GetMetricsResponse struct {
	Success    bool                     `json:"success"`
	Metrics    *external.SuccessMetrics `json:"metrics"`
	TargetRate float64                  `json:"target_success_rate"`
	IsAchieved bool                     `json:"target_achieved"`
}

// GetFailureAnalysisResponse represents the response for failure analysis
type GetFailureAnalysisResponse struct {
	Success  bool                      `json:"success"`
	Analysis *external.FailureAnalysis `json:"analysis"`
	Message  string                    `json:"message,omitempty"`
}

// GetTrendAnalysisResponse represents the response for trend analysis
type GetTrendAnalysisResponse struct {
	Success bool                    `json:"success"`
	Trends  *external.TrendAnalysis `json:"trends"`
	Message string                  `json:"message,omitempty"`
}

// GetSuccessMonitorConfigResponse represents the response for getting configuration
type GetSuccessMonitorConfigResponse struct {
	Success bool                           `json:"success"`
	Config  *external.SuccessMonitorConfig `json:"config"`
}

// UpdateSuccessMonitorConfigRequest represents a request to update configuration
type UpdateSuccessMonitorConfigRequest struct {
	EnableRealTimeMonitoring bool          `json:"enable_real_time_monitoring"`
	EnableFailureAnalysis    bool          `json:"enable_failure_analysis"`
	EnableTrendAnalysis      bool          `json:"enable_trend_analysis"`
	EnableAlerting           bool          `json:"enable_alerting"`
	TargetSuccessRate        float64       `json:"target_success_rate"`
	AlertThreshold           float64       `json:"alert_threshold"`
	MetricsRetentionPeriod   time.Duration `json:"metrics_retention_period"`
	AnalysisWindow           time.Duration `json:"analysis_window"`
	TrendWindow              time.Duration `json:"trend_window"`
	MinDataPoints            int           `json:"min_data_points"`
	MaxDataPoints            int           `json:"max_data_points"`
}

// UpdateSuccessMonitorConfigResponse represents the response for updating configuration
type UpdateSuccessMonitorConfigResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ResetMetricsResponse represents the response for resetting metrics
type ResetMetricsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetStatusResponse represents the response for getting overall status
type GetStatusResponse struct {
	Success       bool      `json:"success"`
	CurrentRate   float64   `json:"current_success_rate"`
	TargetRate    float64   `json:"target_success_rate"`
	IsAchieved    bool      `json:"target_achieved"`
	TotalAttempts int64     `json:"total_attempts"`
	LastUpdated   time.Time `json:"last_updated"`
}

// NewVerificationSuccessMonitorHandler creates a new success monitor handler
func NewVerificationSuccessMonitorHandler(monitor *external.VerificationSuccessMonitor, logger *zap.Logger) *VerificationSuccessMonitorHandler {
	return &VerificationSuccessMonitorHandler{
		monitor: monitor,
		logger:  logger,
	}
}

// RecordAttempt handles POST requests to record verification attempts
func (h *VerificationSuccessMonitorHandler) RecordAttempt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RecordAttemptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	if req.ResponseTime < 0 {
		http.Error(w, "Response time must be non-negative", http.StatusBadRequest)
		return
	}

	// Convert to DataPoint
	dataPoint := external.DataPoint{
		URL:           req.URL,
		Success:       req.Success,
		ResponseTime:  req.ResponseTime,
		StatusCode:    req.StatusCode,
		ErrorType:     req.ErrorType,
		ErrorMessage:  req.ErrorMessage,
		StrategyUsed:  req.StrategyUsed,
		UserAgentUsed: req.UserAgentUsed,
		ProxyUsed:     req.ProxyUsed,
		Metadata:      req.Metadata,
	}

	// Record the attempt
	ctx := r.Context()
	err := h.monitor.RecordAttempt(ctx, dataPoint)
	if err != nil {
		h.logger.Error("Failed to record attempt", zap.Error(err))
		http.Error(w, "Failed to record attempt", http.StatusInternalServerError)
		return
	}

	// Get current metrics
	metrics := h.monitor.GetMetrics()

	response := RecordAttemptResponse{
		Success:       true,
		Message:       "Attempt recorded successfully",
		CurrentRate:   metrics.SuccessRate,
		TotalAttempts: metrics.TotalAttempts,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Verification attempt recorded",
		zap.String("url", req.URL),
		zap.Bool("success", req.Success),
		zap.Duration("response_time", req.ResponseTime),
		zap.Float64("current_rate", metrics.SuccessRate))
}

// GetMetrics handles GET requests to retrieve current metrics
func (h *VerificationSuccessMonitorHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := h.monitor.GetMetrics()
	config := h.monitor.GetConfig()

	response := GetMetricsResponse{
		Success:    true,
		Metrics:    metrics,
		TargetRate: config.TargetSuccessRate,
		IsAchieved: h.monitor.IsTargetAchieved(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetFailureAnalysis handles GET requests to retrieve failure analysis
func (h *VerificationSuccessMonitorHandler) GetFailureAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	analysis, err := h.monitor.AnalyzeFailures(ctx)
	if err != nil {
		h.logger.Error("Failed to analyze failures", zap.Error(err))
		response := GetFailureAnalysisResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to analyze failures: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := GetFailureAnalysisResponse{
		Success:  true,
		Analysis: analysis,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTrendAnalysis handles GET requests to retrieve trend analysis
func (h *VerificationSuccessMonitorHandler) GetTrendAnalysis(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := r.Context()
	trends, err := h.monitor.AnalyzeTrends(ctx)
	if err != nil {
		h.logger.Error("Failed to analyze trends", zap.Error(err))
		response := GetTrendAnalysisResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to analyze trends: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := GetTrendAnalysisResponse{
		Success: true,
		Trends:  trends,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetConfig handles GET requests to retrieve current configuration
func (h *VerificationSuccessMonitorHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := h.monitor.GetConfig()

	response := GetSuccessMonitorConfigResponse{
		Success: true,
		Config:  config,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateConfig handles PUT requests to update configuration
func (h *VerificationSuccessMonitorHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req UpdateSuccessMonitorConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.TargetSuccessRate < 0 || req.TargetSuccessRate > 1 {
		http.Error(w, "Target success rate must be between 0 and 1", http.StatusBadRequest)
		return
	}

	if req.AlertThreshold < 0 || req.AlertThreshold > 1 {
		http.Error(w, "Alert threshold must be between 0 and 1", http.StatusBadRequest)
		return
	}

	if req.AlertThreshold >= req.TargetSuccessRate {
		http.Error(w, "Alert threshold must be less than target success rate", http.StatusBadRequest)
		return
	}

	if req.MinDataPoints < 1 {
		http.Error(w, "Minimum data points must be at least 1", http.StatusBadRequest)
		return
	}

	if req.MaxDataPoints < req.MinDataPoints {
		http.Error(w, "Maximum data points must be greater than or equal to minimum data points", http.StatusBadRequest)
		return
	}

	// Create new config
	newConfig := &external.SuccessMonitorConfig{
		EnableRealTimeMonitoring: req.EnableRealTimeMonitoring,
		EnableFailureAnalysis:    req.EnableFailureAnalysis,
		EnableTrendAnalysis:      req.EnableTrendAnalysis,
		EnableAlerting:           req.EnableAlerting,
		TargetSuccessRate:        req.TargetSuccessRate,
		AlertThreshold:           req.AlertThreshold,
		MetricsRetentionPeriod:   req.MetricsRetentionPeriod,
		AnalysisWindow:           req.AnalysisWindow,
		TrendWindow:              req.TrendWindow,
		MinDataPoints:            req.MinDataPoints,
		MaxDataPoints:            req.MaxDataPoints,
	}

	// Update configuration
	err := h.monitor.UpdateConfig(newConfig)
	if err != nil {
		h.logger.Error("Failed to update configuration", zap.Error(err))
		response := UpdateSuccessMonitorConfigResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to update configuration: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := UpdateSuccessMonitorConfigResponse{
		Success: true,
		Message: "Configuration updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Success monitor configuration updated",
		zap.Float64("target_success_rate", req.TargetSuccessRate),
		zap.Float64("alert_threshold", req.AlertThreshold))
}

// ResetMetrics handles POST requests to reset all metrics
func (h *VerificationSuccessMonitorHandler) ResetMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.monitor.ResetMetrics()

	response := ResetMetricsResponse{
		Success: true,
		Message: "Metrics reset successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Success monitor metrics reset")
}

// GetStatus handles GET requests to retrieve overall status
func (h *VerificationSuccessMonitorHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := h.monitor.GetMetrics()
	config := h.monitor.GetConfig()

	response := GetStatusResponse{
		Success:       true,
		CurrentRate:   metrics.SuccessRate,
		TargetRate:    config.TargetSuccessRate,
		IsAchieved:    h.monitor.IsTargetAchieved(),
		TotalAttempts: metrics.TotalAttempts,
		LastUpdated:   metrics.LastUpdated,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// RegisterRoutes registers all routes for the success monitor handler
func (h *VerificationSuccessMonitorHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/success-monitor/record", h.RecordAttempt)
	mux.HandleFunc("GET /api/v1/success-monitor/metrics", h.GetMetrics)
	mux.HandleFunc("GET /api/v1/success-monitor/failures", h.GetFailureAnalysis)
	mux.HandleFunc("GET /api/v1/success-monitor/trends", h.GetTrendAnalysis)
	mux.HandleFunc("GET /api/v1/success-monitor/config", h.GetConfig)
	mux.HandleFunc("PUT /api/v1/success-monitor/config", h.UpdateConfig)
	mux.HandleFunc("POST /api/v1/success-monitor/reset", h.ResetMetrics)
	mux.HandleFunc("GET /api/v1/success-monitor/status", h.GetStatus)
}
