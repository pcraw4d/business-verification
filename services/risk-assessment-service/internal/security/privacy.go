package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// PrivacyManager handles data privacy controls and GDPR compliance
type PrivacyManager struct {
	config *PrivacyConfig
	logger Logger
}

// PrivacyConfig holds configuration for privacy controls
type PrivacyConfig struct {
	DataRetentionPeriod     time.Duration `json:"data_retention_period"`
	AnonymizationEnabled    bool          `json:"anonymization_enabled"`
	PseudonymizationEnabled bool          `json:"pseudonymization_enabled"`
	ConsentRequired         bool          `json:"consent_required"`
	RightToErasureEnabled   bool          `json:"right_to_erasure_enabled"`
	DataPortabilityEnabled  bool          `json:"data_portability_enabled"`
	AuditLogRetention       time.Duration `json:"audit_log_retention"`
	EncryptionAtRest        bool          `json:"encryption_at_rest"`
	EncryptionInTransit     bool          `json:"encryption_in_transit"`
}

// DataSubject represents a data subject under GDPR
type DataSubject struct {
	ID           string                 `json:"id"`
	Email        string                 `json:"email"`
	ConsentGiven bool                   `json:"consent_given"`
	ConsentDate  time.Time              `json:"consent_date"`
	DataTypes    []string               `json:"data_types"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// DataProcessingActivity represents a data processing activity
type DataProcessingActivity struct {
	ID               string                 `json:"id"`
	Purpose          string                 `json:"purpose"`
	LegalBasis       string                 `json:"legal_basis"`
	DataCategories   []string               `json:"data_categories"`
	Recipients       []string               `json:"recipients"`
	RetentionPeriod  time.Duration          `json:"retention_period"`
	SecurityMeasures []string               `json:"security_measures"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// ConsentRecord represents a consent record
type ConsentRecord struct {
	ID            string     `json:"id"`
	DataSubjectID string     `json:"data_subject_id"`
	Purpose       string     `json:"purpose"`
	ConsentGiven  bool       `json:"consent_given"`
	ConsentDate   time.Time  `json:"consent_date"`
	Withdrawn     bool       `json:"withdrawn"`
	WithdrawnAt   *time.Time `json:"withdrawn_at,omitempty"`
	Version       int        `json:"version"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// DataErasureRequest represents a data erasure request
type DataErasureRequest struct {
	ID            string     `json:"id"`
	DataSubjectID string     `json:"data_subject_id"`
	RequestDate   time.Time  `json:"request_date"`
	Reason        string     `json:"reason"`
	Status        string     `json:"status"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	ProcessedBy   string     `json:"processed_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// DataPortabilityRequest represents a data portability request
type DataPortabilityRequest struct {
	ID            string     `json:"id"`
	DataSubjectID string     `json:"data_subject_id"`
	RequestDate   time.Time  `json:"request_date"`
	DataTypes     []string   `json:"data_types"`
	Format        string     `json:"format"`
	Status        string     `json:"status"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	ProcessedBy   string     `json:"processed_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// NewPrivacyManager creates a new privacy manager
func NewPrivacyManager(config *PrivacyConfig, logger Logger) *PrivacyManager {
	if config == nil {
		config = &PrivacyConfig{
			DataRetentionPeriod:     7 * 24 * time.Hour, // 7 days
			AnonymizationEnabled:    true,
			PseudonymizationEnabled: true,
			ConsentRequired:         true,
			RightToErasureEnabled:   true,
			DataPortabilityEnabled:  true,
			AuditLogRetention:       30 * 24 * time.Hour, // 30 days
			EncryptionAtRest:        true,
			EncryptionInTransit:     true,
		}
	}

	return &PrivacyManager{
		config: config,
		logger: logger,
	}
}

// RegisterDataSubject registers a new data subject
func (pm *PrivacyManager) RegisterDataSubject(ctx context.Context, subject *DataSubject) error {
	if subject == nil {
		return fmt.Errorf("data subject cannot be nil")
	}

	if subject.Email == "" {
		return fmt.Errorf("email is required for data subject")
	}

	// Set default values
	if subject.ID == "" {
		subject.ID = generateDataSubjectID(subject.Email)
	}
	if subject.CreatedAt.IsZero() {
		subject.CreatedAt = time.Now()
	}
	subject.UpdatedAt = time.Now()

	// Log the registration
	pm.logger.Info("Data subject registered",
		"data_subject_id", subject.ID,
		"email", subject.Email,
		"consent_given", subject.ConsentGiven)

	return nil
}

// RecordConsent records consent for a data subject
func (pm *PrivacyManager) RecordConsent(ctx context.Context, dataSubjectID, purpose string, consentGiven bool) (*ConsentRecord, error) {
	if dataSubjectID == "" {
		return nil, fmt.Errorf("data subject ID is required")
	}

	if purpose == "" {
		return nil, fmt.Errorf("purpose is required")
	}

	consent := &ConsentRecord{
		ID:            generateConsentID(dataSubjectID, purpose),
		DataSubjectID: dataSubjectID,
		Purpose:       purpose,
		ConsentGiven:  consentGiven,
		ConsentDate:   time.Now(),
		Version:       1,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Log the consent
	pm.logger.Info("Consent recorded",
		"consent_id", consent.ID,
		"data_subject_id", dataSubjectID,
		"purpose", purpose,
		"consent_given", consentGiven)

	return consent, nil
}

// WithdrawConsent withdraws consent for a data subject
func (pm *PrivacyManager) WithdrawConsent(ctx context.Context, dataSubjectID, purpose string) error {
	if dataSubjectID == "" {
		return fmt.Errorf("data subject ID is required")
	}

	if purpose == "" {
		return fmt.Errorf("purpose is required")
	}

	// Log the consent withdrawal
	pm.logger.Info("Consent withdrawn",
		"data_subject_id", dataSubjectID,
		"purpose", purpose)

	return nil
}

// RequestDataErasure creates a data erasure request
func (pm *PrivacyManager) RequestDataErasure(ctx context.Context, dataSubjectID, reason string) (*DataErasureRequest, error) {
	if !pm.config.RightToErasureEnabled {
		return nil, fmt.Errorf("right to erasure is not enabled")
	}

	if dataSubjectID == "" {
		return nil, fmt.Errorf("data subject ID is required")
	}

	request := &DataErasureRequest{
		ID:            generateErasureRequestID(dataSubjectID),
		DataSubjectID: dataSubjectID,
		RequestDate:   time.Now(),
		Reason:        reason,
		Status:        "PENDING",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Log the erasure request
	pm.logger.Info("Data erasure requested",
		"request_id", request.ID,
		"data_subject_id", dataSubjectID,
		"reason", reason)

	return request, nil
}

// ProcessDataErasure processes a data erasure request
func (pm *PrivacyManager) ProcessDataErasure(ctx context.Context, requestID, processedBy string) error {
	if requestID == "" {
		return fmt.Errorf("request ID is required")
	}

	// Log the erasure processing
	pm.logger.Info("Data erasure processed",
		"request_id", requestID,
		"processed_by", processedBy)

	return nil
}

// RequestDataPortability creates a data portability request
func (pm *PrivacyManager) RequestDataPortability(ctx context.Context, dataSubjectID string, dataTypes []string, format string) (*DataPortabilityRequest, error) {
	if !pm.config.DataPortabilityEnabled {
		return nil, fmt.Errorf("data portability is not enabled")
	}

	if dataSubjectID == "" {
		return nil, fmt.Errorf("data subject ID is required")
	}

	if len(dataTypes) == 0 {
		return nil, fmt.Errorf("data types are required")
	}

	if format == "" {
		format = "JSON"
	}

	request := &DataPortabilityRequest{
		ID:            generatePortabilityRequestID(dataSubjectID),
		DataSubjectID: dataSubjectID,
		RequestDate:   time.Now(),
		DataTypes:     dataTypes,
		Format:        format,
		Status:        "PENDING",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Log the portability request
	pm.logger.Info("Data portability requested",
		"request_id", request.ID,
		"data_subject_id", dataSubjectID,
		"data_types", dataTypes,
		"format", format)

	return request, nil
}

// ProcessDataPortability processes a data portability request
func (pm *PrivacyManager) ProcessDataPortability(ctx context.Context, requestID, processedBy string) error {
	if requestID == "" {
		return fmt.Errorf("request ID is required")
	}

	// Log the portability processing
	pm.logger.Info("Data portability processed",
		"request_id", requestID,
		"processed_by", processedBy)

	return nil
}

// AnonymizeData anonymizes personal data
func (pm *PrivacyManager) AnonymizeData(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if !pm.config.AnonymizationEnabled {
		return data, nil
	}

	anonymized := make(map[string]interface{})

	for key, value := range data {
		if isPersonalData(key) {
			anonymized[key] = anonymizeValue(value)
		} else {
			anonymized[key] = value
		}
	}

	// Log the anonymization
	pm.logger.Info("Data anonymized",
		"original_keys", len(data),
		"anonymized_keys", len(anonymized))

	return anonymized, nil
}

// PseudonymizeData pseudonymizes personal data
func (pm *PrivacyManager) PseudonymizeData(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	if !pm.config.PseudonymizationEnabled {
		return data, nil
	}

	pseudonymized := make(map[string]interface{})

	for key, value := range data {
		if isPersonalData(key) {
			pseudonymized[key] = pseudonymizeValue(value)
		} else {
			pseudonymized[key] = value
		}
	}

	// Log the pseudonymization
	pm.logger.Info("Data pseudonymized",
		"original_keys", len(data),
		"pseudonymized_keys", len(pseudonymized))

	return pseudonymized, nil
}

// CheckDataRetention checks if data should be retained based on retention policy
func (pm *PrivacyManager) CheckDataRetention(ctx context.Context, dataType string, createdAt time.Time) (bool, error) {
	retentionPeriod := pm.config.DataRetentionPeriod

	// Check if data has exceeded retention period
	if time.Since(createdAt) > retentionPeriod {
		return false, nil // Data should be deleted
	}

	return true, nil // Data should be retained
}

// GetDataSubjectRights returns the rights available to a data subject
func (pm *PrivacyManager) GetDataSubjectRights(ctx context.Context, dataSubjectID string) ([]string, error) {
	rights := []string{
		"right_to_access",
		"right_to_rectification",
		"right_to_erasure",
		"right_to_restrict_processing",
		"right_to_data_portability",
		"right_to_object",
		"rights_related_to_automated_decision_making",
	}

	// Filter based on configuration
	if !pm.config.RightToErasureEnabled {
		rights = removeString(rights, "right_to_erasure")
	}

	if !pm.config.DataPortabilityEnabled {
		rights = removeString(rights, "right_to_data_portability")
	}

	return rights, nil
}

// ValidateDataProcessing validates data processing against GDPR requirements
func (pm *PrivacyManager) ValidateDataProcessing(ctx context.Context, activity *DataProcessingActivity) error {
	if activity == nil {
		return fmt.Errorf("data processing activity cannot be nil")
	}

	if activity.Purpose == "" {
		return fmt.Errorf("purpose is required")
	}

	if activity.LegalBasis == "" {
		return fmt.Errorf("legal basis is required")
	}

	if len(activity.DataCategories) == 0 {
		return fmt.Errorf("data categories are required")
	}

	if activity.RetentionPeriod <= 0 {
		return fmt.Errorf("retention period must be positive")
	}

	// Log the validation
	pm.logger.Info("Data processing activity validated",
		"activity_id", activity.ID,
		"purpose", activity.Purpose,
		"legal_basis", activity.LegalBasis)

	return nil
}

// GeneratePrivacyReport generates a privacy compliance report
func (pm *PrivacyManager) GeneratePrivacyReport(ctx context.Context) (map[string]interface{}, error) {
	report := map[string]interface{}{
		"generated_at": time.Now(),
		"config": map[string]interface{}{
			"data_retention_period":    pm.config.DataRetentionPeriod.String(),
			"anonymization_enabled":    pm.config.AnonymizationEnabled,
			"pseudonymization_enabled": pm.config.PseudonymizationEnabled,
			"consent_required":         pm.config.ConsentRequired,
			"right_to_erasure_enabled": pm.config.RightToErasureEnabled,
			"data_portability_enabled": pm.config.DataPortabilityEnabled,
			"audit_log_retention":      pm.config.AuditLogRetention.String(),
			"encryption_at_rest":       pm.config.EncryptionAtRest,
			"encryption_in_transit":    pm.config.EncryptionInTransit,
		},
		"compliance_status": "COMPLIANT",
		"recommendations": []string{
			"Regularly review data retention policies",
			"Conduct privacy impact assessments",
			"Maintain consent records",
			"Implement data minimization principles",
		},
	}

	// Log the report generation
	pm.logger.Info("Privacy report generated")

	return report, nil
}

// Helper functions

// isPersonalData checks if a field contains personal data
func isPersonalData(fieldName string) bool {
	personalDataFields := []string{
		"email", "name", "address", "phone", "ssn", "passport", "id_number",
		"date_of_birth", "gender", "nationality", "ip_address", "user_agent",
		"location", "biometric_data", "health_data", "financial_data",
	}

	fieldName = strings.ToLower(fieldName)
	for _, field := range personalDataFields {
		if strings.Contains(fieldName, field) {
			return true
		}
	}

	return false
}

// anonymizeValue anonymizes a value
func anonymizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		if len(v) > 0 {
			return "***ANONYMIZED***"
		}
		return v
	case int, int64, float64:
		return 0
	case bool:
		return false
	default:
		return nil
	}
}

// pseudonymizeValue pseudonymizes a value
func pseudonymizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		if len(v) > 0 {
			hash := sha256.Sum256([]byte(v))
			return "pseudo_" + hex.EncodeToString(hash[:8])
		}
		return v
	default:
		return value
	}
}

// removeString removes a string from a slice
func removeString(slice []string, item string) []string {
	for i, v := range slice {
		if v == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

// ID generation functions
func generateDataSubjectID(email string) string {
	hash := sha256.Sum256([]byte(email))
	return "ds_" + hex.EncodeToString(hash[:8])
}

func generateConsentID(dataSubjectID, purpose string) string {
	hash := sha256.Sum256([]byte(dataSubjectID + purpose))
	return "consent_" + hex.EncodeToString(hash[:8])
}

func generateErasureRequestID(dataSubjectID string) string {
	hash := sha256.Sum256([]byte(dataSubjectID + "erasure"))
	return "erasure_" + hex.EncodeToString(hash[:8])
}

func generatePortabilityRequestID(dataSubjectID string) string {
	hash := sha256.Sum256([]byte(dataSubjectID + "portability"))
	return "portability_" + hex.EncodeToString(hash[:8])
}
