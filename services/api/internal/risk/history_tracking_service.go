package risk

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// RiskStorageInterface defines the interface for risk data storage
type RiskStorageInterface interface {
	StoreRiskAssessment(ctx context.Context, assessment *RiskAssessment) error
	GetRiskAssessment(ctx context.Context, id string) (*RiskAssessment, error)
	GetRiskAssessmentsByBusinessID(ctx context.Context, businessID string, limit, offset int) ([]*RiskAssessment, error)
	UpdateRiskAssessment(ctx context.Context, assessment *RiskAssessment) error
	DeleteRiskAssessment(ctx context.Context, id string) error
}

// RiskHistoryTrackingService provides risk history tracking functionality
type RiskHistoryTrackingService struct {
	storageService RiskStorageInterface
	logger         *zap.Logger
}

// NewRiskHistoryTrackingService creates a new risk history tracking service
func NewRiskHistoryTrackingService(storageService RiskStorageInterface, logger *zap.Logger) *RiskHistoryTrackingService {
	return &RiskHistoryTrackingService{
		storageService: storageService,
		logger:         logger,
	}
}

// RiskHistoryTrackingEntry represents a risk history entry with trend analysis
type RiskHistoryTrackingEntry struct {
	Assessment     *RiskAssessment `json:"assessment"`
	Trend          string          `json:"trend"`           // "improving", "stable", "declining"
	ScoreChange    float64         `json:"score_change"`    // Change from previous assessment
	LevelChange    string          `json:"level_change"`    // "up", "down", "same"
	DaysSinceLast  int             `json:"days_since_last"` // Days since last assessment
	AlertCount     int             `json:"alert_count"`     // Number of active alerts
	Recommendation string          `json:"recommendation"`  // Key recommendation
}

// RiskHistoryResponse represents the response for risk history queries
type RiskHistoryResponse struct {
	BusinessID       string                     `json:"business_id"`
	BusinessName     string                     `json:"business_name"`
	TotalAssessments int                        `json:"total_assessments"`
	CurrentScore     float64                    `json:"current_score"`
	CurrentLevel     string                     `json:"current_level"`
	Trend            string                     `json:"trend"`
	History          []RiskHistoryTrackingEntry `json:"history"`
	Statistics       map[string]interface{}     `json:"statistics"`
	GeneratedAt      time.Time                  `json:"generated_at"`
}

