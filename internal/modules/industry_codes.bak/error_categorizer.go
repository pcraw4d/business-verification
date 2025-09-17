package industry_codes

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"time"

	"go.uber.org/zap"
)

// ErrorCategorizer provides comprehensive error categorization and prioritization
type ErrorCategorizer struct {
	logger    *zap.Logger
	config    *CategorizationConfig
	patterns  *ErrorPatterns
	analytics *ErrorAnalytics
}

// CategorizationConfig defines configuration for error categorization
type CategorizationConfig struct {
	EnableAnalytics      bool                      `json:"enable_analytics"`
	EnableTrends         bool                      `json:"enable_trends"`
	EnablePrioritization bool                      `json:"enable_prioritization"`
	SeverityThresholds   map[ErrorSeverity]float64 `json:"severity_thresholds"`
	CategoryWeights      map[ErrorCategory]float64 `json:"category_weights"`
	MaxAnalyticsHistory  int                       `json:"max_analytics_history"`
	TrendAnalysisWindow  time.Duration             `json:"trend_analysis_window"`
	PriorityAdjustment   PriorityAdjustmentConfig  `json:"priority_adjustment"`
}

// PriorityAdjustmentConfig defines how priorities are adjusted based on various factors
type PriorityAdjustmentConfig struct {
	FrequencyMultiplier   float64 `json:"frequency_multiplier"`
	RecencyMultiplier     float64 `json:"recency_multiplier"`
	ImpactMultiplier      float64 `json:"impact_multiplier"`
	SystemHealthFactor    float64 `json:"system_health_factor"`
	BusinessCriticalBoost float64 `json:"business_critical_boost"`
}

// ErrorCategory represents different categories of errors
type ErrorCategory string

const (
	CategoryNetwork        ErrorCategory = "network"
	CategoryDatabase       ErrorCategory = "database"
	CategoryValidation     ErrorCategory = "validation"
	CategoryAuthentication ErrorCategory = "authentication"
	CategoryAuthorization  ErrorCategory = "authorization"
	CategoryBusiness       ErrorCategory = "business"
	CategorySystem         ErrorCategory = "system"
	CategoryExternal       ErrorCategory = "external"
	CategoryConfiguration  ErrorCategory = "configuration"
	CategorySecurity       ErrorCategory = "security"
	CategoryPerformance    ErrorCategory = "performance"
	CategoryUnknown        ErrorCategory = "unknown"
)

// ErrorSeverity represents the severity level of errors
type ErrorSeverity string

const (
	SeverityCritical ErrorSeverity = "critical"
	SeverityHigh     ErrorSeverity = "high"
	SeverityMedium   ErrorSeverity = "medium"
	SeverityLow      ErrorSeverity = "low"
	SeverityInfo     ErrorSeverity = "info"
)

// ErrorPriority represents the priority level for error handling
type ErrorPriority string

const (
	PriorityUrgent   ErrorPriority = "urgent"
	PriorityHigh     ErrorPriority = "high"
	PriorityMedium   ErrorPriority = "medium"
	PriorityLow      ErrorPriority = "low"
	PriorityDeferred ErrorPriority = "deferred"
)

// CategorizedError represents a categorized and prioritized error
type CategorizedError struct {
	ID              string                 `json:"id"`
	OriginalError   error                  `json:"original_error"`
	Message         string                 `json:"message"`
	Category        ErrorCategory          `json:"category"`
	Severity        ErrorSeverity          `json:"severity"`
	Priority        ErrorPriority          `json:"priority"`
	Confidence      float64                `json:"confidence"`
	Timestamp       time.Time              `json:"timestamp"`
	Context         map[string]interface{} `json:"context"`
	Source          string                 `json:"source"`
	Operation       string                 `json:"operation"`
	UserID          string                 `json:"user_id"`
	SessionID       string                 `json:"session_id"`
	StackTrace      string                 `json:"stack_trace"`
	Metadata        ErrorMetadata          `json:"metadata"`
	Classification  ErrorClassification    `json:"classification"`
	Recommendations []ErrorRecommendation  `json:"recommendations"`
}

// ErrorMetadata contains additional metadata about the error
type ErrorMetadata struct {
	Frequency          int               `json:"frequency"`
	FirstOccurrence    time.Time         `json:"first_occurrence"`
	LastOccurrence     time.Time         `json:"last_occurrence"`
	AffectedOperations []string          `json:"affected_operations"`
	RelatedErrors      []string          `json:"related_errors"`
	Resolution         *ErrorResolution  `json:"resolution,omitempty"`
	Tags               []string          `json:"tags"`
	CustomFields       map[string]string `json:"custom_fields"`
}

