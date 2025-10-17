package monitoring

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ResourceMonitor monitors system resources and provides cost optimization recommendations
type ResourceMonitor struct {
	config          *ResourceMonitorConfig
	metricsClient   MetricsClient
	costAnalyzer    CostAnalyzer
	logger          *zap.Logger
	mu              sync.RWMutex
	metrics         *ResourceMetrics
	recommendations []CostOptimizationRecommendation
}

// ResourceMonitorConfig defines resource monitoring configuration
type ResourceMonitorConfig struct {
	CollectionInterval         time.Duration `yaml:"collection_interval"`
	RetentionPeriod            time.Duration `yaml:"retention_period"`
	CostThresholdPercent       float64       `yaml:"cost_threshold_percent"`
	ResourceThresholdPercent   float64       `yaml:"resource_threshold_percent"`
	EnableCostOptimization     bool          `yaml:"enable_cost_optimization"`
	EnableResourceOptimization bool          `yaml:"enable_resource_optimization"`
}

// ResourceMetrics represents current resource metrics
type ResourceMetrics struct {
	Timestamp         time.Time      `json:"timestamp"`
	CPUUtilization    float64        `json:"cpu_utilization"`
	MemoryUtilization float64        `json:"memory_utilization"`
	DiskUtilization   float64        `json:"disk_utilization"`
	NetworkIO         NetworkMetrics `json:"network_io"`
	PodCount          int            `json:"pod_count"`
	RequestRate       float64        `json:"request_rate"`
	ResponseTime      float64        `json:"response_time"`
	ErrorRate         float64        `json:"error_rate"`
	CostPerHour       float64        `json:"cost_per_hour"`
	ResourceCosts     ResourceCosts  `json:"resource_costs"`
}

// NetworkMetrics represents network I/O metrics
type NetworkMetrics struct {
	BytesIn    float64 `json:"bytes_in"`
	BytesOut   float64 `json:"bytes_out"`
	PacketsIn  float64 `json:"packets_in"`
	PacketsOut float64 `json:"packets_out"`
}

// ResourceCosts represents cost breakdown by resource type
type ResourceCosts struct {
	ComputeCost  float64 `json:"compute_cost"`
	StorageCost  float64 `json:"storage_cost"`
	NetworkCost  float64 `json:"network_cost"`
	DatabaseCost float64 `json:"database_cost"`
	CacheCost    float64 `json:"cache_cost"`
	CDNCost      float64 `json:"cdn_cost"`
	TotalCost    float64 `json:"total_cost"`
}

// CostOptimizationRecommendation represents a cost optimization recommendation
type CostOptimizationRecommendation struct {
	ID               string    `json:"id"`
	Type             string    `json:"type"`
	Priority         string    `json:"priority"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	PotentialSavings float64   `json:"potential_savings"`
	Impact           string    `json:"impact"`
	Effort           string    `json:"effort"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// MetricsClient interface for collecting metrics
type MetricsClient interface {
	GetCPUUtilization(ctx context.Context) (float64, error)
	GetMemoryUtilization(ctx context.Context) (float64, error)
	GetDiskUtilization(ctx context.Context) (float64, error)
	GetNetworkMetrics(ctx context.Context) (*NetworkMetrics, error)
	GetPodCount(ctx context.Context) (int, error)
	GetRequestRate(ctx context.Context) (float64, error)
	GetResponseTime(ctx context.Context) (float64, error)
	GetErrorRate(ctx context.Context) (float64, error)
}

// CostAnalyzer interface for cost analysis
type CostAnalyzer interface {
	CalculateResourceCosts(ctx context.Context, metrics *ResourceMetrics) (*ResourceCosts, error)
	GenerateRecommendations(ctx context.Context, metrics *ResourceMetrics, costs *ResourceCosts) ([]CostOptimizationRecommendation, error)
	GetCostTrends(ctx context.Context, period time.Duration) ([]CostTrend, error)
}

// CostTrend represents cost trend data
type CostTrend struct {
	Timestamp time.Time `json:"timestamp"`
	Cost      float64   `json:"cost"`
	Type      string    `json:"type"`
}

// NewResourceMonitor creates a new resource monitor instance
func NewResourceMonitor(
	config *ResourceMonitorConfig,
	metricsClient MetricsClient,
	costAnalyzer CostAnalyzer,
	logger *zap.Logger,
) *ResourceMonitor {
	return &ResourceMonitor{
		config:          config,
		metricsClient:   metricsClient,
		costAnalyzer:    costAnalyzer,
		logger:          logger,
		metrics:         &ResourceMetrics{},
		recommendations: make([]CostOptimizationRecommendation, 0),
	}
}

// Start begins the resource monitoring loop
func (rm *ResourceMonitor) Start(ctx context.Context) error {
	rm.logger.Info("Starting resource monitor")

	ticker := time.NewTicker(rm.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			rm.logger.Info("Resource monitor stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := rm.collectMetrics(ctx); err != nil {
				rm.logger.Error("Failed to collect metrics", zap.Error(err))
			}

			if rm.config.EnableCostOptimization {
				if err := rm.analyzeCosts(ctx); err != nil {
					rm.logger.Error("Failed to analyze costs", zap.Error(err))
				}
			}
		}
	}
}

