package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

// PythonMLService represents the Python ML service for all ML models
type PythonMLService struct {
	// Service configuration
	endpoint string
	config   PythonMLServiceConfig

	// Model management
	models map[string]*MLModel

	// Performance tracking
	metrics *ServiceMetrics

	// Health status
	healthStatus *HealthStatus

	// HTTP client for communication
	httpClient *http.Client

	// Thread safety
	mu sync.RWMutex

	// Logging
	logger *log.Logger

	// Control
	ctx    context.Context
	cancel context.CancelFunc
}

// PythonMLServiceConfig holds configuration for the Python ML service
type PythonMLServiceConfig struct {
	// Service configuration
	Host string `json:"host"`
	Port int    `json:"port"`

	// Model configuration
	DefaultModelType    string        `json:"default_model_type"` // bert, distilbert, custom
	SupportedModelTypes []string      `json:"supported_model_types"`
	ModelCacheEnabled   bool          `json:"model_cache_enabled"`
	ModelCacheSize      int           `json:"model_cache_size"`
	ModelUpdateInterval time.Duration `json:"model_update_interval"`

	// Performance configuration
	MaxBatchSize        int           `json:"max_batch_size"`
	InferenceTimeout    time.Duration `json:"inference_timeout"`
	ModelLoadingTimeout time.Duration `json:"model_loading_timeout"`

	// Resource limits
	MaxMemoryUsage      int64 `json:"max_memory_usage"` // in MB
	MaxCPUUsage         int   `json:"max_cpu_usage"`    // percentage
	MaxConcurrentModels int   `json:"max_concurrent_models"`

	// Monitoring
	MetricsEnabled      bool `json:"metrics_enabled"`
	PerformanceTracking bool `json:"performance_tracking"`
	ModelVersioning     bool `json:"model_versioning"`
}

