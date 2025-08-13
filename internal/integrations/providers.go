package integrations

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// DunBradstreetProvider implements Dun & Bradstreet API integration
type DunBradstreetProvider struct {
	config ProviderConfig
	client *http.Client
	health bool
}

// NewDunBradstreetProvider creates a new Dun & Bradstreet provider
func NewDunBradstreetProvider(config ProviderConfig) *DunBradstreetProvider {
	return &DunBradstreetProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		health: true,
	}
}

func (d *DunBradstreetProvider) GetName() string {
	return d.config.Name
}

func (d *DunBradstreetProvider) GetType() string {
	return d.config.Type
}

func (d *DunBradstreetProvider) GetConfig() ProviderConfig {
	return d.config
}

func (d *DunBradstreetProvider) IsHealthy() bool {
	return d.health
}

func (d *DunBradstreetProvider) GetCost() float64 {
	return d.config.CostPerRequest
}

func (d *DunBradstreetProvider) GetQuota() QuotaInfo {
	// Mock quota info - in real implementation would fetch from API
	return QuotaInfo{
		DailyUsed:    50,
		DailyLimit:   1000,
		MonthlyUsed:  1500,
		MonthlyLimit: 30000,
		ResetTime:    time.Now().Add(24 * time.Hour),
		Remaining:    950,
	}
}

func (d *DunBradstreetProvider) SearchBusiness(ctx context.Context, query BusinessSearchQuery) (*BusinessData, error) {
	// Mock implementation - in real implementation would call D&B API
	return &BusinessData{
		ID:             "dnb_12345",
		ProviderID:     "12345",
		ProviderName:   d.config.Name,
		CompanyName:    query.CompanyName,
		LegalName:      query.CompanyName + " Inc.",
		BusinessNumber: "123456789",
		DUNSNumber:     "123456789",
		Address: Address{
			Street1: "123 Business St",
			City:    query.City,
			State:   query.State,
			Country: query.Country,
		},
		Industry:      query.Industry,
		EmployeeCount: 100,
		Revenue:       1000000.0,
		Status:        "active",
		LastUpdated:   time.Now(),
		DataQuality:   0.95,
		Confidence:    0.92,
	}, nil
}

func (d *DunBradstreetProvider) GetBusinessDetails(ctx context.Context, businessID string) (*BusinessData, error) {
	// Mock implementation
	return &BusinessData{
		ID:             businessID,
		ProviderID:     businessID,
		ProviderName:   d.config.Name,
		CompanyName:    "Sample Company",
		LegalName:      "Sample Company Inc.",
		BusinessNumber: "123456789",
		DUNSNumber:     "123456789",
		Address: Address{
			Street1: "123 Business St",
			City:    "New York",
			State:   "NY",
			Country: "US",
		},
		Industry:      "Technology",
		EmployeeCount: 100,
		Revenue:       1000000.0,
		Status:        "active",
		LastUpdated:   time.Now(),
		DataQuality:   0.95,
		Confidence:    0.92,
	}, nil
}

func (d *DunBradstreetProvider) GetFinancialData(ctx context.Context, businessID string) (*FinancialData, error) {
	// Mock implementation
	return &FinancialData{
		FiscalYear:       2023,
		Revenue:          1000000.0,
		NetIncome:        150000.0,
		TotalAssets:      2000000.0,
		TotalLiabilities: 800000.0,
		Equity:           1200000.0,
		CashFlow:         200000.0,
		EBITDA:           250000.0,
		DebtToEquity:     0.67,
		CurrentRatio:     1.5,
		QuickRatio:       1.2,
		ROE:              0.125,
		ROA:              0.075,
		GrossMargin:      0.35,
		NetMargin:        0.15,
		Currency:         "USD",
		LastUpdated:      time.Now(),
	}, nil
}

func (d *DunBradstreetProvider) GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error) {
	// Mock implementation
	return &ComplianceData{
		RegulatoryStatus: "compliant",
		LicenseNumbers:   []string{"LIC123", "LIC456"},
		Certifications:   []string{"ISO9001", "ISO27001"},
		Violations:       []Violation{},
		AuditReports:     []AuditReport{},
		ComplianceScore:  0.95,
		RiskLevel:        "low",
		LastUpdated:      time.Now(),
	}, nil
}

