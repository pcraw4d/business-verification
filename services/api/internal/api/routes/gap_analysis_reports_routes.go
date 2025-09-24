package routes

import (
	"net/http"

	"kyb-platform/internal/api/handlers"

	"github.com/gorilla/mux"
)

// RegisterGapAnalysisReportsRoutes registers all gap analysis reports-related routes
func RegisterGapAnalysisReportsRoutes(router *mux.Router) {
	reportsHandler := handlers.NewGapAnalysisReportsHandler()

	// Gap analysis reports routes
	reportsRouter := router.PathPrefix("/v1/reports").Subrouter()

	// Report generation
	reportsRouter.HandleFunc("/generate", reportsHandler.GenerateReport).Methods("POST")
	reportsRouter.HandleFunc("/templates", reportsHandler.GetReportTemplates).Methods("GET")
	reportsRouter.HandleFunc("/schedule", reportsHandler.ScheduleReport).Methods("POST")

	// Report management
	reportsRouter.HandleFunc("/{id}", reportsHandler.GetReportDetails).Methods("GET")
	reportsRouter.HandleFunc("/{id}/download", reportsHandler.DownloadReport).Methods("GET")
	reportsRouter.HandleFunc("/{id}/preview", reportsHandler.PreviewReport).Methods("GET")

	// Report analytics and metrics
	reportsRouter.HandleFunc("/metrics", reportsHandler.GetReportMetrics).Methods("GET")
	reportsRouter.HandleFunc("/recent", reportsHandler.GetRecentReports).Methods("GET")

	// Data export
	reportsRouter.HandleFunc("/export", reportsHandler.ExportReportData).Methods("GET")

	// Health check for gap analysis reports endpoints
	reportsRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"gap-analysis-reports"}`))
	}).Methods("GET")
}
