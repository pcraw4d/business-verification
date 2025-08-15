package observability

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewSuccessRateTracker(t *testing.T) {
	config := SuccessRateTrackerConfig{
		TrackingWindow:       2 * time.Minute,
		RollingWindowSize:    12 * time.Hour,
		RetentionPeriod:      7 * 24 * time.Hour,
		CriticalThreshold:    0.85,
		WarningThreshold:     0.92,
		DegradationThreshold: 0.97,
		EnableTrendAnalysis:  true,
		MaxDataPoints:        5000,
	}

	tracker := NewSuccessRateTracker(config)

	if tracker == nil {
		t.Fatal("Expected tracker to be created, got nil")
	}

	if tracker.config.TrackingWindow != 2*time.Minute {
		t.Errorf("Expected tracking window to be 2 minutes, got %v", tracker.config.TrackingWindow)
	}

	if tracker.config.CriticalThreshold != 0.85 {
		t.Errorf("Expected critical threshold to be 0.85, got %f", tracker.config.CriticalThreshold)
	}

	if tracker.config.MaxDataPoints != 5000 {
		t.Errorf("Expected max data points to be 5000, got %d", tracker.config.MaxDataPoints)
	}
}

func TestNewSuccessRateTrackerDefaults(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})

	if tracker.config.TrackingWindow != 1*time.Minute {
		t.Errorf("Expected default tracking window to be 1 minute, got %v", tracker.config.TrackingWindow)
	}

	if tracker.config.CriticalThreshold != 0.90 {
		t.Errorf("Expected default critical threshold to be 0.90, got %f", tracker.config.CriticalThreshold)
	}

	if tracker.config.WarningThreshold != 0.95 {
		t.Errorf("Expected default warning threshold to be 0.95, got %f", tracker.config.WarningThreshold)
	}

	if tracker.config.MaxDataPoints != 10000 {
		t.Errorf("Expected default max data points to be 10000, got %d", tracker.config.MaxDataPoints)
	}
}

func TestTrackRequest(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	ctx := context.Background()

	// Track successful request
	req := &SuccessRateRequest{
		Endpoint:     "/api/v1/business",
		UserID:       "user123",
		Success:      true,
		ResponseTime: 150 * time.Millisecond,
		DataSize:     1024,
		Timestamp:    time.Now(),
	}

	err := tracker.TrackRequest(ctx, req)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Track failed request
	failedReq := &SuccessRateRequest{
		Endpoint:     "/api/v1/business",
		UserID:       "user123",
		Success:      false,
		ErrorType:    "timeout",
		ResponseTime: 5000 * time.Millisecond,
		DataSize:     0,
		Timestamp:    time.Now(),
	}

	err = tracker.TrackRequest(ctx, failedReq)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify overall stats
	overallStats := tracker.GetOverallSuccessRate()
	if overallStats.TotalRequests != 2 {
		t.Errorf("Expected total requests to be 2, got %d", overallStats.TotalRequests)
	}

	if overallStats.SuccessfulRequests != 1 {
		t.Errorf("Expected successful requests to be 1, got %d", overallStats.SuccessfulRequests)
	}

	if overallStats.FailedRequests != 1 {
		t.Errorf("Expected failed requests to be 1, got %d", overallStats.FailedRequests)
	}

	if overallStats.OverallSuccessRate != 0.5 {
		t.Errorf("Expected success rate to be 0.5, got %f", overallStats.OverallSuccessRate)
	}
}

