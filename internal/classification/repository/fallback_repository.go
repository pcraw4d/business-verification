package repository

import (
	"context"
	"log"
	"strings"
)

// FallbackKeywordRepository provides a fallback implementation when database is not available
type FallbackKeywordRepository struct {
	logger *log.Logger
}

// NewFallbackKeywordRepository creates a new fallback repository
func NewFallbackKeywordRepository(logger *log.Logger) *FallbackKeywordRepository {
	if logger == nil {
		logger = log.Default()
	}
	return &FallbackKeywordRepository{
		logger: logger,
	}
}

// Industry Management
func (r *FallbackKeywordRepository) GetIndustryByID(ctx context.Context, id int) (*Industry, error) {
	r.logger.Printf("📋 Using fallback industry data for ID: %d", id)

	industries := map[int]*Industry{
		1: {ID: 1, Name: "Technology", Category: "traditional", IsActive: true},
		2: {ID: 2, Name: "Retail", Category: "traditional", IsActive: true},
		3: {ID: 3, Name: "Food Services", Category: "traditional", IsActive: true},
		4: {ID: 4, Name: "Manufacturing", Category: "traditional", IsActive: true},
		5: {ID: 5, Name: "Healthcare", Category: "traditional", IsActive: true},
		6: {ID: 6, Name: "Finance", Category: "traditional", IsActive: true},
		7: {ID: 7, Name: "Agriculture", Category: "traditional", IsActive: true},
		8: {ID: 8, Name: "Construction", Category: "traditional", IsActive: true},
	}

	if industry, exists := industries[id]; exists {
		return industry, nil
	}
	return nil, nil
}

func (r *FallbackKeywordRepository) GetIndustryByName(ctx context.Context, name string) (*Industry, error) {
	r.logger.Printf("📋 Using fallback industry data for name: %s", name)

	industries := []*Industry{
		{ID: 1, Name: "Technology", Category: "traditional", IsActive: true},
		{ID: 2, Name: "Retail", Category: "traditional", IsActive: true},
		{ID: 3, Name: "Food Services", Category: "traditional", IsActive: true},
		{ID: 4, Name: "Manufacturing", Category: "traditional", IsActive: true},
		{ID: 5, Name: "Healthcare", Category: "traditional", IsActive: true},
		{ID: 6, Name: "Finance", Category: "traditional", IsActive: true},
		{ID: 7, Name: "Agriculture", Category: "traditional", IsActive: true},
		{ID: 8, Name: "Construction", Category: "traditional", IsActive: true},
	}

	for _, industry := range industries {
		if strings.EqualFold(industry.Name, name) {
			return industry, nil
		}
	}
	return nil, nil
}

func (r *FallbackKeywordRepository) ListIndustries(ctx context.Context, category string) ([]*Industry, error) {
	r.logger.Printf("📋 Using fallback industry list for category: %s", category)

	allIndustries := []*Industry{
		{ID: 1, Name: "Technology", Category: "traditional", IsActive: true},
		{ID: 2, Name: "Retail", Category: "traditional", IsActive: true},
		{ID: 3, Name: "Food Services", Category: "traditional", IsActive: true},
		{ID: 4, Name: "Manufacturing", Category: "traditional", IsActive: true},
		{ID: 5, Name: "Healthcare", Category: "traditional", IsActive: true},
		{ID: 6, Name: "Finance", Category: "traditional", IsActive: true},
		{ID: 7, Name: "Agriculture", Category: "traditional", IsActive: true},
		{ID: 8, Name: "Construction", Category: "traditional", IsActive: true},
	}

	if category == "" {
		return allIndustries, nil
	}

	var filtered []*Industry
	for _, industry := range allIndustries {
		if strings.EqualFold(industry.Category, category) {
			filtered = append(filtered, industry)
		}
	}

	return filtered, nil
}

func (r *FallbackKeywordRepository) CreateIndustry(ctx context.Context, industry *Industry) error {
	r.logger.Printf("📋 Fallback: Create industry not implemented")
	return nil
}

func (r *FallbackKeywordRepository) UpdateIndustry(ctx context.Context, industry *Industry) error {
	r.logger.Printf("📋 Fallback: Update industry not implemented")
	return nil
}

func (r *FallbackKeywordRepository) DeleteIndustry(ctx context.Context, id int) error {
	r.logger.Printf("📋 Fallback: Delete industry not implemented")
	return nil
}

