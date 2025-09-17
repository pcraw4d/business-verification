package monitoring

import (
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TestAdvancedSecurityMonitor_RecordDataSourceRequest(t *testing.T) {
	// Setup
	logger := observability.NewLogger(zap.NewNop())
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	config := &AdvancedSecurityMonitorConfig{
		TrustRateTarget:            100.0,
		TrustRateAlertThreshold:    95.0,
		TrustRateCheckInterval:     30 * time.Second,
		TrustRateHistorySize:       100,
		VerificationRateTarget:     95.0,
		VerificationAlertThreshold: 90.0,
		VerificationCheckInterval:  30 * time.Second,
		VerificationHistorySize:    100,
		ViolationDetectionEnabled:  true,
		ViolationCheckInterval:     1 * time.Minute,
		ViolationHistorySize:       100,
		ConfidenceIntegrityEnabled: true,
		ConfidenceCheckInterval:    1 * time.Minute,
		ConfidenceHistorySize:      100,
		AlertingEnabled:            true,
		AlertCooldown:              5 * time.Minute,
		AlertHistorySize:           100,
	}

	monitor := NewAdvancedSecurityMonitor(config, logger, tracer)
	defer monitor.Shutdown()

	// Test data source request recording
	tests := []struct {
		name           string
		dataSourceID   string
		dataSourceName string
		trusted        bool
		expectedRate   float64
		expectedStatus TrustStatus
	}{
		{
			name:           "trusted request",
			dataSourceID:   "source1",
			dataSourceName: "Test Source 1",
			trusted:        true,
			expectedRate:   100.0,
			expectedStatus: TrustStatusExcellent,
		},
		{
			name:           "untrusted request",
			dataSourceID:   "source1",
			dataSourceName: "Test Source 1",
			trusted:        false,
			expectedRate:   50.0,
			expectedStatus: TrustStatusCritical,
		},
		{
			name:           "mixed requests",
			dataSourceID:   "source1",
			dataSourceName: "Test Source 1",
			trusted:        true,
			expectedRate:   66.67, // 2 out of 3 trusted
			expectedStatus: TrustStatusCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor.RecordDataSourceRequest(tt.dataSourceID, tt.dataSourceName, tt.trusted)

			trustRates := monitor.GetTrustRates()
			trustData, exists := trustRates[tt.dataSourceID]
			if !exists {
				t.Fatalf("Expected trust data for source %s", tt.dataSourceID)
			}

			if trustData.DataSourceID != tt.dataSourceID {
				t.Errorf("Expected DataSourceID %s, got %s", tt.dataSourceID, trustData.DataSourceID)
			}

			if trustData.DataSourceName != tt.dataSourceName {
				t.Errorf("Expected DataSourceName %s, got %s", tt.dataSourceName, trustData.DataSourceName)
			}

			// Allow for small floating point differences
			if math.Abs(trustData.TrustRate-tt.expectedRate) > 0.01 {
				t.Errorf("Expected trust rate %.2f, got %.2f", tt.expectedRate, trustData.TrustRate)
			}

			if trustData.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, trustData.Status)
			}
		})
	}
}

