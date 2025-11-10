package classification

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// StructuredDataExtractor extracts structured data from web pages
type StructuredDataExtractor struct {
	logger *log.Logger
}

// ExtractorBusinessInfo represents extracted business information
// (renamed to avoid conflict with content_relevance_analyzer.go)
type ExtractorBusinessInfo struct {
	BusinessName  string      `json:"business_name"`
	Description   string      `json:"description"`
	Services      []string    `json:"services"`
	Products      []string    `json:"products"`
	ContactInfo   ExtractorContactInfo `json:"contact_info"`
	BusinessHours string      `json:"business_hours"`
	Location      string      `json:"location"`
	Industry      string      `json:"industry"`
	BusinessType  string      `json:"business_type"`
}

// ExtractorContactInfo represents contact information
// (renamed to avoid conflict with content_relevance_analyzer.go)
type ExtractorContactInfo struct {
	Phone   string            `json:"phone"`
	Email   string            `json:"email"`
	Address string            `json:"address"`
	Website string            `json:"website"`
	Social  map[string]string `json:"social"`
}

// ExtractorStructuredDataResult represents extracted structured data
// (renamed to avoid conflict with smart_crawling_integration.go)
type ExtractorStructuredDataResult struct {
	SchemaOrgData   []SchemaOrgItem   `json:"schema_org_data"`
	OpenGraphData   map[string]string `json:"open_graph_data"`
	TwitterCardData map[string]string `json:"twitter_card_data"`
	Microdata       []MicrodataItem   `json:"microdata"`
	BusinessInfo    ExtractorBusinessInfo      `json:"business_info"`
	ContactInfo     ExtractorContactInfo       `json:"contact_info"`
	ProductInfo     []ProductInfo     `json:"product_info"`
	ServiceInfo     []ServiceInfo     `json:"service_info"`
	EventInfo       []EventInfo       `json:"event_info"`
	ExtractionScore float64           `json:"extraction_score"`
}

// SchemaOrgItem represents a Schema.org structured data item
type SchemaOrgItem struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Context    string                 `json:"context"`
	Confidence float64                `json:"confidence"`
}

// MicrodataItem represents microdata structured data
type MicrodataItem struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Context    string                 `json:"context"`
	Confidence float64                `json:"confidence"`
}

// ProductInfo represents product information
type ProductInfo struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       string  `json:"price"`
	Category    string  `json:"category"`
	Brand       string  `json:"brand"`
	SKU         string  `json:"sku"`
	Image       string  `json:"image"`
	URL         string  `json:"url"`
	Confidence  float64 `json:"confidence"`
}

// ServiceInfo represents service information
type ServiceInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Price       string   `json:"price"`
	Duration    string   `json:"duration"`
	Features    []string `json:"features"`
	URL         string   `json:"url"`
	Confidence  float64  `json:"confidence"`
}

// EventInfo represents event information
type EventInfo struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	Location    string  `json:"location"`
	URL         string  `json:"url"`
	Confidence  float64 `json:"confidence"`
}

// NewStructuredDataExtractor creates a new structured data extractor
func NewStructuredDataExtractor(logger *log.Logger) *StructuredDataExtractor {
	return &StructuredDataExtractor{
		logger: logger,
	}
}

// ExtractStructuredData extracts all types of structured data from HTML content
func (sde *StructuredDataExtractor) ExtractStructuredData(htmlContent string) *ExtractorStructuredDataResult {
	sde.logger.Printf("🔍 [StructuredData] Starting structured data extraction")

	result := &ExtractorStructuredDataResult{
		SchemaOrgData:   []SchemaOrgItem{},
		OpenGraphData:   make(map[string]string),
		TwitterCardData: make(map[string]string),
		Microdata:       []MicrodataItem{},
		BusinessInfo:    ExtractorBusinessInfo{},
		ContactInfo:     ExtractorContactInfo{},
		ProductInfo:     []ProductInfo{},
		ServiceInfo:     []ServiceInfo{},
		EventInfo:       []EventInfo{},
		ExtractionScore: 0.0,
	}

	// Extract Schema.org JSON-LD data
	sde.extractSchemaOrgJSONLD(htmlContent, result)

	// Extract Schema.org microdata
	sde.extractSchemaOrgMicrodata(htmlContent, result)

	// Extract Open Graph data
	sde.extractOpenGraphData(htmlContent, result)

	// Extract Twitter Card data
	sde.extractTwitterCardData(htmlContent, result)

	// Extract business information from structured data
	sde.extractBusinessInfoFromStructuredData(result)

	// Calculate extraction score
	result.ExtractionScore = sde.calculateExtractionScore(result)

	sde.logger.Printf("✅ [StructuredData] Extraction completed - Score: %.2f", result.ExtractionScore)
	return result
}

