package config

import (
	"time"
)

// EnhancedConfig extends the base Config with module-specific configurations
type EnhancedConfig struct {
	// Base configuration (inherits from Config)
	*Config

	// Module-specific configurations
	Modules ModuleConfigs `json:"modules" yaml:"modules"`

	// Enhanced features
	EnhancedFeatures EnhancedFeaturesConfig `json:"enhanced_features" yaml:"enhanced_features"`

	// Performance and optimization
	Performance PerformanceConfig `json:"performance" yaml:"performance"`

	// Advanced monitoring
	AdvancedMonitoring AdvancedMonitoringConfig `json:"advanced_monitoring" yaml:"advanced_monitoring"`
}

// ModuleConfigs holds configuration for all modules
type ModuleConfigs struct {
	// Classification modules
	KeywordClassification KeywordClassificationConfig `json:"keyword_classification" yaml:"keyword_classification"`
	MLClassification      MLClassificationConfig      `json:"ml_classification" yaml:"ml_classification"`
	WebsiteAnalysis       WebsiteAnalysisConfig       `json:"website_analysis" yaml:"website_analysis"`
	WebSearchAnalysis     WebSearchAnalysisConfig     `json:"web_search_analysis" yaml:"web_search_analysis"`

	// Routing and orchestration
	IntelligentRouter IntelligentRouterConfig `json:"intelligent_router" yaml:"intelligent_router"`
	ResourceManager   ResourceManagerConfig   `json:"resource_manager" yaml:"resource_manager"`

	// Data processing modules
	DataExtraction DataExtractionConfig `json:"data_extraction" yaml:"data_extraction"`
	RiskAssessment RiskAssessmentConfig `json:"risk_assessment" yaml:"risk_assessment"`
	Verification   VerificationConfig   `json:"verification" yaml:"verification"`

	// Infrastructure modules
	Cache         CacheConfig         `json:"cache" yaml:"cache"`
	Observability ObservabilityConfig `json:"observability" yaml:"observability"`
	Security      SecurityConfig      `json:"security" yaml:"security"`
	Compliance    ComplianceConfig    `json:"compliance" yaml:"compliance"`
}

