package classification

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/pcraw4d/business-verification/internal/classification/repository"
)

// IndustryDetectionService provides database-driven industry classification
type IndustryDetectionService struct {
	repo   repository.KeywordRepository
	logger *log.Logger
}

// NewIndustryDetectionService creates a new industry detection service
func NewIndustryDetectionService(repo repository.KeywordRepository, logger *log.Logger) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	return &IndustryDetectionService{
		repo:   repo,
		logger: logger,
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
	s.logger.Printf("🔍 Starting database-driven industry detection for content length: %d", len(content))

	if content == "" {
		return s.getDefaultResult("No content provided for analysis"), nil
	}

	// Extract keywords from content
	keywords := s.extractKeywordsFromContent(content)
	s.logger.Printf("🔍 Extracted %d keywords from content", len(keywords))

	if len(keywords) == 0 {
		return s.getDefaultResult("No meaningful keywords found in content"), nil
	}

	// Classify business using the repository
	result, err := s.repo.ClassifyBusinessByKeywords(ctx, keywords)
	if err != nil {
		s.logger.Printf("⚠️ Repository classification failed: %v, falling back to default", err)
		return s.getDefaultResult("Classification failed, using default"), nil
	}

	// Get classification codes for the detected industry
	var codes []*repository.ClassificationCode
	if result.Industry != nil {
		codes, err = s.repo.GetClassificationCodesByIndustry(ctx, result.Industry.ID)
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
		AnalysisMethod:      "database_keyword_classification",
		Evidence:            evidence,
		ClassificationCodes: codes,
	}

	s.logger.Printf("✅ Industry detected: %s (confidence: %.2f%%)",
		detectionResult.Industry.Name, detectionResult.Confidence*100)

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

	// Get classification codes for the detected industry
	var codes []*repository.ClassificationCode
	if result.Industry != nil {
		codes, err = s.repo.GetClassificationCodesByIndustry(ctx, result.Industry.ID)
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
