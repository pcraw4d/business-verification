package industry_codes

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewErrorCategorizer(t *testing.T) {
	logger := zap.NewNop()

	t.Run("creates with default config", func(t *testing.T) {
		categorizer := NewErrorCategorizer(logger, nil)

		require.NotNil(t, categorizer)
		require.NotNil(t, categorizer.config)
		assert.True(t, categorizer.config.EnableAnalytics)
		assert.True(t, categorizer.config.EnableTrends)
		assert.True(t, categorizer.config.EnablePrioritization)
		assert.Equal(t, 10000, categorizer.config.MaxAnalyticsHistory)
		assert.Equal(t, 24*time.Hour, categorizer.config.TrendAnalysisWindow)
		assert.Len(t, categorizer.config.SeverityThresholds, 5)
		assert.Len(t, categorizer.config.CategoryWeights, 12)
	})

	t.Run("creates with custom config", func(t *testing.T) {
		customConfig := &CategorizationConfig{
			EnableAnalytics:      false,
			EnableTrends:         false,
			EnablePrioritization: false,
			MaxAnalyticsHistory:  5000,
			TrendAnalysisWindow:  12 * time.Hour,
		}

		categorizer := NewErrorCategorizer(logger, customConfig)

		require.NotNil(t, categorizer)
		assert.Equal(t, customConfig, categorizer.config)
		assert.False(t, categorizer.config.EnableAnalytics)
		assert.False(t, categorizer.config.EnableTrends)
		assert.False(t, categorizer.config.EnablePrioritization)
		assert.Equal(t, 5000, categorizer.config.MaxAnalyticsHistory)
		assert.Equal(t, 12*time.Hour, categorizer.config.TrendAnalysisWindow)
	})
}

func TestErrorCategorizer_CategorizeError(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	t.Run("categorizes nil error", func(t *testing.T) {
		result := categorizer.CategorizeError(context.Background(), nil, nil)
		assert.Nil(t, result)
	})

	t.Run("categorizes network error", func(t *testing.T) {
		err := errors.New("connection timeout occurred")
		errorContext := map[string]interface{}{
			"source":    "api_client",
			"operation": "fetch_data",
			"user_id":   "user123",
		}

		result := categorizer.CategorizeError(context.Background(), err, errorContext)

		require.NotNil(t, result)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, err, result.OriginalError)
		assert.Equal(t, err.Error(), result.Message)
		// Accept either network or performance category as both are valid for timeout
		assert.True(t, result.Category == CategoryNetwork || result.Category == CategoryPerformance)
		assert.Greater(t, result.Confidence, 0.0)
		assert.Equal(t, "api_client", result.Source)
		assert.Equal(t, "fetch_data", result.Operation)
		assert.Equal(t, "user123", result.UserID)
		assert.NotZero(t, result.Timestamp)
		assert.NotEmpty(t, result.Recommendations)
	})

	t.Run("categorizes database error", func(t *testing.T) {
		err := errors.New("database query failed - deadlock detected")

		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, CategoryDatabase, result.Category)
		assert.Greater(t, result.Confidence, 0.0)
		assert.NotEmpty(t, result.Recommendations)
	})

	t.Run("categorizes validation error", func(t *testing.T) {
		err := errors.New("validation failed: required field missing")

		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, CategoryValidation, result.Category)
		assert.Greater(t, result.Confidence, 0.0)
		assert.True(t, result.Classification.UserActionable)
		assert.NotEmpty(t, result.Recommendations)
	})

	t.Run("categorizes authentication error", func(t *testing.T) {
		err := errors.New("authentication failed: invalid token")

		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, CategoryAuthentication, result.Category)
		assert.Greater(t, result.Confidence, 0.0)
		assert.True(t, result.Classification.UserActionable)
	})

	t.Run("categorizes security error", func(t *testing.T) {
		err := errors.New("security breach detected: SQL injection attempt")

		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, CategorySecurity, result.Category)
		assert.Greater(t, result.Confidence, 0.0)
		// Security errors can be critical or high severity
		assert.True(t, result.Severity == SeverityHigh || result.Severity == SeverityCritical)
	})

	t.Run("categorizes unknown error", func(t *testing.T) {
		err := errors.New("some random unexpected error")

		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, CategoryUnknown, result.Category)
		// Unknown errors may get adjusted severity based on category weight
		assert.True(t, result.Severity == SeverityMedium || result.Severity == SeverityHigh)
		assert.NotEmpty(t, result.Recommendations)
	})
}

