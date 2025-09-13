package placeholders

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPlaceholderConfigManager(t *testing.T) {
	tests := []struct {
		name        string
		environment Environment
		configPath  string
		expectError bool
	}{
		{
			name:        "valid configuration",
			environment: EnvironmentDevelopment,
			configPath:  "/tmp/test-config.json",
			expectError: false,
		},
		{
			name:        "empty config path",
			environment: EnvironmentProduction,
			configPath:  "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewPlaceholderConfigManager(tt.environment, tt.configPath)

			if manager == nil {
				t.Fatal("Expected non-nil manager")
			}

			if manager.environment != tt.environment {
				t.Errorf("Expected environment %s, got %s", tt.environment, manager.environment)
			}

			if manager.configPath != tt.configPath {
				t.Errorf("Expected config path %s, got %s", tt.configPath, manager.configPath)
			}

			if manager.configs == nil {
				t.Error("Expected non-nil configs map")
			}
		})
	}
}

func TestLoadConfigurations(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.json")

	tests := []struct {
		name        string
		setupFile   bool
		fileContent string
		expectError bool
	}{
		{
			name:        "load from existing file",
			setupFile:   true,
			fileContent: `[{"id":"test-feature","name":"Test Feature","description":"Test","category":"analytics","priority":1,"status":"coming_soon","created_at":"2023-01-01T00:00:00Z","updated_at":"2023-01-01T00:00:00Z"}]`,
			expectError: false,
		},
		{
			name:        "file does not exist - create defaults",
			setupFile:   false,
			expectError: false,
		},
		{
			name:        "invalid JSON file",
			setupFile:   true,
			fileContent: `invalid json`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup test file if needed
			if tt.setupFile {
				err := os.WriteFile(configPath, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			// Use empty config path for "file does not exist" test case
			testConfigPath := configPath
			if tt.name == "file does not exist - create defaults" {
				testConfigPath = ""
			}
			manager := NewPlaceholderConfigManager(EnvironmentDevelopment, testConfigPath)
			err := manager.LoadConfigurations()

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Clean up
			if tt.setupFile {
				os.Remove(configPath)
			}
		})
	}
}

func TestSaveConfigurations(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "test-config.json")

	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, configPath)

	// Add a test configuration
	testConfig := &FeatureConfiguration{
		ID:          "test-feature",
		Name:        "Test Feature",
		Description: "Test Description",
		Category:    CategoryAnalytics,
		Priority:    PriorityHigh,
		Status:      StatusComingSoon,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := manager.AddConfiguration(testConfig)
	if err != nil {
		t.Fatalf("Failed to add test configuration: %v", err)
	}

	// Save configurations
	err = manager.SaveConfigurations()
	if err != nil {
		t.Errorf("Failed to save configurations: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("Configuration file was not created")
	}

	// Verify file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Errorf("Failed to read saved configuration file: %v", err)
	}

	if len(content) == 0 {
		t.Error("Saved configuration file is empty")
	}
}

func TestGetConfiguration(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")
	err := manager.LoadConfigurations() // Load default configurations
	if err != nil {
		t.Fatalf("Failed to load configurations: %v", err)
	}

	tests := []struct {
		name        string
		featureID   string
		expectError bool
	}{
		{
			name:        "existing feature",
			featureID:   "advanced_analytics",
			expectError: false,
		},
		{
			name:        "non-existing feature",
			featureID:   "non-existing",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := manager.GetConfiguration(tt.featureID)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && config == nil {
				t.Error("Expected configuration but got nil")
			}
			if !tt.expectError && config != nil && config.ID != tt.featureID {
				t.Errorf("Expected feature ID %s, got %s", tt.featureID, config.ID)
			}
		})
	}
}

func TestGetConfigurationsByCategory(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")
	err := manager.LoadConfigurations()
	if err != nil {
		t.Fatalf("Failed to load configurations: %v", err)
	}

	tests := []struct {
		name     string
		category FeatureCategory
		expected int
	}{
		{
			name:     "analytics category",
			category: CategoryAnalytics,
			expected: 2, // advanced_analytics and merchant_comparison
		},
		{
			name:     "automation category",
			category: CategoryAutomation,
			expected: 1, // bulk_operations
		},
		{
			name:     "non-existing category",
			category: CategoryUI,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := manager.GetConfigurationsByCategory(tt.category)

			if len(configs) != tt.expected {
				t.Errorf("Expected %d configurations, got %d", tt.expected, len(configs))
			}

			// Verify all returned configurations have the correct category
			for _, config := range configs {
				if config.Category != tt.category {
					t.Errorf("Expected category %s, got %s", tt.category, config.Category)
				}
			}
		})
	}
}

