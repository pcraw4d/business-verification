package shared

import (
	"context"
	"time"
)

// =============================================================================
// Core Classification Interfaces
// =============================================================================

// ClassificationService defines the core interface for business classification
type ClassificationService interface {
	// ClassifyBusiness performs classification for a single business
	ClassifyBusiness(ctx context.Context, req *BusinessClassificationRequest) (*BusinessClassificationResponse, error)

	// ClassifyBusinessesBatch performs batch classification for multiple businesses
	ClassifyBusinessesBatch(ctx context.Context, req *BatchClassificationRequest) (*BatchClassificationResponse, error)

	// GetClassification retrieves a specific classification by ID
	GetClassification(ctx context.Context, id string) (*EnhancedClassification, error)

	// HealthCheck performs a health check on the classification service
	HealthCheck(ctx context.Context) error
}

// ClassificationModule defines the interface for individual classification modules
type ClassificationModule interface {
	// ID returns the unique identifier for this module
	ID() string

	// Metadata returns metadata about this module
	Metadata() ModuleMetadata

	// CanHandle determines if this module can handle the given request
	CanHandle(req *BusinessClassificationRequest) bool

	// Classify performs classification using this module
	Classify(ctx context.Context, req *BusinessClassificationRequest) (*ModuleResult, error)

	// HealthCheck performs a health check on this module
	HealthCheck(ctx context.Context) error
}

// ModuleMetadata represents metadata about a classification module
type ModuleMetadata struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Capabilities []ModuleCapability     `json:"capabilities"`
	Priority     ModulePriority         `json:"priority"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// ModuleCapability represents a capability of a classification module
type ModuleCapability string

const (
	ModuleCapabilityClassification ModuleCapability = "classification"
	ModuleCapabilityWebAnalysis    ModuleCapability = "web_analysis"
	ModuleCapabilityDataExtraction ModuleCapability = "data_extraction"
	ModuleCapabilityML             ModuleCapability = "ml"
	ModuleCapabilityEnsemble       ModuleCapability = "ensemble"
)

// ModulePriority represents the priority of a module
type ModulePriority string

const (
	ModulePriorityHigh   ModulePriority = "high"
	ModulePriorityMedium ModulePriority = "medium"
	ModulePriorityLow    ModulePriority = "low"
)

// =============================================================================
// ML Classification Interfaces
// =============================================================================

// MLClassifier defines the interface for ML-based classification
type MLClassifier interface {
	// Classify performs ML-based classification
	Classify(ctx context.Context, req *MLClassificationRequest) (*MLClassificationResult, error)

	// GetModelInfo returns information about the loaded models
	GetModelInfo(ctx context.Context) ([]*ModelInfo, error)

	// HealthCheck performs a health check on the ML classifier
	HealthCheck(ctx context.Context) error
}

// ModelManager defines the interface for ML model management
type ModelManager interface {
	// LoadModel loads a model by type and version
	LoadModel(ctx context.Context, modelType ModelType, version string) error

	// UnloadModel unloads a model by type and version
	UnloadModel(ctx context.Context, modelType ModelType, version string) error

	// GetModelInfo returns information about a specific model
	GetModelInfo(ctx context.Context, modelType ModelType, version string) (*ModelInfo, error)

	// ListModels returns a list of all available models
	ListModels(ctx context.Context) ([]*ModelInfo, error)

	// UpdateModel updates a model to a new version
	UpdateModel(ctx context.Context, modelType ModelType, currentVersion, newVersion string) error
}

// =============================================================================
// Website Analysis Interfaces
// =============================================================================

// WebsiteAnalyzer defines the interface for website analysis
type WebsiteAnalyzer interface {
	// AnalyzeWebsite performs comprehensive website analysis
	AnalyzeWebsite(ctx context.Context, req *WebsiteAnalysisRequest) (*WebsiteAnalysisResult, error)

	// ValidateConnection validates the connection between business and website
	ValidateConnection(ctx context.Context, businessName, websiteURL string) (*ConnectionValidationResult, error)

	// AnalyzeContent analyzes website content for industry indicators
	AnalyzeContent(ctx context.Context, content string, businessName string) (*ContentAnalysisResult, error)

	// HealthCheck performs a health check on the website analyzer
	HealthCheck(ctx context.Context) error
}

// WebScraper defines the interface for web scraping
type WebScraper interface {
	// ScrapeWebsite scrapes content from a website
	ScrapeWebsite(ctx context.Context, url string, config ScrapingConfig) (*ScrapedContent, error)

	// ScrapeMultiplePages scrapes content from multiple pages
	ScrapeMultiplePages(ctx context.Context, baseURL string, config ScrapingConfig) ([]*ScrapedContent, error)

	// HealthCheck performs a health check on the web scraper
	HealthCheck(ctx context.Context) error
}

// ScrapedContent represents scraped content from a website
type ScrapedContent struct {
	URL        string                 `json:"url"`
	Title      string                 `json:"title"`
	Text       string                 `json:"text"`
	HTML       string                 `json:"html,omitempty"`
	MetaTags   map[string]string      `json:"meta_tags,omitempty"`
	Links      []string               `json:"links,omitempty"`
	Images     []string               `json:"images,omitempty"`
	StatusCode int                    `json:"status_code"`
	ScrapedAt  time.Time              `json:"scraped_at"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ScrapingConfig represents configuration for web scraping
