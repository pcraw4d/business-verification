package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/supabase-community/supabase-go"
	"go.uber.org/zap"
)

// RailwayServer represents the Railway deployment server
type RailwayServer struct {
	server         *http.Server
	supabaseClient *supabase.Client
	logger         *log.Logger
	zapLogger      *zap.Logger
}

// NewRailwayServer creates a new Railway server instance
func NewRailwayServer() (*RailwayServer, error) {
	// Initialize logger
	logger := log.New(os.Stdout, "[railway-server] ", log.LstdFlags)
	zapLogger, _ := zap.NewProduction()

	// Initialize Supabase client
	var supabaseClient *supabase.Client
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	if supabaseURL != "" && supabaseKey != "" {
		client, err := supabase.NewClient(supabaseURL, supabaseKey, nil)
		if err != nil {
			logger.Printf("⚠️ Warning: Failed to initialize Supabase client: %v", err)
			supabaseClient = nil
		} else {
			supabaseClient = client
			logger.Printf("✅ Successfully initialized Supabase client")
		}
	} else {
		logger.Printf("⚠️ Supabase configuration incomplete - using fallback mode")
		logger.Printf("📝 Required: SUPABASE_URL, SUPABASE_ANON_KEY")
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create router
	router := mux.NewRouter()

	// Create server
	server := &RailwayServer{
		supabaseClient: supabaseClient,
		logger:         logger,
		zapLogger:      zapLogger,
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

	return map[string]interface{}{
		"connected": true,
		"url":       os.Getenv("SUPABASE_URL"),
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

	// Process classification using Supabase if available
	var result map[string]interface{}
	if s.supabaseClient != nil {
		// Try to use Supabase for classification
		result = s.processClassificationWithSupabase(req.BusinessName, req.Description, req.WebsiteURL)
	} else {
		// Fallback to mock classification
		result = s.getFallbackClassification(req.BusinessName, req.Description, req.WebsiteURL)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// processClassificationWithSupabase processes classification using Supabase
func (s *RailwayServer) processClassificationWithSupabase(businessName, description, websiteURL string) map[string]interface{} {
	// Generate a business ID for tracking
	businessID := fmt.Sprintf("biz_%d", time.Now().Unix())

	// Try to query classification data from Supabase
	var classifications []map[string]interface{}
	_, err := s.supabaseClient.From("classifications").Select("*", "", false).Eq("business_name", businessName).ExecuteTo(&classifications)

	if err != nil || len(classifications) == 0 {
		// If no existing classification, create a new one
		s.logger.Printf("📝 No existing classification found, creating new one")
		return s.createNewClassification(businessName, description, websiteURL, businessID)
	}

	// Return existing classification
	classification := classifications[0]
	return map[string]interface{}{
		"success":          true,
		"business_id":      businessID,
		"business_name":    businessName,
		"description":      description,
		"website_url":      websiteURL,
		"classification":   classification,
		"confidence_score": 0.95,
		"status":           "success",
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"data_source":      "supabase",
	}
}

// createNewClassification creates a new classification and stores it in Supabase
func (s *RailwayServer) createNewClassification(businessName, description, websiteURL, businessID string) map[string]interface{} {
	// Enhanced classification with website scraping and risk detection
	industry := s.classifyBusiness(businessName, description)
	confidence := s.calculateConfidence(businessName, description)

	// Scrape website content if URL provided
	var websiteContent string
	var scrapedKeywords []string
	if websiteURL != "" {
		websiteContent, scrapedKeywords = s.scrapeWebsite(websiteURL)
		s.logger.Printf("🌐 Scraped website content: %d characters, %d keywords", len(websiteContent), len(scrapedKeywords))
	}

	// Combine all text for risk analysis
	allText := fmt.Sprintf("%s %s %s", businessName, description, websiteContent)

	// Perform risk assessment
	riskAssessment := s.performRiskAssessment(businessName, allText, scrapedKeywords)

	// Log risk detection results
	if riskAssessment["risk_level"] != "low" {
		s.logger.Printf("⚠️ Risk detected: %s (score: %.2f) - %s",
			riskAssessment["risk_level"],
			riskAssessment["risk_score"],
			riskAssessment["risk_factors"])
	}

	classification := map[string]interface{}{
		"mcc_codes": []map[string]interface{}{
			{"code": "7372", "description": "Computer Programming Services", "confidence": confidence},
		},
		"sic_codes": []map[string]interface{}{
			{"code": "7372", "description": "Computer Programming Services", "confidence": confidence},
		},
		"naics_codes": []map[string]interface{}{
			{"code": "541511", "description": "Custom Computer Programming Services", "confidence": confidence},
		},
		"industry":        industry,
		"risk_assessment": riskAssessment,
		"website_content": map[string]interface{}{
			"scraped":        len(websiteContent) > 0,
			"content_length": len(websiteContent),
			"keywords_found": len(scrapedKeywords),
		},
	}

	// Try to store in Supabase
	newClassification := map[string]interface{}{
		"business_id":      businessID,
		"business_name":    businessName,
		"description":      description,
		"website_url":      websiteURL,
		"classification":   classification,
		"confidence_score": confidence,
		"created_at":       time.Now().UTC().Format(time.RFC3339),
	}

	_, _, err := s.supabaseClient.From("classifications").Insert(newClassification, false, "", "", "").Execute()
	if err != nil {
		s.logger.Printf("⚠️ Failed to store classification in Supabase: %v", err)
	}

	// Store risk assessment separately
	s.storeRiskAssessment(businessID, businessName, riskAssessment)

	return map[string]interface{}{
		"success":          true,
		"business_id":      businessID,
		"business_name":    businessName,
		"description":      description,
		"website_url":      websiteURL,
		"classification":   classification,
		"confidence_score": confidence,
		"risk_assessment":  riskAssessment,
		"status":           "success",
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"data_source":      "supabase_new",
	}
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

// getFallbackClassification returns mock classification data
func (s *RailwayServer) getFallbackClassification(businessName, description, websiteURL string) map[string]interface{} {
	// Generate a business ID for tracking
	businessID := fmt.Sprintf("biz_%d", time.Now().Unix())

	// Simple classification logic
	industry := s.classifyBusiness(businessName, description)
	confidence := s.calculateConfidence(businessName, description)

	return map[string]interface{}{
		"success":       true,
		"business_id":   businessID,
		"business_name": businessName,
		"description":   description,
		"website_url":   websiteURL,
		"classification": map[string]interface{}{
			"mcc_codes": []map[string]interface{}{
				{"code": "7372", "description": "Computer Programming Services", "confidence": confidence},
			},
			"sic_codes": []map[string]interface{}{
				{"code": "7372", "description": "Computer Programming Services", "confidence": confidence},
			},
			"naics_codes": []map[string]interface{}{
				{"code": "541511", "description": "Custom Computer Programming Services", "confidence": confidence},
			},
			"industry": industry,
		},
		"confidence_score": confidence,
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
	var merchants []map[string]interface{}
	_, err := s.supabaseClient.From("merchants").Select("*", "", false).ExecuteTo(&merchants)
	if err != nil {
		// If merchants table doesn't exist, try mock_merchants
		_, err2 := s.supabaseClient.From("mock_merchants").Select("*", "", false).ExecuteTo(&merchants)
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
	var merchants []map[string]interface{}
	_, err := s.supabaseClient.From("merchants").Select("*", "", false).Eq("id", merchantID).ExecuteTo(&merchants)
	if err != nil {
		// If merchants table doesn't exist, try mock_merchants
		_, err2 := s.supabaseClient.From("mock_merchants").Select("*", "", false).Eq("id", merchantID).ExecuteTo(&merchants)
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
		s.logger.Printf("🔗 Supabase URL: %s", os.Getenv("SUPABASE_URL"))
	}
	return s.server.ListenAndServe()
}

// Stop gracefully stops the server
func (s *RailwayServer) Stop(ctx context.Context) error {
	s.logger.Printf("🛑 Stopping RAILWAY SERVER...")
	return s.server.Shutdown(ctx)
}

// scrapeWebsite scrapes content from a website URL
func (s *RailwayServer) scrapeWebsite(url string) (string, []string) {
	// Add http:// if no protocol specified
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Make request
	resp, err := client.Get(url)
	if err != nil {
		s.logger.Printf("⚠️ Failed to scrape website %s: %v", url, err)
		return "", []string{}
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Printf("⚠️ Failed to read website content: %v", err)
		return "", []string{}
	}

	// Extract text content (simple HTML tag removal)
	content := string(body)
	content = s.extractTextFromHTML(content)

	// Extract keywords (simple approach)
	keywords := s.extractKeywords(content)

	s.logger.Printf("🌐 Successfully scraped %s: %d characters, %d keywords", url, len(content), len(keywords))

	return content, keywords
}

// extractTextFromHTML removes HTML tags and extracts text content
func (s *RailwayServer) extractTextFromHTML(html string) string {
	// Remove HTML tags
	re := regexp.MustCompile(`<[^>]*>`)
	text := re.ReplaceAllString(html, " ")

	// Remove extra whitespace
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	// Remove common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")

	return strings.TrimSpace(text)
}

// extractKeywords extracts relevant keywords from text
func (s *RailwayServer) extractKeywords(text string) []string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Define business-related keywords to look for
	businessKeywords := []string{
		"business", "company", "corp", "inc", "llc", "ltd", "services", "solutions",
		"technology", "software", "tech", "digital", "online", "web", "app",
		"finance", "banking", "investment", "trading", "crypto", "bitcoin",
		"retail", "store", "shop", "ecommerce", "marketplace", "sales",
		"healthcare", "medical", "health", "hospital", "clinic", "pharmacy",
		"food", "restaurant", "catering", "delivery", "dining",
		"education", "school", "university", "training", "learning",
		"real estate", "property", "construction", "building", "development",
		"consulting", "advisory", "professional", "expert", "specialist",
	}

	var foundKeywords []string
	for _, keyword := range businessKeywords {
		if strings.Contains(text, keyword) {
			foundKeywords = append(foundKeywords, keyword)
		}
	}

	return foundKeywords
}

// performRiskAssessment performs comprehensive risk assessment
func (s *RailwayServer) performRiskAssessment(businessName, allText string, scrapedKeywords []string) map[string]interface{} {
	// Get risk keywords from database
	riskKeywords := s.getRiskKeywords()

	// Analyze text for risk indicators
	riskScore := 0.0
	var detectedRisks []string
	var riskFactors map[string]interface{}

	// Check for prohibited keywords
	prohibitedKeywords := s.checkProhibitedKeywords(allText, riskKeywords)
	if len(prohibitedKeywords) > 0 {
		riskScore += 0.4
		detectedRisks = append(detectedRisks, "prohibited_keywords")
		riskFactors = map[string]interface{}{
			"prohibited_keywords": prohibitedKeywords,
			"risk_category":       "prohibited",
		}
	}

	// Check for high-risk keywords
	highRiskKeywords := s.checkHighRiskKeywords(allText, riskKeywords)
	if len(highRiskKeywords) > 0 {
		riskScore += 0.3
		detectedRisks = append(detectedRisks, "high_risk_keywords")
		if riskFactors == nil {
			riskFactors = map[string]interface{}{}
		}
		riskFactors["high_risk_keywords"] = highRiskKeywords
		riskFactors["risk_category"] = "high_risk"
	}

	// Check for fraud indicators
	fraudKeywords := s.checkFraudKeywords(allText, riskKeywords)
	if len(fraudKeywords) > 0 {
		riskScore += 0.5
		detectedRisks = append(detectedRisks, "fraud_indicators")
		if riskFactors == nil {
			riskFactors = map[string]interface{}{}
		}
		riskFactors["fraud_indicators"] = fraudKeywords
		riskFactors["risk_category"] = "fraud"
	}

	// Check business name patterns
	nameRisk := s.analyzeBusinessNameRisk(businessName)
	riskScore += nameRisk

	// Determine risk level
	riskLevel := s.determineRiskLevel(riskScore)

	// Set default risk factors if none detected
	if riskFactors == nil {
		riskFactors = map[string]interface{}{
			"industry":   "general",
			"geographic": "low_risk",
			"regulatory": "compliant",
		}
	}

	return map[string]interface{}{
		"risk_score":                riskScore,
		"risk_level":                riskLevel,
		"risk_factors":              riskFactors,
		"detected_risks":            detectedRisks,
		"prohibited_keywords_found": prohibitedKeywords,
		"assessment_methodology":    "automated",
		"assessment_timestamp":      time.Now().UTC().Format(time.RFC3339),
	}
}

// getRiskKeywords retrieves risk keywords from database
func (s *RailwayServer) getRiskKeywords() []map[string]interface{} {
	if s.supabaseClient == nil {
		// Return default risk keywords if no database connection
		return []map[string]interface{}{
			{"keyword": "gambling", "risk_category": "prohibited", "risk_severity": "high", "risk_score_weight": 1.5},
			{"keyword": "casino", "risk_category": "prohibited", "risk_severity": "high", "risk_score_weight": 1.5},
			{"keyword": "betting", "risk_category": "prohibited", "risk_severity": "high", "risk_score_weight": 1.5},
			{"keyword": "cryptocurrency", "risk_category": "high_risk", "risk_severity": "medium", "risk_score_weight": 1.3},
			{"keyword": "bitcoin", "risk_category": "high_risk", "risk_severity": "medium", "risk_score_weight": 1.3},
			{"keyword": "trading", "risk_category": "high_risk", "risk_severity": "low", "risk_score_weight": 1.1},
			{"keyword": "scam", "risk_category": "fraud", "risk_severity": "high", "risk_score_weight": 1.6},
		}
	}

	var riskKeywords []map[string]interface{}
	_, err := s.supabaseClient.From("risk_keywords").Select("*", "", false).ExecuteTo(&riskKeywords)
	if err != nil {
		s.logger.Printf("⚠️ Failed to get risk keywords from database: %v", err)
		return []map[string]interface{}{}
	}

	return riskKeywords
}

// checkProhibitedKeywords checks for prohibited business keywords
func (s *RailwayServer) checkProhibitedKeywords(text string, riskKeywords []map[string]interface{}) []string {
	var found []string
	textLower := strings.ToLower(text)

	for _, keyword := range riskKeywords {
		if category, ok := keyword["risk_category"].(string); ok && category == "prohibited" {
			if kw, ok := keyword["keyword"].(string); ok {
				if strings.Contains(textLower, strings.ToLower(kw)) {
					found = append(found, kw)
				}
			}
		}
	}

	return found
}

// checkHighRiskKeywords checks for high-risk business keywords
func (s *RailwayServer) checkHighRiskKeywords(text string, riskKeywords []map[string]interface{}) []string {
	var found []string
	textLower := strings.ToLower(text)

	for _, keyword := range riskKeywords {
		if category, ok := keyword["risk_category"].(string); ok && category == "high_risk" {
			if kw, ok := keyword["keyword"].(string); ok {
				if strings.Contains(textLower, strings.ToLower(kw)) {
					found = append(found, kw)
				}
			}
		}
	}

	return found
}

// checkFraudKeywords checks for fraud-related keywords
func (s *RailwayServer) checkFraudKeywords(text string, riskKeywords []map[string]interface{}) []string {
	var found []string
	textLower := strings.ToLower(text)

	for _, keyword := range riskKeywords {
		if category, ok := keyword["risk_category"].(string); ok && category == "fraud" {
			if kw, ok := keyword["keyword"].(string); ok {
				if strings.Contains(textLower, strings.ToLower(kw)) {
					found = append(found, kw)
				}
			}
		}
	}

	return found
}

// analyzeBusinessNameRisk analyzes business name for risk patterns
func (s *RailwayServer) analyzeBusinessNameRisk(businessName string) float64 {
	nameLower := strings.ToLower(businessName)
	riskScore := 0.0

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"investment", "trading", "crypto", "bitcoin", "forex", "gambling", "casino",
		"betting", "lottery", "scam", "fraud", "ponzi", "pyramid",
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(nameLower, pattern) {
			riskScore += 0.1
		}
	}

	// Check for generic names (potential shell companies)
	genericNames := []string{
		"holdings", "investments", "ventures", "capital", "fund", "group",
		"enterprises", "international", "global", "worldwide",
	}

	genericCount := 0
	for _, generic := range genericNames {
		if strings.Contains(nameLower, generic) {
			genericCount++
		}
	}

	if genericCount >= 2 {
		riskScore += 0.2
	}

	return riskScore
}

// determineRiskLevel determines risk level based on score
func (s *RailwayServer) determineRiskLevel(riskScore float64) string {
	if riskScore >= 0.8 {
		return "critical"
	} else if riskScore >= 0.6 {
		return "high"
	} else if riskScore >= 0.3 {
		return "medium"
	} else {
		return "low"
	}
}

// storeRiskAssessment stores risk assessment in database
func (s *RailwayServer) storeRiskAssessment(businessID, businessName string, riskAssessment map[string]interface{}) {
	if s.supabaseClient == nil {
		return
	}

	// Prepare risk assessment data
	riskData := map[string]interface{}{
		"business_id":               businessID,
		"business_name":             businessName,
		"risk_score":                riskAssessment["risk_score"],
		"risk_level":                riskAssessment["risk_level"],
		"risk_factors":              riskAssessment["risk_factors"],
		"prohibited_keywords_found": riskAssessment["prohibited_keywords_found"],
		"assessment_methodology":    riskAssessment["assessment_methodology"],
		"created_at":                time.Now().UTC().Format(time.RFC3339),
	}

	// Store in database
	_, _, err := s.supabaseClient.From("business_risk_assessments").Insert(riskData, false, "", "", "").Execute()
	if err != nil {
		s.logger.Printf("⚠️ Failed to store risk assessment: %v", err)
	} else {
		s.logger.Printf("✅ Risk assessment stored for %s: %s (score: %.2f)",
			businessName, riskAssessment["risk_level"], riskAssessment["risk_score"])
	}
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
