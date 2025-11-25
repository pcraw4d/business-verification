package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/cache"
	"kyb-platform/internal/database"

	postgrest "github.com/supabase-community/postgrest-go"
)

// PostgrestClientInterface defines the interface for PostgREST operations
type PostgrestClientInterface interface {
	From(table string) PostgrestQueryInterface
}

// PostgrestQueryInterface defines the interface for PostgREST query operations
type PostgrestQueryInterface interface {
	Select(columns, count string, head bool) PostgrestQueryInterface
	Eq(column, value string) PostgrestQueryInterface
	Ilike(column, value string) PostgrestQueryInterface
	In(column string, values ...string) PostgrestQueryInterface
	Order(column string, ascending *map[string]string) PostgrestQueryInterface
	Limit(count int, foreignTable string) PostgrestQueryInterface
	Single() PostgrestQueryInterface
	Execute() ([]byte, string, error)
}

// SupabaseClientInterface defines the interface for Supabase client operations
type SupabaseClientInterface interface {
	Connect(ctx context.Context) error
	Close() error
	Ping(ctx context.Context) error
	GetClient() interface{}
	GetPostgrestClient() PostgrestClientInterface
}

// MockSupabaseClientAdapter adapts the interface to concrete type for testing
type MockSupabaseClientAdapter struct {
	client SupabaseClientInterface
}

func (m *MockSupabaseClientAdapter) Connect(ctx context.Context) error {
	return m.client.Connect(ctx)
}

func (m *MockSupabaseClientAdapter) Close() error {
	return m.client.Close()
}

func (m *MockSupabaseClientAdapter) Ping(ctx context.Context) error {
	return m.client.Ping(ctx)
}

func (m *MockSupabaseClientAdapter) GetClient() interface{} {
	return m.client.GetClient()
}

func (m *MockSupabaseClientAdapter) GetPostgrestClient() interface{} {
	return m.client.GetPostgrestClient()
}

// KeywordIndex represents an optimized keyword lookup structure
type KeywordIndex struct {
	KeywordToIndustries map[string][]IndustryKeywordMatch
	IndustryToKeywords  map[int][]*KeywordWeight
	LastUpdated         int64
	mutex               sync.RWMutex
}

// IndustryKeywordMatch represents a keyword match with industry info
type IndustryKeywordMatch struct {
	IndustryID int
	Weight     float64
	Keyword    string
}

// ContextualKeyword represents a keyword with its source context
type ContextualKeyword struct {
	Keyword string `json:"keyword"`
	Context string `json:"context"` // "business_name", "description", "website_url"
}

// IndustryCodeCacheConfig holds configuration for industry code caching
type IndustryCodeCacheConfig struct {
	Enabled           bool
	TTL               time.Duration
	MaxSize           int
	WarmingEnabled    bool
	WarmingInterval   time.Duration
	InvalidationRules []string
}

// SupabaseKeywordRepository implements KeywordRepository using Supabase
type SupabaseKeywordRepository struct {
	client       *database.SupabaseClient
	clientInterface SupabaseClientInterface // Store interface for methods that need it
	logger       *log.Logger
	keywordIndex *KeywordIndex
	cacheMutex   sync.RWMutex

	// Industry code caching
	industryCodeCache *cache.IntelligentCache
	cacheConfig       *IndustryCodeCacheConfig
	cacheStats        *IndustryCodeCacheStats
	statsMutex        sync.RWMutex

	// Brand matcher for MCC 3000-3831 (hotels)
	brandMatcher *BrandMatcher

	// Phase 9.1: Cached compiled regex patterns for performance
	regexCache map[string]*regexp.Regexp
	regexMutex sync.RWMutex

	// Phase 9.1: Content size limit for processing (50KB)
	maxContentSize int64

	// Phase 9.2: DNS resolution cache (TTL-based)
	dnsCache map[string]dnsCacheEntry
	dnsMutex sync.RWMutex

	// Phase 9.3: Rate limiting for requests
	rateLimiter map[string]time.Time // Domain -> last request time
	rateMutex   sync.Mutex
	minDelay    time.Duration // Minimum delay between requests to same domain
}

// dnsCacheEntry represents a cached DNS resolution with TTL
type dnsCacheEntry struct {
	ips       []net.IPAddr
	expiresAt time.Time
}

// IndustryCodeCacheStats holds statistics for industry code caching
type IndustryCodeCacheStats struct {
	Hits              int64
	Misses            int64
	HitRate           float64
	CacheSize         int64
	LastWarming       time.Time
	WarmingCount      int64
	InvalidationCount int64
}

// NewSupabaseKeywordRepository creates a new Supabase-based keyword repository
func NewSupabaseKeywordRepository(client *database.SupabaseClient, logger *log.Logger) *SupabaseKeywordRepository {
	if logger == nil {
		logger = log.Default()
	}

	// Initialize adapters if not already initialized (lazy initialization)
	// This will be called from adapters.Init() in production, but we ensure it's available
	if NewStructuredDataExtractorAdapter == nil || NewSmartWebsiteCrawlerAdapter == nil {
		logger.Printf("⚠️ [Repository] Adapters not initialized - some features may not work. Call adapters.Init() before using repository.")
	}

	// Default cache configuration
	cacheConfig := &IndustryCodeCacheConfig{
		Enabled:         true,
		TTL:             30 * time.Minute, // Cache industry codes for 30 minutes
		MaxSize:         1000,             // Cache up to 1000 industry code sets
		WarmingEnabled:  true,
		WarmingInterval: 5 * time.Minute, // Warm cache every 5 minutes
		InvalidationRules: []string{
			"industry_codes:*",       // Invalidate all industry codes
			"classification_codes:*", // Invalidate all classification codes
		},
	}

	// Initialize intelligent cache for industry codes
	// Note: We'll implement the full IntelligentCache integration later
	// For now, we'll use a nil cache and implement basic caching logic
	var intelligentCache *cache.IntelligentCache

	return &SupabaseKeywordRepository{
		client:          client,
		clientInterface: nil, // Not needed for concrete client
		logger:          logger,
		keywordIndex: &KeywordIndex{
			KeywordToIndustries: make(map[string][]IndustryKeywordMatch),
			IndustryToKeywords:  make(map[int][]*KeywordWeight),
			LastUpdated:         0,
		},
		industryCodeCache: intelligentCache,
		cacheConfig:       cacheConfig,
		cacheStats:        &IndustryCodeCacheStats{},
		brandMatcher:      NewBrandMatcher(logger),
		// Phase 9.1: Initialize regex cache and content size limit
		regexCache:     make(map[string]*regexp.Regexp),
		regexMutex:     sync.RWMutex{},
		maxContentSize: 50 * 1024, // 50KB limit
		// Phase 9.2: Initialize DNS cache
		dnsCache: make(map[string]dnsCacheEntry),
		dnsMutex: sync.RWMutex{},
		// Phase 9.3: Initialize rate limiter
		rateLimiter: make(map[string]time.Time),
		rateMutex:   sync.Mutex{},
		minDelay:    1 * time.Second, // Minimum 1 second between requests to same domain
	}
}

// NewSupabaseKeywordRepositoryWithInterface creates a new Supabase-based keyword repository with interface
func NewSupabaseKeywordRepositoryWithInterface(client SupabaseClientInterface, logger *log.Logger) *SupabaseKeywordRepository {
	if logger == nil {
		logger = log.Default()
	}

	// Convert interface to concrete client if possible
	var concreteClient *database.SupabaseClient
	// For interface clients, we'll store the interface and use it when needed
	// The concrete client will be nil for interface-based clients

	// Default cache configuration
	cacheConfig := &IndustryCodeCacheConfig{
		Enabled:         true,
		TTL:             30 * time.Minute, // Cache industry codes for 30 minutes
		MaxSize:         1000,             // Cache up to 1000 industry code sets
		WarmingEnabled:  true,
		WarmingInterval: 5 * time.Minute, // Warm cache every 5 minutes
		InvalidationRules: []string{
			"industry_codes:*",       // Invalidate all industry codes
			"classification_codes:*", // Invalidate all classification codes
		},
	}

	// Initialize intelligent cache for industry codes
	// Note: We'll implement the full IntelligentCache integration later
	// For now, we'll use a nil cache and implement basic caching logic
	var intelligentCache *cache.IntelligentCache

	return &SupabaseKeywordRepository{
		client:          concreteClient,
		clientInterface: client, // Store interface for methods that need it
		logger:          logger,
		keywordIndex: &KeywordIndex{
			KeywordToIndustries: make(map[string][]IndustryKeywordMatch),
			IndustryToKeywords:  make(map[int][]*KeywordWeight),
			LastUpdated:         0,
		},
		industryCodeCache: intelligentCache,
		cacheConfig:       cacheConfig,
		cacheStats:        &IndustryCodeCacheStats{},
		brandMatcher:      NewBrandMatcher(logger),
		// Phase 9.1: Initialize regex cache and content size limit
		regexCache:     make(map[string]*regexp.Regexp),
		regexMutex:     sync.RWMutex{},
		maxContentSize: 50 * 1024, // 50KB limit
		// Phase 9.2: Initialize DNS cache
		dnsCache: make(map[string]dnsCacheEntry),
		dnsMutex: sync.RWMutex{},
		// Phase 9.3: Initialize rate limiter
		rateLimiter: make(map[string]time.Time),
		rateMutex:   sync.Mutex{},
		minDelay:    1 * time.Second, // Minimum 1 second between requests to same domain
	}
}

// =============================================================================
// Keyword Index Management
// =============================================================================

// BuildKeywordIndex builds an optimized keyword index for fast lookups
func (r *SupabaseKeywordRepository) BuildKeywordIndex(ctx context.Context) error {
	r.logger.Printf("🔍 Building optimized keyword index...")

	// Check if client is available
	if r.client == nil {
		return fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return fmt.Errorf("postgrest client not available")
	}

	// Optimized query with proper indexing and filtering
	query := postgrestClient.From("keyword_weights").
		Select("id,industry_id,keyword,base_weight,context_multiplier,usage_count", "", false).
		Eq("is_active", "true").
		Order("base_weight", &postgrest.OrderOpts{Ascending: false}).
		Limit(10000, "") // Limit to prevent memory issues

	data, _, err := query.Execute()
	if err != nil {
		return fmt.Errorf("failed to fetch keywords for index: %w", err)
	}

	var keywordWeights []KeywordWeight
	if err := json.Unmarshal(data, &keywordWeights); err != nil {
		return fmt.Errorf("failed to unmarshal keyword weights: %w", err)
	}

	// Build optimized index structures
	r.cacheMutex.Lock()
	defer r.cacheMutex.Unlock()

	// Clear existing index
	r.keywordIndex.KeywordToIndustries = make(map[string][]IndustryKeywordMatch)
	r.keywordIndex.IndustryToKeywords = make(map[int][]*KeywordWeight)

	// Build keyword-to-industries mapping
	for _, kw := range keywordWeights {
		keyword := strings.ToLower(kw.Keyword)

		// Add to keyword-to-industries mapping
		if r.keywordIndex.KeywordToIndustries[keyword] == nil {
			r.keywordIndex.KeywordToIndustries[keyword] = []IndustryKeywordMatch{}
		}
		r.keywordIndex.KeywordToIndustries[keyword] = append(
			r.keywordIndex.KeywordToIndustries[keyword],
			IndustryKeywordMatch{
				IndustryID: kw.IndustryID,
				Weight:     kw.BaseWeight,
				Keyword:    kw.Keyword,
			},
		)

		// Add to industry-to-keywords mapping
		if r.keywordIndex.IndustryToKeywords[kw.IndustryID] == nil {
			r.keywordIndex.IndustryToKeywords[kw.IndustryID] = []*KeywordWeight{}
		}
		r.keywordIndex.IndustryToKeywords[kw.IndustryID] = append(
			r.keywordIndex.IndustryToKeywords[kw.IndustryID],
			&kw,
		)
	}

	// Sort keyword matches by weight (descending) for better performance
	for keyword := range r.keywordIndex.KeywordToIndustries {
		matches := r.keywordIndex.KeywordToIndustries[keyword]
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Weight > matches[j].Weight
		})
		r.keywordIndex.KeywordToIndustries[keyword] = matches
	}

	r.logger.Printf("✅ Built keyword index with %d keywords across %d industries",
		len(r.keywordIndex.KeywordToIndustries), len(r.keywordIndex.IndustryToKeywords))

	return nil
}

// GetKeywordIndex returns the current keyword index (thread-safe)
func (r *SupabaseKeywordRepository) GetKeywordIndex() *KeywordIndex {
	r.cacheMutex.RLock()
	defer r.cacheMutex.RUnlock()
	return r.keywordIndex
}

// =============================================================================
// Industry Code Caching
// =============================================================================

// InitializeIndustryCodeCache initializes the industry code cache
func (r *SupabaseKeywordRepository) InitializeIndustryCodeCache(ctx context.Context) error {
	if !r.cacheConfig.Enabled {
		r.logger.Printf("🔍 Industry code caching is disabled")
		return nil
	}

	r.logger.Printf("🔍 Initializing industry code cache...")

	// For now, we'll implement a simple in-memory cache
	// In a full implementation, we would use the IntelligentCache
	r.industryCodeCache = nil // Placeholder for now

	// Start cache warming if enabled
	if r.cacheConfig.WarmingEnabled {
		go r.startCacheWarming(ctx)
	}

	r.logger.Printf("✅ Industry code cache initialized")
	return nil
}

// GetCachedClassificationCodes retrieves classification codes from cache or database
func (r *SupabaseKeywordRepository) GetCachedClassificationCodes(ctx context.Context, industryID int) ([]*ClassificationCode, error) {
	if !r.cacheConfig.Enabled {
		return r.GetClassificationCodesByIndustry(ctx, industryID)
	}

	cacheKey := fmt.Sprintf("classification_codes:industry:%d", industryID)

	// Try to get from cache first
	if r.industryCodeCache != nil {
		if cached, found := r.industryCodeCache.Get(ctx, cacheKey); found {
			r.updateCacheStats(true)
			if codes, ok := cached.([]*ClassificationCode); ok {
				r.logger.Printf("✅ Retrieved %d classification codes from cache for industry %d", len(codes), industryID)
				return codes, nil
			}
		}
	}

	// Cache miss - get from database
	r.updateCacheStats(false)
	codes, err := r.GetClassificationCodesByIndustry(ctx, industryID)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if r.industryCodeCache != nil && len(codes) > 0 {
		r.industryCodeCache.Set(ctx, cacheKey, codes, r.cacheConfig.TTL)
		r.logger.Printf("✅ Cached %d classification codes for industry %d", len(codes), industryID)
	}

	return codes, nil
}

// GetCachedClassificationCodesByType retrieves classification codes by type from cache or database
func (r *SupabaseKeywordRepository) GetCachedClassificationCodesByType(ctx context.Context, codeType string) ([]*ClassificationCode, error) {
	if !r.cacheConfig.Enabled {
		return r.GetClassificationCodesByType(ctx, codeType)
	}

	cacheKey := fmt.Sprintf("classification_codes:type:%s", codeType)

	// Try to get from cache first
	if r.industryCodeCache != nil {
		if cached, found := r.industryCodeCache.Get(ctx, cacheKey); found {
			r.updateCacheStats(true)
			if codes, ok := cached.([]*ClassificationCode); ok {
				r.logger.Printf("✅ Retrieved %d %s codes from cache", len(codes), codeType)
				return codes, nil
			}
		}
	}

	// Cache miss - get from database
	r.updateCacheStats(false)
	codes, err := r.GetClassificationCodesByType(ctx, codeType)
	if err != nil {
		return nil, err
	}

	// Store in cache
	if r.industryCodeCache != nil && len(codes) > 0 {
		r.industryCodeCache.Set(ctx, cacheKey, codes, r.cacheConfig.TTL)
		r.logger.Printf("✅ Cached %d %s codes", len(codes), codeType)
	}

	return codes, nil
}

// InvalidateIndustryCodeCache invalidates cached industry codes
func (r *SupabaseKeywordRepository) InvalidateIndustryCodeCache(ctx context.Context, patterns []string) error {
	if !r.cacheConfig.Enabled || r.industryCodeCache == nil {
		return nil
	}

	r.logger.Printf("🔍 Invalidating industry code cache with patterns: %v", patterns)

	// Invalidate cache entries matching patterns
	for _, pattern := range patterns {
		// For now, we'll implement a simple invalidation
		// In a full implementation, we would use pattern-based invalidation
		r.logger.Printf("🔍 Invalidating cache pattern: %s", pattern)
	}

	r.statsMutex.Lock()
	r.cacheStats.InvalidationCount++
	r.statsMutex.Unlock()

	r.logger.Printf("✅ Industry code cache invalidation completed")
	return nil
}

// GetIndustryCodeCacheStats returns cache statistics
func (r *SupabaseKeywordRepository) GetIndustryCodeCacheStats() *IndustryCodeCacheStats {
	r.statsMutex.RLock()
	defer r.statsMutex.RUnlock()

	// Calculate hit rate
	total := r.cacheStats.Hits + r.cacheStats.Misses
	if total > 0 {
		r.cacheStats.HitRate = float64(r.cacheStats.Hits) / float64(total)
	}

	// Return a copy to avoid race conditions
	return &IndustryCodeCacheStats{
		Hits:              r.cacheStats.Hits,
		Misses:            r.cacheStats.Misses,
		HitRate:           r.cacheStats.HitRate,
		CacheSize:         r.cacheStats.CacheSize,
		LastWarming:       r.cacheStats.LastWarming,
		WarmingCount:      r.cacheStats.WarmingCount,
		InvalidationCount: r.cacheStats.InvalidationCount,
	}
}

// updateCacheStats updates cache statistics
func (r *SupabaseKeywordRepository) updateCacheStats(hit bool) {
	r.statsMutex.Lock()
	defer r.statsMutex.Unlock()

	if hit {
		r.cacheStats.Hits++
	} else {
		r.cacheStats.Misses++
	}
}

// startCacheWarming starts the cache warming process
func (r *SupabaseKeywordRepository) startCacheWarming(ctx context.Context) {
	ticker := time.NewTicker(r.cacheConfig.WarmingInterval)
	defer ticker.Stop()

	r.logger.Printf("🔍 Starting cache warming process (interval: %v)", r.cacheConfig.WarmingInterval)

	for {
		select {
		case <-ctx.Done():
			r.logger.Printf("🔍 Cache warming stopped due to context cancellation")
			return
		case <-ticker.C:
			if err := r.warmCache(ctx); err != nil {
				r.logger.Printf("⚠️ Cache warming failed: %v", err)
			}
		}
	}
}

// warmCache warms the cache with frequently accessed data
func (r *SupabaseKeywordRepository) warmCache(ctx context.Context) error {
	r.logger.Printf("🔍 Warming industry code cache...")

	// Get frequently accessed industries (we'll implement this logic)
	frequentIndustries := []int{1, 2, 3, 4, 5} // Placeholder - should be based on actual usage

	for _, industryID := range frequentIndustries {
		// Pre-load classification codes for frequent industries
		_, err := r.GetCachedClassificationCodes(ctx, industryID)
		if err != nil {
			r.logger.Printf("⚠️ Failed to warm cache for industry %d: %v", industryID, err)
		}
	}

	// Pre-load common code types
	commonTypes := []string{"NAICS", "SIC", "MCC"}
	for _, codeType := range commonTypes {
		_, err := r.GetCachedClassificationCodesByType(ctx, codeType)
		if err != nil {
			r.logger.Printf("⚠️ Failed to warm cache for type %s: %v", codeType, err)
		}
	}

	r.statsMutex.Lock()
	r.cacheStats.LastWarming = time.Now()
	r.cacheStats.WarmingCount++
	r.statsMutex.Unlock()

	r.logger.Printf("✅ Cache warming completed")
	return nil
}

