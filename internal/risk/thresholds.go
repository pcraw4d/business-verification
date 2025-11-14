package risk

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ThresholdRepository interface for threshold persistence
type ThresholdRepository interface {
	CreateThreshold(ctx context.Context, config *ThresholdConfig) error
	GetThreshold(ctx context.Context, id string) (*ThresholdConfig, error)
	UpdateThreshold(ctx context.Context, config *ThresholdConfig) error
	DeleteThreshold(ctx context.Context, id string) error
	ListThresholds(ctx context.Context, category *RiskCategory, industryCode *string, activeOnly bool) ([]*ThresholdConfig, error)
	LoadAllThresholds(ctx context.Context) ([]*ThresholdConfig, error)
}

// ThresholdConfig represents a risk threshold configuration
type ThresholdConfig struct {
	ID             string                 `json:"id"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description"`
	Category       RiskCategory           `json:"category"`
	IndustryCode   string                 `json:"industry_code,omitempty"`
	BusinessType   string                 `json:"business_type,omitempty"`
	RiskLevels     map[RiskLevel]float64  `json:"risk_levels"`
	IsDefault      bool                   `json:"is_default"`
	IsActive       bool                   `json:"is_active"`
	Priority       int                    `json:"priority"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	CreatedBy      string                 `json:"created_by"`
	LastModifiedBy string                 `json:"last_modified_by"`
}

// ThresholdManager manages risk threshold configurations
type ThresholdManager struct {
	configs    map[string]*ThresholdConfig
	mutex      sync.RWMutex
	repository ThresholdRepository
	ctx        context.Context
}

// NewThresholdManager creates a new threshold manager
func NewThresholdManager() *ThresholdManager {
	return &ThresholdManager{
		configs: make(map[string]*ThresholdConfig),
		ctx:     context.Background(),
	}
}

// NewThresholdManagerWithRepository creates a new threshold manager with database persistence
func NewThresholdManagerWithRepository(repository ThresholdRepository) *ThresholdManager {
	return &ThresholdManager{
		configs:    make(map[string]*ThresholdConfig),
		repository: repository,
		ctx:        context.Background(),
	}
}

// RegisterConfig registers a threshold configuration
func (tm *ThresholdManager) RegisterConfig(config *ThresholdConfig) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if config.ID == "" {
		return fmt.Errorf("threshold config ID cannot be empty")
	}

	if len(config.RiskLevels) == 0 {
		return fmt.Errorf("threshold config must have at least one risk level")
	}

	// Validate risk level progression
	if err := tm.validateRiskLevelProgression(config.RiskLevels); err != nil {
		return fmt.Errorf("invalid risk level progression: %w", err)
	}

	config.UpdatedAt = time.Now()

	// Persist to database if repository is available
	if tm.repository != nil {
		if err := tm.repository.CreateThreshold(tm.ctx, config); err != nil {
			return fmt.Errorf("failed to persist threshold: %w", err)
		}
	}

	tm.configs[config.ID] = config
	return nil
}

// GetConfig retrieves a threshold configuration by ID
func (tm *ThresholdManager) GetConfig(id string) (*ThresholdConfig, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	config, exists := tm.configs[id]
	return config, exists
}

// GetConfigsByCategory retrieves all configurations for a specific category
func (tm *ThresholdManager) GetConfigsByCategory(category RiskCategory) []*ThresholdConfig {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var configs []*ThresholdConfig
	for _, config := range tm.configs {
		if config.Category == category && config.IsActive {
			configs = append(configs, config)
		}
	}
	return configs
}

// GetConfigsByIndustry retrieves all configurations for a specific industry
func (tm *ThresholdManager) GetConfigsByIndustry(industryCode string) []*ThresholdConfig {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var configs []*ThresholdConfig
	for _, config := range tm.configs {
		if config.IndustryCode == industryCode && config.IsActive {
			configs = append(configs, config)
		}
	}
	return configs
}

// GetDefaultConfig retrieves the default configuration for a category
func (tm *ThresholdManager) GetDefaultConfig(category RiskCategory) (*ThresholdConfig, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	for _, config := range tm.configs {
		if config.Category == category && config.IsDefault && config.IsActive {
			return config, true
		}
	}
	return nil, false
}

// GetBestMatchConfig finds the best matching configuration for given criteria
func (tm *ThresholdManager) GetBestMatchConfig(category RiskCategory, industryCode string, businessType string) *ThresholdConfig {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var bestMatch *ThresholdConfig
	var bestScore int

	for _, config := range tm.configs {
		if !config.IsActive {
			continue
		}

		if config.Category != category {
			continue
		}

		score := 0

		// Exact industry match gets highest score
		if config.IndustryCode == industryCode {
			score += 100
		} else if config.IndustryCode != "" {
			// Partial industry match
			if len(industryCode) >= len(config.IndustryCode) &&
				industryCode[:len(config.IndustryCode)] == config.IndustryCode {
				score += 50
			}
		}

		// Business type match
		if config.BusinessType == businessType {
			score += 25
		}

		// Default config gets base score
		if config.IsDefault {
			score += 10
		}

		// Higher priority configs get bonus
		score += config.Priority

		if score > bestScore {
			bestScore = score
			bestMatch = config
		}
	}

	return bestMatch
}

// UpdateConfig updates an existing threshold configuration
func (tm *ThresholdManager) UpdateConfig(id string, updates map[string]interface{}) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	config, exists := tm.configs[id]
	if !exists {
		// Try to load from database if repository is available
		if tm.repository != nil {
			dbConfig, err := tm.repository.GetThreshold(tm.ctx, id)
			if err != nil {
				return fmt.Errorf("threshold config with ID %s not found", id)
			}
			config = dbConfig
			tm.configs[id] = config
		} else {
			return fmt.Errorf("threshold config with ID %s not found", id)
		}
	}

	// Update fields based on the updates map
	if name, ok := updates["name"].(string); ok {
		config.Name = name
	}

	if description, ok := updates["description"].(string); ok {
		config.Description = description
	}

	if riskLevels, ok := updates["risk_levels"].(map[RiskLevel]float64); ok {
		if err := tm.validateRiskLevelProgression(riskLevels); err != nil {
			return fmt.Errorf("invalid risk level progression: %w", err)
		}
		config.RiskLevels = riskLevels
	}

	if isActive, ok := updates["is_active"].(bool); ok {
		config.IsActive = isActive
	}

	if priority, ok := updates["priority"].(int); ok {
		config.Priority = priority
	}

	if metadata, ok := updates["metadata"].(map[string]interface{}); ok {
		config.Metadata = metadata
	}

	if modifiedBy, ok := updates["last_modified_by"].(string); ok {
		config.LastModifiedBy = modifiedBy
	}

	config.UpdatedAt = time.Now()

	// Persist to database if repository is available
	if tm.repository != nil {
		if err := tm.repository.UpdateThreshold(tm.ctx, config); err != nil {
			return fmt.Errorf("failed to persist threshold update: %w", err)
		}
	}

	return nil
}

// DeleteConfig deletes a threshold configuration
func (tm *ThresholdManager) DeleteConfig(id string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.configs[id]; !exists {
		// Check if it exists in database if repository is available
		if tm.repository != nil {
			_, err := tm.repository.GetThreshold(tm.ctx, id)
			if err != nil {
				return fmt.Errorf("threshold config with ID %s not found", id)
			}
		} else {
			return fmt.Errorf("threshold config with ID %s not found", id)
		}
	}

	// Delete from database if repository is available
	if tm.repository != nil {
		if err := tm.repository.DeleteThreshold(tm.ctx, id); err != nil {
			return fmt.Errorf("failed to delete threshold from database: %w", err)
		}
	}

	delete(tm.configs, id)
	return nil
}

// ListConfigs returns all threshold configurations
func (tm *ThresholdManager) ListConfigs() []*ThresholdConfig {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	configs := make([]*ThresholdConfig, 0, len(tm.configs))
	for _, config := range tm.configs {
		configs = append(configs, config)
	}
	return configs
}

// LoadFromDatabase loads all thresholds from the database into memory
func (tm *ThresholdManager) LoadFromDatabase(ctx context.Context) error {
	if tm.repository == nil {
		return fmt.Errorf("repository not available")
	}

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	configs, err := tm.repository.LoadAllThresholds(ctx)
	if err != nil {
		return fmt.Errorf("failed to load thresholds from database: %w", err)
	}

	// Clear existing configs and load from database
	tm.configs = make(map[string]*ThresholdConfig)
	for _, config := range configs {
		tm.configs[config.ID] = config
	}

	return nil
}

// SyncToDatabase ensures all in-memory configs are persisted to database
func (tm *ThresholdManager) SyncToDatabase(ctx context.Context) error {
	if tm.repository == nil {
		return fmt.Errorf("repository not available")
	}

	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	for _, config := range tm.configs {
		// Check if it exists in database
		_, err := tm.repository.GetThreshold(ctx, config.ID)
		if err != nil {
			// Doesn't exist, create it
			if err := tm.repository.CreateThreshold(ctx, config); err != nil {
				return fmt.Errorf("failed to sync threshold %s: %w", config.ID, err)
			}
		} else {
			// Exists, update it
			if err := tm.repository.UpdateThreshold(ctx, config); err != nil {
				return fmt.Errorf("failed to sync threshold %s: %w", config.ID, err)
			}
		}
	}

	return nil
}

// validateRiskLevelProgression validates that risk levels progress logically
func (tm *ThresholdManager) validateRiskLevelProgression(levels map[RiskLevel]float64) error {
	// Check that all required risk levels are present
	requiredLevels := []RiskLevel{RiskLevelLow, RiskLevelMedium, RiskLevelHigh, RiskLevelCritical}
	for _, level := range requiredLevels {
		if _, exists := levels[level]; !exists {
			return fmt.Errorf("missing required risk level: %s", level)
		}
	}

	// Validate progression: Low < Medium < High < Critical
	if levels[RiskLevelLow] >= levels[RiskLevelMedium] {
		return fmt.Errorf("low risk threshold must be less than medium risk threshold")
	}

	if levels[RiskLevelMedium] >= levels[RiskLevelHigh] {
		return fmt.Errorf("medium risk threshold must be less than high risk threshold")
	}

	if levels[RiskLevelHigh] >= levels[RiskLevelCritical] {
		return fmt.Errorf("high risk threshold must be less than critical risk threshold")
	}

	// Validate ranges
	if levels[RiskLevelLow] < 0 || levels[RiskLevelCritical] > 100 {
		return fmt.Errorf("risk thresholds must be between 0 and 100")
	}

	return nil
}

// CreateDefaultThresholds creates default threshold configurations
func CreateDefaultThresholds() *ThresholdManager {
	manager := NewThresholdManager()

	// Default configurations for each category
	defaultConfigs := []*ThresholdConfig{
		{
			ID:          "default_financial",
			Name:        "Default Financial Risk Thresholds",
			Description: "Standard financial risk thresholds for general business assessment",
			Category:    RiskCategoryFinancial,
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow:      25.0,
				RiskLevelMedium:   50.0,
				RiskLevelHigh:     75.0,
				RiskLevelCritical: 90.0,
			},
			IsDefault:      true,
			IsActive:       true,
			Priority:       0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      "system",
			LastModifiedBy: "system",
		},
		{
			ID:          "default_operational",
			Name:        "Default Operational Risk Thresholds",
			Description: "Standard operational risk thresholds for general business assessment",
			Category:    RiskCategoryOperational,
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow:      20.0,
				RiskLevelMedium:   45.0,
				RiskLevelHigh:     70.0,
				RiskLevelCritical: 85.0,
			},
			IsDefault:      true,
			IsActive:       true,
			Priority:       0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      "system",
			LastModifiedBy: "system",
		},
		{
			ID:          "default_regulatory",
			Name:        "Default Regulatory Risk Thresholds",
			Description: "Standard regulatory risk thresholds for general business assessment",
			Category:    RiskCategoryRegulatory,
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow:      30.0,
				RiskLevelMedium:   55.0,
				RiskLevelHigh:     80.0,
				RiskLevelCritical: 95.0,
			},
			IsDefault:      true,
			IsActive:       true,
			Priority:       0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      "system",
			LastModifiedBy: "system",
		},
		{
			ID:          "default_reputational",
			Name:        "Default Reputational Risk Thresholds",
			Description: "Standard reputational risk thresholds for general business assessment",
			Category:    RiskCategoryReputational,
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow:      20.0,
				RiskLevelMedium:   45.0,
				RiskLevelHigh:     75.0,
				RiskLevelCritical: 90.0,
			},
			IsDefault:      true,
			IsActive:       true,
			Priority:       0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      "system",
			LastModifiedBy: "system",
		},
		{
			ID:          "default_cybersecurity",
			Name:        "Default Cybersecurity Risk Thresholds",
			Description: "Standard cybersecurity risk thresholds for general business assessment",
			Category:    RiskCategoryCybersecurity,
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow:      15.0,
				RiskLevelMedium:   40.0,
				RiskLevelHigh:     75.0,
				RiskLevelCritical: 90.0,
			},
			IsDefault:      true,
			IsActive:       true,
			Priority:       0,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      "system",
			LastModifiedBy: "system",
		},
	}

	// Industry-specific configurations
	industryConfigs := []*ThresholdConfig{
		{
			ID:           "financial_industry_52",
			Name:         "Financial Industry Risk Thresholds",
			Description:  "Specialized risk thresholds for financial services industry",
			Category:     RiskCategoryFinancial,
			IndustryCode: "52",
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow:      30.0,
				RiskLevelMedium:   55.0,
				RiskLevelHigh:     80.0,
				RiskLevelCritical: 90.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       10,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      "system",
			LastModifiedBy: "system",
		},
		{
			ID:           "regulatory_financial_52",
			Name:         "Financial Industry Regulatory Thresholds",
			Description:  "Specialized regulatory risk thresholds for financial services",
			Category:     RiskCategoryRegulatory,
			IndustryCode: "52",
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow:      35.0,
				RiskLevelMedium:   60.0,
				RiskLevelHigh:     85.0,
				RiskLevelCritical: 95.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       10,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      "system",
			LastModifiedBy: "system",
		},
		{
			ID:           "cybersecurity_tech_54",
			Name:         "Technology Industry Cybersecurity Thresholds",
			Description:  "Specialized cybersecurity risk thresholds for technology industry",
			Category:     RiskCategoryCybersecurity,
			IndustryCode: "54",
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow:      10.0,
				RiskLevelMedium:   35.0,
				RiskLevelHigh:     70.0,
				RiskLevelCritical: 85.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       10,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      "system",
			LastModifiedBy: "system",
		},
		{
			ID:           "regulatory_healthcare_62",
			Name:         "Healthcare Industry Regulatory Thresholds",
			Description:  "Specialized regulatory risk thresholds for healthcare industry",
			Category:     RiskCategoryRegulatory,
			IndustryCode: "62",
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow:      40.0,
				RiskLevelMedium:   65.0,
				RiskLevelHigh:     85.0,
				RiskLevelCritical: 95.0,
			},
			IsDefault:      false,
			IsActive:       true,
			Priority:       10,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
			CreatedBy:      "system",
			LastModifiedBy: "system",
		},
	}

	// Register all configurations
	for _, config := range defaultConfigs {
		manager.RegisterConfig(config)
	}

	for _, config := range industryConfigs {
		manager.RegisterConfig(config)
	}

	return manager
}

// ThresholdConfigService provides business logic for threshold management
type ThresholdConfigService struct {
	manager *ThresholdManager
}

// NewThresholdConfigService creates a new threshold configuration service
func NewThresholdConfigService(manager *ThresholdManager) *ThresholdConfigService {
	return &ThresholdConfigService{
		manager: manager,
	}
}

// GetThresholdsForAssessment retrieves the appropriate thresholds for a risk assessment
func (tcs *ThresholdConfigService) GetThresholdsForAssessment(category RiskCategory, industryCode string, businessType string) map[RiskLevel]float64 {
	config := tcs.manager.GetBestMatchConfig(category, industryCode, businessType)
	if config != nil {
		return config.RiskLevels
	}

	// Fall back to default configuration
	defaultConfig, exists := tcs.manager.GetDefaultConfig(category)
	if exists {
		return defaultConfig.RiskLevels
	}

	// Ultimate fallback to hardcoded defaults
	return map[RiskLevel]float64{
		RiskLevelLow:      25.0,
		RiskLevelMedium:   50.0,
		RiskLevelHigh:     75.0,
		RiskLevelCritical: 90.0,
	}
}

// ValidateThresholds validates threshold configurations
func (tcs *ThresholdConfigService) ValidateThresholds(thresholds map[RiskLevel]float64) error {
	// Check that all required risk levels are present
	requiredLevels := []RiskLevel{RiskLevelLow, RiskLevelMedium, RiskLevelHigh, RiskLevelCritical}
	for _, level := range requiredLevels {
		if _, exists := thresholds[level]; !exists {
			return fmt.Errorf("missing required risk level: %s", level)
		}
	}

	// Validate progression
	if thresholds[RiskLevelLow] >= thresholds[RiskLevelMedium] {
		return fmt.Errorf("low risk threshold must be less than medium risk threshold")
	}

	if thresholds[RiskLevelMedium] >= thresholds[RiskLevelHigh] {
		return fmt.Errorf("medium risk threshold must be less than high risk threshold")
	}

	if thresholds[RiskLevelHigh] >= thresholds[RiskLevelCritical] {
		return fmt.Errorf("high risk threshold must be less than critical risk threshold")
	}

	// Validate ranges
	if thresholds[RiskLevelLow] < 0 || thresholds[RiskLevelCritical] > 100 {
		return fmt.Errorf("risk thresholds must be between 0 and 100")
	}

	return nil
}

// ExportThresholds exports threshold configurations to JSON
// If manager has a repository, exports from database; otherwise exports from memory
func (tcs *ThresholdConfigService) ExportThresholds() ([]byte, error) {
	var configs []*ThresholdConfig

	// If repository is available, load from database to ensure we export all persisted configs
	if tcs.manager.repository != nil {
		ctx := context.Background()
		// LoadAllThresholds returns []*ThresholdConfig (via adapter)
		dbConfigs, err := tcs.manager.repository.LoadAllThresholds(ctx)
		if err != nil {
			// Fall back to memory if database load fails
			configs = tcs.manager.ListConfigs()
		} else {
			// Repository adapter already converts to ThresholdConfig
			configs = dbConfigs
		}
	} else {
		// No repository, export from memory
		configs = tcs.manager.ListConfigs()
	}

	return json.MarshalIndent(configs, "", "  ")
}

// ImportThresholds imports threshold configurations from JSON
// If manager has a repository, imports are persisted to database; otherwise only in memory
func (tcs *ThresholdConfigService) ImportThresholds(data []byte) error {
	var configs []*ThresholdConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to unmarshal threshold configs: %w", err)
	}

	ctx := context.Background()
	
	// Import all configs - use upsert logic if repository is available
	for _, config := range configs {
		if tcs.manager.repository != nil {
			// Check if threshold exists in database
			_, err := tcs.manager.repository.GetThreshold(ctx, config.ID)
			if err != nil {
				// Doesn't exist, create it
				if err := tcs.manager.repository.CreateThreshold(ctx, config); err != nil {
					return fmt.Errorf("failed to create imported config %s: %w", config.ID, err)
				}
			} else {
				// Exists, update it
				if err := tcs.manager.repository.UpdateThreshold(ctx, config); err != nil {
					return fmt.Errorf("failed to update imported config %s: %w", config.ID, err)
				}
			}
		}
		
		// Always register in memory (skip persistence since we already handled it above)
		tcs.manager.mutex.Lock()
		tcs.manager.configs[config.ID] = config
		tcs.manager.mutex.Unlock()
	}

	return nil
}

// ThresholdConfigRequest represents a request to create or update threshold configurations
type ThresholdConfigRequest struct {
	ID           string                 `json:"id,omitempty"`
	Name         string                 `json:"name" validate:"required"`
	Description  string                 `json:"description"`
	Category     RiskCategory           `json:"category" validate:"required"`
	IndustryCode string                 `json:"industry_code,omitempty"`
	BusinessType string                 `json:"business_type,omitempty"`
	RiskLevels   map[RiskLevel]float64  `json:"risk_levels" validate:"required"`
	IsDefault    bool                   `json:"is_default"`
	IsActive     bool                   `json:"is_active"`
	Priority     int                    `json:"priority"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy    string                 `json:"created_by"`
}

// ThresholdConfigResponse represents a response with threshold configuration data
type ThresholdConfigResponse struct {
	Configs []*ThresholdConfig `json:"configs"`
	Total   int                `json:"total"`
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
}
