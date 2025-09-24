package feedback

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MLAwareAnalyzer provides ML-aware feedback analysis capabilities
type MLAwareAnalyzer struct {
	config                    *MLAnalysisConfig
	logger                    *zap.Logger
	patternAnalyzer           *FeedbackPatternAnalyzer
	misclassificationDetector *MisclassificationDetector
	recommendationEngine      *MLRecommendationEngine
	weightOptimizer           *EnsembleWeightOptimizer
	securityAnalyzer          *SecurityFeedbackAnalyzer
	mu                        sync.RWMutex
	analysisCache             map[string]*MLAnalysisResult
	lastAnalysisTime          time.Time
}

// MLAnalysisConfig holds configuration for ML-aware analysis
type MLAnalysisConfig struct {
	// Analysis settings
	AnalysisWindowSize        time.Duration `json:"analysis_window_size"`        // 24 hours
	MinFeedbackThreshold      int           `json:"min_feedback_threshold"`      // 10 feedback items
	ConfidenceThreshold       float64       `json:"confidence_threshold"`        // 0.7
	PatternDetectionThreshold float64       `json:"pattern_detection_threshold"` // 0.8

	// ML-specific settings
	MLModelDriftThreshold         float64 `json:"ml_model_drift_threshold"`        // 0.15
	EnsembleDisagreementThreshold float64 `json:"ensemble_disagreement_threshold"` // 0.3
	WeightAdjustmentStep          float64 `json:"weight_adjustment_step"`          // 0.05
	MaxWeightChange               float64 `json:"max_weight_change"`               // 0.2

	// Security settings
	SecurityViolationThreshold float64 `json:"security_violation_threshold"` // 0.1
	TrustedSourceThreshold     float64 `json:"trusted_source_threshold"`     // 0.95

	// Performance settings
	EnableRealTimeAnalysis bool          `json:"enable_real_time_analysis"`
	CacheAnalysisResults   bool          `json:"cache_analysis_results"`
	CacheTTL               time.Duration `json:"cache_ttl"`               // 1 hour
	MaxConcurrentAnalysis  int           `json:"max_concurrent_analysis"` // 5
}