func TestAdvancedSecurityMonitor_RecordWebsiteVerification(t *testing.T) {
	// Setup
	logger := observability.NewLogger(zap.NewNop())
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	config := &AdvancedSecurityMonitorConfig{
		TrustRateTarget:            100.0,
		TrustRateAlertThreshold:    95.0,
		TrustRateCheckInterval:     30 * time.Second,
		TrustRateHistorySize:       100,
		VerificationRateTarget:     95.0,
		VerificationAlertThreshold: 90.0,
		VerificationCheckInterval:  30 * time.Second,
		VerificationHistorySize:    100,
		ViolationDetectionEnabled:  true,
		ViolationCheckInterval:     1 * time.Minute,
		ViolationHistorySize:       100,
		ConfidenceIntegrityEnabled: true,
		ConfidenceCheckInterval:    1 * time.Minute,
		ConfidenceHistorySize:      100,
		AlertingEnabled:            true,
		AlertCooldown:              5 * time.Minute,
		AlertHistorySize:           100,
	}

	monitor := NewAdvancedSecurityMonitor(config, logger, tracer)
	defer monitor.Shutdown()

	// Test website verification recording
	tests := []struct {
		name           string
		domain         string
		successful     bool
		expectedRate   float64
		expectedStatus VerificationStatus
	}{
		{
			name:           "successful verification",
			domain:         "example.com",
			successful:     true,
			expectedRate:   100.0,
			expectedStatus: VerificationStatusExcellent,
		},
		{
			name:           "failed verification",
			domain:         "example.com",
			successful:     false,
			expectedRate:   50.0,
			expectedStatus: VerificationStatusCritical,
		},
		{
			name:           "mixed verifications",
			domain:         "example.com",
			successful:     true,
			expectedRate:   66.67, // 2 out of 3 successful
			expectedStatus: VerificationStatusCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor.RecordWebsiteVerification(tt.domain, tt.successful)

			verificationRates := monitor.GetVerificationRates()
			verificationData, exists := verificationRates[tt.domain]
			if !exists {
				t.Fatalf("Expected verification data for domain %s", tt.domain)
			}

			if verificationData.WebsiteDomain != tt.domain {
				t.Errorf("Expected WebsiteDomain %s, got %s", tt.domain, verificationData.WebsiteDomain)
			}

			// Allow for small floating point differences
			if math.Abs(verificationData.VerificationRate-tt.expectedRate) > 0.01 {
				t.Errorf("Expected verification rate %.2f, got %.2f", tt.expectedRate, verificationData.VerificationRate)
			}

			if verificationData.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, verificationData.Status)
			}
		})
	}
}

func TestAdvancedSecurityMonitor_RecordSecurityViolation(t *testing.T) {
	// Setup
	logger := observability.NewLogger(zap.NewNop())
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	config := &AdvancedSecurityMonitorConfig{
		TrustRateTarget:            100.0,
		TrustRateAlertThreshold:    95.0,
		TrustRateCheckInterval:     30 * time.Second,
		TrustRateHistorySize:       100,
		VerificationRateTarget:     95.0,
		VerificationAlertThreshold: 90.0,
		VerificationCheckInterval:  30 * time.Second,
		VerificationHistorySize:    100,
		ViolationDetectionEnabled:  true,
		ViolationCheckInterval:     1 * time.Minute,
		ViolationHistorySize:       100,
		ConfidenceIntegrityEnabled: true,
		ConfidenceCheckInterval:    1 * time.Minute,
		ConfidenceHistorySize:      100,
		AlertingEnabled:            true,
		AlertCooldown:              5 * time.Minute,
		AlertHistorySize:           100,
	}

	monitor := NewAdvancedSecurityMonitor(config, logger, tracer)
	defer monitor.Shutdown()

	// Test security violation recording
	tests := []struct {
		name          string
		violationType ViolationType
		severity      ViolationSeverity
		source        string
		description   string
		details       map[string]interface{}
	}{
		{
			name:          "untrusted data source violation",
			violationType: ViolationTypeUntrustedDataSource,
			severity:      ViolationSeverityHigh,
			source:        "classification_service",
			description:   "Untrusted data source detected",
			details: map[string]interface{}{
				"data_source": "untrusted_api",
				"request_id":  "req_123",
			},
		},
		{
			name:          "website verification failure",
			violationType: ViolationTypeWebsiteVerificationFailure,
			severity:      ViolationSeverityMedium,
			source:        "website_verifier",
			description:   "Website verification failed",
			details: map[string]interface{}{
				"domain":     "suspicious-site.com",
				"error_code": "SSL_ERROR",
			},
		},
		{
			name:          "confidence score manipulation",
			violationType: ViolationTypeConfidenceScoreManipulation,
			severity:      ViolationSeverityCritical,
			source:        "confidence_calculator",
			description:   "Confidence score manipulation detected",
			details: map[string]interface{}{
				"expected_score": 0.85,
				"actual_score":   0.99,
				"difference":     0.14,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor.RecordSecurityViolation(tt.violationType, tt.severity, tt.source, tt.description, tt.details)

			violationEvents := monitor.GetViolationEvents()
			if len(violationEvents) == 0 {
				t.Fatal("Expected at least one violation event")
			}

			// Check the most recent violation event
			latestEvent := violationEvents[len(violationEvents)-1]
			if latestEvent.Type != tt.violationType {
				t.Errorf("Expected violation type %s, got %s", tt.violationType, latestEvent.Type)
			}

			if latestEvent.Severity != tt.severity {
				t.Errorf("Expected severity %s, got %s", tt.severity, latestEvent.Severity)
			}

			if latestEvent.Source != tt.source {
				t.Errorf("Expected source %s, got %s", tt.source, latestEvent.Source)
			}

			if latestEvent.Description != tt.description {
				t.Errorf("Expected description %s, got %s", tt.description, latestEvent.Description)
			}

			if latestEvent.Resolved {
				t.Error("Expected violation to be unresolved")
			}
		})
	}
}

