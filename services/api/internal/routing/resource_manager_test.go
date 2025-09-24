package routing

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/observability"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace"
)

func TestNewResourceManager(t *testing.T) {
	logger := observability.NewLogger(nil)
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	metrics, _ := observability.NewMetrics(nil)

	config := ResourceManagerConfig{
		EnableLoadBalancing:        true,
		EnableResourceMonitoring:   true,
		EnableHealthMonitoring:     true,
		EnableCapacityPlanning:     true,
		LoadBalancingStrategy:      LoadBalancingStrategyAdaptive,
		ResourceUpdateInterval:     30 * time.Second,
		HealthCheckInterval:        60 * time.Second,
		CapacityPlanningInterval:   300 * time.Second,
		MaxResourceUtilization:     0.8,
		MinResourceUtilization:     0.2,
		ScalingThreshold:           0.7,
		HealthCheckTimeout:         10 * time.Second,
		EnableResourceOptimization: true,
	}

	rm := NewResourceManager(logger, tracer, metrics, config)

	assert.NotNil(t, rm)
	assert.NotNil(t, rm.loadBalancer)
	assert.NotNil(t, rm.capacityPlanner)
	assert.NotNil(t, rm.healthMonitor)
	assert.Equal(t, config, rm.config)
}

func TestResourceManager_UpdateModuleResource(t *testing.T) {
	rm := createTestResourceManager(t)

	info := &ModuleResourceInfo{
		ModuleID:       "test-module-1",
		ModuleType:     "classification",
		CurrentLoad:    5,
		MaxConcurrency: 10,
		CPUUsage:       45.5,
		MemoryUsage:    60.2,
		DiskUsage:      25.0,
		NetworkIO:      10.5,
		ResponseTime:   2 * time.Second,
		SuccessRate:    95.5,
		ErrorRate:      2.1,
		HealthStatus:   HealthStatusHealthy,
		IsAvailable:    true,
	}

	rm.UpdateModuleResource("test-module-1", info)

	// Verify resource was stored
	retrievedInfo, exists := rm.GetModuleResource("test-module-1")
	assert.True(t, exists)
	assert.Equal(t, info.ModuleID, retrievedInfo.ModuleID)
	assert.Equal(t, info.ModuleType, retrievedInfo.ModuleType)
	assert.Equal(t, info.CurrentLoad, retrievedInfo.CurrentLoad)
	assert.Equal(t, info.MaxConcurrency, retrievedInfo.MaxConcurrency)
	assert.Equal(t, info.CPUUsage, retrievedInfo.CPUUsage)
	assert.Equal(t, info.MemoryUsage, retrievedInfo.MemoryUsage)
	assert.True(t, retrievedInfo.LastUpdated.After(time.Now().Add(-time.Second)))
	assert.Greater(t, retrievedInfo.ResourceUtilization, 0.0)
}

