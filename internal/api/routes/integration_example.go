package routes

// This file provides an example of how to integrate the merchant analytics
// and async risk assessment routes into your main server application.
//
// Copy the relevant code from the setupRoutesExample function into your
// main server file where routes are registered.

/*
Example integration code:

package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/middleware"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/database"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/services"
)

func setupRoutesWithNewEndpoints() *http.ServeMux {
	mux := http.NewServeMux()
	logger := log.Default()

	// 1. Initialize database connection
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 2. Initialize repositories
	merchantRepo := database.NewMerchantPortfolioRepository(db, logger)
	analyticsRepo := database.NewMerchantAnalyticsRepository(db, logger)
	riskAssessmentRepo := database.NewRiskAssessmentRepository(db, logger)

	// 3. Initialize services
	analyticsService := services.NewMerchantAnalyticsService(analyticsRepo, merchantRepo, logger)
	riskAssessmentService := services.NewRiskAssessmentService(riskAssessmentRepo, logger)

	// 4. Initialize handlers
	merchantHandler := handlers.NewMerchantPortfolioHandler(/* your existing params */)
	analyticsHandler := handlers.NewMerchantAnalyticsHandler(analyticsService, logger)
	riskHandler := handlers.NewRiskHandler(/* your existing params */)
	asyncRiskHandler := handlers.NewAsyncRiskAssessmentHandler(riskAssessmentService, logger)

	// 5. Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(/* your auth service */, logger)
	rateLimiter := middleware.NewAPIRateLimiter(100, 1*time.Minute, logger)
	obsLogger := observability.NewLogger(/* your config */)

	// 6. Register merchant routes (includes analytics)
	merchantConfig := &routes.MerchantRouteConfig{
		MerchantPortfolioHandler: merchantHandler,
		MerchantAnalyticsHandler:  analyticsHandler, // Add this
		AuthMiddleware:            authMiddleware,
		RateLimiter:              rateLimiter,
		Logger:                   obsLogger,
	}
	routes.RegisterMerchantRoutes(mux, merchantConfig)

	// 7. Register risk routes (includes async routes)
	asyncRiskConfig := &routes.AsyncRiskAssessmentRouteConfig{
		AsyncRiskHandler: asyncRiskHandler,
		AuthMiddleware:   authMiddleware,
		RateLimiter:      rateLimiter,
	}
	routes.RegisterRiskRoutesWithConfig(mux, riskHandler, asyncRiskConfig)

	return mux
}

*/

