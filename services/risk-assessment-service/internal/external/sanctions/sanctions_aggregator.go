package sanctions

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/external/ofac"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// SanctionsAggregator provides unified sanctions screening across all lists
type SanctionsAggregator struct {
	logger     *zap.Logger
	config     *SanctionsAggregatorConfig
	ofacClient *ofac.OFACMock
	unClient   *UNSanctionsClient
	euClient   *EUSanctionsClient
}

// SanctionsAggregatorConfig holds configuration for the sanctions aggregator
type SanctionsAggregatorConfig struct {
	EnableOFAC bool `json:"enable_ofac"`
	EnableUN   bool `json:"enable_un"`
	EnableEU   bool `json:"enable_eu"`
	// Fuzzy matching configuration
	FuzzyMatchThreshold float64 `json:"fuzzy_match_threshold"` // 0.0-1.0
	PhoneticMatch       bool    `json:"phonetic_match"`
	// Risk scoring configuration
	HighRiskThreshold   float64 `json:"high_risk_threshold"`   // 0.0-1.0
	MediumRiskThreshold float64 `json:"medium_risk_threshold"` // 0.0-1.0
}

// UnifiedSanctionsMatch represents a unified sanctions match across all lists
type UnifiedSanctionsMatch struct {
	MatchID          string     `json:"match_id"`
	EntityID         string     `json:"entity_id"`
	EntityName       string     `json:"entity_name"`
	EntityType       string     `json:"entity_type"`
	Country          string     `json:"country"`
	MatchType        string     `json:"match_type"` // "exact", "partial", "fuzzy", "phonetic"
	MatchScore       float64    `json:"match_score"`
	Confidence       float64    `json:"confidence"`
	RiskLevel        string     `json:"risk_level"`
	Category         string     `json:"category"`
	Subcategory      string     `json:"subcategory"`
	Source           string     `json:"source"`
	SanctionsProgram string     `json:"sanctions_program"`
	ProgramList      string     `json:"program_list"`
	MatchDetails     string     `json:"match_details"`
	AdditionalInfo   string     `json:"additional_info"`
	LastUpdated      time.Time  `json:"last_updated"`
	IsActive         bool       `json:"is_active"`
	ExpiryDate       *time.Time `json:"expiry_date,omitempty"`
	// Additional fields from different sources
	DateOfBirth    *time.Time `json:"date_of_birth,omitempty"`
	PlaceOfBirth   string     `json:"place_of_birth,omitempty"`
	Nationality    string     `json:"nationality,omitempty"`
	PassportNumber string     `json:"passport_number,omitempty"`
	Address        string     `json:"address,omitempty"`
	Title          string     `json:"title,omitempty"`
	Remarks        string     `json:"remarks,omitempty"`
	EUMemberState  string     `json:"eu_member_state,omitempty"`
}

// UnifiedSanctionsResult represents the unified result from all sanctions lists
type UnifiedSanctionsResult struct {
	RequestID         string                  `json:"request_id"`
	QueryName         string                  `json:"query_name"`
	QueryCountry      string                  `json:"query_country"`
	TotalMatches      int                     `json:"total_matches"`
	HighRiskMatches   int                     `json:"high_risk_matches"`
	MediumRiskMatches int                     `json:"medium_risk_matches"`
	LowRiskMatches    int                     `json:"low_risk_matches"`
	Matches           []UnifiedSanctionsMatch `json:"matches"`
	OverallRiskScore  float64                 `json:"overall_risk_score"`
	OverallRiskLevel  string                  `json:"overall_risk_level"`
	ScreeningTime     time.Duration           `json:"screening_time"`
	LastChecked       time.Time               `json:"last_checked"`
	DataQuality       string                  `json:"data_quality"`
	Sources           []string                `json:"sources"`
	// Source-specific results
	OFACResults *ofac.SanctionsSearchResult `json:"ofac_results,omitempty"`
	UNResults   *UNSanctionsSearchResult    `json:"un_results,omitempty"`
	EUResults   *EUSanctionsSearchResult    `json:"eu_results,omitempty"`
}

// NewSanctionsAggregator creates a new sanctions aggregator
func NewSanctionsAggregator(
	config *SanctionsAggregatorConfig,
	ofacClient *ofac.OFACMock,
	unClient *UNSanctionsClient,
	euClient *EUSanctionsClient,
	logger *zap.Logger,
) *SanctionsAggregator {
	return &SanctionsAggregator{
		logger:     logger,
		config:     config,
		ofacClient: ofacClient,
		unClient:   unClient,
		euClient:   euClient,
	}
}

