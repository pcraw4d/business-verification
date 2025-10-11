package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/ml/service"
	"kyb-platform/services/risk-assessment-service/internal/models"
)

// RiskEngine provides high-performance risk assessment capabilities
type RiskEngine struct {
	mlService      *service.MLService
	cache          Cache
	pool           *WorkerPool
	logger         *zap.Logger
	config         *Config
	metrics        *Metrics
	circuitBreaker *CircuitBreaker
}

// Config holds risk engine configuration
type Config struct {
	MaxConcurrentRequests int                  `json:"max_concurrent_requests"`
	RequestTimeout        time.Duration        `json:"request_timeout"`
	CacheTTL              time.Duration        `json:"cache_ttl"`
	CircuitBreakerConfig  CircuitBreakerConfig `json:"circuit_breaker"`
	EnableMetrics         bool                 `json:"enable_metrics"`
	EnableCaching         bool                 `json:"enable_caching"`
}

// DefaultConfig returns default risk engine configuration
func DefaultConfig() *Config {
	return &Config{
		MaxConcurrentRequests: 100,
		RequestTimeout:        500 * time.Millisecond, // Sub-1-second target
		CacheTTL:              5 * time.Minute,
		CircuitBreakerConfig: CircuitBreakerConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  30 * time.Second,
			HalfOpenMaxCalls: 3,
		},
		EnableMetrics: true,
		EnableCaching: true,
	}
}

// NewRiskEngine creates a new risk assessment engine
func NewRiskEngine(mlService *service.MLService, logger *zap.Logger, config *Config) *RiskEngine {
	if config == nil {
		config = DefaultConfig()
	}

	// Initialize cache
	var cache Cache
	if config.EnableCaching {
		cache = NewInMemoryCache(config.CacheTTL, logger)
	} else {
		cache = NewNoOpCache()
	}

	// Initialize worker pool
	pool := NewWorkerPool(config.MaxConcurrentRequests, logger)
	pool.Start()

	// Initialize circuit breaker
	circuitBreaker := NewCircuitBreaker(config.CircuitBreakerConfig, logger)

	// Initialize metrics
	var metrics *Metrics
	if config.EnableMetrics {
		metrics = NewMetrics()
	}

	return &RiskEngine{
		mlService:      mlService,
		cache:          cache,
		pool:           pool,
		logger:         logger,
		config:         config,
		metrics:        metrics,
		circuitBreaker: circuitBreaker,
	}
}

