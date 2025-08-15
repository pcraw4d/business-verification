package webanalysis

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"

	"golang.org/x/net/html"
)

// EnhancedContentAnalysis represents comprehensive content analysis results
type EnhancedContentAnalysis struct {
	MetaTags           *MetaTagAnalysis          `json:"meta_tags"`
	StructuredData     *StructuredDataAnalysis   `json:"structured_data"`
	SemanticAnalysis   *SemanticContentAnalysis  `json:"semantic_analysis"`
	ContentQuality     *ContentQualityAssessment `json:"content_quality"`
	IndustryIndicators []string                  `json:"industry_indicators"`
	BusinessKeywords   []string                  `json:"business_keywords"`
	ConfidenceScore    float64                   `json:"confidence_score"`
	AnalyzedAt         time.Time                 `json:"analyzed_at"`
}

// MetaTagAnalysis represents analysis of HTML meta tags
type MetaTagAnalysis struct {
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Keywords    []string          `json:"keywords"`
	Author      string            `json:"author"`
	Robots      string            `json:"robots"`
	Language    string            `json:"language"`
	Viewport    string            `json:"viewport"`
	OpenGraph   map[string]string `json:"open_graph"`
	Twitter     map[string]string `json:"twitter"`
	Other       map[string]string `json:"other"`
	Quality     float64           `json:"quality"`
}

// StructuredDataAnalysis represents analysis of structured data
type StructuredDataAnalysis struct {
	JSONLD       []map[string]interface{} `json:"json_ld"`
	Microdata    []map[string]interface{} `json:"microdata"`
	RDFa         []map[string]interface{} `json:"rdfa"`
	SchemaTypes  []string                 `json:"schema_types"`
	BusinessInfo *BusinessStructuredData  `json:"business_info"`
	Quality      float64                  `json:"quality"`
}

// BusinessStructuredData represents business information from structured data
type BusinessStructuredData struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Logo        string   `json:"logo"`
	Address     *Address `json:"address"`
	Contact     *Contact `json:"contact"`
	Industry    string   `json:"industry"`
	Founded     string   `json:"founded"`
	Employees   string   `json:"employees"`
}

// Address represents business address information
type Address struct {
	StreetAddress string `json:"street_address"`
	Locality      string `json:"locality"`
	Region        string `json:"region"`
	PostalCode    string `json:"postal_code"`
	Country       string `json:"country"`
}

// Contact represents business contact information
type Contact struct {
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Website string `json:"website"`
}

// SemanticContentAnalysis represents semantic analysis results
type SemanticContentAnalysis struct {
	IndustryKeywords    []string            `json:"industry_keywords"`
	BusinessDescription string              `json:"business_description"`
	ServiceKeywords     []string            `json:"service_keywords"`
	ProductKeywords     []string            `json:"product_keywords"`
	LocationKeywords    []string            `json:"location_keywords"`
	SemanticScore       float64             `json:"semantic_score"`
	TopicClusters       map[string][]string `json:"topic_clusters"`
	EntityExtraction    map[string][]string `json:"entity_extraction"`
}

// ContentQualityAssessment represents content quality analysis
type ContentQualityAssessment struct {
	ReadabilityScore       float64 `json:"readability_score"`
	ContentLength          int     `json:"content_length"`
	KeywordDensity         float64 `json:"keyword_density"`
	StructureQuality       float64 `json:"structure_quality"`
	RelevanceScore         float64 `json:"relevance_score"`
	OverallQuality         float64 `json:"overall_quality"`
	SpamIndicators         int     `json:"spam_indicators"`
	ProfessionalIndicators int     `json:"professional_indicators"`
}

// EnhancedContentAnalyzer manages enhanced content analysis operations
type EnhancedContentAnalyzer struct {
	config               *ContentAnalyzerConfig
	industryKeywords     map[string][]string
	spamPatterns         []*regexp.Regexp
	professionalPatterns []*regexp.Regexp
}

