package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// CompaniesHouseProvider implements Companies House API integration for UK companies
// This is a FREE government API that provides access to UK company registration data
type CompaniesHouseProvider struct {
	config ProviderConfig
	client *http.Client
	health bool
}

// CompaniesHouseCompany represents a company from Companies House API
type CompaniesHouseCompany struct {
	CompanyNumber           string `json:"company_number"`
	CompanyName             string `json:"title"`
	CompanyStatus           string `json:"company_status"`
	CompanyType             string `json:"company_type"`
	DateOfCreation          string `json:"date_of_creation"`
	RegisteredOfficeAddress struct {
		AddressLine1 string `json:"address_line_1"`
		AddressLine2 string `json:"address_line_2"`
		Locality     string `json:"locality"`
		PostalCode   string `json:"postal_code"`
		Country      string `json:"country"`
	} `json:"registered_office_address"`
	SICCodes []struct {
		SICCode        string `json:"sic_code"`
		SICDescription string `json:"sic_description"`
	} `json:"sic_codes"`
}

// CompaniesHouseResponse represents the response from Companies House API
type CompaniesHouseResponse struct {
	Items []CompaniesHouseCompany `json:"items"`
	Total int                     `json:"total_results"`
}

// NewCompaniesHouseProvider creates a new Companies House provider
func NewCompaniesHouseProvider(config ProviderConfig) *CompaniesHouseProvider {
	// Companies House API is free but requires API key
	// Rate limit: 600 requests per 5 minutes
	if config.RateLimit == 0 {
		config.RateLimit = 120 // 600 requests per 5 minutes = 120 per minute
	}
	if config.BurstLimit == 0 {
		config.BurstLimit = 10
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}

	// Companies House is completely free
	config.CostPerRequest = 0.0
	config.CostPerSearch = 0.0
	config.CostPerDetail = 0.0
	config.CostPerFinancial = 0.0

	// Set base URL for Companies House API
	if config.BaseURL == "" {
		config.BaseURL = "https://api.company-information.service.gov.uk"
	}

	// Set provider type
	config.Type = "companies_house"

	return &CompaniesHouseProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		health: true,
	}
}

func (c *CompaniesHouseProvider) GetName() string {
	return c.config.Name
}

func (c *CompaniesHouseProvider) GetType() string {
	return c.config.Type
}

func (c *CompaniesHouseProvider) GetConfig() ProviderConfig {
	return c.config
}

func (c *CompaniesHouseProvider) IsHealthy() bool {
	return c.health
}

func (c *CompaniesHouseProvider) GetCost() float64 {
	return c.config.CostPerRequest // Always 0.0 for Companies House
}

func (c *CompaniesHouseProvider) GetQuota() QuotaInfo {
	// Companies House has rate limits but no quota limits
	return QuotaInfo{
		DailyUsed:    0,
		DailyLimit:   999999, // Effectively unlimited
		MonthlyUsed:  0,
		MonthlyLimit: 999999, // Effectively unlimited
		ResetTime:    time.Now().Add(24 * time.Hour),
		Remaining:    999999,
	}
}

func (c *CompaniesHouseProvider) SearchBusiness(ctx context.Context, query BusinessSearchQuery) (*BusinessData, error) {
	// Companies House API endpoint for company search
	// Documentation: https://developer.company-information.service.gov.uk/
	searchURL := fmt.Sprintf("%s/search/companies", c.config.BaseURL)

	// Build query parameters
	params := url.Values{}
	if query.CompanyName != "" {
		params.Add("q", query.CompanyName)
	}
	if query.City != "" {
		params.Add("location", query.City)
	}

	searchURL += "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Companies House requires API key authentication
	if c.config.APIKey == "" {
		return nil, fmt.Errorf("Companies House API key is required")
	}

	req.SetBasicAuth(c.config.APIKey, "")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Companies House API returned status %d", resp.StatusCode)
	}

	var chResponse CompaniesHouseResponse
	if err := json.NewDecoder(resp.Body).Decode(&chResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Companies House response to our BusinessData format
	if len(chResponse.Items) == 0 {
		return nil, fmt.Errorf("no companies found for query")
	}

	company := chResponse.Items[0] // Take the first result

	businessData := &BusinessData{
		ID:             fmt.Sprintf("ch_%s", company.CompanyNumber),
		ProviderID:     company.CompanyNumber,
		ProviderName:   c.config.Name,
		CompanyName:    company.CompanyName,
		LegalName:      company.CompanyName,
		BusinessNumber: company.CompanyNumber,
		Address: Address{
			Street1:    company.RegisteredOfficeAddress.AddressLine1,
			Street2:    company.RegisteredOfficeAddress.AddressLine2,
			City:       company.RegisteredOfficeAddress.Locality,
			PostalCode: company.RegisteredOfficeAddress.PostalCode,
			Country:    "GB", // Companies House is UK-only
		},
		Status:      c.mapCompanyStatus(company.CompanyStatus),
		LastUpdated: time.Now(),
		DataQuality: 0.95, // Government data is high quality
		Confidence:  0.90, // High confidence for government data
		DataSources: []DataSource{
			{
				Name:        "Companies House",
				Type:        "government",
				TrustLevel:  "high",
				LastUpdated: time.Now(),
			},
		},
	}

	// Add SIC codes if available
	for _, sic := range company.SICCodes {
		businessData.IndustryCodes = append(businessData.IndustryCodes, IndustryCode{
			Type:        "SIC",
			Code:        sic.SICCode,
			Description: sic.SICDescription,
		})
	}

	// Parse creation date
	if company.DateOfCreation != "" {
		if creationDate, err := time.Parse("2006-01-02", company.DateOfCreation); err == nil {
			businessData.FoundedDate = &creationDate
		}
	}

	return businessData, nil
}

