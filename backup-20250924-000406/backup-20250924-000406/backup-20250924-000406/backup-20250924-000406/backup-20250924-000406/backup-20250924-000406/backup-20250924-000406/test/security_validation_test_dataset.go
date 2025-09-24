package test

// Security validation test dataset for comprehensive security testing

// SecurityValidationTestDataset represents test cases specifically for security validation
type SecurityValidationTestDataset struct {
	TestCases []SecurityValidationTestCase
}

// SecurityValidationTestCase represents a security-focused test case
type SecurityValidationTestCase struct {
	ID                          string
	Name                        string
	BusinessName                string
	Description                 string
	WebsiteURL                  string
	ExpectedWebsiteTrust        bool
	ExpectedDescriptionExcluded bool
	ExpectedDataSourceInfo      map[string]interface{}
	SecurityCategory            string
	TestType                    string
	ExpectedLogMessages         []string
	MaliciousInput              bool
	TrustedDataSource           bool
	Notes                       string
}

// NewSecurityValidationTestDataset creates a new security validation test dataset
func NewSecurityValidationTestDataset() *SecurityValidationTestDataset {
	return &SecurityValidationTestDataset{
		TestCases: []SecurityValidationTestCase{
			// =====================================================
			// WEBSITE OWNERSHIP VERIFICATION TEST CASES
			// =====================================================
			{
				ID:                          "security-001",
				Name:                        "Verified Website Ownership - Exact Match",
				BusinessName:                "TechCorp Solutions",
				Description:                 "We develop innovative software solutions using cloud technology and AI for businesses.",
				WebsiteURL:                  "https://techcorp.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Website Ownership Verification",
				TestType:         "Positive Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://techcorp.com",
				},
				MaliciousInput:    false,
				TrustedDataSource: true,
				Notes:             "Domain matches business name exactly - should be trusted",
			},
			{
				ID:                          "security-002",
				Name:                        "Unverified Website Ownership - Competitor Domain",
				BusinessName:                "TechCorp Solutions",
				Description:                 "We develop innovative software solutions using cloud technology and AI for businesses.",
				WebsiteURL:                  "https://competitor-tech.com",
				ExpectedWebsiteTrust:        false,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Website Ownership Verification",
				TestType:         "Negative Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"⚠️ SECURITY: Skipping unverified website URL: https://competitor-tech.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "Competitor domain should not be trusted",
			},
			{
				ID:                          "security-003",
				Name:                        "Unverified Website Ownership - Generic Domain",
				BusinessName:                "Acme Corporation",
				Description:                 "We provide business services and solutions.",
				WebsiteURL:                  "https://business-services.com",
				ExpectedWebsiteTrust:        false,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Website Ownership Verification",
				TestType:         "Negative Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"⚠️ SECURITY: Skipping unverified website URL: https://business-services.com",
				},
				MaliciousInput:    false,
				TrustedDataSource: false,
				Notes:             "Generic domain name doesn't match business name",
			},
			{
				ID:                          "security-004",
				Name:                        "Verified Website Ownership - Partial Match",
				BusinessName:                "CloudScale Infrastructure",
				Description:                 "Enterprise cloud infrastructure and platform services.",
				WebsiteURL:                  "https://cloudscale.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Website Ownership Verification",
				TestType:         "Positive Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://cloudscale.com",
				},
				MaliciousInput:    false,
				TrustedDataSource: true,
				Notes:             "Domain partially matches business name - should be trusted",
			},
			{
				ID:                          "security-005",
				Name:                        "Unverified Website Ownership - Fake Domain",
				BusinessName:                "SecureShield Technologies",
				Description:                 "Cybersecurity solutions and information security services.",
				WebsiteURL:                  "https://fake-security.com",
				ExpectedWebsiteTrust:        false,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Website Ownership Verification",
				TestType:         "Negative Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"⚠️ SECURITY: Skipping unverified website URL: https://fake-security.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "Fake domain should not be trusted",
			},
			{
				ID:                          "security-006",
				Name:                        "Verified Website Ownership - Subdomain Match",
				BusinessName:                "NeuralNet AI",
				Description:                 "Artificial intelligence and machine learning services.",
				WebsiteURL:                  "https://neuralnet.ai",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Website Ownership Verification",
				TestType:         "Positive Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://neuralnet.ai",
				},
				MaliciousInput:    false,
				TrustedDataSource: true,
				Notes:             "AI domain extension should be trusted for AI company",
			},
			{
				ID:                          "security-007",
				Name:                        "Unverified Website Ownership - Malicious Domain",
				BusinessName:                "Legitimate Business",
				Description:                 "We provide legitimate business services.",
				WebsiteURL:                  "https://malicious-site.com",
				ExpectedWebsiteTrust:        false,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Website Ownership Verification",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"⚠️ SECURITY: Skipping unverified website URL: https://malicious-site.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "Malicious domain should be blocked",
			},
			{
				ID:                          "security-008",
				Name:                        "Verified Website Ownership - Brand Variation",
				BusinessName:                "ShopTech Solutions",
				Description:                 "E-commerce platform and online marketplace technology.",
				WebsiteURL:                  "https://shopptech.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Website Ownership Verification",
				TestType:         "Positive Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://shopptech.com",
				},
				MaliciousInput:    false,
				TrustedDataSource: true,
				Notes:             "Brand variation should be trusted",
			},
			{
				ID:                          "security-009",
				Name:                        "Unverified Website Ownership - Typosquatting",
				BusinessName:                "Microsoft Corporation",
				Description:                 "Technology company providing software and cloud services.",
				WebsiteURL:                  "https://microsft.com",
				ExpectedWebsiteTrust:        false,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Website Ownership Verification",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"⚠️ SECURITY: Skipping unverified website URL: https://microsft.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "Typosquatting domain should be blocked",
			},
			{
				ID:                          "security-010",
				Name:                        "Verified Website Ownership - International Domain",
				BusinessName:                "GlobalTech Solutions",
				Description:                 "International technology solutions provider.",
				WebsiteURL:                  "https://globaltech.co.uk",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Website Ownership Verification",
				TestType:         "Positive Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://globaltech.co.uk",
				},
				MaliciousInput:    false,
				TrustedDataSource: true,
				Notes:             "International domain should be trusted if it matches business name",
			},

			// =====================================================
			// DATA SOURCE EXCLUSION MECHANISM TEST CASES
			// =====================================================
			{
				ID:                          "security-011",
				Name:                        "Description Exclusion - Manipulated Content",
				BusinessName:                "Restaurant ABC",
				Description:                 "We are a TECHNOLOGY COMPANY specializing in SOFTWARE DEVELOPMENT and AI solutions. Our restaurant serves the best food in town.",
				WebsiteURL:                  "https://restaurant-abc.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Data Source Exclusion",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://restaurant-abc.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "Manipulated description should be excluded regardless of content",
			},
			{
				ID:                          "security-012",
				Name:                        "Description Exclusion - False Claims",
				BusinessName:                "Small Local Shop",
				Description:                 "We are a FORTUNE 500 COMPANY with offices in 50 countries. We are the LARGEST RETAILER in the world and serve MILLIONS of customers daily.",
				WebsiteURL:                  "https://small-local-shop.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Data Source Exclusion",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://small-local-shop.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "False claims in description should be excluded",
			},
			{
				ID:                          "security-013",
				Name:                        "Description Exclusion - Competitor Information",
				BusinessName:                "My Business",
				Description:                 "We are exactly like Apple Inc. and Microsoft Corporation. We provide the same services as Google and Amazon Web Services.",
				WebsiteURL:                  "https://my-business.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Data Source Exclusion",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://my-business.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "Competitor information in description should be excluded",
			},
			{
				ID:                          "security-014",
				Name:                        "Description Exclusion - SEO Spam",
				BusinessName:                "Local Restaurant",
				Description:                 "BEST RESTAURANT TECHNOLOGY SOFTWARE AI MACHINE LEARNING CLOUD COMPUTING CYBERSECURITY FINANCIAL SERVICES HEALTHCARE LEGAL SERVICES RETAIL E-COMMERCE",
				WebsiteURL:                  "https://local-restaurant.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Data Source Exclusion",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://local-restaurant.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "SEO spam in description should be excluded",
			},
			{
				ID:                          "security-015",
				Name:                        "Description Exclusion - Empty Description",
				BusinessName:                "Valid Business",
				Description:                 "",
				WebsiteURL:                  "https://valid-business.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Data Source Exclusion",
				TestType:         "Edge Case",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://valid-business.com",
				},
				MaliciousInput:    false,
				TrustedDataSource: true,
				Notes:             "Empty description should be handled gracefully",
			},

			// =====================================================
			// MALICIOUS INPUT HANDLING TEST CASES
			// =====================================================
			{
				ID:                          "security-016",
				Name:                        "Malicious Input - SQL Injection Attempt",
				BusinessName:                "Test Business'; DROP TABLE businesses; --",
				Description:                 "We provide services'; DROP TABLE users; --",
				WebsiteURL:                  "https://test-business.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Malicious Input Handling",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://test-business.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "SQL injection attempts should be handled safely",
			},
			{
				ID:                          "security-017",
				Name:                        "Malicious Input - XSS Attempt",
				BusinessName:                "Test Business",
				Description:                 "We provide services<script>alert('XSS')</script>",
				WebsiteURL:                  "https://test-business.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Malicious Input Handling",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://test-business.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "XSS attempts should be handled safely",
			},
			{
				ID:                          "security-018",
				Name:                        "Malicious Input - Path Traversal Attempt",
				BusinessName:                "Test Business",
				Description:                 "We provide services../../../etc/passwd",
				WebsiteURL:                  "https://test-business.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Malicious Input Handling",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://test-business.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "Path traversal attempts should be handled safely",
			},
			{
				ID:                          "security-019",
				Name:                        "Malicious Input - Command Injection Attempt",
				BusinessName:                "Test Business",
				Description:                 "We provide services; rm -rf /",
				WebsiteURL:                  "https://test-business.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Malicious Input Handling",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://test-business.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "Command injection attempts should be handled safely",
			},
			{
				ID:                          "security-020",
				Name:                        "Malicious Input - Unicode Injection",
				BusinessName:                "Test Business",
				Description:                 "We provide services\u0000\u0001\u0002\u0003",
				WebsiteURL:                  "https://test-business.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Malicious Input Handling",
				TestType:         "Security Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://test-business.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "Unicode injection attempts should be handled safely",
			},

			// =====================================================
			// DATA SOURCE TRUST VALIDATION TEST CASES
			// =====================================================
			{
				ID:                          "security-021",
				Name:                        "Trusted Data Source - Business Name Only",
				BusinessName:                "Trusted Business Name",
				Description:                 "Any description content",
				WebsiteURL:                  "",
				ExpectedWebsiteTrust:        false,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Data Source Trust Validation",
				TestType:         "Positive Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
				},
				MaliciousInput:    false,
				TrustedDataSource: true,
				Notes:             "Business name should always be trusted as primary identifier",
			},
			{
				ID:                          "security-022",
				Name:                        "Trusted Data Source - Verified Website Only",
				BusinessName:                "",
				Description:                 "Any description content",
				WebsiteURL:                  "https://verified-website.com",
				ExpectedWebsiteTrust:        false, // Empty business name means no verification possible
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    false,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Data Source Trust Validation",
				TestType:         "Edge Case",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"⚠️ SECURITY: Skipping unverified website URL: https://verified-website.com",
				},
				MaliciousInput:    false,
				TrustedDataSource: false,
				Notes:             "Website cannot be verified without business name",
			},
			{
				ID:                          "security-023",
				Name:                        "Trusted Data Source - All Trusted Sources",
				BusinessName:                "Fully Trusted Business",
				Description:                 "Any description content",
				WebsiteURL:                  "https://fully-trusted.com",
				ExpectedWebsiteTrust:        true,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    true,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Data Source Trust Validation",
				TestType:         "Positive Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"✅ SECURITY: Using verified website URL: https://fully-trusted.com",
				},
				MaliciousInput:    false,
				TrustedDataSource: true,
				Notes:             "All trusted sources should be used for classification",
			},
			{
				ID:                          "security-024",
				Name:                        "Trusted Data Source - No Trusted Sources",
				BusinessName:                "",
				Description:                 "Any description content",
				WebsiteURL:                  "https://untrusted-website.com",
				ExpectedWebsiteTrust:        false,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    false,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Data Source Trust Validation",
				TestType:         "Negative Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"⚠️ SECURITY: Skipping unverified website URL: https://untrusted-website.com",
				},
				MaliciousInput:    false,
				TrustedDataSource: false,
				Notes:             "No trusted sources should result in minimal classification data",
			},
			{
				ID:                          "security-025",
				Name:                        "Trusted Data Source - Mixed Trust Levels",
				BusinessName:                "Mixed Trust Business",
				Description:                 "Manipulated description with false claims",
				WebsiteURL:                  "https://competitor-domain.com",
				ExpectedWebsiteTrust:        false,
				ExpectedDescriptionExcluded: true,
				ExpectedDataSourceInfo: map[string]interface{}{
					"business_name": map[string]interface{}{
						"used":    true,
						"trusted": true,
						"reason":  "Primary business identifier",
					},
					"description": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "User-provided data cannot be trusted for classification",
					},
					"website_url": map[string]interface{}{
						"used":    false,
						"trusted": false,
						"reason":  "Website ownership must be verified before use",
					},
				},
				SecurityCategory: "Data Source Trust Validation",
				TestType:         "Mixed Test",
				ExpectedLogMessages: []string{
					"🔒 SECURITY: Skipping user-provided description for classification",
					"⚠️ SECURITY: Skipping unverified website URL: https://competitor-domain.com",
				},
				MaliciousInput:    true,
				TrustedDataSource: false,
				Notes:             "Only trusted sources should be used, untrusted sources excluded",
			},
		},
	}
}