// ContentAnalyzerConfig holds configuration for content analysis
type ContentAnalyzerConfig struct {
	EnableMetaTagAnalysis       bool    `json:"enable_meta_tag_analysis"`
	EnableStructuredData        bool    `json:"enable_structured_data"`
	EnableSemanticAnalysis      bool    `json:"enable_semantic_analysis"`
	EnableQualityAssessment     bool    `json:"enable_quality_assessment"`
	MinContentLength            int     `json:"min_content_length"`
	MaxKeywordDensity           float64 `json:"max_keyword_density"`
	QualityThreshold            float64 `json:"quality_threshold"`
	SemanticSimilarityThreshold float64 `json:"semantic_similarity_threshold"`
}

// NewEnhancedContentAnalyzer creates a new enhanced content analyzer
func NewEnhancedContentAnalyzer(config *ContentAnalyzerConfig) *EnhancedContentAnalyzer {
	if config == nil {
		config = &ContentAnalyzerConfig{
			EnableMetaTagAnalysis:       true,
			EnableStructuredData:        true,
			EnableSemanticAnalysis:      true,
			EnableQualityAssessment:     true,
			MinContentLength:            100,
			MaxKeywordDensity:           0.05,
			QualityThreshold:            0.6,
			SemanticSimilarityThreshold: 0.7,
		}
	}

	analyzer := &EnhancedContentAnalyzer{
		config:               config,
		industryKeywords:     make(map[string][]string),
		spamPatterns:         []*regexp.Regexp{},
		professionalPatterns: []*regexp.Regexp{},
	}

	analyzer.initializeIndustryKeywords()
	analyzer.initializePatterns()

	return analyzer
}

// AnalyzeContent performs comprehensive content analysis
func (eca *EnhancedContentAnalyzer) AnalyzeContent(content *ScrapedContent, business string) *EnhancedContentAnalysis {
	analysis := &EnhancedContentAnalysis{
		AnalyzedAt: time.Now(),
	}

	// Analyze meta tags
	if eca.config.EnableMetaTagAnalysis {
		analysis.MetaTags = eca.analyzeMetaTags(content.HTML)
	}

	// Analyze structured data
	if eca.config.EnableStructuredData {
		analysis.StructuredData = eca.analyzeStructuredData(content.HTML)
	}

	// Perform semantic analysis
	if eca.config.EnableSemanticAnalysis {
		analysis.SemanticAnalysis = eca.analyzeSemanticContent(content, business)
	}

	// Assess content quality
	if eca.config.EnableQualityAssessment {
		analysis.ContentQuality = eca.assessContentQuality(content, business)
	}

	// Extract industry indicators
	analysis.IndustryIndicators = eca.extractIndustryIndicators(content)

	// Extract business keywords
	analysis.BusinessKeywords = eca.extractBusinessKeywords(content, business)

	// Calculate overall confidence score
	analysis.ConfidenceScore = eca.calculateConfidenceScore(analysis)

	return analysis
}

// analyzeMetaTags analyzes HTML meta tags
func (eca *EnhancedContentAnalyzer) analyzeMetaTags(htmlContent string) *MetaTagAnalysis {
	analysis := &MetaTagAnalysis{
		OpenGraph: make(map[string]string),
		Twitter:   make(map[string]string),
		Other:     make(map[string]string),
	}

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return analysis
	}

	var extractMeta func(*html.Node)
	extractMeta = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			var name, content, property string
			for _, attr := range n.Attr {
				switch attr.Key {
				case "name":
					name = attr.Val
				case "content":
					content = attr.Val
				case "property":
					property = attr.Val
				}
			}

			if content != "" {
				switch {
				case name == "title":
					analysis.Title = content
				case name == "description":
					analysis.Description = content
				case name == "keywords":
					analysis.Keywords = strings.Split(content, ",")
					for i, keyword := range analysis.Keywords {
						analysis.Keywords[i] = strings.TrimSpace(keyword)
					}
				case name == "author":
					analysis.Author = content
				case name == "robots":
					analysis.Robots = content
				case name == "language":
					analysis.Language = content
				case name == "viewport":
					analysis.Viewport = content
				case strings.HasPrefix(property, "og:"):
					analysis.OpenGraph[property] = content
				case strings.HasPrefix(name, "twitter:"):
					analysis.Twitter[name] = content
				default:
					if name != "" {
						analysis.Other[name] = content
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractMeta(c)
		}
	}

	extractMeta(doc)

	// Calculate meta tag quality
	analysis.Quality = eca.calculateMetaTagQuality(analysis)

	return analysis
}

