package config

import (
	"fmt"
	"strings"
)

// ConfigValidator provides comprehensive configuration validation
type ConfigValidator struct {
	errors []ValidationError
}

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
	Level   ValidationLevel
}

// ValidationLevel represents the severity of a validation error
type ValidationLevel string

const (
	Error   ValidationLevel = "error"
	Warning ValidationLevel = "warning"
	Info    ValidationLevel = "info"
)

// NewConfigValidator creates a new configuration validator
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{
		errors: make([]ValidationError, 0),
	}
}

// Validate validates the entire enhanced configuration
func (cv *ConfigValidator) Validate(config *EnhancedConfig) []ValidationError {
	cv.errors = make([]ValidationError, 0)

	// Validate base configuration
	cv.validateBaseConfig(config.Config)

	// Validate module configurations
	cv.validateModuleConfigs(&config.Modules)

	// Validate enhanced features
	cv.validateEnhancedFeatures(&config.EnhancedFeatures)

	// Validate performance configuration
	cv.validatePerformanceConfig(&config.Performance)

	// Validate advanced monitoring
	cv.validateAdvancedMonitoring(&config.AdvancedMonitoring)

	// Validate cross-field dependencies
	cv.validateCrossFieldDependencies(config)

	return cv.errors
}

// validateBaseConfig validates the base configuration
func (cv *ConfigValidator) validateBaseConfig(config *Config) {
	// Environment validation
	if config.Environment == "" {
		cv.addError("Environment", config.Environment, "environment is required", Error)
	} else {
		validEnvironments := []string{"development", "staging", "production", "test"}
		if !contains(validEnvironments, config.Environment) {
			cv.addError("Environment", config.Environment,
				fmt.Sprintf("environment must be one of: %s", strings.Join(validEnvironments, ", ")), Error)
		}
	}

	// Server validation
	cv.validateServerConfig(&config.Server)

	// Database validation
	cv.validateDatabaseConfig(&config.Database)

	// Supabase validation
	cv.validateSupabaseConfig(&config.Supabase)

	// Observability validation
	cv.validateObservabilityConfig(&config.Observability)

	// Features validation
	cv.validateFeaturesConfig(&config.Features)
}

// validateServerConfig validates server configuration
func (cv *ConfigValidator) validateServerConfig(server *ServerConfig) {
	if server.Port <= 0 || server.Port > 65535 {
		cv.addError("Server.Port", server.Port, "port must be between 1 and 65535", Error)
	}

	if server.Host == "" {
		cv.addError("Server.Host", server.Host, "host is required", Error)
	}

	if server.ReadTimeout <= 0 {
		cv.addError("Server.ReadTimeout", server.ReadTimeout, "read timeout must be positive", Error)
	}

	if server.WriteTimeout <= 0 {
		cv.addError("Server.WriteTimeout", server.WriteTimeout, "write timeout must be positive", Error)
	}

	if server.IdleTimeout <= 0 {
		cv.addError("Server.IdleTimeout", server.IdleTimeout, "idle timeout must be positive", Error)
	}

	// Validate timeout relationships
	if server.ReadTimeout >= server.IdleTimeout {
		cv.addError("Server.ReadTimeout", server.ReadTimeout,
			"read timeout should be less than idle timeout", Warning)
	}

	if server.WriteTimeout >= server.IdleTimeout {
		cv.addError("Server.WriteTimeout", server.WriteTimeout,
			"write timeout should be less than idle timeout", Warning)
	}
}

// validateDatabaseConfig validates database configuration
func (cv *ConfigValidator) validateDatabaseConfig(db *DatabaseConfig) {
	if db.Host == "" {
		cv.addError("Database.Host", db.Host, "database host is required", Error)
	}

	if db.Port <= 0 || db.Port > 65535 {
		cv.addError("Database.Port", db.Port, "database port must be between 1 and 65535", Error)
	}

	if db.Database == "" {
		cv.addError("Database.Database", db.Database, "database name is required", Error)
	}

	if db.Username == "" {
		cv.addError("Database.Username", db.Username, "database username is required", Error)
	}

	if db.MaxOpenConns <= 0 {
		cv.addError("Database.MaxOpenConns", db.MaxOpenConns, "max open connections must be positive", Error)
	}

	if db.MaxIdleConns <= 0 {
		cv.addError("Database.MaxIdleConns", db.MaxIdleConns, "max idle connections must be positive", Error)
	}

	if db.MaxIdleConns > db.MaxOpenConns {
		cv.addError("Database.MaxIdleConns", db.MaxIdleConns,
			"max idle connections cannot exceed max open connections", Error)
	}

	if db.ConnMaxLifetime <= 0 {
		cv.addError("Database.ConnMaxLifetime", db.ConnMaxLifetime, "connection max lifetime must be positive", Error)
	}

	// Validate SSL mode
	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
	if !contains(validSSLModes, db.SSLMode) {
		cv.addError("Database.SSLMode", db.SSLMode,
			fmt.Sprintf("SSL mode must be one of: %s", strings.Join(validSSLModes, ", ")), Error)
	}
}

