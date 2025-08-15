package webanalysis

import (
	"strings"
	"sync"
)

// SearchResultFilter provides comprehensive filtering and ranking for search results
type SearchResultFilter struct {
	filters []ResultFilter
	rankers []ResultRanker
	config  FilterConfig
	mu      sync.RWMutex
}

// FilterConfig holds configuration for search result filtering
type FilterConfig struct {
	MinRelevanceScore        float64  `json:"min_relevance_score"`
	MaxResultsToReturn       int      `json:"max_results_to_return"`
	EnableSpamFiltering      bool     `json:"enable_spam_filtering"`
	EnableDuplicateFiltering bool     `json:"enable_duplicate_filtering"`
	EnableQualityFiltering   bool     `json:"enable_quality_filtering"`
	EnableDomainFiltering    bool     `json:"enable_domain_filtering"`
	EnableLanguageFiltering  bool     `json:"enable_language_filtering"`
	EnableDateFiltering      bool     `json:"enable_date_filtering"`
	EnableContentFiltering   bool     `json:"enable_content_filtering"`
	SpamKeywords             []string `json:"spam_keywords"`
	BlockedDomains           []string `json:"blocked_domains"`
	PreferredDomains         []string `json:"preferred_domains"`
	AllowedLanguages         []string `json:"allowed_languages"`
	MinContentLength         int      `json:"min_content_length"`
	MaxContentLength         int      `json:"max_content_length"`
}

// ResultFilter represents a filter for search results
type ResultFilter interface {
	Filter(result *WebSearchResult) bool
	GetName() string
	GetDescription() string
}

// ResultRanker represents a ranker for search results
type ResultRanker interface {
	Rank(result *WebSearchResult, query string) float64
	GetName() string
	GetDescription() string
}

// FilteredSearchResult represents a filtered and ranked search result
type FilteredSearchResult struct {
	Result         *WebSearchResult `json:"result"`
	FilteredBy     []string         `json:"filtered_by"`
	RankingScore   float64          `json:"ranking_score"`
	QualityScore   float64          `json:"quality_score"`
	RelevanceScore float64          `json:"relevance_score"`
	FinalScore     float64          `json:"final_score"`
}

// NewSearchResultFilter creates a new search result filter
func NewSearchResultFilter() *SearchResultFilter {
	config := FilterConfig{
		MinRelevanceScore:        0.3,
		MaxResultsToReturn:       20,
		EnableSpamFiltering:      true,
		EnableDuplicateFiltering: true,
		EnableQualityFiltering:   true,
		EnableDomainFiltering:    true,
		EnableLanguageFiltering:  true,
		EnableDateFiltering:      false,
		EnableContentFiltering:   true,
		SpamKeywords: []string{
			"click here", "buy now", "limited time", "act now", "free trial",
			"make money", "earn money", "work from home", "get rich",
			"lose weight", "diet pills", "miracle cure", "100% free",
			"no cost", "no obligation", "guaranteed", "risk-free",
		},
		BlockedDomains: []string{
			"spam.com", "malware.com", "phishing.com",
		},
		PreferredDomains: []string{
			"wikipedia.org", "linkedin.com", "crunchbase.com", "bloomberg.com",
			"reuters.com", "forbes.com", "techcrunch.com", "wsj.com",
		},
		AllowedLanguages: []string{"en", "en-US", "en-GB"},
		MinContentLength: 50,
		MaxContentLength: 10000,
	}

	filter := &SearchResultFilter{
		config: config,
	}

	// Initialize filters
	filter.initializeFilters()

	// Initialize rankers
	filter.initializeRankers()

	return filter
}

// initializeFilters sets up all available filters
func (srf *SearchResultFilter) initializeFilters() {
	srf.filters = []ResultFilter{
		&RelevanceFilter{config: srf.config},
		&SpamFilter{config: srf.config},
		&DomainFilter{config: srf.config},
		&ContentFilter{config: srf.config},
		&DuplicateFilter{},
		&LanguageFilter{config: srf.config},
		&QualityFilter{config: srf.config},
	}
}

// initializeRankers sets up all available rankers
func (srf *SearchResultFilter) initializeRankers() {
	srf.rankers = []ResultRanker{
		&RelevanceRanker{},
		&DomainRanker{config: srf.config},
		&ContentRanker{},
		&FreshnessRanker{},
		&AuthorityRanker{},
		&PopularityRanker{},
	}
}

