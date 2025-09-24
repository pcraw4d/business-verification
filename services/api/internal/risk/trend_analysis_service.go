package risk

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

// RiskTrendAnalysisService provides comprehensive risk trend analysis
type RiskTrendAnalysisService struct {
	logger        *zap.Logger
	config        *TrendAnalysisConfig
	trendAnalyzer *TrendAnalyzer
	dataStore     TrendDataStore
}

// TrendAnalysisConfig contains configuration for trend analysis
type TrendAnalysisConfig struct {
	EnableHistoricalAnalysis   bool          `json:"enable_historical_analysis"`
	EnablePredictiveAnalysis   bool          `json:"enable_predictive_analysis"`
	EnableSeasonalityAnalysis  bool          `json:"enable_seasonality_analysis"`
	EnableCorrelationAnalysis  bool          `json:"enable_correlation_analysis"`
	MinDataPointsForTrend      int           `json:"min_data_points_for_trend"`
	MinDataPointsForPrediction int           `json:"min_data_points_for_prediction"`
	TrendWindowDays            int           `json:"trend_window_days"`
	PredictionHorizonDays      int           `json:"prediction_horizon_days"`
	SeasonalityWindowDays      int           `json:"seasonality_window_days"`
	CorrelationThreshold       float64       `json:"correlation_threshold"`
	DataRetentionDays          int           `json:"data_retention_days"`
	AnalysisFrequency          time.Duration `json:"analysis_frequency"`
}

// TrendDataStore interface for storing and retrieving trend data
type TrendDataStore interface {
	StoreRiskData(ctx context.Context, data *RiskTrendData) error
	GetRiskData(ctx context.Context, businessID string, factorID string, startDate, endDate time.Time) ([]RiskTrendData, error)
	GetLatestRiskData(ctx context.Context, businessID string, factorID string) (*RiskTrendData, error)
	DeleteOldData(ctx context.Context, olderThan time.Time) error
}