// validateSupabaseConfig validates Supabase configuration
func (cv *ConfigValidator) validateSupabaseConfig(supabase *SupabaseConfig) {
	if supabase.URL == "" {
		cv.addError("Supabase.URL", supabase.URL, "Supabase URL is required", Error)
	}

	if supabase.APIKey == "" {
		cv.addError("Supabase.APIKey", supabase.APIKey, "Supabase API key is required", Error)
	}

	if supabase.ServiceRoleKey == "" {
		cv.addError("Supabase.ServiceRoleKey", supabase.ServiceRoleKey, "Supabase service role key is required", Error)
	}
}

// validateObservabilityConfig validates observability configuration
func (cv *ConfigValidator) validateObservabilityConfig(obs *ObservabilityConfig) {
	// Log level validation
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal"}
	if !contains(validLogLevels, obs.LogLevel) {
		cv.addError("Observability.LogLevel", obs.LogLevel,
			fmt.Sprintf("log level must be one of: %s", strings.Join(validLogLevels, ", ")), Error)
	}

	// Log format validation
	validLogFormats := []string{"json", "text"}
	if !contains(validLogFormats, obs.LogFormat) {
		cv.addError("Observability.LogFormat", obs.LogFormat,
			fmt.Sprintf("log format must be one of: %s", strings.Join(validLogFormats, ", ")), Error)
	}

	// Metrics validation
	if obs.MetricsEnabled {
		if obs.MetricsPort <= 0 || obs.MetricsPort > 65535 {
			cv.addError("Observability.MetricsPort", obs.MetricsPort, "metrics port must be between 1 and 65535", Error)
		}

		if obs.MetricsPath == "" {
			cv.addError("Observability.MetricsPath", obs.MetricsPath, "metrics path is required when metrics are enabled", Error)
		}
	}

	// Tracing validation
	if obs.TracingEnabled {
		if obs.TracingURL == "" {
			cv.addError("Observability.TracingURL", obs.TracingURL, "tracing URL is required when tracing is enabled", Error)
		}
	}

	// Health check validation
	if obs.HealthCheckPath == "" {
		cv.addError("Observability.HealthCheckPath", obs.HealthCheckPath, "health check path is required", Error)
	}
}

// validateFeaturesConfig validates features configuration
func (cv *ConfigValidator) validateFeaturesConfig(features *FeaturesConfig) {
	// No specific validation needed for features
	// All fields are boolean and have sensible defaults
}

// validateModuleConfigs validates module configurations
func (cv *ConfigValidator) validateModuleConfigs(modules *ModuleConfigs) {
	// Keyword Classification validation
	cv.validateKeywordClassificationConfig(&modules.KeywordClassification)

	// ML Classification validation
	cv.validateMLClassificationConfig(&modules.MLClassification)

	// Website Analysis validation
	cv.validateWebsiteAnalysisConfig(&modules.WebsiteAnalysis)

	// Web Search Analysis validation
	cv.validateWebSearchAnalysisConfig(&modules.WebSearchAnalysis)

	// Intelligent Router validation
	cv.validateIntelligentRouterConfig(&modules.IntelligentRouter)

	// Resource Manager validation
	cv.validateResourceManagerConfig(&modules.ResourceManager)

	// Data Extraction validation
	cv.validateDataExtractionConfig(&modules.DataExtraction)

	// Risk Assessment validation
	cv.validateRiskAssessmentConfig(&modules.RiskAssessment)

	// Verification validation
	cv.validateVerificationConfig(&modules.Verification)

	// Cache validation
	cv.validateCacheConfig(&modules.Cache)

	// Note: Module observability and security validation would be added when these configs are implemented

	// Compliance validation
	cv.validateComplianceConfig(&modules.Compliance)
}