// Keyword Management
func (r *FallbackKeywordRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*IndustryKeyword, error) {
	r.logger.Printf("📋 Using fallback keyword data for industry ID: %d", industryID)

	keywords := map[int][]*IndustryKeyword{
		1: { // Technology
			{ID: 1, IndustryID: 1, Keyword: "software", Weight: 0.95, IsActive: true},
			{ID: 2, IndustryID: 1, Keyword: "technology", Weight: 0.90, IsActive: true},
			{ID: 3, IndustryID: 1, Keyword: "computer", Weight: 0.85, IsActive: true},
			{ID: 4, IndustryID: 1, Keyword: "digital", Weight: 0.80, IsActive: true},
			{ID: 5, IndustryID: 1, Keyword: "tech", Weight: 0.75, IsActive: true},
			{ID: 6, IndustryID: 1, Keyword: "app", Weight: 0.70, IsActive: true},
			{ID: 7, IndustryID: 1, Keyword: "development", Weight: 0.85, IsActive: true},
			{ID: 8, IndustryID: 1, Keyword: "programming", Weight: 0.80, IsActive: true},
		},
		2: { // Retail
			{ID: 9, IndustryID: 2, Keyword: "retail", Weight: 0.95, IsActive: true},
			{ID: 10, IndustryID: 2, Keyword: "store", Weight: 0.90, IsActive: true},
			{ID: 11, IndustryID: 2, Keyword: "shop", Weight: 0.85, IsActive: true},
			{ID: 12, IndustryID: 2, Keyword: "merchandise", Weight: 0.80, IsActive: true},
			{ID: 13, IndustryID: 2, Keyword: "sales", Weight: 0.75, IsActive: true},
			{ID: 14, IndustryID: 2, Keyword: "grocery", Weight: 0.70, IsActive: true},
		},
		3: { // Food Services
			{ID: 15, IndustryID: 3, Keyword: "restaurant", Weight: 0.95, IsActive: true},
			{ID: 16, IndustryID: 3, Keyword: "food", Weight: 0.90, IsActive: true},
			{ID: 17, IndustryID: 3, Keyword: "cafe", Weight: 0.85, IsActive: true},
			{ID: 18, IndustryID: 3, Keyword: "dining", Weight: 0.80, IsActive: true},
			{ID: 19, IndustryID: 3, Keyword: "catering", Weight: 0.75, IsActive: true},
			{ID: 20, IndustryID: 3, Keyword: "beverage", Weight: 0.70, IsActive: true},
		},
		4: { // Manufacturing
			{ID: 21, IndustryID: 4, Keyword: "manufacturing", Weight: 0.95, IsActive: true},
			{ID: 22, IndustryID: 4, Keyword: "production", Weight: 0.90, IsActive: true},
			{ID: 23, IndustryID: 4, Keyword: "factory", Weight: 0.85, IsActive: true},
			{ID: 24, IndustryID: 4, Keyword: "industrial", Weight: 0.80, IsActive: true},
			{ID: 25, IndustryID: 4, Keyword: "assembly", Weight: 0.75, IsActive: true},
		},
		5: { // Healthcare
			{ID: 26, IndustryID: 5, Keyword: "healthcare", Weight: 0.95, IsActive: true},
			{ID: 27, IndustryID: 5, Keyword: "medical", Weight: 0.90, IsActive: true},
			{ID: 28, IndustryID: 5, Keyword: "hospital", Weight: 0.85, IsActive: true},
			{ID: 29, IndustryID: 5, Keyword: "clinic", Weight: 0.80, IsActive: true},
			{ID: 30, IndustryID: 5, Keyword: "pharmacy", Weight: 0.75, IsActive: true},
		},
		6: { // Finance
			{ID: 31, IndustryID: 6, Keyword: "finance", Weight: 0.95, IsActive: true},
			{ID: 32, IndustryID: 6, Keyword: "banking", Weight: 0.90, IsActive: true},
			{ID: 33, IndustryID: 6, Keyword: "financial", Weight: 0.85, IsActive: true},
			{ID: 34, IndustryID: 6, Keyword: "investment", Weight: 0.80, IsActive: true},
			{ID: 35, IndustryID: 6, Keyword: "credit", Weight: 0.75, IsActive: true},
		},
		7: { // Agriculture
			{ID: 36, IndustryID: 7, Keyword: "agriculture", Weight: 0.95, IsActive: true},
			{ID: 37, IndustryID: 7, Keyword: "farming", Weight: 0.90, IsActive: true},
			{ID: 38, IndustryID: 7, Keyword: "crop", Weight: 0.85, IsActive: true},
			{ID: 39, IndustryID: 7, Keyword: "livestock", Weight: 0.80, IsActive: true},
		},
		8: { // Construction
			{ID: 40, IndustryID: 8, Keyword: "construction", Weight: 0.95, IsActive: true},
			{ID: 41, IndustryID: 8, Keyword: "building", Weight: 0.90, IsActive: true},
			{ID: 42, IndustryID: 8, Keyword: "contractor", Weight: 0.85, IsActive: true},
			{ID: 43, IndustryID: 8, Keyword: "contracting", Weight: 0.80, IsActive: true},
		},
	}

	if industryKeywords, exists := keywords[industryID]; exists {
		return industryKeywords, nil
	}

	return []*IndustryKeyword{}, nil
}

