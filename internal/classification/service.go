package classification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/datasource"
	"github.com/pcraw4d/business-verification/internal/observability"
	// "github.com/pcraw4d/business-verification/internal/webanalysis" // Temporarily disabled
	"github.com/pcraw4d/business-verification/pkg/validators"
)

// ClassificationService provides business classification functionality
type ClassificationService struct {
	config       *config.ExternalServicesConfig
	db           database.Database
	logger       *observability.Logger
	metrics      *observability.Metrics
	industryData *IndustryCodeData

	// in-memory cache
	cacheEnabled bool
	cacheTTL     time.Duration
	cacheMax     int
	cacheMu      sync.RWMutex
	cache        map[string]cacheEntry
	janitorStop  chan struct{}

	// enrichment
	enricher *datasource.Aggregator

	// web analysis (temporarily disabled)
	// webAnalysis   *webanalysis.ClassificationFlowManager
	// hybridScraper *webanalysis.HybridAPIScraper

	// search integration (temporarily disabled)
	// multiSourceSearch *webanalysis.MultiSourceSearchService
	// searchAnalyzer    *webanalysis.SearchResultAnalyzer

	// crosswalk mapping
	crosswalkMapper *CrosswalkMapper

	// geographic region awareness
	geographicManager *GeographicManager

	// industry-specific mapping
	industryMapper *IndustryMapper

	// feedback collection
	feedbackCollector *FeedbackCollector
}

// NewClassificationService creates a new business classification service
func NewClassificationService(cfg *config.ExternalServicesConfig, db database.Database, logger *observability.Logger, metrics *observability.Metrics) *ClassificationService {
	s := &ClassificationService{
		config:       cfg,
		db:           db,
		logger:       logger,
		metrics:      metrics,
		industryData: nil, // Will be loaded separately
	}
	s.initCache()
	s.initEnrichment(db)
	s.initWebAnalysis()
	s.initSearchIntegration()
	s.initCrosswalkMapper()
	s.initGeographicManager()
	s.initIndustryMapper()
	s.initFeedbackCollector()
	return s
}

// NewClassificationServiceWithData creates a new business classification service with industry data
func NewClassificationServiceWithData(cfg *config.ExternalServicesConfig, db database.Database, logger *observability.Logger, metrics *observability.Metrics, industryData *IndustryCodeData) *ClassificationService {
	s := &ClassificationService{
		config:       cfg,
		db:           db,
		logger:       logger,
		metrics:      metrics,
		industryData: industryData,
	}
	s.initCache()
	s.initEnrichment(db)
	s.initWebAnalysis()
	s.initSearchIntegration()
	s.initCrosswalkMapper()
	s.initGeographicManager()
	s.initIndustryMapper()
	s.initFeedbackCollector()
	return s
}

type cacheEntry struct {
	classifications []IndustryClassification
	expiresAt       time.Time
}

func (c *ClassificationService) initCache() {
	// defaults
	enabled := true
	ttl := 10 * time.Minute
	maxEntries := 10000
	janitorInterval := time.Minute
	if c.config != nil {
		enabled = c.config.ClassificationCache.Enabled
		if c.config.ClassificationCache.TTL > 0 {
			ttl = c.config.ClassificationCache.TTL
		}
		if c.config.ClassificationCache.MaxEntries > 0 {
			maxEntries = c.config.ClassificationCache.MaxEntries
		}
		if c.config.ClassificationCache.JanitorInterval > 0 {
			janitorInterval = c.config.ClassificationCache.JanitorInterval
		}
	}

	c.cacheEnabled = enabled
	if !c.cacheEnabled {
		return
	}
	c.cacheTTL = ttl
	c.cacheMax = maxEntries
	c.cache = make(map[string]cacheEntry, 1024)
	c.janitorStop = make(chan struct{})

	// start janitor
	go func() {
		ticker := time.NewTicker(janitorInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c.evictExpired()
			case <-c.janitorStop:
				return
			}
		}
	}()
}

func (c *ClassificationService) evictExpired() {
	if !c.cacheEnabled {
		return
	}
	now := time.Now()
	c.cacheMu.Lock()
	for k, v := range c.cache {
		if now.After(v.expiresAt) {
			delete(c.cache, k)
		}
	}
	c.cacheMu.Unlock()
}

func (c *ClassificationService) getFromCache(key string) (classifications []IndustryClassification, ok bool) {
	if !c.cacheEnabled {
		return nil, false
	}
	now := time.Now()
	c.cacheMu.RLock()
	ce, exists := c.cache[key]
	c.cacheMu.RUnlock()
	if !exists || now.After(ce.expiresAt) {
		return nil, false
	}
	return ce.classifications, true
}

func (c *ClassificationService) setCache(key string, classifications []IndustryClassification) {
	if !c.cacheEnabled {
		return
	}
	c.cacheMu.Lock()
	// Simple cap: if over max, evict random/oldest by iterating until size below threshold
	if len(c.cache) >= c.cacheMax {
		// opportunistic eviction of expired first
		now := time.Now()
		for k, v := range c.cache {
			if now.After(v.expiresAt) {
				delete(c.cache, k)
			}
			if len(c.cache) < c.cacheMax {
				break
			}
		}
		// if still full, delete arbitrary entries until under cap
		for k := range c.cache {
			if len(c.cache) < c.cacheMax {
				break
			}
			delete(c.cache, k)
		}
	}
	c.cache[key] = cacheEntry{classifications: classifications, expiresAt: time.Now().Add(c.cacheTTL)}
	c.cacheMu.Unlock()
}

func (c *ClassificationService) makeCacheKey(req *ClassificationRequest) string {
	// Normalize primary text fields for stable hashing
	normalized, _ := normalizeBusinessFields(req.BusinessName, req.Description, req.Keywords)
	base := strings.Join([]string{
		strings.ToLower(strings.TrimSpace(normalized)),
		strings.ToLower(strings.TrimSpace(req.BusinessType)),
		strings.ToLower(strings.TrimSpace(req.Industry)),
	}, "|")
	sum := sha256.Sum256([]byte(base))
	return hex.EncodeToString(sum[:])
}

func (c *ClassificationService) initEnrichment(db database.Database) {
	if db == nil {
		return
	}
	// For now only DB source; easily extended with external APIs later
	src := datasource.NewDBSource(db)
	aggr := datasource.NewAggregator([]datasource.DataSource{src}, 1500*time.Millisecond)

	// Note: HTTP client configuration removed as it's not needed for MVP
	// External data sources can be added later when needed

	c.enricher = aggr
}

// initWebAnalysis initializes the web analysis components
func (c *ClassificationService) initWebAnalysis() {
	config := webanalysis.FlowConfig{
		DefaultMaxResults:          10,
		DefaultConfidenceThreshold: 0.5,
		URLFlowTimeout:             30 * time.Second,
		SearchFlowTimeout:          30 * time.Second,
		FallbackEnabled:            true,
		RetryAttempts:              3,
	}
	c.webAnalysis = webanalysis.NewClassificationFlowManager(config)
	c.hybridScraper = webanalysis.NewHybridAPIScraper()
}

// initSearchIntegration initializes the search integration components
func (c *ClassificationService) initSearchIntegration() {
	if c.config == nil {
		return
	}
	// Initialize search integration components with default values
	// Note: API keys would be configured via environment variables or config
	searchConfig := webanalysis.SearchIntegrationConfig{
		GoogleAPIKey:         "", // Would be set from config
		GoogleSearchEngineID: "", // Would be set from config
		BingAPIKey:           "", // Would be set from config
		BingEndpoint:         "", // Would be set from config
		MaxResults:           10,
		Timeout:              30 * time.Second,
		EnableCaching:        true,
		CacheTTL:             1 * time.Hour,
		RetryAttempts:        3,
		RetryDelay:           1 * time.Second,
	}
	c.multiSourceSearch = webanalysis.NewMultiSourceSearchService(searchConfig)

	analyzerConfig := webanalysis.SearchAnalyzerConfig{
		MinSnippetLength:    50,
		MaxSnippetLength:    300,
		MinQualityScore:     0.3,
		MinConfidenceScore:  0.75,
		EnableDeduplication: true,
		EnableRanking:       true,
		IndustryKeywords:    c.getIndustryKeywords(),
		QualityWeights: map[string]float64{
			"title_length":     0.2,
			"snippet_length":   0.3,
			"url_quality":      0.2,
			"provider_quality": 0.15,
			"relevance":        0.15,
		},
	}
	c.searchAnalyzer = webanalysis.NewSearchResultAnalyzer(analyzerConfig)
}

// initCrosswalkMapper initializes the crosswalk mapping component
func (c *ClassificationService) initCrosswalkMapper() {
	c.crosswalkMapper = NewCrosswalkMapper(c.logger, c.metrics)
}

// initGeographicManager initializes the geographic region manager component
func (c *ClassificationService) initGeographicManager() {
	c.geographicManager = NewGeographicManager(c.logger, c.metrics)
}

// initIndustryMapper initializes the industry-specific mapper component
func (c *ClassificationService) initIndustryMapper() {
	c.industryMapper = NewIndustryMapper(c.logger, c.metrics)
}

// initFeedbackCollector initializes the feedback collector component
func (c *ClassificationService) initFeedbackCollector() {
	c.feedbackCollector = NewFeedbackCollector(c.logger, c.metrics)
}

// getIndustryKeywords returns industry keywords for search analysis
func (c *ClassificationService) getIndustryKeywords() map[string][]string {
	// Default industry keywords for search analysis
	// In a production system, these would be loaded from a database or configuration
	return map[string][]string{
		"technology": {
			"software", "technology", "tech", "digital", "computer", "internet", "web", "mobile", "app",
			"development", "programming", "coding", "IT", "information technology", "cybersecurity",
		},
		"healthcare": {
			"healthcare", "medical", "health", "hospital", "clinic", "doctor", "nurse", "pharmacy",
			"dental", "therapy", "wellness", "fitness", "rehabilitation",
		},
		"finance": {
			"finance", "financial", "banking", "investment", "insurance", "accounting", "tax",
			"credit", "loan", "mortgage", "wealth", "advisory", "consulting",
		},
		"retail": {
			"retail", "store", "shop", "commerce", "ecommerce", "online", "marketplace", "sales",
			"merchandise", "product", "consumer", "shopping",
		},
		"manufacturing": {
			"manufacturing", "factory", "production", "industrial", "machinery", "equipment",
			"assembly", "supply chain", "logistics", "warehouse",
		},
		"education": {
			"education", "school", "university", "college", "training", "learning", "academic",
			"teaching", "student", "course", "curriculum",
		},
		"real_estate": {
			"real estate", "property", "realty", "housing", "construction", "development",
			"commercial", "residential", "broker", "agent",
		},
		"legal": {
			"legal", "law", "attorney", "lawyer", "law firm", "litigation", "contract",
			"compliance", "regulatory", "court", "justice",
		},
		"consulting": {
			"consulting", "consultant", "advisory", "strategy", "management", "business",
			"professional services", "expertise", "solutions",
		},
		"transportation": {
			"transportation", "logistics", "shipping", "delivery", "freight", "trucking",
			"warehouse", "supply chain", "distribution",
		},
	}
}