// MLAnalysisResult represents the result of ML-aware feedback analysis
type MLAnalysisResult struct {
	AnalysisID     string        `json:"analysis_id"`
	Timestamp      time.Time     `json:"timestamp"`
	AnalysisWindow time.Duration `json:"analysis_window"`

	// Pattern analysis results
	FeedbackPatterns          []*FeedbackPattern          `json:"feedback_patterns"`
	MethodPerformancePatterns []*MethodPerformancePattern `json:"method_performance_patterns"`
	TemporalPatterns          []*TemporalPattern          `json:"temporal_patterns"`

	// Misclassification analysis
	Misclassifications     []*ModelMisclassification `json:"misclassifications"`
	RootCauseAnalysis      []*RootCauseAnalysis      `json:"root_cause_analysis"`
	ConfidenceCorrelations []*ConfidenceCorrelation  `json:"confidence_correlations"`

	// ML model analysis
	MLModelDrift           *MLModelDriftAnalysis    `json:"ml_model_drift"`
	EnsembleDisagreements  []*EnsembleDisagreement  `json:"ensemble_disagreements"`
	ModelPerformanceTrends []*ModelPerformanceTrend `json:"model_performance_trends"`

	// Recommendations
	MLModelRecommendations        []*MLModelRecommendation  `json:"ml_model_recommendations"`
	EnsembleWeightRecommendations []*WeightRecommendation   `json:"ensemble_weight_recommendations"`
	SecurityRecommendations       []*SecurityRecommendation `json:"security_recommendations"`

	// Security analysis
	SecurityAnalysis      *SecurityFeedbackAnalysis `json:"security_analysis"`
	TrustedSourceAnalysis *TrustedSourceAnalysis    `json:"trusted_source_analysis"`

	// Summary metrics
	OverallAccuracy       float64 `json:"overall_accuracy"`
	OverallConfidence     float64 `json:"overall_confidence"`
	ErrorRate             float64 `json:"error_rate"`
	SecurityViolationRate float64 `json:"security_violation_rate"`
	RecommendationCount   int     `json:"recommendation_count"`

	// Metadata
	ProcessingTimeMs int64                  `json:"processing_time_ms"`
	FeedbackCount    int                    `json:"feedback_count"`
	MethodCount      int                    `json:"method_count"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// FeedbackPattern represents a detected pattern in feedback data
type FeedbackPattern struct {
	PatternID          string                 `json:"pattern_id"`
	PatternType        string                 `json:"pattern_type"`
	PatternDescription string                 `json:"pattern_description"`
	Confidence         float64                `json:"confidence"`
	OccurrenceCount    int                    `json:"occurrence_count"`
	AffectedMethods    []ClassificationMethod `json:"affected_methods"`
	TimeWindow         time.Duration          `json:"time_window"`
	Severity           string                 `json:"severity"`
	Trend              string                 `json:"trend"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// MethodPerformancePattern represents performance patterns for specific methods
type MethodPerformancePattern struct {
	Method           ClassificationMethod `json:"method"`
	PerformanceTrend string               `json:"performance_trend"`
	AccuracyTrend    string               `json:"accuracy_trend"`
	ConfidenceTrend  string               `json:"confidence_trend"`
	ErrorPatterns    []string             `json:"error_patterns"`
	ImprovementAreas []string             `json:"improvement_areas"`
	Recommendations  []string             `json:"recommendations"`
}

// TemporalPattern represents temporal patterns in feedback
type TemporalPattern struct {
	PatternType string   `json:"pattern_type"`
	TimeWindow  string   `json:"time_window"`
	Frequency   int      `json:"frequency"`
	PeakTimes   []string `json:"peak_times"`
	Seasonality string   `json:"seasonality"`
	Trend       string   `json:"trend"`
	Confidence  float64  `json:"confidence"`
}

// ModelMisclassification represents misclassifications specific to ML models
type ModelMisclassification struct {
	ModelID               string   `json:"model_id"`
	ModelType             string   `json:"model_type"`
	MisclassificationType string   `json:"misclassification_type"`
	Frequency             int      `json:"frequency"`
	Confidence            float64  `json:"confidence"`
	AffectedIndustries    []string `json:"affected_industries"`
	CommonInputs          []string `json:"common_inputs"`
	RootCauses            []string `json:"root_causes"`
	Recommendations       []string `json:"recommendations"`
}

// RootCauseAnalysis represents analysis of root causes for issues
type RootCauseAnalysis struct {
	IssueID         string       `json:"issue_id"`
	IssueType       string       `json:"issue_type"`
	RootCauses      []*RootCause `json:"root_causes"`
	Confidence      float64      `json:"confidence"`
	Impact          string       `json:"impact"`
	Recommendations []string     `json:"recommendations"`
}

// RootCause represents a specific root cause
type RootCause struct {
	CauseID            string   `json:"cause_id"`
	Description        string   `json:"description"`
	Confidence         float64  `json:"confidence"`
	Evidence           []string `json:"evidence"`
	AffectedComponents []string `json:"affected_components"`
}

// ConfidenceCorrelation represents correlation between confidence and accuracy
type ConfidenceCorrelation struct {
	Method             ClassificationMethod `json:"method"`
	CorrelationCoeff   float64              `json:"correlation_coefficient"`
	ConfidenceAccuracy map[string]float64   `json:"confidence_accuracy_mapping"`
	CalibrationScore   float64              `json:"calibration_score"`
	Recommendations    []string             `json:"recommendations"`
}

// MLModelDriftAnalysis represents analysis of ML model drift
type MLModelDriftAnalysis struct {
	ModelID              string    `json:"model_id"`
	DriftDetected        bool      `json:"drift_detected"`
	DriftScore           float64   `json:"drift_score"`
	DriftType            string    `json:"drift_type"`
	AffectedFeatures     []string  `json:"affected_features"`
	Recommendations      []string  `json:"recommendations"`
	LastRetrainDate      time.Time `json:"last_retrain_date"`
	SuggestedRetrainDate time.Time `json:"suggested_retrain_date"`
}

// EnsembleDisagreement represents disagreements between ensemble methods
type EnsembleDisagreement struct {
	DisagreementID     string                 `json:"disagreement_id"`
	DisagreementType   string                 `json:"disagreement_type"`
	ConflictingMethods []ClassificationMethod `json:"conflicting_methods"`
	DisagreementScore  float64                `json:"disagreement_score"`
	Frequency          int                    `json:"frequency"`
	CommonInputs       []string               `json:"common_inputs"`
	ResolutionStrategy string                 `json:"resolution_strategy"`
}

// ModelPerformanceTrend represents performance trends for models
type ModelPerformanceTrend struct {
	ModelID        string                  `json:"model_id"`
	TrendType      string                  `json:"trend_type"`
	TrendDirection string                  `json:"trend_direction"`
	TrendStrength  float64                 `json:"trend_strength"`
	TimeWindow     time.Duration           `json:"time_window"`
	DataPoints     []*PerformanceDataPoint `json:"data_points"`
	Projection     *PerformanceProjection  `json:"projection"`
}

// PerformanceDataPoint represents a single performance data point
type PerformanceDataPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	Accuracy       float64   `json:"accuracy"`
	Confidence     float64   `json:"confidence"`
	ProcessingTime int64     `json:"processing_time_ms"`
	ErrorRate      float64   `json:"error_rate"`
}

