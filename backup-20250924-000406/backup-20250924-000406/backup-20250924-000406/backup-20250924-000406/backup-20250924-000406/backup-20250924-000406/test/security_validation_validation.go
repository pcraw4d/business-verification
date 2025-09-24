package test

import (
	"fmt"
	"strings"
)

// SecurityValidationValidator validates security test cases for 100% trusted data source usage
type SecurityValidationValidator struct {
	dataset *SecurityValidationTestDataset
}

// NewSecurityValidationValidator creates a new security validation validator
func NewSecurityValidationValidator() *SecurityValidationValidator {
	return &SecurityValidationValidator{
		dataset: NewSecurityValidationTestDataset(),
	}
}

// Validate100PercentTrustedDataSourceUsage validates that all test cases follow 100% trusted data source principles
func (svv *SecurityValidationValidator) Validate100PercentTrustedDataSourceUsage() *ValidationResult {
	result := &ValidationResult{
		TotalTestCases:     len(svv.dataset.TestCases),
		ValidTestCases:     0,
		InvalidTestCases:   0,
		SecurityViolations: []string{},
		Recommendations:    []string{},
	}

	for _, tc := range svv.dataset.TestCases {
		if svv.validateTestCase(tc) {
			result.ValidTestCases++
		} else {
			result.InvalidTestCases++
		}
	}

	// Calculate compliance percentage
	result.CompliancePercentage = float64(result.ValidTestCases) / float64(result.TotalTestCases) * 100

	// Add recommendations if not 100% compliant
	if result.CompliancePercentage < 100 {
		result.Recommendations = append(result.Recommendations,
			"Review and fix test cases that do not follow 100% trusted data source principles")
	}

	return result
}

// validateTestCase validates a single test case for security compliance
func (svv *SecurityValidationValidator) validateTestCase(tc SecurityValidationTestCase) bool {
	// Check 1: Description should always be excluded
	if !tc.ExpectedDescriptionExcluded {
		return false
	}

	// Check 2: Business name should always be trusted
	businessNameInfo, ok := tc.ExpectedDataSourceInfo["business_name"].(map[string]interface{})
	if !ok || businessNameInfo["trusted"] != true {
		return false
	}

	// Check 3: Description should never be trusted or used
	descriptionInfo, ok := tc.ExpectedDataSourceInfo["description"].(map[string]interface{})
	if !ok || descriptionInfo["trusted"] != false || descriptionInfo["used"] != false {
		return false
	}

	// Check 4: Website URL should only be used if verified
	websiteInfo, ok := tc.ExpectedDataSourceInfo["website_url"].(map[string]interface{})
	if !ok {
		return false
	}

	// Website should only be used if it's expected to be trusted
	if tc.ExpectedWebsiteTrust && websiteInfo["used"] != true {
		return false
	}
	if !tc.ExpectedWebsiteTrust && websiteInfo["used"] != false {
		return false
	}

	// Check 5: Security log messages should be present
	if len(tc.ExpectedLogMessages) == 0 {
		return false
	}

	// Check 6: All log messages should contain SECURITY keyword
	for _, log := range tc.ExpectedLogMessages {
		if !strings.Contains(log, "SECURITY") {
			return false
		}
	}

	return true
}

// ValidateSecurityCategories validates that all security categories are properly tested
func (svv *SecurityValidationValidator) ValidateSecurityCategories() *CategoryValidationResult {
	result := &CategoryValidationResult{
		Categories: make(map[string]int),
		TotalCases: len(svv.dataset.TestCases),
	}

	// Count test cases by category
	for _, tc := range svv.dataset.TestCases {
		result.Categories[tc.SecurityCategory]++
	}

	// Validate minimum requirements
	result.Requirements = map[string]bool{
		"Website Ownership Verification": result.Categories["Website Ownership Verification"] >= 10,
		"Data Source Exclusion":          result.Categories["Data Source Exclusion"] >= 5,
		"Malicious Input Handling":       result.Categories["Malicious Input Handling"] >= 5,
		"Data Source Trust Validation":   result.Categories["Data Source Trust Validation"] >= 5,
	}

	// Calculate overall compliance
	metRequirements := 0
	for _, met := range result.Requirements {
		if met {
			metRequirements++
		}
	}
	result.CompliancePercentage = float64(metRequirements) / float64(len(result.Requirements)) * 100

	return result
}

// ValidateMaliciousInputHandling validates that malicious input is properly handled
func (svv *SecurityValidationValidator) ValidateMaliciousInputHandling() *MaliciousInputValidationResult {
	result := &MaliciousInputValidationResult{
		TotalMaliciousCases: 0,
		BlockedCases:        0,
		SecurityLogsPresent: 0,
	}

	maliciousCases := svv.dataset.GetMaliciousInputTestCases()
	result.TotalMaliciousCases = len(maliciousCases)

	for _, tc := range maliciousCases {
		// Check if malicious input is properly blocked
		if tc.ExpectedDescriptionExcluded && !tc.TrustedDataSource {
			result.BlockedCases++
		}

		// Check if security logs are present
		hasSecurityLog := false
		for _, log := range tc.ExpectedLogMessages {
			if strings.Contains(log, "SECURITY") {
				hasSecurityLog = true
				break
			}
		}
		if hasSecurityLog {
			result.SecurityLogsPresent++
		}
	}

	// Calculate compliance percentages
	if result.TotalMaliciousCases > 0 {
		result.BlockingPercentage = float64(result.BlockedCases) / float64(result.TotalMaliciousCases) * 100
		result.LoggingPercentage = float64(result.SecurityLogsPresent) / float64(result.TotalMaliciousCases) * 100
	}

	return result
}