// ErrorClassification provides detailed classification information
type ErrorClassification struct {
	Type            ErrorType            `json:"type"`
	SubType         string               `json:"sub_type"`
	Domain          string               `json:"domain"`
	Component       string               `json:"component"`
	Layer           string               `json:"layer"`
	Retryable       bool                 `json:"retryable"`
	Transient       bool                 `json:"transient"`
	UserActionable  bool                 `json:"user_actionable"`
	BusinessImpact  BusinessImpactLevel  `json:"business_impact"`
	TechnicalImpact TechnicalImpactLevel `json:"technical_impact"`
}

// ErrorType represents the type of error
type ErrorType string

const (
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeConnection ErrorType = "connection"
	ErrorTypeValidation ErrorType = "validation"
	ErrorTypePermission ErrorType = "permission"
	ErrorTypeNotFound   ErrorType = "not_found"
	ErrorTypeConflict   ErrorType = "conflict"
	ErrorTypeInternal   ErrorType = "internal"
	ErrorTypeExternal   ErrorType = "external"
	ErrorTypeRateLimit  ErrorType = "rate_limit"
	ErrorTypeSecurity   ErrorType = "security"
)

// BusinessImpactLevel represents the impact on business operations
type BusinessImpactLevel string

const (
	BusinessImpactCritical BusinessImpactLevel = "critical"
	BusinessImpactHigh     BusinessImpactLevel = "high"
	BusinessImpactMedium   BusinessImpactLevel = "medium"
	BusinessImpactLow      BusinessImpactLevel = "low"
	BusinessImpactNone     BusinessImpactLevel = "none"
)

// TechnicalImpactLevel represents the technical impact
type TechnicalImpactLevel string

const (
	TechnicalImpactSystem    TechnicalImpactLevel = "system"
	TechnicalImpactService   TechnicalImpactLevel = "service"
	TechnicalImpactComponent TechnicalImpactLevel = "component"
	TechnicalImpactOperation TechnicalImpactLevel = "operation"
	TechnicalImpactNone      TechnicalImpactLevel = "none"
)

// ErrorRecommendation provides recommendations for error resolution
type ErrorRecommendation struct {
	Type          RecommendationType `json:"type"`
	Priority      int                `json:"priority"`
	Description   string             `json:"description"`
	Action        string             `json:"action"`
	EstimatedTime time.Duration      `json:"estimated_time"`
	Resources     []string           `json:"resources"`
	Dependencies  []string           `json:"dependencies"`
	RiskLevel     string             `json:"risk_level"`
	SuccessRate   float64            `json:"success_rate"`
}

// RecommendationType represents different types of recommendations
type RecommendationType string

const (
	RecommendationImmediate  RecommendationType = "immediate"
	RecommendationShortTerm  RecommendationType = "short_term"
	RecommendationLongTerm   RecommendationType = "long_term"
	RecommendationPreventive RecommendationType = "preventive"
	RecommendationMonitoring RecommendationType = "monitoring"
)

// ErrorResolution contains information about error resolution
type ErrorResolution struct {
	Status          ResolutionStatus `json:"status"`
	Method          string           `json:"method"`
	ResolvedAt      time.Time        `json:"resolved_at"`
	ResolvedBy      string           `json:"resolved_by"`
	Resolution      string           `json:"resolution"`
	PreventionSteps []string         `json:"prevention_steps"`
	LessonsLearned  []string         `json:"lessons_learned"`
	TimeToResolve   time.Duration    `json:"time_to_resolve"`
}

// ResolutionStatus represents the status of error resolution
type ResolutionStatus string

const (
	ResolutionStatusOpen       ResolutionStatus = "open"
	ResolutionStatusInProgress ResolutionStatus = "in_progress"
	ResolutionStatusResolved   ResolutionStatus = "resolved"
	ResolutionStatusClosed     ResolutionStatus = "closed"
	ResolutionStatusReopened   ResolutionStatus = "reopened"
)

// ErrorPatterns contains pattern matching rules for error categorization
type ErrorPatterns struct {
	CategoryPatterns  map[ErrorCategory][]*regexp.Regexp
	SeverityPatterns  map[ErrorSeverity][]*regexp.Regexp
	TypePatterns      map[ErrorType][]*regexp.Regexp
	RetryablePatterns []*regexp.Regexp
	TransientPatterns []*regexp.Regexp
	SecurityPatterns  []*regexp.Regexp
}

// ErrorAnalytics provides analytics and trending for errors
type ErrorAnalytics struct {
	ErrorHistory    []CategorizedError            `json:"error_history"`
	CategoryStats   map[ErrorCategory]*ErrorStats `json:"category_stats"`
	SeverityStats   map[ErrorSeverity]*ErrorStats `json:"severity_stats"`
	TrendAnalysis   *ErrorTrendAnalysis           `json:"trend_analysis"`
	HotspotAnalysis *ErrorHotspotAnalysis         `json:"hotspot_analysis"`
	PredictiveData  *ErrorPredictiveData          `json:"predictive_data"`
}