// extractSchemaOrgJSONLD extracts Schema.org JSON-LD structured data
func (sde *StructuredDataExtractor) extractSchemaOrgJSONLD(htmlContent string, result *ExtractorStructuredDataResult) {
	// Find all JSON-LD script tags
	jsonLDPattern := regexp.MustCompile(`<script[^>]*type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	matches := jsonLDPattern.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) > 1 {
			jsonData := strings.TrimSpace(match[1])

			// Parse JSON-LD data
			var data interface{}
			if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
				sde.logger.Printf("⚠️ [StructuredData] Failed to parse JSON-LD: %v", err)
				continue
			}

			// Process the structured data
			sde.processSchemaOrgData(data, result)
		}
	}
}

// processSchemaOrgData processes Schema.org structured data
func (sde *StructuredDataExtractor) processSchemaOrgData(data interface{}, result *ExtractorStructuredDataResult) {
	switch v := data.(type) {
	case map[string]interface{}:
		sde.processSchemaOrgObject(v, result)
	case []interface{}:
		for _, item := range v {
			sde.processSchemaOrgData(item, result)
		}
	}
}

// processSchemaOrgObject processes a single Schema.org object
func (sde *StructuredDataExtractor) processSchemaOrgObject(obj map[string]interface{}, result *ExtractorStructuredDataResult) {
	// Get the type
	typeValue, hasType := obj["@type"]
	if !hasType {
		return
	}

	schemaType := fmt.Sprintf("%v", typeValue)

	// Create Schema.org item
	item := SchemaOrgItem{
		Type:       schemaType,
		Properties: make(map[string]interface{}),
		Context:    "json-ld",
		Confidence: 0.9, // High confidence for structured data
	}

	// Copy properties
	for key, value := range obj {
		if key != "@type" && key != "@context" {
			item.Properties[key] = value
		}
	}

	result.SchemaOrgData = append(result.SchemaOrgData, item)

	// Extract specific business information based on type
	sde.extractBusinessInfoFromSchemaOrg(schemaType, obj, result)
}

// extractBusinessInfoFromSchemaOrg extracts business information from Schema.org data
func (sde *StructuredDataExtractor) extractBusinessInfoFromSchemaOrg(schemaType string, obj map[string]interface{}, result *ExtractorStructuredDataResult) {
	switch schemaType {
	case "Organization", "Corporation", "LocalBusiness", "Store", "Restaurant", "ProfessionalService":
		sde.extractOrganizationInfo(obj, result)
	case "Product":
		sde.extractProductInfo(obj, result)
	case "Service":
		sde.extractServiceInfo(obj, result)
	case "Event":
		sde.extractEventInfo(obj, result)
	}
}

// extractOrganizationInfo extracts organization information
func (sde *StructuredDataExtractor) extractOrganizationInfo(obj map[string]interface{}, result *ExtractorStructuredDataResult) {
	// Business name
	if name, exists := obj["name"]; exists {
		result.BusinessInfo.BusinessName = fmt.Sprintf("%v", name)
	}

	// Description
	if description, exists := obj["description"]; exists {
		result.BusinessInfo.Description = fmt.Sprintf("%v", description)
	}

	// Industry/Business type
	if businessType, exists := obj["@type"]; exists {
		result.BusinessInfo.BusinessType = fmt.Sprintf("%v", businessType)
	}

	// Contact information
	if telephone, exists := obj["telephone"]; exists {
		result.ContactInfo.Phone = fmt.Sprintf("%v", telephone)
	}

	if email, exists := obj["email"]; exists {
		result.ContactInfo.Email = fmt.Sprintf("%v", email)
	}

	if url, exists := obj["url"]; exists {
		result.ContactInfo.Website = fmt.Sprintf("%v", url)
	}

	// Address
	if address, exists := obj["address"]; exists {
		if addrMap, ok := address.(map[string]interface{}); ok {
			var addressParts []string
			if street, exists := addrMap["streetAddress"]; exists {
				addressParts = append(addressParts, fmt.Sprintf("%v", street))
			}
			if city, exists := addrMap["addressLocality"]; exists {
				addressParts = append(addressParts, fmt.Sprintf("%v", city))
			}
			if state, exists := addrMap["addressRegion"]; exists {
				addressParts = append(addressParts, fmt.Sprintf("%v", state))
			}
			if postal, exists := addrMap["postalCode"]; exists {
				addressParts = append(addressParts, fmt.Sprintf("%v", postal))
			}
			if country, exists := addrMap["addressCountry"]; exists {
				addressParts = append(addressParts, fmt.Sprintf("%v", country))
			}
			result.ContactInfo.Address = strings.Join(addressParts, ", ")
		}
	}

	// Business hours
	if openingHours, exists := obj["openingHours"]; exists {
		if hours, ok := openingHours.([]interface{}); ok {
			var hoursList []string
			for _, hour := range hours {
				hoursList = append(hoursList, fmt.Sprintf("%v", hour))
			}
			result.BusinessInfo.BusinessHours = strings.Join(hoursList, "; ")
		}
	}

	// Services
	if services, exists := obj["hasOfferCatalog"]; exists {
		if serviceMap, ok := services.(map[string]interface{}); ok {
			if itemList, exists := serviceMap["itemListElement"]; exists {
				if items, ok := itemList.([]interface{}); ok {
					for _, item := range items {
						if itemMap, ok := item.(map[string]interface{}); ok {
							if name, exists := itemMap["name"]; exists {
								result.BusinessInfo.Services = append(result.BusinessInfo.Services, fmt.Sprintf("%v", name))
							}
						}
					}
				}
			}
		}
	}
}

// extractProductInfo extracts product information
func (sde *StructuredDataExtractor) extractProductInfo(obj map[string]interface{}, result *ExtractorStructuredDataResult) {
	product := ProductInfo{
		Confidence: 0.9,
	}

	if name, exists := obj["name"]; exists {
		product.Name = fmt.Sprintf("%v", name)
	}

	if description, exists := obj["description"]; exists {
		product.Description = fmt.Sprintf("%v", description)
	}

	if price, exists := obj["offers"]; exists {
		if offerMap, ok := price.(map[string]interface{}); ok {
			if priceValue, exists := offerMap["price"]; exists {
				product.Price = fmt.Sprintf("%v", priceValue)
			}
		}
	}

	if category, exists := obj["category"]; exists {
		product.Category = fmt.Sprintf("%v", category)
	}

	if brand, exists := obj["brand"]; exists {
		if brandMap, ok := brand.(map[string]interface{}); ok {
			if brandName, exists := brandMap["name"]; exists {
				product.Brand = fmt.Sprintf("%v", brandName)
			}
		}
	}

	if sku, exists := obj["sku"]; exists {
		product.SKU = fmt.Sprintf("%v", sku)
	}

	if image, exists := obj["image"]; exists {
		if imageMap, ok := image.(map[string]interface{}); ok {
			if imageUrl, exists := imageMap["url"]; exists {
				product.Image = fmt.Sprintf("%v", imageUrl)
			}
		} else {
			product.Image = fmt.Sprintf("%v", image)
		}
	}

	if url, exists := obj["url"]; exists {
		product.URL = fmt.Sprintf("%v", url)
	}

	result.ProductInfo = append(result.ProductInfo, product)
}

// extractServiceInfo extracts service information
func (sde *StructuredDataExtractor) extractServiceInfo(obj map[string]interface{}, result *ExtractorStructuredDataResult) {
	service := ServiceInfo{
		Confidence: 0.9,
	}

	if name, exists := obj["name"]; exists {
		service.Name = fmt.Sprintf("%v", name)
	}

	if description, exists := obj["description"]; exists {
		service.Description = fmt.Sprintf("%v", description)
	}

	if category, exists := obj["category"]; exists {
		service.Category = fmt.Sprintf("%v", category)
	}

	if url, exists := obj["url"]; exists {
		service.URL = fmt.Sprintf("%v", url)
	}

	result.ServiceInfo = append(result.ServiceInfo, service)
}

// extractEventInfo extracts event information
func (sde *StructuredDataExtractor) extractEventInfo(obj map[string]interface{}, result *ExtractorStructuredDataResult) {
	event := EventInfo{
		Confidence: 0.9,
	}

	if name, exists := obj["name"]; exists {
		event.Name = fmt.Sprintf("%v", name)
	}

	if description, exists := obj["description"]; exists {
		event.Description = fmt.Sprintf("%v", description)
	}

	if startDate, exists := obj["startDate"]; exists {
		event.StartDate = fmt.Sprintf("%v", startDate)
	}

	if endDate, exists := obj["endDate"]; exists {
		event.EndDate = fmt.Sprintf("%v", endDate)
	}

	if location, exists := obj["location"]; exists {
		if locationMap, ok := location.(map[string]interface{}); ok {
			if locationName, exists := locationMap["name"]; exists {
				event.Location = fmt.Sprintf("%v", locationName)
			}
		} else {
			event.Location = fmt.Sprintf("%v", location)
		}
	}

	if url, exists := obj["url"]; exists {
		event.URL = fmt.Sprintf("%v", url)
	}

	result.EventInfo = append(result.EventInfo, event)
}

// extractSchemaOrgMicrodata extracts Schema.org microdata
func (sde *StructuredDataExtractor) extractSchemaOrgMicrodata(htmlContent string, result *ExtractorStructuredDataResult) {
	// This is a simplified implementation
	// In a real implementation, you would parse HTML and extract microdata attributes
	sde.logger.Printf("📊 [StructuredData] Microdata extraction not fully implemented")
}

// extractOpenGraphData extracts Open Graph data
func (sde *StructuredDataExtractor) extractOpenGraphData(htmlContent string, result *ExtractorStructuredDataResult) {
	ogPattern := regexp.MustCompile(`<meta[^>]*property=["']og:([^"']+)["'][^>]*content=["']([^"']+)["'][^>]*>`)
	matches := ogPattern.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) > 2 {
			property := match[1]
			content := match[2]
			result.OpenGraphData[property] = content
		}
	}
}

// extractTwitterCardData extracts Twitter Card data
func (sde *StructuredDataExtractor) extractTwitterCardData(htmlContent string, result *ExtractorStructuredDataResult) {
	twitterPattern := regexp.MustCompile(`<meta[^>]*name=["']twitter:([^"']+)["'][^>]*content=["']([^"']+)["'][^>]*>`)
	matches := twitterPattern.FindAllStringSubmatch(htmlContent, -1)

	for _, match := range matches {
		if len(match) > 2 {
			property := match[1]
			content := match[2]
			result.TwitterCardData[property] = content
		}
	}
}

// extractBusinessInfoFromStructuredData extracts business information from all structured data
func (sde *StructuredDataExtractor) extractBusinessInfoFromStructuredData(result *ExtractorStructuredDataResult) {
	// Extract from Open Graph data
	if title, exists := result.OpenGraphData["title"]; exists && result.BusinessInfo.BusinessName == "" {
		result.BusinessInfo.BusinessName = title
	}

	if description, exists := result.OpenGraphData["description"]; exists && result.BusinessInfo.Description == "" {
		result.BusinessInfo.Description = description
	}

	if siteName, exists := result.OpenGraphData["site_name"]; exists && result.BusinessInfo.BusinessName == "" {
		result.BusinessInfo.BusinessName = siteName
	}

	// Extract from Twitter Card data
	if title, exists := result.TwitterCardData["title"]; exists && result.BusinessInfo.BusinessName == "" {
		result.BusinessInfo.BusinessName = title
	}

	if description, exists := result.TwitterCardData["description"]; exists && result.BusinessInfo.Description == "" {
		result.BusinessInfo.Description = description
	}
}

// calculateExtractionScore calculates the quality score of structured data extraction
func (sde *StructuredDataExtractor) calculateExtractionScore(result *ExtractorStructuredDataResult) float64 {
	score := 0.0

	// Schema.org data score
	if len(result.SchemaOrgData) > 0 {
		score += 0.4
	}

	// Open Graph data score
	if len(result.OpenGraphData) > 0 {
		score += 0.2
	}

	// Twitter Card data score
	if len(result.TwitterCardData) > 0 {
		score += 0.1
	}

	// Business information completeness score
	if result.BusinessInfo.BusinessName != "" {
		score += 0.1
	}
	if result.BusinessInfo.Description != "" {
		score += 0.1
	}
	if len(result.BusinessInfo.Services) > 0 {
		score += 0.05
	}
	if len(result.ProductInfo) > 0 {
		score += 0.05
	}

	// Contact information score
	if result.ContactInfo.Phone != "" || result.ContactInfo.Email != "" || result.ContactInfo.Address != "" {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}

	return score
}
