package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pcraw4d/business-verification/internal/auth"
	"go.uber.org/zap"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	authService *auth.AuthService
	logger      *zap.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *auth.AuthService, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
		logger:      logger,
	}
}

// RequireAuth middleware requires a valid JWT token
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Info("Missing authorization header",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token (remove "Bearer " prefix)
		tokenString := ""
		if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		} else {
			m.logger.Info("Invalid authorization header format",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Validate token
		user, err := m.authService.ValidateToken(ctx, tokenString)
		if err != nil {
			m.logger.Info("Invalid token",
				zap.Error(err),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user information to context
		ctx = context.WithValue(ctx, "user_id", user.ID)
		ctx = context.WithValue(ctx, "user_email", user.Email)
		ctx = context.WithValue(ctx, "user_role", user.Role)

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole middleware requires a specific role
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get user role from context (set by RequireAuth middleware)
			userRole, ok := ctx.Value("user_role").(string)
			if !ok {
				m.logger.Error("User role not found in context",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user has required role
			if userRole != role && userRole != "admin" { // admin has access to everything
				userID, _ := ctx.Value("user_id").(string)
				m.logger.Info("Insufficient permissions",
					zap.String("user_id", userID),
					zap.String("required_role", role),
					zap.String("user_role", userRole),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// RequireEmailVerified middleware requires email verification
func (m *AuthMiddleware) RequireEmailVerified(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get user ID from context (set by RequireAuth middleware)
		userID, ok := ctx.Value("user_id").(string)
		if !ok {
			m.logger.Error("User ID not found in context",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Get user details
		user, err := m.authService.GetUserByID(ctx, userID)
		if err != nil {
			m.logger.Error("Failed to get user",
				zap.Error(err),
				zap.String("user_id", userID),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Failed to verify user", http.StatusInternalServerError)
			return
		}

		// Check if email is verified
		if !user.EmailVerified {
			m.logger.Info("Email not verified",
				zap.String("user_id", userID),
				zap.String("email", user.Email),
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Email verification required", http.StatusForbidden)
			return
		}

		// Continue to next handler
		next.ServeHTTP(w, r)
	})
}

// OptionalAuth middleware adds user information to context if token is provided
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			// Extract token (remove "Bearer " prefix)
			if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString := authHeader[7:]

				// Validate token
				user, err := m.authService.ValidateToken(ctx, tokenString)
				if err == nil {
					// Add user information to context
					ctx = context.WithValue(ctx, "user_id", user.ID)
					ctx = context.WithValue(ctx, "user_email", user.Email)
					ctx = context.WithValue(ctx, "user_role", user.Role)
				}
			}
		}

		// Continue to next handler (with or without user context)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequirePermission middleware requires a specific permission
func (m *AuthMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Get user role from context (set by RequireAuth middleware)
			userRole, ok := ctx.Value("user_role").(string)
			if !ok {
				m.logger.Error("User role not found in context",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user has required permission based on role
			hasPermission := m.checkPermission(userRole, permission)
			if !hasPermission {
				userID, _ := ctx.Value("user_id").(string)
				m.logger.Info("Insufficient permissions",
					zap.String("user_id", userID),
					zap.String("required_permission", permission),
					zap.String("user_role", userRole),
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// checkPermission checks if a role has a specific permission
func (m *AuthMiddleware) checkPermission(role, permission string) bool {
	// Define role-based permissions
	permissions := map[string][]string{
		"admin": {
			"read:all",
			"write:all",
			"delete:all",
			"admin:all",
			"user:manage",
			"system:manage",
		},
		"manager": {
			"read:all",
			"write:all",
			"user:read",
			"user:write",
		},
		"user": {
			"read:own",
			"write:own",
			"profile:manage",
		},
		"viewer": {
			"read:own",
		},
	}

	// Check if role has the required permission
	rolePermissions, exists := permissions[role]
	if !exists {
		return false
	}

	for _, perm := range rolePermissions {
		if perm == permission {
			return true
		}
	}

	return false
}

// RequireAPIKey middleware requires a valid API key
func (m *AuthMiddleware) RequireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get API key from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Info("Missing API key",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "API key required", http.StatusUnauthorized)
			return
		}

		// Extract API key (remove "ApiKey " prefix)
		apiKey := ""
		if len(authHeader) > 7 && strings.HasPrefix(authHeader, "ApiKey ") {
			apiKey = authHeader[7:]
		} else {
			// Try without prefix
			apiKey = authHeader
		}

		if apiKey == "" {
			m.logger.Info("Invalid API key format",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Invalid API key format", http.StatusUnauthorized)
			return
		}

		// Validate API key (simplified implementation)
		// In a real system, you would validate against a database
		if !m.isValidAPIKey(apiKey) {
			m.logger.Info("Invalid API key",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// Add API key information to context
		ctx = context.WithValue(ctx, "api_key", apiKey)
		ctx = context.WithValue(ctx, "auth_type", "api_key")

		// Continue to next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// isValidAPIKey validates an API key (simplified implementation)
func (m *AuthMiddleware) isValidAPIKey(apiKey string) bool {
	// In a real implementation, you would check against a database
	// For now, we'll accept any non-empty API key that starts with "test-"
	return len(apiKey) > 0 && strings.HasPrefix(apiKey, "test-")
}

// RequireAnyAuth middleware requires either JWT token or API key
func (m *AuthMiddleware) RequireAnyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Info("Missing authorization header",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))
			http.Error(w, "Authorization required", http.StatusUnauthorized)
			return
		}

		// Try JWT token first
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString := authHeader[7:]
			user, err := m.authService.ValidateToken(ctx, tokenString)
			if err == nil {
				// Add user information to context
				ctx = context.WithValue(ctx, "user_id", user.ID)
				ctx = context.WithValue(ctx, "user_email", user.Email)
				ctx = context.WithValue(ctx, "user_role", user.Role)
				ctx = context.WithValue(ctx, "auth_type", "jwt")
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Try API key
		apiKey := authHeader
		if strings.HasPrefix(authHeader, "ApiKey ") {
			apiKey = authHeader[7:]
		}

		if m.isValidAPIKey(apiKey) {
			// Add API key information to context
			ctx = context.WithValue(ctx, "api_key", apiKey)
			ctx = context.WithValue(ctx, "auth_type", "api_key")
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// No valid authentication found
		m.logger.Info("No valid authentication found",
			zap.String("path", r.URL.Path),
			zap.String("method", r.Method))
		http.Error(w, "Invalid authentication", http.StatusUnauthorized)
	})
}