func (r *FallbackKeywordRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*IndustryKeyword, error) {
	r.logger.Printf("🔍 Performing fallback keyword search for: %s", query)

	query = strings.ToLower(query)

	// Get all industries and their keywords
	industries := []int{1, 2, 3, 4, 5, 6, 7, 8}
	var results []*IndustryKeyword

	for _, industryID := range industries {
		keywords, err := r.GetKeywordsByIndustry(ctx, industryID)
		if err != nil {
			continue
		}

		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(keyword.Keyword), query) {
				results = append(results, keyword)
				if len(results) >= limit {
					return results, nil
				}
			}
		}
	}

	return results, nil
}

func (r *FallbackKeywordRepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	r.logger.Printf("📋 Fallback: Add keyword not implemented")
	return nil
}

func (r *FallbackKeywordRepository) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	r.logger.Printf("📋 Fallback: Update keyword weight not implemented")
	return nil
}

func (r *FallbackKeywordRepository) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	r.logger.Printf("📋 Fallback: Remove keyword not implemented")
	return nil
}

// Classification Codes
func (r *FallbackKeywordRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*ClassificationCode, error) {
	r.logger.Printf("📋 Using fallback classification codes for industry ID: %d", industryID)

	codes := map[int][]*ClassificationCode{
		1: { // Technology
			{ID: 1, IndustryID: 1, CodeType: "NAICS", Code: "541511", Description: "Custom Computer Programming Services", IsActive: true},
			{ID: 2, IndustryID: 1, CodeType: "MCC", Code: "5734", Description: "Computer Software Stores", IsActive: true},
			{ID: 3, IndustryID: 1, CodeType: "SIC", Code: "7371", Description: "Computer Programming Services", IsActive: true},
		},
		2: { // Retail
			{ID: 4, IndustryID: 2, CodeType: "NAICS", Code: "445110", Description: "Supermarkets and Other Grocery (except Convenience) Stores", IsActive: true},
			{ID: 5, IndustryID: 2, CodeType: "MCC", Code: "5411", Description: "Grocery Stores, Supermarkets", IsActive: true},
			{ID: 6, IndustryID: 2, CodeType: "SIC", Code: "5411", Description: "Grocery Stores", IsActive: true},
		},
		3: { // Food Services
			{ID: 7, IndustryID: 3, CodeType: "NAICS", Code: "722511", Description: "Full-Service Restaurants", IsActive: true},
			{ID: 8, IndustryID: 3, CodeType: "MCC", Code: "5812", Description: "Eating Places, Restaurants", IsActive: true},
			{ID: 9, IndustryID: 3, CodeType: "SIC", Code: "5812", Description: "Eating Places", IsActive: true},
		},
		4: { // Manufacturing
			{ID: 10, IndustryID: 4, CodeType: "NAICS", Code: "311111", Description: "Dog and Cat Food Manufacturing", IsActive: true},
			{ID: 11, IndustryID: 4, CodeType: "MCC", Code: "5045", Description: "Computers, Computer Peripheral Equipment, Software", IsActive: true},
			{ID: 12, IndustryID: 4, CodeType: "SIC", Code: "3089", Description: "Plastics Products, Not Elsewhere Classified", IsActive: true},
		},
		5: { // Healthcare
			{ID: 13, IndustryID: 5, CodeType: "NAICS", Code: "621111", Description: "Offices of Physicians (except Mental Health Specialists)", IsActive: true},
			{ID: 14, IndustryID: 5, CodeType: "MCC", Code: "8062", Description: "Hospitals", IsActive: true},
			{ID: 15, IndustryID: 5, CodeType: "SIC", Code: "8011", Description: "Offices and Clinics of Doctors of Medicine", IsActive: true},
		},
		6: { // Finance
			{ID: 16, IndustryID: 6, CodeType: "NAICS", Code: "522110", Description: "Commercial Banking", IsActive: true},
			{ID: 17, IndustryID: 6, CodeType: "MCC", Code: "6011", Description: "Automated Teller Machine Services", IsActive: true},
			{ID: 18, IndustryID: 6, CodeType: "SIC", Code: "6021", Description: "National Commercial Banks", IsActive: true},
		},
		7: { // Agriculture
			{ID: 19, IndustryID: 7, CodeType: "NAICS", Code: "111110", Description: "Soybean Farming", IsActive: true},
			{ID: 20, IndustryID: 7, CodeType: "MCC", Code: "0763", Description: "Agricultural Cooperative", IsActive: true},
			{ID: 21, IndustryID: 7, CodeType: "SIC", Code: "0111", Description: "Wheat", IsActive: true},
		},
		8: { // Construction
			{ID: 22, IndustryID: 8, CodeType: "NAICS", Code: "236116", Description: "New Multifamily Housing Construction (except For-Sale Builders)", IsActive: true},
			{ID: 23, IndustryID: 8, CodeType: "MCC", Code: "1520", Description: "General Contractors - Residential and Commercial", IsActive: true},
			{ID: 24, IndustryID: 8, CodeType: "SIC", Code: "1521", Description: "General Contractors - Single-Family Houses", IsActive: true},
		},
	}

	if industryCodes, exists := codes[industryID]; exists {
		return industryCodes, nil
	}

	return []*ClassificationCode{}, nil
}

