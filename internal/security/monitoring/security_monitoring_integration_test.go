package monitoring

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func TestSecurityMonitoringIntegration_ValidateClassificationSecurity(t *testing.T) {
	// Setup
	logger := observability.NewLogger(zap.NewNop())
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	// Create advanced security monitor
	advancedConfig := &AdvancedSecurityMonitorConfig{
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
	advancedMonitor := NewAdvancedSecurityMonitor(advancedConfig, logger, tracer)
	defer advancedMonitor.Shutdown()

	// Create security monitor
	securityConfig := &SecurityMonitorConfig{
		MaxEvents:      10000,
		EventRetention: 30 * 24 * time.Hour,
		AlertThresholds: map[SecurityEventSeverity]int{
			SeverityCritical: 1,
			SeverityHigh:     5,
			SeverityMedium:   10,
			SeverityLow:      50,
		},
		AlertCooldown:   5 * time.Minute,
		MetricsInterval: 1 * time.Minute,
		WebhookTimeout:  10 * time.Second,
	}
	securityMonitor := NewSecurityMonitor(securityConfig, zap.NewNop())
	defer securityMonitor.Stop()

	// Create integration config
	integrationConfig := &SecurityMonitoringIntegrationConfig{
		EnableAdvancedMonitoring:    true,
		EnableSecurityEventLogging:  true,
		EnablePerformanceMonitoring: true,
		IntegrationCheckInterval:    30 * time.Second,

		TrustedDataSources:       []string{"supabase", "government_apis", "verified_sources"},
		UntrustedDataSources:     []string{"unverified_apis", "suspicious_sources"},
		DataSourceTrustThreshold: 95.0,

		WebsiteVerificationEnabled: true,
		WebsiteVerificationTimeout: 10 * time.Second,
		WebsiteVerificationRetries: 3,

		ConfidenceScoreValidationEnabled: true,
		ConfidenceScoreThreshold:         0.8,
		ConfidenceScoreAnomalyThreshold:  0.1,

		EnableRealTimeAlerts:     true,
		AlertProcessingInterval:  1 * time.Minute,
		AlertEscalationThreshold: 5,

		WebhookTimeout: 10 * time.Second,
	}

	// Create integration
	integration := NewSecurityMonitoringIntegration(
		integrationConfig,
		advancedMonitor,
		securityMonitor,
		logger,
		tracer,
	)
	defer integration.Shutdown()
	defer advancedMonitor.Shutdown()
	defer securityMonitor.Stop()

	// Test cases
	tests := []struct {
		name               string
		securityContext    *ClassificationSecurityContext
		expectedValid      bool
		expectedScore      float64
		expectedViolations int
		expectedWarnings   int
	}{
		{
			name: "valid classification with trusted sources",
			securityContext: &ClassificationSecurityContext{
				RequestID:             "req_001",
				UserID:                "user_123",
				IPAddress:             "192.168.1.1",
				UserAgent:             "Mozilla/5.0",
				BusinessName:          "Test Business",
				BusinessDescription:   "A test business for classification",
				WebsiteURL:            "https://example.com",
				DataSources:           []string{"supabase", "government_apis"},
				ClassificationMethods: []string{"keyword_matching", "ml_classification"},
				ConfidenceScores: map[string]float64{
					"keyword_matching":  0.85,
					"ml_classification": 0.90,
				},
				SecurityFlags: map[string]bool{
					"trusted_source": true,
				},
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"test": true,
				},
			},
			expectedValid:      true,
			expectedScore:      90.0, // 100 - 10 (medium violation for untrusted source check)
			expectedViolations: 1,    // One violation for untrusted source validation
			expectedWarnings:   0,
		},
		{
			name: "invalid classification with untrusted source",
			securityContext: &ClassificationSecurityContext{
				RequestID:             "req_002",
				UserID:                "user_456",
				IPAddress:             "192.168.1.2",
				UserAgent:             "Mozilla/5.0",
				BusinessName:          "Suspicious Business",
				BusinessDescription:   "A suspicious business",
				WebsiteURL:            "https://suspicious-site.com",
				DataSources:           []string{"unverified_apis", "supabase"},
				ClassificationMethods: []string{"keyword_matching"},
				ConfidenceScores: map[string]float64{
					"keyword_matching": 0.75,
				},
				SecurityFlags: map[string]bool{
					"trusted_source": false,
				},
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"test": true,
				},
			},
			expectedValid:      false,
			expectedScore:      73.0, // 100 - 15 (high violation) - 2 (warning) - 10 (medium violation)
			expectedViolations: 2,    // Two violations: untrusted source + confidence anomaly
			expectedWarnings:   1,    // One warning for confidence score anomaly
		},
		{
			name: "invalid classification with confidence score anomaly",
			securityContext: &ClassificationSecurityContext{
				RequestID:             "req_003",
				UserID:                "user_789",
				IPAddress:             "192.168.1.3",
				UserAgent:             "Mozilla/5.0",
				BusinessName:          "Anomaly Business",
				BusinessDescription:   "A business with confidence score anomaly",
				WebsiteURL:            "https://example.com",
				DataSources:           []string{"supabase"},
				ClassificationMethods: []string{"keyword_matching"},
				ConfidenceScores: map[string]float64{
					"keyword_matching": 0.99, // Anomalously high
				},
				SecurityFlags: map[string]bool{
					"trusted_source": true,
				},
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"test": true,
				},
			},
			expectedValid:      true, // Not invalid, just has anomaly
			expectedScore:      90.0, // 100 - 10 (medium severity violation)
			expectedViolations: 1,
			expectedWarnings:   0,
		},
		{
			name: "invalid classification with out of range confidence score",
			securityContext: &ClassificationSecurityContext{
				RequestID:             "req_004",
				UserID:                "user_101",
				IPAddress:             "192.168.1.4",
				UserAgent:             "Mozilla/5.0",
				BusinessName:          "Invalid Score Business",
				BusinessDescription:   "A business with invalid confidence score",
				WebsiteURL:            "https://example.com",
				DataSources:           []string{"supabase"},
				ClassificationMethods: []string{"keyword_matching"},
				ConfidenceScores: map[string]float64{
					"keyword_matching": 1.5, // Out of range
				},
				SecurityFlags: map[string]bool{
					"trusted_source": true,
				},
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"test": true,
				},
			},
			expectedValid:      false,
			expectedScore:      75.0, // 100 - 25 (high severity violation)
			expectedViolations: 1,
			expectedWarnings:   0,
		},
		{
			name: "classification with unknown data source warning",
			securityContext: &ClassificationSecurityContext{
				RequestID:             "req_005",
				UserID:                "user_102",
				IPAddress:             "192.168.1.5",
				UserAgent:             "Mozilla/5.0",
				BusinessName:          "Unknown Source Business",
				BusinessDescription:   "A business with unknown data source",
				WebsiteURL:            "https://example.com",
				DataSources:           []string{"unknown_source"},
				ClassificationMethods: []string{"keyword_matching"},
				ConfidenceScores: map[string]float64{
					"keyword_matching": 0.80,
				},
				SecurityFlags: map[string]bool{
					"trusted_source": true,
				},
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"test": true,
				},
			},
			expectedValid:      true,
			expectedScore:      98.0, // 100 - 2 (warning)
			expectedViolations: 0,
			expectedWarnings:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			result, err := integration.ValidateClassificationSecurity(ctx, tt.securityContext)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if result.Valid != tt.expectedValid {
				t.Errorf("Expected valid %v, got %v", tt.expectedValid, result.Valid)
			}

			// Allow for small floating point differences
			if math.Abs(result.SecurityScore-tt.expectedScore) > 0.01 {
				t.Errorf("Expected security score %.2f, got %.2f", tt.expectedScore, result.SecurityScore)
			}

			if len(result.Violations) != tt.expectedViolations {
				t.Errorf("Expected %d violations, got %d", tt.expectedViolations, len(result.Violations))
			}

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
			}

			// Verify that events were recorded in the monitors
			time.Sleep(100 * time.Millisecond) // Allow time for async processing

			// Check advanced monitor
			trustRates := advancedMonitor.GetTrustRates()
			if len(trustRates) == 0 {
				t.Error("Expected trust rates to be recorded in advanced monitor")
			}

			violationEvents := advancedMonitor.GetViolationEvents()
			if len(violationEvents) != tt.expectedViolations {
				t.Errorf("Expected %d violation events in advanced monitor, got %d", tt.expectedViolations, len(violationEvents))
			}

			// Check security monitor
			events, err := securityMonitor.GetEvents(EventFilters{})
			if err != nil {
				t.Errorf("Expected no error getting events, got %v", err)
			}
			if len(events) == 0 {
				t.Error("Expected events to be recorded in security monitor")
			}
		})
	}
}

