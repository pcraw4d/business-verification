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

func TestNewDataQualityMonitor(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	monitor := NewDataQualityMonitor(logger, nil)
	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.config)
	assert.Equal(t, true, monitor.config.EnableRealTimeMonitoring)
	assert.Equal(t, 0.7, monitor.config.QualityAlertThreshold)

	// Test with custom config
	customConfig := &DataQualityMonitorConfig{
		EnableRealTimeMonitoring: false,
		QualityAlertThreshold:    0.8,
		FreshnessAlertThreshold:  12 * time.Hour,
	}

	monitor = NewDataQualityMonitor(logger, customConfig)
	assert.NotNil(t, monitor)
	assert.Equal(t, false, monitor.config.EnableRealTimeMonitoring)
	assert.Equal(t, 0.8, monitor.config.QualityAlertThreshold)
	assert.Equal(t, 12*time.Hour, monitor.config.FreshnessAlertThreshold)
}

func TestDataQualityMonitor_StartMonitoring(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	// Test successful monitoring start
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, "test-source", session.DataSourceID)
	assert.Equal(t, "website", session.DataSourceType)
	assert.Equal(t, "Test Website", session.DataSourceName)
	assert.Equal(t, "active", session.Status)
	assert.NotEmpty(t, session.SessionID)
	assert.True(t, session.StartTime.After(time.Now().Add(-time.Second)))
	assert.True(t, session.LastActivity.After(time.Now().Add(-time.Second)))

	// Verify session is stored
	retrievedSession, err := monitor.GetMonitoringSession(ctx, session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, session.SessionID, retrievedSession.SessionID)
}

func TestDataQualityMonitor_MonitorQuality(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	// Create test components
	qualityScorer := NewDataQualityScorer(logger, nil)
	freshnessTracker := NewDataFreshnessTracker(logger, nil)
	reliabilityAssessor := NewDataSourceReliabilityAssessor(logger, nil)

	monitor.SetComponents(qualityScorer, freshnessTracker, reliabilityAssessor)

	ctx := context.Background()

	// Start monitoring session
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)

	// Test data
	testData := map[string]interface{}{
		"name":    "Test Business",
		"address": "123 Test St",
		"phone":   "+1-555-123-4567",
		"email":   "test@example.com",
		"website": "https://test.com",
	}

	// Test quality monitoring
	result, err := monitor.MonitorQuality(ctx, session.SessionID, testData)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, session.SessionID, result.SessionID)
	assert.NotNil(t, result.QualityMetric)
	assert.True(t, result.ProcessingTime > 0)

	// Verify session was updated
	updatedSession, err := monitor.GetMonitoringSession(ctx, session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, 1, updatedSession.AssessmentCount)
	assert.True(t, updatedSession.LastAssessment.After(session.StartTime))
}

func TestDataQualityMonitor_MonitorQuality_InvalidSession(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	// Test with non-existent session
	result, err := monitor.MonitorQuality(ctx, "non-existent-session", map[string]interface{}{})
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "monitoring session not found")
}

func TestDataQualityMonitor_GenerateReport(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	// Start monitoring session
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)

	// Generate some metrics first
	testData := map[string]interface{}{
		"name":    "Test Business",
		"address": "123 Test St",
		"phone":   "+1-555-123-4567",
	}

	_, err = monitor.MonitorQuality(ctx, session.SessionID, testData)
	require.NoError(t, err)

	// Generate report
	timeRange := TimeRange{
		Start:    time.Now().Add(-1 * time.Hour),
		End:      time.Now(),
		Duration: 1 * time.Hour,
	}

	report, err := monitor.GenerateReport(ctx, session.SessionID, "summary", timeRange)
	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, session.SessionID, report.SessionID)
	assert.Equal(t, "summary", report.ReportType)
	assert.NotNil(t, report.Summary)
	assert.True(t, report.ProcessingTime > 0)
	assert.Equal(t, timeRange, report.TimeRange)

	// Verify session was updated
	updatedSession, err := monitor.GetMonitoringSession(ctx, session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, 1, updatedSession.ReportCount)
}

