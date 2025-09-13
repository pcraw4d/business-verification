package business_intelligence

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DataCollectionPipeline manages the collection of business intelligence data from various sources
type DataCollectionPipeline struct {
	config       DataCollectionConfig
	logger       *zap.Logger
	dataSources  map[string]DataSource
	processors   map[string]DataProcessor
	aggregators  map[string]DataAggregator
	validators   map[string]DataValidator
	cache        DataCache
	rateLimiters map[string]*RateLimiter
	mu           sync.RWMutex
	metrics      *PipelineMetrics
}

// DataCollectionConfig holds configuration for the data collection pipeline
type DataCollectionConfig struct {
	// Pipeline configuration
	MaxConcurrentCollections int           `json:"max_concurrent_collections"`
	CollectionTimeout        time.Duration `json:"collection_timeout"`
	RetryAttempts            int           `json:"retry_attempts"`
	RetryDelay               time.Duration `json:"retry_delay"`

	// Data source configuration
	DataSources              map[string]DataSourceConfig `json:"data_sources"`
	DefaultDataSource        string                      `json:"default_data_source"`
	FallbackDataSource       string                      `json:"fallback_data_source"`
	EnableDataSourceFailover bool                        `json:"enable_data_source_failover"`

	// Processing configuration
	EnableParallelProcessing bool          `json:"enable_parallel_processing"`
	ProcessingTimeout        time.Duration `json:"processing_timeout"`
	EnableDataValidation     bool          `json:"enable_data_validation"`
	EnableDataEnrichment     bool          `json:"enable_data_enrichment"`

	// Caching configuration
	EnableCaching      bool          `json:"enable_caching"`
	CacheTTL           time.Duration `json:"cache_ttl"`
	CacheSize          int           `json:"cache_size"`
	EnableCacheWarming bool          `json:"enable_cache_warming"`

	// Rate limiting
	EnableRateLimiting bool `json:"enable_rate_limiting"`
	GlobalRateLimit    int  `json:"global_rate_limit"`
	PerSourceRateLimit int  `json:"per_source_rate_limit"`

	// Monitoring and alerting
	EnableMonitoring          bool          `json:"enable_monitoring"`
	EnableAlerting            bool          `json:"enable_alerting"`
	HealthCheckInterval       time.Duration `json:"health_check_interval"`
	MetricsCollectionInterval time.Duration `json:"metrics_collection_interval"`

	// Data quality
	QualityThreshold     float64 `json:"quality_threshold"`
	EnableQualityScoring bool    `json:"enable_quality_scoring"`
	EnableDuplicateCheck bool    `json:"enable_duplicate_check"`
}

// DataSourceConfig holds configuration for a specific data source
type DataSourceConfig struct {
	Name          string             `json:"name"`
	Type          string             `json:"type"` // api, database, file, web_scraping
	Endpoint      string             `json:"endpoint"`
	Credentials   map[string]string  `json:"credentials"`
	RateLimit     int                `json:"rate_limit"`
	Timeout       time.Duration      `json:"timeout"`
	RetryAttempts int                `json:"retry_attempts"`
	Priority      int                `json:"priority"`
	Enabled       bool               `json:"enabled"`
	Quality       float64            `json:"quality"`
	Coverage      map[string]float64 `json:"coverage"`
}

// DataSource represents a data source for business intelligence collection
type DataSource interface {
	GetName() string
	GetType() string
	GetConfig() DataSourceConfig
	IsHealthy() bool
	GetQuality() float64
	GetCoverage() map[string]float64
	CollectData(ctx context.Context, request DataCollectionRequest) (*DataCollectionResult, error)
	ValidateData(data *RawData) (*DataValidationResult, error)
}

// DataProcessor processes raw data from data sources
type DataProcessor interface {
	GetName() string
	GetType() string
	ProcessData(ctx context.Context, rawData *RawData) (*ProcessedData, error)
	ValidateProcessedData(data *ProcessedData) (*DataValidationResult, error)
}

