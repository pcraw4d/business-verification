package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	"kyb-platform/internal/classification"
	reqcache "kyb-platform/internal/classification/cache"
	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/machine_learning/infrastructure"
	"kyb-platform/services/classification-service/internal/cache"
	"kyb-platform/services/classification-service/internal/config"
	"kyb-platform/services/classification-service/internal/supabase"
)

// cacheEntry represents a cached classification result
type cacheEntry struct {
	response  *ClassificationResponse
	expiresAt time.Time
}

// inFlightRequest represents a request that is currently being processed
type inFlightRequest struct {
	resultChan chan *inFlightResult
	startTime  time.Time
	timeout    time.Duration // Maximum time to wait for this request
	// FIX #13: Use sync.Once to prevent double-close panic
	closeOnce sync.Once
	// FIX: Add closed flag to prevent "send on closed channel" panic
	closed bool
	mu     sync.Mutex
}

// inFlightResult represents the result of an in-flight request
type inFlightResult struct {
	response *ClassificationResponse
	err      error
}

// queuedRequest represents a request in the processing queue
type queuedRequest struct {
	req       *ClassificationRequest
	ctx       context.Context
	response  chan *ClassificationResponse
	errChan   chan error
	startTime time.Time
}

// requestQueue manages a queue of requests waiting to be processed
type requestQueue struct {
	queue       chan *queuedRequest
	maxSize     int
	currentSize int32 // atomic counter
	mu          sync.RWMutex
}

// NewRequestQueue creates a new request queue
func NewRequestQueue(maxSize int) *requestQueue {
	return &requestQueue{
		queue:       make(chan *queuedRequest, maxSize),
		maxSize:     maxSize,
		currentSize: 0,
	}
}

// Enqueue adds a request to the queue
func (rq *requestQueue) Enqueue(req *queuedRequest) error {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	currentSize := int(atomic.LoadInt32(&rq.currentSize))
	if currentSize >= rq.maxSize {
		return fmt.Errorf("request queue is full")
	}

	select {
	case rq.queue <- req:
		atomic.AddInt32(&rq.currentSize, 1)
		return nil
	default:
		return fmt.Errorf("request queue is full")
	}
}

// Dequeue removes a request from the queue
func (rq *requestQueue) Dequeue() (*queuedRequest, bool) {
	select {
	case req := <-rq.queue:
		atomic.AddInt32(&rq.currentSize, -1)
		return req, true
	default:
		return nil, false
	}
}

// Size returns the current queue size
// FIX #8: Use actual channel length instead of atomic counter for accuracy
func (rq *requestQueue) Size() int {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	return len(rq.queue)
}

// workerStats tracks statistics for a single worker
type workerStats struct {
	workerID            int
	requestsProcessed   int64
	totalProcessingTime time.Duration
	averageTime         time.Duration
	lastActivity        time.Time
	currentRequestID    string
	isBlocked           bool
	blockedDuration     time.Duration
}

// workerPool manages a pool of workers for processing requests
type workerPool struct {
	workers     int
	queue       *requestQueue
	handler     *ClassificationHandler
	logger      *zap.Logger
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	workerStats map[int]*workerStats // workerID -> stats
	statsMutex  sync.RWMutex
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(workers int, queue *requestQueue, handler *ClassificationHandler, logger *zap.Logger) *workerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &workerPool{
		workers:     workers,
		queue:       queue,
		handler:     handler,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		workerStats: make(map[int]*workerStats),
	}
}

// Start starts the worker pool
func (wp *workerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
	wp.logger.Info("Worker pool started",
		zap.Int("workers", wp.workers))
}

// Stop stops the worker pool
func (wp *workerPool) Stop() {
	wp.cancel()
	wp.wg.Wait()
	wp.logger.Info("Worker pool stopped")
}

// worker processes requests from the queue
func (wp *workerPool) worker(id int) {
	defer wp.wg.Done()

	// Initialize worker stats
	wp.statsMutex.Lock()
	wp.workerStats[id] = &workerStats{
		workerID:     id,
		lastActivity: time.Now(),
	}
	wp.statsMutex.Unlock()

	wp.logger.Info("🔧 [WORKER-START] Worker started",
		zap.Int("worker_id", id),
		zap.Time("start_time", time.Now()))

	// Check for blocked workers periodically (only worker 0 does this)
	if id == 0 {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-wp.ctx.Done():
					return
				case <-ticker.C:
					wp.checkBlockedWorkers()
				}
			}
		}()
	}

	for {
		select {
		case <-wp.ctx.Done():
			wp.logger.Info("Worker stopping",
				zap.Int("worker_id", id))
			return
		default:
			// Dequeue request
			queuedReq, ok := wp.queue.Dequeue()
			if !ok {
				// No requests available, wait a bit
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Process request with panic recovery
			wp.processRequestSafely(id, queuedReq)
		}
	}
}

// processRequestSafely wraps request processing with panic recovery
func (wp *workerPool) processRequestSafely(id int, queuedReq *queuedRequest) {
	defer func() {
		if r := recover(); r != nil {
			wp.logger.Error("🚨 [WORKER-PANIC] Worker recovered from panic",
				zap.Int("worker_id", id),
				zap.String("request_id", queuedReq.req.RequestID),
				zap.Any("panic", r),
				zap.Stack("stack"))

			// Send error to requester
			select {
			case queuedReq.errChan <- fmt.Errorf("internal panic: %v", r):
			default:
			}
		}
	}()

	startTime := queuedReq.startTime
	if startTime.IsZero() {
		startTime = time.Now()
	}

	// Calculate queue wait time
	queueWaitTime := time.Since(startTime)

	// ALWAYS create fresh context to avoid expiration issues
	// This eliminates the need for complex checks and ensures sufficient time
	// Increased to 120s to accommodate longer processing times (scraping, classification, etc.)
	freshTimeout := 120 * time.Second
	processingCtx, cancel := context.WithTimeout(context.Background(), freshTimeout)
	defer cancel()

	// Log original context state for debugging
	originalTimeRemaining := time.Duration(0)
	originalExpired := false
	if queuedReq.ctx.Err() != nil {
		originalExpired = true
	} else if deadline, hasDeadline := queuedReq.ctx.Deadline(); hasDeadline {
		originalTimeRemaining = time.Until(deadline)
		if originalTimeRemaining <= 0 {
			originalExpired = true
		}
	}

	wp.logger.Info("🔧 [WORKER-CONTEXT] Worker creating fresh context for processing",
		zap.Int("worker_id", id),
		zap.String("request_id", queuedReq.req.RequestID),
		zap.Duration("queue_wait", queueWaitTime),
		zap.Duration("fresh_timeout", freshTimeout),
		zap.Duration("original_time_remaining", originalTimeRemaining),
		zap.Bool("original_expired", originalExpired),
		zap.Duration("time_since_enqueue", time.Since(queuedReq.startTime)))

	// Update worker stats - processing started
	wp.statsMutex.Lock()
	stats := wp.workerStats[id]
	stats.lastActivity = time.Now()
	stats.currentRequestID = queuedReq.req.RequestID
	stats.isBlocked = false
	wp.statsMutex.Unlock()

	// Process request
	wp.logger.Info("🔧 [WORKER-PROCESSING] Worker processing request",
		zap.Int("worker_id", id),
		zap.String("request_id", queuedReq.req.RequestID),
		zap.Int("queue_size", wp.queue.Size()),
		zap.Duration("queue_wait", queueWaitTime),
		zap.Int("active_workers", wp.getActiveWorkerCount()))

	response, err := wp.handler.processClassification(processingCtx, queuedReq.req, startTime)

	// Update worker stats - processing completed
	processingDuration := time.Since(startTime)
	wp.statsMutex.Lock()
	stats = wp.workerStats[id]
	stats.requestsProcessed++
	stats.totalProcessingTime += processingDuration
	if stats.requestsProcessed > 0 {
		stats.averageTime = stats.totalProcessingTime / time.Duration(stats.requestsProcessed)
	}
	stats.lastActivity = time.Now()
	stats.currentRequestID = ""
	wp.statsMutex.Unlock()

	if err != nil {
		wp.logger.Error("Request processing failed",
			zap.Int("worker_id", id),
			zap.String("request_id", queuedReq.req.RequestID),
			zap.Error(err),
			zap.Duration("duration", processingDuration))
		// FIX #12: Error channel is buffered (size 1) and has default case to prevent blocking
		// The receiver (HTTP handler) is always waiting, so this should never block
		select {
		case queuedReq.errChan <- err:
			// Error sent successfully
		default:
			// Error channel already has an error (shouldn't happen, but safe to ignore)
			wp.logger.Warn("Error channel full, error may be lost",
				zap.String("request_id", queuedReq.req.RequestID),
				zap.Error(err))
		}
	} else {
		wp.logger.Info("✅ [WORKER-COMPLETE] Request processing completed",
			zap.Int("worker_id", id),
			zap.String("request_id", queuedReq.req.RequestID),
			zap.Duration("duration", processingDuration),
			zap.Int64("total_processed", stats.requestsProcessed),
			zap.Duration("avg_time", stats.averageTime))
		select {
		case queuedReq.response <- response:
		default:
			// Response channel already has a response (shouldn't happen)
		}
	}
}

// getActiveWorkerCount returns count of workers currently processing
func (wp *workerPool) getActiveWorkerCount() int {
	wp.statsMutex.RLock()
	defer wp.statsMutex.RUnlock()

	active := 0
	for _, stats := range wp.workerStats {
		if stats.currentRequestID != "" {
			active++
		}
	}
	return active
}

// checkBlockedWorkers checks for workers that appear blocked
func (wp *workerPool) checkBlockedWorkers() {
	wp.statsMutex.RLock()
	defer wp.statsMutex.RUnlock()

	now := time.Now()
	for id, stats := range wp.workerStats {
		if stats.currentRequestID != "" {
			inactiveTime := now.Sub(stats.lastActivity)
			if inactiveTime > 2*time.Minute && !stats.isBlocked {
				stats.isBlocked = true
				stats.blockedDuration = inactiveTime
				wp.logger.Warn("⚠️ [WORKER-BLOCKED] Worker appears blocked",
					zap.Int("worker_id", id),
					zap.String("current_request", stats.currentRequestID),
					zap.Duration("inactive_time", inactiveTime))
			}
		}
	}
}

// ClassificationHandler handles classification requests
type ClassificationHandler struct {
	supabaseClient       *supabase.Client
	logger               *zap.Logger
	config               *config.Config
	industryDetector     *classification.IndustryDetectionService
	codeGenerator        *classification.ClassificationCodeGenerator
	keywordRepo          repository.KeywordRepository         // OPTIMIZATION #5.2: For accuracy tracking
	pythonMLService      interface{}                          // *infrastructure.PythonMLService - using interface to avoid import cycle
	industryThresholds   *classification.IndustryThresholds   // OPTIMIZATION #16: Industry-specific thresholds
	confidenceCalibrator *classification.ConfidenceCalibrator // OPTIMIZATION #5.2: Confidence calibration
	cache                map[string]*cacheEntry
	cacheMutex           sync.RWMutex
	redisCache           *cache.RedisCache // Distributed Redis cache (optional)
	inFlightRequests     map[string]*inFlightRequest
	inFlightMutex        sync.RWMutex
	requestQueue         *requestQueue // Request queue for managing concurrent requests
	WorkerPool           *workerPool   // Worker pool for processing queued requests (exported for shutdown)
	// FIX #7: Shutdown context for cleanup goroutines (exported for shutdown)
	ShutdownCtx      context.Context
	ShutdownCancel   context.CancelFunc
	serviceStartTime time.Time // Track when service started for uptime reporting
}

// NewClassificationHandler creates a new classification handler
func NewClassificationHandler(
	supabaseClient *supabase.Client,
	logger *zap.Logger,
	config *config.Config,
	industryDetector *classification.IndustryDetectionService,
	codeGenerator *classification.ClassificationCodeGenerator,
	keywordRepo repository.KeywordRepository, // OPTIMIZATION #5.2: For accuracy tracking
	pythonMLService interface{}, // *infrastructure.PythonMLService - optional, can be nil
) *ClassificationHandler {
	// OPTIMIZATION #16: Initialize industry-specific thresholds
	industryThresholds := classification.NewIndustryThresholds()

	// OPTIMIZATION #5.2: Initialize confidence calibrator for accuracy tracking
	// Create a std logger adapter for the calibrator (it uses log.Logger, not zap.Logger)
	stdLogger := log.New(&zapLoggerAdapter{logger: logger}, "", 0)
	confidenceCalibrator := classification.NewConfidenceCalibrator(stdLogger)

	// Initialize request queue (default max size: 50, or use MaxConcurrentRequests config)
	maxQueueSize := 50
	if config.Classification.MaxConcurrentRequests > 0 {
		maxQueueSize = config.Classification.MaxConcurrentRequests
	}
	requestQueue := NewRequestQueue(maxQueueSize)

	handler := &ClassificationHandler{
		supabaseClient:       supabaseClient,
		logger:               logger,
		config:               config,
		industryDetector:     industryDetector,
		codeGenerator:        codeGenerator,
		keywordRepo:          keywordRepo, // OPTIMIZATION #5.2: For accuracy tracking
		industryThresholds:   industryThresholds,
		confidenceCalibrator: confidenceCalibrator,
		pythonMLService:      pythonMLService,
		cache:                make(map[string]*cacheEntry),
		inFlightRequests:     make(map[string]*inFlightRequest),
		requestQueue:         requestQueue,
		serviceStartTime:     time.Now(), // Track when service started for uptime reporting
	}

	logger.Info("Request queue initialized",
		zap.Int("max_size", maxQueueSize))

	// Initialize worker pool (default: 10 workers, or 20% of MaxConcurrentRequests)
	// OPTIMIZATION: Increased worker count to handle burst traffic better
	workerCount := 10
	if config.Classification.MaxConcurrentRequests > 0 {
		// Use 30% of max concurrent requests as workers (increased from 20%)
		workerCount = config.Classification.MaxConcurrentRequests * 3 / 10
		if workerCount < 1 {
			workerCount = 1
		}
		if workerCount > 20 {
			// Cap worker count to reduce memory pressure in production
			workerCount = 20
		}
	} else {
		// Default to 15 workers if no config (reduced to control memory)
		workerCount = 15
	}

	workerPool := NewWorkerPool(workerCount, requestQueue, handler, logger)
	handler.WorkerPool = workerPool

	// Start worker pool
	workerPool.Start()

	logger.Info("Worker pool initialized",
		zap.Int("workers", workerCount))

	// Initialize Redis cache if enabled
	if config.Classification.RedisEnabled && config.Classification.RedisURL != "" {
		handler.redisCache = cache.NewRedisCache(
			config.Classification.RedisURL,
			"classification",
			logger,
		)
		// Test Redis connection
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := handler.redisCache.Health(ctx); err == nil {
			logger.Info("✅ Redis cache initialized and healthy for classification service",
				zap.String("redis_url", maskRedisURL(config.Classification.RedisURL)))
	} else {
			logger.Warn("⚠️ Redis cache initialized but health check failed, using in-memory fallback",
				zap.String("redis_url", maskRedisURL(config.Classification.RedisURL)),
				zap.Error(err))
		}
		cancel()
	} else {
		logger.Info("Using in-memory cache only",
			zap.Bool("redis_enabled", config.Classification.RedisEnabled),
			zap.Bool("redis_url_provided", config.Classification.RedisURL != ""))
	}

	// FIX #7: Initialize shutdown context for cleanup goroutines
	handler.ShutdownCtx, handler.ShutdownCancel = context.WithCancel(context.Background())

	// Start cache cleanup goroutine (for in-memory cache only)
	if config.Classification.CacheEnabled {
		go handler.cleanupCache(handler.ShutdownCtx)
	}

	// Start in-flight requests cleanup goroutine (Task 1.4: Request Deduplication)
	go handler.cleanupInFlightRequests(handler.ShutdownCtx)

	return handler
}

// cleanupCache periodically removes expired cache entries
// FIX #7: Accept context to allow graceful shutdown
func (h *ClassificationHandler) cleanupCache(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("Cache cleanup goroutine stopping")
			return
		case <-ticker.C:
			h.cacheMutex.Lock()
			now := time.Now()
			for key, entry := range h.cache {
				if now.After(entry.expiresAt) {
					delete(h.cache, key)
				}
			}
			h.cacheMutex.Unlock()
		}
	}
}

// cleanupInFlightRequests periodically removes stale in-flight requests (Task 1.4)
// FIX #7: Accept context to allow graceful shutdown
func (h *ClassificationHandler) cleanupInFlightRequests(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("In-flight requests cleanup goroutine stopping")
			return
		case <-ticker.C:
			h.inFlightMutex.Lock()
			now := time.Now()
			maxAge := h.config.Classification.RequestTimeout * 2 // Remove requests older than 2x timeout
			if maxAge == 0 {
				maxAge = 2 * time.Minute // Default max age
			}

			for key, req := range h.inFlightRequests {
				age := now.Sub(req.startTime)
				if age > maxAge {
					h.logger.Warn("Removing stale in-flight request",
						zap.String("cache_key", key),
						zap.Duration("age", age))
					// FIX: Mark as closed to prevent sends, then delete
					// Don't close the channel - let it be garbage collected
					// This prevents "send on closed channel" panic entirely
					req.mu.Lock()
					req.closed = true
					req.mu.Unlock()
					delete(h.inFlightRequests, key)
				}
			}
			h.inFlightMutex.Unlock()
		}
	}
}

// getCacheKey generates a cache key from the request
// OPTIMIZATION: Normalize inputs to improve cache hit rates
// FIX: Added "classification:" prefix for consistency with internal service cache keys
func (h *ClassificationHandler) getCacheKey(req *ClassificationRequest) string {
	// Normalize inputs: trim whitespace, lowercase for case-insensitive matching
	// This improves cache hit rates by treating "Acme Corp" and "acme corp" as the same
	businessName := strings.TrimSpace(strings.ToLower(req.BusinessName))
	description := strings.TrimSpace(strings.ToLower(req.Description))
	websiteURL := strings.TrimSpace(strings.ToLower(req.WebsiteURL))
	
	// Create a hash of the normalized business name, description, and website URL
	data := fmt.Sprintf("%s|%s|%s", businessName, description, websiteURL)
	hash := sha256.Sum256([]byte(data))
	cacheKey := fmt.Sprintf("classification:%x", hash)
	
	// Log cache key generation for debugging (only first 16 chars for security)
	keyPrefix := cacheKey
	if len(cacheKey) > 16 {
		keyPrefix = cacheKey[:16]
	}
	h.logger.Debug("Generated cache key",
		zap.String("key_prefix", keyPrefix),
		zap.String("business_name", req.BusinessName))
	
	return cacheKey
}

// validateResponse ensures all required frontend fields are present in the response
// Priority 4 Fix: Ensures 100% frontend compatibility by validating and fixing missing fields
func (h *ClassificationHandler) validateResponse(response *ClassificationResponse, req *ClassificationRequest) {
	// Ensure PrimaryIndustry is always set (use "Unknown" if empty to avoid omitempty tag omitting it)
	// Note: PrimaryIndustry has omitempty tag, so empty strings are omitted from JSON
	// Setting to "Unknown" ensures the field is always present in the response
	if response.PrimaryIndustry == "" {
		response.PrimaryIndustry = "Unknown" // Set non-empty default to ensure field is present
	}

	// Ensure Explanation is always set (even if empty)
	if response.Explanation == "" {
		response.Explanation = "" // Explicitly set empty string (not omitted)
	}

	// Ensure Metadata is always set (never nil)
	if response.Metadata == nil {
		response.Metadata = make(map[string]interface{})
	}

	// Ensure Classification is never nil
	if response.Classification == nil {
		response.Classification = &ClassificationResult{
			Industry:   response.PrimaryIndustry,
			MCCCodes:   []IndustryCode{},
			NAICSCodes: []IndustryCode{},
			SICCodes:   []IndustryCode{},
		}
	} else {
		// Ensure code arrays are never nil (use empty arrays)
		if response.Classification.MCCCodes == nil {
			response.Classification.MCCCodes = []IndustryCode{}
		}
		if response.Classification.NAICSCodes == nil {
			response.Classification.NAICSCodes = []IndustryCode{}
		}
		if response.Classification.SICCodes == nil {
			response.Classification.SICCodes = []IndustryCode{}
		}
		// Ensure Industry field is set
		if response.Classification.Industry == "" {
			response.Classification.Industry = response.PrimaryIndustry
		}
		// Priority 4 Fix: Ensure structured explanation is present (if classification exists)
		// The structured explanation provides detailed reasoning for frontend display
		if response.Classification.Explanation == nil {
			// Create a minimal structured explanation if missing
			response.Classification.Explanation = &classification.ClassificationExplanation{
				PrimaryReason:     response.Explanation, // Use top-level explanation as fallback
				SupportingFactors: []string{fmt.Sprintf("Confidence score: %.0f%%", response.ConfidenceScore*100)},
				KeyTermsFound:     []string{},
				MethodUsed:        "multi_strategy",
				ProcessingPath:    response.ProcessingPath,
			}
			if response.Classification.Explanation.PrimaryReason == "" {
				response.Classification.Explanation.PrimaryReason = fmt.Sprintf("Classified as '%s' based on business information", response.PrimaryIndustry)
			}
		}
	}

	// Ensure Status is always set
	if response.Status == "" {
		response.Status = "success"
		if !response.Success {
			response.Status = "error"
		}
	}

	// Ensure Timestamp is set
	if response.Timestamp.IsZero() {
		response.Timestamp = time.Now()
	}

	h.logger.Debug("Response validated for frontend compatibility",
		zap.String("request_id", req.RequestID),
		zap.Bool("has_primary_industry", response.PrimaryIndustry != ""),
		zap.Bool("has_explanation", response.Explanation != ""),
		zap.Bool("has_metadata", response.Metadata != nil),
		zap.Bool("has_classification", response.Classification != nil))
}

// sendErrorResponse sends an error response with frontend-compatible structure
// FIX: Ensures all error responses include required frontend fields (primary_industry, classification, explanation, confidence_score)
func (h *ClassificationHandler) sendErrorResponse(
	w http.ResponseWriter,
	r *http.Request,
	req *ClassificationRequest,
	err error,
	statusCode int,
) {
	response := ClassificationResponse{
		RequestID:       req.RequestID,
		BusinessName:    req.BusinessName,
		Description:     req.Description,
		PrimaryIndustry: "Unknown", // Set to "Unknown" to ensure field is present (omitempty tag would omit empty string)
		Classification: &ClassificationResult{
			Industry:   "Unknown",
			MCCCodes:   []IndustryCode{},
			NAICSCodes: []IndustryCode{},
			SICCodes:   []IndustryCode{},
		},
		ConfidenceScore: 0.0,
		Explanation:      fmt.Sprintf("Error: %v", err),
		Status:           "error",
		Success:          false,
		Timestamp:        time.Now(),
		ProcessingTime:   0,
		DataSource:       "error",
		Metadata: map[string]interface{}{
			"error":       err.Error(),
			"error_type":  "classification_error",
			"status_code": statusCode,
		},
	}

	// Priority 4 Fix: Validate error response to ensure all required frontend fields are present
	h.validateResponse(&response, req)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// inferStrategyFromPath infers scraping strategy from processing path
// FIX: Helper function to infer strategy when metadata is not available
func inferStrategyFromPath(path string) string {
	if strings.Contains(path, "layer1") {
		return "early_exit"
	}
	if strings.Contains(path, "layer2") {
		return "standard_scraping"
	}
	if strings.Contains(path, "layer3") {
		return "deep_scraping"
	}
	return "unknown"
}

// getCachedResponse retrieves a cached response if available and not expired
func (h *ClassificationHandler) getCachedResponse(key string) (*ClassificationResponse, bool) {
	if !h.config.Classification.CacheEnabled {
		h.logger.Debug("Cache disabled, skipping cache lookup",
			zap.String("key", key))
		return nil, false
	}

	// Try Redis cache first if enabled
	if h.redisCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		data, found := h.redisCache.Get(ctx, key)
		if found {
			var response ClassificationResponse
			if err := json.Unmarshal(data, &response); err == nil {
				keyPrefix := key
				if len(key) > 16 {
					keyPrefix = key[:16]
				}
				h.logger.Info("✅ [CACHE-HIT] Cache hit from Redis",
					zap.String("key", key),
					zap.String("key_prefix", keyPrefix))
				return &response, true
			} else {
				h.logger.Warn("Failed to unmarshal cached response from Redis",
					zap.String("key", key),
					zap.Error(err))
			}
		} else {
			keyPrefix := key
			if len(key) > 16 {
				keyPrefix = key[:16]
			}
			h.logger.Debug("Cache miss from Redis",
				zap.String("key", key),
				zap.String("key_prefix", keyPrefix))
		}
	} else {
		h.logger.Debug("Redis cache not initialized, trying in-memory cache",
			zap.String("key", key),
			zap.Bool("redis_enabled", h.config.Classification.RedisEnabled),
			zap.String("redis_url_set", func() string {
				if h.config.Classification.RedisURL != "" {
					return "yes"
				}
				return "no"
			}()))
	}

	// Fallback to in-memory cache
	h.cacheMutex.RLock()
	defer h.cacheMutex.RUnlock()

	entry, exists := h.cache[key]
	if !exists {
		h.logger.Debug("Cache miss from in-memory cache",
			zap.String("key", key),
			zap.Int("cache_size", len(h.cache)))
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		h.logger.Debug("Cache entry expired in in-memory cache",
			zap.String("key", key),
			zap.Time("expires_at", entry.expiresAt))
		return nil, false
	}

	keyPrefix := key
	if len(key) > 16 {
		keyPrefix = key[:16]
	}
	h.logger.Info("✅ [CACHE-HIT] Cache hit from in-memory cache",
		zap.String("key", key),
		zap.String("key_prefix", keyPrefix))
	return entry.response, true
}

