package security

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRBACManager_CheckAccess(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	tests := []struct {
		name      string
		request   *AccessRequest
		userRoles []string
		expected  bool
	}{
		{
			name: "admin full access",
			request: &AccessRequest{
				Subject:  "user-123",
				Resource: "risk_assessment",
				Action:   "create",
			},
			userRoles: []string{"admin"},
			expected:  true,
		},
		{
			name: "analyst can create risk assessment",
			request: &AccessRequest{
				Subject:  "user-456",
				Resource: "risk_assessment",
				Action:   "create",
			},
			userRoles: []string{"analyst"},
			expected:  true,
		},
		{
			name: "analyst cannot delete risk assessment",
			request: &AccessRequest{
				Subject:  "user-456",
				Resource: "risk_assessment",
				Action:   "delete",
			},
			userRoles: []string{"analyst"},
			expected:  false,
		},
		{
			name: "viewer can only read",
			request: &AccessRequest{
				Subject:  "user-789",
				Resource: "risk_assessment",
				Action:   "read",
			},
			userRoles: []string{"viewer"},
			expected:  true,
		},
		{
			name: "viewer cannot create",
			request: &AccessRequest{
				Subject:  "user-789",
				Resource: "risk_assessment",
				Action:   "create",
			},
			userRoles: []string{"viewer"},
			expected:  false,
		},
		{
			name: "data scientist can train models",
			request: &AccessRequest{
				Subject:  "user-101",
				Resource: "ml_model",
				Action:   "train",
			},
			userRoles: []string{"data_scientist"},
			expected:  true,
		},
		{
			name: "data scientist cannot manage users",
			request: &AccessRequest{
				Subject:  "user-101",
				Resource: "admin",
				Action:   "manage_users",
			},
			userRoles: []string{"data_scientist"},
			expected:  false,
		},
		{
			name: "api user can create and read",
			request: &AccessRequest{
				Subject:  "api-user-1",
				Resource: "risk_assessment",
				Action:   "create",
			},
			userRoles: []string{"api_user"},
			expected:  true,
		},
		{
			name: "api user cannot update",
			request: &AccessRequest{
				Subject:  "api-user-1",
				Resource: "risk_assessment",
				Action:   "update",
			},
			userRoles: []string{"api_user"},
			expected:  false,
		},
		{
			name: "non-existent role",
			request: &AccessRequest{
				Subject:  "user-999",
				Resource: "risk_assessment",
				Action:   "read",
			},
			userRoles: []string{"non_existent_role"},
			expected:  false,
		},
		{
			name: "multiple roles - admin overrides",
			request: &AccessRequest{
				Subject:  "user-111",
				Resource: "admin",
				Action:   "manage_users",
			},
			userRoles: []string{"viewer", "admin"},
			expected:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decision := rbac.CheckAccess(context.Background(), tt.request, tt.userRoles)
			assert.Equal(t, tt.expected, decision.Allowed, "Access decision should match expected result")
		})
	}
}

func TestRBACManager_CreateRole(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	newRole := &Role{
		ID:          "custom_role",
		Name:        "Custom Role",
		Description: "A custom role for testing",
		Permissions: []string{"risk_assessment:read", "ml_model:predict"},
		Inherits:    []string{"viewer"},
	}

	err := rbac.CreateRole(newRole)
	assert.NoError(t, err, "Creating role should not return error")

	// Verify role was created
	role, err := rbac.GetRole("custom_role")
	assert.NoError(t, err, "Getting created role should not return error")
	assert.Equal(t, "Custom Role", role.Name)
	assert.Equal(t, "A custom role for testing", role.Description)
	assert.True(t, role.IsActive)
	assert.NotZero(t, role.CreatedAt)
}

func TestRBACManager_CreateRole_Duplicate(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	newRole := &Role{
		ID:          "admin", // This role already exists
		Name:        "Duplicate Admin",
		Description: "This should fail",
	}

	err := rbac.CreateRole(newRole)
	assert.Error(t, err, "Creating duplicate role should return error")
	assert.Contains(t, err.Error(), "already exists")
}

