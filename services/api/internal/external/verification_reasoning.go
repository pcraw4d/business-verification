package external

import (
	"fmt"
	"strings"
	"time"
)

// VerificationReasoning provides detailed explanations for verification results
type VerificationReasoning struct {
	Status          string           `json:"status"`
	OverallScore    float64          `json:"overall_score"`
	ConfidenceLevel string           `json:"confidence_level"`
	Explanation     string           `json:"explanation"`
	FieldAnalysis   []FieldAnalysis  `json:"field_analysis"`
	Recommendations []Recommendation `json:"recommendations"`
	RiskFactors     []RiskFactor     `json:"risk_factors"`
	GeneratedAt     time.Time        `json:"generated_at"`
	VerificationID  string           `json:"verification_id"`
	BusinessName    string           `json:"business_name"`
	WebsiteURL      string           `json:"website_url"`
}

// FieldAnalysis provides detailed analysis of each verification field
type FieldAnalysis struct {
	FieldName    string  `json:"field_name"`
	Score        float64 `json:"score"`
	Status       string  `json:"status"`
	Explanation  string  `json:"explanation"`
	Evidence     string  `json:"evidence"`
	Confidence   float64 `json:"confidence"`
	Weight       float64 `json:"weight"`
	Contribution float64 `json:"contribution"`
}

// Recommendation provides actionable recommendations for verification
type Recommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"` // high, medium, low
	Description string `json:"description"`
	Action      string `json:"action"`
	Reason      string `json:"reason"`
	Impact      string `json:"impact"`
}

// RiskFactor identifies potential risks in the verification process
type RiskFactor struct {
	Factor      string  `json:"factor"`
	Severity    string  `json:"severity"` // low, medium, high, critical
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Mitigation  string  `json:"mitigation"`
	Probability float64 `json:"probability"`
}

// VerificationReport contains the complete verification report
type VerificationReport struct {
	ReportID          string                 `json:"report_id"`
	VerificationID    string                 `json:"verification_id"`
	BusinessName      string                 `json:"business_name"`
	WebsiteURL        string                 `json:"website_url"`
	Status            string                 `json:"status"`
	OverallScore      float64                `json:"overall_score"`
	ConfidenceLevel   string                 `json:"confidence_level"`
	GeneratedAt       time.Time              `json:"generated_at"`
	CompletedAt       time.Time              `json:"completed_at"`
	Duration          time.Duration          `json:"duration"`
	Reasoning         *VerificationReasoning `json:"reasoning"`
	ComparisonDetails *ComparisonDetails     `json:"comparison_details"`
	AuditTrail        []AuditEvent           `json:"audit_trail"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// ComparisonDetails contains detailed comparison information
type ComparisonDetails struct {
	BusinessName   ComparisonField `json:"business_name"`
	PhoneNumbers   ComparisonField `json:"phone_numbers"`
	EmailAddresses ComparisonField `json:"email_addresses"`
	Addresses      ComparisonField `json:"addresses"`
	WebsiteURL     ComparisonField `json:"website_url"`
	Industry       ComparisonField `json:"industry"`
	FoundedYear    ComparisonField `json:"founded_year"`
	EmployeeCount  ComparisonField `json:"employee_count"`
	Revenue        ComparisonField `json:"revenue"`
	Services       ComparisonField `json:"services"`
}

// ComparisonField contains detailed comparison for a specific field
type ComparisonField struct {
	FieldName      string      `json:"field_name"`
	InputValue     interface{} `json:"input_value"`
	ExtractedValue interface{} `json:"extracted_value"`
	Score          float64     `json:"score"`
	Status         string      `json:"status"`
	Confidence     float64     `json:"confidence"`
	Algorithm      string      `json:"algorithm"`
	Threshold      float64     `json:"threshold"`
	Matched        bool        `json:"matched"`
	Reasoning      string      `json:"reasoning"`
	Evidence       string      `json:"evidence"`
}

// AuditEvent represents an event in the verification audit trail
type AuditEvent struct {
	EventID     string                 `json:"event_id"`
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Description string                 `json:"description"`
	UserID      string                 `json:"user_id,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Severity    string                 `json:"severity"`
}