func TestTrackRequestEndpointStats(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	ctx := context.Background()

	endpoint := "/api/v1/business"

	// Track multiple requests for the same endpoint
	requests := []*SuccessRateRequest{
		{Endpoint: endpoint, UserID: "user1", Success: true, ResponseTime: 100 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: endpoint, UserID: "user2", Success: true, ResponseTime: 150 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: endpoint, UserID: "user3", Success: false, ErrorType: "timeout", ResponseTime: 5000 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: endpoint, UserID: "user4", Success: true, ResponseTime: 200 * time.Millisecond, Timestamp: time.Now()},
	}

	for _, req := range requests {
		err := tracker.TrackRequest(ctx, req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}

	// Get endpoint stats
	stats, err := tracker.GetEndpointSuccessRate(endpoint)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if stats.TotalRequests != 4 {
		t.Errorf("Expected total requests to be 4, got %d", stats.TotalRequests)
	}

	if stats.SuccessfulRequests != 3 {
		t.Errorf("Expected successful requests to be 3, got %d", stats.SuccessfulRequests)
	}

	if stats.FailedRequests != 1 {
		t.Errorf("Expected failed requests to be 1, got %d", stats.FailedRequests)
	}

	if stats.OverallSuccessRate != 0.75 {
		t.Errorf("Expected success rate to be 0.75, got %f", stats.OverallSuccessRate)
	}

	if stats.TimeoutRequests != 1 {
		t.Errorf("Expected timeout requests to be 1, got %d", stats.TimeoutRequests)
	}

	if len(stats.ErrorBreakdown) != 1 {
		t.Errorf("Expected 1 error type, got %d", len(stats.ErrorBreakdown))
	}

	if stats.ErrorBreakdown["timeout"] != 1 {
		t.Errorf("Expected timeout errors to be 1, got %d", stats.ErrorBreakdown["timeout"])
	}
}

func TestTrackRequestUserStats(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	ctx := context.Background()

	userID := "user123"

	// Track multiple requests for the same user
	requests := []*SuccessRateRequest{
		{Endpoint: "/api/v1/business", UserID: userID, Success: true, ResponseTime: 100 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/analytics", UserID: userID, Success: true, ResponseTime: 150 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/business", UserID: userID, Success: false, ErrorType: "validation", ResponseTime: 200 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/reports", UserID: userID, Success: true, ResponseTime: 300 * time.Millisecond, Timestamp: time.Now()},
	}

	for _, req := range requests {
		err := tracker.TrackRequest(ctx, req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}

	// Get user stats
	stats, err := tracker.GetUserSuccessRate(userID)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if stats.TotalRequests != 4 {
		t.Errorf("Expected total requests to be 4, got %d", stats.TotalRequests)
	}

	if stats.SuccessfulRequests != 3 {
		t.Errorf("Expected successful requests to be 3, got %d", stats.SuccessfulRequests)
	}

	if stats.FailedRequests != 1 {
		t.Errorf("Expected failed requests to be 1, got %d", stats.FailedRequests)
	}

	if stats.OverallSuccessRate != 0.75 {
		t.Errorf("Expected success rate to be 0.75, got %f", stats.OverallSuccessRate)
	}

	if len(stats.MostUsedEndpoints) != 3 {
		t.Errorf("Expected 3 most used endpoints, got %d", len(stats.MostUsedEndpoints))
	}
}

func TestTrackRequestErrorTypeStats(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	ctx := context.Background()

	// Track requests with different error types
	requests := []*SuccessRateRequest{
		{Endpoint: "/api/v1/business", UserID: "user1", Success: false, ErrorType: "timeout", ResponseTime: 5000 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/analytics", UserID: "user2", Success: false, ErrorType: "timeout", ResponseTime: 5000 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/business", UserID: "user3", Success: false, ErrorType: "validation", ResponseTime: 200 * time.Millisecond, Timestamp: time.Now()},
		{Endpoint: "/api/v1/reports", UserID: "user4", Success: false, ErrorType: "database", ResponseTime: 1000 * time.Millisecond, Timestamp: time.Now()},
	}

	for _, req := range requests {
		err := tracker.TrackRequest(ctx, req)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	}

	// Get error analysis
	errorAnalysis := tracker.GetErrorAnalysis()

	if len(errorAnalysis) != 3 {
		t.Errorf("Expected 3 error types, got %d", len(errorAnalysis))
	}

	// Check timeout errors
	if timeoutStats, exists := errorAnalysis["timeout"]; exists {
		if timeoutStats.TotalOccurrences != 2 {
			t.Errorf("Expected 2 timeout occurrences, got %d", timeoutStats.TotalOccurrences)
		}
		if len(timeoutStats.AffectedEndpoints) != 2 {
			t.Errorf("Expected 2 affected endpoints for timeout, got %d", len(timeoutStats.AffectedEndpoints))
		}
	} else {
		t.Error("Expected timeout error stats to exist")
	}

	// Check validation errors
	if validationStats, exists := errorAnalysis["validation"]; exists {
		if validationStats.TotalOccurrences != 1 {
			t.Errorf("Expected 1 validation occurrence, got %d", validationStats.TotalOccurrences)
		}
		if len(validationStats.AffectedEndpoints) != 1 {
			t.Errorf("Expected 1 affected endpoint for validation, got %d", len(validationStats.AffectedEndpoints))
		}
	} else {
		t.Error("Expected validation error stats to exist")
	}
}

func TestGetEndpointSuccessRateNotFound(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})

	_, err := tracker.GetEndpointSuccessRate("/nonexistent/endpoint")
	if err == nil {
		t.Error("Expected error for nonexistent endpoint")
	}

	if err.Error() != "endpoint /nonexistent/endpoint not found" {
		t.Errorf("Expected specific error message, got %v", err)
	}
}

func TestGetUserSuccessRateNotFound(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})

	_, err := tracker.GetUserSuccessRate("nonexistent_user")
	if err == nil {
		t.Error("Expected error for nonexistent user")
	}

	if err.Error() != "user nonexistent_user not found" {
		t.Errorf("Expected specific error message, got %v", err)
	}
}