func TestAdvancedSecurityMonitor_RecordConfidenceIntegrityEvent(t *testing.T) {
	// Setup
	logger := observability.NewLogger(zap.NewNop())
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	config := &AdvancedSecurityMonitorConfig{
		TrustRateTarget:            100.0,
		TrustRateAlertThreshold:    95.0,
		TrustRateCheckInterval:     30 * time.Second,
		TrustRateHistorySize:       100,
		VerificationRateTarget:     95.0,
		VerificationAlertThreshold: 90.0,
		VerificationCheckInterval:  30 * time.Second,
		VerificationHistorySize:    100,
		ViolationDetectionEnabled:  true,
		ViolationCheckInterval:     1 * time.Minute,
		ViolationHistorySize:       100,
		ConfidenceIntegrityEnabled: true,
		ConfidenceCheckInterval:    1 * time.Minute,
		ConfidenceHistorySize:      100,
		AlertingEnabled:            true,
		AlertCooldown:              5 * time.Minute,
		AlertHistorySize:           100,
	}

	monitor := NewAdvancedSecurityMonitor(config, logger, tracer)
	defer monitor.Shutdown()

	// Test confidence integrity event recording
	tests := []struct {
		name               string
		eventType          ConfidenceEventType
		severity           ConfidenceEventSeverity
		classificationID   string
		expectedScore      float64
		actualScore        float64
		expectedDifference float64
		details            map[string]interface{}
	}{
		{
			name:               "confidence anomaly",
			eventType:          ConfidenceEventTypeAnomaly,
			severity:           ConfidenceEventSeverityMedium,
			classificationID:   "class_123",
			expectedScore:      0.85,
			actualScore:        0.95,
			expectedDifference: 0.10,
			details: map[string]interface{}{
				"method":   "keyword_matching",
				"industry": "technology",
			},
		},
		{
			name:               "confidence manipulation",
			eventType:          ConfidenceEventTypeManipulation,
			severity:           ConfidenceEventSeverityCritical,
			classificationID:   "class_456",
			expectedScore:      0.70,
			actualScore:        0.99,
			expectedDifference: 0.29,
			details: map[string]interface{}{
				"method":   "ml_classification",
				"industry": "healthcare",
			},
		},
		{
			name:               "confidence inconsistency",
			eventType:          ConfidenceEventTypeInconsistency,
			severity:           ConfidenceEventSeverityHigh,
			classificationID:   "class_789",
			expectedScore:      0.60,
			actualScore:        0.40,
			expectedDifference: 0.20,
			details: map[string]interface{}{
				"method":   "description_analysis",
				"industry": "retail",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			monitor.RecordConfidenceIntegrityEvent(tt.eventType, tt.severity, tt.classificationID, tt.expectedScore, tt.actualScore, tt.details)

			confidenceEvents := monitor.GetConfidenceEvents()
			if len(confidenceEvents) == 0 {
				t.Fatal("Expected at least one confidence event")
			}

			// Check the most recent confidence event
			latestEvent := confidenceEvents[len(confidenceEvents)-1]
			if latestEvent.Type != tt.eventType {
				t.Errorf("Expected event type %s, got %s", tt.eventType, latestEvent.Type)
			}

			if latestEvent.Severity != tt.severity {
				t.Errorf("Expected severity %s, got %s", tt.severity, latestEvent.Severity)
			}

			if latestEvent.ClassificationID != tt.classificationID {
				t.Errorf("Expected classification ID %s, got %s", tt.classificationID, latestEvent.ClassificationID)
			}

			if latestEvent.ExpectedScore != tt.expectedScore {
				t.Errorf("Expected score %.2f, got %.2f", tt.expectedScore, latestEvent.ExpectedScore)
			}

			if latestEvent.ActualScore != tt.actualScore {
				t.Errorf("Expected actual score %.2f, got %.2f", tt.actualScore, latestEvent.ActualScore)
			}

			// Allow for small floating point differences
			if math.Abs(latestEvent.ScoreDifference-tt.expectedDifference) > 0.01 {
				t.Errorf("Expected score difference %.2f, got %.2f", tt.expectedDifference, latestEvent.ScoreDifference)
			}

			if latestEvent.Resolved {
				t.Error("Expected confidence event to be unresolved")
			}
		})
	}
}

