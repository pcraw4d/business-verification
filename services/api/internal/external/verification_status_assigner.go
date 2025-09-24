package external

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

// VerificationStatus represents the verification status
type VerificationStatus string

// Verification status constants
const (
	StatusPassed  VerificationStatus = "PASSED"
	StatusPartial VerificationStatus = "PARTIAL"
	StatusFailed  VerificationStatus = "FAILED"
	StatusSkipped VerificationStatus = "SKIPPED"
)

// VerificationCriteria defines the criteria for status assignment
type VerificationCriteria struct {
	// Overall score thresholds
	PassedThreshold  float64 `json:"passed_threshold"`  // Minimum score for PASSED status
	PartialThreshold float64 `json:"partial_threshold"` // Minimum score for PARTIAL status

	// Critical field requirements
	CriticalFields []string `json:"critical_fields"` // Fields that must be present and valid

	// Field-specific requirements
	FieldRequirements map[string]FieldRequirement `json:"field_requirements"`

	// Geographic requirements
	MaxDistanceKm float64 `json:"max_distance_km"` // Maximum allowed distance for address matching

	// Confidence requirements
	MinConfidenceLevel string `json:"min_confidence_level"` // Minimum confidence level required
}

// FieldRequirement defines requirements for a specific field
type FieldRequirement struct {
	Required      bool    `json:"required"`       // Whether the field is required
	MinScore      float64 `json:"min_score"`      // Minimum score for the field
	MinConfidence float64 `json:"min_confidence"` // Minimum confidence for the field
	Weight        float64 `json:"weight"`         // Weight in overall scoring
}

