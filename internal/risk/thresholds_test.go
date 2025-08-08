package risk

import (
	"testing"
	"time"
)

func TestThresholdManager_RegisterConfig(t *testing.T) {
	manager := NewThresholdManager()

	// Test valid config registration
	config := &ThresholdConfig{
		ID:          "test_config",
		Name:        "Test Configuration",
		Description: "Test threshold configuration",
		Category:    RiskCategoryFinancial,
		RiskLevels: map[RiskLevel]float64{
			RiskLevelLow:      25.0,
			RiskLevelMedium:   50.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		},
		IsDefault:      false,
		IsActive:       true,
		Priority:       0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      "test",
		LastModifiedBy: "test",
	}

	err := manager.RegisterConfig(config)
	if err != nil {
		t.Errorf("Expected no error when registering valid config, got %v", err)
	}

	// Test invalid config (empty ID)
	invalidConfig := &ThresholdConfig{
		ID:         "",
		Name:       "Invalid Config",
		Category:   RiskCategoryFinancial,
		RiskLevels: map[RiskLevel]float64{},
	}

	err = manager.RegisterConfig(invalidConfig)
	if err == nil {
		t.Error("Expected error when registering config with empty ID")
	}

	// Test invalid config (no risk levels)
	invalidConfig2 := &ThresholdConfig{
		ID:         "invalid_config",
		Name:       "Invalid Config",
		Category:   RiskCategoryFinancial,
		RiskLevels: map[RiskLevel]float64{},
	}

	err = manager.RegisterConfig(invalidConfig2)
	if err == nil {
		t.Error("Expected error when registering config with no risk levels")
	}

	// Test invalid config (invalid progression)
	invalidConfig3 := &ThresholdConfig{
		ID:       "invalid_progression",
		Name:     "Invalid Progression",
		Category: RiskCategoryFinancial,
		RiskLevels: map[RiskLevel]float64{
			RiskLevelLow:      50.0, // Higher than medium
			RiskLevelMedium:   25.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		},
	}

	err = manager.RegisterConfig(invalidConfig3)
	if err == nil {
		t.Error("Expected error when registering config with invalid progression")
	}
}

func TestThresholdManager_GetConfig(t *testing.T) {
	manager := NewThresholdManager()

	// Register a test config
	config := &ThresholdConfig{
		ID:       "test_config",
		Name:     "Test Configuration",
		Category: RiskCategoryFinancial,
		RiskLevels: map[RiskLevel]float64{
			RiskLevelLow:      25.0,
			RiskLevelMedium:   50.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		},
		IsActive: true,
	}
	manager.RegisterConfig(config)

	// Test retrieval
	retrievedConfig, exists := manager.GetConfig("test_config")
	if !exists {
		t.Error("Expected config to exist")
	}
	if retrievedConfig.ID != "test_config" {
		t.Errorf("Expected config ID 'test_config', got %s", retrievedConfig.ID)
	}

	// Test non-existent config
	_, exists = manager.GetConfig("non_existent")
	if exists {
		t.Error("Expected config to not exist")
	}
}

func TestThresholdManager_GetConfigsByCategory(t *testing.T) {
	manager := NewThresholdManager()

	// Register configs for different categories
	configs := []*ThresholdConfig{
		{
			ID:       "financial_config",
			Name:     "Financial Config",
			Category: RiskCategoryFinancial,
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow: 25.0, RiskLevelMedium: 50.0, RiskLevelHigh: 75.0, RiskLevelCritical: 90.0,
			},
			IsActive: true,
		},
		{
			ID:       "operational_config",
			Name:     "Operational Config",
			Category: RiskCategoryOperational,
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow: 20.0, RiskLevelMedium: 45.0, RiskLevelHigh: 70.0, RiskLevelCritical: 85.0,
			},
			IsActive: true,
		},
		{
			ID:       "inactive_financial",
			Name:     "Inactive Financial Config",
			Category: RiskCategoryFinancial,
			RiskLevels: map[RiskLevel]float64{
				RiskLevelLow: 30.0, RiskLevelMedium: 55.0, RiskLevelHigh: 80.0, RiskLevelCritical: 95.0,
			},
			IsActive: false,
		},
	}

	for _, config := range configs {
		manager.RegisterConfig(config)
	}

	// Test retrieval by category
	financialConfigs := manager.GetConfigsByCategory(RiskCategoryFinancial)
	if len(financialConfigs) != 1 {
		t.Errorf("Expected 1 financial config, got %d", len(financialConfigs))
	}

	operationalConfigs := manager.GetConfigsByCategory(RiskCategoryOperational)
	if len(operationalConfigs) != 1 {
		t.Errorf("Expected 1 operational config, got %d", len(operationalConfigs))
	}
}

