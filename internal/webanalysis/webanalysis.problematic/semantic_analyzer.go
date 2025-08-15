package webanalysis

import (
	"math"
	"regexp"
	"strings"
	"time"
)

// SemanticAnalysisResult represents the result of semantic analysis
type SemanticAnalysisResult struct {
	IndustryKeywords    []string            `json:"industry_keywords"`
	BusinessDescription string              `json:"business_description"`
	ServiceKeywords     []string            `json:"service_keywords"`
	ProductKeywords     []string            `json:"product_keywords"`
	LocationKeywords    []string            `json:"location_keywords"`
	SemanticScore       float64             `json:"semantic_score"`
	TopicClusters       map[string][]string `json:"topic_clusters"`
	EntityExtraction    map[string][]string `json:"entity_extraction"`
	SimilarityScores    map[string]float64  `json:"similarity_scores"`
	IndustryConfidence  map[string]float64  `json:"industry_confidence"`
	BusinessEntities    []BusinessEntity    `json:"business_entities"`
	ContentRelevance    float64             `json:"content_relevance"`
	AnalyzedAt          time.Time           `json:"analyzed_at"`
}

// BusinessEntity represents a business-related entity found in content
type BusinessEntity struct {
	Type       string  `json:"type"`
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	Position   int     `json:"position"`
	Context    string  `json:"context"`
}

// SemanticAnalyzer manages semantic analysis operations
type SemanticAnalyzer struct {
	config              *SemanticAnalyzerConfig
	industryKeywords    map[string][]string
	serviceKeywords     map[string][]string
	productKeywords     map[string][]string
	locationPatterns    []*regexp.Regexp
	entityPatterns      map[string]*regexp.Regexp
	sentenceEmbeddings  map[string][]float64
	similarityThreshold float64
}

// SemanticAnalyzerConfig holds configuration for semantic analysis
type SemanticAnalyzerConfig struct {
	EnableSentenceTransformers bool    `json:"enable_sentence_transformers"`
	EnableIndustryExtraction   bool    `json:"enable_industry_extraction"`
	EnableEntityExtraction     bool    `json:"enable_entity_extraction"`
	EnableSimilarityMatching   bool    `json:"enable_similarity_matching"`
	SimilarityThreshold        float64 `json:"similarity_threshold"`
	MinConfidence              float64 `json:"min_confidence"`
	MaxEntities                int     `json:"max_entities"`
	EnableTopicClustering      bool    `json:"enable_topic_clustering"`
}

// NewSemanticAnalyzer creates a new semantic analyzer
func NewSemanticAnalyzer(config *SemanticAnalyzerConfig) *SemanticAnalyzer {
	if config == nil {
		config = &SemanticAnalyzerConfig{
			EnableSentenceTransformers: true,
			EnableIndustryExtraction:   true,
			EnableEntityExtraction:     true,
			EnableSimilarityMatching:   true,
			SimilarityThreshold:        0.7,
			MinConfidence:              0.6,
			MaxEntities:                50,
			EnableTopicClustering:      true,
		}
	}

	analyzer := &SemanticAnalyzer{
		config:              config,
		industryKeywords:    make(map[string][]string),
		serviceKeywords:     make(map[string][]string),
		productKeywords:     make(map[string][]string),
		locationPatterns:    []*regexp.Regexp{},
		entityPatterns:      make(map[string]*regexp.Regexp),
		sentenceEmbeddings:  make(map[string][]float64),
		similarityThreshold: config.SimilarityThreshold,
	}

	analyzer.initializeKeywords()
	analyzer.initializePatterns()
	analyzer.initializeEmbeddings()

	return analyzer
}