// analyzeStructuredData analyzes structured data in HTML
func (eca *EnhancedContentAnalyzer) analyzeStructuredData(htmlContent string) *StructuredDataAnalysis {
	analysis := &StructuredDataAnalysis{
		JSONLD:    []map[string]interface{}{},
		Microdata: []map[string]interface{}{},
		RDFa:      []map[string]interface{}{},
	}

	// Extract JSON-LD
	jsonLDPattern := regexp.MustCompile(`<script[^>]*type=["']application/ld\+json["'][^>]*>(.*?)</script>`)
	matches := jsonLDPattern.FindAllStringSubmatch(htmlContent, -1)
	for _, match := range matches {
		if len(match) > 1 {
			var data interface{}
			if err := json.Unmarshal([]byte(match[1]), &data); err == nil {
				if mapData, ok := data.(map[string]interface{}); ok {
					analysis.JSONLD = append(analysis.JSONLD, mapData)
				}
			}
		}
	}

	// Extract business information from structured data
	analysis.BusinessInfo = eca.extractBusinessFromStructuredData(analysis.JSONLD)

	// Calculate structured data quality
	analysis.Quality = eca.calculateStructuredDataQuality(analysis)

	return analysis
}

// analyzeSemanticContent performs semantic analysis of content
func (eca *EnhancedContentAnalyzer) analyzeSemanticContent(content *ScrapedContent, business string) *SemanticContentAnalysis {
	analysis := &SemanticContentAnalysis{
		TopicClusters:    make(map[string][]string),
		EntityExtraction: make(map[string][]string),
	}

	// Extract industry-specific keywords
	analysis.IndustryKeywords = eca.extractIndustryKeywords(content.Text)

	// Extract business description
	analysis.BusinessDescription = eca.extractBusinessDescription(content)

	// Extract service and product keywords
	analysis.ServiceKeywords = eca.extractServiceKeywords(content.Text)
	analysis.ProductKeywords = eca.extractProductKeywords(content.Text)

	// Extract location keywords
	analysis.LocationKeywords = eca.extractLocationKeywords(content.Text)

	// Calculate semantic score
	analysis.SemanticScore = eca.calculateSemanticScore(content, business)

	// Create topic clusters
	analysis.TopicClusters = eca.createTopicClusters(content.Text)

	// Extract entities
	analysis.EntityExtraction = eca.extractEntities(content.Text)

	return analysis
}

// assessContentQuality assesses the quality of content
func (eca *EnhancedContentAnalyzer) assessContentQuality(content *ScrapedContent, business string) *ContentQualityAssessment {
	assessment := &ContentQualityAssessment{}

	// Calculate readability score
	assessment.ReadabilityScore = eca.calculateReadabilityScore(content.Text)

	// Content length
	assessment.ContentLength = len(content.Text)

	// Keyword density
	assessment.KeywordDensity = eca.calculateKeywordDensity(content.Text, business)

	// Structure quality
	assessment.StructureQuality = eca.calculateStructureQuality(content.HTML)

	// Relevance score
	assessment.RelevanceScore = eca.calculateRelevanceScore(content, business)

	// Spam indicators
	assessment.SpamIndicators = eca.countSpamIndicators(content.Text)

	// Professional indicators
	assessment.ProfessionalIndicators = eca.countProfessionalIndicators(content.Text)

	// Calculate overall quality
	assessment.OverallQuality = eca.calculateOverallQuality(assessment)

	return assessment
}

