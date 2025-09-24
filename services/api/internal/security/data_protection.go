package security

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// DataProtectionService provides comprehensive data protection and anonymization
type DataProtectionService struct {
	logger     *zap.Logger
	config     *DataProtectionConfig
	anonymizer *DataAnonymizer
	encryptor  *DataEncryptor
	validator  *PrivacyValidator
}

// DataProtectionConfig holds configuration for data protection features
type DataProtectionConfig struct {
	// Anonymization settings
	EnableAnonymization bool   `json:"enable_anonymization" yaml:"enable_anonymization"`
	AnonymizationMethod string `json:"anonymization_method" yaml:"anonymization_method"` // "hash", "mask", "pseudonymize"
	SaltLength          int    `json:"salt_length" yaml:"salt_length"`
	HashAlgorithm       string `json:"hash_algorithm" yaml:"hash_algorithm"` // "sha256", "fnv"

	// Encryption settings
	EnableEncryption    bool          `json:"enable_encryption" yaml:"enable_encryption"`
	EncryptionAlgorithm string        `json:"encryption_algorithm" yaml:"encryption_algorithm"` // "aes-256-gcm"
	KeyRotationInterval time.Duration `json:"key_rotation_interval" yaml:"key_rotation_interval"`

	// Privacy validation
	EnablePrivacyValidation bool `json:"enable_privacy_validation" yaml:"enable_privacy_validation"`
	StrictMode              bool `json:"strict_mode" yaml:"strict_mode"`

	// Data retention
	DefaultRetentionPeriod time.Duration `json:"default_retention_period" yaml:"default_retention_period"`
	MaxRetentionPeriod     time.Duration `json:"max_retention_period" yaml:"max_retention_period"`

	// PII detection patterns
	PIIPatterns map[string][]string `json:"pii_patterns" yaml:"pii_patterns"`
}

// SensitiveData represents data that needs protection
type SensitiveData struct {
	ID               string                 `json:"id"`
	DataType         string                 `json:"data_type"` // "email", "phone", "address", "name", "financial"
	SensitivityLevel SensitivityLevel       `json:"sensitivity_level"`
	OriginalValue    string                 `json:"original_value"`
	AnonymizedValue  string                 `json:"anonymized_value,omitempty"`
	EncryptedValue   string                 `json:"encrypted_value,omitempty"`
	HashValue        string                 `json:"hash_value,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	ExpiresAt        *time.Time             `json:"expires_at,omitempty"`
}

// SensitivityLevel represents the sensitivity level of data
type SensitivityLevel string

const (
	SensitivityLevelPublic       SensitivityLevel = "public"
	SensitivityLevelInternal     SensitivityLevel = "internal"
	SensitivityLevelConfidential SensitivityLevel = "confidential"
	SensitivityLevelRestricted   SensitivityLevel = "restricted"
)

// AnonymizationResult represents the result of data anonymization
type AnonymizationResult struct {
	OriginalData        map[string]interface{} `json:"original_data"`
	AnonymizedData      map[string]interface{} `json:"anonymized_data"`
	ProtectedFields     []string               `json:"protected_fields"`
	AnonymizationMethod string                 `json:"anonymization_method"`
	ConfidenceScore     float64                `json:"confidence_score"`
	ProcessingTime      time.Duration          `json:"processing_time"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}

// PrivacyValidationResult represents privacy validation results
type PrivacyValidationResult struct {
	IsCompliant     bool               `json:"is_compliant"`
	Violations      []PrivacyViolation `json:"violations,omitempty"`
	Warnings        []PrivacyWarning   `json:"warnings,omitempty"`
	Recommendations []string           `json:"recommendations,omitempty"`
	ComplianceScore float64            `json:"compliance_score"`
	ValidationTime  time.Duration      `json:"validation_time"`
}

