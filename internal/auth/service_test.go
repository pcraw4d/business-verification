package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewAuthService(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
		MaxLoginAttempts:  5,
		LockoutDuration:   30 * time.Minute,
	}

	logger := zap.NewNop()

	authService := NewAuthService(cfg, logger)
	require.NotNil(t, authService)
	assert.Equal(t, cfg, authService.config)
	assert.NotNil(t, authService.blacklistRepo)
}

func TestRegisterRequest(t *testing.T) {
	req := &RegisterRequest{
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
		Company:   "Test Corp",
	}

	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "testuser", req.Username)
	assert.Equal(t, "password123", req.Password)
	assert.Equal(t, "John", req.FirstName)
	assert.Equal(t, "Doe", req.LastName)
	assert.Equal(t, "Test Corp", req.Company)
}

func TestLoginRequest(t *testing.T) {
	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "password123", req.Password)
}

func TestTokenResponse(t *testing.T) {
	resp := &TokenResponse{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_456",
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes
	}

	assert.Equal(t, "access_token_123", resp.AccessToken)
	assert.Equal(t, "refresh_token_456", resp.RefreshToken)
	assert.Equal(t, "Bearer", resp.TokenType)
	assert.Equal(t, int64(900), resp.ExpiresIn)
}

func TestUserResponse(t *testing.T) {
	now := time.Now()
	user := &UserResponse{
		ID:            "user_123",
		Email:         "test@example.com",
		Username:      "testuser",
		FirstName:     "John",
		LastName:      "Doe",
		Company:       "Test Corp",
		Role:          "user",
		Status:        "active",
		EmailVerified: true,
		LastLoginAt:   &now,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	assert.Equal(t, "user_123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, "Test Corp", user.Company)
	assert.Equal(t, "user", user.Role)
	assert.Equal(t, "active", user.Status)
	assert.True(t, user.EmailVerified)
	assert.Equal(t, &now, user.LastLoginAt)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}

func TestAuthService_RegisterUser(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	req := &RegisterRequest{
		Email:     "test@example.com",
		Username:  "testuser",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
		Company:   "Test Corp",
	}

	user, err := authService.RegisterUser(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.NotEmpty(t, user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, "Test Corp", user.Company)
	assert.Equal(t, "user", user.Role)
	assert.Equal(t, "active", user.Status)
	assert.False(t, user.EmailVerified)
}

func TestAuthService_LoginUser(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	tokenResponse, err := authService.LoginUser(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, tokenResponse)

	assert.NotEmpty(t, tokenResponse.AccessToken)
	assert.NotEmpty(t, tokenResponse.RefreshToken)
	assert.Equal(t, "Bearer", tokenResponse.TokenType)
	assert.Equal(t, int64(900), tokenResponse.ExpiresIn) // 15 minutes
}

func TestAuthService_LoginUser_EmptyPassword(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "",
	}

	_, err := authService.LoginUser(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid credentials")
}

func TestAuthService_ValidateToken(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	// Create a valid token
	user := &User{
		ID:            "user_123",
		Email:         "test@example.com",
		Username:      "testuser",
		Role:          "user",
		Status:        "active",
		EmailVerified: true,
	}

	token, err := authService.generateAccessToken(user)
	require.NoError(t, err)

	// Validate the token
	userResponse, err := authService.ValidateToken(context.Background(), token)
	require.NoError(t, err)
	require.NotNil(t, userResponse)

	assert.Equal(t, "user_123", userResponse.ID)
	assert.Equal(t, "test@example.com", userResponse.Email)
	assert.Equal(t, "testuser", userResponse.Username)
	assert.Equal(t, "user", userResponse.Role)
	assert.Equal(t, "active", userResponse.Status)
	assert.True(t, userResponse.EmailVerified)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	_, err := authService.ValidateToken(context.Background(), "invalid-token")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid token")
}

func TestAuthService_RefreshToken(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	// Create a valid refresh token
	user := &User{
		ID:            "user_123",
		Email:         "test@example.com",
		Username:      "testuser",
		Role:          "user",
		Status:        "active",
		EmailVerified: true,
	}

	refreshToken, err := authService.generateRefreshToken(user)
	require.NoError(t, err)

	// Refresh the token
	tokenResponse, err := authService.RefreshToken(context.Background(), refreshToken)
	require.NoError(t, err)
	require.NotNil(t, tokenResponse)

	assert.NotEmpty(t, tokenResponse.AccessToken)
	assert.NotEmpty(t, tokenResponse.RefreshToken)
	assert.Equal(t, "Bearer", tokenResponse.TokenType)
	assert.Equal(t, int64(900), tokenResponse.ExpiresIn) // 15 minutes
}

func TestAuthService_LogoutUser(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	// Create a valid token
	user := &User{
		ID:            "user_123",
		Email:         "test@example.com",
		Username:      "testuser",
		Role:          "user",
		Status:        "active",
		EmailVerified: true,
	}

	token, err := authService.generateAccessToken(user)
	require.NoError(t, err)

	// Logout user
	err = authService.LogoutUser(context.Background(), token, user.ID)
	require.NoError(t, err)

	// Try to validate the blacklisted token
	_, err = authService.ValidateToken(context.Background(), token)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "token has been revoked")
}

func TestAuthService_ChangePassword(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	err := authService.ChangePassword(context.Background(), "user_123", "current_password", "new_password")
	require.NoError(t, err)
}

func TestAuthService_ChangePassword_EmptyCurrentPassword(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	err := authService.ChangePassword(context.Background(), "user_123", "", "new_password")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "current password is incorrect")
}

func TestAuthService_GetUserByID(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	user, err := authService.GetUserByID(context.Background(), "user_123")
	require.NoError(t, err)
	require.NotNil(t, user)

	assert.Equal(t, "user_123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "Test", user.FirstName)
	assert.Equal(t, "User", user.LastName)
	assert.Equal(t, "Test Corp", user.Company)
	assert.Equal(t, "user", user.Role)
	assert.Equal(t, "active", user.Status)
	assert.True(t, user.EmailVerified)
}

func TestAuthService_CreateEmailVerificationToken(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	token, err := authService.CreateEmailVerificationToken(context.Background(), "user_123")
	require.NoError(t, err)
	require.NotNil(t, token)

	assert.NotEmpty(t, token.ID)
	assert.Equal(t, "user_123", token.UserID)
	assert.NotEmpty(t, token.Token)
	assert.True(t, token.ExpiresAt.After(time.Now()))
	assert.True(t, token.CreatedAt.Before(time.Now().Add(time.Second)))
}

func TestAuthService_VerifyEmail(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	err := authService.VerifyEmail(context.Background(), "test-token")
	require.NoError(t, err)
}

func TestAuthService_CreatePasswordResetToken(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	token, err := authService.CreatePasswordResetToken(context.Background(), "test@example.com")
	require.NoError(t, err)
	require.NotNil(t, token)

	assert.NotEmpty(t, token.ID)
	assert.NotEmpty(t, token.UserID)
	assert.NotEmpty(t, token.Token)
	assert.True(t, token.ExpiresAt.After(time.Now()))
	assert.True(t, token.CreatedAt.Before(time.Now().Add(time.Second)))
}

func TestAuthService_ResetPassword(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}
	logger := zap.NewNop()
	authService := NewAuthService(cfg, logger)

	err := authService.ResetPassword(context.Background(), "test-token", "new_password")
	require.NoError(t, err)
}

func TestTokenBlacklistRepository(t *testing.T) {
	logger := zap.NewNop()
	repo := NewTokenBlacklistRepository(logger)

	// Test blacklisting a token
	tokenID := "token_123"
	expiresAt := time.Now().Add(time.Hour)
	err := repo.BlacklistToken(context.Background(), tokenID, expiresAt)
	require.NoError(t, err)

	// Test checking if token is blacklisted
	isBlacklisted, err := repo.IsTokenBlacklisted(context.Background(), tokenID)
	require.NoError(t, err)
	assert.True(t, isBlacklisted)

	// Test checking non-blacklisted token
	isBlacklisted, err = repo.IsTokenBlacklisted(context.Background(), "non-existent")
	require.NoError(t, err)
	assert.False(t, isBlacklisted)
}

func TestGenerateUUID(t *testing.T) {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	assert.NotEmpty(t, uuid1)
	assert.NotEmpty(t, uuid2)
	assert.NotEqual(t, uuid1, uuid2)
	assert.Len(t, uuid1, 32) // 16 bytes = 32 hex characters
}

func TestGenerateRandomToken(t *testing.T) {
	token1 := generateRandomToken()
	token2 := generateRandomToken()

	assert.NotEmpty(t, token1)
	assert.NotEmpty(t, token2)
	assert.NotEqual(t, token1, token2)
	assert.Len(t, token1, 64) // 32 bytes = 64 hex characters
}

func TestClaims(t *testing.T) {
	claims := &Claims{
		UserID:   "user_123",
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "token_123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kyb-tool",
			Subject:   "user_123",
		},
	}

	assert.Equal(t, "user_123", claims.UserID)
	assert.Equal(t, "test@example.com", claims.Email)
	assert.Equal(t, "testuser", claims.Username)
	assert.Equal(t, "user", claims.Role)
	assert.Equal(t, "token_123", claims.ID)
	assert.Equal(t, "kyb-tool", claims.Issuer)
	assert.Equal(t, "user_123", claims.Subject)
}

func TestUser(t *testing.T) {
	now := time.Now()
	user := &User{
		ID:                  "user_123",
		Email:               "test@example.com",
		Username:            "testuser",
		PasswordHash:        "hashed_password",
		FirstName:           "John",
		LastName:            "Doe",
		Company:             "Test Corp",
		Role:                "user",
		Status:              "active",
		EmailVerified:       true,
		FailedLoginAttempts: 0,
		LockedUntil:         nil,
		LastLoginAt:         &now,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	assert.Equal(t, "user_123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "hashed_password", user.PasswordHash)
	assert.Equal(t, "John", user.FirstName)
	assert.Equal(t, "Doe", user.LastName)
	assert.Equal(t, "Test Corp", user.Company)
	assert.Equal(t, "user", user.Role)
	assert.Equal(t, "active", user.Status)
	assert.True(t, user.EmailVerified)
	assert.Equal(t, 0, user.FailedLoginAttempts)
	assert.Nil(t, user.LockedUntil)
	assert.Equal(t, &now, user.LastLoginAt)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
}