func TestResourceManager_SelectBestModule(t *testing.T) {
	rm := createTestResourceManager(t)

	// Add multiple modules with different loads
	module1 := &ModuleResourceInfo{
		ModuleID:       "module-1",
		ModuleType:     "classification",
		CurrentLoad:    2,
		MaxConcurrency: 10,
		CPUUsage:       30.0,
		MemoryUsage:    40.0,
		ResponseTime:   1 * time.Second,
		SuccessRate:    98.0,
		ErrorRate:      1.0,
		HealthStatus:   HealthStatusHealthy,
		IsAvailable:    true,
	}

	module2 := &ModuleResourceInfo{
		ModuleID:       "module-2",
		ModuleType:     "classification",
		CurrentLoad:    8,
		MaxConcurrency: 10,
		CPUUsage:       70.0,
		MemoryUsage:    80.0,
		ResponseTime:   3 * time.Second,
		SuccessRate:    92.0,
		ErrorRate:      5.0,
		HealthStatus:   HealthStatusHealthy,
		IsAvailable:    true,
	}

	rm.UpdateModuleResource("module-1", module1)
	rm.UpdateModuleResource("module-2", module2)

	// Test least loaded strategy
	rm.config.LoadBalancingStrategy = LoadBalancingStrategyLeastLoaded
	selectedModule, err := rm.SelectBestModule([]string{"module-1", "module-2"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "module-1", selectedModule) // Should select module-1 (less loaded)

	// Test best performance strategy
	rm.config.LoadBalancingStrategy = LoadBalancingStrategyBestPerformance
	selectedModule, err = rm.SelectBestModule([]string{"module-1", "module-2"}, nil)
	assert.NoError(t, err)
	assert.Equal(t, "module-1", selectedModule) // Should select module-1 (better performance)
}

func TestLoadBalancer_SelectModule(t *testing.T) {
	lb := NewLoadBalancer(LoadBalancingStrategyAdaptive)

	modules := []*ModuleResourceInfo{
		{
			ModuleID:            "module-1",
			CurrentLoad:         2,
			MaxConcurrency:      10,
			CPUUsage:            30.0,
			MemoryUsage:         40.0,
			ResponseTime:        1 * time.Second,
			SuccessRate:         98.0,
			ErrorRate:           1.0,
			HealthStatus:        HealthStatusHealthy,
			ResourceUtilization: 0.35,
		},
		{
			ModuleID:            "module-2",
			CurrentLoad:         8,
			MaxConcurrency:      10,
			CPUUsage:            70.0,
			MemoryUsage:         80.0,
			ResponseTime:        3 * time.Second,
			SuccessRate:         92.0,
			ErrorRate:           5.0,
			HealthStatus:        HealthStatusHealthy,
			ResourceUtilization: 0.75,
		},
	}

	// Test round-robin strategy
	lb.strategy = LoadBalancingStrategyRoundRobin
	selected := lb.SelectModule(modules, nil)
	assert.NotNil(t, selected)
	assert.Equal(t, "module-1", selected.ModuleID)

	// Test least loaded strategy
	lb.strategy = LoadBalancingStrategyLeastLoaded
	selected = lb.SelectModule(modules, nil)
	assert.NotNil(t, selected)
	assert.Equal(t, "module-1", selected.ModuleID) // Less loaded

	// Test best performance strategy
	lb.strategy = LoadBalancingStrategyBestPerformance
	selected = lb.SelectModule(modules, nil)
	assert.NotNil(t, selected)
	assert.Equal(t, "module-1", selected.ModuleID) // Better performance

	// Test adaptive strategy
	lb.strategy = LoadBalancingStrategyAdaptive
	selected = lb.SelectModule(modules, nil)
	assert.NotNil(t, selected)
	assert.Equal(t, "module-1", selected.ModuleID) // Better overall score
}

func TestLoadBalancer_CalculatePerformanceScore(t *testing.T) {
	lb := NewLoadBalancer(LoadBalancingStrategyAdaptive)

	module := &ModuleResourceInfo{
		ResponseTime: 2 * time.Second,
		SuccessRate:  95.0,
	}

	score := lb.calculatePerformanceScore(module)
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)

	// Test with better performance
	betterModule := &ModuleResourceInfo{
		ResponseTime: 1 * time.Second,
		SuccessRate:  98.0,
	}

	betterScore := lb.calculatePerformanceScore(betterModule)
	assert.Greater(t, betterScore, score)
}

func TestLoadBalancer_CalculateAdaptiveScore(t *testing.T) {
	lb := NewLoadBalancer(LoadBalancingStrategyAdaptive)

	module := &ModuleResourceInfo{
		CurrentLoad:         5,
		MaxConcurrency:      10,
		ResponseTime:        2 * time.Second,
		SuccessRate:         95.0,
		ResourceUtilization: 0.5,
		HealthStatus:        HealthStatusHealthy,
	}

	score := lb.calculateAdaptiveScore(module, nil)
	assert.Greater(t, score, 0.0)
	assert.LessOrEqual(t, score, 1.0)

	// Test with unhealthy module
	unhealthyModule := &ModuleResourceInfo{
		CurrentLoad:         5,
		MaxConcurrency:      10,
		ResponseTime:        2 * time.Second,
		SuccessRate:         95.0,
		ResourceUtilization: 0.5,
		HealthStatus:        HealthStatusUnhealthy,
	}

	unhealthyScore := lb.calculateAdaptiveScore(unhealthyModule, nil)
	assert.Less(t, unhealthyScore, score)
}

