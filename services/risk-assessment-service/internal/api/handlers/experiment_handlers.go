package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"kyb-platform/services/risk-assessment-service/internal/middleware"
	"kyb-platform/services/risk-assessment-service/internal/ml/testing"
)

// ExperimentHandler handles A/B testing experiment requests
type ExperimentHandler struct {
	abTestManager     *testing.ABTestManager
	experimentManager *testing.ExperimentManager
	logger            *zap.Logger
}

// NewExperimentHandler creates a new experiment handler
func NewExperimentHandler(abTestManager *testing.ABTestManager, experimentManager *testing.ExperimentManager, logger *zap.Logger) *ExperimentHandler {
	return &ExperimentHandler{
		abTestManager:     abTestManager,
		experimentManager: experimentManager,
		logger:            logger,
	}
}

// CreateExperimentRequest represents a request to create an experiment
type CreateExperimentRequest struct {
	TemplateID      string                         `json:"template_id,omitempty"`
	Name            string                         `json:"name" validate:"required"`
	Description     string                         `json:"description"`
	TrafficSplit    map[string]float64             `json:"traffic_split" validate:"required"`
	Models          map[string]testing.ModelConfig `json:"models" validate:"required"`
	SuccessMetrics  []string                       `json:"success_metrics" validate:"required"`
	MinSampleSize   int                            `json:"min_sample_size"`
	ConfidenceLevel float64                        `json:"confidence_level"`
	Customizations  map[string]interface{}         `json:"customizations,omitempty"`
}

// CreateExperimentResponse represents the response for creating an experiment
type CreateExperimentResponse struct {
	Success    bool                      `json:"success"`
	Experiment *testing.Experiment       `json:"experiment,omitempty"`
	Error      *middleware.ErrorResponse `json:"error,omitempty"`
}

// StartExperimentRequest represents a request to start an experiment
type StartExperimentRequest struct {
	ExperimentID string `json:"experiment_id" validate:"required"`
}

// StartExperimentResponse represents the response for starting an experiment
type StartExperimentResponse struct {
	Success bool                      `json:"success"`
	Error   *middleware.ErrorResponse `json:"error,omitempty"`
}

// StopExperimentRequest represents a request to stop an experiment
type StopExperimentRequest struct {
	ExperimentID string `json:"experiment_id" validate:"required"`
}

// StopExperimentResponse represents the response for stopping an experiment
type StopExperimentResponse struct {
	Success bool                      `json:"success"`
	Error   *middleware.ErrorResponse `json:"error,omitempty"`
}