// RiskHistoryQuery represents a query for risk history
type RiskHistoryQuery struct {
	BusinessID    string     `json:"business_id"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Limit         int        `json:"limit,omitempty"`
	Offset        int        `json:"offset,omitempty"`
	IncludeTrends bool       `json:"include_trends,omitempty"`
}

// GetRiskHistory retrieves risk assessment history for a business
func (s *RiskHistoryTrackingService) GetRiskHistory(ctx context.Context, query *RiskHistoryQuery) (*RiskHistoryResponse, error) {
	requestID := "unknown"
	if rid := ctx.Value("request_id"); rid != nil {
		if str, ok := rid.(string); ok {
			requestID = str
		}
	}

	s.logger.Info("Retrieving risk history",
		zap.String("request_id", requestID),
		zap.String("business_id", query.BusinessID),
		zap.Int("limit", query.Limit),
		zap.Int("offset", query.Offset),
	)

	// Set default limit if not provided
	if query.Limit <= 0 {
		query.Limit = 50
	}

	// Get risk assessments from storage service
	assessments, err := s.storageService.GetRiskAssessmentsByBusinessID(ctx, query.BusinessID, query.Limit, query.Offset)
	if err != nil {
		s.logger.Error("Failed to retrieve risk assessments",
			zap.String("request_id", requestID),
			zap.String("business_id", query.BusinessID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to retrieve risk assessments: %w", err)
	}

	if len(assessments) == 0 {
		s.logger.Info("No risk assessments found",
			zap.String("request_id", requestID),
			zap.String("business_id", query.BusinessID),
		)
		return &RiskHistoryResponse{
			BusinessID:       query.BusinessID,
			TotalAssessments: 0,
			History:          []RiskHistoryTrackingEntry{},
			Statistics:       make(map[string]interface{}),
			GeneratedAt:      time.Now(),
		}, nil
	}

	// Create history entries with trend analysis
	historyEntries := s.createHistoryEntries(assessments)

	// Get current assessment (most recent)
	currentAssessment := assessments[0]
	var currentScore float64
	var currentLevel string
	var businessName string
	if currentAssessment != nil {
		currentScore = currentAssessment.OverallScore
		currentLevel = string(currentAssessment.OverallLevel)
		businessName = currentAssessment.BusinessName
	}

	// Determine overall trend
	trend := s.determineOverallTrend(assessments)

	// Calculate statistics
	statistics := s.calculateStatistics(assessments)

	response := &RiskHistoryResponse{
		BusinessID:       query.BusinessID,
		BusinessName:     businessName,
		TotalAssessments: len(assessments),
		CurrentScore:     currentScore,
		CurrentLevel:     currentLevel,
		Trend:            trend,
		History:          historyEntries,
		Statistics:       statistics,
		GeneratedAt:      time.Now(),
	}

	s.logger.Info("Risk history retrieved successfully",
		zap.String("request_id", requestID),
		zap.String("business_id", query.BusinessID),
		zap.Int("total_assessments", len(assessments)),
		zap.String("trend", trend),
	)

	return response, nil
}

// GetRiskTrends analyzes risk trends for a business
func (s *RiskHistoryTrackingService) GetRiskTrends(ctx context.Context, businessID string, days int) (map[string]interface{}, error) {
	requestID := "unknown"
	if rid := ctx.Value("request_id"); rid != nil {
		if str, ok := rid.(string); ok {
			requestID = str
		}
	}

	s.logger.Info("Analyzing risk trends",
		zap.String("request_id", requestID),
		zap.String("business_id", businessID),
		zap.Int("days", days),
	)

	// Get assessments from the last N days
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days)

	// For now, we'll get all assessments and filter by date
	// In a real implementation, you might want to add date filtering to the storage service
	assessments, err := s.storageService.GetRiskAssessmentsByBusinessID(ctx, businessID, 100, 0)
	if err != nil {
		s.logger.Error("Failed to retrieve assessments for trend analysis",
			zap.String("request_id", requestID),
			zap.String("business_id", businessID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to retrieve assessments: %w", err)
	}

	// Filter assessments by date range
	var filteredAssessments []*RiskAssessment
	for _, assessment := range assessments {
		if assessment.AssessedAt.After(startDate) && assessment.AssessedAt.Before(endDate) {
			filteredAssessments = append(filteredAssessments, assessment)
		}
	}

	if len(filteredAssessments) < 2 {
		s.logger.Info("Insufficient data for trend analysis",
			zap.String("request_id", requestID),
			zap.String("business_id", businessID),
			zap.Int("assessments_count", len(filteredAssessments)),
		)
		return map[string]interface{}{
			"trend":             "insufficient_data",
			"assessments_count": len(filteredAssessments),
			"message":           "Insufficient data for trend analysis",
		}, nil
	}

	// Calculate trend metrics
	trends := s.calculateTrendMetrics(filteredAssessments)

	s.logger.Info("Risk trends analyzed successfully",
		zap.String("request_id", requestID),
		zap.String("business_id", businessID),
		zap.Int("assessments_analyzed", len(filteredAssessments)),
	)

	return trends, nil
}

// createHistoryEntries creates history entries with trend analysis
func (s *RiskHistoryTrackingService) createHistoryEntries(assessments []*RiskAssessment) []RiskHistoryTrackingEntry {
	var entries []RiskHistoryTrackingEntry

	for i, assessment := range assessments {
		entry := RiskHistoryTrackingEntry{
			Assessment: assessment,
			AlertCount: len(assessment.Alerts),
		}

		// Calculate trend and changes
		if i > 0 {
			previousAssessment := assessments[i-1]
			entry.ScoreChange = assessment.OverallScore - previousAssessment.OverallScore
			entry.DaysSinceLast = int(assessment.AssessedAt.Sub(previousAssessment.AssessedAt).Hours() / 24)

			// Determine level change
			if assessment.OverallScore > previousAssessment.OverallScore {
				entry.LevelChange = "up"
			} else if assessment.OverallScore < previousAssessment.OverallScore {
				entry.LevelChange = "down"
			} else {
				entry.LevelChange = "same"
			}

			// Determine trend
			if entry.ScoreChange > 5 {
				entry.Trend = "declining"
			} else if entry.ScoreChange < -5 {
				entry.Trend = "improving"
			} else {
				entry.Trend = "stable"
			}
		} else {
			entry.Trend = "baseline"
			entry.LevelChange = "baseline"
			entry.DaysSinceLast = 0
		}

		// Get key recommendation
		if len(assessment.Recommendations) > 0 {
			entry.Recommendation = assessment.Recommendations[0].Title
		}

		entries = append(entries, entry)
	}

	return entries
}

// determineOverallTrend determines the overall trend from assessments
func (s *RiskHistoryTrackingService) determineOverallTrend(assessments []*RiskAssessment) string {
	if len(assessments) < 2 {
		return "insufficient_data"
	}

	// Calculate average score change over the last 3 assessments
	recentCount := 3
	if len(assessments) < recentCount {
		recentCount = len(assessments)
	}

	var totalChange float64
	for i := 0; i < recentCount-1; i++ {
		change := assessments[i].OverallScore - assessments[i+1].OverallScore
		totalChange += change
	}

	averageChange := totalChange / float64(recentCount-1)

	if averageChange > 3 {
		return "declining"
	} else if averageChange < -3 {
		return "improving"
	} else {
		return "stable"
	}
}

// calculateStatistics calculates statistics from assessments
func (s *RiskHistoryTrackingService) calculateStatistics(assessments []*RiskAssessment) map[string]interface{} {
	if len(assessments) == 0 {
		return make(map[string]interface{})
	}

	var totalScore float64
	var minScore, maxScore float64
	levelCounts := make(map[string]int)
	alertCounts := make(map[string]int)

	minScore = assessments[0].OverallScore
	maxScore = assessments[0].OverallScore

	for _, assessment := range assessments {
		totalScore += assessment.OverallScore
		level := string(assessment.OverallLevel)
		levelCounts[level]++

		if assessment.OverallScore < minScore {
			minScore = assessment.OverallScore
		}
		if assessment.OverallScore > maxScore {
			maxScore = assessment.OverallScore
		}

		// Count alerts by level
		for _, alert := range assessment.Alerts {
			alertLevel := string(alert.Level)
			alertCounts[alertLevel]++
		}
	}

	averageScore := totalScore / float64(len(assessments))

	return map[string]interface{}{
		"average_score":      averageScore,
		"min_score":          minScore,
		"max_score":          maxScore,
		"score_range":        maxScore - minScore,
		"level_distribution": levelCounts,
		"alert_distribution": alertCounts,
		"total_alerts":       len(alertCounts),
	}
}

// calculateTrendMetrics calculates detailed trend metrics
func (s *RiskHistoryTrackingService) calculateTrendMetrics(assessments []*RiskAssessment) map[string]interface{} {
	if len(assessments) < 2 {
		return map[string]interface{}{
			"trend": "insufficient_data",
		}
	}

	// Calculate score changes
	var scoreChanges []float64
	for i := 0; i < len(assessments)-1; i++ {
		change := assessments[i].OverallScore - assessments[i+1].OverallScore
		scoreChanges = append(scoreChanges, change)
	}

	// Calculate trend direction
	var positiveChanges, negativeChanges, neutralChanges int
	for _, change := range scoreChanges {
		if change > 1 {
			positiveChanges++
		} else if change < -1 {
			negativeChanges++
		} else {
			neutralChanges++
		}
	}

	// Determine overall trend
	var trend string
	if positiveChanges > negativeChanges {
		trend = "declining"
	} else if negativeChanges > positiveChanges {
		trend = "improving"
	} else {
		trend = "stable"
	}

	// Calculate volatility (standard deviation of score changes)
	var sum, sumSquared float64
	for _, change := range scoreChanges {
		sum += change
		sumSquared += change * change
	}
	mean := sum / float64(len(scoreChanges))
	variance := (sumSquared / float64(len(scoreChanges))) - (mean * mean)
	volatility := variance

	return map[string]interface{}{
		"trend":                trend,
		"volatility":           volatility,
		"positive_changes":     positiveChanges,
		"negative_changes":     negativeChanges,
		"neutral_changes":      neutralChanges,
		"total_changes":        len(scoreChanges),
		"average_change":       mean,
		"score_changes":        scoreChanges,
		"assessments_analyzed": len(assessments),
	}
}
