package error_monitoring

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ErrorAnalyzer provides comprehensive error analysis and root cause identification
type ErrorAnalyzer struct {
	config        *ErrorAnalysisConfig
	logger        *zap.Logger
	mu            sync.RWMutex
	errorPatterns map[string]*ErrorPattern
	rootCauses    map[string]*RootCauseAnalysis
	correlations  map[string]*ErrorCorrelation
	analysisCache map[string]*ErrorAnalysisResult
}

// ErrorAnalysisConfig contains configuration for error analysis
type ErrorAnalysisConfig struct {
	AnalysisWindow            time.Duration `json:"analysis_window"`             // 1 hour
	PatternDetectionThreshold int           `json:"pattern_detection_threshold"` // 3 occurrences
	CorrelationThreshold      float64       `json:"correlation_threshold"`       // 0.7
	RootCauseConfidence       float64       `json:"root_cause_confidence"`       // 0.8
	MaxAnalysisDepth          int           `json:"max_analysis_depth"`          // 5 levels
	EnableMachineLearning     bool          `json:"enable_machine_learning"`
	EnableTemporalAnalysis    bool          `json:"enable_temporal_analysis"`
	EnableDependencyAnalysis  bool          `json:"enable_dependency_analysis"`
	CacheAnalysisResults      bool          `json:"cache_analysis_results"`
	CacheTTL                  time.Duration `json:"cache_ttl"` // 30 minutes
}

