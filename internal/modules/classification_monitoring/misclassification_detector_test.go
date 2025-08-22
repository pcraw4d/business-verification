package classification_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewMisclassificationDetector(t *testing.T) {
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(nil, logger)

	assert.NotNil(t, detector)
	assert.NotNil(t, detector.config)
	assert.True(t, detector.config.EnablePatternDetection)
	assert.True(t, detector.config.EnableRootCauseAnalysis)
	assert.True(t, detector.config.EnableRealTimeDetection)
	assert.NotNil(t, detector.rootCauseAnalyzer)
}

func TestMisclassificationDetector_DetectMisclassification(t *testing.T) {
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(nil, logger)

	tests := []struct {
		name             string
		event            *ClassificationEvent
		expectRecord     bool
		expectedType     string
		expectedSeverity string
	}{
		{
			name: "correct classification - no record",
			event: &ClassificationEvent{
				ID:               "correct-1",
				BusinessName:     "Test Company",
				ExpectedIndustry: "technology",
				ActualIndustry:   "technology",
				ConfidenceScore:  0.9,
				IsCorrect:        true,
				Timestamp:        time.Now(),
			},
			expectRecord: false,
		},
		{
			name: "low confidence error",
			event: &ClassificationEvent{
				ID:               "error-1",
				BusinessName:     "Test Company",
				ExpectedIndustry: "technology",
				ActualIndustry:   "finance",
				ConfidenceScore:  0.3,
				Method:           "ml",
				IsCorrect:        false,
				Timestamp:        time.Now(),
			},
			expectRecord:     true,
			expectedType:     "low_confidence",
			expectedSeverity: "low",
		},
		{
			name: "high confidence error",
			event: &ClassificationEvent{
				ID:               "error-2",
				BusinessName:     "Financial Corp",
				ExpectedIndustry: "financial_services",
				ActualIndustry:   "technology",
				ConfidenceScore:  0.95,
				Method:           "ml",
				IsCorrect:        false,
				Timestamp:        time.Now(),
			},
			expectRecord:     true,
			expectedType:     "very_high_confidence_error",
			expectedSeverity: "high",
		},
		{
			name: "medium confidence error",
			event: &ClassificationEvent{
				ID:               "error-3",
				BusinessName:     "Retail Store",
				ExpectedIndustry: "retail",
				ActualIndustry:   "technology",
				ConfidenceScore:  0.6,
				Method:           "keyword",
				IsCorrect:        false,
				Timestamp:        time.Now(),
			},
			expectRecord:     true,
			expectedType:     "medium_confidence_error",
			expectedSeverity: "medium",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record, err := detector.DetectMisclassification(context.Background(), tt.event)
			assert.NoError(t, err)

			if tt.expectRecord {
				assert.NotNil(t, record)
				assert.Equal(t, tt.event.BusinessName, record.BusinessName)
				assert.Equal(t, tt.event.ExpectedIndustry, record.ExpectedClassification)
				assert.Equal(t, tt.event.ActualIndustry, record.ActualClassification)
				assert.Equal(t, tt.expectedType, record.ErrorType)
				assert.Equal(t, tt.expectedSeverity, record.Severity)
			} else {
				assert.Nil(t, record)
			}
		})
	}
}

func TestMisclassificationDetector_ErrorTypeClassification(t *testing.T) {
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(nil, logger)

	tests := []struct {
		confidence   float64
		expectedType string
	}{
		{0.1, "very_low_confidence"},
		{0.4, "low_confidence"},
		{0.6, "medium_confidence_error"},
		{0.8, "high_confidence_error"},
		{0.95, "very_high_confidence_error"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("confidence_%.2f", tt.confidence), func(t *testing.T) {
			event := &ClassificationEvent{
				ConfidenceScore: tt.confidence,
			}

			errorType := detector.classifyErrorType(event)
			assert.Equal(t, tt.expectedType, errorType)
		})
	}
}

func TestMisclassificationDetector_SeverityCalculation(t *testing.T) {
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(nil, logger)

	tests := []struct {
		name             string
		confidence       float64
		expectedIndustry string
		actualIndustry   string
		expectedSeverity string
	}{
		{
			name:             "high confidence error",
			confidence:       0.9,
			expectedIndustry: "technology",
			actualIndustry:   "finance",
			expectedSeverity: "high",
		},
		{
			name:             "high stakes industry error",
			confidence:       0.7,
			expectedIndustry: "financial_services",
			actualIndustry:   "technology",
			expectedSeverity: "high",
		},
		{
			name:             "medium confidence regular industry",
			confidence:       0.6,
			expectedIndustry: "retail",
			actualIndustry:   "technology",
			expectedSeverity: "medium",
		},
		{
			name:             "low confidence error",
			confidence:       0.3,
			expectedIndustry: "technology",
			actualIndustry:   "retail",
			expectedSeverity: "low",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &ClassificationEvent{
				ConfidenceScore:  tt.confidence,
				ExpectedIndustry: tt.expectedIndustry,
				ActualIndustry:   tt.actualIndustry,
			}

			severity := detector.calculateSeverity(event)
			assert.Equal(t, tt.expectedSeverity, severity)
		})
	}
}

