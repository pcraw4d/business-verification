package feedback

import (
	"context"
	"fmt"
	"math"
	"time"

	"go.uber.org/zap"
)

// SecurityFeedbackAnalysisEngine provides advanced analysis capabilities for security feedback
type SecurityFeedbackAnalysisEngine struct {
	config               *SecurityAnalysisConfig
	logger               *zap.Logger
	patternDetector      *SecurityPatternDetector
	trendAnalyzer        *SecurityTrendAnalyzer
	anomalyDetector      *SecurityAnomalyDetector
	recommendationEngine *SecurityRecommendationEngine
	performanceAnalyzer  *SecurityPerformanceAnalyzer
}

// SecurityAnalysisConfig contains configuration for the security analysis engine
type SecurityAnalysisConfig struct {
	// Analysis settings
	MinDataPoints             int           `json:"min_data_points"`             // 10
	AnalysisWindow            time.Duration `json:"analysis_window"`             // 7 days
	TrendDetectionThreshold   float64       `json:"trend_detection_threshold"`   // 0.1
	AnomalyDetectionThreshold float64       `json:"anomaly_detection_threshold"` // 0.05

	// Pattern detection settings
	PatternMinOccurrences int     `json:"pattern_min_occurrences"` // 3
	PatternMinConfidence  float64 `json:"pattern_min_confidence"`  // 0.7

	// Performance analysis settings
	PerformanceThreshold  float64       `json:"performance_threshold"`   // 0.8
	ResponseTimeThreshold time.Duration `json:"response_time_threshold"` // 500ms

	// Recommendation settings
	RecommendationMinScore float64 `json:"recommendation_min_score"` // 0.6
	MaxRecommendations     int     `json:"max_recommendations"`      // 10
}

// SecurityPattern represents a detected security pattern
type SecurityPattern struct {
	PatternID          string                 `json:"pattern_id"`
	PatternType        string                 `json:"pattern_type"`
	Description        string                 `json:"description"`
	Confidence         float64                `json:"confidence"`
	Occurrences        int                    `json:"occurrences"`
	FirstSeen          time.Time              `json:"first_seen"`
	LastSeen           time.Time              `json:"last_seen"`
	AffectedComponents []string               `json:"affected_components"`
	Severity           string                 `json:"severity"`
	PatternData        map[string]interface{} `json:"pattern_data"`
	Recommendations    []string               `json:"recommendations"`
}

// SecurityTrend represents a detected security trend
type SecurityTrend struct {
	TrendID         string           `json:"trend_id"`
	TrendType       string           `json:"trend_type"`
	Direction       string           `json:"direction"` // "increasing", "decreasing", "stable"
	Magnitude       float64          `json:"magnitude"`
	Confidence      float64          `json:"confidence"`
	StartTime       time.Time        `json:"start_time"`
	EndTime         time.Time        `json:"end_time"`
	DataPoints      []TrendDataPoint `json:"data_points"`
	Impact          string           `json:"impact"`
	Recommendations []string         `json:"recommendations"`
}

// TrendDataPoint represents a single data point in a trend
type TrendDataPoint struct {
	Timestamp time.Time              `json:"timestamp"`
	Value     float64                `json:"value"`
	Context   map[string]interface{} `json:"context"`
}

// SecurityAnomaly represents a detected security anomaly
type SecurityAnomaly struct {
	AnomalyID       string                 `json:"anomaly_id"`
	AnomalyType     string                 `json:"anomaly_type"`
	Description     string                 `json:"description"`
	Severity        string                 `json:"severity"`
	Confidence      float64                `json:"confidence"`
	DetectionTime   time.Time              `json:"detection_time"`
	AffectedData    []string               `json:"affected_data"`
	AnomalyScore    float64                `json:"anomaly_score"`
	Context         map[string]interface{} `json:"context"`
	Recommendations []string               `json:"recommendations"`
}

// SecurityPerformanceMetrics represents performance metrics for security operations
type SecurityPerformanceMetrics struct {
	OperationType             string             `json:"operation_type"`
	AverageResponseTime       time.Duration      `json:"average_response_time"`
	P95ResponseTime           time.Duration      `json:"p95_response_time"`
	P99ResponseTime           time.Duration      `json:"p99_response_time"`
	Throughput                float64            `json:"throughput"` // operations per second
	ErrorRate                 float64            `json:"error_rate"`
	SuccessRate               float64            `json:"success_rate"`
	ResourceUtilization       map[string]float64 `json:"resource_utilization"`
	Bottlenecks               []string           `json:"bottlenecks"`
	OptimizationOpportunities []string           `json:"optimization_opportunities"`
}

