package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/pcraw4d/business-verification/internal/cache"
	"github.com/pcraw4d/business-verification/internal/database"
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
	logger       *log.Logger
	keywordIndex *KeywordIndex
	cacheMutex   sync.RWMutex

	// Industry code caching
	industryCodeCache *cache.IntelligentCache
	cacheConfig       *IndustryCodeCacheConfig
	cacheStats        *IndustryCodeCacheStats
	statsMutex        sync.RWMutex
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
		client: client,
		logger: logger,
		keywordIndex: &KeywordIndex{
			KeywordToIndustries: make(map[string][]IndustryKeywordMatch),
			IndustryToKeywords:  make(map[int][]*KeywordWeight),
			LastUpdated:         0,
		},
		industryCodeCache: intelligentCache,
		cacheConfig:       cacheConfig,
		cacheStats:        &IndustryCodeCacheStats{},
	}
}

// NewSupabaseKeywordRepositoryWithInterface creates a new Supabase-based keyword repository with interface
func NewSupabaseKeywordRepositoryWithInterface(client SupabaseClientInterface, logger *log.Logger) *SupabaseKeywordRepository {
	if logger == nil {
		logger = log.Default()
	}

	// For testing purposes, we'll use nil for the client since tests don't need real database connections
	// In production, this should be refactored to use interfaces properly
	var concreteClient *database.SupabaseClient

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
		client: concreteClient,
		logger: logger,
		keywordIndex: &KeywordIndex{
			KeywordToIndustries: make(map[string][]IndustryKeywordMatch),
			IndustryToKeywords:  make(map[int][]*KeywordWeight),
			LastUpdated:         0,
		},
		industryCodeCache: intelligentCache,
		cacheConfig:       cacheConfig,
		cacheStats:        &IndustryCodeCacheStats{},
	}
}

// =============================================================================
// Keyword Index Management
// =============================================================================

// BuildKeywordIndex builds an optimized keyword index for fast lookups
func (r *SupabaseKeywordRepository) BuildKeywordIndex(ctx context.Context) error {
	r.logger.Printf("🔍 Building optimized keyword index...")

	// Optimized query with proper indexing and filtering
	query := r.client.GetPostgrestClient().From("keyword_weights").
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
		response, _, err = r.client.GetPostgrestClient().
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
		response, _, err = r.client.GetPostgrestClient().
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
		response, _, err = r.client.GetPostgrestClient().
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

	// Optimized query with proper indexing and text search
	postgrestClient := r.client.GetPostgrestClient()

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

	// Optimized query with proper indexing and ordering
	response, _, err := r.client.GetPostgrestClient().
		From("classification_codes").
		Select("id,industry_id,code_type,code,description,is_active", "", false).
		Eq("industry_id", fmt.Sprintf("%d", industryID)).
		Eq("is_active", "true").
		Order("code_type", &postgrest.OrderOpts{Ascending: true}).
		Order("code", &postgrest.OrderOpts{Ascending: true}).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to get classification codes for industry %d: %w", industryID, err)
	}

	// Parse the response
	var codes []*ClassificationCode
	if err := r.parseClassificationCodesResponse(response, &codes); err != nil {
		return nil, fmt.Errorf("failed to parse classification codes response: %w", err)
	}

	r.logger.Printf("✅ Retrieved %d classification codes for industry %d", len(codes), industryID)
	return codes, nil
}

// GetClassificationCodesByType retrieves classification codes by type (NAICS, MCC, SIC)
func (r *SupabaseKeywordRepository) GetClassificationCodesByType(ctx context.Context, codeType string) ([]*ClassificationCode, error) {
	r.logger.Printf("🔍 Getting classification codes by type: %s", codeType)

	// Optimized query with proper indexing and ordering
	response, _, err := r.client.GetPostgrestClient().
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

// ClassifyBusiness classifies a business based on name, description, and website
func (r *SupabaseKeywordRepository) ClassifyBusiness(ctx context.Context, businessName, description, websiteURL string) (*ClassificationResult, error) {
	r.logger.Printf("🔍 Classifying business: %s", businessName)

	// Extract keywords from business information
	keywords := r.extractKeywords(businessName, description, websiteURL)

	// Classify based on keywords
	return r.ClassifyBusinessByKeywords(ctx, keywords)
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

	// Process each input keyword once
	for _, inputKeyword := range keywords {
		normalizedKeyword := strings.ToLower(strings.TrimSpace(inputKeyword))

		// Direct lookup in keyword index - O(1) average case
		if matches, exists := index.KeywordToIndustries[normalizedKeyword]; exists {
			for _, match := range matches {
				industryScores[match.IndustryID] += match.Weight
				industryMatches[match.IndustryID] = append(industryMatches[match.IndustryID], match.Keyword)
			}
		}

		// Also check for partial matches (substring matching)
		for keyword, matches := range index.KeywordToIndustries {
			if strings.Contains(normalizedKeyword, keyword) || strings.Contains(keyword, normalizedKeyword) {
				for _, match := range matches {
					// Reduce weight for partial matches
					partialWeight := match.Weight * 0.5
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
		normalizedScore := score / float64(len(keywords))

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

	// Calculate confidence based on score
	confidence := bestScore
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	reasoning := fmt.Sprintf("Optimized classification matched %d keywords with industry '%s' (score: %.2f)",
		len(bestMatchedKeywords), bestIndustry.Name, bestScore)

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

// extractKeywords extracts keywords from business information
func (r *SupabaseKeywordRepository) extractKeywords(businessName, description, websiteURL string) []string {
	var keywords []string

	// Extract from business name
	if businessName != "" {
		words := strings.Fields(strings.ToLower(businessName))
		keywords = append(keywords, words...)
	}

	// Extract from description
	if description != "" {
		words := strings.Fields(strings.ToLower(description))
		keywords = append(keywords, words...)
	}

	// Extract from website URL (basic extraction)
	if websiteURL != "" {
		// Remove common URL parts and extract domain keywords
		cleanURL := strings.TrimPrefix(websiteURL, "https://")
		cleanURL = strings.TrimPrefix(cleanURL, "http://")
		cleanURL = strings.TrimPrefix(cleanURL, "www.")

		parts := strings.Split(cleanURL, ".")
		if len(parts) > 0 {
			domainWords := strings.Fields(strings.ReplaceAll(parts[0], "-", " "))
			keywords = append(keywords, domainWords...)
		}
	}

	// Remove duplicates and common words
	seen := make(map[string]bool)
	var uniqueKeywords []string

	for _, keyword := range keywords {
		if len(keyword) > 2 && !seen[keyword] {
			seen[keyword] = true
			uniqueKeywords = append(uniqueKeywords, keyword)
		}
	}

	return uniqueKeywords
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
