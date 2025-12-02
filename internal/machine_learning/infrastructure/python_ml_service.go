package infrastructure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/resilience"
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

	// Circuit breaker for resilience
	circuitBreaker *resilience.CircuitBreaker
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
	MaxBatchSize            int           `json:"max_batch_size"`
	InferenceTimeout        time.Duration `json:"inference_timeout"`
	ModelLoadingTimeout     time.Duration `json:"model_loading_timeout"`
	LightweightModelTimeout time.Duration `json:"lightweight_model_timeout"` // Timeout for lightweight model fast-path

	// Resource limits
	MaxMemoryUsage      int64 `json:"max_memory_usage"` // in MB
	MaxCPUUsage         int   `json:"max_cpu_usage"`    // percentage
	MaxConcurrentModels int   `json:"max_concurrent_models"`

	// Monitoring
	MetricsEnabled      bool `json:"metrics_enabled"`
	PerformanceTracking bool `json:"performance_tracking"`
	ModelVersioning     bool `json:"model_versioning"`
}

// Types MLModel, ClassificationRequest, RiskDetectionRequest, ClassificationResponse,
// RiskDetectionResponse, ClassificationPrediction, DetectedRisk, and ModelMetrics
// are defined in types.go to avoid redeclaration

// NewPythonMLService creates a new Python ML service
func NewPythonMLService(endpoint string, logger *log.Logger) *PythonMLService {
	if logger == nil {
		logger = log.Default()
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Normalize endpoint URL - remove trailing slash to avoid double slashes
	normalizedEndpoint := strings.TrimSuffix(endpoint, "/")

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Initialize circuit breaker with enhanced config for production resilience
	// Opens after 10 consecutive failures, stays open for 60s, needs 2 successes to close
	circuitBreakerConfig := resilience.DefaultCircuitBreakerConfig()
	circuitBreakerConfig.FailureThreshold = 10  // Increased from 5 to handle transient initialization failures
	circuitBreakerConfig.Timeout = 60 * time.Second // Increased from 30s to allow service recovery
	circuitBreakerConfig.SuccessThreshold = 2 // Keep at 2
	circuitBreakerConfig.ResetTimeout = 120 * time.Second // Increased from 60s
	circuitBreaker := resilience.NewCircuitBreaker(circuitBreakerConfig)

	return &PythonMLService{
		endpoint: normalizedEndpoint,
		config: PythonMLServiceConfig{
				DefaultModelType:        "bert",
				SupportedModelTypes:     []string{"bert", "distilbert", "custom"},
				ModelCacheEnabled:       true,
				ModelCacheSize:          10,
				ModelUpdateInterval:      24 * time.Hour,
				MaxBatchSize:            32,
				InferenceTimeout:        5 * time.Second,
				ModelLoadingTimeout:      30 * time.Second,
				LightweightModelTimeout:  5 * time.Second, // Default 5s for fast-path
				MaxMemoryUsage:           4096, // 4GB
				MaxCPUUsage:              80,   // 80%
				MaxConcurrentModels:      5,
				MetricsEnabled:           true,
				PerformanceTracking:      true,
				ModelVersioning:          true,
		},
		models:         make(map[string]*MLModel),
		httpClient:      httpClient,
		logger:         logger,
		ctx:            ctx,
		cancel:         cancel,
		circuitBreaker: circuitBreaker,
	}
}

// Initialize initializes the Python ML service
func (pms *PythonMLService) Initialize(ctx context.Context) error {
	pms.logger.Printf("🐍 Initializing Python ML Service at %s", pms.endpoint)

	// Initialize metrics and health status (need lock for these)
	pms.mu.Lock()
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
	pms.mu.Unlock()

	// Test connection to Python service (no lock needed)
	// Use a shorter timeout for initialization to prevent hanging
	initCtx, initCancel := context.WithTimeout(ctx, 5*time.Second)
	defer initCancel()
	
	if err := pms.testConnection(initCtx); err != nil {
		return fmt.Errorf("failed to connect to Python ML service: %w", err)
	}

	// Load available models (this will acquire its own lock)
	// Use a separate timeout for model loading to prevent blocking initialization
	modelsCtx, modelsCancel := context.WithTimeout(ctx, 5*time.Second)
	defer modelsCancel()
	
	if err := pms.loadAvailableModels(modelsCtx); err != nil {
		pms.logger.Printf("⚠️ Warning: failed to load available models: %v", err)
		// Don't fail initialization if models can't be loaded - they can be loaded later
	}

	pms.logger.Printf("✅ Python ML Service initialized successfully")
	return nil
}

// InitializeWithRetry initializes the Python ML service with retry logic and exponential backoff
// This provides resilience during initialization, especially when the service is starting up
func (pms *PythonMLService) InitializeWithRetry(ctx context.Context, maxRetries int) error {
	// Reset circuit breaker before initialization to clear any previous failure state
	pms.ResetCircuitBreaker()
	pms.logger.Printf("🔄 Circuit breaker reset, starting initialization with up to %d retries", maxRetries)

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			// Exponential backoff: wait 2s, 4s, 6s, etc.
			waitTime := time.Duration(i) * 2 * time.Second
			pms.logger.Printf("⏳ Retrying initialization (attempt %d/%d) after %v", i+1, maxRetries, waitTime)
			time.Sleep(waitTime)
		}

		err := pms.Initialize(ctx)
		if err == nil {
			// Verify health before marking as ready
			healthCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			health, healthErr := pms.HealthCheck(healthCtx)
			if healthErr == nil && health != nil && health.Status == "pass" {
				pms.logger.Printf("✅ Python ML Service initialized and healthy")
				return nil
			}
			if healthErr != nil {
				pms.logger.Printf("⚠️ Health check failed: %v", healthErr)
			} else if health != nil {
				pms.logger.Printf("⚠️ Health check returned status: %s", health.Status)
			}
		}
		lastErr = err
	}

	// Graceful degradation: mark as initialized but unavailable
	pms.logger.Printf("⚠️ Python ML Service initialization failed after %d retries, continuing with degraded mode", maxRetries)
	return fmt.Errorf("initialization failed after %d retries: %w", maxRetries, lastErr)
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

