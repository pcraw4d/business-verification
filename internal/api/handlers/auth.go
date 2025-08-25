package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/auth"
	"github.com/pcraw4d/business-verification/internal/config"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService *auth.AuthService
	logger      *zap.Logger
	config      *config.Config
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *auth.AuthService, logger *zap.Logger, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		logger:      logger,
		config:      cfg,
	}
}

func (h *AuthHandler) sameSite() http.SameSite {
	switch strings.ToLower(h.config.Auth.CookieSameSite) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func (h *AuthHandler) setRefreshCookie(w http.ResponseWriter, token string, expiresInSeconds int64) {
	if token == "" {
		return
	}
	expireAt := time.Now().Add(time.Duration(expiresInSeconds) * time.Second)
	cookie := &http.Cookie{
		Name:     h.config.Auth.RefreshCookieName,
		Value:    token,
		Path:     h.config.Auth.CookiePath,
		Domain:   h.config.Auth.CookieDomain,
		Expires:  expireAt,
		HttpOnly: true,
		Secure:   h.config.Auth.CookieSecure,
		SameSite: h.sameSite(),
	}
	http.SetCookie(w, cookie)
}

func (h *AuthHandler) clearCookie(w http.ResponseWriter, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     h.config.Auth.CookiePath,
		Domain:   h.config.Auth.CookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: name == h.config.Auth.RefreshCookieName,
		Secure:   h.config.Auth.CookieSecure,
		SameSite: h.sameSite(),
	}
	http.SetCookie(w, cookie)
}

func (h *AuthHandler) generateCSRFToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (h *AuthHandler) setCSRFCookie(w http.ResponseWriter, token string) {
	cookie := &http.Cookie{
		Name:     h.config.Auth.CSRFCookieName,
		Value:    token,
		Path:     h.config.Auth.CookiePath,
		Domain:   h.config.Auth.CookieDomain,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false, // must be readable by JS to set header
		Secure:   h.config.Auth.CookieSecure,
		SameSite: h.sameSite(),
	}
	http.SetCookie(w, cookie)
}

