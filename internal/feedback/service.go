package feedback

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	// Stub: modelVersion not used in UserFeedback struct
	_, err := s.modelVersionManager.GetCurrentModelVersion(ctx, "ensemble")
	if err != nil {
		s.logger.Warn("failed to get current model version, using default",
			zap.Error(err))
	}

	// Create user feedback record
	// Map FeedbackCollectionRequest fields to UserFeedback, preserving all data
	feedback := UserFeedback{
		ID:                     uuid.New(),
		UserID:                 request.UserID,
		Category:               s.mapFeedbackTypeToCategory(request.FeedbackType),
		Comments:               request.FeedbackText,
		ClassificationAccuracy: request.ConfidenceScore, // Map ConfidenceScore to ClassificationAccuracy
		SubmittedAt:            time.Now(),
		Metadata:               s.buildFeedbackMetadata(request),
	}

	// Extract rating from FeedbackValue if available
	if rating, ok := request.FeedbackValue["rating"].(float64); ok {
		feedback.Rating = int(rating)
	} else if rating, ok := request.FeedbackValue["rating"].(int); ok {
		feedback.Rating = rating
	}

	// Extract specific features from FeedbackValue if available
	if features, ok := request.FeedbackValue["specific_features"].([]interface{}); ok {
		feedback.SpecificFeatures = s.extractStringSlice(features)
	}

	// Extract improvement areas from FeedbackValue if available
	if areas, ok := request.FeedbackValue["improvement_areas"].([]interface{}); ok {
		feedback.ImprovementAreas = s.extractStringSlice(areas)
	}

	// Extract performance rating from FeedbackValue if available
	if perfRating, ok := request.FeedbackValue["performance_rating"].(float64); ok {
		feedback.PerformanceRating = int(perfRating)
	} else if perfRating, ok := request.FeedbackValue["performance_rating"].(int); ok {
		feedback.PerformanceRating = perfRating
	}

	// Extract usability rating from FeedbackValue if available
	if usabilityRating, ok := request.FeedbackValue["usability_rating"].(float64); ok {
		feedback.UsabilityRating = int(usabilityRating)
	} else if usabilityRating, ok := request.FeedbackValue["usability_rating"].(int); ok {
		feedback.UsabilityRating = usabilityRating
	}

	// Extract business impact from FeedbackValue if available
	if businessImpact, ok := request.FeedbackValue["business_impact"].(map[string]interface{}); ok {
		feedback.BusinessImpact = s.mapToBusinessImpactRating(businessImpact)
	}

	// Validate the feedback record
	if err := s.validator.ValidateUserFeedback(ctx, feedback); err != nil {
		s.logger.Error("user feedback validation failed",
			zap.String("feedback_id", feedback.ID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("feedback validation failed: %w", err)
	}

	// Save the feedback
	if err := s.repository.SaveUserFeedback(ctx, feedback); err != nil {
		s.logger.Error("failed to save user feedback",
			zap.String("feedback_id", feedback.ID.String()),
			zap.Error(err))
		return nil, fmt.Errorf("failed to save feedback: %w", err)
	}

	// Process the feedback asynchronously
	go func() {
		processCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.processor.ProcessUserFeedback(processCtx, feedback); err != nil {
			s.logger.Error("failed to process user feedback",
				zap.String("feedback_id", feedback.ID.String()),
				zap.Error(err))
		}
	}()

	processingTime := int(time.Since(startTime).Milliseconds())

	s.logger.Info("user feedback collected successfully",
		zap.String("feedback_id", feedback.ID.String()),
		zap.Int("processing_time_ms", processingTime))

	return &FeedbackCollectionResponse{
		ID:               feedback.ID.String(), // Convert UUID to string
		Status:           "pending",            // Stub - UserFeedback doesn't have Status
		ProcessingTimeMs: processingTime,
		Message:          "Feedback collected successfully",
		CreatedAt:        feedback.SubmittedAt, // Use SubmittedAt as CreatedAt substitute
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

// mapFeedbackTypeToCategory maps FeedbackType to FeedbackCategory
func (s *FeedbackService) mapFeedbackTypeToCategory(feedbackType FeedbackType) FeedbackCategory {
	switch feedbackType {
	case FeedbackTypeAccuracy, FeedbackTypeClassification:
		return CategoryClassificationAccuracy
	case FeedbackTypeConfidence:
		return CategoryClassificationAccuracy
	case FeedbackTypeRelevance:
		return CategoryUserExperience
	case FeedbackTypeSuggestion:
		return CategoryFeatureRequest
	case FeedbackTypeCorrection:
		return CategoryBugReport
	case FeedbackTypeSecurityValidation:
		return CategoryRiskDetection
	default:
		return CategoryOverallSatisfaction
	}
}

// buildFeedbackMetadata builds comprehensive metadata from FeedbackCollectionRequest
// This preserves ALL data from the request to prevent data loss during migration
// All fields from FeedbackCollectionRequest are explicitly preserved for future use
func (s *FeedbackService) buildFeedbackMetadata(request FeedbackCollectionRequest) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Copy existing metadata first (may contain additional context)
	if request.Metadata != nil {
		for k, v := range request.Metadata {
			metadata[k] = v
		}
	}

	// Explicitly preserve ALL request fields to prevent data loss
	// These fields may be needed for downstream processing or future refactoring
	metadata["feedback_type"] = string(request.FeedbackType)
	metadata["business_name"] = request.BusinessName
	metadata["original_classification_id"] = request.OriginalClassificationID
	metadata["suggested_classification_id"] = request.SuggestedClassificationID
	metadata["confidence_score"] = request.ConfidenceScore
	metadata["feedback_text"] = request.FeedbackText // Preserve original FeedbackText (mapped to Comments)
	metadata["feedback_value"] = request.FeedbackValue // Preserve entire FeedbackValue map

	// Preserve field mapping information for traceability
	metadata["field_mapping"] = map[string]interface{}{
		"feedback_type_to_category":           string(s.mapFeedbackTypeToCategory(request.FeedbackType)),
		"feedback_text_to_comments":           true,
		"confidence_score_to_classification_accuracy": true,
		"feedback_value_extracted_fields": []string{
			"rating",
			"specific_features",
			"improvement_areas",
			"performance_rating",
			"usability_rating",
			"business_impact",
		},
	}

	// Add timestamp for when feedback was collected
	metadata["collected_at"] = time.Now().Format(time.RFC3339)
	metadata["migration_version"] = "2.0" // Track migration version for future compatibility

	return metadata
}

// extractStringSlice safely extracts a []string from []interface{}
func (s *FeedbackService) extractStringSlice(items []interface{}) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// mapToBusinessImpactRating maps a map[string]interface{} to BusinessImpactRating
func (s *FeedbackService) mapToBusinessImpactRating(data map[string]interface{}) BusinessImpactRating {
	impact := BusinessImpactRating{}

	if timeSaved, ok := data["time_saved_minutes"].(int); ok {
		impact.TimeSaved = timeSaved
	} else if timeSaved, ok := data["time_saved_minutes"].(float64); ok {
		impact.TimeSaved = int(timeSaved)
	}

	if costReduction, ok := data["cost_reduction"].(string); ok {
		impact.CostReduction = costReduction
	}

	if errorReduction, ok := data["error_reduction_percentage"].(int); ok {
		impact.ErrorReduction = errorReduction
	} else if errorReduction, ok := data["error_reduction_percentage"].(float64); ok {
		impact.ErrorReduction = int(errorReduction)
	}

	if productivityGain, ok := data["productivity_gain_percentage"].(int); ok {
		impact.ProductivityGain = productivityGain
	} else if productivityGain, ok := data["productivity_gain_percentage"].(float64); ok {
		impact.ProductivityGain = int(productivityGain)
	}

	if roi, ok := data["roi_assessment"].(string); ok {
		impact.ROI = roi
	}

	return impact
}
