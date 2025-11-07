package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/config"
	"kyb-platform/services/risk-assessment-service/internal/engine"
	"kyb-platform/services/risk-assessment-service/internal/external"
	"kyb-platform/services/risk-assessment-service/internal/middleware"
	"kyb-platform/services/risk-assessment-service/internal/ml/service"
	"kyb-platform/services/risk-assessment-service/internal/models"
	"kyb-platform/services/risk-assessment-service/internal/supabase"
	"kyb-platform/services/risk-assessment-service/internal/validation"
)

// RiskAssessmentHandler handles risk assessment requests
type RiskAssessmentHandler struct {
	supabaseClient      *supabase.Client
	mlService           *service.MLService
	riskEngine          *engine.RiskEngine
	externalDataService *external.ExternalDataService
	logger              *zap.Logger
	config              *config.Config
	validator           *validation.Validator
	errorHandler        *middleware.ErrorHandler
}

// NewRiskAssessmentHandler creates a new risk assessment handler
func NewRiskAssessmentHandler(
	supabaseClient *supabase.Client,
	mlService *service.MLService,
	riskEngine *engine.RiskEngine,
	externalDataService *external.ExternalDataService,
	logger *zap.Logger,
	config *config.Config,
) *RiskAssessmentHandler {
	return &RiskAssessmentHandler{
		supabaseClient:      supabaseClient,
		mlService:           mlService,
		riskEngine:          riskEngine,
		externalDataService: externalDataService,
		logger:              logger,
		config:              config,
		validator:           validation.NewValidator(),
		errorHandler:        middleware.NewErrorHandler(logger),
	}
}