// ClassifyFast performs fast classification using lightweight model (Task 3.1)
// Protected by circuit breaker to prevent cascading failures
func (pms *PythonMLService) ClassifyFast(
	ctx context.Context,
	req *EnhancedClassificationRequest,
) (*EnhancedClassificationResponse, error) {
	start := time.Now()

	pms.mu.Lock()
	if pms.metrics == nil {
		pms.metrics = &ServiceMetrics{}
	}
	pms.metrics.RequestCount++
	pms.mu.Unlock()

	// Check circuit breaker state before making request
	cbState := pms.circuitBreaker.GetState()
	if cbState == resilience.CircuitOpen {
		pms.logger.Printf("⚠️ [CircuitBreaker] Circuit is OPEN - failing fast for Python ML service")
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, fmt.Errorf("circuit breaker is open: Python ML service unavailable")
	}

	// Execute through circuit breaker
	var enhancedResp *EnhancedClassificationResponse
	var err error

	err = pms.circuitBreaker.Execute(ctx, func() error {
		// Prepare request
		requestBody, marshalErr := json.Marshal(req)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal request: %w", marshalErr)
		}

		// Make HTTP request to fast endpoint
		httpReq, createErr := http.NewRequestWithContext(
			ctx,
			"POST",
			pms.endpoint+"/classify-fast",
			bytes.NewBuffer(requestBody),
		)
		if createErr != nil {
			return fmt.Errorf("failed to create request: %w", createErr)
		}

		httpReq.Header.Set("Content-Type", "application/json")

		// Execute request with timeout for fast path
		// Use configurable timeout with default fallback
		lightweightTimeout := pms.config.LightweightModelTimeout
		if lightweightTimeout == 0 {
			lightweightTimeout = 5 * time.Second // Default fallback
		}
		fastCtx, cancel := context.WithTimeout(ctx, lightweightTimeout)
		defer cancel()
		httpReq = httpReq.WithContext(fastCtx)

		resp, doErr := pms.httpClient.Do(httpReq)
		if doErr != nil {
			return fmt.Errorf("failed to execute request: %w", doErr)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			if len(body) > 0 {
				return fmt.Errorf("Python service returned status %d: %s", resp.StatusCode, string(body))
			}
			return fmt.Errorf("Python service returned status %d", resp.StatusCode)
		}

		// Parse response
		var respData EnhancedClassificationResponse
		if decodeErr := json.NewDecoder(resp.Body).Decode(&respData); decodeErr != nil {
			return fmt.Errorf("failed to decode response: %w", decodeErr)
		}

		enhancedResp = &respData
		return nil
	})

	// Handle circuit breaker errors
	if err != nil {
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, err
	}

	// Update metrics
	processingTime := time.Since(start)
	pms.mu.Lock()
	if enhancedResp.Success {
		pms.metrics.SuccessCount++
	} else {
		pms.metrics.ErrorCount++
	}
	pms.mu.Unlock()

	pms.logger.Printf("✅ Fast classification completed in %v", processingTime)
	return enhancedResp, nil
}

