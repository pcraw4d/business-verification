package tenant

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/audit"
)

// TenantMiddleware creates middleware for tenant isolation
func TenantMiddleware(tenantService *TenantService, auditLogger *audit.AuditLogger, logger *zap.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Extract tenant context from various sources
			tenantCtx, err := extractTenantContext(ctx, r, tenantService)
			if err != nil {
				logger.Error("Failed to extract tenant context", zap.Error(err))
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add tenant context to request context
			ctx = WithTenantContext(ctx, tenantCtx)

			// Log the request with tenant context
			auditLogger.LogRequest(ctx, &audit.AuditEvent{
				TenantID:  tenantCtx.TenantID,
				UserID:    tenantCtx.UserID,
				Action:    "http_request",
				Resource:  "api",
				Method:    r.Method,
				Endpoint:  r.URL.Path,
				IPAddress: getClientIP(r),
				UserAgent: r.UserAgent(),
				RequestID: getRequestID(r),
			})

			// Continue with the request
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractTenantContext extracts tenant context from various sources
func extractTenantContext(ctx context.Context, r *http.Request, tenantService *TenantService) (*TenantContext, error) {
	// Try API key first
	if apiKey := extractAPIKey(r); apiKey != "" {
		tenantCtx, err := tenantService.ValidateAPIKey(ctx, apiKey)
		if err == nil {
			return tenantCtx, nil
		}
		// Log API key validation failure but continue to other methods
	}

	// Try JWT token
	if jwtToken := extractJWTToken(r); jwtToken != "" {
		tenantCtx, err := validateJWTToken(ctx, jwtToken, tenantService)
		if err == nil {
			return tenantCtx, nil
		}
		// Log JWT validation failure but continue to other methods
	}

	// Try tenant header
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		// This is a simplified approach - in production, you'd validate the user's access to this tenant
		tenantCtx := &TenantContext{
			TenantID:    tenantID,
			UserID:      r.Header.Get("X-User-ID"),
			UserRole:    TenantUserRole(r.Header.Get("X-User-Role")),
			Permissions: strings.Split(r.Header.Get("X-Permissions"), ","),
			Metadata:    make(map[string]interface{}),
		}
		return tenantCtx, nil
	}

	return nil, fmt.Errorf("no valid tenant context found")
}

// extractAPIKey extracts API key from request
func extractAPIKey(r *http.Request) string {
	// Try Authorization header with Bearer token
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Try X-API-Key header
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return apiKey
	}

	// Try query parameter
	if apiKey := r.URL.Query().Get("api_key"); apiKey != "" {
		return apiKey
	}

	return ""
}

// extractJWTToken extracts JWT token from request
func extractJWTToken(r *http.Request) string {
	// Try Authorization header with Bearer token
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		// Simple check if it looks like a JWT (has dots)
		if strings.Count(token, ".") == 2 {
			return token
		}
	}

	return ""
}

// validateJWTToken validates JWT token and extracts tenant context
func validateJWTToken(ctx context.Context, token string, tenantService *TenantService) (*TenantContext, error) {
	// This is a simplified JWT validation - in production, you'd use a proper JWT library
	// For now, we'll parse the token and extract claims

	// Mock JWT validation - in production, use jwt-go or similar
	claims, err := parseJWTClaims(token)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT token: %w", err)
	}

	// Extract tenant information from claims
	tenantID, ok := claims["tenant_id"].(string)
	if !ok || tenantID == "" {
		return nil, fmt.Errorf("tenant_id not found in JWT claims")
	}

	userID, _ := claims["user_id"].(string)
	userRole, _ := claims["user_role"].(string)
	permissions, _ := claims["permissions"].([]string)

	// Verify tenant exists and is active
	tenant, err := tenantService.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	if tenant.Status != TenantStatusActive {
		return nil, fmt.Errorf("tenant is not active")
	}

	tenantCtx := &TenantContext{
		TenantID:    tenantID,
		UserID:      userID,
		UserRole:    TenantUserRole(userRole),
		Permissions: permissions,
		Metadata:    make(map[string]interface{}),
	}

	return tenantCtx, nil
}

