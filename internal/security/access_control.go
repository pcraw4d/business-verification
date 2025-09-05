package security

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// AccessControlSystem provides comprehensive access control capabilities
type AccessControlSystem struct {
	logger          *observability.Logger
	roles           map[string]*Role
	permissions     map[string]*Permission
	policies        map[string]*Policy
	userRoles       map[string][]string // userID -> roleIDs
	rolePermissions map[string][]string // roleID -> permissionIDs
	mutex           sync.RWMutex
	config          AccessControlConfig
	auditLogger     *AccessAuditLogger
}

// AccessControlConfig defines configuration for access control
type AccessControlConfig struct {
	DefaultRole           string         `json:"default_role"`
	SessionTimeout        time.Duration  `json:"session_timeout"`
	MaxLoginAttempts      int            `json:"max_login_attempts"`
	LockoutDuration       time.Duration  `json:"lockout_duration"`
	PasswordPolicy        PasswordPolicy `json:"password_policy"`
	MFAEnabled            bool           `json:"mfa_enabled"`
	AuditLoggingEnabled   bool           `json:"audit_logging_enabled"`
	PolicyEnforcementMode string         `json:"policy_enforcement_mode"` // strict, permissive
}

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength           int  `json:"min_length"`
	RequireUppercase    bool `json:"require_uppercase"`
	RequireLowercase    bool `json:"require_lowercase"`
	RequireNumbers      bool `json:"require_numbers"`
	RequireSpecialChars bool `json:"require_special_chars"`
	MaxAge              int  `json:"max_age_days"`
	PreventReuse        int  `json:"prevent_reuse_count"`
}

// Role represents a user role with associated permissions
type Role struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Permissions []string               `json:"permissions"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	IsActive    bool                   `json:"is_active"`
}

// Permission represents a specific permission or capability
type Permission struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	IsActive    bool                   `json:"is_active"`
}

// Policy represents an access control policy
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        PolicyType             `json:"type"`
	Rules       []PolicyRule           `json:"rules"`
	Priority    int                    `json:"priority"`
	Effect      PolicyEffect           `json:"effect"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	IsActive    bool                   `json:"is_active"`
}

// PolicyType defines the type of access control policy
type PolicyType string

const (
	PolicyTypeRBAC      PolicyType = "rbac"
	PolicyTypeABAC      PolicyType = "abac"
	PolicyTypeTimeBased PolicyType = "time_based"
	PolicyTypeLocation  PolicyType = "location_based"
	PolicyTypeDevice    PolicyType = "device_based"
	PolicyTypeRisk      PolicyType = "risk_based"
)

// PolicyEffect defines the effect of a policy
type PolicyEffect string

const (
	PolicyEffectAllow PolicyEffect = "allow"
	PolicyEffectDeny  PolicyEffect = "deny"
)

// PolicyRule represents a rule within a policy
type PolicyRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Effect      PolicyEffect           `json:"effect"`
	Priority    int                    `json:"priority"`
}

// AccessRequest represents a request for access to a resource
type AccessRequest struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Resource  string                 `json:"resource"`
	Action    string                 `json:"action"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Status    AccessRequestStatus    `json:"status"`
	Decision  AccessDecision         `json:"decision"`
	Reason    string                 `json:"reason,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// AccessRequestStatus defines the status of an access request
type AccessRequestStatus string

const (
	AccessRequestStatusPending   AccessRequestStatus = "pending"
	AccessRequestStatusApproved  AccessRequestStatus = "approved"
	AccessRequestStatusDenied    AccessRequestStatus = "denied"
	AccessRequestStatusExpired   AccessRequestStatus = "expired"
	AccessRequestStatusCancelled AccessRequestStatus = "cancelled"
)

