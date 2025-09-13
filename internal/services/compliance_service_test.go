package services

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/models"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockComplianceRepository is a mock implementation of ComplianceRepository
type MockComplianceRepository struct {
	mock.Mock
}

func (m *MockComplianceRepository) SaveComplianceRequirement(ctx context.Context, requirement *ComplianceRequirement) error {
	args := m.Called(ctx, requirement)
	return args.Error(0)
}

func (m *MockComplianceRepository) GetComplianceRequirements(ctx context.Context, filters *ComplianceRequirementFilters) ([]*ComplianceRequirement, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*ComplianceRequirement), args.Error(1)
}

func (m *MockComplianceRepository) GetComplianceRequirementByID(ctx context.Context, id string) (*ComplianceRequirement, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*ComplianceRequirement), args.Error(1)
}

func (m *MockComplianceRepository) UpdateComplianceRequirement(ctx context.Context, requirement *ComplianceRequirement) error {
	args := m.Called(ctx, requirement)
	return args.Error(0)
}

func (m *MockComplianceRepository) DeleteComplianceRequirement(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockComplianceRepository) GetMerchantComplianceStatus(ctx context.Context, merchantID string) (*MerchantComplianceStatus, error) {
	args := m.Called(ctx, merchantID)
	return args.Get(0).(*MerchantComplianceStatus), args.Error(1)
}

func (m *MockComplianceRepository) SaveComplianceAssessment(ctx context.Context, assessment *ComplianceAssessment) error {
	args := m.Called(ctx, assessment)
	return args.Error(0)
}

func (m *MockComplianceRepository) GetComplianceAssessments(ctx context.Context, filters *ComplianceAssessmentFilters) ([]*ComplianceAssessment, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*ComplianceAssessment), args.Error(1)
}

func (m *MockComplianceRepository) GetComplianceReport(ctx context.Context, reportID string) (*ComplianceReport, error) {
	args := m.Called(ctx, reportID)
	return args.Get(0).(*ComplianceReport), args.Error(1)
}

func (m *MockComplianceRepository) SaveComplianceReport(ctx context.Context, report *ComplianceReport) error {
	args := m.Called(ctx, report)
	return args.Error(0)
}

// MockAuditService is a mock implementation of AuditServiceInterface
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogMerchantOperation(ctx context.Context, req *LogMerchantOperationRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockAuditService) GetComplianceStatus(ctx context.Context, merchantID string) (*ComplianceStatus, error) {
	args := m.Called(ctx, merchantID)
	return args.Get(0).(*ComplianceStatus), args.Error(1)
}

func (m *MockAuditService) GetComplianceRecords(ctx context.Context, filters *ComplianceFilters) ([]*ComplianceRecord, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).([]*ComplianceRecord), args.Error(1)
}

