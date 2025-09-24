package classification

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
)

func TestSecurityValidationMonitor_NewAdvancedSecurityValidationMonitor(t *testing.T) {
	tests := []struct {
		name   string
		config *SecurityValidationConfig
	}{
		{
			name:   "with default config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &SecurityValidationConfig{
				Enabled:                      true,
				CollectionInterval:           10 * time.Second,
				SlowValidationThreshold:      100 * time.Millisecond,
				MaxValidationStats:           500,
				AlertingEnabled:              true,
				TrackDataSourceValidation:    true,
				TrackWebsiteVerification:     true,
				TrackTrustScoreValidation:    true,
				TrackSecurityChecks:          true,
				TrackComplianceValidation:    true,
				TrackEncryptionValidation:    true,
				TrackAccessControlValidation: true,
				TrackAuditLogValidation:      true,
				TrackThreatDetection:         true,
				TrackVulnerabilityScanning:   true,
				TrackSecurityMetrics:         true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)

			monitor := NewAdvancedSecurityValidationMonitor(logger, tt.config)

			if monitor == nil {
				t.Fatal("Expected monitor to be created, got nil")
			}

			if monitor.logger != logger {
				t.Error("Expected monitor to use the provided logger")
			}

			if tt.config == nil {
				// Check default config
				if !monitor.config.Enabled {
					t.Error("Expected default config to have monitoring enabled")
				}
				if monitor.config.CollectionInterval != 30*time.Second {
					t.Error("Expected default collection interval to be 30 seconds")
				}
				if monitor.config.SlowValidationThreshold != 200*time.Millisecond {
					t.Error("Expected default slow validation threshold to be 200ms")
				}
			} else {
				// Check custom config
				if monitor.config.Enabled != tt.config.Enabled {
					t.Error("Expected monitor to use custom enabled setting")
				}
				if monitor.config.CollectionInterval != tt.config.CollectionInterval {
					t.Error("Expected monitor to use custom collection interval")
				}
				if monitor.config.SlowValidationThreshold != tt.config.SlowValidationThreshold {
					t.Error("Expected monitor to use custom slow validation threshold")
				}
			}

			// Clean up
			monitor.Stop()
		})
	}
}

func TestSecurityValidationMonitor_StartStop(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SecurityValidationConfig{
		Enabled:            true,
		CollectionInterval: 100 * time.Millisecond,
	}

	monitor := NewAdvancedSecurityValidationMonitor(logger, config)

	// Test starting
	monitor.Start()
	time.Sleep(50 * time.Millisecond) // Give it a moment to start

	// Test stopping
	monitor.Stop()
	time.Sleep(50 * time.Millisecond) // Give it a moment to stop

	// Verify that the monitor stopped without panicking
	// Further checks could involve inspecting logs or mock channels if they were used
}

