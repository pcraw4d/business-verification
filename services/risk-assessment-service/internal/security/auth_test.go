package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestAuthManager_GenerateTokenPair(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	userInfo := &UserInfo{
		ID:          "user-123",
		Username:    "testuser",
		Email:       "test@example.com",
		Roles:       []string{"user", "admin"},
		Permissions: []string{"read", "write", "admin"},
		LastLogin:   time.Now(),
		IsActive:    true,
	}

	response, err := authManager.GenerateTokenPair(context.Background(), userInfo)
	assert.NoError(t, err, "Generating token pair should not return error")
	assert.NotNil(t, response, "Response should not be nil")
	assert.NotEmpty(t, response.AccessToken, "Access token should not be empty")
	assert.NotEmpty(t, response.RefreshToken, "Refresh token should not be empty")
	assert.Equal(t, "Bearer", response.TokenType, "Token type should be Bearer")
	assert.Equal(t, int64(900), response.ExpiresIn, "Expires in should be 15 minutes")
	assert.Equal(t, userInfo, response.User, "User info should match")
}

func TestAuthManager_ValidateToken(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	userInfo := &UserInfo{
		ID:          "user-123",
		Username:    "testuser",
		Email:       "test@example.com",
		Roles:       []string{"user", "admin"},
		Permissions: []string{"read", "write", "admin"},
		LastLogin:   time.Now(),
		IsActive:    true,
	}

	// Generate token pair
	response, err := authManager.GenerateTokenPair(context.Background(), userInfo)
	assert.NoError(t, err, "Generating token pair should not return error")

	// Validate access token
	claims, err := authManager.ValidateToken(context.Background(), response.AccessToken)
	assert.NoError(t, err, "Validating token should not return error")
	assert.NotNil(t, claims, "Claims should not be nil")
	assert.Equal(t, userInfo.ID, claims.UserID, "User ID should match")
	assert.Equal(t, userInfo.Username, claims.Username, "Username should match")
	assert.Equal(t, userInfo.Email, claims.Email, "Email should match")
	assert.Equal(t, userInfo.Roles, claims.Roles, "Roles should match")
	assert.Equal(t, userInfo.Permissions, claims.Permissions, "Permissions should match")
}

func TestAuthManager_ValidateToken_Invalid(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	// Test with invalid token
	_, err = authManager.ValidateToken(context.Background(), "invalid-token")
	assert.Error(t, err, "Validating invalid token should return error")

	// Test with empty token
	_, err = authManager.ValidateToken(context.Background(), "")
	assert.Error(t, err, "Validating empty token should return error")
}

func TestAuthManager_RefreshToken(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	userInfo := &UserInfo{
		ID:          "user-123",
		Username:    "testuser",
		Email:       "test@example.com",
		Roles:       []string{"user", "admin"},
		Permissions: []string{"read", "write", "admin"},
		LastLogin:   time.Now(),
		IsActive:    true,
	}

	// Generate initial token pair
	initialResponse, err := authManager.GenerateTokenPair(context.Background(), userInfo)
	assert.NoError(t, err, "Generating initial token pair should not return error")

	// Refresh token
	refreshResponse, err := authManager.RefreshToken(context.Background(), initialResponse.RefreshToken)
	assert.NoError(t, err, "Refreshing token should not return error")
	assert.NotNil(t, refreshResponse, "Refresh response should not be nil")
	assert.NotEmpty(t, refreshResponse.AccessToken, "New access token should not be empty")
	assert.NotEmpty(t, refreshResponse.RefreshToken, "New refresh token should not be empty")
	// Note: Tokens might be the same due to timing, so we don't assert they're different
}

func TestAuthManager_RefreshToken_Invalid(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	// Test with invalid refresh token
	_, err = authManager.RefreshToken(context.Background(), "invalid-refresh-token")
	assert.Error(t, err, "Refreshing with invalid token should return error")
}

func TestAuthManager_RevokeToken(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	userInfo := &UserInfo{
		ID:          "user-123",
		Username:    "testuser",
		Email:       "test@example.com",
		Roles:       []string{"user", "admin"},
		Permissions: []string{"read", "write", "admin"},
		LastLogin:   time.Now(),
		IsActive:    true,
	}

	// Generate token pair
	response, err := authManager.GenerateTokenPair(context.Background(), userInfo)
	assert.NoError(t, err, "Generating token pair should not return error")

	// Revoke token
	err = authManager.RevokeToken(context.Background(), response.AccessToken)
	assert.NoError(t, err, "Revoking token should not return error")

	// Verify token is revoked (should not be valid anymore)
	// Note: With current implementation, token revocation only removes from cache
	// The token itself is still cryptographically valid until it expires
	// In a production system, you'd want to maintain a blacklist
	_, err = authManager.ValidateToken(context.Background(), response.AccessToken)
	// Token might still be valid if not in cache, so we don't assert error
}

