package microservices

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"kyb-platform/internal/observability"
)

// ServiceClientImpl implements the ServiceClient interface
type ServiceClientImpl struct {
	discovery      *ServiceDiscoveryImpl
	loadBalancer   ServiceLoadBalancer
	circuitBreaker ServiceCircuitBreaker
	metrics        ServiceMetrics
	timeout        ServiceTimeout
	retry          ServiceRetry
	rateLimiter    ServiceRateLimiter
	logger         *observability.Logger
}

// NewServiceClient creates a new service client
func NewServiceClient(
	discovery *ServiceDiscoveryImpl,
	loadBalancer ServiceLoadBalancer,
	circuitBreaker ServiceCircuitBreaker,
	metrics ServiceMetrics,
	timeout ServiceTimeout,
	retry ServiceRetry,
	rateLimiter ServiceRateLimiter,
	logger *observability.Logger,
) *ServiceClientImpl {
	return &ServiceClientImpl{
		discovery:      discovery,
		loadBalancer:   loadBalancer,
		circuitBreaker: circuitBreaker,
		metrics:        metrics,
		timeout:        timeout,
		retry:          retry,
		rateLimiter:    rateLimiter,
		logger:         logger,
	}
}

// Call makes a synchronous call to a service
func (c *ServiceClientImpl) Call(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
	start := time.Now()

	// Check rate limiting
	if !c.rateLimiter.Allow(serviceName) {
		c.metrics.RecordError(serviceName, method, "rate_limited")
		return nil, fmt.Errorf("rate limit exceeded for service %s", serviceName)
	}

	// Use circuit breaker
	result, err := c.circuitBreaker.Execute(ctx, serviceName, method, request)

	// Record metrics
	duration := time.Since(start)
	success := err == nil
	c.metrics.RecordRequest(serviceName, method, duration, success)
	c.metrics.RecordLatency(serviceName, method, duration)

	if !success {
		c.metrics.RecordError(serviceName, method, "call_failed")
	}

	return result, err
}

// CallAsync makes an asynchronous call to a service
func (c *ServiceClientImpl) CallAsync(ctx context.Context, serviceName, method string, request interface{}) (<-chan ServiceResponse, error) {
	responseChan := make(chan ServiceResponse, 1)

	go func() {
		start := time.Now()

		// Check rate limiting
		if !c.rateLimiter.Allow(serviceName) {
			c.metrics.RecordError(serviceName, method, "rate_limited")
			responseChan <- ServiceResponse{
				Data:    nil,
				Error:   fmt.Errorf("rate limit exceeded for service %s", serviceName),
				Latency: time.Since(start),
			}
			close(responseChan)
			return
		}

		// Use circuit breaker
		result, err := c.circuitBreaker.Execute(ctx, serviceName, method, request)

		// Record metrics
		duration := time.Since(start)
		success := err == nil
		c.metrics.RecordRequest(serviceName, method, duration, success)
		c.metrics.RecordLatency(serviceName, method, duration)

		if !success {
			c.metrics.RecordError(serviceName, method, "call_failed")
		}

		responseChan <- ServiceResponse{
			Data:    result,
			Error:   err,
			Latency: duration,
		}
		close(responseChan)
	}()

	return responseChan, nil
}

// Health checks the health of a service
func (c *ServiceClientImpl) Health(ctx context.Context, serviceName string) (ServiceHealth, error) {
	instances, err := c.discovery.Discover(serviceName)
	if err != nil {
		return ServiceHealth{}, err
	}

	if len(instances) == 0 {
		return ServiceHealth{
			Status:    "unhealthy",
			Message:   "No healthy instances available",
			Timestamp: time.Now(),
		}, nil
	}

	// Return health of first instance (simplified)
	return instances[0].Health, nil
}

// ServiceLoadBalancerImpl implements the ServiceLoadBalancer interface
type ServiceLoadBalancerImpl struct {
	discovery *ServiceDiscoveryImpl
	mu        sync.RWMutex
	logger    *observability.Logger
}

// NewServiceLoadBalancer creates a new service load balancer
func NewServiceLoadBalancer(discovery *ServiceDiscoveryImpl, logger *observability.Logger) *ServiceLoadBalancerImpl {
	return &ServiceLoadBalancerImpl{
		discovery: discovery,
		logger:    logger,
	}
}

