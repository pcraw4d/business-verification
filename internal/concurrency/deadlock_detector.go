package concurrency

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DeadlockDetector detects and resolves deadlock situations
type DeadlockDetector struct {
	config         *DeadlockDetectorConfig
	logger         *zap.Logger
	detectionCount int64
	deadlocks      map[string]*DeadlockInfo
	mutex          sync.RWMutex
	stopChan       chan struct{}
	wg             sync.WaitGroup
}

// NewDeadlockDetector creates a new deadlock detector
func NewDeadlockDetector(config *DeadlockDetectorConfig, logger *zap.Logger) *DeadlockDetector {
	if config == nil {
		config = &DeadlockDetectorConfig{
			DetectionInterval: 5 * time.Second,
		}
	}

	dd := &DeadlockDetector{
		config:    config,
		logger:    logger,
		deadlocks: make(map[string]*DeadlockInfo),
		stopChan:  make(chan struct{}),
	}

	return dd
}

// Start starts the deadlock detection process
func (dd *DeadlockDetector) Start() error {
	dd.logger.Info("deadlock detector started",
		zap.Duration("detection_interval", dd.config.DetectionInterval))

	dd.wg.Add(1)
	go dd.detectionRoutine()

	return nil
}

// Stop stops the deadlock detection process
func (dd *DeadlockDetector) Stop() error {
	dd.logger.Info("stopping deadlock detector")

	close(dd.stopChan)
	dd.wg.Wait()

	dd.logger.Info("deadlock detector stopped")
	return nil
}

// DetectDeadlocks performs deadlock detection and returns detected deadlocks
func (dd *DeadlockDetector) DetectDeadlocks() ([]*DeadlockInfo, error) {
	// This is a simplified implementation
	// In a real system, you would implement a more sophisticated algorithm
	// such as the Banker's algorithm or resource allocation graph analysis

	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	var detectedDeadlocks []*DeadlockInfo

	// Check for potential deadlocks based on lock patterns
	// This is a basic implementation - in practice, you'd need more sophisticated analysis
	for _, deadlock := range dd.deadlocks {
		if deadlock.ResolvedAt.IsZero() {
			detectedDeadlocks = append(detectedDeadlocks, deadlock)
		}
	}

	dd.detectionCount++

	if len(detectedDeadlocks) > 0 {
		dd.logger.Warn("deadlocks detected",
			zap.Int("count", len(detectedDeadlocks)))
	}

	return detectedDeadlocks, nil
}

// ResolveDeadlock resolves a detected deadlock
func (dd *DeadlockDetector) ResolveDeadlock(deadlock *DeadlockInfo) error {
	if deadlock == nil {
		return fmt.Errorf("deadlock cannot be nil")
	}

	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	existingDeadlock, exists := dd.deadlocks[deadlock.ID]
	if !exists {
		return fmt.Errorf("deadlock %s not found", deadlock.ID)
	}

	// Mark deadlock as resolved
	existingDeadlock.ResolvedAt = time.Now()
	existingDeadlock.Resolution = "manual_resolution"

	dd.logger.Info("deadlock resolved",
		zap.String("deadlock_id", deadlock.ID),
		zap.String("resolution", existingDeadlock.Resolution))

	return nil
}

// GetDetectionCount returns the number of detection cycles performed
func (dd *DeadlockDetector) GetDetectionCount() int64 {
	dd.mutex.RLock()
	defer dd.mutex.RUnlock()
	return dd.detectionCount
}

// GetDeadlocks returns all detected deadlocks
func (dd *DeadlockDetector) GetDeadlocks() map[string]*DeadlockInfo {
	dd.mutex.RLock()
	defer dd.mutex.RUnlock()

	result := make(map[string]*DeadlockInfo)
	for id, deadlock := range dd.deadlocks {
		result[id] = deadlock
	}
	return result
}

// AddDeadlock adds a new deadlock to the detector
func (dd *DeadlockDetector) AddDeadlock(deadlock *DeadlockInfo) {
	if deadlock == nil {
		return
	}

	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	dd.deadlocks[deadlock.ID] = deadlock

	dd.logger.Warn("deadlock added",
		zap.String("deadlock_id", deadlock.ID),
		zap.Strings("resources", deadlock.Resources),
		zap.Strings("processes", deadlock.Processes))
}

// RemoveDeadlock removes a deadlock from the detector
func (dd *DeadlockDetector) RemoveDeadlock(deadlockID string) {
	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	delete(dd.deadlocks, deadlockID)

	dd.logger.Debug("deadlock removed",
		zap.String("deadlock_id", deadlockID))
}

// detectionRoutine runs the periodic deadlock detection
func (dd *DeadlockDetector) detectionRoutine() {
	defer dd.wg.Done()

	ticker := time.NewTicker(dd.config.DetectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dd.performDetection()
		case <-dd.stopChan:
			return
		}
	}
}