func TestGetTopPerformingEndpoints(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	ctx := context.Background()

	// Create endpoints with different success rates
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
			tracker.TrackRequest(ctx, req)
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
			tracker.TrackRequest(ctx, req)
		}
	}

	// Get top performing endpoints
	topEndpoints := tracker.GetTopPerformingEndpoints(3)

	if len(topEndpoints) != 3 {
		t.Errorf("Expected 3 top endpoints, got %d", len(topEndpoints))
	}

	// Check that they're sorted by success rate (descending)
	if topEndpoints[0].OverallSuccessRate < topEndpoints[1].OverallSuccessRate {
		t.Error("Expected top endpoints to be sorted by success rate (descending)")
	}

	if topEndpoints[1].OverallSuccessRate < topEndpoints[2].OverallSuccessRate {
		t.Error("Expected top endpoints to be sorted by success rate (descending)")
	}

	// Check that the highest success rate endpoint is first
	if topEndpoints[0].Endpoint != "/api/v1/business" {
		t.Errorf("Expected highest success rate endpoint to be first, got %s", topEndpoints[0].Endpoint)
	}
}

func TestGetWorstPerformingEndpoints(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	ctx := context.Background()

	// Create endpoints with different success rates
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
			tracker.TrackRequest(ctx, req)
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
			tracker.TrackRequest(ctx, req)
		}
	}

	// Get worst performing endpoints
	worstEndpoints := tracker.GetWorstPerformingEndpoints(3)

	if len(worstEndpoints) != 3 {
		t.Errorf("Expected 3 worst endpoints, got %d", len(worstEndpoints))
	}

	// Check that they're sorted by success rate (ascending)
	if worstEndpoints[0].OverallSuccessRate > worstEndpoints[1].OverallSuccessRate {
		t.Error("Expected worst endpoints to be sorted by success rate (ascending)")
	}

	if worstEndpoints[1].OverallSuccessRate > worstEndpoints[2].OverallSuccessRate {
		t.Error("Expected worst endpoints to be sorted by success rate (ascending)")
	}

	// Check that the lowest success rate endpoint is first
	if worstEndpoints[0].Endpoint != "/api/v1/users" {
		t.Errorf("Expected lowest success rate endpoint to be first, got %s", worstEndpoints[0].Endpoint)
	}
}

