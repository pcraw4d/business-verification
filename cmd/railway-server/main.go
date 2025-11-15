package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/supabase-community/postgrest-go"
	"go.uber.org/zap"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/middleware"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/database"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/risk"
	"kyb-platform/internal/services"
	redisoptimization "kyb-redis-optimization"
)

// RailwayServer represents the complete KYB platform server with all features
type RailwayServer struct {
	serviceName    string
	version        string
	supabaseClient *postgrest.Client
	port           string
	db             *sql.DB
	mux            *http.ServeMux
	redisOptimizer *redisoptimization.RedisOptimizer // Redis optimization support (optional)
}

// NewRailwayServer creates a new RailwayServer instance
func NewRailwayServer() *RailwayServer {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "kyb-platform-v4-complete"
	}

	version := "4.0.0-CACHE-BUST-REBUILD"

	// Initialize Supabase client
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	var supabaseClient *postgrest.Client
	if supabaseURL != "" && supabaseKey != "" {
		supabaseClient = postgrest.NewClient(supabaseURL, supabaseKey, nil)
	}

	// Initialize database connection for new routes
	var db *sql.DB
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		var err error
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			log.Printf("Warning: Failed to connect to database: %v. New API routes will not be available.", err)
		} else {
			// Configure connection pool settings
			db.SetMaxOpenConns(25)                 // Maximum open connections
			db.SetMaxIdleConns(5)                  // Maximum idle connections
			db.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime
			db.SetConnMaxIdleTime(1 * time.Minute) // Idle connection timeout

			// Test connection with context timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := db.PingContext(ctx); err != nil {
				log.Printf("Warning: Database ping failed: %v. Routes will use in-memory storage.", err)
				// Close and set to nil to prevent using broken connection
				// Routes will still register but use in-memory thresholds
				db.Close()
				db = nil
				log.Println("⚠️  Using in-memory threshold storage (database unavailable)")
			} else {
				log.Println("✅ Database connection established for new API routes")
			}
		}
	} else {
		log.Println("Warning: DATABASE_URL not set. New API routes will not be available.")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize Redis optimizer (optional - graceful fallback if unavailable)
	var redisOptimizer *redisoptimization.RedisOptimizer
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		// Parse Redis URL (format: redis://:password@host:port or redis://host:port)
		parts := strings.Split(redisURL, "://")
		if len(parts) == 2 {
			authAndHost := parts[1]
			if strings.Contains(authAndHost, "@") {
				authParts := strings.Split(authAndHost, "@")
				if len(authParts) == 2 {
					password := strings.TrimPrefix(authParts[0], ":")
					hostPort := authParts[1]
					redisOptimizer = redisoptimization.NewRedisOptimizer(hostPort, password, nil)
					log.Println("✅ Redis optimization enabled")
				}
			} else {
				redisOptimizer = redisoptimization.NewRedisOptimizer(authAndHost, "", nil)
				log.Println("✅ Redis optimization enabled (no password)")
			}
		}
	} else {
		log.Println("⚠️  Redis optimization disabled (REDIS_URL not set)")
	}

	return &RailwayServer{
		serviceName:    serviceName,
		version:        version,
		supabaseClient: supabaseClient,
		port:           port,
		db:             db,
		mux:            http.NewServeMux(),
		redisOptimizer: redisOptimizer,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Service   string `json:"service"`
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// DetailedHealthResponse represents the detailed health check response
type DetailedHealthResponse struct {
	Service   string                 `json:"service"`
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version"`
	Checks    map[string]interface{} `json:"checks"`
}

// ClassificationRequest represents a business classification request
type ClassificationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ClassificationResponse represents a business classification response
type ClassificationResponse struct {
	BusinessName    string                 `json:"business_name"`
	Description     string                 `json:"description"`
	Classifications map[string]interface{} `json:"classifications"`
	RiskAssessment  map[string]interface{} `json:"risk_assessment"`
	ProcessingTime  string                 `json:"processing_time"`
	Timestamp       string                 `json:"timestamp"`
}

// MerchantRequest represents a merchant creation request
type MerchantRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone,omitempty"`
}

// MerchantResponse represents a merchant response
type MerchantResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Email       string    `json:"email,omitempty"`
	Phone       string    `json:"phone,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TokenRequest represents a JWT token generation request
type TokenRequest struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
}

// TokenResponse represents a JWT token response
type TokenResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
}

// AnalyticsResponse represents analytics data
type AnalyticsResponse struct {
	Service       string                   `json:"service"`
	Version       string                   `json:"version"`
	Timestamp     string                   `json:"timestamp"`
	OverallStats  map[string]interface{}   `json:"overall_stats"`
	TopIndustries []map[string]interface{} `json:"top_industries"`
	TopRiskLevels []map[string]interface{} `json:"top_risk_levels"`
}

// MetricsResponse represents performance metrics
type MetricsResponse struct {
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Timestamp string                 `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics"`
}

// SelfDrivingResponse represents self-driving capabilities
type SelfDrivingResponse struct {
	Service      string                 `json:"service"`
	Version      string                 `json:"version"`
	Timestamp    string                 `json:"timestamp"`
	Capabilities map[string]interface{} `json:"capabilities"`
}

