package scaling

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AutoScaler manages intelligent auto-scaling based on traffic predictions
type AutoScaler struct {
	config          *AutoScalingConfig
	trafficAnalyzer *TrafficAnalyzer
	metricsClient   MetricsClient
	k8sClient       KubernetesClient
	logger          *zap.Logger
	mu              sync.RWMutex
	lastScaleTime   time.Time
	scaleHistory    []ScaleEvent
}

// AutoScalingConfig defines auto-scaling configuration
type AutoScalingConfig struct {
	MinReplicas             int           `yaml:"min_replicas"`
	MaxReplicas             int           `yaml:"max_replicas"`
	TargetCPUUtilization    int           `yaml:"target_cpu_utilization"`
	TargetMemoryUtilization int           `yaml:"target_memory_utilization"`
	ScaleUpThreshold        float64       `yaml:"scale_up_threshold"`
	ScaleDownThreshold      float64       `yaml:"scale_down_threshold"`
	ScaleUpCooldown         time.Duration `yaml:"scale_up_cooldown"`
	ScaleDownCooldown       time.Duration `yaml:"scale_down_cooldown"`
	PredictionHorizon       time.Duration `yaml:"prediction_horizon"`
	TrafficBufferPercent    float64       `yaml:"traffic_buffer_percent"`
	EnablePredictiveScaling bool          `yaml:"enable_predictive_scaling"`
}

// ScaleEvent represents a scaling event
type ScaleEvent struct {
	Timestamp     time.Time `json:"timestamp"`
	Action        string    `json:"action"`
	FromReplicas  int       `json:"from_replicas"`
	ToReplicas    int       `json:"to_replicas"`
	Reason        string    `json:"reason"`
	PredictedLoad float64   `json:"predicted_load"`
	ActualLoad    float64   `json:"actual_load"`
}

// MetricsClient interface for metrics collection
type MetricsClient interface {
	GetCurrentCPUUtilization(ctx context.Context) (float64, error)
	GetCurrentMemoryUtilization(ctx context.Context) (float64, error)
	GetCurrentRequestRate(ctx context.Context) (float64, error)
	GetCurrentResponseTime(ctx context.Context) (float64, error)
}

// KubernetesClient interface for Kubernetes operations
type KubernetesClient interface {
	GetCurrentReplicas(ctx context.Context, deployment string) (int, error)
	ScaleDeployment(ctx context.Context, deployment string, replicas int) error
	GetDeploymentStatus(ctx context.Context, deployment string) (*DeploymentStatus, error)
}

// DeploymentStatus represents deployment status
type DeploymentStatus struct {
	Name      string    `json:"name"`
	Replicas  int       `json:"replicas"`
	Ready     int       `json:"ready"`
	Available int       `json:"available"`
	Updated   time.Time `json:"updated"`
}

// NewAutoScaler creates a new auto-scaler instance
func NewAutoScaler(
	config *AutoScalingConfig,
	trafficAnalyzer *TrafficAnalyzer,
	metricsClient MetricsClient,
	k8sClient KubernetesClient,
	logger *zap.Logger,
) *AutoScaler {
	return &AutoScaler{
		config:          config,
		trafficAnalyzer: trafficAnalyzer,
		metricsClient:   metricsClient,
		k8sClient:       k8sClient,
		logger:          logger,
		scaleHistory:    make([]ScaleEvent, 0),
	}
}

// Start begins the auto-scaling loop
func (as *AutoScaler) Start(ctx context.Context) error {
	as.logger.Info("Starting auto-scaler", zap.String("deployment", "risk-assessment-service"))

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			as.logger.Info("Auto-scaler stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := as.evaluateScaling(ctx); err != nil {
				as.logger.Error("Failed to evaluate scaling", zap.Error(err))
			}
		}
	}
}

// evaluateScaling evaluates whether scaling is needed
func (as *AutoScaler) evaluateScaling(ctx context.Context) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Check cooldown period
	if time.Since(as.lastScaleTime) < as.config.ScaleUpCooldown {
		return nil
	}

	// Get current metrics
	currentCPU, err := as.metricsClient.GetCurrentCPUUtilization(ctx)
	if err != nil {
		return fmt.Errorf("failed to get CPU utilization: %w", err)
	}

	currentMemory, err := as.metricsClient.GetCurrentMemoryUtilization(ctx)
	if err != nil {
		return fmt.Errorf("failed to get memory utilization: %w", err)
	}

	currentRequestRate, err := as.metricsClient.GetCurrentRequestRate(ctx)
	if err != nil {
		return fmt.Errorf("failed to get request rate: %w", err)
	}

	// Get current replica count
	currentReplicas, err := as.k8sClient.GetCurrentReplicas(ctx, "risk-assessment-service")
	if err != nil {
		return fmt.Errorf("failed to get current replicas: %w", err)
	}

	// Calculate scaling decision
	scalingDecision, err := as.calculateScalingDecision(ctx, currentCPU, currentMemory, currentRequestRate, currentReplicas)
	if err != nil {
		return fmt.Errorf("failed to calculate scaling decision: %w", err)
	}

	// Execute scaling if needed
	if scalingDecision != nil {
		if err := as.executeScaling(ctx, scalingDecision); err != nil {
			return fmt.Errorf("failed to execute scaling: %w", err)
		}
	}

	return nil
}

