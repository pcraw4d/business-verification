package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/pcraw4d/business-verification/internal/architecture"
	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/modules/database_classification"
	"go.uber.org/zap"
)

// EnhancedClassificationServer represents the enhanced classification server with database-driven classification
type EnhancedClassificationServer struct {
	server                *http.Server
	classificationService *classification.IntegrationService
	databaseModule        *database_classification.DatabaseClassificationModule
	logger                *log.Logger
	zapLogger             *zap.Logger
}

// NewEnhancedClassificationServer creates a new enhanced classification server
func NewEnhancedClassificationServer(port string) *EnhancedClassificationServer {
	logger := log.New(os.Stdout, "[enhanced-classification] ", log.LstdFlags)
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
	server := &EnhancedClassificationServer{
		classificationService: classificationService,
		databaseModule:        databaseModule,
		logger:                logger,
		zapLogger:             zapLogger,
	}

	// Set up HTTP server
	mux := http.NewServeMux()
	server.setupRoutes(mux)

	server.server = &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	return server
}

// setupRoutes sets up the HTTP routes
func (s *EnhancedClassificationServer) setupRoutes(mux *http.ServeMux) {
	// Enhanced classification endpoint with database-driven keyword classification
	mux.HandleFunc("POST /v1/classify", s.handleClassify)

	// Legacy endpoint for backward compatibility
	mux.HandleFunc("POST /classify", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		// Redirect to the main endpoint
		r.URL.Path = "/v1/classify"
		mux.ServeHTTP(w, r)
	})

	// CORS preflight handlers
	mux.HandleFunc("OPTIONS /v1/classify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("OPTIONS /classify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.WriteHeader(http.StatusOK)
	})

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		healthStatus := map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "3.1.0",
			"features": map[string]interface{}{
				"database_driven_classification": true,
				"supabase_integration":           true,
				"enhanced_keyword_matching":      true,
				"confidence_scoring":             true,
				"industry_detection":             true,
			},
		}
		json.NewEncoder(w).Encode(healthStatus)
	})

	// Serve static files from web directory
	fs := http.FileServer(http.Dir("web/"))
	mux.Handle("/", fs)
}

// handleClassify handles the /v1/classify endpoint with database-driven classification
func (s *EnhancedClassificationServer) handleClassify(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	w.Header().Set("Content-Type", "application/json")

	// Parse request
	var request struct {
		BusinessName     string `json:"business_name"`
		GeographicRegion string `json:"geographic_region"`
		WebsiteURL       string `json:"website_url"`
		Description      string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Process with database-driven classification
	ctx := context.Background()

	// Create module request for the database module
	moduleReq := architecture.ModuleRequest{
		ID: fmt.Sprintf("req_%d", time.Now().UnixNano()),
		Data: map[string]interface{}{
			"business_name": request.BusinessName,
			"description":   request.Description,
			"website_url":   request.WebsiteURL,
		},
	}

	// Process classification using the database module
	response, err := s.databaseModule.Process(ctx, moduleReq)
	if err != nil {
		s.logger.Printf("❌ Classification failed: %v", err)
		http.Error(w, "Classification failed", http.StatusInternalServerError)
		return
	}

	// Return the database module response directly
	json.NewEncoder(w).Encode(response)
}

// Start starts the server
func (s *EnhancedClassificationServer) Start() error {
	s.logger.Printf("🚀 Starting Enhanced Classification Server on port %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	server := NewEnhancedClassificationServer(port)

	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