// VerificationReasoningGenerator generates detailed reasoning for verification results
type VerificationReasoningGenerator struct {
	config *VerificationReasoningConfig
}

// VerificationReasoningConfig contains configuration for reasoning generation
type VerificationReasoningConfig struct {
	EnableDetailedExplanations bool    `json:"enable_detailed_explanations"`
	EnableRiskAnalysis         bool    `json:"enable_risk_analysis"`
	EnableRecommendations      bool    `json:"enable_recommendations"`
	EnableAuditTrail           bool    `json:"enable_audit_trail"`
	MinConfidenceThreshold     float64 `json:"min_confidence_threshold"`
	MaxRiskProbability         float64 `json:"max_risk_probability"`
	Language                   string  `json:"language"`
}

// NewVerificationReasoningGenerator creates a new reasoning generator
func NewVerificationReasoningGenerator(config *VerificationReasoningConfig) *VerificationReasoningGenerator {
	if config == nil {
		config = &VerificationReasoningConfig{
			EnableDetailedExplanations: true,
			EnableRiskAnalysis:         true,
			EnableRecommendations:      true,
			EnableAuditTrail:           true,
			MinConfidenceThreshold:     0.6,
			MaxRiskProbability:         0.8,
			Language:                   "en",
		}
	}
	return &VerificationReasoningGenerator{config: config}
}

// GenerateReasoning generates detailed reasoning for verification results
func (g *VerificationReasoningGenerator) GenerateReasoning(
	verificationID string,
	businessName string,
	websiteURL string,
	result *VerificationResult,
	comparison *ComparisonResult,
) (*VerificationReasoning, error) {
	if result == nil {
		return nil, fmt.Errorf("verification result cannot be nil")
	}

	reasoning := &VerificationReasoning{
		Status:          string(result.Status),
		OverallScore:    result.OverallScore,
		ConfidenceLevel: g.calculateConfidenceLevel(result.OverallScore),
		Explanation:     g.generateOverallExplanation(result, comparison),
		FieldAnalysis:   g.generateFieldAnalysis(comparison),
		Recommendations: g.generateRecommendations(result, comparison),
		RiskFactors:     g.generateRiskFactors(result, comparison),
		GeneratedAt:     time.Now(),
		VerificationID:  verificationID,
		BusinessName:    businessName,
		WebsiteURL:      websiteURL,
	}

	return reasoning, nil
}

// calculateConfidenceLevel determines the confidence level based on score
func (g *VerificationReasoningGenerator) calculateConfidenceLevel(score float64) string {
	switch {
	case score >= 0.8:
		return "high"
	case score >= 0.6:
		return "medium"
	case score >= 0.4:
		return "low"
	default:
		return "very_low"
	}
}

// generateOverallExplanation creates the main explanation for verification results
func (g *VerificationReasoningGenerator) generateOverallExplanation(
	result *VerificationResult,
	comparison *ComparisonResult,
) string {
	var explanation strings.Builder

	// Status explanation
	switch result.Status {
	case "PASSED":
		explanation.WriteString("Verification PASSED with high confidence. ")
		explanation.WriteString(fmt.Sprintf("Overall score: %.2f (%.0f%%). ", result.OverallScore, result.OverallScore*100))
		explanation.WriteString("All critical fields matched successfully with strong evidence.")
	case "PARTIAL":
		explanation.WriteString("Verification PARTIALLY PASSED with moderate confidence. ")
		explanation.WriteString(fmt.Sprintf("Overall score: %.2f (%.0f%%). ", result.OverallScore, result.OverallScore*100))
		explanation.WriteString("Some fields matched while others require manual review.")
	case "FAILED":
		explanation.WriteString("Verification FAILED with low confidence. ")
		explanation.WriteString(fmt.Sprintf("Overall score: %.2f (%.0f%%). ", result.OverallScore, result.OverallScore*100))
		explanation.WriteString("Multiple fields failed to match or had insufficient evidence.")
	case "SKIPPED":
		explanation.WriteString("Verification SKIPPED due to insufficient data or technical issues. ")
		explanation.WriteString("Manual verification recommended.")
	default:
		explanation.WriteString("Verification status unclear. Manual review required.")
	}

	// Add field-specific insights
	if comparison != nil {
		explanation.WriteString(" ")
		explanation.WriteString(g.generateFieldInsights(comparison))
	}

	return explanation.String()
}

