package classification

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EnhancedClassification represents an enhanced classification with all new features
type EnhancedClassification struct {
	ID                   uuid.UUID `json:"id" db:"id"`
	BusinessName         string    `json:"business_name" db:"business_name"`
	IndustryCode         string    `json:"industry_code" db:"industry_code"`
	IndustryName         string    `json:"industry_name" db:"industry_name"`
	ConfidenceScore      float64   `json:"confidence_score" db:"confidence_score"`
	ClassificationMethod string    `json:"classification_method" db:"classification_method"`
	Description          string    `json:"description" db:"description"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`

	// Enhanced fields
	MLModelVersion          *string  `json:"ml_model_version,omitempty" db:"ml_model_version"`
	MLConfidenceScore       *float64 `json:"ml_confidence_score,omitempty" db:"ml_confidence_score"`
	CrosswalkMappings       *JSONB   `json:"crosswalk_mappings,omitempty" db:"crosswalk_mappings"`
	GeographicRegion        *string  `json:"geographic_region,omitempty" db:"geographic_region"`
	RegionConfidenceScore   *float64 `json:"region_confidence_score,omitempty" db:"region_confidence_score"`
	IndustrySpecificData    *JSONB   `json:"industry_specific_data,omitempty" db:"industry_specific_data"`
	ClassificationAlgorithm *string  `json:"classification_algorithm,omitempty" db:"classification_algorithm"`
	ValidationRulesApplied  *JSONB   `json:"validation_rules_applied,omitempty" db:"validation_rules_applied"`
	ProcessingTimeMS        *int     `json:"processing_time_ms,omitempty" db:"processing_time_ms"`
	EnhancedMetadata        *JSONB   `json:"enhanced_metadata,omitempty" db:"enhanced_metadata"`
}

// FeedbackModel represents feedback data model
type FeedbackModel struct {
	ID                        uuid.UUID  `json:"id" db:"id"`
	UserID                    string     `json:"user_id" db:"user_id"`
	BusinessName              string     `json:"business_name" db:"business_name"`
	OriginalClassificationID  *uuid.UUID `json:"original_classification_id,omitempty" db:"original_classification_id"`
	FeedbackType              string     `json:"feedback_type" db:"feedback_type"`
	FeedbackValue             *JSONB     `json:"feedback_value,omitempty" db:"feedback_value"`
	FeedbackText              *string    `json:"feedback_text,omitempty" db:"feedback_text"`
	SuggestedClassificationID *uuid.UUID `json:"suggested_classification_id,omitempty" db:"suggested_classification_id"`
	ConfidenceScore           *float64   `json:"confidence_score,omitempty" db:"confidence_score"`
	Status                    string     `json:"status" db:"status"`
	ProcessingTimeMS          *int       `json:"processing_time_ms,omitempty" db:"processing_time_ms"`
	CreatedAt                 time.Time  `json:"created_at" db:"created_at"`
	ProcessedAt               *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	Metadata                  *JSONB     `json:"metadata,omitempty" db:"metadata"`
}

// AccuracyValidationModel represents accuracy validation data model
type AccuracyValidationModel struct {
	ID                       uuid.UUID  `json:"id" db:"id"`
	ClassificationID         *uuid.UUID `json:"classification_id,omitempty" db:"classification_id"`
	MetricType               string     `json:"metric_type" db:"metric_type"`
	Dimension                string     `json:"dimension" db:"dimension"`
	TotalClassifications     int        `json:"total_classifications" db:"total_classifications"`
	CorrectClassifications   int        `json:"correct_classifications" db:"correct_classifications"`
	IncorrectClassifications int        `json:"incorrect_classifications" db:"incorrect_classifications"`
	AccuracyScore            *float64   `json:"accuracy_score,omitempty" db:"accuracy_score"`
	ConfidenceScore          *float64   `json:"confidence_score,omitempty" db:"confidence_score"`
	ProcessingTimeMS         *int       `json:"processing_time_ms,omitempty" db:"processing_time_ms"`
	TimeRangeSeconds         *int       `json:"time_range_seconds,omitempty" db:"time_range_seconds"`
	CreatedAt                time.Time  `json:"created_at" db:"created_at"`
	Metadata                 *JSONB     `json:"metadata,omitempty" db:"metadata"`
}

