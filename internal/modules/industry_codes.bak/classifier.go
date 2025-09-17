package industry_codes

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ClassificationRequest represents a request for industry code classification
type ClassificationRequest struct {
	BusinessName        string     `json:"business_name"`
	BusinessDescription string     `json:"business_description"`
	Website             string     `json:"website,omitempty"`
	Keywords            []string   `json:"keywords,omitempty"`
	PreferredCodeTypes  []CodeType `json:"preferred_code_types,omitempty"`
	MaxResults          int        `json:"max_results,omitempty"`
	MinConfidence       float64    `json:"min_confidence,omitempty"`
}

// ClassificationResult represents the result of industry code classification
type ClassificationResult struct {
	Code       *IndustryCode `json:"code"`
	Confidence float64       `json:"confidence"`
	MatchType  string        `json:"match_type"`
	MatchedOn  []string      `json:"matched_on"`
	Reasons    []string      `json:"reasons"`
	Weight     float64       `json:"weight"`
}

// ClassificationResponse represents the response from classification
type ClassificationResponse struct {
	Request            *ClassificationRequest             `json:"request"`
	Results            []*ClassificationResult            `json:"results"`
	TopResultsByType   map[string][]*ClassificationResult `json:"top_results_by_type"`
	ClassificationTime time.Duration                      `json:"classification_time"`
	TotalCandidates    int                                `json:"total_candidates"`
	Strategy           string                             `json:"strategy"`
	Metadata           map[string]interface{}             `json:"metadata"`
}

// IndustryClassifier provides industry code classification capabilities
type IndustryClassifier struct {
	db               *IndustryCodeDatabase
	lookup           *IndustryCodeLookup
	confidenceFilter *ConfidenceFilter
	resultAggregator *ResultAggregator
	resultValidator  *ResultValidator
	votingEngine     *VotingEngine
	logger           *zap.Logger
}

// NewIndustryClassifier creates a new industry classifier
func NewIndustryClassifier(db *IndustryCodeDatabase, lookup *IndustryCodeLookup, logger *zap.Logger) *IndustryClassifier {
	metadataManager := NewMetadataManager(db.db, logger)
	confidenceScorer := NewConfidenceScorer(db, metadataManager, logger)
	confidenceFilter := NewConfidenceFilter(confidenceScorer, logger)
	rankingEngine := NewRankingEngine(confidenceScorer, logger)
	resultAggregator := NewResultAggregator(confidenceScorer, rankingEngine, logger)
	resultValidator := NewResultValidator(logger)

	// Initialize voting engine with default weighted average configuration
	votingConfig := &VotingConfig{
		Strategy:               VotingStrategyWeightedAverage,
		MinVoters:              2,
		RequiredAgreement:      0.6,
		ConfidenceWeight:       0.4,
		ConsistencyWeight:      0.3,
		DiversityWeight:        0.3,
		EnableTieBreaking:      true,
		EnableOutlierFiltering: true,
		OutlierThreshold:       2.0,
	}
	votingEngine := NewVotingEngine(votingConfig, logger)

	return &IndustryClassifier{
		db:               db,
		lookup:           lookup,
		confidenceFilter: confidenceFilter,
		resultAggregator: resultAggregator,
		resultValidator:  resultValidator,
		votingEngine:     votingEngine,
		logger:           logger,
	}
}