// AnalyzeSemanticContent performs comprehensive semantic analysis
func (sa *SemanticAnalyzer) AnalyzeSemanticContent(content *ScrapedContent, business string) *SemanticAnalysisResult {
	result := &SemanticAnalysisResult{
		TopicClusters:      make(map[string][]string),
		EntityExtraction:   make(map[string][]string),
		SimilarityScores:   make(map[string]float64),
		IndustryConfidence: make(map[string]float64),
		BusinessEntities:   []BusinessEntity{},
		AnalyzedAt:         time.Now(),
	}

	// Extract industry-specific keywords
	if sa.config.EnableIndustryExtraction {
		result.IndustryKeywords = sa.extractIndustryKeywords(content.Text)
	}

	// Extract business description
	result.BusinessDescription = sa.extractBusinessDescription(content, business)

	// Extract service and product keywords
	result.ServiceKeywords = sa.extractServiceKeywords(content.Text)
	result.ProductKeywords = sa.extractProductKeywords(content.Text)

	// Extract location keywords
	result.LocationKeywords = sa.extractLocationKeywords(content.Text)

	// Perform semantic similarity matching
	if sa.config.EnableSimilarityMatching {
		result.SimilarityScores = sa.calculateSimilarityScores(content.Text, business)
	}

	// Calculate industry confidence scores
	result.IndustryConfidence = sa.calculateIndustryConfidence(content.Text)

	// Extract business entities
	if sa.config.EnableEntityExtraction {
		result.BusinessEntities = sa.extractBusinessEntities(content.Text)
	}

	// Create topic clusters
	if sa.config.EnableTopicClustering {
		result.TopicClusters = sa.createTopicClusters(content.Text)
	}

	// Extract entities by type
	result.EntityExtraction = sa.extractEntitiesByType(content.Text)

	// Calculate semantic score
	result.SemanticScore = sa.calculateSemanticScore(content, business, result)

	// Calculate content relevance
	result.ContentRelevance = sa.calculateContentRelevance(content, business, result)

	return result
}

// extractIndustryKeywords extracts industry-specific keywords from text
func (sa *SemanticAnalyzer) extractIndustryKeywords(text string) []string {
	var keywords []string
	textLower := strings.ToLower(text)

	for _, industryKeywords := range sa.industryKeywords {
		for _, keyword := range industryKeywords {
			if strings.Contains(textLower, keyword) {
				keywords = append(keywords, keyword)
			}
		}
	}

	return keywords
}

// extractBusinessDescription extracts business description using semantic analysis
func (sa *SemanticAnalyzer) extractBusinessDescription(content *ScrapedContent, business string) string {
	// Extract from content using semantic patterns
	text := content.Text
	sentences := strings.Split(text, ".")

	// Look for sentences that contain business-related semantic patterns
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		if len(sentence) > 30 && len(sentence) < 300 {
			lowerSentence := strings.ToLower(sentence)

			// Check for business description patterns
			if sa.isBusinessDescriptionSentence(lowerSentence, business) {
				return sentence
			}
		}
	}

	return ""
}

// isBusinessDescriptionSentence checks if a sentence is likely a business description
func (sa *SemanticAnalyzer) isBusinessDescriptionSentence(sentence, business string) bool {
	// Business description indicators
	indicators := []string{
		"company", "business", "organization", "enterprise", "corporation",
		"specializes", "provides", "offers", "delivers", "serves",
		"founded", "established", "since", "leading", "premier",
		"dedicated", "committed", "focused", "expert", "professional",
	}

	indicatorCount := 0
	for _, indicator := range indicators {
		if strings.Contains(sentence, indicator) {
			indicatorCount++
		}
	}

	// Business name presence
	businessLower := strings.ToLower(business)
	businessWords := strings.Fields(businessLower)
	businessWordCount := 0
	for _, word := range businessWords {
		if len(word) > 2 && strings.Contains(sentence, word) {
			businessWordCount++
		}
	}

	// Score the sentence
	score := float64(indicatorCount) * 0.3
	score += float64(businessWordCount) * 0.4

	return score >= 0.5
}

// extractServiceKeywords extracts service-related keywords
func (sa *SemanticAnalyzer) extractServiceKeywords(text string) []string {
	var keywords []string
	textLower := strings.ToLower(text)

	for _, serviceKeywords := range sa.serviceKeywords {
		for _, keyword := range serviceKeywords {
			if strings.Contains(textLower, keyword) {
				keywords = append(keywords, keyword)
			}
		}
	}

	return keywords
}

// extractProductKeywords extracts product-related keywords
func (sa *SemanticAnalyzer) extractProductKeywords(text string) []string {
	var keywords []string
	textLower := strings.ToLower(text)

	for _, productKeywords := range sa.productKeywords {
		for _, keyword := range productKeywords {
			if strings.Contains(textLower, keyword) {
				keywords = append(keywords, keyword)
			}
		}
	}

	return keywords
}

// extractLocationKeywords extracts location-related keywords
func (sa *SemanticAnalyzer) extractLocationKeywords(text string) []string {
	var keywords []string

	for _, pattern := range sa.locationPatterns {
		matches := pattern.FindAllString(text, -1)
		keywords = append(keywords, matches...)
	}

	return keywords
}

