package routing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kyb-platform/internal/architecture"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/shared"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// IntelligentRouter provides intelligent routing for classification requests
type IntelligentRouter struct {
	logger          *observability.Logger
	tracer          trace.Tracer
	metrics         *observability.Metrics
	config          IntelligentRouterConfig
	requestAnalyzer *RequestAnalyzer
	moduleSelector  *ModuleSelector
	moduleManager   ModuleManager
	resourceManager *ResourceManager

	// Request tracking
	activeRequests map[string]*RequestContext
	requestMutex   sync.RWMutex

	// Performance tracking
	routerMetrics *RouterMetrics
	metricsMutex  sync.RWMutex

	// Parallel processing components
	requestSemaphore chan struct{} // Controls concurrent request processing
	workerPool       chan struct{} // Controls worker goroutines
	parallelMutex    sync.RWMutex
}

// IntelligentRouterConfig holds configuration for the intelligent router
type IntelligentRouterConfig struct {
	EnableRequestAnalysis    bool                   `json:"enable_request_analysis"`
	EnableModuleSelection    bool                   `json:"enable_module_selection"`
	EnableParallelProcessing bool                   `json:"enable_parallel_processing"`
	EnableRetryLogic         bool                   `json:"enable_retry_logic"`
	EnableFallbackProcessing bool                   `json:"enable_fallback_processing"`
	MaxConcurrentRequests    int                    `json:"max_concurrent_requests"`
	MaxParallelModules       int                    `json:"max_parallel_modules"`
	WorkerPoolSize           int                    `json:"worker_pool_size"`
	RequestTimeout           time.Duration          `json:"request_timeout"`
	RetryAttempts            int                    `json:"retry_attempts"`
	RetryDelay               time.Duration          `json:"retry_delay"`
	FallbackTimeout          time.Duration          `json:"fallback_timeout"`
	EnableMetricsCollection  bool                   `json:"enable_metrics_collection"`
	ParallelProcessingMode   ParallelProcessingMode `json:"parallel_processing_mode"`
}

// ParallelProcessingMode defines the mode of parallel processing
type ParallelProcessingMode string

const (
	ParallelProcessingModeSequential ParallelProcessingMode = "sequential"
	ParallelProcessingModeConcurrent ParallelProcessingMode = "concurrent"
	ParallelProcessingModeHybrid     ParallelProcessingMode = "hybrid"
)

// ParallelProcessingResult represents the result of parallel processing
type ParallelProcessingResult struct {
	Results        []*ModuleProcessingResult `json:"results"`
	BestResult     *ModuleProcessingResult   `json:"best_result"`
	ProcessingTime time.Duration             `json:"processing_time"`
	SuccessCount   int                       `json:"success_count"`
	FailureCount   int                       `json:"failure_count"`
	Metadata       map[string]interface{}    `json:"metadata"`
}

// ModuleProcessingResult represents the result of processing with a specific module
type ModuleProcessingResult struct {
	ModuleID       string                                 `json:"module_id"`
	ModuleType     string                                 `json:"module_type"`
	Success        bool                                   `json:"success"`
	Classification *shared.BusinessClassificationResponse `json:"classification,omitempty"`
	ProcessingTime time.Duration                          `json:"processing_time"`
	Confidence     float64                                `json:"confidence"`
	Error          string                                 `json:"error,omitempty"`
	AttemptNumber  int                                    `json:"attempt_number"`
	Metadata       map[string]interface{}                 `json:"metadata"`
}

// RequestContext tracks the context of a routing request
type RequestContext struct {
	RequestID        string                                `json:"request_id"`
	OriginalRequest  *shared.BusinessClassificationRequest `json:"original_request"`
	AnalysisResult   *RequestAnalysisResult                `json:"analysis_result"`
	SelectionResult  *SelectionResult                      `json:"selection_result"`
	ProcessingResult *ProcessingResult                     `json:"processing_result"`
	StartTime        time.Time                             `json:"start_time"`
	EndTime          time.Time                             `json:"end_time"`
	Status           RequestStatus                         `json:"status"`
	Error            error                                 `json:"error,omitempty"`
	RetryCount       int                                   `json:"retry_count"`
	Attempts         []ProcessingAttempt                   `json:"attempts"`
}

// RequestStatus represents the status of a routing request
type RequestStatus string

const (
	RequestStatusPending    RequestStatus = "pending"
	RequestStatusAnalyzing  RequestStatus = "analyzing"
	RequestStatusSelecting  RequestStatus = "selecting"
	RequestStatusProcessing RequestStatus = "processing"
	RequestStatusCompleted  RequestStatus = "completed"
	RequestStatusFailed     RequestStatus = "failed"
	RequestStatusRetrying   RequestStatus = "retrying"
)

