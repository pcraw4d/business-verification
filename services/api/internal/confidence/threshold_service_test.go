package confidence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockIndustryThresholdRepository is a mock implementation of IndustryThresholdRepository
type MockIndustryThresholdRepository struct {
	mock.Mock
}

func (m *MockIndustryThresholdRepository) GetIndustryThreshold(ctx context.Context, industryName string) (float64, error) {
	args := m.Called(ctx, industryName)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockIndustryThresholdRepository) GetAllIndustryThresholds(ctx context.Context) (map[string]float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockIndustryThresholdRepository) GetIndustryByID(ctx context.Context, industryID int) (*Industry, error) {
	args := m.Called(ctx, industryID)
	return args.Get(0).(*Industry), args.Error(1)
}

func (m *MockIndustryThresholdRepository) GetIndustryByName(ctx context.Context, industryName string) (*Industry, error) {
	args := m.Called(ctx, industryName)
	return args.Get(0).(*Industry), args.Error(1)
}

// TestIndustryThresholdService_GetIndustryThreshold tests the GetIndustryThreshold method
func TestIndustryThresholdService_GetIndustryThreshold(t *testing.T) {
	tests := []struct {
		name           string
		industryName   string
		mockSetup      func(*MockIndustryThresholdRepository)
		expectedResult float64
		expectedError  bool
	}{
		{
			name:         "successful threshold retrieval",
			industryName: "Restaurants",
			mockSetup: func(mockRepo *MockIndustryThresholdRepository) {
				mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
					ID:                  1,
					Name:                "Restaurants",
					ConfidenceThreshold: 0.75,
					IsActive:            true,
				}, nil)
			},
			expectedResult: 0.75,
			expectedError:  false,
		},
		{
			name:         "industry not found",
			industryName: "Unknown Industry",
			mockSetup: func(mockRepo *MockIndustryThresholdRepository) {
				mockRepo.On("GetIndustryByName", mock.Anything, "Unknown Industry").Return((*Industry)(nil), assert.AnError)
			},
			expectedResult: 0.50, // Default threshold
			expectedError:  false,
		},
		{
			name:         "inactive industry",
			industryName: "Inactive Industry",
			mockSetup: func(mockRepo *MockIndustryThresholdRepository) {
				mockRepo.On("GetIndustryByName", mock.Anything, "Inactive Industry").Return(&Industry{
					ID:                  2,
					Name:                "Inactive Industry",
					ConfidenceThreshold: 0.80,
					IsActive:            false,
				}, nil)
			},
			expectedResult: 0.50, // Default threshold for inactive industry
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockIndustryThresholdRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			// Create service
			service := NewIndustryThresholdService(mockRepo, nil)

			// Execute test
			result, err := service.GetIndustryThreshold(context.Background(), tt.industryName)

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

// TestIndustryThresholdService_GetAllIndustryThresholds tests the GetAllIndustryThresholds method
func TestIndustryThresholdService_GetAllIndustryThresholds(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockIndustryThresholdRepository)
		expectedResult map[string]float64
		expectedError  bool
	}{
		{
			name: "successful retrieval of all thresholds",
			mockSetup: func(mockRepo *MockIndustryThresholdRepository) {
				thresholds := map[string]float64{
					"Restaurants": 0.75,
					"Fast Food":   0.80,
					"Healthcare":  0.80,
				}
				mockRepo.On("GetAllIndustryThresholds", mock.Anything).Return(thresholds, nil)
			},
			expectedResult: map[string]float64{
				"Restaurants": 0.75,
				"Fast Food":   0.80,
				"Healthcare":  0.80,
			},
			expectedError: false,
		},
		{
			name: "database error",
			mockSetup: func(mockRepo *MockIndustryThresholdRepository) {
				mockRepo.On("GetAllIndustryThresholds", mock.Anything).Return((map[string]float64)(nil), assert.AnError)
			},
			expectedResult: map[string]float64{
				"Restaurants":      0.75,
				"Fast Food":        0.80,
				"Food & Beverage":  0.70,
				"Legal Services":   0.75,
				"Healthcare":       0.80,
				"Technology":       0.75,
				"Retail":           0.70,
				"Manufacturing":    0.70,
				"Construction":     0.70,
				"Transportation":   0.70,
				"Education":        0.75,
				"Entertainment":    0.65,
				"Agriculture":      0.70,
				"Energy":           0.70,
				"General Business": 0.50,
			}, // Default thresholds
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock
			mockRepo := &MockIndustryThresholdRepository{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			// Create service
			service := NewIndustryThresholdService(mockRepo, nil)

			// Execute test
			result, err := service.GetAllIndustryThresholds(context.Background())

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

// TestIndustryThresholdService_Caching tests the caching functionality
func TestIndustryThresholdService_Caching(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil).Twice() // Called twice: once initially, once after cache expires

	// Create service with short cache TTL
	service := NewIndustryThresholdService(mockRepo, nil)
	service.SetCacheTTL(100 * time.Millisecond)

	// First call - should hit database
	result1, err := service.GetIndustryThreshold(context.Background(), "Restaurants")
	require.NoError(t, err)
	assert.Equal(t, 0.75, result1)

	// Second call - should hit cache
	result2, err := service.GetIndustryThreshold(context.Background(), "Restaurants")
	require.NoError(t, err)
	assert.Equal(t, 0.75, result2)

	// Wait for cache to expire
	time.Sleep(150 * time.Millisecond)

	// Third call - should hit database again
	result3, err := service.GetIndustryThreshold(context.Background(), "Restaurants")
	require.NoError(t, err)
	assert.Equal(t, 0.75, result3)

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestIndustryThresholdService_RefreshCache tests the cache refresh functionality
func TestIndustryThresholdService_RefreshCache(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	thresholds := map[string]float64{
		"Restaurants": 0.75,
		"Fast Food":   0.80,
	}
	mockRepo.On("GetAllIndustryThresholds", mock.Anything).Return(thresholds, nil)

	// Create service
	service := NewIndustryThresholdService(mockRepo, nil)

	// Refresh cache
	err := service.RefreshCache(context.Background())
	require.NoError(t, err)

	// Verify cache was populated
	stats := service.GetCacheStats()
	assert.Equal(t, 2, stats["cache_size"])
	assert.False(t, stats["is_expired"].(bool))

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestIndustryThresholdService_GetCacheStats tests the cache statistics functionality
func TestIndustryThresholdService_GetCacheStats(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil)

	// Create service
	service := NewIndustryThresholdService(mockRepo, nil)

	// Initially, cache should be empty
	stats := service.GetCacheStats()
	assert.Equal(t, 0, stats["cache_size"])
	assert.True(t, stats["is_expired"].(bool))

	// Get a threshold to populate cache
	_, err := service.GetIndustryThreshold(context.Background(), "Restaurants")
	require.NoError(t, err)

	// Check cache stats after population
	stats = service.GetCacheStats()
	assert.Equal(t, 1, stats["cache_size"])
	assert.False(t, stats["is_expired"].(bool))
	assert.Contains(t, stats["cached_industries"].([]string), "Restaurants")

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestIndustryThresholdService_ClearCache tests the cache clearing functionality
func TestIndustryThresholdService_ClearCache(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil)

	// Create service
	service := NewIndustryThresholdService(mockRepo, nil)

	// Populate cache
	_, err := service.GetIndustryThreshold(context.Background(), "Restaurants")
	require.NoError(t, err)

	// Verify cache is populated
	stats := service.GetCacheStats()
	assert.Equal(t, 1, stats["cache_size"])

	// Clear cache
	service.ClearCache()

	// Verify cache is cleared
	stats = service.GetCacheStats()
	assert.Equal(t, 0, stats["cache_size"])

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestIndustryThresholdService_ValidateThreshold tests the threshold validation
func TestIndustryThresholdService_ValidateThreshold(t *testing.T) {
	service := NewIndustryThresholdService(nil, nil)

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
			err := service.ValidateThreshold(tt.threshold)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestIndustryThresholdService_GetThresholdRecommendation tests the threshold recommendation functionality
func TestIndustryThresholdService_GetThresholdRecommendation(t *testing.T) {
	service := NewIndustryThresholdService(nil, nil)

	tests := []struct {
		name           string
		industryName   string
		expectedResult float64
	}{
		{
			name:           "restaurant recommendation",
			industryName:   "Restaurants",
			expectedResult: 0.75,
		},
		{
			name:           "fast food recommendation",
			industryName:   "Fast Food",
			expectedResult: 0.80,
		},
		{
			name:           "healthcare recommendation",
			industryName:   "Healthcare",
			expectedResult: 0.80,
		},
		{
			name:           "unknown industry recommendation",
			industryName:   "Unknown Industry",
			expectedResult: 0.60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.GetThresholdRecommendation(tt.industryName)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

// TestIndustryThresholdService_ConcurrentAccess tests concurrent access to the service
func TestIndustryThresholdService_ConcurrentAccess(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil).Maybe() // Allow multiple calls due to concurrent access

	// Create service
	service := NewIndustryThresholdService(mockRepo, nil)

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			result, err := service.GetIndustryThreshold(context.Background(), "Restaurants")
			assert.NoError(t, err)
			assert.Equal(t, 0.75, result)
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}

// TestIndustryThresholdService_Performance tests the performance of threshold retrieval
func TestIndustryThresholdService_Performance(t *testing.T) {
	// Setup mock
	mockRepo := &MockIndustryThresholdRepository{}
	mockRepo.On("GetIndustryByName", mock.Anything, "Restaurants").Return(&Industry{
		ID:                  1,
		Name:                "Restaurants",
		ConfidenceThreshold: 0.75,
		IsActive:            true,
	}, nil)

	// Create service
	service := NewIndustryThresholdService(mockRepo, nil)

	// Measure performance
	start := time.Now()
	_, err := service.GetIndustryThreshold(context.Background(), "Restaurants")
	duration := time.Since(start)

	require.NoError(t, err)
	assert.Less(t, duration, 100*time.Millisecond, "Threshold retrieval should be fast")

	// Verify mock expectations
	mockRepo.AssertExpectations(t)
}
