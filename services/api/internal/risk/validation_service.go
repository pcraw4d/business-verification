package risk

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
}

// ValidationResult represents the result of a validation operation
type ValidationResult struct {
	IsValid  bool              `json:"is_valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
}

// RiskDataValidator defines the interface for risk data validation
type RiskDataValidator interface {
	ValidateRiskAssessment(ctx context.Context, assessment *RiskAssessment) *ValidationResult
	ValidateRiskAlert(ctx context.Context, alert *RiskAlert) *ValidationResult
	ValidateRiskTrend(ctx context.Context, trend *RiskTrend) *ValidationResult
	ValidateBusinessData(ctx context.Context, businessData map[string]interface{}) *ValidationResult
	ValidateRiskScore(ctx context.Context, score float64, level RiskLevel) *ValidationResult
}

// RiskValidationService implements comprehensive risk data validation
type RiskValidationService struct {
	logger *zap.Logger
}

// NewRiskValidationService creates a new risk validation service
func NewRiskValidationService(logger *zap.Logger) *RiskValidationService {
	return &RiskValidationService{
		logger: logger,
	}
}

// ValidateRiskAssessment validates a complete risk assessment
func (v *RiskValidationService) ValidateRiskAssessment(ctx context.Context, assessment *RiskAssessment) *ValidationResult {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	v.logger.Info("Validating risk assessment",
		zap.String("request_id", requestID.(string)),
		zap.String("assessment_id", assessment.ID),
		zap.String("business_id", assessment.BusinessID),
	)

	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Validate required fields
	if assessment.ID == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "id",
			Message: "assessment ID is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	if assessment.BusinessID == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "business_id",
			Message: "business ID is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	if assessment.BusinessName == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "business_name",
			Message: "business name is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	// Validate risk score
	scoreResult := v.ValidateRiskScore(ctx, assessment.OverallScore, assessment.OverallLevel)
	if !scoreResult.IsValid {
		result.Errors = append(result.Errors, scoreResult.Errors...)
		result.IsValid = false
	}
	result.Warnings = append(result.Warnings, scoreResult.Warnings...)

	// Validate timestamps
	if assessment.AssessedAt.IsZero() {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "assessed_at",
			Message: "assessment timestamp is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	if assessment.ValidUntil.IsZero() {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "valid_until",
			Message: "valid until timestamp is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	// Validate timestamp logic
	if !assessment.AssessedAt.IsZero() && !assessment.ValidUntil.IsZero() {
		if assessment.ValidUntil.Before(assessment.AssessedAt) {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "valid_until",
				Message: "valid until date must be after assessment date",
				Code:    "INVALID_DATE_RANGE",
			})
			result.IsValid = false
		}
	}

	// Validate assessment method
	if assessment.AssessmentMethod == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "assessment_method",
			Message: "assessment method is not specified",
			Code:    "MISSING_METHOD",
		})
	}

	// Validate source
	if assessment.Source == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "source",
			Message: "assessment source is not specified",
			Code:    "MISSING_SOURCE",
		})
	}

	// Validate category scores
	if len(assessment.CategoryScores) == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "category_scores",
			Message: "no category scores provided",
			Code:    "EMPTY_CATEGORIES",
		})
	} else {
		for category, score := range assessment.CategoryScores {
			if score == nil {
				result.Errors = append(result.Errors, ValidationError{
					Field:   fmt.Sprintf("category_scores.%s", category),
					Message: "category score cannot be null",
					Code:    "NULL_SCORE",
				})
				result.IsValid = false
			}
		}
	}

	// Validate factor scores
	if len(assessment.FactorScores) == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "factor_scores",
			Message: "no factor scores provided",
			Code:    "EMPTY_FACTORS",
		})
	}

	// Validate recommendations
	if len(assessment.Recommendations) == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "recommendations",
			Message: "no recommendations provided",
			Code:    "EMPTY_RECOMMENDATIONS",
		})
	}

	// Validate alerts
	if len(assessment.Alerts) == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "alerts",
			Message: "no alerts provided",
			Code:    "EMPTY_ALERTS",
		})
	}

	// Validate metadata
	if assessment.Metadata != nil {
		for key, value := range assessment.Metadata {
			if value == nil {
				result.Warnings = append(result.Warnings, ValidationError{
					Field:   fmt.Sprintf("metadata.%s", key),
					Message: "metadata value is null",
					Code:    "NULL_METADATA",
				})
			}
		}
	}

	v.logger.Info("Risk assessment validation completed",
		zap.String("request_id", requestID.(string)),
		zap.String("assessment_id", assessment.ID),
		zap.Bool("is_valid", result.IsValid),
		zap.Int("error_count", len(result.Errors)),
		zap.Int("warning_count", len(result.Warnings)),
	)

	return result
}

// ValidateRiskAlert validates a risk alert
func (v *RiskValidationService) ValidateRiskAlert(ctx context.Context, alert *RiskAlert) *ValidationResult {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	v.logger.Info("Validating risk alert",
		zap.String("request_id", requestID.(string)),
		zap.String("alert_id", alert.ID),
		zap.String("business_id", alert.BusinessID),
	)

	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Validate required fields
	if alert.ID == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "id",
			Message: "alert ID is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	if alert.BusinessID == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "business_id",
			Message: "business ID is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	if alert.RiskFactor == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "risk_factor",
			Message: "risk factor is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	if alert.Message == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "message",
			Message: "alert message is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	// Validate risk level
	validLevels := []RiskLevel{RiskLevelLow, RiskLevelMedium, RiskLevelHigh, RiskLevelCritical}
	validLevel := false
	for _, level := range validLevels {
		if alert.Level == level {
			validLevel = true
			break
		}
	}
	if !validLevel {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "level",
			Message: fmt.Sprintf("invalid risk level: %s", alert.Level),
			Code:    "INVALID_LEVEL",
		})
		result.IsValid = false
	}

	// Validate score range
	if alert.Score < 0 || alert.Score > 1 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "score",
			Message: "score must be between 0 and 1",
			Code:    "INVALID_SCORE_RANGE",
		})
		result.IsValid = false
	}

	// Validate threshold range
	if alert.Threshold < 0 || alert.Threshold > 1 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "threshold",
			Message: "threshold must be between 0 and 1",
			Code:    "INVALID_THRESHOLD_RANGE",
		})
		result.IsValid = false
	}

	// Validate timestamp
	if alert.TriggeredAt.IsZero() {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "triggered_at",
			Message: "triggered timestamp is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	// Validate acknowledgment logic
	if alert.Acknowledged && alert.AcknowledgedAt == nil {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "acknowledged_at",
			Message: "acknowledgment timestamp is required when acknowledged is true",
			Code:    "MISSING_ACKNOWLEDGMENT_TIME",
		})
		result.IsValid = false
	}

	if !alert.Acknowledged && alert.AcknowledgedAt != nil {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "acknowledged_at",
			Message: "acknowledgment timestamp should be null when acknowledged is false",
			Code:    "INCONSISTENT_ACKNOWLEDGMENT",
		})
	}

	v.logger.Info("Risk alert validation completed",
		zap.String("request_id", requestID.(string)),
		zap.String("alert_id", alert.ID),
		zap.Bool("is_valid", result.IsValid),
		zap.Int("error_count", len(result.Errors)),
		zap.Int("warning_count", len(result.Warnings)),
	)

	return result
}

// ValidateRiskTrend validates a risk trend
func (v *RiskValidationService) ValidateRiskTrend(ctx context.Context, trend *RiskTrend) *ValidationResult {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	v.logger.Info("Validating risk trend",
		zap.String("request_id", requestID.(string)),
		zap.String("business_id", trend.BusinessID),
	)

	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Validate required fields
	if trend.BusinessID == "" {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "business_id",
			Message: "business ID is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	// Validate trend direction
	validDirections := []string{"improving", "stable", "declining"}
	validDirection := false
	for _, direction := range validDirections {
		if trend.Direction == direction {
			validDirection = true
			break
		}
	}
	if !validDirection {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "direction",
			Message: fmt.Sprintf("invalid trend direction: %s", trend.Direction),
			Code:    "INVALID_DIRECTION",
		})
		result.IsValid = false
	}

	// Validate confidence score
	if trend.Confidence < 0 || trend.Confidence > 1 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "confidence",
			Message: "confidence must be between 0 and 1",
			Code:    "INVALID_CONFIDENCE_RANGE",
		})
		result.IsValid = false
	}

	// Validate period
	if trend.Period <= 0 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "period",
			Message: "period must be greater than 0",
			Code:    "INVALID_PERIOD",
		})
		result.IsValid = false
	}

	// Validate timestamp
	if trend.AnalyzedAt.IsZero() {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "analyzed_at",
			Message: "analysis timestamp is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	v.logger.Info("Risk trend validation completed",
		zap.String("request_id", requestID.(string)),
		zap.String("business_id", trend.BusinessID),
		zap.Bool("is_valid", result.IsValid),
		zap.Int("error_count", len(result.Errors)),
		zap.Int("warning_count", len(result.Warnings)),
	)

	return result
}

// ValidateBusinessData validates business data input
func (v *RiskValidationService) ValidateBusinessData(ctx context.Context, businessData map[string]interface{}) *ValidationResult {
	requestID := ctx.Value("request_id")
	if requestID == nil {
		requestID = "unknown"
	}

	v.logger.Info("Validating business data",
		zap.String("request_id", requestID.(string)),
	)

	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Validate business name
	if name, exists := businessData["name"]; exists {
		if nameStr, ok := name.(string); ok {
			if strings.TrimSpace(nameStr) == "" {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "name",
					Message: "business name cannot be empty",
					Code:    "EMPTY_NAME",
				})
				result.IsValid = false
			} else if len(nameStr) > 255 {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "name",
					Message: "business name exceeds maximum length of 255 characters",
					Code:    "NAME_TOO_LONG",
				})
				result.IsValid = false
			}
		} else {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "name",
				Message: "business name must be a string",
				Code:    "INVALID_TYPE",
			})
			result.IsValid = false
		}
	} else {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "name",
			Message: "business name is required",
			Code:    "REQUIRED_FIELD",
		})
		result.IsValid = false
	}

	// Validate business address
	if address, exists := businessData["address"]; exists {
		if addressStr, ok := address.(string); ok {
			if strings.TrimSpace(addressStr) == "" {
				result.Warnings = append(result.Warnings, ValidationError{
					Field:   "address",
					Message: "business address is empty",
					Code:    "EMPTY_ADDRESS",
				})
			} else if len(addressStr) > 500 {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "address",
					Message: "business address exceeds maximum length of 500 characters",
					Code:    "ADDRESS_TOO_LONG",
				})
				result.IsValid = false
			}
		} else {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "address",
				Message: "business address must be a string",
				Code:    "INVALID_TYPE",
			})
			result.IsValid = false
		}
	}

	// Validate email if provided
	if email, exists := businessData["email"]; exists {
		if emailStr, ok := email.(string); ok {
			if emailStr != "" && !v.isValidEmail(emailStr) {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "email",
					Message: "invalid email format",
					Code:    "INVALID_EMAIL",
				})
				result.IsValid = false
			}
		} else {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "email",
				Message: "email must be a string",
				Code:    "INVALID_TYPE",
			})
			result.IsValid = false
		}
	}

	// Validate phone if provided
	if phone, exists := businessData["phone"]; exists {
		if phoneStr, ok := phone.(string); ok {
			if phoneStr != "" && !v.isValidPhone(phoneStr) {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "phone",
					Message: "invalid phone format",
					Code:    "INVALID_PHONE",
				})
				result.IsValid = false
			}
		} else {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "phone",
				Message: "phone must be a string",
				Code:    "INVALID_TYPE",
			})
			result.IsValid = false
		}
	}

	// Validate website if provided
	if website, exists := businessData["website"]; exists {
		if websiteStr, ok := website.(string); ok {
			if websiteStr != "" && !v.isValidURL(websiteStr) {
				result.Errors = append(result.Errors, ValidationError{
					Field:   "website",
					Message: "invalid website URL format",
					Code:    "INVALID_URL",
				})
				result.IsValid = false
			}
		} else {
			result.Errors = append(result.Errors, ValidationError{
				Field:   "website",
				Message: "website must be a string",
				Code:    "INVALID_TYPE",
			})
			result.IsValid = false
		}
	}

	// Check for potentially malicious content
	for key, value := range businessData {
		if strValue, ok := value.(string); ok {
			if v.containsSuspiciousContent(strValue) {
				result.Warnings = append(result.Warnings, ValidationError{
					Field:   key,
					Message: "content may contain suspicious patterns",
					Code:    "SUSPICIOUS_CONTENT",
				})
			}
		}
	}

	v.logger.Info("Business data validation completed",
		zap.String("request_id", requestID.(string)),
		zap.Bool("is_valid", result.IsValid),
		zap.Int("error_count", len(result.Errors)),
		zap.Int("warning_count", len(result.Warnings)),
	)

	return result
}

// ValidateRiskScore validates risk score and level consistency
func (v *RiskValidationService) ValidateRiskScore(ctx context.Context, score float64, level RiskLevel) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
	}

	// Validate score range
	if score < 0 || score > 1 {
		result.Errors = append(result.Errors, ValidationError{
			Field:   "overall_score",
			Message: "risk score must be between 0 and 1",
			Code:    "INVALID_SCORE_RANGE",
		})
		result.IsValid = false
		return result
	}

	// Validate score and level consistency
	expectedLevel := v.calculateExpectedRiskLevel(score)
	if level != expectedLevel {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "overall_level",
			Message: fmt.Sprintf("risk level '%s' may not match score %.3f (expected: %s)", level, score, expectedLevel),
			Code:    "LEVEL_SCORE_MISMATCH",
		})
	}

	// Check for extreme values
	if score == 0 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "overall_score",
			Message: "risk score is exactly 0, which may indicate missing data",
			Code:    "ZERO_SCORE",
		})
	}

	if score == 1 {
		result.Warnings = append(result.Warnings, ValidationError{
			Field:   "overall_score",
			Message: "risk score is exactly 1, which may indicate maximum risk",
			Code:    "MAXIMUM_SCORE",
		})
	}

	return result
}

// Helper methods

// calculateExpectedRiskLevel calculates the expected risk level based on score
func (v *RiskValidationService) calculateExpectedRiskLevel(score float64) RiskLevel {
	switch {
	case score >= 0.8:
		return RiskLevelCritical
	case score >= 0.6:
		return RiskLevelHigh
	case score >= 0.3:
		return RiskLevelMedium
	default:
		return RiskLevelLow
	}
}

// isValidEmail validates email format
func (v *RiskValidationService) isValidEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(pattern, email)
	return matched && len(email) <= 254
}

// isValidPhone validates phone format (E.164)
func (v *RiskValidationService) isValidPhone(phone string) bool {
	pattern := `^\+[1-9]\d{1,14}$`
	matched, _ := regexp.MatchString(pattern, phone)
	return matched
}

// isValidURL validates URL format
func (v *RiskValidationService) isValidURL(url string) bool {
	pattern := `^https?://[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}(/.*)?$`
	matched, _ := regexp.MatchString(pattern, url)
	return matched
}

// containsSuspiciousContent checks for potentially malicious content
func (v *RiskValidationService) containsSuspiciousContent(content string) bool {
	// Basic patterns for suspicious content
	suspiciousPatterns := []string{
		`(?i)(script|javascript|vbscript|onload|onerror|onclick)`,
		`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`,
		`['\";]`,
		`<[^>]*>`, // HTML tags
	}

	for _, pattern := range suspiciousPatterns {
		matched, _ := regexp.MatchString(pattern, content)
		if matched {
			return true
		}
	}

	return false
}