// DataSourcesHealth proxies health checks for configured enrichment sources
func (c *ClassificationService) DataSourcesHealth(ctx context.Context) []datasource.SourceHealth {
	if c.enricher == nil {
		return nil
	}
	return c.enricher.CheckHealth(ctx)
}

// ClassificationRequest represents a business classification request
type ClassificationRequest struct {
	BusinessName       string `json:"business_name" validate:"required"`
	BusinessType       string `json:"business_type,omitempty"`
	Industry           string `json:"industry,omitempty"`
	Description        string `json:"description,omitempty"`
	Keywords           string `json:"keywords,omitempty"`
	WebsiteURL         string `json:"website_url,omitempty"`
	RegistrationNumber string `json:"registration_number,omitempty"`
	TaxID              string `json:"tax_id,omitempty"`
}

// ClassificationResponse represents a business classification response
type ClassificationResponse struct {
	BusinessID            string                   `json:"business_id"`
	Classifications       []IndustryClassification `json:"classifications"`
	PrimaryClassification *IndustryClassification  `json:"primary_classification"`
	ConfidenceScore       float64                  `json:"confidence_score"`
	ClassificationMethod  string                   `json:"classification_method"`
	ProcessingTime        time.Duration            `json:"processing_time"`
	RawData               map[string]interface{}   `json:"raw_data,omitempty"`
}

// IndustryClassification represents an industry classification result
type IndustryClassification struct {
	IndustryCode         string   `json:"industry_code"`
	IndustryName         string   `json:"industry_name"`
	ConfidenceScore      float64  `json:"confidence_score"`
	ClassificationMethod string   `json:"classification_method"`
	Keywords             []string `json:"keywords,omitempty"`
	Description          string   `json:"description,omitempty"`
}

// BatchClassificationRequest represents a batch classification request
type BatchClassificationRequest struct {
	Businesses []ClassificationRequest `json:"businesses" validate:"required,min=1,max=100"`
}

// BatchClassificationResponse represents a batch classification response
type BatchClassificationResponse struct {
	Results        []ClassificationResponse `json:"results"`
	TotalProcessed int                      `json:"total_processed"`
	SuccessCount   int                      `json:"success_count"`
	ErrorCount     int                      `json:"error_count"`
	ProcessingTime time.Duration            `json:"processing_time"`
}

// ClassifyBusiness classifies a single business
func (c *ClassificationService) ClassifyBusiness(ctx context.Context, req *ClassificationRequest) (*ClassificationResponse, error) {
	start := time.Now()

	// Log classification start
	c.logger.WithComponent("classification").LogBusinessEvent(ctx, "classification_started", "", map[string]interface{}{
		"business_name": req.BusinessName,
		"business_type": req.BusinessType,
		"industry":      req.Industry,
	})

	// Validate request
	if err := c.validateClassificationRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Try cache
	cacheKey := c.makeCacheKey(req)
	if cached, ok := c.getFromCache(cacheKey); ok {
		// Reconstruct response with fresh BusinessID and durations
		primary := c.determinePrimaryClassification(cached)
		confidence := c.calculateOverallConfidence(cached)
		businessID := c.generateBusinessID(req)
		response := &ClassificationResponse{
			BusinessID:            businessID,
			Classifications:       cached,
			PrimaryClassification: primary,
			ConfidenceScore:       confidence,
			ClassificationMethod:  "hybrid",
			ProcessingTime:        time.Since(start),
			RawData: map[string]interface{}{
				"request": req,
				"method":  "hybrid_classification_cache_hit",
			},
		}
		// Log and metrics
		c.logger.WithComponent("classification").LogBusinessEvent(ctx, "classification_completed", businessID, map[string]interface{}{
			"business_name":         req.BusinessName,
			"primary_industry_code": primary.IndustryCode,
			"primary_industry_name": primary.IndustryName,
			"confidence_score":      confidence,
			"processing_time_ms":    time.Since(start).Milliseconds(),
			"total_classifications": len(cached),
			"cache_hit":             true,
		})
		c.metrics.RecordBusinessClassification("success_cache", fmt.Sprintf("%.2f", confidence))
		return response, nil
	}

	// Perform classification (cache miss)
	// Optional enrichment pre-processing
	if c.enricher != nil {
		enrReq := datasource.EnrichmentRequest{
			BusinessName:       req.BusinessName,
			RegistrationNumber: req.RegistrationNumber,
		}
		if enr, err := c.enricher.Enrich(ctx, enrReq); err == nil {
			// sanitize enrichment payload
			enr = validators.CleanEnrichmentResult(enr)
			if enr.CleanBusinessName != "" {
				req.BusinessName = enr.CleanBusinessName
			}
			if enr.Industry != "" && req.Industry == "" {
				req.Industry = enr.Industry
			}
			if enr.Description != "" && req.Description == "" {
				req.Description = enr.Description
			}
			if len(enr.Keywords) > 0 && req.Keywords == "" {
				req.Keywords = strings.Join(enr.Keywords, ",")
			}
		}
	}

	classifications, err := c.performClassification(ctx, req)
	if err != nil {
		c.logger.WithComponent("classification").WithError(err).LogBusinessEvent(ctx, "classification_failed", "", map[string]interface{}{
			"business_name": req.BusinessName,
			"error":         err.Error(),
		})
		c.metrics.RecordBusinessClassification("failed", "error")
		return nil, fmt.Errorf("classification failed: %w", err)
	}

	// Determine primary classification
	primaryClassification := c.determinePrimaryClassification(classifications)

	// Calculate overall confidence score
	confidenceScore := c.calculateOverallConfidence(classifications)

	// Generate business ID for tracking
	businessID := c.generateBusinessID(req)

	// Create response
	response := &ClassificationResponse{
		BusinessID:            businessID,
		Classifications:       classifications,
		PrimaryClassification: primaryClassification,
		ConfidenceScore:       confidenceScore,
		ClassificationMethod:  "hybrid", // Using multiple methods
		ProcessingTime:        time.Since(start),
		RawData: map[string]interface{}{
			"request": req,
			"method":  "hybrid_classification",
		},
	}

	// Store classification in database if available
	if c.db != nil {
		if err := c.storeClassification(ctx, businessID, response); err != nil {
			c.logger.WithComponent("classification").WithError(err).LogBusinessEvent(ctx, "classification_storage_failed", businessID, map[string]interface{}{
				"error": err.Error(),
			})
		}
	}

	// Store in cache
	c.setCache(cacheKey, classifications)

	// Log successful classification
	c.logger.WithComponent("classification").LogBusinessEvent(ctx, "classification_completed", businessID, map[string]interface{}{
		"business_name":         req.BusinessName,
		"primary_industry_code": primaryClassification.IndustryCode,
		"primary_industry_name": primaryClassification.IndustryName,
		"confidence_score":      confidenceScore,
		"processing_time_ms":    time.Since(start).Milliseconds(),
		"total_classifications": len(classifications),
		"cache_hit":             false,
	})

	// Record metrics
	c.metrics.RecordBusinessClassification("success", fmt.Sprintf("%.2f", confidenceScore))
	c.metrics.RecordClassificationDuration("single", time.Since(start))
	// Simple on-host alerting for slow classifications
	if c.logger != nil && c.config != nil {
		thr := 300 * time.Millisecond
		if c.config != nil {
			// use observability thresholds if available via logger's config is not accessible here; keep default 300ms
		}
		if time.Since(start) > thr {
			c.logger.WithComponent("classification").Warn("slow_classification", "duration_ms", time.Since(start).Milliseconds(), "business_name", req.BusinessName)
		}
	}

	return response, nil
}