func TestMisclassificationDetector_HighStakesIndustryDetection(t *testing.T) {
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(nil, logger)

	tests := []struct {
		industry     string
		isHighStakes bool
	}{
		{"financial_services", true},
		{"banking", true},
		{"healthcare", true},
		{"legal_services", true},
		{"government", true},
		{"technology", false},
		{"retail", false},
		{"manufacturing", false},
		{"entertainment", false},
	}

	for _, tt := range tests {
		t.Run(tt.industry, func(t *testing.T) {
			result := detector.isHighStakesIndustry(tt.industry)
			assert.Equal(t, tt.isHighStakes, result)
		})
	}
}

func TestMisclassificationDetector_TemporalPatternDetection(t *testing.T) {
	config := &DetectionConfig{
		EnableTemporalAnalysis: true,
		MinPatternOccurrences:  2,
	}
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(config, logger)

	// Add multiple errors at the same hour
	now := time.Now()
	sameHour := time.Date(now.Year(), now.Month(), now.Day(), 14, 0, 0, 0, now.Location())

	events := []*ClassificationEvent{
		{
			ID:               "temporal-1",
			ExpectedIndustry: "technology",
			ActualIndustry:   "finance",
			ConfidenceScore:  0.7,
			IsCorrect:        false,
			Timestamp:        sameHour,
		},
		{
			ID:               "temporal-2",
			ExpectedIndustry: "retail",
			ActualIndustry:   "technology",
			ConfidenceScore:  0.8,
			IsCorrect:        false,
			Timestamp:        sameHour.Add(30 * time.Minute),
		},
		{
			ID:               "temporal-3",
			ExpectedIndustry: "healthcare",
			ActualIndustry:   "technology",
			ConfidenceScore:  0.6,
			IsCorrect:        false,
			Timestamp:        sameHour.Add(45 * time.Minute),
		},
	}

	// Process events
	for _, event := range events {
		_, err := detector.DetectMisclassification(context.Background(), event)
		assert.NoError(t, err)
	}

	// Check for temporal patterns
	patterns := detector.GetDetectedPatterns()

	// Should detect an hourly pattern
	found := false
	for _, pattern := range patterns {
		if pattern.Type == "temporal" && pattern.Frequency >= 2 {
			found = true
			break
		}
	}
	assert.True(t, found, "Should detect temporal pattern")
}

func TestMisclassificationDetector_SemanticPatternDetection(t *testing.T) {
	config := &DetectionConfig{
		EnableSemanticAnalysis: true,
		MinPatternOccurrences:  2,
	}
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(config, logger)

	// Add semantically similar business names with same error pattern
	events := []*ClassificationEvent{
		{
			ID:               "semantic-1",
			BusinessName:     "Tech Solutions Inc",
			ExpectedIndustry: "technology",
			ActualIndustry:   "consulting",
			ConfidenceScore:  0.7,
			IsCorrect:        false,
			Timestamp:        time.Now(),
		},
		{
			ID:               "semantic-2",
			BusinessName:     "Tech Services Corp",
			ExpectedIndustry: "technology",
			ActualIndustry:   "consulting",
			ConfidenceScore:  0.8,
			IsCorrect:        false,
			Timestamp:        time.Now(),
		},
		{
			ID:               "semantic-3",
			BusinessName:     "Tech Consulting LLC",
			ExpectedIndustry: "technology",
			ActualIndustry:   "consulting",
			ConfidenceScore:  0.6,
			IsCorrect:        false,
			Timestamp:        time.Now(),
		},
	}

	// Process events
	for _, event := range events {
		_, err := detector.DetectMisclassification(context.Background(), event)
		assert.NoError(t, err)
	}

	// Check for semantic patterns
	patterns := detector.GetDetectedPatterns()

	// Should detect a semantic pattern
	found := false
	for _, pattern := range patterns {
		if pattern.Type == "semantic" && pattern.Frequency >= 2 {
			found = true
			assert.Contains(t, pattern.AffectedClassifications, "technology")
			assert.Contains(t, pattern.AffectedClassifications, "consulting")
			break
		}
	}
	assert.True(t, found, "Should detect semantic pattern")
}

