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
	
	// Debug endpoint to check web directory
	router.HandleFunc("/debug/web", s.handleDebugWeb).Methods("GET")

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

	// Serve static files from web directory
	// Create a file server for the web directory
	fileServer := http.FileServer(http.Dir("./web/"))

	// Serve static files with debugging
	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Printf("📁 Serving static file: %s", r.URL.Path)
		fileServer.ServeHTTP(w, r)
	}))
}

// handleDebugWeb handles debug requests to check web directory
func (s *RailwayServer) handleDebugWeb(w http.ResponseWriter, r *http.Request) {
	// Check if web directory exists
	webDir := "./web"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "web directory does not exist",
			"path":  webDir,
		})
		return
	}
	
	// List files in web directory
	files, err := os.ReadDir(webDir)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "failed to read web directory",
			"path":  webDir,
			"err":   err.Error(),
		})
		return
	}
	
	var fileList []string
	for _, file := range files {
		fileList = append(fileList, file.Name())
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"web_directory": webDir,
		"files":         fileList,
		"count":         len(fileList),
	})
}

// handleHealth handles health check requests
func (s *RailwayServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"version":   "3.2.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"features": map[string]bool{
			"confidence_scoring":             true,
			"database_driven_classification": true,
			"enhanced_keyword_matching":      true,
			"industry_detection":             true,
			"supabase_integration":           s.supabaseClient != nil,
		},
	}

	if s.supabaseClient != nil {
		health["supabase_status"] = map[string]interface{}{
			"connected": true,
			"url":       os.Getenv("SUPABASE_URL"),
		}
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

	// Scrape website content if URL provided using enhanced scraper
	var websiteContent string
	var scrapedKeywords []string
	if websiteURL != "" {
		websiteContent, scrapedKeywords = s.scrapeWebsite(websiteURL)
		s.logger.Printf("🌐 Enhanced scraper extracted %d characters, %d keywords from %s",
			len(websiteContent), len(scrapedKeywords), websiteURL)
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

// getFallbackClassification returns mock classification data with enhanced website scraping
func (s *RailwayServer) getFallbackClassification(businessName, description, websiteURL string) map[string]interface{} {
	// Generate a business ID for tracking
	businessID := fmt.Sprintf("biz_%d", time.Now().Unix())

	// Simple classification logic
	industry := s.classifyBusiness(businessName, description)
	confidence := s.calculateConfidence(businessName, description)

	// Scrape website content if URL provided using enhanced scraper
	var websiteContent string
	var scrapedKeywords []string
	if websiteURL != "" {
		websiteContent, scrapedKeywords = s.scrapeWebsite(websiteURL)
		s.logger.Printf("🌐 Enhanced scraper (fallback) extracted %d characters, %d keywords from %s",
			len(websiteContent), len(scrapedKeywords), websiteURL)
	}

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
			"website_content": map[string]interface{}{
				"scraped":        len(websiteContent) > 0,
				"content_length": len(websiteContent),
				"keywords_found": len(scrapedKeywords),
			},
		},
		"confidence_score": confidence,
		"status":           "success",
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
		"data_source":      "fallback_mock",
	}
}

// scrapeWebsite scrapes content from a website URL with enhanced features
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
	// Simple keyword extraction
	words := strings.Fields(strings.ToLower(text))

	// Filter out common words and keep business-relevant terms
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"this": true, "that": true, "these": true, "those": true, "i": true, "you": true,
		"he": true, "she": true, "it": true, "we": true, "they": true, "me": true,
		"him": true, "her": true, "us": true, "them": true, "my": true, "your": true,
		"his": true, "its": true, "our": true, "their": true,
	}

	var keywords []string
	for _, word := range words {
		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	// Limit to first 10 keywords
	if len(keywords) > 10 {
		keywords = keywords[:10]
	}

	return keywords
}

// performRiskAssessment performs risk assessment on business data
func (s *RailwayServer) performRiskAssessment(businessName, allText string, keywords []string) map[string]interface{} {
	// Simple risk assessment logic
	riskScore := 0.0
	riskLevel := "low"
	riskFactors := map[string]string{
		"geographic": "low_risk",
		"industry":   "general",
		"regulatory": "compliant",
	}

	// Check for high-risk keywords
	highRiskKeywords := []string{"crypto", "bitcoin", "gambling", "casino", "adult", "weapon"}
	textLower := strings.ToLower(allText)

	for _, keyword := range highRiskKeywords {
		if strings.Contains(textLower, keyword) {
			riskScore += 0.3
		}
	}

	// Determine risk level
	if riskScore > 0.7 {
		riskLevel = "high"
	} else if riskScore > 0.3 {
		riskLevel = "medium"
	}

	return map[string]interface{}{
		"risk_level":                riskLevel,
		"risk_score":                riskScore,
		"risk_factors":              riskFactors,
		"detected_risks":            nil,
		"prohibited_keywords_found": nil,
		"assessment_methodology":    "automated",
		"assessment_timestamp":      time.Now().UTC().Format(time.RFC3339),
	}
}