// ClassifyBusinessesBatch classifies multiple businesses in batch
func (c *ClassificationService) ClassifyBusinessesBatch(ctx context.Context, req *BatchClassificationRequest) (*BatchClassificationResponse, error) {
	start := time.Now()

	// Log batch classification start
	c.logger.WithComponent("classification").LogBusinessEvent(ctx, "batch_classification_started", "", map[string]interface{}{
		"total_businesses": len(req.Businesses),
	})

	// Validate batch size
	if len(req.Businesses) > 100 {
		return nil, fmt.Errorf("batch size exceeds maximum limit of 100")
	}

	// Pre-allocate output array to preserve order
	out := make([]ClassificationResponse, len(req.Businesses))
	successCount := 0
	errorCount := 0

	// Group identical requests by cache key to avoid duplicate work in the same batch
	type batchJob struct {
		key  string
		req  *ClassificationRequest
		idxs []int // all positions in the original slice that share this key
	}

	keyToJob := make(map[string]*batchJob, len(req.Businesses))
	for i := range req.Businesses {
		// copy to avoid aliasing the loop variable
		r := req.Businesses[i]
		key := c.makeCacheKey(&r)
		if bj, ok := keyToJob[key]; ok {
			bj.idxs = append(bj.idxs, i)
			continue
		}
		keyToJob[key] = &batchJob{key: key, req: &r, idxs: []int{i}}
	}

	jobs := make([]*batchJob, 0, len(keyToJob))
	for _, j := range keyToJob {
		jobs = append(jobs, j)
	}

	// Bounded concurrency worker pool
	workerCount := 8
	if len(jobs) < workerCount {
		workerCount = len(jobs)
	}
	if workerCount == 0 {
		// Nothing to process
		resp := &BatchClassificationResponse{
			Results:        out[:0],
			TotalProcessed: 0,
			SuccessCount:   0,
			ErrorCount:     0,
			ProcessingTime: time.Since(start),
		}
		c.logger.WithComponent("classification").LogBusinessEvent(ctx, "batch_classification_completed", "", map[string]interface{}{
			"total_processed":    resp.TotalProcessed,
			"success_count":      resp.SuccessCount,
			"error_count":        resp.ErrorCount,
			"processing_time_ms": resp.ProcessingTime.Milliseconds(),
		})
		return resp, nil
	}

	jobsCh := make(chan *batchJob)
	var wg sync.WaitGroup
	var mu sync.Mutex // protects successCount, errorCount, and writes into out

	worker := func() {
		defer wg.Done()
		for j := range jobsCh {
			// Respect batch context cancellation
			select {
			case <-ctx.Done():
				mu.Lock()
				errorCount += len(j.idxs)
				mu.Unlock()
				continue
			default:
			}

			res, err := c.ClassifyBusiness(ctx, j.req)
			mu.Lock()
			if err != nil {
				// record error for all deduped positions
				errorCount += len(j.idxs)
				// Log one event per job (not per duplicate index)
				c.logger.WithComponent("classification").WithError(err).LogBusinessEvent(ctx, "batch_item_failed", "", map[string]interface{}{
					"business_name": j.req.BusinessName,
					"error":         err.Error(),
					"duplicates":    len(j.idxs),
				})
			} else {
				for _, idx := range j.idxs {
					out[idx] = *res
				}
				successCount += len(j.idxs)
			}
			mu.Unlock()
		}
	}

	wg.Add(workerCount)
	for w := 0; w < workerCount; w++ {
		go worker()
	}
	for _, j := range jobs {
		jobsCh <- j
	}
	close(jobsCh)
	wg.Wait()

	// Filter out zero-value entries (failed items)
	resultsFiltered := make([]ClassificationResponse, 0, len(out))
	for i := range out {
		if out[i].BusinessID != "" || len(out[i].Classifications) > 0 {
			resultsFiltered = append(resultsFiltered, out[i])
		}
	}

	response := &BatchClassificationResponse{
		Results:        resultsFiltered,
		TotalProcessed: len(req.Businesses),
		SuccessCount:   successCount,
		ErrorCount:     errorCount,
		ProcessingTime: time.Since(start),
	}

	// Log batch classification completion
	c.logger.WithComponent("classification").LogBusinessEvent(ctx, "batch_classification_completed", "", map[string]interface{}{
		"total_processed":    response.TotalProcessed,
		"success_count":      response.SuccessCount,
		"error_count":        response.ErrorCount,
		"processing_time_ms": response.ProcessingTime.Milliseconds(),
		"unique_jobs":        len(jobs),
		"dedup_savings":      len(req.Businesses) - len(jobs),
	})

	// Metrics and slow-alert
	c.metrics.RecordClassificationDuration("batch", time.Since(start))
	if c.logger != nil && c.config != nil {
		thr := 300 * time.Millisecond
		if time.Since(start) > thr {
			c.logger.WithComponent("classification").Warn("slow_batch_classification", "duration_ms", time.Since(start).Milliseconds(), "total_businesses", len(req.Businesses))
		}
	}

	return response, nil
}

// GetClassificationHistory retrieves classification history for a business
func (c *ClassificationService) GetClassificationHistory(ctx context.Context, businessID string, limit, offset int) ([]*database.BusinessClassification, error) {
	if c.db == nil {
		return nil, fmt.Errorf("database not available")
	}

	classifications, err := c.db.GetBusinessClassificationsByBusinessID(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve classification history: %w", err)
	}

	// Apply pagination
	if offset >= len(classifications) {
		return []*database.BusinessClassification{}, nil
	}

	end := offset + limit
	if end > len(classifications) {
		end = len(classifications)
	}

	return classifications[offset:end], nil
}

// validateClassificationRequest validates the classification request
func (c *ClassificationService) validateClassificationRequest(req *ClassificationRequest) error {
	if strings.TrimSpace(req.BusinessName) == "" {
		return fmt.Errorf("business name is required")
	}

	if len(req.BusinessName) > 500 {
		return fmt.Errorf("business name too long (max 500 characters)")
	}

	if req.Description != "" && len(req.Description) > 2000 {
		return fmt.Errorf("description too long (max 2000 characters)")
	}

	return nil
}

// performClassification performs the actual classification using multiple methods
func (c *ClassificationService) performClassification(ctx context.Context, req *ClassificationRequest) ([]IndustryClassification, error) {
	var classifications []IndustryClassification

	// Method 0: Hybrid website scraping and API analysis (highest priority when URL provided)
	if req.WebsiteURL != "" && c.hybridScraper != nil {
		if webClassifications := c.classifyByHybridAnalysis(ctx, req); len(webClassifications) > 0 {
			classifications = append(classifications, webClassifications...)
		}
	}

	// Method 0.5: Search-based classification (high priority when no website URL provided)
	if req.WebsiteURL == "" && c.multiSourceSearch != nil && c.searchAnalyzer != nil {
		if searchClassifications := c.classifyBySearchAnalysis(ctx, req); len(searchClassifications) > 0 {
			classifications = append(classifications, searchClassifications...)
		}
	}

	// Method 1: Keyword-based classification
	if keywordClassifications := c.classifyByKeywords(req); len(keywordClassifications) > 0 {
		classifications = append(classifications, keywordClassifications...)
	}

	// Method 2: Business type classification
	if businessTypeClassifications := c.classifyByBusinessType(req); len(businessTypeClassifications) > 0 {
		classifications = append(classifications, businessTypeClassifications...)
	}

	// Method 3: Industry-based classification
	if industryClassifications := c.classifyByIndustry(req); len(industryClassifications) > 0 {
		classifications = append(classifications, industryClassifications...)
	}

	// Fallback A: Free-text industry mapping via NAICS dataset when simple mapping misses
	if industryTextClassifications := c.classifyByIndustryText(req); len(industryTextClassifications) > 0 {
		classifications = append(classifications, industryTextClassifications...)
	}

	// Method 4: Name-based classification
	if nameClassifications := c.classifyByName(req); len(nameClassifications) > 0 {
		classifications = append(classifications, nameClassifications...)
	}

	// Method 5: Fuzzy matching across industry datasets
	if fuzzyClassifications := c.classifyByFuzzy(req); len(fuzzyClassifications) > 0 {
		classifications = append(classifications, fuzzyClassifications...)
	}

	// Fallback B: History-based fallback using prior classifications from DB
	if historyClassifications := c.classifyByHistory(ctx, req); len(historyClassifications) > 0 {
		classifications = append(classifications, historyClassifications...)
	}

	// Expand with crosswalk mapping to surface related code systems for the primary NAICS
	if c.industryData != nil {
		// Determine primary classification and enrich with MCC/SIC mappings
		primary := c.determinePrimaryClassification(classifications)
		if primary != nil {
			// Only crosswalk when NAICS-style code (simple heuristic: numeric code of length 6)
			if len(primary.IndustryCode) == 6 {
				mcc, sic := crosswalkFromNAICS(primary.IndustryCode, c.industryData)
				for _, code := range mcc {
					classifications = append(classifications, IndustryClassification{
						IndustryCode:         code,
						IndustryName:         c.industryData.GetMCCDescription(code),
						ConfidenceScore:      primary.ConfidenceScore * 0.8,
						ClassificationMethod: "crosswalk_mcc_from_naics",
					})
				}
				for _, code := range sic {
					classifications = append(classifications, IndustryClassification{
						IndustryCode:         code,
						IndustryName:         c.industryData.GetSICDescription(code),
						ConfidenceScore:      primary.ConfidenceScore * 0.75,
						ClassificationMethod: "crosswalk_sic_from_naics",
					})
				}
			}
		}
	}

	// If no classifications found, return default
	if len(classifications) == 0 {
		classifications = append(classifications, c.getDefaultClassification())
	}

	// Normalize and enhance confidence scores, and deduplicate by code
	classifications = c.postProcessConfidence(classifications)

	return classifications, nil
}

// Industry-Specific Mapping Methods

// ClassifyIndustry performs industry-specific classification
func (c *ClassificationService) ClassifyIndustry(ctx context.Context, businessInfo map[string]interface{}) (*IndustryClassificationResult, error) {
	if c.industryMapper == nil {
		return nil, fmt.Errorf("industry mapper not initialized")
	}

	return c.industryMapper.ClassifyIndustry(ctx, businessInfo)
}

// AddIndustryMapping adds a new industry mapping
func (c *ClassificationService) AddIndustryMapping(ctx context.Context, mapping *IndustryMapping) error {
	if c.industryMapper == nil {
		return fmt.Errorf("industry mapper not initialized")
	}

	return c.industryMapper.AddIndustryMapping(ctx, mapping)
}

// GetIndustryMapping retrieves an industry mapping by type
func (c *ClassificationService) GetIndustryMapping(ctx context.Context, industryType IndustryType) (*IndustryMapping, error) {
	if c.industryMapper == nil {
		return nil, fmt.Errorf("industry mapper not initialized")
	}

	return c.industryMapper.GetIndustryMapping(ctx, industryType)
}

// UpdateIndustryMapping updates an existing industry mapping
func (c *ClassificationService) UpdateIndustryMapping(ctx context.Context, industryType IndustryType, updates map[string]interface{}) error {
	if c.industryMapper == nil {
		return fmt.Errorf("industry mapper not initialized")
	}

	return c.industryMapper.UpdateIndustryMapping(ctx, industryType, updates)
}

// ListIndustryMappings returns all industry mappings
func (c *ClassificationService) ListIndustryMappings(ctx context.Context) ([]*IndustryMapping, error) {
	if c.industryMapper == nil {
		return nil, fmt.Errorf("industry mapper not initialized")
	}

	return c.industryMapper.ListIndustryMappings(ctx)
}

// GetIndustryMapperStats returns statistics about the industry mapper
func (c *ClassificationService) GetIndustryMapperStats() map[string]interface{} {
	if c.industryMapper == nil {
		return map[string]interface{}{
			"error": "Industry mapper not initialized",
		}
	}

	return c.industryMapper.GetMapperStats()
}

