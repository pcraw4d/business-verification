package classification_monitoring

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestSecurityMetricsAccuracyTracker_Creation(t *testing.T) {
	config := DefaultSecurityMetricsConfig()
	logger := zap.NewNop()

	tracker := NewSecurityMetricsAccuracyTracker(config, logger)

	if tracker == nil {
		t.Fatal("Expected tracker to be created")
	}

	if tracker.config == nil {
		t.Fatal("Expected config to be set")
	}

	if tracker.logger == nil {
		t.Fatal("Expected logger to be set")
	}

	if tracker.trustedDataSources == nil {
		t.Fatal("Expected trusted data sources map to be initialized")
	}

	if tracker.websiteVerifications == nil {
		t.Fatal("Expected website verifications map to be initialized")
	}

	if tracker.securityViolations == nil {
		t.Fatal("Expected security violations map to be initialized")
	}

	if tracker.confidenceIntegrity == nil {
		t.Fatal("Expected confidence integrity to be initialized")
	}

	if tracker.securityAnalyzer == nil {
		t.Fatal("Expected security analyzer to be initialized")
	}

	if tracker.threatDetector == nil {
		t.Fatal("Expected threat detector to be initialized")
	}

	if tracker.complianceMonitor == nil {
		t.Fatal("Expected compliance monitor to be initialized")
	}

	if tracker.auditLogger == nil {
		t.Fatal("Expected audit logger to be initialized")
	}
}

func TestSecurityMetricsAccuracyTracker_TrackTrustedDataSourceResult(t *testing.T) {
	config := DefaultSecurityMetricsConfig()
	logger := zap.NewNop()

	tracker := NewSecurityMetricsAccuracyTracker(config, logger)

	// Create test classification result with trusted source metadata
	result := &ClassificationResult{
		ID:                     "test_123",
		BusinessName:           "Test Business",
		ActualClassification:   "restaurant",
		ExpectedClassification: stringPtr("restaurant"),
		ConfidenceScore:        0.95,
		ClassificationMethod:   "trusted_source_classification",
		Timestamp:              time.Now(),
		Metadata: map[string]interface{}{
			"industry":       "restaurant",
			"trusted_source": "government_database",
			"source_type":    "government",
		},
		IsCorrect: boolPtr(true),
	}

	// Track trusted data source result
	err := tracker.TrackTrustedDataSourceResult(context.Background(), result)
	if err != nil {
		t.Fatalf("Expected tracking to succeed, got error: %v", err)
	}

	// Verify trusted data source metrics
	metrics := tracker.GetTrustedDataSourceMetrics("government_database")
	if metrics == nil {
		t.Fatal("Expected trusted data source metrics to be available")
	}

	if metrics.SourceName != "government_database" {
		t.Errorf("Expected source name to be 'government_database', got '%s'", metrics.SourceName)
	}

	if metrics.SourceType != "government" {
		t.Errorf("Expected source type to be 'government', got '%s'", metrics.SourceType)
	}

	if metrics.TotalRequests != 1 {
		t.Errorf("Expected total requests to be 1, got %d", metrics.TotalRequests)
	}

	if metrics.SuccessfulRequests != 1 {
		t.Errorf("Expected successful requests to be 1, got %d", metrics.SuccessfulRequests)
	}

	if metrics.AccuracyScore != 1.0 {
		t.Errorf("Expected accuracy score to be 1.0, got %f", metrics.AccuracyScore)
	}

	if metrics.ConfidenceScore != 0.95 {
		t.Errorf("Expected confidence score to be 0.95, got %f", metrics.ConfidenceScore)
	}

	if metrics.TrustLevel != 1.0 {
		t.Errorf("Expected trust level to be 1.0, got %f", metrics.TrustLevel)
	}

	if metrics.Status != "active" {
		t.Errorf("Expected status to be 'active', got '%s'", metrics.Status)
	}
}

