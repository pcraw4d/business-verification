package classification_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewAccuracyTracker(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	assert.NotNil(t, tracker)
	assert.NotNil(t, tracker.config)
	assert.Equal(t, 0.90, tracker.config.TargetAccuracy)
	assert.Equal(t, 0.85, tracker.config.CriticalAccuracyThreshold)
	assert.True(t, tracker.config.EnableRealTimeTracking)
	assert.True(t, tracker.config.EnableMisclassificationLog)
}

func TestAccuracyTracker_TrackClassification(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	tests := []struct {
		name             string
		result           *ClassificationResult
		expectedAccuracy float64
		expectAlert      bool
	}{
		{
			name: "correct classification",
			result: &ClassificationResult{
				ID:                     "test-1",
				BusinessName:           "Test Company",
				ActualClassification:   "technology",
				ExpectedClassification: stringPtr("technology"),
				ConfidenceScore:        0.85,
				ClassificationMethod:   "ml",
				Timestamp:              time.Now(),
				IsCorrect:              boolPtr(true),
			},
			expectedAccuracy: 1.0,
			expectAlert:      false,
		},
		{
			name: "incorrect classification",
			result: &ClassificationResult{
				ID:                     "test-2",
				BusinessName:           "Test Company 2",
				ActualClassification:   "finance",
				ExpectedClassification: stringPtr("technology"),
				ConfidenceScore:        0.75,
				ClassificationMethod:   "ml",
				Timestamp:              time.Now(),
				IsCorrect:              boolPtr(false),
			},
			expectedAccuracy: 0.5,
			expectAlert:      false,
		},
		{
			name: "high confidence error",
			result: &ClassificationResult{
				ID:                     "test-3",
				BusinessName:           "Test Company 3",
				ActualClassification:   "retail",
				ExpectedClassification: stringPtr("technology"),
				ConfidenceScore:        0.95,
				ClassificationMethod:   "ml",
				Timestamp:              time.Now(),
				IsCorrect:              boolPtr(false),
			},
			expectedAccuracy: 0.33,
			expectAlert:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tracker.TrackClassification(context.Background(), tt.result)
			assert.NoError(t, err)

			// Check overall accuracy
			accuracy := tracker.GetOverallAccuracy()
			assert.InDelta(t, tt.expectedAccuracy, accuracy, 0.01)
		})
	}

	// Check final state
	metrics := tracker.GetAccuracyMetrics()
	assert.Contains(t, metrics, "overall")

	overallMetrics := metrics["overall"]
	assert.Equal(t, 3, overallMetrics.TotalClassifications)
	assert.Equal(t, 1, overallMetrics.CorrectClassifications)
	assert.InDelta(t, 0.33, overallMetrics.AccuracyScore, 0.01)
}

func TestAccuracyTracker_DimensionalAnalysis(t *testing.T) {
	config := &AccuracyConfig{
		EnableDimensionalAnalysis: true,
		SampleWindowSize:          10,
	}
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(config, logger)

	// Add classifications with different methods
	results := []*ClassificationResult{
		{
			ID:                     "ml-1",
			ActualClassification:   "technology",
			ExpectedClassification: stringPtr("technology"),
			ConfidenceScore:        0.9,
			ClassificationMethod:   "ml",
			IsCorrect:              boolPtr(true),
			Timestamp:              time.Now(),
		},
		{
			ID:                     "ml-2",
			ActualClassification:   "finance",
			ExpectedClassification: stringPtr("technology"),
			ConfidenceScore:        0.8,
			ClassificationMethod:   "ml",
			IsCorrect:              boolPtr(false),
			Timestamp:              time.Now(),
		},
		{
			ID:                     "keyword-1",
			ActualClassification:   "retail",
			ExpectedClassification: stringPtr("retail"),
			ConfidenceScore:        0.7,
			ClassificationMethod:   "keyword",
			IsCorrect:              boolPtr(true),
			Timestamp:              time.Now(),
		},
	}

	for _, result := range results {
		err := tracker.TrackClassification(context.Background(), result)
		assert.NoError(t, err)
	}

	metrics := tracker.GetAccuracyMetrics()

	// Check method-specific metrics
	assert.Contains(t, metrics, "method:ml")
	assert.Contains(t, metrics, "method:keyword")

	mlMetrics := metrics["method:ml"]
	assert.Equal(t, 2, mlMetrics.TotalClassifications)
	assert.Equal(t, 1, mlMetrics.CorrectClassifications)
	assert.InDelta(t, 0.5, mlMetrics.AccuracyScore, 0.01)

	keywordMetrics := metrics["method:keyword"]
	assert.Equal(t, 1, keywordMetrics.TotalClassifications)
	assert.Equal(t, 1, keywordMetrics.CorrectClassifications)
	assert.InDelta(t, 1.0, keywordMetrics.AccuracyScore, 0.01)

	// Check confidence range metrics
	assert.Contains(t, metrics, "confidence_range:high")
	assert.Contains(t, metrics, "confidence_range:medium")
}

