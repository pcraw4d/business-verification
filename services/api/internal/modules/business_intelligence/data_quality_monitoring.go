package business_intelligence

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DataQualityMonitoringSystem provides comprehensive data quality monitoring and assessment
type DataQualityMonitoringSystem struct {
	config          QualityMonitoringConfig
	logger          *zap.Logger
	monitors        map[string]QualityMonitor
	validators      map[string]QualityValidator
	assessors       map[string]QualityAssessor
	alerters        map[string]QualityAlerter
	reporters       map[string]QualityReporter
	mu              sync.RWMutex
	metrics         *QualityMetrics
	backgroundTasks map[string]*BackgroundTask
}

// QualityMonitoringConfig holds configuration for data quality monitoring
type QualityMonitoringConfig struct {
	// Monitoring configuration
	EnableRealTimeMonitoring   bool          `json:"enable_real_time_monitoring"`
	MonitoringInterval         time.Duration `json:"monitoring_interval"`
	EnableBatchMonitoring      bool          `json:"enable_batch_monitoring"`
	BatchSize                  int           `json:"batch_size"`
	EnableContinuousMonitoring bool          `json:"enable_continuous_monitoring"`

	// Quality thresholds
	DefaultQualityThreshold  float64 `json:"default_quality_threshold"`
	CriticalQualityThreshold float64 `json:"critical_quality_threshold"`
	WarningQualityThreshold  float64 `json:"warning_quality_threshold"`
	EnableAdaptiveThresholds bool    `json:"enable_adaptive_thresholds"`

	// Quality dimensions
	EnableCompletenessMonitoring bool `json:"enable_completeness_monitoring"`
	EnableAccuracyMonitoring     bool `json:"enable_accuracy_monitoring"`
	EnableConsistencyMonitoring  bool `json:"enable_consistency_monitoring"`
	EnableTimelinessMonitoring   bool `json:"enable_timeliness_monitoring"`
	EnableValidityMonitoring     bool `json:"enable_validity_monitoring"`
	EnableUniquenessMonitoring   bool `json:"enable_uniqueness_monitoring"`

	// Alerting configuration
	EnableAlerting      bool          `json:"enable_alerting"`
	AlertThreshold      float64       `json:"alert_threshold"`
	AlertCooldownPeriod time.Duration `json:"alert_cooldown_period"`
	MaxAlertsPerHour    int           `json:"max_alerts_per_hour"`
	EnableEscalation    bool          `json:"enable_escalation"`

	// Reporting configuration
	EnableReporting          bool          `json:"enable_reporting"`
	ReportGenerationInterval time.Duration `json:"report_generation_interval"`
	EnableRealTimeReports    bool          `json:"enable_real_time_reports"`
	EnableHistoricalReports  bool          `json:"enable_historical_reports"`
	ReportRetentionPeriod    time.Duration `json:"report_retention_period"`

	// Performance optimization
	EnableParallelProcessing bool          `json:"enable_parallel_processing"`
	MaxConcurrentMonitors    int           `json:"max_concurrent_monitors"`
	ProcessingTimeout        time.Duration `json:"processing_timeout"`

	// Data storage
	EnableQualityDataStorage bool          `json:"enable_quality_data_storage"`
	StorageRetentionPeriod   time.Duration `json:"storage_retention_period"`
	EnableQualityTrends      bool          `json:"enable_quality_trends"`
}

// QualityMonitor monitors data quality in real-time
type QualityMonitor interface {
	GetName() string
	GetType() string
	GetSupportedDimensions() []QualityDimension
	MonitorData(ctx context.Context, data interface{}) (*QualityAssessment, error)
	GetMonitoringMetrics() *MonitoringMetrics
}

// QualityValidator validates data against quality rules
type QualityValidator interface {
	GetName() string
	GetValidationRules() []QualityRule
	ValidateData(ctx context.Context, data interface{}) (*ValidationResult, error)
	GetValidationMetrics() *ValidationMetrics
}

