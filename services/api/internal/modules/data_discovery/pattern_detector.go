package data_discovery

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// PatternDetector detects patterns in content for automated data discovery
type PatternDetector struct {
	config        *DataDiscoveryConfig
	logger        *zap.Logger
	patterns      []DataPattern
	compiledRegex map[string]*regexp.Regexp
}

// DataPattern represents a pattern for detecting specific types of data
type DataPattern struct {
	PatternID       string   `json:"pattern_id"`
	Name            string   `json:"name"`
	FieldType       string   `json:"field_type"`
	RegexPattern    string   `json:"regex_pattern"`
	ContextClues    []string `json:"context_clues"`
	Priority        int      `json:"priority"`
	ConfidenceBase  float64  `json:"confidence_base"`
	Examples        []string `json:"examples"`
	ValidationRules []string `json:"validation_rules"`
}

// NewPatternDetector creates a new pattern detector
func NewPatternDetector(config *DataDiscoveryConfig, logger *zap.Logger) *PatternDetector {
	detector := &PatternDetector{
		config:        config,
		logger:        logger,
		patterns:      getBuiltInPatterns(),
		compiledRegex: make(map[string]*regexp.Regexp),
	}

	// Pre-compile regex patterns for performance
	detector.compilePatterns()

	return detector
}

// DetectPatterns detects patterns in the provided content
func (pd *PatternDetector) DetectPatterns(ctx context.Context, content *ContentInput) ([]PatternMatch, error) {
	var matches []PatternMatch

	pd.logger.Debug("Starting pattern detection",
		zap.Int("patterns_to_check", len(pd.patterns)),
		zap.Int("content_length", len(content.RawContent)))

	for _, pattern := range pd.patterns {
		select {
		case <-ctx.Done():
			return matches, ctx.Err()
		default:
			// Detect patterns with timeout protection
			patternMatches := pd.detectPattern(content, pattern)
			matches = append(matches, patternMatches...)
		}
	}

	// Sort matches by confidence score (descending)
	pd.sortMatchesByConfidence(matches)

	// Filter matches based on confidence threshold
	filteredMatches := pd.filterMatchesByConfidence(matches)

	pd.logger.Debug("Pattern detection completed",
		zap.Int("total_matches", len(matches)),
		zap.Int("filtered_matches", len(filteredMatches)))

	return filteredMatches, nil
}

// detectPattern detects a specific pattern in content
func (pd *PatternDetector) detectPattern(content *ContentInput, pattern DataPattern) []PatternMatch {
	var matches []PatternMatch

	regex, exists := pd.compiledRegex[pattern.PatternID]
	if !exists {
		pd.logger.Warn("Pattern regex not compiled", zap.String("pattern_id", pattern.PatternID))
		return matches
	}

	// Find regex matches
	regexMatches := regex.FindAllStringSubmatch(content.RawContent, -1)
	regexIndices := regex.FindAllStringIndex(content.RawContent, -1)

	for i, match := range regexMatches {
		if len(match) > 0 {
			matchedText := match[0]
			position := 0
			if i < len(regexIndices) {
				position = regexIndices[i][0]
			}

			// Calculate confidence score
			confidence := pd.calculatePatternConfidence(content, pattern, matchedText, position)

			// Extract context around the match
			context := pd.extractContext(content.RawContent, position, 100)

			patternMatch := PatternMatch{
				PatternID:       pattern.PatternID,
				MatchedText:     matchedText,
				FieldType:       pattern.FieldType,
				ConfidenceScore: confidence,
				Context:         context,
				Position:        position,
				Metadata: map[string]interface{}{
					"pattern_name":     pattern.Name,
					"regex_pattern":    pattern.RegexPattern,
					"context_enhanced": pd.hasContextClues(context, pattern.ContextClues),
				},
			}

			matches = append(matches, patternMatch)
		}
	}

	return matches
}

// calculatePatternConfidence calculates confidence score for a pattern match
func (pd *PatternDetector) calculatePatternConfidence(content *ContentInput, pattern DataPattern, matchedText string, position int) float64 {
	confidence := pattern.ConfidenceBase

	// Boost confidence based on context clues
	context := pd.extractContext(content.RawContent, position, 200)
	if pd.hasContextClues(context, pattern.ContextClues) {
		confidence += 0.2
	}

	// Boost confidence based on match quality
	matchQuality := pd.assessMatchQuality(matchedText, pattern)
	confidence += matchQuality * 0.1

	// Boost confidence if match appears in structured data
	if pd.isInStructuredData(content, matchedText) {
		confidence += 0.15
	}

	// Ensure confidence is within [0, 1] range
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}

