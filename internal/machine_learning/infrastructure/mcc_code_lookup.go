package infrastructure

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
)

// MCCCodeLookup handles MCC code lookup and validation
type MCCCodeLookup struct {
	// MCC code database
	mccCodes map[string]*MCCCodeInfo

	// Prohibited MCC codes
	prohibitedMCCs map[string]bool

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger
}

// MCCCodeInfo represents MCC code information
type MCCCodeInfo struct {
	Code         string `json:"code"`
	Description  string `json:"description"`
	Category     string `json:"category"`
	IsProhibited bool   `json:"is_prohibited"`
	RiskLevel    string `json:"risk_level"` // low, medium, high, critical
}

// NewMCCCodeLookup creates a new MCC code lookup
func NewMCCCodeLookup(logger *log.Logger) *MCCCodeLookup {
	if logger == nil {
		logger = log.Default()
	}

	return &MCCCodeLookup{
		mccCodes:       make(map[string]*MCCCodeInfo),
		prohibitedMCCs: make(map[string]bool),
		logger:         logger,
	}
}

// Initialize initializes the MCC code lookup with MCC code database
func (mcl *MCCCodeLookup) Initialize(ctx context.Context) error {
	mcl.mu.Lock()
	defer mcl.mu.Unlock()

	mcl.logger.Printf("📋 Initializing MCC Code Lookup")

	// Load MCC codes
	if err := mcl.loadMCCCodes(); err != nil {
		return fmt.Errorf("failed to load MCC codes: %w", err)
	}

	// Load prohibited MCC codes
	if err := mcl.loadProhibitedMCCCodes(); err != nil {
		return fmt.Errorf("failed to load prohibited MCC codes: %w", err)
	}

	mcl.logger.Printf("✅ MCC Code Lookup initialized with %d MCC codes (%d prohibited)",
		len(mcl.mccCodes), len(mcl.prohibitedMCCs))

	return nil
}

