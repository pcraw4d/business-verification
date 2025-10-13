package onboarding

import (
	"time"
)

// GetDefaultEnterpriseOnboardingConfig returns the default enterprise onboarding configuration
func GetDefaultEnterpriseOnboardingConfig() *EnterpriseOnboardingConfig {
	return &EnterpriseOnboardingConfig{
		OnboardingSteps: []OnboardingStep{
			{
				ID:            "account_creation",
				Name:          "Account Creation",
				Description:   "Create enterprise customer account with proper permissions and access controls",
				Order:         1,
				IsRequired:    true,
				EstimatedTime: 30 * time.Minute,
				Prerequisites: []string{},
				ValidationRules: []ValidationRule{
					{
						ID:          "account_validation",
						Name:        "Account Validation",
						Description: "Validate account creation requirements",
						RuleType:    "required_field",
						Parameters: map[string]interface{}{
							"fields": []string{"name", "email", "company", "industry", "country"},
						},
						ErrorMessage: "All required fields must be provided",
					},
				},
				SuccessCriteria: []SuccessCriterion{
					{
						ID:            "account_created",
						Name:          "Account Created",
						Description:   "Account successfully created",
						CriterionType: "completion_check",
						Parameters: map[string]interface{}{
							"check_type": "account_exists",
						},
					},
				},
				Metadata: map[string]interface{}{
					"category": "account_management",
					"priority": "high",
				},
			},
			{
				ID:            "document_upload",
				Name:          "Document Upload",
				Description:   "Upload required business documents and certificates",
				Order:         2,
				IsRequired:    true,
				EstimatedTime: 45 * time.Minute,
				Prerequisites: []string{"account_creation"},
				ValidationRules: []ValidationRule{
					{
						ID:          "document_validation",
						Name:        "Document Validation",
						Description: "Validate uploaded documents",
						RuleType:    "format_validation",
						Parameters: map[string]interface{}{
							"allowed_formats": []string{"pdf", "jpg", "png"},
							"max_file_size":   10485760, // 10MB
						},
						ErrorMessage: "Documents must be in valid format and size",
					},
				},
				SuccessCriteria: []SuccessCriterion{
					{
						ID:            "documents_uploaded",
						Name:          "Documents Uploaded",
						Description:   "All required documents uploaded",
						CriterionType: "completion_check",
						Parameters: map[string]interface{}{
							"check_type": "documents_complete",
						},
					},
				},
				Metadata: map[string]interface{}{
					"category": "document_management",
					"priority": "high",
				},
			},
			{
				ID:            "compliance_check",
				Name:          "Compliance Check",
				Description:   "Perform comprehensive compliance and regulatory checks",
				Order:         3,
				IsRequired:    true,
				EstimatedTime: 60 * time.Minute,
				Prerequisites: []string{"document_upload"},
				ValidationRules: []ValidationRule{
					{
						ID:          "compliance_validation",
						Name:        "Compliance Validation",
						Description: "Validate compliance requirements",
						RuleType:    "business_logic",
						Parameters: map[string]interface{}{
							"regulations": []string{"BSA", "GDPR", "PCI-DSS"},
						},
						ErrorMessage: "Compliance requirements not met",
					},
				},
				SuccessCriteria: []SuccessCriterion{
					{
						ID:            "compliance_passed",
						Name:          "Compliance Passed",
						Description:   "All compliance checks passed",
						CriterionType: "quality_check",
						Parameters: map[string]interface{}{
							"check_type": "compliance_score",
							"threshold":  0.95,
						},
					},
				},
				Metadata: map[string]interface{}{
					"category": "compliance",
					"priority": "high",
				},
			},
			{
				ID:            "integration_setup",
				Name:          "Integration Setup",
				Description:   "Set up API integrations and webhooks",
				Order:         4,
				IsRequired:    true,
				EstimatedTime: 90 * time.Minute,
				Prerequisites: []string{"compliance_check"},
				ValidationRules: []ValidationRule{
					{
						ID:          "integration_validation",
						Name:        "Integration Validation",
						Description: "Validate integration setup",
						RuleType:    "business_logic",
						Parameters: map[string]interface{}{
							"required_endpoints": []string{"risk_assessment", "compliance_check"},
						},
						ErrorMessage: "Integration setup incomplete",
					},
				},
				SuccessCriteria: []SuccessCriterion{
					{
						ID:            "integration_configured",
						Name:          "Integration Configured",
						Description:   "Integration successfully configured",
						CriterionType: "completion_check",
						Parameters: map[string]interface{}{
							"check_type": "integration_active",
						},
					},
				},
				Metadata: map[string]interface{}{
					"category": "integration",
					"priority": "medium",
				},
			},
			{
				ID:            "testing",
				Name:          "Testing",
				Description:   "Perform comprehensive testing of all integrations and features",
				Order:         5,
				IsRequired:    true,
				EstimatedTime: 120 * time.Minute,
				Prerequisites: []string{"integration_setup"},
				ValidationRules: []ValidationRule{
					{
						ID:          "testing_validation",
						Name:        "Testing Validation",
						Description: "Validate testing requirements",
						RuleType:    "business_logic",
						Parameters: map[string]interface{}{
							"test_coverage":         0.90,
							"performance_threshold": 500, // ms
						},
						ErrorMessage: "Testing requirements not met",
					},
				},
				SuccessCriteria: []SuccessCriterion{
					{
						ID:            "testing_passed",
						Name:          "Testing Passed",
						Description:   "All tests passed successfully",
						CriterionType: "quality_check",
						Parameters: map[string]interface{}{
							"check_type": "test_results",
							"threshold":  0.95,
						},
					},
				},
				Metadata: map[string]interface{}{
					"category": "testing",
					"priority": "medium",
				},
			},
			{
				ID:            "go_live",
				Name:          "Go Live",
				Description:   "Activate production environment and finalize onboarding",
				Order:         6,
				IsRequired:    true,
				EstimatedTime: 30 * time.Minute,
				Prerequisites: []string{"testing"},
				ValidationRules: []ValidationRule{
					{
						ID:          "go_live_validation",
						Name:        "Go Live Validation",
						Description: "Validate go live requirements",
						RuleType:    "business_logic",
						Parameters: map[string]interface{}{
							"monitoring_enabled":  true,
							"alerting_configured": true,
						},
						ErrorMessage: "Go live requirements not met",
					},
				},
				SuccessCriteria: []SuccessCriterion{
					{
						ID:            "go_live_completed",
						Name:          "Go Live Completed",
						Description:   "Successfully went live",
						CriterionType: "completion_check",
						Parameters: map[string]interface{}{
							"check_type": "production_active",
						},
					},
				},
				Metadata: map[string]interface{}{
					"category": "deployment",
					"priority": "high",
				},
			},
		},
		RequiredDocuments: []RequiredDocument{
			{
				ID:           "business_license",
				Name:         "Business License",
				Description:  "Valid business license or registration certificate",
				DocumentType: "license",
				IsRequired:   true,
				FileFormats:  []string{"pdf", "jpg", "png"},
				MaxFileSize:  10485760, // 10MB
				ValidationRules: []ValidationRule{
					{
						ID:          "license_validation",
						Name:        "License Validation",
						Description: "Validate business license",
						RuleType:    "format_validation",
						Parameters: map[string]interface{}{
							"required_fields": []string{"license_number", "expiry_date"},
						},
						ErrorMessage: "Invalid business license format",
					},
				},
				Metadata: map[string]interface{}{
					"category": "legal",
					"priority": "high",
				},
			},
			{
				ID:           "tax_certificate",
				Name:         "Tax Certificate",
				Description:  "Tax registration certificate or tax ID",
				DocumentType: "tax",
				IsRequired:   true,
				FileFormats:  []string{"pdf", "jpg", "png"},
				MaxFileSize:  10485760, // 10MB
				ValidationRules: []ValidationRule{
					{
						ID:          "tax_validation",
						Name:        "Tax Validation",
						Description: "Validate tax certificate",
						RuleType:    "format_validation",
						Parameters: map[string]interface{}{
							"required_fields": []string{"tax_id", "registration_date"},
						},
						ErrorMessage: "Invalid tax certificate format",
					},
				},
				Metadata: map[string]interface{}{
					"category": "tax",
					"priority": "high",
				},
			},
			{
				ID:           "bank_statement",
				Name:         "Bank Statement",
				Description:  "Recent bank statement for verification",
				DocumentType: "financial",
				IsRequired:   true,
				FileFormats:  []string{"pdf", "jpg", "png"},
				MaxFileSize:  10485760, // 10MB
				ValidationRules: []ValidationRule{
					{
						ID:          "bank_validation",
						Name:        "Bank Validation",
						Description: "Validate bank statement",
						RuleType:    "format_validation",
						Parameters: map[string]interface{}{
							"required_fields": []string{"account_number", "bank_name", "statement_date"},
						},
						ErrorMessage: "Invalid bank statement format",
					},
				},
				Metadata: map[string]interface{}{
					"category": "financial",
					"priority": "medium",
				},
			},
		},
		ComplianceChecks: []ComplianceCheck{
			{
				ID:          "kyb_compliance",
				Name:        "KYB Compliance",
				Description: "Know Your Business compliance check",
				CheckType:   "regulatory",
				IsRequired:  true,
				ValidationRules: []ValidationRule{
					{
						ID:          "kyb_validation",
						Name:        "KYB Validation",
						Description: "Validate KYB requirements",
						RuleType:    "business_logic",
						Parameters: map[string]interface{}{
							"regulations": []string{"BSA", "FATCA"},
						},
						ErrorMessage: "KYB compliance requirements not met",
					},
				},
				SuccessCriteria: []SuccessCriterion{
					{
						ID:            "kyb_passed",
						Name:          "KYB Passed",
						Description:   "KYB compliance check passed",
						CriterionType: "quality_check",
						Parameters: map[string]interface{}{
							"check_type": "compliance_score",
							"threshold":  0.95,
						},
					},
				},
				Metadata: map[string]interface{}{
					"category": "compliance",
					"priority": "high",
				},
			},
			{
				ID:          "data_protection",
				Name:        "Data Protection",
				Description: "Data protection and privacy compliance check",
				CheckType:   "privacy",
				IsRequired:  true,
				ValidationRules: []ValidationRule{
					{
						ID:          "privacy_validation",
						Name:        "Privacy Validation",
						Description: "Validate privacy requirements",
						RuleType:    "business_logic",
						Parameters: map[string]interface{}{
							"regulations": []string{"GDPR", "CCPA", "PIPEDA"},
						},
						ErrorMessage: "Data protection requirements not met",
					},
				},
				SuccessCriteria: []SuccessCriterion{
					{
						ID:            "privacy_passed",
						Name:          "Privacy Passed",
						Description:   "Data protection compliance check passed",
						CriterionType: "quality_check",
						Parameters: map[string]interface{}{
							"check_type": "compliance_score",
							"threshold":  0.95,
						},
					},
				},
				Metadata: map[string]interface{}{
					"category": "privacy",
					"priority": "high",
				},
			},
		},
		IntegrationOptions: []IntegrationOption{
			{
				ID:              "rest_api",
				Name:            "REST API",
				Description:     "RESTful API integration for risk assessment and compliance",
				IntegrationType: "api",
				IsAvailable:     true,
				SetupSteps: []SetupStep{
					{
						ID:            "api_key_generation",
						Name:          "API Key Generation",
						Description:   "Generate API keys for authentication",
						Order:         1,
						IsRequired:    true,
						EstimatedTime: 5 * time.Minute,
						Instructions: []string{
							"Navigate to API settings",
							"Generate new API key",
							"Configure permissions",
							"Test API key",
						},
						Metadata: map[string]interface{}{
							"category": "authentication",
							"priority": "high",
						},
					},
					{
						ID:            "endpoint_configuration",
						Name:          "Endpoint Configuration",
						Description:   "Configure API endpoints and webhooks",
						Order:         2,
						IsRequired:    true,
						EstimatedTime: 15 * time.Minute,
						Instructions: []string{
							"Configure base URL",
							"Set up webhook endpoints",
							"Configure retry policies",
							"Test endpoints",
						},
						Metadata: map[string]interface{}{
							"category": "configuration",
							"priority": "high",
						},
					},
				},
				Documentation: "https://docs.kyb-platform.com/api/rest",
				Metadata: map[string]interface{}{
					"category": "api",
					"priority": "high",
				},
			},
			{
				ID:              "webhook_integration",
				Name:            "Webhook Integration",
				Description:     "Real-time webhook notifications for risk assessment results",
				IntegrationType: "webhook",
				IsAvailable:     true,
				SetupSteps: []SetupStep{
					{
						ID:            "webhook_endpoint_setup",
						Name:          "Webhook Endpoint Setup",
						Description:   "Set up webhook endpoint to receive notifications",
						Order:         1,
						IsRequired:    true,
						EstimatedTime: 10 * time.Minute,
						Instructions: []string{
							"Create webhook endpoint",
							"Configure SSL certificate",
							"Set up authentication",
							"Test webhook",
						},
						Metadata: map[string]interface{}{
							"category": "webhook",
							"priority": "high",
						},
					},
				},
				Documentation: "https://docs.kyb-platform.com/webhooks",
				Metadata: map[string]interface{}{
					"category": "webhook",
					"priority": "medium",
				},
			},
		},
		SupportTiers: []SupportTier{
			{
				ID:           "standard",
				Name:         "Standard Support",
				Description:  "Standard support with business hours coverage",
				ResponseTime: 24 * time.Hour,
				Availability: "Business Hours (9 AM - 5 PM EST)",
				Features: []string{
					"Email support",
					"Documentation access",
					"Basic troubleshooting",
					"Standard SLA",
				},
				Pricing: 0.0,
				Metadata: map[string]interface{}{
					"category": "support",
					"priority": "standard",
				},
			},
			{
				ID:           "premium",
				Name:         "Premium Support",
				Description:  "Premium support with extended hours and priority handling",
				ResponseTime: 4 * time.Hour,
				Availability: "Extended Hours (7 AM - 7 PM EST)",
				Features: []string{
					"Email and phone support",
					"Priority ticket handling",
					"Advanced troubleshooting",
					"Premium SLA",
					"Dedicated support contact",
				},
				Pricing: 500.0,
				Metadata: map[string]interface{}{
					"category": "support",
					"priority": "premium",
				},
			},
			{
				ID:           "enterprise",
				Name:         "Enterprise Support",
				Description:  "Enterprise support with 24/7 coverage and dedicated resources",
				ResponseTime: 1 * time.Hour,
				Availability: "24/7/365",
				Features: []string{
					"24/7 phone and email support",
					"Highest priority handling",
					"Advanced troubleshooting",
					"Enterprise SLA",
					"Dedicated support team",
					"On-site support available",
					"Custom integrations",
				},
				Pricing: 2000.0,
				Metadata: map[string]interface{}{
					"category": "support",
					"priority": "enterprise",
				},
			},
		},
		PricingTiers: []PricingTier{
			{
				ID:              "starter",
				Name:            "Starter",
				Description:     "Entry-level pricing for small businesses",
				BasePrice:       500.0,
				PricePerRequest: 0.10,
				MinCommitment:   1000,
				MaxCommitment:   10000,
				Features: []string{
					"Basic risk assessment",
					"Standard compliance checks",
					"Email support",
					"Basic reporting",
				},
				Metadata: map[string]interface{}{
					"category": "pricing",
					"priority": "starter",
				},
			},
			{
				ID:              "professional",
				Name:            "Professional",
				Description:     "Professional pricing for growing businesses",
				BasePrice:       1500.0,
				PricePerRequest: 0.08,
				MinCommitment:   5000,
				MaxCommitment:   50000,
				Features: []string{
					"Advanced risk assessment",
					"Comprehensive compliance checks",
					"Premium support",
					"Advanced reporting",
					"API access",
					"Webhook integration",
				},
				Metadata: map[string]interface{}{
					"category": "pricing",
					"priority": "professional",
				},
			},
			{
				ID:              "enterprise",
				Name:            "Enterprise",
				Description:     "Enterprise pricing for large organizations",
				BasePrice:       5000.0,
				PricePerRequest: 0.05,
				MinCommitment:   10000,
				MaxCommitment:   100000,
				Features: []string{
					"Enterprise risk assessment",
					"Full compliance suite",
					"24/7 enterprise support",
					"Custom reporting",
					"Full API access",
					"Webhook integration",
					"Custom integrations",
					"Dedicated account manager",
					"SLA guarantees",
				},
				Metadata: map[string]interface{}{
					"category": "pricing",
					"priority": "enterprise",
				},
			},
		},
		OnboardingTimeout: 24 * time.Hour,
		MaxRetryAttempts:  3,
		Metadata: map[string]interface{}{
			"version":    "1.0.0",
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z",
		},
	}
}