// ScreenEntity performs comprehensive sanctions screening across all lists
func (sa *SanctionsAggregator) ScreenEntity(ctx context.Context, entityName, country string) (*UnifiedSanctionsResult, error) {
	startTime := time.Now()
	sa.logger.Info("Performing unified sanctions screening",
		zap.String("entity_name", entityName),
		zap.String("country", country))

	// Perform screening across all enabled sources in parallel
	type screeningResult struct {
		ofacResult *ofac.SanctionsSearchResult
		unResult   *UNSanctionsSearchResult
		euResult   *EUSanctionsSearchResult
		err        error
	}

	results := make(chan screeningResult, 1)

	go func() {
		var result screeningResult

		// OFAC screening
		if sa.config.EnableOFAC && sa.ofacClient != nil {
			ofacResult, err := sa.ofacClient.SearchSanctions(ctx, entityName, country)
			result.ofacResult = ofacResult
			if err != nil {
				sa.logger.Warn("OFAC screening failed", zap.Error(err))
			}
		}

		// UN sanctions screening
		if sa.config.EnableUN && sa.unClient != nil {
			unResult, err := sa.unClient.SearchSanctions(ctx, entityName, country)
			result.unResult = unResult
			if err != nil {
				sa.logger.Warn("UN sanctions screening failed", zap.Error(err))
			}
		}

		// EU sanctions screening
		if sa.config.EnableEU && sa.euClient != nil {
			euResult, err := sa.euClient.SearchSanctions(ctx, entityName, country)
			result.euResult = euResult
			if err != nil {
				sa.logger.Warn("EU sanctions screening failed", zap.Error(err))
			}
		}

		results <- result
	}()

	// Wait for results
	result := <-results
	if result.err != nil {
		return nil, result.err
	}

	// Aggregate and deduplicate matches
	unifiedMatches := sa.aggregateMatches(entityName, country, result.ofacResult, result.unResult, result.euResult)

	// Calculate risk metrics
	highRiskCount := 0
	mediumRiskCount := 0
	lowRiskCount := 0
	totalRiskScore := 0.0

	for _, match := range unifiedMatches {
		switch match.RiskLevel {
		case "high", "critical":
			highRiskCount++
			totalRiskScore += match.MatchScore * 0.9
		case "medium":
			mediumRiskCount++
			totalRiskScore += match.MatchScore * 0.6
		case "low":
			lowRiskCount++
			totalRiskScore += match.MatchScore * 0.3
		}
	}

	overallRiskScore := 0.0
	overallRiskLevel := "low"
	if len(unifiedMatches) > 0 {
		overallRiskScore = totalRiskScore / float64(len(unifiedMatches))
		if overallRiskScore >= sa.config.HighRiskThreshold {
			overallRiskLevel = "critical"
		} else if overallRiskScore >= sa.config.MediumRiskThreshold {
			overallRiskLevel = "high"
		} else if overallRiskScore >= 0.3 {
			overallRiskLevel = "medium"
		}
	}

	// Collect sources
	var sources []string
	if result.ofacResult != nil {
		sources = append(sources, "OFAC")
	}
	if result.unResult != nil {
		sources = append(sources, "UN Security Council")
	}
	if result.euResult != nil {
		sources = append(sources, "European Union")
	}

	unifiedResult := &UnifiedSanctionsResult{
		RequestID:         sa.generateRequestID(),
		QueryName:         entityName,
		QueryCountry:      country,
		TotalMatches:      len(unifiedMatches),
		HighRiskMatches:   highRiskCount,
		MediumRiskMatches: mediumRiskCount,
		LowRiskMatches:    lowRiskCount,
		Matches:           unifiedMatches,
		OverallRiskScore:  overallRiskScore,
		OverallRiskLevel:  overallRiskLevel,
		ScreeningTime:     time.Since(startTime),
		LastChecked:       time.Now(),
		DataQuality:       sa.assessDataQuality(result.ofacResult, result.unResult, result.euResult),
		Sources:           sources,
		OFACResults:       result.ofacResult,
		UNResults:         result.unResult,
		EUResults:         result.euResult,
	}

	sa.logger.Info("Unified sanctions screening completed",
		zap.String("entity_name", entityName),
		zap.Int("total_matches", unifiedResult.TotalMatches),
		zap.String("overall_risk_level", unifiedResult.OverallRiskLevel),
		zap.Duration("screening_time", unifiedResult.ScreeningTime))

	return unifiedResult, nil
}