func TestAdvancedSecurityMonitor_SecurityMetrics(t *testing.T) {
	// Setup
	logger := observability.NewLogger(zap.NewNop())
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	config := &AdvancedSecurityMonitorConfig{
		TrustRateTarget:            100.0,
		TrustRateAlertThreshold:    95.0,
		TrustRateCheckInterval:     30 * time.Second,
		TrustRateHistorySize:       100,
		VerificationRateTarget:     95.0,
		VerificationAlertThreshold: 90.0,
		VerificationCheckInterval:  30 * time.Second,
		VerificationHistorySize:    100,
		ViolationDetectionEnabled:  true,
		ViolationCheckInterval:     1 * time.Minute,
		ViolationHistorySize:       100,
		ConfidenceIntegrityEnabled: true,
		ConfidenceCheckInterval:    1 * time.Minute,
		ConfidenceHistorySize:      100,
		AlertingEnabled:            true,
		AlertCooldown:              5 * time.Minute,
		AlertHistorySize:           100,
		MaxViolationsPerHour:       10,
		ConfidenceAnomalyThreshold: 0.1,
	}

	monitor := NewAdvancedSecurityMonitor(config, logger, tracer)
	defer monitor.Shutdown()

	// Record test data
	monitor.RecordDataSourceRequest("source1", "Test Source 1", true)
	monitor.RecordDataSourceRequest("source1", "Test Source 1", true)
	monitor.RecordDataSourceRequest("source1", "Test Source 1", false)

	monitor.RecordWebsiteVerification("example.com", true)
	monitor.RecordWebsiteVerification("example.com", true)
	monitor.RecordWebsiteVerification("example.com", false)

	monitor.RecordSecurityViolation(ViolationTypeUntrustedDataSource, ViolationSeverityHigh, "test_source", "Test violation", map[string]interface{}{})

	monitor.RecordConfidenceIntegrityEvent(ConfidenceEventTypeAnomaly, ConfidenceEventSeverityMedium, "class_123", 0.85, 0.95, map[string]interface{}{})

	// Force metrics update
	monitor.UpdateSecurityMetrics()

	// Get security metrics
	metrics := monitor.GetSecurityMetrics()
	if metrics == nil {
		t.Fatal("Expected security metrics to be available")
	}

	// Test trust rate metrics
	expectedTrustRate := 66.67 // 2 out of 3 trusted
	if math.Abs(metrics.OverallTrustRate-expectedTrustRate) > 0.01 {
		t.Errorf("Expected overall trust rate %.2f, got %.2f", expectedTrustRate, metrics.OverallTrustRate)
	}

	if metrics.TrustRateBySource["source1"] <= 0 {
		t.Error("Expected trust rate by source to be populated")
	}

	if metrics.TrustRateTarget != config.TrustRateTarget {
		t.Errorf("Expected trust rate target %.2f, got %.2f", config.TrustRateTarget, metrics.TrustRateTarget)
	}

	// Test verification rate metrics
	expectedVerificationRate := 66.67 // 2 out of 3 successful
	if math.Abs(metrics.OverallVerificationRate-expectedVerificationRate) > 0.01 {
		t.Errorf("Expected overall verification rate %.2f, got %.2f", expectedVerificationRate, metrics.OverallVerificationRate)
	}

	if metrics.VerificationRateByDomain["example.com"] <= 0 {
		t.Error("Expected verification rate by domain to be populated")
	}

	if metrics.VerificationRateTarget != config.VerificationRateTarget {
		t.Errorf("Expected verification rate target %.2f, got %.2f", config.VerificationRateTarget, metrics.VerificationRateTarget)
	}

	// Test violation metrics
	if metrics.TotalViolations != 1 {
		t.Errorf("Expected total violations 1, got %d", metrics.TotalViolations)
	}

	if metrics.ViolationsByType[string(ViolationTypeUntrustedDataSource)] != 1 {
		t.Error("Expected violation by type to be populated")
	}

	if metrics.ViolationsBySeverity[string(ViolationSeverityHigh)] != 1 {
		t.Error("Expected violation by severity to be populated")
	}

	// Test confidence integrity metrics
	if metrics.TotalConfidenceEvents != 1 {
		t.Errorf("Expected total confidence events 1, got %d", metrics.TotalConfidenceEvents)
	}

	if metrics.ConfidenceEventsByType[string(ConfidenceEventTypeAnomaly)] != 1 {
		t.Error("Expected confidence event by type to be populated")
	}

	if metrics.ConfidenceEventsBySeverity[string(ConfidenceEventSeverityMedium)] != 1 {
		t.Error("Expected confidence event by severity to be populated")
	}

	// Test overall security score
	if metrics.OverallSecurityScore <= 0 {
		t.Error("Expected overall security score to be calculated")
	}

	// Test compliance flags
	if metrics.TrustRateCompliance {
		t.Error("Expected trust rate compliance to be false (below target)")
	}

	if metrics.VerificationRateCompliance {
		t.Error("Expected verification rate compliance to be false (below target)")
	}

	if !metrics.ViolationRateCompliance {
		t.Error("Expected violation rate compliance to be true (within limits)")
	}

	if !metrics.ConfidenceIntegrityCompliance {
		t.Error("Expected confidence integrity compliance to be true (within threshold)")
	}
}

