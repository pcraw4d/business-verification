package classification_monitoring

import (
	"context"
	"crypto/rand"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// generateID generates a unique identifier
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// PatternAnalysisEngine performs deep analysis of misclassification patterns
type PatternAnalysisEngine struct {
	config           *PatternAnalysisConfig
	logger           *zap.Logger
	mu               sync.RWMutex
	patterns         map[string]*MisclassificationPattern
	patternHistory   []*PatternAnalysisResult
	rootCauseAnalyzer *RootCauseAnalyzer
	recommendationEngine *RecommendationEngine
	predictiveAnalyzer *PredictiveAnalyzer
	startTime        time.Time
}

// PatternAnalysisConfig holds configuration for pattern analysis
type PatternAnalysisConfig struct {
	EnableDeepAnalysis      bool          `json:"enable_deep_analysis"`
	EnablePredictiveAnalysis bool         `json:"enable_predictive_analysis"`
	EnableRootCauseAnalysis bool          `json:"enable_root_cause_analysis"`
	PatternRetentionPeriod  time.Duration `json:"pattern_retention_period"`
	AnalysisWindowSize      time.Duration `json:"analysis_window_size"`
	MinPatternOccurrences   int           `json:"min_pattern_occurrences"`
	ConfidenceThreshold     float64       `json:"confidence_threshold"`
	EnableRealTimeAnalysis  bool          `json:"enable_real_time_analysis"`
	MaxPatternsPerCategory  int           `json:"max_patterns_per_category"`
	EnableCrossDimensionalAnalysis bool   `json:"enable_cross_dimensional_analysis"`
}

// MisclassificationPattern represents a detected pattern in misclassifications
type MisclassificationPattern struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	Description           string                 `json:"description"`
	PatternType           PatternType            `json:"pattern_type"`
	Category              PatternCategory        `json:"category"`
	Severity              PatternSeverity        `json:"severity"`
	Confidence            float64                `json:"confidence"`
	OccurrenceCount       int                    `json:"occurrence_count"`
	FirstSeen             time.Time              `json:"first_seen"`
	LastSeen              time.Time              `json:"last_seen"`
	Frequency             float64                `json:"frequency"` // occurrences per hour
	Trend                 TrendDirection         `json:"trend"`
	Characteristics       PatternCharacteristics  `json:"characteristics"`
	RootCauses            []RootCause            `json:"root_causes"`
	Recommendations       []Recommendation       `json:"recommendations"`
	ImpactScore           float64                `json:"impact_score"`
	BusinessImpact        BusinessImpact         `json:"business_impact"`
	Metadata              map[string]interface{} `json:"metadata"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

// PatternType defines the type of misclassification pattern
type PatternType string

const (
	PatternTypeTemporal     PatternType = "temporal"
	PatternTypeSemantic     PatternType = "semantic"
	PatternTypeConfidence   PatternType = "confidence"
	PatternTypeInput        PatternType = "input"
	PatternTypeMethod       PatternType = "method"
	PatternTypeIndustry     PatternType = "industry"
	PatternTypeCrossDimensional PatternType = "cross_dimensional"
	PatternTypeAnomaly      PatternType = "anomaly"
	PatternTypeTrend        PatternType = "trend"
	PatternTypeSeasonal     PatternType = "seasonal"
)

// PatternCategory defines the category of pattern
type PatternCategory string

const (
	PatternCategoryDataQuality     PatternCategory = "data_quality"
	PatternCategoryModelPerformance PatternCategory = "model_performance"
	PatternCategoryInputProcessing PatternCategory = "input_processing"
	PatternCategoryBusinessLogic   PatternCategory = "business_logic"
	PatternCategoryExternalFactors PatternCategory = "external_factors"
	PatternCategorySystemIssues    PatternCategory = "system_issues"
	PatternCategoryUserBehavior    PatternCategory = "user_behavior"
	PatternCategoryConfiguration   PatternCategory = "configuration"
)

// PatternSeverity defines the severity level of a pattern
type PatternSeverity string

const (
	PatternSeverityLow      PatternSeverity = "low"
	PatternSeverityMedium   PatternSeverity = "medium"
	PatternSeverityHigh     PatternSeverity = "high"
	PatternSeverityCritical PatternSeverity = "critical"
)

// TrendDirection indicates the trend of a pattern
type TrendDirection string

const (
	TrendDirectionIncreasing TrendDirection = "increasing"
	TrendDirectionDecreasing TrendDirection = "decreasing"
	TrendDirectionStable     TrendDirection = "stable"
	TrendDirectionFluctuating TrendDirection = "fluctuating"
)

// PatternCharacteristics describes the characteristics of a pattern
type PatternCharacteristics struct {
	TimeOfDayDistribution    map[string]int     `json:"time_of_day_distribution"`
	DayOfWeekDistribution    map[string]int     `json:"day_of_week_distribution"`
	ConfidenceDistribution   map[string]int     `json:"confidence_distribution"`
	MethodDistribution       map[string]int     `json:"method_distribution"`
	IndustryDistribution     map[string]int     `json:"industry_distribution"`
	InputLengthDistribution  map[string]int     `json:"input_length_distribution"`
	ErrorTypeDistribution    map[string]int     `json:"error_type_distribution"`
	GeographicDistribution   map[string]int     `json:"geographic_distribution"`
	UserAgentDistribution    map[string]int     `json:"user_agent_distribution"`
	CommonKeywords           []string           `json:"common_keywords"`
	CommonPhrases            []string           `json:"common_phrases"`
	InputPatterns            []string           `json:"input_patterns"`
	CorrelationFactors       map[string]float64 `json:"correlation_factors"`
}

// RootCause represents a root cause of misclassifications
type RootCause struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Confidence  float64 `json:"confidence"`
	Evidence    []string `json:"evidence"`
	Impact      float64 `json:"impact"`
	Fixable     bool    `json:"fixable"`
	Priority    int     `json:"priority"`
}

// BusinessImpact describes the business impact of a pattern
type BusinessImpact struct {
	AccuracyImpact    float64 `json:"accuracy_impact"`
	UserExperienceImpact float64 `json:"user_experience_impact"`
	RevenueImpact     float64 `json:"revenue_impact"`
	ComplianceImpact  float64 `json:"compliance_impact"`
	OperationalImpact float64 `json:"operational_impact"`
	RiskLevel         string  `json:"risk_level"`
}

// PatternAnalysisResult represents the result of pattern analysis
type PatternAnalysisResult struct {
	ID              string                    `json:"id"`
	AnalysisTime    time.Time                 `json:"analysis_time"`
	PatternsFound   int                       `json:"patterns_found"`
	PatternsUpdated int                       `json:"patterns_updated"`
	NewPatterns     int                       `json:"new_patterns"`
	CriticalPatterns int                      `json:"critical_patterns"`
	HighImpactPatterns int                    `json:"high_impact_patterns"`
	Recommendations []Recommendation          `json:"recommendations"`
	Summary         PatternAnalysisSummary    `json:"summary"`
	Metadata        map[string]interface{}    `json:"metadata"`
}

// PatternAnalysisSummary provides a summary of pattern analysis
type PatternAnalysisSummary struct {
	TotalPatterns       int                       `json:"total_patterns"`
	PatternsByType      map[PatternType]int       `json:"patterns_by_type"`
	PatternsByCategory  map[PatternCategory]int   `json:"patterns_by_category"`
	PatternsBySeverity  map[PatternSeverity]int   `json:"patterns_by_severity"`
	TopPatterns         []*MisclassificationPattern `json:"top_patterns"`
	TrendingPatterns    []*MisclassificationPattern `json:"trending_patterns"`
	CriticalPatterns    []*MisclassificationPattern `json:"critical_patterns"`
	ImpactScore         float64                   `json:"impact_score"`
	RiskLevel           string                    `json:"risk_level"`
}

// RecommendationEngine generates recommendations based on patterns
type RecommendationEngine struct {
	logger *zap.Logger
}

// NewRecommendationEngine creates a new recommendation engine
func NewRecommendationEngine(logger *zap.Logger) *RecommendationEngine {
	return &RecommendationEngine{
		logger: logger,
	}
}

// GenerateRecommendations generates recommendations based on patterns
func (re *RecommendationEngine) GenerateRecommendations(patterns []*MisclassificationPattern) []Recommendation {
	recommendations := make([]Recommendation, 0)
	
	for _, pattern := range patterns {
		switch pattern.PatternType {
		case PatternTypeConfidence:
			if pattern.Severity == PatternSeverityCritical {
				recommendations = append(recommendations, Recommendation{
					ID:          generateID(),
					Type:        "model_calibration",
					Priority:    "high",
					Title:       "Model Confidence Calibration",
					Description: "Calibrate model confidence scores to reduce high-confidence misclassifications",
					Actions:     []string{"Review confidence thresholds", "Retrain model with calibration data"},
					Impact:      "high",
					Effort:      "medium",
				})
			}
		case PatternTypeTemporal:
			recommendations = append(recommendations, Recommendation{
				ID:          generateID(),
				Type:        "load_balancing",
				Priority:    "medium",
				Title:       "Load Balancing Implementation",
				Description: "Implement load balancing during peak hours to improve classification accuracy",
				Actions:     []string{"Deploy additional resources", "Implement request queuing"},
				Impact:      "medium",
				Effort:      "high",
			})
		case PatternTypeSemantic:
			recommendations = append(recommendations, Recommendation{
				ID:          generateID(),
				Type:        "training_data",
				Priority:    "medium",
				Title:       "Training Data Enhancement",
				Description: "Add training data for problematic keywords and phrases",
				Actions:     []string{"Collect additional training samples", "Annotate problematic cases"},
				Impact:      "high",
				Effort:      "medium",
			})
		}
	}
	
	return recommendations
}

// PredictiveAnalyzer performs predictive analysis on patterns
type PredictiveAnalyzer struct {
	config *PatternAnalysisConfig
	logger *zap.Logger
}

// NewPredictiveAnalyzer creates a new predictive analyzer
func NewPredictiveAnalyzer(config *PatternAnalysisConfig, logger *zap.Logger) *PredictiveAnalyzer {
	return &PredictiveAnalyzer{
		config: config,
		logger: logger,
	}
}

// NewPatternAnalysisEngine creates a new pattern analysis engine
func NewPatternAnalysisEngine(config *PatternAnalysisConfig, logger *zap.Logger) *PatternAnalysisEngine {
	if config == nil {
		config = &PatternAnalysisConfig{
			EnableDeepAnalysis:         true,
			EnablePredictiveAnalysis:   true,
			EnableRootCauseAnalysis:    true,
			PatternRetentionPeriod:     30 * 24 * time.Hour, // 30 days
			AnalysisWindowSize:         1 * time.Hour,
			MinPatternOccurrences:      5,
			ConfidenceThreshold:        0.7,
			EnableRealTimeAnalysis:     true,
			MaxPatternsPerCategory:     10,
			EnableCrossDimensionalAnalysis: true,
		}
	}

	engine := &PatternAnalysisEngine{
		config:         config,
		logger:         logger,
		patterns:       make(map[string]*MisclassificationPattern),
		patternHistory: make([]*PatternAnalysisResult, 0),
		startTime:      time.Now(),
	}

	// Initialize sub-components
	engine.rootCauseAnalyzer = NewRootCauseAnalyzer(&DetectionConfig{}, logger)
	engine.recommendationEngine = NewRecommendationEngine(logger)
	engine.predictiveAnalyzer = NewPredictiveAnalyzer(config, logger)

	return engine
}

// AnalyzeMisclassifications performs comprehensive pattern analysis on misclassifications
func (pae *PatternAnalysisEngine) AnalyzeMisclassifications(ctx context.Context, misclassifications []*MisclassificationRecord) (*PatternAnalysisResult, error) {
	pae.mu.Lock()
	defer pae.mu.Unlock()

	startTime := time.Now()
	pae.logger.Info("Starting pattern analysis", 
		zap.Int("misclassifications_count", len(misclassifications)))

	// Perform different types of pattern analysis
	patterns := make([]*MisclassificationPattern, 0)

	// Temporal pattern analysis
	if temporalPatterns, err := pae.analyzeTemporalPatterns(misclassifications); err == nil {
		patterns = append(patterns, temporalPatterns...)
	}

	// Semantic pattern analysis
	if semanticPatterns, err := pae.analyzeSemanticPatterns(misclassifications); err == nil {
		patterns = append(patterns, semanticPatterns...)
	}

	// Confidence pattern analysis
	if confidencePatterns, err := pae.analyzeConfidencePatterns(misclassifications); err == nil {
		patterns = append(patterns, confidencePatterns...)
	}

	// Input pattern analysis
	if inputPatterns, err := pae.analyzeInputPatterns(misclassifications); err == nil {
		patterns = append(patterns, inputPatterns...)
	}

	// Cross-dimensional pattern analysis
	if pae.config.EnableCrossDimensionalAnalysis {
		if crossPatterns, err := pae.analyzeCrossDimensionalPatterns(misclassifications); err == nil {
			patterns = append(patterns, crossPatterns...)
		}
	}

	// Anomaly detection
	if anomalyPatterns, err := pae.detectAnomalies(misclassifications); err == nil {
		patterns = append(patterns, anomalyPatterns...)
	}

	// Update existing patterns and add new ones
	patternsUpdated := 0
	newPatterns := 0
	for _, pattern := range patterns {
		if existing, exists := pae.patterns[pattern.ID]; exists {
			// Update existing pattern
			pae.updatePattern(existing, pattern)
			patternsUpdated++
		} else {
			// Add new pattern
			pae.patterns[pattern.ID] = pattern
			newPatterns++
		}
	}

	// Generate recommendations
	recommendations := pae.recommendationEngine.GenerateRecommendations(patterns)

	// Create analysis result
	result := &PatternAnalysisResult{
		ID:              generateID(),
		AnalysisTime:    time.Now(),
		PatternsFound:   len(patterns),
		PatternsUpdated: patternsUpdated,
		NewPatterns:     newPatterns,
		Recommendations: recommendations,
		Summary:         pae.generateAnalysisSummary(patterns),
		Metadata: map[string]interface{}{
			"analysis_duration": time.Since(startTime).String(),
			"config":           pae.config,
		},
	}

	// Add to history
	pae.patternHistory = append(pae.patternHistory, result)

	// Cleanup old patterns
	pae.cleanupOldPatterns()

	pae.logger.Info("Pattern analysis completed",
		zap.Int("patterns_found", len(patterns)),
		zap.Int("patterns_updated", patternsUpdated),
		zap.Int("new_patterns", newPatterns),
		zap.Duration("duration", time.Since(startTime)))

	return result, nil
}

// analyzeTemporalPatterns analyzes temporal patterns in misclassifications
func (pae *PatternAnalysisEngine) analyzeTemporalPatterns(misclassifications []*MisclassificationRecord) ([]*MisclassificationPattern, error) {
	patterns := make([]*MisclassificationPattern, 0)

	// Group by time periods
	hourlyDistribution := make(map[int]int)
	dailyDistribution := make(map[string]int)
	weeklyDistribution := make(map[string]int)

	for _, mis := range misclassifications {
		hour := mis.Timestamp.Hour()
		hourlyDistribution[hour]++

		day := mis.Timestamp.Format("Monday")
		dailyDistribution[day]++

		week := mis.Timestamp.Format("2006-01-02")
		weeklyDistribution[week]++
	}

	// Detect time-based patterns
	if pattern := pae.detectTimeOfDayPattern(hourlyDistribution, misclassifications); pattern != nil {
		patterns = append(patterns, pattern)
	}

	if pattern := pae.detectDayOfWeekPattern(dailyDistribution, misclassifications); pattern != nil {
		patterns = append(patterns, pattern)
	}

	if pattern := pae.detectWeeklyPattern(weeklyDistribution, misclassifications); pattern != nil {
		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// analyzeSemanticPatterns analyzes semantic patterns in misclassifications
func (pae *PatternAnalysisEngine) analyzeSemanticPatterns(misclassifications []*MisclassificationRecord) ([]*MisclassificationPattern, error) {
	patterns := make([]*MisclassificationPattern, 0)

	// Extract common keywords and phrases
	keywordFrequency := make(map[string]int)
	phraseFrequency := make(map[string]int)

	for _, mis := range misclassifications {
		// Analyze input text for keywords - convert to string
		inputText := pae.extractInputText(mis.InputData)
		keywords := pae.extractKeywords(inputText)
		for _, keyword := range keywords {
			keywordFrequency[keyword]++
		}

		// Analyze for common phrases
		phrases := pae.extractPhrases(inputText)
		for _, phrase := range phrases {
			phraseFrequency[phrase]++
		}
	}

	// Detect keyword-based patterns
	if pattern := pae.detectKeywordPattern(keywordFrequency, misclassifications); pattern != nil {
		patterns = append(patterns, pattern)
	}

	// Detect phrase-based patterns
	if pattern := pae.detectPhrasePattern(phraseFrequency, misclassifications); pattern != nil {
		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// analyzeConfidencePatterns analyzes confidence-related patterns
func (pae *PatternAnalysisEngine) analyzeConfidencePatterns(misclassifications []*MisclassificationRecord) ([]*MisclassificationPattern, error) {
	patterns := make([]*MisclassificationPattern, 0)

	// Group by confidence ranges
	confidenceRanges := map[string][]*MisclassificationRecord{
		"high":   make([]*MisclassificationRecord, 0),
		"medium": make([]*MisclassificationRecord, 0),
		"low":    make([]*MisclassificationRecord, 0),
	}

	for _, mis := range misclassifications {
		confidence := mis.ConfidenceScore // Use correct field name
		switch {
		case confidence >= 0.8:
			confidenceRanges["high"] = append(confidenceRanges["high"], mis)
		case confidence >= 0.5:
			confidenceRanges["medium"] = append(confidenceRanges["medium"], mis)
		default:
			confidenceRanges["low"] = append(confidenceRanges["low"], mis)
		}
	}

	// Detect high-confidence error patterns (most critical)
	if len(confidenceRanges["high"]) > pae.config.MinPatternOccurrences {
		pattern := &MisclassificationPattern{
			ID:          generateID(),
			Name:        "High Confidence Misclassifications",
			Description: "Misclassifications occurring with high confidence scores",
			PatternType: PatternTypeConfidence,
			Category:    PatternCategoryModelPerformance,
			Severity:    PatternSeverityCritical,
			Confidence:  pae.calculatePatternConfidence(confidenceRanges["high"]),
			OccurrenceCount: len(confidenceRanges["high"]),
			FirstSeen:   time.Now(),
			LastSeen:    time.Now(),
			Frequency:   float64(len(confidenceRanges["high"])) / pae.config.AnalysisWindowSize.Hours(),
			Trend:       TrendDirectionStable,
			Characteristics: PatternCharacteristics{
				ConfidenceDistribution: map[string]int{
					"high": len(confidenceRanges["high"]),
				},
			},
			RootCauses: pae.analyzeRootCauses(confidenceRanges["high"]),
			ImpactScore: 0.9, // High impact for high-confidence errors
			BusinessImpact: BusinessImpact{
				AccuracyImpact: 0.8,
				UserExperienceImpact: 0.7,
				RiskLevel: "high",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		patterns = append(patterns, pattern)
	}

	return patterns, nil
}

// analyzeInputPatterns analyzes input-related patterns
func (pae *PatternAnalysisEngine) analyzeInputPatterns(misclassifications []*MisclassificationRecord) ([]*MisclassificationPattern, error) {
	patterns := make([]*MisclassificationPattern, 0)

	// Analyze input length patterns
	lengthDistribution := make(map[string]int)
	for _, mis := range misclassifications {
		inputText := pae.extractInputText(mis.InputData)
		inputLength := len(inputText)
		switch {
		case inputLength < 10:
			lengthDistribution["very_short"]++
		case inputLength < 50:
			lengthDistribution["short"]++
		case inputLength < 200:
			lengthDistribution["medium"]++
		case inputLength < 500:
			lengthDistribution["long"]++
		default:
			lengthDistribution["very_long"]++
		}
	}

	// Detect input length patterns
	for lengthType, count := range lengthDistribution {
		if count >= pae.config.MinPatternOccurrences {
			pattern := &MisclassificationPattern{
				ID:          generateID(),
				Name:        fmt.Sprintf("%s Input Length Misclassifications", strings.Title(lengthType)),
				Description: fmt.Sprintf("Misclassifications for %s input length", lengthType),
				PatternType: PatternTypeInput,
				Category:    PatternCategoryInputProcessing,
				Severity:    pae.calculateSeverity(count, len(misclassifications)),
				Confidence:  pae.calculatePatternConfidence(misclassifications),
				OccurrenceCount: count,
				FirstSeen:   time.Now(),
				LastSeen:    time.Now(),
				Frequency:   float64(count) / pae.config.AnalysisWindowSize.Hours(),
				Trend:       TrendDirectionStable,
				Characteristics: PatternCharacteristics{
					InputLengthDistribution: map[string]int{
						lengthType: count,
					},
				},
				RootCauses: []RootCause{
					{
						ID:          generateID(),
						Type:        "input_processing",
						Description: fmt.Sprintf("Issues with processing %s input length", lengthType),
						Confidence:  0.7,
						Fixable:     true,
						Priority:    2,
					},
				},
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns, nil
}

// analyzeCrossDimensionalPatterns analyzes patterns across multiple dimensions
func (pae *PatternAnalysisEngine) analyzeCrossDimensionalPatterns(misclassifications []*MisclassificationRecord) ([]*MisclassificationPattern, error) {
	patterns := make([]*MisclassificationPattern, 0)

	// Analyze method + confidence combinations
	methodConfidenceMap := make(map[string]map[string]int)
	for _, mis := range misclassifications {
		method := mis.ClassificationMethod
		confidenceLevel := pae.getConfidenceLevel(mis.ConfidenceScore)
		
		if methodConfidenceMap[method] == nil {
			methodConfidenceMap[method] = make(map[string]int)
		}
		methodConfidenceMap[method][confidenceLevel]++
	}

	// Detect problematic method-confidence combinations
	for method, confidenceMap := range methodConfidenceMap {
		for confidenceLevel, count := range confidenceMap {
			if count >= pae.config.MinPatternOccurrences {
				pattern := &MisclassificationPattern{
					ID:          generateID(),
					Name:        fmt.Sprintf("%s Method %s Confidence Errors", method, confidenceLevel),
					Description: fmt.Sprintf("Misclassifications for %s method with %s confidence", method, confidenceLevel),
					PatternType: PatternTypeCrossDimensional,
					Category:    PatternCategoryModelPerformance,
					Severity:    pae.calculateSeverity(count, len(misclassifications)),
					Confidence:  pae.calculatePatternConfidence(misclassifications),
					OccurrenceCount: count,
					FirstSeen:   time.Now(),
					LastSeen:    time.Now(),
					Frequency:   float64(count) / pae.config.AnalysisWindowSize.Hours(),
					Trend:       TrendDirectionStable,
					Characteristics: PatternCharacteristics{
						MethodDistribution: map[string]int{
							method: count,
						},
						ConfidenceDistribution: map[string]int{
							confidenceLevel: count,
						},
					},
					RootCauses: []RootCause{
						{
							ID:          generateID(),
							Type:        "method_performance",
							Description: fmt.Sprintf("Performance issues with %s method for %s confidence inputs", method, confidenceLevel),
							Confidence:  0.8,
							Fixable:     true,
							Priority:    1,
						},
					},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns, nil
}

// analyzeRootCauses analyzes root causes for a set of misclassifications
func (pae *PatternAnalysisEngine) analyzeRootCauses(misclassifications []*MisclassificationRecord) []RootCause {
	rootCauses := make([]RootCause, 0)
	
	// Simple root cause analysis - in production, use more sophisticated methods
	if len(misclassifications) > 0 {
		// Analyze common patterns
		highConfidenceCount := 0
		for _, mis := range misclassifications {
			if mis.ConfidenceScore >= 0.8 {
				highConfidenceCount++
			}
		}
		
		if highConfidenceCount > len(misclassifications)/2 {
			rootCauses = append(rootCauses, RootCause{
				ID:          generateID(),
				Type:        "model_calibration",
				Description: "Model confidence calibration issues",
				Confidence:  0.7,
				Fixable:     true,
				Priority:    1,
			})
		}
		
		// Add generic root cause
		rootCauses = append(rootCauses, RootCause{
			ID:          generateID(),
			Type:        "classification_error",
			Description: "General classification accuracy issues",
			Confidence:  0.6,
			Fixable:     true,
			Priority:    2,
		})
	}
	
	return rootCauses
}

// detectAnomalies detects anomalous patterns in misclassifications
func (pae *PatternAnalysisEngine) detectAnomalies(misclassifications []*MisclassificationRecord) ([]*MisclassificationPattern, error) {
	patterns := make([]*MisclassificationPattern, 0)

	// Detect sudden spikes in misclassifications
	if anomalyPattern := pae.detectSpikeAnomaly(misclassifications); anomalyPattern != nil {
		patterns = append(patterns, anomalyPattern)
	}

	// Detect unusual confidence distributions
	if anomalyPattern := pae.detectConfidenceAnomaly(misclassifications); anomalyPattern != nil {
		patterns = append(patterns, anomalyPattern)
	}

	// Detect unusual input patterns
	if anomalyPattern := pae.detectInputAnomaly(misclassifications); anomalyPattern != nil {
		patterns = append(patterns, anomalyPattern)
	}

	return patterns, nil
}

// Helper methods for pattern detection
func (pae *PatternAnalysisEngine) detectTimeOfDayPattern(distribution map[int]int, misclassifications []*MisclassificationRecord) *MisclassificationPattern {
	// Find peak hours
	var peakHour int
	maxCount := 0
	for hour, count := range distribution {
		if count > maxCount {
			maxCount = count
			peakHour = hour
		}
	}

	if maxCount >= pae.config.MinPatternOccurrences {
		return &MisclassificationPattern{
			ID:          generateID(),
			Name:        fmt.Sprintf("Peak Hour Misclassifications (%02d:00)", peakHour),
			Description: fmt.Sprintf("Increased misclassifications during hour %02d:00", peakHour),
			PatternType: PatternTypeTemporal,
			Category:    PatternCategorySystemIssues,
			Severity:    pae.calculateSeverity(maxCount, len(misclassifications)),
			Confidence:  pae.calculatePatternConfidence(misclassifications),
			OccurrenceCount: maxCount,
			FirstSeen:   time.Now(),
			LastSeen:    time.Now(),
			Frequency:   float64(maxCount) / pae.config.AnalysisWindowSize.Hours(),
			Trend:       TrendDirectionStable,
			Characteristics: PatternCharacteristics{
				TimeOfDayDistribution: map[string]int{
					fmt.Sprintf("%02d:00", peakHour): maxCount,
				},
			},
			RootCauses: []RootCause{
				{
					ID:          generateID(),
					Type:        "load_related",
					Description: fmt.Sprintf("Increased load during hour %02d:00 affecting classification accuracy", peakHour),
					Confidence:  0.6,
					Fixable:     true,
					Priority:    3,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	return nil
}

func (pae *PatternAnalysisEngine) detectDayOfWeekPattern(distribution map[string]int, misclassifications []*MisclassificationRecord) *MisclassificationPattern {
	// Find peak day
	var peakDay string
	maxCount := 0
	for day, count := range distribution {
		if count > maxCount {
			maxCount = count
			peakDay = day
		}
	}

	if maxCount >= pae.config.MinPatternOccurrences {
		return &MisclassificationPattern{
			ID:          generateID(),
			Name:        fmt.Sprintf("Peak Day Misclassifications (%s)", peakDay),
			Description: fmt.Sprintf("Increased misclassifications on %s", peakDay),
			PatternType: PatternTypeTemporal,
			Category:    PatternCategoryUserBehavior,
			Severity:    pae.calculateSeverity(maxCount, len(misclassifications)),
			Confidence:  pae.calculatePatternConfidence(misclassifications),
			OccurrenceCount: maxCount,
			FirstSeen:   time.Now(),
			LastSeen:    time.Now(),
			Frequency:   float64(maxCount) / (pae.config.AnalysisWindowSize.Hours() / 24),
			Trend:       TrendDirectionStable,
			Characteristics: PatternCharacteristics{
				DayOfWeekDistribution: map[string]int{
					peakDay: maxCount,
				},
			},
			RootCauses: []RootCause{
				{
					ID:          generateID(),
					Type:        "user_behavior",
					Description: fmt.Sprintf("Different user behavior patterns on %s", peakDay),
					Confidence:  0.5,
					Fixable:     false,
					Priority:    4,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	return nil
}

func (pae *PatternAnalysisEngine) detectWeeklyPattern(distribution map[string]int, misclassifications []*MisclassificationRecord) *MisclassificationPattern {
	// This would implement weekly trend analysis
	// For now, return nil as this is a placeholder
	return nil
}

func (pae *PatternAnalysisEngine) detectKeywordPattern(keywordFrequency map[string]int, misclassifications []*MisclassificationRecord) *MisclassificationPattern {
	// Find most common problematic keywords
	var topKeyword string
	maxCount := 0
	for keyword, count := range keywordFrequency {
		if count > maxCount {
			maxCount = count
			topKeyword = keyword
		}
	}

	if maxCount >= pae.config.MinPatternOccurrences {
		return &MisclassificationPattern{
			ID:          generateID(),
			Name:        fmt.Sprintf("Keyword-Based Misclassifications (%s)", topKeyword),
			Description: fmt.Sprintf("Misclassifications associated with keyword '%s'", topKeyword),
			PatternType: PatternTypeSemantic,
			Category:    PatternCategoryModelPerformance,
			Severity:    pae.calculateSeverity(maxCount, len(misclassifications)),
			Confidence:  pae.calculatePatternConfidence(misclassifications),
			OccurrenceCount: maxCount,
			FirstSeen:   time.Now(),
			LastSeen:    time.Now(),
			Frequency:   float64(maxCount) / pae.config.AnalysisWindowSize.Hours(),
			Trend:       TrendDirectionStable,
			Characteristics: PatternCharacteristics{
				CommonKeywords: []string{topKeyword},
			},
			RootCauses: []RootCause{
				{
					ID:          generateID(),
					Type:        "semantic_understanding",
					Description: fmt.Sprintf("Model struggles with semantic understanding of keyword '%s'", topKeyword),
					Confidence:  0.7,
					Fixable:     true,
					Priority:    2,
				},
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}

	return nil
}

func (pae *PatternAnalysisEngine) detectPhrasePattern(phraseFrequency map[string]int, misclassifications []*MisclassificationRecord) *MisclassificationPattern {
	// Similar to keyword pattern but for phrases
	// Implementation would be similar to detectKeywordPattern
	return nil
}

func (pae *PatternAnalysisEngine) detectSpikeAnomaly(misclassifications []*MisclassificationRecord) *MisclassificationPattern {
	// Detect sudden spikes in misclassification rate
	// This would compare current rate to historical baseline
	// For now, return nil as this is a placeholder
	return nil
}

func (pae *PatternAnalysisEngine) detectConfidenceAnomaly(misclassifications []*MisclassificationRecord) *MisclassificationPattern {
	// Detect unusual confidence distributions
	// This would analyze confidence patterns that deviate from normal
	// For now, return nil as this is a placeholder
	return nil
}

func (pae *PatternAnalysisEngine) detectInputAnomaly(misclassifications []*MisclassificationRecord) *MisclassificationPattern {
	// Detect unusual input patterns
	// This would analyze input characteristics that are anomalous
	// For now, return nil as this is a placeholder
	return nil
}

// Helper utility methods
func (pae *PatternAnalysisEngine) extractInputText(inputData map[string]interface{}) string {
	// Extract text from input data map
	if text, ok := inputData["text"].(string); ok {
		return text
	}
	if name, ok := inputData["name"].(string); ok {
		return name
	}
	if description, ok := inputData["description"].(string); ok {
		return description
	}
	// Fallback to empty string
	return ""
}

func (pae *PatternAnalysisEngine) extractKeywords(text string) []string {
	// Simple keyword extraction - in production, use NLP libraries
	words := strings.Fields(strings.ToLower(text))
	keywords := make([]string, 0)
	
	// Filter out common stop words and short words
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true, "with": true, "by": true,
	}
	
	for _, word := range words {
		if len(word) > 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}
	
	return keywords
}

func (pae *PatternAnalysisEngine) extractPhrases(text string) []string {
	// Simple phrase extraction - in production, use NLP libraries
	// This is a basic implementation
	phrases := make([]string, 0)
	words := strings.Fields(text)
	
	for i := 0; i < len(words)-1; i++ {
		phrase := words[i] + " " + words[i+1]
		phrases = append(phrases, phrase)
	}
	
	return phrases
}

func (pae *PatternAnalysisEngine) getConfidenceLevel(confidence float64) string {
	switch {
	case confidence >= 0.8:
		return "high"
	case confidence >= 0.5:
		return "medium"
	default:
		return "low"
	}
}

func (pae *PatternAnalysisEngine) calculatePatternConfidence(misclassifications []*MisclassificationRecord) float64 {
	// Calculate confidence based on pattern strength and consistency
	if len(misclassifications) == 0 {
		return 0.0
	}
	
	// Simple confidence calculation - in production, use more sophisticated methods
	confidenceSum := 0.0
	for _, mis := range misclassifications {
		confidenceSum += mis.ConfidenceScore
	}
	
	return confidenceSum / float64(len(misclassifications))
}

func (pae *PatternAnalysisEngine) calculateSeverity(count, total int) PatternSeverity {
	percentage := float64(count) / float64(total)
	
	switch {
	case percentage >= 0.3:
		return PatternSeverityCritical
	case percentage >= 0.15:
		return PatternSeverityHigh
	case percentage >= 0.05:
		return PatternSeverityMedium
	default:
		return PatternSeverityLow
	}
}

func (pae *PatternAnalysisEngine) updatePattern(existing, new *MisclassificationPattern) {
	existing.OccurrenceCount += new.OccurrenceCount
	existing.LastSeen = new.LastSeen
	existing.Frequency = float64(existing.OccurrenceCount) / time.Since(existing.FirstSeen).Hours()
	existing.UpdatedAt = time.Now()
	
	// Update characteristics
	if existing.Characteristics.TimeOfDayDistribution == nil {
		existing.Characteristics.TimeOfDayDistribution = make(map[string]int)
	}
	for key, value := range new.Characteristics.TimeOfDayDistribution {
		existing.Characteristics.TimeOfDayDistribution[key] += value
	}
	
	// Update root causes if new ones are found
	for _, newRootCause := range new.RootCauses {
		found := false
		for _, existingRootCause := range existing.RootCauses {
			if existingRootCause.Type == newRootCause.Type {
				found = true
				break
			}
		}
		if !found {
			existing.RootCauses = append(existing.RootCauses, newRootCause)
		}
	}
}

func (pae *PatternAnalysisEngine) generateAnalysisSummary(patterns []*MisclassificationPattern) PatternAnalysisSummary {
	summary := PatternAnalysisSummary{
		PatternsByType:     make(map[PatternType]int),
		PatternsByCategory: make(map[PatternCategory]int),
		PatternsBySeverity: make(map[PatternSeverity]int),
		TopPatterns:        make([]*MisclassificationPattern, 0),
		TrendingPatterns:   make([]*MisclassificationPattern, 0),
		CriticalPatterns:   make([]*MisclassificationPattern, 0),
	}

	summary.TotalPatterns = len(patterns)

	// Categorize patterns
	for _, pattern := range patterns {
		summary.PatternsByType[pattern.PatternType]++
		summary.PatternsByCategory[pattern.Category]++
		summary.PatternsBySeverity[pattern.Severity]++

		// Track critical patterns
		if pattern.Severity == PatternSeverityCritical {
			summary.CriticalPatterns = append(summary.CriticalPatterns, pattern)
		}

		// Track high-impact patterns
		if pattern.ImpactScore >= 0.7 {
			summary.TopPatterns = append(summary.TopPatterns, pattern)
		}
	}

	// Sort patterns by impact score
	sort.Slice(summary.TopPatterns, func(i, j int) bool {
		return summary.TopPatterns[i].ImpactScore > summary.TopPatterns[j].ImpactScore
	})

	// Limit to top 5
	if len(summary.TopPatterns) > 5 {
		summary.TopPatterns = summary.TopPatterns[:5]
	}

	// Calculate overall impact score
	totalImpact := 0.0
	for _, pattern := range patterns {
		totalImpact += pattern.ImpactScore
	}
	if len(patterns) > 0 {
		summary.ImpactScore = totalImpact / float64(len(patterns))
	}

	// Determine risk level
	switch {
	case summary.ImpactScore >= 0.8:
		summary.RiskLevel = "critical"
	case summary.ImpactScore >= 0.6:
		summary.RiskLevel = "high"
	case summary.ImpactScore >= 0.4:
		summary.RiskLevel = "medium"
	default:
		summary.RiskLevel = "low"
	}

	return summary
}

func (pae *PatternAnalysisEngine) cleanupOldPatterns() {
	cutoffTime := time.Now().Add(-pae.config.PatternRetentionPeriod)
	
	for id, pattern := range pae.patterns {
		if pattern.LastSeen.Before(cutoffTime) {
			delete(pae.patterns, id)
		}
	}
}

// GetPatterns returns all detected patterns
func (pae *PatternAnalysisEngine) GetPatterns() map[string]*MisclassificationPattern {
	pae.mu.RLock()
	defer pae.mu.RUnlock()
	
	result := make(map[string]*MisclassificationPattern)
	for id, pattern := range pae.patterns {
		result[id] = pattern
	}
	return result
}

// GetPatternsByType returns patterns filtered by type
func (pae *PatternAnalysisEngine) GetPatternsByType(patternType PatternType) []*MisclassificationPattern {
	pae.mu.RLock()
	defer pae.mu.RUnlock()
	
	patterns := make([]*MisclassificationPattern, 0)
	for _, pattern := range pae.patterns {
		if pattern.PatternType == patternType {
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}

// GetPatternsBySeverity returns patterns filtered by severity
func (pae *PatternAnalysisEngine) GetPatternsBySeverity(severity PatternSeverity) []*MisclassificationPattern {
	pae.mu.RLock()
	defer pae.mu.RUnlock()
	
	patterns := make([]*MisclassificationPattern, 0)
	for _, pattern := range pae.patterns {
		if pattern.Severity == severity {
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}

// GetPatternHistory returns the history of pattern analysis results
func (pae *PatternAnalysisEngine) GetPatternHistory() []*PatternAnalysisResult {
	pae.mu.RLock()
	defer pae.mu.RUnlock()
	
	result := make([]*PatternAnalysisResult, len(pae.patternHistory))
	copy(result, pae.patternHistory)
	return result
}
