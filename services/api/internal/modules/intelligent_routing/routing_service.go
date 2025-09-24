package intelligent_routing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// RoutingServiceConfig represents the configuration for the routing service
type RoutingServiceConfig struct {
	DefaultStrategy      RoutingStrategy `json:"default_strategy"`
	LoadBalancingEnabled bool            `json:"load_balancing_enabled"`
	ParallelProcessing   bool            `json:"parallel_processing"`
	FallbackEnabled      bool            `json:"fallback_enabled"`
	HealthCheckInterval  time.Duration   `json:"health_check_interval"`
	DecisionTimeout      time.Duration   `json:"decision_timeout"`
	MaxRetries           int             `json:"max_retries"`
	CacheEnabled         bool            `json:"cache_enabled"`
	CacheTTL             time.Duration   `json:"cache_ttl"`
}

// routingService implements the RoutingService interface
type routingService struct {
	config *RoutingServiceConfig
	logger *zap.Logger

	// Core components
	requestAnalyzer  RequestAnalyzer
	moduleSelector   ModuleSelector
	loadBalancer     LoadBalancer
	healthChecker    HealthChecker
	metricsCollector MetricsCollector

	// Module registry
	modules      map[string]*ModuleCapability
	modulesMutex sync.RWMutex

	// Cache for routing decisions
	decisionCache map[string]*CachedDecision
	cacheMutex    sync.RWMutex

	// Fallback strategies
	fallbackStrategies []*FallbackStrategy
	fallbackMutex      sync.RWMutex

	// Processing coordination
	processingQueue chan *ProcessingRequest
	workerPool      chan struct{}
}

// CachedDecision represents a cached routing decision
type CachedDecision struct {
	Decision    *RoutingDecision
	CreatedAt   time.Time
	ExpiresAt   time.Time
	AccessCount int64
}

// ProcessingRequest represents a request being processed
type ProcessingRequest struct {
	Request         *VerificationRequest
	Analysis        *RequestAnalysis
	Decision        *RoutingDecision
	SelectedModules []*ModuleCapability
	Context         context.Context
	ResultChan      chan *ProcessingResult
	ErrorChan       chan error
}

// NewRoutingService creates a new routing service instance
func NewRoutingService(
	config *RoutingServiceConfig,
	requestAnalyzer RequestAnalyzer,
	moduleSelector ModuleSelector,
	loadBalancer LoadBalancer,
	healthChecker HealthChecker,
	metricsCollector MetricsCollector,
	logger *zap.Logger,
) RoutingService {
	if config == nil {
		config = &RoutingServiceConfig{
			DefaultStrategy:      StrategyOptimized,
			LoadBalancingEnabled: true,
			ParallelProcessing:   true,
			FallbackEnabled:      true,
			HealthCheckInterval:  30 * time.Second,
			DecisionTimeout:      30 * time.Second,
			MaxRetries:           3,
			CacheEnabled:         true,
			CacheTTL:             5 * time.Minute,
		}
	}

	service := &routingService{
		config:             config,
		logger:             logger,
		requestAnalyzer:    requestAnalyzer,
		moduleSelector:     moduleSelector,
		loadBalancer:       loadBalancer,
		healthChecker:      healthChecker,
		metricsCollector:   metricsCollector,
		modules:            make(map[string]*ModuleCapability),
		decisionCache:      make(map[string]*CachedDecision),
		fallbackStrategies: []*FallbackStrategy{},
		processingQueue:    make(chan *ProcessingRequest, 100),
		workerPool:         make(chan struct{}, 10), // 10 concurrent workers
	}

	// Start background workers
	go service.startBackgroundWorkers()

	return service
}

