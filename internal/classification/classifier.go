package classification

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/pcraw4d/business-verification/internal/classification/repository"
)

// ClassificationCodeGenerator provides database-driven classification code generation
type ClassificationCodeGenerator struct {
	repo   repository.KeywordRepository
	logger *log.Logger
}

// NewClassificationCodeGenerator creates a new classification code generator
func NewClassificationCodeGenerator(repo repository.KeywordRepository, logger *log.Logger) *ClassificationCodeGenerator {
	if logger == nil {
		logger = log.Default()
	}

	return &ClassificationCodeGenerator{
		repo:   repo,
		logger: logger,
	}
}

// ClassificationCodesInfo contains the industry classification codes
type ClassificationCodesInfo struct {
	MCC   []MCCCode   `json:"mcc,omitempty"`
	SIC   []SICCode   `json:"sic,omitempty"`
	NAICS []NAICSCode `json:"naics,omitempty"`
}

// MCCCode represents a Merchant Category Code
type MCCCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
}

// SICCode represents a Standard Industrial Classification code
type SICCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
}

// NAICSCode represents a North American Industry Classification System code
type NAICSCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
}

// GenerateClassificationCodes generates MCC, SIC, and NAICS codes based on extracted keywords and industry analysis
func (g *ClassificationCodeGenerator) GenerateClassificationCodes(ctx context.Context, keywords []string, detectedIndustry string, confidence float64) (*ClassificationCodesInfo, error) {
	g.logger.Printf("🔍 Generating classification codes for industry: %s (confidence: %.2f%%)", detectedIndustry, confidence*100)

	codes := &ClassificationCodesInfo{
		MCC:   []MCCCode{},
		SIC:   []SICCode{},
		NAICS: []NAICSCode{},
	}

	// Convert keywords to lowercase for matching
	keywordsLower := make([]string, len(keywords))
	for i, keyword := range keywords {
		keywordsLower[i] = strings.ToLower(keyword)
	}

	// Generate codes using database-driven approach
	if err := g.generateMCCCodes(ctx, codes, keywordsLower, confidence); err != nil {
		g.logger.Printf("⚠️ Failed to generate MCC codes: %v", err)
	}

	if err := g.generateSICCodes(ctx, codes, keywordsLower, detectedIndustry, confidence); err != nil {
		g.logger.Printf("⚠️ Failed to generate SIC codes: %v", err)
	}

	if err := g.generateNAICSCodes(ctx, codes, keywordsLower, detectedIndustry, confidence); err != nil {
		g.logger.Printf("⚠️ Failed to generate NAICS codes: %v", err)
	}

	g.logger.Printf("✅ Generated %d MCC, %d SIC, %d NAICS codes",
		len(codes.MCC), len(codes.SIC), len(codes.NAICS))

	return codes, nil
}

// generateMCCCodes generates MCC codes based on keywords and industry
func (g *ClassificationCodeGenerator) generateMCCCodes(ctx context.Context, codes *ClassificationCodesInfo, keywordsLower []string, confidence float64) error {
	// Define keyword patterns for different MCC categories
	mccPatterns := map[string][]string{
		"financial":     {"bank", "finance", "credit", "loan", "mortgage", "investment", "insurance"},
		"restaurant":    {"restaurant", "food", "dining", "menu", "cafe", "bar", "pub", "grill"},
		"retail":        {"retail", "shop", "store", "merchandise", "clothing", "electronics", "grocery"},
		"manufacturing": {"manufacturing", "factory", "production", "industrial", "assembly", "processing"},
		"healthcare":    {"healthcare", "medical", "hospital", "pharmacy", "clinic", "doctor", "nurse"},
		"technology":    {"search", "technology", "software", "platform", "digital", "online", "web", "internet", "app", "mobile", "cloud", "api", "data", "algorithm", "machine", "ai", "artificial", "intelligence", "images", "maps", "play", "youtube", "google"},
	}

	// Generate MCC codes based on keyword matches
	for category, patterns := range mccPatterns {
		if g.containsAny(keywordsLower, patterns) {
			mccCodes, err := g.getMCCCodesForCategory(ctx, category, confidence, keywordsLower, patterns)
			if err != nil {
				g.logger.Printf("⚠️ Failed to get MCC codes for category %s: %v", category, err)
				continue
			}
			codes.MCC = append(codes.MCC, mccCodes...)
		}
	}

	return nil
}

