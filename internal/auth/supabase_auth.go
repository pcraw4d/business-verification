package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/supabase-community/gotrue-go/types"
	supa "github.com/supabase-community/supabase-go"
)

// SupabaseAuth represents a Supabase authentication client
type SupabaseAuth struct {
	client *supa.Client
	logger *observability.Logger
}

// NewSupabaseAuth creates a new Supabase auth client
func NewSupabaseAuth(cfg *config.SupabaseConfig, logger *observability.Logger) (*SupabaseAuth, error) {
	client, err := supa.NewClient(cfg.URL, cfg.APIKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return &SupabaseAuth{
		client: client,
		logger: logger,
	}, nil
}

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Active    bool      `json:"active"`
}

// AuthenticateUser authenticates a user with email and password
func (s *SupabaseAuth) AuthenticateUser(ctx context.Context, email, password string) (*User, error) {
	s.logger.Debug("Authenticating user with Supabase", "email", email)

	credentials := types.UserCredentials{
		Email:    email,
		Password: password,
	}

	auth, err := s.client.Auth.SignIn(ctx, credentials)
	if err != nil {
		s.logger.Error("Failed to authenticate user with Supabase", "error", err)
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	user := &User{
		ID:        auth.User.ID,
		Email:     auth.User.Email,
		Name:      auth.User.UserMetadata["name"].(string),
		CreatedAt: auth.User.CreatedAt,
		UpdatedAt: auth.User.UpdatedAt,
		Active:    auth.User.EmailConfirmedAt != nil,
	}

	s.logger.Info("Successfully authenticated user with Supabase", "user_id", user.ID)
	return user, nil
}

// RegisterUser registers a new user
func (s *SupabaseAuth) RegisterUser(ctx context.Context, email, password, name string) (*User, error) {
	s.logger.Debug("Registering user with Supabase", "email", email)

	credentials := types.UserCredentials{
		Email:    email,
		Password: password,
		Data: map[string]interface{}{
			"name": name,
		},
	}

	auth, err := s.client.Auth.SignUp(ctx, credentials)
	if err != nil {
		s.logger.Error("Failed to register user with Supabase", "error", err)
		return nil, fmt.Errorf("registration failed: %w", err)
	}

	user := &User{
		ID:        auth.User.ID,
		Email:     auth.User.Email,
		Name:      name,
		CreatedAt: auth.User.CreatedAt,
		UpdatedAt: auth.User.UpdatedAt,
		Active:    false, // User needs to confirm email
	}

	s.logger.Info("Successfully registered user with Supabase", "user_id", user.ID)
	return user, nil
}

// RefreshToken refreshes an access token
func (s *SupabaseAuth) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	s.logger.Debug("Refreshing token with Supabase")

	auth, err := s.client.Auth.RefreshUser(ctx, refreshToken, "")
	if err != nil {
		s.logger.Error("Failed to refresh token with Supabase", "error", err)
		return "", fmt.Errorf("token refresh failed: %w", err)
	}

	s.logger.Info("Successfully refreshed token with Supabase")
	return auth.AccessToken, nil
}

// ValidateToken validates an access token
func (s *SupabaseAuth) ValidateToken(ctx context.Context, accessToken string) (*User, error) {
	s.logger.Debug("Validating token with Supabase")

	user, err := s.client.Auth.GetUser(ctx, accessToken)
	if err != nil {
		s.logger.Error("Failed to validate token with Supabase", "error", err)
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	result := &User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.UserMetadata["name"].(string),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Active:    user.EmailConfirmedAt != nil,
	}

	s.logger.Info("Successfully validated token with Supabase", "user_id", result.ID)
	return result, nil
}

// LogoutUser logs out a user
func (s *SupabaseAuth) LogoutUser(ctx context.Context, accessToken string) error {
	s.logger.Debug("Logging out user with Supabase")

	err := s.client.Auth.SignOut(ctx, accessToken)
	if err != nil {
		s.logger.Error("Failed to logout user with Supabase", "error", err)
		return fmt.Errorf("logout failed: %w", err)
	}

	s.logger.Info("Successfully logged out user with Supabase")
	return nil
}

// RequestPasswordReset requests a password reset
func (s *SupabaseAuth) RequestPasswordReset(ctx context.Context, email string) error {
	s.logger.Debug("Requesting password reset with Supabase", "email", email)

	options := types.Options{
		RedirectTo: "https://your-app.com/reset-password",
	}

	err := s.client.Auth.ResetPasswordForEmail(ctx, email, options)
	if err != nil {
		s.logger.Error("Failed to request password reset with Supabase", "error", err)
		return fmt.Errorf("password reset request failed: %w", err)
	}

	s.logger.Info("Successfully requested password reset with Supabase", "email", email)
	return nil
}

