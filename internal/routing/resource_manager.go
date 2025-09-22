package routing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kyb-platform/internal/observability"
	"go.opentelemetry.io/otel/trace"
)

// ResourceManager manages load balancing and resource allocation for modules
type ResourceManager struct {
	logger  *observability.Logger
	tracer  trace.Tracer
	metrics *observability.Metrics

	// Resource tracking
	moduleResources map[string]*ModuleResourceInfo
	resourceMutex   sync.RWMutex

	// Load balancing state
	loadBalancer *LoadBalancer
	lbMutex      sync.RWMutex

	// Capacity planning
	capacityPlanner *CapacityPlanner
	cpMutex         sync.RWMutex

	// Health monitoring
	healthMonitor *HealthMonitor
	hmMutex       sync.RWMutex

	// Configuration
	config ResourceManagerConfig
}

// ResourceManagerConfig holds configuration for the resource manager
type ResourceManagerConfig struct {
	EnableLoadBalancing        bool                  `json:"enable_load_balancing"`
	EnableResourceMonitoring   bool                  `json:"enable_resource_monitoring"`
	EnableCapacityPlanning     bool                  `json:"enable_capacity_planning"`
	EnableHealthMonitoring     bool                  `json:"enable_health_monitoring"`
	EnableDynamicScaling       bool                  `json:"enable_dynamic_scaling"`
	LoadBalancingStrategy      LoadBalancingStrategy `json:"load_balancing_strategy"`
	ResourceUpdateInterval     time.Duration         `json:"resource_update_interval"`
	HealthCheckInterval        time.Duration         `json:"health_check_interval"`
	CapacityPlanningInterval   time.Duration         `json:"capacity_planning_interval"`
	MaxResourceUtilization     float64               `json:"max_resource_utilization"`
	MinResourceUtilization     float64               `json:"min_resource_utilization"`
	ScalingThreshold           float64               `json:"scaling_threshold"`
	HealthCheckTimeout         time.Duration         `json:"health_check_timeout"`
	EnableResourceOptimization bool                  `json:"enable_resource_optimization"`
}

// ModuleResourceInfo tracks resource usage for a module
type ModuleResourceInfo struct {
	ModuleID            string        `json:"module_id"`
	ModuleType          string        `json:"module_type"`
	CurrentLoad         int           `json:"current_load"`
	MaxConcurrency      int           `json:"max_concurrency"`
	CPUUsage            float64       `json:"cpu_usage"`
	MemoryUsage         float64       `json:"memory_usage"`
	DiskUsage           float64       `json:"disk_usage"`
	NetworkIO           float64       `json:"network_io"`
	ResponseTime        time.Duration `json:"response_time"`
	SuccessRate         float64       `json:"success_rate"`
	ErrorRate           float64       `json:"error_rate"`
	LastUpdated         time.Time     `json:"last_updated"`
	HealthStatus        HealthStatus  `json:"health_status"`
	IsAvailable         bool          `json:"is_available"`
	ResourceUtilization float64       `json:"resource_utilization"`
}

// HealthStatus represents the health status of a module
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
	HealthStatusUnknown   HealthStatus = "unknown"
)

// LoadBalancer manages load distribution across modules
type LoadBalancer struct {
	strategy LoadBalancingStrategy
	state    map[string]interface{}
	mutex    sync.RWMutex
}

// CapacityPlanner manages capacity planning and resource allocation
type CapacityPlanner struct {
	historicalData []*CapacityDataPoint
	forecasts      map[string]*CapacityForecast
	mutex          sync.RWMutex
}

// HealthMonitor monitors module health and availability
type HealthMonitor struct {
	healthChecks map[string]*HealthCheck
	alerts       []*HealthAlert
	mutex        sync.RWMutex
}

// CapacityDataPoint represents a historical capacity data point
type CapacityDataPoint struct {
	Timestamp     time.Time     `json:"timestamp"`
	ModuleID      string        `json:"module_id"`
	Load          int           `json:"load"`
	ResourceUsage float64       `json:"resource_usage"`
	ResponseTime  time.Duration `json:"response_time"`
	SuccessRate   float64       `json:"success_rate"`
}

