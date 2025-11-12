package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"kyb-platform/internal/database"
	"kyb-platform/internal/services"
)

// MerchantAnalyticsHandler handles merchant analytics API endpoints
type MerchantAnalyticsHandler struct {
	analyticsService services.MerchantAnalyticsService
	logger           *log.Logger
}

// NewMerchantAnalyticsHandler creates a new merchant analytics handler
func NewMerchantAnalyticsHandler(
	analyticsService services.MerchantAnalyticsService,
	logger *log.Logger,
) *MerchantAnalyticsHandler {
	if logger == nil {
		logger = log.Default()
	}

	return &MerchantAnalyticsHandler{
		analyticsService: analyticsService,
		logger:           logger,
	}
}

// extractMerchantID extracts merchant ID from request path
func (h *MerchantAnalyticsHandler) extractMerchantID(r *http.Request) string {
	// Try Go 1.22+ PathValue first
	if value := r.PathValue("merchantId"); value != "" {
		return value
	}

	// Fallback: extract from URL path
	path := r.URL.Path
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "merchants" && i+1 < len(parts) {
			return parts[i+1]
		}
	}

	return ""
}

// GetMerchantAnalytics handles GET /api/v1/merchants/{merchantId}/analytics
func (h *MerchantAnalyticsHandler) GetMerchantAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract merchantId from path
	merchantID := h.extractMerchantID(r)
	if merchantID == "" {
		h.logger.Printf("Missing merchantId parameter")
		http.Error(w, "merchant ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Getting analytics for merchant: %s", merchantID)

	// Get analytics data from service
	analytics, err := h.analyticsService.GetMerchantAnalytics(ctx, merchantID)
	if err != nil {
		h.logger.Printf("Error getting analytics: %v", err)
		// Check if error is wrapped ErrMerchantNotFound
		if errors.Is(err, database.ErrMerchantNotFound) || strings.Contains(err.Error(), "merchant not found") {
			http.Error(w, "merchant not found", http.StatusNotFound)
			return
		}
		http.Error(w, "failed to retrieve analytics", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analytics); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetWebsiteAnalysis handles GET /api/v1/merchants/{merchantId}/website-analysis
func (h *MerchantAnalyticsHandler) GetWebsiteAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract merchantId from path
	merchantID := h.extractMerchantID(r)
	if merchantID == "" {
		h.logger.Printf("Missing merchantId parameter")
		http.Error(w, "merchant ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Getting website analysis for merchant: %s", merchantID)

	// Get website analysis data from service
	analysis, err := h.analyticsService.GetWebsiteAnalysis(ctx, merchantID)
	if err != nil {
		h.logger.Printf("Error getting website analysis: %v", err)
		// Check if error is wrapped ErrMerchantNotFound
		if errors.Is(err, database.ErrMerchantNotFound) || strings.Contains(err.Error(), "merchant not found") {
			http.Error(w, "merchant not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "merchant has no website URL") {
			http.Error(w, "merchant has no website URL", http.StatusBadRequest)
			return
		}
		http.Error(w, "failed to retrieve website analysis", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(analysis); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

