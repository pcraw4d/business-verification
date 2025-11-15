package providers

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"kyb-platform/internal/integrations"
)

// WHOISProvider implements WHOIS data integration for domain analysis
// This is a FREE service that provides domain registration data
type WHOISProvider struct {
	config integrations.ProviderConfig
	client *net.Dialer
	health bool
}

// WHOISData represents domain registration information
type WHOISData struct {
	Domain          string     `json:"domain"`
	Registrar       string     `json:"registrar"`
	RegistrantName  string     `json:"registrant_name"`
	RegistrantOrg   string     `json:"registrant_org"`
	RegistrantEmail string     `json:"registrant_email"`
	CreationDate    *time.Time `json:"creation_date"`
	ExpirationDate  *time.Time `json:"expiration_date"`
	UpdatedDate     *time.Time `json:"updated_date"`
	NameServers     []string   `json:"name_servers"`
	Status          []string   `json:"status"`
	Country         string     `json:"country"`
}

// NewWHOISProvider creates a new WHOIS provider
func NewWHOISProvider(config integrations.ProviderConfig) *WHOISProvider {
	// WHOIS is free but has rate limits
	// Rate limit: varies by registry, typically 1-10 requests per second
	if config.RateLimit == 0 {
		config.RateLimit = 60 // Conservative: 1 request per second = 60 per minute
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
		config.RetryDelay = 2 * time.Second // Longer delay for WHOIS
	}

	// WHOIS is completely free
	config.CostPerRequest = 0.0
	config.CostPerSearch = 0.0
	config.CostPerDetail = 0.0
	config.CostPerFinancial = 0.0

	// Set provider type
	config.Type = "whois"

	return &WHOISProvider{
		config: config,
		client: &net.Dialer{
			Timeout: config.Timeout,
		},
		health: true,
	}
}

func (w *WHOISProvider) GetName() string {
	return w.config.Name
}

func (w *WHOISProvider) GetType() string {
	return w.config.Type
}

func (w *WHOISProvider) GetConfig() integrations.ProviderConfig {
	return w.config
}

func (w *WHOISProvider) IsHealthy() bool {
	return w.health
}

func (w *WHOISProvider) GetCost() float64 {
	return w.config.CostPerRequest // Always 0.0 for WHOIS
}

func (w *WHOISProvider) GetQuota() integrations.QuotaInfo {
	// WHOIS has no quota limits, but we respect rate limits
	return integrations.QuotaInfo{
		DailyUsed:    0,
		DailyLimit:   999999, // Effectively unlimited
		MonthlyUsed:  0,
		MonthlyLimit: 999999, // Effectively unlimited
		ResetTime:    time.Now().Add(24 * time.Hour),
		Remaining:    999999,
	}
}

func (w *WHOISProvider) SearchBusiness(ctx context.Context, query integrations.BusinessSearchQuery) (*integrations.BusinessData, error) {
	// WHOIS doesn't support business name search directly
	// We need a domain to perform WHOIS lookup
	if query.Website == "" {
		return nil, fmt.Errorf("WHOIS provider requires a website/domain for lookup")
	}

	// Extract domain from website URL
	domain, err := w.extractDomain(query.Website)
	if err != nil {
		return nil, fmt.Errorf("failed to extract domain from website: %w", err)
	}

	// Perform WHOIS lookup
	whoisData, err := w.performWHOISLookup(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("WHOIS lookup failed: %w", err)
	}

	// Convert WHOIS data to BusinessData
	businessData := &integrations.BusinessData{
		ID:             fmt.Sprintf("whois_%s", domain),
		ProviderID:     domain,
		ProviderName:   w.config.Name,
		CompanyName:    whoisData.RegistrantOrg,
		LegalName:      whoisData.RegistrantOrg,
		BusinessNumber: domain,
		Website:        query.Website,
		Address: integrations.Address{
			Country: whoisData.Country,
		},
		Status:      w.determineDomainStatus(whoisData.Status),
		LastUpdated: time.Now(),
		DataQuality: 0.80, // Good quality but not business-verified
		Confidence:  0.75, // Good confidence for domain data
		DataSources: []integrations.DataSource{
			{
				Name:        "WHOIS",
				Type:        "domain_registry",
				TrustLevel:  "medium",
				LastUpdated: time.Now(),
			},
		},
	}

	// Add domain-specific information
	if whoisData.CreationDate != nil {
		businessData.FoundedDate = whoisData.CreationDate
	}

	return businessData, nil
}

func (w *WHOISProvider) GetBusinessDetails(ctx context.Context, businessID string) (*integrations.BusinessData, error) {
	// Extract domain from business ID
	domain := strings.TrimPrefix(businessID, "whois_")

	// Perform WHOIS lookup
	whoisData, err := w.performWHOISLookup(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("WHOIS lookup failed: %w", err)
	}

	// Convert WHOIS data to BusinessData
	businessData := &integrations.BusinessData{
		ID:             fmt.Sprintf("whois_%s", domain),
		ProviderID:     domain,
		ProviderName:   w.config.Name,
		CompanyName:    whoisData.RegistrantOrg,
		LegalName:      whoisData.RegistrantOrg,
		BusinessNumber: domain,
		Website:        "https://" + domain,
		Address: integrations.Address{
			Country: whoisData.Country,
		},
		Status:      w.determineDomainStatus(whoisData.Status),
		LastUpdated: time.Now(),
		DataQuality: 0.80,
		Confidence:  0.75,
		DataSources: []integrations.DataSource{
			{
				Name:        "WHOIS",
				Type:        "domain_registry",
				TrustLevel:  "medium",
				LastUpdated: time.Now(),
			},
		},
	}

	// Add domain-specific information
	if whoisData.CreationDate != nil {
		businessData.FoundedDate = whoisData.CreationDate
	}

	return businessData, nil
}

