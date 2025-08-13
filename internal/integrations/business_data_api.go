package integrations

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BusinessDataAPIService provides integration with real business data providers
type BusinessDataAPIService struct {
	providers    map[string]BusinessDataProvider
	rateLimiters map[string]*RateLimiter
	cache        *APICache
	config       BusinessDataAPIConfig
	mu           sync.RWMutex
}

// BusinessDataAPIConfig holds configuration for business data API integration
type BusinessDataAPIConfig struct {
	// Provider configuration
	Providers        map[string]ProviderConfig `json:"providers"`
	DefaultProvider  string                    `json:"default_provider"`
	FallbackProvider string                    `json:"fallback_provider"`

	// Rate limiting and quotas
	RateLimiting      bool `json:"rate_limiting"`
	GlobalRateLimit   int  `json:"global_rate_limit"`   // requests per minute
	ProviderRateLimit int  `json:"provider_rate_limit"` // requests per minute per provider

	// Caching configuration
	CachingEnabled bool          `json:"caching_enabled"`
	CacheTTL       time.Duration `json:"cache_ttl"`
	CacheSize      int           `json:"cache_size"`

	// Cost optimization
	CostTracking     bool    `json:"cost_tracking"`
	BudgetLimit      float64 `json:"budget_limit"`
	CostOptimization bool    `json:"cost_optimization"`

	// Data quality and validation
	DataValidation   bool    `json:"data_validation"`
	QualityThreshold float64 `json:"quality_threshold"`
	DuplicateCheck   bool    `json:"duplicate_check"`

	// Monitoring and alerting
	MonitoringEnabled   bool          `json:"monitoring_enabled"`
	AlertThreshold      float64       `json:"alert_threshold"`
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

// BusinessDataProvider represents a business data provider
type BusinessDataProvider interface {
	GetName() string
	GetType() string
	GetConfig() ProviderConfig
	IsHealthy() bool
	GetCost() float64
	GetQuota() QuotaInfo
	SearchBusiness(ctx context.Context, query BusinessSearchQuery) (*BusinessData, error)
	GetBusinessDetails(ctx context.Context, businessID string) (*BusinessData, error)
	GetFinancialData(ctx context.Context, businessID string) (*FinancialData, error)
	GetComplianceData(ctx context.Context, businessID string) (*ComplianceData, error)
	GetNewsData(ctx context.Context, businessID string) ([]NewsItem, error)
	ValidateData(data *BusinessData) (*DataValidationResult, error)
}

// ProviderConfig holds configuration for a specific provider
type ProviderConfig struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // dnb, experian, sec, companies_house, bloomberg, reuters, factiva, lexisnexis
	BaseURL      string `json:"base_url"`
	APIKey       string `json:"api_key"`
	SecretKey    string `json:"secret_key"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`

	// Rate limiting
	RateLimit  int `json:"rate_limit"` // requests per minute
	BurstLimit int `json:"burst_limit"`

	// Cost information
	CostPerRequest   float64 `json:"cost_per_request"`
	CostPerSearch    float64 `json:"cost_per_search"`
	CostPerDetail    float64 `json:"cost_per_detail"`
	CostPerFinancial float64 `json:"cost_per_financial"`

	// Quota information
	DailyQuota     int    `json:"daily_quota"`
	MonthlyQuota   int    `json:"monthly_quota"`
	QuotaResetTime string `json:"quota_reset_time"`

	// Authentication
	AuthType    string            `json:"auth_type"` // api_key, oauth, basic, bearer
	AuthHeaders map[string]string `json:"auth_headers"`

	// Timeout and retry
	Timeout       time.Duration `json:"timeout"`
	RetryAttempts int           `json:"retry_attempts"`
	RetryDelay    time.Duration `json:"retry_delay"`

	// Data quality
	DataQuality     float64            `json:"data_quality"`
	Coverage        map[string]float64 `json:"coverage"` // country/region coverage
	UpdateFrequency string             `json:"update_frequency"`

	// Features
	Features            []string `json:"features"` // search, details, financial, compliance, news
	SupportedCountries  []string `json:"supported_countries"`
	SupportedIndustries []string `json:"supported_industries"`
}

