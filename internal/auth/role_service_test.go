package auth

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// MockDatabase provides a mock implementation for testing
type MockRoleDatabase struct {
	users           map[string]*database.User
	roleAssignments map[string]*database.RoleAssignment
	userAssignments map[string]string // userID -> assignmentID
}

func NewMockRoleDatabase() *MockRoleDatabase {
	return &MockRoleDatabase{
		users:           make(map[string]*database.User),
		roleAssignments: make(map[string]*database.RoleAssignment),
		userAssignments: make(map[string]string),
	}
}

func (m *MockRoleDatabase) GetUserByID(ctx context.Context, id string) (*database.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, database.ErrUserNotFound
}

func (m *MockRoleDatabase) UpdateUser(ctx context.Context, user *database.User) error {
	m.users[user.ID] = user
	return nil
}

func (m *MockRoleDatabase) CreateRoleAssignment(ctx context.Context, assignment *database.RoleAssignment) error {
	m.roleAssignments[assignment.ID] = assignment
	m.userAssignments[assignment.UserID] = assignment.ID
	return nil
}

func (m *MockRoleDatabase) GetActiveRoleAssignmentByUserID(ctx context.Context, userID string) (*database.RoleAssignment, error) {
	if assignmentID, exists := m.userAssignments[userID]; exists {
		if assignment, exists := m.roleAssignments[assignmentID]; exists && assignment.IsActive {
			// Check expiration
			if assignment.ExpiresAt == nil || assignment.ExpiresAt.After(time.Now()) {
				return assignment, nil
			}
		}
	}
	return nil, database.ErrRoleAssignmentNotFound
}

func (m *MockRoleDatabase) GetRoleAssignmentsByUserID(ctx context.Context, userID string) ([]*database.RoleAssignment, error) {
	var assignments []*database.RoleAssignment
	for _, assignment := range m.roleAssignments {
		if assignment.UserID == userID {
			assignments = append(assignments, assignment)
		}
	}
	return assignments, nil
}

func (m *MockRoleDatabase) DeactivateRoleAssignment(ctx context.Context, id string) error {
	if assignment, exists := m.roleAssignments[id]; exists {
		assignment.IsActive = false
		assignment.UpdatedAt = time.Now()
		return nil
	}
	return database.ErrRoleAssignmentNotFound
}

func (m *MockRoleDatabase) DeleteExpiredRoleAssignments(ctx context.Context) error {
	for _, assignment := range m.roleAssignments {
		if assignment.ExpiresAt != nil && assignment.ExpiresAt.Before(time.Now()) {
			assignment.IsActive = false
		}
	}
	return nil
}

// Implement required interface methods (stubs)
func (m *MockRoleDatabase) Connect(ctx context.Context) error                         { return nil }
func (m *MockRoleDatabase) Close() error                                              { return nil }
func (m *MockRoleDatabase) Ping(ctx context.Context) error                            { return nil }
func (m *MockRoleDatabase) CreateUser(ctx context.Context, user *database.User) error { return nil }
func (m *MockRoleDatabase) GetUserByEmail(ctx context.Context, email string) (*database.User, error) {
	return nil, nil
}
func (m *MockRoleDatabase) GetRoleAssignmentByID(ctx context.Context, id string) (*database.RoleAssignment, error) {
	return nil, nil
}
func (m *MockRoleDatabase) UpdateRoleAssignment(ctx context.Context, assignment *database.RoleAssignment) error {
	return nil
}
func (m *MockRoleDatabase) UpdateAPIKeyLastUsed(ctx context.Context, id string, lastUsed time.Time) error {
	return nil
}
func (m *MockRoleDatabase) GetActiveAPIKeysByRole(ctx context.Context, role string) ([]*database.APIKey, error) {
	return nil, nil
}
func (m *MockRoleDatabase) DeactivateAPIKey(ctx context.Context, id string) error  { return nil }
func (m *MockRoleDatabase) BeginTx(ctx context.Context) (database.Database, error) { return m, nil }
func (m *MockRoleDatabase) Commit() error                                          { return nil }
func (m *MockRoleDatabase) Rollback() error                                        { return nil }

