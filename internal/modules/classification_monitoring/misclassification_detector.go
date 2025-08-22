package classification_monitoring

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MisclassificationDetector detects and analyzes classification errors
type MisclassificationDetector struct {
	config               *DetectionConfig
	logger               *zap.Logger
	mu                   sync.RWMutex
	patterns             map[string]*ErrorPattern
	rootCauseAnalyzer    *RootCauseAnalyzer
	classificationBuffer []*ClassificationEvent
	errorDatabase        map[string]*MisclassificationRecord
	startTime            time.Time
}

// DetectionConfig holds configuration for misclassification detection
type DetectionConfig struct {
	EnablePatternDetection  bool          `json:"enable_pattern_detection"`
	EnableRootCauseAnalysis bool          `json:"enable_root_cause_analysis"`
	EnableRealTimeDetection bool          `json:"enable_real_time_detection"`
	PatternAnalysisWindow   time.Duration `json:"pattern_analysis_window"`
	MinPatternOccurrences   int           `json:"min_pattern_occurrences"`
	ConfidenceThreshold     float64       `json:"confidence_threshold"`
	BufferSize              int           `json:"buffer_size"`
	AnalysisInterval        time.Duration `json:"analysis_interval"`
	EnableSemanticAnalysis  bool          `json:"enable_semantic_analysis"`
	EnableTemporalAnalysis  bool          `json:"enable_temporal_analysis"`
	EnableInputAnalysis     bool          `json:"enable_input_analysis"`
}

// ClassificationEvent represents a classification event for analysis
type ClassificationEvent struct {
	ID                string                 `json:"id"`
	Timestamp         time.Time              `json:"timestamp"`
	BusinessName      string                 `json:"business_name"`
	ExpectedIndustry  string                 `json:"expected_industry"`
	ActualIndustry    string                 `json:"actual_industry"`
	ConfidenceScore   float64                `json:"confidence_score"`
	Method            string                 `json:"method"`
	InputFeatures     map[string]interface{} `json:"input_features"`
	IsCorrect         bool                   `json:"is_correct"`
	ProcessingTime    time.Duration          `json:"processing_time"`
	ModelVersion      string                 `json:"model_version"`
	FeatureImportance map[string]float64     `json:"feature_importance"`
}

// ErrorPattern represents a detected pattern in misclassifications
type ErrorPattern struct {
	ID                      string                 `json:"id"`
	Type                    string                 `json:"type"` // semantic, temporal, input_based, confidence_based
	Description             string                 `json:"description"`
	Frequency               int                    `json:"frequency"`
	Confidence              float64                `json:"confidence"`
	FirstOccurrence         time.Time              `json:"first_occurrence"`
	LastOccurrence          time.Time              `json:"last_occurrence"`
	AffectedClassifications []string               `json:"affected_classifications"`
	CommonFeatures          map[string]interface{} `json:"common_features"`
	SuggestedFixes          []string               `json:"suggested_fixes"`
	Severity                string                 `json:"severity"`
	ImpactScore             float64                `json:"impact_score"`
}

// RootCauseAnalyzer analyzes root causes of misclassifications
type RootCauseAnalyzer struct {
	config        *DetectionConfig
	logger        *zap.Logger
	knowledgeBase map[string]*CausePattern
	analysisRules []*AnalysisRule
}

// CausePattern represents a known pattern of root causes
type CausePattern struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Indicators         []string               `json:"indicators"`
	Confidence         float64                `json:"confidence"`
	RecommendedActions []string               `json:"recommended_actions"`
	Metadata           map[string]interface{} `json:"metadata"`
}

// AnalysisRule represents a rule for root cause analysis
type AnalysisRule struct {
	ID        string                                        `json:"id"`
	Name      string                                        `json:"name"`
	Condition func(*ClassificationEvent) bool               `json:"-"`
	Analysis  func(*ClassificationEvent) *RootCauseAnalysis `json:"-"`
	Priority  int                                           `json:"priority"`
	Enabled   bool                                          `json:"enabled"`
}

