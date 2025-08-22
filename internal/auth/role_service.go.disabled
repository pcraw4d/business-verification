package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// RoleService handles role assignment and validation operations
type RoleService struct {
	db     database.Database
	logger *observability.Logger
	rbac   *RBACService
}

// NewRoleService creates a new role service instance
func NewRoleService(db database.Database, logger *observability.Logger, rbac *RBACService) *RoleService {
	return &RoleService{
		db:     db,
		logger: logger,
		rbac:   rbac,
	}
}

// AssignRoleRequest represents a role assignment request
type AssignRoleRequest struct {
	UserID      string     `json:"user_id" validate:"required"`
	Role        Role       `json:"role" validate:"required"`
	AssignedBy  string     `json:"assigned_by" validate:"required"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Reason      string     `json:"reason,omitempty"`
	ForceUpdate bool       `json:"force_update,omitempty"`
}

// RoleAssignmentResponse represents the response after role assignment
type RoleAssignmentResponse struct {
	ID            string       `json:"id"`
	UserID        string       `json:"user_id"`
	Role          Role         `json:"role"`
	AssignedBy    string       `json:"assigned_by"`
	AssignedAt    time.Time    `json:"assigned_at"`
	ExpiresAt     *time.Time   `json:"expires_at,omitempty"`
	IsActive      bool         `json:"is_active"`
	PreviousRole  *Role        `json:"previous_role,omitempty"`
	EffectiveRole Role         `json:"effective_role"`
	Permissions   []Permission `json:"permissions"`
	CreatedAt     time.Time    `json:"created_at"`
}

// UserRoleInfo represents comprehensive user role information
type UserRoleInfo struct {
	UserID      string                     `json:"user_id"`
	CurrentRole Role                       `json:"current_role"`
	Permissions []Permission               `json:"permissions"`
	Assignment  *database.RoleAssignment   `json:"assignment,omitempty"`
	History     []*database.RoleAssignment `json:"history,omitempty"`
	CanAssignTo []Role                     `json:"can_assign_to,omitempty"`
}

// AssignRole assigns a role to a user with proper validation and audit trail
func (rs *RoleService) AssignRole(ctx context.Context, request *AssignRoleRequest) (*RoleAssignmentResponse, error) {
	// Validate the request
	if err := rs.validateAssignRoleRequest(ctx, request); err != nil {
		rs.logger.WithComponent("role_service").WithError(err).Error("Role assignment validation failed")
		return nil, fmt.Errorf("role assignment validation failed: %w", err)
	}

	// Check if assigner has permission to assign this role
	assignerRole, err := rs.getUserRole(ctx, request.AssignedBy)
	if err != nil {
		return nil, fmt.Errorf("failed to get assigner role: %w", err)
	}

	if !CanAssignRole(assignerRole, request.Role) {
		return nil, fmt.Errorf("user with role %s cannot assign role %s", assignerRole, request.Role)
	}

	// Get current role assignment for audit trail
	currentAssignment, err := rs.db.GetActiveRoleAssignmentByUserID(ctx, request.UserID)
	var previousRole *Role
	if err == nil && currentAssignment != nil {
		role := Role(currentAssignment.Role)
		previousRole = &role

		// Deactivate current assignment if not forcing update
		if !request.ForceUpdate {
			if err := rs.db.DeactivateRoleAssignment(ctx, currentAssignment.ID); err != nil {
				return nil, fmt.Errorf("failed to deactivate current role assignment: %w", err)
			}
		}
	}

	// Generate unique ID for the new assignment
	assignmentID, err := rs.generateAssignmentID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate assignment ID: %w", err)
	}

	// Create new role assignment
	now := time.Now()
	assignment := &database.RoleAssignment{
		ID:         assignmentID,
		UserID:     request.UserID,
		Role:       string(request.Role),
		AssignedBy: request.AssignedBy,
		AssignedAt: now,
		ExpiresAt:  request.ExpiresAt,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Create the assignment in database
	if err := rs.db.CreateRoleAssignment(ctx, assignment); err != nil {
		return nil, fmt.Errorf("failed to create role assignment: %w", err)
	}

	// Update user's role in the users table for quick access
	if err := rs.updateUserRole(ctx, request.UserID, request.Role); err != nil {
		rs.logger.WithComponent("role_service").WithError(err).Error("Failed to update user role in users table")
		// Don't fail the assignment, just log the error
	}

	// Log the role assignment event
	rs.logger.WithComponent("role_service").WithFields(map[string]interface{}{
		"user_id":       request.UserID,
		"new_role":      request.Role,
		"previous_role": previousRole,
		"assigned_by":   request.AssignedBy,
		"assignment_id": assignmentID,
		"expires_at":    request.ExpiresAt,
	}).Info("Role assigned successfully")

	// Get effective permissions for the new role
	permissions := GetRolePermissions(request.Role)

	// Create response
	response := &RoleAssignmentResponse{
		ID:            assignment.ID,
		UserID:        assignment.UserID,
		Role:          request.Role,
		AssignedBy:    assignment.AssignedBy,
		AssignedAt:    assignment.AssignedAt,
		ExpiresAt:     assignment.ExpiresAt,
		IsActive:      assignment.IsActive,
		PreviousRole:  previousRole,
		EffectiveRole: request.Role,
		Permissions:   permissions,
		CreatedAt:     assignment.CreatedAt,
	}

	return response, nil
}

// GetUserRoleInfo retrieves comprehensive role information for a user
func (rs *RoleService) GetUserRoleInfo(ctx context.Context, userID string, includeHistory bool) (*UserRoleInfo, error) {
	// Get current active role assignment
	assignment, err := rs.db.GetActiveRoleAssignmentByUserID(ctx, userID)
	if err != nil {
		// If no assignment found, check user's default role
		user, userErr := rs.db.GetUserByID(ctx, userID)
		if userErr != nil {
			return nil, fmt.Errorf("user not found: %w", userErr)
		}

		// Default to user role if no assignment exists
		currentRole := Role(user.Role)
		if !IsValidRole(currentRole) {
			currentRole = RoleUser // Default fallback
		}

		info := &UserRoleInfo{
			UserID:      userID,
			CurrentRole: currentRole,
			Permissions: GetRolePermissions(currentRole),
		}

		return info, nil
	}

	currentRole := Role(assignment.Role)
	permissions := GetRolePermissions(currentRole)

	info := &UserRoleInfo{
		UserID:      userID,
		CurrentRole: currentRole,
		Permissions: permissions,
		Assignment:  assignment,
	}

	// Include role assignment history if requested
	if includeHistory {
		history, err := rs.db.GetRoleAssignmentsByUserID(ctx, userID)
		if err != nil {
			rs.logger.WithComponent("role_service").WithError(err).Error("Failed to get role assignment history")
		} else {
			info.History = history
		}
	}

	// Determine what roles this user can assign to others
	info.CanAssignTo = rs.getAssignableRoles(currentRole)

	return info, nil
}

// ValidateRoleAssignment validates if a role assignment is valid and active
func (rs *RoleService) ValidateRoleAssignment(ctx context.Context, userID string) (*Role, error) {
	assignment, err := rs.db.GetActiveRoleAssignmentByUserID(ctx, userID)
	if err != nil {
		// Fall back to user's default role
		user, userErr := rs.db.GetUserByID(ctx, userID)
		if userErr != nil {
			return nil, fmt.Errorf("user not found: %w", userErr)
		}

		role := Role(user.Role)
		if !IsValidRole(role) {
			role = RoleUser
		}
		return &role, nil
	}

	// Check if assignment has expired
	if assignment.ExpiresAt != nil && assignment.ExpiresAt.Before(time.Now()) {
		// Deactivate expired assignment
		if err := rs.db.DeactivateRoleAssignment(ctx, assignment.ID); err != nil {
			rs.logger.WithComponent("role_service").WithError(err).Error("Failed to deactivate expired role assignment")
		}

		// Fall back to user's default role
		user, err := rs.db.GetUserByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("user not found after role expiration: %w", err)
		}

		role := Role(user.Role)
		if !IsValidRole(role) {
			role = RoleUser
		}
		return &role, nil
	}

	role := Role(assignment.Role)
	return &role, nil
}

// RevokeRole revokes a user's current role assignment
func (rs *RoleService) RevokeRole(ctx context.Context, userID, revokedBy string, reason string) error {
	// Check if revoker has permission
	revokerRole, err := rs.getUserRole(ctx, revokedBy)
	if err != nil {
		return fmt.Errorf("failed to get revoker role: %w", err)
	}

	// Only managers and above can revoke roles
	if revokerRole != RoleManager && revokerRole != RoleAdmin {
		return fmt.Errorf("insufficient permissions to revoke roles")
	}

	// Get current assignment
	assignment, err := rs.db.GetActiveRoleAssignmentByUserID(ctx, userID)
	if err != nil {
		return fmt.Errorf("no active role assignment found for user: %w", err)
	}

	// Deactivate the assignment
	if err := rs.db.DeactivateRoleAssignment(ctx, assignment.ID); err != nil {
		return fmt.Errorf("failed to revoke role assignment: %w", err)
	}

	// Update user's role to default user role
	if err := rs.updateUserRole(ctx, userID, RoleUser); err != nil {
		rs.logger.WithComponent("role_service").WithError(err).Error("Failed to update user role after revocation")
	}

	// Log the revocation
	rs.logger.WithComponent("role_service").WithFields(map[string]interface{}{
		"user_id":       userID,
		"revoked_role":  assignment.Role,
		"revoked_by":    revokedBy,
		"reason":        reason,
		"assignment_id": assignment.ID,
	}).Info("Role revoked")

	return nil
}

// CleanupExpiredRoleAssignments deactivates expired role assignments
func (rs *RoleService) CleanupExpiredRoleAssignments(ctx context.Context) error {
	if err := rs.db.DeleteExpiredRoleAssignments(ctx); err != nil {
		return fmt.Errorf("failed to cleanup expired role assignments: %w", err)
	}

	rs.logger.WithComponent("role_service").Info("Expired role assignments cleaned up")
	return nil
}

// Helper methods

func (rs *RoleService) validateAssignRoleRequest(ctx context.Context, request *AssignRoleRequest) error {
	if request.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if !IsValidRole(request.Role) {
		return fmt.Errorf("invalid role: %s", request.Role)
	}

	if request.AssignedBy == "" {
		return fmt.Errorf("assigned_by is required")
	}

	// Check if user exists
	_, err := rs.db.GetUserByID(ctx, request.UserID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	// Check if assigner exists
	_, err = rs.db.GetUserByID(ctx, request.AssignedBy)
	if err != nil {
		return fmt.Errorf("assigner not found: %w", err)
	}

	// Validate expiration date
	if request.ExpiresAt != nil && request.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("expiration date cannot be in the past")
	}

	return nil
}

func (rs *RoleService) getUserRole(ctx context.Context, userID string) (Role, error) {
	roleInfo, err := rs.GetUserRoleInfo(ctx, userID, false)
	if err != nil {
		return "", err
	}
	return roleInfo.CurrentRole, nil
}

func (rs *RoleService) updateUserRole(ctx context.Context, userID string, role Role) error {
	user, err := rs.db.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	user.Role = string(role)
	user.UpdatedAt = time.Now()

	return rs.db.UpdateUser(ctx, user)
}

func (rs *RoleService) getAssignableRoles(userRole Role) []Role {
	switch userRole {
	case RoleAdmin:
		return []Role{RoleGuest, RoleUser, RoleAnalyst, RoleManager}
	case RoleManager:
		return []Role{RoleGuest, RoleUser, RoleAnalyst}
	default:
		return []Role{} // Other roles cannot assign roles
	}
}

func (rs *RoleService) generateAssignmentID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "role_" + hex.EncodeToString(bytes), nil
}
