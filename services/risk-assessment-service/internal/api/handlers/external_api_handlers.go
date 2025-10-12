package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/external"
	"kyb-platform/services/risk-assessment-service/internal/external/ofac"
	"kyb-platform/services/risk-assessment-service/internal/external/thomson_reuters"
	"kyb-platform/services/risk-assessment-service/internal/external/worldcheck"
	"kyb-platform/services/risk-assessment-service/internal/middleware"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// ExternalAPIHandler handles external API integration requests
type ExternalAPIHandler struct {
	apiManager *external.ExternalAPIManager
	logger     *zap.Logger
}

// NewExternalAPIHandler creates a new external API handler
func NewExternalAPIHandler(apiManager *external.ExternalAPIManager, logger *zap.Logger) *ExternalAPIHandler {
	return &ExternalAPIHandler{
		apiManager: apiManager,
		logger:     logger,
	}
}

// ComprehensiveDataRequest represents a request for comprehensive external data
type ComprehensiveDataRequest struct {
	BusinessName string   `json:"business_name" validate:"required"`
	Country      string   `json:"country" validate:"required"`
	EntityType   string   `json:"entity_type,omitempty"`
	IncludeAPIs  []string `json:"include_apis,omitempty"`
}

// ComprehensiveDataResponse represents the response for comprehensive external data
type ComprehensiveDataResponse struct {
	Success        bool                                `json:"success"`
	Data           *external.PremiumExternalDataResult `json:"data,omitempty"`
	ProcessingTime string                              `json:"processing_time"`
	DataQuality    map[string]float64                  `json:"data_quality,omitempty"`
	Error          *middleware.ErrorResponse           `json:"error,omitempty"`
}

// ThomsonReutersRequest represents a request for Thomson Reuters data
type ThomsonReutersRequest struct {
	BusinessName string   `json:"business_name" validate:"required"`
	Country      string   `json:"country" validate:"required"`
	DataTypes    []string `json:"data_types,omitempty"`
}

// ThomsonReutersResponse represents the response for Thomson Reuters data
type ThomsonReutersResponse struct {
	Success        bool                                  `json:"success"`
	Data           *thomson_reuters.ThomsonReutersResult `json:"data,omitempty"`
	ProcessingTime string                                `json:"processing_time"`
	DataQuality    float64                               `json:"data_quality,omitempty"`
	Error          *middleware.ErrorResponse             `json:"error,omitempty"`
}

// OFACRequest represents a request for OFAC data
type OFACRequest struct {
	EntityName string `json:"entity_name" validate:"required"`
	EntityType string `json:"entity_type,omitempty"`
}

// OFACResponse represents the response for OFAC data
type OFACResponse struct {
	Success        bool                      `json:"success"`
	Data           *ofac.OFACResult          `json:"data,omitempty"`
	ProcessingTime string                    `json:"processing_time"`
	DataQuality    float64                   `json:"data_quality,omitempty"`
	Error          *middleware.ErrorResponse `json:"error,omitempty"`
}

// WorldCheckRequest represents a request for World-Check data
type WorldCheckRequest struct {
	EntityName string `json:"entity_name" validate:"required"`
}

// WorldCheckResponse represents the response for World-Check data
type WorldCheckResponse struct {
	Success        bool                         `json:"success"`
	Data           *worldcheck.WorldCheckResult `json:"data,omitempty"`
	ProcessingTime string                       `json:"processing_time"`
	DataQuality    float64                      `json:"data_quality,omitempty"`
	Error          *middleware.ErrorResponse    `json:"error,omitempty"`
}

// APIStatusResponse represents the response for API status
type APIStatusResponse struct {
	Success bool                      `json:"success"`
	Data    map[string]string         `json:"data,omitempty"`
	Error   *middleware.ErrorResponse `json:"error,omitempty"`
}

// SupportedAPIsResponse represents the response for supported APIs
type SupportedAPIsResponse struct {
	Success bool                      `json:"success"`
	Data    []string                  `json:"data,omitempty"`
	Error   *middleware.ErrorResponse `json:"error,omitempty"`
}

// HealthCheckResponse represents the response for health check
type HealthCheckResponse struct {
	Success bool                      `json:"success"`
	Data    map[string]bool           `json:"data,omitempty"`
	Error   *middleware.ErrorResponse `json:"error,omitempty"`
}