// ErrorPattern represents a recurring error pattern
type ErrorPattern struct {
	ID          string                 `json:"id"`
	PatternType string                 `json:"pattern_type"` // "sequence", "temporal", "dependency", "correlation"
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	ErrorTypes  []string               `json:"error_types"`
	Processes   []string               `json:"processes"`
	Frequency   int                    `json:"frequency"`
	Confidence  float64                `json:"confidence"`
	FirstSeen   time.Time              `json:"first_seen"`
	LastSeen    time.Time              `json:"last_seen"`
	RootCause   *RootCauseAnalysis     `json:"root_cause"`
	Mitigation  *MitigationStrategy    `json:"mitigation"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RootCauseAnalysis represents a root cause analysis result
type RootCauseAnalysis struct {
	ID                  string                 `json:"id"`
	Category            string                 `json:"category"` // "infrastructure", "application", "data", "external", "configuration"
	RootCause           string                 `json:"root_cause"`
	Description         string                 `json:"description"`
	Confidence          float64                `json:"confidence"`
	Impact              string                 `json:"impact"` // "high", "medium", "low"
	AffectedProcesses   []string               `json:"affected_processes"`
	Evidence            []EvidenceItem         `json:"evidence"`
	ContributingFactors []string               `json:"contributing_factors"`
	Recommendations     []string               `json:"recommendations"`
	Timeline            []TimelineEvent        `json:"timeline"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// EvidenceItem represents evidence supporting a root cause analysis
type EvidenceItem struct {
	Type        string                 `json:"type"` // "error_log", "metric", "correlation", "temporal"
	Description string                 `json:"description"`
	Value       interface{}            `json:"value"`
	Confidence  float64                `json:"confidence"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// TimelineEvent represents an event in the error timeline
type TimelineEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	Severity    string                 `json:"severity"`
	Process     string                 `json:"process"`
	ErrorType   string                 `json:"error_type"`
	Context     map[string]interface{} `json:"context"`
}

// MitigationStrategy represents a strategy to mitigate errors
type MitigationStrategy struct {
	ID             string                 `json:"id"`
	StrategyType   string                 `json:"strategy_type"` // "preventive", "reactive", "adaptive"
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Implementation []string               `json:"implementation"`
	ExpectedImpact float64                `json:"expected_impact"`
	Effort         string                 `json:"effort"`   // "low", "medium", "high"
	Priority       string                 `json:"priority"` // "low", "medium", "high", "critical"
	Status         string                 `json:"status"`   // "proposed", "implemented", "testing", "active"
	Metrics        []string               `json:"metrics"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// ErrorCorrelation represents correlation between different error types
type ErrorCorrelation struct {
	ID              string                 `json:"id"`
	PrimaryError    string                 `json:"primary_error"`
	SecondaryError  string                 `json:"secondary_error"`
	CorrelationType string                 `json:"correlation_type"` // "causal", "temporal", "spatial"
	Strength        float64                `json:"strength"`         // 0.0 - 1.0
	Confidence      float64                `json:"confidence"`
	Direction       string                 `json:"direction"` // "forward", "backward", "bidirectional"
	Evidence        []EvidenceItem         `json:"evidence"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ErrorAnalysisResult represents the result of error analysis
type ErrorAnalysisResult struct {
	ID              string                 `json:"id"`
	AnalysisTime    time.Time              `json:"analysis_time"`
	ProcessName     string                 `json:"process_name"`
	TimeRange       TimeRange              `json:"time_range"`
	ErrorPatterns   []*ErrorPattern        `json:"error_patterns"`
	RootCauses      []*RootCauseAnalysis   `json:"root_causes"`
	Correlations    []*ErrorCorrelation    `json:"correlations"`
	Recommendations []string               `json:"recommendations"`
	RiskAssessment  *RiskAssessment        `json:"risk_assessment"`
	Trends          *ErrorTrendAnalysis    `json:"trends"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// TimeRange represents a time range for analysis
type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// RiskAssessment represents risk assessment for errors
type RiskAssessment struct {
	OverallRisk        string          `json:"overall_risk"` // "low", "medium", "high", "critical"
	RiskScore          float64         `json:"risk_score"`   // 0.0 - 1.0
	RiskFactors        []RiskFactor    `json:"risk_factors"`
	ImpactAnalysis     *ImpactAnalysis `json:"impact_analysis"`
	MitigationPriority []string        `json:"mitigation_priority"`
}

// RiskFactor represents a risk factor in error analysis
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Description string  `json:"description"`
	RiskLevel   string  `json:"risk_level"`
	Probability float64 `json:"probability"`
	Impact      string  `json:"impact"`
	Mitigation  string  `json:"mitigation"`
}

// ImpactAnalysis represents impact analysis of errors
type ImpactAnalysis struct {
	UserImpact       string                 `json:"user_impact"`
	BusinessImpact   string                 `json:"business_impact"`
	SystemImpact     string                 `json:"system_impact"`
	FinancialImpact  string                 `json:"financial_impact"`
	ReputationImpact string                 `json:"reputation_impact"`
	Details          map[string]interface{} `json:"details"`
}

// ErrorTrendAnalysis represents trend analysis for errors
type ErrorTrendAnalysis struct {
	OverallTrend     string            `json:"overall_trend"` // "improving", "stable", "degrading"
	TrendConfidence  float64           `json:"trend_confidence"`
	SeasonalPatterns []SeasonalPattern `json:"seasonal_patterns"`
	CyclicalPatterns []CyclicalPattern `json:"cyclical_patterns"`
	Predictions      []Prediction      `json:"predictions"`
}

// SeasonalPattern represents seasonal patterns in errors
type SeasonalPattern struct {
	PatternType string      `json:"pattern_type"` // "daily", "weekly", "monthly"
	Description string      `json:"description"`
	Confidence  float64     `json:"confidence"`
	PeakTimes   []time.Time `json:"peak_times"`
	LowTimes    []time.Time `json:"low_times"`
}

// CyclicalPattern represents cyclical patterns in errors
type CyclicalPattern struct {
	CycleLength time.Duration `json:"cycle_length"`
	Description string        `json:"description"`
	Confidence  float64       `json:"confidence"`
	PeakPeriods []TimeRange   `json:"peak_periods"`
	LowPeriods  []TimeRange   `json:"low_periods"`
}

// Prediction represents error predictions
type Prediction struct {
	PredictionType string        `json:"prediction_type"` // "error_rate", "pattern_occurrence", "root_cause"
	Value          interface{}   `json:"value"`
	Confidence     float64       `json:"confidence"`
	Timeframe      time.Duration `json:"timeframe"`
	Description    string        `json:"description"`
}

// NewErrorAnalyzer creates a new error analyzer
func NewErrorAnalyzer(config *ErrorAnalysisConfig, logger *zap.Logger) *ErrorAnalyzer {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = getDefaultErrorAnalysisConfig()
	}

	return &ErrorAnalyzer{
		config:        config,
		logger:        logger,
		errorPatterns: make(map[string]*ErrorPattern),
		rootCauses:    make(map[string]*RootCauseAnalysis),
		correlations:  make(map[string]*ErrorCorrelation),
		analysisCache: make(map[string]*ErrorAnalysisResult),
	}
}

// getDefaultErrorAnalysisConfig returns default configuration
func getDefaultErrorAnalysisConfig() *ErrorAnalysisConfig {
	return &ErrorAnalysisConfig{
		AnalysisWindow:            1 * time.Hour,
		PatternDetectionThreshold: 3,
		CorrelationThreshold:      0.7,
		RootCauseConfidence:       0.8,
		MaxAnalysisDepth:          5,
		EnableMachineLearning:     false,
		EnableTemporalAnalysis:    true,
		EnableDependencyAnalysis:  true,
		CacheAnalysisResults:      true,
		CacheTTL:                  30 * time.Minute,
	}
}

// AnalyzeErrors performs comprehensive error analysis for a process
func (ea *ErrorAnalyzer) AnalyzeErrors(ctx context.Context, processName string, errors []ErrorEntry, timeRange TimeRange) (*ErrorAnalysisResult, error) {
	ea.mu.Lock()
	defer ea.mu.Unlock()

	// Check cache first
	cacheKey := fmt.Sprintf("%s_%d_%d", processName, timeRange.Start.Unix(), timeRange.End.Unix())
	if ea.config.CacheAnalysisResults {
		if cached, exists := ea.analysisCache[cacheKey]; exists {
			if time.Since(cached.AnalysisTime) < ea.config.CacheTTL {
				ea.logger.Debug("Returning cached analysis result", zap.String("process", processName))
				return cached, nil
			}
		}
	}

	ea.logger.Info("Starting error analysis",
		zap.String("process", processName),
		zap.Int("error_count", len(errors)),
		zap.Time("start", timeRange.Start),
		zap.Time("end", timeRange.End))

	// Filter errors for the specified time range
	filteredErrors := ea.filterErrorsByTimeRange(errors, timeRange)

	// Perform pattern detection
	patterns := ea.detectErrorPatterns(filteredErrors, processName)

	// Perform root cause analysis
	rootCauses := ea.analyzeRootCauses(filteredErrors, patterns, processName)

	// Perform correlation analysis
	correlations := ea.analyzeErrorCorrelations(filteredErrors, processName)

	// Generate recommendations
	recommendations := ea.generateRecommendations(patterns, rootCauses, correlations)

	// Perform risk assessment
	riskAssessment := ea.performRiskAssessment(filteredErrors, patterns, rootCauses)

	// Perform trend analysis
	trends := ea.performErrorTrendAnalysis(filteredErrors, timeRange)

	// Create analysis result
	result := &ErrorAnalysisResult{
		ID:              generateAnalysisID(),
		AnalysisTime:    time.Now(),
		ProcessName:     processName,
		TimeRange:       timeRange,
		ErrorPatterns:   patterns,
		RootCauses:      rootCauses,
		Correlations:    correlations,
		Recommendations: recommendations,
		RiskAssessment:  riskAssessment,
		Trends:          trends,
		Metadata: map[string]interface{}{
			"total_errors":      len(filteredErrors),
			"analysis_duration": time.Since(time.Now()),
			"cache_key":         cacheKey,
		},
	}

	// Cache the result
	if ea.config.CacheAnalysisResults {
		ea.analysisCache[cacheKey] = result
	}

	ea.logger.Info("Error analysis completed",
		zap.String("process", processName),
		zap.Int("patterns_found", len(patterns)),
		zap.Int("root_causes_found", len(rootCauses)),
		zap.Int("correlations_found", len(correlations)))

	return result, nil
}

// filterErrorsByTimeRange filters errors by time range
func (ea *ErrorAnalyzer) filterErrorsByTimeRange(errors []ErrorEntry, timeRange TimeRange) []ErrorEntry {
	var filtered []ErrorEntry

	for _, err := range errors {
		if err.Timestamp.After(timeRange.Start) && err.Timestamp.Before(timeRange.End) {
			filtered = append(filtered, err)
		}
	}

	return filtered
}

// detectErrorPatterns detects patterns in error occurrences
func (ea *ErrorAnalyzer) detectErrorPatterns(errors []ErrorEntry, processName string) []*ErrorPattern {
	var patterns []*ErrorPattern

	// Group errors by type
	errorGroups := ea.groupErrorsByType(errors)

	// Detect sequence patterns
	sequencePatterns := ea.detectSequencePatterns(errors, processName)
	patterns = append(patterns, sequencePatterns...)

	// Detect temporal patterns
	if ea.config.EnableTemporalAnalysis {
		temporalPatterns := ea.detectTemporalPatterns(errors, processName)
		patterns = append(patterns, temporalPatterns...)
	}

	// Detect dependency patterns
	if ea.config.EnableDependencyAnalysis {
		dependencyPatterns := ea.detectDependencyPatterns(errors, processName)
		patterns = append(patterns, dependencyPatterns...)
	}

	// Detect correlation patterns
	correlationPatterns := ea.detectCorrelationPatterns(errorGroups, processName)
	patterns = append(patterns, correlationPatterns...)

	return patterns
}

// groupErrorsByType groups errors by their type
func (ea *ErrorAnalyzer) groupErrorsByType(errors []ErrorEntry) map[string][]ErrorEntry {
	groups := make(map[string][]ErrorEntry)

	for _, err := range errors {
		groups[err.ErrorType] = append(groups[err.ErrorType], err)
	}

	return groups
}

// detectSequencePatterns detects sequential error patterns
func (ea *ErrorAnalyzer) detectSequencePatterns(errors []ErrorEntry, processName string) []*ErrorPattern {
	var patterns []*ErrorPattern

	if len(errors) < ea.config.PatternDetectionThreshold {
		return patterns
	}

	// Sort errors by timestamp
	sort.Slice(errors, func(i, j int) bool {
		return errors[i].Timestamp.Before(errors[j].Timestamp)
	})

	// Look for repeating sequences
	sequences := ea.findRepeatingSequences(errors)

	for _, sequence := range sequences {
		if len(sequence) >= ea.config.PatternDetectionThreshold {
			pattern := &ErrorPattern{
				ID:          generatePatternID(),
				PatternType: "sequence",
				Name:        fmt.Sprintf("Sequential Error Pattern - %s", sequence[0].ErrorType),
				Description: fmt.Sprintf("Repeating sequence of %d errors starting with %s", len(sequence), sequence[0].ErrorType),
				ErrorTypes:  ea.extractErrorTypes(sequence),
				Processes:   []string{processName},
				Frequency:   ea.calculateFrequency(sequence, errors),
				Confidence:  ea.calculatePatternConfidence(sequence, errors),
				FirstSeen:   sequence[0].Timestamp,
				LastSeen:    sequence[len(sequence)-1].Timestamp,
				Metadata: map[string]interface{}{
					"sequence_length": len(sequence),
					"time_span":       sequence[len(sequence)-1].Timestamp.Sub(sequence[0].Timestamp),
				},
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// detectTemporalPatterns detects temporal patterns in errors
func (ea *ErrorAnalyzer) detectTemporalPatterns(errors []ErrorEntry, processName string) []*ErrorPattern {
	var patterns []*ErrorPattern

	if len(errors) < ea.config.PatternDetectionThreshold {
		return patterns
	}

	// Analyze time-based patterns
	hourlyDistribution := ea.analyzeHourlyDistribution(errors)
	dailyDistribution := ea.analyzeDailyDistribution(errors)

	// Detect peak hours
	peakHours := ea.detectPeakHours(hourlyDistribution)
	if len(peakHours) > 0 {
		pattern := &ErrorPattern{
			ID:          generatePatternID(),
			PatternType: "temporal",
			Name:        "Peak Hour Error Pattern",
			Description: fmt.Sprintf("Errors occur more frequently during hours: %v", peakHours),
			Processes:   []string{processName},
			Frequency:   ea.calculateTemporalFrequency(errors, peakHours),
			Confidence:  ea.calculateTemporalConfidence(hourlyDistribution, peakHours),
			FirstSeen:   errors[0].Timestamp,
			LastSeen:    errors[len(errors)-1].Timestamp,
			Metadata: map[string]interface{}{
				"peak_hours":          peakHours,
				"hourly_distribution": hourlyDistribution,
				"daily_distribution":  dailyDistribution,
			},
		}
		patterns = append(patterns, pattern)
	}

	return patterns
}

// detectDependencyPatterns detects dependency-based error patterns
func (ea *ErrorAnalyzer) detectDependencyPatterns(errors []ErrorEntry, processName string) []*ErrorPattern {
	var patterns []*ErrorPattern

	// Look for errors that occur together or in specific order
	dependencies := ea.findErrorDependencies(errors)

	for _, dep := range dependencies {
		if dep.Strength >= ea.config.CorrelationThreshold {
			pattern := &ErrorPattern{
				ID:          generatePatternID(),
				PatternType: "dependency",
				Name:        fmt.Sprintf("Dependency Pattern - %s -> %s", dep.PrimaryError, dep.SecondaryError),
				Description: fmt.Sprintf("Error %s frequently leads to error %s", dep.PrimaryError, dep.SecondaryError),
				ErrorTypes:  []string{dep.PrimaryError, dep.SecondaryError},
				Processes:   []string{processName},
				Frequency:   int(dep.Strength * 100),
				Confidence:  dep.Confidence,
				FirstSeen:   errors[0].Timestamp,
				LastSeen:    errors[len(errors)-1].Timestamp,
				Metadata: map[string]interface{}{
					"dependency_strength":  dep.Strength,
					"dependency_direction": dep.Direction,
				},
			}
			patterns = append(patterns, pattern)
		}
	}

	return patterns
}

// detectCorrelationPatterns detects correlation patterns between error types
func (ea *ErrorAnalyzer) detectCorrelationPatterns(errorGroups map[string][]ErrorEntry, processName string) []*ErrorPattern {
	var patterns []*ErrorPattern

	errorTypes := make([]string, 0, len(errorGroups))
	for errorType := range errorGroups {
		errorTypes = append(errorTypes, errorType)
	}

	// Check correlations between all pairs of error types
	for i := 0; i < len(errorTypes); i++ {
		for j := i + 1; j < len(errorTypes); j++ {
			correlation := ea.calculateErrorCorrelation(errorGroups[errorTypes[i]], errorGroups[errorTypes[j]])

			if correlation.Strength >= ea.config.CorrelationThreshold {
				pattern := &ErrorPattern{
					ID:          generatePatternID(),
					PatternType: "correlation",
					Name:        fmt.Sprintf("Correlation Pattern - %s & %s", errorTypes[i], errorTypes[j]),
					Description: fmt.Sprintf("Strong correlation between %s and %s errors", errorTypes[i], errorTypes[j]),
					ErrorTypes:  []string{errorTypes[i], errorTypes[j]},
					Processes:   []string{processName},
					Frequency:   int(correlation.Strength * 100),
					Confidence:  correlation.Confidence,
					FirstSeen:   ea.getEarliestTimestamp(errorGroups[errorTypes[i]], errorGroups[errorTypes[j]]),
					LastSeen:    ea.getLatestTimestamp(errorGroups[errorTypes[i]], errorGroups[errorTypes[j]]),
					Metadata: map[string]interface{}{
						"correlation_strength": correlation.Strength,
						"correlation_type":     correlation.CorrelationType,
					},
				}
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns
}

// Helper methods continue in the next part...
