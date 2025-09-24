package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// EnhancedConfigLoader handles loading and managing enhanced configuration
type EnhancedConfigLoader struct {
	configPath  string
	environment string
	config      *EnhancedConfig
}

// NewEnhancedConfigLoader creates a new enhanced configuration loader
func NewEnhancedConfigLoader(configPath, environment string) *EnhancedConfigLoader {
	return &EnhancedConfigLoader{
		configPath:  configPath,
		environment: environment,
	}
}

// Load loads the configuration from files and environment variables
func (ecl *EnhancedConfigLoader) Load() (*EnhancedConfig, error) {
	// Load base configuration
	baseConfig, err := ecl.loadBaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load base config: %w", err)
	}

	// Create enhanced configuration
	enhancedConfig := &EnhancedConfig{
		Config: baseConfig,
	}

	// Load module configurations
	if err := ecl.loadModuleConfigs(enhancedConfig); err != nil {
		return nil, fmt.Errorf("failed to load module configs: %w", err)
	}

	// Load environment-specific overrides
	if err := ecl.loadEnvironmentOverrides(enhancedConfig); err != nil {
		return nil, fmt.Errorf("failed to load environment overrides: %w", err)
	}

	// Load environment variables
	ecl.loadEnvironmentVariables(enhancedConfig)

	// Validate configuration
	if err := ecl.validateConfig(enhancedConfig); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Set default values
	ecl.setDefaultValues(enhancedConfig)

	ecl.config = enhancedConfig
	return enhancedConfig, nil
}

// loadBaseConfig loads the base configuration
func (ecl *EnhancedConfigLoader) loadBaseConfig() (*Config, error) {
	configPath := filepath.Join(ecl.configPath, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default configuration if file doesn't exist
		return ecl.getDefaultConfig(), nil
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// loadModuleConfigs loads module-specific configurations
func (ecl *EnhancedConfigLoader) loadModuleConfigs(enhancedConfig *EnhancedConfig) error {
	// Load module configs from separate files
	moduleConfigs := []string{
		"modules.yaml",
		"enhanced_features.yaml",
		"performance.yaml",
		"advanced_monitoring.yaml",
	}

	for _, configFile := range moduleConfigs {
		configPath := filepath.Join(ecl.configPath, configFile)

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			continue // Skip if file doesn't exist
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", configFile, err)
		}

		// Parse based on file type
		switch configFile {
		case "modules.yaml":
			if err := yaml.Unmarshal(data, &enhancedConfig.Modules); err != nil {
				return fmt.Errorf("failed to parse modules config: %w", err)
			}
		case "enhanced_features.yaml":
			if err := yaml.Unmarshal(data, &enhancedConfig.EnhancedFeatures); err != nil {
				return fmt.Errorf("failed to parse enhanced features config: %w", err)
			}
		case "performance.yaml":
			if err := yaml.Unmarshal(data, &enhancedConfig.Performance); err != nil {
				return fmt.Errorf("failed to parse performance config: %w", err)
			}
		case "advanced_monitoring.yaml":
			if err := yaml.Unmarshal(data, &enhancedConfig.AdvancedMonitoring); err != nil {
				return fmt.Errorf("failed to parse advanced monitoring config: %w", err)
			}
		}
	}

	return nil
}

// loadEnvironmentOverrides loads environment-specific configuration overrides
func (ecl *EnhancedConfigLoader) loadEnvironmentOverrides(enhancedConfig *EnhancedConfig) error {
	envConfigPath := filepath.Join(ecl.configPath, "environments", fmt.Sprintf("%s.yaml", ecl.environment))

	if _, err := os.Stat(envConfigPath); os.IsNotExist(err) {
		return nil // No environment-specific config
	}

	data, err := os.ReadFile(envConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read environment config: %w", err)
	}

	// Parse environment-specific overrides
	var envOverrides map[string]interface{}
	if err := yaml.Unmarshal(data, &envOverrides); err != nil {
		return fmt.Errorf("failed to parse environment config: %w", err)
	}

	// Apply overrides to enhanced config
	return ecl.applyOverrides(enhancedConfig, envOverrides)
}

// loadEnvironmentVariables loads configuration from environment variables
func (ecl *EnhancedConfigLoader) loadEnvironmentVariables(enhancedConfig *EnhancedConfig) {
	// Load base config environment variables
	ecl.loadBaseConfigEnvVars(enhancedConfig.Config)

	// Load module-specific environment variables
	ecl.loadModuleConfigEnvVars(&enhancedConfig.Modules)

	// Load enhanced features environment variables
	ecl.loadEnhancedFeaturesEnvVars(&enhancedConfig.EnhancedFeatures)

	// Load performance environment variables
	ecl.loadPerformanceEnvVars(&enhancedConfig.Performance)

	// Load advanced monitoring environment variables
	ecl.loadAdvancedMonitoringEnvVars(&enhancedConfig.AdvancedMonitoring)
}

// loadBaseConfigEnvVars loads base configuration from environment variables
func (ecl *EnhancedConfigLoader) loadBaseConfigEnvVars(config *Config) {
	// Environment
	if env := os.Getenv("ENVIRONMENT"); env != "" {
		config.Environment = env
	}

	// Server configuration
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := parseInt(port); err == nil {
			config.Server.Port = p
		}
	}

	if host := os.Getenv("SERVER_HOST"); host != "" {
		config.Server.Host = host
	}

	// Database configuration
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		config.Database.Host = dbHost
	}

	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		if p, err := parseInt(dbPort); err == nil {
			config.Database.Port = p
		}
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		config.Database.Database = dbName
	}

	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		config.Database.Username = dbUser
	}

	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		config.Database.Password = dbPassword
	}

	// Supabase configuration
	if supabaseURL := os.Getenv("SUPABASE_URL"); supabaseURL != "" {
		config.Supabase.URL = supabaseURL
	}

	if supabaseKey := os.Getenv("SUPABASE_API_KEY"); supabaseKey != "" {
		config.Supabase.APIKey = supabaseKey
	}

	if supabaseServiceKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY"); supabaseServiceKey != "" {
		config.Supabase.ServiceRoleKey = supabaseServiceKey
	}
}

