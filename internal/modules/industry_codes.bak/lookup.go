package industry_codes

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// LookupResult represents a single lookup result with relevance score
type LookupResult struct {
	Code      *IndustryCode `json:"code"`
	Relevance float64       `json:"relevance"`
	MatchType string        `json:"match_type"`
	MatchedOn []string      `json:"matched_on"`
}

// LookupRequest represents a request to lookup industry codes
type LookupRequest struct {
	Query         string     `json:"query"`
	CodeTypes     []CodeType `json:"code_types,omitempty"`
	Categories    []string   `json:"categories,omitempty"`
	MinConfidence float64    `json:"min_confidence,omitempty"`
	MaxResults    int        `json:"max_results,omitempty"`
}

// LookupResponse represents the response from a lookup operation
type LookupResponse struct {
	Query         string          `json:"query"`
	Results       []*LookupResult `json:"results"`
	TotalFound    int             `json:"total_found"`
	SearchTime    time.Duration   `json:"search_time"`
	CodeTypeStats map[string]int  `json:"code_type_stats"`
}

// IndustryCodeLookup provides high-level lookup functionality for industry codes
type IndustryCodeLookup struct {
	db     *IndustryCodeDatabase
	logger *zap.Logger
}

// NewIndustryCodeLookup creates a new industry code lookup instance
func NewIndustryCodeLookup(db *IndustryCodeDatabase, logger *zap.Logger) *IndustryCodeLookup {
	return &IndustryCodeLookup{
		db:     db,
		logger: logger,
	}
}

// Lookup performs a comprehensive search for industry codes based on the request
func (icl *IndustryCodeLookup) Lookup(ctx context.Context, req *LookupRequest) (*LookupResponse, error) {
	startTime := time.Now()

	if req.MaxResults == 0 {
		req.MaxResults = 50 // Default limit
	}

	// Perform the search
	results, err := icl.performSearch(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform lookup: %w", err)
	}

	// Calculate statistics
	codeTypeStats := icl.calculateCodeTypeStats(results)

	response := &LookupResponse{
		Query:         req.Query,
		Results:       results,
		TotalFound:    len(results),
		SearchTime:    time.Since(startTime),
		CodeTypeStats: codeTypeStats,
	}

	icl.logger.Info("Industry code lookup completed",
		zap.String("query", req.Query),
		zap.Int("results", len(results)),
		zap.Duration("search_time", response.SearchTime))

	return response, nil
}

// performSearch executes the actual search logic
func (icl *IndustryCodeLookup) performSearch(ctx context.Context, req *LookupRequest) ([]*LookupResult, error) {
	var allResults []*LookupResult

	// If no specific code types specified, search all types
	if len(req.CodeTypes) == 0 {
		req.CodeTypes = []CodeType{CodeTypeMCC, CodeTypeSIC, CodeTypeNAICS}
	}

	// Search for each code type
	for _, codeType := range req.CodeTypes {
		results, err := icl.searchByCodeType(ctx, req, codeType)
		if err != nil {
			icl.logger.Warn("Failed to search code type",
				zap.String("code_type", string(codeType)),
				zap.Error(err))
			continue
		}
		allResults = append(allResults, results...)
	}

	// Apply confidence filter
	if req.MinConfidence > 0 {
		filteredResults := make([]*LookupResult, 0)
		for _, result := range allResults {
			if result.Code.Confidence >= req.MinConfidence {
				filteredResults = append(filteredResults, result)
			}
		}
		allResults = filteredResults
	}

	// Apply category filter
	if len(req.Categories) > 0 {
		filteredResults := make([]*LookupResult, 0)
		for _, result := range allResults {
			for _, category := range req.Categories {
				if strings.Contains(strings.ToLower(result.Code.Category), strings.ToLower(category)) {
					filteredResults = append(filteredResults, result)
					break
				}
			}
		}
		allResults = filteredResults
	}

	// Sort by relevance and confidence
	sort.Slice(allResults, func(i, j int) bool {
		if allResults[i].Relevance != allResults[j].Relevance {
			return allResults[i].Relevance > allResults[j].Relevance
		}
		return allResults[i].Code.Confidence > allResults[j].Code.Confidence
	})

	// Limit results
	if len(allResults) > req.MaxResults {
		allResults = allResults[:req.MaxResults]
	}

	return allResults, nil
}