// generateFieldInsights provides insights about individual field performance
func (g *VerificationReasoningGenerator) generateFieldInsights(comparison *ComparisonResult) string {
	var insights strings.Builder
	strongMatches := 0
	weakMatches := 0
	failedMatches := 0

	for _, field := range comparison.FieldResults {
		switch {
		case field.Score >= 0.8:
			strongMatches++
		case field.Score >= 0.6:
			weakMatches++
		default:
			failedMatches++
		}
	}

	if strongMatches > 0 {
		insights.WriteString(fmt.Sprintf("%d fields matched strongly, ", strongMatches))
	}
	if weakMatches > 0 {
		insights.WriteString(fmt.Sprintf("%d fields matched partially, ", weakMatches))
	}
	if failedMatches > 0 {
		insights.WriteString(fmt.Sprintf("%d fields failed to match. ", failedMatches))
	}

	return insights.String()
}

// generateFieldAnalysis creates detailed analysis for each verification field
func (g *VerificationReasoningGenerator) generateFieldAnalysis(
	comparison *ComparisonResult,
) []FieldAnalysis {
	if comparison == nil {
		return []FieldAnalysis{}
	}

	var analysis []FieldAnalysis
	for fieldName, field := range comparison.FieldResults {
		fieldAnalysis := FieldAnalysis{
			FieldName:    fieldName,
			Score:        field.Score,
			Status:       g.getFieldStatus(field.Score),
			Explanation:  g.generateFieldExplanation(fieldName, field),
			Evidence:     g.generateFieldEvidence(field),
			Confidence:   field.Confidence,
			Weight:       0.0,               // Weight not available in FieldComparison
			Contribution: field.Score * 0.0, // Weight not available
		}
		analysis = append(analysis, fieldAnalysis)
	}

	return analysis
}

// getFieldStatus determines the status of a field based on its score
func (g *VerificationReasoningGenerator) getFieldStatus(score float64) string {
	switch {
	case score >= 0.8:
		return "passed"
	case score >= 0.6:
		return "partial"
	default:
		return "failed"
	}
}

// generateFieldExplanation creates explanation for a specific field
func (g *VerificationReasoningGenerator) generateFieldExplanation(fieldName string, field FieldComparison) string {
	var explanation strings.Builder

	switch fieldName {
	case "business_name":
		explanation.WriteString("Business name comparison using fuzzy string matching. ")
		if field.Score >= 0.8 {
			explanation.WriteString("Names matched with high similarity.")
		} else if field.Score >= 0.6 {
			explanation.WriteString("Names showed moderate similarity with minor differences.")
		} else {
			explanation.WriteString("Names showed significant differences or no match found.")
		}
	case "phone_numbers":
		explanation.WriteString("Phone number comparison using normalized format matching. ")
		if field.Score >= 0.8 {
			explanation.WriteString("Phone numbers matched exactly or with minor formatting differences.")
		} else if field.Score >= 0.6 {
			explanation.WriteString("Phone numbers showed partial match or similar patterns.")
		} else {
			explanation.WriteString("Phone numbers did not match or were not found.")
		}
	case "email_addresses":
		explanation.WriteString("Email address comparison using domain and local part analysis. ")
		if field.Score >= 0.8 {
			explanation.WriteString("Email addresses matched or used same domain.")
		} else if field.Score >= 0.6 {
			explanation.WriteString("Email addresses showed partial match or similar patterns.")
		} else {
			explanation.WriteString("Email addresses did not match or were not found.")
		}
	case "addresses":
		explanation.WriteString("Address comparison using geographic and component matching. ")
		if field.Score >= 0.8 {
			explanation.WriteString("Addresses matched with high geographic similarity.")
		} else if field.Score >= 0.6 {
			explanation.WriteString("Addresses showed partial geographic match.")
		} else {
			explanation.WriteString("Addresses did not match or geographic data insufficient.")
		}
	default:
		explanation.WriteString(fmt.Sprintf("%s field comparison. ", fieldName))
		if field.Score >= 0.8 {
			explanation.WriteString("Field matched successfully.")
		} else if field.Score >= 0.6 {
			explanation.WriteString("Field showed partial match.")
		} else {
			explanation.WriteString("Field failed to match.")
		}
	}

	return explanation.String()
}