// RecordPredictionRequest represents a request to record a prediction
type RecordPredictionRequest struct {
	ExperimentID  string                 `json:"experiment_id" validate:"required"`
	ModelID       string                 `json:"model_id" validate:"required"`
	RequestID     string                 `json:"request_id" validate:"required"`
	Input         map[string]interface{} `json:"input" validate:"required"`
	Prediction    interface{}            `json:"prediction" validate:"required"`
	Confidence    float64                `json:"confidence" validate:"required"`
	Latency       time.Duration          `json:"latency" validate:"required"`
	ActualOutcome interface{}            `json:"actual_outcome,omitempty"`
	IsError       bool                   `json:"is_error"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
}

// RecordPredictionResponse represents the response for recording a prediction
type RecordPredictionResponse struct {
	Success bool                      `json:"success"`
	Error   *middleware.ErrorResponse `json:"error,omitempty"`
}

// GetExperimentResponse represents the response for getting an experiment
type GetExperimentResponse struct {
	Success    bool                      `json:"success"`
	Experiment *testing.Experiment       `json:"experiment,omitempty"`
	Error      *middleware.ErrorResponse `json:"error,omitempty"`
}

// ListExperimentsResponse represents the response for listing experiments
type ListExperimentsResponse struct {
	Success     bool                      `json:"success"`
	Experiments []*testing.Experiment     `json:"experiments,omitempty"`
	Error       *middleware.ErrorResponse `json:"error,omitempty"`
}

// GetExperimentResultsResponse represents the response for getting experiment results
type GetExperimentResultsResponse struct {
	Success bool                      `json:"success"`
	Results *testing.ExperimentResult `json:"results,omitempty"`
	Error   *middleware.ErrorResponse `json:"error,omitempty"`
}

// GetTemplatesResponse represents the response for getting experiment templates
type GetTemplatesResponse struct {
	Success   bool                          `json:"success"`
	Templates []*testing.ExperimentTemplate `json:"templates,omitempty"`
	Error     *middleware.ErrorResponse     `json:"error,omitempty"`
}

// GetMetricsResponse represents the response for getting experiment metrics
type GetMetricsResponse struct {
	Success bool                             `json:"success"`
	Metrics map[string]*testing.ModelMetrics `json:"metrics,omitempty"`
	Error   *middleware.ErrorResponse        `json:"error,omitempty"`
}

// CreateExperiment handles requests to create a new experiment
func (eh *ExperimentHandler) CreateExperiment(w http.ResponseWriter, r *http.Request) {
	var req CreateExperimentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("Failed to decode create experiment request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	eh.logger.Info("Creating experiment",
		zap.String("name", req.Name),
		zap.String("template_id", req.TemplateID))

	var experiment *testing.Experiment
	var err error

	if req.TemplateID != "" {
		// Create from template
		config, err := eh.experimentManager.CreateExperimentFromTemplate(r.Context(), req.TemplateID, req.Customizations)
		if err != nil {
			eh.logger.Error("Failed to create experiment from template", zap.Error(err))
			response := CreateExperimentResponse{
				Success: false,
				Error: &middleware.ErrorResponse{
					Error: middleware.ErrorDetail{
						Code:    "TEMPLATE_ERROR",
						Message: "Failed to create experiment from template",
						Details: err.Error(),
					},
					Timestamp: time.Now().Format(time.RFC3339),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		experiment, err = eh.abTestManager.CreateExperiment(r.Context(), config)
	} else {
		// Create custom experiment
		config := &testing.ExperimentConfig{
			Name:            req.Name,
			Description:     req.Description,
			TrafficSplit:    req.TrafficSplit,
			Models:          req.Models,
			SuccessMetrics:  req.SuccessMetrics,
			MinSampleSize:   req.MinSampleSize,
			ConfidenceLevel: req.ConfidenceLevel,
		}

		if config.MinSampleSize == 0 {
			config.MinSampleSize = 1000
		}
		if config.ConfidenceLevel == 0 {
			config.ConfidenceLevel = 0.95
		}

		experiment, err = eh.abTestManager.CreateExperiment(r.Context(), config)
	}

	if err != nil {
		eh.logger.Error("Failed to create experiment", zap.Error(err))
		response := CreateExperimentResponse{
			Success: false,
			Error: &middleware.ErrorResponse{
				Error: middleware.ErrorDetail{
					Code:    "EXPERIMENT_ERROR",
					Message: "Failed to create experiment",
					Details: err.Error(),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := CreateExperimentResponse{
		Success:    true,
		Experiment: experiment,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	eh.logger.Info("Experiment created successfully",
		zap.String("experiment_id", experiment.ID),
		zap.String("name", experiment.Name))
}

// StartExperiment handles requests to start an experiment
func (eh *ExperimentHandler) StartExperiment(w http.ResponseWriter, r *http.Request) {
	var req StartExperimentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("Failed to decode start experiment request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.ExperimentID == "" {
		http.Error(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	eh.logger.Info("Starting experiment",
		zap.String("experiment_id", req.ExperimentID))

	err := eh.abTestManager.StartExperiment(r.Context(), req.ExperimentID)
	if err != nil {
		eh.logger.Error("Failed to start experiment", zap.Error(err))
		response := StartExperimentResponse{
			Success: false,
			Error: &middleware.ErrorResponse{
				Error: middleware.ErrorDetail{
					Code:    "EXPERIMENT_ERROR",
					Message: "Failed to start experiment",
					Details: err.Error(),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := StartExperimentResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	eh.logger.Info("Experiment started successfully",
		zap.String("experiment_id", req.ExperimentID))
}

// StopExperiment handles requests to stop an experiment
func (eh *ExperimentHandler) StopExperiment(w http.ResponseWriter, r *http.Request) {
	var req StopExperimentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("Failed to decode stop experiment request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.ExperimentID == "" {
		http.Error(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	eh.logger.Info("Stopping experiment",
		zap.String("experiment_id", req.ExperimentID))

	err := eh.abTestManager.StopExperiment(r.Context(), req.ExperimentID)
	if err != nil {
		eh.logger.Error("Failed to stop experiment", zap.Error(err))
		response := StopExperimentResponse{
			Success: false,
			Error: &middleware.ErrorResponse{
				Error: middleware.ErrorDetail{
					Code:    "EXPERIMENT_ERROR",
					Message: "Failed to stop experiment",
					Details: err.Error(),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := StopExperimentResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

	eh.logger.Info("Experiment stopped successfully",
		zap.String("experiment_id", req.ExperimentID))
}

// GetExperiment handles requests to get an experiment
func (eh *ExperimentHandler) GetExperiment(w http.ResponseWriter, r *http.Request) {
	experimentID := r.URL.Query().Get("experiment_id")
	if experimentID == "" {
		http.Error(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	eh.logger.Info("Getting experiment",
		zap.String("experiment_id", experimentID))

	experiment, err := eh.abTestManager.GetExperiment(experimentID)
	if err != nil {
		eh.logger.Error("Failed to get experiment", zap.Error(err))
		response := GetExperimentResponse{
			Success: false,
			Error: &middleware.ErrorResponse{
				Error: middleware.ErrorDetail{
					Code:    "EXPERIMENT_ERROR",
					Message: "Failed to get experiment",
					Details: err.Error(),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := GetExperimentResponse{
		Success:    true,
		Experiment: experiment,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListExperiments handles requests to list all experiments
func (eh *ExperimentHandler) ListExperiments(w http.ResponseWriter, r *http.Request) {
	eh.logger.Info("Listing experiments")

	experiments := eh.abTestManager.ListExperiments()

	response := ListExperimentsResponse{
		Success:     true,
		Experiments: experiments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetExperimentResults handles requests to get experiment results
func (eh *ExperimentHandler) GetExperimentResults(w http.ResponseWriter, r *http.Request) {
	experimentID := r.URL.Query().Get("experiment_id")
	if experimentID == "" {
		http.Error(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	eh.logger.Info("Getting experiment results",
		zap.String("experiment_id", experimentID))

	results, err := eh.abTestManager.GetExperimentResults(r.Context(), experimentID)
	if err != nil {
		eh.logger.Error("Failed to get experiment results", zap.Error(err))
		response := GetExperimentResultsResponse{
			Success: false,
			Error: &middleware.ErrorResponse{
				Error: middleware.ErrorDetail{
					Code:    "EXPERIMENT_ERROR",
					Message: "Failed to get experiment results",
					Details: err.Error(),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := GetExperimentResultsResponse{
		Success: true,
		Results: results,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTemplates handles requests to get experiment templates
func (eh *ExperimentHandler) GetTemplates(w http.ResponseWriter, r *http.Request) {
	experimentType := r.URL.Query().Get("type")

	eh.logger.Info("Getting experiment templates",
		zap.String("type", experimentType))

	var templates []*testing.ExperimentTemplate
	if experimentType != "" {
		templates = eh.experimentManager.GetTemplatesByType(testing.ExperimentType(experimentType))
	} else {
		templates = eh.experimentManager.ListTemplates()
	}

	response := GetTemplatesResponse{
		Success:   true,
		Templates: templates,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RecordPrediction handles requests to record a prediction
func (eh *ExperimentHandler) RecordPrediction(w http.ResponseWriter, r *http.Request) {
	var req RecordPredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		eh.logger.Error("Failed to decode record prediction request", zap.Error(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.ExperimentID == "" {
		http.Error(w, "experiment_id is required", http.StatusBadRequest)
		return
	}
	if req.ModelID == "" {
		http.Error(w, "model_id is required", http.StatusBadRequest)
		return
	}
	if req.RequestID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}

	eh.logger.Info("Recording prediction",
		zap.String("experiment_id", req.ExperimentID),
		zap.String("model_id", req.ModelID),
		zap.String("request_id", req.RequestID))

	prediction := &testing.PredictionRecord{
		RequestID:     req.RequestID,
		ModelID:       req.ModelID,
		ExperimentID:  req.ExperimentID,
		Input:         req.Input,
		Prediction:    req.Prediction,
		Confidence:    req.Confidence,
		Latency:       req.Latency,
		Timestamp:     time.Now(),
		ActualOutcome: req.ActualOutcome,
		IsError:       req.IsError,
		ErrorMessage:  req.ErrorMessage,
	}

	err := eh.abTestManager.RecordPrediction(r.Context(), req.ExperimentID, req.ModelID, prediction)
	if err != nil {
		eh.logger.Error("Failed to record prediction", zap.Error(err))
		response := RecordPredictionResponse{
			Success: false,
			Error: &middleware.ErrorResponse{
				Error: middleware.ErrorDetail{
					Code:    "PREDICTION_ERROR",
					Message: "Failed to record prediction",
					Details: err.Error(),
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := RecordPredictionResponse{
		Success: true,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetMetrics handles requests to get experiment metrics
func (eh *ExperimentHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	experimentID := r.URL.Query().Get("experiment_id")
	if experimentID == "" {
		http.Error(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	modelID := r.URL.Query().Get("model_id")

	eh.logger.Info("Getting experiment metrics",
		zap.String("experiment_id", experimentID),
		zap.String("model_id", modelID))

	var metrics map[string]*testing.ModelMetrics
	var err error

	if modelID != "" {
		// Get metrics for specific model
		modelMetrics, err := eh.abTestManager.GetModelMetrics(experimentID, modelID)
		if err != nil {
			eh.logger.Error("Failed to get model metrics", zap.Error(err))
			response := GetMetricsResponse{
				Success: false,
				Error: &middleware.ErrorResponse{
					Error: middleware.ErrorDetail{
						Code:    "METRICS_ERROR",
						Message: "Failed to get model metrics",
						Details: err.Error(),
					},
					Timestamp: time.Now().Format(time.RFC3339),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		metrics = map[string]*testing.ModelMetrics{
			modelID: modelMetrics,
		}
	} else {
		// Get metrics for all models in experiment
		metrics, err = eh.abTestManager.GetExperimentMetrics(experimentID)
		if err != nil {
			eh.logger.Error("Failed to get experiment metrics", zap.Error(err))
			response := GetMetricsResponse{
				Success: false,
				Error: &middleware.ErrorResponse{
					Error: middleware.ErrorDetail{
						Code:    "METRICS_ERROR",
						Message: "Failed to get experiment metrics",
						Details: err.Error(),
					},
					Timestamp: time.Now().Format(time.RFC3339),
				},
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	response := GetMetricsResponse{
		Success: true,
		Metrics: metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// SelectModel handles requests to select a model for a request
func (eh *ExperimentHandler) SelectModel(w http.ResponseWriter, r *http.Request) {
	experimentID := r.URL.Query().Get("experiment_id")
	requestID := r.URL.Query().Get("request_id")

	if experimentID == "" {
		http.Error(w, "experiment_id is required", http.StatusBadRequest)
		return
	}
	if requestID == "" {
		http.Error(w, "request_id is required", http.StatusBadRequest)
		return
	}

	eh.logger.Info("Selecting model for request",
		zap.String("experiment_id", experimentID),
		zap.String("request_id", requestID))

	modelID, err := eh.abTestManager.SelectModel(r.Context(), experimentID, requestID)
	if err != nil {
		eh.logger.Error("Failed to select model", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":       true,
		"experiment_id": experimentID,
		"request_id":    requestID,
		"model_id":      modelID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// CreateModelComparisonExperiment creates a model comparison experiment
func (eh *ExperimentHandler) CreateModelComparisonExperiment(w http.ResponseWriter, r *http.Request) {
	eh.CreateExperiment(w, r)
}

// CreateHyperparameterExperiment creates a hyperparameter tuning experiment
func (eh *ExperimentHandler) CreateHyperparameterExperiment(w http.ResponseWriter, r *http.Request) {
	eh.CreateExperiment(w, r)
}

// CreateFeatureExperiment creates a feature experiment
func (eh *ExperimentHandler) CreateFeatureExperiment(w http.ResponseWriter, r *http.Request) {
	eh.CreateExperiment(w, r)
}

// CreateIndustryExperiment creates an industry-specific experiment
func (eh *ExperimentHandler) CreateIndustryExperiment(w http.ResponseWriter, r *http.Request) {
	eh.CreateExperiment(w, r)
}

// GetExperimentStatus gets the status of an experiment
func (eh *ExperimentHandler) GetExperimentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	experimentID := vars["id"]

	if experimentID == "" {
		http.Error(w, "experiment_id is required", http.StatusBadRequest)
		return
	}

	// Use abTestManager to get experiment (same as GetExperiment method)
	if eh.abTestManager == nil {
		http.Error(w, "AB test manager not available", http.StatusInternalServerError)
		return
	}

	experiment, err := eh.abTestManager.GetExperiment(experimentID)
	if err != nil {
		eh.logger.Error("Failed to get experiment status", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success":       true,
		"experiment_id": experimentID,
		"status":        experiment.Status,
		"created_at":    experiment.CreatedAt,
		"updated_at":    experiment.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
