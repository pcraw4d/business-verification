package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
	"golang.org/x/crypto/bcrypt"
)

// AuthService provides authentication and authorization functionality
type AuthService struct {
	config        *config.AuthConfig
	db            database.Database
	logger        *observability.Logger
	metrics       *observability.Metrics
	blacklistRepo *TokenBlacklistRepository
}

// NewAuthService creates a new authentication service
func NewAuthService(cfg *config.AuthConfig, db database.Database, logger *observability.Logger, metrics *observability.Metrics) *AuthService {
	return &AuthService{
		config:        cfg,
		db:            db,
		logger:        logger,
		metrics:       metrics,
		blacklistRepo: NewTokenBlacklistRepository(db, logger),
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

// RegisterUser registers a new user
func (a *AuthService) RegisterUser(ctx context.Context, req *RegisterRequest) (*UserResponse, error) {
	// Check if user already exists
	existingUser, err := a.db.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Check if username is taken
	existingUser, err = a.db.GetUserByEmail(ctx, req.Username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username %s is already taken", req.Username)
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &database.User{
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

	// Save user to database
	if err := a.db.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log registration
	a.logger.WithUser(user.ID).LogBusinessEvent(ctx, "user_registered", user.ID, map[string]interface{}{
		"email":      user.Email,
		"username":   user.Username,
		"company":    user.Company,
		"ip_address": getIPFromContext(ctx),
	})

	// Record metrics
	a.metrics.RecordBusinessClassification("success", "registration")

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
	// Get user by email
	user, err := a.db.GetUserByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Check if account is locked
	if user.LockedUntil != nil && time.Now().Before(*user.LockedUntil) {
		return nil, fmt.Errorf("account is locked until %s", user.LockedUntil.Format(time.RFC3339))
	}

	// Check if account is active
	if user.Status != "active" {
		return nil, fmt.Errorf("account is not active")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		// Increment failed login attempts
		user.FailedLoginAttempts++

		// Lock account if too many failed attempts
		if user.FailedLoginAttempts >= a.config.MaxLoginAttempts {
			lockUntil := time.Now().Add(a.config.LockoutDuration)
			user.LockedUntil = &lockUntil
		}

		a.db.UpdateUser(ctx, user)

		// Log failed login attempt
		a.logger.WithUser(user.ID).LogSecurityEvent(ctx, "login_failed", user.ID, getIPFromContext(ctx), map[string]interface{}{
			"email":           user.Email,
			"failed_attempts": user.FailedLoginAttempts,
		})

		return nil, fmt.Errorf("invalid credentials")
	}

	// Reset failed login attempts on successful login
	user.FailedLoginAttempts = 0
	user.LockedUntil = nil
	now := time.Now()
	user.LastLoginAt = &now
	user.UpdatedAt = now

	if err := a.db.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

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
	a.logger.WithUser(user.ID).LogBusinessEvent(ctx, "user_logged_in", user.ID, map[string]interface{}{
		"email":      user.Email,
		"ip_address": getIPFromContext(ctx),
	})

	// Record metrics
	a.metrics.RecordBusinessClassification("success", "login")

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

	// Get user from database
	user, err := a.db.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
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
	a.logger.WithUser(user.ID).LogBusinessEvent(ctx, "token_refreshed", user.ID, map[string]interface{}{
		"ip_address": getIPFromContext(ctx),
	})

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

	// Get user from database
	user, err := a.db.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
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
	a.logger.WithUser(userID).LogBusinessEvent(ctx, "user_logged_out", userID, map[string]interface{}{
		"ip_address": getIPFromContext(ctx),
		"token_id":   claims.ID,
	})

	// Record metrics
	a.metrics.RecordBusinessClassification("success", "logout")

	return nil
}

// ChangePassword changes a user's password
func (a *AuthService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	// Get user from database
	user, err := a.db.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
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

	if err := a.db.UpdateUser(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// Log password change
	a.logger.WithUser(user.ID).LogSecurityEvent(ctx, "password_changed", user.ID, getIPFromContext(ctx), map[string]interface{}{
		"ip_address": getIPFromContext(ctx),
	})

	// Record metrics
	a.metrics.RecordBusinessClassification("success", "password_change")

	return nil
}

// generateAccessToken generates a new access token
func (a *AuthService) generateAccessToken(user *database.User) (string, error) {
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
func (a *AuthService) generateRefreshToken(user *database.User) (string, error) {
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

// getIPFromContext extracts IP address from context
func getIPFromContext(ctx context.Context) string {
	// This is a placeholder - in a real implementation, you would extract
	// the IP address from the HTTP request context
	return "unknown"
}

// TokenBlacklistRepository handles token blacklisting operations
type TokenBlacklistRepository struct {
	db     database.Database
	logger *observability.Logger
}

// NewTokenBlacklistRepository creates a new token blacklist repository
func NewTokenBlacklistRepository(db database.Database, logger *observability.Logger) *TokenBlacklistRepository {
	return &TokenBlacklistRepository{
		db:     db,
		logger: logger,
	}
}

// BlacklistToken adds a token to the blacklist
func (r *TokenBlacklistRepository) BlacklistToken(ctx context.Context, tokenID string, expiresAt time.Time) error {
	// In a production system, you would store this in a database table or Redis
	// For now, we'll log the blacklisting event
	r.logger.WithComponent("auth").LogSecurityEvent(ctx, "token_blacklisted", "", getIPFromContext(ctx), map[string]interface{}{
		"token_id":   tokenID,
		"expires_at": expiresAt,
	})
	
	// TODO: Implement actual storage in database or Redis
	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func (r *TokenBlacklistRepository) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	// In a production system, you would check the database or Redis
	// For now, we'll return false (no token is blacklisted)
	return false, nil
}
