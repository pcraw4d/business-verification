package routes

import (
	"net/http"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/api/middleware"
)

// RegisterEnhancedRiskRoutes registers enhanced risk assessment routes
func RegisterEnhancedRiskRoutes(mux *http.ServeMux, enhancedRiskHandler *handlers.EnhancedRiskHandler) {
	// Enhanced risk assessment endpoints
	mux.HandleFunc("POST /v1/risk/enhanced/assess",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.EnhancedRiskAssessmentHandler))))

	// Risk factor calculation endpoints
	mux.HandleFunc("POST /v1/risk/factors/calculate",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.RiskFactorCalculationHandler))))

	// Risk recommendations endpoints
	mux.HandleFunc("POST /v1/risk/recommendations",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.RiskRecommendationsHandler))))

	// Risk trend analysis endpoints
	mux.HandleFunc("POST /v1/risk/trends/analyze",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.RiskTrendAnalysisHandler))))

	// Risk alerts endpoints
	mux.HandleFunc("GET /v1/risk/alerts",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.RiskAlertsHandler))))

	// Alert management endpoints
	mux.HandleFunc("POST /v1/risk/alerts/{alert_id}/acknowledge",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.AcknowledgeAlertHandler))))

	mux.HandleFunc("POST /v1/risk/alerts/{alert_id}/resolve",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.ResolveAlertHandler))))

	// Risk factor history endpoints
	mux.HandleFunc("GET /v1/risk/factors/{factor_id}/history",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.RiskFactorHistoryHandler))))

	// Additional enhanced endpoints
	mux.HandleFunc("GET /v1/risk/factors",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.GetRiskFactorsHandler))))

	mux.HandleFunc("GET /v1/risk/categories",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.GetRiskCategoriesHandler))))

	mux.HandleFunc("GET /v1/risk/thresholds",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					enhancedRiskHandler.GetRiskThresholdsHandler))))
}

// RegisterEnhancedRiskAdminRoutes registers admin-specific enhanced risk routes
func RegisterEnhancedRiskAdminRoutes(mux *http.ServeMux, enhancedRiskHandler *handlers.EnhancedRiskHandler) {
	// Admin endpoints for managing risk configuration
	mux.HandleFunc("POST /v1/admin/risk/thresholds",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.CreateRiskThresholdHandler)))))

	mux.HandleFunc("PUT /v1/admin/risk/thresholds/{threshold_id}",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.UpdateRiskThresholdHandler)))))

	mux.HandleFunc("DELETE /v1/admin/risk/thresholds/{threshold_id}",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.DeleteRiskThresholdHandler)))))

	mux.HandleFunc("POST /v1/admin/risk/recommendation-rules",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.CreateRecommendationRuleHandler)))))

	mux.HandleFunc("PUT /v1/admin/risk/recommendation-rules/{rule_id}",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.UpdateRecommendationRuleHandler)))))

	mux.HandleFunc("DELETE /v1/admin/risk/recommendation-rules/{rule_id}",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.DeleteRecommendationRuleHandler)))))

	mux.HandleFunc("POST /v1/admin/risk/notification-channels",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.CreateNotificationChannelHandler)))))

	mux.HandleFunc("PUT /v1/admin/risk/notification-channels/{channel_id}",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.UpdateNotificationChannelHandler)))))

	mux.HandleFunc("DELETE /v1/admin/risk/notification-channels/{channel_id}",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.DeleteNotificationChannelHandler)))))

	// System health and monitoring endpoints
	mux.HandleFunc("GET /v1/admin/risk/system/health",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.GetSystemHealthHandler)))))

	mux.HandleFunc("GET /v1/admin/risk/system/metrics",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.GetSystemMetricsHandler)))))

	mux.HandleFunc("POST /v1/admin/risk/system/cleanup",
		middleware.RequestIDMiddleware(
			middleware.LoggingMiddleware(
				middleware.CORSMiddleware(
					middleware.AdminAuthMiddleware(
						enhancedRiskHandler.CleanupSystemDataHandler)))))
}
