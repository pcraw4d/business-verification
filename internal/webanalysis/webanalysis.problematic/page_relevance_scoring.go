package webanalysis

import (
	"math"
	"regexp"
	"strings"
	"time"
)

// PageRelevanceScore represents a comprehensive relevance score for a web page
type PageRelevanceScore struct {
	OverallScore       float64            `json:"overall_score"`
	ComponentScores    map[string]float64 `json:"component_scores"`
	ConfidenceLevel    float64            `json:"confidence_level"`
	ScoringFactors     []ScoringFactor    `json:"scoring_factors"`
	BusinessRelevance  BusinessRelevance  `json:"business_relevance"`
	ContentRelevance   ContentRelevance   `json:"content_relevance"`
	TechnicalRelevance TechnicalRelevance `json:"technical_relevance"`
	ScoredAt           time.Time          `json:"scored_at"`
}

// ScoringFactor represents a specific factor that contributed to the score
type ScoringFactor struct {
	Factor     string  `json:"factor"`
	Score      float64 `json:"score"`
	Weight     float64 `json:"weight"`
	Confidence float64 `json:"confidence"`
	Reason     string  `json:"reason"`
}

// BusinessRelevance represents business-specific relevance metrics
type BusinessRelevance struct {
	BusinessNameMatch   float64  `json:"business_name_match"`
	IndustryRelevance   float64  `json:"industry_relevance"`
	GeographicRelevance float64  `json:"geographic_relevance"`
	BusinessKeywords    []string `json:"business_keywords"`
	BusinessEntities    []string `json:"business_entities"`
	ContactInformation  float64  `json:"contact_information"`
	RegistrationData    float64  `json:"registration_data"`
}

// ContentRelevance represents content-specific relevance metrics
type ContentRelevance struct {
	ContentQuality     float64  `json:"content_quality"`
	ContentLength      int      `json:"content_length"`
	ContentFreshness   float64  `json:"content_freshness"`
	ContentStructure   float64  `json:"content_structure"`
	ContentReadability float64  `json:"content_readability"`
	ContentTopics      []string `json:"content_topics"`
	ContentSentiment   float64  `json:"content_sentiment"`
	ContentCredibility float64  `json:"content_credibility"`
}

// TechnicalRelevance represents technical aspects of relevance
type TechnicalRelevance struct {
	PageAuthority      float64 `json:"page_authority"`
	PageSpeed          float64 `json:"page_speed"`
	MobileFriendliness float64 `json:"mobile_friendliness"`
	SEOOptimization    float64 `json:"seo_optimization"`
	Accessibility      float64 `json:"accessibility"`
	SecurityScore      float64 `json:"security_score"`
}

// PageRelevanceScorer manages page relevance scoring operations
type PageRelevanceScorer struct {
	config            RelevanceScoringConfig
	businessMatcher   *BusinessMatcher
	contentAnalyzer   *ContentAnalyzer
	technicalAnalyzer *TechnicalAnalyzer
}

// RelevanceScoringConfig holds configuration for relevance scoring
type RelevanceScoringConfig struct {
	Weights                  map[string]float64  `json:"weights"`
	MinConfidence            float64             `json:"min_confidence"`
	MaxContentLength         int                 `json:"max_content_length"`
	BusinessKeywords         []string            `json:"business_keywords"`
	IndustryKeywords         map[string][]string `json:"industry_keywords"`
	GeographicKeywords       []string            `json:"geographic_keywords"`
	ContentQualityThresholds map[string]float64  `json:"content_quality_thresholds"`
	TechnicalThresholds      map[string]float64  `json:"technical_thresholds"`
}

// BusinessMatcher handles business name and entity matching
type BusinessMatcher struct {
	fuzzyMatcher    *FuzzyMatcher
	entityExtractor *EntityExtractor
	keywordMatcher  *KeywordMatcher
}

// ContentAnalyzer handles content analysis and scoring
type ContentAnalyzer struct {
	readabilityAnalyzer *ReadabilityAnalyzer
	topicAnalyzer       *TopicAnalyzer
	sentimentAnalyzer   *SentimentAnalyzer
	credibilityAnalyzer *CredibilityAnalyzer
}

// TechnicalAnalyzer handles technical aspects analysis
type TechnicalAnalyzer struct {
	seoAnalyzer         *SEOAnalyzer
	securityAnalyzer    *SecurityAnalyzer
	performanceAnalyzer *PerformanceAnalyzer
}