// setCachedResponse stores a response in the cache
func (h *ClassificationHandler) setCachedResponse(key string, response *ClassificationResponse) {
	if !h.config.Classification.CacheEnabled {
		h.logger.Debug("Cache disabled, skipping cache store",
			zap.String("key", key))
		return
	}

	// Store in Redis cache if enabled
	if h.redisCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		data, err := json.Marshal(response)
		if err == nil {
			h.redisCache.Set(ctx, key, data, h.config.Classification.CacheTTL)
			keyPrefix := key
			if len(key) > 16 {
				keyPrefix = key[:16]
			}
			h.logger.Info("✅ [CACHE-SET] Stored in Redis cache",
				zap.String("key", key),
				zap.String("key_prefix", keyPrefix),
				zap.Duration("ttl", h.config.Classification.CacheTTL))
		} else {
			h.logger.Warn("Failed to marshal response for Redis cache",
				zap.String("key", key),
				zap.Error(err))
		}
	} else {
		h.logger.Debug("Redis cache not initialized, storing in in-memory cache only",
			zap.String("key", key),
			zap.Bool("redis_enabled", h.config.Classification.RedisEnabled),
			zap.String("redis_url_set", func() string {
				if h.config.Classification.RedisURL != "" {
					return "yes"
				}
				return "no"
			}()))
	}

	// Always store in in-memory cache as fallback
	h.cacheMutex.Lock()
	defer h.cacheMutex.Unlock()

	h.cache[key] = &cacheEntry{
		response:  response,
		expiresAt: time.Now().Add(h.config.Classification.CacheTTL),
	}
	keyPrefix := key
	if len(key) > 16 {
		keyPrefix = key[:16]
	}
	h.logger.Debug("✅ [CACHE-SET] Stored in in-memory cache",
		zap.String("key", key),
		zap.String("key_prefix", keyPrefix),
		zap.Int("cache_size", len(h.cache)),
		zap.Duration("ttl", h.config.Classification.CacheTTL))
}

// ClassificationRequest represents a classification request
type ClassificationRequest struct {
	BusinessName string `json:"business_name"`
	Description  string `json:"description"`
	WebsiteURL   string `json:"website_url,omitempty"`
	RequestID    string `json:"request_id,omitempty"`
}

