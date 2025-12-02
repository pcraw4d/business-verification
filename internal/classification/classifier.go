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
	repo            repository.KeywordRepository
	codeMetadataRepo *repository.CodeMetadataRepository
	logger          *log.Logger
	monitor         *ClassificationAccuracyMonitoring
}

// CodesResult represents a result from parallel code generation
type CodesResult struct {
	Type  string
	Codes []*repository.ClassificationCode
	Error error
}

// ClassificationCodesInfoParallel represents codes queried by type in parallel
type ClassificationCodesInfoParallel struct {
	MCC   []*repository.ClassificationCode
	SIC   []*repository.ClassificationCode
	NAICS []*repository.ClassificationCode
}

// NewClassificationCodeGenerator creates a new classification code generator
func NewClassificationCodeGenerator(repo repository.KeywordRepository, logger *log.Logger) *ClassificationCodeGenerator {
	if logger == nil {
		logger = log.Default()
	}

	// Initialize code metadata repository if Supabase client is available
	var codeMetadataRepo *repository.CodeMetadataRepository
	if supabaseRepo, ok := repo.(*repository.SupabaseKeywordRepository); ok {
		// Access the Supabase client from the repository
		client := supabaseRepo.GetSupabaseClient()
		if client != nil {
			codeMetadataRepo = repository.NewCodeMetadataRepository(client, logger)
			logger.Printf("✅ Code metadata repository initialized")
		} else {
			logger.Printf("⚠️ Supabase client not available, code metadata features disabled")
		}
	} else {
		logger.Printf("⚠️ Repository is not SupabaseKeywordRepository, code metadata features disabled")
	}

	return &ClassificationCodeGenerator{
		repo:            repo,
		codeMetadataRepo: codeMetadataRepo,
		logger:          logger,
		monitor:         nil, // Will be set separately if monitoring is needed
	}
}

// GenerateCodesParallel queries MCC, NAICS, and SIC codes by type in parallel
// This is a simpler parallel method for when you just need to query codes by type
// Phase 2.4: Parallel code generation optimization
func (g *ClassificationCodeGenerator) GenerateCodesParallel(
	ctx context.Context,
	industryID int,
) (*ClassificationCodesInfoParallel, error) {
	g.logger.Printf("🚀 Starting parallel code type queries for industry: %d", industryID)
	
	codesChan := make(chan CodesResult, 3)
	var wg sync.WaitGroup
	
	// Query MCC codes in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		codeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		
		mccCodes, err := g.repo.GetClassificationCodesByType(codeCtx, "MCC")
		if err != nil {
			g.logger.Printf("⚠️ Failed to get MCC codes: %v", err)
			codesChan <- CodesResult{Type: "MCC", Codes: nil, Error: err}
			return
		}
		codesChan <- CodesResult{Type: "MCC", Codes: mccCodes, Error: nil}
	}()
	
	// Query NAICS codes in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		codeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		
		naicsCodes, err := g.repo.GetClassificationCodesByType(codeCtx, "NAICS")
		if err != nil {
			g.logger.Printf("⚠️ Failed to get NAICS codes: %v", err)
			codesChan <- CodesResult{Type: "NAICS", Codes: nil, Error: err}
			return
		}
		codesChan <- CodesResult{Type: "NAICS", Codes: naicsCodes, Error: nil}
	}()
	
	// Query SIC codes in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		codeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		
		sicCodes, err := g.repo.GetClassificationCodesByType(codeCtx, "SIC")
		if err != nil {
			g.logger.Printf("⚠️ Failed to get SIC codes: %v", err)
			codesChan <- CodesResult{Type: "SIC", Codes: nil, Error: err}
			return
		}
		codesChan <- CodesResult{Type: "SIC", Codes: sicCodes, Error: nil}
	}()
	
	// Wait for all queries to complete
	wg.Wait()
	close(codesChan)
	
	// Collect results
	codesInfo := &ClassificationCodesInfoParallel{
		MCC:   []*repository.ClassificationCode{},
		SIC:   []*repository.ClassificationCode{},
		NAICS: []*repository.ClassificationCode{},
	}
	
	for result := range codesChan {
		switch result.Type {
		case "MCC":
			if result.Error == nil {
				codesInfo.MCC = result.Codes
			}
		case "NAICS":
			if result.Error == nil {
				codesInfo.NAICS = result.Codes
			}
		case "SIC":
			if result.Error == nil {
				codesInfo.SIC = result.Codes
			}
		}
	}
	
	g.logger.Printf("✅ Parallel code queries completed: %d MCC, %d NAICS, %d SIC",
		len(codesInfo.MCC), len(codesInfo.NAICS), len(codesInfo.SIC))
	
	return codesInfo, nil
}

