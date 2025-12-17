package classification

import (
	"context"
	"sync"
	"time"

	"kyb-platform/internal/external"
)

// AsyncLLMResult represents the result of an async LLM classification
type AsyncLLMResult struct {
	ProcessingID    string                   `json:"processing_id"`
	Status          AsyncLLMStatus           `json:"status"`
	StartedAt       time.Time                `json:"started_at"`
	CompletedAt     *time.Time               `json:"completed_at,omitempty"`
	Result          *LLMClassificationResult `json:"result,omitempty"`
	Error           string                   `json:"error,omitempty"`
	OriginalResult  *IndustryDetectionResult `json:"original_result,omitempty"` // Layer 1/2 result
}

// AsyncLLMStatus represents the status of async LLM processing
type AsyncLLMStatus string

const (
	AsyncLLMStatusPending    AsyncLLMStatus = "pending"
	AsyncLLMStatusProcessing AsyncLLMStatus = "processing"
	AsyncLLMStatusCompleted  AsyncLLMStatus = "completed"
	AsyncLLMStatusFailed     AsyncLLMStatus = "failed"
	AsyncLLMStatusTimeout    AsyncLLMStatus = "timeout"
)

// AsyncLLMStore manages async LLM processing results
type AsyncLLMStore struct {
	results   map[string]*AsyncLLMResult
	mu        sync.RWMutex
	ttl       time.Duration
	maxSize   int
	cleanupCh chan struct{}
}

// NewAsyncLLMStore creates a new async LLM result store
func NewAsyncLLMStore(ttl time.Duration, maxSize int) *AsyncLLMStore {
	store := &AsyncLLMStore{
		results:   make(map[string]*AsyncLLMResult),
		ttl:       ttl,
		maxSize:   maxSize,
		cleanupCh: make(chan struct{}),
	}
	
	// Start cleanup goroutine
	go store.cleanupLoop()
	
	return store
}

// Store stores an async LLM result
func (s *AsyncLLMStore) Store(result *AsyncLLMResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Evict oldest if at max size
	if len(s.results) >= s.maxSize {
		s.evictOldest()
	}
	
	s.results[result.ProcessingID] = result
}

// Get retrieves an async LLM result by processing ID
func (s *AsyncLLMStore) Get(processingID string) (*AsyncLLMResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result, ok := s.results[processingID]
	return result, ok
}

// Update updates an existing async LLM result
func (s *AsyncLLMStore) Update(processingID string, updateFn func(*AsyncLLMResult)) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	result, ok := s.results[processingID]
	if !ok {
		return false
	}
	
	updateFn(result)
	return true
}

// Delete removes an async LLM result
func (s *AsyncLLMStore) Delete(processingID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.results, processingID)
}

// evictOldest removes the oldest result (must be called with lock held)
func (s *AsyncLLMStore) evictOldest() {
	var oldestID string
	var oldestTime time.Time
	
	for id, result := range s.results {
		if oldestID == "" || result.StartedAt.Before(oldestTime) {
			oldestID = id
			oldestTime = result.StartedAt
		}
	}
	
	if oldestID != "" {
		delete(s.results, oldestID)
	}
}

// cleanupLoop periodically removes expired results
func (s *AsyncLLMStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.cleanupCh:
			return
		}
	}
}

// cleanup removes expired results
func (s *AsyncLLMStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	for id, result := range s.results {
		if now.Sub(result.StartedAt) > s.ttl {
			delete(s.results, id)
		}
	}
}

// Stop stops the cleanup goroutine
func (s *AsyncLLMStore) Stop() {
	close(s.cleanupCh)
}