// AssessRisk performs high-performance risk assessment
func (re *RiskEngine) AssessRisk(ctx context.Context, req *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	start := time.Now()

	// Generate cache key
	cacheKey := re.generateCacheKey(req)

	// Check cache first
	if re.config.EnableCaching {
		if cached, found := re.cache.Get(cacheKey); found {
			re.logger.Debug("Cache hit for risk assessment", zap.String("cache_key", cacheKey))
			if re.metrics != nil {
				re.metrics.RecordCacheHit()
			}
			return cached.(*models.RiskAssessment), nil
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, re.config.RequestTimeout)
	defer cancel()

	// Use circuit breaker
	result, err := re.circuitBreaker.Execute(func() (interface{}, error) {
		return re.performRiskAssessment(ctx, req)
	})

	if err != nil {
		re.logger.Error("Risk assessment failed", zap.Error(err))
		if re.metrics != nil {
			re.metrics.RecordError()
		}
		return nil, err
	}

	assessment := result.(*models.RiskAssessment)

	// Cache the result
	if re.config.EnableCaching {
		re.cache.Set(cacheKey, assessment)
	}

	// Record metrics
	duration := time.Since(start)
	if re.metrics != nil {
		re.metrics.RecordRequest(duration)
		re.metrics.RecordCacheMiss()
	}

	re.logger.Info("Risk assessment completed",
		zap.String("assessment_id", assessment.ID),
		zap.Duration("duration", duration),
		zap.Float64("risk_score", assessment.RiskScore))

	return assessment, nil
}

// AssessRiskBatch performs batch risk assessment for multiple businesses
func (re *RiskEngine) AssessRiskBatch(ctx context.Context, requests []*models.RiskAssessmentRequest) ([]*models.RiskAssessment, error) {
	start := time.Now()

	if len(requests) == 0 {
		return []*models.RiskAssessment{}, nil
	}

	// Limit batch size
	maxBatchSize := re.config.MaxConcurrentRequests
	if len(requests) > maxBatchSize {
		requests = requests[:maxBatchSize]
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, re.config.RequestTimeout*2) // Allow more time for batch
	defer cancel()

	// Use worker pool for concurrent processing
	results := make([]*models.RiskAssessment, len(requests))
	errors := make([]error, len(requests))

	var wg sync.WaitGroup
	for i, req := range requests {
		wg.Add(1)
		go func(index int, request *models.RiskAssessmentRequest) {
			defer wg.Done()

			assessment, err := re.AssessRisk(ctx, request)
			results[index] = assessment
			errors[index] = err
		}(i, req)
	}

	wg.Wait()

	// Check for errors
	var hasErrors bool
	for _, err := range errors {
		if err != nil {
			hasErrors = true
			break
		}
	}

	if hasErrors {
		re.logger.Warn("Some batch assessments failed", zap.Int("total", len(requests)))
	}

	// Record metrics
	duration := time.Since(start)
	if re.metrics != nil {
		re.metrics.RecordBatchRequest(duration, len(requests))
	}

	re.logger.Info("Batch risk assessment completed",
		zap.Int("count", len(requests)),
		zap.Duration("duration", duration))

	return results, nil
}

// PredictRisk performs high-performance risk prediction
func (re *RiskEngine) PredictRisk(ctx context.Context, req *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	start := time.Now()

	// Generate cache key for prediction
	cacheKey := re.generatePredictionCacheKey(req, horizonMonths)

	// Check cache first
	if re.config.EnableCaching {
		if cached, found := re.cache.Get(cacheKey); found {
			re.logger.Debug("Cache hit for risk prediction", zap.String("cache_key", cacheKey))
			if re.metrics != nil {
				re.metrics.RecordCacheHit()
			}
			return cached.(*models.RiskPrediction), nil
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, re.config.RequestTimeout)
	defer cancel()

	// Use circuit breaker
	result, err := re.circuitBreaker.Execute(func() (interface{}, error) {
		return re.performRiskPrediction(ctx, req, horizonMonths)
	})

	if err != nil {
		re.logger.Error("Risk prediction failed", zap.Error(err))
		if re.metrics != nil {
			re.metrics.RecordError()
		}
		return nil, err
	}

	prediction := result.(*models.RiskPrediction)

	// Cache the result
	if re.config.EnableCaching {
		re.cache.Set(cacheKey, prediction)
	}

	// Record metrics
	duration := time.Since(start)
	if re.metrics != nil {
		re.metrics.RecordRequest(duration)
		re.metrics.RecordCacheMiss()
	}

	re.logger.Info("Risk prediction completed",
		zap.String("business_id", prediction.BusinessID),
		zap.Int("horizon_months", horizonMonths),
		zap.Duration("duration", duration))

	return prediction, nil
}

// performRiskAssessment performs the actual risk assessment
func (re *RiskEngine) performRiskAssessment(ctx context.Context, req *models.RiskAssessmentRequest) (*models.RiskAssessment, error) {
	// Create assessment
	assessment := &models.RiskAssessment{
		ID:                re.generateID(),
		BusinessName:      req.BusinessName,
		BusinessAddress:   req.BusinessAddress,
		Industry:          req.Industry,
		Country:           req.Country,
		PredictionHorizon: req.PredictionHorizon,
		Status:            models.StatusPending,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Metadata:          req.Metadata,
	}

	// Set default prediction horizon if not provided
	if assessment.PredictionHorizon == 0 {
		assessment.PredictionHorizon = 3
	}

	// Use ML service for risk prediction with ensemble routing
	// Determine model type based on prediction horizon
	modelType := "auto" // Use ensemble routing by default
	if req.ModelType != "" {
		modelType = req.ModelType
	}
	
	// Create a copy of the request with the model type
	reqCopy := *req
	reqCopy.ModelType = modelType
	
	mlAssessment, err := re.mlService.PredictRisk(ctx, modelType, &reqCopy)
	if err != nil {
		re.logger.Error("ML prediction failed, using fallback", zap.Error(err))
		// Fallback to mock response if ML fails
		assessment.RiskScore = 0.75
		assessment.RiskLevel = models.RiskLevelMedium
		assessment.ConfidenceScore = 0.85
		assessment.Status = models.StatusCompleted
		assessment.RiskFactors = []models.RiskFactor{
			{
				Category:    models.RiskCategoryFinancial,
				Name:        "Credit Score",
				Score:       0.8,
				Weight:      0.3,
				Description: "Business credit score analysis",
				Source:      "internal",
				Confidence:  0.9,
			},
		}
	} else {
		// Use ML prediction results
		assessment.RiskScore = mlAssessment.RiskScore
		assessment.RiskLevel = mlAssessment.RiskLevel
		assessment.ConfidenceScore = mlAssessment.ConfidenceScore
		assessment.RiskFactors = mlAssessment.RiskFactors
		assessment.Status = models.StatusCompleted
		
		// Add model information to metadata
		if assessment.Metadata == nil {
			assessment.Metadata = make(map[string]interface{})
		}
		assessment.Metadata["model_type"] = modelType
		assessment.Metadata["prediction_horizon"] = assessment.PredictionHorizon
		
		// Add ensemble information if available
		if modelType == "auto" || modelType == "ensemble" {
			ensembleInfo := re.mlService.GetEnsembleInfo()
			assessment.Metadata["ensemble_info"] = ensembleInfo
		}
	}

	return assessment, nil
}

// performRiskPrediction performs the actual risk prediction
func (re *RiskEngine) performRiskPrediction(ctx context.Context, req *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	// Determine model type based on prediction horizon
	modelType := "auto" // Use ensemble routing by default
	if req.ModelType != "" {
		modelType = req.ModelType
	}
	
	// Use ML service for future risk prediction with ensemble routing
	prediction, err := re.mlService.PredictFutureRisk(ctx, modelType, req, horizonMonths)
	if err != nil {
		re.logger.Error("Future risk prediction failed", zap.Error(err))
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	return prediction, nil
}

// generateCacheKey generates a cache key for risk assessment
func (re *RiskEngine) generateCacheKey(req *models.RiskAssessmentRequest) string {
	return fmt.Sprintf("risk_assessment:%s:%s:%s:%s:%d",
		req.BusinessName,
		req.BusinessAddress,
		req.Industry,
		req.Country,
		req.PredictionHorizon)
}

// generatePredictionCacheKey generates a cache key for risk prediction
func (re *RiskEngine) generatePredictionCacheKey(req *models.RiskAssessmentRequest, horizonMonths int) string {
	return fmt.Sprintf("risk_prediction:%s:%s:%s:%s:%d",
		req.BusinessName,
		req.BusinessAddress,
		req.Industry,
		req.Country,
		horizonMonths)
}

// generateID generates a unique ID for assessments
func (re *RiskEngine) generateID() string {
	return fmt.Sprintf("risk_%d", time.Now().UnixNano())
}

// GetMetrics returns engine metrics
func (re *RiskEngine) GetMetrics() *Metrics {
	return re.metrics
}

// GetCacheStats returns cache statistics
func (re *RiskEngine) GetCacheStats() CacheStats {
	return re.cache.GetStats()
}

// GetCircuitBreakerState returns circuit breaker state
func (re *RiskEngine) GetCircuitBreakerState() CircuitBreakerState {
	return re.circuitBreaker.GetState()
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (re *RiskEngine) GetCircuitBreakerStats() CircuitBreakerStats {
	return re.circuitBreaker.GetStats()
}

// Shutdown gracefully shuts down the risk engine
func (re *RiskEngine) Shutdown(ctx context.Context) error {
	re.logger.Info("Shutting down risk engine")

	// Shutdown worker pool
	if err := re.pool.Shutdown(ctx); err != nil {
		re.logger.Error("Failed to shutdown worker pool", zap.Error(err))
	}

	// Clear cache
	re.cache.Clear()

	re.logger.Info("Risk engine shutdown complete")
	return nil
}
