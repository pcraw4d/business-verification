package classification

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// FeedbackType represents different types of feedback
type FeedbackType string

const (
	FeedbackTypeAccuracy       FeedbackType = "accuracy"
	FeedbackTypeRelevance      FeedbackType = "relevance"
	FeedbackTypeConfidence     FeedbackType = "confidence"
	FeedbackTypeClassification FeedbackType = "classification"
	FeedbackTypeSuggestion     FeedbackType = "suggestion"
	FeedbackTypeCorrection     FeedbackType = "correction"
)

// FeedbackStatus represents the status of feedback processing
type FeedbackStatus string

const (
	FeedbackStatusPending   FeedbackStatus = "pending"
	FeedbackStatusProcessed FeedbackStatus = "processed"
	FeedbackStatusRejected  FeedbackStatus = "rejected"
	FeedbackStatusApplied   FeedbackStatus = "applied"
)

// Feedback represents user feedback for classification results
type Feedback struct {
	ID                      string                  `json:"id"`
	UserID                  string                  `json:"user_id"`
	BusinessName            string                  `json:"business_name"`
	OriginalClassification  *IndustryClassification `json:"original_classification"`
	FeedbackType            FeedbackType            `json:"feedback_type"`
	FeedbackValue           interface{}             `json:"feedback_value"`
	FeedbackText            string                  `json:"feedback_text"`
	SuggestedClassification *IndustryClassification `json:"suggested_classification,omitempty"`
	Confidence              float64                 `json:"confidence"`
	Status                  FeedbackStatus          `json:"status"`
	ProcessingTime          time.Duration           `json:"processing_time"`
	CreatedAt               time.Time               `json:"created_at"`
	ProcessedAt             *time.Time              `json:"processed_at,omitempty"`
	Metadata                map[string]interface{}  `json:"metadata"`
}

// FeedbackValidationRule represents a validation rule for feedback
type FeedbackValidationRule struct {
	RuleID     string                 `json:"rule_id"`
	RuleType   string                 `json:"rule_type"` // "format", "content", "business_logic"
	Condition  string                 `json:"condition"`
	Severity   string                 `json:"severity"` // "low", "medium", "high", "critical"
	Weight     float64                `json:"weight"`
	Enabled    bool                   `json:"enabled"`
	Parameters map[string]interface{} `json:"parameters"`
}

// FeedbackAccuracyMetrics represents accuracy metrics based on feedback
type FeedbackAccuracyMetrics struct {
	TotalFeedback     int                    `json:"total_feedback"`
	PositiveFeedback  int                    `json:"positive_feedback"`
	NegativeFeedback  int                    `json:"negative_feedback"`
	AccuracyScore     float64                `json:"accuracy_score"`
	ConfidenceScore   float64                `json:"confidence_score"`
	IndustryBreakdown map[string]interface{} `json:"industry_breakdown"`
	TimeRange         time.Duration          `json:"time_range"`
	LastUpdated       time.Time              `json:"last_updated"`
}

// FeedbackCollector provides real-time feedback collection capabilities
type FeedbackCollector struct {
	logger  *observability.Logger
	metrics *observability.Metrics

	// Feedback storage
	feedback      map[string]*Feedback
	feedbackMutex sync.RWMutex

	// Validation rules
	validationRules map[string]FeedbackValidationRule
	rulesMutex      sync.RWMutex

	// Accuracy tracking
	accuracyMetrics map[string]*FeedbackAccuracyMetrics
	accuracyMutex   sync.RWMutex

	// Model update tracking
	modelUpdates      map[string]interface{}
	modelUpdatesMutex sync.RWMutex

	// Configuration
	enableFeedbackCollection bool
	feedbackRetentionDays    int
	maxFeedbackPerUser       int
	accuracyUpdateInterval   time.Duration
}