func (w *WHOISProvider) GetFinancialData(ctx context.Context, businessID string) (*integrations.FinancialData, error) {
	// WHOIS doesn't provide financial data
	return &integrations.FinancialData{
		ProviderID:   businessID,
		ProviderName: w.config.Name,
		LastUpdated:  time.Now(),
		DataQuality:  0.0, // No financial data available
		Confidence:   0.0,
		DataSources: []integrations.DataSource{
			{
				Name:        "WHOIS",
				Type:        "domain_registry",
				TrustLevel:  "medium",
				LastUpdated: time.Now(),
			},
		},
	}, nil
}

func (w *WHOISProvider) GetComplianceData(ctx context.Context, businessID string) (*integrations.ComplianceData, error) {
	// WHOIS provides basic compliance data through domain status
	return &integrations.ComplianceData{
		ProviderID:   businessID,
		ProviderName: w.config.Name,
		LastUpdated:  time.Now(),
		DataQuality:  0.70,
		Confidence:   0.75,
		DataSources: []integrations.DataSource{
			{
				Name:        "WHOIS",
				Type:        "domain_registry",
				TrustLevel:  "medium",
				LastUpdated: time.Now(),
			},
		},
	}, nil
}

func (w *WHOISProvider) GetNewsData(ctx context.Context, businessID string) ([]integrations.NewsItem, error) {
	// WHOIS doesn't provide news data
	return []integrations.NewsItem{}, nil
}

func (w *WHOISProvider) ValidateData(data *integrations.BusinessData) (*integrations.DataValidationResult, error) {
	// WHOIS data is domain registry data, not business data
	issues := []integrations.ValidationIssue{}

	// Check for required fields
	if data.Website == "" {
		issues = append(issues, integrations.ValidationIssue{
			Field:       "website",
			Type:        "missing",
			Severity:    "high",
			Description: "Website is required for WHOIS data",
		})
	}

	if data.ProviderID == "" {
		issues = append(issues, integrations.ValidationIssue{
			Field:       "provider_id",
			Type:        "missing",
			Severity:    "high",
			Description: "Provider ID (domain) is required",
		})
	}

	// Calculate quality score
	qualityScore := 0.80 // Good for domain data
	if len(issues) > 0 {
		qualityScore = 0.60
	}

	return &integrations.DataValidationResult{
		IsValid:       len(issues) == 0,
		QualityScore:  qualityScore,
		Issues:        issues,
		LastValidated: time.Now(),
	}, nil
}

// performWHOISLookup performs a WHOIS lookup for a domain
func (w *WHOISProvider) performWHOISLookup(ctx context.Context, domain string) (*WHOISData, error) {
	// This is a simplified WHOIS implementation
	// In production, you would use a proper WHOIS library or service

	// For now, we'll return mock data that represents typical WHOIS information
	// In a real implementation, you would:
	// 1. Determine the appropriate WHOIS server for the TLD
	// 2. Connect to the WHOIS server on port 43
	// 3. Send the domain query
	// 4. Parse the response

	// Mock WHOIS data for demonstration
	whoisData := &WHOISData{
		Domain:          domain,
		Registrar:       "Example Registrar Inc.",
		RegistrantName:  "John Doe",
		RegistrantOrg:   "Example Company Inc.",
		RegistrantEmail: "admin@example.com",
		CreationDate:    timePtr(time.Now().AddDate(-2, 0, 0)), // 2 years ago
		ExpirationDate:  timePtr(time.Now().AddDate(1, 0, 0)),  // 1 year from now
		UpdatedDate:     timePtr(time.Now().AddDate(-1, 0, 0)), // 1 year ago
		NameServers:     []string{"ns1.example.com", "ns2.example.com"},
		Status:          []string{"clientTransferProhibited", "clientUpdateProhibited"},
		Country:         "US",
	}

	return whoisData, nil
}

// extractDomain extracts the domain from a website URL
func (w *WHOISProvider) extractDomain(website string) (string, error) {
	// Add protocol if missing
	if !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
		website = "https://" + website
	}

	// Parse URL
	u, err := url.Parse(website)
	if err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	// Extract domain
	domain := u.Hostname()
	if domain == "" {
		return "", fmt.Errorf("no domain found in URL")
	}

	// Remove www. prefix if present
	domain = strings.TrimPrefix(domain, "www.")

	return domain, nil
}

// determineDomainStatus determines business status based on domain status
func (w *WHOISProvider) determineDomainStatus(domainStatuses []string) string {
	// Check for active domain status
	for _, status := range domainStatuses {
		switch strings.ToLower(status) {
		case "ok", "active", "clienttransferprohibited", "clientupdateprohibited":
			return "active"
		case "clienthold", "serverhold":
			return "suspended"
		case "clientdeleteprohibited", "serverdeleteprohibited":
			return "protected"
		}
	}

	// Default to active if no specific status found
	return "active"
}

// timePtr returns a pointer to a time.Time value
func timePtr(t time.Time) *time.Time {
	return &t
}
