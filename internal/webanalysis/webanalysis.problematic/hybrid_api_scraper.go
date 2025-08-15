package webanalysis

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// HybridAPIScraper combines ethical scraping with API fallbacks
type HybridAPIScraper struct {
	ethicalScraper  *EthicalScraper
	enhancedScraper *EnhancedEthicalScraper
	apiClients      map[string]APIClient
	config          HybridAPIScraperConfig
}

// HybridAPIScraperConfig holds configuration for hybrid API scraping
type HybridAPIScraperConfig struct {
	PrimaryMethod     string // "ethical_scraping", "api_first"
	EnableAPIFallback bool
	APITimeout        time.Duration
	MaxAPICost        float64 // Maximum cost per request in USD
	RateLimitDelay    time.Duration
}

// APIClient represents an API client interface
type APIClient interface {
	GetBusinessInfo(ctx context.Context, businessName, websiteURL string) (*BusinessInfo, error)
	GetCost() float64
	GetName() string
	IsAvailable() bool
}

// BusinessInfo represents business information from APIs
type BusinessInfo struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Industry    string                 `json:"industry"`
	WebsiteURL  string                 `json:"website_url"`
	Address     string                 `json:"address"`
	Phone       string                 `json:"phone"`
	Email       string                 `json:"email"`
	Founded     string                 `json:"founded"`
	Employees   string                 `json:"employees"`
	Revenue     string                 `json:"revenue"`
	Keywords    []string               `json:"keywords"`
	Categories  []string               `json:"categories"`
	Source      string                 `json:"source"`
	Confidence  float64                `json:"confidence"`
	RawData     map[string]interface{} `json:"raw_data"`
}

// HybridScrapingRequest represents a hybrid scraping request
type HybridScrapingRequest struct {
	BusinessName string
	WebsiteURL   string
	Priority     string  // "high", "medium", "low"
	Budget       float64 // Maximum cost willing to spend
	ContactEmail string
}

// HybridScrapingResponse represents the result of hybrid scraping
type HybridScrapingResponse struct {
	BusinessName   string
	WebsiteURL     string
	Method         string // "ethical_scraping", "api", "hybrid"
	Success        bool
	Data           *BusinessInfo
	ScrapedContent *EthicalScrapingResult
	APIData        *BusinessInfo
	TotalCost      float64
	ProcessingTime time.Duration
	Errors         []string
	Warnings       []string
	LegalStatus    string
	EthicalNotes   []string
}

// NewHybridAPIScraper creates a new hybrid API scraper
func NewHybridAPIScraper() *HybridAPIScraper {
	scraper := &HybridAPIScraper{
		ethicalScraper:  NewEthicalScraper(),
		enhancedScraper: NewEnhancedEthicalScraper(),
		apiClients:      make(map[string]APIClient),
		config: HybridAPIScraperConfig{
			PrimaryMethod:     "enhanced_ethical_scraping",
			EnableAPIFallback: true,
			APITimeout:        30 * time.Second,
			MaxAPICost:        0.50, // $0.50 max per request
			RateLimitDelay:    2 * time.Second,
		},
	}

	// Initialize API clients
	scraper.initializeAPIClients()

	return scraper
}

// initializeAPIClients sets up available API clients
func (has *HybridAPIScraper) initializeAPIClients() {
	// Free APIs (no cost)
	has.apiClients["google_places"] = NewGooglePlacesClient()
	has.apiClients["yelp"] = NewYelpClient()
	has.apiClients["open_corporates"] = NewOpenCorporatesClient()

	// Paid APIs (low cost)
	has.apiClients["crunchbase"] = NewCrunchbaseClient()
	has.apiClients["linkedin"] = NewLinkedInClient()
}

