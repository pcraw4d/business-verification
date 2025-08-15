package webanalysis

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// SearchResultAnalyzer analyzes search results for industry classification
type SearchResultAnalyzer struct {
	industryExtractor *SearchIndustryExtractor
	qualityAssessor   *SearchQualityAssessor
	confidenceScorer  *SearchConfidenceScorer
	deduplicator      *SearchResultDeduplicator
	ranker            *SearchResultRanker
	config            SearchAnalyzerConfig
}

// SearchAnalyzerConfig holds configuration for search result analysis
type SearchAnalyzerConfig struct {
	MinSnippetLength    int                 `json:"min_snippet_length"`
	MaxSnippetLength    int                 `json:"max_snippet_length"`
	MinQualityScore     float64             `json:"min_quality_score"`
	MinConfidenceScore  float64             `json:"min_confidence_score"`
	EnableDeduplication bool                `json:"enable_deduplication"`
	EnableRanking       bool                `json:"enable_ranking"`
	IndustryKeywords    map[string][]string `json:"industry_keywords"`
	QualityWeights      map[string]float64  `json:"quality_weights"`
}

// SearchAnalysisResult represents the result of search result analysis
type SearchAnalysisResult struct {
	Results            []*AnalyzedSearchResult `json:"results"`
	IndustryIndicators map[string]float64      `json:"industry_indicators"`
	OverallQuality     float64                 `json:"overall_quality"`
	OverallConfidence  float64                 `json:"overall_confidence"`
	AnalysisTime       time.Time               `json:"analysis_time"`
	AnalysisMetadata   map[string]interface{}  `json:"analysis_metadata"`
}

// AnalyzedSearchResult represents an analyzed search result
type AnalyzedSearchResult struct {
	OriginalResult     *MultiSourceSearchResult `json:"original_result"`
	IndustryIndicators []string                 `json:"industry_indicators"`
	QualityScore       float64                  `json:"quality_score"`
	ConfidenceScore    float64                  `json:"confidence_score"`
	RelevanceScore     float64                  `json:"relevance_score"`
	RankingScore       float64                  `json:"ranking_score"`
	AnalysisMetadata   map[string]interface{}   `json:"analysis_metadata"`
}

// NewSearchResultAnalyzer creates a new search result analyzer
func NewSearchResultAnalyzer(config SearchAnalyzerConfig) *SearchResultAnalyzer {
	return &SearchResultAnalyzer{
		industryExtractor: NewSearchIndustryExtractor(config.IndustryKeywords),
		qualityAssessor:   NewSearchQualityAssessor(config.QualityWeights),
		confidenceScorer:  NewSearchConfidenceScorer(),
		deduplicator:      NewSearchResultDeduplicator(),
		ranker:            NewSearchResultRanker(),
		config:            config,
	}
}

// AnalyzeSearchResults performs comprehensive analysis of search results
func (sra *SearchResultAnalyzer) AnalyzeSearchResults(ctx context.Context, results []*MultiSourceSearchResult, business string) (*SearchAnalysisResult, error) {
	startTime := time.Now()

	// Step 1: Extract industry indicators
	analyzedResults := sra.extractIndustryIndicators(results, business)

	// Step 2: Assess quality
	analyzedResults = sra.assessQuality(analyzedResults)

	// Step 3: Calculate confidence scores
	analyzedResults = sra.calculateConfidenceScores(analyzedResults, business)

	// Step 4: Deduplicate results
	if sra.config.EnableDeduplication {
		analyzedResults = sra.deduplicator.DeduplicateResults(analyzedResults)
	}

	// Step 5: Rank results
	if sra.config.EnableRanking {
		analyzedResults = sra.ranker.RankResults(analyzedResults)
	}

	// Step 6: Calculate overall metrics
	overallQuality := sra.calculateOverallQuality(analyzedResults)
	overallConfidence := sra.calculateOverallConfidence(analyzedResults)
	industryIndicators := sra.aggregateIndustryIndicators(analyzedResults)

	// Create analysis metadata
	metadata := map[string]interface{}{
		"business":           business,
		"total_results":      len(results),
		"analyzed_results":   len(analyzedResults),
		"analysis_duration":  time.Since(startTime).String(),
		"deduplication_used": sra.config.EnableDeduplication,
		"ranking_used":       sra.config.EnableRanking,
	}

	result := &SearchAnalysisResult{
		Results:            analyzedResults,
		IndustryIndicators: industryIndicators,
		OverallQuality:     overallQuality,
		OverallConfidence:  overallConfidence,
		AnalysisTime:       time.Now(),
		AnalysisMetadata:   metadata,
	}

	return result, nil
}