// FilterAndRank filters and ranks search results
func (srf *SearchResultFilter) FilterAndRank(results []WebSearchResult, query string) []FilteredSearchResult {
	srf.mu.RLock()
	defer srf.mu.RUnlock()

	var filteredResults []FilteredSearchResult

	for _, result := range results {
		// Apply filters
		filteredBy := srf.applyFilters(&result)

		// If result passes all filters, rank it
		if len(filteredBy) == 0 {
			rankingScore := srf.calculateRankingScore(&result, query)
			qualityScore := srf.calculateQualityScore(&result)
			relevanceScore := srf.calculateRelevanceScore(&result, query)
			finalScore := srf.calculateFinalScore(rankingScore, qualityScore, relevanceScore)

			filteredResults = append(filteredResults, FilteredSearchResult{
				Result:         &result,
				FilteredBy:     filteredBy,
				RankingScore:   rankingScore,
				QualityScore:   qualityScore,
				RelevanceScore: relevanceScore,
				FinalScore:     finalScore,
			})
		}
	}

	// Sort by final score (descending)
	srf.sortByFinalScore(filteredResults)

	// Limit results
	if len(filteredResults) > srf.config.MaxResultsToReturn {
		filteredResults = filteredResults[:srf.config.MaxResultsToReturn]
	}

	return filteredResults
}

// applyFilters applies all filters to a result
func (srf *SearchResultFilter) applyFilters(result *WebSearchResult) []string {
	var filteredBy []string

	for _, filter := range srf.filters {
		if !filter.Filter(result) {
			filteredBy = append(filteredBy, filter.GetName())
		}
	}

	return filteredBy
}

