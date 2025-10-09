package external

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// OpenCorporatesClient provides integration with OpenCorporates for company data
type OpenCorporatesClient struct {
	*Client
}

// OpenCorporatesResponse represents the response from OpenCorporates API
type OpenCorporatesResponse struct {
	Results struct {
		Companies []Company `json:"companies"`
	} `json:"results"`
}

// Company represents a company from OpenCorporates
type Company struct {
	Company struct {
		Name                string    `json:"name"`
		CompanyNumber       string    `json:"company_number"`
		JurisdictionCode    string    `json:"jurisdiction_code"`
		CompanyType         string    `json:"company_type"`
		Status              string    `json:"status"`
		DateOfCreation      time.Time `json:"date_of_creation"`
		DateOfDissolution   *time.Time `json:"date_of_dissolution"`
		RegisteredAddress   Address   `json:"registered_address"`
		HomepageURL         string    `json:"homepage_url"`
		RegistryURL         string    `json:"registry_url"`
		Branch              string    `json:"branch"`
		BranchStatus        string    `json:"branch_status"`
		Inactive            bool      `json:"inactive"`
		CurrentOfficers     []Officer `json:"current_officers"`
		PreviousOfficers    []Officer `json:"previous_officers"`
		CurrentDirectors    []Officer `json:"current_directors"`
		PreviousDirectors   []Officer `json:"previous_directors"`
		IndustryCodes       []Code    `json:"industry_codes"`
		PreviousNames       []Name    `json:"previous_names"`
		Source              OpenCorporatesSource    `json:"source"`
		RetrievedAt         time.Time `json:"retrieved_at"`
		Data                Data      `json:"data"`
	} `json:"company"`
}

// Address represents a company address
type Address struct {
	StreetAddress string `json:"street_address"`
	Locality      string `json:"locality"`
	Region        string `json:"region"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

// Officer represents a company officer
type Officer struct {
	Name         string    `json:"name"`
	Position     string    `json:"position"`
	StartDate    time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	Nationality  string    `json:"nationality"`
	Occupation   string    `json:"occupation"`
	DateOfBirth  *time.Time `json:"date_of_birth"`
	Address      Address   `json:"address"`
}

// Code represents an industry or classification code
type Code struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	CodeType    string `json:"code_type"`
}

// Name represents a previous company name
type Name struct {
	Name      string    `json:"name"`
	StartDate time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

// OpenCorporatesSource represents the data source
type OpenCorporatesSource struct {
	Publisher string `json:"publisher"`
	URL       string `json:"url"`
	Terms     string `json:"terms"`
}

// Data represents additional company data
type Data struct {
	NumberOfEmployees *int    `json:"number_of_employees"`
	AnnualRevenue     *int    `json:"annual_revenue"`
	Description       string  `json:"description"`
	Website           string  `json:"website"`
	Phone             string  `json:"phone"`
	Email             string  `json:"email"`
}

// CompanySearchResult represents the result of a company search
type CompanySearchResult struct {
	BusinessName     string    `json:"business_name"`
	Companies        []Company `json:"companies"`
	TotalResults     int       `json:"total_results"`
	RiskScore        float64   `json:"risk_score"`
	ComplianceStatus string    `json:"compliance_status"`
	LastChecked      time.Time `json:"last_checked"`
}

// NewOpenCorporatesClient creates a new OpenCorporates client
func NewOpenCorporatesClient(apiKey string, logger *zap.Logger) *OpenCorporatesClient {
	config := Config{
		BaseURL:    "https://api.opencorporates.com/v0.4",
		APIKey:     apiKey,
		Timeout:    15 * time.Second,
		MaxRetries: 3,
	}

	return &OpenCorporatesClient{
		Client: NewClient(config, logger),
	}
}

// SearchCompany searches for a company by name
func (c *OpenCorporatesClient) SearchCompany(ctx context.Context, companyName, jurisdiction string) (*CompanySearchResult, error) {
	c.logger.Info("Searching for company",
		zap.String("company_name", companyName),
		zap.String("jurisdiction", jurisdiction))

	params := map[string]string{
		"q": companyName,
		"format": "json",
		"per_page": "20",
	}

	if jurisdiction != "" {
		params["jurisdiction_code"] = jurisdiction
	}

	resp, err := c.Get(ctx, "/companies/search", params)
	if err != nil {
		return nil, fmt.Errorf("failed to search company: %w", err)
	}
	defer resp.Body.Close()

	var searchResponse OpenCorporatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Calculate risk score and compliance status
	riskScore, complianceStatus := c.analyzeCompanyRisk(searchResponse.Results.Companies)

	result := &CompanySearchResult{
		BusinessName:     companyName,
		Companies:        searchResponse.Results.Companies,
		TotalResults:     len(searchResponse.Results.Companies),
		RiskScore:        riskScore,
		ComplianceStatus: complianceStatus,
		LastChecked:      time.Now(),
	}

	c.logger.Info("Company search completed",
		zap.String("company_name", companyName),
		zap.Int("companies_found", len(searchResponse.Results.Companies)),
		zap.Float64("risk_score", riskScore),
		zap.String("compliance_status", complianceStatus))

	return result, nil
}

// GetCompanyDetails gets detailed information about a specific company
func (c *OpenCorporatesClient) GetCompanyDetails(ctx context.Context, jurisdictionCode, companyNumber string) (*Company, error) {
	c.logger.Info("Getting company details",
		zap.String("jurisdiction_code", jurisdictionCode),
		zap.String("company_number", companyNumber))

	endpoint := fmt.Sprintf("/companies/%s/%s", jurisdictionCode, companyNumber)
	params := map[string]string{
		"format": "json",
	}

	resp, err := c.Get(ctx, endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to get company details: %w", err)
	}
	defer resp.Body.Close()

	var companyResponse struct {
		Results struct {
			Company Company `json:"company"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&companyResponse); err != nil {
		return nil, fmt.Errorf("failed to decode company response: %w", err)
	}

	c.logger.Info("Company details retrieved",
		zap.String("company_name", companyResponse.Results.Company.Company.Name),
		zap.String("status", companyResponse.Results.Company.Company.Status))

	return &companyResponse.Results.Company, nil
}

