package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
	"golang.org/x/crypto/bcrypt"
)

// AdminService handles administrative user management operations
type AdminService struct {
	db            database.Database
	logger        *observability.Logger
	authService   *AuthService
	roleService   *RoleService
	apiKeyService *APIKeyService
}

// NewAdminService creates a new admin service instance
func NewAdminService(db database.Database, logger *observability.Logger, authService *AuthService, roleService *RoleService, apiKeyService *APIKeyService) *AdminService {
	return &AdminService{
		db:            db,
		logger:        logger,
		authService:   authService,
		roleService:   roleService,
		apiKeyService: apiKeyService,
	}
}

// UserManagementRequest represents a request for user management operations
type UserManagementRequest struct {
	AdminUserID  string                 `json:"admin_user_id" validate:"required"`
	TargetUserID string                 `json:"target_user_id" validate:"required"`
	Action       string                 `json:"action" validate:"required"` // create, update, delete, activate, deactivate
	Data         map[string]interface{} `json:"data,omitempty"`
}

// UserManagementResponse represents the response for user management operations
type UserManagementResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	User    *UserResponse `json:"user,omitempty"`
}

// ListUsersRequest represents a request to list users with filters
type ListUsersRequest struct {
	AdminUserID string `json:"admin_user_id" validate:"required"`
	Role        Role   `json:"role,omitempty"`
	Status      string `json:"status,omitempty"`
	Search      string `json:"search,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users  []*UserResponse `json:"users"`
	Total  int             `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

// SystemStatsResponse represents system statistics for admin dashboard
type SystemStatsResponse struct {
	TotalUsers           int `json:"total_users"`
	ActiveUsers          int `json:"active_users"`
	InactiveUsers        int `json:"inactive_users"`
	TotalAPIKeys         int `json:"total_api_keys"`
	ActiveAPIKeys        int `json:"active_api_keys"`
	TotalRoleAssignments int `json:"total_role_assignments"`
	RecentLogins         int `json:"recent_logins"`
	FailedLogins         int `json:"failed_logins"`
}

// CreateUser creates a new user with admin privileges
func (as *AdminService) CreateUser(ctx context.Context, request *UserManagementRequest) (*UserManagementResponse, error) {
	// Validate admin permissions
	if err := as.validateAdminPermissions(ctx, request.AdminUserID, PermissionCreateUser); err != nil {
		return nil, err
	}

	// Extract user data from request
	userData := request.Data
	if userData == nil {
		return nil, fmt.Errorf("user data is required")
	}

	// Create user object
	user := &database.User{
		ID:                  generateUUID(),
		Email:               userData["email"].(string),
		Username:            userData["username"].(string),
		FirstName:           userData["first_name"].(string),
		LastName:            userData["last_name"].(string),
		Company:             userData["company"].(string),
		Role:                userData["role"].(string),
		Status:              "active",
		EmailVerified:       false,
		FailedLoginAttempts: 0,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Hash password if provided
	if password, ok := userData["password"].(string); ok && password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			as.logger.WithComponent("admin_service").WithError(err).Error("Failed to hash password")
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.PasswordHash = string(hashedPassword)
	}

	// Store user in database
	if err := as.db.CreateUser(ctx, user); err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to create user")
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Log the creation
	as.logger.WithComponent("admin_service").WithFields(map[string]interface{}{
		"admin_user_id":  request.AdminUserID,
		"target_user_id": user.ID,
		"email":          user.Email,
		"role":           user.Role,
	}).Info("User created by admin")

	// Return response
	response := &UserManagementResponse{
		Success: true,
		Message: "User created successfully",
		User: &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Company:   user.Company,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	return response, nil
}

// UpdateUser updates an existing user with admin privileges
func (as *AdminService) UpdateUser(ctx context.Context, request *UserManagementRequest) (*UserManagementResponse, error) {
	// Validate admin permissions
	if err := as.validateAdminPermissions(ctx, request.AdminUserID, PermissionUpdateUser); err != nil {
		return nil, err
	}

	// Get existing user
	user, err := as.db.GetUserByID(ctx, request.TargetUserID)
	if err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to get user for update")
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Update user fields from request data
	userData := request.Data
	if userData != nil {
		if email, ok := userData["email"].(string); ok {
			user.Email = email
		}
		if username, ok := userData["username"].(string); ok {
			user.Username = username
		}
		if firstName, ok := userData["first_name"].(string); ok {
			user.FirstName = firstName
		}
		if lastName, ok := userData["last_name"].(string); ok {
			user.LastName = lastName
		}
		if company, ok := userData["company"].(string); ok {
			user.Company = company
		}
		if role, ok := userData["role"].(string); ok {
			user.Role = role
		}
		if status, ok := userData["status"].(string); ok {
			user.Status = status
		}
		if password, ok := userData["password"].(string); ok && password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				as.logger.WithComponent("admin_service").WithError(err).Error("Failed to hash password")
				return nil, fmt.Errorf("failed to hash password: %w", err)
			}
			user.PasswordHash = string(hashedPassword)
		}
		user.UpdatedAt = time.Now()
	}

	// Update user in database
	if err := as.db.UpdateUser(ctx, user); err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to update user")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Log the update
	as.logger.WithComponent("admin_service").WithFields(map[string]interface{}{
		"admin_user_id":  request.AdminUserID,
		"target_user_id": user.ID,
		"email":          user.Email,
		"role":           user.Role,
	}).Info("User updated by admin")

	// Return response
	response := &UserManagementResponse{
		Success: true,
		Message: "User updated successfully",
		User: &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Company:   user.Company,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	return response, nil
}

// DeleteUser deletes a user with admin privileges
func (as *AdminService) DeleteUser(ctx context.Context, request *UserManagementRequest) (*UserManagementResponse, error) {
	// Validate admin permissions
	if err := as.validateAdminPermissions(ctx, request.AdminUserID, PermissionDeleteUser); err != nil {
		return nil, err
	}

	// Get user before deletion for logging
	user, err := as.db.GetUserByID(ctx, request.TargetUserID)
	if err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to get user for deletion")
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Prevent deletion of admin users (safety check)
	if user.Role == string(RoleAdmin) {
		return nil, fmt.Errorf("cannot delete admin users")
	}

	// Delete user from database
	if err := as.db.DeleteUser(ctx, request.TargetUserID); err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to delete user")
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	// Log the deletion
	as.logger.WithComponent("admin_service").WithFields(map[string]interface{}{
		"admin_user_id":  request.AdminUserID,
		"target_user_id": user.ID,
		"email":          user.Email,
		"role":           user.Role,
	}).Info("User deleted by admin")

	// Return response
	response := &UserManagementResponse{
		Success: true,
		Message: "User deleted successfully",
	}

	return response, nil
}

// ActivateUser activates a deactivated user
func (as *AdminService) ActivateUser(ctx context.Context, request *UserManagementRequest) (*UserManagementResponse, error) {
	// Validate admin permissions
	if err := as.validateAdminPermissions(ctx, request.AdminUserID, PermissionUpdateUser); err != nil {
		return nil, err
	}

	// Get user
	user, err := as.db.GetUserByID(ctx, request.TargetUserID)
	if err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to get user for activation")
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Activate user
	user.Status = "active"
	user.UpdatedAt = time.Now()

	// Update user in database
	if err := as.db.UpdateUser(ctx, user); err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to activate user")
		return nil, fmt.Errorf("failed to activate user: %w", err)
	}

	// Log the activation
	as.logger.WithComponent("admin_service").WithFields(map[string]interface{}{
		"admin_user_id":  request.AdminUserID,
		"target_user_id": user.ID,
		"email":          user.Email,
	}).Info("User activated by admin")

	// Return response
	response := &UserManagementResponse{
		Success: true,
		Message: "User activated successfully",
		User: &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Company:   user.Company,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	return response, nil
}

// DeactivateUser deactivates an active user
func (as *AdminService) DeactivateUser(ctx context.Context, request *UserManagementRequest) (*UserManagementResponse, error) {
	// Validate admin permissions
	if err := as.validateAdminPermissions(ctx, request.AdminUserID, PermissionUpdateUser); err != nil {
		return nil, err
	}

	// Get user
	user, err := as.db.GetUserByID(ctx, request.TargetUserID)
	if err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to get user for deactivation")
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Prevent deactivation of admin users (safety check)
	if user.Role == string(RoleAdmin) {
		return nil, fmt.Errorf("cannot deactivate admin users")
	}

	// Deactivate user
	user.Status = "inactive"
	user.UpdatedAt = time.Now()

	// Update user in database
	if err := as.db.UpdateUser(ctx, user); err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to deactivate user")
		return nil, fmt.Errorf("failed to deactivate user: %w", err)
	}

	// Log the deactivation
	as.logger.WithComponent("admin_service").WithFields(map[string]interface{}{
		"admin_user_id":  request.AdminUserID,
		"target_user_id": user.ID,
		"email":          user.Email,
	}).Info("User deactivated by admin")

	// Return response
	response := &UserManagementResponse{
		Success: true,
		Message: "User deactivated successfully",
		User: &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Company:   user.Company,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}

	return response, nil
}

