package security

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// GDPRComplianceService provides comprehensive GDPR compliance functionality
type GDPRComplianceService struct {
	config         *GDPRComplianceConfig
	logger         *zap.Logger
	consentManager *ConsentManager
	rightsManager  *DataSubjectRightsManager
	breachDetector *BreachDetector
	processor      *DataProcessor
}

// GDPRComplianceConfig contains GDPR compliance configuration
type GDPRComplianceConfig struct {
	// General GDPR settings
	EnableGDPRCompliance bool   `json:"enable_gdpr_compliance" yaml:"enable_gdpr_compliance"`
	DataControllerName   string `json:"data_controller_name" yaml:"data_controller_name"`
	DataControllerEmail  string `json:"data_controller_email" yaml:"data_controller_email"`
	DataControllerPhone  string `json:"data_controller_phone" yaml:"data_controller_phone"`

	// Consent management
	RequireExplicitConsent bool          `json:"require_explicit_consent" yaml:"require_explicit_consent"`
	ConsentExpiryPeriod    time.Duration `json:"consent_expiry_period" yaml:"consent_expiry_period"`
	AllowWithdrawal        bool          `json:"allow_withdrawal" yaml:"allow_withdrawal"`

	// Data subject rights
	EnableDataSubjectRights bool          `json:"enable_data_subject_rights" yaml:"enable_data_subject_rights"`
	ResponseTimeLimit       time.Duration `json:"response_time_limit" yaml:"response_time_limit"`
	MaxDataPortabilitySize  int64         `json:"max_data_portability_size" yaml:"max_data_portability_size"`

	// Breach notification
	EnableBreachDetection       bool          `json:"enable_breach_detection" yaml:"enable_breach_detection"`
	BreachNotificationThreshold int           `json:"breach_notification_threshold" yaml:"breach_notification_threshold"`
	NotificationTimeLimit       time.Duration `json:"notification_time_limit" yaml:"notification_time_limit"`

	// Data processing records
	EnableProcessingRecords bool          `json:"enable_processing_records" yaml:"enable_processing_records"`
	RecordRetentionPeriod   time.Duration `json:"record_retention_period" yaml:"record_retention_period"`

	// Legal basis
	DefaultLegalBasis string   `json:"default_legal_basis" yaml:"default_legal_basis"`
	AllowedLegalBases []string `json:"allowed_legal_bases" yaml:"allowed_legal_bases"`
}

