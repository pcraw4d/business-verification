package auth

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Role represents different user roles in the system
type Role string

const (
	// RoleGuest - Limited read-only access
	RoleGuest Role = "guest"

	// RoleUser - Standard user with basic functionality
	RoleUser Role = "user"

	// RoleAnalyst - Business analyst with enhanced classification access
	RoleAnalyst Role = "analyst"

	// RoleManager - Team manager with user management capabilities
	RoleManager Role = "manager"

	// RoleAdmin - Full system administrator
	RoleAdmin Role = "admin"

	// RoleSystem - System-level access for integrations
	RoleSystem Role = "system"
)

// Permission represents specific actions that can be performed
type Permission string

const (
	// Classification permissions
	PermissionClassifyBusiness     Permission = "classify:business"
	PermissionClassifyBatch        Permission = "classify:batch"
	PermissionViewClassification   Permission = "classify:view"
	PermissionExportClassification Permission = "classify:export"

	// Risk assessment permissions
	PermissionAssessRisk Permission = "risk:assess"
	PermissionViewRisk   Permission = "risk:view"
	PermissionExportRisk Permission = "risk:export"

	// Compliance permissions
	PermissionCheckCompliance  Permission = "compliance:check"
	PermissionViewCompliance   Permission = "compliance:view"
	PermissionExportCompliance Permission = "compliance:export"

	// User management permissions
	PermissionViewUsers   Permission = "users:view"
	PermissionCreateUser  Permission = "users:create"
	PermissionUpdateUser  Permission = "users:update"
	PermissionDeleteUser  Permission = "users:delete"
	PermissionManageRoles Permission = "users:manage_roles"

	// API key management permissions
	PermissionViewAPIKeys   Permission = "api_keys:view"
	PermissionCreateAPIKey  Permission = "api_keys:create"
	PermissionRevokeAPIKey  Permission = "api_keys:revoke"
	PermissionManageAPIKeys Permission = "api_keys:manage"

	// System administration permissions
	PermissionViewMetrics         Permission = "system:metrics"
	PermissionViewLogs            Permission = "system:logs"
	PermissionManageConfiguration Permission = "system:config"
	PermissionSystemBackup        Permission = "system:backup"

	// Audit and monitoring permissions
	PermissionViewAuditLogs   Permission = "audit:view"
	PermissionExportAuditLogs Permission = "audit:export"
)

// RolePermissions defines the permissions for each role
var RolePermissions = map[Role][]Permission{
	RoleGuest: {
		PermissionViewClassification,
		PermissionViewRisk,
		PermissionViewCompliance,
	},

	RoleUser: {
		// Classification permissions
		PermissionClassifyBusiness,
		PermissionViewClassification,
		PermissionExportClassification,

		// Risk assessment permissions
		PermissionAssessRisk,
		PermissionViewRisk,
		PermissionExportRisk,

		// Compliance permissions
		PermissionCheckCompliance,
		PermissionViewCompliance,
		PermissionExportCompliance,
	},

	RoleAnalyst: {
		// All user permissions plus:
		PermissionClassifyBusiness,
		PermissionClassifyBatch,
		PermissionViewClassification,
		PermissionExportClassification,

		PermissionAssessRisk,
		PermissionViewRisk,
		PermissionExportRisk,

		PermissionCheckCompliance,
		PermissionViewCompliance,
		PermissionExportCompliance,

		// Additional analyst permissions
		PermissionViewMetrics,
		PermissionViewAuditLogs,
	},

	RoleManager: {
		// All analyst permissions plus:
		PermissionClassifyBusiness,
		PermissionClassifyBatch,
		PermissionViewClassification,
		PermissionExportClassification,

		PermissionAssessRisk,
		PermissionViewRisk,
		PermissionExportRisk,

		PermissionCheckCompliance,
		PermissionViewCompliance,
		PermissionExportCompliance,

		PermissionViewMetrics,
		PermissionViewAuditLogs,
		PermissionExportAuditLogs,

		// User management permissions
		PermissionViewUsers,
		PermissionCreateUser,
		PermissionUpdateUser,
		PermissionManageRoles, // Can assign user and analyst roles

		// API key management
		PermissionViewAPIKeys,
		PermissionCreateAPIKey,
		PermissionRevokeAPIKey,
	},

	RoleAdmin: {
		// All permissions
		PermissionClassifyBusiness,
		PermissionClassifyBatch,
		PermissionViewClassification,
		PermissionExportClassification,

		PermissionAssessRisk,
		PermissionViewRisk,
		PermissionExportRisk,

		PermissionCheckCompliance,
		PermissionViewCompliance,
		PermissionExportCompliance,

		PermissionViewUsers,
		PermissionCreateUser,
		PermissionUpdateUser,
		PermissionDeleteUser,
		PermissionManageRoles,

		PermissionViewAPIKeys,
		PermissionCreateAPIKey,
		PermissionRevokeAPIKey,
		PermissionManageAPIKeys,

		PermissionViewMetrics,
		PermissionViewLogs,
		PermissionManageConfiguration,
		PermissionSystemBackup,

		PermissionViewAuditLogs,
		PermissionExportAuditLogs,
	},

	RoleSystem: {
		// System integration permissions
		PermissionClassifyBusiness,
		PermissionClassifyBatch,
		PermissionAssessRisk,
		PermissionCheckCompliance,
		PermissionViewMetrics,
	},
}

