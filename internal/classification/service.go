package classification

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification/repository"
)

// IndustryDetectionService provides database-driven industry classification
type IndustryDetectionService struct {
	repo    repository.KeywordRepository
	logger  *log.Logger
	monitor *ClassificationAccuracyMonitoring
}

// NewIndustryDetectionService creates a new industry detection service
func NewIndustryDetectionService(repo repository.KeywordRepository, logger *log.Logger) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	return &IndustryDetectionService{
		repo:    repo,
		logger:  logger,
		monitor: nil, // Will be set separately if monitoring is needed
	}
}

// NewIndustryDetectionServiceWithMonitoring creates a new industry detection service with monitoring
func NewIndustryDetectionServiceWithMonitoring(repo repository.KeywordRepository, logger *log.Logger, monitor *ClassificationAccuracyMonitoring) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	return &IndustryDetectionService{
		repo:    repo,
		logger:  logger,
		monitor: monitor,
	}
}

// IndustryDetectionResult represents the result of industry detection
type IndustryDetectionResult struct {
	Industry            *repository.Industry             `json:"industry"`
	Confidence          float64                          `json:"confidence"`
	KeywordsMatched     []string                         `json:"keywords_matched"`
	AnalysisMethod      string                           `json:"analysis_method"`
	Evidence            string                           `json:"evidence"`
	ClassificationCodes []*repository.ClassificationCode `json:"classification_codes"`
}

// DetectIndustryFromContent analyzes website content to detect industry using database keywords
func (s *IndustryDetectionService) DetectIndustryFromContent(ctx context.Context, content string) (*IndustryDetectionResult, error) {
	startTime := time.Now()
	requestID := s.generateRequestID()

	s.logger.Printf("🔍 Starting database-driven industry detection for content length: %d (request: %s)", len(content), requestID)

	if content == "" {
		result := s.getDefaultResult("No content provided for analysis")
		s.recordClassificationMetrics(ctx, requestID, "", "", "", result, time.Since(startTime), "content_analysis", nil)
		return result, nil
	}

	// Extract keywords from content
	keywords := s.extractKeywordsFromContent(content)
	s.logger.Printf("🔍 Extracted %d keywords from content", len(keywords))

	if len(keywords) == 0 {
		result := s.getDefaultResult("No meaningful keywords found in content")
		s.recordClassificationMetrics(ctx, requestID, "", "", "", result, time.Since(startTime), "content_analysis", nil)
		return result, nil
	}

	// Classify business using the repository
	repoResult, err := s.repo.ClassifyBusinessByKeywords(ctx, keywords)
	if err != nil {
		s.logger.Printf("⚠️ Repository classification failed: %v, falling back to default", err)
		result := s.getDefaultResult("Classification failed, using default")
		s.recordClassificationMetrics(ctx, requestID, "", "", "", result, time.Since(startTime), "content_analysis", err)
		return result, nil
	}

	// Get classification codes for the detected industry (using cache)
	var codes []*repository.ClassificationCode
	if repoResult.Industry != nil {
		codes, err = s.repo.GetCachedClassificationCodes(ctx, repoResult.Industry.ID)
		if err != nil {
			s.logger.Printf("⚠️ Failed to get classification codes: %v", err)
			codes = []*repository.ClassificationCode{}
		}
	}

	// Build evidence string
	evidence := s.buildEvidenceString(keywords, repoResult.Keywords, repoResult.Reasoning)

	detectionResult := &IndustryDetectionResult{
		Industry:            repoResult.Industry,
		Confidence:          repoResult.Confidence,
		KeywordsMatched:     keywords,
		AnalysisMethod:      "database_keyword_classification",
		Evidence:            evidence,
		ClassificationCodes: codes,
	}

	// Record performance metrics
	s.recordClassificationMetrics(ctx, requestID, "", "", "", detectionResult, time.Since(startTime), "content_analysis", nil)

	s.logger.Printf("✅ Industry detected: %s (confidence: %.2f%%) (request: %s)",
		detectionResult.Industry.Name, detectionResult.Confidence*100, requestID)

	return detectionResult, nil
}