func TestSecurityMetricsAccuracyTracker_MultipleTrustedSources(t *testing.T) {
	config := DefaultSecurityMetricsConfig()
	logger := zap.NewNop()

	tracker := NewSecurityMetricsAccuracyTracker(config, logger)

	// Test data for multiple trusted sources
	sources := []struct {
		name       string
		sourceType string
		correct    bool
		confidence float64
	}{
		{"government_database", "government", true, 0.98},
		{"credit_bureau", "credit_bureau", true, 0.95},
		{"business_registry", "business_registry", false, 0.90},
		{"financial_institution", "financial_institution", true, 0.92},
	}

	// Track results for each source
	for i, source := range sources {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        source.confidence,
			ClassificationMethod:   "trusted_source_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":       "restaurant",
				"trusted_source": source.name,
				"source_type":    source.sourceType,
			},
			IsCorrect: boolPtr(source.correct),
		}

		err := tracker.TrackTrustedDataSourceResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track result for source %s: %v", source.name, err)
		}
	}

	// Verify all sources are tracked
	allMetrics := tracker.GetAllTrustedDataSourceMetrics()
	if len(allMetrics) != 4 {
		t.Errorf("Expected 4 trusted sources to be tracked, got %d", len(allMetrics))
	}

	// Verify specific source metrics
	governmentMetrics := tracker.GetTrustedDataSourceMetrics("government_database")
	if governmentMetrics == nil {
		t.Fatal("Expected government database metrics to be available")
	}

	if governmentMetrics.AccuracyScore != 1.0 {
		t.Errorf("Expected government database accuracy to be 1.0, got %f", governmentMetrics.AccuracyScore)
	}

	businessRegistryMetrics := tracker.GetTrustedDataSourceMetrics("business_registry")
	if businessRegistryMetrics == nil {
		t.Fatal("Expected business registry metrics to be available")
	}

	if businessRegistryMetrics.AccuracyScore != 0.0 {
		t.Errorf("Expected business registry accuracy to be 0.0, got %f", businessRegistryMetrics.AccuracyScore)
	}
}

func TestSecurityMetricsAccuracyTracker_TrackWebsiteVerification(t *testing.T) {
	config := DefaultSecurityMetricsConfig()
	logger := zap.NewNop()

	tracker := NewSecurityMetricsAccuracyTracker(config, logger)

	// Test website verification tracking
	domains := []struct {
		domain  string
		method  string
		success bool
		time    time.Duration
	}{
		{"example.com", "ssl_certificate", true, 2 * time.Second},
		{"test.org", "dns_verification", true, 1 * time.Second},
		{"invalid.net", "whois_verification", false, 5 * time.Second},
	}

	// Track verifications for each domain
	for _, domain := range domains {
		err := tracker.TrackWebsiteVerification(context.Background(), domain.domain, domain.method, domain.success, domain.time)
		if err != nil {
			t.Fatalf("Failed to track website verification for %s: %v", domain.domain, err)
		}
	}

	// Verify website verification metrics
	exampleMetrics := tracker.GetWebsiteVerificationMetrics("example.com")
	if exampleMetrics == nil {
		t.Fatal("Expected example.com verification metrics to be available")
	}

	if exampleMetrics.Domain != "example.com" {
		t.Errorf("Expected domain to be 'example.com', got '%s'", exampleMetrics.Domain)
	}

	if exampleMetrics.VerificationMethod != "ssl_certificate" {
		t.Errorf("Expected verification method to be 'ssl_certificate', got '%s'", exampleMetrics.VerificationMethod)
	}

	if exampleMetrics.TotalVerifications != 1 {
		t.Errorf("Expected total verifications to be 1, got %d", exampleMetrics.TotalVerifications)
	}

	if exampleMetrics.SuccessfulVerifications != 1 {
		t.Errorf("Expected successful verifications to be 1, got %d", exampleMetrics.SuccessfulVerifications)
	}

	if exampleMetrics.VerificationAccuracy != 1.0 {
		t.Errorf("Expected verification accuracy to be 1.0, got %f", exampleMetrics.VerificationAccuracy)
	}

	if exampleMetrics.Status != "verified" {
		t.Errorf("Expected status to be 'verified', got '%s'", exampleMetrics.Status)
	}

	// Verify failed verification
	invalidMetrics := tracker.GetWebsiteVerificationMetrics("invalid.net")
	if invalidMetrics == nil {
		t.Fatal("Expected invalid.net verification metrics to be available")
	}

	if invalidMetrics.VerificationAccuracy != 0.0 {
		t.Errorf("Expected verification accuracy to be 0.0, got %f", invalidMetrics.VerificationAccuracy)
	}

	if invalidMetrics.Status != "failed" {
		t.Errorf("Expected status to be 'failed', got '%s'", invalidMetrics.Status)
	}
}