// BusinessSearchQuery represents a business search request
type BusinessSearchQuery struct {
	CompanyName       string         `json:"company_name"`
	BusinessNumber    string         `json:"business_number"`
	TaxID             string         `json:"tax_id"`
	Country           string         `json:"country"`
	State             string         `json:"state"`
	City              string         `json:"city"`
	Industry          string         `json:"industry"`
	SICCode           string         `json:"sic_code"`
	NAICSCode         string         `json:"naics_code"`
	EmployeeCount     *EmployeeRange `json:"employee_count"`
	Revenue           *RevenueRange  `json:"revenue"`
	FoundedYear       *YearRange     `json:"founded_year"`
	Status            string         `json:"status"` // active, inactive, dissolved
	IncludeFinancial  bool           `json:"include_financial"`
	IncludeCompliance bool           `json:"include_compliance"`
	IncludeNews       bool           `json:"include_news"`
	MaxResults        int            `json:"max_results"`
	SortBy            string         `json:"sort_by"`
	SortOrder         string         `json:"sort_order"`
}

// BusinessData represents comprehensive business information
type BusinessData struct {
	ID             string   `json:"id"`
	ProviderID     string   `json:"provider_id"`
	ProviderName   string   `json:"provider_name"`
	CompanyName    string   `json:"company_name"`
	LegalName      string   `json:"legal_name"`
	TradeNames     []string `json:"trade_names"`
	BusinessNumber string   `json:"business_number"`
	TaxID          string   `json:"tax_id"`
	DUNSNumber     string   `json:"duns_number"`

	// Address information
	Address           Address `json:"address"`
	MailingAddress    Address `json:"mailing_address"`
	RegisteredAddress Address `json:"registered_address"`

	// Contact information
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website"`

	// Business details
	Industry          string    `json:"industry"`
	SICCode           string    `json:"sic_code"`
	NAICSCode         string    `json:"naics_code"`
	BusinessType      string    `json:"business_type"`
	LegalStructure    string    `json:"legal_structure"`
	IncorporationDate time.Time `json:"incorporation_date"`
	Status            string    `json:"status"`

	// Size and scale
	EmployeeCount int     `json:"employee_count"`
	Revenue       float64 `json:"revenue"`
	Assets        float64 `json:"assets"`
	MarketCap     float64 `json:"market_cap"`

	// Ownership and management
	Owners     []Owner     `json:"owners"`
	Directors  []Director  `json:"directors"`
	Executives []Executive `json:"executives"`

	// Financial information
	FinancialData *FinancialData `json:"financial_data,omitempty"`

	// Compliance information
	ComplianceData *ComplianceData `json:"compliance_data,omitempty"`

	// News and media
	NewsData []NewsItem `json:"news_data,omitempty"`

	// Metadata
	LastUpdated time.Time `json:"last_updated"`
	DataQuality float64   `json:"data_quality"`
	Confidence  float64   `json:"confidence"`
	SourceURL   string    `json:"source_url"`
}

// Address represents a business address
type Address struct {
	Street1     string  `json:"street1"`
	Street2     string  `json:"street2"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	PostalCode  string  `json:"postal_code"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
}

// Owner represents a business owner
type Owner struct {
	Name             string      `json:"name"`
	Title            string      `json:"title"`
	OwnershipPercent float64     `json:"ownership_percent"`
	Address          Address     `json:"address"`
	ContactInfo      ContactInfo `json:"contact_info"`
}

// Director represents a company director
type Director struct {
	Name            string      `json:"name"`
	Title           string      `json:"title"`
	AppointmentDate time.Time   `json:"appointment_date"`
	ResignationDate *time.Time  `json:"resignation_date,omitempty"`
	Address         Address     `json:"address"`
	ContactInfo     ContactInfo `json:"contact_info"`
}