// ClassifyBusiness performs industry code classification for a business
func (ic *IndustryClassifier) ClassifyBusiness(ctx context.Context, req *ClassificationRequest) (*ClassificationResponse, error) {
	startTime := time.Now()

	// Validate request
	if err := ic.validateRequest(req); err != nil {
		return nil, fmt.Errorf("invalid classification request: %w", err)
	}

	// Set defaults
	ic.setRequestDefaults(req)

	ic.logger.Info("Starting business classification",
		zap.String("business_name", req.BusinessName),
		zap.String("strategy", "multi-strategy"))

	// Extract and process text for analysis
	analysisText := ic.prepareAnalysisText(req)

	// Perform voting-based multi-strategy classification
	results, totalCandidates, err := ic.performVotingBasedClassification(ctx, req, analysisText)
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	// Use result aggregator for enhanced filtering, ranking, and presentation
	aggregationRequest := &AggregationRequest{
		Results:           results,
		MaxResultsPerType: 3, // Return top 3 codes by confidence for each code type
		MinConfidence:     req.MinConfidence,
		IncludeMetadata:   true,
		IncludeAnalytics:  true,
		SortBy:            SortByConfidence,
		Presentation:      PresentationAPI,
	}

	aggregatedResults, err := ic.resultAggregator.AggregateAndPresent(ctx, aggregationRequest)
	if err != nil {
		ic.logger.Warn("Failed to use result aggregator, falling back to basic filtering",
			zap.Error(err))
		// Fallback to basic filtering
		filteredResults := ic.filterAndRankResults(results, req.MinConfidence, req.MaxResults)
		topResultsByType := ic.groupResultsByType(filteredResults)

		classificationTime := time.Since(startTime)
		response := &ClassificationResponse{
			Request:            req,
			Results:            filteredResults,
			TopResultsByType:   topResultsByType,
			ClassificationTime: classificationTime,
			TotalCandidates:    totalCandidates,
			Strategy:           "multi-strategy",
			Metadata: map[string]interface{}{
				"analysis_text_length": len(analysisText),
				"keywords_used":        len(req.Keywords),
				"preferred_types":      len(req.PreferredCodeTypes),
			},
		}
		return response, nil
	}

	// Convert aggregated results back to classification results
	filteredResults := make([]*ClassificationResult, len(aggregatedResults.OverallTopResults))
	for i, aggResult := range aggregatedResults.OverallTopResults {
		filteredResults[i] = aggResult.ClassificationResult
	}

	// Convert top results by type
	topResultsByType := make(map[string][]*ClassificationResult)
	for codeType, aggResults := range aggregatedResults.TopThreeByType {
		typeResults := make([]*ClassificationResult, len(aggResults))
		for i, aggResult := range aggResults {
			typeResults[i] = aggResult.ClassificationResult
		}
		topResultsByType[codeType] = typeResults
	}

	classificationTime := time.Since(startTime)

	response := &ClassificationResponse{
		Request:            req,
		Results:            filteredResults,
		TopResultsByType:   topResultsByType,
		ClassificationTime: classificationTime,
		TotalCandidates:    totalCandidates,
		Strategy:           "enhanced-aggregation",
		Metadata: map[string]interface{}{
			"analysis_text_length": len(analysisText),
			"keywords_used":        len(req.Keywords),
			"preferred_types":      len(req.PreferredCodeTypes),
			"aggregation_metadata": aggregatedResults.AggregationMetadata,
			"analytics":            aggregatedResults.Analytics,
			"presentation_data":    aggregatedResults.PresentationData,
		},
	}

	// Validate the classification results
	validationResult, err := ic.resultValidator.ValidateResults(ctx, response)
	if err != nil {
		ic.logger.Warn("Failed to validate classification results",
			zap.Error(err))
		// Continue without validation if it fails
	} else {
		// Add validation metadata to response
		response.Metadata["validation_result"] = validationResult

		// Log validation results
		ic.logger.Info("Classification results validated",
			zap.Bool("is_valid", validationResult.IsValid),
			zap.Float64("overall_score", validationResult.OverallScore),
			zap.Int("issues_count", len(validationResult.Issues)),
			zap.Duration("validation_time", validationResult.ValidationTime))
	}

	ic.logger.Info("Classification completed",
		zap.String("business_name", req.BusinessName),
		zap.Int("results_count", len(filteredResults)),
		zap.Duration("classification_time", classificationTime))

	return response, nil
}

// validateRequest validates the classification request
func (ic *IndustryClassifier) validateRequest(req *ClassificationRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	if req.BusinessName == "" && req.BusinessDescription == "" {
		return fmt.Errorf("either business_name or business_description must be provided")
	}

	if req.MaxResults < 0 {
		return fmt.Errorf("max_results cannot be negative")
	}

	if req.MinConfidence < 0 || req.MinConfidence > 1 {
		return fmt.Errorf("min_confidence must be between 0 and 1")
	}

	return nil
}