// DataAggregator aggregates processed data from multiple sources
type DataAggregator interface {
	GetName() string
	GetType() string
	AggregateData(ctx context.Context, processedData []*ProcessedData) (*AggregatedData, error)
	ValidateAggregatedData(data *AggregatedData) (*DataValidationResult, error)
}

// DataValidator validates data at various stages of the pipeline
type DataValidator interface {
	GetName() string
	Validate(ctx context.Context, data interface{}) (*DataValidationResult, error)
}

// DataCache provides caching functionality for the pipeline
type DataCache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration)
	Delete(key string)
	Clear()
	Size() int
}

// RateLimiter manages rate limiting for data sources
type RateLimiter struct {
	name       string
	rateLimit  int
	burstLimit int
	tokens     int
	lastRefill time.Time
	mu         sync.Mutex
}

// DataCollectionRequest represents a request for data collection
type DataCollectionRequest struct {
	ID          string                 `json:"id"`
	BusinessID  string                 `json:"business_id"`
	DataType    string                 `json:"data_type"`
	DataSources []string               `json:"data_sources"`
	Parameters  map[string]interface{} `json:"parameters"`
	Options     CollectionOptions      `json:"options"`
	Priority    int                    `json:"priority"`
	RequestedAt time.Time              `json:"requested_at"`
	Timeout     time.Duration          `json:"timeout"`
}

// CollectionOptions holds options for data collection
type CollectionOptions struct {
	RealTime         bool `json:"real_time"`
	BatchMode        bool `json:"batch_mode"`
	Parallel         bool `json:"parallel"`
	ValidateData     bool `json:"validate_data"`
	EnrichData       bool `json:"enrich_data"`
	CacheResults     bool `json:"cache_results"`
	EnableFallback   bool `json:"enable_fallback"`
	EnableRetry      bool `json:"enable_retry"`
	EnableMonitoring bool `json:"enable_monitoring"`
}

// DataCollectionResult represents the result of a data collection operation
type DataCollectionResult struct {
	ID                string                 `json:"id"`
	RequestID         string                 `json:"request_id"`
	BusinessID        string                 `json:"business_id"`
	DataType          string                 `json:"data_type"`
	DataSources       []string               `json:"data_sources"`
	RawData           []*RawData             `json:"raw_data"`
	ProcessedData     []*ProcessedData       `json:"processed_data"`
	AggregatedData    *AggregatedData        `json:"aggregated_data"`
	Metadata          map[string]interface{} `json:"metadata"`
	QualityScore      float64                `json:"quality_score"`
	CompletenessScore float64                `json:"completeness_score"`
	ConsistencyScore  float64                `json:"consistency_score"`
	CollectedAt       time.Time              `json:"collected_at"`
	ProcessingTime    time.Duration          `json:"processing_time"`
	Status            CollectionStatus       `json:"status"`
	Errors            []CollectionError      `json:"errors"`
}

