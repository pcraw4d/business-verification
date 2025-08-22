package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pcraw4d/business-verification/internal/auth"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// PermissionMiddleware handles RBAC permission checking for API endpoints
type PermissionMiddleware struct {
	rbacService auth.RBACServiceInterface
	logger      *observability.Logger
}

// NewPermissionMiddleware creates a new permission middleware instance
func NewPermissionMiddleware(rbacService auth.RBACServiceInterface, logger *observability.Logger) *PermissionMiddleware {
	return &PermissionMiddleware{
		rbacService: rbacService,
		logger:      logger,
	}
}

// Middleware function that checks permissions for each request
func (pm *PermissionMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip permission checking for public endpoints
		if pm.isPublicEndpoint(r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract user context from request
		userID, userRole, err := pm.extractUserContext(r)
		if err != nil {
			pm.logger.WithComponent("permission_middleware").WithError(err).Warn("Failed to extract user context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Check if user has permission for this endpoint
		if err := pm.rbacService.CheckPermission(r.Context(), userRole, r.Method, r.URL.Path); err != nil {
			pm.logger.WithComponent("permission_middleware").WithFields(map[string]interface{}{
				"user_id": userID,
				"role":    userRole,
				"method":  r.Method,
				"path":    r.URL.Path,
			}).Warn("Permission denied")

			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Add user context to request for downstream handlers
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "user_role", userRole)
		r = r.WithContext(ctx)

		// Log successful permission check
		pm.logger.WithComponent("permission_middleware").WithFields(map[string]interface{}{
			"user_id": userID,
			"role":    userRole,
			"method":  r.Method,
			"path":    r.URL.Path,
		}).Debug("Permission check passed")

		next.ServeHTTP(w, r)
	})
}

// isPublicEndpoint checks if an endpoint is publicly accessible
func (pm *PermissionMiddleware) isPublicEndpoint(method, path string) bool {
	publicEndpoints := map[string][]string{
		"GET": {
			"/health",
			"/v1/health",
			"/docs",
			"/docs/",
			"/v1/docs",
			"/v1/docs/",
		},
		"POST": {
			"/v1/auth/login",
			"/v1/auth/register",
			"/v1/auth/refresh",
			"/v1/auth/forgot-password",
			"/v1/auth/reset-password",
		},
	}

	if endpoints, exists := publicEndpoints[method]; exists {
		for _, endpoint := range endpoints {
			if strings.HasPrefix(path, endpoint) {
				return true
			}
		}
	}

	return false
}

// extractUserContext extracts user information from the request
func (pm *PermissionMiddleware) extractUserContext(r *http.Request) (string, auth.Role, error) {
	// First try to get from JWT token in Authorization header
	if userID, userRole, err := pm.extractFromJWT(r); err == nil {
		return userID, userRole, nil
	}

	// Then try to get from API key
	if userID, userRole, err := pm.extractFromAPIKey(r); err == nil {
		return userID, userRole, nil
	}

	// Finally, check for system/internal requests
	if userID, userRole, err := pm.extractFromSystemContext(r); err == nil {
		return userID, userRole, nil
	}

	return "", "", fmt.Errorf("no valid authentication found")
}

// extractFromJWT extracts user context from JWT token
func (pm *PermissionMiddleware) extractFromJWT(r *http.Request) (string, auth.Role, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", "", fmt.Errorf("no authorization header")
	}

	// Check for Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", "", fmt.Errorf("invalid authorization header format")
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// TODO: Implement JWT token validation and user extraction
	// For now, return a placeholder implementation
	// In a real implementation, you would:
	// 1. Validate the JWT token
	// 2. Extract user ID and role from claims
	// 3. Verify token hasn't expired
	// 4. Check if token is blacklisted

	// Placeholder implementation
	if token == "admin-token" {
		return "admin-user", auth.RoleAdmin, nil
	}
	if token == "user-token" {
		return "user-1", auth.RoleUser, nil
	}
	if token == "analyst-token" {
		return "analyst-1", auth.RoleAnalyst, nil
	}
	if token == "manager-token" {
		return "manager-1", auth.RoleManager, nil
	}
	if token == "guest-token" {
		return "guest-1", auth.RoleGuest, nil
	}

	return "", "", fmt.Errorf("invalid token")
}