func (d *DunBradstreetProvider) GetNewsData(ctx context.Context, businessID string) ([]NewsItem, error) {
	// Mock implementation
	return []NewsItem{
		{
			Title:         "Sample Company Reports Strong Q4 Results",
			Summary:       "Sample Company announced strong fourth quarter results...",
			URL:           "https://example.com/news/1",
			Source:        "Business News",
			Author:        "John Doe",
			PublishedDate: time.Now().Add(-24 * time.Hour),
			Sentiment:     "positive",
			Relevance:     0.9,
			Tags:          []string{"earnings", "financial"},
		},
	}, nil
}

func (d *DunBradstreetProvider) ValidateData(data *BusinessData) (*DataValidationResult, error) {
	// Mock validation
	issues := []ValidationIssue{}

	if data.CompanyName == "" {
		issues = append(issues, ValidationIssue{
			Field:       "company_name",
			Type:        "missing",
			Severity:    "high",
			Description: "Company name is required",
		})
	}

	qualityScore := 0.95
	if len(issues) > 0 {
		qualityScore = 0.8
	}

	return &DataValidationResult{
		IsValid:      len(issues) == 0,
		QualityScore: qualityScore,
		Issues:       issues,
		Confidence:   0.9,
	}, nil
}

// ExperianProvider implements Experian API integration
type ExperianProvider struct {
	config ProviderConfig
	client *http.Client
	health bool
}

// NewExperianProvider creates a new Experian provider
func NewExperianProvider(config ProviderConfig) *ExperianProvider {
	return &ExperianProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		health: true,
	}
}

func (e *ExperianProvider) GetName() string {
	return e.config.Name
}

func (e *ExperianProvider) GetType() string {
	return e.config.Type
}

func (e *ExperianProvider) GetConfig() ProviderConfig {
	return e.config
}

func (e *ExperianProvider) IsHealthy() bool {
	return e.health
}

func (e *ExperianProvider) GetCost() float64 {
	return e.config.CostPerRequest
}

func (e *ExperianProvider) GetQuota() QuotaInfo {
	return QuotaInfo{
		DailyUsed:    75,
		DailyLimit:   1200,
		MonthlyUsed:  2000,
		MonthlyLimit: 35000,
		ResetTime:    time.Now().Add(24 * time.Hour),
		Remaining:    1125,
	}
}

func (e *ExperianProvider) SearchBusiness(ctx context.Context, query BusinessSearchQuery) (*BusinessData, error) {
	// Mock implementation
	return &BusinessData{
		ID:             "experian_67890",
		ProviderID:     "67890",
		ProviderName:   e.config.Name,
		CompanyName:    query.CompanyName,
		LegalName:      query.CompanyName + " LLC",
		BusinessNumber: "987654321",
		Address: Address{
			Street1: "456 Corporate Ave",
			City:    query.City,
			State:   query.State,
			Country: query.Country,
		},
		Industry:      query.Industry,
		EmployeeCount: 150,
		Revenue:       1500000.0,
		Status:        "active",
		LastUpdated:   time.Now(),
		DataQuality:   0.93,
		Confidence:    0.89,
	}, nil
}

func (e *ExperianProvider) GetBusinessDetails(ctx context.Context, businessID string) (*BusinessData, error) {
	// Mock implementation
	return &BusinessData{
		ID:             businessID,
		ProviderID:     businessID,
		ProviderName:   e.config.Name,
		CompanyName:    "Sample Company",
		LegalName:      "Sample Company LLC",
		BusinessNumber: "987654321",
		Address: Address{
			Street1: "456 Corporate Ave",
			City:    "Los Angeles",
			State:   "CA",
			Country: "US",
		},
		Industry:      "Finance",
		EmployeeCount: 150,
		Revenue:       1500000.0,
		Status:        "active",
		LastUpdated:   time.Now(),
		DataQuality:   0.93,
		Confidence:    0.89,
	}, nil
}

