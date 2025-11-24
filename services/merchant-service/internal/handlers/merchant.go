package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	
	postgrest "github.com/supabase-community/postgrest-go"
	"go.uber.org/zap"

	"kyb-platform/services/merchant-service/internal/errors"
	"kyb-platform/services/merchant-service/internal/cache"
	"kyb-platform/services/merchant-service/internal/metrics"
	"kyb-platform/services/merchant-service/internal/queue"
	"kyb-platform/services/merchant-service/internal/resilience"
	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/supabase"
	"kyb-platform/services/merchant-service/internal/jobs"
)

// MerchantHandler handles merchant management requests
type MerchantHandler struct {
	supabaseClient  *supabase.Client
	logger          *zap.Logger
	config          *config.Config
	circuitBreaker  *resilience.CircuitBreaker
	redisCache      *cache.RedisCacheImpl // Redis cache for merchant data
	fallbackMetrics *metrics.FallbackMetrics // Metrics for tracking fallback usage
	requestQueue    *queue.RequestQueue // Queue for failed requests
	jobProcessor    *jobs.JobProcessor // Job processor for async analysis jobs
}

// NewMerchantHandler creates a new merchant handler
func NewMerchantHandler(supabaseClient *supabase.Client, logger *zap.Logger, config *config.Config, jobProcessor *jobs.JobProcessor) *MerchantHandler {
	// Create circuit breaker for Supabase connection
	cbConfig := resilience.DefaultCircuitBreakerConfig()
	cbConfig.FailureThreshold = 5
	cbConfig.Timeout = 30 * time.Second
	circuitBreaker := resilience.NewCircuitBreaker(cbConfig)
	
	// Initialize Redis cache if enabled
	var redisCache *cache.RedisCacheImpl
	if config.Merchant.RedisEnabled && config.Merchant.RedisURL != "" {
		cacheConfig := &cache.RedisCacheConfig{
			Addr:     config.Merchant.RedisURL,
			Password: "",
			DB:       0,
			TTL:      config.Merchant.CacheTTL,
			PoolSize: 10,
		}
		var err error
		redisCache, err = cache.NewRedisCache(cacheConfig, logger)
		if err != nil {
			logger.Warn("Failed to initialize Redis cache, continuing without cache",
				zap.Error(err))
		}
	}
	
	// Initialize fallback metrics
	fallbackMetrics := metrics.NewFallbackMetrics(logger)
	
	// Initialize request queue for failed API calls (Phase 2.5: Request Queuing)
	requestQueue := queue.NewRequestQueue(logger, 1000) // Max 1000 queued requests
	
	return &MerchantHandler{
		supabaseClient:  supabaseClient,
		logger:          logger,
		config:          config,
		circuitBreaker:  circuitBreaker,
		redisCache:      redisCache,
		fallbackMetrics: fallbackMetrics,
		requestQueue:    requestQueue,
		jobProcessor:    jobProcessor,
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
		errors.WriteBadRequest(w, r, "Invalid request body: Please provide valid JSON")
		return
	}

	// Validate required fields
	if req.Name == "" || req.LegalName == "" {
		errors.WriteBadRequest(w, r, "Name and legal name are required")
		return
	}

	// Sanitize input to prevent XSS and injection attacks
	req.Name = sanitizeInput(req.Name)
	req.LegalName = sanitizeInput(req.LegalName)
	if req.RegistrationNumber != "" {
		req.RegistrationNumber = sanitizeInput(req.RegistrationNumber)
	}
	if req.TaxID != "" {
		req.TaxID = sanitizeInput(req.TaxID)
	}
	if req.Industry != "" {
		req.Industry = sanitizeInput(req.Industry)
	}
	if req.IndustryCode != "" {
		req.IndustryCode = sanitizeInput(req.IndustryCode)
	}
	if req.BusinessType != "" {
		req.BusinessType = sanitizeInput(req.BusinessType)
	}
	if req.PortfolioType != "" {
		req.PortfolioType = sanitizeInput(req.PortfolioType)
	}
	if req.RiskLevel != "" {
		req.RiskLevel = sanitizeInput(req.RiskLevel)
	}
	if req.ComplianceStatus != "" {
		req.ComplianceStatus = sanitizeInput(req.ComplianceStatus)
	}
	if req.Status != "" {
		req.Status = sanitizeInput(req.Status)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.config.Merchant.RequestTimeout)
	defer cancel()

	// Extract user ID from request (headers or context)
	userID := h.getUserIDFromRequest(r)

	// Create merchant
	merchant, err := h.createMerchant(ctx, &req, startTime, userID)
	if err != nil {
		h.logger.Error("Failed to create merchant", zap.Error(err))
		errors.WriteInternalError(w, r, fmt.Sprintf("Failed to create merchant: %v", err))
		return
	}

	// Send response
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(merchant); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		errors.WriteInternalError(w, r, "Failed to encode response")
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
		errors.WriteBadRequest(w, r, "Merchant ID is required")
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
		
		// Determine appropriate HTTP status code based on error
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			errors.WriteNotFound(w, r, "Merchant not found")
			return
		} else if strings.Contains(err.Error(), "unavailable") {
			// Database unavailable - return 503 Service Unavailable
			w.Header().Set("Retry-After", "30") // Suggest retry after 30 seconds
			errors.WriteServiceUnavailable(w, r, "Service temporarily unavailable")
			return
		}
		
		errors.WriteInternalError(w, r, fmt.Sprintf("Failed to get merchant: %v", err))
		return
	}

	// Send response
	if err := json.NewEncoder(w).Encode(merchant); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		errors.WriteInternalError(w, r, "Failed to encode response")
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

	// Parse filtering parameters
	filters := MerchantFilters{
		PortfolioType: r.URL.Query().Get("portfolio_type"),
		RiskLevel:     r.URL.Query().Get("risk_level"),
		Status:        r.URL.Query().Get("status"),
		SearchQuery:   r.URL.Query().Get("search"),
	}

	// Parse sorting parameters
	sortBy := r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at" // Default sort by creation date
	}
	sortOrder := r.URL.Query().Get("sort_order")
	if sortOrder == "" {
		sortOrder = "desc" // Default to descending (newest first)
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc" // Validate sort order
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), h.config.Merchant.RequestTimeout)
	defer cancel()

	// List merchants with filters and sorting
	response, err := h.listMerchants(ctx, page, pageSize, filters, sortBy, sortOrder, startTime)
	if err != nil {
		h.logger.Error("Failed to list merchants", zap.Error(err))
		errors.WriteInternalError(w, r, fmt.Sprintf("Failed to list merchants: %v", err))
		return
	}

	// Send response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		errors.WriteInternalError(w, r, "Failed to encode response")
		return
	}

	h.logger.Info("Merchants listed successfully",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.Int("total", response.Total),
		zap.Duration("processing_time", time.Since(startTime)))
}