// ResetPassword resets a user's password
func (s *SupabaseAuth) ResetPassword(ctx context.Context, accessToken, newPassword string) error {
	s.logger.Debug("Resetting password with Supabase")

	credentials := types.UserCredentials{
		Password: newPassword,
	}

	_, err := s.client.Auth.UpdateUser(ctx, accessToken, credentials)
	if err != nil {
		s.logger.Error("Failed to reset password with Supabase", "error", err)
		return fmt.Errorf("password reset failed: %w", err)
	}

	s.logger.Info("Successfully reset password with Supabase")
	return nil
}

// ChangePassword changes a user's password
func (s *SupabaseAuth) ChangePassword(ctx context.Context, accessToken, currentPassword, newPassword string) error {
	s.logger.Debug("Changing password with Supabase")

	credentials := types.UserCredentials{
		Password: newPassword,
	}

	_, err := s.client.Auth.UpdateUser(ctx, accessToken, credentials)
	if err != nil {
		s.logger.Error("Failed to change password with Supabase", "error", err)
		return fmt.Errorf("password change failed: %w", err)
	}

	s.logger.Info("Successfully changed password with Supabase")
	return nil
}

// GetUserByID retrieves a user by ID
func (s *SupabaseAuth) GetUserByID(ctx context.Context, userID string) (*User, error) {
	s.logger.Debug("Getting user by ID with Supabase", "user_id", userID)

	// Use service role key for admin operations
	adminClient, err := supa.NewClient(s.client.SupabaseURL, s.client.SupabaseKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin client: %w", err)
	}

	user, err := adminClient.Auth.AdminGetUser(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get user by ID with Supabase", "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	result := &User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.UserMetadata["name"].(string),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Active:    user.EmailConfirmedAt != nil,
	}

	s.logger.Info("Successfully retrieved user by ID with Supabase", "user_id", result.ID)
	return result, nil
}

// UpdateUser updates a user's information
func (s *SupabaseAuth) UpdateUser(ctx context.Context, accessToken string, updates map[string]interface{}) (*User, error) {
	s.logger.Debug("Updating user with Supabase")

	credentials := types.UserCredentials{
		Data: updates,
	}

	user, err := s.client.Auth.UpdateUser(ctx, accessToken, credentials)
	if err != nil {
		s.logger.Error("Failed to update user with Supabase", "error", err)
		return nil, fmt.Errorf("user update failed: %w", err)
	}

	result := &User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.UserMetadata["name"].(string),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Active:    user.EmailConfirmedAt != nil,
	}

	s.logger.Info("Successfully updated user with Supabase", "user_id", result.ID)
	return result, nil
}

// DeleteUser deletes a user
func (s *SupabaseAuth) DeleteUser(ctx context.Context, userID string) error {
	s.logger.Debug("Deleting user with Supabase", "user_id", userID)

	// Use service role key for admin operations
	adminClient, err := supa.NewClient(s.client.SupabaseURL, s.client.SupabaseKey, nil)
	if err != nil {
		return fmt.Errorf("failed to create admin client: %w", err)
	}

	err = adminClient.Auth.AdminDeleteUser(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to delete user with Supabase", "error", err)
		return fmt.Errorf("user deletion failed: %w", err)
	}

	s.logger.Info("Successfully deleted user with Supabase", "user_id", userID)
	return nil
}

// ListUsers lists all users (admin only)
func (s *SupabaseAuth) ListUsers(ctx context.Context, page, perPage int) ([]*User, error) {
	s.logger.Debug("Listing users with Supabase", "page", page, "per_page", perPage)

	// Use service role key for admin operations
	adminClient, err := supa.NewClient(s.client.SupabaseURL, s.client.SupabaseKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create admin client: %w", err)
	}

	users, err := adminClient.Auth.AdminListUsers(ctx, page, perPage)
	if err != nil {
		s.logger.Error("Failed to list users with Supabase", "error", err)
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var result []*User
	for _, user := range users {
		result = append(result, &User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.UserMetadata["name"].(string),
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
			Active:    user.EmailConfirmedAt != nil,
		})
	}

	s.logger.Info("Successfully listed users with Supabase", "count", len(result))
	return result, nil
}
