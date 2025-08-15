package webanalysis

import (
	"context"
	"fmt"
	"strings"
)

// ConnectionValidatorInterface defines the interface for connection validation
type ConnectionValidatorInterface interface {
	ValidateConnection(ctx context.Context, websiteData *WebsiteAnalysis, req *ClassificationRequest) (*ConnectionValidation, error)
}

// ConnectionValidatorAdapter adapts the new ConnectionValidator to the existing interface
type ConnectionValidatorAdapter struct {
	validator *ConnectionValidator
}

// NewConnectionValidatorAdapter creates a new connection validator adapter
func NewConnectionValidatorAdapter() *ConnectionValidatorAdapter {
	return &ConnectionValidatorAdapter{
		validator: NewConnectionValidator(),
	}
}

// ValidateConnection implements the ConnectionValidatorInterface
func (cva *ConnectionValidatorAdapter) ValidateConnection(ctx context.Context, websiteData *WebsiteAnalysis, req *ClassificationRequest) (*ConnectionValidation, error) {
	if websiteData == nil || req == nil {
		return nil, fmt.Errorf("website data and request cannot be nil")
	}

	// Extract business name from the request
	businessName := req.BusinessName
	if businessName == "" {
		return nil, fmt.Errorf("business name is required")
	}

	// Use the website URL from the website data
	websiteURL := websiteData.URL
	if websiteURL == "" {
		return nil, fmt.Errorf("website URL is required")
	}

	// Perform validation using the new connection validator
	result, err := cva.validator.ValidateConnection(businessName, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("connection validation failed: %w", err)
	}

	// Convert the new result to the existing ConnectionValidation format
	connectionValidation := &ConnectionValidation{
		IsConnected:       result.IsConnected,
		Confidence:        result.OverallConfidence,
		Evidence:          cva.formatEvidence(result.Evidence),
		ValidationFactors: cva.convertValidationFactors(result),
		Recommendations:   cva.generateRecommendations(result),
	}

	return connectionValidation, nil
}

// formatEvidence formats the evidence list into a string
func (cva *ConnectionValidatorAdapter) formatEvidence(evidence []ConnectionEvidence) string {
	if len(evidence) == 0 {
		return "No evidence found"
	}

	var evidenceStrings []string
	for _, e := range evidence {
		evidenceStrings = append(evidenceStrings, fmt.Sprintf("%s: %s (confidence: %.2f)", e.Type, e.Description, e.Confidence))
	}

	return fmt.Sprintf("Found %d pieces of evidence: %s", len(evidence), strings.Join(evidenceStrings, "; "))
}

// convertValidationFactors converts the new result to validation factors
func (cva *ConnectionValidatorAdapter) convertValidationFactors(result *ConnectionValidationResult) []ValidationFactor {
	var factors []ValidationFactor

	// Name matching factor
	if result.NameMatchResult != nil {
		factors = append(factors, ValidationFactor{
			Factor:     "Business Name Match",
			Match:      result.NameMatchResult.IsMatch,
			Confidence: result.NameMatchResult.Confidence,
			Details:    fmt.Sprintf("Match type: %s, Similarity: %.2f", result.NameMatchResult.MatchType, result.NameMatchResult.SimilarityScore),
		})
	}

	// Address matching factor
	if result.AddressMatchResult != nil {
		factors = append(factors, ValidationFactor{
			Factor:     "Address Match",
			Match:      result.AddressMatchResult.IsMatch,
			Confidence: result.AddressMatchResult.Confidence,
			Details:    fmt.Sprintf("Similarity: %.2f, Distance: %.2f km", result.AddressMatchResult.SimilarityScore, result.AddressMatchResult.Distance),
		})
	}

	// Contact matching factor
	if result.ContactMatchResult != nil {
		factors = append(factors, ValidationFactor{
			Factor:     "Contact Information Match",
			Match:      result.ContactMatchResult.IsMatch,
			Confidence: result.ContactMatchResult.OverallConfidence,
			Details:    fmt.Sprintf("Phone match: %t, Email match: %t", result.ContactMatchResult.PhoneMatch, result.ContactMatchResult.EmailMatch),
		})
	}

	// Registration matching factor
	if result.RegistrationMatchResult != nil {
		factors = append(factors, ValidationFactor{
			Factor:     "Business Registration Match",
			Match:      result.RegistrationMatchResult.IsMatch,
			Confidence: result.RegistrationMatchResult.Confidence,
			Details:    fmt.Sprintf("Registration found: %t, Sources: %s", result.RegistrationMatchResult.RegistrationFound, strings.Join(result.RegistrationMatchResult.Sources, ", ")),
		})
	}

	// Ownership evidence factor
	if result.OwnershipScore > 0 {
		factors = append(factors, ValidationFactor{
			Factor:     "Ownership Evidence",
			Match:      result.OwnershipScore >= 0.7,
			Confidence: result.OwnershipScore,
			Details:    fmt.Sprintf("Ownership score: %.2f", result.OwnershipScore),
		})
	}

	return factors
}

// generateRecommendations generates recommendations based on the validation result
func (cva *ConnectionValidatorAdapter) generateRecommendations(result *ConnectionValidationResult) []string {
	var recommendations []string

	// Overall connection strength recommendations
	switch result.ConnectionStrength {
	case "strong":
		recommendations = append(recommendations, "Strong connection established between business and website")
	case "moderate":
		recommendations = append(recommendations, "Moderate connection found, consider additional verification")
	case "weak":
		recommendations = append(recommendations, "Weak connection detected, manual verification recommended")
	case "none":
		recommendations = append(recommendations, "No connection found, verify business-website relationship")
	}

	// Specific recommendations based on individual factors
	if result.NameMatchResult != nil && !result.NameMatchResult.IsMatch {
		recommendations = append(recommendations, "Business name does not match website content")
	}

	if result.AddressMatchResult != nil && !result.AddressMatchResult.IsMatch {
		recommendations = append(recommendations, "Address information does not match between business and website")
	}

	if result.ContactMatchResult != nil && !result.ContactMatchResult.IsMatch {
		recommendations = append(recommendations, "Contact information does not match between business and website")
	}

	if result.RegistrationMatchResult != nil && !result.RegistrationMatchResult.IsMatch {
		recommendations = append(recommendations, "Business registration data does not match website information")
	}

	if result.OwnershipScore < 0.5 {
		recommendations = append(recommendations, "Limited ownership evidence found, verify website ownership")
	}

	// Add warnings as recommendations
	for _, warning := range result.Warnings {
		recommendations = append(recommendations, fmt.Sprintf("Warning: %s", warning))
	}

	return recommendations
}