// createMerchant creates a new merchant
func (h *MerchantHandler) createMerchant(ctx context.Context, req *CreateMerchantRequest, startTime time.Time, userID string) (*Merchant, error) {
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
		CreatedBy:          userID,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Save to Supabase
	// Build merchantData map, only including non-empty values
	merchantData := map[string]interface{}{
		"id":                merchant.ID,
		"name":              merchant.Name,
		"created_at":        merchant.CreatedAt.Format(time.RFC3339),
		"updated_at":        merchant.UpdatedAt.Format(time.RFC3339),
	}
	
	// Add optional fields only if they have values
	if merchant.LegalName != "" {
		merchantData["legal_name"] = merchant.LegalName
	}
	if merchant.RegistrationNumber != "" {
		merchantData["registration_number"] = merchant.RegistrationNumber
	}
	if merchant.TaxID != "" {
		merchantData["tax_id"] = merchant.TaxID
	}
	if merchant.Industry != "" {
		merchantData["industry"] = merchant.Industry
	}
	if merchant.IndustryCode != "" {
		merchantData["industry_code"] = merchant.IndustryCode
	}
	if merchant.BusinessType != "" {
		merchantData["business_type"] = merchant.BusinessType
	}
	if merchant.FoundedDate != nil {
		merchantData["founded_date"] = merchant.FoundedDate.Format("2006-01-02")
	}
	if merchant.EmployeeCount != nil {
		merchantData["employee_count"] = *merchant.EmployeeCount
	}
	if merchant.AnnualRevenue != nil {
		merchantData["annual_revenue"] = *merchant.AnnualRevenue
	}
	// Address and ContactInfo as JSONB - include even if empty map (will be stored as empty JSON object)
	if merchant.Address != nil && len(merchant.Address) > 0 {
		merchantData["address"] = merchant.Address
	} else {
		// Set to empty JSON object if nil
		merchantData["address"] = map[string]interface{}{}
	}
	if merchant.ContactInfo != nil && len(merchant.ContactInfo) > 0 {
		merchantData["contact_info"] = merchant.ContactInfo
	} else {
		// Set to empty JSON object if nil
		merchantData["contact_info"] = map[string]interface{}{}
	}
	if merchant.PortfolioType != "" {
		merchantData["portfolio_type"] = merchant.PortfolioType
	}
	if merchant.RiskLevel != "" {
		merchantData["risk_level"] = merchant.RiskLevel
	}
	if merchant.ComplianceStatus != "" {
		merchantData["compliance_status"] = merchant.ComplianceStatus
	}
	if merchant.Status != "" {
		merchantData["status"] = merchant.Status
	}
	if merchant.CreatedBy != "" {
		merchantData["created_by"] = merchant.CreatedBy
	}

	// Save to Supabase with retry logic and circuit breaker
	err := h.circuitBreaker.Execute(ctx, func() error {
		// Use retry logic for the Supabase insert
		retryConfig := resilience.DefaultRetryConfig()
		retryConfig.MaxAttempts = 3
		retryConfig.InitialDelay = 100 * time.Millisecond
		
		_, retryErr := resilience.RetryWithBackoff(ctx, retryConfig, func() ([]map[string]interface{}, error) {
			var queryResult []map[string]interface{}
			_, queryErr := h.supabaseClient.GetClient().From("merchants").
				Insert(merchantData, false, "", "", "").
				ExecuteTo(&queryResult)
			
			if queryErr != nil {
				return nil, queryErr
			}
			
			return queryResult, nil
		})
		
		if retryErr != nil {
			return retryErr
		}
		
		return nil
	})

	if err != nil {
		h.logger.Error("Failed to save merchant to Supabase",
			zap.String("merchant_id", merchant.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to save merchant to database: %w", err)
	}

	h.logger.Info("Merchant saved to Supabase successfully",
		zap.String("merchant_id", merchant.ID),
		zap.String("name", merchant.Name))

	// Trigger async classification job (non-blocking)
	go h.triggerClassificationJob(ctx, merchant)

	// Trigger website analysis job if website URL is provided (non-blocking)
	if merchant.ContactInfo != nil {
		if websiteURL, ok := merchant.ContactInfo["website"].(string); ok && websiteURL != "" {
			go h.triggerWebsiteAnalysisJob(ctx, merchant.ID, websiteURL, merchant.Name)
		} else {
			// Mark website analysis as skipped
			go h.markWebsiteAnalysisSkipped(ctx, merchant.ID)
		}
	} else {
		// Mark website analysis as skipped
		go h.markWebsiteAnalysisSkipped(ctx, merchant.ID)
	}

	return merchant, nil
}

// triggerClassificationJob triggers an async classification job for a merchant
func (h *MerchantHandler) triggerClassificationJob(ctx context.Context, merchant *Merchant) {
	if h.jobProcessor == nil {
		h.logger.Warn("Job processor not initialized, skipping classification job")
		return
	}

	// Extract website URL from contact info
	websiteURL := ""
	if merchant.ContactInfo != nil {
		if url, ok := merchant.ContactInfo["website"].(string); ok {
			websiteURL = url
		}
	}

	// Extract description from industry or business type
	description := merchant.Industry
	if description == "" {
		description = merchant.BusinessType
	}

	// Create classification job
	job := jobs.NewClassificationJob(
		merchant.ID,
		merchant.Name,
		description,
		websiteURL,
		h.supabaseClient,
		h.config,
		h.logger,
	)

	// Enqueue job (non-blocking)
	if err := h.jobProcessor.Enqueue(job); err != nil {
		h.logger.Error("Failed to enqueue classification job",
			zap.String("merchant_id", merchant.ID),
			zap.Error(err))
	} else {
		h.logger.Info("Classification job enqueued",
			zap.String("merchant_id", merchant.ID),
			zap.String("job_id", job.GetID()))
	}
}

// triggerWebsiteAnalysisJob triggers an async website analysis job for a merchant
func (h *MerchantHandler) triggerWebsiteAnalysisJob(ctx context.Context, merchantID, websiteURL, businessName string) {
	if h.jobProcessor == nil {
		h.logger.Warn("Job processor not initialized, skipping website analysis job")
		return
	}

	// Create website analysis job
	job := jobs.NewWebsiteAnalysisJob(
		merchantID,
		websiteURL,
		businessName,
		h.supabaseClient,
		h.config,
		h.logger,
	)

	// Enqueue job (non-blocking)
	if err := h.jobProcessor.Enqueue(job); err != nil {
		h.logger.Error("Failed to enqueue website analysis job",
			zap.String("merchant_id", merchantID),
			zap.Error(err))
	} else {
		h.logger.Info("Website analysis job enqueued",
			zap.String("merchant_id", merchantID),
			zap.String("job_id", job.GetID()),
			zap.String("website_url", websiteURL))
	}
}

// markWebsiteAnalysisSkipped marks website analysis as skipped in the database
func (h *MerchantHandler) markWebsiteAnalysisSkipped(ctx context.Context, merchantID string) {
	updateData := map[string]interface{}{
		"website_analysis_status":      "skipped",
		"website_analysis_updated_at": time.Now().Format(time.RFC3339),
	}

	// Check if merchant_analytics record exists
	var existing []map[string]interface{}
	_, err := h.supabaseClient.GetClient().From("merchant_analytics").
		Select("id", "", false).
		Eq("merchant_id", merchantID).
		Limit(1, "").
		ExecuteTo(&existing)

	if err != nil || len(existing) == 0 {
		// Create new record
		insertData := map[string]interface{}{
			"merchant_id":                merchantID,
			"website_analysis_status":    "skipped",
			"website_analysis_updated_at": time.Now().Format(time.RFC3339),
			"website_analysis_data":      map[string]interface{}{},
		}

		_, _, err := h.supabaseClient.GetClient().From("merchant_analytics").
			Insert(insertData, false, "", "", "").
			Execute()

		if err != nil {
			h.logger.Warn("Failed to mark website analysis as skipped",
				zap.String("merchant_id", merchantID),
				zap.Error(err))
		}
		return
	}

	// Update existing record
	_, _, err = h.supabaseClient.GetClient().From("merchant_analytics").
		Update(updateData, "", "").
		Eq("merchant_id", merchantID).
		Execute()

	if err != nil {
		h.logger.Warn("Failed to mark website analysis as skipped",
			zap.String("merchant_id", merchantID),
			zap.Error(err))
	}
}

// migrateWebsiteURLToContactInfo migrates a website URL from legacy columns to contact_info
func (h *MerchantHandler) migrateWebsiteURLToContactInfo(ctx context.Context, merchantID, websiteURL string) {
	// Get current contact_info
	var merchantResult []map[string]interface{}
	_, err := h.supabaseClient.GetClient().From("merchants").
		Select("contact_info", "", false).
		Eq("id", merchantID).
		Limit(1, "").
		ExecuteTo(&merchantResult)
	
	if err != nil || len(merchantResult) == 0 {
		h.logger.Warn("Failed to fetch merchant for website URL migration",
			zap.String("merchant_id", merchantID),
			zap.Error(err))
		return
	}
	
	// Build updated contact_info
	contactInfo := make(map[string]interface{})
	if existingContactInfo, ok := merchantResult[0]["contact_info"].(map[string]interface{}); ok {
		contactInfo = existingContactInfo
	}
	contactInfo["website"] = websiteURL
	
	// Update merchant with migrated website URL
	updateData := map[string]interface{}{
		"contact_info": contactInfo,
	}
	
	_, _, err = h.supabaseClient.GetClient().From("merchants").
		Update(updateData, "", "").
		Eq("id", merchantID).
		Execute()
	
	if err != nil {
		h.logger.Warn("Failed to migrate website URL to contact_info",
			zap.String("merchant_id", merchantID),
			zap.String("website_url", websiteURL),
			zap.Error(err))
	} else {
		h.logger.Info("Successfully migrated website URL to contact_info",
			zap.String("merchant_id", merchantID),
			zap.String("website_url", websiteURL))
	}
}

// getMerchant retrieves a merchant by ID from Supabase.
//
// FALLBACK BEHAVIOR:
//   - If Supabase query fails (connection error, timeout, etc.), returns mock merchant data
//   - If merchant is not found in database, returns mock merchant data
//   - If data mapping fails, returns mock merchant data
//
// The fallback ensures UI functionality continues even when Supabase is unavailable.
// In production, consider returning proper HTTP 404/503 status codes instead of mock data.
//
	// Retry logic with exponential backoff and circuit breaker are already implemented below
func (h *MerchantHandler) getMerchant(ctx context.Context, merchantID string, startTime time.Time) (*Merchant, error) {
	h.logger.Info("Fetching merchant from Supabase",
		zap.String("merchant_id", merchantID))

	// Check cache first if Redis is enabled
	if h.redisCache != nil {
		cacheKey := fmt.Sprintf("merchant:%s", merchantID)
		cachedData, err := h.redisCache.Get(ctx, cacheKey)
		if err == nil && cachedData != nil {
			// Deserialize cached merchant
			var cachedMerchant Merchant
			if err := json.Unmarshal(cachedData, &cachedMerchant); err == nil {
				h.logger.Info("Retrieved merchant from cache",
					zap.String("merchant_id", merchantID))
				return &cachedMerchant, nil
			}
		}
	}

	// Try to get merchant from Supabase with retry logic and circuit breaker
	var result []map[string]interface{}
	err := h.circuitBreaker.Execute(ctx, func() error {
		// Use retry logic for the Supabase query
		retryConfig := resilience.DefaultRetryConfig()
		retryConfig.MaxAttempts = 3
		retryConfig.InitialDelay = 100 * time.Millisecond
		
		retryResult, retryErr := resilience.RetryWithBackoff(ctx, retryConfig, func() ([]map[string]interface{}, error) {
			var queryResult []map[string]interface{}
			_, queryErr := h.supabaseClient.GetClient().From("merchants").
				Select("*", "", false).
				Eq("id", merchantID).
				Limit(1, "").
				ExecuteTo(&queryResult)
			
			if queryErr != nil {
				return nil, queryErr
			}
			
			return queryResult, nil
		})
		
		if retryErr != nil {
			return retryErr
		}
		
		result = retryResult
		return nil
	})

	if err != nil {
		h.logger.Warn("Failed to fetch merchant from Supabase",
			zap.String("merchant_id", merchantID),
			zap.String("environment", h.config.Environment),
			zap.Error(err))
		
		// Phase 2.5: Queue failed request for retry
		if h.requestQueue != nil {
			queuedReq := &queue.QueuedRequest{
				Type:        "get_merchant",
				Data:        merchantID,
				Priority:    queue.PriorityNormal,
				MaxAttempts: 3,
				Error:       err.Error(),
			}
			if queueErr := h.requestQueue.Enqueue(ctx, queuedReq); queueErr != nil {
				h.logger.Warn("Failed to queue request for retry",
					zap.String("merchant_id", merchantID),
					zap.Error(queueErr))
			} else {
				h.logger.Info("Request queued for retry",
					zap.String("merchant_id", merchantID),
					zap.String("request_id", queuedReq.ID))
			}
		}
		
		// In production, return 503 Service Unavailable instead of mock data
		if h.config.Environment == "production" && !h.config.Merchant.AllowMockData {
			return nil, fmt.Errorf("database unavailable: %w", err)
		}
		
		// FALLBACK: Return mock data when Supabase query fails (development only)
		// This ensures the UI continues to function even when database is unavailable
		fallbackStart := time.Now()
		mockMerchant, mockErr := h.getMockMerchant(merchantID)
		if mockErr != nil {
			return nil, mockErr
		}
		
		// Record fallback usage metrics
		if h.fallbackMetrics != nil {
			h.fallbackMetrics.RecordFallbackUsage(ctx, "merchant-service", "database_fallback", "supabase", time.Since(fallbackStart))
		}
		
		return mockMerchant, nil
	}

	if len(result) == 0 {
		h.logger.Warn("Merchant not found in Supabase",
			zap.String("merchant_id", merchantID),
			zap.String("environment", h.config.Environment))
		
		// Always return "not found" error for non-existent merchants
		// This ensures proper 404 responses instead of returning mock data
		// Mock data should only be used for database connection failures, not missing records
		return nil, fmt.Errorf("merchant not found: %s", merchantID)
	}

	// Convert Supabase result to Merchant struct
	merchant, err := h.mapToMerchant(result[0])
	if err != nil {
		h.logger.Error("Failed to map Supabase data to merchant",
			zap.String("merchant_id", merchantID),
			zap.String("environment", h.config.Environment),
			zap.Error(err))
		
		// In production, return error instead of mock data
		if h.config.Environment == "production" && !h.config.Merchant.AllowMockData {
			return nil, fmt.Errorf("data mapping failed: %w", err)
		}
		
		// FALLBACK: Return mock data when data mapping fails (development only)
		fallbackStart := time.Now()
		mockMerchant, mockErr := h.getMockMerchant(merchantID)
		if mockErr != nil {
			return nil, mockErr
		}
		
		// Record fallback usage metrics
		if h.fallbackMetrics != nil {
			h.fallbackMetrics.RecordFallbackUsage(ctx, "merchant-service", "database_fallback", "supabase", time.Since(fallbackStart))
		}
		
		return mockMerchant, nil
	}

	h.logger.Info("Successfully fetched merchant from Supabase",
		zap.String("merchant_id", merchantID),
		zap.String("name", merchant.Name))

	// Record successful request (non-fallback)
	if h.fallbackMetrics != nil {
		h.fallbackMetrics.RecordRequest(ctx, "merchant-service")
	}

	// Cache the result if Redis is enabled
	if h.redisCache != nil {
		cacheKey := fmt.Sprintf("merchant:%s", merchantID)
		merchantJSON, err := json.Marshal(merchant)
		if err == nil {
			if err := h.redisCache.Set(ctx, cacheKey, merchantJSON, h.config.Merchant.CacheTTL); err != nil {
				h.logger.Warn("Failed to cache merchant",
					zap.String("merchant_id", merchantID),
					zap.Error(err))
			}
		}
	}

	return merchant, nil
}

// getMockMerchant returns a mock merchant for fallback scenarios.
//
// This function is used when:
//   - Supabase query fails (connection error, timeout)
//   - Merchant not found in database
//   - Data mapping fails
//
// The mock data ensures UI functionality continues, but should be replaced with
// proper error handling in production (e.g., return 404/503 HTTP status codes).
//
// FALLBACK DATA - DO NOT USE AS PRIMARY DATA SOURCE
//
// PRODUCTION SAFETY: In production, mock data is only returned if explicitly allowed
// via ALLOW_MOCK_DATA environment variable. Otherwise, this function should not be called.
func (h *MerchantHandler) getMockMerchant(merchantID string) (*Merchant, error) {
	// Production safety check: prevent mock data in production unless explicitly allowed
	if h.config.Environment == "production" && !h.config.Merchant.AllowMockData {
		return nil, fmt.Errorf("mock data not allowed in production environment")
	}
	
	return &Merchant{
		ID:            merchantID,
		Name:          "Sample Merchant",
		LegalName:     "Sample Merchant LLC",
		PortfolioType: "prospective",
		RiskLevel:     "medium",
		Status:        "active",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}, nil
}

// mapToMerchant converts a map from Supabase to a Merchant struct
func (h *MerchantHandler) mapToMerchant(data map[string]interface{}) (*Merchant, error) {
	merchant := &Merchant{}

	// Extract ID
	if id, ok := data["id"].(string); ok {
		merchant.ID = id
	}

	// Extract Name (try multiple field names)
	if name, ok := data["name"].(string); ok {
		merchant.Name = name
	} else if name, ok := data["business_name"].(string); ok {
		merchant.Name = name
	}

	// Extract LegalName
	if legalName, ok := data["legal_name"].(string); ok {
		merchant.LegalName = legalName
	}

	// Extract RegistrationNumber
	if regNum, ok := data["registration_number"].(string); ok {
		merchant.RegistrationNumber = regNum
	}

	// Extract TaxID
	if taxID, ok := data["tax_id"].(string); ok {
		merchant.TaxID = taxID
	}

	// Extract Industry
	if industry, ok := data["industry"].(string); ok {
		merchant.Industry = industry
	}

	// Extract IndustryCode
	if industryCode, ok := data["industry_code"].(string); ok {
		merchant.IndustryCode = industryCode
	}

	// Extract BusinessType
	if businessType, ok := data["business_type"].(string); ok {
		merchant.BusinessType = businessType
	}

	// Extract PortfolioType
	if portfolioType, ok := data["portfolio_type"].(string); ok {
		merchant.PortfolioType = portfolioType
	} else {
		merchant.PortfolioType = "prospective"
	}

	// Extract RiskLevel
	if riskLevel, ok := data["risk_level"].(string); ok {
		merchant.RiskLevel = riskLevel
	} else {
		merchant.RiskLevel = "medium"
	}

	// Extract Status
	if status, ok := data["status"].(string); ok {
		merchant.Status = status
	} else {
		merchant.Status = "active"
	}

	// Extract ComplianceStatus
	if complianceStatus, ok := data["compliance_status"].(string); ok {
		merchant.ComplianceStatus = complianceStatus
	}

	// Extract Address (can be JSON object)
	if address, ok := data["address"].(map[string]interface{}); ok {
		merchant.Address = address
	} else if address, ok := data["address"].(string); ok {
		// Try to parse as JSON string
		var addrMap map[string]interface{}
		if err := json.Unmarshal([]byte(address), &addrMap); err == nil {
			merchant.Address = addrMap
		}
	}

	// Extract ContactInfo (can be JSON object)
	if contactInfo, ok := data["contact_info"].(map[string]interface{}); ok {
		merchant.ContactInfo = contactInfo
	} else if contactInfo, ok := data["contact_info"].(string); ok {
		// Try to parse as JSON string
		var contactMap map[string]interface{}
		if err := json.Unmarshal([]byte(contactInfo), &contactMap); err == nil {
			merchant.ContactInfo = contactMap
		}
	}

	// Extract timestamps
	if createdAt, ok := data["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
			merchant.CreatedAt = t
		}
	}
	if merchant.CreatedAt.IsZero() {
		merchant.CreatedAt = time.Now()
	}

	if updatedAt, ok := data["updated_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
			merchant.UpdatedAt = t
		}
	}
	if merchant.UpdatedAt.IsZero() {
		merchant.UpdatedAt = time.Now()
	}

	return merchant, nil
}

