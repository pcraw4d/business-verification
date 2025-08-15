package observability

import (
	"context"
	"testing"
	"time"
)

func TestNewPerformanceMonitor(t *testing.T) {
	config := PerformanceMonitorConfig{
		MetricsCollectionInterval: 30 * time.Second,
		AlertCheckInterval:        1 * time.Minute,
		OptimizationInterval:      5 * time.Minute,
		PredictionInterval:        2 * time.Minute,
		ResponseTimeThreshold:     500 * time.Millisecond,
		SuccessRateThreshold:      0.95,
		ErrorRateThreshold:        0.05,
		ThroughputThreshold:       1000,
		AutoOptimizationEnabled:   true,
		OptimizationConfidence:    0.8,
		RollbackThreshold:         0.1,
		PredictionHorizon:         1 * time.Hour,
		PredictionConfidence:      0.7,
		TrendAnalysisWindow:       24 * time.Hour,
		DashboardRefreshInterval:  10 * time.Second,
		HistoricalDataRetention:   30 * 24 * time.Hour,
		RealTimeUpdatesEnabled:    true,
		AlertChannels:             []string{"email", "slack"},
		EscalationEnabled:         true,
		EscalationDelay:           15 * time.Minute,
	}

	monitor := NewPerformanceMonitor(config)

	if monitor == nil {
		t.Fatal("Expected monitor to be created, got nil")
	}

	if monitor.config.MetricsCollectionInterval != 30*time.Second {
		t.Errorf("Expected metrics collection interval to be 30s, got %v", monitor.config.MetricsCollectionInterval)
	}

	if monitor.config.AlertCheckInterval != 1*time.Minute {
		t.Errorf("Expected alert check interval to be 1m, got %v", monitor.config.AlertCheckInterval)
	}

	if monitor.config.OptimizationInterval != 5*time.Minute {
		t.Errorf("Expected optimization interval to be 5m, got %v", monitor.config.OptimizationInterval)
	}

	if monitor.config.PredictionInterval != 2*time.Minute {
		t.Errorf("Expected prediction interval to be 2m, got %v", monitor.config.PredictionInterval)
	}

	if monitor.config.ResponseTimeThreshold != 500*time.Millisecond {
		t.Errorf("Expected response time threshold to be 500ms, got %v", monitor.config.ResponseTimeThreshold)
	}

	if monitor.config.SuccessRateThreshold != 0.95 {
		t.Errorf("Expected success rate threshold to be 0.95, got %f", monitor.config.SuccessRateThreshold)
	}

	if monitor.config.ErrorRateThreshold != 0.05 {
		t.Errorf("Expected error rate threshold to be 0.05, got %f", monitor.config.ErrorRateThreshold)
	}

	if monitor.config.ThroughputThreshold != 1000 {
		t.Errorf("Expected throughput threshold to be 1000, got %d", monitor.config.ThroughputThreshold)
	}

	if !monitor.config.AutoOptimizationEnabled {
		t.Error("Expected auto optimization to be enabled")
	}

	if monitor.config.OptimizationConfidence != 0.8 {
		t.Errorf("Expected optimization confidence to be 0.8, got %f", monitor.config.OptimizationConfidence)
	}

	if monitor.config.RollbackThreshold != 0.1 {
		t.Errorf("Expected rollback threshold to be 0.1, got %f", monitor.config.RollbackThreshold)
	}

	if monitor.config.PredictionHorizon != 1*time.Hour {
		t.Errorf("Expected prediction horizon to be 1h, got %v", monitor.config.PredictionHorizon)
	}

	if monitor.config.PredictionConfidence != 0.7 {
		t.Errorf("Expected prediction confidence to be 0.7, got %f", monitor.config.PredictionConfidence)
	}

	if monitor.config.TrendAnalysisWindow != 24*time.Hour {
		t.Errorf("Expected trend analysis window to be 24h, got %v", monitor.config.TrendAnalysisWindow)
	}

	if monitor.config.DashboardRefreshInterval != 10*time.Second {
		t.Errorf("Expected dashboard refresh interval to be 10s, got %v", monitor.config.DashboardRefreshInterval)
	}

	if monitor.config.HistoricalDataRetention != 30*24*time.Hour {
		t.Errorf("Expected historical data retention to be 30d, got %v", monitor.config.HistoricalDataRetention)
	}

	if !monitor.config.RealTimeUpdatesEnabled {
		t.Error("Expected real-time updates to be enabled")
	}

	if len(monitor.config.AlertChannels) != 2 {
		t.Errorf("Expected 2 alert channels, got %d", len(monitor.config.AlertChannels))
	}

	if !monitor.config.EscalationEnabled {
		t.Error("Expected escalation to be enabled")
	}

	if monitor.config.EscalationDelay != 15*time.Minute {
		t.Errorf("Expected escalation delay to be 15m, got %v", monitor.config.EscalationDelay)
	}
}

