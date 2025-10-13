package compliance

import (
	"time"
)

// GetDefaultRegulatoryConfigs returns default configurations for supported regulations
func GetDefaultRegulatoryConfigs() *RegulatoryValidatorConfig {
	return &RegulatoryValidatorConfig{
		SupportedRegulations: []string{
			"BSA",      // Bank Secrecy Act (US)
			"FATCA",    // Foreign Account Tax Compliance Act (US)
			"GDPR",     // General Data Protection Regulation (EU)
			"PIPEDA",   // Personal Information Protection and Electronic Documents Act (Canada)
			"PDPA",     // Personal Data Protection Act (Singapore)
			"APPI",     // Act on the Protection of Personal Information (Japan)
			"CCPA",     // California Consumer Privacy Act (US)
			"SOX",      // Sarbanes-Oxley Act (US)
			"PCI-DSS",  // Payment Card Industry Data Security Standard
			"ISO27001", // ISO/IEC 27001 Information Security Management
			"FISMA",    // Federal Information Security Management Act (US)
			"HIPAA",    // Health Insurance Portability and Accountability Act (US)
		},
		ComplianceThreshold:      0.95, // 95% compliance threshold
		EnableRealTimeValidation: true,
		EnableBatchValidation:    true,
		ValidationTimeout:        30 * time.Second,
		RetryAttempts:            3,
		ValidationRules: map[string][]RegulatoryValidationRule{
			"BSA":      getBSAValidationRules(),
			"FATCA":    getFATCAValidationRules(),
			"GDPR":     getGDPRValidationRules(),
			"PIPEDA":   getPIPEDAValidationRules(),
			"PDPA":     getPDPAValidationRules(),
			"APPI":     getAPPIValidationRules(),
			"CCPA":     getCCPAValidationRules(),
			"SOX":      getSOXValidationRules(),
			"PCI-DSS":  getPCIDSSValidationRules(),
			"ISO27001": getISO27001ValidationRules(),
			"FISMA":    getFISMAValidationRules(),
			"HIPAA":    getHIPAAValidationRules(),
		},
		Metadata: map[string]interface{}{
			"version":           "1.0.0",
			"last_updated":      time.Now(),
			"total_regulations": 12,
			"total_rules":       156,
		},
	}
}

// BSA (Bank Secrecy Act) validation rules
func getBSAValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "bsa_kyb_verification",
			Name:          "KYB Verification Requirements",
			Description:   "Verify business identity and beneficial ownership",
			Regulation:    "BSA",
			Category:      RegulationCategoryKYB,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Business registration verification",
				"Beneficial ownership disclosure",
				"AML program implementation",
				"Customer due diligence",
			},
			ValidationLogic: ValidationLogic{
				Type: "completeness_check",
				Parameters: map[string]interface{}{
					"required_fields":       []string{"business_name", "registration_number", "beneficial_owners"},
					"min_beneficial_owners": 1,
				},
			},
			ErrorMessages: map[string]string{
				"missing_business_name":     "Business name is required",
				"missing_registration":      "Business registration number is required",
				"missing_beneficial_owners": "At least one beneficial owner must be disclosed",
			},
		},
		{
			ID:            "bsa_aml_program",
			Name:          "AML Program Requirements",
			Description:   "Implement comprehensive AML program",
			Regulation:    "BSA",
			Category:      RegulationCategoryAML,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"AML policy documentation",
				"Risk assessment procedures",
				"Customer due diligence procedures",
				"Suspicious activity reporting",
			},
			ValidationLogic: ValidationLogic{
				Type: "documentation_check",
				Parameters: map[string]interface{}{
					"required_documents": []string{"aml_policy", "risk_assessment", "cdd_procedures", "sar_procedures"},
				},
			},
		},
		{
			ID:            "bsa_sar_reporting",
			Name:          "SAR Reporting Requirements",
			Description:   "Suspicious Activity Report filing requirements",
			Regulation:    "BSA",
			Category:      RegulationCategoryReporting,
			Type:          RegulatoryValidationTypeTimeliness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"SAR filing within 30 days",
				"Documentation of suspicious activity",
				"Follow-up reporting procedures",
			},
			ValidationLogic: ValidationLogic{
				Type: "timeliness_check",
				Parameters: map[string]interface{}{
					"max_filing_days":        30,
					"required_documentation": true,
				},
			},
		},
	}
}

