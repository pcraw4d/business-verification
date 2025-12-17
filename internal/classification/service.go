package classification

import (
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/classification/repository"
	"kyb-platform/internal/external"
	"kyb-platform/internal/machine_learning"
)

// ClassificationMetrics tracks classification accuracy and performance metrics
type ClassificationMetrics struct {
	TotalClassifications    int64              `json:"total_classifications"`
	MLClassifications      int64              `json:"ml_classifications"`
	KeywordClassifications int64              `json:"keyword_classifications"`
	FallbackClassifications int64             `json:"fallback_classifications"`
	IndustryAccuracy       map[string]float64 `json:"industry_accuracy"` // Industry name -> accuracy percentage
	MethodAccuracy         map[string]float64 `json:"method_accuracy"`   // Method name -> accuracy percentage
	IndustryCorrect        map[string]int64   `json:"industry_correct"`  // Industry name -> correct count
	IndustryTotal          map[string]int64   `json:"industry_total"`    // Industry name -> total count
	MethodCorrect          map[string]int64   `json:"method_correct"`     // Method name -> correct count
	MethodTotal            map[string]int64   `json:"method_total"`      // Method name -> total count
	mu                     sync.RWMutex       `json:"-"`
}

// NewClassificationMetrics creates a new classification metrics tracker
func NewClassificationMetrics() *ClassificationMetrics {
	return &ClassificationMetrics{
		IndustryAccuracy:       make(map[string]float64),
		MethodAccuracy:         make(map[string]float64),
		IndustryCorrect:        make(map[string]int64),
		IndustryTotal:          make(map[string]int64),
		MethodCorrect:         make(map[string]int64),
		MethodTotal:           make(map[string]int64),
	}
}

// inFlightRequest tracks an in-flight classification request for deduplication
type inFlightRequest struct {
	resultChan chan *IndustryDetectionResult
	errChan    chan error
	done       bool
	mu         sync.Mutex
}

// IndustryDetectionService provides database-driven industry classification
type IndustryDetectionService struct {
	repo                 repository.KeywordRepository
	logger               *log.Logger
	monitor              *ClassificationAccuracyMonitoring
	multiStrategyClassifier *MultiStrategyClassifier
	mlClassifier           *machine_learning.ContentClassifier  // Optional: for ML support
	pythonMLService        interface{}                          // Optional: *infrastructure.PythonMLService - using interface to avoid import cycle
	useML                  bool                                 // Flag to enable ML
	metrics                *ClassificationMetrics               // Classification accuracy metrics
	inFlightRequests       sync.Map                             // map[string]*inFlightRequest - for request deduplication
	confidenceCalibrator   *ConfidenceCalibrator                 // Phase 2: Confidence calibration
	explanationGenerator   *ExplanationGenerator                 // Phase 2: Explanation generation
	embeddingClassifier    *EmbeddingClassifier                 // Phase 3: Embedding-based classification
	llmClassifier          *LLMClassifier                        // Phase 4: LLM-based classification
	asyncLLMProcessor      *AsyncLLMProcessor                    // Phase 4: Async LLM processing
}

// NewIndustryDetectionService creates a new industry detection service
func NewIndustryDetectionService(repo repository.KeywordRepository, logger *log.Logger) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	return &IndustryDetectionService{
		repo:                 repo,
		logger:               logger,
		monitor:              nil, // Will be set separately if monitoring is needed
		multiStrategyClassifier: NewMultiStrategyClassifier(repo, logger),
		mlClassifier:           nil,
		pythonMLService:        nil,
		useML:                  false,
		metrics:                NewClassificationMetrics(),
		confidenceCalibrator:   NewConfidenceCalibrator(logger), // Phase 2: Initialize confidence calibrator
		explanationGenerator:   NewExplanationGenerator(),       // Phase 2: Initialize explanation generator
	}
}

// SetContentCache sets the website content cache (deprecated - kept for backward compatibility)
func (s *IndustryDetectionService) SetContentCache(cache WebsiteContentCacher) {
	// Cache is now handled by request-scoped cache in methods
	s.logger.Printf("ℹ️ SetContentCache called (cache now handled by request-scoped cache)")
}

// NewIndustryDetectionServiceWithML creates a new industry detection service with ML support
func NewIndustryDetectionServiceWithML(
	repo repository.KeywordRepository,
	mlClassifier *machine_learning.ContentClassifier,
	pythonMLService interface{}, // *infrastructure.PythonMLService - using interface to avoid import cycle
	logger *log.Logger,
) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	svc := &IndustryDetectionService{
		repo:                 repo,
		logger:               logger,
		monitor:              nil,
		multiStrategyClassifier: NewMultiStrategyClassifier(repo, logger),
		mlClassifier:           mlClassifier,
		pythonMLService:        pythonMLService,
		useML:                  true,
		metrics:                NewClassificationMetrics(),
		confidenceCalibrator:   NewConfidenceCalibrator(logger), // Phase 2: Initialize confidence calibrator
		explanationGenerator:   NewExplanationGenerator(),       // Phase 2: Initialize explanation generator
	}

	if pythonMLService != nil {
		logger.Printf("✅ IndustryDetectionService initialized with ML support (Python ML service enabled)")
	} else {
		logger.Printf("✅ IndustryDetectionService initialized with ML support (Go ML classifier only)")
	}

	return svc
}

// NewIndustryDetectionServiceWithMonitoring creates a new industry detection service with monitoring
func NewIndustryDetectionServiceWithMonitoring(repo repository.KeywordRepository, logger *log.Logger, monitor *ClassificationAccuracyMonitoring) *IndustryDetectionService {
	if logger == nil {
		logger = log.Default()
	}

	return &IndustryDetectionService{
		repo:                 repo,
		logger:               logger,
		monitor:              monitor,
		multiStrategyClassifier: NewMultiStrategyClassifier(repo, logger),
		metrics:                NewClassificationMetrics(),
	}
}

// RecordClassification records a classification result for accuracy tracking
func (s *IndustryDetectionService) RecordClassification(
	result *IndustryDetectionResult,
	expectedIndustry string,
) {
	if s.metrics == nil {
		return
	}

	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	s.metrics.TotalClassifications++

	// Track by method
	switch result.Method {
	case "ml_distilbart", "ml", "ml_fallback":
		s.metrics.MLClassifications++
	case "keyword", "keyword_classification":
		s.metrics.KeywordClassifications++
	default:
		s.metrics.FallbackClassifications++
	}

	// Track accuracy
	isCorrect := strings.EqualFold(result.IndustryName, expectedIndustry)
	
	// Update industry metrics
	s.metrics.IndustryTotal[expectedIndustry]++
	if isCorrect {
		s.metrics.IndustryCorrect[expectedIndustry]++
	}
	
	// Calculate industry accuracy
	if total := s.metrics.IndustryTotal[expectedIndustry]; total > 0 {
		s.metrics.IndustryAccuracy[expectedIndustry] = float64(s.metrics.IndustryCorrect[expectedIndustry]) / float64(total) * 100.0
	}

	// Update method metrics
	s.metrics.MethodTotal[result.Method]++
	if isCorrect {
		s.metrics.MethodCorrect[result.Method]++
	}
	
	// Calculate method accuracy
	if total := s.metrics.MethodTotal[result.Method]; total > 0 {
		s.metrics.MethodAccuracy[result.Method] = float64(s.metrics.MethodCorrect[result.Method]) / float64(total) * 100.0
	}
}

