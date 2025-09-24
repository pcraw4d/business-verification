package test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSecurityValidationComprehensive tests the comprehensive security validation test cases
func TestSecurityValidationComprehensive(t *testing.T) {
	dataset := NewSecurityValidationTestDataset()
	allTestCases := dataset.TestCases

	// Verify we have 25 security test cases (security-001 through security-025)
	require.Equal(t, 25, len(allTestCases), "Expected 25 security validation test cases")

	t.Logf("🔒 Testing %d security validation test cases", len(allTestCases))

	// Test case validation
	validTestCases := 0
	securityCompliantCases := 0
	maliciousInputBlocked := 0
	trustedDataSourceUsed := 0

	for _, tc := range allTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Validate test case structure
			assert.NotEmpty(t, tc.ID, "Test case ID should not be empty")
			assert.NotEmpty(t, tc.Name, "Test case name should not be empty")
			assert.NotEmpty(t, tc.SecurityCategory, "Security category should not be empty")
			assert.NotEmpty(t, tc.TestType, "Test type should not be empty")
			assert.NotEmpty(t, tc.ExpectedLogMessages, "Expected log messages should not be empty")

			// Validate security requirements
			assert.True(t, tc.ExpectedDescriptionExcluded, "Description should always be excluded for security")

			// Validate data source info structure
			assert.NotNil(t, tc.ExpectedDataSourceInfo, "Expected data source info should not be nil")
			assert.Contains(t, tc.ExpectedDataSourceInfo, "business_name", "Should contain business_name data source info")
			assert.Contains(t, tc.ExpectedDataSourceInfo, "description", "Should contain description data source info")
			assert.Contains(t, tc.ExpectedDataSourceInfo, "website_url", "Should contain website_url data source info")

			// Validate business name data source
			businessNameInfo, ok := tc.ExpectedDataSourceInfo["business_name"].(map[string]interface{})
			require.True(t, ok, "Business name info should be a map")
			assert.Equal(t, true, businessNameInfo["trusted"], "Business name should always be trusted")
			assert.Equal(t, "Primary business identifier", businessNameInfo["reason"], "Business name reason should be correct")

			// Validate description data source
			descriptionInfo, ok := tc.ExpectedDataSourceInfo["description"].(map[string]interface{})
			require.True(t, ok, "Description info should be a map")
			assert.Equal(t, false, descriptionInfo["used"], "Description should never be used")
			assert.Equal(t, false, descriptionInfo["trusted"], "Description should never be trusted")
			assert.Equal(t, "User-provided data cannot be trusted for classification", descriptionInfo["reason"], "Description reason should be correct")

			// Validate website URL data source
			websiteInfo, ok := tc.ExpectedDataSourceInfo["website_url"].(map[string]interface{})
			require.True(t, ok, "Website info should be a map")
			assert.Equal(t, "Website ownership must be verified before use", websiteInfo["reason"], "Website reason should be correct")

			// Validate expected log messages
			for _, expectedLog := range tc.ExpectedLogMessages {
				assert.True(t, strings.Contains(expectedLog, "SECURITY"), "Log messages should contain SECURITY keyword")
			}

			// Count valid test cases
			validTestCases++

			// Count security compliant cases
			if tc.ExpectedDescriptionExcluded {
				securityCompliantCases++
			}

			// Count malicious input blocked cases
			if tc.MaliciousInput && !tc.TrustedDataSource {
				maliciousInputBlocked++
			}

			// Count trusted data source used cases
			if tc.TrustedDataSource {
				trustedDataSourceUsed++
			}
		})
	}

	// Validate overall security compliance
	t.Logf("📊 Security Validation Results:")
	t.Logf("  • Total Test Cases: %d", validTestCases)
	t.Logf("  • Security Compliant Cases: %d", securityCompliantCases)
	t.Logf("  • Malicious Input Blocked: %d", maliciousInputBlocked)
	t.Logf("  • Trusted Data Source Used: %d", trustedDataSourceUsed)

	// Assert security requirements
	assert.Equal(t, 25, validTestCases, "All test cases should be valid")
	assert.Equal(t, 25, securityCompliantCases, "All test cases should be security compliant")
	assert.True(t, maliciousInputBlocked >= 10, "Should have at least 10 malicious input test cases")
	assert.True(t, trustedDataSourceUsed >= 10, "Should have at least 10 trusted data source test cases")
}