// calculateSimilarityScores calculates semantic similarity scores
func (sa *SemanticAnalyzer) calculateSimilarityScores(text, business string) map[string]float64 {
	scores := make(map[string]float64)

	// Business name similarity
	businessLower := strings.ToLower(business)
	textLower := strings.ToLower(text)
	scores["business_name"] = sa.calculateStringSimilarity(textLower, businessLower)

	// Industry similarity
	industryKeywords := sa.extractIndustryKeywords(text)
	if len(industryKeywords) > 0 {
		scores["industry"] = float64(len(industryKeywords)) / 10.0 // Normalize
	}

	// Service similarity
	serviceKeywords := sa.extractServiceKeywords(text)
	if len(serviceKeywords) > 0 {
		scores["services"] = float64(len(serviceKeywords)) / 8.0 // Normalize
	}

	// Product similarity
	productKeywords := sa.extractProductKeywords(text)
	if len(productKeywords) > 0 {
		scores["products"] = float64(len(productKeywords)) / 8.0 // Normalize
	}

	return scores
}

// calculateStringSimilarity calculates similarity between two strings
func (sa *SemanticAnalyzer) calculateStringSimilarity(str1, str2 string) float64 {
	// Simple Jaccard similarity
	words1 := strings.Fields(str1)
	words2 := strings.Fields(str2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Create sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range words1 {
		if len(word) > 2 {
			set1[word] = true
		}
	}

	for _, word := range words2 {
		if len(word) > 2 {
			set2[word] = true
		}
	}

	// Calculate intersection and union
	intersection := 0
	for word := range set1 {
		if set2[word] {
			intersection++
		}
	}

	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// calculateIndustryConfidence calculates confidence scores for different industries
func (sa *SemanticAnalyzer) calculateIndustryConfidence(text string) map[string]float64 {
	confidence := make(map[string]float64)
	textLower := strings.ToLower(text)

	for industry, keywords := range sa.industryKeywords {
		score := 0.0
		keywordCount := 0

		for _, keyword := range keywords {
			if strings.Contains(textLower, keyword) {
				score += 1.0
				keywordCount++
			}
		}

		// Normalize score
		if len(keywords) > 0 {
			confidence[industry] = score / float64(len(keywords))
		} else {
			confidence[industry] = 0.0
		}
	}

	return confidence
}

// extractBusinessEntities extracts business-related entities from text
func (sa *SemanticAnalyzer) extractBusinessEntities(text string) []BusinessEntity {
	var entities []BusinessEntity

	// Extract company names
	companyPattern := sa.entityPatterns["company"]
	if companyPattern != nil {
		matches := companyPattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				entity := BusinessEntity{
					Type:       "company",
					Value:      match[1],
					Confidence: 0.8,
					Position:   strings.Index(text, match[1]),
					Context:    sa.extractContext(text, match[1]),
				}
				entities = append(entities, entity)
			}
		}
	}

	// Extract locations
	locationPattern := sa.entityPatterns["location"]
	if locationPattern != nil {
		matches := locationPattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				entity := BusinessEntity{
					Type:       "location",
					Value:      match[1],
					Confidence: 0.7,
					Position:   strings.Index(text, match[1]),
					Context:    sa.extractContext(text, match[1]),
				}
				entities = append(entities, entity)
			}
		}
	}

	// Extract phone numbers
	phonePattern := sa.entityPatterns["phone"]
	if phonePattern != nil {
		matches := phonePattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				entity := BusinessEntity{
					Type:       "phone",
					Value:      match[1],
					Confidence: 0.9,
					Position:   strings.Index(text, match[1]),
					Context:    sa.extractContext(text, match[1]),
				}
				entities = append(entities, entity)
			}
		}
	}

	// Extract email addresses
	emailPattern := sa.entityPatterns["email"]
	if emailPattern != nil {
		matches := emailPattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				entity := BusinessEntity{
					Type:       "email",
					Value:      match[1],
					Confidence: 0.9,
					Position:   strings.Index(text, match[1]),
					Context:    sa.extractContext(text, match[1]),
				}
				entities = append(entities, entity)
			}
		}
	}

	// Limit number of entities
	if len(entities) > sa.config.MaxEntities {
		entities = entities[:sa.config.MaxEntities]
	}

	return entities
}