// listMerchants lists merchants with pagination.
//
// FALLBACK BEHAVIOR:
//   - In production: Returns empty result set if query fails or no merchants found
//   - In development: Returns mock data if query fails (when allowed)
//
// MerchantFilters represents filtering options for merchant listing
type MerchantFilters struct {
	PortfolioType string
	RiskLevel     string
	Status        string
	SearchQuery   string
}

// listMerchants lists merchants with pagination, filtering, and sorting support
func (h *MerchantHandler) listMerchants(ctx context.Context, page, pageSize int, filters MerchantFilters, sortBy, sortOrder string, startTime time.Time) (*MerchantListResponse, error) {
	// Query Supabase for merchants with pagination, using retry logic and circuit breaker
	var result []map[string]interface{}
	err := h.circuitBreaker.Execute(ctx, func() error {
		// Use retry logic for the Supabase query
		retryConfig := resilience.DefaultRetryConfig()
		retryConfig.MaxAttempts = 3
		retryConfig.InitialDelay = 100 * time.Millisecond
		
		retryResult, retryErr := resilience.RetryWithBackoff(ctx, retryConfig, func() ([]map[string]interface{}, error) {
			var queryResult []map[string]interface{}
			
			// Build query with filters and sorting
			query := h.supabaseClient.GetClient().From("merchants").Select("*", "", false)
			
			// Apply filters
			if filters.PortfolioType != "" {
				query = query.Eq("portfolio_type", filters.PortfolioType)
			}
			if filters.RiskLevel != "" {
				query = query.Eq("risk_level", filters.RiskLevel)
			}
			if filters.Status != "" {
				query = query.Eq("status", filters.Status)
			}
			if filters.SearchQuery != "" {
				// Search in name and legal_name fields
				query = query.Or("name.ilike.%"+filters.SearchQuery+"%,legal_name.ilike.%"+filters.SearchQuery+"%", "")
			}
			
			// Apply sorting
			// Validate sortBy to prevent SQL injection
			validSortFields := map[string]bool{
				"name": true, "legal_name": true, "created_at": true, "updated_at": true,
				"portfolio_type": true, "risk_level": true, "status": true,
			}
			if validSortFields[sortBy] {
				if sortOrder == "desc" {
					query = query.Order(sortBy, &postgrest.OrderOpts{Ascending: false})
				} else {
					query = query.Order(sortBy, &postgrest.OrderOpts{Ascending: true})
				}
			} else {
				// Default to created_at desc if invalid sort field
				query = query.Order("created_at", &postgrest.OrderOpts{Ascending: false})
			}
			
			// Apply pagination
			_, queryErr := query.Range((page-1)*pageSize, page*pageSize-1, "").ExecuteTo(&queryResult)
			
			if queryErr != nil {
				return nil, queryErr
			}
			
			return queryResult, nil
		})
		
		if retryErr != nil {
			return retryErr
		}
		
		result = retryResult
		return nil
	})
	
	if err != nil {
		h.logger.Warn("Failed to fetch merchants from Supabase",
			zap.String("environment", h.config.Environment),
			zap.Error(err))
		
		// In production, return empty result set instead of mock data
		if h.config.Environment == "production" && !h.config.Merchant.AllowMockData {
			return &MerchantListResponse{
				Merchants:   []Merchant{},
				Total:       0,
				Page:        page,
				PageSize:    pageSize,
				TotalPages:  0,
				HasNext:     false,
				HasPrevious: false,
			}, nil
		}
		
		// FALLBACK: Return mock data for development (when allowed)
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
		
		return &MerchantListResponse{
			Merchants:   merchants,
			Total:       total,
			Page:        page,
			PageSize:    pageSize,
			TotalPages:  totalPages,
			HasNext:     page < totalPages,
			HasPrevious: page > 1,
		}, nil
	}
	
	// Convert results to Merchant structs
	merchants := make([]Merchant, 0, len(result))
	for _, row := range result {
		merchant, err := h.mapToMerchant(row)
		if err != nil {
			h.logger.Warn("Failed to map merchant data, skipping",
				zap.Error(err))
			continue
		}
		merchants = append(merchants, *merchant)
	}
	
	// Get total count from Supabase for accurate pagination
	total, err := h.supabaseClient.GetTableCount(ctx, "merchants")
	if err != nil {
		h.logger.Warn("Failed to get total count from Supabase, using current page count",
			zap.Error(err))
		// Fallback to current page count if query fails
		total = len(merchants)
	}
	totalPages := (total + pageSize - 1) / pageSize
	
	// Record successful request (non-fallback)
	if h.fallbackMetrics != nil {
		h.fallbackMetrics.RecordRequest(ctx, "merchant-service")
	}
	
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
// Handles paths like /api/v1/merchants/{id} and /api/v1/merchants/{id}/analytics
func (h *MerchantHandler) extractMerchantIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "merchants" && i+1 < len(parts) {
			merchantID := parts[i+1]
			// Skip empty parts and route segments
			if merchantID != "" && merchantID != "analytics" && merchantID != "website-analysis" && 
			   merchantID != "risk-score" && merchantID != "statistics" && merchantID != "search" {
				return merchantID
			}
		}
	}
	return ""
}

