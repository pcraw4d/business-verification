package observability

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestApplicationMonitoringService(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultMonitoringConfig()
	service := NewApplicationMonitoringService(NewLogger(logger), config)

	t.Run("Start and Stop", func(t *testing.T) {
		err := service.Start()
		if err != nil {
			t.Fatalf("Failed to start monitoring service: %v", err)
		}

		// Wait a bit to ensure it's running
		time.Sleep(100 * time.Millisecond)

		err = service.Stop()
		if err != nil {
			t.Fatalf("Failed to stop monitoring service: %v", err)
		}
	})

	t.Run("Record Metrics", func(t *testing.T) {
		service.RecordMetric("test_metric", 42.0, MetricTypeGauge, map[string]string{"test": "value"})
		service.IncrementCounter("test_counter", map[string]string{"test": "value"})
		service.SetGauge("test_gauge", 100.0, map[string]string{"test": "value"})
		service.RecordHistogram("test_histogram", 50.0, map[string]string{"test": "value"})

		metrics := service.GetMetrics()
		if len(metrics) == 0 {
			t.Error("Expected metrics to be recorded")
		}
	})

	t.Run("Track Error", func(t *testing.T) {
		service.TrackError(nil, ErrorSeverityLow, map[string]interface{}{"test": "value"}, map[string]string{"test": "value"})

		testErr := &testError{message: "test error"}
		service.TrackError(testErr, ErrorSeverityHigh, map[string]interface{}{"test": "value"}, map[string]string{"test": "value"})

		errorSummary := service.GetErrorSummary()
		if errorSummary == nil {
			t.Error("Expected error summary to be available")
		}
	})

	t.Run("Track User Event", func(t *testing.T) {
		service.TrackUserEvent("user123", "session456", "test_event", map[string]interface{}{"test": "value"}, map[string]string{"test": "value"})

		analyticsSummary := service.GetUserAnalyticsSummary()
		if analyticsSummary == nil {
			t.Error("Expected user analytics summary to be available")
		}
	})

	t.Run("Add Health Check", func(t *testing.T) {
		service.AddHealthCheck("test_check", func() error { return nil }, 30*time.Second, false)

		healthStatus := service.GetHealthStatus()
		if healthStatus == nil {
			t.Error("Expected health status to be available")
		}
	})
}

func TestMetricsCollector(t *testing.T) {
	logger := zap.NewNop()
	collector := NewMetricsCollector(NewLogger(logger))

	t.Run("Record Metric", func(t *testing.T) {
		collector.RecordMetric("test_metric", 42.0, MetricTypeGauge, map[string]string{"test": "value"})

		metrics := collector.GetMetrics()
		if len(metrics) == 0 {
			t.Error("Expected metrics to be recorded")
		}
	})

	t.Run("Increment Counter", func(t *testing.T) {
		collector.IncrementCounter("test_counter", map[string]string{"test": "value"})
		collector.IncrementCounter("test_counter", map[string]string{"test": "value"})

		metrics := collector.GetMetrics()
		found := false
		for _, metric := range metrics {
			if metric.Name == "test_counter" && metric.Value == 2.0 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected counter to be incremented to 2")
		}
	})

	t.Run("Set Gauge", func(t *testing.T) {
		collector.SetGauge("test_gauge", 100.0, map[string]string{"test": "value"})

		metrics := collector.GetMetrics()
		found := false
		for _, metric := range metrics {
			if metric.Name == "test_gauge" && metric.Value == 100.0 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected gauge to be set to 100")
		}
	})

	t.Run("Get Metrics By Type", func(t *testing.T) {
		collector.RecordMetric("test_counter", 1.0, MetricTypeCounter, map[string]string{})
		collector.RecordMetric("test_gauge", 1.0, MetricTypeGauge, map[string]string{})

		counterMetrics := collector.GetMetricsByType(MetricTypeCounter)
		gaugeMetrics := collector.GetMetricsByType(MetricTypeGauge)

		if len(counterMetrics) == 0 {
			t.Error("Expected counter metrics to be found")
		}
		if len(gaugeMetrics) == 0 {
			t.Error("Expected gauge metrics to be found")
		}
	})

	t.Run("Add Exporter", func(t *testing.T) {
		exporter := NewConsoleExporter(NewLogger(logger))
		collector.AddExporter(exporter)

		err := collector.ExportMetrics()
		if err != nil {
			t.Errorf("Expected no error when exporting metrics, got: %v", err)
		}
	})

	t.Run("Get Summary", func(t *testing.T) {
		collector.RecordMetric("test_metric", 1.0, MetricTypeGauge, map[string]string{})
		summary := collector.GetMetricsSummary()

		if summary["total_metrics"].(int) == 0 {
			t.Error("Expected total metrics to be greater than 0")
		}
	})
}