// extractContext extracts context around an entity
func (sa *SemanticAnalyzer) extractContext(text, entity string) string {
	pos := strings.Index(text, entity)
	if pos == -1 {
		return ""
	}

	start := pos - 50
	if start < 0 {
		start = 0
	}

	end := pos + len(entity) + 50
	if end > len(text) {
		end = len(text)
	}

	return text[start:end]
}

// createTopicClusters creates topic clusters from text
func (sa *SemanticAnalyzer) createTopicClusters(text string) map[string][]string {
	clusters := make(map[string][]string)
	textLower := strings.ToLower(text)

	// Business-related cluster
	if strings.Contains(textLower, "company") || strings.Contains(textLower, "business") {
		clusters["business"] = []string{"company", "business", "organization", "enterprise", "corporation"}
	}

	// Service-related cluster
	if strings.Contains(textLower, "service") || strings.Contains(textLower, "solutions") {
		clusters["services"] = []string{"service", "solutions", "consulting", "support", "assistance"}
	}

	// Product-related cluster
	if strings.Contains(textLower, "product") || strings.Contains(textLower, "goods") {
		clusters["products"] = []string{"product", "goods", "items", "merchandise", "catalog"}
	}

	// Contact-related cluster
	if strings.Contains(textLower, "contact") || strings.Contains(textLower, "address") {
		clusters["contact"] = []string{"contact", "address", "phone", "email", "location"}
	}

	// Industry-specific clusters
	for industry, keywords := range sa.industryKeywords {
		industryFound := false
		for _, keyword := range keywords {
			if strings.Contains(textLower, keyword) {
				industryFound = true
				break
			}
		}
		if industryFound {
			clusters[industry] = keywords
		}
	}

	return clusters
}

// extractEntitiesByType extracts entities grouped by type
func (sa *SemanticAnalyzer) extractEntitiesByType(text string) map[string][]string {
	entities := make(map[string][]string)

	// Extract companies
	companyPattern := sa.entityPatterns["company"]
	if companyPattern != nil {
		matches := companyPattern.FindAllString(text, -1)
		entities["companies"] = matches
	}

	// Extract locations
	locationPattern := sa.entityPatterns["location"]
	if locationPattern != nil {
		matches := locationPattern.FindAllString(text, -1)
		entities["locations"] = matches
	}

	// Extract phone numbers
	phonePattern := sa.entityPatterns["phone"]
	if phonePattern != nil {
		matches := phonePattern.FindAllString(text, -1)
		entities["phones"] = matches
	}

	// Extract email addresses
	emailPattern := sa.entityPatterns["email"]
	if emailPattern != nil {
		matches := emailPattern.FindAllString(text, -1)
		entities["emails"] = matches
	}

	return entities
}

