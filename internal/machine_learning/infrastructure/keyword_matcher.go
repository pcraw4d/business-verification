package infrastructure

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
)

// KeywordMatcher handles keyword-based matching for classification and risk detection
type KeywordMatcher struct {
	// Keyword databases
	industryKeywords map[string][]string
	riskKeywords     map[string][]string

	// Compiled regex patterns for performance
	compiledPatterns map[string]*regexp.Regexp

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// NewKeywordMatcher creates a new keyword matcher
func NewKeywordMatcher(logger *log.Logger) *KeywordMatcher {
	if logger == nil {
		logger = log.Default()
	}

	return &KeywordMatcher{
		industryKeywords: make(map[string][]string),
		riskKeywords:     make(map[string][]string),
		compiledPatterns: make(map[string]*regexp.Regexp),
		logger:           logger,
	}
}

// Initialize initializes the keyword matcher with keyword databases
func (km *KeywordMatcher) Initialize(ctx context.Context) error {
	km.mu.Lock()
	defer km.mu.Unlock()

	km.logger.Printf("🔍 Initializing Keyword Matcher")

	// Load industry keywords
	if err := km.loadIndustryKeywords(); err != nil {
		return fmt.Errorf("failed to load industry keywords: %w", err)
	}

	// Load risk keywords
	if err := km.loadRiskKeywords(); err != nil {
		return fmt.Errorf("failed to load risk keywords: %w", err)
	}

	// Compile regex patterns for performance
	if err := km.compilePatterns(); err != nil {
		return fmt.Errorf("failed to compile patterns: %w", err)
	}

	km.logger.Printf("✅ Keyword Matcher initialized with %d industry categories and %d risk categories",
		len(km.industryKeywords), len(km.riskKeywords))

	return nil
}

// ClassifyByKeywords performs classification using keyword matching
func (km *KeywordMatcher) ClassifyByKeywords(ctx context.Context, businessName, description string) ([]ClassificationPrediction, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	// Combine business name and description for analysis
	text := strings.ToLower(businessName + " " + description)

	var predictions []ClassificationPrediction
	industryScores := make(map[string]float64)

	// Score each industry based on keyword matches
	for industry, keywords := range km.industryKeywords {
		score := km.calculateKeywordScore(text, keywords)
		if score > 0 {
			industryScores[industry] = score
		}
	}

	// Convert scores to predictions
	rank := 1
	for industry, score := range industryScores {
		predictions = append(predictions, ClassificationPrediction{
			Label:       industry,
			Confidence:  score,
			Probability: score,
			Rank:        rank,
		})
		rank++
	}

	// Sort by confidence (highest first)
	for i := 0; i < len(predictions)-1; i++ {
		for j := i + 1; j < len(predictions); j++ {
			if predictions[i].Confidence < predictions[j].Confidence {
				predictions[i], predictions[j] = predictions[j], predictions[i]
			}
		}
	}

	// Update ranks
	for i := range predictions {
		predictions[i].Rank = i + 1
	}

	return predictions, nil
}

// DetectRiskKeywords detects risk keywords in the provided text
func (km *KeywordMatcher) DetectRiskKeywords(ctx context.Context, businessName, description, websiteContent string) ([]DetectedRisk, error) {
	km.mu.RLock()
	defer km.mu.RUnlock()

	// Combine all text for analysis
	text := strings.ToLower(businessName + " " + description + " " + websiteContent)

	var detectedRisks []DetectedRisk

	// Check each risk category
	for category, keywords := range km.riskKeywords {
		matches := km.findKeywordMatches(text, keywords)
		if len(matches) > 0 {
			risk := DetectedRisk{
				Category:    category,
				Severity:    km.determineRiskSeverity(category, len(matches)),
				Confidence:  km.calculateRiskConfidence(matches),
				Keywords:    matches,
				Description: km.generateRiskDescription(category, matches),
			}
			detectedRisks = append(detectedRisks, risk)
		}
	}

	return detectedRisks, nil
}

// HealthCheck performs a health check on the keyword matcher
func (km *KeywordMatcher) HealthCheck(ctx context.Context) error {
	km.mu.RLock()
	defer km.mu.RUnlock()

	// Check if keyword databases are loaded
	if len(km.industryKeywords) == 0 {
		return fmt.Errorf("industry keywords not loaded")
	}

	if len(km.riskKeywords) == 0 {
		return fmt.Errorf("risk keywords not loaded")
	}

	// Check if patterns are compiled
	if len(km.compiledPatterns) == 0 {
		return fmt.Errorf("patterns not compiled")
	}

	return nil
}

// loadIndustryKeywords loads industry keywords from the database
func (km *KeywordMatcher) loadIndustryKeywords() error {
	// This would typically load from a database or configuration file
	// For now, we'll use a comprehensive set of industry keywords

	km.industryKeywords = map[string][]string{
		"technology": {
			"software", "app", "platform", "saas", "cloud", "api", "development",
			"programming", "coding", "tech", "digital", "online", "web", "mobile",
			"artificial intelligence", "ai", "machine learning", "ml", "data",
			"analytics", "cybersecurity", "blockchain", "cryptocurrency",
		},
		"finance": {
			"bank", "banking", "financial", "finance", "investment", "trading",
			"credit", "loan", "mortgage", "insurance", "wealth", "asset",
			"payment", "fintech", "cryptocurrency", "bitcoin", "ethereum",
		},
		"healthcare": {
			"medical", "health", "healthcare", "hospital", "clinic", "doctor",
			"physician", "nurse", "pharmacy", "pharmaceutical", "drug", "medicine",
			"therapy", "treatment", "diagnosis", "patient", "wellness",
		},
		"retail": {
			"store", "shop", "retail", "merchandise", "product", "goods",
			"ecommerce", "online store", "marketplace", "shopping", "customer",
			"sales", "inventory", "warehouse", "distribution",
		},
		"manufacturing": {
			"manufacturing", "production", "factory", "plant", "assembly",
			"machinery", "equipment", "industrial", "automation", "supply chain",
			"logistics", "shipping", "distribution",
		},
		"education": {
			"education", "school", "university", "college", "learning", "training",
			"course", "curriculum", "student", "teacher", "instructor", "academic",
			"research", "study", "knowledge",
		},
		"real_estate": {
			"real estate", "property", "housing", "construction", "building",
			"development", "rental", "lease", "mortgage", "broker", "agent",
			"commercial", "residential",
		},
		"consulting": {
			"consulting", "consultant", "advisory", "strategy", "management",
			"business", "professional services", "expertise", "guidance",
			"analysis", "planning",
		},
		"media": {
			"media", "entertainment", "broadcasting", "publishing", "content",
			"news", "journalism", "advertising", "marketing", "social media",
			"digital media", "streaming",
		},
		"transportation": {
			"transportation", "logistics", "shipping", "delivery", "freight",
			"trucking", "aviation", "airline", "railway", "automotive",
			"vehicle", "fleet",
		},
	}

	return nil
}

// loadRiskKeywords loads risk keywords from the database
func (km *KeywordMatcher) loadRiskKeywords() error {
	// This would typically load from the risk_keywords table in Supabase
	// For now, we'll use a comprehensive set of risk keywords

	km.riskKeywords = map[string][]string{
		"illegal": {
			"drug", "cocaine", "heroin", "marijuana", "cannabis", "weapon",
			"firearm", "gun", "ammunition", "explosive", "human trafficking",
			"prostitution", "illegal gambling", "money laundering", "fraud",
			"scam", "theft", "robbery", "counterfeit", "stolen",
		},
		"prohibited": {
			"adult entertainment", "pornography", "escort", "dating", "casino",
			"gambling", "betting", "lottery", "tobacco", "cigarette", "alcohol",
			"liquor", "pharmaceutical", "prescription drug", "medical device",
			"cryptocurrency", "bitcoin", "ethereum", "crypto exchange",
		},
		"high_risk": {
			"money service", "check cashing", "wire transfer", "prepaid card",
			"gift card", "cryptocurrency exchange", "forex", "trading",
			"investment", "high risk merchant", "travel", "dating site",
			"adult content", "subscription service",
		},
		"tbml": {
			"shell company", "front company", "trade finance", "import",
			"export", "commodity", "precious metal", "gold", "silver",
			"diamond", "complex trade", "offshore", "tax haven",
		},
		"fraud": {
			"fake", "scam", "phishing", "identity theft", "credit card fraud",
			"payment fraud", "chargeback", "refund abuse", "fake business",
			"stolen identity", "synthetic identity",
		},
		"sanctions": {
			"sanctions", "ofac", "embargo", "restricted country", "terrorist",
			"terrorism", "money laundering", "corruption", "bribery",
		},
	}

	return nil
}

// compilePatterns compiles regex patterns for performance
func (km *KeywordMatcher) compilePatterns() error {
	// Compile patterns for all keywords
	allKeywords := make([]string, 0)

	// Add industry keywords
	for _, keywords := range km.industryKeywords {
		allKeywords = append(allKeywords, keywords...)
	}

	// Add risk keywords
	for _, keywords := range km.riskKeywords {
		allKeywords = append(allKeywords, keywords...)
	}

	// Compile patterns
	for _, keyword := range allKeywords {
		pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(keyword))
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("failed to compile pattern for keyword '%s': %w", keyword, err)
		}
		km.compiledPatterns[keyword] = compiled
	}

	return nil
}