// getUserIDFromRequest extracts user ID from request context or headers
// Checks multiple sources: context (from API Gateway), headers, or falls back to "system"
func (h *MerchantHandler) getUserIDFromRequest(r *http.Request) string {
	ctx := r.Context()

	// Try to get from context (set by API Gateway auth middleware)
	if userID := ctx.Value("user_id"); userID != nil {
		if id, ok := userID.(string); ok && id != "" {
			return id
		}
	}

	// Try to get from X-User-ID header (common pattern for user identification)
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}

	// Try to get from Authorization header and extract user info
	// This is a fallback if API Gateway doesn't set context
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		// Extract token from Bearer header
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		
		// Decode JWT to extract user ID
		// Note: We decode without full validation since API Gateway already validated the token
		userID, err := h.decodeJWTUserID(tokenString)
		if err == nil && userID != "" {
			h.logger.Debug("Extracted user ID from JWT token",
				zap.String("user_id", userID))
			return userID
		}
		
		// Log warning if decoding failed but continue with fallback
		if err != nil {
			h.logger.Warn("Failed to decode JWT token for user ID extraction",
				zap.Error(err))
		}
	}

	// Fallback to "system" if no user ID found
	return "system"
}

// sanitizeInput sanitizes input to prevent XSS and SQL injection
func sanitizeInput(input string) string {
	if input == "" {
		return input
	}
	
	// Trim whitespace
	sanitized := strings.TrimSpace(input)
	
	// Remove HTML tags (basic implementation)
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	sanitized = htmlTagRegex.ReplaceAllString(sanitized, "")
	
	// Remove potentially dangerous SQL patterns (basic protection)
	// Note: Since we use parameterized queries, this is defense-in-depth
	dangerousPatterns := []string{
		"';", "\";", "--", "/*", "*/",
	}
	
	for _, pattern := range dangerousPatterns {
		sanitized = strings.ReplaceAll(sanitized, pattern, "")
	}
	
	return sanitized
}