// ScreenBatch performs batch sanctions screening
func (sa *SanctionsAggregator) ScreenBatch(ctx context.Context, entities []string, country string) ([]*UnifiedSanctionsResult, error) {
	sa.logger.Info("Performing batch unified sanctions screening",
		zap.Int("entity_count", len(entities)),
		zap.String("country", country))

	var results []*UnifiedSanctionsResult
	for _, entity := range entities {
		result, err := sa.ScreenEntity(ctx, entity, country)
		if err != nil {
			sa.logger.Warn("Failed to screen entity in batch",
				zap.String("entity", entity),
				zap.Error(err))
			continue
		}
		results = append(results, result)
	}

	sa.logger.Info("Batch unified sanctions screening completed",
		zap.Int("processed_entities", len(results)),
		zap.Int("total_entities", len(entities)))

	return results, nil
}

// IsHealthy checks if all sanctions services are healthy
func (sa *SanctionsAggregator) IsHealthy(ctx context.Context) error {
	var errors []error

	if sa.config.EnableOFAC && sa.ofacClient != nil {
		if err := sa.ofacClient.IsHealthy(ctx); err != nil {
			errors = append(errors, fmt.Errorf("OFAC: %w", err))
		}
	}

	if sa.config.EnableUN && sa.unClient != nil {
		if err := sa.unClient.IsHealthy(ctx); err != nil {
			errors = append(errors, fmt.Errorf("UN: %w", err))
		}
	}

	if sa.config.EnableEU && sa.euClient != nil {
		if err := sa.euClient.IsHealthy(ctx); err != nil {
			errors = append(errors, fmt.Errorf("EU: %w", err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("sanctions services unhealthy: %v", errors)
	}

	return nil
}

// GenerateRiskFactors generates risk factors from unified sanctions data
func (sa *SanctionsAggregator) GenerateRiskFactors(result *UnifiedSanctionsResult) []models.RiskFactor {
	var riskFactors []models.RiskFactor
	now := time.Now()

	// Overall sanctions risk factor
	riskFactors = append(riskFactors, models.RiskFactor{
		Category:    models.RiskCategoryCompliance,
		Subcategory: "sanctions",
		Name:        "unified_sanctions_risk",
		Score:       result.OverallRiskScore,
		Weight:      0.5,
		Description: "Overall risk score from unified sanctions screening across all lists",
		Source:      "sanctions_aggregator",
		Confidence:  0.95,
		Impact:      "Sanctions violations can result in severe penalties and business restrictions",
		Mitigation:  "Immediate compliance review and potential business relationship termination",
		LastUpdated: &now,
	})

	// Source-specific risk factors
	if result.OFACResults != nil && result.OFACResults.TotalMatches > 0 {
		ofacRisk := 0.95 // Critical risk for OFAC matches
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryCompliance,
			Subcategory: "ofac_sanctions",
			Name:        "ofac_sanctions_risk",
			Score:       ofacRisk,
			Weight:      0.4,
			Description: "Risk associated with OFAC sanctions list matches",
			Source:      "ofac",
			Confidence:  0.98,
			Impact:      "OFAC sanctions violations can result in severe US penalties",
			Mitigation:  "Immediate compliance review and potential business relationship termination",
			LastUpdated: &now,
		})
	}

	if result.UNResults != nil && result.UNResults.TotalMatches > 0 {
		unRisk := 0.95 // Critical risk for UN sanctions matches
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryCompliance,
			Subcategory: "un_sanctions",
			Name:        "un_sanctions_risk",
			Score:       unRisk,
			Weight:      0.4,
			Description: "Risk associated with UN Security Council sanctions list matches",
			Source:      "un_sanctions",
			Confidence:  0.98,
			Impact:      "UN sanctions violations can result in severe international penalties",
			Mitigation:  "Immediate compliance review and potential business relationship termination",
			LastUpdated: &now,
		})
	}

	if result.EUResults != nil && result.EUResults.TotalMatches > 0 {
		euRisk := 0.9 // High risk for EU sanctions matches
		riskFactors = append(riskFactors, models.RiskFactor{
			Category:    models.RiskCategoryCompliance,
			Subcategory: "eu_sanctions",
			Name:        "eu_sanctions_risk",
			Score:       euRisk,
			Weight:      0.4,
			Description: "Risk associated with EU sanctions list matches",
			Source:      "eu_sanctions",
			Confidence:  0.95,
			Impact:      "EU sanctions violations can result in severe penalties and business restrictions",
			Mitigation:  "Immediate compliance review and potential business relationship termination",
			LastUpdated: &now,
		})
	}

	return riskFactors
}

