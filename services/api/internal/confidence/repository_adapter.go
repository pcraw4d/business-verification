package confidence

import (
	"context"
	"fmt"

	"kyb-platform/internal/classification/repository"
)

// SupabaseRepositoryInterface defines the interface for Supabase repository operations needed by the threshold service
type SupabaseRepositoryInterface interface {
	GetIndustryByName(ctx context.Context, name string) (*repository.Industry, error)
	GetAllIndustries(ctx context.Context) ([]*repository.Industry, error)
	GetIndustryByID(ctx context.Context, id int) (*repository.Industry, error)
}

// SupabaseThresholdRepositoryAdapter adapts the existing Supabase repository to the threshold service interface
type SupabaseThresholdRepositoryAdapter struct {
	repository SupabaseRepositoryInterface
	logger     Logger
}

// NewSupabaseThresholdRepositoryAdapter creates a new adapter for the Supabase repository
func NewSupabaseThresholdRepositoryAdapter(supabaseRepo SupabaseRepositoryInterface, logger Logger) *SupabaseThresholdRepositoryAdapter {
	if logger == nil {
		logger = &defaultLogger{}
	}

	return &SupabaseThresholdRepositoryAdapter{
		repository: supabaseRepo,
		logger:     logger,
	}
}

// GetIndustryThreshold retrieves the confidence threshold for a specific industry by name
func (adapter *SupabaseThresholdRepositoryAdapter) GetIndustryThreshold(ctx context.Context, industryName string) (float64, error) {
	adapter.logger.Printf("🔍 Getting threshold for industry: %s", industryName)

	industry, err := adapter.repository.GetIndustryByName(ctx, industryName)
	if err != nil {
		adapter.logger.Printf("❌ Failed to get industry %s: %v", industryName, err)
		return 0, fmt.Errorf("failed to get industry %s: %w", industryName, err)
	}

	if !industry.IsActive {
		adapter.logger.Printf("⚠️ Industry %s is inactive", industryName)
		return 0, fmt.Errorf("industry %s is inactive", industryName)
	}

	adapter.logger.Printf("✅ Retrieved threshold for %s: %.3f", industryName, industry.ConfidenceThreshold)
	return industry.ConfidenceThreshold, nil
}

// GetAllIndustryThresholds retrieves all industry thresholds
func (adapter *SupabaseThresholdRepositoryAdapter) GetAllIndustryThresholds(ctx context.Context) (map[string]float64, error) {
	adapter.logger.Printf("🔍 Getting all industry thresholds")

	// Get all industries from the repository
	industries, err := adapter.repository.GetAllIndustries(ctx)
	if err != nil {
		adapter.logger.Printf("❌ Failed to get all industries: %v", err)
		return nil, fmt.Errorf("failed to get all industries: %w", err)
	}

	thresholds := make(map[string]float64)
	for _, industry := range industries {
		if industry.IsActive {
			thresholds[industry.Name] = industry.ConfidenceThreshold
		}
	}

	adapter.logger.Printf("✅ Retrieved %d industry thresholds", len(thresholds))
	return thresholds, nil
}

// GetIndustryByID retrieves an industry by its ID
func (adapter *SupabaseThresholdRepositoryAdapter) GetIndustryByID(ctx context.Context, industryID int) (*Industry, error) {
	adapter.logger.Printf("🔍 Getting industry by ID: %d", industryID)

	industry, err := adapter.repository.GetIndustryByID(ctx, industryID)
	if err != nil {
		adapter.logger.Printf("❌ Failed to get industry by ID %d: %v", industryID, err)
		return nil, fmt.Errorf("failed to get industry by ID %d: %w", industryID, err)
	}

	// Convert repository.Industry to confidence.Industry
	confidenceIndustry := &Industry{
		ID:                  industry.ID,
		Name:                industry.Name,
		Description:         industry.Description,
		Category:            industry.Category,
		ConfidenceThreshold: industry.ConfidenceThreshold,
		IsActive:            industry.IsActive,
		CreatedAt:           industry.CreatedAt,
		UpdatedAt:           industry.UpdatedAt,
	}

	adapter.logger.Printf("✅ Retrieved industry by ID %d: %s", industryID, industry.Name)
	return confidenceIndustry, nil
}

// GetIndustryByName retrieves an industry by its name
func (adapter *SupabaseThresholdRepositoryAdapter) GetIndustryByName(ctx context.Context, industryName string) (*Industry, error) {
	adapter.logger.Printf("🔍 Getting industry by name: %s", industryName)

	industry, err := adapter.repository.GetIndustryByName(ctx, industryName)
	if err != nil {
		adapter.logger.Printf("❌ Failed to get industry by name %s: %v", industryName, err)
		return nil, fmt.Errorf("failed to get industry by name %s: %w", industryName, err)
	}

	// Convert repository.Industry to confidence.Industry
	confidenceIndustry := &Industry{
		ID:                  industry.ID,
		Name:                industry.Name,
		Description:         industry.Description,
		Category:            industry.Category,
		ConfidenceThreshold: industry.ConfidenceThreshold,
		IsActive:            industry.IsActive,
		CreatedAt:           industry.CreatedAt,
		UpdatedAt:           industry.UpdatedAt,
	}

	adapter.logger.Printf("✅ Retrieved industry by name %s", industryName)
	return confidenceIndustry, nil
}

// ValidateThreshold validates that a threshold is within acceptable bounds
func (adapter *SupabaseThresholdRepositoryAdapter) ValidateThreshold(threshold float64) error {
	if threshold < 0.1 || threshold > 1.0 {
		return fmt.Errorf("threshold %.3f is out of valid range [0.1, 1.0]", threshold)
	}
	return nil
}

// GetThresholdStatistics returns statistics about industry thresholds
func (adapter *SupabaseThresholdRepositoryAdapter) GetThresholdStatistics(ctx context.Context) (map[string]interface{}, error) {
	adapter.logger.Printf("📊 Getting threshold statistics")

	thresholds, err := adapter.GetAllIndustryThresholds(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get threshold statistics: %w", err)
	}

	if len(thresholds) == 0 {
		return map[string]interface{}{
			"total_industries":  0,
			"average_threshold": 0.0,
			"min_threshold":     0.0,
			"max_threshold":     0.0,
		}, nil
	}

	var sum float64
	var min, max float64 = 1.0, 0.0
	count := 0

	for _, threshold := range thresholds {
		sum += threshold
		count++
		if threshold < min {
			min = threshold
		}
		if threshold > max {
			max = threshold
		}
	}

	stats := map[string]interface{}{
		"total_industries":  count,
		"average_threshold": sum / float64(count),
		"min_threshold":     min,
		"max_threshold":     max,
		"threshold_range":   max - min,
	}

	adapter.logger.Printf("📊 Threshold statistics: %+v", stats)
	return stats, nil
}
