package routes

import (
	"net/http"

	"github.com/gorilla/mux"

	"kyb-platform/services/risk-assessment-service/internal/handlers"
)

// RegisterPerformanceRoutes registers performance-related routes
func RegisterPerformanceRoutes(router *mux.Router, performanceHandler *handlers.PerformanceHandler) {
	// Performance metrics routes
	performanceRouter := router.PathPrefix("/api/v1/performance").Subrouter()
	
	// Get performance metrics
	performanceRouter.HandleFunc("/metrics", performanceHandler.GetMetrics).Methods("GET")
	
	// Get system health status
	performanceRouter.HandleFunc("/health", performanceHandler.GetHealth).Methods("GET")
	
	// Get system information
	performanceRouter.HandleFunc("/system", performanceHandler.GetSystemInfo).Methods("GET")
	
	// Reset performance metrics
	performanceRouter.HandleFunc("/metrics/reset", performanceHandler.ResetMetrics).Methods("POST")
	
	// Add CORS headers for performance endpoints
	performanceRouter.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	})
}