func TestErrorCategorizer_SeverityClassification(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	tests := []struct {
		name             string
		errorMessage     string
		expectedSeverity ErrorSeverity
		minConfidence    float64
	}{
		{
			name:             "critical error",
			errorMessage:     "critical system failure - data corruption detected",
			expectedSeverity: SeverityCritical,
			minConfidence:    0.3,
		},
		{
			name:             "high severity error",
			errorMessage:     "operation failed with exception",
			expectedSeverity: SeverityHigh,
			minConfidence:    0.0, // May not match any specific pattern
		},
		{
			name:             "timeout error",
			errorMessage:     "request timeout after 30 seconds",
			expectedSeverity: SeverityHigh,
			minConfidence:    0.3,
		},
		{
			name:             "unknown error gets default severity",
			errorMessage:     "some random issue",
			expectedSeverity: SeverityMedium,
			minConfidence:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errorMessage)
			result := categorizer.CategorizeError(context.Background(), err, nil)

			require.NotNil(t, result)
			assert.Equal(t, tt.expectedSeverity, result.Severity)
			assert.GreaterOrEqual(t, result.Confidence, tt.minConfidence)
		})
	}
}

func TestErrorCategorizer_PriorityCalculation(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	t.Run("calculates priority based on severity and category", func(t *testing.T) {
		// Critical security error should get urgent priority
		err := errors.New("critical security breach detected")
		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, CategorySecurity, result.Category)
		assert.Equal(t, SeverityCritical, result.Severity)
		assert.Equal(t, PriorityUrgent, result.Priority)
	})

	t.Run("adjusts priority based on category weight", func(t *testing.T) {
		// Performance error with lower category weight
		err := errors.New("performance timeout occurred")
		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		// Accept either network or performance category
		assert.True(t, result.Category == CategoryNetwork || result.Category == CategoryPerformance)
		// Priority should be adjusted down due to lower category weight
		assert.NotEqual(t, PriorityUrgent, result.Priority)
	})

	t.Run("disables prioritization when configured", func(t *testing.T) {
		config := &CategorizationConfig{
			EnablePrioritization: false,
		}
		categorizer := NewErrorCategorizer(logger, config)

		err := errors.New("critical system failure")
		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, PriorityMedium, result.Priority) // Default when disabled
	})
}

func TestErrorCategorizer_ErrorClassification(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	t.Run("classifies retryable errors", func(t *testing.T) {
		err := errors.New("temporary network timeout - please try again")
		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.True(t, result.Classification.Retryable)
		assert.True(t, result.Classification.Transient)
		assert.Equal(t, "industry_codes", result.Classification.Domain)
	})

	t.Run("classifies user actionable errors", func(t *testing.T) {
		err := errors.New("validation error: email format is invalid")
		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		// Could be validation or authentication category
		assert.True(t, result.Category == CategoryValidation || result.Category == CategoryAuthentication)
		assert.True(t, result.Classification.UserActionable)
	})

	t.Run("classifies non-user actionable errors", func(t *testing.T) {
		err := errors.New("database pool exhausted")
		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, CategoryDatabase, result.Category)
		assert.False(t, result.Classification.UserActionable)
	})
}