func TestAdvancedSecurityMonitor_TrustStatusCalculation(t *testing.T) {
	tests := []struct {
		name           string
		trustRate      float64
		expectedStatus TrustStatus
	}{
		{"excellent trust rate", 98.5, TrustStatusExcellent},
		{"good trust rate", 92.0, TrustStatusGood},
		{"warning trust rate", 85.0, TrustStatusWarning},
		{"critical trust rate", 75.0, TrustStatusCritical},
		{"boundary excellent", 95.0, TrustStatusExcellent},
		{"boundary good", 90.0, TrustStatusGood},
		{"boundary warning", 80.0, TrustStatusWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := observability.NewLogger(zap.NewNop())
			tracer := trace.NewNoopTracerProvider().Tracer("test")
			config := &AdvancedSecurityMonitorConfig{
				TrustRateTarget:            100.0,
				TrustRateAlertThreshold:    95.0,
				TrustRateCheckInterval:     30 * time.Second,
				TrustRateHistorySize:       100,
				VerificationRateTarget:     95.0,
				VerificationAlertThreshold: 90.0,
				VerificationCheckInterval:  30 * time.Second,
				VerificationHistorySize:    100,
				ViolationDetectionEnabled:  true,
				ViolationCheckInterval:     1 * time.Minute,
				ViolationHistorySize:       100,
				ConfidenceIntegrityEnabled: true,
				ConfidenceCheckInterval:    1 * time.Minute,
				ConfidenceHistorySize:      100,
				AlertingEnabled:            true,
				AlertCooldown:              5 * time.Minute,
				AlertHistorySize:           100,
			}

			monitor := NewAdvancedSecurityMonitor(config, logger, tracer)
			defer monitor.Shutdown()

			// Record requests to achieve the desired trust rate
			totalRequests := 100
			trustedRequests := int(tt.trustRate * float64(totalRequests) / 100.0)

			for i := 0; i < trustedRequests; i++ {
				monitor.RecordDataSourceRequest("test_source", "Test Source", true)
			}

			for i := 0; i < totalRequests-trustedRequests; i++ {
				monitor.RecordDataSourceRequest("test_source", "Test Source", false)
			}

			// Check the trust status
			trustRates := monitor.GetTrustRates()
			trustData, exists := trustRates["test_source"]
			if !exists {
				t.Fatal("Expected trust data for test_source")
			}

			if trustData.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, trustData.Status)
			}
		})
	}
}

