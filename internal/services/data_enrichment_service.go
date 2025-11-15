package services

import (
	"context"
	"fmt"
	"log"
	"time"
)

// DataEnrichmentService provides business logic for data enrichment
type DataEnrichmentService interface {
	TriggerEnrichment(ctx context.Context, merchantID string, source string) (*EnrichmentJob, error)
	GetEnrichmentSources(ctx context.Context) ([]EnrichmentSource, error)
}

// EnrichmentJob represents a data enrichment job
type EnrichmentJob struct {
	JobID      string    `json:"jobId"`
	MerchantID string    `json:"merchantId"`
	Source     string    `json:"source"`
	Status     string    `json:"status"` // pending, processing, completed, failed
	CreatedAt  time.Time `json:"createdAt"`
}

// EnrichmentSource represents an available enrichment source
type EnrichmentSource struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

// dataEnrichmentService implements DataEnrichmentService
type dataEnrichmentService struct {
	logger *log.Logger
}

// NewDataEnrichmentService creates a new data enrichment service
func NewDataEnrichmentService(logger *log.Logger) DataEnrichmentService {
	if logger == nil {
		logger = log.Default()
	}

	return &dataEnrichmentService{
		logger: logger,
	}
}

// TriggerEnrichment triggers a data enrichment job for a merchant
func (s *dataEnrichmentService) TriggerEnrichment(ctx context.Context, merchantID string, source string) (*EnrichmentJob, error) {
	s.logger.Printf("Triggering enrichment for merchant: %s, source: %s", merchantID, source)

	// Validate source
	if !s.isValidSource(source) {
		return nil, fmt.Errorf("invalid enrichment source: %s", source)
	}

	// Generate job ID
	jobID := fmt.Sprintf("enrich-%d", time.Now().UnixNano())

	// Create enrichment job
	job := &EnrichmentJob{
		JobID:      jobID,
		MerchantID: merchantID,
		Source:     source,
		Status:     "pending",
		CreatedAt:  time.Now(),
	}

	// TODO: Save job to repository
	// TODO: Queue job for background processing

	s.logger.Printf("Enrichment job created: %s", jobID)
	return job, nil
}

// GetEnrichmentSources returns available enrichment sources
func (s *dataEnrichmentService) GetEnrichmentSources(ctx context.Context) ([]EnrichmentSource, error) {
	s.logger.Printf("Getting enrichment sources")

	// Return available sources
	sources := []EnrichmentSource{
		{
			ID:          "thomson-reuters",
			Name:        "Thomson Reuters",
			Description: "Business intelligence and compliance data",
			Enabled:     true,
		},
		{
			ID:          "dun-bradstreet",
			Name:        "Dun & Bradstreet",
			Description: "Business credit and company data",
			Enabled:     true,
		},
		{
			ID:          "government-registry",
			Name:        "Government Registry",
			Description: "Official business registration data",
			Enabled:     true,
		},
	}

	return sources, nil
}

// isValidSource validates if a source is valid
func (s *dataEnrichmentService) isValidSource(source string) bool {
	validSources := []string{
		"thomson-reuters",
		"dun-bradstreet",
		"government-registry",
		"industry",
		"query",
	}

	for _, valid := range validSources {
		if source == valid {
			return true
		}
	}

	return false
}

