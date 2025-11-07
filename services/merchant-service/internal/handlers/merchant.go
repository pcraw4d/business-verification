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

	"kyb-platform/internal/cache"
	"kyb-platform/internal/metrics"
	"kyb-platform/internal/queue"
	"kyb-platform/internal/resilience"
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
		
		// Determine appropriate HTTP status code based on error
		statusCode := http.StatusInternalServerError
		errorMsg := fmt.Sprintf("Failed to get merchant: %v", err)
		
		// Check if it's a "not found" error
		if strings.Contains(err.Error(), "not found") {
			statusCode = http.StatusNotFound
			errorMsg = "Merchant not found"
		} else if strings.Contains(err.Error(), "unavailable") {
			// Database unavailable - return 503 Service Unavailable
			statusCode = http.StatusServiceUnavailable
			errorMsg = "Service temporarily unavailable"
			w.Header().Set("Retry-After", "30") // Suggest retry after 30 seconds
		}
		
		http.Error(w, errorMsg, statusCode)
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
// TODO: Add retry logic with exponential backoff for Supabase queries
// TODO: Implement circuit breaker pattern for Supabase connection
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
		
		// In production, return proper 404 response instead of mock data
		if h.config.Environment == "production" && !h.config.Merchant.AllowMockData {
			return nil, fmt.Errorf("merchant not found: %s", merchantID)
		}
		
		// FALLBACK: Return mock data when merchant not found (development only)
		fallbackStart := time.Now()
		mockMerchant, mockErr := h.getMockMerchant(merchantID)
		if mockErr != nil {
			return nil, mockErr
		}
		
		// Record fallback usage metrics
		if h.fallbackMetrics != nil {
			h.fallbackMetrics.RecordFallbackUsage(ctx, "merchant-service", "missing_record", "supabase", time.Since(fallbackStart))
		}
		
		return mockMerchant, nil
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
// TODO: Implement Supabase query with pagination support
// TODO: Add filtering and sorting capabilities
func (h *MerchantHandler) listMerchants(ctx context.Context, page, pageSize int, startTime time.Time) (*MerchantListResponse, error) {
	// Try to query Supabase for merchants
	var result []map[string]interface{}
	_, err := h.supabaseClient.GetClient().From("merchants").
		Select("*", "", false).
		Range((page-1)*pageSize, page*pageSize-1).
		ExecuteTo(&result)
	
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
	
	// TODO: Get total count from Supabase for accurate pagination
	total := len(merchants)
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