// RawData represents raw data collected from a data source
type RawData struct {
	ID           string                 `json:"id"`
	SourceID     string                 `json:"source_id"`
	SourceName   string                 `json:"source_name"`
	DataType     string                 `json:"data_type"`
	Content      interface{}            `json:"content"`
	Metadata     map[string]interface{} `json:"metadata"`
	QualityScore float64                `json:"quality_score"`
	CollectedAt  time.Time              `json:"collected_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
}

// ProcessedData represents processed data
type ProcessedData struct {
	ID           string                 `json:"id"`
	RawDataID    string                 `json:"raw_data_id"`
	ProcessorID  string                 `json:"processor_id"`
	DataType     string                 `json:"data_type"`
	Content      interface{}            `json:"content"`
	Metadata     map[string]interface{} `json:"metadata"`
	QualityScore float64                `json:"quality_score"`
	ProcessedAt  time.Time              `json:"processed_at"`
	ExpiresAt    time.Time              `json:"expires_at"`
}

// AggregatedData represents aggregated data from multiple sources
type AggregatedData struct {
	ID                string                 `json:"id"`
	AggregatorID      string                 `json:"aggregator_id"`
	DataType          string                 `json:"data_type"`
	Content           interface{}            `json:"content"`
	Metadata          map[string]interface{} `json:"metadata"`
	QualityScore      float64                `json:"quality_score"`
	CompletenessScore float64                `json:"completeness_score"`
	ConsistencyScore  float64                `json:"consistency_score"`
	SourceCount       int                    `json:"source_count"`
	AggregatedAt      time.Time              `json:"aggregated_at"`
	ExpiresAt         time.Time              `json:"expires_at"`
}

// DataValidationResult represents the result of data validation
type DataValidationResult struct {
	IsValid         bool                       `json:"is_valid"`
	QualityScore    float64                    `json:"quality_score"`
	Issues          []ValidationIssue          `json:"issues"`
	Recommendations []ValidationRecommendation `json:"recommendations"`
	ValidatedAt     time.Time                  `json:"validated_at"`
}

// ValidationIssue represents a data validation issue
type ValidationIssue struct {
	Type        string      `json:"type"`
	Severity    string      `json:"severity"`
	Description string      `json:"description"`
	Field       string      `json:"field"`
	Value       interface{} `json:"value"`
}

// ValidationRecommendation represents a recommendation for data improvement
type ValidationRecommendation struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
	Action      string `json:"action"`
}

// CollectionStatus represents the status of a data collection operation
type CollectionStatus string

const (
	CollectionStatusPending   CollectionStatus = "pending"
	CollectionStatusRunning   CollectionStatus = "running"
	CollectionStatusCompleted CollectionStatus = "completed"
	CollectionStatusFailed    CollectionStatus = "failed"
	CollectionStatusCancelled CollectionStatus = "cancelled"
	CollectionStatusTimeout   CollectionStatus = "timeout"
)

// CollectionError represents an error during data collection
type CollectionError struct {
	Source    string    `json:"source"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Retryable bool      `json:"retryable"`
}

// PipelineMetrics tracks metrics for the data collection pipeline
type PipelineMetrics struct {
	TotalCollections      int64                        `json:"total_collections"`
	SuccessfulCollections int64                        `json:"successful_collections"`
	FailedCollections     int64                        `json:"failed_collections"`
	AverageProcessingTime time.Duration                `json:"average_processing_time"`
	DataQualityScores     map[string]float64           `json:"data_quality_scores"`
	SourcePerformance     map[string]SourcePerformance `json:"source_performance"`
	LastUpdated           time.Time                    `json:"last_updated"`
}

// SourcePerformance tracks performance metrics for a data source
type SourcePerformance struct {
	TotalRequests       int64         `json:"total_requests"`
	SuccessfulRequests  int64         `json:"successful_requests"`
	FailedRequests      int64         `json:"failed_requests"`
	AverageResponseTime time.Duration `json:"average_response_time"`
	QualityScore        float64       `json:"quality_score"`
	LastRequest         time.Time     `json:"last_request"`
}

// NewDataCollectionPipeline creates a new data collection pipeline
func NewDataCollectionPipeline(config DataCollectionConfig, logger *zap.Logger) *DataCollectionPipeline {
	return &DataCollectionPipeline{
		config:       config,
		logger:       logger,
		dataSources:  make(map[string]DataSource),
		processors:   make(map[string]DataProcessor),
		aggregators:  make(map[string]DataAggregator),
		validators:   make(map[string]DataValidator),
		rateLimiters: make(map[string]*RateLimiter),
		metrics: &PipelineMetrics{
			DataQualityScores: make(map[string]float64),
			SourcePerformance: make(map[string]SourcePerformance),
		},
	}
}

// RegisterDataSource registers a data source with the pipeline
func (p *DataCollectionPipeline) RegisterDataSource(source DataSource) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	name := source.GetName()
	p.dataSources[name] = source

	// Create rate limiter for the source
	config := source.GetConfig()
	rateLimiter := NewRateLimiter(name, config.RateLimit, config.RateLimit*2)
	p.rateLimiters[name] = rateLimiter

	p.logger.Info("Registered data source",
		zap.String("name", name),
		zap.String("type", source.GetType()))

	return nil
}