// AccessDecision represents the decision made on an access request
type AccessDecision struct {
	Allowed     bool                   `json:"allowed"`
	Reason      string                 `json:"reason"`
	PolicyID    string                 `json:"policy_id,omitempty"`
	RoleID      string                 `json:"role_id,omitempty"`
	Permissions []string               `json:"permissions,omitempty"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// AccessAuditLogger provides audit logging for access control events
type AccessAuditLogger struct {
	logger *observability.Logger
	events []AccessAuditEvent
	mutex  sync.RWMutex
	config AuditLoggerConfig
}

// AuditLoggerConfig defines configuration for audit logging
type AuditLoggerConfig struct {
	RetentionDays int    `json:"retention_days"`
	LogToFile     bool   `json:"log_to_file"`
	LogToDatabase bool   `json:"log_to_database"`
	LogLevel      string `json:"log_level"`
}

// AccessAuditEvent represents an audit event for access control
type AccessAuditEvent struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	IPAddress string                 `json:"ip_address,omitempty"`
	UserAgent string                 `json:"user_agent,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewAccessControlSystem creates a new access control system
func NewAccessControlSystem(logger *observability.Logger, config AccessControlConfig) *AccessControlSystem {
	acs := &AccessControlSystem{
		logger:          logger,
		roles:           make(map[string]*Role),
		permissions:     make(map[string]*Permission),
		policies:        make(map[string]*Policy),
		userRoles:       make(map[string][]string),
		rolePermissions: make(map[string][]string),
		config:          config,
		auditLogger: NewAccessAuditLogger(logger, AuditLoggerConfig{
			RetentionDays: 90,
			LogToFile:     true,
			LogToDatabase: false,
			LogLevel:      "info",
		}),
	}

	// Initialize default roles and permissions
	acs.initializeDefaultRoles()
	acs.initializeDefaultPermissions()

	return acs
}

// CheckAccess checks if a user has access to perform an action on a resource
func (acs *AccessControlSystem) CheckAccess(ctx context.Context, userID, resource, action string, context map[string]interface{}) (*AccessDecision, error) {
	acs.mutex.RLock()
	defer acs.mutex.RUnlock()

	// Create access request
	request := &AccessRequest{
		ID:        observability.GenerateRequestID(),
		UserID:    userID,
		Resource:  resource,
		Action:    action,
		Context:   context,
		Timestamp: time.Now(),
		Status:    AccessRequestStatusPending,
	}

	// Evaluate policies
	decision := acs.evaluatePolicies(request)

	// Log audit event
	acs.auditLogger.LogAccessEvent(ctx, userID, "access_check", resource, decision.Allowed, context)

	return &decision, nil
}

// GrantRole grants a role to a user
func (acs *AccessControlSystem) GrantRole(ctx context.Context, userID, roleID string) error {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	// Check if role exists
	if _, exists := acs.roles[roleID]; !exists {
		return fmt.Errorf("role not found: %s", roleID)
	}

	// Check if user already has the role
	for _, existingRole := range acs.userRoles[userID] {
		if existingRole == roleID {
			return fmt.Errorf("user already has role: %s", roleID)
		}
	}

	// Grant role
	acs.userRoles[userID] = append(acs.userRoles[userID], roleID)

	// Log audit event
	acs.auditLogger.LogAccessEvent(ctx, userID, "role_granted", roleID, true, map[string]interface{}{
		"role_id": roleID,
	})

	acs.logger.Info("Role granted to user", map[string]interface{}{
		"user_id": userID,
		"role_id": roleID,
	})

	return nil
}

// RevokeRole revokes a role from a user
func (acs *AccessControlSystem) RevokeRole(ctx context.Context, userID, roleID string) error {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	// Find and remove role
	roles := acs.userRoles[userID]
	for i, role := range roles {
		if role == roleID {
			// Remove role
			acs.userRoles[userID] = append(roles[:i], roles[i+1:]...)

			// Log audit event
			acs.auditLogger.LogAccessEvent(ctx, userID, "role_revoked", roleID, true, map[string]interface{}{
				"role_id": roleID,
			})

			acs.logger.Info("Role revoked from user", map[string]interface{}{
				"user_id": userID,
				"role_id": roleID,
			})

			return nil
		}
	}

	return fmt.Errorf("user does not have role: %s", roleID)
}

// CreateRole creates a new role
func (acs *AccessControlSystem) CreateRole(ctx context.Context, role *Role) error {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	// Generate ID if not provided
	if role.ID == "" {
		role.ID = observability.GenerateRequestID()
	}

	// Set timestamps
	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	// Store role
	acs.roles[role.ID] = role

	// Initialize role permissions
	acs.rolePermissions[role.ID] = make([]string, 0)

	acs.logger.Info("Role created", map[string]interface{}{
		"role_id": role.ID,
		"role_name": role.Name,
	})

	return nil
}

// CreatePermission creates a new permission
func (acs *AccessControlSystem) CreatePermission(ctx context.Context, permission *Permission) error {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	// Generate ID if not provided
	if permission.ID == "" {
		permission.ID = observability.GenerateRequestID()
	}

	// Set timestamps
	now := time.Now()
	permission.CreatedAt = now
	permission.UpdatedAt = now

	// Store permission
	acs.permissions[permission.ID] = permission

	acs.logger.Info("Permission created", map[string]interface{}{
		"permission_id": permission.ID,
		"permission_name": permission.Name,
		"resource": permission.Resource,
		"action": permission.Action,
	})

	return nil
}

// CreatePolicy creates a new access control policy
func (acs *AccessControlSystem) CreatePolicy(ctx context.Context, policy *Policy) error {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	// Generate ID if not provided
	if policy.ID == "" {
		policy.ID = observability.GenerateRequestID()
	}

	// Set timestamps
	now := time.Now()
	policy.CreatedAt = now
	policy.UpdatedAt = now

	// Store policy
	acs.policies[policy.ID] = policy

	acs.logger.Info("Policy created", map[string]interface{}{
		"policy_id": policy.ID,
		"policy_name": policy.Name,
		"policy_type": policy.Type,
	})

	return nil
}

// GetUserRoles retrieves roles assigned to a user
func (acs *AccessControlSystem) GetUserRoles(ctx context.Context, userID string) ([]*Role, error) {
	acs.mutex.RLock()
	defer acs.mutex.RUnlock()

	roleIDs, exists := acs.userRoles[userID]
	if !exists {
		return []*Role{}, nil
	}

	roles := make([]*Role, 0, len(roleIDs))
	for _, roleID := range roleIDs {
		if role, exists := acs.roles[roleID]; exists && role.IsActive {
			roles = append(roles, role)
		}
	}

	return roles, nil
}

// GetRolePermissions retrieves permissions for a role
func (acs *AccessControlSystem) GetRolePermissions(ctx context.Context, roleID string) ([]*Permission, error) {
	acs.mutex.RLock()
	defer acs.mutex.RUnlock()

	permissionIDs, exists := acs.rolePermissions[roleID]
	if !exists {
		return []*Permission{}, nil
	}

	permissions := make([]*Permission, 0, len(permissionIDs))
	for _, permissionID := range permissionIDs {
		if permission, exists := acs.permissions[permissionID]; exists && permission.IsActive {
			permissions = append(permissions, permission)
		}
	}

	return permissions, nil
}

// AssignPermissionToRole assigns a permission to a role
func (acs *AccessControlSystem) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	// Check if role exists
	if _, exists := acs.roles[roleID]; !exists {
		return fmt.Errorf("role not found: %s", roleID)
	}

	// Check if permission exists
	if _, exists := acs.permissions[permissionID]; !exists {
		return fmt.Errorf("permission not found: %s", permissionID)
	}

	// Check if permission is already assigned
	for _, existingPermission := range acs.rolePermissions[roleID] {
		if existingPermission == permissionID {
			return fmt.Errorf("permission already assigned to role: %s", permissionID)
		}
	}

	// Assign permission
	acs.rolePermissions[roleID] = append(acs.rolePermissions[roleID], permissionID)

	acs.logger.Info("Permission assigned to role",
		"role_id", roleID,
		"permission_id", permissionID,
	)

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (acs *AccessControlSystem) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	acs.mutex.Lock()
	defer acs.mutex.Unlock()

	permissions := acs.rolePermissions[roleID]
	for i, permission := range permissions {
		if permission == permissionID {
			// Remove permission
			acs.rolePermissions[roleID] = append(permissions[:i], permissions[i+1:]...)

			acs.logger.Info("Permission removed from role",
				"role_id", roleID,
				"permission_id", permissionID,
			)

			return nil
		}
	}

	return fmt.Errorf("permission not assigned to role: %s", permissionID)
}