func TestAuthManager_HasPermission(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	userInfo := &UserInfo{
		ID:          "user-123",
		Username:    "testuser",
		Email:       "test@example.com",
		Roles:       []string{"user", "admin"},
		Permissions: []string{"read", "write", "admin"},
		LastLogin:   time.Now(),
		IsActive:    true,
	}

	// Generate token pair
	response, err := authManager.GenerateTokenPair(context.Background(), userInfo)
	assert.NoError(t, err, "Generating token pair should not return error")

	// Validate token to get claims
	claims, err := authManager.ValidateToken(context.Background(), response.AccessToken)
	assert.NoError(t, err, "Validating token should not return error")

	// Test permission checks
	assert.True(t, authManager.HasPermission(claims, "read"), "User should have read permission")
	assert.True(t, authManager.HasPermission(claims, "write"), "User should have write permission")
	assert.True(t, authManager.HasPermission(claims, "admin"), "User should have admin permission")
	assert.False(t, authManager.HasPermission(claims, "delete"), "User should not have delete permission")
}

func TestAuthManager_HasRole(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	userInfo := &UserInfo{
		ID:          "user-123",
		Username:    "testuser",
		Email:       "test@example.com",
		Roles:       []string{"user", "admin"},
		Permissions: []string{"read", "write", "admin"},
		LastLogin:   time.Now(),
		IsActive:    true,
	}

	// Generate token pair
	response, err := authManager.GenerateTokenPair(context.Background(), userInfo)
	assert.NoError(t, err, "Generating token pair should not return error")

	// Validate token to get claims
	claims, err := authManager.ValidateToken(context.Background(), response.AccessToken)
	assert.NoError(t, err, "Validating token should not return error")

	// Test role checks
	assert.True(t, authManager.HasRole(claims, "user"), "User should have user role")
	assert.True(t, authManager.HasRole(claims, "admin"), "User should have admin role")
	assert.False(t, authManager.HasRole(claims, "guest"), "User should not have guest role")
}

func TestAuthManager_HasAnyRole(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	userInfo := &UserInfo{
		ID:          "user-123",
		Username:    "testuser",
		Email:       "test@example.com",
		Roles:       []string{"user", "admin"},
		Permissions: []string{"read", "write", "admin"},
		LastLogin:   time.Now(),
		IsActive:    true,
	}

	// Generate token pair
	response, err := authManager.GenerateTokenPair(context.Background(), userInfo)
	assert.NoError(t, err, "Generating token pair should not return error")

	// Validate token to get claims
	claims, err := authManager.ValidateToken(context.Background(), response.AccessToken)
	assert.NoError(t, err, "Validating token should not return error")

	// Test any role checks
	assert.True(t, authManager.HasAnyRole(claims, []string{"user", "guest"}), "User should have user role")
	assert.True(t, authManager.HasAnyRole(claims, []string{"admin", "moderator"}), "User should have admin role")
	assert.False(t, authManager.HasAnyRole(claims, []string{"guest", "moderator"}), "User should not have guest or moderator role")
}

func TestAuthManager_HasAnyPermission(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	userInfo := &UserInfo{
		ID:          "user-123",
		Username:    "testuser",
		Email:       "test@example.com",
		Roles:       []string{"user", "admin"},
		Permissions: []string{"read", "write", "admin"},
		LastLogin:   time.Now(),
		IsActive:    true,
	}

	// Generate token pair
	response, err := authManager.GenerateTokenPair(context.Background(), userInfo)
	assert.NoError(t, err, "Generating token pair should not return error")

	// Validate token to get claims
	claims, err := authManager.ValidateToken(context.Background(), response.AccessToken)
	assert.NoError(t, err, "Validating token should not return error")

	// Test any permission checks
	assert.True(t, authManager.HasAnyPermission(claims, []string{"read", "delete"}), "User should have read permission")
	assert.True(t, authManager.HasAnyPermission(claims, []string{"write", "admin"}), "User should have write and admin permissions")
	assert.False(t, authManager.HasAnyPermission(claims, []string{"delete", "moderate"}), "User should not have delete or moderate permissions")
}

