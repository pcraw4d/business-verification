package security

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthManager provides JWT-based authentication and authorization
type AuthManager struct {
	logger     *zap.Logger
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	config     AuthConfig
	tokenCache map[string]*TokenInfo
}

// AuthConfig holds configuration for authentication
type AuthConfig struct {
	TokenExpiry       time.Duration `json:"token_expiry"`
	RefreshExpiry     time.Duration `json:"refresh_expiry"`
	Issuer            string        `json:"issuer"`
	Audience          string        `json:"audience"`
	Algorithm         string        `json:"algorithm"`
	EnableTokenCache  bool          `json:"enable_token_cache"`
	CacheExpiry       time.Duration `json:"cache_expiry"`
	MaxLoginAttempts  int           `json:"max_login_attempts"`
	LockoutDuration   time.Duration `json:"lockout_duration"`
	PasswordMinLength int           `json:"password_min_length"`
	RequireStrongPass bool          `json:"require_strong_password"`
}

// TokenInfo holds information about a JWT token
type TokenInfo struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	IsActive    bool      `json:"is_active"`
}

// Claims represents JWT claims
type Claims struct {
	UserID      string   `json:"user_id"`
	Username    string   `json:"username"`
	Email       string   `json:"email"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int64     `json:"expires_in"`
	TokenType    string    `json:"token_type"`
	User         *UserInfo `json:"user"`
}

// UserInfo represents user information
type UserInfo struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	Roles       []string  `json:"roles"`
	Permissions []string  `json:"permissions"`
	LastLogin   time.Time `json:"last_login"`
	IsActive    bool      `json:"is_active"`
}

// NewAuthManager creates a new authentication manager
func NewAuthManager(logger *zap.Logger, config AuthConfig) (*AuthManager, error) {
	// Generate RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKey := &privateKey.PublicKey

	return &AuthManager{
		logger:     logger,
		privateKey: privateKey,
		publicKey:  publicKey,
		config:     config,
		tokenCache: make(map[string]*TokenInfo),
	}, nil
}

// GenerateTokenPair generates access and refresh tokens
func (am *AuthManager) GenerateTokenPair(ctx context.Context, userInfo *UserInfo) (*LoginResponse, error) {
	now := time.Now()

	// Create access token claims
	accessClaims := &Claims{
		UserID:      userInfo.ID,
		Username:    userInfo.Username,
		Email:       userInfo.Email,
		Roles:       userInfo.Roles,
		Permissions: userInfo.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    am.config.Issuer,
			Audience:  []string{am.config.Audience},
			Subject:   userInfo.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(am.config.TokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Create refresh token claims
	refreshClaims := &Claims{
		UserID:      userInfo.ID,
		Username:    userInfo.Username,
		Email:       userInfo.Email,
		Roles:       userInfo.Roles,
		Permissions: userInfo.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    am.config.Issuer,
			Audience:  []string{am.config.Audience},
			Subject:   userInfo.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(am.config.RefreshExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(am.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(am.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Cache token info if enabled
	if am.config.EnableTokenCache {
		tokenInfo := &TokenInfo{
			UserID:      userInfo.ID,
			Username:    userInfo.Username,
			Email:       userInfo.Email,
			Roles:       userInfo.Roles,
			Permissions: userInfo.Permissions,
			IssuedAt:    now,
			ExpiresAt:   now.Add(am.config.TokenExpiry),
			IsActive:    true,
		}
		am.tokenCache[accessTokenString] = tokenInfo
	}

	response := &LoginResponse{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(am.config.TokenExpiry.Seconds()),
		TokenType:    "Bearer",
		User:         userInfo,
	}

	am.logger.Info("Token pair generated successfully",
		zap.String("user_id", userInfo.ID),
		zap.String("username", userInfo.Username))

	return response, nil
}

// ValidateToken validates a JWT token and returns claims
func (am *AuthManager) ValidateToken(ctx context.Context, tokenString string) (*Claims, error) {
	// Check cache first if enabled
	if am.config.EnableTokenCache {
		if tokenInfo, exists := am.tokenCache[tokenString]; exists {
			if !tokenInfo.IsActive || time.Now().After(tokenInfo.ExpiresAt) {
				delete(am.tokenCache, tokenString)
				return nil, fmt.Errorf("token expired or inactive")
			}

			// Return cached claims
			return &Claims{
				UserID:      tokenInfo.UserID,
				Username:    tokenInfo.Username,
				Email:       tokenInfo.Email,
				Roles:       tokenInfo.Roles,
				Permissions: tokenInfo.Permissions,
				RegisteredClaims: jwt.RegisteredClaims{
					Subject:   tokenInfo.UserID,
					IssuedAt:  jwt.NewNumericDate(tokenInfo.IssuedAt),
					ExpiresAt: jwt.NewNumericDate(tokenInfo.ExpiresAt),
				},
			}, nil
		}
	}

	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return am.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Validate claims
	if claims.ExpiresAt != nil && time.Now().After(claims.ExpiresAt.Time) {
		return nil, fmt.Errorf("token expired")
	}

	if claims.NotBefore != nil && time.Now().Before(claims.NotBefore.Time) {
		return nil, fmt.Errorf("token not yet valid")
	}

	// Check issuer and audience
	if claims.Issuer != am.config.Issuer {
		return nil, fmt.Errorf("invalid issuer")
	}

	if len(claims.Audience) == 0 || claims.Audience[0] != am.config.Audience {
		return nil, fmt.Errorf("invalid audience")
	}

	return claims, nil
}

// RefreshToken generates a new access token using a refresh token
func (am *AuthManager) RefreshToken(ctx context.Context, refreshTokenString string) (*LoginResponse, error) {
	// Validate refresh token
	claims, err := am.ValidateToken(ctx, refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Create new user info from claims
	userInfo := &UserInfo{
		ID:          claims.UserID,
		Username:    claims.Username,
		Email:       claims.Email,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		LastLogin:   time.Now(),
		IsActive:    true,
	}

	// Generate new token pair
	return am.GenerateTokenPair(ctx, userInfo)
}

// RevokeToken revokes a token
func (am *AuthManager) RevokeToken(ctx context.Context, tokenString string) error {
	if am.config.EnableTokenCache {
		delete(am.tokenCache, tokenString)
	}

	am.logger.Info("Token revoked", zap.String("token", tokenString[:20]+"..."))
	return nil
}

// HasPermission checks if a user has a specific permission
func (am *AuthManager) HasPermission(claims *Claims, permission string) bool {
	for _, userPermission := range claims.Permissions {
		if userPermission == permission {
			return true
		}
	}
	return false
}

// HasRole checks if a user has a specific role
func (am *AuthManager) HasRole(claims *Claims, role string) bool {
	for _, userRole := range claims.Roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// HasAnyRole checks if a user has any of the specified roles
func (am *AuthManager) HasAnyRole(claims *Claims, roles []string) bool {
	for _, requiredRole := range roles {
		if am.HasRole(claims, requiredRole) {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if a user has any of the specified permissions
func (am *AuthManager) HasAnyPermission(claims *Claims, permissions []string) bool {
	for _, requiredPermission := range permissions {
		if am.HasPermission(claims, requiredPermission) {
			return true
		}
	}
	return false
}

// ValidatePassword validates password strength
func (am *AuthManager) ValidatePassword(password string) error {
	if len(password) < am.config.PasswordMinLength {
		return fmt.Errorf("password must be at least %d characters long", am.config.PasswordMinLength)
	}

	if am.config.RequireStrongPass {
		// Check for uppercase letter
		hasUpper := false
		// Check for lowercase letter
		hasLower := false
		// Check for digit
		hasDigit := false
		// Check for special character
		hasSpecial := false

		for _, char := range password {
			switch {
			case 'A' <= char && char <= 'Z':
				hasUpper = true
			case 'a' <= char && char <= 'z':
				hasLower = true
			case '0' <= char && char <= '9':
				hasDigit = true
			case char < 32 || char > 126:
				// Non-printable characters
			default:
				hasSpecial = true
			}
		}

		if !hasUpper {
			return fmt.Errorf("password must contain at least one uppercase letter")
		}
		if !hasLower {
			return fmt.Errorf("password must contain at least one lowercase letter")
		}
		if !hasDigit {
			return fmt.Errorf("password must contain at least one digit")
		}
		if !hasSpecial {
			return fmt.Errorf("password must contain at least one special character")
		}
	}

	return nil
}

// GetPublicKey returns the public key for token verification
func (am *AuthManager) GetPublicKey() *rsa.PublicKey {
	return am.publicKey
}

// GetPublicKeyPEM returns the public key in PEM format
func (am *AuthManager) GetPublicKeyPEM() (string, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(am.publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public key: %w", err)
	}

	pubKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	})

	return string(pubKeyPEM), nil
}

// GetPrivateKeyPEM returns the private key in PEM format
func (am *AuthManager) GetPrivateKeyPEM() (string, error) {
	privKeyBytes := x509.MarshalPKCS1PrivateKey(am.privateKey)

	privKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privKeyBytes,
	})

	return string(privKeyPEM), nil
}

// CleanupExpiredTokens removes expired tokens from cache
func (am *AuthManager) CleanupExpiredTokens() {
	if !am.config.EnableTokenCache {
		return
	}

	now := time.Now()
	for token, tokenInfo := range am.tokenCache {
		if now.After(tokenInfo.ExpiresAt) {
			delete(am.tokenCache, token)
		}
	}
}

// GetTokenCacheSize returns the number of cached tokens
func (am *AuthManager) GetTokenCacheSize() int {
	return len(am.tokenCache)
}

// GetAuthConfig returns the current authentication configuration
func (am *AuthManager) GetAuthConfig() AuthConfig {
	return am.config
}

// UpdateAuthConfig updates the authentication configuration
func (am *AuthManager) UpdateAuthConfig(config AuthConfig) {
	am.config = config
	am.logger.Info("Authentication configuration updated", zap.Any("config", config))
}
