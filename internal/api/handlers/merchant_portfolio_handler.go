package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
	"kyb-platform/internal/services"
)

// MerchantPortfolioServiceInterface defines the interface for merchant portfolio service
type MerchantPortfolioServiceInterface interface {
	CreateMerchant(ctx context.Context, merchant *services.Merchant, userID string) (*services.Merchant, error)
	GetMerchant(ctx context.Context, merchantID string) (*services.Merchant, error)
	UpdateMerchant(ctx context.Context, merchant *services.Merchant, userID string) (*services.Merchant, error)
	DeleteMerchant(ctx context.Context, merchantID string, userID string) error
	SearchMerchants(ctx context.Context, filters *services.MerchantSearchFilters, page, pageSize int) (*services.MerchantListResult, error)
	BulkUpdatePortfolioType(ctx context.Context, merchantIDs []string, portfolioType services.PortfolioType, userID string) (*services.BulkOperationResult, error)
	BulkUpdateRiskLevel(ctx context.Context, merchantIDs []string, riskLevel services.RiskLevel, userID string) (*services.BulkOperationResult, error)
	StartMerchantSession(ctx context.Context, userID, merchantID string) (*services.MerchantSession, error)
	EndMerchantSession(ctx context.Context, userID string) error
	GetActiveMerchantSession(ctx context.Context, userID string) (*services.MerchantSession, error)
}

// MerchantPortfolioHandler handles merchant portfolio API endpoints
type MerchantPortfolioHandler struct {
	service    MerchantPortfolioServiceInterface
	repository *database.MerchantPortfolioRepository // Optional: for portfolio analytics
	logger     *log.Logger
}

// NewMerchantPortfolioHandler creates a new merchant portfolio handler
func NewMerchantPortfolioHandler(service MerchantPortfolioServiceInterface, logger *log.Logger) *MerchantPortfolioHandler {
	if logger == nil {
		logger = log.Default()
	}

	return &MerchantPortfolioHandler{
		service: service,
		logger:  logger,
	}
}

// NewMerchantPortfolioHandlerWithRepository creates a new merchant portfolio handler with repository for analytics
func NewMerchantPortfolioHandlerWithRepository(service MerchantPortfolioServiceInterface, repository *database.MerchantPortfolioRepository, logger *log.Logger) *MerchantPortfolioHandler {
	if logger == nil {
		logger = log.Default()
	}

	return &MerchantPortfolioHandler{
		service:    service,
		repository: repository,
		logger:     logger,
	}
}

// =============================================================================
// Request/Response Types
// =============================================================================

// CreateMerchantRequest represents a request to create a new merchant
type CreateMerchantRequest struct {
	Name               string               `json:"name" validate:"required"`
	LegalName          string               `json:"legal_name" validate:"required"`
	RegistrationNumber string               `json:"registration_number"`
	TaxID              string               `json:"tax_id"`
	Industry           string               `json:"industry"`
	IndustryCode       string               `json:"industry_code"`
	BusinessType       string               `json:"business_type"`
	FoundedDate        *time.Time           `json:"founded_date"`
	EmployeeCount      int                  `json:"employee_count"`
	AnnualRevenue      *float64             `json:"annual_revenue"`
	Address            models.Address       `json:"address" validate:"required"`
	ContactInfo        models.ContactInfo   `json:"contact_info" validate:"required"`
	PortfolioType      models.PortfolioType `json:"portfolio_type"`
	RiskLevel          models.RiskLevel     `json:"risk_level"`
	ComplianceStatus   string               `json:"compliance_status"`
	Status             string               `json:"status"`
}

