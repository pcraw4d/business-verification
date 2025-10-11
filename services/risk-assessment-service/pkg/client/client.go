package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"kyb-platform/services/risk-assessment-service/internal/models"
)

// Client represents the Risk Assessment Service client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	userAgent  string
}

// Config holds client configuration
type Config struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	UserAgent  string
	HTTPClient *http.Client
}

// NewClient creates a new Risk Assessment Service client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}

	if config.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	// Set defaults
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	if config.UserAgent == "" {
		config.UserAgent = "kyb-go-client/1.0.0"
	}

	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: config.Timeout,
		}
	}

	return &Client{
		baseURL:    config.BaseURL,
		apiKey:     config.APIKey,
		httpClient: httpClient,
		userAgent:  config.UserAgent,
	}, nil
}

// AssessRisk performs a risk assessment for a business
func (c *Client) AssessRisk(ctx context.Context, req *models.RiskAssessmentRequest) (*models.RiskAssessmentResponse, error) {
	return c.assessRiskWithOptions(ctx, req, nil)
}

// AssessRiskWithOptions performs a risk assessment with additional options
func (c *Client) AssessRiskWithOptions(ctx context.Context, req *models.RiskAssessmentRequest, opts *RequestOptions) (*models.RiskAssessmentResponse, error) {
	return c.assessRiskWithOptions(ctx, req, opts)
}

func (c *Client) assessRiskWithOptions(ctx context.Context, req *models.RiskAssessmentRequest, opts *RequestOptions) (*models.RiskAssessmentResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	// Validate request
	if err := c.validateRiskAssessmentRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Prepare request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	resp, err := c.makeRequest(ctx, "POST", "/api/v1/assess", reqBody, opts)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var assessment models.RiskAssessmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&assessment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &assessment, nil
}

// GetRiskAssessment retrieves a risk assessment by ID
func (c *Client) GetRiskAssessment(ctx context.Context, id string) (*models.RiskAssessmentResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("assessment ID is required")
	}

	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/api/v1/assess/%s", id), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var assessment models.RiskAssessmentResponse
	if err := json.NewDecoder(resp.Body).Decode(&assessment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &assessment, nil
}

// PredictRisk performs future risk prediction for a business
func (c *Client) PredictRisk(ctx context.Context, id string, req *RiskPredictionRequest) (*models.RiskPrediction, error) {
	if id == "" {
		return nil, fmt.Errorf("assessment ID is required")
	}

	if req == nil {
		return nil, fmt.Errorf("prediction request cannot be nil")
	}

	// Validate prediction request
	if err := c.validatePredictionRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Prepare request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	resp, err := c.makeRequest(ctx, "POST", fmt.Sprintf("/api/v1/assess/%s/predict", id), reqBody, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var prediction models.RiskPrediction
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &prediction, nil
}

// PredictRiskWithHorizon performs risk prediction with specific model selection
func (c *Client) PredictRiskWithHorizon(ctx context.Context, id string, horizonMonths int, modelType string) (*models.RiskPrediction, error) {
	if id == "" {
		return nil, fmt.Errorf("assessment ID is required")
	}

	if horizonMonths <= 0 || horizonMonths > 24 {
		return nil, fmt.Errorf("horizon_months must be between 1 and 24")
	}

	req := &RiskPredictionRequest{
		HorizonMonths:           horizonMonths,
		ModelType:               modelType,
		IncludeTemporalAnalysis: true,
	}

	return c.PredictRisk(ctx, id, req)
}

// PredictMultiHorizon performs advanced multi-horizon risk prediction
func (c *Client) PredictMultiHorizon(ctx context.Context, req *AdvancedPredictionRequest) (*AdvancedPredictionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("prediction request cannot be nil")
	}

	// Validate advanced prediction request
	if err := c.validateAdvancedPredictionRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Prepare request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	resp, err := c.makeRequest(ctx, "POST", "/api/v1/risk/predict-advanced", reqBody, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var prediction AdvancedPredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&prediction); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &prediction, nil
}