// GetClassificationMetrics returns the current classification metrics
func (s *IndustryDetectionService) GetClassificationMetrics() *ClassificationMetrics {
	if s.metrics == nil {
		return NewClassificationMetrics()
	}

	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()

	// Return a copy of metrics
	metricsCopy := &ClassificationMetrics{
		TotalClassifications:     s.metrics.TotalClassifications,
		MLClassifications:         s.metrics.MLClassifications,
		KeywordClassifications:   s.metrics.KeywordClassifications,
		FallbackClassifications:   s.metrics.FallbackClassifications,
		IndustryAccuracy:          make(map[string]float64),
		MethodAccuracy:           make(map[string]float64),
		IndustryCorrect:          make(map[string]int64),
		IndustryTotal:            make(map[string]int64),
		MethodCorrect:           make(map[string]int64),
		MethodTotal:             make(map[string]int64),
	}

	// Copy maps
	for k, v := range s.metrics.IndustryAccuracy {
		metricsCopy.IndustryAccuracy[k] = v
	}
	for k, v := range s.metrics.MethodAccuracy {
		metricsCopy.MethodAccuracy[k] = v
	}
	for k, v := range s.metrics.IndustryCorrect {
		metricsCopy.IndustryCorrect[k] = v
	}
	for k, v := range s.metrics.IndustryTotal {
		metricsCopy.IndustryTotal[k] = v
	}
	for k, v := range s.metrics.MethodCorrect {
		metricsCopy.MethodCorrect[k] = v
	}
	for k, v := range s.metrics.MethodTotal {
		metricsCopy.MethodTotal[k] = v
	}

	return metricsCopy
}

// IndustryDetectionResult represents the result of industry detection
type IndustryDetectionResult struct {
	IndustryName   string                   `json:"industry_name"`
	Confidence     float64                  `json:"confidence"`
	Keywords       []string                 `json:"keywords"`
	ProcessingTime time.Duration            `json:"processing_time"`
	Method         string                   `json:"method"`
	Reasoning      string                   `json:"reasoning"`
	Explanation    *ClassificationExplanation `json:"explanation,omitempty"` // Phase 2: Structured explanation
	CreatedAt      time.Time                `json:"created_at"`
	
	// Phase 4: Async LLM processing fields
	LLMProcessingID string         `json:"llm_processing_id,omitempty"` // ID to poll for LLM result
	LLMStatus       AsyncLLMStatus `json:"llm_status,omitempty"`        // Status of async LLM processing
}