// PerformanceProjection represents future performance projection
type PerformanceProjection struct {
	ProjectedAccuracy   float64   `json:"projected_accuracy"`
	ProjectedConfidence float64   `json:"projected_confidence"`
	ConfidenceInterval  []float64 `json:"confidence_interval"`
	ProjectionDate      time.Time `json:"projection_date"`
}

// MLModelRecommendation represents recommendations for ML model improvements
type MLModelRecommendation struct {
	RecommendationID    string   `json:"recommendation_id"`
	ModelID             string   `json:"model_id"`
	RecommendationType  string   `json:"recommendation_type"`
	Priority            string   `json:"priority"`
	Description         string   `json:"description"`
	ExpectedImprovement float64  `json:"expected_improvement"`
	ImplementationCost  string   `json:"implementation_cost"`
	Timeline            string   `json:"timeline"`
	Evidence            []string `json:"evidence"`
}

// WeightRecommendation represents recommendations for ensemble weight adjustments
type WeightRecommendation struct {
	RecommendationID   string               `json:"recommendation_id"`
	Method             ClassificationMethod `json:"method"`
	CurrentWeight      float64              `json:"current_weight"`
	RecommendedWeight  float64              `json:"recommended_weight"`
	WeightChange       float64              `json:"weight_change"`
	Reasoning          string               `json:"reasoning"`
	ExpectedImpact     float64              `json:"expected_impact"`
	Confidence         float64              `json:"confidence"`
	ImplementationDate time.Time            `json:"implementation_date"`
}

// SecurityRecommendation represents security-related recommendations
type SecurityRecommendation struct {
	RecommendationID    string   `json:"recommendation_id"`
	SecurityType        string   `json:"security_type"`
	Priority            string   `json:"priority"`
	Description         string   `json:"description"`
	AffectedComponents  []string `json:"affected_components"`
	ImplementationSteps []string `json:"implementation_steps"`
	ValidationCriteria  []string `json:"validation_criteria"`
}

// SecurityFeedbackAnalysis represents analysis of security-related feedback
type SecurityFeedbackAnalysis struct {
	SecurityViolations        []*SecurityViolation        `json:"security_violations"`
	TrustedSourceIssues       []*TrustedSourceIssue       `json:"trusted_source_issues"`
	WebsiteVerificationIssues []*WebsiteVerificationIssue `json:"website_verification_issues"`
	OverallSecurityScore      float64                     `json:"overall_security_score"`
	Recommendations           []*SecurityRecommendation   `json:"recommendations"`
}

