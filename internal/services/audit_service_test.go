package services

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/compliance"
	"kyb-platform/internal/models"
	"kyb-platform/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockAuditRepository is a mock implementation of AuditRepository
type MockAuditRepository struct {
	mock.Mock
}

func (m *MockAuditRepository) SaveAuditLog(ctx context.Context, auditLog *models.AuditLog) error {
	args := m.Called(ctx, auditLog)
	return args.Error(0)
}

func (m *MockAuditRepository) GetAuditLogs(ctx context.Context, filters *AuditLogFilters) ([]*models.AuditLog, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) GetAuditLogByID(ctx context.Context, id string) (*models.AuditLog, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) GetAuditTrail(ctx context.Context, merchantID string, limit int, offset int) ([]*models.AuditLog, error) {
	args := m.Called(ctx, merchantID, limit, offset)
	return args.Get(0).([]*models.AuditLog), args.Error(1)
}

func (m *MockAuditRepository) SaveComplianceRecord(ctx context.Context, record *ComplianceRecord) error {
	args := m.Called(ctx, record)
	return args.Error(0)
}

func (m *MockAuditRepository) GetComplianceRecords(ctx context.Context, filters *ComplianceFilters) ([]*ComplianceRecord, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*ComplianceRecord), args.Error(1)
}

func (m *MockAuditRepository) GetComplianceStatus(ctx context.Context, merchantID string) (*ComplianceStatus, error) {
	args := m.Called(ctx, merchantID)
	return args.Get(0).(*ComplianceStatus), args.Error(1)
}

// MockComplianceSystem is a mock implementation of ComplianceAuditSystem
type MockComplianceSystem struct {
	mock.Mock
}

func (m *MockComplianceSystem) RecordAuditEvent(ctx context.Context, event *compliance.AuditEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockComplianceSystem) GetAuditEvents(ctx context.Context, filter *compliance.AuditFilter) ([]*compliance.AuditEvent, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*compliance.AuditEvent), args.Error(1)
}

func (m *MockComplianceSystem) GetAuditTrail(ctx context.Context, entityType, entityID string) (*compliance.ComplianceAuditTrail, error) {
	args := m.Called(ctx, entityType, entityID)
	return args.Get(0).(*compliance.ComplianceAuditTrail), args.Error(1)
}

func (m *MockComplianceSystem) GenerateAuditReport(ctx context.Context, filter *compliance.AuditFilter) (*compliance.AuditSummary, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*compliance.AuditSummary), args.Error(1)
}

func (m *MockComplianceSystem) GetAuditMetrics(ctx context.Context, filter *compliance.AuditFilter) (*compliance.AuditMetrics, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*compliance.AuditMetrics), args.Error(1)
}

func (m *MockComplianceSystem) UpdateAuditMetrics(ctx context.Context, metrics *compliance.AuditMetrics) error {
	args := m.Called(ctx, metrics)
	return args.Error(0)
}

func TestNewAuditService(t *testing.T) {
	logger := observability.NewLogger(zap.NewNop())
	compliance := &MockComplianceSystem{}
	repository := &MockAuditRepository{}

	service := NewAuditService(logger, compliance, repository)

	assert.NotNil(t, service)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, compliance, service.compliance)
	assert.Equal(t, repository, service.repository)
}