// Helper methods for content analysis

func (eca *EnhancedContentAnalyzer) initializeIndustryKeywords() {
	eca.industryKeywords = map[string][]string{
		"technology":    {"software", "hardware", "IT", "digital", "tech", "computing", "programming", "development"},
		"healthcare":    {"medical", "health", "care", "hospital", "clinic", "pharmacy", "wellness", "treatment"},
		"finance":       {"banking", "financial", "investment", "insurance", "credit", "loan", "money", "wealth"},
		"retail":        {"store", "shop", "retail", "commerce", "ecommerce", "merchandise", "sales", "customer"},
		"manufacturing": {"manufacturing", "production", "factory", "industrial", "machinery", "equipment", "assembly"},
		"education":     {"education", "school", "university", "learning", "training", "academic", "teaching"},
		"real_estate":   {"real estate", "property", "housing", "construction", "development", "architecture"},
		"consulting":    {"consulting", "advisory", "strategy", "management", "business", "professional", "expertise"},
	}
}

func (eca *EnhancedContentAnalyzer) initializePatterns() {
	// Spam patterns
	eca.spamPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)click here`),
		regexp.MustCompile(`(?i)buy now`),
		regexp.MustCompile(`(?i)limited time`),
		regexp.MustCompile(`(?i)act now`),
		regexp.MustCompile(`(?i)free offer`),
		regexp.MustCompile(`(?i)make money`),
		regexp.MustCompile(`(?i)work from home`),
	}

	// Professional patterns
	eca.professionalPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)professional`),
		regexp.MustCompile(`(?i)expertise`),
		regexp.MustCompile(`(?i)experience`),
		regexp.MustCompile(`(?i)certified`),
		regexp.MustCompile(`(?i)licensed`),
		regexp.MustCompile(`(?i)accredited`),
		regexp.MustCompile(`(?i)established`),
		regexp.MustCompile(`(?i)since \d{4}`),
	}
}

func (eca *EnhancedContentAnalyzer) calculateMetaTagQuality(analysis *MetaTagAnalysis) float64 {
	score := 0.0

	if analysis.Title != "" {
		score += 0.2
	}
	if analysis.Description != "" {
		score += 0.2
	}
	if len(analysis.Keywords) > 0 {
		score += 0.15
	}
	if len(analysis.OpenGraph) > 0 {
		score += 0.15
	}
	if analysis.Author != "" {
		score += 0.1
	}
	if analysis.Language != "" {
		score += 0.1
	}

	return score
}

func (eca *EnhancedContentAnalyzer) extractBusinessFromStructuredData(jsonLD []map[string]interface{}) *BusinessStructuredData {
	businessInfo := &BusinessStructuredData{}

	for _, data := range jsonLD {
		if schemaType, ok := data["@type"].(string); ok {
			if schemaType == "Organization" || schemaType == "LocalBusiness" || schemaType == "Corporation" {
				if name, ok := data["name"].(string); ok {
					businessInfo.Name = name
				}
				if description, ok := data["description"].(string); ok {
					businessInfo.Description = description
				}
				if url, ok := data["url"].(string); ok {
					businessInfo.URL = url
				}
				if logo, ok := data["logo"].(string); ok {
					businessInfo.Logo = logo
				}
				if founded, ok := data["foundingDate"].(string); ok {
					businessInfo.Founded = founded
				}
				if employees, ok := data["numberOfEmployees"].(string); ok {
					businessInfo.Employees = employees
				}

				// Extract address
				if address, ok := data["address"].(map[string]interface{}); ok {
					businessInfo.Address = &Address{}
					if street, ok := address["streetAddress"].(string); ok {
						businessInfo.Address.StreetAddress = street
					}
					if locality, ok := address["addressLocality"].(string); ok {
						businessInfo.Address.Locality = locality
					}
					if region, ok := address["addressRegion"].(string); ok {
						businessInfo.Address.Region = region
					}
					if postalCode, ok := address["postalCode"].(string); ok {
						businessInfo.Address.PostalCode = postalCode
					}
					if country, ok := address["addressCountry"].(string); ok {
						businessInfo.Address.Country = country
					}
				}

				// Extract contact
				if contact, ok := data["contactPoint"].(map[string]interface{}); ok {
					businessInfo.Contact = &Contact{}
					if phone, ok := contact["telephone"].(string); ok {
						businessInfo.Contact.Phone = phone
					}
					if email, ok := contact["email"].(string); ok {
						businessInfo.Contact.Email = email
					}
				}
			}
		}
	}

	return businessInfo
}