// validateKeywordClassificationConfig validates keyword classification configuration
func (cv *ConfigValidator) validateKeywordClassificationConfig(config *KeywordClassificationConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.MaxConcurrency <= 0 {
		cv.addError("KeywordClassification.MaxConcurrency", config.MaxConcurrency,
			"max concurrency must be positive", Error)
	}

	if config.Timeout <= 0 {
		cv.addError("KeywordClassification.Timeout", config.Timeout,
			"timeout must be positive", Error)
	}

	if config.ConfidenceThreshold < 0 || config.ConfidenceThreshold > 1 {
		cv.addError("KeywordClassification.ConfidenceThreshold", config.ConfidenceThreshold,
			"confidence threshold must be between 0 and 1", Error)
	}

	if config.MaxKeywords <= 0 {
		cv.addError("KeywordClassification.MaxKeywords", config.MaxKeywords,
			"max keywords must be positive", Error)
	}

	if config.CacheTTL < 0 {
		cv.addError("KeywordClassification.CacheTTL", config.CacheTTL,
			"cache TTL cannot be negative", Error)
	}

	// Validate API keys if external APIs are enabled
	// Note: External API keys validation would be added when the field is implemented
}

// validateMLClassificationConfig validates ML classification configuration
func (cv *ConfigValidator) validateMLClassificationConfig(config *MLClassificationConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.ModelPath == "" {
		cv.addError("MLClassification.ModelPath", config.ModelPath,
			"model path is required when ML classification is enabled", Error)
	}

	if config.MaxConcurrency <= 0 {
		cv.addError("MLClassification.MaxConcurrency", config.MaxConcurrency,
			"max concurrency must be positive", Error)
	}

	if config.Timeout <= 0 {
		cv.addError("MLClassification.Timeout", config.Timeout,
			"timeout must be positive", Error)
	}

	if config.ConfidenceThreshold < 0 || config.ConfidenceThreshold > 1 {
		cv.addError("MLClassification.ConfidenceThreshold", config.ConfidenceThreshold,
			"confidence threshold must be between 0 and 1", Error)
	}

	if config.MaxPredictions <= 0 {
		cv.addError("MLClassification.MaxPredictions", config.MaxPredictions,
			"max predictions must be positive", Error)
	}

	if config.CacheTTL < 0 {
		cv.addError("MLClassification.CacheTTL", config.CacheTTL,
			"cache TTL cannot be negative", Error)
	}

	// Validate batch size
	if config.BatchSize <= 0 {
		cv.addError("MLClassification.BatchSize", config.BatchSize,
			"batch size must be positive", Error)
	}

	if config.BatchSize > 1000 {
		cv.addError("MLClassification.BatchSize", config.BatchSize,
			"batch size should not exceed 1000 for performance reasons", Warning)
	}
}

// validateWebsiteAnalysisConfig validates website analysis configuration
func (cv *ConfigValidator) validateWebsiteAnalysisConfig(config *WebsiteAnalysisConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.Timeout <= 0 {
		cv.addError("WebsiteAnalysis.Timeout", config.Timeout,
			"timeout must be positive", Error)
	}

	if config.MaxContentSize <= 0 {
		cv.addError("WebsiteAnalysis.MaxContentSize", config.MaxContentSize,
			"max content size must be positive", Error)
	}

	if config.MaxContentSize > 100*1024*1024 { // 100MB
		cv.addError("WebsiteAnalysis.MaxContentSize", config.MaxContentSize,
			"max content size should not exceed 100MB for performance reasons", Warning)
	}

	if config.UserAgent == "" {
		cv.addError("WebsiteAnalysis.UserAgent", config.UserAgent,
			"user agent is required", Error)
	}

	if config.CacheTTL < 0 {
		cv.addError("WebsiteAnalysis.CacheTTL", config.CacheTTL,
			"cache TTL cannot be negative", Error)
	}

	// Validate proxy settings
	if config.UseProxy {
		if config.ProxyURL == "" {
			cv.addError("WebsiteAnalysis.ProxyURL", config.ProxyURL,
				"proxy URL is required when proxy is enabled", Error)
		}
	}
}

// validateWebSearchAnalysisConfig validates web search analysis configuration
func (cv *ConfigValidator) validateWebSearchAnalysisConfig(config *WebSearchAnalysisConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.SearchTimeout <= 0 {
		cv.addError("WebSearchAnalysis.SearchTimeout", config.SearchTimeout,
			"search timeout must be positive", Error)
	}

	if config.MaxResults <= 0 {
		cv.addError("WebSearchAnalysis.MaxResults", config.MaxResults,
			"max results must be positive", Error)
	}

	if config.MaxResults > 100 {
		cv.addError("WebSearchAnalysis.MaxResults", config.MaxResults,
			"max results should not exceed 100 for performance reasons", Warning)
	}

	// Validate API keys
	if config.APIKeys == nil || len(config.APIKeys) == 0 {
		cv.addError("WebSearchAnalysis.APIKeys", config.APIKeys,
			"API keys are required", Error)
	}
}

