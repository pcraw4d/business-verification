package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/api/handlers"
)

// SetupMonitoringRoutes sets up monitoring API routes
func SetupMonitoringRoutes(router *mux.Router, logger *zap.Logger) {
	// Create monitoring handler
	monitoringHandler := handlers.NewMonitoringHandler(logger)

	// Monitoring API routes
	monitoringRouter := router.PathPrefix("/api/v3/monitoring").Subrouter()

	// Dashboard metrics
	monitoringRouter.HandleFunc("/metrics", monitoringHandler.GetMetrics).Methods("GET")

	// Alerts
	monitoringRouter.HandleFunc("/alerts", monitoringHandler.GetAlerts).Methods("GET")

	// Health checks
	monitoringRouter.HandleFunc("/health", monitoringHandler.GetHealthChecks).Methods("GET")

	// Health check endpoint for load balancers
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","timestamp":"` +
			`"}`))
	}).Methods("GET")

	// Metrics endpoint for Prometheus
	router.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# KYB Platform Metrics\n# This is a placeholder for Prometheus metrics\n"))
	}).Methods("GET")

	logger.Info("Monitoring routes configured", zap.String("prefix", "/api/v3/monitoring"))
}
