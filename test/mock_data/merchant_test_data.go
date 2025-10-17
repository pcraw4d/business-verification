package mock_data

import (
	"fmt"
	"math/rand"
	"time"

	"kyb-platform/internal/models"
)

// TestDataSets provides comprehensive test data for different scenarios
type TestDataSets struct {
	BasicMerchants    []*models.Merchant
	EdgeCaseMerchants []*models.Merchant
	PerformanceData   []*models.Merchant
	ValidationData    []*models.Merchant
	BulkOperationData []*models.Merchant
	ComparisonData    []*models.Merchant
	SessionData       []*models.MerchantSession
	AuditLogData      []*models.AuditLog
	NotificationData  []*models.MerchantNotification
	AnalyticsData     []*models.MerchantAnalytics
}

// GetTestDataSets returns all test data sets
func GetTestDataSets() *TestDataSets {
	return &TestDataSets{
		BasicMerchants:    GetBasicMerchants(),
		EdgeCaseMerchants: GetEdgeCaseMerchants(),
		PerformanceData:   GetPerformanceTestData(),
		ValidationData:    GetValidationTestData(),
		BulkOperationData: GetBulkOperationData(),
		ComparisonData:    GetComparisonData(),
		SessionData:       GetSessionTestData(),
		AuditLogData:      GetAuditLogTestData(),
		NotificationData:  GetNotificationTestData(),
		AnalyticsData:     GetAnalyticsTestData(),
	}
}

// GetBasicMerchants returns basic test merchants for standard testing
func GetBasicMerchants() []*models.Merchant {
	now := time.Now()

	return []*models.Merchant{
		{
			ID:                 "test_merchant_001",
			Name:               "Test Tech Solutions",
			LegalName:          "Test Tech Solutions Inc.",
			RegistrationNumber: "REG001001",
			TaxID:              "TAX001001001",
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "Corporation",
			FoundedDate:        timePtr(now.AddDate(-5, -3, -15)),
			EmployeeCount:      25,
			AnnualRevenue:      float64Ptr(1500000.0),
			Address: models.Address{
				Street1:     "123 Test Street",
				City:        "Test City",
				State:       "CA",
				PostalCode:  "90210",
				Country:     "USA",
				CountryCode: "US",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-555-123-4567",
				Email:          "test@testtech.com",
				Website:        "https://testtech.com",
				PrimaryContact: "John Test",
			},
			PortfolioType:    models.PortfolioTypeOnboarded,
			RiskLevel:        models.RiskLevelLow,
			ComplianceStatus: "compliant",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -1, -5),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
		{
			ID:                 "test_merchant_002",
			Name:               "Sample Retail Store",
			LegalName:          "Sample Retail Store LLC",
			RegistrationNumber: "REG001002",
			TaxID:              "TAX001002002",
			Industry:           "Retail",
			IndustryCode:       "448140",
			BusinessType:       "LLC",
			FoundedDate:        timePtr(now.AddDate(-3, -6, -10)),
			EmployeeCount:      12,
			AnnualRevenue:      float64Ptr(850000.0),
			Address: models.Address{
				Street1:     "456 Sample Avenue",
				City:        "Sample City",
				State:       "NY",
				PostalCode:  "10001",
				Country:     "USA",
				CountryCode: "US",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-555-234-5678",
				Email:          "info@sampleshop.com",
				Website:        "https://sampleshop.com",
				PrimaryContact: "Jane Sample",
			},
			PortfolioType:    models.PortfolioTypeProspective,
			RiskLevel:        models.RiskLevelMedium,
			ComplianceStatus: "pending",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -2, -10),
			UpdatedAt:        now.AddDate(0, 0, -2),
		},
		{
			ID:                 "test_merchant_003",
			Name:               "Demo Manufacturing Co.",
			LegalName:          "Demo Manufacturing Company",
			RegistrationNumber: "REG001003",
			TaxID:              "TAX001003003",
			Industry:           "Manufacturing",
			IndustryCode:       "332710",
			BusinessType:       "Corporation",
			FoundedDate:        timePtr(now.AddDate(-10, -2, -20)),
			EmployeeCount:      150,
			AnnualRevenue:      float64Ptr(12000000.0),
			Address: models.Address{
				Street1:     "789 Demo Drive",
				City:        "Demo City",
				State:       "TX",
				PostalCode:  "77001",
				Country:     "USA",
				CountryCode: "US",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-555-345-6789",
				Email:          "contact@demomfg.com",
				Website:        "https://demomfg.com",
				PrimaryContact: "Bob Demo",
			},
			PortfolioType:    models.PortfolioTypePending,
			RiskLevel:        models.RiskLevelHigh,
			ComplianceStatus: "under_review",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -3, -15),
			UpdatedAt:        now.AddDate(0, 0, -3),
		},
	}
}

