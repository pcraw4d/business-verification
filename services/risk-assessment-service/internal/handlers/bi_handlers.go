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

	"kyb-platform/services/risk-assessment-service/internal/businessintelligence"
)

// BIHandler handles business intelligence API requests
type BIHandler struct {
	biService BIService
	logger    *zap.Logger
}

// BIService interface for business intelligence operations
type BIService interface {
	CreateDataSync(ctx context.Context, request *businessintelligence.DataSyncRequest) (*businessintelligence.DataSyncResponse, error)
	GetDataSync(ctx context.Context, tenantID, syncID string) (*businessintelligence.DataSync, error)
	ListDataSyncs(ctx context.Context, filter *businessintelligence.BIFilter) (*businessintelligence.DataSyncListResponse, error)
	UpdateDataSync(ctx context.Context, tenantID, syncID string, request *businessintelligence.DataSyncRequest) (*businessintelligence.DataSyncResponse, error)
	DeleteDataSync(ctx context.Context, tenantID, syncID string) error
	RunDataSync(ctx context.Context, tenantID, syncID string) (*businessintelligence.DataSyncResponse, error)
	PauseDataSync(ctx context.Context, tenantID, syncID string) error
	ResumeDataSync(ctx context.Context, tenantID, syncID string) error
	CreateDataExport(ctx context.Context, request *businessintelligence.DataExportRequest) (*businessintelligence.DataExportResponse, error)
	GetDataExport(ctx context.Context, tenantID, exportID string) (*businessintelligence.DataExport, error)
	ListDataExports(ctx context.Context, filter *businessintelligence.BIFilter) (*businessintelligence.DataExportListResponse, error)
	DeleteDataExport(ctx context.Context, tenantID, exportID string) error
	CreateBIQuery(ctx context.Context, request *businessintelligence.BIQueryRequest) (*businessintelligence.BIQueryResponse, error)
	GetBIQuery(ctx context.Context, tenantID, queryID string) (*businessintelligence.BIQuery, error)
	ListBIQueries(ctx context.Context, filter *businessintelligence.BIFilter) (*businessintelligence.BIQueryListResponse, error)
	UpdateBIQuery(ctx context.Context, tenantID, queryID string, request *businessintelligence.BIQueryRequest) (*businessintelligence.BIQueryResponse, error)
	DeleteBIQuery(ctx context.Context, tenantID, queryID string) error
	ExecuteBIQuery(ctx context.Context, tenantID, queryID string, parameters map[string]interface{}) (*businessintelligence.BIQueryResult, error)
	CreateBIDashboard(ctx context.Context, request *businessintelligence.BIDashboardRequest) (*businessintelligence.BIDashboardResponse, error)
	GetBIDashboard(ctx context.Context, tenantID, dashboardID string) (*businessintelligence.BIDashboard, error)
	ListBIDashboards(ctx context.Context, filter *businessintelligence.BIFilter) (*businessintelligence.BIDashboardListResponse, error)
	UpdateBIDashboard(ctx context.Context, tenantID, dashboardID string, request *businessintelligence.BIDashboardRequest) (*businessintelligence.BIDashboardResponse, error)
	DeleteBIDashboard(ctx context.Context, tenantID, dashboardID string) error
	GetBIMetrics(ctx context.Context, tenantID string) (*businessintelligence.BIMetrics, error)
	GetDataSyncMetrics(ctx context.Context, tenantID string) (*businessintelligence.DataSyncMetrics, error)
	GetQueryPerformanceMetrics(ctx context.Context, tenantID string) (*businessintelligence.QueryPerformanceMetrics, error)
}

// NewBIHandler creates a new BI handler
func NewBIHandler(biService BIService, logger *zap.Logger) *BIHandler {
	return &BIHandler{
		biService: biService,
		logger:    logger,
	}
}

// Data Sync Handlers