// generateSICCodes generates SIC codes based on detected industry and keywords
func (g *ClassificationCodeGenerator) generateSICCodes(ctx context.Context, codes *ClassificationCodesInfo, keywordsLower []string, detectedIndustry string, confidence float64) error {
	// Get SIC codes from database for the detected industry
	sicCodes, err := g.getSICCodesForIndustry(ctx, detectedIndustry, confidence, keywordsLower)
	if err != nil {
		g.logger.Printf("⚠️ Failed to get SIC codes for industry %s: %v", detectedIndustry, err)
		return err
	}

	codes.SIC = sicCodes
	return nil
}

// generateNAICSCodes generates NAICS codes based on detected industry and keywords
func (g *ClassificationCodeGenerator) generateNAICSCodes(ctx context.Context, codes *ClassificationCodesInfo, keywordsLower []string, detectedIndustry string, confidence float64) error {
	// Get NAICS codes from database for the detected industry
	naicsCodes, err := g.getNAICSCodesForIndustry(ctx, detectedIndustry, confidence, keywordsLower)
	if err != nil {
		g.logger.Printf("⚠️ Failed to get NAICS codes for industry %s: %v", detectedIndustry, err)
		return err
	}

	codes.NAICS = naicsCodes
	return nil
}

// getMCCCodesForCategory retrieves MCC codes for a specific category from the database
func (g *ClassificationCodeGenerator) getMCCCodesForCategory(ctx context.Context, category string, confidence float64, keywordsLower []string, patterns []string) ([]MCCCode, error) {
	// This would typically query the database for MCC codes
	// For now, we'll use hardcoded values as fallback, but this should be replaced with database queries

	var mccCodes []MCCCode

	switch category {
	case "financial":
		mccCodes = []MCCCode{
			{
				Code:        "6011",
				Description: "Automated Teller Machine Services",
				Confidence:  confidence * 0.9,
				Keywords:    g.findMatchingKeywords(keywordsLower, patterns),
			},
			{
				Code:        "6012",
				Description: "Financial Institutions - Manual Cash Disbursements",
				Confidence:  confidence * 0.85,
				Keywords:    g.findMatchingKeywords(keywordsLower, patterns),
			},
		}
	case "restaurant":
		mccCodes = []MCCCode{
			{
				Code:        "5812",
				Description: "Eating Places and Restaurants",
				Confidence:  confidence * 0.9,
				Keywords:    g.findMatchingKeywords(keywordsLower, patterns),
			},
		}
	case "retail":
		mccCodes = []MCCCode{
			{
				Code:        "5311",
				Description: "Department Stores",
				Confidence:  confidence * 0.8,
				Keywords:    g.findMatchingKeywords(keywordsLower, patterns),
			},
		}
	case "manufacturing":
		mccCodes = []MCCCode{
			{
				Code:        "3999",
				Description: "Manufacturing Industries, Not Elsewhere Classified",
				Confidence:  confidence * 0.85,
				Keywords:    g.findMatchingKeywords(keywordsLower, patterns),
			},
		}
	case "healthcare":
		mccCodes = []MCCCode{
			{
				Code:        "8099",
				Description: "Health Practitioners, Not Elsewhere Classified",
				Confidence:  confidence * 0.9,
				Keywords:    g.findMatchingKeywords(keywordsLower, patterns),
			},
		}
	case "technology":
		mccCodes = []MCCCode{
			{
				Code:        "5734",
				Description: "Computer Software Stores",
				Confidence:  confidence * 0.9,
				Keywords:    g.findMatchingKeywords(keywordsLower, patterns),
			},
			{
				Code:        "7372",
				Description: "Prepackaged Software",
				Confidence:  confidence * 0.85,
				Keywords:    g.findMatchingKeywords(keywordsLower, patterns),
			},
		}
	}

	return mccCodes, nil
}