// HandleRiskAssessment handles POST /api/v1/assess
func (h *RiskAssessmentHandler) HandleRiskAssessment(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing risk assessment request")

	// Parse request
	var req models.RiskAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Sanitize input
	req.BusinessName = h.validator.SanitizeInput(req.BusinessName)
	req.BusinessAddress = h.validator.SanitizeInput(req.BusinessAddress)
	req.Industry = h.validator.SanitizeInput(req.Industry)
	req.Country = h.validator.SanitizeInput(req.Country)
	if req.Phone != "" {
		req.Phone = h.validator.SanitizeInput(req.Phone)
	}
	if req.Email != "" {
		req.Email = h.validator.SanitizeInput(req.Email)
	}
	if req.Website != "" {
		req.Website = h.validator.SanitizeInput(req.Website)
	}

	// Validate request using comprehensive validator
	valid, errors := h.validator.ValidateRiskAssessmentRequest(&req)
	if !valid {
		h.logger.Error("Request validation failed",
			zap.Any("errors", errors))

		// Create detailed validation error response
		errorDetail := middleware.ErrorDetail{
			Code:       "VALIDATION_ERROR",
			Message:    "Request validation failed",
			Validation: make([]middleware.ValidationError, len(errors)),
		}

		for i, err := range errors {
			errorDetail.Validation[i] = middleware.ValidationError{
				Field:   "unknown",
				Message: err,
				Code:    "VALIDATION_ERROR",
			}
		}

		errorResponse := middleware.ErrorResponse{
			Error:     errorDetail,
			RequestID: middleware.GetRequestID(r.Context()),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Path:      r.URL.Path,
			Method:    r.Method,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Log warnings if any (simplified for now)
	// Warnings functionality can be added later

	// Use high-performance risk engine for assessment
	assessment, err := h.riskEngine.AssessRisk(r.Context(), &req)
	if err != nil {
		h.logger.Error("Risk assessment failed", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("risk assessment failed: %w", err))
		return
	}

	// Create enhanced response with LSTM support
	response := &models.RiskAssessmentResponse{
		ID:                assessment.ID,
		BusinessID:        assessment.BusinessID,
		RiskScore:         assessment.RiskScore,
		RiskLevel:         assessment.RiskLevel,
		RiskFactors:       assessment.RiskFactors,
		PredictionHorizon: assessment.PredictionHorizon,
		ConfidenceScore:   assessment.ConfidenceScore,
		Status:            assessment.Status,
		CreatedAt:         assessment.CreatedAt,
		UpdatedAt:         assessment.UpdatedAt,
		Metadata:          assessment.Metadata,
	}

	// Add temporal analysis if available
	if temporalAnalysis, exists := assessment.Metadata["temporal_analysis"]; exists {
		response.Metadata["temporal_analysis"] = temporalAnalysis
	}

	// Add model information
	if modelType, exists := assessment.Metadata["model_type"]; exists {
		response.Metadata["model_type"] = modelType
	}

	// Add ensemble information if available
	if ensembleInfo, exists := assessment.Metadata["ensemble_info"]; exists {
		response.Metadata["ensemble_info"] = ensembleInfo
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Risk assessment completed",
		zap.String("assessment_id", assessment.ID),
		zap.Float64("risk_score", assessment.RiskScore),
		zap.String("risk_level", string(assessment.RiskLevel)))
}

// HandleGetRiskAssessment handles GET /api/v1/assess/{id}
func (h *RiskAssessmentHandler) HandleGetRiskAssessment(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement get risk assessment by ID
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleRiskPrediction handles POST /api/v1/assess/{id}/predict
func (h *RiskAssessmentHandler) HandleRiskPrediction(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing risk prediction request")

	// Parse request body for prediction parameters
	var predictionReq struct {
		HorizonMonths           int      `json:"horizon_months"`
		Scenarios               []string `json:"scenarios,omitempty"`
		ModelType               string   `json:"model_type,omitempty"` // "auto", "xgboost", "lstm", "ensemble"
		IncludeTemporalAnalysis bool     `json:"include_temporal_analysis,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&predictionReq); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Sanitize scenarios
	for i, scenario := range predictionReq.Scenarios {
		predictionReq.Scenarios[i] = h.validator.SanitizeInput(scenario)
	}

	// Validate prediction request
	valid, errors := h.validator.ValidatePredictionRequest(&predictionReq)
	if !valid {
		h.logger.Error("Prediction request validation failed",
			zap.Any("errors", errors))

		// Create detailed validation error response
		errorDetail := middleware.ErrorDetail{
			Code:       "VALIDATION_ERROR",
			Message:    "Prediction request validation failed",
			Validation: make([]middleware.ValidationError, len(errors)),
		}

		for i, err := range errors {
			errorDetail.Validation[i] = middleware.ValidationError{
				Field:   "unknown",
				Message: err,
				Code:    "VALIDATION_ERROR",
			}
		}

		errorResponse := middleware.ErrorResponse{
			Error:     errorDetail,
			RequestID: middleware.GetRequestID(r.Context()),
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Path:      r.URL.Path,
			Method:    r.Method,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Log warnings if any (simplified for now)
	// Warnings functionality can be added later

	// TODO: Retrieve business data from database using ID from URL
	// For now, create a mock business request
	business := &models.RiskAssessmentRequest{
		BusinessName:      "Sample Business",
		BusinessAddress:   "123 Sample St, Sample City, SC 12345",
		Industry:          "Technology",
		Country:           "US",
		PredictionHorizon: predictionReq.HorizonMonths,
		Metadata: map[string]interface{}{
			"model_type":                predictionReq.ModelType,
			"include_temporal_analysis": predictionReq.IncludeTemporalAnalysis,
		},
	}

	// Use ML service for prediction with model selection
	var prediction *models.RiskPrediction
	var err error

	if predictionReq.ModelType == "" || predictionReq.ModelType == "auto" {
		// Use ensemble routing for automatic model selection
		prediction, err = h.mlService.PredictFutureRisk(r.Context(), "auto", business, predictionReq.HorizonMonths)
	} else {
		// Use specific model
		prediction, err = h.mlService.PredictFutureRisk(r.Context(), predictionReq.ModelType, business, predictionReq.HorizonMonths)
	}

	if err != nil {
		h.logger.Error("Risk prediction failed", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("risk prediction failed: %w", err))
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(prediction)

	h.logger.Info("Risk prediction completed",
		zap.String("business_id", prediction.BusinessID),
		zap.Int("horizon_months", prediction.HorizonMonths),
		zap.Float64("predicted_score", prediction.PredictedScore),
		zap.String("predicted_level", string(prediction.PredictedLevel)))
}

// HandleRiskHistory handles GET /api/v1/assess/{id}/history
func (h *RiskAssessmentHandler) HandleRiskHistory(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement risk history
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleRiskBenchmarks handles GET /api/v1/risk/benchmarks
// Query parameters: mcc, naics, sic (at least one required)
func (h *RiskAssessmentHandler) HandleRiskBenchmarks(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing risk benchmarks request")

	// Parse query parameters
	mcc := r.URL.Query().Get("mcc")
	naics := r.URL.Query().Get("naics")
	sic := r.URL.Query().Get("sic")

	// At least one industry code is required
	if mcc == "" && naics == "" && sic == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("at least one industry code (mcc, naics, or sic) is required"))
		return
	}

	// Determine industry identifier (prefer MCC, then NAICS, then SIC)
	industryCode := mcc
	industryType := "mcc"
	if industryCode == "" {
		industryCode = naics
		industryType = "naics"
	}
	if industryCode == "" {
		industryCode = sic
		industryType = "sic"
	}

	h.logger.Info("Getting industry benchmarks",
		zap.String("industry_code", industryCode),
		zap.String("industry_type", industryType))

	// Get industry benchmarks from external data service or database
	// For now, return mock benchmarks - can be enhanced with real data
	benchmarks := map[string]interface{}{
		"industry": industryCode,
		"benchmarks": map[string]float64{
			"average_score": 70.0,
			"median_score":  72.0,
			"percentile_75": 80.0,
			"percentile_90": 85.0,
		},
		"last_updated": time.Now().Format(time.RFC3339),
	}

	// Create response
	response := map[string]interface{}{
		"industry_code": industryCode,
		"industry_type": industryType,
		"mcc":          mcc,
		"naics":        naics,
		"sic":          sic,
		"benchmarks":   benchmarks,
		"timestamp":    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Risk benchmarks request completed",
		zap.String("industry_code", industryCode),
		zap.String("industry_type", industryType))
}

// HandleRiskPredictions handles GET /api/v1/risk/predictions/{merchant_id}
// Query parameters: horizons (comma-separated: 3,6,12), includeScenarios (bool), includeConfidence (bool)
func (h *RiskAssessmentHandler) HandleRiskPredictions(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing risk predictions request")

	// Extract merchant ID from URL path
	vars := mux.Vars(r)
	merchantID := vars["merchant_id"]
	if merchantID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("merchant_id is required"))
		return
	}

	// Parse query parameters
	horizonsStr := r.URL.Query().Get("horizons")
	includeScenarios := r.URL.Query().Get("includeScenarios") == "true"
	includeConfidence := r.URL.Query().Get("includeConfidence") == "true"

	// Parse horizons (default: 3, 6, 12 months)
	horizons := []int{3, 6, 12}
	if horizonsStr != "" {
		horizonParts := strings.Split(horizonsStr, ",")
		horizons = []int{}
		for _, part := range horizonParts {
			months := strings.TrimSpace(part)
			if months != "" {
				if m, err := strconv.Atoi(months); err == nil && m > 0 {
					horizons = append(horizons, m)
				}
			}
		}
		if len(horizons) == 0 {
			horizons = []int{3, 6, 12}
		}
	}

	h.logger.Info("Getting risk predictions",
		zap.String("merchant_id", merchantID),
		zap.Ints("horizons", horizons))

	// TODO: Get risk history from database for the merchant
	// For now, generate predictions based on current assessment or mock data
	predictions := []map[string]interface{}{}
	
	for _, months := range horizons {
		// Use ML service to generate prediction
		// Create a mock business request - in production, fetch from database
		business := &models.RiskAssessmentRequest{
			BusinessName:    "Merchant " + merchantID,
			BusinessAddress: "Unknown",
			Industry:        "General",
			Country:         "US",
		}

		// Get prediction from ML service
		prediction, err := h.mlService.PredictFutureRisk(r.Context(), "auto", business, months)
		if err != nil {
			h.logger.Warn("Failed to get ML prediction, using fallback",
				zap.Error(err),
				zap.String("merchant_id", merchantID),
				zap.Int("horizon_months", months))
			
			// Fallback: generate simple prediction
			prediction = &models.RiskPrediction{
				BusinessID:      merchantID,
				PredictionDate:  time.Now(),
				HorizonMonths:   months,
				PredictedScore:  70.0,
				PredictedLevel:  models.RiskLevelMedium,
				ConfidenceScore: 0.75,
				CreatedAt:       time.Now(),
			}
		}

		predictionData := map[string]interface{}{
			"horizon_months":  months,
			"predicted_score": prediction.PredictedScore,
			"trend":           "STABLE", // Can be enhanced with trend analysis
		}

		if includeConfidence {
			predictionData["confidence"] = prediction.ConfidenceScore
		}

		if includeScenarios {
			// Add scenario analysis
			predictionData["scenarios"] = map[string]interface{}{
				"optimistic":  prediction.PredictedScore - 5,
				"realistic":   prediction.PredictedScore,
				"pessimistic": prediction.PredictedScore + 5,
			}
		}

		predictions = append(predictions, predictionData)
	}

	// Create response
	response := map[string]interface{}{
		"merchant_id":  merchantID,
		"predictions":  predictions,
		"generated_at": time.Now().Format(time.RFC3339),
		"data_points":  0, // TODO: Get actual count from database
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Risk predictions request completed",
		zap.String("merchant_id", merchantID),
		zap.Int("num_predictions", len(predictions)))
}

// HandleComplianceCheck handles POST /api/v1/compliance/check
func (h *RiskAssessmentHandler) HandleComplianceCheck(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement compliance check
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleSanctionsScreening handles POST /api/v1/sanctions/screen
func (h *RiskAssessmentHandler) HandleSanctionsScreening(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement sanctions screening
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleAdverseMediaMonitoring handles POST /api/v1/media/monitor
func (h *RiskAssessmentHandler) HandleAdverseMediaMonitoring(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement adverse media monitoring
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleRiskTrends handles GET /api/v1/analytics/trends
func (h *RiskAssessmentHandler) HandleRiskTrends(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement risk trends analytics
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleRiskInsights handles GET /api/v1/analytics/insights
func (h *RiskAssessmentHandler) HandleRiskInsights(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement risk insights analytics
	http.Error(w, "Not implemented", http.StatusNotImplemented)
}

// HandleBatchRiskAssessment handles POST /api/v1/assess/batch
func (h *RiskAssessmentHandler) HandleBatchRiskAssessment(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing batch risk assessment request")

	// Parse request
	var req struct {
		Requests []models.RiskAssessmentRequest `json:"requests"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate batch size
	if len(req.Requests) == 0 {
		h.errorHandler.HandleError(w, r, fmt.Errorf("no requests provided"))
		return
	}

	// For batches larger than 100 requests, redirect to async batch processing
	if len(req.Requests) > 100 {
		h.logger.Info("Large batch detected, redirecting to async processing",
			zap.Int("batch_size", len(req.Requests)))

		// Convert to async batch job request format
		asyncRequests := make([]map[string]interface{}, len(req.Requests))
		for i, request := range req.Requests {
			asyncRequests[i] = map[string]interface{}{
				"business_name":      request.BusinessName,
				"business_address":   request.BusinessAddress,
				"industry":           request.Industry,
				"country":            request.Country,
				"phone":              request.Phone,
				"email":              request.Email,
				"website":            request.Website,
				"prediction_horizon": request.PredictionHorizon,
				"model_type":         request.ModelType,
				"custom_model_id":    request.CustomModelID,
				"metadata":           request.Metadata,
			}
		}

		// Note: In a real implementation, you would submit this to the async batch processor
		// asyncBatchRequest := map[string]interface{}{
		//	"job_type":    "risk_assessment",
		//	"requests":    asyncRequests,
		//	"priority":    5,
		//	"max_retries": 3,
		//	"created_by":  "batch_handler",
		//	"metadata": map[string]interface{}{
		//		"source": "legacy_batch_handler",
		//		"original_batch_size": len(req.Requests),
		//	},
		// }

		// Return async job response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		response := map[string]interface{}{
			"message":         "Large batch redirected to async processing",
			"batch_size":      len(req.Requests),
			"async_endpoint":  "/api/v1/assess/batch/async",
			"status_endpoint": "/api/v1/assess/batch/{job_id}",
			"recommendation":  "Use async batch processing for batches larger than 100 requests",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate each request
	for i, request := range req.Requests {
		// Sanitize input
		request.BusinessName = h.validator.SanitizeInput(request.BusinessName)
		request.BusinessAddress = h.validator.SanitizeInput(request.BusinessAddress)
		request.Industry = h.validator.SanitizeInput(request.Industry)
		request.Country = h.validator.SanitizeInput(request.Country)
		if request.Phone != "" {
			request.Phone = h.validator.SanitizeInput(request.Phone)
		}
		if request.Email != "" {
			request.Email = h.validator.SanitizeInput(request.Email)
		}
		if request.Website != "" {
			request.Website = h.validator.SanitizeInput(request.Website)
		}

		// Validate request
		valid, errors := h.validator.ValidateRiskAssessmentRequest(&request)
		if !valid {
			h.logger.Error("Batch request validation failed",
				zap.Int("index", i),
				zap.Any("errors", errors))

			// Create detailed validation error response
			errorDetail := middleware.ErrorDetail{
				Code:       "VALIDATION_ERROR",
				Message:    fmt.Sprintf("Request at index %d validation failed", i),
				Validation: make([]middleware.ValidationError, len(errors)),
			}

			for j, err := range errors {
				errorDetail.Validation[j] = middleware.ValidationError{
					Field:   "unknown",
					Message: err,
					Code:    "VALIDATION_ERROR",
				}
			}

			errorResponse := middleware.ErrorResponse{
				Error:     errorDetail,
				RequestID: middleware.GetRequestID(r.Context()),
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Path:      r.URL.Path,
				Method:    r.Method,
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(errorResponse)
			return
		}

		req.Requests[i] = request
	}

	// Convert to pointers for batch assessment
	requestPointers := make([]*models.RiskAssessmentRequest, len(req.Requests))
	for i := range req.Requests {
		requestPointers[i] = &req.Requests[i]
	}

	// Use high-performance risk engine for batch assessment
	assessments, err := h.riskEngine.AssessRiskBatch(r.Context(), requestPointers)
	if err != nil {
		h.logger.Error("Batch risk assessment failed", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("batch risk assessment failed: %w", err))
		return
	}

	// Create response
	response := struct {
		Assessments []models.RiskAssessmentResponse `json:"assessments"`
		Count       int                             `json:"count"`
		ProcessedAt time.Time                       `json:"processed_at"`
	}{
		Assessments: make([]models.RiskAssessmentResponse, len(assessments)),
		Count:       len(assessments),
		ProcessedAt: time.Now(),
	}

	// Convert assessments to response format
	for i, assessment := range assessments {
		if assessment != nil {
			response.Assessments[i] = models.RiskAssessmentResponse{
				ID:                assessment.ID,
				BusinessID:        assessment.BusinessID,
				RiskScore:         assessment.RiskScore,
				RiskLevel:         assessment.RiskLevel,
				RiskFactors:       assessment.RiskFactors,
				PredictionHorizon: assessment.PredictionHorizon,
				ConfidenceScore:   assessment.ConfidenceScore,
				Status:            assessment.Status,
				CreatedAt:         assessment.CreatedAt,
				UpdatedAt:         assessment.UpdatedAt,
				Metadata:          assessment.Metadata,
			}
		}
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Batch risk assessment completed",
		zap.Int("count", len(assessments)),
		zap.Int("requested", len(req.Requests)))
}

// HandleExternalAdverseMediaMonitoring handles POST /api/v1/external/adverse-media
func (h *RiskAssessmentHandler) HandleExternalAdverseMediaMonitoring(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing adverse media monitoring request")

	// Parse request
	var req struct {
		BusinessName string `json:"business_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate request
	if req.BusinessName == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("business_name is required"))
		return
	}

	// Sanitize input
	req.BusinessName = h.validator.SanitizeInput(req.BusinessName)

	// Get adverse media data
	adverseMedia, err := h.externalDataService.GetAdverseMedia(r.Context(), req.BusinessName)
	if err != nil {
		h.logger.Error("Failed to get adverse media data", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("adverse media monitoring failed: %w", err))
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adverseMedia)

	h.logger.Info("Adverse media monitoring completed",
		zap.String("business_name", req.BusinessName),
		zap.Int("total_articles", adverseMedia.TotalArticles),
		zap.Float64("risk_score", adverseMedia.RiskScore))
}

// HandleCompanyDataLookup handles POST /api/v1/external/company-data
func (h *RiskAssessmentHandler) HandleCompanyDataLookup(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing company data lookup request")

	// Parse request
	var req struct {
		BusinessName string `json:"business_name"`
		Country      string `json:"country"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate request
	if req.BusinessName == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("business_name is required"))
		return
	}

	// Sanitize input
	req.BusinessName = h.validator.SanitizeInput(req.BusinessName)
	req.Country = h.validator.SanitizeInput(req.Country)

	// Get company data
	companyData, err := h.externalDataService.GetCompanyData(r.Context(), req.BusinessName, req.Country)
	if err != nil {
		h.logger.Error("Failed to get company data", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("company data lookup failed: %w", err))
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(companyData)

	h.logger.Info("Company data lookup completed",
		zap.String("business_name", req.BusinessName),
		zap.String("country", req.Country),
		zap.Int("companies_found", companyData.TotalResults),
		zap.Float64("risk_score", companyData.RiskScore))
}

// HandleExternalComplianceCheck handles POST /api/v1/external/compliance
func (h *RiskAssessmentHandler) HandleExternalComplianceCheck(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing compliance check request")

	// Parse request
	var req struct {
		BusinessName string `json:"business_name"`
		Country      string `json:"country"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Validate request
	if req.BusinessName == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("business_name is required"))
		return
	}

	// Sanitize input
	req.BusinessName = h.validator.SanitizeInput(req.BusinessName)
	req.Country = h.validator.SanitizeInput(req.Country)

	// Get compliance data
	complianceData, err := h.externalDataService.GetComplianceData(r.Context(), req.BusinessName, req.Country)
	if err != nil {
		h.logger.Error("Failed to get compliance data", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("compliance check failed: %w", err))
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(complianceData)

	h.logger.Info("Compliance check completed",
		zap.String("business_name", req.BusinessName),
		zap.String("country", req.Country),
		zap.Int("total_records", complianceData.TotalRecords),
		zap.Float64("risk_score", complianceData.RiskScore),
		zap.String("compliance_status", complianceData.ComplianceStatus))
}

// HandleExternalDataSources handles GET /api/v1/external/sources
func (h *RiskAssessmentHandler) HandleExternalDataSources(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing external data sources request")

	// Get available sources
	sources := h.externalDataService.GetAvailableSources()

	// Create response
	response := struct {
		AvailableSources []string  `json:"available_sources"`
		LastChecked      time.Time `json:"last_checked"`
	}{
		AvailableSources: sources,
		LastChecked:      time.Now(),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("External data sources request completed",
		zap.Strings("sources", sources))
}

// generateID generates a unique ID for risk assessments
func (h *RiskAssessmentHandler) generateID() string {
	return fmt.Sprintf("risk_%d", time.Now().UnixNano())
}
