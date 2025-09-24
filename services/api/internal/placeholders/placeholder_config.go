package placeholders

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Environment represents the deployment environment
type Environment string

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentStaging     Environment = "staging"
	EnvironmentProduction  Environment = "production"
	EnvironmentTesting     Environment = "testing"
)

// FeatureCategory represents the category of a placeholder feature
type FeatureCategory string

const (
	CategoryAnalytics   FeatureCategory = "analytics"
	CategoryReporting   FeatureCategory = "reporting"
	CategoryIntegration FeatureCategory = "integration"
	CategoryAutomation  FeatureCategory = "automation"
	CategoryMonitoring  FeatureCategory = "monitoring"
	CategorySecurity    FeatureCategory = "security"
	CategoryMobile      FeatureCategory = "mobile"
	CategoryUI          FeatureCategory = "ui"
	CategoryAPI         FeatureCategory = "api"
	CategoryDatabase    FeatureCategory = "database"
	CategoryCompliance  FeatureCategory = "compliance"
	CategoryPerformance FeatureCategory = "performance"
)

// FeaturePriority represents the priority level of a feature
type FeaturePriority int

const (
	PriorityCritical   FeaturePriority = 1
	PriorityHigh       FeaturePriority = 2
	PriorityMedium     FeaturePriority = 3
	PriorityLow        FeaturePriority = 4
	PriorityNiceToHave FeaturePriority = 5
)

// FeatureConfiguration defines the configuration for a specific feature
type FeatureConfiguration struct {
	ID                  string                                    `json:"id"`
	Name                string                                    `json:"name"`
	Description         string                                    `json:"description"`
	DetailedDescription string                                    `json:"detailed_description,omitempty"`
	Category            FeatureCategory                           `json:"category"`
	Priority            FeaturePriority                           `json:"priority"`
	Status              FeatureStatus                             `json:"status"`
	ETA                 *time.Time                                `json:"eta,omitempty"`
	DevelopmentPhase    string                                    `json:"development_phase,omitempty"`
	EstimatedEffort     string                                    `json:"estimated_effort,omitempty"`
	Prerequisites       []string                                  `json:"prerequisites,omitempty"`
	Dependencies        []string                                  `json:"dependencies,omitempty"`
	AcceptanceCriteria  []string                                  `json:"acceptance_criteria,omitempty"`
	MockDataConfig      *MockDataConfig                           `json:"mock_data_config,omitempty"`
	EnvironmentConfig   map[Environment]EnvironmentSpecificConfig `json:"environment_config,omitempty"`
	CreatedAt           time.Time                                 `json:"created_at"`
	UpdatedAt           time.Time                                 `json:"updated_at"`
}

// EnvironmentSpecificConfig defines environment-specific configuration
type EnvironmentSpecificConfig struct {
	Enabled         bool                   `json:"enabled"`
	MockDataEnabled bool                   `json:"mock_data_enabled"`
	ShowPlaceholder bool                   `json:"show_placeholder"`
	CustomMessage   string                 `json:"custom_message,omitempty"`
	RedirectURL     string                 `json:"redirect_url,omitempty"`
	CustomConfig    map[string]interface{} `json:"custom_config,omitempty"`
}

// MockDataConfig defines configuration for mock data generation
type MockDataConfig struct {
	Enabled         bool                   `json:"enabled"`
	DataSize        string                 `json:"data_size,omitempty"` // small, medium, large
	DataTypes       []string               `json:"data_types,omitempty"`
	CustomData      map[string]interface{} `json:"custom_data,omitempty"`
	RefreshInterval time.Duration          `json:"refresh_interval,omitempty"`
	RealisticMode   bool                   `json:"realistic_mode"`
}

// PlaceholderConfigManager manages placeholder feature configurations
type PlaceholderConfigManager struct {
	configs     map[string]*FeatureConfiguration
	environment Environment
	configPath  string
}

