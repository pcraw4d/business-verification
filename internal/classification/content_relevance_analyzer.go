package classification

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// ContentRelevanceAnalyzer analyzes content relevance for business classification
type ContentRelevanceAnalyzer struct {
	logger *log.Logger
}

// CrawlResult represents the result of a smart crawl operation (imported from smart_website_crawler.go)
type CrawlResult struct {
	BaseURL       string             `json:"base_url"`
	PagesAnalyzed []PageAnalysis     `json:"pages_analyzed"`
	TotalPages    int                `json:"total_pages"`
	RelevantPages int                `json:"relevant_pages"`
	Keywords      []string           `json:"keywords"`
	IndustryScore map[string]float64 `json:"industry_score"`
	BusinessInfo  BusinessInfo       `json:"business_info"`
	SiteStructure SiteStructure      `json:"site_structure"`
	CrawlDuration time.Duration      `json:"crawl_duration"`
	Success       bool               `json:"success"`
	Error         string             `json:"error,omitempty"`
}

// PageAnalysis represents analysis of a single page (imported from smart_website_crawler.go)
type PageAnalysis struct {
	URL                string                 `json:"url"`
	Title              string                 `json:"title"`
	PageType           string                 `json:"page_type"`
	RelevanceScore     float64                `json:"relevance_score"`
	ContentQuality     float64                `json:"content_quality"`
	Keywords           []string               `json:"keywords"`
	IndustryIndicators []string               `json:"industry_indicators"`
	BusinessInfo       BusinessInfo           `json:"business_info"`
	MetaTags           map[string]string      `json:"meta_tags"`
	StructuredData     map[string]interface{} `json:"structured_data"`
	ResponseTime       time.Duration          `json:"response_time"`
	StatusCode         int                    `json:"status_code"`
	ContentLength      int                    `json:"content_length"`
	LastModified       time.Time              `json:"last_modified"`
	Priority           int                    `json:"priority"`
}

// BusinessInfo represents extracted business information (imported from smart_website_crawler.go)
type BusinessInfo struct {
	BusinessName  string      `json:"business_name"`
	Description   string      `json:"description"`
	Services      []string    `json:"services"`
	Products      []string    `json:"products"`
	ContactInfo   ContactInfo `json:"contact_info"`
	BusinessHours string      `json:"business_hours"`
	Location      string      `json:"location"`
	Industry      string      `json:"industry"`
	BusinessType  string      `json:"business_type"`
}

// ContactInfo represents contact information (imported from smart_website_crawler.go)
type ContactInfo struct {
	Phone   string            `json:"phone"`
	Email   string            `json:"email"`
	Address string            `json:"address"`
	Website string            `json:"website"`
	Social  map[string]string `json:"social"`
}

// SiteStructure represents the discovered site structure (imported from smart_website_crawler.go)
type SiteStructure struct {
	Homepage        string   `json:"homepage"`
	AboutPages      []string `json:"about_pages"`
	ServicePages    []string `json:"service_pages"`
	ProductPages    []string `json:"product_pages"`
	ContactPages    []string `json:"contact_pages"`
	BlogPages       []string `json:"blog_pages"`
	EcommercePages  []string `json:"ecommerce_pages"`
	OtherPages      []string `json:"other_pages"`
	TotalDiscovered int      `json:"total_discovered"`
}

// RelevanceAnalysisResult represents the result of content relevance analysis
type RelevanceAnalysisResult struct {
	OverallRelevance   float64                     `json:"overall_relevance"`
	PageRelevance      map[string]float64          `json:"page_relevance"`
	ContentRelevance   map[string]ContentRelevance `json:"content_relevance"`
	BusinessIndicators []BusinessIndicator         `json:"business_indicators"`
	IndustrySignals    []IndustrySignal            `json:"industry_signals"`
	ConfidenceScore    float64                     `json:"confidence_score"`
	AnalysisTimestamp  time.Time                   `json:"analysis_timestamp"`
}

// ContentRelevance represents relevance analysis for specific content types
type ContentRelevance struct {
	RelevanceScore float64  `json:"relevance_score"`
	Keywords       []string `json:"keywords"`
	Indicators     []string `json:"indicators"`
	Confidence     float64  `json:"confidence"`
}

