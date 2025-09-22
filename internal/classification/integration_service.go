package classification

import (
	"context"
	"log"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/database"
	"kyb-platform/internal/machine_learning"
)

// IntegrationService provides a sophisticated interface for integrating multi-method classification services
type IntegrationService struct {
	multiMethodClassifier *MultiMethodClassifier
	keywordRepo           repository.KeywordRepository
	mlClassifier          *machine_learning.ContentClassifier
	logger                *log.Logger
}

// NewIntegrationService creates a new integration service with multi-method classification
func NewIntegrationService(supabaseClient *database.SupabaseClient, logger *log.Logger) *IntegrationService {
	if logger == nil {
		logger = log.Default()
	}

	// Create repository
	repo := repository.NewSupabaseKeywordRepository(supabaseClient, logger)

	// Create ML classifier with default config
	mlConfig := machine_learning.ContentClassifierConfig{
		ModelType:             "bert",
		MaxSequenceLength:     512,
		BatchSize:             32,
		LearningRate:          0.0001,
		Epochs:                10,
		ValidationSplit:       0.2,
		ConfidenceThreshold:   0.7,
		ExplainabilityEnabled: true,
	}
	mlClassifier := machine_learning.NewContentClassifier(mlConfig)

	// Create multi-method classifier
	multiMethodClassifier := NewMultiMethodClassifier(repo, mlClassifier, logger)

	return &IntegrationService{
		multiMethodClassifier: multiMethodClassifier,
		keywordRepo:           repo,
		mlClassifier:          mlClassifier,
		logger:                logger,
	}
}

// ProcessBusinessClassification processes a business classification request using multi-method voting
func (s *IntegrationService) ProcessBusinessClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*MultiMethodClassificationResult, error) {
	// Use multi-method classification with voting
	result, err := s.multiMethodClassifier.ClassifyWithMultipleMethods(
		ctx,
		businessName,
		description,
		websiteURL,
	)
	if err != nil {
		s.logger.Printf("❌ Multi-method classification failed: %v", err)
		return nil, err
	}

	s.logger.Printf("✅ Multi-method classification successful: %s (ensemble confidence: %.2f%%)",
		result.PrimaryClassification.IndustryName, result.EnsembleConfidence*100)

	return result, nil
}

// GetHealthStatus returns the health status of the integration service
func (s *IntegrationService) GetHealthStatus() map[string]interface{} {
	return map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"service":   "IntegrationService",
		"version":   "2.0.0",
	}
}

// Close closes the integration service
func (s *IntegrationService) Close() error {
	s.logger.Printf("🔄 Closing IntegrationService")
	return nil
}