// RoleHierarchy defines the role hierarchy (higher roles inherit lower role permissions)
var RoleHierarchy = map[Role][]Role{
	RoleUser:    {RoleGuest},
	RoleAnalyst: {RoleUser, RoleGuest},
	RoleManager: {RoleAnalyst, RoleUser, RoleGuest},
	RoleAdmin:   {RoleManager, RoleAnalyst, RoleUser, RoleGuest},
}

// APIKey represents an API key for system integrations
type APIKey struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Key         string     `json:"key" db:"key"` // Hashed in database
	UserID      string     `json:"user_id" db:"user_id"`
	Role        Role       `json:"role" db:"role"`
	Permissions []string   `json:"permissions" db:"permissions"` // JSON array in DB
	ExpiresAt   *time.Time `json:"expires_at" db:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at" db:"last_used_at"`
	IsActive    bool       `json:"is_active" db:"is_active"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// APIKeyResponse represents API key data for responses (without sensitive fields)
type APIKeyResponse struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	UserID      string     `json:"user_id"`
	Role        Role       `json:"role"`
	Permissions []string   `json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// RoleAssignment represents a user's role assignment
type RoleAssignment struct {
	ID         string     `json:"id" db:"id"`
	UserID     string     `json:"user_id" db:"user_id"`
	Role       Role       `json:"role" db:"role"`
	AssignedBy string     `json:"assigned_by" db:"assigned_by"` // User ID who assigned the role
	AssignedAt time.Time  `json:"assigned_at" db:"assigned_at"`
	ExpiresAt  *time.Time `json:"expires_at" db:"expires_at"` // Optional role expiration
	IsActive   bool       `json:"is_active" db:"is_active"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// PermissionContext provides context for permission checking
type PermissionContext struct {
	UserID      string
	Role        Role
	Permissions []Permission
	ResourceID  string // For resource-specific permissions
	IPAddress   string
	UserAgent   string
}

// RBACService handles role-based access control
// RBACServiceInterface defines the interface for RBAC operations
type RBACServiceInterface interface {
	CheckPermission(ctx context.Context, userRole Role, method, path string) error
}

// RBACService handles role-based access control operations
type RBACService struct {
	authService *AuthService
}

// NewRBACService creates a new RBAC service
func NewRBACService(authService *AuthService) *RBACService {
	return &RBACService{
		authService: authService,
	}
}

// IsValidRole checks if a role is valid
func IsValidRole(role Role) bool {
	validRoles := []Role{RoleGuest, RoleUser, RoleAnalyst, RoleManager, RoleAdmin, RoleSystem}
	for _, validRole := range validRoles {
		if role == validRole {
			return true
		}
	}
	return false
}

// IsValidPermission checks if a permission is valid
func IsValidPermission(permission Permission) bool {
	allPermissions := []Permission{
		PermissionClassifyBusiness, PermissionClassifyBatch, PermissionViewClassification, PermissionExportClassification,
		PermissionAssessRisk, PermissionViewRisk, PermissionExportRisk,
		PermissionCheckCompliance, PermissionViewCompliance, PermissionExportCompliance,
		PermissionViewUsers, PermissionCreateUser, PermissionUpdateUser, PermissionDeleteUser, PermissionManageRoles,
		PermissionViewAPIKeys, PermissionCreateAPIKey, PermissionRevokeAPIKey, PermissionManageAPIKeys,
		PermissionViewMetrics, PermissionViewLogs, PermissionManageConfiguration, PermissionSystemBackup,
		PermissionViewAuditLogs, PermissionExportAuditLogs,
	}

	for _, validPermission := range allPermissions {
		if permission == validPermission {
			return true
		}
	}
	return false
}

// GetRolePermissions returns all permissions for a given role
func GetRolePermissions(role Role) []Permission {
	permissions := make([]Permission, 0)

	// Add direct permissions for the role
	if rolePerms, exists := RolePermissions[role]; exists {
		permissions = append(permissions, rolePerms...)
	}

	// Add inherited permissions from lower roles
	if inheritedRoles, exists := RoleHierarchy[role]; exists {
		for _, inheritedRole := range inheritedRoles {
			if inheritedPerms, exists := RolePermissions[inheritedRole]; exists {
				permissions = append(permissions, inheritedPerms...)
			}
		}
	}

	// Remove duplicates
	permissionMap := make(map[Permission]bool)
	uniquePermissions := make([]Permission, 0)

	for _, perm := range permissions {
		if !permissionMap[perm] {
			permissionMap[perm] = true
			uniquePermissions = append(uniquePermissions, perm)
		}
	}

	return uniquePermissions
}

// HasPermission checks if a role has a specific permission
func HasPermission(role Role, permission Permission) bool {
	permissions := GetRolePermissions(role)
	for _, perm := range permissions {
		if perm == permission {
			return true
		}
	}
	return false
}

// CanAssignRole checks if a user with a given role can assign another role
func CanAssignRole(assignerRole Role, targetRole Role) bool {
	// System role cannot be assigned by users
	if targetRole == RoleSystem {
		return false
	}

	// Admin can assign any role except system
	if assignerRole == RoleAdmin {
		return targetRole != RoleSystem
	}

	// Manager can assign user and analyst roles
	if assignerRole == RoleManager {
		return targetRole == RoleUser || targetRole == RoleAnalyst || targetRole == RoleGuest
	}

	// Other roles cannot assign roles
	return false
}

// ValidateAPIKeyPermissions checks if the given permissions are valid for an API key role
func ValidateAPIKeyPermissions(role Role, permissions []Permission) error {
	rolePermissions := GetRolePermissions(role)
	rolePermissionMap := make(map[Permission]bool)

	for _, perm := range rolePermissions {
		rolePermissionMap[perm] = true
	}

	for _, perm := range permissions {
		if !IsValidPermission(perm) {
			return fmt.Errorf("invalid permission: %s", perm)
		}

		if !rolePermissionMap[perm] {
			return fmt.Errorf("permission %s not allowed for role %s", perm, role)
		}
	}

	return nil
}

// GetResourceFromPath extracts resource type from API path
func GetResourceFromPath(path string) string {
	// Extract resource type from API path for resource-specific permissions
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	if len(pathParts) >= 2 && pathParts[0] == "v1" {
		return pathParts[1]
	}

	return ""
}

// RequiresPermission maps API endpoints to required permissions
var EndpointPermissions = map[string]Permission{
	"POST /v1/classify":         PermissionClassifyBusiness,
	"POST /v1/classify/batch":   PermissionClassifyBatch,
	"GET /v1/classify":          PermissionViewClassification,
	"GET /v1/risk":              PermissionViewRisk,
	"POST /v1/risk/assess":      PermissionAssessRisk,
	"GET /v1/compliance":        PermissionViewCompliance,
	"POST /v1/compliance/check": PermissionCheckCompliance,
	"GET /v1/users":             PermissionViewUsers,
	"POST /v1/users":            PermissionCreateUser,
	"PUT /v1/users":             PermissionUpdateUser,
	"DELETE /v1/users":          PermissionDeleteUser,
	"GET /v1/api-keys":          PermissionViewAPIKeys,
	"POST /v1/api-keys":         PermissionCreateAPIKey,
	"DELETE /v1/api-keys":       PermissionRevokeAPIKey,
	"GET /v1/metrics":           PermissionViewMetrics,
	"GET /v1/audit":             PermissionViewAuditLogs,
}

// GetRequiredPermission returns the permission required for a given endpoint
func GetRequiredPermission(method, path string) (Permission, bool) {
	endpoint := method + " " + path
	permission, exists := EndpointPermissions[endpoint]
	return permission, exists
}

// CheckPermission verifies if a user has the required permission for an endpoint
func (r *RBACService) CheckPermission(ctx context.Context, userRole Role, method, path string) error {
	requiredPermission, exists := GetRequiredPermission(method, path)
	if !exists {
		// If no specific permission is required, allow access
		return nil
	}

	if !HasPermission(userRole, requiredPermission) {
		return fmt.Errorf("insufficient permissions: %s required for %s %s", requiredPermission, method, path)
	}

	return nil
}

// ToAPIKeyResponse converts an APIKey to APIKeyResponse (removing sensitive data)
func (ak *APIKey) ToAPIKeyResponse() *APIKeyResponse {
	return &APIKeyResponse{
		ID:          ak.ID,
		Name:        ak.Name,
		UserID:      ak.UserID,
		Role:        ak.Role,
		Permissions: ak.Permissions,
		ExpiresAt:   ak.ExpiresAt,
		LastUsedAt:  ak.LastUsedAt,
		IsActive:    ak.IsActive,
		CreatedAt:   ak.CreatedAt,
		UpdatedAt:   ak.UpdatedAt,
	}
}