// GenerateSecurityValidationReport generates a comprehensive security validation report
func (svv *SecurityValidationValidator) GenerateSecurityValidationReport() *SecurityValidationReport {
	report := &SecurityValidationReport{
		Timestamp: fmt.Sprintf("Generated at: %s", "2024-01-01 00:00:00"), // This would be time.Now() in real implementation
	}

	// Validate 100% trusted data source usage
	report.TrustedDataSourceValidation = svv.Validate100PercentTrustedDataSourceUsage()

	// Validate security categories
	report.CategoryValidation = svv.ValidateSecurityCategories()

	// Validate malicious input handling
	report.MaliciousInputValidation = svv.ValidateMaliciousInputHandling()

	// Calculate overall security score
	report.OverallSecurityScore = (report.TrustedDataSourceValidation.CompliancePercentage +
		report.CategoryValidation.CompliancePercentage +
		report.MaliciousInputValidation.BlockingPercentage) / 3

	// Generate summary
	report.Summary = svv.generateSummary(report)

	return report
}

// generateSummary generates a summary of the security validation report
func (svv *SecurityValidationValidator) generateSummary(report *SecurityValidationReport) string {
	var summary strings.Builder

	summary.WriteString("🔒 Security Validation Report Summary\n")
	summary.WriteString("=====================================\n\n")

	// Trusted data source usage
	summary.WriteString(fmt.Sprintf("📊 Trusted Data Source Usage: %.1f%% (%d/%d test cases)\n",
		report.TrustedDataSourceValidation.CompliancePercentage,
		report.TrustedDataSourceValidation.ValidTestCases,
		report.TrustedDataSourceValidation.TotalTestCases))

	// Security categories
	summary.WriteString(fmt.Sprintf("📊 Security Category Coverage: %.1f%% (%d/%d categories)\n",
		report.CategoryValidation.CompliancePercentage,
		len(report.CategoryValidation.Requirements),
		len(report.CategoryValidation.Requirements)))

	// Malicious input handling
	summary.WriteString(fmt.Sprintf("📊 Malicious Input Blocking: %.1f%% (%d/%d cases)\n",
		report.MaliciousInputValidation.BlockingPercentage,
		report.MaliciousInputValidation.BlockedCases,
		report.MaliciousInputValidation.TotalMaliciousCases))

	// Overall security score
	summary.WriteString(fmt.Sprintf("📊 Overall Security Score: %.1f%%\n\n",
		report.OverallSecurityScore))

	// Recommendations
	if len(report.TrustedDataSourceValidation.Recommendations) > 0 {
		summary.WriteString("💡 Recommendations:\n")
		for _, rec := range report.TrustedDataSourceValidation.Recommendations {
			summary.WriteString(fmt.Sprintf("  • %s\n", rec))
		}
	}

	return summary.String()
}

// ValidationResult represents the result of security validation
type ValidationResult struct {
	TotalTestCases       int      `json:"total_test_cases"`
	ValidTestCases       int      `json:"valid_test_cases"`
	InvalidTestCases     int      `json:"invalid_test_cases"`
	CompliancePercentage float64  `json:"compliance_percentage"`
	SecurityViolations   []string `json:"security_violations"`
	Recommendations      []string `json:"recommendations"`
}

// CategoryValidationResult represents the result of category validation
type CategoryValidationResult struct {
	Categories           map[string]int  `json:"categories"`
	Requirements         map[string]bool `json:"requirements"`
	TotalCases           int             `json:"total_cases"`
	CompliancePercentage float64         `json:"compliance_percentage"`
}

// MaliciousInputValidationResult represents the result of malicious input validation
type MaliciousInputValidationResult struct {
	TotalMaliciousCases int     `json:"total_malicious_cases"`
	BlockedCases        int     `json:"blocked_cases"`
	SecurityLogsPresent int     `json:"security_logs_present"`
	BlockingPercentage  float64 `json:"blocking_percentage"`
	LoggingPercentage   float64 `json:"logging_percentage"`
}

// SecurityValidationReport represents a comprehensive security validation report
type SecurityValidationReport struct {
	Timestamp                   string                          `json:"timestamp"`
	TrustedDataSourceValidation *ValidationResult               `json:"trusted_data_source_validation"`
	CategoryValidation          *CategoryValidationResult       `json:"category_validation"`
	MaliciousInputValidation    *MaliciousInputValidationResult `json:"malicious_input_validation"`
	OverallSecurityScore        float64                         `json:"overall_security_score"`
	Summary                     string                          `json:"summary"`
}
