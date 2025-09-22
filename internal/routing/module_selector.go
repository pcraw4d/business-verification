package routing

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/architecture"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/shared"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// ModuleSelector selects the most appropriate module for processing requests
type ModuleSelector struct {
	logger        *observability.Logger
	tracer        trace.Tracer
	config        ModuleSelectorConfig
	moduleManager ModuleManager
	metrics       *observability.Metrics

	// Module registry and state
	availableModules map[string]ModuleInfo
	moduleMutex      sync.RWMutex

	// Performance tracking
	modulePerformance map[string]*ModulePerformance
	performanceMutex  sync.RWMutex
}

// ModuleSelectorConfig holds configuration for module selection
type ModuleSelectorConfig struct {
	EnablePerformanceTracking bool                  `json:"enable_performance_tracking"`
	EnableLoadBalancing       bool                  `json:"enable_load_balancing"`
	EnableFallbackRouting     bool                  `json:"enable_fallback_routing"`
	MaxRetries                int                   `json:"max_retries"`
	RetryDelay                time.Duration         `json:"retry_delay"`
	PerformanceWindow         time.Duration         `json:"performance_window"`
	LoadBalancingStrategy     LoadBalancingStrategy `json:"load_balancing_strategy"`
	ConfidenceThreshold       float64               `json:"confidence_threshold"`
}

// LoadBalancingStrategy defines the load balancing strategy
type LoadBalancingStrategy string

const (
	LoadBalancingStrategyRoundRobin      LoadBalancingStrategy = "round_robin"
	LoadBalancingStrategyLeastLoaded     LoadBalancingStrategy = "least_loaded"
	LoadBalancingStrategyBestPerformance LoadBalancingStrategy = "best_performance"
	LoadBalancingStrategyAdaptive        LoadBalancingStrategy = "adaptive"
)

// ModuleInfo represents information about an available module
type ModuleInfo struct {
	ModuleID       string                          `json:"module_id"`
	ModuleType     string                          `json:"module_type"`
	Capabilities   []architecture.ModuleCapability `json:"capabilities"`
	Priority       architecture.ModulePriority     `json:"priority"`
	HealthStatus   architecture.ModuleStatus       `json:"health_status"`
	IsRunning      bool                            `json:"is_running"`
	CurrentLoad    int                             `json:"current_load"`
	MaxConcurrency int                             `json:"max_concurrency"`
	LastUpdated    time.Time                       `json:"last_updated"`
	Metadata       map[string]interface{}          `json:"metadata"`
}

// ModulePerformance tracks performance metrics for a module
type ModulePerformance struct {
	ModuleID           string        `json:"module_id"`
	TotalRequests      int64         `json:"total_requests"`
	SuccessfulRequests int64         `json:"successful_requests"`
	FailedRequests     int64         `json:"failed_requests"`
	AverageLatency     time.Duration `json:"average_latency"`
	LastLatency        time.Duration `json:"last_latency"`
	SuccessRate        float64       `json:"success_rate"`
	LastUpdated        time.Time     `json:"last_updated"`
	PerformanceScore   float64       `json:"performance_score"`
}

// ModuleManager interface for managing modules
type ModuleManager interface {
	GetAvailableModules() map[string]architecture.Module
	GetModuleByID(moduleID string) (architecture.Module, bool)
	GetModulesByType(moduleType string) []architecture.Module
	GetModuleHealth(moduleID string) (architecture.ModuleStatus, error)
}

// SelectionResult represents the result of module selection
type SelectionResult struct {
	SelectedModule    *ModuleInfo             `json:"selected_module"`
	FallbackModules   []*ModuleInfo           `json:"fallback_modules"`
	SelectionReason   string                  `json:"selection_reason"`
	Confidence        float64                 `json:"confidence"`
	ExpectedLatency   time.Duration           `json:"expected_latency"`
	SelectionMetadata map[string]interface{}  `json:"selection_metadata"`
	Recommendations   []RoutingRecommendation `json:"recommendations"`
}

// NewModuleSelector creates a new module selector
func NewModuleSelector(
	logger *observability.Logger,
	tracer trace.Tracer,
	config ModuleSelectorConfig,
	moduleManager ModuleManager,
	metrics *observability.Metrics,
) *ModuleSelector {
	// Set default configuration if not provided
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	if config.PerformanceWindow == 0 {
		config.PerformanceWindow = 5 * time.Minute
	}
	if config.LoadBalancingStrategy == "" {
		config.LoadBalancingStrategy = LoadBalancingStrategyAdaptive
	}
	if config.ConfidenceThreshold == 0 {
		config.ConfidenceThreshold = 0.7
	}

	return &ModuleSelector{
		logger:        logger,
		tracer:        tracer,
		config:        config,
		moduleManager: moduleManager,
		metrics:       metrics,

		availableModules:  make(map[string]ModuleInfo),
		modulePerformance: make(map[string]*ModulePerformance),
	}
}

