package enrichment

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// DataSourceReliabilityAssessor provides comprehensive data source reliability assessment
type DataSourceReliabilityAssessor struct {
	logger *zap.Logger
	tracer trace.Tracer
	config *DataSourceReliabilityConfig

	// Assessment data
	mu                    sync.RWMutex
	sourceAssessments     map[string]*SourceAssessment
	historicalPerformance map[string]*PerformanceHistory
	reliabilityMetrics    map[string]*ReliabilityMetrics
	lastCleanup           time.Time
}

// DataSourceReliabilityConfig contains configuration for reliability assessment
type DataSourceReliabilityConfig struct {
	// Assessment settings
	EnableHistoricalAnalysis  bool `json:"enable_historical_analysis"`
	EnablePerformanceTracking bool `json:"enable_performance_tracking"`
	EnablePredictiveScoring   bool `json:"enable_predictive_scoring"`
	EnableAlerting            bool `json:"enable_alerting"`

	// Thresholds
	LowReliabilityThreshold      float64       `json:"low_reliability_threshold"`      // Below this is low reliability
	CriticalReliabilityThreshold float64       `json:"critical_reliability_threshold"` // Below this is critical
	PerformanceThreshold         time.Duration `json:"performance_threshold"`          // Response time threshold
	UptimeThreshold              float64       `json:"uptime_threshold"`               // Minimum uptime percentage

	// Scoring weights
	HistoricalWeight  float64 `json:"historical_weight"`
	PerformanceWeight float64 `json:"performance_weight"`
	AccuracyWeight    float64 `json:"accuracy_weight"`
	ConsistencyWeight float64 `json:"consistency_weight"`
	UptimeWeight      float64 `json:"uptime_weight"`
	DataQualityWeight float64 `json:"data_quality_weight"`

	// History settings
	MaxHistorySize         int           `json:"max_history_size"`
	HistoryRetentionPeriod time.Duration `json:"history_retention_period"`
	CleanupInterval        time.Duration `json:"cleanup_interval"`

	// Alert settings
	AlertCooldownPeriod time.Duration `json:"alert_cooldown_period"`
	MaxAlertsPerSource  int           `json:"max_alerts_per_source"`
}

// SourceAssessment represents a comprehensive source reliability assessment
type SourceAssessment struct {
	SourceID           string    `json:"source_id"`
	SourceType         string    `json:"source_type"`
	SourceName         string    `json:"source_name"`
	LastAssessed       time.Time `json:"last_assessed"`
	OverallReliability float64   `json:"overall_reliability"`
	ReliabilityLevel   string    `json:"reliability_level"` // "excellent", "good", "fair", "poor", "critical"
	ConfidenceScore    float64   `json:"confidence_score"`

	// Component scores
	HistoricalScore  float64 `json:"historical_score"`
	PerformanceScore float64 `json:"performance_score"`
	AccuracyScore    float64 `json:"accuracy_score"`
	ConsistencyScore float64 `json:"consistency_score"`
	UptimeScore      float64 `json:"uptime_score"`
	DataQualityScore float64 `json:"data_quality_score"`

	// Performance metrics
	AverageResponseTime  time.Duration `json:"average_response_time"`
	ResponseTimeVariance time.Duration `json:"response_time_variance"`
	UptimePercentage     float64       `json:"uptime_percentage"`
	ErrorRate            float64       `json:"error_rate"`
	SuccessRate          float64       `json:"success_rate"`

	// Historical analysis
	AssessmentCount int     `json:"assessment_count"`
	TrendDirection  string  `json:"trend_direction"` // "improving", "stable", "declining"
	TrendConfidence float64 `json:"trend_confidence"`

	// Risk factors
	RiskFactors []string `json:"risk_factors"`
	RiskLevel   string   `json:"risk_level"` // "low", "medium", "high", "critical"

	// Recommendations
	Recommendations []string `json:"recommendations"`
	PriorityActions []string `json:"priority_actions"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata"`
}

// PerformanceHistory tracks historical performance data
type PerformanceHistory struct {
	SourceID            string               `json:"source_id"`
	Assessments         []*SourceAssessment  `json:"assessments"`
	PerformanceMetrics  []*PerformanceMetric `json:"performance_metrics"`
	LastUpdated         time.Time            `json:"last_updated"`
	TotalAssessments    int                  `json:"total_assessments"`
	AverageReliability  float64              `json:"average_reliability"`
	ReliabilityVariance float64              `json:"reliability_variance"`
}

