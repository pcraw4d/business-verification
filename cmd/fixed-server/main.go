package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/modules/database_classification"
	"github.com/pcraw4d/business-verification/internal/shared"
	"go.uber.org/zap"
)

// SimplifiedServer represents a simplified server with essential features
type SimplifiedServer struct {
	server                *http.Server
	classificationService *classification.IntegrationService
	databaseModule        *database_classification.DatabaseClassificationModule
	supabaseClient        *database.SupabaseClient
	logger                *log.Logger
	zapLogger             *zap.Logger
	config                *config.Config
}

// NewSimplifiedServer creates a new simplified server
func NewSimplifiedServer(port string) *SimplifiedServer {
	logger := log.New(os.Stdout, "🚀 ", log.LstdFlags|log.Lshortfile)
	zapLogger, _ := zap.NewProduction()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Supabase client
	supabaseConfig := &database.SupabaseConfig{
		URL:            cfg.Supabase.URL,
		APIKey:         cfg.Supabase.APIKey,
		ServiceRoleKey: cfg.Supabase.ServiceRoleKey,
		JWTSecret:      cfg.Supabase.JWTSecret,
	}
	supabaseClient, err := database.NewSupabaseClient(supabaseConfig, logger)
	if err != nil {
		logger.Fatalf("Failed to create Supabase client: %v", err)
	}

	// Connect to Supabase
	ctx := context.Background()
	if err := supabaseClient.Connect(ctx); err != nil {
		logger.Fatalf("Failed to connect to Supabase: %v", err)
	}
	logger.Printf("✅ Successfully connected to Supabase")

	// Create classification service
	classificationService := classification.NewIntegrationService(supabaseClient, logger)

	// Create database classification module
	databaseModule, err := database_classification.NewDatabaseClassificationModule(
		supabaseClient,
		logger,
		database_classification.DefaultConfig(),
	)
	if err != nil {
		logger.Fatalf("Failed to create database classification module: %v", err)
	}

	// Start the database module
	if err := databaseModule.Start(ctx); err != nil {
		logger.Fatalf("Failed to start database classification module: %v", err)
	}

	// Create server
	server := &SimplifiedServer{
		classificationService: classificationService,
		databaseModule:        databaseModule,
		supabaseClient:        supabaseClient,
		logger:                logger,
		zapLogger:             zapLogger,
		config:                cfg,
	}

	// Setup routes
	mux := http.NewServeMux()
	server.setupRoutes(mux)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	server.server = httpServer

	logger.Printf("✅ Fixed Server initialized on port %s", port)
	return server
}

// setupRoutes configures all HTTP routes
func (s *SimplifiedServer) setupRoutes(mux *http.ServeMux) {
	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Classification endpoints
	mux.HandleFunc("/v1/classify", s.handleClassify)
	mux.HandleFunc("/v1/classify-legacy", s.handleClassifyLegacy)

	// Module status endpoints
	mux.HandleFunc("/v1/modules/status", s.handleModuleStatus)
	mux.HandleFunc("/v1/database/status", s.handleDatabaseStatus)

	// Merchant API endpoints (simplified without auth)
	s.setupMerchantRoutes(mux)

	// Static file serving for web interface
	s.setupStaticRoutes(mux)
}

// setupMerchantRoutes configures merchant API routes (without authentication)
func (s *SimplifiedServer) setupMerchantRoutes(mux *http.ServeMux) {
	// Merchant CRUD routes (simplified)
	mux.HandleFunc("/api/v1/merchants", s.handleMerchantsList)
	mux.HandleFunc("/api/v1/merchants/", s.handleMerchantsList) // Handle trailing slash
	mux.HandleFunc("/api/v1/merchants/{id}", s.handleMerchantDetail)

	// Merchant search route
	mux.HandleFunc("/api/v1/merchants/search", s.handleMerchantsSearch)

	// Analytics routes
	mux.HandleFunc("/api/v1/merchants/analytics", s.handleMerchantsAnalytics)
	mux.HandleFunc("/api/v1/merchants/portfolio-types", s.handlePortfolioTypes)
	mux.HandleFunc("/api/v1/merchants/risk-levels", s.handleRiskLevels)
	mux.HandleFunc("/api/v1/merchants/statistics", s.handleMerchantsStatistics)

	s.logger.Printf("✅ Merchant API routes configured (simplified)")
}