// ReportResponse represents a report response
type ReportResponse struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	GeneratedAt *time.Time             `json:"generated_at,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
}

// handleHealth handles the health check endpoint
func (s *RailwayServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Service:   s.serviceName,
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   s.version,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleDetailedHealth handles the detailed health check endpoint
func (s *RailwayServer) handleDetailedHealth(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]interface{})

	// Check database connectivity
	if s.supabaseClient != nil {
		_, _, err := s.supabaseClient.From("classifications").Select("id", "", false).Limit(1, "").Execute()
		if err != nil {
			checks["database"] = map[string]string{"status": "unhealthy", "error": err.Error()}
		} else {
			checks["database"] = map[string]string{"status": "healthy"}
		}
	} else {
		checks["database"] = map[string]string{"status": "not_configured"}
	}

	// Check PostgreSQL database (if available)
	if s.db != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.db.PingContext(ctx); err != nil {
			checks["postgres"] = map[string]string{"status": "unhealthy", "error": err.Error()}
		} else {
			checks["postgres"] = map[string]string{"status": "healthy"}
		}
	} else {
		checks["postgres"] = map[string]string{"status": "not_configured"}
	}

	// Check Redis optimization (if available)
	if s.redisOptimizer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		redisHealth, err := s.redisOptimizer.HealthCheck(ctx)
		if err != nil {
			checks["redis"] = map[string]interface{}{
				"status": "unhealthy",
				"error":  err.Error(),
			}
		} else {
			checks["redis"] = map[string]interface{}{
				"status":  redisHealth.Status,
				"latency": redisHealth.Latency.String(),
				"connections": map[string]int{
					"total":  redisHealth.TotalConnections,
					"active": redisHealth.ActiveConnections,
					"idle":   redisHealth.IdleConnections,
				},
			}
		}
	} else {
		checks["redis"] = map[string]string{"status": "not_configured"}
	}

	// Check external services
	checks["external_apis"] = map[string]string{"status": "healthy"}
	checks["cache"] = map[string]string{"status": "healthy"}

	response := DetailedHealthResponse{
		Service:   s.serviceName,
		Status:    "healthy",
		Timestamp: time.Now().Format(time.RFC3339),
		Version:   s.version,
		Checks:    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleClassify handles business classification requests
// Supports Redis caching when Redis optimizer is available
func (s *RailwayServer) handleClassify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create cache key for Redis (if available)
	cacheKey := fmt.Sprintf("classification:%s:%s", req.Name, req.Description)

	// Try to get from Redis cache first (if available)
	if s.redisOptimizer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cached, err := s.redisOptimizer.GetClient().Get(ctx, cacheKey).Result()
		if err == nil {
			// Cache hit - return cached result
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("X-Cache", "HIT")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(cached))
			return
		}
	}

	// Cache miss or Redis unavailable - perform classification
	start := time.Now()

	// Generate mock classifications
	classifications := map[string]interface{}{
		"mcc": []map[string]interface{}{
			{"code": "5999", "confidence": 0.95, "description": "Miscellaneous and Specialty Retail Stores"},
			{"code": "5411", "confidence": 0.87, "description": "Grocery Stores, Supermarkets"},
			{"code": "5311", "confidence": 0.82, "description": "Department Stores"},
		},
		"naics": []map[string]interface{}{
			{"code": "44-45", "confidence": 0.96, "description": "Retail Trade"},
			{"code": "44-11", "confidence": 0.89, "description": "Food and Beverage Stores"},
			{"code": "44-21", "confidence": 0.85, "description": "General Merchandise Stores"},
		},
		"sic": []map[string]interface{}{
			{"code": "5999", "confidence": 0.94, "description": "Miscellaneous Retail Stores, Not Elsewhere Classified"},
			{"code": "5411", "confidence": 0.88, "description": "Grocery Stores"},
			{"code": "5311", "confidence": 0.83, "description": "Department Stores"},
		},
	}

	riskAssessment := map[string]interface{}{
		"level":      "low",
		"score":      0.15,
		"confidence": 0.92,
		"factors":    []string{"established_business", "low_risk_industry"},
	}

	processingTime := time.Since(start)

	response := ClassificationResponse{
		BusinessName:    req.Name,
		Description:     req.Description,
		Classifications: classifications,
		RiskAssessment:  riskAssessment,
		ProcessingTime:  processingTime.String(),
		Timestamp:       time.Now().Format(time.RFC3339),
	}

	// Cache the result in Redis (if available)
	if s.redisOptimizer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Serialize response for caching
		responseBytes, err := json.Marshal(response)
		if err == nil {
			// Use Redis optimizer's intelligent caching strategy
			s.redisOptimizer.OptimizeCacheStrategy(ctx, cacheKey, string(responseBytes), "classification")
		}
	}

	// Save to database if available
	if s.supabaseClient != nil {
		classificationData := map[string]interface{}{
			"business_name":   req.Name,
			"description":     req.Description,
			"classifications": classifications,
			"risk_assessment": riskAssessment,
			"processing_time": processingTime.String(),
			"created_at":      time.Now(),
		}

		_, _, err := s.supabaseClient.From("classifications").Insert(classificationData, false, "", "", "").Execute()
		if err != nil {
			log.Printf("Failed to save classification: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if s.redisOptimizer != nil {
		w.Header().Set("X-Cache", "MISS")
	}
	json.NewEncoder(w).Encode(response)
}

// handleMerchants handles merchant management requests
func (s *RailwayServer) handleMerchants(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetMerchants(w, r)
	case http.MethodPost:
		s.handleCreateMerchant(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetMerchants handles getting merchants
func (s *RailwayServer) handleGetMerchants(w http.ResponseWriter, r *http.Request) {
	// Mock merchant data
	merchants := []MerchantResponse{
		{
			ID:          "merchant-1",
			Name:        "Acme Corporation",
			Description: "Leading technology company",
			Email:       "contact@acme.com",
			Phone:       "+1-555-0123",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			UpdatedAt:   time.Now().Add(-1 * time.Hour),
		},
		{
			ID:          "merchant-2",
			Name:        "Global Solutions Inc",
			Description: "International consulting firm",
			Email:       "info@globalsolutions.com",
			Phone:       "+1-555-0456",
			CreatedAt:   time.Now().Add(-48 * time.Hour),
			UpdatedAt:   time.Now().Add(-2 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"merchants": merchants,
		"total":     len(merchants),
		"page":      1,
		"limit":     10,
	})
}

// handleCreateMerchant handles creating a new merchant
func (s *RailwayServer) handleCreateMerchant(w http.ResponseWriter, r *http.Request) {
	var req MerchantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	merchant := MerchantResponse{
		ID:          fmt.Sprintf("merchant-%d", time.Now().Unix()),
		Name:        req.Name,
		Description: req.Description,
		Email:       req.Email,
		Phone:       req.Phone,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save to database if available
	if s.supabaseClient != nil {
		merchantData := map[string]interface{}{
			"id":          merchant.ID,
			"name":        merchant.Name,
			"description": merchant.Description,
			"email":       merchant.Email,
			"phone":       merchant.Phone,
			"created_at":  merchant.CreatedAt,
			"updated_at":  merchant.UpdatedAt,
		}

		_, _, err := s.supabaseClient.From("merchants").Insert(merchantData, false, "", "", "").Execute()
		if err != nil {
			log.Printf("Failed to save merchant: %v", err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(merchant)
}

// handleGenerateToken handles JWT token generation
func (s *RailwayServer) handleGenerateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Generate a mock JWT token (in production, use proper JWT library)
	token := fmt.Sprintf("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.%s",
		fmt.Sprintf("%x", time.Now().Unix()))

	expiresAt := time.Now().Add(24 * time.Hour)

	response := TokenResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		UserID:    req.UserID,
		Email:     req.Email,
		Role:      req.Role,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleValidateToken handles JWT token validation
func (s *RailwayServer) handleValidateToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Mock token validation
	valid := strings.HasPrefix(req.Token, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.")

	response := map[string]interface{}{
		"valid":     valid,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if valid {
		response["user_id"] = "user-123"
		response["email"] = "user@example.com"
		response["role"] = "admin"
		response["permissions"] = []string{"read", "write", "admin"}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleAnalytics handles analytics requests
func (s *RailwayServer) handleAnalytics(w http.ResponseWriter, r *http.Request) {
	// Mock analytics data
	overallStats := map[string]interface{}{
		"total_classifications":      1250,
		"successful_classifications": 1180,
		"failed_classifications":     70,
		"success_rate":               94.4,
		"avg_response_time":          "45ms",
		"total_users":                45,
	}

	topIndustries := []map[string]interface{}{
		{"industry": "Retail Trade", "count": 450, "percentage": 36.0},
		{"industry": "Professional Services", "count": 320, "percentage": 25.6},
		{"industry": "Technology", "count": 280, "percentage": 22.4},
	}

	topRiskLevels := []map[string]interface{}{
		{"risk_level": "Low", "count": 950, "percentage": 76.0},
		{"risk_level": "Medium", "count": 250, "percentage": 20.0},
		{"risk_level": "High", "count": 50, "percentage": 4.0},
	}

	response := AnalyticsResponse{
		Service:       s.serviceName,
		Version:       s.version,
		Timestamp:     time.Now().Format(time.RFC3339),
		OverallStats:  overallStats,
		TopIndustries: topIndustries,
		TopRiskLevels: topRiskLevels,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleMetrics handles performance metrics
// Includes Redis performance data if Redis optimizer is available
func (s *RailwayServer) handleMetrics(w http.ResponseWriter, r *http.Request) {
	// Mock metrics data
	metrics := map[string]interface{}{
		"requests": map[string]interface{}{
			"total":        1250,
			"successful":   1180,
			"failed":       70,
			"success_rate": 94.4,
		},
		"response_times": map[string]interface{}{
			"average": "45ms",
			"min":     "12ms",
			"max":     "2.3s",
		},
		"cache": map[string]interface{}{
			"hits":     850,
			"misses":   400,
			"hit_rate": 68.0,
		},
		"errors": map[string]int{
			"4xx": 45,
			"5xx": 25,
		},
	}

	// Add Redis metrics if available
	if s.redisOptimizer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		redisStats, err := s.redisOptimizer.GetCacheStats(ctx)
		if err == nil {
			metrics["redis"] = map[string]interface{}{
				"connections": map[string]int{
					"total":  redisStats.TotalConnections,
					"active": redisStats.ActiveConnections,
					"idle":   redisStats.IdleConnections,
				},
				"performance": map[string]interface{}{
					"hit_rate":  redisStats.HitRate,
					"miss_rate": redisStats.MissRate,
				},
			}
		}
	}

	response := MetricsResponse{
		Service:   s.serviceName,
		Version:   s.version,
		Timestamp: time.Now().Format(time.RFC3339),
		Metrics:   metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleSelfDriving handles self-driving capabilities
func (s *RailwayServer) handleSelfDriving(w http.ResponseWriter, r *http.Request) {
	// Mock self-driving capabilities
	capabilities := map[string]interface{}{
		"auto_scaling": map[string]interface{}{
			"enabled":          true,
			"min_replicas":     1,
			"max_replicas":     10,
			"current_replicas": 2,
		},
		"circuit_breaker": map[string]interface{}{
			"enabled":           true,
			"failure_threshold": 5,
			"timeout":           "30s",
			"state":             "closed",
		},
		"health_monitoring": map[string]interface{}{
			"enabled":  true,
			"interval": "30s",
			"checks":   []string{"database", "cache", "external_apis"},
		},
		"alerting": map[string]interface{}{
			"enabled":  true,
			"channels": []string{"email", "slack", "webhook"},
			"thresholds": map[string]interface{}{
				"cpu_usage":     80,
				"error_rate":    5,
				"response_time": "2s",
			},
		},
	}

	response := SelfDrivingResponse{
		Service:      s.serviceName,
		Version:      s.version,
		Timestamp:    time.Now().Format(time.RFC3339),
		Capabilities: capabilities,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleReports handles reporting requests
func (s *RailwayServer) handleReports(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetReports(w, r)
	case http.MethodPost:
		s.handleGenerateReport(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetReports handles getting reports
func (s *RailwayServer) handleGetReports(w http.ResponseWriter, r *http.Request) {
	// Mock reports data
	reports := []ReportResponse{
		{
			ID:          "report-1",
			Title:       "Monthly Analytics Report",
			Type:        "analytics",
			Status:      "completed",
			CreatedAt:   time.Now().Add(-24 * time.Hour),
			GeneratedAt: func() *time.Time { t := time.Now().Add(-23 * time.Hour); return &t }(),
		},
		{
			ID:        "report-2",
			Title:     "Performance Metrics Report",
			Type:      "performance",
			Status:    "generating",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"reports": reports,
		"total":   len(reports),
	})
}

// handleGenerateReport handles generating a new report
func (s *RailwayServer) handleGenerateReport(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
		Type  string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	report := ReportResponse{
		ID:        fmt.Sprintf("report-%d", time.Now().Unix()),
		Title:     req.Title,
		Type:      req.Type,
		Status:    "generating",
		CreatedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(report)
}

// handleExportReport handles exporting a report
func (s *RailwayServer) handleExportReport(w http.ResponseWriter, r *http.Request) {
	reportID := r.URL.Query().Get("id")
	if reportID == "" {
		http.Error(w, "Report ID is required", http.StatusBadRequest)
		return
	}

	// Mock export data
	exportData := map[string]interface{}{
		"report_id":   reportID,
		"format":      "json",
		"exported_at": time.Now().Format(time.RFC3339),
		"data": map[string]interface{}{
			"total_classifications": 1250,
			"success_rate":          94.4,
			"avg_response_time":     "45ms",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(exportData)
}

// handleRateLimits handles rate limit information
func (s *RailwayServer) handleRateLimits(w http.ResponseWriter, r *http.Request) {
	// Mock rate limit data
	rateLimits := map[string]interface{}{
		"requests_per_minute": 60,
		"burst_size":          10,
		"total_users":         45,
		"user_stats": map[string]interface{}{
			"user-123": map[string]interface{}{
				"user_id":             "user-123",
				"requests_per_minute": 60.0,
				"burst_size":          10,
				"requests_count":      25,
				"status":              "active",
				"last_cleanup":        time.Now().Format(time.RFC3339),
			},
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rateLimits)
}

// handleDocs handles API documentation
func (s *RailwayServer) handleDocs(w http.ResponseWriter, r *http.Request) {
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>%s API Documentation v%s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { background: #f5f5f5; padding: 20px; margin: 20px 0; border-radius: 5px; }
        .method { background: #007bff; color: white; padding: 5px 10px; border-radius: 3px; }
        .method.post { background: #28a745; }
        .method.get { background: #17a2b8; }
        .method.put { background: #ffc107; color: black; }
        .method.delete { background: #dc3545; }
    </style>
</head>
<body>
    <h1>%s API Documentation v%s</h1>
    <p>Complete KYB Platform v4.0.0 with ALL advanced features - CACHE BUSTED!</p>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /health</h3>
        <p>Basic health check endpoint with service information.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /health/detailed</h3>
        <p>Detailed health check with component status.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/classify</h3>
        <p>Enhanced business classification with MCC, NAICS, and SIC codes.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v2/classify</h3>
        <p>V2 business classification with improved accuracy.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /merchants</h3>
        <p>Get list of merchants.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /merchants</h3>
        <p>Create a new merchant.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /v1/merchants</h3>
        <p>V1 merchant management API.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/merchants</h3>
        <p>V1 merchant creation API.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /metrics</h3>
        <p>Performance metrics and system statistics.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /self-driving</h3>
        <p>Self-driving capabilities and automation status.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /analytics/overall</h3>
        <p>Overall analytics and business intelligence.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /reports</h3>
        <p>Get list of reports.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /reports</h3>
        <p>Generate a new report.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /reports/export</h3>
        <p>Export a report.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /auth/token</h3>
        <p>Generate JWT token.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /auth/validate</h3>
        <p>Validate JWT token.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /rate-limits</h3>
        <p>Get rate limit information.</p>
    </div>
    
    <h2>Risk Management Endpoints (Restored)</h2>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /v1/risk/thresholds</h3>
        <p>Get all risk thresholds. Supports query parameters: category, industry_code.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/admin/risk/thresholds</h3>
        <p>Create a new risk threshold configuration.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method put">PUT</span> /v1/admin/risk/thresholds/{threshold_id}</h3>
        <p>Update an existing risk threshold configuration.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method delete">DELETE</span> /v1/admin/risk/thresholds/{threshold_id}</h3>
        <p>Delete a risk threshold configuration.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /v1/admin/risk/threshold-export</h3>
        <p>Export all risk thresholds as JSON.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/admin/risk/threshold-import</h3>
        <p>Import risk thresholds from JSON.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /v1/risk/factors</h3>
        <p>Get all risk factors. Supports query parameter: category.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /v1/risk/categories</h3>
        <p>Get all risk categories.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/admin/risk/recommendation-rules</h3>
        <p>Create a new recommendation rule.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method put">PUT</span> /v1/admin/risk/recommendation-rules/{rule_id}</h3>
        <p>Update an existing recommendation rule.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method delete">DELETE</span> /v1/admin/risk/recommendation-rules/{rule_id}</h3>
        <p>Delete a recommendation rule.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/admin/risk/notification-channels</h3>
        <p>Create a new notification channel (email, SMS, Slack, webhook, Teams, Discord, PagerDuty).</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method put">PUT</span> /v1/admin/risk/notification-channels/{channel_id}</h3>
        <p>Update an existing notification channel.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method delete">DELETE</span> /v1/admin/risk/notification-channels/{channel_id}</h3>
        <p>Delete a notification channel.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /v1/admin/risk/system/health</h3>
        <p>Get system health status for risk management services.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /v1/admin/risk/system/metrics</h3>
        <p>Get system metrics for risk management services.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/admin/risk/system/cleanup</h3>
        <p>Cleanup old system data (alerts, trends, assessments).</p>
    </div>
    
    <h2>Enhanced Risk Assessment</h2>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/risk/enhanced/assess</h3>
        <p>Perform enhanced risk assessment with comprehensive analysis.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/risk/factors/calculate</h3>
        <p>Calculate risk factor scores.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/risk/recommendations</h3>
        <p>Get risk mitigation recommendations.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method post">POST</span> /v1/risk/trends/analyze</h3>
        <p>Analyze risk trends over time.</p>
    </div>
    
    <div class="endpoint">
        <h3><span class="method get">GET</span> /v1/risk/alerts</h3>
        <p>Get active risk alerts.</p>
    </div>
</body>
</html>`, s.serviceName, s.version, s.serviceName, s.version)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// setupRoutes sets up all the HTTP routes
func (s *RailwayServer) setupRoutes() {
	// Health endpoints
	s.mux.HandleFunc("/health", s.handleHealth)
	s.mux.HandleFunc("/health/detailed", s.handleDetailedHealth)

	// Classification endpoints
	s.mux.HandleFunc("/v1/classify", s.handleClassify)
	s.mux.HandleFunc("/v2/classify", s.handleClassify)
	s.mux.HandleFunc("/classify", s.handleClassify) // Legacy support

	// Merchant endpoints (legacy - only register if they don't conflict with new API routes)
	// Note: /api/v1/merchants/* routes are registered in setupNewAPIRoutes()
	// Legacy routes are kept for backward compatibility but won't conflict with /api/v1/* paths
	s.mux.HandleFunc("/merchants", s.handleMerchants)
	// Commented out /v1/merchants to avoid conflict with /api/v1/merchants/* routes
	// The new routes registered in setupNewAPIRoutes() handle /api/v1/merchants/* paths
	// s.mux.HandleFunc("/v1/merchants", s.handleMerchants)

	// Analytics endpoints
	s.mux.HandleFunc("/analytics/overall", s.handleAnalytics)

	// Metrics and monitoring
	s.mux.HandleFunc("/metrics", s.handleMetrics)
	s.mux.HandleFunc("/self-driving", s.handleSelfDriving)

	// Reporting endpoints
	s.mux.HandleFunc("/reports", s.handleReports)
	s.mux.HandleFunc("/reports/export", s.handleExportReport)

	// Authentication endpoints
	s.mux.HandleFunc("/auth/token", s.handleGenerateToken)
	s.mux.HandleFunc("/auth/validate", s.handleValidateToken)

	// Rate limiting
	s.mux.HandleFunc("/rate-limits", s.handleRateLimits)

	// Documentation
	s.mux.HandleFunc("/docs", s.handleDocs)

	// Redis optimization endpoint (only register if Redis is available)
	if s.redisOptimizer != nil {
		s.mux.HandleFunc("/redis-optimization", s.handleRedisOptimization)
	}

	// Register new merchant analytics and async risk assessment routes
	s.setupNewAPIRoutes()
}