func (c *CompaniesHouseProvider) GetBusinessDetails(ctx context.Context, businessID string) (*BusinessData, error) {
	// Extract company number from business ID
	companyNumber := strings.TrimPrefix(businessID, "ch_")

	// Get detailed company information from Companies House
	detailsURL := fmt.Sprintf("%s/company/%s", c.config.BaseURL, companyNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", detailsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.config.APIKey, "")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Companies House API returned status %d", resp.StatusCode)
	}

	var company CompaniesHouseCompany
	if err := json.NewDecoder(resp.Body).Decode(&company); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	businessData := &BusinessData{
		ID:             fmt.Sprintf("ch_%s", company.CompanyNumber),
		ProviderID:     company.CompanyNumber,
		ProviderName:   c.config.Name,
		CompanyName:    company.CompanyName,
		LegalName:      company.CompanyName,
		BusinessNumber: company.CompanyNumber,
		Address: Address{
			Street1:    company.RegisteredOfficeAddress.AddressLine1,
			Street2:    company.RegisteredOfficeAddress.AddressLine2,
			City:       company.RegisteredOfficeAddress.Locality,
			PostalCode: company.RegisteredOfficeAddress.PostalCode,
			Country:    "GB",
		},
		Status:      c.mapCompanyStatus(company.CompanyStatus),
		LastUpdated: time.Now(),
		DataQuality: 0.95,
		Confidence:  0.90,
		DataSources: []DataSource{
			{
				Name:        "Companies House",
				Type:        "government",
				TrustLevel:  "high",
				LastUpdated: time.Now(),
			},
		},
	}

	// Add SIC codes
	for _, sic := range company.SICCodes {
		businessData.IndustryCodes = append(businessData.IndustryCodes, IndustryCode{
			Type:        "SIC",
			Code:        sic.SICCode,
			Description: sic.SICDescription,
		})
	}

	// Parse creation date
	if company.DateOfCreation != "" {
		if creationDate, err := time.Parse("2006-01-02", company.DateOfCreation); err == nil {
			businessData.FoundedDate = &creationDate
		}
	}

	return businessData, nil
}

func (c *CompaniesHouseProvider) GetFinancialData(ctx context.Context, businessID string) (*FinancialData, error) {
	// Companies House provides basic financial data through filings
	// This is a simplified implementation - in production, you'd parse filing data
	return &FinancialData{
		ProviderID:   businessID,
		ProviderName: c.config.Name,
		LastUpdated:  time.Now(),
		DataQuality:  0.90,
		Confidence:   0.85,
		DataSources: []DataSource{
			{
				Name:        "Companies House Filings",
				Type:        "government",
				TrustLevel:  "high",
				LastUpdated: time.Now(),
			},
		},
	}, nil
}

func (c *CompaniesHouseProvider) GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error) {
	// Companies House provides compliance data through company status and filings
	return &ComplianceData{
		ProviderID:   businessID,
		ProviderName: c.config.Name,
		LastUpdated:  time.Now(),
		DataQuality:  0.95,
		Confidence:   0.90,
		DataSources: []DataSource{
			{
				Name:        "Companies House Status",
				Type:        "government",
				TrustLevel:  "high",
				LastUpdated: time.Now(),
			},
		},
	}, nil
}

func (c *CompaniesHouseProvider) GetNewsData(ctx context.Context, businessID string) ([]NewsItem, error) {
	// Companies House doesn't provide news data
	return []NewsItem{}, nil
}

func (c *CompaniesHouseProvider) ValidateData(data *BusinessData) (*DataValidationResult, error) {
	// Companies House data is government-verified, so it's generally high quality
	issues := []ValidationIssue{}

	// Check for required fields
	if data.CompanyName == "" {
		issues = append(issues, ValidationIssue{
			Field:       "company_name",
			Type:        "missing",
			Severity:    "high",
			Description: "Company name is required",
		})
	}

	if data.ProviderID == "" {
		issues = append(issues, ValidationIssue{
			Field:       "provider_id",
			Type:        "missing",
			Severity:    "high",
			Description: "Provider ID (company number) is required",
		})
	}

	// Calculate quality score
	qualityScore := 1.0
	if len(issues) > 0 {
		qualityScore = 0.8 // Still high because it's government data
	}

	return &DataValidationResult{
		IsValid:       len(issues) == 0,
		QualityScore:  qualityScore,
		Issues:        issues,
		LastValidated: time.Now(),
	}, nil
}

// mapCompanyStatus maps Companies House company status to our internal status
func (c *CompaniesHouseProvider) mapCompanyStatus(chStatus string) string {
	switch strings.ToLower(chStatus) {
	case "active":
		return "active"
	case "dissolved":
		return "dissolved"
	case "liquidation":
		return "liquidation"
	case "receivership":
		return "receivership"
	case "administration":
		return "administration"
	case "voluntary-arrangement":
		return "voluntary_arrangement"
	case "converted-closed":
		return "converted_closed"
	case "insolvency-proceedings":
		return "insolvency_proceedings"
	default:
		return "unknown"
	}
}
