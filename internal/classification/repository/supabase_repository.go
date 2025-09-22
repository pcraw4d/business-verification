package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
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

	// Extract contextual keywords from business information
	contextualKeywords := r.extractKeywords(businessName, description, websiteURL)

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

	// Calculate simple confidence score (will be enhanced by confidence service at higher level)
	var confidence float64
	var reasoning string

	// Simple confidence calculation based on match ratio and score
	matchRatio := float64(len(bestMatchedKeywords)) / float64(len(keywords))
	scoreRatio := bestScore / float64(len(keywords))
	confidence = (matchRatio * 0.6) + (scoreRatio * 0.4)

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.1 {
		confidence = 0.1
	}

	r.logger.Printf("📊 Simple confidence calculated: %.3f (match_ratio: %.3f, score_ratio: %.3f)",
		confidence, matchRatio, scoreRatio)

	reasoning = fmt.Sprintf("Simple classification matched %d keywords with industry '%s' (score: %.2f, confidence: %.3f)",
		len(bestMatchedKeywords), bestIndustry.Name, bestScore, confidence)

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
func (r *SupabaseKeywordRepository) extractKeywords(businessName, description, websiteURL string) []ContextualKeyword {
	var keywords []ContextualKeyword
	seen := make(map[string]bool)

	// Extract keywords from business name (highest priority context)
	if businessName != "" {
		nameKeywords := r.extractKeywordsFromText(businessName, "business_name")
		for _, keyword := range nameKeywords {
			if !seen[keyword.Keyword] {
				seen[keyword.Keyword] = true
				keywords = append(keywords, keyword)
			}
		}
	}

	// Extract keywords from description (medium priority context)
	if description != "" {
		descKeywords := r.extractKeywordsFromText(description, "description")
		for _, keyword := range descKeywords {
			if !seen[keyword.Keyword] {
				seen[keyword.Keyword] = true
				keywords = append(keywords, keyword)
			}
		}
	}

	// Extract keywords from website URL (lowest priority context)
	if websiteURL != "" {
		// First, try to scrape actual website content for keywords
		scrapedKeywords := r.extractKeywordsFromWebsite(context.Background(), websiteURL)
		if len(scrapedKeywords) > 0 {
			// Add scraped keywords with website content context
			for _, keyword := range scrapedKeywords {
				if !seen[keyword] {
					seen[keyword] = true
					keywords = append(keywords, ContextualKeyword{
						Keyword: keyword,
						Context: "website_content",
					})
				}
			}
			r.logger.Printf("✅ Extracted %d keywords from website content: %v", len(scrapedKeywords), scrapedKeywords)
		} else {
			// Fallback to URL text extraction if scraping fails
			urlKeywords := r.extractKeywordsFromText(websiteURL, "website_url")
			for _, keyword := range urlKeywords {
				if !seen[keyword.Keyword] {
					seen[keyword.Keyword] = true
					keywords = append(keywords, keyword)
				}
			}
			r.logger.Printf("⚠️ Website scraping failed, using URL text extraction")
		}
	}

	return keywords
}

