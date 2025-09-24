package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/observability"
)

// AdvancedMonitoringRoutes sets up routes for advanced monitoring dashboard
func AdvancedMonitoringRoutes(
	router *mux.Router,
	dashboard *observability.AdvancedMonitoringDashboard,
	logger *zap.Logger,
) {
	// Create handler
	handler := handlers.NewAdvancedMonitoringHandler(dashboard, logger)

	// API version prefix
	apiV1 := router.PathPrefix("/api/v1").Subrouter()

	// Advanced monitoring dashboard routes
	monitoring := apiV1.PathPrefix("/monitoring").Subrouter()

	// Main dashboard data endpoint
	monitoring.HandleFunc("/dashboard", handler.GetDashboardData).Methods("GET")
	monitoring.HandleFunc("/dashboard/data", handler.GetDashboardData).Methods("GET")

	// Individual metrics endpoints
	monitoring.HandleFunc("/ml-models", handler.GetMLModelMetrics).Methods("GET")
	monitoring.HandleFunc("/ensemble", handler.GetEnsembleMetrics).Methods("GET")
	monitoring.HandleFunc("/uncertainty", handler.GetUncertaintyMetrics).Methods("GET")
	monitoring.HandleFunc("/security", handler.GetSecurityMetrics).Methods("GET")
	monitoring.HandleFunc("/performance", handler.GetPerformanceMetrics).Methods("GET")

	// Alerts and health endpoints
	monitoring.HandleFunc("/alerts", handler.GetAlertsSummary).Methods("GET")
	monitoring.HandleFunc("/health", handler.GetHealthStatus).Methods("GET")
	monitoring.HandleFunc("/recommendations", handler.GetRecommendations).Methods("GET")

	// Visualization endpoints
	monitoring.HandleFunc("/visualizations/model-drift", handler.GetModelDriftVisualization).Methods("GET")
	monitoring.HandleFunc("/visualizations/ensemble-contribution", handler.GetEnsembleContributionVisualization).Methods("GET")
	monitoring.HandleFunc("/visualizations/uncertainty", handler.GetUncertaintyVisualization).Methods("GET")
	monitoring.HandleFunc("/visualizations/security-compliance", handler.GetSecurityComplianceVisualization).Methods("GET")

	// Export endpoints
	monitoring.HandleFunc("/export/{format}", handler.ExportDashboardData).Methods("GET")
	monitoring.HandleFunc("/export", handler.ExportDashboardData).Methods("GET").Queries("format", "json")

	// Real-time endpoints (for WebSocket or Server-Sent Events in the future)
	monitoring.HandleFunc("/realtime/dashboard", handler.GetDashboardData).Methods("GET")
	monitoring.HandleFunc("/realtime/health", handler.GetHealthStatus).Methods("GET")

	// Add CORS middleware for monitoring endpoints
	monitoring.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Add logging middleware for monitoring endpoints
	monitoring.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("monitoring request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)
			next.ServeHTTP(w, r)
		})
	})

	// Add rate limiting middleware for monitoring endpoints
	monitoring.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Simple rate limiting - in production, use a proper rate limiter
			// This is a placeholder for rate limiting implementation
			next.ServeHTTP(w, r)
		})
	})

	logger.Info("advanced monitoring routes configured successfully")
}