// ComprehensiveSecurityAnalysis represents a comprehensive security analysis result
type ComprehensiveSecurityAnalysis struct {
	AnalysisID       string        `json:"analysis_id"`
	AnalysisTime     time.Time     `json:"analysis_time"`
	AnalysisDuration time.Duration `json:"analysis_duration"`

	// Overall metrics
	OverallSecurityScore    float64 `json:"overall_security_score"`
	OverallPerformanceScore float64 `json:"overall_performance_score"`
	OverallTrendScore       float64 `json:"overall_trend_score"`

	// Detected patterns
	SecurityPatterns []*SecurityPattern `json:"security_patterns"`
	PatternCount     int                `json:"pattern_count"`
	CriticalPatterns int                `json:"critical_patterns"`

	// Detected trends
	SecurityTrends []*SecurityTrend `json:"security_trends"`
	TrendCount     int              `json:"trend_count"`
	NegativeTrends int              `json:"negative_trends"`

	// Detected anomalies
	SecurityAnomalies []*SecurityAnomaly `json:"security_anomalies"`
	AnomalyCount      int                `json:"anomaly_count"`
	CriticalAnomalies int                `json:"critical_anomalies"`

	// Performance analysis
	PerformanceMetrics map[string]*SecurityPerformanceMetrics `json:"performance_metrics"`
	PerformanceIssues  []string                               `json:"performance_issues"`

	// Recommendations
	Recommendations             []*SecurityRecommendation `json:"recommendations"`
	RecommendationCount         int                       `json:"recommendation_count"`
	HighPriorityRecommendations int                       `json:"high_priority_recommendations"`

	// Risk assessment
	RiskLevel            string   `json:"risk_level"`
	RiskFactors          []string `json:"risk_factors"`
	MitigationStrategies []string `json:"mitigation_strategies"`

	// Data quality
	DataQualityScore float64 `json:"data_quality_score"`
	DataCompleteness float64 `json:"data_completeness"`
	DataFreshness    float64 `json:"data_freshness"`
}

// NewSecurityFeedbackAnalysisEngine creates a new security feedback analysis engine
func NewSecurityFeedbackAnalysisEngine(config *SecurityAnalysisConfig, logger *zap.Logger) *SecurityFeedbackAnalysisEngine {
	if config == nil {
		config = &SecurityAnalysisConfig{
			MinDataPoints:             10,
			AnalysisWindow:            7 * 24 * time.Hour,
			TrendDetectionThreshold:   0.1,
			AnomalyDetectionThreshold: 0.05,
			PatternMinOccurrences:     3,
			PatternMinConfidence:      0.7,
			PerformanceThreshold:      0.8,
			ResponseTimeThreshold:     500 * time.Millisecond,
			RecommendationMinScore:    0.6,
			MaxRecommendations:        10,
		}
	}

	return &SecurityFeedbackAnalysisEngine{
		config:               config,
		logger:               logger,
		patternDetector:      NewSecurityPatternDetector(config, logger),
		trendAnalyzer:        NewSecurityTrendAnalyzer(config, logger),
		anomalyDetector:      NewSecurityAnomalyDetector(config, logger),
		recommendationEngine: NewSecurityRecommendationEngine(config, logger),
		performanceAnalyzer:  NewSecurityPerformanceAnalyzer(config, logger),
	}
}

