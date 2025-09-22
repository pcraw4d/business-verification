package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"kyb-platform/internal/external"
	"go.uber.org/zap"
)

// BusinessComparatorHandler handles business comparison API requests
type BusinessComparatorHandler struct {
	comparator *external.BusinessComparator
	logger     *zap.Logger
}

// NewBusinessComparatorHandler creates a new business comparator handler
func NewBusinessComparatorHandler(comparator *external.BusinessComparator, logger *zap.Logger) *BusinessComparatorHandler {
	return &BusinessComparatorHandler{
		comparator: comparator,
		logger:     logger,
	}
}

// CompareBusinessRequest represents the request for business comparison
type CompareBusinessRequest struct {
	Claimed   *external.ComparisonBusinessInfo `json:"claimed"`
	Extracted *external.ComparisonBusinessInfo `json:"extracted"`
	Config    *ComparisonConfigRequest         `json:"config,omitempty"`
}

// ComparisonConfigRequest represents configuration for comparison
type ComparisonConfigRequest struct {
	MinSimilarityThreshold   *float64                  `json:"min_similarity_threshold,omitempty"`
	MaxEditDistance          *int                      `json:"max_edit_distance,omitempty"`
	PhoneValidationEnabled   *bool                     `json:"phone_validation_enabled,omitempty"`
	EmailValidationEnabled   *bool                     `json:"email_validation_enabled,omitempty"`
	AddressValidationEnabled *bool                     `json:"address_validation_enabled,omitempty"`
	MaxDistanceKm            *float64                  `json:"max_distance_km,omitempty"`
	LocationFuzzyMatch       *bool                     `json:"location_fuzzy_match,omitempty"`
	Weights                  *ComparisonWeightsRequest `json:"weights,omitempty"`
}

// ComparisonWeightsRequest represents weights for different fields
type ComparisonWeightsRequest struct {
	BusinessName    *float64 `json:"business_name,omitempty"`
	PhoneNumber     *float64 `json:"phone_number,omitempty"`
	EmailAddress    *float64 `json:"email_address,omitempty"`
	PhysicalAddress *float64 `json:"physical_address,omitempty"`
	Website         *float64 `json:"website,omitempty"`
	Industry        *float64 `json:"industry,omitempty"`
}

// CompareBusinessResponse represents the response from business comparison
type CompareBusinessResponse struct {
	Success bool                       `json:"success"`
	Result  *external.ComparisonResult `json:"result,omitempty"`
	Error   string                     `json:"error,omitempty"`
	Config  *external.ComparisonConfig `json:"config,omitempty"`
}

// CompareBusiness handles business comparison requests
func (h *BusinessComparatorHandler) CompareBusiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req CompareBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Claimed == nil || req.Extracted == nil {
		h.logger.Error("Missing required fields in request")
		http.Error(w, "Missing claimed or extracted business information", http.StatusBadRequest)
		return
	}

	// Create custom comparator if config is provided
	comparator := h.comparator
	if req.Config != nil {
		customConfig := h.createCustomConfig(req.Config)
		comparator = external.NewBusinessComparator(h.logger, customConfig)
	}

	// Perform comparison
	result, err := comparator.CompareBusinessInfo(r.Context(), req.Claimed, req.Extracted)
	if err != nil {
		h.logger.Error("Business comparison failed", zap.Error(err))
		http.Error(w, "Comparison failed", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := CompareBusinessResponse{
		Success: true,
		Result:  result,
		Config:  comparator.GetConfig(),
	}

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Business comparison completed successfully",
		zap.Float64("overall_score", result.OverallScore),
		zap.String("verification_status", result.VerificationStatus),
		zap.String("confidence_level", result.ConfidenceLevel))
}