// searchByCodeType searches for codes of a specific type
func (icl *IndustryCodeLookup) searchByCodeType(ctx context.Context, req *LookupRequest, codeType CodeType) ([]*LookupResult, error) {
	// First try exact code match
	if icl.isExactCodeMatch(req.Query) {
		code, err := icl.db.GetCodeByCodeAndType(ctx, req.Query, codeType)
		if err == nil {
			return []*LookupResult{{
				Code:      code,
				Relevance: 1.0,
				MatchType: "exact_code",
				MatchedOn: []string{"exact_code"},
			}}, nil
		}
	}

	// Perform text search
	codes, err := icl.db.SearchCodes(ctx, req.Query, &codeType, req.MaxResults)
	if err != nil {
		return nil, err
	}

	var results []*LookupResult
	for _, code := range codes {
		relevance, matchType, matchedOn := icl.calculateRelevance(req.Query, code)
		results = append(results, &LookupResult{
			Code:      code,
			Relevance: relevance,
			MatchType: matchType,
			MatchedOn: matchedOn,
		})
	}

	return results, nil
}

// calculateRelevance calculates the relevance score for a code based on the query
func (icl *IndustryCodeLookup) calculateRelevance(query string, code *IndustryCode) (float64, string, []string) {
	query = strings.ToLower(query)
	matchedOn := make([]string, 0)
	totalScore := 0.0

	// Check exact code match
	if strings.ToLower(code.Code) == query {
		matchedOn = append(matchedOn, "exact_code")
		totalScore += 1.0
	}

	// Check description match
	description := strings.ToLower(code.Description)
	if strings.Contains(description, query) {
		matchedOn = append(matchedOn, "description")
		totalScore += 0.8
	}

	// Check category match
	category := strings.ToLower(code.Category)
	if strings.Contains(category, query) {
		matchedOn = append(matchedOn, "category")
		totalScore += 0.6
	}

	// Check subcategory match
	if code.Subcategory != "" {
		subcategory := strings.ToLower(code.Subcategory)
		if strings.Contains(subcategory, query) {
			matchedOn = append(matchedOn, "subcategory")
			totalScore += 0.4
		}
	}

	// Check keywords match
	for _, keyword := range code.Keywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(keywordLower, query) {
			matchedOn = append(matchedOn, "keywords")
			totalScore += 0.3
			break
		}
	}

	// Determine match type
	var matchType string
	if totalScore >= 1.0 {
		matchType = "exact"
	} else if totalScore >= 0.8 {
		matchType = "high"
	} else if totalScore >= 0.5 {
		matchType = "medium"
	} else {
		matchType = "low"
	}

	// Normalize score to 0-1 range
	relevance := totalScore / 1.0
	if relevance > 1.0 {
		relevance = 1.0
	}

	return relevance, matchType, matchedOn
}

// isExactCodeMatch checks if the query looks like an exact code
func (icl *IndustryCodeLookup) isExactCodeMatch(query string) bool {
	// Check if the original query contains any non-digit characters (except separators)
	originalDigitsOnly := true
	for _, r := range query {
		if r != '-' && r != ' ' && (r < '0' || r > '9') {
			originalDigitsOnly = false
			break
		}
	}

	// If original query contains non-digit characters, it's not an exact code match
	if !originalDigitsOnly {
		return false
	}

	// Handle SIC codes with dash (e.g., "5411-0")
	if strings.Contains(query, "-") {
		parts := strings.Split(query, "-")
		if len(parts) == 2 && len(parts[0]) == 4 && len(parts[1]) == 1 && isAllDigits(parts[0]) && isAllDigits(parts[1]) {
			return true // SIC with dash (e.g., "5411-0")
		}
	}

	// Handle NAICS codes with space (e.g., "5411 00")
	if strings.Contains(query, " ") {
		parts := strings.Fields(query)
		if len(parts) == 2 && len(parts[0]) == 4 && len(parts[1]) == 2 && isAllDigits(parts[0]) && isAllDigits(parts[1]) {
			return true // NAICS with space (e.g., "5411 00")
		}
	}

	// Handle NAICS codes with dash (e.g., "5411-00")
	if strings.Contains(query, "-") {
		parts := strings.Split(query, "-")
		if len(parts) == 2 && len(parts[0]) == 4 && len(parts[1]) == 2 && isAllDigits(parts[0]) && isAllDigits(parts[1]) {
			return true // NAICS with dash (e.g., "5411-00")
		}
	}

	// Remove common separators for remaining checks
	cleanQuery := strings.ReplaceAll(query, "-", "")
	cleanQuery = strings.ReplaceAll(cleanQuery, " ", "")

	// Check if it's all digits (SIC, NAICS) or 4 digits (MCC)
	if len(cleanQuery) == 4 && isAllDigits(cleanQuery) {
		return true // MCC or SIC (4 digits)
	}
	if len(cleanQuery) == 6 && isAllDigits(cleanQuery) {
		return true // NAICS (6 digits)
	}

	return false
}