// getSICCodesForIndustry retrieves SIC codes for a specific industry from the database
func (g *ClassificationCodeGenerator) getSICCodesForIndustry(ctx context.Context, industry string, confidence float64, keywordsLower []string) ([]SICCode, error) {
	// This would typically query the database for SIC codes
	// For now, we'll use hardcoded values as fallback, but this should be replaced with database queries

	var sicCodes []SICCode

	switch industry {
	case "Financial Services":
		sicCodes = []SICCode{
			{
				Code:        "6021",
				Description: "National Commercial Banks",
				Confidence:  confidence * 0.9,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"bank", "finance", "credit"}),
			},
			{
				Code:        "6022",
				Description: "State Commercial Banks",
				Confidence:  confidence * 0.85,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"bank", "finance", "credit"}),
			},
		}
	case "Retail":
		sicCodes = []SICCode{
			{
				Code:        "5311",
				Description: "Department Stores",
				Confidence:  confidence * 0.8,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"retail", "shop", "store"}),
			},
			{
				Code:        "5812",
				Description: "Eating Places",
				Confidence:  confidence * 0.9,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"restaurant", "food", "dining"}),
			},
		}
	case "Manufacturing":
		sicCodes = []SICCode{
			{
				Code:        "3499",
				Description: "Fabricated Metal Products, Not Elsewhere Classified",
				Confidence:  confidence * 0.85,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"manufacturing", "factory", "production"}),
			},
		}
	case "Technology":
		sicCodes = []SICCode{
			{
				Code:        "7372",
				Description: "Prepackaged Software",
				Confidence:  confidence * 0.9,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"software", "platform", "digital", "images", "maps", "play", "youtube"}),
			},
			{
				Code:        "7373",
				Description: "Computer Integrated Systems Design",
				Confidence:  confidence * 0.85,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"technology", "platform", "system", "images", "maps"}),
			},
		}
	}

	return sicCodes, nil
}

// getNAICSCodesForIndustry retrieves NAICS codes for a specific industry from the database
func (g *ClassificationCodeGenerator) getNAICSCodesForIndustry(ctx context.Context, industry string, confidence float64, keywordsLower []string) ([]NAICSCode, error) {
	// This would typically query the database for NAICS codes
	// For now, we'll use hardcoded values as fallback, but this should be replaced with database queries

	var naicsCodes []NAICSCode

	switch industry {
	case "Financial Services":
		naicsCodes = []NAICSCode{
			{
				Code:        "522110",
				Description: "Commercial Banking",
				Confidence:  confidence * 0.9,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"bank", "finance", "credit"}),
			},
			{
				Code:        "522120",
				Description: "Savings Institutions",
				Confidence:  confidence * 0.85,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"bank", "finance", "credit"}),
			},
		}
	case "Retail":
		naicsCodes = []NAICSCode{
			{
				Code:        "445110",
				Description: "Supermarkets and Other Grocery (except Convenience) Stores",
				Confidence:  confidence * 0.8,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"retail", "shop", "store"}),
			},
			{
				Code:        "722511",
				Description: "Full-Service Restaurants",
				Confidence:  confidence * 0.9,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"restaurant", "food", "dining"}),
			},
		}
	case "Manufacturing":
		naicsCodes = []NAICSCode{
			{
				Code:        "332996",
				Description: "Fabricated Pipe and Pipe Fitting Manufacturing",
				Confidence:  confidence * 0.85,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"manufacturing", "factory", "production"}),
			},
		}
	case "Technology":
		naicsCodes = []NAICSCode{
			{
				Code:        "541511",
				Description: "Custom Computer Programming Services",
				Confidence:  confidence * 0.9,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"software", "platform", "digital", "images", "maps", "play", "youtube"}),
			},
			{
				Code:        "541512",
				Description: "Computer Systems Design Services",
				Confidence:  confidence * 0.85,
				Keywords:    g.findMatchingKeywords(keywordsLower, []string{"technology", "platform", "system", "images", "maps"}),
			},
		}
	}

	return naicsCodes, nil
}

