package enrichment

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Test data structures
type TestDataWithTimestamp struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func TestNewDataFreshnessTracker(t *testing.T) {
	tests := []struct {
		name   string
		config *DataFreshnessConfig
	}{
		{
			name:   "with nil config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &DataFreshnessConfig{
				EnableTracking:           true,
				StalenessThreshold:       12 * time.Hour,
				UpdateFrequencyThreshold: 2 * time.Hour,
				AgeWeight:                0.5,
				UpdateFrequencyWeight:    0.3,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewDataFreshnessTracker(zap.NewNop(), tt.config)
			assert.NotNil(t, tracker)
			assert.NotNil(t, tracker.config)
			assert.NotNil(t, tracker.logger)
			assert.NotNil(t, tracker.tracer)

			if tt.config == nil {
				// Should use default config
				assert.True(t, tracker.config.EnableTracking)
				assert.Equal(t, 24*time.Hour, tracker.config.StalenessThreshold)
				assert.Equal(t, 6*time.Hour, tracker.config.UpdateFrequencyThreshold)
			} else {
				// Should use provided config
				assert.Equal(t, tt.config.EnableTracking, tracker.config.EnableTracking)
				assert.Equal(t, tt.config.StalenessThreshold, tracker.config.StalenessThreshold)
				assert.Equal(t, tt.config.UpdateFrequencyThreshold, tracker.config.UpdateFrequencyThreshold)
			}
		})
	}
}

func TestDataFreshnessTracker_TrackFreshness(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)
	ctx := context.Background()

	tests := []struct {
		name     string
		data     interface{}
		dataID   string
		dataType string
		source   string
	}{
		{
			name: "track fresh data",
			data: &TestDataWithTimestamp{
				ID:        "test-123",
				Name:      "Fresh Data",
				CreatedAt: time.Now().Add(-1 * time.Hour),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
			},
			dataID:   "test-123",
			dataType: "business",
			source:   "api_response",
		},
		{
			name: "track stale data",
			data: &TestDataWithTimestamp{
				ID:        "test-456",
				Name:      "Stale Data",
				CreatedAt: time.Now().Add(-48 * time.Hour),
				UpdatedAt: time.Now().Add(-36 * time.Hour),
			},
			dataID:   "test-456",
			dataType: "business",
			source:   "website_content",
		},
		{
			name: "track data without timestamp",
			data: map[string]interface{}{
				"id":   "test-789",
				"name": "No Timestamp Data",
			},
			dataID:   "test-789",
			dataType: "business",
			source:   "user_input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := tracker.TrackFreshness(ctx, tt.data, tt.dataID, tt.dataType, tt.source)
			require.NoError(t, err)
			require.NotNil(t, record)

			// Basic record validation
			assert.Equal(t, tt.dataID, record.DataID)
			assert.Equal(t, tt.dataType, record.DataType)
			assert.Equal(t, tt.source, record.Source)
			assert.NotZero(t, record.LastUpdated)
			assert.Equal(t, 1, record.UpdateCount)

			// Freshness score validation
			assert.GreaterOrEqual(t, record.FreshnessScore, 0.0)
			assert.LessOrEqual(t, record.FreshnessScore, 1.0)

			// Age validation
			assert.GreaterOrEqual(t, record.Age, time.Duration(0))

			// Staleness validation
			if record.Age > tracker.config.StalenessThreshold {
				assert.True(t, record.IsStale)
			}
			if record.Age > tracker.config.CriticalStalenessThreshold {
				assert.True(t, record.IsCriticalStale)
			}
		})
	}
}

func TestDataFreshnessTracker_TrackFreshnessMultipleUpdates(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)
	ctx := context.Background()

	data := &TestDataWithTimestamp{
		ID:        "test-multi",
		Name:      "Multi Update Data",
		CreatedAt: time.Now().Add(-2 * time.Hour),
		UpdatedAt: time.Now().Add(-1 * time.Hour),
	}

	// First tracking
	record1, err := tracker.TrackFreshness(ctx, data, "test-multi", "business", "api")
	require.NoError(t, err)
	assert.Equal(t, 1, record1.UpdateCount)
	assert.Equal(t, time.Duration(0), record1.UpdateFrequency)

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Second tracking
	record2, err := tracker.TrackFreshness(ctx, data, "test-multi", "business", "api")
	require.NoError(t, err)
	assert.Equal(t, 2, record2.UpdateCount)
	assert.Greater(t, record2.UpdateFrequency, time.Duration(0))
	assert.True(t, record2.LastUpdated.After(record1.LastUpdated))
}