func TestMisclassificationDetector_SemanticSimilarity(t *testing.T) {
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(nil, logger)

	tests := []struct {
		name1       string
		name2       string
		threshold   float64
		shouldMatch bool
	}{
		{
			name1:       "Tech Solutions Inc",
			name2:       "Tech Services Corp",
			threshold:   0.5,
			shouldMatch: true,
		},
		{
			name1:       "Apple Inc",
			name2:       "Orange Corp",
			threshold:   0.5,
			shouldMatch: false,
		},
		{
			name1:       "Microsoft Corporation",
			name2:       "Microsoft Corp",
			threshold:   0.7,
			shouldMatch: true,
		},
		{
			name1:       "Google LLC",
			name2:       "Facebook Inc",
			threshold:   0.5,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_vs_%s", tt.name1, tt.name2), func(t *testing.T) {
			similarity := detector.calculateSemanticSimilarity(tt.name1, tt.name2)
			if tt.shouldMatch {
				assert.GreaterOrEqual(t, similarity, tt.threshold)
			} else {
				assert.Less(t, similarity, tt.threshold)
			}
		})
	}
}

func TestMisclassificationDetector_ErrorFrequencyPatterns(t *testing.T) {
	config := &DetectionConfig{
		MinPatternOccurrences: 3,
	}
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(config, logger)

	// Add multiple instances of the same error type
	errorType := "technology->finance"
	for i := 0; i < 5; i++ {
		event := &ClassificationEvent{
			ID:               fmt.Sprintf("freq-%d", i),
			BusinessName:     fmt.Sprintf("Company %d", i),
			ExpectedIndustry: "technology",
			ActualIndustry:   "finance",
			ConfidenceScore:  0.7,
			IsCorrect:        false,
			Timestamp:        time.Now(),
		}

		_, err := detector.DetectMisclassification(context.Background(), event)
		assert.NoError(t, err)
	}

	// Trigger pattern analysis
	detector.performFullPatternAnalysis()

	// Check for frequency pattern
	patterns := detector.GetDetectedPatterns()

	found := false
	for _, pattern := range patterns {
		if pattern.Type == "frequency" && pattern.Frequency >= 3 {
			found = true
			assert.Contains(t, pattern.Description, errorType)
			break
		}
	}
	assert.True(t, found, "Should detect frequency pattern")
}

func TestMisclassificationDetector_GetStatistics(t *testing.T) {
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(nil, logger)

	// Add some test data
	events := []*ClassificationEvent{
		{
			ID:               "stat-1",
			ExpectedIndustry: "technology",
			ActualIndustry:   "finance",
			ConfidenceScore:  0.9,
			Method:           "ml",
			IsCorrect:        false,
			Timestamp:        time.Now(),
		},
		{
			ID:               "stat-2",
			ExpectedIndustry: "retail",
			ActualIndustry:   "technology",
			ConfidenceScore:  0.7,
			Method:           "keyword",
			IsCorrect:        false,
			Timestamp:        time.Now(),
		},
	}

	for _, event := range events {
		_, err := detector.DetectMisclassification(context.Background(), event)
		assert.NoError(t, err)
	}

	stats := detector.GetErrorStatistics()

	assert.Equal(t, 2, stats["total_errors"])
	assert.Contains(t, stats, "error_distribution")
	assert.Contains(t, stats, "severity_distribution")
	assert.Contains(t, stats, "time_range")

	errorDist := stats["error_distribution"].(map[string]int)
	assert.Equal(t, 1, errorDist["very_high_confidence_error"])
	assert.Equal(t, 1, errorDist["high_confidence_error"])
}

func TestMisclassificationDetector_TimeRangeQuery(t *testing.T) {
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(nil, logger)

	now := time.Now()
	oldTime := now.Add(-2 * time.Hour)
	recentTime := now.Add(-30 * time.Minute)

	// Add errors at different times
	events := []*ClassificationEvent{
		{
			ID:               "time-1",
			ExpectedIndustry: "technology",
			ActualIndustry:   "finance",
			IsCorrect:        false,
			Timestamp:        oldTime,
		},
		{
			ID:               "time-2",
			ExpectedIndustry: "retail",
			ActualIndustry:   "technology",
			IsCorrect:        false,
			Timestamp:        recentTime,
		},
		{
			ID:               "time-3",
			ExpectedIndustry: "healthcare",
			ActualIndustry:   "technology",
			IsCorrect:        false,
			Timestamp:        now,
		},
	}

	for _, event := range events {
		_, err := detector.DetectMisclassification(context.Background(), event)
		assert.NoError(t, err)
	}

	// Query recent errors (last hour)
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Minute)

	recentErrors := detector.GetMisclassificationsByTimeRange(start, end)
	assert.Len(t, recentErrors, 2) // Should include time-2 and time-3

	// Verify ordering (most recent first)
	assert.Equal(t, "time-3", recentErrors[0].ID)
	assert.Equal(t, "time-2", recentErrors[1].ID)
}