func (eca *EnhancedContentAnalyzer) calculateStructuredDataQuality(analysis *StructuredDataAnalysis) float64 {
	score := 0.0

	if len(analysis.JSONLD) > 0 {
		score += 0.4
	}
	if analysis.BusinessInfo != nil && analysis.BusinessInfo.Name != "" {
		score += 0.3
	}
	if analysis.BusinessInfo != nil && analysis.BusinessInfo.Description != "" {
		score += 0.2
	}
	if analysis.BusinessInfo != nil && analysis.BusinessInfo.Address != nil {
		score += 0.1
	}

	return score
}

func (eca *EnhancedContentAnalyzer) extractIndustryKeywords(text string) []string {
	var keywords []string
	textLower := strings.ToLower(text)

	for industry, industryKeywords := range eca.industryKeywords {
		for _, keyword := range industryKeywords {
			if strings.Contains(textLower, keyword) {
				keywords = append(keywords, keyword)
			}
		}
	}

	return keywords
}

func (eca *EnhancedContentAnalyzer) extractBusinessDescription(content *ScrapedContent) string {
	// Look for business description in meta tags first
	if content.MetaDescription != "" {
		return content.MetaDescription
	}

	// Extract from content using heuristics
	text := content.Text
	sentences := strings.Split(text, ".")

	// Look for sentences that contain business-related keywords
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 20 && len(sentence) < 200 {
			lowerSentence := strings.ToLower(sentence)
			if strings.Contains(lowerSentence, "company") ||
				strings.Contains(lowerSentence, "business") ||
				strings.Contains(lowerSentence, "organization") ||
				strings.Contains(lowerSentence, "specializes") ||
				strings.Contains(lowerSentence, "provides") {
				return sentence
			}
		}
	}

	return ""
}

func (eca *EnhancedContentAnalyzer) extractServiceKeywords(text string) []string {
	serviceKeywords := []string{"service", "services", "solutions", "consulting", "support", "help", "assistance"}
	var found []string
	textLower := strings.ToLower(text)

	for _, keyword := range serviceKeywords {
		if strings.Contains(textLower, keyword) {
			found = append(found, keyword)
		}
	}

	return found
}

func (eca *EnhancedContentAnalyzer) extractProductKeywords(text string) []string {
	productKeywords := []string{"product", "products", "goods", "items", "merchandise", "catalog", "inventory"}
	var found []string
	textLower := strings.ToLower(text)

	for _, keyword := range productKeywords {
		if strings.Contains(textLower, keyword) {
			found = append(found, keyword)
		}
	}

	return found
}

func (eca *EnhancedContentAnalyzer) extractLocationKeywords(text string) []string {
	locationKeywords := []string{"location", "address", "city", "state", "region", "area", "neighborhood"}
	var found []string
	textLower := strings.ToLower(text)

	for _, keyword := range locationKeywords {
		if strings.Contains(textLower, keyword) {
			found = append(found, keyword)
		}
	}

	return found
}

func (eca *EnhancedContentAnalyzer) calculateSemanticScore(content *ScrapedContent, business string) float64 {
	score := 0.0
	text := strings.ToLower(content.Text)
	businessLower := strings.ToLower(business)

	// Business name match
	if strings.Contains(text, businessLower) {
		score += 0.3
	}

	// Industry keyword density
	industryKeywords := eca.extractIndustryKeywords(content.Text)
	score += float64(len(industryKeywords)) * 0.1

	// Business description presence
	if eca.extractBusinessDescription(content) != "" {
		score += 0.2
	}

	// Professional language indicators
	professionalCount := eca.countProfessionalIndicators(content.Text)
	score += float64(professionalCount) * 0.05

	return score
}

