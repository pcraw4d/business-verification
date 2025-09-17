package feedback

import (
	"time"
)

// FeedbackType represents the type of feedback being collected
type FeedbackType string

const (
	// User feedback types
	FeedbackTypeAccuracy       FeedbackType = "accuracy"
	FeedbackTypeRelevance      FeedbackType = "relevance"
	FeedbackTypeConfidence     FeedbackType = "confidence"
	FeedbackTypeClassification FeedbackType = "classification"
	FeedbackTypeSuggestion     FeedbackType = "suggestion"
	FeedbackTypeCorrection     FeedbackType = "correction"

	// ML model feedback types
	FeedbackTypeMLPerformance   FeedbackType = "ml_performance"
	FeedbackTypeModelDrift      FeedbackType = "model_drift"
	FeedbackTypePredictionError FeedbackType = "prediction_error"

	// Security feedback types
	FeedbackTypeSecurityValidation  FeedbackType = "security_validation"
	FeedbackTypeDataSourceTrust     FeedbackType = "data_source_trust"
	FeedbackTypeWebsiteVerification FeedbackType = "website_verification"
)

// FeedbackStatus represents the processing status of feedback
type FeedbackStatus string

const (
	FeedbackStatusPending   FeedbackStatus = "pending"
	FeedbackStatusProcessed FeedbackStatus = "processed"
	FeedbackStatusRejected  FeedbackStatus = "rejected"
	FeedbackStatusApplied   FeedbackStatus = "applied"
)

// ClassificationMethod represents the classification method used
type ClassificationMethod string

const (
	MethodKeyword    ClassificationMethod = "keyword"
	MethodML         ClassificationMethod = "ml"
	MethodSimilarity ClassificationMethod = "similarity"
	MethodEnsemble   ClassificationMethod = "ensemble"
	MethodSecurity   ClassificationMethod = "security"
)

// UserFeedback represents user-provided feedback on classification results
type UserFeedback struct {
	ID                        string                 `json:"id" db:"id"`
	UserID                    string                 `json:"user_id" db:"user_id"`
	BusinessName              string                 `json:"business_name" db:"business_name"`
	OriginalClassificationID  string                 `json:"original_classification_id" db:"original_classification_id"`
	FeedbackType              FeedbackType           `json:"feedback_type" db:"feedback_type"`
	FeedbackValue             map[string]interface{} `json:"feedback_value" db:"feedback_value"`
	FeedbackText              string                 `json:"feedback_text" db:"feedback_text"`
	SuggestedClassificationID string                 `json:"suggested_classification_id" db:"suggested_classification_id"`
	ConfidenceScore           float64                `json:"confidence_score" db:"confidence_score"`
	Status                    FeedbackStatus         `json:"status" db:"status"`
	ProcessingTimeMs          int                    `json:"processing_time_ms" db:"processing_time_ms"`
	ModelVersionID            string                 `json:"model_version_id" db:"model_version_id"`
	ClassificationMethod      ClassificationMethod   `json:"classification_method" db:"classification_method"`
	EnsembleWeight            float64                `json:"ensemble_weight" db:"ensemble_weight"`
	CreatedAt                 time.Time              `json:"created_at" db:"created_at"`
	ProcessedAt               *time.Time             `json:"processed_at" db:"processed_at"`
	Metadata                  map[string]interface{} `json:"metadata" db:"metadata"`
}