// decodeJWTUserID decodes a JWT token and extracts the user ID from the "sub" claim
// This is a lightweight decode without full signature validation since the token
// was already validated by the API Gateway
func (h *MerchantHandler) decodeJWTUserID(tokenString string) (string, error) {
	// Split token into parts (header.payload.signature)
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid token format: expected 3 parts, got %d", len(parts))
	}

	// Decode the payload (second part)
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("failed to decode JWT payload: %w", err)
	}

	// Parse the payload as JSON to extract claims
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return "", fmt.Errorf("failed to parse JWT claims: %w", err)
	}

	// Extract user ID from "sub" claim (Supabase standard) or "user_id" claim
	userID := ""
	if sub, ok := claims["sub"].(string); ok && sub != "" {
		userID = sub
	} else if userIDClaim, ok := claims["user_id"].(string); ok && userIDClaim != "" {
		userID = userIDClaim
	} else if id, ok := claims["id"].(string); ok && id != "" {
		userID = id
	}

	if userID == "" {
		return "", fmt.Errorf("user ID not found in JWT claims")
	}

	return userID, nil
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
		errors.WriteInternalError(w, r, "Failed to get analytics data")
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

	w.Header().Set("Content-Type", "application/json")

	// Query total merchants
	var merchantCount []map[string]interface{}
	_, err := h.supabaseClient.GetClient().From("merchants").
		Select("count", "", false).
		ExecuteTo(&merchantCount)
	
	totalMerchants := 0
	if err == nil && len(merchantCount) > 0 {
		if count, ok := merchantCount[0]["count"].(float64); ok {
			totalMerchants = int(count)
		}
	}

	// Query total risk assessments
	var assessmentCount []map[string]interface{}
	_, err = h.supabaseClient.GetClient().From("risk_assessments").
		Select("count", "", false).
		ExecuteTo(&assessmentCount)
	
	totalAssessments := 0
	if err == nil && len(assessmentCount) > 0 {
		if count, ok := assessmentCount[0]["count"].(float64); ok {
			totalAssessments = int(count)
		}
	}

	// Query average risk score
	var avgScoreResult []map[string]interface{}
	_, err = h.supabaseClient.GetClient().From("risk_assessments").
		Select("risk_score", "", false).
		ExecuteTo(&avgScoreResult)
	
	averageRiskScore := 0.45 // Default
	if err == nil && len(avgScoreResult) > 0 {
		totalScore := 0.0
		count := 0
		for _, record := range avgScoreResult {
			if score, ok := record["risk_score"].(float64); ok {
				totalScore += score
				count++
			}
		}
		if count > 0 {
			averageRiskScore = totalScore / float64(count)
		}
	}

	// Query risk distribution
	var riskDistributionRecords []map[string]interface{}
	_, err = h.supabaseClient.GetClient().From("risk_assessments").
		Select("risk_level", "", false).
		ExecuteTo(&riskDistributionRecords)
	
	riskDistribution := map[string]interface{}{
		"low":    0.0,
		"medium": 0.0,
		"high":   0.0,
	}
	
	if err == nil && len(riskDistributionRecords) > 0 {
		lowCount := 0
		mediumCount := 0
		highCount := 0
		total := len(riskDistributionRecords)
		
		for _, record := range riskDistributionRecords {
			if level, ok := record["risk_level"].(string); ok {
				switch strings.ToLower(level) {
				case "low":
					lowCount++
				case "medium":
					mediumCount++
				case "high", "critical":
					highCount++
				}
			}
		}
		
		if total > 0 {
			riskDistribution["low"] = float64(lowCount) / float64(total)
			riskDistribution["medium"] = float64(mediumCount) / float64(total)
			riskDistribution["high"] = float64(highCount) / float64(total)
		}
	}

	// Query industry breakdown (from merchants table)
	var industryRecords []map[string]interface{}
	_, err = h.supabaseClient.GetClient().From("merchants").
		Select("industry", "", false).
		ExecuteTo(&industryRecords)
	
	industryBreakdown := []map[string]interface{}{}
	if err == nil {
		industryMap := make(map[string]struct {
			count int
			totalRisk float64
			riskCount int
		})
		
		// Count by industry
		for _, record := range industryRecords {
			if industry, ok := record["industry"].(string); ok && industry != "" {
				stats := industryMap[industry]
				stats.count++
				industryMap[industry] = stats
			}
		}
		
		// Get average risk scores per industry from risk_assessments
		var industryRiskRecords []map[string]interface{}
		_, err = h.supabaseClient.GetClient().From("risk_assessments").
			Select("industry,risk_score", "", false).
			ExecuteTo(&industryRiskRecords)
		
		if err == nil {
			for _, record := range industryRiskRecords {
				if industry, ok := record["industry"].(string); ok && industry != "" {
					if score, ok := record["risk_score"].(float64); ok {
						stats := industryMap[industry]
						stats.totalRisk += score
						stats.riskCount++
						industryMap[industry] = stats
					}
				}
			}
		}
		
		// Build industry breakdown
		for industry, stats := range industryMap {
			avgRisk := 0.45 // Default
			if stats.riskCount > 0 {
				avgRisk = stats.totalRisk / float64(stats.riskCount)
			}
			industryBreakdown = append(industryBreakdown, map[string]interface{}{
				"industry":        industry,
				"count":           stats.count,
				"averageRiskScore": avgRisk,
			})
		}
	}

	// Query country breakdown (from risk_assessments table)
	var countryRecords []map[string]interface{}
	_, err = h.supabaseClient.GetClient().From("risk_assessments").
		Select("country,risk_score", "", false).
		ExecuteTo(&countryRecords)
	
	countryBreakdown := []map[string]interface{}{}
	if err == nil {
		countryMap := make(map[string]struct {
			count int
			totalRisk float64
			riskCount int
		})
		
		for _, record := range countryRecords {
			if country, ok := record["country"].(string); ok && country != "" {
				stats := countryMap[country]
				stats.count++
				if score, ok := record["risk_score"].(float64); ok {
					stats.totalRisk += score
					stats.riskCount++
				}
				countryMap[country] = stats
			}
		}
		
		// Build country breakdown
		for country, stats := range countryMap {
			avgRisk := 0.45 // Default
			if stats.riskCount > 0 {
				avgRisk = stats.totalRisk / float64(stats.riskCount)
			}
			countryBreakdown = append(countryBreakdown, map[string]interface{}{
				"country":         country,
				"count":           stats.count,
				"averageRiskScore": avgRisk,
			})
		}
	}

	// Build response
	response := map[string]interface{}{
		"totalMerchants":    totalMerchants,
		"totalAssessments": totalAssessments,
		"averageRiskScore": averageRiskScore,
		"riskDistribution": riskDistribution,
		"industryBreakdown": industryBreakdown,
		"countryBreakdown":  countryBreakdown,
		"timestamp":         time.Now().Format(time.RFC3339),
	}

	// Log processing time for monitoring
	h.logger.Info("Merchant statistics retrieved",
		zap.Duration("processing_time", time.Since(startTime)),
		zap.Int("total_merchants", totalMerchants),
		zap.Int("total_assessments", totalAssessments),
	)

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
		errors.WriteBadRequest(w, r, "Invalid search request: Please provide valid JSON")
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
		errors.WriteInternalError(w, r, "Search failed")
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
		errors.WriteInternalError(w, r, "Failed to get portfolio types")
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
		errors.WriteInternalError(w, r, "Failed to get risk levels")
		return
	}

	response := map[string]interface{}{
		"risk_levels":     riskLevels,
		"timestamp":       time.Now(),
		"processing_time": time.Since(startTime).String(),
	}

	json.NewEncoder(w).Encode(response)
}

