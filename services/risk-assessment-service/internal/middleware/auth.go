package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret        string         `json:"jwt_secret"`
	JWTPublicKey     *rsa.PublicKey `json:"jwt_public_key"`
	APIKeyHeader     string         `json:"api_key_header"`
	ServiceToken     string         `json:"service_token"`
	TokenCacheTTL    time.Duration  `json:"token_cache_ttl"`
	RequireAuth      bool           `json:"require_auth"`
	AllowServiceAuth bool           `json:"allow_service_auth"`
	AllowAPIKeyAuth  bool           `json:"allow_api_key_auth"`
	AllowJWTAuth     bool           `json:"allow_jwt_auth"`
}

// AuthMiddleware handles authentication for the risk assessment service
type AuthMiddleware struct {
	config      AuthConfig
	redisClient *redis.Client
	logger      *zap.Logger
}

// UserContext holds user information from authentication
type UserContext struct {
	UserID    string `json:"user_id"`
	TenantID  string `json:"tenant_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	APIKeyID  string `json:"api_key_id,omitempty"`
	ServiceID string `json:"service_id,omitempty"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// APIKeyInfo represents API key information
type APIKeyInfo struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	TenantID string    `json:"tenant_id"`
	Role     string    `json:"role"`
	Tier     string    `json:"tier"`
	Expires  time.Time `json:"expires"`
	Active   bool      `json:"active"`
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(config AuthConfig, redisClient *redis.Client, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		config:      config,
		redisClient: redisClient,
		logger:      logger,
	}
}

// AuthMiddleware creates a middleware function for authentication
func (am *AuthMiddleware) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip authentication for health checks and metrics
			if am.shouldSkipAuth(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Extract authentication information
			userCtx, err := am.authenticateRequest(r)
			if err != nil {
				am.logger.Warn("Authentication failed",
					zap.String("path", r.URL.Path),
					zap.String("method", r.Method),
					zap.String("remote_addr", r.RemoteAddr),
					zap.Error(err))

				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user context to request
			ctx := context.WithValue(r.Context(), "user", userCtx)
			ctx = context.WithValue(ctx, "user_id", userCtx.UserID)
			ctx = context.WithValue(ctx, "tenant_id", userCtx.TenantID)
			ctx = context.WithValue(ctx, "role", userCtx.Role)

			// Add headers for downstream services
			r.Header.Set("X-User-ID", userCtx.UserID)
			r.Header.Set("X-Tenant-ID", userCtx.TenantID)
			r.Header.Set("X-User-Role", userCtx.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// authenticateRequest authenticates the incoming request
func (am *AuthMiddleware) authenticateRequest(r *http.Request) (*UserContext, error) {
	// Try different authentication methods in order of preference

	// 1. Service-to-service authentication
	if am.config.AllowServiceAuth {
		if userCtx, err := am.authenticateService(r); err == nil {
			return userCtx, nil
		}
	}

	// 2. API Key authentication
	if am.config.AllowAPIKeyAuth {
		if userCtx, err := am.authenticateAPIKey(r); err == nil {
			return userCtx, nil
		}
	}

	// 3. JWT authentication
	if am.config.AllowJWTAuth {
		if userCtx, err := am.authenticateJWT(r); err == nil {
			return userCtx, nil
		}
	}

	// If no authentication method succeeded and auth is required
	if am.config.RequireAuth {
		return nil, fmt.Errorf("no valid authentication method found")
	}

	// Return anonymous user context if auth is not required
	return &UserContext{
		UserID:   "anonymous",
		TenantID: "default",
		Role:     "guest",
	}, nil
}

// authenticateService authenticates service-to-service requests
func (am *AuthMiddleware) authenticateService(r *http.Request) (*UserContext, error) {
	serviceToken := r.Header.Get("X-Service-Token")
	if serviceToken == "" {
		return nil, fmt.Errorf("service token not provided")
	}

	if serviceToken != am.config.ServiceToken {
		return nil, fmt.Errorf("invalid service token")
	}

	serviceID := r.Header.Get("X-Service-ID")
	if serviceID == "" {
		serviceID = "unknown-service"
	}

	return &UserContext{
		UserID:    "service",
		TenantID:  "system",
		Role:      "service",
		ServiceID: serviceID,
	}, nil
}

// authenticateAPIKey authenticates API key requests
func (am *AuthMiddleware) authenticateAPIKey(r *http.Request) (*UserContext, error) {
	apiKey := r.Header.Get(am.config.APIKeyHeader)
	if apiKey == "" {
		return nil, fmt.Errorf("API key not provided")
	}

	// Check cache first
	if am.redisClient != nil {
		if userCtx, err := am.getAPIKeyFromCache(apiKey); err == nil {
			return userCtx, nil
		}
	}

	// Validate API key (in a real implementation, this would query the database)
	apiKeyInfo, err := am.validateAPIKey(apiKey)
	if err != nil {
		return nil, fmt.Errorf("invalid API key: %w", err)
	}

	userCtx := &UserContext{
		UserID:   apiKeyInfo.UserID,
		TenantID: apiKeyInfo.TenantID,
		Role:     apiKeyInfo.Role,
		APIKeyID: apiKeyInfo.ID,
	}

	// Cache the result
	if am.redisClient != nil {
		am.cacheAPIKey(apiKey, userCtx)
	}

	return userCtx, nil
}

// authenticateJWT authenticates JWT token requests
func (am *AuthMiddleware) authenticateJWT(r *http.Request) (*UserContext, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, fmt.Errorf("authorization header not provided")
	}

	// Extract token from "Bearer <token>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, fmt.Errorf("invalid authorization header format")
	}

	tokenString := parts[1]

	// Parse and validate JWT token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return secret key
		return []byte(am.config.JWTSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	// Validate claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid JWT claims")
	}

	// Check token expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, fmt.Errorf("JWT token expired")
	}

	return &UserContext{
		UserID:   claims.UserID,
		TenantID: claims.TenantID,
		Email:    claims.Email,
		Role:     claims.Role,
	}, nil
}

