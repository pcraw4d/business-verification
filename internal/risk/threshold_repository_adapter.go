package risk

import (
	"context"

	"kyb-platform/internal/database"
)

// thresholdRepositoryAdapter adapts database.ThresholdRepository to risk.ThresholdRepository interface
type thresholdRepositoryAdapter struct {
	repo *database.ThresholdRepository
}

// NewThresholdRepositoryAdapter creates an adapter that implements risk.ThresholdRepository
// using database.ThresholdRepository
func NewThresholdRepositoryAdapter(repo *database.ThresholdRepository) ThresholdRepository {
	return &thresholdRepositoryAdapter{repo: repo}
}

// CreateThreshold creates a new threshold configuration
func (a *thresholdRepositoryAdapter) CreateThreshold(ctx context.Context, config *ThresholdConfig) error {
	data := a.toData(config)
	return a.repo.CreateThreshold(ctx, data)
}

// GetThreshold retrieves a threshold configuration by ID
func (a *thresholdRepositoryAdapter) GetThreshold(ctx context.Context, id string) (*ThresholdConfig, error) {
	data, err := a.repo.GetThreshold(ctx, id)
	if err != nil {
		return nil, err
	}
	return a.fromData(data), nil
}

// UpdateThreshold updates an existing threshold configuration
func (a *thresholdRepositoryAdapter) UpdateThreshold(ctx context.Context, config *ThresholdConfig) error {
	data := a.toData(config)
	return a.repo.UpdateThreshold(ctx, data)
}

// DeleteThreshold deletes a threshold configuration
func (a *thresholdRepositoryAdapter) DeleteThreshold(ctx context.Context, id string) error {
	return a.repo.DeleteThreshold(ctx, id)
}

// ListThresholds retrieves all threshold configurations with optional filters
func (a *thresholdRepositoryAdapter) ListThresholds(ctx context.Context, category *RiskCategory, industryCode *string, activeOnly bool) ([]*ThresholdConfig, error) {
	var categoryStr *string
	if category != nil {
		s := string(*category)
		categoryStr = &s
	}

	dataList, err := a.repo.ListThresholds(ctx, categoryStr, industryCode, activeOnly)
	if err != nil {
		return nil, err
	}

	configs := make([]*ThresholdConfig, len(dataList))
	for i, data := range dataList {
		configs[i] = a.fromData(data)
	}
	return configs, nil
}

// LoadAllThresholds loads all active thresholds from the database
func (a *thresholdRepositoryAdapter) LoadAllThresholds(ctx context.Context) ([]*ThresholdConfig, error) {
	dataList, err := a.repo.LoadAllThresholds(ctx)
	if err != nil {
		return nil, err
	}

	configs := make([]*ThresholdConfig, len(dataList))
	for i, data := range dataList {
		configs[i] = a.fromData(data)
	}
	return configs, nil
}

// toData converts ThresholdConfig to ThresholdConfigData
func (a *thresholdRepositoryAdapter) toData(config *ThresholdConfig) *database.ThresholdConfigData {
	// Convert RiskLevel map to string map
	riskLevelsStr := make(map[string]float64)
	for level, value := range config.RiskLevels {
		riskLevelsStr[string(level)] = value
	}

	return &database.ThresholdConfigData{
		ID:             config.ID,
		Name:           config.Name,
		Description:    config.Description,
		Category:       string(config.Category),
		IndustryCode:   config.IndustryCode,
		BusinessType:   config.BusinessType,
		RiskLevels:     riskLevelsStr,
		IsDefault:      config.IsDefault,
		IsActive:       config.IsActive,
		Priority:       config.Priority,
		Metadata:       config.Metadata,
		CreatedAt:      config.CreatedAt,
		UpdatedAt:      config.UpdatedAt,
		CreatedBy:      config.CreatedBy,
		LastModifiedBy: config.LastModifiedBy,
	}
}

// fromData converts ThresholdConfigData to ThresholdConfig
func (a *thresholdRepositoryAdapter) fromData(data *database.ThresholdConfigData) *ThresholdConfig {
	// Convert string map to RiskLevel map
	riskLevels := make(map[RiskLevel]float64)
	for levelStr, value := range data.RiskLevels {
		riskLevels[RiskLevel(levelStr)] = value
	}

	return &ThresholdConfig{
		ID:             data.ID,
		Name:           data.Name,
		Description:    data.Description,
		Category:       RiskCategory(data.Category),
		IndustryCode:   data.IndustryCode,
		BusinessType:   data.BusinessType,
		RiskLevels:     riskLevels,
		IsDefault:      data.IsDefault,
		IsActive:       data.IsActive,
		Priority:       data.Priority,
		Metadata:       data.Metadata,
		CreatedAt:      data.CreatedAt,
		UpdatedAt:      data.UpdatedAt,
		CreatedBy:      data.CreatedBy,
		LastModifiedBy: data.LastModifiedBy,
	}
}

// Verify adapter implements interface
var _ ThresholdRepository = (*thresholdRepositoryAdapter)(nil)

