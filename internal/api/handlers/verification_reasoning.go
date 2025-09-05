package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/external"
)

// VerificationReasoningGenerateReportRequest represents a request to generate a verification reasoning report
type VerificationReasoningGenerateReportRequest struct {
	VerificationID string                 `json:"verification_id"`
	BusinessName   string                 `json:"business_name"`
	WebsiteURL     string                 `json:"website_url,omitempty"`
	Result         interface{}            `json:"result"`
	Comparison     interface{}            `json:"comparison,omitempty"`
	IncludeAudit   bool                   `json:"include_audit,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// VerificationReasoningGenerateReportResponse represents a response for generating a verification reasoning report
type VerificationReasoningGenerateReportResponse struct {
	Success   bool        `json:"success"`
	Report    interface{} `json:"report,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// VerificationReasoningHandler handles verification reasoning API requests
type VerificationReasoningHandler struct {
	generator *external.VerificationReasoningGenerator
	logger    *zap.Logger
}

// NewVerificationReasoningHandler creates a new verification reasoning handler
func NewVerificationReasoningHandler(generator *external.VerificationReasoningGenerator, logger *zap.Logger) *VerificationReasoningHandler {
	return &VerificationReasoningHandler{
		generator: generator,
		logger:    logger,
	}
}

// GenerateReasoningRequest represents a request to generate verification reasoning
type GenerateReasoningRequest struct {
	VerificationID string                       `json:"verification_id"`
	BusinessName   string                       `json:"business_name"`
	WebsiteURL     string                       `json:"website_url"`
	Result         *external.VerificationResult `json:"result"`
	Comparison     *external.ComparisonResult   `json:"comparison"`
}

// GenerateReasoningResponse represents the response from reasoning generation
type GenerateReasoningResponse struct {
	Success   bool                            `json:"success"`
	Reasoning *external.VerificationReasoning `json:"reasoning,omitempty"`
	Error     string                          `json:"error,omitempty"`
	Timestamp time.Time                       `json:"timestamp"`
}

// UpdateReasoningConfigRequest represents a request to update reasoning configuration
type UpdateReasoningConfigRequest struct {
	EnableDetailedExplanations bool    `json:"enable_detailed_explanations"`
	EnableRiskAnalysis         bool    `json:"enable_risk_analysis"`
	EnableRecommendations      bool    `json:"enable_recommendations"`
	EnableAuditTrail           bool    `json:"enable_audit_trail"`
	MinConfidenceThreshold     float64 `json:"min_confidence_threshold"`
	MaxRiskProbability         float64 `json:"max_risk_probability"`
	Language                   string  `json:"language"`
}

// UpdateReasoningConfigResponse represents the response from config update
type UpdateReasoningConfigResponse struct {
	Success   bool                                  `json:"success"`
	Config    *external.VerificationReasoningConfig `json:"config,omitempty"`
	Error     string                                `json:"error,omitempty"`
	Timestamp time.Time                             `json:"timestamp"`
}

// GenerateReasoning handles POST /generate-reasoning
func (h *VerificationReasoningHandler) GenerateReasoning(w http.ResponseWriter, r *http.Request) {
	var req GenerateReasoningRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		response := GenerateReasoningResponse{
			Success:   false,
			Error:     "Invalid request body",
			Timestamp: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := GenerateReasoningResponse{
		Timestamp: time.Now(),
	}

	// Validate request
	if req.VerificationID == "" {
		response.Success = false
		response.Error = "verification_id is required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.BusinessName == "" {
		response.Success = false
		response.Error = "business_name is required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Result == nil {
		response.Success = false
		response.Error = "result is required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate reasoning
	reasoning, err := h.generator.GenerateReasoning(
		req.VerificationID,
		req.BusinessName,
		req.WebsiteURL,
		req.Result,
		req.Comparison,
	)

	if err != nil {
		h.logger.Error("failed to generate reasoning",
			zap.String("verification_id", req.VerificationID),
			zap.Error(err))
		response.Success = false
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		response.Success = true
		response.Reasoning = reasoning
		w.WriteHeader(http.StatusOK)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GenerateReport handles POST /generate-report
func (h *VerificationReasoningHandler) GenerateReport(w http.ResponseWriter, r *http.Request) {
	var req VerificationReasoningGenerateReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		response := VerificationReasoningGenerateReportResponse{
			Success:   false,
			Error:     "Invalid request body",
			Timestamp: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := VerificationReasoningGenerateReportResponse{
		Timestamp: time.Now(),
	}

	// Validate request
	if req.VerificationID == "" {
		response.Success = false
		response.Error = "verification_id is required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.BusinessName == "" {
		response.Success = false
		response.Error = "business_name is required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Result == nil {
		response.Success = false
		response.Error = "result is required"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate comprehensive verification report
	// Type assertions for interface{} fields
	var verificationResult *external.VerificationResult
	if req.Result != nil {
		if vr, ok := req.Result.(*external.VerificationResult); ok {
			verificationResult = vr
		}
	}

	var comparisonResult *external.ComparisonResult
	if req.Comparison != nil {
		if cr, ok := req.Comparison.(*external.ComparisonResult); ok {
			comparisonResult = cr
		}
	}

	report, err := h.generator.GenerateVerificationReport(
		req.VerificationID,
		req.BusinessName,
		req.WebsiteURL,
		verificationResult,
		comparisonResult,
		req.IncludeAudit,
		req.Metadata,
	)

	if err != nil {
		h.logger.Error("failed to generate verification report",
			zap.String("verification_id", req.VerificationID),
			zap.Error(err))
		response.Success = false
		response.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		h.logger.Info("verification report generated successfully",
			zap.String("verification_id", req.VerificationID),
			zap.String("report_id", report.ReportID),
			zap.String("status", report.Status),
			zap.Float64("overall_score", report.OverallScore))

		response.Success = true
		response.Report = report
		w.WriteHeader(http.StatusOK)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GetConfig handles GET /config
func (h *VerificationReasoningHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	config := h.generator.GetConfig()

	response := UpdateReasoningConfigResponse{
		Success:   true,
		Config:    config,
		Timestamp: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// UpdateConfig handles PUT /config
func (h *VerificationReasoningHandler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	var req UpdateReasoningConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request", zap.Error(err))
		response := UpdateReasoningConfigResponse{
			Success:   false,
			Error:     "Invalid request body",
			Timestamp: time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := UpdateReasoningConfigResponse{
		Timestamp: time.Now(),
	}

	// Validate configuration
	if req.MinConfidenceThreshold < 0 || req.MinConfidenceThreshold > 1 {
		response.Success = false
		response.Error = "min_confidence_threshold must be between 0 and 1"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.MaxRiskProbability < 0 || req.MaxRiskProbability > 1 {
		response.Success = false
		response.Error = "max_risk_probability must be between 0 and 1"
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.Language == "" {
		req.Language = "en"
	}

	// Create new configuration
	newConfig := &external.VerificationReasoningConfig{
		EnableDetailedExplanations: req.EnableDetailedExplanations,
		EnableRiskAnalysis:         req.EnableRiskAnalysis,
		EnableRecommendations:      req.EnableRecommendations,
		EnableAuditTrail:           req.EnableAuditTrail,
		MinConfidenceThreshold:     req.MinConfidenceThreshold,
		MaxRiskProbability:         req.MaxRiskProbability,
		Language:                   req.Language,
	}

	// Update generator with new config
	h.generator = external.NewVerificationReasoningGenerator(newConfig)

	response.Success = true
	response.Config = newConfig

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("failed to encode response", zap.Error(err))
	}
}

// GetHealth handles GET /health
func (h *VerificationReasoningHandler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"service":   "verification_reasoning",
		"timestamp": time.Now(),
		"config":    h.generator.GetConfig() != nil,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(health); err != nil {
		h.logger.Error("failed to encode health response", zap.Error(err))
	}
}

// RegisterRoutes registers the verification reasoning routes
func (h *VerificationReasoningHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/generate-reasoning", h.GenerateReasoning).Methods("POST")
	router.HandleFunc("/generate-report", h.GenerateReport).Methods("POST")
	router.HandleFunc("/config", h.GetConfig).Methods("GET")
	router.HandleFunc("/config", h.UpdateConfig).Methods("PUT")
	router.HandleFunc("/health", h.GetHealth).Methods("GET")
}
