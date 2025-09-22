package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"kyb-platform/internal/database"
)

// MockDatabase implements the database.Database interface for testing
type MockDatabase struct {
	businesses map[string]*database.Business
	auditLogs  []*database.AuditLog
	users      map[string]*database.User
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		businesses: make(map[string]*database.Business),
		auditLogs:  make([]*database.AuditLog, 0),
		users:      make(map[string]*database.User),
	}
}

// Database interface methods
func (m *MockDatabase) Connect(ctx context.Context) error { return nil }
func (m *MockDatabase) Close() error                      { return nil }
func (m *MockDatabase) Ping(ctx context.Context) error    { return nil }

// Business management
func (m *MockDatabase) CreateBusiness(ctx context.Context, business *database.Business) error {
	m.businesses[business.ID] = business
	return nil
}

func (m *MockDatabase) GetBusinessByID(ctx context.Context, id string) (*database.Business, error) {
	if business, exists := m.businesses[id]; exists {
		return business, nil
	}
	return nil, database.ErrUserNotFound
}

func (m *MockDatabase) GetBusinessByRegistrationNumber(ctx context.Context, regNumber string) (*database.Business, error) {
	for _, business := range m.businesses {
		if business.RegistrationNumber == regNumber {
			return business, nil
		}
	}
	return nil, database.ErrUserNotFound
}

func (m *MockDatabase) UpdateBusiness(ctx context.Context, business *database.Business) error {
	if _, exists := m.businesses[business.ID]; !exists {
		return database.ErrUserNotFound
	}
	m.businesses[business.ID] = business
	return nil
}

func (m *MockDatabase) DeleteBusiness(ctx context.Context, id string) error {
	if _, exists := m.businesses[id]; !exists {
		return database.ErrUserNotFound
	}
	delete(m.businesses, id)
	return nil
}

func (m *MockDatabase) ListBusinesses(ctx context.Context, limit, offset int) ([]*database.Business, error) {
	businesses := make([]*database.Business, 0, len(m.businesses))
	for _, business := range m.businesses {
		businesses = append(businesses, business)
	}

	// Simple pagination
	start := offset
	end := start + limit
	if start >= len(businesses) {
		return []*database.Business{}, nil
	}
	if end > len(businesses) {
		end = len(businesses)
	}

	return businesses[start:end], nil
}

func (m *MockDatabase) SearchBusinesses(ctx context.Context, query string, limit, offset int) ([]*database.Business, error) {
	// Simple search implementation
	results := make([]*database.Business, 0)
	for _, business := range m.businesses {
		if contains(business.Name, query) || contains(business.Industry, query) {
			results = append(results, business)
		}
	}

	// Simple pagination
	start := offset
	end := start + limit
	if start >= len(results) {
		return []*database.Business{}, nil
	}
	if end > len(results) {
		end = len(results)
	}

	return results[start:end], nil
}

// Audit log management
func (m *MockDatabase) CreateAuditLog(ctx context.Context, log *database.AuditLog) error {
	m.auditLogs = append(m.auditLogs, log)
	return nil
}

func (m *MockDatabase) GetAuditLogsByUserID(ctx context.Context, userID string, limit, offset int) ([]*database.AuditLog, error) {
	results := make([]*database.AuditLog, 0)
	for _, log := range m.auditLogs {
		if log.UserID == userID {
			results = append(results, log)
		}
	}
	return results, nil
}

func (m *MockDatabase) GetAuditLogsByResource(ctx context.Context, resourceType, resourceID string, limit, offset int) ([]*database.AuditLog, error) {
	results := make([]*database.AuditLog, 0)
	for _, log := range m.auditLogs {
		if log.ResourceType == resourceType && log.ResourceID == resourceID {
			results = append(results, log)
		}
	}
	return results, nil
}

