package business_intelligence_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/your-org/your-repo/internal/modules/business_intelligence"
)

// MockBusinessDataAPI is a mock implementation of BusinessDataAPIService
type MockBusinessDataAPI struct {
	mock.Mock
}

func (m *MockBusinessDataAPI) SearchBusiness(ctx context.Context, query interface{}) (interface{}, error) {
	args := m.Called(ctx, query)
	return args.Get(0), args.Error(1)
}

// MockDataDiscoveryService is a mock implementation of DataDiscoveryService
type MockDataDiscoveryService struct {
	mock.Mock
}

func (m *MockDataDiscoveryService) DiscoverDataPoints(ctx context.Context, input interface{}) (interface{}, error) {
	args := m.Called(ctx, input)
	return args.Get(0), args.Error(1)
}

// MockMetricsCollector is a mock implementation of MetricsCollector
type MockMetricsCollector struct {
	mock.Mock
}

func (m *MockMetricsCollector) RecordMetric(name string, value float64, tags ...string) {
	m.Called(name, value, tags)
}

// MockErrorMonitor is a mock implementation of ErrorMonitor
type MockErrorMonitor struct {
	mock.Mock
}

func (m *MockErrorMonitor) RecordError(ctx context.Context, component string, err error) {
	m.Called(ctx, component, err)
}

func (m *MockErrorMonitor) RecordAlert(ctx context.Context, alertType string, message string) {
	m.Called(ctx, alertType, message)
}