// DetectIndustryFromBusinessInfo analyzes business information for industry detection
func (s *IndustryDetectionService) DetectIndustryFromBusinessInfo(ctx context.Context, businessName, description, websiteURL string) (*IndustryDetectionResult, error) {
	s.logger.Printf("🔍 Starting business info analysis: %s", businessName)

	// Extract keywords from all sources
	keywords := s.extractKeywordsFromBusinessInfo(businessName, description, websiteURL)
	s.logger.Printf("🔍 Extracted %d keywords from business info", len(keywords))

	if len(keywords) == 0 {
		return s.getDefaultResult("No meaningful keywords found in business information"), nil
	}

	// Classify business using the repository
	result, err := s.repo.ClassifyBusiness(ctx, businessName, description, websiteURL)
	if err != nil {
		s.logger.Printf("⚠️ Repository classification failed: %v, falling back to default", err)
		return s.getDefaultResult("Classification failed, using default"), nil
	}

	// Get classification codes for the detected industry (using cache)
	var codes []*repository.ClassificationCode
	if result.Industry != nil {
		codes, err = s.repo.GetCachedClassificationCodes(ctx, result.Industry.ID)
		if err != nil {
			s.logger.Printf("⚠️ Failed to get classification codes: %v", err)
			codes = []*repository.ClassificationCode{}
		}
	}

	// Build evidence string
	evidence := s.buildEvidenceString(keywords, result.Keywords, result.Reasoning)

	detectionResult := &IndustryDetectionResult{
		Industry:            result.Industry,
		Confidence:          result.Confidence,
		KeywordsMatched:     keywords,
		AnalysisMethod:      "multi_source_classification",
		Evidence:            evidence,
		ClassificationCodes: codes,
	}

	s.logger.Printf("✅ Industry detected: %s (confidence: %.2f%%)",
		detectionResult.Industry.Name, detectionResult.Confidence*100)

	return detectionResult, nil
}

// GetTopIndustriesByKeywords finds the top industries matching given keywords
func (s *IndustryDetectionService) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*repository.Industry, error) {
	s.logger.Printf("🔍 Getting top industries for %d keywords (limit: %d)", len(keywords), limit)

	industries, err := s.repo.GetTopIndustriesByKeywords(ctx, keywords, limit)
	if err != nil {
		s.logger.Printf("⚠️ Failed to get top industries: %v", err)
		return []*repository.Industry{}, err
	}

	s.logger.Printf("✅ Found %d matching industries", len(industries))
	return industries, nil
}

// SearchIndustriesByPattern searches industries using pattern matching
func (s *IndustryDetectionService) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*repository.Industry, error) {
	s.logger.Printf("🔍 Searching industries by pattern: %s", pattern)

	industries, err := s.repo.SearchIndustriesByPattern(ctx, pattern)
	if err != nil {
		s.logger.Printf("⚠️ Failed to search industries by pattern: %v", err)
		return []*repository.Industry{}, err
	}

	s.logger.Printf("✅ Found %d industries matching pattern", len(industries))
	return industries, nil
}

// GetIndustryStatistics gets statistics about industries and keywords
func (s *IndustryDetectionService) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	s.logger.Printf("🔍 Getting industry statistics")

	stats, err := s.repo.GetIndustryStatistics(ctx)
	if err != nil {
		s.logger.Printf("⚠️ Failed to get industry statistics: %v", err)
		return map[string]interface{}{}, err
	}

	return stats, nil
}

// =============================================================================
// Helper Methods
// =============================================================================

// extractKeywordsFromContent extracts meaningful keywords from website content
func (s *IndustryDetectionService) extractKeywordsFromContent(content string) []string {
	if content == "" {
		return []string{}
	}

	// Convert to lowercase and split into words
	words := strings.Fields(strings.ToLower(content))

	// Filter out common words and short words
	var keywords []string
	seen := make(map[string]bool)

	for _, word := range words {
		// Clean the word
		cleanWord := strings.Trim(word, ".,!?;:\"'()[]{}")

		// Skip if too short, already seen, or is a common word
		if len(cleanWord) < 3 || seen[cleanWord] || s.isCommonWord(cleanWord) {
			continue
		}

		seen[cleanWord] = true
		keywords = append(keywords, cleanWord)
	}

	// Limit to top keywords to avoid overwhelming the system
	if len(keywords) > 50 {
		keywords = keywords[:50]
	}

	return keywords
}