func TestAccuracyTracker_MisclassificationLogging(t *testing.T) {
	config := &AccuracyConfig{
		EnableMisclassificationLog: true,
	}
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(config, logger)

	// Add a misclassification
	result := &ClassificationResult{
		ID:                     "error-1",
		BusinessName:           "Error Company",
		ActualClassification:   "finance",
		ExpectedClassification: stringPtr("technology"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "ml",
		IsCorrect:              boolPtr(false),
		Timestamp:              time.Now(),
		Metadata:               map[string]interface{}{"source": "test"},
	}

	err := tracker.TrackClassification(context.Background(), result)
	assert.NoError(t, err)

	// Check misclassifications
	misclassifications := tracker.GetMisclassifications(10)
	assert.Len(t, misclassifications, 1)

	mc := misclassifications[0]
	assert.Equal(t, "Error Company", mc.BusinessName)
	assert.Equal(t, "finance", mc.ActualClassification)
	assert.Equal(t, "technology", mc.ExpectedClassification)
	assert.Equal(t, 0.95, mc.ConfidenceScore)
	assert.Equal(t, "high_confidence_error", mc.ErrorType)
	assert.True(t, mc.ActionRequired)
}

func TestAccuracyTracker_AlertGeneration(t *testing.T) {
	config := &AccuracyConfig{
		CriticalAccuracyThreshold: 0.8,
		TargetAccuracy:            0.9,
	}
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(config, logger)

	// Add multiple incorrect classifications to trigger alerts
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("error-%d", i),
			BusinessName:           fmt.Sprintf("Company %d", i),
			ActualClassification:   "finance",
			ExpectedClassification: stringPtr("technology"),
			ConfidenceScore:        0.7,
			ClassificationMethod:   "ml",
			IsCorrect:              boolPtr(false),
			Timestamp:              time.Now(),
		}

		err := tracker.TrackClassification(context.Background(), result)
		assert.NoError(t, err)
	}

	// Add one correct classification
	result := &ClassificationResult{
		ID:                     "correct-1",
		BusinessName:           "Good Company",
		ActualClassification:   "technology",
		ExpectedClassification: stringPtr("technology"),
		ConfidenceScore:        0.9,
		ClassificationMethod:   "ml",
		IsCorrect:              boolPtr(true),
		Timestamp:              time.Now(),
	}

	err := tracker.TrackClassification(context.Background(), result)
	assert.NoError(t, err)

	// Check that alerts were generated
	alerts := tracker.GetActiveAlerts()
	assert.Greater(t, len(alerts), 0)

	// Check accuracy is below threshold
	accuracy := tracker.GetOverallAccuracy()
	assert.Less(t, accuracy, 0.8)
}