// AnalyzeSecurityFeedback performs comprehensive security feedback analysis
func (sae *SecurityFeedbackAnalysisEngine) AnalyzeSecurityFeedback(ctx context.Context, feedback []*UserFeedback) (*ComprehensiveSecurityAnalysis, error) {
	startTime := time.Now()

	sae.logger.Info("Starting comprehensive security feedback analysis",
		zap.Int("feedback_count", len(feedback)))

	analysis := &ComprehensiveSecurityAnalysis{
		AnalysisID:           fmt.Sprintf("analysis_%d", time.Now().Unix()),
		AnalysisTime:         time.Now(),
		SecurityPatterns:     make([]*SecurityPattern, 0),
		SecurityTrends:       make([]*SecurityTrend, 0),
		SecurityAnomalies:    make([]*SecurityAnomaly, 0),
		PerformanceMetrics:   make(map[string]*SecurityPerformanceMetrics),
		Recommendations:      make([]*SecurityRecommendation, 0),
		RiskFactors:          make([]string, 0),
		MitigationStrategies: make([]string, 0),
	}

	// Validate input data
	if err := sae.validateInputData(feedback); err != nil {
		return analysis, fmt.Errorf("input data validation failed: %w", err)
	}

	// Detect security patterns
	patterns, err := sae.patternDetector.DetectPatterns(ctx, feedback)
	if err != nil {
		sae.logger.Error("Failed to detect security patterns", zap.Error(err))
	} else {
		analysis.SecurityPatterns = patterns
		analysis.PatternCount = len(patterns)
		analysis.CriticalPatterns = sae.countCriticalPatterns(patterns)
	}

	// Analyze security trends
	trends, err := sae.trendAnalyzer.AnalyzeTrends(ctx, feedback)
	if err != nil {
		sae.logger.Error("Failed to analyze security trends", zap.Error(err))
	} else {
		analysis.SecurityTrends = trends
		analysis.TrendCount = len(trends)
		analysis.NegativeTrends = sae.countNegativeTrends(trends)
	}

	// Detect security anomalies
	anomalies, err := sae.anomalyDetector.DetectAnomalies(ctx, feedback)
	if err != nil {
		sae.logger.Error("Failed to detect security anomalies", zap.Error(err))
	} else {
		analysis.SecurityAnomalies = anomalies
		analysis.AnomalyCount = len(anomalies)
		analysis.CriticalAnomalies = sae.countCriticalAnomalies(anomalies)
	}

	// Analyze performance metrics
	performanceMetrics, err := sae.performanceAnalyzer.AnalyzePerformance(ctx, feedback)
	if err != nil {
		sae.logger.Error("Failed to analyze performance metrics", zap.Error(err))
	} else {
		analysis.PerformanceMetrics = performanceMetrics
		analysis.PerformanceIssues = sae.identifyPerformanceIssues(performanceMetrics)
	}

	// Generate recommendations
	recommendations, err := sae.recommendationEngine.GenerateRecommendations(ctx, analysis)
	if err != nil {
		sae.logger.Error("Failed to generate recommendations", zap.Error(err))
	} else {
		analysis.Recommendations = recommendations
		analysis.RecommendationCount = len(recommendations)
		analysis.HighPriorityRecommendations = sae.countHighPriorityRecommendations(recommendations)
	}

	// Calculate overall scores
	analysis.OverallSecurityScore = sae.calculateOverallSecurityScore(analysis)
	analysis.OverallPerformanceScore = sae.calculateOverallPerformanceScore(analysis)
	analysis.OverallTrendScore = sae.calculateOverallTrendScore(analysis)

	// Assess risk level
	analysis.RiskLevel = sae.assessRiskLevel(analysis)
	analysis.RiskFactors = sae.identifyRiskFactors(analysis)
	analysis.MitigationStrategies = sae.generateMitigationStrategies(analysis)

	// Calculate data quality metrics
	analysis.DataQualityScore = sae.calculateDataQualityScore(feedback)
	analysis.DataCompleteness = sae.calculateDataCompleteness(feedback)
	analysis.DataFreshness = sae.calculateDataFreshness(feedback)

	analysis.AnalysisDuration = time.Since(startTime)

	sae.logger.Info("Comprehensive security feedback analysis completed",
		zap.String("analysis_id", analysis.AnalysisID),
		zap.Float64("overall_security_score", analysis.OverallSecurityScore),
		zap.Float64("overall_performance_score", analysis.OverallPerformanceScore),
		zap.String("risk_level", analysis.RiskLevel),
		zap.Int("patterns_detected", analysis.PatternCount),
		zap.Int("trends_detected", analysis.TrendCount),
		zap.Int("anomalies_detected", analysis.AnomalyCount),
		zap.Int("recommendations_generated", analysis.RecommendationCount),
		zap.Duration("analysis_duration", analysis.AnalysisDuration))

	return analysis, nil
}