func TestGetConfigurationsByStatus(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")
	err := manager.LoadConfigurations()
	if err != nil {
		t.Fatalf("Failed to load configurations: %v", err)
	}

	tests := []struct {
		name     string
		status   FeatureStatus
		expected int
	}{
		{
			name:     "coming soon status",
			status:   StatusComingSoon,
			expected: 6, // advanced_analytics, external_api_integration, automated_reporting, real_time_monitoring, advanced_security, mobile_app
		},
		{
			name:     "in development status",
			status:   StatusInDevelopment,
			expected: 2, // bulk_operations and merchant_comparison
		},
		{
			name:     "available status",
			status:   StatusAvailable,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := manager.GetConfigurationsByStatus(tt.status)

			if len(configs) != tt.expected {
				t.Errorf("Expected %d configurations, got %d", tt.expected, len(configs))
			}

			// Verify all returned configurations have the correct status
			for _, config := range configs {
				if config.Status != tt.status {
					t.Errorf("Expected status %s, got %s", tt.status, config.Status)
				}
			}
		})
	}
}

func TestGetConfigurationsByPriority(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")
	err := manager.LoadConfigurations()
	if err != nil {
		t.Fatalf("Failed to load configurations: %v", err)
	}

	tests := []struct {
		name     string
		priority FeaturePriority
		expected int
	}{
		{
			name:     "high priority",
			priority: PriorityHigh,
			expected: 3, // advanced_analytics, bulk_operations, real_time_monitoring
		},
		{
			name:     "critical priority",
			priority: PriorityCritical,
			expected: 1, // advanced_security
		},
		{
			name:     "low priority",
			priority: PriorityLow,
			expected: 1, // mobile_app
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configs := manager.GetConfigurationsByPriority(tt.priority)

			if len(configs) != tt.expected {
				t.Errorf("Expected %d configurations, got %d", tt.expected, len(configs))
			}

			// Verify all returned configurations have the correct priority
			for _, config := range configs {
				if config.Priority != tt.priority {
					t.Errorf("Expected priority %d, got %d", tt.priority, config.Priority)
				}
			}
		})
	}
}

func TestGetEnvironmentConfig(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")
	err := manager.LoadConfigurations()
	if err != nil {
		t.Fatalf("Failed to load configurations: %v", err)
	}

	tests := []struct {
		name        string
		featureID   string
		expectError bool
	}{
		{
			name:        "existing feature with environment config",
			featureID:   "advanced_analytics",
			expectError: false,
		},
		{
			name:        "non-existing feature",
			featureID:   "non-existing",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envConfig, err := manager.GetEnvironmentConfig(tt.featureID)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !tt.expectError && envConfig == nil {
				t.Error("Expected environment config but got nil")
			}
		})
	}
}

func TestUpdateConfiguration(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")
	err := manager.LoadConfigurations()
	if err != nil {
		t.Fatalf("Failed to load configurations: %v", err)
	}

	tests := []struct {
		name        string
		featureID   string
		updates     *FeatureConfiguration
		expectError bool
	}{
		{
			name:      "valid update",
			featureID: "advanced_analytics",
			updates: &FeatureConfiguration{
				Name:        "Updated Analytics",
				Description: "Updated Description",
				Priority:    PriorityCritical,
			},
			expectError: false,
		},
		{
			name:      "non-existing feature",
			featureID: "non-existing",
			updates: &FeatureConfiguration{
				Name: "Updated Name",
			},
			expectError: true,
		},
		{
			name:        "nil updates",
			featureID:   "advanced_analytics",
			updates:     nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.UpdateConfiguration(tt.featureID, tt.updates)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify update was applied
			if !tt.expectError && tt.updates != nil {
				config, err := manager.GetConfiguration(tt.featureID)
				if err != nil {
					t.Errorf("Failed to get updated configuration: %v", err)
				}

				if tt.updates.Name != "" && config.Name != tt.updates.Name {
					t.Errorf("Expected name %s, got %s", tt.updates.Name, config.Name)
				}
				if tt.updates.Description != "" && config.Description != tt.updates.Description {
					t.Errorf("Expected description %s, got %s", tt.updates.Description, config.Description)
				}
				if tt.updates.Priority != 0 && config.Priority != tt.updates.Priority {
					t.Errorf("Expected priority %d, got %d", tt.updates.Priority, config.Priority)
				}
			}
		})
	}
}