// NewClassificationCodeGeneratorWithMonitoring creates a new classification code generator with monitoring
func NewClassificationCodeGeneratorWithMonitoring(repo repository.KeywordRepository, logger *log.Logger, monitor *ClassificationAccuracyMonitoring) *ClassificationCodeGenerator {
	if logger == nil {
		logger = log.Default()
	}

	// Initialize code metadata repository if Supabase client is available
	var codeMetadataRepo *repository.CodeMetadataRepository
	if supabaseRepo, ok := repo.(*repository.SupabaseKeywordRepository); ok {
		// Access the Supabase client from the repository
		client := supabaseRepo.GetSupabaseClient()
		if client != nil {
			codeMetadataRepo = repository.NewCodeMetadataRepository(client, logger)
			logger.Printf("✅ Code metadata repository initialized")
		} else {
			logger.Printf("⚠️ Supabase client not available, code metadata features disabled")
		}
	} else {
		logger.Printf("⚠️ Repository is not SupabaseKeywordRepository, code metadata features disabled")
	}

	return &ClassificationCodeGenerator{
		repo:            repo,
		codeMetadataRepo: codeMetadataRepo,
		logger:          logger,
		monitor:         monitor,
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
	// Enhanced metadata from code_metadata table
	CrosswalkCodes []CrosswalkCode `json:"crosswalk_codes,omitempty"` // Related codes from other systems
	IsOfficial     bool            `json:"is_official,omitempty"`      // Whether from official source
}

// SICCode represents a Standard Industrial Classification code
type SICCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
	// Enhanced metadata from code_metadata table
	CrosswalkCodes []CrosswalkCode `json:"crosswalk_codes,omitempty"` // Related codes from other systems
	IsOfficial     bool            `json:"is_official,omitempty"`     // Whether from official source
}

// NAICSCode represents a North American Industry Classification System code
type NAICSCode struct {
	Code        string   `json:"code"`
	Description string   `json:"description"`
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords_matched"`
	// Enhanced metadata from code_metadata table
	CrosswalkCodes []CrosswalkCode `json:"crosswalk_codes,omitempty"` // Related codes from other systems
	Hierarchy      *CodeHierarchy  `json:"hierarchy,omitempty"`       // Parent/child relationships (NAICS only)
	IsOfficial     bool            `json:"is_official,omitempty"`     // Whether from official source
}

// CrosswalkCode represents a related code from another classification system
type CrosswalkCode struct {
	CodeType string `json:"code_type"` // "NAICS", "SIC", or "MCC"
	Code     string `json:"code"`
	Name     string `json:"name,omitempty"`
}