// ErrorStats contains statistical information about errors
type ErrorStats struct {
	Count            int                `json:"count"`
	Frequency        float64            `json:"frequency"`
	AverageFrequency float64            `json:"average_frequency"`
	TrendDirection   TrendDirection     `json:"trend_direction"`
	TrendStrength    float64            `json:"trend_strength"`
	Distribution     map[string]int     `json:"distribution"`
	TimeSeries       []TimeSeriesPoint  `json:"time_series"`
	Correlations     map[string]float64 `json:"correlations"`
}

// TrendDirection represents the direction of an error trend
type TrendDirection string

const (
	TrendIncreasing TrendDirection = "increasing"
	TrendDecreasing TrendDirection = "decreasing"
	TrendStable     TrendDirection = "stable"
	TrendVolatile   TrendDirection = "volatile"
)

// TimeSeriesPoint represents a point in time series data
type TimeSeriesPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Count     int       `json:"count"`
}

// ErrorTrendAnalysis provides trend analysis for errors
type ErrorTrendAnalysis struct {
	OverallTrend       TrendDirection            `json:"overall_trend"`
	CategoryTrends     map[ErrorCategory]float64 `json:"category_trends"`
	SeverityTrends     map[ErrorSeverity]float64 `json:"severity_trends"`
	SeasonalPatterns   []ErrorSeasonalPattern    `json:"seasonal_patterns"`
	AnomalyDetection   []ErrorAnomaly            `json:"anomaly_detection"`
	PredictedIncidents []PredictedIncident       `json:"predicted_incidents"`
}

// ErrorSeasonalPattern represents seasonal error patterns
type ErrorSeasonalPattern struct {
	Pattern     string    `json:"pattern"`
	Frequency   string    `json:"frequency"`
	Amplitude   float64   `json:"amplitude"`
	Confidence  float64   `json:"confidence"`
	NextPeak    time.Time `json:"next_peak"`
	Description string    `json:"description"`
}

// ErrorAnomaly represents detected error anomalies
type ErrorAnomaly struct {
	Timestamp   time.Time     `json:"timestamp"`
	Type        string        `json:"type"`
	Severity    ErrorSeverity `json:"severity"`
	Description string        `json:"description"`
	Expected    float64       `json:"expected"`
	Actual      float64       `json:"actual"`
	Deviation   float64       `json:"deviation"`
	Confidence  float64       `json:"confidence"`
}

// PredictedIncident represents predicted future incidents
type PredictedIncident struct {
	PredictedTime time.Time     `json:"predicted_time"`
	Category      ErrorCategory `json:"category"`
	Severity      ErrorSeverity `json:"severity"`
	Probability   float64       `json:"probability"`
	Impact        string        `json:"impact"`
	Mitigation    []string      `json:"mitigation"`
}

// ErrorHotspotAnalysis identifies error hotspots
type ErrorHotspotAnalysis struct {
	ComponentHotspots  []ComponentHotspot  `json:"component_hotspots"`
	OperationHotspots  []OperationHotspot  `json:"operation_hotspots"`
	UserHotspots       []UserHotspot       `json:"user_hotspots"`
	TimeHotspots       []TimeHotspot       `json:"time_hotspots"`
	GeographicHotspots []GeographicHotspot `json:"geographic_hotspots"`
}

// ComponentHotspot represents a component with high error rates
type ComponentHotspot struct {
	Component   string   `json:"component"`
	ErrorRate   float64  `json:"error_rate"`
	ErrorCount  int      `json:"error_count"`
	Impact      string   `json:"impact"`
	Trend       string   `json:"trend"`
	Suggestions []string `json:"suggestions"`
}

// OperationHotspot represents an operation with high error rates
type OperationHotspot struct {
	Operation   string        `json:"operation"`
	ErrorRate   float64       `json:"error_rate"`
	ErrorCount  int           `json:"error_count"`
	AvgDuration time.Duration `json:"avg_duration"`
	Impact      string        `json:"impact"`
	Suggestions []string      `json:"suggestions"`
}

// UserHotspot represents users experiencing high error rates
type UserHotspot struct {
	UserID      string   `json:"user_id"`
	ErrorRate   float64  `json:"error_rate"`
	ErrorCount  int      `json:"error_count"`
	Pattern     string   `json:"pattern"`
	Suggestions []string `json:"suggestions"`
}

// TimeHotspot represents time periods with high error rates
type TimeHotspot struct {
	TimeRange  string  `json:"time_range"`
	ErrorRate  float64 `json:"error_rate"`
	ErrorCount int     `json:"error_count"`
	Pattern    string  `json:"pattern"`
	Recurrence string  `json:"recurrence"`
}

// GeographicHotspot represents geographic regions with high error rates
type GeographicHotspot struct {
	Region     string  `json:"region"`
	ErrorRate  float64 `json:"error_rate"`
	ErrorCount int     `json:"error_count"`
	Pattern    string  `json:"pattern"`
}