// Consent represents user consent for data processing
type Consent struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	ConsentType    string                 `json:"consent_type"` // "explicit", "implicit", "withdrawn"
	LegalBasis     string                 `json:"legal_basis"`
	Purpose        string                 `json:"purpose"`
	DataCategories []string               `json:"data_categories"`
	ThirdParties   []string               `json:"third_parties,omitempty"`
	Granular       bool                   `json:"granular"`
	Withdrawable   bool                   `json:"withdrawable"`
	GivenAt        time.Time              `json:"given_at"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
	WithdrawnAt    *time.Time             `json:"withdrawn_at,omitempty"`
	Evidence       map[string]interface{} `json:"evidence,omitempty"`
	IPAddress      string                 `json:"ip_address,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty"`
}

// DataSubjectRequest represents a data subject rights request
type DataSubjectRequest struct {
	ID                 string                 `json:"id"`
	UserID             string                 `json:"user_id"`
	RequestType        string                 `json:"request_type"` // "access", "rectification", "erasure", "portability", "restriction"
	Status             string                 `json:"status"`       // "pending", "processing", "completed", "rejected"
	Description        string                 `json:"description"`
	RequestedAt        time.Time              `json:"requested_at"`
	CompletedAt        *time.Time             `json:"completed_at,omitempty"`
	ResponseData       map[string]interface{} `json:"response_data,omitempty"`
	RejectionReason    string                 `json:"rejection_reason,omitempty"`
	VerificationMethod string                 `json:"verification_method,omitempty"`
}

// DataBreach represents a detected data breach
type DataBreach struct {
	ID                 string     `json:"id"`
	BreachType         string     `json:"breach_type"` // "unauthorized_access", "data_loss", "system_breach"
	Severity           string     `json:"severity"`    // "low", "medium", "high", "critical"
	Description        string     `json:"description"`
	DetectedAt         time.Time  `json:"detected_at"`
	ReportedAt         *time.Time `json:"reported_at,omitempty"`
	AffectedRecords    int        `json:"affected_records"`
	AffectedUsers      []string   `json:"affected_users,omitempty"`
	DataCategories     []string   `json:"data_categories"`
	RootCause          string     `json:"root_cause,omitempty"`
	MitigationSteps    []string   `json:"mitigation_steps,omitempty"`
	NotificationSent   bool       `json:"notification_sent"`
	RegulatoryReported bool       `json:"regulatory_reported"`
}

// DataProcessingRecord represents a record of data processing activity
type DataProcessingRecord struct {
	ID                string                 `json:"id"`
	ProcessingPurpose string                 `json:"processing_purpose"`
	LegalBasis        string                 `json:"legal_basis"`
	DataCategories    []string               `json:"data_categories"`
	DataSubjects      []string               `json:"data_subjects"`
	Recipients        []string               `json:"recipients,omitempty"`
	ThirdCountries    []string               `json:"third_countries,omitempty"`
	RetentionPeriod   time.Duration          `json:"retention_period"`
	SecurityMeasures  []string               `json:"security_measures"`
	ProcessingAt      time.Time              `json:"processing_at"`
	DataController    string                 `json:"data_controller"`
	DataProcessor     string                 `json:"data_processor,omitempty"`
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// GDPRComplianceResult represents the result of GDPR compliance validation
type GDPRComplianceResult struct {
	IsCompliant     bool                   `json:"is_compliant"`
	ComplianceScore float64                `json:"compliance_score"`
	Violations      []GDPRViolation        `json:"violations,omitempty"`
	Warnings        []GDPRWarning          `json:"warnings,omitempty"`
	Recommendations []string               `json:"recommendations,omitempty"`
	ValidationTime  time.Duration          `json:"validation_time"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// GDPRViolation represents a GDPR compliance violation
type GDPRViolation struct {
	Type           string `json:"type"`
	Severity       string `json:"severity"` // "low", "medium", "high", "critical"
	Article        string `json:"article"`  // GDPR article number
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
	Penalty        string `json:"penalty,omitempty"`
}

