package confidence

import (
	"context"
	"testing"

	"kyb-platform/internal/classification/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSupabaseRepository is a mock implementation of the SupabaseRepositoryInterface
type MockSupabaseRepository struct {
	mock.Mock
}

func (m *MockSupabaseRepository) GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*repository.Industry), args.Error(1)
}

func (m *MockSupabaseRepository) GetAllIndustries(ctx context.Context) ([]*repository.Industry, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*repository.Industry), args.Error(1)
}

func (m *MockSupabaseRepository) GetIndustryByID(ctx context.Context, id int) (*repository.Industry, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*repository.Industry), args.Error(1)
}

// TestSupabaseThresholdRepositoryAdapter_GetIndustryThreshold tests the GetIndustryThreshold method
func TestSupabaseThresholdRepositoryAdapter_GetIndustryThreshold(t *testing.T) {
	tests := []struct {
		name           string
		industryName   string
		mockSetup      func(*MockSupabaseRepository)
		expectedResult float64
		expectedError  bool
	}{
		{
			name:         "successful threshold retrieval",
			industryName: "Restaurants",
			mockSetup: func(mockRepo *MockSupabaseRepository) {
				mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&repository.Industry{
					ID:                  1,
					Name:                "Restaurants",
					Description:         "Restaurant industry",
					Category:            "food",
					ConfidenceThreshold: 0.75,
					IsActive:            true,
					CreatedAt:           "2023-01-01T00:00:00Z",
					UpdatedAt:           "2023-01-01T00:00:00Z",
				}, nil)
			},
			expectedResult: 0.75,
			expectedError:  false,
		},
		{
			name:         "industry not found",
			industryName: "Unknown Industry",
			mockSetup: func(mockRepo *MockSupabaseRepository) {
				mockRepo.On("GetIndustryByName", mock.Anything, "Unknown Industry").Return((*repository.Industry)(nil), assert.AnError)
			},
			expectedResult: 0.0,
			expectedError:  true,
		},
		{
			name:         "inactive industry",
			industryName: "Inactive Industry",
			mockSetup: func(mockRepo *MockSupabaseRepository) {
				mockRepo.On("GetIndustryByName", mock.Anything, "Inactive Industry").Return(&repository.Industry{
					ID:                  2,
					Name:                "Inactive Industry",
					Description:         "Inactive industry",
					Category:            "other",
					ConfidenceThreshold: 0.80,
					IsActive:            false,
					CreatedAt:           "2023-01-01T00:00:00Z",
					UpdatedAt:           "2023-01-01T00:00:00Z",
				}, nil)
			},
			expectedResult: 0.0,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockSupabaseRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			// Create adapter
			adapter := NewSupabaseThresholdRepositoryAdapter(mockRepo, nil)

			// Execute test
			result, err := adapter.GetIndustryThreshold(context.Background(), tt.industryName)

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestSupabaseThresholdRepositoryAdapter_GetAllIndustryThresholds tests the GetAllIndustryThresholds method
func TestSupabaseThresholdRepositoryAdapter_GetAllIndustryThresholds(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockSupabaseRepository)
		expectedResult map[string]float64
		expectedError  bool
	}{
		{
			name: "successful retrieval of all thresholds",
			mockSetup: func(mockRepo *MockSupabaseRepository) {
				industries := []*repository.Industry{
					{
						ID:                  1,
						Name:                "Restaurants",
						Description:         "Restaurant industry",
						Category:            "food",
						ConfidenceThreshold: 0.75,
						IsActive:            true,
						CreatedAt:           "2023-01-01T00:00:00Z",
						UpdatedAt:           "2023-01-01T00:00:00Z",
					},
					{
						ID:                  2,
						Name:                "Fast Food",
						Description:         "Fast food industry",
						Category:            "food",
						ConfidenceThreshold: 0.80,
						IsActive:            true,
						CreatedAt:           "2023-01-01T00:00:00Z",
						UpdatedAt:           "2023-01-01T00:00:00Z",
					},
					{
						ID:                  3,
						Name:                "Inactive Industry",
						Description:         "Inactive industry",
						Category:            "other",
						ConfidenceThreshold: 0.70,
						IsActive:            false, // Should be excluded
						CreatedAt:           "2023-01-01T00:00:00Z",
						UpdatedAt:           "2023-01-01T00:00:00Z",
					},
				}
				mockRepo.On("GetAllIndustries", mock.Anything).Return(industries, nil)
			},
			expectedResult: map[string]float64{
				"Restaurants": 0.75,
				"Fast Food":   0.80,
				// Inactive industry should be excluded
			},
			expectedError: false,
		},
		{
			name: "database error",
			mockSetup: func(mockRepo *MockSupabaseRepository) {
				mockRepo.On("GetAllIndustries", mock.Anything).Return(([]*repository.Industry)(nil), assert.AnError)
			},
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name: "empty industries list",
			mockSetup: func(mockRepo *MockSupabaseRepository) {
				mockRepo.On("GetAllIndustries", mock.Anything).Return([]*repository.Industry{}, nil)
			},
			expectedResult: map[string]float64{},
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockSupabaseRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			// Create adapter
			adapter := NewSupabaseThresholdRepositoryAdapter(mockRepo, nil)

			// Execute test
			result, err := adapter.GetAllIndustryThresholds(context.Background())

			// Assert results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestSupabaseThresholdRepositoryAdapter_GetIndustryByID tests the GetIndustryByID method
func TestSupabaseThresholdRepositoryAdapter_GetIndustryByID(t *testing.T) {
	// Setup mock
	mockRepo := &MockSupabaseRepository{}
	mockRepo.On("GetIndustryByID", mock.Anything, 1).Return(&repository.Industry{
		ID:                  1,
		Name:                "Restaurants",
		Description:         "Restaurant industry",
		Category:            "food",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
		CreatedAt:           "2023-01-01T00:00:00Z",
		UpdatedAt:           "2023-01-01T00:00:00Z",
	}, nil)

	// Create adapter
	adapter := NewSupabaseThresholdRepositoryAdapter(mockRepo, nil)

	// Execute test
	result, err := adapter.GetIndustryByID(context.Background(), 1)

	// Assert results
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "Restaurants", result.Name)
	assert.Equal(t, 0.75, result.ConfidenceThreshold)
	assert.True(t, result.IsActive)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestSupabaseThresholdRepositoryAdapter_GetIndustryByName tests the GetIndustryByName method
func TestSupabaseThresholdRepositoryAdapter_GetIndustryByName(t *testing.T) {
	// Setup mock
	mockRepo := &MockSupabaseRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&repository.Industry{
		ID:                  1,
		Name:                "Restaurants",
		Description:         "Restaurant industry",
		Category:            "food",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
		CreatedAt:           "2023-01-01T00:00:00Z",
		UpdatedAt:           "2023-01-01T00:00:00Z",
	}, nil)

	// Create adapter
	adapter := NewSupabaseThresholdRepositoryAdapter(mockRepo, nil)

	// Execute test
	result, err := adapter.GetIndustryByName(context.Background(), "Restaurants")

	// Assert results
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "Restaurants", result.Name)
	assert.Equal(t, 0.75, result.ConfidenceThreshold)
	assert.True(t, result.IsActive)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestSupabaseThresholdRepositoryAdapter_ValidateThreshold tests the ValidateThreshold method
func TestSupabaseThresholdRepositoryAdapter_ValidateThreshold(t *testing.T) {
	adapter := NewSupabaseThresholdRepositoryAdapter(nil, nil)

	tests := []struct {
		name        string
		threshold   float64
		expectError bool
	}{
		{
			name:        "valid threshold",
			threshold:   0.75,
			expectError: false,
		},
		{
			name:        "minimum valid threshold",
			threshold:   0.1,
			expectError: false,
		},
		{
			name:        "maximum valid threshold",
			threshold:   1.0,
			expectError: false,
		},
		{
			name:        "threshold too low",
			threshold:   0.05,
			expectError: true,
		},
		{
			name:        "threshold too high",
			threshold:   1.5,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adapter.ValidateThreshold(tt.threshold)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestSupabaseThresholdRepositoryAdapter_GetThresholdStatistics tests the GetThresholdStatistics method
func TestSupabaseThresholdRepositoryAdapter_GetThresholdStatistics(t *testing.T) {
	// Setup mock
	mockRepo := &MockSupabaseRepository{}
	industries := []*repository.Industry{
		{
			ID:                  1,
			Name:                "Restaurants",
			ConfidenceThreshold: 0.75,
			IsActive:            true,
		},
		{
			ID:                  2,
			Name:                "Fast Food",
			ConfidenceThreshold: 0.80,
			IsActive:            true,
		},
		{
			ID:                  3,
			Name:                "Healthcare",
			ConfidenceThreshold: 0.85,
			IsActive:            true,
		},
	}
	mockRepo.On("GetAllIndustries", mock.Anything).Return(industries, nil)

	// Create adapter
	adapter := NewSupabaseThresholdRepositoryAdapter(mockRepo, nil)

	// Execute test
	stats, err := adapter.GetThresholdStatistics(context.Background())

	// Assert results
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 3, stats["total_industries"])
	assert.InDelta(t, 0.80, stats["average_threshold"], 0.01) // (0.75 + 0.80 + 0.85) / 3
	assert.Equal(t, 0.75, stats["min_threshold"])
	assert.Equal(t, 0.85, stats["max_threshold"])
	assert.InDelta(t, 0.10, stats["threshold_range"], 0.01) // 0.85 - 0.75

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestSupabaseThresholdRepositoryAdapter_GetThresholdStatistics_Empty tests statistics with empty data
func TestSupabaseThresholdRepositoryAdapter_GetThresholdStatistics_Empty(t *testing.T) {
	// Setup mock
	mockRepo := &MockSupabaseRepository{}
	mockRepo.On("GetAllIndustries", mock.Anything).Return([]*repository.Industry{}, nil)

	// Create adapter
	adapter := NewSupabaseThresholdRepositoryAdapter(mockRepo, nil)

	// Execute test
	stats, err := adapter.GetThresholdStatistics(context.Background())

	// Assert results
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.Equal(t, 0, stats["total_industries"])
	assert.Equal(t, 0.0, stats["average_threshold"])
	assert.Equal(t, 0.0, stats["min_threshold"])
	assert.Equal(t, 0.0, stats["max_threshold"])

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestSupabaseThresholdRepositoryAdapter_ErrorHandling tests error handling
func TestSupabaseThresholdRepositoryAdapter_ErrorHandling(t *testing.T) {
	// Setup mock that returns an error
	mockRepo := &MockSupabaseRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Error Industry").Return((*repository.Industry)(nil), assert.AnError)

	// Create adapter
	adapter := NewSupabaseThresholdRepositoryAdapter(mockRepo, nil)

	// Execute test
	result, err := adapter.GetIndustryThreshold(context.Background(), "Error Industry")

	// Assert results
	assert.Error(t, err)
	assert.Equal(t, 0.0, result)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestSupabaseThresholdRepositoryAdapter_ConcurrentAccess tests concurrent access
func TestSupabaseThresholdRepositoryAdapter_ConcurrentAccess(t *testing.T) {
	// Setup mock
	mockRepo := &MockSupabaseRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&repository.Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil).Maybe() // Allow multiple calls due to concurrent access

	// Create adapter
	adapter := NewSupabaseThresholdRepositoryAdapter(mockRepo, nil)

	// Test concurrent access
	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			defer func() { done <- true }()
			result, err := adapter.GetIndustryThreshold(context.Background(), "Restaurants")
			assert.NoError(t, err)
			assert.Equal(t, 0.75, result)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}