// TestDataCollectionPipeline_Functionality tests the core functionality of the data collection pipeline
func TestDataCollectionPipeline_Functionality(t *testing.T) {
	logger := zap.NewNop()

	// Create mock dependencies
	mockBusinessDataAPI := &MockBusinessDataAPI{}
	mockDataDiscovery := &MockDataDiscoveryService{}
	mockMetrics := &MockMetricsCollector{}
	mockErrorMonitor := &MockErrorMonitor{}

	// Configure pipeline
	config := business_intelligence.DataCollectionPipelineConfig{
		EnableParallelCollection: true,
		MaxConcurrentCollections: 3,
		CollectionInterval:       30 * time.Second,
		EnableCaching:            true,
		CacheTTL:                 5 * time.Minute,
		EnableDataValidation:     true,
		QualityThreshold:         0.8,
		EnableMetrics:            true,
		EnableAlerting:           true,
		FallbackEnabled:          true,
		DataSources: []business_intelligence.DataSourceConfig{
			{
				Name:          "business_data_api",
				Type:          "business_data_api",
				Endpoint:      "https://api.example.com",
				APIKey:        "test-key",
				Weight:        1.0,
				RateLimit:     100,
				BurstLimit:    10,
				Timeout:       30 * time.Second,
				RetryAttempts: 3,
			},
			{
				Name:          "internal_db",
				Type:          "internal_db",
				Endpoint:      "postgres://localhost:5432/test",
				Weight:        0.8,
				RateLimit:     200,
				BurstLimit:    20,
				Timeout:       15 * time.Second,
				RetryAttempts: 2,
			},
		},
	}

	// Create pipeline
	pipeline := business_intelligence.NewDataCollectionPipeline(
		config,
		logger,
		mockBusinessDataAPI,
		mockDataDiscovery,
		mockMetrics,
		mockErrorMonitor,
	)

	t.Run("Successful Data Collection", func(t *testing.T) {
		// Setup mocks
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Test Company",
				"industry":      "Technology",
				"revenue":       1000000,
			}, nil)

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.95, "type": "completeness"},
					{"score": 0.88, "type": "accuracy"},
				},
			}, nil)

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()
		mockErrorMonitor.On("RecordError", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test data collection
		ctx := context.Background()
		businessID := "test-business-123"
		query := map[string]interface{}{
			"industry": "Technology",
			"location": "San Francisco",
		}

		collectedData, err := pipeline.CollectData(ctx, businessID, query)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, collectedData)
		assert.Greater(t, len(collectedData), 0)

		// Verify collected data structure
		for _, data := range collectedData {
			assert.NotEmpty(t, data.Source)
			assert.NotEmpty(t, data.DataType)
			assert.NotZero(t, data.Timestamp)
			assert.NotNil(t, data.RawData)
			assert.GreaterOrEqual(t, data.QualityScore, 0.0)
			assert.LessOrEqual(t, data.QualityScore, 1.0)
		}

		// Verify mock calls
		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})

	t.Run("Parallel Collection", func(t *testing.T) {
		// Reset mocks
		mockBusinessDataAPI.ExpectedCalls = nil
		mockDataDiscovery.ExpectedCalls = nil
		mockMetrics.ExpectedCalls = nil

		// Setup mocks for parallel collection
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Test Company 2",
				"industry":      "Finance",
			}, nil).Times(2) // Called for each data source

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.92, "type": "completeness"},
				},
			}, nil).Times(2)

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test parallel collection
		ctx := context.Background()
		businessID := "test-business-456"
		query := map[string]interface{}{
			"industry": "Finance",
		}

		startTime := time.Now()
		collectedData, err := pipeline.CollectData(ctx, businessID, query)
		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, collectedData)
		assert.Greater(t, len(collectedData), 0)

		// Verify parallel execution (should be faster than sequential)
		assert.Less(t, duration, 2*time.Second, "Parallel collection should be faster than sequential")

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})

	t.Run("Data Validation Failure", func(t *testing.T) {
		// Reset mocks
		mockBusinessDataAPI.ExpectedCalls = nil
		mockDataDiscovery.ExpectedCalls = nil
		mockMetrics.ExpectedCalls = nil
		mockErrorMonitor.ExpectedCalls = nil

		// Setup mocks for validation failure
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "", // Empty name should trigger validation failure
				"industry":      "Technology",
			}, nil)

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.3, "type": "completeness"}, // Low quality score
				},
			}, nil)

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()
		mockErrorMonitor.On("RecordAlert", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test data collection with validation failure
		ctx := context.Background()
		businessID := "test-business-789"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		collectedData, err := pipeline.CollectData(ctx, businessID, query)

		// Assertions
		require.NoError(t, err) // Collection should succeed even with validation failures
		assert.NotNil(t, collectedData)

		// Verify that validation failures are recorded
		for _, data := range collectedData {
			if !data.IsValid {
				assert.Less(t, data.QualityScore, 0.8, "Invalid data should have low quality score")
				assert.NotEmpty(t, data.Errors, "Invalid data should have error messages")
			}
		}

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
		mockErrorMonitor.AssertExpectations(t)
	})

	t.Run("Source Failure with Fallback", func(t *testing.T) {
		// Reset mocks
		mockBusinessDataAPI.ExpectedCalls = nil
		mockDataDiscovery.ExpectedCalls = nil
		mockMetrics.ExpectedCalls = nil
		mockErrorMonitor.ExpectedCalls = nil

		// Setup mocks for source failure
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			nil, assert.AnError).Once() // First source fails

		// Second source (internal_db) should succeed
		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.85, "type": "completeness"},
				},
			}, nil)

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()
		mockErrorMonitor.On("RecordError", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test data collection with source failure
		ctx := context.Background()
		businessID := "test-business-fallback"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		collectedData, err := pipeline.CollectData(ctx, businessID, query)

		// Assertions
		require.NoError(t, err) // Should succeed with fallback
		assert.NotNil(t, collectedData)
		assert.Greater(t, len(collectedData), 0, "Should have data from fallback sources")

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
		mockErrorMonitor.AssertExpectations(t)
	})

	t.Run("Cache Functionality", func(t *testing.T) {
		// Reset mocks
		mockBusinessDataAPI.ExpectedCalls = nil
		mockDataDiscovery.ExpectedCalls = nil
		mockMetrics.ExpectedCalls = nil

		// Setup mocks for cache test
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Cached Company",
				"industry":      "Technology",
			}, nil).Once() // Should only be called once due to caching

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.90, "type": "completeness"},
				},
			}, nil).Once()

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test cache functionality
		ctx := context.Background()
		businessID := "test-business-cache"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		// First call - should hit the API
		collectedData1, err1 := pipeline.CollectData(ctx, businessID, query)
		require.NoError(t, err1)
		assert.NotNil(t, collectedData1)

		// Second call - should hit cache
		collectedData2, err2 := pipeline.CollectData(ctx, businessID, query)
		require.NoError(t, err2)
		assert.NotNil(t, collectedData2)

		// Verify cache hit (API should only be called once)
		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})

	t.Run("Metrics Collection", func(t *testing.T) {
		// Reset mocks
		mockBusinessDataAPI.ExpectedCalls = nil
		mockDataDiscovery.ExpectedCalls = nil
		mockMetrics.ExpectedCalls = nil

		// Setup mocks
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Metrics Company",
				"industry":      "Technology",
			}, nil)

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.88, "type": "completeness"},
				},
			}, nil)

		// Expect specific metrics to be recorded
		mockMetrics.On("RecordMetric", "data_collection_duration", mock.AnythingOfType("float64"), "business_id", mock.Anything).Return()
		mockMetrics.On("RecordMetric", "data_sources_collected", mock.AnythingOfType("float64"), "business_id", mock.Anything).Return()
		mockMetrics.On("RecordMetric", "data_collection_errors", mock.AnythingOfType("float64"), "business_id", mock.Anything).Return()
		mockMetrics.On("RecordMetric", "data_collection_cache_hit", mock.AnythingOfType("float64"), "source", mock.Anything).Return()
		mockMetrics.On("RecordMetric", "data_collection_item_duration", mock.AnythingOfType("float64"), "source", mock.Anything).Return()

		// Test metrics collection
		ctx := context.Background()
		businessID := "test-business-metrics"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		collectedData, err := pipeline.CollectData(ctx, businessID, query)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, collectedData)

		// Verify all expected metrics were recorded
		mockMetrics.AssertExpectations(t)
	})
}

