package providers

import (
	"fmt"
	"log"
	"time"

	"kyb-platform/internal/integrations"
)

// GovernmentProvidersFactory creates and configures government API providers
type GovernmentProvidersFactory struct {
	logger *log.Logger
}

// NewGovernmentProvidersFactory creates a new factory for government providers
func NewGovernmentProvidersFactory(logger *log.Logger) *GovernmentProvidersFactory {
	if logger == nil {
		logger = log.Default()
	}

	return &GovernmentProvidersFactory{
		logger: logger,
	}
}

// CreateSECEdgarProvider creates a SEC EDGAR provider with default configuration
func (f *GovernmentProvidersFactory) CreateSECEdgarProvider() *SECEdgarProvider {
	config := integrations.ProviderConfig{
		Name:             "SEC EDGAR",
		Type:             "sec_edgar",
		BaseURL:          "https://data.sec.gov",
		RateLimit:        600, // 10 requests per second = 600 per minute
		BurstLimit:       10,
		Timeout:          30 * time.Second,
		RetryAttempts:    3,
		RetryDelay:       1 * time.Second,
		CostPerRequest:   0.0,
		CostPerSearch:    0.0,
		CostPerDetail:    0.0,
		CostPerFinancial: 0.0,
		DataQuality:      0.95, // Government data is high quality
		Coverage: map[string]float64{
			"US": 1.0, // SEC EDGAR covers US companies
		},
		AuthType: "none", // SEC EDGAR doesn't require authentication
	}

	f.logger.Printf("Created SEC EDGAR provider with rate limit: %d requests/minute", config.RateLimit)
	return NewSECEdgarProvider(config)
}

// CreateCompaniesHouseProvider creates a Companies House provider with default configuration
func (f *GovernmentProvidersFactory) CreateCompaniesHouseProvider(apiKey string) *CompaniesHouseProvider {
	config := integrations.ProviderConfig{
		Name:             "Companies House",
		Type:             "companies_house",
		BaseURL:          "https://api.company-information.service.gov.uk",
		APIKey:           apiKey,
		RateLimit:        120, // 600 requests per 5 minutes = 120 per minute
		BurstLimit:       10,
		Timeout:          30 * time.Second,
		RetryAttempts:    3,
		RetryDelay:       1 * time.Second,
		CostPerRequest:   0.0,
		CostPerSearch:    0.0,
		CostPerDetail:    0.0,
		CostPerFinancial: 0.0,
		DataQuality:      0.95, // Government data is high quality
		Coverage: map[string]float64{
			"GB": 1.0, // Companies House covers UK companies
		},
		AuthType: "basic", // Companies House uses basic auth with API key
	}

	if apiKey == "" {
		f.logger.Printf("Warning: Companies House provider created without API key - requests will fail")
	} else {
		f.logger.Printf("Created Companies House provider with rate limit: %d requests/minute", config.RateLimit)
	}

	return NewCompaniesHouseProvider(config)
}

// CreateOpenCorporatesProvider creates an OpenCorporates provider with default configuration
func (f *GovernmentProvidersFactory) CreateOpenCorporatesProvider(apiToken string) *OpenCorporatesProvider {
	config := integrations.ProviderConfig{
		Name:             "OpenCorporates",
		Type:             "opencorporates",
		BaseURL:          "https://api.opencorporates.com",
		APIKey:           apiToken,
		RateLimit:        500, // 500 requests per day = ~20 per hour = ~0.3 per minute
		BurstLimit:       5,
		Timeout:          30 * time.Second,
		RetryAttempts:    3,
		RetryDelay:       1 * time.Second,
		CostPerRequest:   0.0,
		CostPerSearch:    0.0,
		CostPerDetail:    0.0,
		CostPerFinancial: 0.0,
		DataQuality:      0.85, // Good quality but not government-verified
		Coverage: map[string]float64{
			"US": 0.9, // Good coverage of US companies
			"GB": 0.9, // Good coverage of UK companies
			"CA": 0.8, // Good coverage of Canadian companies
			"AU": 0.8, // Good coverage of Australian companies
			"DE": 0.7, // Moderate coverage of German companies
			"FR": 0.7, // Moderate coverage of French companies
			"IT": 0.6, // Limited coverage of Italian companies
			"ES": 0.6, // Limited coverage of Spanish companies
			"NL": 0.7, // Moderate coverage of Dutch companies
			"BE": 0.6, // Limited coverage of Belgian companies
		},
		AuthType: "api_key", // OpenCorporates uses API key authentication
	}

	if apiToken == "" {
		f.logger.Printf("Warning: OpenCorporates provider created without API token - limited to free tier")
	} else {
		f.logger.Printf("Created OpenCorporates provider with rate limit: %d requests/day", config.RateLimit)
	}

	return NewOpenCorporatesProvider(config)
}