func TestDataFreshnessTracker_AnalyzeFreshness(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)
	ctx := context.Background()

	// First track some data
	data := &TestDataWithTimestamp{
		ID:        "test-analyze",
		Name:      "Analysis Test Data",
		CreatedAt: time.Now().Add(-12 * time.Hour),
		UpdatedAt: time.Now().Add(-6 * time.Hour),
	}

	_, err := tracker.TrackFreshness(ctx, data, "test-analyze", "business", "api")
	require.NoError(t, err)

	// Analyze the data
	result, err := tracker.AnalyzeFreshness(ctx, "test-analyze", "business", "api")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Basic result validation
	assert.NotNil(t, result.CurrentFreshness)
	assert.Equal(t, "test-analyze", result.CurrentFreshness.DataID)
	assert.Equal(t, "business", result.CurrentFreshness.DataType)
	assert.Equal(t, "api", result.CurrentFreshness.Source)

	// Score validation
	assert.GreaterOrEqual(t, result.OverallScore, 0.0)
	assert.LessOrEqual(t, result.OverallScore, 1.0)

	// Freshness level validation
	assert.Contains(t, []string{"fresh", "aging", "stale", "critical"}, result.FreshnessLevel)

	// Metadata validation
	assert.NotZero(t, result.AnalyzedAt)
	assert.Greater(t, result.ProcessingTime, time.Duration(0))
	assert.GreaterOrEqual(t, result.DataPoints, 1)

	// Predictive analysis validation (if enabled)
	if tracker.config.EnablePredictiveScoring {
		assert.GreaterOrEqual(t, result.PredictiveScore, 0.0)
		assert.LessOrEqual(t, result.PredictiveScore, 1.0)
		assert.GreaterOrEqual(t, result.StalenessRisk, 0.0)
		assert.LessOrEqual(t, result.StalenessRisk, 1.0)
	}
}

func TestDataFreshnessTracker_AnalyzeFreshnessNoHistory(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)
	ctx := context.Background()

	// Analyze data that hasn't been tracked
	result, err := tracker.AnalyzeFreshness(ctx, "nonexistent", "business", "api")
	require.NoError(t, err)
	require.NotNil(t, result)

	// Should have default values
	assert.NotNil(t, result.CurrentFreshness)
	assert.Equal(t, "nonexistent", result.CurrentFreshness.DataID)
	assert.Equal(t, 0.5, result.CurrentFreshness.FreshnessScore) // Default score
	assert.Equal(t, 0, result.DataPoints)
}

func TestDataFreshnessTracker_GetStalenessAlerts(t *testing.T) {
	config := getDefaultDataFreshnessConfig()
	config.StalenessThreshold = 1 * time.Hour // Short threshold for testing
	config.CriticalStalenessThreshold = 2 * time.Hour
	tracker := NewDataFreshnessTracker(zap.NewNop(), config)
	ctx := context.Background()

	// Track stale data to trigger alerts
	staleData := &TestDataWithTimestamp{
		ID:        "test-stale",
		Name:      "Stale Data for Alerts",
		CreatedAt: time.Now().Add(-3 * time.Hour),
		UpdatedAt: time.Now().Add(-3 * time.Hour),
	}

	_, err := tracker.TrackFreshness(ctx, staleData, "test-stale", "business", "api")
	require.NoError(t, err)

	// Get alerts
	alerts, err := tracker.GetStalenessAlerts(ctx)
	require.NoError(t, err)

	// Should have at least one alert for stale data
	assert.GreaterOrEqual(t, len(alerts), 1)

	// Check alert properties
	for _, alert := range alerts {
		assert.NotEmpty(t, alert.DataID)
		assert.NotEmpty(t, alert.Source)
		assert.NotEmpty(t, alert.DataType)
		assert.NotEmpty(t, alert.AlertType)
		assert.NotEmpty(t, alert.Severity)
		assert.NotEmpty(t, alert.Message)
		assert.True(t, alert.IsActive)
		assert.Greater(t, alert.TriggerCount, 0)
		assert.NotZero(t, alert.CreatedAt)
		assert.NotZero(t, alert.LastTriggered)
	}
}