// Enhanced classification with industry-specific integration
func (c *ClassificationService) ClassifyWithIndustryIntegration(ctx context.Context, req *ClassificationRequest) ([]IndustryClassification, error) {
	// Perform standard classification first
	response, err := c.ClassifyBusiness(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("standard classification failed: %w", err)
	}

	classifications := response.Classifications

	// Apply industry-specific classification if industry mapper is available
	if c.industryMapper != nil {
		businessInfo := map[string]interface{}{
			"business_name": req.BusinessName,
			"business_type": req.BusinessType,
			"industry":      req.Industry,
			"description":   req.Description,
			"keywords":      req.Keywords,
		}

		industryResult, err := c.industryMapper.ClassifyIndustry(ctx, businessInfo)
		if err == nil && industryResult != nil {
			// Create enhanced classification with industry-specific data
			enhancedClassification := IndustryClassification{
				IndustryCode:         industryResult.PrimaryCode,
				IndustryName:         industryResult.IndustryName,
				ConfidenceScore:      industryResult.ConfidenceScore,
				ClassificationMethod: "industry_specific",
				Description:          fmt.Sprintf("Industry-specific classification for %s (confidence: %.2f)", industryResult.IndustryType, industryResult.ConfidenceScore),
			}

			// Log industry integration
			if c.logger != nil {
				c.logger.WithComponent("classification").LogBusinessEvent(ctx, "industry_integration_applied", "", map[string]interface{}{
					"business_name":     req.BusinessName,
					"industry_type":     string(industryResult.IndustryType),
					"industry_confidence": industryResult.ConfidenceScore,
					"rules_applied":     len(industryResult.AppliedRules),
				})
			}

			// Return enhanced classification
			return []IndustryClassification{enhancedClassification}, nil
		}
	}

	return classifications, nil
}

// postProcessConfidence applies method-based weighting, agreement boosts, confidence range validation, and deduplicates by industry code.
func (c *ClassificationService) postProcessConfidence(classifications []IndustryClassification) []IndustryClassification {
	if len(classifications) == 0 {
		return classifications
	}

	// Count occurrences per code to detect agreement
	occurrences := make(map[string]int)
	for _, cl := range classifications {
		occurrences[cl.IndustryCode]++
	}

	// Keep best per code after weighting and confidence range validation
	bestByCode := make(map[string]IndustryClassification)
	for _, cl := range classifications {
		weight := methodWeightFor(cl.ClassificationMethod)
		agreeBoost := 0.0
		if n := occurrences[cl.IndustryCode]; n > 1 {
			// modest boost for multi-method/multi-hit agreement (cap at 0.2)
			agreeBoost = 0.08 * float64(n-1)
			if agreeBoost > 0.2 {
				agreeBoost = 0.2
			}
		}
		score := cl.ConfidenceScore*weight + agreeBoost
		if score > 0.99 {
			score = 0.99
		}
		cl.ConfidenceScore = score

		// Apply confidence range validation
		cl = c.applyConfidenceRangeValidation(cl)

		cur, exists := bestByCode[cl.IndustryCode]
		if !exists || cl.ConfidenceScore > cur.ConfidenceScore {
			bestByCode[cl.IndustryCode] = cl
		}
	}

	out := make([]IndustryClassification, 0, len(bestByCode))
	for _, v := range bestByCode {
		out = append(out, v)
	}

	// Apply confidence-based result ranking
	out = c.applyConfidenceBasedRanking(out)

	return out
}

// applyConfidenceRangeValidation applies confidence range validation to a classification
func (c *ClassificationService) applyConfidenceRangeValidation(cl IndustryClassification) IndustryClassification {
	// Define method-based confidence ranges (as specified in PRD)
	methodConfidenceRanges := map[string]struct {
		min, max float64
	}{
		"website_analysis":           {0.85, 0.95},
		"website_content_analysis":   {0.85, 0.95},
		"website_structure_analysis": {0.85, 0.95},
		"web_search_analysis":        {0.75, 0.85},
		"search_analysis":            {0.75, 0.85},
		"search_result_analysis":     {0.70, 0.80},
		"keyword_based":              {0.60, 0.75},
		"keyword_based_naics":        {0.60, 0.75},
		"keyword_based_mcc":          {0.60, 0.75},
		"keyword_based_sic":          {0.60, 0.75},
		"fuzzy_matching":             {0.50, 0.70},
		"fuzzy_naics_fulltext":       {0.50, 0.70},
		"fuzzy_naics_token":          {0.50, 0.70},
		"fuzzy_mcc_fulltext":         {0.50, 0.70},
		"fuzzy_mcc_token":            {0.50, 0.70},
		"fuzzy_sic_fulltext":         {0.50, 0.70},
		"fuzzy_sic_token":            {0.50, 0.70},
		"crosswalk_mapping":          {0.40, 0.60},
		"crosswalk_mcc_from_naics":   {0.40, 0.60},
		"crosswalk_sic_from_naics":   {0.40, 0.60},
	}

	// Get confidence range for the method
	range_, exists := methodConfidenceRanges[cl.ClassificationMethod]
	if !exists {
		// Default range for unknown methods
		range_ = struct {
			min, max float64
		}{0.50, 0.70}
	}

	// Ensure confidence score is within the valid range
	if cl.ConfidenceScore < range_.min {
		cl.ConfidenceScore = range_.min
	} else if cl.ConfidenceScore > range_.max {
		cl.ConfidenceScore = range_.max
	}

	// Log confidence range validation
	if c.logger != nil {
		c.logger.WithComponent("classification").Debug("confidence_range_validation", map[string]interface{}{
			"method":         cl.ClassificationMethod,
			"original_score": cl.ConfidenceScore,
			"min_confidence": range_.min,
			"max_confidence": range_.max,
			"final_score":    cl.ConfidenceScore,
			"industry_code":  cl.IndustryCode,
		})
	}

	return cl
}

// applyConfidenceBasedRanking applies confidence-based ranking to classifications
func (c *ClassificationService) applyConfidenceBasedRanking(classifications []IndustryClassification) []IndustryClassification {
	if len(classifications) == 0 {
		return classifications
	}

	// Sort classifications by confidence score in descending order
	// This ensures that higher confidence classifications appear first
	for i := 0; i < len(classifications)-1; i++ {
		for j := i + 1; j < len(classifications); j++ {
			if classifications[i].ConfidenceScore < classifications[j].ConfidenceScore {
				classifications[i], classifications[j] = classifications[j], classifications[i]
			}
		}
	}

	// Log confidence-based ranking
	if c.logger != nil && len(classifications) > 0 {
		c.logger.WithComponent("classification").Debug("confidence_based_ranking_applied", map[string]interface{}{
			"total_classifications": len(classifications),
			"top_confidence":        classifications[0].ConfidenceScore,
			"top_method":            classifications[0].ClassificationMethod,
			"top_industry_code":     classifications[0].IndustryCode,
		})
	}

	return classifications
}

// methodWeightFor returns a multiplicative weight reflecting method reliability based on confidence ranges.
func methodWeightFor(method string) float64 {
	switch method {
	// Website analysis methods (highest confidence: 0.85-0.95)
	case "website_analysis":
		return 1.20
	case "website_content_analysis":
		return 1.18
	case "website_structure_analysis":
		return 1.16

	// Web search analysis methods (high confidence: 0.75-0.85)
	case "web_search_analysis":
		return 1.15
	case "search_analysis":
		return 1.14
	case "search_result_analysis":
		return 1.13
	case "multi_source_search":
		return 1.12

	// Keyword-based methods (moderate confidence: 0.60-0.75)
	case "keyword_based_naics":
		return 1.10
	case "keyword_based":
		return 1.08
	case "keyword_based_mcc":
		return 1.07
	case "keyword_based_sic":
		return 1.06
	case "industry_based":
		return 1.05
	case "business_type_based":
		return 1.04

	// Fuzzy matching methods (lower confidence: 0.50-0.70)
	case "fuzzy_naics_fulltext", "fuzzy_naics_token":
		return 1.02
	case "fuzzy_mcc_fulltext", "fuzzy_mcc_token":
		return 1.01
	case "fuzzy_sic_fulltext", "fuzzy_sic_token":
		return 1.00
	case "name_pattern_based":
		return 0.98

	// Crosswalk mapping methods (lowest confidence: 0.40-0.60)
	case "crosswalk_mcc_from_naics", "crosswalk_sic_from_naics":
		return 0.95
	case "crosswalk_mapping":
		return 0.93

	// Default weight for unknown methods
	default:
		return 1.0
	}
}

// classifyByKeywords classifies business based on keywords
func (c *ClassificationService) classifyByKeywords(req *ClassificationRequest) []IndustryClassification {
	var classifications []IndustryClassification

	// Normalize and combine fields for robust matching
	textToSearch, tokens := normalizeBusinessFields(req.BusinessName, req.Description, req.Keywords)

	// Use real industry data if available
	if c.industryData != nil {
		// Token-wise search with deduplication per code system
		naicsSeen := make(map[string]struct{})
		mccSeen := make(map[string]struct{})
		sicSeen := make(map[string]struct{})

		for _, tok := range tokens {
			if len(tok) < 3 {
				continue
			}

			// NAICS
			for _, code := range c.industryData.SearchNAICSByKeyword(tok) {
				if _, exists := naicsSeen[code]; exists {
					continue
				}
				naicsSeen[code] = struct{}{}
				classifications = append(classifications, IndustryClassification{
					IndustryCode:         code,
					IndustryName:         c.industryData.GetNAICSName(code),
					ConfidenceScore:      0.7,
					ClassificationMethod: "keyword_based_naics",
					Keywords:             []string{tok},
				})
			}

			// MCC
			for _, code := range c.industryData.SearchMCCByKeyword(tok) {
				if _, exists := mccSeen[code]; exists {
					continue
				}
				mccSeen[code] = struct{}{}
				classifications = append(classifications, IndustryClassification{
					IndustryCode:         code,
					IndustryName:         c.industryData.GetMCCDescription(code),
					ConfidenceScore:      0.6,
					ClassificationMethod: "keyword_based_mcc",
					Keywords:             []string{tok},
				})
			}

			// SIC
			for _, code := range c.industryData.SearchSICByKeyword(tok) {
				if _, exists := sicSeen[code]; exists {
					continue
				}
				sicSeen[code] = struct{}{}
				classifications = append(classifications, IndustryClassification{
					IndustryCode:         code,
					IndustryName:         c.industryData.GetSICDescription(code),
					ConfidenceScore:      0.5,
					ClassificationMethod: "keyword_based_sic",
					Keywords:             []string{tok},
				})
			}
		}
	} else {
		// Fallback to hardcoded mappings if no industry data available
		keywordMappings := map[string][]string{
			"software":       {"541511", "541512", "541519"},
			"technology":     {"541511", "541512", "541519", "541715"},
			"consulting":     {"541611", "541612", "541618", "541690"},
			"financial":      {"522110", "522120", "522130", "522190", "523150"},
			"healthcare":     {"621111", "621112", "621210", "621310", "621320"},
			"retail":         {"441110", "442110", "443141", "444110", "445110"},
			"manufacturing":  {"332996", "332999", "333415", "334110", "335110"},
			"construction":   {"236115", "236116", "236117", "236118", "236220"},
			"transportation": {"484110", "484121", "484122", "484210", "485110"},
			"education":      {"611110", "611210", "611310", "611410", "611420"},
		}

		for keyword, industryCodes := range keywordMappings {
			if strings.Contains(textToSearch, keyword) {
				for _, code := range industryCodes {
					classifications = append(classifications, IndustryClassification{
						IndustryCode:         code,
						IndustryName:         c.getIndustryName(code),
						ConfidenceScore:      0.7,
						ClassificationMethod: "keyword_based",
						Keywords:             []string{keyword},
					})
				}
			}
		}
	}

	return classifications
}