// ErrorPredictiveData contains predictive analytics for errors
type ErrorPredictiveData struct {
	RiskScore          float64              `json:"risk_score"`
	PredictedErrorRate float64              `json:"predicted_error_rate"`
	ConfidenceInterval [2]float64           `json:"confidence_interval"`
	ModelAccuracy      float64              `json:"model_accuracy"`
	PredictionHorizon  time.Duration        `json:"prediction_horizon"`
	RiskFactors        []RiskFactor         `json:"risk_factors"`
	Scenarios          []PredictiveScenario `json:"scenarios"`
}

// RiskFactor represents factors that contribute to error risk
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Weight      float64 `json:"weight"`
	Impact      float64 `json:"impact"`
	Trend       string  `json:"trend"`
	Description string  `json:"description"`
}

// PredictiveScenario represents different predictive scenarios
type PredictiveScenario struct {
	Name        string   `json:"name"`
	Probability float64  `json:"probability"`
	Impact      string   `json:"impact"`
	Timeline    string   `json:"timeline"`
	Mitigation  []string `json:"mitigation"`
	Triggers    []string `json:"triggers"`
}

// NewErrorCategorizer creates a new error categorizer with default configuration
func NewErrorCategorizer(logger *zap.Logger, config *CategorizationConfig) *ErrorCategorizer {
	if config == nil {
		config = &CategorizationConfig{
			EnableAnalytics:      true,
			EnableTrends:         true,
			EnablePrioritization: true,
			SeverityThresholds: map[ErrorSeverity]float64{
				SeverityCritical: 0.9,
				SeverityHigh:     0.7,
				SeverityMedium:   0.5,
				SeverityLow:      0.3,
				SeverityInfo:     0.1,
			},
			CategoryWeights: map[ErrorCategory]float64{
				CategorySecurity:       1.0,
				CategorySystem:         0.9,
				CategoryDatabase:       0.8,
				CategoryBusiness:       0.7,
				CategoryNetwork:        0.6,
				CategoryAuthentication: 0.8,
				CategoryAuthorization:  0.7,
				CategoryValidation:     0.4,
				CategoryExternal:       0.5,
				CategoryConfiguration:  0.6,
				CategoryPerformance:    0.5,
				CategoryUnknown:        0.3,
			},
			MaxAnalyticsHistory: 10000,
			TrendAnalysisWindow: 24 * time.Hour,
			PriorityAdjustment: PriorityAdjustmentConfig{
				FrequencyMultiplier:   1.5,
				RecencyMultiplier:     1.2,
				ImpactMultiplier:      2.0,
				SystemHealthFactor:    1.3,
				BusinessCriticalBoost: 2.5,
			},
		}
	}

	return &ErrorCategorizer{
		logger:   logger,
		config:   config,
		patterns: initializeErrorPatterns(),
		analytics: &ErrorAnalytics{
			ErrorHistory:  make([]CategorizedError, 0),
			CategoryStats: make(map[ErrorCategory]*ErrorStats),
			SeverityStats: make(map[ErrorSeverity]*ErrorStats),
		},
	}
}

// CategorizeError categorizes and prioritizes an error
func (ec *ErrorCategorizer) CategorizeError(ctx context.Context, err error, context map[string]interface{}) *CategorizedError {
	if err == nil {
		return nil
	}

	categorizedErr := &CategorizedError{
		ID:            generateErrorID(),
		OriginalError: err,
		Message:       err.Error(),
		Timestamp:     time.Now(),
		Context:       context,
		Metadata: ErrorMetadata{
			FirstOccurrence: time.Now(),
			LastOccurrence:  time.Now(),
			Frequency:       1,
			Tags:            []string{},
			CustomFields:    make(map[string]string),
		},
	}

	// Extract context information
	ec.extractContextInfo(categorizedErr, context)

	// Categorize the error
	ec.categorizeError(categorizedErr)

	// Determine severity
	ec.determineSeverity(categorizedErr)

	// Calculate priority
	ec.calculatePriority(categorizedErr)

	// Generate classification
	ec.generateClassification(categorizedErr)

	// Generate recommendations
	ec.generateRecommendations(categorizedErr)

	// Update analytics if enabled
	if ec.config.EnableAnalytics {
		ec.updateAnalytics(categorizedErr)
	}

	ec.logger.Info("error categorized",
		zap.String("error_id", categorizedErr.ID),
		zap.String("category", string(categorizedErr.Category)),
		zap.String("severity", string(categorizedErr.Severity)),
		zap.String("priority", string(categorizedErr.Priority)),
		zap.Float64("confidence", categorizedErr.Confidence))

	return categorizedErr
}

