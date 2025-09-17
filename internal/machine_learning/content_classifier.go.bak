package machine_learning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"
)

// ContentClassifier provides BERT-based content classification for business data
type ContentClassifier struct {
	models           map[string]*ClassificationModel
	modelRegistry    *ModelRegistry
	trainingPipeline *TrainingPipeline
	confidenceScorer *ConfidenceScorer
	explainability   *ModelExplainability
	config           ContentClassifierConfig
	mu               sync.RWMutex
}

// ContentClassifierConfig holds configuration for the content classifier
type ContentClassifierConfig struct {
	// Model configuration
	ModelType         string  `json:"model_type"` // bert, roberta, distilbert
	MaxSequenceLength int     `json:"max_sequence_length"`
	BatchSize         int     `json:"batch_size"`
	LearningRate      float64 `json:"learning_rate"`
	Epochs            int     `json:"epochs"`
	ValidationSplit   float64 `json:"validation_split"`

	// Industry-specific models
	IndustryModels      []string      `json:"industry_models"`
	DefaultModel        string        `json:"default_model"`
	ModelUpdateInterval time.Duration `json:"model_update_interval"`

	// Confidence and explainability
	ConfidenceThreshold    float64 `json:"confidence_threshold"`
	ExplainabilityEnabled  bool    `json:"explainability_enabled"`
	AttentionVisualization bool    `json:"attention_visualization"`

	// Performance and monitoring
	PerformanceTracking bool `json:"performance_tracking"`
	ABTestingEnabled    bool `json:"ab_testing_enabled"`
	ModelVersioning     bool `json:"model_versioning"`

	// Training and retraining
	AutoRetraining      bool    `json:"auto_retraining"`
	RetrainingThreshold float64 `json:"retraining_threshold"`
	DataDriftDetection  bool    `json:"data_drift_detection"`
}

// ClassificationModel represents a trained classification model
type ClassificationModel struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Version        string `json:"version"`
	Industry       string `json:"industry"`
	ModelType      string `json:"model_type"`
	ModelPath      string `json:"model_path"`
	ConfigPath     string `json:"config_path"`
	VocabularyPath string `json:"vocabulary_path"`

	// Performance metrics
	Accuracy  float64 `json:"accuracy"`
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	F1Score   float64 `json:"f1_score"`

	// Training information
	TrainedAt       time.Time              `json:"trained_at"`
	TrainingData    TrainingDataInfo       `json:"training_data"`
	Hyperparameters map[string]interface{} `json:"hyperparameters"`

	// Model state
	IsActive   bool      `json:"is_active"`
	IsDeployed bool      `json:"is_deployed"`
	LastUsed   time.Time `json:"last_used"`

	// Version control
	ParentVersion   string `json:"parent_version"`
	RollbackVersion string `json:"rollback_version"`
}

// TrainingDataInfo holds information about training data
type TrainingDataInfo struct {
	TotalSamples      int       `json:"total_samples"`
	TrainingSamples   int       `json:"training_samples"`
	ValidationSamples int       `json:"validation_samples"`
	TestSamples       int       `json:"test_samples"`
	LastUpdated       time.Time `json:"last_updated"`
	DataSources       []string  `json:"data_sources"`
}

// ClassificationResult represents the result of content classification
type ClassificationResult struct {
	ContentID       string                     `json:"content_id"`
	ModelID         string                     `json:"model_id"`
	ModelVersion    string                     `json:"model_version"`
	Classifications []ClassificationPrediction `json:"classifications"`
	Confidence      float64                    `json:"confidence"`
	ProcessingTime  time.Duration              `json:"processing_time"`
	Timestamp       time.Time                  `json:"timestamp"`

	// Explainability
	Explanations  []Explanation  `json:"explanations,omitempty"`
	AttentionMaps []AttentionMap `json:"attention_maps,omitempty"`

	// Quality assessment
	QualityScore   float64         `json:"quality_score"`
	QualityFactors []QualityFactor `json:"quality_factors"`
}

// ClassificationPrediction represents a single classification prediction
type ClassificationPrediction struct {
	Label       string  `json:"label"`
	Confidence  float64 `json:"confidence"`
	Probability float64 `json:"probability"`
	Rank        int     `json:"rank"`
}

