package risk

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// EnhancedRiskService orchestrates all enhanced risk assessment components
type EnhancedRiskService struct {
	calculator           *EnhancedRiskCalculator
	recommendationEngine *RiskRecommendationEngine
	trendAnalysisService *RiskTrendAnalysisService
	alertSystem          *RiskAlertSystem
	correlationAnalyzer  *CorrelationAnalyzer
	confidenceCalibrator *ConfidenceCalibrator
	logger               *zap.Logger
}

// NewEnhancedRiskService creates a new enhanced risk service
func NewEnhancedRiskService(
	calculator *EnhancedRiskCalculator,
	recommendationEngine *RiskRecommendationEngine,
	trendAnalysisService *RiskTrendAnalysisService,
	alertSystem *RiskAlertSystem,
	correlationAnalyzer *CorrelationAnalyzer,
	confidenceCalibrator *ConfidenceCalibrator,
	logger *zap.Logger,
) *EnhancedRiskService {
	return &EnhancedRiskService{
		calculator:           calculator,
		recommendationEngine: recommendationEngine,
		trendAnalysisService: trendAnalysisService,
		alertSystem:          alertSystem,
		correlationAnalyzer:  correlationAnalyzer,
		confidenceCalibrator: confidenceCalibrator,
		logger:               logger,
	}
}

// PerformEnhancedRiskAssessment performs a comprehensive risk assessment
func (s *EnhancedRiskService) PerformEnhancedRiskAssessment(
	ctx context.Context,
	request *EnhancedRiskAssessmentRequest,
) (*EnhancedRiskAssessmentResponse, error) {
	s.logger.Info("Starting enhanced risk assessment",
		zap.String("business_id", request.BusinessID),
		zap.String("assessment_id", request.AssessmentID))

	// Calculate risk factors
	riskFactors, err := s.calculateRiskFactors(ctx, request)
	if err != nil {
		s.logger.Error("Failed to calculate risk factors",
			zap.Error(err),
			zap.String("business_id", request.BusinessID))
		return nil, fmt.Errorf("failed to calculate risk factors: %w", err)
	}

	// Generate recommendations
	recommendations, err := s.recommendationEngine.GenerateRecommendations(ctx, riskFactors)
	if err != nil {
		s.logger.Error("Failed to generate recommendations",
			zap.Error(err),
			zap.String("business_id", request.BusinessID))
		return nil, fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Analyze trends if historical data is available
	var trendData *RiskTrendData
	if request.IncludeTrendAnalysis {
		trendData, err = s.analyzeRiskTrends(ctx, request.BusinessID, request.TimeRange)
		if err != nil {
			s.logger.Warn("Failed to analyze risk trends",
				zap.Error(err),
				zap.String("business_id", request.BusinessID))
			// Don't fail the entire assessment for trend analysis failure
		}
	}

	// Perform correlation analysis
	var correlationData map[string]float64
	if request.IncludeCorrelationAnalysis {
		correlationData, err = s.performCorrelationAnalysis(ctx, riskFactors)
		if err != nil {
			s.logger.Warn("Failed to perform correlation analysis",
				zap.Error(err),
				zap.String("business_id", request.BusinessID))
			// Don't fail the entire assessment for correlation analysis failure
		}
	}

	// Calibrate confidence scores
	calibratedFactors, err := s.calibrateConfidenceScores(ctx, riskFactors)
	if err != nil {
		s.logger.Warn("Failed to calibrate confidence scores",
			zap.Error(err),
			zap.String("business_id", request.BusinessID))
		// Use original factors if calibration fails
		calibratedFactors = riskFactors
	}

	// Check for alerts
	alerts, err := s.alertSystem.CheckAndTriggerAlerts(ctx, calibratedFactors)
	if err != nil {
		s.logger.Error("Failed to check alerts",
			zap.Error(err),
			zap.String("business_id", request.BusinessID))
		return nil, fmt.Errorf("failed to check alerts: %w", err)
	}

	// Calculate overall risk score
	overallScore := s.calculateOverallRiskScore(calibratedFactors)

	// Build response
	response := &EnhancedRiskAssessmentResponse{
		AssessmentID:     request.AssessmentID,
		BusinessID:       request.BusinessID,
		Timestamp:        time.Now(),
		OverallRiskScore: overallScore,
		OverallRiskLevel: s.determineOverallRiskLevel(overallScore),
		RiskFactors:      calibratedFactors,
		Recommendations:  recommendations,
		TrendData:        trendData,
		CorrelationData:  correlationData,
		Alerts:           alerts,
		ConfidenceScore:  s.calculateOverallConfidence(calibratedFactors),
		ProcessingTimeMs: time.Since(time.Now()).Milliseconds(),
		Metadata: map[string]interface{}{
			"version":              "2.0",
			"assessment_type":      "enhanced",
			"factors_analyzed":     len(calibratedFactors),
			"recommendations":      len(recommendations),
			"alerts_triggered":     len(alerts),
			"trend_analysis":       trendData != nil,
			"correlation_analysis": correlationData != nil,
		},
	}

	s.logger.Info("Enhanced risk assessment completed",
		zap.String("business_id", request.BusinessID),
		zap.String("assessment_id", request.AssessmentID),
		zap.Float64("overall_score", overallScore),
		zap.String("risk_level", string(response.OverallRiskLevel)),
		zap.Int("factors_count", len(calibratedFactors)),
		zap.Int("recommendations_count", len(recommendations)),
		zap.Int("alerts_count", len(alerts)))

	return response, nil
}

// calculateRiskFactors calculates all risk factors for the assessment
func (s *EnhancedRiskService) calculateRiskFactors(
	ctx context.Context,
	request *EnhancedRiskAssessmentRequest,
) ([]RiskFactorDetail, error) {
	var factors []RiskFactorDetail

	// Calculate each risk factor
	for _, factorInput := range request.RiskFactorInputs {
		factor, err := s.calculator.CalculateFactor(ctx, factorInput)
		if err != nil {
			s.logger.Error("Failed to calculate risk factor",
				zap.Error(err),
				zap.String("factor_type", factorInput.FactorType),
				zap.String("business_id", request.BusinessID))
			return nil, fmt.Errorf("failed to calculate factor %s: %w", factorInput.FactorType, err)
		}

		// Convert to detail format
		factorDetail := RiskFactorDetail{
			FactorType:          factorInput.FactorType,
			Score:               factor.Score,
			RiskLevel:           factor.RiskLevel,
			Confidence:          factor.Confidence,
			Weight:              factorInput.Weight,
			Description:         factor.Description,
			ContributingFactors: factor.ContributingFactors,
			LastUpdated:         time.Now(),
			Metadata:            factor.Metadata,
		}

		factors = append(factors, factorDetail)
	}

	return factors, nil
}

// analyzeRiskTrends analyzes risk trends for the business
func (s *EnhancedRiskService) analyzeRiskTrends(
	ctx context.Context,
	businessID string,
	timeRange *TimeRange,
) (*RiskTrendData, error) {
	// Get historical risk data
	historicalData, err := s.getHistoricalRiskData(ctx, businessID, timeRange)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	// Analyze trends
	trends, err := s.trendAnalysisService.GetRiskTrends(ctx, historicalData)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze trends: %w", err)
	}

	// Convert to trend data format
	trendData := &RiskTrendData{
		BusinessID:   businessID,
		TimeRange:    timeRange,
		Trends:       trends,
		LastAnalyzed: time.Now(),
		DataPoints:   len(historicalData),
		TrendSummary: s.generateTrendSummary(trends),
	}

	return trendData, nil
}