// QualityAssessor assesses overall data quality
type QualityAssessor interface {
	GetName() string
	GetAssessmentTypes() []AssessmentType
	AssessQuality(ctx context.Context, assessments []*QualityAssessment) (*QualityScore, error)
	GetAssessmentMetrics() *AssessmentMetrics
}

// QualityAlerter sends alerts for quality issues
type QualityAlerter interface {
	GetName() string
	GetAlertTypes() []AlertType
	SendAlert(ctx context.Context, alert *QualityAlert) error
	GetAlertingMetrics() *AlertingMetrics
}

// QualityReporter generates quality reports
type QualityReporter interface {
	GetName() string
	GetReportTypes() []ReportType
	GenerateReport(ctx context.Context, request *ReportRequest) (*QualityReport, error)
	GetReportingMetrics() *ReportingMetrics
}

// QualityDimension represents a dimension of data quality
type QualityDimension string

const (
	QualityDimensionCompleteness QualityDimension = "completeness"
	QualityDimensionAccuracy     QualityDimension = "accuracy"
	QualityDimensionConsistency  QualityDimension = "consistency"
	QualityDimensionTimeliness   QualityDimension = "timeliness"
	QualityDimensionValidity     QualityDimension = "validity"
	QualityDimensionUniqueness   QualityDimension = "uniqueness"
	QualityDimensionRelevance    QualityDimension = "relevance"
	QualityDimensionPrecision    QualityDimension = "precision"
)

// AssessmentType represents a type of quality assessment
type AssessmentType string

const (
	AssessmentTypeRealTime    AssessmentType = "real_time"
	AssessmentTypeBatch       AssessmentType = "batch"
	AssessmentTypeHistorical  AssessmentType = "historical"
	AssessmentTypePredictive  AssessmentType = "predictive"
	AssessmentTypeComparative AssessmentType = "comparative"
)

// AlertType represents a type of quality alert
type AlertType string

const (
	AlertTypeQualityDegradation AlertType = "quality_degradation"
	AlertTypeThresholdBreach    AlertType = "threshold_breach"
	AlertTypeAnomalyDetection   AlertType = "anomaly_detection"
	AlertTypeTrendAnalysis      AlertType = "trend_analysis"
	AlertTypeSystemFailure      AlertType = "system_failure"
)

// ReportType represents a type of quality report
type ReportType string

const (
	ReportTypeSummary     ReportType = "summary"
	ReportTypeDetailed    ReportType = "detailed"
	ReportTypeTrend       ReportType = "trend"
	ReportTypeComparative ReportType = "comparative"
	ReportTypeExecutive   ReportType = "executive"
	ReportTypeTechnical   ReportType = "technical"
)

// QualityAssessment represents a data quality assessment
type QualityAssessment struct {
	ID              string                               `json:"id"`
	MonitorID       string                               `json:"monitor_id"`
	DataID          string                               `json:"data_id"`
	DataType        string                               `json:"data_type"`
	Dimensions      map[QualityDimension]*DimensionScore `json:"dimensions"`
	OverallScore    float64                              `json:"overall_score"`
	QualityLevel    string                               `json:"quality_level"`
	Issues          []QualityIssue                       `json:"issues"`
	Recommendations []QualityRecommendation              `json:"recommendations"`
	Metadata        map[string]interface{}               `json:"metadata"`
	AssessedAt      time.Time                            `json:"assessed_at"`
	ExpiresAt       time.Time                            `json:"expires_at"`
}

// DimensionScore represents a score for a quality dimension
type DimensionScore struct {
	Dimension   QualityDimension `json:"dimension"`
	Score       float64          `json:"score"`
	Weight      float64          `json:"weight"`
	Issues      []QualityIssue   `json:"issues"`
	Confidence  float64          `json:"confidence"`
	LastUpdated time.Time        `json:"last_updated"`
}