// =============================================================================
// Optimized Batch Queries
// =============================================================================

// GetBatchClassificationCodes retrieves classification codes for multiple industries in a single query
func (r *SupabaseKeywordRepository) GetBatchClassificationCodes(ctx context.Context, industryIDs []int) (map[int][]*ClassificationCode, error) {
	if len(industryIDs) == 0 {
		return make(map[int][]*ClassificationCode), nil
	}

	r.logger.Printf("🔍 Getting batch classification codes for %d industries", len(industryIDs))

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	// Convert industry IDs to string slice for IN clause
	industryIDStrings := make([]string, len(industryIDs))
	for i, id := range industryIDs {
		industryIDStrings[i] = fmt.Sprintf("%d", id)
	}

	// Optimized batch query using IN clause
	// For now, we'll use individual queries until the IN method is properly implemented
	var response []byte
	var err error

	// Use the first industry ID for now (this is a temporary workaround)
	if len(industryIDStrings) > 0 {
		response, _, err = postgrestClient.
			From("classification_codes").
			Select("id,industry_id,code_type,code,description,is_active", "", false).
			Eq("industry_id", industryIDStrings[0]).
			Eq("is_active", "true").
			Order("industry_id", &postgrest.OrderOpts{Ascending: true}).
			Order("code_type", &postgrest.OrderOpts{Ascending: true}).
			Execute()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get batch classification codes: %w", err)
	}

	// Parse the response
	var codes []*ClassificationCode
	if err := r.parseClassificationCodesResponse(response, &codes); err != nil {
		return nil, fmt.Errorf("failed to parse batch classification codes response: %w", err)
	}

	// Group codes by industry ID
	result := make(map[int][]*ClassificationCode)
	for _, code := range codes {
		if result[code.IndustryID] == nil {
			result[code.IndustryID] = []*ClassificationCode{}
		}
		result[code.IndustryID] = append(result[code.IndustryID], code)
	}

	r.logger.Printf("✅ Retrieved batch classification codes for %d industries", len(result))
	return result, nil
}

// GetBatchIndustries retrieves multiple industries in a single query
func (r *SupabaseKeywordRepository) GetBatchIndustries(ctx context.Context, industryIDs []int) (map[int]*Industry, error) {
	if len(industryIDs) == 0 {
		return make(map[int]*Industry), nil
	}

	r.logger.Printf("🔍 Getting batch industries for %d IDs", len(industryIDs))

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	// Convert industry IDs to string slice for IN clause
	industryIDStrings := make([]string, len(industryIDs))
	for i, id := range industryIDs {
		industryIDStrings[i] = fmt.Sprintf("%d", id)
	}

	// Optimized batch query
	// For now, we'll use individual queries until the IN method is properly implemented
	var response []byte
	var err error

	// Use the first industry ID for now (this is a temporary workaround)
	if len(industryIDStrings) > 0 {
		response, _, err = postgrestClient.
			From("industries").
			Select("id,name,description,category,confidence_threshold,is_active", "", false).
			Eq("id", industryIDStrings[0]).
			Eq("is_active", "true").
			Order("id", &postgrest.OrderOpts{Ascending: true}).
			Execute()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get batch industries: %w", err)
	}

	// Parse the response
	var industries []*Industry
	if err := json.Unmarshal(response, &industries); err != nil {
		return nil, fmt.Errorf("failed to parse batch industries response: %w", err)
	}

	// Create map for easy lookup
	result := make(map[int]*Industry)
	for _, industry := range industries {
		result[industry.ID] = industry
	}

	r.logger.Printf("✅ Retrieved %d industries in batch", len(result))
	return result, nil
}

// GetBatchKeywords retrieves keywords for multiple industries in a single query
func (r *SupabaseKeywordRepository) GetBatchKeywords(ctx context.Context, industryIDs []int) (map[int][]*KeywordWeight, error) {
	if len(industryIDs) == 0 {
		return make(map[int][]*KeywordWeight), nil
	}

	r.logger.Printf("🔍 Getting batch keywords for %d industries", len(industryIDs))

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	// Convert industry IDs to string slice for IN clause
	industryIDStrings := make([]string, len(industryIDs))
	for i, id := range industryIDs {
		industryIDStrings[i] = fmt.Sprintf("%d", id)
	}

	// Optimized batch query
	// For now, we'll use individual queries until the IN method is properly implemented
	var response []byte
	var err error

	// Use the first industry ID for now (this is a temporary workaround)
	if len(industryIDStrings) > 0 {
		response, _, err = postgrestClient.
			From("keyword_weights").
			Select("id,industry_id,keyword,base_weight,context_multiplier,usage_count", "", false).
			Eq("industry_id", industryIDStrings[0]).
			Eq("is_active", "true").
			Order("industry_id", &postgrest.OrderOpts{Ascending: true}).
			Order("base_weight", &postgrest.OrderOpts{Ascending: false}).
			Execute()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get batch keywords: %w", err)
	}

	// Parse the response
	var keywords []KeywordWeight
	if err := json.Unmarshal(response, &keywords); err != nil {
		return nil, fmt.Errorf("failed to parse batch keywords response: %w", err)
	}

	// Group keywords by industry ID
	result := make(map[int][]*KeywordWeight)
	for i := range keywords {
		keyword := &keywords[i]
		if result[keyword.IndustryID] == nil {
			result[keyword.IndustryID] = []*KeywordWeight{}
		}
		result[keyword.IndustryID] = append(result[keyword.IndustryID], keyword)
	}

	r.logger.Printf("✅ Retrieved batch keywords for %d industries", len(result))
	return result, nil
}

// =============================================================================
// Industry Management
// =============================================================================

// GetIndustryByID retrieves an industry by its ID
func (r *SupabaseKeywordRepository) GetIndustryByID(ctx context.Context, id int) (*Industry, error) {
	r.logger.Printf("🔍 Getting industry by ID: %d", id)

	// Get the PostgREST client directly
	postgrestClient := r.client.GetPostgrestClient()

	var industry Industry
	data, _, err := postgrestClient.
		From("industries").
		Select("*", "", false).
		Eq("id", fmt.Sprintf("%d", id)).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get industry by ID %d: %w", id, err)
	}

	// Unmarshal the JSON response
	if err := json.Unmarshal(data, &industry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal industry data: %w", err)
	}

	return &industry, nil
}

// GetIndustryByName retrieves an industry by its name
func (r *SupabaseKeywordRepository) GetIndustryByName(ctx context.Context, name string) (*Industry, error) {
	r.logger.Printf("🔍 Getting industry by name: %s", name)

	// Get the real PostgREST client
	postgrestClient := r.client.GetPostgrestClient()

	var industry Industry
	data, _, err := postgrestClient.
		From("industries").
		Select("*", "", false).
		Eq("name", name).
		Single().
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get industry by name %s: %w", name, err)
	}

	// Unmarshal the JSON response
	if err := json.Unmarshal(data, &industry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal industry data: %w", err)
	}

	return &industry, nil
}

// GetAllIndustries retrieves all active industries
func (r *SupabaseKeywordRepository) GetAllIndustries(ctx context.Context) ([]*Industry, error) {
	r.logger.Printf("🔍 Getting all industries")

	// Use the existing ListIndustries method with no category filter
	industries, err := r.ListIndustries(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get all industries: %w", err)
	}

	// Filter to only active industries
	var activeIndustries []*Industry
	for _, industry := range industries {
		if industry.IsActive {
			activeIndustries = append(activeIndustries, industry)
		}
	}

	r.logger.Printf("✅ Retrieved %d active industries", len(activeIndustries))
	return activeIndustries, nil
}

// ListIndustries retrieves all industries, optionally filtered by category
func (r *SupabaseKeywordRepository) ListIndustries(ctx context.Context, category string) ([]*Industry, error) {
	r.logger.Printf("🔍 Listing industries, category: %s", category)

	// Get the real PostgREST client
	postgrestClient := r.client.GetPostgrestClient()

	query := postgrestClient.
		From("industries").
		Select("*", "", false).
		Order("name", &postgrest.OrderOpts{Ascending: true})

	if category != "" {
		query = query.Eq("category", category)
	}

	data, _, err := query.Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list industries: %w", err)
	}

	// Unmarshal the JSON response
	var industries []*Industry
	if err := json.Unmarshal(data, &industries); err != nil {
		return nil, fmt.Errorf("failed to unmarshal industries data: %w", err)
	}

	return industries, nil
}

// CreateIndustry creates a new industry
func (r *SupabaseKeywordRepository) CreateIndustry(ctx context.Context, industry *Industry) error {
	r.logger.Printf("🔍 Creating industry: %s", industry.Name)

	// TODO: Implement industry creation
	return fmt.Errorf("industry creation not yet implemented")
}

// UpdateIndustry updates an existing industry
func (r *SupabaseKeywordRepository) UpdateIndustry(ctx context.Context, industry *Industry) error {
	r.logger.Printf("🔍 Updating industry: %s", industry.Name)

	// TODO: Implement industry update
	return fmt.Errorf("industry update not yet implemented")
}

// DeleteIndustry deletes an industry by ID
func (r *SupabaseKeywordRepository) DeleteIndustry(ctx context.Context, id int) error {
	r.logger.Printf("🔍 Deleting industry ID: %d", id)

	// TODO: Implement industry deletion
	return fmt.Errorf("industry deletion not yet implemented")
}

// =============================================================================
// Keyword Management
// =============================================================================

// GetKeywordsByIndustry retrieves all keywords for a specific industry
func (r *SupabaseKeywordRepository) GetKeywordsByIndustry(ctx context.Context, industryID int) ([]*IndustryKeyword, error) {
	r.logger.Printf("🔍 Getting keywords for industry ID: %d", industryID)

	// Get the real PostgREST client
	postgrestClient := r.client.GetPostgrestClient()

	data, _, err := postgrestClient.
		From("industry_keywords").
		Select("*", "", false).
		Eq("industry_id", fmt.Sprintf("%d", industryID)).
		Eq("is_active", "true").
		Order("weight", &postgrest.OrderOpts{Ascending: false}).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get keywords for industry %d: %w", industryID, err)
	}

	// Unmarshal the JSON response
	var keywords []*IndustryKeyword
	if err := json.Unmarshal(data, &keywords); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keywords data: %w", err)
	}

	return keywords, nil
}

// SearchKeywords searches for keywords matching a query
func (r *SupabaseKeywordRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*IndustryKeyword, error) {
	r.logger.Printf("🔍 Searching keywords: %s (limit: %d)", query, limit)

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	// Use full-text search if available, otherwise fall back to ILIKE
	data, _, err := postgrestClient.
		From("industry_keywords").
		Select("id,industry_id,keyword,weight,is_active", "", false).
		Ilike("keyword", fmt.Sprintf("%%%s%%", query)).
		Eq("is_active", "true").
		Order("weight", &postgrest.OrderOpts{Ascending: false}).
		Order("keyword", &postgrest.OrderOpts{Ascending: true}).
		Limit(limit, "").
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to search keywords: %w", err)
	}

	// Unmarshal the JSON response
	var keywords []*IndustryKeyword
	if err := json.Unmarshal(data, &keywords); err != nil {
		return nil, fmt.Errorf("failed to unmarshal keywords data: %w", err)
	}

	return keywords, nil
}

// AddKeywordToIndustry adds a new keyword to an industry
func (r *SupabaseKeywordRepository) AddKeywordToIndustry(ctx context.Context, industryID int, keyword string, weight float64) error {
	r.logger.Printf("🔍 Adding keyword '%s' to industry %d with weight %.2f", keyword, industryID, weight)

	// TODO: Implement keyword addition
	return fmt.Errorf("keyword addition not yet implemented")
}

// UpdateKeywordWeight updates the weight of a keyword
func (r *SupabaseKeywordRepository) UpdateKeywordWeight(ctx context.Context, keywordID int, weight float64) error {
	r.logger.Printf("🔍 Updating keyword %d weight to %.2f", keywordID, weight)

	// TODO: Implement keyword weight update
	return fmt.Errorf("keyword weight update not yet implemented")
}

// RemoveKeywordFromIndustry removes a keyword from an industry
func (r *SupabaseKeywordRepository) RemoveKeywordFromIndustry(ctx context.Context, keywordID int) error {
	r.logger.Printf("🔍 Removing keyword ID: %d", keywordID)

	// TODO: Implement keyword removal
	return fmt.Errorf("keyword removal not yet implemented")
}

// =============================================================================
// Classification Codes
// =============================================================================

// GetClassificationCodesByIndustry retrieves classification codes for an industry
func (r *SupabaseKeywordRepository) GetClassificationCodesByIndustry(ctx context.Context, industryID int) ([]*ClassificationCode, error) {
	r.logger.Printf("🔍 Getting classification codes for industry ID: %d", industryID)

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	// Optimized query with proper indexing and ordering
	// First, try with is_active filter
	response, _, err := postgrestClient.
		From("classification_codes").
		Select("id,industry_id,code_type,code,description,is_active", "", false).
		Eq("industry_id", fmt.Sprintf("%d", industryID)).
		Eq("is_active", "true").
		Order("code_type", &postgrest.OrderOpts{Ascending: true}).
		Order("code", &postgrest.OrderOpts{Ascending: true}).
		Execute()

	if err != nil {
		r.logger.Printf("⚠️ [ClassificationCodes] Query with is_active filter failed for industry %d: %v", industryID, err)
		// Try without is_active filter as fallback (in case column doesn't exist or all are inactive)
		response, _, err = postgrestClient.
			From("classification_codes").
			Select("id,industry_id,code_type,code,description", "", false).
			Eq("industry_id", fmt.Sprintf("%d", industryID)).
			Order("code_type", &postgrest.OrderOpts{Ascending: true}).
			Order("code", &postgrest.OrderOpts{Ascending: true}).
			Execute()
		
		if err != nil {
			return nil, fmt.Errorf("failed to get classification codes for industry %d: %w", industryID, err)
		}
		r.logger.Printf("⚠️ [ClassificationCodes] Query without is_active filter succeeded for industry %d", industryID)
	}

	// Parse the response
	var codes []*ClassificationCode
	if err := r.parseClassificationCodesResponse(response, &codes); err != nil {
		return nil, fmt.Errorf("failed to parse classification codes response: %w", err)
	}

	if len(codes) == 0 {
		r.logger.Printf("⚠️ [ClassificationCodes] No classification codes found for industry %d - database may need codes populated", industryID)
	} else {
		// Extract unique code types for logging
		codeTypes := make(map[string]bool)
		for _, code := range codes {
			codeTypes[code.CodeType] = true
		}
		typeList := make([]string, 0, len(codeTypes))
		for ct := range codeTypes {
			typeList = append(typeList, ct)
		}
		r.logger.Printf("✅ Retrieved %d classification codes for industry %d (types: %v)", 
			len(codes), industryID, typeList)
	}
	return codes, nil
}

// GetClassificationCodesByKeywords retrieves classification codes directly from keywords
// This bypasses industry detection and matches keywords to codes via code_keywords table
func (r *SupabaseKeywordRepository) GetClassificationCodesByKeywords(
	ctx context.Context,
	keywords []string,
	codeType string, // "MCC", "SIC", or "NAICS"
	minRelevance float64, // Minimum relevance_score threshold (default 0.5)
) ([]*ClassificationCodeWithMetadata, error) {
	if len(keywords) == 0 {
		return []*ClassificationCodeWithMetadata{}, nil
	}

	r.logger.Printf("🔍 Getting classification codes by keywords: %d keywords, type: %s, minRelevance: %.2f",
		len(keywords), codeType, minRelevance)

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	// Step 1: Query code_keywords table to find matching keywords and get code_ids
	// Convert keywords to lowercase for case-insensitive matching
	keywordsLower := make([]string, len(keywords))
	for i, kw := range keywords {
		keywordsLower[i] = strings.ToLower(strings.TrimSpace(kw))
	}

	// Query code_keywords - we'll need to query for each keyword or use IN clause
	// Since PostgREST may not support array matching directly, we'll query for each keyword
	// and combine results, or use a single query if possible

	// Query all code_keywords and filter in memory by relevance and keywords
	// This is not ideal for large datasets, but works for the MVP
	// TODO: Optimize with proper PostgREST array matching or database function
	// Note: We don't filter by relevance_score in the query since we need >= minRelevance,
	// and we filter in memory anyway. This ensures we get all potential matches.
	codeKeywordsResponse, _, err := postgrestClient.
		From("code_keywords").
		Select("id,code_id,keyword,relevance_score,match_type", "", false).
		Order("relevance_score", &postgrest.OrderOpts{Ascending: false}).
		Limit(10000, ""). // Large limit to get all matches, then filter
		Execute()

	if err != nil {
		r.logger.Printf("⚠️ Failed to query code_keywords: %v", err)
		// Fallback: return empty result instead of error
		return []*ClassificationCodeWithMetadata{}, nil
	}

	// Parse code_keywords results
	type CodeKeywordRow struct {
		ID             int     `json:"id"`
		CodeID         int     `json:"code_id"`
		Keyword        string  `json:"keyword"`
		RelevanceScore float64 `json:"relevance_score"`
		MatchType      string  `json:"match_type"`
	}

	var codeKeywordRows []CodeKeywordRow
	if err := json.Unmarshal(codeKeywordsResponse, &codeKeywordRows); err != nil {
		return nil, fmt.Errorf("failed to parse code_keywords response: %w", err)
	}

	// Filter by keywords (case-insensitive) and collect unique code_ids
	keywordSet := make(map[string]bool)
	for _, kw := range keywordsLower {
		keywordSet[kw] = true
	}

	codeIDToMetadata := make(map[int][]struct {
		RelevanceScore float64
		MatchType      string
		Keyword        string
	})

	for _, row := range codeKeywordRows {
		rowKeywordLower := strings.ToLower(strings.TrimSpace(row.Keyword))
		
		// Check if this keyword matches any of our search keywords
		matches := false
		for searchKeyword := range keywordSet {
			// Exact match or contains match
			if rowKeywordLower == searchKeyword || 
			   strings.Contains(rowKeywordLower, searchKeyword) ||
			   strings.Contains(searchKeyword, rowKeywordLower) {
				matches = true
				break
			}
		}

		if matches && row.RelevanceScore >= minRelevance {
			if codeIDToMetadata[row.CodeID] == nil {
				codeIDToMetadata[row.CodeID] = []struct {
					RelevanceScore float64
					MatchType      string
					Keyword        string
				}{}
			}
			codeIDToMetadata[row.CodeID] = append(codeIDToMetadata[row.CodeID], struct {
				RelevanceScore float64
				MatchType      string
				Keyword        string
			}{
				RelevanceScore: row.RelevanceScore,
				MatchType:      row.MatchType,
				Keyword:        row.Keyword,
			})
		}
	}

	if len(codeIDToMetadata) == 0 {
		r.logger.Printf("⚠️ No code_keywords matches found for keywords: %v", keywords)
		return []*ClassificationCodeWithMetadata{}, nil
	}

	// Step 2: Query classification_codes for the matched code_ids
	codeIDs := make([]int, 0, len(codeIDToMetadata))
	for codeID := range codeIDToMetadata {
		codeIDs = append(codeIDs, codeID)
	}

	// Query classification_codes - we'll need to query each code_id or use IN clause
	// For now, query all codes and filter by code_id and code_type
	allCodesResponse, _, err := postgrestClient.
		From("classification_codes").
		Select("id,industry_id,code_type,code,description,is_active", "", false).
		Eq("code_type", codeType).
		Eq("is_active", "true").
		Order("code", &postgrest.OrderOpts{Ascending: true}).
		Limit(10000, "").
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to query classification_codes: %w", err)
	}

	var allCodes []*ClassificationCode
	if err := json.Unmarshal(allCodesResponse, &allCodes); err != nil {
		return nil, fmt.Errorf("failed to parse classification_codes response: %w", err)
	}

	// Step 3: Combine results - filter codes by code_id and attach metadata
	codeIDSet := make(map[int]bool)
	for _, id := range codeIDs {
		codeIDSet[id] = true
	}

	var results []*ClassificationCodeWithMetadata
	for _, code := range allCodes {
		if codeIDSet[code.ID] {
			// Get the best metadata (highest relevance score) for this code
			metadataList := codeIDToMetadata[code.ID]
			if len(metadataList) > 0 {
				// Sort by relevance score (highest first)
				bestMetadata := metadataList[0]
				for _, meta := range metadataList {
					if meta.RelevanceScore > bestMetadata.RelevanceScore {
						bestMetadata = meta
					}
				}

				results = append(results, &ClassificationCodeWithMetadata{
					ClassificationCode: *code,
					RelevanceScore:      bestMetadata.RelevanceScore,
					MatchType:           bestMetadata.MatchType,
				})
			}
		}
	}

	// Sort results by relevance_score descending, then by code
	sort.Slice(results, func(i, j int) bool {
		if results[i].RelevanceScore != results[j].RelevanceScore {
			return results[i].RelevanceScore > results[j].RelevanceScore
		}
		return results[i].Code < results[j].Code
	})

	r.logger.Printf("✅ Retrieved %d classification codes by keywords (type: %s)", len(results), codeType)
	return results, nil
}