func TestNewPerformanceMonitorWithDefaults(t *testing.T) {
	config := PerformanceMonitorConfig{}

	monitor := NewPerformanceMonitor(config)

	if monitor == nil {
		t.Fatal("Expected monitor to be created, got nil")
	}

	// Check default values
	if monitor.config.MetricsCollectionInterval != 30*time.Second {
		t.Errorf("Expected default metrics collection interval to be 30s, got %v", monitor.config.MetricsCollectionInterval)
	}

	if monitor.config.AlertCheckInterval != 1*time.Minute {
		t.Errorf("Expected default alert check interval to be 1m, got %v", monitor.config.AlertCheckInterval)
	}

	if monitor.config.OptimizationInterval != 5*time.Minute {
		t.Errorf("Expected default optimization interval to be 5m, got %v", monitor.config.OptimizationInterval)
	}

	if monitor.config.PredictionInterval != 2*time.Minute {
		t.Errorf("Expected default prediction interval to be 2m, got %v", monitor.config.PredictionInterval)
	}

	if monitor.config.ResponseTimeThreshold != 500*time.Millisecond {
		t.Errorf("Expected default response time threshold to be 500ms, got %v", monitor.config.ResponseTimeThreshold)
	}

	if monitor.config.SuccessRateThreshold != 0.95 {
		t.Errorf("Expected default success rate threshold to be 0.95, got %f", monitor.config.SuccessRateThreshold)
	}

	if monitor.config.ErrorRateThreshold != 0.05 {
		t.Errorf("Expected default error rate threshold to be 0.05, got %f", monitor.config.ErrorRateThreshold)
	}

	if monitor.config.ThroughputThreshold != 1000 {
		t.Errorf("Expected default throughput threshold to be 1000, got %d", monitor.config.ThroughputThreshold)
	}

	if monitor.config.OptimizationConfidence != 0.8 {
		t.Errorf("Expected default optimization confidence to be 0.8, got %f", monitor.config.OptimizationConfidence)
	}

	if monitor.config.RollbackThreshold != 0.1 {
		t.Errorf("Expected default rollback threshold to be 0.1, got %f", monitor.config.RollbackThreshold)
	}

	if monitor.config.PredictionHorizon != 1*time.Hour {
		t.Errorf("Expected default prediction horizon to be 1h, got %v", monitor.config.PredictionHorizon)
	}

	if monitor.config.PredictionConfidence != 0.7 {
		t.Errorf("Expected default prediction confidence to be 0.7, got %f", monitor.config.PredictionConfidence)
	}

	if monitor.config.TrendAnalysisWindow != 24*time.Hour {
		t.Errorf("Expected default trend analysis window to be 24h, got %v", monitor.config.TrendAnalysisWindow)
	}

	if monitor.config.DashboardRefreshInterval != 10*time.Second {
		t.Errorf("Expected default dashboard refresh interval to be 10s, got %v", monitor.config.DashboardRefreshInterval)
	}

	if monitor.config.HistoricalDataRetention != 30*24*time.Hour {
		t.Errorf("Expected default historical data retention to be 30d, got %v", monitor.config.HistoricalDataRetention)
	}

	if monitor.config.EscalationDelay != 15*time.Minute {
		t.Errorf("Expected default escalation delay to be 15m, got %v", monitor.config.EscalationDelay)
	}
}

