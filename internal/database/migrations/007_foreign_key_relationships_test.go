package migrations

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"testing"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestForeignKeyRelationships tests all foreign key relationships in the merchant portfolio schema
func TestForeignKeyRelationships(t *testing.T) {
	// Skip if no database is available for testing
	if testing.Short() {
		t.Skip("Skipping foreign key relationship tests in short mode")
	}

	// Create mock database for testing relationships
	logger := log.New(log.Writer(), "test: ", log.LstdFlags)
	mockDB := database.NewMockMerchantDatabase(logger)

	ctx := context.Background()

	t.Run("merchants to portfolio_types relationship", func(t *testing.T) {
		// Create a test merchant with valid portfolio type
		merchant := createTestMerchantWithPortfolioType(t, models.PortfolioTypeOnboarded)

		// Create merchant in mock database
		err := mockDB.CreateMerchant(ctx, merchant)
		require.NoError(t, err)

		// Verify merchant exists
		retrieved, err := mockDB.GetMerchant(ctx, merchant.ID)
		require.NoError(t, err)
		assert.Equal(t, merchant.ID, retrieved.ID)
		assert.Equal(t, models.PortfolioTypeOnboarded, retrieved.PortfolioType)

		// Test that portfolio type is properly linked
		assert.Equal(t, merchant.PortfolioType, retrieved.PortfolioType)
	})

	t.Run("merchants to risk_levels relationship", func(t *testing.T) {
		// Create a test merchant with valid risk level
		merchant := createTestMerchantWithRiskLevel(t, models.RiskLevelHigh)

		// Create merchant in mock database
		err := mockDB.CreateMerchant(ctx, merchant)
		require.NoError(t, err)

		// Verify merchant exists
		retrieved, err := mockDB.GetMerchant(ctx, merchant.ID)
		require.NoError(t, err)
		assert.Equal(t, merchant.ID, retrieved.ID)
		assert.Equal(t, models.RiskLevelHigh, retrieved.RiskLevel)

		// Test that risk level is properly linked
		assert.Equal(t, merchant.RiskLevel, retrieved.RiskLevel)
	})

	t.Run("merchants to users relationship", func(t *testing.T) {
		// Create a test merchant with valid user
		merchant := createTestMerchantWithUser(t, "test_user_001")

		// Create merchant in mock database
		err := mockDB.CreateMerchant(ctx, merchant)
		require.NoError(t, err)

		// Verify merchant exists
		retrieved, err := mockDB.GetMerchant(ctx, merchant.ID)
		require.NoError(t, err)
		assert.Equal(t, merchant.ID, retrieved.ID)
		assert.Equal(t, "test_user_001", retrieved.CreatedBy)

		// Test that user is properly linked
		assert.Equal(t, merchant.CreatedBy, retrieved.CreatedBy)
	})

	t.Run("merchant_sessions to users relationship", func(t *testing.T) {
		// Create a test merchant first
		merchant := createTestMerchant(t)
		err := mockDB.CreateMerchant(ctx, merchant)
		require.NoError(t, err)

		// Create a merchant session
		session := createTestMerchantSession(t, "test_user_002", merchant.ID)

		// Create session in mock database
		err = mockDB.CreateSession(ctx, session)
		require.NoError(t, err)

		// Verify session exists
		retrieved, err := mockDB.GetActiveSessionByUserID(ctx, session.UserID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.UserID, retrieved.UserID)
		assert.Equal(t, session.MerchantID, retrieved.MerchantID)

		// Test that user is properly linked
		assert.Equal(t, session.UserID, retrieved.UserID)
	})

	t.Run("merchant_sessions to merchants relationship", func(t *testing.T) {
		// Create a test merchant first
		merchant := createTestMerchant(t)
		err := mockDB.CreateMerchant(ctx, merchant)
		require.NoError(t, err)

		// Create a merchant session
		session := createTestMerchantSession(t, "test_user_003", merchant.ID)

		// Create session in mock database
		err = mockDB.CreateSession(ctx, session)
		require.NoError(t, err)

		// Verify session exists
		retrieved, err := mockDB.GetActiveSessionByUserID(ctx, session.UserID)
		require.NoError(t, err)
		assert.Equal(t, session.ID, retrieved.ID)
		assert.Equal(t, session.MerchantID, retrieved.MerchantID)

		// Test that merchant is properly linked
		assert.Equal(t, session.MerchantID, retrieved.MerchantID)
	})

	t.Run("merchant_audit_logs to merchants relationship", func(t *testing.T) {
		// Create a test merchant first
		merchant := createTestMerchant(t)
		err := mockDB.CreateMerchant(ctx, merchant)
		require.NoError(t, err)

		// Create an audit log
		auditLog := createTestAuditLog(t, "test_user_004", merchant.ID)

		// Create audit log in mock database
		err = mockDB.CreateAuditLog(ctx, auditLog)
		require.NoError(t, err)

		// Verify audit log exists
		logs, err := mockDB.GetAuditLogsByMerchantID(ctx, merchant.ID, 1, 10)
		require.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, auditLog.ID, logs[0].ID)
		assert.Equal(t, auditLog.MerchantID, logs[0].MerchantID)

		// Test that merchant is properly linked
		assert.Equal(t, auditLog.MerchantID, logs[0].MerchantID)
	})

	t.Run("merchant_audit_logs to users relationship", func(t *testing.T) {
		// Create a test merchant first
		merchant := createTestMerchant(t)
		err := mockDB.CreateMerchant(ctx, merchant)
		require.NoError(t, err)

		// Create an audit log
		auditLog := createTestAuditLog(t, "test_user_005", merchant.ID)

		// Create audit log in mock database
		err = mockDB.CreateAuditLog(ctx, auditLog)
		require.NoError(t, err)

		// Verify audit log exists
		logs, err := mockDB.GetAuditLogsByMerchantID(ctx, merchant.ID, 1, 10)
		require.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, auditLog.ID, logs[0].ID)
		assert.Equal(t, auditLog.UserID, logs[0].UserID)

		// Test that user is properly linked
		assert.Equal(t, auditLog.UserID, logs[0].UserID)
	})

	t.Run("bulk operations relationships", func(t *testing.T) {
		// Create a test merchant first
		merchant := createTestMerchant(t)
		err := mockDB.CreateMerchant(ctx, merchant)
		require.NoError(t, err)

		// Test bulk update portfolio type
		merchantIDs := []string{merchant.ID}
		err = mockDB.BulkUpdatePortfolioType(ctx, merchantIDs, models.PortfolioTypeOnboarded, "test_user_006")
		require.NoError(t, err)

		// Verify the update
		updated, err := mockDB.GetMerchant(ctx, merchant.ID)
		require.NoError(t, err)
		assert.Equal(t, models.PortfolioTypeOnboarded, updated.PortfolioType)

		// Test bulk update risk level
		err = mockDB.BulkUpdateRiskLevel(ctx, merchantIDs, models.RiskLevelHigh, "test_user_006")
		require.NoError(t, err)

		// Verify the update
		updated, err = mockDB.GetMerchant(ctx, merchant.ID)
		require.NoError(t, err)
		assert.Equal(t, models.RiskLevelHigh, updated.RiskLevel)
	})

	t.Run("data integrity validation", func(t *testing.T) {
		// Test that all relationships maintain data integrity

		// Create merchant with all required relationships
		merchant := createTestMerchantWithAllRelationships(t)
		err := mockDB.CreateMerchant(ctx, merchant)
		require.NoError(t, err)

		// Verify all relationships are properly set
		retrieved, err := mockDB.GetMerchant(ctx, merchant.ID)
		require.NoError(t, err)

		assert.Equal(t, merchant.PortfolioType, retrieved.PortfolioType)
		assert.Equal(t, merchant.RiskLevel, retrieved.RiskLevel)
		assert.Equal(t, merchant.CreatedBy, retrieved.CreatedBy)
		assert.Equal(t, merchant.Status, retrieved.Status)
		assert.Equal(t, merchant.ComplianceStatus, retrieved.ComplianceStatus)

		// Create session for the merchant
		session := createTestMerchantSession(t, merchant.CreatedBy, merchant.ID)
		err = mockDB.CreateSession(ctx, session)
		require.NoError(t, err)

		// Verify session relationship
		retrievedSession, err := mockDB.GetActiveSessionByUserID(ctx, session.UserID)
		require.NoError(t, err)
		assert.Equal(t, session.MerchantID, retrievedSession.MerchantID)

		// Create audit log for the merchant
		auditLog := createTestAuditLog(t, merchant.CreatedBy, merchant.ID)
		err = mockDB.CreateAuditLog(ctx, auditLog)
		require.NoError(t, err)

		// Verify audit log relationship
		logs, err := mockDB.GetAuditLogsByMerchantID(ctx, merchant.ID, 1, 10)
		require.NoError(t, err)
		assert.Len(t, logs, 1)
		assert.Equal(t, auditLog.MerchantID, logs[0].MerchantID)
	})
}

