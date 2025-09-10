package compatibility

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/pkg/validators"
)

// BackwardCompatibilityLayer handles backward compatibility for API versions
type BackwardCompatibilityLayer struct {
	featureFlagManager *config.FeatureFlagManager
	logger             Logger
	metrics            Metrics
	validator          *validators.Validator
}

// Logger interface for logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})
}

// Metrics interface for metrics collection
type Metrics interface {
	IncrementCounter(name string, tags map[string]string)
	RecordHistogram(name string, value float64, tags map[string]string)
}

// ClassificationRequest represents a classification request
type ClassificationRequest struct {
	BusinessName    string                 `json:"business_name"`
	BusinessAddress string                 `json:"business_address"`
	Industry        string                 `json:"industry,omitempty"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// ClassificationResponse represents a classification response
type ClassificationResponse struct {
	ClassificationCodes []ClassificationCode   `json:"classification_codes"`
	Confidence          float64                `json:"confidence"`
	Version             string                 `json:"version"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// ClassificationCode represents a classification code
type ClassificationCode struct {
	Type        string  `json:"type"`
	Code        string  `json:"code"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
}

// NewBackwardCompatibilityLayer creates a new backward compatibility layer
func NewBackwardCompatibilityLayer(
	featureFlagManager *config.FeatureFlagManager,
	logger Logger,
	metrics Metrics,
	validator *validators.Validator,
) *BackwardCompatibilityLayer {
	return &BackwardCompatibilityLayer{
		featureFlagManager: featureFlagManager,
		logger:             logger,
		metrics:            metrics,
		validator:          validator,
	}
}

// GetAPIVersion returns the current API version
func (bcl *BackwardCompatibilityLayer) GetAPIVersion() string {
	return "v1.0.0"
}

// ProcessRequest processes a classification request with backward compatibility
func (bcl *BackwardCompatibilityLayer) ProcessRequest(
	ctx context.Context,
	req *ClassificationRequest,
) (*ClassificationResponse, error) {
	bcl.logger.Info("Processing classification request", "business_name", req.BusinessName)

	// Validate request
	if err := bcl.validateRequest(req); err != nil {
		bcl.metrics.IncrementCounter("classification.request.validation_failed", map[string]string{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check feature flags for backward compatibility
	if bcl.featureFlagManager.IsEnabled("legacy_classification") {
		return bcl.processLegacyRequest(ctx, req)
	}

	// Process with current version
	return bcl.processCurrentRequest(ctx, req)
}

// validateRequest validates the classification request
func (bcl *BackwardCompatibilityLayer) validateRequest(req *ClassificationRequest) error {
	if req.BusinessName == "" {
		return fmt.Errorf("business_name is required")
	}
	if req.BusinessAddress == "" {
		return fmt.Errorf("business_address is required")
	}
	return nil
}

// processLegacyRequest processes a request using legacy classification logic
func (bcl *BackwardCompatibilityLayer) processLegacyRequest(
	ctx context.Context,
	req *ClassificationRequest,
) (*ClassificationResponse, error) {
	bcl.logger.Debug("Processing legacy classification request")

	// Simulate legacy classification logic
	codes := []ClassificationCode{
		{
			Type:        "MCC",
			Code:        "5999",
			Description: "Miscellaneous and Specialty Retail Stores",
			Confidence:  0.85,
		},
		{
			Type:        "SIC",
			Code:        "5999",
			Description: "Retail Stores, Not Elsewhere Classified",
			Confidence:  0.80,
		},
		{
			Type:        "NAICS",
			Code:        "453990",
			Description: "All Other Miscellaneous Store Retailers",
			Confidence:  0.75,
		},
	}

	return &ClassificationResponse{
		ClassificationCodes: codes,
		Confidence:          0.80,
		Version:             "legacy",
		Metadata: map[string]interface{}{
			"legacy_mode":  true,
			"processed_at": "2025-01-01T00:00:00Z",
		},
	}, nil
}

// processCurrentRequest processes a request using current classification logic
func (bcl *BackwardCompatibilityLayer) processCurrentRequest(
	ctx context.Context,
	req *ClassificationRequest,
) (*ClassificationResponse, error) {
	bcl.logger.Debug("Processing current classification request")

	// Simulate current classification logic with enhanced features
	codes := []ClassificationCode{
		{
			Type:        "MCC",
			Code:        "5999",
			Description: "Miscellaneous and Specialty Retail Stores",
			Confidence:  0.92,
		},
		{
			Type:        "SIC",
			Code:        "5999",
			Description: "Retail Stores, Not Elsewhere Classified",
			Confidence:  0.88,
		},
		{
			Type:        "NAICS",
			Code:        "453990",
			Description: "All Other Miscellaneous Store Retailers",
			Confidence:  0.85,
		},
	}

	bcl.metrics.IncrementCounter("classification.request.processed", map[string]string{
		"version": "current",
	})
	bcl.metrics.RecordHistogram("classification.confidence", 0.88, map[string]string{
		"version": "current",
	})

	return &ClassificationResponse{
		ClassificationCodes: codes,
		Confidence:          0.88,
		Version:             "v1.0.0",
		Metadata: map[string]interface{}{
			"enhanced_features": true,
			"processed_at":      "2025-01-01T00:00:00Z",
		},
	}, nil
}

// HandleHTTPRequest handles HTTP requests with backward compatibility
func (bcl *BackwardCompatibilityLayer) HandleHTTPRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req ClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	resp, err := bcl.ProcessRequest(ctx, &req)
	if err != nil {
		bcl.logger.Error("Failed to process request", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// IsLegacyModeEnabled checks if legacy mode is enabled
func (bcl *BackwardCompatibilityLayer) IsLegacyModeEnabled() bool {
	return bcl.featureFlagManager.IsEnabled("legacy_classification")
}

// GetSupportedVersions returns the list of supported API versions
func (bcl *BackwardCompatibilityLayer) GetSupportedVersions() []string {
	return []string{"v1.0.0", "legacy"}
}