// SelectModule selects the most appropriate module for processing a request
func (ms *ModuleSelector) SelectModule(
	ctx context.Context,
	req *shared.BusinessClassificationRequest,
	analysis *RequestAnalysisResult,
) (*SelectionResult, error) {
	ctx, span := ms.tracer.Start(ctx, "ModuleSelector.SelectModule")
	defer span.End()

	span.SetAttributes(
		attribute.String("request.id", req.ID),
		attribute.String("request.type", string(analysis.RequestType)),
		attribute.String("complexity", string(analysis.Complexity)),
		attribute.String("priority", string(analysis.Priority)),
	)

	// Step 1: Update available modules
	ms.updateAvailableModules()

	// Step 2: Filter modules based on capabilities and requirements
	candidateModules := ms.filterCandidateModules(analysis)

	// Step 3: Rank modules based on selection criteria
	rankedModules := ms.rankModules(candidateModules, analysis)

	// Step 4: Select primary module using input type-based selection
	selectedModule := ms.selectModuleByInputType(rankedModules, analysis)
	if selectedModule == nil {
		// Fallback to traditional selection if input type-based selection fails
		selectedModule = ms.selectPrimaryModule(rankedModules, analysis)
	}

	// Step 5: Select fallback modules
	fallbackModules := ms.selectFallbackModules(rankedModules, selectedModule, analysis)

	// Step 6: Create selection result
	result := &SelectionResult{
		SelectedModule:    selectedModule,
		FallbackModules:   fallbackModules,
		SelectionReason:   ms.generateSelectionReason(selectedModule, analysis),
		Confidence:        ms.calculateSelectionConfidence(selectedModule, analysis),
		ExpectedLatency:   ms.calculateExpectedLatency(selectedModule, analysis),
		SelectionMetadata: ms.generateSelectionMetadata(selectedModule, analysis),
		Recommendations:   analysis.Recommendations,
	}

	// Log selection results
	ms.logger.WithComponent("module_selector").Info("module_selection_completed", map[string]interface{}{
		"request_id":          req.ID,
		"selected_module":     selectedModule.ModuleID,
		"module_type":         selectedModule.ModuleType,
		"confidence":          result.Confidence,
		"expected_latency_ms": result.ExpectedLatency.Milliseconds(),
		"fallback_count":      len(fallbackModules),
	})

	return result, nil
}

// updateAvailableModules updates the registry of available modules
func (ms *ModuleSelector) updateAvailableModules() {
	ms.moduleMutex.Lock()
	defer ms.moduleMutex.Unlock()

	availableModules := ms.moduleManager.GetAvailableModules()

	for moduleID, module := range availableModules {
		healthStatus, err := ms.moduleManager.GetModuleHealth(moduleID)
		if err != nil {
			ms.logger.WithComponent("module_selector").Warn("failed_to_get_module_health", map[string]interface{}{
				"module_id": moduleID,
				"error":     err.Error(),
			})
			healthStatus = architecture.ModuleStatusUnhealthy
		}

		metadata := module.Metadata()

		moduleInfo := ModuleInfo{
			ModuleID:       moduleID,
			ModuleType:     ms.getModuleType(module),
			Capabilities:   metadata.Capabilities,
			Priority:       metadata.Priority,
			HealthStatus:   healthStatus,
			IsRunning:      module.IsRunning(),
			CurrentLoad:    ms.getCurrentLoad(moduleID),
			MaxConcurrency: ms.getMaxConcurrency(module),
			LastUpdated:    time.Now(),
			Metadata:       ms.getModuleMetadata(module),
		}

		ms.availableModules[moduleID] = moduleInfo
	}
}

// filterCandidateModules filters modules based on capabilities and requirements
func (ms *ModuleSelector) filterCandidateModules(analysis *RequestAnalysisResult) []ModuleInfo {
	ms.moduleMutex.RLock()
	defer ms.moduleMutex.RUnlock()

	var candidates []ModuleInfo

	for _, moduleInfo := range ms.availableModules {
		// Check if module is healthy and running
		if moduleInfo.HealthStatus != architecture.ModuleStatusHealthy || !moduleInfo.IsRunning {
			continue
		}

		// Check if module has required capabilities
		if !ms.hasRequiredCapabilities(moduleInfo, analysis) {
			continue
		}

		// Check if module can handle the request type
		if !ms.canHandleRequestType(moduleInfo, analysis.RequestType) {
			continue
		}

		// Check if module has capacity
		if moduleInfo.CurrentLoad >= moduleInfo.MaxConcurrency {
			continue
		}

		candidates = append(candidates, moduleInfo)
	}

	return candidates
}

