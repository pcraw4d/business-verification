package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// RiskHistoryService provides risk history tracking functionality
type RiskHistoryService struct {
	logger   *observability.Logger
	database database.Database
}

// NewRiskHistoryService creates a new risk history service
func NewRiskHistoryService(logger *observability.Logger, database database.Database) *RiskHistoryService {
	return &RiskHistoryService{
		logger:   logger,
		database: database,
	}
}

// RiskHistoryEntry represents a risk history entry with trend analysis
type RiskHistoryEntry struct {
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
	BusinessID       string                 `json:"business_id"`
	BusinessName     string                 `json:"business_name"`
	TotalAssessments int                    `json:"total_assessments"`
	CurrentScore     float64                `json:"current_score"`
	CurrentLevel     string                 `json:"current_level"`
	Trend            string                 `json:"trend"`
	History          []RiskHistoryEntry     `json:"history"`
	Statistics       map[string]interface{} `json:"statistics"`
	GeneratedAt      time.Time              `json:"generated_at"`
}

// RiskTrendAnalysis represents trend analysis for risk assessments
type RiskTrendAnalysis struct {
	BusinessID       string            `json:"business_id"`
	Period           string            `json:"period"`      // "1month", "3months", "6months", "1year"
	ScoreTrend       string            `json:"score_trend"` // "improving", "stable", "declining"
	LevelTrend       string            `json:"level_trend"` // "improving", "stable", "declining"
	AverageScore     float64           `json:"average_score"`
	ScoreVolatility  float64           `json:"score_volatility"`
	RiskFactorTrends map[string]string `json:"risk_factor_trends"`
	Recommendations  []string          `json:"recommendations"`
	GeneratedAt      time.Time         `json:"generated_at"`
}

// StoreRiskAssessment stores a risk assessment in the history
func (s *RiskHistoryService) StoreRiskAssessment(ctx context.Context, assessment *RiskAssessment) error {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Storing risk assessment in history",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"assessment_id", assessment.ID,
		"overall_score", assessment.OverallScore,
		"overall_level", assessment.OverallLevel,
	)

	// Convert assessment to database model
	dbAssessment := &database.RiskAssessment{
		ID:               assessment.ID,
		BusinessID:       assessment.BusinessID,
		BusinessName:     assessment.BusinessName,
		OverallScore:     assessment.OverallScore,
		OverallLevel:     string(assessment.OverallLevel),
		CategoryScores:   s.convertCategoryScores(assessment.CategoryScores),
		FactorScores:     s.convertFactorScores(assessment.FactorScores),
		Recommendations:  s.convertRecommendations(assessment.Recommendations),
		Predictions:      []string{}, // Predictions are handled separately in response
		Alerts:           []string{}, // Alerts are handled separately in response
		AssessmentMethod: "comprehensive",
		Source:           "risk_service",
		Metadata:         assessment.Metadata,
		AssessedAt:       assessment.AssessedAt,
		ValidUntil:       assessment.ValidUntil,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Store in database
	if s.database != nil {
		if err := s.database.CreateRiskAssessment(ctx, dbAssessment); err != nil {
			s.logger.Error("Failed to store risk assessment",
				"request_id", requestID,
				"business_id", assessment.BusinessID,
				"error", err.Error(),
			)
			return fmt.Errorf("failed to store risk assessment: %w", err)
		}
	} else {
		s.logger.Info("Database not available, skipping risk assessment storage",
			"request_id", requestID,
			"business_id", assessment.BusinessID,
		)
	}

	s.logger.Info("Risk assessment stored successfully",
		"request_id", requestID,
		"business_id", assessment.BusinessID,
		"assessment_id", assessment.ID,
	)

	return nil
}