func TestSecurityValidationMonitor_RecordSecurityValidation(t *testing.T) {
	tests := []struct {
		name           string
		result         *AdvancedSecurityValidationResult
		expectedStats  int
		expectedAlerts int
	}{
		{
			name: "successful validation",
			result: &AdvancedSecurityValidationResult{
				ValidationID:               "test_validation_1",
				ValidationType:             "data_source_validation",
				ValidationName:             "test_data_source",
				ExecutionTime:              50 * time.Millisecond,
				Success:                    true,
				Error:                      nil,
				SecurityViolation:          false,
				ComplianceViolation:        false,
				ThreatDetected:             false,
				VulnerabilityFound:         false,
				TrustScore:                 0.95,
				ConfidenceLevel:            0.90,
				RiskLevel:                  "low",
				SecurityRecommendations:    []string{},
				PerformanceRecommendations: []string{},
				Metadata:                   make(map[string]interface{}),
				Timestamp:                  time.Now(),
			},
			expectedStats:  1,
			expectedAlerts: 0,
		},
		{
			name: "slow validation",
			result: &AdvancedSecurityValidationResult{
				ValidationID:               "test_validation_2",
				ValidationType:             "website_verification",
				ValidationName:             "test_website",
				ExecutionTime:              300 * time.Millisecond, // Exceeds threshold
				Success:                    true,
				Error:                      nil,
				SecurityViolation:          false,
				ComplianceViolation:        false,
				ThreatDetected:             false,
				VulnerabilityFound:         false,
				TrustScore:                 0.85,
				ConfidenceLevel:            0.80,
				RiskLevel:                  "medium",
				SecurityRecommendations:    []string{},
				PerformanceRecommendations: []string{},
				Metadata:                   make(map[string]interface{}),
				Timestamp:                  time.Now(),
			},
			expectedStats:  1,
			expectedAlerts: 1, // Should trigger slow validation alert
		},
		{
			name: "validation with security violation",
			result: &AdvancedSecurityValidationResult{
				ValidationID:               "test_validation_3",
				ValidationType:             "security_check",
				ValidationName:             "test_security",
				ExecutionTime:              100 * time.Millisecond,
				Success:                    false,
				Error:                      errors.New("security violation detected"),
				SecurityViolation:          true,
				ComplianceViolation:        false,
				ThreatDetected:             false,
				VulnerabilityFound:         false,
				TrustScore:                 0.20,
				ConfidenceLevel:            0.30,
				RiskLevel:                  "critical",
				SecurityRecommendations:    []string{"Review security policies"},
				PerformanceRecommendations: []string{},
				Metadata:                   make(map[string]interface{}),
				Timestamp:                  time.Now(),
			},
			expectedStats:  1,
			expectedAlerts: 1, // Should trigger security violation alert
		},
		{
			name: "validation with compliance violation",
			result: &AdvancedSecurityValidationResult{
				ValidationID:               "test_validation_4",
				ValidationType:             "compliance_validation",
				ValidationName:             "test_compliance",
				ExecutionTime:              150 * time.Millisecond,
				Success:                    false,
				Error:                      errors.New("compliance violation detected"),
				SecurityViolation:          false,
				ComplianceViolation:        true,
				ThreatDetected:             false,
				VulnerabilityFound:         false,
				TrustScore:                 0.40,
				ConfidenceLevel:            0.50,
				RiskLevel:                  "high",
				SecurityRecommendations:    []string{"Update compliance procedures"},
				PerformanceRecommendations: []string{},
				Metadata:                   make(map[string]interface{}),
				Timestamp:                  time.Now(),
			},
			expectedStats:  1,
			expectedAlerts: 1, // Should trigger compliance violation alert
		},
		{
			name: "validation with threat detected",
			result: &AdvancedSecurityValidationResult{
				ValidationID:               "test_validation_5",
				ValidationType:             "threat_detection",
				ValidationName:             "test_threat",
				ExecutionTime:              80 * time.Millisecond,
				Success:                    false,
				Error:                      errors.New("threat detected"),
				SecurityViolation:          false,
				ComplianceViolation:        false,
				ThreatDetected:             true,
				VulnerabilityFound:         false,
				TrustScore:                 0.10,
				ConfidenceLevel:            0.20,
				RiskLevel:                  "critical",
				SecurityRecommendations:    []string{"Implement threat response"},
				PerformanceRecommendations: []string{},
				Metadata:                   make(map[string]interface{}),
				Timestamp:                  time.Now(),
			},
			expectedStats:  1,
			expectedAlerts: 1, // Should trigger threat detected alert
		},
		{
			name: "validation with vulnerability found",
			result: &AdvancedSecurityValidationResult{
				ValidationID:               "test_validation_6",
				ValidationType:             "vulnerability_scanning",
				ValidationName:             "test_vulnerability",
				ExecutionTime:              120 * time.Millisecond,
				Success:                    false,
				Error:                      errors.New("vulnerability found"),
				SecurityViolation:          false,
				ComplianceViolation:        false,
				ThreatDetected:             false,
				VulnerabilityFound:         true,
				TrustScore:                 0.30,
				ConfidenceLevel:            0.40,
				RiskLevel:                  "high",
				SecurityRecommendations:    []string{"Patch vulnerability"},
				PerformanceRecommendations: []string{},
				Metadata:                   make(map[string]interface{}),
				Timestamp:                  time.Now(),
			},
			expectedStats:  1,
			expectedAlerts: 1, // Should trigger vulnerability found alert
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			config := &SecurityValidationConfig{
				Enabled:                 true,
				CollectionInterval:      30 * time.Second,
				SlowValidationThreshold: 200 * time.Millisecond,
				MaxValidationStats:      1000,
				AlertingEnabled:         true,
			}

			monitor := NewAdvancedSecurityValidationMonitor(logger, config)
			defer monitor.Stop()

			// Record multiple executions for the same validation to test statistics
			executions := 3
			for i := 0; i < executions; i++ {
				monitor.RecordSecurityValidation(context.Background(), tt.result)
			}

			// Get validation stats
			stats := monitor.GetValidationStats(10)
			if len(stats) != tt.expectedStats {
				t.Errorf("Expected %d validation stats, got %d", tt.expectedStats, len(stats))
			}

			// Verify the stats for the recorded validation
			for _, stat := range stats {
				if stat.ValidationID != tt.result.ValidationID {
					continue
				}

				if stat.ExecutionCount != int64(executions) {
					t.Errorf("Expected execution count %d, got %d", executions, stat.ExecutionCount)
				}

				if stat.AverageExecutionTime != float64(tt.result.ExecutionTime.Milliseconds()) {
					t.Errorf("Expected average execution time %.2f, got %.2f",
						float64(tt.result.ExecutionTime.Milliseconds()), stat.AverageExecutionTime)
				}

				if tt.result.Success && stat.SuccessCount != int64(executions) {
					t.Errorf("Expected success count %d, got %d", executions, stat.SuccessCount)
				}

				if !tt.result.Success && stat.FailureCount != int64(executions) {
					t.Errorf("Expected failure count %d, got %d", executions, stat.FailureCount)
				}

				if tt.result.SecurityViolation && stat.SecurityViolationCount != int64(executions) {
					t.Errorf("Expected security violation count %d, got %d", executions, stat.SecurityViolationCount)
				}

				if tt.result.ComplianceViolation && stat.ComplianceViolationCount != int64(executions) {
					t.Errorf("Expected compliance violation count %d, got %d", executions, stat.ComplianceViolationCount)
				}

				if tt.result.ThreatDetected && stat.ThreatDetectionCount != int64(executions) {
					t.Errorf("Expected threat detection count %d, got %d", executions, stat.ThreatDetectionCount)
				}

				if tt.result.VulnerabilityFound && stat.VulnerabilityCount != int64(executions) {
					t.Errorf("Expected vulnerability count %d, got %d", executions, stat.VulnerabilityCount)
				}

				// Check performance category
				if tt.result.ExecutionTime > config.SlowValidationThreshold {
					if stat.PerformanceCategory != "poor" && stat.PerformanceCategory != "critical" {
						t.Errorf("Expected performance category to be poor or critical for slow validation, got %s",
							stat.PerformanceCategory)
					}
				}

				// Check security category
				if tt.result.SecurityViolation || tt.result.ComplianceViolation ||
					tt.result.ThreatDetected || tt.result.VulnerabilityFound {
					if stat.SecurityCategory != "at_risk" {
						t.Errorf("Expected security category to be at_risk for validation with issues, got %s",
							stat.SecurityCategory)
					}
				}
			}

			// Get security alerts
			alerts := monitor.GetSecurityAlerts(false, 10)
			if len(alerts) < tt.expectedAlerts {
				t.Errorf("Expected at least %d security alerts, got %d", tt.expectedAlerts, len(alerts))
			}

			// Check alert types
			alertTypes := make(map[string]bool)
			for _, alert := range alerts {
				alertTypes[alert.AlertType] = true

				if alert.Severity == "" {
					t.Error("Expected alert to have a severity level")
				}

				if alert.Message == "" {
					t.Error("Expected alert to have a message")
				}

				if len(alert.Recommendations) == 0 {
					t.Error("Expected alert to have recommendations")
				}

				if alert.SecurityImpact == "" {
					t.Error("Expected alert to have a security impact")
				}
			}

			// Verify expected alert types are present
			if tt.result.ExecutionTime > config.SlowValidationThreshold && !alertTypes["slow_validation"] {
				t.Error("Expected to have slow_validation alert type")
			}
			if tt.result.SecurityViolation && !alertTypes["security_violation"] {
				t.Error("Expected to have security_violation alert type")
			}
			if tt.result.ComplianceViolation && !alertTypes["compliance_violation"] {
				t.Error("Expected to have compliance_violation alert type")
			}
			if tt.result.ThreatDetected && !alertTypes["threat_detected"] {
				t.Error("Expected to have threat_detected alert type")
			}
			if tt.result.VulnerabilityFound && !alertTypes["vulnerability_found"] {
				t.Error("Expected to have vulnerability_found alert type")
			}
		})
	}
}

