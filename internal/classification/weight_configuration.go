package classification

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// WeightConfigurationManager manages method weights and configuration
type WeightConfigurationManager struct {
	configFile string
	configs    map[string]MethodConfig
	mutex      sync.RWMutex
	logger     *log.Logger
	registry   *MethodRegistry
}

// NewWeightConfigurationManager creates a new weight configuration manager
func NewWeightConfigurationManager(configFile string, registry *MethodRegistry, logger *log.Logger) *WeightConfigurationManager {
	if logger == nil {
		logger = log.Default()
	}

	return &WeightConfigurationManager{
		configFile: configFile,
		configs:    make(map[string]MethodConfig),
		logger:     logger,
		registry:   registry,
	}
}

// LoadConfiguration loads method configurations from file
func (wcm *WeightConfigurationManager) LoadConfiguration() error {
	wcm.mutex.Lock()
	defer wcm.mutex.Unlock()

	// Check if config file exists
	if _, err := os.Stat(wcm.configFile); os.IsNotExist(err) {
		wcm.logger.Printf("📝 Configuration file does not exist, creating default configuration: %s", wcm.configFile)
		return wcm.createDefaultConfiguration()
	}

	// Read configuration file
	data, err := os.ReadFile(wcm.configFile)
	if err != nil {
		return fmt.Errorf("failed to read configuration file: %w", err)
	}

	// Parse configuration
	var configs map[string]MethodConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse configuration file: %w", err)
	}

	wcm.configs = configs
	wcm.logger.Printf("✅ Loaded configuration for %d methods from %s", len(configs), wcm.configFile)

	// Apply configurations to registry
	return wcm.applyConfigurationsToRegistry()
}

