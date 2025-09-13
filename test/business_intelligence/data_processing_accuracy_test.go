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

// MockDataProcessor is a mock implementation of DataProcessor
type MockDataProcessor struct {
	mock.Mock
}

func (m *MockDataProcessor) Process(ctx context.Context, data *business_intelligence.CollectedData) (*business_intelligence.ProcessedData, error) {
	args := m.Called(ctx, data)
	return args.Get(0).(*business_intelligence.ProcessedData), args.Error(1)
}

func (m *MockDataProcessor) GetName() string {
	args := m.Called()
	return args.String(0)
}

// TestDataProcessingService_Accuracy tests the accuracy of data processing operations
func TestDataProcessingService_Accuracy(t *testing.T) {
	logger := zap.NewNop()

	// Create mock dependencies
	mockMetrics := &MockMetricsCollector{}
	mockErrorMonitor := &MockErrorMonitor{}

	// Configure processing service
	config := business_intelligence.DataProcessingServiceConfig{
		EnableParallelProcessing: true,
		MaxConcurrentProcessors:  3,
		EnableDataValidation:     true,
		EnableDataEnrichment:     true,
		EnableDataTransformation: true,
		EnableDataCleaning:       true,
		EnableDeduplication:      true,
		EnableMetrics:            true,
		EnableAlerting:           true,
		ValidationRules: []business_intelligence.ValidationRule{
			{
				Field:    "business_name",
				Operator: "not_empty",
				Value:    "",
				Severity: "error",
			},
			{
				Field:    "industry",
				Operator: "not_empty",
				Value:    "",
				Severity: "warning",
			},
		},
		TransformationRules: []business_intelligence.TransformationRule{
			{
				Field:    "business_name",
				Type:     "uppercase",
				NewField: "business_name_upper",
			},
			{
				Field:    "revenue",
				Type:     "format_currency",
				NewField: "revenue_formatted",
			},
		},
	}

	// Create processing service
	service := business_intelligence.NewDataProcessingService(
		config,
		logger,
		mockMetrics,
		mockErrorMonitor,
	)

	t.Run("Accurate Data Processing", func(t *testing.T) {
		// Setup mocks
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()
		mockErrorMonitor.On("RecordAlert", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create test collected data
		collectedData := []*business_intelligence.CollectedData{
			{
				Source:    "test_source_1",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": "Test Company Inc",
					"industry":      "Technology",
					"revenue":       1000000.50,
					"employees":     50,
					"location":      "San Francisco, CA",
				},
				QualityScore: 0.95,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": "test-business-123",
				},
			},
			{
				Source:    "test_source_2",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": "Another Company LLC",
					"industry":      "Finance",
					"revenue":       2500000.75,
					"employees":     100,
					"location":      "New York, NY",
				},
				QualityScore: 0.88,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": "test-business-456",
				},
			},
		}

		// Test data processing
		ctx := context.Background()
		processedData, err := service.ProcessCollectedData(ctx, collectedData)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, processedData)
		assert.Equal(t, len(collectedData), len(processedData))

		// Verify processing accuracy for each item
		for i, processed := range processedData {
			original := collectedData[i]

			// Verify basic fields are preserved
			assert.Equal(t, original.Source, processed.OriginalSource)
			assert.Equal(t, original.DataType, processed.DataType)
			assert.Equal(t, original.Timestamp, processed.Timestamp)
			assert.Equal(t, original.Metadata, processed.Metadata)

			// Verify data cleaning was applied
			assert.NotNil(t, processed.ProcessedContent)
			assert.Greater(t, len(processed.ProcessedContent), 0)

			// Verify validation was performed
			assert.NotNil(t, processed.ValidationStatus)
			assert.True(t, processed.ValidationStatus.IsValid, "Valid data should pass validation")

			// Verify enrichment was applied
			assert.NotNil(t, processed.EnrichmentStatus)
			assert.True(t, processed.EnrichmentStatus.IsEnriched, "Data should be enriched")

			// Verify transformation was applied
			assert.NotNil(t, processed.TransformationStatus)
			assert.True(t, processed.TransformationStatus.IsTransformed, "Data should be transformed")

			// Verify quality score is maintained or improved
			assert.GreaterOrEqual(t, processed.QualityScore, original.QualityScore*0.8, "Quality score should not degrade significantly")

			// Verify specific transformations
			if businessName, exists := processed.ProcessedContent["business_name"]; exists {
				assert.NotEmpty(t, businessName, "Business name should not be empty after processing")
			}

			// Verify no critical errors
			assert.Empty(t, processed.Errors, "No critical errors should be present in valid data")
		}

		mockMetrics.AssertExpectations(t)
	})

	t.Run("Data Validation Accuracy", func(t *testing.T) {
		// Reset mocks
		mockMetrics.ExpectedCalls = nil
		mockErrorMonitor.ExpectedCalls = nil

		// Setup mocks
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()
		mockErrorMonitor.On("RecordAlert", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create test data with validation issues
		collectedData := []*business_intelligence.CollectedData{
			{
				Source:    "test_source_invalid",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": "", // Empty name should trigger validation error
					"industry":      "Technology",
					"revenue":       -1000, // Negative revenue should trigger validation warning
					"employees":     0,     // Zero employees should trigger validation warning
				},
				QualityScore: 0.95,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": "test-business-invalid",
				},
			},
		}

		// Test data processing with validation issues
		ctx := context.Background()
		processedData, err := service.ProcessCollectedData(ctx, collectedData)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, processedData)
		assert.Equal(t, 1, len(processedData))

		processed := processedData[0]

		// Verify validation detected issues
		assert.NotNil(t, processed.ValidationStatus)
		assert.False(t, processed.ValidationStatus.IsValid, "Invalid data should fail validation")
		assert.NotEmpty(t, processed.ValidationStatus.Errors, "Validation errors should be recorded")

		// Verify quality score is reduced due to validation failures
		assert.Less(t, processed.QualityScore, 0.8, "Quality score should be reduced for invalid data")

		// Verify errors are recorded
		assert.NotEmpty(t, processed.Errors, "Processing errors should be recorded for invalid data")

		mockMetrics.AssertExpectations(t)
		mockErrorMonitor.AssertExpectations(t)
	})

	t.Run("Data Enrichment Accuracy", func(t *testing.T) {
		// Reset mocks
		mockMetrics.ExpectedCalls = nil
		mockErrorMonitor.ExpectedCalls = nil

		// Setup mocks
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create test data for enrichment
		collectedData := []*business_intelligence.CollectedData{
			{
				Source:    "test_source_enrichment",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": "Enrichment Test Company",
					"industry":      "Technology",
					"revenue":       1000000,
					"employees":     50,
				},
				QualityScore: 0.90,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": "test-business-enrichment",
				},
			},
		}

		// Test data processing with enrichment
		ctx := context.Background()
		processedData, err := service.ProcessCollectedData(ctx, collectedData)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, processedData)
		assert.Equal(t, 1, len(processedData))

		processed := processedData[0]

		// Verify enrichment was applied
		assert.NotNil(t, processed.EnrichmentStatus)
		assert.True(t, processed.EnrichmentStatus.IsEnriched, "Data should be enriched")

		// Verify enriched content contains additional fields
		assert.Contains(t, processed.ProcessedContent, "enriched_field", "Enriched data should contain additional fields")

		// Verify original data is preserved
		assert.Contains(t, processed.ProcessedContent, "business_name", "Original data should be preserved")
		assert.Contains(t, processed.ProcessedContent, "industry", "Original data should be preserved")

		mockMetrics.AssertExpectations(t)
	})

	t.Run("Data Transformation Accuracy", func(t *testing.T) {
		// Reset mocks
		mockMetrics.ExpectedCalls = nil

		// Setup mocks
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create test data for transformation
		collectedData := []*business_intelligence.CollectedData{
			{
				Source:    "test_source_transformation",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": "transformation test company",
					"industry":      "Technology",
					"revenue":       1000000.50,
					"employees":     50,
				},
				QualityScore: 0.92,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": "test-business-transformation",
				},
			},
		}

		// Test data processing with transformation
		ctx := context.Background()
		processedData, err := service.ProcessCollectedData(ctx, collectedData)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, processedData)
		assert.Equal(t, 1, len(processedData))

		processed := processedData[0]

		// Verify transformation was applied
		assert.NotNil(t, processed.TransformationStatus)
		assert.True(t, processed.TransformationStatus.IsTransformed, "Data should be transformed")

		// Verify specific transformations
		if revenueFormatted, exists := processed.ProcessedContent["revenue_formatted"]; exists {
			assert.NotEmpty(t, revenueFormatted, "Revenue should be formatted")
			assert.IsType(t, "", revenueFormatted, "Formatted revenue should be a string")
		}

		// Verify original data is preserved alongside transformed data
		assert.Contains(t, processed.ProcessedContent, "revenue", "Original revenue should be preserved")
		assert.Contains(t, processed.ProcessedContent, "business_name", "Original business name should be preserved")

		mockMetrics.AssertExpectations(t)
	})

	t.Run("Data Cleaning Accuracy", func(t *testing.T) {
		// Reset mocks
		mockMetrics.ExpectedCalls = nil

		// Setup mocks
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create test data with cleaning issues
		collectedData := []*business_intelligence.CollectedData{
			{
				Source:    "test_source_cleaning",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": "  Clean Test Company  ", // Extra whitespace
					"industry":      "Technology",
					"revenue":       nil, // Null value
					"employees":     50,
					"empty_field":   "", // Empty string
				},
				QualityScore: 0.85,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": "test-business-cleaning",
				},
			},
		}

		// Test data processing with cleaning
		ctx := context.Background()
		processedData, err := service.ProcessCollectedData(ctx, collectedData)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, processedData)
		assert.Equal(t, 1, len(processedData))

		processed := processedData[0]

		// Verify data cleaning was applied
		assert.NotNil(t, processed.ProcessedContent)

		// Verify whitespace was trimmed
		if businessName, exists := processed.ProcessedContent["business_name"]; exists {
			assert.Equal(t, "Clean Test Company", businessName, "Whitespace should be trimmed")
		}

		// Verify null values were handled
		assert.NotContains(t, processed.ProcessedContent, "revenue", "Null values should be removed")

		// Verify empty strings were handled
		assert.NotContains(t, processed.ProcessedContent, "empty_field", "Empty strings should be removed")

		// Verify valid data is preserved
		assert.Contains(t, processed.ProcessedContent, "industry", "Valid data should be preserved")
		assert.Contains(t, processed.ProcessedContent, "employees", "Valid data should be preserved")

		mockMetrics.AssertExpectations(t)
	})

	t.Run("Parallel Processing Accuracy", func(t *testing.T) {
		// Reset mocks
		mockMetrics.ExpectedCalls = nil

		// Setup mocks
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create multiple test data items for parallel processing
		collectedData := make([]*business_intelligence.CollectedData, 10)
		for i := 0; i < 10; i++ {
			collectedData[i] = &business_intelligence.CollectedData{
				Source:    "test_source_parallel",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": fmt.Sprintf("Parallel Test Company %d", i),
					"industry":      "Technology",
					"revenue":       1000000 + float64(i*100000),
					"employees":     50 + i*10,
				},
				QualityScore: 0.90,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": fmt.Sprintf("test-business-parallel-%d", i),
				},
			}
		}

		// Test parallel processing
		ctx := context.Background()
		startTime := time.Now()
		processedData, err := service.ProcessCollectedData(ctx, collectedData)
		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, processedData)
		assert.Equal(t, len(collectedData), len(processedData))

		// Verify all items were processed correctly
		for i, processed := range processedData {
			original := collectedData[i]

			assert.Equal(t, original.Source, processed.OriginalSource)
			assert.Equal(t, original.DataType, processed.DataType)
			assert.True(t, processed.ValidationStatus.IsValid)
			assert.True(t, processed.EnrichmentStatus.IsEnriched)
			assert.True(t, processed.TransformationStatus.IsTransformed)
		}

		// Verify parallel processing was faster than sequential would be
		assert.Less(t, duration, 2*time.Second, "Parallel processing should be faster than sequential")

		mockMetrics.AssertExpectations(t)
	})
}

