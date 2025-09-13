package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pcraw4d/business-verification/internal/api/handlers"
)

// RegisterGapTrackingRoutes registers all gap tracking-related routes
func RegisterGapTrackingRoutes(router *mux.Router) {
	gapTrackingHandler := handlers.NewGapTrackingHandler()

	// Gap tracking routes
	gapTrackingRouter := router.PathPrefix("/v1/gap-tracking").Subrouter()

	// Metrics and overview
	gapTrackingRouter.HandleFunc("/metrics", gapTrackingHandler.GetTrackingMetrics).Methods("GET")

	// Gap tracking management
	gapTrackingRouter.HandleFunc("/gaps", gapTrackingHandler.GetGapTrackingList).Methods("GET")
	gapTrackingRouter.HandleFunc("/gaps", gapTrackingHandler.CreateGapTracking).Methods("POST")
	gapTrackingRouter.HandleFunc("/gaps/{id}", gapTrackingHandler.GetGapTrackingDetails).Methods("GET")
	gapTrackingRouter.HandleFunc("/gaps/{id}/progress", gapTrackingHandler.UpdateGapProgress).Methods("PUT")

	// Progress and history
	gapTrackingRouter.HandleFunc("/gaps/{id}/history", gapTrackingHandler.GetProgressHistory).Methods("GET")

	// Team performance
	gapTrackingRouter.HandleFunc("/teams/{team}/performance", gapTrackingHandler.GetTeamPerformance).Methods("GET")

	// Reporting
	gapTrackingRouter.HandleFunc("/reports/export", gapTrackingHandler.ExportTrackingReport).Methods("GET")

	// Health check for gap tracking endpoints
	gapTrackingRouter.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"gap-tracking-system"}`))
	}).Methods("GET")
}
