package middleware

import (
	"context"
	"net/http"
	"strings"

	"go.uber.org/zap"

	"kyb-platform/services/api-gateway/internal/supabase"
)

// Authentication middleware handles JWT token validation
func Authentication(supabaseClient *supabase.Client, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for health checks and public endpoints
			if isPublicEndpoint(r.URL.Path) {
				logger.Info("Skipping authentication for public endpoint",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method))
				next.ServeHTTP(w, r)
				return
			}

			logger.Info("Authentication required for endpoint",
				zap.String("path", r.URL.Path),
				zap.String("method", r.Method))

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// For now, allow requests without authentication
				// In production, you might want to require authentication
				next.ServeHTTP(w, r)
				return
			}

			// Check if it's a Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				logger.Warn("Invalid authorization header format",
					zap.String("path", r.URL.Path),
					zap.String("header", authHeader))
				// Set CORS headers before returning error
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
				return
			}

			// Extract the token
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate the token with Supabase
			user, err := supabaseClient.ValidateToken(r.Context(), token)
			if err != nil {
				logger.Warn("Token validation failed",
					zap.String("path", r.URL.Path),
					zap.Error(err))
				// Set CORS headers before returning error
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Add user information to context
			ctx := context.WithValue(r.Context(), "user", user)
			if userMap, ok := user.(map[string]interface{}); ok {
				if userID, exists := userMap["id"]; exists {
					ctx = context.WithValue(ctx, "user_id", userID)
				}
				if userEmail, exists := userMap["email"]; exists {
					ctx = context.WithValue(ctx, "user_email", userEmail)
				}
				// Add role to context
				if role, exists := userMap["role"]; exists {
					ctx = context.WithValue(ctx, "user_role", role)
				}
			}

			logger.Info("User authenticated",
				zap.String("path", r.URL.Path))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// isPublicEndpoint checks if the endpoint is public and doesn't require authentication
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/",                             // Root endpoint for debugging
		"/api/v1/classify",              // Classification endpoint is public for now
		"/api/v1/classification/health", // Classification health check
		"/api/v1/merchant/health",       // Merchant health check
		"/api/v1/merchants",             // Merchant endpoints for frontend
		"/api/v1/merchants/",            // Merchant endpoints with ID
		"/api/v1/merchants/search",      // Merchant search
		"/api/v1/merchants/analytics",   // Merchant analytics
		"/api/v1/risk",                  // Risk assessment endpoints
		"/api/v1/risk/",                 // Risk assessment endpoints with path
		"/api/v1/auth/register",         // Registration endpoint is public
	}

	for _, publicPath := range publicPaths {
		if path == publicPath {
			return true
		}
		// Handle dynamic paths like /api/v1/merchants/{id}
		if strings.HasPrefix(path, publicPath) && (publicPath == "/api/v1/merchants/" || publicPath == "/api/v1/risk/") {
			return true
		}
	}

	return false
}

// RequireAdmin is a middleware that requires admin role
func RequireAdmin(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user role from context
			role := r.Context().Value("user_role")
			if role == nil {
				logger.Warn("Admin access denied: no role in context",
					zap.String("path", r.URL.Path))
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Admin access required", http.StatusForbidden)
				return
			}

			// Check if role is admin
			roleStr := ""
			if r, ok := role.(string); ok {
				roleStr = r
			} else if r, ok := role.(interface{ String() string }); ok {
				roleStr = r.String()
			}

			if roleStr != "admin" && roleStr != "Admin" {
				logger.Warn("Admin access denied: insufficient privileges",
					zap.String("path", r.URL.Path),
					zap.String("role", roleStr))
				w.Header().Set("Access-Control-Allow-Origin", "*")
				http.Error(w, "Admin access required", http.StatusForbidden)
				return
			}

			logger.Info("Admin access granted",
				zap.String("path", r.URL.Path))

			next.ServeHTTP(w, r)
		})
	}
}