// hasContextClues checks if context contains any of the specified clues
func (pd *PatternDetector) hasContextClues(context string, clues []string) bool {
	contextLower := strings.ToLower(context)
	for _, clue := range clues {
		if strings.Contains(contextLower, strings.ToLower(clue)) {
			return true
		}
	}
	return false
}

// assessMatchQuality assesses the quality of a pattern match
func (pd *PatternDetector) assessMatchQuality(matchedText string, pattern DataPattern) float64 {
	quality := 0.5 // Base quality

	// Check against examples if available
	for _, example := range pattern.Examples {
		if strings.EqualFold(matchedText, example) {
			quality = 1.0 // Perfect match with example
			break
		}
		// Partial similarity boost
		if pd.calculateStringSimilarity(matchedText, example) > 0.8 {
			quality = 0.9
		}
	}

	// Additional quality checks based on field type
	switch pattern.FieldType {
	case "email":
		quality = pd.assessEmailQuality(matchedText)
	case "phone":
		quality = pd.assessPhoneQuality(matchedText)
	case "url":
		quality = pd.assessURLQuality(matchedText)
	case "address":
		quality = pd.assessAddressQuality(matchedText)
	}

	return quality
}

// assessEmailQuality assesses the quality of an email match
func (pd *PatternDetector) assessEmailQuality(email string) float64 {
	// Basic format check
	if !strings.Contains(email, "@") || !strings.Contains(email, ".") {
		return 0.1
	}

	// Split email into parts
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return 0.3
	}

	localPart := parts[0]
	domain := parts[1]

	quality := 0.5

	// Check local part quality
	if len(localPart) > 0 && len(localPart) < 65 {
		quality += 0.2
	}

	// Check domain quality
	if strings.Contains(domain, ".") && len(domain) > 3 {
		quality += 0.2
	}

	// Bonus for common business domains
	businessDomains := []string{".com", ".org", ".net", ".gov", ".edu"}
	for _, bd := range businessDomains {
		if strings.HasSuffix(domain, bd) {
			quality += 0.1
			break
		}
	}

	return quality
}

// assessPhoneQuality assesses the quality of a phone number match
func (pd *PatternDetector) assessPhoneQuality(phone string) float64 {
	// Remove common formatting characters
	cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

	quality := 0.3

	// Check length (US/international standards)
	if len(cleaned) >= 10 && len(cleaned) <= 15 {
		quality += 0.4
	}

	// Check for country code indicators
	if strings.HasPrefix(phone, "+") || strings.HasPrefix(phone, "00") {
		quality += 0.2
	}

	// Check for proper formatting patterns
	formatPatterns := []string{
		`\(\d{3}\)\s*\d{3}-\d{4}`,       // (123) 456-7890
		`\d{3}-\d{3}-\d{4}`,             // 123-456-7890
		`\+\d{1,3}\s*\d{3,4}\s*\d{3,4}`, // +1 123 456 7890
	}

	for _, pattern := range formatPatterns {
		if matched, _ := regexp.MatchString(pattern, phone); matched {
			quality += 0.1
			break
		}
	}

	return quality
}

// assessURLQuality assesses the quality of a URL match
func (pd *PatternDetector) assessURLQuality(url string) float64 {
	quality := 0.3

	// Check for proper protocol
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		quality += 0.3
	}

	// Check for domain structure
	if strings.Contains(url, ".") {
		quality += 0.2
	}

	// Bonus for HTTPS
	if strings.HasPrefix(url, "https://") {
		quality += 0.1
	}

	// Check for reasonable length
	if len(url) > 10 && len(url) < 2000 {
		quality += 0.1
	}

	return quality
}

// assessAddressQuality assesses the quality of an address match
func (pd *PatternDetector) assessAddressQuality(address string) float64 {
	quality := 0.3

	// Check for street number
	if regexp.MustCompile(`^\d+`).MatchString(address) {
		quality += 0.2
	}

	// Check for street type indicators
	streetTypes := []string{"st", "street", "ave", "avenue", "rd", "road", "blvd", "boulevard", "ln", "lane", "dr", "drive"}
	addressLower := strings.ToLower(address)
	for _, st := range streetTypes {
		if strings.Contains(addressLower, st) {
			quality += 0.2
			break
		}
	}

	// Check for city/state indicators (commas)
	if strings.Count(address, ",") >= 1 {
		quality += 0.2
	}

	// Check for ZIP code pattern
	if regexp.MustCompile(`\d{5}(-\d{4})?`).MatchString(address) {
		quality += 0.1
	}

	return quality
}