func (e *ExperianProvider) GetFinancialData(ctx context.Context, businessID string) (*FinancialData, error) {
	// Mock implementation
	return &FinancialData{
		FiscalYear:       2023,
		Revenue:          1500000.0,
		NetIncome:        225000.0,
		TotalAssets:      3000000.0,
		TotalLiabilities: 1200000.0,
		Equity:           1800000.0,
		CashFlow:         300000.0,
		EBITDA:           375000.0,
		DebtToEquity:     0.67,
		CurrentRatio:     1.8,
		QuickRatio:       1.4,
		ROE:              0.125,
		ROA:              0.075,
		GrossMargin:      0.40,
		NetMargin:        0.15,
		Currency:         "USD",
		LastUpdated:      time.Now(),
	}, nil
}

func (e *ExperianProvider) GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error) {
	// Mock implementation
	return &ComplianceData{
		RegulatoryStatus: "compliant",
		LicenseNumbers:   []string{"FIN123", "SEC456"},
		Certifications:   []string{"SOC2", "PCI-DSS"},
		Violations:       []Violation{},
		AuditReports:     []AuditReport{},
		ComplianceScore:  0.92,
		RiskLevel:        "low",
		LastUpdated:      time.Now(),
	}, nil
}

func (e *ExperianProvider) GetNewsData(ctx context.Context, businessID string) ([]NewsItem, error) {
	// Mock implementation
	return []NewsItem{
		{
			Title:         "Sample Company Expands Operations",
			Summary:       "Sample Company announced expansion plans...",
			URL:           "https://example.com/news/2",
			Source:        "Financial Times",
			Author:        "Jane Smith",
			PublishedDate: time.Now().Add(-48 * time.Hour),
			Sentiment:     "positive",
			Relevance:     0.85,
			Tags:          []string{"expansion", "growth"},
		},
	}, nil
}

func (e *ExperianProvider) ValidateData(data *BusinessData) (*DataValidationResult, error) {
	// Mock validation
	issues := []ValidationIssue{}

	if data.BusinessNumber == "" {
		issues = append(issues, ValidationIssue{
			Field:       "business_number",
			Type:        "missing",
			Severity:    "medium",
			Description: "Business number is recommended",
		})
	}

	qualityScore := 0.93
	if len(issues) > 0 {
		qualityScore = 0.85
	}

	return &DataValidationResult{
		IsValid:      len(issues) == 0,
		QualityScore: qualityScore,
		Issues:       issues,
		Confidence:   0.87,
	}, nil
}

// SECProvider implements SEC API integration
type SECProvider struct {
	config ProviderConfig
	client *http.Client
	health bool
}

// NewSECProvider creates a new SEC provider
func NewSECProvider(config ProviderConfig) *SECProvider {
	return &SECProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		health: true,
	}
}

func (s *SECProvider) GetName() string {
	return s.config.Name
}

func (s *SECProvider) GetType() string {
	return s.config.Type
}

func (s *SECProvider) GetConfig() ProviderConfig {
	return s.config
}

func (s *SECProvider) IsHealthy() bool {
	return s.health
}

func (s *SECProvider) GetCost() float64 {
	return s.config.CostPerRequest
}

func (s *SECProvider) GetQuota() QuotaInfo {
	return QuotaInfo{
		DailyUsed:    200,
		DailyLimit:   5000,
		MonthlyUsed:  15000,
		MonthlyLimit: 100000,
		ResetTime:    time.Now().Add(24 * time.Hour),
		Remaining:    4800,
	}
}

func (s *SECProvider) SearchBusiness(ctx context.Context, query BusinessSearchQuery) (*BusinessData, error) {
	// Mock implementation for SEC data
	return &BusinessData{
		ID:             "sec_11111",
		ProviderID:     "11111",
		ProviderName:   s.config.Name,
		CompanyName:    query.CompanyName,
		LegalName:      query.CompanyName + " Corporation",
		BusinessNumber: "SEC123456",
		Address: Address{
			Street1: "789 Wall Street",
			City:    query.City,
			State:   query.State,
			Country: query.Country,
		},
		Industry:      query.Industry,
		EmployeeCount: 500,
		Revenue:       5000000.0,
		Status:        "active",
		LastUpdated:   time.Now(),
		DataQuality:   0.98,
		Confidence:    0.95,
	}, nil
}