// loadModuleConfigEnvVars loads module configuration from environment variables
func (ecl *EnhancedConfigLoader) loadModuleConfigEnvVars(modules *ModuleConfigs) {
	// Keyword Classification
	if enabled := os.Getenv("KEYWORD_CLASSIFICATION_ENABLED"); enabled != "" {
		modules.KeywordClassification.Enabled = parseBool(enabled)
	}

	if maxConcurrency := os.Getenv("KEYWORD_CLASSIFICATION_MAX_CONCURRENCY"); maxConcurrency != "" {
		if mc, err := parseInt(maxConcurrency); err == nil {
			modules.KeywordClassification.MaxConcurrency = mc
		}
	}

	if timeout := os.Getenv("KEYWORD_CLASSIFICATION_TIMEOUT"); timeout != "" {
		if t, err := parseDuration(timeout); err == nil {
			modules.KeywordClassification.Timeout = t
		}
	}

	// ML Classification
	if enabled := os.Getenv("ML_CLASSIFICATION_ENABLED"); enabled != "" {
		modules.MLClassification.Enabled = parseBool(enabled)
	}

	if modelPath := os.Getenv("ML_CLASSIFICATION_MODEL_PATH"); modelPath != "" {
		modules.MLClassification.ModelPath = modelPath
	}

	if useGPU := os.Getenv("ML_CLASSIFICATION_USE_GPU"); useGPU != "" {
		modules.MLClassification.UseGPU = parseBool(useGPU)
	}

	// Website Analysis
	if enabled := os.Getenv("WEBSITE_ANALYSIS_ENABLED"); enabled != "" {
		modules.WebsiteAnalysis.Enabled = parseBool(enabled)
	}

	if timeout := os.Getenv("WEBSITE_ANALYSIS_TIMEOUT"); timeout != "" {
		if t, err := parseDuration(timeout); err == nil {
			modules.WebsiteAnalysis.Timeout = t
		}
	}

	if userAgent := os.Getenv("WEBSITE_ANALYSIS_USER_AGENT"); userAgent != "" {
		modules.WebsiteAnalysis.UserAgent = userAgent
	}

	// Intelligent Router
	if enabled := os.Getenv("INTELLIGENT_ROUTER_ENABLED"); enabled != "" {
		modules.IntelligentRouter.Enabled = parseBool(enabled)
	}

	if maxConcurrentRequests := os.Getenv("INTELLIGENT_ROUTER_MAX_CONCURRENT_REQUESTS"); maxConcurrentRequests != "" {
		if mcr, err := parseInt(maxConcurrentRequests); err == nil {
			modules.IntelligentRouter.MaxConcurrentRequests = mcr
		}
	}

	// Resource Manager
	if enabled := os.Getenv("RESOURCE_MANAGER_ENABLED"); enabled != "" {
		modules.ResourceManager.Enabled = parseBool(enabled)
	}

	if loadBalancingStrategy := os.Getenv("RESOURCE_MANAGER_LOAD_BALANCING_STRATEGY"); loadBalancingStrategy != "" {
		modules.ResourceManager.LoadBalancingStrategy = loadBalancingStrategy
	}
}