// Select selects a service instance using round-robin load balancing
func (lb *ServiceLoadBalancerImpl) Select(serviceName string) (ServiceInstance, error) {
	instances, err := lb.discovery.Discover(serviceName)
	if err != nil {
		return ServiceInstance{}, err
	}

	if len(instances) == 0 {
		return ServiceInstance{}, fmt.Errorf("no healthy instances available for service %s", serviceName)
	}

	// Simple round-robin selection
	lb.mu.Lock()
	defer lb.mu.Unlock()

	// Use random selection for now (can be enhanced with round-robin)
	selectedIndex := rand.Intn(len(instances))
	selectedInstance := instances[selectedIndex]

	lb.logger.Info("Service instance selected", map[string]interface{}{
		"service_name": serviceName,
		"instance_id":  selectedInstance.ID,
		"host":         selectedInstance.Host,
		"port":         selectedInstance.Port,
	})

	return selectedInstance, nil
}

// UpdateHealth updates the health status of a service instance
func (lb *ServiceLoadBalancerImpl) UpdateHealth(instanceID string, health ServiceHealth) error {
	// This would typically update the health status in the discovery system
	// For now, we'll just log the update
	lb.logger.Info("Service instance health updated", map[string]interface{}{
		"instance_id":    instanceID,
		"health_status":  health.Status,
		"health_message": health.Message,
	})
	return nil
}

// GetInstances returns all instances of a service
func (lb *ServiceLoadBalancerImpl) GetInstances(serviceName string) ([]ServiceInstance, error) {
	return lb.discovery.Discover(serviceName)
}

// ServiceCircuitBreakerImpl implements the ServiceCircuitBreaker interface
type ServiceCircuitBreakerImpl struct {
	states map[string]*CircuitBreakerState
	mu     sync.RWMutex
	logger *observability.Logger
}

// NewServiceCircuitBreaker creates a new service circuit breaker
func NewServiceCircuitBreaker(logger *observability.Logger) *ServiceCircuitBreakerImpl {
	return &ServiceCircuitBreakerImpl{
		states: make(map[string]*CircuitBreakerState),
		logger: logger,
	}
}

// Execute executes a service call with circuit breaker protection
func (cb *ServiceCircuitBreakerImpl) Execute(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
	key := fmt.Sprintf("%s:%s", serviceName, method)

	cb.mu.Lock()
	state, exists := cb.states[key]
	if !exists {
		state = &CircuitBreakerState{
			ServiceName:     serviceName,
			State:           "closed",
			FailureCount:    0,
			SuccessCount:    0,
			Threshold:       5,
			Timeout:         30 * time.Second,
			LastStateChange: time.Now(),
		}
		cb.states[key] = state
	}
	cb.mu.Unlock()

	// Check circuit breaker state
	switch state.State {
	case "open":
		if time.Since(state.LastStateChange) < state.Timeout {
			return nil, fmt.Errorf("circuit breaker is open for %s", key)
		}
		// Try to transition to half-open
		cb.mu.Lock()
		state.State = "half-open"
		state.LastStateChange = time.Now()
		cb.mu.Unlock()

	case "half-open":
		// Allow one request to test if service is back
		break

	case "closed":
		// Normal operation
		break
	}

	// Execute the actual service call (simulated for now)
	result, err := cb.executeServiceCall(ctx, serviceName, method, request)

	// Update circuit breaker state
	cb.updateState(state, err)

	return result, err
}

// GetState returns the current state of a circuit breaker
func (cb *ServiceCircuitBreakerImpl) GetState(serviceName string) CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	// Return the first matching state (simplified)
	for _, state := range cb.states {
		if state.ServiceName == serviceName {
			return *state
		}
	}

	return CircuitBreakerState{
		ServiceName: serviceName,
		State:       "closed",
		Threshold:   5,
		Timeout:     30 * time.Second,
	}
}

// Reset resets a circuit breaker
func (cb *ServiceCircuitBreakerImpl) Reset(serviceName string) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	for _, state := range cb.states {
		if state.ServiceName == serviceName {
			state.State = "closed"
			state.FailureCount = 0
			state.SuccessCount = 0
			state.LastStateChange = time.Now()

			cb.logger.Info("Circuit breaker reset", map[string]interface{}{
				"service_name": serviceName,
			})
			return nil
		}
	}

	return fmt.Errorf("circuit breaker not found for service %s", serviceName)
}

// executeServiceCall simulates a service call (replace with actual implementation)
func (cb *ServiceCircuitBreakerImpl) executeServiceCall(ctx context.Context, serviceName, method string, request interface{}) (interface{}, error) {
	// Simulate service call with random success/failure
	time.Sleep(10 * time.Millisecond) // Simulate network delay

	if rand.Float64() < 0.1 { // 10% failure rate
		return nil, fmt.Errorf("service call failed for %s:%s", serviceName, method)
	}

	return map[string]interface{}{
		"service": serviceName,
		"method":  method,
		"result":  "success",
		"data":    request,
	}, nil
}