// NewFeedbackCollector creates a new feedback collector
func NewFeedbackCollector(logger *observability.Logger, metrics *observability.Metrics) *FeedbackCollector {
	collector := &FeedbackCollector{
		logger:  logger,
		metrics: metrics,

		// Initialize storage
		feedback:        make(map[string]*Feedback),
		validationRules: make(map[string]FeedbackValidationRule),
		accuracyMetrics: make(map[string]*FeedbackAccuracyMetrics),
		modelUpdates:    make(map[string]interface{}),

		// Configuration
		enableFeedbackCollection: true,
		feedbackRetentionDays:    30,
		maxFeedbackPerUser:       100,
		accuracyUpdateInterval:   time.Hour,
	}

	// Initialize validation rules
	collector.initializeValidationRules()

	// Start accuracy update goroutine
	go collector.startAccuracyUpdateLoop()

	return collector
}

// SubmitFeedback submits user feedback for classification results
func (fc *FeedbackCollector) SubmitFeedback(ctx context.Context, feedback *Feedback) error {
	start := time.Now()

	// Log feedback submission
	if fc.logger != nil {
		fc.logger.WithComponent("feedback_collector").LogBusinessEvent(ctx, "feedback_submission_started", "", map[string]interface{}{
			"user_id":       feedback.UserID,
			"business_name": feedback.BusinessName,
			"feedback_type": string(feedback.FeedbackType),
		})
	}

	// Validate feedback
	if err := fc.validateFeedback(feedback); err != nil {
		// Log validation failure
		if fc.logger != nil {
			fc.logger.WithComponent("feedback_collector").LogBusinessEvent(ctx, "feedback_validation_failed", "", map[string]interface{}{
				"user_id":       feedback.UserID,
				"business_name": feedback.BusinessName,
				"error":         err.Error(),
			})
		}
		return fmt.Errorf("feedback validation failed: %w", err)
	}

	// Set feedback metadata
	feedback.ID = fc.generateFeedbackID()
	feedback.Status = FeedbackStatusPending
	feedback.CreatedAt = time.Now()
	if feedback.Metadata == nil {
		feedback.Metadata = make(map[string]interface{})
	}

	// Store feedback
	fc.feedbackMutex.Lock()
	fc.feedback[feedback.ID] = feedback
	fc.feedbackMutex.Unlock()

	// Process feedback
	go fc.processFeedback(ctx, feedback)

	// Log feedback submission completion
	if fc.logger != nil {
		fc.logger.WithComponent("feedback_collector").LogBusinessEvent(ctx, "feedback_submission_completed", "", map[string]interface{}{
			"feedback_id":        feedback.ID,
			"user_id":            feedback.UserID,
			"business_name":      feedback.BusinessName,
			"processing_time_ms": time.Since(start).Milliseconds(),
		})
	}

	// Record metrics
	fc.RecordFeedbackMetrics(ctx, feedback, "submitted")

	return nil
}

// GetFeedback retrieves feedback by ID
func (fc *FeedbackCollector) GetFeedback(ctx context.Context, feedbackID string) (*Feedback, error) {
	fc.feedbackMutex.RLock()
	defer fc.feedbackMutex.RUnlock()

	feedback, exists := fc.feedback[feedbackID]
	if !exists {
		return nil, fmt.Errorf("feedback not found: %s", feedbackID)
	}

	return feedback, nil
}

// ListFeedback returns feedback with optional filtering
func (fc *FeedbackCollector) ListFeedback(ctx context.Context, filters map[string]interface{}) ([]*Feedback, error) {
	fc.feedbackMutex.RLock()
	defer fc.feedbackMutex.RUnlock()

	var feedbackList []*Feedback
	for _, feedback := range fc.feedback {
		if fc.matchesFilters(feedback, filters) {
			feedbackList = append(feedbackList, feedback)
		}
	}

	return feedbackList, nil
}

// UpdateFeedback updates feedback status
func (fc *FeedbackCollector) UpdateFeedback(ctx context.Context, feedbackID string, updates map[string]interface{}) error {
	fc.feedbackMutex.Lock()
	defer fc.feedbackMutex.Unlock()

	feedback, exists := fc.feedback[feedbackID]
	if !exists {
		return fmt.Errorf("feedback not found: %s", feedbackID)
	}

	// Apply updates
	if status, ok := updates["status"].(FeedbackStatus); ok {
		feedback.Status = status
	}
	if feedbackText, ok := updates["feedback_text"].(string); ok {
		feedback.FeedbackText = feedbackText
	}
	if confidence, ok := updates["confidence"].(float64); ok {
		feedback.Confidence = confidence
	}
	if metadata, ok := updates["metadata"].(map[string]interface{}); ok {
		for k, v := range metadata {
			feedback.Metadata[k] = v
		}
	}

	// Update processed timestamp if status changed to processed
	if feedback.Status == FeedbackStatusProcessed && feedback.ProcessedAt == nil {
		now := time.Now()
		feedback.ProcessedAt = &now
	}

	// Log feedback update
	if fc.logger != nil {
		fc.logger.WithComponent("feedback_collector").LogBusinessEvent(ctx, "feedback_updated", feedbackID, map[string]interface{}{
			"updates_applied": len(updates),
		})
	}

	return nil
}