// loadEnhancedFeaturesEnvVars loads enhanced features configuration from environment variables
func (ecl *EnhancedConfigLoader) loadEnhancedFeaturesEnvVars(features *EnhancedFeaturesConfig) {
	if progressiveDisclosure := os.Getenv("ENABLE_PROGRESSIVE_DISCLOSURE"); progressiveDisclosure != "" {
		features.EnableProgressiveDisclosure = parseBool(progressiveDisclosure)
	}

	if realTimeUpdates := os.Getenv("ENABLE_REAL_TIME_UPDATES"); realTimeUpdates != "" {
		features.EnableRealTimeUpdates = parseBool(realTimeUpdates)
	}

	if graphQL := os.Getenv("ENABLE_GRAPHQL"); graphQL != "" {
		features.EnableGraphQL = parseBool(graphQL)
	}

	if webhooks := os.Getenv("ENABLE_WEBHOOKS"); webhooks != "" {
		features.EnableWebhooks = parseBool(webhooks)
	}

	if machineLearning := os.Getenv("ENABLE_MACHINE_LEARNING"); machineLearning != "" {
		features.EnableMachineLearning = parseBool(machineLearning)
	}
}

// loadPerformanceEnvVars loads performance configuration from environment variables
func (ecl *EnhancedConfigLoader) loadPerformanceEnvVars(performance *PerformanceConfig) {
	if connectionPooling := os.Getenv("ENABLE_CONNECTION_POOLING"); connectionPooling != "" {
		performance.EnableConnectionPooling = parseBool(connectionPooling)
	}

	if queryOptimization := os.Getenv("ENABLE_QUERY_OPTIMIZATION"); queryOptimization != "" {
		performance.EnableQueryOptimization = parseBool(queryOptimization)
	}

	if maxMemoryUsage := os.Getenv("MAX_MEMORY_USAGE"); maxMemoryUsage != "" {
		performance.MaxMemoryUsage = maxMemoryUsage
	}

	if maxCPUUsage := os.Getenv("MAX_CPU_USAGE"); maxCPUUsage != "" {
		if mcu, err := parseInt(maxCPUUsage); err == nil {
			performance.MaxCPUUsage = mcu
		}
	}

	if requestTimeout := os.Getenv("REQUEST_TIMEOUT"); requestTimeout != "" {
		if rt, err := parseDuration(requestTimeout); err == nil {
			performance.RequestTimeout = rt
		}
	}
}

// loadAdvancedMonitoringEnvVars loads advanced monitoring configuration from environment variables
func (ecl *EnhancedConfigLoader) loadAdvancedMonitoringEnvVars(monitoring *AdvancedMonitoringConfig) {
	if customMetrics := os.Getenv("ENABLE_CUSTOM_METRICS"); customMetrics != "" {
		monitoring.EnableCustomMetrics = parseBool(customMetrics)
	}

	if businessMetrics := os.Getenv("ENABLE_BUSINESS_METRICS"); businessMetrics != "" {
		monitoring.EnableBusinessMetrics = parseBool(businessMetrics)
	}

	if alerting := os.Getenv("ENABLE_ALERTING"); alerting != "" {
		monitoring.EnableAlerting = parseBool(alerting)
	}

	if escalation := os.Getenv("ENABLE_ESCALATION"); escalation != "" {
		monitoring.EnableEscalation = parseBool(escalation)
	}

	if realTimeDashboards := os.Getenv("ENABLE_REAL_TIME_DASHBOARDS"); realTimeDashboards != "" {
		monitoring.EnableRealTimeDashboards = parseBool(realTimeDashboards)
	}
}

