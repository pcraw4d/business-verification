package webanalysis

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// BetaFeedbackCollector manages user feedback collection and storage
type BetaFeedbackCollector struct {
	logger       *zap.Logger
	mu           sync.RWMutex
	feedback     map[string]*BetaFeedback // feedbackID -> feedback
	userFeedback map[string][]string      // userID -> feedbackIDs
	testFeedback map[string][]string      // testID -> feedbackIDs
}

// NewBetaFeedbackCollector creates a new beta feedback collector
func NewBetaFeedbackCollector(logger *zap.Logger) *BetaFeedbackCollector {
	return &BetaFeedbackCollector{
		logger:       logger,
		feedback:     make(map[string]*BetaFeedback),
		userFeedback: make(map[string][]string),
		testFeedback: make(map[string][]string),
	}
}

// StoreFeedback stores user feedback
func (bfc *BetaFeedbackCollector) StoreFeedback(feedback *BetaFeedback) error {
	bfc.mu.Lock()
	defer bfc.mu.Unlock()

	if feedback == nil {
		return fmt.Errorf("cannot store nil feedback")
	}

	// Generate feedback ID if not provided
	if feedback.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	feedbackID := bfc.generateFeedbackID(feedback)

	// Store feedback
	bfc.feedback[feedbackID] = feedback

	// Update user feedback index
	if bfc.userFeedback[feedback.UserID] == nil {
		bfc.userFeedback[feedback.UserID] = make([]string, 0)
	}
	bfc.userFeedback[feedback.UserID] = append(bfc.userFeedback[feedback.UserID], feedbackID)

	// Update test feedback index
	if bfc.testFeedback[feedback.TestID] == nil {
		bfc.testFeedback[feedback.TestID] = make([]string, 0)
	}
	bfc.testFeedback[feedback.TestID] = append(bfc.testFeedback[feedback.TestID], feedbackID)

	bfc.logger.Info("Stored beta feedback",
		zap.String("feedback_id", feedbackID),
		zap.String("user_id", feedback.UserID),
		zap.String("test_id", feedback.TestID),
		zap.String("method", feedback.Method),
		zap.Int("satisfaction", feedback.Satisfaction),
		zap.Int("accuracy", feedback.Accuracy),
		zap.Int("speed", feedback.Speed),
	)

	return nil
}

// GetFeedback retrieves feedback by ID
func (bfc *BetaFeedbackCollector) GetFeedback(feedbackID string) (*BetaFeedback, error) {
	bfc.mu.RLock()
	defer bfc.mu.RUnlock()

	feedback, exists := bfc.feedback[feedbackID]
	if !exists {
		return nil, fmt.Errorf("feedback not found: %s", feedbackID)
	}

	return feedback, nil
}

// GetUserFeedback retrieves all feedback for a user
func (bfc *BetaFeedbackCollector) GetUserFeedback(userID string) ([]*BetaFeedback, error) {
	bfc.mu.RLock()
	defer bfc.mu.RUnlock()

	feedbackIDs, exists := bfc.userFeedback[userID]
	if !exists {
		return []*BetaFeedback{}, nil
	}

	var feedback []*BetaFeedback
	for _, feedbackID := range feedbackIDs {
		if f, exists := bfc.feedback[feedbackID]; exists {
			feedback = append(feedback, f)
		}
	}

	return feedback, nil
}

// GetTestFeedback retrieves all feedback for a test
func (bfc *BetaFeedbackCollector) GetTestFeedback(testID string) ([]*BetaFeedback, error) {
	bfc.mu.RLock()
	defer bfc.mu.RUnlock()

	feedbackIDs, exists := bfc.testFeedback[testID]
	if !exists {
		return []*BetaFeedback{}, nil
	}

	var feedback []*BetaFeedback
	for _, feedbackID := range feedbackIDs {
		if f, exists := bfc.feedback[feedbackID]; exists {
			feedback = append(feedback, f)
		}
	}

	return feedback, nil
}

