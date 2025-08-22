package routes

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/modules/classification_monitoring"
)

// RegisterClassificationMonitoringRoutes registers all classification monitoring routes
func RegisterClassificationMonitoringRoutes(
	router *mux.Router,
	accuracyTracker *classification_monitoring.AccuracyTracker,
	misclassificationDetector *classification_monitoring.MisclassificationDetector,
	metricsCollector *classification_monitoring.AccuracyMetricsCollector,
	alertingSystem *classification_monitoring.AccuracyAlertingSystem,
	logger *zap.Logger,
) {
	// Create the monitoring handler
	monitoringHandler := handlers.NewClassificationMonitoringHandler(
		accuracyTracker,
		misclassificationDetector,
		metricsCollector,
		alertingSystem,
		logger,
	)

	// Create a subrouter for monitoring routes
	monitoringRouter := router.PathPrefix("/api/v3/monitoring").Subrouter()

	// Accuracy metrics endpoints
	monitoringRouter.HandleFunc("/accuracy/metrics", monitoringHandler.GetAccuracyMetrics).Methods("GET")
	monitoringRouter.HandleFunc("/accuracy/track", monitoringHandler.TrackClassification).Methods("POST")

	// Misclassification endpoints
	monitoringRouter.HandleFunc("/misclassifications", monitoringHandler.GetMisclassifications).Methods("GET")
	monitoringRouter.HandleFunc("/patterns", monitoringHandler.GetErrorPatterns).Methods("GET")
	monitoringRouter.HandleFunc("/statistics", monitoringHandler.GetErrorStatistics).Methods("GET")

	// Alert endpoints
	monitoringRouter.HandleFunc("/alerts", monitoringHandler.GetActiveAlerts).Methods("GET")
	monitoringRouter.HandleFunc("/alerts/history", monitoringHandler.GetAlertHistory).Methods("GET")
	monitoringRouter.HandleFunc("/alerts/{alertId}/resolve", monitoringHandler.ResolveAlert).Methods("POST")

	// Alert rule management endpoints
	monitoringRouter.HandleFunc("/alerts/rules", monitoringHandler.GetAlertRules).Methods("GET")
	monitoringRouter.HandleFunc("/alerts/rules", monitoringHandler.CreateAlertRule).Methods("POST")
	monitoringRouter.HandleFunc("/alerts/rules/{ruleId}", monitoringHandler.UpdateAlertRule).Methods("PUT")
	monitoringRouter.HandleFunc("/alerts/rules/{ruleId}", monitoringHandler.DeleteAlertRule).Methods("DELETE")

	// Metrics collection endpoints
	monitoringRouter.HandleFunc("/metrics/collect", monitoringHandler.CollectMetrics).Methods("POST")

	// Reporting endpoints
	monitoringRouter.HandleFunc("/reports/accuracy", monitoringHandler.GenerateReport).Methods("GET")

	// Health check endpoint
	monitoringRouter.HandleFunc("/health", monitoringHandler.GetHealthStatus).Methods("GET")

	// Add CORS headers for all monitoring routes
	monitoringRouter.Use(corsMiddleware)

	// Add request logging middleware
	monitoringRouter.Use(requestLoggingMiddleware(logger))

	logger.Info("Classification monitoring routes registered",
		zap.String("base_path", "/api/v3/monitoring"),
		zap.Int("route_count", 12))
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
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
}

// requestLoggingMiddleware logs incoming requests
func requestLoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			logger.Info("Request completed",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("query", r.URL.RawQuery),
				zap.Int("status", wrapped.statusCode),
				zap.Duration("duration", duration),
				zap.String("user_agent", r.UserAgent()),
				zap.String("remote_addr", r.RemoteAddr))
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