// collectMetrics collects current resource metrics
func (rm *ResourceMonitor) collectMetrics(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Collect basic metrics
	cpuUtil, err := rm.metricsClient.GetCPUUtilization(ctx)
	if err != nil {
		return fmt.Errorf("failed to get CPU utilization: %w", err)
	}

	memoryUtil, err := rm.metricsClient.GetMemoryUtilization(ctx)
	if err != nil {
		return fmt.Errorf("failed to get memory utilization: %w", err)
	}

	diskUtil, err := rm.metricsClient.GetDiskUtilization(ctx)
	if err != nil {
		return fmt.Errorf("failed to get disk utilization: %w", err)
	}

	networkMetrics, err := rm.metricsClient.GetNetworkMetrics(ctx)
	if err != nil {
		return fmt.Errorf("failed to get network metrics: %w", err)
	}

	podCount, err := rm.metricsClient.GetPodCount(ctx)
	if err != nil {
		return fmt.Errorf("failed to get pod count: %w", err)
	}

	requestRate, err := rm.metricsClient.GetRequestRate(ctx)
	if err != nil {
		return fmt.Errorf("failed to get request rate: %w", err)
	}

	responseTime, err := rm.metricsClient.GetResponseTime(ctx)
	if err != nil {
		return fmt.Errorf("failed to get response time: %w", err)
	}

	errorRate, err := rm.metricsClient.GetErrorRate(ctx)
	if err != nil {
		return fmt.Errorf("failed to get error rate: %w", err)
	}

	// Update metrics
	rm.metrics = &ResourceMetrics{
		Timestamp:         time.Now(),
		CPUUtilization:    cpuUtil,
		MemoryUtilization: memoryUtil,
		DiskUtilization:   diskUtil,
		NetworkIO:         *networkMetrics,
		PodCount:          podCount,
		RequestRate:       requestRate,
		ResponseTime:      responseTime,
		ErrorRate:         errorRate,
	}

	rm.logger.Debug("Collected resource metrics",
		zap.Float64("cpu_utilization", cpuUtil),
		zap.Float64("memory_utilization", memoryUtil),
		zap.Int("pod_count", podCount),
		zap.Float64("request_rate", requestRate),
	)

	return nil
}

// analyzeCosts analyzes costs and generates recommendations
func (rm *ResourceMonitor) analyzeCosts(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	// Calculate resource costs
	costs, err := rm.costAnalyzer.CalculateResourceCosts(ctx, rm.metrics)
	if err != nil {
		return fmt.Errorf("failed to calculate resource costs: %w", err)
	}

	rm.metrics.ResourceCosts = *costs
	rm.metrics.CostPerHour = costs.TotalCost

	// Generate recommendations
	recommendations, err := rm.costAnalyzer.GenerateRecommendations(ctx, rm.metrics, costs)
	if err != nil {
		return fmt.Errorf("failed to generate recommendations: %w", err)
	}

	// Update recommendations
	rm.recommendations = recommendations

	rm.logger.Info("Cost analysis completed",
		zap.Float64("total_cost_per_hour", costs.TotalCost),
		zap.Int("recommendations_count", len(recommendations)),
	)

	return nil
}

// GetCurrentMetrics returns current resource metrics
func (rm *ResourceMonitor) GetCurrentMetrics() *ResourceMetrics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a copy to prevent external modification
	metrics := *rm.metrics
	return &metrics
}

