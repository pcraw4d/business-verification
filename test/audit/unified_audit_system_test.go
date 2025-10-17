package audit

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/models"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockUnifiedAuditRepository is a mock implementation of UnifiedAuditRepository
type MockUnifiedAuditRepository struct {
	mock.Mock
}

func (m *MockUnifiedAuditRepository) SaveAuditLog(ctx context.Context, auditLog *models.UnifiedAuditLog) error {
	args := m.Called(ctx, auditLog)
	return args.Error(0)
}

func (m *MockUnifiedAuditRepository) GetAuditLogs(ctx context.Context, filters *models.UnifiedAuditLogFilters) (*models.UnifiedAuditLogResult, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).(*models.UnifiedAuditLogResult), args.Error(1)
}

func (m *MockUnifiedAuditRepository) GetAuditLogByID(ctx context.Context, id string) (*models.UnifiedAuditLog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.UnifiedAuditLog), args.Error(1)
}

func (m *MockUnifiedAuditRepository) GetAuditTrail(ctx context.Context, merchantID string, limit int, offset int) ([]*models.UnifiedAuditLog, error) {
	args := m.Called(ctx, merchantID, limit, offset)
	return args.Get(0).([]*models.UnifiedAuditLog), args.Error(1)
}

func (m *MockUnifiedAuditRepository) GetAuditLogsByUser(ctx context.Context, userID string, limit int, offset int) ([]*models.UnifiedAuditLog, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*models.UnifiedAuditLog), args.Error(1)
}

func (m *MockUnifiedAuditRepository) GetAuditLogsByAction(ctx context.Context, action string, limit int, offset int) ([]*models.UnifiedAuditLog, error) {
	args := m.Called(ctx, action, limit, offset)
	return args.Get(0).([]*models.UnifiedAuditLog), args.Error(1)
}

func (m *MockUnifiedAuditRepository) DeleteOldAuditLogs(ctx context.Context, olderThan time.Duration) (int64, error) {
	args := m.Called(ctx, olderThan)
	return args.Get(0).(int64), args.Error(1)
}

// TestUnifiedAuditSystem tests the unified audit logging system
func TestUnifiedAuditSystem(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create mock repository
	mockRepo := &MockUnifiedAuditRepository{}

	// Create unified audit service
	auditService := &services.UnifiedAuditService{
		Logger:     logger,
		Repository: mockRepo,
	}

	t.Run("Log Merchant Operation", func(t *testing.T) {
		t.Log("Testing unified audit logging for merchant operations...")

		// Test data
		req := &services.LogMerchantOperationRequest{
			UserID:       "user-123",
			MerchantID:   "merchant-456",
			Action:       "CREATE",
			ResourceType: "merchant",
			ResourceID:   "merchant-456",
			Details: map[string]interface{}{
				"business_name": "Test Business",
				"industry":      "Technology",
			},
			IPAddress: "192.168.1.1",
			UserAgent: "Test Agent",
			RequestID: "req-789",
		}

		// Mock repository expectations
		mockRepo.On("SaveAuditLog", mock.Anything, mock.MatchedBy(func(log *models.UnifiedAuditLog) bool {
			return log.UserID == req.UserID &&
				log.MerchantID == req.MerchantID &&
				log.Action == req.Action &&
				log.ResourceType == req.ResourceType &&
				log.ResourceID == req.ResourceID
		})).Return(nil)

		// Execute test
		err := auditService.LogMerchantOperation(context.Background(), req)

		// Assertions
		assert.NoError(t, err, "Should successfully log merchant operation")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified audit logging for merchant operations passed")
	})

	t.Run("Get Audit Trail", func(t *testing.T) {
		t.Log("Testing unified audit trail retrieval...")

		merchantID := "merchant-456"
		limit := 10
		offset := 0

		// Mock audit logs
		expectedLogs := []*models.UnifiedAuditLog{
			{
				ID:            "audit-1",
				UserID:        "user-123",
				MerchantID:    merchantID,
				Action:        "CREATE",
				ResourceType:  "merchant",
				ResourceID:    merchantID,
				EventType:     "merchant_operation",
				EventCategory: "audit",
				CreatedAt:     time.Now(),
			},
			{
				ID:            "audit-2",
				UserID:        "user-123",
				MerchantID:    merchantID,
				Action:        "UPDATE",
				ResourceType:  "merchant",
				ResourceID:    merchantID,
				EventType:     "merchant_operation",
				EventCategory: "audit",
				CreatedAt:     time.Now().Add(-1 * time.Hour),
			},
		}

		// Mock repository expectations
		mockRepo.On("GetAuditTrail", mock.Anything, merchantID, limit, offset).Return(expectedLogs, nil)

		// Execute test
		logs, err := auditService.GetAuditTrail(context.Background(), merchantID, limit, offset)

		// Assertions
		assert.NoError(t, err, "Should successfully retrieve audit trail")
		assert.Len(t, logs, 2, "Should return 2 audit logs")
		assert.Equal(t, merchantID, logs[0].MerchantID, "First log should have correct merchant ID")
		assert.Equal(t, merchantID, logs[1].MerchantID, "Second log should have correct merchant ID")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified audit trail retrieval passed")
	})

	t.Run("Get Audit Logs by User", func(t *testing.T) {
		t.Log("Testing unified audit logs retrieval by user...")

		userID := "user-123"
		limit := 5
		offset := 0

		// Mock audit logs
		expectedLogs := []*models.UnifiedAuditLog{
			{
				ID:            "audit-1",
				UserID:        userID,
				Action:        "CREATE",
				ResourceType:  "merchant",
				EventType:     "merchant_operation",
				EventCategory: "audit",
				CreatedAt:     time.Now(),
			},
		}

		// Mock repository expectations
		mockRepo.On("GetAuditLogsByUser", mock.Anything, userID, limit, offset).Return(expectedLogs, nil)

		// Execute test
		logs, err := auditService.GetAuditLogsByUser(context.Background(), userID, limit, offset)

		// Assertions
		assert.NoError(t, err, "Should successfully retrieve audit logs by user")
		assert.Len(t, logs, 1, "Should return 1 audit log")
		assert.Equal(t, userID, logs[0].UserID, "Log should have correct user ID")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified audit logs retrieval by user passed")
	})

	t.Run("Get Audit Logs by Action", func(t *testing.T) {
		t.Log("Testing unified audit logs retrieval by action...")

		action := "CREATE"
		limit := 5
		offset := 0

		// Mock audit logs
		expectedLogs := []*models.UnifiedAuditLog{
			{
				ID:            "audit-1",
				Action:        action,
				ResourceType:  "merchant",
				EventType:     "merchant_operation",
				EventCategory: "audit",
				CreatedAt:     time.Now(),
			},
		}

		// Mock repository expectations
		mockRepo.On("GetAuditLogsByAction", mock.Anything, action, limit, offset).Return(expectedLogs, nil)

		// Execute test
		logs, err := auditService.GetAuditLogsByAction(context.Background(), action, limit, offset)

		// Assertions
		assert.NoError(t, err, "Should successfully retrieve audit logs by action")
		assert.Len(t, logs, 1, "Should return 1 audit log")
		assert.Equal(t, action, logs[0].Action, "Log should have correct action")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified audit logs retrieval by action passed")
	})

	t.Run("Delete Old Audit Logs", func(t *testing.T) {
		t.Log("Testing unified audit logs cleanup...")

		olderThan := 30 * 24 * time.Hour // 30 days
		expectedDeleted := int64(150)

		// Mock repository expectations
		mockRepo.On("DeleteOldAuditLogs", mock.Anything, olderThan).Return(expectedDeleted, nil)

		// Execute test
		deleted, err := auditService.DeleteOldAuditLogs(context.Background(), olderThan)

		// Assertions
		assert.NoError(t, err, "Should successfully delete old audit logs")
		assert.Equal(t, expectedDeleted, deleted, "Should return correct number of deleted logs")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified audit logs cleanup passed")
	})

	t.Run("Audit Log Validation", func(t *testing.T) {
		t.Log("Testing unified audit log validation...")

		// Test with invalid request
		invalidReq := &services.LogMerchantOperationRequest{
			// Missing required fields
			Action: "INVALID_ACTION",
		}

		// Execute test
		err := auditService.LogMerchantOperation(context.Background(), invalidReq)

		// Assertions
		assert.Error(t, err, "Should return error for invalid request")

		t.Log("✅ Unified audit log validation passed")
	})
}

