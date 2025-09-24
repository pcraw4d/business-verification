package error_monitoring

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewErrorAnalyzer(t *testing.T) {
	logger := zap.NewNop()
	config := &ErrorAnalysisConfig{
		AnalysisWindow:            2 * time.Hour,
		PatternDetectionThreshold: 5,
		CorrelationThreshold:      0.8,
		RootCauseConfidence:       0.9,
		MaxAnalysisDepth:          10,
		EnableMachineLearning:     true,
		EnableTemporalAnalysis:    true,
		EnableDependencyAnalysis:  true,
		CacheAnalysisResults:      true,
		CacheTTL:                  1 * time.Hour,
	}

	analyzer := NewErrorAnalyzer(config, logger)

	assert.NotNil(t, analyzer)
	assert.Equal(t, config, analyzer.config)
	assert.Equal(t, logger, analyzer.logger)
	assert.NotNil(t, analyzer.errorPatterns)
	assert.NotNil(t, analyzer.rootCauses)
	assert.NotNil(t, analyzer.correlations)
	assert.NotNil(t, analyzer.analysisCache)
}

func TestNewErrorAnalyzer_DefaultConfig(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, nil)

	assert.NotNil(t, analyzer)
	assert.NotNil(t, analyzer.config)
	assert.Equal(t, 1*time.Hour, analyzer.config.AnalysisWindow)
	assert.Equal(t, 3, analyzer.config.PatternDetectionThreshold)
	assert.Equal(t, 0.7, analyzer.config.CorrelationThreshold)
	assert.Equal(t, 0.8, analyzer.config.RootCauseConfidence)
	assert.Equal(t, 5, analyzer.config.MaxAnalysisDepth)
	assert.False(t, analyzer.config.EnableMachineLearning)
	assert.True(t, analyzer.config.EnableTemporalAnalysis)
	assert.True(t, analyzer.config.EnableDependencyAnalysis)
	assert.True(t, analyzer.config.CacheAnalysisResults)
	assert.Equal(t, 30*time.Minute, analyzer.config.CacheTTL)
}

