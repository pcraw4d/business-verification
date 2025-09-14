package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/handlers"
	"github.com/pcraw4d/business-verification/internal/middleware"
	"github.com/pcraw4d/business-verification/internal/modules/database_classification"
)

// RailwayServer represents the Railway deployment server
type RailwayServer struct {
	server                *http.Server
	classificationService *classification.IntegrationService
	databaseModule        *database_classification.DatabaseClassificationModule
	supabaseClient        *database.SupabaseClient
	authMiddleware        *middleware.AuthMiddleware
	authHandler           *handlers.AuthHandler
	logger                *log.Logger
	zapLogger             *zap.Logger
	config                *config.Config
}

// NewRailwayServer creates a new Railway server instance
func NewRailwayServer() (*RailwayServer, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger
	logger := log.New(os.Stdout, "[railway-server] ", log.LstdFlags)
	zapLogger, _ := zap.NewProduction()

	// Initialize Supabase client with proper error handling
	var supabaseClient *database.SupabaseClient
	if cfg.Supabase.URL != "" && cfg.Supabase.APIKey != "" && cfg.Supabase.ServiceRoleKey != "" {
		supabaseConfig := &database.SupabaseConfig{
			URL:            cfg.Supabase.URL,
			APIKey:         cfg.Supabase.APIKey,
			ServiceRoleKey: cfg.Supabase.ServiceRoleKey,
			JWTSecret:      cfg.Supabase.JWTSecret,
		}

		supabaseClient, err = database.NewSupabaseClient(supabaseConfig, logger)
		if err != nil {
			logger.Printf("⚠️ Warning: Failed to initialize Supabase client: %v", err)
			logger.Printf("📝 Supabase URL: %s", cfg.Supabase.URL)
			logger.Printf("📝 API Key present: %t", cfg.Supabase.APIKey != "")
			logger.Printf("📝 Service Role Key present: %t", cfg.Supabase.ServiceRoleKey != "")
			supabaseClient = nil
		} else {
			// Test the connection
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := supabaseClient.Connect(ctx); err != nil {
				logger.Printf("⚠️ Warning: Failed to connect to Supabase: %v", err)
				supabaseClient = nil
			} else {
				logger.Printf("✅ Successfully connected to Supabase")
			}
		}
	} else {
		logger.Printf("⚠️ Supabase configuration incomplete - using fallback mode")
		logger.Printf("📝 Required: SUPABASE_URL, SUPABASE_API_KEY, SUPABASE_SERVICE_ROLE_KEY")
	}

	// Initialize classification service only if Supabase client is available
	var classificationService *classification.IntegrationService
	if supabaseClient != nil {
		classificationService = classification.NewIntegrationService(supabaseClient, logger)
	} else {
		logger.Printf("⚠️ Classification service will use fallback mode (no Supabase)")
	}

	// Initialize database module
	databaseModuleConfig := &database_classification.Config{
		ModuleID:          "railway-classification",
		ModuleName:        "Railway Classification Module",
		ModuleVersion:     "1.0.0",
		ModuleDescription: "Database-driven classification for Railway deployment",
		RequestTimeout:    30 * time.Second,
		MaxConcurrency:    10,
		EnableCaching:     true,
		CacheTTL:          5 * time.Minute,
	}
	databaseModule, err := database_classification.NewDatabaseClassificationModule(supabaseClient, logger, databaseModuleConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database module: %w", err)
	}

	// Initialize authentication middleware
	authConfig := &middleware.AuthConfig{
		JWTSecret:    cfg.Auth.JWTSecret,
		APIKeySecret: cfg.Auth.APIKeySecret,
		TokenExpiry:  cfg.Auth.JWTExpiration,
		RequireAuth:  cfg.Auth.RequireAuth,
	}
	authMiddleware := middleware.NewAuthMiddleware(authConfig)
	authHandler := handlers.NewAuthHandler(authMiddleware, logger)

	// Create router
	router := mux.NewRouter()

	// Create server
	server := &RailwayServer{
		classificationService: classificationService,
		databaseModule:        databaseModule,
		supabaseClient:        supabaseClient,
		authMiddleware:        authMiddleware,
		authHandler:           authHandler,
		logger:                logger,
		zapLogger:             zapLogger,
		config:                cfg,
	}

	// Setup routes
	server.setupRoutes(router)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.server = httpServer

	return server, nil
}

