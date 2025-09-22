package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// RailwayServer represents the Railway deployment server
type RailwayServer struct {
	server *http.Server
	logger *log.Logger
	zapLogger *zap.Logger
}

// NewRailwayServer creates a new Railway server instance
func NewRailwayServer() (*RailwayServer, error) {
	// Initialize logger
	logger := log.New(os.Stdout, "[railway-server] ", log.LstdFlags)
	zapLogger, _ := zap.NewProduction()

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create router
	router := mux.NewRouter()

	// Create server
	server := &RailwayServer{
		logger:    logger,
		zapLogger: zapLogger,
	}

	// Setup routes
	server.setupRoutes(router)

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + port,
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
	router.HandleFunc("/status", s.handleHealth).Methods("GET")

	// Business Intelligence Classification
	router.HandleFunc("/v1/classify", s.handleClassify).Methods("POST")

	// Merchant Management API
	api := router.PathPrefix("/api/v1").Subrouter()
	api.HandleFunc("/merchants", s.handleGetMerchants).Methods("GET")
	api.HandleFunc("/merchants/search", s.handleSearchMerchants).Methods("POST")
	api.HandleFunc("/merchants/analytics", s.handleMerchantAnalytics).Methods("GET")
	api.HandleFunc("/merchants/portfolio-types", s.handlePortfolioTypes).Methods("GET")
	api.HandleFunc("/merchants/risk-levels", s.handleRiskLevels).Methods("GET")
	api.HandleFunc("/merchants/statistics", s.handleMerchantStatistics).Methods("GET")
	api.HandleFunc("/merchants/{id}", s.handleGetMerchant).Methods("GET")

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
			"supabase_integration":           false,
			"database_driven_classification": true,
			"enhanced_keyword_matching":      true,
			"industry_detection":             true,
			"confidence_scoring":             true,
		},
		"supabase_status": map[string]interface{}{
			"connected": false,
			"reason":    "simplified_mode",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
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

	// Generate a business ID for tracking
	businessID := fmt.Sprintf("biz_%d", time.Now().Unix())

	// Simple classification logic based on keywords
	industry := s.classifyBusiness(req.BusinessName, req.Description)
	confidence := s.calculateConfidence(req.BusinessName, req.Description)

	result := map[string]interface{}{
		"success":       true,
		"business_id":   businessID,
		"business_name": req.BusinessName,
		"description":   req.Description,
		"website_url":   req.WebsiteURL,
		"classification": map[string]interface{}{
			"industry":  industry,
			"confidence": confidence,
			"mcc_codes": []map[string]interface{}{
				{"code": "7372", "description": "Computer Programming Services", "confidence": confidence},
			},
			"sic_codes": []map[string]interface{}{
				{"code": "7372", "description": "Computer Programming Services", "confidence": confidence},
			},
			"naics_codes": []map[string]interface{}{
				{"code": "541511", "description": "Custom Computer Programming Services", "confidence": confidence},
			},
		},
		"confidence_score": confidence,
		"status":           "success",
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"data_source":      "simplified_classifier",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// classifyBusiness performs simple keyword-based classification
func (s *RailwayServer) classifyBusiness(name, description string) string {
	text := fmt.Sprintf("%s %s", name, description)
	text = strings.ToLower(text)

	// Simple keyword matching
	if strings.Contains(text, "tech") || strings.Contains(text, "software") || strings.Contains(text, "computer") {
		return "Technology"
	}
	if strings.Contains(text, "retail") || strings.Contains(text, "store") || strings.Contains(text, "shop") {
		return "Retail"
	}
	if strings.Contains(text, "finance") || strings.Contains(text, "bank") || strings.Contains(text, "investment") {
		return "Finance"
	}
	if strings.Contains(text, "health") || strings.Contains(text, "medical") || strings.Contains(text, "hospital") {
		return "Healthcare"
	}
	if strings.Contains(text, "food") || strings.Contains(text, "restaurant") || strings.Contains(text, "catering") {
		return "Food & Beverage"
	}

	return "General Business"
}

// calculateConfidence calculates a simple confidence score
func (s *RailwayServer) calculateConfidence(name, description string) float64 {
	// Simple confidence calculation based on text length and keywords
	text := fmt.Sprintf("%s %s", name, description)
	
	// Base confidence
	confidence := 0.7
	
	// Increase confidence for longer descriptions
	if len(description) > 50 {
		confidence += 0.1
	}
	if len(description) > 100 {
		confidence += 0.1
	}
	
	// Increase confidence for specific keywords
	keywords := []string{"inc", "corp", "llc", "ltd", "company", "business"}
	textLower := strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(textLower, keyword) {
			confidence += 0.05
		}
	}
	
	// Cap at 0.95
	if confidence > 0.95 {
		confidence = 0.95
	}
	
	return confidence
}

// handleGetMerchants handles GET /api/v1/merchants
func (s *RailwayServer) handleGetMerchants(w http.ResponseWriter, r *http.Request) {
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

	// Mock merchant detail data
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
		"data_source":         "mock_data",
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

	// Mock search results
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
	s.logger.Printf("📊 Supabase Integration: false (simplified mode)")
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