// initializeErrorPatterns initializes error pattern matching rules
func initializeErrorPatterns() *ErrorPatterns {
	patterns := &ErrorPatterns{
		CategoryPatterns: make(map[ErrorCategory][]*regexp.Regexp),
		SeverityPatterns: make(map[ErrorSeverity][]*regexp.Regexp),
		TypePatterns:     make(map[ErrorType][]*regexp.Regexp),
	}

	// Network category patterns
	patterns.CategoryPatterns[CategoryNetwork] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(connection|network|timeout|refused|unreachable|dns)`),
		regexp.MustCompile(`(?i)(socket|tcp|udp|http|ssl|tls)`),
		regexp.MustCompile(`(?i)(proxy|gateway|firewall|port)`),
	}

	// Database category patterns
	patterns.CategoryPatterns[CategoryDatabase] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(database|db|sql|query|connection|deadlock)`),
		regexp.MustCompile(`(?i)(postgres|mysql|mongodb|redis|table|index)`),
		regexp.MustCompile(`(?i)(constraint|foreign key|duplicate|unique)`),
	}

	// Validation category patterns
	patterns.CategoryPatterns[CategoryValidation] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(validation|invalid|required|missing|format)`),
		regexp.MustCompile(`(?i)(parse|syntax|schema|type|range)`),
		regexp.MustCompile(`(?i)(malformed|corrupt|encode|decode)`),
	}

	// Authentication category patterns
	patterns.CategoryPatterns[CategoryAuthentication] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(auth|login|password|token|credential)`),
		regexp.MustCompile(`(?i)(session|cookie|jwt|oauth|saml)`),
		regexp.MustCompile(`(?i)(expired|invalid|unauthorized)`),
	}

	// Authorization category patterns
	patterns.CategoryPatterns[CategoryAuthorization] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(permission|access|forbidden|denied|role)`),
		regexp.MustCompile(`(?i)(privilege|scope|policy|acl)`),
		regexp.MustCompile(`(?i)(unauthorized|insufficient)`),
	}

	// Security category patterns
	patterns.CategoryPatterns[CategorySecurity] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(security|attack|intrusion|malware|virus)`),
		regexp.MustCompile(`(?i)(injection|xss|csrf|sql injection)`),
		regexp.MustCompile(`(?i)(vulnerability|exploit|breach|leak)`),
	}

	// Performance category patterns
	patterns.CategoryPatterns[CategoryPerformance] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(performance|slow|timeout|memory|cpu)`),
		regexp.MustCompile(`(?i)(latency|throughput|bottleneck|queue)`),
		regexp.MustCompile(`(?i)(overload|congestion|throttle)`),
	}

	// Critical severity patterns
	patterns.SeverityPatterns[SeverityCritical] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(critical|fatal|emergency|disaster)`),
		regexp.MustCompile(`(?i)(system down|service unavailable|crash)`),
		regexp.MustCompile(`(?i)(data loss|corruption|breach)`),
	}

	// High severity patterns
	patterns.SeverityPatterns[SeverityHigh] = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(error|failed|failure|exception)`),
		regexp.MustCompile(`(?i)(unable|cannot|refused|denied)`),
		regexp.MustCompile(`(?i)(timeout|deadlock|conflict)`),
	}

	// Retryable error patterns
	patterns.RetryablePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(timeout|connection|network|temporary)`),
		regexp.MustCompile(`(?i)(rate limit|throttle|busy|overload)`),
		regexp.MustCompile(`(?i)(service unavailable|try again)`),
	}

	// Transient error patterns
	patterns.TransientPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(temporary|transient|intermittent)`),
		regexp.MustCompile(`(?i)(network|connection|timeout)`),
		regexp.MustCompile(`(?i)(rate limit|throttle|congestion)`),
	}

	return patterns
}

// extractContextInfo extracts relevant information from the context
func (ec *ErrorCategorizer) extractContextInfo(categorizedErr *CategorizedError, context map[string]interface{}) {
	if context == nil {
		return
	}

	if source, ok := context["source"].(string); ok {
		categorizedErr.Source = source
	}

	if operation, ok := context["operation"].(string); ok {
		categorizedErr.Operation = operation
	}

	if userID, ok := context["user_id"].(string); ok {
		categorizedErr.UserID = userID
	}

	if sessionID, ok := context["session_id"].(string); ok {
		categorizedErr.SessionID = sessionID
	}

	if stackTrace, ok := context["stack_trace"].(string); ok {
		categorizedErr.StackTrace = stackTrace
	}
}

// categorizeError determines the category of an error
func (ec *ErrorCategorizer) categorizeError(categorizedErr *CategorizedError) {
	message := categorizedErr.Message
	bestCategory := CategoryUnknown
	highestConfidence := 0.0

	for category, patterns := range ec.patterns.CategoryPatterns {
		confidence := ec.calculatePatternConfidence(message, patterns)
		if confidence > highestConfidence {
			highestConfidence = confidence
			bestCategory = category
		}
	}

	categorizedErr.Category = bestCategory
	categorizedErr.Confidence = highestConfidence
}

// determineSeverity determines the severity of an error
func (ec *ErrorCategorizer) determineSeverity(categorizedErr *CategorizedError) {
	message := categorizedErr.Message
	bestSeverity := SeverityMedium
	highestConfidence := 0.0

	for severity, patterns := range ec.patterns.SeverityPatterns {
		confidence := ec.calculatePatternConfidence(message, patterns)
		if confidence > highestConfidence {
			highestConfidence = confidence
			bestSeverity = severity
		}
	}

	// Adjust severity based on category weight
	if weight, exists := ec.config.CategoryWeights[categorizedErr.Category]; exists {
		if weight >= 0.8 && bestSeverity == SeverityLow {
			bestSeverity = SeverityMedium
		} else if weight >= 0.9 && bestSeverity == SeverityMedium {
			bestSeverity = SeverityHigh
		}
	}

	categorizedErr.Severity = bestSeverity
}

// calculatePriority calculates the priority of an error
func (ec *ErrorCategorizer) calculatePriority(categorizedErr *CategorizedError) {
	if !ec.config.EnablePrioritization {
		categorizedErr.Priority = PriorityMedium
		return
	}

	// Base priority calculation
	var basePriority float64
	switch categorizedErr.Severity {
	case SeverityCritical:
		basePriority = 5.0
	case SeverityHigh:
		basePriority = 4.0
	case SeverityMedium:
		basePriority = 3.0
	case SeverityLow:
		basePriority = 2.0
	case SeverityInfo:
		basePriority = 1.0
	}

	// Apply category weight
	if weight, exists := ec.config.CategoryWeights[categorizedErr.Category]; exists {
		basePriority *= weight
	}

	// Apply adjustments
	adjustedPriority := basePriority

	// Frequency adjustment (would need historical data)
	adjustedPriority *= ec.config.PriorityAdjustment.FrequencyMultiplier

	// Determine final priority
	switch {
	case adjustedPriority >= 4.5:
		categorizedErr.Priority = PriorityUrgent
	case adjustedPriority >= 3.5:
		categorizedErr.Priority = PriorityHigh
	case adjustedPriority >= 2.5:
		categorizedErr.Priority = PriorityMedium
	case adjustedPriority >= 1.5:
		categorizedErr.Priority = PriorityLow
	default:
		categorizedErr.Priority = PriorityDeferred
	}
}

// generateClassification generates detailed classification for an error
func (ec *ErrorCategorizer) generateClassification(categorizedErr *CategorizedError) {
	message := categorizedErr.Message

	classification := ErrorClassification{
		Domain:          "industry_codes",
		Component:       "unknown",
		Layer:           "application",
		BusinessImpact:  BusinessImpactMedium,
		TechnicalImpact: TechnicalImpactOperation,
	}

	// Determine error type
	for errorType, patterns := range ec.patterns.TypePatterns {
		if ec.calculatePatternConfidence(message, patterns) > 0.7 {
			classification.Type = errorType
			break
		}
	}

	// Determine if retryable
	classification.Retryable = ec.calculatePatternConfidence(message, ec.patterns.RetryablePatterns) > 0.5

	// Determine if transient
	classification.Transient = ec.calculatePatternConfidence(message, ec.patterns.TransientPatterns) > 0.5

	// Determine user actionable
	classification.UserActionable = categorizedErr.Category == CategoryValidation ||
		categorizedErr.Category == CategoryAuthentication ||
		categorizedErr.Category == CategoryAuthorization

	categorizedErr.Classification = classification
}

// generateRecommendations generates recommendations for error resolution
func (ec *ErrorCategorizer) generateRecommendations(categorizedErr *CategorizedError) {
	recommendations := []ErrorRecommendation{}

	// Severity-specific recommendations (critical gets highest priority)
	if categorizedErr.Severity == SeverityCritical {
		recommendations = append(recommendations, ErrorRecommendation{
			Type:          RecommendationImmediate,
			Priority:      0,
			Description:   "Escalate to on-call team immediately",
			Action:        "escalate_to_oncall",
			EstimatedTime: 2 * time.Minute,
			Resources:     []string{"oncall_team", "incident_management"},
			RiskLevel:     "critical",
			SuccessRate:   1.0,
		})
	}

	// Category-specific recommendations
	switch categorizedErr.Category {
	case CategoryNetwork:
		recommendations = append(recommendations, ErrorRecommendation{
			Type:          RecommendationImmediate,
			Priority:      1,
			Description:   "Check network connectivity and retry operation",
			Action:        "retry_with_backoff",
			EstimatedTime: 5 * time.Minute,
			Resources:     []string{"network_team", "monitoring"},
			RiskLevel:     "low",
			SuccessRate:   0.8,
		})
	case CategoryDatabase:
		recommendations = append(recommendations, ErrorRecommendation{
			Type:          RecommendationImmediate,
			Priority:      1,
			Description:   "Check database health and connection pool",
			Action:        "monitor_database_metrics",
			EstimatedTime: 10 * time.Minute,
			Resources:     []string{"dba_team", "monitoring"},
			RiskLevel:     "medium",
			SuccessRate:   0.9,
		})
	case CategoryValidation:
		recommendations = append(recommendations, ErrorRecommendation{
			Type:          RecommendationImmediate,
			Priority:      1,
			Description:   "Review input validation and provide user feedback",
			Action:        "improve_validation_messages",
			EstimatedTime: 15 * time.Minute,
			Resources:     []string{"developer", "ux_team"},
			RiskLevel:     "low",
			SuccessRate:   0.95,
		})
	case CategoryAuthentication:
		recommendations = append(recommendations, ErrorRecommendation{
			Type:          RecommendationImmediate,
			Priority:      1,
			Description:   "Verify authentication credentials and token validity",
			Action:        "check_authentication",
			EstimatedTime: 5 * time.Minute,
			Resources:     []string{"security_team", "developer"},
			RiskLevel:     "medium",
			SuccessRate:   0.9,
		})
	case CategoryAuthorization:
		recommendations = append(recommendations, ErrorRecommendation{
			Type:          RecommendationImmediate,
			Priority:      1,
			Description:   "Review user permissions and access control",
			Action:        "check_permissions",
			EstimatedTime: 10 * time.Minute,
			Resources:     []string{"security_team", "admin"},
			RiskLevel:     "medium",
			SuccessRate:   0.85,
		})
	case CategorySecurity:
		recommendations = append(recommendations, ErrorRecommendation{
			Type:          RecommendationImmediate,
			Priority:      1,
			Description:   "Investigate security incident and apply countermeasures",
			Action:        "security_investigation",
			EstimatedTime: 30 * time.Minute,
			Resources:     []string{"security_team", "incident_response"},
			RiskLevel:     "high",
			SuccessRate:   0.95,
		})
	case CategoryPerformance:
		recommendations = append(recommendations, ErrorRecommendation{
			Type:          RecommendationImmediate,
			Priority:      1,
			Description:   "Monitor system performance and optimize if needed",
			Action:        "performance_monitoring",
			EstimatedTime: 15 * time.Minute,
			Resources:     []string{"performance_team", "monitoring"},
			RiskLevel:     "medium",
			SuccessRate:   0.8,
		})
	default:
		// Default recommendation for unknown categories
		recommendations = append(recommendations, ErrorRecommendation{
			Type:          RecommendationImmediate,
			Priority:      2,
			Description:   "Log error details and investigate root cause",
			Action:        "investigate_error",
			EstimatedTime: 20 * time.Minute,
			Resources:     []string{"developer", "support_team"},
			RiskLevel:     "low",
			SuccessRate:   0.7,
		})
	}

	// General recommendation for error monitoring
	recommendations = append(recommendations, ErrorRecommendation{
		Type:          RecommendationMonitoring,
		Priority:      3,
		Description:   "Set up monitoring and alerting for this error pattern",
		Action:        "setup_monitoring",
		EstimatedTime: 30 * time.Minute,
		Resources:     []string{"monitoring_team", "ops"},
		RiskLevel:     "low",
		SuccessRate:   0.9,
	})

	// Sort recommendations by priority
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].Priority < recommendations[j].Priority
	})

	categorizedErr.Recommendations = recommendations
}

// calculatePatternConfidence calculates confidence score for pattern matching
func (ec *ErrorCategorizer) calculatePatternConfidence(message string, patterns []*regexp.Regexp) float64 {
	if len(patterns) == 0 {
		return 0.0
	}

	matches := 0
	for _, pattern := range patterns {
		if pattern.MatchString(message) {
			matches++
		}
	}

	return float64(matches) / float64(len(patterns))
}

// updateAnalytics updates error analytics with the new error
func (ec *ErrorCategorizer) updateAnalytics(categorizedErr *CategorizedError) {
	// Add to history
	ec.analytics.ErrorHistory = append(ec.analytics.ErrorHistory, *categorizedErr)

	// Limit history size
	if len(ec.analytics.ErrorHistory) > ec.config.MaxAnalyticsHistory {
		ec.analytics.ErrorHistory = ec.analytics.ErrorHistory[1:]
	}

	// Update category stats
	if _, exists := ec.analytics.CategoryStats[categorizedErr.Category]; !exists {
		ec.analytics.CategoryStats[categorizedErr.Category] = &ErrorStats{
			Distribution: make(map[string]int),
			TimeSeries:   make([]TimeSeriesPoint, 0),
			Correlations: make(map[string]float64),
		}
	}
	ec.analytics.CategoryStats[categorizedErr.Category].Count++

	// Update severity stats
	if _, exists := ec.analytics.SeverityStats[categorizedErr.Severity]; !exists {
		ec.analytics.SeverityStats[categorizedErr.Severity] = &ErrorStats{
			Distribution: make(map[string]int),
			TimeSeries:   make([]TimeSeriesPoint, 0),
			Correlations: make(map[string]float64),
		}
	}
	ec.analytics.SeverityStats[categorizedErr.Severity].Count++

	// Perform trend analysis if enabled
	if ec.config.EnableTrends {
		ec.performTrendAnalysis()
	}
}

// performTrendAnalysis performs trend analysis on error data
func (ec *ErrorCategorizer) performTrendAnalysis() {
	// This is a simplified trend analysis implementation
	// In a real system, this would be much more sophisticated

	trendAnalysis := &ErrorTrendAnalysis{
		CategoryTrends:     make(map[ErrorCategory]float64),
		SeverityTrends:     make(map[ErrorSeverity]float64),
		SeasonalPatterns:   []ErrorSeasonalPattern{},
		AnomalyDetection:   []ErrorAnomaly{},
		PredictedIncidents: []PredictedIncident{},
	}

	// Calculate category trends
	for category, stats := range ec.analytics.CategoryStats {
		if len(stats.TimeSeries) >= 2 {
			recent := stats.TimeSeries[len(stats.TimeSeries)-1]
			previous := stats.TimeSeries[len(stats.TimeSeries)-2]
			trend := (recent.Value - previous.Value) / previous.Value
			trendAnalysis.CategoryTrends[category] = trend
		}
	}

	// Determine overall trend
	overallTrend := 0.0
	trendCount := 0
	for _, trend := range trendAnalysis.CategoryTrends {
		overallTrend += trend
		trendCount++
	}

	if trendCount > 0 {
		avgTrend := overallTrend / float64(trendCount)
		switch {
		case avgTrend > 0.1:
			trendAnalysis.OverallTrend = TrendIncreasing
		case avgTrend < -0.1:
			trendAnalysis.OverallTrend = TrendDecreasing
		default:
			trendAnalysis.OverallTrend = TrendStable
		}
	}

	ec.analytics.TrendAnalysis = trendAnalysis
}

// GetAnalytics returns current error analytics
func (ec *ErrorCategorizer) GetAnalytics() *ErrorAnalytics {
	return ec.analytics
}

// GetCategoryStats returns statistics for a specific category
func (ec *ErrorCategorizer) GetCategoryStats(category ErrorCategory) *ErrorStats {
	if stats, exists := ec.analytics.CategoryStats[category]; exists {
		return stats
	}
	return nil
}

// ResetAnalytics resets all analytics data
func (ec *ErrorCategorizer) ResetAnalytics() {
	ec.analytics = &ErrorAnalytics{
		ErrorHistory:  make([]CategorizedError, 0),
		CategoryStats: make(map[ErrorCategory]*ErrorStats),
		SeverityStats: make(map[ErrorSeverity]*ErrorStats),
	}
}

// generateErrorID generates a unique error ID
func generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}

// GetErrorsByCategory returns errors filtered by category
func (ec *ErrorCategorizer) GetErrorsByCategory(category ErrorCategory) []CategorizedError {
	var errors []CategorizedError
	for _, err := range ec.analytics.ErrorHistory {
		if err.Category == category {
			errors = append(errors, err)
		}
	}
	return errors
}

// GetErrorsBySeverity returns errors filtered by severity
func (ec *ErrorCategorizer) GetErrorsBySeverity(severity ErrorSeverity) []CategorizedError {
	var errors []CategorizedError
	for _, err := range ec.analytics.ErrorHistory {
		if err.Severity == severity {
			errors = append(errors, err)
		}
	}
	return errors
}

// GetErrorsByPriority returns errors filtered by priority
func (ec *ErrorCategorizer) GetErrorsByPriority(priority ErrorPriority) []CategorizedError {
	var errors []CategorizedError
	for _, err := range ec.analytics.ErrorHistory {
		if err.Priority == priority {
			errors = append(errors, err)
		}
	}
	return errors
}

// GetTopErrorCategories returns the most frequent error categories
func (ec *ErrorCategorizer) GetTopErrorCategories(limit int) []ErrorCategory {
	type categoryCount struct {
		category ErrorCategory
		count    int
	}

	var categories []categoryCount
	for category, stats := range ec.analytics.CategoryStats {
		categories = append(categories, categoryCount{
			category: category,
			count:    stats.Count,
		})
	}

	// Sort by count descending
	sort.Slice(categories, func(i, j int) bool {
		return categories[i].count > categories[j].count
	})

	// Extract top categories
	var result []ErrorCategory
	for i, cat := range categories {
		if i >= limit {
			break
		}
		result = append(result, cat.category)
	}

	return result
}