// setupRoutes configures all API routes
func (s *RailwayServer) setupRoutes(router *mux.Router) {
	// CORS middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check
	router.HandleFunc("/health", s.handleHealth).Methods("GET")

	// Authentication endpoints
	auth := router.PathPrefix("/auth").Subrouter()
	auth.HandleFunc("/login", s.authHandler.Login).Methods("POST")
	auth.HandleFunc("/validate", s.authHandler.ValidateToken).Methods("POST")
	auth.HandleFunc("/status", s.authHandler.GetAuthStatus).Methods("GET")
	auth.HandleFunc("/api-key", s.authHandler.GenerateAPIKey).Methods("POST")

	// Business Intelligence Classification
	router.HandleFunc("/v1/classify", s.handleClassify).Methods("POST")

	// Merchant Management API (with optional authentication)
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/merchants", s.handleGetMerchants).Methods("GET")
	api.HandleFunc("/merchants/search", s.handleSearchMerchants).Methods("POST")
	api.HandleFunc("/merchants/analytics", s.handleMerchantAnalytics).Methods("GET")
	api.HandleFunc("/merchants/portfolio-types", s.handlePortfolioTypes).Methods("GET")
	api.HandleFunc("/merchants/risk-levels", s.handleRiskLevels).Methods("GET")
	api.HandleFunc("/merchants/statistics", s.handleMerchantStatistics).Methods("GET")
	api.HandleFunc("/merchants/{id}", s.handleGetMerchant).Methods("GET")

	// Apply authentication middleware to protected routes (optional)
	if s.config.Auth.RequireAuth {
		api.Use(s.authMiddleware.RequireAuth)
	}

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/")))
}