func TestAccuracyTracker_TrendCalculation(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	// Test trend calculation
	values := []float64{0.8, 0.85, 0.9, 0.95, 1.0}
	trend := tracker.calculateTrend(values)
	assert.Greater(t, trend, 0) // Should be positive (improving)

	// Test declining trend
	decliningValues := []float64{1.0, 0.95, 0.9, 0.85, 0.8}
	decliningTrend := tracker.calculateTrend(decliningValues)
	assert.Less(t, decliningTrend, 0) // Should be negative (declining)

	// Test stable trend
	stableValues := []float64{0.9, 0.9, 0.9, 0.9, 0.9}
	stableTrend := tracker.calculateTrend(stableValues)
	assert.InDelta(t, 0, stableTrend, 0.01) // Should be near zero (stable)
}

func TestAccuracyTracker_ConfidenceRangeClassification(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	tests := []struct {
		confidence    float64
		expectedRange string
	}{
		{0.95, "high"},
		{0.85, "medium"},
		{0.65, "medium"},
		{0.45, "low"},
		{0.25, "very_low"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("confidence_%.2f", tt.confidence), func(t *testing.T) {
			result := tracker.getConfidenceRange(tt.confidence)
			assert.Equal(t, tt.expectedRange, result)
		})
	}
}

func TestAccuracyTracker_TimeOfDayClassification(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	tests := []struct {
		hour           int
		expectedPeriod string
	}{
		{8, "morning"},
		{14, "afternoon"},
		{20, "evening"},
		{2, "night"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("hour_%d", tt.hour), func(t *testing.T) {
			timestamp := time.Date(2023, 1, 1, tt.hour, 0, 0, 0, time.UTC)
			result := tracker.getTimeOfDayCategory(timestamp)
			assert.Equal(t, tt.expectedPeriod, result)
		})
	}
}

func TestAccuracyTracker_AlertResolution(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	// Create an alert manually
	alert := &AccuracyAlert{
		ID:             "test-alert-1",
		Type:           "threshold",
		Severity:       "high",
		DimensionName:  "overall",
		DimensionValue: "all",
		CurrentValue:   0.75,
		ThresholdValue: 0.85,
		Message:        "Test alert",
		Timestamp:      time.Now(),
		Resolved:       false,
	}

	tracker.alerts = append(tracker.alerts, alert)

	// Resolve the alert
	err := tracker.ResolveAlert("test-alert-1")
	assert.NoError(t, err)
	assert.True(t, alert.Resolved)
	assert.NotNil(t, alert.ResolvedAt)

	// Try to resolve non-existent alert
	err = tracker.ResolveAlert("non-existent")
	assert.Error(t, err)
}

func TestAccuracyTracker_MetricCollectors(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	// Create a mock collector
	collector := &MockMetricCollector{
		dimensionName: "test_dimension",
		metrics: []*AccuracyMetrics{
			{
				DimensionName:  "test_dimension",
				DimensionValue: "test_value",
				AccuracyScore:  0.95,
			},
		},
	}

	tracker.AddMetricCollector(collector)

	// Collect metrics
	err := tracker.CollectMetrics(context.Background())
	assert.NoError(t, err)

	// Check that metrics were collected
	metrics := tracker.GetAccuracyMetrics()
	assert.Contains(t, metrics, "test_dimension:test_value")
}

func TestAccuracyTracker_HistoricalSnapshots(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	// Add some classifications
	result := &ClassificationResult{
		ID:                     "hist-1",
		ActualClassification:   "technology",
		ExpectedClassification: stringPtr("technology"),
		ConfidenceScore:        0.9,
		ClassificationMethod:   "ml",
		IsCorrect:              boolPtr(true),
		Timestamp:              time.Now(),
	}

	err := tracker.TrackClassification(context.Background(), result)
	assert.NoError(t, err)

	// Create snapshot
	snapshot := tracker.CreateSnapshot()
	assert.NotNil(t, snapshot)
	assert.Equal(t, 1.0, snapshot.OverallAccuracy)
	assert.Equal(t, 1, snapshot.SampleSize)
	assert.Contains(t, snapshot.Metadata, "total_dimensions")

	// Get historical data
	since := time.Now().Add(-1 * time.Hour)
	historical := tracker.GetHistoricalData(since)
	assert.Len(t, historical, 1)
	assert.Equal(t, snapshot.Timestamp, historical[0].Timestamp)
}

