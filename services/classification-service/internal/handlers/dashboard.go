package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"kyb-platform/internal/classification/repository"
)

// DashboardHandler handles dashboard API requests
type DashboardHandler struct {
	repo   repository.KeywordRepository
	logger *log.Logger
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(repo repository.KeywordRepository) *DashboardHandler {
	return &DashboardHandler{
		repo:   repo,
		logger: log.Default(),
	}
}

// NewDashboardHandlerWithLogger creates a new dashboard handler with custom logger
func NewDashboardHandlerWithLogger(repo repository.KeywordRepository, logger *log.Logger) *DashboardHandler {
	return &DashboardHandler{
		repo:   repo,
		logger: logger,
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

	h.logger.Printf("📊 [Dashboard] Getting summary for %d days", days)

	// Get summary from database
	summary, err := h.repo.GetDashboardSummary(r.Context(), days)
	if err != nil {
		h.logger.Printf("❌ [Dashboard] Failed to get summary: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Failed to get dashboard summary",
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"metrics": summary,
		"days":    days,
	}); err != nil {
		h.logger.Printf("❌ [Dashboard] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Printf("✅ [Dashboard] Summary returned successfully (%d metrics)", len(summary))
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

	h.logger.Printf("📊 [Dashboard] Getting time series for %d days", days)

	// Get time series data from database
	timeSeries, err := h.repo.GetTimeSeriesData(r.Context(), days)
	if err != nil {
		h.logger.Printf("❌ [Dashboard] Failed to get time series: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Failed to get time series data",
			"message": err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"time_series": timeSeries,
		"days":        days,
	}); err != nil {
		h.logger.Printf("❌ [Dashboard] Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Printf("✅ [Dashboard] Time series returned successfully (%d data points)", len(timeSeries))
}