// MLModelFeedback represents feedback on ML model performance
type MLModelFeedback struct {
	ID                   string                 `json:"id" db:"id"`
	ModelVersionID       string                 `json:"model_version_id" db:"model_version_id"`
	ModelType            string                 `json:"model_type" db:"model_type"`
	ClassificationMethod ClassificationMethod   `json:"classification_method" db:"classification_method"`
	PredictionID         string                 `json:"prediction_id" db:"prediction_id"`
	ActualResult         map[string]interface{} `json:"actual_result" db:"actual_result"`
	PredictedResult      map[string]interface{} `json:"predicted_result" db:"predicted_result"`
	AccuracyScore        float64                `json:"accuracy_score" db:"accuracy_score"`
	ConfidenceScore      float64                `json:"confidence_score" db:"confidence_score"`
	ProcessingTimeMs     int                    `json:"processing_time_ms" db:"processing_time_ms"`
	ErrorType            string                 `json:"error_type" db:"error_type"`
	ErrorDescription     string                 `json:"error_description" db:"error_description"`
	Status               FeedbackStatus         `json:"status" db:"status"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	ProcessedAt          *time.Time             `json:"processed_at" db:"processed_at"`
	Metadata             map[string]interface{} `json:"metadata" db:"metadata"`
}

// SecurityValidationFeedback represents feedback on security validation processes
type SecurityValidationFeedback struct {
	ID                 string                 `json:"id" db:"id"`
	ValidationType     string                 `json:"validation_type" db:"validation_type"`
	DataSourceType     string                 `json:"data_source_type" db:"data_source_type"`
	WebsiteURL         string                 `json:"website_url" db:"website_url"`
	ValidationResult   map[string]interface{} `json:"validation_result" db:"validation_result"`
	TrustScore         float64                `json:"trust_score" db:"trust_score"`
	VerificationStatus string                 `json:"verification_status" db:"verification_status"`
	SecurityViolations []string               `json:"security_violations" db:"security_violations"`
	ProcessingTimeMs   int                    `json:"processing_time_ms" db:"processing_time_ms"`
	Status             FeedbackStatus         `json:"status" db:"status"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	ProcessedAt        *time.Time             `json:"processed_at" db:"processed_at"`
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
}

// FeedbackTrend represents aggregated feedback trends across ensemble methods
type FeedbackTrend struct {
	Method                ClassificationMethod `json:"method" db:"method"`
	TimeWindow            string               `json:"time_window" db:"time_window"`
	TotalFeedback         int                  `json:"total_feedback" db:"total_feedback"`
	PositiveFeedback      int                  `json:"positive_feedback" db:"positive_feedback"`
	NegativeFeedback      int                  `json:"negative_feedback" db:"negative_feedback"`
	AverageAccuracy       float64              `json:"average_accuracy" db:"average_accuracy"`
	AverageConfidence     float64              `json:"average_confidence" db:"average_confidence"`
	AverageProcessingTime int                  `json:"average_processing_time" db:"average_processing_time"`
	ErrorRate             float64              `json:"error_rate" db:"error_rate"`
	SecurityViolationRate float64              `json:"security_violation_rate" db:"security_violation_rate"`
	CreatedAt             time.Time            `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at" db:"updated_at"`
}

// FeedbackCollectionRequest represents a request to collect feedback
type FeedbackCollectionRequest struct {
	UserID                    string                 `json:"user_id"`
	BusinessName              string                 `json:"business_name"`
	OriginalClassificationID  string                 `json:"original_classification_id"`
	FeedbackType              FeedbackType           `json:"feedback_type"`
	FeedbackValue             map[string]interface{} `json:"feedback_value"`
	FeedbackText              string                 `json:"feedback_text"`
	SuggestedClassificationID string                 `json:"suggested_classification_id,omitempty"`
	ConfidenceScore           float64                `json:"confidence_score,omitempty"`
	Metadata                  map[string]interface{} `json:"metadata,omitempty"`
}

// FeedbackCollectionResponse represents the response from feedback collection
type FeedbackCollectionResponse struct {
	ID               string    `json:"id"`
	Status           string    `json:"status"`
	ProcessingTimeMs int       `json:"processing_time_ms"`
	Message          string    `json:"message"`
	CreatedAt        time.Time `json:"created_at"`
}

// FeedbackAnalysisRequest represents a request to analyze feedback trends
type FeedbackAnalysisRequest struct {
	Method       ClassificationMethod `json:"method,omitempty"`
	TimeWindow   string               `json:"time_window,omitempty"`
	FeedbackType FeedbackType         `json:"feedback_type,omitempty"`
	StartDate    *time.Time           `json:"start_date,omitempty"`
	EndDate      *time.Time           `json:"end_date,omitempty"`
}

// FeedbackAnalysisResponse represents the response from feedback analysis
type FeedbackAnalysisResponse struct {
	Trends                []FeedbackTrend `json:"trends"`
	TotalFeedback         int             `json:"total_feedback"`
	AverageAccuracy       float64         `json:"average_accuracy"`
	AverageConfidence     float64         `json:"average_confidence"`
	ErrorRate             float64         `json:"error_rate"`
	SecurityViolationRate float64         `json:"security_violation_rate"`
	AnalysisTime          time.Time       `json:"analysis_time"`
}
