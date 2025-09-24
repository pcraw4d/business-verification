package risk

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RiskKeywordMatcher provides keyword matching capabilities for risk detection
type RiskKeywordMatcher struct {
	logger *zap.Logger
	config *RiskDetectionConfig

	// Compiled regex patterns for performance
	compiledPatterns map[string]*regexp.Regexp
	patternMutex     sync.RWMutex
}

// NewRiskKeywordMatcher creates a new risk keyword matcher
func NewRiskKeywordMatcher(logger *zap.Logger, config *RiskDetectionConfig) *RiskKeywordMatcher {
	return &RiskKeywordMatcher{
		logger:           logger,
		config:           config,
		compiledPatterns: make(map[string]*regexp.Regexp),
	}
}

// MatchKeywords matches risk keywords against content
func (rkm *RiskKeywordMatcher) MatchKeywords(
	content string,
	riskKeywords []RiskKeyword,
	source string,
) []DetectedRiskKeyword {
	var detectedKeywords []DetectedRiskKeyword

	// Normalize content for matching
	normalizedContent := rkm.normalizeContent(content)

	// Match each risk keyword
	for _, keyword := range riskKeywords {
		matches := rkm.matchKeyword(normalizedContent, keyword, source)
		detectedKeywords = append(detectedKeywords, matches...)
	}

	// Remove duplicates and sort by confidence
	detectedKeywords = rkm.deduplicateAndSort(detectedKeywords)

	return detectedKeywords
}

// matchKeyword matches a single keyword against content
func (rkm *RiskKeywordMatcher) matchKeyword(
	content string,
	keyword RiskKeyword,
	source string,
) []DetectedRiskKeyword {
	var matches []DetectedRiskKeyword

	// 1. Direct keyword matching
	directMatches := rkm.matchDirectKeyword(content, keyword, source)
	matches = append(matches, directMatches...)

	// 2. Synonym matching
	if rkm.config.EnableSynonymMatching {
		synonymMatches := rkm.matchSynonyms(content, keyword, source)
		matches = append(matches, synonymMatches...)
	}

	// 3. Pattern matching
	if rkm.config.EnableRegexPatterns {
		patternMatches := rkm.matchPatterns(content, keyword, source)
		matches = append(matches, patternMatches...)
	}

	return matches
}

// matchDirectKeyword performs direct keyword matching
func (rkm *RiskKeywordMatcher) matchDirectKeyword(
	content string,
	keyword RiskKeyword,
	source string,
) []DetectedRiskKeyword {
	var matches []DetectedRiskKeyword

	// Case-insensitive matching
	keywordLower := strings.ToLower(keyword.Keyword)
	contentLower := strings.ToLower(content)

	// Find all occurrences
	positions := rkm.findAllOccurrences(contentLower, keywordLower)

	for _, pos := range positions {
		// Extract context around the match
		context := rkm.extractContext(content, pos, len(keyword.Keyword))

		// Calculate confidence based on context
		confidence := rkm.calculateDirectMatchConfidence(keyword, context)

		if confidence >= rkm.config.MinConfidenceThreshold {
			detected := DetectedRiskKeyword{
				Keyword:          keyword.Keyword,
				Category:         keyword.RiskCategory,
				Severity:         keyword.RiskSeverity,
				Confidence:       confidence,
				Context:          context,
				Source:           source,
				Position:         pos,
				MCCCodes:         keyword.MCCCodes,
				CardRestrictions: keyword.CardBrandRestrictions,
				DetectedAt:       time.Now(),
			}
			matches = append(matches, detected)
		}
	}

	return matches
}