// GetComparisonConfig returns the current comparison configuration
func (h *BusinessComparatorHandler) GetComparisonConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	config := h.comparator.GetConfig()

	response := map[string]interface{}{
		"success": true,
		"config":  config,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode config response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateComparisonConfig updates the comparison configuration
func (h *BusinessComparatorHandler) UpdateComparisonConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var configReq ComparisonConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&configReq); err != nil {
		h.logger.Error("Failed to decode config request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create new comparator with updated config
	customConfig := h.createCustomConfig(&configReq)
	newComparator := external.NewBusinessComparator(h.logger, customConfig)
	h.comparator = newComparator

	response := map[string]interface{}{
		"success": true,
		"message": "Configuration updated successfully",
		"config":  customConfig,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode config update response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Comparison configuration updated successfully")
}

// CompareBusinessBatch handles batch business comparison requests
func (h *BusinessComparatorHandler) CompareBusinessBatch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req struct {
		Comparisons []CompareBusinessRequest `json:"comparisons"`
		Config      *ComparisonConfigRequest `json:"config,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode batch request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Comparisons) == 0 {
		h.logger.Error("No comparisons provided in batch request")
		http.Error(w, "No comparisons provided", http.StatusBadRequest)
		return
	}

	// Limit batch size
	if len(req.Comparisons) > 100 {
		h.logger.Error("Batch size too large", zap.Int("size", len(req.Comparisons)))
		http.Error(w, "Batch size too large (max 100)", http.StatusBadRequest)
		return
	}

	// Create custom comparator if config is provided
	comparator := h.comparator
	if req.Config != nil {
		customConfig := h.createCustomConfig(req.Config)
		comparator = external.NewBusinessComparator(h.logger, customConfig)
	}

	// Process each comparison
	results := make([]*external.ComparisonResult, 0, len(req.Comparisons))
	errors := make([]string, 0)

	for i, comparison := range req.Comparisons {
		if comparison.Claimed == nil || comparison.Extracted == nil {
			errors = append(errors, "Comparison "+strconv.Itoa(i+1)+": Missing claimed or extracted business information")
			continue
		}

		result, err := comparator.CompareBusinessInfo(r.Context(), comparison.Claimed, comparison.Extracted)
		if err != nil {
			h.logger.Error("Batch comparison failed",
				zap.Int("index", i),
				zap.Error(err))
			errors = append(errors, "Comparison "+strconv.Itoa(i+1)+": "+err.Error())
			continue
		}

		results = append(results, result)
	}

	// Prepare response
	response := map[string]interface{}{
		"success": len(errors) == 0,
		"results": results,
		"errors":  errors,
		"summary": map[string]interface{}{
			"total_comparisons": len(req.Comparisons),
			"successful":        len(results),
			"failed":            len(errors),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.Error("Failed to encode batch response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Batch business comparison completed",
		zap.Int("total", len(req.Comparisons)),
		zap.Int("successful", len(results)),
		zap.Int("failed", len(errors)))
}

// GetComparisonStats returns statistics about comparison performance
func (h *BusinessComparatorHandler) GetComparisonStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// This would typically come from a metrics/monitoring system
	// For now, return basic stats
	stats := map[string]interface{}{
		"success": true,
		"stats": map[string]interface{}{
			"total_comparisons": 0, // Would be tracked in a real implementation
			"average_score":     0.0,
			"pass_rate":         0.0,
			"partial_rate":      0.0,
			"fail_rate":         0.0,
			"config":            h.comparator.GetConfig(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.logger.Error("Failed to encode stats response", zap.Error(err))
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// createCustomConfig creates a custom configuration from request
func (h *BusinessComparatorHandler) createCustomConfig(configReq *ComparisonConfigRequest) *external.ComparisonConfig {
	// Start with default config
	config := &external.ComparisonConfig{
		MinSimilarityThreshold:   0.8,
		MaxEditDistance:          3,
		PhoneValidationEnabled:   true,
		EmailValidationEnabled:   true,
		AddressValidationEnabled: true,
		MaxDistanceKm:            50.0,
		LocationFuzzyMatch:       true,
		Weights: &external.ComparisonWeights{
			BusinessName:    0.3,
			PhoneNumber:     0.25,
			EmailAddress:    0.2,
			PhysicalAddress: 0.15,
			Website:         0.05,
			Industry:        0.05,
		},
	}

	// Override with provided values
	if configReq.MinSimilarityThreshold != nil {
		config.MinSimilarityThreshold = *configReq.MinSimilarityThreshold
	}
	if configReq.MaxEditDistance != nil {
		config.MaxEditDistance = *configReq.MaxEditDistance
	}
	if configReq.PhoneValidationEnabled != nil {
		config.PhoneValidationEnabled = *configReq.PhoneValidationEnabled
	}
	if configReq.EmailValidationEnabled != nil {
		config.EmailValidationEnabled = *configReq.EmailValidationEnabled
	}
	if configReq.AddressValidationEnabled != nil {
		config.AddressValidationEnabled = *configReq.AddressValidationEnabled
	}
	if configReq.MaxDistanceKm != nil {
		config.MaxDistanceKm = *configReq.MaxDistanceKm
	}
	if configReq.LocationFuzzyMatch != nil {
		config.LocationFuzzyMatch = *configReq.LocationFuzzyMatch
	}
	if configReq.Weights != nil {
		if configReq.Weights.BusinessName != nil {
			config.Weights.BusinessName = *configReq.Weights.BusinessName
		}
		if configReq.Weights.PhoneNumber != nil {
			config.Weights.PhoneNumber = *configReq.Weights.PhoneNumber
		}
		if configReq.Weights.EmailAddress != nil {
			config.Weights.EmailAddress = *configReq.Weights.EmailAddress
		}
		if configReq.Weights.PhysicalAddress != nil {
			config.Weights.PhysicalAddress = *configReq.Weights.PhysicalAddress
		}
		if configReq.Weights.Website != nil {
			config.Weights.Website = *configReq.Weights.Website
		}
		if configReq.Weights.Industry != nil {
			config.Weights.Industry = *configReq.Weights.Industry
		}
	}

	return config
}