// isInStructuredData checks if the matched text appears in structured data
func (pd *PatternDetector) isInStructuredData(content *ContentInput, matchedText string) bool {
	if content.StructuredData == nil {
		return false
	}

	// Convert structured data to string for searching
	structuredString := fmt.Sprintf("%v", content.StructuredData)
	return strings.Contains(structuredString, matchedText)
}

// extractContext extracts context around a position in the text
func (pd *PatternDetector) extractContext(text string, position, contextLength int) string {
	start := position - contextLength/2
	end := position + contextLength/2

	if start < 0 {
		start = 0
	}
	if end > len(text) {
		end = len(text)
	}

	return text[start:end]
}

// calculateStringSimilarity calculates similarity between two strings
func (pd *PatternDetector) calculateStringSimilarity(s1, s2 string) float64 {
	// Simple implementation using longest common subsequence ratio
	if len(s1) == 0 && len(s2) == 0 {
		return 1.0
	}
	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Calculate Levenshtein distance ratio (simplified)
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	// Count common characters (simplified similarity)
	commonChars := 0
	for i := 0; i < len(s1) && i < len(s2); i++ {
		if s1[i] == s2[i] {
			commonChars++
		}
	}

	return float64(commonChars) / float64(maxLen)
}

// compilePatterns pre-compiles regex patterns for performance
func (pd *PatternDetector) compilePatterns() {
	for _, pattern := range pd.patterns {
		regex, err := regexp.Compile(pattern.RegexPattern)
		if err != nil {
			pd.logger.Warn("Failed to compile pattern regex",
				zap.String("pattern_id", pattern.PatternID),
				zap.String("regex", pattern.RegexPattern),
				zap.Error(err))
			continue
		}
		pd.compiledRegex[pattern.PatternID] = regex
	}

	pd.logger.Info("Compiled regex patterns",
		zap.Int("total_patterns", len(pd.patterns)),
		zap.Int("compiled_patterns", len(pd.compiledRegex)))
}

// sortMatchesByConfidence sorts pattern matches by confidence score in descending order
func (pd *PatternDetector) sortMatchesByConfidence(matches []PatternMatch) {
	for i := 0; i < len(matches)-1; i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[i].ConfidenceScore < matches[j].ConfidenceScore {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}
}

// filterMatchesByConfidence filters matches based on minimum confidence threshold
func (pd *PatternDetector) filterMatchesByConfidence(matches []PatternMatch) []PatternMatch {
	var filtered []PatternMatch
	for _, match := range matches {
		if match.ConfidenceScore >= pd.config.MinConfidenceThreshold {
			filtered = append(filtered, match)
		}
	}
	return filtered
}

