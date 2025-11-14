package fixtures

import (
	"time"

	"kyb-platform/internal/models"
)

// GetMerchantsWithRiskAssessments returns test merchants with complete risk assessment data
func GetMerchantsWithRiskAssessments() []*models.Merchant {
	now := time.Now()

	return []*models.Merchant{
		{
			ID:                 "merchant-risk-pending-001",
			Name:               "Pending Assessment Company",
			LegalName:          "Pending Assessment Company Inc.",
			RegistrationNumber: "REG-PENDING-001",
			TaxID:              "TAX-PENDING-001",
			Industry:           "Financial Services",
			IndustryCode:       "522110",
			BusinessType:       "Corporation",
			FoundedDate:        timePtr(now.AddDate(-2, 0, 0)),
			EmployeeCount:      50,
			AnnualRevenue:      float64Ptr(2000000.0),
			Address: models.Address{
				Street1:     "100 Risk Street",
				City:        "Risk City",
				State:       "CA",
				PostalCode:  "94105",
				Country:     "United States",
				CountryCode: "US",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-555-100-0001",
				Email:          "contact@pendingrisk.com",
				Website:        "https://www.pendingrisk.com",
				PrimaryContact: "Jane Risk",
			},
			PortfolioType:    models.PortfolioTypeOnboarded,
			RiskLevel:        models.RiskLevelMedium,
			ComplianceStatus: "pending",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -1, 0),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
		{
			ID:                 "merchant-risk-completed-002",
			Name:               "Completed Assessment Company",
			LegalName:          "Completed Assessment Company LLC",
			RegistrationNumber: "REG-COMPLETED-002",
			TaxID:              "TAX-COMPLETED-002",
			Industry:           "Healthcare",
			IndustryCode:       "621111",
			BusinessType:       "LLC",
			FoundedDate:        timePtr(now.AddDate(-5, 0, 0)),
			EmployeeCount:      120,
			AnnualRevenue:      float64Ptr(5000000.0),
			Address: models.Address{
				Street1:     "200 Completed Ave",
				City:        "Completed City",
				State:       "NY",
				PostalCode:  "10001",
				Country:     "United States",
				CountryCode: "US",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-555-200-0002",
				Email:          "contact@completedrisk.com",
				Website:        "https://www.completedrisk.com",
				PrimaryContact: "Bob Complete",
			},
			PortfolioType:    models.PortfolioTypeOnboarded,
			RiskLevel:        models.RiskLevelLow,
			ComplianceStatus: "compliant",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -2, 0),
			UpdatedAt:        now.AddDate(0, 0, -2),
		},
		{
			ID:                 "merchant-risk-failed-003",
			Name:               "Failed Assessment Company",
			LegalName:          "Failed Assessment Company Inc.",
			RegistrationNumber: "REG-FAILED-003",
			TaxID:              "TAX-FAILED-003",
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "Corporation",
			FoundedDate:        timePtr(now.AddDate(-1, 0, 0)),
			EmployeeCount:      30,
			AnnualRevenue:      float64Ptr(1500000.0),
			Address: models.Address{
				Street1:     "300 Failed Blvd",
				City:        "Failed City",
				State:       "TX",
				PostalCode:  "75001",
				Country:     "United States",
				CountryCode: "US",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-555-300-0003",
				Email:          "contact@failedrisk.com",
				Website:        "https://www.failedrisk.com",
				PrimaryContact: "Alice Failed",
			},
			PortfolioType:    models.PortfolioTypeProspective,
			RiskLevel:        models.RiskLevelHigh,
			ComplianceStatus: "non_compliant",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, 0, -5),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
	}
}