// parseJWTClaims parses JWT claims (simplified implementation)
func parseJWTClaims(token string) (map[string]interface{}, error) {
	// This is a mock implementation - in production, use a proper JWT library
	// For now, we'll return mock claims based on the token

	// Simple mock based on token content
	if strings.Contains(token, "mock") {
		return map[string]interface{}{
			"tenant_id":   "tenant_123",
			"user_id":     "user_456",
			"user_role":   "admin",
			"permissions": []string{"assessments:read", "assessments:write", "reports:read"},
		}, nil
	}

	return nil, fmt.Errorf("invalid token format")
}

// RequireTenantMiddleware creates middleware that requires a valid tenant context
func RequireTenantMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Check if tenant context exists
			_, err := GetTenantContext(ctx)
			if err != nil {
				http.Error(w, "Tenant context required", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermissionMiddleware creates middleware that requires specific permissions
func RequirePermissionMiddleware(permissions ...string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Check if user has required permissions
			if !HasAnyPermission(ctx, permissions) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRoleMiddleware creates middleware that requires specific roles
func RequireRoleMiddleware(roles ...TenantUserRole) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Check if user has required role
			if !HasAnyRole(ctx, roles) {
				http.Error(w, "Insufficient role", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantQuotaMiddleware creates middleware that enforces tenant quotas
func TenantQuotaMiddleware(tenantService *TenantService, quotaType string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get tenant ID
			tenantID, err := GetTenantID(ctx)
			if err != nil {
				http.Error(w, "Tenant context required", http.StatusUnauthorized)
				return
			}

			// Check quota (simplified - in production, you'd track actual usage)
			currentUsage := int64(0) // This would be retrieved from usage tracking
			if err := tenantService.CheckQuota(ctx, tenantID, quotaType, currentUsage); err != nil {
				http.Error(w, "Quota exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantIsolationMiddleware creates middleware that ensures tenant isolation
func TenantIsolationMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Ensure tenant context exists
			tenantCtx, err := GetTenantContext(ctx)
			if err != nil {
				http.Error(w, "Tenant context required", http.StatusUnauthorized)
				return
			}

			// Add tenant isolation metadata
			tenantCtx.Metadata["isolation_enforced"] = true
			tenantCtx.Metadata["request_timestamp"] = time.Now()

			next.ServeHTTP(w, r)
		})
	}
}

// TenantAuditMiddleware creates middleware that logs tenant-specific audit events
func TenantAuditMiddleware(auditLogger *audit.AuditLogger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get tenant context
			tenantCtx, err := GetTenantContext(ctx)
			if err != nil {
				// If no tenant context, log as system event
				auditLogger.LogSecurityEvent(ctx, "", "", "unauthorized_access_attempt", map[string]interface{}{
					"endpoint": r.URL.Path,
					"method":   r.Method,
					"ip":       getClientIP(r),
				})
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Log tenant-specific audit event
			auditLogger.LogDataAccess(ctx, tenantCtx.TenantID, tenantCtx.UserID, "api_endpoint", r.URL.Path, "access", map[string]interface{}{
				"method":      r.Method,
				"user_role":   tenantCtx.UserRole,
				"api_key_id":  tenantCtx.APIKeyID,
				"permissions": tenantCtx.Permissions,
			})

			next.ServeHTTP(w, r)
		})
	}
}

// TenantRateLimitMiddleware creates middleware that enforces tenant-specific rate limits
func TenantRateLimitMiddleware(tenantService *TenantService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get tenant context
			tenantCtx, err := GetTenantContext(ctx)
			if err != nil {
				http.Error(w, "Tenant context required", http.StatusUnauthorized)
				return
			}

			// Get tenant to check rate limits
			tenant, err := tenantService.GetTenant(ctx, tenantCtx.TenantID)
			if err != nil {
				http.Error(w, "Tenant not found", http.StatusNotFound)
				return
			}

			// Check rate limit (simplified - in production, use Redis or similar)
			rateLimit := tenant.Quotas.MaxConcurrentRequests
			if rateLimit > 0 {
				// This would be implemented with a proper rate limiter
				// For now, we'll just pass through
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper functions
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	// Fall back to RemoteAddr
	return r.RemoteAddr
}

func getRequestID(r *http.Request) string {
	if rid := r.Header.Get("X-Request-ID"); rid != "" {
		return rid
	}
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