// validateConfig validates the configuration
func (ecl *EnhancedConfigLoader) validateConfig(config *EnhancedConfig) error {
	var errors []string

	// Validate base configuration
	if err := ecl.validateBaseConfig(config.Config); err != nil {
		errors = append(errors, fmt.Sprintf("base config: %v", err))
	}

	// Validate module configurations
	if err := ecl.validateModuleConfigs(&config.Modules); err != nil {
		errors = append(errors, fmt.Sprintf("module configs: %v", err))
	}

	// Validate enhanced features
	if err := ecl.validateEnhancedFeatures(&config.EnhancedFeatures); err != nil {
		errors = append(errors, fmt.Sprintf("enhanced features: %v", err))
	}

	// Validate performance configuration
	if err := ecl.validatePerformanceConfig(&config.Performance); err != nil {
		errors = append(errors, fmt.Sprintf("performance config: %v", err))
	}

	// Validate advanced monitoring
	if err := ecl.validateAdvancedMonitoring(&config.AdvancedMonitoring); err != nil {
		errors = append(errors, fmt.Sprintf("advanced monitoring: %v", err))
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// validateBaseConfig validates base configuration
func (ecl *EnhancedConfigLoader) validateBaseConfig(config *Config) error {
	if config.Environment == "" {
		return fmt.Errorf("environment is required")
	}

	if config.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}

	if config.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if config.Database.Port <= 0 {
		return fmt.Errorf("database port must be positive")
	}

	return nil
}

// validateModuleConfigs validates module configurations
func (ecl *EnhancedConfigLoader) validateModuleConfigs(modules *ModuleConfigs) error {
	// Validate keyword classification
	if modules.KeywordClassification.Enabled {
		if modules.KeywordClassification.MaxConcurrency <= 0 {
			return fmt.Errorf("keyword classification max concurrency must be positive")
		}
		if modules.KeywordClassification.ConfidenceThreshold < 0 || modules.KeywordClassification.ConfidenceThreshold > 1 {
			return fmt.Errorf("keyword classification confidence threshold must be between 0 and 1")
		}
	}

	// Validate ML classification
	if modules.MLClassification.Enabled {
		if modules.MLClassification.ModelPath == "" {
			return fmt.Errorf("ML classification model path is required when enabled")
		}
		if modules.MLClassification.ConfidenceThreshold < 0 || modules.MLClassification.ConfidenceThreshold > 1 {
			return fmt.Errorf("ML classification confidence threshold must be between 0 and 1")
		}
	}

	// Validate website analysis
	if modules.WebsiteAnalysis.Enabled {
		if modules.WebsiteAnalysis.Timeout <= 0 {
			return fmt.Errorf("website analysis timeout must be positive")
		}
		if modules.WebsiteAnalysis.MaxContentSize <= 0 {
			return fmt.Errorf("website analysis max content size must be positive")
		}
	}

	// Validate intelligent router
	if modules.IntelligentRouter.Enabled {
		if modules.IntelligentRouter.MaxConcurrentRequests <= 0 {
			return fmt.Errorf("intelligent router max concurrent requests must be positive")
		}
		if modules.IntelligentRouter.RequestTimeout <= 0 {
			return fmt.Errorf("intelligent router request timeout must be positive")
		}
	}

	// Validate resource manager
	if modules.ResourceManager.Enabled {
		if modules.ResourceManager.MaxResourceUtilization <= 0 || modules.ResourceManager.MaxResourceUtilization > 1 {
			return fmt.Errorf("resource manager max resource utilization must be between 0 and 1")
		}
		if modules.ResourceManager.MinResourceUtilization < 0 || modules.ResourceManager.MinResourceUtilization > 1 {
			return fmt.Errorf("resource manager min resource utilization must be between 0 and 1")
		}
	}

	return nil
}

// validateEnhancedFeatures validates enhanced features configuration
func (ecl *EnhancedConfigLoader) validateEnhancedFeatures(features *EnhancedFeaturesConfig) error {
	// No specific validation needed for enhanced features
	return nil
}

// validatePerformanceConfig validates performance configuration
func (ecl *EnhancedConfigLoader) validatePerformanceConfig(performance *PerformanceConfig) error {
	if performance.MaxCPUUsage < 0 || performance.MaxCPUUsage > 100 {
		return fmt.Errorf("max CPU usage must be between 0 and 100")
	}

	if performance.MaxConcurrentTasks <= 0 {
		return fmt.Errorf("max concurrent tasks must be positive")
	}

	if performance.RequestTimeout <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}

	return nil
}