// ListUsers lists users with optional filtering
func (as *AdminService) ListUsers(ctx context.Context, request *ListUsersRequest) (*ListUsersResponse, error) {
	// Validate admin permissions
	if err := as.validateAdminPermissions(ctx, request.AdminUserID, PermissionViewUsers); err != nil {
		return nil, err
	}

	// Set default limit if not provided
	if request.Limit <= 0 {
		request.Limit = 50
	}

	// Get users from database
	users, err := as.db.ListUsers(ctx, request.Limit, request.Offset)
	if err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to list users")
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Convert to response format
	userResponses := make([]*UserResponse, 0, len(users))
	for _, user := range users {
		// Apply filters if specified
		if request.Role != "" && Role(user.Role) != request.Role {
			continue
		}
		if request.Status != "" && user.Status != request.Status {
			continue
		}
		if request.Search != "" {
			// Simple search implementation
			searchLower := strings.ToLower(request.Search)
			emailLower := strings.ToLower(user.Email)
			usernameLower := strings.ToLower(user.Username)
			firstNameLower := strings.ToLower(user.FirstName)
			lastNameLower := strings.ToLower(user.LastName)

			if !strings.Contains(emailLower, searchLower) &&
				!strings.Contains(usernameLower, searchLower) &&
				!strings.Contains(firstNameLower, searchLower) &&
				!strings.Contains(lastNameLower, searchLower) {
				continue
			}
		}

		userResponse := &UserResponse{
			ID:        user.ID,
			Email:     user.Email,
			Username:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Company:   user.Company,
			Role:      user.Role,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
		userResponses = append(userResponses, userResponse)
	}

	// Log the listing
	as.logger.WithComponent("admin_service").WithFields(map[string]interface{}{
		"admin_user_id": request.AdminUserID,
		"total_users":   len(userResponses),
		"limit":         request.Limit,
		"offset":        request.Offset,
	}).Info("Users listed by admin")

	response := &ListUsersResponse{
		Users:  userResponses,
		Total:  len(userResponses),
		Limit:  request.Limit,
		Offset: request.Offset,
	}

	return response, nil
}

// GetSystemStats returns system statistics for admin dashboard
func (as *AdminService) GetSystemStats(ctx context.Context, adminUserID string) (*SystemStatsResponse, error) {
	// Validate admin permissions
	if err := as.validateAdminPermissions(ctx, adminUserID, PermissionViewMetrics); err != nil {
		return nil, err
	}

	// Get all users for statistics
	users, err := as.db.ListUsers(ctx, 1000, 0) // Get all users for stats
	if err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to get users for stats")
		return nil, fmt.Errorf("failed to get users for stats: %w", err)
	}

	// Calculate statistics
	stats := &SystemStatsResponse{
		TotalUsers: len(users),
	}

	for _, user := range users {
		if user.Status == "active" {
			stats.ActiveUsers++
		} else {
			stats.InactiveUsers++
		}
	}

	// Get API key statistics
	apiKeys, err := as.db.ListAPIKeysByUserID(ctx, "") // Get all API keys
	if err == nil {
		stats.TotalAPIKeys = len(apiKeys)
		for _, apiKey := range apiKeys {
			if apiKey.Status == "active" {
				stats.ActiveAPIKeys++
			}
		}
	}

	// Log the stats request
	as.logger.WithComponent("admin_service").WithFields(map[string]interface{}{
		"admin_user_id":  adminUserID,
		"total_users":    stats.TotalUsers,
		"active_users":   stats.ActiveUsers,
		"total_api_keys": stats.TotalAPIKeys,
	}).Info("System stats retrieved by admin")

	return stats, nil
}

// validateAdminPermissions validates that the user has admin permissions
func (as *AdminService) validateAdminPermissions(ctx context.Context, adminUserID string, permission Permission) error {
	// Get admin user
	adminUser, err := as.db.GetUserByID(ctx, adminUserID)
	if err != nil {
		as.logger.WithComponent("admin_service").WithError(err).Error("Failed to get admin user")
		return fmt.Errorf("admin user not found: %w", err)
	}

	// Check if user is active
	if adminUser.Status != "active" {
		return fmt.Errorf("admin user is not active")
	}

	// Check if user has admin role
	if adminUser.Role != string(RoleAdmin) {
		return fmt.Errorf("user does not have admin privileges")
	}

	// Check specific permission
	if !HasPermission(Role(adminUser.Role), permission) {
		return fmt.Errorf("user does not have permission: %s", permission)
	}

	return nil
}
