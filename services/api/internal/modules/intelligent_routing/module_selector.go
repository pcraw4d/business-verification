package intelligent_routing

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ModuleSelectorConfig represents the configuration for the module selector
type ModuleSelectorConfig struct {
	MaxModulesPerRequest   int           `json:"max_modules_per_request"`
	MinConfidenceThreshold float64       `json:"min_confidence_threshold"`
	LoadBalancingEnabled   bool          `json:"load_balancing_enabled"`
	HealthCheckEnabled     bool          `json:"health_check_enabled"`
	SpecializationWeight   float64       `json:"specialization_weight"`
	PerformanceWeight      float64       `json:"performance_weight"`
	AvailabilityWeight     float64       `json:"availability_weight"`
	LoadWeight             float64       `json:"load_weight"`
	SelectionTimeout       time.Duration `json:"selection_timeout"`
	EnableLearning         bool          `json:"enable_learning"`
}

// moduleSelector implements the ModuleSelector interface
type moduleSelector struct {
	config *ModuleSelectorConfig
	logger *zap.Logger

	// Module registry
	modules      map[string]*ModuleCapability
	modulesMutex sync.RWMutex

	// Health checker for module availability
	healthChecker HealthChecker

	// Load balancer for load distribution
	loadBalancer LoadBalancer

	// Learning data for optimization
	learningData  map[string]*ModuleLearningData
	learningMutex sync.RWMutex
}

// ModuleLearningData represents learning data for module optimization
type ModuleLearningData struct {
	ModuleID           string                `json:"module_id"`
	SuccessCount       int64                 `json:"success_count"`
	FailureCount       int64                 `json:"failure_count"`
	AverageLatency     float64               `json:"average_latency"`
	LastUsed           time.Time             `json:"last_used"`
	RequestTypeSuccess map[RequestType]int64 `json:"request_type_success"`
	IndustrySuccess    map[string]int64      `json:"industry_success"`
}

// NewModuleSelector creates a new module selector instance
func NewModuleSelector(config *ModuleSelectorConfig, healthChecker HealthChecker, loadBalancer LoadBalancer, logger *zap.Logger) ModuleSelector {
	if config == nil {
		config = &ModuleSelectorConfig{
			MaxModulesPerRequest:   3,
			MinConfidenceThreshold: 0.7,
			LoadBalancingEnabled:   true,
			HealthCheckEnabled:     true,
			SpecializationWeight:   0.3,
			PerformanceWeight:      0.3,
			AvailabilityWeight:     0.2,
			LoadWeight:             0.2,
			SelectionTimeout:       10 * time.Second,
			EnableLearning:         true,
		}
	}

	selector := &moduleSelector{
		config:        config,
		logger:        logger,
		modules:       make(map[string]*ModuleCapability),
		healthChecker: healthChecker,
		loadBalancer:  loadBalancer,
		learningData:  make(map[string]*ModuleLearningData),
	}

	return selector
}

// SelectModules selects the best modules for processing a request
func (ms *moduleSelector) SelectModules(ctx context.Context, request *VerificationRequest, analysis *RequestAnalysis) ([]*ModuleCapability, error) {
	ms.logger.Info("Starting module selection",
		zap.String("request_id", request.ID),
		zap.String("request_type", string(analysis.Classification.RequestType)))

	// Create selection context with timeout
	selectionCtx, cancel := context.WithTimeout(ctx, ms.config.SelectionTimeout)
	defer cancel()

	// Get available modules
	availableModules, err := ms.getAvailableModules(selectionCtx)
	if err != nil {
		return nil, fmt.Errorf("failed to get available modules: %w", err)
	}

	// Filter modules by request type and capabilities
	candidateModules := ms.filterModulesByCapabilities(availableModules, request, analysis)
	if len(candidateModules) == 0 {
		return nil, fmt.Errorf("no suitable modules found for request type %s", analysis.Classification.RequestType)
	}

	// Score and rank modules
	scoredModules := ms.scoreModules(candidateModules, request, analysis)

	// Convert scored modules to regular modules
	modules := make([]*ModuleCapability, len(scoredModules))
	for i, scored := range scoredModules {
		modules[i] = scored.module
	}

	// Optimize selection
	optimizedModules, err := ms.OptimizeSelection(selectionCtx, modules, request)
	if err != nil {
		return nil, fmt.Errorf("optimization failed: %w", err)
	}

	// Apply load balancing if enabled
	if ms.config.LoadBalancingEnabled {
		optimizedModules, err = ms.LoadBalance(selectionCtx, optimizedModules)
		if err != nil {
			return nil, fmt.Errorf("load balancing failed: %w", err)
		}
	}

	// Limit number of modules
	if len(optimizedModules) > ms.config.MaxModulesPerRequest {
		optimizedModules = optimizedModules[:ms.config.MaxModulesPerRequest]
	}

	// Update learning data
	if ms.config.EnableLearning {
		ms.updateLearningData(request, optimizedModules)
	}

	ms.logger.Info("Module selection completed",
		zap.String("request_id", request.ID),
		zap.Int("selected_modules", len(optimizedModules)))

	return optimizedModules, nil
}

