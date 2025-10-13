package soc2

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ConfidentialityControls implements SOC 2 confidentiality control requirements
type ConfidentialityControls struct {
	logger *zap.Logger
	config *ConfidentialityConfig
}

// ConfidentialityConfig represents confidentiality control configuration
type ConfidentialityConfig struct {
	EnableDataEncryption     bool          `json:"enable_data_encryption"`
	EnableAccessLogging      bool          `json:"enable_access_logging"`
	EnableDataClassification bool          `json:"enable_data_classification"`
	EnableDataRetention      bool          `json:"enable_data_retention"`
	EnableDataMasking        bool          `json:"enable_data_masking"`
	EncryptionKey            string        `json:"encryption_key"`
	RetentionPeriod          time.Duration `json:"retention_period"`
	MaskingRules             []MaskingRule `json:"masking_rules"`
	AccessLogRetention       time.Duration `json:"access_log_retention"`
}

// MaskingRule represents a data masking rule
type MaskingRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Pattern     string `json:"pattern"`
	Replacement string `json:"replacement"`
	FieldType   string `json:"field_type"`
	IsActive    bool   `json:"is_active"`
}

// DataAccessLog represents a data access log entry
type DataAccessLog struct {
	ID             string                 `json:"id"`
	UserID         string                 `json:"user_id"`
	TenantID       string                 `json:"tenant_id"`
	Resource       string                 `json:"resource"`
	Action         string                 `json:"action"`
	DataType       string                 `json:"data_type"`
	Classification DataClassification     `json:"classification"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	Timestamp      time.Time              `json:"timestamp"`
	Success        bool                   `json:"success"`
	Reason         string                 `json:"reason,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// DataRetentionPolicy represents a data retention policy
type DataRetentionPolicy struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	DataType        string                 `json:"data_type"`
	RetentionPeriod time.Duration          `json:"retention_period"`
	Action          RetentionAction        `json:"action"`
	IsActive        bool                   `json:"is_active"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// RetentionAction represents the action to take when data expires
type RetentionAction string

const (
	RetentionActionDelete    RetentionAction = "delete"
	RetentionActionArchive   RetentionAction = "archive"
	RetentionActionAnonymize RetentionAction = "anonymize"
	RetentionActionMask      RetentionAction = "mask"
)

// EncryptionResult represents the result of an encryption operation
type EncryptionResult struct {
	EncryptedData string    `json:"encrypted_data"`
	KeyID         string    `json:"key_id"`
	Algorithm     string    `json:"algorithm"`
	Timestamp     time.Time `json:"timestamp"`
}

// NewConfidentialityControls creates a new confidentiality controls instance
func NewConfidentialityControls(config *ConfidentialityConfig, logger *zap.Logger) *ConfidentialityControls {
	return &ConfidentialityControls{
		logger: logger,
		config: config,
	}
}

// EncryptData encrypts sensitive data using AES-256-GCM
func (cc *ConfidentialityControls) EncryptData(ctx context.Context, data string, keyID string) (*EncryptionResult, error) {
	if !cc.config.EnableDataEncryption {
		return &EncryptionResult{
			EncryptedData: data,
			KeyID:         keyID,
			Algorithm:     "none",
			Timestamp:     time.Now(),
		}, nil
	}

	// Generate a random key for this encryption
	key := make([]byte, 32) // AES-256 key
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate encryption key: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt data
	encryptedData := gcm.Seal(nonce, nonce, []byte(data), nil)

	// Encode to base64
	encodedData := base64.StdEncoding.EncodeToString(encryptedData)

	cc.logger.Info("Data encrypted successfully",
		zap.String("key_id", keyID),
		zap.Int("data_length", len(data)),
		zap.Int("encrypted_length", len(encodedData)))

	return &EncryptionResult{
		EncryptedData: encodedData,
		KeyID:         keyID,
		Algorithm:     "AES-256-GCM",
		Timestamp:     time.Now(),
	}, nil
}

// DecryptData decrypts sensitive data using AES-256-GCM
func (cc *ConfidentialityControls) DecryptData(ctx context.Context, encryptedData string, keyID string) (string, error) {
	if !cc.config.EnableDataEncryption {
		return encryptedData, nil
	}

	// Decode from base64
	data, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64 data: %w", err)
	}

	// In a real implementation, we would retrieve the key using keyID
	// For now, we'll use a mock key
	key := make([]byte, 32)
	copy(key, []byte(cc.config.EncryptionKey))
	if len(key) < 32 {
		// Pad key if necessary
		for len(key) < 32 {
			key = append(key, 0)
		}
	}

	// Create AES cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM mode: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("encrypted data too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt data
	decryptedData, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %w", err)
	}

	cc.logger.Info("Data decrypted successfully",
		zap.String("key_id", keyID),
		zap.Int("encrypted_length", len(encryptedData)),
		zap.Int("decrypted_length", len(decryptedData)))

	return string(decryptedData), nil
}

// LogDataAccess logs data access for audit purposes
func (cc *ConfidentialityControls) LogDataAccess(ctx context.Context, log *DataAccessLog) error {
	if !cc.config.EnableAccessLogging {
		return nil
	}

	// Generate log ID if not provided
	if log.ID == "" {
		log.ID = generateAccessLogID()
	}

	// Set timestamp if not provided
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	cc.logger.Info("Data access logged",
		zap.String("log_id", log.ID),
		zap.String("user_id", log.UserID),
		zap.String("tenant_id", log.TenantID),
		zap.String("resource", log.Resource),
		zap.String("action", log.Action),
		zap.String("data_type", log.DataType),
		zap.String("classification", string(log.Classification)),
		zap.Bool("success", log.Success))

	// In a real implementation, this would write to an audit log database
	return nil
}

// GetDataAccessLogs retrieves data access logs
func (cc *ConfidentialityControls) GetDataAccessLogs(ctx context.Context, filters map[string]interface{}) ([]*DataAccessLog, error) {
	if !cc.config.EnableAccessLogging {
		return nil, fmt.Errorf("access logging is disabled")
	}

	// In a real implementation, this would query the audit log database
	// For now, return empty list
	return []*DataAccessLog{}, nil
}

// ClassifyData classifies data based on sensitivity
func (cc *ConfidentialityControls) ClassifyData(ctx context.Context, data string, dataType string) (DataClassification, error) {
	if !cc.config.EnableDataClassification {
		return DataClassificationPublic, nil
	}

	// Simple classification logic based on data type and content
	classification := DataClassificationPublic

	// Classify based on data type
	switch dataType {
	case "password", "secret", "api_key", "token":
		classification = DataClassificationConfidential
	case "email", "phone", "ssn", "credit_card":
		classification = DataClassificationConfidential
	case "business_name", "address":
		classification = DataClassificationInternal
	case "public_info", "marketing":
		classification = DataClassificationPublic
	}

	// Classify based on content keywords
	keywords := map[DataClassification][]string{
		DataClassificationConfidential: {"password", "secret", "private", "confidential", "restricted"},
		DataClassificationInternal:     {"internal", "employee", "staff", "business"},
		DataClassificationPublic:       {"public", "general", "marketing"},
	}

	dataLower := strings.ToLower(data)
	for class, words := range keywords {
		for _, word := range words {
			if strings.Contains(dataLower, word) {
				if class < classification { // More restrictive classification
					classification = class
				}
			}
		}
	}

	cc.logger.Info("Data classified",
		zap.String("data_type", dataType),
		zap.String("classification", string(classification)),
		zap.Int("data_length", len(data)))

	return classification, nil
}

// MaskData masks sensitive data according to masking rules
func (cc *ConfidentialityControls) MaskData(ctx context.Context, data string, dataType string) (string, error) {
	if !cc.config.EnableDataMasking {
		return data, nil
	}

	maskedData := data

	// Apply masking rules
	for _, rule := range cc.config.MaskingRules {
		if !rule.IsActive {
			continue
		}

		// Check if rule applies to this data type
		if rule.FieldType != "" && rule.FieldType != dataType {
			continue
		}

		// Apply masking pattern
		if rule.Pattern != "" {
			// In a real implementation, this would use regex
			// For now, we'll do simple string replacement
			if strings.Contains(maskedData, rule.Pattern) {
				maskedData = strings.ReplaceAll(maskedData, rule.Pattern, rule.Replacement)
			}
		}
	}

	// Apply default masking based on data type
	switch dataType {
	case "email":
		maskedData = maskEmail(data)
	case "phone":
		maskedData = maskPhone(data)
	case "ssn":
		maskedData = maskSSN(data)
	case "credit_card":
		maskedData = maskCreditCard(data)
	case "name":
		maskedData = maskName(data)
	}

	cc.logger.Info("Data masked",
		zap.String("data_type", dataType),
		zap.Int("original_length", len(data)),
		zap.Int("masked_length", len(maskedData)))

	return maskedData, nil
}

// CreateRetentionPolicy creates a new data retention policy
func (cc *ConfidentialityControls) CreateRetentionPolicy(ctx context.Context, policy *DataRetentionPolicy) error {
	if !cc.config.EnableDataRetention {
		return fmt.Errorf("data retention is disabled")
	}

	// Generate policy ID if not provided
	if policy.ID == "" {
		policy.ID = generateRetentionPolicyID()
	}

	// Set timestamps
	now := time.Now()
	if policy.CreatedAt.IsZero() {
		policy.CreatedAt = now
	}
	policy.UpdatedAt = now

	cc.logger.Info("Data retention policy created",
		zap.String("policy_id", policy.ID),
		zap.String("name", policy.Name),
		zap.String("data_type", policy.DataType),
		zap.Duration("retention_period", policy.RetentionPeriod),
		zap.String("action", string(policy.Action)))

	// In a real implementation, this would save to a database
	return nil
}

// GetRetentionPolicies retrieves data retention policies
func (cc *ConfidentialityControls) GetRetentionPolicies(ctx context.Context, filters map[string]interface{}) ([]*DataRetentionPolicy, error) {
	if !cc.config.EnableDataRetention {
		return nil, fmt.Errorf("data retention is disabled")
	}

	// In a real implementation, this would query the database
	// For now, return empty list
	return []*DataRetentionPolicy{}, nil
}

// ProcessDataRetention processes data retention for expired data
func (cc *ConfidentialityControls) ProcessDataRetention(ctx context.Context) error {
	if !cc.config.EnableDataRetention {
		return fmt.Errorf("data retention is disabled")
	}

	cc.logger.Info("Processing data retention")

	// In a real implementation, this would:
	// 1. Query for expired data based on retention policies
	// 2. Apply the appropriate action (delete, archive, anonymize, mask)
	// 3. Log the retention actions
	// 4. Update data status

	return nil
}

// GenerateDataHash generates a hash for data integrity verification
func (cc *ConfidentialityControls) GenerateDataHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// VerifyDataIntegrity verifies data integrity using hash
func (cc *ConfidentialityControls) VerifyDataIntegrity(data string, expectedHash string) bool {
	actualHash := cc.GenerateDataHash(data)
	return actualHash == expectedHash
}

// Helper functions for data masking
func maskEmail(email string) string {
	if len(email) < 5 {
		return "***"
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}

	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return "***@" + domain
	}

	return username[:2] + "***@" + domain
}

func maskPhone(phone string) string {
	if len(phone) < 4 {
		return "***"
	}

	return "***-***-" + phone[len(phone)-4:]
}

func maskSSN(ssn string) string {
	if len(ssn) < 4 {
		return "***"
	}

	return "***-**-" + ssn[len(ssn)-4:]
}

func maskCreditCard(card string) string {
	if len(card) < 4 {
		return "***"
	}

	return "****-****-****-" + card[len(card)-4:]
}

func maskName(name string) string {
	if len(name) < 2 {
		return "***"
	}

	return name[:1] + "***"
}

// Helper functions for ID generation
func generateAccessLogID() string {
	return fmt.Sprintf("access_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func generateRetentionPolicyID() string {
	return fmt.Sprintf("retention_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
