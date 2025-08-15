package webanalysis

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// ConnectionValidator validates the connection between business information and website content
type ConnectionValidator struct {
	nameMatcher      *BusinessNameMatcher
	addressValidator *AddressValidator
	contactValidator *ContactValidator
	domainAnalyzer   *DomainAnalyzer
	confidenceEngine *ConnectionConfidenceEngine
}

// ConnectionValidationResult represents the result of connection validation
type ConnectionValidationResult struct {
	BusinessNameMatch  BusinessNameMatchResult `json:"business_name_match"`
	AddressValidation  AddressValidationResult `json:"address_validation"`
	ContactValidation  ContactValidationResult `json:"contact_validation"`
	DomainAnalysis     DomainAnalysisResult    `json:"domain_analysis"`
	OverallConfidence  float64                 `json:"overall_confidence"`
	ValidationTime     time.Time               `json:"validation_time"`
	ValidationMetadata map[string]interface{}  `json:"validation_metadata"`
}

// BusinessNameMatchResult represents business name matching results
type BusinessNameMatchResult struct {
	IsMatch          bool     `json:"is_match"`
	Confidence       float64  `json:"confidence"`
	MatchType        string   `json:"match_type"` // exact, partial, fuzzy, none
	MatchedName      string   `json:"matched_name"`
	MatchedPositions []int    `json:"matched_positions"`
	SimilarityScore  float64  `json:"similarity_score"`
	Evidence         []string `json:"evidence"`
}

// AddressValidationResult represents address validation results
type AddressValidationResult struct {
	IsValid           bool              `json:"is_valid"`
	Confidence        float64           `json:"confidence"`
	ValidatedAddress  string            `json:"validated_address"`
	AddressComponents map[string]string `json:"address_components"`
	ValidationType    string            `json:"validation_type"` // exact, partial, none
	Evidence          []string          `json:"evidence"`
}

// ContactValidationResult represents contact validation results
type ContactValidationResult struct {
	IsValid        bool     `json:"is_valid"`
	Confidence     float64  `json:"confidence"`
	ValidatedPhone string   `json:"validated_phone"`
	ValidatedEmail string   `json:"validated_email"`
	ValidationType string   `json:"validation_type"` // exact, partial, none
	Evidence       []string `json:"evidence"`
}

// DomainAnalysisResult represents domain analysis results
type DomainAnalysisResult struct {
	IsRelevant        bool     `json:"is_relevant"`
	Confidence        float64  `json:"confidence"`
	DomainName        string   `json:"domain_name"`
	DomainAge         int      `json:"domain_age"` // in days
	DomainAuthority   float64  `json:"domain_authority"`
	BusinessRelevance float64  `json:"business_relevance"`
	Evidence          []string `json:"evidence"`
}

// NewConnectionValidator creates a new connection validator
func NewConnectionValidator() *ConnectionValidator {
	return &ConnectionValidator{
		nameMatcher:      NewBusinessNameMatcher(),
		addressValidator: NewAddressValidator(),
		contactValidator: NewContactValidator(),
		domainAnalyzer:   NewDomainAnalyzer(),
		confidenceEngine: NewConnectionConfidenceEngine(),
	}
}

// ValidateConnection performs comprehensive connection validation
func (cv *ConnectionValidator) ValidateConnection(ctx context.Context, business string, website string, content *ScrapedContent) (*ConnectionValidationResult, error) {
	// Parse domain from website URL
	domain, err := cv.extractDomain(website)
	if err != nil {
		return nil, fmt.Errorf("failed to extract domain: %w", err)
	}

	// Perform business name matching
	nameMatch := cv.nameMatcher.MatchBusinessName(business, content)

	// Perform address validation
	addressValidation := cv.addressValidator.ValidateAddress(business, content)

	// Perform contact validation
	contactValidation := cv.contactValidator.ValidateContact(business, content)

	// Perform domain analysis
	domainAnalysis := cv.domainAnalyzer.AnalyzeDomain(domain, business, content)

	// Calculate overall confidence
	overallConfidence := cv.confidenceEngine.CalculateOverallConfidence(
		nameMatch, addressValidation, contactValidation, domainAnalysis)

	// Create validation metadata
	metadata := map[string]interface{}{
		"business":       business,
		"website":        website,
		"domain":         domain,
		"content_length": len(content.Text),
		"validation_components": []string{
			"business_name_match",
			"address_validation",
			"contact_validation",
			"domain_analysis",
		},
	}

	result := &ConnectionValidationResult{
		BusinessNameMatch:  *nameMatch,
		AddressValidation:  *addressValidation,
		ContactValidation:  *contactValidation,
		DomainAnalysis:     *domainAnalysis,
		OverallConfidence:  overallConfidence,
		ValidationTime:     time.Now(),
		ValidationMetadata: metadata,
	}

	return result, nil
}

