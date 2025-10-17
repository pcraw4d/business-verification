package security

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// RBACManager provides role-based access control functionality
type RBACManager struct {
	logger      *zap.Logger
	roles       map[string]*Role
	permissions map[string]*Permission
	policies    []*Policy
}

// Role represents a user role with associated permissions
type Role struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Permissions []string               `json:"permissions"`
	Inherits    []string               `json:"inherits"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	IsActive    bool                   `json:"is_active"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Permission represents a specific permission
type Permission struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Resource    string    `json:"resource"`
	Action      string    `json:"action"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
}

// Policy represents an access control policy
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Effect      string                 `json:"effect"`   // "allow" or "deny"
	Subjects    []string               `json:"subjects"` // roles or users
	Resources   []string               `json:"resources"`
	Actions     []string               `json:"actions"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	IsActive    bool                   `json:"is_active"`
}

// AccessRequest represents a request for access to a resource
type AccessRequest struct {
	Subject  string                 `json:"subject"` // user ID or role
	Resource string                 `json:"resource"`
	Action   string                 `json:"action"`
	Context  map[string]interface{} `json:"context,omitempty"`
}

// AccessDecision represents the result of an access control decision
type AccessDecision struct {
	Allowed   bool                   `json:"allowed"`
	Reason    string                 `json:"reason,omitempty"`
	Policy    *Policy                `json:"policy,omitempty"`
	Role      *Role                  `json:"role,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Context   map[string]interface{} `json:"context,omitempty"`
}

// UserRole represents a user's role assignment
type UserRole struct {
	UserID     string     `json:"user_id"`
	RoleID     string     `json:"role_id"`
	AssignedAt time.Time  `json:"assigned_at"`
	AssignedBy string     `json:"assigned_by"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	IsActive   bool       `json:"is_active"`
}

// NewRBACManager creates a new RBAC manager with default roles and permissions
func NewRBACManager(logger *zap.Logger) *RBACManager {
	rbac := &RBACManager{
		logger:      logger,
		roles:       make(map[string]*Role),
		permissions: make(map[string]*Permission),
		policies:    make([]*Policy, 0),
	}

	// Initialize default permissions
	rbac.initializeDefaultPermissions()

	// Initialize default roles
	rbac.initializeDefaultRoles()

	// Initialize default policies
	rbac.initializeDefaultPolicies()

	return rbac
}