func TestRecordRequest(t *testing.T) {
	config := PerformanceMonitorConfig{}
	monitor := NewPerformanceMonitor(config)

	// Test successful request
	request := &PerformanceRequest{
		ID:           "req_1",
		Endpoint:     "/api/v1/business",
		Method:       "GET",
		ResponseTime: 200 * time.Millisecond,
		Success:      true,
		Timeout:      false,
		DataSize:     1024,
		UserID:       "user_1",
		Timestamp:    time.Now(),
	}

	err := monitor.RecordRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("Expected to record request successfully, got error: %v", err)
	}

	metrics := monitor.GetMetrics()
	if metrics.TotalRequests != 1 {
		t.Errorf("Expected total requests to be 1, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Expected successful requests to be 1, got %d", metrics.SuccessfulRequests)
	}

	if metrics.FailedRequests != 0 {
		t.Errorf("Expected failed requests to be 0, got %d", metrics.FailedRequests)
	}

	if metrics.AverageResponseTime != 200*time.Millisecond {
		t.Errorf("Expected average response time to be 200ms, got %v", metrics.AverageResponseTime)
	}

	if metrics.SuccessRate != 1.0 {
		t.Errorf("Expected success rate to be 1.0, got %f", metrics.SuccessRate)
	}

	if metrics.ErrorRate != 0.0 {
		t.Errorf("Expected error rate to be 0.0, got %f", metrics.ErrorRate)
	}

	// Test failed request
	failedRequest := &PerformanceRequest{
		ID:           "req_2",
		Endpoint:     "/api/v1/business",
		Method:       "POST",
		ResponseTime: 500 * time.Millisecond,
		Success:      false,
		Timeout:      false,
		Error:        "validation error",
		DataSize:     2048,
		UserID:       "user_2",
		Timestamp:    time.Now(),
	}

	err = monitor.RecordRequest(context.Background(), failedRequest)
	if err != nil {
		t.Fatalf("Expected to record failed request successfully, got error: %v", err)
	}

	metrics = monitor.GetMetrics()
	if metrics.TotalRequests != 2 {
		t.Errorf("Expected total requests to be 2, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Expected successful requests to be 1, got %d", metrics.SuccessfulRequests)
	}

	if metrics.FailedRequests != 1 {
		t.Errorf("Expected failed requests to be 1, got %d", metrics.FailedRequests)
	}

	if metrics.SuccessRate != 0.5 {
		t.Errorf("Expected success rate to be 0.5, got %f", metrics.SuccessRate)
	}

	if metrics.ErrorRate != 0.5 {
		t.Errorf("Expected error rate to be 0.5, got %f", metrics.ErrorRate)
	}

	// Test timeout request
	timeoutRequest := &PerformanceRequest{
		ID:           "req_3",
		Endpoint:     "/api/v1/business",
		Method:       "GET",
		ResponseTime: 5 * time.Second,
		Success:      false,
		Timeout:      true,
		Error:        "timeout",
		DataSize:     512,
		UserID:       "user_3",
		Timestamp:    time.Now(),
	}

	err = monitor.RecordRequest(context.Background(), timeoutRequest)
	if err != nil {
		t.Fatalf("Expected to record timeout request successfully, got error: %v", err)
	}

	metrics = monitor.GetMetrics()
	if metrics.TotalRequests != 3 {
		t.Errorf("Expected total requests to be 3, got %d", metrics.TotalRequests)
	}

	if metrics.TimeoutRequests != 1 {
		t.Errorf("Expected timeout requests to be 1, got %d", metrics.TimeoutRequests)
	}

	if metrics.TimeoutRate != 1.0/3.0 {
		t.Errorf("Expected timeout rate to be 0.333..., got %f", metrics.TimeoutRate)
	}
}

