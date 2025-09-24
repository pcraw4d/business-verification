package routes

import (
	"net/http"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/compliance"
	"kyb-platform/internal/observability"
)

// RegisterComplianceTrackingRoutes registers compliance tracking API routes using Go 1.22 ServeMux
func RegisterComplianceTrackingRoutes(mux *http.ServeMux, logger *observability.Logger) {
	// Create compliance framework service (dependency)
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	// Create compliance tracking service
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)

	// Create compliance tracking handler
	trackingHandler := handlers.NewComplianceTrackingHandler(logger, trackingService)

	// Compliance tracking endpoints - using Go 1.22 ServeMux pattern
	mux.HandleFunc("GET /v1/compliance/tracking/{business_id}/{framework_id}", trackingHandler.GetComplianceTrackingHandler)
	mux.HandleFunc("PUT /v1/compliance/tracking/{business_id}/{framework_id}", trackingHandler.UpdateComplianceTrackingHandler)

	// Compliance milestone endpoints
	mux.HandleFunc("GET /v1/compliance/milestones", trackingHandler.GetComplianceMilestonesHandler)
	mux.HandleFunc("POST /v1/compliance/milestones", trackingHandler.CreateMilestoneHandler)
	mux.HandleFunc("PUT /v1/compliance/milestones/{milestone_id}", trackingHandler.UpdateMilestoneHandler)

	// Compliance metrics and trends endpoints
	mux.HandleFunc("GET /v1/compliance/metrics/{business_id}/{framework_id}", trackingHandler.GetProgressMetricsHandler)
	mux.HandleFunc("GET /v1/compliance/trends/{business_id}/{framework_id}", trackingHandler.GetComplianceTrendsHandler)

	// Log route registration
	logger.Info("Compliance tracking routes registered", map[string]interface{}{
		"endpoints": []string{
			"GET /v1/compliance/tracking/{business_id}/{framework_id}",
			"PUT /v1/compliance/tracking/{business_id}/{framework_id}",
			"GET /v1/compliance/milestones",
			"POST /v1/compliance/milestones",
			"PUT /v1/compliance/milestones/{milestone_id}",
			"GET /v1/compliance/metrics/{business_id}/{framework_id}",
			"GET /v1/compliance/trends/{business_id}/{framework_id}",
		},
		"version": "v1",
		"pattern": "Go 1.22 ServeMux",
		"service": "ComplianceTrackingService",
	})
}