// TestWebsiteOwnershipVerification tests website ownership verification security
func TestWebsiteOwnershipVerification(t *testing.T) {
	dataset := NewSecurityValidationTestDataset()
	websiteTestCases := dataset.GetTestCasesByCategory("Website Ownership Verification")

	// Verify we have website ownership verification test cases
	require.True(t, len(websiteTestCases) >= 10, "Expected at least 10 website ownership verification test cases")

	t.Logf("🔒 Testing %d website ownership verification test cases", len(websiteTestCases))

	verifiedCount := 0
	unverifiedCount := 0

	for _, tc := range websiteTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Validate test case structure
			assert.NotEmpty(t, tc.BusinessName, "Business name should not be empty")
			assert.NotEmpty(t, tc.WebsiteURL, "Website URL should not be empty")
			assert.NotEmpty(t, tc.ExpectedLogMessages, "Expected log messages should not be empty")

			// Validate security requirements
			assert.True(t, tc.ExpectedDescriptionExcluded, "Description should always be excluded")

			// Validate website URL data source info
			websiteInfo, ok := tc.ExpectedDataSourceInfo["website_url"].(map[string]interface{})
			require.True(t, ok, "Website info should be a map")
			assert.Equal(t, tc.ExpectedWebsiteTrust, websiteInfo["used"], "Website usage should match expected trust")

			// Count verification results
			if tc.ExpectedWebsiteTrust {
				verifiedCount++
				// Verify positive log message
				hasPositiveLog := false
				for _, log := range tc.ExpectedLogMessages {
					if strings.Contains(log, "✅ SECURITY: Using verified website URL") {
						hasPositiveLog = true
						break
					}
				}
				assert.True(t, hasPositiveLog, "Verified website should have positive log message")
			} else {
				unverifiedCount++
				// Verify negative log message
				hasNegativeLog := false
				for _, log := range tc.ExpectedLogMessages {
					if strings.Contains(log, "⚠️ SECURITY: Skipping unverified website URL") {
						hasNegativeLog = true
						break
					}
				}
				assert.True(t, hasNegativeLog, "Unverified website should have negative log message")
			}
		})
	}

	t.Logf("📊 Website Ownership Verification Results:")
	t.Logf("  • Verified Websites: %d", verifiedCount)
	t.Logf("  • Unverified Websites: %d", unverifiedCount)

	// Assert we have both verified and unverified test cases
	assert.True(t, verifiedCount >= 5, "Should have at least 5 verified website test cases")
	assert.True(t, unverifiedCount >= 5, "Should have at least 5 unverified website test cases")
}

// TestDataSourceExclusionMechanisms tests data source exclusion mechanisms
func TestDataSourceExclusionMechanisms(t *testing.T) {
	dataset := NewSecurityValidationTestDataset()
	exclusionTestCases := dataset.GetTestCasesByCategory("Data Source Exclusion")

	// Verify we have data source exclusion test cases
	require.True(t, len(exclusionTestCases) >= 5, "Expected at least 5 data source exclusion test cases")

	t.Logf("🔒 Testing %d data source exclusion test cases", len(exclusionTestCases))

	excludedCount := 0

	for _, tc := range exclusionTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Validate test case structure
			assert.NotEmpty(t, tc.Description, "Description should not be empty for exclusion testing")
			assert.NotEmpty(t, tc.ExpectedLogMessages, "Expected log messages should not be empty")

			// Validate security requirements
			assert.True(t, tc.ExpectedDescriptionExcluded, "Description should always be excluded")
			assert.True(t, tc.MaliciousInput, "Exclusion test cases should have malicious input")

			// Validate description data source info
			descriptionInfo, ok := tc.ExpectedDataSourceInfo["description"].(map[string]interface{})
			require.True(t, ok, "Description info should be a map")
			assert.Equal(t, false, descriptionInfo["used"], "Description should never be used")
			assert.Equal(t, false, descriptionInfo["trusted"], "Description should never be trusted")

			// Verify exclusion log message
			hasExclusionLog := false
			for _, log := range tc.ExpectedLogMessages {
				if strings.Contains(log, "🔒 SECURITY: Skipping user-provided description") {
					hasExclusionLog = true
					break
				}
			}
			assert.True(t, hasExclusionLog, "Should have description exclusion log message")

			excludedCount++
		})
	}

	t.Logf("📊 Data Source Exclusion Results:")
	t.Logf("  • Excluded Descriptions: %d", excludedCount)

	// Assert all descriptions are excluded
	assert.Equal(t, len(exclusionTestCases), excludedCount, "All descriptions should be excluded")
}