func TestRBACManager_CreateRole_EmptyID(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	newRole := &Role{
		ID:          "", // Empty ID
		Name:        "No ID Role",
		Description: "This should fail",
	}

	err := rbac.CreateRole(newRole)
	assert.Error(t, err, "Creating role with empty ID should return error")
	assert.Contains(t, err.Error(), "role ID is required")
}

func TestRBACManager_UpdateRole(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	// Create a custom role first
	newRole := &Role{
		ID:          "test_role",
		Name:        "Test Role",
		Description: "Original description",
		Permissions: []string{"risk_assessment:read"},
	}

	err := rbac.CreateRole(newRole)
	assert.NoError(t, err, "Creating role should not return error")

	// Update the role
	updates := &Role{
		Name:        "Updated Test Role",
		Description: "Updated description",
		Permissions: []string{"risk_assessment:read", "risk_assessment:create"},
	}

	err = rbac.UpdateRole("test_role", updates)
	assert.NoError(t, err, "Updating role should not return error")

	// Verify updates
	role, err := rbac.GetRole("test_role")
	assert.NoError(t, err, "Getting updated role should not return error")
	assert.Equal(t, "Updated Test Role", role.Name)
	assert.Equal(t, "Updated description", role.Description)
	assert.Equal(t, []string{"risk_assessment:read", "risk_assessment:create"}, role.Permissions)
	assert.True(t, role.UpdatedAt.After(role.CreatedAt))
}

func TestRBACManager_UpdateRole_NotFound(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	updates := &Role{
		Name: "Non-existent Role",
	}

	err := rbac.UpdateRole("non_existent", updates)
	assert.Error(t, err, "Updating non-existent role should return error")
	assert.Contains(t, err.Error(), "not found")
}

func TestRBACManager_DeleteRole(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	// Create a custom role first
	newRole := &Role{
		ID:          "delete_test_role",
		Name:        "Delete Test Role",
		Description: "Role to be deleted",
		Permissions: []string{"risk_assessment:read"},
	}

	err := rbac.CreateRole(newRole)
	assert.NoError(t, err, "Creating role should not return error")

	// Delete the role
	err = rbac.DeleteRole("delete_test_role")
	assert.NoError(t, err, "Deleting role should not return error")

	// Verify role was deleted
	_, err = rbac.GetRole("delete_test_role")
	assert.Error(t, err, "Getting deleted role should return error")
	assert.Contains(t, err.Error(), "not found")
}

func TestRBACManager_DeleteRole_NotFound(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	err := rbac.DeleteRole("non_existent")
	assert.Error(t, err, "Deleting non-existent role should return error")
	assert.Contains(t, err.Error(), "not found")
}

func TestRBACManager_GetRole(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	// Test getting existing role
	role, err := rbac.GetRole("admin")
	assert.NoError(t, err, "Getting existing role should not return error")
	assert.Equal(t, "admin", role.ID)
	assert.Equal(t, "Administrator", role.Name)
	assert.True(t, role.IsActive)

	// Test getting non-existent role
	_, err = rbac.GetRole("non_existent")
	assert.Error(t, err, "Getting non-existent role should return error")
	assert.Contains(t, err.Error(), "not found")
}

func TestRBACManager_ListRoles(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	roles := rbac.ListRoles()
	assert.NotEmpty(t, roles, "Should return default roles")

	// Check that default roles are present
	roleIDs := make(map[string]bool)
	for _, role := range roles {
		roleIDs[role.ID] = true
	}

	assert.True(t, roleIDs["admin"], "Should include admin role")
	assert.True(t, roleIDs["analyst"], "Should include analyst role")
	assert.True(t, roleIDs["data_scientist"], "Should include data_scientist role")
	assert.True(t, roleIDs["viewer"], "Should include viewer role")
	assert.True(t, roleIDs["api_user"], "Should include api_user role")
}