func TestAuditService_LogMerchantOperation(t *testing.T) {
	tests := []struct {
		name             string
		request          *LogMerchantOperationRequest
		repositoryErr    error
		complianceErr    error
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name: "successful operation logging",
			request: &LogMerchantOperationRequest{
				UserID:       "user123",
				MerchantID:   "merchant123",
				Action:       "create",
				ResourceType: "merchant",
				ResourceID:   "merchant123",
				Details:      "Created new merchant",
				Description:  "Merchant creation",
				IPAddress:    "192.168.1.1",
				UserAgent:    "test-agent",
				RequestID:    "req123",
				SessionID:    "session123",
				UserName:     "testuser",
				UserRole:     "admin",
				UserEmail:    "test@example.com",
				Metadata:     map[string]interface{}{"test": "value"},
			},
			repositoryErr: nil,
			complianceErr: nil,
			expectedError: false,
		},
		{
			name: "repository error",
			request: &LogMerchantOperationRequest{
				UserID:       "user123",
				MerchantID:   "merchant123",
				Action:       "create",
				ResourceType: "merchant",
				ResourceID:   "merchant123",
				Details:      "Created new merchant",
				Description:  "Merchant creation",
			},
			repositoryErr:    assert.AnError,
			complianceErr:    nil,
			expectedError:    true,
			expectedErrorMsg: "failed to save audit log",
		},
		{
			name: "compliance error (should not fail operation)",
			request: &LogMerchantOperationRequest{
				UserID:       "user123",
				MerchantID:   "merchant123",
				Action:       "create",
				ResourceType: "merchant",
				ResourceID:   "merchant123",
				Details:      "Created new merchant",
				Description:  "Merchant creation",
			},
			repositoryErr: nil,
			complianceErr: assert.AnError,
			expectedError: false, // Compliance error should not fail the operation
		},
		{
			name: "invalid request - missing user ID",
			request: &LogMerchantOperationRequest{
				MerchantID:   "merchant123",
				Action:       "create",
				ResourceType: "merchant",
				ResourceID:   "merchant123",
				Details:      "Created new merchant",
				Description:  "Merchant creation",
			},
			repositoryErr:    nil,
			complianceErr:    nil,
			expectedError:    true,
			expectedErrorMsg: "audit log validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAuditRepository{}
			mockCompliance := &MockComplianceSystem{}
			logger := observability.NewLogger(zap.NewNop())

			service := NewAuditService(logger, mockCompliance, mockRepo)

			// Setup expectations
			if tt.repositoryErr == nil && tt.request.UserID != "" {
				mockRepo.On("SaveAuditLog", mock.Anything, mock.AnythingOfType("*models.AuditLog")).Return(nil)
				mockCompliance.On("RecordAuditEvent", mock.Anything, mock.AnythingOfType("*compliance.AuditEvent")).Return(tt.complianceErr)
			} else if tt.repositoryErr != nil {
				mockRepo.On("SaveAuditLog", mock.Anything, mock.AnythingOfType("*models.AuditLog")).Return(tt.repositoryErr)
			}

			// Execute
			err := service.LogMerchantOperation(context.Background(), tt.request)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			if tt.repositoryErr == nil && tt.request.UserID != "" {
				mockCompliance.AssertExpectations(t)
			}
		})
	}
}