// Explanation provides explainability for a classification
type Explanation struct {
	Feature      string  `json:"feature"`
	Importance   float64 `json:"importance"`
	Contribution float64 `json:"contribution"`
	Type         string  `json:"type"` // token, phrase, sentence
}

// AttentionMap represents attention weights for explainability
type AttentionMap struct {
	Layer            int         `json:"layer"`
	Head             int         `json:"head"`
	AttentionWeights [][]float64 `json:"attention_weights"`
	Tokens           []string    `json:"tokens"`
}

// QualityFactor represents a factor affecting content quality
type QualityFactor struct {
	Factor      string  `json:"factor"`
	Score       float64 `json:"score"`
	Weight      float64 `json:"weight"`
	Description string  `json:"description"`
}

// ModelRegistry manages model versions and deployments
type ModelRegistry struct {
	models          map[string]*ClassificationModel
	versions        map[string][]string
	deployments     map[string]string
	rollbackHistory map[string][]string
	mu              sync.RWMutex
}

// TrainingPipeline manages model training and retraining
type TrainingPipeline struct {
	config      TrainingConfig
	datasets    map[string]*TrainingDataset
	experiments map[string]*TrainingExperiment
	mu          sync.RWMutex
}

// TrainingConfig holds training configuration
type TrainingConfig struct {
	BaseModel      string `json:"base_model"`
	Optimizer      string `json:"optimizer"`
	LossFunction   string `json:"loss_function"`
	Regularization string `json:"regularization"`
	EarlyStopping  bool   `json:"early_stopping"`
	Patience       int    `json:"patience"`
	Checkpointing  bool   `json:"checkpointing"`
	MixedPrecision bool   `json:"mixed_precision"`
}

// TrainingDataset represents a dataset for training
type TrainingDataset struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Industry          string             `json:"industry"`
	TotalSamples      int                `json:"total_samples"`
	Labels            []string           `json:"labels"`
	LabelDistribution map[string]int     `json:"label_distribution"`
	CreatedAt         time.Time          `json:"created_at"`
	LastUpdated       time.Time          `json:"last_updated"`
	DataQuality       DataQualityMetrics `json:"data_quality"`
}

// DataQualityMetrics holds data quality information
type DataQualityMetrics struct {
	Completeness float64 `json:"completeness"`
	Consistency  float64 `json:"consistency"`
	Accuracy     float64 `json:"accuracy"`
	Timeliness   float64 `json:"timeliness"`
	Validity     float64 `json:"validity"`
	Uniqueness   float64 `json:"uniqueness"`
}

// TrainingExperiment represents a training experiment
type TrainingExperiment struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	DatasetID       string                 `json:"dataset_id"`
	Hyperparameters map[string]interface{} `json:"hyperparameters"`
	Results         TrainingResults        `json:"results"`
	Status          string                 `json:"status"` // running, completed, failed
	StartedAt       time.Time              `json:"started_at"`
	CompletedAt     time.Time              `json:"completed_at"`
}

// TrainingResults holds training experiment results
type TrainingResults struct {
	FinalAccuracy     float64            `json:"final_accuracy"`
	FinalLoss         float64            `json:"final_loss"`
	TrainingHistory   []TrainingEpoch    `json:"training_history"`
	ValidationMetrics []ValidationMetric `json:"validation_metrics"`
	TestMetrics       TestMetrics        `json:"test_metrics"`
}

// TrainingEpoch represents a training epoch
type TrainingEpoch struct {
	Epoch        int           `json:"epoch"`
	Loss         float64       `json:"loss"`
	Accuracy     float64       `json:"accuracy"`
	LearningRate float64       `json:"learning_rate"`
	TimeElapsed  time.Duration `json:"time_elapsed"`
}

// ValidationMetric represents validation metrics
type ValidationMetric struct {
	Epoch     int     `json:"epoch"`
	Loss      float64 `json:"loss"`
	Accuracy  float64 `json:"accuracy"`
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	F1Score   float64 `json:"f1_score"`
}

// TestMetrics represents test set metrics
type TestMetrics struct {
	Accuracy             float64                       `json:"accuracy"`
	Precision            float64                       `json:"precision"`
	Recall               float64                       `json:"recall"`
	F1Score              float64                       `json:"f1_score"`
	ConfusionMatrix      [][]int                       `json:"confusion_matrix"`
	ClassificationReport map[string]map[string]float64 `json:"classification_report"`
}