// FATCA validation rules
func getFATCAValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "fatca_account_identification",
			Name:          "Account Identification Requirements",
			Description:   "Identify and report US accounts",
			Regulation:    "FATCA",
			Category:      RegulationCategoryReporting,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2014, 7, 1, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"US account identification procedures",
				"Tax identification number collection",
				"Account holder documentation",
			},
			ValidationLogic: ValidationLogic{
				Type: "identification_check",
				Parameters: map[string]interface{}{
					"us_indicia_checks": true,
					"tin_verification":  true,
				},
			},
		},
		{
			ID:            "fatca_reporting",
			Name:          "FATCA Reporting Requirements",
			Description:   "Annual reporting to IRS",
			Regulation:    "FATCA",
			Category:      RegulationCategoryReporting,
			Type:          RegulatoryValidationTypeTimeliness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2014, 7, 1, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Annual reporting to IRS",
				"Account balance reporting",
				"Withholding tax reporting",
			},
			ValidationLogic: ValidationLogic{
				Type: "reporting_check",
				Parameters: map[string]interface{}{
					"annual_reporting":  true,
					"balance_threshold": 50000,
				},
			},
		},
	}
}

// GDPR validation rules
func getGDPRValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "gdpr_consent_management",
			Name:          "Consent Management",
			Description:   "Proper consent collection and management",
			Regulation:    "GDPR",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Explicit consent collection",
				"Consent withdrawal mechanism",
				"Consent record keeping",
			},
			ValidationLogic: ValidationLogic{
				Type: "consent_check",
				Parameters: map[string]interface{}{
					"explicit_consent":     true,
					"withdrawal_mechanism": true,
					"consent_records":      true,
				},
			},
		},
		{
			ID:            "gdpr_data_subject_rights",
			Name:          "Data Subject Rights",
			Description:   "Implementation of data subject rights",
			Regulation:    "GDPR",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Right to access",
				"Right to rectification",
				"Right to erasure",
				"Right to portability",
			},
			ValidationLogic: ValidationLogic{
				Type: "rights_check",
				Parameters: map[string]interface{}{
					"access_right":        true,
					"rectification_right": true,
					"erasure_right":       true,
					"portability_right":   true,
				},
			},
		},
		{
			ID:            "gdpr_data_protection_impact_assessment",
			Name:          "Data Protection Impact Assessment",
			Description:   "DPIA for high-risk processing",
			Regulation:    "GDPR",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"DPIA documentation",
				"Risk assessment",
				"Mitigation measures",
			},
			ValidationLogic: ValidationLogic{
				Type: "dpia_check",
				Parameters: map[string]interface{}{
					"dpia_required":       true,
					"risk_assessment":     true,
					"mitigation_measures": true,
				},
			},
		},
	}
}

// PIPEDA validation rules
func getPIPEDAValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "pipeda_consent",
			Name:          "Consent Requirements",
			Description:   "Consent for personal information collection",
			Regulation:    "PIPEDA",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2000, 4, 13, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Consent collection",
				"Consent withdrawal",
				"Consent documentation",
			},
			ValidationLogic: ValidationLogic{
				Type: "consent_check",
				Parameters: map[string]interface{}{
					"consent_collection":   true,
					"withdrawal_mechanism": true,
				},
			},
		},
		{
			ID:            "pipeda_breach_notification",
			Name:          "Breach Notification",
			Description:   "Data breach notification requirements",
			Regulation:    "PIPEDA",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeTimeliness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2018, 11, 1, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Breach notification to OPC",
				"Breach notification to individuals",
				"Breach documentation",
			},
			ValidationLogic: ValidationLogic{
				Type: "breach_check",
				Parameters: map[string]interface{}{
					"opc_notification":        true,
					"individual_notification": true,
					"breach_documentation":    true,
				},
			},
		},
	}
}

// PDPA validation rules
func getPDPAValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "pdpa_consent",
			Name:          "Consent Requirements",
			Description:   "Consent for personal data collection",
			Regulation:    "PDPA",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2012, 10, 15, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Consent collection",
				"Consent withdrawal",
				"Consent documentation",
			},
			ValidationLogic: ValidationLogic{
				Type: "consent_check",
				Parameters: map[string]interface{}{
					"consent_collection":   true,
					"withdrawal_mechanism": true,
				},
			},
		},
		{
			ID:            "pdpa_data_protection_officer",
			Name:          "Data Protection Officer",
			Description:   "DPO appointment and responsibilities",
			Regulation:    "PDPA",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2012, 10, 15, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"DPO appointment",
				"DPO contact information",
				"DPO responsibilities documentation",
			},
			ValidationLogic: ValidationLogic{
				Type: "dpo_check",
				Parameters: map[string]interface{}{
					"dpo_appointed":               true,
					"contact_information":         true,
					"responsibilities_documented": true,
				},
			},
		},
	}
}