// AdvancedMonitoringRoutesV2 sets up routes for advanced monitoring dashboard v2
func AdvancedMonitoringRoutesV2(
	router *mux.Router,
	dashboard *observability.AdvancedMonitoringDashboard,
	logger *zap.Logger,
) {
	// Create handler
	handler := handlers.NewAdvancedMonitoringHandler(dashboard, logger)

	// API version prefix
	apiV2 := router.PathPrefix("/api/v2").Subrouter()

	// Advanced monitoring dashboard routes
	monitoring := apiV2.PathPrefix("/monitoring").Subrouter()

	// Main dashboard data endpoint
	monitoring.HandleFunc("/dashboard", handler.GetDashboardData).Methods("GET")
	monitoring.HandleFunc("/dashboard/data", handler.GetDashboardData).Methods("GET")

	// Individual metrics endpoints
	monitoring.HandleFunc("/ml-models", handler.GetMLModelMetrics).Methods("GET")
	monitoring.HandleFunc("/ensemble", handler.GetEnsembleMetrics).Methods("GET")
	monitoring.HandleFunc("/uncertainty", handler.GetUncertaintyMetrics).Methods("GET")
	monitoring.HandleFunc("/security", handler.GetSecurityMetrics).Methods("GET")
	monitoring.HandleFunc("/performance", handler.GetPerformanceMetrics).Methods("GET")

	// Alerts and health endpoints
	monitoring.HandleFunc("/alerts", handler.GetAlertsSummary).Methods("GET")
	monitoring.HandleFunc("/health", handler.GetHealthStatus).Methods("GET")
	monitoring.HandleFunc("/recommendations", handler.GetRecommendations).Methods("GET")

	// Visualization endpoints
	monitoring.HandleFunc("/visualizations/model-drift", handler.GetModelDriftVisualization).Methods("GET")
	monitoring.HandleFunc("/visualizations/ensemble-contribution", handler.GetEnsembleContributionVisualization).Methods("GET")
	monitoring.HandleFunc("/visualizations/uncertainty", handler.GetUncertaintyVisualization).Methods("GET")
	monitoring.HandleFunc("/visualizations/security-compliance", handler.GetSecurityComplianceVisualization).Methods("GET")

	// Export endpoints
	monitoring.HandleFunc("/export/{format}", handler.ExportDashboardData).Methods("GET")
	monitoring.HandleFunc("/export", handler.ExportDashboardData).Methods("GET").Queries("format", "json")

	// Real-time endpoints
	monitoring.HandleFunc("/realtime/dashboard", handler.GetDashboardData).Methods("GET")
	monitoring.HandleFunc("/realtime/health", handler.GetHealthStatus).Methods("GET")

	// Add CORS middleware for monitoring endpoints
	monitoring.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Add logging middleware for monitoring endpoints
	monitoring.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("monitoring v2 request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)
			next.ServeHTTP(w, r)
		})
	})

	logger.Info("advanced monitoring v2 routes configured successfully")
}

// AdvancedMonitoringRoutesV3 sets up routes for advanced monitoring dashboard v3
func AdvancedMonitoringRoutesV3(
	router *mux.Router,
	dashboard *observability.AdvancedMonitoringDashboard,
	logger *zap.Logger,
) {
	// Create handler
	handler := handlers.NewAdvancedMonitoringHandler(dashboard, logger)

	// API version prefix
	apiV3 := router.PathPrefix("/api/v3").Subrouter()

	// Advanced monitoring dashboard routes
	monitoring := apiV3.PathPrefix("/monitoring").Subrouter()

	// Main dashboard data endpoint
	monitoring.HandleFunc("/dashboard", handler.GetDashboardData).Methods("GET")
	monitoring.HandleFunc("/dashboard/data", handler.GetDashboardData).Methods("GET")

	// Individual metrics endpoints
	monitoring.HandleFunc("/ml-models", handler.GetMLModelMetrics).Methods("GET")
	monitoring.HandleFunc("/ensemble", handler.GetEnsembleMetrics).Methods("GET")
	monitoring.HandleFunc("/uncertainty", handler.GetUncertaintyMetrics).Methods("GET")
	monitoring.HandleFunc("/security", handler.GetSecurityMetrics).Methods("GET")
	monitoring.HandleFunc("/performance", handler.GetPerformanceMetrics).Methods("GET")

	// Alerts and health endpoints
	monitoring.HandleFunc("/alerts", handler.GetAlertsSummary).Methods("GET")
	monitoring.HandleFunc("/health", handler.GetHealthStatus).Methods("GET")
	monitoring.HandleFunc("/recommendations", handler.GetRecommendations).Methods("GET")

	// Visualization endpoints
	monitoring.HandleFunc("/visualizations/model-drift", handler.GetModelDriftVisualization).Methods("GET")
	monitoring.HandleFunc("/visualizations/ensemble-contribution", handler.GetEnsembleContributionVisualization).Methods("GET")
	monitoring.HandleFunc("/visualizations/uncertainty", handler.GetUncertaintyVisualization).Methods("GET")
	monitoring.HandleFunc("/visualizations/security-compliance", handler.GetSecurityComplianceVisualization).Methods("GET")

	// Export endpoints
	monitoring.HandleFunc("/export/{format}", handler.ExportDashboardData).Methods("GET")
	monitoring.HandleFunc("/export", handler.ExportDashboardData).Methods("GET").Queries("format", "json")

	// Real-time endpoints
	monitoring.HandleFunc("/realtime/dashboard", handler.GetDashboardData).Methods("GET")
	monitoring.HandleFunc("/realtime/health", handler.GetHealthStatus).Methods("GET")

	// Add CORS middleware for monitoring endpoints
	monitoring.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Add logging middleware for monitoring endpoints
	monitoring.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("monitoring v3 request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)
			next.ServeHTTP(w, r)
		})
	})

	logger.Info("advanced monitoring v3 routes configured successfully")
}