// generateFieldEvidence provides evidence for field comparison results
func (g *VerificationReasoningGenerator) generateFieldEvidence(field FieldComparison) string {
	var evidence strings.Builder

	evidence.WriteString(fmt.Sprintf("Score: %.2f, Confidence: %.2f. ",
		field.Score, field.Confidence))

	if field.Score >= 0.8 {
		evidence.WriteString("Strong evidence of match.")
	} else if field.Score >= 0.6 {
		evidence.WriteString("Moderate evidence of match.")
	} else {
		evidence.WriteString("Weak or no evidence of match.")
	}

	return evidence.String()
}

// generateRecommendations creates comprehensive actionable recommendations for manual verification
func (g *VerificationReasoningGenerator) generateRecommendations(
	result *VerificationResult,
	comparison *ComparisonResult,
) []Recommendation {
	var recommendations []Recommendation

	// Overall status recommendations with enhanced logic
	switch result.Status {
	case StatusPassed:
		if result.OverallScore >= 0.95 {
			recommendations = append(recommendations, Recommendation{
				Type:        "verification_complete",
				Priority:    "low",
				Description: "Verification passed with high confidence",
				Action:      "No action required - proceed with onboarding",
				Reason:      "All verification criteria exceeded expectations",
				Impact:      "Positive - business ownership verified with high confidence",
			})
		} else if result.OverallScore >= 0.85 {
			recommendations = append(recommendations, Recommendation{
				Type:        "verification_passed",
				Priority:    "low",
				Description: "Verification passed successfully",
				Action:      "Proceed with standard onboarding process",
				Reason:      "All critical verification criteria met",
				Impact:      "Positive - business ownership verified",
			})
		} else {
			// Passed but with lower confidence - recommend monitoring
			recommendations = append(recommendations, Recommendation{
				Type:        "monitoring",
				Priority:    "low",
				Description: "Monitor for future verification changes",
				Action:      "Schedule periodic re-verification checks",
				Reason:      "Verification passed but with moderate confidence score",
				Impact:      "Low - may need future validation",
			})
		}
	case StatusPartial:
		recommendations = append(recommendations, Recommendation{
			Type:        "manual_review",
			Priority:    "medium",
			Description: "Manual verification recommended for failed fields",
			Action:      "Review and verify failed fields through alternative sources",
			Reason:      "Some critical fields failed automated verification",
			Impact:      "Medium - manual intervention required before approval",
		})

		// Add specific recommendations for failed fields
		recommendations = append(recommendations, g.generateFieldSpecificRecommendations(comparison)...)

	case StatusFailed:
		recommendations = append(recommendations, Recommendation{
			Type:        "investigation",
			Priority:    "high",
			Description: "Comprehensive investigation required",
			Action:      "Contact business directly for documentation and clarification",
			Reason:      "Multiple critical verification failures detected",
			Impact:      "High - potential fraud risk or significant data discrepancies",
		})

		// Add detailed investigation steps
		recommendations = append(recommendations, Recommendation{
			Type:        "documentation_request",
			Priority:    "high",
			Description: "Request additional business documentation",
			Action:      "Request business registration documents, tax records, and contact verification",
			Reason:      "Automated verification failed - manual documentation required",
			Impact:      "High - essential for risk mitigation",
		})
	case StatusSkipped:
		recommendations = append(recommendations, Recommendation{
			Type:        "retry_verification",
			Priority:    "medium",
			Description: "Retry verification with different approach",
			Action:      "Attempt verification using alternative data sources or methods",
			Reason:      "Initial verification was skipped due to technical issues",
			Impact:      "Medium - verification incomplete",
		})
	}

	// Add confidence-based recommendations
	if result.OverallScore < 0.7 {
		recommendations = append(recommendations, Recommendation{
			Type:        "confidence_review",
			Priority:    "medium",
			Description: "Low confidence score requires additional validation",
			Action:      "Gather additional data points and perform secondary verification",
			Reason:      fmt.Sprintf("Verification confidence score is %.2f (below 0.7 threshold)", result.OverallScore),
			Impact:      "Medium - affects decision confidence",
		})
	}

	// Risk-based recommendations
	if result.OverallScore < 0.5 {
		recommendations = append(recommendations, Recommendation{
			Type:        "risk_assessment",
			Priority:    "high",
			Description: "High-risk verification requires enhanced due diligence",
			Action:      "Escalate to senior review team and implement enhanced monitoring",
			Reason:      "Verification score indicates significant risk factors",
			Impact:      "High - potential compliance and fraud risk",
		})
	}

	// Data quality recommendations
	recommendations = append(recommendations, g.generateDataQualityRecommendations(comparison)...)

	return recommendations
}

