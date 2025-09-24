package feedback

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// FeedbackService implements the main feedback service interface
type FeedbackService struct {
	repository          FeedbackRepository
	validator           *FeedbackValidator
	analyzer            FeedbackAnalyzer
	processor           FeedbackProcessor
	modelVersionManager ModelVersionManager
	securityValidator   SecurityValidator
	logger              *zap.Logger
}

// NewFeedbackService creates a new feedback service
func NewFeedbackService(
	repository FeedbackRepository,
	validator *FeedbackValidator,
	analyzer FeedbackAnalyzer,
	processor FeedbackProcessor,
	modelVersionManager ModelVersionManager,
	securityValidator SecurityValidator,
	logger *zap.Logger,
) *FeedbackService {
	return &FeedbackService{
		repository:          repository,
		validator:           validator,
		analyzer:            analyzer,
		processor:           processor,
		modelVersionManager: modelVersionManager,
		securityValidator:   securityValidator,
		logger:              logger,
	}
}

// CollectUserFeedback collects user feedback on classification results
func (s *FeedbackService) CollectUserFeedback(ctx context.Context, request FeedbackCollectionRequest) (*FeedbackCollectionResponse, error) {
	startTime := time.Now()

	s.logger.Info("collecting user feedback",
		zap.String("user_id", request.UserID),
		zap.String("business_name", request.BusinessName),
		zap.String("feedback_type", string(request.FeedbackType)),
		zap.String("original_classification_id", request.OriginalClassificationID))

	// Validate the request
	if err := s.validator.ValidateFeedbackCollectionRequest(ctx, request); err != nil {
		s.logger.Error("feedback validation failed",
			zap.String("user_id", request.UserID),
			zap.Error(err))
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Get current model version for the classification method
	modelVersion, err := s.modelVersionManager.GetCurrentModelVersion(ctx, "ensemble")
	if err != nil {
		s.logger.Warn("failed to get current model version, using default",
			zap.Error(err))
		modelVersion = "default"
	}

	// Create user feedback record
	feedback := UserFeedback{
		ID:                        generateID(),
		UserID:                    request.UserID,
		BusinessName:              request.BusinessName,
		OriginalClassificationID:  request.OriginalClassificationID,
		FeedbackType:              request.FeedbackType,
		FeedbackValue:             request.FeedbackValue,
		FeedbackText:              request.FeedbackText,
		SuggestedClassificationID: request.SuggestedClassificationID,
		ConfidenceScore:           request.ConfidenceScore,
		Status:                    FeedbackStatusPending,
		ProcessingTimeMs:          0,
		ModelVersionID:            modelVersion,
		ClassificationMethod:      MethodEnsemble, // Default to ensemble
		EnsembleWeight:            0.5,            // Default weight
		CreatedAt:                 time.Now(),
		ProcessedAt:               nil,
		Metadata:                  request.Metadata,
	}

	// Validate the feedback record
	if err := s.validator.ValidateUserFeedback(ctx, feedback); err != nil {
		s.logger.Error("user feedback validation failed",
			zap.String("feedback_id", feedback.ID),
			zap.Error(err))
		return nil, fmt.Errorf("feedback validation failed: %w", err)
	}

	// Save the feedback
	if err := s.repository.SaveUserFeedback(ctx, feedback); err != nil {
		s.logger.Error("failed to save user feedback",
			zap.String("feedback_id", feedback.ID),
			zap.Error(err))
		return nil, fmt.Errorf("failed to save feedback: %w", err)
	}

	// Process the feedback asynchronously
	go func() {
		processCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.processor.ProcessUserFeedback(processCtx, feedback); err != nil {
			s.logger.Error("failed to process user feedback",
				zap.String("feedback_id", feedback.ID),
				zap.Error(err))
		}
	}()

	processingTime := int(time.Since(startTime).Milliseconds())

	s.logger.Info("user feedback collected successfully",
		zap.String("feedback_id", feedback.ID),
		zap.Int("processing_time_ms", processingTime))

	return &FeedbackCollectionResponse{
		ID:               feedback.ID,
		Status:           string(feedback.Status),
		ProcessingTimeMs: processingTime,
		Message:          "Feedback collected successfully",
		CreatedAt:        feedback.CreatedAt,
	}, nil
}

// CollectMLModelFeedback collects ML model performance feedback
func (s *FeedbackService) CollectMLModelFeedback(ctx context.Context, feedback MLModelFeedback) error {
	s.logger.Info("collecting ML model feedback",
		zap.String("model_version_id", feedback.ModelVersionID),
		zap.String("model_type", feedback.ModelType),
		zap.String("classification_method", string(feedback.ClassificationMethod)),
		zap.String("prediction_id", feedback.PredictionID))

	// Validate the feedback
	if err := s.validator.ValidateMLModelFeedback(ctx, feedback); err != nil {
		s.logger.Error("ML model feedback validation failed",
			zap.String("feedback_id", feedback.ID),
			zap.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}

	// Save the feedback
	if err := s.repository.SaveMLModelFeedback(ctx, feedback); err != nil {
		s.logger.Error("failed to save ML model feedback",
			zap.String("feedback_id", feedback.ID),
			zap.Error(err))
		return fmt.Errorf("failed to save feedback: %w", err)
	}

	// Process the feedback asynchronously
	go func() {
		processCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.processor.ProcessMLModelFeedback(processCtx, feedback); err != nil {
			s.logger.Error("failed to process ML model feedback",
				zap.String("feedback_id", feedback.ID),
				zap.Error(err))
		}
	}()

	s.logger.Info("ML model feedback collected successfully",
		zap.String("feedback_id", feedback.ID))

	return nil
}

// CollectSecurityValidationFeedback collects security validation feedback
func (s *FeedbackService) CollectSecurityValidationFeedback(ctx context.Context, feedback SecurityValidationFeedback) error {
	s.logger.Info("collecting security validation feedback",
		zap.String("validation_type", feedback.ValidationType),
		zap.String("data_source_type", feedback.DataSourceType),
		zap.String("website_url", feedback.WebsiteURL))

	// Validate the feedback
	if err := s.validator.ValidateSecurityValidationFeedback(ctx, feedback); err != nil {
		s.logger.Error("security validation feedback validation failed",
			zap.String("feedback_id", feedback.ID),
			zap.Error(err))
		return fmt.Errorf("validation failed: %w", err)
	}

	// Save the feedback
	if err := s.repository.SaveSecurityValidationFeedback(ctx, feedback); err != nil {
		s.logger.Error("failed to save security validation feedback",
			zap.String("feedback_id", feedback.ID),
			zap.Error(err))
		return fmt.Errorf("failed to save feedback: %w", err)
	}

	// Process the feedback asynchronously
	go func() {
		processCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.processor.ProcessSecurityValidationFeedback(processCtx, feedback); err != nil {
			s.logger.Error("failed to process security validation feedback",
				zap.String("feedback_id", feedback.ID),
				zap.Error(err))
		}
	}()

	s.logger.Info("security validation feedback collected successfully",
		zap.String("feedback_id", feedback.ID))

	return nil
}

// CollectBatchFeedback collects multiple feedback items in batch
func (s *FeedbackService) CollectBatchFeedback(ctx context.Context, requests []FeedbackCollectionRequest) ([]FeedbackCollectionResponse, error) {
	s.logger.Info("collecting batch feedback",
		zap.Int("request_count", len(requests)))

	responses := make([]FeedbackCollectionResponse, 0, len(requests))

	for i, request := range requests {
		response, err := s.CollectUserFeedback(ctx, request)
		if err != nil {
			s.logger.Error("failed to collect feedback in batch",
				zap.Int("request_index", i),
				zap.String("user_id", request.UserID),
				zap.Error(err))

			// Continue with other requests even if one fails
			responses = append(responses, FeedbackCollectionResponse{
				ID:        "",
				Status:    "failed",
				Message:   fmt.Sprintf("Failed to collect feedback: %v", err),
				CreatedAt: time.Now(),
			})
			continue
		}

		responses = append(responses, *response)
	}

	s.logger.Info("batch feedback collection completed",
		zap.Int("total_requests", len(requests)),
		zap.Int("successful_responses", len(responses)))

	return responses, nil
}

// AnalyzeFeedbackTrends analyzes feedback trends across ensemble methods
func (s *FeedbackService) AnalyzeFeedbackTrends(ctx context.Context, request FeedbackAnalysisRequest) (*FeedbackAnalysisResponse, error) {
	s.logger.Info("analyzing feedback trends",
		zap.String("method", string(request.Method)),
		zap.String("time_window", request.TimeWindow),
		zap.String("feedback_type", string(request.FeedbackType)))

	if s.analyzer == nil {
		return nil, fmt.Errorf("feedback analyzer not configured")
	}

	response, err := s.analyzer.AnalyzeFeedbackTrends(ctx, request)
	if err != nil {
		s.logger.Error("failed to analyze feedback trends",
			zap.Error(err))
		return nil, fmt.Errorf("analysis failed: %w", err)
	}

	s.logger.Info("feedback trends analysis completed",
		zap.Int("trend_count", len(response.Trends)),
		zap.Float64("average_accuracy", response.AverageAccuracy),
		zap.Float64("error_rate", response.ErrorRate))

	return response, nil
}

// IdentifyFeedbackPatterns identifies patterns in feedback data
func (s *FeedbackService) IdentifyFeedbackPatterns(ctx context.Context, method ClassificationMethod, timeWindow string) (map[string]interface{}, error) {
	s.logger.Info("identifying feedback patterns",
		zap.String("method", string(method)),
		zap.String("time_window", timeWindow))

	if s.analyzer == nil {
		return nil, fmt.Errorf("feedback analyzer not configured")
	}

	patterns, err := s.analyzer.IdentifyFeedbackPatterns(ctx, method, timeWindow)
	if err != nil {
		s.logger.Error("failed to identify feedback patterns",
			zap.Error(err))
		return nil, fmt.Errorf("pattern identification failed: %w", err)
	}

	s.logger.Info("feedback patterns identified",
		zap.Int("pattern_count", len(patterns)))

	return patterns, nil
}

// CalculateMethodPerformance calculates performance metrics for a classification method
func (s *FeedbackService) CalculateMethodPerformance(ctx context.Context, method ClassificationMethod) (map[string]float64, error) {
	s.logger.Info("calculating method performance",
		zap.String("method", string(method)))

	if s.analyzer == nil {
		return nil, fmt.Errorf("feedback analyzer not configured")
	}

	performance, err := s.analyzer.CalculateMethodPerformance(ctx, method)
	if err != nil {
		s.logger.Error("failed to calculate method performance",
			zap.Error(err))
		return nil, fmt.Errorf("performance calculation failed: %w", err)
	}

	s.logger.Info("method performance calculated",
		zap.Int("metric_count", len(performance)))

	return performance, nil
}

// DetectAnomalies detects anomalies in feedback data
func (s *FeedbackService) DetectAnomalies(ctx context.Context, method ClassificationMethod) ([]string, error) {
	s.logger.Info("detecting anomalies",
		zap.String("method", string(method)))

	if s.analyzer == nil {
		return nil, fmt.Errorf("feedback analyzer not configured")
	}

	anomalies, err := s.analyzer.DetectAnomalies(ctx, method)
	if err != nil {
		s.logger.Error("failed to detect anomalies",
			zap.Error(err))
		return nil, fmt.Errorf("anomaly detection failed: %w", err)
	}

	s.logger.Info("anomalies detected",
		zap.Int("anomaly_count", len(anomalies)))

	return anomalies, nil
}

// ProcessUserFeedback processes user feedback for learning
func (s *FeedbackService) ProcessUserFeedback(ctx context.Context, feedback UserFeedback) error {
	if s.processor == nil {
		return fmt.Errorf("feedback processor not configured")
	}

	return s.processor.ProcessUserFeedback(ctx, feedback)
}

// ProcessMLModelFeedback processes ML model feedback for learning
func (s *FeedbackService) ProcessMLModelFeedback(ctx context.Context, feedback MLModelFeedback) error {
	if s.processor == nil {
		return fmt.Errorf("feedback processor not configured")
	}

	return s.processor.ProcessMLModelFeedback(ctx, feedback)
}

// ProcessSecurityValidationFeedback processes security validation feedback
func (s *FeedbackService) ProcessSecurityValidationFeedback(ctx context.Context, feedback SecurityValidationFeedback) error {
	if s.processor == nil {
		return fmt.Errorf("feedback processor not configured")
	}

	return s.processor.ProcessSecurityValidationFeedback(ctx, feedback)
}

// ApplyFeedbackCorrections applies feedback corrections to the system
func (s *FeedbackService) ApplyFeedbackCorrections(ctx context.Context, feedbackID string) error {
	if s.processor == nil {
		return fmt.Errorf("feedback processor not configured")
	}

	return s.processor.ApplyFeedbackCorrections(ctx, feedbackID)
}

// GenerateFeedbackInsights generates insights from feedback data
func (s *FeedbackService) GenerateFeedbackInsights(ctx context.Context, method ClassificationMethod) (map[string]interface{}, error) {
	if s.processor == nil {
		return nil, fmt.Errorf("feedback processor not configured")
	}

	return s.processor.GenerateFeedbackInsights(ctx, method)
}

// GetServiceHealth returns the health status of the feedback service
func (s *FeedbackService) GetServiceHealth(ctx context.Context) (map[string]interface{}, error) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"components": map[string]interface{}{
			"repository":            s.repository != nil,
			"validator":             s.validator != nil,
			"analyzer":              s.analyzer != nil,
			"processor":             s.processor != nil,
			"model_version_manager": s.modelVersionManager != nil,
			"security_validator":    s.securityValidator != nil,
		},
	}

	return health, nil
}

// GetServiceMetrics returns metrics for the feedback service
func (s *FeedbackService) GetServiceMetrics(ctx context.Context) (map[string]interface{}, error) {
	metrics := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"service":   "feedback_collection",
		"version":   "1.0.0",
		"metrics": map[string]interface{}{
			"active_components": 6, // All components are active
			"service_status":    "operational",
		},
	}

	return metrics, nil
}

// generateID generates a unique ID for feedback records
func generateID() string {
	return fmt.Sprintf("feedback_%d", time.Now().UnixNano())
}