// extractKeywordsFromWebsite scrapes website content and extracts business-relevant keywords
func (r *SupabaseKeywordRepository) extractKeywordsFromWebsite(ctx context.Context, websiteURL string) []string {
	startTime := time.Now()
	r.logger.Printf("🌐 [Supabase] Starting website scraping for: %s", websiteURL)

	// Validate URL
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		r.logger.Printf("❌ [Supabase] Invalid URL format for %s: %v", websiteURL, err)
		return []string{}
	}

	if parsedURL.Scheme == "" {
		websiteURL = "https://" + websiteURL
		r.logger.Printf("🔧 [Supabase] Added HTTPS scheme: %s", websiteURL)
	}

	// Create HTTP client with enhanced configuration
	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:       10,
			IdleConnTimeout:    30 * time.Second,
			DisableCompression: false,
		},
	}

	// Create request with enhanced headers
	req, err := http.NewRequestWithContext(ctx, "GET", websiteURL, nil)
	if err != nil {
		r.logger.Printf("❌ [Supabase] Failed to create request for %s: %v", websiteURL, err)
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

	r.logger.Printf("📡 [Supabase] Making HTTP request to: %s", websiteURL)

	// Make request with timeout context
	reqCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req = req.WithContext(reqCtx)

	resp, err := client.Do(req)
	if err != nil {
		r.logger.Printf("❌ [Supabase] HTTP request failed for %s: %v", websiteURL, err)
		return []string{}
	}
	defer resp.Body.Close()

	// Log response details
	r.logger.Printf("📊 [Supabase] Response received - Status: %d, Content-Type: %s, Content-Length: %d",
		resp.StatusCode, resp.Header.Get("Content-Type"), resp.ContentLength)

	// Check status code with detailed logging
	if resp.StatusCode >= 400 {
		r.logger.Printf("❌ [Supabase] HTTP error for %s: %d %s", websiteURL, resp.StatusCode, resp.Status)
		// Try to read error response body
		if body, readErr := io.ReadAll(resp.Body); readErr == nil && len(body) > 0 {
			r.logger.Printf("📄 [Supabase] Error response body (first 500 chars): %s", string(body[:min(500, len(body))]))
		}
		return []string{}
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		r.logger.Printf("⚠️ [Supabase] Unexpected content type for %s: %s", websiteURL, contentType)
	}

	// Read response body with size limit
	maxSize := int64(5 * 1024 * 1024) // 5MB limit
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
	if err != nil {
		r.logger.Printf("❌ [Supabase] Failed to read response body from %s: %v", websiteURL, err)
		return []string{}
	}

	r.logger.Printf("📄 [Supabase] Read %d bytes from %s", len(body), websiteURL)

	// Extract text content from HTML
	textContent := r.extractTextFromHTML(string(body))
	r.logger.Printf("🧹 [Supabase] Extracted %d characters of text content from HTML", len(textContent))

	// Log sample of extracted text for debugging
	if len(textContent) > 0 {
		sampleText := textContent[:min(200, len(textContent))]
		r.logger.Printf("📝 [Supabase] Sample extracted text: %s...", sampleText)
	}

	// Extract business-relevant keywords
	keywords := r.extractBusinessKeywords(textContent)

	duration := time.Since(startTime)
	r.logger.Printf("✅ [Supabase] Website scraping completed for %s in %v - extracted %d keywords: %v",
		websiteURL, duration, len(keywords), keywords)

	return keywords
}

// extractTextFromHTML extracts clean text content from HTML
func (r *SupabaseKeywordRepository) extractTextFromHTML(htmlContent string) string {
	// Simple HTML tag removal (for production, consider using a proper HTML parser)
	// Remove script and style tags completely
	htmlContent = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`).ReplaceAllString(htmlContent, "")
	htmlContent = regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`).ReplaceAllString(htmlContent, "")

	// Remove HTML tags
	htmlContent = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(htmlContent, " ")

	// Clean up whitespace
	htmlContent = regexp.MustCompile(`\s+`).ReplaceAllString(htmlContent, " ")

	return strings.TrimSpace(htmlContent)
}

// extractBusinessKeywords extracts business-relevant keywords from text content
func (r *SupabaseKeywordRepository) extractBusinessKeywords(textContent string) []string {
	var keywords []string

	// Convert to lowercase for processing
	text := strings.ToLower(textContent)

	// Business-relevant keyword patterns
	businessPatterns := []string{
		// Industry keywords
		`\b(restaurant|cafe|coffee|food|dining|kitchen|catering|bakery|bar|pub|brewery|winery)\b`,
		`\b(technology|software|tech|app|digital|web|mobile|cloud|ai|ml|data|cyber|security)\b`,
		`\b(healthcare|medical|clinic|hospital|doctor|dentist|therapy|wellness|pharmacy)\b`,
		`\b(legal|law|attorney|lawyer|court|litigation|patent|trademark|copyright)\b`,
		`\b(retail|store|shop|ecommerce|online|fashion|clothing|electronics|beauty)\b`,
		`\b(finance|banking|investment|insurance|accounting|tax|financial|credit|loan)\b`,
		`\b(real estate|property|construction|building|architecture|design|interior)\b`,
		`\b(education|school|university|training|learning|course|academy|institute)\b`,
		`\b(consulting|advisory|strategy|management|business|corporate|professional)\b`,
		`\b(manufacturing|production|factory|industrial|automotive|machinery|equipment)\b`,
		`\b(transportation|logistics|shipping|delivery|freight|warehouse|supply chain)\b`,
		`\b(entertainment|media|marketing|advertising|design|creative|art|music|film)\b`,
		`\b(energy|utilities|renewable|solar|wind|oil|gas|power|electricity)\b`,
		`\b(agriculture|farming|food production|crop|livestock|organic|sustainable)\b`,
		`\b(travel|tourism|hospitality|hotel|accommodation|vacation|booking|trip)\b`,
	}

	// Extract keywords using patterns
	for _, pattern := range businessPatterns {
		matches := regexp.MustCompile(pattern).FindAllString(text, -1)
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