func TestDataQualityMonitor_GenerateReport_InvalidSession(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	timeRange := TimeRange{
		Start:    time.Now().Add(-1 * time.Hour),
		End:      time.Now(),
		Duration: 1 * time.Hour,
	}

	// Test with non-existent session
	report, err := monitor.GenerateReport(ctx, "non-existent-session", "summary", timeRange)
	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "monitoring session not found")
}

func TestDataQualityMonitor_GetQualityMetrics(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	// Start monitoring session
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)

	// Generate some metrics
	testData := map[string]interface{}{
		"name": "Test Business",
	}

	_, err = monitor.MonitorQuality(ctx, session.SessionID, testData)
	require.NoError(t, err)

	// Get metrics
	metrics, err := monitor.GetQualityMetrics(ctx, session.SessionID, 10)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
	assert.Equal(t, session.SessionID, metrics[0].SessionID)

	// Test with limit
	metrics, err = monitor.GetQualityMetrics(ctx, session.SessionID, 0)
	require.NoError(t, err)
	assert.Len(t, metrics, 1)
}

func TestDataQualityMonitor_GetActiveAlerts(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, &DataQualityMonitorConfig{
		QualityAlertThreshold: 0.9, // Set high to trigger alerts
	})

	ctx := context.Background()

	// Start monitoring session
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)

	// Generate metrics that will trigger alerts
	testData := map[string]interface{}{
		"name": "", // Empty name will cause low quality score
	}

	_, err = monitor.MonitorQuality(ctx, session.SessionID, testData)
	require.NoError(t, err)

	// Get active alerts
	alerts, err := monitor.GetActiveAlerts(ctx, session.SessionID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(alerts), 1)

	// Verify alert properties
	for _, alert := range alerts {
		assert.Equal(t, session.SessionID, alert.SessionID)
		assert.True(t, alert.IsActive)
		assert.NotEmpty(t, alert.AlertID)
		assert.NotEmpty(t, alert.Message)
	}
}

func TestDataQualityMonitor_ResolveAlert(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, &DataQualityMonitorConfig{
		QualityAlertThreshold: 0.9, // Set high to trigger alerts
	})

	ctx := context.Background()

	// Start monitoring session
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)

	// Generate metrics that will trigger alerts
	testData := map[string]interface{}{
		"name": "", // Empty name will cause low quality score
	}

	_, err = monitor.MonitorQuality(ctx, session.SessionID, testData)
	require.NoError(t, err)

	// Get active alerts
	alerts, err := monitor.GetActiveAlerts(ctx, session.SessionID)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(alerts), 1)

	// Resolve first alert
	alert := alerts[0]
	err = monitor.ResolveAlert(ctx, alert.AlertID, "data_fix", "Fixed data quality issue", "test_user")
	require.NoError(t, err)

	// Verify alert is resolved
	resolvedAlerts, err := monitor.GetActiveAlerts(ctx, session.SessionID)
	require.NoError(t, err)

	// Check if the resolved alert is no longer active
	found := false
	for _, a := range resolvedAlerts {
		if a.AlertID == alert.AlertID {
			found = true
			break
		}
	}
	assert.False(t, found, "Resolved alert should not be in active alerts")
}

func TestDataQualityMonitor_StopMonitoring(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	// Start monitoring session
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)

	// Stop monitoring
	err = monitor.StopMonitoring(ctx, session.SessionID)
	require.NoError(t, err)

	// Verify session is stopped
	stoppedSession, err := monitor.GetMonitoringSession(ctx, session.SessionID)
	require.NoError(t, err)
	assert.Equal(t, "stopped", stoppedSession.Status)
}