// RegisterDataProcessor registers a data processor with the pipeline
func (p *DataCollectionPipeline) RegisterDataProcessor(processor DataProcessor) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	name := processor.GetName()
	p.processors[name] = processor

	p.logger.Info("Registered data processor",
		zap.String("name", name),
		zap.String("type", processor.GetType()))

	return nil
}

// RegisterDataAggregator registers a data aggregator with the pipeline
func (p *DataCollectionPipeline) RegisterDataAggregator(aggregator DataAggregator) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	name := aggregator.GetName()
	p.aggregators[name] = aggregator

	p.logger.Info("Registered data aggregator",
		zap.String("name", name),
		zap.String("type", aggregator.GetType()))

	return nil
}

// RegisterDataValidator registers a data validator with the pipeline
func (p *DataCollectionPipeline) RegisterDataValidator(validator DataValidator) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	name := validator.GetName()
	p.validators[name] = validator

	p.logger.Info("Registered data validator",
		zap.String("name", name))

	return nil
}

// CollectData collects data from specified sources
func (p *DataCollectionPipeline) CollectData(ctx context.Context, request DataCollectionRequest) (*DataCollectionResult, error) {
	startTime := time.Now()

	p.logger.Info("Starting data collection",
		zap.String("request_id", request.ID),
		zap.String("business_id", request.BusinessID),
		zap.String("data_type", request.DataType),
		zap.Strings("data_sources", request.DataSources))

	// Create result
	result := &DataCollectionResult{
		ID:          generateID(),
		RequestID:   request.ID,
		BusinessID:  request.BusinessID,
		DataType:    request.DataType,
		DataSources: request.DataSources,
		Metadata:    make(map[string]interface{}),
		CollectedAt: time.Now(),
		Status:      CollectionStatusRunning,
		Errors:      []CollectionError{},
	}

	// Set timeout
	if request.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, request.Timeout)
		defer cancel()
	}

	// Collect raw data from sources
	rawData, err := p.collectRawData(ctx, request)
	if err != nil {
		result.Status = CollectionStatusFailed
		result.Errors = append(result.Errors, CollectionError{
			Source:    "pipeline",
			Type:      "collection_error",
			Message:   err.Error(),
			Timestamp: time.Now(),
			Retryable: true,
		})
		return result, fmt.Errorf("failed to collect raw data: %w", err)
	}
	result.RawData = rawData

	// Process data if processors are available
	if len(p.processors) > 0 {
		processedData, err := p.processData(ctx, rawData)
		if err != nil {
			p.logger.Warn("Data processing failed, continuing with raw data",
				zap.Error(err))
			result.Errors = append(result.Errors, CollectionError{
				Source:    "pipeline",
				Type:      "processing_error",
				Message:   err.Error(),
				Timestamp: time.Now(),
				Retryable: true,
			})
		} else {
			result.ProcessedData = processedData
		}
	}

	// Aggregate data if aggregators are available
	if len(p.aggregators) > 0 && len(result.ProcessedData) > 0 {
		aggregatedData, err := p.aggregateData(ctx, result.ProcessedData)
		if err != nil {
			p.logger.Warn("Data aggregation failed",
				zap.Error(err))
			result.Errors = append(result.Errors, CollectionError{
				Source:    "pipeline",
				Type:      "aggregation_error",
				Message:   err.Error(),
				Timestamp: time.Now(),
				Retryable: true,
			})
		} else {
			result.AggregatedData = aggregatedData
		}
	}

	// Calculate quality scores
	result.QualityScore = p.calculateQualityScore(result)
	result.CompletenessScore = p.calculateCompletenessScore(result)
	result.ConsistencyScore = p.calculateConsistencyScore(result)

	// Update processing time
	result.ProcessingTime = time.Since(startTime)
	result.Status = CollectionStatusCompleted

	// Update metrics
	p.updateMetrics(result)

	p.logger.Info("Data collection completed",
		zap.String("request_id", request.ID),
		zap.Duration("processing_time", result.ProcessingTime),
		zap.Float64("quality_score", result.QualityScore),
		zap.Int("raw_data_count", len(result.RawData)),
		zap.Int("processed_data_count", len(result.ProcessedData)))

	return result, nil
}