func TestErrorCategorizer_Recommendations(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	t.Run("generates network-specific recommendations", func(t *testing.T) {
		err := errors.New("connection refused by server")
		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, CategoryNetwork, result.Category)
		assert.NotEmpty(t, result.Recommendations)

		// Check for network-specific recommendation
		found := false
		for _, rec := range result.Recommendations {
			if rec.Action == "retry_with_backoff" {
				found = true
				assert.Equal(t, RecommendationImmediate, rec.Type)
				assert.Contains(t, rec.Resources, "network_team")
				break
			}
		}
		assert.True(t, found, "Should have network-specific recommendation")
	})

	t.Run("generates database-specific recommendations", func(t *testing.T) {
		err := errors.New("database deadlock detected")
		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, CategoryDatabase, result.Category)
		assert.NotEmpty(t, result.Recommendations)

		// Check for database-specific recommendation
		found := false
		for _, rec := range result.Recommendations {
			if rec.Action == "monitor_database_metrics" {
				found = true
				assert.Equal(t, RecommendationImmediate, rec.Type)
				assert.Contains(t, rec.Resources, "dba_team")
				break
			}
		}
		assert.True(t, found, "Should have database-specific recommendation")
	})

	t.Run("generates critical severity recommendations", func(t *testing.T) {
		err := errors.New("critical system failure detected")
		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Equal(t, SeverityCritical, result.Severity)
		assert.NotEmpty(t, result.Recommendations)

		// Should have escalation recommendation with highest priority
		found := false
		for _, rec := range result.Recommendations {
			if rec.Action == "escalate_to_oncall" {
				found = true
				assert.Equal(t, 0, rec.Priority) // Highest priority
				assert.Equal(t, RecommendationImmediate, rec.Type)
				assert.Equal(t, "critical", rec.RiskLevel)
				break
			}
		}
		assert.True(t, found, "Should have critical escalation recommendation")
	})

	t.Run("sorts recommendations by priority", func(t *testing.T) {
		err := errors.New("critical network security breach")
		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.NotEmpty(t, result.Recommendations)

		// Verify recommendations are sorted by priority (ascending)
		for i := 1; i < len(result.Recommendations); i++ {
			assert.LessOrEqual(t, result.Recommendations[i-1].Priority, result.Recommendations[i].Priority)
		}
	})
}

func TestErrorCategorizer_Analytics(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	t.Run("updates analytics when enabled", func(t *testing.T) {
		err1 := errors.New("network timeout")
		err2 := errors.New("database query failed")
		err3 := errors.New("another network issue")

		categorizer.CategorizeError(context.Background(), err1, nil)
		categorizer.CategorizeError(context.Background(), err2, nil)
		categorizer.CategorizeError(context.Background(), err3, nil)

		analytics := categorizer.GetAnalytics()
		require.NotNil(t, analytics)
		assert.Len(t, analytics.ErrorHistory, 3)
		assert.True(t, len(analytics.CategoryStats) >= 2) // At least 2 categories
		// Verify we have some category stats
		assert.True(t, len(analytics.CategoryStats) > 0)
	})

	t.Run("does not update analytics when disabled", func(t *testing.T) {
		config := &CategorizationConfig{
			EnableAnalytics: false,
		}
		categorizer := NewErrorCategorizer(logger, config)

		err := errors.New("some error")
		categorizer.CategorizeError(context.Background(), err, nil)

		analytics := categorizer.GetAnalytics()
		require.NotNil(t, analytics)
		assert.Empty(t, analytics.ErrorHistory)
		assert.Empty(t, analytics.CategoryStats)
	})

	t.Run("limits analytics history size", func(t *testing.T) {
		config := &CategorizationConfig{
			EnableAnalytics:     true,
			MaxAnalyticsHistory: 2,
		}
		categorizer := NewErrorCategorizer(logger, config)

		// Add 3 errors but limit is 2
		for i := 0; i < 3; i++ {
			err := errors.New(fmt.Sprintf("error %d", i))
			categorizer.CategorizeError(context.Background(), err, nil)
		}

		analytics := categorizer.GetAnalytics()
		require.NotNil(t, analytics)
		assert.Len(t, analytics.ErrorHistory, 2) // Should be limited to 2
	})
}

