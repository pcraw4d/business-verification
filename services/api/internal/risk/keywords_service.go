package risk

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/lib/pq"
)

// RiskKeywordsService provides business logic for risk keyword management
type RiskKeywordsService struct {
	db     *sql.DB
	logger *log.Logger
}

// RiskKeyword represents a risk keyword entry
type RiskKeyword struct {
	ID                    int       `json:"id"`
	Keyword               string    `json:"keyword"`
	RiskCategory          string    `json:"risk_category"`
	RiskSeverity          string    `json:"risk_severity"`
	Description           string    `json:"description"`
	MCCCodes              []string  `json:"mcc_codes"`
	NAICSCodes            []string  `json:"naics_codes"`
	SICCodes              []string  `json:"sic_codes"`
	CardBrandRestrictions []string  `json:"card_brand_restrictions"`
	DetectionPatterns     []string  `json:"detection_patterns"`
	Synonyms              []string  `json:"synonyms"`
	IsActive              bool      `json:"is_active"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// RiskDetectionResult represents the result of risk detection
type RiskDetectionResult struct {
	DetectedKeywords []string               `json:"detected_keywords"`
	RiskScore        float64                `json:"risk_score"`
	RiskLevel        string                 `json:"risk_level"`
	RiskCategories   []string               `json:"risk_categories"`
	CardRestrictions []string               `json:"card_restrictions"`
	MCCRestrictions  []string               `json:"mcc_restrictions"`
	Confidence       float64                `json:"confidence"`
	Evidence         []string               `json:"evidence"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// NewRiskKeywordsService creates a new risk keywords service
func NewRiskKeywordsService(db *sql.DB, logger *log.Logger) *RiskKeywordsService {
	return &RiskKeywordsService{
		db:     db,
		logger: logger,
	}
}

// DetectRiskKeywords analyzes text content for risk keywords
func (rks *RiskKeywordsService) DetectRiskKeywords(ctx context.Context, content string) (*RiskDetectionResult, error) {
	rks.logger.Printf("🔍 Detecting risk keywords in content (length: %d)", len(content))

	if content == "" {
		return &RiskDetectionResult{
			DetectedKeywords: []string{},
			RiskScore:        0.0,
			RiskLevel:        "low",
			RiskCategories:   []string{},
			CardRestrictions: []string{},
			MCCRestrictions:  []string{},
			Confidence:       1.0,
			Evidence:         []string{},
			Metadata:         map[string]interface{}{"reason": "no_content"},
		}, nil
	}

	// Normalize content for analysis
	normalizedContent := strings.ToLower(strings.TrimSpace(content))

	// Get all active risk keywords
	keywords, err := rks.getAllActiveKeywords(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get risk keywords: %w", err)
	}

	// Detect matches
	detectedKeywords := make([]string, 0)
	riskCategories := make(map[string]bool)
	cardRestrictions := make(map[string]bool)
	mccRestrictions := make(map[string]bool)
	evidence := make([]string, 0)

	var totalRiskScore float64
	var matchCount int

	for _, keyword := range keywords {
		// Check direct keyword match
		if strings.Contains(normalizedContent, strings.ToLower(keyword.Keyword)) {
			detectedKeywords = append(detectedKeywords, keyword.Keyword)
			riskCategories[keyword.RiskCategory] = true

			// Add card brand restrictions
			for _, restriction := range keyword.CardBrandRestrictions {
				cardRestrictions[restriction] = true
			}

			// Add MCC restrictions
			for _, mcc := range keyword.MCCCodes {
				mccRestrictions[mcc] = true
			}

			// Calculate risk score contribution
			riskScore := rks.calculateKeywordRiskScore(keyword)
			totalRiskScore += riskScore
			matchCount++

			evidence = append(evidence, fmt.Sprintf("Direct keyword match: %s (category: %s, severity: %s)",
				keyword.Keyword, keyword.RiskCategory, keyword.RiskSeverity))
		}

		// Check synonym matches
		for _, synonym := range keyword.Synonyms {
			if strings.Contains(normalizedContent, strings.ToLower(synonym)) {
				detectedKeywords = append(detectedKeywords, synonym)
				riskCategories[keyword.RiskCategory] = true

				// Add restrictions
				for _, restriction := range keyword.CardBrandRestrictions {
					cardRestrictions[restriction] = true
				}
				for _, mcc := range keyword.MCCCodes {
					mccRestrictions[mcc] = true
				}

				// Calculate risk score contribution (reduced for synonyms)
				riskScore := rks.calculateKeywordRiskScore(keyword) * 0.8
				totalRiskScore += riskScore
				matchCount++

				evidence = append(evidence, fmt.Sprintf("Synonym match: %s -> %s (category: %s, severity: %s)",
					synonym, keyword.Keyword, keyword.RiskCategory, keyword.RiskSeverity))
			}
		}

		// Check detection patterns (regex-like patterns)
		for _, pattern := range keyword.DetectionPatterns {
			if rks.matchesPattern(normalizedContent, pattern) {
				detectedKeywords = append(detectedKeywords, keyword.Keyword)
				riskCategories[keyword.RiskCategory] = true

				// Add restrictions
				for _, restriction := range keyword.CardBrandRestrictions {
					cardRestrictions[restriction] = true
				}
				for _, mcc := range keyword.MCCCodes {
					mccRestrictions[mcc] = true
				}

				// Calculate risk score contribution
				riskScore := rks.calculateKeywordRiskScore(keyword)
				totalRiskScore += riskScore
				matchCount++

				evidence = append(evidence, fmt.Sprintf("Pattern match: %s (category: %s, severity: %s)",
					keyword.Keyword, keyword.RiskCategory, keyword.RiskSeverity))
			}
		}
	}

	// Calculate final risk score and level
	finalRiskScore := rks.calculateFinalRiskScore(totalRiskScore, matchCount, len(normalizedContent))
	riskLevel := rks.determineRiskLevel(finalRiskScore)

	// Convert maps to slices
	categorySlice := make([]string, 0, len(riskCategories))
	for category := range riskCategories {
		categorySlice = append(categorySlice, category)
	}

	cardRestrictionSlice := make([]string, 0, len(cardRestrictions))
	for restriction := range cardRestrictions {
		cardRestrictionSlice = append(cardRestrictionSlice, restriction)
	}

	mccRestrictionSlice := make([]string, 0, len(mccRestrictions))
	for mcc := range mccRestrictions {
		mccRestrictionSlice = append(mccRestrictionSlice, mcc)
	}

	// Calculate confidence based on content length and match quality
	confidence := rks.calculateConfidence(len(normalizedContent), matchCount, len(detectedKeywords))

	result := &RiskDetectionResult{
		DetectedKeywords: detectedKeywords,
		RiskScore:        finalRiskScore,
		RiskLevel:        riskLevel,
		RiskCategories:   categorySlice,
		CardRestrictions: cardRestrictionSlice,
		MCCRestrictions:  mccRestrictionSlice,
		Confidence:       confidence,
		Evidence:         evidence,
		Metadata: map[string]interface{}{
			"content_length": len(normalizedContent),
			"match_count":    matchCount,
			"keyword_count":  len(detectedKeywords),
			"analysis_time":  time.Now(),
		},
	}

	rks.logger.Printf("✅ Risk detection completed: %d keywords detected, risk level: %s, score: %.2f",
		len(detectedKeywords), riskLevel, finalRiskScore)

	return result, nil
}

// getAllActiveKeywords retrieves all active risk keywords from the database
func (rks *RiskKeywordsService) getAllActiveKeywords(ctx context.Context) ([]*RiskKeyword, error) {
	query := `
		SELECT id, keyword, risk_category, risk_severity, description,
		       mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
		       detection_patterns, synonyms, is_active, created_at, updated_at
		FROM risk_keywords 
		WHERE is_active = true
		ORDER BY risk_severity DESC, risk_category
	`

	rows, err := rks.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk keywords: %w", err)
	}
	defer rows.Close()

	var keywords []*RiskKeyword
	for rows.Next() {
		var keyword RiskKeyword
		var mccCodes, naicsCodes, sicCodes, cardRestrictions, patterns, synonyms pq.StringArray

		err := rows.Scan(
			&keyword.ID,
			&keyword.Keyword,
			&keyword.RiskCategory,
			&keyword.RiskSeverity,
			&keyword.Description,
			&mccCodes,
			&naicsCodes,
			&sicCodes,
			&cardRestrictions,
			&patterns,
			&synonyms,
			&keyword.IsActive,
			&keyword.CreatedAt,
			&keyword.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk keyword: %w", err)
		}

		keyword.MCCCodes = []string(mccCodes)
		keyword.NAICSCodes = []string(naicsCodes)
		keyword.SICCodes = []string(sicCodes)
		keyword.CardBrandRestrictions = []string(cardRestrictions)
		keyword.DetectionPatterns = []string(patterns)
		keyword.Synonyms = []string(synonyms)

		keywords = append(keywords, &keyword)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating risk keywords: %w", err)
	}

	rks.logger.Printf("📊 Loaded %d active risk keywords", len(keywords))
	return keywords, nil
}

