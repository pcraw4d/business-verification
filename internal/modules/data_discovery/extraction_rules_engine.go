package data_discovery

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// ExtractionRulesEngine generates extraction rules for discovered data points
type ExtractionRulesEngine struct {
	config    *DataDiscoveryConfig
	logger    *zap.Logger
	ruleTypes []RuleTypeDefinition
}

// RuleTypeDefinition defines a type of extraction rule
type RuleTypeDefinition struct {
	RuleType        string   `json:"rule_type"`
	DisplayName     string   `json:"display_name"`
	Description     string   `json:"description"`
	Complexity      int      `json:"complexity"`  // 1=simple, 5=complex
	Reliability     float64  `json:"reliability"` // 0.0-1.0
	Performance     float64  `json:"performance"` // 0.0-1.0 (higher=faster)
	ApplicableTypes []string `json:"applicable_types"`
	Templates       []string `json:"templates"`
}

// NewExtractionRulesEngine creates a new extraction rules engine
func NewExtractionRulesEngine(config *DataDiscoveryConfig, logger *zap.Logger) *ExtractionRulesEngine {
	return &ExtractionRulesEngine{
		config:    config,
		logger:    logger,
		ruleTypes: getRuleTypeDefinitions(),
	}
}

// GenerateRules generates extraction rules for discovered fields
func (ere *ExtractionRulesEngine) GenerateRules(ctx context.Context, fields []DiscoveredField, patterns []PatternMatch) ([]ExtractionRule, error) {
	var rules []ExtractionRule

	ere.logger.Debug("Starting extraction rule generation",
		zap.Int("fields_count", len(fields)),
		zap.Int("patterns_count", len(patterns)))

	// Group patterns by field type for rule generation
	patternGroups := ere.groupPatternsByFieldType(patterns)

	// Generate rules for each discovered field
	for _, field := range fields {
		fieldPatterns := patternGroups[field.FieldType]
		fieldRules := ere.generateFieldRules(field, fieldPatterns)
		rules = append(rules, fieldRules...)
	}

	// Optimize rules for performance and accuracy
	optimizedRules := ere.optimizeRules(rules)

	// Sort rules by priority and confidence
	ere.sortRulesByPriority(optimizedRules)

	ere.logger.Debug("Extraction rule generation completed",
		zap.Int("rules_generated", len(rules)),
		zap.Int("rules_optimized", len(optimizedRules)))

	return optimizedRules, nil
}

// generateFieldRules generates extraction rules for a specific field
func (ere *ExtractionRulesEngine) generateFieldRules(field DiscoveredField, patterns []PatternMatch) []ExtractionRule {
	var rules []ExtractionRule

	// Generate rules based on extraction method
	switch field.ExtractionMethod {
	case "pattern_matching":
		rules = append(rules, ere.generatePatternBasedRules(field, patterns)...)
	case "contextual_analysis":
		rules = append(rules, ere.generateContextualRules(field)...)
	case "xpath":
		rules = append(rules, ere.generateXPathRules(field, patterns)...)
	case "css_selector":
		rules = append(rules, ere.generateCSSRules(field, patterns)...)
	case "regex":
		rules = append(rules, ere.generateRegexRules(field, patterns)...)
	default:
		rules = append(rules, ere.generateFallbackRules(field)...)
	}

	// Enhance rules with validation and confidence scoring
	for i := range rules {
		ere.enhanceRule(&rules[i], field)
	}

	return rules
}

