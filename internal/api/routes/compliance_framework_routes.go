package routes

import (
	"net/http"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// RegisterComplianceFrameworkRoutes registers compliance framework API routes using Go 1.22 ServeMux
func RegisterComplianceFrameworkRoutes(mux *http.ServeMux, logger *observability.Logger) {
	// Create compliance framework service
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	// Create compliance framework handler
	frameworkHandler := handlers.NewComplianceFrameworkHandler(logger, frameworkService)

	// Compliance framework endpoints - using Go 1.22 ServeMux pattern
	mux.HandleFunc("GET /v1/compliance/frameworks", frameworkHandler.GetFrameworksHandler)
	mux.HandleFunc("GET /v1/compliance/frameworks/{framework_id}", frameworkHandler.GetFrameworkHandler)
	mux.HandleFunc("GET /v1/compliance/frameworks/{framework_id}/requirements", frameworkHandler.GetFrameworkRequirementsHandler)

	// Compliance assessment endpoints
	mux.HandleFunc("POST /v1/compliance/assessments", frameworkHandler.CreateAssessmentHandler)
	mux.HandleFunc("GET /v1/compliance/assessments/{assessment_id}", frameworkHandler.GetAssessmentHandler)
	mux.HandleFunc("PUT /v1/compliance/assessments/{assessment_id}", frameworkHandler.UpdateAssessmentHandler)
	mux.HandleFunc("GET /v1/compliance/businesses/{business_id}/assessments", frameworkHandler.GetBusinessAssessmentsHandler)

	// Log route registration
	logger.Info("Compliance framework routes registered", map[string]interface{}{
		"endpoints": []string{
			"GET /v1/compliance/frameworks",
			"GET /v1/compliance/frameworks/{framework_id}",
			"GET /v1/compliance/frameworks/{framework_id}/requirements",
			"POST /v1/compliance/assessments",
			"GET /v1/compliance/assessments/{assessment_id}",
			"PUT /v1/compliance/assessments/{assessment_id}",
			"GET /v1/compliance/businesses/{business_id}/assessments",
		},
		"version": "v1",
		"pattern": "Go 1.22 ServeMux",
		"service": "ComplianceFrameworkService",
	})
}