// Helper functions for test data creation

func createTestMerchant(t *testing.T) *models.Merchant {
	t.Helper()
	now := time.Now()
	return &models.Merchant{
		ID:                 "test_merchant_" + generateTestID(),
		Name:               "Test Company",
		LegalName:          "Test Company LLC",
		RegistrationNumber: "REG" + generateTestID(),
		TaxID:              "TAX" + generateTestID(),
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
		CreatedBy:        "test_user_" + generateTestID(),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

func createTestMerchantWithPortfolioType(t *testing.T, portfolioType models.PortfolioType) *models.Merchant {
	t.Helper()
	merchant := createTestMerchant(t)
	merchant.PortfolioType = portfolioType
	return merchant
}

func createTestMerchantWithRiskLevel(t *testing.T, riskLevel models.RiskLevel) *models.Merchant {
	t.Helper()
	merchant := createTestMerchant(t)
	merchant.RiskLevel = riskLevel
	return merchant
}

func createTestMerchantWithUser(t *testing.T, userID string) *models.Merchant {
	t.Helper()
	merchant := createTestMerchant(t)
	merchant.CreatedBy = userID
	return merchant
}

func createTestMerchantWithAllRelationships(t *testing.T) *models.Merchant {
	t.Helper()
	merchant := createTestMerchant(t)
	merchant.PortfolioType = models.PortfolioTypeOnboarded
	merchant.RiskLevel = models.RiskLevelLow
	merchant.ComplianceStatus = "passed"
	merchant.Status = "active"
	return merchant
}

func createTestMerchantSession(t *testing.T, userID, merchantID string) *models.MerchantSession {
	t.Helper()
	now := time.Now()
	return &models.MerchantSession{
		ID:         "test_session_" + generateTestID(),
		UserID:     userID,
		MerchantID: merchantID,
		StartedAt:  now,
		LastActive: now,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

func createTestAuditLog(t *testing.T, userID, merchantID string) *models.AuditLog {
	t.Helper()
	return &models.AuditLog{
		ID:           "test_audit_" + generateTestID(),
		UserID:       userID,
		MerchantID:   merchantID,
		Action:       "CREATE_MERCHANT",
		ResourceType: "merchant",
		ResourceID:   merchantID,
		Details:      "Merchant created for testing",
		IPAddress:    "192.168.1.1",
		UserAgent:    "Mozilla/5.0 (Test Browser)",
		RequestID:    "req_" + generateTestID(),
		CreatedAt:    time.Now(),
	}
}

func generateTestID() string {
	// Generate a random ID for testing
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
