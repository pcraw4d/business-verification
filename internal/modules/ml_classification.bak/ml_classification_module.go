package ml_classification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"kyb-platform/internal/architecture"
	"kyb-platform/internal/config"
	"kyb-platform/internal/database"
	"kyb-platform/internal/observability"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MLClassificationModule implements the Module interface for ML-based classification
type MLClassificationModule struct {
	id        string
	config    architecture.ModuleConfig
	running   bool
	logger    *observability.ModuleLogger
	metrics   *observability.Metrics
	tracer    trace.Tracer
	db        database.Database
	appConfig *config.Config

	// ML classification specific fields
	modelManager     *ModelManager
	modelOptimizer   *ModelOptimizer
	ensembleConfig   *EnsembleConfig
	featureExtractor *FeatureExtractor
	batchProcessor   *BatchProcessor

	// Caching
	resultCache map[string]*MLClassificationResult
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration

	// Performance tracking
	inferenceTimes  map[string]time.Duration
	accuracyMetrics map[string]float64
	metricsMutex    sync.RWMutex

	// Fallback mechanisms
	fallbackEnabled bool
	fallbackModels  []string
}

// NewMLClassificationModule creates a new ML classification module
func NewMLClassificationModule() *MLClassificationModule {
	return &MLClassificationModule{
		id: "ml_classification_module",

		// Initialize caching
		resultCache: make(map[string]*MLClassificationResult),
		cacheTTL:    1 * time.Hour,

		// Initialize performance tracking
		inferenceTimes:  make(map[string]time.Duration),
		accuracyMetrics: make(map[string]float64),

		// Initialize fallback mechanisms
		fallbackEnabled: true,
		fallbackModels:  []string{"bert_base", "ensemble_fallback"},
	}
}

// Module interface implementation
func (m *MLClassificationModule) ID() string {
	return m.id
}

func (m *MLClassificationModule) Config() architecture.ModuleConfig {
	return m.config
}

func (m *MLClassificationModule) UpdateConfig(config architecture.ModuleConfig) error {
	m.config = config
	return nil
}

func (m *MLClassificationModule) Health() architecture.ModuleHealth {
	status := architecture.ModuleStatusStopped
	if m.running {
		status = architecture.ModuleStatusRunning
	}

	return architecture.ModuleHealth{
		Status:    status,
		LastCheck: time.Now(),
		Message:   "ML classification module health check",
	}
}

func (m *MLClassificationModule) Metadata() architecture.ModuleMetadata {
	return architecture.ModuleMetadata{
		Name:        "ML Classification Module",
		Version:     "1.0.0",
		Description: "Performs business classification using machine learning models",
		Capabilities: []architecture.ModuleCapability{
			architecture.CapabilityClassification,
			architecture.CapabilityMLPrediction,
		},
		Priority: architecture.PriorityHigh,
	}
}

func (m *MLClassificationModule) Start(ctx context.Context) error {
	_, span := m.tracer.Start(ctx, "MLClassificationModule.Start")
	defer span.End()

	if m.running {
		return fmt.Errorf("module %s is already running", m.id)
	}

	// Initialize ML components
	if err := m.initializeMLComponents(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to initialize ML components: %w", err)
	}

	// Initialize ensemble configuration
	if err := m.initializeEnsembleConfig(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to initialize ensemble config: %w", err)
	}

	// Initialize feature extractor
	if err := m.initializeFeatureExtractor(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to initialize feature extractor: %w", err)
	}

	// Initialize batch processor
	if err := m.initializeBatchProcessor(); err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to initialize batch processor: %w", err)
	}

	m.running = true

	// Emit module started event
	if emitEvent != nil {
		emitEvent(architecture.Event{
			Type:     architecture.EventTypeModuleStarted,
			Source:   m.id,
			Priority: architecture.EventPriorityNormal,
			Data: map[string]interface{}{
				"module_id":  m.id,
				"start_time": time.Now(),
			},
		})
	}

	span.SetAttributes(attribute.String("module.id", m.id))
	m.logger.LogModuleStart(ctx, map[string]interface{}{
		"module_id": m.id,
	})

	return nil
}