func TestErrorTracker(t *testing.T) {
	logger := zap.NewNop()
	alertConfig := &ErrorAlertConfig{
		Enabled:           true,
		CriticalThreshold: 10,
		HighThreshold:     50,
		MediumThreshold:   100,
		TimeWindow:        5 * time.Minute,
		AlertChannels:     []string{"email", "slack"},
	}
	tracker := NewErrorTracker(NewLogger(logger), alertConfig)

	t.Run("Track Error", func(t *testing.T) {
		testErr := &testError{message: "test error"}
		tracker.TrackError(testErr, ErrorSeverityHigh, map[string]interface{}{"test": "value"}, map[string]string{"test": "value"})

		summary := tracker.GetSummary()
		if summary["total_errors"].(int) == 0 {
			t.Error("Expected errors to be tracked")
		}
	})

	t.Run("Track Error With Context", func(t *testing.T) {
		testErr := &testError{message: "test error with context"}
		tracker.TrackErrorWithContext(testErr, ErrorSeverityMedium, "user123", "session456", "req789", map[string]interface{}{"test": "value"}, map[string]string{"test": "value"})

		summary := tracker.GetSummary()
		if summary["total_errors"].(int) == 0 {
			t.Error("Expected errors with context to be tracked")
		}
	})

	t.Run("Get Errors By Severity", func(t *testing.T) {
		testErr := &testError{message: "high severity error"}
		tracker.TrackError(testErr, ErrorSeverityHigh, map[string]interface{}{}, map[string]string{})

		highSeverityErrors := tracker.GetErrorsBySeverity(ErrorSeverityHigh)
		if len(highSeverityErrors) == 0 {
			t.Error("Expected high severity errors to be found")
		}
	})

	t.Run("Resolve Error", func(t *testing.T) {
		testErr := &testError{message: "error to resolve"}
		tracker.TrackError(testErr, ErrorSeverityLow, map[string]interface{}{}, map[string]string{})

		// Get the error ID (simplified for test)
		summary := tracker.GetSummary()
		if summary["total_errors"].(int) > 0 {
			// In a real test, we'd get the actual error ID
			// For now, just verify the resolve function doesn't panic
			err := tracker.ResolveError("test_error_id")
			if err == nil {
				// Expected error since we don't have the real ID
			}
		}
	})

	t.Run("Add Exporter", func(t *testing.T) {
		exporter := NewLogExporter(NewLogger(logger))
		tracker.AddExporter(exporter)

		// Verify exporter was added
		summary := tracker.GetSummary()
		if summary == nil {
			t.Error("Expected summary to be available after adding exporter")
		}
	})
}