// generateFieldSpecificRecommendations creates targeted recommendations for specific field failures
func (g *VerificationReasoningGenerator) generateFieldSpecificRecommendations(comparison *ComparisonResult) []Recommendation {
	var recommendations []Recommendation

	if comparison == nil {
		return recommendations
	}

	for fieldName, field := range comparison.FieldResults {
		if field.Score < 0.6 {
			recommendation := g.createFieldRecommendation(fieldName, field.Score, field.Confidence)
			if recommendation != nil {
				recommendations = append(recommendations, *recommendation)
			}
		}
	}

	return recommendations
}

// createFieldRecommendation creates specific recommendations based on field type and scores
func (g *VerificationReasoningGenerator) createFieldRecommendation(fieldName string, score, confidence float64) *Recommendation {
	priority := "medium"
	if score < 0.3 {
		priority = "high"
	} else if score > 0.5 {
		priority = "low"
	}

	switch fieldName {
	case "business_name":
		return &Recommendation{
			Type:        "business_name_verification",
			Priority:    priority,
			Description: "Verify business name through official sources",
			Action:      "Check business registration records, trade name filings, and official documents",
			Reason:      fmt.Sprintf("Business name matching scored %.2f (confidence: %.2f)", score, confidence),
			Impact:      "High - business name is critical for identity verification",
		}
	case "phone_numbers":
		return &Recommendation{
			Type:        "contact_verification",
			Priority:    priority,
			Description: "Verify phone number through direct contact",
			Action:      "Call the provided phone number to confirm business contact details",
			Reason:      fmt.Sprintf("Phone number verification scored %.2f (confidence: %.2f)", score, confidence),
			Impact:      "Medium - affects communication and contact validation",
		}
	case "email_addresses":
		return &Recommendation{
			Type:        "email_verification",
			Priority:    priority,
			Description: "Verify email address through direct communication",
			Action:      "Send verification email and confirm domain ownership",
			Reason:      fmt.Sprintf("Email verification scored %.2f (confidence: %.2f)", score, confidence),
			Impact:      "Medium - important for digital communication validation",
		}
	case "addresses":
		return &Recommendation{
			Type:        "address_verification",
			Priority:    priority,
			Description: "Verify physical address through multiple sources",
			Action:      "Check postal services, mapping services, and physical location validation",
			Reason:      fmt.Sprintf("Address matching scored %.2f (confidence: %.2f)", score, confidence),
			Impact:      "High - physical presence is crucial for business verification",
		}
	case "website_url":
		return &Recommendation{
			Type:        "website_verification",
			Priority:    priority,
			Description: "Verify website ownership and control",
			Action:      "Check domain registration, SSL certificates, and website administrative access",
			Reason:      fmt.Sprintf("Website verification scored %.2f (confidence: %.2f)", score, confidence),
			Impact:      "High - website ownership is core to verification claim",
		}
	case "industry":
		return &Recommendation{
			Type:        "industry_classification",
			Priority:    "low",
			Description: "Verify business industry classification",
			Action:      "Review business activities, licenses, and industry registrations",
			Reason:      fmt.Sprintf("Industry classification scored %.2f (confidence: %.2f)", score, confidence),
			Impact:      "Low - helps with compliance and risk assessment",
		}
	case "founded_year":
		return &Recommendation{
			Type:        "establishment_verification",
			Priority:    "low",
			Description: "Verify business establishment date",
			Action:      "Check incorporation documents and business registration dates",
			Reason:      fmt.Sprintf("Founded year verification scored %.2f (confidence: %.2f)", score, confidence),
			Impact:      "Low - useful for business history validation",
		}
	default:
		return &Recommendation{
			Type:        "field_verification",
			Priority:    priority,
			Description: fmt.Sprintf("Manually verify %s field", fieldName),
			Action:      fmt.Sprintf("Review and validate %s data through appropriate sources", fieldName),
			Reason:      fmt.Sprintf("%s field scored %.2f (confidence: %.2f)", fieldName, score, confidence),
			Impact:      "Medium - affects overall verification confidence",
		}
	}
}