// extractIndustryIndicators extracts industry indicators from search results
func (sra *SearchResultAnalyzer) extractIndustryIndicators(results []*MultiSourceSearchResult, business string) []*AnalyzedSearchResult {
	var analyzedResults []*AnalyzedSearchResult

	for _, result := range results {
		analyzedResult := &AnalyzedSearchResult{
			OriginalResult:     result,
			IndustryIndicators: []string{},
			AnalysisMetadata:   make(map[string]interface{}),
		}

		// Extract industry indicators from title and snippet
		titleIndicators := sra.industryExtractor.ExtractFromText(result.Title)
		snippetIndicators := sra.industryExtractor.ExtractFromText(result.Snippet)

		// Combine indicators
		analyzedResult.IndustryIndicators = append(analyzedResult.IndustryIndicators, titleIndicators...)
		analyzedResult.IndustryIndicators = append(analyzedResult.IndustryIndicators, snippetIndicators...)

		// Remove duplicates
		analyzedResult.IndustryIndicators = sra.removeDuplicateIndicators(analyzedResult.IndustryIndicators)

		analyzedResults = append(analyzedResults, analyzedResult)
	}

	return analyzedResults
}

// assessQuality assesses the quality of search results
func (sra *SearchResultAnalyzer) assessQuality(results []*AnalyzedSearchResult) []*AnalyzedSearchResult {
	for _, result := range results {
		result.QualityScore = sra.qualityAssessor.AssessQuality(result.OriginalResult)
	}
	return results
}

// calculateConfidenceScores calculates confidence scores for search results
func (sra *SearchResultAnalyzer) calculateConfidenceScores(results []*AnalyzedSearchResult, business string) []*AnalyzedSearchResult {
	for _, result := range results {
		result.ConfidenceScore = sra.confidenceScorer.CalculateConfidence(result, business)
	}
	return results
}

// calculateOverallQuality calculates overall quality score
func (sra *SearchResultAnalyzer) calculateOverallQuality(results []*AnalyzedSearchResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalQuality := 0.0
	for _, result := range results {
		totalQuality += result.QualityScore
	}

	return totalQuality / float64(len(results))
}

// calculateOverallConfidence calculates overall confidence score
func (sra *SearchResultAnalyzer) calculateOverallConfidence(results []*AnalyzedSearchResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, result := range results {
		totalConfidence += result.ConfidenceScore
	}

	return totalConfidence / float64(len(results))
}

// aggregateIndustryIndicators aggregates industry indicators across all results
func (sra *SearchResultAnalyzer) aggregateIndustryIndicators(results []*AnalyzedSearchResult) map[string]float64 {
	industryCounts := make(map[string]int)
	totalResults := len(results)

	for _, result := range results {
		for _, indicator := range result.IndustryIndicators {
			industryCounts[indicator]++
		}
	}

	industryIndicators := make(map[string]float64)
	for industry, count := range industryCounts {
		industryIndicators[industry] = float64(count) / float64(totalResults)
	}

	return industryIndicators
}

// removeDuplicateIndicators removes duplicate industry indicators
func (sra *SearchResultAnalyzer) removeDuplicateIndicators(indicators []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, indicator := range indicators {
		if !seen[indicator] {
			seen[indicator] = true
			unique = append(unique, indicator)
		}
	}

	return unique
}

// SearchIndustryExtractor extracts industry indicators from search results
type SearchIndustryExtractor struct {
	industryKeywords map[string][]string
	keywordPatterns  map[string]*regexp.Regexp
}

// NewSearchIndustryExtractor creates a new search industry extractor
func NewSearchIndustryExtractor(industryKeywords map[string][]string) *SearchIndustryExtractor {
	sie := &SearchIndustryExtractor{
		industryKeywords: industryKeywords,
		keywordPatterns:  make(map[string]*regexp.Regexp),
	}
	sie.initializePatterns()
	return sie
}