// GetMerchantsWithAnalytics returns test merchants with various analytics completeness levels
func GetMerchantsWithAnalytics() []*models.Merchant {
	now := time.Now()

	return []*models.Merchant{
		// Complete analytics data
		{
			ID:                 "merchant-analytics-complete-001",
			Name:               "Complete Analytics Company",
			LegalName:          "Complete Analytics Company Inc.",
			RegistrationNumber: "REG-ANALYTICS-COMPLETE-001",
			TaxID:              "TAX-ANALYTICS-COMPLETE-001",
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "Corporation",
			FoundedDate:        timePtr(now.AddDate(-8, 0, 0)),
			EmployeeCount:      200,
			AnnualRevenue:      float64Ptr(10000000.0),
			Address: models.Address{
				Street1:     "400 Complete Street",
				Street2:     "Suite 200",
				City:        "Complete City",
				State:       "CA",
				PostalCode:  "94025",
				Country:     "United States",
				CountryCode: "US",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-555-400-0004",
				Email:          "contact@completeanalytics.com",
				Website:        "https://www.completeanalytics.com",
				PrimaryContact: "David Complete",
			},
			PortfolioType:    models.PortfolioTypeOnboarded,
			RiskLevel:        models.RiskLevelLow,
			ComplianceStatus: "compliant",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -6, 0),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
		// Partial analytics data
		{
			ID:                 "merchant-analytics-partial-002",
			Name:               "Partial Analytics Company",
			LegalName:          "Partial Analytics Company LLC",
			RegistrationNumber: "REG-ANALYTICS-PARTIAL-002",
			Industry:           "Retail",
			IndustryCode:       "448140",
			BusinessType:       "LLC",
			FoundedDate:        timePtr(now.AddDate(-3, 0, 0)),
			EmployeeCount:      45,
			// Missing annual revenue
			Address: models.Address{
				Street1:     "500 Partial Ave",
				City:        "Partial City",
				State:       "NY",
				PostalCode:  "10002",
				Country:     "United States",
				CountryCode: "US",
			},
			ContactInfo: models.ContactInfo{
				Email:   "contact@partialanalytics.com",
				Website: "https://www.partialanalytics.com",
				// Missing phone and primary contact
			},
			PortfolioType:    models.PortfolioTypeProspective,
			RiskLevel:        models.RiskLevelMedium,
			ComplianceStatus: "pending",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -3, 0),
			UpdatedAt:        now.AddDate(0, 0, -2),
		},
		// Missing analytics data
		{
			ID:                 "merchant-analytics-missing-003",
			Name:               "Missing Analytics Company",
			LegalName:          "Missing Analytics Company",
			Industry:           "Services",
			// Missing registration number, tax ID, industry code, business type
			// Missing founded date, employee count, annual revenue
			Address: models.Address{
				City:        "Missing City",
				State:       "TX",
				Country:     "United States",
				CountryCode: "US",
				// Missing street, postal code
			},
			ContactInfo: models.ContactInfo{
				Email: "contact@missinganalytics.com",
				// Missing phone, website, primary contact
			},
			PortfolioType:    models.PortfolioTypePending,
			RiskLevel:        models.RiskLevelHigh,
			ComplianceStatus: "pending",
			Status:           "pending",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, 0, -1),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
	}
}