// TestDataProcessingService_ErrorHandling tests error handling in data processing
func TestDataProcessingService_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()

	// Create mock dependencies
	mockMetrics := &MockMetricsCollector{}
	mockErrorMonitor := &MockErrorMonitor{}

	// Configure processing service
	config := business_intelligence.DataProcessingServiceConfig{
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

	// Create processing service
	service := business_intelligence.NewDataProcessingService(
		config,
		logger,
		mockMetrics,
		mockErrorMonitor,
	)

	t.Run("Processing Error Recovery", func(t *testing.T) {
		// Setup mocks
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()
		mockErrorMonitor.On("RecordError", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create test data with processing issues
		collectedData := []*business_intelligence.CollectedData{
			{
				Source:    "test_source_error",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": "Error Test Company",
					"industry":      "Technology",
					"revenue":       "invalid_revenue", // Invalid type
					"employees":     -50,               // Invalid value
				},
				QualityScore: 0.70,
				IsValid:      false,
				Errors:       []string{"invalid_revenue_format", "negative_employees"},
				Metadata: map[string]interface{}{
					"business_id": "test-business-error",
				},
			},
		}

		// Test data processing with errors
		ctx := context.Background()
		processedData, err := service.ProcessCollectedData(ctx, collectedData)

		// Assertions
		require.NoError(t, err, "Processing should not fail even with data errors")
		assert.NotNil(t, processedData)
		assert.Equal(t, 1, len(processedData))

		processed := processedData[0]

		// Verify error handling
		assert.NotEmpty(t, processed.Errors, "Processing errors should be recorded")
		assert.Less(t, processed.QualityScore, 0.8, "Quality score should be reduced for error data")

		// Verify processing continued despite errors
		assert.NotNil(t, processed.ProcessedContent)
		assert.NotNil(t, processed.ValidationStatus)
		assert.NotNil(t, processed.EnrichmentStatus)
		assert.NotNil(t, processed.TransformationStatus)

		mockMetrics.AssertExpectations(t)
		mockErrorMonitor.AssertExpectations(t)
	})

	t.Run("Context Cancellation", func(t *testing.T) {
		// Setup mocks
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Create test data
		collectedData := []*business_intelligence.CollectedData{
			{
				Source:    "test_source_cancellation",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": "Cancellation Test Company",
					"industry":      "Technology",
					"revenue":       1000000,
					"employees":     50,
				},
				QualityScore: 0.90,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": "test-business-cancellation",
				},
			},
		}

		// Test with context cancellation
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		// Add delay to ensure cancellation
		time.Sleep(20 * time.Millisecond)

		processedData, err := service.ProcessCollectedData(ctx, collectedData)

		// Assertions
		assert.Error(t, err, "Should return error due to context cancellation")
		assert.Contains(t, err.Error(), "context deadline exceeded")

		mockMetrics.AssertExpectations(t)
	})
}