// SecurityViolation represents a security violation detected in feedback
type SecurityViolation struct {
	ViolationID      string    `json:"violation_id"`
	ViolationType    string    `json:"violation_type"`
	Severity         string    `json:"severity"`
	Description      string    `json:"description"`
	AffectedData     []string  `json:"affected_data"`
	DetectionTime    time.Time `json:"detection_time"`
	ResolutionStatus string    `json:"resolution_status"`
}

// TrustedSourceIssue represents issues with trusted data sources
type TrustedSourceIssue struct {
	IssueID          string `json:"issue_id"`
	SourceType       string `json:"source_type"`
	IssueType        string `json:"issue_type"`
	Description      string `json:"description"`
	AffectedRequests int    `json:"affected_requests"`
	ResolutionStatus string `json:"resolution_status"`
}

// WebsiteVerificationIssue represents issues with website verification
type WebsiteVerificationIssue struct {
	IssueID          string   `json:"issue_id"`
	VerificationType string   `json:"verification_type"`
	IssueType        string   `json:"issue_type"`
	Description      string   `json:"description"`
	AffectedWebsites []string `json:"affected_websites"`
	ResolutionStatus string   `json:"resolution_status"`
}

// TrustedSourceAnalysis represents analysis of trusted data source feedback
type TrustedSourceAnalysis struct {
	SourceReliability map[string]float64 `json:"source_reliability"`
	SourceAccuracy    map[string]float64 `json:"source_accuracy"`
	SourcePerformance map[string]float64 `json:"source_performance"`
	Recommendations   []string           `json:"recommendations"`
	OverallTrustScore float64            `json:"overall_trust_score"`
}

// NewMLAwareAnalyzer creates a new ML-aware feedback analyzer
func NewMLAwareAnalyzer(config *MLAnalysisConfig, logger *zap.Logger) *MLAwareAnalyzer {
	if config == nil {
		config = &MLAnalysisConfig{
			AnalysisWindowSize:            24 * time.Hour,
			MinFeedbackThreshold:          10,
			ConfidenceThreshold:           0.7,
			PatternDetectionThreshold:     0.8,
			MLModelDriftThreshold:         0.15,
			EnsembleDisagreementThreshold: 0.3,
			WeightAdjustmentStep:          0.05,
			MaxWeightChange:               0.2,
			SecurityViolationThreshold:    0.1,
			TrustedSourceThreshold:        0.95,
			EnableRealTimeAnalysis:        true,
			CacheAnalysisResults:          true,
			CacheTTL:                      1 * time.Hour,
			MaxConcurrentAnalysis:         5,
		}
	}

	return &MLAwareAnalyzer{
		config:                    config,
		logger:                    logger,
		patternAnalyzer:           NewFeedbackPatternAnalyzer(config, logger),
		misclassificationDetector: NewMisclassificationDetector(config, logger),
		recommendationEngine:      NewMLRecommendationEngine(config, logger),
		weightOptimizer:           NewEnsembleWeightOptimizer(config, logger),
		securityAnalyzer:          NewSecurityFeedbackAnalyzer(config, logger),
		analysisCache:             make(map[string]*MLAnalysisResult),
		lastAnalysisTime:          time.Now(),
	}
}