// collectRawData collects raw data from specified sources
func (p *DataCollectionPipeline) collectRawData(ctx context.Context, request DataCollectionRequest) ([]*RawData, error) {
	var rawData []*RawData
	var errors []CollectionError

	// Determine which sources to use
	sources := p.selectDataSources(request)
	if len(sources) == 0 {
		return nil, fmt.Errorf("no suitable data sources found for request")
	}

	// Collect data from sources
	if request.Options.Parallel && len(sources) > 1 {
		rawData, errors = p.collectDataParallel(ctx, request, sources)
	} else {
		rawData, errors = p.collectDataSequential(ctx, request, sources)
	}

	// Log errors but continue with available data
	if len(errors) > 0 {
		p.logger.Warn("Some data sources failed",
			zap.Int("error_count", len(errors)),
			zap.Int("successful_sources", len(rawData)))
	}

	if len(rawData) == 0 {
		return nil, fmt.Errorf("no data collected from any source")
	}

	return rawData, nil
}

// selectDataSources selects appropriate data sources for the request
func (p *DataCollectionPipeline) selectDataSources(request DataCollectionRequest) []DataSource {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var selectedSources []DataSource

	// If specific sources are requested, use them
	if len(request.DataSources) > 0 {
		for _, sourceName := range request.DataSources {
			if source, exists := p.dataSources[sourceName]; exists && source.IsHealthy() {
				selectedSources = append(selectedSources, source)
			}
		}
	} else {
		// Select sources based on data type and quality
		for _, source := range p.dataSources {
			if source.IsHealthy() {
				coverage := source.GetCoverage()
				if coverage[request.DataType] > 0.5 { // Minimum 50% coverage
					selectedSources = append(selectedSources, source)
				}
			}
		}
	}

	// Sort by priority and quality
	p.sortSourcesByPriority(selectedSources)

	return selectedSources
}

// collectDataParallel collects data from sources in parallel
func (p *DataCollectionPipeline) collectDataParallel(ctx context.Context, request DataCollectionRequest, sources []DataSource) ([]*RawData, []CollectionError) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var rawData []*RawData
	var errors []CollectionError

	// Limit concurrent collections
	semaphore := make(chan struct{}, p.config.MaxConcurrentCollections)

	for _, source := range sources {
		wg.Add(1)
		go func(src DataSource) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()
			case <-ctx.Done():
				return
			}

			// Check rate limit
			if !p.checkRateLimit(src.GetName()) {
				mu.Lock()
				errors = append(errors, CollectionError{
					Source:    src.GetName(),
					Type:      "rate_limit_exceeded",
					Message:   "Rate limit exceeded",
					Timestamp: time.Now(),
					Retryable: true,
				})
				mu.Unlock()
				return
			}

			// Collect data
			data, err := src.CollectData(ctx, request)
			if err != nil {
				mu.Lock()
				errors = append(errors, CollectionError{
					Source:    src.GetName(),
					Type:      "collection_error",
					Message:   err.Error(),
					Timestamp: time.Now(),
					Retryable: true,
				})
				mu.Unlock()
				return
			}

			// Validate data if enabled
			if p.config.EnableDataValidation {
				validation, err := src.ValidateData(data.RawData[0])
				if err != nil || !validation.IsValid {
					mu.Lock()
					errors = append(errors, CollectionError{
						Source:    src.GetName(),
						Type:      "validation_error",
						Message:   "Data validation failed",
						Timestamp: time.Now(),
						Retryable: false,
					})
					mu.Unlock()
					return
				}
			}

			mu.Lock()
			rawData = append(rawData, data.RawData...)
			mu.Unlock()

		}(source)
	}

	wg.Wait()
	return rawData, errors
}

