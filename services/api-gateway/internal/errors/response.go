package errors

import (
	"encoding/json"
	"net/http"
	"time"
)

// WriteError writes a standardized error response
func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, code, message, details string) {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = r.Header.Get("X-Correlation-ID")
	}

	errorResponse := ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:     r.URL.Path,
		Method:   r.Method,
	}

	w.Header().Set("Content-Type", "application/json")
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResponse)
}

// WriteBadRequest writes a bad request error response
func WriteBadRequest(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusBadRequest, "BAD_REQUEST", message, "")
}

// WriteUnauthorized writes an unauthorized error response
func WriteUnauthorized(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusUnauthorized, "UNAUTHORIZED", message, "")
}

// WriteForbidden writes a forbidden error response
func WriteForbidden(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusForbidden, "FORBIDDEN", message, "")
}

// WriteNotFound writes a not found error response
func WriteNotFound(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusNotFound, "NOT_FOUND", message, "")
}

// WriteConflict writes a conflict error response
func WriteConflict(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusConflict, "CONFLICT", message, "")
}

// WriteInternalError writes an internal server error response
func WriteInternalError(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", message)
}

// WriteServiceUnavailable writes a service unavailable error response
func WriteServiceUnavailable(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message, "")
}

// WriteTooManyRequests writes a rate limit error response
func WriteTooManyRequests(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", message, "")
}

// WriteMethodNotAllowed writes a method not allowed error response
func WriteMethodNotAllowed(w http.ResponseWriter, r *http.Request, message string) {
	WriteError(w, r, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", message, "")
}

// ErrorResponse represents a standardized error response format
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
	Path      string      `json:"path,omitempty"`
	Method    string      `json:"method,omitempty"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	Field      string            `json:"field,omitempty"`
	Validation []ValidationError `json:"validation,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