// ClassifyEnhanced performs enhanced classification with summarization and explanation
// Protected by circuit breaker to prevent cascading failures
func (pms *PythonMLService) ClassifyEnhanced(
	ctx context.Context,
	req *EnhancedClassificationRequest,
) (*EnhancedClassificationResponse, error) {
	start := time.Now()

	pms.mu.Lock()
	if pms.metrics == nil {
		pms.metrics = &ServiceMetrics{}
	}
	pms.metrics.RequestCount++
	pms.mu.Unlock()

	// Check circuit breaker state before making request
	cbState := pms.circuitBreaker.GetState()
	if cbState == resilience.CircuitOpen {
		pms.logger.Printf("⚠️ [CircuitBreaker] Circuit is OPEN - failing fast for Python ML service")
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		return nil, fmt.Errorf("circuit breaker is open: Python ML service unavailable")
	}

	// Execute through circuit breaker
	var enhancedResp *EnhancedClassificationResponse
	var err error

	err = pms.circuitBreaker.Execute(ctx, func() error {
		// Prepare request
		requestBody, marshalErr := json.Marshal(req)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal request: %w", marshalErr)
		}

		// Make HTTP request
		httpReq, createErr := http.NewRequestWithContext(
			ctx,
			"POST",
			pms.endpoint+"/classify-enhanced",
			bytes.NewBuffer(requestBody),
		)
		if createErr != nil {
			return fmt.Errorf("failed to create request: %w", createErr)
		}

		httpReq.Header.Set("Content-Type", "application/json")

		// Execute request
		resp, doErr := pms.httpClient.Do(httpReq)
		if doErr != nil {
			return fmt.Errorf("failed to execute request: %w", doErr)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			// Read error response body for better error messages
			body, _ := io.ReadAll(resp.Body)
			if len(body) > 0 {
				return fmt.Errorf("Python service returned status %d: %s", resp.StatusCode, string(body))
			}
			return fmt.Errorf("Python service returned status %d", resp.StatusCode)
		}

		// Parse response
		var respData EnhancedClassificationResponse
		if decodeErr := json.NewDecoder(resp.Body).Decode(&respData); decodeErr != nil {
			return fmt.Errorf("failed to decode response: %w", decodeErr)
		}

		enhancedResp = &respData
		return nil
	})

	// Handle circuit breaker errors
	if err != nil {
		pms.mu.Lock()
		pms.metrics.ErrorCount++
		pms.mu.Unlock()
		
		// Log circuit breaker state changes
		newState := pms.circuitBreaker.GetState()
		if newState != cbState {
			pms.logger.Printf("🔄 [CircuitBreaker] State changed from %s to %s", 
				cbState.String(), newState.String())
		}
		
		return nil, err
	}

	// Update metrics
	processingTime := time.Since(start)
	pms.mu.Lock()
	if enhancedResp.Success {
		pms.metrics.SuccessCount++
	} else {
		pms.metrics.ErrorCount++
		// Treat unsuccessful response as error for circuit breaker
		pms.mu.Unlock()
		_ = pms.circuitBreaker.Execute(ctx, func() error {
			return fmt.Errorf("classification unsuccessful: success=false")
		})
		pms.mu.Lock()
	}
	pms.updateLatencyMetrics(processingTime)
	pms.mu.Unlock()

	// Convert processing_time from float64 to time.Duration for consistency
	// (The response already has processing_time as float64, so we keep it)
	enhancedResp.ProcessingTime = float64(processingTime.Seconds())

	return enhancedResp, nil
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
		// Check if error is due to context timeout/cancellation
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timed out: %w", err)
		}
		if ctx.Err() == context.Canceled {
			return nil, fmt.Errorf("request canceled: %w", err)
		}
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Handle 503 Service Unavailable (models still loading) - return empty list
	if resp.StatusCode == http.StatusServiceUnavailable {
		pms.logger.Printf("⚠️ Models are still loading, returning empty list")
		return []*MLModel{}, nil
	}

	// Handle other non-200 status codes
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

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
		// Check if error is due to context timeout/cancellation
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("ping request timed out: %w", err)
		}
		if ctx.Err() == context.Canceled {
			return fmt.Errorf("ping request canceled: %w", err)
		}
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