// validateInputData validates the input feedback data
func (sae *SecurityFeedbackAnalysisEngine) validateInputData(feedback []*UserFeedback) error {
	if len(feedback) < sae.config.MinDataPoints {
		return fmt.Errorf("insufficient data points: %d < %d", len(feedback), sae.config.MinDataPoints)
	}

	// Check data freshness
	cutoffTime := time.Now().Add(-sae.config.AnalysisWindow)
	validDataCount := 0
	for _, fb := range feedback {
		if fb.CreatedAt.After(cutoffTime) {
			validDataCount++
		}
	}

	if validDataCount < sae.config.MinDataPoints {
		return fmt.Errorf("insufficient recent data points: %d < %d", validDataCount, sae.config.MinDataPoints)
	}

	return nil
}

// countCriticalPatterns counts the number of critical security patterns
func (sae *SecurityFeedbackAnalysisEngine) countCriticalPatterns(patterns []*SecurityPattern) int {
	count := 0
	for _, pattern := range patterns {
		if pattern.Severity == "critical" || pattern.Severity == "high" {
			count++
		}
	}
	return count
}

// countNegativeTrends counts the number of negative security trends
func (sae *SecurityFeedbackAnalysisEngine) countNegativeTrends(trends []*SecurityTrend) int {
	count := 0
	for _, trend := range trends {
		if trend.Direction == "increasing" && trend.Impact == "negative" {
			count++
		}
	}
	return count
}

// countCriticalAnomalies counts the number of critical security anomalies
func (sae *SecurityFeedbackAnalysisEngine) countCriticalAnomalies(anomalies []*SecurityAnomaly) int {
	count := 0
	for _, anomaly := range anomalies {
		if anomaly.Severity == "critical" || anomaly.Severity == "high" {
			count++
		}
	}
	return count
}

// identifyPerformanceIssues identifies performance issues from metrics
func (sae *SecurityFeedbackAnalysisEngine) identifyPerformanceIssues(metrics map[string]*SecurityPerformanceMetrics) []string {
	var issues []string

	for operationType, metric := range metrics {
		if metric.AverageResponseTime > sae.config.ResponseTimeThreshold {
			issues = append(issues, fmt.Sprintf("%s: Average response time %.2fms exceeds threshold %.2fms",
				operationType, float64(metric.AverageResponseTime.Nanoseconds())/1e6,
				float64(sae.config.ResponseTimeThreshold.Nanoseconds())/1e6))
		}

		if metric.ErrorRate > 0.05 { // 5% error rate threshold
			issues = append(issues, fmt.Sprintf("%s: Error rate %.2f%% exceeds threshold 5%%",
				operationType, metric.ErrorRate*100))
		}

		if metric.SuccessRate < sae.config.PerformanceThreshold {
			issues = append(issues, fmt.Sprintf("%s: Success rate %.2f%% below threshold %.2f%%",
				operationType, metric.SuccessRate*100, sae.config.PerformanceThreshold*100))
		}
	}

	return issues
}

// countHighPriorityRecommendations counts high-priority recommendations
func (sae *SecurityFeedbackAnalysisEngine) countHighPriorityRecommendations(recommendations []*SecurityRecommendation) int {
	count := 0
	for _, rec := range recommendations {
		if rec.Priority == "high" || rec.Priority == "critical" {
			count++
		}
	}
	return count
}

// calculateOverallSecurityScore calculates the overall security score
func (sae *SecurityFeedbackAnalysisEngine) calculateOverallSecurityScore(analysis *ComprehensiveSecurityAnalysis) float64 {
	baseScore := 1.0

	// Penalize for critical patterns
	patternPenalty := float64(analysis.CriticalPatterns) * 0.1

	// Penalize for negative trends
	trendPenalty := float64(analysis.NegativeTrends) * 0.05

	// Penalize for critical anomalies
	anomalyPenalty := float64(analysis.CriticalAnomalies) * 0.15

	// Penalize for performance issues
	performancePenalty := float64(len(analysis.PerformanceIssues)) * 0.02

	// Calculate final score
	finalScore := baseScore - patternPenalty - trendPenalty - anomalyPenalty - performancePenalty

	// Ensure score is between 0 and 1
	return math.Max(0.0, math.Min(1.0, finalScore))
}