func TestAuditService_GetAuditTrail(t *testing.T) {
	tests := []struct {
		name           string
		merchantID     string
		limit          int
		offset         int
		repositoryData []*models.AuditLog
		repositoryErr  error
		expectedError  bool
		expectedCount  int
	}{
		{
			name:       "successful audit trail retrieval",
			merchantID: "merchant123",
			limit:      10,
			offset:     0,
			repositoryData: []*models.AuditLog{
				{
					ID:           "audit1",
					UserID:       "user1",
					MerchantID:   "merchant123",
					Action:       "create",
					ResourceType: "merchant",
					ResourceID:   "merchant123",
					Details:      "Created merchant",
					CreatedAt:    time.Now(),
				},
				{
					ID:           "audit2",
					UserID:       "user1",
					MerchantID:   "merchant123",
					Action:       "update",
					ResourceType: "merchant",
					ResourceID:   "merchant123",
					Details:      "Updated merchant",
					CreatedAt:    time.Now(),
				},
			},
			repositoryErr: nil,
			expectedError: false,
			expectedCount: 2,
		},
		{
			name:           "empty merchant ID",
			merchantID:     "",
			limit:          10,
			offset:         0,
			repositoryData: nil,
			repositoryErr:  nil,
			expectedError:  true,
			expectedCount:  0,
		},
		{
			name:           "repository error",
			merchantID:     "merchant123",
			limit:          10,
			offset:         0,
			repositoryData: nil,
			repositoryErr:  assert.AnError,
			expectedError:  true,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAuditRepository{}
			mockCompliance := &MockComplianceSystem{}
			logger := observability.NewLogger(zap.NewNop())

			service := NewAuditService(logger, mockCompliance, mockRepo)

			// Setup expectations
			if tt.merchantID != "" {
				mockRepo.On("GetAuditTrail", mock.Anything, tt.merchantID, tt.limit, tt.offset).Return(tt.repositoryData, tt.repositoryErr)
			}

			// Execute
			result, err := service.GetAuditTrail(context.Background(), tt.merchantID, tt.limit, tt.offset)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.expectedCount)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuditService_CreateComplianceRecord(t *testing.T) {
	tests := []struct {
		name             string
		request          *CreateComplianceRecordRequest
		repositoryErr    error
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name: "successful compliance record creation",
			request: &CreateComplianceRecordRequest{
				MerchantID:     "merchant123",
				ComplianceType: ComplianceTypeAML,
				Requirement:    "Customer Due Diligence",
				Description:    "Perform customer due diligence checks",
				Priority:       CompliancePriorityHigh,
				RiskLevel:      models.RiskLevelHigh,
				CreatedBy:      "user123",
			},
			repositoryErr: nil,
			expectedError: false,
		},
		{
			name: "repository error",
			request: &CreateComplianceRecordRequest{
				MerchantID:     "merchant123",
				ComplianceType: ComplianceTypeAML,
				Requirement:    "Customer Due Diligence",
				Description:    "Perform customer due diligence checks",
				Priority:       CompliancePriorityHigh,
				RiskLevel:      models.RiskLevelHigh,
				CreatedBy:      "user123",
			},
			repositoryErr:    assert.AnError,
			expectedError:    true,
			expectedErrorMsg: "failed to save compliance record",
		},
		{
			name: "invalid compliance type",
			request: &CreateComplianceRecordRequest{
				MerchantID:     "merchant123",
				ComplianceType: "invalid",
				Requirement:    "Customer Due Diligence",
				Description:    "Perform customer due diligence checks",
				Priority:       CompliancePriorityHigh,
				RiskLevel:      models.RiskLevelHigh,
				CreatedBy:      "user123",
			},
			repositoryErr:    nil,
			expectedError:    true,
			expectedErrorMsg: "compliance record validation failed",
		},
		{
			name: "missing merchant ID",
			request: &CreateComplianceRecordRequest{
				ComplianceType: ComplianceTypeAML,
				Requirement:    "Customer Due Diligence",
				Description:    "Perform customer due diligence checks",
				Priority:       CompliancePriorityHigh,
				RiskLevel:      models.RiskLevelHigh,
				CreatedBy:      "user123",
			},
			repositoryErr:    nil,
			expectedError:    true,
			expectedErrorMsg: "compliance record validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAuditRepository{}
			mockCompliance := &MockComplianceSystem{}
			logger := observability.NewLogger(zap.NewNop())

			service := NewAuditService(logger, mockCompliance, mockRepo)

			// Setup expectations
			if tt.repositoryErr == nil && tt.request.MerchantID != "" && tt.request.ComplianceType.IsValid() {
				mockRepo.On("SaveComplianceRecord", mock.Anything, mock.AnythingOfType("*services.ComplianceRecord")).Return(nil)
				mockRepo.On("SaveAuditLog", mock.Anything, mock.AnythingOfType("*models.AuditLog")).Return(nil)
				mockCompliance.On("RecordAuditEvent", mock.Anything, mock.AnythingOfType("*compliance.AuditEvent")).Return(nil)
			} else if tt.repositoryErr != nil && tt.request.MerchantID != "" && tt.request.ComplianceType.IsValid() {
				mockRepo.On("SaveComplianceRecord", mock.Anything, mock.AnythingOfType("*services.ComplianceRecord")).Return(tt.repositoryErr)
			}

			// Execute
			result, err := service.CreateComplianceRecord(context.Background(), tt.request)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.request.MerchantID, result.MerchantID)
				assert.Equal(t, tt.request.ComplianceType, result.ComplianceType)
				assert.Equal(t, tt.request.Requirement, result.Requirement)
				assert.Equal(t, ComplianceStatusPending, result.Status)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			if tt.repositoryErr == nil && tt.request.MerchantID != "" && tt.request.ComplianceType.IsValid() {
				mockCompliance.AssertExpectations(t)
			}
		})
	}
}