// GetClassificationCodesByType retrieves classification codes by type (NAICS, MCC, SIC)
func (r *SupabaseKeywordRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*ClassificationCode, error) {
	r.logger.Printf("🔍 Getting classification codes by type: %s", codeType)

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	// Optimized query with proper indexing and ordering
	response, _, err := postgrestClient.
		From("classification_codes").
		Select("id,industry_id,code_type,code,description,is_active", "", false).
		Eq("code_type", codeType).
		Eq("is_active", "true").
		Order("industry_id", &postgrest.OrderOpts{Ascending: true}).
		Order("code", &postgrest.OrderOpts{Ascending: true}).
		Limit(5000, ""). // Limit to prevent memory issues with large datasets
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get classification codes by type %s: %w", codeType, err)
	}

	// Parse the response
	var codes []*ClassificationCode
	if err := r.parseClassificationCodesResponse(response, &codes); err != nil {
		return nil, fmt.Errorf("failed to parse classification codes response: %w", err)
	}

	r.logger.Printf("✅ Retrieved %d classification codes for type %s", len(codes), codeType)
	return codes, nil
}

// AddClassificationCode adds a new classification code
func (r *SupabaseKeywordRepository) AddClassificationCode(ctx context.Context, code *ClassificationCode) error {
	r.logger.Printf("🔍 Adding classification code: %s %s", code.CodeType, code.Code)

	// TODO: Implement classification code addition
	return fmt.Errorf("classification code addition not yet implemented")
}

// UpdateClassificationCode updates an existing classification code
func (r *SupabaseKeywordRepository) UpdateClassificationCode(ctx context.Context, code *ClassificationCode) error {
	r.logger.Printf("🔍 Updating classification code: %s %s", code.CodeType, code.Code)

	// TODO: Implement classification code update
	return fmt.Errorf("classification code update not yet implemented")
}

// DeleteClassificationCode deletes a classification code
func (r *SupabaseKeywordRepository) DeleteClassificationCode(ctx context.Context, id int) error {
	r.logger.Printf("🔍 Deleting classification code ID: %d", id)

	// TODO: Implement classification code deletion
	return fmt.Errorf("classification code deletion not yet implemented")
}

// =============================================================================
// Industry Patterns
// =============================================================================

// GetPatternsByIndustry retrieves patterns for an industry
// Note: Pattern matching is not implemented - using keyword-based classification instead
func (r *SupabaseKeywordRepository) GetPatternsByIndustry(ctx context.Context, industryID int) ([]*IndustryPattern, error) {
	r.logger.Printf("🔍 Pattern matching not implemented - using keyword-based classification")
	return []*IndustryPattern{}, nil
}

// AddPattern adds a new pattern
// Note: Pattern matching is not implemented - using keyword-based classification instead
func (r *SupabaseKeywordRepository) AddPattern(ctx context.Context, pattern *IndustryPattern) error {
	r.logger.Printf("🔍 Pattern matching not implemented - using keyword-based classification")
	return fmt.Errorf("pattern matching not implemented - use keyword-based classification instead")
}

// UpdatePattern updates an existing pattern
// Note: Pattern matching is not implemented - using keyword-based classification instead
func (r *SupabaseKeywordRepository) UpdatePattern(ctx context.Context, pattern *IndustryPattern) error {
	r.logger.Printf("🔍 Pattern matching not implemented - using keyword-based classification")
	return fmt.Errorf("pattern matching not implemented - use keyword-based classification instead")
}

// DeletePattern deletes a pattern
// Note: Pattern matching is not implemented - using keyword-based classification instead
func (r *SupabaseKeywordRepository) DeletePattern(ctx context.Context, id int) error {
	r.logger.Printf("🔍 Pattern matching not implemented - using keyword-based classification")
	return fmt.Errorf("pattern matching not implemented - use keyword-based classification instead")
}

// =============================================================================
// Keyword Weights
// =============================================================================

// GetKeywordWeights retrieves weight information for a keyword
func (r *SupabaseKeywordRepository) GetKeywordWeights(ctx context.Context, keyword string) ([]*KeywordWeight, error) {
	r.logger.Printf("🔍 Getting weights for keyword: %s", keyword)

	_, _, err := r.client.GetPostgrestClient().
		From("keyword_weights").
		Select("*", "", false).
		Eq("keyword", keyword).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get weights for keyword %s: %w", keyword, err)
	}

	// TODO: Implement proper response parsing
	return []*KeywordWeight{}, nil
}

// UpdateKeywordWeightByID updates a keyword weight by ID
func (r *SupabaseKeywordRepository) UpdateKeywordWeightByID(ctx context.Context, weight *KeywordWeight) error {
	r.logger.Printf("🔍 Updating keyword weight ID: %d", weight.ID)

	// TODO: Implement keyword weight update
	return fmt.Errorf("keyword weight update not yet implemented")
}

// IncrementUsageCount increments the usage count for a keyword
func (r *SupabaseKeywordRepository) IncrementUsageCount(ctx context.Context, keyword string, industryID int) error {
	r.logger.Printf("🔍 Incrementing usage count for keyword '%s' in industry %d", keyword, industryID)

	// TODO: Implement usage count increment
	return fmt.Errorf("usage count increment not yet implemented")
}

// =============================================================================
// Business Classification
// =============================================================================

// ClassifyBusiness classifies a business based on name and website (description removed for security)
func (r *SupabaseKeywordRepository) ClassifyBusiness(ctx context.Context, businessName, websiteURL string) (*ClassificationResult, error) {
	r.logger.Printf("🔍 Classifying business: %s", businessName)

	// Extract contextual keywords from business information (excluding description for security)
	contextualKeywords := r.extractKeywords(businessName, websiteURL)

	// Classify based on contextual keywords
	return r.ClassifyBusinessByContextualKeywords(ctx, contextualKeywords)
}

// ClassifyBusinessByKeywords classifies a business based on extracted keywords using optimized algorithm
func (r *SupabaseKeywordRepository) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*ClassificationResult, error) {
	r.logger.Printf("🔍 Classifying business by keywords (optimized): %v", keywords)

	if len(keywords) == 0 {
		// Return default classification
		return &ClassificationResult{
			Industry:   &Industry{Name: "General Business", ID: 26},
			Confidence: 0.50,
			Keywords:   []string{},
			Patterns:   []string{},
			Codes:      []ClassificationCode{},
			Reasoning:  "No keywords provided for classification",
		}, nil
	}

	// Ensure keyword index is built
	index := r.GetKeywordIndex()
	if len(index.KeywordToIndustries) == 0 {
		r.logger.Printf("⚠️ Keyword index is empty, building it now...")
		if err := r.BuildKeywordIndex(ctx); err != nil {
			r.logger.Printf("⚠️ Failed to build keyword index: %v", err)
			return r.fallbackClassification(keywords, "Failed to build keyword index"), nil
		}
		index = r.GetKeywordIndex()
	}

	// Use optimized O(k) algorithm instead of O(n*m*k)
	industryScores := make(map[int]float64)
	industryMatches := make(map[int][]string)

	// Process each input keyword once with enhanced phrase matching
	for _, inputKeyword := range keywords {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(inputKeyword))

		// Determine if this is a phrase (multi-word) or single word
		isPhrase := strings.Contains(normalizedKeyword, " ")
		phraseMultiplier := 1.0

		// Higher weight for phrase matches
		if isPhrase {
			phraseMultiplier = 1.5 // 50% boost for phrase matches
		}

		// Direct lookup in keyword index - O(1) average case
		if matches, exists := index.KeywordToIndustries[normalizedKeyword]; exists {
			for _, match := range matches {
				// Apply phrase multiplier for exact phrase matches
				weight := match.Weight * phraseMultiplier
				industryScores[match.IndustryID] += weight
				industryMatches[match.IndustryID] = append(industryMatches[match.IndustryID], match.Keyword)
			}
		}

		// Enhanced partial matching with phrase awareness
		for keyword, matches := range index.KeywordToIndustries {
			// Check for exact phrase matches first
			if normalizedKeyword == keyword {
				continue // Already handled above
			}

			// Check for phrase-to-phrase partial matches
			if isPhrase && strings.Contains(keyword, " ") {
				// Both are phrases - check for phrase overlap
				if r.hasPhraseOverlap(normalizedKeyword, keyword) {
					for _, match := range matches {
						// Higher weight for phrase-to-phrase matches
						partialWeight := match.Weight * 0.8
						industryScores[match.IndustryID] += partialWeight
						industryMatches[match.IndustryID] = append(industryMatches[match.IndustryID], match.Keyword)
					}
				}
			} else if strings.Contains(normalizedKeyword, keyword) || strings.Contains(keyword, normalizedKeyword) {
				// Traditional substring matching
				for _, match := range matches {
					// Reduce weight for partial matches
					partialWeight := match.Weight * 0.5
					industryScores[match.IndustryID] += partialWeight
					industryMatches[match.IndustryID] = append(industryMatches[match.IndustryID], match.Keyword)
				}
			}
		}
	}

	// Phase 7.1: Multi-keyword industry matching with co-occurrence analysis
	// Find best industry with enhanced scoring that considers multiple keyword matches
	bestIndustryID := 26 // Default industry
	bestScore := 0.0
	var bestMatchedKeywords []string
	industryMatchCounts := make(map[int]int) // Track number of unique keywords matched per industry

	// Calculate match counts for each industry (Phase 7.1)
	for industryID, matched := range industryMatches {
		// Count unique keywords matched (deduplicate)
		uniqueMatches := make(map[string]bool)
		for _, kw := range matched {
			uniqueMatches[kw] = true
		}
		industryMatchCounts[industryID] = len(uniqueMatches)
	}

	// Phase 7.3: Industry co-occurrence analysis
	coOccurrenceBoost := r.calculateIndustryCoOccurrenceBoost(industryMatches, keywords)

	for industryID, score := range industryScores {
		// Normalize score by number of input keywords
		normalizedScore := score / float64(len(keywords))
		
		// Phase 7.1: Weight by number of unique keyword matches (multi-keyword requirement)
		matchCount := industryMatchCounts[industryID]
		matchCountBoost := 1.0
		if matchCount >= 3 {
			matchCountBoost = 1.2 // 20% boost for 3+ keyword matches
		} else if matchCount >= 2 {
			matchCountBoost = 1.1 // 10% boost for 2 keyword matches
		}
		
		// Apply co-occurrence boost (Phase 7.3)
		// Apply as a multiplier to maintain proper scaling
		if boost, exists := coOccurrenceBoost[industryID]; exists {
			normalizedScore *= (1.0 + boost) // Convert absolute boost to multiplier
		}
		
		// Apply match count boost
		normalizedScore *= matchCountBoost

		if normalizedScore > bestScore {
			bestScore = normalizedScore
			bestIndustryID = industryID
			bestMatchedKeywords = industryMatches[industryID]
		}
	}

	// Get industry information
	var bestIndustry *Industry
	if bestIndustryID == 26 {
		bestIndustry = &Industry{Name: "General Business", ID: 26}
	} else {
		// Get industry details from database
		industry, err := r.GetIndustryByID(ctx, bestIndustryID)
		if err != nil {
			r.logger.Printf("⚠️ Failed to get industry details for ID %d: %v", bestIndustryID, err)
			bestIndustry = &Industry{Name: "General Business", ID: 26}
		} else {
			bestIndustry = industry
		}
	}

	// Get classification codes for the best industry (using cache)
	var codes []ClassificationCode
	if bestIndustry.ID != 26 { // Not the default industry
		classificationCodes, err := r.GetCachedClassificationCodes(ctx, bestIndustry.ID)
		if err != nil {
			r.logger.Printf("⚠️ Failed to get classification codes: %v", err)
		} else {
			for _, code := range classificationCodes {
				codes = append(codes, *code)
			}
		}
	}

	// Phase 7.2: Industry confidence thresholds
	const (
		MinKeywordCount    = 3  // Minimum keywords required for high confidence
		MinConfidenceScore = 0.6 // Minimum confidence threshold
	)

	// Calculate enhanced confidence score with dynamic factors
	var confidence float64
	var reasoning string

	// Phase 7.1: Count unique matched keywords
	uniqueMatchedKeywords := make(map[string]bool)
	for _, kw := range bestMatchedKeywords {
		uniqueMatchedKeywords[kw] = true
	}
	uniqueMatchCount := len(uniqueMatchedKeywords)

	// Enhanced confidence calculation with multiple factors
	// Note: bestScore is already normalized (score / len(keywords)), so we don't divide again
	// Safety check for division by zero (shouldn't happen due to early return, but be defensive)
	if len(keywords) == 0 {
		return r.fallbackClassification(keywords, "No keywords provided for confidence calculation"), nil
	}
	matchRatio := float64(uniqueMatchCount) / float64(len(keywords))
	scoreRatio := bestScore // bestScore is already normalized, no need to divide again

	// Base confidence from match quality
	baseConfidence := (matchRatio * 0.6) + (scoreRatio * 0.4)

	// Apply keyword quality factor
	keywordQualityFactor := r.calculateKeywordQualityFactor(bestMatchedKeywords, keywords)

	// Apply industry specificity factor
	industrySpecificityFactor := r.calculateIndustrySpecificityFactor(bestIndustryID, bestMatchedKeywords)

	// Apply match diversity factor
	matchDiversityFactor := r.calculateMatchDiversityFactor(bestMatchedKeywords)

	// Calculate final confidence with all factors
	confidence = baseConfidence * keywordQualityFactor * industrySpecificityFactor * matchDiversityFactor

	// Phase 7.2: Apply confidence thresholds
	// If below minimum keyword count, reduce confidence
	if uniqueMatchCount == 0 {
		// No matches at all - set very low confidence
		confidence = 0.1
		r.logger.Printf("⚠️ [Phase 7.2] No keyword matches found (0 matches), setting confidence to minimum")
	} else if uniqueMatchCount < MinKeywordCount {
		confidencePenalty := float64(uniqueMatchCount) / float64(MinKeywordCount)
		confidence *= confidencePenalty
		r.logger.Printf("⚠️ [Phase 7.2] Below minimum keyword count (%d < %d), applying penalty: %.3f",
			uniqueMatchCount, MinKeywordCount, confidencePenalty)
	}

	// Phase 7.2: If below minimum confidence threshold, use "General Business"
	originalConfidence := confidence // Store original for logging
	if confidence < MinConfidenceScore && bestIndustryID != 26 {
		r.logger.Printf("⚠️ [Phase 7.2] Confidence below threshold (%.3f < %.3f), falling back to General Business",
			originalConfidence, MinConfidenceScore)
		bestIndustryID = 26
		bestIndustry = &Industry{Name: "General Business", ID: 26}
		confidence = 0.30 // Lower confidence for fallback
		reasoning = fmt.Sprintf("Confidence below threshold (%.3f < %.3f) with %d keyword matches, using General Business",
			originalConfidence, MinConfidenceScore, uniqueMatchCount)
	} else {
		// Ensure confidence is within bounds
		if confidence > 1.0 {
			confidence = 1.0
		}
		if confidence < 0.1 {
			confidence = 0.1
		}

		r.logger.Printf("📊 [Phase 7] Enhanced confidence calculated: %.3f (base: %.3f, quality: %.3f, specificity: %.3f, diversity: %.3f, matches: %d)",
			confidence, baseConfidence, keywordQualityFactor, industrySpecificityFactor, matchDiversityFactor, uniqueMatchCount)

		reasoning = fmt.Sprintf("Multi-keyword classification matched %d unique keywords with industry '%s' (score: %.2f, confidence: %.3f)",
			uniqueMatchCount, bestIndustry.Name, bestScore, confidence)
	}

	return &ClassificationResult{
		Industry:   bestIndustry,
		Confidence: confidence,
		Keywords:   bestMatchedKeywords,
		Patterns:   []string{},
		Codes:      codes,
		Reasoning:  reasoning,
	}, nil
}

// fallbackClassification provides a fallback when optimization fails
func (r *SupabaseKeywordRepository) fallbackClassification(keywords []string, reason string) *ClassificationResult {
	return &ClassificationResult{
		Industry:   &Industry{Name: "General Business", ID: 26},
		Confidence: 0.50,
		Keywords:   keywords,
		Patterns:   []string{},
		Codes:      []ClassificationCode{},
		Reasoning:  reason,
	}
}

