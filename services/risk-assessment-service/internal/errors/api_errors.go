package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// APIError represents a standardized API error response
type APIError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *APIError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Error code constants
const (
	ErrCodeNotFound      = "NOT_FOUND"
	ErrCodeValidation    = "VALIDATION_ERROR"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeForbidden     = "FORBIDDEN"
	ErrCodeBadRequest    = "BAD_REQUEST"
	ErrCodeConflict      = "CONFLICT"
	ErrCodeRateLimited   = "RATE_LIMITED"
	ErrCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// WriteAPIError writes a standardized error response to the HTTP response writer
func WriteAPIError(w http.ResponseWriter, code string, message string, statusCode int, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	apiError := &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}

	if err := json.NewEncoder(w).Encode(apiError); err != nil {
		// Fallback if encoding fails
		http.Error(w, message, statusCode)
	}
}

// getErrorCode maps internal errors to API error codes
func getErrorCode(err error) string {
	errStr := err.Error()

	// Map common error patterns to error codes
	if contains(errStr, "not found") || contains(errStr, "does not exist") {
		return ErrCodeNotFound
	}
	if contains(errStr, "validation") || contains(errStr, "invalid") {
		return ErrCodeValidation
	}
	if contains(errStr, "unauthorized") || contains(errStr, "authentication") {
		return ErrCodeUnauthorized
	}
	if contains(errStr, "forbidden") || contains(errStr, "permission") {
		return ErrCodeForbidden
	}
	if contains(errStr, "conflict") || contains(errStr, "already exists") {
		return ErrCodeConflict
	}

	return ErrCodeInternal
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// RetryWithBackoff retries a function with exponential backoff
func RetryWithBackoff(ctx context.Context, maxRetries int, initialDelay time.Duration, fn func() error) error {
	var lastError error
	delay := initialDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute function
		err := fn()
		if err == nil {
			return nil
		}

		lastError = err

		// Don't retry on last attempt
		if attempt < maxRetries-1 {
			// Wait with exponential backoff
			select {
			case <-time.After(delay):
				delay *= 2 // Exponential backoff
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return fmt.Errorf("retry failed after %d attempts: %w", maxRetries, lastError)
}