// GetAuditLogs retrieves audit logs with optional filtering
func (acs *AccessControlSystem) GetAuditLogs(ctx context.Context, filters map[string]interface{}) ([]AccessAuditEvent, error) {
	return acs.auditLogger.GetAuditLogs(ctx, filters)
}

// ExportAuditLogs exports audit logs to JSON format
func (acs *AccessControlSystem) ExportAuditLogs(ctx context.Context, filters map[string]interface{}) ([]byte, error) {
	return acs.auditLogger.ExportAuditLogs(ctx, filters)
}

// evaluatePolicies evaluates all applicable policies for an access request
func (acs *AccessControlSystem) evaluatePolicies(request *AccessRequest) AccessDecision {
	decision := AccessDecision{
		Allowed:   false,
		Timestamp: time.Now(),
	}

	// Get user roles
	userRoles := acs.userRoles[request.UserID]
	if len(userRoles) == 0 {
		decision.Reason = "No roles assigned to user"
		return decision
	}

	// Check role-based permissions first
	for _, roleID := range userRoles {
		role, exists := acs.roles[roleID]
		if !exists || !role.IsActive {
			continue
		}

		// Check role permissions
		for _, permissionID := range role.Permissions {
			permission, exists := acs.permissions[permissionID]
			if !exists || !permission.IsActive {
				continue
			}

			if acs.matchesPermission(permission, request.Resource, request.Action) {
				decision.Allowed = true
				decision.Reason = "Access granted by role permission"
				decision.RoleID = roleID
				decision.Permissions = append(decision.Permissions, permissionID)
				return decision
			}
		}
	}

	// Evaluate policies
	for _, policy := range acs.policies {
		if !policy.IsActive {
			continue
		}

		if acs.evaluatePolicy(policy, request) {
			decision.Allowed = (policy.Effect == PolicyEffectAllow)
			decision.Reason = fmt.Sprintf("Access %s by policy: %s", policy.Effect, policy.Name)
			decision.PolicyID = policy.ID
			return decision
		}
	}

	decision.Reason = "Access denied - no matching permissions or policies"
	return decision
}