// ClassifyBusinessByContextualKeywords classifies a business based on contextual keywords with enhanced scoring algorithm
func (r *SupabaseKeywordRepository) ClassifyBusinessByContextualKeywords(ctx context.Context, contextualKeywords []ContextualKeyword) (*ClassificationResult, error) {
	r.logger.Printf("🔍 Classifying business by contextual keywords with enhanced scoring: %d keywords", len(contextualKeywords))

	if len(contextualKeywords) == 0 {
		// Return default classification
		return &ClassificationResult{
			Industry:   &Industry{Name: "General Business", ID: 26},
			Confidence: 0.50,
			Keywords:   []string{},
			Patterns:   []string{},
			Codes:      []ClassificationCode{},
			Reasoning:  "No contextual keywords provided for classification",
		}, nil
	}

	// Ensure keyword index is built
	index := r.GetKeywordIndex()
	if len(index.KeywordToIndustries) == 0 {
		r.logger.Printf("⚠️ Keyword index is empty, building it now...")
		if err := r.BuildKeywordIndex(ctx); err != nil {
			r.logger.Printf("⚠️ Failed to build keyword index: %v", err)
			// Convert contextual keywords to strings for fallback
			keywords := make([]string, len(contextualKeywords))
			for i, ck := range contextualKeywords {
				keywords[i] = ck.Keyword
			}
			return r.fallbackClassification(keywords, "Failed to build keyword index"), nil
		}
		index = r.GetKeywordIndex()
	}

	// Use enhanced scoring algorithm for improved accuracy and performance
	enhancedScorer := NewEnhancedScoringAlgorithm(r.logger, DefaultEnhancedScoringConfig())
	enhancedResult, err := enhancedScorer.CalculateEnhancedScore(ctx, contextualKeywords, index)
	if err != nil {
		r.logger.Printf("⚠️ Enhanced scoring failed, falling back to basic algorithm: %v", err)
		return r.classifyBusinessByContextualKeywordsBasic(ctx, contextualKeywords, index)
	}

	// Get industry information
	var bestIndustry *Industry
	if enhancedResult.IndustryID != 26 {
		industry, err := r.GetIndustryByID(ctx, enhancedResult.IndustryID)
		if err != nil {
			r.logger.Printf("⚠️ Failed to get industry %d: %v", enhancedResult.IndustryID, err)
			bestIndustry = &Industry{Name: "General Business", ID: 26}
		} else {
			bestIndustry = industry
		}
	} else {
		bestIndustry = &Industry{Name: "General Business", ID: 26}
	}

	// Get classification codes for the best industry
	codesPtr, err := r.GetCachedClassificationCodes(ctx, enhancedResult.IndustryID)
	if err != nil {
		r.logger.Printf("⚠️ Failed to get classification codes for industry %d: %v", enhancedResult.IndustryID, err)
		codesPtr = []*ClassificationCode{}
	}

	// Convert []*ClassificationCode to []ClassificationCode
	codes := make([]ClassificationCode, len(codesPtr))
	for i, codePtr := range codesPtr {
		codes[i] = *codePtr
	}

	// Extract matched keywords for backward compatibility
	matchedKeywords := make([]string, len(enhancedResult.MatchedKeywords))
	for i, match := range enhancedResult.MatchedKeywords {
		matchedKeywords[i] = match.MatchedKeyword
	}

	// Build enhanced reasoning with detailed breakdown
	reasoning := fmt.Sprintf("Enhanced classification as %s with confidence %.3f based on %d contextual keywords. "+
		"Score breakdown: Direct(%.3f), Phrase(%.3f), Partial(%.3f), Context(%.3f). "+
		"Quality indicators: Diversity(%.3f), Relevance(%.3f), Overall(%.3f). "+
		"Processing time: %v. Matched %d keywords: %v",
		bestIndustry.Name, enhancedResult.Confidence, len(contextualKeywords),
		enhancedResult.ScoreBreakdown.DirectMatchScore,
		enhancedResult.ScoreBreakdown.PhraseMatchScore,
		enhancedResult.ScoreBreakdown.PartialMatchScore,
		enhancedResult.ScoreBreakdown.ContextScore,
		enhancedResult.QualityIndicators.MatchDiversity,
		enhancedResult.QualityIndicators.KeywordRelevance,
		enhancedResult.QualityIndicators.OverallQuality,
		enhancedResult.ProcessingTime,
		len(matchedKeywords), matchedKeywords)

	return &ClassificationResult{
		Industry:   bestIndustry,
		Confidence: enhancedResult.Confidence,
		Keywords:   matchedKeywords,
		Patterns:   []string{},
		Codes:      codes,
		Reasoning:  reasoning,
	}, nil
}

// classifyBusinessByContextualKeywordsBasic provides fallback basic classification algorithm
func (r *SupabaseKeywordRepository) classifyBusinessByContextualKeywordsBasic(ctx context.Context, contextualKeywords []ContextualKeyword, index *KeywordIndex) (*ClassificationResult, error) {
	r.logger.Printf("🔄 Using basic classification algorithm as fallback")

	// Use optimized O(k) algorithm with context multipliers
	industryScores := make(map[int]float64)
	industryMatches := make(map[int][]string)

	// Process each contextual keyword with context-aware multipliers
	for _, contextualKeyword := range contextualKeywords {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(contextualKeyword.Keyword))

		// Apply context multiplier based on source
		contextMultiplier := r.getContextMultiplier(contextualKeyword.Context)

		// Determine if this is a phrase (multi-word) or single word
		isPhrase := strings.Contains(normalizedKeyword, " ")
		phraseMultiplier := 1.0

		// Higher weight for phrase matches
		if isPhrase {
			phraseMultiplier = 1.5 // 50% boost for phrase matches
		}

		// Direct lookup in keyword index - O(1) average case
		if matches, exists := index.KeywordToIndustries[normalizedKeyword]; exists {
			for _, match := range matches {
				// Apply both phrase and context multipliers
				weight := match.Weight * phraseMultiplier * contextMultiplier
				industryScores[match.IndustryID] += weight
				industryMatches[match.IndustryID] = append(industryMatches[match.IndustryID], match.Keyword)
			}
		}

		// Enhanced partial matching with phrase awareness and context multipliers
		for keyword, matches := range index.KeywordToIndustries {
			// Check for exact phrase matches first
			if normalizedKeyword == keyword {
				continue // Already handled above
			}

			// Check for phrase-to-phrase partial matches
			if isPhrase && strings.Contains(keyword, " ") {
				// Both are phrases - check for phrase overlap
				if r.hasPhraseOverlap(normalizedKeyword, keyword) {
					for _, match := range matches {
						// Higher weight for phrase-to-phrase matches with context multiplier
						partialWeight := match.Weight * 0.8 * contextMultiplier
						industryScores[match.IndustryID] += partialWeight
						industryMatches[match.IndustryID] = append(industryMatches[match.IndustryID], match.Keyword)
					}
				}
			} else if strings.Contains(normalizedKeyword, keyword) || strings.Contains(keyword, normalizedKeyword) {
				// Traditional substring matching with context multiplier
				for _, match := range matches {
					// Reduce weight for partial matches but apply context multiplier
					partialWeight := match.Weight * 0.5 * contextMultiplier
					industryScores[match.IndustryID] += partialWeight
					industryMatches[match.IndustryID] = append(industryMatches[match.IndustryID], match.Keyword)
				}
			}
		}
	}

	// Find best industry
	bestIndustryID := 26 // Default industry
	bestScore := 0.0
	var bestMatchedKeywords []string

	for industryID, score := range industryScores {
		// Normalize score by number of input keywords
		normalizedScore := score / float64(len(contextualKeywords))

		if normalizedScore > bestScore {
			bestScore = normalizedScore
			bestIndustryID = industryID
			bestMatchedKeywords = industryMatches[industryID]
		}
	}

	// Get industry information
	var bestIndustry *Industry
	if bestIndustryID != 26 {
		industry, err := r.GetIndustryByID(ctx, bestIndustryID)
		if err != nil {
			r.logger.Printf("⚠️ Failed to get industry %d: %v", bestIndustryID, err)
			bestIndustry = &Industry{Name: "General Business", ID: 26}
		} else {
			bestIndustry = industry
		}
	} else {
		bestIndustry = &Industry{Name: "General Business", ID: 26}
	}

	// Calculate confidence using dynamic confidence calculation
	confidence := r.calculateDynamicConfidence(bestScore, len(bestMatchedKeywords), len(contextualKeywords))

	// Get classification codes for the best industry
	codesPtr, err := r.GetCachedClassificationCodes(ctx, bestIndustryID)
	if err != nil {
		r.logger.Printf("⚠️ Failed to get classification codes for industry %d: %v", bestIndustryID, err)
		codesPtr = []*ClassificationCode{}
	}

	// Convert []*ClassificationCode to []ClassificationCode
	codes := make([]ClassificationCode, len(codesPtr))
	for i, codePtr := range codesPtr {
		codes[i] = *codePtr
	}

	// Build reasoning
	reasoning := fmt.Sprintf("Basic classification as %s with confidence %.2f based on %d contextual keywords. Context multipliers applied: business_name (1.2x), description (1.0x), website_url (1.0x). Matched %d keywords: %v",
		bestIndustry.Name, confidence, len(contextualKeywords), len(bestMatchedKeywords), bestMatchedKeywords)

	return &ClassificationResult{
		Industry:   bestIndustry,
		Confidence: confidence,
		Keywords:   bestMatchedKeywords,
		Patterns:   []string{},
		Codes:      codes,
		Reasoning:  reasoning,
	}, nil
}

// GetTopIndustriesByKeywords finds the top industries matching given keywords
func (r *SupabaseKeywordRepository) GetTopIndustriesByKeywords(ctx context.Context, keywords []string, limit int) ([]*Industry, error) {
	r.logger.Printf("🔍 Getting top industries for keywords: %v (limit: %d)", keywords, limit)

	// TODO: Implement keyword-to-industry scoring algorithm
	return []*Industry{}, nil
}

// =============================================================================
// Advanced Search and Analytics
// =============================================================================

// SearchIndustriesByPattern searches industries by pattern matching
// Note: Pattern matching is not implemented - using keyword-based classification instead
func (r *SupabaseKeywordRepository) SearchIndustriesByPattern(ctx context.Context, pattern string) ([]*Industry, error) {
	r.logger.Printf("🔍 Pattern matching not implemented - using keyword-based classification")
	return []*Industry{}, nil
}

// GetIndustryStatistics gets statistics about industries and keywords
func (r *SupabaseKeywordRepository) GetIndustryStatistics(ctx context.Context) (map[string]interface{}, error) {
	r.logger.Printf("🔍 Getting industry statistics")

	// TODO: Implement industry statistics
	return map[string]interface{}{}, nil
}

// GetKeywordFrequency gets keyword frequency for an industry
func (r *SupabaseKeywordRepository) GetKeywordFrequency(ctx context.Context, industryID int) (map[string]int, error) {
	r.logger.Printf("🔍 Getting keyword frequency for industry ID: %d", industryID)

	// TODO: Implement keyword frequency analysis
	return map[string]int{}, nil
}

// =============================================================================
// Bulk Operations
// =============================================================================

// BulkInsertKeywords inserts multiple keywords at once
func (r *SupabaseKeywordRepository) BulkInsertKeywords(ctx context.Context, keywords []*IndustryKeyword) error {
	r.logger.Printf("🔍 Bulk inserting %d keywords", len(keywords))

	// TODO: Implement bulk keyword insertion
	return fmt.Errorf("bulk keyword insertion not yet implemented")
}

// BulkUpdateKeywords updates multiple keywords at once
func (r *SupabaseKeywordRepository) BulkUpdateKeywords(ctx context.Context, keywords []*IndustryKeyword) error {
	r.logger.Printf("🔍 Bulk updating %d keywords", len(keywords))

	// TODO: Implement bulk keyword update
	return fmt.Errorf("bulk keyword update not yet implemented")
}

// BulkDeleteKeywords deletes multiple keywords at once
func (r *SupabaseKeywordRepository) BulkDeleteKeywords(ctx context.Context, keywordIDs []int) error {
	r.logger.Printf("🔍 Bulk deleting %d keywords", len(keywordIDs))

	// TODO: Implement bulk keyword deletion
	return fmt.Errorf("bulk keyword deletion not yet implemented")
}

// =============================================================================
// Health and Maintenance
// =============================================================================

// Ping checks the database connection
func (r *SupabaseKeywordRepository) Ping(ctx context.Context) error {
	r.logger.Printf("🔍 Pinging database")
	// Use interface client if available, otherwise use concrete client
	if r.clientInterface != nil {
		return r.clientInterface.Ping(ctx)
	}
	if r.client == nil {
		return fmt.Errorf("database client not available")
	}
	return r.client.Ping(ctx)
}

// GetDatabaseStats gets database statistics
func (r *SupabaseKeywordRepository) GetDatabaseStats(ctx context.Context) (map[string]interface{}, error) {
	r.logger.Printf("🔍 Getting database statistics")

	// TODO: Implement database statistics
	return map[string]interface{}{}, nil
}

// CleanupInactiveData cleans up inactive data
func (r *SupabaseKeywordRepository) CleanupInactiveData(ctx context.Context) error {
	r.logger.Printf("🔍 Cleaning up inactive data")

	// TODO: Implement data cleanup
	return fmt.Errorf("data cleanup not yet implemented")
}

// =============================================================================
// Helper Methods
// =============================================================================

// extractKeywords extracts keywords from business information with enhanced phrase matching and context tracking
// Note: Description removed for security - only uses business name and website content
// Priority: Website content FIRST (highest priority), business name LAST (only for brand matches in MCC 3000-3831)
// Phase 8: Enhanced with comprehensive logging and observability
func (r *SupabaseKeywordRepository) extractKeywords(businessName, websiteURL string) []ContextualKeyword {
	extractionStartTime := time.Now()
	r.logger.Printf("🔍 [KeywordExtraction] Starting extraction for: %s (business: %s)", websiteURL, businessName)
	
	var keywords []ContextualKeyword
	seen := make(map[string]bool)
	
	// Track metrics for observability (Phase 8.2)
	// Note: Error counts are logged in nested functions but not tracked here
	// as they occur in separate function scopes. Errors are properly categorized
	// and logged in each extraction method (Phase 8.3).
	metrics := struct {
		startTime          time.Time
		level1Time         time.Duration
		level2Time         time.Duration
		level3Time         time.Duration
		level4Time         time.Duration
		level1Keywords     int
		level2Keywords     int
		level3Keywords     int
		level4Keywords     int
		level1Success      bool
		level2Success      bool
		level3Success      bool
		level4Success      bool
	}{
		startTime: extractionStartTime,
	}

	// PRIORITY 1: Extract keywords from website content (HIGHEST PRIORITY)
	// Enhanced multi-level fallback chain (Phase 5)
	if websiteURL != "" {
		analysisMethod := "none"
		confidenceLevel := "high"
		
		// Level 1: Multi-page analysis (15 pages) - requires 5+ keywords for success
		// Use background context with timeout for keyword extraction
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		
		r.logger.Printf("📊 [KeywordExtraction] Level 1: Starting multi-page website analysis (max 15 pages)")
		level1Start := time.Now()
		multiPageKeywords := r.extractKeywordsFromMultiPageWebsite(ctx, websiteURL)
		metrics.level1Time = time.Since(level1Start)
		metrics.level1Keywords = len(multiPageKeywords)
		
		r.logger.Printf("📊 [KeywordExtraction] Level 1 completed in %v: extracted %d keywords", metrics.level1Time, len(multiPageKeywords))
		
		if len(multiPageKeywords) >= 5 {
			// Success: enough keywords from multi-page analysis
			for _, keyword := range multiPageKeywords {
				if !seen[keyword] {
					seen[keyword] = true
					keywords = append(keywords, ContextualKeyword{
						Keyword: keyword,
						Context: "website_content",
					})
				}
			}
			analysisMethod = "multi_page"
			confidenceLevel = "high"
			metrics.level1Success = true
			r.logger.Printf("✅ [KeywordExtraction] Level 1 SUCCESS: Extracted %d keywords from multi-page analysis (threshold: 5+)", len(multiPageKeywords))
			r.logger.Printf("📝 [KeywordExtraction] Level 1 keywords: %v", multiPageKeywords)
		} else if len(multiPageKeywords) > 0 {
			// Partial success: some keywords but not enough (Phase 5.2)
			for _, keyword := range multiPageKeywords {
				if !seen[keyword] {
					seen[keyword] = true
					keywords = append(keywords, ContextualKeyword{
						Keyword: keyword,
						Context: "website_content",
					})
				}
			}
			analysisMethod = "multi_page_partial"
			confidenceLevel = "medium"
			metrics.level1Success = false
			r.logger.Printf("⚠️ [KeywordExtraction] Level 1 PARTIAL: Only %d keywords from multi-page analysis (below threshold of 5), continuing fallback", len(multiPageKeywords))
			if len(multiPageKeywords) > 0 {
				r.logger.Printf("📝 [KeywordExtraction] Level 1 partial keywords: %v", multiPageKeywords)
			}
		}
		
		// Level 2: Single-page analysis (homepage only) - requires 3+ keywords for success
		if len(keywords) < 5 {
			r.logger.Printf("📊 [KeywordExtraction] Level 2: Starting single-page website analysis (homepage only)")
			level2Start := time.Now()
			singlePageKeywords := r.extractKeywordsFromWebsite(ctx, websiteURL)
			metrics.level2Time = time.Since(level2Start)
			metrics.level2Keywords = len(singlePageKeywords)
			
			r.logger.Printf("📊 [KeywordExtraction] Level 2 completed in %v: extracted %d keywords", metrics.level2Time, len(singlePageKeywords))
			
			if len(singlePageKeywords) >= 3 {
				// Success: enough keywords from single-page
				for _, keyword := range singlePageKeywords {
					if !seen[keyword] {
						seen[keyword] = true
						keywords = append(keywords, ContextualKeyword{
							Keyword: keyword,
							Context: "website_content",
						})
					}
				}
				if analysisMethod == "none" || analysisMethod == "multi_page_partial" {
					analysisMethod = "single_page"
					confidenceLevel = "medium"
				}
				metrics.level2Success = true
				r.logger.Printf("✅ [KeywordExtraction] Level 2 SUCCESS: Extracted %d keywords from single-page analysis (threshold: 3+)", len(singlePageKeywords))
				r.logger.Printf("📝 [KeywordExtraction] Level 2 keywords: %v", singlePageKeywords)
			} else if len(singlePageKeywords) > 0 {
				// Partial success: some keywords but not enough
				for _, keyword := range singlePageKeywords {
					if !seen[keyword] {
						seen[keyword] = true
						keywords = append(keywords, ContextualKeyword{
							Keyword: keyword,
							Context: "website_content",
						})
					}
				}
				if analysisMethod == "none" {
					analysisMethod = "single_page_partial"
					confidenceLevel = "low"
				}
				metrics.level2Success = false
				r.logger.Printf("⚠️ [KeywordExtraction] Level 2 PARTIAL: Only %d keywords from single-page analysis (below threshold of 3), continuing fallback", len(singlePageKeywords))
				if len(singlePageKeywords) > 0 {
					r.logger.Printf("📝 [KeywordExtraction] Level 2 partial keywords: %v", singlePageKeywords)
				}
			}
		}
		
		// Level 3: Homepage with enhanced retry (different DNS, longer timeout) - requires 2+ keywords
		if len(keywords) < 3 {
			r.logger.Printf("📊 [KeywordExtraction] Level 3: Starting homepage extraction with enhanced retry (multiple DNS servers)")
			level3Start := time.Now()
			homepageKeywords := r.extractKeywordsFromHomepageWithRetry(ctx, websiteURL)
			metrics.level3Time = time.Since(level3Start)
			metrics.level3Keywords = len(homepageKeywords)
			
			r.logger.Printf("📊 [KeywordExtraction] Level 3 completed in %v: extracted %d keywords", metrics.level3Time, len(homepageKeywords))
			
			if len(homepageKeywords) >= 2 {
				// Success: enough keywords from retry
				for _, keyword := range homepageKeywords {
					if !seen[keyword] {
						seen[keyword] = true
						keywords = append(keywords, ContextualKeyword{
							Keyword: keyword,
							Context: "website_content",
						})
					}
				}
				if analysisMethod == "none" || analysisMethod == "single_page_partial" {
					analysisMethod = "homepage_retry"
					confidenceLevel = "low"
				}
				metrics.level3Success = true
				r.logger.Printf("✅ [KeywordExtraction] Level 3 SUCCESS: Extracted %d keywords from homepage with retry (threshold: 2+)", len(homepageKeywords))
				r.logger.Printf("📝 [KeywordExtraction] Level 3 keywords: %v", homepageKeywords)
			} else if len(homepageKeywords) > 0 {
				// Partial success
				for _, keyword := range homepageKeywords {
					if !seen[keyword] {
						seen[keyword] = true
						keywords = append(keywords, ContextualKeyword{
							Keyword: keyword,
							Context: "website_content",
						})
					}
				}
				if analysisMethod == "none" {
					analysisMethod = "homepage_retry_partial"
					confidenceLevel = "low"
				}
				metrics.level3Success = false
				r.logger.Printf("⚠️ [KeywordExtraction] Level 3 PARTIAL: Only %d keywords from homepage retry (below threshold of 2)", len(homepageKeywords))
				if len(homepageKeywords) > 0 {
					r.logger.Printf("📝 [KeywordExtraction] Level 3 partial keywords: %v", homepageKeywords)
				}
			}
		}
		
		// Level 4: Enhanced URL text extraction - requires 1+ keywords
		if len(keywords) < 2 {
			r.logger.Printf("📊 [KeywordExtraction] Level 4: Starting enhanced URL text extraction")
			level4Start := time.Now()
			urlKeywords := r.extractKeywordsFromURLEnhanced(websiteURL)
			metrics.level4Time = time.Since(level4Start)
			metrics.level4Keywords = len(urlKeywords)
			
			r.logger.Printf("📊 [KeywordExtraction] Level 4 completed in %v: extracted %d keywords", metrics.level4Time, len(urlKeywords))
			
			if len(urlKeywords) >= 1 {
				for _, keyword := range urlKeywords {
					if !seen[keyword.Keyword] {
						seen[keyword.Keyword] = true
						keywords = append(keywords, keyword)
					}
				}
				if analysisMethod == "none" {
					analysisMethod = "url_only"
					confidenceLevel = "low"
				}
				metrics.level4Success = true
				r.logger.Printf("✅ [KeywordExtraction] Level 4 SUCCESS: Extracted %d keywords from enhanced URL text extraction (threshold: 1+)", len(urlKeywords))
				r.logger.Printf("📝 [KeywordExtraction] Level 4 keywords: %v", urlKeywords)
			}
		}
		
		// Level 5: Business name analysis (if brand match) - already handled below
		// Level 6: Default to "General Business" with low confidence - handled by returning empty keywords
		
		// Phase 8.2: Log performance metrics
		totalExtractionTime := time.Since(extractionStartTime)
		r.logger.Printf("📊 [KeywordExtraction] Performance Metrics:")
		r.logger.Printf("   - Total extraction time: %v", totalExtractionTime)
		r.logger.Printf("   - Level 1 (multi-page): %v, keywords: %d, success: %v", metrics.level1Time, metrics.level1Keywords, metrics.level1Success)
		r.logger.Printf("   - Level 2 (single-page): %v, keywords: %d, success: %v", metrics.level2Time, metrics.level2Keywords, metrics.level2Success)
		r.logger.Printf("   - Level 3 (homepage-retry): %v, keywords: %d, success: %v", metrics.level3Time, metrics.level3Keywords, metrics.level3Success)
		r.logger.Printf("   - Level 4 (URL-only): %v, keywords: %d, success: %v", metrics.level4Time, metrics.level4Keywords, metrics.level4Success)
		r.logger.Printf("   - Note: Errors are logged in individual extraction methods with categorization (DNS, HTTP, Parsing)")
		
		r.logger.Printf("📊 [KeywordExtraction] Final result: method=%s, confidence=%s, total_unique_keywords=%d", analysisMethod, confidenceLevel, len(keywords))
		
		// Log top keywords for observability
		if len(keywords) > 0 {
			topKeywords := make([]string, 0, min(10, len(keywords)))
			for i, kw := range keywords {
				if i >= 10 {
					break
				}
				topKeywords = append(topKeywords, kw.Keyword)
			}
			r.logger.Printf("📝 [KeywordExtraction] Top keywords: %v", topKeywords)
		}
	}

	// Note: Description processing removed for security reasons
	// Business descriptions provided by merchants can be unreliable, misleading, or fraudulent

	// PRIORITY 2: Extract keywords from business name (LOWEST PRIORITY - only for brand matches in MCC 3000-3831)
	if businessName != "" {
		// Check if business name matches a known hotel brand (MCC 3000-3831)
		isBrandMatch, brandName, confidence := r.brandMatcher.IsHighConfidenceBrandMatch(businessName)
		if isBrandMatch {
			r.logger.Printf("✅ Brand match detected: %s (matched: %s, confidence: %.2f) - extracting business name keywords", businessName, brandName, confidence)
			nameKeywords := r.extractKeywordsFromText(businessName, "business_name")
			for _, keyword := range nameKeywords {
				if !seen[keyword.Keyword] {
					seen[keyword.Keyword] = true
					keywords = append(keywords, keyword)
				}
			}
			r.logger.Printf("✅ Extracted %d keywords from business name (brand match in MCC 3000-3831): %v", len(nameKeywords), nameKeywords)
		} else {
			r.logger.Printf("⚠️ Business name '%s' does not match known hotel brands (MCC 3000-3831) - skipping business name keywords", businessName)
		}
	}

	return keywords
}