// PrivacyViolation represents a privacy compliance violation
type PrivacyViolation struct {
	Type           string `json:"type"`     // "pii_exposure", "retention_violation", "consent_missing"
	Severity       string `json:"severity"` // "low", "medium", "high", "critical"
	Field          string `json:"field"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// PrivacyWarning represents a privacy compliance warning
type PrivacyWarning struct {
	Type           string `json:"type"`
	Field          string `json:"field"`
	Description    string `json:"description"`
	Recommendation string `json:"recommendation"`
}

// DataAnonymizer handles data anonymization
type DataAnonymizer struct {
	config *DataProtectionConfig
	logger *zap.Logger
}

// DataEncryptor handles data encryption
type DataEncryptor struct {
	config *DataProtectionConfig
	logger *zap.Logger
	key    []byte
}

// PrivacyValidator validates privacy compliance
type PrivacyValidator struct {
	config *DataProtectionConfig
	logger *zap.Logger
}

// NewDataProtectionService creates a new data protection service
func NewDataProtectionService(config *DataProtectionConfig, logger *zap.Logger) (*DataProtectionService, error) {
	if config == nil {
		config = &DataProtectionConfig{
			EnableAnonymization:     true,
			AnonymizationMethod:     "hash",
			SaltLength:              32,
			HashAlgorithm:           "sha256",
			EnableEncryption:        true,
			EncryptionAlgorithm:     "aes-256-gcm",
			KeyRotationInterval:     24 * time.Hour,
			EnablePrivacyValidation: true,
			StrictMode:              false,
			DefaultRetentionPeriod:  30 * 24 * time.Hour,  // 30 days
			MaxRetentionPeriod:      365 * 24 * time.Hour, // 1 year
			PIIPatterns: map[string][]string{
				"email": {
					`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
				},
				"phone": {
					`(\+?1[-.\s]?)?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})`,
					`\+[1-9]\d{1,14}`,
				},
				"ssn": {
					`\d{3}-\d{2}-\d{4}`,
					`\d{9}`,
				},
				"credit_card": {
					`\d{4}[-.\s]?\d{4}[-.\s]?\d{4}[-.\s]?\d{4}`,
				},
			},
		}
	}

	anonymizer := &DataAnonymizer{
		config: config,
		logger: logger,
	}

	encryptor := &DataEncryptor{
		config: config,
		logger: logger,
		key:    make([]byte, 32), // 256-bit key
	}
	if _, err := rand.Read(encryptor.key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	validator := &PrivacyValidator{
		config: config,
		logger: logger,
	}

	return &DataProtectionService{
		logger:     logger,
		config:     config,
		anonymizer: anonymizer,
		encryptor:  encryptor,
		validator:  validator,
	}, nil
}

// ProtectBusinessData protects and anonymizes business data
func (dps *DataProtectionService) ProtectBusinessData(ctx context.Context, data map[string]interface{}) (*AnonymizationResult, error) {
	startTime := time.Now()
	dps.logger.Info("Starting business data protection", zap.Any("data_keys", getMapKeys(data)))

	// Detect sensitive fields
	sensitiveFields := dps.detectSensitiveFields(data)
	dps.logger.Info("Detected sensitive fields", zap.Strings("fields", sensitiveFields))

	// Create anonymized copy
	anonymizedData := make(map[string]interface{})
	for key, value := range data {
		if contains(sensitiveFields, key) {
			anonymizedValue, err := dps.anonymizer.AnonymizeValue(value, key)
			if err != nil {
				dps.logger.Error("Failed to anonymize value", zap.String("field", key), zap.Error(err))
				continue
			}
			anonymizedData[key] = anonymizedValue
		} else {
			anonymizedData[key] = value
		}
	}

	processingTime := time.Since(startTime)
	result := &AnonymizationResult{
		OriginalData:        data,
		AnonymizedData:      anonymizedData,
		ProtectedFields:     sensitiveFields,
		AnonymizationMethod: dps.config.AnonymizationMethod,
		ConfidenceScore:     dps.calculateConfidenceScore(sensitiveFields, data),
		ProcessingTime:      processingTime,
		Metadata: map[string]interface{}{
			"anonymization_timestamp": startTime,
			"protected_field_count":   len(sensitiveFields),
		},
	}

	dps.logger.Info("Business data protection completed",
		zap.Duration("processing_time", processingTime),
		zap.Float64("confidence_score", result.ConfidenceScore),
		zap.Int("protected_fields", len(sensitiveFields)))

	return result, nil
}