func TestSecurityValidationMonitor_GetSecuritySystemHealth(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SecurityValidationConfig{
		Enabled:                 true,
		CollectionInterval:      30 * time.Second,
		SlowValidationThreshold: 200 * time.Millisecond,
		MaxValidationStats:      1000,
		AlertingEnabled:         true,
	}

	monitor := NewAdvancedSecurityValidationMonitor(logger, config)
	defer monitor.Stop()

	// Record various types of validations
	validations := []struct {
		result     *AdvancedSecurityValidationResult
		executions int
	}{
		{
			result: &AdvancedSecurityValidationResult{
				ValidationID:               "fast_validation",
				ValidationType:             "data_source_validation",
				ValidationName:             "fast_data_source",
				ExecutionTime:              50 * time.Millisecond,
				Success:                    true,
				Error:                      nil,
				SecurityViolation:          false,
				ComplianceViolation:        false,
				ThreatDetected:             false,
				VulnerabilityFound:         false,
				TrustScore:                 0.95,
				ConfidenceLevel:            0.90,
				RiskLevel:                  "low",
				SecurityRecommendations:    []string{},
				PerformanceRecommendations: []string{},
				Metadata:                   make(map[string]interface{}),
				Timestamp:                  time.Now(),
			},
			executions: 5,
		},
		{
			result: &AdvancedSecurityValidationResult{
				ValidationID:               "slow_validation",
				ValidationType:             "website_verification",
				ValidationName:             "slow_website",
				ExecutionTime:              300 * time.Millisecond,
				Success:                    true,
				Error:                      nil,
				SecurityViolation:          false,
				ComplianceViolation:        false,
				ThreatDetected:             false,
				VulnerabilityFound:         false,
				TrustScore:                 0.85,
				ConfidenceLevel:            0.80,
				RiskLevel:                  "medium",
				SecurityRecommendations:    []string{},
				PerformanceRecommendations: []string{},
				Metadata:                   make(map[string]interface{}),
				Timestamp:                  time.Now(),
			},
			executions: 3,
		},
		{
			result: &AdvancedSecurityValidationResult{
				ValidationID:               "security_violation_validation",
				ValidationType:             "security_check",
				ValidationName:             "security_violation",
				ExecutionTime:              100 * time.Millisecond,
				Success:                    false,
				Error:                      errors.New("security violation"),
				SecurityViolation:          true,
				ComplianceViolation:        false,
				ThreatDetected:             false,
				VulnerabilityFound:         false,
				TrustScore:                 0.20,
				ConfidenceLevel:            0.30,
				RiskLevel:                  "critical",
				SecurityRecommendations:    []string{"Review security policies"},
				PerformanceRecommendations: []string{},
				Metadata:                   make(map[string]interface{}),
				Timestamp:                  time.Now(),
			},
			executions: 2,
		},
	}

	for _, v := range validations {
		for i := 0; i < v.executions; i++ {
			monitor.RecordSecurityValidation(context.Background(), v.result)
		}
	}

	// Get security system health
	health := monitor.GetSecuritySystemHealth()
	if health == nil {
		t.Fatal("Expected security system health, got nil")
	}

	// Check health structure
	if health.Timestamp.IsZero() {
		t.Error("Expected health to have a timestamp")
	}

	if health.ValidationCount != 3 {
		t.Errorf("Expected validation count 3, got %d", health.ValidationCount)
	}

	if health.SecurityViolations != 2 {
		t.Errorf("Expected security violations 2, got %d", health.SecurityViolations)
	}

	if health.SlowValidations != 1 {
		t.Errorf("Expected slow validations 1, got %d", health.SlowValidations)
	}

	if health.FailedValidations != 2 {
		t.Errorf("Expected failed validations 2, got %d", health.FailedValidations)
	}

	if health.OverallSecurityScore < 0 || health.OverallSecurityScore > 100 {
		t.Errorf("Expected overall security score between 0 and 100, got %.2f", health.OverallSecurityScore)
	}

	if health.OverallPerformanceScore < 0 || health.OverallPerformanceScore > 100 {
		t.Errorf("Expected overall performance score between 0 and 100, got %.2f", health.OverallPerformanceScore)
	}

	if health.OverallRiskLevel == "" {
		t.Error("Expected overall risk level to be set")
	}

	// With security violations, risk level should be critical
	if health.OverallRiskLevel != "critical" {
		t.Errorf("Expected overall risk level to be critical with security violations, got %s", health.OverallRiskLevel)
	}
}