// matchesPermission checks if a permission matches the requested resource and action
func (acs *AccessControlSystem) matchesPermission(permission *Permission, resource, action string) bool {
	// Check resource and action match
	if permission.Resource != resource || permission.Action != action {
		return false
	}

	// Check conditions if any
	if len(permission.Conditions) > 0 {
		// In a real implementation, this would evaluate conditions
		// For now, we'll assume conditions are met
		return true
	}

	return true
}

// evaluatePolicy evaluates if a policy applies to an access request
func (acs *AccessControlSystem) evaluatePolicy(policy *Policy, request *AccessRequest) bool {
	// Check policy rules
	for _, rule := range policy.Rules {
		if acs.evaluateRule(rule, request) {
			return true
		}
	}

	return false
}

// evaluateRule evaluates if a policy rule applies to an access request
func (acs *AccessControlSystem) evaluateRule(rule PolicyRule, request *AccessRequest) bool {
	// Check resource and action match
	if rule.Resource != request.Resource || rule.Action != request.Action {
		return false
	}

	// Check conditions if any
	if len(rule.Conditions) > 0 {
		// In a real implementation, this would evaluate conditions
		// For now, we'll assume conditions are met
		return true
	}

	return true
}

// initializeDefaultRoles creates default roles for the system
func (acs *AccessControlSystem) initializeDefaultRoles() {
	defaultRoles := []*Role{
		{
			ID:          "admin",
			Name:        "Administrator",
			Description: "Full system administrator with all permissions",
			Permissions: []string{},
			IsActive:    true,
		},
		{
			ID:          "user",
			Name:        "User",
			Description: "Standard user with basic permissions",
			Permissions: []string{},
			IsActive:    true,
		},
		{
			ID:          "readonly",
			Name:        "Read Only",
			Description: "Read-only access to system data",
			Permissions: []string{},
			IsActive:    true,
		},
	}

	for _, role := range defaultRoles {
		acs.roles[role.ID] = role
		acs.rolePermissions[role.ID] = make([]string, 0)
	}
}