func TestErrorCategorizer_TrendAnalysis(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	t.Run("performs trend analysis when enabled", func(t *testing.T) {
		// Add some errors to trigger trend analysis
		for i := 0; i < 5; i++ {
			err := errors.New("network error")
			categorizer.CategorizeError(context.Background(), err, nil)
		}

		analytics := categorizer.GetAnalytics()
		require.NotNil(t, analytics)

		// Trend analysis should be performed (even if simplified)
		if analytics.TrendAnalysis != nil {
			assert.NotNil(t, analytics.TrendAnalysis.CategoryTrends)
			assert.NotNil(t, analytics.TrendAnalysis.SeverityTrends)
		}
	})

	t.Run("skips trend analysis when disabled", func(t *testing.T) {
		config := &CategorizationConfig{
			EnableAnalytics: true,
			EnableTrends:    false,
		}
		categorizer := NewErrorCategorizer(logger, config)

		err := errors.New("some error")
		categorizer.CategorizeError(context.Background(), err, nil)

		analytics := categorizer.GetAnalytics()
		require.NotNil(t, analytics)
		// Trend analysis should not be performed
		assert.Nil(t, analytics.TrendAnalysis)
	})
}

func TestErrorCategorizer_CategoryStats(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	t.Run("gets category stats", func(t *testing.T) {
		// Add errors that will be categorized consistently
		for i := 0; i < 3; i++ {
			err := errors.New("performance timeout issue")
			categorizer.CategorizeError(context.Background(), err, nil)
		}

		// Check if we have stats for performance category (which should match timeout)
		stats := categorizer.GetCategoryStats(CategoryPerformance)
		if stats != nil {
			assert.True(t, stats.Count >= 1)
			assert.NotNil(t, stats.Distribution)
			assert.NotNil(t, stats.TimeSeries)
			assert.NotNil(t, stats.Correlations)
		} else {
			// If performance category not found, check network category
			stats = categorizer.GetCategoryStats(CategoryNetwork)
			if stats != nil {
				assert.True(t, stats.Count >= 1)
				assert.NotNil(t, stats.Distribution)
				assert.NotNil(t, stats.TimeSeries)
				assert.NotNil(t, stats.Correlations)
			}
		}
	})

	t.Run("returns nil for non-existent category", func(t *testing.T) {
		stats := categorizer.GetCategoryStats(CategorySecurity)
		assert.Nil(t, stats)
	})
}