func TestSecurityValidationMonitor_GetPerformanceMetrics(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SecurityValidationConfig{
		Enabled:                 true,
		CollectionInterval:      30 * time.Second,
		SlowValidationThreshold: 200 * time.Millisecond,
		MaxValidationStats:      1000,
		AlertingEnabled:         true,
	}

	monitor := NewAdvancedSecurityValidationMonitor(logger, config)
	defer monitor.Stop()

	// Record a validation
	result := &AdvancedSecurityValidationResult{
		ValidationID:               "test_validation",
		ValidationType:             "data_source_validation",
		ValidationName:             "test_data_source",
		ExecutionTime:              100 * time.Millisecond,
		Success:                    true,
		Error:                      nil,
		SecurityViolation:          false,
		ComplianceViolation:        false,
		ThreatDetected:             false,
		VulnerabilityFound:         false,
		TrustScore:                 0.95,
		ConfidenceLevel:            0.90,
		RiskLevel:                  "low",
		SecurityRecommendations:    []string{},
		PerformanceRecommendations: []string{},
		Metadata:                   make(map[string]interface{}),
		Timestamp:                  time.Now(),
	}

	monitor.RecordSecurityValidation(context.Background(), result)

	// Wait a moment for metrics collection
	time.Sleep(100 * time.Millisecond)

	// Get performance metrics
	metrics := monitor.GetPerformanceMetrics(10)
	if len(metrics) == 0 {
		t.Error("Expected to have performance metrics, but got none")
	}

	// Check metric structure
	for _, metric := range metrics {
		if metric.ID == "" {
			t.Error("Expected metric to have an ID")
		}

		if metric.MetricType == "" {
			t.Error("Expected metric to have a metric type")
		}

		if metric.ValidationType == "" {
			t.Error("Expected metric to have a validation type")
		}

		if metric.ValidationName == "" {
			t.Error("Expected metric to have a validation name")
		}

		if metric.PerformanceScore < 0 || metric.PerformanceScore > 100 {
			t.Error("Expected performance score between 0 and 100")
		}

		if metric.SecurityScore < 0 || metric.SecurityScore > 100 {
			t.Error("Expected security score between 0 and 100")
		}

		if metric.OverallScore < 0 || metric.OverallScore > 100 {
			t.Error("Expected overall score between 0 and 100")
		}
	}
}