// OptimizeSelection optimizes the module selection based on various factors
func (ms *moduleSelector) OptimizeSelection(ctx context.Context, candidates []*ModuleCapability, request *VerificationRequest) ([]*ModuleCapability, error) {
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no candidate modules provided")
	}

	// Score candidates
	scoredModules := ms.scoreModules(candidates, request, nil)

	// Sort by score (highest first)
	sort.Slice(scoredModules, func(i, j int) bool {
		return scoredModules[i].score > scoredModules[j].score
	})

	// Return top modules
	maxModules := ms.config.MaxModulesPerRequest
	if len(scoredModules) < maxModules {
		maxModules = len(scoredModules)
	}

	result := make([]*ModuleCapability, maxModules)
	for i := 0; i < maxModules; i++ {
		result[i] = scoredModules[i].module
	}

	return result, nil
}

// LoadBalance distributes load across available modules
func (ms *moduleSelector) LoadBalance(ctx context.Context, modules []*ModuleCapability) ([]*ModuleCapability, error) {
	if len(modules) <= 1 {
		return modules, nil
	}

	// Get current load for each module
	loadDistribution := make(map[string]float64)
	for _, module := range modules {
		load, err := ms.loadBalancer.GetModuleLoad(ctx, module.ModuleID)
		if err != nil {
			ms.logger.Warn("Failed to get module load, using default",
				zap.String("module_id", module.ModuleID),
				zap.Error(err))
			load = 0.5 // Default load
		}
		loadDistribution[module.ModuleID] = load
	}

	// Sort modules by load (lowest first for better distribution)
	sort.Slice(modules, func(i, j int) bool {
		return loadDistribution[modules[i].ModuleID] < loadDistribution[modules[j].ModuleID]
	})

	return modules, nil
}

// RegisterModule registers a new module with the selector
func (ms *moduleSelector) RegisterModule(module *ModuleCapability) error {
	ms.modulesMutex.Lock()
	defer ms.modulesMutex.Unlock()

	if module.ModuleID == "" {
		return fmt.Errorf("module ID cannot be empty")
	}

	ms.modules[module.ModuleID] = module

	// Initialize learning data
	if ms.config.EnableLearning {
		ms.learningMutex.Lock()
		ms.learningData[module.ModuleID] = &ModuleLearningData{
			ModuleID:           module.ModuleID,
			RequestTypeSuccess: make(map[RequestType]int64),
			IndustrySuccess:    make(map[string]int64),
		}
		ms.learningMutex.Unlock()
	}

	ms.logger.Info("Module registered",
		zap.String("module_id", module.ModuleID),
		zap.String("module_name", module.ModuleName))

	return nil
}

// UnregisterModule removes a module from the selector
func (ms *moduleSelector) UnregisterModule(moduleID string) error {
	ms.modulesMutex.Lock()
	defer ms.modulesMutex.Unlock()

	if _, exists := ms.modules[moduleID]; !exists {
		return fmt.Errorf("module %s not found", moduleID)
	}

	delete(ms.modules, moduleID)

	// Remove learning data
	if ms.config.EnableLearning {
		ms.learningMutex.Lock()
		delete(ms.learningData, moduleID)
		ms.learningMutex.Unlock()
	}

	ms.logger.Info("Module unregistered", zap.String("module_id", moduleID))
	return nil
}

// GetModuleCapabilities returns all registered module capabilities
func (ms *moduleSelector) GetModuleCapabilities() []*ModuleCapability {
	ms.modulesMutex.RLock()
	defer ms.modulesMutex.RUnlock()

	capabilities := make([]*ModuleCapability, 0, len(ms.modules))
	for _, module := range ms.modules {
		capabilities = append(capabilities, module)
	}

	return capabilities
}

// Helper methods

