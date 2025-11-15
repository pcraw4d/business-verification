package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"kyb-platform/internal/database"
	"kyb-platform/internal/services"
)

// RiskIndicatorsHandler handles risk indicators API endpoints
type RiskIndicatorsHandler struct {
	indicatorsService services.RiskIndicatorsService
	logger            *log.Logger
}

// NewRiskIndicatorsHandler creates a new risk indicators handler
func NewRiskIndicatorsHandler(
	indicatorsService services.RiskIndicatorsService,
	logger *log.Logger,
) *RiskIndicatorsHandler {
	if logger == nil {
		logger = log.Default()
	}

	return &RiskIndicatorsHandler{
		indicatorsService: indicatorsService,
		logger:            logger,
	}
}

// extractMerchantID extracts merchant ID from request path
func (h *RiskIndicatorsHandler) extractMerchantID(r *http.Request) string {
	// Try Go 1.22+ PathValue first
	if value := r.PathValue("merchantId"); value != "" {
		return value
	}

	// Fallback: extract from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if (part == "indicators" || part == "alerts" || part == "merchants") && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}

// GetRiskIndicators handles GET /api/v1/risk/indicators/{merchantId}
func (h *RiskIndicatorsHandler) GetRiskIndicators(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	merchantID := h.extractMerchantID(r)
	if merchantID == "" {
		h.logger.Printf("Missing merchantId parameter")
		http.Error(w, "merchant ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	filters := &database.RiskIndicatorFilters{
		Severity: r.URL.Query().Get("severity"),
		Status:   r.URL.Query().Get("status"),
	}

	h.logger.Printf("Getting risk indicators for merchant: %s", merchantID)

	// Get indicators
	indicators, err := h.indicatorsService.GetRiskIndicators(ctx, merchantID, filters)
	if err != nil {
		h.logger.Printf("Error getting risk indicators: %v", err)
		http.Error(w, "failed to get risk indicators", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(indicators); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetRiskAlerts handles GET /api/v1/risk/alerts/{merchantId}
func (h *RiskIndicatorsHandler) GetRiskAlerts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	merchantID := h.extractMerchantID(r)
	if merchantID == "" {
		h.logger.Printf("Missing merchantId parameter")
		http.Error(w, "merchant ID is required", http.StatusBadRequest)
		return
	}

	// Parse query parameters
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")

	// Only return active alerts (status = "active" by default)
	if status == "" {
		status = "active"
	}

	filters := &database.RiskIndicatorFilters{
		Severity: severity,
		Status:   status,
	}

	h.logger.Printf("Getting risk alerts for merchant: %s", merchantID)

	// Get indicators (alerts are active indicators)
	indicators, err := h.indicatorsService.GetRiskIndicators(ctx, merchantID, filters)
	if err != nil {
		h.logger.Printf("Error getting risk alerts: %v", err)
		http.Error(w, "failed to get risk alerts", http.StatusInternalServerError)
		return
	}

	// Filter to only active indicators for alerts
	alerts := []interface{}{}
	for _, indicator := range indicators.Indicators {
		if indicator.Status == "active" {
			alerts = append(alerts, indicator)
		}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"merchantId": merchantID,
		"alerts":     alerts,
		"count":      len(alerts),
	}); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

