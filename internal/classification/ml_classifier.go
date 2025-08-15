package classification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// MLClassificationRequest represents a request for ML-based classification
type MLClassificationRequest struct {
	BusinessName        string                 `json:"business_name"`
	BusinessDescription string                 `json:"business_description"`
	Keywords            []string               `json:"keywords"`
	WebsiteContent      string                 `json:"website_content"`
	IndustryHints       []string               `json:"industry_hints"`
	GeographicRegion    string                 `json:"geographic_region"`
	BusinessType        string                 `json:"business_type"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// MLClassificationResult represents the result of ML-based classification
type MLClassificationResult struct {
	IndustryCode       string                 `json:"industry_code"`
	IndustryName       string                 `json:"industry_name"`
	ConfidenceScore    float64                `json:"confidence_score"`
	ModelType          ModelType              `json:"model_type"`
	ModelVersion       string                 `json:"model_version"`
	InferenceTime      time.Duration          `json:"inference_time"`
	ModelPredictions   []ModelPrediction      `json:"model_predictions"`
	EnsembleScore      float64                `json:"ensemble_score"`
	FeatureImportance  map[string]float64     `json:"feature_importance"`
	ProcessingMetadata map[string]interface{} `json:"processing_metadata"`
}

// ModelPrediction represents a prediction from a single model
type ModelPrediction struct {
	ModelID         string    `json:"model_id"`
	ModelType       ModelType `json:"model_type"`
	IndustryCode    string    `json:"industry_code"`
	IndustryName    string    `json:"industry_name"`
	ConfidenceScore float64   `json:"confidence_score"`
	RawScore        float64   `json:"raw_score"`
}

// EnsembleConfig represents configuration for ensemble methods
type EnsembleConfig struct {
	Enabled          bool               `json:"enabled"`
	Method           string             `json:"method"` // "weighted_average", "voting", "stacking"
	ModelWeights     map[string]float64 `json:"model_weights"`
	MinConfidence    float64            `json:"min_confidence"`
	MaxModels        int                `json:"max_models"`
	FallbackStrategy string             `json:"fallback_strategy"`
}

// FeatureExtractor represents feature extraction configuration
type FeatureExtractor struct {
	TextFeatures        bool `json:"text_features"`
	SemanticFeatures    bool `json:"semantic_features"`
	StatisticalFeatures bool `json:"statistical_features"`
	DomainFeatures      bool `json:"domain_features"`
}

// BatchProcessor represents batch processing configuration
type BatchProcessor struct {
	Enabled        bool          `json:"enabled"`
	BatchSize      int           `json:"batch_size"`
	MaxConcurrency int           `json:"max_concurrency"`
	Timeout        time.Duration `json:"timeout"`
	RetryAttempts  int           `json:"retry_attempts"`
}

// MLClassifier provides ML-based classification capabilities
type MLClassifier struct {
	logger  *observability.Logger
	metrics *observability.Metrics

	// Model management
	modelManager *ModelManager

	// Model optimization
	modelOptimizer *ModelOptimizer

	// Configuration
	ensembleConfig   *EnsembleConfig
	featureExtractor *FeatureExtractor
	batchSize        int
	timeout          time.Duration

	// Caching
	resultCache map[string]*MLClassificationResult
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration

	// Performance tracking
	inferenceTimes  map[string]time.Duration
	accuracyMetrics map[string]float64
	metricsMutex    sync.RWMutex

	// Batch processing
	batchProcessor *BatchProcessor
	batchMutex     sync.RWMutex

	// Fallback mechanisms
	fallbackEnabled bool
	fallbackModels  []string
}

// NewMLClassifier creates a new ML classifier
func NewMLClassifier(logger *observability.Logger, metrics *observability.Metrics, modelManager *ModelManager, modelOptimizer *ModelOptimizer) *MLClassifier {
	classifier := &MLClassifier{
		logger:  logger,
		metrics: metrics,

		// Model management
		modelManager: modelManager,

		// Model optimization
		modelOptimizer: modelOptimizer,

		// Configuration
		ensembleConfig: &EnsembleConfig{
			Enabled:          true,
			Method:           "weighted_average",
			ModelWeights:     make(map[string]float64),
			MinConfidence:    0.5,
			MaxModels:        5,
			FallbackStrategy: "best_single_model",
		},
		featureExtractor: &FeatureExtractor{
			TextFeatures:        true,
			SemanticFeatures:    true,
			StatisticalFeatures: true,
			DomainFeatures:      true,
		},
		batchSize: 32,
		timeout:   30 * time.Second,

		// Caching
		resultCache: make(map[string]*MLClassificationResult),
		cacheTTL:    1 * time.Hour,

		// Performance tracking
		inferenceTimes:  make(map[string]time.Duration),
		accuracyMetrics: make(map[string]float64),
	}

	// Initialize default model weights
	classifier.initializeDefaultWeights()

	// Initialize batch processor
	classifier.batchProcessor = &BatchProcessor{
		Enabled:        true,
		BatchSize:      32,
		MaxConcurrency: 4,
		Timeout:        30 * time.Second,
		RetryAttempts:  3,
	}

	// Initialize fallback mechanisms
	classifier.fallbackEnabled = true
	classifier.fallbackModels = []string{"bert_base", "ensemble_fallback"}

	return classifier
}

// Classify performs ML-based classification
func (c *MLClassifier) Classify(ctx context.Context, request *MLClassificationRequest) (*MLClassificationResult, error) {
	start := time.Now()

	// Log classification start
	if c.logger != nil {
		c.logger.WithComponent("ml_classifier").LogBusinessEvent(ctx, "ml_classification_started", "", map[string]interface{}{
			"business_name": request.BusinessName,
			"model_types":   []string{string(ModelTypeBERT), string(ModelTypeEnsemble)},
		})
	}

	// Check cache first
	cacheKey := c.generateCacheKey(request)
	if cached, exists := c.getFromCache(cacheKey); exists {
		if c.logger != nil {
			c.logger.WithComponent("ml_classifier").LogBusinessEvent(ctx, "ml_classification_cache_hit", "", map[string]interface{}{
				"business_name": request.BusinessName,
			})
		}
		return cached, nil
	}

	// Extract features
	features, err := c.extractFeatures(request)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Perform ensemble classification
	result, err := c.performEnsembleClassification(ctx, features, request)
	if err != nil {
		return nil, fmt.Errorf("failed to perform ensemble classification: %w", err)
	}

	// Calculate ensemble score
	result.EnsembleScore = c.calculateEnsembleScore(result.ModelPredictions)

	// Add processing metadata
	result.ProcessingMetadata = map[string]interface{}{
		"feature_count":   len(features),
		"models_used":     len(result.ModelPredictions),
		"ensemble_method": c.ensembleConfig.Method,
		"processing_time": time.Since(start).Milliseconds(),
		"cache_key":       cacheKey,
	}

	// Cache the result
	c.cacheResult(cacheKey, result)

	// Log classification completion
	if c.logger != nil {
		c.logger.WithComponent("ml_classifier").LogBusinessEvent(ctx, "ml_classification_completed", "", map[string]interface{}{
			"business_name":     request.BusinessName,
			"industry_code":     result.IndustryCode,
			"confidence_score":  result.ConfidenceScore,
			"ensemble_score":    result.EnsembleScore,
			"inference_time_ms": result.InferenceTime.Milliseconds(),
		})
	}

	// Record metrics
	c.RecordClassificationMetrics(ctx, result, request)

	return result, nil
}

// ClassifyBatch performs batch ML-based classification
func (c *MLClassifier) ClassifyBatch(ctx context.Context, requests []*MLClassificationRequest) ([]*MLClassificationResult, error) {
	start := time.Now()

	if c.logger != nil {
		c.logger.WithComponent("ml_classifier").LogBusinessEvent(ctx, "ml_batch_classification_started", "", map[string]interface{}{
			"batch_size": len(requests),
		})
	}

	results := make([]*MLClassificationResult, 0, len(requests))
	errors := make([]error, 0)

	// Process in batches
	for i := 0; i < len(requests); i += c.batchSize {
		end := i + c.batchSize
		if end > len(requests) {
			end = len(requests)
		}

		batch := requests[i:end]
		batchResults, batchErrors := c.processBatch(ctx, batch)

		results = append(results, batchResults...)
		errors = append(errors, batchErrors...)
	}

	// Log batch completion
	if c.logger != nil {
		c.logger.WithComponent("ml_classifier").LogBusinessEvent(ctx, "ml_batch_classification_completed", "", map[string]interface{}{
			"batch_size":         len(requests),
			"successful":         len(results),
			"errors":             len(errors),
			"processing_time_ms": time.Since(start).Milliseconds(),
		})
	}

	// Return results (errors are logged but don't fail the entire batch)
	return results, nil
}

// performEnsembleClassification performs classification using multiple models
func (c *MLClassifier) performEnsembleClassification(ctx context.Context, features map[string]interface{}, request *MLClassificationRequest) (*MLClassificationResult, error) {
	start := time.Now()

	// Get available models
	models, err := c.getAvailableModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available models: %w", err)
	}

	// Perform predictions with each model
	predictions := make([]ModelPrediction, 0)
	for _, model := range models {
		prediction, err := c.predictWithModel(ctx, model, features, request)
		if err != nil {
			if c.logger != nil {
				c.logger.WithComponent("ml_classifier").Warn("model_prediction_failed", map[string]interface{}{
					"model_id": model.ID,
					"error":    err.Error(),
				})
			}
			continue
		}
		predictions = append(predictions, *prediction)
	}

	if len(predictions) == 0 {
		return nil, fmt.Errorf("no successful predictions from any model")
	}

	// Select best prediction based on ensemble method
	var bestPrediction ModelPrediction
	switch c.ensembleConfig.Method {
	case "weighted_average":
		bestPrediction = c.selectWeightedAverage(predictions)
	case "voting":
		bestPrediction = c.selectVoting(predictions)
	case "stacking":
		bestPrediction = c.selectStacking(predictions)
	default:
		bestPrediction = c.selectBestSingle(predictions)
	}

	// Create result
	result := &MLClassificationResult{
		IndustryCode:      bestPrediction.IndustryCode,
		IndustryName:      bestPrediction.IndustryName,
		ConfidenceScore:   bestPrediction.ConfidenceScore,
		ModelType:         bestPrediction.ModelType,
		ModelVersion:      "ensemble",
		InferenceTime:     time.Since(start),
		ModelPredictions:  predictions,
		FeatureImportance: c.calculateFeatureImportance(features),
	}

	return result, nil
}

// predictWithModel performs prediction with a single model
func (c *MLClassifier) predictWithModel(ctx context.Context, model *ModelInfo, features map[string]interface{}, request *MLClassificationRequest) (*ModelPrediction, error) {
	start := time.Now()

	// This would integrate with actual ML frameworks
	// For now, we'll simulate prediction based on model type
	var prediction *ModelPrediction

	switch model.Type {
	case ModelTypeBERT:
		prediction = c.simulateBERTPrediction(model, features, request)
	case ModelTypeEnsemble:
		prediction = c.simulateEnsemblePrediction(model, features, request)
	case ModelTypeTransformer:
		prediction = c.simulateTransformerPrediction(model, features, request)
	default:
		prediction = c.simulateCustomPrediction(model, features, request)
	}

	// Update inference time
	inferenceTime := time.Since(start)
	c.metricsMutex.Lock()
	c.inferenceTimes[model.ID] = inferenceTime
	c.metricsMutex.Unlock()

	prediction.InferenceTime = inferenceTime

	return prediction, nil
}

// extractFeatures extracts features from the classification request
func (c *MLClassifier) extractFeatures(request *MLClassificationRequest) (map[string]interface{}, error) {
	features := make(map[string]interface{})

	// Text features
	if c.featureExtractor.TextFeatures {
		features["business_name_length"] = len(request.BusinessName)
		features["description_length"] = len(request.BusinessDescription)
		features["keyword_count"] = len(request.Keywords)
		features["website_content_length"] = len(request.WebsiteContent)
		features["business_name_words"] = len(c.tokenize(request.BusinessName))
		features["description_words"] = len(c.tokenize(request.BusinessDescription))
	}

	// Semantic features
	if c.featureExtractor.SemanticFeatures {
		features["business_name_embedding"] = c.generateEmbedding(request.BusinessName)
		features["description_embedding"] = c.generateEmbedding(request.BusinessDescription)
		features["keyword_embeddings"] = c.generateKeywordEmbeddings(request.Keywords)
	}

	// Statistical features
	if c.featureExtractor.StatisticalFeatures {
		features["name_word_frequency"] = c.calculateWordFrequency(request.BusinessName)
		features["description_word_frequency"] = c.calculateWordFrequency(request.BusinessDescription)
		features["keyword_overlap"] = c.calculateKeywordOverlap(request.Keywords, request.IndustryHints)
	}

	// Domain features
	if c.featureExtractor.DomainFeatures {
		features["geographic_region"] = request.GeographicRegion
		features["business_type"] = request.BusinessType
		features["industry_hints"] = request.IndustryHints
	}

	return features, nil
}

// getAvailableModels gets available models for classification
func (c *MLClassifier) getAvailableModels(ctx context.Context) ([]*ModelInfo, error) {
	models := make([]*ModelInfo, 0)

	// Try to get BERT model
	if bertModel, err := c.modelManager.GetModelByType(ctx, ModelTypeBERT); err == nil {
		models = append(models, bertModel)
	}

	// Try to get ensemble model
	if ensembleModel, err := c.modelManager.GetModelByType(ctx, ModelTypeEnsemble); err == nil {
		models = append(models, ensembleModel)
	}

	// Try to get transformer model
	if transformerModel, err := c.modelManager.GetModelByType(ctx, ModelTypeTransformer); err == nil {
		models = append(models, transformerModel)
	}

	if len(models) == 0 {
		return nil, fmt.Errorf("no available models for classification")
	}

	return models, nil
}

// selectWeightedAverage selects the best prediction using weighted average
func (c *MLClassifier) selectWeightedAverage(predictions []ModelPrediction) ModelPrediction {
	var bestPrediction ModelPrediction
	var bestScore float64

	for _, pred := range predictions {
		weight := c.ensembleConfig.ModelWeights[pred.ModelID]
		if weight == 0 {
			weight = 1.0 // Default weight
		}

		weightedScore := pred.ConfidenceScore * weight
		if weightedScore > bestScore {
			bestScore = weightedScore
			bestPrediction = pred
		}
	}

	return bestPrediction
}

// selectVoting selects the best prediction using voting
func (c *MLClassifier) selectVoting(predictions []ModelPrediction) ModelPrediction {
	votes := make(map[string]int)
	scores := make(map[string]float64)

	for _, pred := range predictions {
		key := pred.IndustryCode
		votes[key]++
		scores[key] += pred.ConfidenceScore
	}

	var bestIndustryCode string
	var maxVotes int
	var maxScore float64

	for industryCode, voteCount := range votes {
		avgScore := scores[industryCode] / float64(voteCount)
		if voteCount > maxVotes || (voteCount == maxVotes && avgScore > maxScore) {
			maxVotes = voteCount
			maxScore = avgScore
			bestIndustryCode = industryCode
		}
	}

	// Find the prediction with the highest confidence for the winning industry
	var bestPrediction ModelPrediction
	var bestConfidence float64

	for _, pred := range predictions {
		if pred.IndustryCode == bestIndustryCode && pred.ConfidenceScore > bestConfidence {
			bestConfidence = pred.ConfidenceScore
			bestPrediction = pred
		}
	}

	return bestPrediction
}

// selectStacking selects the best prediction using stacking
func (c *MLClassifier) selectStacking(predictions []ModelPrediction) ModelPrediction {
	// Stacking combines predictions using a meta-model
	// For now, we'll use a simple weighted combination
	return c.selectWeightedAverage(predictions)
}

// selectBestSingle selects the best single model prediction
func (c *MLClassifier) selectBestSingle(predictions []ModelPrediction) ModelPrediction {
	var bestPrediction ModelPrediction
	var bestConfidence float64

	for _, pred := range predictions {
		if pred.ConfidenceScore > bestConfidence {
			bestConfidence = pred.ConfidenceScore
			bestPrediction = pred
		}
	}

	return bestPrediction
}

// calculateEnsembleScore calculates the ensemble confidence score
func (c *MLClassifier) calculateEnsembleScore(predictions []ModelPrediction) float64 {
	if len(predictions) == 0 {
		return 0.0
	}

	var totalScore float64
	var totalWeight float64

	for _, pred := range predictions {
		weight := 1.0 // Default weight
		totalScore += pred.ConfidenceScore * weight
		totalWeight += weight
	}

	if totalWeight == 0 {
		return 0.0
	}

	return totalScore / totalWeight
}

// calculateFeatureImportance calculates feature importance scores
func (c *MLClassifier) calculateFeatureImportance(features map[string]interface{}) map[string]float64 {
	importance := make(map[string]float64)

	// This would be calculated by the actual ML models
	// For now, we'll use simple heuristics
	for feature, value := range features {
		switch v := value.(type) {
		case int:
			importance[feature] = float64(v) / 1000.0 // Normalize
		case float64:
			importance[feature] = v
		case string:
			importance[feature] = float64(len(v)) / 100.0 // Normalize by length
		default:
			importance[feature] = 0.5 // Default importance
		}
	}

	return importance
}

// processBatch processes a batch of classification requests
func (c *MLClassifier) processBatch(ctx context.Context, requests []*MLClassificationRequest) ([]*MLClassificationResult, []error) {
	results := make([]*MLClassificationResult, 0, len(requests))
	errors := make([]error, 0)

	// Process requests concurrently
	resultChan := make(chan *MLClassificationResult, len(requests))
	errorChan := make(chan error, len(requests))

	for _, request := range requests {
		go func(req *MLClassificationRequest) {
			result, err := c.Classify(ctx, req)
			if err != nil {
				errorChan <- err
			} else {
				resultChan <- result
			}
		}(request)
	}

	// Collect results
	for i := 0; i < len(requests); i++ {
		select {
		case result := <-resultChan:
			results = append(results, result)
		case err := <-errorChan:
			errors = append(errors, err)
		case <-time.After(c.timeout):
			errors = append(errors, fmt.Errorf("classification timeout"))
		}
	}

	return results, errors
}

// generateCacheKey generates a cache key for the request
func (c *MLClassifier) generateCacheKey(request *MLClassificationRequest) string {
	// Simple hash-based cache key
	key := fmt.Sprintf("%s:%s:%s:%s",
		request.BusinessName,
		request.BusinessDescription,
		request.BusinessType,
		request.GeographicRegion)

	return fmt.Sprintf("ml_classify:%x", c.hashString(key))
}

// getFromCache retrieves a result from cache
func (c *MLClassifier) getFromCache(key string) (*MLClassificationResult, bool) {
	c.cacheMutex.RLock()
	defer c.cacheMutex.RUnlock()

	result, exists := c.resultCache[key]
	if !exists {
		return nil, false
	}

	// Check if cache entry is still valid
	if time.Since(result.ProcessingMetadata["cached_at"].(time.Time)) > c.cacheTTL {
		delete(c.resultCache, key)
		return nil, false
	}

	return result, true
}

// cacheResult stores a result in cache
func (c *MLClassifier) cacheResult(key string, result *MLClassificationResult) {
	c.cacheMutex.Lock()
	defer c.cacheMutex.Unlock()

	result.ProcessingMetadata["cached_at"] = time.Now()
	c.resultCache[key] = result
}

// RecordClassificationMetrics records metrics for ML classification
func (c *MLClassifier) RecordClassificationMetrics(ctx context.Context, result *MLClassificationResult, request *MLClassificationRequest) {
	if c.metrics == nil {
		return
	}

	// Record classification metrics using available method
	c.metrics.RecordBusinessClassification("ml_classification_confidence", fmt.Sprintf("%.3f", result.ConfidenceScore))
	c.metrics.RecordBusinessClassification("ml_classification_inference_time", fmt.Sprintf("%d", result.InferenceTime.Milliseconds()))
	c.metrics.RecordBusinessClassification("ml_classification_ensemble_score", fmt.Sprintf("%.3f", result.EnsembleScore))

	// Record model-specific metrics
	for _, pred := range result.ModelPredictions {
		c.metrics.RecordBusinessClassification("ml_model_prediction_confidence", fmt.Sprintf("%.3f", pred.ConfidenceScore))
	}
}

// initializeDefaultWeights initializes default model weights
func (c *MLClassifier) initializeDefaultWeights() {
	c.ensembleConfig.ModelWeights["default_bert"] = 0.4
	c.ensembleConfig.ModelWeights["default_ensemble"] = 0.3
	c.ensembleConfig.ModelWeights["default_transformer"] = 0.2
	c.ensembleConfig.ModelWeights["default_custom"] = 0.1
}

// Helper methods for feature extraction (simplified implementations)
func (c *MLClassifier) tokenize(text string) []string {
	// Simple tokenization - split by whitespace
	// This would be replaced with proper NLP tokenization
	return []string{} // Placeholder
}

func (c *MLClassifier) generateEmbedding(text string) []float64 {
	// Generate text embedding
	// This would integrate with actual embedding models
	return []float64{} // Placeholder
}

func (c *MLClassifier) generateKeywordEmbeddings(keywords []string) [][]float64 {
	// Generate keyword embeddings
	embeddings := make([][]float64, len(keywords))
	for i := range embeddings {
		embeddings[i] = c.generateEmbedding(keywords[i])
	}
	return embeddings
}

func (c *MLClassifier) calculateWordFrequency(text string) map[string]float64 {
	// Calculate word frequency
	// This would implement proper word frequency analysis
	return make(map[string]float64) // Placeholder
}

func (c *MLClassifier) calculateKeywordOverlap(keywords, hints []string) float64 {
	// Calculate overlap between keywords and industry hints
	if len(keywords) == 0 || len(hints) == 0 {
		return 0.0
	}

	keywordSet := make(map[string]bool)
	for _, keyword := range keywords {
		keywordSet[keyword] = true
	}

	overlap := 0
	for _, hint := range hints {
		if keywordSet[hint] {
			overlap++
		}
	}

	return float64(overlap) / float64(len(hints))
}

func (c *MLClassifier) hashString(s string) uint32 {
	// Simple hash function
	hash := uint32(0)
	for _, char := range s {
		hash = hash*31 + uint32(char)
	}
	return hash
}

// Simulation methods for different model types
func (c *MLClassifier) simulateBERTPrediction(model *ModelInfo, features map[string]interface{}, request *MLClassificationRequest) *ModelPrediction {
	// Simulate BERT-based prediction
	// This would integrate with actual BERT models
	return &ModelPrediction{
		ModelID:         model.ID,
		ModelType:       ModelTypeBERT,
		IndustryCode:    "51", // Information
		IndustryName:    "Information",
		ConfidenceScore: 0.85,
		RawScore:        0.85,
	}
}

func (c *MLClassifier) simulateEnsemblePrediction(model *ModelInfo, features map[string]interface{}, request *MLClassificationRequest) *ModelPrediction {
	// Simulate ensemble prediction
	return &ModelPrediction{
		ModelID:         model.ID,
		ModelType:       ModelTypeEnsemble,
		IndustryCode:    "52", // Finance and Insurance
		IndustryName:    "Finance and Insurance",
		ConfidenceScore: 0.82,
		RawScore:        0.82,
	}
}

func (c *MLClassifier) simulateTransformerPrediction(model *ModelInfo, features map[string]interface{}, request *MLClassificationRequest) *ModelPrediction {
	// Simulate transformer prediction
	return &ModelPrediction{
		ModelID:         model.ID,
		ModelType:       ModelTypeTransformer,
		IndustryCode:    "54", // Professional Services
		IndustryName:    "Professional Services",
		ConfidenceScore: 0.78,
		RawScore:        0.78,
	}
}

func (c *MLClassifier) simulateCustomPrediction(model *ModelInfo, features map[string]interface{}, request *MLClassificationRequest) *ModelPrediction {
	// Simulate custom model prediction
	return &ModelPrediction{
		ModelID:         model.ID,
		ModelType:       ModelTypeCustom,
		IndustryCode:    "44", // Retail Trade
		IndustryName:    "Retail Trade",
		ConfidenceScore: 0.75,
		RawScore:        0.75,
	}
}

// Enhanced ML Classifier methods

// ClassifyWithOptimization performs ML-based classification with optimization
func (c *MLClassifier) ClassifyWithOptimization(ctx context.Context, request *MLClassificationRequest) (*MLClassificationResult, error) {
	start := time.Now()

	// Log optimization classification start
	if c.logger != nil {
		c.logger.WithComponent("ml_classifier").LogBusinessEvent(ctx, "ml_optimized_classification_started", "", map[string]interface{}{
			"business_name":        request.BusinessName,
			"optimization_enabled": c.modelOptimizer != nil,
		})
	}

	// Try to get optimized model first
	if c.modelOptimizer != nil {
		optimizedModel, err := c.modelOptimizer.GetOptimizedModel(ctx, "bert_classifier", "v1.0.0")
		if err == nil {
			// Use optimized model for inference
			result, err := c.performOptimizedInference(ctx, request, optimizedModel)
			if err == nil {
				// Update performance metrics
				c.modelOptimizer.UpdatePerformanceMetrics(ctx, "bert_classifier", "v1.0.0", result.InferenceTime, result.ConfidenceScore, 0)
				return result, nil
			}
		}
	}

	// Fallback to regular classification
	return c.Classify(ctx, request)
}

// ClassifyBatchOptimized performs optimized batch ML-based classification
func (c *MLClassifier) ClassifyBatchOptimized(ctx context.Context, requests []*MLClassificationRequest) ([]*MLClassificationResult, error) {
	if !c.batchProcessor.Enabled {
		return c.ClassifyBatch(ctx, requests)
	}

	start := time.Now()

	// Log batch optimization start
	if c.logger != nil {
		c.logger.WithComponent("ml_classifier").LogBusinessEvent(ctx, "ml_batch_optimized_classification_started", "", map[string]interface{}{
			"batch_size":      len(requests),
			"max_concurrency": c.batchProcessor.MaxConcurrency,
		})
	}

	// Process in batches
	var results []*MLClassificationResult
	batchSize := c.batchProcessor.BatchSize

	for i := 0; i < len(requests); i += batchSize {
		end := i + batchSize
		if end > len(requests) {
			end = len(requests)
		}

		batch := requests[i:end]
		batchResults, err := c.processBatchOptimized(ctx, batch)
		if err != nil {
			// Log batch error but continue with other batches
			if c.logger != nil {
				c.logger.WithComponent("ml_classifier").LogBusinessEvent(ctx, "ml_batch_optimized_classification_error", "", map[string]interface{}{
					"batch_start": i,
					"batch_end":   end,
					"error":       err.Error(),
				})
			}
			continue
		}

		results = append(results, batchResults...)
	}

	// Log batch completion
	if c.logger != nil {
		c.logger.WithComponent("ml_classifier").LogBusinessEvent(ctx, "ml_batch_optimized_classification_completed", "", map[string]interface{}{
			"total_requests":     len(requests),
			"successful_results": len(results),
			"processing_time_ms": time.Since(start).Milliseconds(),
		})
	}

	return results, nil
}

// performOptimizedInference performs inference using optimized models
func (c *MLClassifier) performOptimizedInference(ctx context.Context, request *MLClassificationRequest, optimizedModel *ModelCacheEntry) (*MLClassificationResult, error) {
	// Extract features (not used in simplified implementation but kept for future use)
	_, err := c.extractFeatures(request)
	if err != nil {
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Perform inference with optimized model
	// This is a simplified implementation - in practice, you'd use the optimized model data
	inferenceTime := time.Millisecond * 50 // Simulated inference time
	result := &MLClassificationResult{
		IndustryCode:    "541511", // Example industry code
		IndustryName:    "Custom Computer Programming Services",
		ConfidenceScore: 0.95,
		ModelType:       ModelTypeBERT,
		ModelVersion:    "v1.0.0_optimized",
		InferenceTime:   inferenceTime,
		ModelPredictions: []ModelPrediction{
			{
				ModelID:         "bert_optimized",
				ModelType:       ModelTypeBERT,
				IndustryCode:    "541511",
				IndustryName:    "Custom Computer Programming Services",
				ConfidenceScore: 0.95,
				RawScore:        0.95,
			},
		},
		EnsembleScore: 0.95,
		FeatureImportance: map[string]float64{
			"business_name": 0.3,
			"description":   0.4,
			"keywords":      0.2,
			"website":       0.1,
		},
		ProcessingMetadata: map[string]interface{}{
			"optimized_model":    true,
			"quantization_level": optimizedModel.QuantizationLevel,
			"cache_level":        optimizedModel.CacheLevel,
		},
	}

	return result, nil
}

// processBatchOptimized processes a batch of requests with optimization
func (c *MLClassifier) processBatchOptimized(ctx context.Context, requests []*MLClassificationRequest) ([]*MLClassificationResult, error) {
	// Create a channel for results
	resultChan := make(chan *MLClassificationResult, len(requests))
	errorChan := make(chan error, len(requests))

	// Process requests concurrently
	semaphore := make(chan struct{}, c.batchProcessor.MaxConcurrency)
	var wg sync.WaitGroup

	for _, request := range requests {
		wg.Add(1)
		go func(req *MLClassificationRequest) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			result, err := c.ClassifyWithOptimization(ctx, req)
			if err != nil {
				errorChan <- err
				return
			}
			resultChan <- result
		}(request)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	var results []*MLClassificationResult
	for result := range resultChan {
		results = append(results, result)
	}

	// Check for errors
	select {
	case err := <-errorChan:
		return results, err
	default:
		return results, nil
	}
}

// GetPerformanceMetrics returns performance metrics for the ML classifier
func (c *MLClassifier) GetPerformanceMetrics(ctx context.Context) map[string]interface{} {
	c.metricsMutex.RLock()
	defer c.metricsMutex.RUnlock()

	metrics := map[string]interface{}{
		"total_inferences":        len(c.inferenceTimes),
		"average_inference_time":  c.calculateAverageInferenceTime(),
		"accuracy_metrics":        c.accuracyMetrics,
		"batch_processor_enabled": c.batchProcessor.Enabled,
		"fallback_enabled":        c.fallbackEnabled,
		"cache_size":              len(c.resultCache),
	}

	// Add model optimizer metrics if available
	if c.modelOptimizer != nil {
		// GetStats method is not available, so we'll use a placeholder
		metrics["model_optimizer_enabled"] = true
	}

	return metrics
}

// calculateAverageInferenceTime calculates the average inference time
func (c *MLClassifier) calculateAverageInferenceTime() time.Duration {
	if len(c.inferenceTimes) == 0 {
		return 0
	}

	var total time.Duration
	for _, inferenceTime := range c.inferenceTimes {
		total += inferenceTime
	}

	return total / time.Duration(len(c.inferenceTimes))
}
