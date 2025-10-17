package monitoring

import (
	"context"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
)

// CostAnalyzerImpl implements the CostAnalyzer interface
type CostAnalyzerImpl struct {
	config *CostAnalyzerConfig
	logger *zap.Logger
}

// CostAnalyzerConfig defines cost analyzer configuration
type CostAnalyzerConfig struct {
	// Pricing per hour (in USD)
	CPUCostPerHour       float64 `yaml:"cpu_cost_per_hour"`
	MemoryCostPerGBHour  float64 `yaml:"memory_cost_per_gb_hour"`
	StorageCostPerGBHour float64 `yaml:"storage_cost_per_gb_hour"`
	NetworkCostPerGB     float64 `yaml:"network_cost_per_gb"`
	DatabaseCostPerHour  float64 `yaml:"database_cost_per_hour"`
	CacheCostPerHour     float64 `yaml:"cache_cost_per_hour"`
	CDNCostPerGB         float64 `yaml:"cdn_cost_per_gb"`

	// Resource specifications
	CPUCoresPerPod  float64 `yaml:"cpu_cores_per_pod"`
	MemoryGBPerPod  float64 `yaml:"memory_gb_per_pod"`
	StorageGBPerPod float64 `yaml:"storage_gb_per_pod"`

	// Optimization thresholds
	LowUtilizationThreshold  float64 `yaml:"low_utilization_threshold"`
	HighUtilizationThreshold float64 `yaml:"high_utilization_threshold"`
	CostSavingsThreshold     float64 `yaml:"cost_savings_threshold"`
}

// NewCostAnalyzer creates a new cost analyzer instance
func NewCostAnalyzer(config *CostAnalyzerConfig, logger *zap.Logger) *CostAnalyzerImpl {
	return &CostAnalyzerImpl{
		config: config,
		logger: logger,
	}
}

// CalculateResourceCosts calculates costs for current resource usage
func (ca *CostAnalyzerImpl) CalculateResourceCosts(ctx context.Context, metrics *ResourceMetrics) (*ResourceCosts, error) {
	costs := &ResourceCosts{}

	// Calculate compute costs
	computeCost := float64(metrics.PodCount) * ca.config.CPUCostPerHour
	costs.ComputeCost = computeCost

	// Calculate storage costs
	storageCost := float64(metrics.PodCount) * ca.config.StorageGBPerPod * ca.config.StorageCostPerGBHour
	costs.StorageCost = storageCost

	// Calculate network costs
	networkGB := (metrics.NetworkIO.BytesIn + metrics.NetworkIO.BytesOut) / (1024 * 1024 * 1024)
	networkCost := networkGB * ca.config.NetworkCostPerGB
	costs.NetworkCost = networkCost

	// Calculate database costs (estimated based on request rate)
	databaseCost := ca.config.DatabaseCostPerHour * (1 + metrics.RequestRate/1000) // Scale with request rate
	costs.DatabaseCost = databaseCost

	// Calculate cache costs
	cacheCost := ca.config.CacheCostPerHour * (1 + metrics.RequestRate/1000) // Scale with request rate
	costs.CacheCost = cacheCost

	// Calculate CDN costs
	cdnCost := networkGB * ca.config.CDNCostPerGB
	costs.CDNCost = cdnCost

	// Calculate total cost
	costs.TotalCost = costs.ComputeCost + costs.StorageCost + costs.NetworkCost +
		costs.DatabaseCost + costs.CacheCost + costs.CDNCost

	ca.logger.Debug("Calculated resource costs",
		zap.Float64("compute_cost", costs.ComputeCost),
		zap.Float64("storage_cost", costs.StorageCost),
		zap.Float64("network_cost", costs.NetworkCost),
		zap.Float64("database_cost", costs.DatabaseCost),
		zap.Float64("cache_cost", costs.CacheCost),
		zap.Float64("cdn_cost", costs.CDNCost),
		zap.Float64("total_cost", costs.TotalCost),
	)

	return costs, nil
}