// setRequestDefaults sets default values for the request
func (ic *IndustryClassifier) setRequestDefaults(req *ClassificationRequest) {
	if req.MaxResults == 0 {
		req.MaxResults = 10
	}

	if req.MinConfidence == 0 {
		req.MinConfidence = 0.1
	}

	if len(req.PreferredCodeTypes) == 0 {
		req.PreferredCodeTypes = []CodeType{CodeTypeSIC, CodeTypeNAICS, CodeTypeMCC}
	}
}

// prepareAnalysisText combines and cleans text for analysis
func (ic *IndustryClassifier) prepareAnalysisText(req *ClassificationRequest) string {
	var textParts []string

	if req.BusinessName != "" {
		textParts = append(textParts, req.BusinessName)
	}

	if req.BusinessDescription != "" {
		textParts = append(textParts, req.BusinessDescription)
	}

	// Add keywords as additional context
	if len(req.Keywords) > 0 {
		textParts = append(textParts, strings.Join(req.Keywords, " "))
	}

	analysisText := strings.Join(textParts, " ")

	// Clean and normalize text
	analysisText = ic.cleanText(analysisText)

	return analysisText
}

// cleanText cleans and normalizes text for analysis
func (ic *IndustryClassifier) cleanText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)

	// Remove special characters but keep alphanumeric and spaces
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	text = re.ReplaceAllString(text, " ")

	// Remove extra whitespace (after removing special characters)
	re = regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	// Trim whitespace
	text = strings.TrimSpace(text)

	return text
}

// performVotingBasedClassification performs classification using voting-based multi-strategy approach
func (ic *IndustryClassifier) performVotingBasedClassification(ctx context.Context, req *ClassificationRequest, analysisText string) ([]*ClassificationResult, int, error) {
	var strategyVotes []*StrategyVote
	var totalCandidates int

	// Strategy 1: Keyword-based matching
	keywordResults, keywordCandidates, err := ic.classifyByKeywords(ctx, analysisText, req.PreferredCodeTypes)
	if err != nil {
		ic.logger.Warn("Keyword classification failed", zap.Error(err))
	} else {
		confidence := ic.calculateEnhancedStrategyConfidence("keyword_matching", keywordResults)
		vote := &StrategyVote{
			StrategyName: "keyword_matching",
			Results:      keywordResults,
			Weight:       0.7, // Higher weight for keyword matching
			Confidence:   confidence,
			VoteTime:     time.Now(),
			Metadata: map[string]interface{}{
				"candidates_found": keywordCandidates,
				"analysis_text":    analysisText,
			},
		}
		strategyVotes = append(strategyVotes, vote)
		totalCandidates += keywordCandidates
	}

	// Strategy 2: Description similarity matching
	descriptionResults, descriptionCandidates, err := ic.classifyByDescription(ctx, analysisText, req.PreferredCodeTypes)
	if err != nil {
		ic.logger.Warn("Description classification failed", zap.Error(err))
	} else {
		confidence := ic.calculateEnhancedStrategyConfidence("description_similarity", descriptionResults)
		vote := &StrategyVote{
			StrategyName: "description_similarity",
			Results:      descriptionResults,
			Weight:       0.6, // Medium weight for description similarity
			Confidence:   confidence,
			VoteTime:     time.Now(),
			Metadata: map[string]interface{}{
				"candidates_found": descriptionCandidates,
				"analysis_text":    analysisText,
			},
		}
		strategyVotes = append(strategyVotes, vote)
		totalCandidates += descriptionCandidates
	}

	// Strategy 3: Business name pattern matching
	if req.BusinessName != "" {
		nameResults, nameCandidates, err := ic.classifyByBusinessName(ctx, req.BusinessName, req.PreferredCodeTypes)
		if err != nil {
			ic.logger.Warn("Business name classification failed", zap.Error(err))
		} else {
			confidence := ic.calculateEnhancedStrategyConfidence("business_name_patterns", nameResults)
			vote := &StrategyVote{
				StrategyName: "business_name_patterns",
				Results:      nameResults,
				Weight:       0.5, // Lower weight for business name patterns
				Confidence:   confidence,
				VoteTime:     time.Now(),
				Metadata: map[string]interface{}{
					"candidates_found": nameCandidates,
					"business_name":    req.BusinessName,
				},
			}
			strategyVotes = append(strategyVotes, vote)
			totalCandidates += nameCandidates
		}
	}

	// Perform voting if we have at least 2 strategies
	if len(strategyVotes) >= 2 {
		ic.logger.Info("Conducting voting on classification strategies",
			zap.Int("strategy_count", len(strategyVotes)),
			zap.String("voting_strategy", string(ic.votingEngine.config.Strategy)))

		votingResult, err := ic.votingEngine.ConductVoting(ctx, strategyVotes)
		if err != nil {
			ic.logger.Warn("Voting failed, falling back to simple aggregation", zap.Error(err))
			return ic.fallbackToSimpleAggregation(strategyVotes), totalCandidates, nil
		}

		ic.logger.Info("Voting completed successfully",
			zap.Float64("voting_score", votingResult.VotingScore),
			zap.Float64("agreement", votingResult.Agreement),
			zap.Int("final_results", len(votingResult.FinalResults)))

		return votingResult.FinalResults, totalCandidates, nil
	}

	// Fallback to simple aggregation if insufficient strategies
	ic.logger.Info("Insufficient strategies for voting, using simple aggregation",
		zap.Int("strategy_count", len(strategyVotes)))

	return ic.fallbackToSimpleAggregation(strategyVotes), totalCandidates, nil
}

