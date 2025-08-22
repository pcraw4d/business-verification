package classification_optimization

import (
	"fmt"
	"sync"

	"go.uber.org/zap"
)

// AlgorithmRegistry manages classification algorithms and their parameters
type AlgorithmRegistry struct {
	logger     *zap.Logger
	mu         sync.RWMutex
	algorithms map[string]*ClassificationAlgorithm
}

// ClassificationAlgorithm represents a classification algorithm with configurable parameters
type ClassificationAlgorithm struct {
	ID                    string                 `json:"id"`
	Name                  string                 `json:"name"`
	Category              string                 `json:"category"`
	Version               string                 `json:"version"`
	ConfidenceThreshold   float64                `json:"confidence_threshold"`
	FeatureWeights        map[string]float64     `json:"feature_weights"`
	ModelParameters       map[string]interface{} `json:"model_parameters"`
	PerformanceMetrics    *AlgorithmMetrics      `json:"performance_metrics"`
	IsActive              bool                   `json:"is_active"`
	LastOptimized         *int64                 `json:"last_optimized,omitempty"`
	OptimizationHistory   []*OptimizationResult  `json:"optimization_history"`
}

// NewAlgorithmRegistry creates a new algorithm registry
func NewAlgorithmRegistry(logger *zap.Logger) *AlgorithmRegistry {
	return &AlgorithmRegistry{
		logger:     logger,
		algorithms: make(map[string]*ClassificationAlgorithm),
	}
}

// RegisterAlgorithm registers a new classification algorithm
func (ar *AlgorithmRegistry) RegisterAlgorithm(algorithm *ClassificationAlgorithm) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if algorithm.ID == "" {
		return fmt.Errorf("algorithm ID cannot be empty")
	}

	if _, exists := ar.algorithms[algorithm.ID]; exists {
		return fmt.Errorf("algorithm with ID %s already exists", algorithm.ID)
	}

	// Set default values if not provided
	if algorithm.ConfidenceThreshold == 0 {
		algorithm.ConfidenceThreshold = 0.7
	}

	if algorithm.FeatureWeights == nil {
		algorithm.FeatureWeights = make(map[string]float64)
	}

	if algorithm.ModelParameters == nil {
		algorithm.ModelParameters = make(map[string]interface{})
	}

	if algorithm.PerformanceMetrics == nil {
		algorithm.PerformanceMetrics = &AlgorithmMetrics{
			Accuracy:           0.0,
			Precision:          0.0,
			Recall:             0.0,
			F1Score:            0.0,
			MisclassificationRate: 0.0,
			ConfidenceScore:    0.0,
			ProcessingTime:     0.0,
			Throughput:         0.0,
			ErrorRate:          0.0,
		}
	}

	ar.algorithms[algorithm.ID] = algorithm
	ar.logger.Info("Registered algorithm",
		zap.String("id", algorithm.ID),
		zap.String("name", algorithm.Name),
		zap.String("category", algorithm.Category))

	return nil
}

// GetAlgorithm returns an algorithm by ID
func (ar *AlgorithmRegistry) GetAlgorithm(id string) *ClassificationAlgorithm {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	algorithm, exists := ar.algorithms[id]
	if !exists {
		return nil
	}

	return ar.cloneAlgorithm(algorithm)
}

// GetAlgorithmByCategory returns an algorithm by category
func (ar *AlgorithmRegistry) GetAlgorithmByCategory(category string) *ClassificationAlgorithm {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	for _, algorithm := range ar.algorithms {
		if algorithm.Category == category && algorithm.IsActive {
			return ar.cloneAlgorithm(algorithm)
		}
	}

	return nil
}

// GetAllAlgorithms returns all registered algorithms
func (ar *AlgorithmRegistry) GetAllAlgorithms() map[string]*ClassificationAlgorithm {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	result := make(map[string]*ClassificationAlgorithm)
	for id, algorithm := range ar.algorithms {
		result[id] = ar.cloneAlgorithm(algorithm)
	}

	return result
}

// GetActiveAlgorithms returns all active algorithms
func (ar *AlgorithmRegistry) GetActiveAlgorithms() map[string]*ClassificationAlgorithm {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	result := make(map[string]*ClassificationAlgorithm)
	for id, algorithm := range ar.algorithms {
		if algorithm.IsActive {
			result[id] = ar.cloneAlgorithm(algorithm)
		}
	}

	return result
}