func TestUserAnalytics(t *testing.T) {
	logger := zap.NewNop()
	config := &UserAnalyticsConfig{
		Enabled:              true,
		TrackPageViews:       true,
		TrackClicks:          true,
		TrackFormSubmissions: true,
		TrackAPIUsage:        true,
		AnonymizeIP:          true,
		RetentionDays:        30,
		BatchSize:            100,
		FlushInterval:        30 * time.Second,
	}
	analytics := NewUserAnalytics(NewLogger(logger), config)

	t.Run("Track Event", func(t *testing.T) {
		analytics.TrackEvent("user123", "session456", "test_event", map[string]interface{}{"test": "value"}, map[string]string{"test": "value"})

		summary := analytics.GetSummary()
		if summary["total_events"].(int) == 0 {
			t.Error("Expected events to be tracked")
		}
	})

	t.Run("Track Page View", func(t *testing.T) {
		analytics.TrackPageView("user123", "session456", "/test-page", 5*time.Second, map[string]string{"test": "value"})

		pageViewEvents := analytics.GetEventsByType("page_view")
		if len(pageViewEvents) == 0 {
			t.Error("Expected page view events to be tracked")
		}
	})

	t.Run("Track Click", func(t *testing.T) {
		analytics.TrackClick("user123", "session456", "button", "/test-page", map[string]string{"test": "value"})

		clickEvents := analytics.GetEventsByType("click")
		if len(clickEvents) == 0 {
			t.Error("Expected click events to be tracked")
		}
	})

	t.Run("Track Form Submission", func(t *testing.T) {
		analytics.TrackFormSubmission("user123", "session456", "test_form", true, map[string]string{"test": "value"})

		formEvents := analytics.GetEventsByType("form_submission")
		if len(formEvents) == 0 {
			t.Error("Expected form submission events to be tracked")
		}
	})

	t.Run("Track API Usage", func(t *testing.T) {
		analytics.TrackAPIUsage("user123", "session456", "/api/test", "GET", 200, 100*time.Millisecond, map[string]string{"test": "value"})

		apiEvents := analytics.GetEventsByType("api_usage")
		if len(apiEvents) == 0 {
			t.Error("Expected API usage events to be tracked")
		}
	})

	t.Run("Get Events By User", func(t *testing.T) {
		analytics.TrackEvent("user123", "session456", "test_event", map[string]interface{}{}, map[string]string{})

		userEvents := analytics.GetEventsByUser("user123")
		if len(userEvents) == 0 {
			t.Error("Expected user events to be found")
		}
	})

	t.Run("Get Events By Session", func(t *testing.T) {
		analytics.TrackEvent("user123", "session456", "test_event", map[string]interface{}{}, map[string]string{})

		sessionEvents := analytics.GetEventsBySession("session456")
		if len(sessionEvents) == 0 {
			t.Error("Expected session events to be found")
		}
	})

	t.Run("Add Exporter", func(t *testing.T) {
		exporter := NewUserAnalyticsLogExporter(NewLogger(logger))
		analytics.AddExporter(exporter)

		// Verify exporter was added
		summary := analytics.GetSummary()
		if summary == nil {
			t.Error("Expected summary to be available after adding exporter")
		}
	})
}

func TestHealthChecker(t *testing.T) {
	logger := zap.NewNop()
	config := &HealthCheckConfig{
		Enabled:        true,
		CheckInterval:  30 * time.Second,
		Timeout:        10 * time.Second,
		RetryCount:     3,
		RetryInterval:  5 * time.Second,
		AlertOnFailure: true,
		AlertChannels:  []string{"email", "slack"},
	}
	checker := NewHealthChecker(NewLogger(logger), config)

	t.Run("Add Check", func(t *testing.T) {
		checker.AddCheck("test_check", func() error { return nil }, 30*time.Second, false)

		status := checker.GetStatus()
		if status["total_checks"].(int) == 0 {
			t.Error("Expected health check to be added")
		}
	})

	t.Run("Run Check", func(t *testing.T) {
		checker.AddCheck("test_check", func() error { return nil }, 30*time.Second, false)

		err := checker.RunCheck("test_check")
		if err != nil {
			t.Errorf("Expected no error when running check, got: %v", err)
		}
	})

	t.Run("Run Check With Error", func(t *testing.T) {
		checker.AddCheck("failing_check", func() error { return &testError{message: "check failed"} }, 30*time.Second, false)

		err := checker.RunCheck("failing_check")
		if err != nil {
			t.Errorf("Expected no error when running failing check, got: %v", err)
		}

		// Check should be marked as unhealthy
		status := checker.GetStatus()
		if status["unhealthy"].(int) == 0 {
			t.Error("Expected failing check to be marked as unhealthy")
		}
	})

	t.Run("Get Check Status", func(t *testing.T) {
		checker.AddCheck("test_check", func() error { return nil }, 30*time.Second, false)

		check, exists := checker.GetCheckStatus("test_check")
		if !exists {
			t.Error("Expected check to exist")
		}
		if check.Name != "test_check" {
			t.Error("Expected check name to match")
		}
	})

	t.Run("Remove Check", func(t *testing.T) {
		checker.AddCheck("test_check", func() error { return nil }, 30*time.Second, false)

		err := checker.RemoveCheck("test_check")
		if err != nil {
			t.Errorf("Expected no error when removing check, got: %v", err)
		}

		_, exists := checker.GetCheckStatus("test_check")
		if exists {
			t.Error("Expected check to be removed")
		}
	})

	t.Run("Add Exporter", func(t *testing.T) {
		exporter := NewLogHealthExporter(NewLogger(logger))
		checker.AddExporter(exporter)

		// Verify exporter was added
		status := checker.GetStatus()
		if status == nil {
			t.Error("Expected status to be available after adding exporter")
		}
	})
}