// ConfidenceScorer provides confidence scoring for predictions
type ConfidenceScorer struct {
	config      ConfidenceConfig
	calibration map[string]*CalibrationData
	mu          sync.RWMutex
}

// ConfidenceConfig holds confidence scoring configuration
type ConfidenceConfig struct {
	Method          string  `json:"method"` // temperature_scaling, platt_scaling, isotonic
	CalibrationData bool    `json:"calibration_data"`
	EnsembleMethod  string  `json:"ensemble_method"` // averaging, voting, stacking
	Threshold       float64 `json:"threshold"`
}

// CalibrationData holds calibration information
type CalibrationData struct {
	ModelID        string             `json:"model_id"`
	CalibrationSet []CalibrationPoint `json:"calibration_set"`
	CalibratedAt   time.Time          `json:"calibrated_at"`
	Reliability    float64            `json:"reliability"`
}

// CalibrationPoint represents a calibration data point
type CalibrationPoint struct {
	PredictedProb float64 `json:"predicted_prob"`
	ActualProb    float64 `json:"actual_prob"`
	Count         int     `json:"count"`
}

// ModelExplainability provides model explainability features
type ModelExplainability struct {
	config     ExplainabilityConfig
	explainers map[string]Explainer
	mu         sync.RWMutex
}

// ExplainabilityConfig holds explainability configuration
type ExplainabilityConfig struct {
	Methods         []string `json:"methods"` // attention, gradients, shap, lime
	AttentionLayers []int    `json:"attention_layers"`
	MaxTokens       int      `json:"max_tokens"`
	Visualization   bool     `json:"visualization"`
}

// Explainer represents an explainability method
type Explainer interface {
	Explain(content string, model *ClassificationModel) ([]Explanation, error)
}

// NewContentClassifier creates a new content classifier
func NewContentClassifier(config ContentClassifierConfig) *ContentClassifier {
	if config.ModelType == "" {
		config.ModelType = "bert"
	}

	if config.MaxSequenceLength == 0 {
		config.MaxSequenceLength = 512
	}

	if config.BatchSize == 0 {
		config.BatchSize = 16
	}

	if config.LearningRate == 0 {
		config.LearningRate = 2e-5
	}

	if config.Epochs == 0 {
		config.Epochs = 3
	}

	if config.ValidationSplit == 0 {
		config.ValidationSplit = 0.2
	}

	if config.ConfidenceThreshold == 0 {
		config.ConfidenceThreshold = 0.8
	}

	if config.ModelUpdateInterval == 0 {
		config.ModelUpdateInterval = 24 * time.Hour
	}

	if config.RetrainingThreshold == 0 {
		config.RetrainingThreshold = 0.05
	}

	return &ContentClassifier{
		models:           make(map[string]*ClassificationModel),
		modelRegistry:    NewModelRegistry(),
		trainingPipeline: NewTrainingPipeline(config),
		confidenceScorer: NewConfidenceScorer(config.ConfidenceThreshold),
		explainability:   NewModelExplainability(config.ExplainabilityEnabled),
		config:           config,
	}
}

// ClassifyContent classifies business content using the appropriate model
func (cc *ContentClassifier) ClassifyContent(ctx context.Context, content string, industry string) (*ClassificationResult, error) {
	start := time.Now()

	// Get the appropriate model for the industry
	model, err := cc.getModelForIndustry(industry)
	if err != nil {
		return nil, fmt.Errorf("failed to get model for industry %s: %v", industry, err)
	}

	// Perform classification
	predictions, err := cc.performClassification(ctx, content, model)
	if err != nil {
		return nil, fmt.Errorf("failed to perform classification: %w", err)
	}

	// Calculate confidence
	confidence := cc.confidenceScorer.CalculateConfidence(predictions)

	// Generate explanations if enabled
	var explanations []Explanation
	if cc.config.ExplainabilityEnabled {
		explanations, err = cc.explainability.GenerateExplanations(content, model, predictions)
		if err != nil {
			// Log error but don't fail the classification
			fmt.Printf("Warning: failed to generate explanations: %v\n", err)
		}
	}

	// Assess content quality
	qualityScore, qualityFactors := cc.assessContentQuality(content, predictions)

	result := &ClassificationResult{
		ContentID:       generateContentID(content),
		ModelID:         model.ID,
		ModelVersion:    model.Version,
		Classifications: predictions,
		Confidence:      confidence,
		ProcessingTime:  time.Since(start),
		Timestamp:       time.Now(),
		Explanations:    explanations,
		QualityScore:    qualityScore,
		QualityFactors:  qualityFactors,
	}

	// Update model usage
	cc.updateModelUsage(model.ID)

	return result, nil
}

