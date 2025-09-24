package feedback

import (
	"context"
)

// FeedbackCollector defines the interface for collecting various types of feedback
type FeedbackCollector interface {
	// User feedback collection
	CollectUserFeedback(ctx context.Context, request FeedbackCollectionRequest) (*FeedbackCollectionResponse, error)

	// ML model feedback collection
	CollectMLModelFeedback(ctx context.Context, feedback MLModelFeedback) error

	// Security validation feedback collection
	CollectSecurityValidationFeedback(ctx context.Context, feedback SecurityValidationFeedback) error

	// Batch feedback collection
	CollectBatchFeedback(ctx context.Context, requests []FeedbackCollectionRequest) ([]FeedbackCollectionResponse, error)
}

// FeedbackRepository defines the interface for feedback data persistence
type FeedbackRepository interface {
	// User feedback operations
	SaveUserFeedback(ctx context.Context, feedback UserFeedback) error
	GetUserFeedback(ctx context.Context, id string) (*UserFeedback, error)
	GetUserFeedbackByClassificationID(ctx context.Context, classificationID string) ([]UserFeedback, error)
	UpdateUserFeedbackStatus(ctx context.Context, id string, status FeedbackStatus) error

	// ML model feedback operations
	SaveMLModelFeedback(ctx context.Context, feedback MLModelFeedback) error
	GetMLModelFeedback(ctx context.Context, id string) (*MLModelFeedback, error)
	GetMLModelFeedbackByModelVersion(ctx context.Context, modelVersion string) ([]MLModelFeedback, error)
	UpdateMLModelFeedbackStatus(ctx context.Context, id string, status FeedbackStatus) error

	// Security validation feedback operations
	SaveSecurityValidationFeedback(ctx context.Context, feedback SecurityValidationFeedback) error
	GetSecurityValidationFeedback(ctx context.Context, id string) (*SecurityValidationFeedback, error)
	GetSecurityValidationFeedbackByType(ctx context.Context, validationType string) ([]SecurityValidationFeedback, error)
	UpdateSecurityValidationFeedbackStatus(ctx context.Context, id string, status FeedbackStatus) error

	// Feedback trend analysis
	GetFeedbackTrends(ctx context.Context, request FeedbackAnalysisRequest) ([]FeedbackTrend, error)
	GetFeedbackStatistics(ctx context.Context, request FeedbackAnalysisRequest) (*FeedbackAnalysisResponse, error)

	// Batch operations
	SaveBatchUserFeedback(ctx context.Context, feedback []UserFeedback) error
	SaveBatchMLModelFeedback(ctx context.Context, feedback []MLModelFeedback) error
	SaveBatchSecurityValidationFeedback(ctx context.Context, feedback []SecurityValidationFeedback) error
}

// FeedbackValidatorInterface defines the interface for validating feedback data
type FeedbackValidatorInterface interface {
	ValidateUserFeedback(ctx context.Context, feedback UserFeedback) error
	ValidateMLModelFeedback(ctx context.Context, feedback MLModelFeedback) error
	ValidateSecurityValidationFeedback(ctx context.Context, feedback SecurityValidationFeedback) error
	ValidateFeedbackCollectionRequest(ctx context.Context, request FeedbackCollectionRequest) error
}

// FeedbackAnalyzer defines the interface for analyzing feedback trends and patterns
type FeedbackAnalyzer interface {
	AnalyzeFeedbackTrends(ctx context.Context, request FeedbackAnalysisRequest) (*FeedbackAnalysisResponse, error)
	IdentifyFeedbackPatterns(ctx context.Context, method ClassificationMethod, timeWindow string) (map[string]interface{}, error)
	CalculateMethodPerformance(ctx context.Context, method ClassificationMethod) (map[string]float64, error)
	DetectAnomalies(ctx context.Context, method ClassificationMethod) ([]string, error)
}

// FeedbackProcessor defines the interface for processing and applying feedback
type FeedbackProcessor interface {
	ProcessUserFeedback(ctx context.Context, feedback UserFeedback) error
	ProcessMLModelFeedback(ctx context.Context, feedback MLModelFeedback) error
	ProcessSecurityValidationFeedback(ctx context.Context, feedback SecurityValidationFeedback) error
	ApplyFeedbackCorrections(ctx context.Context, feedbackID string) error
	GenerateFeedbackInsights(ctx context.Context, method ClassificationMethod) (map[string]interface{}, error)
}

// FeedbackServiceInterface defines the main service interface for feedback collection and management
type FeedbackServiceInterface interface {
	FeedbackCollector
	FeedbackAnalyzer
	FeedbackProcessor

	// Service management
	GetServiceHealth(ctx context.Context) (map[string]interface{}, error)
	GetServiceMetrics(ctx context.Context) (map[string]interface{}, error)
}

// ModelVersionManager defines the interface for managing model versions in feedback
type ModelVersionManager interface {
	GetCurrentModelVersion(ctx context.Context, modelType string) (string, error)
	GetModelVersionHistory(ctx context.Context, modelType string) ([]string, error)
	RegisterModelVersion(ctx context.Context, modelType, version string, metadata map[string]interface{}) error
	GetModelVersionMetadata(ctx context.Context, modelType, version string) (map[string]interface{}, error)
}

// SecurityValidator defines the interface for security validation in feedback collection
type SecurityValidator interface {
	ValidateFeedbackSource(ctx context.Context, source string) error
	ValidateFeedbackContent(ctx context.Context, content map[string]interface{}) error
	CheckSecurityViolations(ctx context.Context, feedback interface{}) ([]string, error)
	ValidateUserPermissions(ctx context.Context, userID string, feedbackType FeedbackType) error
}