func TestRBACManager_CreatePermission(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	newPermission := &Permission{
		ID:          "custom:permission",
		Name:        "Custom Permission",
		Description: "A custom permission for testing",
		Resource:    "custom_resource",
		Action:      "custom_action",
	}

	err := rbac.CreatePermission(newPermission)
	assert.NoError(t, err, "Creating permission should not return error")

	// Verify permission was created
	permission, err := rbac.GetPermission("custom:permission")
	assert.NoError(t, err, "Getting created permission should not return error")
	assert.Equal(t, "Custom Permission", permission.Name)
	assert.Equal(t, "custom_resource", permission.Resource)
	assert.Equal(t, "custom_action", permission.Action)
	assert.True(t, permission.IsActive)
}

func TestRBACManager_CreatePermission_Duplicate(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	newPermission := &Permission{
		ID:          "risk_assessment:create", // This permission already exists
		Name:        "Duplicate Permission",
		Description: "This should fail",
		Resource:    "risk_assessment",
		Action:      "create",
	}

	err := rbac.CreatePermission(newPermission)
	assert.Error(t, err, "Creating duplicate permission should return error")
	assert.Contains(t, err.Error(), "already exists")
}

func TestRBACManager_GetPermission(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	// Test getting existing permission
	permission, err := rbac.GetPermission("risk_assessment:create")
	assert.NoError(t, err, "Getting existing permission should not return error")
	assert.Equal(t, "risk_assessment:create", permission.ID)
	assert.Equal(t, "Create Risk Assessment", permission.Name)
	assert.True(t, permission.IsActive)

	// Test getting non-existent permission
	_, err = rbac.GetPermission("non_existent:permission")
	assert.Error(t, err, "Getting non-existent permission should return error")
	assert.Contains(t, err.Error(), "not found")
}

func TestRBACManager_ListPermissions(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	permissions := rbac.ListPermissions()
	assert.NotEmpty(t, permissions, "Should return default permissions")

	// Check that some default permissions are present
	permissionIDs := make(map[string]bool)
	for _, permission := range permissions {
		permissionIDs[permission.ID] = true
	}

	assert.True(t, permissionIDs["risk_assessment:create"], "Should include risk_assessment:create")
	assert.True(t, permissionIDs["ml_model:train"], "Should include ml_model:train")
	assert.True(t, permissionIDs["admin:manage_users"], "Should include admin:manage_users")
}

func TestRBACManager_CreatePolicy(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	newPolicy := &Policy{
		ID:          "custom_policy",
		Name:        "Custom Policy",
		Description: "A custom policy for testing",
		Effect:      "allow",
		Subjects:    []string{"custom_role"},
		Resources:   []string{"custom_resource"},
		Actions:     []string{"custom_action"},
	}

	err := rbac.CreatePolicy(newPolicy)
	assert.NoError(t, err, "Creating policy should not return error")

	// Verify policy was created
	policy, err := rbac.GetPolicy("custom_policy")
	assert.NoError(t, err, "Getting created policy should not return error")
	assert.Equal(t, "Custom Policy", policy.Name)
	assert.Equal(t, "allow", policy.Effect)
	assert.True(t, policy.IsActive)
}

func TestRBACManager_GetPolicy(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	// Test getting existing policy
	policy, err := rbac.GetPolicy("admin_full_access")
	assert.NoError(t, err, "Getting existing policy should not return error")
	assert.Equal(t, "admin_full_access", policy.ID)
	assert.Equal(t, "Admin Full Access", policy.Name)
	assert.Equal(t, "allow", policy.Effect)

	// Test getting non-existent policy
	_, err = rbac.GetPolicy("non_existent_policy")
	assert.Error(t, err, "Getting non-existent policy should return error")
	assert.Contains(t, err.Error(), "not found")
}

func TestRBACManager_ListPolicies(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	policies := rbac.ListPolicies()
	assert.NotEmpty(t, policies, "Should return default policies")

	// Check that some default policies are present
	policyIDs := make(map[string]bool)
	for _, policy := range policies {
		policyIDs[policy.ID] = true
	}

	assert.True(t, policyIDs["admin_full_access"], "Should include admin_full_access policy")
	assert.True(t, policyIDs["analyst_risk_access"], "Should include analyst_risk_access policy")
	assert.True(t, policyIDs["deny_admin_actions"], "Should include deny_admin_actions policy")
}