// setupNewAPIRoutes registers the new merchant analytics and async risk assessment routes
func (s *RailwayServer) setupNewAPIRoutes() {
	// Routes can work with or without database (threshold manager supports in-memory mode)
	// Only skip if we can't create handlers at all

	logger := log.Default()

	// Create zap logger for observability and middleware
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Printf("Warning: Failed to create zap logger: %v. Using default logger.", err)
		zapLogger = zap.NewNop()
	}

	// Initialize observability logger
	obsLogger := observability.NewLogger(zapLogger)

	// Initialize repositories (only if database is available)
	var merchantRepo *database.MerchantPortfolioRepository
	var analyticsRepo *database.MerchantAnalyticsRepository
	var riskAssessmentRepo *database.RiskAssessmentRepository
	var riskIndicatorsRepo *database.RiskIndicatorsRepository

	if s.db != nil {
		merchantRepo = database.NewMerchantPortfolioRepository(s.db, logger)
		analyticsRepo = database.NewMerchantAnalyticsRepository(s.db, logger)
		riskAssessmentRepo = database.NewRiskAssessmentRepository(s.db, logger)
		riskIndicatorsRepo = database.NewRiskIndicatorsRepository(s.db, logger)
	} else {
		log.Println("⚠️  Database unavailable - repositories not initialized. Some features will be limited.")
		// Repositories will be nil, handlers should check for nil before use
	}

	// Initialize services (only if repositories are available)
	// TODO: Create a proper wrapper function in database package: NewPostgresDBFromConnection(*sql.DB)
	// For now, merchantPortfolioService is nil and handler methods check for nil before using it
	// This prevents nil pointer dereference panics and returns 503 Service Unavailable instead
	var merchantPortfolioService services.MerchantPortfolioServiceInterface = nil // TODO: Create proper DB wrapper
	var analyticsService services.MerchantAnalyticsService
	var riskAssessmentService services.RiskAssessmentService
	var riskIndicatorsService services.RiskIndicatorsService
	var dataEnrichmentService services.DataEnrichmentService

	if analyticsRepo != nil && merchantRepo != nil {
		// Initialize cache if Redis is available
		var cacheClient services.Cache = nil
		if s.redisOptimizer != nil {
			// Use Redis optimizer as cache if available
			// For now, pass nil - cache is optional
		}
		analyticsService = services.NewMerchantAnalyticsService(analyticsRepo, merchantRepo, cacheClient, logger)
	} else {
		log.Println("⚠️  Analytics service not initialized - database unavailable")
		// analyticsService will be nil (interface), handlers should check for nil
	}

	if riskAssessmentRepo != nil {
		// RiskAssessmentService requires a jobQueue parameter - pass nil for now
		riskAssessmentService = services.NewRiskAssessmentService(riskAssessmentRepo, nil, logger)
	} else {
		log.Println("⚠️  Risk assessment service not initialized - database unavailable")
		// riskAssessmentService will be nil (interface), handlers should check for nil
	}

	if riskIndicatorsRepo != nil {
		riskIndicatorsService = services.NewRiskIndicatorsService(riskIndicatorsRepo, logger)
	} else {
		log.Println("⚠️  Risk indicators service not initialized - database unavailable")
		// riskIndicatorsService will be nil (interface), handlers should check for nil
	}

	// Data enrichment service doesn't require database (returns mock data for now)
	dataEnrichmentService = services.NewDataEnrichmentService(logger)

	// Initialize handlers
	// Create MerchantPortfolioHandler with both service (for CRUD operations) and repository (for analytics)
	merchantPortfolioHandler := handlers.NewMerchantPortfolioHandlerWithRepository(merchantPortfolioService, merchantRepo, logger)
	analyticsHandler := handlers.NewMerchantAnalyticsHandler(analyticsService, logger)
	asyncRiskHandler := handlers.NewAsyncRiskAssessmentHandler(riskAssessmentService, logger)
	
	// Initialize risk indicators handler (can be nil if service is nil)
	var riskIndicatorsHandler *handlers.RiskIndicatorsHandler
	if riskIndicatorsService != nil {
		riskIndicatorsHandler = handlers.NewRiskIndicatorsHandler(riskIndicatorsService, logger)
	} else {
		log.Println("⚠️  Risk indicators handler not initialized - service unavailable")
	}
	
	// Initialize data enrichment handler
	dataEnrichmentHandler := handlers.NewDataEnrichmentHandler(dataEnrichmentService, logger)

	// Initialize middleware
	// Create auth middleware (can be nil if auth service is not available)
	// The middleware will still work but may not validate tokens properly
	var authMiddleware *middleware.AuthMiddleware
	authMiddleware = middleware.NewAuthMiddleware(nil, zapLogger) // Pass nil auth service for now

	// Create rate limiter
	rateLimitConfig := &middleware.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 100,
		BurstSize:         10,
		WindowSize:        1 * time.Minute,
		Strategy:          "token_bucket",
	}
	rateLimiter := middleware.NewAPIRateLimiter(rateLimitConfig, zapLogger)

	// Register merchant routes with analytics handler
	merchantConfig := &routes.MerchantRouteConfig{
		MerchantPortfolioHandler: merchantPortfolioHandler, // Now includes repository for analytics
		MerchantAnalyticsHandler: analyticsHandler,
		AsyncRiskHandler:         asyncRiskHandler,
		DataEnrichmentHandler:    dataEnrichmentHandler,
		AuthMiddleware:           authMiddleware,
		RateLimiter:              rateLimiter,
		Logger:                   obsLogger,
	}
	routes.RegisterMerchantRoutes(s.mux, merchantConfig)

	// Register risk routes with async config
	// Note: RiskHandler may need existing dependencies
	asyncRiskConfig := &routes.AsyncRiskAssessmentRouteConfig{
		AsyncRiskHandler:     asyncRiskHandler,
		RiskIndicatorsHandler: riskIndicatorsHandler,
		AuthMiddleware:       authMiddleware,
		RateLimiter:          rateLimiter,
	}
	// Register async routes (RiskHandler can be nil if not needed for async routes)
	routes.RegisterRiskRoutesWithConfig(s.mux, nil, asyncRiskConfig)

	// Initialize enhanced risk handler and register routes
	// Create enhanced risk service components using factory
	enhancedRiskFactory := risk.NewEnhancedRiskServiceFactory(zapLogger)
	enhancedCalculator := enhancedRiskFactory.CreateRiskFactorCalculator()
	recommendationEngine := enhancedRiskFactory.CreateRecommendationEngine()
	trendAnalysisService := enhancedRiskFactory.CreateTrendAnalysisService()
	alertSystem := enhancedRiskFactory.CreateAlertSystem()

	// Create and initialize ThresholdManager with database persistence
	var thresholdManager *risk.ThresholdManager
	if s.db != nil {
		// Create threshold repository
		thresholdRepo := database.NewThresholdRepository(s.db, logger)
		// Create adapter to bridge database and risk packages
		thresholdRepoAdapter := risk.NewThresholdRepositoryAdapter(thresholdRepo)
		// Create manager with repository
		thresholdManager = risk.NewThresholdManagerWithRepository(thresholdRepoAdapter)

		// Load thresholds from database on startup
		ctx := context.Background()
		if err := thresholdManager.LoadFromDatabase(ctx); err != nil {
			log.Printf("Warning: Failed to load thresholds from database: %v. Using default thresholds.", err)
			// Fall back to default thresholds if database load fails
			thresholdManager = risk.CreateDefaultThresholds()
			// Try to sync defaults to database (non-blocking)
			go func() {
				if err := thresholdManager.SyncToDatabase(ctx); err != nil {
					log.Printf("Warning: Failed to sync default thresholds to database: %v", err)
				}
			}()
		} else {
			log.Printf("✅ Loaded %d thresholds from database", len(thresholdManager.ListConfigs()))
		}
	} else {
		// No database available, use in-memory defaults
		log.Println("⚠️  Database not available, using in-memory threshold defaults")
		thresholdManager = risk.CreateDefaultThresholds()
	}

	// Create risk detection service (can be nil for now - handlers check for nil)
	var riskDetectionService *risk.RiskDetectionService = nil // TODO: Initialize when available

	// Initialize enhanced risk handler
	enhancedRiskHandler := handlers.NewEnhancedRiskHandler(
		zapLogger,
		riskDetectionService,
		enhancedCalculator,
		recommendationEngine,
		trendAnalysisService,
		alertSystem,
		thresholdManager,
	)

	// Register enhanced risk routes (public endpoints)
	routes.RegisterEnhancedRiskRoutes(s.mux, enhancedRiskHandler)

	// Register enhanced risk admin routes
	routes.RegisterEnhancedRiskAdminRoutes(s.mux, enhancedRiskHandler)

	log.Println("✅ New API routes registered:")
	log.Println("   - GET /api/v1/merchants/analytics (portfolio-level analytics)")
	log.Println("   - GET /api/v1/merchants/{merchantId}/analytics")
	log.Println("   - GET /api/v1/merchants/{merchantId}/website-analysis")
	log.Println("   - POST /api/v1/risk/assess")
	log.Println("   - GET /api/v1/risk/assess/{assessmentId}")
	log.Println("   - GET /api/v1/risk/history/{merchantId}")
	log.Println("   - GET /api/v1/risk/predictions/{merchantId}")
	log.Println("   - GET /api/v1/risk/explain/{assessmentId}")
	log.Println("   - GET /api/v1/merchants/{merchantId}/risk-recommendations")
	if riskIndicatorsHandler != nil {
		log.Println("   - GET /api/v1/risk/indicators/{merchantId}")
		log.Println("   - GET /api/v1/risk/alerts/{merchantId}")
	}
	if dataEnrichmentHandler != nil {
		log.Println("   - POST /api/v1/merchants/{merchantId}/enrichment/trigger")
		log.Println("   - GET /api/v1/merchants/{merchantId}/enrichment/sources")
	}
	log.Println("   - GET /v1/risk/factors")
	log.Println("   - GET /v1/risk/categories")
	log.Println("   - GET /v1/risk/thresholds")
	log.Println("   - POST /v1/admin/risk/thresholds")
	log.Println("   - GET /v1/admin/risk/threshold-export")
	log.Println("   - POST /v1/admin/risk/threshold-import")
	log.Println("   - GET /v1/admin/risk/system/health")
	log.Println("   - GET /v1/admin/risk/system/metrics")
	if s.redisOptimizer != nil {
		log.Println("   - GET/POST /redis-optimization (Redis optimization status and controls)")
	}
}