// Unimplemented methods (not needed for these tests)
func (m *MockDatabase) CreateUser(ctx context.Context, user *database.User) error { return nil }
func (m *MockDatabase) GetUserByID(ctx context.Context, id string) (*database.User, error) {
	return nil, nil
}
func (m *MockDatabase) GetUserByEmail(ctx context.Context, email string) (*database.User, error) {
	return nil, nil
}
func (m *MockDatabase) UpdateUser(ctx context.Context, user *database.User) error { return nil }
func (m *MockDatabase) DeleteUser(ctx context.Context, id string) error           { return nil }
func (m *MockDatabase) ListUsers(ctx context.Context, limit, offset int) ([]*database.User, error) {
	return nil, nil
}
func (m *MockDatabase) CreateEmailVerificationToken(ctx context.Context, token *database.EmailVerificationToken) error {
	return nil
}
func (m *MockDatabase) GetEmailVerificationToken(ctx context.Context, token string) (*database.EmailVerificationToken, error) {
	return nil, nil
}
func (m *MockDatabase) MarkEmailVerificationTokenUsed(ctx context.Context, token string) error {
	return nil
}
func (m *MockDatabase) DeleteExpiredEmailVerificationTokens(ctx context.Context) error { return nil }
func (m *MockDatabase) CreatePasswordResetToken(ctx context.Context, token *database.PasswordResetToken) error {
	return nil
}
func (m *MockDatabase) GetPasswordResetToken(ctx context.Context, token string) (*database.PasswordResetToken, error) {
	return nil, nil
}
func (m *MockDatabase) MarkPasswordResetTokenUsed(ctx context.Context, token string) error {
	return nil
}
func (m *MockDatabase) DeleteExpiredPasswordResetTokens(ctx context.Context) error { return nil }
func (m *MockDatabase) CreateTokenBlacklist(ctx context.Context, blacklist *database.TokenBlacklist) error {
	return nil
}
func (m *MockDatabase) IsTokenBlacklisted(ctx context.Context, tokenID string) (bool, error) {
	return false, nil
}
func (m *MockDatabase) DeleteExpiredTokenBlacklist(ctx context.Context) error { return nil }
func (m *MockDatabase) CreateBusinessClassification(ctx context.Context, classification *database.BusinessClassification) error {
	return nil
}
func (m *MockDatabase) GetBusinessClassificationByID(ctx context.Context, id string) (*database.BusinessClassification, error) {
	return nil, nil
}
func (m *MockDatabase) GetBusinessClassificationsByBusinessID(ctx context.Context, businessID string) ([]*database.BusinessClassification, error) {
	return nil, nil
}
func (m *MockDatabase) UpdateBusinessClassification(ctx context.Context, classification *database.BusinessClassification) error {
	return nil
}
func (m *MockDatabase) DeleteBusinessClassification(ctx context.Context, id string) error { return nil }
func (m *MockDatabase) CreateRiskAssessment(ctx context.Context, assessment *database.RiskAssessment) error {
	return nil
}
func (m *MockDatabase) GetRiskAssessmentByID(ctx context.Context, id string) (*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockDatabase) GetRiskAssessmentsByBusinessID(ctx context.Context, businessID string) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockDatabase) UpdateRiskAssessment(ctx context.Context, assessment *database.RiskAssessment) error {
	return nil
}
func (m *MockDatabase) DeleteRiskAssessment(ctx context.Context, id string) error { return nil }
func (m *MockDatabase) GetRiskAssessmentHistory(ctx context.Context, businessID string, limit, offset int) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockDatabase) GetRiskAssessmentHistoryByDateRange(ctx context.Context, businessID string, startDate, endDate time.Time) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockDatabase) GetLatestRiskAssessment(ctx context.Context, businessID string) (*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockDatabase) GetRiskAssessmentTrends(ctx context.Context, businessID string, days int) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockDatabase) GetRiskAssessmentsByLevel(ctx context.Context, businessID string, riskLevel string) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockDatabase) GetRiskAssessmentsByScoreRange(ctx context.Context, businessID string, minScore, maxScore float64) ([]*database.RiskAssessment, error) {
	return nil, nil
}
func (m *MockDatabase) GetRiskAssessmentStatistics(ctx context.Context, businessID string) (map[string]interface{}, error) {
	return nil, nil
}
func (m *MockDatabase) CreateComplianceCheck(ctx context.Context, check *database.ComplianceCheck) error {
	return nil
}
func (m *MockDatabase) GetComplianceCheckByID(ctx context.Context, id string) (*database.ComplianceCheck, error) {
	return nil, nil
}
func (m *MockDatabase) GetComplianceChecksByBusinessID(ctx context.Context, businessID string) ([]*database.ComplianceCheck, error) {
	return nil, nil
}
func (m *MockDatabase) UpdateComplianceCheck(ctx context.Context, check *database.ComplianceCheck) error {
	return nil
}
func (m *MockDatabase) DeleteComplianceCheck(ctx context.Context, id string) error      { return nil }
func (m *MockDatabase) CreateAPIKey(ctx context.Context, apiKey *database.APIKey) error { return nil }
func (m *MockDatabase) GetAPIKeyByID(ctx context.Context, id string) (*database.APIKey, error) {
	return nil, nil
}
func (m *MockDatabase) GetAPIKeyByHash(ctx context.Context, keyHash string) (*database.APIKey, error) {
	return nil, nil
}
func (m *MockDatabase) UpdateAPIKey(ctx context.Context, apiKey *database.APIKey) error { return nil }
func (m *MockDatabase) DeleteAPIKey(ctx context.Context, id string) error               { return nil }
func (m *MockDatabase) ListAPIKeysByUserID(ctx context.Context, userID string) ([]*database.APIKey, error) {
	return nil, nil
}
func (m *MockDatabase) CreateExternalServiceCall(ctx context.Context, call *database.ExternalServiceCall) error {
	return nil
}
func (m *MockDatabase) GetExternalServiceCallsByUserID(ctx context.Context, userID string, limit, offset int) ([]*database.ExternalServiceCall, error) {
	return nil, nil
}
func (m *MockDatabase) GetExternalServiceCallsByService(ctx context.Context, serviceName string, limit, offset int) ([]*database.ExternalServiceCall, error) {
	return nil, nil
}
func (m *MockDatabase) CreateWebhook(ctx context.Context, webhook *database.Webhook) error {
	return nil
}
func (m *MockDatabase) GetWebhookByID(ctx context.Context, id string) (*database.Webhook, error) {
	return nil, nil
}
func (m *MockDatabase) GetWebhooksByUserID(ctx context.Context, userID string) ([]*database.Webhook, error) {
	return nil, nil
}
func (m *MockDatabase) UpdateWebhook(ctx context.Context, webhook *database.Webhook) error {
	return nil
}
func (m *MockDatabase) DeleteWebhook(ctx context.Context, id string) error { return nil }
func (m *MockDatabase) CreateWebhookEvent(ctx context.Context, event *database.WebhookEvent) error {
	return nil
}
func (m *MockDatabase) GetWebhookEventByID(ctx context.Context, id string) (*database.WebhookEvent, error) {
	return nil, nil
}
func (m *MockDatabase) GetWebhookEventsByWebhookID(ctx context.Context, webhookID string, limit, offset int) ([]*database.WebhookEvent, error) {
	return nil, nil
}
func (m *MockDatabase) UpdateWebhookEvent(ctx context.Context, event *database.WebhookEvent) error {
	return nil
}
func (m *MockDatabase) DeleteWebhookEvent(ctx context.Context, id string) error { return nil }
func (m *MockDatabase) CreateRoleAssignment(ctx context.Context, assignment *database.RoleAssignment) error {
	return nil
}
func (m *MockDatabase) GetRoleAssignmentByID(ctx context.Context, id string) (*database.RoleAssignment, error) {
	return nil, nil
}
func (m *MockDatabase) GetActiveRoleAssignmentByUserID(ctx context.Context, userID string) (*database.RoleAssignment, error) {
	return nil, nil
}
func (m *MockDatabase) GetRoleAssignmentsByUserID(ctx context.Context, userID string) ([]*database.RoleAssignment, error) {
	return nil, nil
}
func (m *MockDatabase) UpdateRoleAssignment(ctx context.Context, assignment *database.RoleAssignment) error {
	return nil
}
func (m *MockDatabase) DeactivateRoleAssignment(ctx context.Context, id string) error { return nil }
func (m *MockDatabase) DeleteExpiredRoleAssignments(ctx context.Context) error        { return nil }
func (m *MockDatabase) UpdateAPIKeyLastUsed(ctx context.Context, id string, lastUsed time.Time) error {
	return nil
}
func (m *MockDatabase) GetActiveAPIKeysByRole(ctx context.Context, role string) ([]*database.APIKey, error) {
	return nil, nil
}
func (m *MockDatabase) DeactivateAPIKey(ctx context.Context, id string) error  { return nil }
func (m *MockDatabase) BeginTx(ctx context.Context) (database.Database, error) { return m, nil }
func (m *MockDatabase) Commit() error                                          { return nil }
func (m *MockDatabase) Rollback() error                                        { return nil }

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Test helper functions
func createTestMerchant() *Merchant {
	return &Merchant{
		Name:               "Test Company",
		LegalName:          "Test Company LLC",
		RegistrationNumber: "REG123456",
		TaxID:              "TAX123456",
		Industry:           "Technology",
		IndustryCode:       "541511",
		BusinessType:       "LLC",
		EmployeeCount:      50,
		Address: database.Address{
			Street1:    "123 Test St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
		},
		ContactInfo: database.ContactInfo{
			Phone:   "+1-555-123-4567",
			Email:   "test@testcompany.com",
			Website: "https://testcompany.com",
		},
		PortfolioType:    PortfolioTypeProspective,
		RiskLevel:        RiskLevelMedium,
		ComplianceStatus: "pending",
		Status:           "active",
	}
}