// normalizeString normalizes a string for cache key generation
func normalizeString(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// DetectIndustry performs database-driven industry detection using multi-strategy classification
// Phase 3.1: Enhanced with three-tier confidence-based ML strategy
// Fix: Added request deduplication to prevent duplicate processing
func (s *IndustryDetectionService) DetectIndustry(ctx context.Context, businessName, description, websiteURL string) (*IndustryDetectionResult, error) {
	// Generate cache key for deduplication
	cacheKey := fmt.Sprintf("%s|%s|%s", normalizeString(businessName), normalizeString(description), normalizeString(websiteURL))
	
	// Check for in-flight request
	if existing, found := s.inFlightRequests.Load(cacheKey); found {
		req := existing.(*inFlightRequest)
		req.mu.Lock()
		if !req.done {
			req.mu.Unlock()
			// Wait for existing request
			s.logger.Printf("♻️ [Deduplication] Reusing in-flight request for: %s", businessName)
			select {
			case result := <-req.resultChan:
				return result, nil
			case err := <-req.errChan:
				return nil, err
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		req.mu.Unlock()
	}
	
	// Create new in-flight request
	resultChan := make(chan *IndustryDetectionResult, 1)
	errChan := make(chan error, 1)
	inFlight := &inFlightRequest{
		resultChan: resultChan,
		errChan:    errChan,
		done:       false,
	}
	s.inFlightRequests.Store(cacheKey, inFlight)
	
	// Perform classification in goroutine
	go func() {
		result, err := s.performClassification(ctx, businessName, description, websiteURL)
		inFlight.mu.Lock()
		inFlight.done = true
		inFlight.mu.Unlock()
		if err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
		// Clean up after a delay to allow concurrent requests to read
		time.AfterFunc(5*time.Second, func() {
			s.inFlightRequests.Delete(cacheKey)
		})
	}()
	
	// Wait for result
	select {
	case result := <-resultChan:
		return result, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// performClassification performs the actual classification (moved from DetectIndustry)
func (s *IndustryDetectionService) performClassification(ctx context.Context, businessName, description, websiteURL string) (*IndustryDetectionResult, error) {
	startTime := time.Now()
	requestID := s.generateRequestID()

	// Debug: Log layer availability
	llmAvailable := s.llmClassifier != nil
	embeddingAvailable := s.embeddingClassifier != nil
	s.logger.Printf("🔍 Starting industry detection for: %s (request: %s) [LLM:%v, Embedding:%v, URL:%s]",
		businessName, requestID, llmAvailable, embeddingAvailable, websiteURL)

	// Step 1: Run multi-strategy classifier (base classification)
	s.logger.Printf("📝 Running MultiStrategyClassifier (base classification) (request: %s)", requestID)
	multiResult, err := s.multiStrategyClassifier.ClassifyWithMultiStrategy(
		ctx, businessName, description, websiteURL)
	if err != nil {
		// Fallback to keyword-based classification if multi-strategy fails
		s.logger.Printf("⚠️ Multi-strategy classification failed, falling back to keyword-based: %v", err)
		return s.fallbackToKeywordClassification(ctx, businessName, description, websiteURL, startTime, requestID)
	}

	if multiResult == nil {
		s.logger.Printf("⚠️ Multi-strategy returned nil, falling back to keyword-based")
		return s.fallbackToKeywordClassification(ctx, businessName, description, websiteURL, startTime, requestID)
	}

	// Phase 2: Use method from multiResult if available, otherwise default to "multi_strategy"
	resultMethod := multiResult.Method
	if resultMethod == "" {
		resultMethod = "multi_strategy"
	}
	s.logger.Printf("📊 [Phase 2] Classification method: %s (request: %s)", resultMethod, requestID)

	// Phase 2: Apply confidence calibration
	// Extract strategy scores from strategies
	strategyScores := make(map[string]float64)
	for _, strategy := range multiResult.Strategies {
		strategyScores[strategy.StrategyName] = strategy.Score
	}

	// For now, we'll use a default content quality (can be enhanced later with actual scraping quality)
	contentQualityForCalibration := 0.7 // Default, can be improved when we have actual content quality data
	if multiResult.Confidence > 0.8 {
		contentQualityForCalibration = 0.85 // High confidence suggests good content
	} else if multiResult.Confidence < 0.5 {
		contentQualityForCalibration = 0.5 // Low confidence suggests poor content
	}

	// Calibrate confidence (code agreement will be calculated later when codes are generated)
	codeAgreement := 0.7 // Default, will be updated when codes are available
	calibratedConfidence := s.confidenceCalibrator.CalibrateConfidence(
		strategyScores,
		contentQualityForCalibration,
		codeAgreement,
		"multi_strategy",
	)

	s.logger.Printf("📊 [Phase 2] Confidence calibration: %.2f%% -> %.2f%% (request: %s)",
		multiResult.Confidence*100, calibratedConfidence*100, requestID)

	// Update multiResult with calibrated confidence
	multiResult.Confidence = calibratedConfidence

	// Phase 3: Layer 2 routing - Try embeddings if Layer 1 confidence is low
	// Decision: Use Layer 1 or try Layer 2?
	const layer2Threshold = 0.80
	const highConfidenceThreshold = 0.95 // Increased from 0.90 to allow more cases to try Layer 2/3

	// Check for ambiguity indicators - ambiguous cases should use Layer 3 even with high confidence
	isAmbiguous := s.isAmbiguousCase(businessName, description)
	s.logger.Printf("📊 [Routing] Ambiguous: %v, LLM available: %v, Website: %v (request: %s)",
		isAmbiguous, s.llmClassifier != nil, websiteURL != "", requestID)

	// Phase 4: If ambiguous, trigger ASYNC Layer 3 (LLM) and return Layer 1 result immediately
	if isAmbiguous && s.llmClassifier != nil && s.asyncLLMProcessor != nil && websiteURL != "" {
		s.logger.Printf("🤖 [Phase 4] Ambiguous case detected, triggering ASYNC Layer 3 (LLM) (confidence: %.2f%%) (request: %s)",
			multiResult.Confidence*100, requestID)

		// Get ScrapedContent for Layer 3
		scrapedContent, err := s.getScrapedContentForLayer2(ctx, websiteURL)
		if err != nil {
			s.logger.Printf("⚠️ [Phase 4] Failed to get scraped content for Layer 3: %v, returning Layer 1 without async LLM (request: %s)", err, requestID)
			result := s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
			result.Method = "layer1_no_content"
			return result, nil
		}

		// Generate processing ID for async tracking
		llmProcessingID := fmt.Sprintf("llm_%s_%d", requestID, time.Now().UnixNano())
		
		// Convert Layer 1 result for the response
		result := s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
		result.Method = "layer1_llm_pending"
		result.LLMProcessingID = llmProcessingID
		result.LLMStatus = AsyncLLMStatusProcessing
		
		// Start async LLM processing (returns immediately)
		s.asyncLLMProcessor.ProcessAsync(
			llmProcessingID,
			ctx, // Note: We pass ctx but the processor creates its own context for the LLM call
			scrapedContent,
			businessName,
			description,
			multiResult,
			nil, // Layer 2 not available yet
			result,
		)
		
		s.logger.Printf("🚀 [Phase 4] Async LLM processing started (id: %s), returning Layer 1 result immediately (request: %s)",
			llmProcessingID, requestID)
		
		return result, nil
	}

	if multiResult.Confidence >= highConfidenceThreshold && !isAmbiguous {
		// Very high confidence AND not ambiguous - use Layer 1
		s.logger.Printf("✅ [Phase 3] Very high confidence (%.2f%%) >= 95%% and not ambiguous, using Layer 1 result (request: %s)",
			multiResult.Confidence*100, requestID)
		result := s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
		result.Method = "layer1"
		return result, nil
	}

	if multiResult.Confidence >= layer2Threshold {
		// Good confidence - use Layer 1
		s.logger.Printf("✅ [Phase 3] Good confidence (%.2f%%) >= 80%%, using Layer 1 result (request: %s)",
			multiResult.Confidence*100, requestID)
		result := s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
		result.Method = "layer1_medium_conf"
		return result, nil
	}

	// Lower confidence (<0.80) - try Layer 2 (Embeddings) if available
	if s.embeddingClassifier != nil && websiteURL != "" {
		s.logger.Printf("🔍 [Phase 3] Layer 1 confidence (%.2f%%) < 80%%, trying Layer 2 (Embeddings) (request: %s)",
			multiResult.Confidence*100, requestID)

		// Get ScrapedContent for Layer 2
		scrapedContent, err := s.getScrapedContentForLayer2(ctx, websiteURL)
		if err != nil {
			s.logger.Printf("⚠️ [Phase 3] Failed to get scraped content for Layer 2: %v, falling back to Layer 1 (request: %s)", err, requestID)
			result := s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
			result.Method = "layer1_fallback"
			return result, nil
		}

		// Try Layer 2 classification
		layer2Result, err := s.embeddingClassifier.ClassifyByEmbedding(ctx, scrapedContent)
		if err != nil {
			s.logger.Printf("⚠️ [Phase 3] Layer 2 classification failed: %v, falling back to Layer 1 (request: %s)", err, requestID)
			result := s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
			result.Method = "layer1_fallback"
			return result, nil
		}

		s.logger.Printf("✅ [Phase 3] Layer 2 complete (confidence: %.2f%%, top_match: %s) (request: %s)",
			layer2Result.Confidence*100, layer2Result.TopMatch, requestID)

		// Compare Layer 1 vs Layer 2
		if layer2Result.Confidence > multiResult.Confidence+0.05 {
			// Layer 2 is meaningfully better
			s.logger.Printf("✅ [Phase 3] Using Layer 2 result (Layer 2: %.2f%% vs Layer 1: %.2f%%) (request: %s)",
				layer2Result.Confidence*100, multiResult.Confidence*100, requestID)
			result := s.buildResultFromEmbedding(layer2Result, "layer2_embedding", requestID)
			return result, nil
		} else {
			// Layer 1 and Layer 2 similar - try async Layer 3 if confidence still low
			s.logger.Printf("ℹ️ [Phase 3] Layer 1 and Layer 2 similar (Layer 1: %.2f%%, Layer 2: %.2f%%) (request: %s)",
				multiResult.Confidence*100, layer2Result.Confidence*100, requestID)
			
			// Phase 4: Trigger ASYNC Layer 3 if confidence is still below threshold
			if layer2Result.Confidence < 0.88 && s.llmClassifier != nil && s.asyncLLMProcessor != nil && websiteURL != "" {
				s.logger.Printf("🤖 [Phase 4] Layer 2 confidence (%.2f%%) < 88%%, triggering ASYNC Layer 3 (LLM) (request: %s)",
					layer2Result.Confidence*100, requestID)
				
				// Get ScrapedContent for Layer 3
				scrapedContent, err := s.getScrapedContentForLayer2(ctx, websiteURL)
				if err != nil {
					s.logger.Printf("⚠️ [Phase 4] Failed to get scraped content for Layer 3: %v, using Layer 2 (request: %s)", err, requestID)
					result := s.buildResultFromEmbedding(layer2Result, "layer2_no_content", requestID)
					return result, nil
				}
				
				// Generate processing ID for async tracking
				llmProcessingID := fmt.Sprintf("llm_%s_%d", requestID, time.Now().UnixNano())
				
				// Use Layer 2 as the immediate result
				result := s.buildResultFromEmbedding(layer2Result, "layer2_llm_pending", requestID)
				result.LLMProcessingID = llmProcessingID
				result.LLMStatus = AsyncLLMStatusProcessing
				
				// Start async LLM processing (returns immediately)
				s.asyncLLMProcessor.ProcessAsync(
					llmProcessingID,
					ctx,
					scrapedContent,
					businessName,
					description,
					multiResult,
					layer2Result,
					result,
				)
				
				s.logger.Printf("🚀 [Phase 4] Async LLM processing started (id: %s), returning Layer 2 result immediately (request: %s)",
					llmProcessingID, requestID)
				
				return result, nil
			}
			
			// Use Layer 1 (Layer 2 similar, Layer 3 not available or didn't help)
			result := s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
			result.Method = "layer1_validated"
			return result, nil
		}
	}

	// Layer 2 not available or no website URL - continue with existing ML logic
	// Step 2: Determine ML strategy based on confidence level
	// Phase 3.1: Three-tier confidence-based ML strategy
	const (
		lowConfidenceThreshold     = 0.5
		mlHighConfidenceThreshold = 0.8 // ML-specific threshold (different from Layer 2/3 routing threshold)
	)

	var result *IndustryDetectionResult

	if !s.useML || s.mlClassifier == nil {
		// ML not available - use base result
		s.logger.Printf("ℹ️ ML not available, using base classification result (request: %s)", requestID)
		result = s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
	} else if multiResult.Confidence < lowConfidenceThreshold {
		// Low confidence (< 0.5): ML-assisted improvement
		s.logger.Printf("🔧 Low confidence (%.2f%%) < %.2f%%, using ML-assisted improvement (request: %s)",
			multiResult.Confidence*100, lowConfidenceThreshold*100, requestID)
		result, err = s.improveWithML(ctx, multiResult, businessName, description, websiteURL, requestID)
		if err != nil {
			s.logger.Printf("⚠️ ML improvement failed (non-fatal): %v, using base result (request: %s)", err, requestID)
			result = s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
		}
	} else if multiResult.Confidence < mlHighConfidenceThreshold {
		// Medium confidence (0.5-0.8): Ensemble validation
		s.logger.Printf("⚖️ Medium confidence (%.2f%%) between %.2f%% and %.2f%%, using ensemble validation (request: %s)",
			multiResult.Confidence*100, lowConfidenceThreshold*100, mlHighConfidenceThreshold*100, requestID)
		result, err = s.validateWithEnsemble(ctx, multiResult, businessName, description, websiteURL, requestID)
		if err != nil {
			s.logger.Printf("⚠️ Ensemble validation failed (non-fatal): %v, using base result (request: %s)", err, requestID)
			result = s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
		}
	} else {
		// High confidence (>= 0.8): ML validation only
		s.logger.Printf("✅ High confidence (%.2f%%) >= %.2f%%, using ML validation (request: %s)",
			multiResult.Confidence*100, mlHighConfidenceThreshold*100, requestID)
		result, err = s.validateWithMLHighConfidence(ctx, multiResult, businessName, description, websiteURL, requestID)
		if err != nil {
			s.logger.Printf("⚠️ ML validation failed (non-fatal): %v, using base result (request: %s)", err, requestID)
			result = s.convertToIndustryDetectionResult(multiResult, resultMethod, requestID)
		}
	}

	// Record metrics if monitoring is enabled
	if s.monitor != nil {
		// Note: RecordClassificationMetrics signature may need adjustment
		// For now, we'll skip monitoring if method doesn't exist
		// This can be added later when monitoring is fully integrated
	}

	// Phase 2: Generate explanation (codes will be added later in handler if available)
	// Estimate content quality for explanation
	contentQuality := 0.7
	if result.Confidence > 0.8 {
		contentQuality = 0.85
	} else if result.Confidence < 0.5 {
		contentQuality = 0.5
	}

	// Convert result back to MultiStrategyResult for explanation generation
	multiResultForExplanation := &MultiStrategyResult{
		PrimaryIndustry: result.IndustryName,
		Confidence:      result.Confidence,
		Keywords:        result.Keywords,
		Method:          result.Method,
		Reasoning:       result.Reasoning,
		ProcessingTime:  result.ProcessingTime,
	}

	// Generate explanation (codes will be nil for now, can be enhanced in handler)
	explanation := s.explanationGenerator.GenerateExplanation(
		multiResultForExplanation,
		nil, // Codes not available at service level, will be added in handler
		contentQuality,
	)
	result.Explanation = explanation

	s.logger.Printf("✅ Industry detection completed: %s (confidence: %.2f%%) (request: %s)",
		result.IndustryName, result.Confidence*100, requestID)

	return result, nil
}

// convertToIndustryDetectionResult converts MultiStrategyResult to IndustryDetectionResult
func (s *IndustryDetectionService) convertToIndustryDetectionResult(
	multiResult *MultiStrategyResult,
	method string,
	requestID string,
) *IndustryDetectionResult {
	// Phase 2: Generate explanation for the result
	contentQuality := 0.7
	if multiResult.Confidence > 0.8 {
		contentQuality = 0.85
	} else if multiResult.Confidence < 0.5 {
		contentQuality = 0.5
	}

	// Generate explanation (codes will be added later in handler if available)
	explanation := s.explanationGenerator.GenerateExplanation(
		multiResult,
		nil, // Codes not available at this level
		contentQuality,
	)

	// Phase 2: Use method from multiResult if provided, otherwise use passed method
	resultMethod := method
	if multiResult.Method != "" {
		resultMethod = multiResult.Method
	}
	if resultMethod == "" {
		resultMethod = "multi_strategy"
	}

	return &IndustryDetectionResult{
		IndustryName:   multiResult.PrimaryIndustry,
		Confidence:     multiResult.Confidence,
		Keywords:       multiResult.Keywords,
		ProcessingTime: multiResult.ProcessingTime,
		Method:         resultMethod, // Phase 2: Use method from multiResult
		Reasoning:      multiResult.Reasoning,
		Explanation:    explanation, // Phase 2: Include explanation
		CreatedAt:      time.Now(),
	}
}

// performMLClassification performs ML classification and returns a MultiStrategyResult
// Phase 5.1: Simplified to use ML classifier directly instead of MultiMethodClassifier
func (s *IndustryDetectionService) performMLClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*MultiStrategyResult, error) {
	if s.mlClassifier == nil {
		return nil, fmt.Errorf("ML classifier not available")
	}

	// Combine business information for ML analysis
	content := strings.TrimSpace(businessName + " " + description)
	if content == "" {
		return nil, fmt.Errorf("no content to analyze")
	}

	// Perform ML classification
	mlResult, err := s.mlClassifier.ClassifyContent(ctx, content, "")
	if err != nil {
		return nil, fmt.Errorf("ML classification failed: %w", err)
	}

	// Find the best classification from ML result
	if len(mlResult.Classifications) == 0 {
		return nil, fmt.Errorf("no classifications returned from ML model")
	}

	// Get the highest confidence classification
	bestClassification := mlResult.Classifications[0]
	for _, classification := range mlResult.Classifications {
		if classification.Confidence > bestClassification.Confidence {
			bestClassification = classification
		}
	}

	// Convert to MultiStrategyResult format
	result := &MultiStrategyResult{
		PrimaryIndustry: bestClassification.Label,
		Confidence:      bestClassification.Confidence,
		Keywords:        []string{}, // ML doesn't provide specific keywords
		Reasoning:       fmt.Sprintf("ML classification: %s (confidence: %.2f%%)", bestClassification.Label, bestClassification.Confidence*100),
	}

	return result, nil
}

// improveWithML performs ML-assisted improvement for low confidence cases (< 0.5)
// Phase 3.1: Low confidence strategy - ML-assisted improvement
// Uses ensemble voting (Base 40% + ML 60%) - favors ML to improve accuracy
func (s *IndustryDetectionService) improveWithML(
	ctx context.Context,
	baseResult *MultiStrategyResult,
	businessName, description, websiteURL string,
	requestID string,
) (*IndustryDetectionResult, error) {
	// Create timeout context for ML improvement
	mlCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Get ML classification
	mlResult, err := s.performMLClassification(mlCtx, businessName, description, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("ML classification failed: %w", err)
	}

	// Ensemble voting: Base 40% + ML 60% (favor ML for improvement)
	const baseWeight = 0.4
	const mlWeight = 0.6

	var finalIndustry string
	var finalConfidence float64
	var finalReasoning string

	// Check if ML confidence is significantly higher than base
	const mlAdvantageThreshold = 0.2
	if mlResult.Confidence > baseResult.Confidence+mlAdvantageThreshold {
		// ML is significantly better - use ML result
		finalIndustry = mlResult.PrimaryIndustry
		finalConfidence = mlResult.Confidence
		finalReasoning = fmt.Sprintf("ML-assisted improvement: ML confidence (%.2f%%) significantly higher than base (%.2f%%)",
			mlResult.Confidence*100, baseResult.Confidence*100)
		s.logger.Printf("✅ ML significantly better (%.2f%% vs %.2f%%), using ML result (request: %s)",
			mlResult.Confidence*100, baseResult.Confidence*100, requestID)
	} else if mlResult.PrimaryIndustry == baseResult.PrimaryIndustry {
		// Consensus - boost confidence
		weightedConfidence := (baseResult.Confidence * baseWeight) + (mlResult.Confidence * mlWeight)
		consensusBoost := 0.15
		finalIndustry = baseResult.PrimaryIndustry
		finalConfidence = math.Min(weightedConfidence+consensusBoost, 1.0)
		finalReasoning = fmt.Sprintf("ML-assisted improvement with consensus: boosted confidence by %.2f%%",
			consensusBoost*100)
		s.logger.Printf("✅ ML consensus: boosted confidence from %.2f%% to %.2f%% (request: %s)",
			weightedConfidence*100, finalConfidence*100, requestID)
	} else {
		// Disagreement - use weighted average, favor ML
		baseScore := baseResult.Confidence * baseWeight
		mlScore := mlResult.Confidence * mlWeight
		
		if mlScore >= baseScore {
			finalIndustry = mlResult.PrimaryIndustry
			finalConfidence = mlResult.Confidence
			finalReasoning = fmt.Sprintf("ML-assisted improvement: ML selected (weighted score: %.2f%% vs base: %.2f%%)",
				mlScore*100, baseScore*100)
			s.logger.Printf("✅ ML selected via weighted voting (ML: %.2f%% vs Base: %.2f%%) (request: %s)",
				mlScore*100, baseScore*100, requestID)
		} else {
			finalIndustry = baseResult.PrimaryIndustry
			finalConfidence = baseResult.Confidence
			finalReasoning = fmt.Sprintf("ML-assisted improvement: Base selected (weighted score: %.2f%% vs ML: %.2f%%)",
				baseScore*100, mlScore*100)
			s.logger.Printf("✅ Base selected via weighted voting (Base: %.2f%% vs ML: %.2f%%) (request: %s)",
				baseScore*100, mlScore*100, requestID)
		}
	}

	// Merge keywords
	keywords := append(baseResult.Keywords, mlResult.Keywords...)
	keywordMap := make(map[string]bool)
	uniqueKeywords := []string{}
	for _, kw := range keywords {
		if !keywordMap[kw] {
			keywordMap[kw] = true
			uniqueKeywords = append(uniqueKeywords, kw)
		}
	}

	return &IndustryDetectionResult{
		IndustryName:   finalIndustry,
		Confidence:     finalConfidence,
		Keywords:       uniqueKeywords,
		ProcessingTime: baseResult.ProcessingTime, // Base processing time
		Method:         "multi_strategy_ml_improved",
		Reasoning:      finalReasoning,
		CreatedAt:      time.Now(),
	}, nil
}

// validateWithEnsemble performs ensemble validation for medium confidence cases (0.5-0.8)
// Phase 3.1: Medium confidence strategy - Ensemble validation
// Uses balanced ensemble (Base 50% + ML 50%)
func (s *IndustryDetectionService) validateWithEnsemble(
	ctx context.Context,
	baseResult *MultiStrategyResult,
	businessName, description, websiteURL string,
	requestID string,
) (*IndustryDetectionResult, error) {
	// Create timeout context for ML validation
	mlCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Get ML classification
	mlResult, err := s.performMLClassification(mlCtx, businessName, description, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("ML classification failed: %w", err)
	}

	// Balanced ensemble: Base 50% + ML 50%
	const baseWeight = 0.5
	const mlWeight = 0.5

	var finalIndustry string
	var finalConfidence float64
	var finalReasoning string

	if mlResult.PrimaryIndustry == baseResult.PrimaryIndustry {
		// Consensus - boost confidence
		weightedConfidence := (baseResult.Confidence * baseWeight) + (mlResult.Confidence * mlWeight)
		consensusBoost := 0.1
		finalIndustry = baseResult.PrimaryIndustry
		finalConfidence = math.Min(weightedConfidence+consensusBoost, 1.0)
		finalReasoning = fmt.Sprintf("Ensemble validation with consensus: boosted confidence by %.2f%%",
			consensusBoost*100)
		s.logger.Printf("✅ Ensemble consensus: boosted confidence from %.2f%% to %.2f%% (request: %s)",
			weightedConfidence*100, finalConfidence*100, requestID)
	} else {
		// Disagreement - use weighted average
		baseScore := baseResult.Confidence * baseWeight
		mlScore := mlResult.Confidence * mlWeight
		weightedConfidence := baseScore + mlScore
		
		// Select industry based on weighted scores
		if mlScore >= baseScore {
			finalIndustry = mlResult.PrimaryIndustry
		} else {
			finalIndustry = baseResult.PrimaryIndustry
		}
		
		finalConfidence = weightedConfidence
		finalReasoning = fmt.Sprintf("Ensemble validation: weighted average (Base: %s %.2f%%, ML: %s %.2f%%)",
			baseResult.PrimaryIndustry, baseScore*100, mlResult.PrimaryIndustry, mlScore*100)
		s.logger.Printf("⚠️ Ensemble disagreement: Base '%s' (%.2f%%) vs ML '%s' (%.2f%%), using weighted average (request: %s)",
			baseResult.PrimaryIndustry, baseScore*100, mlResult.PrimaryIndustry, mlScore*100, requestID)
	}

	// Merge keywords
	keywords := append(baseResult.Keywords, mlResult.Keywords...)
	keywordMap := make(map[string]bool)
	uniqueKeywords := []string{}
	for _, kw := range keywords {
		if !keywordMap[kw] {
			keywordMap[kw] = true
			uniqueKeywords = append(uniqueKeywords, kw)
		}
	}

	return &IndustryDetectionResult{
		IndustryName:   finalIndustry,
		Confidence:     finalConfidence,
		Keywords:       uniqueKeywords,
		ProcessingTime: baseResult.ProcessingTime, // Base processing time
		Method:         "multi_strategy_ml_ensemble",
		Reasoning:      finalReasoning,
		CreatedAt:      time.Now(),
	}, nil
}

// validateWithMLHighConfidence performs ML validation for high confidence cases (>= 0.8)
// Phase 3.1: High confidence strategy - ML validation only
// ML validates, doesn't replace base classification
func (s *IndustryDetectionService) validateWithMLHighConfidence(
	ctx context.Context,
	baseResult *MultiStrategyResult,
	businessName, description, websiteURL string,
	requestID string,
) (*IndustryDetectionResult, error) {
	// Create timeout context for ML validation (shorter timeout since it's validation)
	mlCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Get ML classification
	mlResult, err := s.performMLClassification(mlCtx, businessName, description, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("ML classification failed: %w", err)
	}

	// ML validation logic: validate, don't replace
	var finalIndustry string
	var finalConfidence float64
	var finalReasoning string

	if mlResult.PrimaryIndustry == baseResult.PrimaryIndustry {
		// Consensus - boost confidence
		boostAmount := 0.1
		finalIndustry = baseResult.PrimaryIndustry
		finalConfidence = math.Min(baseResult.Confidence+boostAmount, 1.0)
		finalReasoning = baseResult.Reasoning + " (ML validated with consensus)"
		s.logger.Printf("✅ ML consensus: boosted confidence from %.2f%% to %.2f%% (request: %s)",
			baseResult.Confidence*100, finalConfidence*100, requestID)
	} else {
		// Disagreement - use base result but note ML suggestion
		finalIndustry = baseResult.PrimaryIndustry
		finalConfidence = baseResult.Confidence
		finalReasoning = baseResult.Reasoning + fmt.Sprintf(" (ML suggested: %s, but base classification used)",
			mlResult.PrimaryIndustry)
		s.logger.Printf("⚠️ ML disagreement: ML suggested '%s', but using base result '%s' (request: %s)",
			mlResult.PrimaryIndustry, baseResult.PrimaryIndustry, requestID)
	}

	// Merge keywords
	keywords := append(baseResult.Keywords, mlResult.Keywords...)
	keywordMap := make(map[string]bool)
	uniqueKeywords := []string{}
	for _, kw := range keywords {
		if !keywordMap[kw] {
			keywordMap[kw] = true
			uniqueKeywords = append(uniqueKeywords, kw)
		}
	}

	return &IndustryDetectionResult{
		IndustryName:   finalIndustry,
		Confidence:     finalConfidence,
		Keywords:       uniqueKeywords,
		ProcessingTime: baseResult.ProcessingTime,
		Method:         "multi_strategy_ml_validated",
		Reasoning:      finalReasoning,
		CreatedAt:      time.Now(),
	}, nil
}

// detectIndustryWithML is deprecated - ML is now integrated into DetectIndustry via three-tier strategy
// This method is kept for backward compatibility but should not be used
func (s *IndustryDetectionService) detectIndustryWithML(
	ctx context.Context,
	businessName, description, websiteURL string,
	startTime time.Time,
	requestID string,
) (*IndustryDetectionResult, error) {
	s.logger.Printf("⚠️ detectIndustryWithML is deprecated, using DetectIndustry instead (request: %s)", requestID)
	// Delegate to main DetectIndustry method
	return s.DetectIndustry(ctx, businessName, description, websiteURL)
}

// fallbackToKeywordClassification provides fallback when multi-strategy fails
func (s *IndustryDetectionService) fallbackToKeywordClassification(
	ctx context.Context,
	businessName, description, websiteURL string,
	startTime time.Time,
	requestID string,
) (*IndustryDetectionResult, error) {
	// Extract keywords using database-driven approach (with description fallback)
	keywords, err := s.extractKeywordsFromDatabase(ctx, businessName, description, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to extract keywords: %w", err)
	}

	if len(keywords) == 0 {
		// Enhanced fallback: Extract keywords from business name and description
		s.logger.Printf("⚠️ No keywords extracted from website for: %s, trying fallback extraction", businessName)
		keywords = s.extractKeywordsFromNameAndDescription(businessName, description)

	if len(keywords) == 0 {
		s.logger.Printf("⚠️ No keywords extracted for: %s", businessName)
		return &IndustryDetectionResult{
			IndustryName:   "General Business",
			Confidence:     0.30,
			Keywords:       []string{},
			ProcessingTime: time.Since(startTime),
			Method:         "database_driven",
				Reasoning:      "No relevant keywords found in database or fallback extraction",
			CreatedAt:      time.Now(),
		}, nil
		}
		s.logger.Printf("✅ Fallback extraction found %d keywords from name/description", len(keywords))
	}

	// Classify using database-driven keyword matching
	result, err := s.classifyByKeywords(ctx, keywords)
	if err != nil {
		return nil, fmt.Errorf("failed to classify by keywords: %w", err)
	}

	result.ProcessingTime = time.Since(startTime)
	result.Method = "database_driven"
	result.CreatedAt = time.Now()

	s.logger.Printf("✅ Industry detection completed (fallback): %s (confidence: %.2f%%) (request: %s)",
		result.IndustryName, result.Confidence*100, requestID)

	return result, nil
}

// extractKeywordsFromDatabase extracts keywords using database-driven approach
// Now includes description for enhanced fallback keyword extraction
func (s *IndustryDetectionService) extractKeywordsFromDatabase(ctx context.Context, businessName, description, websiteURL string) ([]string, error) {
	// Use the repository's classification method to get keywords
	// Note: ClassifyBusiness may return "General Business" if only URL keywords are available,
	// but it will still return the expanded keywords from the fallback chain.
	// The actual industry classification happens in classifyByKeywords which uses these expanded keywords.
	result, err := s.repo.ClassifyBusiness(ctx, businessName, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("failed to classify business: %w", err)
	}

	if result == nil {
		return []string{}, nil
	}

	keywords := result.Keywords

	// Enhanced: Always supplement with description keywords if available
	// This ensures we use description even when website keywords exist
	if description != "" {
		descKeywords := s.extractKeywordsFromNameAndDescription(businessName, description)
		// Merge keywords, avoiding duplicates
		keywordSet := make(map[string]bool)
		for _, kw := range keywords {
			keywordSet[kw] = true
		}
		for _, kw := range descKeywords {
			if !keywordSet[kw] {
				keywords = append(keywords, kw)
				keywordSet[kw] = true
			}
		}
		if len(descKeywords) > 0 {
			s.logger.Printf("✅ Supplemented with %d keywords from description (total: %d)", len(descKeywords), len(keywords))
		}
	}

	// If still no keywords, try description-only extraction
	if len(keywords) == 0 && description != "" {
		s.logger.Printf("⚠️ No keywords from website, extracting from description for: %s", businessName)
		fallbackKeywords := s.extractKeywordsFromNameAndDescription(businessName, description)
		if len(fallbackKeywords) > 0 {
			keywords = fallbackKeywords
			s.logger.Printf("✅ Extracted %d keywords from description fallback", len(keywords))
		}
	}

	// Return the keywords - these should be the expanded keywords from extractKeywords
	// The industry from ClassifyBusiness may be "General Business" if only 4 URL keywords were found,
	// but classifyByKeywords will use these keywords (which may be expanded by keyword index matching)
	// to correctly identify the industry (e.g., "Wineries")
	return keywords, nil
}

// extractKeywordsFromNameAndDescription extracts keywords from business name and description
// This is a fallback when website scraping fails
func (s *IndustryDetectionService) extractKeywordsFromNameAndDescription(businessName, description string) []string {
	var keywords []string
	seen := make(map[string]bool)

	// Combine business name and description
	text := strings.ToLower(businessName)
	if description != "" {
		text += " " + strings.ToLower(description)
	}

	// Extract meaningful words (3+ characters, not stop words)
	words := strings.Fields(text)
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "was": true, "are": true, "were": true, "be": true,
		"been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"should": true, "could": true, "may": true, "might": true, "must": true,
		"can": true, "this": true, "that": true, "these": true, "those": true,
		"it": true, "its": true, "they": true, "them": true, "their": true,
		"our": true, "your": true, "my": true, "his": true, "her": true,
		"he": true, "she": true, "we": true, "you": true, "i": true, "me": true, "us": true,
		"inc": true, "llc": true, "ltd": true, "corp": true, "corporation": true,
		"company": true, "co": true, "com": true, "www": true, "http": true, "https": true,
		"services": true, "service": true, // Remove generic service words to focus on industry-specific terms
	}

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:()[]{}\"'")
		if len(word) >= 3 && !stopWords[word] && !seen[word] {
			seen[word] = true
			keywords = append(keywords, word)
		}
	}

	// Also extract industry-specific terms from common patterns
	industryPatterns := map[string][]string{
		"technology":     {"software", "tech", "digital", "computer", "internet", "web", "app", "platform", "system", "data", "cloud", "ai", "ml", "api"},
		"healthcare":     {"health", "medical", "hospital", "clinic", "doctor", "patient", "care", "treatment", "pharmacy", "diagnostic", "therapy"},
		"financial":      {"finance", "financial", "bank", "investment", "credit", "loan", "insurance", "trading", "accounting", "tax", "money"},
		"retail":         {"retail", "store", "shop", "merchandise", "product", "sale", "customer", "shopping", "boutique", "outlet"},
		"manufacturing": {"manufacturing", "production", "factory", "industrial", "machinery", "equipment", "assembly", "fabrication"},
		"construction":  {"construction", "contractor", "building", "construction", "architect", "engineering", "renovation", "development"},
		"transportation": {"transport", "transportation", "logistics", "shipping", "delivery", "freight", "trucking", "airline", "railway"},
		"professional":  {"consulting", "professional", "service", "advisory", "legal", "accounting", "management", "consultant"},
	}

	textLower := strings.ToLower(text)
	for _, terms := range industryPatterns {
		for _, term := range terms {
			if strings.Contains(textLower, term) && !seen[term] {
				seen[term] = true
				keywords = append(keywords, term)
			}
		}
	}

	return keywords
}

// classifyByKeywords performs classification using database-driven keyword matching
func (s *IndustryDetectionService) classifyByKeywords(ctx context.Context, keywords []string) (*IndustryDetectionResult, error) {
	// Use the repository's classification method
	classification, err := s.repo.ClassifyBusinessByKeywords(ctx, keywords)
	if err != nil {
		return nil, fmt.Errorf("database classification failed: %w", err)
	}

	if classification == nil {
		return &IndustryDetectionResult{
			IndustryName: "General Business",
			Confidence:   0.30,
			Keywords:     keywords,
			Reasoning:    "No matching industry found in database",
		}, nil
	}

	return &IndustryDetectionResult{
		IndustryName: classification.Industry.Name,
		Confidence:   classification.Confidence,
		Keywords:     keywords,
		Reasoning:    fmt.Sprintf("Matched %d keywords to %s industry", len(keywords), classification.Industry.Name),
	}, nil
}

// isTechnicalTerm checks if a word is a technical term that should be filtered out
func (s *IndustryDetectionService) isTechnicalTerm(word string) bool {
	// Technical terms that should be filtered out
	technicalTerms := map[string]bool{
		// HTML/CSS/JavaScript terms
		"html": true, "css": true, "javascript": true, "js": true, "jquery": true,
		"bootstrap": true, "react": true, "angular": true, "vue": true, "node": true,
		"php": true, "python": true, "java": true, "csharp": true, "ruby": true,
		"sql": true, "mysql": true, "postgresql": true, "mongodb": true, "redis": true,
		"api": true, "rest": true, "graphql": true, "json": true, "xml": true,
		"http": true, "https": true, "ssl": true, "tls": true, "dns": true,
		"cdn": true, "aws": true, "azure": true, "gcp": true, "docker": true,
		"kubernetes": true, "git": true, "github": true, "gitlab": true,

		// Common web development terms
		"div": true, "span": true, "class": true, "id": true, "src": true,
		"href": true, "alt": true, "title": true, "meta": true, "script": true,
		"style": true, "link": true, "img": true, "button": true, "input": true,
		"form": true, "table": true, "tr": true, "td": true, "th": true,

		// Common programming terms
		"function": true, "variable": true, "array": true, "object": true,
		"string": true, "integer": true, "boolean": true, "null": true,
		"undefined": true, "true": true, "false": true, "if": true,
		"else": true, "for": true, "while": true, "return": true,

		// Common system terms
		"system": true, "server": true, "client": true, "database": true,
		"cache": true, "session": true, "cookie": true, "token": true,
		"auth": true, "login": true, "logout": true, "register": true,
		"admin": true, "user": true, "guest": true, "public": true,
		"private": true, "protected": true, "static": true, "dynamic": true,
	}

	return technicalTerms[strings.ToLower(word)]
}

// generateRequestID generates a unique request ID for tracking
func (s *IndustryDetectionService) generateRequestID() string {
	return fmt.Sprintf("industry_detection_%d", time.Now().UnixNano())
}

// GetIndustryDetectionMetrics returns current industry detection performance metrics
func (s *IndustryDetectionService) GetIndustryDetectionMetrics(ctx context.Context) (*ClassificationAccuracyStats, error) {
	if s.monitor == nil {
		return nil, fmt.Errorf("monitoring not configured")
	}

	// Note: This would call the actual monitoring method when implemented
	// return s.monitor.GetClassificationAccuracyStats(ctx, 24*time.Hour)
	return nil, fmt.Errorf("monitoring not fully implemented")
}

// ValidateIndustryDetectionResult validates that the detection result is consistent
func (s *IndustryDetectionService) ValidateIndustryDetectionResult(result *IndustryDetectionResult) error {
	if result == nil {
		return fmt.Errorf("industry detection result cannot be nil")
	}

	// Validate industry name
	if result.IndustryName == "" {
		return fmt.Errorf("industry name cannot be empty")
	}

	// Validate confidence score
	if result.Confidence < 0.0 || result.Confidence > 1.0 {
		return fmt.Errorf("invalid confidence score: %.2f (must be between 0.0 and 1.0)", result.Confidence)
	}

	// Validate processing time
	if result.ProcessingTime < 0 {
		return fmt.Errorf("invalid processing time: %v (must be non-negative)", result.ProcessingTime)
	}

	// Validate method
	if result.Method == "" {
		return fmt.Errorf("detection method cannot be empty")
	}

	return nil
}

// GetIndustryDetectionStatistics returns statistics about industry detection
func (s *IndustryDetectionService) GetIndustryDetectionStatistics() map[string]interface{} {
	return map[string]interface{}{
		"service_name":       "IndustryDetectionService",
		"version":            "2.0.0",
		"database_driven":    true,
		"hardcoded_patterns": false,
		"monitoring_enabled": s.monitor != nil,
		"created_at":         time.Now(),
	}
}

// SetEmbeddingClassifier sets the embedding classifier for Layer 2 (Phase 3)
func (s *IndustryDetectionService) SetEmbeddingClassifier(embeddingClassifier *EmbeddingClassifier) {
	s.embeddingClassifier = embeddingClassifier
	s.logger.Printf("✅ [Phase 3] Embedding classifier set for Layer 2")
}

// SetLLMClassifier sets the LLM classifier for Layer 3 (Phase 4)
func (s *IndustryDetectionService) SetLLMClassifier(llmClassifier *LLMClassifier) {
	s.llmClassifier = llmClassifier
	s.logger.Printf("✅ [Phase 4] LLM classifier set for Layer 3")
	
	// Initialize async LLM processor
	store := NewAsyncLLMStore(30*time.Minute, 1000) // 30 min TTL, max 1000 results
	s.asyncLLMProcessor = NewAsyncLLMProcessor(
		store,
		llmClassifier,
		s.logger,
		5*time.Minute, // 5 minute timeout for LLM calls
	)
	s.logger.Printf("✅ [Phase 4] Async LLM processor initialized")
}

// GetAsyncLLMResult retrieves the result of async LLM processing
func (s *IndustryDetectionService) GetAsyncLLMResult(processingID string) (*AsyncLLMResult, bool) {
	if s.asyncLLMProcessor == nil {
		return nil, false
	}
	return s.asyncLLMProcessor.GetResult(processingID)
}

// GetAsyncLLMStats returns statistics about async LLM processing
func (s *IndustryDetectionService) GetAsyncLLMStats() map[string]interface{} {
	if s.asyncLLMProcessor == nil {
		return map[string]interface{}{"enabled": false}
	}
	stats := s.asyncLLMProcessor.GetStats()
	stats["enabled"] = true
	return stats
}

// isAmbiguousCase checks if a business description indicates ambiguity
// Ambiguous cases should use Layer 3 (LLM) even if Layer 1 confidence is high
func (s *IndustryDetectionService) isAmbiguousCase(businessName, description string) bool {
	desc := strings.ToLower(description)
	ambiguousKeywords := []string{
		"diversified",
		"multiple sectors",
		"various services",
		"multi-industry",
		"cross-sector",
		"various industries",
		"multiple businesses",
		"wide range",
		"broad range",
		"multiple industries",
		"various sectors",
		"cross-industry",
		"multi-sector",
		"diverse portfolio",
		"multiple markets",
	}

	for _, keyword := range ambiguousKeywords {
		if strings.Contains(desc, keyword) {
			return true
		}
	}

	// Check for very short or vague descriptions
	if len(description) < 50 {
		return true
	}

	// Check for vague business language
	vaguePhrases := []string{
		"help businesses",
		"provide solutions",
		"strategic partnerships",
		"innovative solutions",
		"business services",
		"consulting services",
		"professional services",
	}

	for _, phrase := range vaguePhrases {
		if strings.Contains(desc, phrase) && len(description) < 100 {
			return true
		}
	}

	return false
}

// getScrapedContentForLayer2 gets ScrapedContent for Layer 2 classification
// This is a simplified version - in production, you may want to use a cached scraper
func (s *IndustryDetectionService) getScrapedContentForLayer2(ctx context.Context, websiteURL string) (*external.ScrapedContent, error) {
	if websiteURL == "" {
		return nil, fmt.Errorf("website URL is required for Layer 2")
	}

	// For now, create a basic scraper instance
	// In production, you may want to reuse an existing scraper from the service
	// Note: external.NewWebsiteScraper requires a zap.Logger, but we only have log.Logger
	// For now, we'll create a minimal scraper or use a simpler approach
	// TODO: Consider adding a scraper field to IndustryDetectionService or using a shared scraper
	
	// Create a basic HTTP client to fetch the page
	// This is a simplified version - in production, use the full WebsiteScraper
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	req, err := http.NewRequestWithContext(ctx, "GET", websiteURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "KYB-Platform-Bot/1.0")
	
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch website: %w", err)
	}
	defer resp.Body.Close()
	
	// Read content (simplified - in production, use full scraper)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Create basic ScrapedContent
	// In production, this should use the full WebsiteScraper with structured extraction
	return &external.ScrapedContent{
		RawHTML:   string(body),
		PlainText: string(body),
		Title:     "", // Will be extracted if available
		Domain:    websiteURL,
		ScrapedAt: time.Now(),
	}, nil
}