// Additional stub methods to satisfy the Database interface
func (m *MockRoleDatabase) DeleteUser(ctx context.Context, id string) error {
	delete(m.users, id)
	return nil
}
func (m *MockRoleDatabase) ListBusinesses(ctx context.Context, limit, offset int) ([]*database.Business, error) {
	return nil, nil
}
func (m *MockRoleDatabase) ListUsers(ctx context.Context, limit, offset int) ([]*database.User, error) {
	return nil, nil
}
func (m *MockRoleDatabase) CreateEmailVerificationToken(ctx context.Context, token *database.EmailVerificationToken) error {
	return nil
}
func (m *MockRoleDatabase) GetEmailVerificationToken(ctx context.Context, token string) (*database.EmailVerificationToken, error) {
	return nil, nil
}
func (m *MockRoleDatabase) MarkEmailVerificationTokenUsed(ctx context.Context, token string) error {
	return nil
}
func (m *MockRoleDatabase) DeleteExpiredEmailVerificationTokens(ctx context.Context) error {
	return nil
}
func (m *MockRoleDatabase) CreatePasswordResetToken(ctx context.Context, token *database.PasswordResetToken) error {
	return nil
}
func (m *MockRoleDatabase) GetPasswordResetToken(ctx context.Context, token string) (*database.PasswordResetToken, error) {
	return nil, nil
}
func (m *MockRoleDatabase) MarkPasswordResetTokenUsed(ctx context.Context, token string) error {
	return nil
}
func (m *MockRoleDatabase) DeleteExpiredPasswordResetTokens(ctx context.Context) error { return nil }
func (m *MockRoleDatabase) CreateTokenBlacklist(ctx context.Context, blacklist *database.TokenBlacklist) error {
	return nil
}
func (m *MockRoleDatabase) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	return false, nil
}
func (m *MockRoleDatabase) DeleteExpiredTokenBlacklist(ctx context.Context) error { return nil }
func (m *MockRoleDatabase) CreateBusiness(ctx context.Context, business *database.Business) error {
	return nil
}
func (m *MockRoleDatabase) GetBusinessByID(ctx context.Context, id string) (*database.Business, error) {
	return nil, nil
}
func (m *MockRoleDatabase) GetBusinessByRegistrationNumber(ctx context.Context, regNumber string) (*database.Business, error) {
	return nil, nil
}
func (m *MockRoleDatabase) UpdateBusiness(ctx context.Context, business *database.Business) error {
	return nil
}
func (m *MockRoleDatabase) DeleteBusiness(ctx context.Context, id string) error { return nil }
func (m *MockRoleDatabase) SearchBusinesses(ctx context.Context, query string, limit, offset int) ([]*database.Business, error) {
	return nil, nil
}
func (m *MockRoleDatabase) CreateBusinessClassification(ctx context.Context, classification *database.BusinessClassification) error {
	return nil
}
func (m *MockRoleDatabase) GetBusinessClassificationByID(ctx context.Context, id string) (*database.BusinessClassification, error) {
	return nil, nil
}
func (m *MockRoleDatabase) GetBusinessClassificationsByBusinessID(ctx context.Context, businessID string) ([]*database.BusinessClassification, error) {
	return nil, nil
}
func (m *MockRoleDatabase) UpdateBusinessClassification(ctx context.Context, classification *database.BusinessClassification) error {
	return nil
}
func (m *MockRoleDatabase) DeleteBusinessClassification(ctx context.Context, id string) error {
	return nil
}
func (m *MockRoleDatabase) CreateRiskAssessment(ctx context.Context, assessment *database.RiskAssessment) error {
	return nil
}
func (m *MockRoleDatabase) GetRiskAssessmentByID(ctx context.Context, id string) (*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockRoleDatabase) GetRiskAssessmentsByBusinessID(ctx context.Context, businessID string) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockRoleDatabase) UpdateRiskAssessment(ctx context.Context, assessment *database.RiskAssessment) error {
	return nil
}
func (m *MockRoleDatabase) DeleteRiskAssessment(ctx context.Context, id string) error { return nil }
func (m *MockRoleDatabase) CreateComplianceCheck(ctx context.Context, check *database.ComplianceCheck) error {
	return nil
}
func (m *MockRoleDatabase) GetComplianceCheckByID(ctx context.Context, id string) (*database.ComplianceCheck, error) {
	return nil, nil
}
func (m *MockRoleDatabase) GetComplianceChecksByBusinessID(ctx context.Context, businessID string) ([]*database.ComplianceCheck, error) {
	return nil, nil
}
func (m *MockRoleDatabase) UpdateComplianceCheck(ctx context.Context, check *database.ComplianceCheck) error {
	return nil
}
func (m *MockRoleDatabase) DeleteComplianceCheck(ctx context.Context, id string) error { return nil }
func (m *MockRoleDatabase) CreateAPIKey(ctx context.Context, apiKey *database.APIKey) error {
	return nil
}
func (m *MockRoleDatabase) GetAPIKeyByID(ctx context.Context, id string) (*database.APIKey, error) {
	return nil, nil
}
func (m *MockRoleDatabase) GetAPIKeyByHash(ctx context.Context, keyHash string) (*database.APIKey, error) {
	return nil, nil
}
func (m *MockRoleDatabase) UpdateAPIKey(ctx context.Context, apiKey *database.APIKey) error {
	return nil
}
func (m *MockRoleDatabase) DeleteAPIKey(ctx context.Context, id string) error { return nil }
func (m *MockRoleDatabase) ListAPIKeysByUserID(ctx context.Context, userID string) ([]*database.APIKey, error) {
	return nil, nil
}
func (m *MockRoleDatabase) CreateAuditLog(ctx context.Context, log *database.AuditLog) error {
	return nil
}
func (m *MockRoleDatabase) GetAuditLogsByUserID(ctx context.Context, userID string, limit, offset int) ([]*database.AuditLog, error) {
	return nil, nil
}
func (m *MockRoleDatabase) GetAuditLogsByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*database.AuditLog, error) {
	return nil, nil
}
func (m *MockRoleDatabase) CreateExternalServiceCall(ctx context.Context, call *database.ExternalServiceCall) error {
	return nil
}
func (m *MockRoleDatabase) GetExternalServiceCallsByUserID(ctx context.Context, userID string, limit, offset int) ([]*database.ExternalServiceCall, error) {
	return nil, nil
}
func (m *MockRoleDatabase) GetExternalServiceCallsByService(ctx context.Context, serviceName string, limit, offset int) ([]*database.ExternalServiceCall, error) {
	return nil, nil
}
func (m *MockRoleDatabase) CreateWebhook(ctx context.Context, webhook *database.Webhook) error {
	return nil
}
func (m *MockRoleDatabase) GetWebhookByID(ctx context.Context, id string) (*database.Webhook, error) {
	return nil, nil
}
func (m *MockRoleDatabase) GetWebhooksByUserID(ctx context.Context, userID string) ([]*database.Webhook, error) {
	return nil, nil
}
func (m *MockRoleDatabase) UpdateWebhook(ctx context.Context, webhook *database.Webhook) error {
	return nil
}
func (m *MockRoleDatabase) DeleteWebhook(ctx context.Context, id string) error { return nil }
func (m *MockRoleDatabase) CreateWebhookEvent(ctx context.Context, event *database.WebhookEvent) error {
	return nil
}
func (m *MockRoleDatabase) GetWebhookEventByID(ctx context.Context, id string) (*database.WebhookEvent, error) {
	return nil, nil
}
func (m *MockRoleDatabase) GetWebhookEventsByWebhookID(ctx context.Context, webhookID string, limit, offset int) ([]*database.WebhookEvent, error) {
	return nil, nil
}
func (m *MockRoleDatabase) UpdateWebhookEvent(ctx context.Context, event *database.WebhookEvent) error {
	return nil
}
func (m *MockRoleDatabase) DeleteWebhookEvent(ctx context.Context, id string) error { return nil }

