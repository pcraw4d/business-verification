package risk

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RiskPatternDetector provides pattern-based risk detection capabilities
type RiskPatternDetector struct {
	logger *zap.Logger
	config *RiskDetectionConfig

	// Compiled patterns for performance
	compiledPatterns map[string]*regexp.Regexp
	patternMutex     sync.RWMutex
}

// NewRiskPatternDetector creates a new risk pattern detector
func NewRiskPatternDetector(logger *zap.Logger, config *RiskDetectionConfig) *RiskPatternDetector {
	return &RiskPatternDetector{
		logger:           logger,
		config:           config,
		compiledPatterns: make(map[string]*regexp.Regexp),
	}
}

// DetectPatterns detects risk patterns in business information
func (rpd *RiskPatternDetector) DetectPatterns(
	req *RiskDetectionRequest,
	riskKeywords []RiskKeyword,
) []DetectedPattern {
	var patterns []DetectedPattern

	// Combine all text content for pattern analysis
	content := rpd.combineContent(req)

	// Detect various types of risk patterns
	patterns = append(patterns, rpd.detectMoneyLaunderingPatterns(content)...)
	patterns = append(patterns, rpd.detectFraudPatterns(content)...)
	patterns = append(patterns, rpd.detectShellCompanyPatterns(content)...)
	patterns = append(patterns, rpd.detectSanctionsEvasionPatterns(content)...)
	patterns = append(patterns, rpd.detectTerroristFinancingPatterns(content)...)
	patterns = append(patterns, rpd.detectDrugTraffickingPatterns(content)...)
	patterns = append(patterns, rpd.detectWeaponsTraffickingPatterns(content)...)
	patterns = append(patterns, rpd.detectHumanTraffickingPatterns(content)...)
	patterns = append(patterns, rpd.detectCybercrimePatterns(content)...)
	patterns = append(patterns, rpd.detectSuspiciousActivityPatterns(content)...)

	// Remove duplicates and sort by confidence
	patterns = rpd.deduplicateAndSortPatterns(patterns)

	return patterns
}

// combineContent combines all available content for pattern analysis
func (rpd *RiskPatternDetector) combineContent(req *RiskDetectionRequest) string {
	var contentParts []string

	if req.BusinessName != "" {
		contentParts = append(contentParts, req.BusinessName)
	}

	if req.BusinessDescription != "" {
		contentParts = append(contentParts, req.BusinessDescription)
	}

	if req.WebsiteURL != "" {
		contentParts = append(contentParts, req.WebsiteURL)
	}

	if req.IndustryCode != "" {
		contentParts = append(contentParts, req.IndustryCode)
	}

	if req.MCCCode != "" {
		contentParts = append(contentParts, req.MCCCode)
	}

	if req.NAICSCode != "" {
		contentParts = append(contentParts, req.NAICSCode)
	}

	if req.SICCode != "" {
		contentParts = append(contentParts, req.SICCode)
	}

	return strings.Join(contentParts, " ")
}

// detectMoneyLaunderingPatterns detects money laundering indicators
func (rpd *RiskPatternDetector) detectMoneyLaunderingPatterns(content string) []DetectedPattern {
	var patterns []DetectedPattern

	// Cash-intensive business patterns
	cashPatterns := []string{
		`(?i)(cash\s*only|cash\s*preferred|cash\s*accepted)`,
		`(?i)(no\s*credit\s*card|no\s*debit\s*card)`,
		`(?i)(cash\s*transactions?|cash\s*payments?)`,
	}

	for _, pattern := range cashPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Cash-Intensive Business",
				PatternType: "money_laundering",
				Confidence:  0.6,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	// High-value transaction patterns
	highValuePatterns := []string{
		`(?i)(high\s*value|large\s*transactions?|bulk\s*purchases?)`,
		`(?i)(precious\s*metals?|gold|silver|diamonds?)`,
		`(?i)(art\s*dealer|antique\s*dealer|collectibles?)`,
	}

	for _, pattern := range highValuePatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "High-Value Transactions",
				PatternType: "money_laundering",
				Confidence:  0.7,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	return patterns
}

// detectFraudPatterns detects fraud indicators
func (rpd *RiskPatternDetector) detectFraudPatterns(content string) []DetectedPattern {
	var patterns []DetectedPattern

	// Identity fraud patterns
	identityPatterns := []string{
		`(?i)(fake\s*id|false\s*identity|stolen\s*identity)`,
		`(?i)(identity\s*theft|id\s*theft)`,
		`(?i)(synthetic\s*identity|artificial\s*identity)`,
	}

	for _, pattern := range identityPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Identity Fraud",
				PatternType: "fraud_pattern",
				Confidence:  0.8,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	// Credit card fraud patterns
	ccFraudPatterns := []string{
		`(?i)(credit\s*card\s*fraud|card\s*skimming)`,
		`(?i)(unauthorized\s*charges?|fraudulent\s*transactions?)`,
		`(?i)(card\s*cloning|card\s*counterfeiting)`,
	}

	for _, pattern := range ccFraudPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Credit Card Fraud",
				PatternType: "fraud_pattern",
				Confidence:  0.8,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	return patterns
}