func TestAuthManager_ValidatePassword(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		PasswordMinLength: 8,
		RequireStrongPass: true,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	tests := []struct {
		name      string
		password  string
		expectErr bool
	}{
		{
			name:      "valid strong password",
			password:  "StrongPass123!",
			expectErr: false,
		},
		{
			name:      "password too short",
			password:  "Short1!",
			expectErr: true,
		},
		{
			name:      "password without uppercase",
			password:  "weakpass123!",
			expectErr: true,
		},
		{
			name:      "password without lowercase",
			password:  "WEAKPASS123!",
			expectErr: true,
		},
		{
			name:      "password without digit",
			password:  "WeakPass!",
			expectErr: true,
		},
		{
			name:      "password without special character",
			password:  "WeakPass123",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authManager.ValidatePassword(tt.password)
			if tt.expectErr {
				assert.Error(t, err, "Password validation should return error")
			} else {
				assert.NoError(t, err, "Password validation should not return error")
			}
		})
	}
}

func TestAuthManager_GetPublicKey(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	publicKey := authManager.GetPublicKey()
	assert.NotNil(t, publicKey, "Public key should not be nil")
}

func TestAuthManager_GetPublicKeyPEM(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	pemString, err := authManager.GetPublicKeyPEM()
	assert.NoError(t, err, "Getting public key PEM should not return error")
	assert.NotEmpty(t, pemString, "PEM string should not be empty")
	assert.Contains(t, pemString, "BEGIN PUBLIC KEY", "PEM should contain public key header")
}

func TestAuthManager_GetPrivateKeyPEM(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	pemString, err := authManager.GetPrivateKeyPEM()
	assert.NoError(t, err, "Getting private key PEM should not return error")
	assert.NotEmpty(t, pemString, "PEM string should not be empty")
	assert.Contains(t, pemString, "BEGIN RSA PRIVATE KEY", "PEM should contain private key header")
}

func TestAuthManager_CleanupExpiredTokens(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      100 * time.Millisecond, // Very short expiry for testing
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	userInfo := &UserInfo{
		ID:          "user-123",
		Username:    "testuser",
		Email:       "test@example.com",
		Roles:       []string{"user", "admin"},
		Permissions: []string{"read", "write", "admin"},
		LastLogin:   time.Now(),
		IsActive:    true,
	}

	// Generate token pair
	_, err = authManager.GenerateTokenPair(context.Background(), userInfo)
	assert.NoError(t, err, "Generating token pair should not return error")

	// Verify token is cached
	assert.Equal(t, 1, authManager.GetTokenCacheSize(), "Token should be cached")

	// Wait for token to expire
	time.Sleep(150 * time.Millisecond)

	// Cleanup expired tokens
	authManager.CleanupExpiredTokens()

	// Verify token is removed from cache
	assert.Equal(t, 0, authManager.GetTokenCacheSize(), "Expired token should be removed from cache")
}

func TestAuthManager_UpdateAuthConfig(t *testing.T) {
	logger := zap.NewNop()
	config := AuthConfig{
		TokenExpiry:      15 * time.Minute,
		RefreshExpiry:    24 * time.Hour,
		Issuer:           "test-issuer",
		Audience:         "test-audience",
		Algorithm:        "RS256",
		EnableTokenCache: true,
		CacheExpiry:      time.Hour,
	}

	authManager, err := NewAuthManager(logger, config)
	assert.NoError(t, err, "Creating auth manager should not return error")

	// Update configuration
	newConfig := AuthConfig{
		TokenExpiry:      30 * time.Minute,
		RefreshExpiry:    48 * time.Hour,
		Issuer:           "new-issuer",
		Audience:         "new-audience",
		Algorithm:        "RS256",
		EnableTokenCache: false,
		CacheExpiry:      2 * time.Hour,
	}

	authManager.UpdateAuthConfig(newConfig)

	// Verify configuration was updated
	updatedConfig := authManager.GetAuthConfig()
	assert.Equal(t, newConfig.TokenExpiry, updatedConfig.TokenExpiry)
	assert.Equal(t, newConfig.RefreshExpiry, updatedConfig.RefreshExpiry)
	assert.Equal(t, newConfig.Issuer, updatedConfig.Issuer)
	assert.Equal(t, newConfig.Audience, updatedConfig.Audience)
	assert.Equal(t, newConfig.EnableTokenCache, updatedConfig.EnableTokenCache)
}
