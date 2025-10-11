package data

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// HistoryCollector collects and manages real business assessment history
type HistoryCollector struct {
	cache      Cache
	repository Repository
	logger     *zap.Logger
}

// Cache interface for storing recent assessment data
type Cache interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	GetMultiple(ctx context.Context, keys []string) (map[string]interface{}, error)
}

// Repository interface for persistent storage
type Repository interface {
	GetBusinessAssessments(ctx context.Context, businessID string, limit int) ([]*models.RiskAssessment, error)
	SaveAssessment(ctx context.Context, assessment *models.RiskAssessment) error
	GetAssessmentsByDateRange(ctx context.Context, businessID string, startDate, endDate time.Time) ([]*models.RiskAssessment, error)
}

// NewHistoryCollector creates a new history collector
func NewHistoryCollector(cache Cache, repository Repository, logger *zap.Logger) *HistoryCollector {
	return &HistoryCollector{
		cache:      cache,
		repository: repository,
		logger:     logger,
	}
}

// GetBusinessHistory retrieves historical risk assessments for a business
func (hc *HistoryCollector) GetBusinessHistory(ctx context.Context, businessID string, months int) ([]RiskDataPoint, bool) {
	// Try to get from cache first
	cacheKey := fmt.Sprintf("business_history:%s:%d", businessID, months)
	if cached, err := hc.cache.Get(ctx, cacheKey); err == nil {
		if history, ok := cached.([]RiskDataPoint); ok {
			hc.logger.Debug("Retrieved business history from cache",
				zap.String("business_id", businessID),
				zap.Int("months", months),
				zap.Int("data_points", len(history)))
			return history, true
		}
	}

	// Get from repository
	cutoffDate := time.Now().AddDate(0, -months, 0)
	assessments, err := hc.repository.GetAssessmentsByDateRange(ctx, businessID, cutoffDate, time.Now())
	if err != nil {
		hc.logger.Warn("Failed to retrieve business history from repository",
			zap.String("business_id", businessID),
			zap.Error(err))
		return nil, false
	}

	if len(assessments) == 0 {
		hc.logger.Debug("No historical assessments found",
			zap.String("business_id", businessID),
			zap.Int("months", months))
		return nil, false
	}

	// Convert assessments to risk data points
	history := hc.convertAssessmentsToDataPoints(assessments)

	// Cache the result
	if err := hc.cache.Set(ctx, cacheKey, history, 1*time.Hour); err != nil {
		hc.logger.Warn("Failed to cache business history",
			zap.String("business_id", businessID),
			zap.Error(err))
	}

	hc.logger.Info("Retrieved business history from repository",
		zap.String("business_id", businessID),
		zap.Int("months", months),
		zap.Int("assessments", len(assessments)),
		zap.Int("data_points", len(history)))

	return history, true
}

// RecordAssessment records a new risk assessment for future history collection
func (hc *HistoryCollector) RecordAssessment(ctx context.Context, assessment *models.RiskAssessment) error {
	// Save to repository
	if err := hc.repository.SaveAssessment(ctx, assessment); err != nil {
		hc.logger.Error("Failed to save assessment to repository",
			zap.String("assessment_id", assessment.ID),
			zap.String("business_id", assessment.BusinessID),
			zap.Error(err))
		return fmt.Errorf("failed to save assessment: %w", err)
	}

	// Invalidate cache for this business
	cacheKey := fmt.Sprintf("business_history:%s:*", assessment.BusinessID)
	if err := hc.cache.Delete(ctx, cacheKey); err != nil {
		hc.logger.Warn("Failed to invalidate cache",
			zap.String("business_id", assessment.BusinessID),
			zap.Error(err))
	}

	// Add to recent assessments cache
	recentKey := fmt.Sprintf("recent_assessment:%s", assessment.BusinessID)
	if err := hc.cache.Set(ctx, recentKey, assessment, 24*time.Hour); err != nil {
		hc.logger.Warn("Failed to cache recent assessment",
			zap.String("business_id", assessment.BusinessID),
			zap.Error(err))
	}

	hc.logger.Info("Recorded new assessment",
		zap.String("assessment_id", assessment.ID),
		zap.String("business_id", assessment.BusinessID),
		zap.Float64("risk_score", assessment.RiskScore),
		zap.String("risk_level", string(assessment.RiskLevel)))

	return nil
}

