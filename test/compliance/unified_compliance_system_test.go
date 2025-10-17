package compliance

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

// MockUnifiedComplianceRepository is a mock implementation of UnifiedComplianceRepository
type MockUnifiedComplianceRepository struct {
	mock.Mock
}

func (m *MockUnifiedComplianceRepository) SaveComplianceTracking(ctx context.Context, tracking *models.ComplianceTracking) error {
	args := m.Called(ctx, tracking)
	return args.Error(0)
}

func (m *MockUnifiedComplianceRepository) GetComplianceTracking(ctx context.Context, merchantID, frameworkID string) (*models.ComplianceTracking, error) {
	args := m.Called(ctx, merchantID, frameworkID)
	return args.Get(0).(*models.ComplianceTracking), args.Error(1)
}

func (m *MockUnifiedComplianceRepository) UpdateComplianceTracking(ctx context.Context, tracking *models.ComplianceTracking) error {
	args := m.Called(ctx, tracking)
	return args.Error(0)
}

func (m *MockUnifiedComplianceRepository) GetComplianceTrackingByMerchant(ctx context.Context, merchantID string) ([]*models.ComplianceTracking, error) {
	args := m.Called(ctx, merchantID)
	return args.Get(0).([]*models.ComplianceTracking), args.Error(1)
}

func (m *MockUnifiedComplianceRepository) GetComplianceTrackingByFramework(ctx context.Context, frameworkID string) ([]*models.ComplianceTracking, error) {
	args := m.Called(ctx, frameworkID)
	return args.Get(0).([]*models.ComplianceTracking), args.Error(1)
}

func (m *MockUnifiedComplianceRepository) GetComplianceTrackingByStatus(ctx context.Context, status string) ([]*models.ComplianceTracking, error) {
	args := m.Called(ctx, status)
	return args.Get(0).([]*models.ComplianceTracking), args.Error(1)
}

