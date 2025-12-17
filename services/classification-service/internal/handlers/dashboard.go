package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"kyb-platform/internal/classification/repository"
)

// DashboardHandler handles dashboard API requests
type DashboardHandler struct {
	repo repository.KeywordRepository
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(repo repository.KeywordRepository) *DashboardHandler {
	return &DashboardHandler{
		repo: repo,
	}
}

// GetSummary returns dashboard summary metrics
func (h *DashboardHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	// Get days parameter (default 30)
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	// Get summary from database
	summary, err := h.repo.GetDashboardSummary(r.Context(), days)
	if err != nil {
		http.Error(w, "Failed to get dashboard summary: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": summary,
		"days":    days,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetTimeSeries returns time series data for charts
func (h *DashboardHandler) GetTimeSeries(w http.ResponseWriter, r *http.Request) {
	// Get days parameter (default 30)
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	// Get time series data from database
	timeSeries, err := h.repo.GetTimeSeriesData(r.Context(), days)
	if err != nil {
		http.Error(w, "Failed to get time series data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"time_series": timeSeries,
		"days":        days,
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

