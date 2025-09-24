package testdata

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/risk"
)

// TestDataFactory provides methods to generate test data
type TestDataFactory struct {
	// Random seed for reproducible test data
	seed int64
}

// NewTestDataFactory creates a new test data factory
func NewTestDataFactory(seed int64) *TestDataFactory {
	return &TestDataFactory{
		seed: seed,
	}
}

// GenerateUser creates a test user
func (f *TestDataFactory) GenerateUser() *database.User {
	now := time.Now()

	return &database.User{
		ID:                  f.generateID("user"),
		Email:               f.generateEmail(),
		Username:            f.generateUsername(),
		PasswordHash:        "$2a$10$hashedpassword", // bcrypt hash
		FirstName:           f.generateFirstName(),
		LastName:            f.generateLastName(),
		Company:             f.generateCompanyName(),
		Role:                f.randomChoice([]string{"user", "admin", "analyst"}),
		Status:              "active",
		EmailVerified:       true,
		LastLoginAt:         &now,
		FailedLoginAttempts: 0,
		LockedUntil:         nil,
		CreatedAt:           now,
		UpdatedAt:           now,
	}
}

// GenerateBusiness creates a test business
func (f *TestDataFactory) GenerateBusiness() *database.Business {
	now := time.Now()
	foundedDate := now.AddDate(-f.randomInt(1, 20), 0, 0)
	employeeCount := f.randomInt(1, 1000)
	annualRevenue := float64(f.randomInt(100000, 10000000))

	return &database.Business{
		ID:                 f.generateID("business"),
		Name:               f.generateCompanyName(),
		LegalName:          f.generateLegalName(),
		RegistrationNumber: f.generateRegistrationNumber(),
		TaxID:              f.generateTaxID(),
		Industry:           f.randomChoice([]string{"Technology", "Finance", "Healthcare", "Retail", "Manufacturing"}),
		IndustryCode:       f.generateIndustryCode(),
		BusinessType:       f.randomChoice([]string{"Corporation", "LLC", "Partnership", "Sole Proprietorship"}),
		FoundedDate:        &foundedDate,
		EmployeeCount:      employeeCount,
		AnnualRevenue:      &annualRevenue,
		Address: database.Address{
			Street1:     f.generateStreetAddress(),
			Street2:     "",
			City:        f.randomChoice([]string{"New York", "Los Angeles", "Chicago", "Houston", "Phoenix"}),
			State:       f.randomChoice([]string{"NY", "CA", "IL", "TX", "AZ"}),
			PostalCode:  f.generatePostalCode(),
			Country:     "United States",
			CountryCode: "US",
		},
		ContactInfo: database.ContactInfo{
			Phone:          f.generatePhoneNumber(),
			Email:          f.generateEmail(),
			Website:        f.generateWebsite(),
			PrimaryContact: f.generateFullName(),
		},
		Status:           "active",
		RiskLevel:        f.randomChoice([]string{"low", "medium", "high"}),
		ComplianceStatus: "compliant",
		CreatedBy:        f.generateID("user"),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// GenerateBusinessClassification creates a test business classification
func (f *TestDataFactory) GenerateBusinessClassification(businessID string) *database.BusinessClassification {
	now := time.Now()

	return &database.BusinessClassification{
		ID:                   f.generateID("classification"),
		BusinessID:           businessID,
		IndustryCode:         f.generateIndustryCode(),
		IndustryName:         f.randomChoice([]string{"Software Development", "Financial Services", "Healthcare Services", "Retail Trade", "Manufacturing"}),
		ConfidenceScore:      f.randomFloat(0.5, 1.0),
		ClassificationMethod: f.randomChoice([]string{"keyword", "fuzzy", "industry", "name"}),
		Source:               "test",
		RawData:              `{"test": "data"}`,
		CreatedAt:            now,
	}
}

// GenerateRiskAssessment creates a test risk assessment
func (f *TestDataFactory) GenerateRiskAssessment(businessID string) *database.RiskAssessment {
	now := time.Now()

	return &database.RiskAssessment{
		ID:               f.generateID("assessment"),
		BusinessID:       businessID,
		BusinessName:     f.generateCompanyName(),
		OverallScore:     f.randomFloat(0.0, 100.0),
		OverallLevel:     f.randomChoice([]string{"low", "medium", "high", "critical"}),
		FactorScores:     []string{"financial:75", "operational:60", "regulatory:80"},
		Recommendations:  []string{"Monitor cash flow", "Improve processes", "Stay compliant"},
		Predictions:      []string{"Risk likely to decrease"},
		Alerts:           []string{},
		AssessmentMethod: "rule_based",
		Source:           "test",
		Metadata:         map[string]interface{}{"test": true},
		AssessedAt:       now,
		ValidUntil:       now.AddDate(0, 1, 0),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// GenerateComplianceCheck creates a test compliance check
func (f *TestDataFactory) GenerateComplianceCheck(businessID string) *database.ComplianceCheck {
	now := time.Now()

	return &database.ComplianceCheck{
		ID:             f.generateID("check"),
		BusinessID:     businessID,
		ComplianceType: f.randomChoice([]string{"kyc_aml", "sanctions", "regulatory"}),
		Status:         f.randomChoice([]string{"passed", "failed", "pending"}),
		Score:          f.randomFloat(0.0, 100.0),
		Requirements:   []string{"kyc_verified", "aml_checked", "sanctions_clear"},
		CheckMethod:    "automated",
		Source:         "test",
		RawData:        `{"test": "data"}`,
		CreatedAt:      now,
	}
}

// GenerateAPIKey creates a test API key
func (f *TestDataFactory) GenerateAPIKey(userID string) *database.APIKey {
	now := time.Now()

	return &database.APIKey{
		ID:          f.generateID("key"),
		UserID:      userID,
		Name:        f.generateAPIKeyName(),
		KeyHash:     f.generateKeyHash(),
		Role:        f.randomChoice([]string{"user", "admin", "readonly"}),
		Permissions: `["read", "write"]`,
		Status:      "active",
		LastUsedAt:  &now,
		ExpiresAt:   nil,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// GenerateAuditLog creates a test audit log
func (f *TestDataFactory) GenerateAuditLog(userID string) *database.AuditLog {
	now := time.Now()

	return &database.AuditLog{
		ID:           f.generateID("audit"),
		UserID:       userID,
		Action:       f.randomChoice([]string{"create", "update", "delete", "view"}),
		ResourceType: f.randomChoice([]string{"user", "business", "classification", "risk"}),
		ResourceID:   f.generateID("resource"),
		Details:      `{"test": "details"}`,
		IPAddress:    f.generateIPAddress(),
		UserAgent:    f.generateUserAgent(),
		RequestID:    f.generateID("request"),
		CreatedAt:    now,
	}
}

// GenerateRiskAssessmentRequest creates a test risk assessment request
func (f *TestDataFactory) GenerateRiskAssessmentRequest() risk.RiskAssessmentRequest {
	return risk.RiskAssessmentRequest{
		BusinessID:         f.generateID("business"),
		BusinessName:       f.generateCompanyName(),
		Categories:         []risk.RiskCategory{risk.RiskCategoryFinancial, risk.RiskCategoryOperational},
		Factors:            []string{"cash_flow", "debt_ratio", "operational_efficiency"},
		IncludeHistory:     true,
		IncludePredictions: true,
		Metadata:           map[string]interface{}{"test": true},
	}
}

// GenerateRiskFactor creates a test risk factor
func (f *TestDataFactory) GenerateRiskFactor() *risk.RiskFactor {
	now := time.Now()

	return &risk.RiskFactor{
		ID:          f.generateID("factor"),
		Name:        f.randomChoice([]string{"Cash Flow Risk", "Debt Ratio Risk", "Operational Risk"}),
		Description: "Test risk factor description",
		Category:    f.randomChoice([]risk.RiskCategory{risk.RiskCategoryFinancial, risk.RiskCategoryOperational, risk.RiskCategoryRegulatory}),
		Weight:      f.randomFloat(0.1, 1.0),
		Thresholds: map[risk.RiskLevel]float64{
			risk.RiskLevelLow:      0.0,
			risk.RiskLevelMedium:   25.0,
			risk.RiskLevelHigh:     50.0,
			risk.RiskLevelCritical: 75.0,
		},
		Metadata:  map[string]interface{}{"test": true},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// GenerateRiskScore creates a test risk score
func (f *TestDataFactory) GenerateRiskScore() risk.RiskScore {
	now := time.Now()

	return risk.RiskScore{
		FactorID:     f.generateID("factor"),
		FactorName:   f.randomChoice([]string{"Cash Flow", "Debt Ratio", "Operational Efficiency"}),
		Category:     f.randomChoice([]risk.RiskCategory{risk.RiskCategoryFinancial, risk.RiskCategoryOperational}),
		Score:        f.randomFloat(0.0, 100.0),
		Level:        f.randomChoice([]risk.RiskLevel{risk.RiskLevelLow, risk.RiskLevelMedium, risk.RiskLevelHigh}),
		Confidence:   f.randomFloat(0.5, 1.0),
		Explanation:  "Test risk score explanation",
		Evidence:     []string{"financial_statement", "payment_history"},
		CalculatedAt: now,
	}
}

// Helper methods for generating random data

func (f *TestDataFactory) generateID(prefix string) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(bytes))
}

func (f *TestDataFactory) generateEmail() string {
	domains := []string{"example.com", "test.com", "demo.org"}
	return fmt.Sprintf("user.%s@%s", f.generateID("user"), f.randomChoice(domains))
}

func (f *TestDataFactory) generateUsername() string {
	return fmt.Sprintf("user_%s", f.generateID("user"))
}

func (f *TestDataFactory) generateFirstName() string {
	names := []string{"John", "Jane", "Mike", "Sarah", "David", "Lisa", "Tom", "Emma"}
	return f.randomChoice(names)
}

func (f *TestDataFactory) generateLastName() string {
	names := []string{"Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis"}
	return f.randomChoice(names)
}

func (f *TestDataFactory) generateFullName() string {
	return fmt.Sprintf("%s %s", f.generateFirstName(), f.generateLastName())
}

func (f *TestDataFactory) generateCompanyName() string {
	companies := []string{"TechCorp", "FinancePro", "HealthCare Inc", "Retail Solutions", "Manufacturing Co"}
	return fmt.Sprintf("%s %s", f.randomChoice(companies), f.generateID("company"))
}

func (f *TestDataFactory) generateLegalName() string {
	return fmt.Sprintf("%s, LLC", f.generateCompanyName())
}

func (f *TestDataFactory) generateRegistrationNumber() string {
	return fmt.Sprintf("REG-%s", f.generateID("reg"))
}

func (f *TestDataFactory) generateTaxID() string {
	return fmt.Sprintf("TAX-%s", f.generateID("tax"))
}

func (f *TestDataFactory) generateIndustryCode() string {
	codes := []string{"541511", "522110", "621111", "441110", "332996"}
	return f.randomChoice(codes)
}

func (f *TestDataFactory) generateStreetAddress() string {
	streets := []string{"Main St", "Oak Ave", "Pine Rd", "Elm St", "Cedar Ln"}
	numbers := f.randomInt(100, 9999)
	return fmt.Sprintf("%d %s", numbers, f.randomChoice(streets))
}

func (f *TestDataFactory) generatePostalCode() string {
	return fmt.Sprintf("%05d", f.randomInt(10000, 99999))
}

func (f *TestDataFactory) generatePhoneNumber() string {
	return fmt.Sprintf("+1-%03d-%03d-%04d", f.randomInt(200, 999), f.randomInt(200, 999), f.randomInt(1000, 9999))
}

func (f *TestDataFactory) generateWebsite() string {
	return fmt.Sprintf("https://www.%s.com", f.generateID("site"))
}

func (f *TestDataFactory) generateAPIKeyName() string {
	return fmt.Sprintf("API Key %s", f.generateID("key"))
}

func (f *TestDataFactory) generateKeyHash() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (f *TestDataFactory) generateIPAddress() string {
	return fmt.Sprintf("192.168.%d.%d", f.randomInt(1, 255), f.randomInt(1, 255))
}

func (f *TestDataFactory) generateUserAgent() string {
	agents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
	}
	return f.randomChoice(agents)
}

func (f *TestDataFactory) randomChoice(choices []string) string {
	if len(choices) == 0 {
		return ""
	}
	return choices[f.randomInt(0, len(choices)-1)]
}

func (f *TestDataFactory) randomInt(min, max int) int {
	delta := max - min + 1
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(delta)))
	return min + int(n.Int64())
}

func (f *TestDataFactory) randomFloat(min, max float64) float64 {
	n, _ := rand.Int(rand.Reader, big.NewInt(1000000))
	return min + (max-min)*float64(n.Int64())/1000000.0
}