// GetRecentAssessment gets the most recent assessment for a business
func (hc *HistoryCollector) GetRecentAssessment(ctx context.Context, businessID string) (*models.RiskAssessment, bool) {
	recentKey := fmt.Sprintf("recent_assessment:%s", businessID)
	if cached, err := hc.cache.Get(ctx, recentKey); err == nil {
		if assessment, ok := cached.(*models.RiskAssessment); ok {
			return assessment, true
		}
	}

	// Get from repository
	assessments, err := hc.repository.GetBusinessAssessments(ctx, businessID, 1)
	if err != nil {
		hc.logger.Warn("Failed to get recent assessment from repository",
			zap.String("business_id", businessID),
			zap.Error(err))
		return nil, false
	}

	if len(assessments) == 0 {
		return nil, false
	}

	// Cache the result
	if err := hc.cache.Set(ctx, recentKey, assessments[0], 24*time.Hour); err != nil {
		hc.logger.Warn("Failed to cache recent assessment",
			zap.String("business_id", businessID),
			zap.Error(err))
	}

	return assessments[0], true
}

// GetAssessmentTrends analyzes trends in business assessments over time
func (hc *HistoryCollector) GetAssessmentTrends(ctx context.Context, businessID string, months int) (*AssessmentTrends, error) {
	history, found := hc.GetBusinessHistory(ctx, businessID, months)
	if !found || len(history) < 2 {
		return nil, fmt.Errorf("insufficient history data for trend analysis")
	}

	trends := &AssessmentTrends{
		BusinessID:     businessID,
		AnalysisPeriod: months,
		DataPoints:     len(history),
		StartDate:      history[0].Timestamp,
		EndDate:        history[len(history)-1].Timestamp,
	}

	// Calculate risk score trend
	trends.RiskScoreTrend = hc.calculateTrend(history, func(dp RiskDataPoint) float64 {
		return dp.RiskScore
	})

	// Calculate financial health trend
	trends.FinancialHealthTrend = hc.calculateTrend(history, func(dp RiskDataPoint) float64 {
		return dp.FinancialHealth
	})

	// Calculate compliance score trend
	trends.ComplianceScoreTrend = hc.calculateTrend(history, func(dp RiskDataPoint) float64 {
		return dp.ComplianceScore
	})

	// Calculate volatility
	trends.RiskVolatility = hc.calculateVolatility(history, func(dp RiskDataPoint) float64 {
		return dp.RiskScore
	})

	// Calculate average values
	trends.AverageRiskScore = hc.calculateAverage(history, func(dp RiskDataPoint) float64 {
		return dp.RiskScore
	})
	trends.AverageFinancialHealth = hc.calculateAverage(history, func(dp RiskDataPoint) float64 {
		return dp.FinancialHealth
	})
	trends.AverageComplianceScore = hc.calculateAverage(history, func(dp RiskDataPoint) float64 {
		return dp.ComplianceScore
	})

	hc.logger.Info("Calculated assessment trends",
		zap.String("business_id", businessID),
		zap.Int("months", months),
		zap.Int("data_points", len(history)),
		zap.Float64("risk_trend", trends.RiskScoreTrend),
		zap.Float64("risk_volatility", trends.RiskVolatility))

	return trends, nil
}

// convertAssessmentsToDataPoints converts risk assessments to risk data points
func (hc *HistoryCollector) convertAssessmentsToDataPoints(assessments []*models.RiskAssessment) []RiskDataPoint {
	dataPoints := make([]RiskDataPoint, len(assessments))

	for i, assessment := range assessments {
		// Extract features from risk factors
		financialHealth := hc.extractFeatureFromRiskFactors(assessment.RiskFactors, "financial_health")
		complianceScore := hc.extractFeatureFromRiskFactors(assessment.RiskFactors, "compliance_score")
		marketConditions := hc.extractFeatureFromRiskFactors(assessment.RiskFactors, "market_conditions")
		revenueTrend := hc.extractFeatureFromRiskFactors(assessment.RiskFactors, "revenue_trend")
		employeeGrowth := hc.extractFeatureFromRiskFactors(assessment.RiskFactors, "employee_growth")
		riskVolatility := hc.extractFeatureFromRiskFactors(assessment.RiskFactors, "risk_volatility")

		dataPoints[i] = RiskDataPoint{
			Timestamp:        assessment.CreatedAt,
			RiskScore:        assessment.RiskScore,
			FinancialHealth:  financialHealth,
			ComplianceScore:  complianceScore,
			MarketConditions: marketConditions,
			RevenueTrend:     revenueTrend,
			EmployeeGrowth:   employeeGrowth,
			RiskVolatility:   riskVolatility,
		}
	}

	return dataPoints
}