// KeywordClassificationConfig holds keyword classification module configuration
type KeywordClassificationConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Performance settings
	MaxConcurrency int           `json:"max_concurrency" yaml:"max_concurrency"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
	BatchSize      int           `json:"batch_size" yaml:"batch_size"`

	// Algorithm settings
	ConfidenceThreshold float64 `json:"confidence_threshold" yaml:"confidence_threshold"`
	MaxKeywords         int     `json:"max_keywords" yaml:"max_keywords"`
	MinKeywordLength    int     `json:"min_keyword_length" yaml:"min_keyword_length"`

	// Caching
	EnableCaching bool          `json:"enable_caching" yaml:"enable_caching"`
	CacheTTL      time.Duration `json:"cache_ttl" yaml:"cache_ttl"`

	// Fallback settings
	EnableFallback bool `json:"enable_fallback" yaml:"enable_fallback"`
}

// MLClassificationConfig holds ML classification module configuration
type MLClassificationConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Model settings
	ModelPath           string  `json:"model_path" yaml:"model_path"`
	ModelVersion        string  `json:"model_version" yaml:"model_version"`
	ConfidenceThreshold float64 `json:"confidence_threshold" yaml:"confidence_threshold"`
	MaxPredictions      int     `json:"max_predictions" yaml:"max_predictions"`

	// Performance settings
	MaxConcurrency int           `json:"max_concurrency" yaml:"max_concurrency"`
	Timeout        time.Duration `json:"timeout" yaml:"timeout"`
	BatchSize      int           `json:"batch_size" yaml:"batch_size"`

	// GPU settings
	UseGPU         bool   `json:"use_gpu" yaml:"use_gpu"`
	GPUDeviceID    int    `json:"gpu_device_id" yaml:"gpu_device_id"`
	GPUMemoryLimit string `json:"gpu_memory_limit" yaml:"gpu_memory_limit"`

	// Caching
	EnableCaching bool          `json:"enable_caching" yaml:"enable_caching"`
	CacheTTL      time.Duration `json:"cache_ttl" yaml:"cache_ttl"`

	// Fallback settings
	EnableFallback bool `json:"enable_fallback" yaml:"enable_fallback"`
}

// WebsiteAnalysisConfig holds website analysis module configuration
type WebsiteAnalysisConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// HTTP client settings
	Timeout         time.Duration `json:"timeout" yaml:"timeout"`
	MaxRedirects    int           `json:"max_redirects" yaml:"max_redirects"`
	UserAgent       string        `json:"user_agent" yaml:"user_agent"`
	FollowRedirects bool          `json:"follow_redirects" yaml:"follow_redirects"`

	// Content extraction settings
	MaxContentSize  int64 `json:"max_content_size" yaml:"max_content_size"`
	ExtractImages   bool  `json:"extract_images" yaml:"extract_images"`
	ExtractLinks    bool  `json:"extract_links" yaml:"extract_links"`
	ExtractMetadata bool  `json:"extract_metadata" yaml:"extract_metadata"`

	// Rate limiting
	RateLimitEnabled   bool          `json:"rate_limit_enabled" yaml:"rate_limit_enabled"`
	RateLimitPerMinute int           `json:"rate_limit_per_minute" yaml:"rate_limit_per_minute"`
	RateLimitDelay     time.Duration `json:"rate_limit_delay" yaml:"rate_limit_delay"`

	// Proxy settings
	UseProxy      bool     `json:"use_proxy" yaml:"use_proxy"`
	ProxyURL      string   `json:"proxy_url" yaml:"proxy_url"`
	ProxyRotation bool     `json:"proxy_rotation" yaml:"proxy_rotation"`
	ProxyList     []string `json:"proxy_list" yaml:"proxy_list"`

	// Caching
	EnableCaching bool          `json:"enable_caching" yaml:"enable_caching"`
	CacheTTL      time.Duration `json:"cache_ttl" yaml:"cache_ttl"`

	// Fallback settings
	EnableFallback bool `json:"enable_fallback" yaml:"enable_fallback"`
}

// WebSearchAnalysisConfig holds web search analysis module configuration
type WebSearchAnalysisConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Search engine settings
	SearchEngines []string      `json:"search_engines" yaml:"search_engines"`
	MaxResults    int           `json:"max_results" yaml:"max_results"`
	SearchTimeout time.Duration `json:"search_timeout" yaml:"search_timeout"`

	// API settings
	APIKeys      map[string]string `json:"api_keys" yaml:"api_keys"`
	APIEndpoints map[string]string `json:"api_endpoints" yaml:"api_endpoints"`

	// Rate limiting
	RateLimitEnabled bool          `json:"rate_limit_enabled" yaml:"rate_limit_enabled"`
	RateLimitPerHour int           `json:"rate_limit_per_hour" yaml:"rate_limit_per_hour"`
	RateLimitDelay   time.Duration `json:"rate_limit_delay" yaml:"rate_limit_delay"`

	// Analysis settings
	AnalyzeSentiment bool `json:"analyze_sentiment" yaml:"analyze_sentiment"`
	ExtractKeywords  bool `json:"extract_keywords" yaml:"extract_keywords"`
	DetectLanguage   bool `json:"detect_language" yaml:"detect_language"`

	// Caching
	EnableCaching bool          `json:"enable_caching" yaml:"enable_caching"`
	CacheTTL      time.Duration `json:"cache_ttl" yaml:"cache_ttl"`

	// Fallback settings
	EnableFallback bool `json:"enable_fallback" yaml:"enable_fallback"`
}

// IntelligentRouterConfig holds intelligent router configuration
type IntelligentRouterConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Request analysis
	EnableRequestAnalysis    bool `json:"enable_request_analysis" yaml:"enable_request_analysis"`
	EnableModuleSelection    bool `json:"enable_module_selection" yaml:"enable_module_selection"`
	EnableParallelProcessing bool `json:"enable_parallel_processing" yaml:"enable_parallel_processing"`

	// Performance settings
	MaxConcurrentRequests int           `json:"max_concurrent_requests" yaml:"max_concurrent_requests"`
	MaxParallelModules    int           `json:"max_parallel_modules" yaml:"max_parallel_modules"`
	RequestTimeout        time.Duration `json:"request_timeout" yaml:"request_timeout"`

	// Retry and fallback
	EnableRetryLogic         bool          `json:"enable_retry_logic" yaml:"enable_retry_logic"`
	RetryAttempts            int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay               time.Duration `json:"retry_delay" yaml:"retry_delay"`
	EnableFallbackProcessing bool          `json:"enable_fallback_processing" yaml:"enable_fallback_processing"`

	// Metrics
	EnableMetricsCollection bool `json:"enable_metrics_collection" yaml:"enable_metrics_collection"`
}

// ResourceManagerConfig holds resource manager configuration
type ResourceManagerConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Load balancing
	EnableLoadBalancing   bool   `json:"enable_load_balancing" yaml:"enable_load_balancing"`
	LoadBalancingStrategy string `json:"load_balancing_strategy" yaml:"load_balancing_strategy"`

	// Monitoring
	EnableResourceMonitoring bool `json:"enable_resource_monitoring" yaml:"enable_resource_monitoring"`
	EnableHealthMonitoring   bool `json:"enable_health_monitoring" yaml:"enable_health_monitoring"`
	EnableCapacityPlanning   bool `json:"enable_capacity_planning" yaml:"enable_capacity_planning"`

	// Intervals
	ResourceUpdateInterval   time.Duration `json:"resource_update_interval" yaml:"resource_update_interval"`
	HealthCheckInterval      time.Duration `json:"health_check_interval" yaml:"health_check_interval"`
	CapacityPlanningInterval time.Duration `json:"capacity_planning_interval" yaml:"capacity_planning_interval"`

	// Thresholds
	MaxResourceUtilization float64 `json:"max_resource_utilization" yaml:"max_resource_utilization"`
	MinResourceUtilization float64 `json:"min_resource_utilization" yaml:"min_resource_utilization"`
	ScalingThreshold       float64 `json:"scaling_threshold" yaml:"scaling_threshold"`

	// Scaling
	EnableDynamicScaling       bool `json:"enable_dynamic_scaling" yaml:"enable_dynamic_scaling"`
	EnableResourceOptimization bool `json:"enable_resource_optimization" yaml:"enable_resource_optimization"`
}

// DataExtractionConfig holds data extraction module configuration
type DataExtractionConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Extraction settings
	MaxExtractionDepth int           `json:"max_extraction_depth" yaml:"max_extraction_depth"`
	ExtractionTimeout  time.Duration `json:"extraction_timeout" yaml:"extraction_timeout"`
	MaxDataPoints      int           `json:"max_data_points" yaml:"max_data_points"`

	// Data types to extract
	ExtractContactInfo    bool `json:"extract_contact_info" yaml:"extract_contact_info"`
	ExtractCompanySize    bool `json:"extract_company_size" yaml:"extract_company_size"`
	ExtractBusinessModel  bool `json:"extract_business_model" yaml:"extract_business_model"`
	ExtractGeographicData bool `json:"extract_geographic_data" yaml:"extract_geographic_data"`
	ExtractTechStack      bool `json:"extract_tech_stack" yaml:"extract_tech_stack"`

	// Quality settings
	MinConfidenceScore float64 `json:"min_confidence_score" yaml:"min_confidence_score"`
	EnableValidation   bool    `json:"enable_validation" yaml:"enable_validation"`

	// Caching
	EnableCaching bool          `json:"enable_caching" yaml:"enable_caching"`
	CacheTTL      time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
}

// RiskAssessmentConfig holds risk assessment module configuration
type RiskAssessmentConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Assessment settings
	MaxAssessmentDepth int           `json:"max_assessment_depth" yaml:"max_assessment_depth"`
	AssessmentTimeout  time.Duration `json:"assessment_timeout" yaml:"assessment_timeout"`

	// Risk factors to assess
	AssessSecurityIndicators bool `json:"assess_security_indicators" yaml:"assess_security_indicators"`
	AssessDomainAge          bool `json:"assess_domain_age" yaml:"assess_domain_age"`
	AssessOnlineReputation   bool `json:"assess_online_reputation" yaml:"assess_online_reputation"`
	AssessCompliance         bool `json:"assess_compliance" yaml:"assess_compliance"`
	AssessFinancialHealth    bool `json:"assess_financial_health" yaml:"assess_financial_health"`

	// Thresholds
	HighRiskThreshold   float64 `json:"high_risk_threshold" yaml:"high_risk_threshold"`
	MediumRiskThreshold float64 `json:"medium_risk_threshold" yaml:"medium_risk_threshold"`
	LowRiskThreshold    float64 `json:"low_risk_threshold" yaml:"low_risk_threshold"`

	// External APIs
	EnableExternalAPIs bool `json:"enable_external_apis" yaml:"enable_external_apis"`

	// Caching
	EnableCaching bool          `json:"enable_caching" yaml:"enable_caching"`
	CacheTTL      time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
}

// VerificationConfig holds verification module configuration
type VerificationConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Verification settings
	VerificationTimeout     time.Duration `json:"verification_timeout" yaml:"verification_timeout"`
	MaxVerificationAttempts int           `json:"max_verification_attempts" yaml:"max_verification_attempts"`

	// Verification methods
	EnableWebsiteScraping     bool `json:"enable_website_scraping" yaml:"enable_website_scraping"`
	EnableContactVerification bool `json:"enable_contact_verification" yaml:"enable_contact_verification"`
	EnableDomainVerification  bool `json:"enable_domain_verification" yaml:"enable_domain_verification"`

	// Thresholds
	MinVerificationScore float64 `json:"min_verification_score" yaml:"min_verification_score"`
	SuccessRateThreshold float64 `json:"success_rate_threshold" yaml:"success_rate_threshold"`

	// Fallback strategies
	EnableFallbackStrategies bool `json:"enable_fallback_strategies" yaml:"enable_fallback_strategies"`

	// Caching
	EnableCaching bool          `json:"enable_caching" yaml:"enable_caching"`
	CacheTTL      time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
}

// CacheConfig holds cache configuration
type CacheConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Cache type
	Type string `json:"type" yaml:"type"` // "memory", "redis", "file"

	// Memory cache settings
	MaxSize         int           `json:"max_size" yaml:"max_size"`
	DefaultTTL      time.Duration `json:"default_ttl" yaml:"default_ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`

	// Redis settings
	RedisURL      string `json:"redis_url" yaml:"redis_url"`
	RedisPassword string `json:"redis_password" yaml:"redis_password"`
	RedisDB       int    `json:"redis_db" yaml:"redis_db"`

	// File cache settings
	CacheDir    string `json:"cache_dir" yaml:"cache_dir"`
	MaxFileSize int64  `json:"max_file_size" yaml:"max_file_size"`
	Compression bool   `json:"compression" yaml:"compression"`
}

