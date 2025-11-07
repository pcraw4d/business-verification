package routes

import (
	"net/http"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/middleware"
)

// RegisterRiskRoutes registers risk assessment API routes
func RegisterRiskRoutes(mux *http.ServeMux, riskHandler *handlers.RiskHandler) {
	// Risk assessment endpoints
	mux.HandleFunc("POST /v1/risk/assess",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					riskHandler.AssessRiskHandler))))

	// Risk history endpoints
	mux.HandleFunc("GET /v1/risk/history/{business_id}",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					riskHandler.GetRiskHistoryHandler))))

	// Risk benchmarks endpoint (NEW)
	mux.HandleFunc("GET /v1/risk/benchmarks",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					riskHandler.GetRiskBenchmarksHandler))))

	// Risk predictions endpoint (NEW)
	mux.HandleFunc("GET /v1/risk/predictions/{merchant_id}",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					riskHandler.GetRiskPredictionsHandler))))

	// Risk categories and factors
	mux.HandleFunc("GET /v1/risk/categories",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					riskHandler.GetRiskCategoriesHandler))))

	mux.HandleFunc("GET /v1/risk/factors",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					riskHandler.GetRiskFactorsHandler))))

	// Risk thresholds
	mux.HandleFunc("GET /v1/risk/thresholds",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					riskHandler.GetRiskThresholdsHandler))))

	// Industry benchmarks (legacy endpoint - kept for backward compatibility)
	mux.HandleFunc("GET /v1/risk/industry-benchmarks/{industry}",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					riskHandler.GetIndustryBenchmarksHandler))))
}