func TestThresholdManager_GetBestMatchConfig(t *testing.T) {
	manager := NewThresholdManager()

	// Register configs with different matching criteria
	configs := []*ThresholdConfig{
		{
			ID:         "default_financial",
			Name:       "Default Financial",
			Category:   RiskCategoryFinancial,
			IsDefault:  true,
			IsActive:   true,
			Priority:   0,
			RiskLevels: map[RiskLevel]float64{RiskLevelLow: 25.0, RiskLevelMedium: 50.0, RiskLevelHigh: 75.0, RiskLevelCritical: 90.0},
		},
		{
			ID:           "financial_industry_52",
			Name:         "Financial Industry 52",
			Category:     RiskCategoryFinancial,
			IndustryCode: "52",
			IsDefault:    false,
			IsActive:     true,
			Priority:     5,
			RiskLevels:   map[RiskLevel]float64{RiskLevelLow: 30.0, RiskLevelMedium: 55.0, RiskLevelHigh: 80.0, RiskLevelCritical: 90.0},
		},
		{
			ID:           "financial_bank_52",
			Name:         "Financial Bank 52",
			Category:     RiskCategoryFinancial,
			IndustryCode: "52",
			BusinessType: "bank",
			IsDefault:    false,
			IsActive:     true,
			Priority:     5,
			RiskLevels:   map[RiskLevel]float64{RiskLevelLow: 35.0, RiskLevelMedium: 60.0, RiskLevelHigh: 85.0, RiskLevelCritical: 95.0},
		},
	}

	for _, config := range configs {
		manager.RegisterConfig(config)
	}

	// Test exact industry and business type match
	bestMatch := manager.GetBestMatchConfig(RiskCategoryFinancial, "52", "bank")
	if bestMatch == nil {
		t.Fatal("Expected to find a matching config")
	}
	if bestMatch.ID != "financial_bank_52" {
		t.Errorf("Expected 'financial_bank_52', got %s", bestMatch.ID)
	}

	// Test industry match only
	bestMatch = manager.GetBestMatchConfig(RiskCategoryFinancial, "52", "insurance")
	if bestMatch == nil {
		t.Fatal("Expected to find a matching config")
	}
	if bestMatch.ID != "financial_industry_52" {
		t.Errorf("Expected 'financial_industry_52', got %s", bestMatch.ID)
	}

	// Test default fallback
	bestMatch = manager.GetBestMatchConfig(RiskCategoryFinancial, "99", "unknown")
	if bestMatch == nil {
		t.Fatal("Expected to find a default config")
	}
	if bestMatch.ID != "default_financial" {
		t.Errorf("Expected 'default_financial', got %s", bestMatch.ID)
	}
}