// generatePatternBasedRules generates rules based on pattern matches
func (ere *ExtractionRulesEngine) generatePatternBasedRules(field DiscoveredField, patterns []PatternMatch) []ExtractionRule {
	var rules []ExtractionRule

	if len(patterns) == 0 {
		return rules
	}

	// Analyze patterns to create optimized rules
	patternAnalysis := ere.analyzePatterns(patterns)

	// Create regex-based rule from patterns
	if patternAnalysis.CommonRegex != "" {
		rule := ExtractionRule{
			RuleID:          fmt.Sprintf("regex_%s_%d", field.FieldType, time.Now().UnixNano()),
			FieldType:       field.FieldType,
			Pattern:         "regex",
			RegexPattern:    patternAnalysis.CommonRegex,
			ContextClues:    patternAnalysis.ContextClues,
			Priority:        1,
			ConfidenceScore: patternAnalysis.Confidence,
			ApplicableTypes: []string{field.FieldType},
			Metadata: map[string]interface{}{
				"source":         "pattern_analysis",
				"pattern_count":  len(patterns),
				"sample_matches": ere.getSampleMatches(patterns, 3),
			},
		}
		rules = append(rules, rule)
	}

	// Create context-based rules
	if len(patternAnalysis.ContextClues) > 0 {
		rule := ExtractionRule{
			RuleID:          fmt.Sprintf("context_%s_%d", field.FieldType, time.Now().UnixNano()),
			FieldType:       field.FieldType,
			Pattern:         "context_based",
			ContextClues:    patternAnalysis.ContextClues,
			Priority:        2,
			ConfidenceScore: patternAnalysis.Confidence * 0.8,
			ApplicableTypes: []string{field.FieldType},
			Metadata: map[string]interface{}{
				"source":        "context_analysis",
				"context_clues": len(patternAnalysis.ContextClues),
			},
		}
		rules = append(rules, rule)
	}

	return rules
}

// generateContextualRules generates rules for contextual fields
func (ere *ExtractionRulesEngine) generateContextualRules(field DiscoveredField) []ExtractionRule {
	var rules []ExtractionRule

	// Generate rule based on field type and business context
	rule := ExtractionRule{
		RuleID:          fmt.Sprintf("contextual_%s_%d", field.FieldType, time.Now().UnixNano()),
		FieldType:       field.FieldType,
		Pattern:         "contextual",
		ContextClues:    ere.getContextualClues(field.FieldType),
		Priority:        3,
		ConfidenceScore: 0.6,
		ApplicableTypes: []string{field.FieldType},
		Metadata: map[string]interface{}{
			"source":           "contextual_discovery",
			"business_context": field.Metadata["business_type"],
		},
	}

	rules = append(rules, rule)
	return rules
}

// generateXPathRules generates XPath-based extraction rules
func (ere *ExtractionRulesEngine) generateXPathRules(field DiscoveredField, patterns []PatternMatch) []ExtractionRule {
	var rules []ExtractionRule

	// Generate common XPath selectors for the field type
	xpathSelectors := ere.getXPathSelectors(field.FieldType)

	for i, selector := range xpathSelectors {
		rule := ExtractionRule{
			RuleID:          fmt.Sprintf("xpath_%s_%d_%d", field.FieldType, i, time.Now().UnixNano()),
			FieldType:       field.FieldType,
			Pattern:         "xpath",
			XPathSelector:   selector,
			Priority:        2,
			ConfidenceScore: 0.7,
			ApplicableTypes: []string{field.FieldType},
			Metadata: map[string]interface{}{
				"source":    "xpath_generation",
				"selector":  selector,
				"rule_type": "structural",
			},
		}
		rules = append(rules, rule)
	}

	return rules
}

// generateCSSRules generates CSS selector-based extraction rules
func (ere *ExtractionRulesEngine) generateCSSRules(field DiscoveredField, patterns []PatternMatch) []ExtractionRule {
	var rules []ExtractionRule

	// Generate common CSS selectors for the field type
	cssSelectors := ere.getCSSSelectors(field.FieldType)

	for i, selector := range cssSelectors {
		rule := ExtractionRule{
			RuleID:          fmt.Sprintf("css_%s_%d_%d", field.FieldType, i, time.Now().UnixNano()),
			FieldType:       field.FieldType,
			Pattern:         "css_selector",
			CSSSelector:     selector,
			Priority:        2,
			ConfidenceScore: 0.75,
			ApplicableTypes: []string{field.FieldType},
			Metadata: map[string]interface{}{
				"source":    "css_generation",
				"selector":  selector,
				"rule_type": "structural",
			},
		}
		rules = append(rules, rule)
	}

	return rules
}

// generateRegexRules generates regex-based extraction rules
func (ere *ExtractionRulesEngine) generateRegexRules(field DiscoveredField, patterns []PatternMatch) []ExtractionRule {
	var rules []ExtractionRule

	// Get built-in regex patterns for the field type
	regexPatterns := ere.getRegexPatterns(field.FieldType)

	for i, pattern := range regexPatterns {
		rule := ExtractionRule{
			RuleID:          fmt.Sprintf("regex_%s_%d_%d", field.FieldType, i, time.Now().UnixNano()),
			FieldType:       field.FieldType,
			Pattern:         "regex",
			RegexPattern:    pattern,
			Priority:        1,
			ConfidenceScore: 0.8,
			ApplicableTypes: []string{field.FieldType},
			Metadata: map[string]interface{}{
				"source":    "builtin_regex",
				"pattern":   pattern,
				"rule_type": "pattern_based",
			},
		}
		rules = append(rules, rule)
	}

	return rules
}

