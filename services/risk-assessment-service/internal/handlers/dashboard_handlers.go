package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/reporting"
)

// DashboardHandler handles dashboard API requests
type DashboardHandler struct {
	dashboardService DashboardService
	logger           *zap.Logger
}

// DashboardService interface for dashboard management
type DashboardService interface {
	CreateDashboard(ctx context.Context, request *reporting.DashboardRequest) (*reporting.DashboardResponse, error)
	GetDashboard(ctx context.Context, tenantID, dashboardID string) (*reporting.RiskDashboard, error)
	UpdateDashboard(ctx context.Context, tenantID, dashboardID string, request *reporting.DashboardRequest) (*reporting.DashboardResponse, error)
	DeleteDashboard(ctx context.Context, tenantID, dashboardID string) error
	ListDashboards(ctx context.Context, filter *reporting.DashboardFilter) (*reporting.DashboardListResponse, error)
	GetDashboardData(ctx context.Context, tenantID, dashboardID string, filters *reporting.DashboardFilters) (*reporting.RiskDashboard, error)
	GetRiskOverviewData(ctx context.Context, tenantID string, filters *reporting.DashboardFilters) (*reporting.DashboardSummary, error)
	GetTrendsData(ctx context.Context, tenantID string, filters *reporting.DashboardFilters) (*reporting.DashboardTrends, error)
	GetPredictionsData(ctx context.Context, tenantID string, filters *reporting.DashboardFilters) (*reporting.DashboardPredictions, error)
	GetDashboardMetrics(ctx context.Context, tenantID string) (*reporting.DashboardMetrics, error)
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(dashboardService DashboardService, logger *zap.Logger) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
		logger:           logger,
	}
}

