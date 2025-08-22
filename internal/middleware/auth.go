package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret     string
	APIKeyHeader  string
	APIKeys       map[string]string // API key to user mapping
	RequireAuth   bool
	ExemptPaths   []string // Paths that don't require authentication
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	ClientID string `json:"client_id"`
	jwt.RegisteredClaims
}

// AuthMiddleware provides authentication for API endpoints
func AuthMiddleware(config AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path is exempt from authentication
			if isExemptPath(r.URL.Path, config.ExemptPaths) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check for Bearer token or API key
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if err := validateJWTToken(token, config.JWTSecret); err != nil {
					http.Error(w, "Invalid JWT token", http.StatusUnauthorized)
					return
				}
			} else if strings.HasPrefix(authHeader, "ApiKey ") {
				apiKey := strings.TrimPrefix(authHeader, "ApiKey ")
				if userID, ok := config.APIKeys[apiKey]; !ok {
					http.Error(w, "Invalid API key", http.StatusUnauthorized)
					return
				} else {
					// Add user ID to context
					ctx := context.WithValue(r.Context(), "user_id", userID)
					r = r.WithContext(ctx)
				}
			} else {
				http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// validateJWTToken validates a JWT token
func validateJWTToken(tokenString, secret string) error {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Check if token is expired
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return fmt.Errorf("token expired")
		}
		return nil
	}

	return fmt.Errorf("invalid token")
}

// isExemptPath checks if a path is exempt from authentication
func isExemptPath(path string, exemptPaths []string) bool {
	for _, exemptPath := range exemptPaths {
		if strings.HasPrefix(path, exemptPath) {
			return true
		}
	}
	return false
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}

// GenerateJWTToken generates a new JWT token
func GenerateJWTToken(userID, email, role, clientID, secret string, expiration time.Duration) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Email:    email,
		Role:     role,
		ClientID: clientID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