// extractDomain extracts domain from website URL
func (cv *ConnectionValidator) extractDomain(website string) (string, error) {
	// Add protocol if missing
	if !strings.HasPrefix(website, "http://") && !strings.HasPrefix(website, "https://") {
		website = "https://" + website
	}

	parsedURL, err := url.Parse(website)
	if err != nil {
		return "", err
	}

	return parsedURL.Hostname(), nil
}

// BusinessNameMatcher handles business name matching algorithms
type BusinessNameMatcher struct {
	exactMatcher   *ExactNameMatcher
	partialMatcher *PartialNameMatcher
	fuzzyMatcher   *FuzzyNameMatcher
}

// NewBusinessNameMatcher creates a new business name matcher
func NewBusinessNameMatcher() *BusinessNameMatcher {
	return &BusinessNameMatcher{
		exactMatcher:   NewExactNameMatcher(),
		partialMatcher: NewPartialNameMatcher(),
		fuzzyMatcher:   NewFuzzyNameMatcher(),
	}
}

// MatchBusinessName performs business name matching
func (bnm *BusinessNameMatcher) MatchBusinessName(business string, content *ScrapedContent) *BusinessNameMatchResult {
	// Try exact matching first
	if result := bnm.exactMatcher.Match(business, content); result.IsMatch {
		return result
	}

	// Try partial matching
	if result := bnm.partialMatcher.Match(business, content); result.IsMatch {
		return result
	}

	// Try fuzzy matching
	if result := bnm.fuzzyMatcher.Match(business, content); result.IsMatch {
		return result
	}

	// No match found
	return &BusinessNameMatchResult{
		IsMatch:          false,
		Confidence:       0.0,
		MatchType:        "none",
		MatchedName:      "",
		MatchedPositions: []int{},
		SimilarityScore:  0.0,
		Evidence:         []string{"No business name match found"},
	}
}

// ExactNameMatcher performs exact business name matching
type ExactNameMatcher struct{}

// NewExactNameMatcher creates a new exact name matcher
func NewExactNameMatcher() *ExactNameMatcher {
	return &ExactNameMatcher{}
}

// Match performs exact name matching
func (enm *ExactNameMatcher) Match(business string, content *ScrapedContent) *BusinessNameMatchResult {
	normalizedBusiness := strings.ToLower(strings.TrimSpace(business))
	normalizedContent := strings.ToLower(content.Text)

	if strings.Contains(normalizedContent, normalizedBusiness) {
		// Find all positions where the business name appears
		var positions []int
		start := 0
		for {
			pos := strings.Index(normalizedContent[start:], normalizedBusiness)
			if pos == -1 {
				break
			}
			positions = append(positions, start+pos)
			start += pos + len(normalizedBusiness)
		}

		return &BusinessNameMatchResult{
			IsMatch:          true,
			Confidence:       1.0,
			MatchType:        "exact",
			MatchedName:      business,
			MatchedPositions: positions,
			SimilarityScore:  1.0,
			Evidence:         []string{fmt.Sprintf("Exact match found at positions: %v", positions)},
		}
	}

	return &BusinessNameMatchResult{IsMatch: false}
}

// PartialNameMatcher performs partial business name matching
type PartialNameMatcher struct{}

