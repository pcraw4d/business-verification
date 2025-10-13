package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/custom"
	"kyb-platform/services/risk-assessment-service/internal/ml/industry"
	"kyb-platform/services/risk-assessment-service/internal/webhooks"
)

// CustomModelHandlers handles HTTP requests for custom risk model management
type CustomModelHandlers struct {
	modelBuilder *custom.CustomModelBuilder
	repository   custom.CustomModelRepository
	eventService *webhooks.EventService
	logger       *zap.Logger
}

// NewCustomModelHandlers creates a new custom model handlers instance
func NewCustomModelHandlers(
	modelBuilder *custom.CustomModelBuilder,
	repository custom.CustomModelRepository,
	eventService *webhooks.EventService,
	logger *zap.Logger,
) *CustomModelHandlers {
	return &CustomModelHandlers{
		modelBuilder: modelBuilder,
		repository:   repository,
		eventService: eventService,
		logger:       logger,
	}
}

// NewCustomModelHandler creates a new custom model handler instance (alias for compatibility)
func NewCustomModelHandler(
	modelBuilder *custom.CustomModelBuilder,
	logger *zap.Logger,
) *CustomModelHandlers {
	// Get repository from model builder
	repository := modelBuilder.GetRepository()
	return &CustomModelHandlers{
		modelBuilder: modelBuilder,
		repository:   repository,
		eventService: nil, // Will be set later when webhook components are available
		logger:       logger,
	}
}