// AccuracyAlertModel represents accuracy alert data model
type AccuracyAlertModel struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	MetricType     string     `json:"metric_type" db:"metric_type"`
	Dimension      string     `json:"dimension" db:"dimension"`
	Threshold      float64    `json:"threshold" db:"threshold"`
	CurrentValue   float64    `json:"current_value" db:"current_value"`
	Severity       string     `json:"severity" db:"severity"`
	Message        string     `json:"message" db:"message"`
	Status         string     `json:"status" db:"status"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty" db:"acknowledged_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	Metadata       *JSONB     `json:"metadata,omitempty" db:"metadata"`
}

// AccuracyThresholdModel represents accuracy threshold data model
type AccuracyThresholdModel struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	MetricType    string     `json:"metric_type" db:"metric_type"`
	Threshold     float64    `json:"threshold" db:"threshold"`
	Severity      string     `json:"severity" db:"severity"`
	AlertEnabled  bool       `json:"alert_enabled" db:"alert_enabled"`
	Description   *string    `json:"description,omitempty" db:"description"`
	LastTriggered *time.Time `json:"last_triggered,omitempty" db:"last_triggered"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	Metadata      *JSONB     `json:"metadata,omitempty" db:"metadata"`
}

// MLModelVersionModel represents ML model version data model
type MLModelVersionModel struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	ModelName        string     `json:"model_name" db:"model_name"`
	Version          string     `json:"version" db:"version"`
	ModelType        string     `json:"model_type" db:"model_type"`
	FilePath         *string    `json:"file_path,omitempty" db:"file_path"`
	AccuracyScore    *float64   `json:"accuracy_score,omitempty" db:"accuracy_score"`
	TrainingDataSize *int       `json:"training_data_size,omitempty" db:"training_data_size"`
	TrainingDate     *time.Time `json:"training_date,omitempty" db:"training_date"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	Metadata         *JSONB     `json:"metadata,omitempty" db:"metadata"`
}

// CrosswalkMappingModel represents crosswalk mapping data model
type CrosswalkMappingModel struct {
	ID              uuid.UUID `json:"id" db:"id"`
	SourceCode      string    `json:"source_code" db:"source_code"`
	SourceSystem    string    `json:"source_system" db:"source_system"`
	TargetCode      string    `json:"target_code" db:"target_code"`
	TargetSystem    string    `json:"target_system" db:"target_system"`
	ConfidenceScore float64   `json:"confidence_score" db:"confidence_score"`
	ValidationRules *JSONB    `json:"validation_rules,omitempty" db:"validation_rules"`
	IsValid         bool      `json:"is_valid" db:"is_valid"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
	Metadata        *JSONB    `json:"metadata,omitempty" db:"metadata"`
}

// GeographicRegionModel represents geographic region data model
type GeographicRegionModel struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	RegionName          string     `json:"region_name" db:"region_name"`
	RegionType          string     `json:"region_type" db:"region_type"`
	ParentRegionID      *uuid.UUID `json:"parent_region_id,omitempty" db:"parent_region_id"`
	IndustryPatterns    *JSONB     `json:"industry_patterns,omitempty" db:"industry_patterns"`
	ConfidenceModifiers *JSONB     `json:"confidence_modifiers,omitempty" db:"confidence_modifiers"`
	IsActive            bool       `json:"is_active" db:"is_active"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	Metadata            *JSONB     `json:"metadata,omitempty" db:"metadata"`
}

// IndustryMappingModel represents industry mapping data model
type IndustryMappingModel struct {
	ID                      uuid.UUID `json:"id" db:"id"`
	IndustryType            string    `json:"industry_type" db:"industry_type"`
	IndustryName            string    `json:"industry_name" db:"industry_name"`
	NAICSCodes              []string  `json:"naics_codes" db:"naics_codes"`
	SICCodes                []string  `json:"sic_codes" db:"sic_codes"`
	MCCCodes                []string  `json:"mcc_codes" db:"mcc_codes"`
	Keywords                []string  `json:"keywords" db:"keywords"`
	ConfidenceScore         float64   `json:"confidence_score" db:"confidence_score"`
	ValidationRules         *JSONB    `json:"validation_rules,omitempty" db:"validation_rules"`
	ClassificationAlgorithm string    `json:"classification_algorithm" db:"classification_algorithm"`
	IsActive                bool      `json:"is_active" db:"is_active"`
	CreatedAt               time.Time `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time `json:"updated_at" db:"updated_at"`
	Metadata                *JSONB    `json:"metadata,omitempty" db:"metadata"`
}

