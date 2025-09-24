package routes

import (
	"github.com/gorilla/mux"

	"kyb-platform/internal/api/handlers"
)

// RegisterAlgorithmOptimizationRoutes registers algorithm optimization API routes
func RegisterAlgorithmOptimizationRoutes(router *mux.Router, handler *handlers.AlgorithmOptimizationHandler) {
	// Algorithm optimization endpoints
	router.HandleFunc("/api/v1/algorithm-optimization/analyze", handler.AnalyzeAndOptimizeHandler).Methods("POST")
	router.HandleFunc("/api/v1/algorithm-optimization/history", handler.GetOptimizationHistoryHandler).Methods("GET")
	router.HandleFunc("/api/v1/algorithm-optimization/active", handler.GetActiveOptimizationsHandler).Methods("GET")
	router.HandleFunc("/api/v1/algorithm-optimization/summary", handler.GetOptimizationSummaryHandler).Methods("GET")
	router.HandleFunc("/api/v1/algorithm-optimization/recommendations", handler.GetOptimizationRecommendationsHandler).Methods("GET")

	// Specific optimization endpoints
	router.HandleFunc("/api/v1/algorithm-optimization/{id}", handler.GetOptimizationByIDHandler).Methods("GET")
	router.HandleFunc("/api/v1/algorithm-optimization/{id}/cancel", handler.CancelOptimizationHandler).Methods("POST")
	router.HandleFunc("/api/v1/algorithm-optimization/{id}/rollback", handler.RollbackOptimizationHandler).Methods("POST")

	// Filtered optimization endpoints
	router.HandleFunc("/api/v1/algorithm-optimization/type/{type}", handler.GetOptimizationsByTypeHandler).Methods("GET")
	router.HandleFunc("/api/v1/algorithm-optimization/algorithm/{algorithm_id}", handler.GetOptimizationsByAlgorithmHandler).Methods("GET")
}