// generateDataQualityRecommendations creates recommendations for improving data quality
func (g *VerificationReasoningGenerator) generateDataQualityRecommendations(comparison *ComparisonResult) []Recommendation {
	var recommendations []Recommendation

	if comparison == nil {
		return recommendations
	}

	// Count low-quality fields
	lowQualityFields := 0
	totalFields := len(comparison.FieldResults)

	for _, field := range comparison.FieldResults {
		if field.Confidence < 0.6 {
			lowQualityFields++
		}
	}

	// If many fields have low confidence, recommend data source improvement
	if lowQualityFields > totalFields/2 {
		recommendations = append(recommendations, Recommendation{
			Type:        "data_source_improvement",
			Priority:    "medium",
			Description: "Improve data sources for better verification accuracy",
			Action:      "Consider additional data sources or request more complete business information",
			Reason:      fmt.Sprintf("%d out of %d fields have low confidence scores", lowQualityFields, totalFields),
			Impact:      "Medium - better data leads to more accurate verification",
		})
	}

	// Check for missing critical fields
	criticalFields := []string{"business_name", "phone_numbers", "addresses", "website_url"}
	missingCritical := []string{}

	for _, criticalField := range criticalFields {
		if _, exists := comparison.FieldResults[criticalField]; !exists {
			missingCritical = append(missingCritical, criticalField)
		}
	}

	if len(missingCritical) > 0 {
		recommendations = append(recommendations, Recommendation{
			Type:        "missing_data",
			Priority:    "high",
			Description: "Critical fields missing from verification",
			Action:      fmt.Sprintf("Collect missing data for: %s", strings.Join(missingCritical, ", ")),
			Reason:      "Essential verification fields were not found or extracted",
			Impact:      "High - incomplete verification affects decision quality",
		})
	}

	return recommendations
}

// generateRiskFactors identifies potential risks in verification
func (g *VerificationReasoningGenerator) generateRiskFactors(
	result *VerificationResult,
	comparison *ComparisonResult,
) []RiskFactor {
	var riskFactors []RiskFactor

	// Overall score risk
	if result.OverallScore < 0.6 {
		riskFactors = append(riskFactors, RiskFactor{
			Factor:      "low_verification_score",
			Severity:    "high",
			Description: "Overall verification score is below acceptable threshold",
			Impact:      "May indicate fraudulent business or data quality issues",
			Mitigation:  "Manual verification and additional documentation required",
			Probability: 0.7,
		})
	}

	// Field-specific risks
	if comparison != nil {
		for fieldName, field := range comparison.FieldResults {
			if field.Score < 0.4 {
				riskFactors = append(riskFactors, RiskFactor{
					Factor:      fmt.Sprintf("failed_%s", fieldName),
					Severity:    "medium",
					Description: fmt.Sprintf("%s field failed verification", fieldName),
					Impact:      "Reduces confidence in business ownership claims",
					Mitigation:  "Manual verification of this field required",
					Probability: 0.6,
				})
			}
		}
	}

	return riskFactors
}

// GetConfig returns the current configuration
func (g *VerificationReasoningGenerator) GetConfig() *VerificationReasoningConfig {
	return g.config
}