// RootCauseAnalysis represents the result of root cause analysis
type RootCauseAnalysis struct {
	PrimaryRoot     string                 `json:"primary_root"`
	SecondaryRoots  []string               `json:"secondary_roots"`
	Confidence      float64                `json:"confidence"`
	Evidence        []string               `json:"evidence"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// NewMisclassificationDetector creates a new misclassification detector
func NewMisclassificationDetector(config *DetectionConfig, logger *zap.Logger) *MisclassificationDetector {
	if config == nil {
		config = DefaultDetectionConfig()
	}

	detector := &MisclassificationDetector{
		config:               config,
		logger:               logger,
		patterns:             make(map[string]*ErrorPattern),
		classificationBuffer: make([]*ClassificationEvent, 0, config.BufferSize),
		errorDatabase:        make(map[string]*MisclassificationRecord),
		startTime:            time.Now(),
	}

	// Initialize root cause analyzer
	detector.rootCauseAnalyzer = NewRootCauseAnalyzer(config, logger)

	// Start background pattern analysis if enabled
	if config.EnablePatternDetection {
		go detector.runPatternAnalysis()
	}

	return detector
}

// DefaultDetectionConfig returns default configuration
func DefaultDetectionConfig() *DetectionConfig {
	return &DetectionConfig{
		EnablePatternDetection:  true,
		EnableRootCauseAnalysis: true,
		EnableRealTimeDetection: true,
		PatternAnalysisWindow:   24 * time.Hour,
		MinPatternOccurrences:   3,
		ConfidenceThreshold:     0.7,
		BufferSize:              1000,
		AnalysisInterval:        1 * time.Hour,
		EnableSemanticAnalysis:  true,
		EnableTemporalAnalysis:  true,
		EnableInputAnalysis:     true,
	}
}

// DetectMisclassification processes a classification event and detects misclassifications
func (md *MisclassificationDetector) DetectMisclassification(ctx context.Context, event *ClassificationEvent) (*MisclassificationRecord, error) {
	md.mu.Lock()
	defer md.mu.Unlock()

	// Add to buffer for pattern analysis
	md.addToBuffer(event)

	// Skip if classification is correct
	if event.IsCorrect {
		return nil, nil
	}

	// Create misclassification record
	record := &MisclassificationRecord{
		ID:                     fmt.Sprintf("misclass_%d", time.Now().UnixNano()),
		Timestamp:              event.Timestamp,
		BusinessName:           event.BusinessName,
		ExpectedClassification: event.ExpectedIndustry,
		ActualClassification:   event.ActualIndustry,
		ConfidenceScore:        event.ConfidenceScore,
		ClassificationMethod:   event.Method,
		InputData:              event.InputFeatures,
		ErrorType:              md.classifyErrorType(event),
		Severity:               md.calculateSeverity(event),
		ActionRequired:         md.requiresImmediateAction(event),
	}

	// Perform root cause analysis if enabled
	if md.config.EnableRootCauseAnalysis {
		rootCause := md.rootCauseAnalyzer.AnalyzeRootCause(event)
		if rootCause != nil {
			record.RootCause = rootCause.PrimaryRoot
		}
	}

	// Store in error database
	md.errorDatabase[record.ID] = record

	// Real-time pattern detection if enabled
	if md.config.EnableRealTimeDetection {
		md.detectRealTimePatterns(event, record)
	}

	md.logger.Info("Misclassification detected",
		zap.String("id", record.ID),
		zap.String("business", event.BusinessName),
		zap.String("expected", event.ExpectedIndustry),
		zap.String("actual", event.ActualIndustry),
		zap.Float64("confidence", event.ConfidenceScore),
		zap.String("error_type", record.ErrorType),
		zap.String("severity", record.Severity),
		zap.String("root_cause", record.RootCause))

	return record, nil
}

// addToBuffer adds an event to the analysis buffer
func (md *MisclassificationDetector) addToBuffer(event *ClassificationEvent) {
	md.classificationBuffer = append(md.classificationBuffer, event)

	// Maintain buffer size
	if len(md.classificationBuffer) > md.config.BufferSize {
		md.classificationBuffer = md.classificationBuffer[len(md.classificationBuffer)-md.config.BufferSize:]
	}
}

// classifyErrorType determines the type of classification error
func (md *MisclassificationDetector) classifyErrorType(event *ClassificationEvent) string {
	confidence := event.ConfidenceScore

	switch {
	case confidence < 0.3:
		return "very_low_confidence"
	case confidence < 0.5:
		return "low_confidence"
	case confidence < 0.7:
		return "medium_confidence"
	case confidence < 0.9:
		return "high_confidence_error"
	default:
		return "very_high_confidence_error"
	}
}

// calculateSeverity calculates the severity of a misclassification
func (md *MisclassificationDetector) calculateSeverity(event *ClassificationEvent) string {
	score := 0.0

	// High confidence errors are more severe
	if event.ConfidenceScore > 0.8 {
		score += 3.0
	} else if event.ConfidenceScore > 0.6 {
		score += 2.0
	} else {
		score += 1.0
	}

	// Frequent misclassifications of similar businesses are more severe
	if md.isFrequentError(event.ExpectedIndustry, event.ActualIndustry) {
		score += 2.0
	}

	// High-stakes industries have higher severity
	if md.isHighStakesIndustry(event.ExpectedIndustry) || md.isHighStakesIndustry(event.ActualIndustry) {
		score += 1.5
	}

	switch {
	case score >= 5.0:
		return "critical"
	case score >= 3.5:
		return "high"
	case score >= 2.0:
		return "medium"
	default:
		return "low"
	}
}

// requiresImmediateAction determines if immediate action is required
func (md *MisclassificationDetector) requiresImmediateAction(event *ClassificationEvent) bool {
	return event.ConfidenceScore > 0.9 ||
		md.isHighStakesIndustry(event.ExpectedIndustry) ||
		md.isFrequentError(event.ExpectedIndustry, event.ActualIndustry)
}

// isFrequentError checks if this type of error occurs frequently
func (md *MisclassificationDetector) isFrequentError(expected, actual string) bool {
	count := 0

	for _, record := range md.errorDatabase {
		if record.ExpectedClassification == expected && record.ActualClassification == actual {
			count++
		}
	}

	return count >= md.config.MinPatternOccurrences
}

// isHighStakesIndustry checks if an industry is considered high-stakes
func (md *MisclassificationDetector) isHighStakesIndustry(industry string) bool {
	highStakesIndustries := []string{
		"financial_services", "banking", "insurance", "healthcare",
		"pharmaceuticals", "legal_services", "government", "defense",
		"energy", "utilities", "transportation", "aerospace",
	}

	industry = strings.ToLower(industry)
	for _, highStakes := range highStakesIndustries {
		if strings.Contains(industry, highStakes) {
			return true
		}
	}

	return false
}

// detectRealTimePatterns detects patterns in real-time
func (md *MisclassificationDetector) detectRealTimePatterns(event *ClassificationEvent, record *MisclassificationRecord) {
	// Detect temporal patterns
	if md.config.EnableTemporalAnalysis {
		md.detectTemporalPatterns(event)
	}

	// Detect input-based patterns
	if md.config.EnableInputAnalysis {
		md.detectInputPatterns(event)
	}

	// Detect semantic patterns
	if md.config.EnableSemanticAnalysis {
		md.detectSemanticPatterns(event)
	}
}

// detectTemporalPatterns detects time-based error patterns
func (md *MisclassificationDetector) detectTemporalPatterns(event *ClassificationEvent) {
	hour := event.Timestamp.Hour()
	dayOfWeek := event.Timestamp.Weekday()

	// Check for peak error hours
	hourlyErrors := md.getErrorsByHour()
	if errorCount, exists := hourlyErrors[hour]; exists && errorCount >= md.config.MinPatternOccurrences {
		pattern := &ErrorPattern{
			ID:              fmt.Sprintf("temporal_hour_%d", hour),
			Type:            "temporal",
			Description:     fmt.Sprintf("Increased errors during hour %d", hour),
			Frequency:       errorCount,
			FirstOccurrence: time.Now().Add(-md.config.PatternAnalysisWindow),
			LastOccurrence:  time.Now(),
			Severity:        "medium",
			SuggestedFixes:  []string{"check_system_load", "analyze_data_quality_during_peak_hours"},
		}
		md.patterns[pattern.ID] = pattern
	}

	// Check for day-of-week patterns
	dailyErrors := md.getErrorsByDay()
	if errorCount, exists := dailyErrors[dayOfWeek]; exists && errorCount >= md.config.MinPatternOccurrences {
		pattern := &ErrorPattern{
			ID:              fmt.Sprintf("temporal_day_%s", dayOfWeek.String()),
			Type:            "temporal",
			Description:     fmt.Sprintf("Increased errors on %s", dayOfWeek.String()),
			Frequency:       errorCount,
			FirstOccurrence: time.Now().Add(-md.config.PatternAnalysisWindow),
			LastOccurrence:  time.Now(),
			Severity:        "low",
			SuggestedFixes:  []string{"analyze_weekly_data_patterns", "check_batch_processing_schedules"},
		}
		md.patterns[pattern.ID] = pattern
	}
}

// detectInputPatterns detects patterns based on input features
func (md *MisclassificationDetector) detectInputPatterns(event *ClassificationEvent) {
	// Check for common input features in errors
	featurePatterns := md.analyzeFeaturePatterns(event.InputFeatures)

	for feature, pattern := range featurePatterns {
		if pattern.Frequency >= md.config.MinPatternOccurrences {
			patternID := fmt.Sprintf("input_%s", feature)
			errorPattern := &ErrorPattern{
				ID:              patternID,
				Type:            "input_based",
				Description:     fmt.Sprintf("Errors frequently occur with feature: %s", feature),
				Frequency:       pattern.Frequency,
				FirstOccurrence: pattern.FirstSeen,
				LastOccurrence:  pattern.LastSeen,
				CommonFeatures:  map[string]interface{}{feature: pattern.CommonValue},
				Severity:        "medium",
				SuggestedFixes:  []string{"review_feature_engineering", "check_data_quality", "validate_preprocessing"},
			}
			md.patterns[patternID] = errorPattern
		}
	}
}

// detectSemanticPatterns detects patterns based on semantic similarity
func (md *MisclassificationDetector) detectSemanticPatterns(event *ClassificationEvent) {
	// Check for similar business names that are misclassified
	similarErrors := md.findSemanticallySimilarErrors(event.BusinessName, event.ExpectedIndustry, event.ActualIndustry)

	if len(similarErrors) >= md.config.MinPatternOccurrences {
		pattern := &ErrorPattern{
			ID:                      fmt.Sprintf("semantic_%s_%s", event.ExpectedIndustry, event.ActualIndustry),
			Type:                    "semantic",
			Description:             fmt.Sprintf("Semantic confusion between %s and %s", event.ExpectedIndustry, event.ActualIndustry),
			Frequency:               len(similarErrors),
			FirstOccurrence:         time.Now().Add(-md.config.PatternAnalysisWindow),
			LastOccurrence:          time.Now(),
			AffectedClassifications: []string{event.ExpectedIndustry, event.ActualIndustry},
			Severity:                "high",
			SuggestedFixes:          []string{"improve_semantic_features", "add_domain_specific_training", "review_classification_boundaries"},
		}
		md.patterns[pattern.ID] = pattern
	}
}

// Helper methods for pattern analysis

func (md *MisclassificationDetector) getErrorsByHour() map[int]int {
	hourlyErrors := make(map[int]int)

	for _, record := range md.errorDatabase {
		hour := record.Timestamp.Hour()
		hourlyErrors[hour]++
	}

	return hourlyErrors
}

func (md *MisclassificationDetector) getErrorsByDay() map[time.Weekday]int {
	dailyErrors := make(map[time.Weekday]int)

	for _, record := range md.errorDatabase {
		day := record.Timestamp.Weekday()
		dailyErrors[day]++
	}

	return dailyErrors
}

type FeaturePattern struct {
	Frequency   int
	CommonValue interface{}
	FirstSeen   time.Time
	LastSeen    time.Time
}

func (md *MisclassificationDetector) analyzeFeaturePatterns(features map[string]interface{}) map[string]*FeaturePattern {
	patterns := make(map[string]*FeaturePattern)

	for feature, value := range features {
		count := 0
		var firstSeen, lastSeen time.Time

		for _, record := range md.errorDatabase {
			if inputValue, exists := record.InputData[feature]; exists && inputValue == value {
				count++
				if firstSeen.IsZero() || record.Timestamp.Before(firstSeen) {
					firstSeen = record.Timestamp
				}
				if lastSeen.IsZero() || record.Timestamp.After(lastSeen) {
					lastSeen = record.Timestamp
				}
			}
		}

		if count > 1 {
			patterns[feature] = &FeaturePattern{
				Frequency:   count,
				CommonValue: value,
				FirstSeen:   firstSeen,
				LastSeen:    lastSeen,
			}
		}
	}

	return patterns
}

func (md *MisclassificationDetector) findSemanticallySimilarErrors(businessName, expectedIndustry, actualIndustry string) []string {
	similar := make([]string, 0)

	for _, record := range md.errorDatabase {
		if record.ExpectedClassification == expectedIndustry &&
			record.ActualClassification == actualIndustry &&
			md.calculateSemanticSimilarity(businessName, record.BusinessName) > 0.7 {
			similar = append(similar, record.BusinessName)
		}
	}

	return similar
}

func (md *MisclassificationDetector) calculateSemanticSimilarity(name1, name2 string) float64 {
	// Simple similarity calculation based on common words
	words1 := strings.Fields(strings.ToLower(name1))
	words2 := strings.Fields(strings.ToLower(name2))

	commonWords := 0
	totalWords := len(words1)
	if len(words2) > totalWords {
		totalWords = len(words2)
	}

	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 {
				commonWords++
				break
			}
		}
	}

	if totalWords == 0 {
		return 0.0
	}

	return float64(commonWords) / float64(totalWords)
}

// runPatternAnalysis runs periodic pattern analysis
func (md *MisclassificationDetector) runPatternAnalysis() {
	ticker := time.NewTicker(md.config.AnalysisInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			md.performFullPatternAnalysis()
		}
	}
}

// performFullPatternAnalysis performs comprehensive pattern analysis
func (md *MisclassificationDetector) performFullPatternAnalysis() {
	md.mu.Lock()
	defer md.mu.Unlock()

	md.logger.Info("Starting full pattern analysis",
		zap.Int("error_count", len(md.errorDatabase)),
		zap.Int("buffer_size", len(md.classificationBuffer)))

	// Analyze error frequency patterns
	md.analyzeErrorFrequencyPatterns()

	// Analyze confidence distribution patterns
	md.analyzeConfidencePatterns()

	// Analyze method-specific patterns
	md.analyzeMethodPatterns()

	// Clean up old patterns
	md.cleanupOldPatterns()

	md.logger.Info("Pattern analysis completed",
		zap.Int("patterns_detected", len(md.patterns)))
}

func (md *MisclassificationDetector) analyzeErrorFrequencyPatterns() {
	// Find most common error types
	errorTypeFrequency := make(map[string]int)

	for _, record := range md.errorDatabase {
		errorKey := fmt.Sprintf("%s->%s", record.ExpectedClassification, record.ActualClassification)
		errorTypeFrequency[errorKey]++
	}

	// Create patterns for high-frequency errors
	for errorType, frequency := range errorTypeFrequency {
		if frequency >= md.config.MinPatternOccurrences {
			pattern := &ErrorPattern{
				ID:              fmt.Sprintf("frequency_%s", errorType),
				Type:            "frequency",
				Description:     fmt.Sprintf("High frequency error: %s (occurs %d times)", errorType, frequency),
				Frequency:       frequency,
				FirstOccurrence: time.Now().Add(-md.config.PatternAnalysisWindow),
				LastOccurrence:  time.Now(),
				Severity:        md.calculatePatternSeverity(frequency),
				ImpactScore:     float64(frequency) / float64(len(md.errorDatabase)),
				SuggestedFixes:  []string{"review_training_data", "improve_feature_engineering", "add_classification_rules"},
			}
			md.patterns[pattern.ID] = pattern
		}
	}
}

func (md *MisclassificationDetector) analyzeConfidencePatterns() {
	// Analyze confidence distribution in errors
	confidenceRanges := map[string][2]float64{
		"very_low":  {0.0, 0.3},
		"low":       {0.3, 0.5},
		"medium":    {0.5, 0.7},
		"high":      {0.7, 0.9},
		"very_high": {0.9, 1.0},
	}

	for rangeName, bounds := range confidenceRanges {
		count := 0
		for _, record := range md.errorDatabase {
			if record.ConfidenceScore >= bounds[0] && record.ConfidenceScore < bounds[1] {
				count++
			}
		}

		if count >= md.config.MinPatternOccurrences {
			pattern := &ErrorPattern{
				ID:              fmt.Sprintf("confidence_%s", rangeName),
				Type:            "confidence_based",
				Description:     fmt.Sprintf("High error rate in %s confidence range (%d errors)", rangeName, count),
				Frequency:       count,
				FirstOccurrence: time.Now().Add(-md.config.PatternAnalysisWindow),
				LastOccurrence:  time.Now(),
				Severity:        md.calculateConfidencePatternSeverity(rangeName, count),
				SuggestedFixes:  md.getConfidenceBasedFixes(rangeName),
			}
			md.patterns[pattern.ID] = pattern
		}
	}
}

func (md *MisclassificationDetector) analyzeMethodPatterns() {
	// Analyze errors by classification method
	methodErrors := make(map[string]int)

	for _, record := range md.errorDatabase {
		methodErrors[record.ClassificationMethod]++
	}

	for method, count := range methodErrors {
		if count >= md.config.MinPatternOccurrences {
			pattern := &ErrorPattern{
				ID:              fmt.Sprintf("method_%s", method),
				Type:            "method_based",
				Description:     fmt.Sprintf("High error rate for method %s (%d errors)", method, count),
				Frequency:       count,
				FirstOccurrence: time.Now().Add(-md.config.PatternAnalysisWindow),
				LastOccurrence:  time.Now(),
				Severity:        md.calculateMethodPatternSeverity(method, count),
				SuggestedFixes:  md.getMethodBasedFixes(method),
			}
			md.patterns[pattern.ID] = pattern
		}
	}
}

func (md *MisclassificationDetector) calculatePatternSeverity(frequency int) string {
	totalErrors := len(md.errorDatabase)
	ratio := float64(frequency) / float64(totalErrors)

	switch {
	case ratio > 0.3:
		return "critical"
	case ratio > 0.2:
		return "high"
	case ratio > 0.1:
		return "medium"
	default:
		return "low"
	}
}

func (md *MisclassificationDetector) calculateConfidencePatternSeverity(rangeName string, count int) string {
	// High confidence errors are more severe
	if rangeName == "very_high" || rangeName == "high" {
		return "high"
	}

	// Many low confidence errors indicate system uncertainty
	if (rangeName == "very_low" || rangeName == "low") && count > md.config.MinPatternOccurrences*2 {
		return "medium"
	}

	return "low"
}

func (md *MisclassificationDetector) calculateMethodPatternSeverity(method string, count int) string {
	totalErrors := len(md.errorDatabase)
	ratio := float64(count) / float64(totalErrors)

	// ML methods should have lower error rates
	if strings.Contains(strings.ToLower(method), "ml") || strings.Contains(strings.ToLower(method), "neural") {
		if ratio > 0.2 {
			return "high"
		}
	}

	return md.calculatePatternSeverity(count)
}

func (md *MisclassificationDetector) getConfidenceBasedFixes(rangeName string) []string {
	switch rangeName {
	case "very_high":
		return []string{"review_training_data", "check_model_overfitting", "validate_ground_truth"}
	case "high":
		return []string{"improve_feature_quality", "add_validation_rules", "review_edge_cases"}
	case "medium":
		return []string{"enhance_feature_engineering", "improve_model_complexity", "add_ensemble_methods"}
	case "low":
		return []string{"improve_data_quality", "add_more_training_data", "enhance_preprocessing"}
	case "very_low":
		return []string{"check_input_validation", "improve_data_cleaning", "review_feature_selection"}
	default:
		return []string{"general_model_improvement"}
	}
}

func (md *MisclassificationDetector) getMethodBasedFixes(method string) []string {
	methodLower := strings.ToLower(method)

	if strings.Contains(methodLower, "keyword") {
		return []string{"expand_keyword_dictionary", "improve_text_preprocessing", "add_semantic_analysis"}
	} else if strings.Contains(methodLower, "ml") || strings.Contains(methodLower, "neural") {
		return []string{"retrain_model", "improve_feature_engineering", "add_more_training_data", "tune_hyperparameters"}
	} else if strings.Contains(methodLower, "rule") {
		return []string{"review_classification_rules", "add_exception_handling", "improve_rule_ordering"}
	}

	return []string{"method_specific_optimization", "cross_validation", "performance_tuning"}
}

func (md *MisclassificationDetector) cleanupOldPatterns() {
	cutoffTime := time.Now().Add(-md.config.PatternAnalysisWindow)

	for id, pattern := range md.patterns {
		if pattern.LastOccurrence.Before(cutoffTime) {
			delete(md.patterns, id)
		}
	}
}

// Public methods for accessing detector data

// GetDetectedPatterns returns all detected error patterns
func (md *MisclassificationDetector) GetDetectedPatterns() map[string]*ErrorPattern {
	md.mu.RLock()
	defer md.mu.RUnlock()

	result := make(map[string]*ErrorPattern)
	for k, v := range md.patterns {
		result[k] = v
	}
	return result
}

// GetMisclassificationsByTimeRange returns misclassifications within a time range
func (md *MisclassificationDetector) GetMisclassificationsByTimeRange(start, end time.Time) []*MisclassificationRecord {
	md.mu.RLock()
	defer md.mu.RUnlock()

	result := make([]*MisclassificationRecord, 0)
	for _, record := range md.errorDatabase {
		if record.Timestamp.After(start) && record.Timestamp.Before(end) {
			result = append(result, record)
		}
	}

	// Sort by timestamp
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp.After(result[j].Timestamp)
	})

	return result
}

// GetPatternsBySeverity returns patterns filtered by severity
func (md *MisclassificationDetector) GetPatternsBySeverity(severity string) []*ErrorPattern {
	md.mu.RLock()
	defer md.mu.RUnlock()

	result := make([]*ErrorPattern, 0)
	for _, pattern := range md.patterns {
		if pattern.Severity == severity {
			result = append(result, pattern)
		}
	}

	return result
}

// GetErrorStatistics returns comprehensive error statistics
func (md *MisclassificationDetector) GetErrorStatistics() map[string]interface{} {
	md.mu.RLock()
	defer md.mu.RUnlock()

	stats := make(map[string]interface{})

	// Basic counts
	stats["total_errors"] = len(md.errorDatabase)
	stats["total_patterns"] = len(md.patterns)

	// Error distribution by type
	errorTypes := make(map[string]int)
	for _, record := range md.errorDatabase {
		errorTypes[record.ErrorType]++
	}
	stats["error_distribution"] = errorTypes

	// Severity distribution
	severityDist := make(map[string]int)
	for _, record := range md.errorDatabase {
		severityDist[record.Severity]++
	}
	stats["severity_distribution"] = severityDist

	// Pattern type distribution
	patternTypes := make(map[string]int)
	for _, pattern := range md.patterns {
		patternTypes[pattern.Type]++
	}
	stats["pattern_distribution"] = patternTypes

	// Time range
	if len(md.errorDatabase) > 0 {
		var earliest, latest time.Time
		for _, record := range md.errorDatabase {
			if earliest.IsZero() || record.Timestamp.Before(earliest) {
				earliest = record.Timestamp
			}
			if latest.IsZero() || record.Timestamp.After(latest) {
				latest = record.Timestamp
			}
		}
		stats["time_range"] = map[string]interface{}{
			"earliest": earliest,
			"latest":   latest,
			"span":     latest.Sub(earliest).String(),
		}
	}

	return stats
}

// NewRootCauseAnalyzer creates a new root cause analyzer
func NewRootCauseAnalyzer(config *DetectionConfig, logger *zap.Logger) *RootCauseAnalyzer {
	analyzer := &RootCauseAnalyzer{
		config:        config,
		logger:        logger,
		knowledgeBase: make(map[string]*CausePattern),
		analysisRules: make([]*AnalysisRule, 0),
	}

	// Initialize knowledge base and rules
	analyzer.initializeKnowledgeBase()
	analyzer.initializeAnalysisRules()

	return analyzer
}

// AnalyzeRootCause analyzes the root cause of a misclassification
func (rca *RootCauseAnalyzer) AnalyzeRootCause(event *ClassificationEvent) *RootCauseAnalysis {
	analysis := &RootCauseAnalysis{
		SecondaryRoots:  make([]string, 0),
		Evidence:        make([]string, 0),
		Recommendations: make([]string, 0),
		Metadata:        make(map[string]interface{}),
	}

	// Apply analysis rules
	for _, rule := range rca.analysisRules {
		if rule.Enabled && rule.Condition(event) {
			ruleAnalysis := rule.Analysis(event)
			if ruleAnalysis != nil {
				if analysis.PrimaryRoot == "" || ruleAnalysis.Confidence > analysis.Confidence {
					if analysis.PrimaryRoot != "" {
						analysis.SecondaryRoots = append(analysis.SecondaryRoots, analysis.PrimaryRoot)
					}
					analysis.PrimaryRoot = ruleAnalysis.PrimaryRoot
					analysis.Confidence = ruleAnalysis.Confidence
				} else {
					analysis.SecondaryRoots = append(analysis.SecondaryRoots, ruleAnalysis.PrimaryRoot)
				}

				analysis.Evidence = append(analysis.Evidence, ruleAnalysis.Evidence...)
				analysis.Recommendations = append(analysis.Recommendations, ruleAnalysis.Recommendations...)
			}
		}
	}

	// Deduplicate recommendations
	analysis.Recommendations = rca.deduplicateStrings(analysis.Recommendations)
	analysis.Evidence = rca.deduplicateStrings(analysis.Evidence)
	analysis.SecondaryRoots = rca.deduplicateStrings(analysis.SecondaryRoots)

	return analysis
}

func (rca *RootCauseAnalyzer) initializeKnowledgeBase() {
	// Initialize with common root cause patterns
	patterns := []*CausePattern{
		{
			ID:                 "insufficient_training_data",
			Name:               "Insufficient Training Data",
			Indicators:         []string{"low_confidence", "new_business_type", "rare_industry"},
			Confidence:         0.8,
			RecommendedActions: []string{"collect_more_training_data", "use_data_augmentation", "apply_transfer_learning"},
		},
		{
			ID:                 "feature_quality_issues",
			Name:               "Feature Quality Issues",
			Indicators:         []string{"missing_features", "low_quality_input", "inconsistent_formatting"},
			Confidence:         0.75,
			RecommendedActions: []string{"improve_feature_extraction", "enhance_data_cleaning", "add_validation_rules"},
		},
		{
			ID:                 "model_overfitting",
			Name:               "Model Overfitting",
			Indicators:         []string{"high_confidence_error", "perfect_training_accuracy", "poor_generalization"},
			Confidence:         0.85,
			RecommendedActions: []string{"add_regularization", "use_cross_validation", "reduce_model_complexity"},
		},
		{
			ID:                 "label_noise",
			Name:               "Label Noise or Inconsistency",
			Indicators:         []string{"conflicting_labels", "manual_labeling_errors", "subjective_classification"},
			Confidence:         0.7,
			RecommendedActions: []string{"review_labeling_guidelines", "implement_label_validation", "use_multiple_annotators"},
		},
	}

	for _, pattern := range patterns {
		rca.knowledgeBase[pattern.ID] = pattern
	}
}

func (rca *RootCauseAnalyzer) initializeAnalysisRules() {
	rules := []*AnalysisRule{
		{
			ID:   "high_confidence_error_rule",
			Name: "High Confidence Error Analysis",
			Condition: func(event *ClassificationEvent) bool {
				return event.ConfidenceScore > 0.9 && !event.IsCorrect
			},
			Analysis: func(event *ClassificationEvent) *RootCauseAnalysis {
				return &RootCauseAnalysis{
					PrimaryRoot:     "model_overfitting",
					Confidence:      0.85,
					Evidence:        []string{fmt.Sprintf("High confidence (%.2f) but incorrect classification", event.ConfidenceScore)},
					Recommendations: []string{"review_training_data", "check_model_complexity", "validate_ground_truth"},
				}
			},
			Priority: 1,
			Enabled:  true,
		},
		{
			ID:   "low_confidence_error_rule",
			Name: "Low Confidence Error Analysis",
			Condition: func(event *ClassificationEvent) bool {
				return event.ConfidenceScore < 0.5 && !event.IsCorrect
			},
			Analysis: func(event *ClassificationEvent) *RootCauseAnalysis {
				return &RootCauseAnalysis{
					PrimaryRoot:     "insufficient_training_data",
					Confidence:      0.7,
					Evidence:        []string{fmt.Sprintf("Low confidence (%.2f) indicates model uncertainty", event.ConfidenceScore)},
					Recommendations: []string{"collect_more_training_data", "improve_feature_quality", "enhance_preprocessing"},
				}
			},
			Priority: 2,
			Enabled:  true,
		},
		{
			ID:   "feature_importance_rule",
			Name: "Feature Importance Analysis",
			Condition: func(event *ClassificationEvent) bool {
				return len(event.FeatureImportance) > 0
			},
			Analysis: func(event *ClassificationEvent) *RootCauseAnalysis {
				// Analyze feature importance for anomalies
				var maxImportance float64
				var dominantFeature string

				for feature, importance := range event.FeatureImportance {
					if importance > maxImportance {
						maxImportance = importance
						dominantFeature = feature
					}
				}

				if maxImportance > 0.8 {
					return &RootCauseAnalysis{
						PrimaryRoot:     "feature_quality_issues",
						Confidence:      0.75,
						Evidence:        []string{fmt.Sprintf("Dominant feature '%s' with importance %.2f may indicate feature bias", dominantFeature, maxImportance)},
						Recommendations: []string{"review_feature_engineering", "check_for_feature_leakage", "balance_feature_importance"},
					}
				}

				return nil
			},
			Priority: 3,
			Enabled:  true,
		},
	}

	rca.analysisRules = rules
}

func (rca *RootCauseAnalyzer) deduplicateStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, str := range slice {
		if !seen[str] {
			seen[str] = true
			result = append(result, str)
		}
	}

	return result
}