// initializeDefaultPermissions creates default permissions for the risk assessment service
func (rbac *RBACManager) initializeDefaultPermissions() {
	now := time.Now()

	permissions := []*Permission{
		// Risk Assessment permissions
		{
			ID:          "risk_assessment:create",
			Name:        "Create Risk Assessment",
			Description: "Create new risk assessments",
			Resource:    "risk_assessment",
			Action:      "create",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "risk_assessment:read",
			Name:        "Read Risk Assessment",
			Description: "View risk assessment data",
			Resource:    "risk_assessment",
			Action:      "read",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "risk_assessment:update",
			Name:        "Update Risk Assessment",
			Description: "Modify existing risk assessments",
			Resource:    "risk_assessment",
			Action:      "update",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "risk_assessment:delete",
			Name:        "Delete Risk Assessment",
			Description: "Delete risk assessments",
			Resource:    "risk_assessment",
			Action:      "delete",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},

		// ML Model permissions
		{
			ID:          "ml_model:train",
			Name:        "Train ML Models",
			Description: "Train machine learning models",
			Resource:    "ml_model",
			Action:      "train",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "ml_model:predict",
			Name:        "ML Model Prediction",
			Description: "Use ML models for predictions",
			Resource:    "ml_model",
			Action:      "predict",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "ml_model:validate",
			Name:        "Validate ML Models",
			Description: "Validate ML model performance",
			Resource:    "ml_model",
			Action:      "validate",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},

		// External API permissions
		{
			ID:          "external_api:access",
			Name:        "Access External APIs",
			Description: "Access external data sources",
			Resource:    "external_api",
			Action:      "access",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},

		// Admin permissions
		{
			ID:          "admin:manage_users",
			Name:        "Manage Users",
			Description: "Create, update, and delete users",
			Resource:    "admin",
			Action:      "manage_users",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "admin:manage_roles",
			Name:        "Manage Roles",
			Description: "Create, update, and delete roles",
			Resource:    "admin",
			Action:      "manage_roles",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "admin:view_audit_logs",
			Name:        "View Audit Logs",
			Description: "View system audit logs",
			Resource:    "admin",
			Action:      "view_audit_logs",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "admin:system_config",
			Name:        "System Configuration",
			Description: "Modify system configuration",
			Resource:    "admin",
			Action:      "system_config",
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
	}

	for _, perm := range permissions {
		rbac.permissions[perm.ID] = perm
	}
}

// initializeDefaultRoles creates default roles for the risk assessment service
func (rbac *RBACManager) initializeDefaultRoles() {
	now := time.Now()

	roles := []*Role{
		{
			ID:          "admin",
			Name:        "Administrator",
			Description: "Full system access",
			Permissions: []string{
				"risk_assessment:create",
				"risk_assessment:read",
				"risk_assessment:update",
				"risk_assessment:delete",
				"ml_model:train",
				"ml_model:predict",
				"ml_model:validate",
				"external_api:access",
				"admin:manage_users",
				"admin:manage_roles",
				"admin:view_audit_logs",
				"admin:system_config",
			},
			Inherits:  []string{},
			CreatedAt: now,
			UpdatedAt: now,
			IsActive:  true,
		},
		{
			ID:          "analyst",
			Name:        "Risk Analyst",
			Description: "Risk assessment and analysis",
			Permissions: []string{
				"risk_assessment:create",
				"risk_assessment:read",
				"risk_assessment:update",
				"ml_model:predict",
				"ml_model:validate",
				"external_api:access",
			},
			Inherits:  []string{},
			CreatedAt: now,
			UpdatedAt: now,
			IsActive:  true,
		},
		{
			ID:          "data_scientist",
			Name:        "Data Scientist",
			Description: "ML model development and training",
			Permissions: []string{
				"risk_assessment:read",
				"ml_model:train",
				"ml_model:predict",
				"ml_model:validate",
				"external_api:access",
			},
			Inherits:  []string{},
			CreatedAt: now,
			UpdatedAt: now,
			IsActive:  true,
		},
		{
			ID:          "viewer",
			Name:        "Viewer",
			Description: "Read-only access to risk assessments",
			Permissions: []string{
				"risk_assessment:read",
				"ml_model:predict",
			},
			Inherits:  []string{},
			CreatedAt: now,
			UpdatedAt: now,
			IsActive:  true,
		},
		{
			ID:          "api_user",
			Name:        "API User",
			Description: "API access for risk assessments",
			Permissions: []string{
				"risk_assessment:create",
				"risk_assessment:read",
				"ml_model:predict",
			},
			Inherits:  []string{},
			CreatedAt: now,
			UpdatedAt: now,
			IsActive:  true,
		},
	}

	for _, role := range roles {
		rbac.roles[role.ID] = role
	}
}

// initializeDefaultPolicies creates default access control policies
func (rbac *RBACManager) initializeDefaultPolicies() {
	now := time.Now()

	policies := []*Policy{
		{
			ID:          "admin_full_access",
			Name:        "Admin Full Access",
			Description: "Administrators have full access to all resources",
			Effect:      "allow",
			Subjects:    []string{"admin"},
			Resources:   []string{"*"},
			Actions:     []string{"*"},
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "analyst_risk_access",
			Name:        "Analyst Risk Assessment Access",
			Description: "Analysts can manage risk assessments",
			Effect:      "allow",
			Subjects:    []string{"analyst"},
			Resources:   []string{"risk_assessment", "ml_model", "external_api"},
			Actions:     []string{"create", "read", "update", "predict", "validate", "access"},
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "data_scientist_ml_access",
			Name:        "Data Scientist ML Access",
			Description: "Data scientists can work with ML models",
			Effect:      "allow",
			Subjects:    []string{"data_scientist"},
			Resources:   []string{"risk_assessment", "ml_model", "external_api"},
			Actions:     []string{"read", "train", "predict", "validate", "access"},
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "viewer_read_only",
			Name:        "Viewer Read Only",
			Description: "Viewers have read-only access",
			Effect:      "allow",
			Subjects:    []string{"viewer"},
			Resources:   []string{"risk_assessment", "ml_model"},
			Actions:     []string{"read", "predict"},
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "api_user_limited",
			Name:        "API User Limited Access",
			Description: "API users have limited access",
			Effect:      "allow",
			Subjects:    []string{"api_user"},
			Resources:   []string{"risk_assessment", "ml_model"},
			Actions:     []string{"create", "read", "predict"},
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
		{
			ID:          "deny_admin_actions",
			Name:        "Deny Admin Actions for Non-Admins",
			Description: "Deny admin actions for non-admin users",
			Effect:      "deny",
			Subjects:    []string{"analyst", "data_scientist", "viewer", "api_user"},
			Resources:   []string{"admin"},
			Actions:     []string{"*"},
			CreatedAt:   now,
			UpdatedAt:   now,
			IsActive:    true,
		},
	}

	rbac.policies = policies
}

// CheckAccess checks if a subject has access to perform an action on a resource
func (rbac *RBACManager) CheckAccess(ctx context.Context, req *AccessRequest, userRoles []string) *AccessDecision {
	decision := &AccessDecision{
		Allowed:   false,
		Timestamp: time.Now(),
		Context:   req.Context,
	}

	// Check deny policies first
	for _, policy := range rbac.policies {
		if !policy.IsActive || policy.Effect != "deny" {
			continue
		}

		// Check if deny policy applies to this request
		if rbac.policyApplies(policy, req, userRoles) {
			decision.Policy = policy
			decision.Allowed = false
			decision.Reason = fmt.Sprintf("Denied by policy: %s", policy.Name)
			return decision
		}
	}

	// Check allow policies
	for _, policy := range rbac.policies {
		if !policy.IsActive || policy.Effect != "allow" {
			continue
		}

		// Check if allow policy applies to this request
		if rbac.policyApplies(policy, req, userRoles) {
			decision.Policy = policy
			decision.Allowed = true
			decision.Reason = fmt.Sprintf("Allowed by policy: %s", policy.Name)
			return decision
		}
	}

	// If no policy matched, check role-based permissions
	if !decision.Allowed && decision.Policy == nil {
		decision = rbac.checkRolePermissions(req, userRoles)
	}

	rbac.logger.Info("Access check completed",
		zap.String("subject", req.Subject),
		zap.String("resource", req.Resource),
		zap.String("action", req.Action),
		zap.Bool("allowed", decision.Allowed),
		zap.String("reason", decision.Reason))

	return decision
}

// policyApplies checks if a policy applies to the given request
func (rbac *RBACManager) policyApplies(policy *Policy, req *AccessRequest, userRoles []string) bool {
	// Check subjects (roles)
	subjectMatch := false
	for _, subject := range policy.Subjects {
		if subject == "*" {
			subjectMatch = true
			break
		}
		for _, userRole := range userRoles {
			if subject == userRole {
				subjectMatch = true
				break
			}
		}
	}
	if !subjectMatch {
		return false
	}

	// Check resources
	resourceMatch := false
	for _, resource := range policy.Resources {
		if resource == "*" || resource == req.Resource {
			resourceMatch = true
			break
		}
	}
	if !resourceMatch {
		return false
	}

	// Check actions
	actionMatch := false
	for _, action := range policy.Actions {
		if action == "*" || action == req.Action {
			actionMatch = true
			break
		}
	}
	if !actionMatch {
		return false
	}

	// Check conditions if any
	if len(policy.Conditions) > 0 {
		return rbac.evaluateConditions(policy.Conditions, req.Context)
	}

	return true
}

// checkRolePermissions checks if user roles have the required permissions
func (rbac *RBACManager) checkRolePermissions(req *AccessRequest, userRoles []string) *AccessDecision {
	decision := &AccessDecision{
		Allowed:   false,
		Timestamp: time.Now(),
		Context:   req.Context,
	}

	// Check each user role
	for _, roleID := range userRoles {
		role, exists := rbac.roles[roleID]
		if !exists || !role.IsActive {
			continue
		}

		// Check if role has the required permission
		permissionID := fmt.Sprintf("%s:%s", req.Resource, req.Action)
		for _, permID := range role.Permissions {
			if permID == permissionID {
				decision.Allowed = true
				decision.Role = role
				decision.Reason = fmt.Sprintf("Allowed by role: %s", role.Name)
				return decision
			}
		}

		// Check inherited roles
		for _, inheritedRoleID := range role.Inherits {
			inheritedRole, exists := rbac.roles[inheritedRoleID]
			if !exists || !inheritedRole.IsActive {
				continue
			}

			for _, permID := range inheritedRole.Permissions {
				if permID == permissionID {
					decision.Allowed = true
					decision.Role = role
					decision.Reason = fmt.Sprintf("Allowed by inherited role: %s", inheritedRole.Name)
					return decision
				}
			}
		}
	}

	decision.Reason = "No matching permissions found"
	return decision
}

// evaluateConditions evaluates policy conditions
func (rbac *RBACManager) evaluateConditions(conditions map[string]interface{}, context map[string]interface{}) bool {
	for key, expectedValue := range conditions {
		actualValue, exists := context[key]
		if !exists {
			return false
		}

		// Simple equality check
		if actualValue != expectedValue {
			return false
		}
	}
	return true
}

// CreateRole creates a new role
func (rbac *RBACManager) CreateRole(role *Role) error {
	if role.ID == "" {
		return fmt.Errorf("role ID is required")
	}

	if _, exists := rbac.roles[role.ID]; exists {
		return fmt.Errorf("role with ID %s already exists", role.ID)
	}

	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now
	role.IsActive = true

	rbac.roles[role.ID] = role

	rbac.logger.Info("Role created",
		zap.String("role_id", role.ID),
		zap.String("role_name", role.Name))

	return nil
}

// UpdateRole updates an existing role
func (rbac *RBACManager) UpdateRole(roleID string, updates *Role) error {
	role, exists := rbac.roles[roleID]
	if !exists {
		return fmt.Errorf("role with ID %s not found", roleID)
	}

	// Update fields
	if updates.Name != "" {
		role.Name = updates.Name
	}
	if updates.Description != "" {
		role.Description = updates.Description
	}
	if len(updates.Permissions) > 0 {
		role.Permissions = updates.Permissions
	}
	if len(updates.Inherits) > 0 {
		role.Inherits = updates.Inherits
	}
	if updates.Metadata != nil {
		role.Metadata = updates.Metadata
	}

	role.UpdatedAt = time.Now()

	rbac.logger.Info("Role updated",
		zap.String("role_id", roleID),
		zap.String("role_name", role.Name))

	return nil
}

// DeleteRole deletes a role
func (rbac *RBACManager) DeleteRole(roleID string) error {
	role, exists := rbac.roles[roleID]
	if !exists {
		return fmt.Errorf("role with ID %s not found", roleID)
	}

	// Check if role is in use
	// In a real implementation, you'd check user assignments

	delete(rbac.roles, roleID)

	rbac.logger.Info("Role deleted",
		zap.String("role_id", roleID),
		zap.String("role_name", role.Name))

	return nil
}

// GetRole retrieves a role by ID
func (rbac *RBACManager) GetRole(roleID string) (*Role, error) {
	role, exists := rbac.roles[roleID]
	if !exists {
		return nil, fmt.Errorf("role with ID %s not found", roleID)
	}
	return role, nil
}

// ListRoles returns all roles
func (rbac *RBACManager) ListRoles() []*Role {
	roles := make([]*Role, 0, len(rbac.roles))
	for _, role := range rbac.roles {
		roles = append(roles, role)
	}
	return roles
}

// CreatePermission creates a new permission
func (rbac *RBACManager) CreatePermission(permission *Permission) error {
	if permission.ID == "" {
		return fmt.Errorf("permission ID is required")
	}

	if _, exists := rbac.permissions[permission.ID]; exists {
		return fmt.Errorf("permission with ID %s already exists", permission.ID)
	}

	now := time.Now()
	permission.CreatedAt = now
	permission.UpdatedAt = now
	permission.IsActive = true

	rbac.permissions[permission.ID] = permission

	rbac.logger.Info("Permission created",
		zap.String("permission_id", permission.ID),
		zap.String("permission_name", permission.Name))

	return nil
}

// GetPermission retrieves a permission by ID
func (rbac *RBACManager) GetPermission(permissionID string) (*Permission, error) {
	permission, exists := rbac.permissions[permissionID]
	if !exists {
		return nil, fmt.Errorf("permission with ID %s not found", permissionID)
	}
	return permission, nil
}

// ListPermissions returns all permissions
func (rbac *RBACManager) ListPermissions() []*Permission {
	permissions := make([]*Permission, 0, len(rbac.permissions))
	for _, permission := range rbac.permissions {
		permissions = append(permissions, permission)
	}
	return permissions
}

// CreatePolicy creates a new access control policy
func (rbac *RBACManager) CreatePolicy(policy *Policy) error {
	if policy.ID == "" {
		return fmt.Errorf("policy ID is required")
	}

	// Check if policy already exists
	for _, existingPolicy := range rbac.policies {
		if existingPolicy.ID == policy.ID {
			return fmt.Errorf("policy with ID %s already exists", policy.ID)
		}
	}

	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now
	policy.IsActive = true

	rbac.policies = append(rbac.policies, policy)

	rbac.logger.Info("Policy created",
		zap.String("policy_id", policy.ID),
		zap.String("policy_name", policy.Name))

	return nil
}

// GetPolicy retrieves a policy by ID
func (rbac *RBACManager) GetPolicy(policyID string) (*Policy, error) {
	for _, policy := range rbac.policies {
		if policy.ID == policyID {
			return policy, nil
		}
	}
	return nil, fmt.Errorf("policy with ID %s not found", policyID)
}

// ListPolicies returns all policies
func (rbac *RBACManager) ListPolicies() []*Policy {
	return rbac.policies
}

// GetUserPermissions returns all permissions for a user based on their roles
func (rbac *RBACManager) GetUserPermissions(userRoles []string) []string {
	permissions := make(map[string]bool)

	for _, roleID := range userRoles {
		role, exists := rbac.roles[roleID]
		if !exists || !role.IsActive {
			continue
		}

		// Add role permissions
		for _, permID := range role.Permissions {
			permissions[permID] = true
		}

		// Add inherited role permissions
		for _, inheritedRoleID := range role.Inherits {
			inheritedRole, exists := rbac.roles[inheritedRoleID]
			if !exists || !inheritedRole.IsActive {
				continue
			}

			for _, permID := range inheritedRole.Permissions {
				permissions[permID] = true
			}
		}
	}

	// Convert map to slice
	permList := make([]string, 0, len(permissions))
	for permID := range permissions {
		permList = append(permList, permID)
	}

	return permList
}

// ValidatePermission checks if a permission exists and is valid
func (rbac *RBACManager) ValidatePermission(permissionID string) bool {
	permission, exists := rbac.permissions[permissionID]
	return exists && permission.IsActive
}

// GetRoleHierarchy returns the role hierarchy
func (rbac *RBACManager) GetRoleHierarchy() map[string][]string {
	hierarchy := make(map[string][]string)

	for roleID, role := range rbac.roles {
		if role.IsActive {
			hierarchy[roleID] = role.Inherits
		}
	}

	return hierarchy
}

// CheckPermission checks if a role has a specific permission
func (rbac *RBACManager) CheckPermission(roleID, permissionID string) bool {
	role, exists := rbac.roles[roleID]
	if !exists || !role.IsActive {
		return false
	}

	// Check direct permissions
	for _, permID := range role.Permissions {
		if permID == permissionID {
			return true
		}
	}

	// Check inherited permissions
	for _, inheritedRoleID := range role.Inherits {
		if rbac.CheckPermission(inheritedRoleID, permissionID) {
			return true
		}
	}

	return false
}