// collectDataSequential collects data from sources sequentially
func (p *DataCollectionPipeline) collectDataSequential(ctx context.Context, request DataCollectionRequest, sources []DataSource) ([]*RawData, []CollectionError) {
	var rawData []*RawData
	var errors []CollectionError

	for _, source := range sources {
		// Check rate limit
		if !p.checkRateLimit(source.GetName()) {
			errors = append(errors, CollectionError{
				Source:    source.GetName(),
				Type:      "rate_limit_exceeded",
				Message:   "Rate limit exceeded",
				Timestamp: time.Now(),
				Retryable: true,
			})
			continue
		}

		// Collect data
		data, err := source.CollectData(ctx, request)
		if err != nil {
			errors = append(errors, CollectionError{
				Source:    source.GetName(),
				Type:      "collection_error",
				Message:   err.Error(),
				Timestamp: time.Now(),
				Retryable: true,
			})
			continue
		}

		// Validate data if enabled
		if p.config.EnableDataValidation {
			validation, err := source.ValidateData(data.RawData[0])
			if err != nil || !validation.IsValid {
				errors = append(errors, CollectionError{
					Source:    source.GetName(),
					Type:      "validation_error",
					Message:   "Data validation failed",
					Timestamp: time.Now(),
					Retryable: false,
				})
				continue
			}
		}

		rawData = append(rawData, data.RawData...)
	}

	return rawData, errors
}

// processData processes raw data using registered processors
func (p *DataCollectionPipeline) processData(ctx context.Context, rawData []*RawData) ([]*ProcessedData, error) {
	var processedData []*ProcessedData

	for _, raw := range rawData {
		// Find appropriate processor
		processor := p.selectProcessor(raw.DataType)
		if processor == nil {
			p.logger.Warn("No processor found for data type",
				zap.String("data_type", raw.DataType))
			continue
		}

		// Process data
		processed, err := processor.ProcessData(ctx, raw)
		if err != nil {
			p.logger.Error("Data processing failed",
				zap.String("processor", processor.GetName()),
				zap.String("raw_data_id", raw.ID),
				zap.Error(err))
			continue
		}

		// Validate processed data
		if p.config.EnableDataValidation {
			validation, err := processor.ValidateProcessedData(processed)
			if err != nil || !validation.IsValid {
				p.logger.Warn("Processed data validation failed",
					zap.String("processor", processor.GetName()),
					zap.String("processed_data_id", processed.ID))
				continue
			}
		}

		processedData = append(processedData, processed)
	}

	return processedData, nil
}

// aggregateData aggregates processed data using registered aggregators
func (p *DataCollectionPipeline) aggregateData(ctx context.Context, processedData []*ProcessedData) (*AggregatedData, error) {
	if len(processedData) == 0 {
		return nil, fmt.Errorf("no processed data to aggregate")
	}

	// Find appropriate aggregator
	aggregator := p.selectAggregator(processedData[0].DataType)
	if aggregator == nil {
		return nil, fmt.Errorf("no aggregator found for data type: %s", processedData[0].DataType)
	}

	// Aggregate data
	aggregated, err := aggregator.AggregateData(ctx, processedData)
	if err != nil {
		return nil, fmt.Errorf("data aggregation failed: %w", err)
	}

	// Validate aggregated data
	if p.config.EnableDataValidation {
		validation, err := aggregator.ValidateAggregatedData(aggregated)
		if err != nil || !validation.IsValid {
			return nil, fmt.Errorf("aggregated data validation failed: %w", err)
		}
	}

	return aggregated, nil
}

// selectProcessor selects an appropriate processor for the data type
func (p *DataCollectionPipeline) selectProcessor(dataType string) DataProcessor {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Simple selection logic - can be enhanced
	for _, processor := range p.processors {
		if processor.GetType() == dataType {
			return processor
		}
	}

	// Return first available processor as fallback
	for _, processor := range p.processors {
		return processor
	}

	return nil
}

// selectAggregator selects an appropriate aggregator for the data type
func (p *DataCollectionPipeline) selectAggregator(dataType string) DataAggregator {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Simple selection logic - can be enhanced
	for _, aggregator := range p.aggregators {
		if aggregator.GetType() == dataType {
			return aggregator
		}
	}

	// Return first available aggregator as fallback
	for _, aggregator := range p.aggregators {
		return aggregator
	}

	return nil
}

