package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/middleware"
	"kyb-platform/services/risk-assessment-service/internal/ml/custom"
)

// CustomModelHandler handles custom risk model API requests
type CustomModelHandler struct {
	modelBuilder *custom.CustomModelBuilder
	logger       *zap.Logger
	errorHandler *middleware.ErrorHandler
}

// NewCustomModelHandler creates a new custom model handler
func NewCustomModelHandler(modelBuilder *custom.CustomModelBuilder, logger *zap.Logger) *CustomModelHandler {
	return &CustomModelHandler{
		modelBuilder: modelBuilder,
		logger:       logger,
		errorHandler: middleware.NewErrorHandler(logger),
	}
}

// HandleCreateCustomModel handles POST /api/v1/models/custom
func (h *CustomModelHandler) HandleCreateCustomModel(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing create custom model request")

	// Get tenant ID from context (set by middleware)
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("tenant ID not found in request context"))
		return
	}

	// Parse request
	var req custom.CreateCustomModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Create the custom model
	model, err := h.modelBuilder.CreateCustomModel(r.Context(), tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to create custom model", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("failed to create custom model: %w", err))
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model)

	h.logger.Info("Custom model created successfully",
		zap.String("model_id", model.ID),
		zap.String("tenant_id", tenantID),
		zap.String("name", model.Name))
}

// HandleGetCustomModel handles GET /api/v1/models/custom/{id}
func (h *CustomModelHandler) HandleGetCustomModel(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing get custom model request")

	// Get tenant ID and model ID from request
	tenantID := h.getTenantID(r)
	modelID := h.getModelID(r)
	if tenantID == "" || modelID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("tenant ID or model ID not found"))
		return
	}

	// Get the custom model
	model, err := h.modelBuilder.Repository.GetCustomModel(r.Context(), tenantID, modelID)
	if err != nil {
		h.logger.Error("Failed to get custom model", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("failed to get custom model: %w", err))
		return
	}

	if model == nil {
		http.Error(w, "Custom model not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model)

	h.logger.Info("Custom model retrieved successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))
}

// HandleUpdateCustomModel handles PUT /api/v1/models/custom/{id}
func (h *CustomModelHandler) HandleUpdateCustomModel(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing update custom model request")

	// Get tenant ID and model ID from request
	tenantID := h.getTenantID(r)
	modelID := h.getModelID(r)
	if tenantID == "" || modelID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("tenant ID or model ID not found"))
		return
	}

	// Parse request
	var req custom.UpdateCustomModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Update the custom model
	model, err := h.modelBuilder.UpdateCustomModel(r.Context(), tenantID, modelID, &req)
	if err != nil {
		h.logger.Error("Failed to update custom model", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("failed to update custom model: %w", err))
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model)

	h.logger.Info("Custom model updated successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID),
		zap.Int("version", model.Version))
}

// HandleDeleteCustomModel handles DELETE /api/v1/models/custom/{id}
func (h *CustomModelHandler) HandleDeleteCustomModel(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing delete custom model request")

	// Get tenant ID and model ID from request
	tenantID := h.getTenantID(r)
	modelID := h.getModelID(r)
	if tenantID == "" || modelID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("tenant ID or model ID not found"))
		return
	}

	// Delete the custom model
	err := h.modelBuilder.Repository.DeleteCustomModel(r.Context(), tenantID, modelID)
	if err != nil {
		h.logger.Error("Failed to delete custom model", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("failed to delete custom model: %w", err))
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Custom model deleted successfully",
		"model_id":   modelID,
		"tenant_id":  tenantID,
		"deleted_at": time.Now(),
	})

	h.logger.Info("Custom model deleted successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))
}

// HandleListCustomModels handles GET /api/v1/models/custom
func (h *CustomModelHandler) HandleListCustomModels(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing list custom models request")

	// Get tenant ID from request
	tenantID := h.getTenantID(r)
	if tenantID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("tenant ID not found"))
		return
	}

	// Parse query parameters
	limit := h.getIntQueryParam(r, "limit", 50)
	offset := h.getIntQueryParam(r, "offset", 0)

	// Validate parameters
	if limit < 1 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// List custom models
	models, err := h.modelBuilder.Repository.ListCustomModels(r.Context(), tenantID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to list custom models", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("failed to list custom models: %w", err))
		return
	}

	// Create response
	response := map[string]interface{}{
		"models":   models,
		"count":    len(models),
		"limit":    limit,
		"offset":   offset,
		"has_more": len(models) == limit, // Simple pagination indicator
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Custom models listed successfully",
		zap.String("tenant_id", tenantID),
		zap.Int("count", len(models)))
}

// HandleValidateCustomModel handles POST /api/v1/models/custom/{id}/validate
func (h *CustomModelHandler) HandleValidateCustomModel(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing validate custom model request")

	// Get tenant ID and model ID from request
	tenantID := h.getTenantID(r)
	modelID := h.getModelID(r)
	if tenantID == "" || modelID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("tenant ID or model ID not found"))
		return
	}

	// Validate the custom model
	result, err := h.modelBuilder.ValidateCustomModel(r.Context(), tenantID, modelID)
	if err != nil {
		h.logger.Error("Failed to validate custom model", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("failed to validate custom model: %w", err))
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	h.logger.Info("Custom model validation completed",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID),
		zap.Bool("is_valid", result.IsValid),
		zap.Int("error_count", len(result.Errors)),
		zap.Int("warning_count", len(result.Warnings)))
}