// calculateKeywordRiskScore calculates the risk score contribution for a keyword
func (rks *RiskKeywordsService) calculateKeywordRiskScore(keyword *RiskKeyword) float64 {
	baseScore := 0.1 // Base score for any keyword match

	// Adjust based on risk severity
	switch keyword.RiskSeverity {
	case "critical":
		baseScore = 0.4
	case "high":
		baseScore = 0.3
	case "medium":
		baseScore = 0.2
	case "low":
		baseScore = 0.1
	}

	// Adjust based on risk category
	switch keyword.RiskCategory {
	case "illegal":
		baseScore *= 1.5
	case "prohibited":
		baseScore *= 1.3
	case "high_risk":
		baseScore *= 1.2
	case "tbml":
		baseScore *= 1.4
	case "sanctions":
		baseScore *= 1.6
	case "fraud":
		baseScore *= 1.1
	}

	return baseScore
}

// calculateFinalRiskScore calculates the final risk score
func (rks *RiskKeywordsService) calculateFinalRiskScore(totalScore float64, matchCount int, contentLength int) float64 {
	if matchCount == 0 {
		return 0.0
	}

	// Normalize by content length (longer content = more context)
	lengthFactor := 1.0
	if contentLength > 1000 {
		lengthFactor = 1.2
	} else if contentLength > 500 {
		lengthFactor = 1.1
	} else if contentLength < 100 {
		lengthFactor = 0.8
	}

	// Apply diminishing returns for multiple matches
	diminishingFactor := 1.0
	if matchCount > 5 {
		diminishingFactor = 0.8
	} else if matchCount > 10 {
		diminishingFactor = 0.6
	}

	finalScore := totalScore * lengthFactor * diminishingFactor

	// Cap at 1.0
	if finalScore > 1.0 {
		finalScore = 1.0
	}

	return finalScore
}