// ScalingDecision represents a scaling decision
type ScalingDecision struct {
	Action         string  `json:"action"`
	TargetReplicas int     `json:"target_replicas"`
	Reason         string  `json:"reason"`
	Confidence     float64 `json:"confidence"`
	PredictedLoad  float64 `json:"predicted_load"`
}

// calculateScalingDecision calculates whether scaling is needed
func (as *AutoScaler) calculateScalingDecision(ctx context.Context, currentCPU, currentMemory, currentRequestRate float64, currentReplicas int) (*ScalingDecision, error) {
	// Check if predictive scaling is enabled
	if as.config.EnablePredictiveScaling {
		return as.calculatePredictiveScaling(ctx, currentCPU, currentMemory, currentRequestRate, currentReplicas)
	}

	// Use reactive scaling based on current metrics
	return as.calculateReactiveScaling(currentCPU, currentMemory, currentRequestRate, currentReplicas)
}

// calculatePredictiveScaling calculates scaling based on traffic predictions
func (as *AutoScaler) calculatePredictiveScaling(ctx context.Context, currentCPU, currentMemory, currentRequestRate float64, currentReplicas int) (*ScalingDecision, error) {
	// Get traffic predictions
	predictions, err := as.trafficAnalyzer.PredictTraffic(ctx, as.config.PredictionHorizon)
	if err != nil {
		as.logger.Warn("Failed to get traffic predictions, falling back to reactive scaling", zap.Error(err))
		return as.calculateReactiveScaling(currentCPU, currentMemory, currentRequestRate, currentReplicas)
	}

	// Find the highest predicted load in the horizon
	var maxPredictedLoad float64
	var maxPrediction *TrafficPrediction

	for _, prediction := range predictions {
		if prediction.PredictedRequests > maxPredictedLoad {
			maxPredictedLoad = prediction.PredictedRequests
			maxPrediction = prediction
		}
	}

	// Add buffer to predicted load
	bufferedLoad := maxPredictedLoad * (1 + as.config.TrafficBufferPercent/100)

	// Calculate required replicas based on predicted load
	requiredReplicas := as.calculateRequiredReplicas(bufferedLoad, currentReplicas)

	// Determine scaling action
	var action string
	var reason string

	if requiredReplicas > currentReplicas {
		action = "scale_up"
		reason = fmt.Sprintf("Predicted load %.2f exceeds capacity, scaling up to %d replicas", bufferedLoad, requiredReplicas)
	} else if requiredReplicas < currentReplicas && currentCPU < float64(as.config.TargetCPUUtilization)*0.7 {
		action = "scale_down"
		reason = fmt.Sprintf("Predicted load %.2f allows scaling down to %d replicas", bufferedLoad, requiredReplicas)
	} else {
		return nil, nil // No scaling needed
	}

	// Ensure replicas are within bounds
	if requiredReplicas < as.config.MinReplicas {
		requiredReplicas = as.config.MinReplicas
	}
	if requiredReplicas > as.config.MaxReplicas {
		requiredReplicas = as.config.MaxReplicas
	}

	// Don't scale if target is same as current
	if requiredReplicas == currentReplicas {
		return nil, nil
	}

	return &ScalingDecision{
		Action:         action,
		TargetReplicas: requiredReplicas,
		Reason:         reason,
		Confidence:     maxPrediction.Confidence,
		PredictedLoad:  bufferedLoad,
	}, nil
}

