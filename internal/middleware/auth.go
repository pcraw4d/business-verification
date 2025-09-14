package middleware

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret    string
	APIKeySecret string
	TokenExpiry  time.Duration
	RequireAuth  bool
}

// Claims represents JWT claims
type Claims struct {
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	Issuer    string `json:"iss"`
}

// AuthMiddleware handles authentication for API endpoints
type AuthMiddleware struct {
	config *AuthConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config *AuthConfig) *AuthMiddleware {
	if config.TokenExpiry == 0 {
		config.TokenExpiry = 24 * time.Hour // Default 24 hours
	}
	return &AuthMiddleware{
		config: config,
	}
}

// RequireAuth middleware that validates JWT tokens or API keys
func (a *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health checks and public endpoints
		if a.isPublicEndpoint(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Check for API key in header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			if a.validateAPIKey(apiKey) {
				// Add API key context
				ctx := context.WithValue(r.Context(), "auth_type", "api_key")
				ctx = context.WithValue(ctx, "api_key", apiKey)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// Check for JWT token in Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if claims, err := a.ValidateJWT(token); err == nil {
				// Add user context
				ctx := context.WithValue(r.Context(), "auth_type", "jwt")
				ctx = context.WithValue(ctx, "user_id", claims.UserID)
				ctx = context.WithValue(ctx, "user_role", claims.Role)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		// If authentication is required and no valid token/key found
		if a.config.RequireAuth {
			a.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required")
			return
		}

		// Allow request to proceed without authentication
		next.ServeHTTP(w, r)
	})
}

// isPublicEndpoint checks if the endpoint should be publicly accessible
func (a *AuthMiddleware) isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/health",
		"/",
		"/business-intelligence.html",
		"/merchant-hub.html",
		"/merchant-detail.html",
		"/merchant-portfolio.html",
		"/v1/classify", // Public classification endpoint
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return true
		}
	}
	return false
}

// validateAPIKey validates the provided API key
func (a *AuthMiddleware) validateAPIKey(apiKey string) bool {
	if a.config.APIKeySecret == "" {
		return false
	}

	// For production, you would validate against a database
	// For now, we'll use a simple secret-based validation
	expectedKey := a.generateAPIKey("default")
	return hmac.Equal([]byte(apiKey), []byte(expectedKey))
}

// ValidateJWT validates and parses a JWT token
func (a *AuthMiddleware) ValidateJWT(token string) (*Claims, error) {
	if a.config.JWTSecret == "" {
		return nil, fmt.Errorf("JWT secret not configured")
	}

	// Split token into parts
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode header and payload
	_, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid header: %v", err)
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid payload: %v", err)
	}

	// Verify signature
	signature := parts[2]
	expectedSignature := a.signToken(parts[0] + "." + parts[1])
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		return nil, fmt.Errorf("invalid signature")
	}

	// Parse claims
	var claims Claims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("invalid claims: %v", err)
	}

	// Check expiration
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	return &claims, nil
}

// signToken creates HMAC signature for JWT
func (a *AuthMiddleware) signToken(data string) string {
	h := hmac.New(sha256.New, []byte(a.config.JWTSecret))
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// generateAPIKey generates a new API key
func (a *AuthMiddleware) generateAPIKey(identifier string) string {
	data := fmt.Sprintf("%s:%d", identifier, time.Now().Unix())
	h := hmac.New(sha256.New, []byte(a.config.APIKeySecret))
	h.Write([]byte(data))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// writeErrorResponse writes an error response
func (a *AuthMiddleware) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]interface{}{
		"error":   true,
		"message": message,
		"code":    statusCode,
	}

	json.NewEncoder(w).Encode(response)
}

// GenerateToken generates a new JWT token for a user
func (a *AuthMiddleware) GenerateToken(userID, role string) (string, error) {
	if a.config.JWTSecret == "" {
		return "", fmt.Errorf("JWT secret not configured")
	}

	now := time.Now()
	claims := Claims{
		UserID:    userID,
		Role:      role,
		IssuedAt:  now.Unix(),
		ExpiresAt: now.Add(a.config.TokenExpiry).Unix(),
		Issuer:    "kyb-platform",
	}

	// Encode header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, _ := json.Marshal(header)
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Encode payload
	payloadJSON, _ := json.Marshal(claims)
	payloadEncoded := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	data := headerEncoded + "." + payloadEncoded
	signature := a.signToken(data)

	return data + "." + signature, nil
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(ctx context.Context) (userID, role string, authType string) {
	if authType, ok := ctx.Value("auth_type").(string); ok {
		if authType == "jwt" {
			if userID, ok := ctx.Value("user_id").(string); ok {
				if role, ok := ctx.Value("user_role").(string); ok {
					return userID, role, authType
				}
			}
		} else if authType == "api_key" {
			return "api_user", "api", authType
		}
	}
	return "", "", ""
}