func (s *SECProvider) GetBusinessDetails(ctx context.Context, businessID string) (*BusinessData, error) {
	// Mock implementation
	return &BusinessData{
		ID:             businessID,
		ProviderID:     businessID,
		ProviderName:   s.config.Name,
		CompanyName:    "Sample Company",
		LegalName:      "Sample Company Corporation",
		BusinessNumber: "SEC123456",
		Address: Address{
			Street1: "789 Wall Street",
			City:    "New York",
			State:   "NY",
			Country: "US",
		},
		Industry:      "Technology",
		EmployeeCount: 500,
		Revenue:       5000000.0,
		Status:        "active",
		LastUpdated:   time.Now(),
		DataQuality:   0.98,
		Confidence:    0.95,
	}, nil
}

func (s *SECProvider) GetFinancialData(ctx context.Context, businessID string) (*FinancialData, error) {
	// Mock implementation
	return &FinancialData{
		FiscalYear:       2023,
		Revenue:          5000000.0,
		NetIncome:        750000.0,
		TotalAssets:      10000000.0,
		TotalLiabilities: 4000000.0,
		Equity:           6000000.0,
		CashFlow:         1000000.0,
		EBITDA:           1250000.0,
		DebtToEquity:     0.67,
		CurrentRatio:     2.0,
		QuickRatio:       1.6,
		ROE:              0.125,
		ROA:              0.075,
		GrossMargin:      0.45,
		NetMargin:        0.15,
		Currency:         "USD",
		LastUpdated:      time.Now(),
	}, nil
}

func (s *SECProvider) GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error) {
	// Mock implementation
	return &ComplianceData{
		RegulatoryStatus: "compliant",
		LicenseNumbers:   []string{"SEC001", "FINRA002"},
		Certifications:   []string{"SOX", "Dodd-Frank"},
		Violations:       []Violation{},
		AuditReports:     []AuditReport{},
		ComplianceScore:  0.98,
		RiskLevel:        "low",
		LastUpdated:      time.Now(),
	}, nil
}

func (s *SECProvider) GetNewsData(ctx context.Context, businessID string) ([]NewsItem, error) {
	// Mock implementation
	return []NewsItem{
		{
			Title:         "Sample Company Files 10-K Report",
			Summary:       "Sample Company filed its annual 10-K report...",
			URL:           "https://example.com/news/3",
			Source:        "SEC Filings",
			Author:        "SEC",
			PublishedDate: time.Now().Add(-72 * time.Hour),
			Sentiment:     "neutral",
			Relevance:     0.95,
			Tags:          []string{"filing", "10-k", "annual"},
		},
	}, nil
}

func (s *SECProvider) ValidateData(data *BusinessData) (*DataValidationResult, error) {
	// Mock validation
	issues := []ValidationIssue{}

	qualityScore := 0.98
	if len(issues) > 0 {
		qualityScore = 0.90
	}

	return &DataValidationResult{
		IsValid:      len(issues) == 0,
		QualityScore: qualityScore,
		Issues:       issues,
		Confidence:   0.95,
	}, nil
}

// BloombergProvider implements Bloomberg API integration
type BloombergProvider struct {
	config ProviderConfig
	client *http.Client
	health bool
}

// NewBloombergProvider creates a new Bloomberg provider
func NewBloombergProvider(config ProviderConfig) *BloombergProvider {
	return &BloombergProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		health: true,
	}
}

func (b *BloombergProvider) GetName() string {
	return b.config.Name
}

func (b *BloombergProvider) GetType() string {
	return b.config.Type
}

func (b *BloombergProvider) GetConfig() ProviderConfig {
	return b.config
}

func (b *BloombergProvider) IsHealthy() bool {
	return b.health
}

func (b *BloombergProvider) GetCost() float64 {
	return b.config.CostPerRequest
}

func (b *BloombergProvider) GetQuota() QuotaInfo {
	return QuotaInfo{
		DailyUsed:    100,
		DailyLimit:   2000,
		MonthlyUsed:  8000,
		MonthlyLimit: 50000,
		ResetTime:    time.Now().Add(24 * time.Hour),
		Remaining:    1900,
	}
}