// HandleMerchantSpecificAnalytics handles GET /api/v1/merchants/{id}/analytics
func (h *MerchantHandler) HandleMerchantSpecificAnalytics(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// Extract merchant ID from path
	merchantID := h.extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		errors.WriteBadRequest(w, r, "Merchant ID is required")
		return
	}

	// Get merchant first to ensure it exists
	merchant, err := h.getMerchant(ctx, merchantID, startTime)
	if err != nil {
		h.logger.Error("Failed to get merchant for analytics",
			zap.String("merchant_id", merchantID),
			zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			errors.WriteNotFound(w, r, "Merchant not found")
			return
		}
		errors.WriteInternalError(w, r, "Failed to get merchant")
		return
	}

	// Query merchant_analytics table for real data
	var analyticsRecords []map[string]interface{}
	_, err = h.supabaseClient.GetClient().From("merchant_analytics").
		Select("*", "", false).
		Eq("merchant_id", merchantID).
		Limit(1, "").
		ExecuteTo(&analyticsRecords)

	// Initialize default values
	classification := map[string]interface{}{
		"primaryIndustry": merchant.Industry,
		"confidenceScore": 0.5,
		"riskLevel":       merchant.RiskLevel,
		"status":          "pending",
	}
	security := map[string]interface{}{
		"trustScore": 0.7,
		"sslValid":   false,
	}
	quality := map[string]interface{}{
		"completenessScore": 0.6,
		"dataPoints":        10,
	}

	// If analytics record exists, use real data
	if err == nil && len(analyticsRecords) > 0 {
		record := analyticsRecords[0]
		
		// Extract classification data
		if classificationData, ok := record["classification_data"].(map[string]interface{}); ok {
			if primaryIndustry, ok := classificationData["primaryIndustry"].(string); ok && primaryIndustry != "" {
				classification["primaryIndustry"] = primaryIndustry
			}
			if confidenceScore, ok := classificationData["confidenceScore"].(float64); ok {
				classification["confidenceScore"] = confidenceScore
			}
			if riskLevel, ok := classificationData["riskLevel"].(string); ok && riskLevel != "" {
				classification["riskLevel"] = riskLevel
			}
			if mccCodes, ok := classificationData["mccCodes"].([]interface{}); ok {
				classification["mccCodes"] = mccCodes
			}
			if sicCodes, ok := classificationData["sicCodes"].([]interface{}); ok {
				classification["sicCodes"] = sicCodes
			}
			if naicsCodes, ok := classificationData["naicsCodes"].([]interface{}); ok {
				classification["naicsCodes"] = naicsCodes
			}
		}
		
		// Get classification status
		if status, ok := record["classification_status"].(string); ok {
			classification["status"] = status
		}
		
		// Extract security data
		if securityData, ok := record["security_data"].(map[string]interface{}); ok {
			if trustScore, ok := securityData["trustScore"].(float64); ok {
				security["trustScore"] = trustScore
			}
			if sslValid, ok := securityData["sslValid"].(bool); ok {
				security["sslValid"] = sslValid
			}
		}
		
		// Extract quality data
		if qualityData, ok := record["quality_data"].(map[string]interface{}); ok {
			if completenessScore, ok := qualityData["completenessScore"].(float64); ok {
				quality["completenessScore"] = completenessScore
			}
			if dataPoints, ok := qualityData["dataPoints"].(float64); ok {
				quality["dataPoints"] = int(dataPoints)
			}
		}
	} else {
		// No analytics record found - check if job is processing
		if err == nil {
			// Record doesn't exist yet, status is pending
			classification["status"] = "pending"
		} else {
			h.logger.Warn("Failed to query merchant_analytics",
				zap.String("merchant_id", merchantID),
				zap.Error(err))
			classification["status"] = "pending"
		}
	}
	
	// Build response in the format expected by frontend
	response := map[string]interface{}{
		"merchantId":    merchantID,
		"classification": classification,
		"security":      security,
		"quality":       quality,
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// HandleTriggerAnalyticsRefresh handles POST /api/v1/merchants/{id}/analytics/refresh
// Triggers classification and website analysis jobs for existing merchants
func (h *MerchantHandler) HandleTriggerAnalyticsRefresh(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// Extract merchant ID from path
	merchantID := h.extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		errors.WriteBadRequest(w, r, "Merchant ID is required")
		return
	}

	// Fetch merchant from database
	var merchantResult []map[string]interface{}
	_, err := h.supabaseClient.GetClient().From("merchants").
		Select("*", "", false).
		Eq("id", merchantID).
		Limit(1, "").
		ExecuteTo(&merchantResult)

	if err != nil {
		h.logger.Error("Failed to fetch merchant for analytics refresh",
			zap.String("merchant_id", merchantID),
			zap.Error(err))
		errors.WriteInternalError(w, r, "Failed to fetch merchant")
		return
	}

	if len(merchantResult) == 0 {
		errors.WriteNotFound(w, r, "Merchant not found")
		return
	}

	merchantData := merchantResult[0]

	// Convert to Merchant struct
	// Helper to safely extract string values
	getStringValue := func(data map[string]interface{}, key string) string {
		if val, ok := data[key]; ok {
			if str, ok := val.(string); ok {
				return str
			}
		}
		return ""
	}

	merchant := &Merchant{
		ID:          merchantID,
		Name:        getStringValue(merchantData, "name"),
		Industry:    getStringValue(merchantData, "industry"),
		BusinessType: getStringValue(merchantData, "business_type"),
	}

	// Extract contact info
	if contactInfo, ok := merchantData["contact_info"].(map[string]interface{}); ok {
		merchant.ContactInfo = contactInfo
	}

	// Trigger classification job
	h.triggerClassificationJob(ctx, merchant)

	// Extract website URL and trigger website analysis if available
	websiteURL := ""

	// Primary: Check contact_info (new format)
	if merchant.ContactInfo != nil {
		if url, ok := merchant.ContactInfo["website"].(string); ok && url != "" {
			websiteURL = url
		}
	}

	// Safety fallback: Check legacy columns (only if not found in contact_info)
	if websiteURL == "" {
		// Check contact_website column
		if contactWebsite, ok := merchantData["contact_website"].(string); ok && contactWebsite != "" {
			websiteURL = contactWebsite
			h.logger.Info("Found website URL in legacy contact_website column, migrating to contact_info",
				zap.String("merchant_id", merchantID),
				zap.String("website_url", contactWebsite))
			
			// Auto-migrate: Update contact_info for future use
			if merchant.ContactInfo == nil {
				merchant.ContactInfo = make(map[string]interface{})
			}
			merchant.ContactInfo["website"] = contactWebsite
			
			// Persist the migration back to database
			go h.migrateWebsiteURLToContactInfo(ctx, merchantID, contactWebsite)
		}
		
		// Check website_url column
		if websiteURL == "" {
			if websiteURLField, ok := merchantData["website_url"].(string); ok && websiteURLField != "" {
				websiteURL = websiteURLField
				h.logger.Info("Found website URL in legacy website_url column, migrating to contact_info",
					zap.String("merchant_id", merchantID),
					zap.String("website_url", websiteURLField))
				
				// Auto-migrate
				if merchant.ContactInfo == nil {
					merchant.ContactInfo = make(map[string]interface{})
				}
				merchant.ContactInfo["website"] = websiteURLField
				
				// Persist the migration back to database
				go h.migrateWebsiteURLToContactInfo(ctx, merchantID, websiteURLField)
			}
		}
	}

	if websiteURL != "" {
		h.triggerWebsiteAnalysisJob(ctx, merchantID, websiteURL, merchant.Name)
	} else {
		// Mark website analysis as skipped
		h.markWebsiteAnalysisSkipped(ctx, merchantID)
	}

	// Determine website analysis status
	websiteAnalysisStatus := "skipped"
	websiteAnalysisReason := "no website URL"
	if websiteURL != "" {
		websiteAnalysisStatus = "triggered"
		websiteAnalysisReason = "website URL provided"
	}

	// Return success response
	response := map[string]interface{}{
		"merchant_id": merchantID,
		"status":      "triggered",
		"message":     "Analytics refresh jobs have been triggered",
		"jobs": map[string]interface{}{
			"classification": "triggered",
			"website_analysis": map[string]interface{}{
				"status": websiteAnalysisStatus,
				"reason": websiteAnalysisReason,
			},
		},
		"timestamp":       time.Now(),
		"processing_time": time.Since(startTime).String(),
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(response)
}

// HandleMerchantAnalyticsStatus handles GET /api/v1/merchants/{id}/analytics/status
func (h *MerchantHandler) HandleMerchantAnalyticsStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract merchant ID from path
	merchantID := h.extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		errors.WriteBadRequest(w, r, "Merchant ID is required")
		return
	}

	// Query merchant_analytics table for status
	var analyticsRecords []map[string]interface{}
	_, err := h.supabaseClient.GetClient().From("merchant_analytics").
		Select("classification_status,website_analysis_status,classification_updated_at,website_analysis_updated_at", "", false).
		Eq("merchant_id", merchantID).
		Limit(1, "").
		ExecuteTo(&analyticsRecords)

	// Default statuses
	status := map[string]interface{}{
		"classification": "pending",
		"websiteAnalysis": "pending",
	}

	// If analytics record exists, use real statuses
	if err == nil && len(analyticsRecords) > 0 {
		record := analyticsRecords[0]
		
		if classificationStatus, ok := record["classification_status"].(string); ok {
			status["classification"] = classificationStatus
		}
		
		if websiteAnalysisStatus, ok := record["website_analysis_status"].(string); ok {
			status["websiteAnalysis"] = websiteAnalysisStatus
		}
		
		// Add timestamps if available
		if classificationUpdatedAt, ok := record["classification_updated_at"].(string); ok {
			status["classificationUpdatedAt"] = classificationUpdatedAt
		}
		
		if websiteAnalysisUpdatedAt, ok := record["website_analysis_updated_at"].(string); ok {
			status["websiteAnalysisUpdatedAt"] = websiteAnalysisUpdatedAt
		}
	} else if err != nil {
		h.logger.Warn("Failed to query merchant_analytics for status",
			zap.String("merchant_id", merchantID),
			zap.Error(err))
	}

	response := map[string]interface{}{
		"merchantId": merchantID,
		"status":     status,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}