// UpdateAlgorithm updates an existing algorithm
func (ar *AlgorithmRegistry) UpdateAlgorithm(id string, updates map[string]interface{}) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	algorithm, exists := ar.algorithms[id]
	if !exists {
		return fmt.Errorf("algorithm with ID %s not found", id)
	}

	// Apply updates
	for key, value := range updates {
		switch key {
		case "confidence_threshold":
			if threshold, ok := value.(float64); ok {
				algorithm.ConfidenceThreshold = threshold
			}
		case "feature_weights":
			if weights, ok := value.(map[string]float64); ok {
				algorithm.FeatureWeights = weights
			}
		case "model_parameters":
			if params, ok := value.(map[string]interface{}); ok {
				algorithm.ModelParameters = params
			}
		case "is_active":
			if active, ok := value.(bool); ok {
				algorithm.IsActive = active
			}
		}
	}

	ar.logger.Info("Updated algorithm",
		zap.String("id", id),
		zap.Any("updates", updates))

	return nil
}

// SetConfidenceThreshold sets the confidence threshold for an algorithm
func (ar *AlgorithmRegistry) SetConfidenceThreshold(id string, threshold float64) error {
	return ar.UpdateAlgorithm(id, map[string]interface{}{
		"confidence_threshold": threshold,
	})
}

// GetConfidenceThreshold gets the confidence threshold for an algorithm
func (ar *AlgorithmRegistry) GetConfidenceThreshold(id string) (float64, error) {
	algorithm := ar.GetAlgorithm(id)
	if algorithm == nil {
		return 0, fmt.Errorf("algorithm with ID %s not found", id)
	}

	return algorithm.ConfidenceThreshold, nil
}

// SetFeatureWeights sets the feature weights for an algorithm
func (ar *AlgorithmRegistry) SetFeatureWeights(id string, weights map[string]float64) error {
	return ar.UpdateAlgorithm(id, map[string]interface{}{
		"feature_weights": weights,
	})
}

// GetFeatureWeights gets the feature weights for an algorithm
func (ar *AlgorithmRegistry) GetFeatureWeights(id string) (map[string]float64, error) {
	algorithm := ar.GetAlgorithm(id)
	if algorithm == nil {
		return nil, fmt.Errorf("algorithm with ID %s not found", id)
	}

	return algorithm.FeatureWeights, nil
}

// UpdatePerformanceMetrics updates performance metrics for an algorithm
func (ar *AlgorithmRegistry) UpdatePerformanceMetrics(id string, metrics *AlgorithmMetrics) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	algorithm, exists := ar.algorithms[id]
	if !exists {
		return fmt.Errorf("algorithm with ID %s not found", id)
	}

	algorithm.PerformanceMetrics = metrics
	return nil
}

// AddOptimizationResult adds an optimization result to an algorithm's history
func (ar *AlgorithmRegistry) AddOptimizationResult(id string, result *OptimizationResult) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	algorithm, exists := ar.algorithms[id]
	if !exists {
		return fmt.Errorf("algorithm with ID %s not found", id)
	}

	algorithm.OptimizationHistory = append(algorithm.OptimizationHistory, result)
	return nil
}

// GetOptimizationHistory returns optimization history for an algorithm
func (ar *AlgorithmRegistry) GetOptimizationHistory(id string) ([]*OptimizationResult, error) {
	algorithm := ar.GetAlgorithm(id)
	if algorithm == nil {
		return nil, fmt.Errorf("algorithm with ID %s not found", id)
	}

	return algorithm.OptimizationHistory, nil
}

// DeactivateAlgorithm deactivates an algorithm
func (ar *AlgorithmRegistry) DeactivateAlgorithm(id string) error {
	return ar.UpdateAlgorithm(id, map[string]interface{}{
		"is_active": false,
	})
}

// ActivateAlgorithm activates an algorithm
func (ar *AlgorithmRegistry) ActivateAlgorithm(id string) error {
	return ar.UpdateAlgorithm(id, map[string]interface{}{
		"is_active": true,
	})
}

// RemoveAlgorithm removes an algorithm from the registry
func (ar *AlgorithmRegistry) RemoveAlgorithm(id string) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if _, exists := ar.algorithms[id]; !exists {
		return fmt.Errorf("algorithm with ID %s not found", id)
	}

	delete(ar.algorithms, id)
	ar.logger.Info("Removed algorithm", zap.String("id", id))
	return nil
}