// extractKeywordsFromBusinessInfo extracts keywords from business information
func (s *IndustryDetectionService) extractKeywordsFromBusinessInfo(businessName, description, websiteURL string) []string {
	var keywords []string
	seen := make(map[string]bool)

	// Extract from business name
	if businessName != "" {
		nameWords := strings.Fields(strings.ToLower(businessName))
		for _, word := range nameWords {
			cleanWord := strings.Trim(word, ".,!?;:\"'()[]{}")
			if len(cleanWord) >= 3 && !seen[cleanWord] && !s.isCommonWord(cleanWord) {
				seen[cleanWord] = true
				keywords = append(keywords, cleanWord)
			}
		}
	}

	// Extract from description
	if description != "" {
		descWords := strings.Fields(strings.ToLower(description))
		for _, word := range descWords {
			cleanWord := strings.Trim(word, ".,!?;:\"'()[]{}")
			if len(cleanWord) >= 3 && !seen[cleanWord] && !s.isCommonWord(cleanWord) {
				seen[cleanWord] = true
				keywords = append(keywords, cleanWord)
			}
		}
	}

	// Extract from website URL
	if websiteURL != "" {
		// Remove common URL parts
		cleanURL := strings.TrimPrefix(websiteURL, "https://")
		cleanURL = strings.TrimPrefix(cleanURL, "http://")
		cleanURL = strings.TrimPrefix(cleanURL, "www.")

		// Split by dots and extract meaningful parts
		parts := strings.Split(cleanURL, ".")
		if len(parts) > 0 {
			domainWords := strings.Fields(strings.ReplaceAll(parts[0], "-", " "))
			for _, word := range domainWords {
				cleanWord := strings.Trim(word, ".,!?;:\"'()[]{}")
				if len(cleanWord) >= 3 && !seen[cleanWord] && !s.isCommonWord(cleanWord) {
					seen[cleanWord] = true
					keywords = append(keywords, cleanWord)
				}
			}
		}
	}

	return keywords
}

// isCommonWord checks if a word is a common word that should be filtered out
func (s *IndustryDetectionService) isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "or": true, "but": true, "in": true, "on": true, "at": true,
		"to": true, "for": true, "of": true, "with": true, "by": true, "from": true, "up": true,
		"out": true, "about": true, "into": true, "through": true, "during": true, "before": true,
		"after": true, "above": true, "below": true, "between": true, "among": true, "within": true,
		"without": true, "against": true, "toward": true, "towards": true, "upon": true, "across": true,
		"behind": true, "beneath": true, "beside": true, "beyond": true, "inside": true, "outside": true,
		"under": true, "over": true, "around": true, "along": true, "down": true, "off": true,
		"this": true, "that": true, "these": true, "those": true, "is": true, "are": true, "was": true,
		"were": true, "be": true, "been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "can": true, "must": true, "shall": true, "a": true, "an": true,
		"we": true, "using": true,
		"all": true, "any": true, "each": true, "every": true, "few": true, "many": true, "no": true,
		"some": true, "such": true, "what": true, "which": true, "who": true, "whom": true, "whose": true,
		"where": true, "when": true, "why": true, "how": true, "if": true, "else": true,
		"than": true, "as": true, "so": true, "very": true, "just": true, "only": true, "even": true,
		"still": true, "also": true, "too": true, "well": true, "much": true, "more": true, "most": true,
		"less": true, "least": true, "good": true, "better": true, "best": true, "bad": true, "worse": true,
		"worst": true, "big": true, "bigger": true, "biggest": true, "small": true, "smaller": true,
		"smallest": true, "new": true, "newer": true, "newest": true, "old": true, "older": true, "oldest": true,
		"high": true, "higher": true, "highest": true, "low": true, "lower": true, "lowest": true,
		"long": true, "longer": true, "longest": true, "short": true, "shorter": true, "shortest": true,
		"first": true, "second": true, "third": true, "last": true, "next": true, "previous": true,
		"current": true, "recent": true, "early": true, "late": true, "now": true,
		"here": true, "there": true, "everywhere": true, "nowhere": true, "somewhere": true,
		"anywhere": true, "home": true, "away": true, "abroad": true, "overseas": true, "upstairs": true,
		"downstairs": true, "indoors": true, "outdoors": true,
		"left": true, "right": true, "forward": true, "backward": true, "upward": true, "downward": true,
		"north": true, "south": true, "east": true, "west": true, "northeast": true, "northwest": true,
		"southeast": true, "southwest": true, "northern": true, "southern": true, "eastern": true, "western": true,
	}

	return commonWords[word]
}