// DashboardWidgetModel represents dashboard widget data model
type DashboardWidgetModel struct {
	ID          uuid.UUID `json:"id" db:"id"`
	WidgetID    string    `json:"widget_id" db:"widget_id"`
	WidgetType  string    `json:"widget_type" db:"widget_type"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description,omitempty" db:"description"`
	Position    *JSONB    `json:"position,omitempty" db:"position"`
	Config      *JSONB    `json:"config,omitempty" db:"config"`
	Data        *JSONB    `json:"data,omitempty" db:"data"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Metadata    *JSONB    `json:"metadata,omitempty" db:"metadata"`
}

// DashboardMetricModel represents dashboard metric data model
type DashboardMetricModel struct {
	ID         uuid.UUID `json:"id" db:"id"`
	MetricID   string    `json:"metric_id" db:"metric_id"`
	Name       string    `json:"name" db:"name"`
	Value      float64   `json:"value" db:"value"`
	Unit       *string   `json:"unit,omitempty" db:"unit"`
	Trend      *string   `json:"trend,omitempty" db:"trend"`
	TrendValue *float64  `json:"trend_value,omitempty" db:"trend_value"`
	Status     string    `json:"status" db:"status"`
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
	Metadata   *JSONB    `json:"metadata,omitempty" db:"metadata"`
}

// AlertingRuleModel represents alerting rule data model
type AlertingRuleModel struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	RuleID        string     `json:"rule_id" db:"rule_id"`
	Name          string     `json:"name" db:"name"`
	Description   *string    `json:"description,omitempty" db:"description"`
	MetricType    string     `json:"metric_type" db:"metric_type"`
	Condition     string     `json:"condition" db:"condition"`
	Threshold     float64    `json:"threshold" db:"threshold"`
	Severity      string     `json:"severity" db:"severity"`
	IsEnabled     bool       `json:"is_enabled" db:"is_enabled"`
	Actions       []string   `json:"actions" db:"actions"`
	LastTriggered *time.Time `json:"last_triggered,omitempty" db:"last_triggered"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	Metadata      *JSONB     `json:"metadata,omitempty" db:"metadata"`
}