// HandleCreateDashboard handles POST /api/v1/reporting/dashboards
func (h *DashboardHandler) HandleCreateDashboard(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling create dashboard request")

	// Parse request body
	var request reporting.DashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateDashboardRequest(&request); err != nil {
		h.logger.Error("Invalid dashboard request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID from context
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Set tenant ID in request context
	ctx := r.Context()
	ctx = context.WithValue(ctx, "tenant_id", tenantID)

	// Create dashboard
	response, err := h.dashboardService.CreateDashboard(ctx, &request)
	if err != nil {
		h.logger.Error("Failed to create dashboard", zap.Error(err))
		http.Error(w, "Failed to create dashboard", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetDashboard handles GET /api/v1/reporting/dashboards/{id}
func (h *DashboardHandler) HandleGetDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]

	h.logger.Info("Handling get dashboard request",
		zap.String("dashboard_id", dashboardID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get dashboard
	dashboard, err := h.dashboardService.GetDashboard(r.Context(), tenantID, dashboardID)
	if err != nil {
		h.logger.Error("Failed to get dashboard", zap.Error(err))
		http.Error(w, "Failed to get dashboard", http.StatusInternalServerError)
		return
	}

	if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dashboard); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleUpdateDashboard handles PUT /api/v1/reporting/dashboards/{id}
func (h *DashboardHandler) HandleUpdateDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]

	h.logger.Info("Handling update dashboard request",
		zap.String("dashboard_id", dashboardID))

	// Parse request body
	var request reporting.DashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateDashboardRequest(&request); err != nil {
		h.logger.Error("Invalid dashboard request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Update dashboard
	response, err := h.dashboardService.UpdateDashboard(r.Context(), tenantID, dashboardID, &request)
	if err != nil {
		h.logger.Error("Failed to update dashboard", zap.Error(err))
		http.Error(w, "Failed to update dashboard", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleDeleteDashboard handles DELETE /api/v1/reporting/dashboards/{id}
func (h *DashboardHandler) HandleDeleteDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]

	h.logger.Info("Handling delete dashboard request",
		zap.String("dashboard_id", dashboardID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Delete dashboard
	if err := h.dashboardService.DeleteDashboard(r.Context(), tenantID, dashboardID); err != nil {
		h.logger.Error("Failed to delete dashboard", zap.Error(err))
		http.Error(w, "Failed to delete dashboard", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// HandleListDashboards handles GET /api/v1/reporting/dashboards
func (h *DashboardHandler) HandleListDashboards(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling list dashboards request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := h.parseDashboardFilter(r, tenantID)

	// List dashboards
	response, err := h.dashboardService.ListDashboards(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list dashboards", zap.Error(err))
		http.Error(w, "Failed to list dashboards", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetDashboardData handles GET /api/v1/reporting/dashboards/{id}/data
func (h *DashboardHandler) HandleGetDashboardData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]

	h.logger.Info("Handling get dashboard data request",
		zap.String("dashboard_id", dashboardID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse filters from query parameters
	filters := h.parseDashboardFilters(r)

	// Get dashboard data
	dashboard, err := h.dashboardService.GetDashboardData(r.Context(), tenantID, dashboardID, filters)
	if err != nil {
		h.logger.Error("Failed to get dashboard data", zap.Error(err))
		http.Error(w, "Failed to get dashboard data", http.StatusInternalServerError)
		return
	}

	if dashboard == nil {
		http.Error(w, "Dashboard not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dashboard); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetRiskOverview handles GET /api/v1/reporting/dashboard/risk-overview
func (h *DashboardHandler) HandleGetRiskOverview(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get risk overview request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse filters from query parameters
	filters := h.parseDashboardFilters(r)

	// Get risk overview data
	summary, err := h.dashboardService.GetRiskOverviewData(r.Context(), tenantID, filters)
	if err != nil {
		h.logger.Error("Failed to get risk overview data", zap.Error(err))
		http.Error(w, "Failed to get risk overview data", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(summary); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetTrends handles GET /api/v1/reporting/dashboard/trends
func (h *DashboardHandler) HandleGetTrends(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get trends request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse filters from query parameters
	filters := h.parseDashboardFilters(r)

	// Get trends data
	trends, err := h.dashboardService.GetTrendsData(r.Context(), tenantID, filters)
	if err != nil {
		h.logger.Error("Failed to get trends data", zap.Error(err))
		http.Error(w, "Failed to get trends data", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(trends); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetPredictions handles GET /api/v1/reporting/dashboard/predictions
func (h *DashboardHandler) HandleGetPredictions(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get predictions request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse filters from query parameters
	filters := h.parseDashboardFilters(r)

	// Get predictions data
	predictions, err := h.dashboardService.GetPredictionsData(r.Context(), tenantID, filters)
	if err != nil {
		h.logger.Error("Failed to get predictions data", zap.Error(err))
		http.Error(w, "Failed to get predictions data", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(predictions); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetDashboardMetrics handles GET /api/v1/reporting/dashboards/metrics
func (h *DashboardHandler) HandleGetDashboardMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get dashboard metrics request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get dashboard metrics
	metrics, err := h.dashboardService.GetDashboardMetrics(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get dashboard metrics", zap.Error(err))
		http.Error(w, "Failed to get dashboard metrics", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// Helper functions

// validateDashboardRequest validates a dashboard request
func validateDashboardRequest(request *reporting.DashboardRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(request.Name) > 255 {
		return fmt.Errorf("name exceeds maximum length of 255 characters")
	}

	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Validate dashboard type
	validTypes := map[reporting.DashboardType]bool{
		reporting.DashboardTypeRiskOverview: true,
		reporting.DashboardTypeTrends:       true,
		reporting.DashboardTypePredictions:  true,
		reporting.DashboardTypeCompliance:   true,
		reporting.DashboardTypePerformance:  true,
		reporting.DashboardTypeCustom:       true,
	}

	if !validTypes[request.Type] {
		return fmt.Errorf("invalid dashboard type: %s", request.Type)
	}

	return nil
}

// extractTenantID extracts tenant ID from request context
func (h *DashboardHandler) extractTenantID(r *http.Request) string {
	// This would be implemented based on your authentication/authorization system
	// For now, return a default tenant ID
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}
	return "default"
}

// parseDashboardFilter parses query parameters into a dashboard filter
func (h *DashboardHandler) parseDashboardFilter(r *http.Request, tenantID string) *reporting.DashboardFilter {
	filter := &reporting.DashboardFilter{
		TenantID: tenantID,
	}

	// Parse dashboard type filter
	if dashboardType := r.URL.Query().Get("type"); dashboardType != "" {
		filter.Type = reporting.DashboardType(dashboardType)
	}

	// Parse created by filter
	if createdBy := r.URL.Query().Get("created_by"); createdBy != "" {
		filter.CreatedBy = createdBy
	}

	// Parse public filter
	if isPublicStr := r.URL.Query().Get("is_public"); isPublicStr != "" {
		if isPublic, err := strconv.ParseBool(isPublicStr); err == nil {
			filter.IsPublic = &isPublic
		}
	}

	// Parse date range
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	// Parse pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	return filter
}

// parseDashboardFilters parses query parameters into dashboard filters
func (h *DashboardHandler) parseDashboardFilters(r *http.Request) *reporting.DashboardFilters {
	filters := &reporting.DashboardFilters{}

	// Parse date range
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.DateRange.StartDate = &startDate
		}
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters.DateRange.EndDate = &endDate
		}
	}

	if period := r.URL.Query().Get("period"); period != "" {
		filters.DateRange.Period = period
	}

	// Parse industry filter
	if industries := r.URL.Query()["industry"]; len(industries) > 0 {
		filters.Industry = industries
	}

	// Parse country filter
	if countries := r.URL.Query()["country"]; len(countries) > 0 {
		filters.Country = countries
	}

	// Parse risk level filter
	if riskLevels := r.URL.Query()["risk_level"]; len(riskLevels) > 0 {
		filters.RiskLevel = riskLevels
	}

	return filters
}