func (b *BloombergProvider) SearchBusiness(ctx context.Context, query BusinessSearchQuery) (*BusinessData, error) {
	// Mock implementation
	return &BusinessData{
		ID:             "bloomberg_22222",
		ProviderID:     "22222",
		ProviderName:   b.config.Name,
		CompanyName:    query.CompanyName,
		LegalName:      query.CompanyName + " Group",
		BusinessNumber: "BBG123456",
		Address: Address{
			Street1: "321 Bloomberg Way",
			City:    query.City,
			State:   query.State,
			Country: query.Country,
		},
		Industry:      query.Industry,
		EmployeeCount: 1000,
		Revenue:       10000000.0,
		Status:        "active",
		LastUpdated:   time.Now(),
		DataQuality:   0.96,
		Confidence:    0.94,
	}, nil
}

func (b *BloombergProvider) GetBusinessDetails(ctx context.Context, businessID string) (*BusinessData, error) {
	// Mock implementation
	return &BusinessData{
		ID:             businessID,
		ProviderID:     businessID,
		ProviderName:   b.config.Name,
		CompanyName:    "Sample Company",
		LegalName:      "Sample Company Group",
		BusinessNumber: "BBG123456",
		Address: Address{
			Street1: "321 Bloomberg Way",
			City:    "London",
			State:   "",
			Country: "UK",
		},
		Industry:      "Financial Services",
		EmployeeCount: 1000,
		Revenue:       10000000.0,
		Status:        "active",
		LastUpdated:   time.Now(),
		DataQuality:   0.96,
		Confidence:    0.94,
	}, nil
}

func (b *BloombergProvider) GetFinancialData(ctx context.Context, businessID string) (*FinancialData, error) {
	// Mock implementation
	return &FinancialData{
		FiscalYear:       2023,
		Revenue:          10000000.0,
		NetIncome:        1500000.0,
		TotalAssets:      20000000.0,
		TotalLiabilities: 8000000.0,
		Equity:           12000000.0,
		CashFlow:         2000000.0,
		EBITDA:           2500000.0,
		DebtToEquity:     0.67,
		CurrentRatio:     2.2,
		QuickRatio:       1.8,
		ROE:              0.125,
		ROA:              0.075,
		GrossMargin:      0.50,
		NetMargin:        0.15,
		Currency:         "GBP",
		LastUpdated:      time.Now(),
	}, nil
}

func (b *BloombergProvider) GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error) {
	// Mock implementation
	return &ComplianceData{
		RegulatoryStatus: "compliant",
		LicenseNumbers:   []string{"FCA001", "PRA002"},
		Certifications:   []string{"ISO27001", "PCI-DSS"},
		Violations:       []Violation{},
		AuditReports:     []AuditReport{},
		ComplianceScore:  0.96,
		RiskLevel:        "low",
		LastUpdated:      time.Now(),
	}, nil
}

func (b *BloombergProvider) GetNewsData(ctx context.Context, businessID string) ([]NewsItem, error) {
	// Mock implementation
	return []NewsItem{
		{
			Title:         "Sample Company Announces Merger",
			Summary:       "Sample Company announced a major merger...",
			URL:           "https://example.com/news/4",
			Source:        "Bloomberg News",
			Author:        "Financial Reporter",
			PublishedDate: time.Now().Add(-12 * time.Hour),
			Sentiment:     "positive",
			Relevance:     0.92,
			Tags:          []string{"merger", "acquisition"},
		},
	}, nil
}

func (b *BloombergProvider) ValidateData(data *BusinessData) (*DataValidationResult, error) {
	// Mock validation
	issues := []ValidationIssue{}

	qualityScore := 0.96
	if len(issues) > 0 {
		qualityScore = 0.88
	}

	return &DataValidationResult{
		IsValid:      len(issues) == 0,
		QualityScore: qualityScore,
		Issues:       issues,
		Confidence:   0.94,
	}, nil
}

// FactivaProvider implements Factiva news API integration
type FactivaProvider struct {
	config ProviderConfig
	client *http.Client
	health bool
}

// NewFactivaProvider creates a new Factiva provider
func NewFactivaProvider(config ProviderConfig) *FactivaProvider {
	return &FactivaProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		health: true,
	}
}