// setupStaticRoutes configures static file serving
func (s *SimplifiedServer) setupStaticRoutes(mux *http.ServeMux) {
	// Serve static files from web directory
	fs := http.FileServer(http.Dir("./web/"))
	mux.Handle("/", fs)

	// Specific routes for merchant pages
	mux.HandleFunc("/merchant-portfolio.html", s.serveMerchantPortfolio)
	mux.HandleFunc("/merchant-detail.html", s.serveMerchantDetail)
	mux.HandleFunc("/merchant-hub-integration.html", s.serveMerchantHub)

	s.logger.Printf("✅ Static routes configured")
}

// serveMerchantPortfolio serves the merchant portfolio page
func (s *SimplifiedServer) serveMerchantPortfolio(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/merchant-portfolio.html")
}

// serveMerchantDetail serves the merchant detail page
func (s *SimplifiedServer) serveMerchantDetail(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/merchant-detail.html")
}

// serveMerchantHub serves the merchant hub page
func (s *SimplifiedServer) serveMerchantHub(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/merchant-hub-integration.html")
}

// handleHealth handles health check requests
func (s *SimplifiedServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "3.1.0",
		"services": map[string]interface{}{
			"database_module": s.databaseModule.Health(),
			"supabase":        "connected",
			"classification":  "active",
		},
		"features": map[string]interface{}{
			"confidence_scoring":             true,
			"database_driven_classification": true,
			"enhanced_keyword_matching":      true,
			"industry_detection":             true,
			"supabase_integration":           true,
			"merchant_portfolio_management":  true,
			"merchant_hub_integration":       true,
		},
	}

	json.NewEncoder(w).Encode(health)
}

// handleClassify handles classification requests using the database module
func (s *SimplifiedServer) handleClassify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req shared.BusinessClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set request ID if not provided
	if req.ID == "" {
		req.ID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	// Process with classification service
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Use the classification service directly
	result := s.classificationService.ProcessBusinessClassification(
		ctx,
		req.BusinessName,
		req.Description,
		req.WebsiteURL,
	)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return the result
	json.NewEncoder(w).Encode(result)
}

// handleClassifyLegacy handles classification requests using the legacy service
func (s *SimplifiedServer) handleClassifyLegacy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req shared.BusinessClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Process with legacy service
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result := s.classificationService.ProcessBusinessClassification(
		ctx,
		req.BusinessName,
		req.Description,
		req.WebsiteURL,
	)

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Return the result
	json.NewEncoder(w).Encode(result)
}

// handleModuleStatus handles module status requests
func (s *SimplifiedServer) handleModuleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := map[string]interface{}{
		"database_classification_module": map[string]interface{}{
			"id":         s.databaseModule.ID(),
			"status":     s.databaseModule.Health(),
			"is_running": s.databaseModule.IsRunning(),
			"metadata":   s.databaseModule.Metadata(),
		},
		"timestamp": time.Now().UTC(),
	}

	json.NewEncoder(w).Encode(status)
}

// handleDatabaseStatus handles database status requests
func (s *SimplifiedServer) handleDatabaseStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Get health status from classification service
	healthStatus := s.classificationService.GetHealthStatus()

	status := map[string]interface{}{
		"database_health": healthStatus,
		"supabase_client": map[string]interface{}{
			"connected": true,
			"url":       s.config.Supabase.URL,
		},
		"timestamp": time.Now().UTC(),
	}

	json.NewEncoder(w).Encode(status)
}