// ResetCircuitBreaker resets the circuit breaker to closed state
// This is useful during initialization to clear any previous failure state
func (pms *PythonMLService) ResetCircuitBreaker() {
	pms.circuitBreaker.Reset()
	pms.logger.Printf("🔄 Circuit breaker reset to closed state")
}

// GetCircuitBreakerState returns the current circuit breaker state
func (pms *PythonMLService) GetCircuitBreakerState() resilience.CircuitState {
	return pms.circuitBreaker.GetState()
}

// CircuitBreakerMetrics holds metrics about the circuit breaker
type CircuitBreakerMetrics struct {
	State            string    `json:"state"`
	FailureCount     int       `json:"failure_count"`
	SuccessCount     int       `json:"success_count"`
	StateChangeTime  time.Time `json:"state_change_time"`
	LastFailureTime  time.Time `json:"last_failure_time"`
	TotalRequests    int64     `json:"total_requests"`
	RejectedRequests int64    `json:"rejected_requests"`
}

// GetCircuitBreakerMetrics returns comprehensive metrics about the circuit breaker
func (pms *PythonMLService) GetCircuitBreakerMetrics() CircuitBreakerMetrics {
	stats := pms.circuitBreaker.GetStats()
	
	pms.mu.RLock()
	defer pms.mu.RUnlock()
	
	var totalRequests, rejectedRequests int64
	if pms.metrics != nil {
		totalRequests = pms.metrics.RequestCount
		rejectedRequests = pms.metrics.ErrorCount
	}
	
	return CircuitBreakerMetrics{
		State:            stats.State,
		FailureCount:     stats.FailureCount,
		SuccessCount:     stats.SuccessCount,
		StateChangeTime:  stats.StateChange,
		LastFailureTime:  stats.LastFailure,
		TotalRequests:    totalRequests,
		RejectedRequests: rejectedRequests,
	}
}

// mapCircuitBreakerState maps circuit breaker state to health check status
func mapCircuitBreakerState(state resilience.CircuitState) string {
	switch state {
	case resilience.CircuitClosed:
		return "pass"
	case resilience.CircuitOpen:
		return "fail"
	case resilience.CircuitHalfOpen:
		return "warn"
	default:
		return "unknown"
	}
}

// HealthCheckWithCircuitBreaker performs a health check including circuit breaker status
// Returns HealthStatus which includes circuit breaker information
func (pms *PythonMLService) HealthCheckWithCircuitBreaker(ctx context.Context) (*HealthStatus, error) {
	healthCheck, err := pms.HealthCheck(ctx)
	if err != nil {
		return nil, err
	}

	cbState := pms.circuitBreaker.GetState()
	cbStats := pms.circuitBreaker.GetStats()
	
	// Create health status with circuit breaker information
	healthStatus := &HealthStatus{
		Status:    healthCheck.Status,
		LastCheck: healthCheck.LastCheck,
		Checks:    make(map[string]HealthCheck),
		Message:   healthCheck.Message,
	}
	
	// Add service health check
	healthStatus.Checks["python_ml_service"] = *healthCheck
	
	// Add circuit breaker check
	healthStatus.Checks["circuit_breaker"] = HealthCheck{
		Name:      "circuit_breaker",
		Status:    mapCircuitBreakerState(cbState),
		Message:   fmt.Sprintf("Circuit breaker state: %s (failures: %d, successes: %d)", cbState.String(), cbStats.FailureCount, cbStats.SuccessCount),
		LastCheck: time.Now(),
		Duration:  0,
	}

	return healthStatus, nil
}