// GenerateRecommendations generates cost optimization recommendations
func (ca *CostAnalyzerImpl) GenerateRecommendations(ctx context.Context, metrics *ResourceMetrics, costs *ResourceCosts) ([]CostOptimizationRecommendation, error) {
	var recommendations []CostOptimizationRecommendation

	// Check for over-provisioned resources
	if metrics.CPUUtilization < ca.config.LowUtilizationThreshold {
		recommendations = append(recommendations, ca.generateCPUOptimizationRecommendation(metrics, costs))
	}

	if metrics.MemoryUtilization < ca.config.LowUtilizationThreshold {
		recommendations = append(recommendations, ca.generateMemoryOptimizationRecommendation(metrics, costs))
	}

	// Check for under-provisioned resources
	if metrics.CPUUtilization > ca.config.HighUtilizationThreshold {
		recommendations = append(recommendations, ca.generateCPUUpgradeRecommendation(metrics, costs))
	}

	if metrics.MemoryUtilization > ca.config.HighUtilizationThreshold {
		recommendations = append(recommendations, ca.generateMemoryUpgradeRecommendation(metrics, costs))
	}

	// Check for idle resources
	if metrics.RequestRate < 10 && metrics.PodCount > 2 {
		recommendations = append(recommendations, ca.generatePodReductionRecommendation(metrics, costs))
	}

	// Check for high error rates
	if metrics.ErrorRate > 0.01 {
		recommendations = append(recommendations, ca.generateErrorRateOptimizationRecommendation(metrics, costs))
	}

	// Check for slow response times
	if metrics.ResponseTime > 200 {
		recommendations = append(recommendations, ca.generateResponseTimeOptimizationRecommendation(metrics, costs))
	}

	// Check for high network costs
	if costs.NetworkCost > costs.TotalCost*0.3 {
		recommendations = append(recommendations, ca.generateNetworkOptimizationRecommendation(metrics, costs))
	}

	// Check for high database costs
	if costs.DatabaseCost > costs.TotalCost*0.4 {
		recommendations = append(recommendations, ca.generateDatabaseOptimizationRecommendation(metrics, costs))
	}

	// Sort recommendations by potential savings
	sort.Slice(recommendations, func(i, j int) bool {
		return recommendations[i].PotentialSavings > recommendations[j].PotentialSavings
	})

	ca.logger.Info("Generated cost optimization recommendations",
		zap.Int("count", len(recommendations)),
	)

	return recommendations, nil
}