func TestMisclassificationDetector_PatternSeverityCalculation(t *testing.T) {
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(nil, logger)

	// Add 10 total errors to establish baseline
	for i := 0; i < 10; i++ {
		event := &ClassificationEvent{
			ID:               fmt.Sprintf("baseline-%d", i),
			ExpectedIndustry: "various",
			ActualIndustry:   "other",
			IsCorrect:        false,
			Timestamp:        time.Now(),
		}
		detector.DetectMisclassification(context.Background(), event)
	}

	tests := []struct {
		frequency        int
		expectedSeverity string
	}{
		{4, "critical"}, // 40% of total errors
		{3, "high"},     // 30% of total errors
		{2, "medium"},   // 20% of total errors
		{1, "low"},      // 10% of total errors
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("frequency_%d", tt.frequency), func(t *testing.T) {
			severity := detector.calculatePatternSeverity(tt.frequency)
			assert.Equal(t, tt.expectedSeverity, severity)
		})
	}
}

func TestNewRootCauseAnalyzer(t *testing.T) {
	config := DefaultDetectionConfig()
	logger := zap.NewNop()

	analyzer := NewRootCauseAnalyzer(config, logger)

	assert.NotNil(t, analyzer)
	assert.NotNil(t, analyzer.knowledgeBase)
	assert.Greater(t, len(analyzer.analysisRules), 0)
	assert.Greater(t, len(analyzer.knowledgeBase), 0)
}

func TestRootCauseAnalyzer_AnalyzeRootCause(t *testing.T) {
	config := DefaultDetectionConfig()
	logger := zap.NewNop()
	analyzer := NewRootCauseAnalyzer(config, logger)

	tests := []struct {
		name               string
		event              *ClassificationEvent
		expectedPrimary    string
		expectedConfidence float64
	}{
		{
			name: "high confidence error",
			event: &ClassificationEvent{
				ConfidenceScore: 0.95,
				IsCorrect:       false,
			},
			expectedPrimary:    "model_overfitting",
			expectedConfidence: 0.85,
		},
		{
			name: "low confidence error",
			event: &ClassificationEvent{
				ConfidenceScore: 0.3,
				IsCorrect:       false,
			},
			expectedPrimary:    "insufficient_training_data",
			expectedConfidence: 0.7,
		},
		{
			name: "feature dominance",
			event: &ClassificationEvent{
				ConfidenceScore: 0.7,
				IsCorrect:       false,
				FeatureImportance: map[string]float64{
					"dominant_feature": 0.85,
					"other_feature":    0.15,
				},
			},
			expectedPrimary: "feature_quality_issues",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := analyzer.AnalyzeRootCause(tt.event)

			assert.NotNil(t, analysis)
			assert.Equal(t, tt.expectedPrimary, analysis.PrimaryRoot)

			if tt.expectedConfidence > 0 {
				assert.Equal(t, tt.expectedConfidence, analysis.Confidence)
			}

			assert.NotEmpty(t, analysis.Evidence)
			assert.NotEmpty(t, analysis.Recommendations)
		})
	}
}

func TestRootCauseAnalyzer_DeduplicateRecommendations(t *testing.T) {
	config := DefaultDetectionConfig()
	logger := zap.NewNop()
	analyzer := NewRootCauseAnalyzer(config, logger)

	input := []string{
		"recommendation_1",
		"recommendation_2",
		"recommendation_1", // duplicate
		"recommendation_3",
		"recommendation_2", // duplicate
	}

	result := analyzer.deduplicateStrings(input)

	assert.Len(t, result, 3)
	assert.Contains(t, result, "recommendation_1")
	assert.Contains(t, result, "recommendation_2")
	assert.Contains(t, result, "recommendation_3")
}

func TestMisclassificationDetector_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	detector := NewMisclassificationDetector(nil, logger)

	const numGoroutines = 5
	const numEvents = 20

	done := make(chan bool, numGoroutines)

	// Process events concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for j := 0; j < numEvents; j++ {
				event := &ClassificationEvent{
					ID:               fmt.Sprintf("concurrent-%d-%d", goroutineID, j),
					BusinessName:     fmt.Sprintf("Company %d-%d", goroutineID, j),
					ExpectedIndustry: "technology",
					ActualIndustry:   "finance",
					ConfidenceScore:  0.7,
					IsCorrect:        false,
					Timestamp:        time.Now(),
				}

				_, err := detector.DetectMisclassification(context.Background(), event)
				assert.NoError(t, err)

				// Also test concurrent reads
				_ = detector.GetDetectedPatterns()
				_ = detector.GetErrorStatistics()
			}
		}(i)
	}

	// Wait for completion
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final state
	stats := detector.GetErrorStatistics()
	expectedTotal := numGoroutines * numEvents
	assert.Equal(t, expectedTotal, stats["total_errors"])
}