// validateIntelligentRouterConfig validates intelligent router configuration
func (cv *ConfigValidator) validateIntelligentRouterConfig(config *IntelligentRouterConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.MaxConcurrentRequests <= 0 {
		cv.addError("IntelligentRouter.MaxConcurrentRequests", config.MaxConcurrentRequests,
			"max concurrent requests must be positive", Error)
	}

	if config.RequestTimeout <= 0 {
		cv.addError("IntelligentRouter.RequestTimeout", config.RequestTimeout,
			"request timeout must be positive", Error)
	}

	if config.RetryAttempts < 0 {
		cv.addError("IntelligentRouter.RetryAttempts", config.RetryAttempts,
			"retry attempts cannot be negative", Error)
	}

	if config.RetryAttempts > 10 {
		cv.addError("IntelligentRouter.RetryAttempts", config.RetryAttempts,
			"retry attempts should not exceed 10", Warning)
	}

	if config.RetryDelay < 0 {
		cv.addError("IntelligentRouter.RetryDelay", config.RetryDelay,
			"retry delay cannot be negative", Error)
	}

	// Note: Circuit breaker validation would be added when the fields are implemented
}

// validateResourceManagerConfig validates resource manager configuration
func (cv *ConfigValidator) validateResourceManagerConfig(config *ResourceManagerConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.ResourceUpdateInterval <= 0 {
		cv.addError("ResourceManager.ResourceUpdateInterval", config.ResourceUpdateInterval,
			"resource update interval must be positive", Error)
	}

	if config.HealthCheckInterval <= 0 {
		cv.addError("ResourceManager.HealthCheckInterval", config.HealthCheckInterval,
			"health check interval must be positive", Error)
	}

	if config.MaxResourceUtilization <= 0 || config.MaxResourceUtilization > 1 {
		cv.addError("ResourceManager.MaxResourceUtilization", config.MaxResourceUtilization,
			"max resource utilization must be between 0 and 1", Error)
	}

	if config.MinResourceUtilization < 0 || config.MinResourceUtilization > 1 {
		cv.addError("ResourceManager.MinResourceUtilization", config.MinResourceUtilization,
			"min resource utilization must be between 0 and 1", Error)
	}

	if config.MinResourceUtilization >= config.MaxResourceUtilization {
		cv.addError("ResourceManager.MinResourceUtilization", config.MinResourceUtilization,
			"min resource utilization must be less than max resource utilization", Error)
	}

	if config.ScalingThreshold <= 0 || config.ScalingThreshold > 1 {
		cv.addError("ResourceManager.ScalingThreshold", config.ScalingThreshold,
			"scaling threshold must be between 0 and 1", Error)
	}

	// Validate load balancing strategy
	validStrategies := []string{"round_robin", "least_loaded", "best_performance", "adaptive"}
	if !contains(validStrategies, config.LoadBalancingStrategy) {
		cv.addError("ResourceManager.LoadBalancingStrategy", config.LoadBalancingStrategy,
			fmt.Sprintf("load balancing strategy must be one of: %s", strings.Join(validStrategies, ", ")), Error)
	}
}

// validateDataExtractionConfig validates data extraction configuration
func (cv *ConfigValidator) validateDataExtractionConfig(config *DataExtractionConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.MaxExtractionDepth <= 0 {
		cv.addError("DataExtraction.MaxExtractionDepth", config.MaxExtractionDepth,
			"max extraction depth must be positive", Error)
	}

	if config.ExtractionTimeout <= 0 {
		cv.addError("DataExtraction.ExtractionTimeout", config.ExtractionTimeout,
			"extraction timeout must be positive", Error)
	}

	if config.MaxDataPoints <= 0 {
		cv.addError("DataExtraction.MaxDataPoints", config.MaxDataPoints,
			"max data points must be positive", Error)
	}

	if config.MinConfidenceScore < 0 || config.MinConfidenceScore > 1 {
		cv.addError("DataExtraction.MinConfidenceScore", config.MinConfidenceScore,
			"min confidence score must be between 0 and 1", Error)
	}

	if config.CacheTTL < 0 {
		cv.addError("DataExtraction.CacheTTL", config.CacheTTL,
			"cache TTL cannot be negative", Error)
	}
}

