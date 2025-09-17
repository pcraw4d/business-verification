package classification

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestSecurityMetricsMonitor_Initialization(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled:                      true,
		CollectionInterval:           1 * time.Second,
		DataSourceTrustTarget:        100.0,
		WebsiteVerificationTarget:    95.0,
		ConfidenceIntegrityThreshold: 0.8,
		AlertingEnabled:              true,
		RetentionPeriod:              1 * time.Hour,
		MaxMetricsHistory:            100,
	}

	monitor := NewSecurityMetricsMonitor(config, logger)

	if monitor == nil {
		t.Fatal("Expected monitor to be created")
	}

	if !monitor.config.Enabled {
		t.Error("Expected monitoring to be enabled")
	}

	if monitor.config.DataSourceTrustTarget != 100.0 {
		t.Errorf("Expected data source trust target to be 100.0, got %f", monitor.config.DataSourceTrustTarget)
	}

	if monitor.config.WebsiteVerificationTarget != 95.0 {
		t.Errorf("Expected website verification target to be 95.0, got %f", monitor.config.WebsiteVerificationTarget)
	}
}

func TestSecurityMetricsMonitor_DataSourceTrustTracking(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled:               true,
		DataSourceTrustTarget: 100.0,
		AlertingEnabled:       true,
	}
	config.AlertThresholds = &SecurityAlertThresholds{
		DataSourceTrustBelow: 95.0,
	}

	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Test initial state
	metrics := monitor.GetSecurityMetrics()
	if metrics.DataSourceTrust.TrustRate != 0 {
		t.Errorf("Expected initial trust rate to be 0, got %f", metrics.DataSourceTrust.TrustRate)
	}

	// Record trusted sources
	monitor.RecordDataSourceTrust(ctx, true, "trusted_source_1")
	monitor.RecordDataSourceTrust(ctx, true, "trusted_source_2")
	monitor.RecordDataSourceTrust(ctx, true, "trusted_source_3")

	// Record untrusted source
	monitor.RecordDataSourceTrust(ctx, false, "untrusted_source_1")

	// Check metrics
	metrics = monitor.GetSecurityMetrics()
	expectedTrustRate := 75.0 // 3 out of 4 sources trusted
	if metrics.DataSourceTrust.TrustRate != expectedTrustRate {
		t.Errorf("Expected trust rate to be %f, got %f", expectedTrustRate, metrics.DataSourceTrust.TrustRate)
	}

	if metrics.DataSourceTrust.TrustedCount != 3 {
		t.Errorf("Expected trusted count to be 3, got %d", metrics.DataSourceTrust.TrustedCount)
	}

	if metrics.DataSourceTrust.TotalValidations != 4 {
		t.Errorf("Expected total validations to be 4, got %d", metrics.DataSourceTrust.TotalValidations)
	}
}

func TestSecurityMetricsMonitor_WebsiteVerificationTracking(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled:                   true,
		WebsiteVerificationTarget: 95.0,
		AlertingEnabled:           true,
	}
	config.AlertThresholds = &SecurityAlertThresholds{
		WebsiteVerificationBelow: 90.0,
	}

	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Test initial state
	metrics := monitor.GetSecurityMetrics()
	if metrics.WebsiteVerification.SuccessRate != 0 {
		t.Errorf("Expected initial success rate to be 0, got %f", metrics.WebsiteVerification.SuccessRate)
	}

	// Record successful verifications
	monitor.RecordWebsiteVerification(ctx, true, "example.com", "dns")
	monitor.RecordWebsiteVerification(ctx, true, "test.com", "dns")
	monitor.RecordWebsiteVerification(ctx, true, "demo.com", "dns")

	// Record failed verification
	monitor.RecordWebsiteVerification(ctx, false, "invalid.com", "dns")

	// Check metrics
	metrics = monitor.GetSecurityMetrics()
	expectedSuccessRate := 75.0 // 3 out of 4 verifications successful
	if metrics.WebsiteVerification.SuccessRate != expectedSuccessRate {
		t.Errorf("Expected success rate to be %f, got %f", expectedSuccessRate, metrics.WebsiteVerification.SuccessRate)
	}

	if metrics.WebsiteVerification.SuccessCount != 3 {
		t.Errorf("Expected success count to be 3, got %d", metrics.WebsiteVerification.SuccessCount)
	}

	if metrics.WebsiteVerification.TotalAttempts != 4 {
		t.Errorf("Expected total attempts to be 4, got %d", metrics.WebsiteVerification.TotalAttempts)
	}
}