// handleMerchantsList handles GET /api/v1/merchants
func (s *SimplifiedServer) handleMerchantsList(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Return mock merchant data for now
	mockMerchants := []map[string]interface{}{
		{
			"id":                "merchant-001",
			"name":              "Acme Corporation",
			"legal_name":        "Acme Corporation Inc.",
			"industry":          "Technology",
			"portfolio_type":    "onboarded",
			"risk_level":        "low",
			"compliance_status": "compliant",
			"status":            "active",
			"address": map[string]interface{}{
				"city":  "San Francisco",
				"state": "CA",
			},
			"contact_info": map[string]interface{}{
				"email": "contact@acme.com",
				"phone": "+1-555-123-4567",
			},
			"created_at": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
		},
		{
			"id":                "merchant-002",
			"name":              "TechStart Solutions",
			"legal_name":        "TechStart Solutions LLC",
			"industry":          "Software",
			"portfolio_type":    "prospective",
			"risk_level":        "medium",
			"compliance_status": "pending",
			"status":            "active",
			"address": map[string]interface{}{
				"city":  "Austin",
				"state": "TX",
			},
			"contact_info": map[string]interface{}{
				"email": "info@techstart.com",
				"phone": "+1-555-987-6543",
			},
			"created_at": time.Now().AddDate(0, -2, 0).Format(time.RFC3339),
		},
		{
			"id":                "merchant-003",
			"name":              "Global Retail Co",
			"legal_name":        "Global Retail Company",
			"industry":          "Retail",
			"portfolio_type":    "onboarded",
			"risk_level":        "high",
			"compliance_status": "review_required",
			"status":            "active",
			"address": map[string]interface{}{
				"city":  "New York",
				"state": "NY",
			},
			"contact_info": map[string]interface{}{
				"email": "support@globalretail.com",
				"phone": "+1-555-456-7890",
			},
			"created_at": time.Now().AddDate(0, -3, 0).Format(time.RFC3339),
		},
	}

	response := map[string]interface{}{
		"merchants": mockMerchants,
		"total":     len(mockMerchants),
		"page":      1,
		"page_size": 10,
		"has_more":  false,
	}

	json.NewEncoder(w).Encode(response)
}

// handleMerchantDetail handles GET /api/v1/merchants/{id}
func (s *SimplifiedServer) handleMerchantDetail(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Extract merchant ID from URL path
	path := r.URL.Path
	merchantID := ""
	if len(path) > len("/api/v1/merchants/") {
		merchantID = path[len("/api/v1/merchants/"):]
	}

	if merchantID == "" {
		http.Error(w, "Merchant ID is required", http.StatusBadRequest)
		return
	}

	// Return mock merchant detail
	merchant := map[string]interface{}{
		"id":                  merchantID,
		"name":                "Acme Corporation",
		"legal_name":          "Acme Corporation Inc.",
		"registration_number": "REG-123456",
		"tax_id":              "TAX-789012",
		"industry":            "Technology",
		"industry_code":       "541511",
		"business_type":       "Corporation",
		"founded_date":        "2020-01-15",
		"employee_count":      150,
		"annual_revenue":      5000000.0,
		"portfolio_type":      "onboarded",
		"risk_level":          "low",
		"compliance_status":   "compliant",
		"status":              "active",
		"address": map[string]interface{}{
			"street1":      "123 Tech Street",
			"street2":      "Suite 100",
			"city":         "San Francisco",
			"state":        "CA",
			"postal_code":  "94105",
			"country":      "United States",
			"country_code": "US",
		},
		"contact_info": map[string]interface{}{
			"phone":           "+1-555-123-4567",
			"email":           "contact@acme.com",
			"website":         "https://www.acme.com",
			"primary_contact": "John Smith",
		},
		"created_at": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(merchant)
}

// handleMerchantsSearch handles POST /api/v1/merchants/search
func (s *SimplifiedServer) handleMerchantsSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse search request
	var searchReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&searchReq); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Return mock search results
	response := map[string]interface{}{
		"merchants": []map[string]interface{}{
			{
				"id":                "merchant-001",
				"name":              "Acme Corporation",
				"legal_name":        "Acme Corporation Inc.",
				"industry":          "Technology",
				"portfolio_type":    "onboarded",
				"risk_level":        "low",
				"compliance_status": "compliant",
				"status":            "active",
			},
		},
		"total":     1,
		"page":      1,
		"page_size": 10,
		"has_more":  false,
	}

	json.NewEncoder(w).Encode(response)
}