// validateRiskAssessmentConfig validates risk assessment configuration
func (cv *ConfigValidator) validateRiskAssessmentConfig(config *RiskAssessmentConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.MaxAssessmentDepth <= 0 {
		cv.addError("RiskAssessment.MaxAssessmentDepth", config.MaxAssessmentDepth,
			"max assessment depth must be positive", Error)
	}

	if config.AssessmentTimeout <= 0 {
		cv.addError("RiskAssessment.AssessmentTimeout", config.AssessmentTimeout,
			"assessment timeout must be positive", Error)
	}

	if config.HighRiskThreshold < 0 || config.HighRiskThreshold > 1 {
		cv.addError("RiskAssessment.HighRiskThreshold", config.HighRiskThreshold,
			"high risk threshold must be between 0 and 1", Error)
	}

	if config.MediumRiskThreshold < 0 || config.MediumRiskThreshold > 1 {
		cv.addError("RiskAssessment.MediumRiskThreshold", config.MediumRiskThreshold,
			"medium risk threshold must be between 0 and 1", Error)
	}

	if config.LowRiskThreshold < 0 || config.LowRiskThreshold > 1 {
		cv.addError("RiskAssessment.LowRiskThreshold", config.LowRiskThreshold,
			"low risk threshold must be between 0 and 1", Error)
	}

	if config.CacheTTL < 0 {
		cv.addError("RiskAssessment.CacheTTL", config.CacheTTL,
			"cache TTL cannot be negative", Error)
	}
}

// validateVerificationConfig validates verification configuration
func (cv *ConfigValidator) validateVerificationConfig(config *VerificationConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.VerificationTimeout <= 0 {
		cv.addError("Verification.VerificationTimeout", config.VerificationTimeout,
			"verification timeout must be positive", Error)
	}

	if config.MaxVerificationAttempts <= 0 {
		cv.addError("Verification.MaxVerificationAttempts", config.MaxVerificationAttempts,
			"max verification attempts must be positive", Error)
	}

	if config.MinVerificationScore < 0 || config.MinVerificationScore > 1 {
		cv.addError("Verification.MinVerificationScore", config.MinVerificationScore,
			"min verification score must be between 0 and 1", Error)
	}

	if config.SuccessRateThreshold < 0 || config.SuccessRateThreshold > 1 {
		cv.addError("Verification.SuccessRateThreshold", config.SuccessRateThreshold,
			"success rate threshold must be between 0 and 1", Error)
	}

	if config.CacheTTL < 0 {
		cv.addError("Verification.CacheTTL", config.CacheTTL,
			"cache TTL cannot be negative", Error)
	}
}

// validateCacheConfig validates cache configuration
func (cv *ConfigValidator) validateCacheConfig(config *CacheConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.DefaultTTL <= 0 {
		cv.addError("Cache.DefaultTTL", config.DefaultTTL,
			"default TTL must be positive", Error)
	}

	if config.MaxSize <= 0 {
		cv.addError("Cache.MaxSize", config.MaxSize,
			"max size must be positive", Error)
	}

	if config.CleanupInterval <= 0 {
		cv.addError("Cache.CleanupInterval", config.CleanupInterval,
			"cleanup interval must be positive", Error)
	}
}

// Note: Module observability and security validation functions would be added when these configs are implemented

// validateComplianceConfig validates compliance configuration
func (cv *ConfigValidator) validateComplianceConfig(config *ComplianceConfig) {
	if !config.Enabled {
		return // Skip validation if disabled
	}

	if config.AuditRetentionDays <= 0 {
		cv.addError("Compliance.AuditRetentionDays", config.AuditRetentionDays,
			"audit retention days must be positive", Error)
	}
}

// validateEnhancedFeatures validates enhanced features configuration
func (cv *ConfigValidator) validateEnhancedFeatures(features *EnhancedFeaturesConfig) {
	// No specific validation needed for enhanced features
	// All fields are boolean and have sensible defaults
}

