package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/pkg/validators"
)

// AccuracyValidationHandler handles accuracy validation requests
type AccuracyValidationHandler struct {
	accuracyValidationEngine *classification.AccuracyValidationEngine
	logger                   *observability.Logger
	metrics                  *observability.Metrics
}

// AccuracyValidationRequest represents a request to validate classification accuracy
type AccuracyValidationRequest struct {
	BusinessName   string                                `json:"business_name" validate:"required"`
	Classification classification.IndustryClassification `json:"classification" validate:"required"`
	KnownData      *classification.KnownClassification   `json:"known_data,omitempty"`
	BenchmarkData  *classification.IndustryBenchmark     `json:"benchmark_data,omitempty"`
}

// AccuracyValidationResponse represents the response from accuracy validation
type AccuracyValidationResponse struct {
	Success          bool                             `json:"success"`
	ValidationResult *classification.ValidationResult `json:"validation_result"`
	AccuracyMetrics  map[string]interface{}           `json:"accuracy_metrics"`
	ProcessingTime   time.Duration                    `json:"processing_time"`
	Timestamp        time.Time                        `json:"timestamp"`
}

// AddKnownClassificationRequest represents a request to add known classification data
type AddKnownClassificationRequest struct {
	KnownClassifications []classification.KnownClassification `json:"known_classifications" validate:"required"`
}

// AddBenchmarkRequest represents a request to add industry benchmark data
type AddBenchmarkRequest struct {
	Benchmarks []classification.IndustryBenchmark `json:"benchmarks" validate:"required"`
}

// NewAccuracyValidationHandler creates a new accuracy validation handler
func NewAccuracyValidationHandler(accuracyValidationEngine *classification.AccuracyValidationEngine, logger *observability.Logger, metrics *observability.Metrics) *AccuracyValidationHandler {
	return &AccuracyValidationHandler{
		accuracyValidationEngine: accuracyValidationEngine,
		logger:                   logger,
		metrics:                  metrics,
	}
}