// GenerateVerificationReport creates a comprehensive verification report with all comparison details
func (g *VerificationReasoningGenerator) GenerateVerificationReport(
	verificationID, businessName, websiteURL string,
	result *VerificationResult,
	comparison *ComparisonResult,
	includeAudit bool,
	metadata map[string]interface{},
) (*VerificationReport, error) {
	if result == nil {
		return nil, fmt.Errorf("verification result cannot be nil")
	}

	startTime := time.Now()

	// Generate reasoning
	reasoning, err := g.GenerateReasoning(verificationID, businessName, websiteURL, result, comparison)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reasoning: %w", err)
	}

	// Create comparison details
	comparisonDetails := g.createComparisonDetails(comparison)

	// Create audit trail
	var auditTrail []AuditEvent
	if includeAudit {
		auditTrail = g.generateAuditTrail(verificationID, result, comparison)
	}

	// Create the report
	report := &VerificationReport{
		ReportID:          fmt.Sprintf("report_%s_%d", verificationID, time.Now().Unix()),
		VerificationID:    verificationID,
		BusinessName:      businessName,
		WebsiteURL:        websiteURL,
		Status:            string(result.Status),
		OverallScore:      result.OverallScore,
		ConfidenceLevel:   reasoning.ConfidenceLevel,
		GeneratedAt:       startTime,
		CompletedAt:       time.Now(),
		Duration:          time.Since(startTime),
		Reasoning:         reasoning,
		ComparisonDetails: comparisonDetails,
		AuditTrail:        auditTrail,
		Metadata:          metadata,
	}

	return report, nil
}

// createComparisonDetails converts ComparisonResult to detailed ComparisonDetails
func (g *VerificationReasoningGenerator) createComparisonDetails(comparison *ComparisonResult) *ComparisonDetails {
	if comparison == nil {
		return &ComparisonDetails{}
	}

	details := &ComparisonDetails{}

	// Process each field in the comparison result
	for fieldName, fieldComparison := range comparison.FieldResults {
		comparisonField := ComparisonField{
			FieldName:  fieldName,
			Score:      fieldComparison.Score,
			Status:     g.getFieldStatus(fieldComparison.Score),
			Confidence: fieldComparison.Confidence,
			Algorithm:  g.getAlgorithmForField(fieldName),
			Threshold:  g.getThresholdForField(fieldName),
			Matched:    fieldComparison.Matched,
			Reasoning:  g.generateFieldExplanation(fieldName, fieldComparison),
			Evidence:   g.generateFieldEvidence(fieldComparison),
		}

		// Assign to appropriate field based on field name
		switch fieldName {
		case "business_name":
			details.BusinessName = comparisonField
		case "phone_numbers":
			details.PhoneNumbers = comparisonField
		case "email_addresses":
			details.EmailAddresses = comparisonField
		case "addresses":
			details.Addresses = comparisonField
		case "website_url":
			details.WebsiteURL = comparisonField
		case "industry":
			details.Industry = comparisonField
		case "founded_year":
			details.FoundedYear = comparisonField
		case "employee_count":
			details.EmployeeCount = comparisonField
		case "revenue":
			details.Revenue = comparisonField
		case "services":
			details.Services = comparisonField
		}
	}

	return details
}

// getAlgorithmForField returns the algorithm used for each field type
func (g *VerificationReasoningGenerator) getAlgorithmForField(fieldName string) string {
	switch fieldName {
	case "business_name":
		return "fuzzy_string_matching"
	case "phone_numbers":
		return "normalized_format_matching"
	case "email_addresses":
		return "domain_and_local_analysis"
	case "addresses":
		return "geographic_component_matching"
	case "website_url":
		return "domain_comparison"
	case "industry":
		return "category_classification"
	case "founded_year":
		return "exact_year_matching"
	case "employee_count":
		return "range_comparison"
	case "revenue":
		return "range_comparison"
	case "services":
		return "keyword_similarity"
	default:
		return "similarity_analysis"
	}
}

// getThresholdForField returns the threshold used for each field type
func (g *VerificationReasoningGenerator) getThresholdForField(fieldName string) float64 {
	switch fieldName {
	case "business_name":
		return 0.8
	case "phone_numbers":
		return 0.9
	case "email_addresses":
		return 0.85
	case "addresses":
		return 0.75
	case "website_url":
		return 0.95
	case "industry":
		return 0.7
	case "founded_year":
		return 1.0
	case "employee_count":
		return 0.8
	case "revenue":
		return 0.8
	case "services":
		return 0.6
	default:
		return 0.8
	}
}