// AccuracyReportModel represents accuracy report data model
type AccuracyReportModel struct {
	ID               uuid.UUID `json:"id" db:"id"`
	ReportID         string    `json:"report_id" db:"report_id"`
	Title            string    `json:"title" db:"title"`
	Description      *string   `json:"description,omitempty" db:"description"`
	ReportType       string    `json:"report_type" db:"report_type"`
	TimeRangeSeconds int       `json:"time_range_seconds" db:"time_range_seconds"`
	Data             *JSONB    `json:"data,omitempty" db:"data"`
	Summary          *string   `json:"summary,omitempty" db:"summary"`
	Recommendations  []string  `json:"recommendations" db:"recommendations"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	Metadata         *JSONB    `json:"metadata,omitempty" db:"metadata"`
}

// JSONB is a custom type for handling JSONB data in PostgreSQL
type JSONB map[string]interface{}

// Value implements the driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan %T into JSONB", value)
	}

	return json.Unmarshal(bytes, j)
}

// MarshalJSON implements json.Marshaler
func (j JSONB) MarshalJSON() ([]byte, error) {
	if j == nil {
		return []byte("null"), nil
	}
	return json.Marshal(map[string]interface{}(j))
}

// UnmarshalJSON implements json.Unmarshaler
func (j *JSONB) UnmarshalJSON(data []byte) error {
	if j == nil {
		*j = make(JSONB)
	}
	return json.Unmarshal(data, (*map[string]interface{})(j))
}

// Validation functions

// ValidateEnhancedClassification validates an enhanced classification
func ValidateEnhancedClassification(classification *EnhancedClassification) error {
	if classification.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}
	if classification.IndustryCode == "" {
		return fmt.Errorf("industry code is required")
	}
	if classification.ConfidenceScore < 0.0 || classification.ConfidenceScore > 1.0 {
		return fmt.Errorf("confidence score must be between 0.0 and 1.0")
	}
	if classification.MLConfidenceScore != nil && (*classification.MLConfidenceScore < 0.0 || *classification.MLConfidenceScore > 1.0) {
		return fmt.Errorf("ML confidence score must be between 0.0 and 1.0")
	}
	if classification.RegionConfidenceScore != nil && (*classification.RegionConfidenceScore < 0.0 || *classification.RegionConfidenceScore > 1.0) {
		return fmt.Errorf("region confidence score must be between 0.0 and 1.0")
	}
	if classification.ProcessingTimeMS != nil && *classification.ProcessingTimeMS < 0 {
		return fmt.Errorf("processing time must be non-negative")
	}
	return nil
}

// ValidateFeedback validates feedback data
func ValidateFeedback(feedback *FeedbackModel) error {
	if feedback.UserID == "" {
		return fmt.Errorf("user ID is required")
	}
	if feedback.BusinessName == "" {
		return fmt.Errorf("business name is required")
	}
	if feedback.FeedbackType == "" {
		return fmt.Errorf("feedback type is required")
	}

	// Validate feedback type
	validTypes := []string{"accuracy", "relevance", "confidence", "classification", "suggestion", "correction"}
	valid := false
	for _, t := range validTypes {
		if feedback.FeedbackType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid feedback type: %s", feedback.FeedbackType)
	}

	// Validate status
	validStatuses := []string{"pending", "processed", "rejected", "applied"}
	valid = false
	for _, s := range validStatuses {
		if feedback.Status == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status: %s", feedback.Status)
	}

	if feedback.ConfidenceScore != nil && (*feedback.ConfidenceScore < 0.0 || *feedback.ConfidenceScore > 1.0) {
		return fmt.Errorf("confidence score must be between 0.0 and 1.0")
	}

	return nil
}

// ValidateAccuracyValidation validates accuracy validation data
func ValidateAccuracyValidation(validation *AccuracyValidationModel) error {
	if validation.MetricType == "" {
		return fmt.Errorf("metric type is required")
	}
	if validation.Dimension == "" {
		return fmt.Errorf("dimension is required")
	}
	if validation.TotalClassifications < 0 {
		return fmt.Errorf("total classifications must be non-negative")
	}
	if validation.CorrectClassifications < 0 {
		return fmt.Errorf("correct classifications must be non-negative")
	}
	if validation.IncorrectClassifications < 0 {
		return fmt.Errorf("incorrect classifications must be non-negative")
	}
	if validation.AccuracyScore != nil && (*validation.AccuracyScore < 0.0 || *validation.AccuracyScore > 1.0) {
		return fmt.Errorf("accuracy score must be between 0.0 and 1.0")
	}
	if validation.ConfidenceScore != nil && (*validation.ConfidenceScore < 0.0 || *validation.ConfidenceScore > 1.0) {
		return fmt.Errorf("confidence score must be between 0.0 and 1.0")
	}

	// Validate metric type
	validTypes := []string{"overall", "industry", "business_type", "region", "confidence", "time_range"}
	valid := false
	for _, t := range validTypes {
		if validation.MetricType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid metric type: %s", validation.MetricType)
	}

	return nil
}

// ValidateAccuracyAlert validates accuracy alert data
func ValidateAccuracyAlert(alert *AccuracyAlertModel) error {
	if alert.MetricType == "" {
		return fmt.Errorf("metric type is required")
	}
	if alert.Dimension == "" {
		return fmt.Errorf("dimension is required")
	}
	if alert.Threshold < 0.0 || alert.Threshold > 1.0 {
		return fmt.Errorf("threshold must be between 0.0 and 1.0")
	}
	if alert.CurrentValue < 0.0 || alert.CurrentValue > 1.0 {
		return fmt.Errorf("current value must be between 0.0 and 1.0")
	}
	if alert.Message == "" {
		return fmt.Errorf("message is required")
	}

	// Validate severity
	validSeverities := []string{"low", "medium", "high", "critical"}
	valid := false
	for _, s := range validSeverities {
		if alert.Severity == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid severity: %s", alert.Severity)
	}

	// Validate status
	validStatuses := []string{"active", "acknowledged", "resolved"}
	valid = false
	for _, s := range validStatuses {
		if alert.Status == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status: %s", alert.Status)
	}

	// Validate metric type
	validTypes := []string{"overall", "industry", "business_type", "region", "confidence", "time_range"}
	valid = false
	for _, t := range validTypes {
		if alert.MetricType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid metric type: %s", alert.MetricType)
	}

	return nil
}

// ValidateMLModelVersion validates ML model version data
func ValidateMLModelVersion(model *MLModelVersionModel) error {
	if model.ModelName == "" {
		return fmt.Errorf("model name is required")
	}
	if model.Version == "" {
		return fmt.Errorf("version is required")
	}
	if model.ModelType == "" {
		return fmt.Errorf("model type is required")
	}

	// Validate model type
	validTypes := []string{"bert", "ensemble", "custom"}
	valid := false
	for _, t := range validTypes {
		if model.ModelType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid model type: %s", model.ModelType)
	}

	if model.AccuracyScore != nil && (*model.AccuracyScore < 0.0 || *model.AccuracyScore > 1.0) {
		return fmt.Errorf("accuracy score must be between 0.0 and 1.0")
	}

	if model.TrainingDataSize != nil && *model.TrainingDataSize < 0 {
		return fmt.Errorf("training data size must be non-negative")
	}

	return nil
}

// ValidateCrosswalkMapping validates crosswalk mapping data
func ValidateCrosswalkMapping(mapping *CrosswalkMappingModel) error {
	if mapping.SourceCode == "" {
		return fmt.Errorf("source code is required")
	}
	if mapping.SourceSystem == "" {
		return fmt.Errorf("source system is required")
	}
	if mapping.TargetCode == "" {
		return fmt.Errorf("target code is required")
	}
	if mapping.TargetSystem == "" {
		return fmt.Errorf("target system is required")
	}
	if mapping.ConfidenceScore < 0.0 || mapping.ConfidenceScore > 1.0 {
		return fmt.Errorf("confidence score must be between 0.0 and 1.0")
	}

	// Validate systems
	validSystems := []string{"naics", "sic", "mcc"}
	valid := false
	for _, s := range validSystems {
		if mapping.SourceSystem == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid source system: %s", mapping.SourceSystem)
	}

	valid = false
	for _, s := range validSystems {
		if mapping.TargetSystem == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid target system: %s", mapping.TargetSystem)
	}

	if mapping.SourceSystem == mapping.TargetSystem {
		return fmt.Errorf("source and target systems must be different")
	}

	return nil
}

// ValidateGeographicRegion validates geographic region data
func ValidateGeographicRegion(region *GeographicRegionModel) error {
	if region.RegionName == "" {
		return fmt.Errorf("region name is required")
	}
	if region.RegionType == "" {
		return fmt.Errorf("region type is required")
	}

	// Validate region type
	validTypes := []string{"country", "state", "city", "postal_code"}
	valid := false
	for _, t := range validTypes {
		if region.RegionType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid region type: %s", region.RegionType)
	}

	return nil
}

// ValidateIndustryMapping validates industry mapping data
func ValidateIndustryMapping(mapping *IndustryMappingModel) error {
	if mapping.IndustryType == "" {
		return fmt.Errorf("industry type is required")
	}
	if mapping.IndustryName == "" {
		return fmt.Errorf("industry name is required")
	}
	if mapping.ConfidenceScore < 0.0 || mapping.ConfidenceScore > 1.0 {
		return fmt.Errorf("confidence score must be between 0.0 and 1.0")
	}
	if mapping.ClassificationAlgorithm == "" {
		return fmt.Errorf("classification algorithm is required")
	}

	// Validate industry type
	validTypes := []string{"agriculture", "retail", "food", "manufacturing", "technology", "finance", "healthcare", "other"}
	valid := false
	for _, t := range validTypes {
		if mapping.IndustryType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid industry type: %s", mapping.IndustryType)
	}

	// Validate classification algorithm
	validAlgorithms := []string{"keyword_based", "code_density", "hybrid"}
	valid = false
	for _, a := range validAlgorithms {
		if mapping.ClassificationAlgorithm == a {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid classification algorithm: %s", mapping.ClassificationAlgorithm)
	}

	return nil
}

// ValidateDashboardWidget validates dashboard widget data
func ValidateDashboardWidget(widget *DashboardWidgetModel) error {
	if widget.WidgetID == "" {
		return fmt.Errorf("widget ID is required")
	}
	if widget.WidgetType == "" {
		return fmt.Errorf("widget type is required")
	}
	if widget.Title == "" {
		return fmt.Errorf("title is required")
	}

	// Validate widget type
	validTypes := []string{"metric", "chart", "table", "alert"}
	valid := false
	for _, t := range validTypes {
		if widget.WidgetType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid widget type: %s", widget.WidgetType)
	}

	return nil
}

// ValidateDashboardMetric validates dashboard metric data
func ValidateDashboardMetric(metric *DashboardMetricModel) error {
	if metric.MetricID == "" {
		return fmt.Errorf("metric ID is required")
	}
	if metric.Name == "" {
		return fmt.Errorf("name is required")
	}
	if metric.Status == "" {
		return fmt.Errorf("status is required")
	}

	// Validate status
	validStatuses := []string{"good", "warning", "critical"}
	valid := false
	for _, s := range validStatuses {
		if metric.Status == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid status: %s", metric.Status)
	}

	// Validate trend
	if metric.Trend != nil {
		validTrends := []string{"up", "down", "stable"}
		valid = false
		for _, t := range validTrends {
			if *metric.Trend == t {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid trend: %s", *metric.Trend)
		}
	}

	return nil
}

// ValidateAlertingRule validates alerting rule data
func ValidateAlertingRule(rule *AlertingRuleModel) error {
	if rule.RuleID == "" {
		return fmt.Errorf("rule ID is required")
	}
	if rule.Name == "" {
		return fmt.Errorf("name is required")
	}
	if rule.MetricType == "" {
		return fmt.Errorf("metric type is required")
	}
	if rule.Condition == "" {
		return fmt.Errorf("condition is required")
	}
	if rule.Severity == "" {
		return fmt.Errorf("severity is required")
	}

	// Validate condition
	validConditions := []string{"above", "below", "equals"}
	valid := false
	for _, c := range validConditions {
		if rule.Condition == c {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid condition: %s", rule.Condition)
	}

	// Validate severity
	validSeverities := []string{"low", "medium", "high", "critical"}
	valid = false
	for _, s := range validSeverities {
		if rule.Severity == s {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid severity: %s", rule.Severity)
	}

	return nil
}

// ValidateAccuracyReport validates accuracy report data
func ValidateAccuracyReport(report *AccuracyReportModel) error {
	if report.ReportID == "" {
		return fmt.Errorf("report ID is required")
	}
	if report.Title == "" {
		return fmt.Errorf("title is required")
	}
	if report.ReportType == "" {
		return fmt.Errorf("report type is required")
	}
	if report.TimeRangeSeconds <= 0 {
		return fmt.Errorf("time range must be positive")
	}

	// Validate report type
	validTypes := []string{"daily", "weekly", "monthly"}
	valid := false
	for _, t := range validTypes {
		if report.ReportType == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid report type: %s", report.ReportType)
	}

	return nil
}

// Serialization/Deserialization helpers

// ToJSON converts a model to JSON
func (e *EnhancedClassification) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// FromJSON populates a model from JSON
func (e *EnhancedClassification) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// ToJSON converts a model to JSON
func (f *FeedbackModel) ToJSON() ([]byte, error) {
	return json.Marshal(f)
}

// FromJSON populates a model from JSON
func (f *FeedbackModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, f)
}

// ToJSON converts a model to JSON
func (a *AccuracyValidationModel) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// FromJSON populates a model from JSON
func (a *AccuracyValidationModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, a)
}

// ToJSON converts a model to JSON
func (a *AccuracyAlertModel) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// FromJSON populates a model from JSON
func (a *AccuracyAlertModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, a)
}

// ToJSON converts a model to JSON
func (a *AccuracyThresholdModel) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// FromJSON populates a model from JSON
func (a *AccuracyThresholdModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, a)
}

// ToJSON converts a model to JSON
func (m *MLModelVersionModel) ToJSON() ([]byte, error) {
	return json.Marshal(m)
}

// FromJSON populates a model from JSON
func (m *MLModelVersionModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, m)
}

// ToJSON converts a model to JSON
func (c *CrosswalkMappingModel) ToJSON() ([]byte, error) {
	return json.Marshal(c)
}

// FromJSON populates a model from JSON
func (c *CrosswalkMappingModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, c)
}

// ToJSON converts a model to JSON
func (g *GeographicRegionModel) ToJSON() ([]byte, error) {
	return json.Marshal(g)
}

// FromJSON populates a model from JSON
func (g *GeographicRegionModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, g)
}

// ToJSON converts a model to JSON
func (i *IndustryMappingModel) ToJSON() ([]byte, error) {
	return json.Marshal(i)
}

// FromJSON populates a model from JSON
func (i *IndustryMappingModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, i)
}

// ToJSON converts a model to JSON
func (d *DashboardWidgetModel) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// FromJSON populates a model from JSON
func (d *DashboardWidgetModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, d)
}

// ToJSON converts a model to JSON
func (d *DashboardMetricModel) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// FromJSON populates a model from JSON
func (d *DashboardMetricModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, d)
}

// ToJSON converts a model to JSON
func (a *AlertingRuleModel) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// FromJSON populates a model from JSON
func (a *AlertingRuleModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, a)
}

// ToJSON converts a model to JSON
func (a *AccuracyReportModel) ToJSON() ([]byte, error) {
	return json.Marshal(a)
}

// FromJSON populates a model from JSON
func (a *AccuracyReportModel) FromJSON(data []byte) error {
	return json.Unmarshal(data, a)
}