func TestAddConfiguration(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")

	tests := []struct {
		name        string
		config      *FeatureConfiguration
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &FeatureConfiguration{
				ID:          "test-feature",
				Name:        "Test Feature",
				Description: "Test Description",
				Category:    CategoryAnalytics,
				Priority:    PriorityHigh,
				Status:      StatusComingSoon,
			},
			expectError: false,
		},
		{
			name: "missing ID",
			config: &FeatureConfiguration{
				Name:        "Test Feature",
				Description: "Test Description",
				Category:    CategoryAnalytics,
				Priority:    PriorityHigh,
				Status:      StatusComingSoon,
			},
			expectError: true,
		},
		{
			name: "missing name",
			config: &FeatureConfiguration{
				ID:          "test-feature",
				Description: "Test Description",
				Category:    CategoryAnalytics,
				Priority:    PriorityHigh,
				Status:      StatusComingSoon,
			},
			expectError: true,
		},
		{
			name:        "nil configuration",
			config:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.AddConfiguration(tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify configuration was added
			if !tt.expectError && tt.config != nil {
				config, err := manager.GetConfiguration(tt.config.ID)
				if err != nil {
					t.Errorf("Failed to get added configuration: %v", err)
				}
				if config.ID != tt.config.ID {
					t.Errorf("Expected ID %s, got %s", tt.config.ID, config.ID)
				}
			}
		})
	}
}

func TestRemoveConfiguration(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")
	err := manager.LoadConfigurations()
	if err != nil {
		t.Fatalf("Failed to load configurations: %v", err)
	}

	tests := []struct {
		name        string
		featureID   string
		expectError bool
	}{
		{
			name:        "existing feature",
			featureID:   "advanced_analytics",
			expectError: false,
		},
		{
			name:        "non-existing feature",
			featureID:   "non-existing",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.RemoveConfiguration(tt.featureID)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// Verify configuration was removed
			if !tt.expectError {
				_, err := manager.GetConfiguration(tt.featureID)
				if err == nil {
					t.Error("Expected error when getting removed configuration")
				}
			}
		})
	}
}

func TestGetAllConfigurations(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")
	err := manager.LoadConfigurations()
	if err != nil {
		t.Fatalf("Failed to load configurations: %v", err)
	}

	configs := manager.GetAllConfigurations()

	if len(configs) == 0 {
		t.Error("Expected non-empty configurations map")
	}

	// Verify we can't modify the returned map
	originalCount := len(configs)
	configs["test"] = &FeatureConfiguration{}
	if len(manager.GetAllConfigurations()) != originalCount {
		t.Error("Returned configurations map should be a copy")
	}
}

func TestGetConfigurationStatistics(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")
	err := manager.LoadConfigurations()
	if err != nil {
		t.Fatalf("Failed to load configurations: %v", err)
	}

	stats := manager.GetConfigurationStatistics()

	// Verify required statistics are present
	requiredStats := []string{"total_configurations", "by_status", "by_category", "by_priority", "enabled_in_environment"}
	for _, stat := range requiredStats {
		if _, exists := stats[stat]; !exists {
			t.Errorf("Missing required statistic: %s", stat)
		}
	}

	// Verify total configurations count
	total, ok := stats["total_configurations"].(int)
	if !ok {
		t.Error("total_configurations should be an int")
	}
	if total <= 0 {
		t.Error("Expected positive total configurations count")
	}
}