// HandleTestCustomModel handles POST /api/v1/models/custom/{id}/test
func (h *CustomModelHandler) HandleTestCustomModel(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing test custom model request")

	// Get tenant ID and model ID from request
	tenantID := h.getTenantID(r)
	modelID := h.getModelID(r)
	if tenantID == "" || modelID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("tenant ID or model ID not found"))
		return
	}

	// Parse request
	var req custom.TestModelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorHandler.HandleError(w, r, fmt.Errorf("invalid request body: %w", err))
		return
	}

	// Test the custom model
	result, err := h.modelBuilder.TestCustomModel(r.Context(), tenantID, modelID, &req)
	if err != nil {
		h.logger.Error("Failed to test custom model", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("failed to test custom model: %w", err))
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

	h.logger.Info("Custom model test completed",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID),
		zap.Float64("risk_score", result.RiskScore),
		zap.String("risk_level", string(result.RiskLevel)))
}

// HandleActivateCustomModel handles POST /api/v1/models/custom/{id}/activate
func (h *CustomModelHandler) HandleActivateCustomModel(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing activate custom model request")

	// Get tenant ID and model ID from request
	tenantID := h.getTenantID(r)
	modelID := h.getModelID(r)
	if tenantID == "" || modelID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("tenant ID or model ID not found"))
		return
	}

	// Activate the custom model
	err := h.modelBuilder.Repository.ActivateCustomModel(r.Context(), tenantID, modelID)
	if err != nil {
		h.logger.Error("Failed to activate custom model", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("failed to activate custom model: %w", err))
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":      "Custom model activated successfully",
		"model_id":     modelID,
		"tenant_id":    tenantID,
		"activated_at": time.Now(),
	})

	h.logger.Info("Custom model activated successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))
}

// HandleDeactivateCustomModel handles POST /api/v1/models/custom/{id}/deactivate
func (h *CustomModelHandler) HandleDeactivateCustomModel(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing deactivate custom model request")

	// Get tenant ID and model ID from request
	tenantID := h.getTenantID(r)
	modelID := h.getModelID(r)
	if tenantID == "" || modelID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("tenant ID or model ID not found"))
		return
	}

	// Deactivate the custom model
	err := h.modelBuilder.Repository.DeactivateCustomModel(r.Context(), tenantID, modelID)
	if err != nil {
		h.logger.Error("Failed to deactivate custom model", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("failed to deactivate custom model: %w", err))
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":        "Custom model deactivated successfully",
		"model_id":       modelID,
		"tenant_id":      tenantID,
		"deactivated_at": time.Now(),
	})

	h.logger.Info("Custom model deactivated successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID))
}

// HandleGetCustomModelVersions handles GET /api/v1/models/custom/{id}/versions
func (h *CustomModelHandler) HandleGetCustomModelVersions(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Processing get custom model versions request")

	// Get tenant ID and model ID from request
	tenantID := h.getTenantID(r)
	modelID := h.getModelID(r)
	if tenantID == "" || modelID == "" {
		h.errorHandler.HandleError(w, r, fmt.Errorf("tenant ID or model ID not found"))
		return
	}

	// Get model versions
	versions, err := h.modelBuilder.Repository.GetCustomModelVersions(r.Context(), tenantID, modelID)
	if err != nil {
		h.logger.Error("Failed to get custom model versions", zap.Error(err))
		h.errorHandler.HandleError(w, r, fmt.Errorf("failed to get custom model versions: %w", err))
		return
	}

	// Create response
	response := map[string]interface{}{
		"model_id":  modelID,
		"tenant_id": tenantID,
		"versions":  versions,
		"count":     len(versions),
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	h.logger.Info("Custom model versions retrieved successfully",
		zap.String("model_id", modelID),
		zap.String("tenant_id", tenantID),
		zap.Int("version_count", len(versions)))
}

// Helper methods

// getTenantID extracts tenant ID from request context
func (h *CustomModelHandler) getTenantID(r *http.Request) string {
	// This would be set by tenant middleware
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}
	// Fallback to context value
	if tenantID := r.Context().Value("tenant_id"); tenantID != nil {
		if id, ok := tenantID.(string); ok {
			return id
		}
	}
	return ""
}

// getModelID extracts model ID from URL path
func (h *CustomModelHandler) getModelID(r *http.Request) string {
	// Extract from URL path - this would typically be done by the router
	// For now, we'll assume it's in the URL path
	// In a real implementation, you'd use gorilla/mux or similar
	path := r.URL.Path
	// Simple extraction - in practice, use proper routing
	if len(path) > 20 { // "/api/v1/models/custom/" is ~20 chars
		return path[20:]
	}
	return ""
}

// getIntQueryParam extracts and parses an integer query parameter
func (h *CustomModelHandler) getIntQueryParam(r *http.Request, key string, defaultValue int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}

	return defaultValue
}
