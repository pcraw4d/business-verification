package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/api/handlers"
	"kyb-platform/services/risk-assessment-service/internal/ml/testing"
)

// RegisterExperimentRoutes registers experiment-related routes
func RegisterExperimentRoutes(router *mux.Router, experimentManager *testing.ExperimentManager, logger *zap.Logger) {
	experimentHandlers := handlers.NewExperimentHandlers(experimentManager, logger)

	// Experiment management routes
	experimentRouter := router.PathPrefix("/experiments").Subrouter()

	// Create experiments
	experimentRouter.HandleFunc("/model-comparison", experimentHandlers.CreateModelComparisonExperiment).Methods("POST")
	experimentRouter.HandleFunc("/hyperparameter", experimentHandlers.CreateHyperparameterExperiment).Methods("POST")
	experimentRouter.HandleFunc("/feature", experimentHandlers.CreateFeatureExperiment).Methods("POST")
	experimentRouter.HandleFunc("/industry", experimentHandlers.CreateIndustryExperiment).Methods("POST")

	// List and get experiments
	experimentRouter.HandleFunc("", experimentHandlers.ListExperiments).Methods("GET")
	experimentRouter.HandleFunc("/{id}/status", experimentHandlers.GetExperimentStatus).Methods("GET")
	experimentRouter.HandleFunc("/{id}/results", experimentHandlers.GetExperimentResults).Methods("GET")

	// Control experiments
	experimentRouter.HandleFunc("/{id}/stop", experimentHandlers.StopExperiment).Methods("POST")

	// Add middleware for experiment routes
	experimentRouter.Use(experimentMiddleware(logger))
}

// experimentMiddleware adds common middleware for experiment routes
func experimentMiddleware(logger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log experiment requests
			logger.Info("Experiment API request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()))

			// Add CORS headers for experiment endpoints
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