// classifyByKeywords performs keyword-based classification
func (ic *IndustryClassifier) classifyByKeywords(ctx context.Context, text string, preferredTypes []CodeType) ([]*ClassificationResult, int, error) {
	var results []*ClassificationResult

	// Extract meaningful keywords from text
	keywords := ic.extractKeywords(text)

	var totalCandidates int

	for _, keyword := range keywords {
		if len(keyword) < 3 { // Skip very short keywords
			continue
		}

		// Search for codes matching this keyword
		codes, err := ic.db.SearchCodes(ctx, keyword, nil, 50)
		if err != nil {
			continue
		}

		totalCandidates += len(codes)

		for _, code := range codes {
			// Check if this code type is preferred
			if !ic.isPreferredCodeType(code.Type, preferredTypes) {
				continue
			}

			confidence := ic.calculateKeywordConfidence(keyword, code, text)
			if confidence > 0.1 {
				result := &ClassificationResult{
					Code:       code,
					Confidence: confidence,
					MatchType:  "keyword",
					MatchedOn:  []string{keyword},
					Reasons:    []string{fmt.Sprintf("Matched keyword '%s' in %s", keyword, code.Description)},
					Weight:     0.7, // Keyword matching weight
				}
				results = append(results, result)
			}
		}
	}

	return results, totalCandidates, nil
}

// classifyByDescription performs description similarity matching
func (ic *IndustryClassifier) classifyByDescription(ctx context.Context, text string, preferredTypes []CodeType) ([]*ClassificationResult, int, error) {
	var results []*ClassificationResult
	var totalCandidates int

	// Get all codes for preferred types
	for _, codeType := range preferredTypes {
		codes, err := ic.db.GetCodesByType(ctx, codeType, 1000, 0) // Get many codes for comparison
		if err != nil {
			continue
		}

		totalCandidates += len(codes)

		for _, code := range codes {
			similarity := ic.calculateTextSimilarity(text, code.Description)
			if similarity > 0.2 {
				confidence := similarity * 0.8 // Description similarity weight
				result := &ClassificationResult{
					Code:       code,
					Confidence: confidence,
					MatchType:  "description",
					MatchedOn:  []string{"description_similarity"},
					Reasons:    []string{fmt.Sprintf("Text similarity: %.2f with '%s'", similarity, code.Description)},
					Weight:     0.6, // Description matching weight
				}
				results = append(results, result)
			}
		}
	}

	return results, totalCandidates, nil
}

