package routes

import (
	"net/http"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/middleware"

	"go.uber.org/zap"
)

// RegisterEnhancedRiskRoutes registers enhanced risk assessment routes
// TODO: These routes are currently registered without middleware chaining
// The middleware functions (RequestIDMiddleware, LoggingMiddleware, CORSMiddleware) don't exist as standalone functions.
// They should be refactored to use middleware structs with Middleware() methods, similar to merchant_routes.go
func RegisterEnhancedRiskRoutes(mux *http.ServeMux, enhancedRiskHandler *handlers.EnhancedRiskHandler) {
	// Create middleware instances if needed
	// For now, register routes directly without middleware to avoid compilation errors
	logger := zap.NewNop()
	corsMiddleware := middleware.NewCORSMiddleware(nil, logger)
	loggingMiddleware := middleware.NewRequestLoggingMiddleware(nil, logger)

	// Enhanced risk assessment endpoints
	mux.Handle("POST /v1/risk/enhanced/assess",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.EnhancedRiskAssessmentHandler))))

	// Risk factor calculation endpoints
	mux.Handle("POST /v1/risk/factors/calculate",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.RiskFactorCalculationHandler))))

	// Risk recommendations endpoints
	mux.Handle("POST /v1/risk/recommendations",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.RiskRecommendationsHandler))))

	// Risk trend analysis endpoints
	mux.Handle("POST /v1/risk/trends/analyze",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.RiskTrendAnalysisHandler))))

	// Risk alerts endpoints
	mux.Handle("GET /v1/risk/alerts",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.RiskAlertsHandler))))

	// Alert management endpoints
	mux.Handle("POST /v1/risk/alerts/{alert_id}/acknowledge",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.AcknowledgeAlertHandler))))

	mux.Handle("POST /v1/risk/alerts/{alert_id}/resolve",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.ResolveAlertHandler))))

	// Risk factor history endpoints
	mux.Handle("GET /v1/risk/factors/{factor_id}/history",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.RiskFactorHistoryHandler))))

	// Additional enhanced endpoints
	mux.Handle("GET /v1/risk/factors",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.GetRiskFactorsHandler))))

	mux.Handle("GET /v1/risk/categories",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.GetRiskCategoriesHandler))))

	mux.Handle("GET /v1/risk/thresholds",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.GetRiskThresholdsHandler))))
}

// RegisterEnhancedRiskAdminRoutes registers admin-specific enhanced risk routes
func RegisterEnhancedRiskAdminRoutes(mux *http.ServeMux, enhancedRiskHandler *handlers.EnhancedRiskHandler) {
	// Create middleware instances
	logger := zap.NewNop()
	corsMiddleware := middleware.NewCORSMiddleware(nil, logger)
	loggingMiddleware := middleware.NewRequestLoggingMiddleware(nil, logger)

	// Admin endpoints for managing risk configuration
	// TODO: Add AdminAuthMiddleware when available
	
	// Threshold export/import endpoints
	// Using different base paths (/threshold-export, /threshold-import) to avoid wildcard conflicts
	// This completely avoids the /thresholds/{id} wildcard pattern matching issue
	mux.Handle("GET /v1/admin/risk/threshold-export",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.ExportThresholdsHandler))))

	mux.Handle("POST /v1/admin/risk/threshold-import",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.ImportThresholdsHandler))))

	// Threshold CRUD endpoints (wildcard patterns - register after specific paths)
	mux.Handle("POST /v1/admin/risk/thresholds",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.CreateRiskThresholdHandler))))

	mux.Handle("PUT /v1/admin/risk/thresholds/{threshold_id}",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.UpdateRiskThresholdHandler))))

	mux.Handle("DELETE /v1/admin/risk/thresholds/{threshold_id}",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.DeleteRiskThresholdHandler))))

	mux.Handle("POST /v1/admin/risk/recommendation-rules",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.CreateRecommendationRuleHandler))))

	mux.Handle("PUT /v1/admin/risk/recommendation-rules/{rule_id}",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.UpdateRecommendationRuleHandler))))

	mux.Handle("DELETE /v1/admin/risk/recommendation-rules/{rule_id}",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.DeleteRecommendationRuleHandler))))

	mux.Handle("POST /v1/admin/risk/notification-channels",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.CreateNotificationChannelHandler))))

	mux.Handle("PUT /v1/admin/risk/notification-channels/{channel_id}",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.UpdateNotificationChannelHandler))))

	mux.Handle("DELETE /v1/admin/risk/notification-channels/{channel_id}",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.DeleteNotificationChannelHandler))))

	// System health and monitoring endpoints
	mux.Handle("GET /v1/admin/risk/system/health",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.GetSystemHealthHandler))))

	mux.Handle("GET /v1/admin/risk/system/metrics",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.GetSystemMetricsHandler))))

	mux.Handle("POST /v1/admin/risk/system/cleanup",
		corsMiddleware.Middleware(
			loggingMiddleware.Middleware(
				http.HandlerFunc(enhancedRiskHandler.CleanupSystemDataHandler))))
}
