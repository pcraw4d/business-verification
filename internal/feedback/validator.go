package feedback

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// FeedbackValidator implements the FeedbackValidatorInterface
type FeedbackValidator struct {
	securityValidator SecurityValidator
}

// NewFeedbackValidator creates a new feedback validator
func NewFeedbackValidator(securityValidator SecurityValidator) *FeedbackValidator {
	return &FeedbackValidator{
		securityValidator: securityValidator,
	}
}

// ValidateUserFeedback validates user feedback data
func (v *FeedbackValidator) ValidateUserFeedback(ctx context.Context, feedback UserFeedback) error {
	// Validate required fields
	if feedback.ID == "" {
		return fmt.Errorf("feedback ID is required")
	}

	if feedback.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if feedback.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}

	if feedback.OriginalClassificationID == "" {
		return fmt.Errorf("original classification ID is required")
	}

	// Validate feedback type
	if !isValidFeedbackType(feedback.FeedbackType) {
		return fmt.Errorf("invalid feedback type: %s", feedback.FeedbackType)
	}

	// Validate status
	if !isValidFeedbackStatus(feedback.Status) {
		return fmt.Errorf("invalid feedback status: %s", feedback.Status)
	}

	// Validate confidence score range
	if feedback.ConfidenceScore < 0.0 || feedback.ConfidenceScore > 1.0 {
		return fmt.Errorf("confidence score must be between 0.0 and 1.0, got: %f", feedback.ConfidenceScore)
	}

	// Validate business name format
	if err := v.validateBusinessName(feedback.BusinessName); err != nil {
		return fmt.Errorf("invalid business name: %w", err)
	}

	// Validate feedback text length
	if len(feedback.FeedbackText) > 5000 {
		return fmt.Errorf("feedback text exceeds maximum length of 5000 characters")
	}

	// Validate processing time
	if feedback.ProcessingTimeMs < 0 {
		return fmt.Errorf("processing time cannot be negative")
	}

	// Validate timestamps
	if feedback.ProcessedAt != nil && feedback.ProcessedAt.Before(feedback.CreatedAt) {
		return fmt.Errorf("processed at time cannot be before created at time")
	}

	// Security validation
	if v.securityValidator != nil {
		if err := v.securityValidator.ValidateFeedbackContent(ctx, feedback.FeedbackValue); err != nil {
			return fmt.Errorf("security validation failed: %w", err)
		}
	}

	return nil
}

// ValidateMLModelFeedback validates ML model feedback data
func (v *FeedbackValidator) ValidateMLModelFeedback(ctx context.Context, feedback MLModelFeedback) error {
	// Validate required fields
	if feedback.ID == "" {
		return fmt.Errorf("feedback ID is required")
	}

	if feedback.ModelVersionID == "" {
		return fmt.Errorf("model version ID is required")
	}

	if feedback.ModelType == "" {
		return fmt.Errorf("model type is required")
	}

	if feedback.ClassificationMethod == "" {
		return fmt.Errorf("classification method is required")
	}

	if feedback.PredictionID == "" {
		return fmt.Errorf("prediction ID is required")
	}

	// Validate classification method
	if !isValidClassificationMethod(feedback.ClassificationMethod) {
		return fmt.Errorf("invalid classification method: %s", feedback.ClassificationMethod)
	}

	// Validate status
	if !isValidFeedbackStatus(feedback.Status) {
		return fmt.Errorf("invalid feedback status: %s", feedback.Status)
	}

	// Validate accuracy score range
	if feedback.AccuracyScore < 0.0 || feedback.AccuracyScore > 1.0 {
		return fmt.Errorf("accuracy score must be between 0.0 and 1.0, got: %f", feedback.AccuracyScore)
	}

	// Validate confidence score range
	if feedback.ConfidenceScore < 0.0 || feedback.ConfidenceScore > 1.0 {
		return fmt.Errorf("confidence score must be between 0.0 and 1.0, got: %f", feedback.ConfidenceScore)
	}

	// Validate processing time
	if feedback.ProcessingTimeMs < 0 {
		return fmt.Errorf("processing time cannot be negative")
	}

	// Validate timestamps
	if feedback.ProcessedAt != nil && feedback.ProcessedAt.Before(feedback.CreatedAt) {
		return fmt.Errorf("processed at time cannot be before created at time")
	}

	// Validate actual and predicted results are not empty
	if feedback.ActualResult == nil || len(feedback.ActualResult) == 0 {
		return fmt.Errorf("actual result is required and cannot be empty")
	}

	if feedback.PredictedResult == nil || len(feedback.PredictedResult) == 0 {
		return fmt.Errorf("predicted result is required and cannot be empty")
	}

	// Security validation
	if v.securityValidator != nil {
		if err := v.securityValidator.ValidateFeedbackContent(ctx, feedback.ActualResult); err != nil {
			return fmt.Errorf("security validation failed for actual result: %w", err)
		}

		if err := v.securityValidator.ValidateFeedbackContent(ctx, feedback.PredictedResult); err != nil {
			return fmt.Errorf("security validation failed for predicted result: %w", err)
		}
	}

	return nil
}