// matchSynonyms performs synonym matching
func (rkm *RiskKeywordMatcher) matchSynonyms(
	content string,
	keyword RiskKeyword,
	source string,
) []DetectedRiskKeyword {
	var matches []DetectedRiskKeyword

	contentLower := strings.ToLower(content)

	for _, synonym := range keyword.Synonyms {
		synonymLower := strings.ToLower(synonym)
		positions := rkm.findAllOccurrences(contentLower, synonymLower)

		for _, pos := range positions {
			context := rkm.extractContext(content, pos, len(synonym))
			confidence := rkm.calculateSynonymMatchConfidence(keyword, synonym, context)

			if confidence >= rkm.config.MinConfidenceThreshold {
				detected := DetectedRiskKeyword{
					Keyword:          keyword.Keyword,
					Category:         keyword.RiskCategory,
					Severity:         keyword.RiskSeverity,
					Confidence:       confidence,
					Context:          context,
					Source:           source,
					Position:         pos,
					MCCCodes:         keyword.MCCCodes,
					CardRestrictions: keyword.CardBrandRestrictions,
					DetectedAt:       time.Now(),
				}
				matches = append(matches, detected)
			}
		}
	}

	return matches
}

// matchPatterns performs regex pattern matching
func (rkm *RiskKeywordMatcher) matchPatterns(
	content string,
	keyword RiskKeyword,
	source string,
) []DetectedRiskKeyword {
	var matches []DetectedRiskKeyword

	for _, pattern := range keyword.DetectionPatterns {
		compiledPattern := rkm.getCompiledPattern(pattern)
		if compiledPattern == nil {
			continue
		}

		// Find all matches
		patternMatches := compiledPattern.FindAllStringIndex(content, -1)

		for _, match := range patternMatches {
			start, end := match[0], match[1]
			matchedText := content[start:end]
			context := rkm.extractContext(content, start, end-start)
			confidence := rkm.calculatePatternMatchConfidence(keyword, pattern, matchedText, context)

			if confidence >= rkm.config.MinConfidenceThreshold {
				detected := DetectedRiskKeyword{
					Keyword:          keyword.Keyword,
					Category:         keyword.RiskCategory,
					Severity:         keyword.RiskSeverity,
					Confidence:       confidence,
					Context:          context,
					Source:           source,
					Position:         start,
					MCCCodes:         keyword.MCCCodes,
					CardRestrictions: keyword.CardBrandRestrictions,
					DetectedAt:       time.Now(),
				}
				matches = append(matches, detected)
			}
		}
	}

	return matches
}

// getCompiledPattern gets or compiles a regex pattern
func (rkm *RiskKeywordMatcher) getCompiledPattern(pattern string) *regexp.Regexp {
	rkm.patternMutex.RLock()
	if compiled, exists := rkm.compiledPatterns[pattern]; exists {
		rkm.patternMutex.RUnlock()
		return compiled
	}
	rkm.patternMutex.RUnlock()

	// Compile pattern
	compiled, err := regexp.Compile(`(?i)` + pattern) // Case-insensitive
	if err != nil {
		rkm.logger.Warn("Failed to compile regex pattern",
			zap.String("pattern", pattern),
			zap.Error(err))
		return nil
	}

	// Cache compiled pattern
	rkm.patternMutex.Lock()
	rkm.compiledPatterns[pattern] = compiled
	rkm.patternMutex.Unlock()

	return compiled
}

// normalizeContent normalizes content for better matching
func (rkm *RiskKeywordMatcher) normalizeContent(content string) string {
	// Remove extra whitespace
	content = strings.TrimSpace(content)

	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	content = re.ReplaceAllString(content, " ")

	// Remove special characters that might interfere with matching
	re = regexp.MustCompile(`[^\w\s\-\.]`)
	content = re.ReplaceAllString(content, " ")

	return content
}

// findAllOccurrences finds all occurrences of a substring in content
func (rkm *RiskKeywordMatcher) findAllOccurrences(content, substring string) []int {
	var positions []int
	start := 0

	for {
		pos := strings.Index(content[start:], substring)
		if pos == -1 {
			break
		}

		actualPos := start + pos
		positions = append(positions, actualPos)
		start = actualPos + 1
	}

	return positions
}

// extractContext extracts context around a match
func (rkm *RiskKeywordMatcher) extractContext(content string, position, length int) string {
	contextSize := 50 // Characters before and after

	start := position - contextSize
	if start < 0 {
		start = 0
	}

	end := position + length + contextSize
	if end > len(content) {
		end = len(content)
	}

	context := content[start:end]

	// Add ellipsis if we truncated
	if start > 0 {
		context = "..." + context
	}
	if end < len(content) {
		context = context + "..."
	}

	return context
}

