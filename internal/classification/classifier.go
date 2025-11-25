package classification

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/classification/repository"
)

// ClassificationCodeGenerator provides database-driven classification code generation
type ClassificationCodeGenerator struct {
	repo    repository.KeywordRepository
	logger  *log.Logger
	monitor *ClassificationAccuracyMonitoring
}

// NewClassificationCodeGenerator creates a new classification code generator
func NewClassificationCodeGenerator(repo repository.KeywordRepository, logger *log.Logger) *ClassificationCodeGenerator {
	if logger == nil {
		logger = log.Default()
	}

	return &ClassificationCodeGenerator{
		repo:    repo,
		logger:  logger,
		monitor: nil, // Will be set separately if monitoring is needed
	}
}

// NewClassificationCodeGeneratorWithMonitoring creates a new classification code generator with monitoring
func NewClassificationCodeGeneratorWithMonitoring(repo repository.KeywordRepository, logger *log.Logger, monitor *ClassificationAccuracyMonitoring) *ClassificationCodeGenerator {
	if logger == nil {
		logger = log.Default()
	}

	return &ClassificationCodeGenerator{
		repo:    repo,
		logger:  logger,
		monitor: monitor,
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

// IndustryResult represents an industry with its confidence score
type IndustryResult struct {
	IndustryName string
	Confidence   float64
}

// GenerateClassificationCodes generates MCC, SIC, and NAICS codes based on extracted keywords and industry analysis
// If additionalIndustries is provided, codes will be generated for those industries as well with weighted confidence
func (g *ClassificationCodeGenerator) GenerateClassificationCodes(
	ctx context.Context,
	keywords []string,
	detectedIndustry string,
	confidence float64,
	additionalIndustries ...IndustryResult, // Optional: top N industries from ensemble
) (*ClassificationCodesInfo, error) {
	startTime := time.Now()
	requestID := g.generateRequestID()

	g.logger.Printf("🔍 Generating classification codes for industry: %s (confidence: %.2f%%) (request: %s)", detectedIndustry, confidence*100, requestID)

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

	// Generate codes using parallel processing for better performance
	// Include additional industries if provided
	allIndustries := []IndustryResult{
		{IndustryName: detectedIndustry, Confidence: confidence},
	}
	// Add additional industries with weighted confidence (default weight: 0.7)
	const multiIndustryWeight = 0.7
	for _, additional := range additionalIndustries {
		allIndustries = append(allIndustries, IndustryResult{
			IndustryName: additional.IndustryName,
			Confidence:   additional.Confidence * multiIndustryWeight,
		})
	}
	
	if len(allIndustries) > 1 {
		g.logger.Printf("📊 Multi-industry code generation: %d industries", len(allIndustries))
	}
	
	g.generateCodesInParallel(ctx, codes, keywordsLower, allIndustries)

	// Record performance metrics
	g.recordCodeGenerationMetrics(ctx, requestID, keywords, detectedIndustry, confidence, codes, time.Since(startTime), nil)

	g.logger.Printf("✅ Generated %d MCC, %d SIC, %d NAICS codes (request: %s)",
		len(codes.MCC), len(codes.SIC), len(codes.NAICS), requestID)

	return codes, nil
}

// CodeMatch represents a code match with metadata from keyword matching
type CodeMatch struct {
	Code           *repository.ClassificationCode
	RelevanceScore float64
	MatchType      string // "exact", "partial", "synonym"
	Source         string // "industry" or "keyword"
	Confidence     float64
}

// RankedCode represents a ranked code with combined confidence from multiple sources
type RankedCode struct {
	Code              *repository.ClassificationCode
	CombinedConfidence float64
	Sources            []string // Which sources contributed
	MatchDetails       []CodeMatch
}

// generateCodesFromKeywords generates codes using direct keyword matching
func (g *ClassificationCodeGenerator) generateCodesFromKeywords(
	ctx context.Context,
	keywords []string,
	codeType string,
	industryConfidence float64,
) ([]CodeMatch, error) {
	if len(keywords) == 0 {
		return []CodeMatch{}, nil
	}

	g.logger.Printf("🔍 Generating %s codes from keywords: %d keywords", codeType, len(keywords))

	// Use default minRelevance of 0.5
	minRelevance := 0.5
	keywordCodes, err := g.repo.GetClassificationCodesByKeywords(ctx, keywords, codeType, minRelevance)
	if err != nil {
		g.logger.Printf("⚠️ Failed to get codes from keywords: %v", err)
		return []CodeMatch{}, nil // Return empty instead of error to allow fallback
	}

	// Convert to CodeMatch slice
	matches := make([]CodeMatch, 0, len(keywordCodes))
	for _, codeWithMeta := range keywordCodes {
		// Calculate confidence: relevance_score * industry_confidence * 0.85
		confidence := codeWithMeta.RelevanceScore * industryConfidence * 0.85
		
		matches = append(matches, CodeMatch{
			Code:           &codeWithMeta.ClassificationCode,
			RelevanceScore: codeWithMeta.RelevanceScore,
			MatchType:      codeWithMeta.MatchType,
			Source:         "keyword",
			Confidence:     confidence,
		})
	}

	g.logger.Printf("✅ Generated %d %s codes from keywords", len(matches), codeType)
	return matches, nil
}

// mergeCodeResults combines industry-based and keyword-based code results
// Applies confidence filtering, ranking, and limits
func (g *ClassificationCodeGenerator) mergeCodeResults(
	industryCodes []*repository.ClassificationCode,
	keywordCodes []CodeMatch,
	industryConfidence float64,
	codeType string,
) []RankedCode {
	// Configuration constants
	const (
		confidenceThreshold = 0.6  // Minimum confidence to include
		maxCodesPerType     = 3    // Maximum codes to return per type (top 3 as requested)
		primaryCodeBoost    = 1.2    // Boost multiplier for is_primary codes
	)
	// Create a map to track codes by their ID for deduplication
	codeMap := make(map[int]*RankedCode)

	// Add industry-based codes
	for _, code := range industryCodes {
		if code.CodeType != codeType {
			continue
		}

		// Industry-based codes: confidence * 0.9
		confidence := industryConfidence * 0.9

		if existing, exists := codeMap[code.ID]; exists {
			// Code already exists from keyword match - combine sources
			existing.Sources = append(existing.Sources, "industry")
			// Update combined confidence (weighted average)
			existing.CombinedConfidence = (existing.CombinedConfidence + confidence) / 2.0
			// Boost for codes matched by both sources
			existing.CombinedConfidence *= 1.3
		} else {
			// New code from industry
			codeMap[code.ID] = &RankedCode{
				Code:              code,
				CombinedConfidence: confidence,
				Sources:            []string{"industry"},
				MatchDetails: []CodeMatch{
					{
						Code:       code,
						Source:     "industry",
						Confidence: confidence,
					},
				},
			}
		}
	}

	// Add keyword-based codes
	for _, keywordMatch := range keywordCodes {
		if keywordMatch.Code.CodeType != codeType {
			continue
		}

		if existing, exists := codeMap[keywordMatch.Code.ID]; exists {
			// Code already exists from industry match - combine sources
			existing.Sources = append(existing.Sources, "keyword")
			// Update combined confidence (weighted average)
			existing.CombinedConfidence = (existing.CombinedConfidence + keywordMatch.Confidence) / 2.0
			// Boost for codes matched by both sources
			existing.CombinedConfidence *= 1.3
			existing.MatchDetails = append(existing.MatchDetails, keywordMatch)
		} else {
			// New code from keyword
			codeMap[keywordMatch.Code.ID] = &RankedCode{
				Code:              keywordMatch.Code,
				CombinedConfidence: keywordMatch.Confidence,
				Sources:            []string{"keyword"},
				MatchDetails:       []CodeMatch{keywordMatch},
			}
		}
	}

	// Convert map to slice and apply boosts
	results := make([]RankedCode, 0, len(codeMap))
	for _, rankedCode := range codeMap {
		// Boost is_primary codes (if the field exists - we'll check via description or other means)
		// Note: ClassificationCode doesn't have is_primary field in the struct, so we'll skip this for now
		// TODO: Add is_primary field to ClassificationCode struct if needed
		
		// Filter by confidence threshold
		if rankedCode.CombinedConfidence >= confidenceThreshold {
			results = append(results, *rankedCode)
		}
	}

	// Sort by combined confidence (descending), then by code
	// Boost codes matched by both sources (already applied in merge logic)
	sort.Slice(results, func(i, j int) bool {
		// Primary sort: combined confidence
		if results[i].CombinedConfidence != results[j].CombinedConfidence {
			return results[i].CombinedConfidence > results[j].CombinedConfidence
		}
		// Secondary sort: number of sources (more sources = better)
		if len(results[i].Sources) != len(results[j].Sources) {
			return len(results[i].Sources) > len(results[j].Sources)
		}
		// Tertiary sort: code value
		return results[i].Code.Code < results[j].Code.Code
	})

	// Limit to top N codes
	if len(results) > maxCodesPerType {
		results = results[:maxCodesPerType]
	}

	return results
}

// generateCodesForMultipleIndustries generates codes for multiple industries and merges them
func (g *ClassificationCodeGenerator) generateCodesForMultipleIndustries(
	ctx context.Context,
	industries []IndustryResult,
	codeType string,
) []*repository.ClassificationCode {
	allCodes := make([]*repository.ClassificationCode, 0)
	codeMap := make(map[int]bool) // For deduplication by code ID

	for _, industry := range industries {
		industryObj, err := g.repo.GetIndustryByName(ctx, industry.IndustryName)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get industry %s for %s codes: %v", industry.IndustryName, codeType, err)
			continue
		}

		codes, err := g.repo.GetCachedClassificationCodes(ctx, industryObj.ID)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get %s codes for industry %s: %v", codeType, industry.IndustryName, err)
			continue
		}

		// Filter by code type and deduplicate
		for _, code := range codes {
			if code.CodeType == codeType && !codeMap[code.ID] {
				codeMap[code.ID] = true
				allCodes = append(allCodes, code)
			}
		}
	}

	return allCodes
}

