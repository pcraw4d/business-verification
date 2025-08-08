package database

import (
	"context"
	"errors"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
)

// Common database errors
var (
	ErrUserNotFound           = errors.New("user not found")
	ErrRoleAssignmentNotFound = errors.New("role assignment not found")
	ErrAPIKeyNotFound         = errors.New("api key not found")
	ErrDuplicateUser          = errors.New("user already exists")
	ErrInvalidCredentials     = errors.New("invalid credentials")
)

// User represents a user in the system
type User struct {
	ID                  string     `json:"id" db:"id"`
	Email               string     `json:"email" db:"email"`
	Username            string     `json:"username" db:"username"`
	PasswordHash        string     `json:"-" db:"password_hash"`
	FirstName           string     `json:"first_name" db:"first_name"`
	LastName            string     `json:"last_name" db:"last_name"`
	Company             string     `json:"company" db:"company"`
	Role                string     `json:"role" db:"role"`
	Status              string     `json:"status" db:"status"`
	EmailVerified       bool       `json:"email_verified" db:"email_verified"`
	LastLoginAt         *time.Time `json:"last_login_at" db:"last_login_at"`
	FailedLoginAttempts int        `json:"failed_login_attempts" db:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until" db:"locked_until"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// Business represents a business entity
type Business struct {
	ID                 string      `json:"id" db:"id"`
	Name               string      `json:"name" db:"name"`
	LegalName          string      `json:"legal_name" db:"legal_name"`
	RegistrationNumber string      `json:"registration_number" db:"registration_number"`
	TaxID              string      `json:"tax_id" db:"tax_id"`
	Industry           string      `json:"industry" db:"industry"`
	IndustryCode       string      `json:"industry_code" db:"industry_code"`
	BusinessType       string      `json:"business_type" db:"business_type"`
	FoundedDate        *time.Time  `json:"founded_date" db:"founded_date"`
	EmployeeCount      int         `json:"employee_count" db:"employee_count"`
	AnnualRevenue      *float64    `json:"annual_revenue" db:"annual_revenue"`
	Address            Address     `json:"address" db:"address"`
	ContactInfo        ContactInfo `json:"contact_info" db:"contact_info"`
	Status             string      `json:"status" db:"status"`
	RiskLevel          string      `json:"risk_level" db:"risk_level"`
	ComplianceStatus   string      `json:"compliance_status" db:"compliance_status"`
	CreatedBy          string      `json:"created_by" db:"created_by"`
	CreatedAt          time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time   `json:"updated_at" db:"updated_at"`
}

// Address represents a business address
type Address struct {
	Street1     string `json:"street1" db:"street1"`
	Street2     string `json:"street2" db:"street2"`
	City        string `json:"city" db:"city"`
	State       string `json:"state" db:"state"`
	PostalCode  string `json:"postal_code" db:"postal_code"`
	Country     string `json:"country" db:"country"`
	CountryCode string `json:"country_code" db:"country_code"`
}

// ContactInfo represents business contact information
type ContactInfo struct {
	Phone          string `json:"phone" db:"phone"`
	Email          string `json:"email" db:"email"`
	Website        string `json:"website" db:"website"`
	PrimaryContact string `json:"primary_contact" db:"primary_contact"`
}

// BusinessClassification represents a business classification result
type BusinessClassification struct {
	ID                   string    `json:"id" db:"id"`
	BusinessID           string    `json:"business_id" db:"business_id"`
	IndustryCode         string    `json:"industry_code" db:"industry_code"`
	IndustryName         string    `json:"industry_name" db:"industry_name"`
	ConfidenceScore      float64   `json:"confidence_score" db:"confidence_score"`
	ClassificationMethod string    `json:"classification_method" db:"classification_method"`
	Source               string    `json:"source" db:"source"`
	RawData              string    `json:"raw_data" db:"raw_data"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
}

