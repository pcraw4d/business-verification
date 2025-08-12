package webanalysis

import (
	"context"
	"fmt"
	"strings"
)

// IndustryClassifier handles industry classification with multi-industry support
type IndustryClassifier struct {
	classifiers     map[string]IndustryClassifierRule
	keywords        map[string][]string
	confidenceRules []ConfidenceRule
}

// IndustryClassifierRule represents a classification rule
type IndustryClassifierRule struct {
	Industry   string
	Keywords   []string
	Weight     float64
	Confidence float64
	NAICSCode  string
	SICCode    string
}

// ConfidenceRule represents a confidence calculation rule
type ConfidenceRule struct {
	Condition     string
	Multiplier    float64
	MinConfidence float64
}

// NewIndustryClassifier creates a new industry classifier
func NewIndustryClassifier() *IndustryClassifier {
	ic := &IndustryClassifier{
		classifiers: make(map[string]IndustryClassifierRule),
		keywords:    make(map[string][]string),
	}

	// Initialize with basic industry classifiers
	ic.initializeClassifiers()

	return ic
}

// ClassifyContent performs industry classification on content
func (ic *IndustryClassifier) ClassifyContent(ctx context.Context, content string, maxResults int) ([]IndustryClassification, error) {
	if maxResults == 0 {
		maxResults = 3 // Default to top 3
	}

	// Normalize content
	normalizedContent := strings.ToLower(content)

	// Calculate scores for each industry
	scores := make(map[string]float64)
	evidence := make(map[string]string)

	for industry, rule := range ic.classifiers {
		score := ic.calculateIndustryScore(normalizedContent, rule)
		if score > 0 {
			scores[industry] = score
			evidence[industry] = ic.findEvidence(normalizedContent, rule)
		}
	}

	// Sort industries by score and get top results
	topIndustries := ic.getTopIndustries(scores, maxResults)

	// Create classification results
	var results []IndustryClassification
	for _, industry := range topIndustries {
		rule := ic.classifiers[industry]
		confidence := ic.calculateConfidence(scores[industry], rule, normalizedContent)

		classification := IndustryClassification{
			Industry:   industry,
			NAICSCode:  rule.NAICSCode,
			SICCode:    rule.SICCode,
			Confidence: confidence,
			Evidence:   evidence[industry],
			Keywords:   rule.Keywords,
		}

		results = append(results, classification)
	}

	return results, nil
}

// calculateIndustryScore calculates the score for an industry
func (ic *IndustryClassifier) calculateIndustryScore(content string, rule IndustryClassifierRule) float64 {
	score := 0.0

	// Check for keyword matches
	for _, keyword := range rule.Keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			score += rule.Weight
		}
	}

	// Apply confidence multiplier
	score *= rule.Confidence

	return score
}

// findEvidence finds evidence for the classification
func (ic *IndustryClassifier) findEvidence(content string, rule IndustryClassifierRule) string {
	var foundKeywords []string

	for _, keyword := range rule.Keywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			foundKeywords = append(foundKeywords, keyword)
		}
	}

	if len(foundKeywords) > 0 {
		return fmt.Sprintf("Found keywords: %s", strings.Join(foundKeywords, ", "))
	}

	return "No specific keywords found"
}

// getTopIndustries gets the top industries by score
func (ic *IndustryClassifier) getTopIndustries(scores map[string]float64, maxResults int) []string {
	// Create a slice of industries with scores
	type industryScore struct {
		industry string
		score    float64
	}

	var industryScores []industryScore
	for industry, score := range scores {
		industryScores = append(industryScores, industryScore{industry, score})
	}

	// Sort by score (descending)
	for i := 0; i < len(industryScores); i++ {
		for j := i + 1; j < len(industryScores); j++ {
			if industryScores[i].score < industryScores[j].score {
				industryScores[i], industryScores[j] = industryScores[j], industryScores[i]
			}
		}
	}

	// Get top results
	var topIndustries []string
	for i, is := range industryScores {
		if i >= maxResults {
			break
		}
		topIndustries = append(topIndustries, is.industry)
	}

	return topIndustries
}

