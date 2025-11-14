package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/database"
	"kyb-platform/internal/risk"
)

// Minimal server to test route registration
func main() {
	// Connect to database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://kyb_test:kyb_test_password@localhost:5433/kyb_test?sslmode=disable"
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create mux
	mux := http.NewServeMux()

	// Create logger
	logger := zap.NewNop()
	zapLogger, _ := zap.NewProduction()

	// Initialize enhanced risk handler
	enhancedRiskFactory := risk.NewEnhancedRiskServiceFactory(zapLogger)
	enhancedCalculator := enhancedRiskFactory.CreateRiskFactorCalculator()
	recommendationEngine := enhancedRiskFactory.CreateRecommendationEngine()
	trendAnalysisService := enhancedRiskFactory.CreateTrendAnalysisService()
	alertSystem := enhancedRiskFactory.CreateAlertSystem()

	// Create threshold manager with database
	thresholdRepo := database.NewThresholdRepository(db, log.Default())
	thresholdRepoAdapter := risk.NewThresholdRepositoryAdapter(thresholdRepo)
	thresholdManager := risk.NewThresholdManagerWithRepository(thresholdRepoAdapter)

	enhancedRiskHandler := handlers.NewEnhancedRiskHandler(
		zapLogger,
		nil,
		enhancedCalculator,
		recommendationEngine,
		trendAnalysisService,
		alertSystem,
		thresholdManager,
	)

	// Register routes
	fmt.Println("Registering enhanced risk routes...")
	routes.RegisterEnhancedRiskRoutes(mux, enhancedRiskHandler)
	fmt.Println("Registering enhanced risk admin routes...")
	routes.RegisterEnhancedRiskAdminRoutes(mux, enhancedRiskHandler)

	// Test route registration by checking if handler exists
	req, _ := http.NewRequest("GET", "/v1/risk/thresholds", nil)
	_, pattern := mux.Handler(req)
	fmt.Printf("Pattern matched for GET /v1/risk/thresholds: %s\n", pattern)

	// Start server
	fmt.Println("Starting server on :8081...")
	fmt.Println("Test with: curl http://localhost:8081/v1/risk/thresholds")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal(err)
	}
}

