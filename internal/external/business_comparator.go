package external

import (
	"context"
	"fmt"
	"math"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

// BusinessComparator handles comparison of business information
type BusinessComparator struct {
	logger *zap.Logger
	config *ComparisonConfig
}

// ComparisonConfig holds configuration for business comparison
type ComparisonConfig struct {
	// Fuzzy matching settings
	MinSimilarityThreshold float64
	MaxEditDistance        int

	// Contact validation settings
	PhoneValidationEnabled   bool
	EmailValidationEnabled   bool
	AddressValidationEnabled bool

	// Geographic matching settings
	MaxDistanceKm      float64
	LocationFuzzyMatch bool

	// Confidence scoring weights
	Weights *ComparisonWeights
}

// ComparisonWeights defines weights for different comparison fields
type ComparisonWeights struct {
	BusinessName    float64
	PhoneNumber     float64
	EmailAddress    float64
	PhysicalAddress float64
	Website         float64
	Industry        float64
}

// ComparisonBusinessInfo represents business information for comparison
type ComparisonBusinessInfo struct {
	Name           string              `json:"name"`
	PhoneNumbers   []string            `json:"phone_numbers"`
	EmailAddresses []string            `json:"email_addresses"`
	Addresses      []ComparisonAddress `json:"addresses"`
	Website        string              `json:"website"`
	Industry       string              `json:"industry"`
	Metadata       map[string]string   `json:"metadata"`
}

// ComparisonAddress represents a physical address for comparison
type ComparisonAddress struct {
	Street     string  `json:"street"`
	City       string  `json:"city"`
	State      string  `json:"state"`
	PostalCode string  `json:"postal_code"`
	Country    string  `json:"country"`
	Latitude   float64 `json:"latitude,omitempty"`
	Longitude  float64 `json:"longitude,omitempty"`
}

// ComparisonResult represents the result of comparing two business information sets
type ComparisonResult struct {
	OverallScore       float64                    `json:"overall_score"`
	ConfidenceLevel    string                     `json:"confidence_level"`
	FieldResults       map[string]FieldComparison `json:"field_results"`
	Recommendations    []string                   `json:"recommendations"`
	VerificationStatus string                     `json:"verification_status"`
}

// FieldComparison represents comparison result for a specific field
type FieldComparison struct {
	Score      float64 `json:"score"`
	Confidence float64 `json:"confidence"`
	Matched    bool    `json:"matched"`
	Reasoning  string  `json:"reasoning"`
	Details    string  `json:"details"`
}

// NewBusinessComparator creates a new business comparator
func NewBusinessComparator(logger *zap.Logger, config *ComparisonConfig) *BusinessComparator {
	if config == nil {
		config = &ComparisonConfig{
			MinSimilarityThreshold:   0.8,
			MaxEditDistance:          3,
			PhoneValidationEnabled:   true,
			EmailValidationEnabled:   true,
			AddressValidationEnabled: true,
			MaxDistanceKm:            50.0,
			LocationFuzzyMatch:       true,
			Weights: &ComparisonWeights{
				BusinessName:    0.3,
				PhoneNumber:     0.25,
				EmailAddress:    0.2,
				PhysicalAddress: 0.15,
				Website:         0.05,
				Industry:        0.05,
			},
		}
	}

	return &BusinessComparator{
		logger: logger,
		config: config,
	}
}

// CompareBusinessInfo compares two business information sets
func (bc *BusinessComparator) CompareBusinessInfo(ctx context.Context, claimed, extracted *ComparisonBusinessInfo) (*ComparisonResult, error) {
	bc.logger.Info("Starting business information comparison",
		zap.String("claimed_name", claimed.Name),
		zap.String("extracted_name", extracted.Name))

	result := &ComparisonResult{
		FieldResults:    make(map[string]FieldComparison),
		Recommendations: []string{},
	}

	// Compare business names
	nameResult := bc.compareBusinessNames(claimed.Name, extracted.Name)
	result.FieldResults["business_name"] = nameResult

	// Compare phone numbers
	phoneResult := bc.comparePhoneNumbers(claimed.PhoneNumbers, extracted.PhoneNumbers)
	result.FieldResults["phone_numbers"] = phoneResult

	// Compare email addresses
	emailResult := bc.compareEmailAddresses(claimed.EmailAddresses, extracted.EmailAddresses)
	result.FieldResults["email_addresses"] = emailResult

	// Compare addresses
	addressResult := bc.compareAddresses(claimed.Addresses, extracted.Addresses)
	result.FieldResults["addresses"] = addressResult

	// Compare websites
	websiteResult := bc.compareWebsites(claimed.Website, extracted.Website)
	result.FieldResults["website"] = websiteResult

	// Compare industries
	industryResult := bc.compareIndustries(claimed.Industry, extracted.Industry)
	result.FieldResults["industry"] = industryResult

	// Calculate overall score
	result.OverallScore = bc.calculateOverallScore(result.FieldResults)
	result.ConfidenceLevel = bc.determineConfidenceLevel(result.OverallScore)
	result.VerificationStatus = bc.determineVerificationStatus(result.OverallScore, result.FieldResults)

	// Generate recommendations
	result.Recommendations = bc.generateRecommendations(result)

	bc.logger.Info("Business information comparison completed",
		zap.Float64("overall_score", result.OverallScore),
		zap.String("confidence_level", result.ConfidenceLevel),
		zap.String("verification_status", result.VerificationStatus))

	return result, nil
}

// compareBusinessNames compares business names using fuzzy matching
func (bc *BusinessComparator) compareBusinessNames(claimed, extracted string) FieldComparison {
	if claimed == "" || extracted == "" {
		return FieldComparison{
			Score:      0.0,
			Confidence: 0.0,
			Matched:    false,
			Reasoning:  "Missing business name for comparison",
		}
	}

	// Normalize names
	claimedNorm := bc.normalizeBusinessName(claimed)
	extractedNorm := bc.normalizeBusinessName(extracted)

	// Calculate similarity
	similarity := bc.calculateStringSimilarity(claimedNorm, extractedNorm)

	// Determine match
	matched := similarity >= bc.config.MinSimilarityThreshold

	reasoning := fmt.Sprintf("Similarity: %.2f (threshold: %.2f)", similarity, bc.config.MinSimilarityThreshold)
	if matched {
		reasoning += " - Names match"
	} else {
		reasoning += " - Names do not match"
	}

	return FieldComparison{
		Score:      similarity,
		Confidence: similarity,
		Matched:    matched,
		Reasoning:  reasoning,
		Details:    fmt.Sprintf("Claimed: '%s', Extracted: '%s'", claimed, extracted),
	}
}

// comparePhoneNumbers compares phone number lists
func (bc *BusinessComparator) comparePhoneNumbers(claimed, extracted []string) FieldComparison {
	if len(claimed) == 0 || len(extracted) == 0 {
		return FieldComparison{
			Score:      0.0,
			Confidence: 0.0,
			Matched:    false,
			Reasoning:  "Missing phone numbers for comparison",
		}
	}

	// Normalize phone numbers
	claimedNorm := bc.normalizePhoneNumbers(claimed)
	extractedNorm := bc.normalizePhoneNumbers(extracted)

	// Find best matches
	bestScore := 0.0
	matchedCount := 0

	for _, claimedPhone := range claimedNorm {
		for _, extractedPhone := range extractedNorm {
			similarity := bc.calculateStringSimilarity(claimedPhone, extractedPhone)
			if similarity > bestScore {
				bestScore = similarity
			}
			if similarity >= bc.config.MinSimilarityThreshold {
				matchedCount++
			}
		}
	}

	// Calculate overall score
	totalComparisons := len(claimedNorm) * len(extractedNorm)
	matchRatio := float64(matchedCount) / float64(totalComparisons)
	overallScore := (bestScore + matchRatio) / 2.0

	matched := overallScore >= bc.config.MinSimilarityThreshold

	reasoning := fmt.Sprintf("Best match: %.2f, Match ratio: %.2f", bestScore, matchRatio)
	if matched {
		reasoning += " - Phone numbers match"
	} else {
		reasoning += " - Phone numbers do not match"
	}

	return FieldComparison{
		Score:      overallScore,
		Confidence: overallScore,
		Matched:    matched,
		Reasoning:  reasoning,
		Details:    fmt.Sprintf("Claimed: %v, Extracted: %v", claimed, extracted),
	}
}

// compareEmailAddresses compares email address lists
func (bc *BusinessComparator) compareEmailAddresses(claimed, extracted []string) FieldComparison {
	if len(claimed) == 0 || len(extracted) == 0 {
		return FieldComparison{
			Score:      0.0,
			Confidence: 0.0,
			Matched:    false,
			Reasoning:  "Missing email addresses for comparison",
		}
	}

	// Normalize email addresses
	claimedNorm := bc.normalizeEmailAddresses(claimed)
	extractedNorm := bc.normalizeEmailAddresses(extracted)

	// Find exact matches first
	exactMatches := 0
	for _, claimedEmail := range claimedNorm {
		for _, extractedEmail := range extractedNorm {
			if claimedEmail == extractedEmail {
				exactMatches++
			}
		}
	}

	// Calculate score
	totalComparisons := len(claimedNorm) * len(extractedNorm)
	exactMatchRatio := float64(exactMatches) / float64(totalComparisons)

	// If no exact matches, try fuzzy matching
	if exactMatches == 0 {
		bestFuzzyScore := 0.0
		for _, claimedEmail := range claimedNorm {
			for _, extractedEmail := range extractedNorm {
				similarity := bc.calculateStringSimilarity(claimedEmail, extractedEmail)
				if similarity > bestFuzzyScore {
					bestFuzzyScore = similarity
				}
			}
		}
		exactMatchRatio = bestFuzzyScore * 0.8 // Penalize fuzzy matches
	}

	matched := exactMatchRatio >= bc.config.MinSimilarityThreshold

	reasoning := fmt.Sprintf("Exact match ratio: %.2f", exactMatchRatio)
	if matched {
		reasoning += " - Email addresses match"
	} else {
		reasoning += " - Email addresses do not match"
	}

	return FieldComparison{
		Score:      exactMatchRatio,
		Confidence: exactMatchRatio,
		Matched:    matched,
		Reasoning:  reasoning,
		Details:    fmt.Sprintf("Claimed: %v, Extracted: %v", claimed, extracted),
	}
}

// compareAddresses compares address lists
func (bc *BusinessComparator) compareAddresses(claimed, extracted []ComparisonAddress) FieldComparison {
	if len(claimed) == 0 || len(extracted) == 0 {
		return FieldComparison{
			Score:      0.0,
			Confidence: 0.0,
			Matched:    false,
			Reasoning:  "Missing addresses for comparison",
		}
	}

	bestScore := 0.0
	matchedCount := 0

	for _, claimedAddr := range claimed {
		for _, extractedAddr := range extracted {
			score := bc.compareSingleAddress(claimedAddr, extractedAddr)
			if score > bestScore {
				bestScore = score
			}
			if score >= bc.config.MinSimilarityThreshold {
				matchedCount++
			}
		}
	}

	// Calculate overall score
	totalComparisons := len(claimed) * len(extracted)
	matchRatio := float64(matchedCount) / float64(totalComparisons)
	overallScore := (bestScore + matchRatio) / 2.0

	matched := overallScore >= bc.config.MinSimilarityThreshold

	reasoning := fmt.Sprintf("Best address match: %.2f, Match ratio: %.2f", bestScore, matchRatio)
	if matched {
		reasoning += " - Addresses match"
	} else {
		reasoning += " - Addresses do not match"
	}

	return FieldComparison{
		Score:      overallScore,
		Confidence: overallScore,
		Matched:    matched,
		Reasoning:  reasoning,
		Details:    fmt.Sprintf("Claimed addresses: %d, Extracted addresses: %d", len(claimed), len(extracted)),
	}
}

// compareSingleAddress compares two individual addresses
func (bc *BusinessComparator) compareSingleAddress(claimed, extracted ComparisonAddress) float64 {
	scores := make([]float64, 0)

	// Compare street
	if claimed.Street != "" && extracted.Street != "" {
		streetScore := bc.calculateStringSimilarity(
			bc.normalizeAddress(claimed.Street),
			bc.normalizeAddress(extracted.Street),
		)
		scores = append(scores, streetScore)
	}

	// Compare city
	if claimed.City != "" && extracted.City != "" {
		cityScore := bc.calculateStringSimilarity(
			bc.normalizeAddress(claimed.City),
			bc.normalizeAddress(extracted.City),
		)
		scores = append(scores, cityScore)
	}

	// Compare state
	if claimed.State != "" && extracted.State != "" {
		stateScore := bc.calculateStringSimilarity(
			bc.normalizeAddress(claimed.State),
			bc.normalizeAddress(extracted.State),
		)
		scores = append(scores, stateScore)
	}

	// Compare postal code
	if claimed.PostalCode != "" && extracted.PostalCode != "" {
		postalScore := bc.calculateStringSimilarity(
			bc.normalizeAddress(claimed.PostalCode),
			bc.normalizeAddress(extracted.PostalCode),
		)
		scores = append(scores, postalScore)
	}

	// Compare geographic coordinates if available
	if claimed.Latitude != 0 && claimed.Longitude != 0 &&
		extracted.Latitude != 0 && extracted.Longitude != 0 {
		distance := bc.calculateDistance(claimed.Latitude, claimed.Longitude,
			extracted.Latitude, extracted.Longitude)
		if distance <= bc.config.MaxDistanceKm {
			geoScore := 1.0 - (distance / bc.config.MaxDistanceKm)
			scores = append(scores, geoScore)
		} else {
			scores = append(scores, 0.0)
		}
	}

	if len(scores) == 0 {
		return 0.0
	}

	// Calculate weighted average
	totalScore := 0.0
	for _, score := range scores {
		totalScore += score
	}

	return totalScore / float64(len(scores))
}

// compareWebsites compares website URLs
func (bc *BusinessComparator) compareWebsites(claimed, extracted string) FieldComparison {
	if claimed == "" || extracted == "" {
		return FieldComparison{
			Score:      0.0,
			Confidence: 0.0,
			Matched:    false,
			Reasoning:  "Missing website for comparison",
		}
	}

	// Normalize URLs
	claimedNorm := bc.normalizeURL(claimed)
	extractedNorm := bc.normalizeURL(extracted)

	// Check for exact match
	if claimedNorm == extractedNorm {
		return FieldComparison{
			Score:      1.0,
			Confidence: 1.0,
			Matched:    true,
			Reasoning:  "Exact website match",
			Details:    fmt.Sprintf("Claimed: %s, Extracted: %s", claimed, extracted),
		}
	}

	// Check for domain match
	claimedDomain := bc.extractDomain(claimedNorm)
	extractedDomain := bc.extractDomain(extractedNorm)

	if claimedDomain == extractedDomain {
		return FieldComparison{
			Score:      0.9,
			Confidence: 0.9,
			Matched:    true,
			Reasoning:  "Domain match (different paths)",
			Details:    fmt.Sprintf("Claimed: %s, Extracted: %s", claimed, extracted),
		}
	}

	// Fuzzy match
	similarity := bc.calculateStringSimilarity(claimedNorm, extractedNorm)
	matched := similarity >= bc.config.MinSimilarityThreshold

	reasoning := fmt.Sprintf("Website similarity: %.2f", similarity)
	if matched {
		reasoning += " - Websites match"
	} else {
		reasoning += " - Websites do not match"
	}

	return FieldComparison{
		Score:      similarity,
		Confidence: similarity,
		Matched:    matched,
		Reasoning:  reasoning,
		Details:    fmt.Sprintf("Claimed: %s, Extracted: %s", claimed, extracted),
	}
}

// compareIndustries compares industry classifications
func (bc *BusinessComparator) compareIndustries(claimed, extracted string) FieldComparison {
	if claimed == "" || extracted == "" {
		return FieldComparison{
			Score:      0.0,
			Confidence: 0.0,
			Matched:    false,
			Reasoning:  "Missing industry for comparison",
		}
	}

	// Normalize industries
	claimedNorm := bc.normalizeIndustry(claimed)
	extractedNorm := bc.normalizeIndustry(extracted)

	// Calculate similarity
	similarity := bc.calculateStringSimilarity(claimedNorm, extractedNorm)
	matched := similarity >= bc.config.MinSimilarityThreshold

	reasoning := fmt.Sprintf("Industry similarity: %.2f", similarity)
	if matched {
		reasoning += " - Industries match"
	} else {
		reasoning += " - Industries do not match"
	}

	return FieldComparison{
		Score:      similarity,
		Confidence: similarity,
		Matched:    matched,
		Reasoning:  reasoning,
		Details:    fmt.Sprintf("Claimed: %s, Extracted: %s", claimed, extracted),
	}
}

// calculateOverallScore calculates the weighted overall score
func (bc *BusinessComparator) calculateOverallScore(fieldResults map[string]FieldComparison) float64 {
	totalWeightedScore := 0.0
	totalWeight := 0.0

	// Business name
	if result, exists := fieldResults["business_name"]; exists {
		totalWeightedScore += result.Score * bc.config.Weights.BusinessName
		totalWeight += bc.config.Weights.BusinessName
	}

	// Phone numbers
	if result, exists := fieldResults["phone_numbers"]; exists {
		totalWeightedScore += result.Score * bc.config.Weights.PhoneNumber
		totalWeight += bc.config.Weights.PhoneNumber
	}

	// Email addresses
	if result, exists := fieldResults["email_addresses"]; exists {
		totalWeightedScore += result.Score * bc.config.Weights.EmailAddress
		totalWeight += bc.config.Weights.EmailAddress
	}

	// Addresses
	if result, exists := fieldResults["addresses"]; exists {
		totalWeightedScore += result.Score * bc.config.Weights.PhysicalAddress
		totalWeight += bc.config.Weights.PhysicalAddress
	}

	// Website
	if result, exists := fieldResults["website"]; exists {
		totalWeightedScore += result.Score * bc.config.Weights.Website
		totalWeight += bc.config.Weights.Website
	}

	// Industry
	if result, exists := fieldResults["industry"]; exists {
		totalWeightedScore += result.Score * bc.config.Weights.Industry
		totalWeight += bc.config.Weights.Industry
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalWeightedScore / totalWeight
}

// GetConfig returns the current configuration
func (bc *BusinessComparator) GetConfig() *ComparisonConfig {
	return bc.config
}

// determineConfidenceLevel determines the confidence level based on score
func (bc *BusinessComparator) determineConfidenceLevel(score float64) string {
	switch {
	case score >= 0.9:
		return "high"
	case score >= 0.7:
		return "medium"
	case score >= 0.5:
		return "low"
	default:
		return "very_low"
	}
}

// determineVerificationStatus determines verification status based on score and field results
func (bc *BusinessComparator) determineVerificationStatus(score float64, fieldResults map[string]FieldComparison) string {
	// Count matched fields
	matchedFields := 0
	totalFields := 0

	for _, result := range fieldResults {
		totalFields++
		if result.Matched {
			matchedFields++
		}
	}

	// Determine status based on score and field matches
	switch {
	case score >= 0.9 && matchedFields >= 3:
		return "PASSED"
	case score >= 0.7 && matchedFields >= 2:
		return "PARTIAL"
	case score >= 0.5 && matchedFields >= 1:
		return "PARTIAL"
	case score < 0.3:
		return "FAILED"
	default:
		return "SKIPPED"
	}
}

// generateRecommendations generates recommendations based on comparison results
func (bc *BusinessComparator) generateRecommendations(result *ComparisonResult) []string {
	recommendations := []string{}

	// Check for low confidence fields
	for fieldName, fieldResult := range result.FieldResults {
		if fieldResult.Confidence < 0.5 {
			recommendations = append(recommendations,
				fmt.Sprintf("Manual verification recommended for %s (confidence: %.2f)",
					fieldName, fieldResult.Confidence))
		}
	}

	// Check overall score
	if result.OverallScore < 0.7 {
		recommendations = append(recommendations,
			"Overall verification confidence is low - manual review recommended")
	}

	// Check for missing critical fields
	criticalFields := []string{"business_name", "phone_numbers", "email_addresses"}
	for _, field := range criticalFields {
		if _, exists := result.FieldResults[field]; !exists {
			recommendations = append(recommendations,
				fmt.Sprintf("Critical field '%s' is missing - additional data collection recommended", field))
		}
	}

	return recommendations
}

// Utility functions for string normalization and comparison

func (bc *BusinessComparator) normalizeBusinessName(name string) string {
	// Remove common business suffixes
	suffixes := []string{" inc", " llc", " ltd", " corp", " corporation", " company", " co"}
	normalized := strings.ToLower(strings.TrimSpace(name))

	for _, suffix := range suffixes {
		normalized = strings.TrimSuffix(normalized, suffix)
	}

	// Remove special characters and extra spaces
	normalized = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(normalized, "")
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

func (bc *BusinessComparator) normalizePhoneNumbers(phones []string) []string {
	normalized := make([]string, 0, len(phones))

	for _, phone := range phones {
		// Remove all non-digit characters
		digits := regexp.MustCompile(`[^\d]`).ReplaceAllString(phone, "")

		// Handle different formats
		if len(digits) == 10 {
			// US format: (123) 456-7890 -> 1234567890
			normalized = append(normalized, digits)
		} else if len(digits) == 11 && digits[0] == '1' {
			// US format with country code: +1 (123) 456-7890 -> 1234567890
			normalized = append(normalized, digits[1:])
		} else {
			// Keep as is for international numbers
			normalized = append(normalized, digits)
		}
	}

	return normalized
}

func (bc *BusinessComparator) normalizeEmailAddresses(emails []string) []string {
	normalized := make([]string, 0, len(emails))

	for _, email := range emails {
		normalizedEmail := strings.ToLower(strings.TrimSpace(email))
		normalized = append(normalized, normalizedEmail)
	}

	return normalized
}

func (bc *BusinessComparator) normalizeAddress(addr string) string {
	normalized := strings.ToLower(strings.TrimSpace(addr))

	// Remove common address abbreviations
	abbreviations := map[string]string{
		"street":    "st",
		"avenue":    "ave",
		"boulevard": "blvd",
		"drive":     "dr",
		"road":      "rd",
		"lane":      "ln",
		"court":     "ct",
		"place":     "pl",
		"circle":    "cir",
		"square":    "sq",
	}

	for full, abbr := range abbreviations {
		normalized = strings.ReplaceAll(normalized, full, abbr)
	}

	// Remove special characters and extra spaces
	normalized = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(normalized, "")
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

func (bc *BusinessComparator) normalizeURL(url string) string {
	normalized := strings.ToLower(strings.TrimSpace(url))

	// Remove protocol
	normalized = strings.TrimPrefix(normalized, "http://")
	normalized = strings.TrimPrefix(normalized, "https://")

	// Remove www prefix
	normalized = strings.TrimPrefix(normalized, "www.")

	// Remove trailing slash
	normalized = strings.TrimSuffix(normalized, "/")

	return normalized
}

func (bc *BusinessComparator) extractDomain(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[0]
	}
	return url
}

func (bc *BusinessComparator) normalizeIndustry(industry string) string {
	normalized := strings.ToLower(strings.TrimSpace(industry))

	// Remove special characters and extra spaces
	normalized = regexp.MustCompile(`[^a-z0-9\s]`).ReplaceAllString(normalized, "")
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")

	return strings.TrimSpace(normalized)
}

// calculateStringSimilarity calculates similarity between two strings using Levenshtein distance
func (bc *BusinessComparator) calculateStringSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	if s1 == "" || s2 == "" {
		return 0.0
	}

	// Calculate Levenshtein distance
	distance := bc.levenshteinDistance(s1, s2)
	maxLen := float64(max(len(s1), len(s2)))

	if maxLen == 0 {
		return 1.0
	}

	similarity := 1.0 - (float64(distance) / maxLen)
	return math.Max(0.0, similarity)
}

// levenshteinDistance calculates the Levenshtein distance between two strings
func (bc *BusinessComparator) levenshteinDistance(s1, s2 string) int {
	len1, len2 := len(s1), len(s2)

	// Create matrix
	matrix := make([][]int, len1+1)
	for i := range matrix {
		matrix[i] = make([]int, len2+1)
	}

	// Initialize first row and column
	for i := 0; i <= len1; i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len2; j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len1; i++ {
		for j := 1; j <= len2; j++ {
			if s1[i-1] == s2[j-1] {
				matrix[i][j] = matrix[i-1][j-1]
			} else {
				matrix[i][j] = min(
					matrix[i-1][j]+1, // deletion
					min(
						matrix[i][j-1]+1,   // insertion
						matrix[i-1][j-1]+1, // substitution
					),
				)
			}
		}
	}

	return matrix[len1][len2]
}

// calculateDistance calculates distance between two geographic coordinates in kilometers
func (bc *BusinessComparator) calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371.0 // Earth's radius in kilometers

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dlon/2)*math.Sin(dlon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// Utility functions
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