func (ms *moduleSelector) getAvailableModules(ctx context.Context) ([]*ModuleCapability, error) {
	ms.modulesMutex.RLock()
	allModules := make([]*ModuleCapability, 0, len(ms.modules))
	for _, module := range ms.modules {
		allModules = append(allModules, module)
	}
	ms.modulesMutex.RUnlock()

	if !ms.config.HealthCheckEnabled {
		return allModules, nil
	}

	// Filter by health status
	var availableModules []*ModuleCapability
	for _, module := range allModules {
		health, err := ms.healthChecker.CheckHealth(ctx, module.ModuleID)
		if err != nil {
			ms.logger.Warn("Failed to check module health, assuming available",
				zap.String("module_id", module.ModuleID),
				zap.Error(err))
			availableModules = append(availableModules, module)
			continue
		}

		if health.IsAvailable && health.HealthScore >= ms.config.MinConfidenceThreshold {
			// Update module availability
			module.Availability = *health
			availableModules = append(availableModules, module)
		} else {
			ms.logger.Debug("Module not available",
				zap.String("module_id", module.ModuleID),
				zap.Bool("is_available", health.IsAvailable),
				zap.Float64("health_score", health.HealthScore))
		}
	}

	return availableModules, nil
}

func (ms *moduleSelector) filterModulesByCapabilities(modules []*ModuleCapability, request *VerificationRequest, analysis *RequestAnalysis) []*ModuleCapability {
	var candidates []*ModuleCapability

	for _, module := range modules {
		// Check if module supports the request type
		if !ms.moduleSupportsRequestType(module, analysis.Classification.RequestType) {
			continue
		}

		// Check if module can handle the complexity
		if !ms.moduleCanHandleComplexity(module, analysis.Complexity) {
			continue
		}

		// Check if module has required capabilities
		if !ms.moduleHasRequiredCapabilities(module, request) {
			continue
		}

		candidates = append(candidates, module)
	}

	return candidates
}

func (ms *moduleSelector) moduleSupportsRequestType(module *ModuleCapability, requestType RequestType) bool {
	for _, supportedType := range module.RequestTypes {
		if supportedType == requestType {
			return true
		}
	}
	return false
}

func (ms *moduleSelector) moduleCanHandleComplexity(module *ModuleCapability, complexity RequestComplexity) bool {
	// Simple complexity mapping
	complexityLevels := map[RequestComplexity]int{
		ComplexitySimple:   1,
		ComplexityModerate: 2,
		ComplexityComplex:  3,
		ComplexityAdvanced: 4,
	}

	moduleLevels := map[RequestComplexity]int{
		ComplexitySimple:   1,
		ComplexityModerate: 2,
		ComplexityComplex:  3,
		ComplexityAdvanced: 4,
	}

	requestLevel := complexityLevels[complexity]
	moduleLevel := moduleLevels[module.Complexity]

	return moduleLevel >= requestLevel
}