// ScrapeBusiness performs hybrid business scraping
func (has *HybridAPIScraper) ScrapeBusiness(ctx context.Context, req *HybridScrapingRequest) (*HybridScrapingResponse, error) {
	start := time.Now()
	response := &HybridScrapingResponse{
		BusinessName: req.BusinessName,
		WebsiteURL:   req.WebsiteURL,
		Method:       has.config.PrimaryMethod,
		Errors:       []string{},
		Warnings:     []string{},
		EthicalNotes: []string{},
	}

	// Strategy 1: Enhanced Ethical Web Scraping (if website URL provided)
	if req.WebsiteURL != "" && (has.config.PrimaryMethod == "enhanced_ethical_scraping" || has.config.PrimaryMethod == "ethical_scraping") {
		scrapedData, err := has.performEnhancedEthicalScraping(req)
		if err == nil && scrapedData != nil && scrapedData.ContentLength > 0 {
			response.ScrapedContent = has.convertEnhancedToEthicalResult(scrapedData)
			response.Success = true
			response.Method = "enhanced_ethical_scraping"
			response.LegalStatus = scrapedData.LegalStatus
			response.EthicalNotes = append(response.EthicalNotes, scrapedData.EthicalNotes...)
		} else {
			// Fallback to basic ethical scraping
			basicScrapedData, err := has.performEthicalScraping(req)
			if err == nil && basicScrapedData != nil {
				response.ScrapedContent = basicScrapedData
				response.Success = true
				response.Method = "ethical_scraping"
				response.LegalStatus = basicScrapedData.LegalStatus
				response.EthicalNotes = append(response.EthicalNotes, basicScrapedData.EthicalNotes...)
			} else {
				response.Errors = append(response.Errors, fmt.Sprintf("Enhanced ethical scraping failed: %v", err))
			}
		}
	}

	// Strategy 2: API Fallback (if enabled and scraping failed or for enrichment)
	if has.config.EnableAPIFallback && (!response.Success || has.shouldEnrichWithAPI(req)) {
		apiData, err := has.performAPIScraping(ctx, req)
		if err == nil && apiData != nil {
			response.APIData = apiData
			if !response.Success {
				response.Success = true
				response.Method = "api"
			} else {
				response.Method = "hybrid"
			}
			// Calculate cost based on the API client used
			if client, exists := has.apiClients[apiData.Source]; exists {
				response.TotalCost += client.GetCost()
			}
		} else {
			response.Errors = append(response.Errors, fmt.Sprintf("API scraping failed: %v", err))
		}
	}

	// Strategy 3: Combine and enrich data
	if response.Success {
		response.Data = has.combineData(response.ScrapedContent, response.APIData, req)
	}

	response.ProcessingTime = time.Since(start)
	return response, nil
}

// performEnhancedEthicalScraping performs enhanced ethical web scraping
func (has *HybridAPIScraper) performEnhancedEthicalScraping(req *HybridScrapingRequest) (*EnhancedScrapingResult, error) {
	// Try different extraction modes in order of preference
	modes := []string{"full", "minimal", "meta"}

	for _, mode := range modes {
		job := &EnhancedScrapingJob{
			URL:          req.WebsiteURL,
			BusinessName: req.BusinessName,
			Purpose:      "business_classification",
			ContactEmail: req.ContactEmail,
			ExtractMode:  mode,
		}

		result, err := has.enhancedScraper.ScrapeWebsite(job)
		if err == nil && result != nil && result.ContentLength > 100 { // Require at least 100 characters
			return result, nil
		}
	}

	// If all modes fail, return the last attempt
	job := &EnhancedScrapingJob{
		URL:          req.WebsiteURL,
		BusinessName: req.BusinessName,
		Purpose:      "business_classification",
		ContactEmail: req.ContactEmail,
		ExtractMode:  "minimal",
	}

	return has.enhancedScraper.ScrapeWebsite(job)
}

// convertEnhancedToEthicalResult converts enhanced result to ethical result format
func (has *HybridAPIScraper) convertEnhancedToEthicalResult(enhanced *EnhancedScrapingResult) *EthicalScrapingResult {
	return &EthicalScrapingResult{
		URL:          enhanced.URL,
		Title:        enhanced.Title,
		Text:         enhanced.Text,
		StatusCode:   enhanced.StatusCode,
		Method:       enhanced.Method,
		Error:        enhanced.Error,
		ScrapedAt:    enhanced.ScrapedAt,
		LegalStatus:  enhanced.LegalStatus,
		EthicalNotes: enhanced.EthicalNotes,
	}
}

// performEthicalScraping performs ethical web scraping
func (has *HybridAPIScraper) performEthicalScraping(req *HybridScrapingRequest) (*EthicalScrapingResult, error) {
	job := &EthicalScrapingJob{
		URL:          req.WebsiteURL,
		BusinessName: req.BusinessName,
		Purpose:      "business_classification",
		ContactEmail: req.ContactEmail,
	}

	return has.ethicalScraper.ScrapeWebsite(job)
}