func (r *FallbackKeywordRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*ClassificationCode, error) {
	r.logger.Printf("📋 Fallback: Get classification codes by type not implemented")
	return []*ClassificationCode{}, nil
}

func (r *FallbackKeywordRepository) AddClassificationCode(ctx context.Context, code *ClassificationCode) error {
	r.logger.Printf("📋 Fallback: Add classification code not implemented")
	return nil
}

func (r *FallbackKeywordRepository) UpdateClassificationCode(ctx context.Context, code *ClassificationCode) error {
	r.logger.Printf("📋 Fallback: Update classification code not implemented")
	return nil
}

func (r *FallbackKeywordRepository) DeleteClassificationCode(ctx context.Context, id int) error {
	r.logger.Printf("📋 Fallback: Delete classification code not implemented")
	return nil
}

// Industry Patterns
func (r *FallbackKeywordRepository) GetPatternsByIndustry(ctx context.Context, industryID int) ([]*IndustryPattern, error) {
	r.logger.Printf("📋 Fallback: Get patterns by industry not implemented")
	return []*IndustryPattern{}, nil
}

func (r *FallbackKeywordRepository) AddPattern(ctx context.Context, pattern *IndustryPattern) error {
	r.logger.Printf("📋 Fallback: Add pattern not implemented")
	return nil
}

func (r *FallbackKeywordRepository) UpdatePattern(ctx context.Context, pattern *IndustryPattern) error {
	r.logger.Printf("📋 Fallback: Update pattern not implemented")
	return nil
}

func (r *FallbackKeywordRepository) DeletePattern(ctx context.Context, id int) error {
	r.logger.Printf("📋 Fallback: Delete pattern not implemented")
	return nil
}

// Keyword Weights
func (r *FallbackKeywordRepository) GetKeywordWeights(ctx context.Context, keyword string) ([]*KeywordWeight, error) {
	r.logger.Printf("📋 Fallback: Get keyword weights not implemented")
	return []*KeywordWeight{}, nil
}

func (r *FallbackKeywordRepository) UpdateKeywordWeightByID(ctx context.Context, weight *KeywordWeight) error {
	r.logger.Printf("📋 Fallback: Update keyword weight by ID not implemented")
	return nil
}

func (r *FallbackKeywordRepository) IncrementUsageCount(ctx context.Context, keyword string, industryID int) error {
	r.logger.Printf("📋 Fallback: Increment usage count not implemented")
	return nil
}

