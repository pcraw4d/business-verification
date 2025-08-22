package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ResourceManager manages system resources with thread-safe operations
type ResourceManager struct {
	config    *ResourceManagerConfig
	logger    *zap.Logger
	resources map[string]*Resource
	waitQueue map[string][]*LockRequest
	mutex     sync.RWMutex
	stats     *ResourceStats
	stopChan  chan struct{}
}

// NewResourceManager creates a new resource manager
func NewResourceManager(config *ResourceManagerConfig, logger *zap.Logger) *ResourceManager {
	if config == nil {
		config = &ResourceManagerConfig{
			MaxConcurrentOps: 100,
			ResourceTimeout:  30 * time.Second,
		}
	}

	rm := &ResourceManager{
		config:    config,
		logger:    logger,
		resources: make(map[string]*Resource),
		waitQueue: make(map[string][]*LockRequest),
		stats: &ResourceStats{
			LastUpdated: time.Now(),
		},
		stopChan: make(chan struct{}),
	}

	return rm
}

// RegisterResource registers a new resource with the manager
func (rm *ResourceManager) RegisterResource(id, resourceType string, capacity int) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, exists := rm.resources[id]; exists {
		return fmt.Errorf("resource %s already registered", id)
	}

	rm.resources[id] = &Resource{
		ID:        id,
		Type:      resourceType,
		Capacity:  capacity,
		Used:      0,
		Available: capacity,
	}

	rm.logger.Info("resource registered",
		zap.String("resource_id", id),
		zap.String("type", resourceType),
		zap.Int("capacity", capacity))

	rm.updateStats()
	return nil
}

// Acquire acquires resources with timeout and priority
func (rm *ResourceManager) Acquire(ctx context.Context, resources []string) ([]*Resource, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Check if all resources exist
	for _, resourceID := range resources {
		if _, exists := rm.resources[resourceID]; !exists {
			return nil, fmt.Errorf("resource %s not found", resourceID)
		}
	}

	// Try to acquire all resources immediately
	acquired := make([]*Resource, 0, len(resources))
	for _, resourceID := range resources {
		resource := rm.resources[resourceID]
		if resource.Available > 0 {
			resource.Used++
			resource.Available--
			resource.AcquiredAt = time.Now()
			acquired = append(acquired, resource)
		} else {
			// Release any already acquired resources
			for _, acquiredResource := range acquired {
				acquiredResource.Used--
				acquiredResource.Available++
			}
			return nil, fmt.Errorf("resource %s not available", resourceID)
		}
	}

	rm.logger.Debug("resources acquired",
		zap.Strings("resource_ids", resources),
		zap.Int("count", len(acquired)))

	rm.updateStats()
	return acquired, nil
}

// Release releases previously acquired resources
func (rm *ResourceManager) Release(resources []*Resource) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	for _, resource := range resources {
		if resource.Used > 0 {
			resource.Used--
			resource.Available++
			resource.ReleasedAt = time.Now()
		} else {
			rm.logger.Warn("attempted to release unused resource",
				zap.String("resource_id", resource.ID))
		}
	}

	rm.logger.Debug("resources released",
		zap.Int("count", len(resources)))

	rm.updateStats()
	return nil
}

// GetResource returns a resource by ID
func (rm *ResourceManager) GetResource(id string) (*Resource, bool) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	resource, exists := rm.resources[id]
	return resource, exists
}

// GetResources returns all resources
func (rm *ResourceManager) GetResources() map[string]*Resource {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	result := make(map[string]*Resource)
	for id, resource := range rm.resources {
		result[id] = resource
	}
	return result
}

// GetStats returns resource usage statistics
func (rm *ResourceManager) GetStats() *ResourceStats {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	return rm.stats
}

// UpdateStats updates the resource statistics
func (rm *ResourceManager) UpdateStats() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	rm.updateStats()
}

// updateStats updates the resource statistics
func (rm *ResourceManager) updateStats() {
	totalResources := 0
	availableResources := 0

	for _, resource := range rm.resources {
		totalResources += resource.Capacity
		availableResources += resource.Available
	}

	utilization := 0.0
	if totalResources > 0 {
		utilization = float64(totalResources-availableResources) / float64(totalResources)
	}

	rm.stats = &ResourceStats{
		TotalResources:     totalResources,
		AvailableResources: availableResources,
		Utilization:        utilization,
		LastUpdated:        time.Now(),
	}
}

// Start starts the resource manager
func (rm *ResourceManager) Start() error {
	rm.logger.Info("resource manager started")
	return nil
}

// Stop stops the resource manager
func (rm *ResourceManager) Stop() error {
	close(rm.stopChan)
	rm.logger.Info("resource manager stopped")
	return nil
}

// CleanupExpiredResources cleans up expired resource acquisitions
func (rm *ResourceManager) CleanupExpiredResources() {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	now := time.Now()
	for _, resource := range rm.resources {
		if !resource.AcquiredAt.IsZero() &&
			now.Sub(resource.AcquiredAt) > rm.config.ResourceTimeout {
			// Force release expired resources
			if resource.Used > 0 {
				resource.Used--
				resource.Available++
				resource.ReleasedAt = now

				rm.logger.Warn("expired resource released",
					zap.String("resource_id", resource.ID),
					zap.Duration("duration", now.Sub(resource.AcquiredAt)))
			}
		}
	}

	rm.updateStats()
}

// GetResourceUtilization returns the utilization percentage for a specific resource
func (rm *ResourceManager) GetResourceUtilization(resourceID string) (float64, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	resource, exists := rm.resources[resourceID]
	if !exists {
		return 0, fmt.Errorf("resource %s not found", resourceID)
	}

	if resource.Capacity == 0 {
		return 0, nil
	}

	return float64(resource.Used) / float64(resource.Capacity), nil
}

// GetResourceWaitTime returns the average wait time for a resource
func (rm *ResourceManager) GetResourceWaitTime(resourceID string) time.Duration {
	// This is a simplified implementation
	// In a real system, you would track actual wait times
	return rm.stats.WaitTime
}