// SearchOfficers searches for officers of a company
func (c *OpenCorporatesClient) SearchOfficers(ctx context.Context, jurisdictionCode, companyNumber string) ([]Officer, error) {
	c.logger.Info("Searching for company officers",
		zap.String("jurisdiction_code", jurisdictionCode),
		zap.String("company_number", companyNumber))

	endpoint := fmt.Sprintf("/companies/%s/%s/officers", jurisdictionCode, companyNumber)
	params := map[string]string{
		"format": "json",
		"per_page": "50",
	}

	resp, err := c.Get(ctx, endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search officers: %w", err)
	}
	defer resp.Body.Close()

	var officersResponse struct {
		Results struct {
			Officers []Officer `json:"officers"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&officersResponse); err != nil {
		return nil, fmt.Errorf("failed to decode officers response: %w", err)
	}

	c.logger.Info("Company officers retrieved",
		zap.Int("officers_found", len(officersResponse.Results.Officers)))

	return officersResponse.Results.Officers, nil
}

// SearchByIndustry searches for companies by industry code
func (c *OpenCorporatesClient) SearchByIndustry(ctx context.Context, industryCode, jurisdiction string) ([]Company, error) {
	c.logger.Info("Searching companies by industry",
		zap.String("industry_code", industryCode),
		zap.String("jurisdiction", jurisdiction))

	params := map[string]string{
		"industry_code": industryCode,
		"format": "json",
		"per_page": "20",
	}

	if jurisdiction != "" {
		params["jurisdiction_code"] = jurisdiction
	}

	resp, err := c.Get(ctx, "/companies/search", params)
	if err != nil {
		return nil, fmt.Errorf("failed to search by industry: %w", err)
	}
	defer resp.Body.Close()

	var searchResponse OpenCorporatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode industry search response: %w", err)
	}

	c.logger.Info("Industry search completed",
		zap.String("industry_code", industryCode),
		zap.Int("companies_found", len(searchResponse.Results.Companies)))

	return searchResponse.Results.Companies, nil
}

// analyzeCompanyRisk analyzes company data to determine risk score and compliance status
func (c *OpenCorporatesClient) analyzeCompanyRisk(companies []Company) (float64, string) {
	if len(companies) == 0 {
		return 0.0, "NOT_FOUND"
	}

	riskScore := 0.0
	complianceStatus := "COMPLIANT"

	for _, company := range companies {
		companyRisk := 0.0

		// Check company status
		switch company.Company.Status {
		case "Active", "Active (non-compliant)":
			companyRisk += 0.1
		case "Dissolved", "Inactive":
			companyRisk += 0.3
			complianceStatus = "DISSOLVED"
		case "Struck off":
			companyRisk += 0.4
			complianceStatus = "STRUCK_OFF"
		default:
			companyRisk += 0.2
		}

		// Check if company is inactive
		if company.Company.Inactive {
			companyRisk += 0.2
			complianceStatus = "INACTIVE"
		}

		// Check branch status
		if company.Company.BranchStatus == "Inactive" {
			companyRisk += 0.1
		}

		// Check for dissolution date
		if company.Company.DateOfDissolution != nil {
			companyRisk += 0.3
			complianceStatus = "DISSOLVED"
		}

		// Check company age (newer companies might be riskier)
		if company.Company.DateOfCreation.After(time.Now().AddDate(-1, 0, 0)) {
			companyRisk += 0.1
		}

		// Average the risk scores
		riskScore += companyRisk
	}

	riskScore = riskScore / float64(len(companies))

	// Ensure risk score is between 0 and 1
	if riskScore > 1.0 {
		riskScore = 1.0
	}

	return riskScore, complianceStatus
}

// IsHealthy checks if the OpenCorporates service is healthy
func (c *OpenCorporatesClient) IsHealthy(ctx context.Context) error {
	// Try to search for a well-known company as a health check
	_, err := c.SearchCompany(ctx, "Apple Inc", "us")
	if err != nil {
		return fmt.Errorf("OpenCorporates health check failed: %w", err)
	}
	return nil
}
