package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/classification-service/internal/config"
	"kyb-platform/services/classification-service/internal/supabase"
)

// ClassificationHandler handles classification requests
type ClassificationHandler struct {
	supabaseClient *supabase.Client
	logger         *zap.Logger
	config         *config.Config
}

// NewClassificationHandler creates a new classification handler
func NewClassificationHandler(supabaseClient *supabase.Client, logger *zap.Logger, config *config.Config) *ClassificationHandler {
	return &ClassificationHandler{
		supabaseClient: supabaseClient,
		logger:         logger,
		config:         config,
	}
}

// ClassificationRequest represents a classification request
type ClassificationRequest struct {
	BusinessName string `json:"business_name"`
	Description  string `json:"description"`
	WebsiteURL   string `json:"website_url,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
}

// ClassificationResponse represents a classification response
type ClassificationResponse struct {
	RequestID       string                 `json:"request_id"`
	BusinessName    string                 `json:"business_name"`
	Description     string                 `json:"description"`
	Classification  *ClassificationResult  `json:"classification"`
	RiskAssessment  *RiskAssessmentResult  `json:"risk_assessment"`
	ConfidenceScore float64                `json:"confidence_score"`
	DataSource      string                 `json:"data_source"`
	Status          string                 `json:"status"`
	Success         bool                   `json:"success"`
	Timestamp       time.Time              `json:"timestamp"`
	ProcessingTime  time.Duration          `json:"processing_time"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ClassificationResult represents the classification results
type ClassificationResult struct {
	Industry       string          `json:"industry"`
	MCCCodes       []IndustryCode  `json:"mcc_codes"`
	NAICSCodes     []IndustryCode  `json:"naics_codes"`
	SICCodes       []IndustryCode  `json:"sic_codes"`
	WebsiteContent *WebsiteContent `json:"website_content"`
}

// IndustryCode represents an industry classification code
type IndustryCode struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// WebsiteContent represents website content analysis
type WebsiteContent struct {
	Scraped       bool `json:"scraped"`
	ContentLength int  `json:"content_length"`
	KeywordsFound int  `json:"keywords_found"`
}

// RiskAssessmentResult represents risk assessment results
type RiskAssessmentResult struct {
	RiskLevel               string            `json:"risk_level"`
	RiskScore               float64           `json:"risk_score"`
	RiskFactors             map[string]string `json:"risk_factors"`
	DetectedRisks           []string          `json:"detected_risks,omitempty"`
	ProhibitedKeywordsFound []string          `json:"prohibited_keywords_found,omitempty"`
	AssessmentMethodology   string            `json:"assessment_methodology"`
	AssessmentTimestamp     time.Time         `json:"assessment_timestamp"`
}

// HandleClassification handles classification requests
func (h *ClassificationHandler) HandleClassification(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var req ClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BusinessName == "" {
		http.Error(w, "business_name is required", http.StatusBadRequest)
		return
	}

	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = h.generateRequestID()
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.config.Classification.RequestTimeout)
	defer cancel()

	// Process classification
	response, err := h.processClassification(ctx, &req, startTime)
	if err != nil {
		h.logger.Error("Classification failed",
			zap.String("request_id", req.RequestID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Classification failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Classification completed successfully",
		zap.String("request_id", req.RequestID),
		zap.Duration("processing_time", time.Since(startTime)))
}

// processClassification processes a classification request
func (h *ClassificationHandler) processClassification(ctx context.Context, req *ClassificationRequest, startTime time.Time) (*ClassificationResponse, error) {
	// For now, implement a simplified classification logic
	// This will be enhanced with the full classification algorithms from the main service

	// Generate a simple classification result
	classification := &ClassificationResult{
		Industry: "General Business",
		MCCCodes: []IndustryCode{
			{
				Code:        "7372",
				Description: "Computer Programming Services",
				Confidence:  0.7,
			},
		},
		NAICSCodes: []IndustryCode{
			{
				Code:        "541511",
				Description: "Custom Computer Programming Services",
				Confidence:  0.7,
			},
		},
		SICCodes: []IndustryCode{
			{
				Code:        "7372",
				Description: "Computer Programming Services",
				Confidence:  0.7,
			},
		},
		WebsiteContent: &WebsiteContent{
			Scraped:       false,
			ContentLength: 0,
			KeywordsFound: 0,
		},
	}

	// Generate risk assessment
	riskAssessment := &RiskAssessmentResult{
		RiskLevel: "low",
		RiskScore: 0.0,
		RiskFactors: map[string]string{
			"geographic": "low_risk",
			"industry":   "general",
			"regulatory": "compliant",
		},
		DetectedRisks:           nil,
		ProhibitedKeywordsFound: nil,
		AssessmentMethodology:   "automated",
		AssessmentTimestamp:     time.Now(),
	}

	// Create response
	response := &ClassificationResponse{
		RequestID:       req.RequestID,
		BusinessName:    req.BusinessName,
		Description:     req.Description,
		Classification:  classification,
		RiskAssessment:  riskAssessment,
		ConfidenceScore: 0.7,
		DataSource:      "supabase_new",
		Status:          "success",
		Success:         true,
		Timestamp:       time.Now(),
		ProcessingTime:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"service": "classification-service",
			"version": "1.0.0",
		},
	}

	return response, nil
}

// generateRequestID generates a unique request ID
func (h *ClassificationHandler) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// HandleHealth handles health check requests
func (h *ClassificationHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check Supabase connectivity
	supabaseHealthy := true
	var supabaseError error
	if err := h.supabaseClient.HealthCheck(ctx); err != nil {
		supabaseHealthy = false
		supabaseError = err
	}

	// Get classification data
	classificationData, err := h.supabaseClient.GetClassificationData(ctx)
	if err != nil {
		h.logger.Warn("Failed to get classification data", zap.Error(err))
	}

	// Create health response
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
		"service":   "classification-service",
		"uptime":    time.Since(startTime).String(),
		"supabase_status": map[string]interface{}{
			"connected": supabaseHealthy,
			"url":       h.config.Supabase.URL,
			"error":     supabaseError,
		},
		"classification_data": classificationData,
		"features": map[string]interface{}{
			"ml_enabled":             h.config.Classification.MLEnabled,
			"keyword_method_enabled": h.config.Classification.KeywordMethodEnabled,
			"ensemble_enabled":       h.config.Classification.EnsembleEnabled,
			"cache_enabled":          h.config.Classification.CacheEnabled,
		},
	}

	// Set status code based on health
	statusCode := http.StatusOK
	if !supabaseHealthy {
		statusCode = http.StatusServiceUnavailable
		health["status"] = "unhealthy"
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}