// generateFallbackRules generates fallback rules for fields without specific patterns
func (ere *ExtractionRulesEngine) generateFallbackRules(field DiscoveredField) []ExtractionRule {
	var rules []ExtractionRule

	// Generate generic rule based on field type
	rule := ExtractionRule{
		RuleID:          fmt.Sprintf("fallback_%s_%d", field.FieldType, time.Now().UnixNano()),
		FieldType:       field.FieldType,
		Pattern:         "generic",
		ContextClues:    ere.getGenericContextClues(field.FieldType),
		Priority:        5,
		ConfidenceScore: 0.4,
		ApplicableTypes: []string{field.FieldType},
		Metadata: map[string]interface{}{
			"source":    "fallback_generation",
			"rule_type": "generic",
		},
	}

	rules = append(rules, rule)
	return rules
}

// PatternAnalysis represents analysis results of pattern matches
type PatternAnalysis struct {
	CommonRegex  string   `json:"common_regex"`
	ContextClues []string `json:"context_clues"`
	Confidence   float64  `json:"confidence"`
	PatternCount int      `json:"pattern_count"`
}

// analyzePatterns analyzes pattern matches to derive common patterns
func (ere *ExtractionRulesEngine) analyzePatterns(patterns []PatternMatch) PatternAnalysis {
	analysis := PatternAnalysis{
		PatternCount: len(patterns),
		ContextClues: []string{},
		Confidence:   0.0,
	}

	if len(patterns) == 0 {
		return analysis
	}

	// Extract common regex pattern (simplified approach)
	if len(patterns) > 0 {
		firstPattern := patterns[0]
		if regex, exists := firstPattern.Metadata["regex_pattern"].(string); exists {
			analysis.CommonRegex = regex
		}
	}

	// Extract context clues from all patterns
	contextMap := make(map[string]int)
	for _, pattern := range patterns {
		// Extract words from context
		words := strings.Fields(strings.ToLower(pattern.Context))
		for _, word := range words {
			if len(word) > 3 { // Filter short words
				contextMap[word]++
			}
		}
	}

	// Select most common context clues
	for clue, count := range contextMap {
		if count >= len(patterns)/2 { // Appears in at least half of patterns
			analysis.ContextClues = append(analysis.ContextClues, clue)
		}
	}

	// Calculate confidence based on pattern consistency
	var totalConfidence float64
	for _, pattern := range patterns {
		totalConfidence += pattern.ConfidenceScore
	}
	analysis.Confidence = totalConfidence / float64(len(patterns))

	return analysis
}

// enhanceRule enhances a rule with additional metadata and validation
func (ere *ExtractionRulesEngine) enhanceRule(rule *ExtractionRule, field DiscoveredField) {
	// Add field-specific metadata
	if rule.Metadata == nil {
		rule.Metadata = make(map[string]interface{})
	}

	rule.Metadata["field_name"] = field.FieldName
	rule.Metadata["data_type"] = field.DataType
	rule.Metadata["business_value"] = field.BusinessValue
	rule.Metadata["generated_at"] = time.Now()

	// Adjust confidence based on field confidence
	rule.ConfidenceScore = (rule.ConfidenceScore + field.ConfidenceScore) / 2.0

	// Add validation hints
	rule.Metadata["validation_rules"] = field.ValidationRules
	rule.Metadata["sample_values"] = field.SampleValues
}

// optimizeRules optimizes extraction rules for performance and accuracy
func (ere *ExtractionRulesEngine) optimizeRules(rules []ExtractionRule) []ExtractionRule {
	var optimized []ExtractionRule

	// Group rules by field type
	ruleGroups := make(map[string][]ExtractionRule)
	for _, rule := range rules {
		ruleGroups[rule.FieldType] = append(ruleGroups[rule.FieldType], rule)
	}

	// Optimize each group
	for fieldType, fieldRules := range ruleGroups {
		optimizedGroup := ere.optimizeRuleGroup(fieldType, fieldRules)
		optimized = append(optimized, optimizedGroup...)
	}

	return optimized
}