func (m *MockUnifiedComplianceRepository) DeleteComplianceTracking(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// TestUnifiedComplianceSystem tests the unified compliance tracking system
func TestUnifiedComplianceSystem(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create mock repository
	mockRepo := &MockUnifiedComplianceRepository{}

	// Create unified compliance service
	complianceService := &services.UnifiedComplianceService{
		Logger:     logger,
		Repository: mockRepo,
	}

	t.Run("Create Compliance Tracking", func(t *testing.T) {
		t.Log("Testing unified compliance tracking creation...")

		// Test data
		req := &services.CreateComplianceTrackingRequest{
			MerchantID:          "merchant-456",
			ComplianceType:      "AML",
			ComplianceFramework: "FATF",
			CheckType:           "automated",
			Status:              "pending",
			Requirements: map[string]interface{}{
				"kyc_verification": true,
				"aml_screening":    true,
			},
			CheckMethod: "api_integration",
			Source:      "external_provider",
		}

		// Mock repository expectations
		mockRepo.On("SaveComplianceTracking", mock.Anything, mock.MatchedBy(func(tracking *models.ComplianceTracking) bool {
			return tracking.MerchantID == req.MerchantID &&
				tracking.ComplianceType == req.ComplianceType &&
				tracking.ComplianceFramework == req.ComplianceFramework &&
				tracking.Status == req.Status
		})).Return(nil)

		// Execute test
		tracking, err := complianceService.CreateComplianceTracking(context.Background(), req)

		// Assertions
		assert.NoError(t, err, "Should successfully create compliance tracking")
		assert.NotNil(t, tracking, "Should return compliance tracking")
		assert.Equal(t, req.MerchantID, tracking.MerchantID, "Should have correct merchant ID")
		assert.Equal(t, req.ComplianceType, tracking.ComplianceType, "Should have correct compliance type")
		assert.Equal(t, req.Status, tracking.Status, "Should have correct status")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified compliance tracking creation passed")
	})

	t.Run("Get Compliance Tracking", func(t *testing.T) {
		t.Log("Testing unified compliance tracking retrieval...")

		merchantID := "merchant-456"
		frameworkID := "FATF"

		// Mock compliance tracking
		expectedTracking := &models.ComplianceTracking{
			ID:                  "tracking-123",
			MerchantID:          merchantID,
			ComplianceType:      "AML",
			ComplianceFramework: frameworkID,
			CheckType:           "automated",
			Status:              "completed",
			Score:               0.85,
			RiskLevel:           "medium",
			Requirements: map[string]interface{}{
				"kyc_verification": true,
				"aml_screening":    true,
			},
			CheckMethod: "api_integration",
			Source:      "external_provider",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		// Mock repository expectations
		mockRepo.On("GetComplianceTracking", mock.Anything, merchantID, frameworkID).Return(expectedTracking, nil)

		// Execute test
		tracking, err := complianceService.GetComplianceTracking(context.Background(), merchantID, frameworkID)

		// Assertions
		assert.NoError(t, err, "Should successfully retrieve compliance tracking")
		assert.NotNil(t, tracking, "Should return compliance tracking")
		assert.Equal(t, merchantID, tracking.MerchantID, "Should have correct merchant ID")
		assert.Equal(t, frameworkID, tracking.ComplianceFramework, "Should have correct framework ID")
		assert.Equal(t, "completed", tracking.Status, "Should have correct status")
		assert.Equal(t, 0.85, tracking.Score, "Should have correct score")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified compliance tracking retrieval passed")
	})

	t.Run("Update Compliance Tracking", func(t *testing.T) {
		t.Log("Testing unified compliance tracking update...")

		// Test data
		req := &services.UpdateComplianceTrackingRequest{
			ID:                  "tracking-123",
			MerchantID:          "merchant-456",
			ComplianceType:      "AML",
			ComplianceFramework: "FATF",
			Status:              "completed",
			Score:               0.90,
			RiskLevel:           "low",
			Result: map[string]interface{}{
				"kyc_verified": true,
				"aml_cleared":  true,
			},
			Findings: []string{
				"All KYC requirements met",
				"AML screening passed",
			},
		}

		// Mock repository expectations
		mockRepo.On("UpdateComplianceTracking", mock.Anything, mock.MatchedBy(func(tracking *models.ComplianceTracking) bool {
			return tracking.ID == req.ID &&
				tracking.Status == req.Status &&
				tracking.Score == req.Score &&
				tracking.RiskLevel == req.RiskLevel
		})).Return(nil)

		// Execute test
		tracking, err := complianceService.UpdateComplianceTracking(context.Background(), req)

		// Assertions
		assert.NoError(t, err, "Should successfully update compliance tracking")
		assert.NotNil(t, tracking, "Should return updated compliance tracking")
		assert.Equal(t, req.Status, tracking.Status, "Should have updated status")
		assert.Equal(t, req.Score, tracking.Score, "Should have updated score")
		assert.Equal(t, req.RiskLevel, tracking.RiskLevel, "Should have updated risk level")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified compliance tracking update passed")
	})

	t.Run("Get Compliance Tracking by Merchant", func(t *testing.T) {
		t.Log("Testing unified compliance tracking retrieval by merchant...")

		merchantID := "merchant-456"

		// Mock compliance tracking list
		expectedTrackings := []*models.ComplianceTracking{
			{
				ID:                  "tracking-1",
				MerchantID:          merchantID,
				ComplianceType:      "AML",
				ComplianceFramework: "FATF",
				Status:              "completed",
				Score:               0.85,
			},
			{
				ID:                  "tracking-2",
				MerchantID:          merchantID,
				ComplianceType:      "KYC",
				ComplianceFramework: "FATF",
				Status:              "pending",
				Score:               0.0,
			},
		}

		// Mock repository expectations
		mockRepo.On("GetComplianceTrackingByMerchant", mock.Anything, merchantID).Return(expectedTrackings, nil)

		// Execute test
		trackings, err := complianceService.GetComplianceTrackingByMerchant(context.Background(), merchantID)

		// Assertions
		assert.NoError(t, err, "Should successfully retrieve compliance tracking by merchant")
		assert.Len(t, trackings, 2, "Should return 2 compliance trackings")
		assert.Equal(t, merchantID, trackings[0].MerchantID, "First tracking should have correct merchant ID")
		assert.Equal(t, merchantID, trackings[1].MerchantID, "Second tracking should have correct merchant ID")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified compliance tracking retrieval by merchant passed")
	})

	t.Run("Get Compliance Tracking by Framework", func(t *testing.T) {
		t.Log("Testing unified compliance tracking retrieval by framework...")

		frameworkID := "FATF"

		// Mock compliance tracking list
		expectedTrackings := []*models.ComplianceTracking{
			{
				ID:                  "tracking-1",
				ComplianceFramework: frameworkID,
				ComplianceType:      "AML",
				Status:              "completed",
				Score:               0.85,
			},
			{
				ID:                  "tracking-2",
				ComplianceFramework: frameworkID,
				ComplianceType:      "KYC",
				Status:              "pending",
				Score:               0.0,
			},
		}

		// Mock repository expectations
		mockRepo.On("GetComplianceTrackingByFramework", mock.Anything, frameworkID).Return(expectedTrackings, nil)

		// Execute test
		trackings, err := complianceService.GetComplianceTrackingByFramework(context.Background(), frameworkID)

		// Assertions
		assert.NoError(t, err, "Should successfully retrieve compliance tracking by framework")
		assert.Len(t, trackings, 2, "Should return 2 compliance trackings")
		assert.Equal(t, frameworkID, trackings[0].ComplianceFramework, "First tracking should have correct framework ID")
		assert.Equal(t, frameworkID, trackings[1].ComplianceFramework, "Second tracking should have correct framework ID")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified compliance tracking retrieval by framework passed")
	})

	t.Run("Get Compliance Tracking by Status", func(t *testing.T) {
		t.Log("Testing unified compliance tracking retrieval by status...")

		status := "pending"

		// Mock compliance tracking list
		expectedTrackings := []*models.ComplianceTracking{
			{
				ID:             "tracking-1",
				ComplianceType: "AML",
				Status:         status,
				Score:          0.0,
			},
			{
				ID:             "tracking-2",
				ComplianceType: "KYC",
				Status:         status,
				Score:          0.0,
			},
		}

		// Mock repository expectations
		mockRepo.On("GetComplianceTrackingByStatus", mock.Anything, status).Return(expectedTrackings, nil)

		// Execute test
		trackings, err := complianceService.GetComplianceTrackingByStatus(context.Background(), status)

		// Assertions
		assert.NoError(t, err, "Should successfully retrieve compliance tracking by status")
		assert.Len(t, trackings, 2, "Should return 2 compliance trackings")
		assert.Equal(t, status, trackings[0].Status, "First tracking should have correct status")
		assert.Equal(t, status, trackings[1].Status, "Second tracking should have correct status")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified compliance tracking retrieval by status passed")
	})

	t.Run("Delete Compliance Tracking", func(t *testing.T) {
		t.Log("Testing unified compliance tracking deletion...")

		trackingID := "tracking-123"

		// Mock repository expectations
		mockRepo.On("DeleteComplianceTracking", mock.Anything, trackingID).Return(nil)

		// Execute test
		err := complianceService.DeleteComplianceTracking(context.Background(), trackingID)

		// Assertions
		assert.NoError(t, err, "Should successfully delete compliance tracking")
		mockRepo.AssertExpectations(t)

		t.Log("✅ Unified compliance tracking deletion passed")
	})

	t.Run("Compliance Tracking Validation", func(t *testing.T) {
		t.Log("Testing unified compliance tracking validation...")

		// Test with invalid request
		invalidReq := &services.CreateComplianceTrackingRequest{
			// Missing required fields
			ComplianceType: "INVALID_TYPE",
		}

		// Execute test
		tracking, err := complianceService.CreateComplianceTracking(context.Background(), invalidReq)

		// Assertions
		assert.Error(t, err, "Should return error for invalid request")
		assert.Nil(t, tracking, "Should not return tracking for invalid request")

		t.Log("✅ Unified compliance tracking validation passed")
	})
}

// TestUnifiedCompliancePerformance tests the performance of unified compliance system
func TestUnifiedCompliancePerformance(t *testing.T) {
	// Setup test logger
	logger := observability.NewLogger(zap.NewNop())

	// Create mock repository
	mockRepo := &MockUnifiedComplianceRepository{}

	// Create unified compliance service
	complianceService := &services.UnifiedComplianceService{
		Logger:     logger,
		Repository: mockRepo,
	}

	t.Run("Bulk Compliance Tracking Performance", func(t *testing.T) {
		t.Log("Testing bulk compliance tracking performance...")

		// Mock repository to simulate fast response
		mockRepo.On("SaveComplianceTracking", mock.Anything, mock.Anything).Return(nil)

		// Test bulk creation
		start := time.Now()
		for i := 0; i < 50; i++ {
			req := &services.CreateComplianceTrackingRequest{
				MerchantID:          "merchant-" + string(rune(i)),
				ComplianceType:      "AML",
				ComplianceFramework: "FATF",
				CheckType:           "automated",
				Status:              "pending",
				CheckMethod:         "api_integration",
				Source:              "external_provider",
			}

			_, err := complianceService.CreateComplianceTracking(context.Background(), req)
			assert.NoError(t, err, "Should successfully create compliance tracking")
		}
		duration := time.Since(start)

		// Performance assertions
		assert.Less(t, duration, 3*time.Second, "Bulk creation should complete within 3 seconds")
		assert.Less(t, duration/50, 60*time.Millisecond, "Average creation time should be under 60ms")

		mockRepo.AssertExpectations(t)
		t.Logf("✅ Bulk compliance tracking performance: %v for 50 entries (avg: %v)", duration, duration/50)
	})

	t.Run("Compliance Tracking Query Performance", func(t *testing.T) {
		t.Log("Testing compliance tracking query performance...")

		merchantID := "merchant-456"
		frameworkID := "FATF"

		// Mock large result set
		expectedTracking := &models.ComplianceTracking{
			ID:                  "tracking-123",
			MerchantID:          merchantID,
			ComplianceType:      "AML",
			ComplianceFramework: frameworkID,
			Status:              "completed",
			Score:               0.85,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		// Mock repository expectations
		mockRepo.On("GetComplianceTracking", mock.Anything, merchantID, frameworkID).Return(expectedTracking, nil)

		// Test query performance
		start := time.Now()
		tracking, err := complianceService.GetComplianceTracking(context.Background(), merchantID, frameworkID)
		duration := time.Since(start)

		// Performance assertions
		assert.NoError(t, err, "Should successfully retrieve compliance tracking")
		assert.NotNil(t, tracking, "Should return compliance tracking")
		assert.Less(t, duration, 500*time.Millisecond, "Query should complete within 500ms")

		mockRepo.AssertExpectations(t)
		t.Logf("✅ Compliance tracking query performance: %v", duration)
	})
}