// HandleValidateClassification handles classification accuracy validation requests
func (h *AccuracyValidationHandler) HandleValidateClassification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request
	var req AccuracyValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	// Validate request
	if err := validators.Validate(req); err != nil {
		h.handleError(w, "Validation failed", http.StatusBadRequest, err)
		return
	}

	// Log validation request
	h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("accuracy_validation_request_received", map[string]interface{}{
		"business_name":    req.BusinessName,
		"industry_code":    req.Classification.IndustryCode,
		"confidence_score": req.Classification.ConfidenceScore,
	})

	// Add known classification if provided
	if req.KnownData != nil {
		h.accuracyValidationEngine.AddKnownClassification(ctx, *req.KnownData)
	}

	// Add benchmark data if provided
	if req.BenchmarkData != nil {
		h.accuracyValidationEngine.AddIndustryBenchmark(ctx, *req.BenchmarkData)
	}

	// Perform accuracy validation
	validationResult, err := h.accuracyValidationEngine.ValidateClassification(ctx, req.Classification)
	if err != nil {
		h.handleError(w, "Validation failed", http.StatusInternalServerError, err)
		return
	}

	// Get accuracy metrics
	accuracyMetrics, err := h.accuracyValidationEngine.GetAccuracyMetrics(ctx)
	if err != nil {
		h.handleError(w, "Failed to get accuracy metrics", http.StatusInternalServerError, err)
		return
	}

	// Create response
	response := &AccuracyValidationResponse{
		Success:          true,
		ValidationResult: validationResult,
		AccuracyMetrics:  accuracyMetrics,
		ProcessingTime:   time.Since(start),
		Timestamp:        time.Now(),
	}

	// Log successful completion
	h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("accuracy_validation_request_completed", map[string]interface{}{
		"business_name":      req.BusinessName,
		"accuracy_score":     validationResult.AccuracyScore,
		"is_accurate":        validationResult.IsAccurate,
		"validation_method":  validationResult.ValidationMethod,
		"processing_time_ms": time.Since(start).Milliseconds(),
		"http_status":        http.StatusOK,
	})

	// Record metrics
	h.metrics.RecordBusinessClassification("accuracy_validation_api_success", 1.0)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("accuracy_validation_response_encoding_error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// HandleAddKnownClassification handles requests to add known classification data
func (h *AccuracyValidationHandler) HandleAddKnownClassification(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request
	var req AddKnownClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	// Validate request
	if err := validators.Validate(req); err != nil {
		h.handleError(w, "Validation failed", http.StatusBadRequest, err)
		return
	}

	// Add known classifications
	for _, known := range req.KnownClassifications {
		h.accuracyValidationEngine.AddKnownClassification(ctx, known)
	}

	// Log addition
	h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("known_classifications_added", map[string]interface{}{
		"count":              len(req.KnownClassifications),
		"processing_time_ms": time.Since(start).Milliseconds(),
	})

	// Record metrics
	h.metrics.RecordBusinessClassification("known_classifications_added", float64(len(req.KnownClassifications)))

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success":         true,
		"message":         "Known classifications added successfully",
		"count_added":     len(req.KnownClassifications),
		"processing_time": time.Since(start),
		"timestamp":       time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("known_classifications_response_encoding_error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// HandleAddBenchmark handles requests to add industry benchmark data
func (h *AccuracyValidationHandler) HandleAddBenchmark(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Parse request
	var req AddBenchmarkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.handleError(w, "Invalid request format", http.StatusBadRequest, err)
		return
	}

	// Validate request
	if err := validators.Validate(req); err != nil {
		h.handleError(w, "Validation failed", http.StatusBadRequest, err)
		return
	}

	// Add benchmarks
	for _, benchmark := range req.Benchmarks {
		h.accuracyValidationEngine.AddIndustryBenchmark(ctx, benchmark)
	}

	// Log addition
	h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("benchmarks_added", map[string]interface{}{
		"count":              len(req.Benchmarks),
		"processing_time_ms": time.Since(start).Milliseconds(),
	})

	// Record metrics
	h.metrics.RecordBusinessClassification("benchmarks_added", float64(len(req.Benchmarks)))

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success":         true,
		"message":         "Industry benchmarks added successfully",
		"count_added":     len(req.Benchmarks),
		"processing_time": time.Since(start),
		"timestamp":       time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("benchmarks_response_encoding_error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// HandleGetAccuracyMetrics handles requests to get accuracy metrics
func (h *AccuracyValidationHandler) HandleGetAccuracyMetrics(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	ctx := r.Context()

	// Get accuracy metrics
	metrics, err := h.accuracyValidationEngine.GetAccuracyMetrics(ctx)
	if err != nil {
		h.handleError(w, "Failed to get accuracy metrics", http.StatusInternalServerError, err)
		return
	}

	// Log request
	h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("accuracy_metrics_requested", map[string]interface{}{
		"processing_time_ms": time.Since(start).Milliseconds(),
	})

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success":         true,
		"metrics":         metrics,
		"processing_time": time.Since(start),
		"timestamp":       time.Now(),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("accuracy_metrics_response_encoding_error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}

// handleError handles error responses
func (h *AccuracyValidationHandler) handleError(w http.ResponseWriter, message string, statusCode int, err error) {
	h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("accuracy_validation_error", map[string]interface{}{
		"error":       err.Error(),
		"status_code": statusCode,
		"message":     message,
	})

	// Record error metrics
	h.metrics.RecordBusinessClassification("accuracy_validation_api_error", 1.0)

	// Send error response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResponse := map[string]interface{}{
		"success": false,
		"error":   message,
		"details": err.Error(),
	}

	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		h.logger.WithComponent("accuracy_validation_handler").LogBusinessEvent("error_response_encoding_error", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