// TestUnifiedAuditPerformance tests the performance of unified audit system
func TestUnifiedAuditPerformance(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create mock repository
	mockRepo := &MockUnifiedAuditRepository{}

	// Create unified audit service
	auditService := &services.UnifiedAuditService{
		Logger:     logger,
		Repository: mockRepo,
	}

	t.Run("Bulk Audit Logging Performance", func(t *testing.T) {
		t.Log("Testing bulk audit logging performance...")

		// Mock repository to simulate fast response
		mockRepo.On("SaveAuditLog", mock.Anything, mock.Anything).Return(nil)

		// Test bulk logging
		start := time.Now()
		for i := 0; i < 100; i++ {
			req := &services.LogMerchantOperationRequest{
				UserID:       "user-123",
				MerchantID:   "merchant-456",
				Action:       "READ",
				ResourceType: "merchant",
				ResourceID:   "merchant-456",
				RequestID:    "req-" + string(rune(i)),
			}

			err := auditService.LogMerchantOperation(context.Background(), req)
			assert.NoError(t, err, "Should successfully log audit entry")
		}
		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 5*time.Second, "Bulk logging should complete within 5 seconds")
		assert.Less(t, duration/100, 50*time.Millisecond, "Average logging time should be under 50ms")

		mockRepo.AssertExpectations(t)
		t.Logf("✅ Bulk audit logging performance: %v for 100 entries (avg: %v)", duration, duration/100)
	})

	t.Run("Audit Trail Query Performance", func(t *testing.T) {
		t.Log("Testing audit trail query performance...")

		merchantID := "merchant-456"
		limit := 100
		offset := 0

		// Mock large result set
		expectedLogs := make([]*models.UnifiedAuditLog, 100)
		for i := 0; i < 100; i++ {
			expectedLogs[i] = &models.UnifiedAuditLog{
				ID:            "audit-" + string(rune(i)),
				MerchantID:    merchantID,
				Action:        "READ",
				ResourceType:  "merchant",
				EventType:     "merchant_operation",
				EventCategory: "audit",
				CreatedAt:     time.Now().Add(-time.Duration(i) * time.Minute),
			}
		}

		// Mock repository expectations
		mockRepo.On("GetAuditTrail", mock.Anything, merchantID, limit, offset).Return(expectedLogs, nil)

		// Test query performance
		start := time.Now()
		logs, err := auditService.GetAuditTrail(context.Background(), merchantID, limit, offset)
		duration := time.Since(start)

		// Performance assertions
		assert.NoError(t, err, "Should successfully retrieve audit trail")
		assert.Len(t, logs, 100, "Should return 100 audit logs")
		assert.Less(t, duration, 1*time.Second, "Query should complete within 1 second")

		mockRepo.AssertExpectations(t)
		t.Logf("✅ Audit trail query performance: %v for 100 records", duration)
	})
}