// sortSourcesByPriority sorts data sources by priority and quality
func (p *DataCollectionPipeline) sortSourcesByPriority(sources []DataSource) {
	// Simple sorting by priority - can be enhanced with more sophisticated logic
	for i := 0; i < len(sources)-1; i++ {
		for j := i + 1; j < len(sources); j++ {
			if sources[i].GetConfig().Priority < sources[j].GetConfig().Priority {
				sources[i], sources[j] = sources[j], sources[i]
			}
		}
	}
}

// checkRateLimit checks if a data source is within its rate limit
func (p *DataCollectionPipeline) checkRateLimit(sourceName string) bool {
	if !p.config.EnableRateLimiting {
		return true
	}

	p.mu.RLock()
	rateLimiter, exists := p.rateLimiters[sourceName]
	p.mu.RUnlock()

	if !exists {
		return true
	}

	return rateLimiter.Allow()
}

// calculateQualityScore calculates the overall quality score for the collection result
func (p *DataCollectionPipeline) calculateQualityScore(result *DataCollectionResult) float64 {
	if len(result.RawData) == 0 {
		return 0.0
	}

	var totalScore float64
	for _, raw := range result.RawData {
		totalScore += raw.QualityScore
	}

	return totalScore / float64(len(result.RawData))
}

// calculateCompletenessScore calculates the completeness score for the collection result
func (p *DataCollectionPipeline) calculateCompletenessScore(result *DataCollectionResult) float64 {
	if len(result.RawData) == 0 {
		return 0.0
	}

	// Simple completeness calculation - can be enhanced
	expectedSources := len(result.DataSources)
	actualSources := len(result.RawData)

	if expectedSources == 0 {
		return 1.0
	}

	return float64(actualSources) / float64(expectedSources)
}

// calculateConsistencyScore calculates the consistency score for the collection result
func (p *DataCollectionPipeline) calculateConsistencyScore(result *DataCollectionResult) float64 {
	if len(result.RawData) < 2 {
		return 1.0
	}

	// Simple consistency calculation - can be enhanced
	// For now, return a placeholder value
	return 0.8
}

// updateMetrics updates pipeline metrics
func (p *DataCollectionPipeline) updateMetrics(result *DataCollectionResult) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.metrics.TotalCollections++
	if result.Status == CollectionStatusCompleted {
		p.metrics.SuccessfulCollections++
	} else {
		p.metrics.FailedCollections++
	}

	// Update average processing time
	if p.metrics.TotalCollections == 1 {
		p.metrics.AverageProcessingTime = result.ProcessingTime
	} else {
		// Simple moving average
		p.metrics.AverageProcessingTime = (p.metrics.AverageProcessingTime + result.ProcessingTime) / 2
	}

	// Update data quality scores
	p.metrics.DataQualityScores[result.DataType] = result.QualityScore

	// Update source performance
	for _, raw := range result.RawData {
		if perf, exists := p.metrics.SourcePerformance[raw.SourceName]; exists {
			perf.TotalRequests++
			if result.Status == CollectionStatusCompleted {
				perf.SuccessfulRequests++
			} else {
				perf.FailedRequests++
			}
			perf.QualityScore = raw.QualityScore
			perf.LastRequest = time.Now()
			p.metrics.SourcePerformance[raw.SourceName] = perf
		} else {
			p.metrics.SourcePerformance[raw.SourceName] = SourcePerformance{
				TotalRequests: 1,
				QualityScore:  raw.QualityScore,
				LastRequest:   time.Now(),
			}
		}
	}

	p.metrics.LastUpdated = time.Now()
}

// GetMetrics returns current pipeline metrics
func (p *DataCollectionPipeline) GetMetrics() *PipelineMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *p.metrics
	return &metrics
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(name string, rateLimit, burstLimit int) *RateLimiter {
	return &RateLimiter{
		name:       name,
		rateLimit:  rateLimit,
		burstLimit: burstLimit,
		tokens:     burstLimit,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed.Minutes()) * rl.rateLimit / 60
	if tokensToAdd > 0 {
		rl.tokens = min(rl.tokens+tokensToAdd, rl.burstLimit)
		rl.lastRefill = now
	}

	// Check if we have tokens available
	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("dc_%d", time.Now().UnixNano())
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