// PredictWithLSTM performs LSTM-specific risk prediction
func (c *Client) PredictWithLSTM(ctx context.Context, business *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	if business == nil {
		return nil, fmt.Errorf("business data is required")
	}

	// Create advanced prediction request for LSTM
	req := &AdvancedPredictionRequest{
		Business:                business,
		PredictionHorizons:      []int{horizonMonths},
		ModelPreference:         "lstm",
		IncludeTemporalAnalysis: true,
		IncludeScenarioAnalysis: true,
		IncludeModelComparison:  false,
		ConfidenceThreshold:     0.7,
	}

	// Make request
	response, err := c.PredictMultiHorizon(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LSTM prediction failed: %w", err)
	}

	// Extract prediction for the requested horizon
	horizonKey := fmt.Sprintf("%d", horizonMonths)
	prediction, exists := response.Predictions[horizonKey]
	if !exists {
		return nil, fmt.Errorf("prediction not found for horizon %d months", horizonMonths)
	}

	// Convert to RiskPrediction format
	return &models.RiskPrediction{
		BusinessID:       response.BusinessID,
		PredictionDate:   prediction.PredictionDate,
		HorizonMonths:    prediction.HorizonMonths,
		PredictedScore:   prediction.PredictedScore,
		PredictedLevel:   prediction.PredictedLevel,
		ConfidenceScore:  prediction.ConfidenceScore,
		RiskFactors:      prediction.RiskFactors,
		ScenarioAnalysis: prediction.ScenarioAnalysis,
		CreatedAt:        response.GeneratedAt,
	}, nil
}

// PredictWithEnsemble performs ensemble risk prediction
func (c *Client) PredictWithEnsemble(ctx context.Context, business *models.RiskAssessmentRequest, horizonMonths int) (*models.RiskPrediction, error) {
	if business == nil {
		return nil, fmt.Errorf("business data is required")
	}

	// Create advanced prediction request for ensemble
	req := &AdvancedPredictionRequest{
		Business:                business,
		PredictionHorizons:      []int{horizonMonths},
		ModelPreference:         "ensemble",
		IncludeTemporalAnalysis: true,
		IncludeScenarioAnalysis: true,
		IncludeModelComparison:  true,
		ConfidenceThreshold:     0.7,
	}

	// Make request
	response, err := c.PredictMultiHorizon(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("ensemble prediction failed: %w", err)
	}

	// Extract prediction for the requested horizon
	horizonKey := fmt.Sprintf("%d", horizonMonths)
	prediction, exists := response.Predictions[horizonKey]
	if !exists {
		return nil, fmt.Errorf("prediction not found for horizon %d months", horizonMonths)
	}

	// Convert to RiskPrediction format
	return &models.RiskPrediction{
		BusinessID:       response.BusinessID,
		PredictionDate:   prediction.PredictionDate,
		HorizonMonths:    prediction.HorizonMonths,
		PredictedScore:   prediction.PredictedScore,
		PredictedLevel:   prediction.PredictedLevel,
		ConfidenceScore:  prediction.ConfidenceScore,
		RiskFactors:      prediction.RiskFactors,
		ScenarioAnalysis: prediction.ScenarioAnalysis,
		CreatedAt:        response.GeneratedAt,
	}, nil
}