func (ms *moduleSelector) moduleHasRequiredCapabilities(module *ModuleCapability, request *VerificationRequest) bool {
	// Check for basic capabilities
	requiredCapabilities := []string{"verification", "validation"}

	for _, required := range requiredCapabilities {
		found := false
		for _, capability := range module.Capabilities {
			if capability == required {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check for industry-specific capabilities if industry is specified
	if request.Industry != "" {
		industryCapability := fmt.Sprintf("%s_verification", request.Industry)
		found := false
		for _, capability := range module.Capabilities {
			if capability == industryCapability {
				found = true
				break
			}
		}
		if !found {
			// Check if module has general industry capability
			for _, capability := range module.Capabilities {
				if capability == "industry_verification" {
					found = true
					break
				}
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// scoredModule represents a module with its selection score
type scoredModule struct {
	module *ModuleCapability
	score  float64
}

func (ms *moduleSelector) scoreModules(modules []*ModuleCapability, request *VerificationRequest, analysis *RequestAnalysis) []scoredModule {
	var scoredModules []scoredModule

	for _, module := range modules {
		score := ms.calculateModuleScore(module, request, analysis)
		scoredModules = append(scoredModules, scoredModule{
			module: module,
			score:  score,
		})
	}

	// Sort by score (highest first)
	sort.Slice(scoredModules, func(i, j int) bool {
		return scoredModules[i].score > scoredModules[j].score
	})

	return scoredModules
}

func (ms *moduleSelector) calculateModuleScore(module *ModuleCapability, request *VerificationRequest, analysis *RequestAnalysis) float64 {
	score := 0.0

	// Specialization score
	specializationScore := ms.calculateSpecializationScore(module, request, analysis)
	score += specializationScore * ms.config.SpecializationWeight

	// Performance score
	performanceScore := ms.calculatePerformanceScore(module)
	score += performanceScore * ms.config.PerformanceWeight

	// Availability score
	availabilityScore := ms.calculateAvailabilityScore(module)
	score += availabilityScore * ms.config.AvailabilityWeight

	// Load score (lower load is better)
	loadScore := 1.0 - module.Availability.LoadPercentage
	score += loadScore * ms.config.LoadWeight

	// Learning-based score
	if ms.config.EnableLearning {
		learningScore := ms.calculateLearningScore(module, request, analysis)
		score += learningScore * 0.1 // Small weight for learning
	}

	return score
}

func (ms *moduleSelector) calculateSpecializationScore(module *ModuleCapability, request *VerificationRequest, analysis *RequestAnalysis) float64 {
	score := 0.5 // Base score

	// Industry specialization
	if analysis != nil && analysis.Classification != nil && analysis.Classification.Industry != "" {
		if industryScore, exists := module.Specialization[analysis.Classification.Industry]; exists {
			score += industryScore * 0.3
		}
	}

	// Request type specialization
	if analysis != nil && analysis.Classification != nil {
		requestTypeKey := string(analysis.Classification.RequestType)
		if typeScore, exists := module.Specialization[requestTypeKey]; exists {
			score += typeScore * 0.2
		}
	}

	// Geographic specialization
	if analysis != nil && analysis.Classification != nil && analysis.Classification.GeographicRegion != "" {
		regionKey := fmt.Sprintf("region_%s", analysis.Classification.GeographicRegion)
		if regionScore, exists := module.Specialization[regionKey]; exists {
			score += regionScore * 0.1
		}
	}

	return math.Min(score, 1.0)
}

func (ms *moduleSelector) calculatePerformanceScore(module *ModuleCapability) float64 {
	score := 0.0

	// Success rate (0-1)
	score += module.Performance.SuccessRate * 0.4

	// Latency score (lower is better, normalized to 0-1)
	latencyScore := 1.0 - math.Min(module.Performance.AverageLatency/1000.0, 1.0) // Normalize to 1 second
	score += latencyScore * 0.3

	// Throughput score (normalized to 0-1)
	throughputScore := math.Min(module.Performance.Throughput/100.0, 1.0) // Normalize to 100 req/sec
	score += throughputScore * 0.2

	// Error rate (lower is better)
	errorScore := 1.0 - module.Performance.ErrorRate
	score += errorScore * 0.1

	return score
}

func (ms *moduleSelector) calculateAvailabilityScore(module *ModuleCapability) float64 {
	score := 0.0

	// Health score
	score += module.Availability.HealthScore * 0.6

	// Availability status
	if module.Availability.IsAvailable {
		score += 0.3
	}

	// Queue length (shorter is better)
	queueScore := 1.0 - math.Min(float64(module.Availability.QueueLength)/10.0, 1.0)
	score += queueScore * 0.1

	return score
}

func (ms *moduleSelector) calculateLearningScore(module *ModuleCapability, request *VerificationRequest, analysis *RequestAnalysis) float64 {
	ms.learningMutex.RLock()
	learningData, exists := ms.learningData[module.ModuleID]
	ms.learningMutex.RUnlock()

	if !exists {
		return 0.5 // Neutral score for new modules
	}

	score := 0.0

	// Overall success rate
	totalRequests := learningData.SuccessCount + learningData.FailureCount
	if totalRequests > 0 {
		successRate := float64(learningData.SuccessCount) / float64(totalRequests)
		score += successRate * 0.4
	}

	// Request type success rate
	if analysis != nil {
		requestTypeSuccess := learningData.RequestTypeSuccess[analysis.Classification.RequestType]
		if requestTypeSuccess > 0 {
			score += 0.3
		}
	}

	// Industry success rate
	if analysis != nil && analysis.Classification.Industry != "" {
		industrySuccess := learningData.IndustrySuccess[analysis.Classification.Industry]
		if industrySuccess > 0 {
			score += 0.2
		}
	}

	// Recency bonus (recently used modules get a small bonus)
	timeSinceLastUse := time.Since(learningData.LastUsed)
	if timeSinceLastUse < 1*time.Hour {
		score += 0.1
	}

	return score
}

func (ms *moduleSelector) updateLearningData(request *VerificationRequest, selectedModules []*ModuleCapability) {
	ms.learningMutex.Lock()
	defer ms.learningMutex.Unlock()

	for _, module := range selectedModules {
		learningData, exists := ms.learningData[module.ModuleID]
		if !exists {
			learningData = &ModuleLearningData{
				ModuleID:           module.ModuleID,
				RequestTypeSuccess: make(map[RequestType]int64),
				IndustrySuccess:    make(map[string]int64),
			}
			ms.learningData[module.ModuleID] = learningData
		}

		// Update last used time
		learningData.LastUsed = time.Now()

		// Update request type success (assuming success for now)
		learningData.RequestTypeSuccess[request.RequestType]++

		// Update industry success if available
		if request.Industry != "" {
			learningData.IndustrySuccess[request.Industry]++
		}
	}
}