// GetEdgeCaseMerchants returns merchants with edge cases and boundary conditions
func GetEdgeCaseMerchants() []*models.Merchant {
	now := time.Now()

	return []*models.Merchant{
		// Minimum data merchant
		{
			ID:                 "edge_case_001",
			Name:               "Minimal Business",
			LegalName:          "Minimal Business",
			RegistrationNumber: "MIN001",
			TaxID:              "MIN001",
			Industry:           "Other",
			IndustryCode:       "999999",
			BusinessType:       "Sole Proprietorship",
			EmployeeCount:      1,
			Address: models.Address{
				Street1:    "1 Main St",
				City:       "Small Town",
				State:      "UT",
				PostalCode: "84001",
				Country:    "USA",
			},
			ContactInfo: models.ContactInfo{
				Phone: "+1-555-000-0001",
			},
			PortfolioType:    models.PortfolioTypeProspective,
			RiskLevel:        models.RiskLevelLow,
			ComplianceStatus: "pending",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -1, -1),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
		// Maximum data merchant
		{
			ID:                 "edge_case_002",
			Name:               "Maximum Data Corporation with Very Long Name That Exceeds Normal Limits",
			LegalName:          "Maximum Data Corporation with Very Long Legal Name That Exceeds Normal Limits and Should Be Tested for Boundary Conditions",
			RegistrationNumber: "MAX999999",
			TaxID:              "MAX999999999",
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "Corporation",
			FoundedDate:        timePtr(now.AddDate(-50, 0, 0)),
			EmployeeCount:      10000,
			AnnualRevenue:      float64Ptr(999999999.99),
			Address: models.Address{
				Street1:     "99999 Maximum Data Street with Very Long Street Name",
				Street2:     "Suite 99999 Maximum Data Building",
				City:        "Maximum Data City with Very Long City Name",
				State:       "CA",
				PostalCode:  "99999",
				Country:     "United States of America",
				CountryCode: "US",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-999-999-9999",
				Email:          "maximum.data@maximumdatacorporation.com",
				Website:        "https://maximumdatacorporation.com",
				PrimaryContact: "Maximum Data Contact Person with Very Long Name",
			},
			PortfolioType:    models.PortfolioTypeOnboarded,
			RiskLevel:        models.RiskLevelHigh,
			ComplianceStatus: "compliant",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(-10, 0, 0),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
		// Special characters merchant
		{
			ID:                 "edge_case_003",
			Name:               "Special Chars & Co. (Ltd.)",
			LegalName:          "Special Characters & Company (Limited)",
			RegistrationNumber: "SPEC@001",
			TaxID:              "TAX-SPEC-001",
			Industry:           "Professional Services",
			IndustryCode:       "541110",
			BusinessType:       "LLC",
			FoundedDate:        timePtr(now.AddDate(-2, -1, -1)),
			EmployeeCount:      5,
			AnnualRevenue:      float64Ptr(500000.0),
			Address: models.Address{
				Street1:    "123 Special St. #456",
				City:       "Special City",
				State:      "FL",
				PostalCode: "33101",
				Country:    "USA",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-555-SPEC-001",
				Email:          "special@specialchars.com",
				Website:        "https://special-chars.com",
				PrimaryContact: "Special Contact",
			},
			PortfolioType:    models.PortfolioTypeDeactivated,
			RiskLevel:        models.RiskLevelMedium,
			ComplianceStatus: "non_compliant",
			Status:           "inactive",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -6, -1),
			UpdatedAt:        now.AddDate(0, -1, -1),
		},
	}
}

