package tenant

import (
	"context"
	"errors"
	"fmt"
)

// Context keys for tenant information
type contextKey string

const (
	TenantContextKey contextKey = "tenant_context"
	TenantIDKey      contextKey = "tenant_id"
	UserIDKey        contextKey = "user_id"
	UserRoleKey      contextKey = "user_role"
	PermissionsKey   contextKey = "permissions"
	APIKeyIDKey      contextKey = "api_key_id"
)

// WithTenantContext adds tenant context to the request context
func WithTenantContext(ctx context.Context, tenantCtx *TenantContext) context.Context {
	ctx = context.WithValue(ctx, TenantContextKey, tenantCtx)
	ctx = context.WithValue(ctx, TenantIDKey, tenantCtx.TenantID)
	ctx = context.WithValue(ctx, UserIDKey, tenantCtx.UserID)
	ctx = context.WithValue(ctx, UserRoleKey, tenantCtx.UserRole)
	ctx = context.WithValue(ctx, PermissionsKey, tenantCtx.Permissions)
	if tenantCtx.APIKeyID != "" {
		ctx = context.WithValue(ctx, APIKeyIDKey, tenantCtx.APIKeyID)
	}
	return ctx
}

// GetTenantContext retrieves the tenant context from the request context
func GetTenantContext(ctx context.Context) (*TenantContext, error) {
	tenantCtx, ok := ctx.Value(TenantContextKey).(*TenantContext)
	if !ok || tenantCtx == nil {
		return nil, errors.New("tenant context not found")
	}
	return tenantCtx, nil
}

// GetTenantID retrieves the tenant ID from the request context
func GetTenantID(ctx context.Context) (string, error) {
	tenantID, ok := ctx.Value(TenantIDKey).(string)
	if !ok || tenantID == "" {
		return "", errors.New("tenant ID not found in context")
	}
	return tenantID, nil
}

// GetUserID retrieves the user ID from the request context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok || userID == "" {
		return "", errors.New("user ID not found in context")
	}
	return userID, nil
}

// GetUserRole retrieves the user role from the request context
func GetUserRole(ctx context.Context) (TenantUserRole, error) {
	userRole, ok := ctx.Value(UserRoleKey).(TenantUserRole)
	if !ok {
		return "", errors.New("user role not found in context")
	}
	return userRole, nil
}

// GetPermissions retrieves the user permissions from the request context
func GetPermissions(ctx context.Context) ([]string, error) {
	permissions, ok := ctx.Value(PermissionsKey).([]string)
	if !ok {
		return nil, errors.New("permissions not found in context")
	}
	return permissions, nil
}

// GetAPIKeyID retrieves the API key ID from the request context
func GetAPIKeyID(ctx context.Context) (string, error) {
	apiKeyID, ok := ctx.Value(APIKeyIDKey).(string)
	if !ok {
		return "", errors.New("API key ID not found in context")
	}
	return apiKeyID, nil
}