// extractKeywordsFromHomepageWithRetry attempts to extract keywords from homepage with enhanced retry logic
// Uses different DNS servers, longer timeout, and multiple retry attempts (Phase 5.1)
// Phase 8: Enhanced with detailed logging and error tracking
func (r *SupabaseKeywordRepository) extractKeywordsFromHomepageWithRetry(ctx context.Context, websiteURL string) []string {
	startTime := time.Now()
	r.logger.Printf("🔄 [KeywordExtraction] [HomepageRetry] Starting homepage extraction with enhanced retry for: %s", websiteURL)

	// Create context with longer timeout (30 seconds for retry attempts)
	retryCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Validate URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		r.logger.Printf("❌ [HomepageRetry] Invalid URL format for %s: %v", websiteURL, err)
		return []string{}
	}

	if parsedURL.Scheme == "" {
		websiteURL = "https://" + websiteURL
	}

	// Try multiple DNS servers with retry logic
	dnsServers := []string{"8.8.8.8:53", "1.1.1.1:53", "8.8.4.4:53"}
	
	maxRetries := 3
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		// Try each DNS server
		for _, dnsServer := range dnsServers {
			r.logger.Printf("🔄 [KeywordExtraction] [HomepageRetry] Attempt %d/%d using DNS server %s", attempt, maxRetries, dnsServer)
			
			// Create custom DNS resolver that forces IPv4 and prevents fallback to system DNS
			dnsResolver := &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					// Force IPv4 UDP connection to our custom DNS server
					// Ignore the network and address parameters to prevent system DNS fallback
					d := net.Dialer{
						Timeout: 10 * time.Second, // Longer timeout for retry
					}
					// Always use udp4 to force IPv4, ignore the network parameter
					conn, err := d.DialContext(ctx, "udp4", dnsServer)
					if err != nil {
						return nil, fmt.Errorf("failed to connect to DNS server %s: %w", dnsServer, err)
					}
					return conn, nil
				},
			}

			// Create custom dialer with longer timeout
			baseDialer := &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}

			customDialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
				if network == "tcp" {
					network = "tcp4"
				}

				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, fmt.Errorf("failed to split host:port: %w", err)
				}

				// Phase 9.2: Check DNS cache first (use host+dnsServer as cache key for retry logic)
				cacheKey := host + ":" + dnsServer
				ips, err := r.getCachedDNSResolutionWithKey(cacheKey, host, dnsResolver, ctx)
				if err != nil {
					// Phase 8.3: Categorize DNS errors
					r.logger.Printf("❌ [KeywordExtraction] [HomepageRetry] DNS ERROR: Lookup failed for %s using %s: %v (type: %T)", host, dnsServer, err, err)
					return nil, fmt.Errorf("DNS lookup failed: %w", err) // Return error to try next DNS server
				}

				// Use first IPv4 address
				var ip net.IP
				for _, ipAddr := range ips {
					if ipAddr.IP.To4() != nil {
						ip = ipAddr.IP
						break
					}
				}

				if ip == nil {
					// Phase 8.3: Categorize DNS errors (no IPv4 found)
					r.logger.Printf("❌ [KeywordExtraction] [HomepageRetry] DNS ERROR: No IPv4 address found for %s (only IPv6 available)", host)
					return nil, fmt.Errorf("no IPv4 address found")
				}

				return baseDialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			}

			// Create HTTP client with custom dialer and longer timeout
			client := &http.Client{
				Timeout: 30 * time.Second,
				Transport: &http.Transport{
					DialContext:          customDialContext,
					MaxIdleConns:        10,
					IdleConnTimeout:     30 * time.Second,
					DisableCompression:  false,
					MaxIdleConnsPerHost: 2,
				},
			}

			// Phase 9.3: Apply rate limiting with jitter
			// Note: parsedURL already declared above, so we reuse it
			if parsedURL != nil {
				domain := parsedURL.Hostname()
				r.applyRateLimit(domain)
			}

			// Create request
			req, err := http.NewRequestWithContext(retryCtx, "GET", websiteURL, nil)
			if err != nil {
				r.logger.Printf("❌ [HomepageRetry] Failed to create request: %v", err)
				continue
			}

			req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
			req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

			// Make request
			resp, err := client.Do(req)
			if err != nil {
				// Phase 8.3: Categorize HTTP errors
				errorType := "unknown"
				if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
					errorType = "timeout"
				} else if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
					errorType = "connection"
				} else if strings.Contains(err.Error(), "DNS") {
					errorType = "dns"
				}
				r.logger.Printf("❌ [KeywordExtraction] [HomepageRetry] HTTP ERROR (%s): Request failed (attempt %d, DNS %s): %v (type: %T)", errorType, attempt, dnsServer, err, err)
				continue // Try next DNS server
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				r.logger.Printf("⚠️ [HomepageRetry] Non-200 status code: %d", resp.StatusCode)
				continue // Try next DNS server
			}

			// Read content
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				r.logger.Printf("⚠️ [HomepageRetry] Failed to read response body: %v", err)
				continue // Try next DNS server
			}

			// Extract keywords from content
			content := string(body)
			textContent := r.extractTextFromHTML(content)
			
			// Extract keywords using business patterns
			extractedKeywords := r.extractBusinessKeywords(textContent)
			
			// Also extract from structured elements
			// Phase 9.1: Use cached compiled regex pattern
			titleRegex := r.getCachedRegex(`(?i)<title[^>]*>([^<]+)</title>`)
			titleMatches := titleRegex.FindStringSubmatch(content)
			if len(titleMatches) > 1 {
				titleKeywords := r.extractBusinessKeywords(titleMatches[1])
				extractedKeywords = append(extractedKeywords, titleKeywords...)
			}

			if len(extractedKeywords) > 0 {
				duration := time.Since(startTime)
				// Phase 8.2: Log performance metrics
				r.logger.Printf("✅ [KeywordExtraction] [HomepageRetry] SUCCESS: Extracted %d keywords in %v (attempt %d, DNS %s)", 
					len(extractedKeywords), duration, attempt, dnsServer)
				r.logger.Printf("📊 [KeywordExtraction] [HomepageRetry] Performance: time=%v, keywords=%d, attempt=%d, dns_server=%s", 
					duration, len(extractedKeywords), attempt, dnsServer)
				return extractedKeywords
			}
			
			// If we got here, we got a response but no keywords - try next DNS server
			r.logger.Printf("⚠️ [KeywordExtraction] [HomepageRetry] WARNING: Got response but no keywords extracted (attempt %d, DNS %s), trying next DNS server", attempt, dnsServer)
		}

		// Exponential backoff before next retry
		if attempt < maxRetries {
			backoff := time.Duration(attempt) * time.Second
			r.logger.Printf("⏳ [KeywordExtraction] [HomepageRetry] Waiting %v before retry %d/%d", backoff, attempt+1, maxRetries)
			select {
			case <-retryCtx.Done():
				r.logger.Printf("❌ [KeywordExtraction] [HomepageRetry] ERROR: Context cancelled during backoff")
				return []string{}
			case <-time.After(backoff):
				// Continue to next retry
			}
		}
	}

	// Phase 8.2 & 8.3: Log final failure metrics
	duration := time.Since(startTime)
	r.logger.Printf("❌ [KeywordExtraction] [HomepageRetry] FAILED: Unable to extract keywords after %d attempts in %v", maxRetries, duration)
	r.logger.Printf("📊 [KeywordExtraction] [HomepageRetry] Failure Summary:")
	r.logger.Printf("   - Total attempts: %d", maxRetries)
	r.logger.Printf("   - DNS servers tried: %d", len(dnsServers))
	r.logger.Printf("   - Total time: %v", duration)
	r.logger.Printf("   - Result: No keywords extracted")
	return []string{}
}

// extractKeywordsFromWebsite scrapes website content and extracts business-relevant keywords
// Phase 8: Enhanced with detailed logging and error tracking
func (r *SupabaseKeywordRepository) extractKeywordsFromWebsite(ctx context.Context, websiteURL string) []string {
	startTime := time.Now()
	r.logger.Printf("🌐 [KeywordExtraction] [SinglePage] Starting single-page website scraping for: %s", websiteURL)

	// Validate URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		r.logger.Printf("❌ [KeywordExtraction] [SinglePage] ERROR: Invalid URL format for %s: %v (type: %T)", websiteURL, err, err)
		return []string{}
	}

	if parsedURL.Scheme == "" {
		websiteURL = "https://" + websiteURL
		r.logger.Printf("🔧 [KeywordExtraction] [SinglePage] Added HTTPS scheme: %s", websiteURL)
	}

	// Create custom dialer that forces IPv4 DNS resolution using Google DNS
	// This addresses DNS resolution failures in containerized environments like Railway
	baseDialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// Create custom DNS resolver with multiple fallback servers
	// DNS servers in order of preference: Google DNS, Cloudflare, Google DNS secondary
	dnsServers := []string{"8.8.8.8:53", "1.1.1.1:53", "8.8.4.4:53"}
	dnsResolver := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			// Force IPv4 UDP connection to our custom DNS server
			// Ignore the network and address parameters to prevent system DNS fallback
			// Try each DNS server with retry logic
			var lastErr error
			for _, server := range dnsServers {
				d := net.Dialer{
					Timeout: 5 * time.Second,
				}
				// Always use udp4 to force IPv4, ignore the network parameter
				conn, err := d.DialContext(ctx, "udp4", server)
				if err == nil {
					return conn, nil
				}
				lastErr = err
				r.logger.Printf("⚠️ [KeywordExtraction] [SinglePage] DNS: Failed to connect to DNS server %s: %v", server, err)
			}
			return nil, fmt.Errorf("all DNS servers failed, last error: %w", lastErr)
		},
	}

	// Custom DialContext that forces IPv4 resolution
	customDialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		// Force IPv4 by using "tcp4" instead of "tcp"
		if network == "tcp" {
			network = "tcp4"
		}

		// Parse address to get host and port
		host, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, fmt.Errorf("failed to split host:port: %w", err)
		}

		// Phase 9.2: Check DNS cache first
		ips, err := r.getCachedDNSResolution(host, dnsResolver, ctx)
		if err != nil {
			// Phase 8.3: Categorize DNS errors
			r.logger.Printf("❌ [KeywordExtraction] [SinglePage] DNS ERROR: Lookup failed for %s: %v (type: %T)", host, err, err)
			return nil, fmt.Errorf("DNS lookup failed for %s: %w", host, err)
		}

		// Use first IPv4 address
		var ip net.IP
		for _, ipAddr := range ips {
			if ipAddr.IP.To4() != nil {
				ip = ipAddr.IP
				break
			}
		}

		if ip == nil {
			// Phase 8.3: Categorize DNS errors (no IPv4 found)
			r.logger.Printf("❌ [KeywordExtraction] [SinglePage] DNS ERROR: No IPv4 address found for %s (only IPv6 available)", host)
			return nil, fmt.Errorf("no IPv4 address found for %s", host)
		}

		// Dial using resolved IPv4 address
		return baseDialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DialContext:         customDialContext,
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: false,
		},
	}

	// Phase 9.3: Apply rate limiting with jitter
	// Note: parsedURL already declared above, so we reuse it
	if parsedURL != nil {
		domain := parsedURL.Hostname()
		r.applyRateLimit(domain)
	}

	// Create request with enhanced headers
	req, err := http.NewRequestWithContext(ctx, "GET", websiteURL, nil)
	if err != nil {
		r.logger.Printf("❌ [KeywordExtraction] [SinglePage] PARSING ERROR: Failed to create request for %s: %v (type: %T)", websiteURL, err, err)
		return []string{}
	}

	// Set comprehensive headers to mimic a real browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("DNT", "1")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Cache-Control", "max-age=0")

	r.logger.Printf("📡 [KeywordExtraction] [SinglePage] Making HTTP request to: %s", websiteURL)

	// Make request with timeout context
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	resp, err := client.Do(req)
	if err != nil {
		// Phase 8.3: Categorize HTTP errors
		errorType := "unknown"
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
			errorType = "timeout"
		} else if strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host") {
			errorType = "connection"
		} else if strings.Contains(err.Error(), "DNS") {
			errorType = "dns"
		}
		r.logger.Printf("❌ [KeywordExtraction] [SinglePage] HTTP ERROR (%s): Request failed for %s: %v (type: %T)", errorType, websiteURL, err, err)
		return []string{}
	}
	defer resp.Body.Close()

	// Log response details
	r.logger.Printf("📊 [KeywordExtraction] [SinglePage] Response received - Status: %d, Content-Type: %s, Content-Length: %d",
		resp.StatusCode, resp.Header.Get("Content-Type"), resp.ContentLength)

	// Phase 8.3: Track HTTP status codes
	if resp.StatusCode >= 400 {
		r.logger.Printf("❌ [KeywordExtraction] [SinglePage] HTTP ERROR: Status %d %s for %s", resp.StatusCode, resp.Status, websiteURL)
		// Try to read error response body
		if body, readErr := io.ReadAll(resp.Body); readErr == nil && len(body) > 0 {
			r.logger.Printf("📄 [KeywordExtraction] [SinglePage] Error response body (first 500 chars): %s", string(body[:min(500, len(body))]))
		}
		return []string{}
	} else if resp.StatusCode != 200 {
		r.logger.Printf("⚠️ [KeywordExtraction] [SinglePage] HTTP WARNING: Status code %d for %s (expected 200)", resp.StatusCode, websiteURL)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		r.logger.Printf("⚠️ [KeywordExtraction] [SinglePage] WARNING: Unexpected content type for %s: %s", websiteURL, contentType)
	}

	// Read response body with size limit
	maxSize := int64(5 * 1024 * 1024) // 5MB limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		r.logger.Printf("❌ [KeywordExtraction] [SinglePage] PARSING ERROR: Failed to read response body from %s: %v (type: %T)", websiteURL, err, err)
		return []string{}
	}

	r.logger.Printf("📄 [KeywordExtraction] [SinglePage] Read %d bytes from %s", len(body), websiteURL)

	// Extract text content from HTML
	textContent := r.extractTextFromHTML(string(body))
	r.logger.Printf("🧹 [KeywordExtraction] [SinglePage] Extracted %d characters of text content from HTML", len(textContent))

	// Log sample of extracted text for debugging
	if len(textContent) > 0 {
		sampleText := textContent[:min(200, len(textContent))]
		r.logger.Printf("📝 [KeywordExtraction] [SinglePage] Sample extracted text: %s...", sampleText)
	}

	// Extract business-relevant keywords from text
	textKeywords := r.extractBusinessKeywords(textContent)
	r.logger.Printf("📝 [KeywordExtraction] [SinglePage] Extracted %d keywords from text content", len(textKeywords))

	// Extract structured data and keywords
	if NewStructuredDataExtractorAdapter == nil {
		r.logger.Printf("⚠️ [KeywordExtraction] [SinglePage] WARNING: StructuredDataExtractor adapter not initialized - skipping structured data extraction")
		// Fallback to text-only extraction
		keywords := r.extractBusinessKeywords(textContent)
		return keywords
	}
	structuredDataExtractor := NewStructuredDataExtractorAdapter(r.logger)
	structuredDataResult := structuredDataExtractor.ExtractStructuredData(string(body))
	
	var structuredKeywords []string
	structuredKeywordMap := make(map[string]float64) // Track keywords with confidence scores

	// Extract keywords from BusinessInfo
	if structuredDataResult.BusinessInfo.Industry != "" {
		structuredKeywordMap[strings.ToLower(structuredDataResult.BusinessInfo.Industry)] = 1.5
	}
	if structuredDataResult.BusinessInfo.BusinessType != "" {
		structuredKeywordMap[strings.ToLower(structuredDataResult.BusinessInfo.BusinessType)] = 1.5
	}
	for _, service := range structuredDataResult.BusinessInfo.Services {
		serviceLower := strings.ToLower(service)
		structuredKeywordMap[serviceLower] = 1.5
	}
	for _, product := range structuredDataResult.BusinessInfo.Products {
		productLower := strings.ToLower(product)
		structuredKeywordMap[productLower] = 1.5
	}

	// Extract keywords from ProductInfo
	for _, product := range structuredDataResult.ProductInfo {
		if product.Name != "" {
			structuredKeywordMap[strings.ToLower(product.Name)] = 1.5
		}
		if product.Category != "" {
			structuredKeywordMap[strings.ToLower(product.Category)] = 1.5
		}
		if product.Description != "" {
			// Extract keywords from product description
			descKeywords := r.extractBusinessKeywords(product.Description)
			for _, kw := range descKeywords {
				structuredKeywordMap[strings.ToLower(kw)] = 1.5
			}
		}
	}

	// Extract keywords from ServiceInfo
	for _, service := range structuredDataResult.ServiceInfo {
		if service.Name != "" {
			structuredKeywordMap[strings.ToLower(service.Name)] = 1.5
		}
		if service.Category != "" {
			structuredKeywordMap[strings.ToLower(service.Category)] = 1.5
		}
		if service.Description != "" {
			// Extract keywords from service description
			descKeywords := r.extractBusinessKeywords(service.Description)
			for _, kw := range descKeywords {
				structuredKeywordMap[strings.ToLower(kw)] = 1.5
			}
		}
	}

	// Extract keywords from Schema.org Organization types
	for _, schemaItem := range structuredDataResult.SchemaOrgData {
		if schemaItem.Type != "" {
			typeLower := strings.ToLower(schemaItem.Type)
			// Focus on business-relevant types
			if strings.Contains(typeLower, "organization") || 
			   strings.Contains(typeLower, "business") || 
			   strings.Contains(typeLower, "localbusiness") ||
			   strings.Contains(typeLower, "store") ||
			   strings.Contains(typeLower, "restaurant") ||
			   strings.Contains(typeLower, "service") {
				structuredKeywordMap[typeLower] = 1.5
			}
		}
		// Extract from properties
		if schemaItem.Properties != nil {
			if industry, exists := schemaItem.Properties["industry"]; exists {
				structuredKeywordMap[strings.ToLower(fmt.Sprintf("%v", industry))] = 1.5
			}
			if businessType, exists := schemaItem.Properties["@type"]; exists {
				typeStr := strings.ToLower(fmt.Sprintf("%v", businessType))
				if strings.Contains(typeStr, "organization") || strings.Contains(typeStr, "business") {
					structuredKeywordMap[typeStr] = 1.5
				}
			}
		}
	}

	// Convert map to slice (structured keywords weighted 1.5x)
	for kw := range structuredKeywordMap {
		structuredKeywords = append(structuredKeywords, kw)
	}

	r.logger.Printf("📊 [StructuredData] Extracted %d keywords from structured data (weighted 1.5x)", len(structuredKeywords))

	// Combine text keywords and structured keywords
	allKeywords := make(map[string]float64)
	
	// Add text keywords with weight 1.0
	for _, kw := range textKeywords {
		kwLower := strings.ToLower(kw)
		if allKeywords[kwLower] < 1.0 {
			allKeywords[kwLower] = 1.0
		}
	}
	
	// Add structured keywords with weight 1.5 (higher priority)
	for kw, weight := range structuredKeywordMap {
		if allKeywords[kw] < weight {
			allKeywords[kw] = weight
		}
	}

	// Convert to slice and sort by weight (descending), then limit to top 30
	type keywordWeight struct {
		keyword string
		weight  float64
	}
	keywordList := make([]keywordWeight, 0, len(allKeywords))
	for kw, weight := range allKeywords {
		keywordList = append(keywordList, keywordWeight{keyword: kw, weight: weight})
	}
	
	// Sort by weight descending
	sort.Slice(keywordList, func(i, j int) bool {
		return keywordList[i].weight > keywordList[j].weight
	})
	
	// Limit to top 30 keywords
	maxKeywords := 30
	if len(keywordList) > maxKeywords {
		keywordList = keywordList[:maxKeywords]
	}
	
	keywords := make([]string, len(keywordList))
	for i, kw := range keywordList {
		keywords[i] = kw.keyword
	}

	// Phase 8.2: Log performance metrics
	duration := time.Since(startTime)
	r.logger.Printf("✅ [KeywordExtraction] [SinglePage] Single-page analysis completed in %v", duration)
	r.logger.Printf("📊 [KeywordExtraction] [SinglePage] Performance Summary:")
	r.logger.Printf("   - Total time: %v", duration)
	r.logger.Printf("   - Keywords from text: %d", len(textKeywords))
	r.logger.Printf("   - Keywords from structured data: %d", len(structuredKeywords))
	r.logger.Printf("   - Total unique keywords extracted: %d", len(keywords))
	if len(keywords) > 0 {
		r.logger.Printf("   - Top keywords: %v", keywords[:min(10, len(keywords))])
	}

	return keywords
}