// NewPageRelevanceScorer creates a new page relevance scorer
func NewPageRelevanceScorer() *PageRelevanceScorer {
	scorer := &PageRelevanceScorer{
		config: RelevanceScoringConfig{
			Weights: map[string]float64{
				"business_name_match":  0.25,
				"industry_relevance":   0.20,
				"content_quality":      0.15,
				"geographic_relevance": 0.10,
				"contact_information":  0.10,
				"technical_relevance":  0.10,
				"content_freshness":    0.05,
				"content_structure":    0.05,
			},
			MinConfidence:    0.3,
			MaxContentLength: 50000,
			BusinessKeywords: []string{
				"company", "business", "enterprise", "organization", "corporation",
				"services", "products", "solutions", "offerings", "capabilities",
				"industry", "sector", "market", "clients", "customers",
			},
			IndustryKeywords: map[string][]string{
				"technology":    {"software", "hardware", "IT", "digital", "tech"},
				"finance":       {"banking", "investment", "financial", "insurance", "credit"},
				"healthcare":    {"medical", "health", "pharmaceutical", "clinical", "patient"},
				"retail":        {"commerce", "retail", "ecommerce", "shopping", "store"},
				"manufacturing": {"manufacturing", "production", "industrial", "factory", "supply"},
			},
			GeographicKeywords: []string{
				"location", "address", "city", "state", "country", "region",
				"headquarters", "office", "branch", "location",
			},
			ContentQualityThresholds: map[string]float64{
				"min_length":      100,
				"max_length":      10000,
				"min_readability": 0.6,
				"min_structure":   0.5,
			},
			TechnicalThresholds: map[string]float64{
				"min_seo_score":         0.5,
				"min_security_score":    0.7,
				"min_performance_score": 0.6,
			},
		},
	}

	scorer.initializeComponents()
	return scorer
}

// initializeComponents initializes all analysis components
func (prs *PageRelevanceScorer) initializeComponents() {
	prs.businessMatcher = &BusinessMatcher{
		fuzzyMatcher:    NewFuzzyMatcher(),
		entityExtractor: NewEntityExtractor(),
		keywordMatcher:  NewKeywordMatcher(),
	}

	prs.contentAnalyzer = &ContentAnalyzer{
		readabilityAnalyzer: NewReadabilityAnalyzer(),
		topicAnalyzer:       NewTopicAnalyzer(),
		sentimentAnalyzer:   NewSentimentAnalyzer(),
		credibilityAnalyzer: NewCredibilityAnalyzer(),
	}

	prs.technicalAnalyzer = &TechnicalAnalyzer{
		seoAnalyzer:         NewSEOAnalyzer(),
		securityAnalyzer:    NewSecurityAnalyzer(),
		performanceAnalyzer: NewPerformanceAnalyzer(),
	}
}

// ScorePage calculates a comprehensive relevance score for a web page
func (prs *PageRelevanceScorer) ScorePage(content *ScrapedContent, business string, context *ScoringContext) *PageRelevanceScore {
	score := &PageRelevanceScore{
		ComponentScores: make(map[string]float64),
		ScoringFactors:  []ScoringFactor{},
		ScoredAt:        time.Now(),
	}

	// Calculate business relevance
	score.BusinessRelevance = prs.calculateBusinessRelevance(content, business, context)

	// Calculate content relevance
	score.ContentRelevance = prs.calculateContentRelevance(content, business, context)

	// Calculate technical relevance
	score.TechnicalRelevance = prs.calculateTechnicalRelevance(content, context)

	// Calculate component scores
	prs.calculateComponentScores(score)

	// Calculate overall score
	score.OverallScore = prs.calculateOverallScore(score)

	// Calculate confidence level
	score.ConfidenceLevel = prs.calculateConfidenceLevel(score)

	return score
}

// calculateBusinessRelevance calculates business-specific relevance metrics
func (prs *PageRelevanceScorer) calculateBusinessRelevance(content *ScrapedContent, business string, context *ScoringContext) BusinessRelevance {
	text := strings.ToLower(content.Text + " " + content.Title)

	relevance := BusinessRelevance{
		BusinessKeywords: []string{},
		BusinessEntities: []string{},
	}

	// Business name matching
	relevance.BusinessNameMatch = prs.businessMatcher.fuzzyMatcher.CalculateMatch(text, business)

	// Industry relevance
	relevance.IndustryRelevance = prs.calculateIndustryRelevance(text, context)

	// Geographic relevance
	relevance.GeographicRelevance = prs.calculateGeographicRelevance(text)

	// Business keywords
	relevance.BusinessKeywords = prs.extractBusinessKeywords(text)

	// Business entities
	relevance.BusinessEntities = prs.businessMatcher.entityExtractor.ExtractEntities(text)

	// Contact information
	relevance.ContactInformation = prs.calculateContactInformationScore(text)

	// Registration data
	relevance.RegistrationData = prs.calculateRegistrationDataScore(text)

	return relevance
}