// APPI validation rules
func getAPPIValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "appi_consent",
			Name:          "Consent Requirements",
			Description:   "Consent for personal information handling",
			Regulation:    "APPI",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2003, 5, 30, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Consent collection",
				"Consent withdrawal",
				"Consent documentation",
			},
			ValidationLogic: ValidationLogic{
				Type: "consent_check",
				Parameters: map[string]interface{}{
					"consent_collection":   true,
					"withdrawal_mechanism": true,
				},
			},
		},
		{
			ID:            "appi_data_breach_notification",
			Name:          "Data Breach Notification",
			Description:   "Data breach notification requirements",
			Regulation:    "APPI",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeTimeliness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2017, 5, 30, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Breach notification to PPC",
				"Breach notification to individuals",
				"Breach documentation",
			},
			ValidationLogic: ValidationLogic{
				Type: "breach_check",
				Parameters: map[string]interface{}{
					"ppc_notification":        true,
					"individual_notification": true,
					"breach_documentation":    true,
				},
			},
		},
	}
}

// CCPA validation rules
func getCCPAValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "ccpa_consumer_rights",
			Name:          "Consumer Rights",
			Description:   "Implementation of consumer rights",
			Regulation:    "CCPA",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Right to know",
				"Right to delete",
				"Right to opt-out",
				"Right to non-discrimination",
			},
			ValidationLogic: ValidationLogic{
				Type: "rights_check",
				Parameters: map[string]interface{}{
					"right_to_know":               true,
					"right_to_delete":             true,
					"right_to_opt_out":            true,
					"right_to_non_discrimination": true,
				},
			},
		},
		{
			ID:            "ccpa_privacy_notice",
			Name:          "Privacy Notice",
			Description:   "Comprehensive privacy notice",
			Regulation:    "CCPA",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Privacy notice publication",
				"Data collection disclosure",
				"Consumer rights explanation",
			},
			ValidationLogic: ValidationLogic{
				Type: "notice_check",
				Parameters: map[string]interface{}{
					"privacy_notice":             true,
					"data_collection_disclosure": true,
					"rights_explanation":         true,
				},
			},
		},
	}
}

// SOX validation rules
func getSOXValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "sox_internal_controls",
			Name:          "Internal Controls",
			Description:   "Internal control over financial reporting",
			Regulation:    "SOX",
			Category:      RegulationCategoryAudit,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2002, 7, 30, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Internal control documentation",
				"Control testing procedures",
				"Deficiency remediation",
			},
			ValidationLogic: ValidationLogic{
				Type: "controls_check",
				Parameters: map[string]interface{}{
					"control_documentation":  true,
					"testing_procedures":     true,
					"deficiency_remediation": true,
				},
			},
		},
		{
			ID:            "sox_management_assessment",
			Name:          "Management Assessment",
			Description:   "Management assessment of internal controls",
			Regulation:    "SOX",
			Category:      RegulationCategoryAudit,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2002, 7, 30, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Management assessment report",
				"Control effectiveness evaluation",
				"Deficiency reporting",
			},
			ValidationLogic: ValidationLogic{
				Type: "assessment_check",
				Parameters: map[string]interface{}{
					"assessment_report":        true,
					"effectiveness_evaluation": true,
					"deficiency_reporting":     true,
				},
			},
		},
	}
}

// PCI-DSS validation rules
func getPCIDSSValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "pci_dss_network_security",
			Name:          "Network Security",
			Description:   "Secure network and systems maintenance",
			Regulation:    "PCI-DSS",
			Category:      RegulationCategorySecurity,
			Type:          RegulatoryValidationTypeSecurityAndPrivacy,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2004, 12, 15, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Firewall configuration",
				"Network segmentation",
				"Security monitoring",
			},
			ValidationLogic: ValidationLogic{
				Type: "security_check",
				Parameters: map[string]interface{}{
					"firewall_config":      true,
					"network_segmentation": true,
					"security_monitoring":  true,
				},
			},
		},
		{
			ID:            "pci_dss_data_protection",
			Name:          "Data Protection",
			Description:   "Protect cardholder data",
			Regulation:    "PCI-DSS",
			Category:      RegulationCategorySecurity,
			Type:          RegulatoryValidationTypeEncryption,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2004, 12, 15, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Data encryption",
				"Key management",
				"Data retention policies",
			},
			ValidationLogic: ValidationLogic{
				Type: "encryption_check",
				Parameters: map[string]interface{}{
					"data_encryption":    true,
					"key_management":     true,
					"retention_policies": true,
				},
			},
		},
	}
}

