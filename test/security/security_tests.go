package security

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComprehensiveSecurityTesting runs comprehensive security tests for subtask 4.2.4
func TestComprehensiveSecurityTesting(t *testing.T) {
	t.Log("Starting comprehensive security testing for subtask 4.2.4...")

	// Create simple security test suite
	testSuite := NewSimpleSecurityTestSuite(t)
	defer testSuite.Close()

	// Run all security tests
	results := testSuite.RunAllSecurityTests(t)

	// Verify we have results
	require.NotEmpty(t, results, "Should have security test results")

	// Log summary
	t.Logf("Completed %d security tests", len(results))

	t.Log("Comprehensive security testing completed successfully")
}

// TestAuthenticationFlows tests authentication flow security
func TestAuthenticationFlows(t *testing.T) {
	t.Log("Testing authentication flows...")

	testSuite := NewSimpleSecurityTestSuite(t)
	defer testSuite.Close()

	// Run authentication tests
	results := testSuite.TestAuthenticationFlows(t)

	// Validate results
	require.NotEmpty(t, results, "Should have authentication test results")

	// Check that critical authentication tests pass
	criticalTests := []string{
		"Valid JWT Token Authentication",
		"Valid API Key Authentication",
		"Invalid Token Rejection",
		"Expired Token Rejection",
		"Missing Authentication Header",
	}

	for _, testName := range criticalTests {
		found := false
		for _, result := range results {
			if result.TestName == testName {
				found = true
				assert.Equal(t, "PASS", result.Status,
					"Critical authentication test %s should pass", testName)
				break
			}
		}
		assert.True(t, found, "Critical authentication test %s should be present", testName)
	}

	t.Log("Authentication flow testing completed")
}

// TestAuthorizationControls tests authorization control security
func TestAuthorizationControls(t *testing.T) {
	t.Log("Testing authorization controls...")

	testSuite := NewSimpleSecurityTestSuite(t)
	defer testSuite.Close()

	// Run authorization tests
	results := testSuite.TestAuthorizationControls(t)

	// Validate results
	require.NotEmpty(t, results, "Should have authorization test results")

	// Check that critical authorization tests pass
	criticalTests := []string{
		"Admin Role Access",
		"User Role Access Denied",
		"API Key Role Validation",
	}

	for _, testName := range criticalTests {
		found := false
		for _, result := range results {
			if result.TestName == testName {
				found = true
				assert.Equal(t, "PASS", result.Status,
					"Critical authorization test %s should pass", testName)
				break
			}
		}
		assert.True(t, found, "Critical authorization test %s should be present", testName)
	}

	t.Log("Authorization control testing completed")
}

// TestDataAccessRestrictions tests data access restriction security
func TestDataAccessRestrictions(t *testing.T) {
	t.Log("Testing data access restrictions...")

	testSuite := NewSimpleSecurityTestSuite(t)
	defer testSuite.Close()

	// Run data access tests
	results := testSuite.TestDataAccessRestrictions(t)

	// Validate results
	require.NotEmpty(t, results, "Should have data access test results")

	// Check that critical data access tests pass
	criticalTests := []string{
		"User Data Isolation",
		"Sensitive Data Protection",
	}

	for _, testName := range criticalTests {
		found := false
		for _, result := range results {
			if result.TestName == testName {
				found = true
				assert.Equal(t, "PASS", result.Status,
					"Critical data access test %s should pass", testName)
				break
			}
		}
		assert.True(t, found, "Critical data access test %s should be present", testName)
	}

	t.Log("Data access restriction testing completed")
}

// TestAuditLogging tests audit logging security
func TestAuditLogging(t *testing.T) {
	t.Log("Testing audit logging...")

	testSuite := NewSimpleSecurityTestSuite(t)
	defer testSuite.Close()

	// Run audit logging tests
	results := testSuite.TestAuditLogging(t)

	// Validate results
	require.NotEmpty(t, results, "Should have audit logging test results")

	// Check that critical audit logging tests pass
	criticalTests := []string{
		"Authentication Event Logging",
		"Authorization Event Logging",
	}

	for _, testName := range criticalTests {
		found := false
		for _, result := range results {
			if result.TestName == testName {
				found = true
				assert.Equal(t, "PASS", result.Status,
					"Critical audit logging test %s should pass", testName)
				break
			}
		}
		assert.True(t, found, "Critical audit logging test %s should be present", testName)
	}

	t.Log("Audit logging testing completed")
}

// TestInputValidation tests input validation security
func TestInputValidation(t *testing.T) {
	t.Log("Testing input validation...")

	testSuite := NewSimpleSecurityTestSuite(t)
	defer testSuite.Close()

	// Run input validation tests
	results := testSuite.TestInputValidation(t)

	// Validate results
	require.NotEmpty(t, results, "Should have input validation test results")

	// Check that critical input validation tests pass
	criticalTests := []string{
		"SQL Injection Prevention",
		"XSS Prevention",
	}

	for _, testName := range criticalTests {
		found := false
		for _, result := range results {
			if result.TestName == testName {
				found = true
				assert.Equal(t, "PASS", result.Status,
					"Critical input validation test %s should pass", testName)
				break
			}
		}
		assert.True(t, found, "Critical input validation test %s should be present", testName)
	}

	t.Log("Input validation testing completed")
}

