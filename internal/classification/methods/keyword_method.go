package methods

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification/repository"
	"github.com/pcraw4d/business-verification/internal/shared"
)

// KeywordClassificationMethod implements the ClassificationMethod interface for keyword-based classification
type KeywordClassificationMethod struct {
	name        string
	methodType  string
	weight      float64
	enabled     bool
	description string
	keywordRepo repository.KeywordRepository
	logger      *log.Logger
}

// NewKeywordClassificationMethod creates a new keyword classification method
func NewKeywordClassificationMethod(keywordRepo repository.KeywordRepository, logger *log.Logger) *KeywordClassificationMethod {
	if logger == nil {
		logger = log.Default()
	}

	return &KeywordClassificationMethod{
		name:        "keyword_classification",
		methodType:  "keyword",
		weight:      0.5,
		enabled:     true,
		description: "Keyword-based classification using industry-specific keywords from database",
		keywordRepo: keywordRepo,
		logger:      logger,
	}
}

// GetName returns the unique name of the classification method
func (kcm *KeywordClassificationMethod) GetName() string {
	return kcm.name
}

// GetType returns the type/category of the method
func (kcm *KeywordClassificationMethod) GetType() string {
	return kcm.methodType
}

// GetDescription returns a human-readable description of what this method does
func (kcm *KeywordClassificationMethod) GetDescription() string {
	return kcm.description
}

// GetWeight returns the current weight/importance of this method in the ensemble
func (kcm *KeywordClassificationMethod) GetWeight() float64 {
	return kcm.weight
}

// SetWeight sets the weight/importance of this method in the ensemble
func (kcm *KeywordClassificationMethod) SetWeight(weight float64) {
	kcm.weight = weight
}

// IsEnabled returns whether this method is currently enabled
func (kcm *KeywordClassificationMethod) IsEnabled() bool {
	return kcm.enabled
}

// SetEnabled enables or disables this method
func (kcm *KeywordClassificationMethod) SetEnabled(enabled bool) {
	kcm.enabled = enabled
}

// Classify performs the actual classification using this method
func (kcm *KeywordClassificationMethod) Classify(ctx context.Context, businessName, description, websiteURL string) (*shared.ClassificationMethodResult, error) {
	startTime := time.Now()

	// Validate input
	if err := kcm.ValidateInput(businessName, description, websiteURL); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Perform keyword-based classification
	result, err := kcm.performKeywordClassification(ctx, businessName, description, websiteURL)
	if err != nil {
		return &shared.ClassificationMethodResult{
			MethodName:     kcm.name,
			MethodType:     kcm.methodType,
			Confidence:     0.0,
			ProcessingTime: time.Since(startTime),
			Success:        false,
			Error:          err.Error(),
		}, nil
	}

	// Extract keywords for evidence
	var keywords []string
	if result != nil && len(result.Keywords) > 0 {
		keywords = result.Keywords
	}

	return &shared.ClassificationMethodResult{
		MethodName:     kcm.name,
		MethodType:     kcm.methodType,
		Confidence:     result.ConfidenceScore,
		ProcessingTime: time.Since(startTime),
		Result:         result,
		Evidence:       keywords,
		Keywords:       keywords,
		Success:        true,
	}, nil
}

// GetPerformanceMetrics returns performance metrics for this method
func (kcm *KeywordClassificationMethod) GetPerformanceMetrics() interface{} {
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
func (kcm *KeywordClassificationMethod) ValidateInput(businessName, description, websiteURL string) error {
	if strings.TrimSpace(businessName) == "" && strings.TrimSpace(description) == "" {
		return fmt.Errorf("at least one of business name or description must be provided")
	}
	return nil
}

// GetRequiredDependencies returns a list of dependencies this method requires
func (kcm *KeywordClassificationMethod) GetRequiredDependencies() []string {
	return []string{"supabase", "database"}
}

// Initialize performs any necessary initialization for this method
func (kcm *KeywordClassificationMethod) Initialize(ctx context.Context) error {
	if kcm.keywordRepo == nil {
		return fmt.Errorf("keyword repository is required")
	}
	kcm.logger.Printf("✅ Initialized keyword classification method")
	return nil
}

// Cleanup performs any necessary cleanup when the method is removed
func (kcm *KeywordClassificationMethod) Cleanup() error {
	kcm.logger.Printf("✅ Cleaned up keyword classification method")
	return nil
}

// performKeywordClassification performs the actual keyword-based classification
func (kcm *KeywordClassificationMethod) performKeywordClassification(ctx context.Context, businessName, description, websiteURL string) (*shared.IndustryClassification, error) {
	// Combine business name and description for analysis
	content := strings.TrimSpace(businessName + " " + description)
	if content == "" {
		return nil, fmt.Errorf("no content to analyze")
	}

	// For now, implement a simple keyword-based classification
	// In a real implementation, this would use the repository to find matching industries
	content = strings.ToLower(content)

	// Simple keyword matching
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
			if strings.Contains(content, keyword) {
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
		ClassificationMethod: "keyword",
		Keywords:             []string{}, // Would extract from content analysis
	}, nil
}