func TestAdvancedSecurityMonitor_VerificationStatusCalculation(t *testing.T) {
	tests := []struct {
		name             string
		verificationRate float64
		expectedStatus   VerificationStatus
	}{
		{"excellent verification rate", 98.5, VerificationStatusExcellent},
		{"good verification rate", 92.0, VerificationStatusGood},
		{"warning verification rate", 85.0, VerificationStatusWarning},
		{"critical verification rate", 75.0, VerificationStatusCritical},
		{"boundary excellent", 95.0, VerificationStatusExcellent},
		{"boundary good", 90.0, VerificationStatusGood},
		{"boundary warning", 80.0, VerificationStatusWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := observability.NewLogger(zap.NewNop())
			tracer := trace.NewNoopTracerProvider().Tracer("test")
			config := &AdvancedSecurityMonitorConfig{
				TrustRateTarget:            100.0,
				TrustRateAlertThreshold:    95.0,
				TrustRateCheckInterval:     30 * time.Second,
				TrustRateHistorySize:       100,
				VerificationRateTarget:     95.0,
				VerificationAlertThreshold: 90.0,
				VerificationCheckInterval:  30 * time.Second,
				VerificationHistorySize:    100,
				ViolationDetectionEnabled:  true,
				ViolationCheckInterval:     1 * time.Minute,
				ViolationHistorySize:       100,
				ConfidenceIntegrityEnabled: true,
				ConfidenceCheckInterval:    1 * time.Minute,
				ConfidenceHistorySize:      100,
				AlertingEnabled:            true,
				AlertCooldown:              5 * time.Minute,
				AlertHistorySize:           100,
			}

			monitor := NewAdvancedSecurityMonitor(config, logger, tracer)
			defer monitor.Shutdown()

			// Record verification attempts to achieve the desired verification rate
			totalAttempts := 100
			successfulAttempts := int(tt.verificationRate * float64(totalAttempts) / 100.0)

			for i := 0; i < successfulAttempts; i++ {
				monitor.RecordWebsiteVerification("test.com", true)
			}

			for i := 0; i < totalAttempts-successfulAttempts; i++ {
				monitor.RecordWebsiteVerification("test.com", false)
			}

			// Check the verification status
			verificationRates := monitor.GetVerificationRates()
			verificationData, exists := verificationRates["test.com"]
			if !exists {
				t.Fatal("Expected verification data for test.com")
			}

			if verificationData.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, verificationData.Status)
			}
		})
	}
}

