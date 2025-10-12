package testing

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// ExperimentConfig represents configuration for creating an experiment
type ExperimentConfig struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	TrafficSplit    map[string]float64     `json:"traffic_split"`
	Models          map[string]ModelConfig `json:"models"`
	SuccessMetrics  []string               `json:"success_metrics"`
	MinSampleSize   int                    `json:"min_sample_size"`
	ConfidenceLevel float64                `json:"confidence_level"`
}

// ExperimentTemplate represents a pre-configured experiment template
type ExperimentTemplate struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Type        ExperimentType   `json:"type"`
	Config      ExperimentConfig `json:"config"`
	CreatedAt   time.Time        `json:"created_at"`
}

// ExperimentManager manages experiment templates and configurations
type ExperimentManager struct {
	templates map[string]*ExperimentTemplate
	logger    *zap.Logger
}

// NewExperimentManager creates a new experiment manager
func NewExperimentManager(logger *zap.Logger) *ExperimentManager {
	manager := &ExperimentManager{
		templates: make(map[string]*ExperimentTemplate),
		logger:    logger,
	}

	// Initialize with default templates
	manager.initializeDefaultTemplates()

	return manager
}

// CreateExperimentFromTemplate creates an experiment from a template
func (em *ExperimentManager) CreateExperimentFromTemplate(ctx context.Context, templateID string, customizations map[string]interface{}) (*ExperimentConfig, error) {
	template, exists := em.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template %s not found", templateID)
	}

	// Create config from template
	config := template.Config

	// Apply customizations
	if customizations != nil {
		if name, ok := customizations["name"].(string); ok {
			config.Name = name
		}
		if description, ok := customizations["description"].(string); ok {
			config.Description = description
		}
		if trafficSplit, ok := customizations["traffic_split"].(map[string]float64); ok {
			config.TrafficSplit = trafficSplit
		}
		if minSampleSize, ok := customizations["min_sample_size"].(int); ok {
			config.MinSampleSize = minSampleSize
		}
		if confidenceLevel, ok := customizations["confidence_level"].(float64); ok {
			config.ConfidenceLevel = confidenceLevel
		}
	}

	// Generate unique ID
	config.ID = fmt.Sprintf("%s_%d", templateID, time.Now().Unix())

	em.logger.Info("Experiment created from template",
		zap.String("template_id", templateID),
		zap.String("experiment_id", config.ID),
		zap.String("name", config.Name))

	return &config, nil
}

// GetTemplate retrieves an experiment template
func (em *ExperimentManager) GetTemplate(templateID string) (*ExperimentTemplate, error) {
	template, exists := em.templates[templateID]
	if !exists {
		return nil, fmt.Errorf("template %s not found", templateID)
	}

	return template, nil
}

// ListTemplates returns all available experiment templates
func (em *ExperimentManager) ListTemplates() []*ExperimentTemplate {
	templates := make([]*ExperimentTemplate, 0, len(em.templates))
	for _, template := range em.templates {
		templates = append(templates, template)
	}

	return templates
}

// CreateCustomExperiment creates a custom experiment configuration
func (em *ExperimentManager) CreateCustomExperiment(ctx context.Context, config *ExperimentConfig) (*ExperimentConfig, error) {
	// Validate configuration
	if err := em.validateExperimentConfig(config); err != nil {
		return nil, fmt.Errorf("invalid experiment configuration: %w", err)
	}

	// Generate unique ID if not provided
	if config.ID == "" {
		config.ID = fmt.Sprintf("custom_%d", time.Now().Unix())
	}

	em.logger.Info("Custom experiment created",
		zap.String("experiment_id", config.ID),
		zap.String("name", config.Name),
		zap.Int("model_count", len(config.Models)))

	return config, nil
}

