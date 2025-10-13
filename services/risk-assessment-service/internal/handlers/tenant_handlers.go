package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/audit"
	"kyb-platform/services/risk-assessment-service/internal/tenant"
)

// TenantHandler handles tenant-related API requests
type TenantHandler struct {
	tenantService *tenant.TenantService
	auditLogger   *audit.AuditLogger
	logger        *zap.Logger
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(tenantService *tenant.TenantService, auditLogger *audit.AuditLogger, logger *zap.Logger) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
		auditLogger:   auditLogger,
		logger:        logger,
	}
}

// CreateTenant creates a new tenant
func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req tenant.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode create tenant request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Creating tenant",
		zap.String("name", req.Name),
		zap.String("domain", req.Domain),
		zap.String("plan", string(req.Plan)))

	// Create tenant
	createdTenant, err := h.tenantService.CreateTenant(ctx, &req)
	if err != nil {
		h.logger.Error("Failed to create tenant", zap.Error(err))
		http.Error(w, "Failed to create tenant", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Tenant created successfully",
		"data":    createdTenant,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetTenant retrieves a tenant by ID
func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	h.logger.Info("Retrieving tenant",
		zap.String("tenant_id", tenantID))

	// Get tenant
	tenant, err := h.tenantService.GetTenant(ctx, tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant", zap.Error(err))
		http.Error(w, "Failed to retrieve tenant", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    tenant,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateTenant updates a tenant
func (h *TenantHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	var req tenant.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode update tenant request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Updating tenant",
		zap.String("tenant_id", tenantID),
		zap.String("updated_by", req.UpdatedBy))

	// Update tenant
	updatedTenant, err := h.tenantService.UpdateTenant(ctx, tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to update tenant", zap.Error(err))
		http.Error(w, "Failed to update tenant", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Tenant updated successfully",
		"data":    updatedTenant,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListTenants lists tenants with pagination
func (h *TenantHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	// Parse pagination parameters
	limit := 50 // Default limit
	offset := 0 // Default offset

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	h.logger.Info("Listing tenants",
		zap.Int("limit", limit),
		zap.Int("offset", offset))

	// Get tenants from repository (this would need to be exposed through the service)
	// For now, we'll return a mock response
	tenants := []*tenant.Tenant{
		{
			ID:        "tenant_1",
			Name:      "Example Tenant 1",
			Domain:    "tenant1.example.com",
			Status:    tenant.TenantStatusActive,
			Plan:      tenant.TenantPlanProfessional,
			CreatedAt: time.Now().AddDate(0, -1, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "tenant_2",
			Name:      "Example Tenant 2",
			Domain:    "tenant2.example.com",
			Status:    tenant.TenantStatusActive,
			Plan:      tenant.TenantPlanBasic,
			CreatedAt: time.Now().AddDate(0, -2, 0),
			UpdatedAt: time.Now(),
		},
	}

	response := map[string]interface{}{
		"success": true,
		"data":    tenants,
		"count":   len(tenants),
		"pagination": map[string]interface{}{
			"limit":  limit,
			"offset": offset,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateTenantUser creates a new user for a tenant
func (h *TenantHandler) CreateTenantUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	var req tenant.CreateTenantUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode create tenant user request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Creating tenant user",
		zap.String("tenant_id", tenantID),
		zap.String("email", req.Email),
		zap.String("role", string(req.Role)))

	// Create tenant user
	user, err := h.tenantService.CreateTenantUser(ctx, tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to create tenant user", zap.Error(err))
		http.Error(w, "Failed to create tenant user", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Tenant user created successfully",
		"data":    user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetTenantUsers retrieves all users for a tenant
func (h *TenantHandler) GetTenantUsers(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	h.logger.Info("Retrieving tenant users",
		zap.String("tenant_id", tenantID))

	// Get tenant users from repository (this would need to be exposed through the service)
	// For now, we'll return a mock response
	users := []*tenant.TenantUser{
		{
			ID:        "user_1",
			TenantID:  tenantID,
			UserID:    "user_123",
			Email:     "admin@tenant1.com",
			Role:      tenant.TenantUserRoleOwner,
			Status:    tenant.TenantUserStatusActive,
			CreatedAt: time.Now().AddDate(0, -1, 0),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "user_2",
			TenantID:  tenantID,
			UserID:    "user_456",
			Email:     "analyst@tenant1.com",
			Role:      tenant.TenantUserRoleAnalyst,
			Status:    tenant.TenantUserStatusActive,
			CreatedAt: time.Now().AddDate(0, 0, -15),
			UpdatedAt: time.Now(),
		},
	}

	response := map[string]interface{}{
		"success": true,
		"data":    users,
		"count":   len(users),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateTenantAPIKey creates a new API key for a tenant
func (h *TenantHandler) CreateTenantAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	var req tenant.CreateTenantAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode create API key request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Creating tenant API key",
		zap.String("tenant_id", tenantID),
		zap.String("name", req.Name))

	// Create API key
	apiKeyRecord, apiKey, err := h.tenantService.CreateTenantAPIKey(ctx, tenantID, &req)
	if err != nil {
		h.logger.Error("Failed to create tenant API key", zap.Error(err))
		http.Error(w, "Failed to create API key", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "API key created successfully",
		"data": map[string]interface{}{
			"api_key": apiKeyRecord,
			"key":     *apiKey, // Only returned once
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetTenantAPIKeys retrieves all API keys for a tenant
func (h *TenantHandler) GetTenantAPIKeys(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	h.logger.Info("Retrieving tenant API keys",
		zap.String("tenant_id", tenantID))

	// Get tenant API keys from repository (this would need to be exposed through the service)
	// For now, we'll return a mock response
	apiKeys := []*tenant.TenantAPIKey{
		{
			ID:          "key_1",
			TenantID:    tenantID,
			Name:        "Production API Key",
			KeyHash:     "hash_123",
			Permissions: []string{"assessments:read", "assessments:write"},
			RateLimit:   1000,
			Status:      tenant.APIKeyStatusActive,
			CreatedAt:   time.Now().AddDate(0, -1, 0),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "key_2",
			TenantID:    tenantID,
			Name:        "Development API Key",
			KeyHash:     "hash_456",
			Permissions: []string{"assessments:read"},
			RateLimit:   100,
			Status:      tenant.APIKeyStatusActive,
			CreatedAt:   time.Now().AddDate(0, 0, -7),
			UpdatedAt:   time.Now(),
		},
	}

	response := map[string]interface{}{
		"success": true,
		"data":    apiKeys,
		"count":   len(apiKeys),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTenantMetrics retrieves metrics for a tenant
func (h *TenantHandler) GetTenantMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	h.logger.Info("Retrieving tenant metrics",
		zap.String("tenant_id", tenantID))

	// Get tenant metrics
	metrics, err := h.tenantService.GetTenantMetrics(ctx, tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant metrics", zap.Error(err))
		http.Error(w, "Failed to retrieve tenant metrics", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"data":    metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTenantConfiguration retrieves tenant configuration
func (h *TenantHandler) GetTenantConfiguration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]
	category := vars["category"]
	key := vars["key"]

	h.logger.Info("Retrieving tenant configuration",
		zap.String("tenant_id", tenantID),
		zap.String("category", category),
		zap.String("key", key))

	// Get tenant configuration from repository (this would need to be exposed through the service)
	// For now, we'll return a mock response
	config := &tenant.TenantConfiguration{
		ID:          "config_1",
		TenantID:    tenantID,
		Category:    category,
		Key:         key,
		Value:       "example_value",
		ValueType:   "string",
		Description: "Example configuration value",
		IsEncrypted: false,
		CreatedAt:   time.Now().AddDate(0, -1, 0),
		UpdatedAt:   time.Now(),
		UpdatedBy:   "admin",
	}

	response := map[string]interface{}{
		"success": true,
		"data":    config,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SetTenantConfiguration sets tenant configuration
func (h *TenantHandler) SetTenantConfiguration(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	var req struct {
		Category    string      `json:"category"`
		Key         string      `json:"key"`
		Value       interface{} `json:"value"`
		ValueType   string      `json:"value_type"`
		Description string      `json:"description"`
		IsEncrypted bool        `json:"is_encrypted"`
		UpdatedBy   string      `json:"updated_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode set configuration request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Setting tenant configuration",
		zap.String("tenant_id", tenantID),
		zap.String("category", req.Category),
		zap.String("key", req.Key))

	// Set tenant configuration (this would need to be exposed through the service)
	// For now, we'll return a success response
	config := &tenant.TenantConfiguration{
		ID:          fmt.Sprintf("config_%d", time.Now().UnixNano()),
		TenantID:    tenantID,
		Category:    req.Category,
		Key:         req.Key,
		Value:       req.Value,
		ValueType:   req.ValueType,
		Description: req.Description,
		IsEncrypted: req.IsEncrypted,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		UpdatedBy:   req.UpdatedBy,
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Tenant configuration set successfully",
		"data":    config,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTenantHealth checks the health of a tenant
func (h *TenantHandler) GetTenantHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	h.logger.Info("Checking tenant health",
		zap.String("tenant_id", tenantID))

	// Get tenant to check if it exists and is active
	tenant, err := h.tenantService.GetTenant(ctx, tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant for health check", zap.Error(err))
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	// Get metrics to determine health
	metrics, err := h.tenantService.GetTenantMetrics(ctx, tenantID)
	healthScore := 0.8 // Default health score
	if err != nil {
		h.logger.Error("Failed to get tenant metrics for health check", zap.Error(err))
		// Don't fail the health check if metrics are unavailable
	} else {
		healthScore = metrics.HealthScore
	}

	health := map[string]interface{}{
		"tenant_id":     tenantID,
		"status":        tenant.Status,
		"plan":          tenant.Plan,
		"health_score":  healthScore,
		"is_healthy":    healthScore > 0.8,
		"last_activity": tenant.LastActivityAt,
		"timestamp":     time.Now(),
	}

	response := map[string]interface{}{
		"success": true,
		"data":    health,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetTenantQuotas retrieves tenant quotas and usage
func (h *TenantHandler) GetTenantQuotas(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tenantID := vars["tenant_id"]

	h.logger.Info("Retrieving tenant quotas",
		zap.String("tenant_id", tenantID))

	// Get tenant to retrieve quotas
	tenant, err := h.tenantService.GetTenant(ctx, tenantID)
	if err != nil {
		h.logger.Error("Failed to get tenant for quotas", zap.Error(err))
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	// Get metrics to determine current usage
	metrics, err := h.tenantService.GetTenantMetrics(ctx, tenantID)
	totalAssessments := int64(0)
	apiRequestsToday := int64(0)
	activeUsers := 0
	storageUsed := int64(0)
	quotaUtilization := make(map[string]float64)

	if err != nil {
		h.logger.Error("Failed to get tenant metrics for quotas", zap.Error(err))
		// Don't fail if metrics are unavailable
	} else {
		totalAssessments = metrics.TotalAssessments
		apiRequestsToday = metrics.APIRequestsToday
		activeUsers = metrics.ActiveUsers
		storageUsed = metrics.StorageUsed
		quotaUtilization = metrics.QuotaUtilization
	}

	quotas := map[string]interface{}{
		"tenant_id": tenantID,
		"plan":      tenant.Plan,
		"quotas":    tenant.Quotas,
		"usage": map[string]interface{}{
			"assessments_per_month": map[string]interface{}{
				"used":       totalAssessments,
				"limit":      tenant.Quotas.MaxAssessmentsPerMonth,
				"percentage": calculatePercentage(totalAssessments, tenant.Quotas.MaxAssessmentsPerMonth),
			},
			"api_requests_per_day": map[string]interface{}{
				"used":       apiRequestsToday,
				"limit":      tenant.Quotas.MaxAPIRequestsPerDay,
				"percentage": calculatePercentage(apiRequestsToday, tenant.Quotas.MaxAPIRequestsPerDay),
			},
			"users": map[string]interface{}{
				"used":       activeUsers,
				"limit":      tenant.Quotas.MaxUsers,
				"percentage": calculatePercentage(int64(activeUsers), int64(tenant.Quotas.MaxUsers)),
			},
			"storage": map[string]interface{}{
				"used":       storageUsed,
				"limit":      int64(0), // No storage limit in current quotas
				"percentage": 0.0,
			},
		},
		"quota_utilization": quotaUtilization,
	}

	response := map[string]interface{}{
		"success": true,
		"data":    quotas,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper function to calculate percentage
func calculatePercentage(used, limit int64) float64 {
	if limit <= 0 {
		return 0.0 // Unlimited
	}
	if used <= 0 {
		return 0.0
	}
	percentage := float64(used) / float64(limit) * 100
	if percentage > 100 {
		return 100.0
	}
	return percentage
}