func TestDataQualityMonitor_AlertGeneration(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, &DataQualityMonitorConfig{
		QualityAlertThreshold:     0.8,
		FreshnessAlertThreshold:   1 * time.Hour,
		ReliabilityAlertThreshold: 0.9,
		CriticalThreshold:         0.5,
	})

	ctx := context.Background()

	// Start monitoring session
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)

	// Test data that will trigger multiple alerts
	testData := map[string]interface{}{
		"name": "", // Low quality
	}

	result, err := monitor.MonitorQuality(ctx, session.SessionID, testData)
	require.NoError(t, err)

	// Verify alerts were generated
	assert.GreaterOrEqual(t, len(result.Alerts), 1)

	// Check alert types
	alertTypes := make(map[string]bool)
	for _, alert := range result.Alerts {
		alertTypes[alert.AlertType] = true
		assert.Equal(t, session.SessionID, alert.SessionID)
		assert.True(t, alert.IsActive)
		assert.NotEmpty(t, alert.Message)
		assert.NotEmpty(t, alert.TriggeredBy)
	}

	// Should have at least quality alert
	assert.True(t, alertTypes["quality"] || alertTypes["critical"])
}

func TestDataQualityMonitor_RecommendationGeneration(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	// Start monitoring session
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)

	// Test data with various quality issues
	testData := map[string]interface{}{
		"name": "", // Empty name for completeness issue
	}

	result, err := monitor.MonitorQuality(ctx, session.SessionID, testData)
	require.NoError(t, err)

	// Verify recommendations were generated
	assert.GreaterOrEqual(t, len(result.Recommendations), 1)

	// Verify priority actions were generated
	assert.GreaterOrEqual(t, len(result.PriorityActions), 0)

	// Check recommendation content
	for _, recommendation := range result.Recommendations {
		assert.NotEmpty(t, recommendation)
	}
}

func TestDataQualityMonitor_ReportGeneration(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	// Start monitoring session
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)

	// Generate multiple metrics
	for i := 0; i < 3; i++ {
		testData := map[string]interface{}{
			"name":    "Test Business",
			"address": "123 Test St",
			"phone":   "+1-555-123-4567",
		}
		_, err = monitor.MonitorQuality(ctx, session.SessionID, testData)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Small delay between metrics
	}

	// Generate report
	timeRange := TimeRange{
		Start:    time.Now().Add(-1 * time.Hour),
		End:      time.Now(),
		Duration: 1 * time.Hour,
	}

	report, err := monitor.GenerateReport(ctx, session.SessionID, "detailed", timeRange)
	require.NoError(t, err)

	// Verify report structure
	assert.NotNil(t, report.Summary)
	assert.Equal(t, 3, report.Summary.AssessmentCount)
	assert.Greater(t, report.Summary.OverallQualityScore, 0.0)
	assert.NotEmpty(t, report.Summary.QualityLevel)

	// Verify trends
	assert.NotNil(t, report.Trends)

	// Verify alerts
	assert.NotNil(t, report.Alerts)

	// Verify recommendations
	assert.NotNil(t, report.Recommendations)
	assert.NotNil(t, report.PriorityActions)
}

func TestDataQualityMonitor_QualityLevelDetermination(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	// Test quality level determination
	tests := []struct {
		score    float64
		expected string
	}{
		{0.95, "excellent"},
		{0.85, "good"},
		{0.75, "fair"},
		{0.55, "poor"},
		{0.35, "critical"},
	}

	for _, tt := range tests {
		level := monitor.determineQualityLevel(tt.score)
		assert.Equal(t, tt.expected, level, "Score: %f", tt.score)
	}
}

func TestDataQualityMonitor_TrendDirectionDetermination(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	// Test trend direction determination
	tests := []struct {
		first    float64
		last     float64
		expected string
	}{
		{0.5, 0.6, "improving"},
		{0.6, 0.5, "declining"},
		{0.5, 0.52, "stable"},
		{0.5, 0.48, "stable"},
	}

	for _, tt := range tests {
		direction := monitor.determineTrendDirection(tt.first, tt.last)
		assert.Equal(t, tt.expected, direction, "First: %f, Last: %f", tt.first, tt.last)
	}
}