// RouteRequest routes a verification request to appropriate modules
func (rs *routingService) RouteRequest(ctx context.Context, request *VerificationRequest) (*RoutingDecision, error) {
	rs.logger.Info("Starting request routing",
		zap.String("request_id", request.ID),
		zap.String("business_name", request.BusinessName))

	// Check cache first
	if rs.config.CacheEnabled {
		if cached := rs.getCachedDecision(request); cached != nil {
			rs.logger.Debug("Using cached routing decision",
				zap.String("request_id", request.ID))
			return cached, nil
		}
	}

	// Create routing context with timeout
	routingCtx, cancel := context.WithTimeout(ctx, rs.config.DecisionTimeout)
	defer cancel()

	// Analyze request
	analysis, err := rs.requestAnalyzer.AnalyzeRequest(routingCtx, request)
	if err != nil {
		return nil, fmt.Errorf("request analysis failed: %w", err)
	}

	// Select modules
	selectedModules, err := rs.moduleSelector.SelectModules(routingCtx, request, analysis)
	if err != nil {
		return nil, fmt.Errorf("module selection failed: %w", err)
	}

	// Create routing decision
	decision := &RoutingDecision{
		ID:              generateDecisionID(),
		RequestID:       request.ID,
		SelectedModules: make([]string, len(selectedModules)),
		DecisionReason:  rs.generateDecisionReason(analysis, selectedModules),
		Confidence:      analysis.Confidence,
		RoutingStrategy: rs.determineRoutingStrategy(analysis, selectedModules),
		CreatedAt:       time.Now(),
		Metadata: map[string]interface{}{
			"analysis_id":    analysis.AnalysisID,
			"complexity":     analysis.Complexity,
			"priority":       analysis.Priority,
			"selected_count": len(selectedModules),
		},
	}

	// Extract module IDs
	for i, module := range selectedModules {
		decision.SelectedModules[i] = module.ModuleID
	}

	// Cache the decision
	if rs.config.CacheEnabled {
		rs.cacheDecision(request, decision)
	}

	// Record metrics
	if rs.metricsCollector != nil {
		if err := rs.metricsCollector.RecordRoutingDecision(ctx, decision); err != nil {
			rs.logger.Warn("Failed to record routing metrics", zap.Error(err))
		}
	}

	rs.logger.Info("Routing decision completed",
		zap.String("request_id", request.ID),
		zap.String("strategy", string(decision.RoutingStrategy)),
		zap.Float64("confidence", decision.Confidence),
		zap.Int("modules_selected", len(selectedModules)))

	return decision, nil
}

// ProcessRequest processes a verification request through selected modules
func (rs *routingService) ProcessRequest(ctx context.Context, request *VerificationRequest) ([]*ProcessingResult, error) {
	rs.logger.Info("Starting request processing",
		zap.String("request_id", request.ID))

	// Route the request
	decision, err := rs.RouteRequest(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("routing failed: %w", err)
	}

	// Get selected modules
	selectedModules := rs.getModulesByIDs(decision.SelectedModules)
	if len(selectedModules) == 0 {
		return nil, fmt.Errorf("no modules available for processing")
	}

	// Process based on routing strategy
	var results []*ProcessingResult
	switch decision.RoutingStrategy {
	case StrategySingleModule:
		results, err = rs.processSingleModule(ctx, request, selectedModules[0])
	case StrategyParallelModules:
		results, err = rs.processParallelModules(ctx, request, selectedModules)
	case StrategyFallback:
		results, err = rs.processWithFallback(ctx, request, selectedModules)
	case StrategyLoadBalanced:
		results, err = rs.processLoadBalanced(ctx, request, selectedModules)
	case StrategyOptimized:
		results, err = rs.processOptimized(ctx, request, selectedModules)
	default:
		results, err = rs.processSingleModule(ctx, request, selectedModules[0])
	}

	if err != nil {
		return nil, fmt.Errorf("processing failed: %w", err)
	}

	// Record processing results
	if rs.metricsCollector != nil {
		for _, result := range results {
			if err := rs.metricsCollector.RecordProcessingResult(ctx, result); err != nil {
				rs.logger.Warn("Failed to record processing metrics", zap.Error(err))
			}
		}
	}

	rs.logger.Info("Request processing completed",
		zap.String("request_id", request.ID),
		zap.Int("results_count", len(results)))

	return results, nil
}

// RegisterModule registers a new module with the routing service
func (rs *routingService) RegisterModule(ctx context.Context, capability *ModuleCapability) error {
	rs.modulesMutex.Lock()
	defer rs.modulesMutex.Unlock()

	if capability.ModuleID == "" {
		return fmt.Errorf("module ID cannot be empty")
	}

	rs.modules[capability.ModuleID] = capability

	// Register with module selector
	if rs.moduleSelector != nil {
		if err := rs.moduleSelector.RegisterModule(capability); err != nil {
			return fmt.Errorf("failed to register with module selector: %w", err)
		}
	}

	rs.logger.Info("Module registered with routing service",
		zap.String("module_id", capability.ModuleID),
		zap.String("module_name", capability.ModuleName))

	return nil
}