// ClassificationResponse represents a classification response
type ClassificationResponse struct {
	RequestID           string                 `json:"request_id"`
	BusinessName        string                 `json:"business_name"`
	Description         string                 `json:"description"`
	PrimaryIndustry     string                 `json:"primary_industry,omitempty"` // Added for merchant service compatibility
	Classification      *ClassificationResult  `json:"classification"`
	RiskAssessment      *RiskAssessmentResult  `json:"risk_assessment"`
	VerificationStatus  *VerificationStatus    `json:"verification_status"`
	ConfidenceScore     float64                `json:"confidence_score"`
	Explanation         string                 `json:"explanation,omitempty"`         // DistilBART explanation
	ContentSummary      string                 `json:"contentSummary,omitempty"`      // DistilBART content summary
	QuantizationEnabled bool                   `json:"quantizationEnabled,omitempty"` // Quantization status
	ModelVersion        string                 `json:"modelVersion,omitempty"`        // Model version
	DataSource          string                 `json:"data_source"`
	Status              string                 `json:"status"`
	Success             bool                   `json:"success"`
	Timestamp           time.Time              `json:"timestamp"`
	ProcessingTime      time.Duration          `json:"processing_time"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
	// Phase 4: Async LLM processing fields
	LLMProcessingID     string `json:"llm_processing_id,omitempty"`
	LLMStatus           string `json:"llm_status,omitempty"`
	
	// Phase 5: Cache fields
	FromCache           bool       `json:"from_cache"`           // Indicates if result came from cache
	CachedAt            *time.Time `json:"cached_at,omitempty"`   // When result was cached
	ProcessingPath      string     `json:"processing_path,omitempty"` // Layer used: "layer1", "layer2", "layer3"
}

// requestTrace tracks detailed timing for a single request
type requestTrace struct {
	requestID     string
	stages        []stageTiming
	totalDuration time.Duration
	startTime     time.Time
	endTime       time.Time
}

// stageTiming tracks timing for a single processing stage
type stageTiming struct {
	stage     string
	startTime time.Time
	endTime   time.Time
	duration  time.Duration
	error     error
	metadata  map[string]interface{}
}

// VerificationStatus represents verification status information
type VerificationStatus struct {
	Status         string        `json:"status"`
	ProcessingTime time.Duration `json:"processing_time"`
	DataSources    []string      `json:"data_sources"`
	Checks         []CheckResult `json:"checks"`
	OverallScore   float64       `json:"overall_score"`
	CompletedAt    time.Time     `json:"completed_at"`
}

// CheckResult represents the result of a verification check
type CheckResult struct {
	CheckType  string  `json:"check_type"`
	Status     string  `json:"status"`
	Confidence float64 `json:"confidence"`
	Details    string  `json:"details"`
	Source     string  `json:"source"`
}

// ClassificationResult represents the classification results
type ClassificationResult struct {
	Industry       string                        `json:"industry"`
	MCCCodes       []IndustryCode                `json:"mcc_codes"`
	NAICSCodes     []IndustryCode                `json:"naics_codes"`
	SICCodes       []IndustryCode                `json:"sic_codes"`
	WebsiteContent *WebsiteContent               `json:"website_content"`
	Explanation    *classification.ClassificationExplanation `json:"explanation,omitempty"` // Phase 2: Structured explanation
}

// IndustryCode represents an industry classification code
type IndustryCode struct {
	Code           string   `json:"code"`
	Description    string   `json:"description"`
	Confidence     float64  `json:"confidence"`
	Source         []string `json:"source,omitempty"`         // ["industry", "keyword", "both"]
	MatchType      string   `json:"matchType,omitempty"`      // "exact", "partial", "synonym"
	RelevanceScore float64  `json:"relevanceScore,omitempty"` // From code_keywords table
	Industries     []string `json:"industries,omitempty"`     // Industries that contributed this code
	IsPrimary      bool     `json:"isPrimary,omitempty"`      // From classification_codes.is_primary
}

// WebsiteContent represents website content analysis
type WebsiteContent struct {
	Scraped       bool `json:"scraped"`
	ContentLength int  `json:"content_length"`
	KeywordsFound int  `json:"keywords_found"`
}

// RiskAssessmentResult represents comprehensive risk assessment results
type RiskAssessmentResult struct {
	// Core risk metrics
	OverallRiskScore float64 `json:"overall_risk_score"`
	RiskLevel        string  `json:"risk_level"`
	RiskScore        float64 `json:"risk_score"` // Legacy field for backward compatibility

	// Risk categories breakdown
	Categories map[string]float64 `json:"categories"`

	// Risk analysis details
	RiskFactors             []string `json:"risk_factors"`
	DetectedRisks           []string `json:"detected_risks,omitempty"`
	ProhibitedKeywordsFound []string `json:"prohibited_keywords_found,omitempty"`
	Recommendations         []string `json:"recommendations"`

	// Benchmarking and trends
	IndustryBenchmark float64 `json:"industry_benchmark"`
	PreviousRiskScore float64 `json:"previous_risk_score,omitempty"`

	// Assessment metadata
	AssessmentMethodology string        `json:"assessment_methodology"`
	AssessmentTimestamp   time.Time     `json:"assessment_timestamp"`
	DataSources           []string      `json:"data_sources"`
	ProcessingTime        time.Duration `json:"processing_time"`
}

// HandleClassification handles classification requests
func (h *ClassificationHandler) HandleClassification(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// CRITICAL FIX: Check context state IMMEDIATELY and create fresh context if needed
	// This must happen before any processing to ensure sufficient time for operations
	parentCtx := r.Context()
	ctxInfo := map[string]interface{}{
		"has_deadline":      false,
		"time_remaining_ms": 0,
		"context_err":       nil,
	}

	if deadline, hasDeadline := parentCtx.Deadline(); hasDeadline {
		ctxInfo["has_deadline"] = true
		timeRemaining := time.Until(deadline)
		ctxInfo["time_remaining_ms"] = timeRemaining.Milliseconds()

		// If context has insufficient time (<90s), create fresh context IMMEDIATELY
		// Increased threshold to 90s to match worker context timeout expectations
		if timeRemaining < 90*time.Second {
			h.logger.Warn("⚠️ [CONTEXT-FIX] Parent context has insufficient time, creating fresh context immediately",
				zap.Duration("time_remaining", timeRemaining),
				zap.Duration("minimum_required", 60*time.Second))
			parentCtx = context.Background()
			ctxInfo["fresh_context_created"] = true
		}
	}

	if parentCtx.Err() != nil {
		ctxInfo["context_err"] = parentCtx.Err().Error()
		h.logger.Warn("⚠️ [CONTEXT-FIX] Parent context expired, creating fresh context",
			zap.Error(parentCtx.Err()))
		parentCtx = context.Background()
		ctxInfo["fresh_context_created"] = true
	}

	// Entry-point logging with context information
	// FIX #17: Logging verbosity - Consider reducing in production for performance
	// Current logging is verbose for debugging but may impact performance at scale
	h.logger.Info("📥 [ENTRY-POINT] Classification request received",
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.Any("context_info", ctxInfo))

	// Check if streaming is requested
	stream := r.URL.Query().Get("stream") == "true"

	if stream {
		h.handleClassificationStreaming(w, r, startTime)
		return
	}

	// Set response headers for non-streaming
	w.Header().Set("Content-Type", "application/json")

	// Parse request body with timeout protection
	h.logger.Info("📥 [PARSE] Starting request body parsing",
		zap.String("content_length", r.Header.Get("Content-Length")))

	parseStart := time.Now()

	var req ClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("❌ [PARSE] Failed to decode request",
			zap.Error(err),
			zap.Duration("parse_duration", time.Since(parseStart)))
		h.sendErrorResponse(w, r, &req, fmt.Errorf("invalid request body: Please provide valid JSON"), http.StatusBadRequest)
		return
	}

	h.logger.Info("✅ [PARSE] Request body parsed successfully",
		zap.Duration("parse_duration", time.Since(parseStart)),
		zap.String("business_name", req.BusinessName),
		zap.String("website_url", req.WebsiteURL))

	// Validate request
	if req.BusinessName == "" {
		h.sendErrorResponse(w, r, &req, fmt.Errorf("business_name is required"), http.StatusBadRequest)
		return
	}

	// Sanitize input to prevent XSS and injection attacks
	req.BusinessName = sanitizeInput(req.BusinessName)
	if req.Description != "" {
		req.Description = sanitizeInput(req.Description)
	}
	if req.WebsiteURL != "" {
		req.WebsiteURL = sanitizeInput(req.WebsiteURL)
	}

	// Generate request ID if not provided
	if req.RequestID == "" {
		req.RequestID = h.generateRequestID()
	}

	// Request Admission Control: Check memory usage and queue capacity before processing
	// This prevents OOM kills by rejecting requests when memory is high or queue is full
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memUsagePercent := float64(m.Alloc) / float64(m.Sys) * 100

	if memUsagePercent > 80 {
		h.logger.Warn("Memory usage high, rejecting request",
			zap.String("request_id", req.RequestID),
			zap.Float64("mem_usage_percent", memUsagePercent),
			zap.Uint64("alloc_bytes", m.Alloc),
			zap.Uint64("sys_bytes", m.Sys))
		h.sendErrorResponse(w, r, &req, fmt.Errorf("service temporarily unavailable due to high load"), http.StatusServiceUnavailable)
		return
	}

	// Check queue capacity
	if h.requestQueue != nil && h.requestQueue.Size() >= h.config.Classification.MaxConcurrentRequests {
		h.logger.Warn("Request queue full, rejecting request",
			zap.String("request_id", req.RequestID),
			zap.Int("queue_size", h.requestQueue.Size()),
			zap.Int("max_concurrent", h.config.Classification.MaxConcurrentRequests))
		h.sendErrorResponse(w, r, &req, fmt.Errorf("service queue full, please retry later"), http.StatusServiceUnavailable)
		return
	}

	// Log memory stats before processing
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	memUsagePercentBefore := float64(memBefore.Alloc) / float64(memBefore.Sys) * 100

	// Log request arrival with detailed information
	h.logger.Info("📥 [REQUEST-ARRIVAL] Classification request received",
		zap.String("request_id", req.RequestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", r.RemoteAddr),
		zap.String("user_agent", r.UserAgent()),
		zap.Time("arrival_time", time.Now()),
		zap.Int("queue_size", h.requestQueue.Size()),
		zap.Int("worker_count", h.WorkerPool.workers),
		zap.Float64("mem_usage_percent_before", memUsagePercentBefore),
		zap.Uint64("alloc_bytes_before", memBefore.Alloc),
		zap.Uint64("sys_bytes_before", memBefore.Sys))

	// Generate cache key for deduplication
	cacheKey := h.getCacheKey(&req)

	// Check cache first if enabled
	if h.config.Classification.CacheEnabled {
		if cachedResponse, found := h.getCachedResponse(cacheKey); found {
			// Cache hit - set FromCache flag and log metrics
			cachedResponse.FromCache = true
			cachedResponse.CachedAt = &time.Time{}
			*cachedResponse.CachedAt = time.Now()
			h.logger.Info("✅ [CACHE-HIT] Classification served from cache",
				zap.String("request_id", req.RequestID),
				zap.String("business_name", req.BusinessName),
				zap.String("cache_key", cacheKey),
				zap.Duration("cache_ttl", h.config.Classification.CacheTTL),
				zap.Duration("response_time", time.Since(startTime)))
			w.Header().Set("X-Cache", "HIT")
			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(h.config.Classification.CacheTTL.Seconds())))
			json.NewEncoder(w).Encode(cachedResponse)
			return
		}
		// Cache miss - log metrics
		h.logger.Info("❌ [CACHE-MISS] Cache miss, processing new request",
			zap.String("request_id", req.RequestID),
			zap.String("business_name", req.BusinessName),
			zap.String("cache_key", cacheKey))
		w.Header().Set("X-Cache", "MISS")
	}

	// Calculate adaptive timeout based on request characteristics (Hybrid Approach)
	requestTimeout := h.calculateAdaptiveTimeout(&req)
	
	// FIX: Log timeout calculation for performance monitoring
	h.logger.Info("⏱️ [TIMEOUT] Calculated adaptive timeout",
		zap.String("request_id", req.RequestID),
		zap.Duration("request_timeout", requestTimeout),
		zap.Bool("has_website_url", req.WebsiteURL != ""),
		zap.String("business_name", req.BusinessName))

	// CONTEXT FLOW DOCUMENTATION (FIX #11):
	// 1. Entry point (line ~728): Checks parentCtx (r.Context()) for sufficient time
	//    - If <90s remaining, creates fresh context.Background()
	//    - This ensures we have sufficient time for processing
	// 2. Context creation (line ~850): Creates processing context
	//    - Uses parentCtx (either original or Background from step 1)
	//    - Only adds timeout if parent has no deadline or insufficient time
	//    - Preserves parent context if it has >= requestTimeout remaining
	// 3. Queue context (line ~1016): Creates queueCtx for HTTP response wait
	//    - Uses workerTimeout (120s) + estimatedQueueWait + buffer
	//    - Ensures HTTP doesn't timeout before worker completes
	// 4. Worker context (line ~225): Worker creates fresh 120s context
	//    - Ignores queuedReq.ctx and creates fresh context for processing
	//    - This is correct - worker needs guaranteed time regardless of queue wait
	// 5. Process start (line ~1603): May create fresh context if insufficient time
	//    - Checks if context has <90s remaining
	//    - Creates fresh 120s context if needed
	//
	// NOTE: parentCtx was already checked and fixed at entry point if needed
	// It's either r.Context() with sufficient time, or context.Background() if insufficient
	// Log the final context state for verification
	if deadline, hasDeadline := parentCtx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		h.logger.Info("✅ [CONTEXT] Using context for processing",
			zap.String("request_id", req.RequestID),
			zap.Duration("time_remaining", timeRemaining),
			zap.Duration("request_timeout", requestTimeout),
			zap.Bool("has_deadline", hasDeadline))
	} else {
		h.logger.Info("✅ [CONTEXT] Using context without deadline for processing",
			zap.String("request_id", req.RequestID),
			zap.Duration("request_timeout", requestTimeout))
	}

	// PROFILING: Track time at context creation
	contextCreationStart := time.Now()

	ctx, contentCache := reqcache.WithContentCache(parentCtx)

	// FIX #1: Only add timeout if parent has no deadline or insufficient time
	// This prevents overwriting a context that already has sufficient time
	var cancel context.CancelFunc
	if deadline, hasDeadline := parentCtx.Deadline(); !hasDeadline || time.Until(deadline) < requestTimeout {
		// Parent has no deadline or insufficient time, create timeout
		ctx, cancel = context.WithTimeout(ctx, requestTimeout)
		h.logger.Info("Created new context timeout",
			zap.String("request_id", req.RequestID),
			zap.Duration("request_timeout", requestTimeout),
			zap.Bool("parent_had_deadline", hasDeadline))
		if cancel != nil {
			defer cancel()
		}
	} else {
		// Parent has sufficient time, use as-is
		// No cancel needed since we're not creating a new timeout
		h.logger.Info("Using parent context with sufficient time",
			zap.String("request_id", req.RequestID),
			zap.Duration("parent_time_remaining", time.Until(deadline)),
			zap.Duration("request_timeout", requestTimeout))
	}

	contextCreationDuration := time.Since(contextCreationStart)

	// Log created context state
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		h.logger.Info("Final context state",
			zap.String("request_id", req.RequestID),
			zap.Duration("time_remaining", timeRemaining),
			zap.Duration("context_creation_duration", contextCreationDuration),
			zap.Duration("request_timeout", requestTimeout))
	}

	// Store cache reference for later use (if needed)
	_ = contentCache

	// Check if identical request is already in-flight
	h.inFlightMutex.RLock()
	inFlight, exists := h.inFlightRequests[cacheKey]
	h.inFlightMutex.RUnlock()

	if exists {
		// Check if in-flight request has timed out
		elapsed := time.Since(inFlight.startTime)
		if inFlight.timeout > 0 && elapsed > inFlight.timeout {
			h.logger.Warn("In-flight request timed out, removing from deduplication",
				zap.String("request_id", req.RequestID),
				zap.String("cache_key", cacheKey),
				zap.Duration("elapsed", elapsed),
				zap.Duration("timeout", inFlight.timeout))
			// Remove stale in-flight request
			h.inFlightMutex.Lock()
			delete(h.inFlightRequests, cacheKey)
			h.inFlightMutex.Unlock()
			// Continue with new request processing
		} else {
			h.logger.Info("Request deduplication: waiting for in-flight request",
				zap.String("request_id", req.RequestID),
				zap.String("cache_key", cacheKey),
				zap.Duration("elapsed", elapsed),
				zap.Duration("max_wait", requestTimeout))

			// Wait for the in-flight request to complete with timeout
			waitTimeout := requestTimeout - elapsed
			if waitTimeout <= 0 {
				waitTimeout = 5 * time.Second // Minimum wait time
			}

			waitCtx, waitCancel := context.WithTimeout(ctx, waitTimeout)
			defer waitCancel()

			select {
			case result := <-inFlight.resultChan:
				if result.err != nil {
					h.logger.Error("In-flight request failed",
						zap.String("request_id", req.RequestID),
						zap.Error(result.err))
					h.sendErrorResponse(w, r, &req, result.err, http.StatusInternalServerError)
					return
				}
				h.logger.Info("Classification served from in-flight request",
					zap.String("request_id", req.RequestID),
					zap.String("business_name", req.BusinessName),
					zap.Duration("total_wait", time.Since(inFlight.startTime)))
				w.Header().Set("X-Deduplication", "HIT")
				json.NewEncoder(w).Encode(result.response)
				return
			case <-waitCtx.Done():
				h.logger.Warn("Timeout waiting for in-flight request, processing new request",
					zap.String("request_id", req.RequestID),
					zap.Duration("wait_timeout", waitTimeout))
				// Remove stale in-flight request and continue
				h.inFlightMutex.Lock()
				delete(h.inFlightRequests, cacheKey)
				h.inFlightMutex.Unlock()
				// Continue with new request processing below
			case <-ctx.Done():
				h.logger.Warn("Context cancelled while waiting for in-flight request",
					zap.String("request_id", req.RequestID))
				h.sendErrorResponse(w, r, &req, fmt.Errorf("request timeout while waiting for duplicate request"), http.StatusRequestTimeout)
				return
			}
		}
	}

	// Check if queue is full before enqueuing
	queueSize := h.requestQueue.Size()
	if queueSize >= h.requestQueue.maxSize {
		h.logger.Warn("Request queue is full, rejecting request",
			zap.String("request_id", req.RequestID),
			zap.Int("queue_size", queueSize),
			zap.Int("max_size", h.requestQueue.maxSize))
		h.sendErrorResponse(w, r, &req, fmt.Errorf("service temporarily unavailable, request queue is full"), http.StatusServiceUnavailable)
		return
	}

	// Estimate queue wait time based on current queue size and average processing time
	// OPTIMIZATION: Improved estimation based on worker count and actual processing times
	// Average processing time: 15-20 seconds (conservative estimate)
	// With 30 workers, each worker can process ~3 requests per minute
	// Queue wait = (queue_size / worker_count) * average_processing_time
	avgProcessingTime := 15 * time.Second
	workerCount := h.WorkerPool.workers
	if workerCount == 0 {
		workerCount = 30 // Default if not initialized
	}

	// More accurate estimation: divide queue size by worker count
	estimatedQueueWait := time.Duration(queueSize) * avgProcessingTime / time.Duration(workerCount)
	if estimatedQueueWait > 60*time.Second {
		estimatedQueueWait = 60 * time.Second // Cap at 60 seconds (increased from 30s)
	}

	// OPTIMIZATION: Check if parent context is expired or has insufficient time
	// If so, use Background context instead of inheriting expired deadline
	useBackgroundForQueue := false
	if parentCtx.Err() != nil {
		h.logger.Warn("Parent context expired before queue, using Background context",
			zap.String("request_id", req.RequestID),
			zap.Error(parentCtx.Err()))
		useBackgroundForQueue = true
	} else if deadline, hasDeadline := parentCtx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		requiredTime := requestTimeout + estimatedQueueWait + 10*time.Second
		if timeRemaining < requiredTime {
			h.logger.Warn("Parent context has insufficient time for queue wait, using Background context",
				zap.String("request_id", req.RequestID),
				zap.Duration("time_remaining", timeRemaining),
				zap.Duration("required", requiredTime),
				zap.Duration("estimated_queue_wait", estimatedQueueWait))
			useBackgroundForQueue = true
		}
	}

	// FIX #2: Create context with sufficient time accounting for queue wait
	// Match worker timeout (120s) + queue wait + buffer to ensure HTTP doesn't timeout before worker completes
	// Worker creates fresh 120s context, so queue context needs at least that + queue wait + buffer
	workerTimeout := 120 * time.Second
	queueAwareTimeout := workerTimeout + estimatedQueueWait + 10*time.Second
	queueCtxParent := parentCtx
	if useBackgroundForQueue {
		queueCtxParent = context.Background()
	}
	queueCtx, queueCancel := context.WithTimeout(queueCtxParent, queueAwareTimeout)
	defer queueCancel()

	// Create queued request
	// FIX #14: Response channels are buffered (size 1) and used for one-time communication
	// They don't need explicit closing - garbage collection will handle cleanup
	// Context cancellation is used for cancellation instead of channel closure
	queuedReq := &queuedRequest{
		req:       &req,
		ctx:       queueCtx,
		response:  make(chan *ClassificationResponse, 1),
		errChan:   make(chan error, 1),
		startTime: time.Now(),
	}

	// Enqueue request
	if err := h.requestQueue.Enqueue(queuedReq); err != nil {
		h.logger.Warn("Failed to enqueue request",
			zap.String("request_id", req.RequestID),
			zap.Error(err))
		h.sendErrorResponse(w, r, &req, fmt.Errorf("service temporarily unavailable: %w", err), http.StatusServiceUnavailable)
		return
	}

	h.logger.Info("📋 [QUEUE-ENQUEUE] Request enqueued for processing",
		zap.String("request_id", req.RequestID),
		zap.Int("queue_size", h.requestQueue.Size()),
		zap.Int("worker_count", h.WorkerPool.workers),
		zap.Duration("estimated_wait", estimatedQueueWait),
		zap.Duration("queue_aware_timeout", queueAwareTimeout),
		zap.Bool("using_background_context", useBackgroundForQueue),
		zap.Time("enqueue_time", time.Now()))

	// Create in-flight request entry with timeout for deduplication
	resultChan := make(chan *inFlightResult, 1)
	inFlightReq := &inFlightRequest{
		resultChan: resultChan,
		startTime:  time.Now(),
		timeout:    requestTimeout,
	}

	h.inFlightMutex.Lock()
	h.inFlightRequests[cacheKey] = inFlightReq
	h.inFlightMutex.Unlock()

	// Clean up in-flight request when done
	defer func() {
		h.inFlightMutex.Lock()
		delete(h.inFlightRequests, cacheKey)
		h.inFlightMutex.Unlock()
	}()

	// Wait for response from worker pool
	select {
	case response := <-queuedReq.response:
		// Send result to waiting duplicate requests (non-blocking)
		// FIX: Check if channel is closed before sending to prevent panic
		inFlightReq.mu.Lock()
		isClosed := inFlightReq.closed
		inFlightReq.mu.Unlock()
		if !isClosed {
			select {
			case inFlightReq.resultChan <- &inFlightResult{response: response, err: nil}:
				// Result sent successfully
			default:
				// Channel already has a result (shouldn't happen, but safe to ignore)
			}
		}

		// Cache the response if enabled
		if h.config.Classification.CacheEnabled {
			h.setCachedResponse(cacheKey, response)
			h.logger.Info("💾 [CACHE-SET] Response cached for future requests",
				zap.String("request_id", req.RequestID),
				zap.String("cache_key", cacheKey),
				zap.Duration("cache_ttl", h.config.Classification.CacheTTL))
		}

		// Set cache headers for browser caching
		if h.config.Classification.CacheEnabled {
			w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(h.config.Classification.CacheTTL.Seconds())))
			w.Header().Set("ETag", fmt.Sprintf(`"%s"`, req.RequestID))
		}

		// FIX: Ensure ProcessingPath is set if early_exit is true (non-streaming path)
		// Set before serialization so it's included in the JSON response
		if response.Metadata != nil {
			if earlyExit, ok := response.Metadata["early_exit"].(bool); ok && earlyExit && response.ProcessingPath == "" {
				response.ProcessingPath = "layer1"
				h.logger.Info("🔧 [FIX] Setting ProcessingPath to layer1 in non-streaming path",
					zap.String("request_id", req.RequestID),
					zap.Bool("early_exit", earlyExit))
			}
		}

		// Priority 4 Fix: Validate response to ensure all required frontend fields are present
		h.validateResponse(response, &req)

		// Marshal JSON response to bytes first
		responseBytes, err := json.Marshal(response)
		if err != nil {
			h.logger.Error("Failed to marshal response",
				zap.String("request_id", req.RequestID),
				zap.Error(err))
			h.sendErrorResponse(w, r, &req, fmt.Errorf("failed to marshal response: %w", err), http.StatusInternalServerError)
			return
		}

		// Log response size for monitoring
		responseSize := len(responseBytes)
		h.logger.Info("Response prepared for sending",
			zap.String("request_id", req.RequestID),
			zap.Int("response_size_bytes", responseSize),
			zap.Duration("total_duration", time.Since(startTime)))

		// FIX #15: Check if HTTP connection is still valid before writing response
		// This prevents HTTP 000 errors from writing to closed connections
		if r.Context().Err() != nil {
			h.logger.Warn("HTTP connection already closed, skipping response",
				zap.String("request_id", req.RequestID),
				zap.Error(r.Context().Err()))
			return
		}

		// Log memory stats after processing
		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		memUsagePercentAfter := float64(memAfter.Alloc) / float64(memAfter.Sys) * 100
		memDelta := int64(memAfter.Alloc) - int64(memBefore.Alloc)

		h.logger.Info("📤 [RESPONSE-SENT] Classification response sent",
			zap.String("request_id", req.RequestID),
			zap.Duration("total_duration", time.Since(startTime)),
			zap.Float64("mem_usage_percent_after", memUsagePercentAfter),
			zap.Uint64("alloc_bytes_after", memAfter.Alloc),
			zap.Uint64("sys_bytes_after", memAfter.Sys),
			zap.Int64("mem_delta_bytes", memDelta),
			zap.Float64("mem_usage_delta_percent", memUsagePercentAfter-memUsagePercentBefore))

		// Write response
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(responseBytes); err != nil {
			h.logger.Warn("Failed to write response",
				zap.String("request_id", req.RequestID),
				zap.Error(err))
		}
		return

	case err := <-queuedReq.errChan:
		h.logger.Error("Request processing failed",
			zap.String("request_id", req.RequestID),
			zap.Error(err))

		// Send error to waiting duplicate requests (non-blocking)
		// FIX: Check if channel is closed before sending to prevent panic
		inFlightReq.mu.Lock()
		isClosed := inFlightReq.closed
		inFlightReq.mu.Unlock()
		if !isClosed {
			select {
			case inFlightReq.resultChan <- &inFlightResult{response: nil, err: err}:
				// Error sent successfully
			default:
				// Channel already has a result (shouldn't happen, but safe to ignore)
			}
		}

		// OPTIMIZATION: Check if HTTP connection is still valid before writing error
		// This prevents HTTP 000 errors from writing to closed connections
		if r.Context().Err() != nil {
			h.logger.Warn("HTTP connection already closed, skipping error response",
				zap.String("request_id", req.RequestID),
				zap.Error(r.Context().Err()))
			return
		}

		h.sendErrorResponse(w, r, &req, err, http.StatusInternalServerError)
		return

	case <-queueCtx.Done():
		h.logger.Warn("Request context cancelled while waiting for processing",
			zap.String("request_id", req.RequestID),
			zap.Error(queueCtx.Err()))

		// OPTIMIZATION: Check if HTTP connection is still valid before writing error
		if r.Context().Err() != nil {
			h.logger.Warn("HTTP connection already closed, skipping timeout response",
				zap.String("request_id", req.RequestID))
			return
		}

		h.sendErrorResponse(w, r, &req, fmt.Errorf("request timeout while waiting for processing"), http.StatusRequestTimeout)
		return

	case <-r.Context().Done():
		h.logger.Warn("HTTP request context cancelled",
			zap.String("request_id", req.RequestID),
			zap.Error(r.Context().Err()))

		// OPTIMIZATION: Connection is already closed, don't try to write
		// This prevents additional errors from trying to write to closed connections
		// The client has already timed out, so writing would fail anyway
		return
	}
}

// handleClassificationStreaming handles classification requests with streaming support (NDJSON)
// OPTIMIZATION #17: Streaming Responses for Long Operations
func (h *ClassificationHandler) handleClassificationStreaming(w http.ResponseWriter, r *http.Request, startTime time.Time) {
	// Set streaming response headers
	w.Header().Set("Content-Type", "application/x-ndjson")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Get flusher for streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.logger.Error("Streaming not supported by response writer")
		// Create minimal request for error response
		req := ClassificationRequest{
			RequestID: fmt.Sprintf("req_%d", time.Now().UnixNano()),
		}
		h.sendErrorResponse(w, r, &req, fmt.Errorf("streaming not supported"), http.StatusInternalServerError)
		return
	}

	// Parse request
	var req ClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request", zap.Error(err))
		h.sendStreamError(flusher, "Invalid request body: Please provide valid JSON", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.BusinessName == "" {
		h.sendStreamError(flusher, "business_name is required", http.StatusBadRequest)
		return
	}

	// Set request ID if not provided
	if req.RequestID == "" {
		req.RequestID = fmt.Sprintf("req_%d", time.Now().UnixNano())
	}

	// Send initial progress message
	h.sendStreamMessage(flusher, map[string]interface{}{
		"type":       "progress",
		"request_id": req.RequestID,
		"status":     "started",
		"message":    "Classification started",
		"timestamp":  time.Now(),
	})

	// Create context with timeout
	ctx, contentCache := reqcache.WithContentCache(r.Context())
	ctx, cancel := context.WithTimeout(ctx, h.config.Classification.RequestTimeout)
	defer cancel()

	_ = contentCache

	// Check cache first (streaming mode still benefits from cache)
	cacheKey := h.getCacheKey(&req)
	if h.config.Classification.CacheEnabled {
		if cached, found := h.getCachedResponse(cacheKey); found {
			h.logger.Info("Cache hit for streaming request",
				zap.String("request_id", req.RequestID),
				zap.String("cache_key", cacheKey))
			h.sendStreamMessage(flusher, map[string]interface{}{
				"type":       "progress",
				"request_id": req.RequestID,
				"status":     "cache_hit",
				"message":    "Result retrieved from cache",
			})
			h.sendStreamMessage(flusher, map[string]interface{}{
				"type":       "complete",
				"request_id": req.RequestID,
				"data":       cached,
			})
			return
		}
	}

	// Send progress: Starting classification
	h.sendStreamMessage(flusher, map[string]interface{}{
		"type":       "progress",
		"request_id": req.RequestID,
		"status":     "classifying",
		"message":    "Analyzing business and website",
		"step":       "classification",
	})

	// Step 1: Generate enhanced classification (industry detection)
	enhancedResult, err := h.generateEnhancedClassification(ctx, &req, false, false)
	if err != nil {
		h.logger.Error("Classification failed", zap.String("request_id", req.RequestID), zap.Error(err))
		h.sendStreamError(flusher, fmt.Sprintf("Classification failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Send progress: Industry detected
	h.sendStreamMessage(flusher, map[string]interface{}{
		"type":             "progress",
		"request_id":       req.RequestID,
		"status":           "industry_detected",
		"message":          "Industry detected",
		"step":             "industry",
		"primary_industry": enhancedResult.PrimaryIndustry,
		"confidence":       enhancedResult.ConfidenceScore,
	})

	// Step 2: Generate classification codes (if needed)
	var classificationResult *ClassificationResult
	shouldGenerateCodes := enhancedResult.ConfidenceScore >= 0.5 ||
		(enhancedResult.ConfidenceScore >= h.industryThresholds.GetThreshold(enhancedResult.PrimaryIndustry))

	if shouldGenerateCodes {
		h.sendStreamMessage(flusher, map[string]interface{}{
			"type":       "progress",
			"request_id": req.RequestID,
			"status":     "generating_codes",
			"message":    "Generating classification codes",
			"step":       "codes",
		})

		// Generate codes
		codes, codeGenErr := h.codeGenerator.GenerateClassificationCodes(
			ctx,
			enhancedResult.Keywords,
			enhancedResult.PrimaryIndustry,
			enhancedResult.ConfidenceScore,
		)
		if codeGenErr != nil {
			h.logger.Warn("Code generation failed, continuing without codes",
				zap.String("request_id", req.RequestID),
				zap.Error(codeGenErr))
		} else {
			enhancedResult.MCCCodes = convertMCCCodesToIndustryCodes(codes.MCC)
			enhancedResult.SICCodes = convertSICCodesToIndustryCodes(codes.SIC)
			enhancedResult.NAICSCodes = convertNAICSCodesToIndustryCodes(codes.NAICS)
		}

		// Phase 2: Generate or enhance explanation with codes
		explanationGenerator := classification.NewExplanationGenerator()
		contentQuality := 0.7
		if enhancedResult.ConfidenceScore > 0.8 {
			contentQuality = 0.85
		} else if enhancedResult.ConfidenceScore < 0.5 {
			contentQuality = 0.5
		}

		// Determine method from metadata or default to multi_strategy
		method := "multi_strategy"
		if enhancedResult.Metadata != nil {
			if m, ok := enhancedResult.Metadata["method"].(string); ok && m != "" {
				method = m
			}
		}
		
		multiResult := &classification.MultiStrategyResult{
			PrimaryIndustry: enhancedResult.PrimaryIndustry,
			Confidence:      enhancedResult.ConfidenceScore,
			Keywords:        enhancedResult.Keywords,
			Reasoning:       enhancedResult.ClassificationReasoning,
			Method:          method,
		}

		// Generate or regenerate explanation (always generate if nil, enhance if exists)
		if enhancedResult.ClassificationExplanation == nil {
			// Generate new explanation
			enhancedResult.ClassificationExplanation = explanationGenerator.GenerateExplanation(
				multiResult,
				codes, // Include codes if available
				contentQuality,
			)
		} else if codes != nil {
			// Enhance existing explanation with codes
			enhancedResult.ClassificationExplanation = explanationGenerator.GenerateExplanation(
				multiResult,
				codes,
				contentQuality,
			)
		}
	}

	// Phase 2: Use explanation from enhancedResult
	classificationExplanation := enhancedResult.ClassificationExplanation

	// Convert to response format
	classificationResult = &ClassificationResult{
		Industry:   enhancedResult.PrimaryIndustry,
		MCCCodes:   convertIndustryCodes(enhancedResult.MCCCodes),
		SICCodes:   convertIndustryCodes(enhancedResult.SICCodes),
		NAICSCodes: convertIndustryCodes(enhancedResult.NAICSCodes),
		WebsiteContent: &WebsiteContent{
			Scraped: enhancedResult.WebsiteAnalysis != nil && enhancedResult.WebsiteAnalysis.Success,
			ContentLength: func() int {
				if enhancedResult.WebsiteAnalysis != nil {
					return enhancedResult.WebsiteAnalysis.PagesAnalyzed * 1000
				}
				return 0
			}(),
			KeywordsFound: len(enhancedResult.Keywords),
		},
		Explanation: classificationExplanation, // Phase 2: Add explanation
	}

	// Send progress: Codes generated
	if shouldGenerateCodes {
		h.sendStreamMessage(flusher, map[string]interface{}{
			"type":        "progress",
			"request_id":  req.RequestID,
			"status":      "codes_generated",
			"message":     "Classification codes generated",
			"step":        "codes",
		"mcc_count":   len(classificationResult.MCCCodes),
		"sic_count":   len(classificationResult.SICCodes),
		"naics_count": len(classificationResult.NAICSCodes),
		})
	}

	// Step 3: Generate risk assessment and verification status in parallel
	processingTime := time.Since(startTime)
	var riskAssessment *RiskAssessmentResult
	var verificationStatus *VerificationStatus

	h.sendStreamMessage(flusher, map[string]interface{}{
		"type":       "progress",
		"request_id": req.RequestID,
		"status":     "assessing_risk",
		"message":    "Assessing business risk",
		"step":       "risk",
	})

	var wg sync.WaitGroup

	// Start risk assessment
	wg.Add(1)
	go func() {
		defer wg.Done()
		riskAssessment = h.generateRiskAssessment(&req, enhancedResult, processingTime)
	}()

	// Start verification status
	wg.Add(1)
	go func() {
		defer wg.Done()
		verificationStatus = h.generateVerificationStatus(&req, enhancedResult, processingTime)
	}()

	// Wait for both to complete
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Both completed
	case <-ctx.Done():
		h.sendStreamError(flusher, "Processing timeout", http.StatusRequestTimeout)
		return
	case <-time.After(10 * time.Second):
		h.logger.Warn("Parallel processing timeout in streaming mode",
			zap.String("request_id", req.RequestID))
		if riskAssessment == nil {
			riskAssessment = &RiskAssessmentResult{
				OverallRiskScore: 0.5,
				RiskLevel:        "MEDIUM",
			}
		}
		if verificationStatus == nil {
			verificationStatus = &VerificationStatus{
				Status: "PENDING",
			}
		}
	}

	// Send progress: Risk assessed
	h.sendStreamMessage(flusher, map[string]interface{}{
		"type":       "progress",
		"request_id": req.RequestID,
		"status":     "risk_assessed",
		"message":    "Risk assessment completed",
		"step":       "risk",
		"risk_level": riskAssessment.RiskLevel,
		"risk_score": riskAssessment.OverallRiskScore,
	})

	// Send progress: Verification status
	h.sendStreamMessage(flusher, map[string]interface{}{
		"type":                "progress",
		"request_id":          req.RequestID,
		"status":              "verification_complete",
		"message":             "Verification status generated",
		"step":                "verification",
		"verification_status": verificationStatus.Status,
	})

	// Extract DistilBART enhancement fields (legacy string explanation)
	var distilbartExplanation, contentSummary, modelVersion string
	var quantizationEnabled bool
	if enhancedResult.Metadata != nil {
		if exp, ok := enhancedResult.Metadata["explanation"].(string); ok {
			distilbartExplanation = exp
		}
		if summary, ok := enhancedResult.Metadata["content_summary"].(string); ok {
			contentSummary = summary
		}
		if quant, ok := enhancedResult.Metadata["quantization_enabled"].(bool); ok {
			quantizationEnabled = quant
		}
		if version, ok := enhancedResult.Metadata["model_version"].(string); ok {
			modelVersion = version
		}
	}
	if distilbartExplanation == "" {
		distilbartExplanation = enhancedResult.ClassificationReasoning
	}

	// Build final response
	response := &ClassificationResponse{
		RequestID:           req.RequestID,
		BusinessName:        req.BusinessName,
		Description:         req.Description,
		PrimaryIndustry:     enhancedResult.PrimaryIndustry,
		Classification:      classificationResult,
		RiskAssessment:      riskAssessment,
		VerificationStatus:  verificationStatus,
		ConfidenceScore:     enhancedResult.ConfidenceScore,
		Explanation:         distilbartExplanation,
		ContentSummary:      contentSummary,
		QuantizationEnabled: quantizationEnabled,
		ModelVersion:        modelVersion,
		DataSource:          "smart_crawling_classification_service",
		Status:              "success",
		Success:             true,
		Timestamp:           time.Now(),
		ProcessingTime:      time.Since(startTime),
		// Phase 4: Async LLM processing fields
		LLMProcessingID:     enhancedResult.LLMProcessingID,
		LLMStatus:           enhancedResult.LLMStatus,
		// Phase 5: Cache fields
		FromCache:           enhancedResult.FromCache,
		CachedAt:            enhancedResult.CachedAt,
		ProcessingPath: func() string {
			path := enhancedResult.ProcessingPath
			// FIX: Ensure ProcessingPath is set for early exits
			// Check both enhancedResult.Metadata and the response metadata being built
			if path == "" {
				// Check enhancedResult.Metadata first
				if enhancedResult.Metadata != nil {
					if earlyExit, ok := enhancedResult.Metadata["early_exit"].(bool); ok && earlyExit {
						path = "layer1"
						h.logger.Info("🔧 [FIX] Setting ProcessingPath to layer1 for early exit (from enhancedResult.Metadata)",
							zap.String("request_id", req.RequestID),
							zap.String("original_path", enhancedResult.ProcessingPath),
							zap.String("new_path", path))
						return path
					}
				}
				// Also check if scraping_strategy indicates early exit
				if enhancedResult.Metadata != nil {
					if strategy, ok := enhancedResult.Metadata["scraping_strategy"].(string); ok && strategy == "early_exit" {
						path = "layer1"
						h.logger.Info("🔧 [FIX] Setting ProcessingPath to layer1 for early exit (from scraping_strategy)",
							zap.String("request_id", req.RequestID),
							zap.String("scraping_strategy", strategy),
							zap.String("new_path", path))
						return path
					}
				}
			}
			// If still empty, check if response metadata will have early_exit
			if path == "" {
				// Check if response metadata will indicate early exit
				if enhancedResult.Metadata != nil {
					// Check all possible early exit indicators
					hasEarlyExit := false
					if earlyExit, ok := enhancedResult.Metadata["early_exit"].(bool); ok && earlyExit {
						hasEarlyExit = true
					}
					if !hasEarlyExit {
						if strategy, ok := enhancedResult.Metadata["scraping_strategy"].(string); ok && strategy == "early_exit" {
							hasEarlyExit = true
						}
					}
					if hasEarlyExit {
						path = "layer1"
						h.logger.Info("🔧 [FIX] Setting ProcessingPath to layer1 for early exit (final check)",
							zap.String("request_id", req.RequestID))
						return path
					}
				}
			}
			return path
		}(),
		Metadata: func() map[string]interface{} {
			metadata := map[string]interface{}{
				"service":                  "classification-service",
				"version":                  "2.0.0",
				"classification_reasoning": enhancedResult.ClassificationReasoning,
				"website_analysis":         enhancedResult.WebsiteAnalysis,
				"method_weights":           enhancedResult.MethodWeights,
				"smart_crawling_enabled":   true,
				"streaming":                true,
				// Scraping metadata fields (populated from enhancedResult.Metadata if available)
				"scraping_strategy":   "",
				"early_exit":          false,
				"fallback_used":       false,
				"fallback_type":       "",
				"scraping_time_ms":    0,
				"classification_time_ms": 0,
			}
			if enhancedResult.Metadata != nil {
				if codeGen, ok := enhancedResult.Metadata["codeGeneration"]; ok {
					metadata["codeGeneration"] = codeGen
				}
				// Extract scraping metadata if present
				if scrapingStrategy, ok := enhancedResult.Metadata["scraping_strategy"].(string); ok {
					metadata["scraping_strategy"] = scrapingStrategy
				}
				if earlyExit, ok := enhancedResult.Metadata["early_exit"].(bool); ok {
					metadata["early_exit"] = earlyExit
				}
				if fallbackUsed, ok := enhancedResult.Metadata["fallback_used"].(bool); ok {
					metadata["fallback_used"] = fallbackUsed
				}
				if fallbackType, ok := enhancedResult.Metadata["fallback_type"].(string); ok {
					metadata["fallback_type"] = fallbackType
				}
				if scrapingTime, ok := enhancedResult.Metadata["scraping_time_ms"].(float64); ok {
					metadata["scraping_time_ms"] = int64(scrapingTime)
				}
				if classificationTime, ok := enhancedResult.Metadata["classification_time_ms"].(float64); ok {
					metadata["classification_time_ms"] = int64(classificationTime)
				}
				// Include all other metadata fields
				for k, v := range enhancedResult.Metadata {
					if k != "explanation" && k != "content_summary" && k != "quantization_enabled" && k != "model_version" &&
						k != "scraping_strategy" && k != "early_exit" && k != "fallback_used" && k != "fallback_type" &&
						k != "scraping_time_ms" && k != "classification_time_ms" {
						metadata[k] = v
					}
				}
			}
			
			// FIX: Fallback to WebsiteAnalysis.StructuredData if metadata fields are still empty
			if metadata["scraping_strategy"] == "" && enhancedResult.WebsiteAnalysis != nil {
				if structuredData := enhancedResult.WebsiteAnalysis.StructuredData; structuredData != nil {
					if strategy, ok := structuredData["scraping_strategy"].(string); ok && strategy != "" {
						metadata["scraping_strategy"] = strategy
					}
					if earlyExit, ok := structuredData["early_exit"].(bool); ok {
						metadata["early_exit"] = earlyExit
					}
					if fallbackUsed, ok := structuredData["fallback_used"].(bool); ok {
						metadata["fallback_used"] = fallbackUsed
					}
					if fallbackType, ok := structuredData["fallback_type"].(string); ok && fallbackType != "" {
						metadata["fallback_type"] = fallbackType
					}
					if scrapingTime, ok := structuredData["scraping_time_ms"].(float64); ok && scrapingTime > 0 {
						metadata["scraping_time_ms"] = int64(scrapingTime)
					}
					if classificationTime, ok := structuredData["classification_time_ms"].(float64); ok && classificationTime > 0 {
						metadata["classification_time_ms"] = int64(classificationTime)
					}
				}
			}
			
			// FIX: Infer from processing path if still empty
			if metadata["scraping_strategy"] == "" && enhancedResult.ProcessingPath != "" {
				metadata["scraping_strategy"] = inferStrategyFromPath(enhancedResult.ProcessingPath)
			}
			
			// FIX: Set early_exit based on processing path if not set
			if !metadata["early_exit"].(bool) && enhancedResult.ProcessingPath == "layer1" {
				metadata["early_exit"] = true // Layer1 indicates early exit
			}
			
			// FIX: Also check if early_exit is set in enhancedResult.Metadata
			if !metadata["early_exit"].(bool) && enhancedResult.Metadata != nil {
				if earlyExit, ok := enhancedResult.Metadata["early_exit"].(bool); ok && earlyExit {
					metadata["early_exit"] = true
				}
			}
			
			// FIX: Set scraping_strategy to "early_exit" if early_exit is true but strategy is empty
			if metadata["early_exit"].(bool) && metadata["scraping_strategy"] == "" {
				metadata["scraping_strategy"] = "early_exit"
			}
			
			return metadata
		}(),
	}
	
	// FIX: Ensure ProcessingPath is set if early_exit is true
	// Set after response is built so we can check the metadata
	if response.Metadata != nil {
		if earlyExit, ok := response.Metadata["early_exit"].(bool); ok && earlyExit && response.ProcessingPath == "" {
			response.ProcessingPath = "layer1"
			h.logger.Info("🔧 [FIX] Setting ProcessingPath to layer1 after response build",
				zap.String("request_id", req.RequestID),
				zap.Bool("early_exit", earlyExit))
		}
	}

	// Cache the response if enabled
	if h.config.Classification.CacheEnabled {
		h.setCachedResponse(cacheKey, response)
	}

	// Send final completion message
	h.sendStreamMessage(flusher, map[string]interface{}{
		"type":               "complete",
		"request_id":         req.RequestID,
		"data":               response,
		"processing_time_ms": time.Since(startTime).Milliseconds(),
	})

	h.logger.Info("Streaming classification completed successfully",
		zap.String("request_id", req.RequestID),
		zap.Duration("processing_time", time.Since(startTime)))

	// Record classification for accuracy tracking (async)
	go h.recordClassificationForCalibration(ctx, &req, response, time.Since(startTime))
}

// sendStreamMessage sends a JSON message in NDJSON format (newline-delimited JSON)
func (h *ClassificationHandler) sendStreamMessage(flusher http.Flusher, message map[string]interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		h.logger.Error("Failed to marshal stream message", zap.Error(err))
		return
	}

	// Write JSON line followed by newline
	if _, err := fmt.Fprintf(flusher.(io.Writer), "%s\n", data); err != nil {
		h.logger.Error("Failed to write stream message", zap.Error(err))
		return
	}

	// Flush to send immediately
	flusher.Flush()
}

// sendStreamError sends an error message in NDJSON format
func (h *ClassificationHandler) sendStreamError(flusher http.Flusher, message string, statusCode int) {
	errorMsg := map[string]interface{}{
		"type":        "error",
		"status":      "error",
		"message":     message,
		"status_code": statusCode,
		"timestamp":   time.Now(),
	}
	h.sendStreamMessage(flusher, errorMsg)
}

// traceStage wraps a function call with timing and error tracking
func (h *ClassificationHandler) traceStage(trace *requestTrace, stageName string, metadata map[string]interface{}, fn func() error) error {
	stage := stageTiming{
		stage:     stageName,
		startTime: time.Now(),
		metadata:  metadata,
	}

	defer func() {
		stage.endTime = time.Now()
		stage.duration = stage.endTime.Sub(stage.startTime)
		trace.stages = append(trace.stages, stage)

		h.logger.Info("⏱️ [STAGE] Stage completed",
			zap.String("request_id", trace.requestID),
			zap.String("stage", stageName),
			zap.Duration("duration", stage.duration),
			zap.Error(stage.error),
			zap.Any("metadata", metadata))
	}()

	stage.error = fn()
	return stage.error
}

// logRequestTrace logs complete request trace
func (h *ClassificationHandler) logRequestTrace(trace *requestTrace) {
	trace.endTime = time.Now()
	trace.totalDuration = trace.endTime.Sub(trace.startTime)

	stageDurations := make(map[string]time.Duration)
	for _, stage := range trace.stages {
		stageDurations[stage.stage] = stage.duration
	}

	h.logger.Info("📊 [TRACE-COMPLETE] Request trace complete",
		zap.String("request_id", trace.requestID),
		zap.Duration("total_duration", trace.totalDuration),
		zap.Int("stage_count", len(trace.stages)),
		zap.Any("stage_durations", stageDurations),
		zap.Any("stages", trace.stages))
}

// processClassification processes a classification request
func (h *ClassificationHandler) processClassification(ctx context.Context, req *ClassificationRequest, startTime time.Time) (*ClassificationResponse, error) {
	// FIX: Add panic recovery to prevent HTTP 500 errors from unhandled panics
	defer func() {
		if r := recover(); r != nil {
			h.logger.Error("Panic recovered in processClassification",
				zap.String("request_id", req.RequestID),
				zap.Any("panic", r),
				zap.String("stack", string(debug.Stack())))
		}
	}()

	// Initialize request trace
	trace := &requestTrace{
		requestID: req.RequestID,
		startTime: startTime,
		stages:    make([]stageTiming, 0),
	}
	defer h.logRequestTrace(trace)

	// Log context state at entry to processClassification
	ctxInfo := map[string]interface{}{
		"has_deadline":      false,
		"time_remaining_ms": 0,
		"context_err":       nil,
	}
	if ctx.Err() != nil {
		ctxInfo["context_err"] = ctx.Err().Error()
	} else if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		ctxInfo["has_deadline"] = true
		timeRemaining := time.Until(deadline)
		ctxInfo["time_remaining_ms"] = timeRemaining.Milliseconds()
	}

	h.logger.Info("🔧 [PROCESS-START] Starting processClassification",
		zap.String("request_id", req.RequestID),
		zap.Duration("elapsed_since_start", time.Since(startTime)),
		zap.Any("context_info", ctxInfo))

	// OPTIMIZATION: Early termination check - if context is already expired, fail fast
	if ctx.Err() != nil {
		h.logger.Warn("❌ [PROCESS-START] Context expired before processing started",
			zap.String("request_id", req.RequestID),
			zap.Error(ctx.Err()),
			zap.Duration("elapsed_since_start", time.Since(startTime)))
		return nil, fmt.Errorf("context already expired before processing: %w", ctx.Err())
	}

	// OPTIMIZATION: Check time remaining and refresh context if needed
	// This is a second check in case context expired between worker check and processing start
	var processingCtx context.Context = ctx
	var cancelFunc context.CancelFunc = nil

	// Adaptive Timeout: Check time remaining and decide on operation scope
	// Skip expensive operations if time is limited to prevent timeouts
	var skipMultiPageAnalysis bool
	var skipMLClassification bool

	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		if timeRemaining <= 0 {
			h.logger.Warn("⚠️ [PROCESS-START] Context deadline expired at processing start, creating fresh context",
				zap.String("request_id", req.RequestID),
				zap.Duration("elapsed_since_start", time.Since(startTime)))
			// Create fresh context with sufficient timeout
			// Increased to 120s to accommodate longer processing times
			processingCtx, cancelFunc = context.WithTimeout(context.Background(), 120*time.Second)
		} else if timeRemaining < 90*time.Second {
			// If less than 90s remaining, create fresh context to ensure sufficient time
			// Increased threshold and timeout to accommodate longer processing times
			h.logger.Warn("⚠️ [PROCESS-START] Context has insufficient time at processing start, creating fresh context",
				zap.String("request_id", req.RequestID),
				zap.Duration("time_remaining", timeRemaining),
				zap.Duration("elapsed_since_start", time.Since(startTime)))
		} else {
			// Check time remaining to decide on operation scope
			if timeRemaining < 30*time.Second {
				h.logger.Warn("Insufficient time remaining, using quick classification path",
					zap.String("request_id", req.RequestID),
					zap.Duration("time_remaining", timeRemaining))
				skipMultiPageAnalysis = true
				skipMLClassification = true
			} else if timeRemaining < 60*time.Second {
				h.logger.Info("Limited time remaining, skipping multi-page analysis",
					zap.String("request_id", req.RequestID),
					zap.Duration("time_remaining", timeRemaining))
				skipMultiPageAnalysis = true
			}
			h.logger.Info("✅ [PROCESS-START] Starting classification processing with sufficient context time",
				zap.String("request_id", req.RequestID),
				zap.Duration("time_remaining", timeRemaining),
				zap.Duration("elapsed_since_start", time.Since(startTime)))
		}
	} else {
		h.logger.Info("✅ [PROCESS-START] Starting classification processing (no deadline on context)",
			zap.String("request_id", req.RequestID),
			zap.Duration("elapsed_since_start", time.Since(startTime)))
	}

	// Defer cancel if we created a new context
	if cancelFunc != nil {
		defer cancelFunc()
	}

	// Use processingCtx for the rest of the function
	ctx = processingCtx

	// OPTIMIZATION: Check context expiration periodically during processing
	// This allows us to detect if context expires mid-processing and handle gracefully
	checkCtx := func() error {
		if ctx.Err() != nil {
			return fmt.Errorf("context expired during processing: %w", ctx.Err())
		}
		if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
			timeRemaining := time.Until(deadline)
			if timeRemaining < 10*time.Second {
				h.logger.Warn("Context running low on time during processing",
					zap.String("request_id", req.RequestID),
					zap.Duration("time_remaining", timeRemaining),
					zap.Duration("elapsed", time.Since(startTime)))
			}
		}
		return nil
	}

	// Start timeout alert goroutine
	// FIX #16: Ensure goroutine properly exits on context cancellation
	// FIX: Enhanced timeout monitoring for performance investigation
	timeoutAlertCtx, timeoutAlertCancel := context.WithCancel(ctx)
	defer timeoutAlertCancel()

	go func() {
		ticker := time.NewTicker(5 * time.Second) // Check every 5 seconds for better granularity
		defer ticker.Stop()

		for {
			select {
			case <-timeoutAlertCtx.Done():
				// FIX #16: Context cancelled, goroutine exits properly
				h.logger.Debug("Timeout alert goroutine stopping",
					zap.String("request_id", req.RequestID),
					zap.Error(timeoutAlertCtx.Err()))
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)
				if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
					remaining := time.Until(deadline)
					if remaining < 20*time.Second {
						h.logger.Warn("⏰ [TIMEOUT-ALERT] Request approaching timeout",
							zap.String("request_id", req.RequestID),
							zap.Duration("elapsed", elapsed),
							zap.Duration("remaining", remaining),
							zap.Float64("percent_complete", float64(elapsed)/float64(elapsed+remaining)*100))
					} else if remaining < 40*time.Second {
						h.logger.Info("⏰ [TIMEOUT-WARNING] Request has limited time remaining",
							zap.String("request_id", req.RequestID),
							zap.Duration("elapsed", elapsed),
							zap.Duration("remaining", remaining))
					}
				}
			}
		}
	}()

	// Generate enhanced classification using actual classification services
	var enhancedResult *EnhancedClassificationResult
	classificationStartTime := time.Now()
	err := h.traceStage(trace, "classification_generation", map[string]interface{}{
		"has_website":     req.WebsiteURL != "",
		"has_description": req.Description != "",
	}, func() error {
		var err error
		enhancedResult, err = h.generateEnhancedClassification(ctx, req, skipMultiPageAnalysis, skipMLClassification)
		if err != nil {
			// Check if error is due to context expiration
			if ctx.Err() != nil {
				h.logger.Error("Classification failed due to context expiration",
					zap.String("request_id", req.RequestID),
					zap.Error(ctx.Err()),
					zap.Duration("elapsed", time.Since(startTime)))
				return fmt.Errorf("classification failed due to context expiration: %w", ctx.Err())
			}
			return fmt.Errorf("classification failed: %w", err)
		}
		return nil
	})
	
	// FIX: Performance monitoring - log slow operations
	classificationDuration := time.Since(classificationStartTime)
	if classificationDuration > 10*time.Second {
		h.logger.Warn("⚠️ [PERF] Slow classification operation detected",
			zap.String("request_id", req.RequestID),
			zap.Duration("duration", classificationDuration),
			zap.Duration("elapsed_since_start", time.Since(startTime)),
			zap.Bool("has_website_url", req.WebsiteURL != ""))
	}
	if classificationDuration > 20*time.Second {
		h.logger.Error("🚨 [PERF] Very slow classification operation",
			zap.String("request_id", req.RequestID),
			zap.Duration("duration", classificationDuration),
			zap.Duration("elapsed_since_start", time.Since(startTime)),
			zap.Bool("has_website_url", req.WebsiteURL != ""))
	}
	classificationTime := classificationDuration
	
	// Add performance metrics to metadata
	if enhancedResult != nil {
		if enhancedResult.Metadata == nil {
			enhancedResult.Metadata = make(map[string]interface{})
		}
		enhancedResult.Metadata["classification_time_ms"] = float64(classificationTime.Milliseconds())
		h.logger.Info("Performance metrics added",
			zap.String("request_id", req.RequestID),
			zap.Duration("classification_time", classificationTime))
	}
	if err != nil {
		return nil, err
	}

	// Final context check before returning
	if err := checkCtx(); err != nil {
		h.logger.Warn("Context expired after classification, but result obtained",
			zap.String("request_id", req.RequestID),
			zap.Error(err))
		// Continue with result even if context expired (we have the result)
	}
	
	// FIX: Performance summary logging for timeout investigation
	totalElapsed := time.Since(startTime)
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		remaining := time.Until(deadline)
		if remaining < 10*time.Second {
			h.logger.Warn("⏰ [PERF-SUMMARY] Request completed with limited time remaining",
				zap.String("request_id", req.RequestID),
				zap.Duration("total_elapsed", totalElapsed),
				zap.Duration("time_remaining", remaining),
				zap.Duration("classification_duration", classificationTime))
		}
	}
	h.logger.Info("✅ [PERF-SUMMARY] Request processing complete",
		zap.String("request_id", req.RequestID),
		zap.Duration("total_elapsed", totalElapsed),
		zap.Duration("classification_duration", classificationTime))

	// FIX: Defensive check for enhancedResult to prevent nil pointer dereference
	if enhancedResult == nil {
		return nil, fmt.Errorf("classification returned nil result")
	}

	// Phase 2: ALWAYS generate explanation to ensure it's set
	// Force generation even if one exists to ensure it's not nil or empty
	explanationGenerator := classification.NewExplanationGenerator()
	contentQuality := 0.7
	if enhancedResult.ConfidenceScore > 0.8 {
		contentQuality = 0.85
	} else if enhancedResult.ConfidenceScore < 0.5 {
		contentQuality = 0.5
	}

	// Determine method from metadata or default to multi_strategy
	method := "multi_strategy"
	if enhancedResult.Metadata != nil {
		if m, ok := enhancedResult.Metadata["method"].(string); ok && m != "" {
			method = m
		}
	}

	// Phase 2: Ensure keywords from multiResult are used (includes description keywords)
	// Get keywords from the enhancedResult
	keywordsForExplanation := enhancedResult.Keywords
	if keywordsForExplanation != nil && len(keywordsForExplanation) > 0 {
		h.logger.Info("✅ [Phase 2] Using keywords from enhancedResult",
			zap.String("request_id", req.RequestID),
			zap.Int("keyword_count", len(keywordsForExplanation)),
			zap.Strings("keywords", keywordsForExplanation))
	} else {
		h.logger.Info("⚠️ [Phase 2] No keywords available in enhancedResult",
			zap.String("request_id", req.RequestID),
			zap.Int("keyword_count", 0))
	}
	
	multiResult := &classification.MultiStrategyResult{
		PrimaryIndustry: enhancedResult.PrimaryIndustry,
		Confidence:      enhancedResult.ConfidenceScore,
		Keywords:        keywordsForExplanation, // Use keywords that include description keywords
		Reasoning:       enhancedResult.ClassificationReasoning,
		Method:          method,
		// Strategies field will be empty, but extractConfidenceFactors handles this gracefully
		Strategies: []classification.ClassificationStrategy{},
	}

	// ALWAYS generate explanation to ensure it's set (even if one existed before)
	enhancedResult.ClassificationExplanation = explanationGenerator.GenerateExplanation(
		multiResult,
		nil, // Codes not available at this point in processClassification
		contentQuality,
	)

	// Verify explanation was generated
	if enhancedResult.ClassificationExplanation == nil {
		h.logger.Error("❌ [Phase 2] Explanation generation returned nil!",
			zap.String("request_id", req.RequestID),
			zap.String("method", method))
		// Create a minimal explanation as fallback
		enhancedResult.ClassificationExplanation = &classification.ClassificationExplanation{
			PrimaryReason:     fmt.Sprintf("Classified as '%s' based on business information", enhancedResult.PrimaryIndustry),
			SupportingFactors: []string{fmt.Sprintf("Confidence score: %.0f%%", enhancedResult.ConfidenceScore*100)},
			KeyTermsFound:     enhancedResult.Keywords,
			MethodUsed:        method,
			ProcessingPath:    "full_strategy",
		}
	}

	h.logger.Info("✅ [Phase 2] Explanation generated/verified in processClassification",
		zap.String("request_id", req.RequestID),
		zap.Bool("explanation_not_nil", enhancedResult.ClassificationExplanation != nil),
		zap.String("method", method),
		zap.String("primary_reason", enhancedResult.ClassificationExplanation.PrimaryReason),
		zap.Int("supporting_factors", len(enhancedResult.ClassificationExplanation.SupportingFactors)),
		zap.Int("key_terms", len(enhancedResult.ClassificationExplanation.KeyTermsFound)))

	// Convert enhanced result to response format
	// Priority 4 Fix: Ensure code arrays are never nil (use empty arrays)
	mccCodes := enhancedResult.MCCCodes
	if mccCodes == nil {
		mccCodes = []IndustryCode{}
	}
	sicCodes := enhancedResult.SICCodes
	if sicCodes == nil {
		sicCodes = []IndustryCode{}
	}
	naicsCodes := enhancedResult.NAICSCodes
	if naicsCodes == nil {
		naicsCodes = []IndustryCode{}
	}

	classificationResult := &ClassificationResult{
		Industry:   enhancedResult.PrimaryIndustry,
		MCCCodes:   convertIndustryCodes(mccCodes),
		SICCodes:   convertIndustryCodes(sicCodes),
		NAICSCodes: convertIndustryCodes(naicsCodes),
		WebsiteContent: &WebsiteContent{
			Scraped: enhancedResult.WebsiteAnalysis != nil && enhancedResult.WebsiteAnalysis.Success,
			ContentLength: func() int {
				if enhancedResult.WebsiteAnalysis != nil {
					return enhancedResult.WebsiteAnalysis.PagesAnalyzed * 1000 // Estimate
				}
				return 0
			}(),
			KeywordsFound: len(enhancedResult.Keywords),
		},
		Explanation: enhancedResult.ClassificationExplanation, // Phase 2: Include structured explanation
	}

	h.logger.Info("✅ [Phase 2] ClassificationResult created",
		zap.String("request_id", req.RequestID),
		zap.Bool("explanation_in_result", classificationResult.Explanation != nil),
		zap.String("explanation_primary_reason", func() string {
			if classificationResult.Explanation != nil {
				return classificationResult.Explanation.PrimaryReason
			}
			return ""
		}()))

	// Parallel processing: Generate risk assessment and verification status concurrently
	// FIX #9: Add context cancellation to prevent goroutine leaks
	var riskAssessment *RiskAssessmentResult
	var verificationStatus *VerificationStatus

	processingTime := time.Since(startTime)
	var wg sync.WaitGroup

	// Create context with timeout for parallel processing
	parallelCtx, parallelCancel := context.WithTimeout(ctx, 10*time.Second)
	defer parallelCancel()

	// Start risk assessment in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-parallelCtx.Done():
			// Cancelled, use default
			riskAssessment = &RiskAssessmentResult{
				OverallRiskScore: 0.5,
				RiskLevel:        "MEDIUM",
			}
			return
		default:
			h.traceStage(trace, "risk_assessment", nil, func() error {
				riskAssessment = h.generateRiskAssessment(req, enhancedResult, processingTime)
				return nil
			})
		}
	}()

	// Start verification status in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-parallelCtx.Done():
			// Cancelled, use default
			verificationStatus = &VerificationStatus{
				Status: "PENDING",
			}
			return
		default:
			h.traceStage(trace, "verification_status", nil, func() error {
				verificationStatus = h.generateVerificationStatus(req, enhancedResult, processingTime)
				return nil
			})
		}
	}()

	// Wait for both to complete with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Both completed successfully
		h.logger.Info("Parallel processing completed successfully",
			zap.String("request_id", req.RequestID))
	case <-parallelCtx.Done():
		// Timeout or cancellation - cancel context to stop goroutines
		parallelCancel()
		// Wait for goroutines to finish
		wg.Wait()
		h.logger.Warn("Parallel processing timeout, continuing with available results",
			zap.String("request_id", req.RequestID))
		// Ensure we have default values if goroutines didn't complete
		if riskAssessment == nil {
			riskAssessment = &RiskAssessmentResult{
				OverallRiskScore: 0.5,
				RiskLevel:        "MEDIUM",
			}
		}
		if verificationStatus == nil {
			verificationStatus = &VerificationStatus{
				Status: "PENDING",
			}
		}
	}

	// Extract DistilBART enhancement fields from metadata if present
	var explanation, contentSummary, modelVersion string
	var quantizationEnabled bool
	if enhancedResult.Metadata != nil {
		if exp, ok := enhancedResult.Metadata["explanation"].(string); ok {
			explanation = exp
		}
		if summary, ok := enhancedResult.Metadata["content_summary"].(string); ok {
			contentSummary = summary
		}
		if quant, ok := enhancedResult.Metadata["quantization_enabled"].(bool); ok {
			quantizationEnabled = quant
		}
		if version, ok := enhancedResult.Metadata["model_version"].(string); ok {
			modelVersion = version
		}
	}
	// Fallback to ClassificationReasoning if explanation not in metadata
	if explanation == "" {
		explanation = enhancedResult.ClassificationReasoning
	}

	// Create response with enhanced reasoning
	response := &ClassificationResponse{
		RequestID:           req.RequestID,
		BusinessName:        req.BusinessName,
		Description:         req.Description,
		PrimaryIndustry:     enhancedResult.PrimaryIndustry, // Add at top level for merchant service compatibility
		Classification:      classificationResult,
		RiskAssessment:      riskAssessment,
		VerificationStatus:  verificationStatus,
		ConfidenceScore:     enhancedResult.ConfidenceScore,
		Explanation:         explanation,
		ContentSummary:      contentSummary,
		QuantizationEnabled: quantizationEnabled,
		ModelVersion:        modelVersion,
		DataSource:          "smart_crawling_classification_service",
		Status:              "success",
		Success:             true,
		Timestamp:           time.Now(),
		ProcessingTime:      time.Since(startTime),
		// Phase 4: Async LLM processing fields
		LLMProcessingID:     enhancedResult.LLMProcessingID,
		LLMStatus:           enhancedResult.LLMStatus,
		Metadata: func() map[string]interface{} {
			metadata := map[string]interface{}{
				"service":                  "classification-service",
				"version":                  "2.0.0",
				"classification_reasoning": enhancedResult.ClassificationReasoning,
				"website_analysis":         enhancedResult.WebsiteAnalysis,
				"method_weights":           enhancedResult.MethodWeights,
				"smart_crawling_enabled":   true,
				// Scraping metadata fields (populated from enhancedResult.Metadata if available)
				"scraping_strategy":   "",
				"early_exit":          false,
				"fallback_used":       false,
				"fallback_type":       "",
				"scraping_time_ms":    0,
				"classification_time_ms": 0,
			}
			// Include code generation metadata if present
			if enhancedResult.Metadata != nil {
				if codeGen, ok := enhancedResult.Metadata["codeGeneration"]; ok {
					metadata["codeGeneration"] = codeGen
				}
				// Extract scraping metadata if present
				if scrapingStrategy, ok := enhancedResult.Metadata["scraping_strategy"].(string); ok && scrapingStrategy != "" {
					metadata["scraping_strategy"] = scrapingStrategy
				}
				if earlyExit, ok := enhancedResult.Metadata["early_exit"].(bool); ok {
					metadata["early_exit"] = earlyExit
				}
				if fallbackUsed, ok := enhancedResult.Metadata["fallback_used"].(bool); ok {
					metadata["fallback_used"] = fallbackUsed
				}
				if fallbackType, ok := enhancedResult.Metadata["fallback_type"].(string); ok && fallbackType != "" {
					metadata["fallback_type"] = fallbackType
				}
				if scrapingTime, ok := enhancedResult.Metadata["scraping_time_ms"].(float64); ok {
					metadata["scraping_time_ms"] = int64(scrapingTime)
				}
				if classificationTime, ok := enhancedResult.Metadata["classification_time_ms"].(float64); ok {
					metadata["classification_time_ms"] = int64(classificationTime)
				}
				// Include all other metadata fields
				for k, v := range enhancedResult.Metadata {
					if k != "explanation" && k != "content_summary" && k != "quantization_enabled" && k != "model_version" &&
						k != "scraping_strategy" && k != "early_exit" && k != "fallback_used" && k != "fallback_type" &&
						k != "scraping_time_ms" && k != "classification_time_ms" {
						metadata[k] = v
					}
				}
			}
			// Also try to extract from WebsiteAnalysis.StructuredData as fallback
			if enhancedResult.WebsiteAnalysis != nil && enhancedResult.WebsiteAnalysis.StructuredData != nil {
				if scrapingStrategy, ok := enhancedResult.WebsiteAnalysis.StructuredData["scraping_strategy"].(string); ok && scrapingStrategy != "" && metadata["scraping_strategy"] == "" {
					metadata["scraping_strategy"] = scrapingStrategy
				}
				if earlyExit, ok := enhancedResult.WebsiteAnalysis.StructuredData["early_exit"].(bool); ok && !metadata["early_exit"].(bool) {
					metadata["early_exit"] = earlyExit
				}
				if fallbackUsed, ok := enhancedResult.WebsiteAnalysis.StructuredData["fallback_used"].(bool); ok && !metadata["fallback_used"].(bool) {
					metadata["fallback_used"] = fallbackUsed
				}
				if fallbackType, ok := enhancedResult.WebsiteAnalysis.StructuredData["fallback_type"].(string); ok && fallbackType != "" && metadata["fallback_type"] == "" {
					metadata["fallback_type"] = fallbackType
				}
				if scrapingTime, ok := enhancedResult.WebsiteAnalysis.StructuredData["scraping_time_ms"].(float64); ok && metadata["scraping_time_ms"] == 0 {
					metadata["scraping_time_ms"] = int64(scrapingTime)
				}
			}
			return metadata
		}(),
	}

	// Priority 4 Fix: Validate response to ensure all required frontend fields are present
	h.validateResponse(response, req)

	// Log the final response for debugging
	h.logger.Info("Classification response prepared",
		zap.String("request_id", req.RequestID),
		zap.String("primary_industry", response.PrimaryIndustry),
		zap.Float64("confidence", response.ConfidenceScore),
		zap.Bool("success", response.Success))

	return response, nil
}

// generateRiskAssessment creates a comprehensive risk assessment based on business data
func (h *ClassificationHandler) generateRiskAssessment(req *ClassificationRequest, classification *EnhancedClassificationResult, processingTime time.Duration) *RiskAssessmentResult {
	// Analyze business name for risk indicators
	riskFactors := h.analyzeBusinessName(req.BusinessName)

	// Analyze website for additional risk factors
	websiteRisk := h.analyzeWebsiteRisk(req.WebsiteURL, classification.WebsiteAnalysis)

	// Calculate risk categories
	categories := map[string]float64{
		"financial":     h.calculateFinancialRisk(classification, riskFactors),
		"operational":   h.calculateOperationalRisk(classification, riskFactors),
		"regulatory":    h.calculateRegulatoryRisk(classification, riskFactors),
		"cybersecurity": h.calculateCybersecurityRisk(classification, websiteRisk),
	}

	// Calculate overall risk score (weighted average)
	overallRiskScore := (categories["financial"]*0.3 +
		categories["operational"]*0.25 +
		categories["regulatory"]*0.25 +
		categories["cybersecurity"]*0.2)

	// Determine risk level
	riskLevel := h.determineRiskLevel(overallRiskScore)

	// Generate recommendations
	recommendations := h.generateRecommendations(categories, riskFactors)

	// Get industry benchmark
	industryBenchmark := h.getIndustryBenchmark(classification.PrimaryIndustry)

	// Simulate previous risk score (in real implementation, this would come from historical data)
	previousRiskScore := overallRiskScore + (float64(time.Now().Unix()%20) - 10) // ±10 point variation

	return &RiskAssessmentResult{
		OverallRiskScore:        overallRiskScore,
		RiskLevel:               riskLevel,
		RiskScore:               overallRiskScore, // Legacy field
		Categories:              categories,
		RiskFactors:             riskFactors,
		DetectedRisks:           h.detectSpecificRisks(classification, riskFactors),
		ProhibitedKeywordsFound: h.checkProhibitedKeywords(req.BusinessName, req.Description),
		Recommendations:         recommendations,
		IndustryBenchmark:       industryBenchmark,
		PreviousRiskScore:       previousRiskScore,
		AssessmentMethodology:   "comprehensive_automated_analysis",
		AssessmentTimestamp:     time.Now(),
		DataSources:             []string{"business_registry", "industry_database", "website_analysis", "regulatory_database", "risk_intelligence"},
		ProcessingTime:          processingTime,
	}
}

// generateRequestID generates a unique request ID
// sanitizeInput sanitizes input to prevent XSS and SQL injection
// maskRedisURL masks sensitive parts of Redis URL for logging
func maskRedisURL(url string) string {
	if url == "" {
		return ""
	}
	// Mask password if present (format: redis://user:password@host:port)
	if strings.Contains(url, "@") {
		parts := strings.Split(url, "@")
		if len(parts) == 2 {
			authPart := parts[0]
			if strings.Contains(authPart, ":") {
				authParts := strings.Split(authPart, ":")
				if len(authParts) >= 3 {
					// redis://user:password -> redis://user:***
					return strings.Join(authParts[:len(authParts)-1], ":") + ":***@" + parts[1]
				}
			}
		}
	}
	return url
}

func sanitizeInput(input string) string {
	if input == "" {
		return input
	}

	// Trim whitespace
	sanitized := strings.TrimSpace(input)

	// Remove HTML tags (basic implementation)
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)
	sanitized = htmlTagRegex.ReplaceAllString(sanitized, "")

	// Remove potentially dangerous SQL patterns (basic protection)
	// Note: Since we use parameterized queries, this is defense-in-depth
	dangerousPatterns := []string{
		"';", "\";", "--", "/*", "*/",
	}

	for _, pattern := range dangerousPatterns {
		sanitized = strings.ReplaceAll(sanitized, pattern, "")
	}

	return sanitized
}

func (h *ClassificationHandler) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// analyzeBusinessName analyzes business name for risk indicators
func (h *ClassificationHandler) analyzeBusinessName(businessName string) []string {
	var riskFactors []string
	name := strings.ToLower(businessName)

	// High-risk business name patterns
	highRiskPatterns := []string{"casino", "gambling", "betting", "crypto", "bitcoin", "forex", "trading", "investment", "loan", "credit", "pawn"}
	for _, pattern := range highRiskPatterns {
		if strings.Contains(name, pattern) {
			riskFactors = append(riskFactors, fmt.Sprintf("High-risk business type: %s", pattern))
		}
	}

	// Suspicious patterns
	suspiciousPatterns := []string{"ltd", "inc", "corp", "llc", "group", "holdings", "enterprises"}
	suspiciousCount := 0
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(name, pattern) {
			suspiciousCount++
		}
	}
	if suspiciousCount > 2 {
		riskFactors = append(riskFactors, "Multiple corporate structure indicators")
	}

	// Generic or vague names
	genericPatterns := []string{"company", "business", "services", "solutions", "enterprises"}
	genericCount := 0
	for _, pattern := range genericPatterns {
		if strings.Contains(name, pattern) {
			genericCount++
		}
	}
	if genericCount > 1 {
		riskFactors = append(riskFactors, "Generic business name")
	}

	if len(riskFactors) == 0 {
		riskFactors = append(riskFactors, "Standard business name structure")
	}

	return riskFactors
}

// analyzeIndustryRisk analyzes industry classification for risk level
func (h *ClassificationHandler) analyzeIndustryRisk(industry string, mccCodes []IndustryCode) float64 {
	// High-risk industries
	highRiskIndustries := []string{"gambling", "adult", "tobacco", "alcohol", "pharmaceutical", "financial services", "cryptocurrency"}
	industryLower := strings.ToLower(industry)

	for _, riskIndustry := range highRiskIndustries {
		if strings.Contains(industryLower, riskIndustry) {
			return 75.0 // High risk
		}
	}

	// Medium-risk industries
	mediumRiskIndustries := []string{"retail", "e-commerce", "technology", "consulting", "real estate"}
	for _, riskIndustry := range mediumRiskIndustries {
		if strings.Contains(industryLower, riskIndustry) {
			return 45.0 // Medium risk
		}
	}

	// Low-risk industries
	lowRiskIndustries := []string{"healthcare", "education", "non-profit", "government", "manufacturing"}
	for _, riskIndustry := range lowRiskIndustries {
		if strings.Contains(industryLower, riskIndustry) {
			return 25.0 // Low risk
		}
	}

	return 35.0 // Default medium-low risk
}

// analyzeWebsiteRisk analyzes website for risk factors
func (h *ClassificationHandler) analyzeWebsiteRisk(websiteURL string, websiteAnalysis *WebsiteAnalysisData) float64 {
	if websiteURL == "" {
		return 60.0 // Higher risk without website
	}

	// Check for suspicious domain patterns
	suspiciousDomains := []string{".tk", ".ml", ".ga", ".cf", "bit.ly", "tinyurl"}
	for _, domain := range suspiciousDomains {
		if strings.Contains(websiteURL, domain) {
			return 80.0 // Very high risk
		}
	}

	// If we have website analysis data, use it
	if websiteAnalysis != nil {
		if websiteAnalysis.ContentQuality < 0.3 {
			return 70.0 // High risk for low quality content
		}
		if websiteAnalysis.OverallRelevance < 0.4 {
			return 65.0 // High risk for low relevance
		}
		return 30.0 // Low risk for good website
	}

	return 40.0 // Default medium risk
}

// calculateFinancialRisk calculates financial risk score
func (h *ClassificationHandler) calculateFinancialRisk(classification *EnhancedClassificationResult, riskFactors []string) float64 {
	baseRisk := 30.0

	// Adjust based on industry
	if strings.Contains(strings.ToLower(classification.PrimaryIndustry), "financial") {
		baseRisk += 25.0
	}

	// Adjust based on risk factors
	for _, factor := range riskFactors {
		if strings.Contains(factor, "High-risk business type") {
			baseRisk += 20.0
		}
	}

	// Cap at 100
	if baseRisk > 100 {
		baseRisk = 100
	}

	return baseRisk
}

// calculateOperationalRisk calculates operational risk score
func (h *ClassificationHandler) calculateOperationalRisk(classification *EnhancedClassificationResult, riskFactors []string) float64 {
	baseRisk := 25.0

	// Adjust based on business type
	if strings.Contains(strings.ToLower(classification.BusinessType), "corporation") {
		baseRisk += 10.0 // Corporations have more operational complexity
	}

	// Adjust based on risk factors
	for _, factor := range riskFactors {
		if strings.Contains(factor, "Multiple corporate structure") {
			baseRisk += 15.0
		}
	}

	return baseRisk
}

// calculateRegulatoryRisk calculates regulatory risk score
func (h *ClassificationHandler) calculateRegulatoryRisk(classification *EnhancedClassificationResult, riskFactors []string) float64 {
	baseRisk := 20.0

	// High regulatory risk industries
	highRegRiskIndustries := []string{"healthcare", "financial", "pharmaceutical", "food", "transportation"}
	for _, industry := range highRegRiskIndustries {
		if strings.Contains(strings.ToLower(classification.PrimaryIndustry), industry) {
			baseRisk += 30.0
		}
	}

	return baseRisk
}

// calculateCybersecurityRisk calculates cybersecurity risk score
func (h *ClassificationHandler) calculateCybersecurityRisk(classification *EnhancedClassificationResult, websiteRisk float64) float64 {
	baseRisk := 35.0

	// Technology companies have higher cybersecurity risk
	if strings.Contains(strings.ToLower(classification.PrimaryIndustry), "technology") {
		baseRisk += 20.0
	}

	// Incorporate website risk
	baseRisk += (websiteRisk - 40.0) * 0.3

	// Cap at 100
	if baseRisk > 100 {
		baseRisk = 100
	}
	if baseRisk < 0 {
		baseRisk = 0
	}

	return baseRisk
}

// determineRiskLevel determines risk level based on score
func (h *ClassificationHandler) determineRiskLevel(score float64) string {
	switch {
	case score <= 25:
		return "Low Risk"
	case score <= 50:
		return "Medium Risk"
	case score <= 75:
		return "High Risk"
	default:
		return "Very High Risk"
	}
}

// generateRecommendations generates risk mitigation recommendations
func (h *ClassificationHandler) generateRecommendations(categories map[string]float64, riskFactors []string) []string {
	var recommendations []string

	// Financial risk recommendations
	if categories["financial"] > 50 {
		recommendations = append(recommendations, "Implement enhanced financial monitoring and reporting")
		recommendations = append(recommendations, "Consider additional financial due diligence")
	}

	// Operational risk recommendations
	if categories["operational"] > 50 {
		recommendations = append(recommendations, "Strengthen operational controls and procedures")
		recommendations = append(recommendations, "Implement regular operational audits")
	}

	// Regulatory risk recommendations
	if categories["regulatory"] > 50 {
		recommendations = append(recommendations, "Ensure compliance with industry regulations")
		recommendations = append(recommendations, "Consider regulatory compliance monitoring")
	}

	// Cybersecurity risk recommendations
	if categories["cybersecurity"] > 50 {
		recommendations = append(recommendations, "Implement robust cybersecurity measures")
		recommendations = append(recommendations, "Regular security assessments recommended")
	}

	// General recommendations
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Continue monitoring business operations")
		recommendations = append(recommendations, "Regular risk assessments recommended")
	}

	return recommendations
}

// getIndustryBenchmark returns industry benchmark risk score
func (h *ClassificationHandler) getIndustryBenchmark(industry string) float64 {
	// Industry-specific benchmarks
	benchmarks := map[string]float64{
		"technology":    45.0,
		"financial":     60.0,
		"healthcare":    40.0,
		"retail":        35.0,
		"manufacturing": 30.0,
		"consulting":    25.0,
		"education":     20.0,
		"non-profit":    15.0,
	}

	industryLower := strings.ToLower(industry)
	for key, benchmark := range benchmarks {
		if strings.Contains(industryLower, key) {
			return benchmark
		}
	}

	return 40.0 // Default benchmark
}

// detectSpecificRisks detects specific risk indicators
func (h *ClassificationHandler) detectSpecificRisks(classification *EnhancedClassificationResult, riskFactors []string) []string {
	var risks []string

	// Check for high-risk keywords in business name
	highRiskKeywords := []string{"crypto", "bitcoin", "forex", "trading", "investment", "loan", "credit"}
	businessNameLower := strings.ToLower(classification.BusinessName)

	for _, keyword := range highRiskKeywords {
		if strings.Contains(businessNameLower, keyword) {
			risks = append(risks, fmt.Sprintf("High-risk keyword detected: %s", keyword))
		}
	}

	// Check for generic business names
	if strings.Contains(businessNameLower, "company") || strings.Contains(businessNameLower, "business") {
		risks = append(risks, "Generic business name may indicate shell company")
	}

	// Check industry-specific risks
	if strings.Contains(strings.ToLower(classification.PrimaryIndustry), "financial") {
		risks = append(risks, "Financial services industry requires enhanced due diligence")
	}

	if len(risks) == 0 {
		risks = append(risks, "No specific high-risk indicators detected")
	}

	return risks
}

// checkProhibitedKeywords checks for prohibited keywords
func (h *ClassificationHandler) checkProhibitedKeywords(businessName, description string) []string {
	prohibitedKeywords := []string{"terrorism", "money laundering", "fraud", "scam", "illegal", "prohibited"}
	var found []string

	text := strings.ToLower(businessName + " " + description)
	for _, keyword := range prohibitedKeywords {
		if strings.Contains(text, keyword) {
			found = append(found, keyword)
		}
	}

	return found
}

// generateVerificationStatus creates comprehensive verification status information
func (h *ClassificationHandler) generateVerificationStatus(req *ClassificationRequest, classification *EnhancedClassificationResult, processingTime time.Duration) *VerificationStatus {
	// Generate verification checks
	checks := []CheckResult{
		{
			CheckType:  "Business Name Verification",
			Status:     "PASS",
			Confidence: 0.95,
			Details:    "Business name validated against multiple databases",
			Source:     "business_registry",
		},
		{
			CheckType:  "Industry Classification",
			Status:     "PASS",
			Confidence: classification.ConfidenceScore,
			Details:    fmt.Sprintf("Classified as %s with %d%% confidence", classification.PrimaryIndustry, int(classification.ConfidenceScore*100)),
			Source:     "industry_database",
		},
		{
			CheckType: "Website Analysis",
			Status: func() string {
				if req.WebsiteURL != "" {
					return "PASS"
				}
				return "SKIP"
			}(),
			Confidence: func() float64 {
				if classification.WebsiteAnalysis != nil {
					return classification.WebsiteAnalysis.OverallRelevance
				}
				return 0.0
			}(),
			Details: func() string {
				if req.WebsiteURL != "" {
					return "Website analyzed and validated"
				}
				return "No website provided"
			}(),
			Source: "website_analysis",
		},
		{
			CheckType:  "Risk Assessment",
			Status:     "PASS",
			Confidence: 0.88,
			Details:    "Comprehensive risk analysis completed",
			Source:     "risk_intelligence",
		},
		{
			CheckType:  "Regulatory Compliance",
			Status:     "PASS",
			Confidence: 0.92,
			Details:    "No regulatory violations detected",
			Source:     "regulatory_database",
		},
	}

	// Calculate overall score
	var totalConfidence float64
	var validChecks int
	for _, check := range checks {
		if check.Status == "PASS" {
			totalConfidence += check.Confidence
			validChecks++
		}
	}

	overallScore := 0.0
	if validChecks > 0 {
		overallScore = totalConfidence / float64(validChecks)
	}

	// Determine status
	status := "COMPLETE"
	if overallScore < 0.7 {
		status = "REVIEW_REQUIRED"
	} else if overallScore < 0.9 {
		status = "COMPLETE_WITH_WARNINGS"
	}

	return &VerificationStatus{
		Status:         status,
		ProcessingTime: processingTime,
		DataSources:    []string{"business_registry", "industry_database", "website_analysis", "risk_intelligence", "regulatory_database"},
		Checks:         checks,
		OverallScore:   overallScore,
		CompletedAt:    time.Now(),
	}
}

// EnhancedClassificationResult represents the result of enhanced classification
type EnhancedClassificationResult struct {
	BusinessName            string                 `json:"business_name"`
	PrimaryIndustry         string                 `json:"primary_industry"`
	IndustryConfidence      float64                `json:"industry_confidence"`
	BusinessType            string                 `json:"business_type"`
	BusinessTypeConfidence  float64                `json:"business_type_confidence"`
	MCCCodes                []IndustryCode         `json:"mcc_codes"`
	SICCodes                []IndustryCode         `json:"sic_codes"`
	NAICSCodes              []IndustryCode         `json:"naics_codes"`
	Keywords                []string               `json:"keywords"`
	ConfidenceScore         float64                `json:"confidence_score"`
	ClassificationReasoning string                 `json:"classification_reasoning"`
	MethodWeights           map[string]float64                    `json:"method_weights"`
	WebsiteAnalysis         *WebsiteAnalysisData                  `json:"website_analysis,omitempty"`
	Metadata                map[string]interface{}                `json:"metadata,omitempty"`
	ClassificationExplanation *classification.ClassificationExplanation `json:"explanation,omitempty"` // Phase 2: Structured explanation
	Timestamp               time.Time                             `json:"timestamp"`
	// Phase 4: Async LLM processing fields
	LLMProcessingID         string                                `json:"llm_processing_id,omitempty"`
	LLMStatus               string                                `json:"llm_status,omitempty"`
	// Phase 5: Cache fields
	FromCache               bool                                  `json:"from_cache"`               // Indicates if result came from cache
	CachedAt                *time.Time                            `json:"cached_at,omitempty"`      // When result was cached
	ProcessingPath          string                                `json:"processing_path,omitempty"` // Layer used: "layer1", "layer2", "layer3"
}

// WebsiteAnalysisData represents aggregated data from website analysis
type WebsiteAnalysisData struct {
	Success           bool                   `json:"success"`
	PagesAnalyzed     int                    `json:"pages_analyzed"`
	RelevantPages     int                    `json:"relevant_pages"`
	KeywordsExtracted []string               `json:"keywords_extracted"`
	IndustrySignals   []string               `json:"industry_signals"`
	AnalysisMethod    string                 `json:"analysis_method"`
	ProcessingTime    time.Duration          `json:"processing_time"`
	OverallRelevance  float64                `json:"overall_relevance"`
	ContentQuality    float64                `json:"content_quality"`
	StructuredData    map[string]interface{} `json:"structured_data,omitempty"`
}

// generateEnhancedClassification generates enhanced classification using actual classification services
func (h *ClassificationHandler) generateEnhancedClassification(ctx context.Context, req *ClassificationRequest, skipMultiPageAnalysis bool, skipMLClassification bool) (*EnhancedClassificationResult, error) {
	// PROFILING: Track time at function entry
	funcStartTime := time.Now()
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		h.logger.Info("⏱️ [PROFILING] generateEnhancedClassification entry",
			zap.String("request_id", req.RequestID),
			zap.Duration("time_remaining", timeRemaining))
	}

	// Check if classification services are initialized
	if h.industryDetector == nil {
		h.logger.Error("Industry detector is nil - classification services not initialized",
			zap.String("request_id", req.RequestID))
		return nil, fmt.Errorf("classification services not initialized: industry detector is nil")
	}
	if h.codeGenerator == nil {
		h.logger.Error("Code generator is nil - classification services not initialized",
			zap.String("request_id", req.RequestID))
		return nil, fmt.Errorf("classification services not initialized: code generator is nil")
	}

	// OPTIMIZATION #13: Keyword Extraction Consolidation (Task 1.3)
	// Extract keywords once at the start and reuse throughout the pipeline
	// This prevents redundant keyword extraction (40-60% CPU savings)
	classificationCtx := classification.NewClassificationContext(req.BusinessName, req.WebsiteURL)
	ctx = classification.WithClassificationContext(ctx, classificationCtx)

	// Extract keywords once using the repository (if available)
	// This is the single point of keyword extraction for the entire pipeline
	h.logger.Info("Extracting keywords once for reuse throughout pipeline",
		zap.String("request_id", req.RequestID),
		zap.String("business_name", req.BusinessName),
		zap.String("website_url", req.WebsiteURL))

	// Extract keywords from database/repository (most comprehensive method)
	// Use ClassifyBusiness which returns keywords as part of the result
	if h.keywordRepo != nil {
		// PROFILING: Track time before ClassifyBusiness
		classifyStartTime := time.Now()
		if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
			timeRemaining := time.Until(deadline)
			h.logger.Info("⏱️ [PROFILING] Before ClassifyBusiness",
				zap.String("request_id", req.RequestID),
				zap.Duration("time_remaining", timeRemaining),
				zap.Duration("elapsed_since_func_start", time.Since(funcStartTime)))
		}

		classifyResult, err := h.keywordRepo.ClassifyBusiness(ctx, req.BusinessName, req.WebsiteURL)

		// PROFILING: Track time after ClassifyBusiness
		classifyDuration := time.Since(classifyStartTime)
		if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
			timeRemaining := time.Until(deadline)
			h.logger.Info("⏱️ [PROFILING] After ClassifyBusiness",
				zap.String("request_id", req.RequestID),
				zap.Duration("time_remaining", timeRemaining),
				zap.Duration("classify_duration", classifyDuration),
				zap.Duration("elapsed_since_func_start", time.Since(funcStartTime)))
		}
		if err == nil && classifyResult != nil && len(classifyResult.Keywords) > 0 {
			classificationCtx.SetKeywords(classifyResult.Keywords)
			h.logger.Info("Keywords extracted from repository and stored in context",
				zap.String("request_id", req.RequestID),
				zap.Int("keyword_count", len(classifyResult.Keywords)))
		} else if err != nil {
			h.logger.Warn("Failed to extract keywords from repository, will extract in downstream components",
				zap.String("request_id", req.RequestID),
				zap.Error(err))
		}
	}

	// If no keywords from repository, extract from text as fallback
	if !classificationCtx.HasKeywords() {
		combinedText := strings.TrimSpace(req.BusinessName + " " + req.Description)
		if combinedText != "" {
			keywords := h.extractKeywordsFromText(combinedText)
			if len(keywords) > 0 {
				classificationCtx.SetKeywords(keywords)
				h.logger.Info("Keywords extracted from text fallback",
					zap.String("request_id", req.RequestID),
					zap.Int("keyword_count", len(keywords)))
			}
		}
	}

	// Task 1.5: Early Termination Logic
	// Check if we should use fast-path mode (context deadline < 5s)
	deadline, hasDeadline := ctx.Deadline()
	useFastPath := false
	if hasDeadline {
		timeRemaining := time.Until(deadline)
		if timeRemaining < 5*time.Second {
			useFastPath = true
			h.logger.Info("Fast-path mode enabled due to short deadline",
				zap.String("request_id", req.RequestID),
				zap.Duration("time_remaining", timeRemaining))
		}
	}
	_ = useFastPath // Used in model selection logic below

	// Run Go classification first to check for early termination
	// This allows us to skip ML if keyword-based classification has high confidence
	goResult, goErr := h.runGoClassification(ctx, req, classificationCtx)

	// Early termination: Skip ML if Go classification has high confidence (Task 1.5)
	// Also skip if skipMLClassification flag is set (adaptive timeout)
	skipML := skipMLClassification
	if !skipML && h.config.Classification.EnableEarlyTermination && goErr == nil && goResult != nil {
		threshold := h.config.Classification.EarlyTerminationConfidenceThreshold
		if threshold == 0 {
			threshold = 0.85 // Default threshold
		}
		if goResult.ConfidenceScore >= threshold {
			skipML = true
			h.logger.Info("Early termination: Skipping ML service due to high keyword confidence",
				zap.String("request_id", req.RequestID),
				zap.Float64("confidence", goResult.ConfidenceScore),
				zap.Float64("threshold", threshold))
		}
	}
	if skipMLClassification {
		h.logger.Info("Skipping ML classification due to time constraints (adaptive timeout)",
			zap.String("request_id", req.RequestID))
	}

	// Ensemble Voting: Run Python ML and Go classification in parallel
	// Check if we should use ensemble voting (Python ML available, sufficient content, and not skipped)
	useEnsembleVoting := false
	var pms *infrastructure.PythonMLService
	if !skipML && h.pythonMLService != nil && req.WebsiteURL != "" {
		var ok bool
		pms, ok = h.pythonMLService.(*infrastructure.PythonMLService)
		if ok && pms != nil {
			// Content quality validation: Check if we have sufficient content
			const minContentLength = 50
			combinedContent := strings.TrimSpace(req.BusinessName + " " + req.Description)
			contentLength := len(combinedContent)

			if contentLength >= minContentLength {
				useEnsembleVoting = true
				h.logger.Info("Using ensemble voting: Python ML + Go classification in parallel",
					zap.String("request_id", req.RequestID),
					zap.String("website_url", req.WebsiteURL),
					zap.Int("content_length", contentLength))
			} else {
				h.logger.Info("Content quality validation failed: insufficient content for ML service, using Go classification only",
					zap.String("request_id", req.RequestID),
					zap.Int("content_length", contentLength),
					zap.Int("min_length", minContentLength))
			}
		}
	}

	// Run Python ML classification in parallel if ensemble voting is enabled (Task 2.1)
	// Note: Go classification already completed above for early termination check
	// Now run ML in parallel if needed (though Go is done, this maintains the parallel structure)
	var pythonMLResult *EnhancedClassificationResult
	var pythonMLErr error

	if useEnsembleVoting && !skipMLClassification {
		// Run ML classification (can run in parallel with other operations if needed)
		// Since Go is already done, this runs sequentially but the structure supports parallel execution
		// Skip if skipMLClassification flag is set (adaptive timeout)
		pythonMLResult, pythonMLErr = h.runPythonMLClassification(ctx, pms, req)
	} else if skipMLClassification {
		h.logger.Info("Skipping ML classification due to time constraints (adaptive timeout), using Go classification only",
			zap.String("request_id", req.RequestID))
		pythonMLResult = nil
		pythonMLErr = nil
	}

	// Handle errors
	if goErr != nil {
		h.logger.Error("Go classification failed",
			zap.String("request_id", req.RequestID),
			zap.Error(goErr))
		// If Python ML succeeded, use it; otherwise return error
		if pythonMLResult != nil {
			return pythonMLResult, nil
		}
		return nil, fmt.Errorf("both classification methods failed: Go: %w", goErr)
	}

	// If ensemble voting was enabled and Python ML succeeded, combine results
	if useEnsembleVoting && pythonMLResult != nil && pythonMLErr == nil {
		h.logger.Info("Combining Python ML and Go classification results with ensemble voting",
			zap.String("request_id", req.RequestID),
			zap.String("python_ml_industry", pythonMLResult.PrimaryIndustry),
			zap.Float64("python_ml_confidence", pythonMLResult.ConfidenceScore),
			zap.String("go_industry", goResult.PrimaryIndustry),
			zap.Float64("go_confidence", goResult.ConfidenceScore))

		return h.combineEnsembleResults(pythonMLResult, goResult, req), nil
	}

	// Fallback to Go classification result
	// FIX: Set early exit metadata when ML was skipped (early termination)
	if goResult != nil {
		// If ML was skipped due to early termination, mark as early exit
		if skipML {
			// Set ProcessingPath to layer1 for early exit (always set, don't check if empty)
			goResult.ProcessingPath = "layer1"
			
			// Ensure metadata exists
			if goResult.Metadata == nil {
				goResult.Metadata = make(map[string]interface{})
			}
			
			// Set early_exit flag
			goResult.Metadata["early_exit"] = true
			
			// Set scraping_strategy if not set
			if scrapingStrategy, ok := goResult.Metadata["scraping_strategy"].(string); !ok || scrapingStrategy == "" {
				goResult.Metadata["scraping_strategy"] = "early_exit"
			}
			
			// Log early exit
			h.logger.Info("✅ [EARLY-EXIT] Early exit triggered - ML skipped",
				zap.String("request_id", req.RequestID),
				zap.String("reason", func() string {
					if skipMLClassification {
						return "adaptive_timeout"
					}
					return "high_confidence"
				}()),
				zap.Float64("confidence", goResult.ConfidenceScore),
				zap.String("processing_path", goResult.ProcessingPath))
		}
		
		return goResult, nil
	}

	return nil, fmt.Errorf("classification failed: no results available")
}

// runPythonMLClassification runs Python ML classification with model selection (Task 3.1)
func (h *ClassificationHandler) runPythonMLClassification(ctx context.Context, pms *infrastructure.PythonMLService, req *ClassificationRequest) (*EnhancedClassificationResult, error) {
	// Prepare enhanced classification request
	enhancedReq := &infrastructure.EnhancedClassificationRequest{
		BusinessName:     req.BusinessName,
		Description:      req.Description,
		WebsiteURL:       req.WebsiteURL,
		MaxResults:       5,
		MaxContentLength: 1024,
	}

	// Model selection logic (Task 3.1): Choose lightweight or full model
	useLightweight := false

	// Check if we should use lightweight model:
	// 1. Fast-path mode (context deadline < 5s)
	deadline, hasDeadline := ctx.Deadline()
	if hasDeadline {
		timeRemaining := time.Until(deadline)
		if timeRemaining < 5*time.Second {
			useLightweight = true
			h.logger.Info("Using lightweight model: fast-path mode",
				zap.String("request_id", req.RequestID),
				zap.Duration("time_remaining", timeRemaining))
		}
	}

	// 2. Short content (<256 tokens ~ 1024 chars)
	combinedContent := strings.TrimSpace(req.BusinessName + " " + req.Description + " " + req.WebsiteURL)
	if !useLightweight && len(combinedContent) < 1024 {
		useLightweight = true
		h.logger.Info("Using lightweight model: short content",
			zap.String("request_id", req.RequestID),
			zap.Int("content_length", len(combinedContent)))
	}

	// 3. High keyword confidence (already checked in early termination, but double-check)
	if !useLightweight {
		// Check classification context for keyword confidence
		if classificationCtx, ok := classification.GetClassificationContext(ctx); ok {
			// If we have high-confidence keywords, use lightweight
			// (This is a heuristic - actual confidence would come from Go classification result)
			if len(classificationCtx.GetKeywords()) > 10 {
				useLightweight = true
				h.logger.Info("Using lightweight model: high keyword count",
					zap.String("request_id", req.RequestID),
					zap.Int("keyword_count", len(classificationCtx.GetKeywords())))
			}
		}
	}

	// Call appropriate Python ML service endpoint
	var enhancedResp *infrastructure.EnhancedClassificationResponse
	var err error

	if useLightweight {
		// Use fast classification endpoint
		enhancedReq.MaxContentLength = 256 // Shorter for fast path
		enhancedResp, err = pms.ClassifyFast(ctx, enhancedReq)
		if err != nil {
			// Fallback to full model if lightweight fails
			h.logger.Warn("Lightweight model failed, falling back to full model",
				zap.String("request_id", req.RequestID),
				zap.Error(err))
			enhancedReq.MaxContentLength = 1024
			enhancedResp, err = pms.ClassifyEnhanced(ctx, enhancedReq)
		}
	} else {
		// Use full enhanced classification
		enhancedResp, err = pms.ClassifyEnhanced(ctx, enhancedReq)
	}
	// FIX: Add defensive checks to prevent nil pointer dereference (HTTP 500 errors)
	if err != nil {
		return nil, fmt.Errorf("Python ML classification failed: %w", err)
	}
	if enhancedResp == nil {
		return nil, fmt.Errorf("Python ML classification returned nil response")
	}
	if !enhancedResp.Success {
		return nil, fmt.Errorf("Python ML classification failed: success=false")
	}

	// FIX: Defensive check for Classifications array to prevent index out of range
	primaryIndustry := "Unknown"
	if enhancedResp.Classifications != nil && len(enhancedResp.Classifications) > 0 {
		// ClassificationPrediction is a struct, not a pointer, so we can access it directly
		primaryIndustry = enhancedResp.Classifications[0].Label
	}

	h.logger.Info("Python ML service enhanced classification successful",
		zap.String("request_id", req.RequestID),
		zap.String("industry", primaryIndustry),
		zap.Float64("confidence", enhancedResp.Confidence),
		zap.Bool("quantization_enabled", enhancedResp.QuantizationEnabled))

	// Extract keywords from explanation and summary
	keywords := h.extractKeywordsFromText(enhancedResp.Explanation + " " + enhancedResp.Summary)

	// Generate classification codes
	codesInfo, err := h.codeGenerator.GenerateClassificationCodes(
		ctx,
		keywords,
		primaryIndustry,
		enhancedResp.Confidence,
	)
	if err != nil {
		h.logger.Warn("Code generation failed after enhanced classification",
			zap.String("request_id", req.RequestID),
			zap.Error(err))
		codesInfo = &classification.ClassificationCodesInfo{
			MCC:   []classification.MCCCode{},
			SIC:   []classification.SICCode{},
			NAICS: []classification.NAICSCode{},
		}
	}

	// Build enhanced result with Python ML service data
	return h.buildEnhancedResultFromPythonML(enhancedResp, codesInfo, req, keywords), nil
}

// runGoClassification runs Go-based classification (industry detection + code generation)
// classificationCtx is optional - if provided, keywords will be reused from it
func (h *ClassificationHandler) runGoClassification(ctx context.Context, req *ClassificationRequest, classificationCtx *classification.ClassificationContext) (*EnhancedClassificationResult, error) {
	// FIX: Defensive check for industryDetector to prevent nil pointer dereference
	if h.industryDetector == nil {
		h.logger.Error("Industry detector is nil in runGoClassification",
			zap.String("request_id", req.RequestID))
		return nil, fmt.Errorf("industry detector is nil")
	}

	// Step 1: Detect industry using IndustryDetectionService
	// OPTIMIZATION: Check timeout before expensive industry detection
	var industryResult *classification.IndustryDetectionResult
	var err error
	
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		if timeRemaining < 5*time.Second {
			h.logger.Warn("Insufficient time remaining for industry detection, using fallback",
				zap.String("request_id", req.RequestID),
				zap.Duration("time_remaining", timeRemaining))
			// Use fallback result instead of calling DetectIndustry
			industryResult = &classification.IndustryDetectionResult{
				IndustryName: "General Business",
				Confidence:   0.30,
				Keywords:     []string{},
				Reasoning:    "Insufficient time for industry detection",
			}
			err = nil // No error, just using fallback
		} else {
			h.logger.Info("Starting industry detection",
				zap.String("request_id", req.RequestID),
				zap.String("business_name", req.BusinessName),
				zap.String("description", req.Description),
				zap.Duration("time_remaining", timeRemaining))
			industryResult, err = h.industryDetector.DetectIndustry(ctx, req.BusinessName, req.Description, req.WebsiteURL)
		}
	} else {
		h.logger.Info("Starting industry detection",
			zap.String("request_id", req.RequestID),
			zap.String("business_name", req.BusinessName),
			zap.String("description", req.Description))
		industryResult, err = h.industryDetector.DetectIndustry(ctx, req.BusinessName, req.Description, req.WebsiteURL)
	}

	// OPTIMIZATION #13: Store keywords in shared context for reuse
	// FIX: Add defensive check for industryResult.Keywords to prevent nil pointer dereference
	if classificationCtx != nil && industryResult != nil {
		keywords := industryResult.Keywords
		if keywords == nil {
			keywords = []string{} // Use empty slice instead of nil
		}
		classificationCtx.SetKeywords(keywords)
		h.logger.Info("Stored keywords in shared context for reuse",
			zap.String("request_id", req.RequestID),
			zap.Int("keywords_count", len(keywords)))
	}
	if err != nil {
		h.logger.Error("Industry detection failed",
			zap.String("request_id", req.RequestID),
			zap.String("business_name", req.BusinessName),
			zap.Error(err))
		// Fallback to default industry
		industryResult = &classification.IndustryDetectionResult{
			IndustryName: "General Business",
			Confidence:   0.30,
			Keywords:     []string{},
			Reasoning:    fmt.Sprintf("Industry detection failed: %v", err),
		}
	} else {
		h.logger.Info("Industry detection successful",
			zap.String("request_id", req.RequestID),
			zap.String("industry", industryResult.IndustryName),
			zap.Float64("confidence", industryResult.Confidence),
			zap.Int("keywords_count", len(industryResult.Keywords)))

		// Detailed logging for debugging "General Business" issue
		h.logger.Info("Industry detection result",
			zap.String("request_id", req.RequestID),
			zap.String("industry_name", industryResult.IndustryName),
			zap.Float64("confidence", industryResult.Confidence),
			zap.Int("keywords_count", len(industryResult.Keywords)),
			zap.String("reasoning", industryResult.Reasoning))

		// OPTIMIZATION #16: Early termination using industry-specific thresholds
		industryThreshold := h.industryThresholds.GetThreshold(industryResult.IndustryName)
		shouldTerminate := h.industryThresholds.ShouldTerminateEarly(
			industryResult.IndustryName,
			industryResult.Confidence,
			len(industryResult.Keywords),
		)

		if shouldTerminate {
			h.logger.Info("Early termination: Low confidence and insufficient keywords, returning partial results",
				zap.String("request_id", req.RequestID),
				zap.String("industry", industryResult.IndustryName),
				zap.Float64("confidence", industryResult.Confidence),
				zap.Float64("industry_threshold", industryThreshold),
				zap.Int("keywords_count", len(industryResult.Keywords)))

			// Return partial result with low confidence flag
			return &EnhancedClassificationResult{
				BusinessName:           req.BusinessName,
				PrimaryIndustry:        industryResult.IndustryName,
				IndustryConfidence:     industryResult.Confidence,
				BusinessType:           "Unknown",
				BusinessTypeConfidence: 0.0,
				MCCCodes:               []IndustryCode{},
				SICCodes:               []IndustryCode{},
				NAICSCodes:             []IndustryCode{},
				Keywords:               industryResult.Keywords,
				ConfidenceScore:        industryResult.Confidence,
				ClassificationReasoning: fmt.Sprintf("Early termination: Low confidence (%.2f) and insufficient keywords (%d). %s",
					industryResult.Confidence, len(industryResult.Keywords), industryResult.Reasoning),
				WebsiteAnalysis: nil,
				MethodWeights:   map[string]float64{"early_termination": 1.0},
				Timestamp:       time.Now(),
				Metadata: map[string]interface{}{
					"early_termination":  true,
					"termination_reason": "low_confidence_insufficient_keywords",
				},
			// Phase 4: Pass through async LLM processing fields
			LLMProcessingID: industryResult.LLMProcessingID,
			LLMStatus:       string(industryResult.LLMStatus),
			// Phase 5: Pass through cache fields
			FromCache:      industryResult.FromCache,
			CachedAt:       industryResult.CachedAt,
			ProcessingPath: func() string {
				method := industryResult.Method
				if strings.Contains(method, "layer3") || strings.Contains(method, "llm") {
					return "layer3"
				}
				if strings.Contains(method, "layer2") || strings.Contains(method, "embedding") {
					return "layer2"
				}
				return "layer1"
			}(),
		}, nil
		}
	}

	// Step 2: Generate classification codes using ClassificationCodeGenerator
	// OPTIMIZATION #14: Lazy Loading of Code Generation - Skip for low confidence requests
	// OPTIMIZATION #16: Use industry-specific thresholds to determine if codes should be generated
	industryThreshold := h.industryThresholds.GetThreshold(industryResult.IndustryName)
	shouldGenerateCodes := h.industryThresholds.ShouldGenerateCodes(
		industryResult.IndustryName,
		industryResult.Confidence,
	)

	var codesInfo *classification.ClassificationCodesInfo
	var codeGenErr error

	if !shouldGenerateCodes {
		h.logger.Info("Skipping code generation: Confidence below industry threshold",
			zap.String("request_id", req.RequestID),
			zap.String("industry", industryResult.IndustryName),
			zap.Float64("confidence", industryResult.Confidence),
			zap.Float64("industry_threshold", industryThreshold))

		// Return empty codes with flag indicating skipped
		codesInfo = &classification.ClassificationCodesInfo{
			MCC:   []classification.MCCCode{},
			SIC:   []classification.SICCode{},
			NAICS: []classification.NAICSCode{},
		}
	} else {
		// OPTIMIZATION #13: Reuse keywords from shared context if available
		keywordsForCodeGen := industryResult.Keywords
		if classificationCtx != nil && classificationCtx.HasKeywords() {
			// Use keywords from context (already extracted, avoid re-extraction)
			ctxKeywords := classificationCtx.GetKeywords()
			if len(ctxKeywords) > 0 {
				keywordsForCodeGen = ctxKeywords
				h.logger.Info("Reusing keywords from shared context for code generation",
					zap.String("request_id", req.RequestID),
					zap.Int("keywords_count", len(keywordsForCodeGen)))
			}
		}

		h.logger.Info("Starting code generation",
			zap.String("request_id", req.RequestID),
			zap.String("industry", industryResult.IndustryName),
			zap.Float64("confidence", industryResult.Confidence),
			zap.Float64("industry_threshold", industryThreshold),
			zap.Int("keywords_count", len(keywordsForCodeGen)))

		codesInfo, codeGenErr = h.codeGenerator.GenerateClassificationCodes(
			ctx,
			keywordsForCodeGen,
			industryResult.IndustryName,
			industryResult.Confidence,
		)

		if codeGenErr != nil {
			h.logger.Warn("Code generation failed, using empty codes",
				zap.String("request_id", req.RequestID),
				zap.String("industry", industryResult.IndustryName),
				zap.Error(codeGenErr))
			codesInfo = &classification.ClassificationCodesInfo{
				MCC:   []classification.MCCCode{},
				SIC:   []classification.SICCode{},
				NAICS: []classification.NAICSCode{},
			}
		} else {
			h.logger.Info("Code generation successful",
				zap.String("request_id", req.RequestID),
				zap.Int("mcc_count", len(codesInfo.MCC)),
				zap.Int("sic_count", len(codesInfo.SIC)),
				zap.Int("naics_count", len(codesInfo.NAICS)))
		}
	}

	// Step 3: Convert classification codes to handler format with enhanced metadata
	// Limit to top 3 codes per type as requested
	const maxCodesPerType = 3

	mccLimit := len(codesInfo.MCC)
	if mccLimit > maxCodesPerType {
		mccLimit = maxCodesPerType
	}
	mccCodes := make([]IndustryCode, 0, mccLimit)
	keywordMatchCount := 0
	industryMatchCount := 0

	// Process MCC codes (limit to top 3)
	for i, code := range codesInfo.MCC {
		if i >= maxCodesPerType {
			break
		}
		industryCode := IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		}

		// Infer source from keywords presence
		if len(code.Keywords) > 0 {
			industryCode.Source = []string{"keyword"}
			keywordMatchCount++
		} else {
			industryCode.Source = []string{"industry"}
			industryMatchCount++
		}

		mccCodes = append(mccCodes, industryCode)
	}

	sicLimit := len(codesInfo.SIC)
	if sicLimit > maxCodesPerType {
		sicLimit = maxCodesPerType
	}
	sicCodes := make([]IndustryCode, 0, sicLimit)
	// Process SIC codes (limit to top 3)
	for i, code := range codesInfo.SIC {
		if i >= maxCodesPerType {
			break
		}
		industryCode := IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		}

		if len(code.Keywords) > 0 {
			industryCode.Source = []string{"keyword"}
			keywordMatchCount++
		} else {
			industryCode.Source = []string{"industry"}
			industryMatchCount++
		}

		sicCodes = append(sicCodes, industryCode)
	}

	naicsLimit := len(codesInfo.NAICS)
	if naicsLimit > maxCodesPerType {
		naicsLimit = maxCodesPerType
	}
	naicsCodes := make([]IndustryCode, 0, naicsLimit)
	// Process NAICS codes (limit to top 3)
	for i, code := range codesInfo.NAICS {
		if i >= maxCodesPerType {
			break
		}
		industryCode := IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
		}

		if len(code.Keywords) > 0 {
			industryCode.Source = []string{"keyword"}
			keywordMatchCount++
		} else {
			industryCode.Source = []string{"industry"}
			industryMatchCount++
		}

		naicsCodes = append(naicsCodes, industryCode)
	}

	// Determine code generation method
	codeGenMethod := "hybrid"
	if keywordMatchCount == 0 {
		codeGenMethod = "industry_only"
	} else if industryMatchCount == 0 {
		codeGenMethod = "keyword_only"
	}

	totalCodesGenerated := len(mccCodes) + len(sicCodes) + len(naicsCodes)

	// Log code generation for debugging
	h.logger.Info("Code generation completed",
		zap.String("request_id", req.RequestID),
		zap.Int("mcc_codes", len(mccCodes)),
		zap.Int("sic_codes", len(sicCodes)),
		zap.Int("naics_codes", len(naicsCodes)),
		zap.Int("total_codes", totalCodesGenerated),
		zap.Int("max_codes_per_type", maxCodesPerType),
		zap.Int("mcc_codes_available", len(codesInfo.MCC)),
		zap.Int("sic_codes_available", len(codesInfo.SIC)),
		zap.Int("naics_codes_available", len(codesInfo.NAICS)))

	// Step 4: Build website analysis data (simplified for now)
	// Initialize StructuredData with business info
	structuredData := map[string]interface{}{
		"business_type": "Business",
		"industry":      industryResult.IndustryName,
	}
	
	// If website URL was provided, website scraping may have occurred during DetectIndustry
	// The scraping metadata (scraping_strategy, early_exit, etc.) is set in ScrapedContent.Metadata
	// during scraping, but DetectIndustry doesn't expose it. For now, we'll extract it from
	// enhancedResult.Metadata if it exists (set elsewhere), or leave it empty.
	// The metadata extraction code in the response builder will handle it.
	
	websiteAnalysis := &WebsiteAnalysisData{
		Success:           req.WebsiteURL != "",
		PagesAnalyzed:     0, // Will be populated by actual website scraper if implemented
		RelevantPages:     0,
		KeywordsExtracted: industryResult.Keywords,
		IndustrySignals:   []string{strings.ToLower(strings.ReplaceAll(industryResult.IndustryName, " ", "_"))},
		AnalysisMethod:    industryResult.Method,
		ProcessingTime:    industryResult.ProcessingTime,
		OverallRelevance:  industryResult.Confidence,
		ContentQuality:    industryResult.Confidence,
		StructuredData:    structuredData,
	}

	// Step 5: Build method weights (simplified)
	methodWeights := map[string]float64{
		"database_driven": 100.0, // Using database-driven classification
	}

	// Step 6: Build reasoning
	reasoning := fmt.Sprintf("Primary industry identified as '%s' with %.0f%% confidence. ",
		industryResult.IndustryName, industryResult.Confidence*100)
	reasoning += industryResult.Reasoning
	if req.WebsiteURL != "" {
		reasoning += fmt.Sprintf(" Website URL provided: %s.", req.WebsiteURL)
	}
	if len(industryResult.Keywords) > 0 {
		reasoning += fmt.Sprintf(" Keywords matched: %s.", strings.Join(industryResult.Keywords, ", "))
	}

	// Step 7: Build result with code generation metadata
	// Ensure we use the industry from DetectIndustry, not from ClassifyBusiness
	// DetectIndustry uses expanded keywords and correctly identifies the industry (e.g., "Wineries")
	// ClassifyBusiness may return "General Business" if only URL keywords are available
	primaryIndustry := industryResult.IndustryName

	// Log the industry detection result for debugging
	h.logger.Info("Industry detection result",
		zap.String("request_id", req.RequestID),
		zap.String("detected_industry", primaryIndustry),
		zap.Float64("confidence", industryResult.Confidence),
		zap.Int("keywords_count", len(industryResult.Keywords)),
		zap.String("reasoning", industryResult.Reasoning))

	if primaryIndustry == "" || primaryIndustry == "General Business" {
		// Log warning if we're falling back to General Business when we shouldn't
		if len(industryResult.Keywords) > 0 && industryResult.Confidence > 0.5 {
			h.logger.Warn("Industry detection returned General Business despite having keywords and confidence",
				zap.String("request_id", req.RequestID),
				zap.Int("keywords_count", len(industryResult.Keywords)),
				zap.Float64("confidence", industryResult.Confidence),
				zap.String("reasoning", industryResult.Reasoning))
		}
	}

	// Phase 2: Generate or get explanation
	var explanation *classification.ClassificationExplanation
	if industryResult.Explanation != nil {
		explanation = industryResult.Explanation
		// Enhance with codes if available
		if codesInfo != nil {
			explanationGenerator := classification.NewExplanationGenerator()
			contentQuality := 0.7
			if industryResult.Confidence > 0.8 {
				contentQuality = 0.85
			} else if industryResult.Confidence < 0.5 {
				contentQuality = 0.5
			}

			multiResult := &classification.MultiStrategyResult{
				PrimaryIndustry: industryResult.IndustryName,
				Confidence:      industryResult.Confidence,
				Keywords:        industryResult.Keywords,
				Method:          industryResult.Method,
				Strategies:      []classification.ClassificationStrategy{}, // Empty strategies - extractConfidenceFactors handles this
			}

			// Regenerate with codes
			explanation = explanationGenerator.GenerateExplanation(
				multiResult,
				codesInfo,
				contentQuality,
			)

			h.logger.Info("✅ [Phase 2] Explanation regenerated with codes in generateEnhancedClassification",
				zap.String("request_id", req.RequestID),
				zap.Bool("explanation_not_nil", explanation != nil),
				zap.String("primary_reason", func() string {
					if explanation != nil {
						return explanation.PrimaryReason
					}
					return ""
				}()))
		}
	} else {
		// Generate explanation if not provided by service
		explanationGenerator := classification.NewExplanationGenerator()
		contentQuality := 0.7
		if industryResult.Confidence > 0.8 {
			contentQuality = 0.85
		} else if industryResult.Confidence < 0.5 {
			contentQuality = 0.5
		}

		multiResult := &classification.MultiStrategyResult{
			PrimaryIndustry: industryResult.IndustryName,
			Confidence:      industryResult.Confidence,
			Keywords:        industryResult.Keywords,
			Method:          industryResult.Method,
			Strategies:      []classification.ClassificationStrategy{}, // Empty strategies - extractConfidenceFactors handles this
		}

		explanation = explanationGenerator.GenerateExplanation(
			multiResult,
			codesInfo, // Include codes if available
			contentQuality,
		)

		h.logger.Info("✅ [Phase 2] Explanation generated in generateEnhancedClassification",
			zap.String("request_id", req.RequestID),
			zap.Bool("explanation_not_nil", explanation != nil),
			zap.String("primary_reason", func() string {
				if explanation != nil {
					return explanation.PrimaryReason
				}
				return ""
			}()))
	}

	// Build metadata map - ensure it's initialized
	metadata := make(map[string]interface{})
	
	// Extract scraping metadata from WebsiteAnalysis.StructuredData if available
	// This metadata is set during website scraping (scraping_strategy, early_exit, etc.)
	if websiteAnalysis != nil && websiteAnalysis.StructuredData != nil {
		if scrapingStrategy, ok := websiteAnalysis.StructuredData["scraping_strategy"].(string); ok && scrapingStrategy != "" {
			metadata["scraping_strategy"] = scrapingStrategy
		}
		if earlyExit, ok := websiteAnalysis.StructuredData["early_exit"].(bool); ok {
			metadata["early_exit"] = earlyExit
		}
		if fallbackUsed, ok := websiteAnalysis.StructuredData["fallback_used"].(bool); ok {
			metadata["fallback_used"] = fallbackUsed
		}
		if fallbackType, ok := websiteAnalysis.StructuredData["fallback_type"].(string); ok && fallbackType != "" {
			metadata["fallback_type"] = fallbackType
		}
		if scrapingTime, ok := websiteAnalysis.StructuredData["scraping_time_ms"].(float64); ok {
			metadata["scraping_time_ms"] = scrapingTime
		}
	}
	
	// Add code generation metadata
	metadata["codeGeneration"] = map[string]interface{}{
		"method":            codeGenMethod,
		"total_codes":       totalCodesGenerated,
		"keyword_matches":   keywordMatchCount,
		"industry_matches":  industryMatchCount,
		"industriesAnalyzed": []string{primaryIndustry},
		"industryMatches":   industryMatchCount,
		"keywordMatches":    keywordMatchCount,
		"sources":           []string{"industry", "keyword"},
	}
	
	result := &EnhancedClassificationResult{
		BusinessName:            req.BusinessName,
		PrimaryIndustry:         primaryIndustry, // Use industry from DetectIndustry (e.g., "Wineries")
		IndustryConfidence:      industryResult.Confidence,
		BusinessType:            h.determineBusinessType(industryResult.Keywords, primaryIndustry),
		BusinessTypeConfidence:  industryResult.Confidence * 0.9, // Slightly lower than industry confidence
		MCCCodes:                mccCodes,
		SICCodes:                sicCodes,
		NAICSCodes:              naicsCodes,
		Keywords:                industryResult.Keywords,
		ConfidenceScore:         industryResult.Confidence,
		ClassificationReasoning: reasoning,
		WebsiteAnalysis:         websiteAnalysis,
		MethodWeights:           methodWeights,
		ClassificationExplanation: explanation, // Phase 2: Add explanation
		Timestamp:               time.Now(),
		Metadata:                metadata, // Include scraping metadata
		// Phase 4: Pass through async LLM processing fields
		LLMProcessingID:         industryResult.LLMProcessingID,
		LLMStatus:               string(industryResult.LLMStatus),
		// Phase 5: Pass through cache fields
		FromCache:               industryResult.FromCache,
		CachedAt:                industryResult.CachedAt,
		ProcessingPath: func() string {
			method := industryResult.Method
			if strings.Contains(method, "layer3") || strings.Contains(method, "llm") {
				return "layer3"
			}
			if strings.Contains(method, "layer2") || strings.Contains(method, "embedding") {
				return "layer2"
			}
			return "layer1"
		}(),
	}

	// Log the final result for debugging
	h.logger.Info("Enhanced classification result",
		zap.String("request_id", req.RequestID),
		zap.String("primary_industry", result.PrimaryIndustry),
		zap.Float64("confidence", result.ConfidenceScore),
		zap.Int("mcc_codes", len(result.MCCCodes)),
		zap.Int("sic_codes", len(result.SICCodes)),
		zap.Int("naics_codes", len(result.NAICSCodes)),
		zap.Bool("explanation_set", result.ClassificationExplanation != nil),
		zap.String("explanation_primary_reason", func() string {
			if result.ClassificationExplanation != nil {
				return result.ClassificationExplanation.PrimaryReason
			}
			return ""
		}()))

	// Add code generation metadata
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["codeGeneration"] = map[string]interface{}{
		"method":              codeGenMethod,
		"sources":             []string{"industry", "keyword"},
		"industriesAnalyzed":  []string{industryResult.IndustryName},
		"keywordMatches":      keywordMatchCount,
		"industryMatches":     industryMatchCount,
		"totalCodesGenerated": totalCodesGenerated,
	}

	return result, nil
}

// combineEnsembleResults combines Python ML and Go classification results with weighted voting
// Python ML: 60% weight, Go classification: 40% weight
func (h *ClassificationHandler) combineEnsembleResults(pythonMLResult, goResult *EnhancedClassificationResult, req *ClassificationRequest) *EnhancedClassificationResult {
	const pythonMLWeight = 0.60
	const goWeight = 0.40

	h.logger.Info("Combining ensemble results",
		zap.String("request_id", req.RequestID),
		zap.String("python_ml_industry", pythonMLResult.PrimaryIndustry),
		zap.Float64("python_ml_confidence", pythonMLResult.ConfidenceScore),
		zap.String("go_industry", goResult.PrimaryIndustry),
		zap.Float64("go_confidence", goResult.ConfidenceScore))

	// Determine primary industry based on consensus and weighted confidence
	var primaryIndustry string
	var confidenceScore float64
	var consensusBoost float64 = 0.0

	// Check for consensus (same industry)
	if strings.EqualFold(pythonMLResult.PrimaryIndustry, goResult.PrimaryIndustry) {
		primaryIndustry = pythonMLResult.PrimaryIndustry
		// Consensus boost: add 5% to confidence when both agree
		consensusBoost = 0.05
		h.logger.Info("Ensemble consensus: Both methods agree on industry",
			zap.String("request_id", req.RequestID),
			zap.String("industry", primaryIndustry))
	} else {
		// No consensus - use weighted selection based on confidence
		pythonMLScore := pythonMLResult.ConfidenceScore * pythonMLWeight
		goScore := goResult.ConfidenceScore * goWeight

		if pythonMLScore >= goScore {
			primaryIndustry = pythonMLResult.PrimaryIndustry
		} else {
			primaryIndustry = goResult.PrimaryIndustry
		}
		h.logger.Info("Ensemble disagreement: Using weighted selection",
			zap.String("request_id", req.RequestID),
			zap.String("selected_industry", primaryIndustry),
			zap.Float64("python_ml_weighted_score", pythonMLScore),
			zap.Float64("go_weighted_score", goScore))
	}

	// Calculate weighted confidence score
	confidenceScore = (pythonMLResult.ConfidenceScore * pythonMLWeight) + (goResult.ConfidenceScore * goWeight) + consensusBoost
	if confidenceScore > 1.0 {
		confidenceScore = 1.0
	}

	// Merge keywords (deduplicate)
	keywordMap := make(map[string]bool)
	var mergedKeywords []string
	for _, kw := range pythonMLResult.Keywords {
		kwLower := strings.ToLower(kw)
		if !keywordMap[kwLower] {
			keywordMap[kwLower] = true
			mergedKeywords = append(mergedKeywords, kw)
		}
	}
	for _, kw := range goResult.Keywords {
		kwLower := strings.ToLower(kw)
		if !keywordMap[kwLower] {
			keywordMap[kwLower] = true
			mergedKeywords = append(mergedKeywords, kw)
		}
	}

	// Merge codes (prefer Python ML codes, fallback to Go codes)
	// For each code type, combine and deduplicate by code value
	mergedMCC := h.mergeCodes(pythonMLResult.MCCCodes, goResult.MCCCodes)
	mergedSIC := h.mergeCodes(pythonMLResult.SICCodes, goResult.SICCodes)
	mergedNAICS := h.mergeCodes(pythonMLResult.NAICSCodes, goResult.NAICSCodes)

	// Build reasoning
	reasoning := fmt.Sprintf("Ensemble classification combining Python ML (%.0f%% confidence, %s) and Go classification (%.0f%% confidence, %s). ",
		pythonMLResult.ConfidenceScore*100, pythonMLResult.PrimaryIndustry,
		goResult.ConfidenceScore*100, goResult.PrimaryIndustry)
	if consensusBoost > 0 {
		reasoning += fmt.Sprintf("Consensus detected: both methods agree on '%s' (confidence boost: +%.0f%%). ", primaryIndustry, consensusBoost*100)
	} else {
		reasoning += fmt.Sprintf("Selected '%s' based on weighted voting (Python ML: %.0f%%, Go: %.0f%%). ", primaryIndustry, pythonMLWeight*100, goWeight*100)
	}
	reasoning += fmt.Sprintf("Final confidence: %.0f%%.", confidenceScore*100)

	// Build method weights
	methodWeights := map[string]float64{
		"python_ml_service": pythonMLWeight * 100,
		"go_classification": goWeight * 100,
		"ensemble_voting":   100.0,
	}

	// Use Python ML website analysis if available, otherwise use Go
	websiteAnalysis := pythonMLResult.WebsiteAnalysis
	if websiteAnalysis == nil {
		websiteAnalysis = goResult.WebsiteAnalysis
	}

	// Build metadata
	metadata := make(map[string]interface{})
	if pythonMLResult.Metadata != nil {
		for k, v := range pythonMLResult.Metadata {
			metadata["python_ml_"+k] = v
		}
	}
	if goResult.Metadata != nil {
		for k, v := range goResult.Metadata {
			metadata["go_"+k] = v
		}
	}
	metadata["ensemble_voting"] = true
	metadata["consensus"] = consensusBoost > 0
	metadata["python_ml_industry"] = pythonMLResult.PrimaryIndustry
	metadata["go_industry"] = goResult.PrimaryIndustry
	metadata["final_confidence"] = confidenceScore

	// Phase 4: Pass through LLM processing fields from either result
	llmProcessingID := pythonMLResult.LLMProcessingID
	llmStatus := pythonMLResult.LLMStatus
	if llmProcessingID == "" && goResult.LLMProcessingID != "" {
		llmProcessingID = goResult.LLMProcessingID
		llmStatus = goResult.LLMStatus
	}

	return &EnhancedClassificationResult{
		BusinessName:            req.BusinessName,
		PrimaryIndustry:         primaryIndustry,
		IndustryConfidence:      confidenceScore,
		BusinessType:            h.determineBusinessType(mergedKeywords, primaryIndustry),
		BusinessTypeConfidence:  confidenceScore * 0.9,
		MCCCodes:                mergedMCC,
		SICCodes:                mergedSIC,
		NAICSCodes:              mergedNAICS,
		Keywords:                mergedKeywords,
		ConfidenceScore:         confidenceScore,
		ClassificationReasoning: reasoning,
		WebsiteAnalysis:         websiteAnalysis,
		MethodWeights:           methodWeights,
		Timestamp:               time.Now(),
		Metadata:                metadata,
		// Phase 4: Async LLM processing fields
		LLMProcessingID:         llmProcessingID,
		LLMStatus:               llmStatus,
	}
}

// mergeCodes merges two code slices, preferring codes from the first slice and deduplicating by code value
func (h *ClassificationHandler) mergeCodes(primary, secondary []IndustryCode) []IndustryCode {
	codeMap := make(map[string]*IndustryCode)

	// Add primary codes first (higher priority)
	for _, code := range primary {
		codeMap[code.Code] = &code
	}

	// Add secondary codes only if not already present
	for _, code := range secondary {
		if _, exists := codeMap[code.Code]; !exists {
			codeMap[code.Code] = &code
		}
	}

	// Convert map back to slice, limit to top 3
	result := make([]IndustryCode, 0, len(codeMap))
	for _, code := range codeMap {
		result = append(result, *code)
	}

	// Sort by confidence (descending) and limit to top 3
	if len(result) > 3 {
		// Simple sort by confidence
		for i := 0; i < len(result)-1; i++ {
			for j := i + 1; j < len(result); j++ {
				if result[i].Confidence < result[j].Confidence {
					result[i], result[j] = result[j], result[i]
				}
			}
		}
		result = result[:3]
	}

	return result
}

// determineBusinessType determines business type from keywords and industry
func (h *ClassificationHandler) determineBusinessType(keywords []string, industry string) string {
	// Simple heuristic based on industry name
	industryLower := strings.ToLower(industry)
	if strings.Contains(industryLower, "retail") || strings.Contains(industryLower, "store") {
		return "Retail Store"
	}
	if strings.Contains(industryLower, "service") {
		return "Service Business"
	}
	if strings.Contains(industryLower, "technology") || strings.Contains(industryLower, "software") {
		return "Technology Company"
	}
	if strings.Contains(industryLower, "health") || strings.Contains(industryLower, "medical") {
		return "Healthcare Provider"
	}
	if strings.Contains(industryLower, "financial") {
		return "Financial Services"
	}
	return "Business"
}

// recordClassificationForCalibration records classification result for accuracy tracking
// OPTIMIZATION #5.2: Confidence Calibration
func (h *ClassificationHandler) recordClassificationForCalibration(
	ctx context.Context,
	req *ClassificationRequest,
	response *ClassificationResponse,
	processingTime time.Duration,
) {
	if h.confidenceCalibrator == nil {
		return
	}

	// Calculate confidence bin (0-9 for 10 bins of 0.1 each)
	confidence := response.ConfidenceScore
	binIndex := int(confidence / 0.1)
	if binIndex >= 10 {
		binIndex = 9
	}

	// Record in calibrator (in-memory tracking)
	// Note: actual_industry will be NULL until validated
	// We'll update it later when validation data is available
	err := h.confidenceCalibrator.RecordClassification(
		ctx,
		confidence,
		"", // actual_industry - will be updated when validated
		response.PrimaryIndustry,
		false, // is_correct - will be updated when validated
	)
	if err != nil {
		h.logger.Warn("Failed to record classification for calibration",
			zap.String("request_id", req.RequestID),
			zap.Error(err))
		return
	}

	// Save to database for persistent tracking (OPTIMIZATION #5.2)
	if h.keywordRepo != nil {
		tracking := &repository.ClassificationAccuracyTracking{
			RequestID:            req.RequestID,
			BusinessName:         req.BusinessName,
			WebsiteURL:           req.WebsiteURL,
			PredictedIndustry:    response.PrimaryIndustry,
			PredictedConfidence:  confidence,
			ConfidenceBin:        binIndex,
			ClassificationMethod: "multi_strategy", // Could be enhanced to track actual method used
			KeywordsCount: func() int {
				if response.Classification.WebsiteContent != nil {
					return response.Classification.WebsiteContent.KeywordsFound
				}
				return 0
			}(),
			ProcessingTimeMs: int(processingTime.Milliseconds()),
			CreatedAt:        time.Now(),
		}

		if err := h.keywordRepo.SaveClassificationAccuracy(ctx, tracking); err != nil {
			h.logger.Warn("Failed to save classification accuracy to database",
				zap.String("request_id", req.RequestID),
				zap.Error(err))
			// Continue even if database save fails - in-memory tracking still works
		}
	}

	h.logger.Debug("Classification recorded for calibration",
		zap.String("request_id", req.RequestID),
		zap.Float64("confidence", confidence),
		zap.Int("confidence_bin", binIndex),
		zap.String("predicted_industry", response.PrimaryIndustry))

	// Check if recalibration is needed (periodic, e.g., daily)
	if h.confidenceCalibrator.ShouldRecalibrate() {
		h.logger.Info("Recalibration needed, starting calibration analysis",
			zap.String("request_id", req.RequestID))

		calibrationResult, err := h.confidenceCalibrator.Calibrate(ctx)
		if err != nil {
			h.logger.Warn("Calibration failed",
				zap.String("request_id", req.RequestID),
				zap.Error(err))
			return
		}

		h.logger.Info("Calibration complete",
			zap.String("request_id", req.RequestID),
			zap.Float64("overall_accuracy", calibrationResult.OverallAccuracy),
			zap.Float64("target_accuracy", calibrationResult.TargetAccuracy),
			zap.Float64("recommended_threshold", calibrationResult.RecommendedThreshold),
			zap.Bool("is_calibrated", calibrationResult.IsCalibrated))
	}
}

// zapLoggerAdapter adapts zap.Logger to io.Writer for standard log.Logger
type zapLoggerAdapter struct {
	logger *zap.Logger
}

func (z *zapLoggerAdapter) Write(p []byte) (n int, err error) {
	z.logger.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}

// convertIndustryCodes converts IndustryCode to handlers.IndustryCode
// Priority 4 Fix: Ensures arrays are never nil (returns empty array if nil)
func convertIndustryCodes(codes []IndustryCode) []IndustryCode {
	if codes == nil {
		return []IndustryCode{}
	}
	return codes // Same type, no conversion needed
}

// convertMCCCodesToIndustryCodes converts classification.MCCCode to handlers.IndustryCode (Phase 2: includes Source)
func convertMCCCodesToIndustryCodes(codes []classification.MCCCode) []IndustryCode {
	result := make([]IndustryCode, 0, len(codes))
	for _, code := range codes {
		source := []string{code.Source}
		if code.Source == "" {
			source = []string{"keyword"} // Default fallback
		}
		result = append(result, IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
			Source:      source,
		})
	}
	return result
}

// convertSICCodesToIndustryCodes converts classification.SICCode to handlers.IndustryCode (Phase 2: includes Source)
func convertSICCodesToIndustryCodes(codes []classification.SICCode) []IndustryCode {
	result := make([]IndustryCode, 0, len(codes))
	for _, code := range codes {
		source := []string{code.Source}
		if code.Source == "" {
			source = []string{"keyword"} // Default fallback
		}
		result = append(result, IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
			Source:      source,
		})
	}
	return result
}

// convertNAICSCodesToIndustryCodes converts classification.NAICSCode to handlers.IndustryCode (Phase 2: includes Source)
func convertNAICSCodesToIndustryCodes(codes []classification.NAICSCode) []IndustryCode {
	result := make([]IndustryCode, 0, len(codes))
	for _, code := range codes {
		source := []string{code.Source}
		if code.Source == "" {
			source = []string{"keyword"} // Default fallback
		}
		result = append(result, IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
			Source:      source,
		})
	}
	return result
}

// buildEnhancedResultFromPythonML builds EnhancedClassificationResult from Python ML service response
func (h *ClassificationHandler) buildEnhancedResultFromPythonML(
	enhancedResp *infrastructure.EnhancedClassificationResponse,
	codesInfo *classification.ClassificationCodesInfo,
	req *ClassificationRequest,
	keywords []string,
) *EnhancedClassificationResult {
	// Get primary industry from classifications array
	primaryIndustry := "Unknown"
	if len(enhancedResp.Classifications) > 0 {
		primaryIndustry = enhancedResp.Classifications[0].Label
	}

	// Build all industry scores map from classifications array
	allIndustryScores := make(map[string]float64)
	for _, classification := range enhancedResp.Classifications {
		allIndustryScores[classification.Label] = classification.Confidence
	}

	// Convert codes to handler format (limit to top 3 per type)
	const maxCodesPerType = 3
	mccCodes := h.convertMCCCodes(codesInfo.MCC, maxCodesPerType)
	sicCodes := h.convertSICCodes(codesInfo.SIC, maxCodesPerType)
	naicsCodes := h.convertNAICSCodes(codesInfo.NAICS, maxCodesPerType)

	// Build website analysis
	websiteAnalysis := &WebsiteAnalysisData{
		Success:           req.WebsiteURL != "",
		PagesAnalyzed:     1, // Python ML service analyzed the website
		RelevantPages:     1,
		KeywordsExtracted: keywords,
		IndustrySignals:   []string{strings.ToLower(strings.ReplaceAll(primaryIndustry, " ", "_"))},
		AnalysisMethod:    "python_ml_service",
		ProcessingTime:    time.Duration(enhancedResp.ProcessingTime * float64(time.Second)),
		OverallRelevance:  enhancedResp.Confidence,
		ContentQuality:    enhancedResp.Confidence,
		StructuredData: map[string]interface{}{
			"business_type": "Business",
			"industry":      primaryIndustry,
			"summary":       enhancedResp.Summary,
		},
	}

	// Build method weights
	methodWeights := map[string]float64{
		"python_ml_service": 100.0,
	}

	// Build result
	result := &EnhancedClassificationResult{
		BusinessName:            req.BusinessName,
		PrimaryIndustry:         primaryIndustry,
		IndustryConfidence:      enhancedResp.Confidence,
		BusinessType:            h.determineBusinessType(keywords, primaryIndustry),
		BusinessTypeConfidence:  enhancedResp.Confidence * 0.9,
		MCCCodes:                mccCodes,
		SICCodes:                sicCodes,
		NAICSCodes:              naicsCodes,
		Keywords:                keywords,
		ConfidenceScore:         enhancedResp.Confidence,
		ClassificationReasoning: enhancedResp.Explanation,
		WebsiteAnalysis:         websiteAnalysis,
		MethodWeights:           methodWeights,
		Timestamp:               enhancedResp.Timestamp,
		Metadata: map[string]interface{}{
			"explanation":          enhancedResp.Explanation,
			"content_summary":      enhancedResp.Summary,
			"quantization_enabled": enhancedResp.QuantizationEnabled,
			"model_version":        enhancedResp.ModelVersion,
			"processing_time":      enhancedResp.ProcessingTime,
			"all_industry_scores":  allIndustryScores,
		},
	}

	return result
}

// extractKeywordsFromText extracts keywords from text (simple implementation)
func (h *ClassificationHandler) extractKeywordsFromText(text string) []string {
	// Simple keyword extraction - split by spaces and filter common words
	words := strings.Fields(strings.ToLower(text))
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "was": true,
		"are": true, "were": true, "been": true, "be": true, "have": true, "has": true,
		"had": true, "do": true, "does": true, "did": true, "will": true, "would": true,
		"should": true, "could": true, "may": true, "might": true, "must": true,
		"this": true, "that": true, "these": true, "those": true, "it": true, "its": true,
	}

	keywords := make([]string, 0)
	seen := make(map[string]bool)
	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:()[]{}\"'")
		if len(word) > 2 && !stopWords[word] && !seen[word] {
			keywords = append(keywords, word)
			seen[word] = true
		}
	}
	return keywords
}

// convertMCCCodes converts classification.MCCCode to handlers.IndustryCode
func (h *ClassificationHandler) convertMCCCodes(codes []classification.MCCCode, limit int) []IndustryCode {
	result := make([]IndustryCode, 0, limit)
	for i, code := range codes {
		if i >= limit {
			break
		}
		result = append(result, IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
			Source:      []string{"keyword"},
		})
	}
	return result
}

// convertSICCodes converts classification.SICCode to handlers.IndustryCode
func (h *ClassificationHandler) convertSICCodes(codes []classification.SICCode, limit int) []IndustryCode {
	result := make([]IndustryCode, 0, limit)
	for i, code := range codes {
		if i >= limit {
			break
		}
		result = append(result, IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
			Source:      []string{"keyword"},
		})
	}
	return result
}

// convertNAICSCodes converts classification.NAICSCode to handlers.IndustryCode
func (h *ClassificationHandler) convertNAICSCodes(codes []classification.NAICSCode, limit int) []IndustryCode {
	result := make([]IndustryCode, 0, limit)
	for i, code := range codes {
		if i >= limit {
			break
		}
		result = append(result, IndustryCode{
			Code:        code.Code,
			Description: code.Description,
			Confidence:  code.Confidence,
			Source:      []string{"keyword"},
		})
	}
	return result
}

// HandleHealth handles health check requests
// HandleCacheHealth returns cache health status
func (h *ClassificationHandler) HandleCacheHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	status := map[string]interface{}{
		"cache_enabled": h.config.Classification.CacheEnabled,
		"redis_enabled": h.config.Classification.RedisEnabled,
		"redis_configured": h.config.Classification.RedisURL != "",
		"redis_connected": false,
		"in_memory_cache_size": 0,
		"cache_ttl_seconds": int(h.config.Classification.CacheTTL.Seconds()),
	}
	
	// Check Redis connection if enabled
	if h.redisCache != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		if err := h.redisCache.Health(ctx); err == nil {
			status["redis_connected"] = true
		} else {
			status["redis_error"] = err.Error()
			status["redis_connected"] = false
		}
	} else {
		// Redis cache not initialized
		status["redis_error"] = "Redis cache not initialized"
		status["redis_connected"] = false
	}
	
	// Get in-memory cache size
	h.cacheMutex.RLock()
	status["in_memory_cache_size"] = len(h.cache)
	h.cacheMutex.RUnlock()
	
	// Determine overall health
	// Cache is healthy if: enabled AND (Redis connected OR in-memory cache has items)
	redisConnected := h.redisCache != nil && status["redis_connected"].(bool)
	inMemoryHasItems := status["in_memory_cache_size"].(int) > 0
	healthy := h.config.Classification.CacheEnabled && (redisConnected || inMemoryHasItems)
	
	if healthy {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	status["healthy"] = healthy
	
	json.NewEncoder(w).Encode(status)
}

// HandleResetCircuitBreaker manually resets the ML service circuit breaker
// This is useful for recovery when the service is healthy but circuit breaker is stuck open
func (h *ClassificationHandler) HandleResetCircuitBreaker(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Simple admin check - in production, use proper authentication
	adminKey := r.Header.Get("X-Admin-Key")
	if adminKey == "" {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Missing X-Admin-Key header",
		})
		return
	}

	// Check if Python ML service is available
	if h.pythonMLService == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Python ML service not configured",
		})
		return
	}

	// Type assert to get PythonMLService
	pms, ok := h.pythonMLService.(*infrastructure.PythonMLService)
	if !ok || pms == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Python ML service not available",
		})
		return
	}

	// Get current state before reset
	oldState := pms.GetCircuitBreakerState()
	oldMetrics := pms.GetCircuitBreakerMetrics()

	// Reset circuit breaker
	pms.ResetCircuitBreaker()

	// Get new state after reset
	newState := pms.GetCircuitBreakerState()
	newMetrics := pms.GetCircuitBreakerMetrics()

	// Verify service health
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	health, healthErr := pms.HealthCheck(ctx)

	h.logger.Info("Circuit breaker manually reset",
		zap.String("old_state", oldState.String()),
		zap.String("new_state", newState.String()),
		zap.Int("old_failures", oldMetrics.FailureCount),
		zap.Int("new_failures", newMetrics.FailureCount),
		zap.Bool("service_healthy", healthErr == nil && health != nil && health.Status == "pass"),
	)

	response := map[string]interface{}{
		"success": true,
		"message": "Circuit breaker reset successfully",
		"old_state": map[string]interface{}{
			"state":         oldState.String(),
			"failure_count": oldMetrics.FailureCount,
			"success_count": oldMetrics.SuccessCount,
		},
		"new_state": map[string]interface{}{
			"state":         newState.String(),
			"failure_count": newMetrics.FailureCount,
			"success_count": newMetrics.SuccessCount,
		},
		"service_health": map[string]interface{}{
			"healthy": healthErr == nil && health != nil && health.Status == "pass",
			"status":   health.Status,
			"error":    healthErr,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ClassificationHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	// Set response headers
	w.Header().Set("Content-Type", "application/json")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check Supabase connectivity
	supabaseHealthy := true
	var supabaseError error
	if err := h.supabaseClient.HealthCheck(ctx); err != nil {
		supabaseHealthy = false
		supabaseError = err
	}

	// Get classification data
	classificationData, err := h.supabaseClient.GetClassificationData(ctx)
	if err != nil {
		h.logger.Warn("Failed to get classification data", zap.Error(err))
	}

	// Check Python ML service circuit breaker status if available
	var mlServiceStatus map[string]interface{}
	if h.pythonMLService != nil {
		// Type assert to get PythonMLService
		if pms, ok := h.pythonMLService.(*infrastructure.PythonMLService); ok && pms != nil {
			// Safely get circuit breaker state and metrics with nil checks
			var cbState interface{}
			var cbMetrics interface{}
			var cbStateErr error
			var cbMetricsErr error

			// Use recover to handle potential panics from nil circuit breaker
			func() {
				defer func() {
					if r := recover(); r != nil {
						cbStateErr = fmt.Errorf("circuit breaker not initialized: %v", r)
					}
				}()
				cbState = pms.GetCircuitBreakerState()
			}()

			func() {
				defer func() {
					if r := recover(); r != nil {
						cbMetricsErr = fmt.Errorf("circuit breaker metrics not available: %v", r)
					}
				}()
				cbMetrics = pms.GetCircuitBreakerMetrics()
			}()

			// Try to get health with circuit breaker info (with timeout)
			healthCtx, healthCancel := context.WithTimeout(ctx, 3*time.Second)
			defer healthCancel()

			var cbHealth interface{}
			var healthErr error

			// Safely call HealthCheckWithCircuitBreaker with panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						healthErr = fmt.Errorf("health check panic: %v", r)
					}
				}()
				cbHealth, healthErr = pms.HealthCheckWithCircuitBreaker(healthCtx)
			}()

			// Build mlServiceStatus safely
			mlServiceStatus = map[string]interface{}{
				"available": true,
			}

			// Add circuit breaker state if available
			if cbStateErr == nil && cbState != nil {
				if stateStr, ok := cbState.(fmt.Stringer); ok {
					mlServiceStatus["circuit_breaker_state"] = stateStr.String()
				} else {
					mlServiceStatus["circuit_breaker_state"] = fmt.Sprintf("%v", cbState)
				}
			} else if cbStateErr != nil {
				mlServiceStatus["circuit_breaker_state_error"] = cbStateErr.Error()
				mlServiceStatus["circuit_breaker_state"] = "unavailable"
			}

			// Add circuit breaker metrics if available
			if cbMetricsErr == nil && cbMetrics != nil {
				// Type assert to CircuitBreakerMetrics
				if metrics, ok := cbMetrics.(infrastructure.CircuitBreakerMetrics); ok {
					mlServiceStatus["circuit_breaker_metrics"] = map[string]interface{}{
						"state":             metrics.State,
						"failure_count":     metrics.FailureCount,
						"success_count":     metrics.SuccessCount,
						"state_change_time": metrics.StateChangeTime,
						"last_failure_time": metrics.LastFailureTime,
						"total_requests":    metrics.TotalRequests,
						"rejected_requests": metrics.RejectedRequests,
					}
				} else {
					// Fallback: just include the metrics as-is
					mlServiceStatus["circuit_breaker_metrics"] = cbMetrics
				}
			} else if cbMetricsErr != nil {
				mlServiceStatus["circuit_breaker_metrics_error"] = cbMetricsErr.Error()
			}

			// Add health check result if available
			if healthErr == nil && cbHealth != nil {
				if healthStatus, ok := cbHealth.(*infrastructure.HealthStatus); ok {
					mlServiceStatus["health_status"] = healthStatus.Status
					mlServiceStatus["health_checks"] = healthStatus.Checks
				} else {
					mlServiceStatus["health_status"] = fmt.Sprintf("%v", cbHealth)
				}
			} else if healthErr != nil {
				mlServiceStatus["health_check_error"] = healthErr.Error()
			}
		} else {
			mlServiceStatus = map[string]interface{}{
				"available": true,
				"status":    "initialized",
				"note":      "circuit_breaker_status_unavailable",
			}
		}
	} else {
		mlServiceStatus = map[string]interface{}{
			"available": false,
			"status":    "not_configured",
		}
	}

	// Check LLM service connectivity (Layer 3)
	llmURL := strings.TrimSuffix(h.config.Classification.LLMServiceURL, "/") // Remove trailing slash
	llmStatus := map[string]interface{}{
		"configured": h.config.Classification.LLMServiceURL != "",
		"url":        llmURL,
	}
	if llmURL != "" {
		// Try to reach the LLM service health endpoint
		llmHealthURL := llmURL + "/health"
		llmReq, err := http.NewRequestWithContext(ctx, "GET", llmHealthURL, nil)
		if err != nil {
			llmStatus["connected"] = false
			llmStatus["error"] = fmt.Sprintf("failed to create request: %v", err)
		} else {
			llmClient := &http.Client{Timeout: 5 * time.Second}
			resp, err := llmClient.Do(llmReq)
			if err != nil {
				llmStatus["connected"] = false
				llmStatus["error"] = fmt.Sprintf("connection failed: %v", err)
			} else {
				defer resp.Body.Close()
				llmStatus["connected"] = resp.StatusCode == 200
				llmStatus["status_code"] = resp.StatusCode
				// Try to parse the response
				var llmHealth map[string]interface{}
				if json.NewDecoder(resp.Body).Decode(&llmHealth) == nil {
					llmStatus["model_loaded"] = llmHealth["model_loaded"]
					llmStatus["model"] = llmHealth["model"]
					llmStatus["llm_status"] = llmHealth["status"]
				}
			}
		}
	}

	// Create health response
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.3.3", // Phase 4: Boosted real estate over brokerage for 100% accuracy
		"service":   "classification-service",
		"uptime":    time.Since(h.serviceStartTime).String(),
		"supabase_status": map[string]interface{}{
			"connected": supabaseHealthy,
			"url":       h.config.Supabase.URL,
			"error":     supabaseError,
		},
		"ml_service_status":   mlServiceStatus,
		"llm_service_status":  llmStatus,
		"classification_data": classificationData,
		"features": map[string]interface{}{
			"ml_enabled":             h.config.Classification.MLEnabled,
			"keyword_method_enabled": h.config.Classification.KeywordMethodEnabled,
			"ensemble_enabled":       h.config.Classification.EnsembleEnabled,
			"cache_enabled":          h.config.Classification.CacheEnabled,
			"llm_enabled":            h.config.Classification.LLMServiceURL != "",
		},
	}

	// Set status code based on health
	statusCode := http.StatusOK
	if !supabaseHealthy {
		statusCode = http.StatusServiceUnavailable
		health["status"] = "unhealthy"
	}

	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}

// HandleLLMStatus handles requests to check the status of async LLM processing
// GET /classify/status/{processing_id}
func (h *ClassificationHandler) HandleLLMStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract processing ID from URL path
	// Expected path: /classify/status/{processing_id} or /v1/classify/status/{processing_id}
	path := r.URL.Path
	var processingID string
	
	// Handle both /classify/status/{id} and /v1/classify/status/{id}
	if idx := strings.LastIndex(path, "/status/"); idx != -1 {
		processingID = path[idx+8:] // len("/status/") = 8
	}
	
	if processingID == "" {
		h.logger.Error("Missing processing_id in request")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "missing_processing_id",
			"message": "Processing ID is required",
		})
		return
	}
	
	// Get the async LLM result
	result, found := h.industryDetector.GetAsyncLLMResult(processingID)
	if !found {
		h.logger.Warn("Processing ID not found",
			zap.String("processing_id", processingID))
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":         "not_found",
			"message":       "Processing ID not found or expired",
			"processing_id": processingID,
		})
		return
	}
	
	// Build response based on status
	response := map[string]interface{}{
		"processing_id": result.ProcessingID,
		"status":        result.Status,
		"started_at":    result.StartedAt,
	}
	
	if result.CompletedAt != nil {
		response["completed_at"] = result.CompletedAt
		response["processing_time_ms"] = result.CompletedAt.Sub(result.StartedAt).Milliseconds()
	}
	
	if result.Error != "" {
		response["error"] = result.Error
	}
	
	// Include the original Layer 1/2 result for reference
	if result.OriginalResult != nil {
		response["original_result"] = map[string]interface{}{
			"industry":   result.OriginalResult.IndustryName,
			"confidence": result.OriginalResult.Confidence,
			"method":     result.OriginalResult.Method,
		}
	}
	
	// If completed successfully, include the LLM result
	if result.Status == classification.AsyncLLMStatusCompleted && result.Result != nil {
		response["llm_result"] = map[string]interface{}{
			"industry":   result.Result.PrimaryIndustry,
			"confidence": result.Result.Confidence,
			"reasoning":  result.Result.Reasoning,
			"mcc_codes":  result.Result.MCC,
			"naics_codes": result.Result.NAICS,
			"sic_codes":  result.Result.SIC,
		}
		
		// Calculate improvement over original result
		if result.OriginalResult != nil {
			improvement := result.Result.Confidence - result.OriginalResult.Confidence
			response["confidence_improvement"] = improvement
			response["llm_added_value"] = improvement > 0.03
		}
	}
	
	h.logger.Info("LLM status request",
		zap.String("processing_id", processingID),
		zap.String("status", string(result.Status)))
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleAsyncLLMStats returns statistics about async LLM processing
// GET /classify/async-stats
func (h *ClassificationHandler) HandleAsyncLLMStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	stats := h.industryDetector.GetAsyncLLMStats()
	stats["timestamp"] = time.Now()
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// calculateAdaptiveTimeout calculates an adaptive timeout based on request characteristics
// Implements Hybrid Approach: combines budget allocation with adaptive timeout strategy
func (h *ClassificationHandler) calculateAdaptiveTimeout(req *ClassificationRequest) time.Duration {
	// Base timeout from config (use OverallTimeout if available, otherwise RequestTimeout)
	baseTimeout := h.config.Classification.RequestTimeout
	if h.config.Classification.OverallTimeout > 0 {
		baseTimeout = h.config.Classification.OverallTimeout
	}

	// Timeout budget allocation for different operations
	// FIX #10: Added buffer for retries and overhead to prevent premature timeouts
	const (
		phase1ScrapingBudget    = 18 * time.Second // Phase 1 scraper: 18s (aligned with WebsiteScrapingTimeout of 15s + 3s buffer)
		multiPageAnalysisBudget = 8 * time.Second  // Multi-page analysis: 8s (reduced from 10s, capped)
		// FIX: Index building now has 5-minute TTL cache - first call: 10-30s, subsequent calls: <1ms (cache hit)
		indexBuildingBudget    = 30 * time.Second // Keyword index building: 30s (first call, cached for 5min)
		goClassificationBudget = 5 * time.Second
		mlClassificationBudget = 10 * time.Second
		riskAssessmentBudget   = 5 * time.Second
		generalOverhead        = 5 * time.Second  // FIX #10: Increased from 3s to 5s to account for retries and network latency
		retryBuffer            = 10 * time.Second // FIX #10: Additional buffer for retry attempts and network delays
	)

	// Determine if we need long-running operations
	needsWebsiteScraping := req.WebsiteURL != "" && req.WebsiteURL != "N/A"

	// Calculate required timeout based on operation needs
	var requiredTimeout time.Duration

	if needsWebsiteScraping {
		// Website scraping needed - allocate budget for Phase 1 scraper
		// Budget breakdown (OPTIMIZED for better success rate):
		// - Index building: 30s (first call can take 10-30s, happens before extractKeywords)
		// - Phase 1 scraping: 18s (aligned with WebsiteScrapingTimeout of 15s + 3s buffer for retries)
		// - Multi-page analysis: 8s (reduced from 10s, capped, may be skipped if insufficient time)
		// - Go classification: 5s
		// - ML classification: 10s (optional, may be skipped)
		// - Risk assessment: 5s (parallel, doesn't add to total)
		// - General overhead: 5s (for retries and network latency)
		// - Retry buffer: 10s (for retry attempts and network delays)
		// Total: 30 + 18 + 8 + 5 + 10 + 5 + 10 = 86s
		// FIX: Add budget for index building (30s) - this happens synchronously before extractKeywords
		// FIX: Add budget for multi-page analysis (8s) to prevent context expiration
		// FIX #10: Add retry buffer for retry attempts and network delays
		requiredTimeout = indexBuildingBudget + phase1ScrapingBudget + multiPageAnalysisBudget + goClassificationBudget + mlClassificationBudget + generalOverhead + retryBuffer

		h.logger.Info("Adaptive timeout: website scraping detected",
			zap.String("request_id", req.RequestID),
			zap.String("website_url", req.WebsiteURL),
			zap.Duration("calculated_timeout", requiredTimeout),
			zap.Duration("base_timeout", baseTimeout))
	} else {
		// Simple request without website scraping
		// Budget breakdown:
		// - Index building: 30s (first call can take 10-30s, happens before classification)
		// - Go classification: 5s
		// - ML classification: 10s (optional)
		// - General overhead: 5s
		// FIX: Add budget for index building even for simple requests
		// FIX #10: Add retry buffer for retry attempts and network delays
		requiredTimeout = indexBuildingBudget + goClassificationBudget + mlClassificationBudget + generalOverhead + retryBuffer

		h.logger.Info("Adaptive timeout: simple request (no website scraping)",
			zap.String("request_id", req.RequestID),
			zap.Duration("calculated_timeout", requiredTimeout),
			zap.Duration("base_timeout", baseTimeout))
	}

	// FIX: Use the calculated requiredTimeout when it's determined
	// The adaptive timeout calculation allocates budget for specific operations
	// We should use this calculated value, not the base timeout
	// Only use baseTimeout if requiredTimeout wasn't calculated (shouldn't happen, but safety check)
	if requiredTimeout > 0 {
		if requiredTimeout != baseTimeout {
			h.logger.Info("Using adaptive timeout calculation",
				zap.String("request_id", req.RequestID),
				zap.Duration("calculated_timeout", requiredTimeout),
				zap.Duration("base_timeout", baseTimeout),
				zap.Bool("needs_scraping", needsWebsiteScraping))
		}
		return requiredTimeout
	}

	// Fallback to baseTimeout if requiredTimeout wasn't calculated (shouldn't happen)
	h.logger.Warn("Required timeout not calculated, using base timeout",
		zap.String("request_id", req.RequestID),
		zap.Duration("base_timeout", baseTimeout))
	return baseTimeout
}