// calculateReactiveScaling calculates scaling based on current metrics
func (as *AutoScaler) calculateReactiveScaling(currentCPU, currentMemory, currentRequestRate float64, currentReplicas int) (*ScalingDecision, error) {
	var action string
	var reason string
	var targetReplicas int

	// Check CPU utilization
	if currentCPU > float64(as.config.TargetCPUUtilization) {
		// Scale up based on CPU
		targetReplicas = int(float64(currentReplicas) * (currentCPU / float64(as.config.TargetCPUUtilization)))
		action = "scale_up"
		reason = fmt.Sprintf("CPU utilization %.2f%% exceeds target %d%%, scaling up", currentCPU, as.config.TargetCPUUtilization)
	} else if currentCPU < float64(as.config.TargetCPUUtilization)*0.5 && currentReplicas > as.config.MinReplicas {
		// Scale down based on CPU
		targetReplicas = currentReplicas - 1
		action = "scale_down"
		reason = fmt.Sprintf("CPU utilization %.2f%% is low, scaling down", currentCPU)
	} else {
		return nil, nil // No scaling needed
	}

	// Ensure replicas are within bounds
	if targetReplicas < as.config.MinReplicas {
		targetReplicas = as.config.MinReplicas
	}
	if targetReplicas > as.config.MaxReplicas {
		targetReplicas = as.config.MaxReplicas
	}

	// Don't scale if target is same as current
	if targetReplicas == currentReplicas {
		return nil, nil
	}

	return &ScalingDecision{
		Action:         action,
		TargetReplicas: targetReplicas,
		Reason:         reason,
		Confidence:     1.0, // Reactive scaling has high confidence
		PredictedLoad:  currentRequestRate,
	}, nil
}

// calculateRequiredReplicas calculates required replicas based on load
func (as *AutoScaler) calculateRequiredReplicas(predictedLoad float64, currentReplicas int) int {
	// Assume each replica can handle a certain load
	loadPerReplica := 100.0 // requests per second per replica

	requiredReplicas := int(predictedLoad / loadPerReplica)
	if requiredReplicas < 1 {
		requiredReplicas = 1
	}

	return requiredReplicas
}

// executeScaling executes the scaling decision
func (as *AutoScaler) executeScaling(ctx context.Context, decision *ScalingDecision) error {
	as.logger.Info("Executing scaling decision",
		zap.String("action", decision.Action),
		zap.Int("target_replicas", decision.TargetReplicas),
		zap.String("reason", decision.Reason),
		zap.Float64("confidence", decision.Confidence),
	)

	// Get current replicas for history
	currentReplicas, err := as.k8sClient.GetCurrentReplicas(ctx, "risk-assessment-service")
	if err != nil {
		return fmt.Errorf("failed to get current replicas: %w", err)
	}

	// Scale the deployment
	if err := as.k8sClient.ScaleDeployment(ctx, "risk-assessment-service", decision.TargetReplicas); err != nil {
		return fmt.Errorf("failed to scale deployment: %w", err)
	}

	// Record scaling event
	event := ScaleEvent{
		Timestamp:     time.Now(),
		Action:        decision.Action,
		FromReplicas:  currentReplicas,
		ToReplicas:    decision.TargetReplicas,
		Reason:        decision.Reason,
		PredictedLoad: decision.PredictedLoad,
		ActualLoad:    0, // Will be updated later
	}

	as.scaleHistory = append(as.scaleHistory, event)
	as.lastScaleTime = time.Now()

	// Keep only last 100 events
	if len(as.scaleHistory) > 100 {
		as.scaleHistory = as.scaleHistory[len(as.scaleHistory)-100:]
	}

	as.logger.Info("Scaling completed successfully",
		zap.Int("from_replicas", currentReplicas),
		zap.Int("to_replicas", decision.TargetReplicas),
	)

	return nil
}

// GetScalingHistory returns the scaling history
func (as *AutoScaler) GetScalingHistory() []ScaleEvent {
	as.mu.RLock()
	defer as.mu.RUnlock()

	// Return a copy to prevent external modification
	history := make([]ScaleEvent, len(as.scaleHistory))
	copy(history, as.scaleHistory)

	return history
}

// GetScalingStats returns scaling statistics
func (as *AutoScaler) GetScalingStats() map[string]interface{} {
	as.mu.RLock()
	defer as.mu.RUnlock()

	stats := map[string]interface{}{
		"total_scaling_events": len(as.scaleHistory),
		"last_scale_time":      as.lastScaleTime,
		"scale_up_events":      0,
		"scale_down_events":    0,
	}

	for _, event := range as.scaleHistory {
		if event.Action == "scale_up" {
			stats["scale_up_events"] = stats["scale_up_events"].(int) + 1
		} else if event.Action == "scale_down" {
			stats["scale_down_events"] = stats["scale_down_events"].(int) + 1
		}
	}

	return stats
}

// UpdateActualLoad updates the actual load for the most recent scaling event
func (as *AutoScaler) UpdateActualLoad(actualLoad float64) {
	as.mu.Lock()
	defer as.mu.Unlock()

	if len(as.scaleHistory) > 0 {
		as.scaleHistory[len(as.scaleHistory)-1].ActualLoad = actualLoad
	}
}