func (m *MLClassificationModule) Stop(ctx context.Context) error {
	_, span := m.tracer.Start(ctx, "MLClassificationModule.Stop")
	defer span.End()

	if !m.running {
		return fmt.Errorf("module %s is not running", m.id)
	}

	m.running = false

	// Emit module stopped event
	if emitEvent != nil {
		emitEvent(architecture.Event{
			Type:     architecture.EventTypeModuleStopped,
			Source:   m.id,
			Priority: architecture.EventPriorityNormal,
			Data: map[string]interface{}{
				"module_id": m.id,
				"stop_time": time.Now(),
			},
		})
	}

	span.SetAttributes(attribute.String("module.id", m.id))
	m.logger.LogModuleStop(ctx, "graceful shutdown")

	return nil
}

func (m *MLClassificationModule) IsRunning() bool {
	return m.running
}

func (m *MLClassificationModule) Process(ctx context.Context, req architecture.ModuleRequest) (architecture.ModuleResponse, error) {
	_, span := m.tracer.Start(ctx, "MLClassificationModule.Process")
	defer span.End()

	span.SetAttributes(
		attribute.String("module.id", m.id),
		attribute.String("request.type", req.Type),
	)

	// Check if this module can handle the request
	if !m.CanHandle(req) {
		return architecture.ModuleResponse{
			ID:      req.ID,
			Success: false,
			Error:   "unsupported request type",
		}, nil
	}

	// Parse the request payload
	classificationReq, err := m.parseClassificationRequest(req.Data)
	if err != nil {
		span.RecordError(err)
		return architecture.ModuleResponse{
			ID:      req.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to parse request: %v", err),
		}, nil
	}

	// Perform ML classification
	result, err := m.performMLClassification(ctx, classificationReq)
	if err != nil {
		span.RecordError(err)
		return architecture.ModuleResponse{
			ID:      req.ID,
			Success: false,
			Error:   fmt.Sprintf("ML classification failed: %v", err),
		}, nil
	}

	// Create response
	response := architecture.ModuleResponse{
		ID:      req.ID,
		Success: true,
		Data: map[string]interface{}{
			"classification": result,
			"method":         "ml_classification",
			"module_id":      m.id,
		},
	}

	// Emit classification completed event
	if emitEvent != nil {
		emitEvent(architecture.Event{
			Type:     architecture.EventTypeClassificationCompleted,
			Source:   m.id,
			Priority: architecture.EventPriorityNormal,
			Data: map[string]interface{}{
				"business_name": classificationReq.BusinessName,
				"method":        "ml_classification",
				"model_type":    string(result.ModelType),
				"confidence":    result.ConfidenceScore,
			},
		})
	}

	// Record metrics
	m.metrics.RecordBusinessClassification("ml", 1.0)

	return response, nil
}

func (m *MLClassificationModule) CanHandle(req architecture.ModuleRequest) bool {
	return req.Type == "classify_by_ml"
}

func (m *MLClassificationModule) HealthCheck(ctx context.Context) error {
	_, span := m.tracer.Start(ctx, "MLClassificationModule.HealthCheck")
	defer span.End()

	if !m.running {
		return fmt.Errorf("module is not running")
	}

	// Check if model manager is initialized
	if m.modelManager == nil {
		return fmt.Errorf("model manager not initialized")
	}

	// Check if at least one model is available
	models, err := m.modelManager.GetAvailableModels(ctx)
	if err != nil || len(models) == 0 {
		return fmt.Errorf("no available models for classification")
	}

	span.SetAttributes(attribute.String("module.id", m.id))
	return nil
}

func (m *MLClassificationModule) OnEvent(event architecture.ModuleEvent) error {
	// Handle events if needed
	return nil
}

// ML classification specific methods

