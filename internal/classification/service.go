package classification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
)

// ClassificationService provides business classification functionality
type ClassificationService struct {
	config       *config.ExternalServicesConfig
	db           database.Database
	logger       *observability.Logger
	metrics      *observability.Metrics
	industryData *IndustryCodeData
}

// NewClassificationService creates a new business classification service
func NewClassificationService(cfg *config.ExternalServicesConfig, db database.Database, logger *observability.Logger, metrics *observability.Metrics) *ClassificationService {
	return &ClassificationService{
		config:       cfg,
		db:           db,
		logger:       logger,
		metrics:      metrics,
		industryData: nil, // Will be loaded separately
	}
}

// NewClassificationServiceWithData creates a new business classification service with industry data
func NewClassificationServiceWithData(cfg *config.ExternalServicesConfig, db database.Database, logger *observability.Logger, metrics *observability.Metrics, industryData *IndustryCodeData) *ClassificationService {
	return &ClassificationService{
		config:       cfg,
		db:           db,
		logger:       logger,
		metrics:      metrics,
		industryData: industryData,
	}
}

// ClassificationRequest represents a business classification request
type ClassificationRequest struct {
	BusinessName       string `json:"business_name" validate:"required"`
	BusinessType       string `json:"business_type,omitempty"`
	Industry           string `json:"industry,omitempty"`
	Description        string `json:"description,omitempty"`
	Keywords           string `json:"keywords,omitempty"`
	RegistrationNumber string `json:"registration_number,omitempty"`
	TaxID              string `json:"tax_id,omitempty"`
}

// ClassificationResponse represents a business classification response
type ClassificationResponse struct {
	BusinessID            string                   `json:"business_id"`
	Classifications       []IndustryClassification `json:"classifications"`
	PrimaryClassification *IndustryClassification  `json:"primary_classification"`
	ConfidenceScore       float64                  `json:"confidence_score"`
	ClassificationMethod  string                   `json:"classification_method"`
	ProcessingTime        time.Duration            `json:"processing_time"`
	RawData               map[string]interface{}   `json:"raw_data,omitempty"`
}

// IndustryClassification represents an industry classification result
type IndustryClassification struct {
	IndustryCode         string   `json:"industry_code"`
	IndustryName         string   `json:"industry_name"`
	ConfidenceScore      float64  `json:"confidence_score"`
	ClassificationMethod string   `json:"classification_method"`
	Keywords             []string `json:"keywords,omitempty"`
	Description          string   `json:"description,omitempty"`
}

// BatchClassificationRequest represents a batch classification request
type BatchClassificationRequest struct {
	Businesses []ClassificationRequest `json:"businesses" validate:"required,min=1,max=100"`
}

// BatchClassificationResponse represents a batch classification response
type BatchClassificationResponse struct {
	Results        []ClassificationResponse `json:"results"`
	TotalProcessed int                      `json:"total_processed"`
	SuccessCount   int                      `json:"success_count"`
	ErrorCount     int                      `json:"error_count"`
	ProcessingTime time.Duration            `json:"processing_time"`
}

