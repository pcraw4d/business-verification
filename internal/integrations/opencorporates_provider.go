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

// OpenCorporatesProvider implements OpenCorporates API integration
// This is a FREE API with limited free tier that provides global company data
type OpenCorporatesProvider struct {
	config ProviderConfig
	client *http.Client
	health bool
}

// OpenCorporatesCompany represents a company from OpenCorporates API
type OpenCorporatesCompany struct {
	Company struct {
		CompanyNumber     string `json:"company_number"`
		Name              string `json:"name"`
		JurisdictionCode  string `json:"jurisdiction_code"`
		CompanyType       string `json:"company_type"`
		Status            string `json:"current_status"`
		IncorporationDate string `json:"incorporation_date"`
		DissolutionDate   string `json:"dissolution_date"`
		RegisteredAddress struct {
			StreetAddress string `json:"street_address"`
			Locality      string `json:"locality"`
			Region        string `json:"region"`
			PostalCode    string `json:"postal_code"`
			Country       string `json:"country"`
		} `json:"registered_address"`
		IndustryCodes []struct {
			Code        string `json:"code"`
			Description string `json:"description"`
			CodeType    string `json:"code_type"`
		} `json:"industry_codes"`
	} `json:"company"`
}

// OpenCorporatesResponse represents the response from OpenCorporates API
type OpenCorporatesResponse struct {
	Results struct {
		Companies []OpenCorporatesCompany `json:"companies"`
		Total     int                     `json:"total_count"`
	} `json:"results"`
}

// NewOpenCorporatesProvider creates a new OpenCorporates provider
func NewOpenCorporatesProvider(config ProviderConfig) *OpenCorporatesProvider {
	// OpenCorporates API is free with rate limits
	// Free tier: 500 requests per day
	if config.RateLimit == 0 {
		config.RateLimit = 500 // 500 requests per day = ~20 per hour = ~0.3 per minute
	}
	if config.BurstLimit == 0 {
		config.BurstLimit = 5
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

	// OpenCorporates is free (with limits)
	config.CostPerRequest = 0.0
	config.CostPerSearch = 0.0
	config.CostPerDetail = 0.0
	config.CostPerFinancial = 0.0

	// Set base URL for OpenCorporates API
	if config.BaseURL == "" {
		config.BaseURL = "https://api.opencorporates.com"
	}

	// Set provider type
	config.Type = "opencorporates"

	return &OpenCorporatesProvider{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		health: true,
	}
}

func (o *OpenCorporatesProvider) GetName() string {
	return o.config.Name
}

func (o *OpenCorporatesProvider) GetType() string {
	return o.config.Type
}

func (o *OpenCorporatesProvider) GetConfig() ProviderConfig {
	return o.config
}

func (o *OpenCorporatesProvider) IsHealthy() bool {
	return o.health
}

func (o *OpenCorporatesProvider) GetCost() float64 {
	return o.config.CostPerRequest // Always 0.0 for OpenCorporates free tier
}

func (o *OpenCorporatesProvider) GetQuota() QuotaInfo {
	// OpenCorporates free tier: 500 requests per day
	return QuotaInfo{
		DailyUsed:    0,
		DailyLimit:   500,
		MonthlyUsed:  0,
		MonthlyLimit: 15000, // 500 * 30 days
		ResetTime:    time.Now().Add(24 * time.Hour),
		Remaining:    500,
	}
}

func (o *OpenCorporatesProvider) SearchBusiness(ctx context.Context, query BusinessSearchQuery) (*BusinessData, error) {
	// OpenCorporates API endpoint for company search
	// Documentation: https://api.opencorporates.com/documentation/API-Reference
	searchURL := fmt.Sprintf("%s/v0.4/companies/search", o.config.BaseURL)

	// Build query parameters
	params := url.Values{}
	if query.CompanyName != "" {
		params.Add("q", query.CompanyName)
	}
	if query.Country != "" {
		params.Add("jurisdiction_code", o.mapCountryToJurisdiction(query.Country))
	}

	// Add API token if available (for higher rate limits)
	if o.config.APIKey != "" {
		params.Add("api_token", o.config.APIKey)
	}

	searchURL += "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "KYB-Platform/1.0")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenCorporates API returned status %d", resp.StatusCode)
	}

	var ocResponse OpenCorporatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&ocResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert OpenCorporates response to our BusinessData format
	if len(ocResponse.Results.Companies) == 0 {
		return nil, fmt.Errorf("no companies found for query")
	}

	company := ocResponse.Results.Companies[0] // Take the first result

	businessData := &BusinessData{
		ID:             fmt.Sprintf("oc_%s_%s", company.Company.JurisdictionCode, company.Company.CompanyNumber),
		ProviderID:     company.Company.CompanyNumber,
		ProviderName:   o.config.Name,
		CompanyName:    company.Company.Name,
		LegalName:      company.Company.Name,
		BusinessNumber: company.Company.CompanyNumber,
		Address: Address{
			Street1:    company.Company.RegisteredAddress.StreetAddress,
			City:       company.Company.RegisteredAddress.Locality,
			State:      company.Company.RegisteredAddress.Region,
			PostalCode: company.Company.RegisteredAddress.PostalCode,
			Country:    o.mapJurisdictionToCountry(company.Company.JurisdictionCode),
		},
		Status:      o.mapCompanyStatus(company.Company.Status),
		LastUpdated: time.Now(),
		DataQuality: 0.85, // Good quality but not government-verified
		Confidence:  0.80, // Good confidence for global data
		DataSources: []DataSource{
			{
				Name:        "OpenCorporates",
				Type:        "commercial",
				TrustLevel:  "medium",
				LastUpdated: time.Now(),
			},
		},
	}

	// Add industry codes if available
	for _, code := range company.Company.IndustryCodes {
		businessData.IndustryCodes = append(businessData.IndustryCodes, IndustryCode{
			Type:        code.CodeType,
			Code:        code.Code,
			Description: code.Description,
		})
	}

	// Parse incorporation date
	if company.Company.IncorporationDate != "" {
		if incorporationDate, err := time.Parse("2006-01-02", company.Company.IncorporationDate); err == nil {
			businessData.FoundedDate = &incorporationDate
		}
	}

	return businessData, nil
}

