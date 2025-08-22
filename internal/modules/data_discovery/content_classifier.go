package data_discovery

import (
	"context"
	"regexp"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ContentClassifier classifies content to understand business context and type
type ContentClassifier struct {
	config            *DataDiscoveryConfig
	logger            *zap.Logger
	businessTypeRules []ClassificationRule
	industryRules     []ClassificationRule
	contentTypeRules  []ClassificationRule
	qualityRules      []ClassificationRule
}

// ClassificationRule defines a rule for content classification
type ClassificationRule struct {
	RuleID          string   `json:"rule_id"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	Keywords        []string `json:"keywords"`
	Patterns        []string `json:"patterns"`
	WeightPositive  float64  `json:"weight_positive"`
	WeightNegative  float64  `json:"weight_negative"`
	RequiredMatches int      `json:"required_matches"`
	Priority        int      `json:"priority"`
}

// NewContentClassifier creates a new content classifier
func NewContentClassifier(config *DataDiscoveryConfig, logger *zap.Logger) *ContentClassifier {
	return &ContentClassifier{
		config:            config,
		logger:            logger,
		businessTypeRules: getBusinessTypeRules(),
		industryRules:     getIndustryRules(),
		contentTypeRules:  getContentTypeRules(),
		qualityRules:      getQualityRules(),
	}
}

// ClassifyContent classifies the provided content
func (cc *ContentClassifier) ClassifyContent(ctx context.Context, content *ContentInput) (*ClassificationResult, error) {
	startTime := time.Now()

	cc.logger.Debug("Starting content classification",
		zap.String("content_type", content.ContentType),
		zap.Int("content_length", len(content.RawContent)))

	result := &ClassificationResult{
		ContentCategories:   []string{},
		TechnicalIndicators: []string{},
		Metadata:            make(map[string]interface{}),
	}

	// Prepare content for analysis
	normalizedContent := cc.normalizeContent(content.RawContent)

	// Classify business type
	businessType, businessConfidence := cc.classifyBusinessType(normalizedContent)
	result.BusinessType = businessType

	// Classify industry
	industry, industryConfidence := cc.classifyIndustry(normalizedContent)
	result.IndustryCategory = industry

	// Classify content categories
	categories := cc.classifyContentCategories(normalizedContent)
	result.ContentCategories = categories

	// Assess content quality
	qualityScore := cc.assessContentQuality(content, normalizedContent)
	result.QualityScore = qualityScore

	// Detect technical indicators
	technicalIndicators := cc.detectTechnicalIndicators(content)
	result.TechnicalIndicators = technicalIndicators

	// Calculate overall confidence
	result.ConfidenceScore = cc.calculateOverallConfidence(
		businessConfidence, industryConfidence, qualityScore, len(categories))

	// Add metadata
	result.Metadata["processing_time"] = time.Since(startTime)
	result.Metadata["business_confidence"] = businessConfidence
	result.Metadata["industry_confidence"] = industryConfidence
	result.Metadata["categories_detected"] = len(categories)
	result.Metadata["technical_indicators"] = len(technicalIndicators)

	cc.logger.Debug("Content classification completed",
		zap.String("business_type", result.BusinessType),
		zap.String("industry", result.IndustryCategory),
		zap.Float64("confidence", result.ConfidenceScore),
		zap.Duration("processing_time", time.Since(startTime)))

	return result, nil
}

// normalizeContent normalizes content for analysis
func (cc *ContentClassifier) normalizeContent(content string) string {
	// Convert to lowercase
	normalized := strings.ToLower(content)

	// Remove extra whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	normalized = spaceRegex.ReplaceAllString(normalized, " ")

	// Remove HTML tags if present
	htmlRegex := regexp.MustCompile(`<[^>]*>`)
	normalized = htmlRegex.ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

// classifyBusinessType classifies the business type based on content
func (cc *ContentClassifier) classifyBusinessType(content string) (string, float64) {
	scores := make(map[string]float64)

	for _, rule := range cc.businessTypeRules {
		score := cc.evaluateRule(content, rule)
		if score > 0 {
			scores[rule.Category] = scores[rule.Category] + score
		}
	}

	// Find the highest scoring business type
	maxScore := 0.0
	bestType := "unknown"

	for businessType, score := range scores {
		if score > maxScore {
			maxScore = score
			bestType = businessType
		}
	}

	// Normalize confidence score
	confidence := maxScore
	if confidence > 1.0 {
		confidence = 1.0
	}

	return bestType, confidence
}

// classifyIndustry classifies the industry based on content
func (cc *ContentClassifier) classifyIndustry(content string) (string, float64) {
	scores := make(map[string]float64)

	for _, rule := range cc.industryRules {
		score := cc.evaluateRule(content, rule)
		if score > 0 {
			scores[rule.Category] = scores[rule.Category] + score
		}
	}

	// Find the highest scoring industry
	maxScore := 0.0
	bestIndustry := "unknown"

	for industry, score := range scores {
		if score > maxScore {
			maxScore = score
			bestIndustry = industry
		}
	}

	// Normalize confidence score
	confidence := maxScore
	if confidence > 1.0 {
		confidence = 1.0
	}

	return bestIndustry, confidence
}

// classifyContentCategories classifies content into multiple categories
func (cc *ContentClassifier) classifyContentCategories(content string) []string {
	var categories []string
	categoryScores := make(map[string]float64)

	for _, rule := range cc.contentTypeRules {
		score := cc.evaluateRule(content, rule)
		if score >= 0.3 { // Threshold for category inclusion
			categoryScores[rule.Category] = score
		}
	}

	// Sort categories by score and return top ones
	for category := range categoryScores {
		categories = append(categories, category)
	}

	return categories
}

// assessContentQuality assesses the overall quality of the content
func (cc *ContentClassifier) assessContentQuality(content *ContentInput, normalizedContent string) float64 {
	qualityScore := 0.5 // Base quality score

	// Content length factor
	contentLength := len(content.RawContent)
	if contentLength > 1000 && contentLength < 50000 {
		qualityScore += 0.2
	}

	// HTML structure quality (if HTML content available)
	if content.HTMLContent != "" {
		htmlQuality := cc.assessHTMLQuality(content.HTMLContent)
		qualityScore += htmlQuality * 0.2
	}

	// Structured data presence
	if content.StructuredData != nil && len(content.StructuredData) > 0 {
		qualityScore += 0.15
	}

	// Metadata quality
	if content.MetaData != nil && len(content.MetaData) > 0 {
		qualityScore += 0.1
	}

	// Language detection
	if content.Language != "" && content.Language != "unknown" {
		qualityScore += 0.05
	}

	// Apply quality rules
	for _, rule := range cc.qualityRules {
		ruleScore := cc.evaluateRule(normalizedContent, rule)
		qualityScore += ruleScore * 0.1
	}

	// Ensure score is within bounds
	if qualityScore > 1.0 {
		qualityScore = 1.0
	} else if qualityScore < 0.0 {
		qualityScore = 0.0
	}

	return qualityScore
}

// assessHTMLQuality assesses the quality of HTML structure
func (cc *ContentClassifier) assessHTMLQuality(htmlContent string) float64 {
	quality := 0.5

	// Check for semantic HTML elements
	semanticElements := []string{"<header>", "<nav>", "<main>", "<section>", "<article>", "<aside>", "<footer>"}
	for _, element := range semanticElements {
		if strings.Contains(htmlContent, element) {
			quality += 0.05
		}
	}

	// Check for proper meta tags
	metaTags := []string{"<title>", "<meta name=\"description\"", "<meta name=\"keywords\""}
	for _, tag := range metaTags {
		if strings.Contains(htmlContent, tag) {
			quality += 0.05
		}
	}

	// Check for accessibility features
	accessibilityFeatures := []string{"alt=", "aria-", "role="}
	for _, feature := range accessibilityFeatures {
		if strings.Contains(htmlContent, feature) {
			quality += 0.03
		}
	}

	return quality
}

// detectTechnicalIndicators detects technical aspects of the website
func (cc *ContentClassifier) detectTechnicalIndicators(content *ContentInput) []string {
	var indicators []string

	if content.HTMLContent != "" {
		// Detect CMS/Framework indicators
		cmsIndicators := map[string]string{
			"wp-content":  "WordPress",
			"drupal":      "Drupal",
			"joomla":      "Joomla",
			"shopify":     "Shopify",
			"wix.com":     "Wix",
			"squarespace": "Squarespace",
			"react":       "React",
			"angular":     "Angular",
			"vue":         "Vue.js",
			"bootstrap":   "Bootstrap",
			"jquery":      "jQuery",
		}

		htmlLower := strings.ToLower(content.HTMLContent)
		for indicator, technology := range cmsIndicators {
			if strings.Contains(htmlLower, indicator) {
				indicators = append(indicators, technology)
			}
		}

		// Detect analytics and tracking
		if strings.Contains(htmlLower, "google-analytics") || strings.Contains(htmlLower, "gtag") {
			indicators = append(indicators, "Google Analytics")
		}
		if strings.Contains(htmlLower, "facebook.net") {
			indicators = append(indicators, "Facebook Pixel")
		}
		if strings.Contains(htmlLower, "hotjar") {
			indicators = append(indicators, "Hotjar")
		}

		// Detect SSL/Security indicators
		if content.URL != "" && strings.HasPrefix(content.URL, "https://") {
			indicators = append(indicators, "HTTPS")
		}

		// Detect responsive design
		if strings.Contains(htmlLower, "viewport") || strings.Contains(htmlLower, "responsive") {
			indicators = append(indicators, "Responsive Design")
		}
	}

	return indicators
}

// evaluateRule evaluates a classification rule against content
func (cc *ContentClassifier) evaluateRule(content string, rule ClassificationRule) float64 {
	score := 0.0
	matches := 0

	// Check keywords
	for _, keyword := range rule.Keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			score += rule.WeightPositive
			matches++
		}
	}

	// Check patterns
	for _, pattern := range rule.Patterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			score += rule.WeightPositive
			matches++
		}
	}

	// Apply required matches threshold
	if matches < rule.RequiredMatches {
		score *= 0.5 // Reduce score if not enough matches
	}

	return score
}

// calculateOverallConfidence calculates the overall confidence score
func (cc *ContentClassifier) calculateOverallConfidence(businessConf, industryConf, qualityScore float64, categoriesCount int) float64 {
	// Weight different factors
	confidence := (businessConf * 0.3) + (industryConf * 0.3) + (qualityScore * 0.3)

	// Bonus for multiple categories detected
	categoryBonus := float64(categoriesCount) * 0.02
	if categoryBonus > 0.1 {
		categoryBonus = 0.1
	}
	confidence += categoryBonus

	// Ensure within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// getBusinessTypeRules returns rules for business type classification
func getBusinessTypeRules() []ClassificationRule {
	return []ClassificationRule{
		{
			RuleID:          "b2b_software",
			Name:            "B2B Software",
			Category:        "b2b_software",
			Keywords:        []string{"enterprise", "business", "solution", "platform", "api", "integration", "saas", "software"},
			WeightPositive:  0.3,
			RequiredMatches: 2,
			Priority:        1,
		},
		{
			RuleID:          "ecommerce",
			Name:            "E-commerce",
			Category:        "ecommerce",
			Keywords:        []string{"shop", "buy", "cart", "checkout", "product", "price", "store", "shipping"},
			WeightPositive:  0.4,
			RequiredMatches: 2,
			Priority:        1,
		},
		{
			RuleID:          "consulting",
			Name:            "Consulting",
			Category:        "consulting",
			Keywords:        []string{"consulting", "advisory", "expertise", "strategy", "professional services"},
			WeightPositive:  0.4,
			RequiredMatches: 1,
			Priority:        1,
		},
		{
			RuleID:          "healthcare",
			Name:            "Healthcare",
			Category:        "healthcare",
			Keywords:        []string{"medical", "health", "doctor", "clinic", "patient", "treatment", "therapy"},
			WeightPositive:  0.4,
			RequiredMatches: 1,
			Priority:        1,
		},
		{
			RuleID:          "financial",
			Name:            "Financial Services",
			Category:        "financial",
			Keywords:        []string{"financial", "bank", "investment", "loan", "insurance", "credit", "money"},
			WeightPositive:  0.4,
			RequiredMatches: 1,
			Priority:        1,
		},
	}
}

// getIndustryRules returns rules for industry classification
func getIndustryRules() []ClassificationRule {
	return []ClassificationRule{
		{
			RuleID:          "technology",
			Name:            "Technology",
			Category:        "technology",
			Keywords:        []string{"technology", "software", "development", "programming", "digital", "innovation"},
			WeightPositive:  0.3,
			RequiredMatches: 1,
			Priority:        1,
		},
		{
			RuleID:          "manufacturing",
			Name:            "Manufacturing",
			Category:        "manufacturing",
			Keywords:        []string{"manufacturing", "production", "factory", "industrial", "equipment", "machinery"},
			WeightPositive:  0.4,
			RequiredMatches: 1,
			Priority:        1,
		},
		{
			RuleID:          "retail",
			Name:            "Retail",
			Category:        "retail",
			Keywords:        []string{"retail", "store", "shopping", "merchandise", "customer", "sales"},
			WeightPositive:  0.3,
			RequiredMatches: 1,
			Priority:        1,
		},
		{
			RuleID:          "education",
			Name:            "Education",
			Category:        "education",
			Keywords:        []string{"education", "school", "university", "learning", "course", "training", "academic"},
			WeightPositive:  0.4,
			RequiredMatches: 1,
			Priority:        1,
		},
	}
}

// getContentTypeRules returns rules for content type classification
func getContentTypeRules() []ClassificationRule {
	return []ClassificationRule{
		{
			RuleID:          "corporate_website",
			Name:            "Corporate Website",
			Category:        "corporate",
			Keywords:        []string{"about us", "company", "mission", "values", "team", "leadership"},
			WeightPositive:  0.3,
			RequiredMatches: 1,
			Priority:        1,
		},
		{
			RuleID:          "product_showcase",
			Name:            "Product Showcase",
			Category:        "product",
			Keywords:        []string{"features", "benefits", "demo", "trial", "download", "pricing"},
			WeightPositive:  0.3,
			RequiredMatches: 1,
			Priority:        1,
		},
		{
			RuleID:          "service_provider",
			Name:            "Service Provider",
			Category:        "service",
			Keywords:        []string{"services", "solutions", "expertise", "consulting", "support"},
			WeightPositive:  0.3,
			RequiredMatches: 1,
			Priority:        1,
		},
	}
}

// getQualityRules returns rules for content quality assessment
func getQualityRules() []ClassificationRule {
	return []ClassificationRule{
		{
			RuleID:          "professional_content",
			Name:            "Professional Content",
			Category:        "professional",
			Keywords:        []string{"professional", "certified", "licensed", "accredited", "experienced"},
			WeightPositive:  0.2,
			RequiredMatches: 1,
			Priority:        1,
		},
		{
			RuleID:          "contact_information",
			Name:            "Contact Information",
			Category:        "contact",
			Keywords:        []string{"contact", "phone", "email", "address", "location"},
			WeightPositive:  0.3,
			RequiredMatches: 2,
			Priority:        1,
		},
	}
}