func TestSecurityMonitoringIntegration_DataSourceValidation(t *testing.T) {
	// Setup
	logger := observability.NewLogger(zap.NewNop())
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	advancedConfig := &AdvancedSecurityMonitorConfig{
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
	advancedMonitor := NewAdvancedSecurityMonitor(advancedConfig, logger, tracer)
	defer advancedMonitor.Shutdown()

	securityConfig := &SecurityMonitorConfig{
		MaxEvents:      10000,
		EventRetention: 30 * 24 * time.Hour,
		AlertThresholds: map[SecurityEventSeverity]int{
			SeverityCritical: 1,
			SeverityHigh:     5,
			SeverityMedium:   10,
			SeverityLow:      50,
		},
		AlertCooldown:   5 * time.Minute,
		MetricsInterval: 1 * time.Minute,
		WebhookTimeout:  10 * time.Second,
	}
	securityMonitor := NewSecurityMonitor(securityConfig, zap.NewNop())
	defer securityMonitor.Stop()

	// Performance monitor removed for now - focusing on security monitoring

	integrationConfig := &SecurityMonitoringIntegrationConfig{
		EnableAdvancedMonitoring:    true,
		EnableSecurityEventLogging:  true,
		EnablePerformanceMonitoring: true,
		IntegrationCheckInterval:    30 * time.Second,

		TrustedDataSources:       []string{"supabase", "government_apis", "verified_sources"},
		UntrustedDataSources:     []string{"unverified_apis", "suspicious_sources"},
		DataSourceTrustThreshold: 95.0,

		WebsiteVerificationEnabled: true,
		WebsiteVerificationTimeout: 10 * time.Second,
		WebsiteVerificationRetries: 3,

		ConfidenceScoreValidationEnabled: true,
		ConfidenceScoreThreshold:         0.8,
		ConfidenceScoreAnomalyThreshold:  0.1,

		EnableRealTimeAlerts:     true,
		AlertProcessingInterval:  1 * time.Minute,
		AlertEscalationThreshold: 5,

		WebhookTimeout: 10 * time.Second,
	}

	integration := NewSecurityMonitoringIntegration(
		integrationConfig,
		advancedMonitor,
		securityMonitor,
		logger,
		tracer,
	)
	defer integration.Shutdown()

	// Test data source validation
	tests := []struct {
		name               string
		dataSources        []string
		expectedValid      bool
		expectedViolations int
		expectedWarnings   int
	}{
		{
			name:               "all trusted sources",
			dataSources:        []string{"supabase", "government_apis"},
			expectedValid:      true,
			expectedViolations: 0,
			expectedWarnings:   0,
		},
		{
			name:               "mixed trusted and untrusted sources",
			dataSources:        []string{"supabase", "unverified_apis"},
			expectedValid:      false,
			expectedViolations: 1,
			expectedWarnings:   0,
		},
		{
			name:               "all untrusted sources",
			dataSources:        []string{"unverified_apis", "suspicious_sources"},
			expectedValid:      false,
			expectedViolations: 2,
			expectedWarnings:   0,
		},
		{
			name:               "unknown sources",
			dataSources:        []string{"unknown_source1", "unknown_source2"},
			expectedValid:      true,
			expectedViolations: 0,
			expectedWarnings:   2,
		},
		{
			name:               "mixed trusted and unknown sources",
			dataSources:        []string{"supabase", "unknown_source"},
			expectedValid:      true,
			expectedViolations: 0,
			expectedWarnings:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			securityContext := &ClassificationSecurityContext{
				RequestID:             fmt.Sprintf("req_%s", tt.name),
				UserID:                "user_123",
				IPAddress:             "192.168.1.1",
				UserAgent:             "Mozilla/5.0",
				BusinessName:          "Test Business",
				BusinessDescription:   "A test business",
				WebsiteURL:            "https://example.com",
				DataSources:           tt.dataSources,
				ClassificationMethods: []string{"keyword_matching"},
				ConfidenceScores: map[string]float64{
					"keyword_matching": 0.80,
				},
				SecurityFlags: map[string]bool{},
				Timestamp:     time.Now(),
				Metadata:      map[string]interface{}{},
			}

			ctx := context.Background()
			result, err := integration.ValidateClassificationSecurity(ctx, securityContext)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if result.Valid != tt.expectedValid {
				t.Errorf("Expected valid %v, got %v", tt.expectedValid, result.Valid)
			}

			if len(result.Violations) != tt.expectedViolations {
				t.Errorf("Expected %d violations, got %d", tt.expectedViolations, len(result.Violations))
			}

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
			}
		})
	}
}