// GetFeedbackSummary returns a summary of feedback statistics
func (bfc *BetaFeedbackCollector) GetFeedbackSummary(ctx context.Context) (*FeedbackSummary, error) {
	bfc.mu.RLock()
	defer bfc.mu.RUnlock()

	summary := &FeedbackSummary{
		Generated: time.Now(),
	}

	// Calculate overall statistics
	var totalSatisfaction, totalAccuracy, totalSpeed int
	var satisfactionCount, accuracyCount, speedCount int

	for _, feedback := range bfc.feedback {
		if feedback.Satisfaction > 0 {
			totalSatisfaction += feedback.Satisfaction
			satisfactionCount++
		}
		if feedback.Accuracy > 0 {
			totalAccuracy += feedback.Accuracy
			accuracyCount++
		}
		if feedback.Speed > 0 {
			totalSpeed += feedback.Speed
			speedCount++
		}
	}

	summary.TotalFeedback = len(bfc.feedback)
	summary.UniqueUsers = len(bfc.userFeedback)
	summary.UniqueTests = len(bfc.testFeedback)

	if satisfactionCount > 0 {
		summary.AverageSatisfaction = float64(totalSatisfaction) / float64(satisfactionCount)
	}
	if accuracyCount > 0 {
		summary.AverageAccuracy = float64(totalAccuracy) / float64(accuracyCount)
	}
	if speedCount > 0 {
		summary.AverageSpeed = float64(totalSpeed) / float64(speedCount)
	}

	// Calculate method-specific statistics
	summary.MethodStats = bfc.calculateMethodStats()

	// Calculate satisfaction distribution
	summary.SatisfactionDistribution = bfc.calculateSatisfactionDistribution()

	return summary, nil
}

// GetMethodFeedback retrieves feedback for a specific scraping method
func (bfc *BetaFeedbackCollector) GetMethodFeedback(method string) ([]*BetaFeedback, error) {
	bfc.mu.RLock()
	defer bfc.mu.RUnlock()

	var feedback []*BetaFeedback
	for _, f := range bfc.feedback {
		if f.Method == method {
			feedback = append(feedback, f)
		}
	}

	return feedback, nil
}

// GetFeedbackByTimeRange retrieves feedback within a time range
func (bfc *BetaFeedbackCollector) GetFeedbackByTimeRange(start, end time.Time) ([]*BetaFeedback, error) {
	bfc.mu.RLock()
	defer bfc.mu.RUnlock()

	var feedback []*BetaFeedback
	for _, f := range bfc.feedback {
		if f.Timestamp.After(start) && f.Timestamp.Before(end) {
			feedback = append(feedback, f)
		}
	}

	return feedback, nil
}

// DeleteFeedback deletes feedback by ID
func (bfc *BetaFeedbackCollector) DeleteFeedback(feedbackID string) error {
	bfc.mu.Lock()
	defer bfc.mu.Unlock()

	feedback, exists := bfc.feedback[feedbackID]
	if !exists {
		return fmt.Errorf("feedback not found: %s", feedbackID)
	}

	// Remove from main feedback map
	delete(bfc.feedback, feedbackID)

	// Remove from user feedback index
	if userFeedback, exists := bfc.userFeedback[feedback.UserID]; exists {
		bfc.userFeedback[feedback.UserID] = bfc.removeFromSlice(userFeedback, feedbackID)
	}

	// Remove from test feedback index
	if testFeedback, exists := bfc.testFeedback[feedback.TestID]; exists {
		bfc.testFeedback[feedback.TestID] = bfc.removeFromSlice(testFeedback, feedbackID)
	}

	bfc.logger.Info("Deleted beta feedback",
		zap.String("feedback_id", feedbackID),
		zap.String("user_id", feedback.UserID),
		zap.String("test_id", feedback.TestID),
	)

	return nil
}

// ExportFeedback exports feedback to JSON format
func (bfc *BetaFeedbackCollector) ExportFeedback() ([]byte, error) {
	bfc.mu.RLock()
	defer bfc.mu.RUnlock()

	var feedback []*BetaFeedback
	for _, f := range bfc.feedback {
		feedback = append(feedback, f)
	}

	return json.MarshalIndent(feedback, "", "  ")
}

