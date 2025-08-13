package security

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
)

func TestNewAccessControlSystem(t *testing.T) {
	logger := createTestLogger()
	config := AccessControlConfig{
		DefaultRole:           "user",
		SessionTimeout:        30 * time.Minute,
		MaxLoginAttempts:      5,
		LockoutDuration:       15 * time.Minute,
		MFAEnabled:            true,
		AuditLoggingEnabled:   true,
		PolicyEnforcementMode: "strict",
		PasswordPolicy: PasswordPolicy{
			MinLength:           8,
			RequireUppercase:    true,
			RequireLowercase:    true,
			RequireNumbers:      true,
			RequireSpecialChars: true,
			MaxAge:              90,
			PreventReuse:        5,
		},
	}

	acs := NewAccessControlSystem(logger, config)

	if acs == nil {
		t.Fatal("Expected access control system to be created, got nil")
	}

	if acs.logger != logger {
		t.Error("Expected logger to be set correctly")
	}

	if acs.config.DefaultRole != "user" {
		t.Error("Expected default role to be set correctly")
	}

	// Check that default roles were created
	if _, exists := acs.roles["admin"]; !exists {
		t.Error("Expected admin role to be created")
	}

	if _, exists := acs.roles["user"]; !exists {
		t.Error("Expected user role to be created")
	}

	if _, exists := acs.roles["readonly"]; !exists {
		t.Error("Expected readonly role to be created")
	}

	// Check that default permissions were created
	if _, exists := acs.permissions["read:business"]; !exists {
		t.Error("Expected read:business permission to be created")
	}

	if _, exists := acs.permissions["write:business"]; !exists {
		t.Error("Expected write:business permission to be created")
	}
}