// calculateRankingScore calculates the overall ranking score
func (srf *SearchResultFilter) calculateRankingScore(result *WebSearchResult, query string) float64 {
	totalScore := 0.0
	totalWeight := 0.0

	for _, ranker := range srf.rankers {
		score := ranker.Rank(result, query)
		weight := srf.getRankerWeight(ranker.GetName())

		totalScore += score * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

// calculateQualityScore calculates the quality score
func (srf *SearchResultFilter) calculateQualityScore(result *WebSearchResult) float64 {
	score := 0.0

	// URL quality
	if strings.HasPrefix(result.URL, "https://") {
		score += 0.3
	} else if strings.HasPrefix(result.URL, "http://") {
		score += 0.2
	}

	// Content length
	contentLength := len(result.Title) + len(result.Description)
	if contentLength >= srf.config.MinContentLength && contentLength <= srf.config.MaxContentLength {
		score += 0.2
	}

	// Domain quality
	if srf.isPreferredDomain(result.URL) {
		score += 0.3
	}

	// Content quality indicators
	if strings.Contains(strings.ToLower(result.Title), "official") {
		score += 0.1
	}
	if strings.Contains(strings.ToLower(result.Title), "about") {
		score += 0.1
	}

	return score
}

// calculateRelevanceScore calculates the relevance score
func (srf *SearchResultFilter) calculateRelevanceScore(result *WebSearchResult, query string) float64 {
	queryTerms := strings.Fields(strings.ToLower(query))
	title := strings.ToLower(result.Title)
	description := strings.ToLower(result.Description)
	content := title + " " + description

	score := 0.0
	matchedTerms := 0

	for _, term := range queryTerms {
		if strings.Contains(title, term) {
			score += 0.4
			matchedTerms++
		}
		if strings.Contains(description, term) {
			score += 0.2
			matchedTerms++
		}
	}

	// Bonus for exact phrase match
	if strings.Contains(content, strings.ToLower(query)) {
		score += 0.3
	}

	// Normalize by number of query terms
	if len(queryTerms) > 0 {
		score = score / float64(len(queryTerms))
	}

	return score
}

// calculateFinalScore calculates the final combined score
func (srf *SearchResultFilter) calculateFinalScore(rankingScore, qualityScore, relevanceScore float64) float64 {
	// Weighted combination
	return rankingScore*0.4 + qualityScore*0.3 + relevanceScore*0.3
}

// getRankerWeight returns the weight for a specific ranker
func (srf *SearchResultFilter) getRankerWeight(rankerName string) float64 {
	weights := map[string]float64{
		"RelevanceRanker":  0.3,
		"DomainRanker":     0.2,
		"ContentRanker":    0.2,
		"FreshnessRanker":  0.1,
		"AuthorityRanker":  0.1,
		"PopularityRanker": 0.1,
	}

	if weight, exists := weights[rankerName]; exists {
		return weight
	}
	return 0.1 // Default weight
}

// sortByFinalScore sorts results by final score in descending order
func (srf *SearchResultFilter) sortByFinalScore(results []FilteredSearchResult) {
	// Simple bubble sort for small lists
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			if results[j].FinalScore < results[j+1].FinalScore {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// isPreferredDomain checks if a URL is from a preferred domain
func (srf *SearchResultFilter) isPreferredDomain(url string) bool {
	for _, domain := range srf.config.PreferredDomains {
		if strings.Contains(url, domain) {
			return true
		}
	}
	return false
}

// UpdateConfig updates the filter configuration
func (srf *SearchResultFilter) UpdateConfig(config FilterConfig) {
	srf.mu.Lock()
	defer srf.mu.Unlock()
	srf.config = config
}

// GetConfig returns the current configuration
func (srf *SearchResultFilter) GetConfig() FilterConfig {
	srf.mu.RLock()
	defer srf.mu.RUnlock()
	return srf.config
}

// GetStats returns statistics about filtering and ranking
func (srf *SearchResultFilter) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_filters": len(srf.filters),
		"total_rankers": len(srf.rankers),
		"config":        srf.config,
	}
}

// RelevanceFilter filters results based on relevance score
type RelevanceFilter struct {
	config FilterConfig
}

func (rf *RelevanceFilter) Filter(result *WebSearchResult) bool {
	return result.RelevanceScore >= rf.config.MinRelevanceScore
}

func (rf *RelevanceFilter) GetName() string {
	return "RelevanceFilter"
}

func (rf *RelevanceFilter) GetDescription() string {
	return "Filters results based on minimum relevance score"
}

// SpamFilter filters out spam results
type SpamFilter struct {
	config FilterConfig
}

func (sf *SpamFilter) Filter(result *WebSearchResult) bool {
	if !sf.config.EnableSpamFiltering {
		return true
	}

	content := strings.ToLower(result.Title + " " + result.Description)

	for _, keyword := range sf.config.SpamKeywords {
		if strings.Contains(content, strings.ToLower(keyword)) {
			return false
		}
	}

	return true
}

func (sf *SpamFilter) GetName() string {
	return "SpamFilter"
}

func (sf *SpamFilter) GetDescription() string {
	return "Filters out spam results based on keywords"
}

// DomainFilter filters results based on domain
type DomainFilter struct {
	config FilterConfig
}

func (df *DomainFilter) Filter(result *WebSearchResult) bool {
	if !df.config.EnableDomainFiltering {
		return true
	}

	// Check blocked domains
	for _, domain := range df.config.BlockedDomains {
		if strings.Contains(result.URL, domain) {
			return false
		}
	}

	return true
}

func (df *DomainFilter) GetName() string {
	return "DomainFilter"
}

func (df *DomainFilter) GetDescription() string {
	return "Filters results based on domain restrictions"
}

// ContentFilter filters results based on content quality
type ContentFilter struct {
	config FilterConfig
}

func (cf *ContentFilter) Filter(result *WebSearchResult) bool {
	if !cf.config.EnableContentFiltering {
		return true
	}

	contentLength := len(result.Title) + len(result.Description)

	if contentLength < cf.config.MinContentLength {
		return false
	}

	if contentLength > cf.config.MaxContentLength {
		return false
	}

	return true
}

func (cf *ContentFilter) GetName() string {
	return "ContentFilter"
}

func (cf *ContentFilter) GetDescription() string {
	return "Filters results based on content length and quality"
}

// DuplicateFilter filters out duplicate results
type DuplicateFilter struct {
	seenURLs map[string]bool
	mu       sync.RWMutex
}

func (df *DuplicateFilter) Filter(result *WebSearchResult) bool {
	df.mu.Lock()
	defer df.mu.Unlock()

	if df.seenURLs == nil {
		df.seenURLs = make(map[string]bool)
	}

	if df.seenURLs[result.URL] {
		return false
	}

	df.seenURLs[result.URL] = true
	return true
}

func (df *DuplicateFilter) GetName() string {
	return "DuplicateFilter"
}

func (df *DuplicateFilter) GetDescription() string {
	return "Filters out duplicate URLs"
}

// LanguageFilter filters results based on language
type LanguageFilter struct {
	config FilterConfig
}

func (lf *LanguageFilter) Filter(result *WebSearchResult) bool {
	if !lf.config.EnableLanguageFiltering {
		return true
	}

	// Simple language detection based on common English words
	content := strings.ToLower(result.Title + " " + result.Description)
	englishWords := []string{"the", "and", "or", "but", "in", "on", "at", "to", "for", "of", "with", "by", "test", "business", "official", "website"}

	englishCount := 0
	for _, word := range englishWords {
		if strings.Contains(content, word) {
			englishCount++
		}
	}

	// If we find enough English words, consider it English
	return englishCount >= 2
}

func (lf *LanguageFilter) GetName() string {
	return "LanguageFilter"
}

func (lf *LanguageFilter) GetDescription() string {
	return "Filters results based on language detection"
}

// QualityFilter filters results based on overall quality
type QualityFilter struct {
	config FilterConfig
}

func (qf *QualityFilter) Filter(result *WebSearchResult) bool {
	if !qf.config.EnableQualityFiltering {
		return true
	}

	// Check for obvious quality issues
	if result.Title == "" || result.URL == "" {
		return false
	}

	// Check for suspicious patterns
	suspiciousPatterns := []string{
		"http://", // Prefer HTTPS
		"www.",    // Prefer non-www
	}

	for _, pattern := range suspiciousPatterns {
		if strings.HasPrefix(result.URL, pattern) {
			return false
		}
	}

	return true
}

func (qf *QualityFilter) GetName() string {
	return "QualityFilter"
}

func (qf *QualityFilter) GetDescription() string {
	return "Filters results based on overall quality indicators"
}

// RelevanceRanker ranks results based on relevance
type RelevanceRanker struct{}

func (rr *RelevanceRanker) Rank(result *WebSearchResult, query string) float64 {
	return result.RelevanceScore
}

func (rr *RelevanceRanker) GetName() string {
	return "RelevanceRanker"
}

func (rr *RelevanceRanker) GetDescription() string {
	return "Ranks results based on relevance score"
}

// DomainRanker ranks results based on domain quality
type DomainRanker struct {
	config FilterConfig
}

func (dr *DomainRanker) Rank(result *WebSearchResult, query string) float64 {
	score := 0.5 // Base score

	// Boost preferred domains
	if dr.isPreferredDomain(result.URL) {
		score += 0.4
	}

	// Penalize suspicious domains
	if dr.isSuspiciousDomain(result.URL) {
		score -= 0.3
	}

	return score
}

func (dr *DomainRanker) isPreferredDomain(url string) bool {
	for _, domain := range dr.config.PreferredDomains {
		if strings.Contains(url, domain) {
			return true
		}
	}
	return false
}

func (dr *DomainRanker) isSuspiciousDomain(url string) bool {
	suspiciousPatterns := []string{
		".tk", ".ml", ".ga", ".cf", // Free domains
		"bit.ly", "goo.gl", "tinyurl.com", // URL shorteners
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(url, pattern) {
			return true
		}
	}
	return false
}

func (dr *DomainRanker) GetName() string {
	return "DomainRanker"
}

func (dr *DomainRanker) GetDescription() string {
	return "Ranks results based on domain quality"
}

// ContentRanker ranks results based on content quality
type ContentRanker struct{}

func (cr *ContentRanker) Rank(result *WebSearchResult, query string) float64 {
	score := 0.5 // Base score

	// Boost results with good content length
	contentLength := len(result.Title) + len(result.Description)
	if contentLength > 100 && contentLength < 1000 {
		score += 0.2
	}

	// Boost results with structured content
	if strings.Contains(result.Title, " - ") {
		score += 0.1
	}

	// Boost results with descriptive snippets
	if len(result.Description) > 50 {
		score += 0.1
	}

	return score
}

func (cr *ContentRanker) GetName() string {
	return "ContentRanker"
}

func (cr *ContentRanker) GetDescription() string {
	return "Ranks results based on content quality"
}

// FreshnessRanker ranks results based on freshness
type FreshnessRanker struct{}

func (fr *FreshnessRanker) Rank(result *WebSearchResult, query string) float64 {
	// For now, return a base score since we don't have reliable date information
	// In a real implementation, this would analyze the published date
	return 0.5
}

func (fr *FreshnessRanker) GetName() string {
	return "FreshnessRanker"
}

func (fr *FreshnessRanker) GetDescription() string {
	return "Ranks results based on freshness/date"
}

// AuthorityRanker ranks results based on authority
type AuthorityRanker struct{}

func (ar *AuthorityRanker) Rank(result *WebSearchResult, query string) float64 {
	score := 0.5 // Base score

	// Boost authoritative domains
	authoritativeDomains := []string{
		"wikipedia.org", "linkedin.com", "crunchbase.com", "bloomberg.com",
		"reuters.com", "forbes.com", "techcrunch.com", "wsj.com",
		"nytimes.com", "bbc.com", "cnn.com", "npr.org",
	}

	for _, domain := range authoritativeDomains {
		if strings.Contains(result.URL, domain) {
			score += 0.3
			break
		}
	}

	return score
}

func (ar *AuthorityRanker) GetName() string {
	return "AuthorityRanker"
}

func (ar *AuthorityRanker) GetDescription() string {
	return "Ranks results based on domain authority"
}

// PopularityRanker ranks results based on popularity
type PopularityRanker struct{}

func (pr *PopularityRanker) Rank(result *WebSearchResult, query string) float64 {
	// For now, return a base score since we don't have popularity metrics
	// In a real implementation, this would use metrics like page rank, social shares, etc.
	return 0.5
}

func (pr *PopularityRanker) GetName() string {
	return "PopularityRanker"
}

func (pr *PopularityRanker) GetDescription() string {
	return "Ranks results based on popularity metrics"
}
