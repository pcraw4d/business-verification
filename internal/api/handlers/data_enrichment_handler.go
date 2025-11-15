package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"kyb-platform/internal/services"
)

// DataEnrichmentHandler handles data enrichment API endpoints
type DataEnrichmentHandler struct {
	enrichmentService services.DataEnrichmentService
	logger            *log.Logger
}

// NewDataEnrichmentHandler creates a new data enrichment handler
func NewDataEnrichmentHandler(
	enrichmentService services.DataEnrichmentService,
	logger *log.Logger,
) *DataEnrichmentHandler {
	if logger == nil {
		logger = log.Default()
	}

	return &DataEnrichmentHandler{
		enrichmentService: enrichmentService,
		logger:            logger,
	}
}

// extractMerchantID extracts merchant ID from request path
func (h *DataEnrichmentHandler) extractMerchantID(r *http.Request) string {
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

// TriggerEnrichment handles POST /api/v1/merchants/{merchantId}/enrichment/trigger
func (h *DataEnrichmentHandler) TriggerEnrichment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	merchantID := h.extractMerchantID(r)
	if merchantID == "" {
		h.logger.Printf("Missing merchantId parameter")
		http.Error(w, "merchant ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req struct {
		Source string `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Error decoding request: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Source == "" {
		http.Error(w, "source is required", http.StatusBadRequest)
		return
	}

	h.logger.Printf("Triggering enrichment for merchant: %s, source: %s", merchantID, req.Source)

	// Trigger enrichment
	job, err := h.enrichmentService.TriggerEnrichment(ctx, merchantID, req.Source)
	if err != nil {
		h.logger.Printf("Error triggering enrichment: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(job); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

// GetEnrichmentSources handles GET /api/v1/merchants/{merchantId}/enrichment/sources
func (h *DataEnrichmentHandler) GetEnrichmentSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.Printf("Getting enrichment sources")

	// Get sources
	sources, err := h.enrichmentService.GetEnrichmentSources(ctx)
	if err != nil {
		h.logger.Printf("Error getting enrichment sources: %v", err)
		http.Error(w, "failed to get enrichment sources", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"sources": sources,
	}); err != nil {
		h.logger.Printf("Error encoding response: %v", err)
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}