// determineRiskLevel determines the risk level based on the risk score
func (rks *RiskKeywordsService) determineRiskLevel(riskScore float64) string {
	if riskScore >= 0.8 {
		return "critical"
	} else if riskScore >= 0.6 {
		return "high"
	} else if riskScore >= 0.4 {
		return "medium"
	} else if riskScore >= 0.2 {
		return "low"
	} else {
		return "minimal"
	}
}

// calculateConfidence calculates the confidence in the risk assessment
func (rks *RiskKeywordsService) calculateConfidence(contentLength, matchCount, keywordCount int) float64 {
	confidence := 0.5 // Base confidence

	// Increase confidence with more content
	if contentLength > 500 {
		confidence += 0.2
	} else if contentLength > 200 {
		confidence += 0.1
	}

	// Increase confidence with more matches
	if matchCount > 3 {
		confidence += 0.2
	} else if matchCount > 1 {
		confidence += 0.1
	}

	// Increase confidence with diverse keyword matches
	if keywordCount > 5 {
		confidence += 0.1
	}

	// Cap at 1.0
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// matchesPattern performs simple pattern matching (simplified regex)
func (rks *RiskKeywordsService) matchesPattern(content, pattern string) bool {
	// Convert pattern to simple wildcard matching
	// Replace * with any characters and ? with single character
	pattern = strings.ToLower(pattern)
	content = strings.ToLower(content)

	// Simple wildcard implementation
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.Contains(content, parts[0]) && strings.Contains(content, parts[1])
		}
	}

	// Direct match
	return strings.Contains(content, pattern)
}