// optimizeRuleGroup optimizes a group of rules for the same field type
func (ere *ExtractionRulesEngine) optimizeRuleGroup(fieldType string, rules []ExtractionRule) []ExtractionRule {
	if len(rules) <= 1 {
		return rules
	}

	// Remove duplicate rules
	uniqueRules := ere.removeDuplicateRules(rules)

	// Merge similar rules
	mergedRules := ere.mergeSimilarRules(uniqueRules)

	// Limit rules per field type based on configuration
	if len(mergedRules) > ere.config.MaxPatternsPerField {
		// Keep only the highest confidence rules
		ere.sortRulesByConfidence(mergedRules)
		mergedRules = mergedRules[:ere.config.MaxPatternsPerField]
	}

	return mergedRules
}

// removeDuplicateRules removes duplicate rules
func (ere *ExtractionRulesEngine) removeDuplicateRules(rules []ExtractionRule) []ExtractionRule {
	seen := make(map[string]bool)
	var unique []ExtractionRule

	for _, rule := range rules {
		key := ere.generateRuleKey(rule)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, rule)
		}
	}

	return unique
}

// mergeSimilarRules merges rules that are very similar
func (ere *ExtractionRulesEngine) mergeSimilarRules(rules []ExtractionRule) []ExtractionRule {
	// This is a simplified implementation
	// In practice, you would implement more sophisticated rule merging
	return rules
}

// generateRuleKey generates a unique key for a rule to detect duplicates
func (ere *ExtractionRulesEngine) generateRuleKey(rule ExtractionRule) string {
	return fmt.Sprintf("%s:%s:%s:%s:%s",
		rule.FieldType,
		rule.Pattern,
		rule.RegexPattern,
		rule.XPathSelector,
		rule.CSSSelector)
}

// sortRulesByPriority sorts rules by priority and confidence
func (ere *ExtractionRulesEngine) sortRulesByPriority(rules []ExtractionRule) {
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			// Primary sort: priority (lower number = higher priority)
			if rules[i].Priority > rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			} else if rules[i].Priority == rules[j].Priority {
				// Secondary sort: confidence (higher = better)
				if rules[i].ConfidenceScore < rules[j].ConfidenceScore {
					rules[i], rules[j] = rules[j], rules[i]
				}
			}
		}
	}
}

// sortRulesByConfidence sorts rules by confidence score (descending)
func (ere *ExtractionRulesEngine) sortRulesByConfidence(rules []ExtractionRule) {
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].ConfidenceScore < rules[j].ConfidenceScore {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}
}

// Helper methods for generating selectors and patterns

func (ere *ExtractionRulesEngine) groupPatternsByFieldType(patterns []PatternMatch) map[string][]PatternMatch {
	groups := make(map[string][]PatternMatch)
	for _, pattern := range patterns {
		groups[pattern.FieldType] = append(groups[pattern.FieldType], pattern)
	}
	return groups
}

func (ere *ExtractionRulesEngine) getSampleMatches(patterns []PatternMatch, count int) []string {
	var samples []string
	for i, pattern := range patterns {
		if i >= count {
			break
		}
		samples = append(samples, pattern.MatchedText)
	}
	return samples
}

func (ere *ExtractionRulesEngine) getContextualClues(fieldType string) []string {
	contextMap := map[string][]string{
		"email":             {"email", "contact", "mail", "@"},
		"phone":             {"phone", "tel", "call", "contact"},
		"address":           {"address", "location", "street", "visit"},
		"url":               {"website", "url", "link", "http"},
		"social_media":      {"social", "facebook", "twitter", "linkedin"},
		"business_hours":    {"hours", "open", "closed", "schedule"},
		"payment_methods":   {"payment", "pay", "credit", "card"},
		"shipping_info":     {"shipping", "delivery", "ship"},
		"api_documentation": {"api", "documentation", "docs", "developer"},
	}

	if clues, exists := contextMap[fieldType]; exists {
		return clues
	}
	return []string{}
}

