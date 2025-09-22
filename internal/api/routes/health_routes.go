package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/services"
)

// SetupHealthRoutes sets up health check API routes
func SetupHealthRoutes(router *mux.Router, logger *zap.Logger, healthCheckService *services.HealthCheckService) {
	// Create health check handler
	healthHandler := handlers.NewHealthCheckHandler(logger, healthCheckService)

	// Health check routes
	healthRouter := router.PathPrefix("/health").Subrouter()

	// Basic health check
	healthRouter.HandleFunc("", healthHandler.GetHealth).Methods("GET")
	healthRouter.HandleFunc("/", healthHandler.GetHealth).Methods("GET")

	// Detailed health check
	healthRouter.HandleFunc("/detailed", healthHandler.GetHealthDetailed).Methods("GET")

	// Kubernetes-style health checks
	healthRouter.HandleFunc("/live", healthHandler.GetHealthLiveness).Methods("GET")
	healthRouter.HandleFunc("/ready", healthHandler.GetHealthReadiness).Methods("GET")
	healthRouter.HandleFunc("/startup", healthHandler.GetHealthStartup).Methods("GET")

	// Individual service health checks
	healthRouter.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		check := healthCheckService.CheckAPIHealth()
		w.Header().Set("Content-Type", "application/json")
		if check.Status == services.HealthStatusCritical {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(check)
	}).Methods("GET")

	healthRouter.HandleFunc("/database", func(w http.ResponseWriter, r *http.Request) {
		check := healthCheckService.CheckDatabaseHealth()
		w.Header().Set("Content-Type", "application/json")
		if check.Status == services.HealthStatusCritical {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(check)
	}).Methods("GET")

	healthRouter.HandleFunc("/cache", func(w http.ResponseWriter, r *http.Request) {
		check := healthCheckService.CheckCacheHealth()
		w.Header().Set("Content-Type", "application/json")
		if check.Status == services.HealthStatusCritical {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(check)
	}).Methods("GET")

	healthRouter.HandleFunc("/external-apis", func(w http.ResponseWriter, r *http.Request) {
		check := healthCheckService.CheckExternalAPIsHealth()
		w.Header().Set("Content-Type", "application/json")
		if check.Status == services.HealthStatusCritical {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(check)
	}).Methods("GET")

	healthRouter.HandleFunc("/filesystem", func(w http.ResponseWriter, r *http.Request) {
		check := healthCheckService.CheckFileSystemHealth()
		w.Header().Set("Content-Type", "application/json")
		if check.Status == services.HealthStatusCritical {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(check)
	}).Methods("GET")

	healthRouter.HandleFunc("/memory", func(w http.ResponseWriter, r *http.Request) {
		check := healthCheckService.CheckMemoryHealth()
		w.Header().Set("Content-Type", "application/json")
		if check.Status == services.HealthStatusCritical {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(check)
	}).Methods("GET")

	// Root health check endpoint for load balancers
	router.HandleFunc("/health", healthHandler.GetHealth).Methods("GET")

	// Simple health check for basic monitoring
	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	}).Methods("GET")

	logger.Info("Health check routes configured", zap.String("prefix", "/health"))
}