// GetAccuracyMetrics returns accuracy metrics based on feedback
func (fc *FeedbackCollector) GetAccuracyMetrics(ctx context.Context, filters map[string]interface{}) (*FeedbackAccuracyMetrics, error) {
	fc.accuracyMutex.RLock()
	defer fc.accuracyMutex.RUnlock()

	// Get metrics for the specified filters
	key := fc.generateMetricsKey(filters)
	metrics, exists := fc.accuracyMetrics[key]
	if !exists {
		// Calculate metrics if not cached
		metrics = fc.calculateAccuracyMetrics(filters)
		fc.accuracyMetrics[key] = metrics
	}

	return metrics, nil
}

// GetModelUpdates returns model updates based on feedback
func (fc *FeedbackCollector) GetModelUpdates(ctx context.Context) (map[string]interface{}, error) {
	fc.modelUpdatesMutex.RLock()
	defer fc.modelUpdatesMutex.RUnlock()

	updates := make(map[string]interface{})
	for k, v := range fc.modelUpdates {
		updates[k] = v
	}

	return updates, nil
}

// GetCollectorStats returns statistics about the feedback collector
func (fc *FeedbackCollector) GetCollectorStats() map[string]interface{} {
	fc.feedbackMutex.RLock()
	defer fc.feedbackMutex.RUnlock()

	stats := map[string]interface{}{
		"total_feedback":          len(fc.feedback),
		"validation_rules":        len(fc.validationRules),
		"accuracy_metrics":        len(fc.accuracyMetrics),
		"model_updates":           len(fc.modelUpdates),
		"feedback_collection":     fc.enableFeedbackCollection,
		"feedback_retention_days": fc.feedbackRetentionDays,
		"max_feedback_per_user":   fc.maxFeedbackPerUser,
		"status_breakdown":        make(map[string]int),
		"type_breakdown":          make(map[string]int),
	}

	// Calculate breakdowns
	for _, feedback := range fc.feedback {
		stats["status_breakdown"].(map[string]int)[string(feedback.Status)]++
		stats["type_breakdown"].(map[string]int)[string(feedback.FeedbackType)]++
	}

	return stats
}

// Helper methods

// validateFeedback validates feedback before processing
func (fc *FeedbackCollector) validateFeedback(feedback *Feedback) error {
	// Basic validation
	if feedback.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if feedback.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}
	if feedback.OriginalClassification == nil {
		return fmt.Errorf("original classification is required")
	}
	if feedback.FeedbackType == "" {
		return fmt.Errorf("feedback type is required")
	}

	// Apply validation rules
	for _, rule := range fc.validationRules {
		if !rule.Enabled {
			continue
		}

		if err := fc.evaluateValidationRule(rule, feedback); err != nil {
			return fmt.Errorf("validation rule %s failed: %w", rule.RuleID, err)
		}
	}

	return nil
}

// evaluateValidationRule evaluates a validation rule
func (fc *FeedbackCollector) evaluateValidationRule(rule FeedbackValidationRule, feedback *Feedback) error {
	switch rule.RuleType {
	case "format":
		return fc.evaluateFormatRule(rule, feedback)
	case "content":
		return fc.evaluateContentRule(rule, feedback)
	case "business_logic":
		return fc.evaluateBusinessLogicRule(rule, feedback)
	default:
		return nil
	}
}