// TestDataProcessingService_Configuration tests different configuration scenarios
func TestDataProcessingService_Configuration(t *testing.T) {
	logger := zap.NewNop()

	t.Run("Sequential Processing Mode", func(t *testing.T) {
		// Create mock dependencies
		mockMetrics := &MockMetricsCollector{}
		mockErrorMonitor := &MockErrorMonitor{}

		// Configure for sequential processing
		config := business_intelligence.DataProcessingServiceConfig{
			EnableParallelProcessing: false, // Sequential mode
			MaxConcurrentProcessors:  1,
			EnableDataValidation:     true,
			EnableDataEnrichment:     true,
			EnableDataTransformation: true,
			EnableDataCleaning:       true,
			EnableDeduplication:      true,
			EnableMetrics:            true,
			EnableAlerting:           false,
		}

		// Create processing service
		service := business_intelligence.NewDataProcessingService(
			config,
			logger,
			mockMetrics,
			mockErrorMonitor,
		)

		// Create test data
		collectedData := make([]*business_intelligence.CollectedData, 5)
		for i := 0; i < 5; i++ {
			collectedData[i] = &business_intelligence.CollectedData{
				Source:    "test_source_sequential",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": fmt.Sprintf("Sequential Test Company %d", i),
					"industry":      "Technology",
					"revenue":       1000000 + float64(i*100000),
					"employees":     50 + i*10,
				},
				QualityScore: 0.90,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": fmt.Sprintf("test-business-sequential-%d", i),
				},
			}
		}

		// Setup mocks
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test sequential processing
		ctx := context.Background()
		startTime := time.Now()
		processedData, err := service.ProcessCollectedData(ctx, collectedData)
		duration := time.Since(startTime)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, processedData)
		assert.Equal(t, len(collectedData), len(processedData))

		// Sequential processing should take longer than parallel
		assert.Greater(t, duration, 100*time.Millisecond, "Sequential processing should take more time")

		mockMetrics.AssertExpectations(t)
	})

	t.Run("Disabled Processing Steps", func(t *testing.T) {
		// Create mock dependencies
		mockMetrics := &MockMetricsCollector{}
		mockErrorMonitor := &MockErrorMonitor{}

		// Configure with some processing steps disabled
		config := business_intelligence.DataProcessingServiceConfig{
			EnableParallelProcessing: true,
			MaxConcurrentProcessors:  2,
			EnableDataValidation:     false, // Disable validation
			EnableDataEnrichment:     false, // Disable enrichment
			EnableDataTransformation: true,  // Keep transformation
			EnableDataCleaning:       true,  // Keep cleaning
			EnableDeduplication:      false, // Disable deduplication
			EnableMetrics:            true,
			EnableAlerting:           false,
		}

		// Create processing service
		service := business_intelligence.NewDataProcessingService(
			config,
			logger,
			mockMetrics,
			mockErrorMonitor,
		)

		// Create test data
		collectedData := []*business_intelligence.CollectedData{
			{
				Source:    "test_source_disabled",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": "Disabled Steps Test Company",
					"industry":      "Technology",
					"revenue":       1000000,
					"employees":     50,
				},
				QualityScore: 0.90,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": "test-business-disabled",
				},
			},
		}

		// Setup mocks
		mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

		// Test processing with disabled steps
		ctx := context.Background()
		processedData, err := service.ProcessCollectedData(ctx, collectedData)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, processedData)
		assert.Equal(t, 1, len(processedData))

		processed := processedData[0]

		// Verify disabled steps were skipped
		assert.False(t, processed.ValidationStatus.IsValid, "Validation should be skipped when disabled")
		assert.False(t, processed.EnrichmentStatus.IsEnriched, "Enrichment should be skipped when disabled")
		assert.True(t, processed.TransformationStatus.IsTransformed, "Transformation should still be applied")

		// Verify quality score is maintained when validation is disabled
		assert.Equal(t, 0.90, processed.QualityScore, "Quality score should be maintained when validation is disabled")

		mockMetrics.AssertExpectations(t)
	})
}

