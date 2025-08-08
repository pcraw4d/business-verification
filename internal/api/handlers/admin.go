package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pcraw4d/business-verification/internal/api/middleware"
	"github.com/pcraw4d/business-verification/internal/auth"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// AdminHandler handles admin API endpoints
type AdminHandler struct {
	adminService *auth.AdminService
	logger       *observability.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService *auth.AdminService, logger *observability.Logger) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
		logger:       logger,
	}
}

// CreateUser handles POST /v1/admin/users
func (h *AdminHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var request auth.UserManagementRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.WithComponent("admin_handler").WithError(err).Error("Failed to decode create user request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set action
	request.Action = "create"

	// Call admin service
	response, err := h.adminService.CreateUser(ctx, &request)
	if err != nil {
		h.logger.WithComponent("admin_handler").WithError(err).Error("Failed to create user")
		http.Error(w, fmt.Sprintf("Failed to create user: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateUser handles PUT /v1/admin/users/{id}
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from URL
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var request auth.UserManagementRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.WithComponent("admin_handler").WithError(err).Error("Failed to decode update user request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set target user ID and action
	request.TargetUserID = userID
	request.Action = "update"

	// Call admin service
	response, err := h.adminService.UpdateUser(ctx, &request)
	if err != nil {
		h.logger.WithComponent("admin_handler").WithError(err).Error("Failed to update user")
		http.Error(w, fmt.Sprintf("Failed to update user: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteUser handles DELETE /v1/admin/users/{id}
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from URL
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get admin user ID from context (set by auth middleware)
	adminUserID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || adminUserID == "" {
		http.Error(w, "Admin user ID is required", http.StatusBadRequest)
		return
	}

	// Create request
	request := &auth.UserManagementRequest{
		AdminUserID:  adminUserID,
		TargetUserID: userID,
		Action:       "delete",
	}

	// Call admin service
	response, err := h.adminService.DeleteUser(ctx, request)
	if err != nil {
		h.logger.WithComponent("admin_handler").WithError(err).Error("Failed to delete user")
		http.Error(w, fmt.Sprintf("Failed to delete user: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ActivateUser handles POST /v1/admin/users/{id}/activate
func (h *AdminHandler) ActivateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from URL
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || adminUserID == "" {
		http.Error(w, "Admin user ID is required", http.StatusBadRequest)
		return
	}

	// Create request
	request := &auth.UserManagementRequest{
		AdminUserID:  adminUserID,
		TargetUserID: userID,
		Action:       "activate",
	}

	// Call admin service
	response, err := h.adminService.ActivateUser(ctx, request)
	if err != nil {
		h.logger.WithComponent("admin_handler").WithError(err).Error("Failed to activate user")
		http.Error(w, fmt.Sprintf("Failed to activate user: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeactivateUser handles POST /v1/admin/users/{id}/deactivate
func (h *AdminHandler) DeactivateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract user ID from URL
	userID := r.PathValue("id")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || adminUserID == "" {
		http.Error(w, "Admin user ID is required", http.StatusBadRequest)
		return
	}

	// Create request
	request := &auth.UserManagementRequest{
		AdminUserID:  adminUserID,
		TargetUserID: userID,
		Action:       "deactivate",
	}

	// Call admin service
	response, err := h.adminService.DeactivateUser(ctx, request)
	if err != nil {
		h.logger.WithComponent("admin_handler").WithError(err).Error("Failed to deactivate user")
		http.Error(w, fmt.Sprintf("Failed to deactivate user: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListUsers handles GET /v1/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || adminUserID == "" {
		http.Error(w, "Admin user ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	role := r.URL.Query().Get("role")
	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")

	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Create request
	request := &auth.ListUsersRequest{
		AdminUserID: adminUserID,
		Role:        auth.Role(role),
		Status:      status,
		Search:      search,
		Limit:       limit,
		Offset:      offset,
	}

	// Call admin service
	response, err := h.adminService.ListUsers(ctx, request)
	if err != nil {
		h.logger.WithComponent("admin_handler").WithError(err).Error("Failed to list users")
		http.Error(w, fmt.Sprintf("Failed to list users: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetSystemStats handles GET /v1/admin/stats
func (h *AdminHandler) GetSystemStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get admin user ID from context
	adminUserID, ok := middleware.GetUserIDFromContext(ctx)
	if !ok || adminUserID == "" {
		http.Error(w, "Admin user ID is required", http.StatusBadRequest)
		return
	}

	// Call admin service
	stats, err := h.adminService.GetSystemStats(ctx, adminUserID)
	if err != nil {
		h.logger.WithComponent("admin_handler").WithError(err).Error("Failed to get system stats")
		http.Error(w, fmt.Sprintf("Failed to get system stats: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}