func setupRoleServiceTest() (*RoleService, *MockRoleDatabase) {
	mockDB := NewMockRoleDatabase()

	// Create a simple logger config
	loggerConfig := &config.ObservabilityConfig{
		LogLevel: "debug",
	}
	logger := observability.NewLogger(loggerConfig)

	// Create a minimal RBAC service for testing
	authService := &AuthService{} // Simplified for testing
	rbacService := NewRBACService(authService)

	roleService := NewRoleService(mockDB, logger, rbacService)

	return roleService, mockDB
}

func TestAssignRole(t *testing.T) {
	roleService, mockDB := setupRoleServiceTest()
	ctx := context.Background()

	// Setup test users
	adminUser := &database.User{
		ID:    "admin1",
		Email: "admin@test.com",
		Role:  string(RoleAdmin),
	}
	targetUser := &database.User{
		ID:    "user1",
		Email: "user@test.com",
		Role:  string(RoleUser),
	}

	mockDB.users[adminUser.ID] = adminUser
	mockDB.users[targetUser.ID] = targetUser

	t.Run("successful role assignment", func(t *testing.T) {
		request := &AssignRoleRequest{
			UserID:     targetUser.ID,
			Role:       RoleAnalyst,
			AssignedBy: adminUser.ID,
		}

		response, err := roleService.AssignRole(ctx, request)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if response.Role != RoleAnalyst {
			t.Errorf("Expected role %s, got %s", RoleAnalyst, response.Role)
		}

		if response.UserID != targetUser.ID {
			t.Errorf("Expected user ID %s, got %s", targetUser.ID, response.UserID)
		}

		if response.AssignedBy != adminUser.ID {
			t.Errorf("Expected assigned by %s, got %s", adminUser.ID, response.AssignedBy)
		}
	})

	t.Run("invalid role assignment", func(t *testing.T) {
		request := &AssignRoleRequest{
			UserID:     targetUser.ID,
			Role:       "invalid_role",
			AssignedBy: adminUser.ID,
		}

		_, err := roleService.AssignRole(ctx, request)
		if err == nil {
			t.Fatal("Expected error for invalid role, got none")
		}
	})

	t.Run("unauthorized role assignment", func(t *testing.T) {
		// Regular user trying to assign admin role
		regularUser := &database.User{
			ID:    "user2",
			Email: "user2@test.com",
			Role:  string(RoleUser),
		}
		mockDB.users[regularUser.ID] = regularUser

		request := &AssignRoleRequest{
			UserID:     targetUser.ID,
			Role:       RoleAdmin,
			AssignedBy: regularUser.ID,
		}

		_, err := roleService.AssignRole(ctx, request)
		if err == nil {
			t.Fatal("Expected error for unauthorized assignment, got none")
		}
	})
}