// performCorrelationAnalysis performs correlation analysis on risk factors
func (s *EnhancedRiskService) performCorrelationAnalysis(
	ctx context.Context,
	factors []RiskFactorDetail,
) (map[string]float64, error) {
	// Convert factors to correlation input format
	var factorData [][]float64
	var factorNames []string

	for _, factor := range factors {
		factorData = append(factorData, []float64{factor.Score})
		factorNames = append(factorNames, factor.FactorType)
	}

	// Perform correlation analysis
	correlations, err := s.correlationAnalyzer.AnalyzeCorrelation(ctx, factorData, factorNames)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze correlations: %w", err)
	}

	return correlations, nil
}

// calibrateConfidenceScores calibrates confidence scores for risk factors
func (s *EnhancedRiskService) calibrateConfidenceScores(
	ctx context.Context,
	factors []RiskFactorDetail,
) ([]RiskFactorDetail, error) {
	var calibratedFactors []RiskFactorDetail

	for _, factor := range factors {
		// Calibrate confidence score
		// For now, use empty historical data - this should be populated from actual data
		historicalData := []HistoricalDataPoint{}
		calibration, err := s.confidenceCalibrator.CalibrateConfidence(
			factor.FactorType,
			factor.Confidence,
			historicalData,
		)
		var calibratedConfidence float64
		if err != nil {
			s.logger.Warn("Failed to calibrate confidence for factor",
				zap.Error(err),
				zap.String("factor_type", factor.FactorType))
			// Use original confidence if calibration fails
			calibratedConfidence = factor.Confidence
		} else {
			calibratedConfidence = calibration.CalibratedConfidence
		}

		// Update factor with calibrated confidence
		factor.Confidence = calibratedConfidence
		calibratedFactors = append(calibratedFactors, factor)
	}

	return calibratedFactors, nil
}