// CodeHierarchy represents parent/child relationships for a code (mainly NAICS)
type CodeHierarchy struct {
	ParentCode string   `json:"parent_code,omitempty"`
	ParentType string   `json:"parent_type,omitempty"`
	ChildCodes []string `json:"child_codes,omitempty"`
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

	// PHASE 2: Crosswalk-Enhanced Code Generation
	// Validate and enhance codes using crosswalk relationships
	if g.codeMetadataRepo != nil {
		g.logger.Printf("🔗 [Phase 2] Validating and enhancing codes with crosswalks (request: %s)", requestID)
		enhancedCodes, err := g.validateAndEnhanceCodesWithCrosswalks(ctx, codes)
		if err != nil {
			g.logger.Printf("⚠️ [Phase 2] Crosswalk validation failed: %v, using original codes", err)
		} else {
			codes = enhancedCodes
			g.logger.Printf("✅ [Phase 2] Crosswalk validation completed (request: %s)", requestID)
		}
	}

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

	// Adaptive minRelevance based on industry confidence
	// When industry is "General Business" or confidence is very low, require higher relevance
	// to avoid generating generic/default codes
	minRelevance := 0.5
	if industryConfidence < 0.4 {
		// For very low-confidence industries (likely "General Business"), require higher relevance
		// This prevents generating generic codes when industry detection fails
		minRelevance = 0.5 // Keep at 0.5 to require good keyword matches
		g.logger.Printf("📊 Requiring higher minRelevance (%.2f) due to very low industry confidence (%.2f) - avoiding generic codes", minRelevance, industryConfidence)
	} else if industryConfidence < 0.5 {
		// For low-confidence industries, use moderate threshold
		minRelevance = 0.4
		g.logger.Printf("📊 Set minRelevance to %.2f due to low industry confidence (%.2f)", minRelevance, industryConfidence)
	}
	
	keywordCodes, err := g.repo.GetClassificationCodesByKeywords(ctx, keywords, codeType, minRelevance)
	if err != nil {
		g.logger.Printf("⚠️ Failed to get codes from keywords: %v", err)
		return []CodeMatch{}, nil // Return empty instead of error to allow fallback
	}

	// Convert to CodeMatch slice
	matches := make([]CodeMatch, 0, len(keywordCodes))
	for _, codeWithMeta := range keywordCodes {
		// Calculate confidence: Use relevance_score as base, adjust by industry confidence
		// For low-confidence industries, use a less aggressive multiplier to ensure codes are generated
		var confidence float64
		if industryConfidence < 0.5 {
			// For low-confidence industries, use relevance_score with minimal penalty
			// This ensures codes are generated even with lower industry confidence
			confidence = codeWithMeta.RelevanceScore * 0.7 // Less aggressive multiplier
		} else {
			// For high-confidence industries, use the original formula
			confidence = codeWithMeta.RelevanceScore * industryConfidence * 0.85
		}
		
		// Ensure minimum confidence floor to prevent filtering out valid codes
		if confidence < 0.2 && codeWithMeta.RelevanceScore >= 0.5 {
			confidence = 0.2 // Minimum floor for codes with good relevance
		}
		
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
	// Adaptive confidence threshold: Lower when industry confidence is low
	// This ensures codes are generated even with lower confidence classifications
	confidenceThreshold := 0.6
	if industryConfidence < 0.5 {
		// Lower threshold for low-confidence industries to ensure codes are generated
		confidenceThreshold = 0.3
		g.logger.Printf("📊 Lowered confidence threshold to %.2f due to low industry confidence (%.2f)", confidenceThreshold, industryConfidence)
	}
	
	// Configuration constants
	const (
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

			// Industry-based codes: Use industry confidence with adjustment
			// Skip industry-based codes when confidence is very low (likely "General Business")
			// to avoid generating generic/default codes
			var confidence float64
			if industryConfidence < 0.4 {
				// For very low-confidence industries, skip industry-based codes
				// Rely only on keyword-based matching to avoid generic codes
				continue
			} else if industryConfidence < 0.5 {
				// For low-confidence industries, use a minimum floor
				confidence = 0.3 // Minimum floor for industry-based codes
			} else {
				// For high-confidence industries, use the original formula
				confidence = industryConfidence * 0.9
			}

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
		// Skip "General Business" industry - rely only on keyword-based matching
		// This prevents generating generic/default codes when industry detection fails
		if industry.IndustryName == "General Business" {
			g.logger.Printf("⚠️ Skipping industry-based code generation for 'General Business' - using keyword-based matching only")
			continue
		}

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

		// Convert to MCCCode format with enhanced metadata from code_metadata
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

			// Enhance description with official description from code_metadata if available
			description := rankedCode.Code.Description
			var crosswalkCodes []CrosswalkCode
			isOfficial := false
			
			if g.codeMetadataRepo != nil {
				enhancedDesc := g.codeMetadataRepo.EnhanceCodeDescription(ctx, "MCC", rankedCode.Code.Code, description)
				if enhancedDesc != description {
					g.logger.Printf("📝 Enhanced MCC %s description from code_metadata", rankedCode.Code.Code)
					description = enhancedDesc
				}
				
				// Get crosswalk codes (related NAICS/SIC codes)
				crosswalks, err := g.codeMetadataRepo.GetCrosswalkCodes(ctx, "MCC", rankedCode.Code.Code)
				if err == nil && len(crosswalks) > 0 {
					crosswalkCodes = make([]CrosswalkCode, len(crosswalks))
					for i, cw := range crosswalks {
						crosswalkCodes[i] = CrosswalkCode{
							CodeType: cw.CodeType,
							Code:     cw.Code,
							Name:     cw.Name,
						}
					}
					g.logger.Printf("🔗 Found %d crosswalk codes for MCC %s", len(crosswalkCodes), rankedCode.Code.Code)
				}
				
				// Check if code is official
				metadata, _ := g.codeMetadataRepo.GetCodeMetadata(ctx, "MCC", rankedCode.Code.Code)
				if metadata != nil {
					isOfficial = metadata.IsOfficial
				}
			}

			mccResults = append(mccResults, MCCCode{
				Code:          rankedCode.Code.Code,
				Description:   description,
				Confidence:    rankedCode.CombinedConfidence,
				Keywords:      keywordsMatched,
				CrosswalkCodes: crosswalkCodes,
				IsOfficial:    isOfficial,
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

		// Convert to SICCode format with enhanced descriptions from code_metadata
		sicResults := make([]SICCode, 0, len(rankedCodes))
		for _, rankedCode := range rankedCodes {
			// Extract keywords from match details
			keywordsMatched := make([]string, 0)
			for _, match := range rankedCode.MatchDetails {
				if match.Source == "keyword" {
					keywordsMatched = append(keywordsMatched, "keyword_match")
				}
			}

			// Enhance description with official description from code_metadata if available
			description := rankedCode.Code.Description
			var crosswalkCodes []CrosswalkCode
			isOfficial := false
			
			if g.codeMetadataRepo != nil {
				enhancedDesc := g.codeMetadataRepo.EnhanceCodeDescription(ctx, "SIC", rankedCode.Code.Code, description)
				if enhancedDesc != description {
					g.logger.Printf("📝 Enhanced SIC %s description from code_metadata", rankedCode.Code.Code)
					description = enhancedDesc
				}
				
				// Get crosswalk codes (related NAICS/MCC codes)
				crosswalks, err := g.codeMetadataRepo.GetCrosswalkCodes(ctx, "SIC", rankedCode.Code.Code)
				if err == nil && len(crosswalks) > 0 {
					crosswalkCodes = make([]CrosswalkCode, len(crosswalks))
					for i, cw := range crosswalks {
						crosswalkCodes[i] = CrosswalkCode{
							CodeType: cw.CodeType,
							Code:     cw.Code,
							Name:     cw.Name,
						}
					}
					g.logger.Printf("🔗 Found %d crosswalk codes for SIC %s", len(crosswalkCodes), rankedCode.Code.Code)
				}
				
				// Check if code is official
				metadata, _ := g.codeMetadataRepo.GetCodeMetadata(ctx, "SIC", rankedCode.Code.Code)
				if metadata != nil {
					isOfficial = metadata.IsOfficial
				}
			}

			sicResults = append(sicResults, SICCode{
				Code:          rankedCode.Code.Code,
				Description:   description,
				Confidence:    rankedCode.CombinedConfidence,
				Keywords:      keywordsMatched,
				CrosswalkCodes: crosswalkCodes,
				IsOfficial:    isOfficial,
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

			// Enhance description with official description from code_metadata if available
			description := rankedCode.Code.Description
			var crosswalkCodes []CrosswalkCode
			var hierarchy *CodeHierarchy
			isOfficial := false
			
			if g.codeMetadataRepo != nil {
				enhancedDesc := g.codeMetadataRepo.EnhanceCodeDescription(ctx, "NAICS", rankedCode.Code.Code, description)
				if enhancedDesc != description {
					g.logger.Printf("📝 Enhanced NAICS %s description from code_metadata", rankedCode.Code.Code)
					description = enhancedDesc
				}
				
				// Get crosswalk codes (related SIC/MCC codes)
				crosswalks, err := g.codeMetadataRepo.GetCrosswalkCodes(ctx, "NAICS", rankedCode.Code.Code)
				if err == nil && len(crosswalks) > 0 {
					crosswalkCodes = make([]CrosswalkCode, len(crosswalks))
					for i, cw := range crosswalks {
						crosswalkCodes[i] = CrosswalkCode{
							CodeType: cw.CodeType,
							Code:     cw.Code,
							Name:     cw.Name,
						}
					}
					g.logger.Printf("🔗 Found %d crosswalk codes for NAICS %s", len(crosswalkCodes), rankedCode.Code.Code)
				}
				
				// Get hierarchy (parent/child codes) for NAICS
				parent, children, err := g.codeMetadataRepo.GetHierarchyCodes(ctx, "NAICS", rankedCode.Code.Code)
				if err == nil && (parent != nil || len(children) > 0) {
					hierarchy = &CodeHierarchy{}
					if parent != nil {
						hierarchy.ParentCode = parent.Code
						hierarchy.ParentType = parent.CodeType
					}
					if len(children) > 0 {
						hierarchy.ChildCodes = make([]string, len(children))
						for i, child := range children {
							hierarchy.ChildCodes[i] = child.Code
						}
					}
					if hierarchy.ParentCode != "" || len(hierarchy.ChildCodes) > 0 {
						g.logger.Printf("🌳 Found hierarchy for NAICS %s (parent: %s, children: %d)", 
							rankedCode.Code.Code, hierarchy.ParentCode, len(hierarchy.ChildCodes))
					}
				}
				
				// Check if code is official
				metadata, _ := g.codeMetadataRepo.GetCodeMetadata(ctx, "NAICS", rankedCode.Code.Code)
				if metadata != nil {
					isOfficial = metadata.IsOfficial
				}
			}

			naicsResults = append(naicsResults, NAICSCode{
				Code:          rankedCode.Code.Code,
				Description:   description,
				Confidence:    rankedCode.CombinedConfidence,
				Keywords:      keywordsMatched,
				CrosswalkCodes: crosswalkCodes,
				Hierarchy:     hierarchy,
				IsOfficial:    isOfficial,
			})
		}

		// Thread-safe update of shared codes
		mu.Lock()
		codes.NAICS = naicsResults
		mu.Unlock()

		g.logger.Printf("✅ NAICS code generation completed: %d codes (industries: %d, keyword: %d)",
			len(naicsResults), len(industries), len(keywordMatches))
	}()

	// Wait for all goroutines to complete with timeout (Task 2.2: Enhanced parallel execution)
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	
	// Wait with context timeout
	select {
	case <-done:
		// All goroutines completed successfully
		close(errorChan)
	case <-ctx.Done():
		// Context cancelled or timed out
		g.logger.Printf("⚠️ Parallel code generation timed out or cancelled")
		close(errorChan)
		return
	}

	// Log any errors that occurred
	errorCount := 0
	for err := range errorChan {
		errorCount++
		g.logger.Printf("⚠️ Error in parallel code generation: %v", err)
	}
	
	if errorCount > 0 {
		g.logger.Printf("⚠️ Parallel code generation completed with %d errors", errorCount)
	} else {
		g.logger.Printf("🚀 Parallel code generation completed successfully: %d MCC, %d SIC, %d NAICS codes",
			len(codes.MCC), len(codes.SIC), len(codes.NAICS))
	}
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

// ============================================================================
// PHASE 2: Crosswalk-Enhanced Code Generation
// ============================================================================

// validateAndEnhanceCodesWithCrosswalks validates generated codes against crosswalk relationships
// and infers missing codes from crosswalks
func (g *ClassificationCodeGenerator) validateAndEnhanceCodesWithCrosswalks(
	ctx context.Context,
	codes *ClassificationCodesInfo,
) (*ClassificationCodesInfo, error) {
	if g.codeMetadataRepo == nil {
		return codes, nil // No metadata repo available
	}

	enhancedCodes := &ClassificationCodesInfo{
		MCC:   make([]MCCCode, len(codes.MCC)),
		SIC:   make([]SICCode, len(codes.SIC)),
		NAICS: make([]NAICSCode, len(codes.NAICS)),
	}
	copy(enhancedCodes.MCC, codes.MCC)
	copy(enhancedCodes.SIC, codes.SIC)
	copy(enhancedCodes.NAICS, codes.NAICS)

	// Step 1: Validate existing codes and infer missing related codes
	enhancedCodes = g.inferMissingCodesFromCrosswalks(ctx, enhancedCodes)

	// Step 2: Calculate crosswalk consistency score
	consistencyScore := g.calculateCrosswalkConsistency(ctx, enhancedCodes)
	g.logger.Printf("📊 [Phase 2] Crosswalk consistency score: %.2f", consistencyScore)

	// Step 3: Boost confidence for codes with high crosswalk consistency
	if consistencyScore > 0.8 {
		enhancedCodes = g.boostConfidenceForConsistentCodes(enhancedCodes, consistencyScore)
		g.logger.Printf("✅ [Phase 2] Boosted confidence for codes with high crosswalk consistency (%.2f)", consistencyScore)
	}

	return enhancedCodes, nil
}

// inferMissingCodesFromCrosswalks infers missing codes from crosswalk relationships
func (g *ClassificationCodeGenerator) inferMissingCodesFromCrosswalks(
	ctx context.Context,
	codes *ClassificationCodesInfo,
) *ClassificationCodesInfo {
	enhancedCodes := &ClassificationCodesInfo{
		MCC:   make([]MCCCode, 0, len(codes.MCC)),
		SIC:   make([]SICCode, 0, len(codes.SIC)),
		NAICS: make([]NAICSCode, 0, len(codes.NAICS)),
	}

	// Track which codes we already have
	existingCodes := make(map[string]map[string]bool) // codeType -> code -> true
	existingCodes["MCC"] = make(map[string]bool)
	existingCodes["SIC"] = make(map[string]bool)
	existingCodes["NAICS"] = make(map[string]bool)

	// Add existing codes
	for _, mcc := range codes.MCC {
		existingCodes["MCC"][mcc.Code] = true
		enhancedCodes.MCC = append(enhancedCodes.MCC, mcc)
	}
	for _, sic := range codes.SIC {
		existingCodes["SIC"][sic.Code] = true
		enhancedCodes.SIC = append(enhancedCodes.SIC, sic)
	}
	for _, naics := range codes.NAICS {
		existingCodes["NAICS"][naics.Code] = true
		enhancedCodes.NAICS = append(enhancedCodes.NAICS, naics)
	}

	// Infer missing codes from crosswalks
	// If we have MCC codes, infer NAICS/SIC from crosswalks
	for _, mcc := range codes.MCC {
		crosswalks, err := g.codeMetadataRepo.GetCrosswalkCodes(ctx, "MCC", mcc.Code)
		if err == nil && len(crosswalks) > 0 {
			for _, cw := range crosswalks {
				if !existingCodes[cw.CodeType][cw.Code] {
					// Generate code from crosswalk
					inferredCode := g.generateCodeFromCrosswalk(cw, mcc.Confidence*0.75, mcc.Code, "MCC")
					if inferredCode != nil {
						switch cw.CodeType {
						case "NAICS":
							enhancedCodes.NAICS = append(enhancedCodes.NAICS, *inferredCode.(*NAICSCode))
							existingCodes["NAICS"][cw.Code] = true
							g.logger.Printf("🔗 [Phase 2] Inferred NAICS %s from MCC %s crosswalk", cw.Code, mcc.Code)
						case "SIC":
							enhancedCodes.SIC = append(enhancedCodes.SIC, *inferredCode.(*SICCode))
							existingCodes["SIC"][cw.Code] = true
							g.logger.Printf("🔗 [Phase 2] Inferred SIC %s from MCC %s crosswalk", cw.Code, mcc.Code)
						}
					}
				}
			}
		}
	}

	// If we have NAICS codes, infer MCC/SIC from crosswalks
	for _, naics := range codes.NAICS {
		crosswalks, err := g.codeMetadataRepo.GetCrosswalkCodes(ctx, "NAICS", naics.Code)
		if err == nil && len(crosswalks) > 0 {
			for _, cw := range crosswalks {
				if !existingCodes[cw.CodeType][cw.Code] {
					inferredCode := g.generateCodeFromCrosswalk(cw, naics.Confidence*0.75, naics.Code, "NAICS")
					if inferredCode != nil {
						switch cw.CodeType {
						case "MCC":
							enhancedCodes.MCC = append(enhancedCodes.MCC, *inferredCode.(*MCCCode))
							existingCodes["MCC"][cw.Code] = true
							g.logger.Printf("🔗 [Phase 2] Inferred MCC %s from NAICS %s crosswalk", cw.Code, naics.Code)
						case "SIC":
							enhancedCodes.SIC = append(enhancedCodes.SIC, *inferredCode.(*SICCode))
							existingCodes["SIC"][cw.Code] = true
							g.logger.Printf("🔗 [Phase 2] Inferred SIC %s from NAICS %s crosswalk", cw.Code, naics.Code)
						}
					}
				}
			}
		}
	}

	// If we have SIC codes, infer MCC/NAICS from crosswalks
	for _, sic := range codes.SIC {
		crosswalks, err := g.codeMetadataRepo.GetCrosswalkCodes(ctx, "SIC", sic.Code)
		if err == nil && len(crosswalks) > 0 {
			for _, cw := range crosswalks {
				if !existingCodes[cw.CodeType][cw.Code] {
					inferredCode := g.generateCodeFromCrosswalk(cw, sic.Confidence*0.75, sic.Code, "SIC")
					if inferredCode != nil {
						switch cw.CodeType {
						case "MCC":
							enhancedCodes.MCC = append(enhancedCodes.MCC, *inferredCode.(*MCCCode))
							existingCodes["MCC"][cw.Code] = true
							g.logger.Printf("🔗 [Phase 2] Inferred MCC %s from SIC %s crosswalk", cw.Code, sic.Code)
						case "NAICS":
							enhancedCodes.NAICS = append(enhancedCodes.NAICS, *inferredCode.(*NAICSCode))
							existingCodes["NAICS"][cw.Code] = true
							g.logger.Printf("🔗 [Phase 2] Inferred NAICS %s from SIC %s crosswalk", cw.Code, sic.Code)
						}
					}
				}
			}
		}
	}

	return enhancedCodes
}

// generateCodeFromCrosswalk generates a code from crosswalk relationship
func (g *ClassificationCodeGenerator) generateCodeFromCrosswalk(
	cw struct {
		CodeType string
		Code     string
		Name     string
	},
	confidence float64,
	sourceCode string,
	sourceType string,
) interface{} {
	// Get code metadata for description
	metadata, err := g.codeMetadataRepo.GetCodeMetadata(context.Background(), cw.CodeType, cw.Code)
	if err != nil {
		g.logger.Printf("⚠️ [Phase 2] Failed to get metadata for %s %s: %v", cw.CodeType, cw.Code, err)
		return nil
	}

	description := cw.Name
	if metadata != nil && metadata.OfficialDescription != "" {
		description = metadata.OfficialDescription
	}

	switch cw.CodeType {
	case "MCC":
		return &MCCCode{
			Code:        cw.Code,
			Description: description,
			Confidence:  confidence,
			Keywords:    []string{fmt.Sprintf("inferred_from_%s_%s", sourceType, sourceCode)},
			CrosswalkCodes: []CrosswalkCode{
				{CodeType: sourceType, Code: sourceCode},
			},
			IsOfficial: metadata != nil && metadata.IsOfficial,
		}
	case "SIC":
		return &SICCode{
			Code:        cw.Code,
			Description: description,
			Confidence:  confidence,
			Keywords:    []string{fmt.Sprintf("inferred_from_%s_%s", sourceType, sourceCode)},
			CrosswalkCodes: []CrosswalkCode{
				{CodeType: sourceType, Code: sourceCode},
			},
			IsOfficial: metadata != nil && metadata.IsOfficial,
		}
	case "NAICS":
		return &NAICSCode{
			Code:        cw.Code,
			Description: description,
			Confidence:  confidence,
			Keywords:    []string{fmt.Sprintf("inferred_from_%s_%s", sourceType, sourceCode)},
			CrosswalkCodes: []CrosswalkCode{
				{CodeType: sourceType, Code: sourceCode},
			},
			IsOfficial: metadata != nil && metadata.IsOfficial,
		}
	default:
		return nil
	}
}

// calculateCrosswalkConsistency calculates how well generated codes match crosswalk relationships
func (g *ClassificationCodeGenerator) calculateCrosswalkConsistency(
	ctx context.Context,
	codes *ClassificationCodesInfo,
) float64 {
	if g.codeMetadataRepo == nil {
		return 0.0
	}

	totalChecks := 0
	consistentChecks := 0

	// Check MCC codes
	for _, mcc := range codes.MCC {
		crosswalks, err := g.codeMetadataRepo.GetCrosswalkCodes(ctx, "MCC", mcc.Code)
		if err == nil && len(crosswalks) > 0 {
			for _, cw := range crosswalks {
				totalChecks++
				if g.codeExistsInResults(cw.Code, cw.CodeType, codes) {
					consistentChecks++
				}
			}
		}
	}

	// Check NAICS codes
	for _, naics := range codes.NAICS {
		crosswalks, err := g.codeMetadataRepo.GetCrosswalkCodes(ctx, "NAICS", naics.Code)
		if err == nil && len(crosswalks) > 0 {
			for _, cw := range crosswalks {
				totalChecks++
				if g.codeExistsInResults(cw.Code, cw.CodeType, codes) {
					consistentChecks++
				}
			}
		}
	}

	// Check SIC codes
	for _, sic := range codes.SIC {
		crosswalks, err := g.codeMetadataRepo.GetCrosswalkCodes(ctx, "SIC", sic.Code)
		if err == nil && len(crosswalks) > 0 {
			for _, cw := range crosswalks {
				totalChecks++
				if g.codeExistsInResults(cw.Code, cw.CodeType, codes) {
					consistentChecks++
				}
			}
		}
	}

	if totalChecks == 0 {
		return 0.0 // No crosswalks to check
	}

	return float64(consistentChecks) / float64(totalChecks)
}

// codeExistsInResults checks if a code exists in the generated results
func (g *ClassificationCodeGenerator) codeExistsInResults(code, codeType string, codes *ClassificationCodesInfo) bool {
	switch codeType {
	case "MCC":
		for _, mcc := range codes.MCC {
			if mcc.Code == code {
				return true
			}
		}
	case "SIC":
		for _, sic := range codes.SIC {
			if sic.Code == code {
				return true
			}
		}
	case "NAICS":
		for _, naics := range codes.NAICS {
			if naics.Code == code {
				return true
			}
		}
	}
	return false
}

// boostConfidenceForConsistentCodes boosts confidence for codes with high crosswalk consistency
func (g *ClassificationCodeGenerator) boostConfidenceForConsistentCodes(
	codes *ClassificationCodesInfo,
	consistencyScore float64,
) *ClassificationCodesInfo {
	boostFactor := 1.0 + (consistencyScore * 0.1) // Up to 10% boost for perfect consistency

	enhancedCodes := &ClassificationCodesInfo{
		MCC:   make([]MCCCode, len(codes.MCC)),
		SIC:   make([]SICCode, len(codes.SIC)),
		NAICS: make([]NAICSCode, len(codes.NAICS)),
	}

	// Boost MCC codes
	for i, mcc := range codes.MCC {
		newConfidence := mcc.Confidence * boostFactor
		if newConfidence > 1.0 {
			newConfidence = 1.0
		}
		enhancedCodes.MCC[i] = mcc
		enhancedCodes.MCC[i].Confidence = newConfidence
	}

	// Boost SIC codes
	for i, sic := range codes.SIC {
		newConfidence := sic.Confidence * boostFactor
		if newConfidence > 1.0 {
			newConfidence = 1.0
		}
		enhancedCodes.SIC[i] = sic
		enhancedCodes.SIC[i].Confidence = newConfidence
	}

	// Boost NAICS codes
	for i, naics := range codes.NAICS {
		newConfidence := naics.Confidence * boostFactor
		if newConfidence > 1.0 {
			newConfidence = 1.0
		}
		enhancedCodes.NAICS[i] = naics
		enhancedCodes.NAICS[i].Confidence = newConfidence
	}

	return enhancedCodes
}
