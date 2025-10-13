package compliance

import (
	"time"
)

// GetDefaultCountryConfigs returns default configurations for supported countries
func GetDefaultCountryConfigs() map[string]*CountryConfig {
	configs := make(map[string]*CountryConfig)

	// United States
	configs["US"] = &CountryConfig{
		Code:     "US",
		Name:     "United States",
		Region:   "North America",
		Currency: "USD",
		Language: "en",
		Timezone: "America/New_York",
		RiskFactors: []RiskFactor{
			{
				ID:          "us_political_risk",
				Name:        "Political Risk",
				Description: "Political instability and regulatory changes",
				Category:    "political",
				Severity:    "medium",
				Weight:      0.3,
				IsActive:    true,
				LocalizedName: map[string]string{
					"en": "Political Risk",
					"es": "Riesgo Político",
				},
				LocalizedDesc: map[string]string{
					"en": "Political instability and regulatory changes",
					"es": "Inestabilidad política y cambios regulatorios",
				},
			},
			{
				ID:          "us_regulatory_risk",
				Name:        "Regulatory Risk",
				Description: "Complex regulatory environment and compliance requirements",
				Category:    "regulatory",
				Severity:    "high",
				Weight:      0.4,
				IsActive:    true,
				LocalizedName: map[string]string{
					"en": "Regulatory Risk",
					"es": "Riesgo Regulatorio",
				},
				LocalizedDesc: map[string]string{
					"en": "Complex regulatory environment and compliance requirements",
					"es": "Entorno regulatorio complejo y requisitos de cumplimiento",
				},
			},
		},
		ComplianceRules: []ComplianceRule{
			{
				ID:             "us_kyb_requirement",
				Name:           "KYB Requirements",
				Description:    "Know Your Business requirements under BSA",
				Type:           "verification",
				Category:       "kyb",
				IsMandatory:    true,
				EffectiveDate:  time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
				RegulatoryBody: "FinCEN",
				Requirements: []string{
					"Business registration verification",
					"Beneficial ownership disclosure",
					"AML program implementation",
				},
				LocalizedName: map[string]string{
					"en": "KYB Requirements",
					"es": "Requisitos KYB",
				},
				LocalizedDesc: map[string]string{
					"en": "Know Your Business requirements under BSA",
					"es": "Requisitos de Conocer su Negocio bajo BSA",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				ID:           "us_ein_validation",
				Name:         "EIN Validation",
				Field:        "tax_id",
				Type:         "regex",
				Pattern:      "^\\d{2}-\\d{7}$",
				Required:     true,
				ErrorMessage: "Invalid EIN format",
				LocalizedError: map[string]string{
					"en": "Invalid EIN format",
					"es": "Formato de EIN inválido",
				},
			},
		},
		SanctionsLists: []string{"OFAC", "BIS", "DDTC"},
		RegulatoryBodies: []RegulatoryBody{
			{
				ID:           "fincen",
				Name:         "Financial Crimes Enforcement Network",
				Acronym:      "FinCEN",
				Type:         "financial",
				Jurisdiction: "US",
				Website:      "https://www.fincen.gov",
				Responsibilities: []string{
					"AML enforcement",
					"KYB requirements",
					"SAR reporting",
				},
				LocalizedName: map[string]string{
					"en": "Financial Crimes Enforcement Network",
					"es": "Red de Ejecución de Delitos Financieros",
				},
			},
		},
		DataResidencyRules: DataResidencyRules{
			RequiresLocalStorage: false,
			AllowedRegions:       []string{"US", "EU"},
			CrossBorderTransfer:  true,
			RetentionPeriod:      2555, // 7 years
		},
		BusinessTypes: []BusinessType{
			{
				ID:        "us_corporation",
				Name:      "Corporation",
				Code:      "CORP",
				Category:  "business",
				RiskLevel: "medium",
				Requirements: []string{
					"Articles of Incorporation",
					"EIN",
					"State registration",
				},
				LocalizedName: map[string]string{
					"en": "Corporation",
					"es": "Corporación",
				},
			},
		},
		DocumentTypes: []DocumentType{
			{
				ID:             "us_articles_incorporation",
				Name:           "Articles of Incorporation",
				Code:           "AOI",
				Category:       "registration",
				Required:       true,
				ValidityPeriod: 0, // No expiry
				Format:         "PDF",
				LocalizedName: map[string]string{
					"en": "Articles of Incorporation",
					"es": "Artículos de Incorporación",
				},
			},
		},
	}

	// United Kingdom
	configs["GB"] = &CountryConfig{
		Code:     "GB",
		Name:     "United Kingdom",
		Region:   "Europe",
		Currency: "GBP",
		Language: "en",
		Timezone: "Europe/London",
		RiskFactors: []RiskFactor{
			{
				ID:          "gb_brexit_risk",
				Name:        "Brexit Risk",
				Description: "Post-Brexit regulatory and economic uncertainty",
				Category:    "political",
				Severity:    "high",
				Weight:      0.4,
				IsActive:    true,
				LocalizedName: map[string]string{
					"en": "Brexit Risk",
					"fr": "Risque Brexit",
				},
				LocalizedDesc: map[string]string{
					"en": "Post-Brexit regulatory and economic uncertainty",
					"fr": "Incertitude réglementaire et économique post-Brexit",
				},
			},
		},
		ComplianceRules: []ComplianceRule{
			{
				ID:             "gb_aml_requirement",
				Name:           "AML Requirements",
				Description:    "Anti-Money Laundering requirements under MLR 2017",
				Type:           "monitoring",
				Category:       "aml",
				IsMandatory:    true,
				EffectiveDate:  time.Date(2017, 6, 26, 0, 0, 0, 0, time.UTC),
				RegulatoryBody: "FCA",
				Requirements: []string{
					"Customer due diligence",
					"Enhanced due diligence",
					"Suspicious activity reporting",
				},
				LocalizedName: map[string]string{
					"en": "AML Requirements",
					"fr": "Exigences AML",
				},
				LocalizedDesc: map[string]string{
					"en": "Anti-Money Laundering requirements under MLR 2017",
					"fr": "Exigences de lutte contre le blanchiment d'argent sous MLR 2017",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				ID:           "gb_company_number",
				Name:         "Company Number Validation",
				Field:        "registration_number",
				Type:         "regex",
				Pattern:      "^\\d{8}$",
				Required:     true,
				ErrorMessage: "Invalid company number format",
				LocalizedError: map[string]string{
					"en": "Invalid company number format",
					"fr": "Format de numéro de société invalide",
				},
			},
		},
		SanctionsLists: []string{"OFSI", "UN", "EU"},
		RegulatoryBodies: []RegulatoryBody{
			{
				ID:           "fca",
				Name:         "Financial Conduct Authority",
				Acronym:      "FCA",
				Type:         "financial",
				Jurisdiction: "UK",
				Website:      "https://www.fca.org.uk",
				Responsibilities: []string{
					"Financial services regulation",
					"AML supervision",
					"Market conduct",
				},
				LocalizedName: map[string]string{
					"en": "Financial Conduct Authority",
					"fr": "Autorité de conduite financière",
				},
			},
		},
		DataResidencyRules: DataResidencyRules{
			RequiresLocalStorage: false,
			AllowedRegions:       []string{"UK", "EU"},
			CrossBorderTransfer:  true,
			RetentionPeriod:      1825, // 5 years
		},
	}

	// Germany
	configs["DE"] = &CountryConfig{
		Code:     "DE",
		Name:     "Germany",
		Region:   "Europe",
		Currency: "EUR",
		Language: "de",
		Timezone: "Europe/Berlin",
		RiskFactors: []RiskFactor{
			{
				ID:          "de_gdpr_risk",
				Name:        "GDPR Compliance Risk",
				Description: "Data protection and privacy compliance requirements",
				Category:    "compliance",
				Severity:    "high",
				Weight:      0.5,
				IsActive:    true,
				LocalizedName: map[string]string{
					"de": "DSGVO-Compliance-Risiko",
					"en": "GDPR Compliance Risk",
				},
				LocalizedDesc: map[string]string{
					"de": "Datenschutz- und Datenschutz-Compliance-Anforderungen",
					"en": "Data protection and privacy compliance requirements",
				},
			},
		},
		ComplianceRules: []ComplianceRule{
			{
				ID:             "de_gdpr_requirement",
				Name:           "GDPR Requirements",
				Description:    "General Data Protection Regulation compliance",
				Type:           "privacy",
				Category:       "privacy",
				IsMandatory:    true,
				EffectiveDate:  time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
				RegulatoryBody: "BfDI",
				Requirements: []string{
					"Data protection impact assessment",
					"Privacy by design",
					"Data subject rights",
				},
				LocalizedName: map[string]string{
					"de": "DSGVO-Anforderungen",
					"en": "GDPR Requirements",
				},
				LocalizedDesc: map[string]string{
					"de": "Compliance mit der Datenschutz-Grundverordnung",
					"en": "General Data Protection Regulation compliance",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				ID:           "de_handelsregister",
				Name:         "Handelsregister Validation",
				Field:        "registration_number",
				Type:         "regex",
				Pattern:      "^HRB \\d+$",
				Required:     true,
				ErrorMessage: "Invalid Handelsregister number format",
				LocalizedError: map[string]string{
					"de": "Ungültiges Handelsregister-Nummernformat",
					"en": "Invalid Handelsregister number format",
				},
			},
		},
		SanctionsLists: []string{"EU", "UN", "OFAC"},
		RegulatoryBodies: []RegulatoryBody{
			{
				ID:           "bfdi",
				Name:         "Bundesbeauftragte für den Datenschutz und die Informationsfreiheit",
				Acronym:      "BfDI",
				Type:         "data_protection",
				Jurisdiction: "Germany",
				Website:      "https://www.bfdi.bund.de",
				Responsibilities: []string{
					"Data protection supervision",
					"GDPR enforcement",
					"Privacy rights protection",
				},
				LocalizedName: map[string]string{
					"de": "Bundesbeauftragte für den Datenschutz und die Informationsfreiheit",
					"en": "Federal Commissioner for Data Protection and Freedom of Information",
				},
			},
		},
		DataResidencyRules: DataResidencyRules{
			RequiresLocalStorage: true,
			AllowedRegions:       []string{"DE", "EU"},
			CrossBorderTransfer:  false,
			RetentionPeriod:      1095, // 3 years
		},
	}

	// Canada
	configs["CA"] = &CountryConfig{
		Code:     "CA",
		Name:     "Canada",
		Region:   "North America",
		Currency: "CAD",
		Language: "en",
		Timezone: "America/Toronto",
		RiskFactors: []RiskFactor{
			{
				ID:          "ca_pipeda_risk",
				Name:        "PIPEDA Compliance Risk",
				Description: "Personal Information Protection and Electronic Documents Act compliance",
				Category:    "privacy",
				Severity:    "medium",
				Weight:      0.3,
				IsActive:    true,
				LocalizedName: map[string]string{
					"en": "PIPEDA Compliance Risk",
					"fr": "Risque de conformité PIPEDA",
				},
				LocalizedDesc: map[string]string{
					"en": "Personal Information Protection and Electronic Documents Act compliance",
					"fr": "Conformité à la Loi sur la protection des renseignements personnels et les documents électroniques",
				},
			},
		},
		ComplianceRules: []ComplianceRule{
			{
				ID:             "ca_pipeda_requirement",
				Name:           "PIPEDA Requirements",
				Description:    "Personal Information Protection and Electronic Documents Act",
				Type:           "privacy",
				Category:       "privacy",
				IsMandatory:    true,
				EffectiveDate:  time.Date(2000, 4, 13, 0, 0, 0, 0, time.UTC),
				RegulatoryBody: "OPC",
				Requirements: []string{
					"Privacy impact assessment",
					"Consent management",
					"Data breach notification",
				},
				LocalizedName: map[string]string{
					"en": "PIPEDA Requirements",
					"fr": "Exigences PIPEDA",
				},
				LocalizedDesc: map[string]string{
					"en": "Personal Information Protection and Electronic Documents Act",
					"fr": "Loi sur la protection des renseignements personnels et les documents électroniques",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				ID:           "ca_business_number",
				Name:         "Business Number Validation",
				Field:        "business_number",
				Type:         "regex",
				Pattern:      "^\\d{9}$",
				Required:     true,
				ErrorMessage: "Invalid business number format",
				LocalizedError: map[string]string{
					"en": "Invalid business number format",
					"fr": "Format de numéro d'entreprise invalide",
				},
			},
		},
		SanctionsLists: []string{"OSFI", "UN", "OFAC"},
		RegulatoryBodies: []RegulatoryBody{
			{
				ID:           "opc",
				Name:         "Office of the Privacy Commissioner of Canada",
				Acronym:      "OPC",
				Type:         "data_protection",
				Jurisdiction: "Canada",
				Website:      "https://www.priv.gc.ca",
				Responsibilities: []string{
					"Privacy protection",
					"PIPEDA enforcement",
					"Privacy rights advocacy",
				},
				LocalizedName: map[string]string{
					"en": "Office of the Privacy Commissioner of Canada",
					"fr": "Commissariat à la protection de la vie privée du Canada",
				},
			},
		},
		DataResidencyRules: DataResidencyRules{
			RequiresLocalStorage: false,
			AllowedRegions:       []string{"CA", "US"},
			CrossBorderTransfer:  true,
			RetentionPeriod:      1825, // 5 years
		},
	}

	// Australia
	configs["AU"] = &CountryConfig{
		Code:     "AU",
		Name:     "Australia",
		Region:   "Oceania",
		Currency: "AUD",
		Language: "en",
		Timezone: "Australia/Sydney",
		RiskFactors: []RiskFactor{
			{
				ID:          "au_privacy_risk",
				Name:        "Privacy Act Risk",
				Description: "Privacy Act 1988 compliance requirements",
				Category:    "privacy",
				Severity:    "medium",
				Weight:      0.3,
				IsActive:    true,
				LocalizedName: map[string]string{
					"en": "Privacy Act Risk",
				},
				LocalizedDesc: map[string]string{
					"en": "Privacy Act 1988 compliance requirements",
				},
			},
		},
		ComplianceRules: []ComplianceRule{
			{
				ID:             "au_privacy_requirement",
				Name:           "Privacy Act Requirements",
				Description:    "Privacy Act 1988 compliance",
				Type:           "privacy",
				Category:       "privacy",
				IsMandatory:    true,
				EffectiveDate:  time.Date(1988, 12, 14, 0, 0, 0, 0, time.UTC),
				RegulatoryBody: "OAIC",
				Requirements: []string{
					"Privacy policy",
					"Data breach notification",
					"Consent management",
				},
				LocalizedName: map[string]string{
					"en": "Privacy Act Requirements",
				},
				LocalizedDesc: map[string]string{
					"en": "Privacy Act 1988 compliance",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				ID:           "au_acn_validation",
				Name:         "ACN Validation",
				Field:        "acn",
				Type:         "regex",
				Pattern:      "^\\d{9}$",
				Required:     true,
				ErrorMessage: "Invalid ACN format",
				LocalizedError: map[string]string{
					"en": "Invalid ACN format",
				},
			},
		},
		SanctionsLists: []string{"DFAT", "UN", "OFAC"},
		RegulatoryBodies: []RegulatoryBody{
			{
				ID:           "oaic",
				Name:         "Office of the Australian Information Commissioner",
				Acronym:      "OAIC",
				Type:         "data_protection",
				Jurisdiction: "Australia",
				Website:      "https://www.oaic.gov.au",
				Responsibilities: []string{
					"Privacy protection",
					"Privacy Act enforcement",
					"Information access rights",
				},
				LocalizedName: map[string]string{
					"en": "Office of the Australian Information Commissioner",
				},
			},
		},
		DataResidencyRules: DataResidencyRules{
			RequiresLocalStorage: false,
			AllowedRegions:       []string{"AU", "NZ"},
			CrossBorderTransfer:  true,
			RetentionPeriod:      1825, // 5 years
		},
	}

	// Singapore
	configs["SG"] = &CountryConfig{
		Code:     "SG",
		Name:     "Singapore",
		Region:   "Asia",
		Currency: "SGD",
		Language: "en",
		Timezone: "Asia/Singapore",
		RiskFactors: []RiskFactor{
			{
				ID:          "sg_pdpa_risk",
				Name:        "PDPA Compliance Risk",
				Description: "Personal Data Protection Act compliance requirements",
				Category:    "privacy",
				Severity:    "high",
				Weight:      0.4,
				IsActive:    true,
				LocalizedName: map[string]string{
					"en": "PDPA Compliance Risk",
					"zh": "PDPA合规风险",
				},
				LocalizedDesc: map[string]string{
					"en": "Personal Data Protection Act compliance requirements",
					"zh": "个人数据保护法合规要求",
				},
			},
		},
		ComplianceRules: []ComplianceRule{
			{
				ID:             "sg_pdpa_requirement",
				Name:           "PDPA Requirements",
				Description:    "Personal Data Protection Act compliance",
				Type:           "privacy",
				Category:       "privacy",
				IsMandatory:    true,
				EffectiveDate:  time.Date(2012, 10, 15, 0, 0, 0, 0, time.UTC),
				RegulatoryBody: "PDPC",
				Requirements: []string{
					"Data protection policy",
					"Consent management",
					"Data breach notification",
				},
				LocalizedName: map[string]string{
					"en": "PDPA Requirements",
					"zh": "PDPA要求",
				},
				LocalizedDesc: map[string]string{
					"en": "Personal Data Protection Act compliance",
					"zh": "个人数据保护法合规",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				ID:           "sg_uen_validation",
				Name:         "UEN Validation",
				Field:        "uen",
				Type:         "regex",
				Pattern:      "^\\d{8}[A-Z]$",
				Required:     true,
				ErrorMessage: "Invalid UEN format",
				LocalizedError: map[string]string{
					"en": "Invalid UEN format",
					"zh": "无效的UEN格式",
				},
			},
		},
		SanctionsLists: []string{"MAS", "UN", "OFAC"},
		RegulatoryBodies: []RegulatoryBody{
			{
				ID:           "pdpc",
				Name:         "Personal Data Protection Commission",
				Acronym:      "PDPC",
				Type:         "data_protection",
				Jurisdiction: "Singapore",
				Website:      "https://www.pdpc.gov.sg",
				Responsibilities: []string{
					"Data protection regulation",
					"PDPA enforcement",
					"Privacy rights protection",
				},
				LocalizedName: map[string]string{
					"en": "Personal Data Protection Commission",
					"zh": "个人数据保护委员会",
				},
			},
		},
		DataResidencyRules: DataResidencyRules{
			RequiresLocalStorage: true,
			AllowedRegions:       []string{"SG"},
			CrossBorderTransfer:  false,
			RetentionPeriod:      1095, // 3 years
		},
	}

	// Japan
	configs["JP"] = &CountryConfig{
		Code:     "JP",
		Name:     "Japan",
		Region:   "Asia",
		Currency: "JPY",
		Language: "ja",
		Timezone: "Asia/Tokyo",
		RiskFactors: []RiskFactor{
			{
				ID:          "jp_appi_risk",
				Name:        "APPI Compliance Risk",
				Description: "Act on the Protection of Personal Information compliance",
				Category:    "privacy",
				Severity:    "high",
				Weight:      0.4,
				IsActive:    true,
				LocalizedName: map[string]string{
					"ja": "個人情報保護法コンプライアンスリスク",
					"en": "APPI Compliance Risk",
				},
				LocalizedDesc: map[string]string{
					"ja": "個人情報保護法のコンプライアンス要件",
					"en": "Act on the Protection of Personal Information compliance",
				},
			},
		},
		ComplianceRules: []ComplianceRule{
			{
				ID:             "jp_appi_requirement",
				Name:           "APPI Requirements",
				Description:    "Act on the Protection of Personal Information",
				Type:           "privacy",
				Category:       "privacy",
				IsMandatory:    true,
				EffectiveDate:  time.Date(2003, 5, 30, 0, 0, 0, 0, time.UTC),
				RegulatoryBody: "PPC",
				Requirements: []string{
					"Privacy policy",
					"Data handling procedures",
					"Consent management",
				},
				LocalizedName: map[string]string{
					"ja": "個人情報保護法要件",
					"en": "APPI Requirements",
				},
				LocalizedDesc: map[string]string{
					"ja": "個人情報保護法",
					"en": "Act on the Protection of Personal Information",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				ID:           "jp_corporate_number",
				Name:         "Corporate Number Validation",
				Field:        "corporate_number",
				Type:         "regex",
				Pattern:      "^\\d{13}$",
				Required:     true,
				ErrorMessage: "Invalid corporate number format",
				LocalizedError: map[string]string{
					"ja": "無効な法人番号形式",
					"en": "Invalid corporate number format",
				},
			},
		},
		SanctionsLists: []string{"MOFA", "UN", "OFAC"},
		RegulatoryBodies: []RegulatoryBody{
			{
				ID:           "ppc",
				Name:         "Personal Information Protection Commission",
				Acronym:      "PPC",
				Type:         "data_protection",
				Jurisdiction: "Japan",
				Website:      "https://www.ppc.go.jp",
				Responsibilities: []string{
					"Personal information protection",
					"APPI enforcement",
					"Privacy rights protection",
				},
				LocalizedName: map[string]string{
					"ja": "個人情報保護委員会",
					"en": "Personal Information Protection Commission",
				},
			},
		},
		DataResidencyRules: DataResidencyRules{
			RequiresLocalStorage: true,
			AllowedRegions:       []string{"JP"},
			CrossBorderTransfer:  false,
			RetentionPeriod:      1095, // 3 years
		},
	}

	// France
	configs["FR"] = &CountryConfig{
		Code:     "FR",
		Name:     "France",
		Region:   "Europe",
		Currency: "EUR",
		Language: "fr",
		Timezone: "Europe/Paris",
		RiskFactors: []RiskFactor{
			{
				ID:          "fr_gdpr_risk",
				Name:        "GDPR Compliance Risk",
				Description: "GDPR and French data protection law compliance",
				Category:    "privacy",
				Severity:    "high",
				Weight:      0.5,
				IsActive:    true,
				LocalizedName: map[string]string{
					"fr": "Risque de conformité RGPD",
					"en": "GDPR Compliance Risk",
				},
				LocalizedDesc: map[string]string{
					"fr": "Conformité RGPD et loi française sur la protection des données",
					"en": "GDPR and French data protection law compliance",
				},
			},
		},
		ComplianceRules: []ComplianceRule{
			{
				ID:             "fr_gdpr_requirement",
				Name:           "GDPR Requirements",
				Description:    "GDPR and French data protection compliance",
				Type:           "privacy",
				Category:       "privacy",
				IsMandatory:    true,
				EffectiveDate:  time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
				RegulatoryBody: "CNIL",
				Requirements: []string{
					"Data protection impact assessment",
					"Privacy by design",
					"Data subject rights",
				},
				LocalizedName: map[string]string{
					"fr": "Exigences RGPD",
					"en": "GDPR Requirements",
				},
				LocalizedDesc: map[string]string{
					"fr": "Conformité RGPD et protection des données française",
					"en": "GDPR and French data protection compliance",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				ID:           "fr_siret_validation",
				Name:         "SIRET Validation",
				Field:        "siret",
				Type:         "regex",
				Pattern:      "^\\d{14}$",
				Required:     true,
				ErrorMessage: "Invalid SIRET format",
				LocalizedError: map[string]string{
					"fr": "Format SIRET invalide",
					"en": "Invalid SIRET format",
				},
			},
		},
		SanctionsLists: []string{"EU", "UN", "OFAC"},
		RegulatoryBodies: []RegulatoryBody{
			{
				ID:           "cnil",
				Name:         "Commission Nationale de l'Informatique et des Libertés",
				Acronym:      "CNIL",
				Type:         "data_protection",
				Jurisdiction: "France",
				Website:      "https://www.cnil.fr",
				Responsibilities: []string{
					"Data protection supervision",
					"GDPR enforcement",
					"Privacy rights protection",
				},
				LocalizedName: map[string]string{
					"fr": "Commission Nationale de l'Informatique et des Libertés",
					"en": "National Commission on Informatics and Liberty",
				},
			},
		},
		DataResidencyRules: DataResidencyRules{
			RequiresLocalStorage: true,
			AllowedRegions:       []string{"FR", "EU"},
			CrossBorderTransfer:  false,
			RetentionPeriod:      1095, // 3 years
		},
	}

	// Netherlands
	configs["NL"] = &CountryConfig{
		Code:     "NL",
		Name:     "Netherlands",
		Region:   "Europe",
		Currency: "EUR",
		Language: "nl",
		Timezone: "Europe/Amsterdam",
		RiskFactors: []RiskFactor{
			{
				ID:          "nl_gdpr_risk",
				Name:        "GDPR Compliance Risk",
				Description: "GDPR and Dutch data protection law compliance",
				Category:    "privacy",
				Severity:    "high",
				Weight:      0.5,
				IsActive:    true,
				LocalizedName: map[string]string{
					"nl": "AVG-nalevingsrisico",
					"en": "GDPR Compliance Risk",
				},
				LocalizedDesc: map[string]string{
					"nl": "AVG en Nederlandse wet op gegevensbescherming naleving",
					"en": "GDPR and Dutch data protection law compliance",
				},
			},
		},
		ComplianceRules: []ComplianceRule{
			{
				ID:             "nl_gdpr_requirement",
				Name:           "GDPR Requirements",
				Description:    "GDPR and Dutch data protection compliance",
				Type:           "privacy",
				Category:       "privacy",
				IsMandatory:    true,
				EffectiveDate:  time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
				RegulatoryBody: "AP",
				Requirements: []string{
					"Data protection impact assessment",
					"Privacy by design",
					"Data subject rights",
				},
				LocalizedName: map[string]string{
					"nl": "AVG-vereisten",
					"en": "GDPR Requirements",
				},
				LocalizedDesc: map[string]string{
					"nl": "AVG en Nederlandse gegevensbescherming naleving",
					"en": "GDPR and Dutch data protection compliance",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				ID:           "nl_kvk_validation",
				Name:         "KVK Validation",
				Field:        "kvk_number",
				Type:         "regex",
				Pattern:      "^\\d{8}$",
				Required:     true,
				ErrorMessage: "Invalid KVK number format",
				LocalizedError: map[string]string{
					"nl": "Ongeldig KVK-nummerformaat",
					"en": "Invalid KVK number format",
				},
			},
		},
		SanctionsLists: []string{"EU", "UN", "OFAC"},
		RegulatoryBodies: []RegulatoryBody{
			{
				ID:           "ap",
				Name:         "Autoriteit Persoonsgegevens",
				Acronym:      "AP",
				Type:         "data_protection",
				Jurisdiction: "Netherlands",
				Website:      "https://www.autoriteitpersoonsgegevens.nl",
				Responsibilities: []string{
					"Data protection supervision",
					"GDPR enforcement",
					"Privacy rights protection",
				},
				LocalizedName: map[string]string{
					"nl": "Autoriteit Persoonsgegevens",
					"en": "Dutch Data Protection Authority",
				},
			},
		},
		DataResidencyRules: DataResidencyRules{
			RequiresLocalStorage: true,
			AllowedRegions:       []string{"NL", "EU"},
			CrossBorderTransfer:  false,
			RetentionPeriod:      1095, // 3 years
		},
	}

	// Italy
	configs["IT"] = &CountryConfig{
		Code:     "IT",
		Name:     "Italy",
		Region:   "Europe",
		Currency: "EUR",
		Language: "it",
		Timezone: "Europe/Rome",
		RiskFactors: []RiskFactor{
			{
				ID:          "it_gdpr_risk",
				Name:        "GDPR Compliance Risk",
				Description: "GDPR and Italian data protection law compliance",
				Category:    "privacy",
				Severity:    "high",
				Weight:      0.5,
				IsActive:    true,
				LocalizedName: map[string]string{
					"it": "Rischio di conformità GDPR",
					"en": "GDPR Compliance Risk",
				},
				LocalizedDesc: map[string]string{
					"it": "Conformità GDPR e legge italiana sulla protezione dei dati",
					"en": "GDPR and Italian data protection law compliance",
				},
			},
		},
		ComplianceRules: []ComplianceRule{
			{
				ID:             "it_gdpr_requirement",
				Name:           "GDPR Requirements",
				Description:    "GDPR and Italian data protection compliance",
				Type:           "privacy",
				Category:       "privacy",
				IsMandatory:    true,
				EffectiveDate:  time.Date(2018, 5, 25, 0, 0, 0, 0, time.UTC),
				RegulatoryBody: "Garante",
				Requirements: []string{
					"Data protection impact assessment",
					"Privacy by design",
					"Data subject rights",
				},
				LocalizedName: map[string]string{
					"it": "Requisiti GDPR",
					"en": "GDPR Requirements",
				},
				LocalizedDesc: map[string]string{
					"it": "Conformità GDPR e protezione dei dati italiana",
					"en": "GDPR and Italian data protection compliance",
				},
			},
		},
		ValidationRules: []ValidationRule{
			{
				ID:           "it_codice_fiscale",
				Name:         "Codice Fiscale Validation",
				Field:        "codice_fiscale",
				Type:         "regex",
				Pattern:      "^[A-Z]{6}\\d{2}[A-Z]\\d{2}[A-Z]\\d{3}[A-Z]$",
				Required:     true,
				ErrorMessage: "Invalid Codice Fiscale format",
				LocalizedError: map[string]string{
					"it": "Formato Codice Fiscale non valido",
					"en": "Invalid Codice Fiscale format",
				},
			},
		},
		SanctionsLists: []string{"EU", "UN", "OFAC"},
		RegulatoryBodies: []RegulatoryBody{
			{
				ID:           "garante",
				Name:         "Garante per la protezione dei dati personali",
				Acronym:      "Garante",
				Type:         "data_protection",
				Jurisdiction: "Italy",
				Website:      "https://www.gpdp.it",
				Responsibilities: []string{
					"Data protection supervision",
					"GDPR enforcement",
					"Privacy rights protection",
				},
				LocalizedName: map[string]string{
					"it": "Garante per la protezione dei dati personali",
					"en": "Italian Data Protection Authority",
				},
			},
		},
		DataResidencyRules: DataResidencyRules{
			RequiresLocalStorage: true,
			AllowedRegions:       []string{"IT", "EU"},
			CrossBorderTransfer:  false,
			RetentionPeriod:      1095, // 3 years
		},
	}

	return configs
}