// TestDataCollectionPipeline_Configuration tests different configuration scenarios
func TestDataCollectionPipeline_Configuration(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Sequential Collection Mode", func(t *testing.T) {
		// Create mocks
		mockBusinessDataAPI := &MockBusinessDataAPI{}
		mockDataDiscovery := &MockDataDiscoveryService{}
		mockMetrics := &MockMetricsCollector{}
		mockErrorMonitor := &MockErrorMonitor{}

		// Configure for sequential collection
		config := business_intelligence.DataCollectionPipelineConfig{
			EnableParallelCollection: false, // Sequential mode
			MaxConcurrentCollections: 1,
			CollectionInterval:       30 * time.Second,
			EnableCaching:            false, // Disable caching for this test
			EnableDataValidation:     true,
			QualityThreshold:         0.8,
			EnableMetrics:            true,
			EnableAlerting:           false,
			FallbackEnabled:          true,
			DataSources: []business_intelligence.DataSourceConfig{
				{
					Name:          "source1",
					Type:          "business_data_api",
					Endpoint:      "https://api1.example.com",
					Weight:        1.0,
					RateLimit:     100,
					Timeout:       30 * time.Second,
					RetryAttempts: 3,
				},
				{
					Name:          "source2",
					Type:          "internal_db",
					Endpoint:      "postgres://localhost:5432/test",
					Weight:        0.8,
					RateLimit:     200,
					Timeout:       15 * time.Second,
					RetryAttempts: 2,
				},
			},
		}

		// Create pipeline
		pipeline := business_intelligence.NewDataCollectionPipeline(
			config,
			logger,
			mockBusinessDataAPI,
			mockDataDiscovery,
			mockMetrics,
			mockErrorMonitor,
		)

		// Setup mocks
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Sequential Company",
				"industry":      "Technology",
			}, nil)

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.90, "type": "completeness"},
				},
			}, nil)

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test sequential collection
		ctx := context.Background()
		businessID := "test-business-sequential"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		startTime := time.Now()
		collectedData, err := pipeline.CollectData(ctx, businessID, query)
		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, collectedData)
		assert.Greater(t, len(collectedData), 0)

		// Sequential collection should take longer than parallel
		assert.Greater(t, duration, 100*time.Millisecond, "Sequential collection should take more time")

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})

	t.Run("Disabled Validation", func(t *testing.T) {
		// Create mocks
		mockBusinessDataAPI := &MockBusinessDataAPI{}
		mockDataDiscovery := &MockDataDiscoveryService{}
		mockMetrics := &MockMetricsCollector{}
		mockErrorMonitor := &MockErrorMonitor{}

		// Configure with validation disabled
		config := business_intelligence.DataCollectionPipelineConfig{
			EnableParallelCollection: true,
			MaxConcurrentCollections: 2,
			EnableCaching:            false,
			EnableDataValidation:     false, // Disable validation
			QualityThreshold:         0.8,
			EnableMetrics:            true,
			EnableAlerting:           false,
			FallbackEnabled:          true,
			DataSources: []business_intelligence.DataSourceConfig{
				{
					Name:          "source1",
					Type:          "business_data_api",
					Endpoint:      "https://api.example.com",
					Weight:        1.0,
					RateLimit:     100,
					Timeout:       30 * time.Second,
					RetryAttempts: 3,
				},
			},
		}

		// Create pipeline
		pipeline := business_intelligence.NewDataCollectionPipeline(
			config,
			logger,
			mockBusinessDataAPI,
			mockDataDiscovery,
			mockMetrics,
			mockErrorMonitor,
		)

		// Setup mocks
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "", // Empty name - would normally fail validation
				"industry":      "Technology",
			}, nil)

		// Data discovery should not be called when validation is disabled
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test collection without validation
		ctx := context.Background()
		businessID := "test-business-no-validation"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		collectedData, err := pipeline.CollectData(ctx, businessID, query)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, collectedData)
		assert.Greater(t, len(collectedData), 0)

		// All data should be marked as valid when validation is disabled
		for _, data := range collectedData {
			assert.True(t, data.IsValid, "Data should be valid when validation is disabled")
			assert.Equal(t, 1.0, data.QualityScore, "Quality score should be 1.0 when validation is disabled")
			assert.Empty(t, data.Errors, "No errors should be present when validation is disabled")
		}

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertNotCalled(t, "DiscoverDataPoints")
		mockMetrics.AssertExpectations(t)
	})
}