// calculateOverallRiskScore calculates the overall risk score from all factors
func (s *EnhancedRiskService) calculateOverallRiskScore(factors []RiskFactorDetail) float64 {
	if len(factors) == 0 {
		return 0.0
	}

	var weightedSum float64
	var totalWeight float64

	for _, factor := range factors {
		weight := factor.Weight
		if weight <= 0 {
			weight = 1.0 // Default weight
		}

		weightedSum += factor.Score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return weightedSum / totalWeight
}

// determineOverallRiskLevel determines the overall risk level from the score
func (s *EnhancedRiskService) determineOverallRiskLevel(score float64) RiskLevel {
	switch {
	case score >= 0.8:
		return RiskLevelCritical
	case score >= 0.6:
		return RiskLevelHigh
	case score >= 0.4:
		return RiskLevelMedium
	case score >= 0.2:
		return RiskLevelLow
	default:
		return RiskLevelMinimal
	}
}

// calculateOverallConfidence calculates the overall confidence score
func (s *EnhancedRiskService) calculateOverallConfidence(factors []RiskFactorDetail) float64 {
	if len(factors) == 0 {
		return 0.0
	}

	var totalConfidence float64
	for _, factor := range factors {
		totalConfidence += factor.Confidence
	}

	return totalConfidence / float64(len(factors))
}

// generateTrendSummary generates a summary of risk trends
func (s *EnhancedRiskService) generateTrendSummary(trends []RiskTrend) string {
	if len(trends) == 0 {
		return "No trend data available"
	}

	var improving, deteriorating, stable int
	for _, trend := range trends {
		switch trend.Direction {
		case TrendDirectionImproving:
			improving++
		case TrendDirectionDeteriorating:
			deteriorating++
		case TrendDirectionStable:
			stable++
		}
	}

	return fmt.Sprintf("Trends: %d improving, %d deteriorating, %d stable", improving, deteriorating, stable)
}

// getHistoricalRiskData retrieves historical risk data for trend analysis
func (s *EnhancedRiskService) getHistoricalRiskData(
	ctx context.Context,
	businessID string,
	timeRange *TimeRange,
) ([]RiskHistoryEntry, error) {
	// This would typically query a database for historical risk assessments
	// For now, return empty data - this should be implemented with actual data access
	return []RiskHistoryEntry{}, nil
}

// GetRiskFactorHistory retrieves historical data for a specific risk factor
func (s *EnhancedRiskService) GetRiskFactorHistory(
	ctx context.Context,
	businessID string,
	factorType string,
	timeRange *TimeRange,
) ([]RiskHistoryEntry, error) {
	s.logger.Info("Retrieving risk factor history",
		zap.String("business_id", businessID),
		zap.String("factor_type", factorType))

	// This would typically query a database for historical factor data
	// For now, return empty data - this should be implemented with actual data access
	return []RiskHistoryEntry{}, nil
}

// GetActiveAlerts retrieves active alerts for a business
func (s *EnhancedRiskService) GetActiveAlerts(
	ctx context.Context,
	businessID string,
) ([]AlertDetail, error) {
	s.logger.Info("Retrieving active alerts",
		zap.String("business_id", businessID))

	// This would typically query the alert system for active alerts
	// For now, return empty data - this should be implemented with actual data access
	return []AlertDetail{}, nil
}

// AcknowledgeAlert acknowledges an alert
func (s *EnhancedRiskService) AcknowledgeAlert(
	ctx context.Context,
	alertID string,
	userID string,
	notes string,
) error {
	s.logger.Info("Acknowledging alert",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID))

	// This would typically update the alert status in the database
	// For now, just log the action
	return nil
}

// ResolveAlert resolves an alert
func (s *EnhancedRiskService) ResolveAlert(
	ctx context.Context,
	alertID string,
	userID string,
	resolutionNotes string,
) error {
	s.logger.Info("Resolving alert",
		zap.String("alert_id", alertID),
		zap.String("user_id", userID))

	// This would typically update the alert status in the database
	// For now, just log the action
	return nil
}