func TestErrorAnalyzer_AnalyzeErrors(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	// Create test errors
	now := time.Now()
	errors := []ErrorEntry{
		{
			Timestamp:    now.Add(-10 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
			UserID:       "user1",
			RequestID:    "req1",
		},
		{
			Timestamp:    now.Add(-5 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input data",
			Severity:     "medium",
			UserID:       "user2",
			RequestID:    "req2",
		},
		{
			Timestamp:    now.Add(-2 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
			UserID:       "user3",
			RequestID:    "req3",
		},
	}

	timeRange := TimeRange{
		Start: now.Add(-15 * time.Minute),
		End:   now,
	}

	ctx := context.Background()
	result, err := analyzer.AnalyzeErrors(ctx, "test_process", errors, timeRange)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test_process", result.ProcessName)
	assert.Equal(t, timeRange, result.TimeRange)
	assert.NotEmpty(t, result.ID)
	assert.NotZero(t, result.AnalysisTime)
	assert.NotNil(t, result.ErrorPatterns)
	assert.NotNil(t, result.RootCauses)
	assert.NotNil(t, result.Correlations)
	assert.NotNil(t, result.Recommendations)
	assert.NotNil(t, result.RiskAssessment)
	assert.NotNil(t, result.Trends)
	assert.NotNil(t, result.Metadata)
}

func TestErrorAnalyzer_AnalyzeErrors_Cache(t *testing.T) {
	config := &ErrorAnalysisConfig{
		CacheAnalysisResults: true,
		CacheTTL:             1 * time.Hour,
	}
	analyzer := NewErrorAnalyzer(config, zap.NewNop())

	now := time.Now()
	errors := []ErrorEntry{
		{
			Timestamp:    now.Add(-10 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
	}

	timeRange := TimeRange{
		Start: now.Add(-15 * time.Minute),
		End:   now,
	}

	ctx := context.Background()

	// First analysis
	result1, err := analyzer.AnalyzeErrors(ctx, "test_process", errors, timeRange)
	require.NoError(t, err)

	// Second analysis (should use cache)
	result2, err := analyzer.AnalyzeErrors(ctx, "test_process", errors, timeRange)
	require.NoError(t, err)

	// Results should be the same (cached)
	assert.Equal(t, result1.ID, result2.ID)
	assert.Equal(t, result1.AnalysisTime, result2.AnalysisTime)
}

func TestErrorAnalyzer_DetectErrorPatterns(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	now := time.Now()
	errors := []ErrorEntry{
		// Sequence pattern: network_timeout -> validation_error -> network_timeout
		{
			Timestamp:    now.Add(-10 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
		{
			Timestamp:    now.Add(-9 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input",
			Severity:     "medium",
		},
		{
			Timestamp:    now.Add(-8 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
		// Repeat the sequence
		{
			Timestamp:    now.Add(-5 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
		{
			Timestamp:    now.Add(-4 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input",
			Severity:     "medium",
		},
		{
			Timestamp:    now.Add(-3 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
	}

	patterns := analyzer.detectErrorPatterns(errors, "test_process")

	assert.NotEmpty(t, patterns)

	// Should detect sequence pattern
	foundSequence := false
	for _, pattern := range patterns {
		if pattern.PatternType == "sequence" {
			foundSequence = true
			assert.Contains(t, pattern.ErrorTypes, "network_timeout")
			assert.Contains(t, pattern.ErrorTypes, "validation_error")
			assert.Equal(t, "test_process", pattern.Processes[0])
			assert.True(t, pattern.Frequency > 0)
			assert.True(t, pattern.Confidence > 0)
		}
	}
	assert.True(t, foundSequence, "Should detect sequence pattern")
}

func TestErrorAnalyzer_DetectTemporalPatterns(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	// Create errors at specific hours to create temporal patterns
	baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	errors := []ErrorEntry{}

	// Add errors at hour 9 (peak hour)
	for i := 0; i < 5; i++ {
		errors = append(errors, ErrorEntry{
			Timestamp:    baseTime.Add(time.Duration(9+i) * time.Hour),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		})
	}

	// Add errors at hour 15 (peak hour)
	for i := 0; i < 4; i++ {
		errors = append(errors, ErrorEntry{
			Timestamp:    baseTime.Add(time.Duration(15+i) * time.Hour),
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input",
			Severity:     "medium",
		})
	}

	// Add scattered errors at other hours
	for i := 0; i < 3; i++ {
		errors = append(errors, ErrorEntry{
			Timestamp:    baseTime.Add(time.Duration(2+i*8) * time.Hour),
			ProcessName:  "test_process",
			ErrorType:    "timeout",
			ErrorMessage: "Request timeout",
			Severity:     "medium",
		})
	}

	patterns := analyzer.detectErrorPatterns(errors, "test_process")

	// Should detect temporal patterns
	foundTemporal := false
	for _, pattern := range patterns {
		if pattern.PatternType == "temporal" {
			foundTemporal = true
			assert.Equal(t, "Peak Hour Error Pattern", pattern.Name)
			assert.Contains(t, pattern.Description, "Peak error hours")
			assert.Equal(t, "test_process", pattern.Processes[0])
			assert.True(t, pattern.Frequency > 0)
			assert.True(t, pattern.Confidence > 0)

			// Check metadata for peak hours
			if peakHours, ok := pattern.Metadata["peak_hours"].([]int); ok {
				assert.NotEmpty(t, peakHours)
			}
		}
	}
	assert.True(t, foundTemporal, "Should detect temporal patterns")
}

func TestErrorAnalyzer_AnalyzeRootCauses(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	now := time.Now()
	errors := []ErrorEntry{
		{
			Timestamp:    now.Add(-10 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
		{
			Timestamp:    now.Add(-8 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "connection_failed",
			ErrorMessage: "Connection failed",
			Severity:     "high",
		},
		{
			Timestamp:    now.Add(-5 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input data",
			Severity:     "medium",
		},
	}

	patterns := []*ErrorPattern{}
	rootCauses := analyzer.analyzeRootCauses(errors, patterns, "test_process")

	assert.NotEmpty(t, rootCauses)

	// Should detect infrastructure root cause for network errors
	foundInfra := false
	for _, rootCause := range rootCauses {
		if rootCause.Category == "infrastructure" && rootCause.RootCause == "Network Connectivity Issues" {
			foundInfra = true
			assert.Equal(t, "test_process", rootCause.AffectedProcesses[0])
			assert.True(t, rootCause.Confidence > 0)
			assert.NotEmpty(t, rootCause.Evidence)
			assert.NotEmpty(t, rootCause.ContributingFactors)
			assert.NotEmpty(t, rootCause.Recommendations)
			assert.NotEmpty(t, rootCause.Timeline)
		}
	}
	assert.True(t, foundInfra, "Should detect infrastructure root cause")

	// Should detect application root cause for validation errors
	foundApp := false
	for _, rootCause := range rootCauses {
		if rootCause.Category == "application" && rootCause.RootCause == "Application Logic Errors" {
			foundApp = true
			assert.Equal(t, "test_process", rootCause.AffectedProcesses[0])
			assert.True(t, rootCause.Confidence > 0)
			assert.NotEmpty(t, rootCause.Evidence)
			assert.NotEmpty(t, rootCause.ContributingFactors)
			assert.NotEmpty(t, rootCause.Recommendations)
			assert.NotEmpty(t, rootCause.Timeline)
		}
	}
	assert.True(t, foundApp, "Should detect application root cause")
}

func TestErrorAnalyzer_AnalyzeErrorCorrelations(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	now := time.Now()
	errors := []ErrorEntry{
		{
			Timestamp:    now.Add(-10 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
		{
			Timestamp:    now.Add(-9 * time.Minute), // Close in time
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input",
			Severity:     "medium",
		},
		{
			Timestamp:    now.Add(-5 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
		{
			Timestamp:    now.Add(-4 * time.Minute), // Close in time
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input",
			Severity:     "medium",
		},
	}

	correlations := analyzer.analyzeErrorCorrelations(errors, "test_process")

	// Should detect correlations between network_timeout and validation_error
	foundCorrelation := false
	for _, correlation := range correlations {
		if (correlation.PrimaryError == "network_timeout" && correlation.SecondaryError == "validation_error") ||
			(correlation.PrimaryError == "validation_error" && correlation.SecondaryError == "network_timeout") {
			foundCorrelation = true
			assert.Equal(t, "temporal", correlation.CorrelationType)
			assert.True(t, correlation.Strength > 0)
			assert.True(t, correlation.Confidence > 0)
			assert.NotEmpty(t, correlation.Evidence)
		}
	}
	assert.True(t, foundCorrelation, "Should detect error correlations")
}

func TestErrorAnalyzer_GenerateRecommendations(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	// Create test patterns with mitigation strategies
	patterns := []*ErrorPattern{
		{
			ID:          "pattern1",
			PatternType: "sequence",
			Name:        "Test Pattern",
			Mitigation: &MitigationStrategy{
				ID:             "mit1",
				StrategyType:   "preventive",
				Name:           "Test Mitigation",
				Implementation: []string{"Implement retry logic", "Add circuit breaker"},
			},
		},
	}

	// Create test root causes with recommendations
	rootCauses := []*RootCauseAnalysis{
		{
			ID:              "root1",
			Category:        "infrastructure",
			RootCause:       "Network Issues",
			Recommendations: []string{"Check network connectivity", "Monitor latency"},
		},
	}

	// Create test correlations
	correlations := []*ErrorCorrelation{
		{
			ID:             "corr1",
			PrimaryError:   "error1",
			SecondaryError: "error2",
			Strength:       0.9, // High strength
		},
	}

	recommendations := analyzer.generateRecommendations(patterns, rootCauses, correlations)

	assert.NotEmpty(t, recommendations)

	// Should include recommendations from patterns
	assert.Contains(t, recommendations, "Implement retry logic")
	assert.Contains(t, recommendations, "Add circuit breaker")

	// Should include recommendations from root causes
	assert.Contains(t, recommendations, "Check network connectivity")
	assert.Contains(t, recommendations, "Monitor latency")

	// Should include recommendations from high-strength correlations
	assert.Contains(t, recommendations, "Investigate causal relationship between error1 and error2")
}

func TestErrorAnalyzer_PerformRiskAssessment(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	now := time.Now()
	errors := []ErrorEntry{}

	// Create many errors to trigger high risk
	for i := 0; i < 60; i++ {
		errors = append(errors, ErrorEntry{
			Timestamp:    now.Add(time.Duration(-i) * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		})
	}

	patterns := []*ErrorPattern{
		{
			ID:          "pattern1",
			PatternType: "sequence",
			Name:        "Frequent Pattern",
			Frequency:   15, // High frequency
			Confidence:  0.8,
		},
	}

	rootCauses := []*RootCauseAnalysis{
		{
			ID:         "root1",
			Category:   "infrastructure",
			RootCause:  "Critical Issue",
			Impact:     "high",
			Confidence: 0.9,
		},
	}

	riskAssessment := analyzer.performRiskAssessment(errors, patterns, rootCauses)

	assert.NotNil(t, riskAssessment)
	assert.True(t, riskAssessment.RiskScore > 0.5, "Should have high risk score")
	assert.Equal(t, "high", riskAssessment.OverallRisk)
	assert.NotEmpty(t, riskAssessment.RiskFactors)
	assert.NotNil(t, riskAssessment.ImpactAnalysis)
	assert.NotEmpty(t, riskAssessment.MitigationPriority)

	// Check risk factors
	foundHighErrorRate := false
	for _, factor := range riskAssessment.RiskFactors {
		if factor.Factor == "High Error Rate" {
			foundHighErrorRate = true
			assert.Equal(t, "high", factor.RiskLevel)
			assert.Equal(t, 0.8, factor.Probability)
		}
	}
	assert.True(t, foundHighErrorRate, "Should identify high error rate risk factor")
}

func TestErrorAnalyzer_PerformTrendAnalysis(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	now := time.Now()
	errors := []ErrorEntry{}

	// Create more errors in the second half of the time range (degrading trend)
	for i := 0; i < 5; i++ {
		errors = append(errors, ErrorEntry{
			Timestamp:    now.Add(-20 * time.Minute).Add(time.Duration(i) * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		})
	}

	for i := 0; i < 15; i++ { // More errors in second half
		errors = append(errors, ErrorEntry{
			Timestamp:    now.Add(-10 * time.Minute).Add(time.Duration(i) * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input",
			Severity:     "medium",
		})
	}

	timeRange := TimeRange{
		Start: now.Add(-30 * time.Minute),
		End:   now,
	}

	trends := analyzer.performErrorTrendAnalysis(errors, timeRange)

	assert.NotNil(t, trends)
	assert.Equal(t, "degrading", trends.OverallTrend)
	assert.True(t, trends.TrendConfidence > 0.5)
	assert.NotEmpty(t, trends.SeasonalPatterns)
	assert.NotEmpty(t, trends.CyclicalPatterns)
	assert.NotEmpty(t, trends.Predictions)

	// Check predictions
	foundPrediction := false
	for _, prediction := range trends.Predictions {
		if prediction.PredictionType == "error_rate" {
			foundPrediction = true
			assert.Equal(t, "degrading", prediction.Value)
			assert.True(t, prediction.Confidence > 0)
			assert.Equal(t, 24*time.Hour, prediction.Timeframe)
		}
	}
	assert.True(t, foundPrediction, "Should generate predictions")
}

func TestErrorAnalyzer_FilterErrorsByTimeRange(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	now := time.Now()
	errors := []ErrorEntry{
		{
			Timestamp:    now.Add(-5 * time.Minute), // Within range
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
		{
			Timestamp:    now.Add(-20 * time.Minute), // Outside range
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input",
			Severity:     "medium",
		},
		{
			Timestamp:    now.Add(5 * time.Minute), // Outside range (future)
			ProcessName:  "test_process",
			ErrorType:    "timeout",
			ErrorMessage: "Request timeout",
			Severity:     "medium",
		},
	}

	timeRange := TimeRange{
		Start: now.Add(-10 * time.Minute),
		End:   now,
	}

	filtered := analyzer.filterErrorsByTimeRange(errors, timeRange)

	assert.Len(t, filtered, 1)
	assert.Equal(t, "network_timeout", filtered[0].ErrorType)
}

func TestErrorAnalyzer_GroupErrorsByType(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	errors := []ErrorEntry{
		{
			Timestamp:    time.Now(),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
		{
			Timestamp:    time.Now(),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
		{
			Timestamp:    time.Now(),
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input",
			Severity:     "medium",
		},
	}

	groups := analyzer.groupErrorsByType(errors)

	assert.Len(t, groups, 2)
	assert.Len(t, groups["network_timeout"], 2)
	assert.Len(t, groups["validation_error"], 1)
}

func TestErrorAnalyzer_CalculateErrorCorrelation(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	now := time.Now()
	errors1 := []ErrorEntry{
		{
			Timestamp:    now.Add(-10 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
		{
			Timestamp:    now.Add(-5 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
	}

	errors2 := []ErrorEntry{
		{
			Timestamp:    now.Add(-9 * time.Minute), // Close to err1
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input",
			Severity:     "medium",
		},
		{
			Timestamp:    now.Add(-4 * time.Minute), // Close to err2
			ProcessName:  "test_process",
			ErrorType:    "validation_error",
			ErrorMessage: "Invalid input",
			Severity:     "medium",
		},
	}

	correlation := analyzer.calculateErrorCorrelation(errors1, errors2)

	assert.NotNil(t, correlation)
	assert.Equal(t, "network_timeout", correlation.PrimaryError)
	assert.Equal(t, "validation_error", correlation.SecondaryError)
	assert.Equal(t, "temporal", correlation.CorrelationType)
	assert.True(t, correlation.Strength > 0)
	assert.True(t, correlation.Confidence > 0)
	assert.Equal(t, "bidirectional", correlation.Direction)
	assert.NotEmpty(t, correlation.Evidence)
}

func TestErrorAnalyzer_EmptyErrors(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	timeRange := TimeRange{
		Start: time.Now().Add(-1 * time.Hour),
		End:   time.Now(),
	}

	ctx := context.Background()
	result, err := analyzer.AnalyzeErrors(ctx, "test_process", []ErrorEntry{}, timeRange)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.ErrorPatterns)
	assert.Empty(t, result.RootCauses)
	assert.Empty(t, result.Correlations)
	assert.Empty(t, result.Recommendations)
	assert.NotNil(t, result.RiskAssessment)
	assert.Equal(t, "low", result.RiskAssessment.OverallRisk)
	assert.NotNil(t, result.Trends)
	assert.Equal(t, "stable", result.Trends.OverallTrend)
}

func TestErrorAnalyzer_ConcurrentAccess(t *testing.T) {
	analyzer := NewErrorAnalyzer(nil, zap.NewNop())

	now := time.Now()
	errors := []ErrorEntry{
		{
			Timestamp:    now.Add(-5 * time.Minute),
			ProcessName:  "test_process",
			ErrorType:    "network_timeout",
			ErrorMessage: "Connection timeout",
			Severity:     "high",
		},
	}

	timeRange := TimeRange{
		Start: now.Add(-10 * time.Minute),
		End:   now,
	}

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			ctx := context.Background()
			result, err := analyzer.AnalyzeErrors(ctx, "test_process", errors, timeRange)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestErrorAnalyzer_IDGeneration(t *testing.T) {
	// Test ID generation functions
	analysisID := generateAnalysisID()
	patternID := generatePatternID()
	rootCauseID := generateRootCauseID()
	correlationID := generateCorrelationID()

	assert.NotEmpty(t, analysisID)
	assert.NotEmpty(t, patternID)
	assert.NotEmpty(t, rootCauseID)
	assert.NotEmpty(t, correlationID)

	assert.Contains(t, analysisID, "analysis_")
	assert.Contains(t, patternID, "pattern_")
	assert.Contains(t, rootCauseID, "rootcause_")
	assert.Contains(t, correlationID, "correlation_")

	// IDs should be unique
	assert.NotEqual(t, patternID, rootCauseID)
	assert.NotEqual(t, rootCauseID, correlationID)
}