// HandleMerchantWebsiteAnalysis handles GET /api/v1/merchants/{id}/website-analysis
func (h *MerchantHandler) HandleMerchantWebsiteAnalysis(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// Extract merchant ID from path
	merchantID := h.extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		errors.WriteBadRequest(w, r, "Merchant ID is required")
		return
	}

	// Get merchant first to ensure it exists
	merchant, err := h.getMerchant(ctx, merchantID, startTime)
	if err != nil {
		h.logger.Error("Failed to get merchant for website analysis",
			zap.String("merchant_id", merchantID),
			zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			errors.WriteNotFound(w, r, "Merchant not found")
			return
		}
		errors.WriteInternalError(w, r, "Failed to get merchant")
		return
	}

	// Extract website URL from merchant data if available
	websiteURL := ""
	if merchant.ContactInfo != nil {
		if url, ok := merchant.ContactInfo["website"].(string); ok {
			websiteURL = url
		}
	}

	// Query merchant_analytics table for website analysis data
	var analyticsRecords []map[string]interface{}
	_, err = h.supabaseClient.GetClient().From("merchant_analytics").
		Select("*", "", false).
		Eq("merchant_id", merchantID).
		Limit(1, "").
		ExecuteTo(&analyticsRecords)

	// Initialize default values
	ssl := map[string]interface{}{
		"valid": false,
	}
	securityHeaders := map[string]interface{}{
		"hasHttps":         false,
		"hasHsts":          false,
		"hasCsp":           false,
		"hasXFrameOptions": false,
		"hasXContentType":  false,
		"securityScore":    0.0,
	}
	performance := map[string]interface{}{
		"loadTime":        0.0,
		"pageSize":        0,
		"requestCount":    0,
		"performanceScore": 0.0,
	}
	accessibility := map[string]interface{}{
		"score": 0.0,
	}
	status := "pending"
	lastAnalyzed := ""

	// If analytics record exists, use real data
	if err == nil && len(analyticsRecords) > 0 {
		record := analyticsRecords[0]

		// Get website analysis status
		if analysisStatus, ok := record["website_analysis_status"].(string); ok {
			status = analysisStatus
		}

		// Extract website analysis data
		if analysisData, ok := record["website_analysis_data"].(map[string]interface{}); ok {
			// Extract SSL data
			if sslData, ok := analysisData["ssl"].(map[string]interface{}); ok {
				if valid, ok := sslData["valid"].(bool); ok {
					ssl["valid"] = valid
				}
				if expiresAt, ok := sslData["expiresAt"].(string); ok {
					ssl["expiresAt"] = expiresAt
				}
				if issuer, ok := sslData["issuer"].(string); ok {
					ssl["issuer"] = issuer
				}
			}

			// Extract security headers
			if headersData, ok := analysisData["securityHeaders"].(map[string]interface{}); ok {
				if hasHttps, ok := headersData["hasHttps"].(bool); ok {
					securityHeaders["hasHttps"] = hasHttps
				}
				if hasHsts, ok := headersData["hasHsts"].(bool); ok {
					securityHeaders["hasHsts"] = hasHsts
				}
				if hasCsp, ok := headersData["hasCsp"].(bool); ok {
					securityHeaders["hasCsp"] = hasCsp
				}
				if hasXFrameOptions, ok := headersData["hasXFrameOptions"].(bool); ok {
					securityHeaders["hasXFrameOptions"] = hasXFrameOptions
				}
				if hasXContentType, ok := headersData["hasXContentType"].(bool); ok {
					securityHeaders["hasXContentType"] = hasXContentType
				}
				if securityScore, ok := headersData["securityScore"].(float64); ok {
					securityHeaders["securityScore"] = securityScore
				}
				if missingHeaders, ok := headersData["missingHeaders"].([]interface{}); ok {
					securityHeaders["missingHeaders"] = missingHeaders
				}
			}

			// Extract performance data
			if perfData, ok := analysisData["performance"].(map[string]interface{}); ok {
				if loadTime, ok := perfData["loadTime"].(float64); ok {
					performance["loadTime"] = loadTime
				}
				if pageSize, ok := perfData["pageSize"].(float64); ok {
					performance["pageSize"] = int(pageSize)
				}
				if requestCount, ok := perfData["requestCount"].(float64); ok {
					performance["requestCount"] = int(requestCount)
				}
				if performanceScore, ok := perfData["performanceScore"].(float64); ok {
					performance["performanceScore"] = performanceScore
				}
			}

			// Extract accessibility data
			if accData, ok := analysisData["accessibility"].(map[string]interface{}); ok {
				if score, ok := accData["score"].(float64); ok {
					accessibility["score"] = score
				}
				if issues, ok := accData["issues"].([]interface{}); ok {
					accessibility["issues"] = issues
				}
				if wcagCompliance, ok := accData["wcagCompliance"].(string); ok {
					accessibility["wcagCompliance"] = wcagCompliance
				}
			}

			// Extract last analyzed timestamp
			if lastAnalyzedStr, ok := analysisData["lastAnalyzed"].(string); ok {
				lastAnalyzed = lastAnalyzedStr
			}
		}
	} else {
		// No analytics record found
		if err != nil {
			h.logger.Warn("Failed to query merchant_analytics for website analysis",
				zap.String("merchant_id", merchantID),
				zap.Error(err))
		}
		// Status remains "pending"
	}

	// If status is "skipped", return appropriate message
	if status == "skipped" {
		response := map[string]interface{}{
			"merchantId":   merchantID,
			"websiteUrl":   websiteURL,
			"status":       "skipped",
			"message":      "Website analysis skipped - no website URL provided",
			"lastAnalyzed": "",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Build response in the format expected by frontend
	response := map[string]interface{}{
		"merchantId":      merchantID,
		"websiteUrl":     websiteURL,
		"ssl":            ssl,
		"securityHeaders": securityHeaders,
		"performance":    performance,
		"accessibility":  accessibility,
		"status":         status,
		"lastAnalyzed":   lastAnalyzed,
	}

	json.NewEncoder(w).Encode(response)
}

// HandleMerchantRiskScore handles GET /api/v1/merchants/{id}/risk-score
func (h *MerchantHandler) HandleMerchantRiskScore(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// Extract merchant ID from path
	merchantID := h.extractMerchantIDFromPath(r.URL.Path)
	if merchantID == "" {
		errors.WriteBadRequest(w, r, "Merchant ID is required")
		return
	}

	// Get merchant first to ensure it exists
	merchant, err := h.getMerchant(ctx, merchantID, startTime)
	if err != nil {
		h.logger.Error("Failed to get merchant for risk score",
			zap.String("merchant_id", merchantID),
			zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			errors.WriteNotFound(w, r, "Merchant not found")
			return
		}
		errors.WriteInternalError(w, r, "Failed to get merchant")
		return
	}

	// Query risk_assessments table for real risk assessment data
	var riskAssessments []map[string]interface{}
	_, err = h.supabaseClient.GetClient().From("risk_assessments").
		Select("*", "", false).
		Eq("business_id", merchantID).
		Order("created_at", &postgrest.OrderOpts{Ascending: false}).
		Limit(1, "").
		ExecuteTo(&riskAssessments)

	// Initialize default values
	riskScore := 0.5
	riskLevel := "medium"
	confidenceScore := 0.85
	assessmentDate := time.Now().Format(time.RFC3339)
	factors := []map[string]interface{}{}

	// If risk assessment exists, use real data
	if err == nil && len(riskAssessments) > 0 {
		assessment := riskAssessments[0]

		// Extract risk score
		if score, ok := assessment["risk_score"].(float64); ok {
			riskScore = score
		} else if score, ok := assessment["risk_score"].(string); ok {
			// Handle string conversion if needed
			if parsed, parseErr := strconv.ParseFloat(score, 64); parseErr == nil {
				riskScore = parsed
			}
		}

		// Extract risk level
		if level, ok := assessment["risk_level"].(string); ok {
			riskLevel = strings.ToLower(level)
		}

		// Extract confidence score
		if conf, ok := assessment["confidence_score"].(float64); ok {
			confidenceScore = conf
		}

		// Extract assessment date
		if date, ok := assessment["created_at"].(string); ok {
			assessmentDate = date
		} else if date, ok := assessment["completed_at"].(string); ok {
			assessmentDate = date
		}

		// Extract risk factors
		if riskFactorsData, ok := assessment["risk_factors"].([]interface{}); ok {
			for _, factorData := range riskFactorsData {
				if factorMap, ok := factorData.(map[string]interface{}); ok {
					factor := map[string]interface{}{}
					if category, ok := factorMap["category"].(string); ok {
						factor["category"] = category
					} else if category, ok := factorMap["factor_category"].(string); ok {
						factor["category"] = category
					}
					if score, ok := factorMap["score"].(float64); ok {
						factor["score"] = score
					} else if score, ok := factorMap["impact_score"].(float64); ok {
						factor["score"] = score
					}
					if weight, ok := factorMap["weight"].(float64); ok {
						factor["weight"] = weight
					} else if weight, ok := factorMap["factor_weight"].(float64); ok {
						factor["weight"] = weight
					}
					if len(factor) > 0 {
						factors = append(factors, factor)
					}
				}
			}
		}
	} else {
		// No risk assessment found - use merchant's risk level as fallback
		if err != nil {
			h.logger.Warn("Failed to query risk_assessments",
				zap.String("merchant_id", merchantID),
				zap.Error(err))
		}

		// Map merchant risk level to score
		riskScoreMap := map[string]float64{
			"low":    0.2,
			"medium": 0.5,
			"high":   0.8,
		}

		if score, ok := riskScoreMap[strings.ToLower(merchant.RiskLevel)]; ok {
			riskScore = score
		}

		riskLevel = strings.ToLower(merchant.RiskLevel)
		if riskLevel != "low" && riskLevel != "medium" && riskLevel != "high" {
			riskLevel = "medium"
		}

		// Use default factors if no assessment exists
		factors = []map[string]interface{}{
			{
				"category": "Business Profile",
				"score":    riskScore,
				"weight":   1.0,
			},
		}

		// Add status to indicate no assessment
		response := map[string]interface{}{
			"merchant_id":     merchantID,
			"risk_score":      riskScore,
			"risk_level":      riskLevel,
			"confidence_score": confidenceScore,
			"assessment_date":  assessmentDate,
			"factors":          factors,
			"status":           "no_assessment",
			"message":          "No risk assessment found. Using merchant risk level as fallback.",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	// Build response in the format expected by frontend
	response := map[string]interface{}{
		"merchant_id":     merchantID,
		"risk_score":      riskScore,
		"risk_level":      riskLevel,
		"confidence_score": confidenceScore,
		"assessment_date":  assessmentDate,
		"factors":          factors,
	}

	json.NewEncoder(w).Encode(response)
}
