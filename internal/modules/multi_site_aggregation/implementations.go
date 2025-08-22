package multi_site_aggregation

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// =============================================================================
// Data Extractor Implementation
// =============================================================================

// WebsiteDataExtractor implements DataExtractor interface
type WebsiteDataExtractor struct {
	logger *zap.Logger
}

// NewWebsiteDataExtractor creates a new website data extractor
func NewWebsiteDataExtractor(logger *zap.Logger) *WebsiteDataExtractor {
	return &WebsiteDataExtractor{
		logger: logger,
	}
}

// ExtractData extracts data from a website location
func (e *WebsiteDataExtractor) ExtractData(ctx context.Context, location BusinessLocation) (*SiteData, error) {
	e.logger.Info("Extracting data from website",
		zap.String("location_id", location.ID),
		zap.String("url", location.URL))

	// Simulate data extraction process
	// In a real implementation, this would involve:
	// 1. Making HTTP requests to the website
	// 2. Parsing HTML content
	// 3. Extracting structured data (JSON-LD, microdata, etc.)
	// 4. Scraping contact information, business details, etc.

	extractedData := make(map[string]interface{})
	confidenceScore := 0.0
	extractionMethod := "web_scraping"

	// Simulate different data types based on URL patterns
	if strings.Contains(location.URL, "contact") || strings.Contains(location.URL, "about") {
		extractedData = map[string]interface{}{
			"business_name":  "Sample Business",
			"phone":          "+1-555-123-4567",
			"email":          "contact@samplebusiness.com",
			"address":        "123 Main St, Anytown, ST 12345",
			"business_hours": "Mon-Fri 9AM-5PM",
			"description":    "A sample business for demonstration purposes",
		}
		confidenceScore = 0.85
		extractionMethod = "contact_page_scraping"
	} else if strings.Contains(location.URL, "products") || strings.Contains(location.URL, "services") {
		extractedData = map[string]interface{}{
			"products":     []string{"Product A", "Product B", "Product C"},
			"services":     []string{"Service 1", "Service 2"},
			"categories":   []string{"Technology", "Consulting"},
			"pricing_info": "Contact for pricing",
		}
		confidenceScore = 0.78
		extractionMethod = "product_page_scraping"
	} else {
		// Default extraction for homepage or other pages
		extractedData = map[string]interface{}{
			"business_name":  "Sample Business",
			"tagline":        "Your trusted partner",
			"industry":       "Technology",
			"founded_year":   2020,
			"employee_count": 50,
			"website_title":  "Sample Business - Home",
		}
		confidenceScore = 0.72
		extractionMethod = "general_page_scraping"
	}

	// Add location-specific metadata
	extractedData["region"] = location.Region
	extractedData["language"] = location.Language
	extractedData["country"] = location.Country

	// Calculate data quality score based on extracted data
	dataQuality := e.calculateDataQuality(extractedData)

	siteData := &SiteData{
		ID:               generateID(),
		LocationID:       location.ID,
		BusinessID:       location.BusinessID,
		DataType:         e.determineDataType(extractedData),
		ExtractedData:    extractedData,
		ConfidenceScore:  confidenceScore,
		ExtractionMethod: extractionMethod,
		LastExtracted:    time.Now(),
		DataQuality:      dataQuality,
		IsValid:          true,
		Metadata:         make(map[string]interface{}),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	e.logger.Info("Data extraction completed",
		zap.String("location_id", location.ID),
		zap.Float64("confidence_score", confidenceScore),
		zap.Float64("data_quality", dataQuality),
		zap.String("data_type", siteData.DataType))

	return siteData, nil
}

// calculateDataQuality calculates a quality score for extracted data
func (e *WebsiteDataExtractor) calculateDataQuality(data map[string]interface{}) float64 {
	qualityScore := 0.0
	totalFields := 0

	// Check for essential business information
	essentialFields := []string{"business_name", "phone", "email", "address"}
	for _, field := range essentialFields {
		if value, exists := data[field]; exists && value != "" {
			qualityScore += 0.25
		}
		totalFields++
	}

	// Check for additional valuable information
	additionalFields := []string{"description", "business_hours", "products", "services"}
	for _, field := range additionalFields {
		if value, exists := data[field]; exists && value != "" {
			qualityScore += 0.1
		}
		totalFields++
	}

	// Normalize score to 0-1 range
	if totalFields > 0 {
		qualityScore = math.Min(qualityScore, 1.0)
	}

	return qualityScore
}

// determineDataType determines the type of data extracted
func (e *WebsiteDataExtractor) determineDataType(data map[string]interface{}) string {
	if _, hasContact := data["phone"]; hasContact {
		return "contact_info"
	}
	if _, hasProducts := data["products"]; hasProducts {
		return "product_catalog"
	}
	if _, hasServices := data["services"]; hasServices {
		return "service_catalog"
	}
	return "business_details"
}

// =============================================================================
// Data Validator Implementation
// =============================================================================

// WebsiteDataValidator implements DataValidator interface
type WebsiteDataValidator struct {
	logger *zap.Logger
}

// NewWebsiteDataValidator creates a new website data validator
func NewWebsiteDataValidator(logger *zap.Logger) *WebsiteDataValidator {
	return &WebsiteDataValidator{
		logger: logger,
	}
}

// ValidateData validates extracted website data
func (v *WebsiteDataValidator) ValidateData(data *SiteData) (bool, []string, error) {
	var errors []string

	// Validate business name
	if businessName, exists := data.ExtractedData["business_name"]; exists {
		if name, ok := businessName.(string); !ok || strings.TrimSpace(name) == "" {
			errors = append(errors, "business_name is missing or invalid")
		}
	} else {
		errors = append(errors, "business_name is required")
	}

	// Validate phone number format
	if phone, exists := data.ExtractedData["phone"]; exists {
		if phoneStr, ok := phone.(string); ok && phoneStr != "" {
			if !v.isValidPhoneNumber(phoneStr) {
				errors = append(errors, "phone number format is invalid")
			}
		}
	}

	// Validate email format
	if email, exists := data.ExtractedData["email"]; exists {
		if emailStr, ok := email.(string); ok && emailStr != "" {
			if !v.isValidEmail(emailStr) {
				errors = append(errors, "email format is invalid")
			}
		}
	}

	// Validate address
	if address, exists := data.ExtractedData["address"]; exists {
		if addrStr, ok := address.(string); !ok || strings.TrimSpace(addrStr) == "" {
			errors = append(errors, "address is missing or invalid")
		}
	}

	// Validate confidence score
	if data.ConfidenceScore < 0.0 || data.ConfidenceScore > 1.0 {
		errors = append(errors, "confidence_score must be between 0 and 1")
	}

	// Validate data quality
	if data.DataQuality < 0.0 || data.DataQuality > 1.0 {
		errors = append(errors, "data_quality must be between 0 and 1")
	}

	// Check for minimum required fields
	requiredFields := []string{"business_name"}
	missingFields := 0
	for _, field := range requiredFields {
		if _, exists := data.ExtractedData[field]; !exists {
			missingFields++
		}
	}

	if missingFields > 0 {
		errors = append(errors, fmt.Sprintf("missing %d required fields", missingFields))
	}

	isValid := len(errors) == 0

	v.logger.Info("Data validation completed",
		zap.String("location_id", data.LocationID),
		zap.Bool("is_valid", isValid),
		zap.Int("error_count", len(errors)))

	return isValid, errors, nil
}

// isValidPhoneNumber validates phone number format
func (v *WebsiteDataValidator) isValidPhoneNumber(phone string) bool {
	// Remove common separators
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "(", "")
	phone = strings.ReplaceAll(phone, ")", "")
	phone = strings.ReplaceAll(phone, ".", "")

	// Check if it starts with + and has at least 10 digits
	if strings.HasPrefix(phone, "+") {
		phone = phone[1:]
	}

	// Count digits
	digitCount := 0
	for _, char := range phone {
		if char >= '0' && char <= '9' {
			digitCount++
		}
	}

	return digitCount >= 10 && digitCount <= 15
}