// HandleCreateDataSync handles POST /api/v1/bi/data-syncs
func (h *BIHandler) HandleCreateDataSync(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling create data sync request")

	// Parse request body
	var request businessintelligence.DataSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateDataSyncRequest(&request); err != nil {
		h.logger.Error("Invalid data sync request", zap.Error(err))
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

	// Create data sync
	response, err := h.biService.CreateDataSync(ctx, &request)
	if err != nil {
		h.logger.Error("Failed to create data sync", zap.Error(err))
		http.Error(w, "Failed to create data sync", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetDataSync handles GET /api/v1/bi/data-syncs/{id}
func (h *BIHandler) HandleGetDataSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	syncID := vars["id"]

	h.logger.Info("Handling get data sync request",
		zap.String("sync_id", syncID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get data sync
	sync, err := h.biService.GetDataSync(r.Context(), tenantID, syncID)
	if err != nil {
		h.logger.Error("Failed to get data sync", zap.Error(err))
		http.Error(w, "Failed to get data sync", http.StatusInternalServerError)
		return
	}

	if sync == nil {
		http.Error(w, "Data sync not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(sync); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleListDataSyncs handles GET /api/v1/bi/data-syncs
func (h *BIHandler) HandleListDataSyncs(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling list data syncs request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := h.parseBIFilter(r, tenantID)

	// List data syncs
	response, err := h.biService.ListDataSyncs(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list data syncs", zap.Error(err))
		http.Error(w, "Failed to list data syncs", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleUpdateDataSync handles PUT /api/v1/bi/data-syncs/{id}
func (h *BIHandler) HandleUpdateDataSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	syncID := vars["id"]

	h.logger.Info("Handling update data sync request",
		zap.String("sync_id", syncID))

	// Parse request body
	var request businessintelligence.DataSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateDataSyncRequest(&request); err != nil {
		h.logger.Error("Invalid data sync request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Update data sync
	response, err := h.biService.UpdateDataSync(r.Context(), tenantID, syncID, &request)
	if err != nil {
		h.logger.Error("Failed to update data sync", zap.Error(err))
		http.Error(w, "Failed to update data sync", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleDeleteDataSync handles DELETE /api/v1/bi/data-syncs/{id}
func (h *BIHandler) HandleDeleteDataSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	syncID := vars["id"]

	h.logger.Info("Handling delete data sync request",
		zap.String("sync_id", syncID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Delete data sync
	if err := h.biService.DeleteDataSync(r.Context(), tenantID, syncID); err != nil {
		h.logger.Error("Failed to delete data sync", zap.Error(err))
		http.Error(w, "Failed to delete data sync", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// HandleRunDataSync handles POST /api/v1/bi/data-syncs/{id}/run
func (h *BIHandler) HandleRunDataSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	syncID := vars["id"]

	h.logger.Info("Handling run data sync request",
		zap.String("sync_id", syncID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Run data sync
	response, err := h.biService.RunDataSync(r.Context(), tenantID, syncID)
	if err != nil {
		h.logger.Error("Failed to run data sync", zap.Error(err))
		http.Error(w, "Failed to run data sync", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandlePauseDataSync handles POST /api/v1/bi/data-syncs/{id}/pause
func (h *BIHandler) HandlePauseDataSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	syncID := vars["id"]

	h.logger.Info("Handling pause data sync request",
		zap.String("sync_id", syncID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Pause data sync
	if err := h.biService.PauseDataSync(r.Context(), tenantID, syncID); err != nil {
		h.logger.Error("Failed to pause data sync", zap.Error(err))
		http.Error(w, "Failed to pause data sync", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// HandleResumeDataSync handles POST /api/v1/bi/data-syncs/{id}/resume
func (h *BIHandler) HandleResumeDataSync(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	syncID := vars["id"]

	h.logger.Info("Handling resume data sync request",
		zap.String("sync_id", syncID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Resume data sync
	if err := h.biService.ResumeDataSync(r.Context(), tenantID, syncID); err != nil {
		h.logger.Error("Failed to resume data sync", zap.Error(err))
		http.Error(w, "Failed to resume data sync", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// Data Export Handlers

// HandleCreateDataExport handles POST /api/v1/bi/data-exports
func (h *BIHandler) HandleCreateDataExport(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling create data export request")

	// Parse request body
	var request businessintelligence.DataExportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateDataExportRequest(&request); err != nil {
		h.logger.Error("Invalid data export request", zap.Error(err))
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

	// Create data export
	response, err := h.biService.CreateDataExport(ctx, &request)
	if err != nil {
		h.logger.Error("Failed to create data export", zap.Error(err))
		http.Error(w, "Failed to create data export", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetDataExport handles GET /api/v1/bi/data-exports/{id}
func (h *BIHandler) HandleGetDataExport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exportID := vars["id"]

	h.logger.Info("Handling get data export request",
		zap.String("export_id", exportID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get data export
	export, err := h.biService.GetDataExport(r.Context(), tenantID, exportID)
	if err != nil {
		h.logger.Error("Failed to get data export", zap.Error(err))
		http.Error(w, "Failed to get data export", http.StatusInternalServerError)
		return
	}

	if export == nil {
		http.Error(w, "Data export not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(export); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleListDataExports handles GET /api/v1/bi/data-exports
func (h *BIHandler) HandleListDataExports(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling list data exports request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := h.parseBIFilter(r, tenantID)

	// List data exports
	response, err := h.biService.ListDataExports(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list data exports", zap.Error(err))
		http.Error(w, "Failed to list data exports", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleDeleteDataExport handles DELETE /api/v1/bi/data-exports/{id}
func (h *BIHandler) HandleDeleteDataExport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exportID := vars["id"]

	h.logger.Info("Handling delete data export request",
		zap.String("export_id", exportID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Delete data export
	if err := h.biService.DeleteDataExport(r.Context(), tenantID, exportID); err != nil {
		h.logger.Error("Failed to delete data export", zap.Error(err))
		http.Error(w, "Failed to delete data export", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// BI Query Handlers

// HandleCreateBIQuery handles POST /api/v1/bi/queries
func (h *BIHandler) HandleCreateBIQuery(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling create BI query request")

	// Parse request body
	var request businessintelligence.BIQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateBIQueryRequest(&request); err != nil {
		h.logger.Error("Invalid BI query request", zap.Error(err))
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

	// Create BI query
	response, err := h.biService.CreateBIQuery(ctx, &request)
	if err != nil {
		h.logger.Error("Failed to create BI query", zap.Error(err))
		http.Error(w, "Failed to create BI query", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetBIQuery handles GET /api/v1/bi/queries/{id}
func (h *BIHandler) HandleGetBIQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queryID := vars["id"]

	h.logger.Info("Handling get BI query request",
		zap.String("query_id", queryID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get BI query
	query, err := h.biService.GetBIQuery(r.Context(), tenantID, queryID)
	if err != nil {
		h.logger.Error("Failed to get BI query", zap.Error(err))
		http.Error(w, "Failed to get BI query", http.StatusInternalServerError)
		return
	}

	if query == nil {
		http.Error(w, "BI query not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(query); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleListBIQueries handles GET /api/v1/bi/queries
func (h *BIHandler) HandleListBIQueries(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling list BI queries request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := h.parseBIFilter(r, tenantID)

	// List BI queries
	response, err := h.biService.ListBIQueries(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list BI queries", zap.Error(err))
		http.Error(w, "Failed to list BI queries", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleUpdateBIQuery handles PUT /api/v1/bi/queries/{id}
func (h *BIHandler) HandleUpdateBIQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queryID := vars["id"]

	h.logger.Info("Handling update BI query request",
		zap.String("query_id", queryID))

	// Parse request body
	var request businessintelligence.BIQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateBIQueryRequest(&request); err != nil {
		h.logger.Error("Invalid BI query request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Update BI query
	response, err := h.biService.UpdateBIQuery(r.Context(), tenantID, queryID, &request)
	if err != nil {
		h.logger.Error("Failed to update BI query", zap.Error(err))
		http.Error(w, "Failed to update BI query", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleDeleteBIQuery handles DELETE /api/v1/bi/queries/{id}
func (h *BIHandler) HandleDeleteBIQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queryID := vars["id"]

	h.logger.Info("Handling delete BI query request",
		zap.String("query_id", queryID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Delete BI query
	if err := h.biService.DeleteBIQuery(r.Context(), tenantID, queryID); err != nil {
		h.logger.Error("Failed to delete BI query", zap.Error(err))
		http.Error(w, "Failed to delete BI query", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// HandleExecuteBIQuery handles POST /api/v1/bi/queries/{id}/execute
func (h *BIHandler) HandleExecuteBIQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	queryID := vars["id"]

	h.logger.Info("Handling execute BI query request",
		zap.String("query_id", queryID))

	// Parse request body for parameters
	var request struct {
		Parameters map[string]interface{} `json:"parameters"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Execute BI query
	result, err := h.biService.ExecuteBIQuery(r.Context(), tenantID, queryID, request.Parameters)
	if err != nil {
		h.logger.Error("Failed to execute BI query", zap.Error(err))
		http.Error(w, "Failed to execute BI query", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(result); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// BI Dashboard Handlers

// HandleCreateBIDashboard handles POST /api/v1/bi/dashboards
func (h *BIHandler) HandleCreateBIDashboard(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling create BI dashboard request")

	// Parse request body
	var request businessintelligence.BIDashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateBIDashboardRequest(&request); err != nil {
		h.logger.Error("Invalid BI dashboard request", zap.Error(err))
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

	// Create BI dashboard
	response, err := h.biService.CreateBIDashboard(ctx, &request)
	if err != nil {
		h.logger.Error("Failed to create BI dashboard", zap.Error(err))
		http.Error(w, "Failed to create BI dashboard", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetBIDashboard handles GET /api/v1/bi/dashboards/{id}
func (h *BIHandler) HandleGetBIDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]

	h.logger.Info("Handling get BI dashboard request",
		zap.String("dashboard_id", dashboardID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get BI dashboard
	dashboard, err := h.biService.GetBIDashboard(r.Context(), tenantID, dashboardID)
	if err != nil {
		h.logger.Error("Failed to get BI dashboard", zap.Error(err))
		http.Error(w, "Failed to get BI dashboard", http.StatusInternalServerError)
		return
	}

	if dashboard == nil {
		http.Error(w, "BI dashboard not found", http.StatusNotFound)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(dashboard); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleListBIDashboards handles GET /api/v1/bi/dashboards
func (h *BIHandler) HandleListBIDashboards(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling list BI dashboards request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	filter := h.parseBIFilter(r, tenantID)

	// List BI dashboards
	response, err := h.biService.ListBIDashboards(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list BI dashboards", zap.Error(err))
		http.Error(w, "Failed to list BI dashboards", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleUpdateBIDashboard handles PUT /api/v1/bi/dashboards/{id}
func (h *BIHandler) HandleUpdateBIDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]

	h.logger.Info("Handling update BI dashboard request",
		zap.String("dashboard_id", dashboardID))

	// Parse request body
	var request businessintelligence.BIDashboardRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Error("Failed to decode request body", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if err := validateBIDashboardRequest(&request); err != nil {
		h.logger.Error("Invalid BI dashboard request", zap.Error(err))
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Update BI dashboard
	response, err := h.biService.UpdateBIDashboard(r.Context(), tenantID, dashboardID, &request)
	if err != nil {
		h.logger.Error("Failed to update BI dashboard", zap.Error(err))
		http.Error(w, "Failed to update BI dashboard", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleDeleteBIDashboard handles DELETE /api/v1/bi/dashboards/{id}
func (h *BIHandler) HandleDeleteBIDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	dashboardID := vars["id"]

	h.logger.Info("Handling delete BI dashboard request",
		zap.String("dashboard_id", dashboardID))

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Delete BI dashboard
	if err := h.biService.DeleteBIDashboard(r.Context(), tenantID, dashboardID); err != nil {
		h.logger.Error("Failed to delete BI dashboard", zap.Error(err))
		http.Error(w, "Failed to delete BI dashboard", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusNoContent)
}

// Metrics Handlers

// HandleGetBIMetrics handles GET /api/v1/bi/metrics
func (h *BIHandler) HandleGetBIMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get BI metrics request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get BI metrics
	metrics, err := h.biService.GetBIMetrics(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get BI metrics", zap.Error(err))
		http.Error(w, "Failed to get BI metrics", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetDataSyncMetrics handles GET /api/v1/bi/metrics/data-syncs
func (h *BIHandler) HandleGetDataSyncMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get data sync metrics request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get data sync metrics
	metrics, err := h.biService.GetDataSyncMetrics(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get data sync metrics", zap.Error(err))
		http.Error(w, "Failed to get data sync metrics", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// HandleGetQueryPerformanceMetrics handles GET /api/v1/bi/metrics/query-performance
func (h *BIHandler) HandleGetQueryPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Handling get query performance metrics request")

	// Extract tenant ID
	tenantID := h.extractTenantID(r)
	if tenantID == "" {
		http.Error(w, "Tenant ID not found", http.StatusUnauthorized)
		return
	}

	// Get query performance metrics
	metrics, err := h.biService.GetQueryPerformanceMetrics(r.Context(), tenantID)
	if err != nil {
		h.logger.Error("Failed to get query performance metrics", zap.Error(err))
		http.Error(w, "Failed to get query performance metrics", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

// Helper functions

// validateDataSyncRequest validates a data sync request
func validateDataSyncRequest(request *businessintelligence.DataSyncRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(request.Name) > 255 {
		return fmt.Errorf("name exceeds maximum length of 255 characters")
	}

	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Validate data source type
	validTypes := map[businessintelligence.DataSourceType]bool{
		businessintelligence.DataSourceTypeRiskAssessment: true,
		businessintelligence.DataSourceTypeBatchJob:       true,
		businessintelligence.DataSourceTypeReport:         true,
		businessintelligence.DataSourceTypeDashboard:      true,
		businessintelligence.DataSourceTypeCustomModel:    true,
		businessintelligence.DataSourceTypeWebhook:        true,
		businessintelligence.DataSourceTypePerformance:    true,
	}

	if !validTypes[request.DataSourceType] {
		return fmt.Errorf("invalid data source type: %s", request.DataSourceType)
	}

	// Validate schedule frequency
	validFrequencies := map[businessintelligence.ScheduleFrequency]bool{
		businessintelligence.ScheduleFrequencyRealTime: true,
		businessintelligence.ScheduleFrequencyHourly:   true,
		businessintelligence.ScheduleFrequencyDaily:    true,
		businessintelligence.ScheduleFrequencyWeekly:   true,
		businessintelligence.ScheduleFrequencyMonthly:  true,
		businessintelligence.ScheduleFrequencyManual:   true,
	}

	if !validFrequencies[request.SyncSchedule.Frequency] {
		return fmt.Errorf("invalid schedule frequency: %s", request.SyncSchedule.Frequency)
	}

	return nil
}

// validateDataExportRequest validates a data export request
func validateDataExportRequest(request *businessintelligence.DataExportRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(request.Name) > 255 {
		return fmt.Errorf("name exceeds maximum length of 255 characters")
	}

	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Validate data source type
	validTypes := map[businessintelligence.DataSourceType]bool{
		businessintelligence.DataSourceTypeRiskAssessment: true,
		businessintelligence.DataSourceTypeBatchJob:       true,
		businessintelligence.DataSourceTypeReport:         true,
		businessintelligence.DataSourceTypeDashboard:      true,
		businessintelligence.DataSourceTypeCustomModel:    true,
		businessintelligence.DataSourceTypeWebhook:        true,
		businessintelligence.DataSourceTypePerformance:    true,
	}

	if !validTypes[request.DataSourceType] {
		return fmt.Errorf("invalid data source type: %s", request.DataSourceType)
	}

	// Validate export format
	validFormats := map[businessintelligence.DataExportFormat]bool{
		businessintelligence.DataExportFormatJSON:    true,
		businessintelligence.DataExportFormatCSV:     true,
		businessintelligence.DataExportFormatExcel:   true,
		businessintelligence.DataExportFormatParquet: true,
		businessintelligence.DataExportFormatAvro:    true,
	}

	if !validFormats[request.Format] {
		return fmt.Errorf("invalid export format: %s", request.Format)
	}

	return nil
}

// validateBIQueryRequest validates a BI query request
func validateBIQueryRequest(request *businessintelligence.BIQueryRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(request.Name) > 255 {
		return fmt.Errorf("name exceeds maximum length of 255 characters")
	}

	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Validate query definition
	if len(request.Query.DataSources) == 0 {
		return fmt.Errorf("at least one data source is required")
	}

	return nil
}

// validateBIDashboardRequest validates a BI dashboard request
func validateBIDashboardRequest(request *businessintelligence.BIDashboardRequest) error {
	if request.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(request.Name) > 255 {
		return fmt.Errorf("name exceeds maximum length of 255 characters")
	}

	if request.CreatedBy == "" {
		return fmt.Errorf("created_by is required")
	}

	// Validate layout
	if len(request.Layout.Rows) == 0 {
		return fmt.Errorf("at least one row is required in layout")
	}

	// Validate widgets
	if len(request.Widgets) == 0 {
		return fmt.Errorf("at least one widget is required")
	}

	return nil
}

// extractTenantID extracts tenant ID from request context
func (h *BIHandler) extractTenantID(r *http.Request) string {
	// This would be implemented based on your authentication/authorization system
	// For now, return a default tenant ID
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return tenantID
	}
	return "default"
}

// parseBIFilter parses query parameters into a BI filter
func (h *BIHandler) parseBIFilter(r *http.Request, tenantID string) *businessintelligence.BIFilter {
	filter := &businessintelligence.BIFilter{
		TenantID: tenantID,
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