// VerificationResult represents the complete verification result
type VerificationResult struct {
	ID              string                 `json:"id"`
	Status          VerificationStatus     `json:"status"`
	OverallScore    float64                `json:"overall_score"`
	ConfidenceLevel string                 `json:"confidence_level"`
	FieldResults    map[string]FieldResult `json:"field_results"`
	Reasoning       string                 `json:"reasoning"`
	Recommendations []string               `json:"recommendations"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	Metadata        map[string]string      `json:"metadata"`
}

// FieldResult represents the result for a specific field
type FieldResult struct {
	Status       VerificationStatus `json:"status"`
	Score        float64            `json:"score"`
	Confidence   float64            `json:"confidence"`
	Matched      bool               `json:"matched"`
	Reasoning    string             `json:"reasoning"`
	Details      string             `json:"details"`
	LastVerified time.Time          `json:"last_verified"`
}

// StatusAssigner handles verification status assignment
type StatusAssigner struct {
	criteria *VerificationCriteria
	logger   *zap.Logger
}

// NewStatusAssigner creates a new status assigner
func NewStatusAssigner(criteria *VerificationCriteria, logger *zap.Logger) *StatusAssigner {
	if criteria == nil {
		criteria = &VerificationCriteria{
			PassedThreshold:    0.8,
			PartialThreshold:   0.6,
			CriticalFields:     []string{"business_name", "phone_numbers", "email_addresses"},
			MaxDistanceKm:      50.0,
			MinConfidenceLevel: "medium",
			FieldRequirements: map[string]FieldRequirement{
				"business_name": {
					Required:      true,
					MinScore:      0.7,
					MinConfidence: 0.6,
					Weight:        0.3,
				},
				"phone_numbers": {
					Required:      true,
					MinScore:      0.8,
					MinConfidence: 0.7,
					Weight:        0.25,
				},
				"email_addresses": {
					Required:      true,
					MinScore:      0.8,
					MinConfidence: 0.7,
					Weight:        0.25,
				},
				"addresses": {
					Required:      false,
					MinScore:      0.6,
					MinConfidence: 0.5,
					Weight:        0.1,
				},
				"website": {
					Required:      false,
					MinScore:      0.7,
					MinConfidence: 0.6,
					Weight:        0.05,
				},
				"industry": {
					Required:      false,
					MinScore:      0.6,
					MinConfidence: 0.5,
					Weight:        0.05,
				},
			},
		}
	}

	return &StatusAssigner{
		criteria: criteria,
		logger:   logger,
	}
}

// AssignVerificationStatus assigns verification status based on comparison results
func (sa *StatusAssigner) AssignVerificationStatus(ctx context.Context, comparisonResult *ComparisonResult) (*VerificationResult, error) {
	sa.logger.Info("Starting verification status assignment",
		zap.Float64("overall_score", comparisonResult.OverallScore),
		zap.String("confidence_level", comparisonResult.ConfidenceLevel))

	// Create verification result
	result := &VerificationResult{
		ID:              generateVerificationID(),
		OverallScore:    comparisonResult.OverallScore,
		ConfidenceLevel: comparisonResult.ConfidenceLevel,
		FieldResults:    make(map[string]FieldResult),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		Metadata:        make(map[string]string),
	}

	// Process field results
	sa.processFieldResults(comparisonResult, result)

	// Determine overall status
	result.Status = sa.determineOverallStatus(result)

	// Generate reasoning
	result.Reasoning = sa.generateReasoning(result)

	// Generate recommendations
	result.Recommendations = sa.generateRecommendations(result)

	sa.logger.Info("Verification status assignment completed",
		zap.String("status", string(result.Status)),
		zap.Float64("overall_score", result.OverallScore),
		zap.String("confidence_level", result.ConfidenceLevel))

	return result, nil
}

// processFieldResults processes individual field results
func (sa *StatusAssigner) processFieldResults(comparisonResult *ComparisonResult, result *VerificationResult) {
	for fieldName, fieldComparison := range comparisonResult.FieldResults {
		requirement, exists := sa.criteria.FieldRequirements[fieldName]
		if !exists {
			// Skip fields without requirements
			continue
		}

		fieldResult := FieldResult{
			Score:        fieldComparison.Score,
			Confidence:   fieldComparison.Confidence,
			Matched:      fieldComparison.Matched,
			Reasoning:    fieldComparison.Reasoning,
			Details:      fieldComparison.Details,
			LastVerified: time.Now(),
		}

		// Determine field status
		fieldResult.Status = sa.determineFieldStatus(fieldResult, requirement)

		result.FieldResults[fieldName] = fieldResult
	}
}

// determineFieldStatus determines the status for a specific field
func (sa *StatusAssigner) determineFieldStatus(fieldResult FieldResult, requirement FieldRequirement) VerificationStatus {
	// Check if field is required but missing
	if requirement.Required && fieldResult.Score == 0 {
		return StatusFailed
	}

	// Check if field meets minimum requirements
	if fieldResult.Score < requirement.MinScore || fieldResult.Confidence < requirement.MinConfidence {
		if requirement.Required {
			return StatusFailed
		}
		return StatusSkipped
	}

	// Check if field is matched
	if fieldResult.Matched {
		return StatusPassed
	}

	// Partial match
	return StatusPartial
}

// determineOverallStatus determines the overall verification status
func (sa *StatusAssigner) determineOverallStatus(result *VerificationResult) VerificationStatus {
	// Check critical field requirements
	if !sa.checkCriticalFields(result) {
		return StatusFailed
	}

	// Check overall score thresholds
	if result.OverallScore >= sa.criteria.PassedThreshold {
		return StatusPassed
	}

	if result.OverallScore >= sa.criteria.PartialThreshold {
		return StatusPartial
	}

	return StatusFailed
}

// checkCriticalFields checks if all critical fields meet requirements
func (sa *StatusAssigner) checkCriticalFields(result *VerificationResult) bool {
	for _, fieldName := range sa.criteria.CriticalFields {
		fieldResult, exists := result.FieldResults[fieldName]
		if !exists {
			sa.logger.Warn("Critical field missing", zap.String("field", fieldName))
			return false
		}

		requirement := sa.criteria.FieldRequirements[fieldName]
		if requirement.Required && fieldResult.Status == StatusFailed {
			sa.logger.Warn("Critical field failed",
				zap.String("field", fieldName),
				zap.String("status", string(fieldResult.Status)))
			return false
		}
	}

	return true
}

// generateReasoning generates detailed reasoning for the verification result
func (sa *StatusAssigner) generateReasoning(result *VerificationResult) string {
	var reasoning []string

	// Overall score reasoning
	reasoning = append(reasoning, fmt.Sprintf("Overall verification score: %.2f", result.OverallScore))

	// Status explanation
	switch result.Status {
	case StatusPassed:
		reasoning = append(reasoning, "All critical fields passed verification with high confidence")
	case StatusPartial:
		reasoning = append(reasoning, "Some fields passed verification but others require attention")
	case StatusFailed:
		reasoning = append(reasoning, "Critical fields failed verification or overall score below threshold")
	case StatusSkipped:
		reasoning = append(reasoning, "Verification was skipped due to insufficient data")
	}

	// Field-specific reasoning
	for fieldName, fieldResult := range result.FieldResults {
		fieldReason := fmt.Sprintf("%s: %s (score: %.2f, confidence: %.2f)",
			fieldName, string(fieldResult.Status), fieldResult.Score, fieldResult.Confidence)
		reasoning = append(reasoning, fieldReason)
	}

	return strings.Join(reasoning, "; ")
}

// generateRecommendations generates recommendations based on verification results
func (sa *StatusAssigner) generateRecommendations(result *VerificationResult) []string {
	var recommendations []string

	// Check for failed critical fields
	for _, fieldName := range sa.criteria.CriticalFields {
		if fieldResult, exists := result.FieldResults[fieldName]; exists {
			if fieldResult.Status == StatusFailed {
				recommendations = append(recommendations,
					fmt.Sprintf("Critical field '%s' failed verification - manual review required", fieldName))
			}
		}
	}

	// Check for low confidence fields
	for fieldName, fieldResult := range result.FieldResults {
		if fieldResult.Confidence < 0.5 {
			recommendations = append(recommendations,
				fmt.Sprintf("Field '%s' has low confidence (%.2f) - additional verification recommended",
					fieldName, fieldResult.Confidence))
		}
	}

	// Overall recommendations
	if result.OverallScore < sa.criteria.PartialThreshold {
		recommendations = append(recommendations,
			"Overall verification score is low - comprehensive manual review recommended")
	}

	if result.ConfidenceLevel == "low" {
		recommendations = append(recommendations,
			"Low confidence level - consider additional data sources for verification")
	}

	return recommendations
}

// GetCriteria returns the current verification criteria
func (sa *StatusAssigner) GetCriteria() *VerificationCriteria {
	return sa.criteria
}

// UpdateCriteria updates the verification criteria
func (sa *StatusAssigner) UpdateCriteria(criteria *VerificationCriteria) {
	sa.criteria = criteria
	sa.logger.Info("Verification criteria updated")
}

// generateVerificationID generates a unique verification ID
func generateVerificationID() string {
	return fmt.Sprintf("ver_%d", time.Now().UnixNano())
}