// buildEvidenceString builds a human-readable evidence string
func (s *IndustryDetectionService) buildEvidenceString(keywords, resultKeywords []string, reasoning string) string {
	if len(keywords) == 0 {
		return "No keywords found for analysis"
	}

	evidence := fmt.Sprintf("Analysis based on %d extracted keywords", len(keywords))

	if len(resultKeywords) > 0 {
		evidence += fmt.Sprintf(", with %d matching industry indicators", len(resultKeywords))
	}

	if reasoning != "" {
		evidence += fmt.Sprintf(". %s", reasoning)
	}

	return evidence
}

// getDefaultResult returns a default industry detection result
func (s *IndustryDetectionService) getDefaultResult(reason string) *IndustryDetectionResult {
	return &IndustryDetectionResult{
		Industry: &repository.Industry{
			ID:   26, // General Business ID from our seeded data
			Name: "General Business",
		},
		Confidence:          0.50,
		KeywordsMatched:     []string{},
		AnalysisMethod:      "default_fallback",
		Evidence:            reason,
		ClassificationCodes: []*repository.ClassificationCode{},
	}
}

// =============================================================================
// Parallel Processing Methods
// =============================================================================

// BusinessClassificationRequest represents a request for business classification
type BusinessClassificationRequest struct {
	ID           string
	BusinessName string
	Description  string
	WebsiteURL   string
}

// BusinessClassificationResult represents the result of a business classification
type BusinessClassificationResult struct {
	RequestID string
	Result    *IndustryDetectionResult
	Error     error
}

// ClassifyMultipleBusinessesInParallel processes multiple business classifications in parallel
func (s *IndustryDetectionService) ClassifyMultipleBusinessesInParallel(ctx context.Context, requests []BusinessClassificationRequest) []BusinessClassificationResult {
	s.logger.Printf("🚀 Starting parallel classification for %d businesses", len(requests))

	if len(requests) == 0 {
		return []BusinessClassificationResult{}
	}

	// Create channels for results and errors
	results := make([]BusinessClassificationResult, len(requests))
	resultChan := make(chan BusinessClassificationResult, len(requests))

	// Create a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Process each business classification in parallel
	for i, request := range requests {
		wg.Add(1)
		go func(index int, req BusinessClassificationRequest) {
			defer wg.Done()

			s.logger.Printf("🔄 Processing business %d: %s", index+1, req.BusinessName)

			var result *IndustryDetectionResult
			var err error

			// Choose classification method based on available data
			if req.WebsiteURL != "" {
				// Use website content analysis (simplified for parallel processing)
				websiteContent := s.extractKeywordsFromBusinessInfo(req.BusinessName, req.Description, req.WebsiteURL)
				result, err = s.DetectIndustryFromContent(ctx, strings.Join(websiteContent, " "))
			} else {
				// Use business information analysis
				result, err = s.DetectIndustryFromBusinessInfo(ctx, req.BusinessName, req.Description, req.WebsiteURL)
			}

			// Send result to channel
			resultChan <- BusinessClassificationResult{
				RequestID: req.ID,
				Result:    result,
				Error:     err,
			}

			s.logger.Printf("✅ Completed business %d: %s", index+1, req.BusinessName)
		}(i, request)
	}

	// Close the result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results from the channel
	for result := range resultChan {
		// Find the index for this result
		for i, req := range requests {
			if req.ID == result.RequestID {
				results[i] = result
				break
			}
		}
	}

	s.logger.Printf("🚀 Parallel classification completed for %d businesses", len(requests))
	return results
}