// AnalyzeFeedbackPatterns analyzes feedback patterns across ensemble methods
func (analyzer *MLAwareAnalyzer) AnalyzeFeedbackPatterns(ctx context.Context, feedback []*UserFeedback) ([]*FeedbackPattern, error) {
	analyzer.logger.Info("analyzing feedback patterns across ensemble methods",
		zap.Int("feedback_count", len(feedback)))

	// Group feedback by method
	methodFeedback := make(map[ClassificationMethod][]*UserFeedback)
	for _, fb := range feedback {
		methodFeedback[fb.ClassificationMethod] = append(methodFeedback[fb.ClassificationMethod], fb)
	}

	var patterns []*FeedbackPattern

	// Analyze patterns for each method
	for method, methodFb := range methodFeedback {
		methodPatterns, err := analyzer.patternAnalyzer.AnalyzeMethodPatterns(ctx, method, methodFb)
		if err != nil {
			analyzer.logger.Warn("failed to analyze patterns for method",
				zap.String("method", string(method)),
				zap.Error(err))
			continue
		}
		patterns = append(patterns, methodPatterns...)
	}

	// Analyze cross-method patterns
	crossMethodPatterns, err := analyzer.patternAnalyzer.AnalyzeCrossMethodPatterns(ctx, methodFeedback)
	if err != nil {
		analyzer.logger.Warn("failed to analyze cross-method patterns", zap.Error(err))
	} else {
		patterns = append(patterns, crossMethodPatterns...)
	}

	// Sort patterns by confidence and occurrence
	sort.Slice(patterns, func(i, j int) bool {
		if patterns[i].Confidence == patterns[j].Confidence {
			return patterns[i].OccurrenceCount > patterns[j].OccurrenceCount
		}
		return patterns[i].Confidence > patterns[j].Confidence
	})

	analyzer.logger.Info("feedback pattern analysis completed",
		zap.Int("total_patterns", len(patterns)),
		zap.Int("method_count", len(methodFeedback)))

	return patterns, nil
}

// IdentifyModelSpecificMisclassifications identifies misclassifications specific to ML models
func (analyzer *MLAwareAnalyzer) IdentifyModelSpecificMisclassifications(ctx context.Context, feedback []*UserFeedback) ([]*ModelMisclassification, error) {
	analyzer.logger.Info("identifying model-specific misclassifications",
		zap.Int("feedback_count", len(feedback)))

	// Filter feedback for ML-related issues
	mlFeedback := make([]*UserFeedback, 0)
	for _, fb := range feedback {
		if fb.ClassificationMethod == MethodML ||
			strings.Contains(strings.ToLower(fb.FeedbackText), "ml") ||
			strings.Contains(strings.ToLower(fb.FeedbackText), "model") {
			mlFeedback = append(mlFeedback, fb)
		}
	}

	if len(mlFeedback) < analyzer.config.MinFeedbackThreshold {
		analyzer.logger.Info("insufficient ML feedback for misclassification analysis",
			zap.Int("ml_feedback_count", len(mlFeedback)),
			zap.Int("min_threshold", analyzer.config.MinFeedbackThreshold))
		return []*ModelMisclassification{}, nil
	}

	misclassifications, err := analyzer.misclassificationDetector.DetectModelMisclassifications(ctx, mlFeedback)
	if err != nil {
		return nil, fmt.Errorf("failed to detect model misclassifications: %w", err)
	}

	analyzer.logger.Info("model-specific misclassification analysis completed",
		zap.Int("misclassification_count", len(misclassifications)))

	return misclassifications, nil
}

// GenerateMLModelRecommendations generates recommendations for ML model improvements
func (analyzer *MLAwareAnalyzer) GenerateMLModelRecommendations(ctx context.Context, analysis *MLAnalysisResult) ([]*MLModelRecommendation, error) {
	analyzer.logger.Info("generating ML model improvement recommendations",
		zap.String("analysis_id", analysis.AnalysisID))

	recommendations, err := analyzer.recommendationEngine.GenerateModelRecommendations(ctx, analysis)
	if err != nil {
		return nil, fmt.Errorf("failed to generate model recommendations: %w", err)
	}

	// Sort recommendations by priority and expected improvement
	sort.Slice(recommendations, func(i, j int) bool {
		priorityOrder := map[string]int{"high": 3, "medium": 2, "low": 1}
		if priorityOrder[recommendations[i].Priority] == priorityOrder[recommendations[j].Priority] {
			return recommendations[i].ExpectedImprovement > recommendations[j].ExpectedImprovement
		}
		return priorityOrder[recommendations[i].Priority] > priorityOrder[recommendations[j].Priority]
	})

	analyzer.logger.Info("ML model recommendations generated",
		zap.Int("recommendation_count", len(recommendations)))

	return recommendations, nil
}