// GetRecommendations returns current cost optimization recommendations
func (rm *ResourceMonitor) GetRecommendations() []CostOptimizationRecommendation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a copy to prevent external modification
	recommendations := make([]CostOptimizationRecommendation, len(rm.recommendations))
	copy(recommendations, rm.recommendations)

	return recommendations
}

// GetCostTrends returns cost trends for the specified period
func (rm *ResourceMonitor) GetCostTrends(ctx context.Context, period time.Duration) ([]CostTrend, error) {
	return rm.costAnalyzer.GetCostTrends(ctx, period)
}

// GetResourceUtilization returns current resource utilization
func (rm *ResourceMonitor) GetResourceUtilization() map[string]float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return map[string]float64{
		"cpu_utilization":    rm.metrics.CPUUtilization,
		"memory_utilization": rm.metrics.MemoryUtilization,
		"disk_utilization":   rm.metrics.DiskUtilization,
	}
}

// GetCostBreakdown returns current cost breakdown
func (rm *ResourceMonitor) GetCostBreakdown() *ResourceCosts {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Return a copy to prevent external modification
	costs := rm.metrics.ResourceCosts
	return &costs
}

// GetHighPriorityRecommendations returns high priority recommendations
func (rm *ResourceMonitor) GetHighPriorityRecommendations() []CostOptimizationRecommendation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var highPriority []CostOptimizationRecommendation
	for _, rec := range rm.recommendations {
		if rec.Priority == "high" {
			highPriority = append(highPriority, rec)
		}
	}

	return highPriority
}

// GetRecommendationsByType returns recommendations filtered by type
func (rm *ResourceMonitor) GetRecommendationsByType(recType string) []CostOptimizationRecommendation {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var filtered []CostOptimizationRecommendation
	for _, rec := range rm.recommendations {
		if rec.Type == recType {
			filtered = append(filtered, rec)
		}
	}

	return filtered
}

// UpdateRecommendationStatus updates the status of a recommendation
func (rm *ResourceMonitor) UpdateRecommendationStatus(id, status string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	for i, rec := range rm.recommendations {
		if rec.ID == id {
			rm.recommendations[i].Status = status
			rm.recommendations[i].UpdatedAt = time.Now()
			return nil
		}
	}

	return fmt.Errorf("recommendation with ID %s not found", id)
}

// GetResourceAlerts returns resource utilization alerts
func (rm *ResourceMonitor) GetResourceAlerts() []ResourceAlert {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var alerts []ResourceAlert

	// Check CPU utilization
	if rm.metrics.CPUUtilization > rm.config.ResourceThresholdPercent {
		alerts = append(alerts, ResourceAlert{
			Type:      "cpu_utilization",
			Severity:  "warning",
			Message:   fmt.Sprintf("CPU utilization is %.2f%% (threshold: %.2f%%)", rm.metrics.CPUUtilization, rm.config.ResourceThresholdPercent),
			Value:     rm.metrics.CPUUtilization,
			Threshold: rm.config.ResourceThresholdPercent,
			Timestamp: time.Now(),
		})
	}

	// Check memory utilization
	if rm.metrics.MemoryUtilization > rm.config.ResourceThresholdPercent {
		alerts = append(alerts, ResourceAlert{
			Type:      "memory_utilization",
			Severity:  "warning",
			Message:   fmt.Sprintf("Memory utilization is %.2f%% (threshold: %.2f%%)", rm.metrics.MemoryUtilization, rm.config.ResourceThresholdPercent),
			Value:     rm.metrics.MemoryUtilization,
			Threshold: rm.config.ResourceThresholdPercent,
			Timestamp: time.Now(),
		})
	}

	// Check error rate
	if rm.metrics.ErrorRate > 0.05 { // 5% error rate threshold
		alerts = append(alerts, ResourceAlert{
			Type:      "error_rate",
			Severity:  "critical",
			Message:   fmt.Sprintf("Error rate is %.2f%% (threshold: 5%%)", rm.metrics.ErrorRate*100),
			Value:     rm.metrics.ErrorRate,
			Threshold: 0.05,
			Timestamp: time.Now(),
		})
	}

	return alerts
}

// ResourceAlert represents a resource utilization alert
type ResourceAlert struct {
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
	Timestamp time.Time `json:"timestamp"`
}