// Business Classification
func (r *FallbackKeywordRepository) ClassifyBusiness(ctx context.Context, businessName, description, websiteURL string) (*ClassificationResult, error) {
	r.logger.Printf("🔍 Performing fallback business classification for: %s", businessName)

	// Simple keyword-based classification
	text := strings.ToLower(businessName + " " + description)

	// Check for technology keywords
	if strings.Contains(text, "software") || strings.Contains(text, "tech") || strings.Contains(text, "computer") {
		industry, _ := r.GetIndustryByID(ctx, 1)
		return &ClassificationResult{
			Industry:   industry,
			Confidence: 0.85,
			Keywords:   []string{"software", "technology"},
			Reasoning:  "Technology keywords detected in business name/description",
		}, nil
	}

	// Check for retail keywords
	if strings.Contains(text, "store") || strings.Contains(text, "retail") || strings.Contains(text, "shop") {
		industry, _ := r.GetIndustryByID(ctx, 2)
		return &ClassificationResult{
			Industry:   industry,
			Confidence: 0.80,
			Keywords:   []string{"store", "retail"},
			Reasoning:  "Retail keywords detected in business name/description",
		}, nil
	}

	// Check for food service keywords
	if strings.Contains(text, "restaurant") || strings.Contains(text, "food") || strings.Contains(text, "cafe") {
		industry, _ := r.GetIndustryByID(ctx, 3)
		return &ClassificationResult{
			Industry:   industry,
			Confidence: 0.85,
			Keywords:   []string{"restaurant", "food"},
			Reasoning:  "Food service keywords detected in business name/description",
		}, nil
	}

	// Default to technology if no clear match
	industry, _ := r.GetIndustryByID(ctx, 1)
	return &ClassificationResult{
		Industry:   industry,
		Confidence: 0.50,
		Keywords:   []string{"general"},
		Reasoning:  "No specific industry keywords detected, defaulting to technology",
	}, nil
}

func (r *FallbackKeywordRepository) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*ClassificationResult, error) {
	r.logger.Printf("🔍 Performing fallback business classification by keywords")

	// Simple keyword matching
	for _, keyword := range keywords {
		keyword = strings.ToLower(keyword)

		if keyword == "software" || keyword == "tech" || keyword == "computer" {
			industry, _ := r.GetIndustryByID(ctx, 1)
			return &ClassificationResult{
				Industry:   industry,
				Confidence: 0.85,
				Keywords:   []string{keyword},
				Reasoning:  "Technology keyword detected",
			}, nil
		}

		if keyword == "store" || keyword == "retail" || keyword == "shop" {
			industry, _ := r.GetIndustryByID(ctx, 2)
			return &ClassificationResult{
				Industry:   industry,
				Confidence: 0.80,
				Keywords:   []string{keyword},
				Reasoning:  "Retail keyword detected",
			}, nil
		}
	}

	// Default to technology
	industry, _ := r.GetIndustryByID(ctx, 1)
	return &ClassificationResult{
		Industry:   industry,
		Confidence: 0.50,
		Keywords:   keywords,
		Reasoning:  "No specific industry keywords detected, defaulting to technology",
	}, nil
}

func (r *FallbackKeywordRepository) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*Industry, error) {
	r.logger.Printf("📋 Fallback: Get top industries by keywords not implemented")
	return []*Industry{}, nil
}

// Advanced Search and Analytics
func (r *FallbackKeywordRepository) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*Industry, error) {
	r.logger.Printf("📋 Fallback: Search industries by pattern not implemented")
	return []*Industry{}, nil
}

func (r *FallbackKeywordRepository) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	r.logger.Printf("📋 Fallback: Get industry statistics not implemented")
	return map[string]interface{}{}, nil
}

func (r *FallbackKeywordRepository) GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error) {
	r.logger.Printf("📋 Fallback: Get keyword frequency not implemented")
	return map[string]int{}, nil
}

// Bulk Operations
func (r *FallbackKeywordRepository) BulkInsertKeywords(ctx context.Context, keywords []*IndustryKeyword) error {
	r.logger.Printf("📋 Fallback: Bulk insert keywords not implemented")
	return nil
}

func (r *FallbackKeywordRepository) BulkUpdateKeywords(ctx context.Context, keywords []*IndustryKeyword) error {
	r.logger.Printf("📋 Fallback: Bulk update keywords not implemented")
	return nil
}

func (r *FallbackKeywordRepository) BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error {
	r.logger.Printf("📋 Fallback: Bulk delete keywords not implemented")
	return nil
}

// Health and Maintenance
func (r *FallbackKeywordRepository) Ping(ctx context.Context) error {
	r.logger.Printf("✅ Fallback repository health check passed")
	return nil
}

func (r *FallbackKeywordRepository) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	r.logger.Printf("📋 Fallback: Get database stats not implemented")
	return map[string]interface{}{}, nil
}

func (r *FallbackKeywordRepository) CleanupInactiveData(ctx context.Context) error {
	r.logger.Printf("📋 Fallback: Cleanup inactive data not implemented")
	return nil
}