// generateAuditTrail creates a comprehensive audit trail for the verification process
func (g *VerificationReasoningGenerator) generateAuditTrail(
	verificationID string,
	result *VerificationResult,
	comparison *ComparisonResult,
) []AuditEvent {
	var auditTrail []AuditEvent
	currentTime := time.Now()

	// Initial verification start event
	auditTrail = append(auditTrail, AuditEvent{
		EventID:     fmt.Sprintf("audit_%s_start_%d", verificationID, currentTime.Unix()),
		Timestamp:   currentTime.Add(-time.Minute * 5), // Simulate start time
		EventType:   "verification_started",
		Description: "Business verification process initiated",
		Severity:    "info",
		UserID:      "system",
		Data: map[string]interface{}{
			"verification_id": verificationID,
			"process_type":    "automated_verification",
		},
	})

	// Data extraction event
	auditTrail = append(auditTrail, AuditEvent{
		EventID:     fmt.Sprintf("audit_%s_extract_%d", verificationID, currentTime.Unix()+1),
		Timestamp:   currentTime.Add(-time.Minute * 4),
		EventType:   "data_extracted",
		Description: "Business data extracted from website",
		Severity:    "info",
		UserID:      "system",
		Data: map[string]interface{}{
			"fields_extracted":  len(result.FieldResults),
			"extraction_method": "web_scraping",
		},
	})

	// Comparison events for each field
	if comparison != nil {
		for fieldName, fieldResult := range comparison.FieldResults {
			severity := "info"
			if fieldResult.Score < 0.6 {
				severity = "warning"
			}

			auditTrail = append(auditTrail, AuditEvent{
				EventID:     fmt.Sprintf("audit_%s_%s_%d", verificationID, fieldName, currentTime.Unix()),
				Timestamp:   currentTime.Add(-time.Minute * 3),
				EventType:   "field_compared",
				Description: fmt.Sprintf("%s field comparison completed", fieldName),
				Severity:    severity,
				UserID:      "system",
				Data: map[string]interface{}{
					"field_name": fieldName,
					"score":      fieldResult.Score,
					"matched":    fieldResult.Matched,
					"confidence": fieldResult.Confidence,
				},
			})
		}
	}

	// Confidence scoring event
	auditTrail = append(auditTrail, AuditEvent{
		EventID:     fmt.Sprintf("audit_%s_confidence_%d", verificationID, currentTime.Unix()+2),
		Timestamp:   currentTime.Add(-time.Minute * 2),
		EventType:   "confidence_calculated",
		Description: fmt.Sprintf("Confidence score calculated: %.2f", result.OverallScore),
		Severity:    "info",
		UserID:      "system",
		Data: map[string]interface{}{
			"overall_score":    result.OverallScore,
			"confidence_level": g.calculateConfidenceLevel(result.OverallScore),
		},
	})

	// Status assignment event
	statusSeverity := "info"
	if result.Status == StatusFailed {
		statusSeverity = "warning"
	}

	auditTrail = append(auditTrail, AuditEvent{
		EventID:     fmt.Sprintf("audit_%s_status_%d", verificationID, currentTime.Unix()+3),
		Timestamp:   currentTime.Add(-time.Minute * 1),
		EventType:   "status_assigned",
		Description: fmt.Sprintf("Verification status assigned: %s", result.Status),
		Severity:    statusSeverity,
		UserID:      "system",
		Data: map[string]interface{}{
			"status":        string(result.Status),
			"overall_score": result.OverallScore,
			"threshold":     0.8,
		},
	})

	// Report generation event
	auditTrail = append(auditTrail, AuditEvent{
		EventID:     fmt.Sprintf("audit_%s_report_%d", verificationID, currentTime.Unix()+4),
		Timestamp:   currentTime,
		EventType:   "report_generated",
		Description: "Comprehensive verification report generated",
		Severity:    "info",
		UserID:      "system",
		Data: map[string]interface{}{
			"report_type":     "comprehensive",
			"includes_audit":  true,
			"reasoning_items": 4, // approximate count
		},
	})

	return auditTrail
}