// ClassifyBusinessWithMultipleMethods processes a single business using multiple classification methods in parallel
func (s *IndustryDetectionService) ClassifyBusinessWithMultipleMethods(ctx context.Context, businessName, description, websiteURL string) (*IndustryDetectionResult, error) {
	s.logger.Printf("🚀 Starting multi-method classification for: %s", businessName)

	// Create channels for results
	resultChan := make(chan *IndustryDetectionResult, 2)
	errorChan := make(chan error, 2)

	// Create a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Method 1: Business Information Analysis
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.logger.Printf("🔄 Method 1: Business information analysis")

		result, err := s.DetectIndustryFromBusinessInfo(ctx, businessName, description, websiteURL)
		if err != nil {
			errorChan <- fmt.Errorf("business info analysis: %w", err)
			return
		}

		resultChan <- result
		s.logger.Printf("✅ Method 1 completed: %s (confidence: %.2f%%)", result.Industry.Name, result.Confidence*100)
	}()

	// Method 2: Website Content Analysis (if website URL is available)
	if websiteURL != "" {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.logger.Printf("🔄 Method 2: Website content analysis")

			// Use business info extraction for website content (simplified)
			websiteContent := s.extractKeywordsFromBusinessInfo(businessName, description, websiteURL)
			result, err := s.DetectIndustryFromContent(ctx, strings.Join(websiteContent, " "))
			if err != nil {
				errorChan <- fmt.Errorf("website content analysis: %w", err)
				return
			}

			resultChan <- result
			s.logger.Printf("✅ Method 2 completed: %s (confidence: %.2f%%)", result.Industry.Name, result.Confidence*100)
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	var results []*IndustryDetectionResult
	for result := range resultChan {
		results = append(results, result)
	}

	// Log any errors
	for err := range errorChan {
		s.logger.Printf("⚠️ Error in multi-method classification: %v", err)
	}

	// Choose the best result based on confidence
	if len(results) == 0 {
		return s.getDefaultResult("All classification methods failed"), nil
	}

	// Find the result with highest confidence
	bestResult := results[0]
	for _, result := range results[1:] {
		if result.Confidence > bestResult.Confidence {
			bestResult = result
		}
	}

	// Update analysis method to reflect multi-method approach
	bestResult.AnalysisMethod = "multi_method_parallel_classification"
	bestResult.Evidence = fmt.Sprintf("Multi-method analysis: %s (confidence: %.2f%%)", bestResult.Industry.Name, bestResult.Confidence*100)

	s.logger.Printf("🚀 Multi-method classification completed: %s (confidence: %.2f%%)",
		bestResult.Industry.Name, bestResult.Confidence*100)

	return bestResult, nil
}