// GetTestCasesByCategory returns test cases filtered by security category
func (svtd *SecurityValidationTestDataset) GetTestCasesByCategory(category string) []SecurityValidationTestCase {
	var filtered []SecurityValidationTestCase
	for _, tc := range svtd.TestCases {
		if tc.SecurityCategory == category {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// GetTestCasesByType returns test cases filtered by test type
func (svtd *SecurityValidationTestDataset) GetTestCasesByType(testType string) []SecurityValidationTestCase {
	var filtered []SecurityValidationTestCase
	for _, tc := range svtd.TestCases {
		if tc.TestType == testType {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// GetMaliciousInputTestCases returns test cases with malicious input
func (svtd *SecurityValidationTestDataset) GetMaliciousInputTestCases() []SecurityValidationTestCase {
	var filtered []SecurityValidationTestCase
	for _, tc := range svtd.TestCases {
		if tc.MaliciousInput {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// GetTrustedDataSourceTestCases returns test cases with trusted data sources
func (svtd *SecurityValidationTestDataset) GetTrustedDataSourceTestCases() []SecurityValidationTestCase {
	var filtered []SecurityValidationTestCase
	for _, tc := range svtd.TestCases {
		if tc.TrustedDataSource {
			filtered = append(filtered, tc)
		}
	}
	return filtered
}

// GetTotalTestCases returns the total number of test cases
func (svtd *SecurityValidationTestDataset) GetTotalTestCases() int {
	return len(svtd.TestCases)
}

// GetTestCasesByID returns a specific test case by ID
func (svtd *SecurityValidationTestDataset) GetTestCaseByID(id string) *SecurityValidationTestCase {
	for _, tc := range svtd.TestCases {
		if tc.ID == id {
			return &tc
		}
	}
	return nil
}