// calculateContentRelevance calculates content-specific relevance metrics
func (prs *PageRelevanceScorer) calculateContentRelevance(content *ScrapedContent, business string, context *ScoringContext) ContentRelevance {
	relevance := ContentRelevance{
		ContentTopics: []string{},
	}

	// Enhanced content quality assessment using the new quality assessor
	qualityConfig := ContentQualityConfig{
		Weights: map[string]float64{
			"readability":       0.2,
			"structure":         0.2,
			"completeness":      0.2,
			"business_content":  0.2,
			"technical_content": 0.2,
		},
	}

	qualityAssessor := NewPageContentQualityAssessor(qualityConfig)
	contentQuality := qualityAssessor.AssessContentQuality(content, business)

	// Use the comprehensive content quality assessment
	relevance.ContentQuality = contentQuality.OverallQuality

	// Content length
	relevance.ContentLength = len(content.Text)

	// Content freshness
	relevance.ContentFreshness = prs.calculateContentFreshness(content, context)

	// Content structure (enhanced with quality assessment)
	relevance.ContentStructure = contentQuality.StructureMetrics.StructureScore

	// Content readability (enhanced with quality assessment)
	relevance.ContentReadability = contentQuality.ReadabilityMetrics.ReadabilityScore

	// Content topics
	relevance.ContentTopics = prs.contentAnalyzer.topicAnalyzer.ExtractTopics(content.Text)

	// Content sentiment
	relevance.ContentSentiment = prs.contentAnalyzer.sentimentAnalyzer.AnalyzeSentiment(content.Text)

	// Content credibility (enhanced with quality assessment)
	relevance.ContentCredibility = contentQuality.BusinessMetrics.BusinessScore

	return relevance
}

// calculateTechnicalRelevance calculates technical relevance metrics
func (prs *PageRelevanceScorer) calculateTechnicalRelevance(content *ScrapedContent, context *ScoringContext) TechnicalRelevance {
	relevance := TechnicalRelevance{}

	// Page authority (simulated)
	relevance.PageAuthority = prs.calculatePageAuthority(content.URL, context)

	// Page speed (simulated)
	relevance.PageSpeed = prs.technicalAnalyzer.performanceAnalyzer.CalculateSpeed(content.ResponseTime)

	// Mobile friendliness (simulated)
	relevance.MobileFriendliness = prs.calculateMobileFriendliness(content.HTML)

	// SEO optimization
	relevance.SEOOptimization = prs.technicalAnalyzer.seoAnalyzer.CalculateSEO(content)

	// Accessibility
	relevance.Accessibility = prs.calculateAccessibilityScore(content.HTML)

	// Security score
	relevance.SecurityScore = prs.technicalAnalyzer.securityAnalyzer.CalculateSecurity(content)

	return relevance
}