func TestRBACManager_GetUserPermissions(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	// Test admin permissions
	adminPermissions := rbac.GetUserPermissions([]string{"admin"})
	assert.NotEmpty(t, adminPermissions, "Admin should have permissions")

	// Check for specific admin permissions
	permissionMap := make(map[string]bool)
	for _, perm := range adminPermissions {
		permissionMap[perm] = true
	}

	assert.True(t, permissionMap["risk_assessment:create"], "Admin should have risk_assessment:create")
	assert.True(t, permissionMap["admin:manage_users"], "Admin should have admin:manage_users")
	assert.True(t, permissionMap["ml_model:train"], "Admin should have ml_model:train")

	// Test viewer permissions
	viewerPermissions := rbac.GetUserPermissions([]string{"viewer"})
	assert.NotEmpty(t, viewerPermissions, "Viewer should have permissions")

	viewerPermissionMap := make(map[string]bool)
	for _, perm := range viewerPermissions {
		viewerPermissionMap[perm] = true
	}

	assert.True(t, viewerPermissionMap["risk_assessment:read"], "Viewer should have risk_assessment:read")
	assert.False(t, viewerPermissionMap["risk_assessment:create"], "Viewer should not have risk_assessment:create")
	assert.False(t, viewerPermissionMap["admin:manage_users"], "Viewer should not have admin:manage_users")

	// Test multiple roles
	multiRolePermissions := rbac.GetUserPermissions([]string{"viewer", "analyst"})
	assert.NotEmpty(t, multiRolePermissions, "Multiple roles should have permissions")

	multiRolePermissionMap := make(map[string]bool)
	for _, perm := range multiRolePermissions {
		multiRolePermissionMap[perm] = true
	}

	assert.True(t, multiRolePermissionMap["risk_assessment:read"], "Should have viewer permissions")
	assert.True(t, multiRolePermissionMap["risk_assessment:create"], "Should have analyst permissions")
}

func TestRBACManager_ValidatePermission(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	// Test valid permission
	valid := rbac.ValidatePermission("risk_assessment:create")
	assert.True(t, valid, "Valid permission should return true")

	// Test invalid permission
	invalid := rbac.ValidatePermission("non_existent:permission")
	assert.False(t, invalid, "Invalid permission should return false")
}

func TestRBACManager_GetRoleHierarchy(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	hierarchy := rbac.GetRoleHierarchy()
	assert.NotEmpty(t, hierarchy, "Should return role hierarchy")

	// Check that default roles are in hierarchy
	assert.Contains(t, hierarchy, "admin", "Should include admin in hierarchy")
	assert.Contains(t, hierarchy, "analyst", "Should include analyst in hierarchy")
	assert.Contains(t, hierarchy, "viewer", "Should include viewer in hierarchy")
}

func TestRBACManager_CheckPermission(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	// Test admin permission
	hasPermission := rbac.CheckPermission("admin", "risk_assessment:create")
	assert.True(t, hasPermission, "Admin should have risk_assessment:create permission")

	// Test viewer permission
	hasPermission = rbac.CheckPermission("viewer", "risk_assessment:read")
	assert.True(t, hasPermission, "Viewer should have risk_assessment:read permission")

	// Test viewer without permission
	hasPermission = rbac.CheckPermission("viewer", "risk_assessment:create")
	assert.False(t, hasPermission, "Viewer should not have risk_assessment:create permission")

	// Test non-existent role
	hasPermission = rbac.CheckPermission("non_existent", "risk_assessment:read")
	assert.False(t, hasPermission, "Non-existent role should not have any permissions")
}

