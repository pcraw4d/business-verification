package routes

import (
	"net/http"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// RegisterComplianceReportingRoutes registers compliance reporting API routes using Go 1.22 ServeMux
func RegisterComplianceReportingRoutes(mux *http.ServeMux, logger *observability.Logger) {
	// Create compliance framework service (dependency)
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	// Create compliance tracking service (dependency)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)

	// Create compliance reporting service
	reportingService := compliance.NewComplianceReportingService(logger, frameworkService, trackingService)

	// Create compliance reporting handler
	reportingHandler := handlers.NewComplianceReportingHandler(logger, reportingService)

	// Compliance reporting endpoints - using Go 1.22 ServeMux pattern
	mux.HandleFunc("POST /v1/compliance/reports", reportingHandler.GenerateReportHandler)
	mux.HandleFunc("GET /v1/compliance/reports/{report_id}", reportingHandler.GetReportHandler)
	mux.HandleFunc("GET /v1/compliance/reports", reportingHandler.ListReportsHandler)
	mux.HandleFunc("GET /v1/compliance/reports/{report_id}/export", reportingHandler.ExportReportHandler)

	// Report template endpoints
	mux.HandleFunc("GET /v1/compliance/report-templates", reportingHandler.GetReportTemplatesHandler)

	// Log route registration
	logger.Info("Compliance reporting routes registered", map[string]interface{}{
		"endpoints": []string{
			"POST /v1/compliance/reports",
			"GET /v1/compliance/reports/{report_id}",
			"GET /v1/compliance/reports",
			"GET /v1/compliance/reports/{report_id}/export",
			"GET /v1/compliance/report-templates",
		},
		"version": "v1",
		"pattern": "Go 1.22 ServeMux",
		"service": "ComplianceReportingService",
	})
}