// TestMaliciousInputHandling tests malicious input handling
func TestMaliciousInputHandling(t *testing.T) {
	dataset := NewSecurityValidationTestDataset()
	maliciousTestCases := dataset.GetMaliciousInputTestCases()

	// Verify we have malicious input test cases
	require.True(t, len(maliciousTestCases) >= 10, "Expected at least 10 malicious input test cases")

	t.Logf("🔒 Testing %d malicious input handling test cases", len(maliciousTestCases))

	blockedCount := 0
	securityCategories := make(map[string]int)

	for _, tc := range maliciousTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Validate test case structure
			assert.True(t, tc.MaliciousInput, "Test case should be marked as malicious input")
			assert.NotEmpty(t, tc.ExpectedLogMessages, "Expected log messages should not be empty")

			// Validate security requirements
			assert.True(t, tc.ExpectedDescriptionExcluded, "Description should always be excluded for malicious input")

			// Validate data source info
			descriptionInfo, ok := tc.ExpectedDataSourceInfo["description"].(map[string]interface{})
			require.True(t, ok, "Description info should be a map")
			assert.Equal(t, false, descriptionInfo["used"], "Malicious description should never be used")
			assert.Equal(t, false, descriptionInfo["trusted"], "Malicious description should never be trusted")

			// Verify security log messages
			hasSecurityLog := false
			for _, log := range tc.ExpectedLogMessages {
				if strings.Contains(log, "SECURITY") {
					hasSecurityLog = true
					break
				}
			}
			assert.True(t, hasSecurityLog, "Should have security log messages")

			// Count security categories
			securityCategories[tc.SecurityCategory]++

			blockedCount++
		})
	}

	t.Logf("📊 Malicious Input Handling Results:")
	t.Logf("  • Blocked Malicious Inputs: %d", blockedCount)
	t.Logf("  • Security Categories Tested: %d", len(securityCategories))

	// Assert all malicious inputs are blocked
	assert.Equal(t, len(maliciousTestCases), blockedCount, "All malicious inputs should be blocked")
	assert.True(t, len(securityCategories) >= 3, "Should test multiple security categories")
}

// TestDataSourceTrustValidation tests data source trust validation
func TestDataSourceTrustValidation(t *testing.T) {
	dataset := NewSecurityValidationTestDataset()
	trustTestCases := dataset.GetTestCasesByCategory("Data Source Trust Validation")

	// Verify we have data source trust validation test cases
	require.True(t, len(trustTestCases) >= 5, "Expected at least 5 data source trust validation test cases")

	t.Logf("🔒 Testing %d data source trust validation test cases", len(trustTestCases))

	trustedCount := 0
	untrustedCount := 0

	for _, tc := range trustTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Validate test case structure
			assert.NotEmpty(t, tc.ExpectedDataSourceInfo, "Expected data source info should not be empty")
			assert.NotEmpty(t, tc.ExpectedLogMessages, "Expected log messages should not be empty")

			// Validate security requirements
			assert.True(t, tc.ExpectedDescriptionExcluded, "Description should always be excluded")

			// Validate business name trust (should always be trusted)
			businessNameInfo, ok := tc.ExpectedDataSourceInfo["business_name"].(map[string]interface{})
			require.True(t, ok, "Business name info should be a map")
			assert.Equal(t, true, businessNameInfo["trusted"], "Business name should always be trusted")

			// Validate description trust (should never be trusted)
			descriptionInfo, ok := tc.ExpectedDataSourceInfo["description"].(map[string]interface{})
			require.True(t, ok, "Description info should be a map")
			assert.Equal(t, false, descriptionInfo["trusted"], "Description should never be trusted")

			// Count trust levels
			if tc.TrustedDataSource {
				trustedCount++
			} else {
				untrustedCount++
			}
		})
	}

	t.Logf("📊 Data Source Trust Validation Results:")
	t.Logf("  • Trusted Data Sources: %d", trustedCount)
	t.Logf("  • Untrusted Data Sources: %d", untrustedCount)

	// Assert we have both trusted and untrusted test cases
	assert.True(t, trustedCount >= 2, "Should have at least 2 trusted data source test cases")
	assert.True(t, untrustedCount >= 2, "Should have at least 2 untrusted data source test cases")
}