// GetPerformanceTestData returns large datasets for performance testing
func GetPerformanceTestData() []*models.Merchant {
	merchants := make([]*models.Merchant, 1000)
	now := time.Now()

	// Generate 1000 merchants for performance testing
	for i := 0; i < 1000; i++ {
		merchants[i] = &models.Merchant{
			ID:                 fmt.Sprintf("perf_merchant_%04d", i+1),
			Name:               fmt.Sprintf("Performance Test Business %d", i+1),
			LegalName:          fmt.Sprintf("Performance Test Business %d Inc.", i+1),
			RegistrationNumber: fmt.Sprintf("PERF%06d", i+1),
			TaxID:              fmt.Sprintf("PERF%09d", i+1),
			Industry:           getRandomIndustry(),
			IndustryCode:       getRandomIndustryCode(),
			BusinessType:       getRandomBusinessType(),
			FoundedDate:        timePtr(now.AddDate(-rand.Intn(20), -rand.Intn(12), -rand.Intn(30))),
			EmployeeCount:      rand.Intn(500) + 1,
			AnnualRevenue:      float64Ptr(float64(rand.Intn(10000000) + 100000)),
			Address: models.Address{
				Street1:     fmt.Sprintf("%d Performance St", i+1),
				City:        fmt.Sprintf("Perf City %d", (i%50)+1),
				State:       getRandomState(),
				PostalCode:  fmt.Sprintf("%05d", 10000+i),
				Country:     "USA",
				CountryCode: "US",
			},
			ContactInfo: models.ContactInfo{
				Phone:          fmt.Sprintf("+1-555-%04d", i+1),
				Email:          fmt.Sprintf("perf%d@perftest.com", i+1),
				Website:        fmt.Sprintf("https://perf%d.com", i+1),
				PrimaryContact: fmt.Sprintf("Perf Contact %d", i+1),
			},
			PortfolioType:    getRandomPortfolioType(),
			RiskLevel:        getRandomRiskLevel(),
			ComplianceStatus: getRandomComplianceStatus(),
			Status:           "active",
			CreatedBy:        "perf_test_user",
			CreatedAt:        now.AddDate(0, -rand.Intn(12), -rand.Intn(30)),
			UpdatedAt:        now.AddDate(0, 0, -1-rand.Intn(7)),
		}
	}

	return merchants
}