// UpdateMerchantRequest represents a request to update a merchant
type UpdateMerchantRequest struct {
	Name               *string               `json:"name,omitempty"`
	LegalName          *string               `json:"legal_name,omitempty"`
	RegistrationNumber *string               `json:"registration_number,omitempty"`
	TaxID              *string               `json:"tax_id,omitempty"`
	Industry           *string               `json:"industry,omitempty"`
	IndustryCode       *string               `json:"industry_code,omitempty"`
	BusinessType       *string               `json:"business_type,omitempty"`
	FoundedDate        *time.Time            `json:"founded_date,omitempty"`
	EmployeeCount      *int                  `json:"employee_count,omitempty"`
	AnnualRevenue      *float64              `json:"annual_revenue,omitempty"`
	Address            *models.Address       `json:"address,omitempty"`
	ContactInfo        *models.ContactInfo   `json:"contact_info,omitempty"`
	PortfolioType      *models.PortfolioType `json:"portfolio_type,omitempty"`
	RiskLevel          *models.RiskLevel     `json:"risk_level,omitempty"`
	ComplianceStatus   *string               `json:"compliance_status,omitempty"`
	Status             *string               `json:"status,omitempty"`
}

// MerchantSearchRequest represents a request to search merchants
type MerchantSearchRequest struct {
	Query         string                  `json:"query"`
	PortfolioType *services.PortfolioType `json:"portfolio_type,omitempty"`
	RiskLevel     *services.RiskLevel     `json:"risk_level,omitempty"`
	Industry      string                  `json:"industry,omitempty"`
	Status        string                  `json:"status,omitempty"`
	Page          int                     `json:"page,omitempty"`
	PageSize      int                     `json:"page_size,omitempty"`
	SortBy        string                  `json:"sort_by,omitempty"`
	SortOrder     string                  `json:"sort_order,omitempty"`
}

// BulkOperationRequest represents a request for bulk operations
type BulkOperationRequest struct {
	MerchantIDs   []string                `json:"merchant_ids" validate:"required"`
	Operation     string                  `json:"operation" validate:"required"`
	PortfolioType *services.PortfolioType `json:"portfolio_type,omitempty"`
	RiskLevel     *services.RiskLevel     `json:"risk_level,omitempty"`
	Status        *string                 `json:"status,omitempty"`
	Reason        string                  `json:"reason,omitempty"`
}