func TestGetMetrics(t *testing.T) {
	config := PerformanceMonitorConfig{}
	monitor := NewPerformanceMonitor(config)

	// Record some requests
	requests := []*PerformanceRequest{
		{
			ID:           "req_1",
			Endpoint:     "/api/v1/business",
			Method:       "GET",
			ResponseTime: 100 * time.Millisecond,
			Success:      true,
			DataSize:     1024,
			Timestamp:    time.Now(),
		},
		{
			ID:           "req_2",
			Endpoint:     "/api/v1/business",
			Method:       "POST",
			ResponseTime: 200 * time.Millisecond,
			Success:      true,
			DataSize:     2048,
			Timestamp:    time.Now(),
		},
		{
			ID:           "req_3",
			Endpoint:     "/api/v1/business",
			Method:       "PUT",
			ResponseTime: 300 * time.Millisecond,
			Success:      false,
			Error:        "validation error",
			DataSize:     512,
			Timestamp:    time.Now(),
		},
	}

	for _, request := range requests {
		err := monitor.RecordRequest(context.Background(), request)
		if err != nil {
			t.Fatalf("Expected to record request successfully, got error: %v", err)
		}
	}

	metrics := monitor.GetMetrics()

	if metrics.TotalRequests != 3 {
		t.Errorf("Expected total requests to be 3, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 2 {
		t.Errorf("Expected successful requests to be 2, got %d", metrics.SuccessfulRequests)
	}

	if metrics.FailedRequests != 1 {
		t.Errorf("Expected failed requests to be 1, got %d", metrics.FailedRequests)
	}

	if metrics.SuccessRate != 2.0/3.0 {
		t.Errorf("Expected success rate to be 0.666..., got %f", metrics.SuccessRate)
	}

	if metrics.ErrorRate != 1.0/3.0 {
		t.Errorf("Expected error rate to be 0.333..., got %f", metrics.ErrorRate)
	}

	if metrics.AverageResponseTime != 200*time.Millisecond {
		t.Errorf("Expected average response time to be 200ms, got %v", metrics.AverageResponseTime)
	}

	if metrics.MinResponseTime != 100*time.Millisecond {
		t.Errorf("Expected min response time to be 100ms, got %v", metrics.MinResponseTime)
	}

	if metrics.MaxResponseTime != 300*time.Millisecond {
		t.Errorf("Expected max response time to be 300ms, got %v", metrics.MaxResponseTime)
	}

	if metrics.DataProcessingVolume != 3584 {
		t.Errorf("Expected data processing volume to be 3584, got %d", metrics.DataProcessingVolume)
	}

	if len(metrics.APIUsageByEndpoint) != 1 {
		t.Errorf("Expected 1 endpoint in API usage, got %d", len(metrics.APIUsageByEndpoint))
	}

	if metrics.APIUsageByEndpoint["/api/v1/business"] != 3 {
		t.Errorf("Expected 3 requests for /api/v1/business endpoint, got %d", metrics.APIUsageByEndpoint["/api/v1/business"])
	}
}

func TestGetAlerts(t *testing.T) {
	config := PerformanceMonitorConfig{
		ResponseTimeThreshold: 100 * time.Millisecond,
		SuccessRateThreshold:  0.9,
		ErrorRateThreshold:    0.1,
	}
	monitor := NewPerformanceMonitor(config)

	// Record requests that should trigger alerts
	requests := []*PerformanceRequest{
		{
			ID:           "req_1",
			Endpoint:     "/api/v1/business",
			Method:       "GET",
			ResponseTime: 150 * time.Millisecond, // Above threshold
			Success:      true,
			DataSize:     1024,
			Timestamp:    time.Now(),
		},
		{
			ID:           "req_2",
			Endpoint:     "/api/v1/business",
			Method:       "POST",
			ResponseTime: 200 * time.Millisecond, // Above threshold
			Success:      false,                  // Failed request
			Error:        "validation error",
			DataSize:     2048,
			Timestamp:    time.Now(),
		},
	}

	for _, request := range requests {
		err := monitor.RecordRequest(context.Background(), request)
		if err != nil {
			t.Fatalf("Expected to record request successfully, got error: %v", err)
		}
	}

	// Trigger alert checks
	monitor.checkThresholdAlerts()

	alerts := monitor.GetAlerts()
	if len(alerts) == 0 {
		t.Error("Expected alerts to be generated, got none")
	}

	// Check for specific alerts
	foundResponseTimeAlert := false
	foundErrorRateAlert := false

	for _, alert := range alerts {
		if alert.Metric == "response_time" {
			foundResponseTimeAlert = true
			if alert.Severity != "high" {
				t.Errorf("Expected response time alert severity to be 'high', got '%s'", alert.Severity)
			}
		}
		if alert.Metric == "error_rate" {
			foundErrorRateAlert = true
			if alert.Severity != "critical" {
				t.Errorf("Expected error rate alert severity to be 'critical', got '%s'", alert.Severity)
			}
		}
	}

	if !foundResponseTimeAlert {
		t.Error("Expected response time alert to be generated")
	}

	if !foundErrorRateAlert {
		t.Error("Expected error rate alert to be generated")
	}
}

func TestGetOptimizations(t *testing.T) {
	config := PerformanceMonitorConfig{
		AutoOptimizationEnabled: true,
		ResponseTimeThreshold:   100 * time.Millisecond,
		SuccessRateThreshold:    0.9,
	}
	monitor := NewPerformanceMonitor(config)

	// Record requests that should trigger optimizations
	requests := []*PerformanceRequest{
		{
			ID:           "req_1",
			Endpoint:     "/api/v1/business",
			Method:       "GET",
			ResponseTime: 150 * time.Millisecond, // Above threshold
			Success:      true,
			DataSize:     1024,
			Timestamp:    time.Now(),
		},
		{
			ID:           "req_2",
			Endpoint:     "/api/v1/business",
			Method:       "POST",
			ResponseTime: 200 * time.Millisecond, // Above threshold
			Success:      false,                  // Failed request
			Error:        "validation error",
			DataSize:     2048,
			Timestamp:    time.Now(),
		},
	}

	for _, request := range requests {
		err := monitor.RecordRequest(context.Background(), request)
		if err != nil {
			t.Fatalf("Expected to record request successfully, got error: %v", err)
		}
	}

	// Trigger optimization
	monitor.runAutoOptimizations()

	optimizations := monitor.GetOptimizations()
	if len(optimizations) == 0 {
		t.Error("Expected optimizations to be generated, got none")
	}

	// Check for specific optimizations
	foundCachingOptimization := false
	foundRetryOptimization := false

	for _, optimization := range optimizations {
		if optimization.Action == "enable_caching" {
			foundCachingOptimization = true
			if optimization.TargetMetric != "response_time" {
				t.Errorf("Expected caching optimization target metric to be 'response_time', got '%s'", optimization.TargetMetric)
			}
			if optimization.Status != "applied" {
				t.Errorf("Expected caching optimization status to be 'applied', got '%s'", optimization.Status)
			}
		}
		if optimization.Action == "retry_configuration" {
			foundRetryOptimization = true
			if optimization.TargetMetric != "success_rate" {
				t.Errorf("Expected retry optimization target metric to be 'success_rate', got '%s'", optimization.TargetMetric)
			}
			if optimization.Status != "applied" {
				t.Errorf("Expected retry optimization status to be 'applied', got '%s'", optimization.Status)
			}
		}
	}

	if !foundCachingOptimization {
		t.Error("Expected caching optimization to be generated")
	}

	if !foundRetryOptimization {
		t.Error("Expected retry optimization to be generated")
	}
}

func TestGetPredictions(t *testing.T) {
	config := PerformanceMonitorConfig{
		PredictionHorizon:    1 * time.Hour,
		PredictionConfidence: 0.7,
	}
	monitor := NewPerformanceMonitor(config)

	// Record some requests
	requests := []*PerformanceRequest{
		{
			ID:           "req_1",
			Endpoint:     "/api/v1/business",
			Method:       "GET",
			ResponseTime: 100 * time.Millisecond,
			Success:      true,
			DataSize:     1024,
			Timestamp:    time.Now(),
		},
		{
			ID:           "req_2",
			Endpoint:     "/api/v1/business",
			Method:       "POST",
			ResponseTime: 200 * time.Millisecond,
			Success:      false,
			Error:        "validation error",
			DataSize:     2048,
			Timestamp:    time.Now(),
		},
	}

	for _, request := range requests {
		err := monitor.RecordRequest(context.Background(), request)
		if err != nil {
			t.Fatalf("Expected to record request successfully, got error: %v", err)
		}
	}

	// Trigger predictions
	monitor.runPerformancePredictions()

	predictions := monitor.GetPredictions()
	if len(predictions) == 0 {
		t.Error("Expected predictions to be generated, got none")
	}

	// Check for specific predictions
	foundResponseTimePrediction := false
	foundSuccessRatePrediction := false

	for _, prediction := range predictions {
		if prediction.Metric == "response_time" {
			foundResponseTimePrediction = true
			if prediction.Confidence != 0.7 {
				t.Errorf("Expected response time prediction confidence to be 0.7, got %f", prediction.Confidence)
			}
			if prediction.PredictionHorizon != 1*time.Hour {
				t.Errorf("Expected response time prediction horizon to be 1h, got %v", prediction.PredictionHorizon)
			}
			if prediction.Trend != "degrading" {
				t.Errorf("Expected response time prediction trend to be 'degrading', got '%s'", prediction.Trend)
			}
		}
		if prediction.Metric == "success_rate" {
			foundSuccessRatePrediction = true
			if prediction.Confidence != 0.7 {
				t.Errorf("Expected success rate prediction confidence to be 0.7, got %f", prediction.Confidence)
			}
			if prediction.PredictionHorizon != 1*time.Hour {
				t.Errorf("Expected success rate prediction horizon to be 1h, got %v", prediction.PredictionHorizon)
			}
			if prediction.Trend != "degrading" {
				t.Errorf("Expected success rate prediction trend to be 'degrading', got '%s'", prediction.Trend)
			}
		}
	}

	if !foundResponseTimePrediction {
		t.Error("Expected response time prediction to be generated")
	}

	if !foundSuccessRatePrediction {
		t.Error("Expected success rate prediction to be generated")
	}
}

func TestGetDashboard(t *testing.T) {
	config := PerformanceMonitorConfig{}
	monitor := NewPerformanceMonitor(config)

	// Record some requests
	requests := []*PerformanceRequest{
		{
			ID:           "req_1",
			Endpoint:     "/api/v1/business",
			Method:       "GET",
			ResponseTime: 100 * time.Millisecond,
			Success:      true,
			DataSize:     1024,
			Timestamp:    time.Now(),
		},
		{
			ID:           "req_2",
			Endpoint:     "/api/v1/business",
			Method:       "POST",
			ResponseTime: 200 * time.Millisecond,
			Success:      false,
			Error:        "validation error",
			DataSize:     2048,
			Timestamp:    time.Now(),
		},
	}

	for _, request := range requests {
		err := monitor.RecordRequest(context.Background(), request)
		if err != nil {
			t.Fatalf("Expected to record request successfully, got error: %v", err)
		}
	}

	dashboard := monitor.GetDashboard()

	if dashboard == nil {
		t.Fatal("Expected dashboard to be returned, got nil")
	}

	if dashboard.CurrentMetrics == nil {
		t.Fatal("Expected current metrics to be present in dashboard")
	}

	if dashboard.CurrentMetrics.TotalRequests != 2 {
		t.Errorf("Expected dashboard to show 2 total requests, got %d", dashboard.CurrentMetrics.TotalRequests)
	}

	if dashboard.OverallHealth == "" {
		t.Error("Expected overall health to be calculated")
	}

	if dashboard.LastUpdated.IsZero() {
		t.Error("Expected last updated timestamp to be set")
	}
}

func TestCalculateOverallHealth(t *testing.T) {
	config := PerformanceMonitorConfig{
		ResponseTimeThreshold: 100 * time.Millisecond,
		SuccessRateThreshold:  0.9,
		ErrorRateThreshold:    0.1,
	}
	monitor := NewPerformanceMonitor(config)

	// Test excellent health
	requests := []*PerformanceRequest{
		{
			ID:           "req_1",
			Endpoint:     "/api/v1/business",
			Method:       "GET",
			ResponseTime: 50 * time.Millisecond,
			Success:      true,
			DataSize:     1024,
			Timestamp:    time.Now(),
		},
		{
			ID:           "req_2",
			Endpoint:     "/api/v1/business",
			Method:       "POST",
			ResponseTime: 75 * time.Millisecond,
			Success:      true,
			DataSize:     2048,
			Timestamp:    time.Now(),
		},
	}

	for _, request := range requests {
		err := monitor.RecordRequest(context.Background(), request)
		if err != nil {
			t.Fatalf("Expected to record request successfully, got error: %v", err)
		}
	}

	health := monitor.calculateOverallHealth()
	if health != "excellent" {
		t.Errorf("Expected health to be 'excellent', got '%s'", health)
	}

	// Test poor health
	monitor = NewPerformanceMonitor(config)
	poorRequests := []*PerformanceRequest{
		{
			ID:           "req_1",
			Endpoint:     "/api/v1/business",
			Method:       "GET",
			ResponseTime: 200 * time.Millisecond, // Above threshold
			Success:      false,                  // Failed request
			Error:        "validation error",
			DataSize:     1024,
			Timestamp:    time.Now(),
		},
		{
			ID:           "req_2",
			Endpoint:     "/api/v1/business",
			Method:       "POST",
			ResponseTime: 300 * time.Millisecond, // Above threshold
			Success:      false,                  // Failed request
			Error:        "timeout",
			DataSize:     2048,
			Timestamp:    time.Now(),
		},
	}

	for _, request := range poorRequests {
		err := monitor.RecordRequest(context.Background(), request)
		if err != nil {
			t.Fatalf("Expected to record request successfully, got error: %v", err)
		}
	}

	health = monitor.calculateOverallHealth()
	if health != "poor" {
		t.Errorf("Expected health to be 'poor', got '%s'", health)
	}
}

func TestStartMonitoring(t *testing.T) {
	config := PerformanceMonitorConfig{
		MetricsCollectionInterval: 100 * time.Millisecond,
		AlertCheckInterval:        200 * time.Millisecond,
		OptimizationInterval:      300 * time.Millisecond,
		PredictionInterval:        400 * time.Millisecond,
		DashboardRefreshInterval:  500 * time.Millisecond,
	}
	monitor := NewPerformanceMonitor(config)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := monitor.StartMonitoring(ctx)
	if err != nil {
		t.Fatalf("Expected to start monitoring successfully, got error: %v", err)
	}

	// Wait for monitoring to start
	time.Sleep(100 * time.Millisecond)

	// The monitoring should be running in the background
	// We can't easily test the background goroutines without more complex setup
	// This test mainly ensures the StartMonitoring function doesn't return an error
}

func TestPerformanceMetricsClone(t *testing.T) {
	metrics := NewPerformanceMetrics()
	metrics.TotalRequests = 10
	metrics.SuccessfulRequests = 8
	metrics.FailedRequests = 2
	metrics.AverageResponseTime = 150 * time.Millisecond
	metrics.SuccessRate = 0.8
	metrics.ErrorRate = 0.2
	metrics.APIUsageByEndpoint = map[string]int64{
		"/api/v1/business": 5,
		"/api/v1/user":     5,
	}

	clone := metrics.Clone()

	if clone.TotalRequests != metrics.TotalRequests {
		t.Errorf("Expected cloned total requests to match original, got %d vs %d", clone.TotalRequests, metrics.TotalRequests)
	}

	if clone.SuccessfulRequests != metrics.SuccessfulRequests {
		t.Errorf("Expected cloned successful requests to match original, got %d vs %d", clone.SuccessfulRequests, metrics.SuccessfulRequests)
	}

	if clone.FailedRequests != metrics.FailedRequests {
		t.Errorf("Expected cloned failed requests to match original, got %d vs %d", clone.FailedRequests, metrics.FailedRequests)
	}

	if clone.AverageResponseTime != metrics.AverageResponseTime {
		t.Errorf("Expected cloned average response time to match original, got %v vs %v", clone.AverageResponseTime, metrics.AverageResponseTime)
	}

	if clone.SuccessRate != metrics.SuccessRate {
		t.Errorf("Expected cloned success rate to match original, got %f vs %f", clone.SuccessRate, metrics.SuccessRate)
	}

	if clone.ErrorRate != metrics.ErrorRate {
		t.Errorf("Expected cloned error rate to match original, got %f vs %f", clone.ErrorRate, metrics.ErrorRate)
	}

	if len(clone.APIUsageByEndpoint) != len(metrics.APIUsageByEndpoint) {
		t.Errorf("Expected cloned API usage to have same length, got %d vs %d", len(clone.APIUsageByEndpoint), len(metrics.APIUsageByEndpoint))
	}

	for endpoint, count := range metrics.APIUsageByEndpoint {
		if clone.APIUsageByEndpoint[endpoint] != count {
			t.Errorf("Expected cloned API usage for %s to match original, got %d vs %d", endpoint, clone.APIUsageByEndpoint[endpoint], count)
		}
	}

	// Test that modifying clone doesn't affect original
	clone.TotalRequests = 999
	if metrics.TotalRequests == 999 {
		t.Error("Expected modifying clone to not affect original")
	}

	clone.APIUsageByEndpoint["/api/v1/test"] = 100
	if _, exists := metrics.APIUsageByEndpoint["/api/v1/test"]; exists {
		t.Error("Expected modifying clone API usage to not affect original")
	}
}

func TestPerformanceAlertManager(t *testing.T) {
	config := PerformanceMonitorConfig{}
	alertManager := NewPerformanceAlertManager(config)

	// Create test alert
	alert := &PerformanceAlert{
		Type:         "threshold",
		Severity:     "high",
		Title:        "Test Alert",
		Description:  "Test alert description",
		Metric:       "response_time",
		CurrentValue: 150.0,
		Threshold:    100.0,
		Timestamp:    time.Now(),
		Recommendations: []string{
			"Check database performance",
			"Review query optimization",
		},
	}

	alertManager.CreateAlert(alert)

	alerts := alertManager.GetActiveAlerts()
	if len(alerts) != 1 {
		t.Errorf("Expected 1 active alert, got %d", len(alerts))
	}

	retrievedAlert := alerts[0]
	if retrievedAlert.ID == "" {
		t.Error("Expected alert to have an ID")
	}

	if retrievedAlert.Status != "active" {
		t.Errorf("Expected alert status to be 'active', got '%s'", retrievedAlert.Status)
	}

	if retrievedAlert.Type != "threshold" {
		t.Errorf("Expected alert type to be 'threshold', got '%s'", retrievedAlert.Type)
	}

	if retrievedAlert.Severity != "high" {
		t.Errorf("Expected alert severity to be 'high', got '%s'", retrievedAlert.Severity)
	}

	if retrievedAlert.Title != "Test Alert" {
		t.Errorf("Expected alert title to be 'Test Alert', got '%s'", retrievedAlert.Title)
	}

	if len(retrievedAlert.Recommendations) != 2 {
		t.Errorf("Expected 2 recommendations, got %d", len(retrievedAlert.Recommendations))
	}
}

func TestPerformanceOptimizer(t *testing.T) {
	config := PerformanceMonitorConfig{}
	optimizer := NewPerformanceOptimizer(config)

	// Create test optimization
	optimization := &PerformanceOptimization{
		Type:                "auto",
		Action:              "enable_caching",
		Description:         "Enable response caching to improve performance",
		TargetMetric:        "response_time",
		ExpectedImprovement: 0.3,
		Confidence:          0.8,
		RiskLevel:           "low",
		Status:              "pending",
	}

	optimizer.ApplyOptimization(optimization)

	optimizations := optimizer.GetRecentOptimizations()
	if len(optimizations) != 1 {
		t.Errorf("Expected 1 recent optimization, got %d", len(optimizations))
	}

	retrievedOptimization := optimizations[0]
	if retrievedOptimization.ID == "" {
		t.Error("Expected optimization to have an ID")
	}

	if retrievedOptimization.Status != "applied" {
		t.Errorf("Expected optimization status to be 'applied', got '%s'", retrievedOptimization.Status)
	}

	if retrievedOptimization.Type != "auto" {
		t.Errorf("Expected optimization type to be 'auto', got '%s'", retrievedOptimization.Type)
	}

	if retrievedOptimization.Action != "enable_caching" {
		t.Errorf("Expected optimization action to be 'enable_caching', got '%s'", retrievedOptimization.Action)
	}

	if retrievedOptimization.TargetMetric != "response_time" {
		t.Errorf("Expected optimization target metric to be 'response_time', got '%s'", retrievedOptimization.TargetMetric)
	}

	if retrievedOptimization.ExpectedImprovement != 0.3 {
		t.Errorf("Expected optimization expected improvement to be 0.3, got %f", retrievedOptimization.ExpectedImprovement)
	}

	if retrievedOptimization.Confidence != 0.8 {
		t.Errorf("Expected optimization confidence to be 0.8, got %f", retrievedOptimization.Confidence)
	}

	if retrievedOptimization.RiskLevel != "low" {
		t.Errorf("Expected optimization risk level to be 'low', got '%s'", retrievedOptimization.RiskLevel)
	}

	if retrievedOptimization.AppliedAt == nil {
		t.Error("Expected optimization to have applied timestamp")
	}
}

func TestPerformancePredictor(t *testing.T) {
	config := PerformanceMonitorConfig{
		PredictionHorizon: 1 * time.Hour,
	}
	predictor := NewPerformancePredictor(config)

	// Create test prediction
	prediction := &PerformancePrediction{
		Metric:            "response_time",
		PredictedValue:    150.0,
		Confidence:        0.7,
		PredictionHorizon: 1 * time.Hour,
		Trend:             "degrading",
		Factors:           []string{"increasing_load", "resource_constraints"},
		Timestamp:         time.Now(),
	}

	predictor.AddPrediction(prediction)

	predictions := predictor.GetCurrentPredictions()
	if len(predictions) != 1 {
		t.Errorf("Expected 1 current prediction, got %d", len(predictions))
	}

	retrievedPrediction := predictions[0]
	if retrievedPrediction.ID == "" {
		t.Error("Expected prediction to have an ID")
	}

	if retrievedPrediction.Metric != "response_time" {
		t.Errorf("Expected prediction metric to be 'response_time', got '%s'", retrievedPrediction.Metric)
	}

	if retrievedPrediction.PredictedValue != 150.0 {
		t.Errorf("Expected prediction value to be 150.0, got %f", retrievedPrediction.PredictedValue)
	}

	if retrievedPrediction.Confidence != 0.7 {
		t.Errorf("Expected prediction confidence to be 0.7, got %f", retrievedPrediction.Confidence)
	}

	if retrievedPrediction.PredictionHorizon != 1*time.Hour {
		t.Errorf("Expected prediction horizon to be 1h, got %v", retrievedPrediction.PredictionHorizon)
	}

	if retrievedPrediction.Trend != "degrading" {
		t.Errorf("Expected prediction trend to be 'degrading', got '%s'", retrievedPrediction.Trend)
	}

	if len(retrievedPrediction.Factors) != 2 {
		t.Errorf("Expected 2 factors, got %d", len(retrievedPrediction.Factors))
	}
}

func TestPerformanceDashboard(t *testing.T) {
	config := PerformanceMonitorConfig{}
	dashboard := NewPerformanceDashboard(config)

	// Create test dashboard data
	dashboardData := &PerformanceDashboard{
		CurrentMetrics:      NewPerformanceMetrics(),
		ActiveAlerts:        []*PerformanceAlert{},
		RecentOptimizations: []*PerformanceOptimization{},
		CurrentPredictions:  []*PerformancePrediction{},
		OverallHealth:       "good",
		LastUpdated:         time.Now(),
	}

	dashboard.UpdateDashboard(dashboardData.CurrentMetrics, dashboardData.ActiveAlerts, dashboardData.RecentOptimizations, dashboardData.CurrentPredictions)

	// Note: The dashboard doesn't have a getter method in the current implementation
	// This test mainly ensures the UpdateDashboard function doesn't panic
}
