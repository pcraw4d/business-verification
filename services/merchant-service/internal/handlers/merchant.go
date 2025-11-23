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

	"go.uber.org/zap"
	postgrest "github.com/supabase-community/postgrest-go"

	"kyb-platform/services/merchant-service/internal/errors"
	"kyb-platform/services/merchant-service/internal/cache"
	"kyb-platform/services/merchant-service/internal/metrics"
	"kyb-platform/services/merchant-service/internal/queue"
	"kyb-platform/services/merchant-service/internal/resilience"
	"kyb-platform/services/merchant-service/internal/config"
	"kyb-platform/services/merchant-service/internal/supabase"
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
}

// NewMerchantHandler creates a new merchant handler
func NewMerchantHandler(supabaseClient *supabase.Client, logger *zap.Logger, config *config.Config) *MerchantHandler {
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

	return merchant, nil
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
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	// Return statistics matching frontend PortfolioStatisticsSchema
	// Schema requires: totalMerchants, totalAssessments, averageRiskScore, riskDistribution, industryBreakdown, countryBreakdown, timestamp
	// Note: This is a mock response. In production, this should query the database.
	response := map[string]interface{}{
		"totalMerchants":   5000,
		"totalAssessments": 7500, // Total risk assessments performed
		"averageRiskScore": 0.45, // Average risk score (0-1)
		"riskDistribution": map[string]interface{}{
			"low":    0.2, // 20%
			"medium": 0.6, // 60%
			"high":   0.2, // 20%
		},
		"industryBreakdown": []map[string]interface{}{
			{
				"industry":        "Technology",
				"count":           2000,
				"averageRiskScore": 0.3,
			},
			{
				"industry":        "Finance",
				"count":           1500,
				"averageRiskScore": 0.5,
			},
			{
				"industry":        "Retail",
				"count":           1000,
				"averageRiskScore": 0.4,
			},
		},
		"countryBreakdown": []map[string]interface{}{
			{
				"country":         "US",
				"count":           3000,
				"averageRiskScore": 0.4,
			},
			{
				"country":         "GB",
				"count":           1000,
				"averageRiskScore": 0.5,
			},
			{
				"country":         "CA",
				"count":           500,
				"averageRiskScore": 0.35,
			},
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	// Log processing time for monitoring
	h.logger.Info("Merchant statistics retrieved",
		zap.Duration("processing_time", time.Since(startTime)),
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

	// Generate merchant-specific analytics in the format expected by the frontend
	// Frontend expects: { merchantId, classification, security, quality, timestamp }
	
	// Map risk level to confidence score (default to medium)
	confidenceScore := 0.5
	if strings.ToLower(merchant.RiskLevel) == "low" {
		confidenceScore = 0.8
	} else if strings.ToLower(merchant.RiskLevel) == "high" {
		confidenceScore = 0.3
	}
	
	// Build classification data
	classification := map[string]interface{}{
		"primaryIndustry": merchant.Industry,
		"confidenceScore": confidenceScore,
		"riskLevel":       merchant.RiskLevel,
	}
	
	// Build security data
	security := map[string]interface{}{
		"trustScore": 0.7, // Default trust score
		"sslValid":   true, // TODO: Implement actual SSL validation
	}
	
	// Build quality data
	quality := map[string]interface{}{
		"completenessScore": 0.6, // Default completeness
		"dataPoints":        10,  // Default data points
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

	// Generate website analysis data in the format expected by the frontend
	// Frontend expects: { merchantId, websiteUrl, ssl, securityHeaders, performance, accessibility, lastAnalyzed }
	
	ssl := map[string]interface{}{
		"valid": true, // TODO: Implement actual SSL validation
	}
	
	securityHeaders := []map[string]interface{}{}
	
	performance := map[string]interface{}{
		"loadTime": 0,    // TODO: Get from actual analysis
		"pageSize": 0,    // TODO: Get from actual analysis
		"requests": 0,   // TODO: Get from actual analysis
		"score":    75,  // Default score
	}
	
	accessibility := map[string]interface{}{
		"score": 0.8, // Default accessibility score
	}
	
	// Build response in the format expected by frontend
	response := map[string]interface{}{
		"merchantId":      merchantID,
		"websiteUrl":     websiteURL,
		"ssl":            ssl,
		"securityHeaders": securityHeaders,
		"performance":    performance,
		"accessibility":  accessibility,
		"lastAnalyzed":   time.Now().Format(time.RFC3339),
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

	// Map risk level to numeric score
	riskScoreMap := map[string]float64{
		"low":    0.2,
		"medium": 0.5,
		"high":   0.8,
	}

	riskScore := riskScoreMap[strings.ToLower(merchant.RiskLevel)]
	if riskScore == 0 {
		riskScore = 0.5 // Default to medium
	}

	// Ensure risk_level is one of the valid enum values
	riskLevel := strings.ToLower(merchant.RiskLevel)
	if riskLevel != "low" && riskLevel != "medium" && riskLevel != "high" {
		riskLevel = "medium" // Default to medium if invalid
	}

	// Generate risk score response matching MerchantRiskScoreSchema
	// Schema requires: merchant_id, risk_level, assessment_date, factors (array of objects)
	// Optional: risk_score, confidence_score
	response := map[string]interface{}{
		"merchant_id":    merchantID,
		"risk_score":     riskScore,
		"risk_level":     riskLevel,
		"confidence_score": 0.85, // Default confidence score
		"assessment_date": time.Now().Format(time.RFC3339),
		"factors": []map[string]interface{}{
			{
				"category": "Business Age",
				"score":    0.2,
				"weight":   0.3,
			},
			{
				"category": "Financial Stability",
				"score":    0.4,
				"weight":   0.4,
			},
			{
				"category": "Compliance History",
				"score":    0.3,
				"weight":   0.3,
			},
		},
	}

	json.NewEncoder(w).Encode(response)
}