// calculateOverallPerformanceScore calculates the overall performance score
func (sae *SecurityFeedbackAnalysisEngine) calculateOverallPerformanceScore(analysis *ComprehensiveSecurityAnalysis) float64 {
	if len(analysis.PerformanceMetrics) == 0 {
		return 1.0
	}

	var totalScore float64
	var count int

	for _, metric := range analysis.PerformanceMetrics {
		// Calculate performance score based on response time, error rate, and success rate
		responseTimeScore := math.Max(0.0, 1.0-float64(metric.AverageResponseTime)/float64(sae.config.ResponseTimeThreshold))
		errorRateScore := math.Max(0.0, 1.0-metric.ErrorRate*10) // Scale error rate
		successRateScore := metric.SuccessRate

		// Weighted average
		operationScore := responseTimeScore*0.3 + errorRateScore*0.3 + successRateScore*0.4
		totalScore += operationScore
		count++
	}

	if count == 0 {
		return 1.0
	}

	return totalScore / float64(count)
}

// calculateOverallTrendScore calculates the overall trend score
func (sae *SecurityFeedbackAnalysisEngine) calculateOverallTrendScore(analysis *ComprehensiveSecurityAnalysis) float64 {
	if len(analysis.SecurityTrends) == 0 {
		return 1.0
	}

	var totalScore float64
	var count int

	for _, trend := range analysis.SecurityTrends {
		// Positive trends increase score, negative trends decrease score
		if trend.Direction == "decreasing" && trend.Impact == "positive" {
			totalScore += 1.0
		} else if trend.Direction == "increasing" && trend.Impact == "negative" {
			totalScore += 0.0
		} else {
			totalScore += 0.5 // Neutral trends
		}
		count++
	}

	if count == 0 {
		return 1.0
	}

	return totalScore / float64(count)
}

// assessRiskLevel assesses the overall risk level
func (sae *SecurityFeedbackAnalysisEngine) assessRiskLevel(analysis *ComprehensiveSecurityAnalysis) string {
	riskScore := 0.0

	// Add risk factors
	riskScore += float64(analysis.CriticalPatterns) * 0.3
	riskScore += float64(analysis.NegativeTrends) * 0.2
	riskScore += float64(analysis.CriticalAnomalies) * 0.4
	riskScore += float64(len(analysis.PerformanceIssues)) * 0.1

	// Determine risk level
	if riskScore >= 2.0 {
		return "critical"
	} else if riskScore >= 1.0 {
		return "high"
	} else if riskScore >= 0.5 {
		return "medium"
	} else {
		return "low"
	}
}

// identifyRiskFactors identifies specific risk factors
func (sae *SecurityFeedbackAnalysisEngine) identifyRiskFactors(analysis *ComprehensiveSecurityAnalysis) []string {
	var riskFactors []string

	if analysis.CriticalPatterns > 0 {
		riskFactors = append(riskFactors, fmt.Sprintf("%d critical security patterns detected", analysis.CriticalPatterns))
	}

	if analysis.NegativeTrends > 0 {
		riskFactors = append(riskFactors, fmt.Sprintf("%d negative security trends identified", analysis.NegativeTrends))
	}

	if analysis.CriticalAnomalies > 0 {
		riskFactors = append(riskFactors, fmt.Sprintf("%d critical security anomalies found", analysis.CriticalAnomalies))
	}

	if len(analysis.PerformanceIssues) > 0 {
		riskFactors = append(riskFactors, fmt.Sprintf("%d performance issues identified", len(analysis.PerformanceIssues)))
	}

	if analysis.OverallSecurityScore < 0.7 {
		riskFactors = append(riskFactors, "Overall security score below acceptable threshold")
	}

	if analysis.OverallPerformanceScore < 0.8 {
		riskFactors = append(riskFactors, "Overall performance score below acceptable threshold")
	}

	return riskFactors
}

