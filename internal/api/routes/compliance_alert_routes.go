package routes

import (
	"net/http"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// RegisterComplianceAlertRoutes registers compliance alert API routes using Go 1.22 ServeMux
func RegisterComplianceAlertRoutes(mux *http.ServeMux, logger *observability.Logger) {
	// Create compliance framework service (dependency)
	frameworkService := compliance.NewComplianceFrameworkService(logger)

	// Create compliance tracking service (dependency)
	trackingService := compliance.NewComplianceTrackingService(logger, frameworkService)

	// Create compliance alert service
	alertService := compliance.NewComplianceAlertService(logger, frameworkService, trackingService)

	// Create compliance alert handler
	alertHandler := handlers.NewComplianceAlertHandler(logger, alertService)

	// Compliance alert endpoints - using Go 1.22 ServeMux pattern
	mux.HandleFunc("POST /v1/compliance/alerts", alertHandler.CreateAlertHandler)
	mux.HandleFunc("GET /v1/compliance/alerts/{alert_id}", alertHandler.GetAlertHandler)
	mux.HandleFunc("GET /v1/compliance/alerts", alertHandler.ListAlertsHandler)
	mux.HandleFunc("PUT /v1/compliance/alerts/{alert_id}/status", alertHandler.UpdateAlertStatusHandler)

	// Alert rule endpoints
	mux.HandleFunc("POST /v1/compliance/alert-rules", alertHandler.CreateAlertRuleHandler)
	mux.HandleFunc("POST /v1/compliance/alerts/evaluate", alertHandler.EvaluateAlertRulesHandler)

	// Notification endpoints
	mux.HandleFunc("GET /v1/compliance/notifications", alertHandler.GetNotificationsHandler)

	// Log route registration
	logger.Info("Compliance alert routes registered", map[string]interface{}{
		"endpoints": []string{
			"POST /v1/compliance/alerts",
			"GET /v1/compliance/alerts/{alert_id}",
			"GET /v1/compliance/alerts",
			"PUT /v1/compliance/alerts/{alert_id}/status",
			"POST /v1/compliance/alert-rules",
			"POST /v1/compliance/alerts/evaluate",
			"GET /v1/compliance/notifications",
		},
		"version": "v1",
		"pattern": "Go 1.22 ServeMux",
		"service": "ComplianceAlertService",
	})
}