// GetComprehensiveData handles requests for comprehensive external data
func (h *ExternalAPIHandler) GetComprehensiveData(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req ComprehensiveDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode comprehensive data request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BusinessName == "" {
		http.Error(w, "business_name is required", http.StatusBadRequest)
		return
	}
	if req.Country == "" {
		http.Error(w, "country is required", http.StatusBadRequest)
		return
	}

	// Set default entity type if not provided
	if req.EntityType == "" {
		req.EntityType = "corporation"
	}

	// Set default APIs if not provided
	if len(req.IncludeAPIs) == 0 {
		req.IncludeAPIs = []string{"thomson_reuters", "ofac", "worldcheck"}
	}

	h.logger.Info("Getting comprehensive external data",
		zap.String("business_name", req.BusinessName),
		zap.String("country", req.Country),
		zap.Strings("include_apis", req.IncludeAPIs))

	// Get comprehensive data
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Create risk assessment request
	riskRequest := &models.RiskAssessmentRequest{
		BusinessName: req.BusinessName,
		Country:      req.Country,
		Industry:     req.EntityType,
	}

	result, err := h.apiManager.GetComprehensiveData(ctx, riskRequest)
	if err != nil {
		h.logger.Error("Failed to get comprehensive external data", zap.Error(err))
		response := ComprehensiveDataResponse{
			Success: false,
			Error: &middleware.ErrorResponse{
				Error: middleware.ErrorDetail{
					Code:    "API_ERROR",
					Message: "Failed to retrieve comprehensive external data",
					Details: err.Error(),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Calculate data quality scores
	dataQuality := make(map[string]float64)
	if result.ThomsonReuters != nil {
		dataQuality["thomson_reuters"] = 0.95 // Mock data quality score
	}
	if result.OFAC != nil {
		dataQuality["ofac"] = 0.98
	}
	if result.WorldCheck != nil {
		dataQuality["worldcheck"] = 0.92
	}

	processingTime := time.Since(startTime)
	response := ComprehensiveDataResponse{
		Success:        true,
		Data:           result,
		ProcessingTime: processingTime.String(),
		DataQuality:    dataQuality,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Comprehensive external data retrieved successfully",
		zap.String("business_name", req.BusinessName),
		zap.Duration("processing_time", processingTime))
}

// GetThomsonReutersData handles requests for Thomson Reuters data
func (h *ExternalAPIHandler) GetThomsonReutersData(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req ThomsonReutersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode Thomson Reuters request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BusinessName == "" {
		http.Error(w, "business_name is required", http.StatusBadRequest)
		return
	}
	if req.Country == "" {
		http.Error(w, "country is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting Thomson Reuters data",
		zap.String("business_name", req.BusinessName),
		zap.String("country", req.Country))

	// Get Thomson Reuters data
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Create risk assessment request
	riskRequest := &models.RiskAssessmentRequest{
		BusinessName: req.BusinessName,
		Country:      req.Country,
	}

	result, err := h.apiManager.GetThomsonReutersData(ctx, riskRequest)
	if err != nil {
		h.logger.Error("Failed to get Thomson Reuters data", zap.Error(err))
		response := ThomsonReutersResponse{
			Success: false,
			Error: &middleware.ErrorResponse{
				Error: middleware.ErrorDetail{
					Code:    "API_ERROR",
					Message: "Failed to retrieve Thomson Reuters data",
					Details: err.Error(),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	processingTime := time.Since(startTime)
	response := ThomsonReutersResponse{
		Success:        true,
		Data:           result,
		ProcessingTime: processingTime.String(),
		DataQuality:    0.95, // Mock data quality score
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Thomson Reuters data retrieved successfully",
		zap.String("business_name", req.BusinessName),
		zap.Duration("processing_time", processingTime))
}

// GetOFACData handles requests for OFAC data
func (h *ExternalAPIHandler) GetOFACData(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req OFACRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode OFAC request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.EntityName == "" {
		http.Error(w, "entity_name is required", http.StatusBadRequest)
		return
	}

	// Set default entity type if not provided
	if req.EntityType == "" {
		req.EntityType = "corporation"
	}

	h.logger.Info("Getting OFAC data",
		zap.String("entity_name", req.EntityName),
		zap.String("entity_type", req.EntityType))

	// Get OFAC data
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Create risk assessment request
	riskRequest := &models.RiskAssessmentRequest{
		BusinessName: req.EntityName,
		Country:      "US", // Default country for OFAC
	}

	result, err := h.apiManager.GetOFACData(ctx, riskRequest)
	if err != nil {
		h.logger.Error("Failed to get OFAC data", zap.Error(err))
		response := OFACResponse{
			Success: false,
			Error: &middleware.ErrorResponse{
				Error: middleware.ErrorDetail{
					Code:    "API_ERROR",
					Message: "Failed to retrieve OFAC data",
					Details: err.Error(),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	processingTime := time.Since(startTime)
	response := OFACResponse{
		Success:        true,
		Data:           result,
		ProcessingTime: processingTime.String(),
		DataQuality:    0.98, // Mock data quality score
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Info("OFAC data retrieved successfully",
		zap.String("entity_name", req.EntityName),
		zap.Duration("processing_time", processingTime))
}

// GetWorldCheckData handles requests for World-Check data
func (h *ExternalAPIHandler) GetWorldCheckData(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var req WorldCheckRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode World-Check request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.EntityName == "" {
		http.Error(w, "entity_name is required", http.StatusBadRequest)
		return
	}

	h.logger.Info("Getting World-Check data",
		zap.String("entity_name", req.EntityName))

	// Get World-Check data
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Create risk assessment request
	riskRequest := &models.RiskAssessmentRequest{
		BusinessName: req.EntityName,
		Country:      "US", // Default country for World-Check
	}

	result, err := h.apiManager.GetWorldCheckData(ctx, riskRequest)
	if err != nil {
		h.logger.Error("Failed to get World-Check data", zap.Error(err))
		response := WorldCheckResponse{
			Success: false,
			Error: &middleware.ErrorResponse{
				Error: middleware.ErrorDetail{
					Code:    "API_ERROR",
					Message: "Failed to retrieve World-Check data",
					Details: err.Error(),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	processingTime := time.Since(startTime)
	response := WorldCheckResponse{
		Success:        true,
		Data:           result,
		ProcessingTime: processingTime.String(),
		DataQuality:    0.92, // Mock data quality score
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Info("World-Check data retrieved successfully",
		zap.String("entity_name", req.EntityName),
		zap.Duration("processing_time", processingTime))
}

// GetAPIStatus handles requests for API status
func (h *ExternalAPIHandler) GetAPIStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting API status")

	status := h.apiManager.GetAPIStatus()

	response := APIStatusResponse{
		Success: true,
		Data:    status,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Info("API status retrieved successfully")
}

// GetSupportedAPIs handles requests for supported APIs
func (h *ExternalAPIHandler) GetSupportedAPIs(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting supported APIs")

	supported := h.apiManager.GetSupportedAPIs()

	response := SupportedAPIsResponse{
		Success: true,
		Data:    supported,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Supported APIs retrieved successfully")
}

// HealthCheck handles health check requests
func (h *ExternalAPIHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Performing external API health check")

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	health := h.apiManager.HealthCheck(ctx)

	response := HealthCheckResponse{
		Success: true,
		Data:    health,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Info("External API health check completed successfully",
		zap.Any("health_status", health))
}

// GetRiskFactorsFromExternalData extracts risk factors from external API data
func (h *ExternalAPIHandler) GetRiskFactorsFromExternalData(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BusinessName string   `json:"business_name" validate:"required"`
		Country      string   `json:"country" validate:"required"`
		APIs         []string `json:"apis,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode risk factors request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BusinessName == "" {
		http.Error(w, "business_name is required", http.StatusBadRequest)
		return
	}
	if req.Country == "" {
		http.Error(w, "country is required", http.StatusBadRequest)
		return
	}

	// Set default APIs if not provided
	if len(req.APIs) == 0 {
		req.APIs = []string{"thomson_reuters", "ofac", "worldcheck"}
	}

	h.logger.Info("Getting risk factors from external data",
		zap.String("business_name", req.BusinessName),
		zap.String("country", req.Country),
		zap.Strings("apis", req.APIs))

	// Get comprehensive data
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	// Create risk assessment request
	riskRequest := &models.RiskAssessmentRequest{
		BusinessName: req.BusinessName,
		Country:      req.Country,
	}

	result, err := h.apiManager.GetComprehensiveData(ctx, riskRequest)
	if err != nil {
		h.logger.Error("Failed to get external data for risk factors", zap.Error(err))
		http.Error(w, "Failed to retrieve external data", http.StatusInternalServerError)
		return
	}

	// Extract risk factors from the comprehensive result
	allRiskFactors := result.RiskFactors

	response := struct {
		Success     bool                `json:"success"`
		RiskFactors []models.RiskFactor `json:"risk_factors"`
		Count       int                 `json:"count"`
		Sources     []string            `json:"sources"`
	}{
		Success:     true,
		RiskFactors: allRiskFactors,
		Count:       len(allRiskFactors),
		Sources:     req.APIs,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Risk factors extracted from external data successfully",
		zap.String("business_name", req.BusinessName),
		zap.Int("risk_factor_count", len(allRiskFactors)))
}
