package classification

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestSecurityMetricsMonitor_Integration tests the complete security metrics monitoring system
func TestSecurityMetricsMonitor_Integration(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled:                      true,
		CollectionInterval:           100 * time.Millisecond,
		DataSourceTrustTarget:        100.0,
		WebsiteVerificationTarget:    95.0,
		ConfidenceIntegrityThreshold: 0.8,
		AlertingEnabled:              true,
		RetentionPeriod:              1 * time.Hour,
		MaxMetricsHistory:            100,
	}
	config.AlertThresholds = &SecurityAlertThresholds{
		DataSourceTrustBelow:     95.0,
		WebsiteVerificationBelow: 90.0,
		SecurityViolationsAbove:  5,
		ConfidenceIntegrityBelow: 0.7,
	}

	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Start monitoring
	monitor.Start()
	defer monitor.Stop()

	// Test data source trust tracking
	t.Run("DataSourceTrustTracking", func(t *testing.T) {
		// Record trusted sources
		monitor.RecordDataSourceTrust(ctx, true, "trusted_source_1")
		monitor.RecordDataSourceTrust(ctx, true, "trusted_source_2")
		monitor.RecordDataSourceTrust(ctx, true, "trusted_source_3")

		// Record untrusted source
		monitor.RecordDataSourceTrust(ctx, false, "untrusted_source_1")

		// Get metrics
		metrics := monitor.GetSecurityMetrics()
		if metrics.DataSourceTrust.TrustRate != 75.0 {
			t.Errorf("Expected trust rate to be 75.0, got %f", metrics.DataSourceTrust.TrustRate)
		}

		if metrics.DataSourceTrust.TrustedCount != 3 {
			t.Errorf("Expected trusted count to be 3, got %d", metrics.DataSourceTrust.TrustedCount)
		}

		if metrics.DataSourceTrust.TotalValidations != 4 {
			t.Errorf("Expected total validations to be 4, got %d", metrics.DataSourceTrust.TotalValidations)
		}
	})

	// Test website verification tracking
	t.Run("WebsiteVerificationTracking", func(t *testing.T) {
		// Record successful verifications
		monitor.RecordWebsiteVerification(ctx, true, "example.com", "dns")
		monitor.RecordWebsiteVerification(ctx, true, "test.com", "dns")
		monitor.RecordWebsiteVerification(ctx, true, "demo.com", "dns")

		// Record failed verification
		monitor.RecordWebsiteVerification(ctx, false, "invalid.com", "dns")

		// Get metrics
		metrics := monitor.GetSecurityMetrics()
		if metrics.WebsiteVerification.SuccessRate != 75.0 {
			t.Errorf("Expected success rate to be 75.0, got %f", metrics.WebsiteVerification.SuccessRate)
		}

		if metrics.WebsiteVerification.SuccessCount != 3 {
			t.Errorf("Expected success count to be 3, got %d", metrics.WebsiteVerification.SuccessCount)
		}

		if metrics.WebsiteVerification.TotalAttempts != 4 {
			t.Errorf("Expected total attempts to be 4, got %d", metrics.WebsiteVerification.TotalAttempts)
		}
	})

	// Test security violation tracking
	t.Run("SecurityViolationTracking", func(t *testing.T) {
		// Record security violations
		monitor.RecordSecurityViolation(ctx, "untrusted_data", "high", "Untrusted data source detected", "classifier")
		monitor.RecordSecurityViolation(ctx, "unverified_website", "medium", "Unverified website URL provided", "verifier")
		monitor.RecordSecurityViolation(ctx, "untrusted_data", "high", "Another untrusted data source", "classifier")

		// Get metrics
		metrics := monitor.GetSecurityMetrics()
		if metrics.SecurityViolations.TotalViolations != 3 {
			t.Errorf("Expected total violations to be 3, got %d", metrics.SecurityViolations.TotalViolations)
		}

		violationsByType := metrics.SecurityViolations.ViolationsByType
		if violationsByType["untrusted_data"] != 2 {
			t.Errorf("Expected 2 untrusted_data violations, got %d", violationsByType["untrusted_data"])
		}

		if violationsByType["unverified_website"] != 1 {
			t.Errorf("Expected 1 unverified_website violation, got %d", violationsByType["unverified_website"])
		}
	})

	// Test confidence integrity tracking
	t.Run("ConfidenceIntegrityTracking", func(t *testing.T) {
		// Record valid confidence scores
		monitor.RecordConfidenceScore(ctx, 0.95, true, "classifier")
		monitor.RecordConfidenceScore(ctx, 0.87, true, "classifier")
		monitor.RecordConfidenceScore(ctx, 0.92, true, "classifier")

		// Record invalid confidence score
		monitor.RecordConfidenceScore(ctx, 1.5, false, "classifier") // Invalid: > 1.0

		// Get metrics
		metrics := monitor.GetSecurityMetrics()
		if metrics.ConfidenceIntegrity.IntegrityRate != 0.75 {
			t.Errorf("Expected integrity rate to be 0.75, got %f", metrics.ConfidenceIntegrity.IntegrityRate)
		}

		if metrics.ConfidenceIntegrity.ValidScores != 3 {
			t.Errorf("Expected valid scores to be 3, got %d", metrics.ConfidenceIntegrity.ValidScores)
		}

		if metrics.ConfidenceIntegrity.TotalScores != 4 {
			t.Errorf("Expected total scores to be 4, got %d", metrics.ConfidenceIntegrity.TotalScores)
		}
	})

	// Test alerting system
	t.Run("AlertingSystem", func(t *testing.T) {
		// Wait for monitoring loop to run
		time.Sleep(200 * time.Millisecond)

		// Get metrics and check for alerts
		metrics := monitor.GetSecurityMetrics()
		if len(metrics.Alerts) == 0 {
			t.Error("Expected alerts to be generated for low trust rates")
		}

		// Check for specific alert types
		alertTypes := make(map[string]bool)
		for _, alert := range metrics.Alerts {
			alertTypes[alert.Type] = true
		}

		expectedAlertTypes := []string{"data_source_trust", "website_verification"}
		for _, expectedType := range expectedAlertTypes {
			if !alertTypes[expectedType] {
				t.Errorf("Expected alert type %s to be present", expectedType)
			}
		}
	})

	// Test performance tracking
	t.Run("PerformanceTracking", func(t *testing.T) {
		// Wait for some collections
		time.Sleep(300 * time.Millisecond)

		// Get metrics and check performance
		metrics := monitor.GetSecurityMetrics()
		if metrics.Performance.TotalCollections == 0 {
			t.Error("Expected performance collections to be tracked")
		}

		if metrics.Performance.AverageCollectionTime == 0 {
			t.Error("Expected average collection time to be tracked")
		}
	})

	// Test comprehensive metrics
	t.Run("ComprehensiveMetrics", func(t *testing.T) {
		metrics := monitor.GetSecurityMetrics()

		// Verify all components are present
		if metrics.DataSourceTrust == nil {
			t.Error("Expected DataSourceTrust metrics to be present")
		}

		if metrics.WebsiteVerification == nil {
			t.Error("Expected WebsiteVerification metrics to be present")
		}

		if metrics.SecurityViolations == nil {
			t.Error("Expected SecurityViolations metrics to be present")
		}

		if metrics.ConfidenceIntegrity == nil {
			t.Error("Expected ConfidenceIntegrity metrics to be present")
		}

		if metrics.Alerts == nil {
			t.Error("Expected Alerts to be present")
		}

		if metrics.Performance == nil {
			t.Error("Expected Performance metrics to be present")
		}

		// Verify timestamp is recent
		if time.Since(metrics.Timestamp) > 1*time.Second {
			t.Error("Expected metrics timestamp to be recent")
		}
	})
}