// validateAdvancedMonitoring validates advanced monitoring configuration
func (ecl *EnhancedConfigLoader) validateAdvancedMonitoring(monitoring *AdvancedMonitoringConfig) error {
	// No specific validation needed for advanced monitoring
	return nil
}

// setDefaultValues sets default values for configuration
func (ecl *EnhancedConfigLoader) setDefaultValues(config *EnhancedConfig) {
	// Set default values for modules
	ecl.setModuleDefaultValues(&config.Modules)

	// Set default values for enhanced features
	ecl.setEnhancedFeaturesDefaultValues(&config.EnhancedFeatures)

	// Set default values for performance
	ecl.setPerformanceDefaultValues(&config.Performance)

	// Set default values for advanced monitoring
	ecl.setAdvancedMonitoringDefaultValues(&config.AdvancedMonitoring)
}

// setModuleDefaultValues sets default values for modules
func (ecl *EnhancedConfigLoader) setModuleDefaultValues(modules *ModuleConfigs) {
	// Keyword Classification defaults
	if modules.KeywordClassification.MaxConcurrency == 0 {
		modules.KeywordClassification.MaxConcurrency = 10
	}
	if modules.KeywordClassification.Timeout == 0 {
		modules.KeywordClassification.Timeout = 30 * time.Second
	}
	if modules.KeywordClassification.ConfidenceThreshold == 0 {
		modules.KeywordClassification.ConfidenceThreshold = 0.7
	}
	if modules.KeywordClassification.CacheTTL == 0 {
		modules.KeywordClassification.CacheTTL = 1 * time.Hour
	}

	// ML Classification defaults
	if modules.MLClassification.MaxConcurrency == 0 {
		modules.MLClassification.MaxConcurrency = 5
	}
	if modules.MLClassification.Timeout == 0 {
		modules.MLClassification.Timeout = 60 * time.Second
	}
	if modules.MLClassification.ConfidenceThreshold == 0 {
		modules.MLClassification.ConfidenceThreshold = 0.8
	}
	if modules.MLClassification.CacheTTL == 0 {
		modules.MLClassification.CacheTTL = 2 * time.Hour
	}

	// Website Analysis defaults
	if modules.WebsiteAnalysis.Timeout == 0 {
		modules.WebsiteAnalysis.Timeout = 30 * time.Second
	}
	if modules.WebsiteAnalysis.MaxContentSize == 0 {
		modules.WebsiteAnalysis.MaxContentSize = 10 * 1024 * 1024 // 10MB
	}
	if modules.WebsiteAnalysis.CacheTTL == 0 {
		modules.WebsiteAnalysis.CacheTTL = 1 * time.Hour
	}

	// Intelligent Router defaults
	if modules.IntelligentRouter.MaxConcurrentRequests == 0 {
		modules.IntelligentRouter.MaxConcurrentRequests = 100
	}
	if modules.IntelligentRouter.RequestTimeout == 0 {
		modules.IntelligentRouter.RequestTimeout = 60 * time.Second
	}
	if modules.IntelligentRouter.RetryAttempts == 0 {
		modules.IntelligentRouter.RetryAttempts = 3
	}
	if modules.IntelligentRouter.RetryDelay == 0 {
		modules.IntelligentRouter.RetryDelay = 1 * time.Second
	}

	// Resource Manager defaults
	if modules.ResourceManager.ResourceUpdateInterval == 0 {
		modules.ResourceManager.ResourceUpdateInterval = 30 * time.Second
	}
	if modules.ResourceManager.HealthCheckInterval == 0 {
		modules.ResourceManager.HealthCheckInterval = 60 * time.Second
	}
	if modules.ResourceManager.MaxResourceUtilization == 0 {
		modules.ResourceManager.MaxResourceUtilization = 0.8
	}
	if modules.ResourceManager.MinResourceUtilization == 0 {
		modules.ResourceManager.MinResourceUtilization = 0.2
	}
}

