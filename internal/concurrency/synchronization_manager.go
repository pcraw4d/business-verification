package concurrency

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// SynchronizationManager manages locks, mutexes, and synchronization primitives
type SynchronizationManager struct {
	config           *SynchronizationManagerConfig
	logger           *zap.Logger
	locks            map[string]*Lock
	lockRequests     map[string][]*LockRequest
	mutex            sync.RWMutex
	deadlockDetector *DeadlockDetector
	stopChan         chan struct{}
}

// NewSynchronizationManager creates a new synchronization manager
func NewSynchronizationManager(config *SynchronizationManagerConfig, logger *zap.Logger) *SynchronizationManager {
	if config == nil {
		config = &SynchronizationManagerConfig{
			EnableDeadlockDetection: true,
		}
	}

	sm := &SynchronizationManager{
		config:       config,
		logger:       logger,
		locks:        make(map[string]*Lock),
		lockRequests: make(map[string][]*LockRequest),
		stopChan:     make(chan struct{}),
	}

	if config.EnableDeadlockDetection {
		sm.deadlockDetector = NewDeadlockDetector(&DeadlockDetectorConfig{
			DetectionInterval: 5 * time.Second,
		}, logger)
	}

	return sm
}

// AcquireLock acquires a lock on a resource
func (sm *SynchronizationManager) AcquireLock(ctx context.Context, request *LockRequest) (*Lock, error) {
	if request == nil {
		return nil, fmt.Errorf("lock request cannot be nil")
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if resource is already locked
	if existingLock, exists := sm.locks[request.ResourceID]; exists {
		if time.Now().Before(existingLock.ExpiresAt) {
			// Resource is locked, add to wait queue
			sm.lockRequests[request.ResourceID] = append(sm.lockRequests[request.ResourceID], request)

			sm.logger.Debug("lock request queued",
				zap.String("resource_id", request.ResourceID),
				zap.Duration("timeout", request.Timeout))

			// Wait for lock to be released or timeout
			sm.mutex.Unlock()

			select {
			case <-ctx.Done():
				sm.mutex.Lock()
				sm.removeLockRequest(request.ResourceID, request)
				sm.mutex.Unlock()
				return nil, ctx.Err()
			case <-time.After(request.Timeout):
				sm.mutex.Lock()
				sm.removeLockRequest(request.ResourceID, request)
				sm.mutex.Unlock()
				return nil, fmt.Errorf("lock acquisition timeout")
			}

			sm.mutex.Lock()
		} else {
			// Lock has expired, remove it
			delete(sm.locks, request.ResourceID)
		}
	}

	// Acquire the lock
	lock := &Lock{
		ResourceID:  request.ResourceID,
		AcquiredAt:  time.Now(),
		ExpiresAt:   time.Now().Add(request.Timeout),
		Owner:       fmt.Sprintf("process-%d", time.Now().UnixNano()),
		IsExclusive: true,
	}

	sm.locks[request.ResourceID] = lock

	sm.logger.Debug("lock acquired",
		zap.String("resource_id", request.ResourceID),
		zap.String("owner", lock.Owner),
		zap.Time("expires_at", lock.ExpiresAt))

	return lock, nil
}

// ReleaseLock releases a previously acquired lock
func (sm *SynchronizationManager) ReleaseLock(lock *Lock) error {
	if lock == nil {
		return fmt.Errorf("lock cannot be nil")
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	existingLock, exists := sm.locks[lock.ResourceID]
	if !exists {
		return fmt.Errorf("lock not found for resource %s", lock.ResourceID)
	}

	if existingLock.Owner != lock.Owner {
		return fmt.Errorf("lock owned by different process")
	}

	// Release the lock
	delete(sm.locks, lock.ResourceID)

	sm.logger.Debug("lock released",
		zap.String("resource_id", lock.ResourceID),
		zap.String("owner", lock.Owner))

	// Process waiting lock requests
	if requests, exists := sm.lockRequests[lock.ResourceID]; exists && len(requests) > 0 {
		// Grant lock to the next request in queue
		nextRequest := requests[0]
		sm.lockRequests[lock.ResourceID] = requests[1:]

		newLock := &Lock{
			ResourceID:  lock.ResourceID,
			AcquiredAt:  time.Now(),
			ExpiresAt:   time.Now().Add(nextRequest.Timeout),
			Owner:       fmt.Sprintf("process-%d", time.Now().UnixNano()),
			IsExclusive: true,
		}

		sm.locks[lock.ResourceID] = newLock

		sm.logger.Debug("lock granted to waiting request",
			zap.String("resource_id", lock.ResourceID),
			zap.String("owner", newLock.Owner))
	}

	return nil
}

// IsLocked checks if a resource is currently locked
func (sm *SynchronizationManager) IsLocked(resourceID string) bool {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	lock, exists := sm.locks[resourceID]
	if !exists {
		return false
	}

	// Check if lock has expired
	if time.Now().After(lock.ExpiresAt) {
		return false
	}

	return true
}

// GetLockInfo returns information about a lock
func (sm *SynchronizationManager) GetLockInfo(resourceID string) (*Lock, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	lock, exists := sm.locks[resourceID]
	if !exists {
		return nil, false
	}

	// Check if lock has expired
	if time.Now().After(lock.ExpiresAt) {
		return nil, false
	}

	return lock, true
}

// GetLockedResources returns all currently locked resources
func (sm *SynchronizationManager) GetLockedResources() map[string]*Lock {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := make(map[string]*Lock)
	now := time.Now()

	for resourceID, lock := range sm.locks {
		if now.Before(lock.ExpiresAt) {
			result[resourceID] = lock
		}
	}

	return result
}

// GetWaitingRequests returns all waiting lock requests
func (sm *SynchronizationManager) GetWaitingRequests() map[string][]*LockRequest {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	result := make(map[string][]*LockRequest)
	for resourceID, requests := range sm.lockRequests {
		if len(requests) > 0 {
			result[resourceID] = requests
		}
	}

	return result
}

// CleanupExpiredLocks removes expired locks
func (sm *SynchronizationManager) CleanupExpiredLocks() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for resourceID, lock := range sm.locks {
		if now.After(lock.ExpiresAt) {
			delete(sm.locks, resourceID)
			expiredCount++

			sm.logger.Debug("expired lock removed",
				zap.String("resource_id", resourceID),
				zap.String("owner", lock.Owner))
		}
	}

	if expiredCount > 0 {
		sm.logger.Info("expired locks cleaned up",
			zap.Int("count", expiredCount))
	}
}

// Start starts the synchronization manager
func (sm *SynchronizationManager) Start() error {
	sm.logger.Info("synchronization manager started")

	if sm.deadlockDetector != nil {
		if err := sm.deadlockDetector.Start(); err != nil {
			return fmt.Errorf("failed to start deadlock detector: %w", err)
		}
	}

	// Start cleanup goroutine
	go sm.cleanupRoutine()

	return nil
}

// Stop stops the synchronization manager
func (sm *SynchronizationManager) Stop() error {
	sm.logger.Info("stopping synchronization manager")

	close(sm.stopChan)

	if sm.deadlockDetector != nil {
		if err := sm.deadlockDetector.Stop(); err != nil {
			sm.logger.Error("failed to stop deadlock detector", zap.Error(err))
		}
	}

	sm.logger.Info("synchronization manager stopped")
	return nil
}

// cleanupRoutine periodically cleans up expired locks
func (sm *SynchronizationManager) cleanupRoutine() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			sm.CleanupExpiredLocks()
		case <-sm.stopChan:
			return
		}
	}
}

// removeLockRequest removes a lock request from the wait queue
func (sm *SynchronizationManager) removeLockRequest(resourceID string, request *LockRequest) {
	requests, exists := sm.lockRequests[resourceID]
	if !exists {
		return
	}

	for i, req := range requests {
		if req == request {
			sm.lockRequests[resourceID] = append(requests[:i], requests[i+1:]...)
			break
		}
	}

	if len(sm.lockRequests[resourceID]) == 0 {
		delete(sm.lockRequests, resourceID)
	}
}

// GetDeadlockDetector returns the deadlock detector
func (sm *SynchronizationManager) GetDeadlockDetector() *DeadlockDetector {
	return sm.deadlockDetector
}

// ForceReleaseLock forcefully releases a lock (admin function)
func (sm *SynchronizationManager) ForceReleaseLock(resourceID string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	lock, exists := sm.locks[resourceID]
	if !exists {
		return fmt.Errorf("lock not found for resource %s", resourceID)
	}

	delete(sm.locks, resourceID)

	sm.logger.Warn("lock forcefully released",
		zap.String("resource_id", resourceID),
		zap.String("owner", lock.Owner))

	return nil
}