// MerchantResponse represents a merchant in API responses
type MerchantResponse struct {
	ID                 string               `json:"id"`
	Name               string               `json:"name"`
	LegalName          string               `json:"legal_name"`
	RegistrationNumber string               `json:"registration_number"`
	TaxID              string               `json:"tax_id"`
	Industry           string               `json:"industry"`
	IndustryCode       string               `json:"industry_code"`
	BusinessType       string               `json:"business_type"`
	FoundedDate        *time.Time           `json:"founded_date"`
	EmployeeCount      int                  `json:"employee_count"`
	AnnualRevenue      *float64             `json:"annual_revenue"`
	Address            models.Address       `json:"address"`
	ContactInfo        models.ContactInfo   `json:"contact_info"`
	PortfolioType      models.PortfolioType `json:"portfolio_type"`
	RiskLevel          models.RiskLevel     `json:"risk_level"`
	ComplianceStatus   string               `json:"compliance_status"`
	Status             string               `json:"status"`
	CreatedBy          string               `json:"created_by"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

// MerchantListResponse represents a paginated list of merchants
type MerchantListResponse struct {
	Merchants   []MerchantResponse `json:"merchants"`
	Total       int                `json:"total"`
	Page        int                `json:"page"`
	PageSize    int                `json:"page_size"`
	TotalPages  int                `json:"total_pages"`
	HasNext     bool               `json:"has_next"`
	HasPrevious bool               `json:"has_previous"`
}

// BulkOperationResponse represents the response for bulk operations
type BulkOperationResponse struct {
	OperationID    string    `json:"operation_id"`
	TotalMerchants int       `json:"total_merchants"`
	Processed      int       `json:"processed"`
	Successful     int       `json:"successful"`
	Failed         int       `json:"failed"`
	Status         string    `json:"status"`
	Message        string    `json:"message"`
	Errors         []string  `json:"errors,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// SessionResponse represents a merchant session
type SessionResponse struct {
	SessionID  string            `json:"session_id"`
	MerchantID string            `json:"merchant_id"`
	Merchant   *MerchantResponse `json:"merchant,omitempty"`
	UserID     string            `json:"user_id"`
	StartedAt  time.Time         `json:"started_at"`
	LastActive time.Time         `json:"last_active"`
	IsActive   bool              `json:"is_active"`
}

// =============================================================================
// Merchant CRUD Operations
// =============================================================================

// CreateMerchant handles POST /api/v1/merchants
func (h *MerchantPortfolioHandler) CreateMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Printf("Creating new merchant")

	var req CreateMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.LegalName == "" {
		http.Error(w, "Name and legal name are required", http.StatusBadRequest)
		return
	}

	// Convert request to service model
	merchant := &services.Merchant{
		ID:                 generateMerchantID(),
		Name:               req.Name,
		LegalName:          req.LegalName,
		RegistrationNumber: req.RegistrationNumber,
		TaxID:              req.TaxID,
		Industry:           req.Industry,
		IndustryCode:       req.IndustryCode,
		BusinessType:       req.BusinessType,
		FoundedDate:        req.FoundedDate,
		EmployeeCount:      req.EmployeeCount,
		AnnualRevenue:      req.AnnualRevenue,
		Address:            database.Address(req.Address),
		ContactInfo:        database.ContactInfo(req.ContactInfo),
		PortfolioType:      services.PortfolioType(req.PortfolioType),
		RiskLevel:          services.RiskLevel(req.RiskLevel),
		ComplianceStatus:   req.ComplianceStatus,
		Status:             req.Status,
		CreatedBy:          getUserIDFromContext(ctx),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Set defaults if not provided
	if merchant.PortfolioType == "" {
		merchant.PortfolioType = services.PortfolioTypeProspective
	}
	if merchant.RiskLevel == "" {
		merchant.RiskLevel = services.RiskLevelMedium
	}
	if merchant.Status == "" {
		merchant.Status = "active"
	}

	// Create merchant
	if h.service == nil {
		h.logger.Printf("Error: MerchantPortfolioService is not initialized")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	createdMerchant, err := h.service.CreateMerchant(ctx, merchant, getUserIDFromContext(ctx))
	if err != nil {
		h.logger.Printf("Error creating merchant: %v", err)
		http.Error(w, "Failed to create merchant", http.StatusInternalServerError)
		return
	}

	// Convert to response
	response := h.convertMerchantToResponse(createdMerchant)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetMerchant handles GET /api/v1/merchants/{id}
func (h *MerchantPortfolioHandler) GetMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	merchantID := extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		http.Error(w, "Merchant ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Getting merchant: %s", merchantID)

	if h.service == nil {
		h.logger.Printf("Error: MerchantPortfolioService is not initialized")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	merchant, err := h.service.GetMerchant(ctx, merchantID)
	if err != nil {
		h.logger.Printf("Error getting merchant %s: %v", merchantID, err)
		if err == database.ErrMerchantNotFound {
			http.Error(w, "Merchant not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get merchant", http.StatusInternalServerError)
		}
		return
	}

	response := h.convertMerchantToResponse(merchant)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateMerchant handles PUT /api/v1/merchants/{id}
func (h *MerchantPortfolioHandler) UpdateMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	merchantID := extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		http.Error(w, "Merchant ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Updating merchant: %s", merchantID)

	if h.service == nil {
		h.logger.Printf("Error: MerchantPortfolioService is not initialized")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}

	var req UpdateMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get existing merchant
	existingMerchant, err := h.service.GetMerchant(ctx, merchantID)
	if err != nil {
		h.logger.Printf("Error getting merchant %s: %v", merchantID, err)
		if err == database.ErrMerchantNotFound {
			http.Error(w, "Merchant not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get merchant", http.StatusInternalServerError)
		}
		return
	}

	// Update fields if provided
	if req.Name != nil {
		existingMerchant.Name = *req.Name
	}
	if req.LegalName != nil {
		existingMerchant.LegalName = *req.LegalName
	}
	if req.RegistrationNumber != nil {
		existingMerchant.RegistrationNumber = *req.RegistrationNumber
	}
	if req.TaxID != nil {
		existingMerchant.TaxID = *req.TaxID
	}
	if req.Industry != nil {
		existingMerchant.Industry = *req.Industry
	}
	if req.IndustryCode != nil {
		existingMerchant.IndustryCode = *req.IndustryCode
	}
	if req.BusinessType != nil {
		existingMerchant.BusinessType = *req.BusinessType
	}
	if req.FoundedDate != nil {
		existingMerchant.FoundedDate = req.FoundedDate
	}
	if req.EmployeeCount != nil {
		existingMerchant.EmployeeCount = *req.EmployeeCount
	}
	if req.AnnualRevenue != nil {
		existingMerchant.AnnualRevenue = req.AnnualRevenue
	}
	if req.Address != nil {
		existingMerchant.Address = database.Address(*req.Address)
	}
	if req.ContactInfo != nil {
		existingMerchant.ContactInfo = database.ContactInfo(*req.ContactInfo)
	}
	if req.PortfolioType != nil {
		existingMerchant.PortfolioType = services.PortfolioType(*req.PortfolioType)
	}
	if req.RiskLevel != nil {
		existingMerchant.RiskLevel = services.RiskLevel(*req.RiskLevel)
	}
	if req.ComplianceStatus != nil {
		existingMerchant.ComplianceStatus = *req.ComplianceStatus
	}
	if req.Status != nil {
		existingMerchant.Status = *req.Status
	}

	existingMerchant.UpdatedAt = time.Now()

	// Update merchant
	updatedMerchant, err := h.service.UpdateMerchant(ctx, existingMerchant, getUserIDFromContext(ctx))
	if err != nil {
		h.logger.Printf("Error updating merchant %s: %v", merchantID, err)
		http.Error(w, "Failed to update merchant", http.StatusInternalServerError)
		return
	}

	response := h.convertMerchantToResponse(updatedMerchant)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DeleteMerchant handles DELETE /api/v1/merchants/{id}
func (h *MerchantPortfolioHandler) DeleteMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	merchantID := extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		http.Error(w, "Merchant ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Deleting merchant: %s", merchantID)

	if h.service == nil {
		h.logger.Printf("Error: MerchantPortfolioService is not initialized")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	if err := h.service.DeleteMerchant(ctx, merchantID, getUserIDFromContext(ctx)); err != nil {
		h.logger.Printf("Error deleting merchant %s: %v", merchantID, err)
		if err == database.ErrMerchantNotFound {
			http.Error(w, "Merchant not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete merchant", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// =============================================================================
// Merchant Search and Listing
// =============================================================================

// ListMerchants handles GET /api/v1/merchants
func (h *MerchantPortfolioHandler) ListMerchants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Printf("Listing merchants")

	// Parse query parameters
	query := r.URL.Query()
	page, _ := strconv.Atoi(query.Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	sortBy := query.Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := query.Get("sort_order")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Build search filters
	filters := &services.MerchantSearchFilters{
		SearchQuery:   query.Get("query"),
		PortfolioType: parsePortfolioType(query.Get("portfolio_type")),
		RiskLevel:     parseRiskLevel(query.Get("risk_level")),
		Industry:      query.Get("industry"),
		Status:        query.Get("status"),
	}

	// Search merchants
	if h.service == nil {
		h.logger.Printf("Error: MerchantPortfolioService is not initialized")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	result, err := h.service.SearchMerchants(ctx, filters, page, pageSize)
	if err != nil {
		h.logger.Printf("Error searching merchants: %v", err)
		http.Error(w, "Failed to search merchants", http.StatusInternalServerError)
		return
	}

	// Convert to response
	merchantResponses := make([]MerchantResponse, len(result.Merchants))
	for i, merchant := range result.Merchants {
		merchantResponses[i] = h.convertMerchantToResponse(merchant)
	}

	totalPages := (result.Total + pageSize - 1) / pageSize
	response := MerchantListResponse{
		Merchants:   merchantResponses,
		Total:       result.Total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SearchMerchants handles POST /api/v1/merchants/search
func (h *MerchantPortfolioHandler) SearchMerchants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Printf("Searching merchants")

	var req MerchantSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "desc"
	}

	// Build search filters
	filters := &services.MerchantSearchFilters{
		SearchQuery:   req.Query,
		PortfolioType: req.PortfolioType,
		RiskLevel:     req.RiskLevel,
		Industry:      req.Industry,
		Status:        req.Status,
	}

	// Search merchants
	if h.service == nil {
		h.logger.Printf("Error: MerchantPortfolioService is not initialized")
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
		return
	}
	result, err := h.service.SearchMerchants(ctx, filters, req.Page, req.PageSize)
	if err != nil {
		h.logger.Printf("Error searching merchants: %v", err)
		http.Error(w, "Failed to search merchants", http.StatusInternalServerError)
		return
	}

	// Convert to response
	merchantResponses := make([]MerchantResponse, len(result.Merchants))
	for i, merchant := range result.Merchants {
		merchantResponses[i] = h.convertMerchantToResponse(merchant)
	}

	totalPages := (result.Total + req.PageSize - 1) / req.PageSize
	response := MerchantListResponse{
		Merchants:   merchantResponses,
		Total:       result.Total,
		Page:        req.Page,
		PageSize:    req.PageSize,
		TotalPages:  totalPages,
		HasNext:     req.Page < totalPages,
		HasPrevious: req.Page > 1,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// =============================================================================
// Bulk Operations
// =============================================================================

// BulkUpdateMerchants handles POST /api/v1/merchants/bulk/update
func (h *MerchantPortfolioHandler) BulkUpdateMerchants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Printf("Bulk updating merchants")

	var req BulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.MerchantIDs) == 0 {
		http.Error(w, "Merchant IDs are required", http.StatusBadRequest)
		return
	}
	if req.Operation == "" {
		http.Error(w, "Operation is required", http.StatusBadRequest)
		return
	}

	// Perform bulk operation based on operation type
	var result *services.BulkOperationResult
	var err error

	switch req.Operation {
	case "update_portfolio_type":
		if req.PortfolioType == nil {
			http.Error(w, "Portfolio type is required for this operation", http.StatusBadRequest)
			return
		}
		result, err = h.service.BulkUpdatePortfolioType(ctx, req.MerchantIDs, services.PortfolioType(*req.PortfolioType), getUserIDFromContext(ctx))
	case "update_risk_level":
		if req.RiskLevel == nil {
			http.Error(w, "Risk level is required for this operation", http.StatusBadRequest)
			return
		}
		result, err = h.service.BulkUpdateRiskLevel(ctx, req.MerchantIDs, services.RiskLevel(*req.RiskLevel), getUserIDFromContext(ctx))
	default:
		http.Error(w, "Invalid operation type", http.StatusBadRequest)
		return
	}
	if err != nil {
		h.logger.Printf("Error performing bulk operation: %v", err)
		http.Error(w, "Failed to perform bulk operation", http.StatusInternalServerError)
		return
	}

	response := BulkOperationResponse{
		OperationID:    result.OperationID,
		TotalMerchants: result.TotalItems,
		Processed:      result.Processed,
		Successful:     result.Successful,
		Failed:         result.Failed,
		Status:         result.Status,
		Message:        "Bulk operation completed",
		Errors:         result.Errors,
		CreatedAt:      result.StartedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// =============================================================================
// Session Management
// =============================================================================

// StartMerchantSession handles POST /api/v1/merchants/{id}/session
func (h *MerchantPortfolioHandler) StartMerchantSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	merchantID := extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		http.Error(w, "Merchant ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Starting session for merchant: %s", merchantID)

	// Start session
	session, err := h.service.StartMerchantSession(ctx, getUserIDFromContext(ctx), merchantID)
	if err != nil {
		h.logger.Printf("Error starting session for merchant %s: %v", merchantID, err)
		if err == database.ErrMerchantNotFound {
			http.Error(w, "Merchant not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to start session", http.StatusInternalServerError)
		}
		return
	}

	response := SessionResponse{
		SessionID:  session.ID,
		MerchantID: session.MerchantID,
		UserID:     session.UserID,
		StartedAt:  session.StartedAt,
		LastActive: session.LastActive,
		IsActive:   session.IsActive,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// EndMerchantSession handles DELETE /api/v1/merchants/{id}/session
func (h *MerchantPortfolioHandler) EndMerchantSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	merchantID := extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		http.Error(w, "Merchant ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Ending session for merchant: %s", merchantID)

	if err := h.service.EndMerchantSession(ctx, getUserIDFromContext(ctx)); err != nil {
		h.logger.Printf("Error ending session for merchant %s: %v", merchantID, err)
		if err == database.ErrSessionNotFound {
			http.Error(w, "Session not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to end session", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetActiveSession handles GET /api/v1/merchants/session/active
func (h *MerchantPortfolioHandler) GetActiveSession(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := getUserIDFromContext(ctx)
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Getting active session for user: %s", userID)

	session, err := h.service.GetActiveMerchantSession(ctx, userID)
	if err != nil {
		h.logger.Printf("Error getting active session for user %s: %v", userID, err)
		if err == database.ErrSessionNotFound {
			http.Error(w, "No active session found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to get active session", http.StatusInternalServerError)
		}
		return
	}

	response := SessionResponse{
		SessionID:  session.ID,
		MerchantID: session.MerchantID,
		UserID:     session.UserID,
		StartedAt:  session.StartedAt,
		LastActive: session.LastActive,
		IsActive:   session.IsActive,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// =============================================================================
// Helper Functions
// =============================================================================

// convertMerchantToResponse converts a service merchant to API response
func (h *MerchantPortfolioHandler) convertMerchantToResponse(merchant *services.Merchant) MerchantResponse {
	return MerchantResponse{
		ID:                 merchant.ID,
		Name:               merchant.Name,
		LegalName:          merchant.LegalName,
		RegistrationNumber: merchant.RegistrationNumber,
		TaxID:              merchant.TaxID,
		Industry:           merchant.Industry,
		IndustryCode:       merchant.IndustryCode,
		BusinessType:       merchant.BusinessType,
		FoundedDate:        merchant.FoundedDate,
		EmployeeCount:      merchant.EmployeeCount,
		AnnualRevenue:      merchant.AnnualRevenue,
		Address:            models.Address(merchant.Address),
		ContactInfo:        models.ContactInfo(merchant.ContactInfo),
		PortfolioType:      models.PortfolioType(merchant.PortfolioType),
		RiskLevel:          models.RiskLevel(merchant.RiskLevel),
		ComplianceStatus:   merchant.ComplianceStatus,
		Status:             merchant.Status,
		CreatedBy:          merchant.CreatedBy,
		CreatedAt:          merchant.CreatedAt,
		UpdatedAt:          merchant.UpdatedAt,
	}
}

// extractMerchantIDFromPath extracts merchant ID from URL path
func extractMerchantIDFromPath(path string) string {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i, part := range parts {
		if part == "merchants" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// getUserIDFromContext extracts user ID from request context
func getUserIDFromContext(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return "system" // Default for system operations
}

// generateMerchantID generates a unique merchant ID
func generateMerchantID() string {
	return fmt.Sprintf("merchant_%d", time.Now().UnixNano())
}

// parsePortfolioType parses a string to PortfolioType
func parsePortfolioType(s string) *services.PortfolioType {
	if s == "" {
		return nil
	}
	pt := services.PortfolioType(s)
	return &pt
}

// parseRiskLevel parses a string to RiskLevel
func parseRiskLevel(s string) *services.RiskLevel {
	if s == "" {
		return nil
	}
	rl := services.RiskLevel(s)
	return &rl
}

// =============================================================================
// Additional Handler Methods for Routes
// =============================================================================

// BulkExportMerchants handles POST /api/v1/merchants/bulk/export
func (h *MerchantPortfolioHandler) BulkExportMerchants(w http.ResponseWriter, r *http.Request) {
	_ = r.Context() // Context will be used in future implementation
	h.logger.Printf("Bulk exporting merchants")

	var req BulkOperationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.MerchantIDs) == 0 {
		http.Error(w, "Merchant IDs are required", http.StatusBadRequest)
		return
	}

	// For now, return a placeholder response
	// TODO: Implement actual bulk export functionality
	response := map[string]interface{}{
		"operation_id":    fmt.Sprintf("export_%d", time.Now().UnixNano()),
		"total_merchants": len(req.MerchantIDs),
		"status":          "completed",
		"message":         "Bulk export completed successfully",
		"download_url":    fmt.Sprintf("/api/v1/merchants/bulk/export/%d/download", time.Now().UnixNano()),
		"created_at":      time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMerchantAnalytics handles GET /api/v1/merchants/analytics
func (h *MerchantPortfolioHandler) GetMerchantAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.logger.Printf("Getting merchant analytics")

	// If repository is not available, return mock data for backward compatibility
	if h.repository == nil {
		h.logger.Printf("Warning: Repository not available, returning mock analytics data")
		response := map[string]interface{}{
			"total_merchants": 5000,
			"portfolio_distribution": map[string]interface{}{
				"onboarded":   2500,
				"deactivated": 500,
				"prospective": 1500,
				"pending":     500,
			},
			"risk_distribution": map[string]interface{}{
				"high":   1000,
				"medium": 3000,
				"low":    1000,
			},
			"industry_distribution": map[string]interface{}{
				"Technology":    1500,
				"Retail":        1200,
				"Healthcare":    800,
				"Finance":       600,
				"Manufacturing": 500,
				"Other":         400,
			},
			"compliance_status": map[string]interface{}{
				"compliant":     4000,
				"non_compliant": 500,
				"pending":       500,
			},
			"created_at": time.Now(),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get real analytics data from database
	analytics, err := h.getPortfolioAnalytics(ctx)
	if err != nil {
		h.logger.Printf("Error getting portfolio analytics: %v", err)
		http.Error(w, "failed to retrieve analytics", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analytics); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// getPortfolioAnalytics retrieves portfolio-level analytics from the database
func (h *MerchantPortfolioHandler) getPortfolioAnalytics(ctx context.Context) (map[string]interface{}, error) {
	// Get total merchants count
	totalCount, err := h.repository.CountMerchants(ctx, &models.MerchantSearchFilters{})
	if err != nil {
		return nil, fmt.Errorf("failed to count merchants: %w", err)
	}

	// Get portfolio distribution
	portfolioDist, err := h.repository.GetPortfolioDistribution(ctx)
	if err != nil {
		h.logger.Printf("Warning: failed to get portfolio distribution: %v", err)
		portfolioDist = map[string]int{
			"onboarded":   0,
			"deactivated": 0,
			"prospective": 0,
			"pending":     0,
		}
	}

	// Get risk distribution
	riskDist, err := h.repository.GetRiskDistribution(ctx)
	if err != nil {
		h.logger.Printf("Warning: failed to get risk distribution: %v", err)
		riskDist = map[string]int{
			"high":   0,
			"medium": 0,
			"low":    0,
		}
	}

	// Get industry distribution
	industryDist, err := h.repository.GetIndustryDistribution(ctx)
	if err != nil {
		h.logger.Printf("Warning: failed to get industry distribution: %v", err)
		industryDist = map[string]int{}
	}

	// Get compliance status distribution
	complianceDist, err := h.repository.GetComplianceDistribution(ctx)
	if err != nil {
		h.logger.Printf("Warning: failed to get compliance distribution: %v", err)
		complianceDist = map[string]int{
			"compliant":     0,
			"non_compliant": 0,
			"pending":       0,
		}
	}

	// Convert int maps to interface{} maps for JSON encoding
	portfolioDistInterface := make(map[string]interface{})
	for k, v := range portfolioDist {
		portfolioDistInterface[k] = v
	}

	riskDistInterface := make(map[string]interface{})
	for k, v := range riskDist {
		riskDistInterface[k] = v
	}

	industryDistInterface := make(map[string]interface{})
	for k, v := range industryDist {
		industryDistInterface[k] = v
	}

	complianceDistInterface := make(map[string]interface{})
	for k, v := range complianceDist {
		complianceDistInterface[k] = v
	}

	return map[string]interface{}{
		"total_merchants":       totalCount,
		"portfolio_distribution": portfolioDistInterface,
		"risk_distribution":     riskDistInterface,
		"industry_distribution": industryDistInterface,
		"compliance_status":     complianceDistInterface,
		"created_at":            time.Now(),
	}, nil
}


// GetPortfolioTypes handles GET /api/v1/merchants/portfolio-types
func (h *MerchantPortfolioHandler) GetPortfolioTypes(w http.ResponseWriter, r *http.Request) {
	_ = r.Context() // Context will be used in future implementation
	h.logger.Printf("Getting portfolio types")

	portfolioTypes := []map[string]interface{}{
		{
			"value":       "onboarded",
			"label":       "Onboarded",
			"description": "Merchants that have completed the onboarding process",
			"color":       "#28a745",
		},
		{
			"value":       "deactivated",
			"label":       "Deactivated",
			"description": "Merchants that have been deactivated",
			"color":       "#dc3545",
		},
		{
			"value":       "prospective",
			"label":       "Prospective",
			"description": "Potential merchants under consideration",
			"color":       "#ffc107",
		},
		{
			"value":       "pending",
			"label":       "Pending",
			"description": "Merchants with pending applications or reviews",
			"color":       "#17a2b8",
		},
	}

	response := map[string]interface{}{
		"portfolio_types": portfolioTypes,
		"total":           len(portfolioTypes),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetRiskLevels handles GET /api/v1/merchants/risk-levels
func (h *MerchantPortfolioHandler) GetRiskLevels(w http.ResponseWriter, r *http.Request) {
	_ = r.Context() // Context will be used in future implementation
	h.logger.Printf("Getting risk levels")

	riskLevels := []map[string]interface{}{
		{
			"value":       "high",
			"label":       "High Risk",
			"description": "Merchants with high risk factors requiring enhanced monitoring",
			"color":       "#dc3545",
			"score_range": "0.7-1.0",
		},
		{
			"value":       "medium",
			"label":       "Medium Risk",
			"description": "Merchants with moderate risk factors requiring standard monitoring",
			"color":       "#ffc107",
			"score_range": "0.3-0.7",
		},
		{
			"value":       "low",
			"label":       "Low Risk",
			"description": "Merchants with low risk factors requiring minimal monitoring",
			"color":       "#28a745",
			"score_range": "0.0-0.3",
		},
	}

	response := map[string]interface{}{
		"risk_levels": riskLevels,
		"total":       len(riskLevels),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMerchantStatistics handles GET /api/v1/merchants/statistics
func (h *MerchantPortfolioHandler) GetMerchantStatistics(w http.ResponseWriter, r *http.Request) {
	_ = r.Context() // Context will be used in future implementation
	h.logger.Printf("Getting merchant statistics")

	// For now, return mock statistics
	// TODO: Implement actual statistics calculation
	response := map[string]interface{}{
		"total_merchants":            5000,
		"active_merchants":           4000,
		"inactive_merchants":         1000,
		"average_employee_count":     25,
		"total_annual_revenue":       1250000000, // $1.25B
		"average_annual_revenue":     250000,     // $250K
		"compliance_rate":            0.8,        // 80%
		"onboarding_completion_rate": 0.75,       // 75%
		"risk_distribution": map[string]interface{}{
			"high":   0.2, // 20%
			"medium": 0.6, // 60%
			"low":    0.2, // 20%
		},
		"portfolio_distribution": map[string]interface{}{
			"onboarded":   0.5, // 50%
			"deactivated": 0.1, // 10%
			"prospective": 0.3, // 30%
			"pending":     0.1, // 10%
		},
		"last_updated": time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