// NewPartialNameMatcher creates a new partial name matcher
func NewPartialNameMatcher() *PartialNameMatcher {
	return &PartialNameMatcher{}
}

// Match performs partial name matching
func (pnm *PartialNameMatcher) Match(business string, content *ScrapedContent) *BusinessNameMatchResult {
	normalizedBusiness := strings.ToLower(strings.TrimSpace(business))
	normalizedContent := strings.ToLower(content.Text)

	// Split business name into words
	businessWords := strings.Fields(normalizedBusiness)
	if len(businessWords) < 2 {
		return &BusinessNameMatchResult{IsMatch: false}
	}

	// Find matching words
	var matchedWords []string
	var positions []int
	totalWords := len(businessWords)
	matchedCount := 0

	for _, word := range businessWords {
		if len(word) < 3 { // Skip very short words
			continue
		}
		if strings.Contains(normalizedContent, word) {
			matchedWords = append(matchedWords, word)
			matchedCount++
		}
	}

	// Calculate match ratio
	matchRatio := float64(matchedCount) / float64(totalWords)
	confidence := matchRatio * 0.8 // Partial match gets max 80% confidence

	if matchRatio >= 0.6 { // At least 60% of words must match
		return &BusinessNameMatchResult{
			IsMatch:          true,
			Confidence:       confidence,
			MatchType:        "partial",
			MatchedName:      strings.Join(matchedWords, " "),
			MatchedPositions: positions,
			SimilarityScore:  matchRatio,
			Evidence:         []string{fmt.Sprintf("Partial match: %d/%d words matched", matchedCount, totalWords)},
		}
	}

	return &BusinessNameMatchResult{IsMatch: false}
}

// FuzzyNameMatcher performs fuzzy business name matching
type FuzzyNameMatcher struct{}

// NewFuzzyNameMatcher creates a new fuzzy name matcher
func NewFuzzyNameMatcher() *FuzzyNameMatcher {
	return &FuzzyNameMatcher{}
}

// Match performs fuzzy name matching
func (fnm *FuzzyNameMatcher) Match(business string, content *ScrapedContent) *BusinessNameMatchResult {
	normalizedBusiness := strings.ToLower(strings.TrimSpace(business))
	normalizedContent := strings.ToLower(content.Text)

	// Calculate Jaccard similarity
	similarity := fnm.calculateJaccardSimilarity(normalizedBusiness, normalizedContent)
	confidence := similarity * 0.6 // Fuzzy match gets max 60% confidence

	if similarity >= 0.3 { // At least 30% similarity
		return &BusinessNameMatchResult{
			IsMatch:          true,
			Confidence:       confidence,
			MatchType:        "fuzzy",
			MatchedName:      business,
			MatchedPositions: []int{},
			SimilarityScore:  similarity,
			Evidence:         []string{fmt.Sprintf("Fuzzy match with similarity: %.2f", similarity)},
		}
	}

	return &BusinessNameMatchResult{IsMatch: false}
}

// calculateJaccardSimilarity calculates Jaccard similarity between two strings
func (fnm *FuzzyNameMatcher) calculateJaccardSimilarity(str1, str2 string) float64 {
	// Create character sets
	set1 := make(map[rune]bool)
	set2 := make(map[rune]bool)

	for _, char := range str1 {
		if char != ' ' {
			set1[char] = true
		}
	}

	for _, char := range str2 {
		if char != ' ' {
			set2[char] = true
		}
	}

	// Calculate intersection
	intersection := 0
	for char := range set1 {
		if set2[char] {
			intersection++
		}
	}

	// Calculate union
	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)
}

// AddressValidator validates address information
type AddressValidator struct {
	addressPatterns []*regexp.Regexp
}

// NewAddressValidator creates a new address validator
func NewAddressValidator() *AddressValidator {
	av := &AddressValidator{}
	av.initializePatterns()
	return av
}

