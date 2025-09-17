package methods

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/shared"
)

// DescriptionClassificationMethod implements the ClassificationMethod interface for description-based classification
type DescriptionClassificationMethod struct {
	name        string
	methodType  string
	weight      float64
	enabled     bool
	description string
	logger      *log.Logger
}

// NewDescriptionClassificationMethod creates a new description classification method
func NewDescriptionClassificationMethod(logger *log.Logger) *DescriptionClassificationMethod {
	if logger == nil {
		logger = log.Default()
	}

	return &DescriptionClassificationMethod{
		name:        "description_classification",
		methodType:  "description",
		weight:      0.1,
		enabled:     true,
		description: "Description-based classification using business description analysis",
		logger:      logger,
	}
}

// GetName returns the unique name of the classification method
func (dcm *DescriptionClassificationMethod) GetName() string {
	return dcm.name
}

// GetType returns the type/category of the method
func (dcm *DescriptionClassificationMethod) GetType() string {
	return dcm.methodType
}

// GetDescription returns a human-readable description of what this method does
func (dcm *DescriptionClassificationMethod) GetDescription() string {
	return dcm.description
}

// GetWeight returns the current weight/importance of this method in the ensemble
func (dcm *DescriptionClassificationMethod) GetWeight() float64 {
	return dcm.weight
}

// SetWeight sets the weight/importance of this method in the ensemble
func (dcm *DescriptionClassificationMethod) SetWeight(weight float64) {
	dcm.weight = weight
}

// IsEnabled returns whether this method is currently enabled
func (dcm *DescriptionClassificationMethod) IsEnabled() bool {
	return dcm.enabled
}

// SetEnabled enables or disables this method
func (dcm *DescriptionClassificationMethod) SetEnabled(enabled bool) {
	dcm.enabled = enabled
}

// Classify performs the actual classification using this method
func (dcm *DescriptionClassificationMethod) Classify(ctx context.Context, businessName, description, websiteURL string) (*shared.ClassificationMethodResult, error) {
	startTime := time.Now()

	// Validate input
	if err := dcm.ValidateInput(businessName, description, websiteURL); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Perform description-based classification
	result, err := dcm.performDescriptionClassification(ctx, businessName, description, websiteURL)
	if err != nil {
		return &shared.ClassificationMethodResult{
			MethodName:     dcm.name,
			MethodType:     dcm.methodType,
			Confidence:     0.0,
			ProcessingTime: time.Since(startTime),
			Success:        false,
			Error:          err.Error(),
		}, nil
	}

	// Create evidence from description analysis
	evidence := []string{fmt.Sprintf("Description analysis with confidence %.2f%%", result.ConfidenceScore*100)}

	return &shared.ClassificationMethodResult{
		MethodName:     dcm.name,
		MethodType:     dcm.methodType,
		Confidence:     result.ConfidenceScore,
		ProcessingTime: time.Since(startTime),
		Result:         result,
		Evidence:       evidence,
		Success:        true,
	}, nil
}

// GetPerformanceMetrics returns performance metrics for this method
func (dcm *DescriptionClassificationMethod) GetPerformanceMetrics() interface{} {
	// This would typically be stored and updated by the registry
	// For now, return a basic metrics structure
	return map[string]interface{}{
		"total_requests":        0,
		"successful_requests":   0,
		"failed_requests":       0,
		"average_response_time": 0,
		"accuracy_score":        0.0,
	}
}

// ValidateInput validates the input parameters before classification
func (dcm *DescriptionClassificationMethod) ValidateInput(businessName, description, websiteURL string) error {
	if strings.TrimSpace(description) == "" {
		return fmt.Errorf("description is required for description-based classification")
	}
	return nil
}

// GetRequiredDependencies returns a list of dependencies this method requires
func (dcm *DescriptionClassificationMethod) GetRequiredDependencies() []string {
	return []string{} // No external dependencies
}

// Initialize performs any necessary initialization for this method
func (dcm *DescriptionClassificationMethod) Initialize(ctx context.Context) error {
	dcm.logger.Printf("✅ Initialized description classification method")
	return nil
}

// Cleanup performs any necessary cleanup when the method is removed
func (dcm *DescriptionClassificationMethod) Cleanup() error {
	dcm.logger.Printf("✅ Cleaned up description classification method")
	return nil
}

// performDescriptionClassification performs the actual description-based classification
func (dcm *DescriptionClassificationMethod) performDescriptionClassification(ctx context.Context, businessName, description, websiteURL string) (*shared.IndustryClassification, error) {
	// Simple description-based classification using keyword matching
	// This is a simplified implementation - in practice, this would use more sophisticated NLP

	description = strings.ToLower(strings.TrimSpace(description))
	if description == "" {
		return nil, fmt.Errorf("no description to analyze")
	}

	// Simple keyword-based industry detection from description
	industryKeywords := map[string][]string{
		"restaurant":    {"restaurant", "food", "dining", "cuisine", "menu", "chef", "kitchen", "cafe", "bistro", "bar", "grill"},
		"technology":    {"software", "technology", "tech", "app", "digital", "computer", "programming", "development", "IT", "cyber"},
		"healthcare":    {"medical", "health", "healthcare", "doctor", "clinic", "hospital", "patient", "medicine", "therapy", "wellness"},
		"retail":        {"retail", "store", "shop", "selling", "merchandise", "products", "commerce", "ecommerce", "shopping"},
		"finance":       {"financial", "finance", "banking", "investment", "money", "credit", "loan", "insurance", "accounting"},
		"legal":         {"legal", "law", "attorney", "lawyer", "court", "litigation", "legal services", "advocacy"},
		"education":     {"education", "school", "university", "learning", "teaching", "training", "academic", "student"},
		"manufacturing": {"manufacturing", "production", "factory", "industrial", "machinery", "assembly", "production"},
	}

	var bestMatch string
	var maxMatches int
	var confidence float64

	for industry, keywords := range industryKeywords {
		matches := 0
		for _, keyword := range keywords {
			if strings.Contains(description, keyword) {
				matches++
			}
		}

		if matches > maxMatches {
			maxMatches = matches
			bestMatch = industry
		}
	}

	// Calculate confidence based on matches
	if maxMatches > 0 {
		confidence = float64(maxMatches) / float64(len(industryKeywords[bestMatch]))
		if confidence > 1.0 {
			confidence = 1.0
		}
	} else {
		bestMatch = "Unknown"
		confidence = 0.0
	}

	return &shared.IndustryClassification{
		IndustryCode:         bestMatch,
		IndustryName:         strings.Title(bestMatch),
		ConfidenceScore:      confidence,
		ClassificationMethod: "description",
		Keywords:             []string{}, // Would extract from description analysis
	}, nil
}