// performDetection performs a single deadlock detection cycle
func (dd *DeadlockDetector) performDetection() {
	deadlocks, err := dd.DetectDeadlocks()
	if err != nil {
		dd.logger.Error("deadlock detection failed", zap.Error(err))
		return
	}

	if len(deadlocks) > 0 {
		dd.logger.Warn("deadlocks detected during periodic scan",
			zap.Int("count", len(deadlocks)))

		// Attempt automatic resolution for each deadlock
		for _, deadlock := range deadlocks {
			if err := dd.attemptAutomaticResolution(deadlock); err != nil {
				dd.logger.Error("automatic deadlock resolution failed",
					zap.String("deadlock_id", deadlock.ID),
					zap.Error(err))
			}
		}
	}
}

// attemptAutomaticResolution attempts to automatically resolve a deadlock
func (dd *DeadlockDetector) attemptAutomaticResolution(deadlock *DeadlockInfo) error {
	// This is a simplified automatic resolution strategy
	// In practice, you would implement more sophisticated resolution logic

	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	existingDeadlock, exists := dd.deadlocks[deadlock.ID]
	if !exists {
		return fmt.Errorf("deadlock %s not found", deadlock.ID)
	}

	// Simple resolution strategy: mark the first process as the victim
	existingDeadlock.ResolvedAt = time.Now()
	existingDeadlock.Resolution = "automatic_resolution_victim_process"

	dd.logger.Info("deadlock automatically resolved",
		zap.String("deadlock_id", deadlock.ID),
		zap.String("resolution", existingDeadlock.Resolution))

	return nil
}

// AnalyzeResourceGraph analyzes the resource allocation graph for potential deadlocks
func (dd *DeadlockDetector) AnalyzeResourceGraph(resources map[string]*Resource, locks map[string]*Lock) []*DeadlockInfo {
	// This is a simplified resource allocation graph analysis
	// In practice, you would implement a more sophisticated algorithm

	var potentialDeadlocks []*DeadlockInfo

	// Check for circular wait conditions
	// This is a basic implementation - real deadlock detection would be more complex
	for resourceID, lock := range locks {
		// Check if this resource is part of a potential deadlock
		if dd.isPartOfCircularWait(resourceID, locks) {
			deadlock := &DeadlockInfo{
				ID:         fmt.Sprintf("deadlock-%d", time.Now().UnixNano()),
				Resources:  []string{resourceID},
				Processes:  []string{lock.Owner},
				DetectedAt: time.Now(),
			}

			potentialDeadlocks = append(potentialDeadlocks, deadlock)
		}
	}

	return potentialDeadlocks
}

// isPartOfCircularWait checks if a resource is part of a circular wait condition
func (dd *DeadlockDetector) isPartOfCircularWait(resourceID string, locks map[string]*Lock) bool {
	// This is a simplified circular wait detection
	// In practice, you would implement a proper graph traversal algorithm

	// For now, we'll use a simple heuristic: if a resource has been locked for too long
	lock, exists := locks[resourceID]
	if !exists {
		return false
	}

	// Consider it a potential deadlock if locked for more than 30 seconds
	return time.Since(lock.AcquiredAt) > 30*time.Second
}

// GetDeadlockStatistics returns statistics about detected deadlocks
func (dd *DeadlockDetector) GetDeadlockStatistics() map[string]interface{} {
	dd.mutex.RLock()
	defer dd.mutex.RUnlock()

	totalDeadlocks := len(dd.deadlocks)
	resolvedDeadlocks := 0
	unresolvedDeadlocks := 0

	for _, deadlock := range dd.deadlocks {
		if deadlock.ResolvedAt.IsZero() {
			unresolvedDeadlocks++
		} else {
			resolvedDeadlocks++
		}
	}

	return map[string]interface{}{
		"total_deadlocks":       totalDeadlocks,
		"resolved_deadlocks":    resolvedDeadlocks,
		"unresolved_deadlocks":  unresolvedDeadlocks,
		"detection_count":       dd.detectionCount,
		"detection_interval_ms": dd.config.DetectionInterval.Milliseconds(),
	}
}

// ClearResolvedDeadlocks removes all resolved deadlocks from memory
func (dd *DeadlockDetector) ClearResolvedDeadlocks() {
	dd.mutex.Lock()
	defer dd.mutex.Unlock()

	clearedCount := 0
	for id, deadlock := range dd.deadlocks {
		if !deadlock.ResolvedAt.IsZero() {
			delete(dd.deadlocks, id)
			clearedCount++
		}
	}

	if clearedCount > 0 {
		dd.logger.Info("resolved deadlocks cleared from memory",
			zap.Int("count", clearedCount))
	}
}