// detectShellCompanyPatterns detects shell company indicators
func (rpd *RiskPatternDetector) detectShellCompanyPatterns(content string) []DetectedPattern {
	var patterns []DetectedPattern

	// Shell company indicators
	shellPatterns := []string{
		`(?i)(shell\s*company|front\s*company|paper\s*company)`,
		`(?i)(nominee\s*company|holding\s*company)`,
		`(?i)(offshore\s*company|tax\s*haven)`,
		`(?i)(bearer\s*shares?|anonymous\s*shares?)`,
	}

	for _, pattern := range shellPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Shell Company Indicators",
				PatternType: "shell_company",
				Confidence:  0.7,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	return patterns
}

// detectSanctionsEvasionPatterns detects sanctions evasion indicators
func (rpd *RiskPatternDetector) detectSanctionsEvasionPatterns(content string) []DetectedPattern {
	var patterns []DetectedPattern

	// Sanctions evasion patterns
	sanctionsPatterns := []string{
		`(?i)(sanctions?\s*evasion|embargo\s*evasion)`,
		`(?i)(prohibited\s*country|restricted\s*country)`,
		`(?i)(ofac\s*list|sanctions?\s*list)`,
		`(?i)(blocked\s*entity|designated\s*entity)`,
	}

	for _, pattern := range sanctionsPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Sanctions Evasion",
				PatternType: "sanctions_evasion",
				Confidence:  0.9,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	return patterns
}

// detectTerroristFinancingPatterns detects terrorist financing indicators
func (rpd *RiskPatternDetector) detectTerroristFinancingPatterns(content string) []DetectedPattern {
	var patterns []DetectedPattern

	// Terrorist financing patterns
	terrorismPatterns := []string{
		`(?i)(terrorist\s*financing|terrorism\s*financing)`,
		`(?i)(terrorist\s*organization|terrorist\s*group)`,
		`(?i)(extremist\s*organization|radical\s*group)`,
		`(?i)(jihad|jihadi|militant)`,
	}

	for _, pattern := range terrorismPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Terrorist Financing",
				PatternType: "terrorist_financing",
				Confidence:  0.95,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	return patterns
}

// detectDrugTraffickingPatterns detects drug trafficking indicators
func (rpd *RiskPatternDetector) detectDrugTraffickingPatterns(content string) []DetectedPattern {
	var patterns []DetectedPattern

	// Drug trafficking patterns
	drugPatterns := []string{
		`(?i)(drug\s*trafficking|narcotics?\s*trafficking)`,
		`(?i)(drug\s*dealing|drug\s*distribution)`,
		`(?i)(cocaine|heroin|marijuana|methamphetamine)`,
		`(?i)(drug\s*smuggling|narcotics?\s*smuggling)`,
	}

	for _, pattern := range drugPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Drug Trafficking",
				PatternType: "drug_trafficking",
				Confidence:  0.9,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	return patterns
}

// detectWeaponsTraffickingPatterns detects weapons trafficking indicators
func (rpd *RiskPatternDetector) detectWeaponsTraffickingPatterns(content string) []DetectedPattern {
	var patterns []DetectedPattern

	// Weapons trafficking patterns
	weaponsPatterns := []string{
		`(?i)(weapons?\s*trafficking|arms?\s*trafficking)`,
		`(?i)(weapons?\s*dealing|arms?\s*dealing)`,
		`(?i)(illegal\s*weapons?|illegal\s*arms?)`,
		`(?i)(firearms?\s*trafficking|gun\s*trafficking)`,
	}

	for _, pattern := range weaponsPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Weapons Trafficking",
				PatternType: "weapons_trafficking",
				Confidence:  0.9,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	return patterns
}

// detectHumanTraffickingPatterns detects human trafficking indicators
func (rpd *RiskPatternDetector) detectHumanTraffickingPatterns(content string) []DetectedPattern {
	var patterns []DetectedPattern

	// Human trafficking patterns
	humanTraffickingPatterns := []string{
		`(?i)(human\s*trafficking|people\s*trafficking)`,
		`(?i)(sex\s*trafficking|forced\s*labor)`,
		`(?i)(human\s*smuggling|people\s*smuggling)`,
		`(?i)(forced\s*prostitution|sex\s*slave)`,
	}

	for _, pattern := range humanTraffickingPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Human Trafficking",
				PatternType: "human_trafficking",
				Confidence:  0.95,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	return patterns
}

