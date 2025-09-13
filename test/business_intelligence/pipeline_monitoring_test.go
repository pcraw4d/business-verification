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

	"github.com/pcraw4d/business-verification/internal/modules/business_intelligence"
)

// TestBusinessIntelligencePipeline_EndToEndMonitoring tests end-to-end pipeline monitoring
func TestBusinessIntelligencePipeline_EndToEndMonitoring(t *testing.T) {
	logger := zap.NewNop()

	// Create mock dependencies for the entire pipeline
	mockBusinessDataAPI := &MockBusinessDataAPI{}
	mockDataDiscovery := &MockDataDiscoveryService{}
	mockMetrics := &MockMetricsCollector{}
	mockErrorMonitor := &MockErrorMonitor{}

	// Configure data collection pipeline
	collectionConfig := business_intelligence.DataCollectionPipelineConfig{
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
		},
	}

	// Configure data processing service
	processingConfig := business_intelligence.DataProcessingServiceConfig{
		EnableParallelProcessing: true,
		MaxConcurrentProcessors:  3,
		EnableDataValidation:     true,
		EnableDataEnrichment:     true,
		EnableDataTransformation: true,
		EnableDataCleaning:       true,
		EnableDeduplication:      true,
		EnableMetrics:            true,
		EnableAlerting:           true,
	}

	// Configure data aggregation service
	aggregationConfig := business_intelligence.DataAggregationConfig{
		EnableParallelAggregation: true,
		MaxConcurrentAggregators:  3,
		EnableCorrelation:         true,
		CorrelationThreshold:      0.7,
		EnableAnalysis:            true,
		EnableTrendAnalysis:       true,
		EnablePatternAnalysis:     true,
		EnableAnomalyDetection:    true,
		EnableMetricCalculation:   true,
		EnableQualityAssessment:   true,
		QualityThreshold:          0.8,
		EnableMetrics:             true,
		EnablePerformanceTracking: true,
	}

	// Configure caching system
	cachingConfig := business_intelligence.CacheConfig{
		DefaultTTL:              5 * time.Minute,
		MaxCacheSize:            1000,
		EnableCompression:       true,
		EnableEncryption:        false,
		EnableSerialization:     true,
		EnableCacheWarming:      true,
		WarmingInterval:         1 * time.Minute,
		EnableCacheInvalidation: true,
		EnableMetrics:           true,
	}

	// Configure quality monitoring system
	qualityConfig := business_intelligence.QualityMonitoringConfig{
		EnableRealTimeMonitoring: true,
		MonitoringInterval:       30 * time.Second,
		EnableBatchMonitoring:    true,
		DefaultQualityThreshold:  0.8,
		CriticalQualityThreshold: 0.5,
		WarningQualityThreshold:  0.7,
		EnableAlerting:           true,
		AlertThreshold:           0.6,
		EnableReporting:          true,
		EnableParallelProcessing: true,
		MaxConcurrentMonitors:    5,
	}

	// Create pipeline components
	collectionPipeline := business_intelligence.NewDataCollectionPipeline(
		collectionConfig,
		logger,
		mockBusinessDataAPI,
		mockDataDiscovery,
		mockMetrics,
		mockErrorMonitor,
	)

	processingService := business_intelligence.NewDataProcessingService(
		processingConfig,
		logger,
		mockMetrics,
		mockErrorMonitor,
	)

	aggregationService := business_intelligence.NewDataAggregationService(
		aggregationConfig,
		logger,
		mockMetrics,
		mockErrorMonitor,
	)

	cachingSystem := business_intelligence.NewDataCachingSystem(cachingConfig, logger)

	qualitySystem := business_intelligence.NewDataQualityMonitoringSystem(qualityConfig, logger)

	t.Run("End-to-End Pipeline Execution", func(t *testing.T) {
		// Setup mocks for end-to-end execution
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "End-to-End Test Company",
				"industry":      "Technology",
				"revenue":       1000000,
				"employees":     50,
				"location":      "San Francisco, CA",
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

		// Execute end-to-end pipeline
		ctx := context.Background()
		businessID := "e2e-test-business-123"
		query := map[string]interface{}{
			"industry": "Technology",
			"location": "San Francisco",
		}

		// Step 1: Data Collection
		startTime := time.Now()
		collectedData, err := collectionPipeline.CollectData(ctx, businessID, query)
		collectionDuration := time.Since(startTime)

		require.NoError(t, err)
		assert.NotNil(t, collectedData)
		assert.Greater(t, len(collectedData), 0)

		// Step 2: Data Processing
		startTime = time.Now()
		processedData, err := processingService.ProcessCollectedData(ctx, collectedData)
		processingDuration := time.Since(startTime)

		require.NoError(t, err)
		assert.NotNil(t, processedData)
		assert.Equal(t, len(collectedData), len(processedData))

		// Step 3: Data Aggregation
		startTime = time.Now()
		aggregatedData, err := aggregationService.AggregateProcessedData(ctx, businessID, processedData)
		aggregationDuration := time.Since(startTime)

		require.NoError(t, err)
		assert.NotNil(t, aggregatedData)
		assert.Equal(t, businessID, aggregatedData.BusinessID)

		// Step 4: Data Quality Monitoring
		startTime = time.Now()
		qualityAssessment, err := qualitySystem.MonitorDataQuality(ctx, aggregatedData, businessID)
		qualityDuration := time.Since(startTime)

		require.NoError(t, err)
		assert.NotNil(t, qualityAssessment)
		assert.Equal(t, businessID, qualityAssessment.DataID)

		// Step 5: Caching
		startTime = time.Now()
		cacheKey := fmt.Sprintf("business_intelligence:%s", businessID)
		err = cachingSystem.Set(ctx, "business_intelligence", cacheKey, aggregatedData, 5*time.Minute)
		cachingDuration := time.Since(startTime)

		require.NoError(t, err)

		// Verify cached data
		cachedData, err := cachingSystem.Get(ctx, "business_intelligence", cacheKey)
		require.NoError(t, err)
		assert.NotNil(t, cachedData)

		// Performance assertions
		assert.Less(t, collectionDuration, 2*time.Second, "Data collection should be fast")
		assert.Less(t, processingDuration, 1*time.Second, "Data processing should be fast")
		assert.Less(t, aggregationDuration, 1*time.Second, "Data aggregation should be fast")
		assert.Less(t, qualityDuration, 500*time.Millisecond, "Quality monitoring should be fast")
		assert.Less(t, cachingDuration, 100*time.Millisecond, "Caching should be very fast")

		// Verify data flow integrity
		assert.Equal(t, len(collectedData), len(processedData), "Data count should be preserved through processing")
		assert.NotNil(t, aggregatedData.AggregatedMetrics, "Aggregated data should have metrics")
		assert.Greater(t, qualityAssessment.OverallScore, 0.0, "Quality assessment should have a score")

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})

	t.Run("Pipeline Error Recovery", func(t *testing.T) {
		// Reset mocks
		mockBusinessDataAPI.ExpectedCalls = nil
		mockDataDiscovery.ExpectedCalls = nil
		mockMetrics.ExpectedCalls = nil
		mockErrorMonitor.ExpectedCalls = nil

		// Setup mocks for error scenario
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			nil, assert.AnError).Once() // First call fails

		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Recovery Test Company",
				"industry":      "Technology",
				"revenue":       1000000,
			}, nil).Once() // Second call succeeds

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.85, "type": "completeness"},
				},
			}, nil)

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()
		mockErrorMonitor.On("RecordError", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test error recovery
		ctx := context.Background()
		businessID := "error-recovery-test-business"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		// Execute pipeline with error recovery
		collectedData, err := collectionPipeline.CollectData(ctx, businessID, query)

		// Should succeed with fallback
		require.NoError(t, err)
		assert.NotNil(t, collectedData)
		assert.Greater(t, len(collectedData), 0)

		// Continue with processing
		processedData, err := processingService.ProcessCollectedData(ctx, collectedData)
		require.NoError(t, err)
		assert.NotNil(t, processedData)

		// Continue with aggregation
		aggregatedData, err := aggregationService.AggregateProcessedData(ctx, businessID, processedData)
		require.NoError(t, err)
		assert.NotNil(t, aggregatedData)

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
		mockErrorMonitor.AssertExpectations(t)
	})

	t.Run("Pipeline Performance Monitoring", func(t *testing.T) {
		// Reset mocks
		mockBusinessDataAPI.ExpectedCalls = nil
		mockDataDiscovery.ExpectedCalls = nil
		mockMetrics.ExpectedCalls = nil

		// Setup mocks for performance monitoring
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Performance Test Company",
				"industry":      "Technology",
				"revenue":       1000000,
			}, nil)

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.90, "type": "completeness"},
				},
			}, nil)

		// Expect specific performance metrics
		mockMetrics.On("RecordMetric", "data_collection_duration", mock.AnythingOfType("float64"), "business_id", mock.Anything).Return()
		mockMetrics.On("RecordMetric", "data_processing_duration", mock.AnythingOfType("float64"), "items_count", mock.Anything).Return()
		mockMetrics.On("RecordMetric", "data_aggregation_duration", mock.AnythingOfType("float64"), "business_id", mock.Anything).Return()

		// Test performance monitoring
		ctx := context.Background()
		businessID := "performance-test-business"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		// Execute pipeline with performance monitoring
		collectedData, err := collectionPipeline.CollectData(ctx, businessID, query)
		require.NoError(t, err)

		processedData, err := processingService.ProcessCollectedData(ctx, collectedData)
		require.NoError(t, err)

		aggregatedData, err := aggregationService.AggregateProcessedData(ctx, businessID, processedData)
		require.NoError(t, err)

		// Verify performance metrics were recorded
		mockMetrics.AssertExpectations(t)

		// Verify aggregated data contains performance information
		assert.NotNil(t, aggregatedData.Metadata)
		assert.Contains(t, aggregatedData.Metadata, "performance_metrics")
	})

	t.Run("Pipeline Resource Management", func(t *testing.T) {
		// Reset mocks
		mockBusinessDataAPI.ExpectedCalls = nil
		mockDataDiscovery.ExpectedCalls = nil
		mockMetrics.ExpectedCalls = nil

		// Setup mocks for resource management test
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Resource Test Company",
				"industry":      "Technology",
				"revenue":       1000000,
			}, nil).Times(10) // Multiple calls

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.88, "type": "completeness"},
				},
			}, nil).Times(10)

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test resource management with concurrent operations
		ctx := context.Background()
		done := make(chan bool, 5)

		// Start multiple concurrent pipeline executions
		for i := 0; i < 5; i++ {
			go func(workerID int) {
				businessID := fmt.Sprintf("resource-test-business-%d", workerID)
				query := map[string]interface{}{
					"industry": "Technology",
				}

				// Execute pipeline
				collectedData, err := collectionPipeline.CollectData(ctx, businessID, query)
				if err != nil {
					t.Errorf("Collection failed for worker %d: %v", workerID, err)
					done <- false
					return
				}

				processedData, err := processingService.ProcessCollectedData(ctx, collectedData)
				if err != nil {
					t.Errorf("Processing failed for worker %d: %v", workerID, err)
					done <- false
					return
				}

				aggregatedData, err := aggregationService.AggregateProcessedData(ctx, businessID, processedData)
				if err != nil {
					t.Errorf("Aggregation failed for worker %d: %v", workerID, err)
					done <- false
					return
				}

				assert.NotNil(t, aggregatedData)
				done <- true
			}(i)
		}

		// Wait for all workers to complete
		successCount := 0
		for i := 0; i < 5; i++ {
			if <-done {
				successCount++
			}
		}

		// Verify all workers completed successfully
		assert.Equal(t, 5, successCount, "All concurrent pipeline executions should succeed")

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})
}