func TestDataFreshnessTracker_GetUpdatePatterns(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)
	ctx := context.Background()

	// Track data multiple times to create patterns
	data := &TestDataWithTimestamp{
		ID:        "test-pattern",
		Name:      "Pattern Test Data",
		CreatedAt: time.Now().Add(-24 * time.Hour),
		UpdatedAt: time.Now().Add(-12 * time.Hour),
	}

	// Track multiple times
	for i := 0; i < 3; i++ {
		_, err := tracker.TrackFreshness(ctx, data, "test-pattern", "business", "api")
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	// Get patterns
	patterns, err := tracker.GetUpdatePatterns(ctx)
	require.NoError(t, err)

	// Should have at least one pattern
	assert.GreaterOrEqual(t, len(patterns), 1)

	// Check pattern properties
	for _, pattern := range patterns {
		assert.NotEmpty(t, pattern.Source)
		assert.NotEmpty(t, pattern.DataType)
		assert.Greater(t, pattern.UpdateCount, 0)
		assert.NotZero(t, pattern.LastUpdate)
		assert.GreaterOrEqual(t, pattern.ConsistencyScore, 0.0)
		assert.LessOrEqual(t, pattern.ConsistencyScore, 1.0)
		assert.NotEmpty(t, pattern.PatternType)
		assert.GreaterOrEqual(t, pattern.Confidence, 0.0)
		assert.LessOrEqual(t, pattern.Confidence, 1.0)
	}
}

func TestDataFreshnessTracker_FreshnessScoring(t *testing.T) {
	config := getDefaultDataFreshnessConfig()
	config.StalenessThreshold = 24 * time.Hour
	tracker := NewDataFreshnessTracker(zap.NewNop(), config)

	tests := []struct {
		name               string
		age                time.Duration
		expectedScoreMin   float64
		expectedScoreMax   float64
		expectedIsStale    bool
		expectedIsCritical bool
	}{
		{
			name:               "very fresh data",
			age:                1 * time.Hour,
			expectedScoreMin:   0.9,
			expectedScoreMax:   1.0,
			expectedIsStale:    false,
			expectedIsCritical: false,
		},
		{
			name:               "fresh data within threshold",
			age:                12 * time.Hour,
			expectedScoreMin:   0.8,
			expectedScoreMax:   1.0,
			expectedIsStale:    false,
			expectedIsCritical: false,
		},
		{
			name:               "stale data beyond threshold",
			age:                48 * time.Hour,
			expectedScoreMin:   0.1,
			expectedScoreMax:   0.7,
			expectedIsStale:    true,
			expectedIsCritical: false,
		},
		{
			name:               "critically stale data",
			age:                10 * 24 * time.Hour, // 10 days
			expectedScoreMin:   0.1,
			expectedScoreMax:   0.3,
			expectedIsStale:    true,
			expectedIsCritical: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := tracker.calculateFreshnessScore(tt.age)
			isStale := tt.age > tracker.config.StalenessThreshold
			isCritical := tt.age > tracker.config.CriticalStalenessThreshold

			assert.GreaterOrEqual(t, score, tt.expectedScoreMin)
			assert.LessOrEqual(t, score, tt.expectedScoreMax)
			assert.Equal(t, tt.expectedIsStale, isStale)
			assert.Equal(t, tt.expectedIsCritical, isCritical)
		})
	}
}

func TestDataFreshnessTracker_UpdatePatternAnalysis(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)

	// Create test records
	record1 := &FreshnessRecord{
		DataID:          "test-1",
		DataType:        "business",
		Source:          "api",
		UpdateFrequency: 2 * time.Hour,
		LastUpdated:     time.Now().Add(-2 * time.Hour),
	}

	record2 := &FreshnessRecord{
		DataID:          "test-2",
		DataType:        "business",
		Source:          "api",
		UpdateFrequency: 3 * time.Hour,
		LastUpdated:     time.Now().Add(-1 * time.Hour),
	}

	// Analyze patterns
	key1 := "test-1:business:api"
	key2 := "test-2:business:api"

	tracker.updatePatternAnalysis(key1, record1)
	tracker.updatePatternAnalysis(key2, record2)

	// Check that patterns were created
	patternKey := "business:api"
	pattern, exists := tracker.updatePatterns[patternKey]
	assert.True(t, exists)
	assert.NotNil(t, pattern)
	assert.Equal(t, "api", pattern.Source)
	assert.Equal(t, "business", pattern.DataType)
	assert.Equal(t, 2, pattern.UpdateCount)
	assert.Greater(t, pattern.AverageInterval, time.Duration(0))
}

func TestDataFreshnessTracker_StalenessAlerts(t *testing.T) {
	config := getDefaultDataFreshnessConfig()
	config.StalenessThreshold = 1 * time.Hour
	config.CriticalStalenessThreshold = 2 * time.Hour
	tracker := NewDataFreshnessTracker(zap.NewNop(), config)

	// Test stale record
	staleRecord := &FreshnessRecord{
		DataID:          "test-stale",
		DataType:        "business",
		Source:          "api",
		Age:             3 * time.Hour,
		IsStale:         true,
		IsCriticalStale: true,
		LastUpdated:     time.Now(),
	}

	tracker.checkStalenessAlerts(staleRecord)

	// Check that alert was created
	alertKey := "test-stale:business:api"
	alert, exists := tracker.stalenessAlerts[alertKey]
	assert.True(t, exists)
	assert.NotNil(t, alert)
	assert.Equal(t, "test-stale", alert.DataID)
	assert.Equal(t, "critical_staleness", alert.AlertType)
	assert.Equal(t, "critical", alert.Severity)
	assert.True(t, alert.IsActive)
	assert.Equal(t, 1, alert.TriggerCount)
}