// ComplianceConfig holds compliance module configuration
type ComplianceConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Compliance frameworks
	EnableSOC2     bool `json:"enable_soc2" yaml:"enable_soc2"`
	EnableGDPR     bool `json:"enable_gdpr" yaml:"enable_gdpr"`
	EnableCCPA     bool `json:"enable_ccpa" yaml:"enable_ccpa"`
	EnableISO27001 bool `json:"enable_iso27001" yaml:"enable_iso27001"`

	// Audit settings
	EnableAuditLogging bool `json:"enable_audit_logging" yaml:"enable_audit_logging"`
	AuditRetentionDays int  `json:"audit_retention_days" yaml:"audit_retention_days"`

	// Data protection
	EnableDataEncryption    bool `json:"enable_data_encryption" yaml:"enable_data_encryption"`
	EnableDataAnonymization bool `json:"enable_data_anonymization" yaml:"enable_data_anonymization"`
}

// EnhancedFeaturesConfig holds enhanced features configuration
type EnhancedFeaturesConfig struct {
	// Dashboard features
	EnableProgressiveDisclosure bool `json:"enable_progressive_disclosure" yaml:"enable_progressive_disclosure"`
	EnableRealTimeUpdates       bool `json:"enable_real_time_updates" yaml:"enable_real_time_updates"`
	EnableCustomDashboards      bool `json:"enable_custom_dashboards" yaml:"enable_custom_dashboards"`

	// API features
	EnableGraphQL          bool `json:"enable_graphql" yaml:"enable_graphql"`
	EnableWebhooks         bool `json:"enable_webhooks" yaml:"enable_webhooks"`
	EnableRateLimiting     bool `json:"enable_rate_limiting" yaml:"enable_rate_limiting"`
	EnableAPIKeyManagement bool `json:"enable_api_key_management" yaml:"enable_api_key_management"`

	// Advanced features
	EnableMachineLearning       bool `json:"enable_machine_learning" yaml:"enable_machine_learning"`
	EnablePredictiveAnalytics   bool `json:"enable_predictive_analytics" yaml:"enable_predictive_analytics"`
	EnableAutomatedOptimization bool `json:"enable_automated_optimization" yaml:"enable_automated_optimization"`
}