// extractFeatureFromRiskFactors extracts a specific feature value from risk factors
func (hc *HistoryCollector) extractFeatureFromRiskFactors(factors []models.RiskFactor, featureName string) float64 {
	for _, factor := range factors {
		if factor.Name == featureName {
			return factor.Score
		}
	}
	return 0.5 // Default value if feature not found
}

// calculateTrend calculates the trend (slope) of a feature over time
func (hc *HistoryCollector) calculateTrend(history []RiskDataPoint, extractor func(RiskDataPoint) float64) float64 {
	if len(history) < 2 {
		return 0.0
	}

	// Simple linear regression to calculate trend
	n := len(history)
	sumX, sumY, sumXY, sumXX := 0.0, 0.0, 0.0, 0.0

	for i, dp := range history {
		x := float64(i)
		y := extractor(dp)
		sumX += x
		sumY += y
		sumXY += x * y
		sumXX += x * x
	}

	// Calculate slope
	slope := (float64(n)*sumXY - sumX*sumY) / (float64(n)*sumXX - sumX*sumX)
	return slope
}

// calculateVolatility calculates the volatility (standard deviation) of a feature
func (hc *HistoryCollector) calculateVolatility(history []RiskDataPoint, extractor func(RiskDataPoint) float64) float64 {
	if len(history) < 2 {
		return 0.0
	}

	// Calculate mean
	sum := 0.0
	for _, dp := range history {
		sum += extractor(dp)
	}
	mean := sum / float64(len(history))

	// Calculate variance
	variance := 0.0
	for _, dp := range history {
		diff := extractor(dp) - mean
		variance += diff * diff
	}
	variance /= float64(len(history) - 1)

	// Return standard deviation
	return math.Sqrt(variance)
}

// calculateAverage calculates the average of a feature
func (hc *HistoryCollector) calculateAverage(history []RiskDataPoint, extractor func(RiskDataPoint) float64) float64 {
	if len(history) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, dp := range history {
		sum += extractor(dp)
	}
	return sum / float64(len(history))
}

// AssessmentTrends contains trend analysis results
type AssessmentTrends struct {
	BusinessID             string    `json:"business_id"`
	AnalysisPeriod         int       `json:"analysis_period_months"`
	DataPoints             int       `json:"data_points"`
	StartDate              time.Time `json:"start_date"`
	EndDate                time.Time `json:"end_date"`
	RiskScoreTrend         float64   `json:"risk_score_trend"`
	FinancialHealthTrend   float64   `json:"financial_health_trend"`
	ComplianceScoreTrend   float64   `json:"compliance_score_trend"`
	RiskVolatility         float64   `json:"risk_volatility"`
	AverageRiskScore       float64   `json:"average_risk_score"`
	AverageFinancialHealth float64   `json:"average_financial_health"`
	AverageComplianceScore float64   `json:"average_compliance_score"`
}

// Health checks the health of the history collector
func (hc *HistoryCollector) Health(ctx context.Context) error {
	// Test cache connectivity
	testKey := "health_check"
	testValue := "test"
	if err := hc.cache.Set(ctx, testKey, testValue, 1*time.Minute); err != nil {
		return fmt.Errorf("cache health check failed: %w", err)
	}

	if _, err := hc.cache.Get(ctx, testKey); err != nil {
		return fmt.Errorf("cache retrieval health check failed: %w", err)
	}

	// Test repository connectivity (if available)
	// This would depend on the specific repository implementation
	// For now, we'll assume it's healthy if cache is working

	return nil
}