// CreateWHOISProvider creates a WHOIS provider with default configuration
func (f *GovernmentProvidersFactory) CreateWHOISProvider() *WHOISProvider {
	config := integrations.ProviderConfig{
		Name:             "WHOIS",
		Type:             "whois",
		RateLimit:        60, // Conservative: 1 request per second = 60 per minute
		BurstLimit:       5,
		Timeout:          30 * time.Second,
		RetryAttempts:    3,
		RetryDelay:       2 * time.Second, // Longer delay for WHOIS
		CostPerRequest:   0.0,
		CostPerSearch:    0.0,
		CostPerDetail:    0.0,
		CostPerFinancial: 0.0,
		DataQuality:      0.80, // Good quality for domain data
		Coverage: map[string]float64{
			"global": 1.0, // WHOIS covers all domains globally
		},
		AuthType: "none", // WHOIS doesn't require authentication
	}

	f.logger.Printf("Created WHOIS provider with rate limit: %d requests/minute", config.RateLimit)
	return NewWHOISProvider(config)
}

// RegisterAllGovernmentProviders registers all government API providers with the service
func (f *GovernmentProvidersFactory) RegisterAllGovernmentProviders(service *integrations.BusinessDataAPIService, config integrations.GovernmentAPIsConfig) error {
	var errors []error

	// Register SEC EDGAR provider (always free, no API key required)
	secProvider := f.CreateSECEdgarProvider()
	if err := service.RegisterProvider(secProvider); err != nil {
		errors = append(errors, fmt.Errorf("failed to register SEC EDGAR provider: %w", err))
	} else {
		f.logger.Printf("Successfully registered SEC EDGAR provider")
	}

	// Register Companies House provider (free but requires API key)
	if config.CompaniesHouseAPIKey != "" {
		chProvider := f.CreateCompaniesHouseProvider(config.CompaniesHouseAPIKey)
		if err := service.RegisterProvider(chProvider); err != nil {
			errors = append(errors, fmt.Errorf("failed to register Companies House provider: %w", err))
		} else {
			f.logger.Printf("Successfully registered Companies House provider")
		}
	} else {
		f.logger.Printf("Skipping Companies House provider - no API key provided")
	}

	// Register OpenCorporates provider (free tier available)
	ocProvider := f.CreateOpenCorporatesProvider(config.OpenCorporatesAPIToken)
	if err := service.RegisterProvider(ocProvider); err != nil {
		errors = append(errors, fmt.Errorf("failed to register OpenCorporates provider: %w", err))
	} else {
		f.logger.Printf("Successfully registered OpenCorporates provider")
	}

	// Register WHOIS provider (always free, no API key required)
	whoisProvider := f.CreateWHOISProvider()
	if err := service.RegisterProvider(whoisProvider); err != nil {
		errors = append(errors, fmt.Errorf("failed to register WHOIS provider: %w", err))
	} else {
		f.logger.Printf("Successfully registered WHOIS provider")
	}

	// Return combined errors if any
	if len(errors) > 0 {
		return fmt.Errorf("failed to register some government providers: %v", errors)
	}

	f.logger.Printf("Successfully registered all government API providers")
	return nil
}

// Note: GovernmentAPIsConfig, GetDefaultGovernmentAPIsConfig, and ValidateGovernmentAPIsConfig
// are now defined in the parent integrations package