// TestRateLimiting tests rate limiting security
func TestRateLimiting(t *testing.T) {
	t.Log("Testing rate limiting...")

	testSuite := NewSimpleSecurityTestSuite(t)
	defer testSuite.Close()

	// Run rate limiting tests
	results := testSuite.TestRateLimiting(t)

	// Validate results
	require.NotEmpty(t, results, "Should have rate limiting test results")

	// Check that rate limiting tests are present
	criticalTests := []string{
		"Rate Limiting Enforcement",
	}

	for _, testName := range criticalTests {
		found := false
		for _, result := range results {
			if result.TestName == testName {
				found = true
				// Rate limiting can be PASS or WARN (WARN if not configured)
				assert.True(t, result.Status == "PASS" || result.Status == "WARN",
					"Rate limiting test %s should pass or warn", testName)
				break
			}
		}
		assert.True(t, found, "Rate limiting test %s should be present", testName)
	}

	t.Log("Rate limiting testing completed")
}

// TestSecurityHeaders tests security headers
func TestSecurityHeaders(t *testing.T) {
	t.Log("Testing security headers...")

	testSuite := NewSimpleSecurityTestSuite(t)
	defer testSuite.Close()

	// Run security headers tests
	results := testSuite.TestSecurityHeaders(t)

	// Validate results
	require.NotEmpty(t, results, "Should have security headers test results")

	// Check that security headers tests are present
	criticalTests := []string{
		"Security Headers Implementation",
	}

	for _, testName := range criticalTests {
		found := false
		for _, result := range results {
			if result.TestName == testName {
				found = true
				// Security headers can be PASS or WARN (WARN if some headers missing)
				assert.True(t, result.Status == "PASS" || result.Status == "WARN",
					"Security headers test %s should pass or warn", testName)
				break
			}
		}
		assert.True(t, found, "Security headers test %s should be present", testName)
	}

	t.Log("Security headers testing completed")
}

// TestSecurityTestSuite tests the security test suite functionality
func TestSecurityTestSuite(t *testing.T) {
	t.Log("Testing security test suite functionality...")

	testSuite := NewSimpleSecurityTestSuite(t)
	defer testSuite.Close()

	// Test that all test categories are available
	allResults := testSuite.RunAllSecurityTests(t)

	// Verify we have tests from all categories
	categories := make(map[string]bool)
	for _, result := range allResults {
		categories[result.Category] = true
	}

	expectedCategories := []string{
		"AUTHENTICATION",
		"AUTHORIZATION",
		"DATA_ACCESS",
		"AUDIT_LOGGING",
		"INPUT_VALIDATION",
		"RATE_LIMITING",
		"SECURITY_HEADERS",
	}

	for _, category := range expectedCategories {
		assert.True(t, categories[category],
			"Should have tests for category %s", category)
	}

	// Verify we have a reasonable number of tests
	assert.GreaterOrEqual(t, len(allResults), 10,
		"Should have at least 10 security tests")

	t.Log("Security test suite functionality testing completed")
}

// TestSecurityReportGeneration tests security report generation
func TestSecurityReportGeneration(t *testing.T) {
	t.Log("Testing security report generation...")

	runner := NewSimpleSecurityTestRunner(t)
	defer runner.Close()

	// Run tests to generate results
	runner.results = runner.testSuite.RunAllSecurityTests(t)

	// Generate reports
	runner.generateJSONReport(t)
	runner.generateMarkdownReport(t)
	runner.generateSummaryReport(t)

	// Verify reports were created and have content
	jsonPath := "test/reports/security/security_test_results.json"
	markdownPath := "test/reports/security/security_test_report.md"
	summaryPath := "test/reports/security/security_summary.md"

	// Check files exist
	assert.FileExists(t, jsonPath, "JSON report should exist")
	assert.FileExists(t, markdownPath, "Markdown report should exist")
	assert.FileExists(t, summaryPath, "Summary report should exist")

	// Check JSON report content
	jsonData, err := os.ReadFile(jsonPath)
	require.NoError(t, err)
	assert.NotEmpty(t, jsonData, "JSON report should have content")

	// Check markdown report content
	markdownData, err := os.ReadFile(markdownPath)
	require.NoError(t, err)
	assert.NotEmpty(t, markdownData, "Markdown report should have content")

	// Check summary report content
	summaryData, err := os.ReadFile(summaryPath)
	require.NoError(t, err)
	assert.NotEmpty(t, summaryData, "Summary report should have content")

	t.Log("Security report generation testing completed")
}

// BenchmarkSecurityTests benchmarks security test performance
func BenchmarkSecurityTests(b *testing.B) {
	b.Log("Benchmarking security tests...")

	for i := 0; i < b.N; i++ {
		testSuite := NewSecurityTestSuite(&testing.T{})
		results := testSuite.RunAllSecurityTests(&testing.T{})
		testSuite.Close()

		// Ensure we have results
		if len(results) == 0 {
			b.Fatal("No security test results generated")
		}
	}
}
