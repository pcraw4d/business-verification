package routes

import (
	"github.com/gorilla/mux"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
)

// RegisterPatternAnalysisRoutes registers pattern analysis API routes
func RegisterPatternAnalysisRoutes(router *mux.Router, handler *handlers.PatternAnalysisHandler) {
	// Pattern analysis endpoints
	router.HandleFunc("/api/v1/pattern-analysis/analyze", handler.AnalyzeMisclassificationsHandler).Methods("POST")
	router.HandleFunc("/api/v1/pattern-analysis/patterns", handler.GetPatternsHandler).Methods("GET")
	router.HandleFunc("/api/v1/pattern-analysis/patterns/type/{type}", handler.GetPatternsByTypeHandler).Methods("GET")
	router.HandleFunc("/api/v1/pattern-analysis/patterns/severity/{severity}", handler.GetPatternsBySeverityHandler).Methods("GET")
	router.HandleFunc("/api/v1/pattern-analysis/patterns/{id}", handler.GetPatternDetailsHandler).Methods("GET")
	router.HandleFunc("/api/v1/pattern-analysis/history", handler.GetPatternHistoryHandler).Methods("GET")
	router.HandleFunc("/api/v1/pattern-analysis/summary", handler.GetPatternSummaryHandler).Methods("GET")
	router.HandleFunc("/api/v1/pattern-analysis/recommendations", handler.GetRecommendationsHandler).Methods("GET")
	router.HandleFunc("/api/v1/pattern-analysis/health", handler.HealthCheckHandler).Methods("GET")
}