// RiskAssessment represents a risk assessment result
type RiskAssessment struct {
	ID               string    `json:"id" db:"id"`
	BusinessID       string    `json:"business_id" db:"business_id"`
	RiskLevel        string    `json:"risk_level" db:"risk_level"`
	RiskScore        float64   `json:"risk_score" db:"risk_score"`
	RiskFactors      []string  `json:"risk_factors" db:"risk_factors"`
	AssessmentMethod string    `json:"assessment_method" db:"assessment_method"`
	Source           string    `json:"source" db:"source"`
	RawData          string    `json:"raw_data" db:"raw_data"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// ComplianceCheck represents a compliance check result
type ComplianceCheck struct {
	ID             string    `json:"id" db:"id"`
	BusinessID     string    `json:"business_id" db:"business_id"`
	ComplianceType string    `json:"compliance_type" db:"compliance_type"`
	Status         string    `json:"status" db:"status"`
	Score          float64   `json:"score" db:"score"`
	Requirements   []string  `json:"requirements" db:"requirements"`
	CheckMethod    string    `json:"check_method" db:"check_method"`
	Source         string    `json:"source" db:"source"`
	RawData        string    `json:"raw_data" db:"raw_data"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// APIKey represents an API key for external integrations
type APIKey struct {
	ID          string     `json:"id" db:"id"`
	UserID      string     `json:"user_id" db:"user_id"`
	Name        string     `json:"name" db:"name"`
	KeyHash     string     `json:"-" db:"key_hash"`
	Role        string     `json:"role" db:"role"`
	Permissions string     `json:"permissions" db:"permissions"` // JSON array as string
	Status      string     `json:"status" db:"status"`
	LastUsedAt  *time.Time `json:"last_used_at" db:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// RoleAssignment represents a user's role assignment with audit trail
type RoleAssignment struct {
	ID         string     `json:"id" db:"id"`
	UserID     string     `json:"user_id" db:"user_id"`
	Role       string     `json:"role" db:"role"`
	AssignedBy string     `json:"assigned_by" db:"assigned_by"` // User ID who assigned the role
	AssignedAt time.Time  `json:"assigned_at" db:"assigned_at"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"` // Optional role expiration
	IsActive   bool       `json:"is_active" db:"is_active"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// AuditLog represents an audit log entry
type AuditLog struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	Action       string    `json:"action" db:"action"`
	ResourceType string    `json:"resource_type" db:"resource_type"`
	ResourceID   string    `json:"resource_id" db:"resource_id"`
	Details      string    `json:"details" db:"details"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	RequestID    string    `json:"request_id" db:"request_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// ExternalServiceCall represents an external service API call
type ExternalServiceCall struct {
	ID           string    `json:"id" db:"id"`
	UserID       string    `json:"user_id" db:"user_id"`
	ServiceName  string    `json:"service_name" db:"service_name"`
	Endpoint     string    `json:"endpoint" db:"endpoint"`
	Method       string    `json:"method" db:"method"`
	RequestData  string    `json:"request_data" db:"request_data"`
	ResponseData string    `json:"response_data" db:"response_data"`
	StatusCode   int       `json:"status_code" db:"status_code"`
	Duration     int64     `json:"duration_ms" db:"duration_ms"`
	Error        string    `json:"error" db:"error"`
	RequestID    string    `json:"request_id" db:"request_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Webhook represents a webhook configuration
type Webhook struct {
	ID              string     `json:"id" db:"id"`
	UserID          string     `json:"user_id" db:"user_id"`
	Name            string     `json:"name" db:"name"`
	URL             string     `json:"url" db:"url"`
	Events          []string   `json:"events" db:"events"`
	Secret          string     `json:"-" db:"secret"`
	Status          string     `json:"status" db:"status"`
	LastTriggeredAt *time.Time `json:"last_triggered_at" db:"last_triggered_at"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// WebhookEvent represents a webhook event
type WebhookEvent struct {
	ID           string     `json:"id" db:"id"`
	WebhookID    string     `json:"webhook_id" db:"webhook_id"`
	EventType    string     `json:"event_type" db:"event_type"`
	Payload      string     `json:"payload" db:"payload"`
	Status       string     `json:"status" db:"status"`
	ResponseCode *int       `json:"response_code" db:"response_code"`
	ResponseBody string     `json:"response_body" db:"response_body"`
	Attempts     int        `json:"attempts" db:"attempts"`
	NextRetryAt  *time.Time `json:"next_retry_at" db:"next_retry_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Token     string     `json:"token" db:"token"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at" db:"used_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        string     `json:"id" db:"id"`
	UserID    string     `json:"user_id" db:"user_id"`
	Token     string     `json:"token" db:"token"`
	ExpiresAt time.Time  `json:"expires_at" db:"expires_at"`
	UsedAt    *time.Time `json:"used_at" db:"used_at"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// TokenBlacklist represents a blacklisted JWT token
type TokenBlacklist struct {
	ID            string    `json:"id" db:"id"`
	TokenID       string    `json:"token_id" db:"token_id"`
	UserID        *string   `json:"user_id" db:"user_id"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
	BlacklistedAt time.Time `json:"blacklisted_at" db:"blacklisted_at"`
	Reason        string    `json:"reason" db:"reason"`
}

// Database represents the database interface
type Database interface {
	// Connection management
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error

	// User management
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)

	// Email verification management
	CreateEmailVerificationToken(ctx context.Context, token *EmailVerificationToken) error
	GetEmailVerificationToken(ctx context.Context, token string) (*EmailVerificationToken, error)
	MarkEmailVerificationTokenUsed(ctx context.Context, token string) error
	DeleteExpiredEmailVerificationTokens(ctx context.Context) error

	// Password reset management
	CreatePasswordResetToken(ctx context.Context, token *PasswordResetToken) error
	GetPasswordResetToken(ctx context.Context, token string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, token string) error
	DeleteExpiredPasswordResetTokens(ctx context.Context) error

	// Token blacklist management
	CreateTokenBlacklist(ctx context.Context, blacklist *TokenBlacklist) error
	IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error)
	DeleteExpiredTokenBlacklist(ctx context.Context) error

	// Business management
	CreateBusiness(ctx context.Context, business *Business) error
	GetBusinessByID(ctx context.Context, id string) (*Business, error)
	GetBusinessByRegistrationNumber(ctx context.Context, regNumber string) (*Business, error)
	UpdateBusiness(ctx context.Context, business *Business) error
	DeleteBusiness(ctx context.Context, id string) error
	ListBusinesses(ctx context.Context, limit, offset int) ([]*Business, error)
	SearchBusinesses(ctx context.Context, query string, limit, offset int) ([]*Business, error)

	// Classification management
	CreateBusinessClassification(ctx context.Context, classification *BusinessClassification) error
	GetBusinessClassificationByID(ctx context.Context, id string) (*BusinessClassification, error)
	GetBusinessClassificationsByBusinessID(ctx context.Context, businessID string) ([]*BusinessClassification, error)
	UpdateBusinessClassification(ctx context.Context, classification *BusinessClassification) error
	DeleteBusinessClassification(ctx context.Context, id string) error

	// Risk assessment management
	CreateRiskAssessment(ctx context.Context, assessment *RiskAssessment) error
	GetRiskAssessmentByID(ctx context.Context, id string) (*RiskAssessment, error)
	GetRiskAssessmentsByBusinessID(ctx context.Context, businessID string) ([]*RiskAssessment, error)
	UpdateRiskAssessment(ctx context.Context, assessment *RiskAssessment) error
	DeleteRiskAssessment(ctx context.Context, id string) error

	// Compliance management
	CreateComplianceCheck(ctx context.Context, check *ComplianceCheck) error
	GetComplianceCheckByID(ctx context.Context, id string) (*ComplianceCheck, error)
	GetComplianceChecksByBusinessID(ctx context.Context, businessID string) ([]*ComplianceCheck, error)
	UpdateComplianceCheck(ctx context.Context, check *ComplianceCheck) error
	DeleteComplianceCheck(ctx context.Context, id string) error

	// API key management
	CreateAPIKey(ctx context.Context, apiKey *APIKey) error
	GetAPIKeyByID(ctx context.Context, id string) (*APIKey, error)
	GetAPIKeyByHash(ctx context.Context, keyHash string) (*APIKey, error)
	UpdateAPIKey(ctx context.Context, apiKey *APIKey) error
	DeleteAPIKey(ctx context.Context, id string) error
	ListAPIKeysByUserID(ctx context.Context, userID string) ([]*APIKey, error)

	// Audit log management
	CreateAuditLog(ctx context.Context, log *AuditLog) error
	GetAuditLogsByUserID(ctx context.Context, userID string, limit, offset int) ([]*AuditLog, error)
	GetAuditLogsByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*AuditLog, error)

	// External service call management
	CreateExternalServiceCall(ctx context.Context, call *ExternalServiceCall) error
	GetExternalServiceCallsByUserID(ctx context.Context, userID string, limit, offset int) ([]*ExternalServiceCall, error)
	GetExternalServiceCallsByService(ctx context.Context, serviceName string, limit, offset int) ([]*ExternalServiceCall, error)

	// Webhook management
	CreateWebhook(ctx context.Context, webhook *Webhook) error
	GetWebhookByID(ctx context.Context, id string) (*Webhook, error)
	GetWebhooksByUserID(ctx context.Context, userID string) ([]*Webhook, error)
	UpdateWebhook(ctx context.Context, webhook *Webhook) error
	DeleteWebhook(ctx context.Context, id string) error

	// Webhook event management
	CreateWebhookEvent(ctx context.Context, event *WebhookEvent) error
	GetWebhookEventByID(ctx context.Context, id string) (*WebhookEvent, error)
	GetWebhookEventsByWebhookID(ctx context.Context, webhookID string, limit, offset int) ([]*WebhookEvent, error)
	UpdateWebhookEvent(ctx context.Context, event *WebhookEvent) error
	DeleteWebhookEvent(ctx context.Context, id string) error

	// Role assignment management
	CreateRoleAssignment(ctx context.Context, assignment *RoleAssignment) error
	GetRoleAssignmentByID(ctx context.Context, id string) (*RoleAssignment, error)
	GetActiveRoleAssignmentByUserID(ctx context.Context, userID string) (*RoleAssignment, error)
	GetRoleAssignmentsByUserID(ctx context.Context, userID string) ([]*RoleAssignment, error)
	UpdateRoleAssignment(ctx context.Context, assignment *RoleAssignment) error
	DeactivateRoleAssignment(ctx context.Context, id string) error
	DeleteExpiredRoleAssignments(ctx context.Context) error

	// Enhanced API key management with RBAC
	UpdateAPIKeyLastUsed(ctx context.Context, id string, lastUsed time.Time) error
	GetActiveAPIKeysByRole(ctx context.Context, role string) ([]*APIKey, error)
	DeactivateAPIKey(ctx context.Context, id string) error

	// Transaction support
	BeginTx(ctx context.Context) (Database, error)
	Commit() error
	Rollback() error
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	*config.DatabaseConfig
}

// NewDatabaseConfig creates a new database configuration
func NewDatabaseConfig(cfg *config.DatabaseConfig) *DatabaseConfig {
	return &DatabaseConfig{
		DatabaseConfig: cfg,
	}
}