// calculateComponentScores calculates individual component scores
func (prs *PageRelevanceScorer) calculateComponentScores(score *PageRelevanceScore) {
	// Business name match score
	businessNameScore := score.BusinessRelevance.BusinessNameMatch
	score.ComponentScores["business_name_match"] = businessNameScore
	score.ScoringFactors = append(score.ScoringFactors, ScoringFactor{
		Factor:     "business_name_match",
		Score:      businessNameScore,
		Weight:     prs.config.Weights["business_name_match"],
		Confidence: 0.9,
		Reason:     "Business name matching accuracy",
	})

	// Industry relevance score
	industryScore := score.BusinessRelevance.IndustryRelevance
	score.ComponentScores["industry_relevance"] = industryScore
	score.ScoringFactors = append(score.ScoringFactors, ScoringFactor{
		Factor:     "industry_relevance",
		Score:      industryScore,
		Weight:     prs.config.Weights["industry_relevance"],
		Confidence: 0.8,
		Reason:     "Industry keyword presence and relevance",
	})

	// Content quality score
	contentQualityScore := score.ContentRelevance.ContentQuality
	score.ComponentScores["content_quality"] = contentQualityScore
	score.ScoringFactors = append(score.ScoringFactors, ScoringFactor{
		Factor:     "content_quality",
		Score:      contentQualityScore,
		Weight:     prs.config.Weights["content_quality"],
		Confidence: 0.7,
		Reason:     "Content quality and readability",
	})

	// Geographic relevance score
	geographicScore := score.BusinessRelevance.GeographicRelevance
	score.ComponentScores["geographic_relevance"] = geographicScore
	score.ScoringFactors = append(score.ScoringFactors, ScoringFactor{
		Factor:     "geographic_relevance",
		Score:      geographicScore,
		Weight:     prs.config.Weights["geographic_relevance"],
		Confidence: 0.6,
		Reason:     "Geographic information presence",
	})

	// Contact information score
	contactScore := score.BusinessRelevance.ContactInformation
	score.ComponentScores["contact_information"] = contactScore
	score.ScoringFactors = append(score.ScoringFactors, ScoringFactor{
		Factor:     "contact_information",
		Score:      contactScore,
		Weight:     prs.config.Weights["contact_information"],
		Confidence: 0.8,
		Reason:     "Contact information completeness",
	})

	// Technical relevance score
	technicalScore := (score.TechnicalRelevance.SEOOptimization +
		score.TechnicalRelevance.SecurityScore +
		score.TechnicalRelevance.PageSpeed) / 3.0
	score.ComponentScores["technical_relevance"] = technicalScore
	score.ScoringFactors = append(score.ScoringFactors, ScoringFactor{
		Factor:     "technical_relevance",
		Score:      technicalScore,
		Weight:     prs.config.Weights["technical_relevance"],
		Confidence: 0.7,
		Reason:     "Technical quality and optimization",
	})

	// Content freshness score
	freshnessScore := score.ContentRelevance.ContentFreshness
	score.ComponentScores["content_freshness"] = freshnessScore
	score.ScoringFactors = append(score.ScoringFactors, ScoringFactor{
		Factor:     "content_freshness",
		Score:      freshnessScore,
		Weight:     prs.config.Weights["content_freshness"],
		Confidence: 0.5,
		Reason:     "Content freshness and recency",
	})

	// Content structure score
	structureScore := score.ContentRelevance.ContentStructure
	score.ComponentScores["content_structure"] = structureScore
	score.ScoringFactors = append(score.ScoringFactors, ScoringFactor{
		Factor:     "content_structure",
		Score:      structureScore,
		Weight:     prs.config.Weights["content_structure"],
		Confidence: 0.6,
		Reason:     "Content structure and organization",
	})
}

// calculateOverallScore calculates the weighted overall score
func (prs *PageRelevanceScorer) calculateOverallScore(score *PageRelevanceScore) float64 {
	overallScore := 0.0
	totalWeight := 0.0

	for factor, componentScore := range score.ComponentScores {
		weight := prs.config.Weights[factor]
		overallScore += componentScore * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		return overallScore / totalWeight
	}
	return 0.0
}

// calculateConfidenceLevel calculates the confidence level of the scoring
func (prs *PageRelevanceScorer) calculateConfidenceLevel(score *PageRelevanceScore) float64 {
	if len(score.ScoringFactors) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, factor := range score.ScoringFactors {
		totalConfidence += factor.Confidence
	}

	return totalConfidence / float64(len(score.ScoringFactors))
}

// calculateIndustryRelevance calculates industry-specific relevance
func (prs *PageRelevanceScorer) calculateIndustryRelevance(text string, context *ScoringContext) float64 {
	if context == nil || context.Industry == "" {
		return 0.5 // Default neutral score
	}

	industryKeywords, exists := prs.config.IndustryKeywords[strings.ToLower(context.Industry)]
	if !exists {
		return 0.5
	}

	keywordCount := 0
	for _, keyword := range industryKeywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			keywordCount++
		}
	}

	if len(industryKeywords) == 0 {
		return 0.5
	}

	return float64(keywordCount) / float64(len(industryKeywords))
}

// calculateGeographicRelevance calculates geographic relevance
func (prs *PageRelevanceScorer) calculateGeographicRelevance(text string) float64 {
	keywordCount := 0
	for _, keyword := range prs.config.GeographicKeywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			keywordCount++
		}
	}

	if len(prs.config.GeographicKeywords) == 0 {
		return 0.0
	}

	return float64(keywordCount) / float64(len(prs.config.GeographicKeywords))
}