// storeRiskAssessment stores risk assessment data
func (s *RailwayServer) storeRiskAssessment(businessID, businessName string, riskAssessment map[string]interface{}) {
	// Store risk assessment in Supabase if available
	if s.supabaseClient != nil {
		riskData := map[string]interface{}{
			"business_id":     businessID,
			"business_name":   businessName,
			"risk_level":      riskAssessment["risk_level"],
			"risk_score":      riskAssessment["risk_score"],
			"risk_factors":    riskAssessment["risk_factors"],
			"assessment_time": time.Now().UTC().Format(time.RFC3339),
		}

		_, _, err := s.supabaseClient.From("risk_assessments").Insert(riskData, false, "", "", "").Execute()
		if err != nil {
			s.logger.Printf("⚠️ Failed to store risk assessment: %v", err)
		} else {
			s.logger.Printf("✅ Risk assessment stored for %s: %s (score: %.2f)",
				businessName, riskAssessment["risk_level"], riskAssessment["risk_score"])
		}
	}
}

// Additional handler methods for merchant management API
func (s *RailwayServer) handleGetMerchants(w http.ResponseWriter, r *http.Request) {
	// Mock merchant data
	merchants := []map[string]interface{}{
		{
			"id":         "merchant_1",
			"name":       "TechCorp Solutions",
			"industry":   "Technology",
			"risk_level": "low",
			"status":     "active",
			"created_at": time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
		},
		{
			"id":         "merchant_2",
			"name":       "Retail Store Inc",
			"industry":   "Retail",
			"risk_level": "medium",
			"status":     "active",
			"created_at": time.Now().AddDate(0, -2, 0).Format(time.RFC3339),
		},
	}

	response := map[string]interface{}{
		"merchants": merchants,
		"total":     len(merchants),
		"page":      1,
		"limit":     10,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *RailwayServer) handleSearchMerchants(w http.ResponseWriter, r *http.Request) {
	// Mock search functionality
	response := map[string]interface{}{
		"merchants": []map[string]interface{}{},
		"total":     0,
		"page":      1,
		"limit":     10,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *RailwayServer) handleMerchantAnalytics(w http.ResponseWriter, r *http.Request) {
	analytics := map[string]interface{}{
		"total_merchants":   150,
		"active_merchants":  142,
		"pending_merchants": 8,
		"risk_distribution": map[string]int{
			"low":    120,
			"medium": 25,
			"high":   5,
		},
		"industry_breakdown": map[string]int{
			"Technology": 45,
			"Retail":     35,
			"Finance":    25,
			"Healthcare": 20,
			"Other":      25,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(analytics)
}

func (s *RailwayServer) handlePortfolioTypes(w http.ResponseWriter, r *http.Request) {
	portfolioTypes := []map[string]interface{}{
		{"id": "enterprise", "name": "Enterprise", "count": 45},
		{"id": "sme", "name": "Small & Medium Enterprise", "count": 78},
		{"id": "startup", "name": "Startup", "count": 27},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(portfolioTypes)
}

func (s *RailwayServer) handleRiskLevels(w http.ResponseWriter, r *http.Request) {
	riskLevels := []map[string]interface{}{
		{"id": "low", "name": "Low Risk", "count": 120},
		{"id": "medium", "name": "Medium Risk", "count": 25},
		{"id": "high", "name": "High Risk", "count": 5},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(riskLevels)
}

func (s *RailwayServer) handleGetMerchant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	merchantID := vars["id"]

	merchant := map[string]interface{}{
		"id":          merchantID,
		"name":        "Sample Merchant",
		"industry":    "Technology",
		"risk_level":  "low",
		"status":      "active",
		"created_at":  time.Now().AddDate(0, -1, 0).Format(time.RFC3339),
		"description": "A sample merchant for testing purposes",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(merchant)
}

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

func main() {
	server, err := NewRailwayServer()
	if err != nil {
		log.Fatal("Failed to create server:", err)
	}

	log.Fatal(server.Start())
}