func (ere *ExtractionRulesEngine) getXPathSelectors(fieldType string) []string {
	selectorMap := map[string][]string{
		"email": {
			"//a[contains(@href, 'mailto:')]/@href",
			"//*[contains(text(), '@')]",
			"//input[@type='email']/@value",
		},
		"phone": {
			"//a[contains(@href, 'tel:')]/@href",
			"//*[contains(text(), 'phone') or contains(text(), 'tel')]",
		},
		"address": {
			"//*[contains(@class, 'address')]",
			"//*[contains(text(), 'address') or contains(text(), 'location')]",
		},
		"url": {
			"//a/@href",
			"//link[@rel='canonical']/@href",
		},
	}

	if selectors, exists := selectorMap[fieldType]; exists {
		return selectors
	}
	return []string{}
}

func (ere *ExtractionRulesEngine) getCSSSelectors(fieldType string) []string {
	selectorMap := map[string][]string{
		"email": {
			"a[href^='mailto:']",
			"input[type='email']",
			".contact-email, .email",
		},
		"phone": {
			"a[href^='tel:']",
			".phone, .telephone, .contact-phone",
		},
		"address": {
			".address, .location, .contact-address",
			"[itemprop='address']",
		},
		"url": {
			"a[href^='http']",
			"link[rel='canonical']",
		},
		"social_media": {
			"a[href*='facebook.com']",
			"a[href*='twitter.com']",
			"a[href*='linkedin.com']",
		},
	}

	if selectors, exists := selectorMap[fieldType]; exists {
		return selectors
	}
	return []string{}
}

func (ere *ExtractionRulesEngine) getRegexPatterns(fieldType string) []string {
	patternMap := map[string][]string{
		"email": {
			`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`,
		},
		"phone": {
			`\+?1?[-.\s]?\(?([0-9]{3})\)?[-.\s]?([0-9]{3})[-.\s]?([0-9]{4})`,
			`\+[1-9]\d{1,14}`,
		},
		"url": {
			`https?://[^\s<>"{}|\\^` + "`" + `\[\]]+`,
		},
		"social_media": {
			`https?://(?:www\.)?facebook\.com/[a-zA-Z0-9._-]+`,
			`https?://(?:www\.)?twitter\.com/[a-zA-Z0-9._-]+`,
			`https?://(?:www\.)?linkedin\.com/(?:company|in)/[a-zA-Z0-9._-]+`,
		},
	}

	if patterns, exists := patternMap[fieldType]; exists {
		return patterns
	}
	return []string{}
}

func (ere *ExtractionRulesEngine) getGenericContextClues(fieldType string) []string {
	// Return generic context clues for unknown field types
	return []string{strings.Replace(fieldType, "_", " ", -1)}
}

// getRuleTypeDefinitions returns built-in rule type definitions
func getRuleTypeDefinitions() []RuleTypeDefinition {
	return []RuleTypeDefinition{
		{
			RuleType:        "regex",
			DisplayName:     "Regular Expression",
			Description:     "Pattern-based extraction using regular expressions",
			Complexity:      2,
			Reliability:     0.8,
			Performance:     0.9,
			ApplicableTypes: []string{"email", "phone", "url", "social_media"},
			Templates:       []string{},
		},
		{
			RuleType:        "xpath",
			DisplayName:     "XPath Selector",
			Description:     "Structural extraction using XPath selectors",
			Complexity:      3,
			Reliability:     0.9,
			Performance:     0.7,
			ApplicableTypes: []string{"email", "phone", "address", "url"},
			Templates:       []string{},
		},
		{
			RuleType:        "css_selector",
			DisplayName:     "CSS Selector",
			Description:     "Structural extraction using CSS selectors",
			Complexity:      2,
			Reliability:     0.85,
			Performance:     0.8,
			ApplicableTypes: []string{"email", "phone", "address", "url", "social_media"},
			Templates:       []string{},
		},
		{
			RuleType:        "contextual",
			DisplayName:     "Contextual Analysis",
			Description:     "Context-based extraction using surrounding content",
			Complexity:      4,
			Reliability:     0.7,
			Performance:     0.6,
			ApplicableTypes: []string{"business_hours", "services", "industry"},
			Templates:       []string{},
		},
		{
			RuleType:        "structured_data",
			DisplayName:     "Structured Data",
			Description:     "Extraction from JSON-LD, microdata, and other structured formats",
			Complexity:      3,
			Reliability:     0.95,
			Performance:     0.9,
			ApplicableTypes: []string{"all"},
			Templates:       []string{},
		},
	}
}