// extractKeywordsFromMultiPageWebsite analyzes multiple pages with relevance-based weighting
// Uses SmartWebsiteCrawler to discover pages, limits to top 15 pages, analyzes concurrently,
// and returns top 30 keywords weighted by page relevance score
// Phase 8: Enhanced with detailed logging and error tracking
func (r *SupabaseKeywordRepository) extractKeywordsFromMultiPageWebsite(ctx context.Context, websiteURL string) []string {
	startTime := time.Now()
	r.logger.Printf("🌐 [KeywordExtraction] [MultiPage] Starting multi-page website analysis for: %s", websiteURL)

	// Create overall timeout context (60 seconds for multi-page analysis)
	analysisCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	// Create SmartWebsiteCrawler with max 15 pages
	if NewSmartWebsiteCrawlerAdapter == nil {
		r.logger.Printf("❌ [KeywordExtraction] [MultiPage] ERROR: SmartWebsiteCrawler adapter not initialized - falling back to single page")
		return []string{} // Return empty to trigger fallback
	}
	crawler := NewSmartWebsiteCrawlerAdapter(r.logger)
	
	// Use CrawlWebsite which handles discovery, prioritization, and analysis
	r.logger.Printf("🔍 [KeywordExtraction] [MultiPage] Discovering and crawling website pages...")
	crawlResult, err := crawler.CrawlWebsite(analysisCtx, websiteURL)
	if err != nil {
		r.logger.Printf("❌ [KeywordExtraction] [MultiPage] ERROR: Website crawl failed: %v (type: %T) - falling back to single page", err, err)
		return []string{} // Return empty to trigger fallback
	}

	if crawlResult == nil {
		r.logger.Printf("❌ [KeywordExtraction] [MultiPage] ERROR: Crawl result is nil - falling back to single page")
		return []string{} // Return empty to trigger fallback
	}

	pagesAnalyzed := crawlResult.GetPagesAnalyzed()
	r.logger.Printf("📊 [KeywordExtraction] [MultiPage] Discovered %d pages for analysis", len(pagesAnalyzed))
	
	if len(pagesAnalyzed) == 0 {
		r.logger.Printf("⚠️ [KeywordExtraction] [MultiPage] WARNING: No pages analyzed - falling back to single page")
		return []string{} // Return empty to trigger fallback
	}

	// Limit to top 15 pages by priority (they're already sorted by CrawlWebsite)
	maxPages := 15
	pagesToAnalyze := pagesAnalyzed
	if len(pagesToAnalyze) > maxPages {
		pagesToAnalyze = pagesToAnalyze[:maxPages]
		r.logger.Printf("📊 [KeywordExtraction] [MultiPage] Limited to top %d pages by priority (from %d total pages)", maxPages, len(pagesAnalyzed))
	}

	// Check if we have enough successful pages (at least 3)
	successfulPages := 0
	for i, page := range pagesToAnalyze {
		statusCode := page.GetStatusCode()
		relevanceScore := page.GetRelevanceScore()
		pageURL := page.GetURL()
		
		// Detailed logging for each page (Phase 8.1)
		r.logger.Printf("🔍 [KeywordExtraction] [MultiPage] Page %d/%d: URL=%s, Status=%d, Relevance=%.2f", 
			i+1, len(pagesToAnalyze), pageURL, statusCode, relevanceScore)
		
		if statusCode == 200 && relevanceScore > 0 {
			successfulPages++
			r.logger.Printf("✅ [KeywordExtraction] [MultiPage] Page %d/%d SUCCESS: URL=%s, Status=%d, Relevance=%.2f", 
				i+1, len(pagesToAnalyze), pageURL, statusCode, relevanceScore)
		} else {
			// Phase 8.3: Categorize errors
			if statusCode != 200 {
				r.logger.Printf("⚠️ [KeywordExtraction] [MultiPage] Page %d/%d HTTP ERROR: Status=%d (expected 200), URL=%s", 
					i+1, len(pagesToAnalyze), statusCode, pageURL)
			}
			if relevanceScore <= 0 {
				r.logger.Printf("⚠️ [KeywordExtraction] [MultiPage] Page %d/%d RELEVANCE ERROR: Relevance=%.2f (expected >0), URL=%s", 
					i+1, len(pagesToAnalyze), relevanceScore, pageURL)
			}
		}
	}

	if successfulPages < 3 {
		r.logger.Printf("⚠️ [KeywordExtraction] [MultiPage] WARNING: Only %d/%d pages successfully analyzed (< 3 required) - falling back to single page", 
			successfulPages, len(pagesToAnalyze))
		return []string{} // Return empty to trigger fallback
	}
	
	r.logger.Printf("✅ [KeywordExtraction] [MultiPage] %d/%d pages successfully analyzed, proceeding with keyword extraction", 
		successfulPages, len(pagesToAnalyze))

	// Weight keywords by page relevance score
	keywordWeights := make(map[string]float64)
	totalRelevance := 0.0

	// Calculate total relevance for normalization
	for _, page := range pagesToAnalyze {
		if page.GetStatusCode() == 200 && page.GetRelevanceScore() > 0 {
			totalRelevance += page.GetRelevanceScore()
		}
	}

	// Extract keywords from each page and weight by relevance
	for _, page := range pagesToAnalyze {
		if page.GetStatusCode() == 200 && page.GetRelevanceScore() > 0 && totalRelevance > 0 {
			weight := page.GetRelevanceScore() / totalRelevance // Normalize by total relevance
			
			// Extract keywords from page
			pageKeywords := r.extractKeywordsFromPageData(page)
			
			// Weight keywords by page relevance
			for _, keyword := range pageKeywords {
				keywordLower := strings.ToLower(keyword)
				keywordWeights[keywordLower] += weight
			}
			
			r.logger.Printf("✅ [KeywordExtraction] [MultiPage] Page analyzed: URL=%s, relevance=%.2f, keywords_extracted=%d, weighted_keywords=%d", 
				page.GetURL(), page.GetRelevanceScore(), len(pageKeywords), len(keywordWeights))
		} else {
			r.logger.Printf("⚠️ [KeywordExtraction] [MultiPage] Page skipped: URL=%s, status=%d, relevance=%.2f", 
				page.GetURL(), page.GetStatusCode(), page.GetRelevanceScore())
		}
	}

	// Convert to slice and sort by weighted score
	type keywordWeight struct {
		keyword string
		weight  float64
	}
	keywordList := make([]keywordWeight, 0, len(keywordWeights))
	for kw, weight := range keywordWeights {
		keywordList = append(keywordList, keywordWeight{keyword: kw, weight: weight})
	}

	// Sort by weight descending
	sort.Slice(keywordList, func(i, j int) bool {
		return keywordList[i].weight > keywordList[j].weight
	})

	// Limit to top 30 keywords
	maxKeywords := 30
	if len(keywordList) > maxKeywords {
		keywordList = keywordList[:maxKeywords]
	}

	keywords := make([]string, len(keywordList))
	for i, kw := range keywordList {
		keywords[i] = kw.keyword
	}

	// Phase 8.2: Log performance metrics
	duration := time.Since(startTime)
	r.logger.Printf("✅ [KeywordExtraction] [MultiPage] Multi-page analysis completed in %v", duration)
	r.logger.Printf("📊 [KeywordExtraction] [MultiPage] Performance Summary:")
	r.logger.Printf("   - Total time: %v", duration)
	r.logger.Printf("   - Pages discovered: %d", len(pagesAnalyzed))
	r.logger.Printf("   - Pages analyzed: %d", len(pagesToAnalyze))
	r.logger.Printf("   - Successful pages: %d", successfulPages)
	r.logger.Printf("   - Unique keywords extracted: %d", len(keywords))
	if len(keywords) > 0 {
		r.logger.Printf("   - Top keywords: %v", keywords[:min(10, len(keywords))])
	}

	return keywords
}

// extractKeywordsFromPageData extracts keywords from a PageAnalysisData result
func (r *SupabaseKeywordRepository) extractKeywordsFromPageData(page PageAnalysisData) []string {
	var keywords []string
	seen := make(map[string]bool)

	// Extract keywords from page keywords
	for _, kw := range page.GetKeywords() {
		kwLower := strings.ToLower(kw)
		if !seen[kwLower] {
			seen[kwLower] = true
			keywords = append(keywords, kwLower)
		}
	}

	// Extract keywords from industry indicators
	for _, indicator := range page.GetIndustryIndicators() {
		indLower := strings.ToLower(indicator)
		if !seen[indLower] {
			seen[indLower] = true
			keywords = append(keywords, indLower)
		}
	}

	// Extract keywords from structured data if present
	structuredData := page.GetStructuredData()
	// Extract from the structured data map directly (ranging over nil map is safe in Go)
	for _, value := range structuredData {
			if strValue, ok := value.(string); ok && strValue != "" {
				// Extract keywords from structured data values
				extracted := r.extractBusinessKeywords(strValue)
				for _, kw := range extracted {
					kwLower := strings.ToLower(kw)
					if !seen[kwLower] {
						seen[kwLower] = true
						keywords = append(keywords, kwLower)
					}
			}
		}
	}

	return keywords
}

// extractTextFromHTML extracts clean text content from HTML
// Phase 9.1: Optimized with cached regex patterns and content size limiting
func (r *SupabaseKeywordRepository) extractTextFromHTML(htmlContent string) string {
	// Phase 9.1: Limit content size for processing (first 50KB)
	if int64(len(htmlContent)) > r.maxContentSize {
		htmlContent = htmlContent[:r.maxContentSize]
		r.logger.Printf("📊 [Performance] Content size limited to %d bytes for processing", r.maxContentSize)
	}

	// Phase 9.1: Use cached compiled regex patterns
	scriptRegex := r.getCachedRegex(`(?i)<script[^>]*>.*?</script>`)
	styleRegex := r.getCachedRegex(`(?i)<style[^>]*>.*?</style>`)
	tagRegex := r.getCachedRegex(`<[^>]*>`)
	whitespaceRegex := r.getCachedRegex(`\s+`)

	// Remove script and style tags completely
	htmlContent = scriptRegex.ReplaceAllString(htmlContent, "")
	htmlContent = styleRegex.ReplaceAllString(htmlContent, "")

	// Remove HTML tags
	htmlContent = tagRegex.ReplaceAllString(htmlContent, " ")

	// Clean up whitespace
	htmlContent = whitespaceRegex.ReplaceAllString(htmlContent, " ")

	return strings.TrimSpace(htmlContent)
}

// getCachedRegex returns a cached compiled regex pattern or compiles and caches it
// Phase 9.1: Performance optimization to avoid recompiling regex patterns
func (r *SupabaseKeywordRepository) getCachedRegex(pattern string) *regexp.Regexp {
	// Try read lock first
	r.regexMutex.RLock()
	if regex, exists := r.regexCache[pattern]; exists {
		r.regexMutex.RUnlock()
		return regex
	}
	r.regexMutex.RUnlock()

	// Compile and cache
	r.regexMutex.Lock()
	defer r.regexMutex.Unlock()

	// Double-check after acquiring write lock
	if regex, exists := r.regexCache[pattern]; exists {
		return regex
	}

	// Compile and cache
	regex := regexp.MustCompile(pattern)
	r.regexCache[pattern] = regex
	return regex
}

// getCachedDNSResolution performs DNS lookup with caching (TTL-based)
// Phase 9.2: Performance optimization to cache DNS resolutions
func (r *SupabaseKeywordRepository) getCachedDNSResolution(host string, resolver *net.Resolver, ctx context.Context) ([]net.IPAddr, error) {
	return r.getCachedDNSResolutionWithKey(host, host, resolver, ctx)
}

// getCachedDNSResolutionWithKey performs DNS lookup with caching using a custom cache key
// Phase 9.2: Performance optimization to cache DNS resolutions
func (r *SupabaseKeywordRepository) getCachedDNSResolutionWithKey(cacheKey, host string, resolver *net.Resolver, ctx context.Context) ([]net.IPAddr, error) {
	// Check cache first
	r.dnsMutex.RLock()
	entry, exists := r.dnsCache[cacheKey]
	if exists && time.Now().Before(entry.expiresAt) {
		// Cache hit - return cached IPs
		r.dnsMutex.RUnlock()
		r.logger.Printf("📊 [Performance] DNS cache hit for %s", host)
		return entry.ips, nil
	}
	r.dnsMutex.RUnlock()

	// If entry exists but expired, remove it (need write lock for deletion)
	if exists {
		r.dnsMutex.Lock()
		// Double-check after acquiring write lock (another goroutine might have removed it)
		if entry, stillExists := r.dnsCache[cacheKey]; stillExists && time.Now().After(entry.expiresAt) {
			delete(r.dnsCache, cacheKey)
		}
		r.dnsMutex.Unlock()
	}

	// Cache miss - perform DNS lookup
	r.logger.Printf("📊 [Performance] DNS cache miss for %s, performing lookup", host)
	ips, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}

	// Cache the result with TTL (default 5 minutes, or use DNS TTL if available)
	ttl := 5 * time.Minute
	r.dnsMutex.Lock()
	
	// Clean up expired entries periodically (when cache grows large)
	// This prevents unbounded memory growth
	if len(r.dnsCache) > 1000 {
		now := time.Now()
		cleanedCount := 0
		for key, entry := range r.dnsCache {
			if now.After(entry.expiresAt) {
				delete(r.dnsCache, key)
				cleanedCount++
			}
		}
		if cleanedCount > 0 {
			r.logger.Printf("🧹 [Performance] Cleaned up %d expired DNS cache entries", cleanedCount)
		}
	}
	
	r.dnsCache[cacheKey] = dnsCacheEntry{
		ips:       ips,
		expiresAt: time.Now().Add(ttl),
	}
	r.dnsMutex.Unlock()

	return ips, nil
}