// classifyByBusinessName performs business name pattern matching
func (ic *IndustryClassifier) classifyByBusinessName(ctx context.Context, businessName string, preferredTypes []CodeType) ([]*ClassificationResult, int, error) {
	var results []*ClassificationResult

	// Extract business name indicators
	nameIndicators := ic.extractBusinessNameIndicators(businessName)

	var totalCandidates int

	for _, indicator := range nameIndicators {
		// Search for codes matching this indicator
		codes, err := ic.db.SearchCodes(ctx, indicator, nil, 20)
		if err != nil {
			continue
		}

		totalCandidates += len(codes)

		for _, code := range codes {
			if !ic.isPreferredCodeType(code.Type, preferredTypes) {
				continue
			}

			confidence := ic.calculateNameIndicatorConfidence(indicator, code)
			if confidence > 0.15 {
				result := &ClassificationResult{
					Code:       code,
					Confidence: confidence,
					MatchType:  "business_name",
					MatchedOn:  []string{indicator},
					Reasons:    []string{fmt.Sprintf("Business name indicator '%s' matches %s", indicator, code.Description)},
					Weight:     0.5, // Business name weight
				}
				results = append(results, result)
			}
		}
	}

	return results, totalCandidates, nil
}

// extractKeywords extracts meaningful keywords from text
func (ic *IndustryClassifier) extractKeywords(text string) []string {
	// Common stop words to filter out
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "is": true, "are": true, "was": true, "were": true,
		"be": true, "been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "must": true, "can": true, "we": true, "you": true,
		"they": true, "it": true, "this": true, "that": true, "these": true, "those": true,
		"company": true, "business": true, "inc": true, "corp": true, "llc": true,
	}

	words := strings.Fields(text)
	var keywords []string

	for _, word := range words {
		word = strings.ToLower(strings.TrimSpace(word))
		if len(word) >= 3 && !stopWords[word] {
			keywords = append(keywords, word)
		}
	}

	return ic.deduplicateStringSlice(keywords)
}

// extractBusinessNameIndicators extracts industry indicators from business name
func (ic *IndustryClassifier) extractBusinessNameIndicators(businessName string) []string {
	var indicators []string

	// Common business name patterns that indicate industry
	patterns := map[string][]string{
		"restaurant":   {"restaurant", "cafe", "diner", "bistro", "grill", "eatery"},
		"retail":       {"store", "shop", "market", "outlet", "boutique"},
		"service":      {"services", "solutions", "consulting", "advisors"},
		"tech":         {"tech", "technology", "software", "systems", "digital"},
		"health":       {"medical", "health", "clinic", "pharmacy", "dental"},
		"finance":      {"financial", "bank", "credit", "investment", "insurance"},
		"legal":        {"law", "legal", "attorney", "lawyers"},
		"real estate":  {"realty", "properties", "real estate", "homes"},
		"automotive":   {"auto", "automotive", "cars", "motors"},
		"construction": {"construction", "building", "contractors"},
	}

	nameLower := strings.ToLower(businessName)

	for category, keywords := range patterns {
		for _, keyword := range keywords {
			if strings.Contains(nameLower, keyword) {
				indicators = append(indicators, category)
				indicators = append(indicators, keyword)
			}
		}
	}

	return ic.deduplicateStringSlice(indicators)
}

// calculateKeywordConfidence calculates confidence for keyword matches
func (ic *IndustryClassifier) calculateKeywordConfidence(keyword string, code *IndustryCode, fullText string) float64 {
	confidence := 0.0

	// Base confidence based on keyword match
	if ic.containsWord(code.Description, keyword) {
		confidence += 0.3
	}

	// Check keyword match in code keywords
	for _, codeKeyword := range code.Keywords {
		if strings.Contains(strings.ToLower(codeKeyword), keyword) {
			confidence += 0.4
			break
		}
	}

	// Check category match
	if ic.containsWord(code.Category, keyword) {
		confidence += 0.2
	}

	// Boost confidence based on keyword frequency in text
	frequency := float64(strings.Count(strings.ToLower(fullText), keyword))
	if frequency > 1 {
		confidence += math.Min(frequency*0.1, 0.3)
	}

	// Factor in code confidence
	confidence *= code.Confidence

	return math.Min(confidence, 1.0)
}