// buildResultFromEmbedding builds IndustryDetectionResult from EmbeddingClassificationResult
func (s *IndustryDetectionService) buildResultFromEmbedding(
	embResult *EmbeddingClassificationResult,
	method string,
	requestID string,
) *IndustryDetectionResult {
	// Derive primary industry from top MCC match
	primaryIndustry := "Unknown"
	if len(embResult.MCC) > 0 {
		primaryIndustry = embResult.MCC[0].Description
	}

	// Generate reasoning
	reasoning := fmt.Sprintf(
		"Semantic similarity analysis matched '%s' with %.0f%% confidence",
		embResult.TopMatch,
		embResult.TopSimilarity*100,
	)

	// Estimate content quality for explanation
	contentQuality := 0.7
	if embResult.Confidence > 0.8 {
		contentQuality = 0.85
	} else if embResult.Confidence < 0.5 {
		contentQuality = 0.5
	}

	// Create MultiStrategyResult for explanation generation
	multiResultForExplanation := &MultiStrategyResult{
		PrimaryIndustry: primaryIndustry,
		Confidence:      embResult.Confidence,
		Keywords:        []string{}, // Embeddings don't use keywords
		Method:          embResult.Method,
		Reasoning:       reasoning,
		ProcessingTime:  time.Duration(embResult.ProcessingTimeMs) * time.Millisecond,
	}

	// Generate explanation
	explanation := s.explanationGenerator.GenerateExplanation(
		multiResultForExplanation,
		nil, // Codes not available at service level
		contentQuality,
	)

	return &IndustryDetectionResult{
		IndustryName:   primaryIndustry,
		Confidence:     embResult.Confidence,
		Keywords:       []string{}, // Embeddings don't use keywords
		ProcessingTime: time.Duration(embResult.ProcessingTimeMs) * time.Millisecond,
		Method:         method,
		Reasoning:      reasoning,
		Explanation:    explanation,
		CreatedAt:      time.Now(),
	}
}