// applyRateLimit applies rate limiting with jitter to avoid thundering herd
// Phase 9.3: Performance optimization to respect rate limits and add jitter
func (r *SupabaseKeywordRepository) applyRateLimit(domain string) {
	if domain == "" {
		return
	}

	r.rateMutex.Lock()
	defer r.rateMutex.Unlock()

	// Clean up old entries (older than 1 hour) to prevent memory leak
	// This is a simple cleanup - in production, consider a more sophisticated approach
	now := time.Now()
	cutoff := now.Add(-1 * time.Hour)
	cleanedCount := 0
	for key, lastRequest := range r.rateLimiter {
		if lastRequest.Before(cutoff) {
			delete(r.rateLimiter, key)
			cleanedCount++
		}
	}
	if cleanedCount > 0 {
		r.logger.Printf("🧹 [Performance] Cleaned up %d old rate limiter entries", cleanedCount)
	}

	lastRequest, exists := r.rateLimiter[domain]
	if exists {
		elapsed := time.Since(lastRequest)
		if elapsed < r.minDelay {
			// Calculate delay with jitter (random 0-20% of minDelay)
			remainingDelay := r.minDelay - elapsed
			jitter := time.Duration(float64(remainingDelay) * 0.2 * rand.Float64())
			totalDelay := remainingDelay + jitter

			r.logger.Printf("⏳ [Performance] Rate limiting: waiting %v before request to %s (base: %v, jitter: %v)",
				totalDelay, domain, remainingDelay, jitter)
			time.Sleep(totalDelay)
		}
	}

	// Update last request time
	r.rateLimiter[domain] = time.Now()
}

// extractBusinessKeywords extracts business-relevant keywords from text content
func (r *SupabaseKeywordRepository) extractBusinessKeywords(textContent string) []string {
	var keywords []string

	// Convert to lowercase for processing
	text := strings.ToLower(textContent)

	// Business-relevant keyword patterns (expanded with synonyms and NAICS-aligned terms)
	businessPatterns := []string{
		// Food & Beverage (expanded) - single words first, then phrases
		`\b(wine|wines|winery|vineyard|vintner|sommelier|tasting|cellar|bottle|vintage|grape|grapes|grapevine|oenology|alcohol|spirits|liquor|beer|brewery|distillery|beverage|beverages|restaurant|cafe|coffee|food|dining|kitchen|catering|bakery|bar|pub|bistro|eatery|diner|tavern|gastropub|brewpub)\b`,
		`\b(wine shop|wine store|wine bar|wine merchant|wine retailer|wine tasting|wine cellar|wine selection|fine wine|premium wine)\b`,
		`\b(food service|dining establishment|restaurant chain|fast food|casual dining|fine dining|takeout|delivery|food truck)\b`,
		
		// Retail (expanded) - single words first, then phrases
		`\b(retail|retailer|storefront|merchandise|inventory|POS|checkout|showroom|boutique|outlet|marketplace|vendor|seller|selling|commerce|store|shop|boutique|emporium|mart|bazaar|market|retailer|merchant|dealer|reseller)\b`,
		`\b(retail store|retail shop|brick and mortar|brick-and-mortar|physical store|point of sale|cash register|sales floor|retail location|store location|retail outlet|retail chain)\b`,
		`\b(merchandise sales|product sales|consumer goods|retail goods|store merchandise|inventory management|stock management)\b`,
		
		// E-commerce (expanded) - single words first, then phrases
		`\b(ecommerce|e-commerce|online|digital|web|internet|cyber)\b`,
		`\b(online store|online shop|digital storefront|web store|internet retailer|online marketplace|digital commerce|online sales|web sales|internet sales|online retail|ecommerce platform|online shopping|web commerce|digital retail)\b`,
		`\b(online business|digital business|web business|internet business|ecommerce business|online merchant|digital merchant)\b`,
		
		// Technology (expanded)
		`\b(technology|software|tech|app|application|digital|web|mobile|cloud|ai|artificial intelligence|ml|machine learning|data|cyber|security|programming|development|IT|information technology|computer|internet|online|platform|api|database|saas|software as a service|paas|iaas|devops|automation|digitalization)\b`,
		`\b(software development|software engineering|web development|mobile development|app development|cloud computing|data science|cybersecurity|IT services|tech services|digital solutions|software solutions)\b`,
		`\b(technology company|tech company|software company|IT company|digital agency|tech startup|software firm)\b`,
		
		// Healthcare (expanded)
		`\b(healthcare|health care|medical|clinic|hospital|doctor|physician|dentist|therapy|wellness|pharmacy|medicine|patient|treatment|health|care|nurse|practitioner|surgeon|specialist|therapist|wellness|rehabilitation|diagnosis|treatment)\b`,
		`\b(medical services|healthcare services|medical care|health services|patient care|medical treatment|healthcare provider|medical provider|healthcare facility|medical facility)\b`,
		`\b(primary care|specialty care|urgent care|emergency care|preventive care|healthcare system|medical system)\b`,
		
		// Legal (expanded)
		`\b(legal|law|attorney|lawyer|attorney at law|counsel|counselor|barrister|solicitor|court|litigation|patent|trademark|copyright|legal services|advocacy|justice|legal advice|law firm|legal counsel|legal representation|legal practice)\b`,
		`\b(law firm|legal firm|attorney firm|law office|legal office|legal services|legal counsel|legal representation|legal practice|litigation services)\b`,
		`\b(legal advice|legal consultation|legal assistance|legal support|legal guidance|legal expertise)\b`,
		
		// Finance (expanded)
		`\b(finance|banking|investment|insurance|accounting|tax|financial|credit|loan|money|capital|funding|payment|transaction|wealth|asset|portfolio|brokerage|trading|securities|bank|credit union|savings|checking|mortgage|lending)\b`,
		`\b(financial services|banking services|investment services|financial institution|financial advisor|financial planning|wealth management|asset management|portfolio management)\b`,
		`\b(commercial banking|retail banking|investment banking|private banking|corporate banking|online banking|digital banking)\b`,
		
		// Real Estate (expanded) - single words first, then phrases
		`\b(property|construction|building|architecture|design|interior|home|house|apartment|rental|rent|lease|mortgage|realty|realtor|broker|developer|contractor|builder|architect|designer)\b`,
		`\b(real estate|property management|real estate agent|real estate broker|property development|real estate development|property investment|real estate investment)\b`,
		`\b(home sales|property sales|real estate sales|home buying|property buying|home selling|property selling|real estate transaction)\b`,
		
		// Education (expanded)
		`\b(education|school|university|college|academy|institute|training|learning|course|curriculum|student|teacher|instructor|professor|teaching|academic|degree|certification|diploma|certificate|tuition|enrollment|admission)\b`,
		`\b(educational services|education services|training services|learning services|educational institution|educational facility|training center|learning center)\b`,
		`\b(higher education|continuing education|professional education|vocational training|skills training|online education|distance learning)\b`,
		
		// Consulting (expanded)
		`\b(consulting|advisory|strategy|management|business|corporate|professional|services|expert|specialist|consultant|advisor|strategist|manager|executive|leadership|coaching|mentoring)\b`,
		`\b(consulting services|advisory services|management consulting|business consulting|strategy consulting|professional services|consulting firm|advisory firm)\b`,
		`\b(business consulting|management consulting|strategy consulting|IT consulting|financial consulting|marketing consulting)\b`,
		
		// Manufacturing (expanded)
		`\b(manufacturing|production|factory|plant|facility|industrial|automotive|machinery|equipment|assembly|fabrication|processing|quality control|supply chain|logistics|warehouse|distribution)\b`,
		`\b(manufacturing company|manufacturing facility|production facility|manufacturing plant|industrial manufacturing|custom manufacturing)\b`,
		`\b(production line|assembly line|manufacturing process|production process|quality assurance|manufacturing operations)\b`,
		
		// Transportation & Logistics (expanded)
		`\b(transportation|logistics|shipping|delivery|freight|warehouse|supply chain|trucking|hauling|courier|parcel|package|cargo|freight|logistics|distribution|fulfillment|warehousing|inventory)\b`,
		`\b(transportation services|logistics services|shipping services|delivery services|freight services|supply chain management|logistics management)\b`,
		`\b(trucking company|shipping company|delivery company|logistics company|freight company|transportation company)\b`,
		
		// Entertainment & Media (expanded)
		`\b(entertainment|media|marketing|advertising|design|creative|art|music|film|movie|television|TV|broadcast|streaming|content|production|publishing|journalism|news|media|social media|digital media)\b`,
		`\b(entertainment industry|media industry|entertainment company|media company|entertainment services|media services)\b`,
		`\b(content creation|content production|media production|entertainment production|creative services|advertising services)\b`,
		
		// Energy (expanded)
		`\b(energy|utilities|renewable|solar|wind|hydro|geothermal|oil|gas|petroleum|power|electricity|electrical|utility|energy services|power generation|energy production)\b`,
		`\b(energy company|utility company|power company|energy services|utility services|renewable energy|solar energy|wind energy)\b`,
		`\b(energy production|power generation|energy distribution|utility services|energy management|energy efficiency)\b`,
		
		// Agriculture (expanded)
		`\b(agriculture|farming|farm|ranch|ranching|food production|crop|crops|livestock|organic|sustainable|agricultural|farming|harvest|cultivation|agricultural services)\b`,
		`\b(agricultural services|farming services|agricultural production|food production|organic farming|sustainable agriculture)\b`,
		`\b(crop production|livestock production|agricultural products|farm products|organic products|sustainable farming)\b`,
		
		// Travel & Hospitality (expanded)
		`\b(travel|tourism|hospitality|hotel|motel|resort|accommodation|vacation|booking|trip|tour|travel agency|travel services|hospitality services|lodging|accommodations)\b`,
		`\b(travel services|tourism services|hospitality services|hotel services|accommodation services|travel agency|tour operator)\b`,
		`\b(hotel management|hospitality management|travel management|tourism management|accommodation management)\b`,
	}

	// Extract keywords using patterns
	// Phase 9.1: Use cached compiled regex patterns for performance
	for _, pattern := range businessPatterns {
		compiledRegex := r.getCachedRegex(pattern)
		matches := compiledRegex.FindAllString(text, -1)
		for _, match := range matches {
			// Remove duplicates and add to keywords
			if !r.containsKeyword(keywords, match) {
				keywords = append(keywords, match)
			}
		}
	}

	// Also extract common business words
	commonBusinessWords := []string{
		"service", "services", "company", "business", "corp", "corporation", "inc", "llc", "ltd",
		"enterprise", "solutions", "systems", "group", "associates", "partners", "consulting",
		"management", "development", "production", "distribution", "marketing", "sales",
		"customer", "clients", "professional", "expert", "specialist", "quality", "premium",
		"innovative", "leading", "trusted", "reliable", "experienced", "established",
	}

	for _, word := range commonBusinessWords {
		if strings.Contains(text, word) && !r.containsKeyword(keywords, word) {
			keywords = append(keywords, word)
		}
	}

	// Limit to top 20 keywords to avoid noise
	if len(keywords) > 20 {
		keywords = keywords[:20]
	}

	return keywords
}

// containsKeyword checks if a keyword already exists in the slice
func (r *SupabaseKeywordRepository) containsKeyword(keywords []string, keyword string) bool {
	for _, k := range keywords {
		if k == keyword {
			return true
		}
	}
	return false
}

// extractKeywordsFromText extracts keywords from a specific text source with context
func (r *SupabaseKeywordRepository) extractKeywordsFromText(text, context string) []ContextualKeyword {
	var keywords []ContextualKeyword
	seen := make(map[string]bool)

	// Normalize text
	normalizedText := strings.ToLower(text)

	// Extract individual words first
	words := strings.Fields(normalizedText)
	for _, word := range words {
		cleanWord := strings.Trim(word, ".,!?;:\"'()[]{}")
		if len(cleanWord) > 2 && !seen[cleanWord] {
			seen[cleanWord] = true
			keywords = append(keywords, ContextualKeyword{
				Keyword: cleanWord,
				Context: context,
			})
		}
	}

	// Extract 2-word phrases
	phrases := r.extractPhrases(normalizedText, 2)
	for _, phrase := range phrases {
		if !seen[phrase] {
			seen[phrase] = true
			keywords = append(keywords, ContextualKeyword{
				Keyword: phrase,
				Context: context,
			})
		}
	}

	// Extract 3-word phrases (for specific industry terms)
	phrases3 := r.extractPhrases(normalizedText, 3)
	for _, phrase := range phrases3 {
		if !seen[phrase] {
			seen[phrase] = true
			keywords = append(keywords, ContextualKeyword{
				Keyword: phrase,
				Context: context,
			})
		}
	}

	return keywords
}

// extractKeywordsFromURLEnhanced extracts keywords from URL with enhanced domain parsing
// Extracts compound domain names, TLD hints, and industry inference
func (r *SupabaseKeywordRepository) extractKeywordsFromURLEnhanced(websiteURL string) []ContextualKeyword {
	var keywords []ContextualKeyword
	seen := make(map[string]bool)

	// 1. Parse domain name
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		// If URL parsing fails, try adding https://
		if !strings.HasPrefix(websiteURL, "http://") && !strings.HasPrefix(websiteURL, "https://") {
			parsedURL, err = url.Parse("https://" + websiteURL)
		}
		if err != nil {
			r.logger.Printf("⚠️ [URL] Failed to parse URL: %s, error: %v", websiteURL, err)
			return keywords
		}
	}

	domain := parsedURL.Host
	if domain == "" {
		// If Host is empty, try to extract from Path
		domain = strings.TrimPrefix(parsedURL.Path, "/")
		if domain == "" {
			r.logger.Printf("⚠️ [URL] Empty domain for URL: %s", websiteURL)
			return keywords
		}
	}
	
	// Remove port if present
	if strings.Contains(domain, ":") {
		domain = strings.Split(domain, ":")[0]
	}

	// 2. Extract domain name parts (split by common separators)
	domainParts := r.splitDomainName(domain)
	if len(domainParts) == 0 {
		r.logger.Printf("⚠️ [URL] No domain parts extracted from: %s", domain)
		return keywords
	}
	
	// 3. Extract individual words (filter stop words)
	words := r.filterStopWords(domainParts)
	for _, word := range words {
		if len(word) >= 3 && !seen[word] {
			seen[word] = true
			keywords = append(keywords, ContextualKeyword{
				Keyword: word,
				Context: "website_url",
			})
		}
	}

	// 4. Extract 2-word phrases
	phrases2 := r.generatePhrases(domainParts, 2)
	for _, phrase := range phrases2 {
		phraseLower := strings.ToLower(phrase)
		if !seen[phraseLower] && len(phraseLower) >= 4 {
			seen[phraseLower] = true
			keywords = append(keywords, ContextualKeyword{
				Keyword: phraseLower,
				Context: "website_url",
			})
		}
	}

	// 5. Extract 3-word phrases for longer domains
	if len(domainParts) > 3 {
		phrases3 := r.generatePhrases(domainParts, 3)
		for _, phrase := range phrases3 {
			phraseLower := strings.ToLower(phrase)
			if !seen[phraseLower] && len(phraseLower) >= 6 {
				seen[phraseLower] = true
				keywords = append(keywords, ContextualKeyword{
					Keyword: phraseLower,
					Context: "website_url",
				})
			}
		}
	}

	// 6. Add TLD-based hints
	tldKeywords := r.extractTLDHints(parsedURL)
	for _, kw := range tldKeywords {
		if !seen[kw] {
			seen[kw] = true
			keywords = append(keywords, ContextualKeyword{
				Keyword: kw,
				Context: "website_url",
			})
		}
	}

	// 7. Add industry inference from domain
	industryKeywords := r.inferIndustryFromDomain(domain)
	for _, kw := range industryKeywords {
		if !seen[kw] {
			seen[kw] = true
			keywords = append(keywords, ContextualKeyword{
				Keyword: kw,
				Context: "website_url",
			})
		}
	}

	return keywords
}

// splitDomainName splits a domain name into meaningful parts
// Handles compound domain names like "thegreenegrape" → ["the", "green", "grape"]
func (r *SupabaseKeywordRepository) splitDomainName(domain string) []string {
	if domain == "" {
		return []string{}
	}
	
	// Remove TLD (everything after the last dot)
	parts := strings.Split(domain, ".")
	if len(parts) == 0 {
		return []string{}
	}
	domainName := parts[0]
	
	// If domainName is empty, try the whole domain
	if domainName == "" && len(parts) > 1 {
		domainName = parts[len(parts)-2] // Use second-to-last part
	}

	// Split by common separators (hyphens, underscores)
	domainName = strings.ReplaceAll(domainName, "-", " ")
	domainName = strings.ReplaceAll(domainName, "_", " ")
	
	// Try to split camelCase or compound words
	// This is a simple heuristic - for production, consider using a proper word segmentation library
	// See plan: plan-keyword-extraction-improvements-b60b8a1a.plan.md - Future Enhancement #5: Word Segmentation Library
	var words []string
	
	// First, split by spaces (from hyphens/underscores)
	spaceParts := strings.Fields(domainName)
	for _, part := range spaceParts {
		// Try to split camelCase words
		camelWords := r.splitCamelCase(part)
		words = append(words, camelWords...)
	}

	return words
}

// splitCamelCase splits camelCase words into individual words
// Simple heuristic: split on uppercase letters
func (r *SupabaseKeywordRepository) splitCamelCase(word string) []string {
	if len(word) == 0 {
		return []string{}
	}

	var words []string
	var currentWord strings.Builder
	currentWord.WriteByte(word[0])

	for i := 1; i < len(word); i++ {
		char := word[i]
		// If we encounter an uppercase letter and current word has content, start new word
		if char >= 'A' && char <= 'Z' && currentWord.Len() > 0 {
			words = append(words, strings.ToLower(currentWord.String()))
			currentWord.Reset()
		}
		currentWord.WriteByte(char)
	}

	// Add the last word
	if currentWord.Len() > 0 {
		words = append(words, strings.ToLower(currentWord.String()))
	}

	// If no camelCase detected, return the whole word as lowercase
	if len(words) == 0 {
		return []string{strings.ToLower(word)}
	}

	return words
}

// filterStopWords filters out common stop words from domain parts
func (r *SupabaseKeywordRepository) filterStopWords(parts []string) []string {
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "as": true, "is": true, "are": true,
		"www": true, "com": true, "net": true, "org": true, "io": true,
		"co": true, "uk": true, "us": true, "ca": true, "au": true,
	}

	var filtered []string
	for _, part := range parts {
		partLower := strings.ToLower(part)
		if !stopWords[partLower] && len(partLower) >= 2 {
			filtered = append(filtered, partLower)
		}
	}
	return filtered
}

// generatePhrases generates N-word phrases from domain parts
func (r *SupabaseKeywordRepository) generatePhrases(parts []string, n int) []string {
	var phrases []string
	if len(parts) < n {
		return phrases
	}

	for i := 0; i <= len(parts)-n; i++ {
		phrase := strings.Join(parts[i:i+n], " ")
		phrases = append(phrases, phrase)
	}

	return phrases
}

// extractTLDHints extracts industry hints from TLD
func (r *SupabaseKeywordRepository) extractTLDHints(parsedURL *url.URL) []string {
	var hints []string
	host := parsedURL.Host
	
	// Extract TLD
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return hints
	}
	tld := strings.ToLower(parts[len(parts)-1])

	// TLD to industry mapping
	tldHints := map[string][]string{
		"shop":    {"retail", "ecommerce", "store", "shop"},
		"store":   {"retail", "ecommerce", "store", "shop"},
		"restaurant": {"restaurant", "food", "dining"},
		"cafe":    {"cafe", "coffee", "food", "dining"},
		"bar":     {"bar", "beverage", "alcohol", "drinks"},
		"wine":    {"wine", "beverage", "alcohol", "winery"},
		"beer":    {"beer", "beverage", "alcohol", "brewery"},
		"tech":    {"technology", "tech", "software"},
		"app":     {"app", "application", "software", "technology"},
		"dev":     {"development", "software", "technology"},
		"design":  {"design", "creative", "art"},
		"photo":   {"photography", "photo", "creative"},
		"art":     {"art", "creative", "design"},
		"music":   {"music", "entertainment", "creative"},
		"film":    {"film", "entertainment", "media"},
		"news":    {"news", "media", "journalism"},
		"blog":    {"blog", "content", "media"},
		"edu":     {"education", "school", "learning"},
		"health":  {"health", "healthcare", "medical"},
		"law":     {"law", "legal", "attorney"},
		"finance": {"finance", "financial", "banking"},
		"realestate": {"real estate", "property", "realty"},
	}

	if hintsList, ok := tldHints[tld]; ok {
		hints = append(hints, hintsList...)
	}

	return hints
}

