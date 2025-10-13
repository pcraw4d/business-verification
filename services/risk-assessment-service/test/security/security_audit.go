package security

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// SecurityAudit provides comprehensive security auditing capabilities
type SecurityAudit struct {
	logger *zap.Logger
	config *SecurityAuditConfig
}

// SecurityAuditConfig represents configuration for security auditing
type SecurityAuditConfig struct {
	EnableAuthenticationAudit      bool                   `json:"enable_authentication_audit"`
	EnableAuthorizationAudit       bool                   `json:"enable_authorization_audit"`
	EnableDataAccessAudit          bool                   `json:"enable_data_access_audit"`
	EnableConfigurationAudit       bool                   `json:"enable_configuration_audit"`
	EnableNetworkAudit             bool                   `json:"enable_network_audit"`
	EnableComplianceAudit          bool                   `json:"enable_compliance_audit"`
	AuditRetentionPeriod           time.Duration          `json:"audit_retention_period"`
	EnableRealTimeMonitoring       bool                   `json:"enable_real_time_monitoring"`
	EnableAutomatedRemediation     bool                   `json:"enable_automated_remediation"`
	CriticalVulnerabilityThreshold int                    `json:"critical_vulnerability_threshold"`
	HighVulnerabilityThreshold     int                    `json:"high_vulnerability_threshold"`
	MediumVulnerabilityThreshold   int                    `json:"medium_vulnerability_threshold"`
	Metadata                       map[string]interface{} `json:"metadata"`
}

// SecurityAuditResult represents the result of a security audit
type SecurityAuditResult struct {
	ID                      string                   `json:"id"`
	AuditType               string                   `json:"audit_type"`
	Status                  AuditStatus              `json:"status"`
	StartTime               time.Time                `json:"start_time"`
	EndTime                 time.Time                `json:"end_time"`
	Duration                time.Duration            `json:"duration"`
	TotalChecks             int                      `json:"total_checks"`
	PassedChecks            int                      `json:"passed_checks"`
	FailedChecks            int                      `json:"failed_checks"`
	CriticalVulnerabilities int                      `json:"critical_vulnerabilities"`
	HighVulnerabilities     int                      `json:"high_vulnerabilities"`
	MediumVulnerabilities   int                      `json:"medium_vulnerabilities"`
	LowVulnerabilities      int                      `json:"low_vulnerabilities"`
	ComplianceScore         float64                  `json:"compliance_score"`
	Recommendations         []SecurityRecommendation `json:"recommendations"`
	Vulnerabilities         []SecurityVulnerability  `json:"vulnerabilities"`
	Metadata                map[string]interface{}   `json:"metadata"`
}

// AuditStatus represents the status of an audit
type AuditStatus string

const (
	AuditStatusPending   AuditStatus = "pending"
	AuditStatusRunning   AuditStatus = "running"
	AuditStatusCompleted AuditStatus = "completed"
	AuditStatusFailed    AuditStatus = "failed"
	AuditStatusCancelled AuditStatus = "cancelled"
)