// hasRequiredCapabilities checks if a module has the required capabilities
func (ms *ModuleSelector) hasRequiredCapabilities(moduleInfo ModuleInfo, analysis *RequestAnalysisResult) bool {
	requiredCapabilities := ms.getRequiredCapabilities(analysis)

	for _, required := range requiredCapabilities {
		found := false
		for _, capability := range moduleInfo.Capabilities {
			if capability == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// getRequiredCapabilities determines required capabilities based on analysis
func (ms *ModuleSelector) getRequiredCapabilities(analysis *RequestAnalysisResult) []architecture.ModuleCapability {
	var capabilities []architecture.ModuleCapability

	// Always require classification capability
	capabilities = append(capabilities, architecture.CapabilityClassification)

	// Add specific capabilities based on request characteristics
	switch analysis.RequestType {
	case RequestTypeComplex:
		capabilities = append(capabilities, architecture.CapabilityDataExtraction)
	case RequestTypeBatch:
		capabilities = append(capabilities, architecture.CapabilityDataExtraction) // Use data extraction for batch
	case RequestTypeUrgent:
		capabilities = append(capabilities, architecture.CapabilityVerification) // Use verification for urgent
	}

	// Add capabilities based on recommendations
	for _, recommendation := range analysis.Recommendations {
		switch recommendation.ModuleType {
		case "website_analysis":
			capabilities = append(capabilities, architecture.CapabilityWebAnalysis)
		case "web_search_analysis":
			capabilities = append(capabilities, architecture.CapabilityWebAnalysis)
		case "ml_classification":
			capabilities = append(capabilities, architecture.CapabilityMLPrediction)
		}
	}

	return capabilities
}

// canHandleRequestType checks if a module can handle the request type
func (ms *ModuleSelector) canHandleRequestType(moduleInfo ModuleInfo, requestType RequestType) bool {
	// Check module metadata for supported request types
	if metadata, ok := moduleInfo.Metadata["supported_request_types"]; ok {
		if supportedTypes, ok := metadata.([]string); ok {
			for _, supportedType := range supportedTypes {
				if supportedType == string(requestType) {
					return true
				}
			}
		}
	}

	// Enhanced input type-based selection
	return ms.canHandleRequestTypeByInput(moduleInfo, requestType)
}

// canHandleRequestTypeByInput determines if a module can handle a request type based on input characteristics
func (ms *ModuleSelector) canHandleRequestTypeByInput(moduleInfo ModuleInfo, requestType RequestType) bool {
	switch moduleInfo.ModuleType {
	case "website_analysis":
		// Website analysis is best for requests with website URLs
		return requestType == RequestTypeSimple || requestType == RequestTypeStandard || requestType == RequestTypeComplex
	case "web_search_analysis":
		// Web search is best for research and complex requests without websites
		return requestType == RequestTypeStandard || requestType == RequestTypeComplex || requestType == RequestTypeResearch
	case "ml_classification":
		// ML classification is best for standard and complex requests with good data
		return requestType == RequestTypeStandard || requestType == RequestTypeComplex
	case "keyword_classification":
		// Keyword classification is best for simple requests
		return requestType == RequestTypeSimple || requestType == RequestTypeStandard
	default:
		return true // Generic modules can handle most request types
	}
}

// selectModuleByInputType implements input type-based module selection
func (ms *ModuleSelector) selectModuleByInputType(candidates []ModuleInfo, analysis *RequestAnalysisResult) *ModuleInfo {
	// Get input characteristics from the request
	inputCharacteristics := ms.analyzeInputCharacteristics(analysis)

	// Score modules based on input type compatibility
	var bestModule *ModuleInfo
	bestScore := 0.0

	for i := range candidates {
		score := ms.calculateInputTypeScore(candidates[i], inputCharacteristics, analysis.RequestType)
		if score > bestScore {
			bestScore = score
			bestModule = &candidates[i]
		}
	}

	return bestModule
}

// rankModules ranks candidate modules based on selection criteria
func (ms *ModuleSelector) rankModules(candidates []ModuleInfo, analysis *RequestAnalysisResult) []ModuleInfo {
	// Create a copy of candidates for ranking
	ranked := make([]ModuleInfo, len(candidates))
	copy(ranked, candidates)

	// Sort modules based on multiple criteria
	sort.Slice(ranked, func(i, j int) bool {
		scoreI := ms.calculateModuleScore(ranked[i], analysis)
		scoreJ := ms.calculateModuleScore(ranked[j], analysis)
		return scoreI > scoreJ
	})

	return ranked
}

// calculateModuleScore calculates a score for module selection
func (ms *ModuleSelector) calculateModuleScore(moduleInfo ModuleInfo, analysis *RequestAnalysisResult) float64 {
	score := 0.0

	// Priority score (higher priority modules get higher scores)
	priorityScore := float64(moduleInfo.Priority) / float64(architecture.PriorityCritical)
	score += priorityScore * 0.3

	// Performance score
	performanceScore := ms.getPerformanceScore(moduleInfo.ModuleID)
	score += performanceScore * 0.3

	// Load score (less loaded modules get higher scores)
	loadScore := 1.0 - (float64(moduleInfo.CurrentLoad) / float64(moduleInfo.MaxConcurrency))
	score += loadScore * 0.2

	// Capability match score
	capabilityScore := ms.calculateCapabilityMatchScore(moduleInfo, analysis)
	score += capabilityScore * 0.2

	return score
}

// getPerformanceScore gets the performance score for a module
func (ms *ModuleSelector) getPerformanceScore(moduleID string) float64 {
	ms.performanceMutex.RLock()
	defer ms.performanceMutex.RUnlock()

	if performance, exists := ms.modulePerformance[moduleID]; exists {
		return performance.PerformanceScore
	}

	return 0.5 // Default score for modules without performance data
}

// calculateCapabilityMatchScore calculates how well a module matches the requirements
func (ms *ModuleSelector) calculateCapabilityMatchScore(moduleInfo ModuleInfo, analysis *RequestAnalysisResult) float64 {
	requiredCapabilities := ms.getRequiredCapabilities(analysis)
	matchedCapabilities := 0

	for _, required := range requiredCapabilities {
		for _, capability := range moduleInfo.Capabilities {
			if capability == required {
				matchedCapabilities++
				break
			}
		}
	}

	if len(requiredCapabilities) == 0 {
		return 1.0
	}

	return float64(matchedCapabilities) / float64(len(requiredCapabilities))
}

// selectPrimaryModule selects the primary module for processing
func (ms *ModuleSelector) selectPrimaryModule(rankedModules []ModuleInfo, analysis *RequestAnalysisResult) *ModuleInfo {
	if len(rankedModules) == 0 {
		return nil
	}

	// Apply load balancing strategy
	switch ms.config.LoadBalancingStrategy {
	case LoadBalancingStrategyRoundRobin:
		return ms.selectRoundRobin(rankedModules)
	case LoadBalancingStrategyLeastLoaded:
		return ms.selectLeastLoaded(rankedModules)
	case LoadBalancingStrategyBestPerformance:
		return ms.selectBestPerformance(rankedModules)
	case LoadBalancingStrategyAdaptive:
		return ms.selectAdaptive(rankedModules, analysis)
	default:
		return &rankedModules[0] // Default to first ranked module
	}
}

// selectRoundRobin selects module using round-robin strategy
func (ms *ModuleSelector) selectRoundRobin(rankedModules []ModuleInfo) *ModuleInfo {
	// Simple round-robin implementation
	// In a real implementation, this would maintain state across requests
	return &rankedModules[0]
}

// selectLeastLoaded selects the least loaded module
func (ms *ModuleSelector) selectLeastLoaded(rankedModules []ModuleInfo) *ModuleInfo {
	if len(rankedModules) == 0 {
		return nil
	}

	leastLoaded := rankedModules[0]
	lowestLoad := float64(leastLoaded.CurrentLoad) / float64(leastLoaded.MaxConcurrency)

	for _, module := range rankedModules[1:] {
		load := float64(module.CurrentLoad) / float64(module.MaxConcurrency)
		if load < lowestLoad {
			leastLoaded = module
			lowestLoad = load
		}
	}

	return &leastLoaded
}

// selectBestPerformance selects the module with best performance
func (ms *ModuleSelector) selectBestPerformance(rankedModules []ModuleInfo) *ModuleInfo {
	if len(rankedModules) == 0 {
		return nil
	}

	bestModule := rankedModules[0]
	bestScore := ms.getPerformanceScore(bestModule.ModuleID)

	for _, module := range rankedModules[1:] {
		score := ms.getPerformanceScore(module.ModuleID)
		if score > bestScore {
			bestModule = module
			bestScore = score
		}
	}

	return &bestModule
}

// selectAdaptive selects module using adaptive strategy
func (ms *ModuleSelector) selectAdaptive(rankedModules []ModuleInfo, analysis *RequestAnalysisResult) *ModuleInfo {
	// Adaptive strategy considers multiple factors
	// For now, use a weighted combination of performance and load
	if len(rankedModules) == 0 {
		return nil
	}

	bestModule := rankedModules[0]
	bestScore := ms.calculateAdaptiveScore(bestModule, analysis)

	for _, module := range rankedModules[1:] {
		score := ms.calculateAdaptiveScore(module, analysis)
		if score > bestScore {
			bestModule = module
			bestScore = score
		}
	}

	return &bestModule
}

// calculateAdaptiveScore calculates adaptive selection score
func (ms *ModuleSelector) calculateAdaptiveScore(moduleInfo ModuleInfo, analysis *RequestAnalysisResult) float64 {
	score := 0.0

	// Performance weight (40%)
	performanceScore := ms.getPerformanceScore(moduleInfo.ModuleID)
	score += performanceScore * 0.4

	// Load weight (30%)
	loadScore := 1.0 - (float64(moduleInfo.CurrentLoad) / float64(moduleInfo.MaxConcurrency))
	score += loadScore * 0.3

	// Priority weight (20%)
	priorityScore := float64(moduleInfo.Priority) / float64(architecture.PriorityCritical)
	score += priorityScore * 0.2

	// Request type match weight (10%)
	typeMatchScore := ms.calculateRequestTypeMatchScore(moduleInfo, analysis.RequestType)
	score += typeMatchScore * 0.1

	return score
}

// calculateRequestTypeMatchScore calculates how well a module matches the request type
func (ms *ModuleSelector) calculateRequestTypeMatchScore(moduleInfo ModuleInfo, requestType RequestType) float64 {
	// Simple scoring based on module type and request type compatibility
	switch moduleInfo.ModuleType {
	case "website_analysis":
		switch requestType {
		case RequestTypeSimple, RequestTypeStandard, RequestTypeComplex:
			return 1.0
		default:
			return 0.5
		}
	case "web_search_analysis":
		switch requestType {
		case RequestTypeStandard, RequestTypeComplex, RequestTypeResearch:
			return 1.0
		default:
			return 0.7
		}
	case "ml_classification":
		switch requestType {
		case RequestTypeStandard, RequestTypeComplex:
			return 1.0
		default:
			return 0.6
		}
	case "keyword_classification":
		switch requestType {
		case RequestTypeSimple, RequestTypeStandard:
			return 1.0
		default:
			return 0.8
		}
	default:
		return 0.5
	}
}

// selectFallbackModules selects fallback modules
func (ms *ModuleSelector) selectFallbackModules(
	rankedModules []ModuleInfo,
	primaryModule *ModuleInfo,
	analysis *RequestAnalysisResult,
) []*ModuleInfo {
	if !ms.config.EnableFallbackRouting {
		return nil
	}

	var fallbacks []*ModuleInfo
	maxFallbacks := 2

	for _, module := range rankedModules {
		if len(fallbacks) >= maxFallbacks {
			break
		}

		// Skip the primary module
		if primaryModule != nil && module.ModuleID == primaryModule.ModuleID {
			continue
		}

		// Check if module is suitable as fallback
		if ms.isSuitableFallback(module, primaryModule, analysis) {
			fallbacks = append(fallbacks, &module)
		}
	}

	return fallbacks
}

// isSuitableFallback checks if a module is suitable as a fallback
func (ms *ModuleSelector) isSuitableFallback(
	moduleInfo ModuleInfo,
	primaryModule *ModuleInfo,
	analysis *RequestAnalysisResult,
) bool {
	// Must be healthy and running
	if moduleInfo.HealthStatus != architecture.ModuleStatusHealthy || !moduleInfo.IsRunning {
		return false
	}

	// Must have capacity
	if moduleInfo.CurrentLoad >= moduleInfo.MaxConcurrency {
		return false
	}

	// Must have different characteristics than primary module
	if primaryModule != nil {
		if moduleInfo.ModuleType == primaryModule.ModuleType {
			return false
		}
	}

	// Must have minimum performance score
	performanceScore := ms.getPerformanceScore(moduleInfo.ModuleID)
	if performanceScore < 0.3 {
		return false
	}

	return true
}

// generateSelectionReason generates a human-readable reason for the selection
func (ms *ModuleSelector) generateSelectionReason(moduleInfo *ModuleInfo, analysis *RequestAnalysisResult) string {
	if moduleInfo == nil {
		return "No suitable module available"
	}

	reasons := []string{}

	// Add reason based on module type
	switch moduleInfo.ModuleType {
	case "website_analysis":
		reasons = append(reasons, "Website analysis module selected for URL-based classification")
	case "web_search_analysis":
		reasons = append(reasons, "Web search analysis module selected for comprehensive research")
	case "ml_classification":
		reasons = append(reasons, "ML classification module selected for rich data processing")
	case "keyword_classification":
		reasons = append(reasons, "Keyword classification module selected as efficient fallback")
	default:
		reasons = append(reasons, "Generic module selected for classification")
	}

	// Add reason based on load balancing strategy
	switch ms.config.LoadBalancingStrategy {
	case LoadBalancingStrategyRoundRobin:
		reasons = append(reasons, "using round-robin load balancing")
	case LoadBalancingStrategyLeastLoaded:
		reasons = append(reasons, "using least-loaded selection")
	case LoadBalancingStrategyBestPerformance:
		reasons = append(reasons, "using best performance selection")
	case LoadBalancingStrategyAdaptive:
		reasons = append(reasons, "using adaptive load balancing")
	}

	// Add reason based on performance
	performanceScore := ms.getPerformanceScore(moduleInfo.ModuleID)
	if performanceScore > 0.8 {
		reasons = append(reasons, "high performance module")
	} else if performanceScore > 0.6 {
		reasons = append(reasons, "good performance module")
	}

	return strings.Join(reasons, " - ")
}

// calculateSelectionConfidence calculates confidence in the selection
func (ms *ModuleSelector) calculateSelectionConfidence(moduleInfo *ModuleInfo, analysis *RequestAnalysisResult) float64 {
	if moduleInfo == nil {
		return 0.0
	}

	confidence := 0.5 // Base confidence

	// Add confidence based on performance
	performanceScore := ms.getPerformanceScore(moduleInfo.ModuleID)
	confidence += performanceScore * 0.2

	// Add confidence based on health status
	if moduleInfo.HealthStatus == architecture.ModuleStatusHealthy {
		confidence += 0.2
	}

	// Add confidence based on load
	loadScore := 1.0 - (float64(moduleInfo.CurrentLoad) / float64(moduleInfo.MaxConcurrency))
	confidence += loadScore * 0.1

	// Add confidence based on capability match
	capabilityScore := ms.calculateCapabilityMatchScore(*moduleInfo, analysis)
	confidence += capabilityScore * 0.2

	// Normalize to 0-1 range
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// calculateExpectedLatency calculates expected latency for the selected module
func (ms *ModuleSelector) calculateExpectedLatency(moduleInfo *ModuleInfo, analysis *RequestAnalysisResult) time.Duration {
	if moduleInfo == nil {
		return 30 * time.Second // Default timeout
	}

	baseLatency := analysis.ResourceRequirements.EstimatedTime

	// Adjust based on module performance
	performanceScore := ms.getPerformanceScore(moduleInfo.ModuleID)
	if performanceScore > 0.8 {
		baseLatency = time.Duration(float64(baseLatency) * 0.8)
	} else if performanceScore < 0.4 {
		baseLatency = time.Duration(float64(baseLatency) * 1.5)
	}

	// Adjust based on current load
	loadFactor := float64(moduleInfo.CurrentLoad) / float64(moduleInfo.MaxConcurrency)
	if loadFactor > 0.8 {
		baseLatency = time.Duration(float64(baseLatency) * 1.3)
	}

	return baseLatency
}

// generateSelectionMetadata generates metadata for the selection
func (ms *ModuleSelector) generateSelectionMetadata(moduleInfo *ModuleInfo, analysis *RequestAnalysisResult) map[string]interface{} {
	metadata := map[string]interface{}{
		"selector_version":        "1.0.0",
		"selection_timestamp":     time.Now().Unix(),
		"load_balancing_strategy": string(ms.config.LoadBalancingStrategy),
		"enable_fallback_routing": ms.config.EnableFallbackRouting,
		"confidence_threshold":    ms.config.ConfidenceThreshold,
	}

	if moduleInfo != nil {
		metadata["selected_module_id"] = moduleInfo.ModuleID
		metadata["selected_module_type"] = moduleInfo.ModuleType
		metadata["module_health_status"] = string(moduleInfo.HealthStatus)
		metadata["module_current_load"] = moduleInfo.CurrentLoad
		metadata["module_max_concurrency"] = moduleInfo.MaxConcurrency
		metadata["module_performance_score"] = ms.getPerformanceScore(moduleInfo.ModuleID)
	}

	return metadata
}

// UpdateModulePerformance updates performance metrics for a module
func (ms *ModuleSelector) UpdateModulePerformance(
	moduleID string,
	success bool,
	latency time.Duration,
) {
	if !ms.config.EnablePerformanceTracking {
		return
	}

	ms.performanceMutex.Lock()
	defer ms.performanceMutex.Unlock()

	performance, exists := ms.modulePerformance[moduleID]
	if !exists {
		performance = &ModulePerformance{
			ModuleID: moduleID,
		}
		ms.modulePerformance[moduleID] = performance
	}

	// Update metrics
	performance.TotalRequests++
	if success {
		performance.SuccessfulRequests++
	} else {
		performance.FailedRequests++
	}

	// Update latency (simple moving average)
	if performance.TotalRequests == 1 {
		performance.AverageLatency = latency
	} else {
		// Exponential moving average
		alpha := 0.1
		performance.AverageLatency = time.Duration(
			float64(performance.AverageLatency)*(1-alpha) + float64(latency)*alpha,
		)
	}

	performance.LastLatency = latency
	performance.LastUpdated = time.Now()

	// Calculate success rate
	if performance.TotalRequests > 0 {
		performance.SuccessRate = float64(performance.SuccessfulRequests) / float64(performance.TotalRequests)
	}

	// Calculate performance score
	performance.PerformanceScore = ms.calculatePerformanceScore(performance)
}

// calculatePerformanceScore calculates overall performance score
func (ms *ModuleSelector) calculatePerformanceScore(performance *ModulePerformance) float64 {
	score := 0.0

	// Success rate weight (60%)
	score += performance.SuccessRate * 0.6

	// Latency weight (40%) - lower latency = higher score
	// Normalize latency to 0-1 range (assuming max acceptable latency is 30 seconds)
	maxLatency := 30 * time.Second
	latencyScore := 1.0 - (float64(performance.AverageLatency) / float64(maxLatency))
	if latencyScore < 0 {
		latencyScore = 0
	}
	score += latencyScore * 0.4

	return score
}

// InputCharacteristics represents the characteristics of the input data
type InputCharacteristics struct {
	HasWebsiteURL       bool    `json:"has_website_url"`
	HasBusinessName     bool    `json:"has_business_name"`
	HasDescription      bool    `json:"has_description"`
	HasKeywords         bool    `json:"has_keywords"`
	HasIndustry         bool    `json:"has_industry"`
	HasGeographicRegion bool    `json:"has_geographic_region"`
	DataQuality         float64 `json:"data_quality"`
	DataCompleteness    float64 `json:"data_completeness"`
	InputComplexity     float64 `json:"input_complexity"`
}

// analyzeInputCharacteristics analyzes the characteristics of the input data
func (ms *ModuleSelector) analyzeInputCharacteristics(analysis *RequestAnalysisResult) *InputCharacteristics {
	// This would be populated from the request analysis
	// For now, we'll create a basic implementation
	return &InputCharacteristics{
		HasWebsiteURL:       analysis.AnalysisMetadata["has_website_url"] == true,
		HasBusinessName:     analysis.AnalysisMetadata["has_business_name"] == true,
		HasDescription:      analysis.AnalysisMetadata["has_description"] == true,
		HasKeywords:         analysis.AnalysisMetadata["has_keywords"] == true,
		HasIndustry:         analysis.AnalysisMetadata["has_industry"] == true,
		HasGeographicRegion: analysis.AnalysisMetadata["has_geographic_region"] == true,
		DataQuality:         ms.extractFloat64(analysis.AnalysisMetadata, "data_quality", 0.5),
		DataCompleteness:    ms.extractFloat64(analysis.AnalysisMetadata, "data_completeness", 0.5),
		InputComplexity:     ms.extractFloat64(analysis.AnalysisMetadata, "input_complexity", 0.5),
	}
}

// calculateInputTypeScore calculates how well a module matches the input type
func (ms *ModuleSelector) calculateInputTypeScore(moduleInfo ModuleInfo, characteristics *InputCharacteristics, requestType RequestType) float64 {
	score := 0.0

	// Base score for module type compatibility
	score += ms.getModuleTypeCompatibilityScore(moduleInfo.ModuleType, characteristics, requestType)

	// Input-specific scoring
	switch moduleInfo.ModuleType {
	case "website_analysis":
		score += ms.calculateWebsiteAnalysisScore(characteristics, requestType)
	case "web_search_analysis":
		score += ms.calculateWebSearchAnalysisScore(characteristics, requestType)
	case "ml_classification":
		score += ms.calculateMLClassificationScore(characteristics, requestType)
	case "keyword_classification":
		score += ms.calculateKeywordClassificationScore(characteristics, requestType)
	}

	// Performance and load considerations
	performanceScore := ms.getPerformanceScore(moduleInfo.ModuleID)
	loadScore := 1.0 - (float64(moduleInfo.CurrentLoad) / float64(moduleInfo.MaxConcurrency))

	score += performanceScore * 0.2
	score += loadScore * 0.1

	return score
}

// getModuleTypeCompatibilityScore returns base compatibility score for module type
func (ms *ModuleSelector) getModuleTypeCompatibilityScore(moduleType string, characteristics *InputCharacteristics, requestType RequestType) float64 {
	switch moduleType {
	case "website_analysis":
		if characteristics.HasWebsiteURL {
			return 0.9 // Excellent match
		}
		return 0.3 // Poor match without website
	case "web_search_analysis":
		if requestType == RequestTypeResearch || requestType == RequestTypeComplex {
			return 0.8 // Good match for research/complex requests
		}
		return 0.6 // Moderate match for other types
	case "ml_classification":
		if characteristics.DataQuality > 0.7 && characteristics.DataCompleteness > 0.6 {
			return 0.8 // Good match for high-quality data
		}
		return 0.5 // Moderate match for lower quality data
	case "keyword_classification":
		if requestType == RequestTypeSimple || characteristics.DataCompleteness < 0.4 {
			return 0.7 // Good match for simple requests or incomplete data
		}
		return 0.4 // Lower match for complex requests
	default:
		return 0.5 // Default score for unknown module types
	}
}

// calculateWebsiteAnalysisScore calculates score for website analysis module
func (ms *ModuleSelector) calculateWebsiteAnalysisScore(characteristics *InputCharacteristics, requestType RequestType) float64 {
	score := 0.0

	// Website URL is essential for website analysis
	if characteristics.HasWebsiteURL {
		score += 0.4

		// Additional data improves analysis quality
		if characteristics.HasBusinessName {
			score += 0.2
		}
		if characteristics.HasDescription {
			score += 0.15
		}
		if characteristics.HasKeywords {
			score += 0.1
		}
		if characteristics.HasIndustry {
			score += 0.1
		}
		if characteristics.HasGeographicRegion {
			score += 0.05
		}
	}

	// Request type considerations
	switch requestType {
	case RequestTypeComplex:
		score += 0.1 // Website analysis excels with complex requests
	case RequestTypeStandard:
		score += 0.05 // Good for standard requests
	case RequestTypeSimple:
		score += 0.0 // May be overkill for simple requests
	}

	return score
}

// calculateWebSearchAnalysisScore calculates score for web search analysis module
func (ms *ModuleSelector) calculateWebSearchAnalysisScore(characteristics *InputCharacteristics, requestType RequestType) float64 {
	score := 0.0

	// Web search works well with business names and descriptions
	if characteristics.HasBusinessName {
		score += 0.3
	}
	if characteristics.HasDescription {
		score += 0.25
	}
	if characteristics.HasKeywords {
		score += 0.2
	}
	if characteristics.HasIndustry {
		score += 0.15
	}
	if characteristics.HasGeographicRegion {
		score += 0.1
	}

	// Request type considerations
	switch requestType {
	case RequestTypeResearch:
		score += 0.2 // Excellent for research requests
	case RequestTypeComplex:
		score += 0.15 // Good for complex requests
	case RequestTypeStandard:
		score += 0.1 // Moderate for standard requests
	case RequestTypeSimple:
		score += 0.05 // Lower for simple requests
	}

	// Web search is particularly useful when no website URL is available
	if !characteristics.HasWebsiteURL {
		score += 0.1
	}

	return score
}

// calculateMLClassificationScore calculates score for ML classification module
func (ms *ModuleSelector) calculateMLClassificationScore(characteristics *InputCharacteristics, requestType RequestType) float64 {
	score := 0.0

	// ML classification benefits from high-quality, complete data
	if characteristics.DataQuality > 0.7 {
		score += 0.3
	} else if characteristics.DataQuality > 0.5 {
		score += 0.2
	} else {
		score += 0.1
	}

	if characteristics.DataCompleteness > 0.6 {
		score += 0.25
	} else if characteristics.DataCompleteness > 0.4 {
		score += 0.15
	} else {
		score += 0.05
	}

	// Individual field contributions
	if characteristics.HasBusinessName {
		score += 0.2
	}
	if characteristics.HasDescription {
		score += 0.15
	}
	if characteristics.HasKeywords {
		score += 0.1
	}
	if characteristics.HasIndustry {
		score += 0.1
	}

	// Request type considerations
	switch requestType {
	case RequestTypeComplex:
		score += 0.1 // ML excels with complex patterns
	case RequestTypeStandard:
		score += 0.05 // Good for standard requests
	case RequestTypeSimple:
		score += 0.0 // May be overkill for simple requests
	}

	return score
}

// calculateKeywordClassificationScore calculates score for keyword classification module
func (ms *ModuleSelector) calculateKeywordClassificationScore(characteristics *InputCharacteristics, requestType RequestType) float64 {
	score := 0.0

	// Keyword classification works well with keywords and business names
	if characteristics.HasKeywords {
		score += 0.4
	}
	if characteristics.HasBusinessName {
		score += 0.3
	}
	if characteristics.HasDescription {
		score += 0.2
	}
	if characteristics.HasIndustry {
		score += 0.1
	}

	// Request type considerations
	switch requestType {
	case RequestTypeSimple:
		score += 0.2 // Excellent for simple requests
	case RequestTypeStandard:
		score += 0.1 // Good for standard requests
	case RequestTypeComplex:
		score += 0.0 // May be insufficient for complex requests
	}

	// Keyword classification is good for incomplete data
	if characteristics.DataCompleteness < 0.4 {
		score += 0.1
	}

	return score
}

// extractFloat64 safely extracts a float64 value from metadata
func (ms *ModuleSelector) extractFloat64(metadata map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := metadata[key]; exists {
		if floatValue, ok := value.(float64); ok {
			return floatValue
		}
	}
	return defaultValue
}

// Helper methods for module information
func (ms *ModuleSelector) getModuleType(module architecture.Module) string {
	metadata := module.Metadata()
	if metadata.Name != "" {
		// Extract type from name
		if strings.Contains(strings.ToLower(metadata.Name), "website") {
			return "website_analysis"
		} else if strings.Contains(strings.ToLower(metadata.Name), "web search") {
			return "web_search_analysis"
		} else if strings.Contains(strings.ToLower(metadata.Name), "ml") {
			return "ml_classification"
		} else if strings.Contains(strings.ToLower(metadata.Name), "keyword") {
			return "keyword_classification"
		}
	}
	return "generic"
}

func (ms *ModuleSelector) getCurrentLoad(moduleID string) int {
	// In a real implementation, this would query the module's current load
	// For now, return a placeholder value
	return 0
}

func (ms *ModuleSelector) getMaxConcurrency(module architecture.Module) int {
	// In a real implementation, this would get from module configuration
	// For now, return a default value
	return 10
}

func (ms *ModuleSelector) getModuleMetadata(module architecture.Module) map[string]interface{} {
	metadata := module.Metadata()
	return map[string]interface{}{
		"name":         metadata.Name,
		"version":      metadata.Version,
		"description":  metadata.Description,
		"capabilities": metadata.Capabilities,
		"priority":     metadata.Priority,
	}
}