// evaluateFormatRule evaluates a format validation rule
func (fc *FeedbackCollector) evaluateFormatRule(rule FeedbackValidationRule, feedback *Feedback) error {
	// Check feedback text length
	if rule.Condition == "text_length" {
		maxLength := 1000
		if param, ok := rule.Parameters["max_length"].(int); ok {
			maxLength = param
		}
		if len(feedback.FeedbackText) > maxLength {
			return fmt.Errorf("feedback text exceeds maximum length of %d", maxLength)
		}
	}

	// Check confidence range
	if rule.Condition == "confidence_range" {
		if feedback.Confidence < 0.0 || feedback.Confidence > 1.0 {
			return fmt.Errorf("confidence must be between 0.0 and 1.0")
		}
	}

	return nil
}

// evaluateContentRule evaluates a content validation rule
func (fc *FeedbackCollector) evaluateContentRule(rule FeedbackValidationRule, feedback *Feedback) error {
	// Check for inappropriate content
	if rule.Condition == "inappropriate_content" {
		inappropriateWords := []string{"spam", "test", "invalid"}
		if param, ok := rule.Parameters["words"].([]string); ok {
			inappropriateWords = param
		}

		textLower := feedback.FeedbackText
		for _, word := range inappropriateWords {
			if strings.Contains(textLower, word) {
				return fmt.Errorf("feedback contains inappropriate content")
			}
		}
	}

	return nil
}

// evaluateBusinessLogicRule evaluates a business logic validation rule
func (fc *FeedbackCollector) evaluateBusinessLogicRule(rule FeedbackValidationRule, feedback *Feedback) error {
	// Check user feedback limit
	if rule.Condition == "user_feedback_limit" {
		userFeedbackCount := fc.getUserFeedbackCount(feedback.UserID)
		if userFeedbackCount >= fc.maxFeedbackPerUser {
			return fmt.Errorf("user has reached maximum feedback limit")
		}
	}

	return nil
}

// processFeedback processes feedback asynchronously
func (fc *FeedbackCollector) processFeedback(ctx context.Context, feedback *Feedback) {
	start := time.Now()

	// Update status to processing
	fc.UpdateFeedback(ctx, feedback.ID, map[string]interface{}{
		"status": FeedbackStatusProcessed,
	})

	// Apply feedback to accuracy metrics
	fc.updateAccuracyMetrics(feedback)

	// Generate model updates if applicable
	if fc.shouldGenerateModelUpdate(feedback) {
		fc.generateModelUpdate(feedback)
	}

	// Update feedback processing time
	fc.UpdateFeedback(ctx, feedback.ID, map[string]interface{}{
		"processing_time": time.Since(start),
	})

	// Log feedback processing completion
	if fc.logger != nil {
		fc.logger.WithComponent("feedback_collector").LogBusinessEvent(ctx, "feedback_processing_completed", feedback.ID, map[string]interface{}{
			"processing_time_ms": time.Since(start).Milliseconds(),
		})
	}

	// Record metrics
	fc.RecordFeedbackMetrics(ctx, feedback, "processed")
}

// updateAccuracyMetrics updates accuracy metrics based on feedback
func (fc *FeedbackCollector) updateAccuracyMetrics(feedback *Feedback) {
	fc.accuracyMutex.Lock()
	defer fc.accuracyMutex.Unlock()

	// Get or create metrics for the current time period
	key := fc.generateMetricsKey(nil)
	metrics, exists := fc.accuracyMetrics[key]
	if !exists {
		metrics = &FeedbackAccuracyMetrics{
			IndustryBreakdown: make(map[string]interface{}),
			LastUpdated:       time.Now(),
		}
		fc.accuracyMetrics[key] = metrics
	}

	// Update metrics
	metrics.TotalFeedback++
	if fc.isPositiveFeedback(feedback) {
		metrics.PositiveFeedback++
	} else {
		metrics.NegativeFeedback++
	}

	// Recalculate accuracy score
	if metrics.TotalFeedback > 0 {
		metrics.AccuracyScore = float64(metrics.PositiveFeedback) / float64(metrics.TotalFeedback)
	}

	// Update confidence score
	metrics.ConfidenceScore = (metrics.ConfidenceScore + feedback.Confidence) / 2.0

	// Update industry breakdown
	if feedback.OriginalClassification != nil {
		industryCode := feedback.OriginalClassification.IndustryCode
		if industryCode != "" {
			if breakdown, ok := metrics.IndustryBreakdown[industryCode].(map[string]interface{}); ok {
				breakdown["total"] = breakdown["total"].(int) + 1
				if fc.isPositiveFeedback(feedback) {
					breakdown["positive"] = breakdown["positive"].(int) + 1
				} else {
					breakdown["negative"] = breakdown["negative"].(int) + 1
				}
			} else {
				metrics.IndustryBreakdown[industryCode] = map[string]interface{}{
					"total":    1,
					"positive": 0,
					"negative": 0,
				}
				if fc.isPositiveFeedback(feedback) {
					metrics.IndustryBreakdown[industryCode].(map[string]interface{})["positive"] = 1
				} else {
					metrics.IndustryBreakdown[industryCode].(map[string]interface{})["negative"] = 1
				}
			}
		}
	}

	metrics.LastUpdated = time.Now()
}