// calculateTextSimilarity calculates similarity between two texts
func (ic *IndustryClassifier) calculateTextSimilarity(text1, text2 string) float64 {
	words1 := ic.extractKeywords(text1)
	words2 := ic.extractKeywords(text2)

	if len(words1) == 0 || len(words2) == 0 {
		return 0.0
	}

	// Simple word overlap similarity
	overlap := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 {
				overlap++
				break
			}
		}
	}

	// Jaccard similarity
	union := len(words1) + len(words2) - overlap
	if union == 0 {
		return 0.0
	}

	return float64(overlap) / float64(union)
}

// calculateNameIndicatorConfidence calculates confidence for business name indicators
func (ic *IndustryClassifier) calculateNameIndicatorConfidence(indicator string, code *IndustryCode) float64 {
	confidence := 0.0

	// Check direct matches in description
	if ic.containsWord(code.Description, indicator) {
		confidence += 0.4
	}

	// Check category match
	if ic.containsWord(code.Category, indicator) {
		confidence += 0.3
	}

	// Check keyword matches
	for _, keyword := range code.Keywords {
		if ic.containsWord(keyword, indicator) {
			confidence += 0.2
			break
		}
	}

	// Factor in code confidence
	confidence *= code.Confidence

	return math.Min(confidence, 1.0)
}

// containsWord checks if a text contains a word (case-insensitive, word boundaries)
func (ic *IndustryClassifier) containsWord(text, word string) bool {
	pattern := fmt.Sprintf(`\b%s\b`, regexp.QuoteMeta(strings.ToLower(word)))
	matched, _ := regexp.MatchString(pattern, strings.ToLower(text))
	return matched
}

// isPreferredCodeType checks if a code type is in the preferred list
func (ic *IndustryClassifier) isPreferredCodeType(codeType CodeType, preferredTypes []CodeType) bool {
	for _, preferred := range preferredTypes {
		if codeType == preferred {
			return true
		}
	}
	return false
}

// deduplicateAndMergeResults deduplicates and merges classification results
func (ic *IndustryClassifier) deduplicateAndMergeResults(results []*ClassificationResult) []*ClassificationResult {
	resultMap := make(map[string]*ClassificationResult)

	for _, result := range results {
		key := fmt.Sprintf("%s-%s", result.Code.Code, result.Code.Type)

		if existing, exists := resultMap[key]; exists {
			// Merge results for the same code
			existing.Confidence = math.Max(existing.Confidence, result.Confidence)
			existing.MatchedOn = ic.deduplicateStringSlice(append(existing.MatchedOn, result.MatchedOn...))
			existing.Reasons = ic.deduplicateStringSlice(append(existing.Reasons, result.Reasons...))

			// Update match type to include multiple strategies
			if existing.MatchType != result.MatchType {
				existing.MatchType = "multi-strategy"
			}
		} else {
			resultMap[key] = result
		}
	}

	// Convert map back to slice
	var mergedResults []*ClassificationResult
	for _, result := range resultMap {
		mergedResults = append(mergedResults, result)
	}

	return mergedResults
}

// filterAndRankResults filters results by confidence and ranks them
func (ic *IndustryClassifier) filterAndRankResults(results []*ClassificationResult, minConfidence float64, maxResults int) []*ClassificationResult {
	if len(results) == 0 {
		return results
	}

	// Create a mock request for filtering (we don't have the original request here)
	request := &ClassificationRequest{
		MinConfidence: minConfidence,
		MaxResults:    maxResults,
	}

	// Use the confidence filter for advanced filtering
	filteringResult, filteredResults, err := ic.confidenceFilter.FilterByConfidence(context.Background(), results, request, nil)
	if err != nil {
		ic.logger.Warn("Failed to apply confidence filtering, falling back to basic filtering",
			zap.Error(err))
		// Fallback to basic filtering
		return ic.basicFilterAndRankResults(results, minConfidence, maxResults)
	}

	ic.logger.Info("Applied advanced confidence filtering",
		zap.Int("original_count", filteringResult.OriginalCount),
		zap.Int("filtered_count", filteringResult.FilteredCount),
		zap.Int("rejected_count", filteringResult.RejectedCount),
		zap.Float64("threshold_used", filteringResult.ThresholdUsed))

	// Limit to max results if needed
	if len(filteredResults) > maxResults {
		filteredResults = filteredResults[:maxResults]
	}

	return filteredResults
}

