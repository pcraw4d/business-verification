package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSecurityValidationValidator tests the security validation validator
func TestSecurityValidationValidator(t *testing.T) {
	validator := NewSecurityValidationValidator()
	require.NotNil(t, validator, "Validator should be created successfully")
	require.NotNil(t, validator.dataset, "Dataset should be initialized")

	// Test dataset has test cases
	assert.True(t, len(validator.dataset.TestCases) > 0, "Dataset should have test cases")
}

// TestValidate100PercentTrustedDataSourceUsage tests the 100% trusted data source validation
func TestValidate100PercentTrustedDataSourceUsage(t *testing.T) {
	validator := NewSecurityValidationValidator()
	result := validator.Validate100PercentTrustedDataSourceUsage()

	// Validate result structure
	require.NotNil(t, result, "Validation result should not be nil")
	assert.True(t, result.TotalTestCases > 0, "Should have test cases")
	assert.True(t, result.ValidTestCases >= 0, "Should have valid test cases count")
	assert.True(t, result.InvalidTestCases >= 0, "Should have invalid test cases count")
	assert.True(t, result.CompliancePercentage >= 0 && result.CompliancePercentage <= 100, "Compliance percentage should be between 0 and 100")

	// Validate that we have test cases
	assert.Equal(t, 25, result.TotalTestCases, "Should have 25 test cases")

	// Validate that all test cases follow security principles
	assert.Equal(t, 25, result.ValidTestCases, "All test cases should be valid")
	assert.Equal(t, 0, result.InvalidTestCases, "No test cases should be invalid")
	assert.Equal(t, 100.0, result.CompliancePercentage, "Should have 100% compliance")

	t.Logf("📊 Trusted Data Source Validation Results:")
	t.Logf("  • Total Test Cases: %d", result.TotalTestCases)
	t.Logf("  • Valid Test Cases: %d", result.ValidTestCases)
	t.Logf("  • Invalid Test Cases: %d", result.InvalidTestCases)
	t.Logf("  • Compliance Percentage: %.1f%%", result.CompliancePercentage)
}

// TestValidateSecurityCategories tests the security category validation
func TestValidateSecurityCategories(t *testing.T) {
	validator := NewSecurityValidationValidator()
	result := validator.ValidateSecurityCategories()

	// Validate result structure
	require.NotNil(t, result, "Category validation result should not be nil")
	assert.True(t, result.TotalCases > 0, "Should have test cases")
	assert.NotNil(t, result.Categories, "Categories should not be nil")
	assert.NotNil(t, result.Requirements, "Requirements should not be nil")

	// Validate that we have the expected categories
	expectedCategories := []string{
		"Website Ownership Verification",
		"Data Source Exclusion",
		"Malicious Input Handling",
		"Data Source Trust Validation",
	}

	for _, category := range expectedCategories {
		assert.Contains(t, result.Categories, category, "Should have %s category", category)
		assert.True(t, result.Categories[category] > 0, "Category %s should have test cases", category)
	}

	// Validate minimum requirements
	assert.True(t, result.Requirements["Website Ownership Verification"], "Should have at least 10 website ownership verification test cases")
	assert.True(t, result.Requirements["Data Source Exclusion"], "Should have at least 5 data source exclusion test cases")
	assert.True(t, result.Requirements["Malicious Input Handling"], "Should have at least 5 malicious input handling test cases")
	assert.True(t, result.Requirements["Data Source Trust Validation"], "Should have at least 5 data source trust validation test cases")

	// Validate compliance percentage
	assert.Equal(t, 100.0, result.CompliancePercentage, "Should have 100% category compliance")

	t.Logf("📊 Security Category Validation Results:")
	t.Logf("  • Total Test Cases: %d", result.TotalCases)
	t.Logf("  • Categories: %d", len(result.Categories))
	t.Logf("  • Compliance Percentage: %.1f%%", result.CompliancePercentage)

	for category, count := range result.Categories {
		t.Logf("  • %s: %d test cases", category, count)
	}
}

// TestValidateMaliciousInputHandling tests the malicious input handling validation
func TestValidateMaliciousInputHandling(t *testing.T) {
	validator := NewSecurityValidationValidator()
	result := validator.ValidateMaliciousInputHandling()

	// Validate result structure
	require.NotNil(t, result, "Malicious input validation result should not be nil")
	assert.True(t, result.TotalMaliciousCases > 0, "Should have malicious input test cases")
	assert.True(t, result.BlockedCases >= 0, "Should have blocked cases count")
	assert.True(t, result.SecurityLogsPresent >= 0, "Should have security logs count")

	// Validate that all malicious inputs are blocked
	assert.Equal(t, result.TotalMaliciousCases, result.BlockedCases, "All malicious inputs should be blocked")
	assert.Equal(t, 100.0, result.BlockingPercentage, "Should have 100% blocking rate")

	// Validate that all malicious inputs have security logs
	assert.Equal(t, result.TotalMaliciousCases, result.SecurityLogsPresent, "All malicious inputs should have security logs")
	assert.Equal(t, 100.0, result.LoggingPercentage, "Should have 100% logging rate")

	t.Logf("📊 Malicious Input Handling Validation Results:")
	t.Logf("  • Total Malicious Cases: %d", result.TotalMaliciousCases)
	t.Logf("  • Blocked Cases: %d", result.BlockedCases)
	t.Logf("  • Security Logs Present: %d", result.SecurityLogsPresent)
	t.Logf("  • Blocking Percentage: %.1f%%", result.BlockingPercentage)
	t.Logf("  • Logging Percentage: %.1f%%", result.LoggingPercentage)
}