// handleHealth handles health check requests
func (s *RailwayServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "3.2.0",
		"features": map[string]bool{
			"supabase_integration":           s.supabaseClient != nil,
			"database_driven_classification": true,
			"enhanced_keyword_matching":      true,
			"industry_detection":             true,
			"confidence_scoring":             true,
		},
		"supabase_status": s.getSupabaseStatus(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// getSupabaseStatus returns the current Supabase connection status
func (s *RailwayServer) getSupabaseStatus() map[string]interface{} {
	if s.supabaseClient == nil {
		return map[string]interface{}{
			"connected": false,
			"reason":    "client_not_initialized",
		}
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.supabaseClient.Ping(ctx); err != nil {
		return map[string]interface{}{
			"connected": false,
			"reason":    "connection_failed",
			"error":     err.Error(),
		}
	}

	return map[string]interface{}{
		"connected": true,
		"url":       s.supabaseClient.GetURL(),
	}
}

// handleClassify handles business classification requests
func (s *RailwayServer) handleClassify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		BusinessName string `json:"business_name"`
		Description  string `json:"description"`
		WebsiteURL   string `json:"website_url,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.BusinessName == "" || req.Description == "" {
		http.Error(w, "business_name and description are required", http.StatusBadRequest)
		return
	}

	// Process classification
	var result map[string]interface{}
	if s.classificationService != nil {
		result = s.classificationService.ProcessBusinessClassification(
			context.Background(),
			req.BusinessName,
			req.Description,
			req.WebsiteURL,
		)
	} else {
		// Fallback mock classification when Supabase is not available
		result = s.getFallbackClassification(req.BusinessName, req.Description, req.WebsiteURL)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// getFallbackClassification returns mock classification data
func (s *RailwayServer) getFallbackClassification(businessName, description, websiteURL string) map[string]interface{} {
	// Generate a business ID for tracking
	businessID := fmt.Sprintf("biz_%d", time.Now().Unix())
	
	return map[string]interface{}{
		"success":       true,
		"business_id":   businessID,
		"business_name": businessName,
		"description":   description,
		"website_url":   websiteURL,
		"classification": map[string]interface{}{
			"mcc_codes": []map[string]interface{}{
				{"code": "7372", "description": "Computer Programming Services", "confidence": 0.95},
				{"code": "7373", "description": "Computer Integrated Systems Design", "confidence": 0.88},
			},
			"sic_codes": []map[string]interface{}{
				{"code": "7372", "description": "Computer Programming Services", "confidence": 0.92},
				{"code": "7373", "description": "Computer Integrated Systems Design", "confidence": 0.85},
			},
			"naics_codes": []map[string]interface{}{
				{"code": "541511", "description": "Custom Computer Programming Services", "confidence": 0.98},
				{"code": "541512", "description": "Computer Systems Design Services", "confidence": 0.90},
			},
		},
		"confidence_score": 0.94,
		"status":           "success",
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"data_source":      "fallback_mock",
	}
}

// handleGetMerchants handles GET /api/v1/merchants
func (s *RailwayServer) handleGetMerchants(w http.ResponseWriter, r *http.Request) {
	// Try to get merchants from Supabase first
	if s.supabaseClient != nil {
		merchants, err := s.getMerchantsFromSupabase()
		if err != nil {
			s.logger.Printf("⚠️ Failed to get merchants from Supabase: %v", err)
			// Fall back to mock data
			s.handleGetMerchantsMock(w, r)
			return
		}

		response := map[string]interface{}{
			"merchants":   merchants,
			"total":       len(merchants),
			"page":        1,
			"limit":       10,
			"data_source": "supabase",
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		return
	}

	// Use mock data if Supabase is not available
	s.handleGetMerchantsMock(w, r)
}

// getMerchantsFromSupabase retrieves merchants from Supabase database
func (s *RailwayServer) getMerchantsFromSupabase() ([]map[string]interface{}, error) {
	// Query merchants from Supabase
	postgrestClient := s.supabaseClient.GetPostgrestClient()

	// Try to get merchants from the merchants table
	var merchants []map[string]interface{}
	_, err := postgrestClient.From("merchants").Select("*", "", false).ExecuteTo(&merchants)
	if err != nil {
		// If merchants table doesn't exist, try mock_merchants
		_, err2 := postgrestClient.From("mock_merchants").Select("*", "", false).ExecuteTo(&merchants)
		if err2 != nil {
			return nil, fmt.Errorf("failed to query merchants: %w", err2)
		}
	}

	// If no merchants found, return empty array
	if len(merchants) == 0 {
		return []map[string]interface{}{}, nil
	}

	return merchants, nil
}

// handleGetMerchantsMock handles GET /api/v1/merchants with mock data
func (s *RailwayServer) handleGetMerchantsMock(w http.ResponseWriter, r *http.Request) {
	// Enhanced mock merchant data
	merchants := []map[string]interface{}{
		{
			"id":                  "merchant_001",
			"name":                "Acme Corporation",
			"industry":            "Technology",
			"portfolio_type":      "High Volume",
			"risk_level":          "Medium",
			"status":              "Active",
			"created_at":          "2024-01-15T10:30:00Z",
			"revenue":             1500000,
			"address":             "123 Tech Street, Silicon Valley, CA 94000",
			"phone":               "+1-555-0123",
			"email":               "contact@acme.com",
			"website":             "https://www.acme.com",
			"employees":           150,
			"founded":             "2015",
			"verification_status": "Verified",
			"compliance_score":    95,
			"risk_factors": []string{
				"High transaction volume",
				"International operations",
			},
			"recent_activity": []map[string]interface{}{
				{
					"date":        "2024-03-15T10:30:00Z",
					"type":        "Transaction",
					"description": "Large payment processed",
					"amount":      50000,
				},
				{
					"date":        "2024-03-14T14:20:00Z",
					"type":        "Verification",
					"description": "Document verification completed",
					"amount":      0,
				},
			},
		},
		{
			"id":                  "merchant_002",
			"name":                "Global Retail Inc",
			"industry":            "Retail",
			"portfolio_type":      "Standard",
			"risk_level":          "Low",
			"status":              "Active",
			"created_at":          "2024-02-20T14:45:00Z",
			"revenue":             850000,
			"address":             "456 Commerce Ave, New York, NY 10001",
			"phone":               "+1-555-0456",
			"email":               "info@globalretail.com",
			"website":             "https://www.globalretail.com",
			"employees":           75,
			"founded":             "2018",
			"verification_status": "Verified",
			"compliance_score":    88,
			"risk_factors": []string{
				"Standard retail operations",
			},
			"recent_activity": []map[string]interface{}{
				{
					"date":        "2024-03-12T09:15:00Z",
					"type":        "Transaction",
					"description": "Monthly payment processed",
					"amount":      25000,
				},
			},
		},
		{
			"id":                  "merchant_003",
			"name":                "TechStart Solutions",
			"industry":            "Software",
			"portfolio_type":      "High Volume",
			"risk_level":          "High",
			"status":              "Pending",
			"created_at":          "2024-03-10T09:15:00Z",
			"revenue":             250000,
			"address":             "789 Innovation Blvd, Austin, TX 78701",
			"phone":               "+1-555-0789",
			"email":               "hello@techstart.com",
			"website":             "https://www.techstart.com",
			"employees":           25,
			"founded":             "2022",
			"verification_status": "Pending",
			"compliance_score":    72,
			"risk_factors": []string{
				"New business",
				"High growth potential",
				"Limited financial history",
			},
			"recent_activity": []map[string]interface{}{
				{
					"date":        "2024-03-08T16:30:00Z",
					"type":        "Verification",
					"description": "Initial verification submitted",
					"amount":      0,
				},
			},
		},
		{
			"id":                  "merchant_004",
			"name":                "Healthcare Partners LLC",
			"industry":            "Healthcare",
			"portfolio_type":      "Standard",
			"risk_level":          "Low",
			"status":              "Active",
			"created_at":          "2024-01-05T08:00:00Z",
			"revenue":             2200000,
			"address":             "321 Medical Center Dr, Boston, MA 02115",
			"phone":               "+1-555-0321",
			"email":               "contact@healthcarepartners.com",
			"website":             "https://www.healthcarepartners.com",
			"employees":           200,
			"founded":             "2010",
			"verification_status": "Verified",
			"compliance_score":    98,
			"risk_factors": []string{
				"Regulated industry",
			},
			"recent_activity": []map[string]interface{}{
				{
					"date":        "2024-03-16T11:45:00Z",
					"type":        "Transaction",
					"description": "Insurance payment processed",
					"amount":      75000,
				},
			},
		},
		{
			"id":                  "merchant_005",
			"name":                "Finance First Corp",
			"industry":            "Finance",
			"portfolio_type":      "High Volume",
			"risk_level":          "Medium",
			"status":              "Active",
			"created_at":          "2024-02-01T12:00:00Z",
			"revenue":             5000000,
			"address":             "555 Wall Street, New York, NY 10005",
			"phone":               "+1-555-0555",
			"email":               "info@financefirst.com",
			"website":             "https://www.financefirst.com",
			"employees":           300,
			"founded":             "2008",
			"verification_status": "Verified",
			"compliance_score":    92,
			"risk_factors": []string{
				"High transaction volume",
				"Financial services",
				"Regulatory compliance",
			},
			"recent_activity": []map[string]interface{}{
				{
					"date":        "2024-03-17T14:20:00Z",
					"type":        "Transaction",
					"description": "Investment transaction processed",
					"amount":      150000,
				},
			},
		},
	}

	response := map[string]interface{}{
		"merchants":   merchants,
		"total":       len(merchants),
		"page":        1,
		"limit":       10,
		"data_source": "mock_data",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetMerchant handles GET /api/v1/merchants/{id}
func (s *RailwayServer) handleGetMerchant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	merchantID := vars["id"]

	// Try to get merchant from Supabase first
	if s.supabaseClient != nil {
		merchant, err := s.getMerchantFromSupabase(merchantID)
		if err != nil {
			s.logger.Printf("⚠️ Failed to get merchant from Supabase: %v", err)
			// Fall back to mock data
			s.handleGetMerchantMock(w, r, merchantID)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(merchant)
		return
	}

	// Use mock data if Supabase is not available
	s.handleGetMerchantMock(w, r, merchantID)
}

// getMerchantFromSupabase retrieves a specific merchant from Supabase database
func (s *RailwayServer) getMerchantFromSupabase(merchantID string) (map[string]interface{}, error) {
	// Query merchant from Supabase
	postgrestClient := s.supabaseClient.GetPostgrestClient()

	// Try to get merchant from the merchants table
	var merchants []map[string]interface{}
	_, err := postgrestClient.From("merchants").Select("*", "", false).Eq("id", merchantID).ExecuteTo(&merchants)
	if err != nil {
		// If merchants table doesn't exist, try mock_merchants
		_, err2 := postgrestClient.From("mock_merchants").Select("*", "", false).Eq("id", merchantID).ExecuteTo(&merchants)
		if err2 != nil {
			return nil, fmt.Errorf("failed to query merchant: %w", err2)
		}
	}

	if len(merchants) == 0 {
		return nil, fmt.Errorf("merchant not found")
	}

	return merchants[0], nil
}

// handleGetMerchantMock handles GET /api/v1/merchants/{id} with mock data
func (s *RailwayServer) handleGetMerchantMock(w http.ResponseWriter, r *http.Request, merchantID string) {
	// Mock merchant detail data based on ID
	merchant := map[string]interface{}{
		"id":                  merchantID,
		"name":                "Acme Corporation",
		"industry":            "Technology",
		"portfolio_type":      "High Volume",
		"risk_level":          "Medium",
		"status":              "Active",
		"created_at":          "2024-01-15T10:30:00Z",
		"revenue":             1500000,
		"description":         "A leading technology company specializing in innovative solutions",
		"address":             "123 Tech Street, Silicon Valley, CA 94000",
		"phone":               "+1-555-0123",
		"email":               "contact@acme.com",
		"website":             "https://www.acme.com",
		"employees":           150,
		"founded":             "2015",
		"verification_status": "Verified",
		"compliance_score":    95,
		"risk_factors": []string{
			"High transaction volume",
			"International operations",
		},
		"recent_activity": []map[string]interface{}{
			{
				"date":        "2024-03-15T10:30:00Z",
				"type":        "Transaction",
				"description": "Large payment processed",
				"amount":      50000,
			},
			{
				"date":        "2024-03-14T14:20:00Z",
				"type":        "Verification",
				"description": "Document verification completed",
				"amount":      0,
			},
		},
		"data_source": "mock_data",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(merchant)
}

// handleSearchMerchants handles POST /api/v1/merchants/search
func (s *RailwayServer) handleSearchMerchants(w http.ResponseWriter, r *http.Request) {
	var searchReq struct {
		Query         string `json:"query,omitempty"`
		Industry      string `json:"industry,omitempty"`
		PortfolioType string `json:"portfolio_type,omitempty"`
		RiskLevel     string `json:"risk_level,omitempty"`
		Status        string `json:"status,omitempty"`
		Page          int    `json:"page,omitempty"`
		Limit         int    `json:"limit,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&searchReq); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// For now, return mock search results
	// TODO: Implement Supabase search when available
	results := []map[string]interface{}{
		{
			"id":             "merchant_001",
			"name":           "Acme Corporation",
			"industry":       "Technology",
			"portfolio_type": "High Volume",
			"risk_level":     "Medium",
			"status":         "Active",
			"revenue":        1500000,
		},
	}

	response := map[string]interface{}{
		"merchants":   results,
		"total":       len(results),
		"page":        searchReq.Page,
		"limit":       searchReq.Limit,
		"data_source": "mock_data",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMerchantAnalytics handles GET /api/v1/merchants/analytics
func (s *RailwayServer) handleMerchantAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics := map[string]interface{}{
		"total_merchants":   150,
		"active_merchants":  142,
		"pending_merchants": 8,
		"total_revenue":     25000000,
		"average_revenue":   166667,
		"portfolio_distribution": map[string]int{
			"High Volume": 45,
			"Standard":    78,
			"Low Volume":  27,
		},
		"risk_distribution": map[string]int{
			"Low":    65,
			"Medium": 52,
			"High":   33,
		},
		"industry_distribution": map[string]int{
			"Technology": 42,
			"Retail":     38,
			"Finance":    25,
			"Healthcare": 20,
			"Other":      25,
		},
		"monthly_growth": []map[string]interface{}{
			{"month": "2024-01", "merchants": 12, "revenue": 2100000},
			{"month": "2024-02", "merchants": 18, "revenue": 3200000},
			{"month": "2024-03", "merchants": 15, "revenue": 2800000},
		},
		"data_source": "mock_data",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

// handlePortfolioTypes handles GET /api/v1/merchants/portfolio-types
func (s *RailwayServer) handlePortfolioTypes(w http.ResponseWriter, r *http.Request) {
	portfolioTypes := []map[string]interface{}{
		{"id": "high_volume", "name": "High Volume", "description": "Merchants with high transaction volumes"},
		{"id": "standard", "name": "Standard", "description": "Standard merchant portfolio"},
		{"id": "low_volume", "name": "Low Volume", "description": "Merchants with low transaction volumes"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(portfolioTypes)
}

// handleRiskLevels handles GET /api/v1/merchants/risk-levels
func (s *RailwayServer) handleRiskLevels(w http.ResponseWriter, r *http.Request) {
	riskLevels := []map[string]interface{}{
		{"id": "low", "name": "Low", "description": "Low risk merchants"},
		{"id": "medium", "name": "Medium", "description": "Medium risk merchants"},
		{"id": "high", "name": "High", "description": "High risk merchants"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(riskLevels)
}

// handleMerchantStatistics handles GET /api/v1/merchants/statistics
func (s *RailwayServer) handleMerchantStatistics(w http.ResponseWriter, r *http.Request) {
	statistics := map[string]interface{}{
		"total_merchants":   150,
		"active_merchants":  142,
		"pending_merchants": 8,
		"total_revenue":     25000000,
		"average_revenue":   166667,
		"verification_rate": 94.7,
		"compliance_score":  92.3,
		"data_source":       "mock_data",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statistics)
}

// Start starts the server
func (s *RailwayServer) Start() error {
	s.logger.Printf("🚀 Starting RAILWAY SERVER v3.2.0 on %s", s.server.Addr)
	s.logger.Printf("📊 Supabase Integration: %t", s.supabaseClient != nil)
	if s.supabaseClient != nil {
		s.logger.Printf("🔗 Supabase URL: %s", s.supabaseClient.GetURL())
	}
	return s.server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *RailwayServer) Stop(ctx context.Context) error {
	s.logger.Printf("🛑 Stopping RAILWAY SERVER...")
	return s.server.Shutdown(ctx)
}

func main() {
	server, err := NewRailwayServer()
	if err != nil {
		log.Fatalf("Failed to create Railway server: %v", err)
	}

	// Start server
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