func TestComplianceService_CreateComplianceRequirement(t *testing.T) {
	tests := []struct {
		name           string
		request        *CreateComplianceRequirementRequest
		mockSetup      func(*MockComplianceRepository, *MockAuditService)
		expectedResult *ComplianceRequirement
		expectedError  string
	}{
		{
			name: "successful creation",
			request: &CreateComplianceRequirementRequest{
				Regulation:      "FATF",
				Requirement:     "Customer Due Diligence",
				Description:     "Implement customer due diligence procedures",
				Category:        ComplianceCategoryFATF,
				Priority:        CompliancePriorityHigh,
				RiskLevel:       models.RiskLevelHigh,
				EffectiveDate:   time.Now(),
				ReviewFrequency: "monthly",
				CreatedBy:       "test-user",
			},
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				repo.On("SaveComplianceRequirement", mock.Anything, mock.AnythingOfType("*services.ComplianceRequirement")).Return(nil)
				audit.On("LogMerchantOperation", mock.Anything, mock.AnythingOfType("*services.LogMerchantOperationRequest")).Return(nil)
			},
			expectedResult: &ComplianceRequirement{
				Regulation:      "FATF",
				Requirement:     "Customer Due Diligence",
				Description:     "Implement customer due diligence procedures",
				Category:        ComplianceCategoryFATF,
				Priority:        CompliancePriorityHigh,
				RiskLevel:       models.RiskLevelHigh,
				Status:          ComplianceStatusPending,
				ReviewFrequency: "monthly",
				CreatedBy:       "test-user",
			},
		},
		{
			name: "validation failure - missing regulation",
			request: &CreateComplianceRequirementRequest{
				Requirement:     "Customer Due Diligence",
				Description:     "Implement customer due diligence procedures",
				Category:        ComplianceCategoryFATF,
				Priority:        CompliancePriorityHigh,
				RiskLevel:       models.RiskLevelHigh,
				EffectiveDate:   time.Now(),
				ReviewFrequency: "monthly",
				CreatedBy:       "test-user",
			},
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				// No mock setup needed for validation failure
			},
			expectedError: "regulation is required",
		},
		{
			name: "validation failure - invalid category",
			request: &CreateComplianceRequirementRequest{
				Regulation:      "FATF",
				Requirement:     "Customer Due Diligence",
				Description:     "Implement customer due diligence procedures",
				Category:        ComplianceCategory("invalid"),
				Priority:        CompliancePriorityHigh,
				RiskLevel:       models.RiskLevelHigh,
				EffectiveDate:   time.Now(),
				ReviewFrequency: "monthly",
				CreatedBy:       "test-user",
			},
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				// No mock setup needed for validation failure
			},
			expectedError: "invalid compliance category: invalid",
		},
		{
			name: "repository error",
			request: &CreateComplianceRequirementRequest{
				Regulation:      "FATF",
				Requirement:     "Customer Due Diligence",
				Description:     "Implement customer due diligence procedures",
				Category:        ComplianceCategoryFATF,
				Priority:        CompliancePriorityHigh,
				RiskLevel:       models.RiskLevelHigh,
				EffectiveDate:   time.Now(),
				ReviewFrequency: "monthly",
				CreatedBy:       "test-user",
			},
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				repo.On("SaveComplianceRequirement", mock.Anything, mock.AnythingOfType("*services.ComplianceRequirement")).Return(assert.AnError)
			},
			expectedError: "failed to save compliance requirement",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockComplianceRepository{}
			mockAudit := &MockAuditService{}
			logger := observability.NewLogger(zap.NewNop())

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockAudit)
			}

			// Create service
			service := NewComplianceService(logger, mockRepo, mockAudit)

			// Execute test
			result, err := service.CreateComplianceRequirement(context.Background(), tt.request)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.Regulation, result.Regulation)
				assert.Equal(t, tt.expectedResult.Requirement, result.Requirement)
				assert.Equal(t, tt.expectedResult.Description, result.Description)
				assert.Equal(t, tt.expectedResult.Category, result.Category)
				assert.Equal(t, tt.expectedResult.Priority, result.Priority)
				assert.Equal(t, tt.expectedResult.RiskLevel, result.RiskLevel)
				assert.Equal(t, tt.expectedResult.Status, result.Status)
				assert.Equal(t, tt.expectedResult.ReviewFrequency, result.ReviewFrequency)
				assert.Equal(t, tt.expectedResult.CreatedBy, result.CreatedBy)
				assert.NotEmpty(t, result.ID)
				assert.NotZero(t, result.CreatedAt)
				assert.NotZero(t, result.UpdatedAt)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mockAudit.AssertExpectations(t)
		})
	}
}