func (eca *EnhancedContentAnalyzer) createTopicClusters(text string) map[string][]string {
	clusters := make(map[string][]string)
	textLower := strings.ToLower(text)

	// Business-related clusters
	if strings.Contains(textLower, "company") || strings.Contains(textLower, "business") {
		clusters["business"] = []string{"company", "business", "organization", "enterprise"}
	}

	// Service-related clusters
	if strings.Contains(textLower, "service") || strings.Contains(textLower, "solutions") {
		clusters["services"] = []string{"service", "solutions", "consulting", "support"}
	}

	// Product-related clusters
	if strings.Contains(textLower, "product") || strings.Contains(textLower, "goods") {
		clusters["products"] = []string{"product", "goods", "items", "merchandise"}
	}

	// Contact-related clusters
	if strings.Contains(textLower, "contact") || strings.Contains(textLower, "address") {
		clusters["contact"] = []string{"contact", "address", "phone", "email"}
	}

	return clusters
}

func (eca *EnhancedContentAnalyzer) extractEntities(text string) map[string][]string {
	entities := make(map[string][]string)
	textLower := strings.ToLower(text)

	// Extract company names (simple heuristic)
	companyPattern := regexp.MustCompile(`(?i)([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\s+(?:Inc|LLC|Corp|Company|Ltd|Limited))`)
	matches := companyPattern.FindAllString(text, -1)
	if len(matches) > 0 {
		entities["companies"] = matches
	}

	// Extract locations
	locationPattern := regexp.MustCompile(`(?i)([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*,\s*[A-Z]{2})`)
	locationMatches := locationPattern.FindAllString(text, -1)
	if len(locationMatches) > 0 {
		entities["locations"] = locationMatches
	}

	return entities
}

func (eca *EnhancedContentAnalyzer) calculateReadabilityScore(text string) float64 {
	// Simple Flesch Reading Ease approximation
	words := strings.Fields(text)
	sentences := strings.Split(text, ".")

	if len(words) == 0 || len(sentences) == 0 {
		return 0.0
	}

	// Count syllables (approximation)
	syllables := 0
	for _, word := range words {
		syllables += eca.countSyllables(word)
	}

	// Calculate Flesch score
	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))
	avgSyllablesPerWord := float64(syllables) / float64(len(words))

	score := 206.835 - (1.015 * avgWordsPerSentence) - (84.6 * avgSyllablesPerWord)

	// Normalize to 0-1 range
	if score > 100 {
		score = 100
	} else if score < 0 {
		score = 0
	}

	return score / 100.0
}

func (eca *EnhancedContentAnalyzer) countSyllables(word string) int {
	// Simple syllable counting approximation
	vowels := regexp.MustCompile(`[aeiouy]+`)
	matches := vowels.FindAllString(strings.ToLower(word), -1)
	return len(matches)
}

func (eca *EnhancedContentAnalyzer) calculateKeywordDensity(text string, business string) float64 {
	words := strings.Fields(strings.ToLower(text))
	if len(words) == 0 {
		return 0.0
	}

	businessWords := strings.Fields(strings.ToLower(business))
	keywordCount := 0

	for _, businessWord := range businessWords {
		if len(businessWord) > 2 {
			for _, word := range words {
				if word == businessWord {
					keywordCount++
				}
			}
		}
	}

	return float64(keywordCount) / float64(len(words))
}

func (eca *EnhancedContentAnalyzer) calculateStructureQuality(html string) float64 {
	score := 0.0

	// Check for proper HTML structure
	if strings.Contains(html, "<h1>") {
		score += 0.2
	}
	if strings.Contains(html, "<h2>") {
		score += 0.15
	}
	if strings.Contains(html, "<p>") {
		score += 0.15
	}
	if strings.Contains(html, "<ul>") || strings.Contains(html, "<ol>") {
		score += 0.1
	}
	if strings.Contains(html, "<div>") {
		score += 0.1
	}

	return score
}