// Helper types and functions for testing

type MockMetricCollector struct {
	dimensionName string
	metrics       []*AccuracyMetrics
}

func (m *MockMetricCollector) CollectMetrics(ctx context.Context) ([]*AccuracyMetrics, error) {
	return m.metrics, nil
}

func (m *MockMetricCollector) GetDimensions() []string {
	return []string{m.dimensionName}
}

func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func TestAccuracyTracker_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	// Test concurrent access
	const numGoroutines = 10
	const numOperations = 100

	done := make(chan bool, numGoroutines)

	// Start multiple goroutines performing operations
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for j := 0; j < numOperations; j++ {
				result := &ClassificationResult{
					ID:                     fmt.Sprintf("concurrent-%d-%d", goroutineID, j),
					BusinessName:           fmt.Sprintf("Company %d-%d", goroutineID, j),
					ActualClassification:   "technology",
					ExpectedClassification: stringPtr("technology"),
					ConfidenceScore:        0.8,
					ClassificationMethod:   "ml",
					IsCorrect:              boolPtr(true),
					Timestamp:              time.Now(),
				}

				err := tracker.TrackClassification(context.Background(), result)
				assert.NoError(t, err)

				// Also test concurrent reads
				_ = tracker.GetOverallAccuracy()
				_ = tracker.GetAccuracyMetrics()
				_ = tracker.GetActiveAlerts()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final state
	metrics := tracker.GetAccuracyMetrics()
	overallMetrics := metrics["overall"]
	expectedTotal := numGoroutines * numOperations
	assert.Equal(t, expectedTotal, overallMetrics.TotalClassifications)
	assert.Equal(t, expectedTotal, overallMetrics.CorrectClassifications)
	assert.Equal(t, 1.0, overallMetrics.AccuracyScore)
}

func TestAccuracyTracker_ErrorTypeClassification(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	tests := []struct {
		name         string
		confidence   float64
		expectedType string
	}{
		{"very low confidence", 0.2, "low_confidence"},
		{"low confidence", 0.4, "low_confidence"},
		{"medium confidence", 0.6, "medium_confidence_error"},
		{"high confidence error", 0.85, "high_confidence_error"},
		{"very high confidence error", 0.95, "high_confidence_error"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ClassificationResult{
				ConfidenceScore: tt.confidence,
			}

			errorType := tracker.classifyErrorType(result)
			assert.Equal(t, tt.expectedType, errorType)
		})
	}
}

func TestAccuracyTracker_SeverityCalculation(t *testing.T) {
	logger := zap.NewNop()
	tracker := NewAccuracyTracker(nil, logger)

	tests := []struct {
		name             string
		confidence       float64
		expectedIndustry string
		actualIndustry   string
		expectedSeverity string
	}{
		{
			name:             "high confidence financial error",
			confidence:       0.95,
			expectedIndustry: "financial_services",
			actualIndustry:   "technology",
			expectedSeverity: "critical",
		},
		{
			name:             "medium confidence error",
			confidence:       0.7,
			expectedIndustry: "technology",
			actualIndustry:   "retail",
			expectedSeverity: "medium",
		},
		{
			name:             "low confidence error",
			confidence:       0.4,
			expectedIndustry: "technology",
			actualIndustry:   "retail",
			expectedSeverity: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &ClassificationResult{
				ConfidenceScore:        tt.confidence,
				ExpectedClassification: &tt.expectedIndustry,
				ActualClassification:   tt.actualIndustry,
			}

			severity := tracker.calculateMisclassificationSeverity(result)
			assert.Equal(t, tt.expectedSeverity, severity)
		})
	}
}
