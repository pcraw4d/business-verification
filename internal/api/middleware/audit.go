package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pcraw4d/business-verification/internal/compliance"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// AuditMiddleware provides automatic audit logging for API requests
type AuditMiddleware struct {
	auditSystem *compliance.ComplianceAuditSystem
	logger      *observability.Logger
	enabled     bool
}

// NewAuditMiddleware creates a new audit middleware
func NewAuditMiddleware(auditSystem *compliance.ComplianceAuditSystem, logger *observability.Logger, enabled bool) *AuditMiddleware {
	return &AuditMiddleware{
		auditSystem: auditSystem,
		logger:      logger,
		enabled:     enabled,
	}
}

// AuditLoggingMiddleware wraps HTTP handlers to automatically log audit events
func (m *AuditMiddleware) AuditLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !m.enabled {
			next.ServeHTTP(w, r)
			return
		}

		start := time.Now()
		ctx := r.Context()

		// Extract user information from context (set by auth middleware)
		userID := extractUserID(ctx)
		userName := extractUserName(ctx)
		userRole := extractUserRole(ctx)
		userEmail := extractUserEmail(ctx)

		// Extract business ID from request
		businessID := extractBusinessID(r)

		// Create a custom response writer to capture status code
		responseWriter := &auditResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Process the request
		next.ServeHTTP(responseWriter, r)

		// Calculate duration
		duration := time.Since(start)

		// Determine event type based on request path and method
		eventType := determineEventType(r.Method, r.URL.Path)
		eventCategory := determineEventCategory(r.URL.Path)
		entityType := determineEntityType(r.URL.Path)
		entityID := extractEntityID(r.URL.Path)

		// Determine severity based on request characteristics
		severity := determineSeverity(r.Method, r.URL.Path, responseWriter.statusCode)
		impact := determineImpact(r.Method, r.URL.Path)

		// Create audit event
		event := &compliance.AuditEvent{
			ID:            generateAuditID(),
			BusinessID:    businessID,
			EventType:     eventType,
			EventCategory: eventCategory,
			EntityType:    entityType,
			EntityID:      entityID,
			Action:        determineAuditAction(r.Method),
			Description:   fmt.Sprintf("%s %s", r.Method, r.URL.Path),
			UserID:        userID,
			UserName:      userName,
			UserRole:      userRole,
			UserEmail:     userEmail,
			IPAddress:     extractIPAddress(r),
			UserAgent:     r.UserAgent(),
			SessionID:     extractSessionID(ctx),
			RequestID:     extractRequestID(ctx),
			Timestamp:     start,
			Duration:      duration,
			Success:       responseWriter.statusCode < 400,
			ErrorCode:     extractErrorCode(responseWriter.statusCode),
			ErrorMessage:  extractErrorMessage(responseWriter.statusCode),
			Metadata: map[string]interface{}{
				"method":     r.Method,
				"path":       r.URL.Path,
				"query":      r.URL.RawQuery,
				"status":     responseWriter.statusCode,
				"duration":   duration.String(),
				"user_agent": r.UserAgent(),
			},
			Severity: severity,
			Impact:   impact,
			Tags:     extractTags(r.URL.Path),
		}

		// Record the audit event asynchronously to avoid blocking the response
		go func() {
			if err := m.auditSystem.RecordAuditEvent(context.Background(), event); err != nil {
				m.logger.WithComponent("audit").LogAPIRequest(context.Background(), "AUDIT", "record_event", "system", http.StatusInternalServerError, time.Since(start))
			}
		}()
	})
}

// auditResponseWriter wraps http.ResponseWriter to capture status code
type auditResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *auditResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *auditResponseWriter) Write(data []byte) (int, error) {
	return w.ResponseWriter.Write(data)
}

// Helper functions for extracting information from requests

func extractUserID(ctx context.Context) string {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID
	}
	return "anonymous"
}

func extractUserName(ctx context.Context) string {
	if userName, ok := ctx.Value("user_name").(string); ok {
		return userName
	}
	return "Anonymous User"
}

func extractUserRole(ctx context.Context) string {
	if userRole, ok := ctx.Value("user_role").(string); ok {
		return userRole
	}
	return "guest"
}

func extractUserEmail(ctx context.Context) string {
	if userEmail, ok := ctx.Value("user_email").(string); ok {
		return userEmail
	}
	return "anonymous@example.com"
}

func extractBusinessID(r *http.Request) string {
	// Try to extract from query parameters
	if businessID := r.URL.Query().Get("business_id"); businessID != "" {
		return businessID
	}

	// Try to extract from path parameters
	if businessID := extractPathParam(r.URL.Path, "business_id"); businessID != "" {
		return businessID
	}

	// Try to extract from request body for POST/PUT requests
	if r.Method == "POST" || r.Method == "PUT" {
		if businessID := extractBusinessIDFromBody(r); businessID != "" {
			return businessID
		}
	}

	return "unknown"
}