// ValidateSecurityValidationFeedback validates security validation feedback data
func (v *FeedbackValidator) ValidateSecurityValidationFeedback(ctx context.Context, feedback SecurityValidationFeedback) error {
	// Validate required fields
	if feedback.ID == "" {
		return fmt.Errorf("feedback ID is required")
	}

	if feedback.ValidationType == "" {
		return fmt.Errorf("validation type is required")
	}

	if feedback.DataSourceType == "" {
		return fmt.Errorf("data source type is required")
	}

	if feedback.VerificationStatus == "" {
		return fmt.Errorf("verification status is required")
	}

	// Validate status
	if !isValidFeedbackStatus(feedback.Status) {
		return fmt.Errorf("invalid feedback status: %s", feedback.Status)
	}

	// Validate trust score range
	if feedback.TrustScore < 0.0 || feedback.TrustScore > 1.0 {
		return fmt.Errorf("trust score must be between 0.0 and 1.0, got: %f", feedback.TrustScore)
	}

	// Validate processing time
	if feedback.ProcessingTimeMs < 0 {
		return fmt.Errorf("processing time cannot be negative")
	}

	// Validate timestamps
	if feedback.ProcessedAt != nil && feedback.ProcessedAt.Before(feedback.CreatedAt) {
		return fmt.Errorf("processed at time cannot be before created at time")
	}

	// Validate website URL format if provided
	if feedback.WebsiteURL != "" {
		if err := v.validateWebsiteURL(feedback.WebsiteURL); err != nil {
			return fmt.Errorf("invalid website URL: %w", err)
		}
	}

	// Validate validation result is not empty
	if feedback.ValidationResult == nil || len(feedback.ValidationResult) == 0 {
		return fmt.Errorf("validation result is required and cannot be empty")
	}

	// Security validation
	if v.securityValidator != nil {
		if err := v.securityValidator.ValidateFeedbackContent(ctx, feedback.ValidationResult); err != nil {
			return fmt.Errorf("security validation failed: %w", err)
		}
	}

	return nil
}

// ValidateFeedbackCollectionRequest validates a feedback collection request
func (v *FeedbackValidator) ValidateFeedbackCollectionRequest(ctx context.Context, request FeedbackCollectionRequest) error {
	// Validate required fields
	if request.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if request.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}

	if request.OriginalClassificationID == "" {
		return fmt.Errorf("original classification ID is required")
	}

	// Validate feedback type
	if !isValidFeedbackType(request.FeedbackType) {
		return fmt.Errorf("invalid feedback type: %s", request.FeedbackType)
	}

	// Validate confidence score range if provided
	if request.ConfidenceScore < 0.0 || request.ConfidenceScore > 1.0 {
		return fmt.Errorf("confidence score must be between 0.0 and 1.0, got: %f", request.ConfidenceScore)
	}

	// Validate business name format
	if err := v.validateBusinessName(request.BusinessName); err != nil {
		return fmt.Errorf("invalid business name: %w", err)
	}

	// Validate feedback text length
	if len(request.FeedbackText) > 5000 {
		return fmt.Errorf("feedback text exceeds maximum length of 5000 characters")
	}

	// Security validation
	if v.securityValidator != nil {
		if err := v.securityValidator.ValidateFeedbackContent(ctx, request.FeedbackValue); err != nil {
			return fmt.Errorf("security validation failed: %w", err)
		}

		if err := v.securityValidator.ValidateUserPermissions(ctx, request.UserID, request.FeedbackType); err != nil {
			return fmt.Errorf("user permission validation failed: %w", err)
		}
	}

	return nil
}

// validateBusinessName validates business name format
func (v *FeedbackValidator) validateBusinessName(name string) error {
	// Check length
	if len(name) < 1 || len(name) > 255 {
		return fmt.Errorf("business name must be between 1 and 255 characters")
	}

	// Check for valid characters (alphanumeric, spaces, common punctuation)
	validNameRegex := regexp.MustCompile(`^[a-zA-Z0-9\s\-\.&',()]+$`)
	if !validNameRegex.MatchString(name) {
		return fmt.Errorf("business name contains invalid characters")
	}

	// Check for excessive whitespace
	trimmed := strings.TrimSpace(name)
	if len(trimmed) == 0 {
		return fmt.Errorf("business name cannot be empty or only whitespace")
	}

	return nil
}

// validateWebsiteURL validates website URL format
func (v *FeedbackValidator) validateWebsiteURL(url string) error {
	// Basic URL validation regex
	urlRegex := regexp.MustCompile(`^https?://[a-zA-Z0-9\-\.]+\.[a-zA-Z]{2,}(/.*)?$`)
	if !urlRegex.MatchString(url) {
		return fmt.Errorf("invalid URL format")
	}

	// Check length
	if len(url) > 1000 {
		return fmt.Errorf("URL exceeds maximum length of 1000 characters")
	}

	return nil
}

// Helper functions for validation
func isValidFeedbackType(feedbackType FeedbackType) bool {
	validTypes := map[FeedbackType]bool{
		FeedbackTypeAccuracy:            true,
		FeedbackTypeRelevance:           true,
		FeedbackTypeConfidence:          true,
		FeedbackTypeClassification:      true,
		FeedbackTypeSuggestion:          true,
		FeedbackTypeCorrection:          true,
		FeedbackTypeMLPerformance:       true,
		FeedbackTypeModelDrift:          true,
		FeedbackTypePredictionError:     true,
		FeedbackTypeSecurityValidation:  true,
		FeedbackTypeDataSourceTrust:     true,
		FeedbackTypeWebsiteVerification: true,
	}
	return validTypes[feedbackType]
}

func isValidFeedbackStatus(status FeedbackStatus) bool {
	validStatuses := map[FeedbackStatus]bool{
		FeedbackStatusPending:   true,
		FeedbackStatusProcessed: true,
		FeedbackStatusRejected:  true,
		FeedbackStatusApplied:   true,
	}
	return validStatuses[status]
}

func isValidClassificationMethod(method ClassificationMethod) bool {
	validMethods := map[ClassificationMethod]bool{
		MethodKeyword:    true,
		MethodML:         true,
		MethodSimilarity: true,
		MethodEnsemble:   true,
		MethodSecurity:   true,
	}
	return validMethods[method]
}