func TestSecurityMetricsMonitor_SecurityViolationTracking(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled:         true,
		AlertingEnabled: true,
	}
	config.AlertThresholds = &SecurityAlertThresholds{
		SecurityViolationsAbove: 5,
	}

	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Test initial state
	metrics := monitor.GetSecurityMetrics()
	if metrics.SecurityViolations.TotalViolations != 0 {
		t.Errorf("Expected initial violations to be 0, got %d", metrics.SecurityViolations.TotalViolations)
	}

	// Record security violations
	monitor.RecordSecurityViolation(ctx, "untrusted_data", "high", "Untrusted data source detected", "classifier")
	monitor.RecordSecurityViolation(ctx, "unverified_website", "medium", "Unverified website URL provided", "verifier")
	monitor.RecordSecurityViolation(ctx, "untrusted_data", "high", "Another untrusted data source", "classifier")

	// Check metrics
	metrics = monitor.GetSecurityMetrics()
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

	// Check recent violations
	recentViolations := metrics.SecurityViolations.RecentViolations
	if len(recentViolations) != 3 {
		t.Errorf("Expected 3 recent violations, got %d", len(recentViolations))
	}
}

func TestSecurityMetricsMonitor_ConfidenceIntegrityTracking(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled:                      true,
		ConfidenceIntegrityThreshold: 0.8,
		AlertingEnabled:              true,
	}
	config.AlertThresholds = &SecurityAlertThresholds{
		ConfidenceIntegrityBelow: 0.7,
	}

	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Test initial state
	metrics := monitor.GetSecurityMetrics()
	if metrics.ConfidenceIntegrity.IntegrityRate != 0 {
		t.Errorf("Expected initial integrity rate to be 0, got %f", metrics.ConfidenceIntegrity.IntegrityRate)
	}

	// Record valid confidence scores
	monitor.RecordConfidenceScore(ctx, 0.95, true, "classifier")
	monitor.RecordConfidenceScore(ctx, 0.87, true, "classifier")
	monitor.RecordConfidenceScore(ctx, 0.92, true, "classifier")

	// Record invalid confidence score
	monitor.RecordConfidenceScore(ctx, 1.5, false, "classifier") // Invalid: > 1.0

	// Check metrics
	metrics = monitor.GetSecurityMetrics()
	expectedIntegrityRate := 0.75 // 3 out of 4 scores valid
	if metrics.ConfidenceIntegrity.IntegrityRate != expectedIntegrityRate {
		t.Errorf("Expected integrity rate to be %f, got %f", expectedIntegrityRate, metrics.ConfidenceIntegrity.IntegrityRate)
	}

	if metrics.ConfidenceIntegrity.ValidScores != 3 {
		t.Errorf("Expected valid scores to be 3, got %d", metrics.ConfidenceIntegrity.ValidScores)
	}

	if metrics.ConfidenceIntegrity.TotalScores != 4 {
		t.Errorf("Expected total scores to be 4, got %d", metrics.ConfidenceIntegrity.TotalScores)
	}
}

func TestSecurityMetricsMonitor_Alerting(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled:         true,
		AlertingEnabled: true,
	}
	config.AlertThresholds = &SecurityAlertThresholds{
		DataSourceTrustBelow:     95.0,
		WebsiteVerificationBelow: 90.0,
		SecurityViolationsAbove:  2,
		ConfidenceIntegrityBelow: 0.7,
	}

	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Trigger data source trust alert
	monitor.RecordDataSourceTrust(ctx, false, "untrusted_source") // 0% trust rate

	// Trigger website verification alert
	monitor.RecordWebsiteVerification(ctx, false, "invalid.com", "dns") // 0% success rate

	// Trigger security violation alert
	monitor.RecordSecurityViolation(ctx, "test_violation", "high", "Test violation", "test")
	monitor.RecordSecurityViolation(ctx, "test_violation", "high", "Test violation 2", "test")
	monitor.RecordSecurityViolation(ctx, "test_violation", "high", "Test violation 3", "test") // 3 > 2 threshold

	// Trigger confidence integrity alert
	monitor.RecordConfidenceScore(ctx, 1.5, false, "test") // Invalid score

	// Check alerts
	metrics := monitor.GetSecurityMetrics()
	alerts := metrics.Alerts

	if len(alerts) == 0 {
		t.Error("Expected alerts to be generated")
	}

	// Check for specific alert types
	alertTypes := make(map[string]bool)
	for _, alert := range alerts {
		alertTypes[alert.Type] = true
	}

	expectedAlertTypes := []string{"data_source_trust", "website_verification", "security_violations", "confidence_integrity"}
	for _, expectedType := range expectedAlertTypes {
		if !alertTypes[expectedType] {
			t.Errorf("Expected alert type %s to be present", expectedType)
		}
	}
}