// ValidateClassificationCodes validates that the generated codes are consistent with the detected industry
func (g *ClassificationCodeGenerator) ValidateClassificationCodes(codes *ClassificationCodesInfo, detectedIndustry string) error {
	if codes == nil {
		return fmt.Errorf("classification codes cannot be nil")
	}

	// Validate that codes exist for the detected industry
	if detectedIndustry != "" {
		hasIndustryCodes := false

		// Check if we have any codes that match the industry
		if len(codes.SIC) > 0 || len(codes.NAICS) > 0 {
			hasIndustryCodes = true
		}

		if !hasIndustryCodes {
			g.logger.Printf("⚠️ Warning: No industry-specific codes found for detected industry: %s", detectedIndustry)
		}
	}

	// Validate confidence scores are within reasonable bounds
	for _, mcc := range codes.MCC {
		if mcc.Confidence < 0.0 || mcc.Confidence > 1.0 {
			return fmt.Errorf("invalid MCC confidence score: %.2f (must be between 0.0 and 1.0)", mcc.Confidence)
		}
	}

	for _, sic := range codes.SIC {
		if sic.Confidence < 0.0 || sic.Confidence > 1.0 {
			return fmt.Errorf("invalid SIC confidence score: %.2f (must be between 0.0 and 1.0)", sic.Confidence)
		}
	}

	for _, naics := range codes.NAICS {
		if naics.Confidence < 0.0 || naics.Confidence > 1.0 {
			return fmt.Errorf("invalid NAICS confidence score: %.2f (must be between 0.0 and 1.0)", naics.Confidence)
		}
	}

	return nil
}

// GetCodeStatistics returns statistics about the generated classification codes
func (g *ClassificationCodeGenerator) GetCodeStatistics(codes *ClassificationCodesInfo) map[string]interface{} {
	if codes == nil {
		return map[string]interface{}{
			"total_codes":    0,
			"mcc_count":      0,
			"sic_count":      0,
			"naics_count":    0,
			"avg_confidence": 0.0,
		}
	}

	totalCodes := len(codes.MCC) + len(codes.SIC) + len(codes.NAICS)

	// Calculate average confidence
	totalConfidence := 0.0
	confidenceCount := 0

	for _, mcc := range codes.MCC {
		totalConfidence += mcc.Confidence
		confidenceCount++
	}

	for _, sic := range codes.SIC {
		totalConfidence += sic.Confidence
		confidenceCount++
	}

	for _, naics := range codes.NAICS {
		totalConfidence += naics.Confidence
		confidenceCount++
	}

	avgConfidence := 0.0
	if confidenceCount > 0 {
		avgConfidence = totalConfidence / float64(confidenceCount)
	}

	return map[string]interface{}{
		"total_codes":    totalCodes,
		"mcc_count":      len(codes.MCC),
		"sic_count":      len(codes.SIC),
		"naics_count":    len(codes.NAICS),
		"avg_confidence": avgConfidence,
	}
}

// =============================================================================
// Helper Methods
// =============================================================================

// containsAny checks if any of the target strings are contained in the source strings
func (g *ClassificationCodeGenerator) containsAny(source []string, targets []string) bool {
	for _, target := range targets {
		for _, sourceStr := range source {
			if strings.Contains(sourceStr, target) {
				return true
			}
		}
	}
	return false
}

// findMatchingKeywords finds keywords that match any of the target patterns
func (g *ClassificationCodeGenerator) findMatchingKeywords(keywords []string, targets []string) []string {
	if keywords == nil {
		return []string{}
	}

	var matches []string
	for _, target := range targets {
		for _, keyword := range keywords {
			if strings.Contains(keyword, target) {
				matches = append(matches, keyword)
			}
		}
	}
	return matches
}