// isPositiveFeedback determines if feedback is positive
func (fc *FeedbackCollector) isPositiveFeedback(feedback *Feedback) bool {
	switch feedback.FeedbackType {
	case FeedbackTypeAccuracy:
		if value, ok := feedback.FeedbackValue.(bool); ok {
			return value
		}
		if value, ok := feedback.FeedbackValue.(float64); ok {
			return value >= 0.7
		}
	case FeedbackTypeConfidence:
		if value, ok := feedback.FeedbackValue.(float64); ok {
			return value >= 0.7
		}
	case FeedbackTypeRelevance:
		if value, ok := feedback.FeedbackValue.(bool); ok {
			return value
		}
	}
	return false
}

// shouldGenerateModelUpdate determines if feedback should trigger model updates
func (fc *FeedbackCollector) shouldGenerateModelUpdate(feedback *Feedback) bool {
	// Generate updates for negative feedback with high confidence
	if !fc.isPositiveFeedback(feedback) && feedback.Confidence >= 0.8 {
		return true
	}

	// Generate updates for correction feedback
	if feedback.FeedbackType == FeedbackTypeCorrection {
		return true
	}

	return false
}

// generateModelUpdate generates model updates based on feedback
func (fc *FeedbackCollector) generateModelUpdate(feedback *Feedback) {
	fc.modelUpdatesMutex.Lock()
	defer fc.modelUpdatesMutex.Unlock()

	updateKey := fmt.Sprintf("update_%s_%s", feedback.ID, time.Now().Format("20060102_150405"))
	update := map[string]interface{}{
		"feedback_id":    feedback.ID,
		"business_name":  feedback.BusinessName,
		"feedback_type":  string(feedback.FeedbackType),
		"original_code":  feedback.OriginalClassification.IndustryCode,
		"suggested_code": "",
		"confidence":     feedback.Confidence,
		"created_at":     time.Now(),
	}

	// Add suggested classification if available
	if feedback.SuggestedClassification != nil {
		update["suggested_code"] = feedback.SuggestedClassification.IndustryCode
	}

	fc.modelUpdates[updateKey] = update
}

// getUserFeedbackCount gets the number of feedback submissions by a user
func (fc *FeedbackCollector) getUserFeedbackCount(userID string) int {
	fc.feedbackMutex.RLock()
	defer fc.feedbackMutex.RUnlock()

	count := 0
	for _, feedback := range fc.feedback {
		if feedback.UserID == userID {
			count++
		}
	}
	return count
}

// matchesFilters checks if feedback matches the specified filters
func (fc *FeedbackCollector) matchesFilters(feedback *Feedback, filters map[string]interface{}) bool {
	if filters == nil {
		return true
	}

	for key, value := range filters {
		switch key {
		case "user_id":
			if userID, ok := value.(string); ok && feedback.UserID != userID {
				return false
			}
		case "feedback_type":
			if feedbackType, ok := value.(FeedbackType); ok && feedback.FeedbackType != feedbackType {
				return false
			}
		case "status":
			if status, ok := value.(FeedbackStatus); ok && feedback.Status != status {
				return false
			}
		case "business_name":
			if businessName, ok := value.(string); ok && feedback.BusinessName != businessName {
				return false
			}
		}
	}

	return true
}

// generateFeedbackID generates a unique feedback ID
func (fc *FeedbackCollector) generateFeedbackID() string {
	return fmt.Sprintf("feedback_%d", time.Now().UnixNano())
}