// QualityIssue represents a data quality issue
type QualityIssue struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Severity    string                 `json:"severity"`
	Description string                 `json:"description"`
	Field       string                 `json:"field"`
	Value       interface{}            `json:"value"`
	Expected    interface{}            `json:"expected"`
	Rule        string                 `json:"rule"`
	Data        map[string]interface{} `json:"data"`
	DetectedAt  time.Time              `json:"detected_at"`
}

// QualityRecommendation represents a recommendation for improving data quality
type QualityRecommendation struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Priority    string                 `json:"priority"`
	Description string                 `json:"description"`
	Action      string                 `json:"action"`
	Impact      string                 `json:"impact"`
	Effort      string                 `json:"effort"`
	Data        map[string]interface{} `json:"data"`
	CreatedAt   time.Time              `json:"created_at"`
}

// QualityScore represents an overall quality score
type QualityScore struct {
	ID              string                       `json:"id"`
	AssessorID      string                       `json:"assessor_id"`
	AssessmentType  AssessmentType               `json:"assessment_type"`
	OverallScore    float64                      `json:"overall_score"`
	QualityLevel    string                       `json:"quality_level"`
	DimensionScores map[QualityDimension]float64 `json:"dimension_scores"`
	Trend           string                       `json:"trend"`
	Comparison      QualityComparison            `json:"comparison"`
	Metadata        map[string]interface{}       `json:"metadata"`
	CalculatedAt    time.Time                    `json:"calculated_at"`
}

// QualityComparison represents quality comparison data
type QualityComparison struct {
	PreviousScore    float64 `json:"previous_score"`
	Change           float64 `json:"change"`
	ChangePercentage float64 `json:"change_percentage"`
	Benchmark        float64 `json:"benchmark"`
	BenchmarkGap     float64 `json:"benchmark_gap"`
}