// extractBusinessKeywords extracts business-related keywords
func (prs *PageRelevanceScorer) extractBusinessKeywords(text string) []string {
	var keywords []string
	for _, keyword := range prs.config.BusinessKeywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			keywords = append(keywords, keyword)
		}
	}
	return keywords
}

// calculateContactInformationScore calculates contact information completeness
func (prs *PageRelevanceScorer) calculateContactInformationScore(text string) float64 {
	contactPatterns := []string{
		`\b\d{3}[-.]?\d{3}[-.]?\d{4}\b`,                                                         // Phone numbers
		`\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,                                   // Email addresses
		`\b\d+\s+[A-Za-z\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Lane|Ln|Drive|Dr)\b`, // Addresses
	}

	score := 0.0
	for _, pattern := range contactPatterns {
		matched, _ := regexp.MatchString(pattern, text)
		if matched {
			score += 0.33
		}
	}

	return math.Min(score, 1.0)
}

// calculateRegistrationDataScore calculates registration data presence
func (prs *PageRelevanceScorer) calculateRegistrationDataScore(text string) float64 {
	registrationPatterns := []string{
		`\b(?:incorporated|inc|corporation|corp|limited|ltd|llc|company|co)\b`,
		`\b(?:established|founded|since|established in|founded in)\b`,
		`\b(?:registration|registered|license|licensed)\b`,
	}

	score := 0.0
	for _, pattern := range registrationPatterns {
		matched, _ := regexp.MatchString(pattern, strings.ToLower(text))
		if matched {
			score += 0.33
		}
	}

	return math.Min(score, 1.0)
}

// calculateContentFreshness calculates content freshness score
func (prs *PageRelevanceScorer) calculateContentFreshness(content *ScrapedContent, context *ScoringContext) float64 {
	// This would typically use actual content dates
	// For now, we'll use a simulated freshness based on content characteristics
	return 0.7 // Simulated freshness score
}

// calculateContentStructure calculates content structure score
func (prs *PageRelevanceScorer) calculateContentStructure(html string) float64 {
	structureElements := []string{"<h1>", "<h2>", "<h3>", "<p>", "<ul>", "<ol>", "<table>"}
	score := 0.0

	for _, element := range structureElements {
		if strings.Contains(html, element) {
			score += 0.14 // 1.0 / 7 elements
		}
	}

	return math.Min(score, 1.0)
}

// calculatePageAuthority calculates page authority (simulated)
func (prs *PageRelevanceScorer) calculatePageAuthority(url string, context *ScoringContext) float64 {
	// This would typically use actual page authority metrics
	// For now, we'll use a simulated authority based on URL characteristics
	return 0.6 // Simulated authority score
}

// calculateMobileFriendliness calculates mobile friendliness score
func (prs *PageRelevanceScorer) calculateMobileFriendliness(html string) float64 {
	mobileIndicators := []string{
		`viewport`,
		`mobile`,
		`responsive`,
		`@media.*max-width`,
	}

	score := 0.0
	for _, indicator := range mobileIndicators {
		if strings.Contains(strings.ToLower(html), indicator) {
			score += 0.25
		}
	}

	return math.Min(score, 1.0)
}

// calculateAccessibilityScore calculates accessibility score
func (prs *PageRelevanceScorer) calculateAccessibilityScore(html string) float64 {
	accessibilityElements := []string{
		`alt=`,
		`aria-`,
		`role=`,
		`tabindex=`,
	}

	score := 0.0
	for _, element := range accessibilityElements {
		if strings.Contains(html, element) {
			score += 0.25
		}
	}

	return math.Min(score, 1.0)
}

// ScoringContext provides context for scoring calculations
type ScoringContext struct {
	Industry     string            `json:"industry"`
	Location     string            `json:"location"`
	BusinessType string            `json:"business_type"`
	Metadata     map[string]string `json:"metadata"`
}

// CalculateMatch calculates fuzzy match score between text and business name
func (fm *FuzzyMatcher) CalculateMatch(text, business string) float64 {
	// Simplified fuzzy matching - in production, use a proper fuzzy matching library
	textLower := strings.ToLower(text)
	businessLower := strings.ToLower(business)

	// Exact match
	if strings.Contains(textLower, businessLower) {
		return 1.0
	}

	// Word-by-word matching
	businessWords := strings.Fields(businessLower)
	matchedWords := 0

	for _, word := range businessWords {
		if len(word) > 2 && strings.Contains(textLower, word) {
			matchedWords++
		}
	}

	if len(businessWords) == 0 {
		return 0.0
	}

	return float64(matchedWords) / float64(len(businessWords))
}

// EntityExtractor extracts business entities from text
type EntityExtractor struct{}

// NewEntityExtractor creates a new entity extractor
func NewEntityExtractor() *EntityExtractor {
	return &EntityExtractor{}
}

// ExtractEntities extracts business entities from text
func (ee *EntityExtractor) ExtractEntities(text string) []string {
	// Simplified entity extraction - in production, use NLP libraries
	entities := []string{}

	// Extract potential company names (capitalized words)
	words := strings.Fields(text)
	for i, word := range words {
		if len(word) > 2 && strings.ToUpper(word) == word {
			entities = append(entities, word)
		}
		// Look for patterns like "Company Name Inc" or "Company Name LLC"
		if i > 0 && strings.Contains(strings.ToLower(word), "inc") ||
			strings.Contains(strings.ToLower(word), "llc") ||
			strings.Contains(strings.ToLower(word), "corp") {
			if i > 0 {
				entities = append(entities, words[i-1]+" "+word)
			}
		}
	}

	return entities
}

// KeywordMatcher handles keyword matching
type KeywordMatcher struct{}

// NewKeywordMatcher creates a new keyword matcher
func NewKeywordMatcher() *KeywordMatcher {
	return &KeywordMatcher{}
}

// ReadabilityAnalyzer analyzes text readability
type ReadabilityAnalyzer struct{}

// NewReadabilityAnalyzer creates a new readability analyzer
func NewReadabilityAnalyzer() *ReadabilityAnalyzer {
	return &ReadabilityAnalyzer{}
}

// CalculateQuality calculates content quality score
func (ra *ReadabilityAnalyzer) CalculateQuality(text string) float64 {
	if len(text) == 0 {
		return 0.0
	}

	score := 0.0

	// Length score
	if len(text) > 1000 {
		score += 0.3
	} else if len(text) > 500 {
		score += 0.2
	} else if len(text) > 100 {
		score += 0.1
	}

	// Sentence structure score
	sentences := strings.Split(text, ".")
	if len(sentences) > 5 {
		score += 0.3
	}

	// Word variety score
	words := strings.Fields(text)
	uniqueWords := make(map[string]bool)
	for _, word := range words {
		uniqueWords[strings.ToLower(word)] = true
	}

	if len(uniqueWords) > 50 {
		score += 0.4
	}

	return math.Min(score, 1.0)
}

// CalculateReadability calculates readability score
func (ra *ReadabilityAnalyzer) CalculateReadability(text string) float64 {
	// Simplified readability calculation
	words := strings.Fields(text)
	sentences := strings.Split(text, ".")

	if len(sentences) == 0 || len(words) == 0 {
		return 0.0
	}

	avgWordsPerSentence := float64(len(words)) / float64(len(sentences))

	// Flesch Reading Ease approximation
	if avgWordsPerSentence < 15 {
		return 0.9
	} else if avgWordsPerSentence < 20 {
		return 0.7
	} else if avgWordsPerSentence < 25 {
		return 0.5
	} else {
		return 0.3
	}
}

// TopicAnalyzer analyzes content topics
type TopicAnalyzer struct{}

// NewTopicAnalyzer creates a new topic analyzer
func NewTopicAnalyzer() *TopicAnalyzer {
	return &TopicAnalyzer{}
}

// ExtractTopics extracts topics from text
func (ta *TopicAnalyzer) ExtractTopics(text string) []string {
	// Simplified topic extraction
	topics := []string{}

	// Look for common business topics
	businessTopics := []string{
		"business", "company", "services", "products", "solutions",
		"technology", "finance", "healthcare", "retail", "manufacturing",
		"marketing", "sales", "customer", "client", "industry",
	}

	for _, topic := range businessTopics {
		if strings.Contains(strings.ToLower(text), topic) {
			topics = append(topics, topic)
		}
	}

	return topics
}

// SentimentAnalyzer analyzes text sentiment
type SentimentAnalyzer struct{}

// NewSentimentAnalyzer creates a new sentiment analyzer
func NewSentimentAnalyzer() *SentimentAnalyzer {
	return &SentimentAnalyzer{}
}

// AnalyzeSentiment analyzes sentiment of text
func (sa *SentimentAnalyzer) AnalyzeSentiment(text string) float64 {
	// Simplified sentiment analysis
	positiveWords := []string{"excellent", "great", "good", "best", "leading", "innovative", "quality"}
	negativeWords := []string{"poor", "bad", "worst", "terrible", "failed", "problem", "issue"}

	positiveCount := 0
	negativeCount := 0

	textLower := strings.ToLower(text)
	for _, word := range positiveWords {
		positiveCount += strings.Count(textLower, word)
	}
	for _, word := range negativeWords {
		negativeCount += strings.Count(textLower, word)
	}

	total := positiveCount + negativeCount
	if total == 0 {
		return 0.5 // Neutral
	}

	return float64(positiveCount) / float64(total)
}

// CredibilityAnalyzer analyzes content credibility
type CredibilityAnalyzer struct{}

// NewCredibilityAnalyzer creates a new credibility analyzer
func NewCredibilityAnalyzer() *CredibilityAnalyzer {
	return &CredibilityAnalyzer{}
}

// CalculateCredibility calculates content credibility score
func (ca *CredibilityAnalyzer) CalculateCredibility(content *ScrapedContent) float64 {
	score := 0.0

	// Check for professional indicators
	professionalIndicators := []string{
		"certified", "licensed", "accredited", "professional", "expert",
		"experience", "years", "established", "trusted", "reliable",
	}

	text := strings.ToLower(content.Text + " " + content.Title)
	for _, indicator := range professionalIndicators {
		if strings.Contains(text, indicator) {
			score += 0.1
		}
	}

	// Check for contact information
	if strings.Contains(text, "@") || strings.Contains(text, "phone") || strings.Contains(text, "contact") {
		score += 0.2
	}

	// Check for SSL (HTTPS)
	if strings.Contains(content.URL, "https://") {
		score += 0.1
	}

	return math.Min(score, 1.0)
}

// SEOAnalyzer analyzes SEO aspects
type SEOAnalyzer struct{}

// NewSEOAnalyzer creates a new SEO analyzer
func NewSEOAnalyzer() *SEOAnalyzer {
	return &SEOAnalyzer{}
}

// CalculateSEO calculates SEO score
func (sa *SEOAnalyzer) CalculateSEO(content *ScrapedContent) float64 {
	score := 0.0

	// Title analysis
	if len(content.Title) > 10 && len(content.Title) < 60 {
		score += 0.2
	}

	// Meta description check (simulated)
	score += 0.2

	// Heading structure
	if strings.Contains(content.HTML, "<h1>") {
		score += 0.2
	}
	if strings.Contains(content.HTML, "<h2>") {
		score += 0.1
	}

	// Content length
	if len(content.Text) > 300 {
		score += 0.2
	}

	// Internal links (simulated)
	score += 0.1

	return math.Min(score, 1.0)
}

// SecurityAnalyzer analyzes security aspects
type SecurityAnalyzer struct{}

// NewSecurityAnalyzer creates a new security analyzer
func NewSecurityAnalyzer() *SecurityAnalyzer {
	return &SecurityAnalyzer{}
}

// CalculateSecurity calculates security score
func (sa *SecurityAnalyzer) CalculateSecurity(content *ScrapedContent) float64 {
	score := 0.0

	// HTTPS check
	if strings.Contains(content.URL, "https://") {
		score += 0.4
	}

	// SSL certificate (simulated)
	score += 0.3

	// Security headers (simulated)
	score += 0.2

	// Content security (simulated)
	score += 0.1

	return math.Min(score, 1.0)
}

// PerformanceAnalyzer analyzes performance aspects
type PerformanceAnalyzer struct{}

// NewPerformanceAnalyzer creates a new performance analyzer
func NewPerformanceAnalyzer() *PerformanceAnalyzer {
	return &PerformanceAnalyzer{}
}

// CalculateSpeed calculates page speed score
func (pa *PerformanceAnalyzer) CalculateSpeed(responseTime time.Duration) float64 {
	// Convert response time to score (lower is better)
	if responseTime < time.Second {
		return 1.0
	} else if responseTime < 2*time.Second {
		return 0.8
	} else if responseTime < 3*time.Second {
		return 0.6
	} else if responseTime < 5*time.Second {
		return 0.4
	} else {
		return 0.2
	}
}