// SaveConfiguration saves method configurations to file
func (wcm *WeightConfigurationManager) SaveConfiguration() error {
	wcm.mutex.RLock()
	defer wcm.mutex.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(wcm.configFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create configuration directory: %w", err)
	}

	// Marshal configuration
	data, err := json.MarshalIndent(wcm.configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	// Write to file
	if err := os.WriteFile(wcm.configFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write configuration file: %w", err)
	}

	wcm.logger.Printf("✅ Saved configuration for %d methods to %s", len(wcm.configs), wcm.configFile)
	return nil
}

// SetMethodWeight sets the weight for a specific method
func (wcm *WeightConfigurationManager) SetMethodWeight(methodName string, weight float64) error {
	wcm.mutex.Lock()
	defer wcm.mutex.Unlock()

	// Validate weight
	if weight < 0 || weight > 1 {
		return fmt.Errorf("weight must be between 0 and 1, got %.2f", weight)
	}

	// Update configuration
	config, exists := wcm.configs[methodName]
	if !exists {
		// Create new configuration
		config = MethodConfig{
			Name:        methodName,
			Weight:      weight,
			Enabled:     true,
			Description: fmt.Sprintf("Configuration for %s method", methodName),
		}
	} else {
		config.Weight = weight
	}

	wcm.configs[methodName] = config

	// Update registry
	if wcm.registry != nil {
		if err := wcm.registry.UpdateMethodConfig(methodName, config); err != nil {
			wcm.logger.Printf("⚠️ Warning: Failed to update registry for method '%s': %v", methodName, err)
		}
	}

	wcm.logger.Printf("✅ Set weight for method '%s' to %.2f", methodName, weight)
	return nil
}

// SetMethodEnabled enables or disables a specific method
func (wcm *WeightConfigurationManager) SetMethodEnabled(methodName string, enabled bool) error {
	wcm.mutex.Lock()
	defer wcm.mutex.Unlock()

	// Update configuration
	config, exists := wcm.configs[methodName]
	if !exists {
		// Create new configuration
		config = MethodConfig{
			Name:        methodName,
			Weight:      0.5, // Default weight
			Enabled:     enabled,
			Description: fmt.Sprintf("Configuration for %s method", methodName),
		}
	} else {
		config.Enabled = enabled
	}

	wcm.configs[methodName] = config

	// Update registry
	if wcm.registry != nil {
		if err := wcm.registry.UpdateMethodConfig(methodName, config); err != nil {
			wcm.logger.Printf("⚠️ Warning: Failed to update registry for method '%s': %v", methodName, err)
		}
	}

	status := "disabled"
	if enabled {
		status = "enabled"
	}
	wcm.logger.Printf("✅ %s method '%s'", status, methodName)
	return nil
}

// GetMethodWeight returns the weight for a specific method
func (wcm *WeightConfigurationManager) GetMethodWeight(methodName string) (float64, error) {
	wcm.mutex.RLock()
	defer wcm.mutex.RUnlock()

	config, exists := wcm.configs[methodName]
	if !exists {
		return 0, fmt.Errorf("method '%s' not found in configuration", methodName)
	}

	return config.Weight, nil
}

// IsMethodEnabled returns whether a specific method is enabled
func (wcm *WeightConfigurationManager) IsMethodEnabled(methodName string) (bool, error) {
	wcm.mutex.RLock()
	defer wcm.mutex.RUnlock()

	config, exists := wcm.configs[methodName]
	if !exists {
		return false, fmt.Errorf("method '%s' not found in configuration", methodName)
	}

	return config.Enabled, nil
}

// GetAllConfigurations returns all method configurations
func (wcm *WeightConfigurationManager) GetAllConfigurations() map[string]MethodConfig {
	wcm.mutex.RLock()
	defer wcm.mutex.RUnlock()

	// Return a copy to prevent external modification
	configs := make(map[string]MethodConfig)
	for name, config := range wcm.configs {
		configs[name] = config
	}

	return configs
}

// UpdateConfiguration updates the configuration for a specific method
func (wcm *WeightConfigurationManager) UpdateConfiguration(methodName string, config MethodConfig) error {
	wcm.mutex.Lock()
	defer wcm.mutex.Unlock()

	// Validate configuration
	if config.Weight < 0 || config.Weight > 1 {
		return fmt.Errorf("weight must be between 0 and 1, got %.2f", config.Weight)
	}

	config.Name = methodName
	wcm.configs[methodName] = config

	// Update registry
	if wcm.registry != nil {
		if err := wcm.registry.UpdateMethodConfig(methodName, config); err != nil {
			wcm.logger.Printf("⚠️ Warning: Failed to update registry for method '%s': %v", methodName, err)
		}
	}

	wcm.logger.Printf("✅ Updated configuration for method '%s'", methodName)
	return nil
}

// NormalizeWeights normalizes all method weights so they sum to 1.0
func (wcm *WeightConfigurationManager) NormalizeWeights() error {
	wcm.mutex.Lock()
	defer wcm.mutex.Unlock()

	// Calculate total weight
	var totalWeight float64
	for _, config := range wcm.configs {
		if config.Enabled {
			totalWeight += config.Weight
		}
	}

	if totalWeight == 0 {
		return fmt.Errorf("no enabled methods found to normalize")
	}

	// Normalize weights
	for methodName, config := range wcm.configs {
		if config.Enabled {
			config.Weight = config.Weight / totalWeight
			wcm.configs[methodName] = config

			// Update registry
			if wcm.registry != nil {
				if err := wcm.registry.UpdateMethodConfig(methodName, config); err != nil {
					wcm.logger.Printf("⚠️ Warning: Failed to update registry for method '%s': %v", methodName, err)
				}
			}
		}
	}

	wcm.logger.Printf("✅ Normalized weights for %d enabled methods", len(wcm.configs))
	return nil
}

// SetDefaultWeights sets default weights based on method type
func (wcm *WeightConfigurationManager) SetDefaultWeights() error {
	wcm.mutex.Lock()
	defer wcm.mutex.Unlock()

	// Default weights by method type
	defaultWeights := map[string]float64{
		"keyword":      0.5, // 50% - Primary method
		"ml":           0.4, // 40% - Secondary method
		"description":  0.1, // 10% - Tertiary method
		"external_api": 0.3, // 30% - External validation
		"hybrid":       0.6, // 60% - Combined approach
	}

	// Apply default weights
	for methodName, config := range wcm.configs {
		if weight, exists := defaultWeights[config.Type]; exists {
			config.Weight = weight
			wcm.configs[methodName] = config

			// Update registry
			if wcm.registry != nil {
				if err := wcm.registry.UpdateMethodConfig(methodName, config); err != nil {
					wcm.logger.Printf("⚠️ Warning: Failed to update registry for method '%s': %v", methodName, err)
				}
			}
		}
	}

	wcm.logger.Printf("✅ Set default weights for methods")
	return nil
}

// GetWeightSummary returns a summary of current weights
func (wcm *WeightConfigurationManager) GetWeightSummary() *WeightSummary {
	wcm.mutex.RLock()
	defer wcm.mutex.RUnlock()

	summary := &WeightSummary{
		TotalMethods:    len(wcm.configs),
		EnabledMethods:  0,
		DisabledMethods: 0,
		TotalWeight:     0.0,
		MethodWeights:   make(map[string]float64),
		MethodStatus:    make(map[string]bool),
		LastUpdated:     time.Now(),
	}

	for methodName, config := range wcm.configs {
		summary.MethodWeights[methodName] = config.Weight
		summary.MethodStatus[methodName] = config.Enabled

		if config.Enabled {
			summary.EnabledMethods++
			summary.TotalWeight += config.Weight
		} else {
			summary.DisabledMethods++
		}
	}

	return summary
}

// createDefaultConfiguration creates a default configuration file
func (wcm *WeightConfigurationManager) createDefaultConfiguration() error {
	// Default configurations for known methods
	defaultConfigs := map[string]MethodConfig{
		"keyword_classification": {
			Name:        "keyword_classification",
			Type:        "keyword",
			Weight:      0.5,
			Enabled:     true,
			Description: "Keyword-based classification using industry-specific keywords",
		},
		"ml_classification": {
			Name:        "ml_classification",
			Type:        "ml",
			Weight:      0.4,
			Enabled:     true,
			Description: "Machine learning-based classification using BERT models",
		},
		"description_classification": {
			Name:        "description_classification",
			Type:        "description",
			Weight:      0.1,
			Enabled:     true,
			Description: "Description-based classification using business descriptions",
		},
	}

	wcm.configs = defaultConfigs
	return wcm.SaveConfiguration()
}

// applyConfigurationsToRegistry applies loaded configurations to the registry
func (wcm *WeightConfigurationManager) applyConfigurationsToRegistry() error {
	if wcm.registry == nil {
		return nil
	}

	for methodName, config := range wcm.configs {
		if err := wcm.registry.UpdateMethodConfig(methodName, config); err != nil {
			wcm.logger.Printf("⚠️ Warning: Failed to apply configuration for method '%s': %v", methodName, err)
		}
	}

	return nil
}

// WeightSummary represents a summary of method weights
type WeightSummary struct {
	TotalMethods    int                `json:"total_methods"`
	EnabledMethods  int                `json:"enabled_methods"`
	DisabledMethods int                `json:"disabled_methods"`
	TotalWeight     float64            `json:"total_weight"`
	MethodWeights   map[string]float64 `json:"method_weights"`
	MethodStatus    map[string]bool    `json:"method_status"`
	LastUpdated     time.Time          `json:"last_updated"`
}