// calculateDirectMatchConfidence calculates confidence for direct keyword matches
func (rkm *RiskKeywordMatcher) calculateDirectMatchConfidence(
	keyword RiskKeyword,
	context string,
) float64 {
	confidence := 0.8 // Base confidence for direct matches

	// Increase confidence for exact word boundaries
	if rkm.isExactWordMatch(context, keyword.Keyword) {
		confidence += 0.2
	}

	// Increase confidence for multiple occurrences
	occurrences := strings.Count(strings.ToLower(context), strings.ToLower(keyword.Keyword))
	if occurrences > 1 {
		confidence += 0.1 * float64(occurrences-1)
	}

	// Adjust based on severity (higher severity = higher confidence threshold)
	switch keyword.RiskSeverity {
	case "critical":
		confidence *= 1.1
	case "high":
		confidence *= 1.05
	case "medium":
		confidence *= 1.0
	case "low":
		confidence *= 0.95
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateSynonymMatchConfidence calculates confidence for synonym matches
func (rkm *RiskKeywordMatcher) calculateSynonymMatchConfidence(
	keyword RiskKeyword,
	synonym string,
	context string,
) float64 {
	confidence := 0.6 // Base confidence for synonym matches (lower than direct)

	// Increase confidence for exact word boundaries
	if rkm.isExactWordMatch(context, synonym) {
		confidence += 0.2
	}

	// Adjust based on keyword severity
	switch keyword.RiskSeverity {
	case "critical":
		confidence *= 1.1
	case "high":
		confidence *= 1.05
	case "medium":
		confidence *= 1.0
	case "low":
		confidence *= 0.95
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculatePatternMatchConfidence calculates confidence for pattern matches
func (rkm *RiskKeywordMatcher) calculatePatternMatchConfidence(
	keyword RiskKeyword,
	pattern string,
	matchedText string,
	context string,
) float64 {
	confidence := 0.7 // Base confidence for pattern matches

	// Increase confidence for longer matches
	if len(matchedText) > len(keyword.Keyword) {
		confidence += 0.1
	}

	// Increase confidence for exact word boundaries
	if rkm.isExactWordMatch(context, matchedText) {
		confidence += 0.2
	}

	// Adjust based on keyword severity
	switch keyword.RiskSeverity {
	case "critical":
		confidence *= 1.1
	case "high":
		confidence *= 1.05
	case "medium":
		confidence *= 1.0
	case "low":
		confidence *= 0.95
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// isExactWordMatch checks if a match is on word boundaries
func (rkm *RiskKeywordMatcher) isExactWordMatch(context, match string) bool {
	// Simple word boundary check
	contextLower := strings.ToLower(context)
	matchLower := strings.ToLower(match)

	// Find the match in context
	pos := strings.Index(contextLower, matchLower)
	if pos == -1 {
		return false
	}

	// Check character before match
	if pos > 0 {
		charBefore := contextLower[pos-1]
		if isWordCharacter(charBefore) {
			return false
		}
	}

	// Check character after match
	if pos+len(match) < len(contextLower) {
		charAfter := contextLower[pos+len(match)]
		if isWordCharacter(charAfter) {
			return false
		}
	}

	return true
}

// isWordCharacter checks if a character is a word character
func isWordCharacter(char byte) bool {
	return (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || char == '_'
}

// deduplicateAndSort removes duplicate matches and sorts by confidence
func (rkm *RiskKeywordMatcher) deduplicateAndSort(matches []DetectedRiskKeyword) []DetectedRiskKeyword {
	// Create a map to track unique matches
	uniqueMatches := make(map[string]DetectedRiskKeyword)

	for _, match := range matches {
		// Create a key based on keyword, position, and source
		key := fmt.Sprintf("%s_%d_%s", match.Keyword, match.Position, match.Source)

		// Keep the match with higher confidence
		if existing, exists := uniqueMatches[key]; !exists || match.Confidence > existing.Confidence {
			uniqueMatches[key] = match
		}
	}

	// Convert back to slice
	var result []DetectedRiskKeyword
	for _, match := range uniqueMatches {
		result = append(result, match)
	}

	// Sort by confidence (highest first)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Confidence < result[j].Confidence {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}