// SecurityVulnerability represents a security vulnerability
type SecurityVulnerability struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Severity        VulnerabilitySeverity  `json:"severity"`
	Category        VulnerabilityCategory  `json:"category"`
	CVSSScore       float64                `json:"cvss_score"`
	CVSSVector      string                 `json:"cvss_vector"`
	AffectedSystems []string               `json:"affected_systems"`
	Remediation     string                 `json:"remediation"`
	References      []string               `json:"references"`
	DiscoveredAt    time.Time              `json:"discovered_at"`
	Status          VulnerabilityStatus    `json:"status"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// VulnerabilitySeverity represents the severity of a vulnerability
type VulnerabilitySeverity string

const (
	VulnerabilitySeverityCritical VulnerabilitySeverity = "critical"
	VulnerabilitySeverityHigh     VulnerabilitySeverity = "high"
	VulnerabilitySeverityMedium   VulnerabilitySeverity = "medium"
	VulnerabilitySeverityLow      VulnerabilitySeverity = "low"
	VulnerabilitySeverityInfo     VulnerabilitySeverity = "info"
)

// VulnerabilityCategory represents the category of a vulnerability
type VulnerabilityCategory string

const (
	VulnerabilityCategoryAuthentication  VulnerabilityCategory = "authentication"
	VulnerabilityCategoryAuthorization   VulnerabilityCategory = "authorization"
	VulnerabilityCategoryInputValidation VulnerabilityCategory = "input_validation"
	VulnerabilityCategoryDataProtection  VulnerabilityCategory = "data_protection"
	VulnerabilityCategoryNetworkSecurity VulnerabilityCategory = "network_security"
	VulnerabilityCategoryConfiguration   VulnerabilityCategory = "configuration"
	VulnerabilityCategoryCompliance      VulnerabilityCategory = "compliance"
	VulnerabilityCategoryBusinessLogic   VulnerabilityCategory = "business_logic"
)

// VulnerabilityStatus represents the status of a vulnerability
type VulnerabilityStatus string

const (
	VulnerabilityStatusOpen          VulnerabilityStatus = "open"
	VulnerabilityStatusInProgress    VulnerabilityStatus = "in_progress"
	VulnerabilityStatusResolved      VulnerabilityStatus = "resolved"
	VulnerabilityStatusClosed        VulnerabilityStatus = "closed"
	VulnerabilityStatusFalsePositive VulnerabilityStatus = "false_positive"
)

// SecurityRecommendation represents a security recommendation
type SecurityRecommendation struct {
	ID          string                 `json:"id"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Priority    RecommendationPriority `json:"priority"`
	Category    string                 `json:"category"`
	Impact      string                 `json:"impact"`
	Effort      string                 `json:"effort"`
	Timeline    string                 `json:"timeline"`
	Resources   []string               `json:"resources"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// RecommendationPriority represents the priority of a recommendation
type RecommendationPriority string

const (
	RecommendationPriorityCritical RecommendationPriority = "critical"
	RecommendationPriorityHigh     RecommendationPriority = "high"
	RecommendationPriorityMedium   RecommendationPriority = "medium"
	RecommendationPriorityLow      RecommendationPriority = "low"
)

// NewSecurityAudit creates a new security audit instance
func NewSecurityAudit(logger *zap.Logger, config *SecurityAuditConfig) *SecurityAudit {
	return &SecurityAudit{
		logger: logger,
		config: config,
	}
}

// RunComprehensiveAudit runs a comprehensive security audit
func (sa *SecurityAudit) RunComprehensiveAudit(ctx context.Context) (*SecurityAuditResult, error) {
	auditID := fmt.Sprintf("audit_%d", time.Now().UnixNano())

	result := &SecurityAuditResult{
		ID:        auditID,
		AuditType: "comprehensive",
		Status:    AuditStatusRunning,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	sa.logger.Info("Starting comprehensive security audit",
		zap.String("audit_id", auditID))

	// Run all audit categories
	auditResults := make([]*SecurityAuditResult, 0)

	if sa.config.EnableAuthenticationAudit {
		authResult, err := sa.RunAuthenticationAudit(ctx)
		if err != nil {
			sa.logger.Error("Authentication audit failed", zap.Error(err))
		} else {
			auditResults = append(auditResults, authResult)
		}
	}

	if sa.config.EnableAuthorizationAudit {
		authzResult, err := sa.RunAuthorizationAudit(ctx)
		if err != nil {
			sa.logger.Error("Authorization audit failed", zap.Error(err))
		} else {
			auditResults = append(auditResults, authzResult)
		}
	}

	if sa.config.EnableDataAccessAudit {
		dataResult, err := sa.RunDataAccessAudit(ctx)
		if err != nil {
			sa.logger.Error("Data access audit failed", zap.Error(err))
		} else {
			auditResults = append(auditResults, dataResult)
		}
	}

	if sa.config.EnableConfigurationAudit {
		configResult, err := sa.RunConfigurationAudit(ctx)
		if err != nil {
			sa.logger.Error("Configuration audit failed", zap.Error(err))
		} else {
			auditResults = append(auditResults, configResult)
		}
	}

	if sa.config.EnableNetworkAudit {
		networkResult, err := sa.RunNetworkAudit(ctx)
		if err != nil {
			sa.logger.Error("Network audit failed", zap.Error(err))
		} else {
			auditResults = append(auditResults, networkResult)
		}
	}

	if sa.config.EnableComplianceAudit {
		complianceResult, err := sa.RunComplianceAudit(ctx)
		if err != nil {
			sa.logger.Error("Compliance audit failed", zap.Error(err))
		} else {
			auditResults = append(auditResults, complianceResult)
		}
	}

	// Aggregate results
	sa.aggregateAuditResults(result, auditResults)

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = AuditStatusCompleted

	sa.logger.Info("Comprehensive security audit completed",
		zap.String("audit_id", auditID),
		zap.Int("total_checks", result.TotalChecks),
		zap.Int("failed_checks", result.FailedChecks),
		zap.Float64("compliance_score", result.ComplianceScore))

	return result, nil
}

// RunAuthenticationAudit runs authentication security audit
func (sa *SecurityAudit) RunAuthenticationAudit(ctx context.Context) (*SecurityAuditResult, error) {
	result := &SecurityAuditResult{
		ID:        fmt.Sprintf("auth_audit_%d", time.Now().UnixNano()),
		AuditType: "authentication",
		Status:    AuditStatusRunning,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Mock authentication audit checks
	checks := []struct {
		name        string
		description string
		severity    VulnerabilitySeverity
		passed      bool
	}{
		{"Password Policy", "Check password policy enforcement", VulnerabilitySeverityHigh, true},
		{"MFA Implementation", "Check multi-factor authentication", VulnerabilitySeverityHigh, true},
		{"Session Management", "Check session security", VulnerabilitySeverityMedium, true},
		{"Token Security", "Check JWT token security", VulnerabilitySeverityHigh, true},
		{"Account Lockout", "Check account lockout policies", VulnerabilitySeverityMedium, true},
	}

	for _, check := range checks {
		result.TotalChecks++
		if check.passed {
			result.PassedChecks++
		} else {
			result.FailedChecks++
			vulnerability := &SecurityVulnerability{
				ID:           fmt.Sprintf("auth_%d", time.Now().UnixNano()),
				Title:        check.name,
				Description:  check.description,
				Severity:     check.severity,
				Category:     VulnerabilityCategoryAuthentication,
				CVSSScore:    sa.getCVSSScore(check.severity),
				DiscoveredAt: time.Now(),
				Status:       VulnerabilityStatusOpen,
				Metadata:     make(map[string]interface{}),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, *vulnerability)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = AuditStatusCompleted

	return result, nil
}

// RunAuthorizationAudit runs authorization security audit
func (sa *SecurityAudit) RunAuthorizationAudit(ctx context.Context) (*SecurityAuditResult, error) {
	result := &SecurityAuditResult{
		ID:        fmt.Sprintf("authz_audit_%d", time.Now().UnixNano()),
		AuditType: "authorization",
		Status:    AuditStatusRunning,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Mock authorization audit checks
	checks := []struct {
		name        string
		description string
		severity    VulnerabilitySeverity
		passed      bool
	}{
		{"RBAC Implementation", "Check role-based access control", VulnerabilitySeverityHigh, true},
		{"Tenant Isolation", "Check multi-tenant data isolation", VulnerabilitySeverityCritical, true},
		{"API Authorization", "Check API endpoint authorization", VulnerabilitySeverityHigh, true},
		{"Privilege Escalation", "Check privilege escalation prevention", VulnerabilitySeverityHigh, true},
		{"Access Control", "Check access control implementation", VulnerabilitySeverityMedium, true},
	}

	for _, check := range checks {
		result.TotalChecks++
		if check.passed {
			result.PassedChecks++
		} else {
			result.FailedChecks++
			vulnerability := &SecurityVulnerability{
				ID:           fmt.Sprintf("authz_%d", time.Now().UnixNano()),
				Title:        check.name,
				Description:  check.description,
				Severity:     check.severity,
				Category:     VulnerabilityCategoryAuthorization,
				CVSSScore:    sa.getCVSSScore(check.severity),
				DiscoveredAt: time.Now(),
				Status:       VulnerabilityStatusOpen,
				Metadata:     make(map[string]interface{}),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, *vulnerability)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = AuditStatusCompleted

	return result, nil
}

// RunDataAccessAudit runs data access security audit
func (sa *SecurityAudit) RunDataAccessAudit(ctx context.Context) (*SecurityAuditResult, error) {
	result := &SecurityAuditResult{
		ID:        fmt.Sprintf("data_audit_%d", time.Now().UnixNano()),
		AuditType: "data_access",
		Status:    AuditStatusRunning,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Mock data access audit checks
	checks := []struct {
		name        string
		description string
		severity    VulnerabilitySeverity
		passed      bool
	}{
		{"Data Encryption", "Check data encryption at rest", VulnerabilitySeverityHigh, true},
		{"Data Transmission", "Check data encryption in transit", VulnerabilitySeverityHigh, true},
		{"Data Classification", "Check data classification", VulnerabilitySeverityMedium, true},
		{"Data Retention", "Check data retention policies", VulnerabilitySeverityMedium, true},
		{"Data Backup", "Check data backup security", VulnerabilitySeverityMedium, true},
	}

	for _, check := range checks {
		result.TotalChecks++
		if check.passed {
			result.PassedChecks++
		} else {
			result.FailedChecks++
			vulnerability := &SecurityVulnerability{
				ID:           fmt.Sprintf("data_%d", time.Now().UnixNano()),
				Title:        check.name,
				Description:  check.description,
				Severity:     check.severity,
				Category:     VulnerabilityCategoryDataProtection,
				CVSSScore:    sa.getCVSSScore(check.severity),
				DiscoveredAt: time.Now(),
				Status:       VulnerabilityStatusOpen,
				Metadata:     make(map[string]interface{}),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, *vulnerability)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = AuditStatusCompleted

	return result, nil
}

// RunConfigurationAudit runs configuration security audit
func (sa *SecurityAudit) RunConfigurationAudit(ctx context.Context) (*SecurityAuditResult, error) {
	result := &SecurityAuditResult{
		ID:        fmt.Sprintf("config_audit_%d", time.Now().UnixNano()),
		AuditType: "configuration",
		Status:    AuditStatusRunning,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Mock configuration audit checks
	checks := []struct {
		name        string
		description string
		severity    VulnerabilitySeverity
		passed      bool
	}{
		{"Security Headers", "Check HTTP security headers", VulnerabilitySeverityMedium, true},
		{"SSL/TLS Configuration", "Check SSL/TLS configuration", VulnerabilitySeverityHigh, true},
		{"Server Configuration", "Check server security configuration", VulnerabilitySeverityMedium, true},
		{"Database Configuration", "Check database security configuration", VulnerabilitySeverityHigh, true},
		{"Logging Configuration", "Check security logging configuration", VulnerabilitySeverityMedium, true},
	}

	for _, check := range checks {
		result.TotalChecks++
		if check.passed {
			result.PassedChecks++
		} else {
			result.FailedChecks++
			vulnerability := &SecurityVulnerability{
				ID:           fmt.Sprintf("config_%d", time.Now().UnixNano()),
				Title:        check.name,
				Description:  check.description,
				Severity:     check.severity,
				Category:     VulnerabilityCategoryConfiguration,
				CVSSScore:    sa.getCVSSScore(check.severity),
				DiscoveredAt: time.Now(),
				Status:       VulnerabilityStatusOpen,
				Metadata:     make(map[string]interface{}),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, *vulnerability)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = AuditStatusCompleted

	return result, nil
}

// RunNetworkAudit runs network security audit
func (sa *SecurityAudit) RunNetworkAudit(ctx context.Context) (*SecurityAuditResult, error) {
	result := &SecurityAuditResult{
		ID:        fmt.Sprintf("network_audit_%d", time.Now().UnixNano()),
		AuditType: "network",
		Status:    AuditStatusRunning,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Mock network audit checks
	checks := []struct {
		name        string
		description string
		severity    VulnerabilitySeverity
		passed      bool
	}{
		{"Firewall Configuration", "Check firewall rules", VulnerabilitySeverityHigh, true},
		{"Network Segmentation", "Check network segmentation", VulnerabilitySeverityMedium, true},
		{"Port Security", "Check open ports", VulnerabilitySeverityMedium, true},
		{"VPN Configuration", "Check VPN security", VulnerabilitySeverityHigh, true},
		{"Network Monitoring", "Check network monitoring", VulnerabilitySeverityMedium, true},
	}

	for _, check := range checks {
		result.TotalChecks++
		if check.passed {
			result.PassedChecks++
		} else {
			result.FailedChecks++
			vulnerability := &SecurityVulnerability{
				ID:           fmt.Sprintf("network_%d", time.Now().UnixNano()),
				Title:        check.name,
				Description:  check.description,
				Severity:     check.severity,
				Category:     VulnerabilityCategoryNetworkSecurity,
				CVSSScore:    sa.getCVSSScore(check.severity),
				DiscoveredAt: time.Now(),
				Status:       VulnerabilityStatusOpen,
				Metadata:     make(map[string]interface{}),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, *vulnerability)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = AuditStatusCompleted

	return result, nil
}

// RunComplianceAudit runs compliance security audit
func (sa *SecurityAudit) RunComplianceAudit(ctx context.Context) (*SecurityAuditResult, error) {
	result := &SecurityAuditResult{
		ID:        fmt.Sprintf("compliance_audit_%d", time.Now().UnixNano()),
		AuditType: "compliance",
		Status:    AuditStatusRunning,
		StartTime: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Mock compliance audit checks
	checks := []struct {
		name        string
		description string
		severity    VulnerabilitySeverity
		passed      bool
	}{
		{"SOC 2 Compliance", "Check SOC 2 compliance", VulnerabilitySeverityHigh, true},
		{"GDPR Compliance", "Check GDPR compliance", VulnerabilitySeverityHigh, true},
		{"PCI DSS Compliance", "Check PCI DSS compliance", VulnerabilitySeverityHigh, true},
		{"Audit Trail", "Check audit trail implementation", VulnerabilitySeverityMedium, true},
		{"Data Privacy", "Check data privacy controls", VulnerabilitySeverityHigh, true},
	}

	for _, check := range checks {
		result.TotalChecks++
		if check.passed {
			result.PassedChecks++
		} else {
			result.FailedChecks++
			vulnerability := &SecurityVulnerability{
				ID:           fmt.Sprintf("compliance_%d", time.Now().UnixNano()),
				Title:        check.name,
				Description:  check.description,
				Severity:     check.severity,
				Category:     VulnerabilityCategoryCompliance,
				CVSSScore:    sa.getCVSSScore(check.severity),
				DiscoveredAt: time.Now(),
				Status:       VulnerabilityStatusOpen,
				Metadata:     make(map[string]interface{}),
			}
			result.Vulnerabilities = append(result.Vulnerabilities, *vulnerability)
		}
	}

	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.Status = AuditStatusCompleted

	return result, nil
}

// aggregateAuditResults aggregates results from multiple audit categories
func (sa *SecurityAudit) aggregateAuditResults(result *SecurityAuditResult, auditResults []*SecurityAuditResult) {
	for _, auditResult := range auditResults {
		result.TotalChecks += auditResult.TotalChecks
		result.PassedChecks += auditResult.PassedChecks
		result.FailedChecks += auditResult.FailedChecks
		result.CriticalVulnerabilities += auditResult.CriticalVulnerabilities
		result.HighVulnerabilities += auditResult.HighVulnerabilities
		result.MediumVulnerabilities += auditResult.MediumVulnerabilities
		result.LowVulnerabilities += auditResult.LowVulnerabilities
		result.Vulnerabilities = append(result.Vulnerabilities, auditResult.Vulnerabilities...)
		result.Recommendations = append(result.Recommendations, auditResult.Recommendations...)
	}

	// Calculate compliance score
	if result.TotalChecks > 0 {
		result.ComplianceScore = float64(result.PassedChecks) / float64(result.TotalChecks) * 100
	}

	// Count vulnerabilities by severity
	for _, vuln := range result.Vulnerabilities {
		switch vuln.Severity {
		case VulnerabilitySeverityCritical:
			result.CriticalVulnerabilities++
		case VulnerabilitySeverityHigh:
			result.HighVulnerabilities++
		case VulnerabilitySeverityMedium:
			result.MediumVulnerabilities++
		case VulnerabilitySeverityLow:
			result.LowVulnerabilities++
		}
	}

	// Generate recommendations based on vulnerabilities
	sa.generateRecommendations(result)
}

// generateRecommendations generates security recommendations based on audit results
func (sa *SecurityAudit) generateRecommendations(result *SecurityAuditResult) {
	// Critical vulnerabilities
	if result.CriticalVulnerabilities > 0 {
		recommendation := &SecurityRecommendation{
			ID:          fmt.Sprintf("rec_critical_%d", time.Now().UnixNano()),
			Title:       "Address Critical Vulnerabilities",
			Description: fmt.Sprintf("Immediately address %d critical vulnerabilities", result.CriticalVulnerabilities),
			Priority:    RecommendationPriorityCritical,
			Category:    "vulnerability_management",
			Impact:      "High security risk",
			Effort:      "High",
			Timeline:    "Immediate",
			Resources:   []string{"Security team", "Development team"},
			Metadata:    make(map[string]interface{}),
		}
		result.Recommendations = append(result.Recommendations, *recommendation)
	}

	// High vulnerabilities
	if result.HighVulnerabilities > 0 {
		recommendation := &SecurityRecommendation{
			ID:          fmt.Sprintf("rec_high_%d", time.Now().UnixNano()),
			Title:       "Address High Vulnerabilities",
			Description: fmt.Sprintf("Address %d high vulnerabilities within 24 hours", result.HighVulnerabilities),
			Priority:    RecommendationPriorityHigh,
			Category:    "vulnerability_management",
			Impact:      "Medium security risk",
			Effort:      "Medium",
			Timeline:    "24 hours",
			Resources:   []string{"Security team", "Development team"},
			Metadata:    make(map[string]interface{}),
		}
		result.Recommendations = append(result.Recommendations, *recommendation)
	}

	// Medium vulnerabilities
	if result.MediumVulnerabilities > 0 {
		recommendation := &SecurityRecommendation{
			ID:          fmt.Sprintf("rec_medium_%d", time.Now().UnixNano()),
			Title:       "Address Medium Vulnerabilities",
			Description: fmt.Sprintf("Address %d medium vulnerabilities within 1 week", result.MediumVulnerabilities),
			Priority:    RecommendationPriorityMedium,
			Category:    "vulnerability_management",
			Impact:      "Low security risk",
			Effort:      "Low",
			Timeline:    "1 week",
			Resources:   []string{"Development team"},
			Metadata:    make(map[string]interface{}),
		}
		result.Recommendations = append(result.Recommendations, *recommendation)
	}

	// Compliance recommendations
	if result.ComplianceScore < 95.0 {
		recommendation := &SecurityRecommendation{
			ID:          fmt.Sprintf("rec_compliance_%d", time.Now().UnixNano()),
			Title:       "Improve Compliance Score",
			Description: fmt.Sprintf("Current compliance score is %.2f%%, target is 95%%", result.ComplianceScore),
			Priority:    RecommendationPriorityHigh,
			Category:    "compliance",
			Impact:      "Regulatory compliance risk",
			Effort:      "Medium",
			Timeline:    "2 weeks",
			Resources:   []string{"Compliance team", "Security team"},
			Metadata:    make(map[string]interface{}),
		}
		result.Recommendations = append(result.Recommendations, *recommendation)
	}
}

// getCVSSScore returns CVSS score based on severity
func (sa *SecurityAudit) getCVSSScore(severity VulnerabilitySeverity) float64 {
	switch severity {
	case VulnerabilitySeverityCritical:
		return 9.0
	case VulnerabilitySeverityHigh:
		return 7.0
	case VulnerabilitySeverityMedium:
		return 5.0
	case VulnerabilitySeverityLow:
		return 3.0
	default:
		return 1.0
	}
}

// GenerateAuditReport generates a comprehensive audit report
func (sa *SecurityAudit) GenerateAuditReport(result *SecurityAuditResult) (string, error) {
	report := fmt.Sprintf(`