// Executive represents a company executive
type Executive struct {
	Name         string      `json:"name"`
	Title        string      `json:"title"`
	StartDate    time.Time   `json:"start_date"`
	EndDate      *time.Time  `json:"end_date,omitempty"`
	Compensation float64     `json:"compensation"`
	Address      Address     `json:"address"`
	ContactInfo  ContactInfo `json:"contact_info"`
}

// ContactInfo represents contact information
type ContactInfo struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Fax   string `json:"fax"`
}

// FinancialData represents financial information
type FinancialData struct {
	FiscalYear       int       `json:"fiscal_year"`
	Revenue          float64   `json:"revenue"`
	NetIncome        float64   `json:"net_income"`
	TotalAssets      float64   `json:"total_assets"`
	TotalLiabilities float64   `json:"total_liabilities"`
	Equity           float64   `json:"equity"`
	CashFlow         float64   `json:"cash_flow"`
	EBITDA           float64   `json:"ebitda"`
	DebtToEquity     float64   `json:"debt_to_equity"`
	CurrentRatio     float64   `json:"current_ratio"`
	QuickRatio       float64   `json:"quick_ratio"`
	ROE              float64   `json:"roe"`
	ROA              float64   `json:"roa"`
	GrossMargin      float64   `json:"gross_margin"`
	NetMargin        float64   `json:"net_margin"`
	Currency         string    `json:"currency"`
	LastUpdated      time.Time `json:"last_updated"`
}

// ComplianceData represents compliance information
type ComplianceData struct {
	RegulatoryStatus string        `json:"regulatory_status"`
	LicenseNumbers   []string      `json:"license_numbers"`
	Certifications   []string      `json:"certifications"`
	Violations       []Violation   `json:"violations"`
	AuditReports     []AuditReport `json:"audit_reports"`
	ComplianceScore  float64       `json:"compliance_score"`
	RiskLevel        string        `json:"risk_level"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// Violation represents a compliance violation
type Violation struct {
	Type           string     `json:"type"`
	Description    string     `json:"description"`
	Date           time.Time  `json:"date"`
	Penalty        float64    `json:"penalty"`
	Status         string     `json:"status"`
	ResolutionDate *time.Time `json:"resolution_date,omitempty"`
}

// AuditReport represents an audit report
type AuditReport struct {
	Type            string    `json:"type"`
	Date            time.Time `json:"date"`
	Auditor         string    `json:"auditor"`
	Result          string    `json:"result"`
	Findings        []string  `json:"findings"`
	Recommendations []string  `json:"recommendations"`
}

// NewsItem represents a news item
type NewsItem struct {
	Title         string    `json:"title"`
	Summary       string    `json:"summary"`
	Content       string    `json:"content"`
	URL           string    `json:"url"`
	Source        string    `json:"source"`
	Author        string    `json:"author"`
	PublishedDate time.Time `json:"published_date"`
	Sentiment     string    `json:"sentiment"`
	Relevance     float64   `json:"relevance"`
	Tags          []string  `json:"tags"`
}

// EmployeeRange represents employee count range
type EmployeeRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

// RevenueRange represents revenue range
type RevenueRange struct {
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Currency string  `json:"currency"`
}

// YearRange represents year range
type YearRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// QuotaInfo represents API quota information
type QuotaInfo struct {
	DailyUsed    int       `json:"daily_used"`
	DailyLimit   int       `json:"daily_limit"`
	MonthlyUsed  int       `json:"monthly_used"`
	MonthlyLimit int       `json:"monthly_limit"`
	ResetTime    time.Time `json:"reset_time"`
	Remaining    int       `json:"remaining"`
}

// DataValidationResult represents data validation results
type DataValidationResult struct {
	IsValid      bool              `json:"is_valid"`
	QualityScore float64           `json:"quality_score"`
	Issues       []ValidationIssue `json:"issues"`
	Suggestions  []string          `json:"suggestions"`
	Confidence   float64           `json:"confidence"`
}

// ValidationIssue represents a data validation issue
type ValidationIssue struct {
	Field          string `json:"field"`
	Type           string `json:"type"`     // missing, invalid, inconsistent, outdated
	Severity       string `json:"severity"` // low, medium, high, critical
	Description    string `json:"description"`
	SuggestedValue string `json:"suggested_value,omitempty"`
}

// RateLimiter manages API rate limiting
type RateLimiter struct {
	provider   string
	rateLimit  int
	burstLimit int
	tokens     chan struct{}
	lastRefill time.Time
	mu         sync.Mutex
}

// APICache manages API response caching
type APICache struct {
	cache   map[string]CacheEntry
	ttl     time.Duration
	maxSize int
	mu      sync.RWMutex
}

// CacheEntry represents a cached API response
type CacheEntry struct {
	Data      interface{}   `json:"data"`
	Timestamp time.Time     `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

// NewBusinessDataAPIService creates a new business data API service
func NewBusinessDataAPIService(config BusinessDataAPIConfig) *BusinessDataAPIService {
	if config.GlobalRateLimit == 0 {
		config.GlobalRateLimit = 1000
	}

	if config.ProviderRateLimit == 0 {
		config.ProviderRateLimit = 100
	}

	if config.CacheTTL == 0 {
		config.CacheTTL = 1 * time.Hour
	}

	if config.CacheSize == 0 {
		config.CacheSize = 10000
	}

	if config.BudgetLimit == 0 {
		config.BudgetLimit = 1000.0
	}

	if config.QualityThreshold == 0 {
		config.QualityThreshold = 0.8
	}

	if config.AlertThreshold == 0 {
		config.AlertThreshold = 0.9
	}

	if config.HealthCheckInterval == 0 {
		config.HealthCheckInterval = 5 * time.Minute
	}

	return &BusinessDataAPIService{
		providers:    make(map[string]BusinessDataProvider),
		rateLimiters: make(map[string]*RateLimiter),
		cache:        NewAPICache(config.CacheTTL, config.CacheSize),
		config:       config,
	}
}

// RegisterProvider registers a business data provider
func (s *BusinessDataAPIService) RegisterProvider(provider BusinessDataProvider) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := provider.GetName()
	s.providers[name] = provider

	// Create rate limiter for provider
	config := provider.GetConfig()
	rateLimiter := NewRateLimiter(name, config.RateLimit, config.BurstLimit)
	s.rateLimiters[name] = rateLimiter

	return nil
}