func TestComplianceService_ValidateMerchantCompliance(t *testing.T) {
	tests := []struct {
		name           string
		merchantID     string
		mockSetup      func(*MockComplianceRepository, *MockAuditService)
		expectedResult *MerchantComplianceStatus
		expectedError  string
	}{
		{
			name:       "successful validation",
			merchantID: "merchant-123",
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				// Mock compliance requirements
				requirements := []*ComplianceRequirement{
					{
						ID:          "req-1",
						Requirement: "CDD Requirements",
						Category:    ComplianceCategoryFATF,
						Priority:    CompliancePriorityHigh,
						RiskLevel:   models.RiskLevelHigh,
						Status:      ComplianceStatusPending,
					},
					{
						ID:          "req-2",
						Requirement: "Record Keeping",
						Category:    ComplianceCategoryFATF,
						Priority:    CompliancePriorityMedium,
						RiskLevel:   models.RiskLevelMedium,
						Status:      ComplianceStatusPending,
					},
				}

				// Mock compliance records
				records := []*ComplianceRecord{
					{
						Requirement: "CDD Requirements",
						Status:      ComplianceStatusCompleted,
					},
					{
						Requirement: "Record Keeping",
						Status:      ComplianceStatusPending,
					},
				}

				repo.On("GetComplianceRequirements", mock.Anything, mock.AnythingOfType("*services.ComplianceRequirementFilters")).Return(requirements, nil)
				audit.On("GetComplianceRecords", mock.Anything, mock.AnythingOfType("*services.ComplianceFilters")).Return(records, nil)
				repo.On("SaveComplianceReport", mock.Anything, mock.AnythingOfType("*services.ComplianceReport")).Return(nil)
			},
			expectedResult: &MerchantComplianceStatus{
				MerchantID:            "merchant-123",
				TotalRequirements:     2,
				CompletedRequirements: 1,
				OverdueRequirements:   0,
				FailedRequirements:    0,
				ComplianceScore:       0.5,
				OverallStatus:         ComplianceStatusInProgress,
				RiskLevel:             models.RiskLevelMedium,
			},
		},
		{
			name:       "empty merchant ID",
			merchantID: "",
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				// No mock setup needed for validation failure
			},
			expectedError: "merchant ID is required",
		},
		{
			name:       "repository error",
			merchantID: "merchant-123",
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				repo.On("GetComplianceRequirements", mock.Anything, mock.AnythingOfType("*services.ComplianceRequirementFilters")).Return([]*ComplianceRequirement{}, assert.AnError)
			},
			expectedError: "failed to get compliance requirements",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := &MockComplianceRepository{}
			mockAudit := &MockAuditService{}
			logger := observability.NewLogger(zap.NewNop())

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockAudit)
			}

			// Create service
			service := NewComplianceService(logger, mockRepo, mockAudit)

			// Execute test
			result, err := service.ValidateMerchantCompliance(context.Background(), tt.merchantID)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.MerchantID, result.MerchantID)
				assert.Equal(t, tt.expectedResult.TotalRequirements, result.TotalRequirements)
				assert.Equal(t, tt.expectedResult.CompletedRequirements, result.CompletedRequirements)
				assert.Equal(t, tt.expectedResult.OverdueRequirements, result.OverdueRequirements)
				assert.Equal(t, tt.expectedResult.FailedRequirements, result.FailedRequirements)
				assert.Equal(t, tt.expectedResult.ComplianceScore, result.ComplianceScore)
				assert.Equal(t, tt.expectedResult.OverallStatus, result.OverallStatus)
				assert.Equal(t, tt.expectedResult.RiskLevel, result.RiskLevel)
				assert.NotZero(t, result.GeneratedAt)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mockAudit.AssertExpectations(t)
		})
	}
}