func extractBusinessIDFromBody(r *http.Request) string {
	// Read the body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return ""
	}

	// Restore the body for other middleware/handlers
	r.Body = io.NopCloser(bytes.NewReader(body))

	// Try to parse as JSON
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return ""
	}

	// Extract business_id
	if businessID, ok := data["business_id"].(string); ok {
		return businessID
	}

	return ""
}

func extractPathParam(path, paramName string) string {
	// Simple path parameter extraction
	// This could be enhanced with a proper router
	// For now, we'll look for common patterns
	if paramName == "business_id" {
		// Look for patterns like /business/{business_id}/...
		// This is a simplified implementation
		return ""
	}
	return ""
}

func extractIPAddress(r *http.Request) string {
	// Check for forwarded headers first
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		return forwarded
	}
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}

func extractSessionID(ctx context.Context) string {
	if sessionID, ok := ctx.Value("session_id").(string); ok {
		return sessionID
	}
	return ""
}

func extractRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return ""
}

func extractErrorCode(statusCode int) string {
	if statusCode >= 400 {
		switch {
		case statusCode >= 500:
			return "server_error"
		case statusCode >= 400:
			return "client_error"
		default:
			return "unknown_error"
		}
	}
	return ""
}

func extractErrorMessage(statusCode int) string {
	if statusCode >= 400 {
		switch statusCode {
		case 400:
			return "Bad Request"
		case 401:
			return "Unauthorized"
		case 403:
			return "Forbidden"
		case 404:
			return "Not Found"
		case 500:
			return "Internal Server Error"
		default:
			return "Unknown Error"
		}
	}
	return ""
}

// Event type determination functions

func determineEventType(method, path string) string {
	switch {
	case path == "/v1/compliance/check":
		return "compliance_check"
	case path == "/v1/compliance/status":
		return "status_check"
	case path == "/v1/compliance/report":
		return "report_generation"
	case path == "/v1/compliance/audit":
		return "audit_access"
	case method == "GET":
		return "data_access"
	case method == "POST":
		return "data_creation"
	case method == "PUT":
		return "data_update"
	case method == "DELETE":
		return "data_deletion"
	default:
		return "api_request"
	}
}

func determineEventCategory(path string) string {
	switch {
	case contains(path, "/compliance/"):
		return "compliance"
	case contains(path, "/audit/"):
		return "audit"
	case contains(path, "/risk/"):
		return "risk"
	case contains(path, "/auth/"):
		return "authentication"
	case contains(path, "/admin/"):
		return "administration"
	default:
		return "general"
	}
}

func determineEntityType(path string) string {
	switch {
	case contains(path, "/compliance/"):
		return "compliance"
	case contains(path, "/audit/"):
		return "audit"
	case contains(path, "/risk/"):
		return "risk"
	case contains(path, "/auth/"):
		return "authentication"
	case contains(path, "/admin/"):
		return "administration"
	default:
		return "api"
	}
}

func extractEntityID(path string) string {
	// Extract entity ID from path
	// This is a simplified implementation
	return path
}

func determineAuditAction(method string) compliance.AuditAction {
	switch method {
	case "GET":
		return compliance.AuditActionRead
	case "POST":
		return compliance.AuditActionCreate
	case "PUT":
		return compliance.AuditActionUpdate
	case "DELETE":
		return compliance.AuditActionDelete
	default:
		return compliance.AuditActionRead
	}
}

func determineSeverity(method, path string, statusCode int) string {
	// High severity for admin operations
	if contains(path, "/admin/") {
		return "high"
	}

	// High severity for authentication operations
	if contains(path, "/auth/") {
		return "high"
	}

	// Medium severity for compliance operations
	if contains(path, "/compliance/") {
		return "medium"
	}

	// High severity for errors
	if statusCode >= 400 {
		return "high"
	}

	// Low severity for read operations
	if method == "GET" {
		return "low"
	}

	return "medium"
}

func determineImpact(method, path string) string {
	// Critical impact for admin operations
	if contains(path, "/admin/") {
		return "critical"
	}

	// High impact for compliance operations
	if contains(path, "/compliance/") {
		return "high"
	}

	// Medium impact for write operations
	if method == "POST" || method == "PUT" || method == "DELETE" {
		return "medium"
	}

	// Low impact for read operations
	return "low"
}

func extractTags(path string) []string {
	tags := []string{"api_request"}

	if contains(path, "/compliance/") {
		tags = append(tags, "compliance")
	}
	if contains(path, "/audit/") {
		tags = append(tags, "audit")
	}
	if contains(path, "/risk/") {
		tags = append(tags, "risk")
	}
	if contains(path, "/auth/") {
		tags = append(tags, "authentication")
	}
	if contains(path, "/admin/") {
		tags = append(tags, "administration")
	}

	return tags
}

// Utility functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func generateAuditID() string {
	return fmt.Sprintf("audit-%d", time.Now().UnixNano())
}