func (o *OpenCorporatesProvider) GetBusinessDetails(ctx context.Context, businessID string) (*BusinessData, error) {
	// Extract jurisdiction and company number from business ID
	parts := strings.Split(strings.TrimPrefix(businessID, "oc_"), "_")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid business ID format")
	}

	jurisdictionCode := parts[0]
	companyNumber := parts[1]

	// Get detailed company information from OpenCorporates
	detailsURL := fmt.Sprintf("%s/v0.4/companies/%s/%s", o.config.BaseURL, jurisdictionCode, companyNumber)

	// Add API token if available
	if o.config.APIKey != "" {
		detailsURL += "?api_token=" + o.config.APIKey
	}

	req, err := http.NewRequestWithContext(ctx, "GET", detailsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "KYB-Platform/1.0")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenCorporates API returned status %d", resp.StatusCode)
	}

	var company OpenCorporatesCompany
	if err := json.NewDecoder(resp.Body).Decode(&company); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	businessData := &BusinessData{
		ID:             fmt.Sprintf("oc_%s_%s", company.Company.JurisdictionCode, company.Company.CompanyNumber),
		ProviderID:     company.Company.CompanyNumber,
		ProviderName:   o.config.Name,
		CompanyName:    company.Company.Name,
		LegalName:      company.Company.Name,
		BusinessNumber: company.Company.CompanyNumber,
		Address: Address{
			Street1:    company.Company.RegisteredAddress.StreetAddress,
			City:       company.Company.RegisteredAddress.Locality,
			State:      company.Company.RegisteredAddress.Region,
			PostalCode: company.Company.RegisteredAddress.PostalCode,
			Country:    o.mapJurisdictionToCountry(company.Company.JurisdictionCode),
		},
		Status:      o.mapCompanyStatus(company.Company.Status),
		LastUpdated: time.Now(),
		DataQuality: 0.85,
		Confidence:  0.80,
		DataSources: []DataSource{
			{
				Name:        "OpenCorporates",
				Type:        "commercial",
				TrustLevel:  "medium",
				LastUpdated: time.Now(),
			},
		},
	}

	// Add industry codes
	for _, code := range company.Company.IndustryCodes {
		businessData.IndustryCodes = append(businessData.IndustryCodes, IndustryCode{
			Type:        code.CodeType,
			Code:        code.Code,
			Description: code.Description,
		})
	}

	// Parse incorporation date
	if company.Company.IncorporationDate != "" {
		if incorporationDate, err := time.Parse("2006-01-02", company.Company.IncorporationDate); err == nil {
			businessData.FoundedDate = &incorporationDate
		}
	}

	return businessData, nil
}