func TestGetUserRoleInfo(t *testing.T) {
	roleService, mockDB := setupRoleServiceTest()
	ctx := context.Background()

	// Setup test user
	testUser := &database.User{
		ID:    "user1",
		Email: "user@test.com",
		Role:  string(RoleUser),
	}
	mockDB.users[testUser.ID] = testUser

	t.Run("user with default role", func(t *testing.T) {
		info, err := roleService.GetUserRoleInfo(ctx, testUser.ID, false)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if info.CurrentRole != RoleUser {
			t.Errorf("Expected role %s, got %s", RoleUser, info.CurrentRole)
		}

		if len(info.Permissions) == 0 {
			t.Error("Expected permissions to be populated")
		}
	})

	t.Run("user with assigned role", func(t *testing.T) {
		// Create role assignment
		assignment := &database.RoleAssignment{
			ID:         "assignment1",
			UserID:     testUser.ID,
			Role:       string(RoleAnalyst),
			AssignedBy: "admin1",
			AssignedAt: time.Now(),
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockDB.roleAssignments[assignment.ID] = assignment
		mockDB.userAssignments[testUser.ID] = assignment.ID

		info, err := roleService.GetUserRoleInfo(ctx, testUser.ID, true)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if info.CurrentRole != RoleAnalyst {
			t.Errorf("Expected role %s, got %s", RoleAnalyst, info.CurrentRole)
		}

		if info.Assignment == nil {
			t.Error("Expected assignment to be populated")
		}

		if len(info.History) == 0 {
			t.Error("Expected history to be populated when requested")
		}
	})
}

func TestValidateRoleAssignment(t *testing.T) {
	roleService, mockDB := setupRoleServiceTest()
	ctx := context.Background()

	// Setup test user
	testUser := &database.User{
		ID:    "user1",
		Email: "user@test.com",
		Role:  string(RoleUser),
	}
	mockDB.users[testUser.ID] = testUser

	t.Run("user with active assignment", func(t *testing.T) {
		// Create active role assignment
		assignment := &database.RoleAssignment{
			ID:         "assignment1",
			UserID:     testUser.ID,
			Role:       string(RoleAnalyst),
			AssignedBy: "admin1",
			AssignedAt: time.Now(),
			IsActive:   true,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		mockDB.roleAssignments[assignment.ID] = assignment
		mockDB.userAssignments[testUser.ID] = assignment.ID

		role, err := roleService.ValidateRoleAssignment(ctx, testUser.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		if *role != RoleAnalyst {
			t.Errorf("Expected role %s, got %s", RoleAnalyst, *role)
		}
	})

	t.Run("user with expired assignment", func(t *testing.T) {
		// Create expired role assignment
		expiredTime := time.Now().Add(-24 * time.Hour)
		assignment := &database.RoleAssignment{
			ID:         "assignment2",
			UserID:     testUser.ID,
			Role:       string(RoleManager),
			AssignedBy: "admin1",
			AssignedAt: time.Now().Add(-48 * time.Hour),
			ExpiresAt:  &expiredTime,
			IsActive:   true,
			CreatedAt:  time.Now().Add(-48 * time.Hour),
			UpdatedAt:  time.Now().Add(-48 * time.Hour),
		}

		mockDB.roleAssignments[assignment.ID] = assignment
		mockDB.userAssignments[testUser.ID] = assignment.ID

		role, err := roleService.ValidateRoleAssignment(ctx, testUser.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Should fall back to default user role
		if *role != RoleUser {
			t.Errorf("Expected fallback role %s, got %s", RoleUser, *role)
		}
	})
}

func TestRolePermissions(t *testing.T) {
	t.Run("admin has all permissions", func(t *testing.T) {
		permissions := GetRolePermissions(RoleAdmin)

		expectedPermissions := []Permission{
			PermissionClassifyBusiness,
			PermissionViewUsers,
			PermissionManageRoles,
			PermissionViewMetrics,
		}

		for _, expected := range expectedPermissions {
			found := false
			for _, perm := range permissions {
				if perm == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Admin role missing permission: %s", expected)
			}
		}
	})

	t.Run("user has limited permissions", func(t *testing.T) {
		permissions := GetRolePermissions(RoleUser)

		// User should have basic permissions
		hasClassify := false
		for _, perm := range permissions {
			if perm == PermissionClassifyBusiness {
				hasClassify = true
				break
			}
		}
		if !hasClassify {
			t.Error("User role should have classify permission")
		}

		// User should not have admin permissions
		for _, perm := range permissions {
			if perm == PermissionManageRoles {
				t.Error("User role should not have manage roles permission")
			}
		}
	})
}

func TestCanAssignRole(t *testing.T) {
	tests := []struct {
		assignerRole Role
		targetRole   Role
		expected     bool
		description  string
	}{
		{RoleAdmin, RoleManager, true, "Admin can assign manager role"},
		{RoleAdmin, RoleUser, true, "Admin can assign user role"},
		{RoleAdmin, RoleSystem, false, "Admin cannot assign system role"},
		{RoleManager, RoleUser, true, "Manager can assign user role"},
		{RoleManager, RoleAdmin, false, "Manager cannot assign admin role"},
		{RoleUser, RoleAnalyst, false, "User cannot assign any role"},
		{RoleGuest, RoleUser, false, "Guest cannot assign any role"},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := CanAssignRole(test.assignerRole, test.targetRole)
			if result != test.expected {
				t.Errorf("Expected %v, got %v for %s assigning %s",
					test.expected, result, test.assignerRole, test.targetRole)
			}
		})
	}
}