func TestSecurityValidationMonitor_ValidationKeyGeneration(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SecurityValidationConfig{
		Enabled: true,
	}

	monitor := NewAdvancedSecurityValidationMonitor(logger, config)
	defer monitor.Stop()

	// Test that same validation type and name generates same key
	validationType1 := "data_source_validation"
	validationName1 := "test_data_source"
	validationType2 := "data_source_validation"
	validationName2 := "test_data_source"
	validationType3 := "website_verification"
	validationName3 := "test_data_source"

	key1 := monitor.generateValidationKey(validationType1, validationName1)
	key2 := monitor.generateValidationKey(validationType2, validationName2)
	key3 := monitor.generateValidationKey(validationType3, validationName3)

	if key1 != key2 {
		t.Error("Expected same validation type and name to generate same key")
	}

	if key1 == key3 {
		t.Error("Expected different validation type to generate different key")
	}

	if key1 == "" {
		t.Error("Expected validation key to be non-empty")
	}
}

func TestSecurityValidationMonitor_AlertSeverityDetermination(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &SecurityValidationConfig{
		Enabled: true,
	}

	monitor := NewAdvancedSecurityValidationMonitor(logger, config)
	defer monitor.Stop()

	tests := []struct {
		actual    float64
		threshold float64
		expected  string
	}{
		{100, 200, "low"},       // 0.5x threshold
		{300, 200, "low"},       // 1.5x threshold
		{400, 200, "medium"},    // 2x threshold
		{600, 200, "high"},      // 3x threshold
		{1000, 200, "critical"}, // 5x threshold
		{1200, 200, "critical"}, // 6x threshold
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("actual_%.0f_threshold_%.0f", tt.actual, tt.threshold), func(t *testing.T) {
			severity := monitor.determineSecurityAlertSeverity(tt.actual, tt.threshold)
			if severity != tt.expected {
				t.Errorf("Expected severity %s, got %s", tt.expected, severity)
			}
		})
	}
}