func (o *OpenCorporatesProvider) GetFinancialData(ctx context.Context, businessID string) (*FinancialData, error) {
	// OpenCorporates provides limited financial data
	return &FinancialData{
		ProviderID:   businessID,
		ProviderName: o.config.Name,
		LastUpdated:  time.Now(),
		DataQuality:  0.70,
		Confidence:   0.75,
		DataSources: []DataSource{
			{
				Name:        "OpenCorporates",
				Type:        "commercial",
				TrustLevel:  "medium",
				LastUpdated: time.Now(),
			},
		},
	}, nil
}

func (o *OpenCorporatesProvider) GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error) {
	// OpenCorporates provides basic compliance data through company status
	return &ComplianceData{
		ProviderID:   businessID,
		ProviderName: o.config.Name,
		LastUpdated:  time.Now(),
		DataQuality:  0.80,
		Confidence:   0.75,
		DataSources: []DataSource{
			{
				Name:        "OpenCorporates",
				Type:        "commercial",
				TrustLevel:  "medium",
				LastUpdated: time.Now(),
			},
		},
	}, nil
}

func (o *OpenCorporatesProvider) GetNewsData(ctx context.Context, businessID string) ([]NewsItem, error) {
	// OpenCorporates doesn't provide news data
	return []NewsItem{}, nil
}

func (o *OpenCorporatesProvider) ValidateData(data *BusinessData) (*DataValidationResult, error) {
	// OpenCorporates data is commercial but generally reliable
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
	qualityScore := 0.85 // Good but not government-verified
	if len(issues) > 0 {
		qualityScore = 0.70
	}

	return &DataValidationResult{
		IsValid:       len(issues) == 0,
		QualityScore:  qualityScore,
		Issues:        issues,
		LastValidated: time.Now(),
	}, nil
}

// mapCountryToJurisdiction maps country codes to OpenCorporates jurisdiction codes
func (o *OpenCorporatesProvider) mapCountryToJurisdiction(country string) string {
	switch strings.ToUpper(country) {
	case "US", "USA", "UNITED STATES":
		return "us"
	case "GB", "UK", "UNITED KINGDOM":
		return "gb"
	case "CA", "CANADA":
		return "ca"
	case "AU", "AUSTRALIA":
		return "au"
	case "DE", "GERMANY":
		return "de"
	case "FR", "FRANCE":
		return "fr"
	case "IT", "ITALY":
		return "it"
	case "ES", "SPAIN":
		return "es"
	case "NL", "NETHERLANDS":
		return "nl"
	case "BE", "BELGIUM":
		return "be"
	default:
		return strings.ToLower(country)
	}
}

// mapJurisdictionToCountry maps OpenCorporates jurisdiction codes to country codes
func (o *OpenCorporatesProvider) mapJurisdictionToCountry(jurisdiction string) string {
	switch strings.ToLower(jurisdiction) {
	case "us":
		return "US"
	case "gb":
		return "GB"
	case "ca":
		return "CA"
	case "au":
		return "AU"
	case "de":
		return "DE"
	case "fr":
		return "FR"
	case "it":
		return "IT"
	case "es":
		return "ES"
	case "nl":
		return "NL"
	case "be":
		return "BE"
	default:
		return strings.ToUpper(jurisdiction)
	}
}

// mapCompanyStatus maps OpenCorporates company status to our internal status
func (o *OpenCorporatesProvider) mapCompanyStatus(ocStatus string) string {
	switch strings.ToLower(ocStatus) {
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