// calculateConfidence calculates the confidence score
func (ic *IndustryClassifier) calculateConfidence(score float64, rule IndustryClassifierRule, content string) float64 {
	confidence := score * rule.Confidence

	// Apply confidence rules
	for _, confidenceRule := range ic.confidenceRules {
		if ic.matchesCondition(content, confidenceRule.Condition) {
			confidence *= confidenceRule.Multiplier
		}
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// matchesCondition checks if content matches a condition
func (ic *IndustryClassifier) matchesCondition(content, condition string) bool {
	// Simple condition matching for now
	return strings.Contains(content, strings.ToLower(condition))
}

// initializeClassifiers initializes the industry classifiers
func (ic *IndustryClassifier) initializeClassifiers() {
	// Technology
	ic.classifiers["Technology"] = IndustryClassifierRule{
		Industry:   "Technology",
		Keywords:   []string{"software", "technology", "digital", "app", "platform", "saas", "cloud", "ai", "machine learning", "data", "analytics"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "511210",
		SICCode:    "7372",
	}

	// Financial Services
	ic.classifiers["Financial Services"] = IndustryClassifierRule{
		Industry:   "Financial Services",
		Keywords:   []string{"bank", "financial", "investment", "insurance", "credit", "loan", "mortgage", "trading", "wealth", "asset"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "522110",
		SICCode:    "6021",
	}

	// Healthcare
	ic.classifiers["Healthcare"] = IndustryClassifierRule{
		Industry:   "Healthcare",
		Keywords:   []string{"health", "medical", "pharmaceutical", "hospital", "clinic", "doctor", "patient", "treatment", "medicine", "therapy"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "621111",
		SICCode:    "8011",
	}

	// Manufacturing
	ic.classifiers["Manufacturing"] = IndustryClassifierRule{
		Industry:   "Manufacturing",
		Keywords:   []string{"manufacturing", "factory", "production", "industrial", "machinery", "equipment", "assembly", "supply chain"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "332996",
		SICCode:    "3499",
	}

	// Retail
	ic.classifiers["Retail"] = IndustryClassifierRule{
		Industry:   "Retail",
		Keywords:   []string{"retail", "store", "shop", "commerce", "ecommerce", "online store", "marketplace", "consumer", "product"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "454110",
		SICCode:    "5961",
	}

	// Consulting
	ic.classifiers["Consulting"] = IndustryClassifierRule{
		Industry:   "Consulting",
		Keywords:   []string{"consulting", "advisory", "strategy", "management", "business", "professional services", "expertise"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "541611",
		SICCode:    "8742",
	}

	// Real Estate
	ic.classifiers["Real Estate"] = IndustryClassifierRule{
		Industry:   "Real Estate",
		Keywords:   []string{"real estate", "property", "realty", "broker", "agent", "development", "construction", "housing"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "531210",
		SICCode:    "6531",
	}

	// Education
	ic.classifiers["Education"] = IndustryClassifierRule{
		Industry:   "Education",
		Keywords:   []string{"education", "school", "university", "college", "training", "learning", "academic", "student"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "611110",
		SICCode:    "8221",
	}

	// Transportation
	ic.classifiers["Transportation"] = IndustryClassifierRule{
		Industry:   "Transportation",
		Keywords:   []string{"transportation", "logistics", "shipping", "delivery", "freight", "trucking", "warehouse", "supply chain"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "484121",
		SICCode:    "4213",
	}

	// Energy
	ic.classifiers["Energy"] = IndustryClassifierRule{
		Industry:   "Energy",
		Keywords:   []string{"energy", "oil", "gas", "renewable", "solar", "wind", "power", "utility", "electricity"},
		Weight:     1.0,
		Confidence: 0.8,
		NAICSCode:  "221111",
		SICCode:    "4911",
	}

	// Initialize confidence rules
	ic.confidenceRules = []ConfidenceRule{
		{
			Condition:     "about us",
			Multiplier:    1.2,
			MinConfidence: 0.1,
		},
		{
			Condition:     "services",
			Multiplier:    1.1,
			MinConfidence: 0.1,
		},
		{
			Condition:     "products",
			Multiplier:    1.1,
			MinConfidence: 0.1,
		},
	}
}