// SearchBusiness searches for business information across all providers
func (s *BusinessDataAPIService) SearchBusiness(ctx context.Context, query BusinessSearchQuery) (*BusinessData, error) {
	// Check cache first
	cacheKey := generateCacheKey("search", query)
	if s.config.CachingEnabled {
		if cached := s.cache.Get(cacheKey); cached != nil {
			if data, ok := cached.(*BusinessData); ok {
				return data, nil
			}
		}
	}

	// Select best provider based on query
	provider := s.selectBestProvider(query)
	if provider == nil {
		return nil, fmt.Errorf("no suitable provider found for query")
	}

	// Check rate limits
	if !s.checkRateLimit(provider.GetName()) {
		return nil, fmt.Errorf("rate limit exceeded for provider %s", provider.GetName())
	}

	// Perform search
	data, err := provider.SearchBusiness(ctx, query)
	if err != nil {
		// Try fallback provider
		if fallback := s.getFallbackProvider(); fallback != nil {
			data, err = fallback.SearchBusiness(ctx, query)
		}
		if err != nil {
			return nil, fmt.Errorf("search failed: %w", err)
		}
	}

	// Validate data quality
	if s.config.DataValidation {
		validation, err := provider.ValidateData(data)
		if err != nil {
			return nil, fmt.Errorf("data validation failed: %w", err)
		}
		if validation.QualityScore < s.config.QualityThreshold {
			return nil, fmt.Errorf("data quality below threshold: %f", validation.QualityScore)
		}
	}

	// Cache result
	if s.config.CachingEnabled {
		s.cache.Set(cacheKey, data, s.config.CacheTTL)
	}

	// Track cost
	if s.config.CostTracking {
		s.trackCost(provider.GetName(), provider.GetConfig().CostPerSearch)
	}

	return data, nil
}