func TestThresholdManager_UpdateConfig(t *testing.T) {
	manager := NewThresholdManager()

	// Register a test config
	config := &ThresholdConfig{
		ID:       "test_config",
		Name:     "Test Configuration",
		Category: RiskCategoryFinancial,
		RiskLevels: map[RiskLevel]float64{
			RiskLevelLow:      25.0,
			RiskLevelMedium:   50.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		},
		IsActive: true,
	}
	manager.RegisterConfig(config)

	// Test updating config
	updates := map[string]interface{}{
		"name":        "Updated Configuration",
		"description": "Updated description",
		"is_active":   false,
		"priority":    15,
	}

	err := manager.UpdateConfig("test_config", updates)
	if err != nil {
		t.Errorf("Expected no error when updating config, got %v", err)
	}

	// Verify updates
	updatedConfig, exists := manager.GetConfig("test_config")
	if !exists {
		t.Fatal("Expected config to exist after update")
	}

	if updatedConfig.Name != "Updated Configuration" {
		t.Errorf("Expected name 'Updated Configuration', got %s", updatedConfig.Name)
	}

	if updatedConfig.IsActive {
		t.Error("Expected IsActive to be false after update")
	}

	if updatedConfig.Priority != 15 {
		t.Errorf("Expected priority 15, got %d", updatedConfig.Priority)
	}

	// Test updating non-existent config
	err = manager.UpdateConfig("non_existent", updates)
	if err == nil {
		t.Error("Expected error when updating non-existent config")
	}

	// Test updating with invalid risk levels
	invalidUpdates := map[string]interface{}{
		"risk_levels": map[RiskLevel]float64{
			RiskLevelLow:      50.0, // Invalid progression
			RiskLevelMedium:   25.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		},
	}

	err = manager.UpdateConfig("test_config", invalidUpdates)
	if err == nil {
		t.Error("Expected error when updating with invalid risk levels")
	}
}

func TestThresholdManager_DeleteConfig(t *testing.T) {
	manager := NewThresholdManager()

	// Register a test config
	config := &ThresholdConfig{
		ID:       "test_config",
		Name:     "Test Configuration",
		Category: RiskCategoryFinancial,
		RiskLevels: map[RiskLevel]float64{
			RiskLevelLow:      25.0,
			RiskLevelMedium:   50.0,
			RiskLevelHigh:     75.0,
			RiskLevelCritical: 90.0,
		},
		IsActive: true,
	}
	manager.RegisterConfig(config)

	// Test deletion
	err := manager.DeleteConfig("test_config")
	if err != nil {
		t.Errorf("Expected no error when deleting config, got %v", err)
	}

	// Verify deletion
	_, exists := manager.GetConfig("test_config")
	if exists {
		t.Error("Expected config to not exist after deletion")
	}

	// Test deleting non-existent config
	err = manager.DeleteConfig("non_existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent config")
	}
}

func TestThresholdManager_validateRiskLevelProgression(t *testing.T) {
	manager := NewThresholdManager()

	// Test valid progression
	validLevels := map[RiskLevel]float64{
		RiskLevelLow:      25.0,
		RiskLevelMedium:   50.0,
		RiskLevelHigh:     75.0,
		RiskLevelCritical: 90.0,
	}

	err := manager.validateRiskLevelProgression(validLevels)
	if err != nil {
		t.Errorf("Expected no error for valid progression, got %v", err)
	}

	// Test missing risk level
	invalidLevels1 := map[RiskLevel]float64{
		RiskLevelLow:    25.0,
		RiskLevelMedium: 50.0,
		RiskLevelHigh:   75.0,
		// Missing RiskLevelCritical
	}

	err = manager.validateRiskLevelProgression(invalidLevels1)
	if err == nil {
		t.Error("Expected error for missing risk level")
	}

	// Test invalid progression (low >= medium)
	invalidLevels2 := map[RiskLevel]float64{
		RiskLevelLow:      50.0, // Higher than medium
		RiskLevelMedium:   25.0,
		RiskLevelHigh:     75.0,
		RiskLevelCritical: 90.0,
	}

	err = manager.validateRiskLevelProgression(invalidLevels2)
	if err == nil {
		t.Error("Expected error for invalid progression")
	}

	// Test out of range values
	invalidLevels3 := map[RiskLevel]float64{
		RiskLevelLow:      -5.0, // Below 0
		RiskLevelMedium:   50.0,
		RiskLevelHigh:     75.0,
		RiskLevelCritical: 90.0,
	}

	err = manager.validateRiskLevelProgression(invalidLevels3)
	if err == nil {
		t.Error("Expected error for out of range values")
	}

	invalidLevels4 := map[RiskLevel]float64{
		RiskLevelLow:      25.0,
		RiskLevelMedium:   50.0,
		RiskLevelHigh:     75.0,
		RiskLevelCritical: 105.0, // Above 100
	}

	err = manager.validateRiskLevelProgression(invalidLevels4)
	if err == nil {
		t.Error("Expected error for out of range values")
	}
}