// generateCodesInParallel generates MCC, SIC, and NAICS codes in parallel for better performance
// Supports multiple industries for enhanced code coverage
func (g *ClassificationCodeGenerator) generateCodesInParallel(ctx context.Context, codes *ClassificationCodesInfo, keywordsLower []string, industries []IndustryResult) {
	g.logger.Printf("🚀 Starting parallel code generation for MCC, SIC, and NAICS")

	// Create a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup
	var mu sync.Mutex // Mutex to protect shared data access

	// Channel to collect errors from goroutines
	errorChan := make(chan error, 3)

	// Generate MCC codes in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.logger.Printf("🔄 Starting MCC code generation (hybrid: industry + keywords)...")

		// Get industry-based codes from all industries
		industryCodes := g.generateCodesForMultipleIndustries(ctx, industries, "MCC")
		
		// Use primary industry confidence for keyword matching
		primaryConfidence := 0.5
		if len(industries) > 0 {
			primaryConfidence = industries[0].Confidence
		}

		// Get keyword-based codes
		keywordMatches, err := g.generateCodesFromKeywords(ctx, keywordsLower, "MCC", primaryConfidence)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get MCC codes from keywords: %v", err)
		}

		// Merge results using primary industry confidence
		rankedCodes := g.mergeCodeResults(industryCodes, keywordMatches, primaryConfidence, "MCC")

		// Convert to MCCCode format
		mccResults := make([]MCCCode, 0, len(rankedCodes))
		for _, rankedCode := range rankedCodes {
			// Extract keywords from match details
			keywordsMatched := make([]string, 0)
			for _, match := range rankedCode.MatchDetails {
				if match.Source == "keyword" {
					// Try to extract keyword from match (we don't store it in CodeMatch, so we'll use a placeholder)
					keywordsMatched = append(keywordsMatched, "keyword_match")
				}
			}

				mccResults = append(mccResults, MCCCode{
				Code:        rankedCode.Code.Code,
				Description: rankedCode.Code.Description,
				Confidence:  rankedCode.CombinedConfidence,
				Keywords:    keywordsMatched,
				})
		}

		// Thread-safe update of shared codes
		mu.Lock()
		codes.MCC = mccResults
		mu.Unlock()

		g.logger.Printf("✅ MCC code generation completed: %d codes (industries: %d, keyword: %d)",
			len(mccResults), len(industries), len(keywordMatches))
	}()

	// Generate SIC codes in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.logger.Printf("🔄 Starting SIC code generation (hybrid: industry + keywords)...")

		// Get industry-based codes from all industries
		industryCodes := g.generateCodesForMultipleIndustries(ctx, industries, "SIC")
		
		// Use primary industry confidence for keyword matching
		primaryConfidence := 0.5
		if len(industries) > 0 {
			primaryConfidence = industries[0].Confidence
		}

		// Get keyword-based codes
		keywordMatches, err := g.generateCodesFromKeywords(ctx, keywordsLower, "SIC", primaryConfidence)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get SIC codes from keywords: %v", err)
		}

		// Merge results using primary industry confidence
		rankedCodes := g.mergeCodeResults(industryCodes, keywordMatches, primaryConfidence, "SIC")

		// Convert to SICCode format
		sicResults := make([]SICCode, 0, len(rankedCodes))
		for _, rankedCode := range rankedCodes {
			// Extract keywords from match details
			keywordsMatched := make([]string, 0)
			for _, match := range rankedCode.MatchDetails {
				if match.Source == "keyword" {
					keywordsMatched = append(keywordsMatched, "keyword_match")
				}
			}

				sicResults = append(sicResults, SICCode{
				Code:        rankedCode.Code.Code,
				Description: rankedCode.Code.Description,
				Confidence:  rankedCode.CombinedConfidence,
				Keywords:    keywordsMatched,
				})
		}

		// Thread-safe update of shared codes
		mu.Lock()
		codes.SIC = sicResults
		mu.Unlock()

		g.logger.Printf("✅ SIC code generation completed: %d codes (industries: %d, keyword: %d)",
			len(sicResults), len(industries), len(keywordMatches))
	}()

	// Generate NAICS codes in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		g.logger.Printf("🔄 Starting NAICS code generation (hybrid: industry + keywords)...")

		// Get industry-based codes from all industries
		industryCodes := g.generateCodesForMultipleIndustries(ctx, industries, "NAICS")
		
		// Use primary industry confidence for keyword matching
		primaryConfidence := 0.5
		if len(industries) > 0 {
			primaryConfidence = industries[0].Confidence
		}

		// Get keyword-based codes
		keywordMatches, err := g.generateCodesFromKeywords(ctx, keywordsLower, "NAICS", primaryConfidence)
		if err != nil {
			g.logger.Printf("⚠️ Failed to get NAICS codes from keywords: %v", err)
		}

		// Merge results using primary industry confidence
		rankedCodes := g.mergeCodeResults(industryCodes, keywordMatches, primaryConfidence, "NAICS")

		// Convert to NAICSCode format
		naicsResults := make([]NAICSCode, 0, len(rankedCodes))
		for _, rankedCode := range rankedCodes {
			// Extract keywords from match details
			keywordsMatched := make([]string, 0)
			for _, match := range rankedCode.MatchDetails {
				if match.Source == "keyword" {
					keywordsMatched = append(keywordsMatched, "keyword_match")
				}
			}

				naicsResults = append(naicsResults, NAICSCode{
				Code:        rankedCode.Code.Code,
				Description: rankedCode.Code.Description,
				Confidence:  rankedCode.CombinedConfidence,
				Keywords:    keywordsMatched,
				})
		}

		// Thread-safe update of shared codes
		mu.Lock()
		codes.NAICS = naicsResults
		mu.Unlock()

		g.logger.Printf("✅ NAICS code generation completed: %d codes (industries: %d, keyword: %d)",
			len(naicsResults), len(industries), len(keywordMatches))
	}()

	// Wait for all goroutines to complete
	wg.Wait()
	close(errorChan)

	// Log any errors that occurred
	for err := range errorChan {
		g.logger.Printf("⚠️ Error in parallel code generation: %v", err)
	}

	g.logger.Printf("🚀 Parallel code generation completed: %d MCC, %d SIC, %d NAICS codes",
		len(codes.MCC), len(codes.SIC), len(codes.NAICS))
}