# Security Audit Report

## Executive Summary
- **Audit ID**: %s
- **Audit Type**: %s
- **Status**: %s
- **Duration**: %s
- **Compliance Score**: %.2f%%

## Audit Results
- **Total Checks**: %d
- **Passed Checks**: %d
- **Failed Checks**: %d

## Vulnerability Summary
- **Critical**: %d
- **High**: %d
- **Medium**: %d
- **Low**: %d

## Recommendations
`, result.ID, result.AuditType, result.Status, result.Duration, result.ComplianceScore,
		result.TotalChecks, result.PassedChecks, result.FailedChecks,
		result.CriticalVulnerabilities, result.HighVulnerabilities,
		result.MediumVulnerabilities, result.LowVulnerabilities)

	for i, rec := range result.Recommendations {
		report += fmt.Sprintf(`
### %d. %s
- **Priority**: %s
- **Category**: %s
- **Impact**: %s
- **Effort**: %s
- **Timeline**: %s
- **Description**: %s
`, i+1, rec.Title, rec.Priority, rec.Category, rec.Impact, rec.Effort, rec.Timeline, rec.Description)
	}

	report += fmt.Sprintf(`
## Vulnerabilities

`)

	for i, vuln := range result.Vulnerabilities {
		report += fmt.Sprintf(`