func TestResourceManager_CalculateResourceUtilization(t *testing.T) {
	rm := createTestResourceManager(t)

	info := &ModuleResourceInfo{
		CurrentLoad:    5,
		MaxConcurrency: 10,
		CPUUsage:       50.0,
		MemoryUsage:    60.0,
	}

	utilization := rm.calculateResourceUtilization(info)
	assert.Greater(t, utilization, 0.0)
	assert.LessOrEqual(t, utilization, 1.0)

	// Test with higher resource usage
	highUsageInfo := &ModuleResourceInfo{
		CurrentLoad:    9,
		MaxConcurrency: 10,
		CPUUsage:       90.0,
		MemoryUsage:    95.0,
	}

	highUtilization := rm.calculateResourceUtilization(highUsageInfo)
	assert.Greater(t, highUtilization, utilization)
}

func TestCapacityPlanner_AddDataPoint(t *testing.T) {
	cp := NewCapacityPlanner()

	dataPoint := &CapacityDataPoint{
		Timestamp:     time.Now(),
		ModuleID:      "test-module",
		Load:          5,
		ResourceUsage: 0.5,
		ResponseTime:  2 * time.Second,
		SuccessRate:   95.0,
	}

	cp.AddDataPoint(dataPoint)
	assert.Equal(t, 1, len(cp.historicalData))
	assert.Equal(t, dataPoint, cp.historicalData[0])
}

func TestCapacityPlanner_GenerateForecast(t *testing.T) {
	cp := NewCapacityPlanner()

	// Add some historical data
	now := time.Now()
	for i := 0; i < 5; i++ {
		dataPoint := &CapacityDataPoint{
			Timestamp:     now.Add(time.Duration(i) * time.Hour),
			ModuleID:      "test-module",
			Load:          5 + i,
			ResourceUsage: 0.5 + float64(i)*0.1,
			ResponseTime:  2 * time.Second,
			SuccessRate:   95.0,
		}
		cp.AddDataPoint(dataPoint)
	}

	forecast := cp.GenerateForecast("test-module")
	assert.NotNil(t, forecast)
	assert.Equal(t, "test-module", forecast.ModuleID)
	assert.Greater(t, forecast.PredictedLoad, 0)
	assert.Greater(t, forecast.PredictedUsage, 0.0)
	assert.LessOrEqual(t, forecast.Confidence, 1.0)
	assert.NotEmpty(t, forecast.RecommendedScaling.Action)
}

func TestResourceManager_GetResourceSummary(t *testing.T) {
	rm := createTestResourceManager(t)

	// Add some modules
	module1 := &ModuleResourceInfo{
		ModuleID:            "module-1",
		CurrentLoad:         5,
		MaxConcurrency:      10,
		CPUUsage:            50.0,
		MemoryUsage:         60.0,
		HealthStatus:        HealthStatusHealthy,
		ResourceUtilization: 0.55,
	}

	module2 := &ModuleResourceInfo{
		ModuleID:            "module-2",
		CurrentLoad:         8,
		MaxConcurrency:      10,
		CPUUsage:            70.0,
		MemoryUsage:         80.0,
		HealthStatus:        HealthStatusDegraded,
		ResourceUtilization: 0.75,
	}

	rm.UpdateModuleResource("module-1", module1)
	rm.UpdateModuleResource("module-2", module2)

	summary := rm.GetResourceSummary()
	assert.Equal(t, 2, summary["total_modules"])
	assert.Equal(t, 1, summary["healthy_modules"]) // Only module-1 is healthy
	assert.Equal(t, 13, summary["total_load"])     // 5 + 8
	assert.Greater(t, summary["avg_utilization"], 0.0)
	assert.True(t, summary["load_balancing"].(bool))
	assert.True(t, summary["resource_monitoring"].(bool))
}