// UnregisterModule removes a module from the routing service
func (rs *routingService) UnregisterModule(ctx context.Context, moduleID string) error {
	rs.modulesMutex.Lock()
	defer rs.modulesMutex.Unlock()

	if _, exists := rs.modules[moduleID]; !exists {
		return fmt.Errorf("module %s not found", moduleID)
	}

	delete(rs.modules, moduleID)

	// Unregister from module selector
	if rs.moduleSelector != nil {
		if err := rs.moduleSelector.UnregisterModule(moduleID); err != nil {
			rs.logger.Warn("Failed to unregister from module selector", zap.Error(err))
		}
	}

	rs.logger.Info("Module unregistered from routing service", zap.String("module_id", moduleID))
	return nil
}

// GetModuleCapabilities returns all registered module capabilities
func (rs *routingService) GetModuleCapabilities(ctx context.Context) ([]*ModuleCapability, error) {
	rs.modulesMutex.RLock()
	defer rs.modulesMutex.RUnlock()

	capabilities := make([]*ModuleCapability, 0, len(rs.modules))
	for _, module := range rs.modules {
		capabilities = append(capabilities, module)
	}

	return capabilities, nil
}

// CheckModuleHealth checks the health of a specific module
func (rs *routingService) CheckModuleHealth(ctx context.Context, moduleID string) (*ModuleAvailability, error) {
	if rs.healthChecker == nil {
		return nil, fmt.Errorf("health checker not available")
	}

	return rs.healthChecker.CheckHealth(ctx, moduleID)
}

// GetRoutingMetrics returns routing performance metrics
func (rs *routingService) GetRoutingMetrics(ctx context.Context) (*RoutingMetrics, error) {
	if rs.metricsCollector == nil {
		return nil, fmt.Errorf("metrics collector not available")
	}

	return rs.metricsCollector.GetMetrics(ctx)
}

// UpdateConfig updates the routing service configuration
func (rs *routingService) UpdateConfig(ctx context.Context, config *RoutingConfig) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}

	// Update internal config
	rs.config.DefaultStrategy = config.DefaultStrategy
	rs.config.LoadBalancingEnabled = config.LoadBalancingEnabled
	rs.config.ParallelProcessing = config.ParallelProcessing
	rs.config.FallbackEnabled = config.FallbackEnabled
	rs.config.HealthCheckInterval = config.HealthCheckInterval
	rs.config.DecisionTimeout = config.DecisionTimeout
	rs.config.MaxRetries = config.MaxRetries
	rs.config.CacheEnabled = config.CacheEnabled
	rs.config.CacheTTL = config.CacheTTL

	rs.logger.Info("Routing service configuration updated")
	return nil
}

// GetConfig returns the current routing service configuration
func (rs *routingService) GetConfig(ctx context.Context) (*RoutingConfig, error) {
	config := &RoutingConfig{
		DefaultStrategy:      rs.config.DefaultStrategy,
		LoadBalancingEnabled: rs.config.LoadBalancingEnabled,
		ParallelProcessing:   rs.config.ParallelProcessing,
		FallbackEnabled:      rs.config.FallbackEnabled,
		HealthCheckInterval:  rs.config.HealthCheckInterval,
		DecisionTimeout:      rs.config.DecisionTimeout,
		MaxRetries:           rs.config.MaxRetries,
		CacheEnabled:         rs.config.CacheEnabled,
		CacheTTL:             rs.config.CacheTTL,
	}

	return config, nil
}

// Helper methods for processing strategies

func (rs *routingService) processSingleModule(ctx context.Context, request *VerificationRequest, module *ModuleCapability) ([]*ProcessingResult, error) {
	result, err := rs.executeModule(ctx, request, module)
	if err != nil {
		return nil, err
	}
	return []*ProcessingResult{result}, nil
}

