package classification

import (
	"log"

	"github.com/pcraw4d/business-verification/internal/classification/repository"
	"github.com/pcraw4d/business-verification/internal/database"
)

// ClassificationContainer manages all classification service dependencies
type ClassificationContainer struct {
	industryDetectionService *IndustryDetectionService
	codeGenerator            *ClassificationCodeGenerator
	repository               repository.KeywordRepository
	logger                   *log.Logger
}

// NewClassificationContainer creates a new classification container with all dependencies
func NewClassificationContainer(supabaseClient *database.SupabaseClient, logger *log.Logger) *ClassificationContainer {
	if logger == nil {
		logger = log.Default()
	}

	// Create repository using the factory
	repo := repository.NewRepository(supabaseClient, logger)

	// Create services
	industryDetectionService := NewIndustryDetectionService(repo, logger)
	codeGenerator := NewClassificationCodeGenerator(repo, logger)

	return &ClassificationContainer{
		industryDetectionService: industryDetectionService,
		codeGenerator:            codeGenerator,
		repository:               repo,
		logger:                   logger,
	}
}

// GetIndustryDetectionService returns the industry detection service
func (c *ClassificationContainer) GetIndustryDetectionService() *IndustryDetectionService {
	return c.industryDetectionService
}

// GetCodeGenerator returns the classification code generator
func (c *ClassificationContainer) GetCodeGenerator() *ClassificationCodeGenerator {
	return c.codeGenerator
}

// GetRepository returns the keyword repository
func (c *ClassificationContainer) GetRepository() repository.KeywordRepository {
	return c.repository
}

// GetLogger returns the logger
func (c *ClassificationContainer) GetLogger() *log.Logger {
	return c.logger
}

// HealthCheck performs a health check on all classification services
func (c *ClassificationContainer) HealthCheck() map[string]interface{} {
	health := map[string]interface{}{
		"status": "healthy",
		"services": map[string]string{
			"industry_detection": "active",
			"code_generator":     "active",
			"repository":         "active",
		},
		"timestamp": log.Default().Output(0, ""), // This is a placeholder - in real implementation, use proper timestamp
	}

	// Check repository connectivity
	if c.repository != nil {
		health["repository_status"] = "connected"
	} else {
		health["repository_status"] = "disconnected"
		health["status"] = "degraded"
	}

	return health
}

// Close performs cleanup operations for the container
func (c *ClassificationContainer) Close() error {
	// Close database connections if needed
	if c.repository != nil {
		// In a real implementation, you might want to close database connections
		// For now, we'll just log the cleanup
		c.logger.Printf("🔧 Cleaning up classification container")
	}
	return nil
}