// ValidatePrivacyCompliance validates privacy compliance of data
func (dps *DataProtectionService) ValidatePrivacyCompliance(ctx context.Context, data map[string]interface{}) (*PrivacyValidationResult, error) {
	startTime := time.Now()
	dps.logger.Info("Starting privacy compliance validation")

	result := &PrivacyValidationResult{
		IsCompliant:     true,
		Violations:      []PrivacyViolation{},
		Warnings:        []PrivacyWarning{},
		Recommendations: []string{},
	}

	// Check for PII exposure
	piiViolations := dps.validator.checkPIIExposure(data)
	result.Violations = append(result.Violations, piiViolations...)

	// Check data retention compliance
	retentionViolations := dps.validator.checkRetentionCompliance(data)
	result.Violations = append(result.Violations, retentionViolations...)

	// Check consent requirements
	consentViolations := dps.validator.checkConsentRequirements(data)
	result.Violations = append(result.Violations, consentViolations...)

	// Generate warnings for potential issues
	warnings := dps.validator.generateWarnings(data)
	result.Warnings = append(result.Warnings, warnings...)

	// Calculate compliance score
	result.ComplianceScore = dps.validator.calculateComplianceScore(result.Violations, result.Warnings)
	result.IsCompliant = len(result.Violations) == 0

	// Generate recommendations
	result.Recommendations = dps.validator.generateRecommendations(result.Violations, result.Warnings)

	result.ValidationTime = time.Since(startTime)

	dps.logger.Info("Privacy compliance validation completed",
		zap.Bool("is_compliant", result.IsCompliant),
		zap.Float64("compliance_score", result.ComplianceScore),
		zap.Int("violations", len(result.Violations)),
		zap.Int("warnings", len(result.Warnings)),
		zap.Duration("validation_time", result.ValidationTime))

	return result, nil
}

// EncryptSensitiveData encrypts sensitive data
func (dps *DataProtectionService) EncryptSensitiveData(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error) {
	dps.logger.Info("Starting sensitive data encryption")

	encryptedData := make(map[string]interface{})
	for key, value := range data {
		if dps.isSensitiveField(key) {
			encryptedValue, err := dps.encryptor.EncryptValue(value)
			if err != nil {
				dps.logger.Error("Failed to encrypt value", zap.String("field", key), zap.Error(err))
				continue
			}
			encryptedData[key] = encryptedValue
		} else {
			encryptedData[key] = value
		}
	}

	dps.logger.Info("Sensitive data encryption completed", zap.Int("encrypted_fields", len(encryptedData)))
	return encryptedData, nil
}

// DecryptSensitiveData decrypts sensitive data
func (dps *DataProtectionService) DecryptSensitiveData(ctx context.Context, encryptedData map[string]interface{}) (map[string]interface{}, error) {
	dps.logger.Info("Starting sensitive data decryption")

	decryptedData := make(map[string]interface{})
	for key, value := range encryptedData {
		if dps.isSensitiveField(key) {
			decryptedValue, err := dps.encryptor.DecryptValue(value)
			if err != nil {
				dps.logger.Error("Failed to decrypt value", zap.String("field", key), zap.Error(err))
				continue
			}
			decryptedData[key] = decryptedValue
		} else {
			decryptedData[key] = value
		}
	}

	dps.logger.Info("Sensitive data decryption completed", zap.Int("decrypted_fields", len(decryptedData)))
	return decryptedData, nil
}

// AnonymizeValue anonymizes a single value
func (da *DataAnonymizer) AnonymizeValue(value interface{}, fieldName string) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	valueStr := fmt.Sprintf("%v", value)
	if valueStr == "" {
		return "", nil
	}

	switch da.config.AnonymizationMethod {
	case "hash":
		return da.hashValue(valueStr, fieldName)
	case "mask":
		return da.maskValue(valueStr, fieldName)
	case "pseudonymize":
		return da.pseudonymizeValue(valueStr, fieldName)
	default:
		return da.hashValue(valueStr, fieldName)
	}
}