func (rs *routingService) processParallelModules(ctx context.Context, request *VerificationRequest, modules []*ModuleCapability) ([]*ProcessingResult, error) {
	if !rs.config.ParallelProcessing {
		return rs.processSequentialModules(ctx, request, modules)
	}

	// Create channels for results
	resultChan := make(chan *ProcessingResult, len(modules))
	errorChan := make(chan error, len(modules))

	// Start parallel processing
	var wg sync.WaitGroup
	for _, module := range modules {
		wg.Add(1)
		go func(m *ModuleCapability) {
			defer wg.Done()
			result, err := rs.executeModule(ctx, request, m)
			if err != nil {
				errorChan <- err
			} else {
				resultChan <- result
			}
		}(module)
	}

	// Wait for completion
	wg.Wait()
	close(resultChan)
	close(errorChan)

	// Collect results
	var results []*ProcessingResult
	for result := range resultChan {
		results = append(results, result)
	}

	// Check for errors
	select {
	case err := <-errorChan:
		if len(results) == 0 {
			return nil, err
		}
		rs.logger.Warn("Some modules failed during parallel processing", zap.Error(err))
	default:
	}

	return results, nil
}

func (rs *routingService) processSequentialModules(ctx context.Context, request *VerificationRequest, modules []*ModuleCapability) ([]*ProcessingResult, error) {
	var results []*ProcessingResult

	for _, module := range modules {
		result, err := rs.executeModule(ctx, request, module)
		if err != nil {
			rs.logger.Warn("Module processing failed, trying next",
				zap.String("module_id", module.ModuleID),
				zap.Error(err))
			continue
		}
		results = append(results, result)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("all modules failed to process request")
	}

	return results, nil
}

func (rs *routingService) processWithFallback(ctx context.Context, request *VerificationRequest, modules []*ModuleCapability) ([]*ProcessingResult, error) {
	// Try primary modules first
	for _, module := range modules {
		result, err := rs.executeModule(ctx, request, module)
		if err == nil {
			return []*ProcessingResult{result}, nil
		}
		rs.logger.Warn("Primary module failed, trying fallback",
			zap.String("module_id", module.ModuleID),
			zap.Error(err))
	}

	// Try fallback modules
	fallbackModules := rs.getFallbackModules(request)
	for _, module := range fallbackModules {
		result, err := rs.executeModule(ctx, request, module)
		if err == nil {
			return []*ProcessingResult{result}, nil
		}
		rs.logger.Warn("Fallback module failed",
			zap.String("module_id", module.ModuleID),
			zap.Error(err))
	}

	return nil, fmt.Errorf("all modules and fallbacks failed")
}

func (rs *routingService) processLoadBalanced(ctx context.Context, request *VerificationRequest, modules []*ModuleCapability) ([]*ProcessingResult, error) {
	// Get load distribution
	loadDistribution, err := rs.loadBalancer.DistributeLoad(ctx, modules, request)
	if err != nil {
		rs.logger.Warn("Failed to get load distribution, using first module", zap.Error(err))
		return rs.processSingleModule(ctx, request, modules[0])
	}

	// Find module with lowest load
	var bestModule *ModuleCapability
	var lowestLoad float64 = 1.0

	for _, module := range modules {
		if load, exists := loadDistribution[module.ModuleID]; exists && load < lowestLoad {
			lowestLoad = load
			bestModule = module
		}
	}

	if bestModule == nil {
		bestModule = modules[0]
	}

	return rs.processSingleModule(ctx, request, bestModule)
}

func (rs *routingService) processOptimized(ctx context.Context, request *VerificationRequest, modules []*ModuleCapability) ([]*ProcessingResult, error) {
	// Use the first module (already optimized by module selector)
	return rs.processSingleModule(ctx, request, modules[0])
}

// executeModule executes a single module
func (rs *routingService) executeModule(ctx context.Context, request *VerificationRequest, module *ModuleCapability) (*ProcessingResult, error) {
	startTime := time.Now()

	result := &ProcessingResult{
		ID:        generateResultID(),
		RequestID: request.ID,
		ModuleID:  module.ModuleID,
		Status:    StatusProcessing,
	}

	// Simulate module execution (in real implementation, this would call the actual module)
	// For now, we'll simulate a successful result
	time.Sleep(100 * time.Millisecond) // Simulate processing time

	result.Status = StatusCompleted
	result.ProcessingTime = time.Since(startTime)
	result.CompletedAt = time.Now()
	result.Result = map[string]interface{}{
		"verification_status": "verified",
		"confidence_score":    0.95,
		"module_name":         module.ModuleName,
	}

	return result, nil
}

// Helper methods