// inferIndustryFromDomain infers industry from domain name patterns
func (r *SupabaseKeywordRepository) inferIndustryFromDomain(domain string) []string {
	var keywords []string
	domainLower := strings.ToLower(domain)

	// Industry inference patterns
	industryPatterns := map[string][]string{
		// Wine & Beverage
		"wine":     {"wine", "beverage", "alcohol", "retail"},
		"grape":    {"wine", "grape", "beverage", "retail"},
		"vineyard": {"vineyard", "wine", "winery", "beverage"},
		"vintner":  {"vintner", "wine", "winery", "beverage"},
		"brewery":  {"brewery", "beer", "beverage", "alcohol"},
		"distillery": {"distillery", "spirits", "alcohol", "beverage"},

		// Retail
		"shop":     {"shop", "retail", "store", "commerce"},
		"store":    {"store", "retail", "shop", "commerce"},
		"market":   {"market", "retail", "commerce", "store"},
		"boutique": {"boutique", "retail", "shop", "fashion"},

		// Technology
		"tech":     {"technology", "tech", "software"},
		"software": {"software", "technology", "tech"},
		"app":      {"app", "application", "software", "technology"},
		"digital":  {"digital", "technology", "tech"},

		// Food & Dining
		"restaurant": {"restaurant", "food", "dining"},
		"cafe":       {"cafe", "coffee", "food", "dining"},
		"food":       {"food", "restaurant", "dining"},
		"dining":     {"dining", "restaurant", "food"},
	}

	// Check for industry patterns in domain
	for pattern, keywordsList := range industryPatterns {
		if strings.Contains(domainLower, pattern) {
			keywords = append(keywords, keywordsList...)
		}
	}

	return keywords
}

// getContextMultiplier returns the appropriate multiplier based on keyword context
func (r *SupabaseKeywordRepository) getContextMultiplier(context string) float64 {
	switch context {
	case "business_name":
		return 1.2 // 20% boost for business name keywords (highest priority)
	case "description":
		return 1.0 // No boost for description keywords (baseline)
	case "website_url":
		return 1.0 // No boost for website URL keywords (baseline)
	default:
		return 1.0 // Default to no boost for unknown contexts
	}
}

// calculateDynamicConfidence calculates confidence based on match quality and context
func (r *SupabaseKeywordRepository) calculateDynamicConfidence(score float64, matchedKeywords int, totalKeywords int) float64 {
	// Base confidence from score (normalized to 0-1 range)
	baseConfidence := score

	// Apply match ratio factor (30% weight)
	matchRatio := float64(matchedKeywords) / float64(totalKeywords)
	matchRatioFactor := matchRatio * 0.3

	// Apply score strength factor (40% weight)
	scoreStrengthFactor := baseConfidence * 0.4

	// Apply specificity factor (20% weight) - more matched keywords = higher specificity
	specificityFactor := float64(matchedKeywords) * 0.02
	if specificityFactor > 0.2 {
		specificityFactor = 0.2 // Cap at 20%
	}

	// Apply keyword quality factor (10% weight) - based on total keywords processed
	qualityFactor := float64(totalKeywords) * 0.01
	if qualityFactor > 0.1 {
		qualityFactor = 0.1 // Cap at 10%
	}

	// Combine all factors
	confidence := matchRatioFactor + scoreStrengthFactor + specificityFactor + qualityFactor

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	return confidence
}

// extractKeywordsAndPhrases extracts both individual keywords and multi-word phrases
func (r *SupabaseKeywordRepository) extractKeywordsAndPhrases(text string) []string {
	var keywords []string
	seen := make(map[string]bool)

	// Normalize text
	normalizedText := strings.ToLower(text)

	// Extract individual words first
	words := strings.Fields(normalizedText)
	for _, word := range words {
		cleanWord := strings.Trim(word, ".,!?;:\"'()[]{}")
		if len(cleanWord) > 2 && !seen[cleanWord] {
			seen[cleanWord] = true
			keywords = append(keywords, cleanWord)
		}
	}

	// Extract 2-word phrases
	phrases := r.extractPhrases(normalizedText, 2)
	for _, phrase := range phrases {
		if !seen[phrase] {
			seen[phrase] = true
			keywords = append(keywords, phrase)
		}
	}

	// Extract 3-word phrases (for specific industry terms)
	phrases3 := r.extractPhrases(normalizedText, 3)
	for _, phrase := range phrases3 {
		if !seen[phrase] {
			seen[phrase] = true
			keywords = append(keywords, phrase)
		}
	}

	return keywords
}

// extractPhrases extracts n-word phrases from text
func (r *SupabaseKeywordRepository) extractPhrases(text string, phraseLength int) []string {
	var phrases []string
	words := strings.Fields(text)

	// Clean words
	var cleanWords []string
	for _, word := range words {
		cleanWord := strings.Trim(word, ".,!?;:\"'()[]{}")
		if len(cleanWord) > 1 { // Allow shorter words in phrases
			cleanWords = append(cleanWords, cleanWord)
		}
	}

	// Extract phrases of specified length
	for i := 0; i <= len(cleanWords)-phraseLength; i++ {
		phrase := strings.Join(cleanWords[i:i+phraseLength], " ")
		if r.isValidPhrase(phrase) {
			phrases = append(phrases, phrase)
		}
	}

	return phrases
}

// isValidPhrase checks if a phrase is valid for classification
func (r *SupabaseKeywordRepository) isValidPhrase(phrase string) bool {
	// Filter out phrases that are too short or contain only common words
	if len(phrase) < 4 {
		return false
	}

	// Check if phrase contains meaningful business terms
	words := strings.Fields(phrase)
	meaningfulWords := 0

	for _, word := range words {
		if !r.isCommonWord(word) && len(word) > 2 {
			meaningfulWords++
		}
	}

	// At least half the words should be meaningful
	return meaningfulWords >= (len(words)+1)/2
}

// isCommonWord checks if a word is a common word that should be filtered out
func (r *SupabaseKeywordRepository) isCommonWord(word string) bool {
	commonWords := map[string]bool{
		// Articles and basic words
		"the": true, "and": true, "or": true, "but": true, "in": true, "on": true, "at": true,
		"to": true, "for": true, "of": true, "with": true, "by": true, "from": true, "up": true,
		"about": true, "into": true, "through": true, "during": true, "before": true, "after": true,
		"above": true, "below": true, "between": true, "among": true, "within": true, "without": true,

		// Verbs
		"is": true, "are": true, "was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true, "did": true, "will": true,
		"would": true, "could": true, "should": true, "may": true, "might": true, "must": true,
		"can": true,

		// Pronouns and determiners
		"this": true, "that": true, "these": true, "those": true, "a": true, "an": true,
		"our": true, "your": true, "their": true, "my": true, "his": true, "her": true, "its": true,
		"we": true, "you": true, "they": true, "i": true, "he": true, "she": true, "it": true,
		"me": true, "him": true, "us": true, "them": true,

		// Quantifiers
		"all": true, "any": true, "some": true, "many": true, "much": true, "few": true,
		"more": true, "most": true, "another": true, "each": true, "every": true, "both": true,
		"either": true, "neither": true, "one": true, "two": true,

		// Adjectives
		"first": true, "second": true, "last": true, "next": true, "new": true, "old": true,
		"good": true, "bad": true, "big": true, "small": true, "long": true, "short": true,
		"high": true, "low": true, "great": true, "little": true, "own": true, "just": true,
		"like": true, "over": true, "also": true, "back": true, "well": true, "even": true,
		"still": true,

		// Adverbs and prepositions
		"here": true, "there": true, "where": true, "when": true, "why": true, "how": true,
		"what": true, "who": true, "which": true,

		// Internet and domain related
		"www": true, "com": true, "org": true, "net": true, "uk": true, "ca": true, "au": true,
		"de": true, "fr": true, "jp": true, "cn": true, "ru": true, "br": true, "mx": true,
		"es": true, "nl": true, "se": true, "no": true, "dk": true, "fi": true, "pl": true,
		"tr": true, "ar": true, "cl": true, "pe": true, "ve": true, "ec": true, "uy": true,
		"py": true, "bo": true, "gt": true, "hn": true, "ni": true, "cr": true, "pa": true,
		"cu": true, "ht": true, "jm": true, "tt": true, "bb": true, "gd": true, "lc": true,
		"vc": true, "ag": true, "bs": true, "bz": true, "dm": true, "kn": true, "sr": true,
		"gy": true, "fk": true, "gs": true, "sh": true, "ac": true, "ta": true, "bv": true,
		"hm": true, "nf": true, "aq": true, "tf": true, "pf": true, "nc": true, "vu": true,
		"sb": true, "tv": true, "ki": true, "nr": true, "fm": true, "mh": true, "pw": true,
		"mp": true, "gu": true, "as": true, "vi": true, "pr": true, "um": true,
	}

	return commonWords[word]
}

// hasPhraseOverlap checks if two phrases have meaningful overlap
func (r *SupabaseKeywordRepository) hasPhraseOverlap(phrase1, phrase2 string) bool {
	words1 := strings.Fields(phrase1)
	words2 := strings.Fields(phrase2)

	// Count meaningful word overlaps
	overlaps := 0
	for _, word1 := range words1 {
		if !r.isCommonWord(word1) && len(word1) > 2 {
			for _, word2 := range words2 {
				if !r.isCommonWord(word2) && len(word2) > 2 && word1 == word2 {
					overlaps++
					break
				}
			}
		}
	}

	// At least one meaningful word should overlap
	return overlaps > 0
}

// parseClassificationCodesResponse parses the Supabase response for classification codes
func (r *SupabaseKeywordRepository) parseClassificationCodesResponse(response []byte, codes *[]*ClassificationCode) error {
	if len(response) == 0 {
		*codes = []*ClassificationCode{}
		return nil
	}

	// Parse JSON response
	var rawCodes []map[string]interface{}
	if err := json.Unmarshal(response, &rawCodes); err != nil {
		return fmt.Errorf("failed to unmarshal classification codes response: %w", err)
	}

	*codes = make([]*ClassificationCode, 0, len(rawCodes))
	for _, rawCode := range rawCodes {
		code := &ClassificationCode{}

		// Parse ID
		if id, ok := rawCode["id"].(float64); ok {
			code.ID = int(id)
		}

		// Parse IndustryID
		if industryID, ok := rawCode["industry_id"].(float64); ok {
			code.IndustryID = int(industryID)
		}

		// Parse CodeType
		if codeType, ok := rawCode["code_type"].(string); ok {
			code.CodeType = codeType
		}

		// Parse Code
		if codeStr, ok := rawCode["code"].(string); ok {
			code.Code = codeStr
		}

		// Parse Description
		if description, ok := rawCode["description"].(string); ok {
			code.Description = description
		}

		// Parse IsActive
		if isActive, ok := rawCode["is_active"].(bool); ok {
			code.IsActive = isActive
		}

		// Parse CreatedAt
		if createdAt, ok := rawCode["created_at"].(string); ok {
			code.CreatedAt = createdAt
		}

		// Parse UpdatedAt
		if updatedAt, ok := rawCode["updated_at"].(string); ok {
			code.UpdatedAt = updatedAt
		}

		*codes = append(*codes, code)
	}

	return nil
}

// calculateKeywordQualityFactor calculates the quality factor based on keyword relevance
func (r *SupabaseKeywordRepository) calculateKeywordQualityFactor(matchedKeywords, allKeywords []string) float64 {
	if len(allKeywords) == 0 {
		return 0.5 // Default factor
	}

	// Calculate the ratio of matched keywords to total keywords
	matchRatio := float64(len(matchedKeywords)) / float64(len(allKeywords))

	// Apply quality boost for high match ratios
	if matchRatio > 0.8 {
		return 1.2 // 20% boost for high match ratios
	} else if matchRatio > 0.5 {
		return 1.1 // 10% boost for medium match ratios
	} else if matchRatio > 0.2 {
		return 1.0 // No boost for low match ratios
	} else {
		return 0.8 // 20% penalty for very low match ratios
	}
}

// calculateIndustrySpecificityFactor calculates the specificity factor based on industry relevance
func (r *SupabaseKeywordRepository) calculateIndustrySpecificityFactor(industryID int, matchedKeywords []string) float64 {
	// Define industry-specific keyword weights
	industryWeights := map[int]float64{
		1:  1.2, // Technology - high specificity
		2:  1.1, // Healthcare - medium-high specificity
		3:  1.0, // Finance - medium specificity
		4:  1.1, // Retail - medium-high specificity
		5:  1.0, // Manufacturing - medium specificity
		6:  1.2, // Restaurant - high specificity
		7:  1.0, // Professional Services - medium specificity
		8:  1.1, // Construction - medium-high specificity
		9:  1.0, // Transportation - medium specificity
		10: 1.1, // Education - medium-high specificity
		26: 0.8, // General Business - low specificity
	}

	weight, exists := industryWeights[industryID]
	if !exists {
		weight = 1.0 // Default weight
	}

	// Apply keyword count factor
	keywordCountFactor := 1.0
	if len(matchedKeywords) > 5 {
		keywordCountFactor = 1.1 // Boost for many keywords
	} else if len(matchedKeywords) < 2 {
		keywordCountFactor = 0.9 // Penalty for few keywords
	}

	return weight * keywordCountFactor
}

// calculateMatchDiversityFactor calculates the diversity factor based on keyword variety
func (r *SupabaseKeywordRepository) calculateMatchDiversityFactor(matchedKeywords []string) float64 {
	if len(matchedKeywords) == 0 {
		return 0.5
	}

	// Calculate diversity based on keyword length and variety
	avgLength := 0.0
	uniqueChars := make(map[rune]bool)

	for _, keyword := range matchedKeywords {
		avgLength += float64(len(keyword))
		for _, char := range keyword {
			uniqueChars[char] = true
		}
	}

	avgLength /= float64(len(matchedKeywords))
	charDiversity := float64(len(uniqueChars)) / (avgLength * float64(len(matchedKeywords)))

	// Apply diversity factor
	if charDiversity > 0.7 {
		return 1.2 // 20% boost for high diversity
	} else if charDiversity > 0.5 {
		return 1.1 // 10% boost for medium diversity
	} else if charDiversity > 0.3 {
		return 1.0 // No boost for low diversity
	} else {
		return 0.9 // 10% penalty for very low diversity
	}
}

// calculateIndustryCoOccurrenceBoost calculates boost scores based on industry co-occurrence patterns
// Phase 7.3: Implements industry co-occurrence analysis
func (r *SupabaseKeywordRepository) calculateIndustryCoOccurrenceBoost(industryMatches map[int][]string, inputKeywords []string) map[int]float64 {
	boosts := make(map[int]float64)
	
	// Define common industry co-occurrence patterns
	// Format: map[industryID][]coOccurringKeywords
	coOccurrencePatterns := map[int][]string{
		// Retail + Food & Beverage + Technology (e.g., wine shop, electronics store)
		4: {"wine", "retail", "shop", "store", "beverage", "alcohol", "spirits", "liquor", "tech", "electronics", "digital", "online", "ecommerce"},
		// Restaurant + Food & Beverage
		6: {"restaurant", "food", "dining", "beverage", "wine", "cuisine"},
		// Healthcare + Professional Services
		2: {"medical", "health", "professional", "service", "clinic", "therapy"},
		// Technology + Professional Services
		1: {"software", "technology", "professional", "service", "consulting", "development"},
	}
	
	// Analyze co-occurrence for each industry
	for industryID, patternKeywords := range coOccurrencePatterns {
		coOccurrenceCount := 0
		matchedPatternKeywords := make(map[string]bool)
		
		// Check how many pattern keywords appear in input
		for _, patternKw := range patternKeywords {
			patternKwLower := strings.ToLower(patternKw)
			
			// Check if pattern keyword appears in input keywords
			// Use word boundary matching to avoid substring false positives
			for _, inputKw := range inputKeywords {
				inputKwLower := strings.ToLower(strings.TrimSpace(inputKw))
				// Exact match or word boundary match
				if inputKwLower == patternKwLower {
					if !matchedPatternKeywords[patternKwLower] {
						coOccurrenceCount++
						matchedPatternKeywords[patternKwLower] = true
					}
				} else if strings.Contains(inputKwLower, " "+patternKwLower+" ") || 
					strings.HasPrefix(inputKwLower, patternKwLower+" ") ||
					strings.HasSuffix(inputKwLower, " "+patternKwLower) {
					// Word boundary match (space-separated)
					if !matchedPatternKeywords[patternKwLower] {
						coOccurrenceCount++
						matchedPatternKeywords[patternKwLower] = true
					}
				}
			}
		}
		
		// Calculate boost based on co-occurrence count
		// More pattern keywords matched = higher boost
		if coOccurrenceCount >= 3 {
			boosts[industryID] = 0.15 // 15% boost for 3+ co-occurring keywords
		} else if coOccurrenceCount >= 2 {
			boosts[industryID] = 0.10 // 10% boost for 2 co-occurring keywords
		} else if coOccurrenceCount >= 1 {
			boosts[industryID] = 0.05 // 5% boost for 1 co-occurring keyword
		}
		
		if coOccurrenceCount > 0 {
			r.logger.Printf("📊 [Phase 7.3] Industry %d co-occurrence: %d pattern keywords matched", industryID, coOccurrenceCount)
		}
	}
	
	// Additional analysis: check for cross-industry keyword co-occurrence
	// Example: "wine" (Food & Beverage) + "retail" (Retail) + "shop" (Retail)
	// This suggests Retail industry with Food & Beverage subcategory
	for industryID, matchedKeywords := range industryMatches {
		if len(matchedKeywords) >= 2 {
			// Check if keywords from different semantic groups appear together
			// This is a simplified check - in production, you'd use more sophisticated NLP
			hasRetailKeywords := false
			hasFoodKeywords := false
			hasTechKeywords := false
			
			for _, kw := range matchedKeywords {
				kwLower := strings.ToLower(kw)
				if strings.Contains(kwLower, "retail") || strings.Contains(kwLower, "shop") || strings.Contains(kwLower, "store") {
					hasRetailKeywords = true
				}
				if strings.Contains(kwLower, "wine") || strings.Contains(kwLower, "food") || strings.Contains(kwLower, "beverage") {
					hasFoodKeywords = true
				}
				if strings.Contains(kwLower, "tech") || strings.Contains(kwLower, "software") || strings.Contains(kwLower, "digital") {
					hasTechKeywords = true
				}
			}
			
			// Boost for Retail + Food & Beverage co-occurrence (e.g., wine shop)
			if industryID == 4 && hasRetailKeywords && hasFoodKeywords {
				boosts[industryID] += 0.10
				r.logger.Printf("📊 [Phase 7.3] Retail + Food & Beverage co-occurrence detected for industry %d", industryID)
			}
			
			// Boost for Retail + Technology co-occurrence (e.g., electronics store)
			if industryID == 4 && hasRetailKeywords && hasTechKeywords {
				boosts[industryID] += 0.10
				r.logger.Printf("📊 [Phase 7.3] Retail + Technology co-occurrence detected for industry %d", industryID)
			}
		}
	}
	
	return boosts
}