// BusinessIndicator represents a business-related indicator found in content
type BusinessIndicator struct {
	Type       string  `json:"type"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	Context    string  `json:"context"`
	Source     string  `json:"source"`
}

// IndustrySignal represents an industry-specific signal
type IndustrySignal struct {
	Industry   string  `json:"industry"`
	Signal     string  `json:"signal"`
	Strength   float64 `json:"strength"`
	Confidence float64 `json:"confidence"`
	Context    string  `json:"context"`
	Source     string  `json:"source"`
}

// NewContentRelevanceAnalyzer creates a new content relevance analyzer
func NewContentRelevanceAnalyzer(logger *log.Logger) *ContentRelevanceAnalyzer {
	return &ContentRelevanceAnalyzer{
		logger: logger,
	}
}

// AnalyzeContentRelevance analyzes the relevance of content for business classification
func (cra *ContentRelevanceAnalyzer) AnalyzeContentRelevance(ctx context.Context, crawlResult *CrawlResult) (*RelevanceAnalysisResult, error) {
	cra.logger.Printf("🔍 [RelevanceAnalyzer] Starting content relevance analysis for %s", crawlResult.BaseURL)

	result := &RelevanceAnalysisResult{
		PageRelevance:      make(map[string]float64),
		ContentRelevance:   make(map[string]ContentRelevance),
		BusinessIndicators: []BusinessIndicator{},
		IndustrySignals:    []IndustrySignal{},
		AnalysisTimestamp:  time.Now(),
	}

	// Analyze each page
	for _, page := range crawlResult.PagesAnalyzed {
		pageRelevance := cra.analyzePageRelevance(page)
		result.PageRelevance[page.URL] = pageRelevance.RelevanceScore

		// Aggregate business indicators
		indicators := cra.extractBusinessIndicators(page)
		result.BusinessIndicators = append(result.BusinessIndicators, indicators...)

		// Aggregate industry signals
		signals := cra.extractIndustrySignals(page)
		result.IndustrySignals = append(result.IndustrySignals, signals...)
	}

	// Calculate overall relevance
	result.OverallRelevance = cra.calculateOverallRelevance(result.PageRelevance)

	// Calculate confidence score
	result.ConfidenceScore = cra.calculateConfidenceScore(result)

	cra.logger.Printf("✅ [RelevanceAnalyzer] Analysis completed - Overall relevance: %.2f, Confidence: %.2f",
		result.OverallRelevance, result.ConfidenceScore)

	return result, nil
}

// analyzePageRelevance analyzes the relevance of a single page
func (cra *ContentRelevanceAnalyzer) analyzePageRelevance(page PageAnalysis) ContentRelevance {
	relevance := ContentRelevance{
		RelevanceScore: 0.0,
		Keywords:       []string{},
		Indicators:     []string{},
		Confidence:     0.0,
	}

	// Base relevance by page type
	baseRelevance := cra.getBaseRelevanceByPageType(page.PageType)
	relevance.RelevanceScore = baseRelevance

	// Adjust based on content quality
	relevance.RelevanceScore *= page.ContentQuality

	// Adjust based on keyword density
	keywordScore := cra.calculateKeywordRelevanceScore(page.Keywords)
	relevance.RelevanceScore += keywordScore * 0.3

	// Adjust based on industry indicators
	industryScore := cra.calculateIndustryIndicatorScore(page.IndustryIndicators)
	relevance.RelevanceScore += industryScore * 0.2

	// Adjust based on business information completeness
	businessInfoScore := cra.calculateBusinessInfoScore(page.BusinessInfo)
	relevance.RelevanceScore += businessInfoScore * 0.2

	// Cap at 1.0
	if relevance.RelevanceScore > 1.0 {
		relevance.RelevanceScore = 1.0
	}

	// Calculate confidence
	relevance.Confidence = cra.calculatePageConfidence(page)

	relevance.Keywords = page.Keywords
	relevance.Indicators = page.IndustryIndicators

	return relevance
}

// getBaseRelevanceByPageType returns base relevance score by page type
func (cra *ContentRelevanceAnalyzer) getBaseRelevanceByPageType(pageType string) float64 {
	switch pageType {
	case "about", "about-us", "company", "mission", "vision":
		return 0.95 // Highest relevance for business information
	case "services", "products":
		return 0.90 // Very high relevance for business offerings
	case "homepage":
		return 0.85 // High relevance for business overview
	case "contact", "contact-us":
		return 0.80 // High relevance for business details
	case "team", "careers", "jobs":
		return 0.70 // Medium-high relevance for business structure
	case "blog", "news":
		return 0.60 // Medium relevance for business insights
	case "support", "help", "faq":
		return 0.50 // Medium relevance for business operations
	case "privacy", "terms", "legal":
		return 0.30 // Low relevance for business classification
	default:
		return 0.50 // Default medium relevance
	}
}

// calculateKeywordRelevanceScore calculates relevance score based on keywords
func (cra *ContentRelevanceAnalyzer) calculateKeywordRelevanceScore(keywords []string) float64 {
	if len(keywords) == 0 {
		return 0.0
	}

	// High-value business keywords
	highValueKeywords := map[string]float64{
		"company": 0.9, "business": 0.9, "services": 0.8, "products": 0.8,
		"industry": 0.8, "solutions": 0.7, "consulting": 0.7, "management": 0.7,
		"technology": 0.7, "software": 0.7, "retail": 0.7, "manufacturing": 0.7,
		"healthcare": 0.7, "finance": 0.7, "education": 0.7, "real estate": 0.7,
	}

	score := 0.0
	for _, keyword := range keywords {
		if value, exists := highValueKeywords[strings.ToLower(keyword)]; exists {
			score += value
		} else {
			score += 0.3 // Default score for other keywords
		}
	}

	// Normalize by number of keywords
	return score / float64(len(keywords))
}

// calculateIndustryIndicatorScore calculates score based on industry indicators
func (cra *ContentRelevanceAnalyzer) calculateIndustryIndicatorScore(indicators []string) float64 {
	if len(indicators) == 0 {
		return 0.0
	}

	// Industry-specific indicator weights
	indicatorWeights := map[string]float64{
		"restaurant": 0.9, "cafe": 0.8, "food": 0.8, "dining": 0.8,
		"software": 0.9, "technology": 0.8, "digital": 0.7, "app": 0.7,
		"medical": 0.9, "healthcare": 0.8, "clinic": 0.8, "hospital": 0.8,
		"legal": 0.9, "law": 0.8, "attorney": 0.8, "lawyer": 0.8,
		"retail": 0.8, "store": 0.7, "shop": 0.7, "ecommerce": 0.7,
		"banking": 0.9, "finance": 0.8, "investment": 0.8, "insurance": 0.8,
		"construction": 0.8, "building": 0.7, "architecture": 0.7,
		"school": 0.8, "education": 0.8, "university": 0.8, "training": 0.7,
	}

	score := 0.0
	for _, indicator := range indicators {
		if weight, exists := indicatorWeights[strings.ToLower(indicator)]; exists {
			score += weight
		} else {
			score += 0.4 // Default weight for other indicators
		}
	}

	// Normalize by number of indicators
	return score / float64(len(indicators))
}

// calculateBusinessInfoScore calculates score based on business information completeness
func (cra *ContentRelevanceAnalyzer) calculateBusinessInfoScore(businessInfo BusinessInfo) float64 {
	score := 0.0

	// Business name (most important)
	if businessInfo.BusinessName != "" {
		score += 0.3
	}

	// Description
	if businessInfo.Description != "" {
		score += 0.2
	}

	// Services/Products
	if len(businessInfo.Services) > 0 {
		score += 0.2
	}
	if len(businessInfo.Products) > 0 {
		score += 0.2
	}

	// Contact information
	if businessInfo.ContactInfo.Phone != "" {
		score += 0.05
	}
	if businessInfo.ContactInfo.Email != "" {
		score += 0.05
	}
	if businessInfo.ContactInfo.Address != "" {
		score += 0.05
	}

	// Industry/Business type
	if businessInfo.Industry != "" {
		score += 0.1
	}
	if businessInfo.BusinessType != "" {
		score += 0.1
	}

	return score
}

// calculatePageConfidence calculates confidence score for a page analysis
func (cra *ContentRelevanceAnalyzer) calculatePageConfidence(page PageAnalysis) float64 {
	confidence := 0.5 // Base confidence

	// Content quality factor
	confidence += page.ContentQuality * 0.2

	// Response time factor (faster is better)
	if page.ResponseTime < 2*time.Second {
		confidence += 0.1
	} else if page.ResponseTime < 5*time.Second {
		confidence += 0.05
	}

	// Content length factor
	if page.ContentLength > 1000 {
		confidence += 0.1
	} else if page.ContentLength > 500 {
		confidence += 0.05
	}

	// Keyword density factor
	if len(page.Keywords) > 10 {
		confidence += 0.1
	} else if len(page.Keywords) > 5 {
		confidence += 0.05
	}

	// Industry indicators factor
	if len(page.IndustryIndicators) > 5 {
		confidence += 0.1
	} else if len(page.IndustryIndicators) > 2 {
		confidence += 0.05
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// extractBusinessIndicators extracts business indicators from page content
func (cra *ContentRelevanceAnalyzer) extractBusinessIndicators(page PageAnalysis) []BusinessIndicator {
	var indicators []BusinessIndicator

	// Extract from title
	if page.Title != "" {
		indicators = append(indicators, BusinessIndicator{
			Type:       "title",
			Value:      page.Title,
			Confidence: 0.9,
			Context:    "page_title",
			Source:     page.URL,
		})
	}

	// Extract from meta description
	if desc, exists := page.MetaTags["description"]; exists {
		indicators = append(indicators, BusinessIndicator{
			Type:       "description",
			Value:      desc,
			Confidence: 0.8,
			Context:    "meta_description",
			Source:     page.URL,
		})
	}

	// Extract from business info
	if page.BusinessInfo.BusinessName != "" {
		indicators = append(indicators, BusinessIndicator{
			Type:       "business_name",
			Value:      page.BusinessInfo.BusinessName,
			Confidence: 0.95,
			Context:    "business_info",
			Source:     page.URL,
		})
	}

	// Extract services
	for _, service := range page.BusinessInfo.Services {
		indicators = append(indicators, BusinessIndicator{
			Type:       "service",
			Value:      service,
			Confidence: 0.8,
			Context:    "services",
			Source:     page.URL,
		})
	}

	// Extract products
	for _, product := range page.BusinessInfo.Products {
		indicators = append(indicators, BusinessIndicator{
			Type:       "product",
			Value:      product,
			Confidence: 0.8,
			Context:    "products",
			Source:     page.URL,
		})
	}

	return indicators
}

// extractIndustrySignals extracts industry signals from page content
func (cra *ContentRelevanceAnalyzer) extractIndustrySignals(page PageAnalysis) []IndustrySignal {
	var signals []IndustrySignal

	// Industry-specific signal patterns
	industryPatterns := map[string][]string{
		"food_beverage": {"restaurant", "cafe", "food", "dining", "kitchen", "catering", "bakery", "bar", "pub", "brewery", "winery", "wine", "beer"},
		"technology":    {"software", "technology", "tech", "app", "digital", "web", "mobile", "cloud", "ai", "ml", "data", "cyber", "security", "programming"},
		"healthcare":    {"healthcare", "medical", "clinic", "hospital", "doctor", "dentist", "therapy", "wellness", "pharmacy", "medicine", "patient"},
		"legal":         {"legal", "law", "attorney", "lawyer", "court", "litigation", "patent", "trademark", "copyright", "legal services"},
		"retail":        {"retail", "store", "shop", "ecommerce", "online", "fashion", "clothing", "electronics", "beauty", "products", "merchandise"},
		"finance":       {"finance", "banking", "investment", "insurance", "accounting", "tax", "financial", "credit", "loan", "money", "capital"},
		"real_estate":   {"real estate", "property", "construction", "building", "architecture", "design", "interior", "home", "house", "apartment"},
		"education":     {"education", "school", "university", "training", "learning", "course", "academy", "institute", "student", "teacher"},
		"consulting":    {"consulting", "advisory", "strategy", "management", "business", "corporate", "professional", "services", "expert", "specialist"},
		"manufacturing": {"manufacturing", "production", "factory", "industrial", "automotive", "machinery", "equipment", "assembly"},
	}

	// Check keywords against industry patterns
	for industry, patterns := range industryPatterns {
		strength := 0.0
		matches := 0

		for _, keyword := range page.Keywords {
			for _, pattern := range patterns {
				if strings.Contains(strings.ToLower(keyword), pattern) {
					strength += 1.0
					matches++
					break
				}
			}
		}

		if matches > 0 {
			// Normalize strength
			strength = strength / float64(len(page.Keywords))

			// Calculate confidence based on number of matches and page relevance
			confidence := float64(matches) / float64(len(patterns)) * page.RelevanceScore

			signals = append(signals, IndustrySignal{
				Industry:   industry,
				Signal:     fmt.Sprintf("%d keyword matches", matches),
				Strength:   strength,
				Confidence: confidence,
				Context:    page.PageType,
				Source:     page.URL,
			})
		}
	}

	return signals
}

// calculateOverallRelevance calculates overall relevance from page relevances
func (cra *ContentRelevanceAnalyzer) calculateOverallRelevance(pageRelevances map[string]float64) float64 {
	if len(pageRelevances) == 0 {
		return 0.0
	}

	totalRelevance := 0.0
	count := 0

	for _, relevance := range pageRelevances {
		totalRelevance += relevance
		count++
	}

	return totalRelevance / float64(count)
}

// calculateConfidenceScore calculates overall confidence score
func (cra *ContentRelevanceAnalyzer) calculateConfidenceScore(result *RelevanceAnalysisResult) float64 {
	confidence := 0.5 // Base confidence

	// Factor in number of pages analyzed
	if len(result.PageRelevance) > 5 {
		confidence += 0.2
	} else if len(result.PageRelevance) > 2 {
		confidence += 0.1
	}

	// Factor in business indicators
	if len(result.BusinessIndicators) > 10 {
		confidence += 0.2
	} else if len(result.BusinessIndicators) > 5 {
		confidence += 0.1
	}

	// Factor in industry signals
	if len(result.IndustrySignals) > 5 {
		confidence += 0.1
	} else if len(result.IndustrySignals) > 2 {
		confidence += 0.05
	}

	// Factor in overall relevance
	confidence += result.OverallRelevance * 0.2

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}
