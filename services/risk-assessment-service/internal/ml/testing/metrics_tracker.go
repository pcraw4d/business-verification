package testing

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PredictionRecord represents a single prediction record for metrics tracking
type PredictionRecord struct {
	RequestID     string                 `json:"request_id"`
	ModelID       string                 `json:"model_id"`
	ExperimentID  string                 `json:"experiment_id"`
	Input         map[string]interface{} `json:"input"`
	Prediction    interface{}            `json:"prediction"`
	Confidence    float64                `json:"confidence"`
	Latency       time.Duration          `json:"latency"`
	Timestamp     time.Time              `json:"timestamp"`
	ActualOutcome interface{}            `json:"actual_outcome,omitempty"` // For validation
	IsError       bool                   `json:"is_error"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
}

// ModelMetrics represents aggregated metrics for a model
type ModelMetrics struct {
	ModelID         string             `json:"model_id"`
	ExperimentID    string             `json:"experiment_id"`
	RequestCount    int64              `json:"request_count"`
	ErrorCount      int64              `json:"error_count"`
	Metrics         map[string]float64 `json:"metrics"`
	AverageLatency  time.Duration      `json:"average_latency"`
	ErrorRate       float64            `json:"error_rate"`
	ConfidenceScore float64            `json:"confidence_score"`
	Accuracy        float64            `json:"accuracy"`
	Precision       float64            `json:"precision"`
	Recall          float64            `json:"recall"`
	F1Score         float64            `json:"f1_score"`
	LastUpdated     time.Time          `json:"last_updated"`
}

// MetricsTracker tracks metrics for A/B testing experiments
type MetricsTracker struct {
	metrics     map[string]map[string]*ModelMetrics // experiment_id -> model_id -> metrics
	predictions map[string][]*PredictionRecord      // experiment_id -> predictions
	logger      *zap.Logger
	mu          sync.RWMutex
}

// NewMetricsTracker creates a new metrics tracker
func NewMetricsTracker(logger *zap.Logger) *MetricsTracker {
	return &MetricsTracker{
		metrics:     make(map[string]map[string]*ModelMetrics),
		predictions: make(map[string][]*PredictionRecord),
		logger:      logger,
	}
}

// RecordPrediction records a prediction for metrics tracking
func (mt *MetricsTracker) RecordPrediction(experimentID, modelID string, prediction *PredictionRecord) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	// Initialize experiment metrics if not exists
	if mt.metrics[experimentID] == nil {
		mt.metrics[experimentID] = make(map[string]*ModelMetrics)
	}

	// Initialize model metrics if not exists
	if mt.metrics[experimentID][modelID] == nil {
		mt.metrics[experimentID][modelID] = &ModelMetrics{
			ModelID:      modelID,
			ExperimentID: experimentID,
			Metrics:      make(map[string]float64),
			LastUpdated:  time.Now(),
		}
	}

	// Initialize predictions slice if not exists
	if mt.predictions[experimentID] == nil {
		mt.predictions[experimentID] = make([]*PredictionRecord, 0)
	}

	// Add prediction to history
	mt.predictions[experimentID] = append(mt.predictions[experimentID], prediction)

	// Update metrics
	metrics := mt.metrics[experimentID][modelID]
	metrics.RequestCount++

	if prediction.IsError {
		metrics.ErrorCount++
	} else {
		// Update latency metrics
		if metrics.RequestCount == 1 {
			metrics.AverageLatency = prediction.Latency
		} else {
			// Calculate running average
			metrics.AverageLatency = time.Duration(
				(int64(metrics.AverageLatency)*int64(metrics.RequestCount-1) + int64(prediction.Latency)) / int64(metrics.RequestCount),
			)
		}

		// Update confidence metrics
		if metrics.RequestCount == 1 {
			metrics.ConfidenceScore = prediction.Confidence
		} else {
			metrics.ConfidenceScore = (metrics.ConfidenceScore*float64(metrics.RequestCount-1) + prediction.Confidence) / float64(metrics.RequestCount)
		}

		// Update accuracy metrics if actual outcome is available
		if prediction.ActualOutcome != nil {
			mt.updateAccuracyMetrics(metrics, prediction)
		}
	}

	// Calculate error rate
	metrics.ErrorRate = float64(metrics.ErrorCount) / float64(metrics.RequestCount)

	metrics.LastUpdated = time.Now()

	return nil
}

// GetModelMetrics retrieves metrics for a specific model in an experiment
func (mt *MetricsTracker) GetModelMetrics(experimentID, modelID string) (*ModelMetrics, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	experimentMetrics, exists := mt.metrics[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	modelMetrics, exists := experimentMetrics[modelID]
	if !exists {
		return nil, fmt.Errorf("model %s not found in experiment %s", modelID, experimentID)
	}

	// Return a copy to avoid race conditions
	return &ModelMetrics{
		ModelID:         modelMetrics.ModelID,
		ExperimentID:    modelMetrics.ExperimentID,
		RequestCount:    modelMetrics.RequestCount,
		ErrorCount:      modelMetrics.ErrorCount,
		Metrics:         copyMap(modelMetrics.Metrics),
		AverageLatency:  modelMetrics.AverageLatency,
		ErrorRate:       modelMetrics.ErrorRate,
		ConfidenceScore: modelMetrics.ConfidenceScore,
		Accuracy:        modelMetrics.Accuracy,
		Precision:       modelMetrics.Precision,
		Recall:          modelMetrics.Recall,
		F1Score:         modelMetrics.F1Score,
		LastUpdated:     modelMetrics.LastUpdated,
	}, nil
}

// GetExperimentMetrics retrieves metrics for all models in an experiment
func (mt *MetricsTracker) GetExperimentMetrics(experimentID string) (map[string]*ModelMetrics, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	experimentMetrics, exists := mt.metrics[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	// Return a copy of all model metrics
	result := make(map[string]*ModelMetrics)
	for modelID, metrics := range experimentMetrics {
		result[modelID] = &ModelMetrics{
			ModelID:         metrics.ModelID,
			ExperimentID:    metrics.ExperimentID,
			RequestCount:    metrics.RequestCount,
			ErrorCount:      metrics.ErrorCount,
			Metrics:         copyMap(metrics.Metrics),
			AverageLatency:  metrics.AverageLatency,
			ErrorRate:       metrics.ErrorRate,
			ConfidenceScore: metrics.ConfidenceScore,
			Accuracy:        metrics.Accuracy,
			Precision:       metrics.Precision,
			Recall:          metrics.Recall,
			F1Score:         metrics.F1Score,
			LastUpdated:     metrics.LastUpdated,
		}
	}

	return result, nil
}

// GetPredictions retrieves prediction history for an experiment
func (mt *MetricsTracker) GetPredictions(experimentID string, limit int) ([]*PredictionRecord, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	predictions, exists := mt.predictions[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	// Return recent predictions
	start := 0
	if limit > 0 && len(predictions) > limit {
		start = len(predictions) - limit
	}

	result := make([]*PredictionRecord, len(predictions[start:]))
	copy(result, predictions[start:])

	return result, nil
}

// GetPredictionsByModel retrieves predictions for a specific model in an experiment
func (mt *MetricsTracker) GetPredictionsByModel(experimentID, modelID string, limit int) ([]*PredictionRecord, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	predictions, exists := mt.predictions[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	var modelPredictions []*PredictionRecord
	for _, prediction := range predictions {
		if prediction.ModelID == modelID {
			modelPredictions = append(modelPredictions, prediction)
		}
	}

	// Return recent predictions
	start := 0
	if limit > 0 && len(modelPredictions) > limit {
		start = len(modelPredictions) - limit
	}

	result := make([]*PredictionRecord, len(modelPredictions[start:]))
	copy(result, modelPredictions[start:])

	return result, nil
}

// ClearExperimentData clears all data for an experiment
func (mt *MetricsTracker) ClearExperimentData(experimentID string) error {
	mt.mu.Lock()
	defer mt.mu.Unlock()

	delete(mt.metrics, experimentID)
	delete(mt.predictions, experimentID)

	mt.logger.Info("Experiment data cleared",
		zap.String("experiment_id", experimentID))

	return nil
}

// GetExperimentSummary returns a summary of experiment metrics
func (mt *MetricsTracker) GetExperimentSummary(experimentID string) (*ExperimentSummary, error) {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	experimentMetrics, exists := mt.metrics[experimentID]
	if !exists {
		return nil, fmt.Errorf("experiment %s not found", experimentID)
	}

	summary := &ExperimentSummary{
		ExperimentID:   experimentID,
		TotalRequests:  0,
		TotalErrors:    0,
		AverageLatency: 0,
		ModelCount:     len(experimentMetrics),
		StartTime:      time.Time{},
		LastUpdated:    time.Time{},
		ModelSummaries: make(map[string]*ModelSummary),
	}

	for modelID, metrics := range experimentMetrics {
		summary.TotalRequests += metrics.RequestCount
		summary.TotalErrors += metrics.ErrorCount

		// Update start time (earliest last updated)
		if summary.StartTime.IsZero() || metrics.LastUpdated.Before(summary.StartTime) {
			summary.StartTime = metrics.LastUpdated
		}

		// Update last updated (latest last updated)
		if metrics.LastUpdated.After(summary.LastUpdated) {
			summary.LastUpdated = metrics.LastUpdated
		}

		// Add model summary
		summary.ModelSummaries[modelID] = &ModelSummary{
			ModelID:         modelID,
			RequestCount:    metrics.RequestCount,
			ErrorCount:      metrics.ErrorCount,
			ErrorRate:       metrics.ErrorRate,
			AverageLatency:  metrics.AverageLatency,
			ConfidenceScore: metrics.ConfidenceScore,
			Accuracy:        metrics.Accuracy,
			F1Score:         metrics.F1Score,
		}
	}

	// Calculate overall average latency
	if summary.TotalRequests > 0 {
		totalLatency := int64(0)
		for _, metrics := range experimentMetrics {
			totalLatency += int64(metrics.AverageLatency) * int64(metrics.RequestCount)
		}
		summary.AverageLatency = time.Duration(totalLatency / int64(summary.TotalRequests))
	}

	return summary, nil
}

// updateAccuracyMetrics updates accuracy-related metrics
func (mt *MetricsTracker) updateAccuracyMetrics(metrics *ModelMetrics, prediction *PredictionRecord) {
	// This is a simplified implementation
	// In a real system, you would implement proper accuracy calculation
	// based on the prediction type and actual outcome

	// For now, we'll use a mock calculation
	// In practice, you would compare prediction with actual outcome
	// and calculate true positives, false positives, etc.

	// Mock accuracy calculation (replace with real implementation)
	if prediction.Confidence > 0.8 {
		metrics.Accuracy = 0.95 // Mock high accuracy for high confidence
	} else {
		metrics.Accuracy = 0.75 // Mock lower accuracy for low confidence
	}

	// Mock precision and recall (replace with real implementation)
	metrics.Precision = metrics.Accuracy * 0.9
	metrics.Recall = metrics.Accuracy * 0.85
	metrics.F1Score = 2 * (metrics.Precision * metrics.Recall) / (metrics.Precision + metrics.Recall)
}

// copyMap creates a deep copy of a map
func copyMap(original map[string]float64) map[string]float64 {
	copy := make(map[string]float64)
	for k, v := range original {
		copy[k] = v
	}
	return copy
}

// ExperimentSummary represents a summary of experiment metrics
type ExperimentSummary struct {
	ExperimentID   string                   `json:"experiment_id"`
	TotalRequests  int64                    `json:"total_requests"`
	TotalErrors    int64                    `json:"total_errors"`
	AverageLatency time.Duration            `json:"average_latency"`
	ModelCount     int                      `json:"model_count"`
	StartTime      time.Time                `json:"start_time"`
	LastUpdated    time.Time                `json:"last_updated"`
	ModelSummaries map[string]*ModelSummary `json:"model_summaries"`
}

// ModelSummary represents a summary of model metrics
type ModelSummary struct {
	ModelID         string        `json:"model_id"`
	RequestCount    int64         `json:"request_count"`
	ErrorCount      int64         `json:"error_count"`
	ErrorRate       float64       `json:"error_rate"`
	AverageLatency  time.Duration `json:"average_latency"`
	ConfidenceScore float64       `json:"confidence_score"`
	Accuracy        float64       `json:"accuracy"`
	F1Score         float64       `json:"f1_score"`
}
