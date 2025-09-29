package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/supabase"
)

// MerchantHandler handles merchant management requests
type MerchantHandler struct {
	supabaseClient *supabase.Client
	logger         *zap.Logger
	config         *config.Config
}

// NewMerchantHandler creates a new merchant handler
func NewMerchantHandler(supabaseClient *supabase.Client, logger *zap.Logger, config *config.Config) *MerchantHandler {
	return &MerchantHandler{
		supabaseClient: supabaseClient,
		logger:         logger,
		config:         config,
	}
}

// Merchant represents a merchant entity
type Merchant struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	LegalName          string                 `json:"legal_name"`
	RegistrationNumber string                 `json:"registration_number,omitempty"`
	TaxID              string                 `json:"tax_id,omitempty"`
	Industry           string                 `json:"industry,omitempty"`
	IndustryCode       string                 `json:"industry_code,omitempty"`
	BusinessType       string                 `json:"business_type,omitempty"`
	FoundedDate        *time.Time             `json:"founded_date,omitempty"`
	EmployeeCount      *int                   `json:"employee_count,omitempty"`
	AnnualRevenue      *float64               `json:"annual_revenue,omitempty"`
	Address            map[string]interface{} `json:"address,omitempty"`
	ContactInfo        map[string]interface{} `json:"contact_info,omitempty"`
	PortfolioType      string                 `json:"portfolio_type"`
	RiskLevel          string                 `json:"risk_level"`
	ComplianceStatus   string                 `json:"compliance_status,omitempty"`
	Status             string                 `json:"status"`
	CreatedBy          string                 `json:"created_by,omitempty"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// CreateMerchantRequest represents a create merchant request
type CreateMerchantRequest struct {
	Name               string                 `json:"name"`
	LegalName          string                 `json:"legal_name"`
	RegistrationNumber string                 `json:"registration_number,omitempty"`
	TaxID              string                 `json:"tax_id,omitempty"`
	Industry           string                 `json:"industry,omitempty"`
	IndustryCode       string                 `json:"industry_code,omitempty"`
	BusinessType       string                 `json:"business_type,omitempty"`
	FoundedDate        *time.Time             `json:"founded_date,omitempty"`
	EmployeeCount      *int                   `json:"employee_count,omitempty"`
	AnnualRevenue      *float64               `json:"annual_revenue,omitempty"`
	Address            map[string]interface{} `json:"address,omitempty"`
	ContactInfo        map[string]interface{} `json:"contact_info,omitempty"`
	PortfolioType      string                 `json:"portfolio_type,omitempty"`
	RiskLevel          string                 `json:"risk_level,omitempty"`
	ComplianceStatus   string                 `json:"compliance_status,omitempty"`
	Status             string                 `json:"status,omitempty"`
}

// UpdateMerchantRequest represents an update merchant request
type UpdateMerchantRequest struct {
	Name               *string                 `json:"name,omitempty"`
	LegalName          *string                 `json:"legal_name,omitempty"`
	RegistrationNumber *string                 `json:"registration_number,omitempty"`
	TaxID              *string                 `json:"tax_id,omitempty"`
	Industry           *string                 `json:"industry,omitempty"`
	IndustryCode       *string                 `json:"industry_code,omitempty"`
	BusinessType       *string                 `json:"business_type,omitempty"`
	FoundedDate        *time.Time              `json:"founded_date,omitempty"`
	EmployeeCount      *int                    `json:"employee_count,omitempty"`
	AnnualRevenue      *float64                `json:"annual_revenue,omitempty"`
	Address            *map[string]interface{} `json:"address,omitempty"`
	ContactInfo        *map[string]interface{} `json:"contact_info,omitempty"`
	PortfolioType      *string                 `json:"portfolio_type,omitempty"`
	RiskLevel          *string                 `json:"risk_level,omitempty"`
	ComplianceStatus   *string                 `json:"compliance_status,omitempty"`
	Status             *string                 `json:"status,omitempty"`
}

// MerchantListResponse represents a list of merchants response
type MerchantListResponse struct {
	Merchants   []Merchant `json:"merchants"`
	Total       int        `json:"total"`
	Page        int        `json:"page"`
	PageSize    int        `json:"page_size"`
	TotalPages  int        `json:"total_pages"`
	HasNext     bool       `json:"has_next"`
	HasPrevious bool       `json:"has_previous"`
}

// HandleCreateMerchant handles POST /api/v1/merchants
func (h *MerchantHandler) HandleCreateMerchant(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var req CreateMerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" || req.LegalName == "" {
		http.Error(w, "Name and legal name are required", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.config.Merchant.RequestTimeout)
	defer cancel()

	// Create merchant
	merchant, err := h.createMerchant(ctx, &req, startTime)
	if err != nil {
		h.logger.Error("Failed to create merchant", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to create merchant: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(merchant); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Merchant created successfully",
		zap.String("merchant_id", merchant.ID),
		zap.Duration("processing_time", time.Since(startTime)))
}

// HandleGetMerchant handles GET /api/v1/merchants/{id}
func (h *MerchantHandler) HandleGetMerchant(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Extract merchant ID from path
	merchantID := h.extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		http.Error(w, "Merchant ID is required", http.StatusBadRequest)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.config.Merchant.RequestTimeout)
	defer cancel()

	// Get merchant
	merchant, err := h.getMerchant(ctx, merchantID, startTime)
	if err != nil {
		h.logger.Error("Failed to get merchant",
			zap.String("merchant_id", merchantID),
			zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get merchant: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	if err := json.NewEncoder(w).Encode(merchant); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Merchant retrieved successfully",
		zap.String("merchant_id", merchantID),
		zap.Duration("processing_time", time.Since(startTime)))
}

// HandleListMerchants handles GET /api/v1/merchants
func (h *MerchantHandler) HandleListMerchants(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > h.config.Merchant.SearchLimit {
		pageSize = h.config.Merchant.SearchLimit
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.config.Merchant.RequestTimeout)
	defer cancel()

	// List merchants
	response, err := h.listMerchants(ctx, page, pageSize, startTime)
	if err != nil {
		h.logger.Error("Failed to list merchants", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to list merchants: %v", err), http.StatusInternalServerError)
		return
	}

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Merchants listed successfully",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.Int("total", response.Total),
		zap.Duration("processing_time", time.Since(startTime)))
}

// createMerchant creates a new merchant
func (h *MerchantHandler) createMerchant(ctx context.Context, req *CreateMerchantRequest, startTime time.Time) (*Merchant, error) {
	// Generate merchant ID
	merchantID := h.generateMerchantID()

	// Set defaults
	portfolioType := req.PortfolioType
	if portfolioType == "" {
		portfolioType = "prospective"
	}

	riskLevel := req.RiskLevel
	if riskLevel == "" {
		riskLevel = "medium"
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	// Create merchant
	merchant := &Merchant{
		ID:                 merchantID,
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
		Address:            req.Address,
		ContactInfo:        req.ContactInfo,
		PortfolioType:      portfolioType,
		RiskLevel:          riskLevel,
		ComplianceStatus:   req.ComplianceStatus,
		Status:             status,
		CreatedBy:          "system", // TODO: Get from auth context
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// TODO: Save to Supabase
	// For now, return the created merchant

	return merchant, nil
}

// getMerchant retrieves a merchant by ID
func (h *MerchantHandler) getMerchant(ctx context.Context, merchantID string, startTime time.Time) (*Merchant, error) {
	// TODO: Retrieve from Supabase
	// For now, return a mock merchant

	merchant := &Merchant{
		ID:            merchantID,
		Name:          "Sample Merchant",
		LegalName:     "Sample Merchant LLC",
		PortfolioType: "prospective",
		RiskLevel:     "medium",
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return merchant, nil
}

// listMerchants lists merchants with pagination
func (h *MerchantHandler) listMerchants(ctx context.Context, page, pageSize int, startTime time.Time) (*MerchantListResponse, error) {
	// TODO: Retrieve from Supabase
	// For now, return mock data

	merchants := []Merchant{
		{
			ID:            "merchant_1",
			Name:          "Sample Merchant 1",
			LegalName:     "Sample Merchant 1 LLC",
			PortfolioType: "prospective",
			RiskLevel:     "medium",
			Status:        "active",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "merchant_2",
			Name:          "Sample Merchant 2",
			LegalName:     "Sample Merchant 2 LLC",
			PortfolioType: "active",
			RiskLevel:     "low",
			Status:        "active",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	total := len(merchants)
	totalPages := (total + pageSize - 1) / pageSize

	response := &MerchantListResponse{
		Merchants:   merchants,
		Total:       total,
		Page:        page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasNext:     page < totalPages,
		HasPrevious: page > 1,
	}

	return response, nil
}

// generateMerchantID generates a unique merchant ID
func (h *MerchantHandler) generateMerchantID() string {
	return fmt.Sprintf("merchant_%d", time.Now().UnixNano())
}

// extractMerchantIDFromPath extracts merchant ID from URL path
func (h *MerchantHandler) extractMerchantIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "merchants" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// HandleHealth handles health check requests
func (h *MerchantHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
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

	// Get merchant data
	merchantData, err := h.supabaseClient.GetMerchantData(ctx)
	if err != nil {
		h.logger.Warn("Failed to get merchant data", zap.Error(err))
	}

	// Create health response
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.1.0-ENHANCED-ENDPOINTS",
		"service":   "merchant-service",
		"uptime":    time.Since(startTime).String(),
		"supabase_status": map[string]interface{}{
			"connected": supabaseHealthy,
			"url":       h.config.Supabase.URL,
			"error":     supabaseError,
		},
		"merchant_data": merchantData,
		"features": map[string]interface{}{
			"cache_enabled":        h.config.Merchant.CacheEnabled,
			"bulk_operation_limit": h.config.Merchant.BulkOperationLimit,
			"search_limit":         h.config.Merchant.SearchLimit,
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

// HandleMerchantAnalytics handles merchant analytics requests
func (h *MerchantHandler) HandleMerchantAnalytics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// Get analytics data from Supabase
	analytics, err := h.supabaseClient.GetMerchantAnalytics(ctx)
	if err != nil {
		h.logger.Error("Failed to get merchant analytics", zap.Error(err))
		http.Error(w, "Failed to get analytics data", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"analytics":       analytics,
		"timestamp":       time.Now(),
		"processing_time": time.Since(startTime).String(),
	}

	json.NewEncoder(w).Encode(response)
}

// HandleMerchantStatistics handles merchant statistics requests
func (h *MerchantHandler) HandleMerchantStatistics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// Get statistics data from Supabase
	statistics, err := h.supabaseClient.GetMerchantStatistics(ctx)
	if err != nil {
		h.logger.Error("Failed to get merchant statistics", zap.Error(err))
		http.Error(w, "Failed to get statistics data", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"statistics":      statistics,
		"timestamp":       time.Now(),
		"processing_time": time.Since(startTime).String(),
	}

	json.NewEncoder(w).Encode(response)
}

// HandleMerchantSearch handles merchant search requests
func (h *MerchantHandler) HandleMerchantSearch(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// Parse search request
	var searchReq struct {
		Query     string `json:"query"`
		Page      int    `json:"page,omitempty"`
		PageSize  int    `json:"page_size,omitempty"`
		SortBy    string `json:"sort_by,omitempty"`
		SortOrder string `json:"sort_order,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&searchReq); err != nil {
		http.Error(w, "Invalid search request", http.StatusBadRequest)
		return
	}

	// Set defaults
	if searchReq.Page <= 0 {
		searchReq.Page = 1
	}
	if searchReq.PageSize <= 0 {
		searchReq.PageSize = 20
	}
	if searchReq.SortBy == "" {
		searchReq.SortBy = "name"
	}
	if searchReq.SortOrder == "" {
		searchReq.SortOrder = "asc"
	}

	// Perform search
	results, err := h.supabaseClient.SearchMerchants(ctx, searchReq.Query, searchReq.Page, searchReq.PageSize, searchReq.SortBy, searchReq.SortOrder)
	if err != nil {
		h.logger.Error("Failed to search merchants", zap.Error(err))
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"results":         results,
		"query":           searchReq.Query,
		"page":            searchReq.Page,
		"page_size":       searchReq.PageSize,
		"sort_by":         searchReq.SortBy,
		"sort_order":      searchReq.SortOrder,
		"timestamp":       time.Now(),
		"processing_time": time.Since(startTime).String(),
	}

	json.NewEncoder(w).Encode(response)
}

// HandleMerchantPortfolioTypes handles merchant portfolio types requests
func (h *MerchantHandler) HandleMerchantPortfolioTypes(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// Get portfolio types from Supabase
	portfolioTypes, err := h.supabaseClient.GetMerchantPortfolioTypes(ctx)
	if err != nil {
		h.logger.Error("Failed to get portfolio types", zap.Error(err))
		http.Error(w, "Failed to get portfolio types", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"portfolio_types": portfolioTypes,
		"timestamp":       time.Now(),
		"processing_time": time.Since(startTime).String(),
	}

	json.NewEncoder(w).Encode(response)
}

// HandleMerchantRiskLevels handles merchant risk levels requests
func (h *MerchantHandler) HandleMerchantRiskLevels(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// Get risk levels from Supabase
	riskLevels, err := h.supabaseClient.GetMerchantRiskLevels(ctx)
	if err != nil {
		h.logger.Error("Failed to get risk levels", zap.Error(err))
		http.Error(w, "Failed to get risk levels", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"risk_levels":     riskLevels,
		"timestamp":       time.Now(),
		"processing_time": time.Since(startTime).String(),
	}

	json.NewEncoder(w).Encode(response)
}