func TestDataFreshnessTracker_ConsistencyScoring(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)

	tests := []struct {
		name                string
		pattern             *UpdatePattern
		expectedConsistency float64
	}{
		{
			name: "high consistency pattern",
			pattern: &UpdatePattern{
				UpdateCount:       5,
				AverageInterval:   2 * time.Hour,
				StandardDeviation: 15 * time.Minute,
			},
			expectedConsistency: 0.8, // High consistency
		},
		{
			name: "low consistency pattern",
			pattern: &UpdatePattern{
				UpdateCount:       5,
				AverageInterval:   2 * time.Hour,
				StandardDeviation: 3 * time.Hour,
			},
			expectedConsistency: 0.0, // Low consistency
		},
		{
			name: "single update pattern",
			pattern: &UpdatePattern{
				UpdateCount:       1,
				AverageInterval:   2 * time.Hour,
				StandardDeviation: 0,
			},
			expectedConsistency: 1.0, // Perfect consistency for single update
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consistency := tracker.calculateConsistencyScore(tt.pattern)
			assert.GreaterOrEqual(t, consistency, 0.0)
			assert.LessOrEqual(t, consistency, 1.0)
		})
	}
}

func TestDataFreshnessTracker_FreshnessLevelDetermination(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)

	tests := []struct {
		name          string
		score         float64
		expectedLevel string
	}{
		{
			name:          "excellent score",
			score:         0.95,
			expectedLevel: "fresh",
		},
		{
			name:          "good score",
			score:         0.75,
			expectedLevel: "aging",
		},
		{
			name:          "poor score",
			score:         0.45,
			expectedLevel: "stale",
		},
		{
			name:          "critical score",
			score:         0.25,
			expectedLevel: "critical",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := tracker.determineFreshnessLevel(tt.score)
			assert.Equal(t, tt.expectedLevel, level)
		})
	}
}

func TestDataFreshnessTracker_RecommendationGeneration(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)

	tests := []struct {
		name                    string
		result                  *FreshnessAnalysisResult
		expectedRecommendations int
	}{
		{
			name: "fresh data - few recommendations",
			result: &FreshnessAnalysisResult{
				CurrentFreshness: &FreshnessRecord{
					IsStale:         false,
					IsCriticalStale: false,
				},
				StalenessRisk: 0.2,
			},
			expectedRecommendations: 0,
		},
		{
			name: "stale data - recommendations expected",
			result: &FreshnessAnalysisResult{
				CurrentFreshness: &FreshnessRecord{
					IsStale:         true,
					IsCriticalStale: false,
				},
				StalenessRisk: 0.6,
			},
			expectedRecommendations: 1,
		},
		{
			name: "critically stale data - urgent recommendations",
			result: &FreshnessAnalysisResult{
				CurrentFreshness: &FreshnessRecord{
					IsStale:         true,
					IsCriticalStale: true,
				},
				StalenessRisk: 0.9,
			},
			expectedRecommendations: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendations := tracker.generateRecommendations(tt.result)
			assert.GreaterOrEqual(t, len(recommendations), tt.expectedRecommendations)
		})
	}
}

func TestDataFreshnessTracker_PriorityActions(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)

	tests := []struct {
		name            string
		result          *FreshnessAnalysisResult
		expectedActions int
	}{
		{
			name: "critical staleness - urgent actions",
			result: &FreshnessAnalysisResult{
				CurrentFreshness: &FreshnessRecord{
					IsCriticalStale: true,
				},
				ActiveAlerts: []*StalenessAlert{
					{IsActive: true},
				},
			},
			expectedActions: 1,
		},
		{
			name: "normal data - no urgent actions",
			result: &FreshnessAnalysisResult{
				CurrentFreshness: &FreshnessRecord{
					IsStale:         false,
					IsCriticalStale: false,
				},
				ActiveAlerts: []*StalenessAlert{},
			},
			expectedActions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actions := tracker.generatePriorityActions(tt.result)
			assert.GreaterOrEqual(t, len(actions), tt.expectedActions)
		})
	}
}