// AnalyzeEnsembleWeightOptimization analyzes opportunities for ensemble weight optimization
func (analyzer *MLAwareAnalyzer) AnalyzeEnsembleWeightOptimization(ctx context.Context, feedback []*UserFeedback) ([]*WeightRecommendation, error) {
	analyzer.logger.Info("analyzing ensemble weight optimization opportunities",
		zap.Int("feedback_count", len(feedback)))

	recommendations, err := analyzer.weightOptimizer.AnalyzeWeightOptimization(ctx, feedback)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze weight optimization: %w", err)
	}

	// Filter recommendations based on confidence and impact
	filteredRecommendations := make([]*WeightRecommendation, 0)
	for _, rec := range recommendations {
		if rec.Confidence >= analyzer.config.ConfidenceThreshold &&
			rec.ExpectedImpact >= analyzer.config.WeightAdjustmentStep {
			filteredRecommendations = append(filteredRecommendations, rec)
		}
	}

	analyzer.logger.Info("ensemble weight optimization analysis completed",
		zap.Int("total_recommendations", len(recommendations)),
		zap.Int("filtered_recommendations", len(filteredRecommendations)))

	return filteredRecommendations, nil
}

// AnalyzeTrustedDataSourceFeedback analyzes feedback related to trusted data sources
func (analyzer *MLAwareAnalyzer) AnalyzeTrustedDataSourceFeedback(ctx context.Context, feedback []*UserFeedback) (*TrustedSourceAnalysis, error) {
	analyzer.logger.Info("analyzing trusted data source feedback",
		zap.Int("feedback_count", len(feedback)))

	analysis, err := analyzer.securityAnalyzer.AnalyzeTrustedSourceFeedback(ctx, feedback)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze trusted source feedback: %w", err)
	}

	analyzer.logger.Info("trusted data source feedback analysis completed",
		zap.Float64("overall_trust_score", analysis.OverallTrustScore))

	return analysis, nil
}

