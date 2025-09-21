package feedback

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// UserFeedback represents structured user feedback data
type UserFeedback struct {
	ID                     uuid.UUID              `json:"id"`
	UserID                 string                 `json:"user_id"`
	Category               FeedbackCategory       `json:"category"`
	Rating                 int                    `json:"rating"` // 1-5 scale
	Comments               string                 `json:"comments"`
	SpecificFeatures       []string               `json:"specific_features"`
	ImprovementAreas       []string               `json:"improvement_areas"`
	ClassificationAccuracy float64                `json:"classification_accuracy"`
	PerformanceRating      int                    `json:"performance_rating"` // 1-5 scale
	UsabilityRating        int                    `json:"usability_rating"`   // 1-5 scale
	BusinessImpact         BusinessImpactRating   `json:"business_impact"`
	SubmittedAt            time.Time              `json:"submitted_at"`
	Metadata               map[string]interface{} `json:"metadata"`
}

// FeedbackCategory represents different types of feedback
type FeedbackCategory string

const (
	CategoryDatabasePerformance    FeedbackCategory = "database_performance"
	CategoryClassificationAccuracy FeedbackCategory = "classification_accuracy"
	CategoryUserExperience         FeedbackCategory = "user_experience"
	CategoryRiskDetection          FeedbackCategory = "risk_detection"
	CategoryOverallSatisfaction    FeedbackCategory = "overall_satisfaction"
	CategoryFeatureRequest         FeedbackCategory = "feature_request"
	CategoryBugReport              FeedbackCategory = "bug_report"
)

// BusinessImpactRating represents the business impact assessment
type BusinessImpactRating struct {
	TimeSaved        int    `json:"time_saved_minutes"`
	CostReduction    string `json:"cost_reduction"`
	ErrorReduction   int    `json:"error_reduction_percentage"`
	ProductivityGain int    `json:"productivity_gain_percentage"`
	ROI              string `json:"roi_assessment"`
}

// FeedbackStats provides aggregated feedback statistics
type FeedbackStats struct {
	TotalResponses     int                  `json:"total_responses"`
	AverageRating      float64              `json:"average_rating"`
	CategoryBreakdown  map[string]int       `json:"category_breakdown"`
	RatingDistribution map[int]int          `json:"rating_distribution"`
	CommonImprovements []string             `json:"common_improvements"`
	BusinessImpactAvg  BusinessImpactRating `json:"business_impact_average"`
	ResponseRate       float64              `json:"response_rate"`
	LastUpdated        time.Time            `json:"last_updated"`
}

// FeedbackAnalysis provides detailed analysis of feedback data
type FeedbackAnalysis struct {
	Category           FeedbackCategory     `json:"category"`
	TotalResponses     int                  `json:"total_responses"`
	AverageRating      float64              `json:"average_rating"`
	AveragePerformance float64              `json:"average_performance"`
	AverageUsability   float64              `json:"average_usability"`
	AverageAccuracy    float64              `json:"average_accuracy"`
	TopImprovements    []string             `json:"top_improvements"`
	TopFeatures        []string             `json:"top_features"`
	BusinessImpact     BusinessImpactRating `json:"business_impact"`
	SentimentScore     float64              `json:"sentiment_score"`
	Recommendations    []string             `json:"recommendations"`
	GeneratedAt        time.Time            `json:"generated_at"`
}

// FeedbackStorage interface for storing and retrieving feedback data
type FeedbackStorage interface {
	StoreFeedback(ctx context.Context, feedback *UserFeedback) error
	GetFeedbackByCategory(ctx context.Context, category string) ([]*UserFeedback, error)
	GetFeedbackByTimeRange(ctx context.Context, start, end time.Time) ([]*UserFeedback, error)
	GetFeedbackStats(ctx context.Context) (*FeedbackStats, error)
}