// isAllDigits checks if a string contains only digits
func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// calculateCodeTypeStats calculates statistics by code type
func (icl *IndustryCodeLookup) calculateCodeTypeStats(results []*LookupResult) map[string]int {
	stats := make(map[string]int)
	for _, result := range results {
		codeType := string(result.Code.Type)
		stats[codeType]++
	}
	return stats
}

// GetTopCodesByType retrieves the top codes by confidence for each code type
func (icl *IndustryCodeLookup) GetTopCodesByType(ctx context.Context, limit int) (map[CodeType][]*IndustryCode, error) {
	codeTypes := []CodeType{CodeTypeMCC, CodeTypeSIC, CodeTypeNAICS}
	results := make(map[CodeType][]*IndustryCode)

	for _, codeType := range codeTypes {
		codes, err := icl.db.GetCodesByType(ctx, codeType, limit, 0)
		if err != nil {
			icl.logger.Warn("Failed to get top codes for type",
				zap.String("code_type", string(codeType)),
				zap.Error(err))
			continue
		}

		// Sort by confidence
		sort.Slice(codes, func(i, j int) bool {
			return codes[i].Confidence > codes[j].Confidence
		})

		results[codeType] = codes
	}

	return results, nil
}

// GetCodesByCategory retrieves codes by category with relevance scoring
func (icl *IndustryCodeLookup) GetCodesByCategory(ctx context.Context, category string, codeTypes []CodeType, limit int) ([]*LookupResult, error) {
	if len(codeTypes) == 0 {
		codeTypes = []CodeType{CodeTypeMCC, CodeTypeSIC, CodeTypeNAICS}
	}

	var allResults []*LookupResult

	for _, codeType := range codeTypes {
		codes, err := icl.db.GetCodesByCategory(ctx, category, &codeType, limit)
		if err != nil {
			icl.logger.Warn("Failed to get codes by category",
				zap.String("category", category),
				zap.String("code_type", string(codeType)),
				zap.Error(err))
			continue
		}

		for _, code := range codes {
			relevance, matchType, matchedOn := icl.calculateRelevance(category, code)
			allResults = append(allResults, &LookupResult{
				Code:      code,
				Relevance: relevance,
				MatchType: matchType,
				MatchedOn: matchedOn,
			})
		}
	}

	// Sort by relevance and confidence
	sort.Slice(allResults, func(i, j int) bool {
		if allResults[i].Relevance != allResults[j].Relevance {
			return allResults[i].Relevance > allResults[j].Relevance
		}
		return allResults[i].Code.Confidence > allResults[j].Code.Confidence
	})

	// Limit results
	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	return allResults, nil
}

// GetCodeSuggestions provides autocomplete suggestions for industry codes
func (icl *IndustryCodeLookup) GetCodeSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if len(query) < 2 {
		return []string{}, nil
	}

	// Search for codes that start with the query
	req := &LookupRequest{
		Query:      query,
		MaxResults: limit * 2, // Get more results to filter
	}

	results, err := icl.Lookup(ctx, req)
	if err != nil {
		return nil, err
	}

	// Extract unique suggestions
	suggestions := make(map[string]bool)
	var uniqueSuggestions []string

	for _, result := range results.Results {
		// Add code as suggestion
		if !suggestions[result.Code.Code] {
			suggestions[result.Code.Code] = true
			uniqueSuggestions = append(uniqueSuggestions, result.Code.Code)
		}

		// Add description words as suggestions
		words := strings.Fields(result.Code.Description)
		for _, word := range words {
			word = strings.ToLower(word)
			if len(word) >= 3 && strings.HasPrefix(word, strings.ToLower(query)) {
				if !suggestions[word] {
					suggestions[word] = true
					uniqueSuggestions = append(uniqueSuggestions, word)
				}
			}
		}

		if len(uniqueSuggestions) >= limit {
			break
		}
	}

	// Limit to requested number
	if len(uniqueSuggestions) > limit {
		uniqueSuggestions = uniqueSuggestions[:limit]
	}

	return uniqueSuggestions, nil
}

// GetCodeStats retrieves statistics about the industry codes database
func (icl *IndustryCodeLookup) GetCodeStats(ctx context.Context) (map[string]interface{}, error) {
	return icl.db.GetCodeStats(ctx)
}