// GetTopIndustriesByKeywordsInParallel finds top industries for multiple keyword sets in parallel
func (s *IndustryDetectionService) GetTopIndustriesByKeywordsInParallel(ctx context.Context, keywordSets [][]string, limit int) []*IndustryDetectionResult {
	s.logger.Printf("🚀 Starting parallel top industries lookup for %d keyword sets", len(keywordSets))

	if len(keywordSets) == 0 {
		return []*IndustryDetectionResult{}
	}

	// Create channels for results
	resultChan := make(chan *IndustryDetectionResult, len(keywordSets))

	// Create a WaitGroup to wait for all goroutines to complete
	var wg sync.WaitGroup

	// Process each keyword set in parallel
	for i, keywords := range keywordSets {
		wg.Add(1)
		go func(index int, keywordSet []string) {
			defer wg.Done()

			s.logger.Printf("🔄 Processing keyword set %d: %v", index+1, keywordSet)

			result, err := s.repo.ClassifyBusinessByKeywords(ctx, keywordSet)
			if err != nil {
				s.logger.Printf("⚠️ Failed to classify keyword set %d: %v", index+1, err)
				// Send default result
				resultChan <- s.getDefaultResult(fmt.Sprintf("Classification failed for keyword set %d", index+1))
				return
			}

			// Convert to IndustryDetectionResult
			// Convert []ClassificationCode to []*ClassificationCode
			var codes []*repository.ClassificationCode
			for i := range result.Codes {
				codes = append(codes, &result.Codes[i])
			}

			detectionResult := &IndustryDetectionResult{
				Industry:            result.Industry,
				Confidence:          result.Confidence,
				KeywordsMatched:     keywordSet,
				AnalysisMethod:      "parallel_keyword_classification",
				Evidence:            result.Reasoning,
				ClassificationCodes: codes,
			}

			resultChan <- detectionResult
			s.logger.Printf("✅ Completed keyword set %d: %s (confidence: %.2f%%)",
				index+1, result.Industry.Name, result.Confidence*100)
		}(i, keywords)
	}

	// Close the result channel when all goroutines are done
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	var results []*IndustryDetectionResult
	for result := range resultChan {
		results = append(results, result)
	}

	// Sort results by confidence (highest first)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Confidence > results[i].Confidence {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	// Limit results if requested
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	s.logger.Printf("🚀 Parallel top industries lookup completed: %d results", len(results))
	return results
}

// =============================================================================
// Performance Monitoring Helper Methods
// =============================================================================

// generateRequestID generates a unique request ID for tracking
func (s *IndustryDetectionService) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// recordClassificationMetrics records classification performance metrics
func (s *IndustryDetectionService) recordClassificationMetrics(
	ctx context.Context,
	requestID string,
	businessName, description, websiteURL string,
	result *IndustryDetectionResult,
	responseTime time.Duration,
	method string,
	err error,
) {
	if s.monitor == nil {
		return // No monitoring configured
	}

	// Prepare metrics data
	metrics := &ClassificationAccuracyMetrics{
		Timestamp:            time.Now(),
		RequestID:            requestID,
		BusinessName:         &businessName,
		BusinessDescription:  &description,
		WebsiteURL:           &websiteURL,
		PredictedIndustry:    result.Industry.Name,
		PredictedConfidence:  result.Confidence,
		ResponseTimeMs:       float64(responseTime.Nanoseconds()) / 1e6, // Convert to milliseconds
		ClassificationMethod: &method,
		KeywordsUsed:         result.KeywordsMatched,
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
		// if err := s.monitor.RecordClassificationMetrics(ctx, metrics); err != nil {
		//     s.logger.Printf("⚠️ Failed to record classification metrics: %v", err)
		// }
	}()
}

// GetPerformanceMetrics returns current performance metrics
func (s *IndustryDetectionService) GetPerformanceMetrics(ctx context.Context) (*ClassificationAccuracyStats, error) {
	if s.monitor == nil {
		return nil, fmt.Errorf("monitoring not configured")
	}

	// Note: This would call the actual monitoring method when implemented
	// return s.monitor.GetClassificationAccuracyStats(ctx, 24*time.Hour)
	return nil, fmt.Errorf("monitoring not fully implemented")
}

// GetPerformanceTrends returns performance trend data
func (s *IndustryDetectionService) GetPerformanceTrends(ctx context.Context, hours int) ([]*ClassificationAccuracyTrend, error) {
	if s.monitor == nil {
		return nil, fmt.Errorf("monitoring not configured")
	}

	// Note: This would call the actual monitoring method when implemented
	// return s.monitor.GetClassificationAccuracyTrends(ctx, time.Duration(hours)*time.Hour)
	return nil, fmt.Errorf("monitoring not fully implemented")
}

// GetPerformanceAlerts returns current performance alerts
func (s *IndustryDetectionService) GetPerformanceAlerts(ctx context.Context) ([]*ClassificationAccuracyAlert, error) {
	if s.monitor == nil {
		return nil, fmt.Errorf("monitoring not configured")
	}

	// Note: This would call the actual monitoring method when implemented
	// return s.monitor.GetClassificationAccuracyAlerts(ctx)
	return nil, fmt.Errorf("monitoring not fully implemented")
}