// getBuiltInPatterns returns built-in patterns for common data types
func getBuiltInPatterns() []DataPattern {
	return []DataPattern{
		{
			PatternID:      "email_basic",
			Name:           "Email Address",
			FieldType:      "email",
			RegexPattern:   `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
			ContextClues:   []string{"email", "contact", "mail", "@"},
			Priority:       1,
			ConfidenceBase: 0.8,
			Examples:       []string{"info@example.com", "contact@company.org"},
		},
		{
			PatternID:      "phone_us",
			Name:           "US Phone Number",
			FieldType:      "phone",
			RegexPattern:   `(?:\+?1[-.\s]?)?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})`,
			ContextClues:   []string{"phone", "tel", "call", "contact"},
			Priority:       1,
			ConfidenceBase: 0.8,
			Examples:       []string{"(555) 123-4567", "+1-555-123-4567"},
		},
		{
			PatternID:      "phone_international",
			Name:           "International Phone Number",
			FieldType:      "phone",
			RegexPattern:   `\+[1-9]\d{1,14}`,
			ContextClues:   []string{"phone", "tel", "international", "call"},
			Priority:       2,
			ConfidenceBase: 0.7,
			Examples:       []string{"+44 20 7946 0958", "+33 1 42 86 83 26"},
		},
		{
			PatternID:      "url_http",
			Name:           "HTTP URL",
			FieldType:      "url",
			RegexPattern:   `https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`,
			ContextClues:   []string{"website", "url", "link", "http"},
			Priority:       1,
			ConfidenceBase: 0.9,
			Examples:       []string{"https://example.com", "http://company.org"},
		},
		{
			PatternID:      "address_us",
			Name:           "US Street Address",
			FieldType:      "address",
			RegexPattern:   `\d+\s+[A-Za-z\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Lane|Ln|Drive|Dr)[,\s]+[A-Za-z\s]+[,\s]+[A-Z]{2}\s+\d{5}(?:-\d{4})?`,
			ContextClues:   []string{"address", "location", "street", "visit"},
			Priority:       1,
			ConfidenceBase: 0.8,
			Examples:       []string{"123 Main Street, Anytown, NY 12345"},
		},
		{
			PatternID:      "address_flexible",
			Name:           "Flexible US Address",
			FieldType:      "address",
			RegexPattern:   `\d+\s+[A-Za-z\s]+(?:Ave|Street|St|Avenue|Road|Rd|Boulevard|Blvd|Lane|Ln|Drive|Dr)[,\s]+(?:[A-Za-z\s]+(?:Suite\s+\d+)?[,\s]+)?[A-Za-z\s]+[,\s]+[A-Z]{2}\s+\d{5}(?:-\d{4})?`,
			ContextClues:   []string{"address", "location", "street", "visit"},
			Priority:       1,
			ConfidenceBase: 0.7,
			Examples:       []string{"123 Business Ave, Suite 100, Anytown, NY 12345", "123 Main Street, Anytown, NY 12345"},
		},
		{
			PatternID:      "social_facebook",
			Name:           "Facebook URL",
			FieldType:      "social_media",
			RegexPattern:   `https?://(?:www\.)?facebook\.com/[a-zA-Z0-9._-]+`,
			ContextClues:   []string{"facebook", "social", "follow"},
			Priority:       2,
			ConfidenceBase: 0.9,
			Examples:       []string{"https://www.facebook.com/company"},
		},
		{
			PatternID:      "social_twitter",
			Name:           "Twitter URL",
			FieldType:      "social_media",
			RegexPattern:   `https?://(?:www\.)?twitter\.com/[a-zA-Z0-9._-]+`,
			ContextClues:   []string{"twitter", "social", "follow", "@"},
			Priority:       2,
			ConfidenceBase: 0.9,
			Examples:       []string{"https://twitter.com/company"},
		},
		{
			PatternID:      "social_linkedin",
			Name:           "LinkedIn URL",
			FieldType:      "social_media",
			RegexPattern:   `https?://(?:www\.)?linkedin\.com/(?:company|in)/[a-zA-Z0-9._-]+`,
			ContextClues:   []string{"linkedin", "professional", "connect"},
			Priority:       2,
			ConfidenceBase: 0.9,
			Examples:       []string{"https://www.linkedin.com/company/example"},
		},
		{
			PatternID:      "business_hours",
			Name:           "Business Hours",
			FieldType:      "business_hours",
			RegexPattern:   `(?:Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday|Mon|Tue|Wed|Thu|Fri|Sat|Sun).*?(?:\d{1,2}:\d{2}|\d{1,2}\s*(?:AM|PM|am|pm))`,
			ContextClues:   []string{"hours", "open", "closed", "schedule"},
			Priority:       3,
			ConfidenceBase: 0.7,
			Examples:       []string{"Monday: 9:00 AM - 5:00 PM"},
		},
		{
			PatternID:      "ein_number",
			Name:           "EIN/Tax ID Number",
			FieldType:      "tax_id",
			RegexPattern:   `\b\d{2}-\d{7}\b`,
			ContextClues:   []string{"ein", "tax id", "federal", "employer"},
			Priority:       1,
			ConfidenceBase: 0.9,
			Examples:       []string{"12-3456789"},
		},
		{
			PatternID:      "zip_code_us",
			Name:           "US ZIP Code",
			FieldType:      "postal_code",
			RegexPattern:   `\b\d{5}(?:-\d{4})?\b`,
			ContextClues:   []string{"zip", "postal", "code", "mail"},
			Priority:       2,
			ConfidenceBase: 0.8,
			Examples:       []string{"12345", "12345-6789"},
		},
		{
			PatternID:      "founded_year",
			Name:           "Founded Year",
			FieldType:      "founded",
			RegexPattern:   `(?:founded|established|since)\s+(?:in\s+)?(\d{4})`,
			ContextClues:   []string{"founded", "established", "since", "year"},
			Priority:       3,
			ConfidenceBase: 0.7,
			Examples:       []string{"Founded in 1995", "Established 2010"},
		},
	}
}