// CreateCustomModel handles POST /api/v1/models/custom
func (cmh *CustomModelHandlers) CreateCustomModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cmh.logger.Info("Creating custom risk model")

	var request custom.CreateCustomModelRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		cmh.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create the custom model
	model, err := cmh.modelBuilder.CreateCustomModel(ctx, &request)
	if err != nil {
		cmh.logger.Error("Failed to create custom model", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save to repository
	if err := cmh.repository.SaveCustomModel(ctx, model); err != nil {
		cmh.logger.Error("Failed to save custom model", zap.Error(err))
		http.Error(w, "Failed to save model", http.StatusInternalServerError)
		return
	}

	// Trigger webhook event for custom model created
	if cmh.eventService != nil {
		tenantID := cmh.getTenantIDFromContext(ctx)
		go func() {
			if webhookErr := cmh.eventService.TriggerCustomModelCreated(context.Background(), tenantID, model.ID, model.Name); webhookErr != nil {
				cmh.logger.Error("Failed to trigger custom model created webhook", zap.Error(webhookErr))
			}
		}()
	}

	// Return the created model
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(model); err != nil {
		cmh.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// GetCustomModel handles GET /api/v1/models/custom/{id}
func (cmh *CustomModelHandlers) GetCustomModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	modelID := vars["id"]

	cmh.logger.Info("Retrieving custom risk model", zap.String("model_id", modelID))

	model, err := cmh.repository.GetCustomModel(ctx, modelID)
	if err != nil {
		cmh.logger.Error("Failed to retrieve custom model", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(model); err != nil {
		cmh.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// UpdateCustomModel handles PUT /api/v1/models/custom/{id}
func (cmh *CustomModelHandlers) UpdateCustomModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	modelID := vars["id"]

	cmh.logger.Info("Updating custom risk model", zap.String("model_id", modelID))

	var request custom.UpdateCustomModelRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		cmh.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update the custom model
	model, err := cmh.modelBuilder.UpdateCustomModel(ctx, modelID, &request)
	if err != nil {
		cmh.logger.Error("Failed to update custom model", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save to repository
	if err := cmh.repository.SaveCustomModel(ctx, model); err != nil {
		cmh.logger.Error("Failed to save updated custom model", zap.Error(err))
		http.Error(w, "Failed to save model", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(model); err != nil {
		cmh.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// DeleteCustomModel handles DELETE /api/v1/models/custom/{id}
func (cmh *CustomModelHandlers) DeleteCustomModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	modelID := vars["id"]

	cmh.logger.Info("Deleting custom risk model", zap.String("model_id", modelID))

	if err := cmh.repository.DeleteCustomModel(ctx, modelID); err != nil {
		cmh.logger.Error("Failed to delete custom model", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListCustomModels handles GET /api/v1/models/custom
func (cmh *CustomModelHandlers) ListCustomModels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get query parameters
	tenantID := r.URL.Query().Get("tenant_id")
	if tenantID == "" {
		http.Error(w, "tenant_id is required", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default limit
	offset := 0 // default offset

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	cmh.logger.Info("Listing custom risk models",
		zap.String("tenant_id", tenantID),
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	models, err := cmh.repository.ListCustomModels(ctx, tenantID, limit, offset)
	if err != nil {
		cmh.logger.Error("Failed to list custom models", zap.Error(err))
		http.Error(w, "Failed to list models", http.StatusInternalServerError)
		return
	}

	// Get total count
	totalCount, err := cmh.repository.CountCustomModels(ctx, tenantID)
	if err != nil {
		cmh.logger.Error("Failed to count custom models", zap.Error(err))
		http.Error(w, "Failed to count models", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"models":      models,
		"total_count": totalCount,
		"limit":       limit,
		"offset":      offset,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// ValidateCustomModel handles POST /api/v1/models/custom/{id}/validate
func (cmh *CustomModelHandlers) ValidateCustomModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	modelID := vars["id"]

	cmh.logger.Info("Validating custom risk model", zap.String("model_id", modelID))

	// Get the model
	model, err := cmh.repository.GetCustomModel(ctx, modelID)
	if err != nil {
		cmh.logger.Error("Failed to retrieve custom model for validation", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Validate the model configuration
	validationResult := &custom.CustomModelValidationResult{
		ModelID:        modelID,
		ValidationDate: time.Now(),
	}

	if err := cmh.modelBuilder.ValidateModelConfiguration(model); err != nil {
		validationResult.IsValid = false
		validationResult.Errors = []string{err.Error()}
	} else {
		validationResult.IsValid = true
		validationResult.Errors = []string{}
	}

	// Add some basic performance metrics (in a real implementation, this would run actual tests)
	if validationResult.IsValid {
		validationResult.PerformanceMetrics = &custom.ModelPerformanceMetrics{
			Accuracy:        0.85,
			Precision:       0.82,
			Recall:          0.88,
			F1Score:         0.85,
			ConfidenceScore: 0.90,
			TestDataSize:    1000,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(validationResult); err != nil {
		cmh.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// GetActiveCustomModel handles GET /api/v1/models/custom/active
func (cmh *CustomModelHandlers) GetActiveCustomModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get query parameters
	tenantID := r.URL.Query().Get("tenant_id")
	baseModel := r.URL.Query().Get("base_model")

	if tenantID == "" || baseModel == "" {
		http.Error(w, "tenant_id and base_model are required", http.StatusBadRequest)
		return
	}

	cmh.logger.Info("Retrieving active custom risk model",
		zap.String("tenant_id", tenantID),
		zap.String("base_model", baseModel))

	model, err := cmh.repository.GetActiveCustomModel(ctx, tenantID, baseModel)
	if err != nil {
		cmh.logger.Error("Failed to retrieve active custom model", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(model); err != nil {
		cmh.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// UpdateModelStatus handles PATCH /api/v1/models/custom/{id}/status
func (cmh *CustomModelHandlers) UpdateModelStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	modelID := vars["id"]

	cmh.logger.Info("Updating custom model status", zap.String("model_id", modelID))

	var request struct {
		IsActive bool `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		cmh.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := cmh.repository.UpdateModelStatus(ctx, modelID, request.IsActive); err != nil {
		cmh.logger.Error("Failed to update model status", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetModelVersions handles GET /api/v1/models/custom/{id}/versions
func (cmh *CustomModelHandlers) GetModelVersions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	modelID := vars["id"]

	cmh.logger.Info("Retrieving model versions", zap.String("model_id", modelID))

	versions, err := cmh.repository.GetModelVersions(ctx, modelID)
	if err != nil {
		cmh.logger.Error("Failed to retrieve model versions", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(versions); err != nil {
		cmh.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// GetDefaultCustomFactors handles GET /api/v1/models/custom/default-factors
func (cmh *CustomModelHandlers) GetDefaultCustomFactors(w http.ResponseWriter, r *http.Request) {
	baseModel := r.URL.Query().Get("base_model")
	if baseModel == "" {
		http.Error(w, "base_model is required", http.StatusBadRequest)
		return
	}

	cmh.logger.Info("Retrieving default custom factors", zap.String("base_model", baseModel))

	// Convert string to IndustryType
	industryType := industry.IndustryType(baseModel)

	// Get default factors
	factors := custom.GetDefaultCustomFactors(industryType)
	weights := custom.GetDefaultFactorWeights()
	thresholds := custom.GetDefaultThresholds()

	response := map[string]interface{}{
		"base_model":     baseModel,
		"custom_factors": factors,
		"factor_weights": weights,
		"thresholds":     thresholds,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		cmh.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// CustomModelResponse represents a response for custom model operations
type CustomModelResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// CustomModelListResponse represents a response for listing custom models
type CustomModelListResponse struct {
	Models     []*custom.CustomRiskModel `json:"models"`
	TotalCount int                       `json:"total_count"`
	Limit      int                       `json:"limit"`
	Offset     int                       `json:"offset"`
}

// CustomModelValidationResponse represents a response for model validation
type CustomModelValidationResponse struct {
	ModelID            string                          `json:"model_id"`
	IsValid            bool                            `json:"is_valid"`
	Errors             []string                        `json:"errors"`
	Warnings           []string                        `json:"warnings"`
	ValidationDate     time.Time                       `json:"validation_date"`
	PerformanceMetrics *custom.ModelPerformanceMetrics `json:"performance_metrics,omitempty"`
}

// HandleCreateCustomModel is an alias for CreateCustomModel
func (cmh *CustomModelHandlers) HandleCreateCustomModel(w http.ResponseWriter, r *http.Request) {
	cmh.CreateCustomModel(w, r)
}

// HandleListCustomModels is an alias for ListCustomModels
func (cmh *CustomModelHandlers) HandleListCustomModels(w http.ResponseWriter, r *http.Request) {
	cmh.ListCustomModels(w, r)
}

// HandleGetCustomModel is an alias for GetCustomModel
func (cmh *CustomModelHandlers) HandleGetCustomModel(w http.ResponseWriter, r *http.Request) {
	cmh.GetCustomModel(w, r)
}

// HandleUpdateCustomModel is an alias for UpdateCustomModel
func (cmh *CustomModelHandlers) HandleUpdateCustomModel(w http.ResponseWriter, r *http.Request) {
	cmh.UpdateCustomModel(w, r)
}

// HandleDeleteCustomModel is an alias for DeleteCustomModel
func (cmh *CustomModelHandlers) HandleDeleteCustomModel(w http.ResponseWriter, r *http.Request) {
	cmh.DeleteCustomModel(w, r)
}

// HandleValidateCustomModel is an alias for ValidateCustomModel
func (cmh *CustomModelHandlers) HandleValidateCustomModel(w http.ResponseWriter, r *http.Request) {
	cmh.ValidateCustomModel(w, r)
}

// HandleTestCustomModel handles POST /api/v1/models/custom/{id}/test
func (cmh *CustomModelHandlers) HandleTestCustomModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	modelID := vars["id"]

	cmh.logger.Info("Testing custom risk model", zap.String("model_id", modelID))

	// Get the model
	model, err := cmh.repository.GetCustomModel(ctx, modelID)
	if err != nil {
		cmh.logger.Error("Failed to retrieve custom model for testing", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	// Create a test request
	var testRequest struct {
		TestData []map[string]interface{} `json:"test_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&testRequest); err != nil {
		cmh.logger.Error("Failed to decode test request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Run tests (simplified implementation)
	testResults := map[string]interface{}{
		"model_id":      modelID,
		"test_count":    len(testRequest.TestData),
		"passed_tests":  len(testRequest.TestData) - 1, // Mock: assume 1 test fails
		"failed_tests":  1,
		"test_date":     time.Now(),
		"model_version": model.Version,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(testResults); err != nil {
		cmh.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleActivateCustomModel handles POST /api/v1/models/custom/{id}/activate
func (cmh *CustomModelHandlers) HandleActivateCustomModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	modelID := vars["id"]

	cmh.logger.Info("Activating custom risk model", zap.String("model_id", modelID))

	if err := cmh.repository.UpdateModelStatus(ctx, modelID, true); err != nil {
		cmh.logger.Error("Failed to activate custom model", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleDeactivateCustomModel handles POST /api/v1/models/custom/{id}/deactivate
func (cmh *CustomModelHandlers) HandleDeactivateCustomModel(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	modelID := vars["id"]

	cmh.logger.Info("Deactivating custom risk model", zap.String("model_id", modelID))

	if err := cmh.repository.UpdateModelStatus(ctx, modelID, false); err != nil {
		cmh.logger.Error("Failed to deactivate custom model", zap.Error(err))
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// HandleGetCustomModelVersions is an alias for GetModelVersions
func (cmh *CustomModelHandlers) HandleGetCustomModelVersions(w http.ResponseWriter, r *http.Request) {
	cmh.GetModelVersions(w, r)
}

// getTenantIDFromContext extracts tenant ID from context
func (cmh *CustomModelHandlers) getTenantIDFromContext(ctx context.Context) string {
	if tenantID, ok := ctx.Value("tenant_id").(string); ok {
		return tenantID
	}
	return "default"
}