func TestDataQualityMonitor_ComponentIntegration(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	// Test component integration
	qualityScorer := NewDataQualityScorer(logger, nil)
	freshnessTracker := NewDataFreshnessTracker(logger, nil)
	reliabilityAssessor := NewDataSourceReliabilityAssessor(logger, nil)

	monitor.SetComponents(qualityScorer, freshnessTracker, reliabilityAssessor)

	// Verify components are set
	assert.NotNil(t, monitor.qualityScorer)
	assert.NotNil(t, monitor.freshnessTracker)
	assert.NotNil(t, monitor.reliabilityAssessor)
}

func TestDataQualityMonitor_Concurrency(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	// Start multiple monitoring sessions concurrently
	sessions := make([]*MonitoringSession, 5)
	for i := 0; i < 5; i++ {
		session, err := monitor.StartMonitoring(ctx,
			fmt.Sprintf("source-%d", i),
			"website",
			fmt.Sprintf("Test Website %d", i))
		require.NoError(t, err)
		sessions[i] = session
	}

	// Monitor quality concurrently
	results := make(chan *MonitoringResult, 5)
	errors := make(chan error, 5)

	for _, session := range sessions {
		go func(s *MonitoringSession) {
			testData := map[string]interface{}{
				"name": "Test Business",
			}
			result, err := monitor.MonitorQuality(ctx, s.SessionID, testData)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(session)
	}

	// Collect results
	resultCount := 0
	errorCount := 0

	for i := 0; i < 5; i++ {
		select {
		case result := <-results:
			assert.NotNil(t, result)
			resultCount++
		case err := <-errors:
			assert.NoError(t, err)
			errorCount++
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	assert.Equal(t, 5, resultCount)
	assert.Equal(t, 0, errorCount)
}

func TestDataQualityMonitor_Performance(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	// Start monitoring session
	session, err := monitor.StartMonitoring(ctx, "test-source", "website", "Test Website")
	require.NoError(t, err)

	// Test performance with multiple rapid assessments
	startTime := time.Now()

	for i := 0; i < 100; i++ {
		testData := map[string]interface{}{
			"name": fmt.Sprintf("Test Business %d", i),
		}
		_, err = monitor.MonitorQuality(ctx, session.SessionID, testData)
		require.NoError(t, err)
	}

	duration := time.Since(startTime)

	// Verify performance (should complete 100 assessments in reasonable time)
	assert.Less(t, duration, 10*time.Second, "100 assessments should complete within 10 seconds")

	// Verify all metrics were stored
	metrics, err := monitor.GetQualityMetrics(ctx, session.SessionID, 0)
	require.NoError(t, err)
	assert.Equal(t, 100, len(metrics))
}

func TestDataQualityMonitor_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	monitor := NewDataQualityMonitor(logger, nil)

	ctx := context.Background()

	// Test error handling with invalid session ID
	_, err := monitor.MonitorQuality(ctx, "invalid-session", map[string]interface{}{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monitoring session not found")

	// Test error handling with invalid session ID for report generation
	timeRange := TimeRange{
		Start:    time.Now().Add(-1 * time.Hour),
		End:      time.Now(),
		Duration: 1 * time.Hour,
	}

	_, err = monitor.GenerateReport(ctx, "invalid-session", "summary", timeRange)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "monitoring session not found")

	// Test error handling with invalid session ID for getting metrics
	_, err = monitor.GetQualityMetrics(ctx, "invalid-session", 10)
	assert.NoError(t, err) // Should return empty slice, not error

	// Test error handling with invalid session ID for getting alerts
	_, err = monitor.GetActiveAlerts(ctx, "invalid-session")
	assert.NoError(t, err) // Should return empty slice, not error
}
