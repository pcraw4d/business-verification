package classification

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/classification/methods"
	"github.com/pcraw4d/business-verification/internal/machine_learning"
	"github.com/pcraw4d/business-verification/internal/shared"
)

// MLIntegrationManager manages ML classifier integration with the ensemble system
type MLIntegrationManager struct {
	// Core components
	mlClassifier     *machine_learning.ContentClassifier
	methodRegistry   *MethodRegistry
	confidenceRouter *ConfidenceBasedRouter

	// Configuration
	config MLIntegrationConfig

	// Thread safety
	mutex sync.RWMutex

	// Logging
	logger *log.Logger

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// MLIntegrationConfig holds configuration for ML integration
type MLIntegrationConfig struct {
	// ML Method Configuration
	MLMethodEnabled  bool    `json:"ml_method_enabled"`
	MLMethodWeight   float64 `json:"ml_method_weight"`
	MLMethodPriority int     `json:"ml_method_priority"`

	// Confidence-based routing
	ConfidenceRoutingEnabled bool    `json:"confidence_routing_enabled"`
	HighConfidenceThreshold  float64 `json:"high_confidence_threshold"`
	LowConfidenceThreshold   float64 `json:"low_confidence_threshold"`

	// ML Model Configuration
	ModelType           string        `json:"model_type"`
	MaxSequenceLength   int           `json:"max_sequence_length"`
	BatchSize           int           `json:"batch_size"`
	ConfidenceThreshold float64       `json:"confidence_threshold"`
	ModelUpdateInterval time.Duration `json:"model_update_interval"`

	// Performance and monitoring
	PerformanceTracking bool `json:"performance_tracking"`
	ABTestingEnabled    bool `json:"ab_testing_enabled"`
	ModelVersioning     bool `json:"model_versioning"`

	// Fallback configuration
	FallbackToKeywordMethod bool    `json:"fallback_to_keyword_method"`
	FallbackConfidence      float64 `json:"fallback_confidence"`
}

// ConfidenceBasedRouter routes requests based on confidence thresholds
type ConfidenceBasedRouter struct {
	config ConfidenceRoutingConfig
	logger *log.Logger
}

// ConfidenceRoutingConfig holds configuration for confidence-based routing
type ConfidenceRoutingConfig struct {
	HighConfidenceThreshold float64 `json:"high_confidence_threshold"`
	LowConfidenceThreshold  float64 `json:"low_confidence_threshold"`
	MLMethodWeight          float64 `json:"ml_method_weight"`
	KeywordMethodWeight     float64 `json:"keyword_method_weight"`
	EnsembleMethodWeight    float64 `json:"ensemble_method_weight"`
}

// MLMethodRegistration represents registration of an ML method
type MLMethodRegistration struct {
	MethodName   string                 `json:"method_name"`
	ModelType    string                 `json:"model_type"`
	ModelVersion string                 `json:"model_version"`
	Weight       float64                `json:"weight"`
	Priority     int                    `json:"priority"`
	Enabled      bool                   `json:"enabled"`
	Config       map[string]interface{} `json:"config"`
	RegisteredAt time.Time              `json:"registered_at"`
	LastUsed     time.Time              `json:"last_used"`
}

// NewMLIntegrationManager creates a new ML integration manager
func NewMLIntegrationManager(
	mlClassifier *machine_learning.ContentClassifier,
	methodRegistry *MethodRegistry,
	config MLIntegrationConfig,
	logger *log.Logger,
) *MLIntegrationManager {
	if logger == nil {
		logger = log.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create confidence-based router
	confidenceRouter := NewConfidenceBasedRouter(ConfidenceRoutingConfig{
		HighConfidenceThreshold: config.HighConfidenceThreshold,
		LowConfidenceThreshold:  config.LowConfidenceThreshold,
		MLMethodWeight:          config.MLMethodWeight,
		KeywordMethodWeight:     0.5, // Default weight for keyword methods
		EnsembleMethodWeight:    0.3, // Default weight for ensemble methods
	}, logger)

	return &MLIntegrationManager{
		mlClassifier:     mlClassifier,
		methodRegistry:   methodRegistry,
		confidenceRouter: confidenceRouter,
		config:           config,
		logger:           logger,
		ctx:              ctx,
		cancel:           cancel,
	}
}

// NewConfidenceBasedRouter creates a new confidence-based router
func NewConfidenceBasedRouter(config ConfidenceRoutingConfig, logger *log.Logger) *ConfidenceBasedRouter {
	if logger == nil {
		logger = log.Default()
	}

	return &ConfidenceBasedRouter{
		config: config,
		logger: logger,
	}
}

// RegisterMLMethod registers an ML method with the ensemble system
func (mim *MLIntegrationManager) RegisterMLMethod(ctx context.Context) error {
	mim.mutex.Lock()
	defer mim.mutex.Unlock()

	if !mim.config.MLMethodEnabled {
		mim.logger.Printf("⚠️ ML method registration skipped - ML methods disabled")
		return nil
	}

	// Create ML classification method
	mlMethod := methods.NewMLClassificationMethod(mim.mlClassifier, mim.logger)

	// Set method weight and priority
	mlMethod.SetWeight(mim.config.MLMethodWeight)

	// Create method configuration
	methodConfig := methods.MethodConfig{
		Name:        "ml_classification",
		Type:        "ml",
		Weight:      mim.config.MLMethodWeight,
		Enabled:     true,
		Description: "Machine learning-based classification using BERT models",
		Config: map[string]interface{}{
			"model_type":            mim.config.ModelType,
			"max_sequence_length":   mim.config.MaxSequenceLength,
			"batch_size":            mim.config.BatchSize,
			"confidence_threshold":  mim.config.ConfidenceThreshold,
			"model_update_interval": mim.config.ModelUpdateInterval,
			"performance_tracking":  mim.config.PerformanceTracking,
			"ab_testing_enabled":    mim.config.ABTestingEnabled,
			"model_versioning":      mim.config.ModelVersioning,
			"priority":              mim.config.MLMethodPriority,
		},
	}

	// Register the method
	if err := mim.methodRegistry.RegisterMethod(mlMethod, methodConfig); err != nil {
		return fmt.Errorf("failed to register ML method: %w", err)
	}

	mim.logger.Printf("✅ Successfully registered ML method with weight %.2f and priority %d",
		mim.config.MLMethodWeight, mim.config.MLMethodPriority)

	return nil
}

// RouteByConfidence routes classification requests based on confidence thresholds
func (mim *MLIntegrationManager) RouteByConfidence(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.ClassificationMethodResult, error) {

	if !mim.config.ConfidenceRoutingEnabled {
		// If confidence routing is disabled, use standard ensemble
		return mim.classifyWithEnsemble(ctx, businessName, description, websiteURL)
	}

	// Get preliminary confidence from keyword method
	keywordResult, err := mim.getKeywordMethodConfidence(ctx, businessName, description, websiteURL)
	if err != nil {
		mim.logger.Printf("⚠️ Failed to get keyword method confidence: %v", err)
		// Fallback to ensemble
		return mim.classifyWithEnsemble(ctx, businessName, description, websiteURL)
	}

	// Route based on confidence
	confidence := keywordResult.Confidence

	if confidence >= mim.config.HighConfidenceThreshold {
		// High confidence - use keyword method result
		mim.logger.Printf("🎯 High confidence (%.2f) - using keyword method result", confidence)
		return keywordResult, nil

	} else if confidence <= mim.config.LowConfidenceThreshold {
		// Low confidence - use ML method
		mim.logger.Printf("🤖 Low confidence (%.2f) - routing to ML method", confidence)
		return mim.classifyWithMLMethod(ctx, businessName, description, websiteURL)

	} else {
		// Medium confidence - use ensemble
		mim.logger.Printf("⚖️ Medium confidence (%.2f) - using ensemble method", confidence)
		return mim.classifyWithEnsemble(ctx, businessName, description, websiteURL)
	}
}

// classifyWithMLMethod performs classification using ML method
func (mim *MLIntegrationManager) classifyWithMLMethod(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.ClassificationMethodResult, error) {

	// Get ML method from registry
	mlMethod, err := mim.methodRegistry.GetMethod("ml_classification")
	if err != nil {
		return nil, fmt.Errorf("ML method not found: %w", err)
	}

	// Perform ML classification
	result, err := mlMethod.Classify(ctx, businessName, description, websiteURL)
	if err != nil {
		// If ML method fails and fallback is enabled, use keyword method
		if mim.config.FallbackToKeywordMethod {
			mim.logger.Printf("⚠️ ML method failed, falling back to keyword method: %v", err)
			return mim.getKeywordMethodConfidence(ctx, businessName, description, websiteURL)
		}
		return nil, fmt.Errorf("ML classification failed: %w", err)
	}

	return result, nil
}

// classifyWithEnsemble performs classification using ensemble of methods
func (mim *MLIntegrationManager) classifyWithEnsemble(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.ClassificationMethodResult, error) {

	// Get all enabled methods
	enabledMethods := mim.methodRegistry.GetEnabledMethods()
	if len(enabledMethods) == 0 {
		return nil, fmt.Errorf("no enabled classification methods found")
	}

	// Perform classification with all methods
	var methodResults []shared.ClassificationMethodResult
	var totalWeight float64

	for _, method := range enabledMethods {
		result, err := method.Classify(ctx, businessName, description, websiteURL)
		if err != nil {
			mim.logger.Printf("⚠️ Method %s failed: %v", method.GetName(), err)
			continue
		}

		if result.Success {
			methodResults = append(methodResults, *result)
			totalWeight += method.GetWeight()
		}
	}

	if len(methodResults) == 0 {
		return nil, fmt.Errorf("all classification methods failed")
	}

	// Calculate ensemble result
	ensembleResult := mim.calculateEnsembleResult(methodResults, totalWeight)

	return ensembleResult, nil
}

// getKeywordMethodConfidence gets confidence from keyword method only
func (mim *MLIntegrationManager) getKeywordMethodConfidence(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*shared.ClassificationMethodResult, error) {

	// Get keyword method from registry
	keywordMethod, err := mim.methodRegistry.GetMethod("keyword_classification")
	if err != nil {
		return nil, fmt.Errorf("keyword method not found: %w", err)
	}

	// Perform keyword classification
	result, err := keywordMethod.Classify(ctx, businessName, description, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("keyword classification failed: %w", err)
	}

	return result, nil
}

// calculateEnsembleResult calculates the ensemble result from multiple method results
func (mim *MLIntegrationManager) calculateEnsembleResult(
	methodResults []shared.ClassificationMethodResult,
	totalWeight float64,
) *shared.ClassificationMethodResult {

	if len(methodResults) == 0 {
		return &shared.ClassificationMethodResult{
			MethodName:     "ensemble",
			MethodType:     "ensemble",
			Confidence:     mim.config.FallbackConfidence,
			ProcessingTime: 0,
			Success:        false,
			Error:          "no successful method results",
		}
	}

	// Calculate weighted average
	var weightedConfidence float64
	var bestResult *shared.ClassificationMethodResult
	var bestScore float64

	for _, result := range methodResults {
		weight := result.Confidence * (result.Confidence / totalWeight) // Weight by confidence and method weight
		weightedConfidence += weight

		if result.Confidence > bestScore {
			bestScore = result.Confidence
			bestResult = &result
		}
	}

	// Use the best result as base and adjust confidence
	if bestResult != nil {
		ensembleResult := *bestResult
		ensembleResult.MethodName = "ensemble"
		ensembleResult.MethodType = "ensemble"
		ensembleResult.Confidence = weightedConfidence
		ensembleResult.Evidence = append(ensembleResult.Evidence,
			fmt.Sprintf("Ensemble result from %d methods with weighted confidence %.2f",
				len(methodResults), weightedConfidence))

		return &ensembleResult
	}

	// Fallback
	return &shared.ClassificationMethodResult{
		MethodName:     "ensemble",
		MethodType:     "ensemble",
		Confidence:     mim.config.FallbackConfidence,
		ProcessingTime: 0,
		Success:        true,
		Result: &shared.IndustryClassification{
			IndustryCode:         "unknown",
			IndustryName:         "Unknown",
			ConfidenceScore:      mim.config.FallbackConfidence,
			ClassificationMethod: "ensemble",
			Description:          "Ensemble classification result",
		},
	}
}

// GetMLMethodInfo returns information about registered ML methods
func (mim *MLIntegrationManager) GetMLMethodInfo(ctx context.Context) ([]MLMethodRegistration, error) {
	mim.mutex.RLock()
	defer mim.mutex.RUnlock()

	var mlMethods []MLMethodRegistration

	// Get all methods from registry
	allMethods := mim.methodRegistry.GetAllMethods()

	for _, method := range allMethods {
		if method.GetType() == "ml" {
			mlMethods = append(mlMethods, MLMethodRegistration{
				MethodName:   method.GetName(),
				ModelType:    "bert", // Default model type
				ModelVersion: "1.0",  // Default version
				Weight:       method.GetWeight(),
				Priority:     1, // Default priority
				Enabled:      method.IsEnabled(),
				Config: map[string]interface{}{
					"description": method.GetDescription(),
				},
				RegisteredAt: time.Now(),
				LastUsed:     time.Now(),
			})
		}
	}

	return mlMethods, nil
}

// UpdateMLMethodWeight updates the weight of an ML method
func (mim *MLIntegrationManager) UpdateMLMethodWeight(methodName string, newWeight float64) error {
	mim.mutex.Lock()
	defer mim.mutex.Unlock()

	method, err := mim.methodRegistry.GetMethod(methodName)
	if err != nil {
		return fmt.Errorf("method not found: %w", err)
	}

	if method.GetType() != "ml" {
		return fmt.Errorf("method %s is not an ML method", methodName)
	}

	method.SetWeight(newWeight)

	mim.logger.Printf("✅ Updated ML method %s weight to %.2f", methodName, newWeight)
	return nil
}

// HealthCheck performs a health check on ML integration
func (mim *MLIntegrationManager) HealthCheck(ctx context.Context) error {
	// Check ML classifier
	if mim.mlClassifier == nil {
		return fmt.Errorf("ML classifier is nil")
	}

	// Check method registry
	if mim.methodRegistry == nil {
		return fmt.Errorf("method registry is nil")
	}

	// Check if ML method is registered
	_, err := mim.methodRegistry.GetMethod("ml_classification")
	if err != nil {
		return fmt.Errorf("ML method not registered: %w", err)
	}

	return nil
}

// Stop stops the ML integration manager
func (mim *MLIntegrationManager) Stop() {
	mim.logger.Printf("🛑 Stopping ML integration manager")
	mim.cancel()
}