// GetRiskHistory retrieves risk assessment history for a business
func (s *RiskHistoryService) GetRiskHistory(ctx context.Context, businessID string, limit, offset int) (*RiskHistoryResponse, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving risk history",
		"request_id", requestID,
		"business_id", businessID,
		"limit", limit,
		"offset", offset,
	)

	// Get risk assessments from database
	var dbAssessments []*database.RiskAssessment
	var err error

	if s.database != nil {
		dbAssessments, err = s.database.GetRiskAssessmentHistory(ctx, businessID, limit, offset)
		if err != nil {
			s.logger.Error("Failed to retrieve risk history",
				"request_id", requestID,
				"business_id", businessID,
				"error", err.Error(),
			)
			return nil, fmt.Errorf("failed to retrieve risk history: %w", err)
		}
	} else {
		s.logger.Info("Database not available, returning empty risk history",
			"request_id", requestID,
			"business_id", businessID,
		)
		dbAssessments = []*database.RiskAssessment{}
	}

	// Convert database models to risk assessment models
	assessments := make([]*RiskAssessment, len(dbAssessments))
	for i, dbAssessment := range dbAssessments {
		assessment, err := s.convertDBAssessmentToRiskAssessment(dbAssessment)
		if err != nil {
			s.logger.Warn("Failed to convert database assessment",
				"request_id", requestID,
				"assessment_id", dbAssessment.ID,
				"error", err.Error(),
			)
			continue
		}
		assessments[i] = assessment
	}

	// Create history entries with trend analysis
	historyEntries := s.createHistoryEntries(assessments)

	// Get current assessment
	var currentScore float64
	var currentLevel string
	if len(assessments) > 0 {
		currentScore = assessments[0].OverallScore
		currentLevel = string(assessments[0].OverallLevel)
	}

	// Determine overall trend
	trend := s.determineOverallTrend(assessments)

	// Get statistics
	var statistics map[string]interface{}
	if s.database != nil {
		var err error
		statistics, err = s.database.GetRiskAssessmentStatistics(ctx, businessID)
		if err != nil {
			s.logger.Warn("Failed to get risk assessment statistics",
				"request_id", requestID,
				"business_id", businessID,
				"error", err.Error(),
			)
			statistics = make(map[string]interface{})
		}
	} else {
		statistics = make(map[string]interface{})
	}

	response := &RiskHistoryResponse{
		BusinessID:       businessID,
		BusinessName:     s.getBusinessName(assessments),
		TotalAssessments: len(assessments),
		CurrentScore:     currentScore,
		CurrentLevel:     currentLevel,
		Trend:            trend,
		History:          historyEntries,
		Statistics:       statistics,
		GeneratedAt:      time.Now(),
	}

	s.logger.Info("Risk history retrieved successfully",
		"request_id", requestID,
		"business_id", businessID,
		"total_assessments", len(assessments),
		"trend", trend,
	)

	return response, nil
}

// GetRiskTrends retrieves trend analysis for a business
func (s *RiskHistoryService) GetRiskTrends(ctx context.Context, businessID string, days int) (*RiskTrendAnalysis, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving risk trends",
		"request_id", requestID,
		"business_id", businessID,
		"days", days,
	)

	// Get risk assessments for trend analysis
	var dbAssessments []*database.RiskAssessment
	var err error
	
	if s.database != nil {
		dbAssessments, err = s.database.GetRiskAssessmentTrends(ctx, businessID, days)
		if err != nil {
			s.logger.Error("Failed to retrieve risk trends",
				"request_id", requestID,
				"business_id", businessID,
				"error", err.Error(),
			)
			return nil, fmt.Errorf("failed to retrieve risk trends: %w", err)
		}
	} else {
		s.logger.Info("Database not available, returning empty risk trends",
			"request_id", requestID,
			"business_id", businessID,
		)
		dbAssessments = []*database.RiskAssessment{}
	}

	// Convert to risk assessment models
	assessments := make([]*RiskAssessment, len(dbAssessments))
	for i, dbAssessment := range dbAssessments {
		assessment, err := s.convertDBAssessmentToRiskAssessment(dbAssessment)
		if err != nil {
			s.logger.Warn("Failed to convert database assessment for trends",
				"request_id", requestID,
				"assessment_id", dbAssessment.ID,
				"error", err.Error(),
			)
			continue
		}
		assessments[i] = assessment
	}

	// Analyze trends
	trendAnalysis := s.analyzeTrends(assessments, days)

	s.logger.Info("Risk trends retrieved successfully",
		"request_id", requestID,
		"business_id", businessID,
		"score_trend", trendAnalysis.ScoreTrend,
		"level_trend", trendAnalysis.LevelTrend,
	)

	return trendAnalysis, nil
}

