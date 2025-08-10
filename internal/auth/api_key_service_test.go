package auth

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// MockAPIKeyDatabase provides a mock implementation for testing
type MockAPIKeyDatabase struct {
	users   map[string]*database.User
	apiKeys map[string]*database.APIKey
}

func NewMockAPIKeyDatabase() *MockAPIKeyDatabase {
	return &MockAPIKeyDatabase{
		users:   make(map[string]*database.User),
		apiKeys: make(map[string]*database.APIKey),
	}
}

// Implement required interface methods
func (m *MockAPIKeyDatabase) Connect(ctx context.Context) error { return nil }
func (m *MockAPIKeyDatabase) Close() error                      { return nil }
func (m *MockAPIKeyDatabase) Ping(ctx context.Context) error    { return nil }
func (m *MockAPIKeyDatabase) CreateUser(ctx context.Context, user *database.User) error {
	m.users[user.ID] = user
	return nil
}
func (m *MockAPIKeyDatabase) GetUserByID(ctx context.Context, id string) (*database.User, error) {
	if user, exists := m.users[id]; exists {
		return user, nil
	}
	return nil, database.ErrUserNotFound
}
func (m *MockAPIKeyDatabase) GetUserByEmail(ctx context.Context, email string) (*database.User, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) UpdateUser(ctx context.Context, user *database.User) error { return nil }
func (m *MockAPIKeyDatabase) DeleteUser(ctx context.Context, id string) error {
	delete(m.users, id)
	return nil
}
func (m *MockAPIKeyDatabase) ListUsers(ctx context.Context, limit, offset int) ([]*database.User, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) CreateEmailVerificationToken(ctx context.Context, token *database.EmailVerificationToken) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetEmailVerificationToken(ctx context.Context, token string) (*database.EmailVerificationToken, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) MarkEmailVerificationTokenUsed(ctx context.Context, token string) error {
	return nil
}
func (m *MockAPIKeyDatabase) DeleteExpiredEmailVerificationTokens(ctx context.Context) error {
	return nil
}
func (m *MockAPIKeyDatabase) CreatePasswordResetToken(ctx context.Context, token *database.PasswordResetToken) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetPasswordResetToken(ctx context.Context, token string) (*database.PasswordResetToken, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) MarkPasswordResetTokenUsed(ctx context.Context, token string) error {
	return nil
}
func (m *MockAPIKeyDatabase) DeleteExpiredPasswordResetTokens(ctx context.Context) error { return nil }
func (m *MockAPIKeyDatabase) CreateTokenBlacklist(ctx context.Context, blacklist *database.TokenBlacklist) error {
	return nil
}
func (m *MockAPIKeyDatabase) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	return false, nil
}
func (m *MockAPIKeyDatabase) DeleteExpiredTokenBlacklist(ctx context.Context) error { return nil }
func (m *MockAPIKeyDatabase) CreateBusiness(ctx context.Context, business *database.Business) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetBusinessByID(ctx context.Context, id string) (*database.Business, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetBusinessByRegistrationNumber(ctx context.Context, regNumber string) (*database.Business, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) UpdateBusiness(ctx context.Context, business *database.Business) error {
	return nil
}
func (m *MockAPIKeyDatabase) DeleteBusiness(ctx context.Context, id string) error { return nil }
func (m *MockAPIKeyDatabase) ListBusinesses(ctx context.Context, limit, offset int) ([]*database.Business, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) SearchBusinesses(ctx context.Context, query string, limit, offset int) ([]*database.Business, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) CreateBusinessClassification(ctx context.Context, classification *database.BusinessClassification) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetBusinessClassificationByID(ctx context.Context, id string) (*database.BusinessClassification, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetBusinessClassificationsByBusinessID(ctx context.Context, businessID string) ([]*database.BusinessClassification, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) UpdateBusinessClassification(ctx context.Context, classification *database.BusinessClassification) error {
	return nil
}
func (m *MockAPIKeyDatabase) DeleteBusinessClassification(ctx context.Context, id string) error {
	return nil
}
func (m *MockAPIKeyDatabase) CreateRiskAssessment(ctx context.Context, assessment *database.RiskAssessment) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetRiskAssessmentByID(ctx context.Context, id string) (*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetRiskAssessmentsByBusinessID(ctx context.Context, businessID string) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) UpdateRiskAssessment(ctx context.Context, assessment *database.RiskAssessment) error {
	return nil
}
func (m *MockAPIKeyDatabase) DeleteRiskAssessment(ctx context.Context, id string) error { return nil }
func (m *MockAPIKeyDatabase) GetRiskAssessmentHistory(ctx context.Context, businessID string, limit, offset int) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetRiskAssessmentHistoryByDateRange(ctx context.Context, businessID string, startDate, endDate time.Time) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetLatestRiskAssessment(ctx context.Context, businessID string) (*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetRiskAssessmentTrends(ctx context.Context, businessID string, days int) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetRiskAssessmentsByLevel(ctx context.Context, businessID string, riskLevel string) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetRiskAssessmentsByScoreRange(ctx context.Context, businessID string, minScore, maxScore float64) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetRiskAssessmentStatistics(ctx context.Context, businessID string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) CreateComplianceCheck(ctx context.Context, check *database.ComplianceCheck) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetComplianceCheckByID(ctx context.Context, id string) (*database.ComplianceCheck, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetComplianceChecksByBusinessID(ctx context.Context, businessID string) ([]*database.ComplianceCheck, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) UpdateComplianceCheck(ctx context.Context, check *database.ComplianceCheck) error {
	return nil
}
func (m *MockAPIKeyDatabase) DeleteComplianceCheck(ctx context.Context, id string) error { return nil }
func (m *MockAPIKeyDatabase) CreateAPIKey(ctx context.Context, apiKey *database.APIKey) error {
	m.apiKeys[apiKey.ID] = apiKey
	return nil
}
func (m *MockAPIKeyDatabase) GetAPIKeyByID(ctx context.Context, id string) (*database.APIKey, error) {
	if apiKey, exists := m.apiKeys[id]; exists {
		return apiKey, nil
	}
	return nil, database.ErrAPIKeyNotFound
}
func (m *MockAPIKeyDatabase) GetAPIKeyByHash(ctx context.Context, keyHash string) (*database.APIKey, error) {
	for _, apiKey := range m.apiKeys {
		if apiKey.KeyHash == keyHash {
			return apiKey, nil
		}
	}
	return nil, database.ErrAPIKeyNotFound
}
func (m *MockAPIKeyDatabase) UpdateAPIKey(ctx context.Context, apiKey *database.APIKey) error {
	m.apiKeys[apiKey.ID] = apiKey
	return nil
}
func (m *MockAPIKeyDatabase) DeleteAPIKey(ctx context.Context, id string) error {
	delete(m.apiKeys, id)
	return nil
}
func (m *MockAPIKeyDatabase) ListAPIKeysByUserID(ctx context.Context, userID string) ([]*database.APIKey, error) {
	var keys []*database.APIKey
	for _, apiKey := range m.apiKeys {
		if apiKey.UserID == userID {
			keys = append(keys, apiKey)
		}
	}
	return keys, nil
}
func (m *MockAPIKeyDatabase) CreateAuditLog(ctx context.Context, log *database.AuditLog) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetAuditLogsByUserID(ctx context.Context, userID string, limit, offset int) ([]*database.AuditLog, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetAuditLogsByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*database.AuditLog, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) CreateExternalServiceCall(ctx context.Context, call *database.ExternalServiceCall) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetExternalServiceCallsByUserID(ctx context.Context, userID string, limit, offset int) ([]*database.ExternalServiceCall, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetExternalServiceCallsByService(ctx context.Context, serviceName string, limit, offset int) ([]*database.ExternalServiceCall, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) CreateWebhook(ctx context.Context, webhook *database.Webhook) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetWebhookByID(ctx context.Context, id string) (*database.Webhook, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetWebhooksByUserID(ctx context.Context, userID string) ([]*database.Webhook, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) UpdateWebhook(ctx context.Context, webhook *database.Webhook) error {
	return nil
}
func (m *MockAPIKeyDatabase) DeleteWebhook(ctx context.Context, id string) error { return nil }
func (m *MockAPIKeyDatabase) CreateWebhookEvent(ctx context.Context, event *database.WebhookEvent) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetWebhookEventByID(ctx context.Context, id string) (*database.WebhookEvent, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetWebhookEventsByWebhookID(ctx context.Context, webhookID string, limit, offset int) ([]*database.WebhookEvent, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) UpdateWebhookEvent(ctx context.Context, event *database.WebhookEvent) error {
	return nil
}
func (m *MockAPIKeyDatabase) DeleteWebhookEvent(ctx context.Context, id string) error { return nil }
func (m *MockAPIKeyDatabase) CreateRoleAssignment(ctx context.Context, assignment *database.RoleAssignment) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetRoleAssignmentByID(ctx context.Context, id string) (*database.RoleAssignment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetActiveRoleAssignmentByUserID(ctx context.Context, userID string) (*database.RoleAssignment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) GetRoleAssignmentsByUserID(ctx context.Context, userID string) ([]*database.RoleAssignment, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) UpdateRoleAssignment(ctx context.Context, assignment *database.RoleAssignment) error {
	return nil
}
func (m *MockAPIKeyDatabase) DeactivateRoleAssignment(ctx context.Context, id string) error {
	return nil
}
func (m *MockAPIKeyDatabase) DeleteExpiredRoleAssignments(ctx context.Context) error { return nil }
func (m *MockAPIKeyDatabase) UpdateAPIKeyLastUsed(ctx context.Context, id string, lastUsed time.Time) error {
	return nil
}
func (m *MockAPIKeyDatabase) GetActiveAPIKeysByRole(ctx context.Context, role string) ([]*database.APIKey, error) {
	return nil, nil
}
func (m *MockAPIKeyDatabase) DeactivateAPIKey(ctx context.Context, id string) error  { return nil }
func (m *MockAPIKeyDatabase) BeginTx(ctx context.Context) (database.Database, error) { return m, nil }
func (m *MockAPIKeyDatabase) Commit() error                                          { return nil }
func (m *MockAPIKeyDatabase) Rollback() error                                        { return nil }

func setupAPIKeyServiceTest() (*APIKeyService, *MockAPIKeyDatabase) {
	mockDB := NewMockAPIKeyDatabase()
	loggerConfig := &config.ObservabilityConfig{
		LogLevel: "debug",
	}
	logger := observability.NewLogger(loggerConfig)

	apiKeyService := NewAPIKeyService(mockDB, logger)

	return apiKeyService, mockDB
}

func TestAPIKeyService_CreateAPIKey(t *testing.T) {
	aks, mockDB := setupAPIKeyServiceTest()

	// Create a test user
	testUser := &database.User{
		ID:    "user1",
		Email: "user@test.com",
		Role:  string(RoleAdmin),
	}
	mockDB.users[testUser.ID] = testUser

	t.Run("successful creation", func(t *testing.T) {
		request := &CreateAPIKeyRequest{
			Name:        "Test API Key",
			UserID:      "user1",
			Role:        RoleUser,
			Permissions: []string{"classify:business", "risk:view"},
			Description: "Test API key for integration",
		}

		response, err := aks.CreateAPIKey(context.Background(), request)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if response == nil {
			t.Error("Expected response, got nil")
			return
		}

		if response.Name != request.Name {
			t.Errorf("Expected name %s, got %s", request.Name, response.Name)
		}

		if response.UserID != request.UserID {
			t.Errorf("Expected user ID %s, got %s", request.UserID, response.UserID)
		}

		if response.Role != request.Role {
			t.Errorf("Expected role %s, got %s", request.Role, response.Role)
		}

		if response.Key == "" {
			t.Error("Expected API key to be generated")
		}

		if !strings.HasPrefix(response.Key, "kyb_") {
			t.Errorf("Expected API key to start with 'kyb_', got %s", response.Key)
		}
	})

	t.Run("invalid role", func(t *testing.T) {
		request := &CreateAPIKeyRequest{
			Name:   "Test API Key",
			UserID: "user1",
			Role:   "invalid_role",
		}

		_, err := aks.CreateAPIKey(context.Background(), request)

		if err == nil {
			t.Error("Expected error for invalid role")
		}

		if !strings.Contains(err.Error(), "invalid role") {
			t.Errorf("Expected error about invalid role, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		request := &CreateAPIKeyRequest{
			Name:   "Test API Key",
			UserID: "nonexistent",
			Role:   RoleUser,
		}

		_, err := aks.CreateAPIKey(context.Background(), request)

		if err == nil {
			t.Error("Expected error for nonexistent user")
		}

		if !strings.Contains(err.Error(), "user not found") {
			t.Errorf("Expected error about user not found, got %v", err)
		}
	})
}

func TestAPIKeyService_ValidateAPIKey(t *testing.T) {
	aks, mockDB := setupAPIKeyServiceTest()

	// Create a test API key
	testAPIKey := &database.APIKey{
		ID:          "key1",
		Name:        "Test API Key",
		KeyHash:     aks.hashAPIKey("kyb_testkey123456789012345678901234567890"),
		UserID:      "user1",
		Role:        string(RoleUser),
		Permissions: "classify:business,risk:view",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockDB.apiKeys[testAPIKey.ID] = testAPIKey

	t.Run("valid API key", func(t *testing.T) {
		context, err := aks.ValidateAPIKey(context.Background(), "kyb_testkey123456789012345678901234567890")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if context == nil {
			t.Error("Expected context, got nil")
			return
		}

		if context.APIKeyID != testAPIKey.ID {
			t.Errorf("Expected API key ID %s, got %s", testAPIKey.ID, context.APIKeyID)
		}

		if context.UserID != testAPIKey.UserID {
			t.Errorf("Expected user ID %s, got %s", testAPIKey.UserID, context.UserID)
		}

		if context.Role != Role(testAPIKey.Role) {
			t.Errorf("Expected role %s, got %s", testAPIKey.Role, context.Role)
		}
	})

	t.Run("invalid API key", func(t *testing.T) {
		_, err := aks.ValidateAPIKey(context.Background(), "invalid_key")

		if err == nil {
			t.Error("Expected error for invalid API key")
		}

		if !strings.Contains(err.Error(), "invalid API key") {
			t.Errorf("Expected error about invalid API key, got %v", err)
		}
	})

	t.Run("inactive API key", func(t *testing.T) {
		// Create an inactive API key
		inactiveAPIKey := &database.APIKey{
			ID:          "key2",
			Name:        "Inactive API Key",
			KeyHash:     aks.hashAPIKey("kyb_inactivekey123456789012345678901234567890"),
			UserID:      "user1",
			Role:        string(RoleUser),
			Permissions: "classify:business",
			Status:      "inactive",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		mockDB.apiKeys[inactiveAPIKey.ID] = inactiveAPIKey

		_, err := aks.ValidateAPIKey(context.Background(), "kyb_inactivekey123456789012345678901234567890")

		if err == nil {
			t.Error("Expected error for inactive API key")
		}

		if !strings.Contains(err.Error(), "inactive") {
			t.Errorf("Expected error about inactive API key, got %v", err)
		}
	})

	t.Run("expired API key", func(t *testing.T) {
		// Create an expired API key
		expiredTime := time.Now().Add(-1 * time.Hour)
		expiredAPIKey := &database.APIKey{
			ID:          "key3",
			Name:        "Expired API Key",
			KeyHash:     aks.hashAPIKey("kyb_expiredkey123456789012345678901234567890"),
			UserID:      "user1",
			Role:        string(RoleUser),
			Permissions: "classify:business",
			Status:      "active",
			ExpiresAt:   &expiredTime,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		mockDB.apiKeys[expiredAPIKey.ID] = expiredAPIKey

		_, err := aks.ValidateAPIKey(context.Background(), "kyb_expiredkey123456789012345678901234567890")

		if err == nil {
			t.Error("Expected error for expired API key")
		}

		if !strings.Contains(err.Error(), "expired") {
			t.Errorf("Expected error about expired API key, got %v", err)
		}
	})
}

func TestAPIKeyService_RevokeAPIKey(t *testing.T) {
	aks, mockDB := setupAPIKeyServiceTest()

	// Create a test API key
	testAPIKey := &database.APIKey{
		ID:          "key1",
		Name:        "Test API Key",
		KeyHash:     "test_hash",
		UserID:      "user1",
		Role:        string(RoleUser),
		Permissions: "classify:business",
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	mockDB.apiKeys[testAPIKey.ID] = testAPIKey

	t.Run("successful revocation", func(t *testing.T) {
		request := &RevokeAPIKeyRequest{
			APIKeyID: "key1",
			UserID:   "user1",
			Reason:   "Security concern",
		}

		err := aks.RevokeAPIKey(context.Background(), request)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("unauthorized revocation", func(t *testing.T) {
		request := &RevokeAPIKeyRequest{
			APIKeyID: "key1",
			UserID:   "user2", // Different user
			Reason:   "Security concern",
		}

		err := aks.RevokeAPIKey(context.Background(), request)

		if err == nil {
			t.Error("Expected error for unauthorized revocation")
		}

		if !strings.Contains(err.Error(), "insufficient permissions") {
			t.Errorf("Expected error about insufficient permissions, got %v", err)
		}
	})

	t.Run("API key not found", func(t *testing.T) {
		request := &RevokeAPIKeyRequest{
			APIKeyID: "nonexistent",
			UserID:   "user1",
			Reason:   "Security concern",
		}

		err := aks.RevokeAPIKey(context.Background(), request)

		if err == nil {
			t.Error("Expected error for nonexistent API key")
		}

		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected error about API key not found, got %v", err)
		}
	})
}

func TestAPIKeyService_GenerateAPIKey(t *testing.T) {
	aks, _ := setupAPIKeyServiceTest()

	key, hash, err := aks.generateAPIKey()

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if key == "" {
		t.Error("Expected generated key")
	}

	if hash == "" {
		t.Error("Expected generated hash")
	}

	if !strings.HasPrefix(key, "kyb_") {
		t.Errorf("Expected key to start with 'kyb_', got %s", key)
	}

	// Test that the hash matches the key
	expectedHash := aks.hashAPIKey(key)
	if hash != expectedHash {
		t.Errorf("Expected hash %s, got %s", expectedHash, hash)
	}
}

func TestAPIKeyService_HashAPIKey(t *testing.T) {
	aks, _ := setupAPIKeyServiceTest()

	testKey := "kyb_testkey123456789012345678901234567890"
	hash1 := aks.hashAPIKey(testKey)
	hash2 := aks.hashAPIKey(testKey)

	if hash1 != hash2 {
		t.Errorf("Expected consistent hashing, got %s and %s", hash1, hash2)
	}

	if len(hash1) != 64 { // SHA-256 hash is 64 characters in hex
		t.Errorf("Expected 64-character hash, got %d characters", len(hash1))
	}
}