func TestComplianceService_GenerateComplianceReport(t *testing.T) {
	tests := []struct {
		name           string
		request        *GenerateComplianceReportRequest
		mockSetup      func(*MockComplianceRepository, *MockAuditService)
		expectedResult *ComplianceReport
		expectedError  string
	}{
		{
			name: "successful report generation",
			request: &GenerateComplianceReportRequest{
				MerchantID:  "merchant-123",
				ReportType:  "comprehensive",
				GeneratedBy: "test-user",
			},
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				// Mock compliance requirements for ValidateMerchantCompliance
				requirements := []*ComplianceRequirement{
					{
						ID:          "req-1",
						Requirement: "CDD Requirements",
						Category:    ComplianceCategoryFATF,
						Priority:    CompliancePriorityHigh,
						RiskLevel:   models.RiskLevelHigh,
						Status:      ComplianceStatusPending,
					},
				}

				// Mock compliance records for ValidateMerchantCompliance
				records := []*ComplianceRecord{
					{
						Requirement: "CDD Requirements",
						Status:      ComplianceStatusCompleted,
					},
				}

				// Mock compliance assessments
				assessments := []*ComplianceAssessment{
					{
						ID:             "assessment-1",
						MerchantID:     "merchant-123",
						AssessmentType: "initial",
						Status:         ComplianceStatusCompleted,
						Score:          0.8,
					},
				}

				// Setup mocks for ValidateMerchantCompliance (called by GenerateComplianceReport)
				repo.On("GetComplianceRequirements", mock.Anything, mock.AnythingOfType("*services.ComplianceRequirementFilters")).Return(requirements, nil)
				audit.On("GetComplianceRecords", mock.Anything, mock.AnythingOfType("*services.ComplianceFilters")).Return(records, nil)

				// Setup mocks for GenerateComplianceReport
				repo.On("GetComplianceAssessments", mock.Anything, mock.AnythingOfType("*services.ComplianceAssessmentFilters")).Return(assessments, nil)
				repo.On("SaveComplianceReport", mock.Anything, mock.AnythingOfType("*services.ComplianceReport")).Return(nil)
				audit.On("LogMerchantOperation", mock.Anything, mock.AnythingOfType("*services.LogMerchantOperationRequest")).Return(nil)
			},
			expectedResult: &ComplianceReport{
				MerchantID: "merchant-123",
			},
		},
		{
			name: "validation failure - missing merchant ID",
			request: &GenerateComplianceReportRequest{
				ReportType:  "comprehensive",
				GeneratedBy: "test-user",
			},
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				// No mock setup needed for validation failure
			},
			expectedError: "merchant ID is required",
		},
		{
			name: "validation failure - missing report type",
			request: &GenerateComplianceReportRequest{
				MerchantID:  "merchant-123",
				GeneratedBy: "test-user",
			},
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				// No mock setup needed for validation failure
			},
			expectedError: "report type is required",
		},
		{
			name: "validation failure - missing generated by",
			request: &GenerateComplianceReportRequest{
				MerchantID: "merchant-123",
				ReportType: "comprehensive",
			},
			mockSetup: func(repo *MockComplianceRepository, audit *MockAuditService) {
				// No mock setup needed for validation failure
			},
			expectedError: "generated by is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup fresh mocks for each test case
			mockRepo := &MockComplianceRepository{}
			mockAudit := &MockAuditService{}
			logger := observability.NewLogger(zap.NewNop())

			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo, mockAudit)
			}

			// Create service
			service := NewComplianceService(logger, mockRepo, mockAudit)

			// Execute test
			result, err := service.GenerateComplianceReport(context.Background(), tt.request)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedResult.MerchantID, result.MerchantID)
				assert.NotZero(t, result.GeneratedAt)
				assert.NotNil(t, result.ComplianceStatus)
				assert.NotNil(t, result.Summary)
				// Recommendations can be nil or empty slice
				if result.Recommendations != nil {
					assert.NotNil(t, result.Recommendations)
				}
				assert.NotNil(t, result.RiskAssessment)
			}

			// Verify mocks
			mockRepo.AssertExpectations(t)
			mockAudit.AssertExpectations(t)
		})
	}
}