func (eca *EnhancedContentAnalyzer) calculateRelevanceScore(content *ScrapedContent, business string) float64 {
	score := 0.0
	text := strings.ToLower(content.Text)
	businessLower := strings.ToLower(business)

	// Business name relevance
	if strings.Contains(text, businessLower) {
		score += 0.4
	}

	// Industry keyword relevance
	industryKeywords := eca.extractIndustryKeywords(content.Text)
	score += float64(len(industryKeywords)) * 0.1

	// Content length relevance
	if len(content.Text) > 500 {
		score += 0.2
	} else if len(content.Text) > 200 {
		score += 0.1
	}

	return score
}

func (eca *EnhancedContentAnalyzer) countSpamIndicators(text string) int {
	count := 0
	textLower := strings.ToLower(text)

	for _, pattern := range eca.spamPatterns {
		matches := pattern.FindAllString(textLower, -1)
		count += len(matches)
	}

	return count
}

func (eca *EnhancedContentAnalyzer) countProfessionalIndicators(text string) int {
	count := 0
	textLower := strings.ToLower(text)

	for _, pattern := range eca.professionalPatterns {
		matches := pattern.FindAllString(textLower, -1)
		count += len(matches)
	}

	return count
}

func (eca *EnhancedContentAnalyzer) calculateOverallQuality(assessment *ContentQualityAssessment) float64 {
	score := 0.0

	// Weighted combination of quality factors
	score += assessment.ReadabilityScore * 0.2
	score += assessment.StructureQuality * 0.2
	score += assessment.RelevanceScore * 0.3

	// Content length bonus
	if assessment.ContentLength > 1000 {
		score += 0.1
	} else if assessment.ContentLength > 500 {
		score += 0.05
	}

	// Professional indicators bonus
	score += float64(assessment.ProfessionalIndicators) * 0.02

	// Spam indicators penalty
	score -= float64(assessment.SpamIndicators) * 0.05

	// Ensure score is between 0 and 1
	if score > 1.0 {
		score = 1.0
	} else if score < 0.0 {
		score = 0.0
	}

	return score
}

func (eca *EnhancedContentAnalyzer) extractIndustryIndicators(content *ScrapedContent) []string {
	return eca.extractIndustryKeywords(content.Text)
}

func (eca *EnhancedContentAnalyzer) extractBusinessKeywords(content *ScrapedContent, business string) []string {
	var keywords []string
	text := strings.ToLower(content.Text)
	businessLower := strings.ToLower(business)

	// Add business name words
	businessWords := strings.Fields(businessLower)
	keywords = append(keywords, businessWords...)

	// Add industry-related keywords found in content
	industryKeywords := eca.extractIndustryKeywords(content.Text)
	keywords = append(keywords, industryKeywords...)

	// Add service and product keywords
	serviceKeywords := eca.extractServiceKeywords(content.Text)
	keywords = append(keywords, serviceKeywords...)

	productKeywords := eca.extractProductKeywords(content.Text)
	keywords = append(keywords, productKeywords...)

	return keywords
}

func (eca *EnhancedContentAnalyzer) calculateConfidenceScore(analysis *EnhancedContentAnalysis) float64 {
	score := 0.0

	// Meta tag quality contribution
	if analysis.MetaTags != nil {
		score += analysis.MetaTags.Quality * 0.2
	}

	// Structured data quality contribution
	if analysis.StructuredData != nil {
		score += analysis.StructuredData.Quality * 0.25
	}

	// Semantic analysis contribution
	if analysis.SemanticAnalysis != nil {
		score += analysis.SemanticAnalysis.SemanticScore * 0.3
	}

	// Content quality contribution
	if analysis.ContentQuality != nil {
		score += analysis.ContentQuality.OverallQuality * 0.25
	}

	return score
}