// getModelForIndustry returns the appropriate model for the given industry
func (cc *ContentClassifier) getModelForIndustry(industry string) (*ClassificationModel, error) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	// Try to find industry-specific model
	if model, exists := cc.models[industry]; exists && model.IsActive && model.IsDeployed {
		return model, nil
	}

	// Fall back to default model
	if model, exists := cc.models[cc.config.DefaultModel]; exists && model.IsActive && model.IsDeployed {
		return model, nil
	}

	return nil, fmt.Errorf("no suitable model found for industry %s", industry)
}

// performClassification performs the actual classification
func (cc *ContentClassifier) performClassification(ctx context.Context, content string, model *ClassificationModel) ([]ClassificationPrediction, error) {
	// This would integrate with actual BERT model inference
	// For now, return mock predictions

	predictions := []ClassificationPrediction{
		{
			Label:       "business_registration",
			Confidence:  0.95,
			Probability: 0.95,
			Rank:        1,
		},
		{
			Label:       "financial_report",
			Confidence:  0.03,
			Probability: 0.03,
			Rank:        2,
		},
		{
			Label:       "legal_document",
			Confidence:  0.02,
			Probability: 0.02,
			Rank:        3,
		},
	}

	return predictions, nil
}

// assessContentQuality assesses the quality of the content
func (cc *ContentClassifier) assessContentQuality(content string, predictions []ClassificationPrediction) (float64, []QualityFactor) {
	var qualityFactors []QualityFactor

	// Content length factor
	lengthScore := math.Min(float64(len(content))/1000.0, 1.0)
	qualityFactors = append(qualityFactors, QualityFactor{
		Factor:      "content_length",
		Score:       lengthScore,
		Weight:      0.2,
		Description: "Content length adequacy",
	})

	// Confidence factor
	if len(predictions) > 0 {
		confidenceScore := predictions[0].Confidence
		qualityFactors = append(qualityFactors, QualityFactor{
			Factor:      "classification_confidence",
			Score:       confidenceScore,
			Weight:      0.4,
			Description: "Classification confidence level",
		})
	}

	// Content structure factor
	structureScore := cc.assessContentStructure(content)
	qualityFactors = append(qualityFactors, QualityFactor{
		Factor:      "content_structure",
		Score:       structureScore,
		Weight:      0.3,
		Description: "Content structure and formatting",
	})

	// Language quality factor
	languageScore := cc.assessLanguageQuality(content)
	qualityFactors = append(qualityFactors, QualityFactor{
		Factor:      "language_quality",
		Score:       languageScore,
		Weight:      0.1,
		Description: "Language and grammar quality",
	})

	// Calculate overall quality score
	var totalScore float64
	var totalWeight float64

	for _, factor := range qualityFactors {
		totalScore += factor.Score * factor.Weight
		totalWeight += factor.Weight
	}

	overallScore := totalScore / totalWeight

	return overallScore, qualityFactors
}

// assessContentStructure assesses the structure of the content
func (cc *ContentClassifier) assessContentStructure(content string) float64 {
	// Simple structure assessment
	// In a real implementation, this would analyze formatting, sections, etc.

	if len(content) < 100 {
		return 0.3
	} else if len(content) < 500 {
		return 0.6
	} else if len(content) < 2000 {
		return 0.8
	} else {
		return 0.9
	}
}

// assessLanguageQuality assesses the language quality of the content
func (cc *ContentClassifier) assessLanguageQuality(content string) float64 {
	// Simple language quality assessment
	// In a real implementation, this would use NLP libraries

	// Check for basic indicators of quality
	hasCapitalization := false
	hasPunctuation := false

	for _, char := range content {
		if char >= 'A' && char <= 'Z' {
			hasCapitalization = true
		}
		if char == '.' || char == ',' || char == '!' || char == '?' {
			hasPunctuation = true
		}
	}

	score := 0.5 // Base score

	if hasCapitalization {
		score += 0.2
	}

	if hasPunctuation {
		score += 0.3
	}

	return math.Min(score, 1.0)
}