// GetRiskHistoryByDateRange retrieves risk history within a date range
func (s *RiskHistoryService) GetRiskHistoryByDateRange(ctx context.Context, businessID string, startDate, endDate time.Time) ([]*RiskAssessment, error) {
	requestID := ctx.Value("request_id").(string)

	s.logger.Info("Retrieving risk history by date range",
		"request_id", requestID,
		"business_id", businessID,
		"start_date", startDate,
		"end_date", endDate,
	)

	// Get risk assessments from database
	dbAssessments, err := s.database.GetRiskAssessmentHistoryByDateRange(ctx, businessID, startDate, endDate)
	if err != nil {
		s.logger.Error("Failed to retrieve risk history by date range",
			"request_id", requestID,
			"business_id", businessID,
			"error", err.Error(),
		)
		return nil, fmt.Errorf("failed to retrieve risk history by date range: %w", err)
	}

	// Convert database models to risk assessment models
	assessments := make([]*RiskAssessment, len(dbAssessments))
	for i, dbAssessment := range dbAssessments {
		assessment, err := s.convertDBAssessmentToRiskAssessment(dbAssessment)
		if err != nil {
			s.logger.Warn("Failed to convert database assessment for date range",
				"request_id", requestID,
				"assessment_id", dbAssessment.ID,
				"error", err.Error(),
			)
			continue
		}
		assessments[i] = assessment
	}

	s.logger.Info("Risk history by date range retrieved successfully",
		"request_id", requestID,
		"business_id", businessID,
		"total_assessments", len(assessments),
	)

	return assessments, nil
}

// Helper methods for data conversion and analysis

func (s *RiskHistoryService) convertCategoryScores(categoryScores map[RiskCategory]RiskScore) map[string]interface{} {
	result := make(map[string]interface{})
	for category, score := range categoryScores {
		result[string(category)] = map[string]interface{}{
			"score":       score.Score,
			"level":       string(score.Level),
			"confidence":  score.Confidence,
			"explanation": score.Explanation,
		}
	}
	return result
}

func (s *RiskHistoryService) convertFactorScores(factorScores []RiskScore) []string {
	result := make([]string, len(factorScores))
	for i, score := range factorScores {
		data := map[string]interface{}{
			"factor_id":   score.FactorID,
			"factor_name": score.FactorName,
			"category":    string(score.Category),
			"score":       score.Score,
			"level":       string(score.Level),
			"confidence":  score.Confidence,
			"explanation": score.Explanation,
		}
		jsonData, _ := json.Marshal(data)
		result[i] = string(jsonData)
	}
	return result
}

func (s *RiskHistoryService) convertRecommendations(recommendations []RiskRecommendation) []string {
	result := make([]string, len(recommendations))
	for i, rec := range recommendations {
		data := map[string]interface{}{
			"id":          rec.ID,
			"risk_factor": rec.RiskFactor,
			"title":       rec.Title,
			"description": rec.Description,
			"priority":    string(rec.Priority),
			"action":      rec.Action,
			"impact":      rec.Impact,
			"timeline":    rec.Timeline,
		}
		jsonData, _ := json.Marshal(data)
		result[i] = string(jsonData)
	}
	return result
}