func TestSecurityMonitoringIntegration_WebsiteVerification(t *testing.T) {
	// Setup
	logger := observability.NewLogger(zap.NewNop())
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	advancedConfig := &AdvancedSecurityMonitorConfig{
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
	advancedMonitor := NewAdvancedSecurityMonitor(advancedConfig, logger, tracer)
	defer advancedMonitor.Shutdown()

	securityConfig := &SecurityMonitorConfig{
		MaxEvents:      10000,
		EventRetention: 30 * 24 * time.Hour,
		AlertThresholds: map[SecurityEventSeverity]int{
			SeverityCritical: 1,
			SeverityHigh:     5,
			SeverityMedium:   10,
			SeverityLow:      50,
		},
		AlertCooldown:   5 * time.Minute,
		MetricsInterval: 1 * time.Minute,
		WebhookTimeout:  10 * time.Second,
	}
	securityMonitor := NewSecurityMonitor(securityConfig, zap.NewNop())
	defer securityMonitor.Stop()

	// Performance monitor removed for now - focusing on security monitoring

	integrationConfig := &SecurityMonitoringIntegrationConfig{
		EnableAdvancedMonitoring:    true,
		EnableSecurityEventLogging:  true,
		EnablePerformanceMonitoring: true,
		IntegrationCheckInterval:    30 * time.Second,

		TrustedDataSources:       []string{"supabase"},
		UntrustedDataSources:     []string{},
		DataSourceTrustThreshold: 95.0,

		WebsiteVerificationEnabled: true,
		WebsiteVerificationTimeout: 10 * time.Second,
		WebsiteVerificationRetries: 3,

		ConfidenceScoreValidationEnabled: true,
		ConfidenceScoreThreshold:         0.8,
		ConfidenceScoreAnomalyThreshold:  0.1,

		EnableRealTimeAlerts:     true,
		AlertProcessingInterval:  1 * time.Minute,
		AlertEscalationThreshold: 5,

		WebhookTimeout: 10 * time.Second,
	}

	integration := NewSecurityMonitoringIntegration(
		integrationConfig,
		advancedMonitor,
		securityMonitor,
		logger,
		tracer,
	)
	defer integration.Shutdown()

	// Test website verification
	tests := []struct {
		name               string
		websiteURL         string
		expectedValid      bool
		expectedViolations int
		expectedWarnings   int
	}{
		{
			name:               "verified website",
			websiteURL:         "https://example.com",
			expectedValid:      true,
			expectedViolations: 0,
			expectedWarnings:   0,
		},
		{
			name:               "unverified website",
			websiteURL:         "https://suspicious-site.com",
			expectedValid:      false,
			expectedViolations: 1,
			expectedWarnings:   0,
		},
		{
			name:               "malicious website",
			websiteURL:         "https://malicious-domain.net",
			expectedValid:      false,
			expectedViolations: 1,
			expectedWarnings:   0,
		},
		{
			name:               "fake website",
			websiteURL:         "https://fake-website.org",
			expectedValid:      false,
			expectedViolations: 1,
			expectedWarnings:   0,
		},
		{
			name:               "no website URL",
			websiteURL:         "",
			expectedValid:      true,
			expectedViolations: 0,
			expectedWarnings:   0,
		},
		{
			name:               "invalid website URL",
			websiteURL:         "not-a-valid-url",
			expectedValid:      true,
			expectedViolations: 0,
			expectedWarnings:   1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			securityContext := &ClassificationSecurityContext{
				RequestID:             fmt.Sprintf("req_%s", tt.name),
				UserID:                "user_123",
				IPAddress:             "192.168.1.1",
				UserAgent:             "Mozilla/5.0",
				BusinessName:          "Test Business",
				BusinessDescription:   "A test business",
				WebsiteURL:            tt.websiteURL,
				DataSources:           []string{"supabase"},
				ClassificationMethods: []string{"keyword_matching"},
				ConfidenceScores: map[string]float64{
					"keyword_matching": 0.80,
				},
				SecurityFlags: map[string]bool{},
				Timestamp:     time.Now(),
				Metadata:      map[string]interface{}{},
			}

			ctx := context.Background()
			result, err := integration.ValidateClassificationSecurity(ctx, securityContext)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if result.Valid != tt.expectedValid {
				t.Errorf("Expected valid %v, got %v", tt.expectedValid, result.Valid)
			}

			if len(result.Violations) != tt.expectedViolations {
				t.Errorf("Expected %d violations, got %d", tt.expectedViolations, len(result.Violations))
			}

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
			}
		})
	}
}