func TestResourceManager_HealthStatus(t *testing.T) {
	rm := createTestResourceManager(t)

	// Test healthy module
	healthyModule := &ModuleResourceInfo{
		ModuleID:            "healthy-module",
		CurrentLoad:         5,
		MaxConcurrency:      10,
		CPUUsage:            50.0,
		MemoryUsage:         60.0,
		ResponseTime:        2 * time.Second,
		ErrorRate:           2.0, // Less than 10%
		ResourceUtilization: 0.55,
	}

	rm.UpdateModuleResource("healthy-module", healthyModule)
	healthStatus := rm.checkModuleHealth(context.Background(), "healthy-module", healthyModule)
	assert.Equal(t, HealthStatusHealthy, healthStatus)

	// Test unhealthy module (high error rate)
	unhealthyModule := &ModuleResourceInfo{
		ModuleID:            "unhealthy-module",
		CurrentLoad:         5,
		MaxConcurrency:      10,
		CPUUsage:            50.0,
		MemoryUsage:         60.0,
		ResponseTime:        2 * time.Second,
		ErrorRate:           15.0, // More than 10%
		ResourceUtilization: 0.55,
	}

	healthStatus = rm.checkModuleHealth(context.Background(), "unhealthy-module", unhealthyModule)
	assert.Equal(t, HealthStatusUnhealthy, healthStatus)

	// Test degraded module (high resource utilization)
	degradedModule := &ModuleResourceInfo{
		ModuleID:            "degraded-module",
		CurrentLoad:         5,
		MaxConcurrency:      10,
		CPUUsage:            50.0,
		MemoryUsage:         60.0,
		ResponseTime:        2 * time.Second,
		ErrorRate:           2.0,
		ResourceUtilization: 0.9, // Above max utilization (0.8)
	}

	healthStatus = rm.checkModuleHealth(context.Background(), "degraded-module", degradedModule)
	assert.Equal(t, HealthStatusDegraded, healthStatus)
}

func TestScalingRecommendation(t *testing.T) {
	cp := NewCapacityPlanner()

	// Test scale up recommendation
	scaleUpRec := cp.generateScalingRecommendation(100, 0.85) // High usage
	assert.Equal(t, ScalingActionScaleUp, scaleUpRec.Action)
	assert.Equal(t, "predicted_high_utilization", scaleUpRec.Reason)
	assert.Equal(t, 1, scaleUpRec.Priority)

	// Test scale down recommendation
	scaleDownRec := cp.generateScalingRecommendation(10, 0.15) // Low usage
	assert.Equal(t, ScalingActionScaleDown, scaleDownRec.Action)
	assert.Equal(t, "predicted_low_utilization", scaleDownRec.Reason)
	assert.Equal(t, 2, scaleDownRec.Priority)

	// Test maintain recommendation
	maintainRec := cp.generateScalingRecommendation(50, 0.5) // Optimal usage
	assert.Equal(t, ScalingActionMaintain, maintainRec.Action)
	assert.Equal(t, "predicted_optimal_utilization", maintainRec.Reason)
	assert.Equal(t, 3, maintainRec.Priority)
}

// Helper function to create a test resource manager
func createTestResourceManager(t *testing.T) *ResourceManager {
	logger := observability.NewLogger(nil)
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	metrics, _ := observability.NewMetrics(nil)

	config := ResourceManagerConfig{
		EnableLoadBalancing:        true,
		EnableResourceMonitoring:   true,
		EnableHealthMonitoring:     true,
		EnableCapacityPlanning:     true,
		LoadBalancingStrategy:      LoadBalancingStrategyAdaptive,
		ResourceUpdateInterval:     30 * time.Second,
		HealthCheckInterval:        60 * time.Second,
		CapacityPlanningInterval:   300 * time.Second,
		MaxResourceUtilization:     0.8,
		MinResourceUtilization:     0.2,
		ScalingThreshold:           0.7,
		HealthCheckTimeout:         10 * time.Second,
		EnableResourceOptimization: true,
	}

	return NewResourceManager(logger, tracer, metrics, config)
}