// QualityRule represents a data quality rule
type QualityRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Dimension   QualityDimension       `json:"dimension"`
	Severity    string                 `json:"severity"`
	Pattern     string                 `json:"pattern"`
	Threshold   float64                `json:"threshold"`
	Weight      float64                `json:"weight"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// ValidationResult represents the result of data validation
type ValidationResult struct {
	ID             string                 `json:"id"`
	ValidatorID    string                 `json:"validator_id"`
	DataID         string                 `json:"data_id"`
	IsValid        bool                   `json:"is_valid"`
	QualityScore   float64                `json:"quality_score"`
	Issues         []QualityIssue         `json:"issues"`
	RulesApplied   []string               `json:"rules_applied"`
	ValidationTime time.Duration          `json:"validation_time"`
	Metadata       map[string]interface{} `json:"metadata"`
	ValidatedAt    time.Time              `json:"validated_at"`
}

// QualityAlert represents a data quality alert
type QualityAlert struct {
	ID              string                  `json:"id"`
	AlerterID       string                  `json:"alerter_id"`
	Type            AlertType               `json:"type"`
	Severity        string                  `json:"severity"`
	Title           string                  `json:"title"`
	Description     string                  `json:"description"`
	DataID          string                  `json:"data_id"`
	QualityScore    float64                 `json:"quality_score"`
	Threshold       float64                 `json:"threshold"`
	Issues          []QualityIssue          `json:"issues"`
	Recommendations []QualityRecommendation `json:"recommendations"`
	Metadata        map[string]interface{}  `json:"metadata"`
	TriggeredAt     time.Time               `json:"triggered_at"`
	ResolvedAt      *time.Time              `json:"resolved_at"`
}

// QualityReport represents a data quality report
type QualityReport struct {
	ID              string                               `json:"id"`
	ReporterID      string                               `json:"reporter_id"`
	Type            ReportType                           `json:"type"`
	Title           string                               `json:"title"`
	Summary         string                               `json:"summary"`
	QualityScore    float64                              `json:"quality_score"`
	QualityLevel    string                               `json:"quality_level"`
	Dimensions      map[QualityDimension]*DimensionScore `json:"dimensions"`
	Issues          []QualityIssue                       `json:"issues"`
	Recommendations []QualityRecommendation              `json:"recommendations"`
	Trends          []QualityTrend                       `json:"trends"`
	Metadata        map[string]interface{}               `json:"metadata"`
	GeneratedAt     time.Time                            `json:"generated_at"`
	Period          TimePeriod                           `json:"period"`
}

// QualityTrend represents a quality trend
type QualityTrend struct {
	ID         string                 `json:"id"`
	Dimension  QualityDimension       `json:"dimension"`
	Direction  string                 `json:"direction"`
	Magnitude  float64                `json:"magnitude"`
	Confidence float64                `json:"confidence"`
	TimeRange  TimePeriod             `json:"time_range"`
	Data       map[string]interface{} `json:"data"`
	CreatedAt  time.Time              `json:"created_at"`
}

// TimePeriod represents a time period
type TimePeriod struct {
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
	Duration  time.Duration `json:"duration"`
}

// ReportRequest represents a request for a quality report
type ReportRequest struct {
	ID          string                 `json:"id"`
	ReporterID  string                 `json:"reporter_id"`
	Type        ReportType             `json:"type"`
	DataIDs     []string               `json:"data_ids"`
	Dimensions  []QualityDimension     `json:"dimensions"`
	TimeRange   TimePeriod             `json:"time_range"`
	Format      string                 `json:"format"`
	Options     map[string]interface{} `json:"options"`
	RequestedAt time.Time              `json:"requested_at"`
}

// QualityMetrics tracks metrics for the quality monitoring system
type QualityMetrics struct {
	TotalAssessments         int64                         `json:"total_assessments"`
	TotalValidations         int64                         `json:"total_validations"`
	TotalAlerts              int64                         `json:"total_alerts"`
	TotalReports             int64                         `json:"total_reports"`
	AverageQualityScore      float64                       `json:"average_quality_score"`
	QualityScoreDistribution map[string]int64              `json:"quality_score_distribution"`
	IssueDistribution        map[string]int64              `json:"issue_distribution"`
	AlertDistribution        map[string]int64              `json:"alert_distribution"`
	MonitorMetrics           map[string]*MonitoringMetrics `json:"monitor_metrics"`
	ValidatorMetrics         map[string]*ValidationMetrics `json:"validator_metrics"`
	AssessorMetrics          map[string]*AssessmentMetrics `json:"assessor_metrics"`
	AlerterMetrics           map[string]*AlertingMetrics   `json:"alerter_metrics"`
	ReporterMetrics          map[string]*ReportingMetrics  `json:"reporter_metrics"`
	LastUpdated              time.Time                     `json:"last_updated"`
}

// MonitoringMetrics tracks metrics for quality monitoring
type MonitoringMetrics struct {
	MonitorName      string    `json:"monitor_name"`
	TotalAssessments int64     `json:"total_assessments"`
	AverageScore     float64   `json:"average_score"`
	IssuesDetected   int64     `json:"issues_detected"`
	LastAssessment   time.Time `json:"last_assessment"`
}

// ValidationMetrics tracks metrics for data validation
type ValidationMetrics struct {
	ValidatorName         string        `json:"validator_name"`
	TotalValidations      int64         `json:"total_validations"`
	ValidData             int64         `json:"valid_data"`
	InvalidData           int64         `json:"invalid_data"`
	AverageValidationTime time.Duration `json:"average_validation_time"`
	LastValidation        time.Time     `json:"last_validation"`
}

// AssessmentMetrics tracks metrics for quality assessment
type AssessmentMetrics struct {
	AssessorName     string    `json:"assessor_name"`
	TotalAssessments int64     `json:"total_assessments"`
	AverageScore     float64   `json:"average_score"`
	LastAssessment   time.Time `json:"last_assessment"`
}

// AlertingMetrics tracks metrics for quality alerting
type AlertingMetrics struct {
	AlerterName  string    `json:"alerter_name"`
	TotalAlerts  int64     `json:"total_alerts"`
	AlertsSent   int64     `json:"alerts_sent"`
	AlertsFailed int64     `json:"alerts_failed"`
	LastAlert    time.Time `json:"last_alert"`
}

// ReportingMetrics tracks metrics for quality reporting
type ReportingMetrics struct {
	ReporterName     string    `json:"reporter_name"`
	TotalReports     int64     `json:"total_reports"`
	ReportsGenerated int64     `json:"reports_generated"`
	ReportsFailed    int64     `json:"reports_failed"`
	LastReport       time.Time `json:"last_report"`
}

// NewDataQualityMonitoringSystem creates a new data quality monitoring system
func NewDataQualityMonitoringSystem(config QualityMonitoringConfig, logger *zap.Logger) *DataQualityMonitoringSystem {
	return &DataQualityMonitoringSystem{
		config:     config,
		logger:     logger,
		monitors:   make(map[string]QualityMonitor),
		validators: make(map[string]QualityValidator),
		assessors:  make(map[string]QualityAssessor),
		alerters:   make(map[string]QualityAlerter),
		reporters:  make(map[string]QualityReporter),
		metrics: &QualityMetrics{
			QualityScoreDistribution: make(map[string]int64),
			IssueDistribution:        make(map[string]int64),
			AlertDistribution:        make(map[string]int64),
			MonitorMetrics:           make(map[string]*MonitoringMetrics),
			ValidatorMetrics:         make(map[string]*ValidationMetrics),
			AssessorMetrics:          make(map[string]*AssessmentMetrics),
			AlerterMetrics:           make(map[string]*AlertingMetrics),
			ReporterMetrics:          make(map[string]*ReportingMetrics),
		},
		backgroundTasks: make(map[string]*BackgroundTask),
	}
}

// RegisterMonitor registers a quality monitor
func (s *DataQualityMonitoringSystem) RegisterMonitor(monitor QualityMonitor) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := monitor.GetName()
	s.monitors[name] = monitor

	// Initialize metrics
	s.metrics.MonitorMetrics[name] = &MonitoringMetrics{
		MonitorName: name,
	}

	s.logger.Info("Registered quality monitor",
		zap.String("name", name),
		zap.String("type", monitor.GetType()),
		zap.Strings("dimensions", s.qualityDimensionsToStrings(monitor.GetSupportedDimensions())))

	return nil
}

// RegisterValidator registers a quality validator
func (s *DataQualityMonitoringSystem) RegisterValidator(validator QualityValidator) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := validator.GetName()
	s.validators[name] = validator

	// Initialize metrics
	s.metrics.ValidatorMetrics[name] = &ValidationMetrics{
		ValidatorName: name,
	}

	s.logger.Info("Registered quality validator",
		zap.String("name", name),
		zap.Int("rules_count", len(validator.GetValidationRules())))

	return nil
}

// RegisterAssessor registers a quality assessor
func (s *DataQualityMonitoringSystem) RegisterAssessor(assessor QualityAssessor) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := assessor.GetName()
	s.assessors[name] = assessor

	// Initialize metrics
	s.metrics.AssessorMetrics[name] = &AssessmentMetrics{
		AssessorName: name,
	}

	s.logger.Info("Registered quality assessor",
		zap.String("name", name),
		zap.Strings("assessment_types", s.assessmentTypesToStrings(assessor.GetAssessmentTypes())))

	return nil
}

// RegisterAlerter registers a quality alerter
func (s *DataQualityMonitoringSystem) RegisterAlerter(alerter QualityAlerter) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := alerter.GetName()
	s.alerters[name] = alerter

	// Initialize metrics
	s.metrics.AlerterMetrics[name] = &AlertingMetrics{
		AlerterName: name,
	}

	s.logger.Info("Registered quality alerter",
		zap.String("name", name),
		zap.Strings("alert_types", s.alertTypesToStrings(alerter.GetAlertTypes())))

	return nil
}

// RegisterReporter registers a quality reporter
func (s *DataQualityMonitoringSystem) RegisterReporter(reporter QualityReporter) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := reporter.GetName()
	s.reporters[name] = reporter

	// Initialize metrics
	s.metrics.ReporterMetrics[name] = &ReportingMetrics{
		ReporterName: name,
	}

	s.logger.Info("Registered quality reporter",
		zap.String("name", name),
		zap.Strings("report_types", s.reportTypesToStrings(reporter.GetReportTypes())))

	return nil
}

// MonitorDataQuality monitors data quality in real-time
func (s *DataQualityMonitoringSystem) MonitorDataQuality(ctx context.Context, data interface{}, dataID string) (*QualityAssessment, error) {
	s.logger.Debug("Monitoring data quality",
		zap.String("data_id", dataID))

	// Find appropriate monitor
	monitor := s.selectMonitor(data)
	if monitor == nil {
		return nil, fmt.Errorf("no suitable monitor found for data")
	}

	// Monitor data
	assessment, err := monitor.MonitorData(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("monitoring failed: %w", err)
	}

	// Set data ID
	assessment.DataID = dataID

	// Update metrics
	s.updateMonitoringMetrics(monitor.GetName(), assessment)

	// Check for alerts
	if s.config.EnableAlerting {
		s.checkAndSendAlerts(ctx, assessment)
	}

	s.logger.Debug("Data quality monitoring completed",
		zap.String("data_id", dataID),
		zap.Float64("quality_score", assessment.OverallScore),
		zap.String("quality_level", assessment.QualityLevel))

	return assessment, nil
}

// ValidateData validates data against quality rules
func (s *DataQualityMonitoringSystem) ValidateData(ctx context.Context, data interface{}, dataID string) (*ValidationResult, error) {
	s.logger.Debug("Validating data",
		zap.String("data_id", dataID))

	// Find appropriate validator
	validator := s.selectValidator(data)
	if validator == nil {
		return nil, fmt.Errorf("no suitable validator found for data")
	}

	// Validate data
	result, err := validator.ValidateData(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Set data ID
	result.DataID = dataID

	// Update metrics
	s.updateValidationMetrics(validator.GetName(), result)

	s.logger.Debug("Data validation completed",
		zap.String("data_id", dataID),
		zap.Bool("is_valid", result.IsValid),
		zap.Float64("quality_score", result.QualityScore))

	return result, nil
}

// AssessQuality assesses overall data quality
func (s *DataQualityMonitoringSystem) AssessQuality(ctx context.Context, assessments []*QualityAssessment) (*QualityScore, error) {
	s.logger.Debug("Assessing data quality",
		zap.Int("assessment_count", len(assessments)))

	// Find appropriate assessor
	assessor := s.selectAssessor()
	if assessor == nil {
		return nil, fmt.Errorf("no suitable assessor found")
	}

	// Assess quality
	score, err := assessor.AssessQuality(ctx, assessments)
	if err != nil {
		return nil, fmt.Errorf("quality assessment failed: %w", err)
	}

	// Update metrics
	s.updateAssessmentMetrics(assessor.GetName(), score)

	s.logger.Debug("Quality assessment completed",
		zap.Float64("overall_score", score.OverallScore),
		zap.String("quality_level", score.QualityLevel))

	return score, nil
}

// GenerateQualityReport generates a quality report
func (s *DataQualityMonitoringSystem) GenerateQualityReport(ctx context.Context, request *ReportRequest) (*QualityReport, error) {
	s.logger.Info("Generating quality report",
		zap.String("report_id", request.ID),
		zap.String("type", string(request.Type)))

	// Find appropriate reporter
	reporter := s.selectReporter(request.Type)
	if reporter == nil {
		return nil, fmt.Errorf("no suitable reporter found for type: %s", request.Type)
	}

	// Generate report
	report, err := reporter.GenerateReport(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("report generation failed: %w", err)
	}

	// Update metrics
	s.updateReportingMetrics(reporter.GetName(), report)

	s.logger.Info("Quality report generated",
		zap.String("report_id", report.ID),
		zap.String("type", string(report.Type)),
		zap.Float64("quality_score", report.QualityScore))

	return report, nil
}

// Helper methods

// selectMonitor selects an appropriate monitor for the data
func (s *DataQualityMonitoringSystem) selectMonitor(data interface{}) QualityMonitor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return first available monitor as fallback
	for _, monitor := range s.monitors {
		return monitor
	}

	return nil
}

// selectValidator selects an appropriate validator for the data
func (s *DataQualityMonitoringSystem) selectValidator(data interface{}) QualityValidator {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return first available validator as fallback
	for _, validator := range s.validators {
		return validator
	}

	return nil
}

// selectAssessor selects an appropriate assessor
func (s *DataQualityMonitoringSystem) selectAssessor() QualityAssessor {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return first available assessor as fallback
	for _, assessor := range s.assessors {
		return assessor
	}

	return nil
}

// selectReporter selects an appropriate reporter for the report type
func (s *DataQualityMonitoringSystem) selectReporter(reportType ReportType) QualityReporter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find reporter that supports this report type
	for _, reporter := range s.reporters {
		for _, supportedType := range reporter.GetReportTypes() {
			if supportedType == reportType {
				return reporter
			}
		}
	}

	// Return first available reporter as fallback
	for _, reporter := range s.reporters {
		return reporter
	}

	return nil
}

// checkAndSendAlerts checks if alerts should be sent and sends them
func (s *DataQualityMonitoringSystem) checkAndSendAlerts(ctx context.Context, assessment *QualityAssessment) {
	if assessment.OverallScore < s.config.AlertThreshold {
		// Create alert
		alert := &QualityAlert{
			ID:              generateID(),
			Type:            AlertTypeQualityDegradation,
			Severity:        s.determineAlertSeverity(assessment.OverallScore),
			Title:           "Data Quality Degradation Detected",
			Description:     fmt.Sprintf("Data quality score %f is below threshold %f", assessment.OverallScore, s.config.AlertThreshold),
			DataID:          assessment.DataID,
			QualityScore:    assessment.OverallScore,
			Threshold:       s.config.AlertThreshold,
			Issues:          assessment.Issues,
			Recommendations: assessment.Recommendations,
			TriggeredAt:     time.Now(),
		}

		// Send alert
		s.sendAlert(ctx, alert)
	}
}

// determineAlertSeverity determines the severity of an alert based on quality score
func (s *DataQualityMonitoringSystem) determineAlertSeverity(score float64) string {
	if score < s.config.CriticalQualityThreshold {
		return "critical"
	} else if score < s.config.WarningQualityThreshold {
		return "warning"
	}
	return "info"
}

// sendAlert sends an alert using available alerters
func (s *DataQualityMonitoringSystem) sendAlert(ctx context.Context, alert *QualityAlert) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, alerter := range s.alerters {
		err := alerter.SendAlert(ctx, alert)
		if err != nil {
			s.logger.Error("Failed to send alert",
				zap.String("alerter", alerter.GetName()),
				zap.String("alert_id", alert.ID),
				zap.Error(err))
		} else {
			s.updateAlertingMetrics(alerter.GetName(), true)
		}
	}
}

// Metrics update methods

// updateMonitoringMetrics updates monitoring metrics
func (s *DataQualityMonitoringSystem) updateMonitoringMetrics(monitorName string, assessment *QualityAssessment) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.TotalAssessments++
	s.metrics.AverageQualityScore = (s.metrics.AverageQualityScore + assessment.OverallScore) / 2

	// Update quality score distribution
	scoreRange := s.getScoreRange(assessment.OverallScore)
	s.metrics.QualityScoreDistribution[scoreRange]++

	// Update issue distribution
	for _, issue := range assessment.Issues {
		s.metrics.IssueDistribution[issue.Type]++
	}

	// Update monitor-specific metrics
	if metrics, exists := s.metrics.MonitorMetrics[monitorName]; exists {
		metrics.TotalAssessments++
		metrics.AverageScore = (metrics.AverageScore + assessment.OverallScore) / 2
		metrics.IssuesDetected += int64(len(assessment.Issues))
		metrics.LastAssessment = time.Now()
	}

	s.metrics.LastUpdated = time.Now()
}

// updateValidationMetrics updates validation metrics
func (s *DataQualityMonitoringSystem) updateValidationMetrics(validatorName string, result *ValidationResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.TotalValidations++

	// Update validator-specific metrics
	if metrics, exists := s.metrics.ValidatorMetrics[validatorName]; exists {
		metrics.TotalValidations++
		if result.IsValid {
			metrics.ValidData++
		} else {
			metrics.InvalidData++
		}
		metrics.AverageValidationTime = (metrics.AverageValidationTime + result.ValidationTime) / 2
		metrics.LastValidation = time.Now()
	}

	s.metrics.LastUpdated = time.Now()
}

// updateAssessmentMetrics updates assessment metrics
func (s *DataQualityMonitoringSystem) updateAssessmentMetrics(assessorName string, score *QualityScore) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update assessor-specific metrics
	if metrics, exists := s.metrics.AssessorMetrics[assessorName]; exists {
		metrics.TotalAssessments++
		metrics.AverageScore = (metrics.AverageScore + score.OverallScore) / 2
		metrics.LastAssessment = time.Now()
	}

	s.metrics.LastUpdated = time.Now()
}

// updateAlertingMetrics updates alerting metrics
func (s *DataQualityMonitoringSystem) updateAlertingMetrics(alerterName string, success bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.TotalAlerts++

	// Update alerter-specific metrics
	if metrics, exists := s.metrics.AlerterMetrics[alerterName]; exists {
		metrics.TotalAlerts++
		if success {
			metrics.AlertsSent++
		} else {
			metrics.AlertsFailed++
		}
		metrics.LastAlert = time.Now()
	}

	s.metrics.LastUpdated = time.Now()
}

// updateReportingMetrics updates reporting metrics
func (s *DataQualityMonitoringSystem) updateReportingMetrics(reporterName string, report *QualityReport) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.metrics.TotalReports++

	// Update reporter-specific metrics
	if metrics, exists := s.metrics.ReporterMetrics[reporterName]; exists {
		metrics.TotalReports++
		metrics.ReportsGenerated++
		metrics.LastReport = time.Now()
	}

	s.metrics.LastUpdated = time.Now()
}

// getScoreRange returns the score range for a given score
func (s *DataQualityMonitoringSystem) getScoreRange(score float64) string {
	switch {
	case score >= 0.9:
		return "excellent"
	case score >= 0.8:
		return "good"
	case score >= 0.7:
		return "fair"
	case score >= 0.6:
		return "poor"
	default:
		return "critical"
	}
}

// GetMetrics returns current quality monitoring metrics
func (s *DataQualityMonitoringSystem) GetMetrics() *QualityMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *s.metrics
	return &metrics
}

// Utility methods

// qualityDimensionsToStrings converts quality dimensions to strings
func (s *DataQualityMonitoringSystem) qualityDimensionsToStrings(dimensions []QualityDimension) []string {
	var strings []string
	for _, d := range dimensions {
		strings = append(strings, string(d))
	}
	return strings
}

// assessmentTypesToStrings converts assessment types to strings
func (s *DataQualityMonitoringSystem) assessmentTypesToStrings(types []AssessmentType) []string {
	var strings []string
	for _, t := range types {
		strings = append(strings, string(t))
	}
	return strings
}

// alertTypesToStrings converts alert types to strings
func (s *DataQualityMonitoringSystem) alertTypesToStrings(types []AlertType) []string {
	var strings []string
	for _, t := range types {
		strings = append(strings, string(t))
	}
	return strings
}

// reportTypesToStrings converts report types to strings
func (s *DataQualityMonitoringSystem) reportTypesToStrings(types []ReportType) []string {
	var strings []string
	for _, t := range types {
		strings = append(strings, string(t))
	}
	return strings
}