// ProcessingResult represents the result of processing a request
type ProcessingResult struct {
	Success        bool                                   `json:"success"`
	ModuleID       string                                 `json:"module_id"`
	ModuleType     string                                 `json:"module_type"`
	Classification *shared.BusinessClassificationResponse `json:"classification,omitempty"`
	ProcessingTime time.Duration                          `json:"processing_time"`
	Confidence     float64                                `json:"confidence"`
	FallbackUsed   bool                                   `json:"fallback_used"`
	RetryCount     int                                    `json:"retry_count"`
	Error          string                                 `json:"error,omitempty"`
	Metadata       map[string]interface{}                 `json:"metadata"`
}

// ProcessingAttempt represents a single processing attempt
type ProcessingAttempt struct {
	AttemptNumber  int                                    `json:"attempt_number"`
	ModuleID       string                                 `json:"module_id"`
	ModuleType     string                                 `json:"module_type"`
	StartTime      time.Time                              `json:"start_time"`
	EndTime        time.Time                              `json:"end_time"`
	Success        bool                                   `json:"success"`
	Result         *shared.BusinessClassificationResponse `json:"result,omitempty"`
	Error          string                                 `json:"error,omitempty"`
	ProcessingTime time.Duration                          `json:"processing_time"`
	Confidence     float64                                `json:"confidence"`
}

