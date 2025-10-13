package soc2

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	mathrand "math/rand"
	"strings"
	"time"

	"go.uber.org/zap"
)

// SecurityControls implements SOC 2 security control requirements
type SecurityControls struct {
	logger *zap.Logger
	config *SecurityConfig
}

// SecurityConfig represents SOC 2 security configuration
type SecurityConfig struct {
	EnableAccessControl      bool           `json:"enable_access_control"`
	EnableEncryption         bool           `json:"enable_encryption"`
	EnableAuditLogging       bool           `json:"enable_audit_logging"`
	EnableIncidentResponse   bool           `json:"enable_incident_response"`
	EnableVulnerabilityMgmt  bool           `json:"enable_vulnerability_mgmt"`
	PasswordPolicy           PasswordPolicy `json:"password_policy"`
	SessionTimeout           time.Duration  `json:"session_timeout"`
	MaxLoginAttempts         int            `json:"max_login_attempts"`
	LockoutDuration          time.Duration  `json:"lockout_duration"`
	EncryptionKey            string         `json:"encryption_key"`
	RequireMFA               bool           `json:"require_mfa"`
	EnableDataClassification bool           `json:"enable_data_classification"`
}

// PasswordPolicy represents password security requirements
type PasswordPolicy struct {
	MinLength           int  `json:"min_length"`
	RequireUppercase    bool `json:"require_uppercase"`
	RequireLowercase    bool `json:"require_lowercase"`
	RequireNumbers      bool `json:"require_numbers"`
	RequireSpecialChars bool `json:"require_special_chars"`
	MaxAge              int  `json:"max_age_days"`
	HistoryCount        int  `json:"history_count"`
}