// ClassifyByMCC performs classification using MCC code lookup
func (mcl *MCCCodeLookup) ClassifyByMCC(ctx context.Context, businessName, description string) ([]ClassificationPrediction, error) {
	mcl.mu.RLock()
	defer mcl.mu.RUnlock()

	// Combine business name and description for analysis
	text := strings.ToLower(businessName + " " + description)

	var predictions []ClassificationPrediction
	industryScores := make(map[string]float64)

	// Match business description to MCC codes
	for code, mccInfo := range mcl.mccCodes {
		score := mcl.calculateMCCScore(text, mccInfo)
		if score > 0 {
			// Map MCC category to industry
			industry := mcl.mapMCCToIndustry(mccInfo.Category)
			if industryScores[industry] < score {
				industryScores[industry] = score
			}
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

// CheckMCCRestrictions checks for MCC code restrictions and risks
func (mcl *MCCCodeLookup) CheckMCCRestrictions(ctx context.Context, businessName, description string) ([]DetectedRisk, error) {
	mcl.mu.RLock()
	defer mcl.mu.RUnlock()

	// Combine business name and description for analysis
	text := strings.ToLower(businessName + " " + description)

	var detectedRisks []DetectedRisk

	// Check for prohibited MCC codes
	for code, mccInfo := range mcl.mccCodes {
		if mccInfo.IsProhibited {
			score := mcl.calculateMCCScore(text, mccInfo)
			if score > 0.5 { // High threshold for prohibited activities
				risk := DetectedRisk{
					Category:    "prohibited",
					Severity:    mcl.mapRiskLevelToSeverity(mccInfo.RiskLevel),
					Confidence:  score,
					Keywords:    []string{mccInfo.Description},
					Description: fmt.Sprintf("Prohibited MCC code %s detected: %s", code, mccInfo.Description),
				}
				detectedRisks = append(detectedRisks, risk)
			}
		}
	}

	// Check for high-risk MCC codes
	for code, mccInfo := range mcl.mccCodes {
		if mccInfo.RiskLevel == "high" || mccInfo.RiskLevel == "critical" {
			score := mcl.calculateMCCScore(text, mccInfo)
			if score > 0.3 { // Lower threshold for high-risk activities
				risk := DetectedRisk{
					Category:    "high_risk",
					Severity:    mcl.mapRiskLevelToSeverity(mccInfo.RiskLevel),
					Confidence:  score,
					Keywords:    []string{mccInfo.Description},
					Description: fmt.Sprintf("High-risk MCC code %s detected: %s", code, mccInfo.Description),
				}
				detectedRisks = append(detectedRisks, risk)
			}
		}
	}

	return detectedRisks, nil
}

// GetMCCInfo returns information about a specific MCC code
func (mcl *MCCCodeLookup) GetMCCInfo(code string) (*MCCCodeInfo, bool) {
	mcl.mu.RLock()
	defer mcl.mu.RUnlock()

	info, exists := mcl.mccCodes[code]
	return info, exists
}

// IsProhibitedMCC checks if an MCC code is prohibited
func (mcl *MCCCodeLookup) IsProhibitedMCC(code string) bool {
	mcl.mu.RLock()
	defer mcl.mu.RUnlock()

	return mcl.prohibitedMCCs[code]
}

// HealthCheck performs a health check on the MCC code lookup
func (mcl *MCCCodeLookup) HealthCheck(ctx context.Context) error {
	mcl.mu.RLock()
	defer mcl.mu.RUnlock()

	// Check if MCC codes are loaded
	if len(mcl.mccCodes) == 0 {
		return fmt.Errorf("MCC codes not loaded")
	}

	// Check if prohibited MCC codes are loaded
	if len(mcl.prohibitedMCCs) == 0 {
		return fmt.Errorf("prohibited MCC codes not loaded")
	}

	return nil
}

// loadMCCCodes loads MCC codes from the database
func (mcl *MCCCodeLookup) loadMCCCodes() error {
	// This would typically load from the Supabase database
	// For now, we'll use a comprehensive set of MCC codes

	mcl.mccCodes = map[string]*MCCCodeInfo{
		// Technology
		"5734": {"5734", "Computer Software Stores", "technology", false, "low"},
		"7372": {"7372", "Computer Programming Services", "technology", false, "low"},
		"7379": {"7379", "Computer Maintenance and Repair Services", "technology", false, "low"},
		"4816": {"4816", "Computer Network/Information Services", "technology", false, "low"},

		// Finance
		"6010": {"6010", "Financial Institutions - Manual Cash Disbursements", "finance", false, "medium"},
		"6011": {"6011", "Financial Institutions - Automated Cash Disbursements", "finance", false, "medium"},
		"6012": {"6012", "Financial Institutions - Merchandise and Services", "finance", false, "medium"},
		"6051": {"6051", "Non-Financial Institutions - Foreign Currency", "finance", false, "high"},

		// Healthcare
		"8011": {"8011", "Doctors and Physicians", "healthcare", false, "low"},
		"8021": {"8021", "Dentists and Orthodontists", "healthcare", false, "low"},
		"8041": {"8041", "Chiropractors", "healthcare", false, "low"},
		"8062": {"8062", "Hospitals", "healthcare", false, "low"},

		// Retail
		"5310": {"5310", "Discount Stores", "retail", false, "low"},
		"5311": {"5311", "Department Stores", "retail", false, "low"},
		"5411": {"5411", "Grocery Stores, Supermarkets", "retail", false, "low"},
		"5999": {"5999", "Miscellaneous and Specialty Retail", "retail", false, "low"},

		// Manufacturing
		"5085": {"5085", "Industrial Supplies", "manufacturing", false, "low"},
		"5087": {"5087", "Service Establishment Equipment", "manufacturing", false, "low"},
		"5094": {"5094", "Jewelry, Watches, Silverware, and Other Precious Metals", "manufacturing", false, "medium"},

		// Education
		"8220": {"8220", "Colleges, Universities, Professional Schools", "education", false, "low"},
		"8241": {"8241", "Correspondence Schools", "education", false, "low"},
		"8244": {"8244", "Business and Secretarial Schools", "education", false, "low"},

		// Real Estate
		"6513": {"6513", "Real Estate Agents and Managers", "real_estate", false, "low"},
		"7011": {"7011", "Hotels, Motels, and Resorts", "real_estate", false, "low"},
		"7012": {"7012", "Timeshares", "real_estate", false, "medium"},

		// Consulting
		"7392": {"7392", "Management, Consulting, and Public Relations Services", "consulting", false, "low"},
		"7393": {"7393", "Detective Agencies, Protective Services", "consulting", false, "medium"},

		// Media
		"4812": {"4812", "Telecommunications Equipment", "media", false, "low"},
		"4813": {"4813", "Telecommunications Services", "media", false, "low"},
		"5993": {"5993", "Cigar Stores and Stands", "media", false, "medium"},

		// Transportation
		"4111": {"4111", "Local and Suburban Commuter Passenger Transportation", "transportation", false, "low"},
		"4119": {"4119", "Ambulance Services", "transportation", false, "low"},
		"4121": {"4121", "Taxicabs and Limousines", "transportation", false, "low"},

		// Prohibited MCC codes
		"7995": {"7995", "Betting, including Lottery Tickets, Casino Gaming Chips, Off-Track Betting", "gambling", true, "critical"},
		"7273": {"7273", "Dating Services", "adult_entertainment", true, "high"},
		"7841": {"7841", "Video Tape Rental Stores", "adult_entertainment", true, "high"},
		"5967": {"5967", "Direct Marketing - Continuity/Subscription Merchants", "high_risk", true, "high"},
		"6010": {"6010", "Financial Institutions - Manual Cash Disbursements", "money_services", true, "high"},
		"6051": {"6051", "Non-Financial Institutions - Foreign Currency", "money_services", true, "critical"},

		// High-risk MCC codes
		"5993": {"5993", "Cigar Stores and Stands", "tobacco", false, "high"},
		"5921": {"5921", "Package Stores - Beer, Wine, and Liquor", "alcohol", false, "high"},
		"7012": {"7012", "Timeshares", "travel", false, "high"},
		"5094": {"5094", "Jewelry, Watches, Silverware, and Other Precious Metals", "precious_metals", false, "high"},
	}

	return nil
}

// loadProhibitedMCCCodes loads prohibited MCC codes
func (mcl *MCCCodeLookup) loadProhibitedMCCCodes() error {
	// Load prohibited MCC codes from the MCC codes database
	for code, mccInfo := range mcl.mccCodes {
		if mccInfo.IsProhibited {
			mcl.prohibitedMCCs[code] = true
		}
	}

	return nil
}

// calculateMCCScore calculates the score for MCC code matching
func (mcl *MCCCodeLookup) calculateMCCScore(text string, mccInfo *MCCCodeInfo) float64 {
	// Simple keyword matching against MCC description
	description := strings.ToLower(mccInfo.Description)
	textWords := strings.Fields(text)
	descriptionWords := strings.Fields(description)

	var matchCount int
	var totalWords int

	// Count word matches
	for _, textWord := range textWords {
		totalWords++
		for _, descWord := range descriptionWords {
			if strings.Contains(textWord, descWord) || strings.Contains(descWord, textWord) {
				matchCount++
				break
			}
		}
	}

	if totalWords == 0 {
		return 0.0
	}

	// Calculate score based on match ratio
	score := float64(matchCount) / float64(totalWords)

	// Boost score for exact phrase matches
	if strings.Contains(text, description) {
		score += 0.3
	}

	// Cap at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// mapMCCToIndustry maps MCC category to industry
func (mcl *MCCCodeLookup) mapMCCToIndustry(category string) string {
	industryMap := map[string]string{
		"technology":          "technology",
		"finance":             "finance",
		"healthcare":          "healthcare",
		"retail":              "retail",
		"manufacturing":       "manufacturing",
		"education":           "education",
		"real_estate":         "real_estate",
		"consulting":          "consulting",
		"media":               "media",
		"transportation":      "transportation",
		"gambling":            "gambling",
		"adult_entertainment": "adult_entertainment",
		"high_risk":           "high_risk",
		"money_services":      "money_services",
		"tobacco":             "tobacco",
		"alcohol":             "alcohol",
		"travel":              "travel",
		"precious_metals":     "precious_metals",
	}

	industry, exists := industryMap[category]
	if !exists {
		industry = "other"
	}

	return industry
}

// mapRiskLevelToSeverity maps MCC risk level to severity
func (mcl *MCCCodeLookup) mapRiskLevelToSeverity(riskLevel string) string {
	severityMap := map[string]string{
		"low":      "low",
		"medium":   "medium",
		"high":     "high",
		"critical": "critical",
	}

	severity, exists := severityMap[riskLevel]
	if !exists {
		severity = "medium"
	}

	return severity
}