// generateMitigationStrategies generates mitigation strategies based on analysis
func (sae *SecurityFeedbackAnalysisEngine) generateMitigationStrategies(analysis *ComprehensiveSecurityAnalysis) []string {
	var strategies []string

	// Pattern-based mitigation strategies
	for _, pattern := range analysis.SecurityPatterns {
		if pattern.Severity == "critical" || pattern.Severity == "high" {
			strategies = append(strategies, fmt.Sprintf("Address %s pattern: %s", pattern.PatternType, pattern.Description))
		}
	}

	// Trend-based mitigation strategies
	for _, trend := range analysis.SecurityTrends {
		if trend.Direction == "increasing" && trend.Impact == "negative" {
			strategies = append(strategies, fmt.Sprintf("Mitigate %s trend", trend.TrendType))
		}
	}

	// Anomaly-based mitigation strategies
	for _, anomaly := range analysis.SecurityAnomalies {
		if anomaly.Severity == "critical" || anomaly.Severity == "high" {
			strategies = append(strategies, fmt.Sprintf("Investigate %s anomaly: %s", anomaly.AnomalyType, anomaly.Description))
		}
	}

	// Performance-based mitigation strategies
	for _, issue := range analysis.PerformanceIssues {
		strategies = append(strategies, fmt.Sprintf("Optimize performance: %s", issue))
	}

	// General mitigation strategies
	if analysis.OverallSecurityScore < 0.7 {
		strategies = append(strategies, "Implement comprehensive security review and enhancement")
	}

	if analysis.OverallPerformanceScore < 0.8 {
		strategies = append(strategies, "Conduct performance optimization review")
	}

	return strategies
}

// calculateDataQualityScore calculates the quality score of the input data
func (sae *SecurityFeedbackAnalysisEngine) calculateDataQualityScore(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 0.0
	}

	var qualityScore float64
	var count int

	for _, fb := range feedback {
		score := 1.0

		// Check completeness
		if fb.BusinessName == "" {
			score -= 0.1
		}
		if fb.FeedbackText == "" {
			score -= 0.1
		}
		if fb.FeedbackType == "" {
			score -= 0.2
		}

		// Check validity
		if fb.ConfidenceScore < 0.0 || fb.ConfidenceScore > 1.0 {
			score -= 0.2
		}

		// Check timestamp validity
		if fb.CreatedAt.IsZero() || fb.CreatedAt.After(time.Now()) {
			score -= 0.1
		}

		qualityScore += math.Max(0.0, score)
		count++
	}

	if count == 0 {
		return 0.0
	}

	return qualityScore / float64(count)
}

// calculateDataCompleteness calculates the completeness of the input data
func (sae *SecurityFeedbackAnalysisEngine) calculateDataCompleteness(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 0.0
	}

	requiredFields := []string{"business_name", "feedback_text", "feedback_type", "confidence_score"}
	var completenessScore float64
	var count int

	for _, fb := range feedback {
		fieldCount := 0
		if fb.BusinessName != "" {
			fieldCount++
		}
		if fb.FeedbackText != "" {
			fieldCount++
		}
		if fb.FeedbackType != "" {
			fieldCount++
		}
		if fb.ConfidenceScore >= 0.0 {
			fieldCount++
		}

		completenessScore += float64(fieldCount) / float64(len(requiredFields))
		count++
	}

	if count == 0 {
		return 0.0
	}

	return completenessScore / float64(count)
}

// calculateDataFreshness calculates the freshness of the input data
func (sae *SecurityFeedbackAnalysisEngine) calculateDataFreshness(feedback []*UserFeedback) float64 {
	if len(feedback) == 0 {
		return 0.0
	}

	now := time.Now()
	var freshnessScore float64
	var count int

	for _, fb := range feedback {
		age := now.Sub(fb.CreatedAt)
		// Score decreases with age, with data older than analysis window getting 0
		if age <= sae.config.AnalysisWindow {
			score := 1.0 - (age.Hours() / sae.config.AnalysisWindow.Hours())
			freshnessScore += math.Max(0.0, score)
		}
		count++
	}

	if count == 0 {
		return 0.0
	}

	return freshnessScore / float64(count)
}

// GetAnalysisConfig returns the current analysis configuration
func (sae *SecurityFeedbackAnalysisEngine) GetAnalysisConfig() *SecurityAnalysisConfig {
	return sae.config
}

// UpdateAnalysisConfig updates the analysis configuration
func (sae *SecurityFeedbackAnalysisEngine) UpdateAnalysisConfig(config *SecurityAnalysisConfig) {
	sae.config = config
}