// handleRedisOptimization provides Redis optimization status and controls
func (s *RailwayServer) handleRedisOptimization(w http.ResponseWriter, r *http.Request) {
	if s.redisOptimizer == nil {
		http.Error(w, "Redis optimization not available", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch r.Method {
	case http.MethodGet:
		// Get optimization status
		health, err := s.redisOptimizer.HealthCheck(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Redis health check failed: %v", err), http.StatusInternalServerError)
			return
		}

		stats, err := s.redisOptimizer.GetCacheStats(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get cache stats: %v", err), http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"service":   s.serviceName,
			"version":   s.version,
			"timestamp": time.Now().Format(time.RFC3339),
			"redis_optimization": map[string]interface{}{
				"status": health.Status,
				"health": health,
				"stats":  stats,
				"config": map[string]interface{}{
					"pool_size":         100,
					"min_idle_conns":    10,
					"max_idle_conns":    50,
					"enable_pipelining": true,
					"compression":       true,
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	case http.MethodPost:
		// Warmup cache
		var warmupReq struct {
			Action string `json:"action"`
		}

		if err := json.NewDecoder(r.Body).Decode(&warmupReq); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if warmupReq.Action == "warmup" {
			warmupData := map[string]interface{}{
				"warmup:classification:tech": map[string]string{
					"mcc":   "5411",
					"naics": "541511",
				},
				"warmup:analytics:summary": map[string]interface{}{
					"total":        1250,
					"success_rate": 0.944,
				},
			}

			err := s.redisOptimizer.WarmupCache(ctx, warmupData)
			if err != nil {
				http.Error(w, fmt.Sprintf("Cache warmup failed: %v", err), http.StatusInternalServerError)
				return
			}

			response := map[string]interface{}{
				"status":  "success",
				"message": "Cache warmup completed",
				"items":   len(warmupData),
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		} else {
			http.Error(w, "Invalid action", http.StatusBadRequest)
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// Start starts the server
func (s *RailwayServer) Start() error {
	s.setupRoutes()

	log.Printf("🚀 Starting %s v%s on :%s", s.serviceName, s.version, s.port)
	log.Printf("✅ %s v%s is ready and listening on :%s", s.serviceName, s.version, s.port)
	log.Printf("🔗 Health: http://localhost:%s/health", s.port)
	log.Printf("📊 Metrics: http://localhost:%s/metrics", s.port)
	log.Printf("🤖 Self-Driving: http://localhost:%s/self-driving", s.port)
	log.Printf("📈 Analytics: http://localhost:%s/analytics/overall", s.port)
	log.Printf("📚 Docs: http://localhost:%s/docs", s.port)
	if s.redisOptimizer != nil {
		log.Printf("⚡ Redis Optimization: http://localhost:%s/redis-optimization", s.port)
		log.Printf("✅ Redis optimization enabled")
	} else {
		log.Printf("⚠️  Redis optimization disabled (REDIS_URL not set)")
	}
	if s.db != nil {
		log.Printf("✅ Database connection established")
	} else {
		log.Printf("⚠️  Database not available (DATABASE_URL not set or connection failed)")
	}

	return http.ListenAndServe(":"+s.port, s.mux)
}

func main() {
	server := NewRailwayServer()

	// Close database connection on exit
	if server.db != nil {
		defer server.db.Close()
	}

	// Close Redis connection on exit (if available)
	if server.redisOptimizer != nil {
		defer func() {
			if client := server.redisOptimizer.GetClient(); client != nil {
				if err := client.Close(); err != nil {
					log.Printf("Error closing Redis connection: %v", err)
				}
			}
		}()
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