// handleMerchantsAnalytics handles GET /api/v1/merchants/analytics
func (s *SimplifiedServer) handleMerchantsAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	analytics := map[string]interface{}{
		"total_merchants":       150,
		"active_merchants":      120,
		"pending_merchants":     20,
		"deactivated_merchants": 10,
		"portfolio_types": map[string]int{
			"onboarded":   100,
			"prospective": 30,
			"pending":     15,
			"deactivated": 5,
		},
		"risk_levels": map[string]int{
			"low":    80,
			"medium": 50,
			"high":   20,
		},
		"industries": map[string]int{
			"Technology": 60,
			"Retail":     40,
			"Finance":    30,
			"Healthcare": 20,
		},
		"compliance_status": map[string]int{
			"compliant":       100,
			"pending":         30,
			"review_required": 15,
			"non_compliant":   5,
		},
	}

	json.NewEncoder(w).Encode(analytics)
}

// handlePortfolioTypes handles GET /api/v1/merchants/portfolio-types
func (s *SimplifiedServer) handlePortfolioTypes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	portfolioTypes := []map[string]interface{}{
		{
			"id":          "onboarded",
			"type":        "onboarded",
			"name":        "Onboarded",
			"description": "Fully onboarded and active merchants",
			"color":       "#27ae60",
		},
		{
			"id":          "prospective",
			"type":        "prospective",
			"name":        "Prospective",
			"description": "Potential merchants under evaluation",
			"color":       "#3498db",
		},
		{
			"id":          "pending",
			"type":        "pending",
			"name":        "Pending",
			"description": "Merchants awaiting approval",
			"color":       "#f39c12",
		},
		{
			"id":          "deactivated",
			"type":        "deactivated",
			"name":        "Deactivated",
			"description": "Deactivated or suspended merchants",
			"color":       "#e74c3c",
		},
	}

	json.NewEncoder(w).Encode(portfolioTypes)
}

// handleRiskLevels handles GET /api/v1/merchants/risk-levels
func (s *SimplifiedServer) handleRiskLevels(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	riskLevels := []map[string]interface{}{
		{
			"id":          "low",
			"level":       "low",
			"name":        "Low Risk",
			"description": "Low risk merchants with good compliance history",
			"color":       "#27ae60",
		},
		{
			"id":          "medium",
			"level":       "medium",
			"name":        "Medium Risk",
			"description": "Medium risk merchants requiring regular monitoring",
			"color":       "#f39c12",
		},
		{
			"id":          "high",
			"level":       "high",
			"name":        "High Risk",
			"description": "High risk merchants requiring enhanced due diligence",
			"color":       "#e74c3c",
		},
	}

	json.NewEncoder(w).Encode(riskLevels)
}

// handleMerchantsStatistics handles GET /api/v1/merchants/statistics
func (s *SimplifiedServer) handleMerchantsStatistics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	statistics := map[string]interface{}{
		"total_merchants":       150,
		"active_merchants":      120,
		"pending_merchants":     20,
		"deactivated_merchants": 10,
		"average_revenue":       2500000.0,
		"total_employees":       15000,
		"compliance_rate":       0.85,
		"risk_distribution": map[string]float64{
			"low":    0.53,
			"medium": 0.33,
			"high":   0.14,
		},
		"monthly_growth": 0.05,
		"last_updated":   time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(statistics)
}

// Start starts the server
func (s *SimplifiedServer) Start() error {
	s.logger.Printf("🚀 Starting PRODUCTION Fixed Server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *SimplifiedServer) Stop(ctx context.Context) error {
	s.logger.Printf("🛑 Stopping Fixed Server...")

	// Stop the database module
	if err := s.databaseModule.Stop(ctx); err != nil {
		s.logger.Printf("⚠️ Error stopping database module: %v", err)
	}

	// Shutdown HTTP server
	return s.server.Shutdown(ctx)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewSimplifiedServer(port)

	// Handle graceful shutdown
	go func() {
		<-make(chan os.Signal, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Stop(ctx)
	}()

	// Start server
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