func TestAdvancedSecurityMonitor_AlertSeverityCalculation(t *testing.T) {
	tests := []struct {
		name             string
		rate             float64
		threshold        float64
		expectedSeverity AlertSeverity
	}{
		{"critical severity", 40.0, 95.0, AlertSeverityCritical}, // 40% < 47.5% (50% of threshold)
		{"error severity", 70.0, 95.0, AlertSeverityError},       // 70% < 76% (80% of threshold)
		{"warning severity", 90.0, 95.0, AlertSeverityWarning},   // 90% < 95% (threshold)
		{"info severity", 98.0, 95.0, AlertSeverityInfo},         // 98% >= 95% (threshold)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger := observability.NewLogger(zap.NewNop())
			tracer := trace.NewNoopTracerProvider().Tracer("test")
			config := &AdvancedSecurityMonitorConfig{
				TrustRateTarget:            100.0,
				TrustRateAlertThreshold:    95.0,
				TrustRateCheckInterval:     30 * time.Second,
				TrustRateHistorySize:       100,
				VerificationRateTarget:     95.0,
				VerificationAlertThreshold: 90.0,
				VerificationCheckInterval:  30 * time.Second,
				VerificationHistorySize:    100,
				ViolationDetectionEnabled:  true,
				ViolationCheckInterval:     1 * time.Minute,
				ViolationHistorySize:       100,
				ConfidenceIntegrityEnabled: true,
				ConfidenceCheckInterval:    1 * time.Minute,
				ConfidenceHistorySize:      100,
				AlertingEnabled:            true,
				AlertCooldown:              5 * time.Minute,
				AlertHistorySize:           100,
			}

			monitor := NewAdvancedSecurityMonitor(config, logger, tracer)
			defer monitor.Shutdown()

			// Test alert severity calculation
			severity := monitor.calculateAlertSeverity(tt.rate, tt.threshold)
			if severity != tt.expectedSeverity {
				t.Errorf("Expected severity %s, got %s", tt.expectedSeverity, severity)
			}
		})
	}
}

func TestAdvancedSecurityMonitor_ConcurrentAccess(t *testing.T) {
	// Setup
	logger := observability.NewLogger(zap.NewNop())
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	config := &AdvancedSecurityMonitorConfig{
		TrustRateTarget:            100.0,
		TrustRateAlertThreshold:    95.0,
		TrustRateCheckInterval:     30 * time.Second,
		TrustRateHistorySize:       100,
		VerificationRateTarget:     95.0,
		VerificationAlertThreshold: 90.0,
		VerificationCheckInterval:  30 * time.Second,
		VerificationHistorySize:    100,
		ViolationDetectionEnabled:  true,
		ViolationCheckInterval:     1 * time.Minute,
		ViolationHistorySize:       100,
		ConfidenceIntegrityEnabled: true,
		ConfidenceCheckInterval:    1 * time.Minute,
		ConfidenceHistorySize:      100,
		AlertingEnabled:            true,
		AlertCooldown:              5 * time.Minute,
		AlertHistorySize:           100,
	}

	monitor := NewAdvancedSecurityMonitor(config, logger, tracer)
	defer monitor.Shutdown()

	// Test concurrent access
	done := make(chan bool, 10)

	// Start multiple goroutines to record data concurrently
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			// Record data source requests
			for j := 0; j < 10; j++ {
				monitor.RecordDataSourceRequest(fmt.Sprintf("source_%d", id), fmt.Sprintf("Source %d", id), j%2 == 0)
			}

			// Record website verifications
			for j := 0; j < 10; j++ {
				monitor.RecordWebsiteVerification(fmt.Sprintf("domain_%d.com", id), j%3 != 0)
			}

			// Record security violations
			monitor.RecordSecurityViolation(ViolationTypeUntrustedDataSource, ViolationSeverityMedium, fmt.Sprintf("source_%d", id), "Test violation", map[string]interface{}{})

			// Record confidence integrity events
			monitor.RecordConfidenceIntegrityEvent(ConfidenceEventTypeAnomaly, ConfidenceEventSeverityLow, fmt.Sprintf("class_%d", id), 0.8, 0.9, map[string]interface{}{})
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify data integrity
	trustRates := monitor.GetTrustRates()
	if len(trustRates) != 10 {
		t.Errorf("Expected 10 trust rate entries, got %d", len(trustRates))
	}

	verificationRates := monitor.GetVerificationRates()
	if len(verificationRates) != 10 {
		t.Errorf("Expected 10 verification rate entries, got %d", len(verificationRates))
	}

	violationEvents := monitor.GetViolationEvents()
	if len(violationEvents) != 10 {
		t.Errorf("Expected 10 violation events, got %d", len(violationEvents))
	}

	confidenceEvents := monitor.GetConfidenceEvents()
	if len(confidenceEvents) != 10 {
		t.Errorf("Expected 10 confidence events, got %d", len(confidenceEvents))
	}
}
