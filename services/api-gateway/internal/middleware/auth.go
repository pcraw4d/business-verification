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
				next.ServeHTTP(w, r)
				return
			}

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
		"/", // Root endpoint for debugging
		"/api/v1/classify", // Classification endpoint is public for now
		"/api/v1/classification/health", // Classification health check
		"/api/v1/merchant/health", // Merchant health check
	}

	for _, publicPath := range publicPaths {
		if path == publicPath {
			return true
		}
	}

	return false
}