// calculateSemanticScore calculates overall semantic score
func (sa *SemanticAnalyzer) calculateSemanticScore(content *ScrapedContent, business string, result *SemanticAnalysisResult) float64 {
	score := 0.0

	// Business description score
	if result.BusinessDescription != "" {
		score += 0.3
	}

	// Industry keywords score
	score += float64(len(result.IndustryKeywords)) * 0.1

	// Service keywords score
	score += float64(len(result.ServiceKeywords)) * 0.1

	// Product keywords score
	score += float64(len(result.ProductKeywords)) * 0.1

	// Business entities score
	score += float64(len(result.BusinessEntities)) * 0.05

	// Similarity scores
	for _, similarityScore := range result.SimilarityScores {
		score += similarityScore * 0.1
	}

	// Industry confidence
	for _, confidence := range result.IndustryConfidence {
		score += confidence * 0.05
	}

	// Normalize score
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// calculateContentRelevance calculates content relevance score
func (sa *SemanticAnalyzer) calculateContentRelevance(content *ScrapedContent, business string, result *SemanticAnalysisResult) float64 {
	score := 0.0
	text := strings.ToLower(content.Text)
	businessLower := strings.ToLower(business)

	// Business name relevance
	if strings.Contains(text, businessLower) {
		score += 0.4
	}

	// Industry relevance
	if len(result.IndustryKeywords) > 0 {
		score += 0.2
	}

	// Business description relevance
	if result.BusinessDescription != "" {
		score += 0.2
	}

	// Entity relevance
	if len(result.BusinessEntities) > 0 {
		score += 0.1
	}

	// Semantic score contribution
	score += result.SemanticScore * 0.1

	return score
}

// Helper methods for initialization

func (sa *SemanticAnalyzer) initializeKeywords() {
	// Industry keywords
	sa.industryKeywords = map[string][]string{
		"technology":    {"software", "hardware", "IT", "digital", "tech", "computing", "programming", "development", "technology", "innovative"},
		"healthcare":    {"medical", "health", "care", "hospital", "clinic", "pharmacy", "wellness", "treatment", "healthcare", "medical services"},
		"finance":       {"banking", "financial", "investment", "insurance", "credit", "loan", "money", "wealth", "finance", "financial services"},
		"retail":        {"store", "shop", "retail", "commerce", "ecommerce", "merchandise", "sales", "customer", "retail", "shopping"},
		"manufacturing": {"manufacturing", "production", "factory", "industrial", "machinery", "equipment", "assembly", "manufacturing", "industrial"},
		"education":     {"education", "school", "university", "learning", "training", "academic", "teaching", "education", "educational"},
		"real_estate":   {"real estate", "property", "housing", "construction", "development", "architecture", "real estate", "property"},
		"consulting":    {"consulting", "advisory", "strategy", "management", "business", "professional", "expertise", "consulting", "advisory"},
	}

	// Service keywords
	sa.serviceKeywords = map[string][]string{
		"professional": {"consulting", "advisory", "strategy", "management", "professional services"},
		"technical":    {"technical support", "IT services", "software development", "technical consulting"},
		"business":     {"business services", "management consulting", "business solutions", "corporate services"},
		"customer":     {"customer service", "client support", "customer care", "support services"},
	}

	// Product keywords
	sa.productKeywords = map[string][]string{
		"software":   {"software", "application", "platform", "system", "solution"},
		"hardware":   {"hardware", "equipment", "device", "machine", "tool"},
		"consumer":   {"product", "goods", "item", "merchandise", "consumer goods"},
		"industrial": {"industrial", "machinery", "equipment", "industrial products"},
	}
}

func (sa *SemanticAnalyzer) initializePatterns() {
	// Location patterns
	sa.locationPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*,\s*[A-Z]{2}\s+\d{5})`),                  // City, State ZIP
		regexp.MustCompile(`(?i)([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*,\s*[A-Z]{2})`),                          // City, State
		regexp.MustCompile(`(?i)(\d+\s+[A-Z][a-z]+\s+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd))`), // Address
	}

	// Entity patterns
	sa.entityPatterns = map[string]*regexp.Regexp{
		"company":  regexp.MustCompile(`(?i)([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*\s+(?:Inc|LLC|Corp|Company|Ltd|Limited|Corporation))`),
		"location": regexp.MustCompile(`(?i)([A-Z][a-z]+(?:\s+[A-Z][a-z]+)*,\s*[A-Z]{2})`),
		"phone":    regexp.MustCompile(`(?i)(\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4})`),
		"email":    regexp.MustCompile(`(?i)([a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,})`),
	}
}

func (sa *SemanticAnalyzer) initializeEmbeddings() {
	// Initialize sentence embeddings for common business terms
	// This is a simplified implementation - in production, you'd use actual sentence transformers
	sa.sentenceEmbeddings = map[string][]float64{
		"business":      {0.1, 0.2, 0.3, 0.4, 0.5},
		"company":       {0.2, 0.3, 0.4, 0.5, 0.6},
		"services":      {0.3, 0.4, 0.5, 0.6, 0.7},
		"products":      {0.4, 0.5, 0.6, 0.7, 0.8},
		"consulting":    {0.5, 0.6, 0.7, 0.8, 0.9},
		"technology":    {0.6, 0.7, 0.8, 0.9, 1.0},
		"healthcare":    {0.7, 0.8, 0.9, 1.0, 0.1},
		"finance":       {0.8, 0.9, 1.0, 0.1, 0.2},
		"retail":        {0.9, 1.0, 0.1, 0.2, 0.3},
		"manufacturing": {1.0, 0.1, 0.2, 0.3, 0.4},
	}
}

// calculateCosineSimilarity calculates cosine similarity between two vectors
func (sa *SemanticAnalyzer) calculateCosineSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) || len(vec1) == 0 {
		return 0.0
	}

	dotProduct := 0.0
	magnitude1 := 0.0
	magnitude2 := 0.0

	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		magnitude1 += vec1[i] * vec1[i]
		magnitude2 += vec2[i] * vec2[i]
	}

	magnitude1 = math.Sqrt(magnitude1)
	magnitude2 = math.Sqrt(magnitude2)

	if magnitude1 == 0 || magnitude2 == 0 {
		return 0.0
	}

	return dotProduct / (magnitude1 * magnitude2)
}
