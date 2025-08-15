package observability

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewRealTimeDashboard(t *testing.T) {
	config := RealTimeDashboardConfig{
		MetricsUpdateInterval:    10 * time.Second,
		DashboardRefreshRate:     2 * time.Second,
		ConnectionTimeout:        60 * time.Second,
		MaxDataPoints:            5000,
		EnableRealTimeUpdates:    true,
		EnableHistoricalView:     true,
		MaxConcurrentConnections: 50,
		EnableCompression:        true,
		EnableCaching:            true,
		RequireAuthentication:    false,
		AllowedOrigins:           []string{"*"},
		APIKeyRequired:           false,
	}

	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, config, logger)

	if dashboard == nil {
		t.Fatal("Expected dashboard to be created, got nil")
	}

	if dashboard.config.MetricsUpdateInterval != 10*time.Second {
		t.Errorf("Expected metrics update interval to be 10 seconds, got %v", dashboard.config.MetricsUpdateInterval)
	}

	if dashboard.config.DashboardRefreshRate != 2*time.Second {
		t.Errorf("Expected dashboard refresh rate to be 2 seconds, got %v", dashboard.config.DashboardRefreshRate)
	}

	if dashboard.config.MaxDataPoints != 5000 {
		t.Errorf("Expected max data points to be 5000, got %d", dashboard.config.MaxDataPoints)
	}

	if !dashboard.config.EnableRealTimeUpdates {
		t.Error("Expected real-time updates to be enabled")
	}

	if !dashboard.config.EnableHistoricalView {
		t.Error("Expected historical view to be enabled")
	}
}

func TestNewRealTimeDashboardDefaults(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	if dashboard.config.MetricsUpdateInterval != 5*time.Second {
		t.Errorf("Expected default metrics update interval to be 5 seconds, got %v", dashboard.config.MetricsUpdateInterval)
	}

	if dashboard.config.DashboardRefreshRate != 1*time.Second {
		t.Errorf("Expected default dashboard refresh rate to be 1 second, got %v", dashboard.config.DashboardRefreshRate)
	}

	if dashboard.config.ConnectionTimeout != 30*time.Second {
		t.Errorf("Expected default connection timeout to be 30 seconds, got %v", dashboard.config.ConnectionTimeout)
	}

	if dashboard.config.MaxDataPoints != 1000 {
		t.Errorf("Expected default max data points to be 1000, got %d", dashboard.config.MaxDataPoints)
	}

	if dashboard.config.MaxConcurrentConnections != 100 {
		t.Errorf("Expected default max concurrent connections to be 100, got %d", dashboard.config.MaxConcurrentConnections)
	}
}

func TestRealTimeDashboardStartStop(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	ctx := context.Background()

	// Start the dashboard
	err := dashboard.Start(ctx)
	if err != nil {
		t.Errorf("Expected no error when starting dashboard, got %v", err)
	}

	// Wait a bit for goroutines to start
	time.Sleep(100 * time.Millisecond)

	// Stop the dashboard
	dashboard.Stop()

	// Wait a bit for goroutines to stop
	time.Sleep(100 * time.Millisecond)
}

func TestGetDashboardState(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Get initial dashboard state
	state := dashboard.GetDashboardState()

	if state == nil {
		t.Fatal("Expected dashboard state to be returned, got nil")
	}

	if state.CurrentMetrics == nil {
		t.Error("Expected current metrics to be initialized")
	}

	if state.PerformanceIndicators == nil {
		t.Error("Expected performance indicators to be initialized")
	}

	if state.SystemHealth == nil {
		t.Error("Expected system health to be initialized")
	}

	if state.UserActivity == nil {
		t.Error("Expected user activity to be initialized")
	}

	if state.ErrorAnalysis == nil {
		t.Error("Expected error analysis to be initialized")
	}
}