// PerformComprehensiveAnalysis performs comprehensive ML-aware feedback analysis
func (analyzer *MLAwareAnalyzer) PerformComprehensiveAnalysis(ctx context.Context, feedback []*UserFeedback) (*MLAnalysisResult, error) {
	startTime := time.Now()
	analysisID := generateAnalysisID()

	analyzer.logger.Info("performing comprehensive ML-aware feedback analysis",
		zap.String("analysis_id", analysisID),
		zap.Int("feedback_count", len(feedback)))

	// Check cache first
	if analyzer.config.CacheAnalysisResults {
		if cached, exists := analyzer.analysisCache[analysisID]; exists {
			if time.Since(cached.Timestamp) < analyzer.config.CacheTTL {
				analyzer.logger.Info("returning cached analysis result",
					zap.String("analysis_id", analysisID))
				return cached, nil
			}
		}
	}

	result := &MLAnalysisResult{
		AnalysisID:     analysisID,
		Timestamp:      time.Now(),
		AnalysisWindow: analyzer.config.AnalysisWindowSize,
		Metadata:       make(map[string]interface{}),
	}

	// Perform all analysis components in parallel
	var wg sync.WaitGroup
	errChan := make(chan error, 6)

	// 1. Analyze feedback patterns
	wg.Add(1)
	go func() {
		defer wg.Done()
		patterns, err := analyzer.AnalyzeFeedbackPatterns(ctx, feedback)
		if err != nil {
			errChan <- fmt.Errorf("pattern analysis failed: %w", err)
			return
		}
		result.FeedbackPatterns = patterns
	}()

	// 2. Identify model-specific misclassifications
	wg.Add(1)
	go func() {
		defer wg.Done()
		misclassifications, err := analyzer.IdentifyModelSpecificMisclassifications(ctx, feedback)
		if err != nil {
			errChan <- fmt.Errorf("misclassification analysis failed: %w", err)
			return
		}
		result.Misclassifications = misclassifications
	}()

	// 3. Analyze ensemble weight optimization
	wg.Add(1)
	go func() {
		defer wg.Done()
		weightRecs, err := analyzer.AnalyzeEnsembleWeightOptimization(ctx, feedback)
		if err != nil {
			errChan <- fmt.Errorf("weight optimization analysis failed: %w", err)
			return
		}
		result.EnsembleWeightRecommendations = weightRecs
	}()

	// 4. Analyze trusted data source feedback
	wg.Add(1)
	go func() {
		defer wg.Done()
		trustAnalysis, err := analyzer.AnalyzeTrustedDataSourceFeedback(ctx, feedback)
		if err != nil {
			errChan <- fmt.Errorf("trusted source analysis failed: %w", err)
			return
		}
		result.TrustedSourceAnalysis = trustAnalysis
	}()

	// 5. Perform security analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		securityAnalysis, err := analyzer.securityAnalyzer.AnalyzeSecurityFeedback(ctx, feedback)
		if err != nil {
			errChan <- fmt.Errorf("security analysis failed: %w", err)
			return
		}
		result.SecurityAnalysis = securityAnalysis
	}()

	// 6. Generate ML model recommendations
	wg.Add(1)
	go func() {
		defer wg.Done()
		// We need to wait for other analyses to complete first
		// This will be handled after the wait group
	}()

	wg.Wait()

	// Check for errors
	select {
	case err := <-errChan:
		return nil, err
	default:
	}

	// Generate ML model recommendations (depends on other analyses)
	mlRecs, err := analyzer.GenerateMLModelRecommendations(ctx, result)
	if err != nil {
		analyzer.logger.Warn("failed to generate ML model recommendations", zap.Error(err))
	} else {
		result.MLModelRecommendations = mlRecs
	}

	// Calculate summary metrics
	analyzer.calculateSummaryMetrics(result, feedback)

	// Set processing time
	result.ProcessingTimeMs = time.Since(startTime).Milliseconds()
	result.FeedbackCount = len(feedback)

	// Cache the result
	if analyzer.config.CacheAnalysisResults {
		analyzer.mu.Lock()
		analyzer.analysisCache[analysisID] = result
		analyzer.mu.Unlock()
	}

	analyzer.logger.Info("comprehensive ML-aware feedback analysis completed",
		zap.String("analysis_id", analysisID),
		zap.Int64("processing_time_ms", result.ProcessingTimeMs),
		zap.Float64("overall_accuracy", result.OverallAccuracy),
		zap.Int("recommendation_count", result.RecommendationCount))

	return result, nil
}

// calculateSummaryMetrics calculates summary metrics for the analysis result
func (analyzer *MLAwareAnalyzer) calculateSummaryMetrics(result *MLAnalysisResult, feedback []*UserFeedback) {
	if len(feedback) == 0 {
		return
	}

	var totalAccuracy, totalConfidence float64
	var errorCount, securityViolationCount int

	for _, fb := range feedback {
		if fb.FeedbackType == FeedbackTypeAccuracy {
			if accuracy, ok := fb.FeedbackValue["accuracy"].(float64); ok {
				totalAccuracy += accuracy
			}
		}
		if fb.ConfidenceScore > 0 {
			totalConfidence += fb.ConfidenceScore
		}
		if fb.FeedbackType == FeedbackTypeCorrection {
			errorCount++
		}
		if fb.FeedbackType == FeedbackTypeSecurityValidation {
			if violated, ok := fb.FeedbackValue["violation"].(bool); ok && violated {
				securityViolationCount++
			}
		}
	}

	result.OverallAccuracy = totalAccuracy / float64(len(feedback))
	result.OverallConfidence = totalConfidence / float64(len(feedback))
	result.ErrorRate = float64(errorCount) / float64(len(feedback))
	result.SecurityViolationRate = float64(securityViolationCount) / float64(len(feedback))
	result.RecommendationCount = len(result.MLModelRecommendations) +
		len(result.EnsembleWeightRecommendations) +
		len(result.SecurityRecommendations)
}

// generateAnalysisID generates a unique analysis ID
func generateAnalysisID() string {
	return fmt.Sprintf("ml_analysis_%d", time.Now().UnixNano())
}