// CapacityForecast represents a capacity forecast for a module
type CapacityForecast struct {
	ModuleID           string                `json:"module_id"`
	ForecastTime       time.Time             `json:"forecast_time"`
	PredictedLoad      int                   `json:"predicted_load"`
	PredictedUsage     float64               `json:"predicted_usage"`
	RecommendedScaling ScalingRecommendation `json:"recommended_scaling"`
	Confidence         float64               `json:"confidence"`
}

// ScalingRecommendation represents a scaling recommendation
type ScalingRecommendation struct {
	Action   ScalingAction `json:"action"`
	Reason   string        `json:"reason"`
	Priority int           `json:"priority"`
	Impact   float64       `json:"impact"`
}

// ScalingAction represents a scaling action
type ScalingAction string

const (
	ScalingActionNone      ScalingAction = "none"
	ScalingActionScaleUp   ScalingAction = "scale_up"
	ScalingActionScaleDown ScalingAction = "scale_down"
	ScalingActionMaintain  ScalingAction = "maintain"
)

// HealthCheck represents a health check for a module
type HealthCheck struct {
	ModuleID     string        `json:"module_id"`
	LastCheck    time.Time     `json:"last_check"`
	Status       HealthStatus  `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	ErrorCount   int           `json:"error_count"`
	SuccessCount int           `json:"success_count"`
	LastError    string        `json:"last_error"`
}

// HealthAlert represents a health alert
type HealthAlert struct {
	ID        string    `json:"id"`
	ModuleID  string    `json:"module_id"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Resolved  bool      `json:"resolved"`
}

// NewResourceManager creates a new resource manager
func NewResourceManager(
	logger *observability.Logger,
	tracer trace.Tracer,
	metrics *observability.Metrics,
	config ResourceManagerConfig,
) *ResourceManager {
	rm := &ResourceManager{
		logger:          logger,
		tracer:          tracer,
		metrics:         metrics,
		moduleResources: make(map[string]*ModuleResourceInfo),
		config:          config,
	}

	// Initialize components
	if config.EnableLoadBalancing {
		rm.loadBalancer = NewLoadBalancer(config.LoadBalancingStrategy)
	}

	if config.EnableCapacityPlanning {
		rm.capacityPlanner = NewCapacityPlanner()
	}

	if config.EnableHealthMonitoring {
		rm.healthMonitor = NewHealthMonitor()
	}

	// Start background tasks
	rm.startBackgroundTasks()

	return rm
}

// NewLoadBalancer creates a new load balancer
func NewLoadBalancer(strategy LoadBalancingStrategy) *LoadBalancer {
	return &LoadBalancer{
		strategy: strategy,
		state:    make(map[string]interface{}),
	}
}