// AccessControlEntry represents an access control entry
type AccessControlEntry struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	Resource   string                 `json:"resource"`
	Action     string                 `json:"action"`
	Permission string                 `json:"permission"`
	GrantedBy  string                 `json:"granted_by"`
	GrantedAt  time.Time              `json:"granted_at"`
	ExpiresAt  *time.Time             `json:"expires_at"`
	IsActive   bool                   `json:"is_active"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// SecurityIncident represents a security incident
type SecurityIncident struct {
	ID          string                 `json:"id"`
	Type        IncidentType           `json:"type"`
	Severity    IncidentSeverity       `json:"severity"`
	Status      IncidentStatus         `json:"status"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	DetectedAt  time.Time              `json:"detected_at"`
	ReportedBy  string                 `json:"reported_by"`
	AssignedTo  string                 `json:"assigned_to"`
	Resolution  string                 `json:"resolution"`
	ResolvedAt  *time.Time             `json:"resolved_at"`
	Impact      string                 `json:"impact"`
	RootCause   string                 `json:"root_cause"`
	Prevention  string                 `json:"prevention"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// IncidentType represents the type of security incident
type IncidentType string

const (
	IncidentTypeDataBreach         IncidentType = "data_breach"
	IncidentTypeUnauthorizedAccess IncidentType = "unauthorized_access"
	IncidentTypeMalware            IncidentType = "malware"
	IncidentTypePhishing           IncidentType = "phishing"
	IncidentTypeInsiderThreat      IncidentType = "insider_threat"
	IncidentTypeSystemCompromise   IncidentType = "system_compromise"
	IncidentTypeDDoS               IncidentType = "ddos"
	IncidentTypeVulnerability      IncidentType = "vulnerability"
	IncidentTypeOther              IncidentType = "other"
)

// IncidentSeverity represents the severity of a security incident
type IncidentSeverity string

const (
	IncidentSeverityCritical IncidentSeverity = "critical"
	IncidentSeverityHigh     IncidentSeverity = "high"
	IncidentSeverityMedium   IncidentSeverity = "medium"
	IncidentSeverityLow      IncidentSeverity = "low"
)

// IncidentStatus represents the status of a security incident
type IncidentStatus string

const (
	IncidentStatusOpen       IncidentStatus = "open"
	IncidentStatusInProgress IncidentStatus = "in_progress"
	IncidentStatusResolved   IncidentStatus = "resolved"
	IncidentStatusClosed     IncidentStatus = "closed"
)

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID              string                 `json:"id"`
	CVE             string                 `json:"cve"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Severity        VulnerabilitySeverity  `json:"severity"`
	CVSSScore       float64                `json:"cvss_score"`
	AffectedSystems []string               `json:"affected_systems"`
	Status          VulnerabilityStatus    `json:"status"`
	DiscoveredAt    time.Time              `json:"discovered_at"`
	DiscoveredBy    string                 `json:"discovered_by"`
	Remediation     string                 `json:"remediation"`
	RemediatedAt    *time.Time             `json:"remediated_at"`
	VerifiedAt      *time.Time             `json:"verified_at"`
	Metadata        map[string]interface{} `json:"metadata"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// VulnerabilitySeverity represents the severity of a vulnerability
type VulnerabilitySeverity string

const (
	VulnerabilitySeverityCritical VulnerabilitySeverity = "critical"
	VulnerabilitySeverityHigh     VulnerabilitySeverity = "high"
	VulnerabilitySeverityMedium   VulnerabilitySeverity = "medium"
	VulnerabilitySeverityLow      VulnerabilitySeverity = "low"
)

// VulnerabilityStatus represents the status of a vulnerability
type VulnerabilityStatus string

const (
	VulnerabilityStatusOpen       VulnerabilityStatus = "open"
	VulnerabilityStatusInProgress VulnerabilityStatus = "in_progress"
	VulnerabilityStatusRemediated VulnerabilityStatus = "remediated"
	VulnerabilityStatusVerified   VulnerabilityStatus = "verified"
	VulnerabilityStatusClosed     VulnerabilityStatus = "closed"
)

// NewSecurityControls creates a new security controls instance
func NewSecurityControls(config *SecurityConfig, logger *zap.Logger) *SecurityControls {
	return &SecurityControls{
		logger: logger,
		config: config,
	}
}

// ValidatePassword validates a password against the password policy
func (sc *SecurityControls) ValidatePassword(password string) error {
	policy := sc.config.PasswordPolicy

	if len(password) < policy.MinLength {
		return fmt.Errorf("password must be at least %d characters long", policy.MinLength)
	}

	if policy.RequireUppercase && !hasUppercase(password) {
		return fmt.Errorf("password must contain at least one uppercase letter")
	}

	if policy.RequireLowercase && !hasLowercase(password) {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}

	if policy.RequireNumbers && !hasNumbers(password) {
		return fmt.Errorf("password must contain at least one number")
	}

	if policy.RequireSpecialChars && !hasSpecialChars(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}

// GenerateSecurePassword generates a secure password that meets the policy requirements
func (sc *SecurityControls) GenerateSecurePassword() (string, error) {
	policy := sc.config.PasswordPolicy

	// Generate a password that meets all requirements
	password := ""

	// Add required characters
	if policy.RequireUppercase {
		password += getRandomUppercase()
	}
	if policy.RequireLowercase {
		password += getRandomLowercase()
	}
	if policy.RequireNumbers {
		password += getRandomNumber()
	}
	if policy.RequireSpecialChars {
		password += getRandomSpecialChar()
	}

	// Fill remaining length with random characters
	remaining := policy.MinLength - len(password)
	for i := 0; i < remaining; i++ {
		password += getRandomChar()
	}

	// Shuffle the password
	shuffled := shuffleString(password)
	return shuffled, nil
}

// HashPassword securely hashes a password
func (sc *SecurityControls) HashPassword(password string) (string, error) {
	// Generate a random salt
	salt := make([]byte, 32)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Hash password with salt
	hash := sha256.Sum256(append([]byte(password), salt...))

	// Return hex-encoded salt + hash
	return hex.EncodeToString(salt) + hex.EncodeToString(hash[:]), nil
}

// VerifyPassword verifies a password against its hash
func (sc *SecurityControls) VerifyPassword(password, hash string) bool {
	if len(hash) < 64 { // 32 bytes salt + 32 bytes hash = 64 hex chars
		return false
	}

	// Extract salt and hash
	saltHex := hash[:64]
	hashHex := hash[64:]

	salt, err := hex.DecodeString(saltHex)
	if err != nil {
		return false
	}

	expectedHash, err := hex.DecodeString(hashHex)
	if err != nil {
		return false
	}

	// Compute hash
	computedHash := sha256.Sum256(append([]byte(password), salt...))

	// Compare hashes
	return string(computedHash[:]) == string(expectedHash)
}

// CheckAccessControl checks if a user has access to a resource
func (sc *SecurityControls) CheckAccessControl(ctx context.Context, userID, resource, action string) (bool, error) {
	if !sc.config.EnableAccessControl {
		return true, nil // Access control disabled
	}

	// In a real implementation, this would check against a database
	// For now, we'll implement a simple mock check
	sc.logger.Info("Checking access control",
		zap.String("user_id", userID),
		zap.String("resource", resource),
		zap.String("action", action))

	// Mock access control logic
	// In production, this would query the access control database
	return true, nil
}

// LogSecurityEvent logs a security-related event
func (sc *SecurityControls) LogSecurityEvent(ctx context.Context, eventType, userID, resource, action string, metadata map[string]interface{}) error {
	if !sc.config.EnableAuditLogging {
		return nil // Audit logging disabled
	}

	sc.logger.Info("Security event logged",
		zap.String("event_type", eventType),
		zap.String("user_id", userID),
		zap.String("resource", resource),
		zap.String("action", action),
		zap.Any("metadata", metadata))

	// In a real implementation, this would write to an audit log database
	return nil
}

// CreateSecurityIncident creates a new security incident
func (sc *SecurityControls) CreateSecurityIncident(ctx context.Context, incident *SecurityIncident) error {
	if !sc.config.EnableIncidentResponse {
		return fmt.Errorf("incident response is disabled")
	}

	// Generate incident ID if not provided
	if incident.ID == "" {
		incident.ID = generateIncidentID()
	}

	// Set timestamps
	now := time.Now()
	if incident.CreatedAt.IsZero() {
		incident.CreatedAt = now
	}
	incident.UpdatedAt = now

	sc.logger.Info("Security incident created",
		zap.String("incident_id", incident.ID),
		zap.String("type", string(incident.Type)),
		zap.String("severity", string(incident.Severity)),
		zap.String("title", incident.Title))

	// In a real implementation, this would save to a database
	return nil
}

// UpdateSecurityIncident updates an existing security incident
func (sc *SecurityControls) UpdateSecurityIncident(ctx context.Context, incidentID string, updates map[string]interface{}) error {
	if !sc.config.EnableIncidentResponse {
		return fmt.Errorf("incident response is disabled")
	}

	sc.logger.Info("Security incident updated",
		zap.String("incident_id", incidentID),
		zap.Any("updates", updates))

	// In a real implementation, this would update the database
	return nil
}

// GetSecurityIncidents retrieves security incidents
func (sc *SecurityControls) GetSecurityIncidents(ctx context.Context, filters map[string]interface{}) ([]*SecurityIncident, error) {
	if !sc.config.EnableIncidentResponse {
		return nil, fmt.Errorf("incident response is disabled")
	}

	// In a real implementation, this would query the database
	// For now, return empty list
	return []*SecurityIncident{}, nil
}

// CreateVulnerability creates a new vulnerability record
func (sc *SecurityControls) CreateVulnerability(ctx context.Context, vulnerability *Vulnerability) error {
	if !sc.config.EnableVulnerabilityMgmt {
		return fmt.Errorf("vulnerability management is disabled")
	}

	// Generate vulnerability ID if not provided
	if vulnerability.ID == "" {
		vulnerability.ID = generateVulnerabilityID()
	}

	// Set timestamps
	now := time.Now()
	if vulnerability.CreatedAt.IsZero() {
		vulnerability.CreatedAt = now
	}
	vulnerability.UpdatedAt = now

	sc.logger.Info("Vulnerability created",
		zap.String("vulnerability_id", vulnerability.ID),
		zap.String("cve", vulnerability.CVE),
		zap.String("severity", string(vulnerability.Severity)),
		zap.String("title", vulnerability.Title))

	// In a real implementation, this would save to a database
	return nil
}

// UpdateVulnerability updates an existing vulnerability
func (sc *SecurityControls) UpdateVulnerability(ctx context.Context, vulnerabilityID string, updates map[string]interface{}) error {
	if !sc.config.EnableVulnerabilityMgmt {
		return fmt.Errorf("vulnerability management is disabled")
	}

	sc.logger.Info("Vulnerability updated",
		zap.String("vulnerability_id", vulnerabilityID),
		zap.Any("updates", updates))

	// In a real implementation, this would update the database
	return nil
}

// GetVulnerabilities retrieves vulnerabilities
func (sc *SecurityControls) GetVulnerabilities(ctx context.Context, filters map[string]interface{}) ([]*Vulnerability, error) {
	if !sc.config.EnableVulnerabilityMgmt {
		return nil, fmt.Errorf("vulnerability management is disabled")
	}

	// In a real implementation, this would query the database
	// For now, return empty list
	return []*Vulnerability{}, nil
}

// EncryptData encrypts sensitive data
func (sc *SecurityControls) EncryptData(data string) (string, error) {
	if !sc.config.EnableEncryption {
		return data, nil // Encryption disabled
	}

	// In a real implementation, this would use proper encryption
	// For now, we'll use a simple hash-based approach
	hash := sha256.Sum256([]byte(data + sc.config.EncryptionKey))
	return hex.EncodeToString(hash[:]), nil
}

// DecryptData decrypts sensitive data
func (sc *SecurityControls) DecryptData(encryptedData string) (string, error) {
	if !sc.config.EnableEncryption {
		return encryptedData, nil // Encryption disabled
	}

	// In a real implementation, this would use proper decryption
	// For now, we'll return an error since we can't decrypt a hash
	return "", fmt.Errorf("decryption not implemented in mock")
}

// ClassifyData classifies data based on sensitivity
func (sc *SecurityControls) ClassifyData(data string) (DataClassification, error) {
	if !sc.config.EnableDataClassification {
		return DataClassificationPublic, nil
	}

	// Simple classification logic based on keywords
	// In production, this would use more sophisticated classification
	keywords := map[DataClassification][]string{
		DataClassificationConfidential: {"password", "secret", "private", "confidential"},
		DataClassificationInternal:     {"internal", "employee", "staff"},
		DataClassificationPublic:       {"public", "general"},
	}

	dataLower := strings.ToLower(data)
	for classification, words := range keywords {
		for _, word := range words {
			if strings.Contains(dataLower, word) {
				return classification, nil
			}
		}
	}

	return DataClassificationInternal, nil
}

// DataClassification represents data sensitivity levels
type DataClassification string

const (
	DataClassificationPublic       DataClassification = "public"
	DataClassificationInternal     DataClassification = "internal"
	DataClassificationConfidential DataClassification = "confidential"
	DataClassificationRestricted   DataClassification = "restricted"
)

// Helper functions
func hasUppercase(s string) bool {
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			return true
		}
	}
	return false
}

func hasLowercase(s string) bool {
	for _, r := range s {
		if r >= 'a' && r <= 'z' {
			return true
		}
	}
	return false
}

func hasNumbers(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

func hasSpecialChars(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, r := range s {
		if strings.ContainsRune(specialChars, r) {
			return true
		}
	}
	return false
}

func getRandomUppercase() string {
	return string(rune('A' + mathrand.Intn(26)))
}

func getRandomLowercase() string {
	return string(rune('a' + mathrand.Intn(26)))
}

func getRandomNumber() string {
	return string(rune('0' + mathrand.Intn(10)))
}

func getRandomSpecialChar() string {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	return string(specialChars[mathrand.Intn(len(specialChars))])
}

func getRandomChar() string {
	chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+-=[]{}|;:,.<>?"
	return string(chars[mathrand.Intn(len(chars))])
}

func shuffleString(s string) string {
	runes := []rune(s)
	for i := len(runes) - 1; i > 0; i-- {
		j := mathrand.Intn(i + 1)
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func generateIncidentID() string {
	return fmt.Sprintf("inc_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}

func generateVulnerabilityID() string {
	return fmt.Sprintf("vuln_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