// RegisterHandler handles user registration requests
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req auth.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode register request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Register user
	user, err := h.authService.RegisterUser(ctx, &req)
	if err != nil {
		h.logger.Error("User registration failed", zap.Error(err), zap.String("email", req.Email))

		// Handle different error types
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			http.Error(w, err.Error(), http.StatusConflict)
		} else {
			http.Error(w, "Registration failed", http.StatusInternalServerError)
		}
		return
	}

	// Create email verification token
	verificationToken, err := h.authService.CreateEmailVerificationToken(ctx, user.ID)
	if err != nil {
		h.logger.Error("Failed to create email verification token", zap.Error(err), zap.String("user_id", user.ID))
		// Don't fail registration, just log the error
	}

	// Response
	response := map[string]interface{}{
		"user":    user,
		"message": "Registration successful. Please check your email for verification instructions.",
	}

	if verificationToken != nil {
		response["verification_token"] = verificationToken.Token // Only for testing/development
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// LoginHandler handles user login requests
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode login request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Authenticate user
	tokenResponse, err := h.authService.LoginUser(ctx, &req)
	if err != nil {
		h.logger.Error("User login failed", zap.Error(err), zap.String("email", req.Email))

		// Handle different error types
		if err.Error() == "invalid credentials" {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		} else if err.Error() == "account locked" {
			http.Error(w, "Account temporarily locked due to too many failed attempts", http.StatusLocked)
		} else {
			http.Error(w, "Login failed", http.StatusInternalServerError)
		}
		return
	}

	// Set secure cookies for session (refresh token) and CSRF token
	h.setRefreshCookie(w, tokenResponse.RefreshToken, tokenResponse.ExpiresIn)
	csrf := h.generateCSRFToken()
	h.setCSRFCookie(w, csrf)

	// Response (still returns tokens for API clients; browsers use cookies)
	response := map[string]interface{}{
		"tokens":  tokenResponse,
		"message": "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LogoutHandler handles user logout requests
func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Extract token (remove "Bearer " prefix)
	tokenString := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	// Get user ID from token
	userResponse, err := h.authService.ValidateToken(ctx, tokenString)
	if err != nil {
		h.logger.Error("Invalid token for logout", zap.Error(err))
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Logout user (blacklist token)
	if err := h.authService.LogoutUser(ctx, tokenString, userResponse.ID); err != nil {
		h.logger.Error("Logout failed", zap.Error(err), zap.String("user_id", userResponse.ID))
		http.Error(w, "Logout failed", http.StatusInternalServerError)
		return
	}

	// Clear cookies
	h.clearCookie(w, h.config.Auth.RefreshCookieName)
	h.clearCookie(w, h.config.Auth.CSRFCookieName)

	// Response
	response := map[string]interface{}{
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RefreshTokenHandler handles token refresh requests
func (h *AuthHandler) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// CSRF validation: require header X-CSRF-Token to match CSRF cookie
	csrfHeader := r.Header.Get("X-CSRF-Token")
	csrfCookie, _ := r.Cookie(h.config.Auth.CSRFCookieName)
	if csrfCookie == nil || csrfHeader == "" || !strings.EqualFold(csrfCookie.Value, csrfHeader) {
		http.Error(w, "CSRF validation failed", http.StatusForbidden)
		return
	}

	// Try to read refresh token from cookie first; fallback to body for non-browser clients
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.RefreshToken == "" {
		if c, err := r.Cookie(h.config.Auth.RefreshCookieName); err == nil {
			req.RefreshToken = c.Value
		}
	}
	if req.RefreshToken == "" {
		http.Error(w, "refresh token required", http.StatusUnauthorized)
		return
	}

	// Refresh token
	tokenResponse, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		h.logger.WithComponent("auth").Error("Token refresh failed", "error", err)
		http.Error(w, "Token refresh failed", http.StatusUnauthorized)
		return
	}

	// Rotate cookies: set new refresh cookie and rotate CSRF token
	h.setRefreshCookie(w, tokenResponse.RefreshToken, tokenResponse.ExpiresIn)
	h.setCSRFCookie(w, h.generateCSRFToken())

	// Response
	response := map[string]interface{}{
		"tokens":  tokenResponse,
		"message": "Token refreshed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Record metrics
	h.metrics.RecordBusinessClassification("success", "token_refresh")
}

// VerifyEmailHandler handles email verification requests
func (h *AuthHandler) VerifyEmailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Verification token required", http.StatusBadRequest)
		return
	}

	// Verify email
	if err := h.authService.VerifyEmail(ctx, token); err != nil {
		h.logger.WithComponent("auth").Error("Email verification failed", "error", err, "token", token)

		// Handle different error types
		if err.Error() == "verification token has expired" {
			http.Error(w, "Verification token has expired", http.StatusGone)
		} else if err.Error() == "verification token has already been used" {
			http.Error(w, "Email already verified", http.StatusConflict)
		} else {
			http.Error(w, "Email verification failed", http.StatusBadRequest)
		}
		return
	}

	// Response
	response := map[string]interface{}{
		"message": "Email verified successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Record metrics
	h.metrics.RecordBusinessClassification("success", "email_verification")
}

// RequestPasswordResetHandler handles password reset requests
func (h *AuthHandler) RequestPasswordResetHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithComponent("auth").Error("Failed to decode password reset request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create password reset token
	resetToken, err := h.authService.CreatePasswordResetToken(ctx, req.Email)
	if err != nil {
		h.logger.WithComponent("auth").Error("Password reset token creation failed", "error", err, "email", req.Email)
		// Always return success to prevent email enumeration
	}

	// Response (always successful to prevent email enumeration)
	response := map[string]interface{}{
		"message": "If the email exists, a password reset link has been sent",
	}

	if resetToken != nil {
		response["reset_token"] = resetToken.Token // Only for testing/development
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Record metrics
	h.metrics.RecordBusinessClassification("success", "password_reset_request")
}

// ResetPasswordHandler handles password reset confirmation requests
func (h *AuthHandler) ResetPasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithComponent("auth").Error("Failed to decode password reset confirmation request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Reset password
	if err := h.authService.ResetPassword(ctx, req.Token, req.NewPassword); err != nil {
		h.logger.WithComponent("auth").Error("Password reset failed", "error", err, "token", req.Token)

		// Handle different error types
		if err.Error() == "reset token has expired" {
			http.Error(w, "Reset token has expired", http.StatusGone)
		} else if err.Error() == "reset token has already been used" {
			http.Error(w, "Reset token has already been used", http.StatusConflict)
		} else {
			http.Error(w, "Password reset failed", http.StatusBadRequest)
		}
		return
	}

	// Response
	response := map[string]interface{}{
		"message": "Password reset successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Record metrics
	h.metrics.RecordBusinessClassification("success", "password_reset")
}

// ChangePasswordHandler handles password change requests
func (h *AuthHandler) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context (set by auth middleware)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithComponent("auth").Error("Failed to decode change password request", "error", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Change password
	if err := h.authService.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword); err != nil {
		h.logger.WithComponent("auth").Error("Password change failed", "error", err, "user_id", userID)

		if err.Error() == "current password is incorrect" {
			http.Error(w, "Current password is incorrect", http.StatusBadRequest)
		} else {
			http.Error(w, "Password change failed", http.StatusInternalServerError)
		}
		return
	}

	// Response
	response := map[string]interface{}{
		"message": "Password changed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	// Record metrics
	h.metrics.RecordBusinessClassification("success", "password_change")
}

// ProfileHandler returns the current user's profile
func (h *AuthHandler) ProfileHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context (set by auth middleware)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		http.Error(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get user profile
	user, err := h.authService.GetUserByID(ctx, userID)
	if err != nil {
		h.logger.WithComponent("auth").Error("Failed to get user profile", "error", err, "user_id", userID)
		http.Error(w, "Failed to get user profile", http.StatusInternalServerError)
		return
	}

	// Response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}