// classifyByBusinessType classifies business based on business type
func (c *ClassificationService) classifyByBusinessType(req *ClassificationRequest) []IndustryClassification {
	if req.BusinessType == "" {
		return nil
	}

	businessTypeMappings := map[string]string{
		"llc":                 "541611", // Management consulting
		"corporation":         "541611", // Management consulting
		"partnership":         "541611", // Management consulting
		"sole_proprietorship": "541611", // Management consulting
		"nonprofit":           "813211", // Grantmaking foundations
		"charity":             "813211", // Grantmaking foundations
		"foundation":          "813211", // Grantmaking foundations
	}

	if code, exists := businessTypeMappings[strings.ToLower(req.BusinessType)]; exists {
		return []IndustryClassification{
			{
				IndustryCode:         code,
				IndustryName:         c.getIndustryName(code),
				ConfidenceScore:      0.8,
				ClassificationMethod: "business_type_based",
				Description:          fmt.Sprintf("Classified based on business type: %s", req.BusinessType),
			},
		}
	}

	return nil
}

// classifyByIndustry classifies business based on provided industry
func (c *ClassificationService) classifyByIndustry(req *ClassificationRequest) []IndustryClassification {
	if req.Industry == "" {
		return nil
	}

	industryMappings := map[string]string{
		"technology":     "541511",
		"software":       "541511",
		"consulting":     "541611",
		"finance":        "522110",
		"healthcare":     "621111",
		"retail":         "441110",
		"manufacturing":  "332996",
		"construction":   "236115",
		"transportation": "484110",
		"education":      "611110",
		"real_estate":    "531110",
		"legal":          "541110",
		"accounting":     "541211",
		"marketing":      "541810",
		"advertising":    "541810",
	}

	if code, exists := industryMappings[strings.ToLower(req.Industry)]; exists {
		return []IndustryClassification{
			{
				IndustryCode:         code,
				IndustryName:         c.getIndustryName(code),
				ConfidenceScore:      0.9,
				ClassificationMethod: "industry_based",
				Description:          fmt.Sprintf("Classified based on industry: %s", req.Industry),
			},
		}
	}

	return nil
}

// classifyByIndustryText uses free-text industry label mapping via NAICS dataset as a fallback
func (c *ClassificationService) classifyByIndustryText(req *ClassificationRequest) []IndustryClassification {
	if req.Industry == "" || c.industryData == nil {
		return nil
	}
	codes := mapIndustryTextToNAICS(req.Industry, c.industryData)
	if len(codes) == 0 {
		return nil
	}
	out := make([]IndustryClassification, 0, len(codes))
	for _, code := range codes {
		out = append(out, IndustryClassification{
			IndustryCode:         code,
			IndustryName:         c.industryData.GetNAICSName(code),
			ConfidenceScore:      0.68,
			ClassificationMethod: "industry_text_mapping",
			Description:          "Mapped from free-text industry label",
		})
	}
	return out
}

// classifyByHistory falls back to previous classifications stored in the database for the same business
func (c *ClassificationService) classifyByHistory(ctx context.Context, req *ClassificationRequest) []IndustryClassification {
	if c.db == nil {
		return nil
	}

	// 1) Try by registration number
	if req.RegistrationNumber != "" {
		if b, err := c.db.GetBusinessByRegistrationNumber(ctx, req.RegistrationNumber); err == nil && b != nil {
			if cl := c.latestClassificationForBusiness(ctx, b.ID); cl != nil {
				return []IndustryClassification{{
					IndustryCode:         cl.IndustryCode,
					IndustryName:         cl.IndustryName,
					ConfidenceScore:      minFloat(0.75, cl.ConfidenceScore),
					ClassificationMethod: "history_fallback",
					Description:          "Reused most recent prior classification by registration number",
				}}
			}
			// If no history but business has industry fields, use them
			if b.IndustryCode != "" {
				return []IndustryClassification{{
					IndustryCode:         b.IndustryCode,
					IndustryName:         c.getIndustryName(b.IndustryCode),
					ConfidenceScore:      0.6,
					ClassificationMethod: "history_business_profile",
					Description:          "Derived from stored business profile",
				}}
			}
		}
	}

	// 2) Try by searching businesses by name and reuse their last classification
	name := strings.TrimSpace(req.BusinessName)
	if name == "" {
		return nil
	}
	if matches, err := c.db.SearchBusinesses(ctx, name, 1, 0); err == nil && len(matches) > 0 {
		b := matches[0]
		if cl := c.latestClassificationForBusiness(ctx, b.ID); cl != nil {
			return []IndustryClassification{{
				IndustryCode:         cl.IndustryCode,
				IndustryName:         cl.IndustryName,
				ConfidenceScore:      minFloat(0.7, cl.ConfidenceScore),
				ClassificationMethod: "history_fallback_name_match",
				Description:          "Reused most recent prior classification by business name",
			}}
		}
		if b.IndustryCode != "" {
			return []IndustryClassification{{
				IndustryCode:         b.IndustryCode,
				IndustryName:         c.getIndustryName(b.IndustryCode),
				ConfidenceScore:      0.55,
				ClassificationMethod: "history_business_profile_name_match",
				Description:          "Derived from stored business profile (name match)",
			}}
		}
	}

	return nil
}