// ClassificationRequest represents an ML classification request
type ClassificationRequest struct {
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

// ModelType represents the type of ML model
type ModelType string

const (
	ModelTypeBERT        ModelType = "bert"
	ModelTypeEnsemble    ModelType = "ensemble"
	ModelTypeTransformer ModelType = "transformer"
	ModelTypeCustom      ModelType = "custom"
)

// ModelManager provides ML model management capabilities
type ModelManager struct {
	models map[string]*ModelInfo
}

// GetAvailableModels returns all available models
func (mm *ModelManager) GetAvailableModels(ctx context.Context) ([]*ModelInfo, error) {
	models := make([]*ModelInfo, 0, len(mm.models))
	for _, model := range mm.models {
		if model.Status == "ready" {
			models = append(models, model)
		}
	}
	return models, nil
}

// ModelInfo represents information about a loaded model
type ModelInfo struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        ModelType              `json:"type"`
	Version     string                 `json:"version"`
	Status      string                 `json:"status"`
	LoadedAt    time.Time              `json:"loaded_at"`
	LastUsed    time.Time              `json:"last_used"`
	UsageCount  int64                  `json:"usage_count"`
	Performance *ModelPerformance      `json:"performance,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ModelPerformance represents performance metrics for a model
type ModelPerformance struct {
	Accuracy        float64   `json:"accuracy"`
	Precision       float64   `json:"precision"`
	Recall          float64   `json:"recall"`
	F1Score         float64   `json:"f1_score"`
	InferenceTime   float64   `json:"inference_time_ms"`
	Throughput      float64   `json:"throughput_requests_per_sec"`
	MemoryUsage     float64   `json:"memory_usage_mb"`
	LastEvaluated   time.Time `json:"last_evaluated"`
	EvaluationCount int       `json:"evaluation_count"`
}

// ModelOptimizer provides model optimization capabilities
type ModelOptimizer struct {
	enabled bool
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

// parseClassificationRequest parses the module request into a classification request
func (m *MLClassificationModule) parseClassificationRequest(payload map[string]interface{}) (*ClassificationRequest, error) {
	req := &ClassificationRequest{}

	if businessName, ok := payload["business_name"].(string); ok {
		req.BusinessName = businessName
	}

	if businessDescription, ok := payload["business_description"].(string); ok {
		req.BusinessDescription = businessDescription
	}

	if keywords, ok := payload["keywords"].([]interface{}); ok {
		req.Keywords = make([]string, len(keywords))
		for i, keyword := range keywords {
			if str, ok := keyword.(string); ok {
				req.Keywords[i] = str
			}
		}
	}

	if websiteContent, ok := payload["website_content"].(string); ok {
		req.WebsiteContent = websiteContent
	}

	if industryHints, ok := payload["industry_hints"].([]interface{}); ok {
		req.IndustryHints = make([]string, len(industryHints))
		for i, hint := range industryHints {
			if str, ok := hint.(string); ok {
				req.IndustryHints[i] = str
			}
		}
	}

	if geographicRegion, ok := payload["geographic_region"].(string); ok {
		req.GeographicRegion = geographicRegion
	}

	if businessType, ok := payload["business_type"].(string); ok {
		req.BusinessType = businessType
	}

	if metadata, ok := payload["metadata"].(map[string]interface{}); ok {
		req.Metadata = metadata
	} else {
		req.Metadata = make(map[string]interface{})
	}

	return req, nil
}

// performMLClassification performs ML-based classification
func (m *MLClassificationModule) performMLClassification(ctx context.Context, req *ClassificationRequest) (*MLClassificationResult, error) {
	_, span := m.tracer.Start(ctx, "performMLClassification")
	defer span.End()

	span.SetAttributes(attribute.String("business_name", req.BusinessName))

	// Check cache first
	cacheKey := m.generateCacheKey(req)
	if cached, exists := m.getFromCache(cacheKey); exists {
		span.AddEvent("Cache hit")
		return cached, nil
	}

	// Extract features
	features, err := m.extractFeatures(req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to extract features: %w", err)
	}

	// Perform ensemble classification
	result, err := m.performEnsembleClassification(ctx, features, req)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to perform ensemble classification: %w", err)
	}

	// Calculate ensemble score
	result.EnsembleScore = m.calculateEnsembleScore(result.ModelPredictions)

	// Add processing metadata
	result.ProcessingMetadata = map[string]interface{}{
		"feature_count":   len(features),
		"models_used":     len(result.ModelPredictions),
		"ensemble_method": m.ensembleConfig.Method,
		"cache_key":       cacheKey,
	}

	// Cache the result
	m.cacheResult(cacheKey, result)

	span.SetAttributes(
		attribute.String("model_type", string(result.ModelType)),
		attribute.Float64("confidence_score", result.ConfidenceScore),
		attribute.Int("models_used", len(result.ModelPredictions)),
	)

	return result, nil
}

// generateCacheKey generates a cache key for the request
func (m *MLClassificationModule) generateCacheKey(req *ClassificationRequest) string {
	// Create a hash of the request data for caching
	data := fmt.Sprintf("%s|%s|%v|%s|%v|%s|%s",
		req.BusinessName,
		req.BusinessDescription,
		req.Keywords,
		req.WebsiteContent,
		req.IndustryHints,
		req.GeographicRegion,
		req.BusinessType,
	)

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// getFromCache retrieves a result from cache
func (m *MLClassificationModule) getFromCache(cacheKey string) (*MLClassificationResult, bool) {
	m.cacheMutex.RLock()
	defer m.cacheMutex.RUnlock()

	if result, exists := m.resultCache[cacheKey]; exists {
		// Check if cache entry is still valid
		if time.Since(result.ProcessingMetadata["timestamp"].(time.Time)) < m.cacheTTL {
			return result, true
		}
		// Remove expired entry
		delete(m.resultCache, cacheKey)
	}

	return nil, false
}

// cacheResult stores a result in cache
func (m *MLClassificationModule) cacheResult(cacheKey string, result *MLClassificationResult) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	// Add timestamp to metadata
	if result.ProcessingMetadata == nil {
		result.ProcessingMetadata = make(map[string]interface{})
	}
	result.ProcessingMetadata["timestamp"] = time.Now()

	m.resultCache[cacheKey] = result
}

// extractFeatures extracts features from the classification request
func (m *MLClassificationModule) extractFeatures(req *ClassificationRequest) (map[string]interface{}, error) {
	features := make(map[string]interface{})

	// Text features
	if m.featureExtractor.TextFeatures {
		features["business_name"] = req.BusinessName
		features["business_description"] = req.BusinessDescription
		features["website_content"] = req.WebsiteContent
		features["keywords"] = req.Keywords
	}

	// Semantic features
	if m.featureExtractor.SemanticFeatures {
		features["industry_hints"] = req.IndustryHints
		features["business_type"] = req.BusinessType
	}

	// Statistical features
	if m.featureExtractor.StatisticalFeatures {
		features["name_length"] = len(req.BusinessName)
		features["description_length"] = len(req.BusinessDescription)
		features["content_length"] = len(req.WebsiteContent)
		features["keyword_count"] = len(req.Keywords)
	}

	// Domain features
	if m.featureExtractor.DomainFeatures {
		features["geographic_region"] = req.GeographicRegion
		features["has_website"] = req.WebsiteContent != ""
		features["has_description"] = req.BusinessDescription != ""
	}

	return features, nil
}

// performEnsembleClassification performs classification using multiple models
func (m *MLClassificationModule) performEnsembleClassification(ctx context.Context, features map[string]interface{}, req *ClassificationRequest) (*MLClassificationResult, error) {
	start := time.Now()

	// Get available models
	models, err := m.getAvailableModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available models: %w", err)
	}

	// Perform predictions with each model
	predictions := make([]ModelPrediction, 0)
	for _, model := range models {
		prediction, err := m.predictWithModel(ctx, model, features, req)
		if err != nil {
			m.logger.LogModuleError(ctx, "model_prediction", err, map[string]interface{}{
				"model_id": model.ID,
			})
			continue
		}
		predictions = append(predictions, *prediction)
	}

	if len(predictions) == 0 {
		return nil, fmt.Errorf("no successful predictions from any model")
	}

	// Select best prediction based on ensemble method
	var bestPrediction ModelPrediction
	switch m.ensembleConfig.Method {
	case "weighted_average":
		bestPrediction = m.selectWeightedAverage(predictions)
	case "voting":
		bestPrediction = m.selectVoting(predictions)
	case "stacking":
		bestPrediction = m.selectStacking(predictions)
	default:
		bestPrediction = m.selectBestSingle(predictions)
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
		FeatureImportance: m.calculateFeatureImportance(features),
	}

	return result, nil
}

// getAvailableModels gets available models for classification
func (m *MLClassificationModule) getAvailableModels(ctx context.Context) ([]*ModelInfo, error) {
	if m.modelManager == nil {
		return nil, fmt.Errorf("model manager not initialized")
	}

	return m.modelManager.GetAvailableModels(ctx)
}

// predictWithModel performs prediction with a single model
func (m *MLClassificationModule) predictWithModel(ctx context.Context, model *ModelInfo, features map[string]interface{}, req *ClassificationRequest) (*ModelPrediction, error) {
	start := time.Now()

	// Simulate prediction based on model type
	var prediction *ModelPrediction

	switch model.Type {
	case ModelTypeBERT:
		prediction = m.simulateBERTPrediction(model, features, req)
	case ModelTypeEnsemble:
		prediction = m.simulateEnsemblePrediction(model, features, req)
	case ModelTypeTransformer:
		prediction = m.simulateTransformerPrediction(model, features, req)
	default:
		prediction = m.simulateCustomPrediction(model, features, req)
	}

	// Update inference time
	inferenceTime := time.Since(start)
	m.metricsMutex.Lock()
	m.inferenceTimes[model.ID] = inferenceTime
	m.metricsMutex.Unlock()

	return prediction, nil
}

// simulateBERTPrediction simulates a BERT model prediction
func (m *MLClassificationModule) simulateBERTPrediction(model *ModelInfo, features map[string]interface{}, req *ClassificationRequest) *ModelPrediction {
	// Simulate BERT-based classification
	// In a real implementation, this would call the actual BERT model

	// Simple heuristic based on business name and description
	confidence := 0.85
	industryCode := "511210" // Technology
	industryName := "Technology"

	if req.BusinessDescription != "" {
		if containsAny(req.BusinessDescription, []string{"bank", "financial", "credit"}) {
			industryCode = "522110"
			industryName = "Financial Services"
			confidence = 0.90
		} else if containsAny(req.BusinessDescription, []string{"health", "medical", "hospital"}) {
			industryCode = "621111"
			industryName = "Healthcare"
			confidence = 0.88
		}
	}

	return &ModelPrediction{
		ModelID:         model.ID,
		ModelType:       model.Type,
		IndustryCode:    industryCode,
		IndustryName:    industryName,
		ConfidenceScore: confidence,
		RawScore:        confidence,
	}
}

// simulateEnsemblePrediction simulates an ensemble model prediction
func (m *MLClassificationModule) simulateEnsemblePrediction(model *ModelInfo, features map[string]interface{}, req *ClassificationRequest) *ModelPrediction {
	// Simulate ensemble-based classification
	confidence := 0.92
	industryCode := "511210" // Technology
	industryName := "Technology"

	// Ensemble logic would combine multiple model predictions
	if len(req.IndustryHints) > 0 {
		confidence = 0.95
	}

	return &ModelPrediction{
		ModelID:         model.ID,
		ModelType:       model.Type,
		IndustryCode:    industryCode,
		IndustryName:    industryName,
		ConfidenceScore: confidence,
		RawScore:        confidence,
	}
}

// simulateTransformerPrediction simulates a transformer model prediction
func (m *MLClassificationModule) simulateTransformerPrediction(model *ModelInfo, features map[string]interface{}, req *ClassificationRequest) *ModelPrediction {
	// Simulate transformer-based classification
	confidence := 0.87
	industryCode := "511210" // Technology
	industryName := "Technology"

	return &ModelPrediction{
		ModelID:         model.ID,
		ModelType:       model.Type,
		IndustryCode:    industryCode,
		IndustryName:    industryName,
		ConfidenceScore: confidence,
		RawScore:        confidence,
	}
}

// simulateCustomPrediction simulates a custom model prediction
func (m *MLClassificationModule) simulateCustomPrediction(model *ModelInfo, features map[string]interface{}, req *ClassificationRequest) *ModelPrediction {
	// Simulate custom model prediction
	confidence := 0.80
	industryCode := "511210" // Technology
	industryName := "Technology"

	return &ModelPrediction{
		ModelID:         model.ID,
		ModelType:       model.Type,
		IndustryCode:    industryCode,
		IndustryName:    industryName,
		ConfidenceScore: confidence,
		RawScore:        confidence,
	}
}

// selectWeightedAverage selects the best prediction using weighted average
func (m *MLClassificationModule) selectWeightedAverage(predictions []ModelPrediction) ModelPrediction {
	var bestPrediction ModelPrediction
	var bestScore float64

	for _, pred := range predictions {
		weight := m.ensembleConfig.ModelWeights[pred.ModelID]
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
func (m *MLClassificationModule) selectVoting(predictions []ModelPrediction) ModelPrediction {
	// Simple voting - return the prediction with highest confidence
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

// selectStacking selects the best prediction using stacking
func (m *MLClassificationModule) selectStacking(predictions []ModelPrediction) ModelPrediction {
	// Simple stacking - average the predictions
	if len(predictions) == 0 {
		return ModelPrediction{}
	}

	var totalConfidence float64
	var bestPrediction ModelPrediction
	var bestConfidence float64

	for _, pred := range predictions {
		totalConfidence += pred.ConfidenceScore
		if pred.ConfidenceScore > bestConfidence {
			bestConfidence = pred.ConfidenceScore
			bestPrediction = pred
		}
	}

	// Update confidence to average
	bestPrediction.ConfidenceScore = totalConfidence / float64(len(predictions))
	return bestPrediction
}

// selectBestSingle selects the best single prediction
func (m *MLClassificationModule) selectBestSingle(predictions []ModelPrediction) ModelPrediction {
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

// calculateEnsembleScore calculates the ensemble score from predictions
func (m *MLClassificationModule) calculateEnsembleScore(predictions []ModelPrediction) float64 {
	if len(predictions) == 0 {
		return 0.0
	}

	var totalScore float64
	for _, pred := range predictions {
		totalScore += pred.ConfidenceScore
	}

	return totalScore / float64(len(predictions))
}

// calculateFeatureImportance calculates feature importance
func (m *MLClassificationModule) calculateFeatureImportance(features map[string]interface{}) map[string]float64 {
	importance := make(map[string]float64)

	// Simple feature importance calculation
	for feature, value := range features {
		switch v := value.(type) {
		case string:
			importance[feature] = float64(len(v)) / 100.0 // Normalize by length
		case int:
			importance[feature] = float64(v) / 100.0 // Normalize by value
		case []string:
			importance[feature] = float64(len(v)) / 10.0 // Normalize by count
		default:
			importance[feature] = 0.5 // Default importance
		}
	}

	return importance
}

// initializeMLComponents initializes ML components
func (m *MLClassificationModule) initializeMLComponents() error {
	// Initialize model manager
	m.modelManager = &ModelManager{
		models: make(map[string]*ModelInfo),
	}

	// Add some default models
	m.modelManager.models["bert_base"] = &ModelInfo{
		ID:       "bert_base",
		Name:     "BERT Base Model",
		Type:     ModelTypeBERT,
		Version:  "1.0.0",
		Status:   "ready",
		LoadedAt: time.Now(),
	}

	m.modelManager.models["ensemble_model"] = &ModelInfo{
		ID:       "ensemble_model",
		Name:     "Ensemble Model",
		Type:     ModelTypeEnsemble,
		Version:  "1.0.0",
		Status:   "ready",
		LoadedAt: time.Now(),
	}

	// Initialize model optimizer
	m.modelOptimizer = &ModelOptimizer{
		enabled: true,
	}

	return nil
}

// initializeEnsembleConfig initializes ensemble configuration
func (m *MLClassificationModule) initializeEnsembleConfig() error {
	m.ensembleConfig = &EnsembleConfig{
		Enabled:          true,
		Method:           "weighted_average",
		ModelWeights:     make(map[string]float64),
		MinConfidence:    0.5,
		MaxModels:        5,
		FallbackStrategy: "best_single_model",
	}

	// Set default model weights
	m.ensembleConfig.ModelWeights["bert_base"] = 0.4
	m.ensembleConfig.ModelWeights["ensemble_model"] = 0.6

	return nil
}

// initializeFeatureExtractor initializes feature extractor
func (m *MLClassificationModule) initializeFeatureExtractor() error {
	m.featureExtractor = &FeatureExtractor{
		TextFeatures:        true,
		SemanticFeatures:    true,
		StatisticalFeatures: true,
		DomainFeatures:      true,
	}
	return nil
}

// initializeBatchProcessor initializes batch processor
func (m *MLClassificationModule) initializeBatchProcessor() error {
	m.batchProcessor = &BatchProcessor{
		Enabled:        true,
		BatchSize:      32,
		MaxConcurrency: 4,
		Timeout:        30 * time.Second,
		RetryAttempts:  3,
	}
	return nil
}

// containsAny checks if a string contains any of the given substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

// containsSubstring checks if a string contains a substring (case-insensitive)
func containsSubstring(s, substr string) bool {
	sLower := toLower(s)
	substrLower := toLower(substr)
	return len(sLower) >= len(substrLower) &&
		(sLower == substrLower ||
			len(sLower) > len(substrLower) &&
				(sLower[:len(substrLower)] == substrLower ||
					sLower[len(sLower)-len(substrLower):] == substrLower ||
					containsSubstringHelper(sLower, substrLower)))
}

// containsSubstringHelper helper function for substring search
func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toLower converts string to lowercase
func toLower(s string) string {
	// Simple lowercase conversion for ASCII
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// Event emission function (will be injected by the module manager)
var emitEvent func(architecture.Event) error

// SetEventEmitter sets the event emission function
func (m *MLClassificationModule) SetEventEmitter(emitter func(architecture.Event) error) {
	emitEvent = emitter
}
