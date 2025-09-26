package security

import (
	"context"
	"net/http"
	"strings"
)

// SecurityMiddleware provides security-related middleware functions
type SecurityMiddleware struct {
	jwtManager  *JWTManager
	rateLimiter *RateLimiter
	validator   *InputValidator
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(jwtManager *JWTManager, rateLimiter *RateLimiter, validator *InputValidator) *SecurityMiddleware {
	return &SecurityMiddleware{
		jwtManager:  jwtManager,
		rateLimiter: rateLimiter,
		validator:   validator,
	}
}

// JWTMiddleware validates JWT tokens
func (sm *SecurityMiddleware) JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip JWT validation for certain endpoints
		if sm.shouldSkipJWTValidation(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		tokenString, err := sm.jwtManager.ExtractTokenFromHeader(authHeader)
		if err != nil {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := sm.jwtManager.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user information to context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)
		ctx = context.WithValue(ctx, "user_role", claims.Role)
		ctx = context.WithValue(ctx, "user_permissions", claims.Permissions)
		ctx = context.WithValue(ctx, "user_claims", claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RateLimitMiddleware applies rate limiting per user
func (sm *SecurityMiddleware) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from context (set by JWT middleware)
		userID, ok := r.Context().Value("user_id").(string)
		if !ok {
			// If no user ID, use IP address as fallback
			userID = r.RemoteAddr
		}

		// Check rate limit
		allowed, err := sm.rateLimiter.Allow(r.Context(), userID)
		if err != nil {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		if !allowed {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// InputValidationMiddleware validates and sanitizes request inputs
func (sm *SecurityMiddleware) InputValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate query parameters for all requests
		for _, values := range r.URL.Query() {
			for _, value := range values {
				if sm.containsSuspiciousContent(value) {
					http.Error(w, "Invalid query parameter", http.StatusBadRequest)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeadersMiddleware adds security headers
func (sm *SecurityMiddleware) SecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		next.ServeHTTP(w, r)
	})
}

// PermissionMiddleware checks if user has required permissions
func (sm *SecurityMiddleware) PermissionMiddleware(requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user_claims").(*Claims)
			if !ok {
				http.Error(w, "User claims not found", http.StatusUnauthorized)
				return
			}

			if !claims.HasPermission(requiredPermission) {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RoleMiddleware checks if user has required role
func (sm *SecurityMiddleware) RoleMiddleware(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("user_claims").(*Claims)
			if !ok {
				http.Error(w, "User claims not found", http.StatusUnauthorized)
				return
			}

			if claims.Role != requiredRole {
				http.Error(w, "Insufficient role", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper methods
func (sm *SecurityMiddleware) shouldSkipJWTValidation(path string) bool {
	// Skip JWT validation for public endpoints
	publicPaths := []string{
		"/health",
		"/health/detailed",
		"/metrics",
		"/self-driving",
		"/debug/web",
		"/auth/token",
		"/auth/validate",
		"/v1/auth/token",
		"/v1/auth/validate",
		"/v2/auth/token",
		"/v2/auth/validate",
		"/docs",
		"/docs/openapi",
		"/v1/classify",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}

	return false
}

func (sm *SecurityMiddleware) containsSuspiciousContent(input string) bool {
	// Check for SQL injection patterns
	if sm.validator.sqlInjectionRegex.MatchString(input) {
		return true
	}

	// Check for XSS patterns
	if sm.validator.xssRegex.MatchString(input) {
		return true
	}

	return false
}
