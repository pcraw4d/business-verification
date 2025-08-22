package metadata

import (
	"context"
	"time"
)

// DataSourceMetadata represents metadata about a data source used in processing
type DataSourceMetadata struct {
	SourceID           string                 `json:"source_id"`
	SourceName         string                 `json:"source_name"`
	SourceType         string                 `json:"source_type"`
	SourceURL          string                 `json:"source_url,omitempty"`
	SourceDescription  string                 `json:"source_description,omitempty"`
	ReliabilityScore   float64                `json:"reliability_score"`
	LastUpdated        time.Time              `json:"last_updated"`
	DataFreshness      time.Duration          `json:"data_freshness"`
	Coverage           map[string]float64     `json:"coverage,omitempty"`
	QualityMetrics     *DataSourceQuality     `json:"quality_metrics,omitempty"`
	PerformanceMetrics *DataSourcePerformance `json:"performance_metrics,omitempty"`
	Attribution        *DataSourceAttribution `json:"attribution,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// DataSourceQuality represents quality metrics for a data source
type DataSourceQuality struct {
	AccuracyScore     float64 `json:"accuracy_score"`
	CompletenessScore float64 `json:"completeness_score"`
	ConsistencyScore  float64 `json:"consistency_score"`
	TimelinessScore   float64 `json:"timeliness_score"`
	ValidityScore     float64 `json:"validity_score"`
	OverallQuality    float64 `json:"overall_quality"`
}

// DataSourcePerformance represents performance metrics for a data source
type DataSourcePerformance struct {
	ResponseTime     time.Duration `json:"response_time"`
	UptimePercentage float64       `json:"uptime_percentage"`
	ErrorRate        float64       `json:"error_rate"`
	SuccessRate      float64       `json:"success_rate"`
	Throughput       float64       `json:"throughput"`
}

// DataSourceAttribution represents attribution information for a data source
type DataSourceAttribution struct {
	ProviderName    string `json:"provider_name"`
	ProviderURL     string `json:"provider_url,omitempty"`
	License         string `json:"license,omitempty"`
	TermsOfService  string `json:"terms_of_service,omitempty"`
	AttributionText string `json:"attribution_text,omitempty"`
	Required        bool   `json:"required"`
}

// ConfidenceLevel represents the confidence level categorization
type ConfidenceLevel string

const (
	ConfidenceLevelVeryHigh ConfidenceLevel = "very_high"
	ConfidenceLevelHigh     ConfidenceLevel = "high"
	ConfidenceLevelMedium   ConfidenceLevel = "medium"
	ConfidenceLevelLow      ConfidenceLevel = "low"
	ConfidenceLevelVeryLow  ConfidenceLevel = "very_low"
)

// ConfidenceMetadata represents detailed confidence information
type ConfidenceMetadata struct {
	OverallConfidence float64                `json:"overall_confidence"`
	ConfidenceLevel   ConfidenceLevel        `json:"confidence_level"`
	ComponentScores   map[string]float64     `json:"component_scores"`
	Factors           []ConfidenceFactor     `json:"factors"`
	Uncertainty       *UncertaintyMetrics    `json:"uncertainty,omitempty"`
	Calibration       *CalibrationData       `json:"calibration,omitempty"`
	CalculatedAt      time.Time              `json:"calculated_at"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ConfidenceFactor represents a factor that contributes to confidence
type ConfidenceFactor struct {
	FactorName   string  `json:"factor_name"`
	FactorValue  float64 `json:"factor_value"`
	FactorWeight float64 `json:"factor_weight"`
	Description  string  `json:"description"`
	Impact       string  `json:"impact"` // "positive", "negative", "neutral"
	Confidence   float64 `json:"confidence"`
}

// UncertaintyMetrics represents uncertainty quantification
type UncertaintyMetrics struct {
	UncertaintyLevel   float64            `json:"uncertainty_level"`
	ConfidenceInterval ConfidenceInterval `json:"confidence_interval"`
	StandardError      float64            `json:"standard_error"`
	Variance           float64            `json:"variance"`
	ReliabilityScore   float64            `json:"reliability_score"`
}

// ConfidenceInterval represents confidence interval bounds
type ConfidenceInterval struct {
	LowerBound float64 `json:"lower_bound"`
	UpperBound float64 `json:"upper_bound"`
	Level      float64 `json:"level"` // 0.95 for 95% confidence interval
}

// CalibrationData represents calibration information
type CalibrationData struct {
	SampleSize         int       `json:"sample_size"`
	AverageScore       float64   `json:"average_score"`
	StandardDeviation  float64   `json:"standard_deviation"`
	CalibrationFactor  float64   `json:"calibration_factor"`
	LastCalibrated     time.Time `json:"last_calibrated"`
	CalibrationMethod  string    `json:"calibration_method"`
	HistoricalAccuracy float64   `json:"historical_accuracy"`
}

// ResponseMetadata represents comprehensive metadata for API responses
type ResponseMetadata struct {
	RequestID      string                 `json:"request_id"`
	ProcessingTime time.Duration          `json:"processing_time"`
	Timestamp      time.Time              `json:"timestamp"`
	APIVersion     string                 `json:"api_version"`
	DataSources    []DataSourceMetadata   `json:"data_sources"`
	Confidence     *ConfidenceMetadata    `json:"confidence,omitempty"`
	Validation     *ValidationMetadata    `json:"validation,omitempty"`
	Quality        *QualityMetadata       `json:"quality,omitempty"`
	Traceability   *TraceabilityMetadata  `json:"traceability,omitempty"`
	Compliance     *ComplianceMetadata    `json:"compliance,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationMetadata represents validation information
type ValidationMetadata struct {
	ValidationStatus string                 `json:"validation_status"`
	ValidationScore  float64                `json:"validation_score"`
	ValidationRules  []ValidationRule       `json:"validation_rules"`
	ValidationErrors []ValidationError      `json:"validation_errors,omitempty"`
	ValidatedAt      time.Time              `json:"validated_at"`
	ValidatorVersion string                 `json:"validator_version"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationRule represents a validation rule that was applied
type ValidationRule struct {
	RuleID          string  `json:"rule_id"`
	RuleName        string  `json:"rule_name"`
	RuleDescription string  `json:"rule_description"`
	Status          string  `json:"status"`   // "passed", "failed", "warning"
	Severity        string  `json:"severity"` // "low", "medium", "high", "critical"
	Confidence      float64 `json:"confidence"`
}

// ValidationError represents a validation error
type ValidationError struct {
	ErrorCode    string   `json:"error_code"`
	ErrorMessage string   `json:"error_message"`
	Field        string   `json:"field,omitempty"`
	Severity     string   `json:"severity"`
	Suggestions  []string `json:"suggestions,omitempty"`
}

// QualityMetadata represents overall quality metrics
type QualityMetadata struct {
	OverallQuality float64                `json:"overall_quality"`
	DataQuality    float64                `json:"data_quality"`
	ProcessQuality float64                `json:"process_quality"`
	OutputQuality  float64                `json:"output_quality"`
	QualityFactors []QualityFactor        `json:"quality_factors"`
	QualityScore   float64                `json:"quality_score"`
	QualityLevel   string                 `json:"quality_level"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// QualityFactor represents a quality factor
type QualityFactor struct {
	FactorName   string  `json:"factor_name"`
	FactorValue  float64 `json:"factor_value"`
	FactorWeight float64 `json:"factor_weight"`
	Description  string  `json:"description"`
	Impact       string  `json:"impact"`
}

// TraceabilityMetadata represents traceability information
type TraceabilityMetadata struct {
	TraceID         string                 `json:"trace_id"`
	CorrelationID   string                 `json:"correlation_id"`
	RequestTrace    []TraceEvent           `json:"request_trace"`
	DataLineage     []DataLineageItem      `json:"data_lineage"`
	ProcessingSteps []ProcessingStep       `json:"processing_steps"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// TraceEvent represents a trace event
type TraceEvent struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	EventName string                 `json:"event_name"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Status    string                 `json:"status"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// DataLineageItem represents a data lineage item
type DataLineageItem struct {
	ItemID         string                 `json:"item_id"`
	SourceID       string                 `json:"source_id"`
	Transformation string                 `json:"transformation"`
	InputData      map[string]interface{} `json:"input_data,omitempty"`
	OutputData     map[string]interface{} `json:"output_data,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
	Confidence     float64                `json:"confidence"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ProcessingStep represents a processing step
type ProcessingStep struct {
	StepID      string                 `json:"step_id"`
	StepName    string                 `json:"step_name"`
	StepType    string                 `json:"step_type"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	Duration    time.Duration          `json:"duration"`
	Status      string                 `json:"status"`
	InputCount  int                    `json:"input_count"`
	OutputCount int                    `json:"output_count"`
	ErrorCount  int                    `json:"error_count"`
	Confidence  float64                `json:"confidence"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ComplianceMetadata represents compliance information
type ComplianceMetadata struct {
	ComplianceStatus string                  `json:"compliance_status"`
	ComplianceScore  float64                 `json:"compliance_score"`
	Frameworks       []ComplianceFramework   `json:"frameworks"`
	Requirements     []ComplianceRequirement `json:"requirements"`
	AuditTrail       []AuditEvent            `json:"audit_trail"`
	Metadata         map[string]interface{}  `json:"metadata,omitempty"`
}

// ComplianceFramework represents a compliance framework
type ComplianceFramework struct {
	FrameworkID   string  `json:"framework_id"`
	FrameworkName string  `json:"framework_name"`
	Version       string  `json:"version"`
	Status        string  `json:"status"`
	Score         float64 `json:"score"`
	Requirements  int     `json:"requirements"`
	Compliant     int     `json:"compliant"`
	NonCompliant  int     `json:"non_compliant"`
}

// ComplianceRequirement represents a compliance requirement
type ComplianceRequirement struct {
	RequirementID   string `json:"requirement_id"`
	RequirementName string `json:"requirement_name"`
	Framework       string `json:"framework"`
	Status          string `json:"status"`
	Severity        string `json:"severity"`
	Description     string `json:"description"`
	Evidence        string `json:"evidence,omitempty"`
}

// AuditEvent represents an audit event
type AuditEvent struct {
	EventID   string                 `json:"event_id"`
	EventType string                 `json:"event_type"`
	EventName string                 `json:"event_name"`
	Timestamp time.Time              `json:"timestamp"`
	UserID    string                 `json:"user_id,omitempty"`
	IPAddress string                 `json:"ip_address,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// MetadataManager provides functionality for managing metadata
type MetadataManager interface {
	// Data source management
	AddDataSource(ctx context.Context, source *DataSourceMetadata) error
	GetDataSource(ctx context.Context, sourceID string) (*DataSourceMetadata, error)
	UpdateDataSource(ctx context.Context, source *DataSourceMetadata) error
	ListDataSources(ctx context.Context) ([]*DataSourceMetadata, error)

	// Confidence management
	CalculateConfidence(ctx context.Context, factors []ConfidenceFactor) (*ConfidenceMetadata, error)
	UpdateConfidence(ctx context.Context, confidence *ConfidenceMetadata) error
	GetConfidenceHistory(ctx context.Context, requestID string) ([]*ConfidenceMetadata, error)

	// Response metadata management
	CreateResponseMetadata(ctx context.Context, requestID string) (*ResponseMetadata, error)
	UpdateResponseMetadata(ctx context.Context, metadata *ResponseMetadata) error
	GetResponseMetadata(ctx context.Context, requestID string) (*ResponseMetadata, error)

	// Validation
	ValidateMetadata(ctx context.Context, metadata *ResponseMetadata) (*ValidationMetadata, error)

	// Quality assessment
	AssessQuality(ctx context.Context, metadata *ResponseMetadata) (*QualityMetadata, error)

	// Traceability
	CreateTraceability(ctx context.Context, requestID string) (*TraceabilityMetadata, error)
	AddTraceEvent(ctx context.Context, traceID string, event *TraceEvent) error
	AddDataLineage(ctx context.Context, traceID string, lineage *DataLineageItem) error
	AddProcessingStep(ctx context.Context, traceID string, step *ProcessingStep) error

	// Compliance
	AssessCompliance(ctx context.Context, metadata *ResponseMetadata) (*ComplianceMetadata, error)

	// Utility methods
	GenerateRequestID() string
	GenerateTraceID() string
	GenerateCorrelationID() string
}
