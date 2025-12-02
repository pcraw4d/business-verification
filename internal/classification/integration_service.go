package classification

import (
	"context"
	"log"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/database"
	"kyb-platform/internal/machine_learning"
)

// IntegrationService provides a sophisticated interface for integrating classification services
// Phase 5.1: Updated to use IndustryDetectionService instead of MultiMethodClassifier
type IntegrationService struct {
	detectionService *IndustryDetectionService
	keywordRepo      repository.KeywordRepository
	mlClassifier     *machine_learning.ContentClassifier
	logger           *log.Logger
}

// NewIntegrationService creates a new integration service with industry detection
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

	// Create industry detection service with ML support
	detectionService := NewIndustryDetectionServiceWithML(repo, mlClassifier, nil, logger)

	return &IntegrationService{
		detectionService: detectionService,
		keywordRepo:      repo,
		mlClassifier:     mlClassifier,
		logger:           logger,
	}
}

// ProcessBusinessClassification processes a business classification request using industry detection
func (s *IntegrationService) ProcessBusinessClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (interface{}, error) {
	// Use industry detection service (which uses three-tier ML strategy)
	result, err := s.detectionService.DetectIndustry(ctx, businessName, description, websiteURL)
	if err != nil {
		s.logger.Printf("❌ Industry detection failed: %v", err)
		return nil, err
	}

	s.logger.Printf("✅ Industry detection successful: %s (confidence: %.2f%%)",
		result.IndustryName, result.Confidence*100)

	// Return result in a compatible format
	return map[string]interface{}{
		"industry_name": result.IndustryName,
		"confidence":    result.Confidence,
		"method":        result.Method,
		"reasoning":     result.Reasoning,
	}, nil
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