// GetRiskKeywordsByCategory retrieves risk keywords by category
func (rks *RiskKeywordsService) GetRiskKeywordsByCategory(ctx context.Context, category string) ([]*RiskKeyword, error) {
	query := `
		SELECT id, keyword, risk_category, risk_severity, description,
		       mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
		       detection_patterns, synonyms, is_active, created_at, updated_at
		FROM risk_keywords 
		WHERE risk_category = $1 AND is_active = true
		ORDER BY risk_severity DESC, keyword
	`

	rows, err := rks.db.QueryContext(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk keywords by category: %w", err)
	}
	defer rows.Close()

	var keywords []*RiskKeyword
	for rows.Next() {
		var keyword RiskKeyword
		var mccCodes, naicsCodes, sicCodes, cardRestrictions, patterns, synonyms pq.StringArray

		err := rows.Scan(
			&keyword.ID,
			&keyword.Keyword,
			&keyword.RiskCategory,
			&keyword.RiskSeverity,
			&keyword.Description,
			&mccCodes,
			&naicsCodes,
			&sicCodes,
			&cardRestrictions,
			&patterns,
			&synonyms,
			&keyword.IsActive,
			&keyword.CreatedAt,
			&keyword.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk keyword: %w", err)
		}

		keyword.MCCCodes = []string(mccCodes)
		keyword.NAICSCodes = []string(naicsCodes)
		keyword.SICCodes = []string(sicCodes)
		keyword.CardBrandRestrictions = []string(cardRestrictions)
		keyword.DetectionPatterns = []string(patterns)
		keyword.Synonyms = []string(synonyms)

		keywords = append(keywords, &keyword)
	}

	return keywords, nil
}

// GetRiskKeywordsBySeverity retrieves risk keywords by severity level
func (rks *RiskKeywordsService) GetRiskKeywordsBySeverity(ctx context.Context, severity string) ([]*RiskKeyword, error) {
	query := `
		SELECT id, keyword, risk_category, risk_severity, description,
		       mcc_codes, naics_codes, sic_codes, card_brand_restrictions,
		       detection_patterns, synonyms, is_active, created_at, updated_at
		FROM risk_keywords 
		WHERE risk_severity = $1 AND is_active = true
		ORDER BY risk_category, keyword
	`

	rows, err := rks.db.QueryContext(ctx, query, severity)
	if err != nil {
		return nil, fmt.Errorf("failed to query risk keywords by severity: %w", err)
	}
	defer rows.Close()

	var keywords []*RiskKeyword
	for rows.Next() {
		var keyword RiskKeyword
		var mccCodes, naicsCodes, sicCodes, cardRestrictions, patterns, synonyms pq.StringArray

		err := rows.Scan(
			&keyword.ID,
			&keyword.Keyword,
			&keyword.RiskCategory,
			&keyword.RiskSeverity,
			&keyword.Description,
			&mccCodes,
			&naicsCodes,
			&sicCodes,
			&cardRestrictions,
			&patterns,
			&synonyms,
			&keyword.IsActive,
			&keyword.CreatedAt,
			&keyword.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan risk keyword: %w", err)
		}

		keyword.MCCCodes = []string(mccCodes)
		keyword.NAICSCodes = []string(naicsCodes)
		keyword.SICCodes = []string(sicCodes)
		keyword.CardBrandRestrictions = []string(cardRestrictions)
		keyword.DetectionPatterns = []string(patterns)
		keyword.Synonyms = []string(synonyms)

		keywords = append(keywords, &keyword)
	}

	return keywords, nil
}

// GetRiskStatistics returns statistics about risk keywords
func (rks *RiskKeywordsService) GetRiskStatistics(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	var totalCount int
	err := rks.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM risk_keywords WHERE is_active = true").Scan(&totalCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats["total_keywords"] = totalCount

	// Count by category
	categoryQuery := `
		SELECT risk_category, COUNT(*) 
		FROM risk_keywords 
		WHERE is_active = true 
		GROUP BY risk_category
	`
	rows, err := rks.db.QueryContext(ctx, categoryQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query category stats: %w", err)
	}
	defer rows.Close()

	categoryStats := make(map[string]int)
	for rows.Next() {
		var category string
		var count int
		if err := rows.Scan(&category, &count); err != nil {
			return nil, fmt.Errorf("failed to scan category stats: %w", err)
		}
		categoryStats[category] = count
	}
	stats["by_category"] = categoryStats

	// Count by severity
	severityQuery := `
		SELECT risk_severity, COUNT(*) 
		FROM risk_keywords 
		WHERE is_active = true 
		GROUP BY risk_severity
	`
	rows, err = rks.db.QueryContext(ctx, severityQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to query severity stats: %w", err)
	}
	defer rows.Close()

	severityStats := make(map[string]int)
	for rows.Next() {
		var severity string
		var count int
		if err := rows.Scan(&severity, &count); err != nil {
			return nil, fmt.Errorf("failed to scan severity stats: %w", err)
		}
		severityStats[severity] = count
	}
	stats["by_severity"] = severityStats

	return stats, nil
}