// PerformanceConfig holds performance configuration
type PerformanceConfig struct {
	// Optimization settings
	EnableConnectionPooling bool `json:"enable_connection_pooling" yaml:"enable_connection_pooling"`
	EnableQueryOptimization bool `json:"enable_query_optimization" yaml:"enable_query_optimization"`
	EnableCaching           bool `json:"enable_caching" yaml:"enable_caching"`

	// Resource limits
	MaxMemoryUsage     string `json:"max_memory_usage" yaml:"max_memory_usage"`
	MaxCPUUsage        int    `json:"max_cpu_usage" yaml:"max_cpu_usage"`
	MaxConcurrentTasks int    `json:"max_concurrent_tasks" yaml:"max_concurrent_tasks"`

	// Timeouts
	RequestTimeout     time.Duration `json:"request_timeout" yaml:"request_timeout"`
	DatabaseTimeout    time.Duration `json:"database_timeout" yaml:"database_timeout"`
	ExternalAPITimeout time.Duration `json:"external_api_timeout" yaml:"external_api_timeout"`
}

// AdvancedMonitoringConfig holds advanced monitoring configuration
type AdvancedMonitoringConfig struct {
	// Metrics collection
	EnableCustomMetrics      bool `json:"enable_custom_metrics" yaml:"enable_custom_metrics"`
	EnableBusinessMetrics    bool `json:"enable_business_metrics" yaml:"enable_business_metrics"`
	EnablePerformanceMetrics bool `json:"enable_performance_metrics" yaml:"enable_performance_metrics"`

	// Alerting
	EnableAlerting     bool `json:"enable_alerting" yaml:"enable_alerting"`
	EnableEscalation   bool `json:"enable_escalation" yaml:"enable_escalation"`
	EnableNotification bool `json:"enable_notification" yaml:"enable_notification"`

	// Dashboards
	EnableRealTimeDashboards bool `json:"enable_real_time_dashboards" yaml:"enable_real_time_dashboards"`
	EnableHistoricalAnalysis bool `json:"enable_historical_analysis" yaml:"enable_historical_analysis"`
	EnablePredictiveAnalysis bool `json:"enable_predictive_analysis" yaml:"enable_predictive_analysis"`
}