// RiskTrendData represents a data point for trend analysis
type RiskTrendData struct {
	ID         string                 `json:"id"`
	BusinessID string                 `json:"business_id"`
	FactorID   string                 `json:"factor_id"`
	FactorName string                 `json:"factor_name"`
	Category   RiskCategory           `json:"category"`
	Score      float64                `json:"score"`
	Level      RiskLevel              `json:"level"`
	Confidence float64                `json:"confidence"`
	Timestamp  time.Time              `json:"timestamp"`
	Source     string                 `json:"source"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// RiskTrendAnalysisRequest represents a request for trend analysis
type RiskTrendAnalysisRequest struct {
	BusinessID          string                 `json:"business_id"`
	FactorID            string                 `json:"factor_id,omitempty"`
	Category            RiskCategory           `json:"category,omitempty"`
	StartDate           time.Time              `json:"start_date,omitempty"`
	EndDate             time.Time              `json:"end_date,omitempty"`
	IncludePredictions  bool                   `json:"include_predictions"`
	IncludeSeasonality  bool                   `json:"include_seasonality"`
	IncludeCorrelations bool                   `json:"include_correlations"`
	AnalysisOptions     map[string]interface{} `json:"analysis_options,omitempty"`
}

// RiskTrendAnalysisResponse represents the response from trend analysis
type RiskTrendAnalysisResponse struct {
	BusinessID          string                 `json:"business_id"`
	AnalysisTimestamp   time.Time              `json:"analysis_timestamp"`
	Trends              []FactorTrendAnalysis  `json:"trends"`
	OverallTrend        *OverallTrendAnalysis  `json:"overall_trend,omitempty"`
	Predictions         []FactorPrediction     `json:"predictions,omitempty"`
	SeasonalityAnalysis *SeasonalityAnalysis   `json:"seasonality_analysis,omitempty"`
	CorrelationAnalysis *CorrelationAnalysis   `json:"correlation_analysis,omitempty"`
	Summary             TrendAnalysisSummary   `json:"summary"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// FactorTrendAnalysis represents trend analysis for a specific factor
type FactorTrendAnalysis struct {
	FactorID        string          `json:"factor_id"`
	FactorName      string          `json:"factor_name"`
	Category        RiskCategory    `json:"category"`
	CurrentScore    float64         `json:"current_score"`
	CurrentLevel    RiskLevel       `json:"current_level"`
	TrendDirection  string          `json:"trend_direction"`
	TrendStrength   float64         `json:"trend_strength"`
	TrendSlope      float64         `json:"trend_slope"`
	R2Score         float64         `json:"r2_score"`
	DataPoints      int             `json:"data_points"`
	TrendConfidence float64         `json:"trend_confidence"`
	Volatility      float64         `json:"volatility"`
	MinScore        float64         `json:"min_score"`
	MaxScore        float64         `json:"max_score"`
	AvgScore        float64         `json:"avg_score"`
	ScoreRange      float64         `json:"score_range"`
	LastUpdated     time.Time       `json:"last_updated"`
	HistoricalData  []RiskTrendData `json:"historical_data,omitempty"`
}

// OverallTrendAnalysis represents overall business risk trend
type OverallTrendAnalysis struct {
	OverallDirection    string  `json:"overall_direction"`
	OverallStrength     float64 `json:"overall_strength"`
	RiskImprovement     int     `json:"risk_improvement"`
	RiskDeterioration   int     `json:"risk_deterioration"`
	RiskStable          int     `json:"risk_stable"`
	RiskVolatile        int     `json:"risk_volatile"`
	TotalFactors        int     `json:"total_factors"`
	HighRiskFactors     int     `json:"high_risk_factors"`
	CriticalRiskFactors int     `json:"critical_risk_factors"`
	TrendConfidence     float64 `json:"trend_confidence"`
}

// FactorPrediction represents a prediction for a specific factor
type FactorPrediction struct {
	FactorID          string       `json:"factor_id"`
	FactorName        string       `json:"factor_name"`
	Category          RiskCategory `json:"category"`
	PredictedScore    float64      `json:"predicted_score"`
	PredictedLevel    RiskLevel    `json:"predicted_level"`
	PredictionDate    time.Time    `json:"prediction_date"`
	Confidence        float64      `json:"confidence"`
	LowerBound        float64      `json:"lower_bound"`
	UpperBound        float64      `json:"upper_bound"`
	PredictionMethod  string       `json:"prediction_method"`
	RiskChange        float64      `json:"risk_change"`
	RiskChangePercent float64      `json:"risk_change_percent"`
}

// TrendAnalysisSummary contains summary statistics
type TrendAnalysisSummary struct {
	TotalFactorsAnalyzed int     `json:"total_factors_analyzed"`
	ImprovingFactors     int     `json:"improving_factors"`
	DecliningFactors     int     `json:"declining_factors"`
	StableFactors        int     `json:"stable_factors"`
	VolatileFactors      int     `json:"volatile_factors"`
	AvgTrendStrength     float64 `json:"avg_trend_strength"`
	AvgTrendConfidence   float64 `json:"avg_trend_confidence"`
	DataQualityScore     float64 `json:"data_quality_score"`
	AnalysisCoverage     float64 `json:"analysis_coverage"`
}

// NewRiskTrendAnalysisService creates a new trend analysis service
func NewRiskTrendAnalysisService(logger *zap.Logger, config *TrendAnalysisConfig, dataStore TrendDataStore) *RiskTrendAnalysisService {
	return &RiskTrendAnalysisService{
		logger:        logger,
		config:        config,
		trendAnalyzer: &TrendAnalyzer{logger: logger},
		dataStore:     dataStore,
	}
}

// AnalyzeTrends performs comprehensive trend analysis
func (rtas *RiskTrendAnalysisService) AnalyzeTrends(ctx context.Context, request RiskTrendAnalysisRequest) (*RiskTrendAnalysisResponse, error) {
	startTime := time.Now()

	rtas.logger.Info("Starting trend analysis",
		zap.String("business_id", request.BusinessID),
		zap.String("factor_id", request.FactorID))

	// Validate request
	if err := rtas.validateRequest(request); err != nil {
		return nil, fmt.Errorf("invalid trend analysis request: %w", err)
	}

	// Set default date range if not provided
	if request.StartDate.IsZero() {
		request.StartDate = time.Now().AddDate(0, 0, -rtas.config.TrendWindowDays)
	}
	if request.EndDate.IsZero() {
		request.EndDate = time.Now()
	}

	// Get historical data
	historicalData, err := rtas.getHistoricalData(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to get historical data: %w", err)
	}

	// Group data by factor
	factorData := rtas.groupDataByFactor(historicalData)

	// Analyze trends for each factor
	var factorTrends []FactorTrendAnalysis
	for factorID, data := range factorData {
		trend, err := rtas.analyzeFactorTrend(ctx, factorID, data)
		if err != nil {
			rtas.logger.Warn("Failed to analyze trend for factor",
				zap.String("factor_id", factorID),
				zap.Error(err))
			continue
		}
		factorTrends = append(factorTrends, trend)
	}

	// Calculate overall trend
	overallTrend := rtas.calculateOverallTrend(factorTrends)

	// Generate predictions if requested
	var predictions []FactorPrediction
	if request.IncludePredictions {
		predictions, err = rtas.generatePredictions(ctx, factorTrends)
		if err != nil {
			rtas.logger.Warn("Failed to generate predictions", zap.Error(err))
		}
	}

	// Perform seasonality analysis if requested
	var seasonalityAnalysis *SeasonalityAnalysis
	if request.IncludeSeasonality {
		seasonalityAnalysis, err = rtas.analyzeSeasonality(ctx, historicalData)
		if err != nil {
			rtas.logger.Warn("Failed to analyze seasonality", zap.Error(err))
		}
	}

	// Perform correlation analysis if requested
	var correlationAnalysis *CorrelationAnalysis
	if request.IncludeCorrelations {
		correlationAnalysis, err = rtas.analyzeCorrelations(ctx, factorTrends)
		if err != nil {
			rtas.logger.Warn("Failed to analyze correlations", zap.Error(err))
		}
	}

	// Generate summary
	summary := rtas.generateSummary(factorTrends, historicalData)

	processingTime := time.Since(startTime)

	rtas.logger.Info("Trend analysis completed",
		zap.Int("factors_analyzed", len(factorTrends)),
		zap.Duration("processing_time", processingTime))

	return &RiskTrendAnalysisResponse{
		BusinessID:          request.BusinessID,
		AnalysisTimestamp:   time.Now(),
		Trends:              factorTrends,
		OverallTrend:        overallTrend,
		Predictions:         predictions,
		SeasonalityAnalysis: seasonalityAnalysis,
		CorrelationAnalysis: correlationAnalysis,
		Summary:             summary,
	}, nil
}

// validateRequest validates the trend analysis request
func (rtas *RiskTrendAnalysisService) validateRequest(request RiskTrendAnalysisRequest) error {
	if request.BusinessID == "" {
		return fmt.Errorf("business_id is required")
	}

	if !request.StartDate.IsZero() && !request.EndDate.IsZero() {
		if request.StartDate.After(request.EndDate) {
			return fmt.Errorf("start_date cannot be after end_date")
		}
	}

	return nil
}

// getHistoricalData retrieves historical data for analysis
func (rtas *RiskTrendAnalysisService) getHistoricalData(ctx context.Context, request RiskTrendAnalysisRequest) ([]RiskTrendData, error) {
	// If specific factor requested, get data for that factor
	if request.FactorID != "" {
		return rtas.dataStore.GetRiskData(ctx, request.BusinessID, request.FactorID, request.StartDate, request.EndDate)
	}

	// Get data for all factors (this would need to be implemented in the data store)
	// For now, we'll simulate getting data for common factors
	var allData []RiskTrendData

	commonFactors := []string{
		"cash_flow_coverage",
		"debt_to_equity_ratio",
		"operational_efficiency",
		"employee_turnover_rate",
		"compliance_score",
		"security_score",
		"sentiment_score",
	}

	for _, factorID := range commonFactors {
		factorData, err := rtas.dataStore.GetRiskData(ctx, request.BusinessID, factorID, request.StartDate, request.EndDate)
		if err != nil {
			rtas.logger.Warn("Failed to get data for factor",
				zap.String("factor_id", factorID),
				zap.Error(err))
			continue
		}
		allData = append(allData, factorData...)
	}

	return allData, nil
}

// groupDataByFactor groups historical data by factor ID
func (rtas *RiskTrendAnalysisService) groupDataByFactor(data []RiskTrendData) map[string][]RiskTrendData {
	grouped := make(map[string][]RiskTrendData)

	for _, point := range data {
		grouped[point.FactorID] = append(grouped[point.FactorID], point)
	}

	// Sort each group by timestamp
	for factorID := range grouped {
		sort.Slice(grouped[factorID], func(i, j int) bool {
			return grouped[factorID][i].Timestamp.Before(grouped[factorID][j].Timestamp)
		})
	}

	return grouped
}

// analyzeFactorTrend analyzes trend for a specific factor
func (rtas *RiskTrendAnalysisService) analyzeFactorTrend(ctx context.Context, factorID string, data []RiskTrendData) (FactorTrendAnalysis, error) {
	if len(data) < rtas.config.MinDataPointsForTrend {
		return FactorTrendAnalysis{}, fmt.Errorf("insufficient data points for trend analysis: %d", len(data))
	}

	// Convert to historical data points for trend analyzer
	historicalPoints := make([]HistoricalDataPoint, len(data))
	for i, point := range data {
		historicalPoints[i] = HistoricalDataPoint{
			Timestamp: point.Timestamp,
			Value:     point.Score,
			Source:    point.Source,
			Metadata:  point.Metadata,
		}
	}

	// Perform trend analysis
	trendAnalysis, err := rtas.trendAnalyzer.AnalyzeTrend(historicalPoints)
	if err != nil {
		return FactorTrendAnalysis{}, fmt.Errorf("failed to analyze trend: %w", err)
	}

	// Calculate additional statistics
	currentPoint := data[len(data)-1]
	stats := rtas.calculateFactorStatistics(data)

	return FactorTrendAnalysis{
		FactorID:        factorID,
		FactorName:      currentPoint.FactorName,
		Category:        currentPoint.Category,
		CurrentScore:    currentPoint.Score,
		CurrentLevel:    currentPoint.Level,
		TrendDirection:  trendAnalysis.TrendDirection,
		TrendStrength:   trendAnalysis.TrendStrength,
		TrendSlope:      trendAnalysis.TrendSlope,
		R2Score:         trendAnalysis.R2Score,
		DataPoints:      trendAnalysis.DataPoints,
		TrendConfidence: trendAnalysis.TrendConfidence,
		Volatility:      stats.Volatility,
		MinScore:        stats.MinScore,
		MaxScore:        stats.MaxScore,
		AvgScore:        stats.AvgScore,
		ScoreRange:      stats.MaxScore - stats.MinScore,
		LastUpdated:     currentPoint.Timestamp,
		HistoricalData:  data,
	}, nil
}

// FactorStatistics contains statistical information about a factor
type FactorStatistics struct {
	Volatility float64
	MinScore   float64
	MaxScore   float64
	AvgScore   float64
}

// calculateFactorStatistics calculates statistical measures for a factor
func (rtas *RiskTrendAnalysisService) calculateFactorStatistics(data []RiskTrendData) FactorStatistics {
	if len(data) == 0 {
		return FactorStatistics{}
	}

	// Extract scores
	scores := make([]float64, len(data))
	for i, point := range data {
		scores[i] = point.Score
	}

	// Calculate statistics
	minScore := scores[0]
	maxScore := scores[0]
	sum := 0.0

	for _, score := range scores {
		if score < minScore {
			minScore = score
		}
		if score > maxScore {
			maxScore = score
		}
		sum += score
	}

	avgScore := sum / float64(len(scores))

	// Calculate volatility (standard deviation)
	variance := 0.0
	for _, score := range scores {
		diff := score - avgScore
		variance += diff * diff
	}
	variance /= float64(len(scores))
	volatility := math.Sqrt(variance)

	return FactorStatistics{
		Volatility: volatility,
		MinScore:   minScore,
		MaxScore:   maxScore,
		AvgScore:   avgScore,
	}
}

// calculateOverallTrend calculates overall business risk trend
func (rtas *RiskTrendAnalysisService) calculateOverallTrend(factorTrends []FactorTrendAnalysis) *OverallTrendAnalysis {
	if len(factorTrends) == 0 {
		return nil
	}

	var improving, declining, stable, volatile int
	var totalStrength, totalConfidence float64
	var highRisk, criticalRisk int

	for _, trend := range factorTrends {
		// Count trend directions
		switch trend.TrendDirection {
		case "improving":
			improving++
		case "declining":
			declining++
		case "stable":
			stable++
		case "volatile":
			volatile++
		}

		// Accumulate strength and confidence
		totalStrength += trend.TrendStrength
		totalConfidence += trend.TrendConfidence

		// Count risk levels
		if trend.CurrentLevel == RiskLevelHigh {
			highRisk++
		} else if trend.CurrentLevel == RiskLevelCritical {
			criticalRisk++
		}
	}

	// Calculate overall direction
	var overallDirection string
	if improving > declining && improving > stable && improving > volatile {
		overallDirection = "improving"
	} else if declining > improving && declining > stable && declining > volatile {
		overallDirection = "declining"
	} else if volatile > improving && volatile > declining && volatile > stable {
		overallDirection = "volatile"
	} else {
		overallDirection = "stable"
	}

	// Calculate overall strength
	overallStrength := totalStrength / float64(len(factorTrends))
	overallConfidence := totalConfidence / float64(len(factorTrends))

	return &OverallTrendAnalysis{
		OverallDirection:    overallDirection,
		OverallStrength:     overallStrength,
		RiskImprovement:     improving,
		RiskDeterioration:   declining,
		RiskStable:          stable,
		RiskVolatile:        volatile,
		TotalFactors:        len(factorTrends),
		HighRiskFactors:     highRisk,
		CriticalRiskFactors: criticalRisk,
		TrendConfidence:     overallConfidence,
	}
}

// generatePredictions generates predictions for factors
func (rtas *RiskTrendAnalysisService) generatePredictions(ctx context.Context, factorTrends []FactorTrendAnalysis) ([]FactorPrediction, error) {
	var predictions []FactorPrediction

	for _, trend := range factorTrends {
		if trend.DataPoints < rtas.config.MinDataPointsForPrediction {
			continue
		}

		// Convert to historical data points
		historicalPoints := make([]HistoricalDataPoint, len(trend.HistoricalData))
		for i, point := range trend.HistoricalData {
			historicalPoints[i] = HistoricalDataPoint{
				Timestamp: point.Timestamp,
				Value:     point.Score,
				Source:    point.Source,
				Metadata:  point.Metadata,
			}
		}

		// Analyze trend to get projection
		trendAnalysis, err := rtas.trendAnalyzer.AnalyzeTrend(historicalPoints)
		if err != nil {
			continue
		}

		if trendAnalysis.ProjectedValue == 0 {
			continue
		}

		// Calculate risk change
		riskChange := trendAnalysis.ProjectedValue - trend.CurrentScore
		riskChangePercent := (riskChange / trend.CurrentScore) * 100

		// Determine predicted level
		predictedLevel := rtas.determineRiskLevel(trendAnalysis.ProjectedValue)

		prediction := FactorPrediction{
			FactorID:          trend.FactorID,
			FactorName:        trend.FactorName,
			Category:          trend.Category,
			PredictedScore:    trendAnalysis.ProjectedValue,
			PredictedLevel:    predictedLevel,
			PredictionDate:    time.Now().AddDate(0, 0, rtas.config.PredictionHorizonDays),
			Confidence:        trendAnalysis.ProjectionConfidence,
			LowerBound:        trendAnalysis.ProjectedValue - (trend.Volatility * 2),
			UpperBound:        trendAnalysis.ProjectedValue + (trend.Volatility * 2),
			PredictionMethod:  "linear_regression",
			RiskChange:        riskChange,
			RiskChangePercent: riskChangePercent,
		}

		predictions = append(predictions, prediction)
	}

	return predictions, nil
}

// determineRiskLevel determines risk level based on score
func (rtas *RiskTrendAnalysisService) determineRiskLevel(score float64) RiskLevel {
	if score <= 25 {
		return RiskLevelLow
	} else if score <= 50 {
		return RiskLevelMedium
	} else if score <= 75 {
		return RiskLevelHigh
	} else {
		return RiskLevelCritical
	}
}

// analyzeSeasonality analyzes seasonality patterns
func (rtas *RiskTrendAnalysisService) analyzeSeasonality(ctx context.Context, data []RiskTrendData) (*SeasonalityAnalysis, error) {
	// Convert to historical data points
	historicalPoints := make([]HistoricalDataPoint, len(data))
	for i, point := range data {
		historicalPoints[i] = HistoricalDataPoint{
			Timestamp: point.Timestamp,
			Value:     point.Score,
			Source:    point.Source,
			Metadata:  point.Metadata,
		}
	}

	return rtas.trendAnalyzer.DetectSeasonality(historicalPoints)
}

// analyzeCorrelations analyzes correlations between factors
func (rtas *RiskTrendAnalysisService) analyzeCorrelations(ctx context.Context, factorTrends []FactorTrendAnalysis) (*CorrelationAnalysis, error) {
	// This would require more sophisticated correlation analysis
	// For now, return a basic analysis
	return &CorrelationAnalysis{
		CorrelatedFactors:   []CorrelatedFactor{},
		MaxCorrelation:      0.0,
		AvgCorrelation:      0.0,
		CorrelationStrength: "none",
	}, nil
}

// generateSummary generates summary statistics
func (rtas *RiskTrendAnalysisService) generateSummary(factorTrends []FactorTrendAnalysis, historicalData []RiskTrendData) TrendAnalysisSummary {
	var improving, declining, stable, volatile int
	var totalStrength, totalConfidence float64

	for _, trend := range factorTrends {
		switch trend.TrendDirection {
		case "improving":
			improving++
		case "declining":
			declining++
		case "stable":
			stable++
		case "volatile":
			volatile++
		}

		totalStrength += trend.TrendStrength
		totalConfidence += trend.TrendConfidence
	}

	// Calculate data quality score
	dataQualityScore := rtas.calculateDataQualityScore(historicalData)

	// Calculate analysis coverage
	analysisCoverage := float64(len(factorTrends)) / float64(len(historicalData)) * 100

	return TrendAnalysisSummary{
		TotalFactorsAnalyzed: len(factorTrends),
		ImprovingFactors:     improving,
		DecliningFactors:     declining,
		StableFactors:        stable,
		VolatileFactors:      volatile,
		AvgTrendStrength:     totalStrength / float64(len(factorTrends)),
		AvgTrendConfidence:   totalConfidence / float64(len(factorTrends)),
		DataQualityScore:     dataQualityScore,
		AnalysisCoverage:     analysisCoverage,
	}
}

// calculateDataQualityScore calculates the quality of the data
func (rtas *RiskTrendAnalysisService) calculateDataQualityScore(data []RiskTrendData) float64 {
	if len(data) == 0 {
		return 0.0
	}

	// Calculate completeness (non-zero values)
	completeData := 0
	for _, point := range data {
		if point.Score > 0 && point.Confidence > 0 {
			completeData++
		}
	}

	completeness := float64(completeData) / float64(len(data))

	// Calculate recency (data within last 30 days)
	recentData := 0
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	for _, point := range data {
		if point.Timestamp.After(thirtyDaysAgo) {
			recentData++
		}
	}

	recency := float64(recentData) / float64(len(data))

	// Calculate consistency (low volatility)
	avgConfidence := 0.0
	for _, point := range data {
		avgConfidence += point.Confidence
	}
	avgConfidence /= float64(len(data))

	// Overall quality score
	qualityScore := (completeness * 0.4) + (recency * 0.3) + (avgConfidence * 0.3)

	return qualityScore * 100
}

// StoreRiskData stores risk data for trend analysis
func (rtas *RiskTrendAnalysisService) StoreRiskData(ctx context.Context, data *RiskTrendData) error {
	return rtas.dataStore.StoreRiskData(ctx, data)
}

// GetLatestRiskData gets the latest risk data for a factor
func (rtas *RiskTrendAnalysisService) GetLatestRiskData(ctx context.Context, businessID, factorID string) (*RiskTrendData, error) {
	return rtas.dataStore.GetLatestRiskData(ctx, businessID, factorID)
}

// CleanupOldData removes old data beyond retention period
func (rtas *RiskTrendAnalysisService) CleanupOldData(ctx context.Context) error {
	cutoffDate := time.Now().AddDate(0, 0, -rtas.config.DataRetentionDays)
	return rtas.dataStore.DeleteOldData(ctx, cutoffDate)
}
