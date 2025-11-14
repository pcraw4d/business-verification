package routes

import (
	"net/http"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/middleware"

	"go.uber.org/zap"
)

// RegisterRiskRoutes registers risk assessment API routes
// This is a convenience wrapper that calls RegisterRiskRoutesWithConfig with nil asyncConfig
// for backward compatibility
func RegisterRiskRoutes(mux *http.ServeMux, riskHandler *handlers.RiskHandler) {
	RegisterRiskRoutesWithConfig(mux, riskHandler, nil)
}

// RegisterRiskRoutesWithConfig registers risk assessment API routes
// If asyncConfig is provided, it will also register async risk assessment routes
func RegisterRiskRoutesWithConfig(mux *http.ServeMux, riskHandler *handlers.RiskHandler, asyncConfig *AsyncRiskAssessmentRouteConfig) {
	// Guard against nil riskHandler - only register legacy routes if handler is provided
	if riskHandler != nil {
		// Create middleware instances
		logger := zap.NewNop()
		corsMiddleware := middleware.NewCORSMiddleware(nil, logger)
		loggingMiddleware := middleware.NewRequestLoggingMiddleware(nil, logger)

		// Risk assessment endpoints
		mux.Handle("POST /v1/risk/assess",
			corsMiddleware.Middleware(
				loggingMiddleware.Middleware(
					http.HandlerFunc(riskHandler.AssessRiskHandler))))

		// Risk history endpoints
		mux.Handle("GET /v1/risk/history/{business_id}",
			corsMiddleware.Middleware(
				loggingMiddleware.Middleware(
					http.HandlerFunc(riskHandler.GetRiskHistoryHandler))))

		// Risk benchmarks endpoint (NEW)
		mux.Handle("GET /v1/risk/benchmarks",
			corsMiddleware.Middleware(
				loggingMiddleware.Middleware(
					http.HandlerFunc(riskHandler.GetRiskBenchmarksHandler))))

		// Risk predictions endpoint (NEW)
		mux.Handle("GET /v1/risk/predictions/{merchant_id}",
			corsMiddleware.Middleware(
				loggingMiddleware.Middleware(
					http.HandlerFunc(riskHandler.GetRiskPredictionsHandler))))

		// Risk categories and factors
		// NOTE: These routes are now registered in RegisterEnhancedRiskRoutes
		// to avoid conflict with enhanced risk handler. Commented out to prevent duplicate registration.
		// mux.Handle("GET /v1/risk/categories",
		// 	corsMiddleware.Middleware(
		// 		loggingMiddleware.Middleware(
		// 			http.HandlerFunc(riskHandler.GetRiskCategoriesHandler))))

		// mux.Handle("GET /v1/risk/factors",
		// 	corsMiddleware.Middleware(
		// 		loggingMiddleware.Middleware(
		// 			http.HandlerFunc(riskHandler.GetRiskFactorsHandler))))

		// Risk thresholds
		// NOTE: This route is now registered in RegisterEnhancedRiskRoutes
		// to avoid conflict with enhanced risk handler. Commented out to prevent duplicate registration.
		// mux.Handle("GET /v1/risk/thresholds",
		// 	corsMiddleware.Middleware(
		// 		loggingMiddleware.Middleware(
		// 			http.HandlerFunc(riskHandler.GetRiskThresholdsHandler))))

		// Industry benchmarks (legacy endpoint - kept for backward compatibility)
		mux.Handle("GET /v1/risk/industry-benchmarks/{industry}",
			corsMiddleware.Middleware(
				loggingMiddleware.Middleware(
					http.HandlerFunc(riskHandler.GetIndustryBenchmarksHandler))))
	}

	// Register async risk assessment routes if config is provided
	if asyncConfig != nil {
		RegisterAsyncRiskAssessmentRoutes(mux, asyncConfig)
	}
}

// AsyncRiskAssessmentRouteConfig holds configuration for async risk assessment routes
type AsyncRiskAssessmentRouteConfig struct {
	AsyncRiskHandler *handlers.AsyncRiskAssessmentHandler
	AuthMiddleware   *middleware.AuthMiddleware
	RateLimiter      *middleware.APIRateLimiter
}

// RegisterAsyncRiskAssessmentRoutes registers async risk assessment routes
// This function must be called separately to register the async risk assessment endpoints
func RegisterAsyncRiskAssessmentRoutes(mux *http.ServeMux, config *AsyncRiskAssessmentRouteConfig) {
	if config == nil || config.AsyncRiskHandler == nil {
		return
	}

	// POST /api/v1/risk/assess - Start async assessment
	mux.Handle("POST /api/v1/risk/assess",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.AsyncRiskHandler.AssessRisk),
			),
		),
	)

	// GET /api/v1/risk/assess/{assessmentId} - Get assessment status
	mux.Handle("GET /api/v1/risk/assess/{assessmentId}",
		config.AuthMiddleware.RequireAuth(
			config.RateLimiter.Middleware(
				http.HandlerFunc(config.AsyncRiskHandler.GetAssessmentStatus),
			),
		),
	)
}

// RegisterAsyncRiskAssessmentRoutesWithConfig is an alias for RegisterAsyncRiskAssessmentRoutes
// for backward compatibility
func RegisterAsyncRiskAssessmentRoutesWithConfig(mux *http.ServeMux, config *AsyncRiskAssessmentRouteConfig) {
	RegisterAsyncRiskAssessmentRoutes(mux, config)
}