// GetStats returns statistics about the store
func (s *AsyncLLMStore) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	pending := 0
	processing := 0
	completed := 0
	failed := 0
	
	for _, result := range s.results {
		switch result.Status {
		case AsyncLLMStatusPending:
			pending++
		case AsyncLLMStatusProcessing:
			processing++
		case AsyncLLMStatusCompleted:
			completed++
		case AsyncLLMStatusFailed, AsyncLLMStatusTimeout:
			failed++
		}
	}
	
	return map[string]interface{}{
		"total":      len(s.results),
		"pending":    pending,
		"processing": processing,
		"completed":  completed,
		"failed":     failed,
		"max_size":   s.maxSize,
		"ttl":        s.ttl.String(),
	}
}

// AsyncLLMProcessor handles async LLM processing
type AsyncLLMProcessor struct {
	store         *AsyncLLMStore
	llmClassifier *LLMClassifier
	logger        interface{ Printf(format string, v ...interface{}) }
	timeout       time.Duration
}

// NewAsyncLLMProcessor creates a new async LLM processor
func NewAsyncLLMProcessor(
	store *AsyncLLMStore,
	llmClassifier *LLMClassifier,
	logger interface{ Printf(format string, v ...interface{}) },
	timeout time.Duration,
) *AsyncLLMProcessor {
	return &AsyncLLMProcessor{
		store:         store,
		llmClassifier: llmClassifier,
		logger:        logger,
		timeout:       timeout,
	}
}

// ProcessAsync starts async LLM processing and returns immediately
func (p *AsyncLLMProcessor) ProcessAsync(
	processingID string,
	ctx context.Context,
	scrapedContent *external.ScrapedContent,
	businessName string,
	description string,
	layer1Result *MultiStrategyResult,
	layer2Result *EmbeddingClassificationResult,
	originalResult *IndustryDetectionResult,
) {
	// Store initial pending state
	asyncResult := &AsyncLLMResult{
		ProcessingID:   processingID,
		Status:         AsyncLLMStatusProcessing,
		StartedAt:      time.Now(),
		OriginalResult: originalResult,
	}
	p.store.Store(asyncResult)
	
	// Process in background
	go func() {
		defer func() {
			if r := recover(); r != nil {
				p.logger.Printf("⚠️ [AsyncLLM] Panic in LLM processing for %s: %v", processingID, r)
				p.store.Update(processingID, func(result *AsyncLLMResult) {
					result.Status = AsyncLLMStatusFailed
					result.Error = "Internal error during LLM processing"
					now := time.Now()
					result.CompletedAt = &now
				})
			}
		}()
		
		p.logger.Printf("🤖 [AsyncLLM] Starting LLM processing for %s", processingID)
		
		// Create a new context with timeout for the LLM call
		llmCtx, cancel := context.WithTimeout(context.Background(), p.timeout)
		defer cancel()
		
		// Call LLM
		llmResult, err := p.llmClassifier.ClassifyWithLLM(
			llmCtx,
			scrapedContent,
			businessName,
			description,
			layer1Result,
			layer2Result,
		)
		
		// Update result
		p.store.Update(processingID, func(result *AsyncLLMResult) {
			now := time.Now()
			result.CompletedAt = &now
			
			if err != nil {
				if llmCtx.Err() == context.DeadlineExceeded {
					result.Status = AsyncLLMStatusTimeout
					result.Error = "LLM processing timed out"
					p.logger.Printf("⏱️ [AsyncLLM] Timeout for %s after %v", processingID, p.timeout)
				} else {
					result.Status = AsyncLLMStatusFailed
					result.Error = err.Error()
					p.logger.Printf("❌ [AsyncLLM] Failed for %s: %v", processingID, err)
				}
			} else {
				result.Status = AsyncLLMStatusCompleted
				result.Result = llmResult
				p.logger.Printf("✅ [AsyncLLM] Completed for %s (confidence: %.2f%%)", 
					processingID, llmResult.Confidence*100)
			}
		})
	}()
}

// GetResult retrieves the result of async LLM processing
func (p *AsyncLLMProcessor) GetResult(processingID string) (*AsyncLLMResult, bool) {
	return p.store.Get(processingID)
}

// GetStats returns processor statistics
func (p *AsyncLLMProcessor) GetStats() map[string]interface{} {
	return p.store.GetStats()
}