// hashValue creates a hash of the value
func (da *DataAnonymizer) hashValue(value, fieldName string) (string, error) {
	salt := make([]byte, da.config.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	var hash []byte
	switch da.config.HashAlgorithm {
	case "sha256":
		h := sha256.New()
		h.Write(salt)
		h.Write([]byte(value))
		hash = h.Sum(nil)
	case "fnv":
		h := fnv.New64a()
		h.Write(salt)
		h.Write([]byte(value))
		hash = h.Sum(nil)
	default:
		h := sha256.New()
		h.Write(salt)
		h.Write([]byte(value))
		hash = h.Sum(nil)
	}

	return base64.StdEncoding.EncodeToString(hash), nil
}

// maskValue masks sensitive parts of the value
func (da *DataAnonymizer) maskValue(value, fieldName string) (string, error) {
	if len(value) <= 4 {
		return strings.Repeat("*", len(value)), nil
	}

	// Keep first and last character, mask the rest
	return string(value[0]) + strings.Repeat("*", len(value)-2) + string(value[len(value)-1]), nil
}

// pseudonymizeValue creates a pseudonym for the value
func (da *DataAnonymizer) pseudonymizeValue(value, fieldName string) (string, error) {
	// Create a deterministic pseudonym based on the value and field name
	h := fnv.New64a()
	h.Write([]byte(value))
	h.Write([]byte(fieldName))
	hash := h.Sum64()

	// Convert to a readable pseudonym
	pseudonym := fmt.Sprintf("pseudo_%s_%x", fieldName, hash)
	return pseudonym, nil
}

// EncryptValue encrypts a single value
func (de *DataEncryptor) EncryptValue(value interface{}) (string, error) {
	if value == nil {
		return "", nil
	}

	valueBytes, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal value: %w", err)
	}

	block, err := aes.NewCipher(de.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := gcm.Seal(nonce, nonce, valueBytes, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptValue decrypts a single value
func (de *DataEncryptor) DecryptValue(encryptedValue interface{}) (interface{}, error) {
	if encryptedValue == nil {
		return nil, nil
	}

	encryptedStr, ok := encryptedValue.(string)
	if !ok {
		return nil, fmt.Errorf("encrypted value is not a string")
	}

	// Handle empty string (represents nil value)
	if encryptedStr == "" {
		return nil, nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(encryptedStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	block, err := aes.NewCipher(de.key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	var result interface{}
	if err := json.Unmarshal(plaintext, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal decrypted value: %w", err)
	}

	return result, nil
}

// detectSensitiveFields detects sensitive fields in the data
func (dps *DataProtectionService) detectSensitiveFields(data map[string]interface{}) []string {
	var sensitiveFields []string

	for key, value := range data {
		if dps.isSensitiveField(key) || dps.containsSensitiveData(value) {
			sensitiveFields = append(sensitiveFields, key)
		}
	}

	return sensitiveFields
}

// isSensitiveField checks if a field name indicates sensitive data
func (dps *DataProtectionService) isSensitiveField(fieldName string) bool {
	sensitivePatterns := []string{
		"email", "phone", "address", "ssn", "social", "credit", "card",
		"password", "secret", "key", "token", "auth", "login", "user",
		"personal", "private", "confidential", "sensitive",
	}

	fieldLower := strings.ToLower(fieldName)
	for _, pattern := range sensitivePatterns {
		if strings.Contains(fieldLower, pattern) {
			return true
		}
	}

	return false
}

// containsSensitiveData checks if a value contains sensitive data
func (dps *DataProtectionService) containsSensitiveData(value interface{}) bool {
	if value == nil {
		return false
	}

	valueStr := fmt.Sprintf("%v", value)
	for dataType, patterns := range dps.config.PIIPatterns {
		for _, pattern := range patterns {
			matched, _ := regexp.MatchString(pattern, valueStr)
			if matched {
				dps.logger.Debug("Detected sensitive data", zap.String("type", dataType), zap.String("pattern", pattern))
				return true
			}
		}
	}

	return false
}

// calculateConfidenceScore calculates confidence score for anonymization
func (dps *DataProtectionService) calculateConfidenceScore(sensitiveFields []string, data map[string]interface{}) float64 {
	if len(data) == 0 {
		return 0.0
	}

	detectedRatio := float64(len(sensitiveFields)) / float64(len(data))
	confidence := 1.0 - detectedRatio // Higher confidence when fewer sensitive fields detected

	// Adjust based on anonymization method
	switch dps.config.AnonymizationMethod {
	case "hash":
		confidence *= 0.95
	case "mask":
		confidence *= 0.85
	case "pseudonymize":
		confidence *= 0.90
	}

	return confidence
}

// checkPIIExposure checks for PII exposure violations
func (pv *PrivacyValidator) checkPIIExposure(data map[string]interface{}) []PrivacyViolation {
	var violations []PrivacyViolation

	for key, value := range data {
		if pv.containsPII(value) {
			violations = append(violations, PrivacyViolation{
				Type:           "pii_exposure",
				Severity:       "high",
				Field:          key,
				Description:    fmt.Sprintf("Field '%s' contains PII data", key),
				Recommendation: "Anonymize or encrypt this field before processing",
			})
		}
	}

	return violations
}

// checkRetentionCompliance checks data retention compliance
func (pv *PrivacyValidator) checkRetentionCompliance(data map[string]interface{}) []PrivacyViolation {
	var violations []PrivacyViolation

	// Check if data has retention metadata
	if retentionData, ok := data["retention_period"]; ok {
		if retentionPeriod, ok := retentionData.(time.Duration); ok {
			if retentionPeriod > pv.config.MaxRetentionPeriod {
				violations = append(violations, PrivacyViolation{
					Type:           "retention_violation",
					Severity:       "medium",
					Field:          "retention_period",
					Description:    fmt.Sprintf("Retention period %v exceeds maximum allowed %v", retentionPeriod, pv.config.MaxRetentionPeriod),
					Recommendation: "Reduce retention period to comply with privacy regulations",
				})
			}
		}
	}

	return violations
}

// checkConsentRequirements checks consent requirements
func (pv *PrivacyValidator) checkConsentRequirements(data map[string]interface{}) []PrivacyViolation {
	var violations []PrivacyViolation

	// Check if consent is documented for sensitive data
	if pv.containsSensitiveData(data) {
		if _, hasConsent := data["consent_given"]; !hasConsent {
			violations = append(violations, PrivacyViolation{
				Type:           "consent_missing",
				Severity:       "high",
				Field:          "consent_given",
				Description:    "Sensitive data detected without consent documentation",
				Recommendation: "Document explicit consent for processing sensitive data",
			})
		}
	}

	return violations
}

// generateWarnings generates privacy warnings
func (pv *PrivacyValidator) generateWarnings(data map[string]interface{}) []PrivacyWarning {
	var warnings []PrivacyWarning

	// Check for potential PII patterns
	for key, value := range data {
		if pv.matchesPIIPattern(value) {
			warnings = append(warnings, PrivacyWarning{
				Type:           "potential_pii",
				Field:          key,
				Description:    fmt.Sprintf("Field '%s' may contain PII data", key),
				Recommendation: "Review and validate if this field contains sensitive information",
			})
		}
	}

	return warnings
}

// calculateComplianceScore calculates privacy compliance score
func (pv *PrivacyValidator) calculateComplianceScore(violations []PrivacyViolation, warnings []PrivacyWarning) float64 {
	baseScore := 100.0

	// Deduct points for violations
	for _, violation := range violations {
		switch violation.Severity {
		case "critical":
			baseScore -= 25
		case "high":
			baseScore -= 15
		case "medium":
			baseScore -= 10
		case "low":
			baseScore -= 5
		}
	}

	// Deduct points for warnings
	for range warnings {
		baseScore -= 2
	}

	if baseScore < 0 {
		baseScore = 0
	}

	return baseScore
}

// generateRecommendations generates privacy recommendations
func (pv *PrivacyValidator) generateRecommendations(violations []PrivacyViolation, warnings []PrivacyWarning) []string {
	var recommendations []string

	// Add recommendations based on violations
	for _, violation := range violations {
		recommendations = append(recommendations, violation.Recommendation)
	}

	// Add recommendations based on warnings
	for _, warning := range warnings {
		recommendations = append(recommendations, warning.Recommendation)
	}

	// Add general recommendations
	if len(violations) > 0 {
		recommendations = append(recommendations, "Implement comprehensive data protection measures")
		recommendations = append(recommendations, "Conduct regular privacy impact assessments")
	}

	return recommendations
}

// containsPII checks if data contains PII
func (pv *PrivacyValidator) containsPII(value interface{}) bool {
	if value == nil {
		return false
	}

	valueStr := fmt.Sprintf("%v", value)
	for _, patterns := range pv.config.PIIPatterns {
		for _, pattern := range patterns {
			matched, _ := regexp.MatchString(pattern, valueStr)
			if matched {
				return true
			}
		}
	}

	return false
}

// matchesPIIPattern checks if data matches PII patterns
func (pv *PrivacyValidator) matchesPIIPattern(value interface{}) bool {
	// Similar to containsPII but with less strict matching
	if value == nil {
		return false
	}

	valueStr := fmt.Sprintf("%v", value)

	// Check for common PII indicators
	piiIndicators := []string{
		"@", // Email indicator
		"-", // Phone/SSN indicator
		"(", // Phone indicator
		")", // Phone indicator
		" ", // Name indicator
	}

	for _, indicator := range piiIndicators {
		if strings.Contains(valueStr, indicator) {
			return true
		}
	}

	return false
}

// containsSensitiveData checks if data contains sensitive information
func (pv *PrivacyValidator) containsSensitiveData(data map[string]interface{}) bool {
	for _, value := range data {
		if pv.containsPII(value) {
			return true
		}
	}
	return false
}

// Helper functions
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
