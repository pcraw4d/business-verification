package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/pcraw4d/business-verification/internal/auth"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	authService *auth.AuthService
	logger      *observability.Logger
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *auth.AuthService, logger *observability.Logger) *AuthMiddleware {
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
			m.logger.WithComponent("auth").Info("Missing authorization header", "path", r.URL.Path, "method", r.Method)
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token (remove "Bearer " prefix)
		tokenString := ""
		if len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		} else {
			m.logger.WithComponent("auth").Info("Invalid authorization header format", "path", r.URL.Path, "method", r.Method)
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Validate token
		user, err := m.authService.ValidateToken(ctx, tokenString)
		if err != nil {
			m.logger.WithComponent("auth").Info("Invalid token", "error", err, "path", r.URL.Path, "method", r.Method)
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
				m.logger.WithComponent("auth").Error("User role not found in context", "path", r.URL.Path, "method", r.Method)
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if user has required role
			if userRole != role && userRole != "admin" { // admin has access to everything
				userID, _ := ctx.Value("user_id").(string)
				m.logger.WithComponent("auth").Info("Insufficient permissions", "user_id", userID, "required_role", role, "user_role", userRole, "path", r.URL.Path, "method", r.Method)
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
			m.logger.WithComponent("auth").Error("User ID not found in context", "path", r.URL.Path, "method", r.Method)
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Get user details
		user, err := m.authService.GetUserByID(ctx, userID)
		if err != nil {
			m.logger.WithComponent("auth").Error("Failed to get user", "error", err, "user_id", userID, "path", r.URL.Path, "method", r.Method)
			http.Error(w, "Failed to verify user", http.StatusInternalServerError)
			return
		}

		// Check if email is verified
		if !user.EmailVerified {
			m.logger.WithComponent("auth").Info("Email not verified", "user_id", userID, "email", user.Email, "path", r.URL.Path, "method", r.Method)
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