// TestBusinessIntelligencePipeline_IntegrationMonitoring tests integration monitoring
func TestBusinessIntelligencePipeline_IntegrationMonitoring(t *testing.T) {
	logger := zap.NewNop()

	// Create mock dependencies
	mockBusinessDataAPI := &MockBusinessDataAPI{}
	mockDataDiscovery := &MockDataDiscoveryService{}
	mockMetrics := &MockMetricsCollector{}
	mockErrorMonitor := &MockErrorMonitor{}

	// Configure pipeline components
	collectionConfig := business_intelligence.DataCollectionPipelineConfig{
		EnableParallelCollection: true,
		MaxConcurrentCollections: 2,
		EnableCaching:            true,
		CacheTTL:                 5 * time.Minute,
		EnableDataValidation:     true,
		QualityThreshold:         0.8,
		EnableMetrics:            true,
		EnableAlerting:           true,
		FallbackEnabled:          true,
		DataSources: []business_intelligence.DataSourceConfig{
			{
				Name:          "primary_source",
				Type:          "business_data_api",
				Endpoint:      "https://primary-api.example.com",
				Weight:        1.0,
				RateLimit:     100,
				Timeout:       30 * time.Second,
				RetryAttempts: 3,
			},
			{
				Name:          "secondary_source",
				Type:          "business_data_api",
				Endpoint:      "https://secondary-api.example.com",
				Weight:        0.8,
				RateLimit:     50,
				Timeout:       20 * time.Second,
				RetryAttempts: 2,
			},
		},
	}

	processingConfig := business_intelligence.DataProcessingServiceConfig{
		EnableParallelProcessing: true,
		MaxConcurrentProcessors:  2,
		EnableDataValidation:     true,
		EnableDataEnrichment:     true,
		EnableDataTransformation: true,
		EnableDataCleaning:       true,
		EnableDeduplication:      true,
		EnableMetrics:            true,
		EnableAlerting:           true,
	}

	aggregationConfig := business_intelligence.DataAggregationConfig{
		EnableParallelAggregation: true,
		MaxConcurrentAggregators:  2,
		EnableCorrelation:         true,
		CorrelationThreshold:      0.7,
		EnableAnalysis:            true,
		EnableTrendAnalysis:       true,
		EnablePatternAnalysis:     true,
		EnableAnomalyDetection:    true,
		EnableMetricCalculation:   true,
		EnableQualityAssessment:   true,
		QualityThreshold:          0.8,
		EnableMetrics:             true,
		EnablePerformanceTracking: true,
	}

	// Create pipeline components
	collectionPipeline := business_intelligence.NewDataCollectionPipeline(
		collectionConfig,
		logger,
		mockBusinessDataAPI,
		mockDataDiscovery,
		mockMetrics,
		mockErrorMonitor,
	)

	processingService := business_intelligence.NewDataProcessingService(
		processingConfig,
		logger,
		mockMetrics,
		mockErrorMonitor,
	)

	aggregationService := business_intelligence.NewDataAggregationService(
		aggregationConfig,
		logger,
		mockMetrics,
		mockErrorMonitor,
	)

	t.Run("Multi-Source Integration", func(t *testing.T) {
		// Setup mocks for multi-source integration
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Multi-Source Company",
				"industry":      "Technology",
				"revenue":       1000000,
				"employees":     50,
			}, nil).Times(2) // Called for each source

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.92, "type": "completeness"},
					{"score": 0.88, "type": "accuracy"},
				},
			}, nil).Times(2)

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()
		mockErrorMonitor.On("RecordError", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test multi-source integration
		ctx := context.Background()
		businessID := "multi-source-test-business"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		// Execute pipeline with multiple sources
		collectedData, err := collectionPipeline.CollectData(ctx, businessID, query)
		require.NoError(t, err)
		assert.NotNil(t, collectedData)
		assert.Equal(t, 2, len(collectedData), "Should collect data from both sources")

		// Verify data from different sources
		sources := make(map[string]bool)
		for _, data := range collectedData {
			sources[data.Source] = true
			assert.NotNil(t, data.RawData)
			assert.Greater(t, data.QualityScore, 0.0)
		}
		assert.True(t, sources["primary_source"], "Should have data from primary source")
		assert.True(t, sources["secondary_source"], "Should have data from secondary source")

		// Continue with processing and aggregation
		processedData, err := processingService.ProcessCollectedData(ctx, collectedData)
		require.NoError(t, err)
		assert.Equal(t, len(collectedData), len(processedData))

		aggregatedData, err := aggregationService.AggregateProcessedData(ctx, businessID, processedData)
		require.NoError(t, err)
		assert.NotNil(t, aggregatedData)

		// Verify aggregation includes data from multiple sources
		assert.NotNil(t, aggregatedData.AggregatedMetrics)
		assert.Contains(t, aggregatedData.Metadata, "source_count")
		assert.Equal(t, 2, aggregatedData.Metadata["source_count"])

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})

	t.Run("Data Flow Integrity", func(t *testing.T) {
		// Reset mocks
		mockBusinessDataAPI.ExpectedCalls = nil
		mockDataDiscovery.ExpectedCalls = nil
		mockMetrics.ExpectedCalls = nil

		// Setup mocks for data flow integrity test
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Integrity Test Company",
				"industry":      "Technology",
				"revenue":       1000000,
				"employees":     50,
				"location":      "San Francisco, CA",
				"founded":       "2010",
				"website":       "https://integrity-test.com",
			}, nil)

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.95, "type": "completeness"},
					{"score": 0.90, "type": "accuracy"},
					{"score": 0.88, "type": "consistency"},
				},
			}, nil)

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test data flow integrity
		ctx := context.Background()
		businessID := "integrity-test-business"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		// Execute pipeline and verify data integrity at each step
		collectedData, err := collectionPipeline.CollectData(ctx, businessID, query)
		require.NoError(t, err)
		assert.NotNil(t, collectedData)

		// Verify collected data integrity
		for _, data := range collectedData {
			assert.NotEmpty(t, data.Source)
			assert.NotEmpty(t, data.DataType)
			assert.NotZero(t, data.Timestamp)
			assert.NotNil(t, data.RawData)
			assert.GreaterOrEqual(t, data.QualityScore, 0.0)
			assert.LessOrEqual(t, data.QualityScore, 1.0)
		}

		// Process data and verify integrity
		processedData, err := processingService.ProcessCollectedData(ctx, collectedData)
		require.NoError(t, err)
		assert.Equal(t, len(collectedData), len(processedData))

		// Verify processed data integrity
		for i, processed := range processedData {
			original := collectedData[i]
			assert.Equal(t, original.Source, processed.OriginalSource)
			assert.Equal(t, original.DataType, processed.DataType)
			assert.Equal(t, original.Timestamp, processed.Timestamp)
			assert.NotNil(t, processed.ProcessedContent)
			assert.NotNil(t, processed.ValidationStatus)
			assert.NotNil(t, processed.EnrichmentStatus)
			assert.NotNil(t, processed.TransformationStatus)
		}

		// Aggregate data and verify integrity
		aggregatedData, err := aggregationService.AggregateProcessedData(ctx, businessID, processedData)
		require.NoError(t, err)
		assert.NotNil(t, aggregatedData)

		// Verify aggregated data integrity
		assert.Equal(t, businessID, aggregatedData.BusinessID)
		assert.NotNil(t, aggregatedData.AggregatedMetrics)
		assert.NotZero(t, aggregatedData.LastAggregated)
		assert.NotNil(t, aggregatedData.Metadata)
		assert.NotNil(t, aggregatedData.QualityAssessment)

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
	})

	t.Run("Error Propagation and Recovery", func(t *testing.T) {
		// Reset mocks
		mockBusinessDataAPI.ExpectedCalls = nil
		mockDataDiscovery.ExpectedCalls = nil
		mockMetrics.ExpectedCalls = nil
		mockErrorMonitor.ExpectedCalls = nil

		// Setup mocks for error propagation test
		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			nil, assert.AnError).Once() // First source fails

		mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"business_name": "Recovery Test Company",
				"industry":      "Technology",
				"revenue":       1000000,
			}, nil).Once() // Second source succeeds

		mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
			map[string]interface{}{
				"quality_assessments": []map[string]interface{}{
					{"score": 0.85, "type": "completeness"},
				},
			}, nil)

		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()
		mockErrorMonitor.On("RecordError", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test error propagation and recovery
		ctx := context.Background()
		businessID := "error-propagation-test-business"
		query := map[string]interface{}{
			"industry": "Technology",
		}

		// Execute pipeline with error recovery
		collectedData, err := collectionPipeline.CollectData(ctx, businessID, query)

		// Should succeed with partial data
		require.NoError(t, err)
		assert.NotNil(t, collectedData)
		assert.Equal(t, 1, len(collectedData), "Should have data from successful source")

		// Continue with processing
		processedData, err := processingService.ProcessCollectedData(ctx, collectedData)
		require.NoError(t, err)
		assert.NotNil(t, processedData)

		// Continue with aggregation
		aggregatedData, err := aggregationService.AggregateProcessedData(ctx, businessID, processedData)
		require.NoError(t, err)
		assert.NotNil(t, aggregatedData)

		// Verify error information is preserved
		assert.NotEmpty(t, aggregatedData.Errors, "Should have error information")
		assert.Contains(t, aggregatedData.Errors[0], "data collection completed with")

		mockBusinessDataAPI.AssertExpectations(t)
		mockDataDiscovery.AssertExpectations(t)
		mockMetrics.AssertExpectations(t)
		mockErrorMonitor.AssertExpectations(t)
	})
}