func TestAuditService_GetComplianceStatus(t *testing.T) {
	tests := []struct {
		name           string
		merchantID     string
		repositoryData *ComplianceStatus
		repositoryErr  error
		expectedError  bool
	}{
		{
			name:       "successful compliance status retrieval",
			merchantID: "merchant123",
			repositoryData: &ComplianceStatus{
				MerchantID:            "merchant123",
				OverallStatus:         ComplianceStatusCompleted,
				ComplianceScore:       0.85,
				TotalRequirements:     10,
				CompletedRequirements: 8,
				OverdueRequirements:   1,
				FailedRequirements:    1,
				RiskLevel:             models.RiskLevelMedium,
				LastAssessmentDate:    time.Now(),
				NextAssessmentDate:    time.Now().Add(30 * 24 * time.Hour),
			},
			repositoryErr: nil,
			expectedError: false,
		},
		{
			name:           "empty merchant ID",
			merchantID:     "",
			repositoryData: nil,
			repositoryErr:  nil,
			expectedError:  true,
		},
		{
			name:           "repository error",
			merchantID:     "merchant123",
			repositoryData: nil,
			repositoryErr:  assert.AnError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAuditRepository{}
			mockCompliance := &MockComplianceSystem{}
			logger := observability.NewLogger(zap.NewNop())

			service := NewAuditService(logger, mockCompliance, mockRepo)

			// Setup expectations
			if tt.merchantID != "" {
				mockRepo.On("GetComplianceStatus", mock.Anything, tt.merchantID).Return(tt.repositoryData, tt.repositoryErr)
			}

			// Execute
			result, err := service.GetComplianceStatus(context.Background(), tt.merchantID)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.repositoryData.MerchantID, result.MerchantID)
				assert.Equal(t, tt.repositoryData.OverallStatus, result.OverallStatus)
				assert.Equal(t, tt.repositoryData.ComplianceScore, result.ComplianceScore)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestAuditService_TrackFATFCompliance(t *testing.T) {
	tests := []struct {
		name           string
		merchantID     string
		recommendation *FATFRecommendation
		repositoryErr  error
		expectedError  bool
	}{
		{
			name:       "successful FATF compliance tracking",
			merchantID: "merchant123",
			recommendation: &FATFRecommendation{
				ID:             "fatf1",
				Recommendation: "Customer Due Diligence",
				Description:    "Implement customer due diligence procedures",
				Category:       "CDD",
				Priority:       CompliancePriorityHigh,
				Status:         ComplianceStatusPending,
				Implementation: "Implement CDD procedures",
				Evidence:       []string{"policy1", "procedure1"},
			},
			repositoryErr: nil,
			expectedError: false,
		},
		{
			name:           "empty merchant ID",
			merchantID:     "",
			recommendation: &FATFRecommendation{},
			repositoryErr:  nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAuditRepository{}
			mockCompliance := &MockComplianceSystem{}
			logger := observability.NewLogger(zap.NewNop())

			service := NewAuditService(logger, mockCompliance, mockRepo)

			// Setup expectations
			if tt.merchantID != "" {
				mockRepo.On("SaveComplianceRecord", mock.Anything, mock.AnythingOfType("*services.ComplianceRecord")).Return(nil)
				mockRepo.On("SaveAuditLog", mock.Anything, mock.AnythingOfType("*models.AuditLog")).Return(nil)
				mockCompliance.On("RecordAuditEvent", mock.Anything, mock.AnythingOfType("*compliance.AuditEvent")).Return(nil)
			}

			// Execute
			err := service.TrackFATFCompliance(context.Background(), tt.merchantID, tt.recommendation)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			if tt.merchantID != "" {
				mockCompliance.AssertExpectations(t)
			}
		})
	}
}

func TestComplianceType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		ct       ComplianceType
		expected bool
	}{
		{"valid AML", ComplianceTypeAML, true},
		{"valid KYC", ComplianceTypeKYC, true},
		{"valid KYB", ComplianceTypeKYB, true},
		{"valid FATF", ComplianceTypeFATF, true},
		{"valid GDPR", ComplianceTypeGDPR, true},
		{"valid SOX", ComplianceTypeSOX, true},
		{"valid PCI", ComplianceTypePCI, true},
		{"valid ISO27001", ComplianceTypeISO27001, true},
		{"valid SOC2", ComplianceTypeSOC2, true},
		{"valid Custom", ComplianceTypeCustom, true},
		{"invalid type", ComplianceType("invalid"), false},
		{"empty type", ComplianceType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.ct.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestComplianceStatusType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		cst      ComplianceStatusType
		expected bool
	}{
		{"valid pending", ComplianceStatusPending, true},
		{"valid in_progress", ComplianceStatusInProgress, true},
		{"valid completed", ComplianceStatusCompleted, true},
		{"valid overdue", ComplianceStatusOverdue, true},
		{"valid failed", ComplianceStatusFailed, true},
		{"valid waived", ComplianceStatusWaived, true},
		{"valid exempt", ComplianceStatusExempt, true},
		{"invalid status", ComplianceStatusType("invalid"), false},
		{"empty status", ComplianceStatusType(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cst.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompliancePriority_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		cp       CompliancePriority
		expected bool
	}{
		{"valid low", CompliancePriorityLow, true},
		{"valid medium", CompliancePriorityMedium, true},
		{"valid high", CompliancePriorityHigh, true},
		{"valid critical", CompliancePriorityCritical, true},
		{"invalid priority", CompliancePriority("invalid"), false},
		{"empty priority", CompliancePriority(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cp.IsValid()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompliancePriority_GetNumericValue(t *testing.T) {
	tests := []struct {
		name     string
		cp       CompliancePriority
		expected int
	}{
		{"low priority", CompliancePriorityLow, 1},
		{"medium priority", CompliancePriorityMedium, 2},
		{"high priority", CompliancePriorityHigh, 3},
		{"critical priority", CompliancePriorityCritical, 4},
		{"invalid priority", CompliancePriority("invalid"), 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cp.GetNumericValue()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAuditService_validateComplianceRecord(t *testing.T) {
	service := &AuditService{}

	tests := []struct {
		name     string
		record   *ComplianceRecord
		expected bool
	}{
		{
			name: "valid record",
			record: &ComplianceRecord{
				MerchantID:     "merchant123",
				ComplianceType: ComplianceTypeAML,
				Status:         ComplianceStatusPending,
				Priority:       CompliancePriorityHigh,
				RiskLevel:      models.RiskLevelHigh,
				Requirement:    "CDD",
				Description:    "Customer Due Diligence",
			},
			expected: true,
		},
		{
			name: "missing merchant ID",
			record: &ComplianceRecord{
				ComplianceType: ComplianceTypeAML,
				Status:         ComplianceStatusPending,
				Priority:       CompliancePriorityHigh,
				RiskLevel:      models.RiskLevelHigh,
				Requirement:    "CDD",
				Description:    "Customer Due Diligence",
			},
			expected: false,
		},
		{
			name: "invalid compliance type",
			record: &ComplianceRecord{
				MerchantID:     "merchant123",
				ComplianceType: "invalid",
				Status:         ComplianceStatusPending,
				Priority:       CompliancePriorityHigh,
				RiskLevel:      models.RiskLevelHigh,
				Requirement:    "CDD",
				Description:    "Customer Due Diligence",
			},
			expected: false,
		},
		{
			name: "invalid status",
			record: &ComplianceRecord{
				MerchantID:     "merchant123",
				ComplianceType: ComplianceTypeAML,
				Status:         "invalid",
				Priority:       CompliancePriorityHigh,
				RiskLevel:      models.RiskLevelHigh,
				Requirement:    "CDD",
				Description:    "Customer Due Diligence",
			},
			expected: false,
		},
		{
			name: "invalid priority",
			record: &ComplianceRecord{
				MerchantID:     "merchant123",
				ComplianceType: ComplianceTypeAML,
				Status:         ComplianceStatusPending,
				Priority:       "invalid",
				RiskLevel:      models.RiskLevelHigh,
				Requirement:    "CDD",
				Description:    "Customer Due Diligence",
			},
			expected: false,
		},
		{
			name: "invalid risk level",
			record: &ComplianceRecord{
				MerchantID:     "merchant123",
				ComplianceType: ComplianceTypeAML,
				Status:         ComplianceStatusPending,
				Priority:       CompliancePriorityHigh,
				RiskLevel:      "invalid",
				Requirement:    "CDD",
				Description:    "Customer Due Diligence",
			},
			expected: false,
		},
		{
			name: "missing requirement",
			record: &ComplianceRecord{
				MerchantID:     "merchant123",
				ComplianceType: ComplianceTypeAML,
				Status:         ComplianceStatusPending,
				Priority:       CompliancePriorityHigh,
				RiskLevel:      models.RiskLevelHigh,
				Description:    "Customer Due Diligence",
			},
			expected: false,
		},
		{
			name: "missing description",
			record: &ComplianceRecord{
				MerchantID:     "merchant123",
				ComplianceType: ComplianceTypeAML,
				Status:         ComplianceStatusPending,
				Priority:       CompliancePriorityHigh,
				RiskLevel:      models.RiskLevelHigh,
				Requirement:    "CDD",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateComplianceRecord(tt.record)
			if tt.expected {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestAuditService_GenerateComplianceReport(t *testing.T) {
	tests := []struct {
		name          string
		merchantID    string
		statusData    *ComplianceStatus
		statusErr     error
		auditData     []*models.AuditLog
		auditErr      error
		expectedError bool
	}{
		{
			name:       "successful report generation",
			merchantID: "merchant123",
			statusData: &ComplianceStatus{
				MerchantID:            "merchant123",
				OverallStatus:         ComplianceStatusCompleted,
				ComplianceScore:       0.85,
				TotalRequirements:     10,
				CompletedRequirements: 8,
				OverdueRequirements:   1,
				FailedRequirements:    1,
				RiskLevel:             models.RiskLevelMedium,
				LastAssessmentDate:    time.Now(),
				NextAssessmentDate:    time.Now().Add(30 * 24 * time.Hour),
			},
			statusErr: nil,
			auditData: []*models.AuditLog{
				{
					ID:           "audit1",
					UserID:       "user1",
					MerchantID:   "merchant123",
					Action:       "create",
					ResourceType: "compliance_record",
					ResourceID:   "record1",
					Details:      "Created compliance record",
					CreatedAt:    time.Now(),
				},
			},
			auditErr:      nil,
			expectedError: false,
		},
		{
			name:          "status retrieval error",
			merchantID:    "merchant123",
			statusData:    nil,
			statusErr:     assert.AnError,
			auditData:     nil,
			auditErr:      nil,
			expectedError: true,
		},
		{
			name:       "audit trail retrieval error",
			merchantID: "merchant123",
			statusData: &ComplianceStatus{
				MerchantID:            "merchant123",
				OverallStatus:         ComplianceStatusCompleted,
				ComplianceScore:       0.85,
				TotalRequirements:     10,
				CompletedRequirements: 8,
				OverdueRequirements:   1,
				FailedRequirements:    1,
				RiskLevel:             models.RiskLevelMedium,
				LastAssessmentDate:    time.Now(),
				NextAssessmentDate:    time.Now().Add(30 * 24 * time.Hour),
			},
			statusErr:     nil,
			auditData:     nil,
			auditErr:      assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockAuditRepository{}
			mockCompliance := &MockComplianceSystem{}
			logger := observability.NewLogger(zap.NewNop())

			service := NewAuditService(logger, mockCompliance, mockRepo)

			// Setup expectations
			mockRepo.On("GetComplianceStatus", mock.Anything, tt.merchantID).Return(tt.statusData, tt.statusErr)
			if tt.statusErr == nil {
				mockRepo.On("GetAuditTrail", mock.Anything, tt.merchantID, 100, 0).Return(tt.auditData, tt.auditErr)
			}

			// Execute
			result, err := service.GenerateComplianceReport(context.Background(), tt.merchantID)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.merchantID, result.MerchantID)
				assert.NotNil(t, result.ComplianceStatus)
				assert.NotNil(t, result.AuditTrail)
				assert.NotNil(t, result.Summary)
				assert.NotNil(t, result.Recommendations)
				assert.NotNil(t, result.RiskAssessment)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
		})
	}
}
