package routes

import (
	"net/http"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// RegisterComplianceStatusRoutes registers compliance status API routes using Go 1.22 ServeMux
func RegisterComplianceStatusRoutes(mux *http.ServeMux, logger *observability.Logger) {
	// Create compliance status handler
	complianceStatusHandler := handlers.NewComplianceStatusHandler(logger)

	// Compliance status endpoints - using Go 1.22 ServeMux pattern
	mux.HandleFunc("GET /v1/compliance/status/{business_id}", complianceStatusHandler.GetComplianceStatusHandler)
	mux.HandleFunc("PUT /v1/compliance/status/{business_id}", complianceStatusHandler.UpdateComplianceStatusHandler)
	mux.HandleFunc("GET /v1/compliance/status/{business_id}/history", complianceStatusHandler.GetComplianceStatusHistoryHandler)

	// Log route registration
	logger.Info("Compliance status routes registered", map[string]interface{}{
		"endpoints": []string{
			"GET /v1/compliance/status/{business_id}",
			"PUT /v1/compliance/status/{business_id}",
			"GET /v1/compliance/status/{business_id}/history",
		},
		"version": "v1",
		"pattern": "Go 1.22 ServeMux",
	})
}