// NewPlaceholderConfigManager creates a new configuration manager
func NewPlaceholderConfigManager(environment Environment, configPath string) *PlaceholderConfigManager {
	return &PlaceholderConfigManager{
		configs:     make(map[string]*FeatureConfiguration),
		environment: environment,
		configPath:  configPath,
	}
}

// LoadConfigurations loads feature configurations from file or creates defaults
func (pcm *PlaceholderConfigManager) LoadConfigurations() error {
	// If no config path is specified, create default configurations
	if pcm.configPath == "" {
		pcm.createDefaultConfigurations()
		return nil
	}

	// Try to load from file first
	if err := pcm.loadFromFile(); err != nil {
		// If file doesn't exist, create default configurations
		if os.IsNotExist(err) {
			pcm.createDefaultConfigurations()
			return nil
		}
		// If file exists but has invalid content, return the error
		return err
	}
	return nil
}

// loadFromFile loads configurations from JSON file
func (pcm *PlaceholderConfigManager) loadFromFile() error {
	if pcm.configPath == "" {
		return fmt.Errorf("config path not specified")
	}

	data, err := os.ReadFile(pcm.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var configs []*FeatureConfiguration
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Load configurations into map
	for _, config := range configs {
		pcm.configs[config.ID] = config
	}

	return nil
}

// SaveConfigurations saves current configurations to file
func (pcm *PlaceholderConfigManager) SaveConfigurations() error {
	if pcm.configPath == "" {
		return fmt.Errorf("config path not specified")
	}

	// Ensure directory exists
	dir := filepath.Dir(pcm.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Convert map to slice
	var configs []*FeatureConfiguration
	for _, config := range pcm.configs {
		configs = append(configs, config)
	}

	data, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(pcm.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfiguration retrieves a feature configuration by ID
func (pcm *PlaceholderConfigManager) GetConfiguration(featureID string) (*FeatureConfiguration, error) {
	config, exists := pcm.configs[featureID]
	if !exists {
		return nil, fmt.Errorf("configuration for feature %s not found", featureID)
	}
	return config, nil
}

// GetConfigurationsByCategory returns configurations filtered by category
func (pcm *PlaceholderConfigManager) GetConfigurationsByCategory(category FeatureCategory) []*FeatureConfiguration {
	var configs []*FeatureConfiguration
	for _, config := range pcm.configs {
		if config.Category == category {
			configs = append(configs, config)
		}
	}
	return configs
}

// GetConfigurationsByStatus returns configurations filtered by status
func (pcm *PlaceholderConfigManager) GetConfigurationsByStatus(status FeatureStatus) []*FeatureConfiguration {
	var configs []*FeatureConfiguration
	for _, config := range pcm.configs {
		if config.Status == status {
			configs = append(configs, config)
		}
	}
	return configs
}

// GetConfigurationsByPriority returns configurations filtered by priority
func (pcm *PlaceholderConfigManager) GetConfigurationsByPriority(priority FeaturePriority) []*FeatureConfiguration {
	var configs []*FeatureConfiguration
	for _, config := range pcm.configs {
		if config.Priority == priority {
			configs = append(configs, config)
		}
	}
	return configs
}

// GetEnvironmentConfig returns environment-specific configuration for a feature
func (pcm *PlaceholderConfigManager) GetEnvironmentConfig(featureID string) (*EnvironmentSpecificConfig, error) {
	config, err := pcm.GetConfiguration(featureID)
	if err != nil {
		return nil, err
	}

	envConfig, exists := config.EnvironmentConfig[pcm.environment]
	if !exists {
		// Return default environment config
		return &EnvironmentSpecificConfig{
			Enabled:         true,
			MockDataEnabled: true,
			ShowPlaceholder: true,
		}, nil
	}

	return &envConfig, nil
}

// UpdateConfiguration updates a feature configuration
func (pcm *PlaceholderConfigManager) UpdateConfiguration(featureID string, updates *FeatureConfiguration) error {
	if updates == nil {
		return fmt.Errorf("updates cannot be nil")
	}

	config, exists := pcm.configs[featureID]
	if !exists {
		return fmt.Errorf("configuration for feature %s not found", featureID)
	}

	// Update fields if provided
	if updates.Name != "" {
		config.Name = updates.Name
	}
	if updates.Description != "" {
		config.Description = updates.Description
	}
	if updates.DetailedDescription != "" {
		config.DetailedDescription = updates.DetailedDescription
	}
	if updates.Category != "" {
		config.Category = updates.Category
	}
	if updates.Priority != 0 {
		config.Priority = updates.Priority
	}
	if updates.Status != "" {
		config.Status = updates.Status
	}
	if updates.ETA != nil {
		config.ETA = updates.ETA
	}
	if updates.DevelopmentPhase != "" {
		config.DevelopmentPhase = updates.DevelopmentPhase
	}
	if updates.EstimatedEffort != "" {
		config.EstimatedEffort = updates.EstimatedEffort
	}
	if len(updates.Prerequisites) > 0 {
		config.Prerequisites = updates.Prerequisites
	}
	if len(updates.Dependencies) > 0 {
		config.Dependencies = updates.Dependencies
	}
	if len(updates.AcceptanceCriteria) > 0 {
		config.AcceptanceCriteria = updates.AcceptanceCriteria
	}
	if updates.MockDataConfig != nil {
		config.MockDataConfig = updates.MockDataConfig
	}
	if len(updates.EnvironmentConfig) > 0 {
		config.EnvironmentConfig = updates.EnvironmentConfig
	}

	config.UpdatedAt = time.Now()
	return nil
}

// AddConfiguration adds a new feature configuration
func (pcm *PlaceholderConfigManager) AddConfiguration(config *FeatureConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	if config.ID == "" {
		return fmt.Errorf("feature ID is required")
	}

	if _, exists := pcm.configs[config.ID]; exists {
		return fmt.Errorf("configuration for feature %s already exists", config.ID)
	}

	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()
	pcm.configs[config.ID] = config
	return nil
}

// RemoveConfiguration removes a feature configuration
func (pcm *PlaceholderConfigManager) RemoveConfiguration(featureID string) error {
	if _, exists := pcm.configs[featureID]; !exists {
		return fmt.Errorf("configuration for feature %s not found", featureID)
	}

	delete(pcm.configs, featureID)
	return nil
}

// GetAllConfigurations returns all feature configurations
func (pcm *PlaceholderConfigManager) GetAllConfigurations() map[string]*FeatureConfiguration {
	// Return a copy to prevent external modifications
	result := make(map[string]*FeatureConfiguration)
	for id, config := range pcm.configs {
		result[id] = config
	}
	return result
}

// GetConfigurationStatistics returns statistics about configurations
func (pcm *PlaceholderConfigManager) GetConfigurationStatistics() map[string]interface{} {
	stats := make(map[string]interface{})

	// Total configurations
	stats["total_configurations"] = len(pcm.configs)

	// Configurations by status
	statusCount := make(map[FeatureStatus]int)
	for _, config := range pcm.configs {
		statusCount[config.Status]++
	}
	stats["by_status"] = statusCount

	// Configurations by category
	categoryCount := make(map[FeatureCategory]int)
	for _, config := range pcm.configs {
		categoryCount[config.Category]++
	}
	stats["by_category"] = categoryCount

	// Configurations by priority
	priorityCount := make(map[FeaturePriority]int)
	for _, config := range pcm.configs {
		priorityCount[config.Priority]++
	}
	stats["by_priority"] = priorityCount

	// Environment-specific stats
	envStats := make(map[Environment]int)
	for _, config := range pcm.configs {
		if envConfig, exists := config.EnvironmentConfig[pcm.environment]; exists && envConfig.Enabled {
			envStats[pcm.environment]++
		}
	}
	stats["enabled_in_environment"] = envStats

	return stats
}

// ValidateConfiguration validates a feature configuration
func (pcm *PlaceholderConfigManager) ValidateConfiguration(config *FeatureConfiguration) error {
	if config == nil {
		return fmt.Errorf("configuration cannot be nil")
	}

	if config.ID == "" {
		return fmt.Errorf("feature ID is required")
	}

	if config.Name == "" {
		return fmt.Errorf("feature name is required")
	}

	if config.Description == "" {
		return fmt.Errorf("feature description is required")
	}

	if config.Category == "" {
		return fmt.Errorf("feature category is required")
	}

	if config.Priority == 0 {
		return fmt.Errorf("feature priority is required")
	}

	if config.Status == "" {
		return fmt.Errorf("feature status is required")
	}

	// Validate status
	if !isValidFeatureStatus(config.Status) {
		return fmt.Errorf("invalid feature status: %s", config.Status)
	}

	// Validate category
	if !isValidFeatureCategory(config.Category) {
		return fmt.Errorf("invalid feature category: %s", config.Category)
	}

	// Validate priority
	if !isValidFeaturePriority(config.Priority) {
		return fmt.Errorf("invalid feature priority: %d", config.Priority)
	}

	return nil
}

// createDefaultConfigurations creates default feature configurations for the KYB platform
func (pcm *PlaceholderConfigManager) createDefaultConfigurations() {
	defaultConfigs := []*FeatureConfiguration{
		{
			ID:                  "advanced_analytics",
			Name:                "Advanced Analytics Dashboard",
			Description:         "Comprehensive analytics and reporting dashboard with real-time insights",
			DetailedDescription: "A powerful analytics dashboard that provides real-time insights into merchant performance, risk metrics, compliance status, and business intelligence. Features include interactive charts, customizable reports, data export capabilities, and automated alerts.",
			Category:            CategoryAnalytics,
			Priority:            PriorityHigh,
			Status:              StatusComingSoon,
			ETA:                 timePtr(time.Now().Add(30 * 24 * time.Hour)),
			DevelopmentPhase:    "Planning",
			EstimatedEffort:     "4-6 weeks",
			Prerequisites:       []string{"database_optimization", "api_enhancement"},
			Dependencies:        []string{"merchant_portfolio_service"},
			AcceptanceCriteria: []string{
				"Real-time data visualization",
				"Customizable dashboard layouts",
				"Export functionality for reports",
				"Performance metrics tracking",
				"User role-based access control",
			},
			MockDataConfig: &MockDataConfig{
				Enabled:         true,
				DataSize:        "large",
				DataTypes:       []string{"metrics", "charts", "reports"},
				RealisticMode:   true,
				RefreshInterval: 5 * time.Minute,
			},
			EnvironmentConfig: map[Environment]EnvironmentSpecificConfig{
				EnvironmentDevelopment: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "Advanced Analytics - Coming Soon (Dev Environment)",
				},
				EnvironmentStaging: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "Advanced Analytics - Coming Soon (Staging)",
				},
				EnvironmentProduction: {
					Enabled:         false,
					MockDataEnabled: false,
					ShowPlaceholder: true,
					CustomMessage:   "Advanced Analytics - Coming Soon",
				},
			},
		},
		{
			ID:                  "bulk_operations",
			Name:                "Bulk Operations Management",
			Description:         "Perform bulk operations on multiple merchants with progress tracking",
			DetailedDescription: "A comprehensive bulk operations system that allows users to perform various operations on multiple merchants simultaneously. Features include progress tracking, pause/resume functionality, error handling, and detailed operation logs.",
			Category:            CategoryAutomation,
			Priority:            PriorityHigh,
			Status:              StatusInDevelopment,
			ETA:                 timePtr(time.Now().Add(14 * 24 * time.Hour)),
			DevelopmentPhase:    "Implementation",
			EstimatedEffort:     "2-3 weeks",
			Prerequisites:       []string{"merchant_portfolio_service"},
			Dependencies:        []string{"audit_service"},
			AcceptanceCriteria: []string{
				"Select multiple merchants for operations",
				"Progress tracking with real-time updates",
				"Pause and resume functionality",
				"Error handling and rollback",
				"Operation history and logs",
			},
			MockDataConfig: &MockDataConfig{
				Enabled:       true,
				DataSize:      "medium",
				DataTypes:     []string{"operations", "progress", "logs"},
				RealisticMode: true,
			},
			EnvironmentConfig: map[Environment]EnvironmentSpecificConfig{
				EnvironmentDevelopment: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: false,
				},
				EnvironmentStaging: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: false,
				},
				EnvironmentProduction: {
					Enabled:         false,
					MockDataEnabled: false,
					ShowPlaceholder: true,
					CustomMessage:   "Bulk Operations - In Development",
				},
			},
		},
		{
			ID:                  "merchant_comparison",
			Name:                "Merchant Comparison Tool",
			Description:         "Compare up to 2 merchants side-by-side with detailed analysis",
			DetailedDescription: "A sophisticated comparison tool that allows side-by-side analysis of up to 2 merchants. Features include detailed comparison metrics, visual indicators, exportable reports, and comprehensive analysis across multiple dimensions.",
			Category:            CategoryAnalytics,
			Priority:            PriorityMedium,
			Status:              StatusInDevelopment,
			ETA:                 timePtr(time.Now().Add(21 * 24 * time.Hour)),
			DevelopmentPhase:    "Implementation",
			EstimatedEffort:     "3-4 weeks",
			Prerequisites:       []string{"merchant_portfolio_service"},
			Dependencies:        []string{"reporting_service"},
			AcceptanceCriteria: []string{
				"Side-by-side merchant comparison",
				"Visual comparison indicators",
				"Exportable comparison reports",
				"Multiple comparison criteria",
				"Historical comparison data",
			},
			MockDataConfig: &MockDataConfig{
				Enabled:       true,
				DataSize:      "small",
				DataTypes:     []string{"comparison", "metrics", "reports"},
				RealisticMode: true,
			},
			EnvironmentConfig: map[Environment]EnvironmentSpecificConfig{
				EnvironmentDevelopment: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: false,
				},
				EnvironmentStaging: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: false,
				},
				EnvironmentProduction: {
					Enabled:         false,
					MockDataEnabled: false,
					ShowPlaceholder: true,
					CustomMessage:   "Merchant Comparison - In Development",
				},
			},
		},
		{
			ID:                  "external_api_integration",
			Name:                "External API Integration",
			Description:         "Integrate with external data sources and third-party APIs",
			DetailedDescription: "A comprehensive integration system that connects the KYB platform with external data sources, government databases, credit bureaus, and third-party APIs. Features include secure API management, data synchronization, error handling, and compliance tracking.",
			Category:            CategoryIntegration,
			Priority:            PriorityMedium,
			Status:              StatusComingSoon,
			ETA:                 timePtr(time.Now().Add(45 * 24 * time.Hour)),
			DevelopmentPhase:    "Planning",
			EstimatedEffort:     "6-8 weeks",
			Prerequisites:       []string{"security_enhancement", "api_gateway"},
			Dependencies:        []string{"compliance_service"},
			AcceptanceCriteria: []string{
				"Secure API key management",
				"Data synchronization capabilities",
				"Error handling and retry logic",
				"Compliance tracking",
				"Rate limiting and monitoring",
			},
			MockDataConfig: &MockDataConfig{
				Enabled:       true,
				DataSize:      "large",
				DataTypes:     []string{"api_responses", "external_data"},
				RealisticMode: true,
			},
			EnvironmentConfig: map[Environment]EnvironmentSpecificConfig{
				EnvironmentDevelopment: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "External API Integration - Coming Soon (Dev)",
				},
				EnvironmentStaging: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "External API Integration - Coming Soon (Staging)",
				},
				EnvironmentProduction: {
					Enabled:         false,
					MockDataEnabled: false,
					ShowPlaceholder: true,
					CustomMessage:   "External API Integration - Coming Soon",
				},
			},
		},
		{
			ID:                  "automated_reporting",
			Name:                "Automated Reporting System",
			Description:         "Generate and schedule automated compliance and risk reports",
			DetailedDescription: "An automated reporting system that generates and schedules various compliance and risk reports. Features include customizable report templates, automated scheduling, email delivery, and comprehensive audit trails.",
			Category:            CategoryReporting,
			Priority:            PriorityMedium,
			Status:              StatusComingSoon,
			ETA:                 timePtr(time.Now().Add(60 * 24 * time.Hour)),
			DevelopmentPhase:    "Planning",
			EstimatedEffort:     "5-7 weeks",
			Prerequisites:       []string{"reporting_framework", "email_service"},
			Dependencies:        []string{"compliance_service", "audit_service"},
			AcceptanceCriteria: []string{
				"Customizable report templates",
				"Automated scheduling",
				"Email delivery system",
				"Report generation history",
				"Compliance audit trails",
			},
			MockDataConfig: &MockDataConfig{
				Enabled:       true,
				DataSize:      "medium",
				DataTypes:     []string{"reports", "schedules", "templates"},
				RealisticMode: true,
			},
			EnvironmentConfig: map[Environment]EnvironmentSpecificConfig{
				EnvironmentDevelopment: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "Automated Reporting - Coming Soon (Dev)",
				},
				EnvironmentStaging: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "Automated Reporting - Coming Soon (Staging)",
				},
				EnvironmentProduction: {
					Enabled:         false,
					MockDataEnabled: false,
					ShowPlaceholder: true,
					CustomMessage:   "Automated Reporting - Coming Soon",
				},
			},
		},
		{
			ID:                  "real_time_monitoring",
			Name:                "Real-time Monitoring",
			Description:         "Real-time monitoring and alerting for merchant activities",
			DetailedDescription: "A comprehensive real-time monitoring system that tracks merchant activities, system performance, and compliance metrics. Features include real-time dashboards, automated alerting, performance metrics, and incident management.",
			Category:            CategoryMonitoring,
			Priority:            PriorityHigh,
			Status:              StatusComingSoon,
			ETA:                 timePtr(time.Now().Add(90 * 24 * time.Hour)),
			DevelopmentPhase:    "Planning",
			EstimatedEffort:     "8-10 weeks",
			Prerequisites:       []string{"monitoring_infrastructure", "alerting_system"},
			Dependencies:        []string{"performance_service"},
			AcceptanceCriteria: []string{
				"Real-time activity monitoring",
				"Automated alerting system",
				"Performance metrics dashboard",
				"Incident management",
				"Historical data analysis",
			},
			MockDataConfig: &MockDataConfig{
				Enabled:         true,
				DataSize:        "large",
				DataTypes:       []string{"metrics", "alerts", "events"},
				RealisticMode:   true,
				RefreshInterval: 1 * time.Minute,
			},
			EnvironmentConfig: map[Environment]EnvironmentSpecificConfig{
				EnvironmentDevelopment: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "Real-time Monitoring - Coming Soon (Dev)",
				},
				EnvironmentStaging: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "Real-time Monitoring - Coming Soon (Staging)",
				},
				EnvironmentProduction: {
					Enabled:         false,
					MockDataEnabled: false,
					ShowPlaceholder: true,
					CustomMessage:   "Real-time Monitoring - Coming Soon",
				},
			},
		},
		{
			ID:                  "advanced_security",
			Name:                "Advanced Security Features",
			Description:         "Enhanced security features including multi-factor authentication",
			DetailedDescription: "Advanced security features including multi-factor authentication, role-based access control, session management, audit logging, and compliance with security standards. Features include biometric authentication, advanced encryption, and security monitoring.",
			Category:            CategorySecurity,
			Priority:            PriorityCritical,
			Status:              StatusComingSoon,
			ETA:                 timePtr(time.Now().Add(75 * 24 * time.Hour)),
			DevelopmentPhase:    "Planning",
			EstimatedEffort:     "6-8 weeks",
			Prerequisites:       []string{"security_audit", "compliance_review"},
			Dependencies:        []string{"authentication_service"},
			AcceptanceCriteria: []string{
				"Multi-factor authentication",
				"Role-based access control",
				"Advanced session management",
				"Security audit logging",
				"Compliance reporting",
			},
			MockDataConfig: &MockDataConfig{
				Enabled:       true,
				DataSize:      "small",
				DataTypes:     []string{"security_events", "audit_logs"},
				RealisticMode: true,
			},
			EnvironmentConfig: map[Environment]EnvironmentSpecificConfig{
				EnvironmentDevelopment: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "Advanced Security - Coming Soon (Dev)",
				},
				EnvironmentStaging: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "Advanced Security - Coming Soon (Staging)",
				},
				EnvironmentProduction: {
					Enabled:         false,
					MockDataEnabled: false,
					ShowPlaceholder: true,
					CustomMessage:   "Advanced Security - Coming Soon",
				},
			},
		},
		{
			ID:                  "mobile_app",
			Name:                "Mobile Application",
			Description:         "Native mobile application for iOS and Android",
			DetailedDescription: "A native mobile application for iOS and Android that provides full access to the KYB platform features. Features include offline capabilities, push notifications, biometric authentication, and optimized mobile user experience.",
			Category:            CategoryMobile,
			Priority:            PriorityLow,
			Status:              StatusComingSoon,
			ETA:                 timePtr(time.Now().Add(120 * 24 * time.Hour)),
			DevelopmentPhase:    "Planning",
			EstimatedEffort:     "12-16 weeks",
			Prerequisites:       []string{"api_completion", "ui_framework"},
			Dependencies:        []string{"mobile_sdk"},
			AcceptanceCriteria: []string{
				"Native iOS and Android apps",
				"Offline functionality",
				"Push notifications",
				"Biometric authentication",
				"Mobile-optimized UI",
			},
			MockDataConfig: &MockDataConfig{
				Enabled:       true,
				DataSize:      "medium",
				DataTypes:     []string{"mobile_data", "notifications"},
				RealisticMode: true,
			},
			EnvironmentConfig: map[Environment]EnvironmentSpecificConfig{
				EnvironmentDevelopment: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "Mobile App - Coming Soon (Dev)",
				},
				EnvironmentStaging: {
					Enabled:         true,
					MockDataEnabled: true,
					ShowPlaceholder: true,
					CustomMessage:   "Mobile App - Coming Soon (Staging)",
				},
				EnvironmentProduction: {
					Enabled:         false,
					MockDataEnabled: false,
					ShowPlaceholder: true,
					CustomMessage:   "Mobile App - Coming Soon",
				},
			},
		},
	}

	// Add configurations to manager
	for _, config := range defaultConfigs {
		config.CreatedAt = time.Now()
		config.UpdatedAt = time.Now()
		pcm.configs[config.ID] = config
	}
}

// Helper functions for validation
func isValidFeatureStatus(status FeatureStatus) bool {
	switch status {
	case StatusComingSoon, StatusInDevelopment, StatusAvailable, StatusDeprecated:
		return true
	default:
		return false
	}
}

func isValidFeatureCategory(category FeatureCategory) bool {
	switch category {
	case CategoryAnalytics, CategoryReporting, CategoryIntegration, CategoryAutomation,
		CategoryMonitoring, CategorySecurity, CategoryMobile, CategoryUI, CategoryAPI,
		CategoryDatabase, CategoryCompliance, CategoryPerformance:
		return true
	default:
		return false
	}
}

func isValidFeaturePriority(priority FeaturePriority) bool {
	return priority >= PriorityCritical && priority <= PriorityNiceToHave
}