// buildResultFromLLM builds IndustryDetectionResult from LLMClassificationResult
func (s *IndustryDetectionService) buildResultFromLLM(
	llmResult *LLMClassificationResult,
	method string,
	requestID string,
) *IndustryDetectionResult {
	// Use primary industry from LLM
	primaryIndustry := llmResult.PrimaryIndustry
	if primaryIndustry == "" {
		primaryIndustry = "Unknown"
	}

	// Use LLM reasoning
	reasoning := llmResult.Reasoning
	if reasoning == "" {
		reasoning = "LLM-based classification with advanced reasoning"
	}

	// Estimate content quality for explanation
	contentQuality := 0.7
	if llmResult.Confidence > 0.8 {
		contentQuality = 0.85
	} else if llmResult.Confidence < 0.5 {
		contentQuality = 0.5
	}

	// Create MultiStrategyResult for explanation generation
	multiResultForExplanation := &MultiStrategyResult{
		PrimaryIndustry: primaryIndustry,
		Confidence:      llmResult.Confidence,
		Keywords:        []string{}, // LLM doesn't use keywords
		Method:          "llm_reasoning",
		Reasoning:       reasoning,
		ProcessingTime:  time.Duration(llmResult.ProcessingTimeMs) * time.Millisecond,
	}

	// Generate explanation
	explanation := s.explanationGenerator.GenerateExplanation(
		multiResultForExplanation,
		nil, // Codes not available at service level
		contentQuality,
	)

	// Enhance explanation with LLM-specific details
	if explanation != nil {
		explanation.PrimaryReason = reasoning
		explanation.SupportingFactors = append(explanation.SupportingFactors,
			"Advanced LLM reasoning with context understanding",
			"Considers business model complexity and nuance",
			"Provides detailed rationale for classification",
		)
		// Add alternative classifications as supporting factors if available
		if len(llmResult.AlternativeClassifications) > 0 {
			for _, alt := range llmResult.AlternativeClassifications {
				explanation.SupportingFactors = append(explanation.SupportingFactors,
					fmt.Sprintf("Alternative classification considered: %s", alt))
			}
		}
		explanation.MethodUsed = "llm_reasoning"
	}

	return &IndustryDetectionResult{
		IndustryName:   primaryIndustry,
		Confidence:     llmResult.Confidence,
		Keywords:       []string{}, // LLM doesn't use keywords
		ProcessingTime: time.Duration(llmResult.ProcessingTimeMs) * time.Millisecond,
		Method:         method,
		Reasoning:      reasoning,
		Explanation:    explanation,
		CreatedAt:      time.Now(),
	}
}