// isValidEmail validates email format
func (v *WebsiteDataValidator) isValidEmail(email string) bool {
	// Simple email validation
	if !strings.Contains(email, "@") {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	localPart := parts[0]
	domainPart := parts[1]

	if len(localPart) == 0 || len(domainPart) == 0 {
		return false
	}

	if !strings.Contains(domainPart, ".") {
		return false
	}

	return true
}

// =============================================================================
// Data Aggregator Implementation
// =============================================================================

// WebsiteDataAggregator implements DataAggregator interface
type WebsiteDataAggregator struct {
	logger *zap.Logger
}

// NewWebsiteDataAggregator creates a new website data aggregator
func NewWebsiteDataAggregator(logger *zap.Logger) *WebsiteDataAggregator {
	return &WebsiteDataAggregator{
		logger: logger,
	}
}

// AggregateData aggregates data from multiple sites
func (a *WebsiteDataAggregator) AggregateData(ctx context.Context, sitesData []SiteData) (*AggregatedBusinessData, error) {
	a.logger.Info("Starting data aggregation",
		zap.Int("sites_count", len(sitesData)))

	if len(sitesData) == 0 {
		return nil, fmt.Errorf("no site data to aggregate")
	}

	// Group data by type
	dataByType := make(map[string][]SiteData)
	for _, siteData := range sitesData {
		dataType := siteData.DataType
		dataByType[dataType] = append(dataByType[dataType], siteData)
	}

	// Aggregate data by type
	aggregatedData := make(map[string]interface{})
	consistencyIssues := []DataConsistencyIssue{}

	// Aggregate contact information
	if contactData, exists := dataByType["contact_info"]; exists {
		contactInfo, issues := a.aggregateContactInfo(contactData)
		aggregatedData["contact_info"] = contactInfo
		consistencyIssues = append(consistencyIssues, issues...)
	}

	// Aggregate business details
	if businessData, exists := dataByType["business_details"]; exists {
		businessDetails, issues := a.aggregateBusinessDetails(businessData)
		aggregatedData["business_details"] = businessDetails
		consistencyIssues = append(consistencyIssues, issues...)
	}

	// Aggregate product catalog
	if productData, exists := dataByType["product_catalog"]; exists {
		productCatalog, issues := a.aggregateProductCatalog(productData)
		aggregatedData["product_catalog"] = productCatalog
		consistencyIssues = append(consistencyIssues, issues...)
	}

	// Aggregate service catalog
	if serviceData, exists := dataByType["service_catalog"]; exists {
		serviceCatalog, issues := a.aggregateServiceCatalog(serviceData)
		aggregatedData["service_catalog"] = serviceCatalog
		consistencyIssues = append(consistencyIssues, issues...)
	}

	// Calculate scores
	consistencyScore := a.calculateConsistencyScore(sitesData, consistencyIssues)
	completenessScore := a.calculateCompletenessScore(sitesData)
	qualityScore := a.calculateQualityScore(sitesData)

	// Create site data map
	siteDataMap := make(map[string][]SiteData)
	for _, siteData := range sitesData {
		siteDataMap[siteData.LocationID] = append(siteDataMap[siteData.LocationID], siteData)
	}

	aggregatedBusinessData := &AggregatedBusinessData{
		ID:                    generateID(),
		AggregatedData:        aggregatedData,
		DataConsistencyScore:  consistencyScore,
		DataCompletenessScore: completenessScore,
		DataQualityScore:      qualityScore,
		SiteDataMap:           siteDataMap,
		ConsistencyIssues:     consistencyIssues,
		AggregationMethod:     "weighted_average",
		LastAggregated:        time.Now(),
		Metadata:              make(map[string]interface{}),
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	a.logger.Info("Data aggregation completed",
		zap.Float64("consistency_score", consistencyScore),
		zap.Float64("completeness_score", completenessScore),
		zap.Float64("quality_score", qualityScore),
		zap.Int("consistency_issues", len(consistencyIssues)))

	return aggregatedBusinessData, nil
}

// aggregateContactInfo aggregates contact information from multiple sites
func (a *WebsiteDataAggregator) aggregateContactInfo(sitesData []SiteData) (map[string]interface{}, []DataConsistencyIssue) {
	aggregated := make(map[string]interface{})
	var issues []DataConsistencyIssue

	// Collect all values for each field
	fieldValues := make(map[string][]interface{})
	for _, siteData := range sitesData {
		for field, value := range siteData.ExtractedData {
			if value != "" && value != nil {
				fieldValues[field] = append(fieldValues[field], value)
			}
		}
	}

	// Aggregate each field
	for field, values := range fieldValues {
		if len(values) == 0 {
			continue
		}

		// Check for consistency
		if !a.areValuesConsistent(values) {
			issue := DataConsistencyIssue{
				ID:                generateID(),
				FieldName:         field,
				DataType:          "contact_info",
				IssueType:         "conflict",
				Severity:          "medium",
				Description:       fmt.Sprintf("Conflicting values found for field: %s", field),
				AffectedSites:     a.getAffectedSiteIDs(sitesData, field),
				ConflictingValues: a.getConflictingValues(values),
				Recommendation:    "Review and resolve conflicting values",
				CreatedAt:         time.Now(),
			}
			issues = append(issues, issue)
		}

		// Use the most common value or the highest confidence value
		aggregated[field] = a.selectBestValue(values, sitesData)
	}

	return aggregated, issues
}

// aggregateBusinessDetails aggregates business details from multiple sites
func (a *WebsiteDataAggregator) aggregateBusinessDetails(sitesData []SiteData) (map[string]interface{}, []DataConsistencyIssue) {
	aggregated := make(map[string]interface{})
	var issues []DataConsistencyIssue

	// Similar logic to contact info aggregation
	fieldValues := make(map[string][]interface{})
	for _, siteData := range sitesData {
		for field, value := range siteData.ExtractedData {
			if value != "" && value != nil {
				fieldValues[field] = append(fieldValues[field], value)
			}
		}
	}

	for field, values := range fieldValues {
		if len(values) == 0 {
			continue
		}

		if !a.areValuesConsistent(values) {
			issue := DataConsistencyIssue{
				ID:                generateID(),
				FieldName:         field,
				DataType:          "business_details",
				IssueType:         "conflict",
				Severity:          "medium",
				Description:       fmt.Sprintf("Conflicting values found for field: %s", field),
				AffectedSites:     a.getAffectedSiteIDs(sitesData, field),
				ConflictingValues: a.getConflictingValues(values),
				Recommendation:    "Review and resolve conflicting values",
				CreatedAt:         time.Now(),
			}
			issues = append(issues, issue)
		}

		aggregated[field] = a.selectBestValue(values, sitesData)
	}

	return aggregated, issues
}

// aggregateProductCatalog aggregates product catalog from multiple sites
func (a *WebsiteDataAggregator) aggregateProductCatalog(sitesData []SiteData) (map[string]interface{}, []DataConsistencyIssue) {
	aggregated := make(map[string]interface{})
	var issues []DataConsistencyIssue

	// Collect all products
	allProducts := make(map[string]bool)
	for _, siteData := range sitesData {
		if products, exists := siteData.ExtractedData["products"]; exists {
			if productList, ok := products.([]string); ok {
				for _, product := range productList {
					allProducts[product] = true
				}
			}
		}
	}

	// Convert to slice
	productList := make([]string, 0, len(allProducts))
	for product := range allProducts {
		productList = append(productList, product)
	}

	// Sort for consistency
	sort.Strings(productList)
	aggregated["products"] = productList

	return aggregated, issues
}

// aggregateServiceCatalog aggregates service catalog from multiple sites
func (a *WebsiteDataAggregator) aggregateServiceCatalog(sitesData []SiteData) (map[string]interface{}, []DataConsistencyIssue) {
	aggregated := make(map[string]interface{})
	var issues []DataConsistencyIssue

	// Collect all services
	allServices := make(map[string]bool)
	for _, siteData := range sitesData {
		if services, exists := siteData.ExtractedData["services"]; exists {
			if serviceList, ok := services.([]string); ok {
				for _, service := range serviceList {
					allServices[service] = true
				}
			}
		}
	}

	// Convert to slice
	serviceList := make([]string, 0, len(allServices))
	for service := range allServices {
		serviceList = append(serviceList, service)
	}

	// Sort for consistency
	sort.Strings(serviceList)
	aggregated["services"] = serviceList

	return aggregated, issues
}

// calculateConsistencyScore calculates the consistency score across all sites
func (a *WebsiteDataAggregator) calculateConsistencyScore(sitesData []SiteData, issues []DataConsistencyIssue) float64 {
	if len(sitesData) == 0 {
		return 0.0
	}

	// Count total fields
	totalFields := 0
	consistentFields := 0

	fieldValues := make(map[string][]interface{})
	for _, siteData := range sitesData {
		for field, value := range siteData.ExtractedData {
			if value != "" && value != nil {
				fieldValues[field] = append(fieldValues[field], value)
			}
		}
	}

	for _, values := range fieldValues {
		totalFields++
		if a.areValuesConsistent(values) {
			consistentFields++
		}
	}

	if totalFields == 0 {
		return 1.0
	}

	return float64(consistentFields) / float64(totalFields)
}

// calculateCompletenessScore calculates the completeness score
func (a *WebsiteDataAggregator) calculateCompletenessScore(sitesData []SiteData) float64 {
	if len(sitesData) == 0 {
		return 0.0
	}

	// Define required fields
	requiredFields := []string{"business_name", "phone", "email", "address"}
	optionalFields := []string{"description", "business_hours", "products", "services"}

	foundFields := 0

	// Check if fields exist in any site
	allFields := make(map[string]bool)
	for _, siteData := range sitesData {
		for field := range siteData.ExtractedData {
			allFields[field] = true
		}
	}

	// Count found fields
	for _, field := range requiredFields {
		if allFields[field] {
			foundFields += 2 // Weight required fields more
		}
	}

	for _, field := range optionalFields {
		if allFields[field] {
			foundFields += 1
		}
	}

	maxPossible := len(requiredFields)*2 + len(optionalFields)
	if maxPossible == 0 {
		return 1.0
	}

	return math.Min(float64(foundFields)/float64(maxPossible), 1.0)
}

// calculateQualityScore calculates the overall quality score
func (a *WebsiteDataAggregator) calculateQualityScore(sitesData []SiteData) float64 {
	if len(sitesData) == 0 {
		return 0.0
	}

	totalQuality := 0.0
	totalConfidence := 0.0

	for _, siteData := range sitesData {
		totalQuality += siteData.DataQuality
		totalConfidence += siteData.ConfidenceScore
	}

	avgQuality := totalQuality / float64(len(sitesData))
	avgConfidence := totalConfidence / float64(len(sitesData))

	// Weight quality and confidence equally
	return (avgQuality + avgConfidence) / 2.0
}

// areValuesConsistent checks if values are consistent across sites
func (a *WebsiteDataAggregator) areValuesConsistent(values []interface{}) bool {
	if len(values) <= 1 {
		return true
	}

	// Convert all values to strings for comparison
	strValues := make([]string, len(values))
	for i, value := range values {
		strValues[i] = fmt.Sprintf("%v", value)
	}

	// Check if all values are the same
	firstValue := strValues[0]
	for _, value := range strValues[1:] {
		if value != firstValue {
			return false
		}
	}

	return true
}

// selectBestValue selects the best value from multiple options
func (a *WebsiteDataAggregator) selectBestValue(values []interface{}, sitesData []SiteData) interface{} {
	if len(values) == 0 {
		return nil
	}

	if len(values) == 1 {
		return values[0]
	}

	// For now, return the first non-empty value
	// In a more sophisticated implementation, this could consider:
	// - Confidence scores
	// - Data quality scores
	// - Source reliability
	// - Data freshness

	for _, value := range values {
		if value != "" && value != nil {
			return value
		}
	}

	return values[0]
}

// getAffectedSiteIDs gets the IDs of sites affected by a field
func (a *WebsiteDataAggregator) getAffectedSiteIDs(sitesData []SiteData, field string) []string {
	var affectedSites []string
	for _, siteData := range sitesData {
		if _, exists := siteData.ExtractedData[field]; exists {
			affectedSites = append(affectedSites, siteData.LocationID)
		}
	}
	return affectedSites
}

// getConflictingValues gets the conflicting values for a field
func (a *WebsiteDataAggregator) getConflictingValues(values []interface{}) map[string]interface{} {
	conflictingValues := make(map[string]interface{})
	for i, value := range values {
		conflictingValues[fmt.Sprintf("value_%d", i)] = value
	}
	return conflictingValues
}