// validateExperimentConfig validates an experiment configuration
func (em *ExperimentManager) validateExperimentConfig(config *ExperimentConfig) error {
	// Validate required fields
	if config.Name == "" {
		return fmt.Errorf("experiment name is required")
	}

	if len(config.Models) < 2 {
		return fmt.Errorf("at least 2 models are required for A/B testing")
	}

	if len(config.TrafficSplit) == 0 {
		return fmt.Errorf("traffic split is required")
	}

	// Validate traffic split
	totalSplit := 0.0
	for modelID, split := range config.TrafficSplit {
		if split < 0 || split > 1 {
			return fmt.Errorf("traffic split for model %s must be between 0 and 1", modelID)
		}
		totalSplit += split
	}

	if totalSplit < 0.99 || totalSplit > 1.01 {
		return fmt.Errorf("traffic split must sum to 1.0, got %.2f", totalSplit)
	}

	// Validate that all models in traffic split exist
	for modelID := range config.TrafficSplit {
		if _, exists := config.Models[modelID]; !exists {
			return fmt.Errorf("model %s in traffic split not found in models", modelID)
		}
	}

	// Validate success metrics
	if len(config.SuccessMetrics) == 0 {
		return fmt.Errorf("at least one success metric is required")
	}

	// Validate minimum sample size
	if config.MinSampleSize < 100 {
		return fmt.Errorf("minimum sample size must be at least 100")
	}

	// Validate confidence level
	if config.ConfidenceLevel < 0.8 || config.ConfidenceLevel > 0.99 {
		return fmt.Errorf("confidence level must be between 0.8 and 0.99")
	}

	return nil
}

// initializeDefaultTemplates initializes default experiment templates
func (em *ExperimentManager) initializeDefaultTemplates() {
	// Model Comparison Template
	modelComparisonTemplate := &ExperimentTemplate{
		ID:          "model_comparison",
		Name:        "Model Comparison",
		Description: "Compare performance of different ML models",
		Type:        ExperimentTypeModelComparison,
		Config: ExperimentConfig{
			Name:        "Model Comparison Experiment",
			Description: "Compare XGBoost vs LSTM vs Ensemble models",
			TrafficSplit: map[string]float64{
				"xgboost":  0.33,
				"lstm":     0.33,
				"ensemble": 0.34,
			},
			Models: map[string]ModelConfig{
				"xgboost": {
					ID:          "xgboost",
					Name:        "XGBoost Model",
					Type:        "xgboost",
					Version:     "1.0",
					Description: "Gradient boosting model for risk assessment",
				},
				"lstm": {
					ID:          "lstm",
					Name:        "LSTM Model",
					Type:        "lstm",
					Version:     "1.0",
					Description: "Long short-term memory model for time series prediction",
				},
				"ensemble": {
					ID:          "ensemble",
					Name:        "Ensemble Model",
					Type:        "ensemble",
					Version:     "1.0",
					Description: "Combined model using multiple algorithms",
				},
			},
			SuccessMetrics:  []string{"accuracy", "precision", "recall", "f1_score"},
			MinSampleSize:   1000,
			ConfidenceLevel: 0.95,
		},
		CreatedAt: time.Now(),
	}

	// Hyperparameter Tuning Template
	hyperparameterTemplate := &ExperimentTemplate{
		ID:          "hyperparameter_tuning",
		Name:        "Hyperparameter Tuning",
		Description: "Test different hyperparameter configurations",
		Type:        ExperimentTypeHyperparameterTuning,
		Config: ExperimentConfig{
			Name:        "Hyperparameter Tuning Experiment",
			Description: "Compare different hyperparameter settings for XGBoost",
			TrafficSplit: map[string]float64{
				"default": 0.5,
				"tuned":   0.5,
			},
			Models: map[string]ModelConfig{
				"default": {
					ID:          "default",
					Name:        "Default Parameters",
					Type:        "xgboost",
					Version:     "1.0",
					Description: "XGBoost with default hyperparameters",
					Parameters: map[string]interface{}{
						"n_estimators":  100,
						"max_depth":     6,
						"learning_rate": 0.1,
					},
				},
				"tuned": {
					ID:          "tuned",
					Name:        "Tuned Parameters",
					Type:        "xgboost",
					Version:     "1.1",
					Description: "XGBoost with optimized hyperparameters",
					Parameters: map[string]interface{}{
						"n_estimators":  200,
						"max_depth":     8,
						"learning_rate": 0.05,
					},
				},
			},
			SuccessMetrics:  []string{"accuracy", "f1_score"},
			MinSampleSize:   500,
			ConfidenceLevel: 0.95,
		},
		CreatedAt: time.Now(),
	}

	// Feature Testing Template
	featureTestingTemplate := &ExperimentTemplate{
		ID:          "feature_testing",
		Name:        "Feature Testing",
		Description: "Test impact of different feature sets",
		Type:        ExperimentTypeFeatureTesting,
		Config: ExperimentConfig{
			Name:        "Feature Testing Experiment",
			Description: "Compare models with different feature sets",
			TrafficSplit: map[string]float64{
				"basic":    0.5,
				"enhanced": 0.5,
			},
			Models: map[string]ModelConfig{
				"basic": {
					ID:          "basic",
					Name:        "Basic Features",
					Type:        "xgboost",
					Version:     "1.0",
					Description: "Model with basic feature set",
					Parameters: map[string]interface{}{
						"feature_set": "basic",
						"features":    []string{"revenue", "employees", "age"},
					},
				},
				"enhanced": {
					ID:          "enhanced",
					Name:        "Enhanced Features",
					Type:        "xgboost",
					Version:     "1.1",
					Description: "Model with enhanced feature set",
					Parameters: map[string]interface{}{
						"feature_set": "enhanced",
						"features":    []string{"revenue", "employees", "age", "industry", "location", "financial_ratios"},
					},
				},
			},
			SuccessMetrics:  []string{"accuracy", "precision"},
			MinSampleSize:   750,
			ConfidenceLevel: 0.95,
		},
		CreatedAt: time.Now(),
	}

	// Industry-Specific Template
	industrySpecificTemplate := &ExperimentTemplate{
		ID:          "industry_specific",
		Name:        "Industry-Specific Testing",
		Description: "Test industry-specific model variants",
		Type:        ExperimentTypeIndustrySpecific,
		Config: ExperimentConfig{
			Name:        "Industry-Specific Experiment",
			Description: "Compare general vs industry-specific models",
			TrafficSplit: map[string]float64{
				"general":  0.5,
				"specific": 0.5,
			},
			Models: map[string]ModelConfig{
				"general": {
					ID:          "general",
					Name:        "General Model",
					Type:        "xgboost",
					Version:     "1.0",
					Description: "General-purpose risk assessment model",
					Parameters: map[string]interface{}{
						"model_type": "general",
						"industry":   "all",
					},
				},
				"specific": {
					ID:          "specific",
					Name:        "Industry-Specific Model",
					Type:        "xgboost",
					Version:     "1.1",
					Description: "Industry-specific risk assessment model",
					Parameters: map[string]interface{}{
						"model_type": "industry_specific",
						"industry":   "fintech",
					},
				},
			},
			SuccessMetrics:  []string{"accuracy", "f1_score"},
			MinSampleSize:   1000,
			ConfidenceLevel: 0.95,
		},
		CreatedAt: time.Now(),
	}

	// Add templates to manager
	em.templates[modelComparisonTemplate.ID] = modelComparisonTemplate
	em.templates[hyperparameterTemplate.ID] = hyperparameterTemplate
	em.templates[featureTestingTemplate.ID] = featureTestingTemplate
	em.templates[industrySpecificTemplate.ID] = industrySpecificTemplate

	em.logger.Info("Default experiment templates initialized",
		zap.Int("template_count", len(em.templates)))
}

