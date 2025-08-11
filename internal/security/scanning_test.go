package security

import (
	"context"
	"testing"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewSecurityScanningSystem(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled:                 true,
		ScanInterval:            1 * time.Hour,
		MaxConcurrentScans:      5,
		ScanTimeout:             30 * time.Minute,
		OutputDirectory:         "./security-reports",
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
		VulnerabilityScanning: VulnerabilityScanConfig{
			Enabled:           true,
			Tools:             []string{"trivy", "snyk"},
			SeverityThreshold: "HIGH",
			FailOnCritical:    true,
			FailOnHigh:        false,
			MaxCriticalVulns:  0,
			MaxHighVulns:      5,
			MaxMediumVulns:    20,
			ScanTimeout:       10 * time.Minute,
		},
		ContainerScanning: ContainerScanConfig{
			Enabled:      true,
			Tools:        []string{"hadolint", "docker-bench-security"},
			FailOnIssues: true,
			MaxIssues:    10,
			ScanTimeout:  5 * time.Minute,
		},
		SecretScanning: SecretScanConfig{
			Enabled:       true,
			Tools:         []string{"trufflehog", "git-secrets"},
			Patterns:      []string{"api_key", "password", "secret"},
			FailOnSecrets: true,
			ScanTimeout:   5 * time.Minute,
		},
		DependencyScanning: DependencyScanConfig{
			Enabled:               true,
			Tools:                 []string{"govulncheck", "snyk"},
			FailOnVulnerabilities: true,
			AutoUpdate:            false,
			ScanTimeout:           10 * time.Minute,
		},
		ComplianceScanning: ComplianceScanConfig{
			Enabled:      true,
			Frameworks:   []string{"SOC2", "PCI-DSS", "GDPR"},
			FailOnIssues: false,
			ScanTimeout:  15 * time.Minute,
		},
		Reporting: ReportConfig{
			Enabled:            true,
			Format:             "json",
			OutputDirectory:    "./security-reports",
			IncludeDetails:     true,
			IncludeRemediation: true,
			EmailRecipients:    []string{"security@kybplatform.com"},
			SlackWebhook:       "https://hooks.slack.com/services/...",
		},
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	// Create log aggregation config for error tracking
	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	assert.NotNil(t, sss)
	assert.Equal(t, config, sss.config)
	assert.Equal(t, logger, sss.logger)
	assert.Equal(t, monitoring, sss.monitoring)
	assert.Equal(t, errorTracking, sss.errorTracking)
	assert.NotNil(t, sss.results)
	assert.NotNil(t, sss.history)
	assert.NotNil(t, sss.vulnDB)
}

func TestRunVulnerabilityScan(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		VulnerabilityScanning: VulnerabilityScanConfig{
			Enabled:           true,
			Tools:             []string{"trivy", "snyk"},
			SeverityThreshold: "HIGH",
			FailOnCritical:    true,
			FailOnHigh:        false,
			MaxCriticalVulns:  0,
			MaxHighVulns:      5,
			MaxMediumVulns:    20,
			ScanTimeout:       10 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Test vulnerability scan
	scanResult, err := sss.RunScan(context.Background(), ScanTypeVulnerability, "kyb-platform:latest")

	assert.NoError(t, err)
	assert.NotNil(t, scanResult)
	assert.Equal(t, ScanTypeVulnerability, scanResult.ScanType)
	assert.Equal(t, "kyb-platform:latest", scanResult.Target)
	assert.Equal(t, StatusCompleted, scanResult.Status)
	assert.NotEmpty(t, scanResult.ID)
	assert.True(t, scanResult.Duration > 0)
	assert.NotNil(t, scanResult.Summary)

	// Verify vulnerabilities were found
	assert.Len(t, scanResult.Vulnerabilities, 1)
	assert.Equal(t, "CVE-2023-1234", scanResult.Vulnerabilities[0].ID)
	assert.Equal(t, SeverityMedium, scanResult.Vulnerabilities[0].Severity)
}

func TestRunContainerScan(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		ContainerScanning: ContainerScanConfig{
			Enabled:      true,
			Tools:        []string{"hadolint", "docker-bench-security"},
			FailOnIssues: true,
			MaxIssues:    10,
			ScanTimeout:  5 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Test container scan
	scanResult, err := sss.RunScan(context.Background(), ScanTypeContainer, "Dockerfile")

	assert.NoError(t, err)
	assert.NotNil(t, scanResult)
	assert.Equal(t, ScanTypeContainer, scanResult.ScanType)
	assert.Equal(t, "Dockerfile", scanResult.Target)
	assert.Equal(t, StatusCompleted, scanResult.Status)
	assert.NotEmpty(t, scanResult.ID)
	assert.True(t, scanResult.Duration > 0)
	assert.NotNil(t, scanResult.Summary)

	// Verify container issues were found
	assert.Len(t, scanResult.ContainerIssues, 1)
	assert.Equal(t, "HADOLINT-001", scanResult.ContainerIssues[0].ID)
	assert.Equal(t, SeverityLow, scanResult.ContainerIssues[0].Severity)
}

func TestRunSecretScan(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		SecretScanning: SecretScanConfig{
			Enabled:       true,
			Tools:         []string{"trufflehog", "git-secrets"},
			Patterns:      []string{"api_key", "password", "secret"},
			FailOnSecrets: true,
			ScanTimeout:   5 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Test secret scan
	scanResult, err := sss.RunScan(context.Background(), ScanTypeSecret, "./")

	assert.NoError(t, err)
	assert.NotNil(t, scanResult)
	assert.Equal(t, ScanTypeSecret, scanResult.ScanType)
	assert.Equal(t, "./", scanResult.Target)
	assert.Equal(t, StatusCompleted, scanResult.Status)
	assert.NotEmpty(t, scanResult.ID)
	assert.True(t, scanResult.Duration > 0)
	assert.NotNil(t, scanResult.Summary)
}

func TestRunDependencyScan(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		DependencyScanning: DependencyScanConfig{
			Enabled:               true,
			Tools:                 []string{"govulncheck", "snyk"},
			FailOnVulnerabilities: true,
			AutoUpdate:            false,
			ScanTimeout:           10 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Test dependency scan
	scanResult, err := sss.RunScan(context.Background(), ScanTypeDependency, "./")

	assert.NoError(t, err)
	assert.NotNil(t, scanResult)
	assert.Equal(t, ScanTypeDependency, scanResult.ScanType)
	assert.Equal(t, "./", scanResult.Target)
	assert.Equal(t, StatusCompleted, scanResult.Status)
	assert.NotEmpty(t, scanResult.ID)
	assert.True(t, scanResult.Duration > 0)
	assert.NotNil(t, scanResult.Summary)
}

func TestRunComplianceScan(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		ComplianceScanning: ComplianceScanConfig{
			Enabled:      true,
			Frameworks:   []string{"SOC2", "PCI-DSS", "GDPR"},
			FailOnIssues: false,
			ScanTimeout:  15 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Test compliance scan
	scanResult, err := sss.RunScan(context.Background(), ScanTypeCompliance, "kyb-platform")

	assert.NoError(t, err)
	assert.NotNil(t, scanResult)
	assert.Equal(t, ScanTypeCompliance, scanResult.ScanType)
	assert.Equal(t, "kyb-platform", scanResult.Target)
	assert.Equal(t, StatusCompleted, scanResult.Status)
	assert.NotEmpty(t, scanResult.ID)
	assert.True(t, scanResult.Duration > 0)
	assert.NotNil(t, scanResult.Summary)

	// Verify compliance issues were found
	assert.Len(t, scanResult.ComplianceIssues, 3) // One for each framework
	for _, issue := range scanResult.ComplianceIssues {
		assert.Contains(t, []string{"SOC2", "PCI-DSS", "GDPR"}, issue.Framework)
		assert.Equal(t, "failed", issue.Status)
		assert.Equal(t, SeverityMedium, issue.Severity)
	}
}

func TestRunFullScan(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		VulnerabilityScanning: VulnerabilityScanConfig{
			Enabled:           true,
			Tools:             []string{"trivy"},
			SeverityThreshold: "HIGH",
			FailOnCritical:    true,
			FailOnHigh:        false,
			MaxCriticalVulns:  0,
			MaxHighVulns:      5,
			MaxMediumVulns:    20,
			ScanTimeout:       10 * time.Minute,
		},
		ContainerScanning: ContainerScanConfig{
			Enabled:      true,
			Tools:        []string{"hadolint"},
			FailOnIssues: true,
			MaxIssues:    10,
			ScanTimeout:  5 * time.Minute,
		},
		SecretScanning: SecretScanConfig{
			Enabled:       true,
			Tools:         []string{"trufflehog"},
			Patterns:      []string{"api_key", "password", "secret"},
			FailOnSecrets: true,
			ScanTimeout:   5 * time.Minute,
		},
		DependencyScanning: DependencyScanConfig{
			Enabled:               true,
			Tools:                 []string{"govulncheck"},
			FailOnVulnerabilities: true,
			AutoUpdate:            false,
			ScanTimeout:           10 * time.Minute,
		},
		ComplianceScanning: ComplianceScanConfig{
			Enabled:      true,
			Frameworks:   []string{"SOC2"},
			FailOnIssues: false,
			ScanTimeout:  15 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Test full scan
	scanResult, err := sss.RunScan(context.Background(), ScanTypeFull, "kyb-platform")

	assert.NoError(t, err)
	assert.NotNil(t, scanResult)
	assert.Equal(t, ScanTypeFull, scanResult.ScanType)
	assert.Equal(t, "kyb-platform", scanResult.Target)
	assert.Equal(t, StatusCompleted, scanResult.Status)
	assert.NotEmpty(t, scanResult.ID)
	assert.True(t, scanResult.Duration > 0)
	assert.NotNil(t, scanResult.Summary)

	// Verify all scan types were executed
	assert.Len(t, scanResult.Vulnerabilities, 1)
	assert.Len(t, scanResult.ContainerIssues, 1)
	assert.Len(t, scanResult.ComplianceIssues, 1)
}

func TestScanWithOptions(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		VulnerabilityScanning: VulnerabilityScanConfig{
			Enabled:           true,
			Tools:             []string{"trivy"},
			SeverityThreshold: "HIGH",
			FailOnCritical:    true,
			FailOnHigh:        false,
			MaxCriticalVulns:  0,
			MaxHighVulns:      5,
			MaxMediumVulns:    20,
			ScanTimeout:       10 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Test scan with options
	scanResult, err := sss.RunScan(context.Background(), ScanTypeVulnerability, "kyb-platform:latest",
		WithScanMetadata("environment", "production"),
		WithScanMetadata("team", "security"),
		WithScanTimeout(5*time.Minute),
	)

	assert.NoError(t, err)
	assert.NotNil(t, scanResult)
	assert.Equal(t, "production", scanResult.Metadata["environment"])
	assert.Equal(t, "security", scanResult.Metadata["team"])
	assert.Equal(t, 5*time.Minute, scanResult.Metadata["timeout"])
}

func TestScanDisabled(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: false, // Disabled
		VulnerabilityScanning: VulnerabilityScanConfig{
			Enabled:           true,
			Tools:             []string{"trivy"},
			SeverityThreshold: "HIGH",
			FailOnCritical:    true,
			FailOnHigh:        false,
			MaxCriticalVulns:  0,
			MaxHighVulns:      5,
			MaxMediumVulns:    20,
			ScanTimeout:       10 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Test scan when disabled
	scanResult, err := sss.RunScan(context.Background(), ScanTypeVulnerability, "kyb-platform:latest")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "security scanning is disabled")
	assert.Nil(t, scanResult)
}

func TestUnknownScanType(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		VulnerabilityScanning: VulnerabilityScanConfig{
			Enabled:           true,
			Tools:             []string{"trivy"},
			SeverityThreshold: "HIGH",
			FailOnCritical:    true,
			FailOnHigh:        false,
			MaxCriticalVulns:  0,
			MaxHighVulns:      5,
			MaxMediumVulns:    20,
			ScanTimeout:       10 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Test unknown scan type
	scanResult, err := sss.RunScan(context.Background(), "unknown-scan-type", "kyb-platform:latest")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown scan type")
	assert.NotNil(t, scanResult) // Scan result is created even for failed scans
	assert.Equal(t, StatusFailed, scanResult.Status)
}

func TestGetScanResult(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		VulnerabilityScanning: VulnerabilityScanConfig{
			Enabled:           true,
			Tools:             []string{"trivy"},
			SeverityThreshold: "HIGH",
			FailOnCritical:    true,
			FailOnHigh:        false,
			MaxCriticalVulns:  0,
			MaxHighVulns:      5,
			MaxMediumVulns:    20,
			ScanTimeout:       10 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Run a scan
	scanResult, err := sss.RunScan(context.Background(), ScanTypeVulnerability, "kyb-platform:latest")
	require.NoError(t, err)
	require.NotNil(t, scanResult)

	// Get scan result
	retrievedResult, exists := sss.GetScanResult(scanResult.ID)

	assert.True(t, exists)
	assert.NotNil(t, retrievedResult)
	assert.Equal(t, scanResult.ID, retrievedResult.ID)
	assert.Equal(t, scanResult.ScanType, retrievedResult.ScanType)
	assert.Equal(t, scanResult.Target, retrievedResult.Target)
	assert.Equal(t, scanResult.Status, retrievedResult.Status)

	// Test getting non-existent scan result
	nonExistentResult, exists := sss.GetScanResult("non-existent-id")
	assert.False(t, exists)
	assert.Nil(t, nonExistentResult)
}

func TestGetScanHistory(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		VulnerabilityScanning: VulnerabilityScanConfig{
			Enabled:           true,
			Tools:             []string{"trivy"},
			SeverityThreshold: "HIGH",
			FailOnCritical:    true,
			FailOnHigh:        false,
			MaxCriticalVulns:  0,
			MaxHighVulns:      5,
			MaxMediumVulns:    20,
			ScanTimeout:       10 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Run multiple scans
	scanResult1, err := sss.RunScan(context.Background(), ScanTypeVulnerability, "kyb-platform:latest")
	require.NoError(t, err)

	scanResult2, err := sss.RunScan(context.Background(), ScanTypeContainer, "Dockerfile")
	require.NoError(t, err)

	// Get scan history
	history := sss.GetScanHistory()

	assert.Len(t, history, 2)
	assert.Equal(t, scanResult1.ID, history[0].ID)
	assert.Equal(t, scanResult2.ID, history[1].ID)
	assert.Equal(t, ScanTypeVulnerability, history[0].ScanType)
	assert.Equal(t, ScanTypeContainer, history[1].ScanType)
	assert.Equal(t, StatusCompleted, history[0].Status)
	assert.Equal(t, StatusCompleted, history[1].Status)
	assert.True(t, history[0].Duration > 0)
	assert.True(t, history[1].Duration > 0)
}

func TestGetScanResults(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled: true,
		VulnerabilityScanning: VulnerabilityScanConfig{
			Enabled:           true,
			Tools:             []string{"trivy"},
			SeverityThreshold: "HIGH",
			FailOnCritical:    true,
			FailOnHigh:        false,
			MaxCriticalVulns:  0,
			MaxHighVulns:      5,
			MaxMediumVulns:    20,
			ScanTimeout:       10 * time.Minute,
		},
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Run multiple scans
	scanResult1, err := sss.RunScan(context.Background(), ScanTypeVulnerability, "kyb-platform:latest")
	require.NoError(t, err)

	scanResult2, err := sss.RunScan(context.Background(), ScanTypeContainer, "Dockerfile")
	require.NoError(t, err)

	// Get all scan results
	results := sss.GetScanResults()

	assert.Len(t, results, 2)
	assert.Contains(t, results, scanResult1.ID)
	assert.Contains(t, results, scanResult2.ID)
	assert.Equal(t, scanResult1, results[scanResult1.ID])
	assert.Equal(t, scanResult2, results[scanResult2.ID])
}

func TestCalculateSummary(t *testing.T) {
	config := &SecurityScanningConfig{
		Enabled:                 true,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
		EnableErrorTracking:     true,
	}

	logger := zap.NewNop()
	monitoring := observability.NewMonitoringSystem(logger)

	logConfig := &observability.LogAggregationConfig{
		EnableConsole: true,
		Environment:   "test",
		Application:   "kyb-platform",
		Version:       "1.0.0",
	}

	logAggregation, err := observability.NewLogAggregationSystem(logConfig)
	require.NoError(t, err)

	errorTracking := observability.NewErrorTrackingSystem(monitoring, logAggregation, &observability.ErrorTrackingConfig{
		EnableErrorTracking:     true,
		MaxErrorsStored:         1000,
		ErrorRetentionPeriod:    24 * time.Hour,
		EnablePrometheusMetrics: true,
		EnableLogIntegration:    true,
	}, logger)

	sss := NewSecurityScanningSystem(monitoring, errorTracking, config, logger)

	// Create a scan result with various issues
	scanResult := &ScanResult{
		ID:        "test-scan-1",
		Timestamp: time.Now(),
		ScanType:  ScanTypeFull,
		Target:    "test-target",
		Status:    StatusCompleted,
		Vulnerabilities: []*Vulnerability{
			{ID: "CVE-1", Severity: SeverityCritical},
			{ID: "CVE-2", Severity: SeverityHigh},
			{ID: "CVE-3", Severity: SeverityMedium},
		},
		ContainerIssues: []*ContainerIssue{
			{ID: "CONT-1", Severity: SeverityLow},
		},
		Secrets: []*Secret{
			{ID: "SEC-1", Severity: SeverityHigh},
		},
		ComplianceIssues: []*ComplianceIssue{
			{ID: "COMP-1", Severity: SeverityMedium, Status: "failed"},
			{ID: "COMP-2", Severity: SeverityLow, Status: "passed"},
		},
		Summary: &ScanSummary{},
	}

	// Calculate summary
	sss.calculateSummary(scanResult)

	// Verify summary calculations
	summary := scanResult.Summary
	assert.Equal(t, 7, summary.TotalIssues) // 3 vulns + 1 container + 1 secret + 2 compliance
	assert.Equal(t, 1, summary.CriticalIssues)
	assert.Equal(t, 2, summary.HighIssues)   // 1 vuln + 1 secret
	assert.Equal(t, 2, summary.MediumIssues) // 1 vuln + 1 compliance
	assert.Equal(t, 2, summary.LowIssues)    // 1 container + 1 compliance
	assert.Equal(t, 0, summary.InfoIssues)
	assert.Equal(t, 50.0, summary.ComplianceScore) // 1 passed out of 2
	assert.True(t, summary.RiskScore > 0)
	assert.NotEmpty(t, summary.Recommendations)
}

func TestHelperFunctions(t *testing.T) {
	// Test generateScanID
	scanID1 := generateScanID()
	scanID2 := generateScanID()

	assert.NotEmpty(t, scanID1)
	assert.NotEmpty(t, scanID2)
	assert.NotEqual(t, scanID1, scanID2)
	assert.Contains(t, scanID1, "scan_")
	assert.Contains(t, scanID2, "scan_")
}