func (c *ClassificationService) latestClassificationForBusiness(ctx context.Context, businessID string) *database.BusinessClassification {
	cls, err := c.db.GetBusinessClassificationsByBusinessID(ctx, businessID)
	if err != nil || len(cls) == 0 {
		return nil
	}
	latest := cls[0]
	for _, v := range cls[1:] {
		if v.CreatedAt.After(latest.CreatedAt) {
			latest = v
		}
	}
	return latest
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// classifyByName classifies business based on business name patterns
func (c *ClassificationService) classifyByName(req *ClassificationRequest) []IndustryClassification {
	name, _ := normalizeBusinessFields(req.BusinessName, "", "")

	// Define name pattern mappings
	namePatterns := map[string]string{
		"tech":          "541511",
		"software":      "541511",
		"systems":       "541511",
		"consult":       "541611",
		"advisory":      "541611",
		"financial":     "522110",
		"bank":          "522110",
		"credit":        "522110",
		"medical":       "621111",
		"health":        "621111",
		"clinic":        "621111",
		"store":         "441110",
		"shop":          "441110",
		"market":        "441110",
		"factory":       "332996",
		"manufacturing": "332996",
		"build":         "236115",
		"construction":  "236115",
		"transport":     "484110",
		"logistics":     "484110",
		"school":        "611110",
		"university":    "611110",
		"college":       "611110",
		"realty":        "531110",
		"properties":    "531110",
		"law":           "541110",
		"legal":         "541110",
		"accounting":    "541211",
		"cpa":           "541211",
		"marketing":     "541810",
		"advertising":   "541810",
	}

	for pattern, code := range namePatterns {
		if strings.Contains(name, pattern) {
			return []IndustryClassification{
				{
					IndustryCode:         code,
					IndustryName:         c.getIndustryName(code),
					ConfidenceScore:      0.6,
					ClassificationMethod: "name_pattern_based",
					Description:          fmt.Sprintf("Classified based on name pattern: %s", pattern),
				},
			}
		}
	}

	return nil
}

// classifyByFuzzy applies fuzzy matching on business name/description/keywords against industry datasets.
// It leverages token and full-text similarity to identify likely industries even when exact keywords do not match.
func (c *ClassificationService) classifyByFuzzy(req *ClassificationRequest) []IndustryClassification {
	if c.industryData == nil {
		return nil
	}

	normalized, tokens := normalizeBusinessFields(req.BusinessName, req.Description, req.Keywords)
	if normalized == "" {
		return nil
	}

	// Thresholds tuned for precision>recall initially; can be adjusted with config later
	const naicsThreshold = 0.82
	const mccThreshold = 0.85
	const sicThreshold = 0.85

	naicsSeen := make(map[string]struct{})
	mccSeen := make(map[string]struct{})
	sicSeen := make(map[string]struct{})

	var out []IndustryClassification

	// Full-text pass (captures multi-word semantics)
	for _, code := range c.industryData.SearchNAICSByFuzzy(normalized, naicsThreshold) {
		if _, exists := naicsSeen[code]; exists {
			continue
		}
		naicsSeen[code] = struct{}{}
		out = append(out, IndustryClassification{
			IndustryCode:         code,
			IndustryName:         c.industryData.GetNAICSName(code),
			ConfidenceScore:      0.65,
			ClassificationMethod: "fuzzy_naics_fulltext",
		})
	}
	for _, code := range c.industryData.SearchMCCByFuzzy(normalized, mccThreshold) {
		if _, exists := mccSeen[code]; exists {
			continue
		}
		mccSeen[code] = struct{}{}
		out = append(out, IndustryClassification{
			IndustryCode:         code,
			IndustryName:         c.industryData.GetMCCDescription(code),
			ConfidenceScore:      0.55,
			ClassificationMethod: "fuzzy_mcc_fulltext",
		})
	}
	for _, code := range c.industryData.SearchSICByFuzzy(normalized, sicThreshold) {
		if _, exists := sicSeen[code]; exists {
			continue
		}
		sicSeen[code] = struct{}{}
		out = append(out, IndustryClassification{
			IndustryCode:         code,
			IndustryName:         c.industryData.GetSICDescription(code),
			ConfidenceScore:      0.5,
			ClassificationMethod: "fuzzy_sic_fulltext",
		})
	}

	// Token pass to catch strong token-specific signals
	for _, tok := range tokens {
		if len(tok) < 3 {
			continue
		}
		for _, code := range c.industryData.SearchNAICSByFuzzy(tok, naicsThreshold) {
			if _, exists := naicsSeen[code]; exists {
				continue
			}
			naicsSeen[code] = struct{}{}
			out = append(out, IndustryClassification{
				IndustryCode:         code,
				IndustryName:         c.industryData.GetNAICSName(code),
				ConfidenceScore:      0.62,
				ClassificationMethod: "fuzzy_naics_token",
				Keywords:             []string{tok},
			})
		}
		for _, code := range c.industryData.SearchMCCByFuzzy(tok, mccThreshold) {
			if _, exists := mccSeen[code]; exists {
				continue
			}
			mccSeen[code] = struct{}{}
			out = append(out, IndustryClassification{
				IndustryCode:         code,
				IndustryName:         c.industryData.GetMCCDescription(code),
				ConfidenceScore:      0.52,
				ClassificationMethod: "fuzzy_mcc_token",
				Keywords:             []string{tok},
			})
		}
		for _, code := range c.industryData.SearchSICByFuzzy(tok, sicThreshold) {
			if _, exists := sicSeen[code]; exists {
				continue
			}
			sicSeen[code] = struct{}{}
			out = append(out, IndustryClassification{
				IndustryCode:         code,
				IndustryName:         c.industryData.GetSICDescription(code),
				ConfidenceScore:      0.48,
				ClassificationMethod: "fuzzy_sic_token",
				Keywords:             []string{tok},
			})
		}
	}

	return out
}

// determinePrimaryClassification determines the primary classification from multiple results
func (c *ClassificationService) determinePrimaryClassification(classifications []IndustryClassification) *IndustryClassification {
	if len(classifications) == 0 {
		return nil
	}

	// Find the classification with the highest confidence score
	var primary *IndustryClassification
	highestConfidence := 0.0

	for i := range classifications {
		if classifications[i].ConfidenceScore > highestConfidence {
			highestConfidence = classifications[i].ConfidenceScore
			primary = &classifications[i]
		}
	}

	return primary
}

// calculateOverallConfidence calculates the overall confidence score
func (c *ClassificationService) calculateOverallConfidence(classifications []IndustryClassification) float64 {
	if len(classifications) == 0 {
		return 0.0
	}

	totalConfidence := 0.0
	for _, classification := range classifications {
		totalConfidence += classification.ConfidenceScore
	}

	return totalConfidence / float64(len(classifications))
}

// getIndustryName returns the industry name for a given NAICS code
func (c *ClassificationService) getIndustryName(code string) string {
	// Use real industry data if available
	if c.industryData != nil {
		return c.industryData.GetNAICSName(code)
	}

	// Fallback to hardcoded mappings
	industryNames := map[string]string{
		"541511": "Custom Computer Programming Services",
		"541512": "Computer Systems Design Services",
		"541519": "Other Computer Related Services",
		"541611": "Administrative Management and General Management Consulting Services",
		"541612": "Human Resources Consulting Services",
		"541618": "Other Management Consulting Services",
		"541690": "Other Scientific and Technical Consulting Services",
		"541715": "Research and Development in the Physical, Engineering, and Life Sciences",
		"522110": "Commercial Banking",
		"522120": "Savings Institutions",
		"522130": "Credit Unions",
		"522190": "Other Depository Credit Intermediation",
		"523150": "Securities and Commodity Exchanges",
		"621111": "Offices of Physicians (except Mental Health Specialists)",
		"621112": "Offices of Physicians, Mental Health Specialists",
		"621210": "Offices of Dentists",
		"621310": "Offices of Chiropractors",
		"621320": "Offices of Optometrists",
		"441110": "New Car Dealers",
		"442110": "Furniture Stores",
		"443141": "Household Appliance Stores",
		"444110": "Home Centers",
		"445110": "Supermarkets and Other Grocery (except Convenience) Stores",
		"332996": "Fabricated Pipe and Pipe Fitting Manufacturing",
		"332999": "Miscellaneous Fabricated Metal Product Manufacturing",
		"333415": "Air-Conditioning and Warm Air Heating Equipment and Commercial and Industrial Refrigeration Equipment Manufacturing",
		"334110": "Computer and Peripheral Equipment Manufacturing",
		"335110": "Electric Lamp Bulb and Part Manufacturing",
		"236115": "New Single-Family Housing Construction (except For-Sale Builders)",
		"236116": "New Multifamily Housing Construction (except For-Sale Builders)",
		"236117": "New Housing For-Sale Builders",
		"236118": "Residential Remodelers",
		"236220": "Commercial Building Construction",
		"484110": "General Freight Trucking, Local",
		"484121": "General Freight Trucking, Long-Distance, Truckload",
		"484122": "General Freight Trucking, Long-Distance, Less Than Truckload",
		"484210": "Used Household and Office Goods Moving",
		"485110": "Urban Transit Systems",
		"611110": "Elementary and Secondary Schools",
		"611210": "Junior Colleges",
		"611310": "Colleges, Universities, and Professional Schools",
		"611410": "Business and Secretarial Schools",
		"611420": "Computer Training",
		"531110": "Lessors of Residential Buildings and Dwellings",
		"541110": "Offices of Lawyers",
		"541211": "Offices of Certified Public Accountants",
		"541810": "Advertising Agencies",
		"813211": "Grantmaking Foundations",
	}

	if name, exists := industryNames[code]; exists {
		return name
	}

	return "Unknown Industry"
}

// getDefaultClassification returns a default classification
func (c *ClassificationService) getDefaultClassification() IndustryClassification {
	return IndustryClassification{
		IndustryCode:         "541611",
		IndustryName:         "Administrative Management and General Management Consulting Services",
		ConfidenceScore:      0.3,
		ClassificationMethod: "default",
		Description:          "Default classification applied when no specific classification could be determined",
	}
}

// classifyByHybridAnalysis performs classification using hybrid scraping and API analysis
func (c *ClassificationService) classifyByHybridAnalysis(ctx context.Context, req *ClassificationRequest) []IndustryClassification {
	if c.hybridScraper == nil {
		return nil
	}

	// Convert our request to hybrid scraping request format
	hybridReq := &webanalysis.HybridScrapingRequest{
		BusinessName: req.BusinessName,
		WebsiteURL:   req.WebsiteURL,
		Priority:     "medium",
		Budget:       0.50, // $0.50 max budget
		ContactEmail: "contact@kyb-platform.com",
	}

	// Perform hybrid analysis
	result, err := c.hybridScraper.ScrapeBusiness(ctx, hybridReq)
	if err != nil {
		c.logger.WithComponent("classification").Warn("Hybrid analysis failed", "error", err.Error(), "website_url", req.WebsiteURL)
		return nil
	}

	// Convert hybrid results to our format
	var classifications []IndustryClassification

	// Add classifications from scraped content
	if result.ScrapedContent != nil && result.ScrapedContent.Text != "" {
		// Use keyword-based classification on scraped text
		scrapedReq := &ClassificationRequest{
			BusinessName: req.BusinessName,
			Description:  result.ScrapedContent.Text,
		}
		scrapedClassifications := c.classifyByKeywords(scrapedReq)
		for _, classification := range scrapedClassifications {
			classification.ClassificationMethod = "hybrid_website_analysis"
			classification.ConfidenceScore *= 0.9 // Slightly reduce confidence for hybrid method
			classifications = append(classifications, classification)
		}
	}

	// Add classifications from API data
	if result.APIData != nil {
		apiReq := &ClassificationRequest{
			BusinessName: req.BusinessName,
			Description:  result.APIData.Description,
			Industry:     result.APIData.Industry,
		}
		apiClassifications := c.classifyByKeywords(apiReq)
		for _, classification := range apiClassifications {
			classification.ClassificationMethod = "hybrid_api_analysis"
			classification.ConfidenceScore *= result.APIData.Confidence // Weight by API confidence
			classifications = append(classifications, classification)
		}
	}

	return classifications
}

// classifyByWebsiteAnalysis performs classification using website scraping and analysis
func (c *ClassificationService) classifyByWebsiteAnalysis(ctx context.Context, req *ClassificationRequest) []IndustryClassification {
	if c.webAnalysis == nil {
		return nil
	}

	// Convert our request to webanalysis request format
	webReq := &webanalysis.ClassificationRequest{
		BusinessName:                req.BusinessName,
		BusinessType:                req.BusinessType,
		Industry:                    req.Industry,
		WebsiteURL:                  req.WebsiteURL,
		MaxResults:                  10,
		ConfidenceThreshold:         0.5,
		IncludeRiskAnalysis:         false,
		IncludeConnectionValidation: false,
	}

	// Perform web analysis
	result, err := c.webAnalysis.ClassifyBusiness(ctx, webReq)
	if err != nil {
		c.logger.WithComponent("classification").Warn("Website analysis failed", "error", err.Error(), "website_url", req.WebsiteURL)
		return nil
	}

	// Convert webanalysis results to our format
	var classifications []IndustryClassification
	for _, industry := range result.Industries {
		classification := IndustryClassification{
			IndustryCode:         industry.NAICSCode,
			IndustryName:         industry.Industry,
			ConfidenceScore:      industry.Confidence,
			ClassificationMethod: "website_analysis",
			Description:          industry.Evidence,
		}
		classifications = append(classifications, classification)
	}

	return classifications
}

// generateBusinessID generates a unique business ID
func (c *ClassificationService) generateBusinessID(req *ClassificationRequest) string {
	// In a real implementation, this would generate a proper UUID
	// For now, we'll create a simple hash-based ID
	return fmt.Sprintf("business_%d", time.Now().UnixNano())
}

// storeClassification stores the classification result in the database
func (c *ClassificationService) storeClassification(ctx context.Context, businessID string, response *ClassificationResponse) error {
	if response.PrimaryClassification == nil {
		return fmt.Errorf("no primary classification to store")
	}

	classification := &database.BusinessClassification{
		ID:                   fmt.Sprintf("classification_%d", time.Now().UnixNano()),
		BusinessID:           businessID,
		IndustryCode:         response.PrimaryClassification.IndustryCode,
		IndustryName:         response.PrimaryClassification.IndustryName,
		ConfidenceScore:      response.PrimaryClassification.ConfidenceScore,
		ClassificationMethod: response.PrimaryClassification.ClassificationMethod,
		Source:               "internal_classifier",
		RawData:              fmt.Sprintf("%+v", response.RawData),
		CreatedAt:            time.Now(),
	}

	return c.db.CreateBusinessClassification(ctx, classification)
}

// classifyBySearchAnalysis performs classification using search-based analysis
func (c *ClassificationService) classifyBySearchAnalysis(ctx context.Context, req *ClassificationRequest) []IndustryClassification {
	if c.multiSourceSearch == nil || c.searchAnalyzer == nil {
		return nil
	}

	// Create search query from business information
	searchQuery := c.buildSearchQuery(req)

	// Perform search
	searchResponse, err := c.multiSourceSearch.Search(ctx, searchQuery, req.BusinessName)
	if err != nil {
		c.logger.WithComponent("classification").Warn("Search analysis failed", "error", err.Error(), "business_name", req.BusinessName)
		return nil
	}

	// Analyze search results
	analysisResult, err := c.searchAnalyzer.AnalyzeSearchResults(ctx, searchResponse.Results, req.BusinessName)
	if err != nil {
		c.logger.WithComponent("classification").Warn("Search result analysis failed", "error", err.Error(), "business_name", req.BusinessName)
		return nil
	}

	// Convert analysis results to classifications
	var classifications []IndustryClassification
	for industry, confidence := range analysisResult.IndustryIndicators {
		// Map industry names to NAICS codes (simplified mapping)
		naicsCode := c.mapIndustryToNAICS(industry)
		if naicsCode != "" {
			classification := IndustryClassification{
				IndustryCode:         naicsCode,
				IndustryName:         industry,
				ConfidenceScore:      confidence * 0.8, // Scale confidence to 0.75-0.85 range
				ClassificationMethod: "search_analysis",
				Description:          fmt.Sprintf("Search-based classification with confidence %.2f", confidence),
			}
			classifications = append(classifications, classification)
		}
	}

	// Add classifications from individual search results
	for _, analyzedResult := range analysisResult.Results {
		for _, indicator := range analyzedResult.IndustryIndicators {
			naicsCode := c.mapIndustryToNAICS(indicator)
			if naicsCode != "" {
				classification := IndustryClassification{
					IndustryCode:         naicsCode,
					IndustryName:         indicator,
					ConfidenceScore:      analyzedResult.ConfidenceScore,
					ClassificationMethod: "search_result_analysis",
					Description:          fmt.Sprintf("Search result analysis with quality score %.2f", analyzedResult.QualityScore),
				}
				classifications = append(classifications, classification)
			}
		}
	}

	return classifications
}

// buildSearchQuery builds a search query from business information
func (c *ClassificationService) buildSearchQuery(req *ClassificationRequest) string {
	var queryParts []string

	// Add business name
	if req.BusinessName != "" {
		queryParts = append(queryParts, req.BusinessName)
	}

	// Add business type
	if req.BusinessType != "" {
		queryParts = append(queryParts, req.BusinessType)
	}

	// Add industry
	if req.Industry != "" {
		queryParts = append(queryParts, req.Industry)
	}

	// Add description keywords
	if req.Description != "" {
		// Extract key terms from description
		words := strings.Fields(req.Description)
		if len(words) > 0 {
			// Take first few meaningful words
			keyWords := words[:min(len(words), 5)]
			queryParts = append(queryParts, strings.Join(keyWords, " "))
		}
	}

	// Add keywords
	if req.Keywords != "" {
		queryParts = append(queryParts, req.Keywords)
	}

	// Combine all parts
	query := strings.Join(queryParts, " ")
	if query == "" {
		query = req.BusinessName // Fallback to just business name
	}

	return query
}

// mapIndustryToNAICS maps industry names to NAICS codes
func (c *ClassificationService) mapIndustryToNAICS(industry string) string {
	// Simplified mapping - in production this would be more comprehensive
	industryLower := strings.ToLower(industry)

	switch {
	case strings.Contains(industryLower, "technology") || strings.Contains(industryLower, "software") || strings.Contains(industryLower, "tech"):
		return "511200" // Software Publishers
	case strings.Contains(industryLower, "healthcare") || strings.Contains(industryLower, "medical") || strings.Contains(industryLower, "health"):
		return "621100" // Offices of Physicians
	case strings.Contains(industryLower, "finance") || strings.Contains(industryLower, "financial") || strings.Contains(industryLower, "banking"):
		return "522100" // Depository Credit Intermediation
	case strings.Contains(industryLower, "retail") || strings.Contains(industryLower, "store") || strings.Contains(industryLower, "shop"):
		return "441100" // Automobile Dealers
	case strings.Contains(industryLower, "manufacturing") || strings.Contains(industryLower, "factory") || strings.Contains(industryLower, "production"):
		return "332000" // Fabricated Metal Product Manufacturing
	case strings.Contains(industryLower, "education") || strings.Contains(industryLower, "school") || strings.Contains(industryLower, "university"):
		return "611100" // Elementary and Secondary Schools
	case strings.Contains(industryLower, "real estate") || strings.Contains(industryLower, "property") || strings.Contains(industryLower, "realty"):
		return "531100" // Lessors of Real Estate
	case strings.Contains(industryLower, "legal") || strings.Contains(industryLower, "law") || strings.Contains(industryLower, "attorney"):
		return "541100" // Legal Services
	case strings.Contains(industryLower, "consulting") || strings.Contains(industryLower, "advisory") || strings.Contains(industryLower, "strategy"):
		return "541600" // Management, Scientific, and Technical Consulting Services
	case strings.Contains(industryLower, "transportation") || strings.Contains(industryLower, "logistics") || strings.Contains(industryLower, "shipping"):
		return "484000" // Truck Transportation
	default:
		return "" // No mapping found
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Crosswalk Mapping Methods

// ClassifyWithCrosswalk performs classification using crosswalk mappings
func (c *ClassificationService) ClassifyWithCrosswalk(ctx context.Context, inputCode, inputSystem string, businessInfo map[string]interface{}) (*CrosswalkClassification, error) {
	if c.crosswalkMapper == nil {
		return nil, fmt.Errorf("crosswalk mapper not initialized")
	}

	return c.crosswalkMapper.ClassifyWithCrosswalk(ctx, inputCode, inputSystem, businessInfo)
}

// AddCrosswalkMapping adds a new crosswalk mapping
func (c *ClassificationService) AddCrosswalkMapping(ctx context.Context, mapping *CrosswalkMapping) error {
	if c.crosswalkMapper == nil {
		return fmt.Errorf("crosswalk mapper not initialized")
	}

	return c.crosswalkMapper.AddMapping(ctx, mapping)
}

// GetCrosswalkMapping retrieves a crosswalk mapping
func (c *ClassificationService) GetCrosswalkMapping(ctx context.Context, mccCode, naicsCode, sicCode string) (*CrosswalkMapping, error) {
	if c.crosswalkMapper == nil {
		return nil, fmt.Errorf("crosswalk mapper not initialized")
	}

	return c.crosswalkMapper.GetMapping(ctx, mccCode, naicsCode, sicCode)
}

// UpdateCrosswalkMapping updates an existing crosswalk mapping
func (c *ClassificationService) UpdateCrosswalkMapping(ctx context.Context, key string, updates map[string]interface{}) error {
	if c.crosswalkMapper == nil {
		return fmt.Errorf("crosswalk mapper not initialized")
	}

	return c.crosswalkMapper.UpdateMapping(ctx, key, updates)
}

// ListCrosswalkMappings returns all crosswalk mappings with optional filters
func (c *ClassificationService) ListCrosswalkMappings(ctx context.Context, filters map[string]interface{}) ([]*CrosswalkMapping, error) {
	if c.crosswalkMapper == nil {
		return nil, fmt.Errorf("crosswalk mapper not initialized")
	}

	return c.crosswalkMapper.ListMappings(ctx, filters)
}

// ValidateCrosswalkMapping validates a crosswalk mapping
func (c *ClassificationService) ValidateCrosswalkMapping(ctx context.Context, mapping *CrosswalkMapping) *CrosswalkValidation {
	if c.crosswalkMapper == nil {
		return &CrosswalkValidation{
			IsValid:    false,
			Confidence: 0.0,
			Issues: []ValidationIssue{
				{
					Type:        "system_error",
					Severity:    "critical",
					Description: "Crosswalk mapper not initialized",
				},
			},
		}
	}

	return c.crosswalkMapper.ValidateMapping(ctx, mapping)
}

// CalculateCrosswalkConfidence calculates confidence score for a crosswalk mapping
func (c *ClassificationService) CalculateCrosswalkConfidence(ctx context.Context, mapping *CrosswalkMapping) *CrosswalkConfidence {
	if c.crosswalkMapper == nil {
		return &CrosswalkConfidence{
			OverallConfidence: 0.0,
			Factors:           make(map[string]float64),
			Metadata: map[string]interface{}{
				"error": "Crosswalk mapper not initialized",
			},
		}
	}

	return c.crosswalkMapper.CalculateConfidence(ctx, mapping)
}

// GetCrosswalkMapperStats returns statistics about the crosswalk mapper
func (c *ClassificationService) GetCrosswalkMapperStats() map[string]interface{} {
	if c.crosswalkMapper == nil {
		return map[string]interface{}{
			"error": "Crosswalk mapper not initialized",
		}
	}

	return c.crosswalkMapper.GetMapperStats()
}

// Enhanced classification with crosswalk integration
func (c *ClassificationService) ClassifyWithCrosswalkIntegration(ctx context.Context, req *ClassificationRequest) ([]IndustryClassification, error) {
	// Perform standard classification first
	response, err := c.ClassifyBusiness(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("standard classification failed: %w", err)
	}

	classifications := response.Classifications

	// Enhance with crosswalk mappings if available
	if c.crosswalkMapper != nil && len(classifications) > 0 {
		enhancedClassifications := make([]IndustryClassification, 0, len(classifications))

		for _, classification := range classifications {
			// Try to find crosswalk mapping for the classification
			if classification.IndustryCode != "" {
				businessInfo := map[string]interface{}{
					"business_name": req.BusinessName,
					"business_type": req.BusinessType,
					"industry":      req.Industry,
					"description":   req.Description,
				}

				crosswalkResult, err := c.crosswalkMapper.ClassifyWithCrosswalk(ctx, classification.IndustryCode, "naics", businessInfo)
				if err == nil && crosswalkResult != nil {
					// Create enhanced classification with crosswalk data
					enhancedClassification := IndustryClassification{
						IndustryCode:         classification.IndustryCode,
						IndustryName:         classification.IndustryName,
						ConfidenceScore:      classification.ConfidenceScore * crosswalkResult.Confidence,
						ClassificationMethod: "crosswalk_enhanced",
						Description:          fmt.Sprintf("Enhanced with crosswalk mapping (confidence: %.2f)", crosswalkResult.Confidence),
					}
					enhancedClassifications = append(enhancedClassifications, enhancedClassification)
				} else {
					// Keep original classification if crosswalk mapping fails
					enhancedClassifications = append(enhancedClassifications, classification)
				}
			} else {
				enhancedClassifications = append(enhancedClassifications, classification)
			}
		}

			return enhancedClassifications, nil
}

// Geographic Region Methods

// DetectGeographicRegion detects the geographic region for a business
func (c *ClassificationService) DetectGeographicRegion(ctx context.Context, businessInfo map[string]interface{}) (*RegionDetectionResult, error) {
	if c.geographicManager == nil {
		return nil, fmt.Errorf("geographic manager not initialized")
	}

	return c.geographicManager.DetectRegion(ctx, businessInfo)
}

// ClassifyWithGeographicRegion performs classification with geographic region awareness
func (c *ClassificationService) ClassifyWithGeographicRegion(ctx context.Context, classifications []IndustryClassification, region *GeographicRegion) (*RegionClassificationResult, error) {
	if c.geographicManager == nil {
		return &RegionClassificationResult{
			Classifications:       classifications,
			ConfidenceAdjustments: make(map[string]float64),
			AppliedRules:          make([]string, 0),
			ProcessingTime:        0,
			Metadata:              make(map[string]interface{}),
		}, nil
	}

	return c.geographicManager.ClassifyWithRegion(ctx, classifications, region)
}

// AddGeographicRegion adds a new geographic region
func (c *ClassificationService) AddGeographicRegion(ctx context.Context, region *GeographicRegion) error {
	if c.geographicManager == nil {
		return fmt.Errorf("geographic manager not initialized")
	}

	return c.geographicManager.AddRegion(ctx, region)
}

// GetGeographicRegion retrieves a geographic region by ID
func (c *ClassificationService) GetGeographicRegion(ctx context.Context, regionID string) (*GeographicRegion, error) {
	if c.geographicManager == nil {
		return nil, fmt.Errorf("geographic manager not initialized")
	}

	return c.geographicManager.GetRegion(ctx, regionID)
}

// UpdateGeographicRegion updates an existing geographic region
func (c *ClassificationService) UpdateGeographicRegion(ctx context.Context, regionID string, updates map[string]interface{}) error {
	if c.geographicManager == nil {
		return fmt.Errorf("geographic manager not initialized")
	}

	return c.geographicManager.UpdateRegion(ctx, regionID, updates)
}

// ListGeographicRegions returns all geographic regions with optional filters
func (c *ClassificationService) ListGeographicRegions(ctx context.Context, filters map[string]interface{}) ([]*GeographicRegion, error) {
	if c.geographicManager == nil {
		return nil, fmt.Errorf("geographic manager not initialized")
	}

	return c.geographicManager.ListRegions(ctx, filters)
}

// GetGeographicManagerStats returns statistics about the geographic manager
func (c *ClassificationService) GetGeographicManagerStats() map[string]interface{} {
	if c.geographicManager == nil {
		return map[string]interface{}{
			"error": "Geographic manager not initialized",
		}
	}

	return c.geographicManager.GetManagerStats()
}

// Enhanced classification with geographic region integration
func (c *ClassificationService) ClassifyWithGeographicIntegration(ctx context.Context, req *ClassificationRequest) ([]IndustryClassification, error) {
	// Perform standard classification first
	response, err := c.ClassifyBusiness(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("standard classification failed: %w", err)
	}

	classifications := response.Classifications

	// Detect geographic region if geographic manager is available
	if c.geographicManager != nil {
		businessInfo := map[string]interface{}{
			"business_name": req.BusinessName,
			"business_type": req.BusinessType,
			"industry":      req.Industry,
			"description":   req.Description,
			// Add any location information from the request
		}

		regionResult, err := c.geographicManager.DetectRegion(ctx, businessInfo)
		if err == nil && regionResult != nil && regionResult.DetectedRegion != nil {
			// Apply geographic region adjustments
			regionClassificationResult, err := c.geographicManager.ClassifyWithRegion(ctx, classifications, regionResult.DetectedRegion)
			if err == nil && regionClassificationResult != nil {
				// Log geographic integration
				if c.logger != nil {
					c.logger.WithComponent("classification").LogBusinessEvent(ctx, "geographic_integration_applied", "", map[string]interface{}{
						"business_name":     req.BusinessName,
						"detected_region":   regionResult.DetectedRegion.Name,
						"region_confidence": regionResult.Confidence,
						"rules_applied":     len(regionClassificationResult.AppliedRules),
					})
				}

				return regionClassificationResult.Classifications, nil
			}
		}
	}

	return classifications, nil
}

// Feedback Collection Methods

// SubmitFeedback submits user feedback for classification results
func (c *ClassificationService) SubmitFeedback(ctx context.Context, feedback *Feedback) error {
	if c.feedbackCollector == nil {
		return fmt.Errorf("feedback collector not initialized")
	}

	return c.feedbackCollector.SubmitFeedback(ctx, feedback)
}

// GetFeedback retrieves feedback by ID
func (c *ClassificationService) GetFeedback(ctx context.Context, feedbackID string) (*Feedback, error) {
	if c.feedbackCollector == nil {
		return nil, fmt.Errorf("feedback collector not initialized")
	}

	return c.feedbackCollector.GetFeedback(ctx, feedbackID)
}

// ListFeedback returns feedback with optional filtering
func (c *ClassificationService) ListFeedback(ctx context.Context, filters map[string]interface{}) ([]*Feedback, error) {
	if c.feedbackCollector == nil {
		return nil, fmt.Errorf("feedback collector not initialized")
	}

	return c.feedbackCollector.ListFeedback(ctx, filters)
}

// UpdateFeedback updates feedback status
func (c *ClassificationService) UpdateFeedback(ctx context.Context, feedbackID string, updates map[string]interface{}) error {
	if c.feedbackCollector == nil {
		return fmt.Errorf("feedback collector not initialized")
	}

	return c.feedbackCollector.UpdateFeedback(ctx, feedbackID, updates)
}

// GetAccuracyMetrics returns accuracy metrics based on feedback
func (c *ClassificationService) GetAccuracyMetrics(ctx context.Context, filters map[string]interface{}) (*FeedbackAccuracyMetrics, error) {
	if c.feedbackCollector == nil {
		return nil, fmt.Errorf("feedback collector not initialized")
	}

	return c.feedbackCollector.GetAccuracyMetrics(ctx, filters)
}

// GetModelUpdates returns model updates based on feedback
func (c *ClassificationService) GetModelUpdates(ctx context.Context) (map[string]interface{}, error) {
	if c.feedbackCollector == nil {
		return nil, fmt.Errorf("feedback collector not initialized")
	}

	return c.feedbackCollector.GetModelUpdates(ctx)
}

// GetFeedbackStats returns statistics about the feedback collector
func (c *ClassificationService) GetFeedbackStats() map[string]interface{} {
	if c.feedbackCollector == nil {
		return map[string]interface{}{
			"error": "Feedback collector not initialized",
		}
	}

	return c.feedbackCollector.GetCollectorStats()
}

// Enhanced classification with feedback integration
func (c *ClassificationService) ClassifyWithFeedbackIntegration(ctx context.Context, req *ClassificationRequest) ([]IndustryClassification, error) {
	// Perform standard classification first
	response, err := c.ClassifyBusiness(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("standard classification failed: %w", err)
	}

	classifications := response.Classifications

	// Apply feedback-based adjustments if feedback collector is available
	if c.feedbackCollector != nil {
		// Get accuracy metrics for this business type/industry
		businessInfo := map[string]interface{}{
			"business_name": req.BusinessName,
			"business_type": req.BusinessType,
			"industry":      req.Industry,
		}

		accuracyMetrics, err := c.feedbackCollector.GetAccuracyMetrics(ctx, businessInfo)
		if err == nil && accuracyMetrics != nil {
			// Apply feedback-based confidence adjustments
			for i := range classifications {
				// Adjust confidence based on feedback accuracy
				if accuracyMetrics.AccuracyScore > 0.8 {
					// High accuracy feedback - boost confidence slightly
					classifications[i].ConfidenceScore *= 1.05
				} else if accuracyMetrics.AccuracyScore < 0.6 {
					// Low accuracy feedback - reduce confidence slightly
					classifications[i].ConfidenceScore *= 0.95
				}

				// Ensure confidence stays within bounds
				if classifications[i].ConfidenceScore > 1.0 {
					classifications[i].ConfidenceScore = 1.0
				}
			}

			// Log feedback integration
			if c.logger != nil {
				c.logger.WithComponent("classification").LogBusinessEvent(ctx, "feedback_integration_applied", "", map[string]interface{}{
					"business_name":     req.BusinessName,
					"accuracy_score":    accuracyMetrics.AccuracyScore,
					"total_feedback":    accuracyMetrics.TotalFeedback,
					"confidence_adjustment": "applied",
				})
			}
		}
	}

	return classifications, nil
}