// GetBusinessDetails gets detailed business information
func (s *BusinessDataAPIService) GetBusinessDetails(ctx context.Context, businessID string, providerName string) (*BusinessData, error) {
	// Check cache first
	cacheKey := generateCacheKey("details", businessID, providerName)
	if s.config.CachingEnabled {
		if cached := s.cache.Get(cacheKey); cached != nil {
			if data, ok := cached.(*BusinessData); ok {
				return data, nil
			}
		}
	}

	// Get provider
	provider := s.getProvider(providerName)
	if provider == nil {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	// Check rate limits
	if !s.checkRateLimit(providerName) {
		return nil, fmt.Errorf("rate limit exceeded for provider %s", providerName)
	}

	// Get details
	data, err := provider.GetBusinessDetails(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get business details: %w", err)
	}

	// Cache result
	if s.config.CachingEnabled {
		s.cache.Set(cacheKey, data, s.config.CacheTTL)
	}

	// Track cost
	if s.config.CostTracking {
		s.trackCost(providerName, provider.GetConfig().CostPerDetail)
	}

	return data, nil
}

// GetFinancialData gets financial data for a business
func (s *BusinessDataAPIService) GetFinancialData(ctx context.Context, businessID string, providerName string) (*FinancialData, error) {
	// Check cache first
	cacheKey := generateCacheKey("financial", businessID, providerName)
	if s.config.CachingEnabled {
		if cached := s.cache.Get(cacheKey); cached != nil {
			if data, ok := cached.(*FinancialData); ok {
				return data, nil
			}
		}
	}

	// Get provider
	provider := s.getProvider(providerName)
	if provider == nil {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	// Check rate limits
	if !s.checkRateLimit(providerName) {
		return nil, fmt.Errorf("rate limit exceeded for provider %s", providerName)
	}

	// Get financial data
	data, err := provider.GetFinancialData(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial data: %w", err)
	}

	// Cache result
	if s.config.CachingEnabled {
		s.cache.Set(cacheKey, data, s.config.CacheTTL)
	}

	// Track cost
	if s.config.CostTracking {
		s.trackCost(providerName, provider.GetConfig().CostPerFinancial)
	}

	return data, nil
}

// GetComplianceData gets compliance data for a business
func (s *BusinessDataAPIService) GetComplianceData(ctx context.Context, businessID string, providerName string) (*ComplianceData, error) {
	// Get provider
	provider := s.getProvider(providerName)
	if provider == nil {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	// Check rate limits
	if !s.checkRateLimit(providerName) {
		return nil, fmt.Errorf("rate limit exceeded for provider %s", providerName)
	}

	// Get compliance data
	data, err := provider.GetComplianceData(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get compliance data: %w", err)
	}

	// Track cost
	if s.config.CostTracking {
		s.trackCost(providerName, provider.GetConfig().CostPerRequest)
	}

	return data, nil
}

// GetNewsData gets news data for a business
func (s *BusinessDataAPIService) GetNewsData(ctx context.Context, businessID string, providerName string) ([]NewsItem, error) {
	// Get provider
	provider := s.getProvider(providerName)
	if provider == nil {
		return nil, fmt.Errorf("provider %s not found", providerName)
	}

	// Check rate limits
	if !s.checkRateLimit(providerName) {
		return nil, fmt.Errorf("rate limit exceeded for provider %s", providerName)
	}

	// Get news data
	data, err := provider.GetNewsData(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get news data: %w", err)
	}

	// Track cost
	if s.config.CostTracking {
		s.trackCost(providerName, provider.GetConfig().CostPerRequest)
	}

	return data, nil
}

// selectBestProvider selects the best provider for a given query
func (s *BusinessDataAPIService) selectBestProvider(query BusinessSearchQuery) BusinessDataProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestProvider BusinessDataProvider
	var bestScore float64

	for _, provider := range s.providers {
		if !provider.IsHealthy() {
			continue
		}

		score := s.calculateProviderScore(provider, query)
		if score > bestScore {
			bestScore = score
			bestProvider = provider
		}
	}

	return bestProvider
}

// calculateProviderScore calculates a score for provider selection
func (s *BusinessDataAPIService) calculateProviderScore(provider BusinessDataProvider, query BusinessSearchQuery) float64 {
	config := provider.GetConfig()
	score := 0.0

	// Data quality score
	score += config.DataQuality * 0.3

	// Coverage score
	if coverage, exists := config.Coverage[query.Country]; exists {
		score += coverage * 0.2
	}

	// Cost optimization score
	if s.config.CostOptimization {
		costScore := 1.0 - (config.CostPerSearch / 10.0) // Normalize cost
		score += costScore * 0.2
	}

	// Feature availability score
	featureScore := 0.0
	if query.IncludeFinancial {
		if contains(config.Features, "financial") {
			featureScore += 0.25
		}
	}
	if query.IncludeCompliance {
		if contains(config.Features, "compliance") {
			featureScore += 0.25
		}
	}
	if query.IncludeNews {
		if contains(config.Features, "news") {
			featureScore += 0.25
		}
	}
	score += featureScore * 0.3

	return score
}

// getProvider gets a provider by name
func (s *BusinessDataAPIService) getProvider(name string) BusinessDataProvider {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.providers[name]
}

// getFallbackProvider gets the fallback provider
func (s *BusinessDataAPIService) getFallbackProvider() BusinessDataProvider {
	if s.config.FallbackProvider == "" {
		return nil
	}

	return s.getProvider(s.config.FallbackProvider)
}

// checkRateLimit checks if rate limit allows the request
func (s *BusinessDataAPIService) checkRateLimit(providerName string) bool {
	s.mu.RLock()
	rateLimiter, exists := s.rateLimiters[providerName]
	s.mu.RUnlock()

	if !exists {
		return true
	}

	return rateLimiter.Allow()
}

// trackCost tracks API usage cost
func (s *BusinessDataAPIService) trackCost(providerName string, cost float64) {
	// Implementation would track costs and check budget limits
	// For now, just log the cost
	fmt.Printf("API cost for %s: $%.4f\n", providerName, cost)
}

// generateCacheKey generates a cache key for the given parameters
func generateCacheKey(operation string, params ...interface{}) string {
	key := operation
	for _, param := range params {
		key += fmt.Sprintf(":%v", param)
	}
	return key
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(provider string, rateLimit, burstLimit int) *RateLimiter {
	rl := &RateLimiter{
		provider:   provider,
		rateLimit:  rateLimit,
		burstLimit: burstLimit,
		tokens:     make(chan struct{}, burstLimit),
		lastRefill: time.Now(),
	}

	// Fill initial tokens
	for i := 0; i < burstLimit; i++ {
		rl.tokens <- struct{}{}
	}

	return rl
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on time passed
	now := time.Now()
	timePassed := now.Sub(rl.lastRefill)
	tokensToAdd := int(timePassed.Minutes() * float64(rl.rateLimit))

	if tokensToAdd > 0 {
		for i := 0; i < tokensToAdd && len(rl.tokens) < rl.burstLimit; i++ {
			select {
			case rl.tokens <- struct{}{}:
			default:
				break
			}
		}
		rl.lastRefill = now
	}

	// Try to consume a token
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// NewAPICache creates a new API cache
func NewAPICache(ttl time.Duration, maxSize int) *APICache {
	return &APICache{
		cache:   make(map[string]CacheEntry),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Get gets a value from cache
func (c *APICache) Get(key string) interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil
	}

	// Check if entry is expired
	if time.Since(entry.Timestamp) > entry.TTL {
		delete(c.cache, key)
		return nil
	}

	return entry.Data
}

// Set sets a value in cache
func (c *APICache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache size
	if len(c.cache) >= c.maxSize {
		// Remove oldest entry
		var oldestKey string
		var oldestTime time.Time

		for k, v := range c.cache {
			if oldestKey == "" || v.Timestamp.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.Timestamp
			}
		}

		if oldestKey != "" {
			delete(c.cache, oldestKey)
		}
	}

	c.cache[key] = CacheEntry{
		Data:      value,
		Timestamp: time.Now(),
		TTL:       ttl,
	}
}