// GDPRWarning represents a GDPR compliance warning
type GDPRWarning struct {
	Type           string `json:"type"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// ConsentManager handles consent management
type ConsentManager struct {
	config *GDPRComplianceConfig
	logger *zap.Logger
}

// DataSubjectRightsManager handles data subject rights requests
type DataSubjectRightsManager struct {
	config *GDPRComplianceConfig
	logger *zap.Logger
}

// BreachDetector handles data breach detection and notification
type BreachDetector struct {
	config *GDPRComplianceConfig
	logger *zap.Logger
}

// DataProcessor handles data processing records
type DataProcessor struct {
	config *GDPRComplianceConfig
	logger *zap.Logger
}

// NewGDPRComplianceService creates a new GDPR compliance service
func NewGDPRComplianceService(config *GDPRComplianceConfig, logger *zap.Logger) (*GDPRComplianceService, error) {
	if config == nil {
		config = &GDPRComplianceConfig{
			EnableGDPRCompliance:        true,
			DataControllerName:          "KYB Platform",
			DataControllerEmail:         "privacy@kyb-platform.com",
			RequireExplicitConsent:      true,
			ConsentExpiryPeriod:         365 * 24 * time.Hour, // 1 year
			AllowWithdrawal:             true,
			EnableDataSubjectRights:     true,
			ResponseTimeLimit:           30 * 24 * time.Hour, // 30 days
			MaxDataPortabilitySize:      10 * 1024 * 1024,    // 10MB
			EnableBreachDetection:       true,
			BreachNotificationThreshold: 100,
			NotificationTimeLimit:       72 * time.Hour, // 72 hours
			EnableProcessingRecords:     true,
			RecordRetentionPeriod:       5 * 365 * 24 * time.Hour, // 5 years
			DefaultLegalBasis:           "legitimate_interest",
			AllowedLegalBases: []string{
				"consent",
				"contract",
				"legal_obligation",
				"vital_interests",
				"public_task",
				"legitimate_interest",
			},
		}
	}

	consentManager := &ConsentManager{
		config: config,
		logger: logger,
	}

	rightsManager := &DataSubjectRightsManager{
		config: config,
		logger: logger,
	}

	breachDetector := &BreachDetector{
		config: config,
		logger: logger,
	}

	processor := &DataProcessor{
		config: config,
		logger: logger,
	}

	return &GDPRComplianceService{
		config:         config,
		logger:         logger,
		consentManager: consentManager,
		rightsManager:  rightsManager,
		breachDetector: breachDetector,
		processor:      processor,
	}, nil
}

// ValidateGDPRCompliance validates GDPR compliance for data processing
func (gcs *GDPRComplianceService) ValidateGDPRCompliance(ctx context.Context, data map[string]interface{}, processingPurpose string) (*GDPRComplianceResult, error) {
	startTime := time.Now()

	if !gcs.config.EnableGDPRCompliance {
		return &GDPRComplianceResult{
			IsCompliant:     true,
			ComplianceScore: 100.0,
			ValidationTime:  time.Since(startTime),
			Metadata: map[string]interface{}{
				"validation_timestamp":     startTime,
				"gdpr_compliance_disabled": true,
			},
		}, nil
	}

	var violations []GDPRViolation
	var warnings []GDPRWarning

	// Check legal basis
	if err := gcs.validateLegalBasis(ctx, data, processingPurpose); err != nil {
		violations = append(violations, GDPRViolation{
			Type:           "invalid_legal_basis",
			Severity:       "high",
			Article:        "6",
			Description:    fmt.Sprintf("Invalid or missing legal basis: %v", err),
			Recommendation: "Ensure valid legal basis is documented for data processing",
			Penalty:        "Up to €20 million or 4% of global annual turnover",
		})
	}

	// Check consent requirements
	if gcs.config.RequireExplicitConsent {
		if err := gcs.validateConsent(ctx, data); err != nil {
			violations = append(violations, GDPRViolation{
				Type:           "missing_consent",
				Severity:       "high",
				Article:        "7",
				Description:    fmt.Sprintf("Missing or invalid consent: %v", err),
				Recommendation: "Obtain explicit consent before processing personal data",
				Penalty:        "Up to €20 million or 4% of global annual turnover",
			})
		}
	}

	// Check data minimization
	if err := gcs.validateDataMinimization(ctx, data, processingPurpose); err != nil {
		violations = append(violations, GDPRViolation{
			Type:           "data_minimization_violation",
			Severity:       "medium",
			Article:        "5(1)(c)",
			Description:    fmt.Sprintf("Data minimization principle violated: %v", err),
			Recommendation: "Only collect data that is adequate, relevant, and limited to what is necessary",
		})
	}

	// Check purpose limitation
	if err := gcs.validatePurposeLimitation(ctx, data, processingPurpose); err != nil {
		violations = append(violations, GDPRViolation{
			Type:           "purpose_limitation_violation",
			Severity:       "medium",
			Article:        "5(1)(b)",
			Description:    fmt.Sprintf("Purpose limitation principle violated: %v", err),
			Recommendation: "Ensure data is collected for specified, explicit, and legitimate purposes",
		})
	}

	// Check data retention
	if err := gcs.validateDataRetention(ctx, data); err != nil {
		warnings = append(warnings, GDPRWarning{
			Type:           "retention_period_warning",
			Description:    fmt.Sprintf("Data retention period may be excessive: %v", err),
			Recommendation: "Review and justify data retention periods",
		})
	}

	// Check security measures
	if err := gcs.validateSecurityMeasures(ctx, data); err != nil {
		violations = append(violations, GDPRViolation{
			Type:           "inadequate_security",
			Severity:       "high",
			Article:        "32",
			Description:    fmt.Sprintf("Inadequate security measures: %v", err),
			Recommendation: "Implement appropriate technical and organizational security measures",
			Penalty:        "Up to €10 million or 2% of global annual turnover",
		})
	}

	// Calculate compliance score
	complianceScore := gcs.calculateComplianceScore(violations, warnings)
	isCompliant := len(violations) == 0

	// Generate recommendations
	recommendations := gcs.generateRecommendations(violations, warnings)

	result := &GDPRComplianceResult{
		IsCompliant:     isCompliant,
		ComplianceScore: complianceScore,
		Violations:      violations,
		Warnings:        warnings,
		Recommendations: recommendations,
		ValidationTime:  time.Since(startTime),
		Metadata: map[string]interface{}{
			"validation_timestamp": startTime,
			"processing_purpose":   processingPurpose,
			"data_categories":      gcs.extractDataCategories(data),
		},
	}

	gcs.logger.Info("GDPR compliance validation completed",
		zap.Bool("is_compliant", isCompliant),
		zap.Float64("compliance_score", complianceScore),
		zap.Int("violations_count", len(violations)),
		zap.Int("warnings_count", len(warnings)),
		zap.Duration("validation_time", result.ValidationTime))

	return result, nil
}

// RecordConsent records user consent for data processing
func (gcs *GDPRComplianceService) RecordConsent(ctx context.Context, userID, consentType, legalBasis, purpose string, dataCategories []string) (*Consent, error) {
	if !gcs.config.EnableGDPRCompliance {
		return nil, fmt.Errorf("GDPR compliance is disabled")
	}

	consentID, err := gcs.generateConsentID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate consent ID: %w", err)
	}

	expiresAt := time.Now().Add(gcs.config.ConsentExpiryPeriod)

	consent := &Consent{
		ID:             consentID,
		UserID:         userID,
		ConsentType:    consentType,
		LegalBasis:     legalBasis,
		Purpose:        purpose,
		DataCategories: dataCategories,
		Granular:       true,
		Withdrawable:   gcs.config.AllowWithdrawal,
		GivenAt:        time.Now(),
		ExpiresAt:      &expiresAt,
		Evidence: map[string]interface{}{
			"recorded_at": time.Now(),
			"ip_address":  gcs.extractIPAddress(ctx),
			"user_agent":  gcs.extractUserAgent(ctx),
		},
	}

	gcs.logger.Info("Consent recorded",
		zap.String("consent_id", consent.ID),
		zap.String("user_id", consent.UserID),
		zap.String("consent_type", consent.ConsentType),
		zap.String("legal_basis", consent.LegalBasis))

	return consent, nil
}

// WithdrawConsent withdraws user consent
func (gcs *GDPRComplianceService) WithdrawConsent(ctx context.Context, consentID, userID string) error {
	if !gcs.config.EnableGDPRCompliance {
		return fmt.Errorf("GDPR compliance is disabled")
	}

	if !gcs.config.AllowWithdrawal {
		return fmt.Errorf("consent withdrawal is not allowed")
	}

	// In a real implementation, this would update the consent record in the database
	withdrawnAt := time.Now()

	gcs.logger.Info("Consent withdrawn",
		zap.String("consent_id", consentID),
		zap.String("user_id", userID),
		zap.Time("withdrawn_at", withdrawnAt))

	return nil
}

// SubmitDataSubjectRequest submits a data subject rights request
func (gcs *GDPRComplianceService) SubmitDataSubjectRequest(ctx context.Context, userID, requestType, description string) (*DataSubjectRequest, error) {
	if !gcs.config.EnableGDPRCompliance {
		return nil, fmt.Errorf("GDPR compliance is disabled")
	}

	if !gcs.config.EnableDataSubjectRights {
		return nil, fmt.Errorf("data subject rights are disabled")
	}

	requestID, err := gcs.generateRequestID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate request ID: %w", err)
	}

	request := &DataSubjectRequest{
		ID:                 requestID,
		UserID:             userID,
		RequestType:        requestType,
		Status:             "pending",
		Description:        description,
		RequestedAt:        time.Now(),
		VerificationMethod: "email_verification", // Default verification method
	}

	gcs.logger.Info("Data subject request submitted",
		zap.String("request_id", request.ID),
		zap.String("user_id", request.UserID),
		zap.String("request_type", request.RequestType))

	return request, nil
}

// ProcessDataSubjectRequest processes a data subject rights request
func (gcs *GDPRComplianceService) ProcessDataSubjectRequest(ctx context.Context, requestID string) (*DataSubjectRequest, error) {
	if !gcs.config.EnableGDPRCompliance {
		return nil, fmt.Errorf("GDPR compliance is disabled")
	}

	// In a real implementation, this would retrieve the request from the database
	// and process it according to the request type

	completedAt := time.Now()

	// Simulate processing time
	time.Sleep(100 * time.Millisecond)

	gcs.logger.Info("Data subject request processed",
		zap.String("request_id", requestID),
		zap.Time("completed_at", completedAt))

	return &DataSubjectRequest{
		ID:          requestID,
		Status:      "completed",
		CompletedAt: &completedAt,
	}, nil
}

// DetectDataBreach detects and records a data breach
func (gcs *GDPRComplianceService) DetectDataBreach(ctx context.Context, breachType, description string, affectedRecords int, dataCategories []string) (*DataBreach, error) {
	if !gcs.config.EnableGDPRCompliance {
		return nil, fmt.Errorf("GDPR compliance is disabled")
	}

	if !gcs.config.EnableBreachDetection {
		return nil, fmt.Errorf("breach detection is disabled")
	}

	breachID, err := gcs.generateBreachID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate breach ID: %w", err)
	}

	severity := gcs.calculateBreachSeverity(affectedRecords, dataCategories)

	breach := &DataBreach{
		ID:              breachID,
		BreachType:      breachType,
		Severity:        severity,
		Description:     description,
		DetectedAt:      time.Now(),
		AffectedRecords: affectedRecords,
		DataCategories:  dataCategories,
		RootCause:       "system_analysis_required",
		MitigationSteps: []string{
			"Immediate system isolation",
			"Forensic analysis",
			"Notification to affected individuals",
			"Regulatory reporting",
		},
	}

	// Check if notification is required
	if affectedRecords >= gcs.config.BreachNotificationThreshold {
		breach.NotificationSent = true
		breach.RegulatoryReported = true
		breach.ReportedAt = &time.Time{}
		*breach.ReportedAt = time.Now()
	}

	gcs.logger.Warn("Data breach detected",
		zap.String("breach_id", breach.ID),
		zap.String("breach_type", breach.BreachType),
		zap.String("severity", breach.Severity),
		zap.Int("affected_records", breach.AffectedRecords),
		zap.Bool("notification_required", breach.NotificationSent))

	return breach, nil
}

// RecordDataProcessing records data processing activity
func (gcs *GDPRComplianceService) RecordDataProcessing(ctx context.Context, purpose, legalBasis string, dataCategories []string, dataSubjects []string) (*DataProcessingRecord, error) {
	if !gcs.config.EnableGDPRCompliance {
		return nil, fmt.Errorf("GDPR compliance is disabled")
	}

	if !gcs.config.EnableProcessingRecords {
		return nil, fmt.Errorf("processing records are disabled")
	}

	recordID, err := gcs.generateRecordID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate record ID: %w", err)
	}

	record := &DataProcessingRecord{
		ID:                recordID,
		ProcessingPurpose: purpose,
		LegalBasis:        legalBasis,
		DataCategories:    dataCategories,
		DataSubjects:      dataSubjects,
		RetentionPeriod:   gcs.config.RecordRetentionPeriod,
		SecurityMeasures: []string{
			"Encryption at rest",
			"Encryption in transit",
			"Access controls",
			"Audit logging",
		},
		ProcessingAt:   time.Now(),
		DataController: gcs.config.DataControllerName,
		Metadata: map[string]interface{}{
			"recorded_at": time.Now(),
			"ip_address":  gcs.extractIPAddress(ctx),
		},
	}

	gcs.logger.Info("Data processing recorded",
		zap.String("record_id", record.ID),
		zap.String("purpose", record.ProcessingPurpose),
		zap.String("legal_basis", record.LegalBasis),
		zap.Int("data_subjects_count", len(record.DataSubjects)))

	return record, nil
}

// Helper methods for GDPR compliance validation

func (gcs *GDPRComplianceService) validateLegalBasis(ctx context.Context, data map[string]interface{}, purpose string) error {
	// Check if legal basis is documented
	if _, hasLegalBasis := data["legal_basis"]; !hasLegalBasis {
		return fmt.Errorf("legal basis not documented")
	}

	legalBasis := data["legal_basis"].(string)

	// Validate against allowed legal bases
	allowed := false
	for _, allowedBasis := range gcs.config.AllowedLegalBases {
		if legalBasis == allowedBasis {
			allowed = true
			break
		}
	}

	if !allowed {
		return fmt.Errorf("legal basis '%s' is not allowed", legalBasis)
	}

	return nil
}

func (gcs *GDPRComplianceService) validateConsent(ctx context.Context, data map[string]interface{}) error {
	// Check if consent is documented
	if _, hasConsent := data["consent_given"]; !hasConsent {
		return fmt.Errorf("consent not documented")
	}

	consentGiven := data["consent_given"].(bool)
	if !consentGiven {
		return fmt.Errorf("explicit consent not given")
	}

	// Check if consent is recent enough
	if consentDate, hasConsentDate := data["consent_date"]; hasConsentDate {
		if consentTime, ok := consentDate.(time.Time); ok {
			if time.Since(consentTime) > gcs.config.ConsentExpiryPeriod {
				return fmt.Errorf("consent has expired")
			}
		}
	}

	return nil
}

func (gcs *GDPRComplianceService) validateDataMinimization(ctx context.Context, data map[string]interface{}, purpose string) error {
	// Check if data collected is excessive for the stated purpose
	collectedFields := make([]string, 0, len(data))

	for field := range data {
		collectedFields = append(collectedFields, field)
	}

	// Check for potentially excessive data collection
	excessiveFields := gcs.identifyExcessiveFields(collectedFields, purpose)
	if len(excessiveFields) > 0 {
		return fmt.Errorf("excessive data collection detected: %v", excessiveFields)
	}

	return nil
}

func (gcs *GDPRComplianceService) validatePurposeLimitation(ctx context.Context, data map[string]interface{}, purpose string) error {
	// Check if data is being used for the stated purpose
	if originalPurpose, hasOriginalPurpose := data["original_purpose"]; hasOriginalPurpose {
		if originalPurpose != purpose {
			return fmt.Errorf("data being used for different purpose than originally stated")
		}
	}

	return nil
}

func (gcs *GDPRComplianceService) validateDataRetention(ctx context.Context, data map[string]interface{}) error {
	// Check if data retention period is reasonable
	if retentionPeriod, hasRetention := data["retention_period"]; hasRetention {
		if period, ok := retentionPeriod.(time.Duration); ok {
			if period > gcs.config.RecordRetentionPeriod {
				return fmt.Errorf("retention period %v exceeds maximum allowed %v", period, gcs.config.RecordRetentionPeriod)
			}
		}
	}

	return nil
}

func (gcs *GDPRComplianceService) validateSecurityMeasures(ctx context.Context, data map[string]interface{}) error {
	// Check if appropriate security measures are documented
	requiredMeasures := []string{"encryption", "access_controls", "audit_logging"}

	for _, measure := range requiredMeasures {
		if _, hasMeasure := data[measure]; !hasMeasure {
			return fmt.Errorf("security measure '%s' not documented", measure)
		}
	}

	return nil
}

func (gcs *GDPRComplianceService) calculateComplianceScore(violations []GDPRViolation, warnings []GDPRWarning) float64 {
	score := 100.0

	// Deduct points for violations
	for _, violation := range violations {
		switch violation.Severity {
		case "critical":
			score -= 25
		case "high":
			score -= 15
		case "medium":
			score -= 10
		case "low":
			score -= 5
		}
	}

	// Deduct points for warnings
	score -= float64(len(warnings)) * 2

	// Ensure score doesn't go below 0
	if score < 0 {
		score = 0
	}

	return score
}

func (gcs *GDPRComplianceService) generateRecommendations(violations []GDPRViolation, warnings []GDPRWarning) []string {
	var recommendations []string

	// Add violation-specific recommendations
	for _, violation := range violations {
		recommendations = append(recommendations, violation.Recommendation)
	}

	// Add warning-specific recommendations
	for _, warning := range warnings {
		recommendations = append(recommendations, warning.Recommendation)
	}

	// Add general recommendations if violations exist
	if len(violations) > 0 {
		recommendations = append(recommendations,
			"Conduct a comprehensive GDPR compliance audit",
			"Implement regular compliance monitoring and reporting",
			"Provide GDPR training to all staff members",
			"Establish a data protection impact assessment process",
		)
	}

	// Add general recommendations if warnings exist
	if len(warnings) > 0 {
		recommendations = append(recommendations,
			"Review and update privacy policies",
			"Enhance data protection measures",
			"Improve consent management processes",
		)
	}

	return recommendations
}

// Helper methods for data processing

func (gcs *GDPRComplianceService) extractDataCategories(data map[string]interface{}) []string {
	categories := make([]string, 0)

	// Simple categorization based on field names
	for field := range data {
		switch {
		case strings.Contains(strings.ToLower(field), "email"):
			categories = append(categories, "contact_information")
		case strings.Contains(strings.ToLower(field), "phone"):
			categories = append(categories, "contact_information")
		case strings.Contains(strings.ToLower(field), "address"):
			categories = append(categories, "location_data")
		case strings.Contains(strings.ToLower(field), "financial"):
			categories = append(categories, "financial_data")
		case strings.Contains(strings.ToLower(field), "first_name") || strings.Contains(strings.ToLower(field), "last_name"):
			categories = append(categories, "personal_identification")
		case strings.Contains(strings.ToLower(field), "business_name") || strings.Contains(strings.ToLower(field), "industry") || strings.Contains(strings.ToLower(field), "revenue"):
			categories = append(categories, "business_data")
		default:
			// For any other fields, categorize as business_data
			categories = append(categories, "business_data")
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, category := range categories {
		if !seen[category] {
			seen[category] = true
			result = append(result, category)
		}
	}

	return result
}

func (gcs *GDPRComplianceService) getRequiredFieldsForPurpose(purpose string) []string {
	// Define required fields for different purposes
	purposeFields := map[string][]string{
		"business_verification": {"business_name", "business_address", "industry"},
		"risk_assessment":       {"business_name", "financial_data", "compliance_history"},
		"marketing":             {"contact_information", "preferences"},
		"analytics":             {"usage_data", "performance_metrics"},
	}

	if fields, exists := purposeFields[purpose]; exists {
		return fields
	}

	return []string{}
}

func (gcs *GDPRComplianceService) identifyExcessiveFields(fields []string, purpose string) []string {
	requiredFields := gcs.getRequiredFieldsForPurpose(purpose)
	excessive := make([]string, 0)

	// If no required fields are defined for this purpose, don't flag anything as excessive
	if len(requiredFields) == 0 {
		return excessive
	}

	for _, field := range fields {
		required := false
		for _, requiredField := range requiredFields {
			if field == requiredField {
				required = true
				break
			}
		}
		if !required {
			excessive = append(excessive, field)
		}
	}

	return excessive
}

func (gcs *GDPRComplianceService) calculateBreachSeverity(affectedRecords int, dataCategories []string) string {
	// Calculate severity based on affected records and data sensitivity
	severity := "low"

	if affectedRecords >= 10000 {
		severity = "critical"
	} else if affectedRecords >= 1000 {
		severity = "high"
	} else if affectedRecords >= 100 {
		severity = "medium"
	}

	// Adjust severity based on data sensitivity
	for _, category := range dataCategories {
		switch category {
		case "financial_data", "personal_identification":
			if severity == "low" {
				severity = "medium"
			} else if severity == "medium" {
				severity = "high"
			}
		}
	}

	return severity
}

// ID generation helpers

func (gcs *GDPRComplianceService) generateConsentID() (string, error) {
	return gcs.generateID("consent")
}

func (gcs *GDPRComplianceService) generateRequestID() (string, error) {
	return gcs.generateID("dsr")
}

func (gcs *GDPRComplianceService) generateBreachID() (string, error) {
	return gcs.generateID("breach")
}

func (gcs *GDPRComplianceService) generateRecordID() (string, error) {
	return gcs.generateID("record")
}

func (gcs *GDPRComplianceService) generateID(prefix string) (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s_%x", prefix, bytes), nil
}

// Context extraction helpers

func (gcs *GDPRComplianceService) extractIPAddress(ctx context.Context) string {
	// In a real implementation, this would extract IP from context
	return "192.168.1.1"
}

func (gcs *GDPRComplianceService) extractUserAgent(ctx context.Context) string {
	// In a real implementation, this would extract user agent from context
	return "Mozilla/5.0 (compatible; KYB-Platform/1.0)"
}