func (rs *routingService) getCachedDecision(request *VerificationRequest) *RoutingDecision {
	rs.cacheMutex.RLock()
	defer rs.cacheMutex.RUnlock()

	cacheKey := rs.generateCacheKey(request)
	cached, exists := rs.decisionCache[cacheKey]
	if !exists {
		return nil
	}

	// Check if cache entry is expired
	if time.Now().After(cached.ExpiresAt) {
		rs.cacheMutex.RUnlock()
		rs.cacheMutex.Lock()
		delete(rs.decisionCache, cacheKey)
		rs.cacheMutex.Unlock()
		rs.cacheMutex.RLock()
		return nil
	}

	cached.AccessCount++
	return cached.Decision
}

func (rs *routingService) cacheDecision(request *VerificationRequest, decision *RoutingDecision) {
	rs.cacheMutex.Lock()
	defer rs.cacheMutex.Unlock()

	cacheKey := rs.generateCacheKey(request)
	rs.decisionCache[cacheKey] = &CachedDecision{
		Decision:    decision,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(rs.config.CacheTTL),
		AccessCount: 1,
	}
}

func (rs *routingService) generateCacheKey(request *VerificationRequest) string {
	return fmt.Sprintf("%s_%s_%s", request.ID, request.BusinessName, request.BusinessAddress)
}

func (rs *routingService) generateDecisionReason(analysis *RequestAnalysis, modules []*ModuleCapability) string {
	return fmt.Sprintf("Selected %d modules based on %s complexity and %s priority",
		len(modules), analysis.Complexity, analysis.Priority)
}

func (rs *routingService) determineRoutingStrategy(analysis *RequestAnalysis, modules []*ModuleCapability) RoutingStrategy {
	if len(modules) == 1 {
		return StrategySingleModule
	}

	if analysis.Priority == PriorityUrgent {
		return StrategyParallelModules
	}

	if analysis.Complexity == ComplexityAdvanced {
		return StrategyFallback
	}

	return rs.config.DefaultStrategy
}

func (rs *routingService) getModulesByIDs(moduleIDs []string) []*ModuleCapability {
	rs.modulesMutex.RLock()
	defer rs.modulesMutex.RUnlock()

	var modules []*ModuleCapability
	for _, id := range moduleIDs {
		if module, exists := rs.modules[id]; exists {
			modules = append(modules, module)
		}
	}

	return modules
}

func (rs *routingService) getFallbackModules(request *VerificationRequest) []*ModuleCapability {
	// Return basic verification modules as fallbacks
	var fallbacks []*ModuleCapability

	rs.modulesMutex.RLock()
	for _, module := range rs.modules {
		for _, reqType := range module.RequestTypes {
			if reqType == RequestTypeBasic {
				fallbacks = append(fallbacks, module)
				break
			}
		}
	}
	rs.modulesMutex.RUnlock()

	return fallbacks
}

func (rs *routingService) startBackgroundWorkers() {
	// Start health check worker
	go rs.healthCheckWorker()

	// Start cache cleanup worker
	go rs.cacheCleanupWorker()
}

func (rs *routingService) healthCheckWorker() {
	ticker := time.NewTicker(rs.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		rs.performHealthChecks()
	}
}

func (rs *routingService) performHealthChecks() {
	rs.modulesMutex.RLock()
	modules := make([]*ModuleCapability, 0, len(rs.modules))
	for _, module := range rs.modules {
		modules = append(modules, module)
	}
	rs.modulesMutex.RUnlock()

	for _, module := range modules {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		health, err := rs.healthChecker.CheckHealth(ctx, module.ModuleID)
		cancel()

		if err != nil {
			rs.logger.Warn("Health check failed",
				zap.String("module_id", module.ModuleID),
				zap.Error(err))
			continue
		}

		// Update module availability
		rs.modulesMutex.Lock()
		if existingModule, exists := rs.modules[module.ModuleID]; exists {
			existingModule.Availability = *health
		}
		rs.modulesMutex.Unlock()
	}
}

func (rs *routingService) cacheCleanupWorker() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rs.cleanupExpiredCache()
	}
}

func (rs *routingService) cleanupExpiredCache() {
	rs.cacheMutex.Lock()
	defer rs.cacheMutex.Unlock()

	now := time.Now()
	for key, cached := range rs.decisionCache {
		if now.After(cached.ExpiresAt) {
			delete(rs.decisionCache, key)
		}
	}
}

// ID generation helpers
func generateDecisionID() string {
	return fmt.Sprintf("decision_%d", time.Now().UnixNano())
}

func generateResultID() string {
	return fmt.Sprintf("result_%d", time.Now().UnixNano())
}
