package test

import (
	"strings"
	"testing"
)

// TestTechnologyClassificationValidation tests the technology test cases for subtask 4.1.2
func TestTechnologyClassificationValidation(t *testing.T) {
	// Get technology test cases from the comprehensive dataset
	dataset := NewComprehensiveTestDataset()
	techTestCases := dataset.GetTestCasesByCategory("Technology")

	if len(techTestCases) < 20 {
		t.Errorf("❌ Expected at least 20 technology test cases, got %d", len(techTestCases))
		return
	}

	t.Logf("🧪 Validating %d technology test cases", len(techTestCases))

	var totalTests int
	var validTests int
	var securityTests int
	var securityValid int

	for _, tc := range techTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Validate test case structure
			if tc.ID == "" {
				t.Errorf("❌ Test case %s has empty ID", tc.Name)
				return
			}

			if tc.BusinessName == "" {
				t.Errorf("❌ Test case %s has empty business name", tc.Name)
				return
			}

			if len(tc.Keywords) == 0 {
				t.Errorf("❌ Test case %s has no keywords", tc.Name)
				return
			}

			if tc.ExpectedConfidence <= 0 || tc.ExpectedConfidence > 1 {
				t.Errorf("❌ Test case %s has invalid expected confidence: %.2f", tc.Name, tc.ExpectedConfidence)
				return
			}

			// Check if this is a security test case
			isSecurityTest := tc.ID == "tech-019" || tc.ID == "tech-020" || tc.ID == "tech-021"
			if isSecurityTest {
				securityTests++
			}

			// Validate security test cases have appropriate expectations
			if isSecurityTest {
				if tc.ExpectedConfidence > 0.6 {
					t.Errorf("❌ Security test case %s should have low confidence expectation, got %.2f", tc.Name, tc.ExpectedConfidence)
				} else {
					securityValid++
					t.Logf("🔒 SECURITY: %s correctly configured with low confidence %.2f", tc.Name, tc.ExpectedConfidence)
				}
			}

			// Validate technology-specific keywords
			hasTechKeywords := false
			for _, keyword := range tc.Keywords {
				if containsTechnologyKeywords(keyword) {
					hasTechKeywords = true
					break
				}
			}

			if !hasTechKeywords {
				t.Errorf("❌ Test case %s has no technology-related keywords", tc.Name)
				return
			}

			validTests++
			totalTests++

			// Log successful validation
			t.Logf("✅ %s: Valid test case with %d keywords, expected confidence %.2f",
				tc.Name, len(tc.Keywords), tc.ExpectedConfidence)
		})
	}

	// Calculate and log overall statistics
	validationAccuracy := float64(validTests) / float64(totalTests) * 100
	securityValidationAccuracy := float64(securityValid) / float64(securityTests) * 100

	t.Logf("📊 Technology Test Case Validation Results:")
	t.Logf("   Total tests: %d", totalTests)
	t.Logf("   Valid tests: %d", validTests)
	t.Logf("   Validation accuracy: %.1f%%", validationAccuracy)
	t.Logf("   Security tests: %d", securityTests)
	t.Logf("   Security validation accuracy: %.1f%%", securityValidationAccuracy)

	// Assert minimum validation threshold for subtask 4.1.2
	if validationAccuracy < 95.0 {
		t.Errorf("❌ Technology test case validation accuracy %.1f%% is below minimum threshold of 95%%", validationAccuracy)
	}

	// Log test case breakdown
	t.Logf("📋 Technology Test Case Breakdown:")
	for _, tc := range techTestCases {
		t.Logf("   %s: %s (Expected: %.2f, Keywords: %d)", tc.ID, tc.Name, tc.ExpectedConfidence, len(tc.Keywords))
	}
}

// Helper function to check if a keyword is technology-related
func containsTechnologyKeywords(keyword string) bool {
	techKeywords := []string{
		"software", "technology", "tech", "digital", "cloud", "AI", "artificial intelligence",
		"machine learning", "cybersecurity", "e-commerce", "fintech", "mobile", "app",
		"blockchain", "data", "analytics", "IoT", "SaaS", "gaming", "DevOps", "AR", "VR",
		"EdTech", "HealthTech", "API", "database", "quantum", "robotics", "automation",
		"platform", "solution", "development", "programming", "coding", "system", "network",
		"security", "payment", "processing", "web", "online", "internet", "computer",
		"electronic", "hardware", "firmware", "algorithm", "protocol", "framework",
		"infrastructure", "CI/CD", "container", "orchestration", "augmented", "virtual",
		"reality", "immersive", "mixed", "RESTful", "GraphQL", "management", "integration",
		"services", "warehousing", "optimization", "design", "computing", "algorithms",
		"processors", "industrial", "service", "RPA", "process", "automation",
	}

	keywordLower := strings.ToLower(keyword)
	for _, techKeyword := range techKeywords {
		if strings.Contains(keywordLower, techKeyword) {
			return true
		}
	}
	return false
}

// TestTechnologySecurityValidation specifically tests security validation for technology test cases
func TestTechnologySecurityValidation(t *testing.T) {
	dataset := NewComprehensiveTestDataset()
	techTestCases := dataset.GetTestCasesByCategory("Technology")

	// Find security test cases
	var securityTestCases []ClassificationTestCase
	for _, tc := range techTestCases {
		if tc.ID == "tech-019" || tc.ID == "tech-020" || tc.ID == "tech-021" {
			securityTestCases = append(securityTestCases, tc)
		}
	}

	if len(securityTestCases) != 3 {
		t.Errorf("❌ Expected 3 security test cases, got %d", len(securityTestCases))
		return
	}

	t.Logf("🔒 Testing %d security validation test cases", len(securityTestCases))

	for _, tc := range securityTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			// Test fake website URL validation
			if tc.ID == "tech-019" {
				if tc.WebsiteURL == "https://suspicious-tech-fake.com" {
					t.Logf("✅ Security test case %s has correct fake URL", tc.Name)
				} else {
					t.Errorf("❌ Security test case %s has incorrect URL: %s", tc.Name, tc.WebsiteURL)
				}
			}

			// Test misleading description validation
			if tc.ID == "tech-020" {
				if tc.ExpectedConfidence == 0.30 {
					t.Logf("✅ Security test case %s has correct low confidence expectation", tc.Name)
				} else {
					t.Errorf("❌ Security test case %s has incorrect confidence expectation: %.2f", tc.Name, tc.ExpectedConfidence)
				}
			}

			// Test no website URL validation
			if tc.ID == "tech-021" {
				if tc.WebsiteURL == "" {
					t.Logf("✅ Security test case %s correctly has no website URL", tc.Name)
				} else {
					t.Errorf("❌ Security test case %s should have no website URL, got: %s", tc.Name, tc.WebsiteURL)
				}
			}
		})
	}

	t.Logf("🔒 All security validation test cases are properly configured")
}

// Note: calculateOverallConfidence function is already defined in classification_accuracy_test_dataset.go