// TestSecurityMetricsMonitor_DisabledState tests the monitor when disabled
func TestSecurityMetricsMonitor_DisabledState(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled: false, // Disabled
	}

	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Record metrics when disabled
	monitor.RecordDataSourceTrust(ctx, true, "test_source")
	monitor.RecordWebsiteVerification(ctx, true, "test.com", "dns")
	monitor.RecordSecurityViolation(ctx, "test", "low", "Test violation", "test")
	monitor.RecordConfidenceScore(ctx, 0.9, true, "test")

	// Check that no data was recorded
	metrics := monitor.GetSecurityMetrics()
	if metrics.DataSourceTrust.TotalValidations != 0 {
		t.Error("Expected no data to be recorded when monitoring is disabled")
	}
}

// TestSecurityMetricsMonitor_ConcurrentAccess tests concurrent access to the monitor
func TestSecurityMetricsMonitor_ConcurrentAccess(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled: true,
	}

	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Test concurrent access
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Record various metrics concurrently
			monitor.RecordDataSourceTrust(ctx, id%2 == 0, "source_"+string(rune(id)))
			monitor.RecordWebsiteVerification(ctx, id%3 == 0, "domain_"+string(rune(id))+".com", "dns")
			monitor.RecordSecurityViolation(ctx, "concurrent_test", "low", "Test "+string(rune(id)), "test")
			monitor.RecordConfidenceScore(ctx, float64(id)/10.0, id%2 == 0, "test")

			// Get metrics concurrently
			_ = monitor.GetSecurityMetrics()
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify final state
	metrics := monitor.GetSecurityMetrics()
	if metrics.DataSourceTrust.TotalValidations != 10 {
		t.Errorf("Expected 10 data source validations, got %d", metrics.DataSourceTrust.TotalValidations)
	}

	if metrics.WebsiteVerification.TotalAttempts != 10 {
		t.Errorf("Expected 10 website verification attempts, got %d", metrics.WebsiteVerification.TotalAttempts)
	}

	if metrics.SecurityViolations.TotalViolations != 10 {
		t.Errorf("Expected 10 security violations, got %d", metrics.SecurityViolations.TotalViolations)
	}

	if metrics.ConfidenceIntegrity.TotalScores != 10 {
		t.Errorf("Expected 10 confidence scores, got %d", metrics.ConfidenceIntegrity.TotalScores)
	}
}