func TestSecurityMonitoringIntegration_ConfidenceScoreValidation(t *testing.T) {
	// Setup
	logger := observability.NewLogger(zap.NewNop())
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	advancedConfig := &AdvancedSecurityMonitorConfig{
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
	advancedMonitor := NewAdvancedSecurityMonitor(advancedConfig, logger, tracer)
	defer advancedMonitor.Shutdown()

	securityConfig := &SecurityMonitorConfig{
		MaxEvents:      10000,
		EventRetention: 30 * 24 * time.Hour,
		AlertThresholds: map[SecurityEventSeverity]int{
			SeverityCritical: 1,
			SeverityHigh:     5,
			SeverityMedium:   10,
			SeverityLow:      50,
		},
		AlertCooldown:   5 * time.Minute,
		MetricsInterval: 1 * time.Minute,
		WebhookTimeout:  10 * time.Second,
	}
	securityMonitor := NewSecurityMonitor(securityConfig, zap.NewNop())
	defer securityMonitor.Stop()

	// Performance monitor removed for now - focusing on security monitoring

	integrationConfig := &SecurityMonitoringIntegrationConfig{
		EnableAdvancedMonitoring:    true,
		EnableSecurityEventLogging:  true,
		EnablePerformanceMonitoring: true,
		IntegrationCheckInterval:    30 * time.Second,

		TrustedDataSources:       []string{"supabase"},
		UntrustedDataSources:     []string{},
		DataSourceTrustThreshold: 95.0,

		WebsiteVerificationEnabled: true,
		WebsiteVerificationTimeout: 10 * time.Second,
		WebsiteVerificationRetries: 3,

		ConfidenceScoreValidationEnabled: true,
		ConfidenceScoreThreshold:         0.8,
		ConfidenceScoreAnomalyThreshold:  0.1,

		EnableRealTimeAlerts:     true,
		AlertProcessingInterval:  1 * time.Minute,
		AlertEscalationThreshold: 5,

		WebhookTimeout: 10 * time.Second,
	}

	integration := NewSecurityMonitoringIntegration(
		integrationConfig,
		advancedMonitor,
		securityMonitor,
		logger,
		tracer,
	)
	defer integration.Shutdown()

	// Test confidence score validation
	tests := []struct {
		name               string
		confidenceScores   map[string]float64
		expectedValid      bool
		expectedViolations int
		expectedWarnings   int
	}{
		{
			name: "valid confidence scores",
			confidenceScores: map[string]float64{
				"keyword_matching":  0.85,
				"ml_classification": 0.90,
			},
			expectedValid:      true,
			expectedViolations: 0,
			expectedWarnings:   0,
		},
		{
			name: "out of range confidence score (too high)",
			confidenceScores: map[string]float64{
				"keyword_matching": 1.5,
			},
			expectedValid:      false,
			expectedViolations: 1,
			expectedWarnings:   0,
		},
		{
			name: "out of range confidence score (negative)",
			confidenceScores: map[string]float64{
				"keyword_matching": -0.1,
			},
			expectedValid:      false,
			expectedViolations: 1,
			expectedWarnings:   0,
		},
		{
			name: "confidence score anomaly",
			confidenceScores: map[string]float64{
				"keyword_matching": 0.99, // Anomalously high
			},
			expectedValid:      true, // Not invalid, just has anomaly
			expectedViolations: 1,
			expectedWarnings:   0,
		},
		{
			name: "multiple confidence score issues",
			confidenceScores: map[string]float64{
				"keyword_matching":  1.2,  // Out of range
				"ml_classification": 0.99, // Anomaly
			},
			expectedValid:      false,
			expectedViolations: 2,
			expectedWarnings:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			securityContext := &ClassificationSecurityContext{
				RequestID:             fmt.Sprintf("req_%s", tt.name),
				UserID:                "user_123",
				IPAddress:             "192.168.1.1",
				UserAgent:             "Mozilla/5.0",
				BusinessName:          "Test Business",
				BusinessDescription:   "A test business",
				WebsiteURL:            "https://example.com",
				DataSources:           []string{"supabase"},
				ClassificationMethods: []string{"keyword_matching"},
				ConfidenceScores:      tt.confidenceScores,
				SecurityFlags:         map[string]bool{},
				Timestamp:             time.Now(),
				Metadata:              map[string]interface{}{},
			}

			ctx := context.Background()
			result, err := integration.ValidateClassificationSecurity(ctx, securityContext)
			if err != nil {
				t.Fatalf("Expected no error, got %v", err)
			}

			if result.Valid != tt.expectedValid {
				t.Errorf("Expected valid %v, got %v", tt.expectedValid, result.Valid)
			}

			if len(result.Violations) != tt.expectedViolations {
				t.Errorf("Expected %d violations, got %d", tt.expectedViolations, len(result.Violations))
			}

			if len(result.Warnings) != tt.expectedWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectedWarnings, len(result.Warnings))
			}
		})
	}
}

// Helper function for absolute value
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