// initializePatterns initializes regex patterns for industry keywords
func (sie *SearchIndustryExtractor) initializePatterns() {
	for industry, keywords := range sie.industryKeywords {
		for _, keyword := range keywords {
			pattern := regexp.MustCompile(fmt.Sprintf(`(?i)\b%s\b`, regexp.QuoteMeta(keyword)))
			sie.keywordPatterns[keyword] = pattern
		}
	}
}

// ExtractFromText extracts industry indicators from text
func (sie *SearchIndustryExtractor) ExtractFromText(text string) []string {
	var indicators []string

	for industry, keywords := range sie.industryKeywords {
		for _, keyword := range keywords {
			if pattern, exists := sie.keywordPatterns[keyword]; exists {
				if pattern.MatchString(text) {
					indicators = append(indicators, industry)
					break // Found one keyword for this industry, move to next
				}
			}
		}
	}

	return indicators
}

// SearchQualityAssessor assesses the quality of search results
type SearchQualityAssessor struct {
	qualityWeights map[string]float64
}

// NewSearchQualityAssessor creates a new search quality assessor
func NewSearchQualityAssessor(qualityWeights map[string]float64) *SearchQualityAssessor {
	if qualityWeights == nil {
		qualityWeights = map[string]float64{
			"title_length":     0.2,
			"snippet_length":   0.3,
			"url_quality":      0.2,
			"provider_quality": 0.15,
			"relevance":        0.15,
		}
	}

	return &SearchQualityAssessor{
		qualityWeights: qualityWeights,
	}
}

// AssessQuality assesses the quality of a search result
func (sqa *SearchQualityAssessor) AssessQuality(result *MultiSourceSearchResult) float64 {
	score := 0.0

	// Title length quality
	titleLength := len(result.Title)
	if titleLength >= 10 && titleLength <= 100 {
		score += sqa.qualityWeights["title_length"]
	} else if titleLength > 5 && titleLength < 200 {
		score += sqa.qualityWeights["title_length"] * 0.5
	}

	// Snippet length quality
	snippetLength := len(result.Snippet)
	if snippetLength >= 50 && snippetLength <= 300 {
		score += sqa.qualityWeights["snippet_length"]
	} else if snippetLength > 20 && snippetLength < 500 {
		score += sqa.qualityWeights["snippet_length"] * 0.5
	}

	// URL quality
	if strings.HasPrefix(result.URL, "https://") {
		score += sqa.qualityWeights["url_quality"]
	} else if strings.HasPrefix(result.URL, "http://") {
		score += sqa.qualityWeights["url_quality"] * 0.5
	}

	// Provider quality
	if result.Provider == "google" {
		score += sqa.qualityWeights["provider_quality"]
	} else if result.Provider == "bing" {
		score += sqa.qualityWeights["provider_quality"] * 0.8
	}

	// Relevance quality
	score += result.RelevanceScore * sqa.qualityWeights["relevance"]

	return score
}

// SearchConfidenceScorer calculates confidence scores for search results
type SearchConfidenceScorer struct {
	baseConfidence float64
}

// NewSearchConfidenceScorer creates a new search confidence scorer
func NewSearchConfidenceScorer() *SearchConfidenceScorer {
	return &SearchConfidenceScorer{
		baseConfidence: 0.75, // Base confidence for search-based analysis (0.75-0.85 range)
	}
}

// CalculateConfidence calculates confidence score for a search result
func (scs *SearchConfidenceScorer) CalculateConfidence(result *AnalyzedSearchResult, business string) float64 {
	confidence := scs.baseConfidence

	// Boost confidence based on industry indicators
	if len(result.IndustryIndicators) > 0 {
		confidence += 0.05
	}

	// Boost confidence based on quality score
	confidence += result.QualityScore * 0.1

	// Boost confidence based on business name match
	businessLower := strings.ToLower(business)
	titleLower := strings.ToLower(result.OriginalResult.Title)
	snippetLower := strings.ToLower(result.OriginalResult.Snippet)

	if strings.Contains(titleLower, businessLower) {
		confidence += 0.05
	}
	if strings.Contains(snippetLower, businessLower) {
		confidence += 0.03
	}

	// Cap confidence at 0.85 (upper bound for search-based analysis)
	if confidence > 0.85 {
		confidence = 0.85
	}

	return confidence
}

