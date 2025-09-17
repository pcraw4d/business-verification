package methods

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pcraw4d/business-verification/internal/machine_learning"
	"github.com/pcraw4d/business-verification/internal/shared"
)

// MLClassificationMethod implements the ClassificationMethod interface for ML-based classification
type MLClassificationMethod struct {
	name         string
	methodType   string
	weight       float64
	enabled      bool
	description  string
	mlClassifier *machine_learning.ContentClassifier
	logger       *log.Logger
}

// NewMLClassificationMethod creates a new ML classification method
func NewMLClassificationMethod(mlClassifier *machine_learning.ContentClassifier, logger *log.Logger) *MLClassificationMethod {
	if logger == nil {
		logger = log.Default()
	}

	return &MLClassificationMethod{
		name:         "ml_classification",
		methodType:   "ml",
		weight:       0.4,
		enabled:      true,
		description:  "Machine learning-based classification using BERT models and content analysis",
		mlClassifier: mlClassifier,
		logger:       logger,
	}
}

// GetName returns the unique name of the classification method
func (mlcm *MLClassificationMethod) GetName() string {
	return mlcm.name
}

// GetType returns the type/category of the method
func (mlcm *MLClassificationMethod) GetType() string {
	return mlcm.methodType
}

// GetDescription returns a human-readable description of what this method does
func (mlcm *MLClassificationMethod) GetDescription() string {
	return mlcm.description
}

// GetWeight returns the current weight/importance of this method in the ensemble
func (mlcm *MLClassificationMethod) GetWeight() float64 {
	return mlcm.weight
}

// SetWeight sets the weight/importance of this method in the ensemble
func (mlcm *MLClassificationMethod) SetWeight(weight float64) {
	mlcm.weight = weight
}

// IsEnabled returns whether this method is currently enabled
func (mlcm *MLClassificationMethod) IsEnabled() bool {
	return mlcm.enabled
}

// SetEnabled enables or disables this method
func (mlcm *MLClassificationMethod) SetEnabled(enabled bool) {
	mlcm.enabled = enabled
}

// Classify performs the actual classification using this method
func (mlcm *MLClassificationMethod) Classify(ctx context.Context, businessName, description, websiteURL string) (*shared.ClassificationMethodResult, error) {
	startTime := time.Now()

	// Validate input
	if err := mlcm.ValidateInput(businessName, description, websiteURL); err != nil {
		return nil, fmt.Errorf("input validation failed: %w", err)
	}

	// Perform ML-based classification
	result, err := mlcm.performMLClassification(ctx, businessName, description, websiteURL)
	if err != nil {
		return &shared.ClassificationMethodResult{
			MethodName:     mlcm.name,
			MethodType:     mlcm.methodType,
			Confidence:     0.0,
			ProcessingTime: time.Since(startTime),
			Success:        false,
			Error:          err.Error(),
		}, nil
	}

	// Create evidence from ML result
	evidence := []string{fmt.Sprintf("ML model prediction with confidence %.2f%%", result.ConfidenceScore*100)}

	return &shared.ClassificationMethodResult{
		MethodName:     mlcm.name,
		MethodType:     mlcm.methodType,
		Confidence:     result.ConfidenceScore,
		ProcessingTime: time.Since(startTime),
		Result:         result,
		Evidence:       evidence,
		Success:        true,
	}, nil
}

// GetPerformanceMetrics returns performance metrics for this method
func (mlcm *MLClassificationMethod) GetPerformanceMetrics() interface{} {
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
func (mlcm *MLClassificationMethod) ValidateInput(businessName, description, websiteURL string) error {
	if strings.TrimSpace(businessName) == "" && strings.TrimSpace(description) == "" {
		return fmt.Errorf("at least one of business name or description must be provided")
	}
	return nil
}

// GetRequiredDependencies returns a list of dependencies this method requires
func (mlcm *MLClassificationMethod) GetRequiredDependencies() []string {
	return []string{"ml_model", "bert"}
}

// Initialize performs any necessary initialization for this method
func (mlcm *MLClassificationMethod) Initialize(ctx context.Context) error {
	if mlcm.mlClassifier == nil {
		return fmt.Errorf("ML classifier is required")
	}
	mlcm.logger.Printf("✅ Initialized ML classification method")
	return nil
}

// Cleanup performs any necessary cleanup when the method is removed
func (mlcm *MLClassificationMethod) Cleanup() error {
	mlcm.logger.Printf("✅ Cleaned up ML classification method")
	return nil
}

// performMLClassification performs the actual ML-based classification
func (mlcm *MLClassificationMethod) performMLClassification(ctx context.Context, businessName, description, websiteURL string) (*shared.IndustryClassification, error) {
	// Combine business name and description for analysis
	content := strings.TrimSpace(businessName + " " + description)
	if content == "" {
		return nil, fmt.Errorf("no content to analyze")
	}

	// Use the ML classifier to classify the content
	mlResult, err := mlcm.mlClassifier.ClassifyContent(ctx, content, "")
	if err != nil {
		return nil, fmt.Errorf("failed to classify content with ML: %w", err)
	}

	// Convert ML result to IndustryClassification
	if len(mlResult.Classifications) == 0 {
		return &shared.IndustryClassification{
			IndustryCode:         "unknown",
			IndustryName:         "Unknown",
			ConfidenceScore:      0.0,
			ClassificationMethod: "ml",
			Keywords:             []string{},
		}, nil
	}

	// Get the top classification
	topClassification := mlResult.Classifications[0]

	// Convert to IndustryClassification format
	industryClassification := &shared.IndustryClassification{
		IndustryCode:         topClassification.Label,
		IndustryName:         strings.Title(topClassification.Label),
		ConfidenceScore:      topClassification.Confidence,
		ClassificationMethod: "ml",
		Keywords:             []string{}, // ML doesn't provide specific keywords
	}

	return industryClassification, nil
}
