package routes

import (
	"github.com/gorilla/mux"

	"kyb-platform/internal/api/handlers"
)

// RegisterAccuracyValidationRoutes registers all accuracy validation API routes
func RegisterAccuracyValidationRoutes(router *mux.Router, handler *handlers.ClassificationOptimizationValidationHandler) {
	// Base path for accuracy validation endpoints
	basePath := "/api/v1/accuracy-validation"

	// Accuracy validation endpoints
	router.HandleFunc(basePath+"/validate", handler.ValidateAccuracy).Methods("POST")
	router.HandleFunc(basePath+"/cross-validation", handler.PerformCrossValidation).Methods("POST")

	// History and summary endpoints
	router.HandleFunc(basePath+"/history", handler.GetValidationHistory).Methods("GET")
	router.HandleFunc(basePath+"/summary", handler.GetValidationSummary).Methods("GET")
	router.HandleFunc(basePath+"/active", handler.GetActiveValidations).Methods("GET")

	// Specific validation endpoints
	router.HandleFunc(basePath+"/validation/{id}", handler.GetValidationByID).Methods("GET")
	router.HandleFunc(basePath+"/validation/{id}/cancel", handler.CancelValidation).Methods("POST")

	// Filtered history endpoints
	router.HandleFunc(basePath+"/algorithm/{algorithm_id}", handler.GetValidationsByAlgorithm).Methods("GET")
	router.HandleFunc(basePath+"/type/{type}", handler.GetValidationsByType).Methods("GET")

	// Health check endpoint
	router.HandleFunc(basePath+"/health", handler.HealthCheck).Methods("GET")
}