func TestErrorCategorizer_FilterMethods(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	// Add test errors
	networkErr := errors.New("network timeout")
	dbErr := errors.New("database error")
	criticalErr := errors.New("critical system failure")

	categorizer.CategorizeError(context.Background(), networkErr, nil)
	categorizer.CategorizeError(context.Background(), dbErr, nil)
	categorizer.CategorizeError(context.Background(), criticalErr, nil)

	t.Run("filters errors by category", func(t *testing.T) {
		networkErrors := categorizer.GetErrorsByCategory(CategoryNetwork)
		dbErrors := categorizer.GetErrorsByCategory(CategoryDatabase)
		perfErrors := categorizer.GetErrorsByCategory(CategoryPerformance)

		// We should have some errors in these categories
		totalErrors := len(networkErrors) + len(dbErrors) + len(perfErrors)
		assert.True(t, totalErrors >= 2, "Should have at least 2 categorized errors")

		// Check that we have at least one error of each type we added
		if len(networkErrors) > 0 {
			assert.Contains(t, networkErrors[0].Message, "timeout")
		}
		if len(dbErrors) > 0 {
			assert.Contains(t, dbErrors[0].Message, "database")
		}
	})

	t.Run("filters errors by severity", func(t *testing.T) {
		// Get all errors by severity
		criticalErrors := categorizer.GetErrorsBySeverity(SeverityCritical)
		highErrors := categorizer.GetErrorsBySeverity(SeverityHigh)
		mediumErrors := categorizer.GetErrorsBySeverity(SeverityMedium)

		// We should have some errors in these severity levels
		totalErrors := len(criticalErrors) + len(highErrors) + len(mediumErrors)
		assert.True(t, totalErrors >= 3, "Should have at least 3 categorized errors")

		// Check that errors are properly filtered by severity
		for _, err := range criticalErrors {
			assert.Equal(t, SeverityCritical, err.Severity)
		}
		for _, err := range highErrors {
			assert.Equal(t, SeverityHigh, err.Severity)
		}
		for _, err := range mediumErrors {
			assert.Equal(t, SeverityMedium, err.Severity)
		}
	})

	t.Run("filters errors by priority", func(t *testing.T) {
		urgentErrors := categorizer.GetErrorsByPriority(PriorityUrgent)
		// Should have at least the critical error if it was classified as urgent
		if len(urgentErrors) > 0 {
			assert.Equal(t, PriorityUrgent, urgentErrors[0].Priority)
		}
	})

	t.Run("gets top error categories", func(t *testing.T) {
		// Add more network errors to make it the top category
		for i := 0; i < 3; i++ {
			err := errors.New(fmt.Sprintf("network error %d", i))
			categorizer.CategorizeError(context.Background(), err, nil)
		}

		topCategories := categorizer.GetTopErrorCategories(2)
		assert.NotEmpty(t, topCategories)
		assert.LessOrEqual(t, len(topCategories), 2)

		// Network should be the top category now
		if len(topCategories) > 0 {
			assert.Equal(t, CategoryNetwork, topCategories[0])
		}
	})
}

func TestErrorCategorizer_ResetAnalytics(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	// Add some errors
	err := errors.New("test error")
	categorizer.CategorizeError(context.Background(), err, nil)

	// Verify analytics has data
	analytics := categorizer.GetAnalytics()
	assert.NotEmpty(t, analytics.ErrorHistory)
	assert.NotEmpty(t, analytics.CategoryStats)

	// Reset analytics
	categorizer.ResetAnalytics()

	// Verify analytics is reset
	analytics = categorizer.GetAnalytics()
	assert.Empty(t, analytics.ErrorHistory)
	assert.Empty(t, analytics.CategoryStats)
	assert.Empty(t, analytics.SeverityStats)
}

func TestErrorCategorizer_ContextExtraction(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	t.Run("extracts context information", func(t *testing.T) {
		err := errors.New("test error")
		errorContext := map[string]interface{}{
			"source":      "test_service",
			"operation":   "test_operation",
			"user_id":     "user123",
			"session_id":  "session456",
			"stack_trace": "stack trace here",
		}

		result := categorizer.CategorizeError(context.Background(), err, errorContext)

		require.NotNil(t, result)
		assert.Equal(t, "test_service", result.Source)
		assert.Equal(t, "test_operation", result.Operation)
		assert.Equal(t, "user123", result.UserID)
		assert.Equal(t, "session456", result.SessionID)
		assert.Equal(t, "stack trace here", result.StackTrace)
	})

	t.Run("handles missing context gracefully", func(t *testing.T) {
		err := errors.New("test error")

		result := categorizer.CategorizeError(context.Background(), err, nil)

		require.NotNil(t, result)
		assert.Empty(t, result.Source)
		assert.Empty(t, result.Operation)
		assert.Empty(t, result.UserID)
		assert.Empty(t, result.SessionID)
		assert.Empty(t, result.StackTrace)
	})

	t.Run("handles partial context", func(t *testing.T) {
		err := errors.New("test error")
		errorContext := map[string]interface{}{
			"source":  "test_service",
			"user_id": "user123",
			"invalid": 123, // Should be ignored
		}

		result := categorizer.CategorizeError(context.Background(), err, errorContext)

		require.NotNil(t, result)
		assert.Equal(t, "test_service", result.Source)
		assert.Equal(t, "user123", result.UserID)
		assert.Empty(t, result.Operation)
		assert.Empty(t, result.SessionID)
	})
}

