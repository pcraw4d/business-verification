package handlers

import (
	"fmt"
	"time"
)

// ValidationError represents a validation error
type ValidationError struct {
	Code       ErrorCode   `json:"code"`
	Message    string      `json:"message"`
	Field      string      `json:"field,omitempty"`
	Constraint string      `json:"constraint,omitempty"`
	Value      interface{} `json:"value,omitempty"`
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error in field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// AuthenticationError represents an authentication error
type AuthenticationError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *AuthenticationError) Error() string {
	return fmt.Sprintf("authentication error: %s", e.Message)
}

// AuthorizationError represents an authorization error
type AuthorizationError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *AuthorizationError) Error() string {
	return fmt.Sprintf("authorization error: %s", e.Message)
}

// RateLimitError represents a rate limit error
type RateLimitError struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	RetryAfter int       `json:"retry_after"`
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit error: %s (retry after %d seconds)", e.Message, e.RetryAfter)
}

// ClassificationError represents a classification error
type ClassificationError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *ClassificationError) Error() string {
	return fmt.Sprintf("classification error: %s", e.Message)
}

// ExternalServiceError represents an external service error
type ExternalServiceError struct {
	Code    ErrorCode `json:"code"`
	Service string    `json:"service"`
	Message string    `json:"message"`
}

func (e *ExternalServiceError) Error() string {
	return fmt.Sprintf("external service error (%s): %s", e.Service, e.Message)
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	Code      ErrorCode     `json:"code"`
	Operation string        `json:"operation"`
	Timeout   time.Duration `json:"timeout"`
	Message   string        `json:"message"`
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("timeout error: %s", e.Message)
}

// NotFoundError represents a not found error
type NotFoundError struct {
	Code     ErrorCode `json:"code"`
	Resource string    `json:"resource"`
	ID       string    `json:"id,omitempty"`
	Message  string    `json:"message"`
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("not found error: %s with ID '%s' not found", e.Resource, e.ID)
	}
	return fmt.Sprintf("not found error: %s", e.Message)
}

// ConflictError represents a conflict error
type ConflictError struct {
	Code     ErrorCode `json:"code"`
	Resource string    `json:"resource"`
	Message  string    `json:"message"`
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict error: %s", e.Message)
}

// DatabaseError represents a database error
type DatabaseError struct {
	Code      ErrorCode `json:"code"`
	Operation string    `json:"operation"`
	Message   string    `json:"message"`
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %s", e.Operation, e.Message)
}

// ServiceUnavailableError represents a service unavailable error
type ServiceUnavailableError struct {
	Code       ErrorCode `json:"code"`
	Service    string    `json:"service"`
	Message    string    `json:"message"`
	RetryAfter *int      `json:"retry_after,omitempty"`
}

func (e *ServiceUnavailableError) Error() string {
	return fmt.Sprintf("service unavailable error (%s): %s", e.Service, e.Message)
}

// GatewayTimeoutError represents a gateway timeout error
type GatewayTimeoutError struct {
	Code    ErrorCode     `json:"code"`
	Service string        `json:"service"`
	Timeout time.Duration `json:"timeout"`
	Message string        `json:"message"`
}

func (e *GatewayTimeoutError) Error() string {
	return fmt.Sprintf("gateway timeout error (%s): %s", e.Service, e.Message)
}

// VerificationError represents a verification error
type VerificationError struct {
	Code    ErrorCode `json:"code"`
	Type    string    `json:"type"`
	Message string    `json:"message"`
}

func (e *VerificationError) Error() string {
	return fmt.Sprintf("verification error (%s): %s", e.Type, e.Message)
}

// RiskAssessmentError represents a risk assessment error
type RiskAssessmentError struct {
	Code    ErrorCode `json:"code"`
	Type    string    `json:"type"`
	Message string    `json:"message"`
}

func (e *RiskAssessmentError) Error() string {
	return fmt.Sprintf("risk assessment error (%s): %s", e.Type, e.Message)
}

// DataExtractionError represents a data extraction error
type DataExtractionError struct {
	Code    ErrorCode `json:"code"`
	Source  string    `json:"source"`
	Message string    `json:"message"`
}

func (e *DataExtractionError) Error() string {
	return fmt.Sprintf("data extraction error (%s): %s", e.Source, e.Message)
}