// detectCybercrimePatterns detects cybercrime indicators
func (rpd *RiskPatternDetector) detectCybercrimePatterns(content string) []DetectedPattern {
	var patterns []DetectedPattern

	// Cybercrime patterns
	cyberPatterns := []string{
		`(?i)(cybercrime|cyber\s*crime)`,
		`(?i)(hacking|computer\s*hacking)`,
		`(?i)(malware|virus|trojan)`,
		`(?i)(phishing|spoofing|identity\s*theft)`,
		`(?i)(ransomware|cryptocurrency\s*mining)`,
	}

	for _, pattern := range cyberPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Cybercrime",
				PatternType: "cybercrime",
				Confidence:  0.7,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	return patterns
}

// detectSuspiciousActivityPatterns detects general suspicious activity indicators
func (rpd *RiskPatternDetector) detectSuspiciousActivityPatterns(content string) []DetectedPattern {
	var patterns []DetectedPattern

	// Suspicious activity patterns
	suspiciousPatterns := []string{
		`(?i)(suspicious\s*activity|unusual\s*activity)`,
		`(?i)(money\s*laundering|ml\s*activity)`,
		`(?i)(structuring|smurfing)`,
		`(?i)(layering|integration)`,
		`(?i)(placement|placement\s*phase)`,
	}

	for _, pattern := range suspiciousPatterns {
		if rpd.matchesPattern(content, pattern) {
			patterns = append(patterns, DetectedPattern{
				PatternName: "Suspicious Activity",
				PatternType: "suspicious_activity",
				Confidence:  0.6,
				Context:     rpd.extractPatternContext(content, pattern),
				Source:      "business_description",
				DetectedAt:  time.Now(),
			})
		}
	}

	return patterns
}

// matchesPattern checks if content matches a regex pattern
func (rpd *RiskPatternDetector) matchesPattern(content, pattern string) bool {
	compiledPattern := rpd.getCompiledPattern(pattern)
	if compiledPattern == nil {
		return false
	}

	return compiledPattern.MatchString(content)
}

// extractPatternContext extracts context around a pattern match
func (rpd *RiskPatternDetector) extractPatternContext(content, pattern string) string {
	compiledPattern := rpd.getCompiledPattern(pattern)
	if compiledPattern == nil {
		return ""
	}

	matches := compiledPattern.FindAllStringIndex(content, -1)
	if len(matches) == 0 {
		return ""
	}

	// Use the first match
	start, end := matches[0][0], matches[0][1]
	contextSize := 50

	contextStart := start - contextSize
	if contextStart < 0 {
		contextStart = 0
	}

	contextEnd := end + contextSize
	if contextEnd > len(content) {
		contextEnd = len(content)
	}

	context := content[contextStart:contextEnd]

	// Add ellipsis if truncated
	if contextStart > 0 {
		context = "..." + context
	}
	if contextEnd < len(content) {
		context = context + "..."
	}

	return context
}

// getCompiledPattern gets or compiles a regex pattern
func (rpd *RiskPatternDetector) getCompiledPattern(pattern string) *regexp.Regexp {
	rpd.patternMutex.RLock()
	if compiled, exists := rpd.compiledPatterns[pattern]; exists {
		rpd.patternMutex.RUnlock()
		return compiled
	}
	rpd.patternMutex.RUnlock()

	// Compile pattern
	compiled, err := regexp.Compile(pattern)
	if err != nil {
		rpd.logger.Warn("Failed to compile pattern",
			zap.String("pattern", pattern),
			zap.Error(err))
		return nil
	}

	// Cache compiled pattern
	rpd.patternMutex.Lock()
	rpd.compiledPatterns[pattern] = compiled
	rpd.patternMutex.Unlock()

	return compiled
}

// deduplicateAndSortPatterns removes duplicate patterns and sorts by confidence
func (rpd *RiskPatternDetector) deduplicateAndSortPatterns(patterns []DetectedPattern) []DetectedPattern {
	// Create a map to track unique patterns
	uniquePatterns := make(map[string]DetectedPattern)

	for _, pattern := range patterns {
		// Create a key based on pattern name and type
		key := pattern.PatternName + "_" + pattern.PatternType

		// Keep the pattern with higher confidence
		if existing, exists := uniquePatterns[key]; !exists || pattern.Confidence > existing.Confidence {
			uniquePatterns[key] = pattern
		}
	}

	// Convert back to slice
	var result []DetectedPattern
	for _, pattern := range uniquePatterns {
		result = append(result, pattern)
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