func (s *RiskHistoryService) convertPredictions(predictions []RiskPrediction) []string {
	result := make([]string, len(predictions))
	for i, pred := range predictions {
		data := map[string]interface{}{
			"horizon":         pred.Horizon,
			"predicted_score": pred.PredictedScore,
			"predicted_level": string(pred.PredictedLevel),
			"confidence":      pred.Confidence,
			"factors":         pred.Factors,
		}
		jsonData, _ := json.Marshal(data)
		result[i] = string(jsonData)
	}
	return result
}

func (s *RiskHistoryService) convertAlerts(alerts []RiskAlert) []string {
	result := make([]string, len(alerts))
	for i, alert := range alerts {
		data := map[string]interface{}{
			"id":           alert.ID,
			"risk_factor":  alert.RiskFactor,
			"level":        string(alert.Level),
			"message":      alert.Message,
			"score":        alert.Score,
			"threshold":    alert.Threshold,
			"triggered_at": alert.TriggeredAt,
		}
		jsonData, _ := json.Marshal(data)
		result[i] = string(jsonData)
	}
	return result
}

func (s *RiskHistoryService) convertDBAssessmentToRiskAssessment(dbAssessment *database.RiskAssessment) (*RiskAssessment, error) {
	// Convert category scores back to map
	var categoryScores map[RiskCategory]RiskScore
	if dbAssessment.CategoryScores != nil {
		categoryScores = make(map[RiskCategory]RiskScore)
		for key, value := range dbAssessment.CategoryScores {
			if scoreData, ok := value.(map[string]interface{}); ok {
				categoryScores[RiskCategory(key)] = RiskScore{
					Score:       scoreData["score"].(float64),
					Level:       RiskLevel(scoreData["level"].(string)),
					Confidence:  scoreData["confidence"].(float64),
					Explanation: scoreData["explanation"].(string),
				}
			}
		}
	}

	// Convert factor scores back to slice
	var factorScores []RiskScore
	for _, scoreJSON := range dbAssessment.FactorScores {
		var scoreData map[string]interface{}
		if err := json.Unmarshal([]byte(scoreJSON), &scoreData); err == nil {
			factorScores = append(factorScores, RiskScore{
				FactorID:    scoreData["factor_id"].(string),
				FactorName:  scoreData["factor_name"].(string),
				Category:    RiskCategory(scoreData["category"].(string)),
				Score:       scoreData["score"].(float64),
				Level:       RiskLevel(scoreData["level"].(string)),
				Confidence:  scoreData["confidence"].(float64),
				Explanation: scoreData["explanation"].(string),
			})
		}
	}

	// Convert recommendations back to slice
	var recommendations []RiskRecommendation
	for _, recJSON := range dbAssessment.Recommendations {
		var recData map[string]interface{}
		if err := json.Unmarshal([]byte(recJSON), &recData); err == nil {
			recommendations = append(recommendations, RiskRecommendation{
				ID:          recData["id"].(string),
				RiskFactor:  recData["risk_factor"].(string),
				Title:       recData["title"].(string),
				Description: recData["description"].(string),
				Priority:    RiskLevel(recData["priority"].(string)),
				Action:      recData["action"].(string),
				Impact:      recData["impact"].(string),
				Timeline:    recData["timeline"].(string),
			})
		}
	}

	// Convert predictions back to slice
	var predictions []RiskPrediction
	for _, predJSON := range dbAssessment.Predictions {
		var predData map[string]interface{}
		if err := json.Unmarshal([]byte(predJSON), &predData); err == nil {
			predictions = append(predictions, RiskPrediction{
				Horizon:        predData["horizon"].(string),
				PredictedScore: predData["predicted_score"].(float64),
				PredictedLevel: RiskLevel(predData["predicted_level"].(string)),
				Confidence:     predData["confidence"].(float64),
				Factors:        s.convertToStringSlice(predData["factors"]),
			})
		}
	}

	// Convert alerts back to slice
	var alerts []RiskAlert
	for _, alertJSON := range dbAssessment.Alerts {
		var alertData map[string]interface{}
		if err := json.Unmarshal([]byte(alertJSON), &alertData); err == nil {
			alerts = append(alerts, RiskAlert{
				ID:           alertData["id"].(string),
				BusinessID:   dbAssessment.BusinessID,
				RiskFactor:   alertData["risk_factor"].(string),
				Level:        RiskLevel(alertData["level"].(string)),
				Message:      alertData["message"].(string),
				Score:        alertData["score"].(float64),
				Threshold:    alertData["threshold"].(float64),
				TriggeredAt:  time.Unix(int64(alertData["triggered_at"].(float64)), 0),
				Acknowledged: false,
			})
		}
	}

	return &RiskAssessment{
		ID:              dbAssessment.ID,
		BusinessID:      dbAssessment.BusinessID,
		BusinessName:    dbAssessment.BusinessName,
		OverallScore:    dbAssessment.OverallScore,
		OverallLevel:    RiskLevel(dbAssessment.OverallLevel),
		CategoryScores:  categoryScores,
		FactorScores:    factorScores,
		Recommendations: recommendations,
		AlertLevel:      RiskLevelLow, // Default alert level
		Metadata:        dbAssessment.Metadata,
		AssessedAt:      dbAssessment.AssessedAt,
		ValidUntil:      dbAssessment.ValidUntil,
	}, nil
}

