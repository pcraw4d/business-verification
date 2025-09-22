package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kyb-platform/internal/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// AdvancedSecurityMonitor provides comprehensive security monitoring for classification system
type AdvancedSecurityMonitor struct {
	// Configuration
	config *AdvancedSecurityMonitorConfig

	// Observability
	logger *observability.Logger
	tracer trace.Tracer

	// Core monitoring components
	dataSourceTrustMonitor     *DataSourceTrustMonitor
	websiteVerificationMonitor *WebsiteVerificationMonitor
	securityViolationMonitor   *SecurityViolationMonitor
	confidenceIntegrityMonitor *ConfidenceIntegrityMonitor

	// Metrics and alerts
	securityMetrics *AdvancedSecurityMetrics
	alertManager    *SecurityAlertManager

	// Data storage
	trustRates        map[string]*TrustRateData
	verificationRates map[string]*VerificationRateData
	violationEvents   []*SecurityViolationEvent
	confidenceEvents  []*ConfidenceIntegrityEvent

	// Synchronization
	trustMux        sync.RWMutex
	verificationMux sync.RWMutex
	violationMux    sync.RWMutex
	confidenceMux   sync.RWMutex
	metricsMux      sync.RWMutex

	// Context for shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// AdvancedSecurityMonitorConfig configuration for advanced security monitoring
type AdvancedSecurityMonitorConfig struct {
	// Data source trust monitoring
	TrustRateTarget         float64       `json:"trust_rate_target" yaml:"trust_rate_target"`
	TrustRateCheckInterval  time.Duration `json:"trust_rate_check_interval" yaml:"trust_rate_check_interval"`
	TrustRateHistorySize    int           `json:"trust_rate_history_size" yaml:"trust_rate_history_size"`
	TrustRateAlertThreshold float64       `json:"trust_rate_alert_threshold" yaml:"trust_rate_alert_threshold"`

	// Website verification monitoring
	VerificationRateTarget     float64       `json:"verification_rate_target" yaml:"verification_rate_target"`
	VerificationCheckInterval  time.Duration `json:"verification_check_interval" yaml:"verification_check_interval"`
	VerificationHistorySize    int           `json:"verification_history_size" yaml:"verification_history_size"`
	VerificationAlertThreshold float64       `json:"verification_alert_threshold" yaml:"verification_alert_threshold"`

	// Security violation monitoring
	ViolationDetectionEnabled bool          `json:"violation_detection_enabled" yaml:"violation_detection_enabled"`
	ViolationCheckInterval    time.Duration `json:"violation_check_interval" yaml:"violation_check_interval"`
	ViolationHistorySize      int           `json:"violation_history_size" yaml:"violation_history_size"`
	MaxViolationsPerHour      int           `json:"max_violations_per_hour" yaml:"max_violations_per_hour"`

	// Confidence integrity monitoring
	ConfidenceIntegrityEnabled bool          `json:"confidence_integrity_enabled" yaml:"confidence_integrity_enabled"`
	ConfidenceCheckInterval    time.Duration `json:"confidence_check_interval" yaml:"confidence_check_interval"`
	ConfidenceHistorySize      int           `json:"confidence_history_size" yaml:"confidence_history_size"`
	ConfidenceAnomalyThreshold float64       `json:"confidence_anomaly_threshold" yaml:"confidence_anomaly_threshold"`

	// Alerting
	AlertingEnabled  bool          `json:"alerting_enabled" yaml:"alerting_enabled"`
	AlertCooldown    time.Duration `json:"alert_cooldown" yaml:"alert_cooldown"`
	AlertHistorySize int           `json:"alert_history_size" yaml:"alert_history_size"`

	// External integrations
	WebhookURL     string        `json:"webhook_url" yaml:"webhook_url"`
	WebhookTimeout time.Duration `json:"webhook_timeout" yaml:"webhook_timeout"`
}

// DataSourceTrustMonitor monitors data source trust rates
type DataSourceTrustMonitor struct {
	TrustRates map[string]*TrustRateData
	LastCheck  time.Time
	CheckCount int64
	Mux        sync.RWMutex
}

// TrustRateData represents trust rate data for a data source
type TrustRateData struct {
	DataSourceID      string           `json:"data_source_id"`
	DataSourceName    string           `json:"data_source_name"`
	TrustRate         float64          `json:"trust_rate"`
	TotalRequests     int64            `json:"total_requests"`
	TrustedRequests   int64            `json:"trusted_requests"`
	UntrustedRequests int64            `json:"untrusted_requests"`
	LastUpdated       time.Time        `json:"last_updated"`
	History           []TrustRatePoint `json:"history"`
	Status            TrustStatus      `json:"status"`
	Alerts            []TrustAlert     `json:"alerts"`
}

// TrustRatePoint represents a point in trust rate history
type TrustRatePoint struct {
	TrustRate    float64   `json:"trust_rate"`
	Timestamp    time.Time `json:"timestamp"`
	RequestCount int64     `json:"request_count"`
}

// TrustStatus represents the trust status of a data source
type TrustStatus string

const (
	TrustStatusExcellent TrustStatus = "excellent" // 95-100%
	TrustStatusGood      TrustStatus = "good"      // 90-94%
	TrustStatusWarning   TrustStatus = "warning"   // 80-89%
	TrustStatusCritical  TrustStatus = "critical"  // <80%
)

// TrustAlert represents a trust rate alert
type TrustAlert struct {
	ID           string        `json:"id"`
	DataSourceID string        `json:"data_source_id"`
	Type         AlertType     `json:"type"`
	Severity     AlertSeverity `json:"severity"`
	Message      string        `json:"message"`
	TrustRate    float64       `json:"trust_rate"`
	Threshold    float64       `json:"threshold"`
	Timestamp    time.Time     `json:"timestamp"`
	Resolved     bool          `json:"resolved"`
	ResolvedAt   *time.Time    `json:"resolved_at,omitempty"`
}

// WebsiteVerificationMonitor monitors website verification success rates
type WebsiteVerificationMonitor struct {
	VerificationRates map[string]*VerificationRateData
	LastCheck         time.Time
	CheckCount        int64
	Mux               sync.RWMutex
}

// VerificationRateData represents verification rate data
type VerificationRateData struct {
	WebsiteDomain           string                  `json:"website_domain"`
	VerificationRate        float64                 `json:"verification_rate"`
	TotalAttempts           int64                   `json:"total_attempts"`
	SuccessfulVerifications int64                   `json:"successful_verifications"`
	FailedVerifications     int64                   `json:"failed_verifications"`
	LastUpdated             time.Time               `json:"last_updated"`
	History                 []VerificationRatePoint `json:"history"`
	Status                  VerificationStatus      `json:"status"`
	Alerts                  []VerificationAlert     `json:"alerts"`
}

// VerificationRatePoint represents a point in verification rate history
type VerificationRatePoint struct {
	VerificationRate float64   `json:"verification_rate"`
	Timestamp        time.Time `json:"timestamp"`
	AttemptCount     int64     `json:"attempt_count"`
}

// VerificationStatus represents the verification status
type VerificationStatus string

const (
	VerificationStatusExcellent VerificationStatus = "excellent" // 95-100%
	VerificationStatusGood      VerificationStatus = "good"      // 90-94%
	VerificationStatusWarning   VerificationStatus = "warning"   // 80-89%
	VerificationStatusCritical  VerificationStatus = "critical"  // <80%
)

// VerificationAlert represents a verification rate alert
type VerificationAlert struct {
	ID               string        `json:"id"`
	WebsiteDomain    string        `json:"website_domain"`
	Type             AlertType     `json:"type"`
	Severity         AlertSeverity `json:"severity"`
	Message          string        `json:"message"`
	VerificationRate float64       `json:"verification_rate"`
	Threshold        float64       `json:"threshold"`
	Timestamp        time.Time     `json:"timestamp"`
	Resolved         bool          `json:"resolved"`
	ResolvedAt       *time.Time    `json:"resolved_at,omitempty"`
}

// SecurityViolationMonitor monitors security violations
type SecurityViolationMonitor struct {
	ViolationEvents []*SecurityViolationEvent
	LastCheck       time.Time
	CheckCount      int64
	Mux             sync.RWMutex
}

// SecurityViolationEvent represents a security violation event
type SecurityViolationEvent struct {
	ID          string                 `json:"id"`
	Type        ViolationType          `json:"type"`
	Severity    ViolationSeverity      `json:"severity"`
	Source      string                 `json:"source"`
	Description string                 `json:"description"`
	Details     map[string]interface{} `json:"details"`
	Timestamp   time.Time              `json:"timestamp"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy  string                 `json:"resolved_by,omitempty"`
	Resolution  string                 `json:"resolution,omitempty"`
}

// ViolationType represents the type of security violation
type ViolationType string

const (
	ViolationTypeUntrustedDataSource         ViolationType = "untrusted_data_source"
	ViolationTypeWebsiteVerificationFailure  ViolationType = "website_verification_failure"
	ViolationTypeConfidenceScoreManipulation ViolationType = "confidence_score_manipulation"
	ViolationTypeDataSourceTampering         ViolationType = "data_source_tampering"
	ViolationTypeUnauthorizedAccess          ViolationType = "unauthorized_access"
	ViolationTypeDataIntegrityViolation      ViolationType = "data_integrity_violation"
	ViolationTypeRateLimitViolation          ViolationType = "rate_limit_violation"
	ViolationTypeInputValidationFailure      ViolationType = "input_validation_failure"
)

// ViolationSeverity represents the severity of a security violation
type ViolationSeverity string

const (
	ViolationSeverityLow      ViolationSeverity = "low"
	ViolationSeverityMedium   ViolationSeverity = "medium"
	ViolationSeverityHigh     ViolationSeverity = "high"
	ViolationSeverityCritical ViolationSeverity = "critical"
)

// ConfidenceIntegrityMonitor monitors confidence score integrity
type ConfidenceIntegrityMonitor struct {
	ConfidenceEvents []*ConfidenceIntegrityEvent
	LastCheck        time.Time
	CheckCount       int64
	Mux              sync.RWMutex
}

// ConfidenceIntegrityEvent represents a confidence score integrity event
type ConfidenceIntegrityEvent struct {
	ID               string                  `json:"id"`
	Type             ConfidenceEventType     `json:"type"`
	Severity         ConfidenceEventSeverity `json:"severity"`
	ClassificationID string                  `json:"classification_id"`
	ExpectedScore    float64                 `json:"expected_score"`
	ActualScore      float64                 `json:"actual_score"`
	ScoreDifference  float64                 `json:"score_difference"`
	Details          map[string]interface{}  `json:"details"`
	Timestamp        time.Time               `json:"timestamp"`
	Resolved         bool                    `json:"resolved"`
	ResolvedAt       *time.Time              `json:"resolved_at,omitempty"`
	ResolvedBy       string                  `json:"resolved_by,omitempty"`
	Resolution       string                  `json:"resolution,omitempty"`
}

// ConfidenceEventType represents the type of confidence integrity event
type ConfidenceEventType string

const (
	ConfidenceEventTypeAnomaly           ConfidenceEventType = "anomaly"
	ConfidenceEventTypeManipulation      ConfidenceEventType = "manipulation"
	ConfidenceEventTypeInconsistency     ConfidenceEventType = "inconsistency"
	ConfidenceEventTypeOutOfRange        ConfidenceEventType = "out_of_range"
	ConfidenceEventTypeUnexpectedPattern ConfidenceEventType = "unexpected_pattern"
)

// ConfidenceEventSeverity represents the severity of a confidence integrity event
type ConfidenceEventSeverity string

const (
	ConfidenceEventSeverityLow      ConfidenceEventSeverity = "low"
	ConfidenceEventSeverityMedium   ConfidenceEventSeverity = "medium"
	ConfidenceEventSeverityHigh     ConfidenceEventSeverity = "high"
	ConfidenceEventSeverityCritical ConfidenceEventSeverity = "critical"
)

// SecurityAlertManager manages security alerts
type SecurityAlertManager struct {
	Alerts     []*AdvancedSecurityAlert
	LastAlert  time.Time
	AlertCount int64
	Mux        sync.RWMutex
}

// AdvancedSecurityAlert represents a security alert
type AdvancedSecurityAlert struct {
	ID             string                    `json:"id"`
	Type           AdvancedSecurityAlertType `json:"type"`
	Severity       AlertSeverity             `json:"severity"`
	Title          string                    `json:"title"`
	Message        string                    `json:"message"`
	Source         string                    `json:"source"`
	Details        map[string]interface{}    `json:"details"`
	Timestamp      time.Time                 `json:"timestamp"`
	Acknowledged   bool                      `json:"acknowledged"`
	AcknowledgedAt *time.Time                `json:"acknowledged_at,omitempty"`
	AcknowledgedBy string                    `json:"acknowledged_by,omitempty"`
	Resolved       bool                      `json:"resolved"`
	ResolvedAt     *time.Time                `json:"resolved_at,omitempty"`
	ResolvedBy     string                    `json:"resolved_by,omitempty"`
}

// AdvancedSecurityAlertType represents the type of security alert
type AdvancedSecurityAlertType string

const (
	AdvancedSecurityAlertTypeTrustRateViolation  AdvancedSecurityAlertType = "trust_rate_violation"
	AdvancedSecurityAlertTypeVerificationFailure AdvancedSecurityAlertType = "verification_failure"
	AdvancedSecurityAlertTypeSecurityViolation   AdvancedSecurityAlertType = "security_violation"
	AdvancedSecurityAlertTypeConfidenceIntegrity AdvancedSecurityAlertType = "confidence_integrity"
	AdvancedSecurityAlertTypeSystemSecurity      AdvancedSecurityAlertType = "system_security"
)

// AlertType represents the type of alert
type AlertType string

const (
	AlertTypeTrustRate    AlertType = "trust_rate"
	AlertTypeVerification AlertType = "verification"
	AlertTypeViolation    AlertType = "violation"
	AlertTypeConfidence   AlertType = "confidence"
)

// AlertSeverity represents the severity of an alert
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AdvancedSecurityMetrics represents comprehensive security metrics
type AdvancedSecurityMetrics struct {
	// Data source trust metrics
	OverallTrustRate    float64            `json:"overall_trust_rate"`
	TrustRateBySource   map[string]float64 `json:"trust_rate_by_source"`
	TrustRateTarget     float64            `json:"trust_rate_target"`
	TrustRateCompliance bool               `json:"trust_rate_compliance"`

	// Website verification metrics
	OverallVerificationRate    float64            `json:"overall_verification_rate"`
	VerificationRateByDomain   map[string]float64 `json:"verification_rate_by_domain"`
	VerificationRateTarget     float64            `json:"verification_rate_target"`
	VerificationRateCompliance bool               `json:"verification_rate_compliance"`

	// Security violation metrics
	TotalViolations         int64            `json:"total_violations"`
	ViolationsByType        map[string]int64 `json:"violations_by_type"`
	ViolationsBySeverity    map[string]int64 `json:"violations_by_severity"`
	ViolationsPerHour       float64          `json:"violations_per_hour"`
	ViolationRateCompliance bool             `json:"violation_rate_compliance"`

	// Confidence integrity metrics
	TotalConfidenceEvents         int64            `json:"total_confidence_events"`
	ConfidenceEventsByType        map[string]int64 `json:"confidence_events_by_type"`
	ConfidenceEventsBySeverity    map[string]int64 `json:"confidence_events_by_severity"`
	AverageScoreDifference        float64          `json:"average_score_difference"`
	ConfidenceIntegrityCompliance bool             `json:"confidence_integrity_compliance"`

	// Overall security metrics
	OverallSecurityScore float64   `json:"overall_security_score"`
	SecurityCompliance   bool      `json:"security_compliance"`
	LastUpdated          time.Time `json:"last_updated"`
}

// NewAdvancedSecurityMonitor creates a new advanced security monitor
func NewAdvancedSecurityMonitor(config *AdvancedSecurityMonitorConfig, logger *observability.Logger, tracer trace.Tracer) *AdvancedSecurityMonitor {
	if config == nil {
		config = &AdvancedSecurityMonitorConfig{
			TrustRateTarget:         100.0, // 100% target
			TrustRateCheckInterval:  30 * time.Second,
			TrustRateHistorySize:    1000,
			TrustRateAlertThreshold: 95.0, // Alert if below 95%

			VerificationRateTarget:     95.0, // 95% target
			VerificationCheckInterval:  30 * time.Second,
			VerificationHistorySize:    1000,
			VerificationAlertThreshold: 90.0, // Alert if below 90%

			ViolationDetectionEnabled: true,
			ViolationCheckInterval:    1 * time.Minute,
			ViolationHistorySize:      1000,
			MaxViolationsPerHour:      10, // Max 10 violations per hour

			ConfidenceIntegrityEnabled: true,
			ConfidenceCheckInterval:    1 * time.Minute,
			ConfidenceHistorySize:      1000,
			ConfidenceAnomalyThreshold: 0.1, // 10% difference threshold

			AlertingEnabled:  true,
			AlertCooldown:    5 * time.Minute,
			AlertHistorySize: 1000,
			WebhookTimeout:   10 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	asm := &AdvancedSecurityMonitor{
		config:            config,
		logger:            logger,
		tracer:            tracer,
		trustRates:        make(map[string]*TrustRateData),
		verificationRates: make(map[string]*VerificationRateData),
		violationEvents:   make([]*SecurityViolationEvent, 0),
		confidenceEvents:  make([]*ConfidenceIntegrityEvent, 0),
		ctx:               ctx,
		cancel:            cancel,
	}

	// Initialize monitoring components
	asm.dataSourceTrustMonitor = &DataSourceTrustMonitor{
		TrustRates: make(map[string]*TrustRateData),
	}
	asm.websiteVerificationMonitor = &WebsiteVerificationMonitor{
		VerificationRates: make(map[string]*VerificationRateData),
	}
	asm.securityViolationMonitor = &SecurityViolationMonitor{
		ViolationEvents: make([]*SecurityViolationEvent, 0),
	}
	asm.confidenceIntegrityMonitor = &ConfidenceIntegrityMonitor{
		ConfidenceEvents: make([]*ConfidenceIntegrityEvent, 0),
	}
	asm.alertManager = &SecurityAlertManager{
		Alerts: make([]*AdvancedSecurityAlert, 0),
	}

	// Initialize security metrics
	asm.securityMetrics = &AdvancedSecurityMetrics{
		TrustRateBySource:          make(map[string]float64),
		VerificationRateByDomain:   make(map[string]float64),
		ViolationsByType:           make(map[string]int64),
		ViolationsBySeverity:       make(map[string]int64),
		ConfidenceEventsByType:     make(map[string]int64),
		ConfidenceEventsBySeverity: make(map[string]int64),
		TrustRateTarget:            config.TrustRateTarget,
		VerificationRateTarget:     config.VerificationRateTarget,
	}

	// Start background monitoring workers
	asm.startBackgroundWorkers()

	return asm
}

// startBackgroundWorkers starts background monitoring workers
func (asm *AdvancedSecurityMonitor) startBackgroundWorkers() {
	// Data source trust monitoring worker
	go asm.dataSourceTrustMonitoringWorker()

	// Website verification monitoring worker
	go asm.websiteVerificationMonitoringWorker()

	// Security violation monitoring worker
	go asm.securityViolationMonitoringWorker()

	// Confidence integrity monitoring worker
	go asm.confidenceIntegrityMonitoringWorker()

	// Metrics update worker
	go asm.metricsUpdateWorker()

	// Alert processing worker
	go asm.alertProcessingWorker()
}

// dataSourceTrustMonitoringWorker monitors data source trust rates
func (asm *AdvancedSecurityMonitor) dataSourceTrustMonitoringWorker() {
	ticker := time.NewTicker(asm.config.TrustRateCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-asm.ctx.Done():
			return
		case <-ticker.C:
			asm.monitorDataSourceTrust()
		}
	}
}

// websiteVerificationMonitoringWorker monitors website verification rates
func (asm *AdvancedSecurityMonitor) websiteVerificationMonitoringWorker() {
	ticker := time.NewTicker(asm.config.VerificationCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-asm.ctx.Done():
			return
		case <-ticker.C:
			asm.monitorWebsiteVerification()
		}
	}
}

// securityViolationMonitoringWorker monitors security violations
func (asm *AdvancedSecurityMonitor) securityViolationMonitoringWorker() {
	ticker := time.NewTicker(asm.config.ViolationCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-asm.ctx.Done():
			return
		case <-ticker.C:
			asm.monitorSecurityViolations()
		}
	}
}

// confidenceIntegrityMonitoringWorker monitors confidence score integrity
func (asm *AdvancedSecurityMonitor) confidenceIntegrityMonitoringWorker() {
	ticker := time.NewTicker(asm.config.ConfidenceCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-asm.ctx.Done():
			return
		case <-ticker.C:
			asm.monitorConfidenceIntegrity()
		}
	}
}

// metricsUpdateWorker updates security metrics
func (asm *AdvancedSecurityMonitor) metricsUpdateWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-asm.ctx.Done():
			return
		case <-ticker.C:
			asm.updateSecurityMetrics()
		}
	}
}

// alertProcessingWorker processes security alerts
func (asm *AdvancedSecurityMonitor) alertProcessingWorker() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-asm.ctx.Done():
			return
		case <-ticker.C:
			asm.processSecurityAlerts()
		}
	}
}

// RecordDataSourceRequest records a data source request for trust monitoring
func (asm *AdvancedSecurityMonitor) RecordDataSourceRequest(dataSourceID, dataSourceName string, trusted bool) {
	_, span := asm.tracer.Start(asm.ctx, "AdvancedSecurityMonitor.RecordDataSourceRequest")
	defer span.End()

	asm.trustMux.Lock()
	defer asm.trustMux.Unlock()

	trustData, exists := asm.trustRates[dataSourceID]
	if !exists {
		trustData = &TrustRateData{
			DataSourceID:   dataSourceID,
			DataSourceName: dataSourceName,
			History:        make([]TrustRatePoint, 0),
			Alerts:         make([]TrustAlert, 0),
		}
		asm.trustRates[dataSourceID] = trustData
	}

	// Update request counts
	trustData.TotalRequests++
	if trusted {
		trustData.TrustedRequests++
	} else {
		trustData.UntrustedRequests++
	}

	// Calculate trust rate
	if trustData.TotalRequests > 0 {
		trustData.TrustRate = float64(trustData.TrustedRequests) / float64(trustData.TotalRequests) * 100.0
	}

	// Update status
	trustData.Status = asm.calculateTrustStatus(trustData.TrustRate)
	trustData.LastUpdated = time.Now()

	// Add to history
	trustData.History = append(trustData.History, TrustRatePoint{
		TrustRate:    trustData.TrustRate,
		Timestamp:    trustData.LastUpdated,
		RequestCount: trustData.TotalRequests,
	})

	// Maintain history size
	if len(trustData.History) > asm.config.TrustRateHistorySize {
		trustData.History = trustData.History[1:]
	}

	// Check for alerts
	if trustData.TrustRate < asm.config.TrustRateAlertThreshold {
		asm.createTrustRateAlert(trustData)
	}

	span.SetAttributes(
		attribute.String("data_source_id", dataSourceID),
		attribute.Bool("trusted", trusted),
		attribute.Float64("trust_rate", trustData.TrustRate),
		attribute.String("status", string(trustData.Status)),
	)

	asm.logger.Info("data source request recorded",
		map[string]interface{}{
			"data_source_id":   dataSourceID,
			"data_source_name": dataSourceName,
			"trusted":          trusted,
			"trust_rate":       trustData.TrustRate,
			"total_requests":   trustData.TotalRequests,
		})
}

// RecordWebsiteVerification records a website verification attempt
func (asm *AdvancedSecurityMonitor) RecordWebsiteVerification(domain string, successful bool) {
	_, span := asm.tracer.Start(asm.ctx, "AdvancedSecurityMonitor.RecordWebsiteVerification")
	defer span.End()

	asm.verificationMux.Lock()
	defer asm.verificationMux.Unlock()

	verificationData, exists := asm.verificationRates[domain]
	if !exists {
		verificationData = &VerificationRateData{
			WebsiteDomain: domain,
			History:       make([]VerificationRatePoint, 0),
			Alerts:        make([]VerificationAlert, 0),
		}
		asm.verificationRates[domain] = verificationData
	}

	// Update attempt counts
	verificationData.TotalAttempts++
	if successful {
		verificationData.SuccessfulVerifications++
	} else {
		verificationData.FailedVerifications++
	}

	// Calculate verification rate
	if verificationData.TotalAttempts > 0 {
		verificationData.VerificationRate = float64(verificationData.SuccessfulVerifications) / float64(verificationData.TotalAttempts) * 100.0
	}

	// Update status
	verificationData.Status = asm.calculateVerificationStatus(verificationData.VerificationRate)
	verificationData.LastUpdated = time.Now()

	// Add to history
	verificationData.History = append(verificationData.History, VerificationRatePoint{
		VerificationRate: verificationData.VerificationRate,
		Timestamp:        verificationData.LastUpdated,
		AttemptCount:     verificationData.TotalAttempts,
	})

	// Maintain history size
	if len(verificationData.History) > asm.config.VerificationHistorySize {
		verificationData.History = verificationData.History[1:]
	}

	// Check for alerts
	if verificationData.VerificationRate < asm.config.VerificationAlertThreshold {
		asm.createVerificationAlert(verificationData)
	}

	span.SetAttributes(
		attribute.String("domain", domain),
		attribute.Bool("successful", successful),
		attribute.Float64("verification_rate", verificationData.VerificationRate),
		attribute.String("status", string(verificationData.Status)),
	)

	asm.logger.Info("website verification recorded",
		map[string]interface{}{
			"domain":            domain,
			"successful":        successful,
			"verification_rate": verificationData.VerificationRate,
			"total_attempts":    verificationData.TotalAttempts,
		})
}

// RecordSecurityViolation records a security violation
func (asm *AdvancedSecurityMonitor) RecordSecurityViolation(violationType ViolationType, severity ViolationSeverity, source, description string, details map[string]interface{}) {
	_, span := asm.tracer.Start(asm.ctx, "AdvancedSecurityMonitor.RecordSecurityViolation")
	defer span.End()

	asm.violationMux.Lock()
	defer asm.violationMux.Unlock()

	violation := &SecurityViolationEvent{
		ID:          generateViolationID(),
		Type:        violationType,
		Severity:    severity,
		Source:      source,
		Description: description,
		Details:     details,
		Timestamp:   time.Now(),
		Resolved:    false,
	}

	asm.violationEvents = append(asm.violationEvents, violation)

	// Maintain history size
	if len(asm.violationEvents) > asm.config.ViolationHistorySize {
		asm.violationEvents = asm.violationEvents[1:]
	}

	span.SetAttributes(
		attribute.String("violation_id", violation.ID),
		attribute.String("type", string(violationType)),
		attribute.String("severity", string(severity)),
		attribute.String("source", source),
	)

	asm.logger.Warn("security violation recorded",
		map[string]interface{}{
			"violation_id": violation.ID,
			"type":         violationType,
			"severity":     severity,
			"source":       source,
			"description":  description,
		})
}

// RecordConfidenceIntegrityEvent records a confidence score integrity event
func (asm *AdvancedSecurityMonitor) RecordConfidenceIntegrityEvent(eventType ConfidenceEventType, severity ConfidenceEventSeverity, classificationID string, expectedScore, actualScore float64, details map[string]interface{}) {
	_, span := asm.tracer.Start(asm.ctx, "AdvancedSecurityMonitor.RecordConfidenceIntegrityEvent")
	defer span.End()

	asm.confidenceMux.Lock()
	defer asm.confidenceMux.Unlock()

	scoreDifference := actualScore - expectedScore
	if scoreDifference < 0 {
		scoreDifference = -scoreDifference
	}

	event := &ConfidenceIntegrityEvent{
		ID:               generateConfidenceEventID(),
		Type:             eventType,
		Severity:         severity,
		ClassificationID: classificationID,
		ExpectedScore:    expectedScore,
		ActualScore:      actualScore,
		ScoreDifference:  scoreDifference,
		Details:          details,
		Timestamp:        time.Now(),
		Resolved:         false,
	}

	asm.confidenceEvents = append(asm.confidenceEvents, event)

	// Maintain history size
	if len(asm.confidenceEvents) > asm.config.ConfidenceHistorySize {
		asm.confidenceEvents = asm.confidenceEvents[1:]
	}

	span.SetAttributes(
		attribute.String("event_id", event.ID),
		attribute.String("type", string(eventType)),
		attribute.String("severity", string(severity)),
		attribute.String("classification_id", classificationID),
		attribute.Float64("expected_score", expectedScore),
		attribute.Float64("actual_score", actualScore),
		attribute.Float64("score_difference", scoreDifference),
	)

	asm.logger.Warn("confidence integrity event recorded",
		map[string]interface{}{
			"event_id":          event.ID,
			"type":              eventType,
			"severity":          severity,
			"classification_id": classificationID,
			"expected_score":    expectedScore,
			"actual_score":      actualScore,
			"score_difference":  scoreDifference,
		})
}

// Helper methods

func (asm *AdvancedSecurityMonitor) calculateTrustStatus(trustRate float64) TrustStatus {
	if trustRate >= 95.0 {
		return TrustStatusExcellent
	} else if trustRate >= 90.0 {
		return TrustStatusGood
	} else if trustRate >= 80.0 {
		return TrustStatusWarning
	} else {
		return TrustStatusCritical
	}
}

func (asm *AdvancedSecurityMonitor) calculateVerificationStatus(verificationRate float64) VerificationStatus {
	if verificationRate >= 95.0 {
		return VerificationStatusExcellent
	} else if verificationRate >= 90.0 {
		return VerificationStatusGood
	} else if verificationRate >= 80.0 {
		return VerificationStatusWarning
	} else {
		return VerificationStatusCritical
	}
}

func (asm *AdvancedSecurityMonitor) createTrustRateAlert(trustData *TrustRateData) {
	alert := TrustAlert{
		ID:           generateTrustAlertID(),
		DataSourceID: trustData.DataSourceID,
		Type:         AlertTypeTrustRate,
		Severity:     asm.calculateAlertSeverity(trustData.TrustRate, asm.config.TrustRateAlertThreshold),
		Message:      fmt.Sprintf("Trust rate for data source %s is %.2f%%, below threshold of %.2f%%", trustData.DataSourceName, trustData.TrustRate, asm.config.TrustRateAlertThreshold),
		TrustRate:    trustData.TrustRate,
		Threshold:    asm.config.TrustRateAlertThreshold,
		Timestamp:    time.Now(),
		Resolved:     false,
	}

	trustData.Alerts = append(trustData.Alerts, alert)

	asm.logger.Warn("trust rate alert created",
		map[string]interface{}{
			"alert_id":       alert.ID,
			"data_source_id": trustData.DataSourceID,
			"trust_rate":     trustData.TrustRate,
			"threshold":      asm.config.TrustRateAlertThreshold,
		})
}

func (asm *AdvancedSecurityMonitor) createVerificationAlert(verificationData *VerificationRateData) {
	alert := VerificationAlert{
		ID:               generateVerificationAlertID(),
		WebsiteDomain:    verificationData.WebsiteDomain,
		Type:             AlertTypeVerification,
		Severity:         asm.calculateAlertSeverity(verificationData.VerificationRate, asm.config.VerificationAlertThreshold),
		Message:          fmt.Sprintf("Verification rate for domain %s is %.2f%%, below threshold of %.2f%%", verificationData.WebsiteDomain, verificationData.VerificationRate, asm.config.VerificationAlertThreshold),
		VerificationRate: verificationData.VerificationRate,
		Threshold:        asm.config.VerificationAlertThreshold,
		Timestamp:        time.Now(),
		Resolved:         false,
	}

	verificationData.Alerts = append(verificationData.Alerts, alert)

	asm.logger.Warn("verification alert created",
		map[string]interface{}{
			"alert_id":          alert.ID,
			"website_domain":    verificationData.WebsiteDomain,
			"verification_rate": verificationData.VerificationRate,
			"threshold":         asm.config.VerificationAlertThreshold,
		})
}

func (asm *AdvancedSecurityMonitor) calculateAlertSeverity(rate, threshold float64) AlertSeverity {
	if rate < threshold*0.5 {
		return AlertSeverityCritical
	} else if rate < threshold*0.8 {
		return AlertSeverityError
	} else if rate < threshold {
		return AlertSeverityWarning
	} else {
		return AlertSeverityInfo
	}
}

// Monitoring worker implementations (simplified for now)
func (asm *AdvancedSecurityMonitor) monitorDataSourceTrust() {
	// Implementation would check actual data source trust rates
	// For now, this is a placeholder
}

func (asm *AdvancedSecurityMonitor) monitorWebsiteVerification() {
	// Implementation would check actual website verification rates
	// For now, this is a placeholder
}

func (asm *AdvancedSecurityMonitor) monitorSecurityViolations() {
	// Implementation would check for security violations
	// For now, this is a placeholder
}

func (asm *AdvancedSecurityMonitor) monitorConfidenceIntegrity() {
	// Implementation would check confidence score integrity
	// For now, this is a placeholder
}

func (asm *AdvancedSecurityMonitor) updateSecurityMetrics() {
	asm.metricsMux.Lock()
	defer asm.metricsMux.Unlock()

	// Update overall trust rate
	totalTrusted := int64(0)
	totalRequests := int64(0)
	asm.trustMux.RLock()
	for _, trustData := range asm.trustRates {
		totalTrusted += trustData.TrustedRequests
		totalRequests += trustData.TotalRequests
		asm.securityMetrics.TrustRateBySource[trustData.DataSourceID] = trustData.TrustRate
	}
	asm.trustMux.RUnlock()

	if totalRequests > 0 {
		asm.securityMetrics.OverallTrustRate = float64(totalTrusted) / float64(totalRequests) * 100.0
	}

	// Update overall verification rate
	totalSuccessful := int64(0)
	totalAttempts := int64(0)
	asm.verificationMux.RLock()
	for _, verificationData := range asm.verificationRates {
		totalSuccessful += verificationData.SuccessfulVerifications
		totalAttempts += verificationData.TotalAttempts
		asm.securityMetrics.VerificationRateByDomain[verificationData.WebsiteDomain] = verificationData.VerificationRate
	}
	asm.verificationMux.RUnlock()

	if totalAttempts > 0 {
		asm.securityMetrics.OverallVerificationRate = float64(totalSuccessful) / float64(totalAttempts) * 100.0
	}

	// Update violation metrics
	asm.violationMux.RLock()
	asm.securityMetrics.TotalViolations = int64(len(asm.violationEvents))
	for _, violation := range asm.violationEvents {
		asm.securityMetrics.ViolationsByType[string(violation.Type)]++
		asm.securityMetrics.ViolationsBySeverity[string(violation.Severity)]++
	}
	asm.violationMux.RUnlock()

	// Update confidence integrity metrics
	asm.confidenceMux.RLock()
	asm.securityMetrics.TotalConfidenceEvents = int64(len(asm.confidenceEvents))
	totalScoreDifference := 0.0
	for _, event := range asm.confidenceEvents {
		asm.securityMetrics.ConfidenceEventsByType[string(event.Type)]++
		asm.securityMetrics.ConfidenceEventsBySeverity[string(event.Severity)]++
		totalScoreDifference += event.ScoreDifference
	}
	asm.confidenceMux.RUnlock()

	if asm.securityMetrics.TotalConfidenceEvents > 0 {
		asm.securityMetrics.AverageScoreDifference = totalScoreDifference / float64(asm.securityMetrics.TotalConfidenceEvents)
	}

	// Calculate compliance
	asm.securityMetrics.TrustRateCompliance = asm.securityMetrics.OverallTrustRate >= asm.config.TrustRateTarget
	asm.securityMetrics.VerificationRateCompliance = asm.securityMetrics.OverallVerificationRate >= asm.config.VerificationRateTarget
	asm.securityMetrics.ViolationRateCompliance = asm.securityMetrics.TotalViolations < int64(asm.config.MaxViolationsPerHour)
	asm.securityMetrics.ConfidenceIntegrityCompliance = asm.securityMetrics.AverageScoreDifference < asm.config.ConfidenceAnomalyThreshold

	// Calculate overall security score
	asm.calculateOverallSecurityScore()

	asm.securityMetrics.LastUpdated = time.Now()
}

func (asm *AdvancedSecurityMonitor) calculateOverallSecurityScore() {
	score := 0.0
	factors := 0

	// Trust rate factor (25%)
	if asm.securityMetrics.TrustRateCompliance {
		score += 25.0
	} else {
		score += (asm.securityMetrics.OverallTrustRate / asm.config.TrustRateTarget) * 25.0
	}
	factors++

	// Verification rate factor (25%)
	if asm.securityMetrics.VerificationRateCompliance {
		score += 25.0
	} else {
		score += (asm.securityMetrics.OverallVerificationRate / asm.config.VerificationRateTarget) * 25.0
	}
	factors++

	// Violation rate factor (25%)
	if asm.securityMetrics.ViolationRateCompliance {
		score += 25.0
	} else {
		violationRate := float64(asm.securityMetrics.TotalViolations) / float64(asm.config.MaxViolationsPerHour)
		if violationRate > 1.0 {
			score += 0.0
		} else {
			score += (1.0 - violationRate) * 25.0
		}
	}
	factors++

	// Confidence integrity factor (25%)
	if asm.securityMetrics.ConfidenceIntegrityCompliance {
		score += 25.0
	} else {
		integrityRate := 1.0 - (asm.securityMetrics.AverageScoreDifference / asm.config.ConfidenceAnomalyThreshold)
		if integrityRate < 0.0 {
			integrityRate = 0.0
		}
		score += integrityRate * 25.0
	}
	factors++

	asm.securityMetrics.OverallSecurityScore = score
	asm.securityMetrics.SecurityCompliance = score >= 90.0 // 90% threshold for compliance
}

func (asm *AdvancedSecurityMonitor) processSecurityAlerts() {
	// Implementation would process and send security alerts
	// For now, this is a placeholder
}

// Getter methods

func (asm *AdvancedSecurityMonitor) GetTrustRates() map[string]*TrustRateData {
	asm.trustMux.RLock()
	defer asm.trustMux.RUnlock()
	return asm.trustRates
}

func (asm *AdvancedSecurityMonitor) GetVerificationRates() map[string]*VerificationRateData {
	asm.verificationMux.RLock()
	defer asm.verificationMux.RUnlock()
	return asm.verificationRates
}

func (asm *AdvancedSecurityMonitor) GetViolationEvents() []*SecurityViolationEvent {
	asm.violationMux.RLock()
	defer asm.violationMux.RUnlock()
	return asm.violationEvents
}

func (asm *AdvancedSecurityMonitor) GetConfidenceEvents() []*ConfidenceIntegrityEvent {
	asm.confidenceMux.RLock()
	defer asm.confidenceMux.RUnlock()
	return asm.confidenceEvents
}

func (asm *AdvancedSecurityMonitor) GetSecurityMetrics() *AdvancedSecurityMetrics {
	asm.metricsMux.RLock()
	defer asm.metricsMux.RUnlock()
	return asm.securityMetrics
}

// UpdateSecurityMetrics forces an update of security metrics (useful for testing)
func (asm *AdvancedSecurityMonitor) UpdateSecurityMetrics() {
	asm.updateSecurityMetrics()
}

func (asm *AdvancedSecurityMonitor) GetAlerts() []*AdvancedSecurityAlert {
	asm.alertManager.Mux.RLock()
	defer asm.alertManager.Mux.RUnlock()
	return asm.alertManager.Alerts
}

// Shutdown shuts down the advanced security monitor
func (asm *AdvancedSecurityMonitor) Shutdown() {
	asm.cancel()
	asm.logger.Info("advanced security monitor shutting down", map[string]interface{}{})
}

// Helper functions for generating IDs
func generateViolationID() string {
	return fmt.Sprintf("viol_%d", time.Now().UnixNano())
}

func generateConfidenceEventID() string {
	return fmt.Sprintf("conf_%d", time.Now().UnixNano())
}

func generateTrustAlertID() string {
	return fmt.Sprintf("trust_alert_%d", time.Now().UnixNano())
}

func generateVerificationAlertID() string {
	return fmt.Sprintf("verif_alert_%d", time.Now().UnixNano())
}