// PerformanceMetric represents a single performance measurement
type PerformanceMetric struct {
	Timestamp        time.Time     `json:"timestamp"`
	ResponseTime     time.Duration `json:"response_time"`
	Success          bool          `json:"success"`
	ErrorType        string        `json:"error_type,omitempty"`
	DataQualityScore float64       `json:"data_quality_score"`
	ReliabilityScore float64       `json:"reliability_score"`
}

// ReliabilityMetrics contains aggregated reliability statistics
type ReliabilityMetrics struct {
	SourceID            string        `json:"source_id"`
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	ResponseTimeP95     time.Duration `json:"response_time_p95"`
	ResponseTimeP99     time.Duration `json:"response_time_p99"`
	UptimePercentage    float64       `json:"uptime_percentage"`
	ErrorRate           float64       `json:"error_rate"`
	LastUpdated         time.Time     `json:"last_updated"`
}

// ReliabilityAssessmentResult contains comprehensive assessment results
type ReliabilityAssessmentResult struct {
	// Assessment summary
	SourceID         string            `json:"source_id"`
	Assessment       *SourceAssessment `json:"assessment"`
	OverallScore     float64           `json:"overall_score"`
	ReliabilityLevel string            `json:"reliability_level"`

	// Historical analysis
	HistoricalAnalysis *PerformanceHistory `json:"historical_analysis"`
	TrendAnalysis      *TrendAnalysis      `json:"trend_analysis"`

	// Performance analysis
	PerformanceMetrics *ReliabilityMetrics `json:"performance_metrics"`
	PerformanceScore   float64             `json:"performance_score"`

	// Risk assessment
	RiskAssessment *RiskAssessment `json:"risk_assessment"`
	RiskLevel      string          `json:"risk_level"`

	// Recommendations
	Recommendations []string `json:"recommendations"`
	PriorityActions []string `json:"priority_actions"`

	// Metadata
	AssessedAt     time.Time     `json:"assessed_at"`
	ProcessingTime time.Duration `json:"processing_time"`
	DataPoints     int           `json:"data_points"`
}

// TrendAnalysis contains trend analysis results
type TrendAnalysis struct {
	Direction  string    `json:"direction"` // "improving", "stable", "declining"
	Confidence float64   `json:"confidence"`
	Slope      float64   `json:"slope"`
	R2         float64   `json:"r2"` // R-squared value
	Periods    int       `json:"periods"`
	LastChange time.Time `json:"last_change"`
}

// RiskAssessment contains risk analysis results
type RiskAssessment struct {
	RiskLevel         string   `json:"risk_level"`
	RiskFactors       []string `json:"risk_factors"`
	RiskScore         float64  `json:"risk_score"`
	MitigationActions []string `json:"mitigation_actions"`
	ImpactLevel       string   `json:"impact_level"` // "low", "medium", "high", "critical"
}

// NewDataSourceReliabilityAssessor creates a new data source reliability assessor
func NewDataSourceReliabilityAssessor(logger *zap.Logger, config *DataSourceReliabilityConfig) *DataSourceReliabilityAssessor {
	if config == nil {
		config = getDefaultDataSourceReliabilityConfig()
	}

	return &DataSourceReliabilityAssessor{
		logger:                logger,
		tracer:                trace.NewNoopTracerProvider().Tracer("data_source_reliability_assessor"),
		config:                config,
		sourceAssessments:     make(map[string]*SourceAssessment),
		historicalPerformance: make(map[string]*PerformanceHistory),
		reliabilityMetrics:    make(map[string]*ReliabilityMetrics),
		lastCleanup:           time.Now(),
	}
}