func TestCheckAccess(t *testing.T) {
	logger := createTestLogger()
	config := AccessControlConfig{
		DefaultRole:           "user",
		SessionTimeout:        30 * time.Minute,
		MaxLoginAttempts:      5,
		LockoutDuration:       15 * time.Minute,
		MFAEnabled:            true,
		AuditLoggingEnabled:   true,
		PolicyEnforcementMode: "strict",
	}

	acs := NewAccessControlSystem(logger, config)

	// Test access without any roles assigned
	decision, err := acs.CheckAccess(context.Background(), "user1", "business", "read", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if decision.Allowed {
		t.Error("Expected access to be denied for user without roles")
	}

	// Grant role to user
	err = acs.GrantRole(context.Background(), "user1", "user")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assign permission to role
	err = acs.AssignPermissionToRole(context.Background(), "user", "read:business")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Test access with role and permission
	decision, err = acs.CheckAccess(context.Background(), "user1", "business", "read", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !decision.Allowed {
		t.Error("Expected access to be allowed for user with role and permission")
	}

	if decision.RoleID != "user" {
		t.Error("Expected role ID to be set in decision")
	}

	if len(decision.Permissions) == 0 {
		t.Error("Expected permissions to be set in decision")
	}
}

func TestGrantAndRevokeRole(t *testing.T) {
	logger := createTestLogger()
	config := AccessControlConfig{
		DefaultRole:           "user",
		SessionTimeout:        30 * time.Minute,
		MaxLoginAttempts:      5,
		LockoutDuration:       15 * time.Minute,
		MFAEnabled:            true,
		AuditLoggingEnabled:   true,
		PolicyEnforcementMode: "strict",
	}

	acs := NewAccessControlSystem(logger, config)

	// Grant role
	err := acs.GrantRole(context.Background(), "user1", "admin")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that role was granted
	roles, err := acs.GetUserRoles(context.Background(), "user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(roles) != 1 {
		t.Fatalf("Expected 1 role, got %d", len(roles))
	}

	if roles[0].ID != "admin" {
		t.Error("Expected admin role to be granted")
	}

	// Try to grant the same role again
	err = acs.GrantRole(context.Background(), "user1", "admin")
	if err == nil {
		t.Error("Expected error when granting duplicate role")
	}

	// Revoke role
	err = acs.RevokeRole(context.Background(), "user1", "admin")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that role was revoked
	roles, err = acs.GetUserRoles(context.Background(), "user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(roles) != 0 {
		t.Fatalf("Expected 0 roles, got %d", len(roles))
	}

	// Try to revoke non-existent role
	err = acs.RevokeRole(context.Background(), "user1", "admin")
	if err == nil {
		t.Error("Expected error when revoking non-existent role")
	}
}

func TestCreateRole(t *testing.T) {
	logger := createTestLogger()
	config := AccessControlConfig{
		DefaultRole:           "user",
		SessionTimeout:        30 * time.Minute,
		MaxLoginAttempts:      5,
		LockoutDuration:       15 * time.Minute,
		MFAEnabled:            true,
		AuditLoggingEnabled:   true,
		PolicyEnforcementMode: "strict",
	}

	acs := NewAccessControlSystem(logger, config)

	role := &Role{
		Name:        "Test Role",
		Description: "A test role for testing",
		Permissions: []string{},
		IsActive:    true,
	}

	err := acs.CreateRole(context.Background(), role)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if role.ID == "" {
		t.Error("Expected role ID to be generated")
	}

	if role.CreatedAt.IsZero() {
		t.Error("Expected created timestamp to be set")
	}

	if role.UpdatedAt.IsZero() {
		t.Error("Expected updated timestamp to be set")
	}

	// Check that role was stored
	if _, exists := acs.roles[role.ID]; !exists {
		t.Error("Expected role to be stored")
	}
}

func TestCreatePermission(t *testing.T) {
	logger := createTestLogger()
	config := AccessControlConfig{
		DefaultRole:           "user",
		SessionTimeout:        30 * time.Minute,
		MaxLoginAttempts:      5,
		LockoutDuration:       15 * time.Minute,
		MFAEnabled:            true,
		AuditLoggingEnabled:   true,
		PolicyEnforcementMode: "strict",
	}

	acs := NewAccessControlSystem(logger, config)

	permission := &Permission{
		Name:        "Test Permission",
		Description: "A test permission for testing",
		Resource:    "test",
		Action:      "read",
		IsActive:    true,
	}

	err := acs.CreatePermission(context.Background(), permission)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if permission.ID == "" {
		t.Error("Expected permission ID to be generated")
	}

	if permission.CreatedAt.IsZero() {
		t.Error("Expected created timestamp to be set")
	}

	if permission.UpdatedAt.IsZero() {
		t.Error("Expected updated timestamp to be set")
	}

	// Check that permission was stored
	if _, exists := acs.permissions[permission.ID]; !exists {
		t.Error("Expected permission to be stored")
	}
}

func TestCreatePolicy(t *testing.T) {
	logger := createTestLogger()
	config := AccessControlConfig{
		DefaultRole:           "user",
		SessionTimeout:        30 * time.Minute,
		MaxLoginAttempts:      5,
		LockoutDuration:       15 * time.Minute,
		MFAEnabled:            true,
		AuditLoggingEnabled:   true,
		PolicyEnforcementMode: "strict",
	}

	acs := NewAccessControlSystem(logger, config)

	policy := &Policy{
		Name:        "Test Policy",
		Description: "A test policy for testing",
		Type:        PolicyTypeRBAC,
		Rules: []PolicyRule{
			{
				Name:     "Test Rule",
				Resource: "test",
				Action:   "read",
				Effect:   PolicyEffectAllow,
				Priority: 1,
			},
		},
		Effect:   PolicyEffectAllow,
		Priority: 1,
		IsActive: true,
	}

	err := acs.CreatePolicy(context.Background(), policy)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if policy.ID == "" {
		t.Error("Expected policy ID to be generated")
	}

	if policy.CreatedAt.IsZero() {
		t.Error("Expected created timestamp to be set")
	}

	if policy.UpdatedAt.IsZero() {
		t.Error("Expected updated timestamp to be set")
	}

	// Check that policy was stored
	if _, exists := acs.policies[policy.ID]; !exists {
		t.Error("Expected policy to be stored")
	}
}

func TestAssignAndRemovePermissionFromRole(t *testing.T) {
	logger := createTestLogger()
	config := AccessControlConfig{
		DefaultRole:           "user",
		SessionTimeout:        30 * time.Minute,
		MaxLoginAttempts:      5,
		LockoutDuration:       15 * time.Minute,
		MFAEnabled:            true,
		AuditLoggingEnabled:   true,
		PolicyEnforcementMode: "strict",
	}

	acs := NewAccessControlSystem(logger, config)

	// Create a custom role
	role := &Role{
		Name:        "Custom Role",
		Description: "A custom role for testing",
		Permissions: []string{},
		IsActive:    true,
	}

	err := acs.CreateRole(context.Background(), role)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Create a custom permission
	permission := &Permission{
		Name:        "Custom Permission",
		Description: "A custom permission for testing",
		Resource:    "custom",
		Action:      "write",
		IsActive:    true,
	}

	err = acs.CreatePermission(context.Background(), permission)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Assign permission to role
	err = acs.AssignPermissionToRole(context.Background(), role.ID, permission.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that permission was assigned
	permissions, err := acs.GetRolePermissions(context.Background(), role.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(permissions) != 1 {
		t.Fatalf("Expected 1 permission, got %d", len(permissions))
	}

	if permissions[0].ID != permission.ID {
		t.Error("Expected permission to be assigned to role")
	}

	// Try to assign the same permission again
	err = acs.AssignPermissionToRole(context.Background(), role.ID, permission.ID)
	if err == nil {
		t.Error("Expected error when assigning duplicate permission")
	}

	// Remove permission from role
	err = acs.RemovePermissionFromRole(context.Background(), role.ID, permission.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Check that permission was removed
	permissions, err = acs.GetRolePermissions(context.Background(), role.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(permissions) != 0 {
		t.Fatalf("Expected 0 permissions, got %d", len(permissions))
	}

	// Try to remove non-existent permission
	err = acs.RemovePermissionFromRole(context.Background(), role.ID, permission.ID)
	if err == nil {
		t.Error("Expected error when removing non-existent permission")
	}
}

func TestGetUserRoles(t *testing.T) {
	logger := createTestLogger()
	config := AccessControlConfig{
		DefaultRole:           "user",
		SessionTimeout:        30 * time.Minute,
		MaxLoginAttempts:      5,
		LockoutDuration:       15 * time.Minute,
		MFAEnabled:            true,
		AuditLoggingEnabled:   true,
		PolicyEnforcementMode: "strict",
	}

	acs := NewAccessControlSystem(logger, config)

	// Grant multiple roles to user
	err := acs.GrantRole(context.Background(), "user1", "admin")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = acs.GrantRole(context.Background(), "user1", "user")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Get user roles
	roles, err := acs.GetUserRoles(context.Background(), "user1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(roles) != 2 {
		t.Fatalf("Expected 2 roles, got %d", len(roles))
	}

	// Check that both roles are present
	roleIDs := make(map[string]bool)
	for _, role := range roles {
		roleIDs[role.ID] = true
	}

	if !roleIDs["admin"] {
		t.Error("Expected admin role to be present")
	}

	if !roleIDs["user"] {
		t.Error("Expected user role to be present")
	}

	// Test with non-existent user
	roles, err = acs.GetUserRoles(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(roles) != 0 {
		t.Fatalf("Expected 0 roles for non-existent user, got %d", len(roles))
	}
}

func TestGetRolePermissions(t *testing.T) {
	logger := createTestLogger()
	config := AccessControlConfig{
		DefaultRole:           "user",
		SessionTimeout:        30 * time.Minute,
		MaxLoginAttempts:      5,
		LockoutDuration:       15 * time.Minute,
		MFAEnabled:            true,
		AuditLoggingEnabled:   true,
		PolicyEnforcementMode: "strict",
	}

	acs := NewAccessControlSystem(logger, config)

	// Assign permissions to user role
	err := acs.AssignPermissionToRole(context.Background(), "user", "read:business")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = acs.AssignPermissionToRole(context.Background(), "user", "read:reports")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Get role permissions
	permissions, err := acs.GetRolePermissions(context.Background(), "user")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(permissions) != 2 {
		t.Fatalf("Expected 2 permissions, got %d", len(permissions))
	}

	// Check that both permissions are present
	permissionIDs := make(map[string]bool)
	for _, permission := range permissions {
		permissionIDs[permission.ID] = true
	}

	if !permissionIDs["read:business"] {
		t.Error("Expected read:business permission to be present")
	}

	if !permissionIDs["read:reports"] {
		t.Error("Expected read:reports permission to be present")
	}

	// Test with non-existent role
	permissions, err = acs.GetRolePermissions(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(permissions) != 0 {
		t.Fatalf("Expected 0 permissions for non-existent role, got %d", len(permissions))
	}
}

func TestAccessAuditLogger(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggerConfig{
		RetentionDays: 90,
		LogToFile:     true,
		LogToDatabase: false,
		LogLevel:      "info",
	}

	auditLogger := NewAccessAuditLogger(logger, config)

	if auditLogger == nil {
		t.Fatal("Expected audit logger to be created, got nil")
	}

	// Log an access event
	auditLogger.LogAccessEvent(context.Background(), "user1", "access_check", "business", true, map[string]interface{}{
		"ip_address": "192.168.1.1",
		"user_agent": "test-agent",
	})

	// Get audit logs
	logs, err := auditLogger.GetAuditLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 audit log, got %d", len(logs))
	}

	log := logs[0]
	if log.UserID != "user1" {
		t.Error("Expected user ID to match")
	}

	if log.Action != "access_check" {
		t.Error("Expected action to match")
	}

	if log.Resource != "business" {
		t.Error("Expected resource to match")
	}

	if log.Result != "success" {
		t.Error("Expected result to match")
	}

	if log.IPAddress != "192.168.1.1" {
		t.Error("Expected IP address to match")
	}

	if log.UserAgent != "test-agent" {
		t.Error("Expected user agent to match")
	}

	// Test filtering
	logs, err = auditLogger.GetAuditLogs(context.Background(), map[string]interface{}{
		"user_id": "user1",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(logs) != 1 {
		t.Fatalf("Expected 1 filtered audit log, got %d", len(logs))
	}

	// Test filtering with non-matching criteria
	logs, err = auditLogger.GetAuditLogs(context.Background(), map[string]interface{}{
		"user_id": "nonexistent",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(logs) != 0 {
		t.Fatalf("Expected 0 filtered audit logs, got %d", len(logs))
	}
}

func TestExportAuditLogs(t *testing.T) {
	logger := createTestLogger()
	config := AuditLoggerConfig{
		RetentionDays: 90,
		LogToFile:     true,
		LogToDatabase: false,
		LogLevel:      "info",
	}

	auditLogger := NewAccessAuditLogger(logger, config)

	// Log some access events
	auditLogger.LogAccessEvent(context.Background(), "user1", "access_check", "business", true, nil)
	auditLogger.LogAccessEvent(context.Background(), "user2", "role_granted", "admin", true, nil)

	// Export audit logs
	exported, err := auditLogger.ExportAuditLogs(context.Background(), nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(exported) == 0 {
		t.Error("Expected exported data to not be empty")
	}

	// Verify it's valid JSON
	if !isValidJSON(exported) {
		t.Error("Expected exported data to be valid JSON")
	}
}

// Helper function to check if data is valid JSON
func isValidJSON(data []byte) bool {
	var v interface{}
	return json.Unmarshal(data, &v) == nil
}