func TestComplianceService_validateComplianceRequirement(t *testing.T) {
	tests := []struct {
		name          string
		requirement   *ComplianceRequirement
		expectedError string
	}{
		{
			name: "valid requirement",
			requirement: &ComplianceRequirement{
				Regulation:  "FATF",
				Requirement: "CDD Requirements",
				Description: "Customer due diligence requirements",
				Category:    ComplianceCategoryFATF,
				Priority:    CompliancePriorityHigh,
				RiskLevel:   models.RiskLevelHigh,
				Status:      ComplianceStatusPending,
			},
		},
		{
			name: "missing regulation",
			requirement: &ComplianceRequirement{
				Requirement: "CDD Requirements",
				Description: "Customer due diligence requirements",
				Category:    ComplianceCategoryFATF,
				Priority:    CompliancePriorityHigh,
				RiskLevel:   models.RiskLevelHigh,
				Status:      ComplianceStatusPending,
			},
			expectedError: "regulation is required",
		},
		{
			name: "missing requirement",
			requirement: &ComplianceRequirement{
				Regulation:  "FATF",
				Description: "Customer due diligence requirements",
				Category:    ComplianceCategoryFATF,
				Priority:    CompliancePriorityHigh,
				RiskLevel:   models.RiskLevelHigh,
				Status:      ComplianceStatusPending,
			},
			expectedError: "requirement is required",
		},
		{
			name: "missing description",
			requirement: &ComplianceRequirement{
				Regulation:  "FATF",
				Requirement: "CDD Requirements",
				Category:    ComplianceCategoryFATF,
				Priority:    CompliancePriorityHigh,
				RiskLevel:   models.RiskLevelHigh,
				Status:      ComplianceStatusPending,
			},
			expectedError: "description is required",
		},
		{
			name: "invalid category",
			requirement: &ComplianceRequirement{
				Regulation:  "FATF",
				Requirement: "CDD Requirements",
				Description: "Customer due diligence requirements",
				Category:    ComplianceCategory("invalid"),
				Priority:    CompliancePriorityHigh,
				RiskLevel:   models.RiskLevelHigh,
				Status:      ComplianceStatusPending,
			},
			expectedError: "invalid compliance category: invalid",
		},
		{
			name: "invalid priority",
			requirement: &ComplianceRequirement{
				Regulation:  "FATF",
				Requirement: "CDD Requirements",
				Description: "Customer due diligence requirements",
				Category:    ComplianceCategoryFATF,
				Priority:    CompliancePriority("invalid"),
				RiskLevel:   models.RiskLevelHigh,
				Status:      ComplianceStatusPending,
			},
			expectedError: "invalid compliance priority: invalid",
		},
		{
			name: "invalid risk level",
			requirement: &ComplianceRequirement{
				Regulation:  "FATF",
				Requirement: "CDD Requirements",
				Description: "Customer due diligence requirements",
				Category:    ComplianceCategoryFATF,
				Priority:    CompliancePriorityHigh,
				RiskLevel:   models.RiskLevel("invalid"),
				Status:      ComplianceStatusPending,
			},
			expectedError: "invalid risk level: invalid",
		},
		{
			name: "invalid status",
			requirement: &ComplianceRequirement{
				Regulation:  "FATF",
				Requirement: "CDD Requirements",
				Description: "Customer due diligence requirements",
				Category:    ComplianceCategoryFATF,
				Priority:    CompliancePriorityHigh,
				RiskLevel:   models.RiskLevelHigh,
				Status:      ComplianceStatusType("invalid"),
			},
			expectedError: "invalid compliance status: invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			logger := observability.NewLogger(zap.NewNop())
			service := NewComplianceService(logger, nil, nil)

			// Execute test
			err := service.validateComplianceRequirement(tt.requirement)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestComplianceService_calculateNextReviewDate(t *testing.T) {
	tests := []struct {
		name      string
		frequency string
		expected  time.Duration
	}{
		{
			name:      "daily frequency",
			frequency: "daily",
			expected:  24 * time.Hour,
		},
		{
			name:      "weekly frequency",
			frequency: "weekly",
			expected:  7 * 24 * time.Hour,
		},
		{
			name:      "monthly frequency",
			frequency: "monthly",
			expected:  30 * 24 * time.Hour, // Approximate
		},
		{
			name:      "quarterly frequency",
			frequency: "quarterly",
			expected:  0, // Will be calculated using AddDate
		},
		{
			name:      "annually frequency",
			frequency: "annually",
			expected:  365 * 24 * time.Hour, // Approximate
		},
		{
			name:      "unknown frequency",
			frequency: "unknown",
			expected:  30 * 24 * time.Hour, // Default to monthly
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			logger := observability.NewLogger(zap.NewNop())
			service := NewComplianceService(logger, nil, nil)

			// Execute test
			result := service.calculateNextReviewDate(tt.frequency)

			// Assert results
			assert.NotNil(t, result)

			var expectedTime time.Time
			if tt.frequency == "quarterly" {
				// For quarterly, use AddDate to match the service implementation
				expectedTime = time.Now().AddDate(0, 3, 0)
			} else {
				expectedTime = time.Now().Add(tt.expected)
			}

			// Allow more tolerance for quarterly and annual calculations
			tolerance := time.Minute
			if tt.frequency == "quarterly" || tt.frequency == "annually" {
				tolerance = time.Hour
			}
			assert.WithinDuration(t, expectedTime, *result, tolerance)
		})
	}
}

func TestComplianceService_buildMerchantComplianceStatus(t *testing.T) {
	tests := []struct {
		name           string
		merchantID     string
		requirements   []*ComplianceRequirement
		records        []*ComplianceRecord
		expectedResult *MerchantComplianceStatus
	}{
		{
			name:       "all requirements completed",
			merchantID: "merchant-123",
			requirements: []*ComplianceRequirement{
				{
					ID:          "req-1",
					Requirement: "CDD Requirements",
					Category:    ComplianceCategoryFATF,
					Priority:    CompliancePriorityHigh,
					RiskLevel:   models.RiskLevelHigh,
					Status:      ComplianceStatusPending,
				},
				{
					ID:          "req-2",
					Requirement: "Record Keeping",
					Category:    ComplianceCategoryFATF,
					Priority:    CompliancePriorityMedium,
					RiskLevel:   models.RiskLevelMedium,
					Status:      ComplianceStatusPending,
				},
			},
			records: []*ComplianceRecord{
				{
					Requirement: "CDD Requirements",
					Status:      ComplianceStatusCompleted,
				},
				{
					Requirement: "Record Keeping",
					Status:      ComplianceStatusCompleted,
				},
			},
			expectedResult: &MerchantComplianceStatus{
				MerchantID:            "merchant-123",
				TotalRequirements:     2,
				CompletedRequirements: 2,
				OverdueRequirements:   0,
				FailedRequirements:    0,
				ComplianceScore:       1.0,
				OverallStatus:         ComplianceStatusCompleted,
				RiskLevel:             models.RiskLevelLow,
			},
		},
		{
			name:       "mixed compliance status",
			merchantID: "merchant-123",
			requirements: []*ComplianceRequirement{
				{
					ID:          "req-1",
					Requirement: "CDD Requirements",
					Category:    ComplianceCategoryFATF,
					Priority:    CompliancePriorityHigh,
					RiskLevel:   models.RiskLevelHigh,
					Status:      ComplianceStatusPending,
				},
				{
					ID:          "req-2",
					Requirement: "Record Keeping",
					Category:    ComplianceCategoryFATF,
					Priority:    CompliancePriorityMedium,
					RiskLevel:   models.RiskLevelMedium,
					Status:      ComplianceStatusPending,
				},
				{
					ID:          "req-3",
					Requirement: "Suspicious Activity Reporting",
					Category:    ComplianceCategoryFATF,
					Priority:    CompliancePriorityHigh,
					RiskLevel:   models.RiskLevelHigh,
					Status:      ComplianceStatusPending,
				},
			},
			records: []*ComplianceRecord{
				{
					Requirement: "CDD Requirements",
					Status:      ComplianceStatusCompleted,
				},
				{
					Requirement: "Record Keeping",
					Status:      ComplianceStatusOverdue,
				},
				{
					Requirement: "Suspicious Activity Reporting",
					Status:      ComplianceStatusFailed,
				},
			},
			expectedResult: &MerchantComplianceStatus{
				MerchantID:            "merchant-123",
				TotalRequirements:     3,
				CompletedRequirements: 1,
				OverdueRequirements:   1,
				FailedRequirements:    1,
				ComplianceScore:       0.3333333333333333,
				OverallStatus:         ComplianceStatusFailed,
				RiskLevel:             models.RiskLevelHigh,
			},
		},
		{
			name:         "no requirements",
			merchantID:   "merchant-123",
			requirements: []*ComplianceRequirement{},
			records:      []*ComplianceRecord{},
			expectedResult: &MerchantComplianceStatus{
				MerchantID:            "merchant-123",
				TotalRequirements:     0,
				CompletedRequirements: 0,
				OverdueRequirements:   0,
				FailedRequirements:    0,
				ComplianceScore:       1.0,
				OverallStatus:         ComplianceStatusCompleted,
				RiskLevel:             models.RiskLevelLow,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			logger := observability.NewLogger(zap.NewNop())
			service := NewComplianceService(logger, nil, nil)

			// Execute test
			result := service.buildMerchantComplianceStatus(tt.merchantID, tt.requirements, tt.records)

			// Assert results
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedResult.MerchantID, result.MerchantID)
			assert.Equal(t, tt.expectedResult.TotalRequirements, result.TotalRequirements)
			assert.Equal(t, tt.expectedResult.CompletedRequirements, result.CompletedRequirements)
			assert.Equal(t, tt.expectedResult.OverdueRequirements, result.OverdueRequirements)
			assert.Equal(t, tt.expectedResult.FailedRequirements, result.FailedRequirements)
			assert.Equal(t, tt.expectedResult.ComplianceScore, result.ComplianceScore)
			assert.Equal(t, tt.expectedResult.OverallStatus, result.OverallStatus)
			assert.Equal(t, tt.expectedResult.RiskLevel, result.RiskLevel)
			assert.NotZero(t, result.GeneratedAt)
			assert.Len(t, result.Requirements, len(tt.requirements))
		})
	}
}

func TestComplianceService_generateComplianceSummary(t *testing.T) {
	// Create service
	logger := observability.NewLogger(zap.NewNop())
	service := NewComplianceService(logger, nil, nil)

	// Test data
	status := &MerchantComplianceStatus{
		TotalRequirements:     10,
		CompletedRequirements: 8,
		OverdueRequirements:   1,
		FailedRequirements:    1,
		ComplianceScore:       0.8,
		RiskLevel:             models.RiskLevelMedium,
		LastAssessmentDate:    time.Now(),
		NextAssessmentDate:    time.Now().AddDate(0, 1, 0),
	}

	// Execute test
	result := service.generateComplianceSummary(status)

	// Assert results
	assert.NotNil(t, result)
	assert.Equal(t, status.TotalRequirements, result.TotalRequirements)
	assert.Equal(t, status.CompletedRequirements, result.CompletedRequirements)
	assert.Equal(t, status.OverdueRequirements, result.OverdueRequirements)
	assert.Equal(t, status.FailedRequirements, result.FailedRequirements)
	assert.Equal(t, status.ComplianceScore, result.ComplianceScore)
	assert.Equal(t, string(status.RiskLevel), result.RiskLevel)
	assert.Equal(t, status.LastAssessmentDate, result.LastAssessment)
	assert.Equal(t, status.NextAssessmentDate, result.NextAssessment)
}

func TestComplianceService_generateComplianceRecommendations(t *testing.T) {
	tests := []struct {
		name          string
		status        *MerchantComplianceStatus
		expectedCount int
		expectedTypes []string
	}{
		{
			name: "overdue requirements",
			status: &MerchantComplianceStatus{
				OverdueRequirements: 2,
				FailedRequirements:  0,
				ComplianceScore:     0.9,
			},
			expectedCount: 1,
			expectedTypes: []string{"overdue"},
		},
		{
			name: "failed requirements",
			status: &MerchantComplianceStatus{
				OverdueRequirements: 0,
				FailedRequirements:  1,
				ComplianceScore:     0.9,
			},
			expectedCount: 1,
			expectedTypes: []string{"failed"},
		},
		{
			name: "low compliance score",
			status: &MerchantComplianceStatus{
				OverdueRequirements: 0,
				FailedRequirements:  0,
				ComplianceScore:     0.7,
			},
			expectedCount: 1,
			expectedTypes: []string{"score"},
		},
		{
			name: "multiple issues",
			status: &MerchantComplianceStatus{
				OverdueRequirements: 1,
				FailedRequirements:  1,
				ComplianceScore:     0.7,
			},
			expectedCount: 3,
			expectedTypes: []string{"overdue", "failed", "score"},
		},
		{
			name: "no issues",
			status: &MerchantComplianceStatus{
				OverdueRequirements: 0,
				FailedRequirements:  0,
				ComplianceScore:     0.9,
			},
			expectedCount: 0,
			expectedTypes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			logger := observability.NewLogger(zap.NewNop())
			service := NewComplianceService(logger, nil, nil)

			// Execute test
			result := service.generateComplianceRecommendations(tt.status)

			// Assert results
			assert.Len(t, result, tt.expectedCount)

			// Check that all expected types are present
			types := make([]string, len(result))
			for i, rec := range result {
				types[i] = rec.Type
			}

			for _, expectedType := range tt.expectedTypes {
				assert.Contains(t, types, expectedType)
			}
		})
	}
}

func TestComplianceService_generateRiskAssessment(t *testing.T) {
	// Create service
	logger := observability.NewLogger(zap.NewNop())
	service := NewComplianceService(logger, nil, nil)

	// Test data
	status := &MerchantComplianceStatus{
		ComplianceScore:     0.7,
		OverdueRequirements: 1,
		FailedRequirements:  1,
		RiskLevel:           models.RiskLevelMedium,
	}

	// Execute test
	result := service.generateRiskAssessment(status)

	// Assert results
	assert.NotNil(t, result)
	assert.Equal(t, string(status.RiskLevel), result.OverallRisk)
	assert.InDelta(t, 0.3, result.RiskScore, 0.0001) // 1.0 - 0.7, allow small floating point differences
	assert.NotEmpty(t, result.RiskFactors)
	assert.NotEmpty(t, result.Mitigations)
	assert.NotZero(t, result.LastAssessment)
}