func TestGetSuccessRateTrend(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	ctx := context.Background()

	// Track requests over time
	baseTime := time.Now().Add(-1 * time.Hour)

	for i := 0; i < 10; i++ {
		req := &SuccessRateRequest{
			Endpoint:     "/api/v1/business",
			UserID:       "user1",
			Success:      i < 8, // 80% success rate
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    baseTime.Add(time.Duration(i) * 10 * time.Minute),
		}
		tracker.TrackRequest(ctx, req)
	}

	// Get trend for last 30 minutes
	trend, err := tracker.GetSuccessRateTrend(30 * time.Minute)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(trend) == 0 {
		t.Error("Expected trend data, got empty slice")
	}

	// Check that data points are sorted by timestamp
	for i := 1; i < len(trend); i++ {
		if trend[i-1].Timestamp.After(trend[i].Timestamp) {
			t.Error("Expected trend data points to be sorted by timestamp")
		}
	}
}

func TestCalculateHealthScore(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	ctx := context.Background()

	// Track requests with 90% success rate
	for i := 0; i < 10; i++ {
		req := &SuccessRateRequest{
			Endpoint:     "/api/v1/business",
			UserID:       "user1",
			Success:      i < 9, // 90% success rate
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    time.Now(),
		}
		tracker.TrackRequest(ctx, req)
	}

	overallStats := tracker.GetOverallSuccessRate()

	// Health score should be based on success rate with some penalty for errors
	expectedHealthScore := 90.0 - (1.0/10.0)*20.0 // 90% - (10% error rate * 20 penalty)

	if overallStats.HealthScore != expectedHealthScore {
		t.Errorf("Expected health score to be %f, got %f", expectedHealthScore, overallStats.HealthScore)
	}
}

func TestCalculateDegradationTrend(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	ctx := context.Background()

	// Track requests with improving success rate over time
	baseTime := time.Now().Add(-1 * time.Hour)

	// First 5 requests: 60% success rate
	for i := 0; i < 5; i++ {
		req := &SuccessRateRequest{
			Endpoint:     "/api/v1/business",
			UserID:       "user1",
			Success:      i < 3, // 60% success rate
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    baseTime.Add(time.Duration(i) * 10 * time.Minute),
		}
		tracker.TrackRequest(ctx, req)
	}

	// Next 5 requests: 100% success rate
	for i := 5; i < 10; i++ {
		req := &SuccessRateRequest{
			Endpoint:     "/api/v1/business",
			UserID:       "user1",
			Success:      true, // 100% success rate
			ResponseTime: 100 * time.Millisecond,
			Timestamp:    baseTime.Add(time.Duration(i) * 10 * time.Minute),
		}
		tracker.TrackRequest(ctx, req)
	}

	overallStats := tracker.GetOverallSuccessRate()

	// Overall success rate should be 80% (8/10)
	if overallStats.OverallSuccessRate != 0.8 {
		t.Errorf("Expected overall success rate to be 0.8, got %f", overallStats.OverallSuccessRate)
	}

	// Degradation trend should be calculated (simplified implementation)
	// In a real implementation, this would compare recent vs overall trends
	if overallStats.DegradationTrend == "" {
		t.Error("Expected degradation trend to be calculated")
	}
}

func TestConcurrentAccess(t *testing.T) {
	tracker := NewSuccessRateTracker(SuccessRateTrackerConfig{})
	ctx := context.Background()

	// Test concurrent access to the tracker
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				req := &SuccessRateRequest{
					Endpoint:     "/api/v1/business",
					UserID:       fmt.Sprintf("user%d", id),
					Success:      j%10 != 0, // 90% success rate
					ResponseTime: 100 * time.Millisecond,
					Timestamp:    time.Now(),
				}
				tracker.TrackRequest(ctx, req)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify that all requests were tracked correctly
	overallStats := tracker.GetOverallSuccessRate()
	expectedTotal := int64(10 * 100) // 10 goroutines * 100 requests each

	if overallStats.TotalRequests != expectedTotal {
		t.Errorf("Expected total requests to be %d, got %d", expectedTotal, overallStats.TotalRequests)
	}

	expectedSuccessful := int64(10 * 90) // 10 goroutines * 90 successful requests each
	if overallStats.SuccessfulRequests != expectedSuccessful {
		t.Errorf("Expected successful requests to be %d, got %d", expectedSuccessful, overallStats.SuccessfulRequests)
	}
}