func TestSecurityMetricsAccuracyTracker_TrackSecurityViolation(t *testing.T) {
	config := DefaultSecurityMetricsConfig()
	logger := zap.NewNop()

	tracker := NewSecurityMetricsAccuracyTracker(config, logger)

	// Test security violation tracking
	violations := []struct {
		violationType string
		source        string
		description   string
		severity      float64
	}{
		{"data_tampering", "external_api", "Suspicious data modification detected", 0.9},
		{"unauthorized_access", "internal_system", "Unauthorized access attempt", 0.8},
		{"suspicious_pattern", "user_behavior", "Unusual access pattern detected", 0.6},
		{"integrity_failure", "data_source", "Data integrity check failed", 0.7},
	}

	// Track violations
	for _, violation := range violations {
		err := tracker.TrackSecurityViolation(context.Background(), violation.violationType, violation.source, violation.description, violation.severity)
		if err != nil {
			t.Fatalf("Failed to track security violation %s: %v", violation.violationType, err)
		}
	}

	// Verify security violation metrics
	tamperingMetrics := tracker.GetSecurityViolationMetrics("data_tampering")
	if tamperingMetrics == nil {
		t.Fatal("Expected data tampering violation metrics to be available")
	}

	if tamperingMetrics.ViolationType != "data_tampering" {
		t.Errorf("Expected violation type to be 'data_tampering', got '%s'", tamperingMetrics.ViolationType)
	}

	if tamperingMetrics.TotalViolations != 1 {
		t.Errorf("Expected total violations to be 1, got %d", tamperingMetrics.TotalViolations)
	}

	if tamperingMetrics.CriticalViolations != 1 {
		t.Errorf("Expected critical violations to be 1, got %d", tamperingMetrics.CriticalViolations)
	}

	if tamperingMetrics.AverageSeverity != 0.9 {
		t.Errorf("Expected average severity to be 0.9, got %f", tamperingMetrics.AverageSeverity)
	}

	if tamperingMetrics.Status != "monitoring" {
		t.Errorf("Expected status to be 'monitoring', got '%s'", tamperingMetrics.Status)
	}
}

func TestSecurityMetricsAccuracyTracker_HistoricalData(t *testing.T) {
	config := DefaultSecurityMetricsConfig()
	logger := zap.NewNop()

	tracker := NewSecurityMetricsAccuracyTracker(config, logger)

	// Add multiple results to generate historical data
	for i := 0; i < 15; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_historical_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "trusted_source_classification",
			Timestamp:              time.Now().Add(-time.Duration(i) * time.Minute),
			Metadata: map[string]interface{}{
				"industry":       "restaurant",
				"trusted_source": "government_database",
				"source_type":    "government",
			},
			IsCorrect: boolPtr(true),
		}

		err := tracker.TrackTrustedDataSourceResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track result: %v", err)
		}
	}

	// Verify historical data
	metrics := tracker.GetTrustedDataSourceMetrics("government_database")
	if metrics == nil {
		t.Fatal("Expected government database metrics to be available")
	}

	if len(metrics.HistoricalAccuracy) == 0 {
		t.Error("Expected historical accuracy data to be populated")
	}

	if len(metrics.HistoricalConfidence) == 0 {
		t.Error("Expected historical confidence data to be populated")
	}

	if len(metrics.HistoricalLatency) == 0 {
		t.Error("Expected historical latency data to be populated")
	}

	// Verify data points have correct values
	if len(metrics.HistoricalAccuracy) != 15 {
		t.Errorf("Expected 15 historical accuracy points, got %d", len(metrics.HistoricalAccuracy))
	}

	if len(metrics.HistoricalConfidence) != 15 {
		t.Errorf("Expected 15 historical confidence points, got %d", len(metrics.HistoricalConfidence))
	}

	if len(metrics.HistoricalLatency) != 15 {
		t.Errorf("Expected 15 historical latency points, got %d", len(metrics.HistoricalLatency))
	}
}