// validateAPIKey validates an API key (placeholder implementation)
func (am *AuthMiddleware) validateAPIKey(apiKey string) (*APIKeyInfo, error) {
	// In a real implementation, this would query the database
	// For now, we'll use a simple validation based on key format

	if len(apiKey) < 32 {
		return nil, fmt.Errorf("API key too short")
	}

	// Simple tier detection based on prefix
	var tier string
	if strings.HasPrefix(apiKey, "pro_") {
		tier = "pro"
	} else if strings.HasPrefix(apiKey, "ent_") {
		tier = "enterprise"
	} else if strings.HasPrefix(apiKey, "int_") {
		tier = "internal"
	} else {
		tier = "free"
	}

	// Generate a deterministic user/tenant ID based on the key
	// In reality, this would come from the database
	userID := fmt.Sprintf("user_%s", apiKey[:8])
	tenantID := fmt.Sprintf("tenant_%s", apiKey[8:16])

	return &APIKeyInfo{
		ID:       apiKey,
		UserID:   userID,
		TenantID: tenantID,
		Role:     "user",
		Tier:     tier,
		Expires:  time.Now().Add(365 * 24 * time.Hour), // 1 year
		Active:   true,
	}, nil
}

// getAPIKeyFromCache retrieves API key information from cache
func (am *AuthMiddleware) getAPIKeyFromCache(apiKey string) (*UserContext, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("api_key:%s", apiKey)
	data, err := am.redisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var userCtx UserContext
	if err := json.Unmarshal([]byte(data), &userCtx); err != nil {
		return nil, err
	}

	return &userCtx, nil
}

// cacheAPIKey caches API key information
func (am *AuthMiddleware) cacheAPIKey(apiKey string, userCtx *UserContext) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := fmt.Sprintf("api_key:%s", apiKey)
	data, err := json.Marshal(userCtx)
	if err != nil {
		am.logger.Error("Failed to marshal user context", zap.Error(err))
		return
	}

	am.redisClient.Set(ctx, key, data, am.config.TokenCacheTTL)
}

// shouldSkipAuth determines if authentication should be skipped for a path
func (am *AuthMiddleware) shouldSkipAuth(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/api/v1/health",
		"/api/v1/metrics",
		"/api/v1/status",
		"/docs",
		"/swagger",
		"/openapi.json",
	}

	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}

	return false
}

// RequireRole creates a middleware that requires a specific role
func (am *AuthMiddleware) RequireRole(requiredRole string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := r.Context().Value("user").(*UserContext)
			if !ok {
				http.Error(w, "User context not found", http.StatusUnauthorized)
				return
			}

			if !am.hasRole(userCtx.Role, requiredRole) {
				am.logger.Warn("Insufficient permissions",
					zap.String("user_id", userCtx.UserID),
					zap.String("user_role", userCtx.Role),
					zap.String("required_role", requiredRole))

				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// hasRole checks if a user role has the required permissions
func (am *AuthMiddleware) hasRole(userRole, requiredRole string) bool {
	// Role hierarchy (higher roles have permissions of lower roles)
	roleHierarchy := map[string]int{
		"guest":      0,
		"user":       1,
		"analyst":    2,
		"admin":      3,
		"service":    4,
		"superadmin": 5,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}

// RequireTenant creates a middleware that requires a specific tenant
func (am *AuthMiddleware) RequireTenant(tenantID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, ok := r.Context().Value("user").(*UserContext)
			if !ok {
				http.Error(w, "User context not found", http.StatusUnauthorized)
				return
			}

			if userCtx.TenantID != tenantID && userCtx.Role != "superadmin" {
				am.logger.Warn("Tenant access denied",
					zap.String("user_id", userCtx.UserID),
					zap.String("user_tenant", userCtx.TenantID),
					zap.String("required_tenant", tenantID))

				http.Error(w, "Tenant access denied", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(ctx context.Context) (*UserContext, bool) {
	userCtx, ok := ctx.Value("user").(*UserContext)
	return userCtx, ok
}