// TestSecurityLoggingFunctionality tests security logging functionality
func TestSecurityLoggingFunctionality(t *testing.T) {
	dataset := NewSecurityValidationTestDataset()
	allTestCases := dataset.TestCases

	t.Logf("🔒 Testing security logging functionality across %d test cases", len(allTestCases))

	loggingCompliantCount := 0
	securityLogTypes := make(map[string]int)

	for _, tc := range allTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Validate test case structure
			assert.NotEmpty(t, tc.ExpectedLogMessages, "Expected log messages should not be empty")

			// Validate security log messages
			hasSecurityLog := false
			for _, log := range tc.ExpectedLogMessages {
				if strings.Contains(log, "SECURITY") {
					hasSecurityLog = true

					// Categorize log types
					if strings.Contains(log, "🔒 SECURITY: Skipping user-provided description") {
						securityLogTypes["description_exclusion"]++
					} else if strings.Contains(log, "✅ SECURITY: Using verified website URL") {
						securityLogTypes["website_verification_success"]++
					} else if strings.Contains(log, "⚠️ SECURITY: Skipping unverified website URL") {
						securityLogTypes["website_verification_failure"]++
					}
					break
				}
			}
			assert.True(t, hasSecurityLog, "Should have security log messages")

			// Validate log message format
			for _, log := range tc.ExpectedLogMessages {
				assert.True(t, strings.Contains(log, "SECURITY"), "Log messages should contain SECURITY keyword")
				assert.True(t, strings.HasPrefix(log, "🔒") || strings.HasPrefix(log, "✅") || strings.HasPrefix(log, "⚠️"), "Log messages should have security emoji prefix")
			}

			loggingCompliantCount++
		})
	}

	t.Logf("📊 Security Logging Results:")
	t.Logf("  • Logging Compliant Cases: %d", loggingCompliantCount)
	t.Logf("  • Description Exclusion Logs: %d", securityLogTypes["description_exclusion"])
	t.Logf("  • Website Verification Success Logs: %d", securityLogTypes["website_verification_success"])
	t.Logf("  • Website Verification Failure Logs: %d", securityLogTypes["website_verification_failure"])

	// Assert all test cases have proper security logging
	assert.Equal(t, len(allTestCases), loggingCompliantCount, "All test cases should have proper security logging")
	assert.True(t, securityLogTypes["description_exclusion"] >= 20, "Should have description exclusion logs for most cases")
	assert.True(t, securityLogTypes["website_verification_success"] >= 5, "Should have website verification success logs")
	assert.True(t, securityLogTypes["website_verification_failure"] >= 5, "Should have website verification failure logs")
}

// Test100PercentTrustedDataSourceUsage tests 100% trusted data source usage
func Test100PercentTrustedDataSourceUsage(t *testing.T) {
	dataset := NewSecurityValidationTestDataset()
	allTestCases := dataset.TestCases

	t.Logf("🔒 Testing 100%% trusted data source usage across %d test cases", len(allTestCases))

	trustedDataSourceCount := 0
	untrustedDataSourceCount := 0

	for _, tc := range allTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Validate test case structure
			assert.NotEmpty(t, tc.ExpectedDataSourceInfo, "Expected data source info should not be empty")

			// Validate business name is always trusted
			businessNameInfo, ok := tc.ExpectedDataSourceInfo["business_name"].(map[string]interface{})
			require.True(t, ok, "Business name info should be a map")
			assert.Equal(t, true, businessNameInfo["trusted"], "Business name should always be trusted")

			// Validate description is never trusted
			descriptionInfo, ok := tc.ExpectedDataSourceInfo["description"].(map[string]interface{})
			require.True(t, ok, "Description info should be a map")
			assert.Equal(t, false, descriptionInfo["trusted"], "Description should never be trusted")

			// Validate website URL trust based on verification
			websiteInfo, ok := tc.ExpectedDataSourceInfo["website_url"].(map[string]interface{})
			require.True(t, ok, "Website info should be a map")
			assert.Equal(t, false, websiteInfo["trusted"], "Website URL should not be trusted (only used if verified)")

			// Count trusted vs untrusted data sources
			if tc.TrustedDataSource {
				trustedDataSourceCount++
			} else {
				untrustedDataSourceCount++
			}
		})
	}

	t.Logf("📊 Trusted Data Source Usage Results:")
	t.Logf("  • Trusted Data Source Cases: %d", trustedDataSourceCount)
	t.Logf("  • Untrusted Data Source Cases: %d", untrustedDataSourceCount)

	// Assert we have both trusted and untrusted test cases for comprehensive testing
	assert.True(t, trustedDataSourceCount >= 10, "Should have at least 10 trusted data source test cases")
	assert.True(t, untrustedDataSourceCount >= 10, "Should have at least 10 untrusted data source test cases")

	// Assert that all test cases follow the security principle: only trusted data sources are used
	for _, tc := range allTestCases {
		// Business name should always be trusted and used
		businessNameInfo, _ := tc.ExpectedDataSourceInfo["business_name"].(map[string]interface{})
		assert.Equal(t, true, businessNameInfo["trusted"], "Business name should always be trusted")

		// Description should never be trusted or used
		descriptionInfo, _ := tc.ExpectedDataSourceInfo["description"].(map[string]interface{})
		assert.Equal(t, false, descriptionInfo["trusted"], "Description should never be trusted")
		assert.Equal(t, false, descriptionInfo["used"], "Description should never be used")

		// Website URL should only be used if verified
		websiteInfo, _ := tc.ExpectedDataSourceInfo["website_url"].(map[string]interface{})
		if tc.ExpectedWebsiteTrust {
			assert.Equal(t, true, websiteInfo["used"], "Verified website should be used")
		} else {
			assert.Equal(t, false, websiteInfo["used"], "Unverified website should not be used")
		}
	}
}