// ValidateAddress validates address information
func (av *AddressValidator) ValidateAddress(business string, content *ScrapedContent) *AddressValidationResult {
	normalizedContent := strings.ToLower(content.Text)

	// Look for address patterns
	var foundAddresses []string
	for _, pattern := range av.addressPatterns {
		matches := pattern.FindAllString(content.Text, -1)
		foundAddresses = append(foundAddresses, matches...)
	}

	if len(foundAddresses) > 0 {
		// Extract address components
		components := av.extractAddressComponents(foundAddresses[0])

		return &AddressValidationResult{
			IsValid:           true,
			Confidence:        0.7,
			ValidatedAddress:  foundAddresses[0],
			AddressComponents: components,
			ValidationType:    "partial",
			Evidence:          []string{fmt.Sprintf("Found address: %s", foundAddresses[0])},
		}
	}

	return &AddressValidationResult{
		IsValid:        false,
		Confidence:     0.0,
		ValidationType: "none",
		Evidence:       []string{"No address information found"},
	}
}

// initializePatterns initializes address patterns
func (av *AddressValidator) initializePatterns() {
	av.addressPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\d+\s+[A-Za-z\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way|Terrace|Ter),\s*[A-Za-z\s]+,\s*[A-Z]{2}\s+\d{5}`),
		regexp.MustCompile(`\d+\s+[A-Za-z\s]+(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way|Terrace|Ter)`),
		regexp.MustCompile(`[A-Za-z\s]+,\s*[A-Z]{2}\s+\d{5}`),
	}
}

// extractAddressComponents extracts address components
func (av *AddressValidator) extractAddressComponents(address string) map[string]string {
	components := make(map[string]string)
	components["full_address"] = address

	// Extract street number
	if match := regexp.MustCompile(`^(\d+)`).FindStringSubmatch(address); len(match) > 1 {
		components["street_number"] = match[1]
	}

	// Extract street name
	if match := regexp.MustCompile(`^\d+\s+([A-Za-z\s]+?)(?:Street|St|Avenue|Ave|Road|Rd|Boulevard|Blvd|Drive|Dr|Lane|Ln|Court|Ct|Place|Pl|Way|Terrace|Ter)`).FindStringSubmatch(address); len(match) > 1 {
		components["street_name"] = strings.TrimSpace(match[1])
	}

	// Extract city and state
	if match := regexp.MustCompile(`([A-Za-z\s]+),\s*([A-Z]{2})\s+(\d{5})`).FindStringSubmatch(address); len(match) > 3 {
		components["city"] = strings.TrimSpace(match[1])
		components["state"] = match[2]
		components["zip_code"] = match[3]
	}

	return components
}

// ContactValidator validates contact information
type ContactValidator struct {
	phonePattern *regexp.Regexp
	emailPattern *regexp.Regexp
}

// NewContactValidator creates a new contact validator
func NewContactValidator() *ContactValidator {
	return &ContactValidator{
		phonePattern: regexp.MustCompile(`\(?\d{3}\)?[\s.-]?\d{3}[\s.-]?\d{4}`),
		emailPattern: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
	}
}

// ValidateContact validates contact information
func (cv *ContactValidator) ValidateContact(business string, content *ScrapedContent) *ContactValidationResult {
	// Find phone numbers
	phoneMatches := cv.phonePattern.FindAllString(content.Text, -1)
	var validatedPhone string
	if len(phoneMatches) > 0 {
		validatedPhone = phoneMatches[0]
	}

	// Find email addresses
	emailMatches := cv.emailPattern.FindAllString(content.Text, -1)
	var validatedEmail string
	if len(emailMatches) > 0 {
		validatedEmail = emailMatches[0]
	}

	confidence := 0.0
	validationType := "none"
	var evidence []string

	if validatedPhone != "" && validatedEmail != "" {
		confidence = 0.8
		validationType = "exact"
		evidence = append(evidence, fmt.Sprintf("Found phone: %s", validatedPhone))
		evidence = append(evidence, fmt.Sprintf("Found email: %s", validatedEmail))
	} else if validatedPhone != "" {
		confidence = 0.6
		validationType = "partial"
		evidence = append(evidence, fmt.Sprintf("Found phone: %s", validatedPhone))
	} else if validatedEmail != "" {
		confidence = 0.5
		validationType = "partial"
		evidence = append(evidence, fmt.Sprintf("Found email: %s", validatedEmail))
	} else {
		evidence = append(evidence, "No contact information found")
	}

	return &ContactValidationResult{
		IsValid:        validatedPhone != "" || validatedEmail != "",
		Confidence:     confidence,
		ValidatedPhone: validatedPhone,
		ValidatedEmail: validatedEmail,
		ValidationType: validationType,
		Evidence:       evidence,
	}
}

// DomainAnalyzer analyzes domain information
type DomainAnalyzer struct{}

// NewDomainAnalyzer creates a new domain analyzer
func NewDomainAnalyzer() *DomainAnalyzer {
	return &DomainAnalyzer{}
}

// AnalyzeDomain analyzes domain information
func (da *DomainAnalyzer) AnalyzeDomain(domain, business string, content *ScrapedContent) *DomainAnalysisResult {
	// Simple domain analysis - in production, you'd use domain authority APIs
	domainAge := 365       // Placeholder - would be calculated from WHOIS data
	domainAuthority := 0.7 // Placeholder - would be calculated from SEO metrics
	businessRelevance := da.calculateBusinessRelevance(domain, business, content)

	confidence := (domainAuthority + businessRelevance) / 2.0
	isRelevant := confidence >= 0.5

	var evidence []string
	if isRelevant {
		evidence = append(evidence, fmt.Sprintf("Domain relevance score: %.2f", businessRelevance))
		evidence = append(evidence, fmt.Sprintf("Domain authority: %.2f", domainAuthority))
	} else {
		evidence = append(evidence, "Low domain relevance or authority")
	}

	return &DomainAnalysisResult{
		IsRelevant:        isRelevant,
		Confidence:        confidence,
		DomainName:        domain,
		DomainAge:         domainAge,
		DomainAuthority:   domainAuthority,
		BusinessRelevance: businessRelevance,
		Evidence:          evidence,
	}
}

// calculateBusinessRelevance calculates business relevance of domain
func (da *DomainAnalyzer) calculateBusinessRelevance(domain, business string, content *ScrapedContent) float64 {
	normalizedDomain := strings.ToLower(domain)
	normalizedBusiness := strings.ToLower(business)

	// Check if business name appears in domain
	if strings.Contains(normalizedDomain, normalizedBusiness) {
		return 0.9
	}

	// Check if business name words appear in domain
	businessWords := strings.Fields(normalizedBusiness)
	matchedWords := 0
	for _, word := range businessWords {
		if len(word) > 2 && strings.Contains(normalizedDomain, word) {
			matchedWords++
		}
	}

	if len(businessWords) > 0 {
		return float64(matchedWords) / float64(len(businessWords)) * 0.7
	}

	return 0.3 // Default low relevance
}

// ConnectionConfidenceEngine calculates overall connection confidence
type ConnectionConfidenceEngine struct {
	weights map[string]float64
}

// NewConnectionConfidenceEngine creates a new confidence engine
func NewConnectionConfidenceEngine() *ConnectionConfidenceEngine {
	return &ConnectionConfidenceEngine{
		weights: map[string]float64{
			"business_name": 0.4,
			"address":       0.2,
			"contact":       0.2,
			"domain":        0.2,
		},
	}
}

// CalculateOverallConfidence calculates overall connection confidence
func (cce *ConnectionConfidenceEngine) CalculateOverallConfidence(
	nameMatch *BusinessNameMatchResult,
	addressValidation *AddressValidationResult,
	contactValidation *ContactValidationResult,
	domainAnalysis *DomainAnalysisResult) float64 {

	weightedSum := nameMatch.Confidence*cce.weights["business_name"] +
		addressValidation.Confidence*cce.weights["address"] +
		contactValidation.Confidence*cce.weights["contact"] +
		domainAnalysis.Confidence*cce.weights["domain"]

	// Normalize to 0-1 range
	if weightedSum > 1.0 {
		weightedSum = 1.0
	}

	return weightedSum
}