func TestSecurityMetricsAccuracyTracker_SecurityAlerts(t *testing.T) {
	config := DefaultSecurityMetricsConfig()
	config.TrustedSourceAccuracyThreshold = 0.95 // Set high threshold to trigger alerts
	logger := zap.NewNop()

	tracker := NewSecurityMetricsAccuracyTracker(config, logger)

	// Create results that will trigger security alerts (low accuracy)
	for i := 0; i < 10; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_alert_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("technology"), // Wrong classification
			ConfidenceScore:        0.95,
			ClassificationMethod:   "trusted_source_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":       "restaurant",
				"trusted_source": "degraded_source",
				"source_type":    "government",
			},
			IsCorrect: boolPtr(false), // All incorrect to trigger accuracy alert
		}

		err := tracker.TrackTrustedDataSourceResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track result: %v", err)
		}
	}

	// Verify security alerts
	alerts := tracker.GetSecurityAlerts()
	if len(alerts) == 0 {
		t.Error("Expected security alerts to be generated")
	}

	// Check for accuracy degradation alert
	accuracyAlertFound := false
	for _, alert := range alerts {
		if alert.AlertType == "trusted_source_degradation" {
			accuracyAlertFound = true
			if alert.Severity == "" {
				t.Error("Expected alert to have severity")
			}
			if alert.Message == "" {
				t.Error("Expected alert to have message")
			}
			if alert.Source == "" {
				t.Error("Expected alert to have source")
			}
			break
		}
	}

	if !accuracyAlertFound {
		t.Error("Expected trusted source degradation alert to be generated")
	}

	// Verify active alerts
	activeAlerts := tracker.GetActiveSecurityAlerts()
	if len(activeAlerts) == 0 {
		t.Error("Expected active security alerts")
	}

	// All alerts should be unresolved initially
	for _, alert := range activeAlerts {
		if alert.Resolved {
			t.Error("Expected active alerts to be unresolved")
		}
	}
}

func TestSecurityMetricsAccuracyTracker_ConfidenceIntegrityMetrics(t *testing.T) {
	config := DefaultSecurityMetricsConfig()
	logger := zap.NewNop()

	tracker := NewSecurityMetricsAccuracyTracker(config, logger)

	// Add results to generate confidence and integrity data
	for i := 0; i < 20; i++ {
		result := &ClassificationResult{
			ID:                     fmt.Sprintf("test_integrity_%d", i),
			BusinessName:           fmt.Sprintf("Test Business %d", i),
			ActualClassification:   "restaurant",
			ExpectedClassification: stringPtr("restaurant"),
			ConfidenceScore:        0.95,
			ClassificationMethod:   "trusted_source_classification",
			Timestamp:              time.Now(),
			Metadata: map[string]interface{}{
				"industry":       "restaurant",
				"trusted_source": "reliable_source",
				"source_type":    "government",
			},
			IsCorrect: boolPtr(true),
		}

		err := tracker.TrackTrustedDataSourceResult(context.Background(), result)
		if err != nil {
			t.Fatalf("Failed to track result: %v", err)
		}
	}

	// Verify confidence and integrity metrics
	integrityMetrics := tracker.GetConfidenceIntegrityMetrics()
	if integrityMetrics == nil {
		t.Fatal("Expected confidence integrity metrics to be available")
	}

	// Check integrity status
	if integrityMetrics.IntegrityStatus == "" {
		t.Error("Expected integrity status to be set")
	}

	// Status should be reasonable
	validStatuses := []string{"high", "medium", "low", "critical"}
	statusValid := false
	for _, status := range validStatuses {
		if integrityMetrics.IntegrityStatus == status {
			statusValid = true
			break
		}
	}

	if !statusValid {
		t.Errorf("Expected integrity status to be one of %v, got '%s'", validStatuses, integrityMetrics.IntegrityStatus)
	}
}