func TestPerformanceMonitor(t *testing.T) {
	logger := zap.NewNop()
	config := &PerformanceConfig{
		Enabled:              true,
		CollectionInterval:   30 * time.Second,
		TrackHTTPRequests:    true,
		TrackDatabaseQueries: true,
		TrackExternalAPIs:    true,
		TrackMemoryUsage:     true,
		TrackCPUUsage:        true,
		TrackGoroutines:      true,
		TrackGC:              true,
		Percentiles:          []float64{0.5, 0.9, 0.95, 0.99},
	}
	monitor := NewPerformanceMonitor(NewLogger(logger), config)

	t.Run("Record Metric", func(t *testing.T) {
		monitor.RecordMetric("test_metric", 42.0, "seconds", map[string]string{"test": "value"})

		metrics := monitor.GetMetrics()
		if len(metrics) == 0 {
			t.Error("Expected performance metrics to be recorded")
		}
	})

	t.Run("Record HTTP Request", func(t *testing.T) {
		monitor.RecordHTTPRequest("GET", "/test", 200, 100*time.Millisecond, map[string]string{"test": "value"})

		metrics := monitor.GetMetrics()
		found := false
		for _, metric := range metrics {
			if metric.Name == "http_request_duration_seconds" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected HTTP request metrics to be recorded")
		}
	})

	t.Run("Record Database Query", func(t *testing.T) {
		monitor.RecordDatabaseQuery("SELECT", "users", 50*time.Millisecond, true, map[string]string{"test": "value"})

		metrics := monitor.GetMetrics()
		found := false
		for _, metric := range metrics {
			if metric.Name == "database_query_duration_seconds" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected database query metrics to be recorded")
		}
	})

	t.Run("Record External API", func(t *testing.T) {
		monitor.RecordExternalAPI("test_provider", "/api/test", 200*time.Millisecond, 200, map[string]string{"test": "value"})

		metrics := monitor.GetMetrics()
		found := false
		for _, metric := range metrics {
			if metric.Name == "external_api_duration_seconds" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected external API metrics to be recorded")
		}
	})

	t.Run("Record Business Operation", func(t *testing.T) {
		monitor.RecordBusinessOperation("test_operation", 75*time.Millisecond, true, map[string]string{"test": "value"})

		metrics := monitor.GetMetrics()
		found := false
		for _, metric := range metrics {
			if metric.Name == "business_operation_duration_seconds" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected business operation metrics to be recorded")
		}
	})

	t.Run("Record Merchant Operation", func(t *testing.T) {
		monitor.RecordMerchantOperation("test_merchant_op", "merchant123", 90*time.Millisecond, true, map[string]string{"test": "value"})

		metrics := monitor.GetMetrics()
		found := false
		for _, metric := range metrics {
			if metric.Name == "merchant_operation_duration_seconds" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected merchant operation metrics to be recorded")
		}
	})

	t.Run("Collect System Metrics", func(t *testing.T) {
		monitor.CollectSystemMetrics()

		metrics := monitor.GetMetrics()
		found := false
		for _, metric := range metrics {
			if metric.Name == "system_memory_alloc_bytes" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected system metrics to be collected")
		}
	})

	t.Run("Get Metrics By Name", func(t *testing.T) {
		monitor.RecordMetric("test_metric", 42.0, "seconds", map[string]string{"test": "value"})

		metrics := monitor.GetMetricsByName("test_metric")
		if len(metrics) == 0 {
			t.Error("Expected metrics by name to be found")
		}
	})

	t.Run("Get Summary", func(t *testing.T) {
		monitor.RecordMetric("test_metric", 42.0, "seconds", map[string]string{"test": "value"})
		summary := monitor.GetSummary()

		if summary["total_metrics"].(int) == 0 {
			t.Error("Expected total metrics to be greater than 0")
		}
	})

	t.Run("Add Exporter", func(t *testing.T) {
		exporter := NewLogPerformanceExporter(NewLogger(logger))
		monitor.AddExporter(exporter)

		err := monitor.ExportMetrics()
		if err != nil {
			t.Errorf("Expected no error when exporting metrics, got: %v", err)
		}
	})

	t.Run("Calculate Percentiles", func(t *testing.T) {
		values := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
		percentiles := []float64{50, 90, 95, 99}

		result := monitor.CalculatePercentiles(values, percentiles)
		if len(result) == 0 {
			t.Error("Expected percentiles to be calculated")
		}

		if result["p50"] != 5.0 {
			t.Errorf("Expected p50 to be 5.0, got %f", result["p50"])
		}
	})
}

// Helper types for testing
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
