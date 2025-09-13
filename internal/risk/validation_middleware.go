package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// ValidationMiddleware provides validation for risk data operations
type ValidationMiddleware struct {
	validator RiskDataValidator
	logger    *zap.Logger
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware(validator RiskDataValidator, logger *zap.Logger) *ValidationMiddleware {
	return &ValidationMiddleware{
		validator: validator,
		logger:    logger,
	}
}

// ValidateRiskAssessmentMiddleware validates risk assessment data before storage
func (vm *ValidationMiddleware) ValidateRiskAssessmentMiddleware(next func(context.Context, *RiskAssessment) error) func(context.Context, *RiskAssessment) error {
	return func(ctx context.Context, assessment *RiskAssessment) error {
		requestID := ctx.Value("request_id")
		if requestID == nil {
			requestID = "unknown"
		}

		vm.logger.Info("Validating risk assessment before storage",
			zap.String("request_id", requestID.(string)),
			zap.String("assessment_id", assessment.ID),
		)

		// Validate the assessment
		result := vm.validator.ValidateRiskAssessment(ctx, assessment)

		if !result.IsValid {
			vm.logger.Error("Risk assessment validation failed",
				zap.String("request_id", requestID.(string)),
				zap.String("assessment_id", assessment.ID),
				zap.Int("error_count", len(result.Errors)),
				zap.Any("errors", result.Errors),
			)
			return fmt.Errorf("validation failed: %d errors found", len(result.Errors))
		}

		// Log warnings if any
		if len(result.Warnings) > 0 {
			vm.logger.Warn("Risk assessment validation warnings",
				zap.String("request_id", requestID.(string)),
				zap.String("assessment_id", assessment.ID),
				zap.Int("warning_count", len(result.Warnings)),
				zap.Any("warnings", result.Warnings),
			)
		}

		vm.logger.Info("Risk assessment validation passed",
			zap.String("request_id", requestID.(string)),
			zap.String("assessment_id", assessment.ID),
		)

		// Proceed with the next operation
		return next(ctx, assessment)
	}
}

// ValidateRiskAlertMiddleware validates risk alert data before storage
func (vm *ValidationMiddleware) ValidateRiskAlertMiddleware(next func(context.Context, *RiskAlert) error) func(context.Context, *RiskAlert) error {
	return func(ctx context.Context, alert *RiskAlert) error {
		requestID := ctx.Value("request_id")
		if requestID == nil {
			requestID = "unknown"
		}

		vm.logger.Info("Validating risk alert before storage",
			zap.String("request_id", requestID.(string)),
			zap.String("alert_id", alert.ID),
		)

		// Validate the alert
		result := vm.validator.ValidateRiskAlert(ctx, alert)

		if !result.IsValid {
			vm.logger.Error("Risk alert validation failed",
				zap.String("request_id", requestID.(string)),
				zap.String("alert_id", alert.ID),
				zap.Int("error_count", len(result.Errors)),
				zap.Any("errors", result.Errors),
			)
			return fmt.Errorf("validation failed: %d errors found", len(result.Errors))
		}

		// Log warnings if any
		if len(result.Warnings) > 0 {
			vm.logger.Warn("Risk alert validation warnings",
				zap.String("request_id", requestID.(string)),
				zap.String("alert_id", alert.ID),
				zap.Int("warning_count", len(result.Warnings)),
				zap.Any("warnings", result.Warnings),
			)
		}

		vm.logger.Info("Risk alert validation passed",
			zap.String("request_id", requestID.(string)),
			zap.String("alert_id", alert.ID),
		)

		// Proceed with the next operation
		return next(ctx, alert)
	}
}

// ValidateBusinessDataMiddleware validates business data before processing
func (vm *ValidationMiddleware) ValidateBusinessDataMiddleware(next func(context.Context, map[string]interface{}) error) func(context.Context, map[string]interface{}) error {
	return func(ctx context.Context, businessData map[string]interface{}) error {
		requestID := ctx.Value("request_id")
		if requestID == nil {
			requestID = "unknown"
		}

		vm.logger.Info("Validating business data before processing",
			zap.String("request_id", requestID.(string)),
		)

		// Validate the business data
		result := vm.validator.ValidateBusinessData(ctx, businessData)

		if !result.IsValid {
			vm.logger.Error("Business data validation failed",
				zap.String("request_id", requestID.(string)),
				zap.Int("error_count", len(result.Errors)),
				zap.Any("errors", result.Errors),
			)
			return fmt.Errorf("validation failed: %d errors found", len(result.Errors))
		}

		// Log warnings if any
		if len(result.Warnings) > 0 {
			vm.logger.Warn("Business data validation warnings",
				zap.String("request_id", requestID.(string)),
				zap.Int("warning_count", len(result.Warnings)),
				zap.Any("warnings", result.Warnings),
			)
		}

		vm.logger.Info("Business data validation passed",
			zap.String("request_id", requestID.(string)),
		)

		// Proceed with the next operation
		return next(ctx, businessData)
	}
}

// HTTPValidationHandler provides HTTP-level validation for API endpoints
type HTTPValidationHandler struct {
	validator RiskDataValidator
	logger    *zap.Logger
}

// NewHTTPValidationHandler creates a new HTTP validation handler
func NewHTTPValidationHandler(validator RiskDataValidator, logger *zap.Logger) *HTTPValidationHandler {
	return &HTTPValidationHandler{
		validator: validator,
		logger:    logger,
	}
}

// ValidateRiskAssessmentRequest validates risk assessment data from HTTP request
func (h *HTTPValidationHandler) ValidateRiskAssessmentRequest(r *http.Request, assessment *RiskAssessment) *ValidationResult {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	h.logger.Info("Validating risk assessment HTTP request",
		zap.String("request_id", requestID.(string)),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	result := h.validator.ValidateRiskAssessment(ctx, assessment)

	if !result.IsValid {
		h.logger.Error("Risk assessment HTTP request validation failed",
			zap.String("request_id", requestID.(string)),
			zap.Int("error_count", len(result.Errors)),
			zap.Any("errors", result.Errors),
		)
	} else {
		h.logger.Info("Risk assessment HTTP request validation passed",
			zap.String("request_id", requestID.(string)),
		)
	}

	return result
}

// ValidateBusinessDataRequest validates business data from HTTP request
func (h *HTTPValidationHandler) ValidateBusinessDataRequest(r *http.Request, businessData map[string]interface{}) *ValidationResult {
	ctx := r.Context()
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	h.logger.Info("Validating business data HTTP request",
		zap.String("request_id", requestID.(string)),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
	)

	result := h.validator.ValidateBusinessData(ctx, businessData)

	if !result.IsValid {
		h.logger.Error("Business data HTTP request validation failed",
			zap.String("request_id", requestID.(string)),
			zap.Int("error_count", len(result.Errors)),
			zap.Any("errors", result.Errors),
		)
	} else {
		h.logger.Info("Business data HTTP request validation passed",
			zap.String("request_id", requestID.(string)),
		)
	}

	return result
}

// WriteValidationErrorResponse writes validation error response to HTTP response writer
func (h *HTTPValidationHandler) WriteValidationErrorResponse(w http.ResponseWriter, result *ValidationResult) {
	requestID := "unknown"
	if ctx := w.(*http.Request).Context(); ctx != nil {
		if id := ctx.Value("request_id"); id != nil {
			requestID = id.(string)
		}
	}

	h.logger.Error("Writing validation error response",
		zap.String("request_id", requestID),
		zap.Int("error_count", len(result.Errors)),
		zap.Int("warning_count", len(result.Warnings)),
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	// Create error response
	errorResponse := map[string]interface{}{
		"error":   "validation_failed",
		"message": "Request data validation failed",
		"details": map[string]interface{}{
			"errors":   result.Errors,
			"warnings": result.Warnings,
		},
	}

	// Write JSON response
	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
		h.logger.Error("Failed to write validation error response",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
	}
}

// ValidationConfig provides configuration for validation behavior
type ValidationConfig struct {
	// StrictMode enables strict validation (treats warnings as errors)
	StrictMode bool

	// MaxWarnings is the maximum number of warnings allowed before failing
	MaxWarnings int

	// EnableContentValidation enables content validation for suspicious patterns
	EnableContentValidation bool

	// EnableScoreValidation enables risk score validation
	EnableScoreValidation bool
}

// DefaultValidationConfig returns the default validation configuration
func DefaultValidationConfig() *ValidationConfig {
	return &ValidationConfig{
		StrictMode:              false,
		MaxWarnings:             10,
		EnableContentValidation: true,
		EnableScoreValidation:   true,
	}
}

// ValidationServiceWithConfig provides validation with configurable behavior
type ValidationServiceWithConfig struct {
	*RiskValidationService
	config *ValidationConfig
}

// NewValidationServiceWithConfig creates a validation service with configuration
func NewValidationServiceWithConfig(logger *zap.Logger, config *ValidationConfig) *ValidationServiceWithConfig {
	return &ValidationServiceWithConfig{
		RiskValidationService: NewRiskValidationService(logger),
		config:                config,
	}
}

// ValidateRiskAssessmentWithConfig validates risk assessment with configuration
func (v *ValidationServiceWithConfig) ValidateRiskAssessmentWithConfig(ctx context.Context, assessment *RiskAssessment) *ValidationResult {
	result := v.RiskValidationService.ValidateRiskAssessment(ctx, assessment)

	// Apply configuration rules
	if v.config.StrictMode && len(result.Warnings) > 0 {
		// Convert warnings to errors in strict mode
		result.Errors = append(result.Errors, result.Warnings...)
		result.Warnings = []ValidationError{}
		result.IsValid = false
	}

	if len(result.Warnings) > v.config.MaxWarnings {
		// Too many warnings, fail validation
		result.Errors = append(result.Errors, ValidationError{
			Field:   "warnings",
			Message: fmt.Sprintf("too many warnings: %d (max: %d)", len(result.Warnings), v.config.MaxWarnings),
			Code:    "TOO_MANY_WARNINGS",
		})
		result.IsValid = false
	}

	return result
}