// generateCPUOptimizationRecommendation generates CPU optimization recommendation
func (ca *CostAnalyzerImpl) generateCPUOptimizationRecommendation(metrics *ResourceMetrics, costs *ResourceCosts) CostOptimizationRecommendation {
	potentialSavings := costs.ComputeCost * 0.2 // 20% savings from CPU optimization

	return CostOptimizationRecommendation{
		ID:               fmt.Sprintf("cpu_opt_%d", time.Now().Unix()),
		Type:             "compute_optimization",
		Priority:         ca.getPriority(potentialSavings),
		Title:            "Optimize CPU Resources",
		Description:      fmt.Sprintf("CPU utilization is %.2f%%, consider reducing CPU allocation or pod count", metrics.CPUUtilization),
		PotentialSavings: potentialSavings,
		Impact:           "medium",
		Effort:           "low",
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// generateMemoryOptimizationRecommendation generates memory optimization recommendation
func (ca *CostAnalyzerImpl) generateMemoryOptimizationRecommendation(metrics *ResourceMetrics, costs *ResourceCosts) CostOptimizationRecommendation {
	potentialSavings := costs.ComputeCost * 0.15 // 15% savings from memory optimization

	return CostOptimizationRecommendation{
		ID:               fmt.Sprintf("memory_opt_%d", time.Now().Unix()),
		Type:             "compute_optimization",
		Priority:         ca.getPriority(potentialSavings),
		Title:            "Optimize Memory Resources",
		Description:      fmt.Sprintf("Memory utilization is %.2f%%, consider reducing memory allocation", metrics.MemoryUtilization),
		PotentialSavings: potentialSavings,
		Impact:           "medium",
		Effort:           "low",
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// generateCPUUpgradeRecommendation generates CPU upgrade recommendation
func (ca *CostAnalyzerImpl) generateCPUUpgradeRecommendation(metrics *ResourceMetrics, costs *ResourceCosts) CostOptimizationRecommendation {
	additionalCost := costs.ComputeCost * 0.3 // 30% increase for better performance

	return CostOptimizationRecommendation{
		ID:               fmt.Sprintf("cpu_upgrade_%d", time.Now().Unix()),
		Type:             "performance_improvement",
		Priority:         "high",
		Title:            "Upgrade CPU Resources",
		Description:      fmt.Sprintf("CPU utilization is %.2f%%, consider upgrading CPU resources for better performance", metrics.CPUUtilization),
		PotentialSavings: -additionalCost, // Negative value indicates cost increase
		Impact:           "high",
		Effort:           "medium",
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// generateMemoryUpgradeRecommendation generates memory upgrade recommendation
func (ca *CostAnalyzerImpl) generateMemoryUpgradeRecommendation(metrics *ResourceMetrics, costs *ResourceCosts) CostOptimizationRecommendation {
	additionalCost := costs.ComputeCost * 0.2 // 20% increase for better performance

	return CostOptimizationRecommendation{
		ID:               fmt.Sprintf("memory_upgrade_%d", time.Now().Unix()),
		Type:             "performance_improvement",
		Priority:         "high",
		Title:            "Upgrade Memory Resources",
		Description:      fmt.Sprintf("Memory utilization is %.2f%%, consider upgrading memory resources", metrics.MemoryUtilization),
		PotentialSavings: -additionalCost, // Negative value indicates cost increase
		Impact:           "high",
		Effort:           "medium",
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// generatePodReductionRecommendation generates pod reduction recommendation
func (ca *CostAnalyzerImpl) generatePodReductionRecommendation(metrics *ResourceMetrics, costs *ResourceCosts) CostOptimizationRecommendation {
	potentialSavings := costs.ComputeCost * 0.5 // 50% savings from reducing pods

	return CostOptimizationRecommendation{
		ID:               fmt.Sprintf("pod_reduction_%d", time.Now().Unix()),
		Type:             "resource_optimization",
		Priority:         ca.getPriority(potentialSavings),
		Title:            "Reduce Pod Count",
		Description:      fmt.Sprintf("Request rate is %.2f req/s with %d pods, consider reducing pod count", metrics.RequestRate, metrics.PodCount),
		PotentialSavings: potentialSavings,
		Impact:           "high",
		Effort:           "low",
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// generateErrorRateOptimizationRecommendation generates error rate optimization recommendation
func (ca *CostAnalyzerImpl) generateErrorRateOptimizationRecommendation(metrics *ResourceMetrics, costs *ResourceCosts) CostOptimizationRecommendation {
	// Estimate cost of errors (lost revenue, support costs, etc.)
	errorCost := costs.TotalCost * metrics.ErrorRate * 10 // 10x multiplier for error impact

	return CostOptimizationRecommendation{
		ID:               fmt.Sprintf("error_rate_opt_%d", time.Now().Unix()),
		Type:             "reliability_improvement",
		Priority:         "critical",
		Title:            "Reduce Error Rate",
		Description:      fmt.Sprintf("Error rate is %.2f%%, investigate and fix issues to improve reliability", metrics.ErrorRate*100),
		PotentialSavings: errorCost,
		Impact:           "critical",
		Effort:           "high",
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// generateResponseTimeOptimizationRecommendation generates response time optimization recommendation
func (ca *CostAnalyzerImpl) generateResponseTimeOptimizationRecommendation(metrics *ResourceMetrics, costs *ResourceCosts) CostOptimizationRecommendation {
	// Estimate cost of slow response times (user experience impact)
	performanceCost := costs.TotalCost * 0.1 // 10% of total cost for performance issues

	return CostOptimizationRecommendation{
		ID:               fmt.Sprintf("response_time_opt_%d", time.Now().Unix()),
		Type:             "performance_improvement",
		Priority:         "high",
		Title:            "Optimize Response Time",
		Description:      fmt.Sprintf("Response time is %.2fms, optimize for better user experience", metrics.ResponseTime),
		PotentialSavings: performanceCost,
		Impact:           "high",
		Effort:           "medium",
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// generateNetworkOptimizationRecommendation generates network optimization recommendation
func (ca *CostAnalyzerImpl) generateNetworkOptimizationRecommendation(metrics *ResourceMetrics, costs *ResourceCosts) CostOptimizationRecommendation {
	potentialSavings := costs.NetworkCost * 0.3 // 30% savings from network optimization

	return CostOptimizationRecommendation{
		ID:               fmt.Sprintf("network_opt_%d", time.Now().Unix()),
		Type:             "network_optimization",
		Priority:         ca.getPriority(potentialSavings),
		Title:            "Optimize Network Usage",
		Description:      fmt.Sprintf("Network costs are %.2f%% of total costs, consider optimizing data transfer", (costs.NetworkCost/costs.TotalCost)*100),
		PotentialSavings: potentialSavings,
		Impact:           "medium",
		Effort:           "medium",
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// generateDatabaseOptimizationRecommendation generates database optimization recommendation
func (ca *CostAnalyzerImpl) generateDatabaseOptimizationRecommendation(metrics *ResourceMetrics, costs *ResourceCosts) CostOptimizationRecommendation {
	potentialSavings := costs.DatabaseCost * 0.25 // 25% savings from database optimization

	return CostOptimizationRecommendation{
		ID:               fmt.Sprintf("database_opt_%d", time.Now().Unix()),
		Type:             "database_optimization",
		Priority:         ca.getPriority(potentialSavings),
		Title:            "Optimize Database Usage",
		Description:      fmt.Sprintf("Database costs are %.2f%% of total costs, consider query optimization and caching", (costs.DatabaseCost/costs.TotalCost)*100),
		PotentialSavings: potentialSavings,
		Impact:           "high",
		Effort:           "high",
		Status:           "pending",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

// getPriority determines recommendation priority based on potential savings
func (ca *CostAnalyzerImpl) getPriority(potentialSavings float64) string {
	if potentialSavings >= ca.config.CostSavingsThreshold*2 {
		return "critical"
	} else if potentialSavings >= ca.config.CostSavingsThreshold {
		return "high"
	} else if potentialSavings >= ca.config.CostSavingsThreshold*0.5 {
		return "medium"
	}
	return "low"
}

// GetCostTrends returns cost trends for the specified period
func (ca *CostAnalyzerImpl) GetCostTrends(ctx context.Context, period time.Duration) ([]CostTrend, error) {
	// This would typically query a time-series database
	// For now, we'll return mock data
	trends := []CostTrend{
		{
			Timestamp: time.Now().Add(-period),
			Cost:      100.0,
			Type:      "total",
		},
		{
			Timestamp: time.Now().Add(-period / 2),
			Cost:      120.0,
			Type:      "total",
		},
		{
			Timestamp: time.Now(),
			Cost:      110.0,
			Type:      "total",
		},
	}

	return trends, nil
}