// ClassifyBusiness classifies a single business
func (c *ClassificationService) ClassifyBusiness(ctx context.Context, req *ClassificationRequest) (*ClassificationResponse, error) {
	start := time.Now()

	// Log classification start
	c.logger.WithComponent("classification").LogBusinessEvent(ctx, "classification_started", "", map[string]interface{}{
		"business_name": req.BusinessName,
		"business_type": req.BusinessType,
		"industry":      req.Industry,
	})

	// Validate request
	if err := c.validateClassificationRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Perform classification
	classifications, err := c.performClassification(ctx, req)
	if err != nil {
		c.logger.WithComponent("classification").WithError(err).LogBusinessEvent(ctx, "classification_failed", "", map[string]interface{}{
			"business_name": req.BusinessName,
			"error":         err.Error(),
		})
		c.metrics.RecordBusinessClassification("failed", "error")
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	// Determine primary classification
	primaryClassification := c.determinePrimaryClassification(classifications)

	// Calculate overall confidence score
	confidenceScore := c.calculateOverallConfidence(classifications)

	// Generate business ID for tracking
	businessID := c.generateBusinessID(req)

	// Create response
	response := &ClassificationResponse{
		BusinessID:            businessID,
		Classifications:       classifications,
		PrimaryClassification: primaryClassification,
		ConfidenceScore:       confidenceScore,
		ClassificationMethod:  "hybrid", // Using multiple methods
		ProcessingTime:        time.Since(start),
		RawData: map[string]interface{}{
			"request": req,
			"method":  "hybrid_classification",
		},
	}

	// Store classification in database if available
	if c.db != nil {
		if err := c.storeClassification(ctx, businessID, response); err != nil {
			c.logger.WithComponent("classification").WithError(err).LogBusinessEvent(ctx, "classification_storage_failed", businessID, map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Log successful classification
	c.logger.WithComponent("classification").LogBusinessEvent(ctx, "classification_completed", businessID, map[string]interface{}{
		"business_name":         req.BusinessName,
		"primary_industry_code": primaryClassification.IndustryCode,
		"primary_industry_name": primaryClassification.IndustryName,
		"confidence_score":      confidenceScore,
		"processing_time_ms":    time.Since(start).Milliseconds(),
		"total_classifications": len(classifications),
	})

	// Record metrics
	c.metrics.RecordBusinessClassification("success", fmt.Sprintf("%.2f", confidenceScore))

	return response, nil
}

// ClassifyBusinessesBatch classifies multiple businesses in batch
func (c *ClassificationService) ClassifyBusinessesBatch(ctx context.Context, req *BatchClassificationRequest) (*BatchClassificationResponse, error) {
	start := time.Now()

	// Log batch classification start
	c.logger.WithComponent("classification").LogBusinessEvent(ctx, "batch_classification_started", "", map[string]interface{}{
		"total_businesses": len(req.Businesses),
	})

	// Validate batch size
	if len(req.Businesses) > 100 {
		return nil, fmt.Errorf("batch size exceeds maximum limit of 100")
	}

	results := make([]ClassificationResponse, 0, len(req.Businesses))
	successCount := 0
	errorCount := 0

	// Process each business
	for i, businessReq := range req.Businesses {
		result, err := c.ClassifyBusiness(ctx, &businessReq)
		if err != nil {
			errorCount++
			c.logger.WithComponent("classification").WithError(err).LogBusinessEvent(ctx, "batch_item_failed", "", map[string]interface{}{
				"index":         i,
				"business_name": businessReq.BusinessName,
				"error":         err.Error(),
			})
			continue
		}

		results = append(results, *result)
		successCount++
	}

	response := &BatchClassificationResponse{
		Results:        results,
		TotalProcessed: len(req.Businesses),
		SuccessCount:   successCount,
		ErrorCount:     errorCount,
		ProcessingTime: time.Since(start),
	}

	// Log batch classification completion
	c.logger.WithComponent("classification").LogBusinessEvent(ctx, "batch_classification_completed", "", map[string]interface{}{
		"total_processed":    response.TotalProcessed,
		"success_count":      response.SuccessCount,
		"error_count":        response.ErrorCount,
		"processing_time_ms": response.ProcessingTime.Milliseconds(),
	})

	return response, nil
}

// GetClassificationHistory retrieves classification history for a business
func (c *ClassificationService) GetClassificationHistory(ctx context.Context, businessID string, limit, offset int) ([]*database.BusinessClassification, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	classifications, err := c.db.GetBusinessClassificationsByBusinessID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve classification history: %w", err)
	}

	// Apply pagination
	if offset >= len(classifications) {
		return []*database.BusinessClassification{}, nil
	}

	end := offset + limit
	if end > len(classifications) {
		end = len(classifications)
	}

	return classifications[offset:end], nil
}

// validateClassificationRequest validates the classification request
func (c *ClassificationService) validateClassificationRequest(req *ClassificationRequest) error {
	if strings.TrimSpace(req.BusinessName) == "" {
		return fmt.Errorf("business name is required")
	}

	if len(req.BusinessName) > 500 {
		return fmt.Errorf("business name too long (max 500 characters)")
	}

	if req.Description != "" && len(req.Description) > 2000 {
		return fmt.Errorf("description too long (max 2000 characters)")
	}

	return nil
}

// performClassification performs the actual classification using multiple methods
func (c *ClassificationService) performClassification(ctx context.Context, req *ClassificationRequest) ([]IndustryClassification, error) {
	var classifications []IndustryClassification

	// Method 1: Keyword-based classification
	if keywordClassifications := c.classifyByKeywords(req); len(keywordClassifications) > 0 {
		classifications = append(classifications, keywordClassifications...)
	}

	// Method 2: Business type classification
	if businessTypeClassifications := c.classifyByBusinessType(req); len(businessTypeClassifications) > 0 {
		classifications = append(classifications, businessTypeClassifications...)
	}

	// Method 3: Industry-based classification
	if industryClassifications := c.classifyByIndustry(req); len(industryClassifications) > 0 {
		classifications = append(classifications, industryClassifications...)
	}

	// Method 4: Name-based classification
	if nameClassifications := c.classifyByName(req); len(nameClassifications) > 0 {
		classifications = append(classifications, nameClassifications...)
	}

	// If no classifications found, return default
	if len(classifications) == 0 {
		classifications = append(classifications, c.getDefaultClassification())
	}

	return classifications, nil
}

// classifyByKeywords classifies business based on keywords
func (c *ClassificationService) classifyByKeywords(req *ClassificationRequest) []IndustryClassification {
	var classifications []IndustryClassification

	// Check keywords in business name and description
	textToSearch := strings.ToLower(req.BusinessName + " " + req.Description + " " + req.Keywords)

	// Use real industry data if available
	if c.industryData != nil {
		// Search NAICS codes by keywords
		naicsCodes := c.industryData.SearchNAICSByKeyword(textToSearch)
		for _, code := range naicsCodes {
			classifications = append(classifications, IndustryClassification{
				IndustryCode:         code,
				IndustryName:         c.industryData.GetNAICSName(code),
				ConfidenceScore:      0.7,
				ClassificationMethod: "keyword_based_naics",
				Keywords:             []string{textToSearch},
			})
		}

		// Search MCC codes by keywords
		mccCodes := c.industryData.SearchMCCByKeyword(textToSearch)
		for _, code := range mccCodes {
			classifications = append(classifications, IndustryClassification{
				IndustryCode:         code,
				IndustryName:         c.industryData.GetMCCDescription(code),
				ConfidenceScore:      0.6,
				ClassificationMethod: "keyword_based_mcc",
				Keywords:             []string{textToSearch},
			})
		}

		// Search SIC codes by keywords
		sicCodes := c.industryData.SearchSICByKeyword(textToSearch)
		for _, code := range sicCodes {
			classifications = append(classifications, IndustryClassification{
				IndustryCode:         code,
				IndustryName:         c.industryData.GetSICDescription(code),
				ConfidenceScore:      0.5,
				ClassificationMethod: "keyword_based_sic",
				Keywords:             []string{textToSearch},
			})
		}
	} else {
		// Fallback to hardcoded mappings if no industry data available
		keywordMappings := map[string][]string{
			"software":       {"541511", "541512", "541519"},
			"technology":     {"541511", "541512", "541519", "541715"},
			"consulting":     {"541611", "541612", "541618", "541690"},
			"financial":      {"522110", "522120", "522130", "522190", "523150"},
			"healthcare":     {"621111", "621112", "621210", "621310", "621320"},
			"retail":         {"441110", "442110", "443141", "444110", "445110"},
			"manufacturing":  {"332996", "332999", "333415", "334110", "335110"},
			"construction":   {"236115", "236116", "236117", "236118", "236220"},
			"transportation": {"484110", "484121", "484122", "484210", "485110"},
			"education":      {"611110", "611210", "611310", "611410", "611420"},
		}

		for keyword, industryCodes := range keywordMappings {
			if strings.Contains(textToSearch, keyword) {
				for _, code := range industryCodes {
					classifications = append(classifications, IndustryClassification{
						IndustryCode:         code,
						IndustryName:         c.getIndustryName(code),
						ConfidenceScore:      0.7,
						ClassificationMethod: "keyword_based",
						Keywords:             []string{keyword},
					})
				}
			}
		}
	}

	return classifications
}

// classifyByBusinessType classifies business based on business type
func (c *ClassificationService) classifyByBusinessType(req *ClassificationRequest) []IndustryClassification {
	if req.BusinessType == "" {
		return nil
	}

	businessTypeMappings := map[string]string{
		"llc":                 "541611", // Management consulting
		"corporation":         "541611", // Management consulting
		"partnership":         "541611", // Management consulting
		"sole_proprietorship": "541611", // Management consulting
		"nonprofit":           "813211", // Grantmaking foundations
		"charity":             "813211", // Grantmaking foundations
		"foundation":          "813211", // Grantmaking foundations
	}

	if code, exists := businessTypeMappings[strings.ToLower(req.BusinessType)]; exists {
		return []IndustryClassification{
			{
				IndustryCode:         code,
				IndustryName:         c.getIndustryName(code),
				ConfidenceScore:      0.8,
				ClassificationMethod: "business_type_based",
				Description:          fmt.Sprintf("Classified based on business type: %s", req.BusinessType),
			},
		}
	}

	return nil
}

// classifyByIndustry classifies business based on provided industry
func (c *ClassificationService) classifyByIndustry(req *ClassificationRequest) []IndustryClassification {
	if req.Industry == "" {
		return nil
	}

	industryMappings := map[string]string{
		"technology":     "541511",
		"software":       "541511",
		"consulting":     "541611",
		"finance":        "522110",
		"healthcare":     "621111",
		"retail":         "441110",
		"manufacturing":  "332996",
		"construction":   "236115",
		"transportation": "484110",
		"education":      "611110",
		"real_estate":    "531110",
		"legal":          "541110",
		"accounting":     "541211",
		"marketing":      "541810",
		"advertising":    "541810",
	}

	if code, exists := industryMappings[strings.ToLower(req.Industry)]; exists {
		return []IndustryClassification{
			{
				IndustryCode:         code,
				IndustryName:         c.getIndustryName(code),
				ConfidenceScore:      0.9,
				ClassificationMethod: "industry_based",
				Description:          fmt.Sprintf("Classified based on industry: %s", req.Industry),
			},
		}
	}

	return nil
}

// classifyByName classifies business based on business name patterns
func (c *ClassificationService) classifyByName(req *ClassificationRequest) []IndustryClassification {
	name := strings.ToLower(req.BusinessName)

	// Define name pattern mappings
	namePatterns := map[string]string{
		"tech":          "541511",
		"software":      "541511",
		"systems":       "541511",
		"consult":       "541611",
		"advisory":      "541611",
		"financial":     "522110",
		"bank":          "522110",
		"credit":        "522110",
		"medical":       "621111",
		"health":        "621111",
		"clinic":        "621111",
		"store":         "441110",
		"shop":          "441110",
		"market":        "441110",
		"factory":       "332996",
		"manufacturing": "332996",
		"build":         "236115",
		"construction":  "236115",
		"transport":     "484110",
		"logistics":     "484110",
		"school":        "611110",
		"university":    "611110",
		"college":       "611110",
		"realty":        "531110",
		"properties":    "531110",
		"law":           "541110",
		"legal":         "541110",
		"accounting":    "541211",
		"cpa":           "541211",
		"marketing":     "541810",
		"advertising":   "541810",
	}

	for pattern, code := range namePatterns {
		if strings.Contains(name, pattern) {
			return []IndustryClassification{
				{
					IndustryCode:         code,
					IndustryName:         c.getIndustryName(code),
					ConfidenceScore:      0.6,
					ClassificationMethod: "name_pattern_based",
					Description:          fmt.Sprintf("Classified based on name pattern: %s", pattern),
				},
			}
		}
	}

	return nil
}

// determinePrimaryClassification determines the primary classification from multiple results
func (c *ClassificationService) determinePrimaryClassification(classifications []IndustryClassification) *IndustryClassification {
	if len(classifications) == 0 {
		return nil
	}

	// Find the classification with the highest confidence score
	var primary *IndustryClassification
	highestConfidence := 0.0

	for i := range classifications {
		if classifications[i].ConfidenceScore > highestConfidence {
			highestConfidence = classifications[i].ConfidenceScore
			primary = &classifications[i]
		}
	}

	return primary
}

// calculateOverallConfidence calculates the overall confidence score
func (c *ClassificationService) calculateOverallConfidence(classifications []IndustryClassification) float64 {
	if len(classifications) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, classification := range classifications {
		totalConfidence += classification.ConfidenceScore
	}

	return totalConfidence / float64(len(classifications))
}

// getIndustryName returns the industry name for a given NAICS code
func (c *ClassificationService) getIndustryName(code string) string {
	// Use real industry data if available
	if c.industryData != nil {
		return c.industryData.GetNAICSName(code)
	}

	// Fallback to hardcoded mappings
	industryNames := map[string]string{
		"541511": "Custom Computer Programming Services",
		"541512": "Computer Systems Design Services",
		"541519": "Other Computer Related Services",
		"541611": "Administrative Management and General Management Consulting Services",
		"541612": "Human Resources Consulting Services",
		"541618": "Other Management Consulting Services",
		"541690": "Other Scientific and Technical Consulting Services",
		"541715": "Research and Development in the Physical, Engineering, and Life Sciences",
		"522110": "Commercial Banking",
		"522120": "Savings Institutions",
		"522130": "Credit Unions",
		"522190": "Other Depository Credit Intermediation",
		"523150": "Securities and Commodity Exchanges",
		"621111": "Offices of Physicians (except Mental Health Specialists)",
		"621112": "Offices of Physicians, Mental Health Specialists",
		"621210": "Offices of Dentists",
		"621310": "Offices of Chiropractors",
		"621320": "Offices of Optometrists",
		"441110": "New Car Dealers",
		"442110": "Furniture Stores",
		"443141": "Household Appliance Stores",
		"444110": "Home Centers",
		"445110": "Supermarkets and Other Grocery (except Convenience) Stores",
		"332996": "Fabricated Pipe and Pipe Fitting Manufacturing",
		"332999": "Miscellaneous Fabricated Metal Product Manufacturing",
		"333415": "Air-Conditioning and Warm Air Heating Equipment and Commercial and Industrial Refrigeration Equipment Manufacturing",
		"334110": "Computer and Peripheral Equipment Manufacturing",
		"335110": "Electric Lamp Bulb and Part Manufacturing",
		"236115": "New Single-Family Housing Construction (except For-Sale Builders)",
		"236116": "New Multifamily Housing Construction (except For-Sale Builders)",
		"236117": "New Housing For-Sale Builders",
		"236118": "Residential Remodelers",
		"236220": "Commercial Building Construction",
		"484110": "General Freight Trucking, Local",
		"484121": "General Freight Trucking, Long-Distance, Truckload",
		"484122": "General Freight Trucking, Long-Distance, Less Than Truckload",
		"484210": "Used Household and Office Goods Moving",
		"485110": "Urban Transit Systems",
		"611110": "Elementary and Secondary Schools",
		"611210": "Junior Colleges",
		"611310": "Colleges, Universities, and Professional Schools",
		"611410": "Business and Secretarial Schools",
		"611420": "Computer Training",
		"531110": "Lessors of Residential Buildings and Dwellings",
		"541110": "Offices of Lawyers",
		"541211": "Offices of Certified Public Accountants",
		"541810": "Advertising Agencies",
		"813211": "Grantmaking Foundations",
	}

	if name, exists := industryNames[code]; exists {
		return name
	}

	return "Unknown Industry"
}

// getDefaultClassification returns a default classification
func (c *ClassificationService) getDefaultClassification() IndustryClassification {
	return IndustryClassification{
		IndustryCode:         "541611",
		IndustryName:         "Administrative Management and General Management Consulting Services",
		ConfidenceScore:      0.3,
		ClassificationMethod: "default",
		Description:          "Default classification applied when no specific classification could be determined",
	}
}

// generateBusinessID generates a unique business ID
func (c *ClassificationService) generateBusinessID(req *ClassificationRequest) string {
	// In a real implementation, this would generate a proper UUID
	// For now, we'll create a simple hash-based ID
	return fmt.Sprintf("business_%d", time.Now().UnixNano())
}

// storeClassification stores the classification result in the database
func (c *ClassificationService) storeClassification(ctx context.Context, businessID string, response *ClassificationResponse) error {
	if response.PrimaryClassification == nil {
		return fmt.Errorf("no primary classification to store")
	}

	classification := &database.BusinessClassification{
		ID:                   fmt.Sprintf("classification_%d", time.Now().UnixNano()),
		BusinessID:           businessID,
		IndustryCode:         response.PrimaryClassification.IndustryCode,
		IndustryName:         response.PrimaryClassification.IndustryName,
		ConfidenceScore:      response.PrimaryClassification.ConfidenceScore,
		ClassificationMethod: response.PrimaryClassification.ClassificationMethod,
		Source:               "internal_classifier",
		RawData:              fmt.Sprintf("%+v", response.RawData),
		CreatedAt:            time.Now(),
	}

	return c.db.CreateBusinessClassification(ctx, classification)
}