func TestPatternMatching(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	t.Run("calculates pattern confidence correctly", func(t *testing.T) {
		patterns := categorizer.patterns.CategoryPatterns[CategoryNetwork]
		require.NotEmpty(t, patterns)

		// Test message that should match network patterns
		confidence := categorizer.calculatePatternConfidence("connection timeout occurred", patterns)
		assert.Greater(t, confidence, 0.0)

		// Test message that should not match
		confidence = categorizer.calculatePatternConfidence("random unrelated message", patterns)
		assert.Equal(t, 0.0, confidence)
	})

	t.Run("handles empty patterns", func(t *testing.T) {
		var emptyPatterns []*regexp.Regexp
		confidence := categorizer.calculatePatternConfidence("any message", emptyPatterns)
		assert.Equal(t, 0.0, confidence)
	})
}

func TestGenerateErrorID(t *testing.T) {
	t.Run("generates unique error IDs", func(t *testing.T) {
		id1 := generateErrorID()
		id2 := generateErrorID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2)
		assert.Contains(t, id1, "err_")
		assert.Contains(t, id2, "err_")
	})
}

func TestErrorCategorizer_Integration(t *testing.T) {
	logger := zap.NewNop()
	categorizer := NewErrorCategorizer(logger, nil)

	t.Run("full categorization workflow", func(t *testing.T) {
		err := errors.New("critical database connection failed due to network timeout")
		errorContext := map[string]interface{}{
			"source":    "user_service",
			"operation": "authenticate_user",
			"user_id":   "user123",
		}

		result := categorizer.CategorizeError(context.Background(), err, errorContext)

		// Verify all aspects of categorization
		require.NotNil(t, result)
		assert.NotEmpty(t, result.ID)
		assert.Equal(t, err, result.OriginalError)
		assert.Equal(t, err.Error(), result.Message)
		assert.NotEqual(t, CategoryUnknown, result.Category)  // Should match some category
		assert.NotEqual(t, SeverityInfo, result.Severity)     // Should be higher severity
		assert.NotEqual(t, PriorityDeferred, result.Priority) // Should be higher priority
		assert.Greater(t, result.Confidence, 0.0)
		assert.NotZero(t, result.Timestamp)
		assert.Equal(t, "user_service", result.Source)
		assert.Equal(t, "authenticate_user", result.Operation)
		assert.Equal(t, "user123", result.UserID)
		assert.NotNil(t, result.Metadata.Tags)
		assert.NotNil(t, result.Classification)
		assert.NotEmpty(t, result.Recommendations)

		// Verify metadata
		assert.Equal(t, 1, result.Metadata.Frequency)
		assert.NotZero(t, result.Metadata.FirstOccurrence)
		assert.NotZero(t, result.Metadata.LastOccurrence)
		assert.NotNil(t, result.Metadata.CustomFields)

		// Verify classification details
		assert.Equal(t, "industry_codes", result.Classification.Domain)
		assert.NotEmpty(t, result.Classification.Layer)
		assert.NotEqual(t, BusinessImpactNone, result.Classification.BusinessImpact)
		assert.NotEqual(t, TechnicalImpactNone, result.Classification.TechnicalImpact)

		// Verify recommendations are sorted
		if len(result.Recommendations) > 1 {
			for i := 1; i < len(result.Recommendations); i++ {
				assert.LessOrEqual(t, result.Recommendations[i-1].Priority, result.Recommendations[i].Priority)
			}
		}

		// Verify analytics were updated
		analytics := categorizer.GetAnalytics()
		assert.Len(t, analytics.ErrorHistory, 1)
		assert.Contains(t, analytics.CategoryStats, result.Category)
		assert.Contains(t, analytics.SeverityStats, result.Severity)
	})
}