// calculateKeywordScore calculates the score for a keyword match
func (km *KeywordMatcher) calculateKeywordScore(text string, keywords []string) float64 {
	var totalScore float64
	var matchCount int

	for _, keyword := range keywords {
		if pattern, exists := km.compiledPatterns[keyword]; exists {
			matches := pattern.FindAllString(text, -1)
			if len(matches) > 0 {
				matchCount++
				// Weight by keyword length and match frequency
				keywordWeight := float64(len(keyword)) / 20.0 // Normalize by average keyword length
				frequencyWeight := float64(len(matches)) * 0.1
				totalScore += keywordWeight + frequencyWeight
			}
		}
	}

	// Normalize score based on number of keywords and matches
	if len(keywords) == 0 {
		return 0.0
	}

	normalizedScore := totalScore / float64(len(keywords))

	// Boost score for multiple matches
	if matchCount > 1 {
		normalizedScore *= 1.2
	}

	// Cap at 1.0
	if normalizedScore > 1.0 {
		normalizedScore = 1.0
	}

	return normalizedScore
}

// findKeywordMatches finds matching keywords in the text
func (km *KeywordMatcher) findKeywordMatches(text string, keywords []string) []string {
	var matches []string

	for _, keyword := range keywords {
		if pattern, exists := km.compiledPatterns[keyword]; exists {
			if pattern.MatchString(text) {
				matches = append(matches, keyword)
			}
		}
	}

	return matches
}