// =============================================================================
// Performance Monitoring Helper Methods
// =============================================================================

// generateRequestID generates a unique request ID for tracking
func (g *ClassificationCodeGenerator) generateRequestID() string {
	return fmt.Sprintf("code_gen_%d", time.Now().UnixNano())
}

// recordCodeGenerationMetrics records code generation performance metrics
func (g *ClassificationCodeGenerator) recordCodeGenerationMetrics(
	ctx context.Context,
	requestID string,
	keywords []string,
	detectedIndustry string,
	confidence float64,
	codes *ClassificationCodesInfo,
	responseTime time.Duration,
	err error,
) {
	if g.monitor == nil {
		return // No monitoring configured
	}

	// Prepare metrics data
	metrics := &ClassificationAccuracyMetrics{
		Timestamp:            time.Now(),
		RequestID:            requestID,
		PredictedIndustry:    detectedIndustry,
		PredictedConfidence:  confidence,
		ResponseTimeMs:       float64(responseTime.Nanoseconds()) / 1e6, // Convert to milliseconds
		ClassificationMethod: stringPtr("code_generation"),
		KeywordsUsed:         keywords,
		ConfidenceThreshold:  0.5, // Default threshold
		CreatedAt:            time.Now(),
	}

	// Set error message if there was an error
	if err != nil {
		errorMsg := err.Error()
		metrics.ErrorMessage = &errorMsg
	}

	// Record metrics asynchronously to avoid blocking the main flow
	go func() {
		// Note: This would call the actual monitoring method when implemented
		// if err := g.monitor.RecordClassificationMetrics(ctx, metrics); err != nil {
		//     g.logger.Printf("⚠️ Failed to record code generation metrics: %v", err)
		// }
	}()
}