// TestGenerateSecurityValidationReport tests the comprehensive security validation report
func TestGenerateSecurityValidationReport(t *testing.T) {
	validator := NewSecurityValidationValidator()
	report := validator.GenerateSecurityValidationReport()

	// Validate report structure
	require.NotNil(t, report, "Security validation report should not be nil")
	assert.NotEmpty(t, report.Timestamp, "Report should have timestamp")
	assert.NotNil(t, report.TrustedDataSourceValidation, "Report should have trusted data source validation")
	assert.NotNil(t, report.CategoryValidation, "Report should have category validation")
	assert.NotNil(t, report.MaliciousInputValidation, "Report should have malicious input validation")
	assert.NotEmpty(t, report.Summary, "Report should have summary")

	// Validate overall security score
	assert.True(t, report.OverallSecurityScore >= 0 && report.OverallSecurityScore <= 100, "Overall security score should be between 0 and 100")
	assert.Equal(t, 100.0, report.OverallSecurityScore, "Should have 100% overall security score")

	// Validate individual components
	assert.Equal(t, 100.0, report.TrustedDataSourceValidation.CompliancePercentage, "Should have 100% trusted data source compliance")
	assert.Equal(t, 100.0, report.CategoryValidation.CompliancePercentage, "Should have 100% category compliance")
	assert.Equal(t, 100.0, report.MaliciousInputValidation.BlockingPercentage, "Should have 100% malicious input blocking")

	t.Logf("📊 Security Validation Report:")
	t.Logf("  • Overall Security Score: %.1f%%", report.OverallSecurityScore)
	t.Logf("  • Trusted Data Source Compliance: %.1f%%", report.TrustedDataSourceValidation.CompliancePercentage)
	t.Logf("  • Category Compliance: %.1f%%", report.CategoryValidation.CompliancePercentage)
	t.Logf("  • Malicious Input Blocking: %.1f%%", report.MaliciousInputValidation.BlockingPercentage)
	t.Logf("\n%s", report.Summary)
}

// TestSecurityValidationComprehensiveIntegration tests the comprehensive integration of security validation
func TestSecurityValidationComprehensiveIntegration(t *testing.T) {
	validator := NewSecurityValidationValidator()

	// Test all validation methods
	trustedDataSourceResult := validator.Validate100PercentTrustedDataSourceUsage()
	categoryResult := validator.ValidateSecurityCategories()
	maliciousInputResult := validator.ValidateMaliciousInputHandling()
	report := validator.GenerateSecurityValidationReport()

	// Validate that all results are consistent
	assert.Equal(t, 25, trustedDataSourceResult.TotalTestCases, "All validations should have same total test cases")
	assert.Equal(t, 25, categoryResult.TotalCases, "All validations should have same total test cases")
	assert.True(t, maliciousInputResult.TotalMaliciousCases > 0, "Should have malicious input test cases")

	// Validate that all validations pass
	assert.Equal(t, 100.0, trustedDataSourceResult.CompliancePercentage, "Trusted data source validation should pass")
	assert.Equal(t, 100.0, categoryResult.CompliancePercentage, "Category validation should pass")
	assert.Equal(t, 100.0, maliciousInputResult.BlockingPercentage, "Malicious input validation should pass")
	assert.Equal(t, 100.0, report.OverallSecurityScore, "Overall security score should be 100%")

	// Validate that the report is comprehensive
	assert.Contains(t, report.Summary, "Security Validation Report Summary", "Report summary should contain title")
	assert.Contains(t, report.Summary, "Trusted Data Source Usage", "Report should contain trusted data source info")
	assert.Contains(t, report.Summary, "Security Category Coverage", "Report should contain category info")
	assert.Contains(t, report.Summary, "Malicious Input Blocking", "Report should contain malicious input info")
	assert.Contains(t, report.Summary, "Overall Security Score", "Report should contain overall score")

	t.Logf("✅ All security validation tests passed successfully!")
	t.Logf("📊 Comprehensive Security Validation Results:")
	t.Logf("  • 100%% Trusted Data Source Usage: ✅")
	t.Logf("  • Security Category Coverage: ✅")
	t.Logf("  • Malicious Input Handling: ✅")
	t.Logf("  • Overall Security Score: 100%%")
}