// extractFromAPIKey extracts user context from API key
func (pm *PermissionMiddleware) extractFromAPIKey(r *http.Request) (string, auth.Role, error) {
	apiKey := r.Header.Get("X-API-Key")
	if apiKey == "" {
		return "", "", fmt.Errorf("no API key provided")
	}

	// TODO: Implement API key validation
	// In a real implementation, you would:
	// 1. Look up the API key in the database
	// 2. Check if it's active and not expired
	// 3. Extract user ID and role from the API key record
	// 4. Update last used timestamp

	// Placeholder implementation
	if apiKey == "system-api-key" {
		return "system-user", auth.RoleSystem, nil
	}

	return "", "", fmt.Errorf("invalid API key")
}

// extractFromSystemContext extracts user context from system/internal requests
func (pm *PermissionMiddleware) extractFromSystemContext(r *http.Request) (string, auth.Role, error) {
	// Check for internal system headers
	systemToken := r.Header.Get("X-System-Token")
	if systemToken == "internal-system" {
		return "system-internal", auth.RoleSystem, nil
	}

	// Check for health check or monitoring requests
	userAgent := r.Header.Get("User-Agent")
	if strings.Contains(userAgent, "health-check") || strings.Contains(userAgent, "monitoring") {
		return "monitoring-system", auth.RoleSystem, nil
	}

	return "", "", fmt.Errorf("not a system request")
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}

// GetUserRoleFromContext extracts user role from request context
func GetUserRoleFromContext(ctx context.Context) (auth.Role, bool) {
	userRole, ok := ctx.Value("user_role").(auth.Role)
	return userRole, ok
}

// RequirePermission creates a middleware that requires a specific permission
func (pm *PermissionMiddleware) RequirePermission(permission auth.Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRoleFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !auth.HasPermission(userRole, permission) {
				pm.logger.WithComponent("permission_middleware").WithFields(map[string]interface{}{
					"user_role":  userRole,
					"permission": permission,
					"method":     r.Method,
					"path":       r.URL.Path,
				}).Warn("Permission denied for specific permission check")

				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole creates a middleware that requires a specific role
func (pm *PermissionMiddleware) RequireRole(role auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRoleFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if userRole != role {
				pm.logger.WithComponent("permission_middleware").WithFields(map[string]interface{}{
					"user_role": userRole,
					"required":  role,
					"method":    r.Method,
					"path":      r.URL.Path,
				}).Warn("Role access denied")

				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireMinimumRole creates a middleware that requires at least a specific role
func (pm *PermissionMiddleware) RequireMinimumRole(minRole auth.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := GetUserRoleFromContext(r.Context())
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if !pm.hasMinimumRole(userRole, minRole) {
				pm.logger.WithComponent("permission_middleware").WithFields(map[string]interface{}{
					"user_role":    userRole,
					"minimum_role": minRole,
					"method":       r.Method,
					"path":         r.URL.Path,
				}).Warn("Minimum role requirement not met")

				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasMinimumRole checks if a user role meets the minimum requirement
func (pm *PermissionMiddleware) hasMinimumRole(userRole, minRole auth.Role) bool {
	roleHierarchy := map[auth.Role]int{
		auth.RoleGuest:   0,
		auth.RoleUser:    1,
		auth.RoleAnalyst: 2,
		auth.RoleManager: 3,
		auth.RoleAdmin:   4,
		auth.RoleSystem:  5,
	}

	userLevel, userExists := roleHierarchy[userRole]
	minLevel, minExists := roleHierarchy[minRole]

	if !userExists || !minExists {
		return false
	}

	return userLevel >= minLevel
}
