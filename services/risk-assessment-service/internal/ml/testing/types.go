package testing

// ModelConfig represents configuration for a model in an experiment
type ModelConfig struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"` // "xgboost", "lstm", "ensemble"
	Version     string                 `json:"version"`
	Parameters  map[string]interface{} `json:"parameters"`
	Description string                 `json:"description"`
}

// ExperimentStatus represents the status of an experiment
type ExperimentStatus string

const (
	StatusDraft     ExperimentStatus = "draft"
	StatusRunning   ExperimentStatus = "running"
	StatusPaused    ExperimentStatus = "paused"
	StatusCompleted ExperimentStatus = "completed"
	StatusCancelled ExperimentStatus = "cancelled"
)

// TrafficSplit represents how traffic is split between models
type TrafficSplit struct {
	ModelID string  `json:"model_id"`
	Weight  float64 `json:"weight"` // 0.0 to 1.0
}

// ExperimentType represents the type of experiment
type ExperimentType string

const (
	ExperimentTypeModelComparison      ExperimentType = "model_comparison"
	ExperimentTypeHyperparameterTuning ExperimentType = "hyperparameter_tuning"
	ExperimentTypeFeatureTesting       ExperimentType = "feature_testing"
	ExperimentTypeIndustrySpecific     ExperimentType = "industry_specific"
)

// MetricType represents the type of metric being tracked
type MetricType string

const (
	MetricTypeAccuracy   MetricType = "accuracy"
	MetricTypePrecision  MetricType = "precision"
	MetricTypeRecall     MetricType = "recall"
	MetricTypeF1Score    MetricType = "f1_score"
	MetricTypeAUC        MetricType = "auc"
	MetricTypeLatency    MetricType = "latency"
	MetricTypeThroughput MetricType = "throughput"
	MetricTypeErrorRate  MetricType = "error_rate"
)

// StatisticalSignificance represents the result of statistical significance testing
type StatisticalSignificance struct {
	IsSignificant   bool    `json:"is_significant"`
	PValue          float64 `json:"p_value"`
	ConfidenceLevel float64 `json:"confidence_level"`
	EffectSize      float64 `json:"effect_size"`
	SampleSize      int     `json:"sample_size"`
}
