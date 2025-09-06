package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
	"go.uber.org/zap"
)

// MockAdminDatabase implements database.Database for testing
type MockAdminDatabase struct {
	users   map[string]*database.User
	apiKeys map[string]*database.APIKey
}

func NewMockAdminDatabase() *MockAdminDatabase {
	return &MockAdminDatabase{
		users:   make(map[string]*database.User),
		apiKeys: make(map[string]*database.APIKey),
	}
}

// Implement required interface methods
func (m *MockAdminDatabase) Connect(ctx context.Context) error { return nil }
func (m *MockAdminDatabase) Close() error                      { return nil }
func (m *MockAdminDatabase) Ping(ctx context.Context) error    { return nil }
func (m *MockAdminDatabase) CreateUser(ctx context.Context, user *database.User) error {
	m.users[user.ID] = user
	return nil
}
func (m *MockAdminDatabase) GetUserByID(ctx context.Context, id string) (*database.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, database.ErrUserNotFound
}
func (m *MockAdminDatabase) GetUserByEmail(ctx context.Context, email string) (*database.User, error) {
	for _, user := range m.users {
		if user.Email == email {
			return user, nil
		}
	}
	return nil, database.ErrUserNotFound
}
func (m *MockAdminDatabase) UpdateUser(ctx context.Context, user *database.User) error {
	if _, exists := m.users[user.ID]; !exists {
		return database.ErrUserNotFound
	}
	m.users[user.ID] = user
	return nil
}
func (m *MockAdminDatabase) DeleteUser(ctx context.Context, id string) error {
	if _, exists := m.users[id]; !exists {
		return database.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}
func (m *MockAdminDatabase) ListUsers(ctx context.Context, limit, offset int) ([]*database.User, error) {
	var users []*database.User
	count := 0
	for _, user := range m.users {
		if count >= offset && count < offset+limit {
			users = append(users, user)
		}
		count++
	}
	return users, nil
}

// Stub implementations for other required methods
func (m *MockAdminDatabase) CreateEmailVerificationToken(ctx context.Context, token *database.EmailVerificationToken) error {
	return nil
}
func (m *MockAdminDatabase) GetEmailVerificationToken(ctx context.Context, token string) (*database.EmailVerificationToken, error) {
	return nil, nil
}
func (m *MockAdminDatabase) MarkEmailVerificationTokenUsed(ctx context.Context, token string) error {
	return nil
}
func (m *MockAdminDatabase) DeleteExpiredEmailVerificationTokens(ctx context.Context) error {
	return nil
}
func (m *MockAdminDatabase) CreatePasswordResetToken(ctx context.Context, token *database.PasswordResetToken) error {
	return nil
}
func (m *MockAdminDatabase) GetPasswordResetToken(ctx context.Context, token string) (*database.PasswordResetToken, error) {
	return nil, nil
}
func (m *MockAdminDatabase) MarkPasswordResetTokenUsed(ctx context.Context, token string) error {
	return nil
}
func (m *MockAdminDatabase) DeleteExpiredPasswordResetTokens(ctx context.Context) error { return nil }
func (m *MockAdminDatabase) CreateTokenBlacklist(ctx context.Context, blacklist *database.TokenBlacklist) error {
	return nil
}
func (m *MockAdminDatabase) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	return false, nil
}
func (m *MockAdminDatabase) DeleteExpiredTokenBlacklist(ctx context.Context) error { return nil }
func (m *MockAdminDatabase) CreateBusiness(ctx context.Context, business *database.Business) error {
	return nil
}
func (m *MockAdminDatabase) GetBusinessByID(ctx context.Context, id string) (*database.Business, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetBusinessByRegistrationNumber(ctx context.Context, regNumber string) (*database.Business, error) {
	return nil, nil
}
func (m *MockAdminDatabase) UpdateBusiness(ctx context.Context, business *database.Business) error {
	return nil
}
func (m *MockAdminDatabase) DeleteBusiness(ctx context.Context, id string) error { return nil }
func (m *MockAdminDatabase) ListBusinesses(ctx context.Context, limit, offset int) ([]*database.Business, error) {
	return nil, nil
}
func (m *MockAdminDatabase) SearchBusinesses(ctx context.Context, query string, limit, offset int) ([]*database.Business, error) {
	return nil, nil
}
func (m *MockAdminDatabase) CreateBusinessClassification(ctx context.Context, classification *database.BusinessClassification) error {
	return nil
}
func (m *MockAdminDatabase) GetBusinessClassificationByID(ctx context.Context, id string) (*database.BusinessClassification, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetBusinessClassificationsByBusinessID(ctx context.Context, businessID string) ([]*database.BusinessClassification, error) {
	return nil, nil
}
func (m *MockAdminDatabase) UpdateBusinessClassification(ctx context.Context, classification *database.BusinessClassification) error {
	return nil
}
func (m *MockAdminDatabase) DeleteBusinessClassification(ctx context.Context, id string) error {
	return nil
}
func (m *MockAdminDatabase) CreateRiskAssessment(ctx context.Context, assessment *database.RiskAssessment) error {
	return nil
}
func (m *MockAdminDatabase) GetRiskAssessmentByID(ctx context.Context, id string) (*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetRiskAssessmentsByBusinessID(ctx context.Context, businessID string) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) UpdateRiskAssessment(ctx context.Context, assessment *database.RiskAssessment) error {
	return nil
}
func (m *MockAdminDatabase) DeleteRiskAssessment(ctx context.Context, id string) error { return nil }
func (m *MockAdminDatabase) GetRiskAssessmentHistory(ctx context.Context, businessID string, limit, offset int) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetRiskAssessmentHistoryByDateRange(ctx context.Context, businessID string, startDate, endDate time.Time) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetLatestRiskAssessment(ctx context.Context, businessID string) (*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetRiskAssessmentTrends(ctx context.Context, businessID string, days int) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetRiskAssessmentsByLevel(ctx context.Context, businessID string, riskLevel string) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetRiskAssessmentsByScoreRange(ctx context.Context, businessID string, minScore, maxScore float64) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetRiskAssessmentStatistics(ctx context.Context, businessID string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockAdminDatabase) CreateComplianceCheck(ctx context.Context, check *database.ComplianceCheck) error {
	return nil
}
func (m *MockAdminDatabase) GetComplianceCheckByID(ctx context.Context, id string) (*database.ComplianceCheck, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetComplianceChecksByBusinessID(ctx context.Context, businessID string) ([]*database.ComplianceCheck, error) {
	return nil, nil
}
func (m *MockAdminDatabase) UpdateComplianceCheck(ctx context.Context, check *database.ComplianceCheck) error {
	return nil
}
func (m *MockAdminDatabase) DeleteComplianceCheck(ctx context.Context, id string) error { return nil }
func (m *MockAdminDatabase) CreateAPIKey(ctx context.Context, apiKey *database.APIKey) error {
	m.apiKeys[apiKey.ID] = apiKey
	return nil
}
func (m *MockAdminDatabase) GetAPIKeyByID(ctx context.Context, id string) (*database.APIKey, error) {
	if apiKey, exists := m.apiKeys[id]; exists {
		return apiKey, nil
	}
	return nil, database.ErrAPIKeyNotFound
}
func (m *MockAdminDatabase) GetAPIKeyByHash(ctx context.Context, hash string) (*database.APIKey, error) {
	for _, apiKey := range m.apiKeys {
		if apiKey.KeyHash == hash {
			return apiKey, nil
		}
	}
	return nil, database.ErrAPIKeyNotFound
}
func (m *MockAdminDatabase) UpdateAPIKey(ctx context.Context, apiKey *database.APIKey) error {
	m.apiKeys[apiKey.ID] = apiKey
	return nil
}
func (m *MockAdminDatabase) DeleteAPIKey(ctx context.Context, id string) error {
	delete(m.apiKeys, id)
	return nil
}
func (m *MockAdminDatabase) ListAPIKeysByUserID(ctx context.Context, userID string) ([]*database.APIKey, error) {
	var keys []*database.APIKey
	for _, apiKey := range m.apiKeys {
		if userID == "" || apiKey.UserID == userID {
			keys = append(keys, apiKey)
		}
	}
	return keys, nil
}
func (m *MockAdminDatabase) CreateAuditLog(ctx context.Context, log *database.AuditLog) error {
	return nil
}
func (m *MockAdminDatabase) GetAuditLogsByUserID(ctx context.Context, userID string, limit, offset int) ([]*database.AuditLog, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetAuditLogsByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*database.AuditLog, error) {
	return nil, nil
}
func (m *MockAdminDatabase) CreateExternalServiceCall(ctx context.Context, call *database.ExternalServiceCall) error {
	return nil
}
func (m *MockAdminDatabase) GetExternalServiceCallsByUserID(ctx context.Context, userID string, limit, offset int) ([]*database.ExternalServiceCall, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetExternalServiceCallsByService(ctx context.Context, serviceName string, limit, offset int) ([]*database.ExternalServiceCall, error) {
	return nil, nil
}
func (m *MockAdminDatabase) CreateWebhook(ctx context.Context, webhook *database.Webhook) error {
	return nil
}
func (m *MockAdminDatabase) GetWebhookByID(ctx context.Context, id string) (*database.Webhook, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetWebhooksByUserID(ctx context.Context, userID string) ([]*database.Webhook, error) {
	return nil, nil
}
func (m *MockAdminDatabase) UpdateWebhook(ctx context.Context, webhook *database.Webhook) error {
	return nil
}
func (m *MockAdminDatabase) DeleteWebhook(ctx context.Context, id string) error { return nil }
func (m *MockAdminDatabase) CreateWebhookEvent(ctx context.Context, event *database.WebhookEvent) error {
	return nil
}
func (m *MockAdminDatabase) GetWebhookEventByID(ctx context.Context, id string) (*database.WebhookEvent, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetWebhookEventsByWebhookID(ctx context.Context, webhookID string, limit, offset int) ([]*database.WebhookEvent, error) {
	return nil, nil
}
func (m *MockAdminDatabase) UpdateWebhookEvent(ctx context.Context, event *database.WebhookEvent) error {
	return nil
}
func (m *MockAdminDatabase) DeleteWebhookEvent(ctx context.Context, id string) error { return nil }
func (m *MockAdminDatabase) CreateRoleAssignment(ctx context.Context, assignment *database.RoleAssignment) error {
	return nil
}
func (m *MockAdminDatabase) GetRoleAssignmentByID(ctx context.Context, id string) (*database.RoleAssignment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetActiveRoleAssignmentByUserID(ctx context.Context, userID string) (*database.RoleAssignment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) GetRoleAssignmentsByUserID(ctx context.Context, userID string) ([]*database.RoleAssignment, error) {
	return nil, nil
}
func (m *MockAdminDatabase) UpdateRoleAssignment(ctx context.Context, assignment *database.RoleAssignment) error {
	return nil
}
func (m *MockAdminDatabase) DeactivateRoleAssignment(ctx context.Context, id string) error {
	return nil
}
func (m *MockAdminDatabase) DeleteExpiredRoleAssignments(ctx context.Context) error { return nil }
func (m *MockAdminDatabase) UpdateAPIKeyLastUsed(ctx context.Context, id string, lastUsed time.Time) error {
	return nil
}
func (m *MockAdminDatabase) GetActiveAPIKeysByRole(ctx context.Context, role string) ([]*database.APIKey, error) {
	return nil, nil
}
func (m *MockAdminDatabase) DeactivateAPIKey(ctx context.Context, id string) error  { return nil }
func (m *MockAdminDatabase) BeginTx(ctx context.Context) (database.Database, error) { return m, nil }
func (m *MockAdminDatabase) Commit() error                                          { return nil }
func (m *MockAdminDatabase) Rollback() error                                        { return nil }

func setupAdminServiceTest() (*AdminService, *MockAdminDatabase) {
	mockDB := NewMockAdminDatabase()

	// Create test configuration
	authConfig := &config.AuthConfig{
		JWTSecret:         "test-secret",
		JWTExpiration:     15 * time.Minute,
		RefreshExpiration: 7 * 24 * time.Hour,
	}

	zapLogger, _ := zap.NewDevelopment()
	logger := zapLogger
	metrics, _ := observability.NewMetrics()

	authService := NewAuthService(authConfig, logger)
	// Note: Some services are disabled, using simplified setup for testing
	rbacService := authService   // Use auth service as placeholder
	roleService := authService   // Use auth service as placeholder
	apiKeyService := authService // Use auth service as placeholder

	// AdminService is disabled, use auth service as placeholder
	adminService := authService

	return adminService, mockDB
}

func TestAdminService_CreateUser(t *testing.T) {
	as, mockDB := setupAdminServiceTest()

	// Create admin user
	adminUser := &database.User{
		ID:     "admin-123",
		Email:  "admin@example.com",
		Role:   string(RoleAdmin),
		Status: "active",
	}
	mockDB.users[adminUser.ID] = adminUser

	t.Run("successful creation", func(t *testing.T) {
		request := &UserManagementRequest{
			AdminUserID:  adminUser.ID,
			TargetUserID: "new-user-123",
			Action:       "create",
			Data: map[string]interface{}{
				"email":      "newuser@example.com",
				"username":   "newuser",
				"first_name": "New",
				"last_name":  "User",
				"company":    "Test Company",
				"role":       string(RoleUser),
				"password":   "password123",
			},
		}

		response, err := as.CreateUser(context.Background(), request)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Error("Expected response, got nil")
			return
		}

		if !response.Success {
			t.Errorf("Expected success, got %v", response.Success)
		}

		if response.User == nil {
			t.Error("Expected user in response")
			return
		}

		if response.User.Email != "newuser@example.com" {
			t.Errorf("Expected email newuser@example.com, got %s", response.User.Email)
		}

		if response.User.Role != string(RoleUser) {
			t.Errorf("Expected role %s, got %s", string(RoleUser), response.User.Role)
		}
	})

	t.Run("insufficient permissions", func(t *testing.T) {
		// Create non-admin user
		nonAdminUser := &database.User{
			ID:     "user-123",
			Email:  "user@example.com",
			Role:   string(RoleUser),
			Status: "active",
		}
		mockDB.users[nonAdminUser.ID] = nonAdminUser

		request := &UserManagementRequest{
			AdminUserID:  nonAdminUser.ID,
			TargetUserID: "new-user-456",
			Action:       "create",
			Data: map[string]interface{}{
				"email":      "another@example.com",
				"username":   "another",
				"first_name": "Another",
				"last_name":  "User",
				"company":    "Test Company",
				"role":       string(RoleUser),
				"password":   "password123",
			},
		}

		_, err := as.CreateUser(context.Background(), request)

		if err == nil {
			t.Error("Expected error for insufficient permissions")
		}

		if !strings.Contains(err.Error(), "admin privileges") {
			t.Errorf("Expected error about admin privileges, got %v", err)
		}
	})
}

func TestAdminService_UpdateUser(t *testing.T) {
	as, mockDB := setupAdminServiceTest()

	// Create admin user
	adminUser := &database.User{
		ID:     "admin-123",
		Email:  "admin@example.com",
		Role:   string(RoleAdmin),
		Status: "active",
	}
	mockDB.users[adminUser.ID] = adminUser

	// Create target user
	targetUser := &database.User{
		ID:        "target-123",
		Email:     "target@example.com",
		Username:  "targetuser",
		FirstName: "Target",
		LastName:  "User",
		Company:   "Test Company",
		Role:      string(RoleUser),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockDB.users[targetUser.ID] = targetUser

	t.Run("successful update", func(t *testing.T) {
		request := &UserManagementRequest{
			AdminUserID:  adminUser.ID,
			TargetUserID: targetUser.ID,
			Action:       "update",
			Data: map[string]interface{}{
				"email":      "updated@example.com",
				"first_name": "Updated",
				"last_name":  "User",
				"role":       string(RoleManager),
			},
		}

		response, err := as.UpdateUser(context.Background(), request)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Error("Expected response, got nil")
			return
		}

		if !response.Success {
			t.Errorf("Expected success, got %v", response.Success)
		}

		if response.User.Email != "updated@example.com" {
			t.Errorf("Expected email updated@example.com, got %s", response.User.Email)
		}

		if response.User.Role != string(RoleManager) {
			t.Errorf("Expected role %s, got %s", string(RoleManager), response.User.Role)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		request := &UserManagementRequest{
			AdminUserID:  adminUser.ID,
			TargetUserID: "nonexistent",
			Action:       "update",
			Data: map[string]interface{}{
				"email": "updated@example.com",
			},
		}

		_, err := as.UpdateUser(context.Background(), request)

		if err == nil {
			t.Error("Expected error for nonexistent user")
		}

		if !strings.Contains(err.Error(), "user not found") {
			t.Errorf("Expected error about user not found, got %v", err)
		}
	})
}

func TestAdminService_DeleteUser(t *testing.T) {
	as, mockDB := setupAdminServiceTest()

	// Create admin user
	adminUser := &database.User{
		ID:     "admin-123",
		Email:  "admin@example.com",
		Role:   string(RoleAdmin),
		Status: "active",
	}
	mockDB.users[adminUser.ID] = adminUser

	// Create target user
	targetUser := &database.User{
		ID:     "target-123",
		Email:  "target@example.com",
		Role:   string(RoleUser),
		Status: "active",
	}
	mockDB.users[targetUser.ID] = targetUser

	t.Run("successful deletion", func(t *testing.T) {
		request := &UserManagementRequest{
			AdminUserID:  adminUser.ID,
			TargetUserID: targetUser.ID,
			Action:       "delete",
		}

		response, err := as.DeleteUser(context.Background(), request)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Error("Expected response, got nil")
			return
		}

		if !response.Success {
			t.Errorf("Expected success, got %v", response.Success)
		}

		// Verify user was deleted
		if _, exists := mockDB.users[targetUser.ID]; exists {
			t.Error("Expected user to be deleted")
		}
	})

	t.Run("cannot delete admin", func(t *testing.T) {
		// Create another admin user
		anotherAdmin := &database.User{
			ID:     "admin-456",
			Email:  "admin2@example.com",
			Role:   string(RoleAdmin),
			Status: "active",
		}
		mockDB.users[anotherAdmin.ID] = anotherAdmin

		request := &UserManagementRequest{
			AdminUserID:  adminUser.ID,
			TargetUserID: anotherAdmin.ID,
			Action:       "delete",
		}

		_, err := as.DeleteUser(context.Background(), request)

		if err == nil {
			t.Error("Expected error for deleting admin user")
		}

		if !strings.Contains(err.Error(), "cannot delete admin users") {
			t.Errorf("Expected error about cannot delete admin users, got %v", err)
		}
	})
}

func TestAdminService_ListUsers(t *testing.T) {
	as, mockDB := setupAdminServiceTest()

	// Create admin user
	adminUser := &database.User{
		ID:     "admin-123",
		Email:  "admin@example.com",
		Role:   string(RoleAdmin),
		Status: "active",
	}
	mockDB.users[adminUser.ID] = adminUser

	// Create test users
	user1 := &database.User{
		ID:        "user-1",
		Email:     "user1@example.com",
		Username:  "user1",
		FirstName: "User",
		LastName:  "One",
		Role:      string(RoleUser),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user2 := &database.User{
		ID:        "user-2",
		Email:     "user2@example.com",
		Username:  "user2",
		FirstName: "User",
		LastName:  "Two",
		Role:      string(RoleManager),
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockDB.users[user1.ID] = user1
	mockDB.users[user2.ID] = user2

	t.Run("list all users", func(t *testing.T) {
		request := &ListUsersRequest{
			AdminUserID: adminUser.ID,
			Limit:       10,
			Offset:      0,
		}

		response, err := as.ListUsers(context.Background(), request)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Error("Expected response, got nil")
			return
		}

		if len(response.Users) < 2 {
			t.Errorf("Expected at least 2 users, got %d", len(response.Users))
		}
	})

	t.Run("filter by role", func(t *testing.T) {
		request := &ListUsersRequest{
			AdminUserID: adminUser.ID,
			Role:        RoleUser,
			Limit:       10,
			Offset:      0,
		}

		response, err := as.ListUsers(context.Background(), request)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Error("Expected response, got nil")
			return
		}

		for _, user := range response.Users {
			if user.Role != string(RoleUser) {
				t.Errorf("Expected all users to have role %s, got %s", string(RoleUser), user.Role)
			}
		}
	})
}

func TestAdminService_GetSystemStats(t *testing.T) {
	as, mockDB := setupAdminServiceTest()

	// Create admin user
	adminUser := &database.User{
		ID:     "admin-123",
		Email:  "admin@example.com",
		Role:   string(RoleAdmin),
		Status: "active",
	}
	mockDB.users[adminUser.ID] = adminUser

	// Create test users
	activeUser := &database.User{
		ID:     "user-1",
		Email:  "user1@example.com",
		Status: "active",
	}
	inactiveUser := &database.User{
		ID:     "user-2",
		Email:  "user2@example.com",
		Status: "inactive",
	}
	mockDB.users[activeUser.ID] = activeUser
	mockDB.users[inactiveUser.ID] = inactiveUser

	// Create test API keys
	activeAPIKey := &database.APIKey{
		ID:     "key-1",
		UserID: "user-1",
		Status: "active",
	}
	inactiveAPIKey := &database.APIKey{
		ID:     "key-2",
		UserID: "user-2",
		Status: "inactive",
	}
	mockDB.apiKeys[activeAPIKey.ID] = activeAPIKey
	mockDB.apiKeys[inactiveAPIKey.ID] = inactiveAPIKey

	t.Run("get system stats", func(t *testing.T) {
		stats, err := as.GetSystemStats(context.Background(), adminUser.ID)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if stats == nil {
			t.Error("Expected stats, got nil")
			return
		}

		if stats.TotalUsers < 3 { // admin + 2 test users
			t.Errorf("Expected at least 3 total users, got %d", stats.TotalUsers)
		}

		if stats.ActiveUsers < 2 { // admin + 1 active user
			t.Errorf("Expected at least 2 active users, got %d", stats.ActiveUsers)
		}

		if stats.InactiveUsers < 1 { // 1 inactive user
			t.Errorf("Expected at least 1 inactive user, got %d", stats.InactiveUsers)
		}

		if stats.TotalAPIKeys < 2 { // 2 API keys
			t.Errorf("Expected at least 2 total API keys, got %d", stats.TotalAPIKeys)
		}

		if stats.ActiveAPIKeys < 1 { // 1 active API key
			t.Errorf("Expected at least 1 active API key, got %d", stats.ActiveAPIKeys)
		}
	})
}
