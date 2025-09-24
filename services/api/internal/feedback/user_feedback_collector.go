package feedback

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// UserFeedbackCollector handles collection and processing of user feedback
// on database improvements and classification system enhancements
type UserFeedbackCollector struct {
	storage FeedbackStorage
	logger  *log.Logger
}

// NewUserFeedbackCollector creates a new user feedback collector
func NewUserFeedbackCollector(storage FeedbackStorage, logger *log.Logger) *UserFeedbackCollector {
	return &UserFeedbackCollector{
		storage: storage,
		logger:  logger,
	}
}

// CollectFeedback processes and stores user feedback
func (ufc *UserFeedbackCollector) CollectFeedback(ctx context.Context, feedback *UserFeedback) error {
	// Validate feedback data
	if err := ufc.validateFeedback(feedback); err != nil {
		return fmt.Errorf("feedback validation failed: %w", err)
	}

	// Set metadata
	feedback.ID = uuid.New()
	feedback.SubmittedAt = time.Now()

	// Add system metadata
	feedback.Metadata = map[string]interface{}{
		"collector_version": "1.0.0",
		"collection_method": "api",
		"timestamp":         feedback.SubmittedAt,
	}

	// Store feedback
	if err := ufc.storage.StoreFeedback(ctx, feedback); err != nil {
		return fmt.Errorf("failed to store feedback: %w", err)
	}

	ufc.logger.Printf("User feedback collected successfully: ID=%s, Category=%s, Rating=%d",
		feedback.ID, feedback.Category, feedback.Rating)

	return nil
}

// validateFeedback validates the feedback data
func (ufc *UserFeedbackCollector) validateFeedback(feedback *UserFeedback) error {
	if feedback.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if feedback.Category == "" {
		return fmt.Errorf("feedback category is required")
	}

	if feedback.Rating < 1 || feedback.Rating > 5 {
		return fmt.Errorf("rating must be between 1 and 5")
	}

	if feedback.PerformanceRating < 1 || feedback.PerformanceRating > 5 {
		return fmt.Errorf("performance rating must be between 1 and 5")
	}

	if feedback.UsabilityRating < 1 || feedback.UsabilityRating > 5 {
		return fmt.Errorf("usability rating must be between 1 and 5")
	}

	if feedback.ClassificationAccuracy < 0 || feedback.ClassificationAccuracy > 1 {
		return fmt.Errorf("classification accuracy must be between 0 and 1")
	}

	return nil
}

// GetFeedbackAnalysis retrieves and analyzes feedback data
func (ufc *UserFeedbackCollector) GetFeedbackAnalysis(ctx context.Context, category FeedbackCategory) (*FeedbackAnalysis, error) {
	// Get feedback for specific category
	feedback, err := ufc.storage.GetFeedbackByCategory(ctx, string(category))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve feedback: %w", err)
	}

	// Analyze feedback
	analysis := ufc.analyzeFeedback(feedback)

	return analysis, nil
}

// analyzeFeedback performs comprehensive analysis of feedback data
func (ufc *UserFeedbackCollector) analyzeFeedback(feedback []*UserFeedback) *FeedbackAnalysis {
	if len(feedback) == 0 {
		return &FeedbackAnalysis{
			GeneratedAt: time.Now(),
		}
	}

	analysis := &FeedbackAnalysis{
		Category:       feedback[0].Category,
		TotalResponses: len(feedback),
		GeneratedAt:    time.Now(),
	}

	// Calculate averages
	var totalRating, totalPerformance, totalUsability, totalAccuracy float64
	improvementCounts := make(map[string]int)
	featureCounts := make(map[string]int)
	var totalTimeSaved, totalErrorReduction, totalProductivity int

	for _, f := range feedback {
		totalRating += float64(f.Rating)
		totalPerformance += float64(f.PerformanceRating)
		totalUsability += float64(f.UsabilityRating)
		totalAccuracy += f.ClassificationAccuracy

		// Count improvements
		for _, improvement := range f.ImprovementAreas {
			improvementCounts[improvement]++
		}

		// Count features
		for _, feature := range f.SpecificFeatures {
			featureCounts[feature]++
		}

		// Aggregate business impact
		totalTimeSaved += f.BusinessImpact.TimeSaved
		totalErrorReduction += f.BusinessImpact.ErrorReduction
		totalProductivity += f.BusinessImpact.ProductivityGain
	}

	// Calculate averages
	count := float64(len(feedback))
	analysis.AverageRating = totalRating / count
	analysis.AveragePerformance = totalPerformance / count
	analysis.AverageUsability = totalUsability / count
	analysis.AverageAccuracy = totalAccuracy / count

	// Calculate business impact averages
	analysis.BusinessImpact = BusinessImpactRating{
		TimeSaved:        totalTimeSaved / len(feedback),
		ErrorReduction:   totalErrorReduction / len(feedback),
		ProductivityGain: totalProductivity / len(feedback),
	}

	// Find top improvements and features
	analysis.TopImprovements = ufc.getTopItems(improvementCounts, 5)
	analysis.TopFeatures = ufc.getTopItems(featureCounts, 5)

	// Calculate sentiment score (simplified)
	analysis.SentimentScore = ufc.calculateSentimentScore(feedback)

	// Generate recommendations
	analysis.Recommendations = ufc.generateRecommendations(analysis)

	return analysis
}