func createTestService() (*MerchantPortfolioService, *MockDatabase) {
	mockDB := NewMockDatabase()
	logger := log.New(log.Writer(), "test: ", log.LstdFlags)
	service := NewMerchantPortfolioService(mockDB, logger)
	return service, mockDB
}

// Tests
func TestNewMerchantPortfolioService(t *testing.T) {
	mockDB := NewMockDatabase()
	service := NewMerchantPortfolioService(mockDB, nil)

	if service == nil {
		t.Fatal("Expected service to be created")
	}

	if service.db != mockDB {
		t.Error("Expected database to be set correctly")
	}

	if service.logger == nil {
		t.Error("Expected logger to be set (default logger)")
	}
}

func TestCreateMerchant(t *testing.T) {
	service, mockDB := createTestService()
	ctx := context.Background()
	userID := "user123"

	merchant := createTestMerchant()

	// Test successful creation
	created, err := service.CreateMerchant(ctx, merchant, userID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if created.ID == "" {
		t.Error("Expected merchant ID to be generated")
	}

	if created.CreatedBy != userID {
		t.Error("Expected CreatedBy to be set to userID")
	}

	if created.PortfolioType != PortfolioTypeProspective {
		t.Error("Expected default portfolio type to be set")
	}

	if created.RiskLevel != RiskLevelMedium {
		t.Error("Expected default risk level to be set")
	}

	// Verify merchant was stored in database
	if len(mockDB.businesses) != 1 {
		t.Error("Expected merchant to be stored in database")
	}

	// Verify audit log was created
	if len(mockDB.auditLogs) != 1 {
		t.Error("Expected audit log to be created")
	}
}

func TestCreateMerchantValidation(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Test nil merchant
	_, err := service.CreateMerchant(ctx, nil, userID)
	if err == nil {
		t.Error("Expected error for nil merchant")
	}

	// Test empty name
	merchant := createTestMerchant()
	merchant.Name = ""
	_, err = service.CreateMerchant(ctx, merchant, userID)
	if err == nil {
		t.Error("Expected error for empty merchant name")
	}

	// Test invalid portfolio type
	merchant = createTestMerchant()
	merchant.PortfolioType = "invalid"
	_, err = service.CreateMerchant(ctx, merchant, userID)
	if err == nil {
		t.Error("Expected error for invalid portfolio type")
	}

	// Test invalid risk level
	merchant = createTestMerchant()
	merchant.RiskLevel = "invalid"
	_, err = service.CreateMerchant(ctx, merchant, userID)
	if err == nil {
		t.Error("Expected error for invalid risk level")
	}
}

func TestGetMerchant(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()

	// Create a merchant first
	merchant := createTestMerchant()
	created, err := service.CreateMerchant(ctx, merchant, "user123")
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	// Test successful retrieval
	retrieved, err := service.GetMerchant(ctx, created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if retrieved.ID != created.ID {
		t.Error("Expected retrieved merchant ID to match created ID")
	}

	if retrieved.Name != created.Name {
		t.Error("Expected retrieved merchant name to match created name")
	}

	// Test merchant not found
	_, err = service.GetMerchant(ctx, "nonexistent")
	if !errors.Is(err, database.ErrMerchantNotFound) {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}
}

func TestUpdateMerchant(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Create a merchant first
	merchant := createTestMerchant()
	created, err := service.CreateMerchant(ctx, merchant, userID)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	// Update merchant
	created.Name = "Updated Company"
	created.Industry = "Updated Industry"

	updated, err := service.UpdateMerchant(ctx, created, userID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if updated.Name != "Updated Company" {
		t.Error("Expected merchant name to be updated")
	}

	if updated.Industry != "Updated Industry" {
		t.Error("Expected merchant industry to be updated")
	}

	// Test updating non-existent merchant
	nonExistent := createTestMerchant()
	nonExistent.ID = "nonexistent"
	_, err = service.UpdateMerchant(ctx, nonExistent, userID)
	if !errors.Is(err, database.ErrMerchantNotFound) {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}
}

func TestDeleteMerchant(t *testing.T) {
	service, mockDB := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Create a merchant first
	merchant := createTestMerchant()
	created, err := service.CreateMerchant(ctx, merchant, userID)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	// Delete merchant
	err = service.DeleteMerchant(ctx, created.ID, userID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify merchant was deleted
	if len(mockDB.businesses) != 0 {
		t.Error("Expected merchant to be deleted from database")
	}

	// Test deleting non-existent merchant
	err = service.DeleteMerchant(ctx, "nonexistent", userID)
	if !errors.Is(err, database.ErrMerchantNotFound) {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}
}

func TestUpdateMerchantPortfolioType(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Create a merchant first
	merchant := createTestMerchant()
	created, err := service.CreateMerchant(ctx, merchant, userID)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	// Test successful update
	err = service.UpdateMerchantPortfolioType(ctx, created.ID, PortfolioTypeOnboarded, userID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify update
	updated, err := service.GetMerchant(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated merchant: %v", err)
	}

	if updated.PortfolioType != PortfolioTypeOnboarded {
		t.Error("Expected portfolio type to be updated")
	}

	// Test invalid portfolio type
	err = service.UpdateMerchantPortfolioType(ctx, created.ID, "invalid", userID)
	if !errors.Is(err, ErrInvalidPortfolioType) {
		t.Errorf("Expected ErrInvalidPortfolioType, got: %v", err)
	}

	// Test non-existent merchant
	err = service.UpdateMerchantPortfolioType(ctx, "nonexistent", PortfolioTypeOnboarded, userID)
	if !errors.Is(err, database.ErrMerchantNotFound) {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}
}

func TestUpdateMerchantRiskLevel(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Create a merchant first
	merchant := createTestMerchant()
	created, err := service.CreateMerchant(ctx, merchant, userID)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	// Test successful update
	err = service.UpdateMerchantRiskLevel(ctx, created.ID, RiskLevelHigh, userID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Verify update
	updated, err := service.GetMerchant(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated merchant: %v", err)
	}

	if updated.RiskLevel != RiskLevelHigh {
		t.Error("Expected risk level to be updated")
	}

	// Test invalid risk level
	err = service.UpdateMerchantRiskLevel(ctx, created.ID, "invalid", userID)
	if !errors.Is(err, ErrInvalidRiskLevel) {
		t.Errorf("Expected ErrInvalidRiskLevel, got: %v", err)
	}

	// Test non-existent merchant
	err = service.UpdateMerchantRiskLevel(ctx, "nonexistent", RiskLevelHigh, userID)
	if !errors.Is(err, database.ErrMerchantNotFound) {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}
}

func TestGetMerchantsByPortfolioType(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()

	// Create merchants with different portfolio types
	merchant1 := createTestMerchant()
	merchant1.Name = "Company 1"
	merchant1.PortfolioType = PortfolioTypeOnboarded
	_, err := service.CreateMerchant(ctx, merchant1, "user1")
	if err != nil {
		t.Fatalf("Failed to create merchant 1: %v", err)
	}

	merchant2 := createTestMerchant()
	merchant2.Name = "Company 2"
	merchant2.PortfolioType = PortfolioTypeProspective
	_, err = service.CreateMerchant(ctx, merchant2, "user2")
	if err != nil {
		t.Fatalf("Failed to create merchant 2: %v", err)
	}

	// Test filtering by portfolio type
	result, err := service.GetMerchantsByPortfolioType(ctx, PortfolioTypeOnboarded, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 1 {
		t.Errorf("Expected 1 merchant, got %d", len(result.Merchants))
	}

	if result.Merchants[0].PortfolioType != PortfolioTypeOnboarded {
		t.Error("Expected merchant to have onboarded portfolio type")
	}

	// Test invalid portfolio type
	_, err = service.GetMerchantsByPortfolioType(ctx, "invalid", 1, 10)
	if !errors.Is(err, ErrInvalidPortfolioType) {
		t.Errorf("Expected ErrInvalidPortfolioType, got: %v", err)
	}
}

func TestGetMerchantsByRiskLevel(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()

	// Create merchants with different risk levels
	merchant1 := createTestMerchant()
	merchant1.Name = "Company 1"
	merchant1.RiskLevel = RiskLevelHigh
	_, err := service.CreateMerchant(ctx, merchant1, "user1")
	if err != nil {
		t.Fatalf("Failed to create merchant 1: %v", err)
	}

	merchant2 := createTestMerchant()
	merchant2.Name = "Company 2"
	merchant2.RiskLevel = RiskLevelLow
	_, err = service.CreateMerchant(ctx, merchant2, "user2")
	if err != nil {
		t.Fatalf("Failed to create merchant 2: %v", err)
	}

	// Test filtering by risk level
	result, err := service.GetMerchantsByRiskLevel(ctx, RiskLevelHigh, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 1 {
		t.Errorf("Expected 1 merchant, got %d", len(result.Merchants))
	}

	if result.Merchants[0].RiskLevel != RiskLevelHigh {
		t.Error("Expected merchant to have high risk level")
	}

	// Test invalid risk level
	_, err = service.GetMerchantsByRiskLevel(ctx, "invalid", 1, 10)
	if !errors.Is(err, ErrInvalidRiskLevel) {
		t.Errorf("Expected ErrInvalidRiskLevel, got: %v", err)
	}
}

func TestMerchantSessionManagement(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Create a merchant first
	merchant := createTestMerchant()
	created, err := service.CreateMerchant(ctx, merchant, userID)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	// Test starting a session
	session, err := service.StartMerchantSession(ctx, userID, created.ID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if session.UserID != userID {
		t.Error("Expected session user ID to match")
	}

	if session.MerchantID != created.ID {
		t.Error("Expected session merchant ID to match")
	}

	if !session.IsActive {
		t.Error("Expected session to be active")
	}

	// Test getting active session
	activeSession, err := service.GetActiveMerchantSession(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if activeSession.ID != session.ID {
		t.Error("Expected active session ID to match")
	}

	// Test ending session
	err = service.EndMerchantSession(ctx, userID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	// Test getting session after ending
	_, err = service.GetActiveMerchantSession(ctx, userID)
	if !errors.Is(err, ErrNoActiveSession) {
		t.Errorf("Expected ErrNoActiveSession, got: %v", err)
	}

	// Test starting session with non-existent merchant
	_, err = service.StartMerchantSession(ctx, userID, "nonexistent")
	if !errors.Is(err, database.ErrMerchantNotFound) {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}
}

func TestSearchMerchants(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()

	// Create test merchants
	merchant1 := createTestMerchant()
	merchant1.Name = "Tech Company"
	merchant1.Industry = "Technology"
	merchant1.PortfolioType = PortfolioTypeOnboarded
	merchant1.RiskLevel = RiskLevelLow
	_, err := service.CreateMerchant(ctx, merchant1, "user1")
	if err != nil {
		t.Fatalf("Failed to create merchant 1: %v", err)
	}

	merchant2 := createTestMerchant()
	merchant2.Name = "Finance Corp"
	merchant2.Industry = "Finance"
	merchant2.PortfolioType = PortfolioTypeProspective
	merchant2.RiskLevel = RiskLevelHigh
	_, err = service.CreateMerchant(ctx, merchant2, "user2")
	if err != nil {
		t.Fatalf("Failed to create merchant 2: %v", err)
	}

	// Test search with portfolio type filter
	portfolioType := PortfolioTypeOnboarded
	filters := &MerchantSearchFilters{
		PortfolioType: &portfolioType,
	}
	result, err := service.SearchMerchants(ctx, filters, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 1 {
		t.Errorf("Expected 1 merchant, got %d", len(result.Merchants))
	}

	// Test search with risk level filter
	riskLevel := RiskLevelHigh
	filters = &MerchantSearchFilters{
		RiskLevel: &riskLevel,
	}
	result, err = service.SearchMerchants(ctx, filters, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 1 {
		t.Errorf("Expected 1 merchant, got %d", len(result.Merchants))
	}

	// Test search with query
	filters = &MerchantSearchFilters{
		SearchQuery: "Tech",
	}
	result, err = service.SearchMerchants(ctx, filters, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 1 {
		t.Errorf("Expected 1 merchant, got %d", len(result.Merchants))
	}

	// Test search with no filters
	result, err = service.SearchMerchants(ctx, nil, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 2 {
		t.Errorf("Expected 2 merchants, got %d", len(result.Merchants))
	}
}

func TestBulkUpdatePortfolioType(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Create test merchants
	merchant1 := createTestMerchant()
	merchant1.Name = "Company 1"
	_, err := service.CreateMerchant(ctx, merchant1, userID)
	if err != nil {
		t.Fatalf("Failed to create merchant 1: %v", err)
	}

	merchant2 := createTestMerchant()
	merchant2.Name = "Company 2"
	created2, err := service.CreateMerchant(ctx, merchant2, userID)
	if err != nil {
		t.Fatalf("Failed to create merchant 2: %v", err)
	}

	// Test bulk update
	merchantIDs := []string{created2.ID}
	result, err := service.BulkUpdatePortfolioType(ctx, merchantIDs, PortfolioTypeOnboarded, userID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.TotalItems != 1 {
		t.Error("Expected total items to be 1")
	}

	if result.Successful != 1 {
		t.Error("Expected 1 successful update")
	}

	if result.Failed != 0 {
		t.Error("Expected 0 failed updates")
	}

	// Test with invalid portfolio type
	_, err = service.BulkUpdatePortfolioType(ctx, merchantIDs, "invalid", userID)
	if !errors.Is(err, ErrInvalidPortfolioType) {
		t.Errorf("Expected ErrInvalidPortfolioType, got: %v", err)
	}
}

func TestBulkUpdateRiskLevel(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Create test merchants
	merchant1 := createTestMerchant()
	merchant1.Name = "Company 1"
	_, err := service.CreateMerchant(ctx, merchant1, userID)
	if err != nil {
		t.Fatalf("Failed to create merchant 1: %v", err)
	}

	merchant2 := createTestMerchant()
	merchant2.Name = "Company 2"
	created2, err := service.CreateMerchant(ctx, merchant2, userID)
	if err != nil {
		t.Fatalf("Failed to create merchant 2: %v", err)
	}

	// Test bulk update
	merchantIDs := []string{created2.ID}
	result, err := service.BulkUpdateRiskLevel(ctx, merchantIDs, RiskLevelHigh, userID)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.TotalItems != 1 {
		t.Error("Expected total items to be 1")
	}

	if result.Successful != 1 {
		t.Error("Expected 1 successful update")
	}

	if result.Failed != 0 {
		t.Error("Expected 0 failed updates")
	}

	// Test with invalid risk level
	_, err = service.BulkUpdateRiskLevel(ctx, merchantIDs, "invalid", userID)
	if !errors.Is(err, ErrInvalidRiskLevel) {
		t.Errorf("Expected ErrInvalidRiskLevel, got: %v", err)
	}
}

func TestValidationHelpers(t *testing.T) {
	service, _ := createTestService()

	// Test portfolio type validation
	if !service.isValidPortfolioType(PortfolioTypeOnboarded) {
		t.Error("Expected onboarded to be valid")
	}

	if !service.isValidPortfolioType(PortfolioTypeDeactivated) {
		t.Error("Expected deactivated to be valid")
	}

	if !service.isValidPortfolioType(PortfolioTypeProspective) {
		t.Error("Expected prospective to be valid")
	}

	if !service.isValidPortfolioType(PortfolioTypePending) {
		t.Error("Expected pending to be valid")
	}

	if service.isValidPortfolioType("invalid") {
		t.Error("Expected invalid to be invalid")
	}

	// Test risk level validation
	if !service.isValidRiskLevel(RiskLevelHigh) {
		t.Error("Expected high to be valid")
	}

	if !service.isValidRiskLevel(RiskLevelMedium) {
		t.Error("Expected medium to be valid")
	}

	if !service.isValidRiskLevel(RiskLevelLow) {
		t.Error("Expected low to be valid")
	}

	if service.isValidRiskLevel("invalid") {
		t.Error("Expected invalid to be invalid")
	}
}

// =============================================================================
// Additional Comprehensive Tests
// =============================================================================

func TestMerchantPortfolioService_EdgeCases(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Test creating merchant with minimal data
	minimalMerchant := &Merchant{
		Name: "Minimal Company",
	}
	created, err := service.CreateMerchant(ctx, minimalMerchant, userID)
	if err != nil {
		t.Fatalf("Expected no error for minimal merchant, got: %v", err)
	}

	if created.PortfolioType != PortfolioTypeProspective {
		t.Error("Expected default portfolio type for minimal merchant")
	}

	if created.RiskLevel != RiskLevelMedium {
		t.Error("Expected default risk level for minimal merchant")
	}

	if created.Status != "active" {
		t.Error("Expected default status for minimal merchant")
	}
}

func TestMerchantPortfolioService_ConcurrentOperations(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Test concurrent merchant creation
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			merchant := createTestMerchant()
			merchant.Name = fmt.Sprintf("Concurrent Company %d", id)
			_, err := service.CreateMerchant(ctx, merchant, userID)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Errorf("Concurrent creation failed: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent operations timed out")
		}
	}
}

func TestMerchantPortfolioService_DataConversion(t *testing.T) {
	service, _ := createTestService()

	// Test business to merchant conversion
	business := &database.Business{
		ID:                 "test_business",
		Name:               "Test Business",
		LegalName:          "Test Business LLC",
		RegistrationNumber: "REG123",
		TaxID:              "TAX123",
		Industry:           "Technology",
		IndustryCode:       "541511",
		BusinessType:       "LLC",
		EmployeeCount:      25,
		Address: database.Address{
			Street1:    "123 Test St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
		},
		ContactInfo: database.ContactInfo{
			Phone:   "+1-555-123-4567",
			Email:   "test@testbusiness.com",
			Website: "https://testbusiness.com",
		},
		Status:           "onboarded",
		RiskLevel:        "high",
		ComplianceStatus: "compliant",
		CreatedBy:        "user123",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	merchant := service.businessToMerchant(business)

	if merchant.ID != business.ID {
		t.Error("Expected merchant ID to match business ID")
	}

	if merchant.PortfolioType != PortfolioTypeOnboarded {
		t.Error("Expected portfolio type to be mapped from status")
	}

	if merchant.RiskLevel != RiskLevelHigh {
		t.Error("Expected risk level to be mapped correctly")
	}
}

func TestMerchantPortfolioService_ErrorHandling(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Test various error scenarios
	tests := []struct {
		name        string
		merchant    *Merchant
		expectedErr string
	}{
		{
			name:        "nil merchant",
			merchant:    nil,
			expectedErr: "merchant cannot be nil",
		},
		{
			name: "empty name",
			merchant: &Merchant{
				Name: "",
			},
			expectedErr: "merchant name is required",
		},
		{
			name: "whitespace only name",
			merchant: &Merchant{
				Name: "   ",
			},
			expectedErr: "merchant name is required",
		},
		{
			name: "invalid portfolio type",
			merchant: &Merchant{
				Name:          "Test Company",
				PortfolioType: "invalid_type",
			},
			expectedErr: "invalid portfolio type",
		},
		{
			name: "invalid risk level",
			merchant: &Merchant{
				Name:      "Test Company",
				RiskLevel: "invalid_level",
			},
			expectedErr: "invalid risk level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.CreateMerchant(ctx, tt.merchant, userID)
			if err == nil {
				t.Errorf("Expected error for %s, got nil", tt.name)
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error to contain '%s', got '%s'", tt.expectedErr, err.Error())
			}
		})
	}
}

func TestMerchantPortfolioService_Pagination(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()

	// Create multiple merchants for pagination testing
	for i := 0; i < 25; i++ {
		merchant := createTestMerchant()
		merchant.Name = fmt.Sprintf("Company %d", i)
		_, err := service.CreateMerchant(ctx, merchant, "user123")
		if err != nil {
			t.Fatalf("Failed to create merchant %d: %v", i, err)
		}
	}

	// Test first page
	result, err := service.SearchMerchants(ctx, nil, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 10 {
		t.Errorf("Expected 10 merchants on first page, got %d", len(result.Merchants))
	}

	if result.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Page)
	}

	if result.PageSize != 10 {
		t.Errorf("Expected page size 10, got %d", result.PageSize)
	}

	if !result.HasMore {
		t.Error("Expected HasMore to be true for first page")
	}

	// Test second page
	result, err = service.SearchMerchants(ctx, nil, 2, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 10 {
		t.Errorf("Expected 10 merchants on second page, got %d", len(result.Merchants))
	}

	if result.Page != 2 {
		t.Errorf("Expected page 2, got %d", result.Page)
	}

	// Test last page
	result, err = service.SearchMerchants(ctx, nil, 3, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 5 {
		t.Errorf("Expected 5 merchants on last page, got %d", len(result.Merchants))
	}

	if result.HasMore {
		t.Error("Expected HasMore to be false for last page")
	}
}

func TestMerchantPortfolioService_Filtering(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()

	// Create merchants with different attributes
	merchants := []struct {
		name          string
		portfolioType PortfolioType
		riskLevel     RiskLevel
		industry      string
	}{
		{"Tech Company 1", PortfolioTypeOnboarded, RiskLevelLow, "Technology"},
		{"Tech Company 2", PortfolioTypeProspective, RiskLevelMedium, "Technology"},
		{"Finance Corp", PortfolioTypeOnboarded, RiskLevelHigh, "Finance"},
		{"Retail Store", PortfolioTypePending, RiskLevelLow, "Retail"},
	}

	for _, m := range merchants {
		merchant := createTestMerchant()
		merchant.Name = m.name
		merchant.PortfolioType = m.portfolioType
		merchant.RiskLevel = m.riskLevel
		merchant.Industry = m.industry
		_, err := service.CreateMerchant(ctx, merchant, "user123")
		if err != nil {
			t.Fatalf("Failed to create merchant %s: %v", m.name, err)
		}
	}

	// Test portfolio type filtering
	portfolioType := PortfolioTypeOnboarded
	filters := &MerchantSearchFilters{
		PortfolioType: &portfolioType,
	}
	result, err := service.SearchMerchants(ctx, filters, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 2 {
		t.Errorf("Expected 2 onboarded merchants, got %d", len(result.Merchants))
	}

	// Test risk level filtering
	riskLevel := RiskLevelHigh
	filters = &MerchantSearchFilters{
		RiskLevel: &riskLevel,
	}
	result, err = service.SearchMerchants(ctx, filters, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 1 {
		t.Errorf("Expected 1 high risk merchant, got %d", len(result.Merchants))
	}

	// Test industry filtering
	filters = &MerchantSearchFilters{
		Industry: "Technology",
	}
	result, err = service.SearchMerchants(ctx, filters, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 2 {
		t.Errorf("Expected 2 technology merchants, got %d", len(result.Merchants))
	}

	// Test combined filters
	filters = &MerchantSearchFilters{
		PortfolioType: &portfolioType,
		Industry:      "Technology",
	}
	result, err = service.SearchMerchants(ctx, filters, 1, 10)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(result.Merchants) != 1 {
		t.Errorf("Expected 1 onboarded technology merchant, got %d", len(result.Merchants))
	}
}

func TestMerchantPortfolioService_SessionManagement_EdgeCases(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Test starting session with non-existent merchant
	_, err := service.StartMerchantSession(ctx, userID, "nonexistent")
	if !errors.Is(err, database.ErrMerchantNotFound) {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}

	// Test ending session when no active session exists
	err = service.EndMerchantSession(ctx, userID)
	if !errors.Is(err, ErrNoActiveSession) {
		t.Errorf("Expected ErrNoActiveSession, got: %v", err)
	}

	// Test getting active session when none exists
	_, err = service.GetActiveMerchantSession(ctx, userID)
	if !errors.Is(err, ErrNoActiveSession) {
		t.Errorf("Expected ErrNoActiveSession, got: %v", err)
	}
}

func TestMerchantPortfolioService_BulkOperations_EdgeCases(t *testing.T) {
	service, _ := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Test bulk update with empty list
	result, err := service.BulkUpdatePortfolioType(ctx, []string{}, PortfolioTypeOnboarded, userID)
	if err != nil {
		t.Errorf("Expected no error for empty list, got: %v", err)
	}

	if result.TotalItems != 0 {
		t.Errorf("Expected 0 total items, got %d", result.TotalItems)
	}

	// Test bulk update with non-existent merchants
	result, err = service.BulkUpdatePortfolioType(ctx, []string{"nonexistent1", "nonexistent2"}, PortfolioTypeOnboarded, userID)
	if err != nil {
		t.Errorf("Expected no error for non-existent merchants, got: %v", err)
	}

	if result.Successful != 0 {
		t.Errorf("Expected 0 successful updates, got %d", result.Successful)
	}

	if result.Failed != 2 {
		t.Errorf("Expected 2 failed updates, got %d", result.Failed)
	}
}

func TestMerchantPortfolioService_IDGeneration(t *testing.T) {
	service, _ := createTestService()

	// Test ID generation uniqueness
	id1 := service.generateID()
	id2 := service.generateID()

	if id1 == id2 {
		t.Error("Expected generated IDs to be unique")
	}

	if !strings.HasPrefix(id1, "merchant_") {
		t.Error("Expected ID to have merchant_ prefix")
	}

	if !strings.HasPrefix(id2, "merchant_") {
		t.Error("Expected ID to have merchant_ prefix")
	}
}

func TestMerchantPortfolioService_AuditLogging(t *testing.T) {
	service, mockDB := createTestService()
	ctx := context.Background()
	userID := "user123"

	// Create a merchant to test audit logging
	merchant := createTestMerchant()
	_, err := service.CreateMerchant(ctx, merchant, userID)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	// Verify audit log was created
	if len(mockDB.auditLogs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(mockDB.auditLogs))
	}

	auditLog := mockDB.auditLogs[0]
	if auditLog.UserID != userID {
		t.Errorf("Expected audit log user ID %s, got %s", userID, auditLog.UserID)
	}

	if auditLog.Action != "CREATE_MERCHANT" {
		t.Errorf("Expected audit log action CREATE_MERCHANT, got %s", auditLog.Action)
	}

	if auditLog.ResourceType != "merchant" {
		t.Errorf("Expected audit log resource type merchant, got %s", auditLog.ResourceType)
	}
}