// updateModelUsage updates model usage statistics
func (cc *ContentClassifier) updateModelUsage(modelID string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	if model, exists := cc.models[modelID]; exists {
		model.LastUsed = time.Now()
	}
}

// generateContentID generates a unique ID for content
func generateContentID(content string) string {
	// In a real implementation, this would use a proper hash function
	return fmt.Sprintf("content_%d", time.Now().UnixNano())
}

// NewModelRegistry creates a new model registry
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		models:          make(map[string]*ClassificationModel),
		versions:        make(map[string][]string),
		deployments:     make(map[string]string),
		rollbackHistory: make(map[string][]string),
	}
}

// NewTrainingPipeline creates a new training pipeline
func NewTrainingPipeline(config ContentClassifierConfig) *TrainingPipeline {
	return &TrainingPipeline{
		config: TrainingConfig{
			BaseModel:      "bert-base-uncased",
			Optimizer:      "adamw",
			LossFunction:   "cross_entropy",
			Regularization: "dropout",
			EarlyStopping:  true,
			Patience:       3,
			Checkpointing:  true,
			MixedPrecision: true,
		},
		datasets:    make(map[string]*TrainingDataset),
		experiments: make(map[string]*TrainingExperiment),
	}
}

// NewConfidenceScorer creates a new confidence scorer
func NewConfidenceScorer(threshold float64) *ConfidenceScorer {
	return &ConfidenceScorer{
		config: ConfidenceConfig{
			Method:          "temperature_scaling",
			CalibrationData: true,
			EnsembleMethod:  "averaging",
			Threshold:       threshold,
		},
		calibration: make(map[string]*CalibrationData),
	}
}

// NewModelExplainability creates a new model explainability system
func NewModelExplainability(enabled bool) *ModelExplainability {
	return &ModelExplainability{
		config: ExplainabilityConfig{
			Methods:         []string{"attention", "gradients"},
			AttentionLayers: []int{0, 6, 11},
			MaxTokens:       100,
			Visualization:   true,
		},
		explainers: make(map[string]Explainer),
	}
}

// CalculateConfidence calculates confidence for predictions
func (cs *ConfidenceScorer) CalculateConfidence(predictions []ClassificationPrediction) float64 {
	if len(predictions) == 0 {
		return 0.0
	}

	// Use the highest confidence prediction
	return predictions[0].Confidence
}

// GenerateExplanations generates explanations for predictions
func (me *ModelExplainability) GenerateExplanations(content string, model *ClassificationModel, predictions []ClassificationPrediction) ([]Explanation, error) {
	var explanations []Explanation

	// Generate token-level explanations
	tokenExplanations := me.generateTokenExplanations(content, predictions)
	explanations = append(explanations, tokenExplanations...)

	// Generate phrase-level explanations
	phraseExplanations := me.generatePhraseExplanations(content, predictions)
	explanations = append(explanations, phraseExplanations...)

	// Sort by importance
	sort.Slice(explanations, func(i, j int) bool {
		return explanations[i].Importance > explanations[j].Importance
	})

	return explanations, nil
}

// generateTokenExplanations generates token-level explanations
func (me *ModelExplainability) generateTokenExplanations(content string, predictions []ClassificationPrediction) []Explanation {
	var explanations []Explanation

	// Simple token importance based on frequency and position
	// In a real implementation, this would use attention weights or gradients

	tokens := []string{"business", "registration", "company", "incorporated", "llc"}

	for i, token := range tokens {
		importance := 0.8 - float64(i)*0.1
		if importance < 0.1 {
			importance = 0.1
		}

		explanations = append(explanations, Explanation{
			Feature:      token,
			Importance:   importance,
			Contribution: importance * 0.8,
			Type:         "token",
		})
	}

	return explanations
}

// generatePhraseExplanations generates phrase-level explanations
func (me *ModelExplainability) generatePhraseExplanations(content string, predictions []ClassificationPrediction) []Explanation {
	var explanations []Explanation

	// Simple phrase importance
	phrases := []string{"business registration", "company formation", "legal entity"}

	for i, phrase := range phrases {
		importance := 0.9 - float64(i)*0.2
		if importance < 0.2 {
			importance = 0.2
		}

		explanations = append(explanations, Explanation{
			Feature:      phrase,
			Importance:   importance,
			Contribution: importance * 0.9,
			Type:         "phrase",
		})
	}

	return explanations
}