// setEnhancedFeaturesDefaultValues sets default values for enhanced features
func (ecl *EnhancedConfigLoader) setEnhancedFeaturesDefaultValues(features *EnhancedFeaturesConfig) {
	// Enable progressive disclosure by default
	features.EnableProgressiveDisclosure = true

	// Enable rate limiting by default
	features.EnableRateLimiting = true
}

// setPerformanceDefaultValues sets default values for performance
func (ecl *EnhancedConfigLoader) setPerformanceDefaultValues(performance *PerformanceConfig) {
	// Enable connection pooling by default
	performance.EnableConnectionPooling = true

	// Enable caching by default
	performance.EnableCaching = true

	// Set default memory usage
	if performance.MaxMemoryUsage == "" {
		performance.MaxMemoryUsage = "512MB"
	}

	// Set default CPU usage
	if performance.MaxCPUUsage == 0 {
		performance.MaxCPUUsage = 80
	}

	// Set default concurrent tasks
	if performance.MaxConcurrentTasks == 0 {
		performance.MaxConcurrentTasks = 50
	}

	// Set default timeouts
	if performance.RequestTimeout == 0 {
		performance.RequestTimeout = 30 * time.Second
	}
	if performance.DatabaseTimeout == 0 {
		performance.DatabaseTimeout = 10 * time.Second
	}
	if performance.ExternalAPITimeout == 0 {
		performance.ExternalAPITimeout = 15 * time.Second
	}
}

// setAdvancedMonitoringDefaultValues sets default values for advanced monitoring
func (ecl *EnhancedConfigLoader) setAdvancedMonitoringDefaultValues(monitoring *AdvancedMonitoringConfig) {
	// Enable custom metrics by default
	monitoring.EnableCustomMetrics = true

	// Enable business metrics by default
	monitoring.EnableBusinessMetrics = true

	// Enable performance metrics by default
	monitoring.EnablePerformanceMetrics = true

	// Enable alerting by default
	monitoring.EnableAlerting = true
}

// applyOverrides applies configuration overrides
func (ecl *EnhancedConfigLoader) applyOverrides(config *EnhancedConfig, overrides map[string]interface{}) error {
	// This is a simplified implementation
	// In a real implementation, you would use reflection or a more sophisticated approach
	// to apply overrides to nested structures

	// For now, we'll just log that overrides were found
	fmt.Printf("Found %d configuration overrides for environment %s\n", len(overrides), ecl.environment)

	return nil
}

// getDefaultConfig returns a default configuration
func (ecl *EnhancedConfigLoader) getDefaultConfig() *Config {
	return &Config{
		Environment: ecl.environment,
		Server: ServerConfig{
			Port:         8080,
			Host:         "localhost",
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			Driver:          "postgres",
			Host:            "localhost",
			Port:            5432,
			Database:        "business_verification",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    25,
			ConnMaxLifetime: 5 * time.Minute,
			AutoMigrate:     true,
		},
		Observability: ObservabilityConfig{
			LogLevel:        "info",
			LogFormat:       "json",
			MetricsEnabled:  true,
			MetricsPort:     9090,
			MetricsPath:     "/metrics",
			TracingEnabled:  true,
			TracingURL:      "http://localhost:14268/api/traces",
			HealthCheckPath: "/health",
		},
		Features: FeaturesConfig{
			BusinessClassification: true,
			RiskAssessment:         true,
			ComplianceFramework:    true,
			AdvancedAnalytics:      true,
			RealTimeMonitoring:     true,
		},
	}
}

// Helper functions for parsing environment variables
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func parseBool(s string) bool {
	return strings.ToLower(s) == "true" || s == "1"
}

func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}