func TestRBACManager_AccessRequestWithContext(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	// Create a policy with conditions
	conditionalPolicy := &Policy{
		ID:          "conditional_policy",
		Name:        "Conditional Policy",
		Description: "Policy with conditions",
		Effect:      "allow",
		Subjects:    []string{"analyst"},
		Resources:   []string{"risk_assessment"},
		Actions:     []string{"create"},
		Conditions: map[string]interface{}{
			"time_of_day": "business_hours",
			"location":    "office",
		},
	}

	err := rbac.CreatePolicy(conditionalPolicy)
	assert.NoError(t, err, "Creating conditional policy should not return error")

	// Test with matching context
	request := &AccessRequest{
		Subject:  "user-123",
		Resource: "risk_assessment",
		Action:   "create",
		Context: map[string]interface{}{
			"time_of_day": "business_hours",
			"location":    "office",
		},
	}

	decision := rbac.CheckAccess(context.Background(), request, []string{"analyst"})
	assert.True(t, decision.Allowed, "Should allow access with matching context")

	// Test with non-matching context
	request.Context = map[string]interface{}{
		"time_of_day": "after_hours",
		"location":    "office",
	}

	decision = rbac.CheckAccess(context.Background(), request, []string{"analyst"})
	assert.False(t, decision.Allowed, "Should deny access with non-matching context")
}

func TestRBACManager_DefaultRolesAndPermissions(t *testing.T) {
	logger := zap.NewNop()
	rbac := NewRBACManager(logger)

	// Verify default roles exist
	roles := rbac.ListRoles()
	roleMap := make(map[string]*Role)
	for _, role := range roles {
		roleMap[role.ID] = role
	}

	// Check admin role
	adminRole, exists := roleMap["admin"]
	assert.True(t, exists, "Admin role should exist")
	assert.Equal(t, "Administrator", adminRole.Name)
	assert.Contains(t, adminRole.Permissions, "admin:manage_users", "Admin should have manage_users permission")

	// Check analyst role
	analystRole, exists := roleMap["analyst"]
	assert.True(t, exists, "Analyst role should exist")
	assert.Equal(t, "Risk Analyst", analystRole.Name)
	assert.Contains(t, analystRole.Permissions, "risk_assessment:create", "Analyst should have create permission")
	assert.NotContains(t, analystRole.Permissions, "admin:manage_users", "Analyst should not have admin permissions")

	// Check viewer role
	viewerRole, exists := roleMap["viewer"]
	assert.True(t, exists, "Viewer role should exist")
	assert.Equal(t, "Viewer", viewerRole.Name)
	assert.Contains(t, viewerRole.Permissions, "risk_assessment:read", "Viewer should have read permission")
	assert.NotContains(t, viewerRole.Permissions, "risk_assessment:create", "Viewer should not have create permission")

	// Verify default permissions exist
	permissions := rbac.ListPermissions()
	permissionMap := make(map[string]*Permission)
	for _, permission := range permissions {
		permissionMap[permission.ID] = permission
	}

	// Check risk assessment permissions
	assert.True(t, permissionMap["risk_assessment:create"] != nil, "risk_assessment:create permission should exist")
	assert.True(t, permissionMap["risk_assessment:read"] != nil, "risk_assessment:read permission should exist")
	assert.True(t, permissionMap["risk_assessment:update"] != nil, "risk_assessment:update permission should exist")
	assert.True(t, permissionMap["risk_assessment:delete"] != nil, "risk_assessment:delete permission should exist")

	// Check ML model permissions
	assert.True(t, permissionMap["ml_model:train"] != nil, "ml_model:train permission should exist")
	assert.True(t, permissionMap["ml_model:predict"] != nil, "ml_model:predict permission should exist")
	assert.True(t, permissionMap["ml_model:validate"] != nil, "ml_model:validate permission should exist")

	// Check admin permissions
	assert.True(t, permissionMap["admin:manage_users"] != nil, "admin:manage_users permission should exist")
	assert.True(t, permissionMap["admin:manage_roles"] != nil, "admin:manage_roles permission should exist")
	assert.True(t, permissionMap["admin:view_audit_logs"] != nil, "admin:view_audit_logs permission should exist")
}