// MLModel represents a machine learning model in the Python service
type MLModel struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"` // bert, distilbert, custom
	Version    string    `json:"version"`
	ModelPath  string    `json:"model_path"`
	ConfigPath string    `json:"config_path"`
	IsActive   bool      `json:"is_active"`
	IsDeployed bool      `json:"is_deployed"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	LastUsed   time.Time `json:"last_used"`
}

// ClassificationRequest represents a request for classification
type ClassificationRequest struct {
	BusinessName        string  `json:"business_name"`
	Description         string  `json:"description"`
	WebsiteURL          string  `json:"website_url"`
	ModelType           string  `json:"model_type,omitempty"` // bert, distilbert, custom
	ModelVersion        string  `json:"model_version,omitempty"`
	MaxResults          int     `json:"max_results,omitempty"`
	ConfidenceThreshold float64 `json:"confidence_threshold,omitempty"`
}

// RiskDetectionRequest represents a request for risk detection
type RiskDetectionRequest struct {
	BusinessName   string   `json:"business_name"`
	Description    string   `json:"description"`
	WebsiteURL     string   `json:"website_url"`
	WebsiteContent string   `json:"website_content,omitempty"`
	ModelType      string   `json:"model_type,omitempty"`
	ModelVersion   string   `json:"model_version,omitempty"`
	RiskCategories []string `json:"risk_categories,omitempty"` // illegal, prohibited, high_risk, tbml
}

// ClassificationResponse represents a response from classification
type ClassificationResponse struct {
	RequestID       string                     `json:"request_id"`
	ModelID         string                     `json:"model_id"`
	ModelVersion    string                     `json:"model_version"`
	Classifications []ClassificationPrediction `json:"classifications"`
	Confidence      float64                    `json:"confidence"`
	ProcessingTime  time.Duration              `json:"processing_time"`
	Timestamp       time.Time                  `json:"timestamp"`
	Success         bool                       `json:"success"`
	Error           string                     `json:"error,omitempty"`
}

// RiskDetectionResponse represents a response from risk detection
type RiskDetectionResponse struct {
	RequestID      string         `json:"request_id"`
	ModelID        string         `json:"model_id"`
	ModelVersion   string         `json:"model_version"`
	RiskScore      float64        `json:"risk_score"`
	RiskLevel      string         `json:"risk_level"` // low, medium, high, critical
	DetectedRisks  []DetectedRisk `json:"detected_risks"`
	ProcessingTime time.Duration  `json:"processing_time"`
	Timestamp      time.Time      `json:"timestamp"`
	Success        bool           `json:"success"`
	Error          string         `json:"error,omitempty"`
}

// ClassificationPrediction represents a single classification prediction
type ClassificationPrediction struct {
	Label       string  `json:"label"`
	Confidence  float64 `json:"confidence"`
	Probability float64 `json:"probability"`
	Rank        int     `json:"rank"`
}

// DetectedRisk represents a detected risk
type DetectedRisk struct {
	Category    string   `json:"category"` // illegal, prohibited, high_risk, tbml
	Severity    string   `json:"severity"` // low, medium, high, critical
	Confidence  float64  `json:"confidence"`
	Keywords    []string `json:"keywords"`
	Description string   `json:"description"`
}

// ModelMetrics represents model performance metrics
type ModelMetrics struct {
	ModelID       string        `json:"model_id"`
	ModelVersion  string        `json:"model_version"`
	Accuracy      float64       `json:"accuracy"`
	Precision     float64       `json:"precision"`
	Recall        float64       `json:"recall"`
	F1Score       float64       `json:"f1_score"`
	InferenceTime time.Duration `json:"inference_time"`
	Throughput    int           `json:"throughput"` // requests per second
	RequestCount  int64         `json:"request_count"`
	SuccessCount  int64         `json:"success_count"`
	ErrorCount    int64         `json:"error_count"`
	LastUpdated   time.Time     `json:"last_updated"`
}

// NewPythonMLService creates a new Python ML service
func NewPythonMLService(endpoint string, logger *log.Logger) *PythonMLService {
	if logger == nil {
		logger = log.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &PythonMLService{
		endpoint: endpoint,
		config: PythonMLServiceConfig{
			DefaultModelType:    "bert",
			SupportedModelTypes: []string{"bert", "distilbert", "custom"},
			ModelCacheEnabled:   true,
			ModelCacheSize:      10,
			ModelUpdateInterval: 24 * time.Hour,
			MaxBatchSize:        32,
			InferenceTimeout:    5 * time.Second,
			ModelLoadingTimeout: 30 * time.Second,
			MaxMemoryUsage:      4096, // 4GB
			MaxCPUUsage:         80,   // 80%
			MaxConcurrentModels: 5,
			MetricsEnabled:      true,
			PerformanceTracking: true,
			ModelVersioning:     true,
		},
		models:     make(map[string]*MLModel),
		httpClient: httpClient,
		logger:     logger,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Initialize initializes the Python ML service
func (pms *PythonMLService) Initialize(ctx context.Context) error {
	pms.mu.Lock()
	defer pms.mu.Unlock()

	pms.logger.Printf("🐍 Initializing Python ML Service at %s", pms.endpoint)

	// Initialize metrics
	pms.metrics = &ServiceMetrics{
		RequestCount:   0,
		SuccessCount:   0,
		ErrorCount:     0,
		AverageLatency: 0,
		P95Latency:     0,
		P99Latency:     0,
		Throughput:     0,
		ErrorRate:      0,
		LastUpdated:    time.Now(),
	}

	// Initialize health status
	pms.healthStatus = &HealthStatus{
		Status:    "unknown",
		LastCheck: time.Now(),
		Checks:    make(map[string]HealthCheck),
	}

	// Test connection to Python service
	if err := pms.testConnection(ctx); err != nil {
		return fmt.Errorf("failed to connect to Python ML service: %w", err)
	}

	// Load available models
	if err := pms.loadAvailableModels(ctx); err != nil {
		pms.logger.Printf("⚠️ Warning: failed to load available models: %v", err)
	}

	pms.logger.Printf("✅ Python ML Service initialized successfully")
	return nil
}

// Start starts the Python ML service
func (pms *PythonMLService) Start(ctx context.Context) error {
	pms.mu.Lock()
	defer pms.mu.Unlock()

	pms.logger.Printf("🚀 Starting Python ML Service")

	// Start health monitoring
	go pms.startHealthMonitoring(ctx)

	// Start metrics collection
	if pms.config.MetricsEnabled {
		go pms.startMetricsCollection(ctx)
	}

	// Start model cache management
	if pms.config.ModelCacheEnabled {
		go pms.startModelCacheManagement(ctx)
	}

	pms.logger.Printf("✅ Python ML Service started successfully")
	return nil
}

// Stop stops the Python ML service
func (pms *PythonMLService) Stop() {
	pms.mu.Lock()
	defer pms.mu.Unlock()

	pms.logger.Printf("🛑 Stopping Python ML Service")

	// Cancel context
	pms.cancel()

	pms.logger.Printf("✅ Python ML Service stopped successfully")
}

// Classify performs business classification using the Python ML service
func (pms *PythonMLService) Classify(ctx context.Context, req *ClassificationRequest) (*ClassificationResponse, error) {
	start := time.Now()

	pms.mu.Lock()
	pms.metrics.RequestCount++
	pms.mu.Unlock()

	// Prepare request
	requestBody, err := json.Marshal(req)
	if err != nil {
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to Python service
	httpReq, err := http.NewRequestWithContext(ctx, "POST", pms.endpoint+"/classify",
		bytes.NewBuffer(requestBody))
	if err != nil {
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := pms.httpClient.Do(httpReq)
	if err != nil {
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var classificationResp ClassificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&classificationResp); err != nil {
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Update metrics
	processingTime := time.Since(start)
	pms.mu.Lock()
	if classificationResp.Success {
		pms.metrics.SuccessCount++
	} else {
		pms.metrics.ErrorCount++
	}
	pms.updateLatencyMetrics(processingTime)
	pms.mu.Unlock()

	// Set processing time
	classificationResp.ProcessingTime = processingTime

	return &classificationResp, nil
}

// DetectRisk performs risk detection using the Python ML service
func (pms *PythonMLService) DetectRisk(ctx context.Context, req *RiskDetectionRequest) (*RiskDetectionResponse, error) {
	start := time.Now()

	pms.mu.Lock()
	pms.metrics.RequestCount++
	pms.mu.Unlock()

	// Prepare request
	requestBody, err := json.Marshal(req)
	if err != nil {
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to Python service
	httpReq, err := http.NewRequestWithContext(ctx, "POST", pms.endpoint+"/detect-risk",
		bytes.NewBuffer(requestBody))
	if err != nil {
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := pms.httpClient.Do(httpReq)
	if err != nil {
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Parse response
	var riskResp RiskDetectionResponse
	if err := json.NewDecoder(resp.Body).Decode(&riskResp); err != nil {
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Update metrics
	processingTime := time.Since(start)
	pms.mu.Lock()
	if riskResp.Success {
		pms.metrics.SuccessCount++
	} else {
		pms.metrics.ErrorCount++
	}
	pms.updateLatencyMetrics(processingTime)
	pms.mu.Unlock()

	// Set processing time
	riskResp.ProcessingTime = processingTime

	return &riskResp, nil
}

// GetAvailableModels returns available models in the Python service
func (pms *PythonMLService) GetAvailableModels(ctx context.Context) ([]*MLModel, error) {
	// Make HTTP request to get models
	httpReq, err := http.NewRequestWithContext(ctx, "GET", pms.endpoint+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := pms.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var models []*MLModel
	if err := json.NewDecoder(resp.Body).Decode(&models); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return models, nil
}

// GetModelMetrics returns metrics for a specific model
func (pms *PythonMLService) GetModelMetrics(ctx context.Context, modelID string) (*ModelMetrics, error) {
	// Make HTTP request to get model metrics
	httpReq, err := http.NewRequestWithContext(ctx, "GET",
		fmt.Sprintf("%s/models/%s/metrics", pms.endpoint, modelID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := pms.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	var metrics ModelMetrics
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &metrics, nil
}

// HealthCheck performs a health check on the Python ML service
func (pms *PythonMLService) HealthCheck(ctx context.Context) (*HealthCheck, error) {
	start := time.Now()

	// Make HTTP request to health endpoint
	httpReq, err := http.NewRequestWithContext(ctx, "GET", pms.endpoint+"/health", nil)
	if err != nil {
		return &HealthCheck{
			Name:      "python_ml_service",
			Status:    "fail",
			Message:   fmt.Sprintf("Failed to create health check request: %v", err),
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}, nil
	}

	resp, err := pms.httpClient.Do(httpReq)
	if err != nil {
		return &HealthCheck{
			Name:      "python_ml_service",
			Status:    "fail",
			Message:   fmt.Sprintf("Health check request failed: %v", err),
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}, nil
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return &HealthCheck{
			Name:      "python_ml_service",
			Status:    "fail",
			Message:   fmt.Sprintf("Health check returned status %d", resp.StatusCode),
			LastCheck: time.Now(),
			Duration:  time.Since(start),
		}, nil
	}

	return &HealthCheck{
		Name:      "python_ml_service",
		Status:    "pass",
		Message:   "Service is healthy",
		LastCheck: time.Now(),
		Duration:  time.Since(start),
	}, nil
}

// GetMetrics returns service metrics
func (pms *PythonMLService) GetMetrics(ctx context.Context) (*ServiceMetrics, error) {
	pms.mu.RLock()
	defer pms.mu.RUnlock()

	// Return a copy of metrics
	metrics := *pms.metrics
	return &metrics, nil
}

// testConnection tests the connection to the Python ML service
func (pms *PythonMLService) testConnection(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", pms.endpoint+"/ping", nil)
	if err != nil {
		return fmt.Errorf("failed to create ping request: %w", err)
	}

	resp, err := pms.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("ping request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping returned status %d", resp.StatusCode)
	}

	return nil
}

// loadAvailableModels loads available models from the Python service
func (pms *PythonMLService) loadAvailableModels(ctx context.Context) error {
	models, err := pms.GetAvailableModels(ctx)
	if err != nil {
		return fmt.Errorf("failed to get available models: %w", err)
	}

	pms.mu.Lock()
	defer pms.mu.Unlock()

	for _, model := range models {
		pms.models[model.ID] = model
	}

	pms.logger.Printf("📚 Loaded %d models from Python ML service", len(models))
	return nil
}

// startHealthMonitoring starts health monitoring for the service
func (pms *PythonMLService) startHealthMonitoring(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			healthCheck, err := pms.HealthCheck(ctx)
			if err != nil {
				pms.logger.Printf("⚠️ Health check failed: %v", err)
				continue
			}

			pms.mu.Lock()
			pms.healthStatus.Status = healthCheck.Status
			pms.healthStatus.LastCheck = healthCheck.LastCheck
			pms.healthStatus.Checks["python_ml_service"] = *healthCheck
			pms.mu.Unlock()

			if healthCheck.Status != "pass" {
				pms.logger.Printf("⚠️ Python ML Service health check failed: %s", healthCheck.Message)
			}
		}
	}
}

// startMetricsCollection starts metrics collection
func (pms *PythonMLService) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			pms.updateMetrics()
		}
	}
}

// startModelCacheManagement starts model cache management
func (pms *PythonMLService) startModelCacheManagement(ctx context.Context) {
	ticker := time.NewTicker(pms.config.ModelUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := pms.loadAvailableModels(ctx); err != nil {
				pms.logger.Printf("⚠️ Failed to update model cache: %v", err)
			}
		}
	}
}

// updateLatencyMetrics updates latency metrics
func (pms *PythonMLService) updateLatencyMetrics(latency time.Duration) {
	// Simple moving average for average latency
	if pms.metrics.AverageLatency == 0 {
		pms.metrics.AverageLatency = latency
	} else {
		pms.metrics.AverageLatency = (pms.metrics.AverageLatency + latency) / 2
	}

	// Update P95 and P99 (simplified implementation)
	if latency > pms.metrics.P95Latency {
		pms.metrics.P95Latency = latency
	}
	if latency > pms.metrics.P99Latency {
		pms.metrics.P99Latency = latency
	}

	pms.metrics.LastUpdated = time.Now()
}

// updateMetrics updates service metrics
func (pms *PythonMLService) updateMetrics() {
	pms.mu.Lock()
	defer pms.mu.Unlock()

	// Calculate error rate
	if pms.metrics.RequestCount > 0 {
		pms.metrics.ErrorRate = float64(pms.metrics.ErrorCount) / float64(pms.metrics.RequestCount)
	}

	// Calculate throughput (requests per second over last minute)
	// This is a simplified calculation
	pms.metrics.Throughput = float64(pms.metrics.RequestCount) / 60.0

	pms.metrics.LastUpdated = time.Now()
}