// NewCapacityPlanner creates a new capacity planner
func NewCapacityPlanner() *CapacityPlanner {
	return &CapacityPlanner{
		historicalData: make([]*CapacityDataPoint, 0),
		forecasts:      make(map[string]*CapacityForecast),
	}
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor() *HealthMonitor {
	return &HealthMonitor{
		healthChecks: make(map[string]*HealthCheck),
		alerts:       make([]*HealthAlert, 0),
	}
}

// startBackgroundTasks starts background monitoring tasks
func (rm *ResourceManager) startBackgroundTasks() {
	if rm.config.EnableResourceMonitoring {
		go rm.monitorResources()
	}

	if rm.config.EnableHealthMonitoring {
		go rm.monitorHealth()
	}

	if rm.config.EnableCapacityPlanning {
		go rm.planCapacity()
	}
}

// monitorResources monitors resource usage for all modules
func (rm *ResourceManager) monitorResources() {
	ticker := time.NewTicker(rm.config.ResourceUpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		rm.updateResourceMetrics()
	}
}

// monitorHealth monitors health of all modules
func (rm *ResourceManager) monitorHealth() {
	ticker := time.NewTicker(rm.config.HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		rm.performHealthChecks()
	}
}

// planCapacity performs capacity planning
func (rm *ResourceManager) planCapacity() {
	ticker := time.NewTicker(rm.config.CapacityPlanningInterval)
	defer ticker.Stop()

	for range ticker.C {
		rm.updateCapacityForecasts()
	}
}

// UpdateModuleResource updates resource information for a module
func (rm *ResourceManager) UpdateModuleResource(moduleID string, info *ModuleResourceInfo) {
	rm.resourceMutex.Lock()
	defer rm.resourceMutex.Unlock()

	// Calculate resource utilization
	info.ResourceUtilization = rm.calculateResourceUtilization(info)
	info.LastUpdated = time.Now()

	rm.moduleResources[moduleID] = info

	// Record metrics
	if rm.metrics != nil {
		// TODO: Implement metrics recording when RecordHistogram is available
		// rm.metrics.RecordHistogram("module.resource.utilization", info.ResourceUtilization, map[string]string{
		// 	"module_id": moduleID,
		// 	"module_type": info.ModuleType,
		// })
		// rm.metrics.RecordHistogram("module.load.current", float64(info.CurrentLoad), map[string]string{
		// 	"module_id": moduleID,
		// })
		// rm.metrics.RecordHistogram("module.cpu.usage", info.CPUUsage, map[string]string{
		// 	"module_id": moduleID,
		// })
		// rm.metrics.RecordHistogram("module.memory.usage", info.MemoryUsage, map[string]string{
		// 	"module_id": moduleID,
		// })
	}

	rm.logger.WithComponent("resource_manager").Debug("module_resource_updated", map[string]interface{}{
		"module_id":    moduleID,
		"utilization":  info.ResourceUtilization,
		"load":         info.CurrentLoad,
		"cpu_usage":    info.CPUUsage,
		"memory_usage": info.MemoryUsage,
	})
}

// GetModuleResource gets resource information for a module
func (rm *ResourceManager) GetModuleResource(moduleID string) (*ModuleResourceInfo, bool) {
	rm.resourceMutex.RLock()
	defer rm.resourceMutex.RUnlock()

	info, exists := rm.moduleResources[moduleID]
	return info, exists
}

// GetAllModuleResources gets resource information for all modules
func (rm *ResourceManager) GetAllModuleResources() map[string]*ModuleResourceInfo {
	rm.resourceMutex.RLock()
	defer rm.resourceMutex.RUnlock()

	result := make(map[string]*ModuleResourceInfo)
	for id, info := range rm.moduleResources {
		result[id] = info
	}
	return result
}

// SelectBestModule selects the best module based on load balancing strategy
func (rm *ResourceManager) SelectBestModule(availableModules []string, requirements map[string]interface{}) (string, error) {
	if !rm.config.EnableLoadBalancing || rm.loadBalancer == nil {
		// Return first available module if load balancing is disabled
		if len(availableModules) > 0 {
			return availableModules[0], nil
		}
		return "", fmt.Errorf("no available modules")
	}

	rm.lbMutex.RLock()
	defer rm.lbMutex.RUnlock()

	// Get resource info for available modules
	var moduleInfos []*ModuleResourceInfo
	for _, moduleID := range availableModules {
		if info, exists := rm.GetModuleResource(moduleID); exists && info.IsAvailable {
			moduleInfos = append(moduleInfos, info)
		}
	}

	if len(moduleInfos) == 0 {
		return "", fmt.Errorf("no healthy modules available")
	}

	// Apply load balancing strategy
	selectedModule := rm.loadBalancer.SelectModule(moduleInfos, requirements)
	if selectedModule == nil {
		return "", fmt.Errorf("no suitable module found")
	}

	return selectedModule.ModuleID, nil
}

// SelectModule selects a module using the configured load balancing strategy
func (lb *LoadBalancer) SelectModule(modules []*ModuleResourceInfo, requirements map[string]interface{}) *ModuleResourceInfo {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	switch lb.strategy {
	case LoadBalancingStrategyRoundRobin:
		return lb.selectRoundRobin(modules)
	case LoadBalancingStrategyLeastLoaded:
		return lb.selectLeastLoaded(modules)
	case LoadBalancingStrategyBestPerformance:
		return lb.selectBestPerformance(modules)
	case LoadBalancingStrategyAdaptive:
		return lb.selectAdaptive(modules, requirements)
	default:
		return lb.selectLeastLoaded(modules) // Default to least loaded
	}
}

// selectRoundRobin selects module using round-robin strategy
func (lb *LoadBalancer) selectRoundRobin(modules []*ModuleResourceInfo) *ModuleResourceInfo {
	if len(modules) == 0 {
		return nil
	}

	// Get current round-robin state
	currentIndex := 0
	if state, exists := lb.state["round_robin_index"]; exists {
		if index, ok := state.(int); ok {
			currentIndex = index
		}
	}

	// Select module and update state
	selectedModule := modules[currentIndex%len(modules)]
	lb.state["round_robin_index"] = (currentIndex + 1) % len(modules)

	return selectedModule
}

// selectLeastLoaded selects the least loaded module
func (lb *LoadBalancer) selectLeastLoaded(modules []*ModuleResourceInfo) *ModuleResourceInfo {
	if len(modules) == 0 {
		return nil
	}

	leastLoaded := modules[0]
	lowestLoad := float64(leastLoaded.CurrentLoad) / float64(leastLoaded.MaxConcurrency)

	for _, module := range modules[1:] {
		load := float64(module.CurrentLoad) / float64(module.MaxConcurrency)
		if load < lowestLoad {
			leastLoaded = module
			lowestLoad = load
		}
	}

	return leastLoaded
}

// selectBestPerformance selects the module with best performance
func (lb *LoadBalancer) selectBestPerformance(modules []*ModuleResourceInfo) *ModuleResourceInfo {
	if len(modules) == 0 {
		return nil
	}

	bestModule := modules[0]
	bestScore := lb.calculatePerformanceScore(bestModule)

	for _, module := range modules[1:] {
		score := lb.calculatePerformanceScore(module)
		if score > bestScore {
			bestModule = module
			bestScore = score
		}
	}

	return bestModule
}

// selectAdaptive selects module using adaptive strategy
func (lb *LoadBalancer) selectAdaptive(modules []*ModuleResourceInfo, requirements map[string]interface{}) *ModuleResourceInfo {
	if len(modules) == 0 {
		return nil
	}

	bestModule := modules[0]
	bestScore := lb.calculateAdaptiveScore(bestModule, requirements)

	for _, module := range modules[1:] {
		score := lb.calculateAdaptiveScore(module, requirements)
		if score > bestScore {
			bestModule = module
			bestScore = score
		}
	}

	return bestModule
}

// calculatePerformanceScore calculates performance score for a module
func (lb *LoadBalancer) calculatePerformanceScore(module *ModuleResourceInfo) float64 {
	// Performance score based on response time and success rate
	responseTimeScore := 1.0 - (float64(module.ResponseTime.Milliseconds()) / 5000.0) // Normalize to 5 seconds
	if responseTimeScore < 0 {
		responseTimeScore = 0
	}

	successRateScore := module.SuccessRate / 100.0

	// Weighted combination
	return (responseTimeScore * 0.6) + (successRateScore * 0.4)
}

// calculateAdaptiveScore calculates adaptive selection score
func (lb *LoadBalancer) calculateAdaptiveScore(module *ModuleResourceInfo, requirements map[string]interface{}) float64 {
	score := 0.0

	// Performance weight (40%)
	performanceScore := lb.calculatePerformanceScore(module)
	score += performanceScore * 0.4

	// Load weight (30%)
	loadScore := 1.0 - (float64(module.CurrentLoad) / float64(module.MaxConcurrency))
	score += loadScore * 0.3

	// Resource utilization weight (20%)
	resourceScore := 1.0 - module.ResourceUtilization
	score += resourceScore * 0.2

	// Health weight (10%)
	healthScore := 0.0
	switch module.HealthStatus {
	case HealthStatusHealthy:
		healthScore = 1.0
	case HealthStatusDegraded:
		healthScore = 0.5
	case HealthStatusUnhealthy:
		healthScore = 0.0
	default:
		healthScore = 0.5
	}
	score += healthScore * 0.1

	return score
}

// calculateResourceUtilization calculates resource utilization for a module
func (rm *ResourceManager) calculateResourceUtilization(info *ModuleResourceInfo) float64 {
	// Weighted combination of CPU, memory, and load
	cpuWeight := 0.4
	memoryWeight := 0.3
	loadWeight := 0.3

	cpuUtil := info.CPUUsage / 100.0
	memoryUtil := info.MemoryUsage / 100.0
	loadUtil := float64(info.CurrentLoad) / float64(info.MaxConcurrency)

	return (cpuUtil * cpuWeight) + (memoryUtil * memoryWeight) + (loadUtil * loadWeight)
}

// updateResourceMetrics updates resource metrics
func (rm *ResourceManager) updateResourceMetrics() {
	ctx, span := rm.tracer.Start(context.Background(), "ResourceManager.updateResourceMetrics")
	defer span.End()

	rm.resourceMutex.RLock()
	modules := make(map[string]*ModuleResourceInfo)
	for id, info := range rm.moduleResources {
		modules[id] = info
	}
	rm.resourceMutex.RUnlock()

	for moduleID, info := range modules {
		// Update metrics
		if rm.metrics != nil {
			// TODO: Implement metrics recording when RecordHistogram is available
			// rm.metrics.RecordHistogram("module.resource.utilization", info.ResourceUtilization, map[string]string{
			// 	"module_id": moduleID,
			// })
		}

		// Check for scaling recommendations
		if rm.config.EnableDynamicScaling {
			rm.checkScalingRecommendations(ctx, moduleID, info)
		}
	}
}

// performHealthChecks performs health checks for all modules
func (rm *ResourceManager) performHealthChecks() {
	ctx, span := rm.tracer.Start(context.Background(), "ResourceManager.performHealthChecks")
	defer span.End()

	rm.resourceMutex.RLock()
	modules := make(map[string]*ModuleResourceInfo)
	for id, info := range rm.moduleResources {
		modules[id] = info
	}
	rm.resourceMutex.RUnlock()

	for moduleID, info := range modules {
		healthStatus := rm.checkModuleHealth(ctx, moduleID, info)

		// Update health status
		rm.resourceMutex.Lock()
		if existingInfo, exists := rm.moduleResources[moduleID]; exists {
			existingInfo.HealthStatus = healthStatus
			existingInfo.IsAvailable = healthStatus == HealthStatusHealthy || healthStatus == HealthStatusDegraded
		}
		rm.resourceMutex.Unlock()

		// Create health alert if needed
		if healthStatus == HealthStatusUnhealthy {
			rm.createHealthAlert(moduleID, "module_unhealthy", "high", "Module is unhealthy")
		}
	}
}

// checkModuleHealth checks health of a specific module
func (rm *ResourceManager) checkModuleHealth(ctx context.Context, moduleID string, info *ModuleResourceInfo) HealthStatus {
	// Simple health check based on resource utilization and error rate
	if info.ErrorRate > 0.1 { // More than 10% error rate
		return HealthStatusUnhealthy
	}

	if info.ResourceUtilization > rm.config.MaxResourceUtilization {
		return HealthStatusDegraded
	}

	if info.ResponseTime > 5*time.Second { // More than 5 seconds response time
		return HealthStatusDegraded
	}

	return HealthStatusHealthy
}

// createHealthAlert creates a health alert
func (rm *ResourceManager) createHealthAlert(moduleID, alertType, severity, message string) {
	if rm.healthMonitor == nil {
		return
	}

	alert := &HealthAlert{
		ID:        fmt.Sprintf("health_%s_%d", moduleID, time.Now().Unix()),
		ModuleID:  moduleID,
		Type:      alertType,
		Severity:  severity,
		Message:   message,
		Timestamp: time.Now(),
		Resolved:  false,
	}

	rm.healthMonitor.mutex.Lock()
	rm.healthMonitor.alerts = append(rm.healthMonitor.alerts, alert)
	rm.healthMonitor.mutex.Unlock()

	rm.logger.WithComponent("resource_manager").Warn("health_alert_created", map[string]interface{}{
		"module_id":  moduleID,
		"alert_type": alertType,
		"severity":   severity,
		"message":    message,
	})
}

// checkScalingRecommendations checks for scaling recommendations
func (rm *ResourceManager) checkScalingRecommendations(ctx context.Context, moduleID string, info *ModuleResourceInfo) {
	if info.ResourceUtilization > rm.config.ScalingThreshold {
		// Recommend scale up
		rm.logger.WithComponent("resource_manager").Info("scaling_recommendation", map[string]interface{}{
			"module_id":   moduleID,
			"action":      "scale_up",
			"reason":      "high_resource_utilization",
			"utilization": info.ResourceUtilization,
			"threshold":   rm.config.ScalingThreshold,
		})
	} else if info.ResourceUtilization < rm.config.MinResourceUtilization {
		// Recommend scale down
		rm.logger.WithComponent("resource_manager").Info("scaling_recommendation", map[string]interface{}{
			"module_id":   moduleID,
			"action":      "scale_down",
			"reason":      "low_resource_utilization",
			"utilization": info.ResourceUtilization,
			"threshold":   rm.config.MinResourceUtilization,
		})
	}
}

// updateCapacityForecasts updates capacity forecasts
func (rm *ResourceManager) updateCapacityForecasts() {
	if rm.capacityPlanner == nil {
		return
	}

	_, span := rm.tracer.Start(context.Background(), "ResourceManager.updateCapacityForecasts")
	defer span.End()

	// Collect current capacity data
	rm.resourceMutex.RLock()
	modules := make(map[string]*ModuleResourceInfo)
	for id, info := range rm.moduleResources {
		modules[id] = info
	}
	rm.resourceMutex.RUnlock()

	for moduleID, info := range modules {
		// Add data point
		dataPoint := &CapacityDataPoint{
			Timestamp:     time.Now(),
			ModuleID:      moduleID,
			Load:          info.CurrentLoad,
			ResourceUsage: info.ResourceUtilization,
			ResponseTime:  info.ResponseTime,
			SuccessRate:   info.SuccessRate,
		}

		rm.capacityPlanner.AddDataPoint(dataPoint)

		// Generate forecast
		forecast := rm.capacityPlanner.GenerateForecast(moduleID)
		if forecast != nil {
			rm.logger.WithComponent("resource_manager").Debug("capacity_forecast_generated", map[string]interface{}{
				"module_id":          moduleID,
				"predicted_load":     forecast.PredictedLoad,
				"predicted_usage":    forecast.PredictedUsage,
				"recommended_action": forecast.RecommendedScaling.Action,
				"confidence":         forecast.Confidence,
			})
		}
	}
}

// AddDataPoint adds a capacity data point
func (cp *CapacityPlanner) AddDataPoint(dataPoint *CapacityDataPoint) {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	cp.historicalData = append(cp.historicalData, dataPoint)

	// Keep only last 1000 data points
	if len(cp.historicalData) > 1000 {
		cp.historicalData = cp.historicalData[len(cp.historicalData)-1000:]
	}
}

// GenerateForecast generates a capacity forecast for a module
func (cp *CapacityPlanner) GenerateForecast(moduleID string) *CapacityForecast {
	cp.mutex.Lock()
	defer cp.mutex.Unlock()

	// Simple forecasting based on recent trends
	// In a real implementation, this would use more sophisticated algorithms
	recentData := cp.getRecentData(moduleID, 10) // Last 10 data points
	if len(recentData) < 3 {
		return nil
	}

	// Calculate average load and usage
	var totalLoad int
	var totalUsage float64
	for _, data := range recentData {
		totalLoad += data.Load
		totalUsage += data.ResourceUsage
	}

	avgLoad := totalLoad / len(recentData)
	avgUsage := totalUsage / float64(len(recentData))

	// Simple trend calculation
	trend := cp.calculateTrend(recentData)

	// Predict next values
	predictedLoad := avgLoad + int(float64(avgLoad)*trend)
	predictedUsage := avgUsage + (avgUsage * trend)

	// Generate scaling recommendation
	recommendation := cp.generateScalingRecommendation(predictedLoad, predictedUsage)

	forecast := &CapacityForecast{
		ModuleID:           moduleID,
		ForecastTime:       time.Now(),
		PredictedLoad:      predictedLoad,
		PredictedUsage:     predictedUsage,
		RecommendedScaling: recommendation,
		Confidence:         0.8, // Simple confidence calculation
	}

	cp.forecasts[moduleID] = forecast
	return forecast
}

// getRecentData gets recent data points for a module
func (cp *CapacityPlanner) getRecentData(moduleID string, count int) []*CapacityDataPoint {
	var moduleData []*CapacityDataPoint
	for _, data := range cp.historicalData {
		if data.ModuleID == moduleID {
			moduleData = append(moduleData, data)
		}
	}

	if len(moduleData) <= count {
		return moduleData
	}

	return moduleData[len(moduleData)-count:]
}

// calculateTrend calculates trend from recent data
func (cp *CapacityPlanner) calculateTrend(data []*CapacityDataPoint) float64 {
	if len(data) < 2 {
		return 0.0
	}

	// Simple linear trend calculation
	first := data[0]
	last := data[len(data)-1]

	loadDiff := float64(last.Load - first.Load)
	timeDiff := float64(last.Timestamp.Sub(first.Timestamp).Hours())

	if timeDiff == 0 {
		return 0.0
	}

	return loadDiff / timeDiff / float64(first.Load) // Normalized trend
}

// generateScalingRecommendation generates a scaling recommendation
func (cp *CapacityPlanner) generateScalingRecommendation(predictedLoad int, predictedUsage float64) ScalingRecommendation {
	if predictedUsage > 0.8 { // 80% utilization threshold
		return ScalingRecommendation{
			Action:   ScalingActionScaleUp,
			Reason:   "predicted_high_utilization",
			Priority: 1,
			Impact:   0.8,
		}
	} else if predictedUsage < 0.2 { // 20% utilization threshold
		return ScalingRecommendation{
			Action:   ScalingActionScaleDown,
			Reason:   "predicted_low_utilization",
			Priority: 2,
			Impact:   0.6,
		}
	}

	return ScalingRecommendation{
		Action:   ScalingActionMaintain,
		Reason:   "predicted_optimal_utilization",
		Priority: 3,
		Impact:   0.0,
	}
}

// GetResourceSummary returns a summary of resource usage
func (rm *ResourceManager) GetResourceSummary() map[string]interface{} {
	rm.resourceMutex.RLock()
	defer rm.resourceMutex.RUnlock()

	totalModules := len(rm.moduleResources)
	healthyModules := 0
	totalLoad := 0
	totalUtilization := 0.0

	for _, info := range rm.moduleResources {
		if info.HealthStatus == HealthStatusHealthy {
			healthyModules++
		}
		totalLoad += info.CurrentLoad
		totalUtilization += info.ResourceUtilization
	}

	avgUtilization := 0.0
	if totalModules > 0 {
		avgUtilization = totalUtilization / float64(totalModules)
	}

	return map[string]interface{}{
		"total_modules":       totalModules,
		"healthy_modules":     healthyModules,
		"total_load":          totalLoad,
		"avg_utilization":     avgUtilization,
		"load_balancing":      rm.config.EnableLoadBalancing,
		"resource_monitoring": rm.config.EnableResourceMonitoring,
		"health_monitoring":   rm.config.EnableHealthMonitoring,
		"capacity_planning":   rm.config.EnableCapacityPlanning,
	}
}