func (s *RiskHistoryService) convertToStringSlice(data interface{}) []string {
	if slice, ok := data.([]interface{}); ok {
		result := make([]string, len(slice))
		for i, item := range slice {
			result[i] = item.(string)
		}
		return result
	}
	return []string{}
}

func (s *RiskHistoryService) createHistoryEntries(assessments []*RiskAssessment) []RiskHistoryEntry {
	entries := make([]RiskHistoryEntry, len(assessments))

	for i, assessment := range assessments {
		entry := RiskHistoryEntry{
			Assessment: assessment,
		}

		// Calculate trend if we have previous assessment
		if i < len(assessments)-1 {
			prevAssessment := assessments[i+1]
			entry.ScoreChange = assessment.OverallScore - prevAssessment.OverallScore

			if entry.ScoreChange > 5 {
				entry.Trend = "declining"
			} else if entry.ScoreChange < -5 {
				entry.Trend = "improving"
			} else {
				entry.Trend = "stable"
			}

			// Determine level change
			if assessment.OverallLevel > prevAssessment.OverallLevel {
				entry.LevelChange = "up"
			} else if assessment.OverallLevel < prevAssessment.OverallLevel {
				entry.LevelChange = "down"
			} else {
				entry.LevelChange = "same"
			}

			// Calculate days since last assessment
			entry.DaysSinceLast = int(assessment.AssessedAt.Sub(prevAssessment.AssessedAt).Hours() / 24)
		}

		// Count alerts
		entry.AlertCount = len(assessment.Alerts)

		// Get key recommendation
		if len(assessment.Recommendations) > 0 {
			entry.Recommendation = assessment.Recommendations[0].Title
		}

		entries[i] = entry
	}

	return entries
}

func (s *RiskHistoryService) determineOverallTrend(assessments []*RiskAssessment) string {
	if len(assessments) < 2 {
		return "insufficient_data"
	}

	// Calculate average score change over the last 3 assessments
	recentAssessments := assessments
	if len(assessments) > 3 {
		recentAssessments = assessments[:3]
	}

	totalChange := 0.0
	for i := 0; i < len(recentAssessments)-1; i++ {
		totalChange += recentAssessments[i].OverallScore - recentAssessments[i+1].OverallScore
	}

	averageChange := totalChange / float64(len(recentAssessments)-1)

	if averageChange > 10 {
		return "declining"
	} else if averageChange < -10 {
		return "improving"
	} else {
		return "stable"
	}
}

func (s *RiskHistoryService) getBusinessName(assessments []*RiskAssessment) string {
	if len(assessments) > 0 {
		return assessments[0].BusinessName
	}
	return ""
}