// SearchResultDeduplicator removes duplicate search results
type SearchResultDeduplicator struct {
	similarityThreshold float64
}

// NewSearchResultDeduplicator creates a new search result deduplicator
func NewSearchResultDeduplicator() *SearchResultDeduplicator {
	return &SearchResultDeduplicator{
		similarityThreshold: 0.8,
	}
}

// DeduplicateResults removes duplicate search results
func (srd *SearchResultDeduplicator) DeduplicateResults(results []*AnalyzedSearchResult) []*AnalyzedSearchResult {
	var uniqueResults []*AnalyzedSearchResult

	for _, result := range results {
		isDuplicate := false

		for _, uniqueResult := range uniqueResults {
			if srd.isDuplicate(result, uniqueResult) {
				isDuplicate = true
				break
			}
		}

		if !isDuplicate {
			uniqueResults = append(uniqueResults, result)
		}
	}

	return uniqueResults
}

// isDuplicate checks if two search results are duplicates
func (srd *SearchResultDeduplicator) isDuplicate(result1, result2 *AnalyzedSearchResult) bool {
	// Check URL similarity
	if result1.OriginalResult.URL == result2.OriginalResult.URL {
		return true
	}

	// Check title similarity
	titleSimilarity := srd.calculateSimilarity(
		result1.OriginalResult.Title,
		result2.OriginalResult.Title,
	)
	if titleSimilarity > srd.similarityThreshold {
		return true
	}

	// Check snippet similarity
	snippetSimilarity := srd.calculateSimilarity(
		result1.OriginalResult.Snippet,
		result2.OriginalResult.Snippet,
	)
	if snippetSimilarity > srd.similarityThreshold {
		return true
	}

	return false
}

// calculateSimilarity calculates similarity between two strings
func (srd *SearchResultDeduplicator) calculateSimilarity(str1, str2 string) float64 {
	if str1 == str2 {
		return 1.0
	}

	// Simple Jaccard similarity
	words1 := strings.Fields(strings.ToLower(str1))
	words2 := strings.Fields(strings.ToLower(str2))

	// Create word sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range words1 {
		set1[word] = true
	}
	for _, word := range words2 {
		set2[word] = true
	}

	// Calculate intersection and union
	intersection := 0
	union := len(set1)

	for word := range set2 {
		if set1[word] {
			intersection++
		} else {
			union++
		}
	}

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// SearchResultRanker ranks search results by relevance and quality
type SearchResultRanker struct {
	rankingWeights map[string]float64
}

// NewSearchResultRanker creates a new search result ranker
func NewSearchResultRanker() *SearchResultRanker {
	return &SearchResultRanker{
		rankingWeights: map[string]float64{
			"relevance":  0.4,
			"quality":    0.3,
			"confidence": 0.2,
			"industry":   0.1,
		},
	}
}

// RankResults ranks search results by relevance and quality
func (srr *SearchResultRanker) RankResults(results []*AnalyzedSearchResult) []*AnalyzedSearchResult {
	// Calculate ranking scores
	for _, result := range results {
		result.RankingScore = srr.calculateRankingScore(result)
	}

	// Sort results by ranking score (descending)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].RankingScore < results[j].RankingScore {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// calculateRankingScore calculates ranking score for a search result
func (srr *SearchResultRanker) calculateRankingScore(result *AnalyzedSearchResult) float64 {
	score := 0.0

	// Relevance score
	score += result.OriginalResult.RelevanceScore * srr.rankingWeights["relevance"]

	// Quality score
	score += result.QualityScore * srr.rankingWeights["quality"]

	// Confidence score
	score += result.ConfidenceScore * srr.rankingWeights["confidence"]

	// Industry indicator score
	industryScore := 0.0
	if len(result.IndustryIndicators) > 0 {
		industryScore = float64(len(result.IndustryIndicators)) / 10.0 // Normalize to 0-1
		if industryScore > 1.0 {
			industryScore = 1.0
		}
	}
	score += industryScore * srr.rankingWeights["industry"]

	return score
}