// validatePerformanceConfig validates performance configuration
func (cv *ConfigValidator) validatePerformanceConfig(performance *PerformanceConfig) {
	if performance.MaxCPUUsage < 0 || performance.MaxCPUUsage > 100 {
		cv.addError("Performance.MaxCPUUsage", performance.MaxCPUUsage,
			"max CPU usage must be between 0 and 100", Error)
	}

	if performance.MaxConcurrentTasks <= 0 {
		cv.addError("Performance.MaxConcurrentTasks", performance.MaxConcurrentTasks,
			"max concurrent tasks must be positive", Error)
	}

	if performance.RequestTimeout <= 0 {
		cv.addError("Performance.RequestTimeout", performance.RequestTimeout,
			"request timeout must be positive", Error)
	}

	if performance.DatabaseTimeout <= 0 {
		cv.addError("Performance.DatabaseTimeout", performance.DatabaseTimeout,
			"database timeout must be positive", Error)
	}

	if performance.ExternalAPITimeout <= 0 {
		cv.addError("Performance.ExternalAPITimeout", performance.ExternalAPITimeout,
			"external API timeout must be positive", Error)
	}

	// Validate timeout relationships
	if performance.DatabaseTimeout >= performance.RequestTimeout {
		cv.addError("Performance.DatabaseTimeout", performance.DatabaseTimeout,
			"database timeout should be less than request timeout", Warning)
	}

	if performance.ExternalAPITimeout >= performance.RequestTimeout {
		cv.addError("Performance.ExternalAPITimeout", performance.ExternalAPITimeout,
			"external API timeout should be less than request timeout", Warning)
	}
}

// validateAdvancedMonitoring validates advanced monitoring configuration
func (cv *ConfigValidator) validateAdvancedMonitoring(monitoring *AdvancedMonitoringConfig) {
	// No specific validation needed for advanced monitoring
	// All fields are boolean and have sensible defaults
}

// validateCrossFieldDependencies validates dependencies between different configuration sections
func (cv *ConfigValidator) validateCrossFieldDependencies(config *EnhancedConfig) {
	// Validate that if ML classification is enabled, the model path exists
	if config.Modules.MLClassification.Enabled {
		if config.Modules.MLClassification.ModelPath == "" {
			cv.addError("MLClassification.ModelPath", config.Modules.MLClassification.ModelPath,
				"model path is required when ML classification is enabled", Error)
		}
	}

	// Validate that if caching is enabled, cache configuration is provided
	if config.Performance.EnableCaching {
		if !config.Modules.Cache.Enabled {
			cv.addError("Cache.Enabled", config.Modules.Cache.Enabled,
				"cache module must be enabled when performance caching is enabled", Error)
		}
	}

	// Note: Cross-field dependency validation for observability and security would be added when these configs are implemented
}

// addError adds a validation error to the list
func (cv *ConfigValidator) addError(field string, value interface{}, message string, level ValidationLevel) {
	cv.errors = append(cv.errors, ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
		Level:   level,
	})
}

// GetErrors returns all validation errors
func (cv *ConfigValidator) GetErrors() []ValidationError {
	return cv.errors
}

// GetErrorsByLevel returns validation errors filtered by level
func (cv *ConfigValidator) GetErrorsByLevel(level ValidationLevel) []ValidationError {
	var filtered []ValidationError
	for _, err := range cv.errors {
		if err.Level == level {
			filtered = append(filtered, err)
		}
	}
	return filtered
}

// HasErrors returns true if there are any validation errors
func (cv *ConfigValidator) HasErrors() bool {
	return len(cv.errors) > 0
}

// HasErrorsByLevel returns true if there are validation errors of the specified level
func (cv *ConfigValidator) HasErrorsByLevel(level ValidationLevel) bool {
	for _, err := range cv.errors {
		if err.Level == level {
			return true
		}
	}
	return false
}

// GetErrorSummary returns a summary of validation errors
func (cv *ConfigValidator) GetErrorSummary() string {
	if len(cv.errors) == 0 {
		return "Configuration is valid"
	}

	errorCount := 0
	warningCount := 0
	infoCount := 0

	for _, err := range cv.errors {
		switch err.Level {
		case Error:
			errorCount++
		case Warning:
			warningCount++
		case Info:
			infoCount++
		}
	}

	summary := fmt.Sprintf("Configuration validation found %d errors, %d warnings, %d info messages",
		errorCount, warningCount, infoCount)

	if errorCount > 0 {
		summary += "\nErrors:"
		for _, err := range cv.GetErrorsByLevel(Error) {
			summary += fmt.Sprintf("\n  %s: %s (value: %v)", err.Field, err.Message, err.Value)
		}
	}

	if warningCount > 0 {
		summary += "\nWarnings:"
		for _, err := range cv.GetErrorsByLevel(Warning) {
			summary += fmt.Sprintf("\n  %s: %s (value: %v)", err.Field, err.Message, err.Value)
		}
	}

	return summary
}

// Helper function to check if a slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