// Helper methods

func (sa *SanctionsAggregator) aggregateMatches(
	entityName, country string,
	ofacResult *ofac.SanctionsSearchResult,
	unResult *UNSanctionsSearchResult,
	euResult *EUSanctionsSearchResult,
) []UnifiedSanctionsMatch {
	var allMatches []UnifiedSanctionsMatch

	// Convert OFAC matches
	if ofacResult != nil {
		for _, match := range ofacResult.Matches {
			unifiedMatch := UnifiedSanctionsMatch{
				MatchID:          fmt.Sprintf("UNIFIED_OFAC_%s", match.EntityID),
				EntityID:         match.EntityID,
				EntityName:       match.EntityName,
				EntityType:       match.EntityType,
				Country:          match.Country,
				MatchType:        "exact", // OFAC matches are typically exact
				MatchScore:       match.MatchScore,
				Confidence:       0.95, // High confidence for OFAC matches
				RiskLevel:        match.RiskLevel,
				Category:         "sanctions",
				Subcategory:      match.SanctionsProgram,
				Source:           "OFAC",
				SanctionsProgram: match.SanctionsProgram,
				ProgramList:      match.ProgramList,
				MatchDetails:     fmt.Sprintf("Entity matches OFAC sanctions list: %s", match.ProgramList),
				AdditionalInfo:   "Immediate compliance review required",
				LastUpdated:      match.LastUpdated,
				IsActive:         true,
				Title:            match.Title,
				Remarks:          match.Remarks,
			}
			allMatches = append(allMatches, unifiedMatch)
		}
	}

	// Convert UN sanctions matches
	if unResult != nil {
		for _, match := range unResult.Matches {
			unifiedMatch := UnifiedSanctionsMatch{
				MatchID:          fmt.Sprintf("UNIFIED_UN_%s", match.EntityID),
				EntityID:         match.EntityID,
				EntityName:       match.EntityName,
				EntityType:       match.EntityType,
				Country:          match.Country,
				MatchType:        "exact", // UN matches are typically exact
				MatchScore:       match.MatchScore,
				Confidence:       0.95,
				RiskLevel:        match.RiskLevel,
				Category:         "sanctions",
				Subcategory:      match.SanctionsProgram,
				Source:           "UN Security Council",
				SanctionsProgram: match.SanctionsProgram,
				ProgramList:      match.ProgramList,
				MatchDetails:     fmt.Sprintf("Entity matches UN sanctions list: %s", match.ProgramList),
				AdditionalInfo:   "Immediate compliance review required",
				LastUpdated:      match.LastUpdated,
				IsActive:         true,
				DateOfBirth:      match.DateOfBirth,
				PlaceOfBirth:     match.PlaceOfBirth,
				Nationality:      match.Nationality,
				PassportNumber:   match.PassportNumber,
				Address:          match.Address,
				Title:            match.Title,
				Remarks:          match.Remarks,
			}
			allMatches = append(allMatches, unifiedMatch)
		}
	}

	// Convert EU sanctions matches
	if euResult != nil {
		for _, match := range euResult.Matches {
			unifiedMatch := UnifiedSanctionsMatch{
				MatchID:          fmt.Sprintf("UNIFIED_EU_%s", match.EntityID),
				EntityID:         match.EntityID,
				EntityName:       match.EntityName,
				EntityType:       match.EntityType,
				Country:          match.Country,
				MatchType:        "exact", // EU matches are typically exact
				MatchScore:       match.MatchScore,
				Confidence:       0.95,
				RiskLevel:        match.RiskLevel,
				Category:         "sanctions",
				Subcategory:      match.SanctionsProgram,
				Source:           "European Union",
				SanctionsProgram: match.SanctionsProgram,
				ProgramList:      match.ProgramList,
				MatchDetails:     fmt.Sprintf("Entity matches EU sanctions list: %s", match.ProgramList),
				AdditionalInfo:   "Immediate compliance review required",
				LastUpdated:      match.LastUpdated,
				IsActive:         true,
				DateOfBirth:      match.DateOfBirth,
				PlaceOfBirth:     match.PlaceOfBirth,
				Nationality:      match.Nationality,
				PassportNumber:   match.PassportNumber,
				Address:          match.Address,
				Title:            match.Title,
				Remarks:          match.Remarks,
				EUMemberState:    match.EUMemberState,
			}
			allMatches = append(allMatches, unifiedMatch)
		}
	}

	// Deduplicate matches using fuzzy matching
	deduplicatedMatches := sa.deduplicateMatches(allMatches)

	// Sort by risk level and match score
	sort.Slice(deduplicatedMatches, func(i, j int) bool {
		// First sort by risk level
		riskOrder := map[string]int{"critical": 4, "high": 3, "medium": 2, "low": 1}
		if riskOrder[deduplicatedMatches[i].RiskLevel] != riskOrder[deduplicatedMatches[j].RiskLevel] {
			return riskOrder[deduplicatedMatches[i].RiskLevel] > riskOrder[deduplicatedMatches[j].RiskLevel]
		}
		// Then by match score
		return deduplicatedMatches[i].MatchScore > deduplicatedMatches[j].MatchScore
	})

	return deduplicatedMatches
}