func BenchmarkSecurityValidationMonitor_RecordSecurityValidation(b *testing.B) {
	logger := zaptest.NewLogger(b)
	config := &SecurityValidationConfig{
		Enabled:                 true,
		CollectionInterval:      30 * time.Second,
		SlowValidationThreshold: 200 * time.Millisecond,
		MaxValidationStats:      1000,
		AlertingEnabled:         false, // Disable alerting for benchmark
	}

	monitor := NewAdvancedSecurityValidationMonitor(logger, config)
	defer monitor.Stop()

	result := &AdvancedSecurityValidationResult{
		ValidationID:               "benchmark_validation",
		ValidationType:             "data_source_validation",
		ValidationName:             "benchmark_data_source",
		ExecutionTime:              100 * time.Millisecond,
		Success:                    true,
		Error:                      nil,
		SecurityViolation:          false,
		ComplianceViolation:        false,
		ThreatDetected:             false,
		VulnerabilityFound:         false,
		TrustScore:                 0.95,
		ConfidenceLevel:            0.90,
		RiskLevel:                  "low",
		SecurityRecommendations:    []string{},
		PerformanceRecommendations: []string{},
		Metadata:                   make(map[string]interface{}),
		Timestamp:                  time.Now(),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			monitor.RecordSecurityValidation(context.Background(), result)
		}
	})
}

func BenchmarkSecurityValidationMonitor_GetValidationStats(b *testing.B) {
	logger := zaptest.NewLogger(b)
	config := &SecurityValidationConfig{
		Enabled:                 true,
		CollectionInterval:      30 * time.Second,
		SlowValidationThreshold: 200 * time.Millisecond,
		MaxValidationStats:      1000,
		AlertingEnabled:         false,
	}

	monitor := NewAdvancedSecurityValidationMonitor(logger, config)
	defer monitor.Stop()

	// Pre-populate with some validation stats
	for i := 0; i < 100; i++ {
		result := &AdvancedSecurityValidationResult{
			ValidationID:               fmt.Sprintf("validation_%d", i),
			ValidationType:             "data_source_validation",
			ValidationName:             fmt.Sprintf("data_source_%d", i),
			ExecutionTime:              100 * time.Millisecond,
			Success:                    true,
			Error:                      nil,
			SecurityViolation:          false,
			ComplianceViolation:        false,
			ThreatDetected:             false,
			VulnerabilityFound:         false,
			TrustScore:                 0.95,
			ConfidenceLevel:            0.90,
			RiskLevel:                  "low",
			SecurityRecommendations:    []string{},
			PerformanceRecommendations: []string{},
			Metadata:                   make(map[string]interface{}),
			Timestamp:                  time.Now(),
		}
		monitor.RecordSecurityValidation(context.Background(), result)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		monitor.GetValidationStats(50)
	}
}