// generateMetricsKey generates a key for metrics caching
func (fc *FeedbackCollector) generateMetricsKey(filters map[string]interface{}) string {
	if filters == nil {
		return "default"
	}
	// Simple key generation - in production, use a more sophisticated approach
	return fmt.Sprintf("metrics_%d", time.Now().Unix()/3600) // Hourly buckets
}

// calculateAccuracyMetrics calculates accuracy metrics for the specified filters
func (fc *FeedbackCollector) calculateAccuracyMetrics(filters map[string]interface{}) *FeedbackAccuracyMetrics {
	fc.feedbackMutex.RLock()
	defer fc.feedbackMutex.RUnlock()

	metrics := &FeedbackAccuracyMetrics{
		IndustryBreakdown: make(map[string]interface{}),
		LastUpdated:       time.Now(),
	}

	for _, feedback := range fc.feedback {
		if fc.matchesFilters(feedback, filters) {
			metrics.TotalFeedback++
			if fc.isPositiveFeedback(feedback) {
				metrics.PositiveFeedback++
			} else {
				metrics.NegativeFeedback++
			}
		}
	}

	if metrics.TotalFeedback > 0 {
		metrics.AccuracyScore = float64(metrics.PositiveFeedback) / float64(metrics.TotalFeedback)
	}

	return metrics
}

// startAccuracyUpdateLoop starts the accuracy metrics update loop
func (fc *FeedbackCollector) startAccuracyUpdateLoop() {
	ticker := time.NewTicker(fc.accuracyUpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		// Recalculate all accuracy metrics
		fc.accuracyMutex.Lock()
		for key := range fc.accuracyMetrics {
			delete(fc.accuracyMetrics, key)
		}
		fc.accuracyMutex.Unlock()

		// Clean up old feedback
		fc.cleanupOldFeedback()
	}
}

// cleanupOldFeedback removes feedback older than retention period
func (fc *FeedbackCollector) cleanupOldFeedback() {
	fc.feedbackMutex.Lock()
	defer fc.feedbackMutex.Unlock()

	cutoff := time.Now().AddDate(0, 0, -fc.feedbackRetentionDays)
	for id, feedback := range fc.feedback {
		if feedback.CreatedAt.Before(cutoff) {
			delete(fc.feedback, id)
		}
	}
}

// initializeValidationRules initializes validation rules
func (fc *FeedbackCollector) initializeValidationRules() {
	fc.validationRules = map[string]FeedbackValidationRule{
		"text_length": {
			RuleID:    "text_length",
			RuleType:  "format",
			Condition: "text_length",
			Severity:  "medium",
			Weight:    0.3,
			Enabled:   true,
			Parameters: map[string]interface{}{
				"max_length": 1000,
			},
		},
		"confidence_range": {
			RuleID:    "confidence_range",
			RuleType:  "format",
			Condition: "confidence_range",
			Severity:  "high",
			Weight:    0.5,
			Enabled:   true,
		},
		"inappropriate_content": {
			RuleID:    "inappropriate_content",
			RuleType:  "content",
			Condition: "inappropriate_content",
			Severity:  "critical",
			Weight:    1.0,
			Enabled:   true,
			Parameters: map[string]interface{}{
				"words": []string{"spam", "test", "invalid"},
			},
		},
		"user_feedback_limit": {
			RuleID:    "user_feedback_limit",
			RuleType:  "business_logic",
			Condition: "user_feedback_limit",
			Severity:  "medium",
			Weight:    0.4,
			Enabled:   true,
		},
	}
}

// RecordFeedbackMetrics records metrics for feedback operations
func (fc *FeedbackCollector) RecordFeedbackMetrics(ctx context.Context, feedback *Feedback, operation string) {
	if fc.metrics == nil {
		return
	}

	fc.metrics.RecordHistogram(ctx, "feedback_confidence", feedback.Confidence, map[string]string{
		"operation":     operation,
		"feedback_type": string(feedback.FeedbackType),
	})

	fc.metrics.RecordHistogram(ctx, "feedback_operations", 1.0, map[string]string{
		"operation":     operation,
		"feedback_type": string(feedback.FeedbackType),
		"status":        string(feedback.Status),
	})
}