// initializeDefaultPermissions creates default permissions for the system
func (acs *AccessControlSystem) initializeDefaultPermissions() {
	defaultPermissions := []*Permission{
		{
			ID:          "read:business",
			Name:        "Read Business Data",
			Description: "Read access to business verification data",
			Resource:    "business",
			Action:      "read",
			IsActive:    true,
		},
		{
			ID:          "write:business",
			Name:        "Write Business Data",
			Description: "Write access to business verification data",
			Resource:    "business",
			Action:      "write",
			IsActive:    true,
		},
		{
			ID:          "read:reports",
			Name:        "Read Reports",
			Description: "Read access to system reports",
			Resource:    "reports",
			Action:      "read",
			IsActive:    true,
		},
		{
			ID:          "admin:users",
			Name:        "Manage Users",
			Description: "Manage user accounts and permissions",
			Resource:    "users",
			Action:      "admin",
			IsActive:    true,
		},
		{
			ID:          "admin:system",
			Name:        "System Administration",
			Description: "Full system administration access",
			Resource:    "system",
			Action:      "admin",
			IsActive:    true,
		},
	}

	for _, permission := range defaultPermissions {
		acs.permissions[permission.ID] = permission
	}
}

// NewAccessAuditLogger creates a new access audit logger
func NewAccessAuditLogger(logger *observability.Logger, config AuditLoggerConfig) *AccessAuditLogger {
	auditLogger := &AccessAuditLogger{
		logger: logger,
		events: make([]AccessAuditEvent, 0),
		config: config,
	}

	// Start cleanup routine
	go auditLogger.cleanupRoutine()

	return auditLogger
}

// LogAccessEvent logs an access control event
func (aal *AccessAuditLogger) LogAccessEvent(ctx context.Context, userID, action, resource string, success bool, context map[string]interface{}) {
	aal.mutex.Lock()
	defer aal.mutex.Unlock()

	event := AccessAuditEvent{
		ID:        observability.GenerateRequestID(),
		UserID:    userID,
		Action:    action,
		Resource:  resource,
		Result:    map[bool]string{true: "success", false: "denied"}[success],
		Context:   context,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Extract IP address and user agent from context if available
	if ip, ok := context["ip_address"].(string); ok {
		event.IPAddress = ip
	}
	if ua, ok := context["user_agent"].(string); ok {
		event.UserAgent = ua
	}
	if session, ok := context["session_id"].(string); ok {
		event.SessionID = session
	}

	// Add to events
	aal.events = append(aal.events, event)

	// Log to logger
	aal.logger.Info("Access control event",
		"user_id", userID,
		"action", action,
		"resource", resource,
		"result", event.Result,
		"ip_address", event.IPAddress,
	)
}

// GetAuditLogs retrieves audit logs with optional filtering
func (aal *AccessAuditLogger) GetAuditLogs(ctx context.Context, filters map[string]interface{}) ([]AccessAuditEvent, error) {
	aal.mutex.RLock()
	defer aal.mutex.RUnlock()

	var filteredEvents []AccessAuditEvent

	for _, event := range aal.events {
		if aal.matchesFilters(event, filters) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	return filteredEvents, nil
}

// ExportAuditLogs exports audit logs to JSON format
func (aal *AccessAuditLogger) ExportAuditLogs(ctx context.Context, filters map[string]interface{}) ([]byte, error) {
	events, err := aal.GetAuditLogs(ctx, filters)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(events, "", "  ")
}

// matchesFilters checks if an event matches the given filters
func (aal *AccessAuditLogger) matchesFilters(event AccessAuditEvent, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "user_id":
			if userID, ok := value.(string); ok && event.UserID != userID {
				return false
			}
		case "action":
			if action, ok := value.(string); ok && event.Action != action {
				return false
			}
		case "resource":
			if resource, ok := value.(string); ok && event.Resource != resource {
				return false
			}
		case "result":
			if result, ok := value.(string); ok && event.Result != result {
				return false
			}
		}
	}
	return true
}

// cleanupRoutine periodically cleans up old audit events
func (aal *AccessAuditLogger) cleanupRoutine() {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	defer ticker.Stop()

	for range ticker.C {
		aal.cleanup()
	}
}

// cleanup removes old audit events based on retention policy
func (aal *AccessAuditLogger) cleanup() {
	aal.mutex.Lock()
	defer aal.mutex.Unlock()

	cutoff := time.Now().AddDate(0, 0, -aal.config.RetentionDays)

	var filteredEvents []AccessAuditEvent
	for _, event := range aal.events {
		if event.Timestamp.After(cutoff) {
			filteredEvents = append(filteredEvents, event)
		}
	}

	aal.events = filteredEvents
}