// TestDataCollectionPipeline_ErrorHandling tests error handling scenarios
func TestDataCollectionPipeline_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()

	t.Run("All Sources Fail", func(t *testing.T) {
		// Create mocks
		mockBusinessDataAPI := &MockBusinessDataAPI{}
		mockDataDiscovery := &MockDataDiscoveryService{}
		mockMetrics := &MockMetricsCollector{}
		mockErrorMonitor := &MockErrorMonitor{}

		// Configure pipeline
		config := business_intelligence.DataCollectionPipelineConfig{
			EnableParallelCollection: true,
			MaxConcurrentCollections: 2,
			EnableCaching:            false,
			EnableDataValidation:     true,
			QualityThreshold:         0.8,
			EnableMetrics:            true,
			EnableAlerting:           true,
			FallbackEnabled:          false, // Disable fallback
			DataSources: []business_intelligence.DataSourceConfig{
				{
					Name:          "failing_source1",
					Type:          "business_data_api",
					Endpoint:      "https://failing-api1.example.com",
					Weight:        1.0,
					RateLimit:     100,
					Timeout:       5 * time.Second,
					RetryAttempts: 1,
				},
				{
					Name:          "failing_source2",
					Type:          "business_data_api",
					Endpoint:      "https://failing-api2.example.com",
					Weight:        0.8,
					RateLimit:     100,
					Timeout:       5 * time.Second,
					RetryAttempts: 1,
				},
			},
		}

		// Create pipeline
		pipeline := business_intelligence.NewDataCollectionPipeline(
			config,
			logger,
			mockBusinessDataAPI,
			mockDataDiscovery,
			mockMetrics,
			mockErrorMonitor,
		)

		// Setup mocks for all sources failing
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			nil, assert.AnError).Times(2) // Both sources fail

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()
		mockErrorMonitor.On("RecordError", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test collection with all sources failing
		ctx := context.Background()
		businessID := "test-business-all-fail"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		collectedData, err := pipeline.CollectData(ctx, businessID, query)

		// Assertions
		assert.Error(t, err, "Should return error when all sources fail")
		assert.Contains(t, err.Error(), "data collection completed with")
		assert.NotNil(t, collectedData) // Should still return partial data
		assert.Equal(t, 0, len(collectedData), "Should return empty data when all sources fail")

		mockBusinessDataAPI.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
		mockErrorMonitor.AssertExpectations(t)
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		// Create mocks
		mockBusinessDataAPI := &MockBusinessDataAPI{}
		mockDataDiscovery := &MockDataDiscoveryService{}
		mockMetrics := &MockMetricsCollector{}
		mockErrorMonitor := &MockErrorMonitor{}

		// Configure pipeline
		config := business_intelligence.DataCollectionPipelineConfig{
			EnableParallelCollection: true,
			MaxConcurrentCollections: 2,
			EnableCaching:            false,
			EnableDataValidation:     true,
			QualityThreshold:         0.8,
			EnableMetrics:            true,
			EnableAlerting:           false,
			FallbackEnabled:          true,
			DataSources: []business_intelligence.DataSourceConfig{
				{
					Name:          "slow_source",
					Type:          "business_data_api",
					Endpoint:      "https://slow-api.example.com",
					Weight:        1.0,
					RateLimit:     100,
					Timeout:       30 * time.Second,
					RetryAttempts: 3,
				},
			},
		}

		// Create pipeline
		pipeline := business_intelligence.NewDataCollectionPipeline(
			config,
			logger,
			mockBusinessDataAPI,
			mockDataDiscovery,
			mockMetrics,
			mockErrorMonitor,
		)

		// Setup mock for slow response
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Slow Company",
				"industry":      "Technology",
			}, nil).Run(func(args mock.Arguments) {
			// Simulate slow response
			time.Sleep(100 * time.Millisecond)
		})

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test with context cancellation
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		businessID := "test-business-cancellation"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		collectedData, err := pipeline.CollectData(ctx, businessID, query)

		// Assertions
		assert.Error(t, err, "Should return error due to context cancellation")
		assert.Contains(t, err.Error(), "context deadline exceeded")

		mockBusinessDataAPI.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})
}