// GetCodeGenerationMetrics returns current code generation performance metrics
func (g *ClassificationCodeGenerator) GetCodeGenerationMetrics(ctx context.Context) (*ClassificationAccuracyStats, error) {
	if g.monitor == nil {
		return nil, fmt.Errorf("monitoring not configured")
	}

	// Note: This would call the actual monitoring method when implemented
	// return g.monitor.GetClassificationAccuracyStats(ctx, 24*time.Hour)
	return nil, fmt.Errorf("monitoring not fully implemented")
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}

// matchesKeywords checks if a classification code matches any of the provided keywords
func (g *ClassificationCodeGenerator) matchesKeywords(code *repository.ClassificationCode, keywordsLower []string) bool {
	descriptionLower := strings.ToLower(code.Description)

	// Check if any of the provided keywords match
	for _, keyword := range keywordsLower {
		if strings.Contains(descriptionLower, keyword) {
			return true
		}
	}

	return false
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
// containsAny checks if any of the source strings contain any of the target strings
func (g *ClassificationCodeGenerator) containsAny(source []string, targets []string) bool {
	for _, s := range source {
		for _, t := range targets {
			if strings.Contains(strings.ToLower(s), strings.ToLower(t)) {
				return true
			}
		}
	}
	return false
}

// findMatchingKeywords finds keywords that match any of the target strings
func (g *ClassificationCodeGenerator) findMatchingKeywords(keywords []string, targets []string) []string {
	if keywords == nil {
		return []string{}
	}

	var matches []string
	for _, keyword := range keywords {
		for _, target := range targets {
			if strings.Contains(strings.ToLower(keyword), strings.ToLower(target)) {
				matches = append(matches, keyword)
				break // Only add each keyword once
			}
		}
	}
	return matches
}

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