// HasPermission checks if the user has a specific permission
func HasPermission(ctx context.Context, permission string) bool {
	permissions, err := GetPermissions(ctx)
	if err != nil {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// HasAnyPermission checks if the user has any of the specified permissions
func HasAnyPermission(ctx context.Context, permissions []string) bool {
	userPermissions, err := GetPermissions(ctx)
	if err != nil {
		return false
	}

	for _, required := range permissions {
		for _, user := range userPermissions {
			if user == required {
				return true
			}
		}
	}
	return false
}

// HasAllPermissions checks if the user has all of the specified permissions
func HasAllPermissions(ctx context.Context, permissions []string) bool {
	userPermissions, err := GetPermissions(ctx)
	if err != nil {
		return false
	}

	permissionMap := make(map[string]bool)
	for _, p := range userPermissions {
		permissionMap[p] = true
	}

	for _, required := range permissions {
		if !permissionMap[required] {
			return false
		}
	}
	return true
}

// HasRole checks if the user has a specific role
func HasRole(ctx context.Context, role TenantUserRole) bool {
	userRole, err := GetUserRole(ctx)
	if err != nil {
		return false
	}
	return userRole == role
}

// HasAnyRole checks if the user has any of the specified roles
func HasAnyRole(ctx context.Context, roles []TenantUserRole) bool {
	userRole, err := GetUserRole(ctx)
	if err != nil {
		return false
	}

	for _, role := range roles {
		if userRole == role {
			return true
		}
	}
	return false
}

// IsOwnerOrAdmin checks if the user is an owner or admin
func IsOwnerOrAdmin(ctx context.Context) bool {
	return HasAnyRole(ctx, []TenantUserRole{TenantUserRoleOwner, TenantUserRoleAdmin})
}

// IsOwner checks if the user is an owner
func IsOwner(ctx context.Context) bool {
	return HasRole(ctx, TenantUserRoleOwner)
}

// IsAdmin checks if the user is an admin
func IsAdmin(ctx context.Context) bool {
	return HasRole(ctx, TenantUserRoleAdmin)
}

// IsManagerOrAbove checks if the user is a manager or above
func IsManagerOrAbove(ctx context.Context) bool {
	return HasAnyRole(ctx, []TenantUserRole{
		TenantUserRoleOwner,
		TenantUserRoleAdmin,
		TenantUserRoleManager,
	})
}

// IsAnalystOrAbove checks if the user is an analyst or above
func IsAnalystOrAbove(ctx context.Context) bool {
	return HasAnyRole(ctx, []TenantUserRole{
		TenantUserRoleOwner,
		TenantUserRoleAdmin,
		TenantUserRoleManager,
		TenantUserRoleAnalyst,
	})
}

// RequirePermission creates a context that requires a specific permission
func RequirePermission(ctx context.Context, permission string) error {
	if !HasPermission(ctx, permission) {
		return fmt.Errorf("permission required: %s", permission)
	}
	return nil
}

// RequireAnyPermission creates a context that requires any of the specified permissions
func RequireAnyPermission(ctx context.Context, permissions []string) error {
	if !HasAnyPermission(ctx, permissions) {
		return fmt.Errorf("one of the following permissions required: %v", permissions)
	}
	return nil
}

// RequireAllPermissions creates a context that requires all of the specified permissions
func RequireAllPermissions(ctx context.Context, permissions []string) error {
	if !HasAllPermissions(ctx, permissions) {
		return fmt.Errorf("all of the following permissions required: %v", permissions)
	}
	return nil
}

// RequireRole creates a context that requires a specific role
func RequireRole(ctx context.Context, role TenantUserRole) error {
	if !HasRole(ctx, role) {
		return fmt.Errorf("role required: %s", role)
	}
	return nil
}

// RequireAnyRole creates a context that requires any of the specified roles
func RequireAnyRole(ctx context.Context, roles []TenantUserRole) error {
	if !HasAnyRole(ctx, roles) {
		return fmt.Errorf("one of the following roles required: %v", roles)
	}
	return nil
}

// RequireOwnerOrAdmin creates a context that requires owner or admin role
func RequireOwnerOrAdmin(ctx context.Context) error {
	if !IsOwnerOrAdmin(ctx) {
		return errors.New("owner or admin role required")
	}
	return nil
}

// RequireOwner creates a context that requires owner role
func RequireOwner(ctx context.Context) error {
	if !IsOwner(ctx) {
		return errors.New("owner role required")
	}
	return nil
}

// RequireAdmin creates a context that requires admin role
func RequireAdmin(ctx context.Context) error {
	if !IsAdmin(ctx) {
		return errors.New("admin role required")
	}
	return nil
}

// RequireManagerOrAbove creates a context that requires manager role or above
func RequireManagerOrAbove(ctx context.Context) error {
	if !IsManagerOrAbove(ctx) {
		return errors.New("manager role or above required")
	}
	return nil
}

// RequireAnalystOrAbove creates a context that requires analyst role or above
func RequireAnalystOrAbove(ctx context.Context) error {
	if !IsAnalystOrAbove(ctx) {
		return errors.New("analyst role or above required")
	}
	return nil
}

// ValidateTenantAccess validates that the user has access to the specified tenant
func ValidateTenantAccess(ctx context.Context, tenantID string) error {
	userTenantID, err := GetTenantID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get tenant ID from context: %w", err)
	}

	if userTenantID != tenantID {
		return fmt.Errorf("access denied: user does not belong to tenant %s", tenantID)
	}

	return nil
}

// ValidateResourceAccess validates that the user has access to a specific resource
func ValidateResourceAccess(ctx context.Context, resourceType, action string) error {
	permission := fmt.Sprintf("%s:%s", resourceType, action)
	return RequirePermission(ctx, permission)
}

// GetTenantContextFromRequest extracts tenant context from various sources
func GetTenantContextFromRequest(ctx context.Context, tenantID, userID, userRole string, permissions []string, apiKeyID string) *TenantContext {
	return &TenantContext{
		TenantID:    tenantID,
		UserID:      userID,
		UserRole:    TenantUserRole(userRole),
		Permissions: permissions,
		APIKeyID:    apiKeyID,
		Metadata:    make(map[string]interface{}),
	}
}

// CloneTenantContext creates a copy of the tenant context
func CloneTenantContext(tenantCtx *TenantContext) *TenantContext {
	if tenantCtx == nil {
		return nil
	}

	clone := &TenantContext{
		TenantID:    tenantCtx.TenantID,
		UserID:      tenantCtx.UserID,
		UserRole:    tenantCtx.UserRole,
		Permissions: make([]string, len(tenantCtx.Permissions)),
		APIKeyID:    tenantCtx.APIKeyID,
		Metadata:    make(map[string]interface{}),
	}

	copy(clone.Permissions, tenantCtx.Permissions)

	for k, v := range tenantCtx.Metadata {
		clone.Metadata[k] = v
	}

	return clone
}

// AddMetadata adds metadata to the tenant context
func AddMetadata(ctx context.Context, key string, value interface{}) error {
	tenantCtx, err := GetTenantContext(ctx)
	if err != nil {
		return err
	}

	if tenantCtx.Metadata == nil {
		tenantCtx.Metadata = make(map[string]interface{})
	}

	tenantCtx.Metadata[key] = value
	return nil
}

// GetMetadata retrieves metadata from the tenant context
func GetMetadata(ctx context.Context, key string) (interface{}, error) {
	tenantCtx, err := GetTenantContext(ctx)
	if err != nil {
		return nil, err
	}

	if tenantCtx.Metadata == nil {
		return nil, fmt.Errorf("metadata key not found: %s", key)
	}

	value, exists := tenantCtx.Metadata[key]
	if !exists {
		return nil, fmt.Errorf("metadata key not found: %s", key)
	}

	return value, nil
}

// RemoveMetadata removes metadata from the tenant context
func RemoveMetadata(ctx context.Context, key string) error {
	tenantCtx, err := GetTenantContext(ctx)
	if err != nil {
		return err
	}

	if tenantCtx.Metadata == nil {
		return nil
	}

	delete(tenantCtx.Metadata, key)
	return nil
}