// basicFilterAndRankResults provides basic filtering as fallback
func (ic *IndustryClassifier) basicFilterAndRankResults(results []*ClassificationResult, minConfidence float64, maxResults int) []*ClassificationResult {
	// Filter by minimum confidence
	var filteredResults []*ClassificationResult
	for _, result := range results {
		if result.Confidence >= minConfidence {
			filteredResults = append(filteredResults, result)
		}
	}

	// Sort by confidence (descending)
	sort.Slice(filteredResults, func(i, j int) bool {
		return filteredResults[i].Confidence > filteredResults[j].Confidence
	})

	// Limit to max results
	if len(filteredResults) > maxResults {
		filteredResults = filteredResults[:maxResults]
	}

	return filteredResults
}

// groupResultsByType groups results by code type
func (ic *IndustryClassifier) groupResultsByType(results []*ClassificationResult) map[string][]*ClassificationResult {
	groupedResults := make(map[string][]*ClassificationResult)

	for _, result := range results {
		codeType := string(result.Code.Type)
		groupedResults[codeType] = append(groupedResults[codeType], result)
	}

	// Limit each group to top 3 results
	for codeType, typeResults := range groupedResults {
		if len(typeResults) > 3 {
			groupedResults[codeType] = typeResults[:3]
		}
	}

	return groupedResults
}

// deduplicateStringSlice removes duplicates from a string slice
func (ic *IndustryClassifier) deduplicateStringSlice(slice []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}

// calculateEnhancedStrategyConfidence calculates advanced confidence using the confidence calculator
func (ic *IndustryClassifier) calculateEnhancedStrategyConfidence(strategyName string, results []*ClassificationResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	// Use enhanced confidence calculation if voting engine has confidence calculator
	if ic.votingEngine != nil && ic.votingEngine.confidenceCalculator != nil {
		ctx := context.Background()

		metrics, err := ic.votingEngine.confidenceCalculator.CalculateAdvancedStrategyConfidence(ctx, strategyName, results)
		if err == nil && metrics != nil {
			return metrics.FinalConfidence
		}
		// Fall back to simple calculation if enhanced fails
		ic.logger.Warn("Enhanced confidence calculation failed, using simple method",
			zap.String("strategy", strategyName),
			zap.Error(err))
	}

	// Fall back to simple calculation
	return ic.calculateStrategyConfidence(results)
}

// calculateStrategyConfidence calculates the overall confidence of a strategy based on its results (simple version)
func (ic *IndustryClassifier) calculateStrategyConfidence(results []*ClassificationResult) float64 {
	if len(results) == 0 {
		return 0.0
	}

	// Simple confidence calculation
	totalConfidence := 0.0
	for _, result := range results {
		totalConfidence += result.Confidence
	}

	averageConfidence := totalConfidence / float64(len(results))

	// Boost confidence slightly for strategies with more results (up to a point)
	resultCountFactor := math.Min(1.0, float64(len(results))/10.0) * 0.1

	return math.Min(1.0, averageConfidence+resultCountFactor)
}

// fallbackToSimpleAggregation provides simple result aggregation when voting fails
func (ic *IndustryClassifier) fallbackToSimpleAggregation(strategyVotes []*StrategyVote) []*ClassificationResult {
	if len(strategyVotes) == 0 {
		return []*ClassificationResult{}
	}

	// Collect all results from all strategies
	var allResults []*ClassificationResult
	for _, vote := range strategyVotes {
		for _, result := range vote.Results {
			// Apply strategy weight to confidence
			weightedResult := &ClassificationResult{
				Code:       result.Code,
				Confidence: result.Confidence * vote.Weight,
				MatchType:  result.MatchType + "_" + vote.StrategyName,
				MatchedOn:  result.MatchedOn,
				Reasons:    append(result.Reasons, fmt.Sprintf("Strategy: %s (weight: %.2f)", vote.StrategyName, vote.Weight)),
				Weight:     vote.Weight,
			}
			allResults = append(allResults, weightedResult)
		}
	}

	// Deduplicate and merge results using existing logic
	return ic.deduplicateAndMergeResults(allResults)
}