// performAPIScraping performs API-based business information gathering
func (has *HybridAPIScraper) performAPIScraping(ctx context.Context, req *HybridScrapingRequest) (*BusinessInfo, error) {
	var bestResult *BusinessInfo
	var bestConfidence float64

	// Try free APIs first
	freeAPIs := []string{"google_places", "yelp", "open_corporates"}
	for _, apiName := range freeAPIs {
		if client, exists := has.apiClients[apiName]; exists && client.IsAvailable() {
			result, err := client.GetBusinessInfo(ctx, req.BusinessName, req.WebsiteURL)
			if err == nil && result != nil && result.Confidence > bestConfidence {
				bestResult = result
				bestConfidence = result.Confidence
			}
		}
	}

	// Try paid APIs if free APIs didn't work and budget allows
	if bestResult == nil && req.Budget > 0 {
		paidAPIs := []string{"crunchbase", "linkedin"}
		for _, apiName := range paidAPIs {
			if client, exists := has.apiClients[apiName]; exists && client.IsAvailable() {
				if client.GetCost() <= req.Budget {
					result, err := client.GetBusinessInfo(ctx, req.BusinessName, req.WebsiteURL)
					if err == nil && result != nil && result.Confidence > bestConfidence {
						bestResult = result
						bestConfidence = result.Confidence
					}
				}
			}
		}
	}

	return bestResult, nil
}

// shouldEnrichWithAPI determines if we should enrich with API data
func (has *HybridAPIScraper) shouldEnrichWithAPI(req *HybridScrapingRequest) bool {
	// Enrich if:
	// 1. High priority request
	// 2. Budget allows
	// 3. Business name is generic (needs enrichment)
	return req.Priority == "high" && req.Budget > 0 && has.isGenericBusinessName(req.BusinessName)
}

// isGenericBusinessName checks if business name is generic
func (has *HybridAPIScraper) isGenericBusinessName(name string) bool {
	genericWords := []string{"company", "corp", "inc", "llc", "ltd", "business", "enterprise", "group"}
	nameLower := strings.ToLower(name)

	for _, word := range genericWords {
		if strings.Contains(nameLower, word) {
			return true
		}
	}
	return false
}

// combineData combines scraped and API data
func (has *HybridAPIScraper) combineData(scraped *EthicalScrapingResult, api *BusinessInfo, req *HybridScrapingRequest) *BusinessInfo {
	combined := &BusinessInfo{
		Name:       req.BusinessName,
		WebsiteURL: req.WebsiteURL,
		Source:     "hybrid",
		RawData:    make(map[string]interface{}),
	}

	// Combine data sources
	if scraped != nil {
		combined.Description = scraped.Text
		combined.RawData["scraped_content"] = scraped
		combined.Confidence += 0.3 // Base confidence from scraping
	}

	if api != nil {
		// Use API data to enrich
		if combined.Description == "" {
			combined.Description = api.Description
		}
		if combined.Industry == "" {
			combined.Industry = api.Industry
		}
		if combined.Address == "" {
			combined.Address = api.Address
		}
		if combined.Phone == "" {
			combined.Phone = api.Phone
		}
		if combined.Email == "" {
			combined.Email = api.Email
		}
		if len(combined.Keywords) == 0 {
			combined.Keywords = api.Keywords
		}
		if len(combined.Categories) == 0 {
			combined.Categories = api.Categories
		}
		combined.RawData["api_data"] = api
		combined.Confidence += api.Confidence * 0.7 // Weight API confidence higher
	}

	// Normalize confidence
	if combined.Confidence > 1.0 {
		combined.Confidence = 1.0
	}

	return combined
}

// GetAvailableAPIs returns list of available APIs
func (has *HybridAPIScraper) GetAvailableAPIs() map[string]APIClient {
	available := make(map[string]APIClient)
	for name, client := range has.apiClients {
		if client.IsAvailable() {
			available[name] = client
		}
	}
	return available
}

// GetCostEstimate estimates the cost for a scraping request
func (has *HybridAPIScraper) GetCostEstimate(req *HybridScrapingRequest) float64 {
	cost := 0.0

	// Ethical scraping is free
	if req.WebsiteURL != "" {
		cost += 0.0
	}

	// API costs
	if has.config.EnableAPIFallback {
		// Estimate API costs based on available APIs
		for _, client := range has.apiClients {
			if client.IsAvailable() {
				cost += client.GetCost()
			}
		}
	}

	return cost
}
