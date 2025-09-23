package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"kyb-platform/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// AuthService provides authentication and authorization functionality
type AuthService struct {
	config        *config.AuthConfig
	logger        *zap.Logger
	blacklistRepo *TokenBlacklistRepository
}

// NewAuthService creates a new authentication service
func NewAuthService(cfg *config.AuthConfig, logger *zap.Logger) *AuthService {
	return &AuthService{
		config:        cfg,
		logger:        logger,
		blacklistRepo: NewTokenBlacklistRepository(logger),
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Company   string `json:"company" validate:"required"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// TokenResponse represents a JWT token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// UserResponse represents a user response
type UserResponse struct {
	ID            string     `json:"id"`
	Email         string     `json:"email"`
	Username      string     `json:"username"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Company       string     `json:"company"`
	Role          string     `json:"role"`
	Status        string     `json:"status"`
	EmailVerified bool       `json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Claims represents JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// User represents a user in the system
type User struct {
	ID                  string     `json:"id"`
	Email               string     `json:"email"`
	Username            string     `json:"username"`
	PasswordHash        string     `json:"-"`
	FirstName           string     `json:"first_name"`
	LastName            string     `json:"last_name"`
	Company             string     `json:"company"`
	Role                string     `json:"role"`
	Status              string     `json:"status"`
	EmailVerified       bool       `json:"email_verified"`
	FailedLoginAttempts int        `json:"-"`
	LockedUntil         *time.Time `json:"-"`
	LastLoginAt         *time.Time `json:"last_login_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, limit, offset int) ([]*User, error)
}

// TokenBlacklistRepository handles token blacklisting operations
type TokenBlacklistRepository struct {
	logger *zap.Logger
	// In a real implementation, this would use a database
	blacklistedTokens map[string]time.Time
}

// NewTokenBlacklistRepository creates a new token blacklist repository
func NewTokenBlacklistRepository(logger *zap.Logger) *TokenBlacklistRepository {
	return &TokenBlacklistRepository{
		logger:            logger,
		blacklistedTokens: make(map[string]time.Time),
	}
}

// BlacklistToken adds a token to the blacklist
func (r *TokenBlacklistRepository) BlacklistToken(ctx context.Context, tokenID string, expiresAt time.Time) error {
	r.blacklistedTokens[tokenID] = expiresAt
	r.logger.Info("Token blacklisted", zap.String("token_id", tokenID))
	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *TokenBlacklistRepository) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	expiresAt, exists := r.blacklistedTokens[tokenID]
	if !exists {
		return false, nil
	}

	// Clean up expired tokens
	if time.Now().After(expiresAt) {
		delete(r.blacklistedTokens, tokenID)
		return false, nil
	}

	return true, nil
}