func (sa *SanctionsAggregator) deduplicateMatches(matches []UnifiedSanctionsMatch) []UnifiedSanctionsMatch {
	if len(matches) <= 1 {
		return matches
	}

	var deduplicated []UnifiedSanctionsMatch
	seen := make(map[string]bool)

	for _, match := range matches {
		// Create a key for deduplication based on entity name and source
		key := fmt.Sprintf("%s_%s", strings.ToLower(match.EntityName), match.Source)

		if !seen[key] {
			// Check for fuzzy matches with existing entries
			isDuplicate := false
			for _, existing := range deduplicated {
				if sa.isFuzzyMatch(match.EntityName, existing.EntityName) && match.Source == existing.Source {
					// Keep the match with higher score
					if match.MatchScore > existing.MatchScore {
						// Replace existing match
						for i, existingMatch := range deduplicated {
							if existingMatch.MatchID == existing.MatchID {
								deduplicated[i] = match
								break
							}
						}
					}
					isDuplicate = true
					break
				}
			}

			if !isDuplicate {
				deduplicated = append(deduplicated, match)
				seen[key] = true
			}
		}
	}

	return deduplicated
}

func (sa *SanctionsAggregator) isFuzzyMatch(name1, name2 string) bool {
	// Simple fuzzy matching based on string similarity
	// In a real implementation, you would use more sophisticated algorithms
	// like Levenshtein distance, Jaro-Winkler, or phonetic matching

	// Normalize names
	name1 = strings.ToLower(strings.TrimSpace(name1))
	name2 = strings.ToLower(strings.TrimSpace(name2))

	// Exact match
	if name1 == name2 {
		return true
	}

	// Check if one name contains the other
	if strings.Contains(name1, name2) || strings.Contains(name2, name1) {
		return true
	}

	// Simple similarity check (in real implementation, use proper string similarity)
	// This is a simplified version - in production, use a proper string similarity library
	similarity := sa.calculateStringSimilarity(name1, name2)
	return similarity >= sa.config.FuzzyMatchThreshold
}

func (sa *SanctionsAggregator) calculateStringSimilarity(s1, s2 string) float64 {
	// Simplified string similarity calculation
	// In production, use a proper string similarity library like go-fuzzywuzzy

	if len(s1) == 0 && len(s2) == 0 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Simple character-based similarity
	commonChars := 0
	minLen := int(math.Min(float64(len(s1)), float64(len(s2))))

	for i := 0; i < minLen; i++ {
		if s1[i] == s2[i] {
			commonChars++
		}
	}

	return float64(commonChars) / float64(int(math.Max(float64(len(s1)), float64(len(s2)))))
}

func (sa *SanctionsAggregator) generateRequestID() string {
	return fmt.Sprintf("UNIFIED_%d", time.Now().UnixNano())
}

func (sa *SanctionsAggregator) assessDataQuality(
	ofacResult *ofac.SanctionsSearchResult,
	unResult *UNSanctionsSearchResult,
	euResult *EUSanctionsSearchResult,
) string {
	// Assess data quality based on available sources and their quality
	var qualities []string

	if ofacResult != nil {
		qualities = append(qualities, ofacResult.DataQuality)
	}
	if unResult != nil {
		qualities = append(qualities, unResult.DataQuality)
	}
	if euResult != nil {
		qualities = append(qualities, euResult.DataQuality)
	}

	if len(qualities) == 0 {
		return "unknown"
	}

	// Return the best quality available
	qualityOrder := map[string]int{"excellent": 3, "good": 2, "average": 1}
	bestQuality := "average"
	bestScore := 0

	for _, quality := range qualities {
		if score, exists := qualityOrder[quality]; exists && score > bestScore {
			bestScore = score
			bestQuality = quality
		}
	}

	return bestQuality
}