// GetMerchantsWithEdgeCases returns test merchants with edge case scenarios
func GetMerchantsWithEdgeCases() []*models.Merchant {
	now := time.Now()

	return []*models.Merchant{
		// Missing required fields
		{
			ID:        "merchant-edge-missing-001",
			Name:      "Missing Fields Company",
			LegalName: "",
			// Missing most required fields
			Status:    "pending",
			CreatedBy: "test_user",
			CreatedAt: now,
			UpdatedAt: now,
		},
		// Invalid data
		{
			ID:                 "merchant-edge-invalid-002",
			Name:               "Invalid Data Company",
			LegalName:          "Invalid Data Company Inc.",
			RegistrationNumber: "INVALID-REG",
			TaxID:              "INVALID-TAX",
			Industry:           "Invalid Industry",
			IndustryCode:       "999999", // Invalid code
			BusinessType:       "InvalidType",
			EmployeeCount:      -1, // Invalid count
			AnnualRevenue:      float64Ptr(-1000.0), // Invalid revenue
			Address: models.Address{
				Street1:     "",
				City:        "",
				State:       "XX", // Invalid state
				PostalCode:  "INVALID",
				Country:     "",
				CountryCode: "XX", // Invalid country code
			},
			ContactInfo: models.ContactInfo{
				Phone:   "invalid-phone",
				Email:   "invalid-email",
				Website: "not-a-url",
			},
			PortfolioType:    models.PortfolioTypePending,
			RiskLevel:        models.RiskLevelMedium,
			ComplianceStatus: "invalid_status",
			Status:           "invalid",
			CreatedBy:        "test_user",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		// Null/empty values
		{
			ID:                 "merchant-edge-null-003",
			Name:               "Null Values Company",
			LegalName:          "Null Values Company LLC",
			RegistrationNumber: "",
			TaxID:              "",
			Industry:           "",
			IndustryCode:       "",
			BusinessType:       "",
			FoundedDate:        nil,
			EmployeeCount:      0,
			AnnualRevenue:      nil,
			Address: models.Address{
				Street1:     "",
				Street2:     "",
				City:        "",
				State:       "",
				PostalCode:  "",
				Country:     "",
				CountryCode: "",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "",
				Email:          "",
				Website:        "",
				PrimaryContact: "",
			},
			PortfolioType:    models.PortfolioTypePending,
			RiskLevel:        models.RiskLevelMedium,
			ComplianceStatus: "",
			Status:           "pending",
			CreatedBy:        "",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		// Very long values
		{
			ID:                 "merchant-edge-long-004",
			Name:               string(make([]byte, 300)), // Very long name
			LegalName:          string(make([]byte, 300)),
			RegistrationNumber: string(make([]byte, 200)),
			TaxID:              string(make([]byte, 200)),
			Industry:           string(make([]byte, 150)),
			IndustryCode:       string(make([]byte, 50)),
			BusinessType:       string(make([]byte, 100)),
			EmployeeCount:      999999,
			AnnualRevenue:      float64Ptr(999999999999.99),
			Address: models.Address{
				Street1:     string(make([]byte, 200)),
				Street2:     string(make([]byte, 200)),
				City:        string(make([]byte, 150)),
				State:       string(make([]byte, 50)),
				PostalCode:  string(make([]byte, 50)),
				Country:     string(make([]byte, 200)),
				CountryCode: string(make([]byte, 10)),
			},
			ContactInfo: models.ContactInfo{
				Phone:          string(make([]byte, 100)),
				Email:          string(make([]byte, 200)),
				Website:        string(make([]byte, 300)),
				PrimaryContact: string(make([]byte, 200)),
			},
			PortfolioType:    models.PortfolioTypeOnboarded,
			RiskLevel:        models.RiskLevelLow,
			ComplianceStatus: "compliant",
			Status:           "active",
			CreatedBy:        string(make([]byte, 100)),
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}
}

// GetRiskAssessmentTestData returns test risk assessment data
func GetRiskAssessmentTestData() []*models.RiskAssessment {
	now := time.Now()
	completedAt := now.Add(-1 * time.Hour)
	estimatedCompletion := now.Add(30 * time.Minute)

	return []*models.RiskAssessment{
		// Pending assessment
		{
			ID:        "assessment-pending-001",
			MerchantID: "merchant-risk-pending-001",
			Status:    models.AssessmentStatusPending,
			Options: models.AssessmentOptions{
				IncludeHistory:    true,
				IncludePredictions: true,
			},
			Progress:            0,
			EstimatedCompletion: &estimatedCompletion,
			CreatedAt:           now.Add(-10 * time.Minute),
			UpdatedAt:           now.Add(-5 * time.Minute),
			CompletedAt:         nil,
		},
		// Processing assessment
		{
			ID:        "assessment-processing-002",
			MerchantID: "merchant-risk-pending-001",
			Status:    models.AssessmentStatusProcessing,
			Options: models.AssessmentOptions{
				IncludeHistory:    true,
				IncludePredictions: false,
			},
			Progress:            45,
			EstimatedCompletion: &estimatedCompletion,
			CreatedAt:           now.Add(-20 * time.Minute),
			UpdatedAt:           now.Add(-1 * time.Minute),
			CompletedAt:         nil,
		},
		// Completed assessment
		{
			ID:        "assessment-completed-003",
			MerchantID: "merchant-risk-completed-002",
			Status:    models.AssessmentStatusCompleted,
			Options: models.AssessmentOptions{
				IncludeHistory:    true,
				IncludePredictions: true,
			},
			Result: &models.RiskAssessmentResult{
				OverallScore: 0.25,
				RiskLevel:    "low",
				Factors: []models.RiskFactor{
					{
						Name:   "Financial Stability",
						Score:  0.2,
						Weight: 0.4,
					},
					{
						Name:   "Compliance History",
						Score:  0.3,
						Weight: 0.3,
					},
					{
						Name:   "Industry Risk",
						Score:  0.25,
						Weight: 0.3,
					},
				},
			},
			Progress:            100,
			EstimatedCompletion: nil,
			CreatedAt:           now.Add(-2 * time.Hour),
			UpdatedAt:           completedAt,
			CompletedAt:         &completedAt,
		},
		// Failed assessment
		{
			ID:        "assessment-failed-004",
			MerchantID: "merchant-risk-failed-003",
			Status:    models.AssessmentStatusFailed,
			Options: models.AssessmentOptions{
				IncludeHistory:    false,
				IncludePredictions: false,
			},
			Progress:            30,
			EstimatedCompletion: nil,
			CreatedAt:           now.Add(-1 * time.Hour),
			UpdatedAt:           now.Add(-30 * time.Minute),
			CompletedAt:         nil,
		},
	}
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func float64Ptr(f float64) *float64 {
	return &f
}

