package security

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTManager handles JWT token operations
type JWTManager struct {
	secretKey []byte
	issuer    string
	audience  string
}

// Claims represents JWT claims
type Claims struct {
	UserID      string   `json:"user_id"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions"`
	Role        string   `json:"role"`
	jwt.RegisteredClaims
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey, issuer, audience string) *JWTManager {
	return &JWTManager{
		secretKey: []byte(secretKey),
		issuer:    issuer,
		audience:  audience,
	}
}

// GenerateToken generates a new JWT token
func (jm *JWTManager) GenerateToken(userID, email, role string, permissions []string) (string, error) {
	now := time.Now()

	claims := &Claims{
		UserID:      userID,
		Email:       email,
		Permissions: permissions,
		Role:        role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    jm.issuer,
			Audience:  []string{jm.audience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jm.secretKey)
}

// ValidateToken validates a JWT token
func (jm *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jm.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Validate issuer and audience
		if claims.Issuer != jm.issuer {
			return nil, errors.New("invalid issuer")
		}

		if len(claims.Audience) == 0 || claims.Audience[0] != jm.audience {
			return nil, errors.New("invalid audience")
		}

		// Check if token is expired
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			return nil, errors.New("token expired")
		}

		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func (jm *JWTManager) ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}

	return parts[1], nil
}

// HasPermission checks if the user has a specific permission
func (c *Claims) HasPermission(permission string) bool {
	for _, p := range c.Permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// IsAdmin checks if the user has admin role
func (c *Claims) IsAdmin() bool {
	return c.Role == "admin"
}

// IsUser checks if the user has user role
func (c *Claims) IsUser() bool {
	return c.Role == "user"
}

// GetUserID returns the user ID from context
func GetUserID(ctx context.Context) (string, error) {
	if userID, ok := ctx.Value("user_id").(string); ok {
		return userID, nil
	}
	return "", errors.New("user ID not found in context")
}

// GetUserClaims returns the user claims from context
func GetUserClaims(ctx context.Context) (*Claims, error) {
	if claims, ok := ctx.Value("user_claims").(*Claims); ok {
		return claims, nil
	}
	return nil, errors.New("user claims not found in context")
}