// BenchmarkDataCollectionPipeline benchmarks the data collection pipeline performance
func BenchmarkDataCollectionPipeline(b *testing.B) {
	logger := zap.NewNop()

	// Create mocks
	mockBusinessDataAPI := &MockBusinessDataAPI{}
	mockDataDiscovery := &MockDataDiscoveryService{}
	mockMetrics := &MockMetricsCollector{}
	mockErrorMonitor := &MockErrorMonitor{}

	// Configure pipeline
	config := business_intelligence.DataCollectionPipelineConfig{
		EnableParallelCollection: true,
		MaxConcurrentCollections: 4,
		EnableCaching:            true,
		CacheTTL:                 5 * time.Minute,
		EnableDataValidation:     true,
		QualityThreshold:         0.8,
		EnableMetrics:            true,
		EnableAlerting:           false,
		FallbackEnabled:          true,
		DataSources: []business_intelligence.DataSourceConfig{
			{
				Name:          "benchmark_source1",
				Type:          "business_data_api",
				Endpoint:      "https://benchmark-api1.example.com",
				Weight:        1.0,
				RateLimit:     1000,
				Timeout:       10 * time.Second,
				RetryAttempts: 3,
			},
			{
				Name:          "benchmark_source2",
				Type:          "business_data_api",
				Endpoint:      "https://benchmark-api2.example.com",
				Weight:        0.9,
				RateLimit:     1000,
				Timeout:       10 * time.Second,
				RetryAttempts: 3,
			},
		},
	}

	// Create pipeline
	pipeline := business_intelligence.NewDataCollectionPipeline(
		config,
		logger,
		mockBusinessDataAPI,
		mockDataDiscovery,
		mockMetrics,
		mockErrorMonitor,
	)

	// Setup mocks
	mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"business_name": "Benchmark Company",
			"industry":      "Technology",
			"revenue":       1000000,
		}, nil)

	mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"quality_assessments": []map[string]interface{}{
				{"score": 0.95, "type": "completeness"},
				{"score": 0.88, "type": "accuracy"},
			},
		}, nil)

	mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

	b.ResetTimer()

	// Benchmark data collection
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		businessID := fmt.Sprintf("benchmark-business-%d", i)
		query := map[string]interface{}{
			"industry": "Technology",
		}

		_, err := pipeline.CollectData(ctx, businessID, query)
		if err != nil {
			b.Fatalf("Data collection failed: %v", err)
		}
	}
}