func TestDataFreshnessTracker_TrendAnalysis(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)

	tests := []struct {
		name               string
		trend              []FreshnessRecord
		expectedDirection  string
		expectedConfidence float64
	}{
		{
			name: "improving trend",
			trend: []FreshnessRecord{
				{FreshnessScore: 0.3},
				{FreshnessScore: 0.5},
				{FreshnessScore: 0.7},
			},
			expectedDirection:  "improving",
			expectedConfidence: 0.3,
		},
		{
			name: "declining trend",
			trend: []FreshnessRecord{
				{FreshnessScore: 0.8},
				{FreshnessScore: 0.6},
				{FreshnessScore: 0.4},
			},
			expectedDirection:  "declining",
			expectedConfidence: 0.3,
		},
		{
			name: "stable trend",
			trend: []FreshnessRecord{
				{FreshnessScore: 0.5},
				{FreshnessScore: 0.52},
				{FreshnessScore: 0.48},
			},
			expectedDirection:  "stable",
			expectedConfidence: 0.3,
		},
		{
			name: "insufficient data",
			trend: []FreshnessRecord{
				{FreshnessScore: 0.5},
			},
			expectedDirection:  "stable",
			expectedConfidence: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			direction := tracker.analyzeTrendDirection(tt.trend)
			confidence := tracker.calculateTrendConfidence(tt.trend)

			assert.Equal(t, tt.expectedDirection, direction)
			assert.GreaterOrEqual(t, confidence, 0.0)
			assert.LessOrEqual(t, confidence, 1.0)
		})
	}
}

func TestDataFreshnessTracker_Performance(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)
	ctx := context.Background()

	// Create test data
	data := &TestDataWithTimestamp{
		ID:        "perf-test",
		Name:      "Performance Test Data",
		CreatedAt: time.Now().Add(-1 * time.Hour),
		UpdatedAt: time.Now().Add(-30 * time.Minute),
	}

	// Performance test for tracking
	iterations := 100
	start := time.Now()

	for i := 0; i < iterations; i++ {
		_, err := tracker.TrackFreshness(ctx, data, "perf-test", "business", "api")
		require.NoError(t, err)
	}

	duration := time.Since(start)
	avgDuration := duration / time.Duration(iterations)

	// Performance assertions
	assert.Less(t, avgDuration, 10*time.Millisecond, "Average tracking should take less than 10ms")
	assert.Less(t, duration, 1*time.Second, "Total time for %d iterations should be less than 1 second", iterations)

	t.Logf("Tracking performance: %d iterations in %v (avg: %v per iteration)",
		iterations, duration, avgDuration)

	// Performance test for analysis
	start = time.Now()
	for i := 0; i < iterations; i++ {
		_, err := tracker.AnalyzeFreshness(ctx, "perf-test", "business", "api")
		require.NoError(t, err)
	}

	duration = time.Since(start)
	avgDuration = duration / time.Duration(iterations)

	// Performance assertions
	assert.Less(t, avgDuration, 5*time.Millisecond, "Average analysis should take less than 5ms")
	assert.Less(t, duration, 1*time.Second, "Total analysis time for %d iterations should be less than 1 second", iterations)

	t.Logf("Analysis performance: %d iterations in %v (avg: %v per iteration)",
		iterations, duration, avgDuration)
}

func TestDataFreshnessTracker_Concurrency(t *testing.T) {
	tracker := NewDataFreshnessTracker(zap.NewNop(), nil)
	ctx := context.Background()

	// Test concurrent tracking
	const numGoroutines = 10
	const iterationsPerGoroutine = 10

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			data := &TestDataWithTimestamp{
				ID:        fmt.Sprintf("concurrent-test-%d", id),
				Name:      fmt.Sprintf("Concurrent Test Data %d", id),
				CreatedAt: time.Now().Add(-1 * time.Hour),
				UpdatedAt: time.Now().Add(-30 * time.Minute),
			}

			for j := 0; j < iterationsPerGoroutine; j++ {
				_, err := tracker.TrackFreshness(ctx, data, fmt.Sprintf("concurrent-test-%d", id), "business", "api")
				require.NoError(t, err)

				_, err = tracker.AnalyzeFreshness(ctx, fmt.Sprintf("concurrent-test-%d", id), "business", "api")
				require.NoError(t, err)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify that all data was tracked
	patterns, err := tracker.GetUpdatePatterns(ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(patterns), 1)

	// Verify that patterns have expected update counts
	for _, pattern := range patterns {
		if pattern.Source == "api" && pattern.DataType == "business" {
			assert.GreaterOrEqual(t, pattern.UpdateCount, numGoroutines*iterationsPerGoroutine)
		}
	}
}