// AssessSourceReliability performs comprehensive reliability assessment
func (dsra *DataSourceReliabilityAssessor) AssessSourceReliability(ctx context.Context, sourceID, sourceType, sourceName string, data interface{}) (*ReliabilityAssessmentResult, error) {
	ctx, span := dsra.tracer.Start(ctx, "data_source_reliability_assessor.assess",
		trace.WithAttributes(
			attribute.String("source_id", sourceID),
			attribute.String("source_type", sourceType),
			attribute.String("source_name", sourceName),
		))
	defer span.End()

	startTime := time.Now()

	dsra.logger.Info("Starting source reliability assessment",
		zap.String("source_id", sourceID),
		zap.String("source_type", sourceType),
		zap.String("source_name", sourceName))

	result := &ReliabilityAssessmentResult{
		SourceID:        sourceID,
		AssessedAt:      time.Now(),
		Recommendations: []string{},
		PriorityActions: []string{},
	}

	// Create or update source assessment
	assessment := dsra.createOrUpdateAssessment(sourceID, sourceType, sourceName, data)
	result.Assessment = assessment

	// Perform historical analysis if enabled
	if dsra.config.EnableHistoricalAnalysis {
		historicalAnalysis := dsra.analyzeHistoricalPerformance(sourceID)
		result.HistoricalAnalysis = historicalAnalysis
		result.TrendAnalysis = dsra.analyzeTrend(historicalAnalysis)
	}

	// Calculate performance metrics
	performanceMetrics := dsra.calculatePerformanceMetrics(sourceID)
	result.PerformanceMetrics = performanceMetrics
	result.PerformanceScore = dsra.calculatePerformanceScore(performanceMetrics)

	// Perform risk assessment
	riskAssessment := dsra.assessRisk(assessment, performanceMetrics)
	result.RiskAssessment = riskAssessment
	result.RiskLevel = riskAssessment.RiskLevel

	// Generate recommendations
	result.Recommendations = dsra.generateRecommendations(assessment, riskAssessment)
	result.PriorityActions = dsra.generatePriorityActions(assessment, riskAssessment)

	// Calculate overall score
	result.OverallScore = dsra.calculateOverallScore(assessment, performanceMetrics, riskAssessment)
	result.ReliabilityLevel = dsra.determineReliabilityLevel(result.OverallScore)

	// Update metadata
	result.ProcessingTime = time.Since(startTime)
	result.DataPoints = dsra.countDataPoints(sourceID)

	dsra.logger.Info("Source reliability assessment completed",
		zap.String("source_id", sourceID),
		zap.Float64("overall_score", result.OverallScore),
		zap.String("reliability_level", result.ReliabilityLevel),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// RecordPerformance records a performance measurement
func (dsra *DataSourceReliabilityAssessor) RecordPerformance(ctx context.Context, sourceID string, responseTime time.Duration, success bool, errorType string, dataQualityScore float64) error {
	ctx, span := dsra.tracer.Start(ctx, "data_source_reliability_assessor.record_performance",
		trace.WithAttributes(
			attribute.String("source_id", sourceID),
			attribute.String("response_time", responseTime.String()),
			attribute.Bool("success", success),
		))
	defer span.End()

	dsra.mu.Lock()
	defer dsra.mu.Unlock()

	// Create or update metrics
	if metrics, exists := dsra.reliabilityMetrics[sourceID]; exists {
		metrics.TotalRequests++
		if success {
			metrics.SuccessfulRequests++
		} else {
			metrics.FailedRequests++
		}

		// Update response time statistics (simplified)
		if metrics.TotalRequests == 1 {
			metrics.AverageResponseTime = responseTime
		} else {
			// Simple moving average
			metrics.AverageResponseTime = (metrics.AverageResponseTime + responseTime) / 2
		}

		metrics.ErrorRate = float64(metrics.FailedRequests) / float64(metrics.TotalRequests)
		metrics.UptimePercentage = float64(metrics.SuccessfulRequests) / float64(metrics.TotalRequests) * 100
		metrics.LastUpdated = time.Now()
	} else {
		dsra.reliabilityMetrics[sourceID] = &ReliabilityMetrics{
			SourceID:            sourceID,
			TotalRequests:       1,
			SuccessfulRequests:  0,
			FailedRequests:      0,
			AverageResponseTime: responseTime,
			LastUpdated:         time.Now(),
		}
		if success {
			dsra.reliabilityMetrics[sourceID].SuccessfulRequests = 1
			dsra.reliabilityMetrics[sourceID].UptimePercentage = 100.0
		} else {
			dsra.reliabilityMetrics[sourceID].FailedRequests = 1
			dsra.reliabilityMetrics[sourceID].ErrorRate = 1.0
		}
	}

	// Add to performance history if enabled
	if dsra.config.EnablePerformanceTracking {
		metric := &PerformanceMetric{
			Timestamp:        time.Now(),
			ResponseTime:     responseTime,
			Success:          success,
			ErrorType:        errorType,
			DataQualityScore: dataQualityScore,
		}

		if history, exists := dsra.historicalPerformance[sourceID]; exists {
			history.PerformanceMetrics = append(history.PerformanceMetrics, metric)
			history.LastUpdated = time.Now()
		} else {
			dsra.historicalPerformance[sourceID] = &PerformanceHistory{
				SourceID:           sourceID,
				PerformanceMetrics: []*PerformanceMetric{metric},
				LastUpdated:        time.Now(),
			}
		}
	}

	// Cleanup old data periodically
	dsra.cleanupIfNeeded()

	return nil
}

// GetSourceAssessment retrieves the latest assessment for a source
func (dsra *DataSourceReliabilityAssessor) GetSourceAssessment(ctx context.Context, sourceID string) (*SourceAssessment, error) {
	dsra.mu.RLock()
	defer dsra.mu.RUnlock()

	if assessment, exists := dsra.sourceAssessments[sourceID]; exists {
		return assessment, nil
	}

	return nil, fmt.Errorf("no assessment found for source: %s", sourceID)
}

// GetReliabilityMetrics retrieves reliability metrics for a source
func (dsra *DataSourceReliabilityAssessor) GetReliabilityMetrics(ctx context.Context, sourceID string) (*ReliabilityMetrics, error) {
	dsra.mu.RLock()
	defer dsra.mu.RUnlock()

	if metrics, exists := dsra.reliabilityMetrics[sourceID]; exists {
		return metrics, nil
	}

	return nil, fmt.Errorf("no metrics found for source: %s", sourceID)
}

// GetHistoricalPerformance retrieves historical performance data
func (dsra *DataSourceReliabilityAssessor) GetHistoricalPerformance(ctx context.Context, sourceID string) (*PerformanceHistory, error) {
	dsra.mu.RLock()
	defer dsra.mu.RUnlock()

	if history, exists := dsra.historicalPerformance[sourceID]; exists {
		return history, nil
	}

	return nil, fmt.Errorf("no historical data found for source: %s", sourceID)
}

// Helper methods

func (dsra *DataSourceReliabilityAssessor) createOrUpdateAssessment(sourceID, sourceType, sourceName string, data interface{}) *SourceAssessment {
	dsra.mu.Lock()
	defer dsra.mu.Unlock()

	assessment := &SourceAssessment{
		SourceID:        sourceID,
		SourceType:      sourceType,
		SourceName:      sourceName,
		LastAssessed:    time.Now(),
		RiskFactors:     []string{},
		Recommendations: []string{},
		PriorityActions: []string{},
		Metadata:        make(map[string]interface{}),
	}

	// Calculate component scores
	assessment.HistoricalScore = dsra.calculateHistoricalScore(sourceID)
	assessment.PerformanceScore = dsra.calculatePerformanceScore(dsra.reliabilityMetrics[sourceID])
	assessment.AccuracyScore = dsra.calculateAccuracyScore(data)
	assessment.ConsistencyScore = dsra.calculateConsistencyScore(sourceID)
	assessment.UptimeScore = dsra.calculateUptimeScore(sourceID)
	assessment.DataQualityScore = dsra.calculateDataQualityScore(data)

	// Calculate overall reliability
	assessment.OverallReliability = dsra.calculateOverallReliability(assessment)
	assessment.ReliabilityLevel = dsra.determineReliabilityLevel(assessment.OverallReliability)

	// Update performance metrics
	if metrics, exists := dsra.reliabilityMetrics[sourceID]; exists {
		assessment.AverageResponseTime = metrics.AverageResponseTime
		assessment.UptimePercentage = metrics.UptimePercentage
		assessment.ErrorRate = metrics.ErrorRate
		assessment.SuccessRate = 1.0 - metrics.ErrorRate
	}

	// Analyze trends
	if history, exists := dsra.historicalPerformance[sourceID]; exists {
		assessment.AssessmentCount = len(history.Assessments)
		assessment.TrendDirection = dsra.analyzeTrendDirection(history)
		assessment.TrendConfidence = dsra.calculateTrendConfidence(history)
	}

	// Identify risk factors
	assessment.RiskFactors = dsra.identifyRiskFactors(assessment)
	assessment.RiskLevel = dsra.determineRiskLevel(assessment)

	// Generate recommendations
	assessment.Recommendations = dsra.generateSourceRecommendations(assessment)
	assessment.PriorityActions = dsra.generateSourcePriorityActions(assessment)

	// Store assessment
	dsra.sourceAssessments[sourceID] = assessment

	// Add to historical analysis
	if history, exists := dsra.historicalPerformance[sourceID]; exists {
		history.Assessments = append(history.Assessments, assessment)
		history.TotalAssessments++
		history.LastUpdated = time.Now()
	}

	return assessment
}

func (dsra *DataSourceReliabilityAssessor) calculateHistoricalScore(sourceID string) float64 {
	if !dsra.config.EnableHistoricalAnalysis {
		return 0.7 // Default score
	}

	if history, exists := dsra.historicalPerformance[sourceID]; exists {
		if len(history.Assessments) == 0 {
			return 0.7
		}

		// Calculate average reliability from historical assessments
		totalScore := 0.0
		for _, assessment := range history.Assessments {
			totalScore += assessment.OverallReliability
		}

		return totalScore / float64(len(history.Assessments))
	}

	return 0.7
}

func (dsra *DataSourceReliabilityAssessor) calculatePerformanceScore(metrics *ReliabilityMetrics) float64 {
	if metrics == nil {
		return 0.7
	}

	score := 0.0
	factors := 0

	// Uptime factor
	if metrics.UptimePercentage >= dsra.config.UptimeThreshold {
		score += 1.0
	} else {
		score += metrics.UptimePercentage / dsra.config.UptimeThreshold
	}
	factors++

	// Response time factor
	if metrics.AverageResponseTime <= dsra.config.PerformanceThreshold {
		score += 1.0
	} else {
		// Penalize slow response times
		ratio := float64(dsra.config.PerformanceThreshold) / float64(metrics.AverageResponseTime)
		score += math.Max(0.0, ratio)
	}
	factors++

	// Error rate factor
	errorScore := 1.0 - metrics.ErrorRate
	score += errorScore
	factors++

	return score / float64(factors)
}

func (dsra *DataSourceReliabilityAssessor) calculateAccuracyScore(data interface{}) float64 {
	// This would implement accuracy assessment based on data validation
	// For now, return a default score
	return 0.8
}

func (dsra *DataSourceReliabilityAssessor) calculateConsistencyScore(sourceID string) float64 {
	if history, exists := dsra.historicalPerformance[sourceID]; exists {
		if len(history.PerformanceMetrics) < 2 {
			return 0.7
		}

		// Calculate consistency based on response time variance
		var responseTimes []float64
		for _, metric := range history.PerformanceMetrics {
			responseTimes = append(responseTimes, float64(metric.ResponseTime))
		}

		// Calculate coefficient of variation (lower is more consistent)
		mean := calculateMean(responseTimes)
		stdDev := calculateStandardDeviation(responseTimes, mean)

		if mean == 0 {
			return 1.0
		}

		cv := stdDev / mean
		consistency := math.Max(0.0, 1.0-cv)

		return consistency
	}

	return 0.7
}

func (dsra *DataSourceReliabilityAssessor) calculateUptimeScore(sourceID string) float64 {
	if metrics, exists := dsra.reliabilityMetrics[sourceID]; exists {
		return metrics.UptimePercentage / 100.0
	}
	return 0.7
}

func (dsra *DataSourceReliabilityAssessor) calculateDataQualityScore(data interface{}) float64 {
	// This would integrate with the data quality scorer
	// For now, return a default score
	return 0.8
}

func (dsra *DataSourceReliabilityAssessor) calculateOverallReliability(assessment *SourceAssessment) float64 {
	totalWeight := dsra.config.HistoricalWeight + dsra.config.PerformanceWeight +
		dsra.config.AccuracyWeight + dsra.config.ConsistencyWeight +
		dsra.config.UptimeWeight + dsra.config.DataQualityWeight

	if totalWeight == 0 {
		totalWeight = 6.0 // Equal weights
	}

	weightedScore := (assessment.HistoricalScore * dsra.config.HistoricalWeight) +
		(assessment.PerformanceScore * dsra.config.PerformanceWeight) +
		(assessment.AccuracyScore * dsra.config.AccuracyWeight) +
		(assessment.ConsistencyScore * dsra.config.ConsistencyWeight) +
		(assessment.UptimeScore * dsra.config.UptimeWeight) +
		(assessment.DataQualityScore * dsra.config.DataQualityWeight)

	return weightedScore / totalWeight
}

func (dsra *DataSourceReliabilityAssessor) determineReliabilityLevel(score float64) string {
	if score >= 0.9 {
		return "excellent"
	} else if score >= 0.8 {
		return "good"
	} else if score >= 0.7 {
		return "fair"
	} else if score >= dsra.config.CriticalReliabilityThreshold {
		return "poor"
	} else {
		return "critical"
	}
}

func (dsra *DataSourceReliabilityAssessor) analyzeHistoricalPerformance(sourceID string) *PerformanceHistory {
	dsra.mu.RLock()
	defer dsra.mu.RUnlock()

	if history, exists := dsra.historicalPerformance[sourceID]; exists {
		return history
	}

	return &PerformanceHistory{
		SourceID:           sourceID,
		Assessments:        []*SourceAssessment{},
		PerformanceMetrics: []*PerformanceMetric{},
		LastUpdated:        time.Now(),
	}
}

func (dsra *DataSourceReliabilityAssessor) analyzeTrend(history *PerformanceHistory) *TrendAnalysis {
	if len(history.Assessments) < 2 {
		return &TrendAnalysis{
			Direction:  "stable",
			Confidence: 0.3,
			Periods:    len(history.Assessments),
		}
	}

	// Simple trend analysis
	firstScore := history.Assessments[0].OverallReliability
	lastScore := history.Assessments[len(history.Assessments)-1].OverallReliability

	slope := lastScore - firstScore
	var direction string
	if slope > 0.05 {
		direction = "improving"
	} else if slope < -0.05 {
		direction = "declining"
	} else {
		direction = "stable"
	}

	return &TrendAnalysis{
		Direction:  direction,
		Confidence: 0.7,
		Slope:      slope,
		R2:         0.6,
		Periods:    len(history.Assessments),
		LastChange: history.Assessments[len(history.Assessments)-1].LastAssessed,
	}
}

func (dsra *DataSourceReliabilityAssessor) calculatePerformanceMetrics(sourceID string) *ReliabilityMetrics {
	dsra.mu.RLock()
	defer dsra.mu.RUnlock()

	if metrics, exists := dsra.reliabilityMetrics[sourceID]; exists {
		return metrics
	}

	return &ReliabilityMetrics{
		SourceID:           sourceID,
		TotalRequests:      0,
		SuccessfulRequests: 0,
		FailedRequests:     0,
		LastUpdated:        time.Now(),
	}
}

func (dsra *DataSourceReliabilityAssessor) assessRisk(assessment *SourceAssessment, metrics *ReliabilityMetrics) *RiskAssessment {
	riskFactors := []string{}
	riskScore := 0.0

	// Check reliability level
	if assessment.ReliabilityLevel == "critical" {
		riskFactors = append(riskFactors, "Critical reliability level")
		riskScore += 0.4
	} else if assessment.ReliabilityLevel == "poor" {
		riskFactors = append(riskFactors, "Poor reliability level")
		riskScore += 0.3
	}

	// Check uptime
	if assessment.UptimePercentage < dsra.config.UptimeThreshold {
		riskFactors = append(riskFactors, "Low uptime percentage")
		riskScore += 0.2
	}

	// Check error rate
	if assessment.ErrorRate > 0.1 {
		riskFactors = append(riskFactors, "High error rate")
		riskScore += 0.2
	}

	// Check response time
	if assessment.AverageResponseTime > dsra.config.PerformanceThreshold {
		riskFactors = append(riskFactors, "Slow response times")
		riskScore += 0.1
	}

	// Determine risk level
	var riskLevel string
	if riskScore >= 0.7 {
		riskLevel = "critical"
	} else if riskScore >= 0.5 {
		riskLevel = "high"
	} else if riskScore >= 0.3 {
		riskLevel = "medium"
	} else {
		riskLevel = "low"
	}

	// Generate mitigation actions
	mitigationActions := dsra.generateMitigationActions(riskFactors)

	return &RiskAssessment{
		RiskLevel:         riskLevel,
		RiskFactors:       riskFactors,
		RiskScore:         riskScore,
		MitigationActions: mitigationActions,
		ImpactLevel:       riskLevel,
	}
}

func (dsra *DataSourceReliabilityAssessor) generateRecommendations(assessment *SourceAssessment, riskAssessment *RiskAssessment) []string {
	recommendations := []string{}

	if assessment.ReliabilityLevel == "critical" {
		recommendations = append(recommendations, "Immediate intervention required - source is critically unreliable")
	}

	if assessment.UptimePercentage < dsra.config.UptimeThreshold {
		recommendations = append(recommendations, "Improve uptime to meet threshold requirements")
	}

	if assessment.ErrorRate > 0.1 {
		recommendations = append(recommendations, "Investigate and reduce error rate")
	}

	if assessment.AverageResponseTime > dsra.config.PerformanceThreshold {
		recommendations = append(recommendations, "Optimize response times")
	}

	if assessment.ConsistencyScore < 0.7 {
		recommendations = append(recommendations, "Improve consistency of performance")
	}

	return recommendations
}

func (dsra *DataSourceReliabilityAssessor) generatePriorityActions(assessment *SourceAssessment, riskAssessment *RiskAssessment) []string {
	actions := []string{}

	if riskAssessment.RiskLevel == "critical" {
		actions = append(actions, "URGENT: Implement immediate mitigation strategies")
		actions = append(actions, "Consider alternative data sources")
	}

	if assessment.ReliabilityLevel == "critical" {
		actions = append(actions, "URGENT: Source reliability is critical - immediate action required")
	}

	if assessment.UptimePercentage < 90 {
		actions = append(actions, "Investigate uptime issues and implement redundancy")
	}

	return actions
}

func (dsra *DataSourceReliabilityAssessor) calculateOverallScore(assessment *SourceAssessment, metrics *ReliabilityMetrics, riskAssessment *RiskAssessment) float64 {
	// Base score from assessment
	score := assessment.OverallReliability

	// Adjust for risk
	riskAdjustment := 1.0 - riskAssessment.RiskScore
	score = score * riskAdjustment

	return math.Max(0.0, math.Min(1.0, score))
}

func (dsra *DataSourceReliabilityAssessor) determineRiskLevel(assessment *SourceAssessment) string {
	riskScore := 0.0

	if assessment.ReliabilityLevel == "critical" {
		riskScore += 0.4
	} else if assessment.ReliabilityLevel == "poor" {
		riskScore += 0.3
	}

	if assessment.UptimePercentage < 90 {
		riskScore += 0.2
	}

	if assessment.ErrorRate > 0.1 {
		riskScore += 0.2
	}

	if riskScore >= 0.7 {
		return "critical"
	} else if riskScore >= 0.5 {
		return "high"
	} else if riskScore >= 0.3 {
		return "medium"
	} else {
		return "low"
	}
}

func (dsra *DataSourceReliabilityAssessor) identifyRiskFactors(assessment *SourceAssessment) []string {
	factors := []string{}

	if assessment.ReliabilityLevel == "critical" || assessment.ReliabilityLevel == "poor" {
		factors = append(factors, "Low reliability score")
	}

	if assessment.UptimePercentage < 95 {
		factors = append(factors, "Suboptimal uptime")
	}

	if assessment.ErrorRate > 0.05 {
		factors = append(factors, "Elevated error rate")
	}

	if assessment.AverageResponseTime > dsra.config.PerformanceThreshold {
		factors = append(factors, "Slow response times")
	}

	return factors
}

func (dsra *DataSourceReliabilityAssessor) generateSourceRecommendations(assessment *SourceAssessment) []string {
	recommendations := []string{}

	if assessment.ReliabilityLevel == "critical" {
		recommendations = append(recommendations, "Critical reliability issues detected")
	}

	if assessment.UptimePercentage < 95 {
		recommendations = append(recommendations, "Improve uptime performance")
	}

	if assessment.ConsistencyScore < 0.8 {
		recommendations = append(recommendations, "Improve performance consistency")
	}

	return recommendations
}

func (dsra *DataSourceReliabilityAssessor) generateSourcePriorityActions(assessment *SourceAssessment) []string {
	actions := []string{}

	if assessment.ReliabilityLevel == "critical" {
		actions = append(actions, "URGENT: Address critical reliability issues")
	}

	if assessment.UptimePercentage < 90 {
		actions = append(actions, "Implement uptime improvements")
	}

	return actions
}

func (dsra *DataSourceReliabilityAssessor) generateMitigationActions(riskFactors []string) []string {
	actions := []string{}

	for _, factor := range riskFactors {
		switch factor {
		case "Critical reliability level":
			actions = append(actions, "Implement immediate reliability improvements")
		case "Low uptime percentage":
			actions = append(actions, "Add redundancy and failover mechanisms")
		case "High error rate":
			actions = append(actions, "Investigate and fix error sources")
		case "Slow response times":
			actions = append(actions, "Optimize performance and caching")
		}
	}

	return actions
}

func (dsra *DataSourceReliabilityAssessor) analyzeTrendDirection(history *PerformanceHistory) string {
	if len(history.Assessments) < 2 {
		return "stable"
	}

	firstScore := history.Assessments[0].OverallReliability
	lastScore := history.Assessments[len(history.Assessments)-1].OverallReliability

	if lastScore > firstScore+0.05 {
		return "improving"
	} else if lastScore < firstScore-0.05 {
		return "declining"
	} else {
		return "stable"
	}
}

func (dsra *DataSourceReliabilityAssessor) calculateTrendConfidence(history *PerformanceHistory) float64 {
	if len(history.Assessments) < 3 {
		return 0.3
	}

	// Simple confidence calculation based on consistency
	var scores []float64
	for _, assessment := range history.Assessments {
		scores = append(scores, assessment.OverallReliability)
	}

	mean := calculateMean(scores)
	stdDev := calculateStandardDeviation(scores, mean)

	if mean == 0 {
		return 0.3
	}

	cv := stdDev / mean
	confidence := math.Max(0.3, 1.0-cv)

	return confidence
}

func (dsra *DataSourceReliabilityAssessor) countDataPoints(sourceID string) int {
	dsra.mu.RLock()
	defer dsra.mu.RUnlock()

	count := 0

	if metrics, exists := dsra.reliabilityMetrics[sourceID]; exists {
		count += int(metrics.TotalRequests)
	}

	if history, exists := dsra.historicalPerformance[sourceID]; exists {
		count += len(history.Assessments)
		count += len(history.PerformanceMetrics)
	}

	return count
}

func (dsra *DataSourceReliabilityAssessor) cleanupIfNeeded() {
	if time.Since(dsra.lastCleanup) < dsra.config.CleanupInterval {
		return
	}

	dsra.lastCleanup = time.Now()
	cutoff := time.Now().Add(-dsra.config.HistoryRetentionPeriod)

	// Cleanup old performance metrics
	for sourceID, history := range dsra.historicalPerformance {
		var validMetrics []*PerformanceMetric
		for _, metric := range history.PerformanceMetrics {
			if metric.Timestamp.After(cutoff) {
				validMetrics = append(validMetrics, metric)
			}
		}
		history.PerformanceMetrics = validMetrics

		// Remove if no data left
		if len(history.Assessments) == 0 && len(history.PerformanceMetrics) == 0 {
			delete(dsra.historicalPerformance, sourceID)
		}
	}
}

// Utility functions

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}

func calculateStandardDeviation(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, value := range values {
		diff := value - mean
		sum += diff * diff
	}

	variance := sum / float64(len(values))
	return math.Sqrt(variance)
}

func getDefaultDataSourceReliabilityConfig() *DataSourceReliabilityConfig {
	return &DataSourceReliabilityConfig{
		// Assessment settings
		EnableHistoricalAnalysis:  true,
		EnablePerformanceTracking: true,
		EnablePredictiveScoring:   true,
		EnableAlerting:            true,

		// Thresholds
		LowReliabilityThreshold:      0.7,
		CriticalReliabilityThreshold: 0.5,
		PerformanceThreshold:         2 * time.Second,
		UptimeThreshold:              95.0,

		// Scoring weights
		HistoricalWeight:  0.2,
		PerformanceWeight: 0.25,
		AccuracyWeight:    0.15,
		ConsistencyWeight: 0.15,
		UptimeWeight:      0.15,
		DataQualityWeight: 0.1,

		// History settings
		MaxHistorySize:         1000,
		HistoryRetentionPeriod: 30 * 24 * time.Hour, // 30 days
		CleanupInterval:        1 * time.Hour,       // 1 hour

		// Alert settings
		AlertCooldownPeriod: 1 * time.Hour, // 1 hour
		MaxAlertsPerSource:  10,
	}
}