// BenchmarkBusinessIntelligencePipeline benchmarks the entire pipeline performance
func BenchmarkBusinessIntelligencePipeline(b *testing.B) {
	logger := zap.NewNop()

	// Create mock dependencies
	mockBusinessDataAPI := &MockBusinessDataAPI{}
	mockDataDiscovery := &MockDataDiscoveryService{}
	mockMetrics := &MockMetricsCollector{}
	mockErrorMonitor := &MockErrorMonitor{}

	// Configure pipeline components
	collectionConfig := business_intelligence.DataCollectionPipelineConfig{
		EnableParallelCollection: true,
		MaxConcurrentCollections: 4,
		EnableCaching:            true,
		CacheTTL:                 5 * time.Minute,
		EnableDataValidation:     true,
		QualityThreshold:         0.8,
		EnableMetrics:            false, // Disable metrics for benchmarking
		EnableAlerting:           false,
		FallbackEnabled:          true,
		DataSources: []business_intelligence.DataSourceConfig{
			{
				Name:          "benchmark_source",
				Type:          "business_data_api",
				Endpoint:      "https://benchmark-api.example.com",
				Weight:        1.0,
				RateLimit:     1000,
				Timeout:       10 * time.Second,
				RetryAttempts: 3,
			},
		},
	}

	processingConfig := business_intelligence.DataProcessingServiceConfig{
		EnableParallelProcessing: true,
		MaxConcurrentProcessors:  4,
		EnableDataValidation:     true,
		EnableDataEnrichment:     true,
		EnableDataTransformation: true,
		EnableDataCleaning:       true,
		EnableDeduplication:      true,
		EnableMetrics:            false,
		EnableAlerting:           false,
	}

	aggregationConfig := business_intelligence.DataAggregationConfig{
		EnableParallelAggregation: true,
		MaxConcurrentAggregators:  4,
		EnableCorrelation:         true,
		CorrelationThreshold:      0.7,
		EnableAnalysis:            true,
		EnableTrendAnalysis:       true,
		EnablePatternAnalysis:     true,
		EnableAnomalyDetection:    true,
		EnableMetricCalculation:   true,
		EnableQualityAssessment:   true,
		QualityThreshold:          0.8,
		EnableMetrics:             false,
		EnablePerformanceTracking: false,
	}

	// Create pipeline components
	collectionPipeline := business_intelligence.NewDataCollectionPipeline(
		collectionConfig,
		logger,
		mockBusinessDataAPI,
		mockDataDiscovery,
		mockMetrics,
		mockErrorMonitor,
	)

	processingService := business_intelligence.NewDataProcessingService(
		processingConfig,
		logger,
		mockMetrics,
		mockErrorMonitor,
	)

	aggregationService := business_intelligence.NewDataAggregationService(
		aggregationConfig,
		logger,
		mockMetrics,
		mockErrorMonitor,
	)

	// Setup mocks for benchmarking
	mockBusinessDataAPI.On("SearchBusiness", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"business_name": "Benchmark Company",
			"industry":      "Technology",
			"revenue":       1000000,
			"employees":     50,
			"location":      "San Francisco, CA",
		}, nil)

	mockDataDiscovery.On("DiscoverDataPoints", mock.Anything, mock.Anything).Return(
		map[string]interface{}{
			"quality_assessments": []map[string]interface{}{
				{"score": 0.90, "type": "completeness"},
				{"score": 0.85, "type": "accuracy"},
			},
		}, nil)

	mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

	b.ResetTimer()

	// Benchmark entire pipeline
	for i := 0; i < b.N; i++ {
		ctx := context.Background()
		businessID := fmt.Sprintf("benchmark-business-%d", i)
		query := map[string]interface{}{
			"industry": "Technology",
		}

		// Execute entire pipeline
		collectedData, err := collectionPipeline.CollectData(ctx, businessID, query)
		if err != nil {
			b.Fatalf("Data collection failed: %v", err)
		}

		processedData, err := processingService.ProcessCollectedData(ctx, collectedData)
		if err != nil {
			b.Fatalf("Data processing failed: %v", err)
		}

		aggregatedData, err := aggregationService.AggregateProcessedData(ctx, businessID, processedData)
		if err != nil {
			b.Fatalf("Data aggregation failed: %v", err)
		}

		// Verify result
		if aggregatedData == nil {
			b.Fatalf("Aggregated data is nil")
		}
	}
}