// RegisterUser registers a new user
func (a *AuthService) RegisterUser(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
	// This is a simplified implementation - in a real system, you would use the UserRepository
	// For now, we'll create a mock user response

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		ID:            generateUUID(),
		Email:         strings.ToLower(req.Email),
		Username:      req.Username,
		PasswordHash:  string(hashedPassword),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Company:       req.Company,
		Role:          "user",
		Status:        "active",
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Log registration
	a.logger.Info("User registered",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("username", user.Username),
		zap.String("company", user.Company))

	return &UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Company:       user.Company,
		Role:          user.Role,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
		LastLoginAt:   user.LastLoginAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

// LoginUser authenticates a user and returns JWT tokens
func (a *AuthService) LoginUser(ctx context.Context, req *LoginRequest) (*TokenResponse, error) {
	// This is a simplified implementation - in a real system, you would use the UserRepository
	// For now, we'll create a mock user for demonstration

	// Mock user - in real implementation, get from database
	user := &User{
		ID:            generateUUID(),
		Email:         strings.ToLower(req.Email),
		Username:      "testuser",
		PasswordHash:  "$2a$10$test_hash", // This would be the actual hash
		FirstName:     "Test",
		LastName:      "User",
		Company:       "Test Corp",
		Role:          "user",
		Status:        "active",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, fmt.Errorf("account is locked until %s", user.LockedUntil.Format(time.RFC3339))
	}

	// Check if account is active
	if user.Status != "active" {
		return nil, fmt.Errorf("account is not active")
	}

	// In a real implementation, you would verify the password hash
	// For now, we'll just check if the password is not empty
	if req.Password == "" {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now

	// Generate tokens
	accessToken, err := a.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := a.generateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Log successful login
	a.logger.Info("User logged in",
		zap.String("user_id", user.ID),
		zap.String("email", user.Email))

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(a.config.JWTExpiration.Seconds()),
	}, nil
}

// RefreshToken refreshes an access token using a refresh token
func (a *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*TokenResponse, error) {
	// Parse and validate refresh token
	claims, err := a.parseToken(refreshToken, a.config.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// Get user from database (mock for now)
	user := &User{
		ID:            claims.UserID,
		Email:         claims.Email,
		Username:      claims.Username,
		Role:          claims.Role,
		Status:        "active",
		EmailVerified: true,
	}

	// Check if user is still active
	if user.Status != "active" {
		return nil, fmt.Errorf("user account is not active")
	}

	// Generate new tokens
	accessToken, err := a.generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := a.generateRefreshToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Log token refresh
	a.logger.Info("Token refreshed", zap.String("user_id", user.ID))

	return &TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(a.config.JWTExpiration.Seconds()),
	}, nil
}

// ValidateToken validates an access token and returns user information
func (a *AuthService) ValidateToken(ctx context.Context, tokenString string) (*UserResponse, error) {
	// Parse and validate token
	claims, err := a.parseToken(tokenString, a.config.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Check if token is blacklisted
	isBlacklisted, err := a.blacklistRepo.IsTokenBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, fmt.Errorf("token has been revoked")
	}

	// Get user from database (mock for now)
	user := &User{
		ID:            claims.UserID,
		Email:         claims.Email,
		Username:      claims.Username,
		FirstName:     "Test",
		LastName:      "User",
		Company:       "Test Corp",
		Role:          claims.Role,
		Status:        "active",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Check if user is still active
	if user.Status != "active" {
		return nil, fmt.Errorf("user account is not active")
	}

	return &UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Company:       user.Company,
		Role:          user.Role,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
		LastLoginAt:   user.LastLoginAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

// LogoutUser invalidates the current session by blacklisting the token
func (a *AuthService) LogoutUser(ctx context.Context, tokenString string, userID string) error {
	// Parse token to get claims
	claims, err := a.parseToken(tokenString, a.config.JWTSecret)
	if err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	// Add token to blacklist
	if err := a.blacklistRepo.BlacklistToken(ctx, claims.ID, claims.ExpiresAt.Time); err != nil {
		return fmt.Errorf("failed to blacklist token: %w", err)
	}

	// Log logout event
	a.logger.Info("User logged out",
		zap.String("user_id", userID),
		zap.String("token_id", claims.ID))

	return nil
}

// ChangePassword changes a user's password
func (a *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get user from database (mock for now)
	user := &User{
		ID:           userID,
		PasswordHash: "$2a$10$test_hash", // This would be the actual hash
	}

	// In a real implementation, you would verify the current password
	// For now, we'll just check if the current password is not empty
	if currentPassword == "" {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update user password
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	// Log password change
	a.logger.Info("Password changed", zap.String("user_id", userID))

	return nil
}

// GetUserByID gets a user by ID
func (a *AuthService) GetUserByID(ctx context.Context, userID string) (*UserResponse, error) {
	// Get user from database (mock for now)
	user := &User{
		ID:            userID,
		Email:         "test@example.com",
		Username:      "testuser",
		FirstName:     "Test",
		LastName:      "User",
		Company:       "Test Corp",
		Role:          "user",
		Status:        "active",
		EmailVerified: true,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return &UserResponse{
		ID:            user.ID,
		Email:         user.Email,
		Username:      user.Username,
		FirstName:     user.FirstName,
		LastName:      user.LastName,
		Company:       user.Company,
		Role:          user.Role,
		Status:        user.Status,
		EmailVerified: user.EmailVerified,
		LastLoginAt:   user.LastLoginAt,
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}, nil
}

// CreateEmailVerificationToken creates a new email verification token
func (a *AuthService) CreateEmailVerificationToken(ctx context.Context, userID string) (*EmailVerificationToken, error) {
	token := generateRandomToken()
	expiresAt := time.Now().Add(24 * time.Hour) // Token expires in 24 hours

	verificationToken := &EmailVerificationToken{
		ID:        generateUUID(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	// Log token creation
	a.logger.Info("Email verification token created",
		zap.String("user_id", userID),
		zap.Time("expires_at", expiresAt))

	return verificationToken, nil
}

// VerifyEmail verifies a user's email using the provided token
func (a *AuthService) VerifyEmail(ctx context.Context, token string) error {
	// This is a simplified implementation
	// In a real system, you would validate the token against the database

	// Log email verification
	a.logger.Info("Email verified", zap.String("token", token))

	return nil
}

// CreatePasswordResetToken creates a new password reset token
func (a *AuthService) CreatePasswordResetToken(ctx context.Context, email string) (*PasswordResetToken, error) {
	// This is a simplified implementation
	// In a real system, you would create and store the token

	token := generateRandomToken()
	expiresAt := time.Now().Add(1 * time.Hour) // Token expires in 1 hour

	resetToken := &PasswordResetToken{
		ID:        generateUUID(),
		UserID:    generateUUID(), // Mock user ID
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}

	// Log token creation
	a.logger.Info("Password reset token created",
		zap.String("email", email),
		zap.Time("expires_at", expiresAt))

	return resetToken, nil
}

// ResetPassword resets a user's password using the provided token
func (a *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	// This is a simplified implementation
	// In a real system, you would validate the token and update the password

	// Log password reset
	a.logger.Info("Password reset", zap.String("token", token))

	return nil
}

// generateAccessToken generates a new access token
func (a *AuthService) generateAccessToken(user *User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(a.config.JWTExpiration)
	tokenID := generateUUID()

	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "kyb-tool",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.JWTSecret))
}

// generateRefreshToken generates a new refresh token
func (a *AuthService) generateRefreshToken(user *User) (string, error) {
	now := time.Now()
	expiresAt := now.Add(a.config.RefreshExpiration)
	tokenID := generateUUID()

	claims := &Claims{
		UserID:   user.ID,
		Email:    user.Email,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "kyb-tool",
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.config.JWTSecret))
}

// parseToken parses and validates a JWT token
func (a *AuthService) parseToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// generateUUID generates a new UUID
func generateUUID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// generateRandomToken generates a secure random token
func generateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// EmailVerificationToken represents an email verification token
type EmailVerificationToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}