// GetAlgorithmSummary returns a summary of all algorithms
func (ar *AlgorithmRegistry) GetAlgorithmSummary() *AlgorithmRegistrySummary {
	ar.mu.RLock()
	defer ar.mu.RUnlock()

	summary := &AlgorithmRegistrySummary{
		TotalAlgorithms:     len(ar.algorithms),
		ActiveAlgorithms:    0,
		AlgorithmsByCategory: make(map[string]int),
		AverageAccuracy:     0.0,
		AverageConfidence:   0.0,
	}

	var totalAccuracy float64
	var totalConfidence float64
	algorithmCount := 0

	for _, algorithm := range ar.algorithms {
		if algorithm.IsActive {
			summary.ActiveAlgorithms++
		}

		summary.AlgorithmsByCategory[algorithm.Category]++

		if algorithm.PerformanceMetrics != nil {
			totalAccuracy += algorithm.PerformanceMetrics.Accuracy
			totalConfidence += algorithm.PerformanceMetrics.ConfidenceScore
			algorithmCount++
		}
	}

	if algorithmCount > 0 {
		summary.AverageAccuracy = totalAccuracy / float64(algorithmCount)
		summary.AverageConfidence = totalConfidence / float64(algorithmCount)
	}

	return summary
}

// AlgorithmRegistrySummary represents a summary of the algorithm registry
type AlgorithmRegistrySummary struct {
	TotalAlgorithms      int               `json:"total_algorithms"`
	ActiveAlgorithms     int               `json:"active_algorithms"`
	AlgorithmsByCategory map[string]int    `json:"algorithms_by_category"`
	AverageAccuracy      float64           `json:"average_accuracy"`
	AverageConfidence    float64           `json:"average_confidence"`
}

// cloneAlgorithm creates a deep copy of an algorithm
func (ar *AlgorithmRegistry) cloneAlgorithm(algorithm *ClassificationAlgorithm) *ClassificationAlgorithm {
	if algorithm == nil {
		return nil
	}

	cloned := &ClassificationAlgorithm{
		ID:                  algorithm.ID,
		Name:                algorithm.Name,
		Category:            algorithm.Category,
		Version:             algorithm.Version,
		ConfidenceThreshold: algorithm.ConfidenceThreshold,
		IsActive:            algorithm.IsActive,
		LastOptimized:       algorithm.LastOptimized,
	}

	// Clone feature weights
	if algorithm.FeatureWeights != nil {
		cloned.FeatureWeights = make(map[string]float64)
		for k, v := range algorithm.FeatureWeights {
			cloned.FeatureWeights[k] = v
		}
	}

	// Clone model parameters
	if algorithm.ModelParameters != nil {
		cloned.ModelParameters = make(map[string]interface{})
		for k, v := range algorithm.ModelParameters {
			cloned.ModelParameters[k] = v
		}
	}

	// Clone performance metrics
	if algorithm.PerformanceMetrics != nil {
		cloned.PerformanceMetrics = &AlgorithmMetrics{
			Accuracy:           algorithm.PerformanceMetrics.Accuracy,
			Precision:          algorithm.PerformanceMetrics.Precision,
			Recall:             algorithm.PerformanceMetrics.Recall,
			F1Score:            algorithm.PerformanceMetrics.F1Score,
			MisclassificationRate: algorithm.PerformanceMetrics.MisclassificationRate,
			ConfidenceScore:    algorithm.PerformanceMetrics.ConfidenceScore,
			ProcessingTime:     algorithm.PerformanceMetrics.ProcessingTime,
			Throughput:         algorithm.PerformanceMetrics.Throughput,
			ErrorRate:          algorithm.PerformanceMetrics.ErrorRate,
		}
	}

	// Clone optimization history
	if algorithm.OptimizationHistory != nil {
		cloned.OptimizationHistory = make([]*OptimizationResult, len(algorithm.OptimizationHistory))
		copy(cloned.OptimizationHistory, algorithm.OptimizationHistory)
	}

	return cloned
}

// GetConfidenceThreshold returns the confidence threshold for the algorithm
func (ca *ClassificationAlgorithm) GetConfidenceThreshold() float64 {
	return ca.ConfidenceThreshold
}

// SetConfidenceThreshold sets the confidence threshold for the algorithm
func (ca *ClassificationAlgorithm) SetConfidenceThreshold(threshold float64) {
	ca.ConfidenceThreshold = threshold
}