// getTopItems returns the top N items by count
func (ufc *UserFeedbackCollector) getTopItems(counts map[string]int, n int) []string {
	var items []string
	for item, count := range counts {
		items = append(items, fmt.Sprintf("%s (%d)", item, count))
	}

	// Sort by count (simplified - in production, use proper sorting)
	if len(items) > n {
		items = items[:n]
	}

	return items
}

// calculateSentimentScore calculates a simple sentiment score based on ratings
func (ufc *UserFeedbackCollector) calculateSentimentScore(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 0.0
	}

	var totalScore float64
	for _, f := range feedback {
		// Weighted average of ratings
		score := (float64(f.Rating) + float64(f.PerformanceRating) + float64(f.UsabilityRating)) / 3.0
		totalScore += score
	}

	return totalScore / float64(len(feedback))
}

// generateRecommendations generates actionable recommendations based on analysis
func (ufc *UserFeedbackCollector) generateRecommendations(analysis *FeedbackAnalysis) []string {
	var recommendations []string

	// Performance recommendations
	if analysis.AveragePerformance < 3.0 {
		recommendations = append(recommendations, "Focus on improving system performance and response times")
	}

	// Usability recommendations
	if analysis.AverageUsability < 3.0 {
		recommendations = append(recommendations, "Enhance user interface and user experience design")
	}

	// Accuracy recommendations
	if analysis.AverageAccuracy < 0.8 {
		recommendations = append(recommendations, "Improve classification accuracy through model refinement")
	}

	// Business impact recommendations
	if analysis.BusinessImpact.TimeSaved < 30 {
		recommendations = append(recommendations, "Optimize workflows to increase time savings")
	}

	// General recommendations
	if analysis.AverageRating < 3.0 {
		recommendations = append(recommendations, "Address user concerns and implement requested improvements")
	}

	return recommendations
}

// GetFeedbackStats retrieves overall feedback statistics
func (ufc *UserFeedbackCollector) GetFeedbackStats(ctx context.Context) (*FeedbackStats, error) {
	return ufc.storage.GetFeedbackStats(ctx)
}

// GetFeedbackByTimeRange retrieves feedback within a time range
func (ufc *UserFeedbackCollector) GetFeedbackByTimeRange(ctx context.Context, start, end time.Time) ([]*UserFeedback, error) {
	return ufc.storage.GetFeedbackByTimeRange(ctx, start, end)
}

// ExportFeedback exports feedback data in various formats
func (ufc *UserFeedbackCollector) ExportFeedback(ctx context.Context, format string, category FeedbackCategory) ([]byte, error) {
	feedback, err := ufc.storage.GetFeedbackByCategory(ctx, string(category))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve feedback for export: %w", err)
	}

	switch format {
	case "json":
		return json.MarshalIndent(feedback, "", "  ")
	case "csv":
		return ufc.exportToCSV(feedback), nil
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportToCSV exports feedback data to CSV format
func (ufc *UserFeedbackCollector) exportToCSV(feedback []*UserFeedback) []byte {
	// Simplified CSV export - in production, use proper CSV library
	var csvData string
	csvData += "ID,UserID,Category,Rating,Comments,PerformanceRating,UsabilityRating,ClassificationAccuracy,SubmittedAt\n"

	for _, f := range feedback {
		csvData += fmt.Sprintf("%s,%s,%s,%d,\"%s\",%d,%d,%.2f,%s\n",
			f.ID, f.UserID, f.Category, f.Rating, f.Comments,
			f.PerformanceRating, f.UsabilityRating, f.ClassificationAccuracy,
			f.SubmittedAt.Format(time.RFC3339))
	}

	return []byte(csvData)
}