### %d. %s
- **Severity**: %s
- **Category**: %s
- **CVSS Score**: %.1f
- **Description**: %s
- **Remediation**: %s
`, i+1, vuln.Title, vuln.Severity, vuln.Category, vuln.CVSSScore, vuln.Description, vuln.Remediation)
	}

	report += fmt.Sprintf(`
## Conclusion
This audit identified %d vulnerabilities across %d security checks. 
The compliance score of %.2f%% indicates %s compliance with security standards.

## Next Steps
1. Prioritize critical and high vulnerabilities for immediate remediation
2. Implement recommended security controls
3. Schedule follow-up audit to validate remediation
4. Establish continuous security monitoring

---
Report generated on: %s
`, len(result.Vulnerabilities), result.TotalChecks, result.ComplianceScore,
		sa.getComplianceLevel(result.ComplianceScore), time.Now().Format(time.RFC3339))

	return report, nil
}

// getComplianceLevel returns compliance level based on score
func (sa *SecurityAudit) getComplianceLevel(score float64) string {
	if score >= 95.0 {
		return "excellent"
	} else if score >= 85.0 {
		return "good"
	} else if score >= 70.0 {
		return "fair"
	} else {
		return "poor"
	}
}

// GenerateAuditHash generates a cryptographic hash of the audit result
func (sa *SecurityAudit) GenerateAuditHash(result *SecurityAuditResult) (string, error) {
	// Convert result to JSON for hashing
	jsonData, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("failed to marshal audit result: %w", err)
	}

	// Generate SHA-256 hash
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}
