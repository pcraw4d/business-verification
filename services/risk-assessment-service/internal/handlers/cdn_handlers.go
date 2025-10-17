package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"kyb-platform/services/risk-assessment-service/internal/cdn"

	"go.uber.org/zap"
)

// CDNHandlers handles CDN-related operations
type CDNHandlers struct {
	cloudflare *cdn.CloudFlareCDN
	logger     *zap.Logger
}

// NewCDNHandlers creates a new CDN handlers instance
func NewCDNHandlers(cloudflare *cdn.CloudFlareCDN, logger *zap.Logger) *CDNHandlers {
	return &CDNHandlers{
		cloudflare: cloudflare,
		logger:     logger,
	}
}

// PurgeCacheRequest represents a cache purge request
type PurgeCacheRequest struct {
	URLs            []string `json:"urls,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	Prefixes        []string `json:"prefixes,omitempty"`
	PurgeEverything bool     `json:"purge_everything,omitempty"`
}

// PurgeCacheResponse represents a cache purge response
type PurgeCacheResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	PurgeID string `json:"purge_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

// CDNStatsResponse represents CDN statistics response
type CDNStatsResponse struct {
	Success bool          `json:"success"`
	Stats   *cdn.CDNStats `json:"stats"`
	Error   string        `json:"error,omitempty"`
}

// ZoneInfoResponse represents zone information response
type ZoneInfoResponse struct {
	Success bool          `json:"success"`
	Zone    *cdn.ZoneInfo `json:"zone"`
	Error   string        `json:"error,omitempty"`
}

// PurgeCache purges CDN cache
func (ch *CDNHandlers) PurgeCache(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req PurgeCacheRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.logger.Error("Failed to parse purge cache request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ch.logger.Info("Processing cache purge request",
		zap.Strings("urls", req.URLs),
		zap.Strings("tags", req.Tags),
		zap.Strings("prefixes", req.Prefixes),
		zap.Bool("purge_everything", req.PurgeEverything))

	// Create CloudFlare purge request
	purgeReq := &cdn.PurgeRequest{
		URLs:            req.URLs,
		Tags:            req.Tags,
		Prefixes:        req.Prefixes,
		PurgeEverything: req.PurgeEverything,
	}

	// Execute purge
	purgeResp, err := ch.cloudflare.PurgeCache(ctx, purgeReq)
	if err != nil {
		ch.logger.Error("Failed to purge cache", zap.Error(err))

		response := PurgeCacheResponse{
			Success: false,
			Message: "Failed to purge cache",
			Error:   err.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return success response
	response := PurgeCacheResponse{
		Success: true,
		Message: "Cache purged successfully",
		PurgeID: purgeResp.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Cache purge completed successfully",
		zap.String("purge_id", purgeResp.ID))
}

// PurgeByURL purges cache by URL
func (ch *CDNHandlers) PurgeByURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req struct {
		URLs []string `json:"urls"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.logger.Error("Failed to parse purge by URL request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.URLs) == 0 {
		http.Error(w, "URLs are required", http.StatusBadRequest)
		return
	}

	ch.logger.Info("Processing purge by URL request",
		zap.Strings("urls", req.URLs))

	// Execute purge
	purgeResp, err := ch.cloudflare.PurgeByURL(ctx, req.URLs)
	if err != nil {
		ch.logger.Error("Failed to purge cache by URL", zap.Error(err))

		response := PurgeCacheResponse{
			Success: false,
			Message: "Failed to purge cache by URL",
			Error:   err.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return success response
	response := PurgeCacheResponse{
		Success: true,
		Message: "Cache purged by URL successfully",
		PurgeID: purgeResp.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Cache purge by URL completed successfully",
		zap.String("purge_id", purgeResp.ID))
}

// PurgeByTag purges cache by tag
func (ch *CDNHandlers) PurgeByTag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req struct {
		Tags []string `json:"tags"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.logger.Error("Failed to parse purge by tag request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Tags) == 0 {
		http.Error(w, "Tags are required", http.StatusBadRequest)
		return
	}

	ch.logger.Info("Processing purge by tag request",
		zap.Strings("tags", req.Tags))

	// Execute purge
	purgeResp, err := ch.cloudflare.PurgeByTag(ctx, req.Tags)
	if err != nil {
		ch.logger.Error("Failed to purge cache by tag", zap.Error(err))

		response := PurgeCacheResponse{
			Success: false,
			Message: "Failed to purge cache by tag",
			Error:   err.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return success response
	response := PurgeCacheResponse{
		Success: true,
		Message: "Cache purged by tag successfully",
		PurgeID: purgeResp.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Cache purge by tag completed successfully",
		zap.String("purge_id", purgeResp.ID))
}

// PurgeByPrefix purges cache by prefix
func (ch *CDNHandlers) PurgeByPrefix(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req struct {
		Prefixes []string `json:"prefixes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.logger.Error("Failed to parse purge by prefix request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Prefixes) == 0 {
		http.Error(w, "Prefixes are required", http.StatusBadRequest)
		return
	}

	ch.logger.Info("Processing purge by prefix request",
		zap.Strings("prefixes", req.Prefixes))

	// Execute purge
	purgeResp, err := ch.cloudflare.PurgeByPrefix(ctx, req.Prefixes)
	if err != nil {
		ch.logger.Error("Failed to purge cache by prefix", zap.Error(err))

		response := PurgeCacheResponse{
			Success: false,
			Message: "Failed to purge cache by prefix",
			Error:   err.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return success response
	response := PurgeCacheResponse{
		Success: true,
		Message: "Cache purged by prefix successfully",
		PurgeID: purgeResp.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Cache purge by prefix completed successfully",
		zap.String("purge_id", purgeResp.ID))
}

// PurgeEverything purges entire cache
func (ch *CDNHandlers) PurgeEverything(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ch.logger.Info("Processing purge everything request")

	// Execute purge
	purgeResp, err := ch.cloudflare.PurgeEverything(ctx)
	if err != nil {
		ch.logger.Error("Failed to purge entire cache", zap.Error(err))

		response := PurgeCacheResponse{
			Success: false,
			Message: "Failed to purge entire cache",
			Error:   err.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return success response
	response := PurgeCacheResponse{
		Success: true,
		Message: "Entire cache purged successfully",
		PurgeID: purgeResp.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Entire cache purge completed successfully",
		zap.String("purge_id", purgeResp.ID))
}

// GetCDNStats returns CDN statistics
func (ch *CDNHandlers) GetCDNStats(w http.ResponseWriter, r *http.Request) {
	ch.logger.Debug("Retrieving CDN statistics")

	// Get stats
	stats := ch.cloudflare.GetStats()

	// Return response
	response := CDNStatsResponse{
		Success: true,
		Stats:   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Debug("CDN statistics retrieved successfully")
}

// GetZoneInfo returns zone information
func (ch *CDNHandlers) GetZoneInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ch.logger.Debug("Retrieving zone information")

	// Get zone info
	zone, err := ch.cloudflare.GetZoneInfo(ctx)
	if err != nil {
		ch.logger.Error("Failed to get zone information", zap.Error(err))

		response := ZoneInfoResponse{
			Success: false,
			Error:   err.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return response
	response := ZoneInfoResponse{
		Success: true,
		Zone:    zone,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Debug("Zone information retrieved successfully",
		zap.String("zone_id", zone.ID),
		zap.String("zone_name", zone.Name))
}

// ConfigureCacheRules configures default cache rules
func (ch *CDNHandlers) ConfigureCacheRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	ch.logger.Info("Configuring default cache rules")

	// Configure cache rules
	err := ch.cloudflare.ConfigureCacheRules(ctx)
	if err != nil {
		ch.logger.Error("Failed to configure cache rules", zap.Error(err))

		response := map[string]interface{}{
			"success": false,
			"message": "Failed to configure cache rules",
			"error":   err.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "Cache rules configured successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Cache rules configured successfully")
}

// ConfigureGeographicRouting configures geographic routing
func (ch *CDNHandlers) ConfigureGeographicRouting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request
	var req struct {
		Regions []string `json:"regions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.logger.Error("Failed to parse geographic routing request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Regions) == 0 {
		http.Error(w, "Regions are required", http.StatusBadRequest)
		return
	}

	ch.logger.Info("Configuring geographic routing",
		zap.Strings("regions", req.Regions))

	// Configure geographic routing
	err := ch.cloudflare.ConfigureGeographicRouting(ctx, req.Regions)
	if err != nil {
		ch.logger.Error("Failed to configure geographic routing", zap.Error(err))

		response := map[string]interface{}{
			"success": false,
			"message": "Failed to configure geographic routing",
			"error":   err.Error(),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "Geographic routing configured successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Info("Geographic routing configured successfully")
}

// UpdateCDNStats updates CDN statistics
func (ch *CDNHandlers) UpdateCDNStats(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		Hits         int64         `json:"hits"`
		Misses       int64         `json:"misses"`
		Bandwidth    int64         `json:"bandwidth"`
		ResponseTime time.Duration `json:"response_time"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ch.logger.Error("Failed to parse update stats request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ch.logger.Debug("Updating CDN statistics",
		zap.Int64("hits", req.Hits),
		zap.Int64("misses", req.Misses),
		zap.Int64("bandwidth", req.Bandwidth),
		zap.Duration("response_time", req.ResponseTime))

	// Update stats
	ch.cloudflare.UpdateStats(req.Hits, req.Misses, req.Bandwidth, req.ResponseTime)

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "CDN statistics updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	ch.logger.Debug("CDN statistics updated successfully")
}