type ScrapingConfig struct {
	Timeout         time.Duration `json:"timeout"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
	MaxConcurrent   int           `json:"max_concurrent"`
	RateLimitPerSec int           `json:"rate_limit_per_sec"`
	UserAgents      []string      `json:"user_agents,omitempty"`
	FollowRedirects bool          `json:"follow_redirects"`
	ExtractLinks    bool          `json:"extract_links"`
	ExtractImages   bool          `json:"extract_images"`
}

// =============================================================================
// Web Search Analysis Interfaces
// =============================================================================

// WebSearchAnalyzer defines the interface for web search analysis
type WebSearchAnalyzer interface {
	// AnalyzeWebSearch performs comprehensive web search analysis
	AnalyzeWebSearch(ctx context.Context, req *WebSearchAnalysisRequest) (*WebSearchAnalysisResult, error)

	// Search performs a basic web search
	Search(ctx context.Context, query string, maxResults int, engines []string) ([]*SearchResult, error)

	// ExtractBusinessInfo extracts business information from search results
	ExtractBusinessInfo(ctx context.Context, results []*SearchResult, businessName string) (*BusinessExtractionResult, error)

	// HealthCheck performs a health check on the web search analyzer
	HealthCheck(ctx context.Context) error
}

// SearchEngine defines the interface for search engines
type SearchEngine interface {
	// Search performs a search using this engine
	Search(ctx context.Context, query string, maxResults int) ([]*SearchResult, error)

	// GetName returns the name of this search engine
	GetName() string

	// GetRateLimit returns the rate limit for this engine
	GetRateLimit() time.Duration

	// HealthCheck performs a health check on this search engine
	HealthCheck(ctx context.Context) error
}

// =============================================================================
// Data Storage Interfaces
// =============================================================================

// ClassificationRepository defines the interface for classification data storage
type ClassificationRepository interface {
	// StoreClassification stores a classification result
	StoreClassification(ctx context.Context, classification *EnhancedClassification) error

	// GetClassification retrieves a classification by ID
	GetClassification(ctx context.Context, id string) (*EnhancedClassification, error)

	// GetClassificationsByBusiness retrieves classifications for a business
	GetClassificationsByBusiness(ctx context.Context, businessName string) ([]*EnhancedClassification, error)

	// UpdateClassification updates an existing classification
	UpdateClassification(ctx context.Context, classification *EnhancedClassification) error

	// DeleteClassification deletes a classification by ID
	DeleteClassification(ctx context.Context, id string) error

	// HealthCheck performs a health check on the repository
	HealthCheck(ctx context.Context) error
}

// FeedbackRepository defines the interface for feedback data storage
type FeedbackRepository interface {
	// StoreFeedback stores feedback data
	StoreFeedback(ctx context.Context, feedback *FeedbackModel) error

	// GetFeedback retrieves feedback by ID
	GetFeedback(ctx context.Context, id string) (*FeedbackModel, error)

	// GetFeedbackByClassification retrieves feedback for a classification
	GetFeedbackByClassification(ctx context.Context, classificationID string) ([]*FeedbackModel, error)

	// UpdateFeedback updates existing feedback
	UpdateFeedback(ctx context.Context, feedback *FeedbackModel) error

	// HealthCheck performs a health check on the repository
	HealthCheck(ctx context.Context) error
}

// =============================================================================
// Validation Interfaces
// =============================================================================

// ValidationService defines the interface for validation services
type ValidationService interface {
	// ValidateClassification validates a classification result
	ValidateClassification(ctx context.Context, classification *IndustryClassification) (*ValidationResult, error)

	// ValidateRequest validates a classification request
	ValidateRequest(ctx context.Context, req *BusinessClassificationRequest) (*ValidationResult, error)

	// HealthCheck performs a health check on the validation service
	HealthCheck(ctx context.Context) error
}

// ValidationResult represents the result of validation
type ValidationResult struct {
	IsValid  bool                   `json:"is_valid"`
	Errors   []ValidationError      `json:"errors,omitempty"`
	Warnings []ValidationWarning    `json:"warnings,omitempty"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// =============================================================================
// Event System Interfaces
// =============================================================================

// EventPublisher defines the interface for publishing events
type EventPublisher interface {
	// PublishEvent publishes an event
	PublishEvent(ctx context.Context, event *ClassificationEvent) error

	// PublishBatch publishes multiple events
	PublishBatch(ctx context.Context, events []*ClassificationEvent) error

	// HealthCheck performs a health check on the event publisher
	HealthCheck(ctx context.Context) error
}

// EventSubscriber defines the interface for subscribing to events
type EventSubscriber interface {
	// Subscribe subscribes to events of a specific type
	Subscribe(ctx context.Context, eventType string, handler EventHandler) error

	// Unsubscribe unsubscribes from events
	Unsubscribe(ctx context.Context, eventType string) error

	// HealthCheck performs a health check on the event subscriber
	HealthCheck(ctx context.Context) error
}

// EventHandler defines the interface for handling events
type EventHandler func(ctx context.Context, event *ClassificationEvent) error

// ClassificationEvent represents a classification-related event
type ClassificationEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Source    string                 `json:"source"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// =============================================================================
// Configuration Interfaces
// =============================================================================

// ConfigurationProvider defines the interface for configuration management
type ConfigurationProvider interface {
	// GetConfiguration retrieves configuration for a module
	GetConfiguration(ctx context.Context, moduleID string) (map[string]interface{}, error)

	// SetConfiguration sets configuration for a module
	SetConfiguration(ctx context.Context, moduleID string, config map[string]interface{}) error

	// GetDefaultConfiguration retrieves default configuration for a module type
	GetDefaultConfiguration(ctx context.Context, moduleType string) (map[string]interface{}, error)

	// HealthCheck performs a health check on the configuration provider
	HealthCheck(ctx context.Context) error
}

// =============================================================================
// Monitoring and Observability Interfaces
// =============================================================================

// MetricsCollector defines the interface for metrics collection
type MetricsCollector interface {
	// RecordClassification records a classification metric
	RecordClassification(ctx context.Context, method string, success bool, duration time.Duration) error

	// RecordBatchClassification records a batch classification metric
	RecordBatchClassification(ctx context.Context, totalCount, successCount int, duration time.Duration) error

	// RecordModuleHealth records module health metrics
	RecordModuleHealth(ctx context.Context, moduleID string, healthy bool) error

	// HealthCheck performs a health check on the metrics collector
	HealthCheck(ctx context.Context) error
}

// Logger defines the interface for logging
type Logger interface {
	// Log logs a message with the given level
	Log(ctx context.Context, level LogLevel, message string, fields map[string]interface{}) error

	// LogClassification logs classification-specific information
	LogClassification(ctx context.Context, req *BusinessClassificationRequest, resp *BusinessClassificationResponse, err error) error

	// HealthCheck performs a health check on the logger
	HealthCheck(ctx context.Context) error
}

// LogLevel represents the level of logging
type LogLevel string

const (
	LogLevelDebug   LogLevel = "debug"
	LogLevelInfo    LogLevel = "info"
	LogLevelWarning LogLevel = "warning"
	LogLevelError   LogLevel = "error"
	LogLevelFatal   LogLevel = "fatal"
)

// =============================================================================
// Utility Interfaces
// =============================================================================

// Cache defines the interface for caching
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) (interface{}, bool, error)

	// Set stores a value in cache
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Clear clears all cached values
	Clear(ctx context.Context) error

	// HealthCheck performs a health check on the cache
	HealthCheck(ctx context.Context) error
}

// RateLimiter defines the interface for rate limiting
type RateLimiter interface {
	// Allow checks if a request is allowed
	Allow(ctx context.Context, key string) (bool, error)

	// Reset resets the rate limit for a key
	Reset(ctx context.Context, key string) error

	// HealthCheck performs a health check on the rate limiter
	HealthCheck(ctx context.Context) error
}

// =============================================================================
// Factory Interfaces
// =============================================================================

// ModuleFactory defines the interface for creating classification modules
type ModuleFactory interface {
	// CreateModule creates a new module instance
	CreateModule(ctx context.Context, moduleType string, config map[string]interface{}) (ClassificationModule, error)

	// GetSupportedModules returns a list of supported module types
	GetSupportedModules(ctx context.Context) ([]string, error)

	// HealthCheck performs a health check on the factory
	HealthCheck(ctx context.Context) error
}

// ServiceFactory defines the interface for creating classification services
type ServiceFactory interface {
	// CreateService creates a new classification service
	CreateService(ctx context.Context, config map[string]interface{}) (ClassificationService, error)

	// HealthCheck performs a health check on the factory
	HealthCheck(ctx context.Context) error
}