// GetValidationTestData returns merchants for validation testing
func GetValidationTestData() []*models.Merchant {
	now := time.Now()

	return []*models.Merchant{
		// Invalid merchant (empty name)
		{
			ID:                 "invalid_001",
			Name:               "", // Invalid: empty name
			LegalName:          "Invalid Test Legal Name",
			RegistrationNumber: "INV001",
			TaxID:              "INV001",
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "Corporation",
			EmployeeCount:      10,
			Address: models.Address{
				Street1:    "123 Invalid St",
				City:       "Invalid City",
				State:      "CA",
				PostalCode: "90210",
				Country:    "USA",
			},
			ContactInfo: models.ContactInfo{
				Phone: "+1-555-INV-001",
			},
			PortfolioType:    models.PortfolioTypeOnboarded,
			RiskLevel:        models.RiskLevelLow,
			ComplianceStatus: "compliant",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -1, -1),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
		// Invalid portfolio type
		{
			ID:                 "invalid_002",
			Name:               "Invalid Portfolio Type Test",
			LegalName:          "Invalid Portfolio Type Test Inc.",
			RegistrationNumber: "INV002",
			TaxID:              "INV002",
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "Corporation",
			EmployeeCount:      10,
			Address: models.Address{
				Street1:    "123 Invalid St",
				City:       "Invalid City",
				State:      "CA",
				PostalCode: "90210",
				Country:    "USA",
			},
			ContactInfo: models.ContactInfo{
				Phone: "+1-555-INV-002",
			},
			PortfolioType:    "invalid_type", // Invalid portfolio type
			RiskLevel:        models.RiskLevelLow,
			ComplianceStatus: "compliant",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -1, -1),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
		// Invalid risk level
		{
			ID:                 "invalid_003",
			Name:               "Invalid Risk Level Test",
			LegalName:          "Invalid Risk Level Test Inc.",
			RegistrationNumber: "INV003",
			TaxID:              "INV003",
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "Corporation",
			EmployeeCount:      10,
			Address: models.Address{
				Street1:    "123 Invalid St",
				City:       "Invalid City",
				State:      "CA",
				PostalCode: "90210",
				Country:    "USA",
			},
			ContactInfo: models.ContactInfo{
				Phone: "+1-555-INV-003",
			},
			PortfolioType:    models.PortfolioTypeOnboarded,
			RiskLevel:        "invalid_risk", // Invalid risk level
			ComplianceStatus: "compliant",
			Status:           "active",
			CreatedBy:        "test_user",
			CreatedAt:        now.AddDate(0, -1, -1),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
	}
}

// GetBulkOperationData returns merchants for bulk operation testing
func GetBulkOperationData() []*models.Merchant {
	merchants := make([]*models.Merchant, 100)
	now := time.Now()

	for i := 0; i < 100; i++ {
		merchants[i] = &models.Merchant{
			ID:                 fmt.Sprintf("bulk_merchant_%03d", i+1),
			Name:               fmt.Sprintf("Bulk Test Business %d", i+1),
			LegalName:          fmt.Sprintf("Bulk Test Business %d LLC", i+1),
			RegistrationNumber: fmt.Sprintf("BULK%06d", i+1),
			TaxID:              fmt.Sprintf("BULK%09d", i+1),
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "LLC",
			FoundedDate:        timePtr(now.AddDate(-2, -i%12, -i%30)),
			EmployeeCount:      (i % 50) + 1,
			AnnualRevenue:      float64Ptr(float64((i % 1000000) + 100000)),
			Address: models.Address{
				Street1:    fmt.Sprintf("%d Bulk St", i+1),
				City:       "Bulk City",
				State:      "CA",
				PostalCode: "90210",
				Country:    "USA",
			},
			ContactInfo: models.ContactInfo{
				Phone:          fmt.Sprintf("+1-555-BULK-%03d", i+1),
				Email:          fmt.Sprintf("bulk%d@bulktest.com", i+1),
				Website:        fmt.Sprintf("https://bulk%d.com", i+1),
				PrimaryContact: fmt.Sprintf("Bulk Contact %d", i+1),
			},
			PortfolioType:    models.PortfolioTypeProspective,
			RiskLevel:        models.RiskLevelMedium,
			ComplianceStatus: "pending",
			Status:           "active",
			CreatedBy:        "bulk_test_user",
			CreatedAt:        now.AddDate(0, -1, -i%30),
			UpdatedAt:        now.AddDate(0, 0, -1-i%7),
		}
	}

	return merchants
}

// GetComparisonData returns merchants specifically for comparison testing
func GetComparisonData() []*models.Merchant {
	now := time.Now()

	return []*models.Merchant{
		{
			ID:                 "compare_merchant_001",
			Name:               "Comparison Tech Corp",
			LegalName:          "Comparison Tech Corporation",
			RegistrationNumber: "COMP001",
			TaxID:              "COMP001",
			Industry:           "Technology",
			IndustryCode:       "541511",
			BusinessType:       "Corporation",
			FoundedDate:        timePtr(now.AddDate(-5, 0, 0)),
			EmployeeCount:      50,
			AnnualRevenue:      float64Ptr(5000000.0),
			Address: models.Address{
				Street1:    "123 Compare St",
				City:       "Compare City",
				State:      "CA",
				PostalCode: "90210",
				Country:    "USA",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-555-COMP-001",
				Email:          "compare1@compare.com",
				Website:        "https://compare1.com",
				PrimaryContact: "Compare Contact 1",
			},
			PortfolioType:    models.PortfolioTypeOnboarded,
			RiskLevel:        models.RiskLevelLow,
			ComplianceStatus: "compliant",
			Status:           "active",
			CreatedBy:        "compare_test_user",
			CreatedAt:        now.AddDate(0, -6, 0),
			UpdatedAt:        now.AddDate(0, 0, -1),
		},
		{
			ID:                 "compare_merchant_002",
			Name:               "Comparison Retail LLC",
			LegalName:          "Comparison Retail LLC",
			RegistrationNumber: "COMP002",
			TaxID:              "COMP002",
			Industry:           "Retail",
			IndustryCode:       "448140",
			BusinessType:       "LLC",
			FoundedDate:        timePtr(now.AddDate(-3, 0, 0)),
			EmployeeCount:      25,
			AnnualRevenue:      float64Ptr(2500000.0),
			Address: models.Address{
				Street1:    "456 Compare Ave",
				City:       "Compare City",
				State:      "NY",
				PostalCode: "10001",
				Country:    "USA",
			},
			ContactInfo: models.ContactInfo{
				Phone:          "+1-555-COMP-002",
				Email:          "compare2@compare.com",
				Website:        "https://compare2.com",
				PrimaryContact: "Compare Contact 2",
			},
			PortfolioType:    models.PortfolioTypeProspective,
			RiskLevel:        models.RiskLevelMedium,
			ComplianceStatus: "pending",
			Status:           "active",
			CreatedBy:        "compare_test_user",
			CreatedAt:        now.AddDate(0, -3, 0),
			UpdatedAt:        now.AddDate(0, 0, -2),
		},
	}
}

// GetSessionTestData returns test data for session management
func GetSessionTestData() []*models.MerchantSession {
	now := time.Now()

	return []*models.MerchantSession{
		{
			ID:         "session_001",
			UserID:     "test_user_001",
			MerchantID: "test_merchant_001",
			StartedAt:  now.Add(-2 * time.Hour),
			LastActive: now.Add(-5 * time.Minute),
			IsActive:   true,
			CreatedAt:  now.Add(-2 * time.Hour),
			UpdatedAt:  now.Add(-5 * time.Minute),
		},
		{
			ID:         "session_002",
			UserID:     "test_user_002",
			MerchantID: "test_merchant_002",
			StartedAt:  now.Add(-1 * time.Hour),
			LastActive: now.Add(-10 * time.Minute),
			IsActive:   true,
			CreatedAt:  now.Add(-1 * time.Hour),
			UpdatedAt:  now.Add(-10 * time.Minute),
		},
		{
			ID:         "session_003",
			UserID:     "test_user_001",
			MerchantID: "test_merchant_003",
			StartedAt:  now.Add(-25 * time.Hour), // Expired session
			LastActive: now.Add(-25 * time.Hour),
			IsActive:   false,
			CreatedAt:  now.Add(-25 * time.Hour),
			UpdatedAt:  now.Add(-25 * time.Hour),
		},
	}
}

// GetAuditLogTestData returns test data for audit logging
func GetAuditLogTestData() []*models.AuditLog {
	now := time.Now()

	return []*models.AuditLog{
		{
			ID:           "audit_001",
			UserID:       "test_user_001",
			MerchantID:   "test_merchant_001",
			Action:       "CREATE",
			ResourceType: "merchant",
			ResourceID:   "test_merchant_001",
			Details:      "Created new merchant",
			IPAddress:    "192.168.1.100",
			UserAgent:    "Mozilla/5.0 (Test Browser)",
			RequestID:    "req_001",
			CreatedAt:    now.Add(-1 * time.Hour),
		},
		{
			ID:           "audit_002",
			UserID:       "test_user_001",
			MerchantID:   "test_merchant_001",
			Action:       "UPDATE",
			ResourceType: "merchant",
			ResourceID:   "test_merchant_001",
			Details:      "Updated merchant portfolio type",
			IPAddress:    "192.168.1.100",
			UserAgent:    "Mozilla/5.0 (Test Browser)",
			RequestID:    "req_002",
			CreatedAt:    now.Add(-30 * time.Minute),
		},
		{
			ID:           "audit_003",
			UserID:       "test_user_002",
			MerchantID:   "test_merchant_002",
			Action:       "VIEW",
			ResourceType: "merchant",
			ResourceID:   "test_merchant_002",
			Details:      "Viewed merchant details",
			IPAddress:    "192.168.1.101",
			UserAgent:    "Mozilla/5.0 (Test Browser)",
			RequestID:    "req_003",
			CreatedAt:    now.Add(-15 * time.Minute),
		},
	}
}

// GetNotificationTestData returns test data for notifications
func GetNotificationTestData() []*models.MerchantNotification {
	now := time.Now()

	return []*models.MerchantNotification{
		{
			ID:         "notif_001",
			MerchantID: "test_merchant_001",
			UserID:     "test_user_001",
			Type:       string(models.NotificationTypeRiskAlert),
			Title:      "Risk Level Changed",
			Message:    "Merchant risk level has been updated to High",
			IsRead:     false,
			Priority:   string(models.NotificationPriorityHigh),
			CreatedAt:  now.Add(-1 * time.Hour),
		},
		{
			ID:         "notif_002",
			MerchantID: "test_merchant_002",
			UserID:     "test_user_002",
			Type:       string(models.NotificationTypeCompliance),
			Title:      "Compliance Review Required",
			Message:    "Merchant requires compliance review",
			IsRead:     true,
			Priority:   string(models.NotificationPriorityMedium),
			CreatedAt:  now.Add(-2 * time.Hour),
			ReadAt:     timePtr(now.Add(-1 * time.Hour)),
		},
		{
			ID:         "notif_003",
			MerchantID: "test_merchant_003",
			UserID:     "test_user_001",
			Type:       string(models.NotificationTypeBulkOperation),
			Title:      "Bulk Operation Completed",
			Message:    "Bulk portfolio type update completed successfully",
			IsRead:     false,
			Priority:   string(models.NotificationPriorityLow),
			CreatedAt:  now.Add(-30 * time.Minute),
		},
	}
}

// GetAnalyticsTestData returns test data for analytics
func GetAnalyticsTestData() []*models.MerchantAnalytics {
	now := time.Now()

	return []*models.MerchantAnalytics{
		{
			MerchantID:        "test_merchant_001",
			RiskScore:         0.25,
			ComplianceScore:   0.95,
			TransactionVolume: 150000.0,
			LastActivity:      timePtr(now.Add(-1 * time.Hour)),
			Flags:             []string{"low_risk", "high_compliance"},
			Metadata: map[string]interface{}{
				"risk_factors": []string{"stable_revenue", "good_compliance_history"},
				"alerts":       0,
			},
			CalculatedAt: now.Add(-1 * time.Hour),
			UpdatedAt:    now.Add(-1 * time.Hour),
		},
		{
			MerchantID:        "test_merchant_002",
			RiskScore:         0.65,
			ComplianceScore:   0.70,
			TransactionVolume: 85000.0,
			LastActivity:      timePtr(now.Add(-2 * time.Hour)),
			Flags:             []string{"medium_risk", "pending_compliance"},
			Metadata: map[string]interface{}{
				"risk_factors": []string{"new_business", "limited_history"},
				"alerts":       2,
			},
			CalculatedAt: now.Add(-2 * time.Hour),
			UpdatedAt:    now.Add(-2 * time.Hour),
		},
		{
			MerchantID:        "test_merchant_003",
			RiskScore:         0.85,
			ComplianceScore:   0.45,
			TransactionVolume: 25000.0,
			LastActivity:      timePtr(now.Add(-3 * time.Hour)),
			Flags:             []string{"high_risk", "compliance_issues"},
			Metadata: map[string]interface{}{
				"risk_factors": []string{"high_risk_industry", "compliance_violations"},
				"alerts":       5,
			},
			CalculatedAt: now.Add(-3 * time.Hour),
			UpdatedAt:    now.Add(-3 * time.Hour),
		},
	}
}

// Helper functions for generating random data

func getRandomIndustry() string {
	industries := []string{"Technology", "Finance", "Healthcare", "Retail", "Manufacturing", "Professional Services", "Construction", "Food & Beverage", "Transportation"}
	return industries[rand.Intn(len(industries))]
}

func getRandomIndustryCode() string {
	codes := []string{"541511", "522110", "621111", "448140", "332710", "541110", "236220", "311812", "484121"}
	return codes[rand.Intn(len(codes))]
}

func getRandomBusinessType() string {
	types := []string{"Corporation", "LLC", "Partnership", "Sole Proprietorship", "LLP"}
	return types[rand.Intn(len(types))]
}

func getRandomPortfolioType() models.PortfolioType {
	types := []models.PortfolioType{
		models.PortfolioTypeOnboarded,
		models.PortfolioTypeProspective,
		models.PortfolioTypePending,
		models.PortfolioTypeDeactivated,
	}
	return types[rand.Intn(len(types))]
}

func getRandomRiskLevel() models.RiskLevel {
	levels := []models.RiskLevel{
		models.RiskLevelLow,
		models.RiskLevelMedium,
		models.RiskLevelHigh,
	}
	return levels[rand.Intn(len(levels))]
}

func getRandomComplianceStatus() string {
	statuses := []string{"compliant", "pending", "non_compliant", "under_review"}
	return statuses[rand.Intn(len(statuses))]
}

func getRandomState() string {
	states := []string{"CA", "NY", "TX", "FL", "IL", "PA", "OH", "GA", "NC", "MI"}
	return states[rand.Intn(len(states))]
}

// Utility functions

func timePtr(t time.Time) *time.Time {
	return &t
}

func float64Ptr(f float64) *float64 {
	return &f
}

// GetTestDataByScenario returns test data for specific testing scenarios
func GetTestDataByScenario(scenario string) interface{} {
	switch scenario {
	case "basic":
		return GetBasicMerchants()
	case "edge_cases":
		return GetEdgeCaseMerchants()
	case "performance":
		return GetPerformanceTestData()
	case "validation":
		return GetValidationTestData()
	case "bulk_operations":
		return GetBulkOperationData()
	case "comparison":
		return GetComparisonData()
	case "sessions":
		return GetSessionTestData()
	case "audit_logs":
		return GetAuditLogTestData()
	case "notifications":
		return GetNotificationTestData()
	case "analytics":
		return GetAnalyticsTestData()
	default:
		return GetTestDataSets()
	}
}

// GetTestDataCount returns the count of test data items for each scenario
func GetTestDataCount(scenario string) int {
	switch scenario {
	case "basic":
		return len(GetBasicMerchants())
	case "edge_cases":
		return len(GetEdgeCaseMerchants())
	case "performance":
		return len(GetPerformanceTestData())
	case "validation":
		return len(GetValidationTestData())
	case "bulk_operations":
		return len(GetBulkOperationData())
	case "comparison":
		return len(GetComparisonData())
	case "sessions":
		return len(GetSessionTestData())
	case "audit_logs":
		return len(GetAuditLogTestData())
	case "notifications":
		return len(GetNotificationTestData())
	case "analytics":
		return len(GetAnalyticsTestData())
	default:
		return 0
	}
}
