package database

import (
	"context"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/models"
)

// Test helper functions
func createTestMockDB() *MockMerchantDatabase {
	logger := log.New(log.Writer(), "test: ", log.LstdFlags)
	return NewMockMerchantDatabase(logger)
}

func createTestMerchantForMock() *models.Merchant {
	now := time.Now()
	return &models.Merchant{
		ID:                 "test_merchant_001",
		Name:               "Test Company",
		LegalName:          "Test Company LLC",
		RegistrationNumber: "REG123456",
		TaxID:              "TAX123456",
		Industry:           "Technology",
		IndustryCode:       "541511",
		BusinessType:       "LLC",
		FoundedDate:        &now,
		EmployeeCount:      50,
		AnnualRevenue:      func() *float64 { v := 2500000.0; return &v }(),
		Address: models.Address{
			Street1:    "123 Test St",
			City:       "Test City",
			State:      "TS",
			PostalCode: "12345",
			Country:    "USA",
		},
		ContactInfo: models.ContactInfo{
			Phone:   "+1-555-123-4567",
			Email:   "test@testcompany.com",
			Website: "https://testcompany.com",
		},
		PortfolioType:    models.PortfolioTypeProspective,
		RiskLevel:        models.RiskLevelMedium,
		ComplianceStatus: "pending",
		Status:           "active",
		CreatedBy:        "user123",
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func createTestMerchantSessionForMock() *models.MerchantSession {
	now := time.Now()
	return &models.MerchantSession{
		ID:         "test_session_001",
		UserID:     "user123",
		MerchantID: "test_merchant_001",
		StartedAt:  now,
		LastActive: now,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func createTestAuditLogForMock() *models.AuditLog {
	return &models.AuditLog{
		ID:           "test_audit_001",
		UserID:       "user123",
		MerchantID:   "test_merchant_001",
		Action:       "CREATE_MERCHANT",
		ResourceType: "merchant",
		ResourceID:   "test_merchant_001",
		Details:      "Merchant created",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0",
		RequestID:    "req_123",
		CreatedAt:    time.Now(),
	}
}

// Tests
func TestNewMockMerchantDatabase(t *testing.T) {
	mockDB := createTestMockDB()

	if mockDB == nil {
		t.Fatal("Expected mock database to be created")
	}

	if mockDB.merchants == nil {
		t.Error("Expected merchants map to be initialized")
	}

	if mockDB.sessions == nil {
		t.Error("Expected sessions map to be initialized")
	}

	if mockDB.auditLogs == nil {
		t.Error("Expected audit logs slice to be initialized")
	}

	if mockDB.logger == nil {
		t.Error("Expected logger to be set")
	}

	// Check that initial data was loaded
	if len(mockDB.merchants) != 100 {
		t.Errorf("Expected 100 initial merchants, got %d", len(mockDB.merchants))
	}
}

func TestMockMerchantDatabase_CreateMerchant(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	merchant := createTestMerchantForMock()

	err := mockDB.CreateMerchant(ctx, merchant)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify merchant was created
	retrieved, err := mockDB.GetMerchant(ctx, merchant.ID)
	if err != nil {
		t.Errorf("Expected to retrieve merchant, got error: %v", err)
	}

	if retrieved.ID != merchant.ID {
		t.Error("Retrieved merchant ID doesn't match")
	}

	// Test duplicate creation
	err = mockDB.CreateMerchant(ctx, merchant)
	if err != ErrDuplicateMerchant {
		t.Errorf("Expected ErrDuplicateMerchant, got: %v", err)
	}
}

func TestMockMerchantDatabase_GetMerchant(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	// Test getting existing merchant (from initial data)
	merchant, err := mockDB.GetMerchant(ctx, "merchant_001")
	if err != nil {
		t.Errorf("Expected to get merchant, got error: %v", err)
	}

	if merchant.ID != "merchant_001" {
		t.Error("Retrieved merchant ID doesn't match")
	}

	// Test getting non-existent merchant
	_, err = mockDB.GetMerchant(ctx, "non_existent")
	if err != ErrMerchantNotFound {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}
}

func TestMockMerchantDatabase_UpdateMerchant(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	// Get an existing merchant
	merchant, err := mockDB.GetMerchant(ctx, "merchant_001")
	if err != nil {
		t.Fatalf("Failed to get merchant: %v", err)
	}

	// Update the merchant
	merchant.Name = "Updated Company Name"
	err = mockDB.UpdateMerchant(ctx, merchant)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify update
	updated, err := mockDB.GetMerchant(ctx, merchant.ID)
	if err != nil {
		t.Errorf("Expected to get updated merchant, got error: %v", err)
	}

	if updated.Name != "Updated Company Name" {
		t.Error("Merchant name was not updated")
	}

	// Test updating non-existent merchant
	nonExistent := createTestMerchant()
	nonExistent.ID = "non_existent"
	err = mockDB.UpdateMerchant(ctx, nonExistent)
	if err != ErrMerchantNotFound {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}
}

func TestMockMerchantDatabase_DeleteMerchant(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	// Create a merchant to delete
	merchant := createTestMerchantForMock()
	err := mockDB.CreateMerchant(ctx, merchant)
	if err != nil {
		t.Fatalf("Failed to create merchant: %v", err)
	}

	// Delete the merchant
	err = mockDB.DeleteMerchant(ctx, merchant.ID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify deletion
	_, err = mockDB.GetMerchant(ctx, merchant.ID)
	if err != ErrMerchantNotFound {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}

	// Test deleting non-existent merchant
	err = mockDB.DeleteMerchant(ctx, "non_existent")
	if err != ErrMerchantNotFound {
		t.Errorf("Expected ErrMerchantNotFound, got: %v", err)
	}
}

func TestMockMerchantDatabase_ListMerchants(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	// Test valid pagination
	merchants, err := mockDB.ListMerchants(ctx, 1, 10)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(merchants) != 10 {
		t.Errorf("Expected 10 merchants, got %d", len(merchants))
	}

	// Test invalid pagination
	_, err = mockDB.ListMerchants(ctx, 0, 10)
	if err != ErrInvalidPagination {
		t.Errorf("Expected ErrInvalidPagination, got: %v", err)
	}

	_, err = mockDB.ListMerchants(ctx, 1, 0)
	if err != ErrInvalidPagination {
		t.Errorf("Expected ErrInvalidPagination, got: %v", err)
	}

	_, err = mockDB.ListMerchants(ctx, 1, 1001)
	if err != ErrInvalidPagination {
		t.Errorf("Expected ErrInvalidPagination, got: %v", err)
	}
}

func TestMockMerchantDatabase_SearchMerchants(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	// Test search with portfolio type filter
	portfolioType := models.PortfolioTypeOnboarded
	filters := &models.MerchantSearchFilters{
		PortfolioType: &portfolioType,
	}

	merchants, err := mockDB.SearchMerchants(ctx, filters, 1, 10)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify all returned merchants have the correct portfolio type
	for _, merchant := range merchants {
		if merchant.PortfolioType != portfolioType {
			t.Errorf("Expected portfolio type %s, got %s", portfolioType, merchant.PortfolioType)
		}
	}

	// Test search with risk level filter
	riskLevel := models.RiskLevelHigh
	filters = &models.MerchantSearchFilters{
		RiskLevel: &riskLevel,
	}

	merchants, err = mockDB.SearchMerchants(ctx, filters, 1, 10)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify all returned merchants have the correct risk level
	for _, merchant := range merchants {
		if merchant.RiskLevel != riskLevel {
			t.Errorf("Expected risk level %s, got %s", riskLevel, merchant.RiskLevel)
		}
	}

	// Test search with text query
	filters = &models.MerchantSearchFilters{
		SearchQuery: "Tech",
	}

	merchants, err = mockDB.SearchMerchants(ctx, filters, 1, 10)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify all returned merchants contain "Tech" in name, legal name, or industry
	for _, merchant := range merchants {
		if !containsIgnoreCase(merchant.Name, "Tech") &&
			!containsIgnoreCase(merchant.LegalName, "Tech") &&
			!containsIgnoreCase(merchant.Industry, "Tech") {
			t.Errorf("Merchant %s doesn't contain 'Tech' in searchable fields", merchant.Name)
		}
	}
}

func TestMockMerchantDatabase_CountMerchants(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	// Test count with no filters
	count, err := mockDB.CountMerchants(ctx, nil)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if count != 100 {
		t.Errorf("Expected 100 merchants, got %d", count)
	}

	// Test count with portfolio type filter
	portfolioType := models.PortfolioTypeOnboarded
	filters := &models.MerchantSearchFilters{
		PortfolioType: &portfolioType,
	}

	count, err = mockDB.CountMerchants(ctx, filters)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Count should be less than total
	if count >= 100 {
		t.Errorf("Expected count less than 100, got %d", count)
	}
}

func TestMockMerchantDatabase_CreateSession(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	session := createTestMerchantSessionForMock()

	err := mockDB.CreateSession(ctx, session)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test duplicate creation
	err = mockDB.CreateSession(ctx, session)
	if err != ErrDuplicateSession {
		t.Errorf("Expected ErrDuplicateSession, got: %v", err)
	}
}

func TestMockMerchantDatabase_GetActiveSessionByUserID(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	session := createTestMerchantSessionForMock()
	err := mockDB.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Test getting active session
	retrieved, err := mockDB.GetActiveSessionByUserID(ctx, session.UserID)
	if err != nil {
		t.Errorf("Expected to get session, got error: %v", err)
	}

	if retrieved.ID != session.ID {
		t.Error("Retrieved session ID doesn't match")
	}

	// Test getting session for non-existent user
	_, err = mockDB.GetActiveSessionByUserID(ctx, "non_existent_user")
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got: %v", err)
	}
}

func TestMockMerchantDatabase_UpdateSession(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	session := createTestMerchantSessionForMock()
	err := mockDB.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Update the session
	session.IsActive = false
	err = mockDB.UpdateSession(ctx, session)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify update - session should be inactive
	_, err = mockDB.GetActiveSessionByUserID(ctx, session.UserID)
	if err != ErrSessionNotFound {
		t.Error("Expected session to be inactive")
	}

	// Test updating non-existent session
	nonExistent := createTestMerchantSessionForMock()
	nonExistent.ID = "non_existent"
	err = mockDB.UpdateSession(ctx, nonExistent)
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got: %v", err)
	}
}

func TestMockMerchantDatabase_DeactivateSession(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	session := createTestMerchantSessionForMock()
	err := mockDB.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("Failed to create session: %v", err)
	}

	// Deactivate the session
	err = mockDB.DeactivateSession(ctx, session.ID)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify deactivation
	_, err = mockDB.GetActiveSessionByUserID(ctx, session.UserID)
	if err != ErrSessionNotFound {
		t.Error("Expected session to be inactive")
	}

	// Test deactivating non-existent session
	err = mockDB.DeactivateSession(ctx, "non_existent")
	if err != ErrSessionNotFound {
		t.Errorf("Expected ErrSessionNotFound, got: %v", err)
	}
}

func TestMockMerchantDatabase_CreateAuditLog(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	auditLog := createTestAuditLogForMock()

	err := mockDB.CreateAuditLog(ctx, auditLog)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify audit log was created
	logs, err := mockDB.GetAuditLogsByMerchantID(ctx, auditLog.MerchantID, 1, 10)
	if err != nil {
		t.Errorf("Expected to get audit logs, got error: %v", err)
	}

	if len(logs) != 1 {
		t.Errorf("Expected 1 audit log, got %d", len(logs))
	}

	if logs[0].ID != auditLog.ID {
		t.Error("Retrieved audit log ID doesn't match")
	}
}

func TestMockMerchantDatabase_GetAuditLogsByMerchantID(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	// Create multiple audit logs for the same merchant
	merchantID := "test_merchant_001"
	for i := 0; i < 5; i++ {
		auditLog := &models.AuditLog{
			ID:           fmt.Sprintf("audit_%d", i),
			UserID:       "user123",
			MerchantID:   merchantID,
			Action:       "TEST_ACTION",
			ResourceType: "merchant",
			ResourceID:   merchantID,
			Details:      fmt.Sprintf("Test action %d", i),
			CreatedAt:    time.Now(),
		}

		err := mockDB.CreateAuditLog(ctx, auditLog)
		if err != nil {
			t.Fatalf("Failed to create audit log: %v", err)
		}
	}

	// Test pagination
	logs, err := mockDB.GetAuditLogsByMerchantID(ctx, merchantID, 1, 3)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(logs) != 3 {
		t.Errorf("Expected 3 audit logs, got %d", len(logs))
	}

	// Test invalid pagination
	_, err = mockDB.GetAuditLogsByMerchantID(ctx, merchantID, 0, 10)
	if err != ErrInvalidPagination {
		t.Errorf("Expected ErrInvalidPagination, got: %v", err)
	}
}

func TestMockMerchantDatabase_BulkUpdatePortfolioType(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	// Get some existing merchant IDs
	merchants, err := mockDB.ListMerchants(ctx, 1, 5)
	if err != nil {
		t.Fatalf("Failed to list merchants: %v", err)
	}

	merchantIDs := make([]string, len(merchants))
	for i, merchant := range merchants {
		merchantIDs[i] = merchant.ID
	}

	// Test bulk update
	err = mockDB.BulkUpdatePortfolioType(ctx, merchantIDs, models.PortfolioTypeOnboarded, "user123")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify updates
	for _, merchantID := range merchantIDs {
		merchant, err := mockDB.GetMerchant(ctx, merchantID)
		if err != nil {
			t.Errorf("Failed to get merchant %s: %v", merchantID, err)
		}

		if merchant.PortfolioType != models.PortfolioTypeOnboarded {
			t.Errorf("Expected portfolio type Onboarded, got %s", merchant.PortfolioType)
		}
	}

	// Test with empty list
	err = mockDB.BulkUpdatePortfolioType(ctx, []string{}, models.PortfolioTypeOnboarded, "user123")
	if err != nil {
		t.Errorf("Expected no error for empty list, got: %v", err)
	}
}

func TestMockMerchantDatabase_BulkUpdateRiskLevel(t *testing.T) {
	mockDB := createTestMockDB()
	ctx := context.Background()

	// Get some existing merchant IDs
	merchants, err := mockDB.ListMerchants(ctx, 1, 5)
	if err != nil {
		t.Fatalf("Failed to list merchants: %v", err)
	}

	merchantIDs := make([]string, len(merchants))
	for i, merchant := range merchants {
		merchantIDs[i] = merchant.ID
	}

	// Test bulk update
	err = mockDB.BulkUpdateRiskLevel(ctx, merchantIDs, models.RiskLevelHigh, "user123")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Verify updates
	for _, merchantID := range merchantIDs {
		merchant, err := mockDB.GetMerchant(ctx, merchantID)
		if err != nil {
			t.Errorf("Failed to get merchant %s: %v", merchantID, err)
		}

		if merchant.RiskLevel != models.RiskLevelHigh {
			t.Errorf("Expected risk level High, got %s", merchant.RiskLevel)
		}
	}

	// Test with empty list
	err = mockDB.BulkUpdateRiskLevel(ctx, []string{}, models.RiskLevelHigh, "user123")
	if err != nil {
		t.Errorf("Expected no error for empty list, got: %v", err)
	}
}

func TestMockMerchantDatabase_UtilityMethods(t *testing.T) {
	mockDB := createTestMockDB()

	// Test counts
	merchantCount := mockDB.GetMerchantCount()
	if merchantCount != 100 {
		t.Errorf("Expected 100 merchants, got %d", merchantCount)
	}

	sessionCount := mockDB.GetSessionCount()
	if sessionCount != 0 {
		t.Errorf("Expected 0 sessions, got %d", sessionCount)
	}

	auditLogCount := mockDB.GetAuditLogCount()
	if auditLogCount != 0 {
		t.Errorf("Expected 0 audit logs, got %d", auditLogCount)
	}

	// Test clear all data
	mockDB.ClearAllData()

	merchantCount = mockDB.GetMerchantCount()
	if merchantCount != 0 {
		t.Errorf("Expected 0 merchants after clear, got %d", merchantCount)
	}

	// Test reset to initial data
	mockDB.ResetToInitialData()

	merchantCount = mockDB.GetMerchantCount()
	if merchantCount != 100 {
		t.Errorf("Expected 100 merchants after reset, got %d", merchantCount)
	}
}

// Helper function for case-insensitive string contains
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