// GetModelInfo retrieves information about available models
func (c *Client) GetModelInfo(ctx context.Context, modelType string) (*ModelInfo, error) {
	if modelType == "" {
		return nil, fmt.Errorf("model type is required")
	}

	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/api/v1/models/%s/info", modelType), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var modelInfo ModelInfo
	if err := json.NewDecoder(resp.Body).Decode(&modelInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &modelInfo, nil
}

// GetModelPerformance retrieves performance metrics for models
func (c *Client) GetModelPerformance(ctx context.Context) (*ModelPerformanceResponse, error) {
	resp, err := c.makeRequest(ctx, "GET", "/api/v1/models/performance", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var performance ModelPerformanceResponse
	if err := json.NewDecoder(resp.Body).Decode(&performance); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &performance, nil
}

// GetRiskHistory retrieves risk assessment history for a business
func (c *Client) GetRiskHistory(ctx context.Context, id string) (*RiskHistoryResponse, error) {
	if id == "" {
		return nil, fmt.Errorf("assessment ID is required")
	}

	resp, err := c.makeRequest(ctx, "GET", fmt.Sprintf("/api/v1/assess/%s/history", id), nil, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var history RiskHistoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &history, nil
}

// CheckCompliance performs compliance checks for a business
func (c *Client) CheckCompliance(ctx context.Context, req *ComplianceCheckRequest) (*ComplianceCheckResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("compliance request cannot be nil")
	}

	// Validate compliance request
	if err := c.validateComplianceRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Prepare request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	resp, err := c.makeRequest(ctx, "POST", "/api/v1/compliance/check", reqBody, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var compliance ComplianceCheckResponse
	if err := json.NewDecoder(resp.Body).Decode(&compliance); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &compliance, nil
}

// ScreenSanctions performs sanctions screening for a business
func (c *Client) ScreenSanctions(ctx context.Context, req *SanctionsScreeningRequest) (*SanctionsScreeningResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("sanctions request cannot be nil")
	}

	// Validate sanctions request
	if err := c.validateSanctionsRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Prepare request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	resp, err := c.makeRequest(ctx, "POST", "/api/v1/sanctions/screen", reqBody, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var sanctions SanctionsScreeningResponse
	if err := json.NewDecoder(resp.Body).Decode(&sanctions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &sanctions, nil
}

// MonitorMedia sets up adverse media monitoring for a business
func (c *Client) MonitorMedia(ctx context.Context, req *MediaMonitoringRequest) (*MediaMonitoringResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("media monitoring request cannot be nil")
	}

	// Validate media monitoring request
	if err := c.validateMediaMonitoringRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Prepare request
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request
	resp, err := c.makeRequest(ctx, "POST", "/api/v1/media/monitor", reqBody, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var media MediaMonitoringResponse
	if err := json.NewDecoder(resp.Body).Decode(&media); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &media, nil
}

// GetRiskTrends retrieves risk trends and analytics
func (c *Client) GetRiskTrends(ctx context.Context, opts *RiskTrendsOptions) (*RiskTrendsResponse, error) {
	// Build query parameters
	params := url.Values{}
	if opts != nil {
		if opts.Industry != "" {
			params.Set("industry", opts.Industry)
		}
		if opts.Country != "" {
			params.Set("country", opts.Country)
		}
		if opts.Timeframe != "" {
			params.Set("timeframe", opts.Timeframe)
		}
		if opts.Limit > 0 {
			params.Set("limit", fmt.Sprintf("%d", opts.Limit))
		}
	}

	// Build URL
	url := "/api/v1/analytics/trends"
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	resp, err := c.makeRequest(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var trends RiskTrendsResponse
	if err := json.NewDecoder(resp.Body).Decode(&trends); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &trends, nil
}

// GetRiskInsights retrieves risk insights and recommendations
func (c *Client) GetRiskInsights(ctx context.Context, opts *RiskInsightsOptions) (*RiskInsightsResponse, error) {
	// Build query parameters
	params := url.Values{}
	if opts != nil {
		if opts.Industry != "" {
			params.Set("industry", opts.Industry)
		}
		if opts.Country != "" {
			params.Set("country", opts.Country)
		}
		if opts.RiskLevel != "" {
			params.Set("risk_level", opts.RiskLevel)
		}
	}

	// Build URL
	url := "/api/v1/analytics/insights"
	if len(params) > 0 {
		url += "?" + params.Encode()
	}

	resp, err := c.makeRequest(ctx, "GET", url, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var insights RiskInsightsResponse
	if err := json.NewDecoder(resp.Body).Decode(&insights); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &insights, nil
}

// makeRequest makes an HTTP request to the API
func (c *Client) makeRequest(ctx context.Context, method, path string, body []byte, opts *RequestOptions) (*http.Response, error) {
	// Build URL
	fullURL := c.baseURL + path

	// Create request
	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	// Add custom headers if provided
	if opts != nil && opts.Headers != nil {
		for key, value := range opts.Headers {
			req.Header.Set(key, value)
		}
	}

	// Make request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()

		// Try to parse error response
		var errorResp ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			return nil, &APIError{
				StatusCode: resp.StatusCode,
				Response:   &errorResp,
			}
		}

		// Fallback to generic error
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status),
		}
	}

	return resp, nil
}

// Validation methods
func (c *Client) validateRiskAssessmentRequest(req *models.RiskAssessmentRequest) error {
	if req.BusinessName == "" {
		return fmt.Errorf("business_name is required")
	}
	if req.BusinessAddress == "" {
		return fmt.Errorf("business_address is required")
	}
	if req.Industry == "" {
		return fmt.Errorf("industry is required")
	}
	if req.Country == "" {
		return fmt.Errorf("country is required")
	}
	if len(req.Country) != 2 {
		return fmt.Errorf("country must be a 2-letter ISO code")
	}
	if req.PredictionHorizon < 0 || req.PredictionHorizon > 24 {
		return fmt.Errorf("prediction_horizon must be between 0 and 24 months")
	}
	return nil
}

func (c *Client) validatePredictionRequest(req *RiskPredictionRequest) error {
	if req.HorizonMonths <= 0 || req.HorizonMonths > 24 {
		return fmt.Errorf("horizon_months must be between 1 and 24")
	}
	return nil
}

func (c *Client) validateAdvancedPredictionRequest(req *AdvancedPredictionRequest) error {
	if req.Business == nil {
		return fmt.Errorf("business data is required")
	}
	if len(req.PredictionHorizons) == 0 {
		return fmt.Errorf("at least one prediction horizon is required")
	}
	if len(req.PredictionHorizons) > 5 {
		return fmt.Errorf("maximum of 5 prediction horizons allowed")
	}
	for _, horizon := range req.PredictionHorizons {
		if horizon < 1 || horizon > 24 {
			return fmt.Errorf("prediction horizon must be between 1 and 24 months")
		}
	}
	if req.ModelPreference != "" {
		validModels := []string{"auto", "xgboost", "lstm", "ensemble"}
		valid := false
		for _, model := range validModels {
			if req.ModelPreference == model {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid model preference: %s", req.ModelPreference)
		}
	}
	if req.ConfidenceThreshold < 0 || req.ConfidenceThreshold > 1 {
		return fmt.Errorf("confidence threshold must be between 0 and 1")
	}
	return nil
}

func (c *Client) validateComplianceRequest(req *ComplianceCheckRequest) error {
	if req.BusinessName == "" {
		return fmt.Errorf("business_name is required")
	}
	if req.BusinessAddress == "" {
		return fmt.Errorf("business_address is required")
	}
	if req.Industry == "" {
		return fmt.Errorf("industry is required")
	}
	if req.Country == "" {
		return fmt.Errorf("country is required")
	}
	return nil
}

func (c *Client) validateSanctionsRequest(req *SanctionsScreeningRequest) error {
	if req.BusinessName == "" {
		return fmt.Errorf("business_name is required")
	}
	if req.BusinessAddress == "" {
		return fmt.Errorf("business_address is required")
	}
	if req.Country == "" {
		return fmt.Errorf("country is required")
	}
	return nil
}

func (c *Client) validateMediaMonitoringRequest(req *MediaMonitoringRequest) error {
	if req.BusinessName == "" {
		return fmt.Errorf("business_name is required")
	}
	if req.BusinessAddress == "" {
		return fmt.Errorf("business_address is required")
	}
	return nil
}