func TestCreateDefaultThresholds(t *testing.T) {
	manager := CreateDefaultThresholds()

	// Test that default configs are created
	expectedDefaultConfigs := []string{
		"default_financial",
		"default_operational",
		"default_regulatory",
		"default_reputational",
		"default_cybersecurity",
	}

	for _, expectedID := range expectedDefaultConfigs {
		config, exists := manager.GetConfig(expectedID)
		if !exists {
			t.Errorf("Expected default config %s to exist", expectedID)
		}
		if !config.IsDefault {
			t.Errorf("Expected config %s to be default", expectedID)
		}
		if !config.IsActive {
			t.Errorf("Expected config %s to be active", expectedID)
		}
	}

	// Test industry-specific configs
	expectedIndustryConfigs := []string{
		"financial_industry_52",
		"regulatory_financial_52",
		"cybersecurity_tech_54",
		"regulatory_healthcare_62",
	}

	for _, expectedID := range expectedIndustryConfigs {
		config, exists := manager.GetConfig(expectedID)
		if !exists {
			t.Errorf("Expected industry config %s to exist", expectedID)
		}
		if config.IsDefault {
			t.Errorf("Expected config %s to not be default", expectedID)
		}
		if !config.IsActive {
			t.Errorf("Expected config %s to be active", expectedID)
		}
	}
}

func TestThresholdConfigService_GetThresholdsForAssessment(t *testing.T) {
	manager := CreateDefaultThresholds()
	service := NewThresholdConfigService(manager)

	// Test industry-specific thresholds
	thresholds := service.GetThresholdsForAssessment(RiskCategoryFinancial, "52", "bank")
	if thresholds == nil {
		t.Fatal("Expected thresholds to be returned")
	}

	// Verify thresholds are appropriate for financial industry
	if thresholds[RiskLevelLow] != 30.0 {
		t.Errorf("Expected low threshold 30.0, got %f", thresholds[RiskLevelLow])
	}

	// Test default fallback with empty manager (no default configs)
	emptyManager := NewThresholdManager()
	emptyService := NewThresholdConfigService(emptyManager)
	thresholds = emptyService.GetThresholdsForAssessment(RiskCategoryFinancial, "99", "unknown")
	if thresholds == nil {
		t.Fatal("Expected default thresholds to be returned")
	}

	// Verify default thresholds
	if thresholds[RiskLevelLow] != 25.0 {
		t.Errorf("Expected default low threshold 25.0, got %f", thresholds[RiskLevelLow])
	}
}

func TestThresholdConfigService_ValidateThresholds(t *testing.T) {
	service := NewThresholdConfigService(NewThresholdManager())

	// Test valid thresholds
	validThresholds := map[RiskLevel]float64{
		RiskLevelLow:      25.0,
		RiskLevelMedium:   50.0,
		RiskLevelHigh:     75.0,
		RiskLevelCritical: 90.0,
	}

	err := service.ValidateThresholds(validThresholds)
	if err != nil {
		t.Errorf("Expected no error for valid thresholds, got %v", err)
	}

	// Test invalid thresholds
	invalidThresholds := map[RiskLevel]float64{
		RiskLevelLow:      50.0, // Higher than medium
		RiskLevelMedium:   25.0,
		RiskLevelHigh:     75.0,
		RiskLevelCritical: 90.0,
	}

	err = service.ValidateThresholds(invalidThresholds)
	if err == nil {
		t.Error("Expected error for invalid thresholds")
	}
}