func TestSecurityMetricsMonitor_PerformanceTracking(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled:            true,
		CollectionInterval: 100 * time.Millisecond,
	}

	monitor := NewSecurityMetricsMonitor(config, logger)

	// Start monitoring
	monitor.Start()
	defer monitor.Stop()

	// Wait for some collections
	time.Sleep(500 * time.Millisecond)

	// Check performance metrics
	metrics := monitor.GetSecurityMetrics()
	performance := metrics.Performance

	if performance.TotalCollections == 0 {
		t.Error("Expected performance collections to be tracked")
	}

	if performance.AverageCollectionTime == 0 {
		t.Error("Expected average collection time to be tracked")
	}
}

func TestSecurityMetricsMonitor_DataCleanup(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled:         true,
		RetentionPeriod: 1 * time.Millisecond, // Very short retention for testing
	}

	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Record some data
	monitor.RecordDataSourceTrust(ctx, true, "test_source")
	monitor.RecordWebsiteVerification(ctx, true, "test.com", "dns")
	monitor.RecordSecurityViolation(ctx, "test", "low", "Test violation", "test")
	monitor.RecordConfidenceScore(ctx, 0.9, true, "test")

	// Wait for retention period to pass
	time.Sleep(10 * time.Millisecond)

	// Trigger cleanup
	monitor.cleanupOldData()

	// Data should still be present (cleanup only affects historical data)
	metrics := monitor.GetSecurityMetrics()
	if metrics.DataSourceTrust.TotalValidations == 0 {
		t.Error("Expected current data to be preserved after cleanup")
	}
}

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
			monitor.RecordDataSourceTrust(ctx, id%2 == 0, fmt.Sprintf("source_%d", id))
			monitor.RecordWebsiteVerification(ctx, id%3 == 0, fmt.Sprintf("domain_%d.com", id), "dns")
			monitor.RecordSecurityViolation(ctx, "concurrent_test", "low", fmt.Sprintf("Test %d", id), "test")
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

func TestSecurityMetricsMonitor_AlertCooldown(t *testing.T) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{
		Enabled:         true,
		AlertingEnabled: true,
	}
	config.AlertThresholds = &SecurityAlertThresholds{
		DataSourceTrustBelow: 95.0,
	}

	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Trigger alert
	monitor.RecordDataSourceTrust(ctx, false, "untrusted_source")

	// Get initial alerts
	metrics1 := monitor.GetSecurityMetrics()
	initialAlertCount := len(metrics1.Alerts)

	// Trigger same condition immediately (should be cooldown)
	monitor.RecordDataSourceTrust(ctx, false, "another_untrusted_source")

	// Check that no new alert was created due to cooldown
	metrics2 := monitor.GetSecurityMetrics()
	if len(metrics2.Alerts) != initialAlertCount {
		t.Error("Expected no new alert due to cooldown period")
	}
}

// Benchmark tests
func BenchmarkSecurityMetricsMonitor_RecordDataSourceTrust(b *testing.B) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{Enabled: true}
	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.RecordDataSourceTrust(ctx, i%2 == 0, fmt.Sprintf("source_%d", i))
	}
}

func BenchmarkSecurityMetricsMonitor_GetMetrics(b *testing.B) {
	logger := zap.NewNop()
	config := &SecurityMetricsConfig{Enabled: true}
	monitor := NewSecurityMetricsMonitor(config, logger)
	ctx := context.Background()

	// Pre-populate with some data
	for i := 0; i < 1000; i++ {
		monitor.RecordDataSourceTrust(ctx, i%2 == 0, fmt.Sprintf("source_%d", i))
		monitor.RecordWebsiteVerification(ctx, i%3 == 0, fmt.Sprintf("domain_%d.com", i), "dns")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = monitor.GetSecurityMetrics()
	}
}
