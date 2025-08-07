package auth

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

func TestNewAuthService(t *testing.T) {
	cfg := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
		MaxLoginAttempts:  5,
		LockoutDuration:   30 * time.Minute,
	}

	// Create mock dependencies
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})

	// Note: We can't easily mock the database interface for this test
	// In a real implementation, you would use a mock database

	authService := NewAuthService(cfg, nil, logger, metrics)
	if authService == nil {
		t.Fatal("Expected auth service to be created")
	}

	if authService.config != cfg {
		t.Error("Expected config to match input config")
	}
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

	if req.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", req.Email)
	}

	if req.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", req.Username)
	}

	if req.Password != "password123" {
		t.Errorf("Expected password 'password123', got %s", req.Password)
	}

	if req.FirstName != "John" {
		t.Errorf("Expected first name 'John', got %s", req.FirstName)
	}

	if req.LastName != "Doe" {
		t.Errorf("Expected last name 'Doe', got %s", req.LastName)
	}

	if req.Company != "Test Corp" {
		t.Errorf("Expected company 'Test Corp', got %s", req.Company)
	}
}

func TestLoginRequest(t *testing.T) {
	req := &LoginRequest{
		Email:    "test@example.com",
		Password: "password123",
	}

	if req.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", req.Email)
	}

	if req.Password != "password123" {
		t.Errorf("Expected password 'password123', got %s", req.Password)
	}
}

func TestTokenResponse(t *testing.T) {
	resp := &TokenResponse{
		AccessToken:  "access_token_123",
		RefreshToken: "refresh_token_456",
		TokenType:    "Bearer",
		ExpiresIn:    900, // 15 minutes
	}

	if resp.AccessToken != "access_token_123" {
		t.Errorf("Expected access token 'access_token_123', got %s", resp.AccessToken)
	}

	if resp.RefreshToken != "refresh_token_456" {
		t.Errorf("Expected refresh token 'refresh_token_456', got %s", resp.RefreshToken)
	}

	if resp.TokenType != "Bearer" {
		t.Errorf("Expected token type 'Bearer', got %s", resp.TokenType)
	}

	if resp.ExpiresIn != 900 {
		t.Errorf("Expected expires in 900, got %d", resp.ExpiresIn)
	}
}

func TestUserResponse(t *testing.T) {
	now := time.Now()
	lastLogin := now.Add(-1 * time.Hour)

	resp := &UserResponse{
		ID:            "user-123",
		Email:         "test@example.com",
		Username:      "testuser",
		FirstName:     "John",
		LastName:      "Doe",
		Company:       "Test Corp",
		Role:          "user",
		Status:        "active",
		EmailVerified: true,
		LastLoginAt:   &lastLogin,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if resp.ID != "user-123" {
		t.Errorf("Expected ID 'user-123', got %s", resp.ID)
	}

	if resp.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", resp.Email)
	}

	if resp.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", resp.Username)
	}

	if resp.FirstName != "John" {
		t.Errorf("Expected first name 'John', got %s", resp.FirstName)
	}

	if resp.LastName != "Doe" {
		t.Errorf("Expected last name 'Doe', got %s", resp.LastName)
	}

	if resp.Company != "Test Corp" {
		t.Errorf("Expected company 'Test Corp', got %s", resp.Company)
	}

	if resp.Role != "user" {
		t.Errorf("Expected role 'user', got %s", resp.Role)
	}

	if resp.Status != "active" {
		t.Errorf("Expected status 'active', got %s", resp.Status)
	}

	if !resp.EmailVerified {
		t.Error("Expected email verified to be true")
	}

	if resp.LastLoginAt == nil {
		t.Error("Expected last login at to be set")
	}

	if resp.CreatedAt != now {
		t.Errorf("Expected created at to be %v, got %v", now, resp.CreatedAt)
	}

	if resp.UpdatedAt != now {
		t.Errorf("Expected updated at to be %v, got %v", now, resp.UpdatedAt)
	}
}

func TestClaims(t *testing.T) {
	now := time.Now()
	claims := &Claims{
		UserID:   "user-123",
		Email:    "test@example.com",
		Username: "testuser",
		Role:     "user",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "kyb-tool",
			Subject:   "user-123",
		},
	}

	if claims.UserID != "user-123" {
		t.Errorf("Expected user ID 'user-123', got %s", claims.UserID)
	}

	if claims.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", claims.Email)
	}

	if claims.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", claims.Username)
	}

	if claims.Role != "user" {
		t.Errorf("Expected role 'user', got %s", claims.Role)
	}

	if claims.Subject != "user-123" {
		t.Errorf("Expected subject 'user-123', got %s", claims.Subject)
	}

	if claims.Issuer != "kyb-tool" {
		t.Errorf("Expected issuer 'kyb-tool', got %s", claims.Issuer)
	}
}

func TestGenerateUUID(t *testing.T) {
	uuid1 := generateUUID()
	uuid2 := generateUUID()

	if uuid1 == "" {
		t.Error("Expected UUID to be generated")
	}

	if uuid2 == "" {
		t.Error("Expected UUID to be generated")
	}

	if uuid1 == uuid2 {
		t.Error("Expected UUIDs to be unique")
	}

	// Check that UUIDs are hex strings
	if len(uuid1) != 32 {
		t.Errorf("Expected UUID length 32, got %d", len(uuid1))
	}

	if len(uuid2) != 32 {
		t.Errorf("Expected UUID length 32, got %d", len(uuid2))
	}
}

func TestGetIPFromContext(t *testing.T) {
	ctx := context.Background()
	ip := getIPFromContext(ctx)

	if ip != "unknown" {
		t.Errorf("Expected IP 'unknown', got %s", ip)
	}
}