func TestThresholdConfigService_ExportImportThresholds(t *testing.T) {
	manager := CreateDefaultThresholds()
	service := NewThresholdConfigService(manager)

	// Test export
	exportData, err := service.ExportThresholds()
	if err != nil {
		t.Errorf("Expected no error when exporting thresholds, got %v", err)
	}

	if len(exportData) == 0 {
		t.Error("Expected export data to not be empty")
	}

	// Test import
	newManager := NewThresholdManager()
	newService := NewThresholdConfigService(newManager)

	err = newService.ImportThresholds(exportData)
	if err != nil {
		t.Errorf("Expected no error when importing thresholds, got %v", err)
	}

	// Verify imported data
	configs := newManager.ListConfigs()
	if len(configs) == 0 {
		t.Error("Expected imported configs to not be empty")
	}

	// Test import with invalid JSON
	err = newService.ImportThresholds([]byte("invalid json"))
	if err == nil {
		t.Error("Expected error when importing invalid JSON")
	}
}

func TestThresholdManager_ListConfigs(t *testing.T) {
	manager := NewThresholdManager()

	// Register multiple configs
	configs := []*ThresholdConfig{
		{
			ID:         "config1",
			Name:       "Config 1",
			Category:   RiskCategoryFinancial,
			RiskLevels: map[RiskLevel]float64{RiskLevelLow: 25.0, RiskLevelMedium: 50.0, RiskLevelHigh: 75.0, RiskLevelCritical: 90.0},
			IsActive:   true,
		},
		{
			ID:         "config2",
			Name:       "Config 2",
			Category:   RiskCategoryOperational,
			RiskLevels: map[RiskLevel]float64{RiskLevelLow: 20.0, RiskLevelMedium: 45.0, RiskLevelHigh: 70.0, RiskLevelCritical: 85.0},
			IsActive:   true,
		},
	}

	for _, config := range configs {
		manager.RegisterConfig(config)
	}

	// Test listing configs
	listedConfigs := manager.ListConfigs()
	if len(listedConfigs) != 2 {
		t.Errorf("Expected 2 configs, got %d", len(listedConfigs))
	}

	// Verify all configs are present
	ids := make(map[string]bool)
	for _, config := range listedConfigs {
		ids[config.ID] = true
	}

	if !ids["config1"] || !ids["config2"] {
		t.Error("Expected both config IDs to be present in listed configs")
	}
}

func TestThresholdManager_GetDefaultConfig(t *testing.T) {
	manager := CreateDefaultThresholds()

	// Test getting default config for each category
	categories := []RiskCategory{
		RiskCategoryFinancial,
		RiskCategoryOperational,
		RiskCategoryRegulatory,
		RiskCategoryReputational,
		RiskCategoryCybersecurity,
	}

	for _, category := range categories {
		config, exists := manager.GetDefaultConfig(category)
		if !exists {
			t.Errorf("Expected default config to exist for category %s", category)
		}
		if !config.IsDefault {
			t.Errorf("Expected config to be default for category %s", category)
		}
		if config.Category != category {
			t.Errorf("Expected config category to be %s, got %s", category, config.Category)
		}
	}

	// Test getting default config for non-existent category
	config, exists := manager.GetDefaultConfig("non_existent_category")
	if exists {
		t.Error("Expected no default config to exist for non-existent category")
	}
	if config != nil {
		t.Error("Expected config to be nil for non-existent category")
	}
}

func TestThresholdManager_Performance(t *testing.T) {
	manager := CreateDefaultThresholds()

	// Test performance with multiple lookups
	categories := []RiskCategory{
		RiskCategoryFinancial,
		RiskCategoryOperational,
		RiskCategoryRegulatory,
		RiskCategoryReputational,
		RiskCategoryCybersecurity,
	}

	industries := []string{"52", "54", "62", "31", "44", "99"}
	businessTypes := []string{"bank", "insurance", "tech", "healthcare", "manufacturing", "retail"}

	for _, category := range categories {
		for _, industry := range industries {
			for _, businessType := range businessTypes {
				config := manager.GetBestMatchConfig(category, industry, businessType)
				if config == nil {
					t.Errorf("Expected to find config for category %s, industry %s, business type %s", category, industry, businessType)
				}
				if config.Category != category {
					t.Errorf("Expected config category to be %s, got %s", category, config.Category)
				}
			}
		}
	}
}