func TestGetRealTimeMetrics(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Track some requests to generate metrics
	ctx := context.Background()

	requests := []*SuccessRateRequest{
		{Endpoint: "/api/v1/business", UserID: "user1", Success: true, ResponseTime: 100 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/analytics", UserID: "user2", Success: true, ResponseTime: 150 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/business", UserID: "user3", Success: false, ErrorType: "timeout", ResponseTime: 5000 * time.Millisecond, Timestamp: time.Now()},
	}

	for _, req := range requests {
		successRateTracker.TrackRequest(ctx, req)
	}

	// Update metrics manually
	dashboard.updateMetrics()

	// Get real-time metrics
	metrics := dashboard.GetRealTimeMetrics()

	if metrics == nil {
		t.Fatal("Expected real-time metrics to be returned, got nil")
	}

	if metrics.TotalRequests != 3 {
		t.Errorf("Expected total requests to be 3, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 2 {
		t.Errorf("Expected successful requests to be 2, got %d", metrics.SuccessfulRequests)
	}

	if metrics.FailedRequests != 1 {
		t.Errorf("Expected failed requests to be 1, got %d", metrics.FailedRequests)
	}

	if metrics.OverallSuccessRate != 2.0/3.0 {
		t.Errorf("Expected success rate to be 0.666..., got %f", metrics.OverallSuccessRate)
	}
}

func TestGetPerformanceIndicators(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Track some requests
	ctx := context.Background()

	requests := []*SuccessRateRequest{
		{Endpoint: "/api/v1/business", UserID: "user1", Success: true, ResponseTime: 100 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/analytics", UserID: "user2", Success: true, ResponseTime: 150 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/business", UserID: "user3", Success: true, ResponseTime: 200 * time.Millisecond, Timestamp: time.Now()},
	}

	for _, req := range requests {
		successRateTracker.TrackRequest(ctx, req)
	}

	// Update metrics manually
	dashboard.updateMetrics()

	// Get performance indicators
	indicators := dashboard.GetPerformanceIndicators()

	if indicators == nil {
		t.Fatal("Expected performance indicators to be returned, got nil")
	}

	if indicators.SystemHealthScore <= 0 {
		t.Error("Expected system health score to be positive")
	}

	if indicators.APIHealthScore <= 0 {
		t.Error("Expected API health score to be positive")
	}

	if indicators.DatabaseHealthScore <= 0 {
		t.Error("Expected database health score to be positive")
	}

	if indicators.CacheHitRate < 0 || indicators.CacheHitRate > 1 {
		t.Errorf("Expected cache hit rate to be between 0 and 1, got %f", indicators.CacheHitRate)
	}

	if indicators.DatabaseQueryEfficiency < 0 || indicators.DatabaseQueryEfficiency > 1 {
		t.Errorf("Expected database query efficiency to be between 0 and 1, got %f", indicators.DatabaseQueryEfficiency)
	}
}

func TestGetSystemHealth(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Track some requests to generate health data
	ctx := context.Background()

	requests := []*SuccessRateRequest{
		{Endpoint: "/api/v1/business", UserID: "user1", Success: true, ResponseTime: 100 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/analytics", UserID: "user2", Success: true, ResponseTime: 150 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/business", UserID: "user3", Success: true, ResponseTime: 200 * time.Millisecond, Timestamp: time.Now()},
	}

	for _, req := range requests {
		successRateTracker.TrackRequest(ctx, req)
	}

	// Update metrics manually
	dashboard.updateMetrics()

	// Get system health
	health := dashboard.GetSystemHealth()

	if health == nil {
		t.Fatal("Expected system health to be returned, got nil")
	}

	if health.OverallStatus == "" {
		t.Error("Expected overall status to be set")
	}

	if health.StatusMessage == "" {
		t.Error("Expected status message to be set")
	}

	if health.APIStatus == "" {
		t.Error("Expected API status to be set")
	}

	if health.DatabaseStatus == "" {
		t.Error("Expected database status to be set")
	}

	if health.CacheStatus == "" {
		t.Error("Expected cache status to be set")
	}

	if health.QueueStatus == "" {
		t.Error("Expected queue status to be set")
	}
}

func TestRealTimeDashboardGetTopPerformingEndpoints(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Track requests for different endpoints with varying success rates
	ctx := context.Background()

	endpoints := []struct {
		endpoint    string
		successRate float64
		totalReqs   int
	}{
		{"/api/v1/business", 0.95, 20},
		{"/api/v1/analytics", 0.88, 15},
		{"/api/v1/reports", 0.92, 25},
		{"/api/v1/users", 0.78, 10},
		{"/api/v1/settings", 0.85, 12},
	}

	// Track requests for each endpoint
	for _, ep := range endpoints {
		successfulReqs := int(float64(ep.totalReqs) * ep.successRate)
		failedReqs := ep.totalReqs - successfulReqs

		// Add successful requests
		for i := 0; i < successfulReqs; i++ {
			req := &SuccessRateRequest{
				Endpoint:     ep.endpoint,
				UserID:       "user1",
				Success:      true,
				ResponseTime: 100 * time.Millisecond,
				Timestamp:    time.Now(),
			}
			successRateTracker.TrackRequest(ctx, req)
		}

		// Add failed requests
		for i := 0; i < failedReqs; i++ {
			req := &SuccessRateRequest{
				Endpoint:     ep.endpoint,
				UserID:       "user1",
				Success:      false,
				ErrorType:    "error",
				ResponseTime: 200 * time.Millisecond,
				Timestamp:    time.Now(),
			}
			successRateTracker.TrackRequest(ctx, req)
		}
	}

	// Get top performing endpoints
	topEndpoints := dashboard.GetTopPerformingEndpoints(3)

	if len(topEndpoints) != 3 {
		t.Errorf("Expected 3 top endpoints, got %d", len(topEndpoints))
	}

	// Check that they're sorted by success rate (descending)
	if topEndpoints[0].SuccessRate < topEndpoints[1].SuccessRate {
		t.Error("Expected top endpoints to be sorted by success rate (descending)")
	}

	if topEndpoints[1].SuccessRate < topEndpoints[2].SuccessRate {
		t.Error("Expected top endpoints to be sorted by success rate (descending)")
	}

	// Check that the highest success rate endpoint is first
	if topEndpoints[0].Endpoint != "/api/v1/business" {
		t.Errorf("Expected highest success rate endpoint to be first, got %s", topEndpoints[0].Endpoint)
	}
}

func TestRealTimeDashboardGetWorstPerformingEndpoints(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Track requests for different endpoints with varying success rates
	ctx := context.Background()

	endpoints := []struct {
		endpoint    string
		successRate float64
		totalReqs   int
	}{
		{"/api/v1/business", 0.95, 20},
		{"/api/v1/analytics", 0.88, 15},
		{"/api/v1/reports", 0.92, 25},
		{"/api/v1/users", 0.78, 10},
		{"/api/v1/settings", 0.85, 12},
	}

	// Track requests for each endpoint
	for _, ep := range endpoints {
		successfulReqs := int(float64(ep.totalReqs) * ep.successRate)
		failedReqs := ep.totalReqs - successfulReqs

		// Add successful requests
		for i := 0; i < successfulReqs; i++ {
			req := &SuccessRateRequest{
				Endpoint:     ep.endpoint,
				UserID:       "user1",
				Success:      true,
				ResponseTime: 100 * time.Millisecond,
				Timestamp:    time.Now(),
			}
			successRateTracker.TrackRequest(ctx, req)
		}

		// Add failed requests
		for i := 0; i < failedReqs; i++ {
			req := &SuccessRateRequest{
				Endpoint:     ep.endpoint,
				UserID:       "user1",
				Success:      false,
				ErrorType:    "error",
				ResponseTime: 200 * time.Millisecond,
				Timestamp:    time.Now(),
			}
			successRateTracker.TrackRequest(ctx, req)
		}
	}

	// Get worst performing endpoints
	worstEndpoints := dashboard.GetWorstPerformingEndpoints(3)

	if len(worstEndpoints) != 3 {
		t.Errorf("Expected 3 worst endpoints, got %d", len(worstEndpoints))
	}

	// Check that they're sorted by success rate (ascending)
	if worstEndpoints[0].SuccessRate > worstEndpoints[1].SuccessRate {
		t.Error("Expected worst endpoints to be sorted by success rate (ascending)")
	}

	if worstEndpoints[1].SuccessRate > worstEndpoints[2].SuccessRate {
		t.Error("Expected worst endpoints to be sorted by success rate (ascending)")
	}

	// Check that the lowest success rate endpoint is first
	if worstEndpoints[0].Endpoint != "/api/v1/users" {
		t.Errorf("Expected lowest success rate endpoint to be first, got %s", worstEndpoints[0].Endpoint)
	}
}

func TestGetErrorAnalysis(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Track requests with different error types
	ctx := context.Background()

	requests := []*SuccessRateRequest{
		{Endpoint: "/api/v1/business", UserID: "user1", Success: false, ErrorType: "timeout", ResponseTime: 5000 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/analytics", UserID: "user2", Success: false, ErrorType: "timeout", ResponseTime: 5000 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/business", UserID: "user3", Success: false, ErrorType: "validation", ResponseTime: 200 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/reports", UserID: "user4", Success: false, ErrorType: "database", ResponseTime: 1000 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/business", UserID: "user5", Success: true, ResponseTime: 100 * time.Millisecond, Timestamp: time.Now()},
	}

	for _, req := range requests {
		successRateTracker.TrackRequest(ctx, req)
	}

	// Get error analysis
	errorAnalysis := dashboard.GetErrorAnalysis()

	if errorAnalysis == nil {
		t.Fatal("Expected error analysis to be returned, got nil")
	}

	if errorAnalysis.TotalErrors != 4 {
		t.Errorf("Expected total errors to be 4, got %d", errorAnalysis.TotalErrors)
	}

	if errorAnalysis.ErrorRate != 0.8 {
		t.Errorf("Expected error rate to be 0.8, got %f", errorAnalysis.ErrorRate)
	}

	if len(errorAnalysis.MostCommonErrors) != 3 {
		t.Errorf("Expected 3 most common errors, got %d", len(errorAnalysis.MostCommonErrors))
	}

	// Check that errors are sorted by count (descending)
	if errorAnalysis.MostCommonErrors[0].Count < errorAnalysis.MostCommonErrors[1].Count {
		t.Error("Expected most common errors to be sorted by count (descending)")
	}

	// Check that timeout errors are first (most common)
	if errorAnalysis.MostCommonErrors[0].ErrorType != "timeout" {
		t.Errorf("Expected timeout errors to be first, got %s", errorAnalysis.MostCommonErrors[0].ErrorType)
	}
}

func TestCalculateAPIHealthScore(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Track requests with good performance
	ctx := context.Background()

	requests := []*SuccessRateRequest{
		{Endpoint: "/api/v1/business", UserID: "user1", Success: true, ResponseTime: 100 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/analytics", UserID: "user2", Success: true, ResponseTime: 150 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/business", UserID: "user3", Success: true, ResponseTime: 200 * time.Millisecond, Timestamp: time.Now()},
	}

	for _, req := range requests {
		successRateTracker.TrackRequest(ctx, req)
	}

	// Calculate API health score
	score := dashboard.calculateAPIHealthScore()

	if score <= 0 {
		t.Error("Expected API health score to be positive")
	}

	if score > 100 {
		t.Error("Expected API health score to be <= 100")
	}

	// With 100% success rate and good response times, score should be high
	if score < 90 {
		t.Errorf("Expected high API health score for good performance, got %f", score)
	}
}

func TestCalculateOverallStatus(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Test healthy status (95%+ success rate)
	ctx := context.Background()
	for i := 0; i < 20; i++ {
		req := &SuccessRateRequest{
			Endpoint:     "/api/v1/business",
			UserID:       "user1",
			Success:      i < 19, // 95% success rate
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    time.Now(),
		}
		successRateTracker.TrackRequest(ctx, req)
	}

	status := dashboard.calculateOverallStatus()
	if status != "healthy" {
		t.Errorf("Expected status to be 'healthy', got %s", status)
	}

	// Test warning status (90-95% success rate)
	successRateTracker = NewSuccessRateTracker(SuccessRateTrackerConfig{})
	dashboard.successRateTracker = successRateTracker

	for i := 0; i < 20; i++ {
		req := &SuccessRateRequest{
			Endpoint:     "/api/v1/business",
			UserID:       "user1",
			Success:      i < 18, // 90% success rate
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    time.Now(),
		}
		successRateTracker.TrackRequest(ctx, req)
	}

	status = dashboard.calculateOverallStatus()
	if status != "warning" {
		t.Errorf("Expected status to be 'warning', got %s", status)
	}

	// Test critical status (<90% success rate)
	successRateTracker = NewSuccessRateTracker(SuccessRateTrackerConfig{})
	dashboard.successRateTracker = successRateTracker

	for i := 0; i < 20; i++ {
		req := &SuccessRateRequest{
			Endpoint:     "/api/v1/business",
			UserID:       "user1",
			Success:      i < 15, // 75% success rate
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    time.Now(),
		}
		successRateTracker.TrackRequest(ctx, req)
	}

	status = dashboard.calculateOverallStatus()
	if status != "critical" {
		t.Errorf("Expected status to be 'critical', got %s", status)
	}
}

func TestGenerateStatusMessage(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Test healthy message
	ctx := context.Background()
	for i := 0; i < 20; i++ {
		req := &SuccessRateRequest{
			Endpoint:     "/api/v1/business",
			UserID:       "user1",
			Success:      i < 19, // 95% success rate
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    time.Now(),
		}
		successRateTracker.TrackRequest(ctx, req)
	}

	message := dashboard.generateStatusMessage()
	if message != "System operating normally" {
		t.Errorf("Expected message to be 'System operating normally', got %s", message)
	}

	// Test warning message
	successRateTracker = NewSuccessRateTracker(SuccessRateTrackerConfig{})
	dashboard.successRateTracker = successRateTracker

	for i := 0; i < 20; i++ {
		req := &SuccessRateRequest{
			Endpoint:     "/api/v1/business",
			UserID:       "user1",
			Success:      i < 18, // 90% success rate
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    time.Now(),
		}
		successRateTracker.TrackRequest(ctx, req)
	}

	message = dashboard.generateStatusMessage()
	if message != "System experiencing minor issues" {
		t.Errorf("Expected message to be 'System experiencing minor issues', got %s", message)
	}

	// Test critical message
	successRateTracker = NewSuccessRateTracker(SuccessRateTrackerConfig{})
	dashboard.successRateTracker = successRateTracker

	for i := 0; i < 20; i++ {
		req := &SuccessRateRequest{
			Endpoint:     "/api/v1/business",
			UserID:       "user1",
			Success:      i < 15, // 75% success rate
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    time.Now(),
		}
		successRateTracker.TrackRequest(ctx, req)
	}

	message = dashboard.generateStatusMessage()
	if message != "System experiencing critical issues" {
		t.Errorf("Expected message to be 'System experiencing critical issues', got %s", message)
	}
}

func TestGetUserActivityMetrics(t *testing.T) {
	successRateTracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	performanceMonitor := NewPerformanceMonitor(PerformanceMonitorConfig{})
	logger := zap.NewNop()

	dashboard := NewRealTimeDashboard(successRateTracker, performanceMonitor, RealTimeDashboardConfig{}, logger)

	// Get user activity metrics
	userActivity := dashboard.getUserActivityMetrics()

	if userActivity == nil {
		t.Fatal("Expected user activity metrics to be returned, got nil")
	}

	if userActivity.ActiveUsers <= 0 {
		t.Error("Expected active users to be positive")
	}

	if userActivity.TotalUsers <= 0 {
		t.Error("Expected total users to be positive")
	}

	if userActivity.NewUsersToday < 0 {
		t.Error("Expected new users today to be non-negative")
	}

	if userActivity.PeakConcurrency <= 0 {
		t.Error("Expected peak concurrency to be positive")
	}

	if userActivity.UserSessions <= 0 {
		t.Error("Expected user sessions to be positive")
	}

	if userActivity.AverageSessionDuration <= 0 {
		t.Error("Expected average session duration to be positive")
	}
}

func TestSortErrorSummaries(t *testing.T) {
	summaries := []*ErrorSummary{
		{ErrorType: "timeout", Count: 5},
		{ErrorType: "validation", Count: 10},
		{ErrorType: "database", Count: 3},
		{ErrorType: "network", Count: 8},
	}

	// Sort by count (descending)
	sortErrorSummaries(summaries)

	// Check that they're sorted correctly
	if summaries[0].Count != 10 {
		t.Errorf("Expected first error to have count 10, got %d", summaries[0].Count)
	}

	if summaries[0].ErrorType != "validation" {
		t.Errorf("Expected first error to be validation, got %s", summaries[0].ErrorType)
	}

	if summaries[1].Count != 8 {
		t.Errorf("Expected second error to have count 8, got %d", summaries[1].Count)
	}

	if summaries[1].ErrorType != "network" {
		t.Errorf("Expected second error to be network, got %s", summaries[1].ErrorType)
	}

	if summaries[2].Count != 5 {
		t.Errorf("Expected third error to have count 5, got %d", summaries[2].Count)
	}

	if summaries[3].Count != 3 {
		t.Errorf("Expected fourth error to have count 3, got %d", summaries[3].Count)
	}
}