// determineRiskSeverity determines the risk severity based on category and match count
func (km *KeywordMatcher) determineRiskSeverity(category string, matchCount int) string {
	// Base severity by category
	baseSeverity := map[string]string{
		"illegal":    "critical",
		"prohibited": "high",
		"high_risk":  "medium",
		"tbml":       "high",
		"fraud":      "high",
		"sanctions":  "critical",
	}

	severity := baseSeverity[category]
	if severity == "" {
		severity = "medium"
	}

	// Adjust based on match count
	if matchCount > 3 {
		switch severity {
		case "low":
			severity = "medium"
		case "medium":
			severity = "high"
		case "high":
			severity = "critical"
		}
	}

	return severity
}

// calculateRiskConfidence calculates confidence based on matches
func (km *KeywordMatcher) calculateRiskConfidence(matches []string) float64 {
	if len(matches) == 0 {
		return 0.0
	}

	// Base confidence
	confidence := 0.7

	// Boost confidence for multiple matches
	if len(matches) > 1 {
		confidence += 0.1 * float64(len(matches)-1)
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// generateRiskDescription generates a description for the detected risk
func (km *KeywordMatcher) generateRiskDescription(category string, matches []string) string {
	descriptions := map[string]string{
		"illegal":    "Illegal activities detected",
		"prohibited": "Prohibited business activities detected",
		"high_risk":  "High-risk business activities detected",
		"tbml":       "Trade-based money laundering indicators detected",
		"fraud":      "Fraud indicators detected",
		"sanctions":  "Sanctions-related keywords detected",
	}

	baseDescription := descriptions[category]
	if baseDescription == "" {
		baseDescription = "Risk indicators detected"
	}

	if len(matches) > 0 {
		baseDescription += fmt.Sprintf(": %s", strings.Join(matches[:min(3, len(matches))], ", "))
		if len(matches) > 3 {
			baseDescription += "..."
		}
	}

	return baseDescription
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