// RouterMetrics tracks router performance metrics
type RouterMetrics struct {
	TotalRequests         int64         `json:"total_requests"`
	SuccessfulRequests    int64         `json:"successful_requests"`
	FailedRequests        int64         `json:"failed_requests"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	AverageAnalysisTime   time.Duration `json:"average_analysis_time"`
	AverageSelectionTime  time.Duration `json:"average_selection_time"`
	RetryCount            int64         `json:"retry_count"`
	FallbackCount         int64         `json:"fallback_count"`
	LastUpdated           time.Time     `json:"last_updated"`
}

// NewIntelligentRouter creates a new intelligent router
func NewIntelligentRouter(
	logger *observability.Logger,
	tracer trace.Tracer,
	metrics *observability.Metrics,
	config IntelligentRouterConfig,
	requestAnalyzer *RequestAnalyzer,
	moduleSelector *ModuleSelector,
	moduleManager ModuleManager,
) *IntelligentRouter {
	// Set default configuration if not provided
	if config.MaxConcurrentRequests == 0 {
		config.MaxConcurrentRequests = 100
	}
	if config.MaxParallelModules == 0 {
		config.MaxParallelModules = 5
	}
	if config.WorkerPoolSize == 0 {
		config.WorkerPoolSize = 20
	}
	if config.RequestTimeout == 0 {
		config.RequestTimeout = 60 * time.Second
	}
	if config.RetryAttempts == 0 {
		config.RetryAttempts = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.FallbackTimeout == 0 {
		config.FallbackTimeout = 30 * time.Second
	}
	if config.ParallelProcessingMode == "" {
		config.ParallelProcessingMode = ParallelProcessingModeHybrid
	}

	return &IntelligentRouter{
		logger:          logger,
		tracer:          tracer,
		metrics:         metrics,
		config:          config,
		requestAnalyzer: requestAnalyzer,
		moduleSelector:  moduleSelector,
		moduleManager:   moduleManager,
		resourceManager: NewResourceManager(
			logger,
			tracer,
			metrics,
			ResourceManagerConfig{
				EnableLoadBalancing:      config.EnableRequestAnalysis,
				EnableResourceMonitoring: config.EnableMetricsCollection,
				EnableHealthMonitoring:   config.EnableMetricsCollection,
				EnableCapacityPlanning:   config.EnableMetricsCollection,
				LoadBalancingStrategy:    LoadBalancingStrategyAdaptive,
				ResourceUpdateInterval:   30 * time.Second,
				HealthCheckInterval:      60 * time.Second,
				CapacityPlanningInterval: 300 * time.Second,
				MaxResourceUtilization:   0.8,
				MinResourceUtilization:   0.2,
				ScalingThreshold:         0.7,
				HealthCheckTimeout:       10 * time.Second,
			},
		),

		activeRequests: make(map[string]*RequestContext),
		routerMetrics: &RouterMetrics{
			LastUpdated: time.Now(),
		},

		// Initialize parallel processing components
		requestSemaphore: make(chan struct{}, config.MaxConcurrentRequests),
		workerPool:       make(chan struct{}, config.WorkerPoolSize),
	}
}

// RouteRequest routes a classification request through the intelligent routing system
func (ir *IntelligentRouter) RouteRequest(
	ctx context.Context,
	req *shared.BusinessClassificationRequest,
) (*shared.BusinessClassificationResponse, error) {
	ctx, span := ir.tracer.Start(ctx, "IntelligentRouter.RouteRequest")
	defer span.End()

	span.SetAttributes(
		attribute.String("request.id", req.ID),
		attribute.String("business.name", req.BusinessName),
	)

	// Acquire request semaphore for concurrent request control
	select {
	case ir.requestSemaphore <- struct{}{}:
		defer func() { <-ir.requestSemaphore }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Create request context
	requestContext := &RequestContext{
		RequestID:       req.ID,
		OriginalRequest: req,
		StartTime:       time.Now(),
		Status:          RequestStatusPending,
		Attempts:        make([]ProcessingAttempt, 0),
	}

	// Track active request
	ir.trackActiveRequest(requestContext)
	defer ir.untrackActiveRequest(req.ID)

	// Step 1: Analyze request
	analysisResult, err := ir.analyzeRequest(ctx, requestContext)
	if err != nil {
		ir.updateRequestStatus(requestContext, RequestStatusFailed, err)
		ir.recordMetrics(requestContext, false)
		return nil, fmt.Errorf("request analysis failed: %w", err)
	}

	// Step 2: Select module
	selectionResult, err := ir.selectModule(ctx, requestContext, analysisResult)
	if err != nil {
		ir.updateRequestStatus(requestContext, RequestStatusFailed, err)
		ir.recordMetrics(requestContext, false)
		return nil, fmt.Errorf("module selection failed: %w", err)
	}

	// Step 3: Process request with parallel processing capabilities
	var processingResult *ProcessingResult
	if ir.config.EnableParallelProcessing {
		processingResult, err = ir.processRequestParallel(ctx, requestContext, analysisResult, selectionResult)
	} else {
		processingResult, err = ir.processRequest(ctx, requestContext, analysisResult, selectionResult)
	}

	if err != nil {
		ir.updateRequestStatus(requestContext, RequestStatusFailed, err)
		ir.recordMetrics(requestContext, false)
		return nil, fmt.Errorf("request processing failed: %w", err)
	}

	// Update final status
	ir.updateRequestStatus(requestContext, RequestStatusCompleted, nil)
	ir.recordMetrics(requestContext, true)

	// Log successful routing
	ir.logger.WithComponent("intelligent_router").Info("request_routing_completed", map[string]interface{}{
		"request_id":         req.ID,
		"selected_module":    processingResult.ModuleID,
		"module_type":        processingResult.ModuleType,
		"processing_time_ms": processingResult.ProcessingTime.Milliseconds(),
		"confidence":         processingResult.Confidence,
		"fallback_used":      processingResult.FallbackUsed,
		"retry_count":        processingResult.RetryCount,
		"parallel_mode":      ir.config.ParallelProcessingMode,
	})

	return processingResult.Classification, nil
}

// analyzeRequest analyzes the classification request
func (ir *IntelligentRouter) analyzeRequest(
	ctx context.Context,
	requestContext *RequestContext,
) (*RequestAnalysisResult, error) {
	ctx, span := ir.tracer.Start(ctx, "IntelligentRouter.analyzeRequest")
	defer span.End()

	ir.updateRequestStatus(requestContext, RequestStatusAnalyzing, nil)

	startTime := time.Now()

	// Perform request analysis
	analysisResult, err := ir.requestAnalyzer.AnalyzeRequest(ctx, requestContext.OriginalRequest)
	if err != nil {
		return nil, err
	}

	// Store analysis result
	requestContext.AnalysisResult = analysisResult

	// Record analysis time
	analysisTime := time.Since(startTime)
	ir.updateAnalysisMetrics(analysisTime)

	span.SetAttributes(
		attribute.String("request.type", string(analysisResult.RequestType)),
		attribute.String("complexity", string(analysisResult.Complexity)),
		attribute.String("priority", string(analysisResult.Priority)),
		attribute.Int64("analysis_time_ms", analysisTime.Milliseconds()),
	)

	return analysisResult, nil
}

// selectModule selects the appropriate module for processing
func (ir *IntelligentRouter) selectModule(
	ctx context.Context,
	requestContext *RequestContext,
	analysisResult *RequestAnalysisResult,
) (*SelectionResult, error) {
	ctx, span := ir.tracer.Start(ctx, "IntelligentRouter.selectModule")
	defer span.End()

	ir.updateRequestStatus(requestContext, RequestStatusSelecting, nil)

	startTime := time.Now()

	// Select module
	selectionResult, err := ir.moduleSelector.SelectModule(ctx, requestContext.OriginalRequest, analysisResult)
	if err != nil {
		return nil, err
	}

	// Store selection result
	requestContext.SelectionResult = selectionResult

	// Record selection time
	selectionTime := time.Since(startTime)
	ir.updateSelectionMetrics(selectionTime)

	span.SetAttributes(
		attribute.String("selected_module", selectionResult.SelectedModule.ModuleID),
		attribute.String("module_type", selectionResult.SelectedModule.ModuleType),
		attribute.Float64("confidence", selectionResult.Confidence),
		attribute.Int64("selection_time_ms", selectionTime.Milliseconds()),
	)

	return selectionResult, nil
}

// processRequest processes the request using the selected module
func (ir *IntelligentRouter) processRequest(
	ctx context.Context,
	requestContext *RequestContext,
	analysisResult *RequestAnalysisResult,
	selectionResult *SelectionResult,
) (*ProcessingResult, error) {
	ctx, span := ir.tracer.Start(ctx, "IntelligentRouter.processRequest")
	defer span.End()

	ir.updateRequestStatus(requestContext, RequestStatusProcessing, nil)

	// Create processing result
	processingResult := &ProcessingResult{
		ModuleID:   selectionResult.SelectedModule.ModuleID,
		ModuleType: selectionResult.SelectedModule.ModuleType,
		Metadata:   make(map[string]interface{}),
	}

	// Process with primary module
	result, err := ir.processWithModule(ctx, requestContext, selectionResult.SelectedModule, 1)
	if err == nil {
		processingResult.Success = true
		processingResult.Classification = result
		processingResult.Confidence = selectionResult.Confidence
		processingResult.ProcessingTime = time.Since(requestContext.StartTime)
		requestContext.ProcessingResult = processingResult
		return processingResult, nil
	}

	// If primary module fails and retry logic is enabled
	if ir.config.EnableRetryLogic {
		for attempt := 2; attempt <= ir.config.RetryAttempts; attempt++ {
			ir.updateRequestStatus(requestContext, RequestStatusRetrying, nil)

			// Wait before retry
			time.Sleep(ir.config.RetryDelay)

			result, err := ir.processWithModule(ctx, requestContext, selectionResult.SelectedModule, attempt)
			if err == nil {
				processingResult.Success = true
				processingResult.Classification = result
				processingResult.Confidence = selectionResult.Confidence
				processingResult.RetryCount = attempt - 1
				processingResult.ProcessingTime = time.Since(requestContext.StartTime)
				requestContext.ProcessingResult = processingResult
				return processingResult, nil
			}
		}
	}

	// If fallback processing is enabled, try fallback modules
	if ir.config.EnableFallbackProcessing && len(selectionResult.FallbackModules) > 0 {
		for _, fallbackModule := range selectionResult.FallbackModules {
			result, err := ir.processWithModule(ctx, requestContext, fallbackModule, 1)
			if err == nil {
				processingResult.Success = true
				processingResult.Classification = result
				processingResult.ModuleID = fallbackModule.ModuleID
				processingResult.ModuleType = fallbackModule.ModuleType
				processingResult.Confidence = selectionResult.Confidence * 0.8 // Reduce confidence for fallback
				processingResult.FallbackUsed = true
				processingResult.RetryCount = ir.config.RetryAttempts
				processingResult.ProcessingTime = time.Since(requestContext.StartTime)
				requestContext.ProcessingResult = processingResult
				return processingResult, nil
			}
		}
	}

	// All attempts failed
	processingResult.Success = false
	processingResult.Error = err.Error()
	processingResult.ProcessingTime = time.Since(requestContext.StartTime)
	requestContext.ProcessingResult = processingResult

	return processingResult, fmt.Errorf("all processing attempts failed: %w", err)
}

// convertRequestToModuleData converts a business classification request to module data
func (ir *IntelligentRouter) convertRequestToModuleData(req *shared.BusinessClassificationRequest) map[string]interface{} {
	return map[string]interface{}{
		"business_name":     req.BusinessName,
		"website_url":       req.WebsiteURL,
		"description":       req.Description,
		"keywords":          req.Keywords,
		"industry":          req.Industry,
		"geographic_region": req.GeographicRegion,
		"metadata":          req.Metadata,
	}
}

// convertModuleResponseToClassification converts a module response to classification response
func (ir *IntelligentRouter) convertModuleResponseToClassification(moduleResponse architecture.ModuleResponse) (*shared.BusinessClassificationResponse, error) {
	if !moduleResponse.Success {
		return nil, fmt.Errorf("module processing failed: %s", moduleResponse.Error)
	}

	// Extract classification data from module response
	classificationData, ok := moduleResponse.Data["classification"]
	if !ok {
		return nil, fmt.Errorf("no classification data in module response")
	}

	// Convert to classification response
	// This is a simplified conversion - in practice, you'd want more robust type checking
	classificationMap, ok := classificationData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid classification data format")
	}

	// Create basic classification response
	response := &shared.BusinessClassificationResponse{
		ID:                moduleResponse.ID,
		OverallConfidence: moduleResponse.Confidence,
		ProcessingTime:    moduleResponse.Latency,
		Classifications:   make([]shared.IndustryClassification, 0),
		Metadata:          moduleResponse.Metadata,
	}

	// Extract classifications if available
	if classificationsData, ok := classificationMap["classifications"]; ok {
		if classificationsList, ok := classificationsData.([]interface{}); ok {
			for _, item := range classificationsList {
				if classificationMap, ok := item.(map[string]interface{}); ok {
					classification := shared.IndustryClassification{
						IndustryCode:         ir.extractString(classificationMap, "industry_code"),
						IndustryName:         ir.extractString(classificationMap, "industry_name"),
						ConfidenceScore:      ir.extractFloat64(classificationMap, "confidence_score"),
						ClassificationMethod: ir.extractString(classificationMap, "classification_method"),
						Description:          ir.extractString(classificationMap, "description"),
						Evidence:             ir.extractString(classificationMap, "evidence"),
					}
					response.Classifications = append(response.Classifications, classification)
				}
			}
		}
	}

	// Set primary classification if available
	if len(response.Classifications) > 0 {
		response.PrimaryClassification = &response.Classifications[0]
	}

	return response, nil
}

// extractString safely extracts a string value from a map
func (ir *IntelligentRouter) extractString(data map[string]interface{}, key string) string {
	if value, ok := data[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// extractFloat64 safely extracts a float64 value from a map
func (ir *IntelligentRouter) extractFloat64(data map[string]interface{}, key string) float64 {
	if value, ok := data[key]; ok {
		if f, ok := value.(float64); ok {
			return f
		}
	}
	return 0.0
}

// processRequestParallel processes the request using parallel processing capabilities
func (ir *IntelligentRouter) processRequestParallel(
	ctx context.Context,
	requestContext *RequestContext,
	analysisResult *RequestAnalysisResult,
	selectionResult *SelectionResult,
) (*ProcessingResult, error) {
	ctx, span := ir.tracer.Start(ctx, "IntelligentRouter.processRequestParallel")
	defer span.End()

	ir.updateRequestStatus(requestContext, RequestStatusProcessing, nil)

	// Determine processing strategy based on mode
	switch ir.config.ParallelProcessingMode {
	case ParallelProcessingModeConcurrent:
		return ir.processConcurrent(ctx, requestContext, analysisResult, selectionResult)
	case ParallelProcessingModeHybrid:
		return ir.processHybrid(ctx, requestContext, analysisResult, selectionResult)
	default:
		// Fallback to sequential processing
		return ir.processRequest(ctx, requestContext, analysisResult, selectionResult)
	}
}

// processConcurrent processes the request using concurrent module execution
func (ir *IntelligentRouter) processConcurrent(
	ctx context.Context,
	requestContext *RequestContext,
	analysisResult *RequestAnalysisResult,
	selectionResult *SelectionResult,
) (*ProcessingResult, error) {
	ctx, span := ir.tracer.Start(ctx, "IntelligentRouter.processConcurrent")
	defer span.End()

	// Collect all modules to process (primary + fallbacks)
	modulesToProcess := []*ModuleInfo{selectionResult.SelectedModule}
	modulesToProcess = append(modulesToProcess, selectionResult.FallbackModules...)

	// Limit the number of parallel modules
	if len(modulesToProcess) > ir.config.MaxParallelModules {
		modulesToProcess = modulesToProcess[:ir.config.MaxParallelModules]
	}

	// Create channels for results and errors
	resultChan := make(chan *ModuleProcessingResult, len(modulesToProcess))
	errorChan := make(chan error, len(modulesToProcess))
	doneChan := make(chan struct{})

	// Process modules concurrently
	var wg sync.WaitGroup
	for i, moduleInfo := range modulesToProcess {
		wg.Add(1)
		go func(module *ModuleInfo, attempt int) {
			defer wg.Done()

			// Acquire worker from pool
			select {
			case ir.workerPool <- struct{}{}:
				defer func() { <-ir.workerPool }()
			case <-ctx.Done():
				errorChan <- ctx.Err()
				return
			}

			// Process with module
			result := ir.processModuleParallel(ctx, requestContext, module, attempt)
			if result.Success {
				resultChan <- result
			} else {
				errorChan <- fmt.Errorf("module %s failed: %s", module.ModuleID, result.Error)
			}
		}(moduleInfo, i+1)
	}

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(doneChan)
	}()

	// Wait for first successful result or all failures
	select {
	case result := <-resultChan:
		// Success - return the result
		processingResult := &ProcessingResult{
			Success:        true,
			ModuleID:       result.ModuleID,
			ModuleType:     result.ModuleType,
			Classification: result.Classification,
			Confidence:     result.Confidence,
			ProcessingTime: time.Since(requestContext.StartTime),
			FallbackUsed:   result.ModuleID != selectionResult.SelectedModule.ModuleID,
			Metadata:       result.Metadata,
		}
		requestContext.ProcessingResult = processingResult
		return processingResult, nil

	case <-doneChan:
		// All modules failed
		processingResult := &ProcessingResult{
			Success:        false,
			Error:          "all parallel processing attempts failed",
			ProcessingTime: time.Since(requestContext.StartTime),
		}
		requestContext.ProcessingResult = processingResult
		return processingResult, fmt.Errorf("all parallel processing attempts failed")

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// processHybrid processes the request using hybrid parallel/sequential approach
func (ir *IntelligentRouter) processHybrid(
	ctx context.Context,
	requestContext *RequestContext,
	analysisResult *RequestAnalysisResult,
	selectionResult *SelectionResult,
) (*ProcessingResult, error) {
	ctx, span := ir.tracer.Start(ctx, "IntelligentRouter.processHybrid")
	defer span.End()

	// Try primary module first
	result, err := ir.processWithModule(ctx, requestContext, selectionResult.SelectedModule, 1)
	if err == nil {
		processingResult := &ProcessingResult{
			Success:        true,
			ModuleID:       selectionResult.SelectedModule.ModuleID,
			ModuleType:     selectionResult.SelectedModule.ModuleType,
			Classification: result,
			Confidence:     selectionResult.Confidence,
			ProcessingTime: time.Since(requestContext.StartTime),
			Metadata:       make(map[string]interface{}),
		}
		requestContext.ProcessingResult = processingResult
		return processingResult, nil
	}

	// If primary fails and fallbacks are available, process them in parallel
	if ir.config.EnableFallbackProcessing && len(selectionResult.FallbackModules) > 0 {
		return ir.processConcurrent(ctx, requestContext, analysisResult, selectionResult)
	}

	// No fallbacks or fallbacks disabled - return error
	processingResult := &ProcessingResult{
		Success:        false,
		Error:          err.Error(),
		ProcessingTime: time.Since(requestContext.StartTime),
	}
	requestContext.ProcessingResult = processingResult
	return processingResult, err
}

// processModuleParallel processes a single module in parallel context
func (ir *IntelligentRouter) processModuleParallel(
	ctx context.Context,
	requestContext *RequestContext,
	moduleInfo *ModuleInfo,
	attemptNumber int,
) *ModuleProcessingResult {
	ctx, span := ir.tracer.Start(ctx, "IntelligentRouter.processModuleParallel")
	defer span.End()

	span.SetAttributes(
		attribute.String("module.id", moduleInfo.ModuleID),
		attribute.String("module.type", moduleInfo.ModuleType),
		attribute.Int("attempt", attemptNumber),
	)

	startTime := time.Now()

	// Get module instance
	module, exists := ir.moduleManager.GetModuleByID(moduleInfo.ModuleID)
	if !exists {
		return &ModuleProcessingResult{
			ModuleID:       moduleInfo.ModuleID,
			ModuleType:     moduleInfo.ModuleType,
			Success:        false,
			Error:          fmt.Sprintf("module %s not found", moduleInfo.ModuleID),
			ProcessingTime: time.Since(startTime),
			AttemptNumber:  attemptNumber,
		}
	}

	// Create module request
	moduleRequest := architecture.ModuleRequest{
		ID:       requestContext.RequestID,
		Type:     "classification",
		Data:     ir.convertRequestToModuleData(requestContext.OriginalRequest),
		Priority: architecture.PriorityMedium,
		Timeout:  ir.config.RequestTimeout,
		Context:  ctx,
	}

	// Process with module
	moduleResponse, err := module.Process(ctx, moduleRequest)
	if err != nil {
		return &ModuleProcessingResult{
			ModuleID:       moduleInfo.ModuleID,
			ModuleType:     moduleInfo.ModuleType,
			Success:        false,
			Error:          err.Error(),
			ProcessingTime: time.Since(startTime),
			AttemptNumber:  attemptNumber,
		}
	}

	// Convert module response to classification response
	classificationResponse, err := ir.convertModuleResponseToClassification(moduleResponse)
	if err != nil {
		return &ModuleProcessingResult{
			ModuleID:       moduleInfo.ModuleID,
			ModuleType:     moduleInfo.ModuleType,
			Success:        false,
			Error:          err.Error(),
			ProcessingTime: time.Since(startTime),
			AttemptNumber:  attemptNumber,
		}
	}

	return &ModuleProcessingResult{
		ModuleID:       moduleInfo.ModuleID,
		ModuleType:     moduleInfo.ModuleType,
		Success:        true,
		Classification: classificationResponse,
		ProcessingTime: time.Since(startTime),
		Confidence:     moduleResponse.Confidence,
		AttemptNumber:  attemptNumber,
		Metadata:       moduleResponse.Metadata,
	}
}

// processWithModule processes the request with a specific module
func (ir *IntelligentRouter) processWithModule(
	ctx context.Context,
	requestContext *RequestContext,
	moduleInfo *ModuleInfo,
	attemptNumber int,
) (*shared.BusinessClassificationResponse, error) {
	ctx, span := ir.tracer.Start(ctx, "IntelligentRouter.processWithModule")
	defer span.End()

	span.SetAttributes(
		attribute.String("module.id", moduleInfo.ModuleID),
		attribute.String("module.type", moduleInfo.ModuleType),
		attribute.Int("attempt", attemptNumber),
	)

	// Get module instance
	module, exists := ir.moduleManager.GetModuleByID(moduleInfo.ModuleID)
	if !exists {
		return nil, fmt.Errorf("module %s not found", moduleInfo.ModuleID)
	}

	// Create processing attempt
	attempt := ProcessingAttempt{
		AttemptNumber: attemptNumber,
		ModuleID:      moduleInfo.ModuleID,
		ModuleType:    moduleInfo.ModuleType,
		StartTime:     time.Now(),
	}

	// Convert request to module request format
	moduleRequest := ir.convertToModuleRequest(requestContext.OriginalRequest)

	// Process with module
	startTime := time.Now()
	moduleResponse, err := module.Process(ctx, moduleRequest)
	processingTime := time.Since(startTime)

	attempt.EndTime = time.Now()
	attempt.ProcessingTime = processingTime

	if err != nil {
		attempt.Success = false
		attempt.Error = err.Error()
		requestContext.Attempts = append(requestContext.Attempts, attempt)

		// Update module performance
		ir.moduleSelector.UpdateModulePerformance(moduleInfo.ModuleID, false, processingTime)

		return nil, err
	}

	// Convert module response to classification response
	classificationResponse, err := ir.convertFromModuleResponse(moduleResponse)
	if err != nil {
		attempt.Success = false
		attempt.Error = err.Error()
		requestContext.Attempts = append(requestContext.Attempts, attempt)

		// Update module performance
		ir.moduleSelector.UpdateModulePerformance(moduleInfo.ModuleID, false, processingTime)

		return nil, err
	}

	attempt.Success = true
	attempt.Result = classificationResponse
	attempt.Confidence = moduleResponse.Confidence
	requestContext.Attempts = append(requestContext.Attempts, attempt)

	// Update module performance
	ir.moduleSelector.UpdateModulePerformance(moduleInfo.ModuleID, true, processingTime)

	return classificationResponse, nil
}

// convertToModuleRequest converts a business classification request to a module request
func (ir *IntelligentRouter) convertToModuleRequest(req *shared.BusinessClassificationRequest) architecture.ModuleRequest {
	return architecture.ModuleRequest{
		ID:   req.ID,
		Type: "business_classification",
		Data: map[string]interface{}{
			"business_name":     req.BusinessName,
			"website_url":       req.WebsiteURL,
			"keywords":          req.Keywords,
			"description":       req.Description,
			"industry":          req.Industry,
			"geographic_region": req.GeographicRegion,
			"metadata":          req.Metadata,
		},
		Priority: architecture.PriorityMedium, // Default priority since BusinessClassificationRequest doesn't have Priority field
		Timeout:  ir.config.RequestTimeout,
		Context:  context.Background(),
	}
}

// convertFromModuleResponse converts a module response to a business classification response
func (ir *IntelligentRouter) convertFromModuleResponse(response architecture.ModuleResponse) (*shared.BusinessClassificationResponse, error) {
	if !response.Success {
		return nil, fmt.Errorf("module processing failed: %s", response.Error)
	}

	// Extract classification data from response
	classificationData, ok := response.Data["classification"]
	if !ok {
		return nil, fmt.Errorf("no classification data in module response")
	}

	// Convert to classification response
	// This is a simplified conversion - in a real implementation, you'd have more robust conversion logic
	classificationResponse := &shared.BusinessClassificationResponse{
		ID:                response.ID,
		BusinessName:      "", // Will be set from original request
		OverallConfidence: response.Confidence,
		ProcessingTime:    response.Latency,
		Metadata:          response.Metadata,
		CreatedAt:         time.Now(),
	}

	// Extract specific classification fields
	if data, ok := classificationData.(map[string]interface{}); ok {
		// Create industry classification from module data
		if industry, ok := data["industry"].(string); ok {
			industryClassification := &shared.IndustryClassification{
				IndustryName:         industry,
				ConfidenceScore:      response.Confidence,
				ClassificationMethod: "module_classification",
				ProcessingTime:       response.Latency,
			}
			classificationResponse.Classifications = append(classificationResponse.Classifications, *industryClassification)
			classificationResponse.PrimaryClassification = industryClassification
		}
		// Add more field conversions as needed
	}

	return classificationResponse, nil
}

// trackActiveRequest tracks an active request
func (ir *IntelligentRouter) trackActiveRequest(requestContext *RequestContext) {
	ir.requestMutex.Lock()
	defer ir.requestMutex.Unlock()
	ir.activeRequests[requestContext.RequestID] = requestContext
}

// untrackActiveRequest removes an active request from tracking
func (ir *IntelligentRouter) untrackActiveRequest(requestID string) {
	ir.requestMutex.Lock()
	defer ir.requestMutex.Unlock()
	delete(ir.activeRequests, requestID)
}

// updateRequestStatus updates the status of a request
func (ir *IntelligentRouter) updateRequestStatus(requestContext *RequestContext, status RequestStatus, err error) {
	requestContext.Status = status
	if err != nil {
		requestContext.Error = err
	}
	if status == RequestStatusCompleted || status == RequestStatusFailed {
		requestContext.EndTime = time.Now()
	}
}

// recordMetrics records metrics for a completed request
func (ir *IntelligentRouter) recordMetrics(requestContext *RequestContext, success bool) {
	ir.metricsMutex.Lock()
	defer ir.metricsMutex.Unlock()

	ir.routerMetrics.TotalRequests++
	if success {
		ir.routerMetrics.SuccessfulRequests++
	} else {
		ir.routerMetrics.FailedRequests++
	}

	if requestContext.ProcessingResult != nil {
		// Update average processing time
		if ir.routerMetrics.TotalRequests == 1 {
			ir.routerMetrics.AverageProcessingTime = requestContext.ProcessingResult.ProcessingTime
		} else {
			// Exponential moving average
			alpha := 0.1
			ir.routerMetrics.AverageProcessingTime = time.Duration(
				float64(ir.routerMetrics.AverageProcessingTime)*(1-alpha) + float64(requestContext.ProcessingResult.ProcessingTime)*alpha,
			)
		}

		ir.routerMetrics.RetryCount += int64(requestContext.ProcessingResult.RetryCount)
		if requestContext.ProcessingResult.FallbackUsed {
			ir.routerMetrics.FallbackCount++
		}
	}

	ir.routerMetrics.LastUpdated = time.Now()
}

// updateAnalysisMetrics updates analysis time metrics
func (ir *IntelligentRouter) updateAnalysisMetrics(analysisTime time.Duration) {
	ir.metricsMutex.Lock()
	defer ir.metricsMutex.Unlock()

	if ir.routerMetrics.TotalRequests == 1 {
		ir.routerMetrics.AverageAnalysisTime = analysisTime
	} else {
		alpha := 0.1
		ir.routerMetrics.AverageAnalysisTime = time.Duration(
			float64(ir.routerMetrics.AverageAnalysisTime)*(1-alpha) + float64(analysisTime)*alpha,
		)
	}
}

// updateSelectionMetrics updates selection time metrics
func (ir *IntelligentRouter) updateSelectionMetrics(selectionTime time.Duration) {
	ir.metricsMutex.Lock()
	defer ir.metricsMutex.Unlock()

	if ir.routerMetrics.TotalRequests == 1 {
		ir.routerMetrics.AverageSelectionTime = selectionTime
	} else {
		alpha := 0.1
		ir.routerMetrics.AverageSelectionTime = time.Duration(
			float64(ir.routerMetrics.AverageSelectionTime)*(1-alpha) + float64(selectionTime)*alpha,
		)
	}
}

// GetActiveRequests returns the currently active requests
func (ir *IntelligentRouter) GetActiveRequests() map[string]*RequestContext {
	ir.requestMutex.RLock()
	defer ir.requestMutex.RUnlock()

	activeRequests := make(map[string]*RequestContext)
	for id, context := range ir.activeRequests {
		activeRequests[id] = context
	}

	return activeRequests
}

// GetRouterMetrics returns the router performance metrics
func (ir *IntelligentRouter) GetRouterMetrics() *RouterMetrics {
	ir.metricsMutex.RLock()
	defer ir.metricsMutex.RUnlock()

	// Return a copy to avoid race conditions
	metrics := *ir.routerMetrics
	return &metrics
}

// GetRequestContext returns the context for a specific request
func (ir *IntelligentRouter) GetRequestContext(requestID string) (*RequestContext, bool) {
	ir.requestMutex.RLock()
	defer ir.requestMutex.RUnlock()

	context, exists := ir.activeRequests[requestID]
	return context, exists
}