// ExportFeedbackSummary exports feedback summary to JSON format
func (bfc *BetaFeedbackCollector) ExportFeedbackSummary(ctx context.Context) ([]byte, error) {
	summary, err := bfc.GetFeedbackSummary(ctx)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(summary, "", "  ")
}

// generateFeedbackID generates a unique feedback ID
func (bfc *BetaFeedbackCollector) generateFeedbackID(feedback *BetaFeedback) string {
	return fmt.Sprintf("feedback_%s_%s_%d",
		feedback.UserID,
		feedback.TestID,
		feedback.Timestamp.UnixNano())
}

// calculateMethodStats calculates statistics for each scraping method
func (bfc *BetaFeedbackCollector) calculateMethodStats() map[string]*MethodStats {
	methodStats := make(map[string]*MethodStats)

	for _, feedback := range bfc.feedback {
		if methodStats[feedback.Method] == nil {
			methodStats[feedback.Method] = &MethodStats{
				Method: feedback.Method,
			}
		}

		stats := methodStats[feedback.Method]
		stats.TotalFeedback++

		if feedback.Satisfaction > 0 {
			stats.TotalSatisfaction += feedback.Satisfaction
			stats.SatisfactionCount++
		}
		if feedback.Accuracy > 0 {
			stats.TotalAccuracy += feedback.Accuracy
			stats.AccuracyCount++
		}
		if feedback.Speed > 0 {
			stats.TotalSpeed += feedback.Speed
			stats.SpeedCount++
		}
	}

	// Calculate averages
	for _, stats := range methodStats {
		if stats.SatisfactionCount > 0 {
			stats.AverageSatisfaction = float64(stats.TotalSatisfaction) / float64(stats.SatisfactionCount)
		}
		if stats.AccuracyCount > 0 {
			stats.AverageAccuracy = float64(stats.TotalAccuracy) / float64(stats.AccuracyCount)
		}
		if stats.SpeedCount > 0 {
			stats.AverageSpeed = float64(stats.TotalSpeed) / float64(stats.SpeedCount)
		}
	}

	return methodStats
}

// calculateSatisfactionDistribution calculates distribution of satisfaction scores
func (bfc *BetaFeedbackCollector) calculateSatisfactionDistribution() map[int]int {
	distribution := make(map[int]int)

	for _, feedback := range bfc.feedback {
		if feedback.Satisfaction > 0 {
			distribution[feedback.Satisfaction]++
		}
	}

	return distribution
}

// removeFromSlice removes an item from a slice
func (bfc *BetaFeedbackCollector) removeFromSlice(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// FeedbackSummary represents a summary of feedback statistics
type FeedbackSummary struct {
	TotalFeedback            int                     `json:"total_feedback"`
	UniqueUsers              int                     `json:"unique_users"`
	UniqueTests              int                     `json:"unique_tests"`
	AverageSatisfaction      float64                 `json:"average_satisfaction"`
	AverageAccuracy          float64                 `json:"average_accuracy"`
	AverageSpeed             float64                 `json:"average_speed"`
	MethodStats              map[string]*MethodStats `json:"method_stats"`
	SatisfactionDistribution map[int]int             `json:"satisfaction_distribution"`
	Generated                time.Time               `json:"generated"`
}

// MethodStats represents statistics for a specific scraping method
type MethodStats struct {
	Method              string  `json:"method"`
	TotalFeedback       int     `json:"total_feedback"`
	TotalSatisfaction   int     `json:"total_satisfaction"`
	SatisfactionCount   int     `json:"satisfaction_count"`
	AverageSatisfaction float64 `json:"average_satisfaction"`
	TotalAccuracy       int     `json:"total_accuracy"`
	AccuracyCount       int     `json:"accuracy_count"`
	AverageAccuracy     float64 `json:"average_accuracy"`
	TotalSpeed          int     `json:"total_speed"`
	SpeedCount          int     `json:"speed_count"`
	AverageSpeed        float64 `json:"average_speed"`
}