// updateState updates the circuit breaker state based on the result
func (cb *ServiceCircuitBreakerImpl) updateState(state *CircuitBreakerState, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		state.FailureCount++
		now := time.Now()
		state.LastFailure = &now

		if state.State == "half-open" {
			// Back to open
			state.State = "open"
			state.LastStateChange = now
			cb.logger.Warn("Circuit breaker opened", map[string]interface{}{
				"service_name":  state.ServiceName,
				"failure_count": state.FailureCount,
			})
		} else if state.State == "closed" && state.FailureCount >= state.Threshold {
			// Transition to open
			state.State = "open"
			state.LastStateChange = now
			cb.logger.Warn("Circuit breaker opened", map[string]interface{}{
				"service_name":  state.ServiceName,
				"failure_count": state.FailureCount,
				"threshold":     state.Threshold,
			})
		}
	} else {
		state.SuccessCount++
		now := time.Now()
		state.LastSuccess = &now

		if state.State == "half-open" {
			// Back to closed
			state.State = "closed"
			state.FailureCount = 0
			state.LastStateChange = now
			cb.logger.Info("Circuit breaker closed", map[string]interface{}{
				"service_name": state.ServiceName,
			})
		} else if state.State == "closed" {
			// Reset failure count on success
			state.FailureCount = 0
		}
	}
}

// ServiceMetricsImpl implements the ServiceMetrics interface
type ServiceMetricsImpl struct {
	metrics map[string]*ServiceMetricsData
	mu      sync.RWMutex
	logger  *observability.Logger
}

// NewServiceMetrics creates a new service metrics instance
func NewServiceMetrics(logger *observability.Logger) *ServiceMetricsImpl {
	return &ServiceMetricsImpl{
		metrics: make(map[string]*ServiceMetricsData),
		logger:  logger,
	}
}

// RecordRequest records a service request
func (m *ServiceMetricsImpl) RecordRequest(serviceName, method string, duration time.Duration, success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.metrics[serviceName] == nil {
		m.metrics[serviceName] = &ServiceMetricsData{
			ServiceName:   serviceName,
			MethodMetrics: make(map[string]MethodMetrics),
			LastUpdated:   time.Now(),
		}
	}

	// Update service-level metrics
	serviceMetrics := m.metrics[serviceName]
	serviceMetrics.RequestCount++
	serviceMetrics.LastUpdated = time.Now()

	if success {
		serviceMetrics.SuccessCount++
	} else {
		serviceMetrics.ErrorCount++
	}

	// Update method-level metrics
	if serviceMetrics.MethodMetrics[method].Method == "" {
		serviceMetrics.MethodMetrics[method] = MethodMetrics{
			Method: method,
		}
	}

	methodMetrics := serviceMetrics.MethodMetrics[method]
	methodMetrics.RequestCount++

	if success {
		methodMetrics.SuccessCount++
	} else {
		methodMetrics.ErrorCount++
	}

	// Update latency metrics (simplified)
	methodMetrics.AverageLatency = duration
	serviceMetrics.AverageLatency = duration

	serviceMetrics.MethodMetrics[method] = methodMetrics

	// Calculate rates
	if serviceMetrics.RequestCount > 0 {
		serviceMetrics.SuccessRate = float64(serviceMetrics.SuccessCount) / float64(serviceMetrics.RequestCount)
		serviceMetrics.ErrorRate = float64(serviceMetrics.ErrorCount) / float64(serviceMetrics.RequestCount)
	}

	if methodMetrics.RequestCount > 0 {
		methodMetrics.SuccessRate = float64(methodMetrics.SuccessCount) / float64(methodMetrics.RequestCount)
		methodMetrics.ErrorRate = float64(methodMetrics.ErrorCount) / float64(methodMetrics.RequestCount)
		serviceMetrics.MethodMetrics[method] = methodMetrics
	}
}

// RecordLatency records service latency
func (m *ServiceMetricsImpl) RecordLatency(serviceName, method string, latency time.Duration) {
	// Already handled in RecordRequest
}

// RecordError records a service error
func (m *ServiceMetricsImpl) RecordError(serviceName, method, errorType string) {
	// Already handled in RecordRequest
}

// GetMetrics returns metrics for a service
func (m *ServiceMetricsImpl) GetMetrics(serviceName string) ServiceMetricsData {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if metrics, exists := m.metrics[serviceName]; exists {
		return *metrics
	}

	return ServiceMetricsData{
		ServiceName: serviceName,
		LastUpdated: time.Now(),
	}
}