// GetTemplatesByType returns templates filtered by type
func (em *ExperimentManager) GetTemplatesByType(experimentType ExperimentType) []*ExperimentTemplate {
	var templates []*ExperimentTemplate
	for _, template := range em.templates {
		if template.Type == experimentType {
			templates = append(templates, template)
		}
	}
	return templates
}

// CreateTemplate creates a new experiment template
func (em *ExperimentManager) CreateTemplate(ctx context.Context, template *ExperimentTemplate) error {
	// Validate template
	if err := em.validateExperimentConfig(&template.Config); err != nil {
		return fmt.Errorf("invalid template configuration: %w", err)
	}

	// Check if template ID already exists
	if _, exists := em.templates[template.ID]; exists {
		return fmt.Errorf("template with ID %s already exists", template.ID)
	}

	template.CreatedAt = time.Now()
	em.templates[template.ID] = template

	em.logger.Info("New experiment template created",
		zap.String("template_id", template.ID),
		zap.String("name", template.Name),
		zap.String("type", string(template.Type)))

	return nil
}

// DeleteTemplate deletes an experiment template
func (em *ExperimentManager) DeleteTemplate(templateID string) error {
	if _, exists := em.templates[templateID]; !exists {
		return fmt.Errorf("template %s not found", templateID)
	}

	delete(em.templates, templateID)

	em.logger.Info("Experiment template deleted",
		zap.String("template_id", templateID))

	return nil
}