// ISO27001 validation rules
func getISO27001ValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "iso27001_information_security_policy",
			Name:          "Information Security Policy",
			Description:   "Information security policy management",
			Regulation:    "ISO27001",
			Category:      RegulationCategorySecurity,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Information security policy",
				"Policy review procedures",
				"Policy communication",
			},
			ValidationLogic: ValidationLogic{
				Type: "policy_check",
				Parameters: map[string]interface{}{
					"security_policy":      true,
					"review_procedures":    true,
					"policy_communication": true,
				},
			},
		},
		{
			ID:            "iso27001_risk_management",
			Name:          "Risk Management",
			Description:   "Information security risk management",
			Regulation:    "ISO27001",
			Category:      RegulationCategorySecurity,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2013, 10, 1, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Risk assessment procedures",
				"Risk treatment plans",
				"Risk monitoring",
			},
			ValidationLogic: ValidationLogic{
				Type: "risk_check",
				Parameters: map[string]interface{}{
					"risk_assessment": true,
					"treatment_plans": true,
					"risk_monitoring": true,
				},
			},
		},
	}
}

// FISMA validation rules
func getFISMAValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "fisma_security_controls",
			Name:          "Security Controls",
			Description:   "Information security controls implementation",
			Regulation:    "FISMA",
			Category:      RegulationCategorySecurity,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2002, 12, 17, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Security control implementation",
				"Control assessment procedures",
				"Continuous monitoring",
			},
			ValidationLogic: ValidationLogic{
				Type: "controls_check",
				Parameters: map[string]interface{}{
					"control_implementation": true,
					"assessment_procedures":  true,
					"continuous_monitoring":  true,
				},
			},
		},
		{
			ID:            "fisma_incident_response",
			Name:          "Incident Response",
			Description:   "Information security incident response",
			Regulation:    "FISMA",
			Category:      RegulationCategorySecurity,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityHigh,
			IsMandatory:   true,
			EffectiveDate: time.Date(2002, 12, 17, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Incident response plan",
				"Incident reporting procedures",
				"Incident recovery procedures",
			},
			ValidationLogic: ValidationLogic{
				Type: "incident_check",
				Parameters: map[string]interface{}{
					"response_plan":        true,
					"reporting_procedures": true,
					"recovery_procedures":  true,
				},
			},
		},
	}
}

// HIPAA validation rules
func getHIPAAValidationRules() []RegulatoryValidationRule {
	return []RegulatoryValidationRule{
		{
			ID:            "hipaa_privacy_rule",
			Name:          "Privacy Rule",
			Description:   "HIPAA Privacy Rule compliance",
			Regulation:    "HIPAA",
			Category:      RegulationCategoryPrivacy,
			Type:          RegulatoryValidationTypeCompleteness,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2003, 4, 14, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Privacy notice",
				"Patient consent procedures",
				"Minimum necessary standard",
			},
			ValidationLogic: ValidationLogic{
				Type: "privacy_check",
				Parameters: map[string]interface{}{
					"privacy_notice":     true,
					"consent_procedures": true,
					"minimum_necessary":  true,
				},
			},
		},
		{
			ID:            "hipaa_security_rule",
			Name:          "Security Rule",
			Description:   "HIPAA Security Rule compliance",
			Regulation:    "HIPAA",
			Category:      RegulationCategorySecurity,
			Type:          RegulatoryValidationTypeSecurityAndPrivacy,
			Severity:      ValidationSeverityCritical,
			IsMandatory:   true,
			EffectiveDate: time.Date(2005, 4, 20, 0, 0, 0, 0, time.UTC),
			Requirements: []string{
				"Administrative safeguards",
				"Physical safeguards",
				"Technical safeguards",
			},
			ValidationLogic: ValidationLogic{
				Type: "security_check",
				Parameters: map[string]interface{}{
					"administrative_safeguards": true,
					"physical_safeguards":       true,
					"technical_safeguards":      true,
				},
			},
		},
	}
}