func (f *FactivaProvider) GetName() string {
	return f.config.Name
}

func (f *FactivaProvider) GetType() string {
	return f.config.Type
}

func (f *FactivaProvider) GetConfig() ProviderConfig {
	return f.config
}

func (f *FactivaProvider) IsHealthy() bool {
	return f.health
}

func (f *FactivaProvider) GetCost() float64 {
	return f.config.CostPerRequest
}

func (f *FactivaProvider) GetQuota() QuotaInfo {
	return QuotaInfo{
		DailyUsed:    150,
		DailyLimit:   3000,
		MonthlyUsed:  12000,
		MonthlyLimit: 80000,
		ResetTime:    time.Now().Add(24 * time.Hour),
		Remaining:    2850,
	}
}

func (f *FactivaProvider) SearchBusiness(ctx context.Context, query BusinessSearchQuery) (*BusinessData, error) {
	// Factiva is primarily for news, so return basic business info
	return &BusinessData{
		ID:           "factiva_33333",
		ProviderID:   "33333",
		ProviderName: f.config.Name,
		CompanyName:  query.CompanyName,
		LegalName:    query.CompanyName + " Limited",
		Address: Address{
			Street1: "555 News Street",
			City:    query.City,
			State:   query.State,
			Country: query.Country,
		},
		Industry:    query.Industry,
		Status:      "active",
		LastUpdated: time.Now(),
		DataQuality: 0.85,
		Confidence:  0.82,
	}, nil
}

func (f *FactivaProvider) GetBusinessDetails(ctx context.Context, businessID string) (*BusinessData, error) {
	// Mock implementation
	return &BusinessData{
		ID:           businessID,
		ProviderID:   businessID,
		ProviderName: f.config.Name,
		CompanyName:  "Sample Company",
		LegalName:    "Sample Company Limited",
		Address: Address{
			Street1: "555 News Street",
			City:    "Toronto",
			State:   "ON",
			Country: "CA",
		},
		Industry:    "Media",
		Status:      "active",
		LastUpdated: time.Now(),
		DataQuality: 0.85,
		Confidence:  0.82,
	}, nil
}

func (f *FactivaProvider) GetFinancialData(ctx context.Context, businessID string) (*FinancialData, error) {
	// Factiva doesn't provide financial data
	return nil, fmt.Errorf("financial data not available from Factiva")
}

func (f *FactivaProvider) GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error) {
	// Factiva doesn't provide compliance data
	return nil, fmt.Errorf("compliance data not available from Factiva")
}

func (f *FactivaProvider) GetNewsData(ctx context.Context, businessID string) ([]NewsItem, error) {
	// Mock implementation with rich news data
	return []NewsItem{
		{
			Title:         "Sample Company Reports Record Profits",
			Summary:       "Sample Company announced record-breaking profits...",
			Content:       "Full article content would be here...",
			URL:           "https://example.com/news/5",
			Source:        "Reuters",
			Author:        "Business Reporter",
			PublishedDate: time.Now().Add(-6 * time.Hour),
			Sentiment:     "positive",
			Relevance:     0.95,
			Tags:          []string{"earnings", "profits", "financial"},
		},
		{
			Title:         "Sample Company Expands to New Markets",
			Summary:       "Sample Company announced expansion into new markets...",
			Content:       "Full article content would be here...",
			URL:           "https://example.com/news/6",
			Source:        "Financial Times",
			Author:        "Market Analyst",
			PublishedDate: time.Now().Add(-18 * time.Hour),
			Sentiment:     "positive",
			Relevance:     0.88,
			Tags:          []string{"expansion", "markets", "growth"},
		},
	}, nil
}

func (f *FactivaProvider) ValidateData(data *BusinessData) (*DataValidationResult, error) {
	// Mock validation
	issues := []ValidationIssue{}

	qualityScore := 0.85
	if len(issues) > 0 {
		qualityScore = 0.75
	}

	return &DataValidationResult{
		IsValid:      len(issues) == 0,
		QualityScore: qualityScore,
		Issues:       issues,
		Confidence:   0.82,
	}, nil
}