// BenchmarkDataProcessingService benchmarks the data processing service performance
func BenchmarkDataProcessingService(b *testing.B) {
	logger := zap.NewNop()

	// Create mock dependencies
	mockMetrics := &MockMetricsCollector{}
	mockErrorMonitor := &MockErrorMonitor{}

	// Configure processing service
	config := business_intelligence.DataProcessingServiceConfig{
		EnableParallelProcessing: true,
		MaxConcurrentProcessors:  4,
		EnableDataValidation:     true,
		EnableDataEnrichment:     true,
		EnableDataTransformation: true,
		EnableDataCleaning:       true,
		EnableDeduplication:      true,
		EnableMetrics:            true,
		EnableAlerting:           false,
	}

	// Create processing service
	service := business_intelligence.NewDataProcessingService(
		config,
		logger,
		mockMetrics,
		mockErrorMonitor,
	)

	// Setup mocks
	mockMetrics.On("RecordMetric", mock.Anything, mock.Anything, mock.Anything).Return()

	b.ResetTimer()

	// Benchmark data processing
	for i := 0; i < b.N; i++ {
		// Create test data
		collectedData := []*business_intelligence.CollectedData{
			{
				Source:    "benchmark_source",
				DataType:  "business_data",
				Timestamp: time.Now(),
				RawData: map[string]interface{}{
					"business_name": fmt.Sprintf("Benchmark Company %d", i),
					"industry":      "Technology",
					"revenue":       1000000 + float64(i*100000),
					"employees":     50 + i*10,
					"location":      "San Francisco, CA",
				},
				QualityScore: 0.90,
				IsValid:      true,
				Errors:       []string{},
				Metadata: map[string]interface{}{
					"business_id": fmt.Sprintf("benchmark-business-%d", i),
				},
			},
		}

		ctx := context.Background()
		_, err := service.ProcessCollectedData(ctx, collectedData)
		if err != nil {
			b.Fatalf("Data processing failed: %v", err)
		}
	}
}