func TestValidateConfiguration(t *testing.T) {
	manager := NewPlaceholderConfigManager(EnvironmentDevelopment, "")

	tests := []struct {
		name        string
		config      *FeatureConfiguration
		expectError bool
	}{
		{
			name: "valid configuration",
			config: &FeatureConfiguration{
				ID:          "test-feature",
				Name:        "Test Feature",
				Description: "Test Description",
				Category:    CategoryAnalytics,
				Priority:    PriorityHigh,
				Status:      StatusComingSoon,
			},
			expectError: false,
		},
		{
			name:        "nil configuration",
			config:      nil,
			expectError: true,
		},
		{
			name: "missing ID",
			config: &FeatureConfiguration{
				Name:        "Test Feature",
				Description: "Test Description",
				Category:    CategoryAnalytics,
				Priority:    PriorityHigh,
				Status:      StatusComingSoon,
			},
			expectError: true,
		},
		{
			name: "missing name",
			config: &FeatureConfiguration{
				ID:          "test-feature",
				Description: "Test Description",
				Category:    CategoryAnalytics,
				Priority:    PriorityHigh,
				Status:      StatusComingSoon,
			},
			expectError: true,
		},
		{
			name: "missing description",
			config: &FeatureConfiguration{
				ID:       "test-feature",
				Name:     "Test Feature",
				Category: CategoryAnalytics,
				Priority: PriorityHigh,
				Status:   StatusComingSoon,
			},
			expectError: true,
		},
		{
			name: "missing category",
			config: &FeatureConfiguration{
				ID:          "test-feature",
				Name:        "Test Feature",
				Description: "Test Description",
				Priority:    PriorityHigh,
				Status:      StatusComingSoon,
			},
			expectError: true,
		},
		{
			name: "missing priority",
			config: &FeatureConfiguration{
				ID:          "test-feature",
				Name:        "Test Feature",
				Description: "Test Description",
				Category:    CategoryAnalytics,
				Status:      StatusComingSoon,
			},
			expectError: true,
		},
		{
			name: "missing status",
			config: &FeatureConfiguration{
				ID:          "test-feature",
				Name:        "Test Feature",
				Description: "Test Description",
				Category:    CategoryAnalytics,
				Priority:    PriorityHigh,
			},
			expectError: true,
		},
		{
			name: "invalid status",
			config: &FeatureConfiguration{
				ID:          "test-feature",
				Name:        "Test Feature",
				Description: "Test Description",
				Category:    CategoryAnalytics,
				Priority:    PriorityHigh,
				Status:      "invalid_status",
			},
			expectError: true,
		},
		{
			name: "invalid category",
			config: &FeatureConfiguration{
				ID:          "test-feature",
				Name:        "Test Feature",
				Description: "Test Description",
				Category:    "invalid_category",
				Priority:    PriorityHigh,
				Status:      StatusComingSoon,
			},
			expectError: true,
		},
		{
			name: "invalid priority",
			config: &FeatureConfiguration{
				ID:          "test-feature",
				Name:        "Test Feature",
				Description: "Test Description",
				Category:    CategoryAnalytics,
				Priority:    999,
				Status:      StatusComingSoon,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateConfiguration(tt.config)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestEnvironmentConstants(t *testing.T) {
	// Test environment constants
	environments := []Environment{
		EnvironmentDevelopment,
		EnvironmentStaging,
		EnvironmentProduction,
		EnvironmentTesting,
	}

	for _, env := range environments {
		if string(env) == "" {
			t.Errorf("Environment constant should not be empty: %s", env)
		}
	}
}

func TestFeatureCategoryConstants(t *testing.T) {
	// Test feature category constants
	categories := []FeatureCategory{
		CategoryAnalytics,
		CategoryReporting,
		CategoryIntegration,
		CategoryAutomation,
		CategoryMonitoring,
		CategorySecurity,
		CategoryMobile,
		CategoryUI,
		CategoryAPI,
		CategoryDatabase,
		CategoryCompliance,
		CategoryPerformance,
	}

	for _, category := range categories {
		if string(category) == "" {
			t.Errorf("Feature category constant should not be empty: %s", category)
		}
	}
}

func TestFeaturePriorityConstants(t *testing.T) {
	// Test feature priority constants
	priorities := []FeaturePriority{
		PriorityCritical,
		PriorityHigh,
		PriorityMedium,
		PriorityLow,
		PriorityNiceToHave,
	}

	for _, priority := range priorities {
		if priority <= 0 {
			t.Errorf("Feature priority constant should be positive: %d", priority)
		}
	}
}

func TestMockDataConfig(t *testing.T) {
	config := &MockDataConfig{
		Enabled:         true,
		DataSize:        "large",
		DataTypes:       []string{"metrics", "charts"},
		RealisticMode:   true,
		RefreshInterval: 5 * time.Minute,
	}

	if !config.Enabled {
		t.Error("Mock data config should be enabled")
	}
	if config.DataSize != "large" {
		t.Errorf("Expected data size 'large', got %s", config.DataSize)
	}
	if len(config.DataTypes) != 2 {
		t.Errorf("Expected 2 data types, got %d", len(config.DataTypes))
	}
	if !config.RealisticMode {
		t.Error("Realistic mode should be enabled")
	}
	if config.RefreshInterval != 5*time.Minute {
		t.Errorf("Expected refresh interval 5 minutes, got %v", config.RefreshInterval)
	}
}

func TestEnvironmentSpecificConfig(t *testing.T) {
	config := &EnvironmentSpecificConfig{
		Enabled:         true,
		MockDataEnabled: true,
		ShowPlaceholder: true,
		CustomMessage:   "Test message",
		RedirectURL:     "https://example.com",
		CustomConfig: map[string]interface{}{
			"key": "value",
		},
	}

	if !config.Enabled {
		t.Error("Environment config should be enabled")
	}
	if !config.MockDataEnabled {
		t.Error("Mock data should be enabled")
	}
	if !config.ShowPlaceholder {
		t.Error("Should show placeholder")
	}
	if config.CustomMessage != "Test message" {
		t.Errorf("Expected custom message 'Test message', got %s", config.CustomMessage)
	}
	if config.RedirectURL != "https://example.com" {
		t.Errorf("Expected redirect URL 'https://example.com', got %s", config.RedirectURL)
	}
	if config.CustomConfig["key"] != "value" {
		t.Errorf("Expected custom config value 'value', got %v", config.CustomConfig["key"])
	}
}