func (s *RiskHistoryService) analyzeTrends(assessments []*RiskAssessment, days int) *RiskTrendAnalysis {
	analysis := &RiskTrendAnalysis{
		Period:      fmt.Sprintf("%dmonths", days/30),
		GeneratedAt: time.Now(),
	}

	if len(assessments) == 0 {
		analysis.ScoreTrend = "insufficient_data"
		analysis.LevelTrend = "insufficient_data"
		return analysis
	}

	analysis.BusinessID = assessments[0].BusinessID

	if len(assessments) < 2 {
		analysis.ScoreTrend = "insufficient_data"
		analysis.LevelTrend = "insufficient_data"
		return analysis
	}

	// Calculate average score
	totalScore := 0.0
	for _, assessment := range assessments {
		totalScore += assessment.OverallScore
	}
	analysis.AverageScore = totalScore / float64(len(assessments))

	// Calculate score volatility (standard deviation)
	variance := 0.0
	for _, assessment := range assessments {
		variance += (assessment.OverallScore - analysis.AverageScore) * (assessment.OverallScore - analysis.AverageScore)
	}
	analysis.ScoreVolatility = variance / float64(len(assessments))

	// Determine score trend
	firstScore := assessments[len(assessments)-1].OverallScore
	lastScore := assessments[0].OverallScore
	scoreChange := lastScore - firstScore

	if scoreChange > 10 {
		analysis.ScoreTrend = "declining"
	} else if scoreChange < -10 {
		analysis.ScoreTrend = "improving"
	} else {
		analysis.ScoreTrend = "stable"
	}

	// Determine level trend
	firstLevel := assessments[len(assessments)-1].OverallLevel
	lastLevel := assessments[0].OverallLevel

	if lastLevel > firstLevel {
		analysis.LevelTrend = "declining"
	} else if lastLevel < firstLevel {
		analysis.LevelTrend = "improving"
	} else {
		analysis.LevelTrend = "stable"
	}

	// Analyze risk factor trends
	analysis.RiskFactorTrends = s.analyzeRiskFactorTrends(assessments)

	// Generate recommendations based on trends
	analysis.Recommendations = s.generateTrendRecommendations(analysis)

	return analysis
}

func (s *RiskHistoryService) analyzeRiskFactorTrends(assessments []*RiskAssessment) map[string]string {
	trends := make(map[string]string)

	if len(assessments) < 2 {
		return trends
	}

	// Analyze trends for each risk factor
	firstAssessment := assessments[len(assessments)-1]
	lastAssessment := assessments[0]

	for _, lastFactor := range lastAssessment.FactorScores {
		// Find corresponding factor in first assessment
		for _, firstFactor := range firstAssessment.FactorScores {
			if lastFactor.FactorID == firstFactor.FactorID {
				scoreChange := lastFactor.Score - firstFactor.Score
				if scoreChange > 5 {
					trends[lastFactor.FactorID] = "increasing"
				} else if scoreChange < -5 {
					trends[lastFactor.FactorID] = "decreasing"
				} else {
					trends[lastFactor.FactorID] = "stable"
				}
				break
			}
		}
	}

	return trends
}

func (s *RiskHistoryService) generateTrendRecommendations(analysis *RiskTrendAnalysis) []string {
	var recommendations []string

	if analysis.ScoreTrend == "declining" {
		recommendations = append(recommendations, "Risk level is increasing. Consider implementing additional risk mitigation measures.")
	}

	if analysis.ScoreVolatility > 20 {
		recommendations = append(recommendations, "High score volatility detected. Review risk assessment methodology for consistency.")
	}

	if analysis.LevelTrend == "declining" {
		recommendations = append(recommendations, "Risk level has increased. Immediate attention required for risk management.")
	}

	// Add factor-specific recommendations
	for factorID, trend := range analysis.RiskFactorTrends {
		if trend == "increasing" {
			recommendations = append(recommendations, fmt.Sprintf("Risk factor %s is trending upward. Review and address underlying issues.", factorID))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Risk profile appears stable. Continue monitoring and maintain current risk management practices.")
	}

	return recommendations
}
