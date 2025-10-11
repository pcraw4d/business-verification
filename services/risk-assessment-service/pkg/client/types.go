package client

import (
	"time"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// RequestOptions provides additional options for API requests
type RequestOptions struct {
	Headers map[string]string
	Timeout time.Duration
}

// RiskPredictionRequest represents a risk prediction request
type RiskPredictionRequest struct {
	HorizonMonths           int      `json:"horizon_months"`
	ModelType               string   `json:"model_type,omitempty"`
	IncludeTemporalAnalysis bool     `json:"include_temporal_analysis,omitempty"`
	Scenarios               []string `json:"scenarios,omitempty"`
}

// AdvancedPredictionRequest represents an advanced multi-horizon prediction request
type AdvancedPredictionRequest struct {
	Business                *models.RiskAssessmentRequest `json:"business"`
	PredictionHorizons      []int                         `json:"prediction_horizons"`
	ModelPreference         string                        `json:"model_preference,omitempty"`
	IncludeTemporalAnalysis bool                          `json:"include_temporal_analysis,omitempty"`
	IncludeScenarioAnalysis bool                          `json:"include_scenario_analysis,omitempty"`
	IncludeModelComparison  bool                          `json:"include_model_comparison,omitempty"`
	ConfidenceThreshold     float64                       `json:"confidence_threshold,omitempty"`
	CustomScenarios         []string                      `json:"custom_scenarios,omitempty"`
	Metadata                map[string]interface{}        `json:"metadata,omitempty"`
}

// AdvancedPredictionResponse represents an advanced prediction response
type AdvancedPredictionResponse struct {
	RequestID          string                       `json:"request_id"`
	BusinessID         string                       `json:"business_id"`
	Predictions        map[string]HorizonPrediction `json:"predictions"`
	ModelComparison    *ModelComparison             `json:"model_comparison,omitempty"`
	TemporalAnalysis   interface{}                  `json:"temporal_analysis,omitempty"`
	ScenarioAnalysis   []AdvancedScenarioAnalysis   `json:"scenario_analysis,omitempty"`
	ConfidenceAnalysis *ConfidenceAnalysis          `json:"confidence_analysis,omitempty"`
	ProcessingTime     time.Duration                `json:"processing_time"`
	GeneratedAt        time.Time                    `json:"generated_at"`
	Metadata           map[string]interface{}       `json:"metadata,omitempty"`
}

// HorizonPrediction represents a prediction for a specific horizon
type HorizonPrediction struct {
	HorizonMonths    int                       `json:"horizon_months"`
	ModelType        string                    `json:"model_type"`
	PredictedScore   float64                   `json:"predicted_score"`
	PredictedLevel   models.RiskLevel          `json:"predicted_level"`
	ConfidenceScore  float64                   `json:"confidence_score"`
	RiskFactors      []models.RiskFactor       `json:"risk_factors"`
	ScenarioAnalysis []models.ScenarioAnalysis `json:"scenario_analysis"`
	PredictionDate   time.Time                 `json:"prediction_date"`
	ModelInfo        interface{}               `json:"model_info,omitempty"`
	Metadata         map[string]interface{}    `json:"metadata,omitempty"`
}

// ModelComparison represents model comparison analysis
type ModelComparison struct {
	Horizons            map[string]ModelComparisonHorizon `json:"horizons"`
	BestModelPerHorizon map[string]string                 `json:"best_model_per_horizon"`
	AgreementAnalysis   *AgreementAnalysis                `json:"agreement_analysis,omitempty"`
	PerformanceMetrics  *PerformanceMetrics               `json:"performance_metrics,omitempty"`
}

// ModelComparisonHorizon represents model comparison for a specific horizon
type ModelComparisonHorizon struct {
	HorizonMonths        int                `json:"horizon_months"`
	XGBoostPrediction    *HorizonPrediction `json:"xgboost_prediction,omitempty"`
	LSTMPrediction       *HorizonPrediction `json:"lstm_prediction,omitempty"`
	EnsemblePrediction   *HorizonPrediction `json:"ensemble_prediction,omitempty"`
	BestModel            string             `json:"best_model"`
	ConfidenceDifference float64            `json:"confidence_difference"`
	ScoreDifference      float64            `json:"score_difference"`
}

// AgreementAnalysis represents model agreement analysis
type AgreementAnalysis struct {
	OverallAgreement         float64            `json:"overall_agreement"`
	AgreementByHorizon       map[string]float64 `json:"agreement_by_horizon"`
	DisagreementThreshold    float64            `json:"disagreement_threshold"`
	HighDisagreementHorizons []int              `json:"high_disagreement_horizons"`
}

// PerformanceMetrics represents performance metrics
type PerformanceMetrics struct {
	AverageLatency time.Duration            `json:"average_latency"`
	LatencyByModel map[string]time.Duration `json:"latency_by_model"`
	MemoryUsage    map[string]int64         `json:"memory_usage"`
	Throughput     float64                  `json:"throughput"`
}

// AdvancedScenarioAnalysis represents advanced scenario analysis
type AdvancedScenarioAnalysis struct {
	ScenarioName         string           `json:"scenario_name"`
	Description          string           `json:"description"`
	Probability          float64          `json:"probability"`
	RiskScore            float64          `json:"risk_score"`
	RiskLevel            models.RiskLevel `json:"risk_level"`
	Impact               string           `json:"impact"`
	TimeHorizon          int              `json:"time_horizon"`
	KeyFactors           []string         `json:"key_factors"`
	MitigationStrategies []string         `json:"mitigation_strategies,omitempty"`
}

// ConfidenceAnalysis represents confidence analysis
type ConfidenceAnalysis struct {
	OverallConfidence         float64            `json:"overall_confidence"`
	ConfidenceByHorizon       map[string]float64 `json:"confidence_by_horizon"`
	ConfidenceByModel         map[string]float64 `json:"confidence_by_model"`
	LowConfidencePredictions  []int              `json:"low_confidence_predictions"`
	HighConfidencePredictions []int              `json:"high_confidence_predictions"`
	CalibrationScore          float64            `json:"calibration_score"`
}

// ModelInfo represents model information
type ModelInfo struct {
	Name            string                 `json:"name"`
	Version         string                 `json:"version"`
	Type            string                 `json:"type"`
	TrainingDate    time.Time              `json:"training_date"`
	Accuracy        float64                `json:"accuracy"`
	Precision       float64                `json:"precision"`
	Recall          float64                `json:"recall"`
	F1Score         float64                `json:"f1_score"`
	Features        []string               `json:"features"`
	Hyperparameters map[string]interface{} `json:"hyperparameters"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ModelPerformanceResponse represents model performance metrics
type ModelPerformanceResponse struct {
	Models         map[string]ModelPerformance `json:"models"`
	OverallMetrics OverallPerformanceMetrics   `json:"overall_metrics"`
	LastUpdated    time.Time                   `json:"last_updated"`
}

// ModelPerformance represents performance metrics for a specific model
type ModelPerformance struct {
	ModelType   string        `json:"model_type"`
	LatencyP50  time.Duration `json:"latency_p50"`
	LatencyP95  time.Duration `json:"latency_p95"`
	LatencyP99  time.Duration `json:"latency_p99"`
	Throughput  float64       `json:"throughput"`
	MemoryUsage int64         `json:"memory_usage"`
	Accuracy    float64       `json:"accuracy"`
	ErrorRate   float64       `json:"error_rate"`
	LastUpdated time.Time     `json:"last_updated"`
}

// OverallPerformanceMetrics represents overall performance metrics
type OverallPerformanceMetrics struct {
	TotalRequests     int64         `json:"total_requests"`
	AverageLatency    time.Duration `json:"average_latency"`
	AverageThroughput float64       `json:"average_throughput"`
	TotalMemoryUsage  int64         `json:"total_memory_usage"`
	OverallAccuracy   float64       `json:"overall_accuracy"`
	OverallErrorRate  float64       `json:"overall_error_rate"`
}

// RiskHistoryResponse represents a risk history response
type RiskHistoryResponse struct {
	BusinessID  string                          `json:"business_id"`
	Assessments []models.RiskAssessmentResponse `json:"assessments"`
	Trends      RiskTrends                      `json:"trends"`
}

// RiskTrends represents risk trends
type RiskTrends struct {
	ScoreTrend      string `json:"score_trend"`
	LevelTrend      string `json:"level_trend"`
	ConfidenceTrend string `json:"confidence_trend"`
}

// ComplianceCheckRequest represents a compliance check request
type ComplianceCheckRequest struct {
	BusinessName    string   `json:"business_name"`
	BusinessAddress string   `json:"business_address"`
	Industry        string   `json:"industry"`
	Country         string   `json:"country"`
	ComplianceTypes []string `json:"compliance_types"`
}

// ComplianceCheckResponse represents a compliance check response
type ComplianceCheckResponse struct {
	BusinessID       string            `json:"business_id"`
	ComplianceStatus string            `json:"compliance_status"`
	Checks           []ComplianceCheck `json:"checks"`
	CreatedAt        time.Time         `json:"created_at"`
}

// ComplianceCheck represents a compliance check result
type ComplianceCheck struct {
	Type    string  `json:"type"`
	Status  string  `json:"status"`
	Score   float64 `json:"score"`
	Details string  `json:"details"`
}

// SanctionsScreeningRequest represents a sanctions screening request
type SanctionsScreeningRequest struct {
	BusinessName    string `json:"business_name"`
	BusinessAddress string `json:"business_address"`
	Country         string `json:"country"`
}

// SanctionsScreeningResponse represents a sanctions screening response
type SanctionsScreeningResponse struct {
	BusinessID      string           `json:"business_id"`
	SanctionsStatus string           `json:"sanctions_status"`
	Matches         []SanctionsMatch `json:"matches"`
	ScreeningDate   time.Time        `json:"screening_date"`
}

// SanctionsMatch represents a sanctions match
type SanctionsMatch struct {
	Source    string    `json:"source"`
	Score     float64   `json:"score"`
	Details   string    `json:"details"`
	MatchDate time.Time `json:"match_date"`
}

// MediaMonitoringRequest represents a media monitoring request
type MediaMonitoringRequest struct {
	BusinessName    string   `json:"business_name"`
	BusinessAddress string   `json:"business_address"`
	MonitoringTypes []string `json:"monitoring_types"`
}

// MediaMonitoringResponse represents a media monitoring response
type MediaMonitoringResponse struct {
	BusinessID   string       `json:"business_id"`
	MonitoringID string       `json:"monitoring_id"`
	Status       string       `json:"status"`
	Alerts       []MediaAlert `json:"alerts"`
	CreatedAt    time.Time    `json:"created_at"`
}

// MediaAlert represents a media alert
type MediaAlert struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	Severity    string    `json:"severity"`
	CreatedAt   time.Time `json:"created_at"`
}

// RiskTrendsOptions provides options for risk trends requests
type RiskTrendsOptions struct {
	Industry  string
	Country   string
	Timeframe string
	Limit     int
}

// RiskTrendsResponse represents a risk trends response
type RiskTrendsResponse struct {
	Trends  []RiskTrend  `json:"trends"`
	Summary TrendSummary `json:"summary"`
}

// RiskTrend represents a risk trend
type RiskTrend struct {
	Industry         string  `json:"industry"`
	Country          string  `json:"country"`
	AverageRiskScore float64 `json:"average_risk_score"`
	TrendDirection   string  `json:"trend_direction"`
	ChangePercentage float64 `json:"change_percentage"`
	SampleSize       int     `json:"sample_size"`
}

// TrendSummary represents a trend summary
type TrendSummary struct {
	TotalAssessments   int     `json:"total_assessments"`
	AverageRiskScore   float64 `json:"average_risk_score"`
	HighRiskPercentage float64 `json:"high_risk_percentage"`
}

// RiskInsightsOptions provides options for risk insights requests
type RiskInsightsOptions struct {
	Industry  string
	Country   string
	RiskLevel string
}

// RiskInsightsResponse represents a risk insights response
type RiskInsightsResponse struct {
	Insights        []RiskInsight    `json:"insights"`
	Recommendations []Recommendation `json:"recommendations"`
}

// RiskInsight represents a risk insight
type RiskInsight struct {
	Type           string `json:"type"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	Impact         string `json:"impact"`
	Recommendation string `json:"recommendation"`
}

// Recommendation represents a recommendation
type Recommendation struct {
	Category string `json:"category"`
	Action   string `json:"action"`
	Priority string `json:"priority"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id,omitempty"`
	Timestamp string      `json:"timestamp"`
	Path      string      `json:"path"`
	Method    string      `json:"method"`
}

// ErrorDetail represents error details
type ErrorDetail struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Details    string            `json:"details,omitempty"`
	Field      string            `json:"field,omitempty"`
	Validation []ValidationError `json:"validation,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

// APIError represents an API error
type APIError struct {
	StatusCode int
	Message    string
	Response   *ErrorResponse
}

// Error implements the error interface
func (e *APIError) Error() string {
	if e.Response != nil {
		return e.Response.Error.Message
	}
	return e.Message
}

// IsValidationError checks if the error is a validation error
func (e *APIError) IsValidationError() bool {
	return e.Response != nil && e.Response.Error.Code == "VALIDATION_ERROR"
}

// IsAuthenticationError checks if the error is an authentication error
func (e *APIError) IsAuthenticationError() bool {
	return e.Response != nil && e.Response.Error.Code == "AUTHENTICATION_ERROR"
}

// IsAuthorizationError checks if the error is an authorization error
func (e *APIError) IsAuthorizationError() bool {
	return e.Response != nil && e.Response.Error.Code == "AUTHORIZATION_ERROR"
}

// IsRateLimitError checks if the error is a rate limit error
func (e *APIError) IsRateLimitError() bool {
	return e.Response != nil && e.Response.Error.Code == "RATE_LIMIT_EXCEEDED"
}

// IsNotFoundError checks if the error is a not found error
func (e *APIError) IsNotFoundError() bool {
	return e.Response != nil && e.Response.Error.Code == "NOT_FOUND"
}

// IsServiceUnavailableError checks if the error is a service unavailable error
func (e *APIError) IsServiceUnavailableError() bool {
	return e.Response != nil && e.Response.Error.Code == "SERVICE_UNAVAILABLE"
}

// IsTimeoutError checks if the error is a timeout error
func (e *APIError) IsTimeoutError() bool {
	return e.Response != nil && e.Response.Error.Code == "REQUEST_TIMEOUT"
}

// IsInternalError checks if the error is an internal error
func (e *APIError) IsInternalError() bool {
	return e.Response != nil && e.Response.Error.Code == "INTERNAL_ERROR"
}

// GetValidationErrors returns validation errors if any
func (e *APIError) GetValidationErrors() []ValidationError {
	if e.Response != nil {
		return e.Response.Error.Validation
	}
	return nil
}

// GetRequestID returns the request ID if available
func (e *APIError) GetRequestID() string {
	if e.Response != nil {
		return e.Response.RequestID
	}
	return ""
}

// GetTimestamp returns the error timestamp if available
func (e *APIError) GetTimestamp() string {
	if e.Response != nil {
		return e.Response.Timestamp
	}
	return ""
}

// GetPath returns the request path if available
func (e *APIError) GetPath() string {
	if e.Response != nil {
		return e.Response.Path
	}
	return ""
}

// GetMethod returns the request method if available
func (e *APIError) GetMethod() string {
	if e.Response != nil {
		return e.Response.Method
	}
	return ""
}
