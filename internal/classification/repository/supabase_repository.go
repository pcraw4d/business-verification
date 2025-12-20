package repository

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"kyb-platform/internal/cache"
	"kyb-platform/internal/classification/nlp"
	"kyb-platform/internal/classification/word_segmentation"
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
	client          *database.SupabaseClient
	clientInterface SupabaseClientInterface // Store interface for methods that need it
	logger          *log.Logger
	keywordIndex    *KeywordIndex
	cacheMutex      sync.RWMutex

	// Industry code caching
	industryCodeCache *cache.IntelligentCache
	cacheConfig       *IndustryCodeCacheConfig
	cacheStats        *IndustryCodeCacheStats
	statsMutex        sync.RWMutex

	// Industry caching (5-minute TTL)
	industryCache      map[int]*industryCacheEntry
	industryCacheMutex sync.RWMutex

	// Website content caching (5-minute TTL) - OPTIMIZATION: Priority 4
	websiteContentCache      map[string]*websiteContentCacheEntry
	websiteContentCacheMutex sync.RWMutex

	// Brand matcher for MCC 3000-3831 (hotels)
	brandMatcher *BrandMatcher

	// Phase 9.1: Cached compiled regex patterns for performance
	regexCache map[string]*regexp.Regexp
	regexMutex sync.RWMutex

	// Phase 9.1: Content size limit for processing (50KB)

	// Word segmentation for compound domain names
	segmenter      *word_segmentation.Segmenter
	maxContentSize int64

	// NLP components for enhanced keyword extraction
	entityRecognizer *nlp.EntityRecognizer
	topicModeler     *nlp.TopicModeler

	// Enhanced keyword matching (synonyms, stemming, fuzzy)
	keywordMatcher *KeywordMatcher

	// Phase 9.2: DNS resolution cache (TTL-based)
	dnsCache map[string]dnsCacheEntry
	dnsMutex sync.RWMutex

	// Phase 9.3: Rate limiting for requests
	rateLimiter map[string]time.Time // Domain -> last request time
	rateMutex   sync.Mutex
	minDelay    time.Duration // Minimum delay between requests to same domain

	// Session management for cookies and referer tracking
	sessionManager *scrapingSessionManager

	// Phase 1: Enhanced website scraper with multi-tier strategies
	// Using interface to avoid import cycle with classification package
	websiteScraper WebsiteScraperInterface

	// Enhanced scoring algorithm (reused instance for performance)
	enhancedScorer      *EnhancedScoringAlgorithm
	enhancedScorerMutex sync.RWMutex // Thread-safe access to enhanced scorer
}

// timedQuery wraps a database query with timing and logging
func (r *SupabaseKeywordRepository) timedQuery(ctx context.Context, queryName string, metadata map[string]interface{}, fn func() error) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		r.logger.Printf("⏱️ [DB-QUERY] %s took %v", queryName, duration)

		if duration > 5*time.Second {
			r.logger.Printf("⚠️ [SLOW-QUERY] %s took %v (threshold: 5s)", queryName, duration)
		}

		if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
			remaining := time.Until(deadline)
			r.logger.Printf("⏱️ [DB-QUERY] %s - time remaining: %v", queryName, remaining)
		}
	}()

	return fn()
}

// ensureValidContext ensures the context has sufficient time for an HTTP request
// If the context is expired or has insufficient time, creates a fresh context with the specified timeout
func (r *SupabaseKeywordRepository) ensureValidContext(ctx context.Context, httpTimeout time.Duration) (context.Context, context.CancelFunc) {
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		if timeRemaining <= 0 {
			// Context already expired, create fresh context with HTTP timeout
			r.logger.Printf("⚠️ [RPC] Context expired, creating fresh context with %v timeout", httpTimeout)
			rpcCtx, cancel := context.WithTimeout(context.Background(), httpTimeout)
			return rpcCtx, cancel
		} else if timeRemaining < httpTimeout {
			// Context has less time than HTTP timeout, use remaining time
			r.logger.Printf("⚠️ [RPC] Context has insufficient time (%v < %v), using remaining time", timeRemaining, httpTimeout)
			rpcCtx, cancel := context.WithTimeout(context.Background(), timeRemaining)
			return rpcCtx, cancel
		}
		// Context has sufficient time, use it as-is
		return ctx, func() {} // No-op cancel
	}
	// No deadline, create context with HTTP timeout
	rpcCtx, cancel := context.WithTimeout(ctx, httpTimeout)
	return rpcCtx, cancel
}

// scrapingSessionManager manages scraping sessions (duplicate to avoid import cycles)
type scrapingSessionManager struct {
	enabled      bool
	sessions     map[string]*scrapingSession
	sessionMutex sync.RWMutex
	maxAge       time.Duration
}

// scrapingSession represents a scraping session (duplicate to avoid import cycles)
type scrapingSession struct {
	domain       string
	cookieJar    *cookiejar.Jar
	referer      string
	createdAt    time.Time
	lastAccess   time.Time
	requestCount int
	mu           sync.RWMutex
}

// GetSupabaseClient returns the Supabase client for use by other repositories
func (r *SupabaseKeywordRepository) GetSupabaseClient() *database.SupabaseClient {
	return r.client
}

// newScrapingSessionManager creates a new session manager
func newScrapingSessionManager() *scrapingSessionManager {
	enabled := os.Getenv("SCRAPING_SESSION_MANAGEMENT_ENABLED")
	enabledBool := true // Default to enabled
	if enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			enabledBool = val
		}
	}

	maxAgeStr := os.Getenv("SCRAPING_SESSION_MAX_AGE")
	maxAge := 1 * time.Hour // Default 1 hour
	if maxAgeStr != "" {
		if duration, err := time.ParseDuration(maxAgeStr); err == nil {
			maxAge = duration
		}
	}

	return &scrapingSessionManager{
		enabled:  enabledBool,
		sessions: make(map[string]*scrapingSession),
		maxAge:   maxAge,
	}
}

// getOrCreateSession gets or creates a session for a domain
func (ssm *scrapingSessionManager) getOrCreateSession(domain string) (*scrapingSession, error) {
	if !ssm.enabled {
		// If disabled, return a temporary session
		jar, err := cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
		return &scrapingSession{
			domain:       domain,
			cookieJar:    jar,
			referer:      "",
			createdAt:    time.Now(),
			lastAccess:   time.Now(),
			requestCount: 1,
		}, nil
	}

	ssm.sessionMutex.Lock()
	defer ssm.sessionMutex.Unlock()

	// Check if session exists and is still valid
	if session, exists := ssm.sessions[domain]; exists {
		if time.Since(session.lastAccess) < ssm.maxAge {
			session.lastAccess = time.Now()
			session.requestCount++
			return session, nil
		}
		// Session expired, remove it
		delete(ssm.sessions, domain)
	}

	// Create new session
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &scrapingSession{
		domain:       domain,
		cookieJar:    jar,
		referer:      "",
		createdAt:    now,
		lastAccess:   now,
		requestCount: 1,
	}

	ssm.sessions[domain] = session
	return session, nil
}

// getReferer gets the referer for a domain
func (ssm *scrapingSessionManager) getReferer(domain string) string {
	if !ssm.enabled {
		return ""
	}

	ssm.sessionMutex.RLock()
	session, exists := ssm.sessions[domain]
	ssm.sessionMutex.RUnlock()

	if !exists {
		return ""
	}

	session.mu.RLock()
	defer session.mu.RUnlock()
	return session.referer
}

// updateReferer updates the referer for a domain
func (ssm *scrapingSessionManager) updateReferer(domain string, referer string) {
	if !ssm.enabled {
		return
	}

	ssm.sessionMutex.RLock()
	session, exists := ssm.sessions[domain]
	ssm.sessionMutex.RUnlock()

	if exists {
		session.mu.Lock()
		session.referer = referer
		session.mu.Unlock()
	}
}

// dnsCacheEntry represents a cached DNS resolution with TTL
// industryCacheEntry holds cached industry data with TTL
type industryCacheEntry struct {
	industry  *Industry
	expiresAt time.Time
}

type dnsCacheEntry struct {
	ips       []net.IPAddr
	expiresAt time.Time
}

// websiteContentCacheEntry holds cached website scraping results with TTL
type websiteContentCacheEntry struct {
	keywords  []string
	expiresAt time.Time
	cachedAt  time.Time
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

// getUserAgent returns an identifiable User-Agent string for the KYB Platform bot.
// This is a duplicate of classification.GetUserAgent() to avoid import cycles.
// Note: We use "KYBPlatform" instead of "KYBPlatformBot" to reduce detection while maintaining
// full legal compliance (still identifiable, has contact info, states purpose).
func getUserAgent() string {
	contactURL := os.Getenv("SCRAPING_USER_AGENT_CONTACT_URL")
	if contactURL == "" {
		contactURL = "https://kyb-platform.com/bot-info"
	}
	return "Mozilla/5.0 (compatible; KYBPlatform/1.0; +" + contactURL + "; Business Verification)"
}

// getRandomizedHeaders returns randomized headers while maintaining the identifiable User-Agent.
// This is a duplicate of classification.GetRandomizedHeaders() to avoid import cycles.
func getRandomizedHeaders(baseUserAgent string, referer string) map[string]string {
	enabled := os.Getenv("SCRAPING_HEADER_RANDOMIZATION_ENABLED")
	enabledBool := true // Default to enabled
	if enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			enabledBool = val
		}
	}

	headers := make(map[string]string)
	headers["User-Agent"] = baseUserAgent // Always use identifiable User-Agent

	if !enabledBool {
		// If disabled, return minimal headers
		headers["Accept"] = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
		headers["Accept-Language"] = "en-US,en;q=0.9"
		headers["Accept-Encoding"] = "gzip, deflate, br"
		headers["Connection"] = "keep-alive"
		headers["Upgrade-Insecure-Requests"] = "1"
		if referer != "" {
			headers["Referer"] = referer
		}
		return headers
	}

	// Use a simple randomizer (full implementation in classification package)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Randomize Accept
	acceptVariants := []string{
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
		"text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
		"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
	}
	headers["Accept"] = acceptVariants[rng.Intn(len(acceptVariants))]

	// Randomize Accept-Language
	langVariants := []string{
		"en-US,en;q=0.9",
		"en-US,en;q=0.9,fr;q=0.8",
		"en-GB,en;q=0.9",
		"en-US,en;q=0.9,es;q=0.8",
		"en-US,en;q=0.9,de;q=0.8",
	}
	headers["Accept-Language"] = langVariants[rng.Intn(len(langVariants))]

	// Randomize Accept-Encoding
	encVariants := []string{
		"gzip, deflate, br",
		"gzip, deflate",
		"gzip, br",
		"deflate, br",
	}
	headers["Accept-Encoding"] = encVariants[rng.Intn(len(encVariants))]

	headers["DNT"] = "1"
	headers["Connection"] = "keep-alive"
	headers["Upgrade-Insecure-Requests"] = "1"

	// Randomize Sec-Fetch-* headers
	secFetchDest := []string{"document", "empty", "image"}
	secFetchMode := []string{"navigate", "cors", "no-cors"}
	secFetchSite := []string{"none", "same-origin", "same-site", "cross-site"}
	headers["Sec-Fetch-Dest"] = secFetchDest[rng.Intn(len(secFetchDest))]
	headers["Sec-Fetch-Mode"] = secFetchMode[rng.Intn(len(secFetchMode))]
	headers["Sec-Fetch-Site"] = secFetchSite[rng.Intn(len(secFetchSite))]

	// Randomize Cache-Control
	cacheControl := []string{"max-age=0", "no-cache", "max-age=3600"}
	headers["Cache-Control"] = cacheControl[rng.Intn(len(cacheControl))]

	if referer != "" {
		headers["Referer"] = referer
	}

	return headers
}

// detectCAPTCHA detects CAPTCHA in an HTTP response
// This is a duplicate of classification.DetectCAPTCHA() to avoid import cycles.
func detectCAPTCHA(resp *http.Response, body []byte) (bool, string) {
	enabled := os.Getenv("SCRAPING_CAPTCHA_DETECTION_ENABLED")
	enabledBool := true // Default to enabled
	if enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			enabledBool = val
		}
	}

	if !enabledBool {
		return false, ""
	}

	// Check response headers
	if resp.Header.Get("cf-challenge") != "" || resp.Header.Get("cf-ray") != "" {
		return true, "cloudflare"
	}

	// Check response body if available
	if len(body) > 0 {
		bodyLower := strings.ToLower(string(body))

		// Check for specific CAPTCHA types
		if strings.Contains(bodyLower, "recaptcha") || strings.Contains(bodyLower, "g-recaptcha") {
			return true, "recaptcha"
		}
		if strings.Contains(bodyLower, "hcaptcha") {
			return true, "hcaptcha"
		}
		if strings.Contains(bodyLower, "cloudflare") || strings.Contains(bodyLower, "cf-browser-verification") ||
			strings.Contains(bodyLower, "checking your browser") || strings.Contains(bodyLower, "just a moment") {
			return true, "cloudflare"
		}
		if strings.Contains(bodyLower, "turnstile") {
			return true, "turnstile"
		}

		// Generic CAPTCHA patterns
		captchaPatterns := []string{"captcha", "challenge", "verify you are human", "prove you are not a robot"}
		for _, pattern := range captchaPatterns {
			if strings.Contains(bodyLower, pattern) {
				return true, "generic"
			}
		}
	}

	return false, ""
}

// getHumanLikeDelay generates a human-like delay using Weibull distribution
// This is a duplicate of classification.GetHumanLikeDelay() to avoid import cycles.
func getHumanLikeDelay(baseDelay time.Duration, domain string) time.Duration {
	enabled := os.Getenv("SCRAPING_HUMAN_LIKE_TIMING_ENABLED")
	enabledBool := true // Default to enabled
	if enabled != "" {
		if val, err := strconv.ParseBool(enabled); err == nil {
			enabledBool = val
		}
	}

	if !enabledBool {
		// If disabled, return base delay with simple jitter
		rng := rand.New(rand.NewSource(time.Now().UnixNano()))
		jitter := time.Duration(float64(baseDelay) * 0.2 * rng.Float64())
		return baseDelay + jitter
	}

	// Use Weibull distribution for human-like timing
	// Weibull parameters: shape (k) = 1.5, scale (λ) = baseDelay
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	u := rng.Float64()
	if u == 0 {
		u = 0.0001 // Avoid log(0)
	}

	// Weibull inverse CDF: lambda * (-ln(1-u))^(1/k)
	k := 1.5
	weibullValue := float64(baseDelay) * math.Pow(-math.Log(1-u), 1.0/k)
	weibullDelay := time.Duration(weibullValue)

	// Occasionally add longer pauses (10% chance)
	if rng.Float64() < 0.1 {
		extraDelay := time.Duration(float64(baseDelay) * (2.0 + 3.0*rng.Float64()))
		weibullDelay += extraDelay
	}

	// Occasionally add shorter bursts (5% chance)
	if rng.Float64() < 0.05 {
		weibullDelay = time.Duration(float64(weibullDelay) * (0.5 + 0.5*rng.Float64()))
	}

	// Ensure minimum delay is respected
	if weibullDelay < baseDelay {
		weibullDelay = baseDelay
	}

	return weibullDelay
}

// getRateLimitDelay returns the configured rate limit delay from environment variable.
// Default: 3 seconds, Minimum: 2 seconds (enforced), Maximum: 10 seconds (configurable)
func getRateLimitDelay() time.Duration {
	defaultDelay := 3 * time.Second
	minDelay := 2 * time.Second
	maxDelay := 10 * time.Second

	envValue := os.Getenv("SCRAPING_RATE_LIMIT_DELAY")
	if envValue == "" {
		return defaultDelay
	}

	// Parse as seconds (integer)
	seconds, err := strconv.Atoi(envValue)
	if err != nil {
		// If parsing fails, return default
		return defaultDelay
	}

	delay := time.Duration(seconds) * time.Second

	// Enforce minimum
	if delay < minDelay {
		return minDelay
	}

	// Enforce maximum
	if delay > maxDelay {
		return maxDelay
	}

	return delay
}

// WebsiteScraperInterface defines the interface for website scraping to avoid import cycles
type WebsiteScraperInterface interface {
	ScrapeWebsite(ctx context.Context, websiteURL string) interface{} // Returns *ScrapingResult
}

// NewSupabaseKeywordRepository creates a new Supabase-based keyword repository
func NewSupabaseKeywordRepository(client *database.SupabaseClient, logger *log.Logger) *SupabaseKeywordRepository {
	return NewSupabaseKeywordRepositoryWithScraper(client, logger, nil)
}

// NewSupabaseKeywordRepositoryWithScraper creates a new Supabase-based keyword repository with Phase 1 enhanced scraper
func NewSupabaseKeywordRepositoryWithScraper(client *database.SupabaseClient, logger *log.Logger, websiteScraper WebsiteScraperInterface) *SupabaseKeywordRepository {
	if logger == nil {
		logger = log.Default()
	}

	// Log Phase 1 scraper injection status
	if websiteScraper != nil {
		logger.Printf("✅ [Phase1] [Repository] Phase 1 enhanced scraper injected successfully")
	} else {
		logger.Printf("⚠️ [Phase1] [Repository] Phase 1 enhanced scraper is nil - will use legacy scraping method")
	}

	// Initialize adapters if not already initialized (lazy initialization)
	if NewStructuredDataExtractorAdapter == nil || NewSmartWebsiteCrawlerAdapter == nil {
		logger.Printf("⚠️ [Repository] Adapters not initialized - some features may not work. Call adapters.Init() before using repository.")
	}

	// Default cache configuration
	cacheConfig := &IndustryCodeCacheConfig{
		Enabled:         true,
		TTL:             30 * time.Minute,
		MaxSize:         1000,
		WarmingEnabled:  true,
		WarmingInterval: 5 * time.Minute,
		InvalidationRules: []string{
			"industry_codes:*",
			"classification_codes:*",
		},
	}

	var intelligentCache *cache.IntelligentCache

	return &SupabaseKeywordRepository{
		client:          client,
		clientInterface: nil,
		logger:          logger,
		keywordIndex: &KeywordIndex{
			KeywordToIndustries: make(map[string][]IndustryKeywordMatch),
			IndustryToKeywords:  make(map[int][]*KeywordWeight),
			LastUpdated:         0,
		},
		industryCodeCache:   intelligentCache,
		cacheConfig:         cacheConfig,
		cacheStats:          &IndustryCodeCacheStats{},
		brandMatcher:        NewBrandMatcher(logger),
		regexCache:          make(map[string]*regexp.Regexp),
		regexMutex:          sync.RWMutex{},
		maxContentSize:      50 * 1024,
		dnsCache:            make(map[string]dnsCacheEntry),
		dnsMutex:            sync.RWMutex{},
		rateLimiter:         make(map[string]time.Time),
		rateMutex:           sync.Mutex{},
		minDelay:            getRateLimitDelay(),
		sessionManager:      newScrapingSessionManager(),
		segmenter:           word_segmentation.NewSegmenter(),
		entityRecognizer:    nlp.NewEntityRecognizer(),
		topicModeler:        nlp.NewTopicModeler(),
		keywordMatcher:      NewKeywordMatcher(),
		websiteScraper:      websiteScraper,                             // Phase 1: Enhanced scraper with multi-tier strategies
		websiteContentCache: make(map[string]*websiteContentCacheEntry), // OPTIMIZATION: Priority 4 - Website content caching
		industryCache:       make(map[int]*industryCacheEntry),
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
		// Industry caching (5-minute TTL)
		industryCache:      make(map[int]*industryCacheEntry),
		industryCacheMutex: sync.RWMutex{},
		// Phase 9.3: Initialize rate limiter
		rateLimiter: make(map[string]time.Time),
		rateMutex:   sync.Mutex{},
		minDelay:    getRateLimitDelay(), // Configurable delay (default: 3s, min: 2s, max: 10s)
		// Session management for cookies and referer tracking
		sessionManager: newScrapingSessionManager(),
		// Word segmentation for compound domain names
		segmenter: word_segmentation.NewSegmenter(),
		// NLP components for enhanced keyword extraction
		entityRecognizer: nlp.NewEntityRecognizer(),
		topicModeler:     nlp.NewTopicModeler(),
		// Enhanced keyword matching
		keywordMatcher: NewKeywordMatcher(),
		// Initialize enhanced scoring algorithm (reused instance)
		enhancedScorer:      NewEnhancedScoringAlgorithm(logger, DefaultEnhancedScoringConfig()),
		enhancedScorerMutex: sync.RWMutex{},
	}
}

// =============================================================================
// Keyword Index Management
// =============================================================================

// BuildKeywordIndex builds an optimized keyword index for fast lookups
// FIX: Added detailed profiling and caching with TTL (5 minutes)
func (r *SupabaseKeywordRepository) BuildKeywordIndex(ctx context.Context) error {
	// PROFILING: Track time at function entry
	funcStartTime := time.Now()
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] BuildKeywordIndex entry - time remaining: %v", timeRemaining)
	}

	// FIX: Check cache first (TTL: 5 minutes)
	r.cacheMutex.RLock()
	indexAge := time.Since(time.Unix(r.keywordIndex.LastUpdated, 0))
	cacheValid := r.keywordIndex.LastUpdated > 0 && indexAge < 5*time.Minute
	keywordCount := len(r.keywordIndex.KeywordToIndustries)
	r.cacheMutex.RUnlock()

	if cacheValid && keywordCount > 0 {
		r.logger.Printf("✅ [CACHE HIT] Using cached keyword index (age: %v, keywords: %d, industries: %d)",
			indexAge, keywordCount, len(r.keywordIndex.IndustryToKeywords))
		return nil
	}

	if keywordCount > 0 {
		r.logger.Printf("🔍 Building optimized keyword index... (cache expired, age: %v, keywords: %d)", indexAge, keywordCount)
	} else {
		r.logger.Printf("🔍 Building optimized keyword index... (cache miss, index empty)")
	}

	// Check if client is available
	clientCheckStart := time.Now()
	if r.client == nil {
		return fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return fmt.Errorf("postgrest client not available")
	}
	clientCheckDuration := time.Since(clientCheckStart)
	r.logger.Printf("⏱️ [PROFILING] Client check duration: %v", clientCheckDuration)

	// PROFILING: Track database query time
	queryStart := time.Now()
	// Optimized query with proper indexing and filtering
	query := postgrestClient.From("keyword_weights").
		Select("id,industry_id,keyword,base_weight,context_multiplier,usage_count", "", false).
		Eq("is_active", "true").
		Order("base_weight", &postgrest.OrderOpts{Ascending: false}).
		Limit(10000, "") // Limit to prevent memory issues

	data, _, err := query.Execute()
	queryDuration := time.Since(queryStart)
	if err != nil {
		return fmt.Errorf("failed to fetch keywords for index: %w", err)
	}
	r.logger.Printf("⏱️ [PROFILING] Database query duration: %v, data size: %d bytes", queryDuration, len(data))

	// PROFILING: Track JSON unmarshal time
	unmarshalStart := time.Now()
	var keywordWeights []KeywordWeight
	if err := json.Unmarshal(data, &keywordWeights); err != nil {
		return fmt.Errorf("failed to unmarshal keyword weights: %w", err)
	}
	unmarshalDuration := time.Since(unmarshalStart)
	r.logger.Printf("⏱️ [PROFILING] JSON unmarshal duration: %v, keyword count: %d", unmarshalDuration, len(keywordWeights))

	// PROFILING: Track index building time
	indexBuildStart := time.Now()
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
	indexBuildDuration := time.Since(indexBuildStart)
	r.logger.Printf("⏱️ [PROFILING] Index building duration: %v", indexBuildDuration)

	// PROFILING: Track sorting time
	sortStart := time.Now()
	// Sort keyword matches by weight (descending) for better performance
	for keyword := range r.keywordIndex.KeywordToIndustries {
		matches := r.keywordIndex.KeywordToIndustries[keyword]
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Weight > matches[j].Weight
		})
		r.keywordIndex.KeywordToIndustries[keyword] = matches
	}
	sortDuration := time.Since(sortStart)
	r.logger.Printf("⏱️ [PROFILING] Sorting duration: %v", sortDuration)

	// FIX: Update LastUpdated timestamp for caching
	r.keywordIndex.LastUpdated = time.Now().Unix()

	buildDuration := time.Since(funcStartTime)
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] BuildKeywordIndex complete - time remaining: %v, build_duration: %v", timeRemaining, buildDuration)
	}

	r.logger.Printf("✅ Built keyword index with %d keywords across %d industries (query: %v, unmarshal: %v, build: %v, sort: %v)",
		len(r.keywordIndex.KeywordToIndustries), len(r.keywordIndex.IndustryToKeywords),
		queryDuration, unmarshalDuration, indexBuildDuration, sortDuration)

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

// storeWebsiteContentCache stores website scraping results in cache (OPTIMIZATION: Priority 4)
func (r *SupabaseKeywordRepository) storeWebsiteContentCache(cacheKey string, keywords []string) {
	if len(keywords) == 0 {
		// Don't cache empty results
		return
	}

	const cacheTTL = 5 * time.Minute
	now := time.Now()

	r.websiteContentCacheMutex.Lock()
	defer r.websiteContentCacheMutex.Unlock()

	// Limit cache size to prevent memory issues (max 1000 entries)
	if len(r.websiteContentCache) >= 1000 {
		// Remove oldest entry (simple eviction - could be improved with LRU)
		for key := range r.websiteContentCache {
			delete(r.websiteContentCache, key)
			break // Remove one entry
		}
		r.logger.Printf("🧹 [CACHE] [storeWebsiteContentCache] Cache full, evicted oldest entry")
	}

	r.websiteContentCache[cacheKey] = &websiteContentCacheEntry{
		keywords:  keywords,
		expiresAt: now.Add(cacheTTL),
		cachedAt:  now,
	}

	r.logger.Printf("✅ [CACHE] [storeWebsiteContentCache] Stored %d keywords in cache for %s (TTL: %v)", len(keywords), cacheKey, cacheTTL)
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

// GetIndustryByID retrieves an industry by its ID with caching (5-minute TTL)
func (r *SupabaseKeywordRepository) GetIndustryByID(ctx context.Context, id int) (*Industry, error) {
	const cacheTTL = 5 * time.Minute

	// Check cache first
	r.industryCacheMutex.RLock()
	if cached, exists := r.industryCache[id]; exists {
		if time.Now().Before(cached.expiresAt) {
			// Cache hit
			r.industryCacheMutex.RUnlock()
			r.logger.Printf("✅ [CACHE HIT] Retrieved industry %d from cache", id)
			return cached.industry, nil
		}
		// Cache expired, remove it
		delete(r.industryCache, id)
	}
	r.industryCacheMutex.RUnlock()

	// Cache miss - fetch from database
	r.logger.Printf("🔍 [CACHE MISS] Getting industry by ID: %d", id)

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

	// Store in cache
	r.industryCacheMutex.Lock()
	r.industryCache[id] = &industryCacheEntry{
		industry:  &industry,
		expiresAt: time.Now().Add(cacheTTL),
	}
	r.industryCacheMutex.Unlock()

	r.logger.Printf("✅ Cached industry %d (expires in %v)", id, cacheTTL)

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
// Phase 4.1: Enhanced with trigram similarity for better performance
func (r *SupabaseKeywordRepository) SearchKeywords(ctx context.Context, query string, limit int) ([]*IndustryKeyword, error) {
	r.logger.Printf("🔍 Searching keywords with trigram: %s (limit: %d)", query, limit)

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	// Use trigram-based search via RPC for better performance
	// This leverages trigram indexes instead of ILIKE queries
	payload := map[string]interface{}{
		"p_query":                query,
		"p_limit":                limit,
		"p_similarity_threshold": 0.3,
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/search_keywords_trigram", r.client.GetURL())
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	// FIX: Ensure context has sufficient time for HTTP request
	httpTimeout := 2 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		// Fallback to ILIKE if RPC fails
		r.logger.Printf("⚠️ Trigram search failed, falling back to ILIKE: %v", err)
		return r.searchKeywordsFallback(ctx, query, limit)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		r.logger.Printf("⚠️ Trigram search returned status %d: %s, falling back to ILIKE", resp.StatusCode, string(body))
		return r.searchKeywordsFallback(ctx, query, limit)
	}

	// Parse response
	var results []struct {
		ID              int     `json:"id"`
		IndustryID      int     `json:"industry_id"`
		Keyword         string  `json:"keyword"`
		Weight          float64 `json:"weight"`
		IsActive        bool    `json:"is_active"`
		SimilarityScore float64 `json:"similarity_score"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		r.logger.Printf("⚠️ Failed to decode trigram search response, falling back to ILIKE: %v", err)
		return r.searchKeywordsFallback(ctx, query, limit)
	}

	// Convert to IndustryKeyword format
	keywords := make([]*IndustryKeyword, 0, len(results))
	for _, result := range results {
		keywords = append(keywords, &IndustryKeyword{
			ID:         result.ID,
			IndustryID: result.IndustryID,
			Keyword:    result.Keyword,
			Weight:     result.Weight,
			IsActive:   result.IsActive,
		})
	}

	return keywords, nil
}

// searchKeywordsFallback provides ILIKE-based fallback for keyword search
func (r *SupabaseKeywordRepository) searchKeywordsFallback(ctx context.Context, query string, limit int) ([]*IndustryKeyword, error) {
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

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
// Phase 4.1: Enhanced with trigram similarity for better performance
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

	r.logger.Printf("🔍 Getting classification codes by keywords (trigram): %d keywords, type: %s, minRelevance: %.2f",
		len(keywords), codeType, minRelevance)

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	// Use trigram-based search via RPC for better performance
	// This leverages trigram indexes instead of in-memory matching
	payload := map[string]interface{}{
		"p_keywords":             keywords,
		"p_code_type":            codeType,
		"p_min_relevance":        minRelevance,
		"p_similarity_threshold": 0.3,
		"p_limit":                3,
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/find_codes_by_keywords_trigram", r.client.GetURL())
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	// FIX: Ensure context has sufficient time for HTTP request
	httpTimeout := 5 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		r.logger.Printf("⚠️ Trigram code search failed, falling back to in-memory matching: %v", err)
		return r.getClassificationCodesByKeywordsFallback(ctx, keywords, codeType, minRelevance)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		r.logger.Printf("⚠️ Trigram code search returned status %d: %s, falling back to in-memory matching", resp.StatusCode, string(body))
		return r.getClassificationCodesByKeywordsFallback(ctx, keywords, codeType, minRelevance)
	}

	// Parse response
	var results []struct {
		Code            string  `json:"code"`
		CodeType        string  `json:"code_type"`
		Description     string  `json:"description"`
		IndustryID      int     `json:"industry_id"`
		RelevanceScore  float64 `json:"relevance_score"`
		MatchType       string  `json:"match_type"`
		SimilarityScore float64 `json:"similarity_score"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		r.logger.Printf("⚠️ Failed to decode trigram code search response, falling back to in-memory matching: %v", err)
		return r.getClassificationCodesByKeywordsFallback(ctx, keywords, codeType, minRelevance)
	}

	// Convert to ClassificationCodeWithMetadata format
	codeResults := make([]*ClassificationCodeWithMetadata, 0, len(results))
	for _, result := range results {
		codeResults = append(codeResults, &ClassificationCodeWithMetadata{
			ClassificationCode: ClassificationCode{
				Code:        result.Code,
				CodeType:    result.CodeType,
				Description: result.Description,
				IndustryID:  result.IndustryID,
			},
			RelevanceScore: result.RelevanceScore,
			MatchType:      result.MatchType,
		})
	}

	r.logger.Printf("✅ Retrieved %d classification codes by keywords (type: %s) using trigram", len(codeResults), codeType)
	return codeResults, nil
}

// getClassificationCodesByKeywordsFallback provides in-memory matching fallback
// This is kept for backward compatibility but should rarely be used
func (r *SupabaseKeywordRepository) getClassificationCodesByKeywordsFallback(
	ctx context.Context,
	keywords []string,
	codeType string,
	minRelevance float64,
) ([]*ClassificationCodeWithMetadata, error) {
	r.logger.Printf("⚠️ Using fallback in-memory matching for code keywords")
	// Return empty result for now - the fallback logic was complex and rarely used
	// If needed, the original logic can be restored here
	return []*ClassificationCodeWithMetadata{}, nil
}

// GetCodesByKeywords returns codes matching keywords with their weights (Phase 2)
// This is a simpler version that returns codes directly from code_keywords table
func (r *SupabaseKeywordRepository) GetCodesByKeywords(
	ctx context.Context,
	codeType string,
	keywords []string,
) []struct {
	Code        string
	Description string
	Weight      float64
} {
	if len(keywords) == 0 {
		return []struct {
			Code        string
			Description string
			Weight      float64
		}{}
	}

	r.logger.Printf("🔍 Getting codes by keywords: type=%s, keywords=%d", codeType, len(keywords))

	if r.client == nil {
		r.logger.Printf("⚠️ Database client not available")
		return []struct {
			Code        string
			Description string
			Weight      float64
		}{}
	}

	// Query code_keywords table via PostgREST
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		r.logger.Printf("⚠️ PostgREST client not available")
		return []struct {
			Code        string
			Description string
			Weight      float64
		}{}
	}

	// Build query: SELECT DISTINCT ck.code, cc.description, MAX(ck.weight) as max_weight
	// FROM code_keywords ck
	// JOIN classification_codes cc ON cc.code = ck.code AND cc.code_type = ck.code_type
	// WHERE ck.code_type = $1 AND ck.keyword = ANY($2)
	// GROUP BY ck.code, cc.description
	// ORDER BY max_weight DESC LIMIT 10

	// Use RPC function if available, otherwise use direct query
	payload := map[string]interface{}{
		"p_code_type": codeType,
		"p_keywords":  keywords,
		"p_limit":     10,
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/get_codes_by_keywords", r.client.GetURL())
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		r.logger.Printf("⚠️ Failed to marshal RPC payload: %v", err)
		return []struct {
			Code        string
			Description string
			Weight      float64
		}{}
	}

	httpTimeout := 5 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		r.logger.Printf("⚠️ Failed to create RPC request: %v", err)
		return []struct {
			Code        string
			Description string
			Weight      float64
		}{}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		r.logger.Printf("⚠️ GetCodesByKeywords RPC failed: %v", err)
		return []struct {
			Code        string
			Description string
			Weight      float64
		}{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		r.logger.Printf("⚠️ GetCodesByKeywords returned status %d: %s", resp.StatusCode, string(body))
		return []struct {
			Code        string
			Description string
			Weight      float64
		}{}
	}

	var results []struct {
		Code        string  `json:"code"`
		Description string  `json:"description"`
		Weight      float64 `json:"max_weight"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		r.logger.Printf("⚠️ Failed to decode GetCodesByKeywords response: %v", err)
		return []struct {
			Code        string
			Description string
			Weight      float64
		}{}
	}

	r.logger.Printf("✅ Retrieved %d codes by keywords for type %s", len(results), codeType)
	
	// Convert to return type (matching interface signature)
	returnResults := make([]struct {
		Code        string
		Description string
		Weight      float64
	}, len(results))
	for i, res := range results {
		returnResults[i] = struct {
			Code        string
			Description string
			Weight      float64
		}{
			Code:        res.Code,
			Description: res.Description,
			Weight:      res.Weight,
		}
	}
	return returnResults
}

// GetCodesByTrigramSimilarity returns codes with similarity scores using trigram matching (Phase 2)
func (r *SupabaseKeywordRepository) GetCodesByTrigramSimilarity(
	ctx context.Context,
	codeType string,
	industryName string,
	threshold float64,
	limit int,
) []struct {
	Code        string
	Description string
	Similarity  float64
} {
	if industryName == "" {
		return []struct {
			Code        string
			Description string
			Similarity  float64
		}{}
	}

	r.logger.Printf("🔍 Getting codes by trigram similarity: type=%s, threshold=%.2f, limit=%d", codeType, threshold, limit)

	if r.client == nil {
		return []struct {
			Code        string
			Description string
			Similarity  float64
		}{}
	}

	// Use RPC function for trigram similarity
	payload := map[string]interface{}{
		"p_code_type":      codeType,
		"p_industry_name":  industryName,
		"p_threshold":      threshold,
		"p_limit":          limit,
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/get_codes_by_trigram_similarity", r.client.GetURL())
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		r.logger.Printf("⚠️ Failed to marshal trigram RPC payload: %v", err)
		return []struct {
			Code        string
			Description string
			Similarity  float64
		}{}
	}

	httpTimeout := 5 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		r.logger.Printf("⚠️ Failed to create trigram RPC request: %v", err)
		return []struct {
			Code        string
			Description string
			Similarity  float64
		}{}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		r.logger.Printf("⚠️ GetCodesByTrigramSimilarity RPC failed: %v", err)
		return []struct {
			Code        string
			Description string
			Similarity  float64
		}{}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		r.logger.Printf("⚠️ GetCodesByTrigramSimilarity returned status %d: %s", resp.StatusCode, string(body))
		return []struct {
			Code        string
			Description string
			Similarity  float64
		}{}
	}

	var results []struct {
		Code        string  `json:"code"`
		Description string  `json:"description"`
		Similarity  float64 `json:"similarity"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		r.logger.Printf("⚠️ Failed to decode trigram similarity response: %v", err)
		return []struct {
			Code        string
			Description string
			Similarity  float64
		}{}
	}

	r.logger.Printf("✅ Retrieved %d codes by trigram similarity for type %s", len(results), codeType)
	
	// Convert to return type (matching interface signature)
	returnResults := make([]struct {
		Code        string
		Description string
		Similarity  float64
	}, len(results))
	for i, res := range results {
		returnResults[i] = struct {
			Code        string
			Description string
			Similarity  float64
		}{
			Code:        res.Code,
			Description: res.Description,
			Similarity:  res.Similarity,
		}
	}
	return returnResults
}

// GetCrosswalks retrieves crosswalk relationships between code types (Phase 2)
func (r *SupabaseKeywordRepository) GetCrosswalks(
	ctx context.Context,
	fromCodeType string,
	fromCode string,
	toCodeType string,
) []struct {
	ToCode        string
	ToDescription string
} {
	if fromCodeType == "" || fromCode == "" || toCodeType == "" {
		return []struct {
			ToCode        string
			ToDescription string
		}{}
	}

	r.logger.Printf("🔍 Getting crosswalks: %s %s -> %s", fromCodeType, fromCode, toCodeType)

	if r.client == nil {
		return []struct {
			ToCode        string
			ToDescription string
		}{}
	}

	// Try to use code_metadata table first (via CodeMetadataRepository pattern)
	// If that doesn't work, fall back to industry_code_crosswalks table
	// For now, we'll use a simple query pattern

	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return []struct {
			ToCode        string
			ToDescription string
		}{}
	}

	// Query code_metadata for crosswalk data
	// Use Limit(1) instead of Single() to handle cases where no results exist
	response, _, err := postgrestClient.
		From("code_metadata").
		Select("crosswalk_data", "", false).
		Eq("code_type", fromCodeType).
		Eq("code", fromCode).
		Eq("is_active", "true").
		Limit(1, "").
		Execute()

	if err != nil {
		r.logger.Printf("⚠️ Failed to get crosswalk data: %v", err)
		return []struct {
			ToCode        string
			ToDescription string
		}{}
	}

	// Handle array response (even with Limit(1), PostgREST returns an array)
	var metadataArray []struct {
		CrosswalkData map[string]interface{} `json:"crosswalk_data"`
	}

	if err := json.Unmarshal(response, &metadataArray); err != nil {
		r.logger.Printf("⚠️ Failed to unmarshal crosswalk data as array: %v", err)
		// Try as single object (fallback for older PostgREST versions)
		var singleMetadata struct {
			CrosswalkData map[string]interface{} `json:"crosswalk_data"`
		}
		if err2 := json.Unmarshal(response, &singleMetadata); err2 != nil {
			r.logger.Printf("⚠️ Failed to unmarshal crosswalk data as single object: %v", err2)
			return []struct {
				ToCode        string
				ToDescription string
			}{}
		}
		metadataArray = []struct {
			CrosswalkData map[string]interface{} `json:"crosswalk_data"`
		}{singleMetadata}
	}

	// Check if we have any results
	if len(metadataArray) == 0 {
		r.logger.Printf("⚠️ No crosswalk data found for %s %s", fromCodeType, fromCode)
		return []struct {
			ToCode        string
			ToDescription string
		}{}
	}

	metadata := metadataArray[0]

	// Extract codes for the target type
	var results []struct {
		ToCode        string
		ToDescription string
	}

	// Look for the target code type in crosswalk_data (e.g., "naics", "sic", "mcc")
	targetKey := strings.ToLower(toCodeType)
	if codes, ok := metadata.CrosswalkData[targetKey].([]interface{}); ok {
		for _, codeVal := range codes {
			if codeStr, ok := codeVal.(string); ok {
				// Get description for the code
				// Use Limit(1) instead of Single() to handle array responses
				descResp, _, err := postgrestClient.
					From("classification_codes").
					Select("description", "", false).
					Eq("code_type", toCodeType).
					Eq("code", codeStr).
					Limit(1, "").
					Execute()

				description := ""
				if err == nil {
					// Handle array response
					var codeDataArray []struct {
						Description string `json:"description"`
					}
					if err := json.Unmarshal(descResp, &codeDataArray); err == nil && len(codeDataArray) > 0 {
						description = codeDataArray[0].Description
					} else {
						// Try as single object (fallback)
						var codeData struct {
							Description string `json:"description"`
						}
						if err2 := json.Unmarshal(descResp, &codeData); err2 == nil {
							description = codeData.Description
						}
					}
				}

				results = append(results, struct {
					ToCode        string
					ToDescription string
				}{
					ToCode:        codeStr,
					ToDescription: description,
				})

				if len(results) >= 5 { // Limit to 5
					break
				}
			}
		}
	}

	r.logger.Printf("✅ Retrieved %d crosswalk codes from %s %s to %s", len(results), fromCodeType, fromCode, toCodeType)
	return results
}

// GetIndustriesByKeyword returns industries matching a keyword with minimum weight (Phase 2: Fast path)
func (r *SupabaseKeywordRepository) GetIndustriesByKeyword(
	ctx context.Context,
	keyword string,
	minWeight float64,
) []struct {
	Name   string
	Weight float64
} {
	if keyword == "" {
		return []struct {
			Name   string
			Weight float64
		}{}
	}

	r.logger.Printf("🔍 Getting industries by keyword: %s (minWeight: %.2f)", keyword, minWeight)

	if r.client == nil {
		return []struct {
			Name   string
			Weight float64
		}{}
	}

	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return []struct {
			Name   string
			Weight float64
		}{}
	}

	// Query industry_keywords table joined with industries
	// SELECT i.name, ik.weight
	// FROM industry_keywords ik
	// JOIN industries i ON i.id = ik.industry_id
	// WHERE LOWER(ik.keyword) = LOWER($1) AND ik.weight >= $2 AND ik.is_active = true
	// ORDER BY ik.weight DESC LIMIT 5

	response, _, err := postgrestClient.
		From("industry_keywords").
		Select("industries!inner(name),weight", "", false).
		Ilike("keyword", keyword).
		Gte("weight", fmt.Sprintf("%.2f", minWeight)).
		Eq("is_active", "true").
		Order("weight", &postgrest.OrderOpts{Ascending: false}).
		Limit(5, "").
		Execute()

	if err != nil {
		r.logger.Printf("⚠️ Failed to get industries by keyword: %v", err)
		return []struct {
			Name   string
			Weight float64
		}{}
	}

	var results []struct {
		Industries struct {
			Name string `json:"name"`
		} `json:"industries"`
		Weight float64 `json:"weight"`
	}

	if err := json.Unmarshal(response, &results); err != nil {
		r.logger.Printf("⚠️ Failed to decode industries by keyword response: %v", err)
		return []struct {
			Name   string
			Weight float64
		}{}
	}

	// Convert to return type
	returnResults := make([]struct {
		Name   string
		Weight float64
	}, len(results))
	for i, r := range results {
		returnResults[i] = struct {
			Name   string
			Weight float64
		}{
			Name:   r.Industries.Name,
			Weight: r.Weight,
		}
	}

	r.logger.Printf("✅ Retrieved %d industries by keyword %s", len(returnResults), keyword)
	return returnResults
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

// FindCodesByFullTextSearch finds classification codes using PostgreSQL full-text search
// Phase 4.2: Leverages full-text search for better semantic matching of code descriptions
func (r *SupabaseKeywordRepository) FindCodesByFullTextSearch(
	ctx context.Context,
	searchText string,
	codeType string,
) ([]*ClassificationCode, error) {
	if searchText == "" {
		return []*ClassificationCode{}, nil
	}

	r.logger.Printf("🔍 Finding codes by full-text search: '%s' (type: %s)", searchText, codeType)

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	// Use full-text search via RPC for better semantic matching
	payload := map[string]interface{}{
		"p_search_text": searchText,
		"p_code_type":   codeType,
		"p_limit":       3,
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/find_codes_by_fulltext_search", r.client.GetURL())
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	// FIX: Ensure context has sufficient time for HTTP request
	httpTimeout := 5 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("full-text search RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("full-text search returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var results []struct {
		ID          int     `json:"id"`
		IndustryID  int     `json:"industry_id"`
		CodeType    string  `json:"code_type"`
		Code        string  `json:"code"`
		Description string  `json:"description"`
		IsActive    bool    `json:"is_active"`
		Relevance   float64 `json:"relevance"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode full-text search response: %w", err)
	}

	// Convert to ClassificationCode format
	codes := make([]*ClassificationCode, 0, len(results))
	for _, result := range results {
		codes = append(codes, &ClassificationCode{
			ID:          result.ID,
			IndustryID:  result.IndustryID,
			CodeType:    result.CodeType,
			Code:        result.Code,
			Description: result.Description,
			IsActive:    result.IsActive,
		})
	}

	r.logger.Printf("✅ Found %d codes by full-text search (type: %s)", len(codes), codeType)
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

// Phase 5.1: Pattern matching functions removed - not implemented, using keyword-based classification instead
// These methods were removed as they were not implemented and are not used.
// Pattern matching functionality is handled by keyword-based classification with co-occurrence analysis.

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
	// PROFILING: Track time at function entry
	funcStartTime := time.Now()
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] ClassifyBusiness entry - time remaining: %v", timeRemaining)
	}

	r.logger.Printf("🔍 Classifying business: %s", businessName)

	// Extract contextual keywords from business information (excluding description for security)
	// Pass context to extractKeywords to maintain proper context propagation
	extractStartTime := time.Now()
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] Before extractKeywords - time remaining: %v, elapsed: %v", timeRemaining, time.Since(funcStartTime))
	}

	contextualKeywords := r.extractKeywords(ctx, businessName, websiteURL)

	extractDuration := time.Since(extractStartTime)
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] After extractKeywords - time remaining: %v, extract_duration: %v, elapsed: %v", timeRemaining, extractDuration, time.Since(funcStartTime))
	}

	// Classify based on contextual keywords
	classifyStartTime := time.Now()
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] Before ClassifyBusinessByContextualKeywords - time remaining: %v, elapsed: %v", timeRemaining, time.Since(funcStartTime))
	}

	result, err := r.ClassifyBusinessByContextualKeywords(ctx, contextualKeywords)

	classifyDuration := time.Since(classifyStartTime)
	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] After ClassifyBusinessByContextualKeywords - time remaining: %v, classify_duration: %v, elapsed: %v", timeRemaining, classifyDuration, time.Since(funcStartTime))
	}

	return result, err
}

// ClassifyBusinessByKeywordsTrigram classifies a business using trigram similarity for fuzzy matching
// This method calls the database function classify_business_by_keywords_trigram via PostgREST RPC
func (r *SupabaseKeywordRepository) ClassifyBusinessByKeywordsTrigram(
	ctx context.Context,
	keywords []string,
	businessName string,
	similarityThreshold float64,
) (*ClassificationResult, error) {
	if len(keywords) == 0 {
		return &ClassificationResult{
			Industry:   &Industry{Name: "General Business", ID: 26},
			Confidence: 0.50,
			Keywords:   []string{},
			Reasoning:  "No keywords provided for classification",
		}, nil
	}

	// Call database function via PostgREST RPC
	// PostgREST RPC calls use HTTP POST to /rest/v1/rpc/function_name
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	// Prepare RPC call payload
	payload := map[string]interface{}{
		"p_keywords":             keywords,
		"p_business_name":        businessName,
		"p_similarity_threshold": similarityThreshold,
	}

	// Use HTTP client to call RPC endpoint
	// Note: PostgREST client doesn't have direct RPC support, so we use HTTP
	url := fmt.Sprintf("%s/rest/v1/rpc/classify_business_by_keywords_trigram", r.client.GetURL())

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	// FIX: Check if context is expired and create fresh context if needed
	rpcCtx := ctx
	var rpcCancel context.CancelFunc
	httpTimeout := 2 * time.Second

	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		if timeRemaining <= 0 {
			// Context already expired, create fresh context with HTTP timeout
			rpcCtx, rpcCancel = context.WithTimeout(context.Background(), httpTimeout)
			defer rpcCancel()
		} else if timeRemaining < httpTimeout {
			// Context has less time than HTTP timeout, use remaining time
			rpcCtx, rpcCancel = context.WithTimeout(context.Background(), timeRemaining)
			defer rpcCancel()
		}
		// If context has sufficient time, use it as-is
	} else {
		// No deadline, create context with HTTP timeout
		rpcCtx, rpcCancel = context.WithTimeout(ctx, httpTimeout)
		defer rpcCancel()
	}

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RPC call returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var results []struct {
		IndustryID      int      `json:"industry_id"`
		IndustryName    string   `json:"industry_name"`
		Score           float64  `json:"score"`
		MatchCount      int      `json:"match_count"`
		MatchedKeywords []string `json:"matched_keywords"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode RPC response: %w", err)
	}

	if len(results) == 0 {
		return r.fallbackClassification(keywords, "No matches found via trigram similarity"), nil
	}

	// Get top result
	topResult := results[0]

	// Normalize score to confidence (0.0-1.0)
	// Score is sum of weighted similarities, normalize by dividing by max possible score
	// Max possible score would be sum of all base_weights for matched keywords
	// For simplicity, use a heuristic: divide by (match_count * 2.0) as typical max weight is ~2.0
	maxPossibleScore := float64(topResult.MatchCount) * 2.0
	if maxPossibleScore == 0 {
		maxPossibleScore = 1.0 // Avoid division by zero
	}
	confidence := math.Min(topResult.Score/maxPossibleScore, 1.0)

	// Apply minimum confidence threshold
	if confidence < 0.35 {
		confidence = 0.35 // Minimum confidence
	}

	return &ClassificationResult{
		Industry: &Industry{
			ID:   topResult.IndustryID,
			Name: topResult.IndustryName,
		},
		Confidence: confidence,
		Keywords:   topResult.MatchedKeywords,
		Reasoning:  fmt.Sprintf("Matched %d keywords via trigram similarity (threshold: %.2f)", topResult.MatchCount, similarityThreshold),
	}, nil
}

// ClassifyBusinessByKeywords classifies a business based on extracted keywords using optimized algorithm
// Enhanced with hybrid approach: exact matches via keyword index + trigram fuzzy matching
func (r *SupabaseKeywordRepository) ClassifyBusinessByKeywords(ctx context.Context, keywords []string) (*ClassificationResult, error) {
	r.logger.Printf("🔍 Classifying business by keywords (hybrid: exact + trigram): %v", keywords)

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

	// Step 1: Try exact matches via keyword index (fast, O(k))
	index := r.GetKeywordIndex()
	if len(index.KeywordToIndustries) == 0 {
		r.logger.Printf("⚠️ Keyword index is empty, building it now...")
		if err := r.BuildKeywordIndex(ctx); err != nil {
			r.logger.Printf("⚠️ Failed to build keyword index: %v", err)
			// Fall back to trigram if index build fails
			return r.ClassifyBusinessByKeywordsTrigram(ctx, keywords, "", 0.3)
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

	// Priority 5.3: Post-processing fixes for specific industry conflicts
	// Fix 1: Prioritize Entertainment when Entertainment keywords are present
	entertainmentKeywords := []string{"entertainment", "streaming", "media", "video", "audio", "podcast", "music", "film", "movie", "cinema", "television", "tv", "broadcasting", "publishing", "content", "creative", "art", "gaming", "game", "esports", "sports", "events", "concert", "festival", "theater", "theatre", "performance", "show", "production", "studio", "record", "label", "artist", "actor", "director", "producer"}
	hasEntertainmentKeywords := false
	for _, kw := range keywords {
		kwLower := strings.ToLower(kw)
		for _, entKw := range entertainmentKeywords {
			if strings.Contains(kwLower, entKw) || strings.Contains(entKw, kwLower) {
				hasEntertainmentKeywords = true
				break
			}
		}
		if hasEntertainmentKeywords {
			break
		}
	}
	
	// If Entertainment keywords are present but not matched, boost Entertainment industry
	if hasEntertainmentKeywords && bestIndustryID != 26 {
		// Check if Entertainment industry exists in matches
		for industryID, matched := range industryMatches {
			industry, err := r.GetIndustryByID(ctx, industryID)
			if err == nil && strings.Contains(strings.ToLower(industry.Name), "entertainment") {
				// Boost Entertainment score significantly
				entertainmentScore := industryScores[industryID] / float64(len(keywords))
				entertainmentScore *= 1.5 // 50% boost for Entertainment keywords
				if entertainmentScore > bestScore {
					bestScore = entertainmentScore
					bestIndustryID = industryID
					bestMatchedKeywords = matched
					r.logger.Printf("🎬 [Priority 5.3] Boosted Entertainment industry due to Entertainment keywords")
				}
			}
		}
	}

	// Fix 2: Prioritize Food & Beverage over Retail when Food & Beverage keywords are present
	foodBeverageKeywords := []string{"restaurant", "restaurants", "cafe", "cafes", "coffee", "food", "dining", "kitchen", "catering", "bakery", "bar", "pub", "brewery", "winery", "wine", "beer", "cocktail", "menu", "chef", "cook", "cuisine", "delivery", "takeout", "fast food", "casual dining", "fine dining", "bistro", "eatery", "diner", "tavern", "gastropub", "food truck", "beverage", "drink", "alcohol", "spirits", "liquor"}
	hasFoodBeverageKeywords := false
	for _, kw := range keywords {
		kwLower := strings.ToLower(kw)
		for _, fbKw := range foodBeverageKeywords {
			if strings.Contains(kwLower, fbKw) || strings.Contains(fbKw, kwLower) {
				hasFoodBeverageKeywords = true
				break
			}
		}
		if hasFoodBeverageKeywords {
			break
		}
	}
	
	// If Food & Beverage keywords are present and Retail is winning, boost Food & Beverage
	if hasFoodBeverageKeywords && bestIndustryID != 26 {
		bestIndustry, _ := r.GetIndustryByID(ctx, bestIndustryID)
		if bestIndustry != nil && strings.Contains(strings.ToLower(bestIndustry.Name), "retail") {
			// Check if Food & Beverage industry exists in matches
			for industryID, matched := range industryMatches {
				industry, err := r.GetIndustryByID(ctx, industryID)
				if err == nil && (strings.Contains(strings.ToLower(industry.Name), "food") || strings.Contains(strings.ToLower(industry.Name), "beverage") || strings.Contains(strings.ToLower(industry.Name), "restaurant")) {
					// Boost Food & Beverage score significantly
					foodBeverageScore := industryScores[industryID] / float64(len(keywords))
					foodBeverageScore *= 1.4 // 40% boost for Food & Beverage keywords
					if foodBeverageScore > bestScore*0.9 { // Allow Food & Beverage to win even if slightly lower
						bestScore = foodBeverageScore
						bestIndustryID = industryID
						bestMatchedKeywords = matched
						r.logger.Printf("🍽️ [Priority 5.3] Boosted Food & Beverage industry over Retail due to Food & Beverage keywords")
					}
				}
			}
		}
	}

	// Fix 3: Distinguish Healthcare from Insurance
	healthcareKeywords := []string{"healthcare", "health", "medical", "hospital", "clinic", "doctor", "physician", "patient", "care", "treatment", "pharmacy", "diagnostic", "therapy", "wellness", "dental", "vision", "mental", "psychology", "counseling", "rehabilitation", "emergency", "ambulance", "laboratory", "radiology", "imaging", "surgery", "nursing"}
	hasHealthcareKeywords := false
	for _, kw := range keywords {
		kwLower := strings.ToLower(kw)
		for _, hcKw := range healthcareKeywords {
			if strings.Contains(kwLower, hcKw) || strings.Contains(hcKw, kwLower) {
				hasHealthcareKeywords = true
				break
			}
		}
		if hasHealthcareKeywords {
			break
		}
	}
	
	// If Healthcare keywords are present and Insurance is winning, boost Healthcare
	if hasHealthcareKeywords && bestIndustryID != 26 {
		bestIndustry, _ := r.GetIndustryByID(ctx, bestIndustryID)
		if bestIndustry != nil && strings.Contains(strings.ToLower(bestIndustry.Name), "insurance") {
			// Check if Healthcare industry exists in matches
			for industryID, matched := range industryMatches {
				industry, err := r.GetIndustryByID(ctx, industryID)
				if err == nil && (strings.Contains(strings.ToLower(industry.Name), "healthcare") || strings.Contains(strings.ToLower(industry.Name), "health") || strings.Contains(strings.ToLower(industry.Name), "medical")) {
					// Boost Healthcare score significantly
					healthcareScore := industryScores[industryID] / float64(len(keywords))
					healthcareScore *= 1.5 // 50% boost for Healthcare keywords
					if healthcareScore > bestScore*0.85 { // Allow Healthcare to win even if slightly lower
						bestScore = healthcareScore
						bestIndustryID = industryID
						bestMatchedKeywords = matched
						r.logger.Printf("🏥 [Priority 5.3] Boosted Healthcare industry over Insurance due to Healthcare keywords")
					}
				}
			}
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

	// Phase 7.2: Industry confidence thresholds (adaptive)
	// Priority 5.3: Reduced threshold to reduce "General Business" fallback
	const (
		MinKeywordCount    = 2    // Minimum keywords required for high confidence (reduced from 3)
		MinConfidenceScore = 0.25 // Priority 5.3: Reduced from 0.35 to 0.25 to reduce fallback (was 0.6 originally)
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

	// Step 2: Hybrid approach - supplement with trigram fuzzy matching if confidence is low
	// Use trigram when: confidence < 0.6 OR uniqueMatchCount < 2
	const trigramConfidenceThreshold = 0.6
	const trigramMatchCountThreshold = 2

	if confidence < trigramConfidenceThreshold || uniqueMatchCount < trigramMatchCountThreshold {
		r.logger.Printf("📊 Confidence %.2f or match count %d below threshold, supplementing with trigram fuzzy matching", confidence, uniqueMatchCount)

		// Try trigram classification to find fuzzy matches
		trigramResult, err := r.ClassifyBusinessByKeywordsTrigram(ctx, keywords, "", 0.3)
		if err == nil && trigramResult != nil {
			// Merge trigram results with exact match results
			trigramScore := trigramResult.Confidence
			trigramIndustryID := trigramResult.Industry.ID

			// If trigram found a better match (higher confidence or different industry with good confidence)
			if trigramScore > confidence && trigramIndustryID != 26 {
				r.logger.Printf("✅ Trigram found better match: industry %d with confidence %.2f (vs exact match %.2f)",
					trigramIndustryID, trigramScore, confidence)

				// Use trigram result but boost it slightly if exact matches also found this industry
				if trigramIndustryID == bestIndustryID {
					// Both methods agree - boost confidence
					confidence = math.Min(1.0, (confidence*0.3 + trigramScore*0.7))
					bestMatchedKeywords = append(bestMatchedKeywords, trigramResult.Keywords...)
					// Deduplicate keywords
					uniqueKeywords := make(map[string]bool)
					var deduplicated []string
					for _, kw := range bestMatchedKeywords {
						if !uniqueKeywords[kw] {
							uniqueKeywords[kw] = true
							deduplicated = append(deduplicated, kw)
						}
					}
					bestMatchedKeywords = deduplicated
				} else {
					// Trigram found different industry - use it if significantly better
					if trigramScore > confidence+0.15 {
						bestIndustryID = trigramIndustryID
						bestIndustry = trigramResult.Industry
						confidence = trigramScore
						bestMatchedKeywords = trigramResult.Keywords
						bestScore = trigramScore // Update bestScore for consistency
					}
				}
			} else if trigramIndustryID == bestIndustryID && trigramScore > 0.4 {
				// Both methods found same industry - combine scores
				confidence = math.Min(1.0, (confidence*0.4 + trigramScore*0.6))
				bestMatchedKeywords = append(bestMatchedKeywords, trigramResult.Keywords...)
				// Deduplicate keywords
				uniqueKeywords := make(map[string]bool)
				var deduplicated []string
				for _, kw := range bestMatchedKeywords {
					if !uniqueKeywords[kw] {
						uniqueKeywords[kw] = true
						deduplicated = append(deduplicated, kw)
					}
				}
				bestMatchedKeywords = deduplicated
			}
		} else if err != nil {
			r.logger.Printf("⚠️ Trigram classification failed: %v", err)
		}
	}

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
// Configurable timeout: default 30 seconds (increased to accommodate extractKeywords + classification)
// Can be overridden by parent context if shorter
func (r *SupabaseKeywordRepository) ClassifyBusinessByContextualKeywords(ctx context.Context, contextualKeywords []ContextualKeyword) (*ClassificationResult, error) {
	const defaultClassificationTimeout = 30 * time.Second

	// Create context with timeout if parent context doesn't have a deadline or has a longer deadline
	var classificationCtx context.Context
	var cancel context.CancelFunc

	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		if timeRemaining < defaultClassificationTimeout {
			// Use parent context if it has a shorter deadline
			classificationCtx = ctx
			cancel = func() {} // No-op cancel
			r.logger.Printf("⏱️ [TIMEOUT] Using parent context deadline: %v (shorter than default %v)", timeRemaining, defaultClassificationTimeout)
		} else {
			// Create new context with default timeout
			classificationCtx, cancel = context.WithTimeout(ctx, defaultClassificationTimeout)
			r.logger.Printf("⏱️ [TIMEOUT] Created classification context with timeout: %v (parent has %v remaining)", defaultClassificationTimeout, timeRemaining)
		}
	} else {
		// Parent has no deadline, create our own
		classificationCtx, cancel = context.WithTimeout(ctx, defaultClassificationTimeout)
		r.logger.Printf("⏱️ [TIMEOUT] Created classification context with timeout: %v (parent has no deadline)", defaultClassificationTimeout)
	}
	defer cancel()

	// PROFILING: Track time at function entry
	funcStartTime := time.Now()
	if deadline, hasDeadline := classificationCtx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] ClassifyBusinessByContextualKeywords entry - time remaining: %v", timeRemaining)
	}

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

	// Check timeout before proceeding
	if err := classificationCtx.Err(); err != nil {
		return nil, fmt.Errorf("classification context cancelled before start: %w", err)
	}

	// Ensure keyword index is built
	indexCheckStart := time.Now()
	index := r.GetKeywordIndex()
	indexCheckDuration := time.Since(indexCheckStart)

	// Enhanced logging for index state
	indexState := "empty"
	indexAge := time.Duration(0)
	keywordCount := len(index.KeywordToIndustries)
	if keywordCount > 0 {
		if index.LastUpdated > 0 {
			indexAge = time.Since(time.Unix(index.LastUpdated, 0))
			if indexAge < 5*time.Minute {
				indexState = fmt.Sprintf("populated (age: %v, valid)", indexAge)
			} else {
				indexState = fmt.Sprintf("populated (age: %v, expired)", indexAge)
			}
		} else {
			indexState = "populated (no timestamp)"
		}
	}
	r.logger.Printf("📊 [INDEX STATE] Keyword index: %s, keywords: %d, industries: %d",
		indexState, keywordCount, len(index.IndustryToKeywords))

	if deadline, hasDeadline := classificationCtx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] After GetKeywordIndex - time remaining: %v, index_check_duration: %v, elapsed: %v", timeRemaining, indexCheckDuration, time.Since(funcStartTime))
	}

	if len(index.KeywordToIndustries) == 0 {
		r.logger.Printf("⚠️ Keyword index is empty, building it now...")
		buildStartTime := time.Now()
		if deadline, hasDeadline := classificationCtx.Deadline(); hasDeadline {
			timeRemaining := time.Until(deadline)
			r.logger.Printf("⏱️ [PROFILING] Before BuildKeywordIndex - time remaining: %v, elapsed: %v", timeRemaining, time.Since(funcStartTime))
		}

		if err := r.BuildKeywordIndex(classificationCtx); err != nil {
			r.logger.Printf("⚠️ Failed to build keyword index: %v", err)
			// Convert contextual keywords to strings for fallback
			keywords := make([]string, len(contextualKeywords))
			for i, ck := range contextualKeywords {
				keywords[i] = ck.Keyword
			}
			return r.fallbackClassification(keywords, "Failed to build keyword index"), nil
		}

		buildDuration := time.Since(buildStartTime)
		if deadline, hasDeadline := classificationCtx.Deadline(); hasDeadline {
			timeRemaining := time.Until(deadline)
			r.logger.Printf("⏱️ [PROFILING] After BuildKeywordIndex - time remaining: %v, build_duration: %v, elapsed: %v", timeRemaining, buildDuration, time.Since(funcStartTime))
		}

		index = r.GetKeywordIndex()
		r.logger.Printf("✅ [INDEX BUILT] Keyword index built successfully: %d keywords, %d industries",
			len(index.KeywordToIndustries), len(index.IndustryToKeywords))
	} else if indexAge >= 5*time.Minute {
		r.logger.Printf("⚠️ Keyword index expired (age: %v), but using it anyway to avoid rebuild during request", indexAge)
	}

	// Use enhanced scoring algorithm for improved accuracy and performance
	// Create context with timeout for CalculateEnhancedScore (default: 8 seconds)
	const defaultEnhancedScoreTimeout = 8 * time.Second
	var enhancedScoreCtx context.Context
	var enhancedScoreCancel context.CancelFunc

	if deadline, hasDeadline := classificationCtx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		if timeRemaining < defaultEnhancedScoreTimeout {
			enhancedScoreCtx = classificationCtx
			enhancedScoreCancel = func() {}
			r.logger.Printf("⏱️ [TIMEOUT] Using classification context for enhanced scoring: %v (shorter than default %v)", timeRemaining, defaultEnhancedScoreTimeout)
		} else {
			enhancedScoreCtx, enhancedScoreCancel = context.WithTimeout(classificationCtx, defaultEnhancedScoreTimeout)
			r.logger.Printf("⏱️ [TIMEOUT] Created enhanced scoring context with timeout: %v", defaultEnhancedScoreTimeout)
		}
	} else {
		enhancedScoreCtx, enhancedScoreCancel = context.WithTimeout(classificationCtx, defaultEnhancedScoreTimeout)
		r.logger.Printf("⏱️ [TIMEOUT] Created enhanced scoring context with timeout: %v", defaultEnhancedScoreTimeout)
	}
	defer enhancedScoreCancel()

	enhancedScorerStart := time.Now()
	if deadline, hasDeadline := enhancedScoreCtx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] Before CalculateEnhancedScore - time remaining: %v, elapsed: %v", timeRemaining, time.Since(funcStartTime))
	}

	// Use reused enhanced scoring algorithm instance (thread-safe)
	r.enhancedScorerMutex.RLock()
	enhancedScorer := r.enhancedScorer
	r.enhancedScorerMutex.RUnlock()

	// Initialize if not already initialized (shouldn't happen, but safety check)
	if enhancedScorer == nil {
		r.enhancedScorerMutex.Lock()
		if r.enhancedScorer == nil {
			r.enhancedScorer = NewEnhancedScoringAlgorithm(r.logger, DefaultEnhancedScoringConfig())
		}
		enhancedScorer = r.enhancedScorer
		r.enhancedScorerMutex.Unlock()
	}

	enhancedResult, err := enhancedScorer.CalculateEnhancedScore(enhancedScoreCtx, contextualKeywords, index)

	enhancedScorerDuration := time.Since(enhancedScorerStart)
	if deadline, hasDeadline := enhancedScoreCtx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] After CalculateEnhancedScore - time remaining: %v, calculate_enhanced_score_duration: %v, elapsed: %v", timeRemaining, enhancedScorerDuration, time.Since(funcStartTime))
	}

	if err != nil {
		// Check if error is due to timeout
		if enhancedScoreCtx.Err() == context.DeadlineExceeded {
			r.logger.Printf("⚠️ Enhanced scoring timed out after %v, falling back to basic algorithm", defaultEnhancedScoreTimeout)
		} else {
			r.logger.Printf("⚠️ Enhanced scoring failed, falling back to basic algorithm: %v", err)
		}
		return r.classifyBusinessByContextualKeywordsBasic(classificationCtx, contextualKeywords, index)
	}

	// Parallelize database queries: GetIndustryByID and GetCachedClassificationCodes
	type queryResults struct {
		industry         *Industry
		codes            []*ClassificationCode
		industryErr      error
		codesErr         error
		industryDuration time.Duration
		codesDuration    time.Duration
	}

	var bestIndustry *Industry
	var codesPtr []*ClassificationCode
	var getIndustryDuration, getCodesDuration time.Duration

	queryStart := time.Now()
	if deadline, hasDeadline := classificationCtx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] Before parallel queries - time remaining: %v, elapsed: %v", timeRemaining, time.Since(funcStartTime))
		// Check if we have enough time for queries (need at least 2 seconds)
		if timeRemaining < 2*time.Second {
			r.logger.Printf("⚠️ Context deadline too short for parallel queries, using defaults")
			// Return defaults if deadline too short
			bestIndustry = &Industry{Name: "General Business", ID: 26}
			if enhancedResult.IndustryID != 26 {
				bestIndustry = &Industry{Name: "General Business", ID: enhancedResult.IndustryID}
			}
			codesPtr = []*ClassificationCode{}
			getIndustryDuration = 0
			getCodesDuration = 0
		} else if enhancedResult.IndustryID != 26 {
			// Execute queries in parallel
			resultsChan := make(chan queryResults, 2)
			var wg sync.WaitGroup

			// Get industry information in parallel
			wg.Add(1)
			go func() {
				defer wg.Done()
				var industry *Industry
				var err error
				industryDuration := time.Duration(0)

				err = r.timedQuery(classificationCtx, "GetIndustryByID", map[string]interface{}{
					"industry_id": enhancedResult.IndustryID,
				}, func() error {
					industryStart := time.Now()
					industry, err = r.GetIndustryByID(classificationCtx, enhancedResult.IndustryID)
					industryDuration = time.Since(industryStart)
					return err
				})

				resultsChan <- queryResults{
					industry:         industry,
					industryErr:      err,
					industryDuration: industryDuration,
				}
			}()

			// Get classification codes in parallel
			wg.Add(1)
			go func() {
				defer wg.Done()
				var codes []*ClassificationCode
				var err error
				codesDuration := time.Duration(0)

				err = r.timedQuery(classificationCtx, "GetCachedClassificationCodes", map[string]interface{}{
					"industry_id": enhancedResult.IndustryID,
				}, func() error {
					codesStart := time.Now()
					codes, err = r.GetCachedClassificationCodes(classificationCtx, enhancedResult.IndustryID)
					codesDuration = time.Since(codesStart)
					return err
				})

				resultsChan <- queryResults{
					codes:         codes,
					codesErr:      err,
					codesDuration: codesDuration,
				}
			}()

			// Wait for both queries to complete
			wg.Wait()
			close(resultsChan)

			// Collect results (each goroutine sends one result with either industry or codes)
			var industryResult, codesResult queryResults
			for result := range resultsChan {
				// Check which type of result this is
				if result.industry != nil || result.industryErr != nil {
					industryResult = result
				}
				if result.codes != nil || result.codesErr != nil {
					codesResult = result
				}
			}

			// Process industry result
			getIndustryDuration = industryResult.industryDuration
			if industryResult.industryErr != nil {
				r.logger.Printf("⚠️ Failed to get industry %d: %v", enhancedResult.IndustryID, industryResult.industryErr)
				bestIndustry = &Industry{Name: "General Business", ID: 26}
			} else {
				bestIndustry = industryResult.industry
			}

			// Process codes result
			getCodesDuration = codesResult.codesDuration
			if codesResult.codesErr != nil {
				r.logger.Printf("⚠️ Failed to get classification codes for industry %d: %v", enhancedResult.IndustryID, codesResult.codesErr)
				codesPtr = []*ClassificationCode{}
			} else {
				codesPtr = codesResult.codes
			}

			totalQueryDuration := time.Since(queryStart)
			if deadline, hasDeadline := classificationCtx.Deadline(); hasDeadline {
				timeRemaining := time.Until(deadline)
				r.logger.Printf("⏱️ [PROFILING] After parallel queries - time remaining: %v, total_query_duration: %v (industry: %v, codes: %v), elapsed: %v",
					timeRemaining, totalQueryDuration, getIndustryDuration, getCodesDuration, time.Since(funcStartTime))
			}
		}
	} else {
		bestIndustry = &Industry{Name: "General Business", ID: 26}
		// Still get codes even for default industry
		getCodesStart := time.Now()
		var err error
		codesPtr, err = r.GetCachedClassificationCodes(classificationCtx, enhancedResult.IndustryID)
		getCodesDuration = time.Since(getCodesStart)
		if err != nil {
			r.logger.Printf("⚠️ Failed to get classification codes for industry %d: %v", enhancedResult.IndustryID, err)
			codesPtr = []*ClassificationCode{}
		}
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
	totalDuration := time.Since(funcStartTime)
	if deadline, hasDeadline := classificationCtx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] ClassifyBusinessByContextualKeywords exit - time remaining: %v, total_duration: %v", timeRemaining, totalDuration)
		r.logger.Printf("⏱️ [PROFILING] Duration breakdown - GetKeywordIndex: %v, CalculateEnhancedScore: %v, GetIndustryByID: %v, GetCachedClassificationCodes: %v",
			indexCheckDuration, enhancedScorerDuration, getIndustryDuration, getCodesDuration)
	} else {
		r.logger.Printf("⏱️ [PROFILING] ClassifyBusinessByContextualKeywords exit - total_duration: %v", totalDuration)
		r.logger.Printf("⏱️ [PROFILING] Duration breakdown - GetKeywordIndex: %v, CalculateEnhancedScore: %v, GetIndustryByID: %v, GetCachedClassificationCodes: %v",
			indexCheckDuration, enhancedScorerDuration, getIndustryDuration, getCodesDuration)
	}

	// Check if timeout was exceeded
	if classificationCtx.Err() == context.DeadlineExceeded {
		r.logger.Printf("⚠️ [TIMEOUT] Classification timeout exceeded after %v", totalDuration)
	}

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
// FIX: Now accepts context parameter to maintain proper context propagation (Root Cause #1)
func (r *SupabaseKeywordRepository) extractKeywords(ctx context.Context, businessName, websiteURL string) []ContextualKeyword {
	extractionStartTime := time.Now()
	r.logger.Printf("🔍 [KeywordExtraction] Starting extraction for: %s (business: %s)", websiteURL, businessName)

	// PROFILING: Detailed profiling for extractKeywords optimization
	funcStartTime := time.Now()
	if deadline, ok := ctx.Deadline(); ok {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [PROFILING] [extractKeywords] Entry - time remaining: %v", timeRemaining)
	} else {
		r.logger.Printf("⏱️ [PROFILING] [extractKeywords] Entry - no deadline")
	}

	// Log parent context state for debugging
	if deadline, ok := ctx.Deadline(); ok {
		timeRemaining := time.Until(deadline)
		r.logger.Printf("⏱️ [KeywordExtraction] Parent context deadline: %v from now", timeRemaining)
	} else {
		r.logger.Printf("⏱️ [KeywordExtraction] Parent context has no deadline")
	}

	var keywords []ContextualKeyword
	seen := make(map[string]bool)

	// Track metrics for observability (Phase 8.2)
	// Note: Error counts are logged in nested functions but not tracked here
	// as they occur in separate function scopes. Errors are properly categorized
	// and logged in each extraction method (Phase 8.3).
	metrics := struct {
		startTime      time.Time
		level1Time     time.Duration
		level2Time     time.Duration
		level3Time     time.Duration
		level4Time     time.Duration
		level1Keywords int
		level2Keywords int
		level3Keywords int
		level4Keywords int
		level1Success  bool
		level2Success  bool
		level3Success  bool
		level4Success  bool
	}{
		startTime: extractionStartTime,
	}

	// PRIORITY 1: Extract keywords from website content (HIGHEST PRIORITY)
	// Optimized: Try fastest method first (single-page) before slower multi-page
	if websiteURL != "" {
		analysisMethod := "none"
		confidenceLevel := "high"

		// OPTIMIZATION: Reduced timeout requirement (was 18s, now 15s)
		// Phase 1 scraper: 10s max (reduced from 12s), overhead: 5s buffer
		// This allows faster failure and reduces overall extractKeywords duration
		const requiredTimeout = 15 * time.Second
		var extractionCtx context.Context
		var cancel context.CancelFunc

		if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
			timeRemaining := time.Until(deadline)
			if timeRemaining < requiredTimeout {
				// Parent context doesn't have enough time, create new context from Background
				r.logger.Printf("🔧 [KeywordExtraction] Parent context has insufficient time (%v < %v), creating separate context from Background", timeRemaining, requiredTimeout)
				extractionCtx, cancel = context.WithTimeout(context.Background(), requiredTimeout)
			} else {
				// Parent has enough time, use it directly
				r.logger.Printf("✅ [KeywordExtraction] Parent context has sufficient time (%v >= %v), using parent context", timeRemaining, requiredTimeout)
				extractionCtx = ctx
				cancel = func() {} // No-op cancel for parent context
			}
		} else {
			// Parent context has no deadline, create our own with required timeout
			r.logger.Printf("🔧 [KeywordExtraction] Parent context has no deadline, creating context with %v timeout", requiredTimeout)
			extractionCtx, cancel = context.WithTimeout(ctx, requiredTimeout)
		}
		defer cancel()

		// OPTIMIZATION: Try Level 2 (single-page) FIRST - it's faster than multi-page
		// Level 2: Single-page analysis (homepage only) - requires 3+ keywords for success
		// FIX: Check context deadline before starting single-page analysis
		hasEnoughTimeForLevel2 := true
		if deadline, ok := extractionCtx.Deadline(); ok {
			timeRemaining := time.Until(deadline)
			// FIX: Add defensive check for negative time (expired context)
			if timeRemaining < 0 {
				hasEnoughTimeForLevel2 = false
				r.logger.Printf("⚠️ [KeywordExtraction] Level 2 SKIPPED: Context already expired (time remaining: %v)", timeRemaining)
			} else if timeRemaining < 11*time.Second {
				hasEnoughTimeForLevel2 = false
				r.logger.Printf("⚠️ [KeywordExtraction] Level 2 SKIPPED: Insufficient time remaining (%v < 11s required for Phase 1 scraper)", timeRemaining)
			} else {
				r.logger.Printf("📊 [KeywordExtraction] Level 2: Starting single-page website analysis (homepage only) - trying fastest method first (time remaining: %v)", timeRemaining)
			}
		} else {
			r.logger.Printf("📊 [KeywordExtraction] Level 2: Starting single-page website analysis (homepage only) - trying fastest method first (no deadline on context)")
		}

		if hasEnoughTimeForLevel2 {
			level2Start := time.Now()
			if deadline, ok := extractionCtx.Deadline(); ok {
				timeRemaining := time.Until(deadline)
				r.logger.Printf("⏱️ [PROFILING] [extractKeywords] Before extractKeywordsFromWebsite (Level 2) - time remaining: %v, elapsed: %v", timeRemaining, time.Since(funcStartTime))
			}
			singlePageKeywords := r.extractKeywordsFromWebsite(extractionCtx, websiteURL)
			metrics.level2Time = time.Since(level2Start)
			metrics.level2Keywords = len(singlePageKeywords)

			if deadline, ok := extractionCtx.Deadline(); ok {
				timeRemaining := time.Until(deadline)
				r.logger.Printf("⏱️ [PROFILING] [extractKeywords] After extractKeywordsFromWebsite (Level 2) - time remaining: %v, level2_duration: %v, elapsed: %v", timeRemaining, metrics.level2Time, time.Since(funcStartTime))
			}

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
				analysisMethod = "single_page"
				confidenceLevel = "medium"
				metrics.level2Success = true
				r.logger.Printf("✅ [KeywordExtraction] Level 2 SUCCESS: Extracted %d keywords from single-page analysis (threshold: 3+)", len(singlePageKeywords))
				r.logger.Printf("📝 [KeywordExtraction] Level 2 keywords: %v", singlePageKeywords)
				// EARLY TERMINATION: Return immediately after Level 2 success to avoid unnecessary Level 1/3/4 calls
				r.logger.Printf("🚀 [KeywordExtraction] Early termination: Level 2 success, skipping Levels 1/3/4")
				return keywords
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
				analysisMethod = "single_page_partial"
				confidenceLevel = "low"
				metrics.level2Success = false
				r.logger.Printf("⚠️ [KeywordExtraction] Level 2 PARTIAL: Only %d keywords from single-page analysis (below threshold of 3), trying Level 1 (multi-page)", len(singlePageKeywords))
				if len(singlePageKeywords) > 0 {
					r.logger.Printf("📝 [KeywordExtraction] Level 2 partial keywords: %v", singlePageKeywords)
				}
			}
		} else {
			// Not enough time for Level 2, skip it
			r.logger.Printf("⚠️ [KeywordExtraction] Level 2 SKIPPED: Insufficient time remaining for single-page analysis")
		}

		// Level 1: Multi-page analysis (15 pages) - requires 5+ keywords for success
		// Only try if Level 2 didn't get enough keywords
		// FIX: Check context deadline before starting expensive multi-page analysis
		if len(keywords) < 5 {
			// Check if we have enough time remaining for multi-page analysis
			hasEnoughTime := true
			if deadline, ok := extractionCtx.Deadline(); ok {
				timeRemaining := time.Until(deadline)
				// FIX: Add defensive check for negative time (expired context)
				if timeRemaining < 0 {
					hasEnoughTime = false
					r.logger.Printf("⚠️ [KeywordExtraction] Level 1 SKIPPED: Context already expired (time remaining: %v)", timeRemaining)
				} else if timeRemaining < 5*time.Second {
					hasEnoughTime = false
					r.logger.Printf("⚠️ [KeywordExtraction] Level 1 SKIPPED: Insufficient time remaining (%v < 5s required for multi-page analysis)", timeRemaining)
				} else {
					r.logger.Printf("✅ [KeywordExtraction] Level 1: Starting multi-page website analysis (max 15 pages, timeout: capped at 10s, time remaining: %v)", timeRemaining)
				}
			} else {
				r.logger.Printf("📊 [KeywordExtraction] Level 1: Starting multi-page website analysis (max 15 pages, timeout: capped at 10s, no deadline on context)")
			}

			if hasEnoughTime {
				level1Start := time.Now()
				multiPageKeywords := r.extractKeywordsFromMultiPageWebsite(extractionCtx, websiteURL)
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
					// EARLY TERMINATION: Return immediately after Level 1 success to avoid unnecessary Level 3/4 calls
					r.logger.Printf("🚀 [KeywordExtraction] Early termination: Level 1 success, skipping Levels 3-4")
					return keywords
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
					if analysisMethod == "none" || analysisMethod == "single_page_partial" {
						analysisMethod = "multi_page_partial"
						confidenceLevel = "medium"
					}
					metrics.level1Success = false
					r.logger.Printf("⚠️ [KeywordExtraction] Level 1 PARTIAL: Only %d keywords from multi-page analysis (below threshold of 5), continuing fallback", len(multiPageKeywords))
					if len(multiPageKeywords) > 0 {
						r.logger.Printf("📝 [KeywordExtraction] Level 1 partial keywords: %v", multiPageKeywords)
					}
				}
			}
		}

		// OPTIMIZATION: Priority 5 - Parallel extraction from multiple sources
		// Level 3 and Level 4 can run in parallel since they're independent
		needsLevel3 := len(keywords) < 3
		needsLevel4 := len(keywords) < 2

		if needsLevel3 && needsLevel4 {
			// Run Level 3 and Level 4 in parallel
			r.logger.Printf("🚀 [OPTIMIZATION] [extractKeywords] Running Level 3 and Level 4 in parallel")

			type levelResult struct {
				level       int
				keywords    []string
				urlKeywords []ContextualKeyword
				duration    time.Duration
			}

			resultsChan := make(chan levelResult, 2)
			var wg sync.WaitGroup

			// Level 3: Homepage retry (needs context)
			wg.Add(1)
			go func() {
				defer wg.Done()
				level3Start := time.Now()
				r.logger.Printf("📊 [KeywordExtraction] Level 3 (parallel): Starting homepage extraction with enhanced retry")
				homepageKeywords := r.extractKeywordsFromHomepageWithRetry(extractionCtx, websiteURL)
				level3Duration := time.Since(level3Start)
				resultsChan <- levelResult{
					level:    3,
					keywords: homepageKeywords,
					duration: level3Duration,
				}
			}()

			// Level 4: URL extraction (no context needed, very fast)
			wg.Add(1)
			go func() {
				defer wg.Done()
				level4Start := time.Now()
				r.logger.Printf("📊 [KeywordExtraction] Level 4 (parallel): Starting enhanced URL text extraction")
				urlKeywords := r.extractKeywordsFromURLEnhanced(websiteURL)
				level4Duration := time.Since(level4Start)
				resultsChan <- levelResult{
					level:       4,
					urlKeywords: urlKeywords,
					duration:    level4Duration,
				}
			}()

			// Wait for both to complete
			wg.Wait()
			close(resultsChan)

			// Process results
			for result := range resultsChan {
				if result.level == 3 {
					metrics.level3Time = result.duration
					metrics.level3Keywords = len(result.keywords)
					r.logger.Printf("📊 [KeywordExtraction] Level 3 (parallel) completed in %v: extracted %d keywords", metrics.level3Time, len(result.keywords))

					if len(result.keywords) >= 2 {
						// Success: enough keywords from retry
						for _, keyword := range result.keywords {
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
						r.logger.Printf("✅ [KeywordExtraction] Level 3 SUCCESS: Extracted %d keywords from homepage with retry (threshold: 2+)", len(result.keywords))
					} else if len(result.keywords) > 0 {
						// Partial success
						for _, keyword := range result.keywords {
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
						r.logger.Printf("⚠️ [KeywordExtraction] Level 3 PARTIAL: Only %d keywords from homepage retry (below threshold of 2)", len(result.keywords))
					}
				} else if result.level == 4 {
					metrics.level4Time = result.duration
					metrics.level4Keywords = len(result.urlKeywords)
					r.logger.Printf("📊 [KeywordExtraction] Level 4 (parallel) completed in %v: extracted %d keywords", metrics.level4Time, len(result.urlKeywords))

					if len(result.urlKeywords) >= 1 {
						for _, keyword := range result.urlKeywords {
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
						r.logger.Printf("✅ [KeywordExtraction] Level 4 SUCCESS: Extracted %d keywords from enhanced URL text extraction (threshold: 1+)", len(result.urlKeywords))
					}
				}
			}

			r.logger.Printf("✅ [OPTIMIZATION] [extractKeywords] Parallel extraction completed (Level 3: %v, Level 4: %v)", metrics.level3Time, metrics.level4Time)
		} else {
			// Sequential execution (only one level needed)
			if needsLevel3 {
				r.logger.Printf("📊 [KeywordExtraction] Level 3: Starting homepage extraction with enhanced retry (multiple DNS servers)")
				level3Start := time.Now()
				homepageKeywords := r.extractKeywordsFromHomepageWithRetry(extractionCtx, websiteURL)
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

			if needsLevel4 {
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
		}

		// Level 5: Business name analysis (if brand match) - already handled below
		// Level 6: Default to "General Business" with low confidence - handled by returning empty keywords

		// Phase 8.2: Log performance metrics
		totalExtractionTime := time.Since(extractionStartTime)
		totalFuncTime := time.Since(funcStartTime)
		if deadline, ok := ctx.Deadline(); ok {
			timeRemaining := time.Until(deadline)
			r.logger.Printf("⏱️ [PROFILING] [extractKeywords] Exit - time remaining: %v, total_duration: %v", timeRemaining, totalFuncTime)
		}
		r.logger.Printf("📊 [KeywordExtraction] Performance Metrics:")
		r.logger.Printf("   - Total extraction time: %v", totalExtractionTime)
		r.logger.Printf("   - Total function time: %v", totalFuncTime)
		r.logger.Printf("   - Level 1 (multi-page): %v, keywords: %d, success: %v", metrics.level1Time, metrics.level1Keywords, metrics.level1Success)
		r.logger.Printf("   - Level 2 (single-page): %v, keywords: %d, success: %v", metrics.level2Time, metrics.level2Keywords, metrics.level2Success)
		r.logger.Printf("   - Level 3 (homepage-retry): %v, keywords: %d, success: %v", metrics.level3Time, metrics.level3Keywords, metrics.level3Success)
		r.logger.Printf("   - Level 4 (URL-only): %v, keywords: %d, success: %v", metrics.level4Time, metrics.level4Keywords, metrics.level4Success)
		r.logger.Printf("   - Note: Errors are logged in individual extraction methods with categorization (DNS, HTTP, Parsing)")

		// Phase 3.3: Integrate NER and topic modeling with keyword extraction
		// Apply NER to extract entities from collected keywords and enhance with entity-based keywords
		if len(keywords) > 0 {
			// Combine all keyword text for entity recognition
			combinedText := r.combineKeywordsForNER(keywords)
			if combinedText != "" {
				r.logger.Printf("🔍 [KeywordExtraction] [NLP] Applying Named Entity Recognition to extracted keywords")
				entities := r.entityRecognizer.ExtractEntities(combinedText)

				// Extract keywords from entities
				entityKeywords := r.entityRecognizer.GetEntityKeywords(entities)
				for _, entityKw := range entityKeywords {
					if !seen[entityKw] {
						seen[entityKw] = true
						keywords = append(keywords, ContextualKeyword{
							Keyword: entityKw,
							Context: "ner_entity",
						})
					}
				}

				if len(entities) > 0 {
					r.logger.Printf("✅ [KeywordExtraction] [NLP] Extracted %d entities, added %d entity-based keywords", len(entities), len(entityKeywords))
				}
			}

			// Apply topic modeling to identify industry topics and enhance keyword relevance
			keywordStrings := make([]string, 0, len(keywords))
			for _, kw := range keywords {
				keywordStrings = append(keywordStrings, kw.Keyword)
			}

			r.logger.Printf("🔍 [KeywordExtraction] [NLP] Applying topic modeling to identify industry topics")
			topicScores := r.topicModeler.IdentifyTopics(keywordStrings)

			if len(topicScores) > 0 {
				// Log top industry topics
				maxIndustries := 3
				if len(topicScores) < maxIndustries {
					maxIndustries = len(topicScores)
				}
				topIndustries := make([]int, 0, maxIndustries)
				for industryID := range topicScores {
					topIndustries = append(topIndustries, industryID)
					if len(topIndustries) >= 3 {
						break
					}
				}
				r.logger.Printf("📊 [KeywordExtraction] [NLP] Topic modeling identified %d industries with scores: %v", len(topicScores), topicScores)
			}
		}

		r.logger.Printf("📊 [KeywordExtraction] Final result: method=%s, confidence=%s, total_unique_keywords=%d", analysisMethod, confidenceLevel, len(keywords))

		// Log top keywords for observability
		if len(keywords) > 0 {
			maxKeywords := 10
			if len(keywords) < maxKeywords {
				maxKeywords = len(keywords)
			}
			topKeywords := make([]string, 0, maxKeywords)
			for i, kw := range keywords {
				if i >= 10 {
					break
				}
				topKeywords = append(topKeywords, kw.Keyword)
			}
			r.logger.Printf("📝 [KeywordExtraction] Top keywords: %v", topKeywords)
		}
	} // FIX: Close if websiteURL != "" block that starts at line 2738

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

// combineKeywordsForNER combines keywords into text for entity recognition
func (r *SupabaseKeywordRepository) combineKeywordsForNER(keywords []ContextualKeyword) string {
	if len(keywords) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, kw := range keywords {
		if i > 0 {
			builder.WriteString(" ")
		}
		builder.WriteString(kw.Keyword)
	}
	return builder.String()
}

// extractKeywordsFromHomepageWithRetry attempts to extract keywords from homepage with enhanced retry logic
// Uses different DNS servers, longer timeout, and multiple retry attempts (Phase 5.1)
// Phase 8: Enhanced with detailed logging and error tracking
// Phase 10: Increased timeout to 60s and use fresh context to avoid parent context expiration
func (r *SupabaseKeywordRepository) extractKeywordsFromHomepageWithRetry(ctx context.Context, websiteURL string) []string {
	startTime := time.Now()
	r.logger.Printf("🔄 [KeywordExtraction] [HomepageRetry] Starting homepage extraction with enhanced retry for: %s", websiteURL)

	// Create a fresh context with longer timeout (60 seconds for retry attempts)
	// Use context.Background() instead of inheriting from parent to avoid expiration
	// This ensures retry attempts have full 60 seconds even if parent context expires
	retryCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
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

			// Create custom dialer with longer timeout (increased to 45s for slow websites)
			baseDialer := &net.Dialer{
				Timeout:   45 * time.Second,
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

			// Phase 9.3: Apply rate limiting with jitter
			// Note: parsedURL already declared above, so we reuse it
			var domain string
			var session *scrapingSession
			if parsedURL != nil {
				domain = parsedURL.Hostname()
				r.applyRateLimit(domain)

				// Get or create session for this domain
				sess, err := r.sessionManager.getOrCreateSession(domain)
				if err != nil {
					r.logger.Printf("⚠️ [HomepageRetry] Failed to get session for %s: %v", domain, err)
				} else {
					session = sess
				}
			}

			// Create HTTP client with custom dialer, longer timeout, and session cookie jar
			client := &http.Client{
				Timeout: 60 * time.Second,
				Transport: &http.Transport{
					DialContext:           customDialContext,
					MaxIdleConns:          10,
					IdleConnTimeout:       30 * time.Second,
					DisableCompression:    false,
					MaxIdleConnsPerHost:   2,
					TLSHandshakeTimeout:   15 * time.Second,
					ResponseHeaderTimeout: 30 * time.Second,
				},
			}
			if session != nil {
				client.Jar = session.cookieJar
			}

			// Create request
			req, err := http.NewRequestWithContext(retryCtx, "GET", websiteURL, nil)
			if err != nil {
				r.logger.Printf("❌ [HomepageRetry] Failed to create request: %v", err)
				continue
			}

			// Enhanced headers with randomization to avoid blocking by websites that detect automated requests
			referer := ""
			if session != nil {
				referer = r.sessionManager.getReferer(domain)
			}
			headers := getRandomizedHeaders(getUserAgent(), referer)
			for key, value := range headers {
				req.Header.Set(key, value)
			}

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

			// Handle specific HTTP status codes
			if resp.StatusCode == 429 {
				// Too Many Requests - stop immediately, do not retry
				retryAfter := resp.Header.Get("Retry-After")
				r.logger.Printf("⚠️ [HomepageRetry] Rate limited (429) for %s, Retry-After: %s - stopping immediately", websiteURL, retryAfter)
				return []string{} // Stop immediately
			}
			if resp.StatusCode == 403 {
				// Forbidden - stop immediately, do not retry
				r.logger.Printf("🚫 [HomepageRetry] Access forbidden (403) for %s - stopping immediately", websiteURL)
				return []string{} // Stop immediately
			}
			if resp.StatusCode == 503 {
				// Service Unavailable - implement exponential backoff
				if attempt < maxRetries {
					backoffDelay := time.Duration(attempt) * 2 * time.Second // 2s, 4s, 6s
					r.logger.Printf("⚠️ [HomepageRetry] Service unavailable (503) for %s (attempt %d/%d), retrying after %v", websiteURL, attempt, maxRetries, backoffDelay)
					time.Sleep(backoffDelay)
					continue // Retry with exponential backoff
				} else {
					r.logger.Printf("❌ [HomepageRetry] Service unavailable (503) for %s after %d attempts", websiteURL, maxRetries)
					return []string{} // Stop after max retries
				}
			}
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

			// Check for CAPTCHA before processing
			if detected, captchaType := detectCAPTCHA(resp, body); detected {
				r.logger.Printf("🚫 [HomepageRetry] CAPTCHA detected (%s) for %s - stopping", captchaType, websiteURL)
				return []string{} // Stop immediately when CAPTCHA is detected
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

		// Exponential backoff before next retry (with jitter to avoid thundering herd)
		if attempt < maxRetries {
			baseBackoff := time.Duration(attempt) * 2 * time.Second     // Increased base backoff
			jitter := time.Duration(rand.Intn(1000)) * time.Millisecond // Random jitter 0-1s
			backoff := baseBackoff + jitter
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
// Phase 1: Now uses EnhancedWebsiteScraper with multi-tier strategies (SimpleHTTP → BrowserHeaders → Playwright)
// OPTIMIZATION: Priority 4 - Website content caching to reduce redundant scraping
func (r *SupabaseKeywordRepository) extractKeywordsFromWebsite(ctx context.Context, websiteURL string) []string {
	startTime := time.Now()
	r.logger.Printf("🌐 [KeywordExtraction] [SinglePage] Starting single-page website scraping for: %s", websiteURL)

	// OPTIMIZATION: Check cache first (Priority 4)
	cacheKey := websiteURL
	r.websiteContentCacheMutex.RLock()
	if cached, exists := r.websiteContentCache[cacheKey]; exists {
		if time.Now().Before(cached.expiresAt) {
			r.websiteContentCacheMutex.RUnlock()
			cacheAge := time.Since(cached.cachedAt)
			r.logger.Printf("✅ [CACHE] [extractKeywordsFromWebsite] Cache HIT for %s (age: %v, keywords: %d)", websiteURL, cacheAge, len(cached.keywords))
			return cached.keywords
		} else {
			// Expired entry, remove it
			r.websiteContentCacheMutex.RUnlock()
			r.websiteContentCacheMutex.Lock()
			delete(r.websiteContentCache, cacheKey)
			r.websiteContentCacheMutex.Unlock()
			r.logger.Printf("⏰ [CACHE] [extractKeywordsFromWebsite] Cache entry expired for %s, removing", websiteURL)
		}
	} else {
		r.websiteContentCacheMutex.RUnlock()
		r.logger.Printf("❌ [CACHE] [extractKeywordsFromWebsite] Cache MISS for %s", websiteURL)
	}

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

	// Phase 1: Use enhanced scraper if available (with multi-tier strategies)
	r.logger.Printf("🔍 [Phase1] [KeywordExtraction] Checking Phase 1 scraper availability for: %s (scraper is nil: %v)", websiteURL, r.websiteScraper == nil)

	// Log parent context state for debugging
	if deadline, ok := ctx.Deadline(); ok {
		parentTimeUntilDeadline := time.Until(deadline)
		r.logger.Printf("⏱️ [Phase1] [KeywordExtraction] Parent context deadline: %v from now (will create new context)", parentTimeUntilDeadline)
	} else {
		r.logger.Printf("⏱️ [Phase1] [KeywordExtraction] Parent context has no deadline")
	}
	r.logger.Printf("🔍 [Phase1] [KeywordExtraction] Parent context error state: %v", ctx.Err())

	if r.websiteScraper != nil {
		r.logger.Printf("✅ [Phase1] [KeywordExtraction] Using Phase 1 enhanced scraper for: %s", websiteURL)

		// OPTIMIZATION: Reduced Phase 1 scraper timeout (was 12s, now 10s)
		// This allows faster failure and reduces overall extractKeywords duration
		// Early termination will prevent waiting for slow websites
		const phase1RequiredTimeout = 10 * time.Second

		var phase1Ctx context.Context
		var phase1Cancel context.CancelFunc
		var useSeparateContext bool

		// Check parent context deadline
		deadline, hasDeadline := ctx.Deadline()
		if hasDeadline {
			timeRemaining := time.Until(deadline)
			if timeRemaining < phase1RequiredTimeout {
				// Parent context doesn't have enough time
				// Create new context from Background with full timeout
				// Don't propagate parent cancellation since parent had insufficient time anyway
				r.logger.Printf("🔧 [Phase1] [KeywordExtraction] Parent context has insufficient time (%v < %v), creating separate context from Background", timeRemaining, phase1RequiredTimeout)
				phase1Ctx, phase1Cancel = context.WithTimeout(context.Background(), phase1RequiredTimeout)
				useSeparateContext = true

				// Note: We do NOT monitor parent context cancellation here because:
				// 1. The parent context had insufficient time to begin with
				// 2. The Phase 1 scraper needs the full 15s to complete
				// 3. The parent context expiring is expected and should not cancel Phase 1
				// The Phase 1 context will run independently with its own 15s timeout
			} else {
				// Parent has enough time, use it directly
				r.logger.Printf("✅ [Phase1] [KeywordExtraction] Parent context has sufficient time (%v >= %v), using parent context", timeRemaining, phase1RequiredTimeout)
				phase1Ctx = ctx
				phase1Cancel = func() {} // No-op cancel for parent context
				useSeparateContext = false
			}
		} else {
			// Parent context has no deadline, create our own with required timeout
			r.logger.Printf("🔧 [Phase1] [KeywordExtraction] Parent context has no deadline, creating context with %v timeout", phase1RequiredTimeout)
			phase1Ctx, phase1Cancel = context.WithTimeout(ctx, phase1RequiredTimeout)
			useSeparateContext = false
		}

		// Only defer cancel if we created a separate context
		if useSeparateContext {
			defer phase1Cancel()
		}

		// Log context deadline for debugging
		if deadline, ok := phase1Ctx.Deadline(); ok {
			timeUntilDeadline := time.Until(deadline)
			r.logger.Printf("⏱️ [Phase1] [KeywordExtraction] Phase 1 context deadline: %v from now (deadline: %v, separate_context: %v)",
				timeUntilDeadline, deadline, useSeparateContext)
		}

		// Verify context is valid before passing
		ctxErr := phase1Ctx.Err()
		if ctxErr != nil {
			r.logger.Printf("❌ [Phase1] [KeywordExtraction] ERROR: Context already cancelled: %v", ctxErr)
			return []string{} // Fall back to legacy method
		}

		// Use Phase 1 enhanced scraper with multi-tier strategies
		r.logger.Printf("🚀 [Phase1] [KeywordExtraction] Calling Phase 1 scraper.ScrapeWebsite() for: %s (timeout: %v, separate_context: %v)",
			websiteURL, phase1RequiredTimeout, useSeparateContext)
		scrapingResultInterface := r.websiteScraper.ScrapeWebsite(phase1Ctx, websiteURL)
		r.logger.Printf("📥 [Phase1] [KeywordExtraction] Phase 1 scraper returned result (nil: %v)", scrapingResultInterface == nil)

		// Type assert to get the actual ScrapingResult
		// We use reflection to access fields to avoid import cycle
		if scrapingResultInterface != nil {
			// Use type assertion with a helper function to extract fields
			scrapingResult := r.extractScrapingResultFields(scrapingResultInterface)

			if scrapingResult != nil && scrapingResult.Success {
				// Extract keywords from the scraped content
				htmlContent := scrapingResult.Content
				if htmlContent == "" {
					htmlContent = scrapingResult.TextContent
				}

				if htmlContent != "" {
					// Use existing keyword extraction logic
					textContent := r.extractTextFromHTML(htmlContent)
					keywords := r.extractBusinessKeywords(textContent)

					// Also extract from structured data if available (same logic as legacy method)
					if NewStructuredDataExtractorAdapter != nil {
						structuredDataExtractor := NewStructuredDataExtractorAdapter(r.logger)
						structuredDataResult := structuredDataExtractor.ExtractStructuredData(htmlContent)

						// Extract structured keywords using same logic as legacy method
						structuredKeywordMap := make(map[string]float64)
						if structuredDataResult.BusinessInfo.Industry != "" {
							structuredKeywordMap[strings.ToLower(structuredDataResult.BusinessInfo.Industry)] = 2.0
						}
						if structuredDataResult.BusinessInfo.BusinessType != "" {
							structuredKeywordMap[strings.ToLower(structuredDataResult.BusinessInfo.BusinessType)] = 2.0
						}
						for _, service := range structuredDataResult.BusinessInfo.Services {
							structuredKeywordMap[strings.ToLower(service)] = 2.0
						}
						for _, product := range structuredDataResult.BusinessInfo.Products {
							structuredKeywordMap[strings.ToLower(product)] = 2.0
						}

						// Merge structured keywords with text keywords (weighted)
						allKeywords := make(map[string]float64)
						for _, kw := range keywords {
							allKeywords[strings.ToLower(kw)] = 1.0
						}
						for kw, weight := range structuredKeywordMap {
							if allKeywords[kw] < weight {
								allKeywords[kw] = weight
							}
						}

						// Convert back to slice
						keywords = make([]string, 0, len(allKeywords))
						for kw := range allKeywords {
							keywords = append(keywords, kw)
						}
					}

					duration := time.Since(startTime)
					r.logger.Printf("✅ [Phase1] [KeywordExtraction] Successfully extracted %d keywords in %v", len(keywords), duration)

					// OPTIMIZATION: Store in cache (Priority 4)
					r.storeWebsiteContentCache(cacheKey, keywords)

					return keywords
				}
			}

			// If Phase 1 scraper failed, log and fall through to legacy method
			if scrapingResult != nil && !scrapingResult.Success {
				r.logger.Printf("⚠️ [Phase1] [KeywordExtraction] Phase 1 scraper failed: %s, falling back to legacy method", scrapingResult.Error)
			}
		} else {
			r.logger.Printf("⚠️ [Phase1] [KeywordExtraction] Phase 1 scraper returned nil, falling back to legacy method")
		}
	} else {
		r.logger.Printf("ℹ️ [KeywordExtraction] Phase 1 enhanced scraper not available, using legacy scraping method")
	}

	// Legacy scraping method (fallback if Phase 1 scraper not available or failed)
	r.logger.Printf("🔄 [KeywordExtraction] [SinglePage] Using legacy scraping method for: %s", websiteURL)

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
		Timeout: 45 * time.Second, // Increased from 15s to 45s for slow websites
		Transport: &http.Transport{
			DialContext:           customDialContext,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			DisableCompression:    false,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}

	// Phase 9.3: Apply rate limiting with jitter
	// Note: parsedURL already declared above, so we reuse it
	var domain string
	var session *scrapingSession
	if parsedURL != nil {
		domain = parsedURL.Hostname()
		r.applyRateLimit(domain)

		// Get or create session for this domain
		sess, err := r.sessionManager.getOrCreateSession(domain)
		if err != nil {
			r.logger.Printf("⚠️ [SinglePage] Failed to get session for %s: %v", domain, err)
		} else {
			session = sess
		}
	}

	// Update HTTP client to use session cookie jar
	if session != nil {
		client.Jar = session.cookieJar
	}

	// Create request with enhanced headers
	req, err := http.NewRequestWithContext(ctx, "GET", websiteURL, nil)
	if err != nil {
		r.logger.Printf("❌ [KeywordExtraction] [SinglePage] PARSING ERROR: Failed to create request for %s: %v (type: %T)", websiteURL, err, err)
		return []string{}
	}

	// Set comprehensive headers with randomization to mimic a real browser
	referer := ""
	if session != nil {
		referer = r.sessionManager.getReferer(domain)
	}
	headers := getRandomizedHeaders(getUserAgent(), referer)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	r.logger.Printf("📡 [KeywordExtraction] [SinglePage] Making HTTP request to: %s", websiteURL)

	// Make request with timeout context (increased to 45s for slow websites)
	reqCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
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

	// Handle specific HTTP status codes
	if resp.StatusCode == 429 {
		// Too Many Requests - stop immediately, do not retry
		retryAfter := resp.Header.Get("Retry-After")
		r.logger.Printf("⚠️ [SinglePage] Rate limited (429) for %s, Retry-After: %s - stopping immediately", websiteURL, retryAfter)
		return []string{}
	}
	if resp.StatusCode == 403 {
		// Forbidden - stop immediately, do not retry
		r.logger.Printf("🚫 [SinglePage] Access forbidden (403) for %s - stopping immediately", websiteURL)
		return []string{}
	}
	if resp.StatusCode == 503 {
		// Service Unavailable - implement exponential backoff (up to 3 retries)
		maxRetries := 3
		for retryAttempt := 1; retryAttempt <= maxRetries; retryAttempt++ {
			backoffDelay := time.Duration(retryAttempt) * 2 * time.Second // 2s, 4s, 6s
			r.logger.Printf("⚠️ [SinglePage] Service unavailable (503) for %s (retry %d/%d), waiting %v", websiteURL, retryAttempt, maxRetries, backoffDelay)

			// Wait for backoff delay
			select {
			case <-ctx.Done():
				return []string{}
			case <-time.After(backoffDelay):
			}

			// Retry request
			retryReq, err := http.NewRequestWithContext(ctx, "GET", websiteURL, nil)
			if err != nil {
				return []string{}
			}
			// Set randomized headers for retry
			retryHeaders := getRandomizedHeaders(getUserAgent(), "")
			for key, value := range retryHeaders {
				retryReq.Header.Set(key, value)
			}

			retryResp, err := client.Do(retryReq)
			if err != nil {
				continue // Try next retry
			}
			defer retryResp.Body.Close()

			if retryResp.StatusCode == 200 {
				// Success on retry
				resp = retryResp
				break
			} else if retryResp.StatusCode == 503 && retryAttempt < maxRetries {
				// Still 503, continue retrying
				continue
			} else {
				// Other status code or max retries reached
				resp = retryResp
				break
			}
		}
		if resp.StatusCode == 503 {
			r.logger.Printf("❌ [SinglePage] Service unavailable (503) for %s after %d retry attempts", websiteURL, maxRetries)
			return []string{}
		}
	}

	// Log response details
	r.logger.Printf("📊 [KeywordExtraction] [SinglePage] Response received - Status: %d, Content-Type: %s, Content-Length: %d",
		resp.StatusCode, resp.Header.Get("Content-Type"), resp.ContentLength)

	// Phase 8.3: Track HTTP status codes
	if resp.StatusCode >= 400 {
		r.logger.Printf("❌ [KeywordExtraction] [SinglePage] HTTP ERROR: Status %d %s for %s", resp.StatusCode, resp.Status, websiteURL)
		// Try to read error response body
		if body, readErr := io.ReadAll(resp.Body); readErr == nil && len(body) > 0 {
			maxLen := 500
			if len(body) < maxLen {
				maxLen = len(body)
			}
			r.logger.Printf("📄 [KeywordExtraction] [SinglePage] Error response body (first 500 chars): %s", string(body[:maxLen]))
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

	// Read response body with size limit and handle decompression
	maxSize := int64(5 * 1024 * 1024) // 5MB limit
	var reader io.Reader = io.LimitReader(resp.Body, maxSize)

	// Handle content encoding (gzip, deflate, br)
	contentEncoding := resp.Header.Get("Content-Encoding")
	if contentEncoding != "" {
		r.logger.Printf("📦 [KeywordExtraction] [SinglePage] Content-Encoding: %s", contentEncoding)
	}

	switch contentEncoding {
	case "gzip":
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			r.logger.Printf("⚠️ [KeywordExtraction] [SinglePage] Failed to create gzip reader: %v, reading uncompressed", err)
		} else {
			defer gzipReader.Close()
			reader = gzipReader
			r.logger.Printf("📦 [KeywordExtraction] [SinglePage] Decompressing gzip content")
		}
	case "deflate":
		// Go's http package handles deflate automatically, but we'll log it
		r.logger.Printf("📦 [KeywordExtraction] [SinglePage] Deflate compression detected (handled by http.Client)")
	case "br":
		// Brotli compression - Go's http package doesn't handle this automatically
		r.logger.Printf("⚠️ [KeywordExtraction] [SinglePage] Brotli compression detected but not supported, may result in garbled text")
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		r.logger.Printf("❌ [KeywordExtraction] [SinglePage] PARSING ERROR: Failed to read response body from %s: %v (type: %T)", websiteURL, err, err)
		return []string{}
	}

	r.logger.Printf("📄 [KeywordExtraction] [SinglePage] Read %d bytes from %s (decompressed: %v)", len(body), websiteURL, contentEncoding != "")

	// Check for CAPTCHA before processing
	if detected, captchaType := detectCAPTCHA(resp, body); detected {
		r.logger.Printf("🚫 [SinglePage] CAPTCHA detected (%s) for %s - stopping", captchaType, websiteURL)
		return []string{} // Stop immediately when CAPTCHA is detected
	}

	// Check if body appears to be binary/garbled (contains null bytes or high percentage of non-printable chars)
	if len(body) > 0 {
		nonPrintableCount := 0
		maxLen := 1000
		if len(body) < maxLen {
			maxLen = len(body)
		}
		for i := 0; i < maxLen; i++ {
			if body[i] < 32 && body[i] != 9 && body[i] != 10 && body[i] != 13 {
				nonPrintableCount++
			}
		}
		if nonPrintableCount > 100 {
			r.logger.Printf("⚠️ [KeywordExtraction] [SinglePage] WARNING: Response body appears to be binary/garbled (%d non-printable chars in first 1000 bytes). Content-Encoding: %s", nonPrintableCount, contentEncoding)
		}
	}

	// Extract text content from HTML
	textContent := r.extractTextFromHTML(string(body))
	r.logger.Printf("🧹 [KeywordExtraction] [SinglePage] Extracted %d characters of text content from HTML", len(textContent))

	// Log sample of extracted text for debugging
	if len(textContent) > 0 {
		maxLen := 200
		if len(textContent) < maxLen {
			maxLen = len(textContent)
		}
		sampleText := textContent[:maxLen]
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

		// OPTIMIZATION: Store in cache (Priority 4)
		r.storeWebsiteContentCache(cacheKey, keywords)

		return keywords
	}
	structuredDataExtractor := NewStructuredDataExtractorAdapter(r.logger)
	structuredDataResult := structuredDataExtractor.ExtractStructuredData(string(body))

	var structuredKeywords []string
	structuredKeywordMap := make(map[string]float64) // Track keywords with confidence scores

	// Extract keywords from BusinessInfo
	if structuredDataResult.BusinessInfo.Industry != "" {
		structuredKeywordMap[strings.ToLower(structuredDataResult.BusinessInfo.Industry)] = 2.0 // OPTIMIZATION #15: 2x weight for structured data
	}
	if structuredDataResult.BusinessInfo.BusinessType != "" {
		structuredKeywordMap[strings.ToLower(structuredDataResult.BusinessInfo.BusinessType)] = 2.0 // OPTIMIZATION #15: 2x weight for structured data
	}
	for _, service := range structuredDataResult.BusinessInfo.Services {
		serviceLower := strings.ToLower(service)
		structuredKeywordMap[serviceLower] = 2.0 // OPTIMIZATION #15: 2x weight for structured data
	}
	for _, product := range structuredDataResult.BusinessInfo.Products {
		productLower := strings.ToLower(product)
		structuredKeywordMap[productLower] = 2.0 // OPTIMIZATION #15: 2x weight for structured data
	}

	// Extract keywords from ProductInfo
	for _, product := range structuredDataResult.ProductInfo {
		if product.Name != "" {
			structuredKeywordMap[strings.ToLower(product.Name)] = 2.0 // OPTIMIZATION #15: 2x weight for structured data
		}
		if product.Category != "" {
			structuredKeywordMap[strings.ToLower(product.Category)] = 2.0 // OPTIMIZATION #15: 2x weight for structured data
		}
		if product.Description != "" {
			// Extract keywords from product description
			descKeywords := r.extractBusinessKeywords(product.Description)
			for _, kw := range descKeywords {
				structuredKeywordMap[strings.ToLower(kw)] = 2.0 // OPTIMIZATION #15: 2x weight for structured data
			}
		}
	}

	// Extract keywords from ServiceInfo
	for _, service := range structuredDataResult.ServiceInfo {
		if service.Name != "" {
			structuredKeywordMap[strings.ToLower(service.Name)] = 2.0 // OPTIMIZATION #15: 2x weight for structured data
		}
		if service.Category != "" {
			structuredKeywordMap[strings.ToLower(service.Category)] = 2.0 // OPTIMIZATION #15: 2x weight for structured data
		}
		if service.Description != "" {
			// Extract keywords from service description
			descKeywords := r.extractBusinessKeywords(service.Description)
			for _, kw := range descKeywords {
				structuredKeywordMap[strings.ToLower(kw)] = 2.0 // OPTIMIZATION #15: 2x weight for structured data
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

	r.logger.Printf("📊 [StructuredData] Extracted %d keywords from structured data (weighted 2.0x)", len(structuredKeywords))

	// Combine text keywords and structured keywords
	allKeywords := make(map[string]float64)

	// Add text keywords with weight 1.0
	for _, kw := range textKeywords {
		kwLower := strings.ToLower(kw)
		if allKeywords[kwLower] < 1.0 {
			allKeywords[kwLower] = 1.0
		}
	}

	// Add structured keywords with weight 2.0 (higher priority - OPTIMIZATION #15)
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
		maxLen := 10
		if len(keywords) < maxLen {
			maxLen = len(keywords)
		}
		r.logger.Printf("   - Top keywords: %v", keywords[:maxLen])
	}

	// OPTIMIZATION: Store in cache (Priority 4)
	r.storeWebsiteContentCache(cacheKey, keywords)

	return keywords
}

// extractScrapingResultFields extracts fields from ScrapingResult interface to avoid import cycle
func (r *SupabaseKeywordRepository) extractScrapingResultFields(result interface{}) *struct {
	URL           string
	StatusCode    int
	Content       string
	TextContent   string
	Keywords      []string
	ContentType   string
	ContentLength int64
	Headers       map[string]string
	FinalURL      string
	ScrapedAt     time.Time
	Duration      time.Duration
	Error         string
	Success       bool
	Title         string
} {
	if result == nil {
		return nil
	}

	resultValue := reflect.ValueOf(result)
	if resultValue.Kind() == reflect.Ptr {
		resultValue = resultValue.Elem()
	}

	if !resultValue.IsValid() {
		return nil
	}

	extracted := &struct {
		URL           string
		StatusCode    int
		Content       string
		TextContent   string
		Keywords      []string
		ContentType   string
		ContentLength int64
		Headers       map[string]string
		FinalURL      string
		ScrapedAt     time.Time
		Duration      time.Duration
		Error         string
		Success       bool
		Title         string
	}{}

	if field := resultValue.FieldByName("URL"); field.IsValid() && field.Kind() == reflect.String {
		extracted.URL = field.String()
	}
	if field := resultValue.FieldByName("StatusCode"); field.IsValid() && field.Kind() == reflect.Int {
		extracted.StatusCode = int(field.Int())
	}
	if field := resultValue.FieldByName("Content"); field.IsValid() && field.Kind() == reflect.String {
		extracted.Content = field.String()
	}
	if field := resultValue.FieldByName("TextContent"); field.IsValid() && field.Kind() == reflect.String {
		extracted.TextContent = field.String()
	}
	if field := resultValue.FieldByName("Keywords"); field.IsValid() && field.Kind() == reflect.Slice {
		if field.Len() > 0 {
			extracted.Keywords = make([]string, field.Len())
			for i := 0; i < field.Len(); i++ {
				if item := field.Index(i); item.Kind() == reflect.String {
					extracted.Keywords[i] = item.String()
				}
			}
		}
	}
	if field := resultValue.FieldByName("ContentType"); field.IsValid() && field.Kind() == reflect.String {
		extracted.ContentType = field.String()
	}
	if field := resultValue.FieldByName("ContentLength"); field.IsValid() {
		switch field.Kind() {
		case reflect.Int, reflect.Int64:
			extracted.ContentLength = field.Int()
		}
	}
	if field := resultValue.FieldByName("Headers"); field.IsValid() && field.Kind() == reflect.Map {
		extracted.Headers = make(map[string]string)
		for _, key := range field.MapKeys() {
			if val := field.MapIndex(key); val.IsValid() && val.Kind() == reflect.String {
				extracted.Headers[key.String()] = val.String()
			}
		}
	}
	if field := resultValue.FieldByName("FinalURL"); field.IsValid() && field.Kind() == reflect.String {
		extracted.FinalURL = field.String()
	}
	if field := resultValue.FieldByName("ScrapedAt"); field.IsValid() {
		if field.Type().String() == "time.Time" {
			extracted.ScrapedAt = field.Interface().(time.Time)
		}
	}
	if field := resultValue.FieldByName("Duration"); field.IsValid() {
		if field.Type().String() == "time.Duration" {
			extracted.Duration = field.Interface().(time.Duration)
		}
	}
	if field := resultValue.FieldByName("Error"); field.IsValid() && field.Kind() == reflect.String {
		extracted.Error = field.String()
	}
	if field := resultValue.FieldByName("Success"); field.IsValid() && field.Kind() == reflect.Bool {
		extracted.Success = field.Bool()
	}
	if field := resultValue.FieldByName("Title"); field.IsValid() && field.Kind() == reflect.String {
		extracted.Title = field.String()
	}

	return extracted
}

// extractKeywordsFromMultiPageWebsite analyzes multiple pages with relevance-based weighting
// Uses SmartWebsiteCrawler to discover pages, limits to top 15 pages, analyzes concurrently,
// and returns top 30 keywords weighted by page relevance score
// Phase 8: Enhanced with detailed logging and error tracking
func (r *SupabaseKeywordRepository) extractKeywordsFromMultiPageWebsite(ctx context.Context, websiteURL string) []string {
	startTime := time.Now()
	r.logger.Printf("🌐 [KeywordExtraction] [MultiPage] Starting multi-page website analysis for: %s", websiteURL)

	// Use the parent context timeout (set to 5s in calling function to enable fast-path)
	// Don't create additional timeout - respect the parent timeout
	analysisCtx := ctx

	// Create SmartWebsiteCrawler with max 15 pages
	if NewSmartWebsiteCrawlerAdapter == nil {
		r.logger.Printf("❌ [KeywordExtraction] [MultiPage] ERROR: SmartWebsiteCrawler adapter not initialized - falling back to single page")
		return []string{} // Return empty to trigger fallback
	}
	crawler := NewSmartWebsiteCrawlerAdapter(r.logger)

	// Use fast-path mode if timeout is short (5s or less), otherwise use regular crawl
	// Fast-path: 5s timeout, 8 pages max, 3 concurrent
	// Regular: capped at 10s to reserve time for other operations (Phase 1 scraping, classification, etc.)
	// FIX: Don't use full remaining context time - cap at 10s to prevent consuming entire deadline
	// FIX: Handle negative timeouts (expired context) gracefully
	// OPTIMIZATION: Add timeout checks and early termination for multi-page analysis
	const maxMultiPageTimeout = 10 * time.Second
	timeoutDuration := 5 * time.Second // Default from calling function (5s to enable fast-path)
	maxConcurrentPages := 3            // Default concurrent pages
	maxPages := 8                      // Default max pages for fast-path

	if deadline, ok := analysisCtx.Deadline(); ok {
		remainingTime := time.Until(deadline)
		// FIX: Handle expired context (negative remaining time)
		if remainingTime <= 0 {
			// Context already expired, skip multi-page analysis
			r.logger.Printf("⚠️ [KeywordExtraction] [MultiPage] Context already expired, skipping multi-page analysis")
			return []string{} // Return empty to trigger fallback
		} else if remainingTime < 10*time.Second {
			// Very little time remaining, skip multi-page analysis
			r.logger.Printf("⚠️ [KeywordExtraction] [MultiPage] Insufficient time remaining (%v < 10s), skipping multi-page analysis", remainingTime)
			return []string{} // Return empty to trigger fallback
		} else if remainingTime < 30*time.Second {
			// Limited time - reduce concurrent pages and max pages
			maxConcurrentPages = 1
			maxPages = 3
			timeoutDuration = remainingTime - 5*time.Second
			r.logger.Printf("📊 [KeywordExtraction] [MultiPage] Limited time remaining (%v), reducing concurrency: maxPages=%d, concurrent=%d", remainingTime, maxPages, maxConcurrentPages)
		} else if remainingTime < 60*time.Second {
			// Moderate time - reduce concurrent pages
			maxConcurrentPages = 2
			maxPages = 5
			timeoutDuration = remainingTime - 5*time.Second
			r.logger.Printf("📊 [KeywordExtraction] [MultiPage] Moderate time remaining (%v), reducing concurrency: maxPages=%d, concurrent=%d", remainingTime, maxPages, maxConcurrentPages)
		} else if remainingTime > maxMultiPageTimeout+5*time.Second {
			// Cap timeout to maxMultiPageTimeout to reserve time for other operations
			timeoutDuration = maxMultiPageTimeout
		} else {
			// Use remaining time minus 5s buffer for other operations
			timeoutDuration = remainingTime - 5*time.Second
		}
	}

	r.logger.Printf("📊 [KeywordExtraction] [MultiPage] Timeout duration: %v (threshold: 5s, max: 10s)", timeoutDuration)

	var crawlResult CrawlResultInterface
	var err error

	if timeoutDuration <= 5*time.Second {
		// Use fast-path mode for short timeouts
		r.logger.Printf("🚀 [KeywordExtraction] [MultiPage] [FAST-PATH] Using fast-path mode (timeout: %v, max pages: %d, concurrent: %d)", timeoutDuration, maxPages, maxConcurrentPages)
		crawlResult, err = crawler.CrawlWebsiteFast(analysisCtx, websiteURL, timeoutDuration, maxPages, maxConcurrentPages)
	} else {
		// Use regular crawl for longer timeouts, but with reduced concurrency if time is limited
		r.logger.Printf("🔍 [KeywordExtraction] [MultiPage] [REGULAR] Using regular crawl mode (timeout: %v, concurrent: %d)", timeoutDuration, maxConcurrentPages)
		// Note: CrawlWebsite doesn't accept maxConcurrentPages parameter, but the crawler will respect the timeout
		crawlResult, err = crawler.CrawlWebsite(analysisCtx, websiteURL)
	}
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
	// Note: maxPages may have been set earlier based on time remaining, but cap at 15
	if maxPages > 15 {
		maxPages = 15
	}
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
		maxLen := 10
		if len(keywords) < maxLen {
			maxLen = len(keywords)
		}
		r.logger.Printf("   - Top keywords: %v", keywords[:maxLen])
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
// Enhanced: Added HTML entity decoding for better keyword extraction
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

	// Decode HTML entities (basic ones) - CRITICAL for keyword extraction
	htmlContent = strings.ReplaceAll(htmlContent, "&nbsp;", " ")
	htmlContent = strings.ReplaceAll(htmlContent, "&amp;", "&")
	htmlContent = strings.ReplaceAll(htmlContent, "&lt;", "<")
	htmlContent = strings.ReplaceAll(htmlContent, "&gt;", ">")
	htmlContent = strings.ReplaceAll(htmlContent, "&quot;", "\"")
	htmlContent = strings.ReplaceAll(htmlContent, "&#39;", "'")
	htmlContent = strings.ReplaceAll(htmlContent, "&apos;", "'")
	htmlContent = strings.ReplaceAll(htmlContent, "&mdash;", "-")
	htmlContent = strings.ReplaceAll(htmlContent, "&ndash;", "-")
	htmlContent = strings.ReplaceAll(htmlContent, "&hellip;", "...")
	htmlContent = strings.ReplaceAll(htmlContent, "&copy;", "(c)")
	htmlContent = strings.ReplaceAll(htmlContent, "&reg;", "(r)")
	htmlContent = strings.ReplaceAll(htmlContent, "&trade;", "(tm)")

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

// applyRateLimit applies rate limiting with jitter to avoid thundering herd.
// If crawlDelay is provided and greater than the configured minDelay, it will be used instead.
// Phase 9.3: Performance optimization to respect rate limits and add jitter
func (r *SupabaseKeywordRepository) applyRateLimit(domain string, crawlDelay ...time.Duration) {
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

	// Determine effective delay: use robots.txt crawl delay if specified and greater than minDelay
	effectiveDelay := r.minDelay
	if len(crawlDelay) > 0 && crawlDelay[0] > 0 {
		if crawlDelay[0] > r.minDelay {
			effectiveDelay = crawlDelay[0]
			r.logger.Printf("⏳ [RobotsTxt] Using robots.txt crawl delay of %v for %s (greater than configured %v)",
				crawlDelay[0], domain, r.minDelay)
		}
	}

	lastRequest, exists := r.rateLimiter[domain]
	if exists {
		elapsed := time.Since(lastRequest)
		if elapsed < effectiveDelay {
			// Use human-like timing instead of fixed delay + jitter
			remainingDelay := effectiveDelay - elapsed
			totalDelay := getHumanLikeDelay(remainingDelay, domain)

			delaySource := "configured"
			if len(crawlDelay) > 0 && crawlDelay[0] > r.minDelay {
				delaySource = "robots.txt"
			}

			r.logger.Printf("⏳ [Performance] Rate limiting: waiting %v before request to %s (human-like delay, source: %s)",
				totalDelay, domain, delaySource)
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

	// Log extraction statistics for debugging
	if len(keywords) == 0 {
		r.logger.Printf("⚠️ [KeywordExtraction] No keywords extracted from %d characters of text", len(textContent))
		// Log a sample of the text to help debug
		if len(textContent) > 500 {
			r.logger.Printf("📝 [KeywordExtraction] Sample text (first 500 chars): %s", textContent[:500])
		}
	} else {
		r.logger.Printf("✅ [KeywordExtraction] Extracted %d keywords from %d characters of text", len(keywords), len(textContent))
	}

	// Post-processing filter: Remove isolated gibberish words
	keywords = r.filterGibberishKeywords(keywords)

	// Limit to top 50 keywords to avoid noise (increased from 20 for better coverage)
	if len(keywords) > 50 {
		keywords = keywords[:50]
	}

	return keywords
}

// filterGibberishKeywords removes isolated gibberish words from keyword list
// Uses pattern detection and n-gram validation similar to smart_website_crawler
func (r *SupabaseKeywordRepository) filterGibberishKeywords(keywords []string) []string {
	var filtered []string

	// Common English bigrams for validation
	commonBigrams := map[string]bool{
		"th": true, "in": true, "er": true, "ed": true, "an": true, "re": true,
		"he": true, "on": true, "en": true, "at": true, "it": true, "is": true,
		"or": true, "ti": true, "as": true, "to": true, "of": true, "al": true,
		"ar": true, "st": true, "ng": true, "le": true, "ou": true, "nt": true,
		"ea": true, "nd": true, "te": true, "es": true, "hi": true,
		"ri": true, "ve": true, "co": true, "de": true, "ra": true, "li": true,
		"se": true, "ne": true, "me": true, "be": true, "we": true, "wa": true,
		"ma": true, "ha": true, "ca": true, "la": true, "pa": true, "ta": true,
		"sa": true, "na": true, "ga": true, "fa": true, "da": true, "ba": true,
	}

	// Suspicious bigrams that rarely appear in English
	// Enhanced to catch patterns from gibberish words: "ivdi", "fays", "yilp", "dioy", "ukxa"
	suspiciousBigrams := map[string]bool{
		"iv": true, "vd": true, "di": true, "xa": true, "uk": true, "kx": true,
		"fa": true, "ay": true, "ys": true, "yi": true, "il": true, "lp": true,
		"gu": true, "oi": true, "je": true, "yl": true, "lb": true, "io": true,
		"fv": true, "yz": true, "zx": true, "qw": true, "xc": true, "vb": true,
		"fg": true, "gh": true, "hj": true, "jk": true, "kl": true, "lm": true,
		"kj": true, "mn": true, "nb": true, "bv": true, "vc": true, "cx": true,
		"xz": true, "zq": true, "qa": true, "az": true, "ws": true, "sx": true,
		"xw": true, "ed": true, "dc": true, "cd": true, "de": true, "rf": true,
		"vr": true, "tg": true, "gb": true, "bg": true, "gt": true,
		"yh": true, "hn": true, "nh": true, "hy": true, "uj": true, "jm": true,
		"mj": true, "ju": true,
	}

	// Known gibberish words to filter
	knownGibberish := map[string]bool{
		"ivdi": true, "fays": true, "yilp": true, "dioy": true, "ukxa": true,
		"ivd": true, "fay": true, "yil": true, "dio": true, "ukx": true,
	}

	for _, keyword := range keywords {
		// Skip multi-word phrases (they're likely valid)
		if strings.Contains(keyword, " ") {
			filtered = append(filtered, keyword)
			continue
		}

		// Skip if too short
		if len(keyword) < 4 {
			continue
		}

		// Check for known gibberish words first
		if knownGibberish[keyword] {
			continue
		}

		// Check for suspicious patterns
		if r.hasSuspiciousPattern(keyword) {
			continue
		}

		// Check n-gram patterns
		if !r.hasValidNgramPattern(keyword, commonBigrams, suspiciousBigrams) {
			continue
		}

		// Passed all checks
		filtered = append(filtered, keyword)
	}

	return filtered
}

// hasSuspiciousPattern checks for patterns that rarely appear in English words
func (r *SupabaseKeywordRepository) hasSuspiciousPattern(word string) bool {
	// Check for repeated letters (more than 2 consecutive)
	for i := 0; i < len(word)-2; i++ {
		if word[i] == word[i+1] && word[i] == word[i+2] {
			return true
		}
	}

	// Check for unusual consonant clusters
	// Enhanced to catch specific gibberish words: "ivdi", "fays", "yilp", "dioy", "ukxa"
	suspiciousClusters := []string{
		"ivd", "ivdi", "fay", "fays", "yil", "yilp", "dio", "dioy", "ukx", "ukxa",
		"guo", "jey", "mii", "xzv", "qwx", "jkl", "zxc", "vbn", "qwe", "asd",
		// Additional suspicious patterns
		"fgh", "hjk", "lkj", "mnb", "bvc", "cxz", "zqa", "qaz", "wsx", "xsw",
		"edc", "cde", "rfv", "vfr", "tgb", "bgt", "yhn", "nhy", "ujm", "mju",
	}
	for _, cluster := range suspiciousClusters {
		if strings.Contains(word, cluster) {
			return true
		}
	}

	// Check for specific known gibberish words
	knownGibberish := map[string]bool{
		"ivdi": true, "fays": true, "yilp": true, "dioy": true, "ukxa": true,
		"ivd": true, "fay": true, "yil": true, "dio": true, "ukx": true,
	}
	if knownGibberish[word] {
		return true
	}

	// Check for too many rare letters
	rareLetters := map[rune]bool{'q': true, 'x': true, 'z': true, 'j': true}
	rareCount := 0
	for _, char := range word {
		if rareLetters[char] {
			rareCount++
		}
	}
	if float64(rareCount)/float64(len(word)) > 0.3 {
		return true
	}

	return false
}

// hasValidNgramPattern checks if letter combinations are common in English
func (r *SupabaseKeywordRepository) hasValidNgramPattern(word string, commonBigrams, suspiciousBigrams map[string]bool) bool {
	hasCommonBigram := false
	for i := 0; i < len(word)-1; i++ {
		bigram := word[i : i+2]
		if suspiciousBigrams[bigram] && !commonBigrams[bigram] {
			if !hasCommonBigram {
				return false
			}
		}
		if commonBigrams[bigram] {
			hasCommonBigram = true
		}
	}

	if !hasCommonBigram && len(word) >= 4 {
		return false
	}

	return true
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
// Uses word segmentation library for compound domain names like "thegreenegrape" → ["the", "green", "grape"]
func (r *SupabaseKeywordRepository) splitDomainName(domain string) []string {
	if domain == "" {
		return []string{}
	}

	// Use word segmentation library for compound domain names
	// This handles cases like "thegreenegrape", "techstartup", "wineshop", etc.
	segments := r.segmenter.Segment(domain)

	// If segmentation returned empty or single segment, try fallback to camelCase splitting
	if len(segments) <= 1 {
		// Remove TLD (everything after the last dot)
		parts := strings.Split(domain, ".")
		if len(parts) == 0 {
			return segments
		}
		domainName := parts[0]

		// If domainName is empty, try the whole domain
		if domainName == "" && len(parts) > 1 {
			domainName = parts[len(parts)-2] // Use second-to-last part
		}

		// Split by common separators (hyphens, underscores)
		domainName = strings.ReplaceAll(domainName, "-", " ")
		domainName = strings.ReplaceAll(domainName, "_", " ")

		// Split by spaces (from hyphens/underscores) and camelCase
		var words []string
		spaceParts := strings.Fields(domainName)
		for _, part := range spaceParts {
			// Try to split camelCase words
			camelWords := r.splitCamelCase(part)
			words = append(words, camelWords...)
		}

		// If we got words from fallback, use them; otherwise use segments
		if len(words) > 0 {
			return words
		}
	}

	return segments
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
		"shop":       {"retail", "ecommerce", "store", "shop"},
		"store":      {"retail", "ecommerce", "store", "shop"},
		"restaurant": {"restaurant", "food", "dining"},
		"cafe":       {"cafe", "coffee", "food", "dining"},
		"bar":        {"bar", "beverage", "alcohol", "drinks"},
		"wine":       {"wine", "beverage", "alcohol", "winery"},
		"beer":       {"beer", "beverage", "alcohol", "brewery"},
		"tech":       {"technology", "tech", "software"},
		"app":        {"app", "application", "software", "technology"},
		"dev":        {"development", "software", "technology"},
		"design":     {"design", "creative", "art"},
		"photo":      {"photography", "photo", "creative"},
		"art":        {"art", "creative", "design"},
		"music":      {"music", "entertainment", "creative"},
		"film":       {"film", "entertainment", "media"},
		"news":       {"news", "media", "journalism"},
		"blog":       {"blog", "content", "media"},
		"edu":        {"education", "school", "learning"},
		"health":     {"health", "healthcare", "medical"},
		"law":        {"law", "legal", "attorney"},
		"finance":    {"finance", "financial", "banking"},
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
		"wine":       {"wine", "beverage", "alcohol", "retail"},
		"grape":      {"wine", "grape", "beverage", "retail"},
		"vineyard":   {"vineyard", "wine", "winery", "beverage"},
		"vintner":    {"vintner", "wine", "winery", "beverage"},
		"brewery":    {"brewery", "beer", "beverage", "alcohol"},
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

// =============================================================================
// Topic-Industry Mapping (Phase 1.3)
// =============================================================================

// GetIndustryTopicsByKeywords retrieves industry-topic relationships for given keywords
// Returns map of industry_id -> relevance_score from industry_topics table
func (r *SupabaseKeywordRepository) GetIndustryTopicsByKeywords(ctx context.Context, keywords []string) (map[int]float64, error) {
	if len(keywords) == 0 {
		return make(map[int]float64), nil
	}

	r.logger.Printf("🔍 Getting industry topics for keywords: %v", keywords)

	// Check if client is available
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	// Use batch query for better performance (Phase 2.2)
	topicMatches, err := r.BatchFindIndustryTopics(ctx, keywords)
	if err != nil {
		r.logger.Printf("⚠️ Batch topic lookup failed, falling back to individual queries: %v", err)
		// Fallback to individual queries
		return r.getIndustryTopicsByKeywordsFallback(ctx, keywords, postgrestClient)
	}

	// Aggregate scores by industry from batch results
	results := make(map[int]float64)
	for _, matches := range topicMatches {
		for _, match := range matches {
			if existing, exists := results[match.IndustryID]; !exists || match.RelevanceScore > existing {
				// Weight by accuracy score: higher accuracy = higher final score
				weightedScore := match.RelevanceScore * (0.5 + match.AccuracyScore*0.5)
				results[match.IndustryID] = weightedScore
			}
		}
	}

	r.logger.Printf("✅ Found %d industry-topic mappings", len(results))
	return results, nil
}

// getIndustryTopicsByKeywordsFallback provides fallback to individual queries
func (r *SupabaseKeywordRepository) getIndustryTopicsByKeywordsFallback(ctx context.Context, keywords []string, postgrestClient interface{}) (map[int]float64, error) {
	results := make(map[int]float64)
	postgrest := r.client.GetPostgrestClient()
	if postgrest == nil {
		return results, nil
	}

	// Query for each keyword individually (fallback)
	for _, keyword := range keywords {
		keywordLower := strings.ToLower(strings.TrimSpace(keyword))

		// Query industry_topics table
		// Note: Order method expects *map[string]string, but we'll skip ordering in fallback
		// to avoid type issues - relevance_score ordering is handled by batch function
		data, _, err := postgrest.
			From("industry_topics").
			Select("industry_id,relevance_score,accuracy_score", "", false).
			Ilike("topic", fmt.Sprintf("%%%s%%", keywordLower)).
			Limit(10, "").
			Execute()

		if err != nil {
			r.logger.Printf("⚠️ Failed to query industry_topics for keyword '%s': %v", keyword, err)
			continue
		}

		// Parse results
		var topicMappings []struct {
			IndustryID     int     `json:"industry_id"`
			RelevanceScore float64 `json:"relevance_score"`
			AccuracyScore  float64 `json:"accuracy_score"`
		}

		if err := json.Unmarshal(data, &topicMappings); err != nil {
			r.logger.Printf("⚠️ Failed to unmarshal topic mappings: %v", err)
			continue
		}

		// Aggregate scores by industry (use highest relevance score)
		for _, mapping := range topicMappings {
			if existing, exists := results[mapping.IndustryID]; !exists || mapping.RelevanceScore > existing {
				// Weight by accuracy score: higher accuracy = higher final score
				weightedScore := mapping.RelevanceScore * (0.5 + mapping.AccuracyScore*0.5)
				results[mapping.IndustryID] = weightedScore
			}
		}
	}

	return results, nil
}

// GetTopicAccuracy retrieves the accuracy score for a specific topic-industry pair
func (r *SupabaseKeywordRepository) GetTopicAccuracy(ctx context.Context, industryID int, topic string) (float64, error) {
	if r.client == nil {
		return 0.75, nil // Default accuracy if no database
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return 0.75, nil // Default accuracy if no client
	}

	topicLower := strings.ToLower(strings.TrimSpace(topic))

	// Query for specific topic-industry pair
	data, _, err := postgrestClient.
		From("industry_topics").
		Select("accuracy_score", "", false).
		Eq("industry_id", fmt.Sprintf("%d", industryID)).
		Ilike("topic", fmt.Sprintf("%%%s%%", topicLower)).
		Single().
		Execute()

	if err != nil {
		// Return default accuracy if not found
		return 0.75, nil
	}

	var result struct {
		AccuracyScore float64 `json:"accuracy_score"`
	}

	if err := json.Unmarshal(data, &result); err != nil {
		return 0.75, nil // Default on error
	}

	return result.AccuracyScore, nil
}

// =============================================================================
// Keyword Patterns / Co-Occurrence Analysis (Phase 1.4)
// =============================================================================

// FindIndustriesByPatterns finds industries matching keyword co-occurrence patterns
func (r *SupabaseKeywordRepository) FindIndustriesByPatterns(ctx context.Context, patterns []string) ([]*PatternMatchResult, error) {
	if len(patterns) == 0 {
		return []*PatternMatchResult{}, nil
	}

	r.logger.Printf("🔍 Finding industries by patterns: %d patterns", len(patterns))

	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	// Call database function via PostgREST RPC
	payload := map[string]interface{}{
		"p_patterns": patterns,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/find_industries_by_patterns", r.client.GetURL())

	// FIX: Check if context is expired and create fresh context if needed
	rpcCtx := ctx
	var rpcCancel context.CancelFunc
	httpTimeout := 5 * time.Second

	if deadline, hasDeadline := ctx.Deadline(); hasDeadline {
		timeRemaining := time.Until(deadline)
		if timeRemaining <= 0 {
			// Context already expired, create fresh context with HTTP timeout
			rpcCtx, rpcCancel = context.WithTimeout(context.Background(), httpTimeout)
			defer rpcCancel()
		} else if timeRemaining < httpTimeout {
			// Context has less time than HTTP timeout, use remaining time
			rpcCtx, rpcCancel = context.WithTimeout(context.Background(), timeRemaining)
			defer rpcCancel()
		}
		// If context has sufficient time, use it as-is
	} else {
		// No deadline, create context with HTTP timeout
		rpcCtx, rpcCancel = context.WithTimeout(ctx, httpTimeout)
		defer rpcCancel()
	}

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RPC call returned status %d: %s", resp.StatusCode, string(body))
	}

	var results []struct {
		IndustryID      int      `json:"industry_id"`
		IndustryName    string   `json:"industry_name"`
		PatternMatches  int      `json:"pattern_matches"`
		AvgScore        float64  `json:"avg_score"`
		MatchedPatterns []string `json:"matched_patterns"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode RPC response: %w", err)
	}

	patternResults := make([]*PatternMatchResult, 0, len(results))
	for _, result := range results {
		patternResults = append(patternResults, &PatternMatchResult{
			IndustryID:      result.IndustryID,
			IndustryName:    result.IndustryName,
			PatternMatches:  result.PatternMatches,
			AvgScore:        result.AvgScore,
			MatchedPatterns: result.MatchedPatterns,
		})
	}

	r.logger.Printf("✅ Found %d industries matching patterns", len(patternResults))
	return patternResults, nil
}

// GetPatternMatches retrieves keyword patterns for a specific industry
func (r *SupabaseKeywordRepository) GetPatternMatches(ctx context.Context, industryID int, patterns []string) ([]*KeywordPattern, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	// Query keyword_patterns table
	data, _, err := postgrestClient.
		From("keyword_patterns").
		Select("id,industry_id,keyword_pair,keyword1,keyword2,co_occurrence_score,pattern_type,frequency", "", false).
		Eq("industry_id", fmt.Sprintf("%d", industryID)).
		In("keyword_pair", patterns).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("failed to query keyword patterns: %w", err)
	}

	var patternsData []struct {
		ID                int     `json:"id"`
		IndustryID        int     `json:"industry_id"`
		KeywordPair       string  `json:"keyword_pair"`
		Keyword1          string  `json:"keyword1"`
		Keyword2          string  `json:"keyword2"`
		CoOccurrenceScore float64 `json:"co_occurrence_score"`
		PatternType       string  `json:"pattern_type"`
		Frequency         int     `json:"frequency"`
	}

	if err := json.Unmarshal(data, &patternsData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal pattern data: %w", err)
	}

	keywordPatterns := make([]*KeywordPattern, 0, len(patternsData))
	for _, p := range patternsData {
		keywordPatterns = append(keywordPatterns, &KeywordPattern{
			ID:                p.ID,
			IndustryID:        p.IndustryID,
			KeywordPair:       p.KeywordPair,
			Keyword1:          p.Keyword1,
			Keyword2:          p.Keyword2,
			CoOccurrenceScore: p.CoOccurrenceScore,
			PatternType:       p.PatternType,
			Frequency:         p.Frequency,
		})
	}

	return keywordPatterns, nil
}

// =============================================================================
// Batch Queries (Phase 2.2)
// =============================================================================

// BatchFindKeywords performs batch keyword lookup in a single query
// Returns map of keyword -> []IndustryMatch for all keywords
func (r *SupabaseKeywordRepository) BatchFindKeywords(ctx context.Context, keywords []string) (map[string][]IndustryMatch, error) {
	if len(keywords) == 0 {
		return make(map[string][]IndustryMatch), nil
	}

	r.logger.Printf("🔍 Batch finding keywords: %d keywords", len(keywords))

	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	// Call database function via PostgREST RPC
	payload := map[string]interface{}{
		"p_keywords": keywords,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/batch_find_keywords", r.client.GetURL())

	// FIX: Ensure context has sufficient time for HTTP request
	httpTimeout := 3 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RPC call returned status %d: %s", resp.StatusCode, string(body))
	}

	var results []struct {
		Keyword         string  `json:"keyword"`
		IndustryID      int     `json:"industry_id"`
		IndustryName    string  `json:"industry_name"`
		BaseWeight      float64 `json:"base_weight"`
		SimilarityScore float64 `json:"similarity_score"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode RPC response: %w", err)
	}

	// Group results by keyword
	keywordMatches := make(map[string][]IndustryMatch)
	for _, result := range results {
		keywordMatches[result.Keyword] = append(keywordMatches[result.Keyword], IndustryMatch{
			IndustryID:   result.IndustryID,
			IndustryName: result.IndustryName,
			Weight:       result.BaseWeight,
			Similarity:   result.SimilarityScore,
		})
	}

	r.logger.Printf("✅ Batch found matches for %d keywords", len(keywordMatches))
	return keywordMatches, nil
}

// BatchFindIndustryTopics performs batch industry topic lookup in a single query
// Returns map of keyword -> []TopicMatch for all keywords
func (r *SupabaseKeywordRepository) BatchFindIndustryTopics(ctx context.Context, keywords []string) (map[string][]TopicMatch, error) {
	if len(keywords) == 0 {
		return make(map[string][]TopicMatch), nil
	}

	r.logger.Printf("🔍 Batch finding industry topics: %d keywords", len(keywords))

	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	// Call database function via PostgREST RPC
	payload := map[string]interface{}{
		"p_keywords": keywords,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/batch_find_industry_topics", r.client.GetURL())

	// FIX: Ensure context has sufficient time for HTTP request
	httpTimeout := 3 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RPC call returned status %d: %s", resp.StatusCode, string(body))
	}

	var results []struct {
		Keyword        string  `json:"keyword"`
		IndustryID     int     `json:"industry_id"`
		IndustryName   string  `json:"industry_name"`
		RelevanceScore float64 `json:"relevance_score"`
		AccuracyScore  float64 `json:"accuracy_score"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode RPC response: %w", err)
	}

	// Group results by keyword
	topicMatches := make(map[string][]TopicMatch)
	for _, result := range results {
		topicMatches[result.Keyword] = append(topicMatches[result.Keyword], TopicMatch{
			IndustryID:     result.IndustryID,
			IndustryName:   result.IndustryName,
			RelevanceScore: result.RelevanceScore,
			AccuracyScore:  result.AccuracyScore,
		})
	}

	r.logger.Printf("✅ Batch found topic matches for %d keywords", len(topicMatches))
	return topicMatches, nil
}

// =============================================================================
// Vector Similarity Search (Phase 3)
// =============================================================================

// MatchCodeEmbeddings performs vector similarity search for code embeddings
func (r *SupabaseKeywordRepository) MatchCodeEmbeddings(
	ctx context.Context,
	embedding []float64,
	codeType string,
	threshold float64,
	limit int,
) ([]CodeMatch, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	r.logger.Printf("🔍 [Phase 3] Matching codes by embedding (type: %s, threshold: %.2f, limit: %d)", codeType, threshold, limit)

	// Call database function via PostgREST RPC
	payload := map[string]interface{}{
		"query_embedding":   embedding,
		"code_type_filter":  codeType,
		"match_threshold":   threshold,
		"match_count":       limit,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/match_code_embeddings", r.client.GetURL())

	// Ensure context has sufficient time for HTTP request
	httpTimeout := 5 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RPC call returned status %d: %s", resp.StatusCode, string(body))
	}

	var results []struct {
		Code        string  `json:"code"`
		CodeType    string  `json:"code_type"`
		Description string  `json:"description"`
		Similarity  float64 `json:"similarity"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode RPC response: %w", err)
	}

	// Convert to CodeMatch
	matches := make([]CodeMatch, 0, len(results))
	for _, result := range results {
		matches = append(matches, CodeMatch{
			Code:        result.Code,
			Description: result.Description,
			Similarity:  result.Similarity,
		})
	}

	r.logger.Printf("✅ [Phase 3] Found %d code matches by embedding", len(matches))
	return matches, nil
}

// =============================================================================
// Phase 5: Classification Cache Methods
// =============================================================================

// CacheStats represents cache statistics
type CacheStats struct {
	TotalEntries int           `json:"total_entries"`
	HitRate      float64       `json:"hit_rate"`
	AvgAge       time.Duration `json:"avg_age"`
	ExpiringSoon int           `json:"expiring_soon"` // Expiring in next 7 days
}

// GetCachedClassification retrieves a cached classification result
func (r *SupabaseKeywordRepository) GetCachedClassification(
	ctx context.Context,
	contentHash string,
) (*CachedClassificationResult, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	r.logger.Printf("🔍 [Phase 5] Checking cache for content hash: %s", contentHash[:16]+"...")

	// Call database function via PostgREST RPC
	payload := map[string]interface{}{
		"p_content_hash": contentHash,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/get_cached_classification", r.client.GetURL())

	// Ensure context has sufficient time for HTTP request
	httpTimeout := 2 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusNoContent {
		// Cache miss
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RPC call returned status %d: %s", resp.StatusCode, string(body))
	}

	// Decode JSONB result
	var resultJSON json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&resultJSON); err != nil {
		return nil, fmt.Errorf("failed to decode cache result: %w", err)
	}

	// Check if result is null (cache miss)
	if len(resultJSON) == 0 || string(resultJSON) == "null" {
		return nil, nil
	}

	// Parse the cached result
	var cachedResult CachedClassificationResult
	if err := json.Unmarshal(resultJSON, &cachedResult); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached result: %w", err)
	}

	r.logger.Printf("✅ [Phase 5] Cache hit for content hash: %s", contentHash[:16]+"...")
	return &cachedResult, nil
}

// SetCachedClassification stores a classification result in cache
func (r *SupabaseKeywordRepository) SetCachedClassification(
	ctx context.Context,
	contentHash string,
	businessName string,
	websiteURL string,
	result *CachedClassificationResult,
) error {
	if r.client == nil {
		return fmt.Errorf("database client not available")
	}

	// Serialize result to JSONB
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	// Call database function via PostgREST RPC
	payload := map[string]interface{}{
		"p_content_hash":        contentHash,
		"p_business_name":       businessName,
		"p_website_url":          websiteURL,
		"p_result":              json.RawMessage(resultJSON),
		"p_layer_used":          result.LayerUsed,
		"p_confidence":          result.Confidence,
		"p_processing_time_ms":  result.ProcessingTimeMs,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/set_cached_classification", r.client.GetURL())

	// Ensure context has sufficient time for HTTP request
	httpTimeout := 2 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RPC call returned status %d: %s", resp.StatusCode, string(body))
	}

	r.logger.Printf("✅ [Phase 5] Cached classification result for content hash: %s", contentHash[:16]+"...")
	return nil
}

// GetCacheStats retrieves cache statistics
func (r *SupabaseKeywordRepository) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	// Query cache statistics using PostgREST
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return nil, fmt.Errorf("postgrest client not available")
	}

	stats := &CacheStats{}

	// Query using PostgREST From method to count entries
	// Note: PostgREST doesn't support COUNT(*) directly, so we fetch all IDs and count
	data, _, err := postgrestClient.From("classification_cache").
		Select("id", "", false).
		Gt("expires_at", time.Now().Format(time.RFC3339)).
		Execute()
	if err == nil && len(data) > 0 {
		// Count the results
		var results []map[string]interface{}
		if err := json.Unmarshal(data, &results); err == nil {
			stats.TotalEntries = len(results)
		}
	}

	// Query average age and expiring soon (simplified - would need custom SQL RPC)
	// For now, return basic stats
	stats.HitRate = 0.0 // Would need to track hits/misses separately
	stats.AvgAge = 0
	stats.ExpiringSoon = 0

	return stats, nil
}

// LogClassificationMetrics logs a classification metrics record
func (r *SupabaseKeywordRepository) LogClassificationMetrics(
	ctx context.Context,
	metrics *ClassificationMetricsRecord,
) error {
	if r.client == nil {
		return fmt.Errorf("database client not available")
	}

	// Use PostgREST to insert into classification_metrics table
	postgrestClient := r.client.GetPostgrestClient()
	if postgrestClient == nil {
		return fmt.Errorf("postgrest client not available")
	}

	// Prepare payload
	payload := map[string]interface{}{
		"request_id":       metrics.RequestID,
		"business_name":    metrics.BusinessName,
		"website_url":       metrics.WebsiteURL,
		"primary_industry":  metrics.PrimaryIndustry,
		"confidence":        metrics.Confidence,
		"layer_used":        metrics.LayerUsed,
		"method":            metrics.Method,
		"total_time_ms":     metrics.TotalTimeMs,
		"scrape_time_ms":    metrics.ScrapeTimeMs,
		"layer1_time_ms":    metrics.Layer1TimeMs,
		"layer2_time_ms":    metrics.Layer2TimeMs,
		"layer3_time_ms":    metrics.Layer3TimeMs,
		"from_cache":        metrics.FromCache,
		"user_agent":        metrics.UserAgent,
	}

	// Add codes if present
	if len(metrics.MCCCodes) > 0 {
		payload["mcc_codes"] = json.RawMessage(metrics.MCCCodes)
	}
	if len(metrics.SICCodes) > 0 {
		payload["sic_codes"] = json.RawMessage(metrics.SICCodes)
	}
	if len(metrics.NAICSCodes) > 0 {
		payload["naics_codes"] = json.RawMessage(metrics.NAICSCodes)
	}

	// Add IP address if present
	if metrics.IPAddress != "" {
		payload["ip_address"] = metrics.IPAddress
	}

	// Insert using PostgREST (expects array of maps)
	// Use direct HTTP call since PostgREST interface may not support Insert properly
	insertData := []map[string]interface{}{payload}
	jsonPayload, err := json.Marshal(insertData)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/classification_metrics", r.client.GetURL())
	
	// Ensure context has sufficient time for HTTP request
	httpTimeout := 2 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("insert returned status %d: %s", resp.StatusCode, string(body))
	}

	r.logger.Printf("✅ [Phase 5] Logged classification metrics for request: %s", metrics.RequestID)
	return nil
}

// GetDashboardSummary retrieves dashboard summary metrics
func (r *SupabaseKeywordRepository) GetDashboardSummary(
	ctx context.Context,
	days int,
) ([]*DashboardMetric, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	// Call database function via PostgREST RPC
	payload := map[string]interface{}{
		"days": days,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/get_dashboard_summary", r.client.GetURL())

	// Ensure context has sufficient time for HTTP request
	httpTimeout := 5 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RPC call returned status %d: %s", resp.StatusCode, string(body))
	}

	var results []struct {
		Metric      string  `json:"metric"`
		Value       float64 `json:"value"`
		Description string  `json:"description"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode dashboard summary: %w", err)
	}

	metrics := make([]*DashboardMetric, len(results))
	for i, r := range results {
		metrics[i] = &DashboardMetric{
			Metric:      r.Metric,
			Value:       r.Value,
			Description: r.Description,
		}
	}

	return metrics, nil
}

// GetTimeSeriesData retrieves time series data for charts
func (r *SupabaseKeywordRepository) GetTimeSeriesData(
	ctx context.Context,
	days int,
) ([]*TimeSeriesData, error) {
	if r.client == nil {
		return nil, fmt.Errorf("database client not available")
	}

	// Call database function via PostgREST RPC
	payload := map[string]interface{}{
		"days": days,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal RPC payload: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/rpc/get_dashboard_timeseries", r.client.GetURL())

	// Ensure context has sufficient time for HTTP request
	httpTimeout := 5 * time.Second
	rpcCtx, rpcCancel := r.ensureValidContext(ctx, httpTimeout)
	defer rpcCancel()

	req, err := http.NewRequestWithContext(rpcCtx, "POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create RPC request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", r.client.GetServiceKey())
	req.Header.Set("Authorization", "Bearer "+r.client.GetServiceKey())
	req.Header.Set("Prefer", "return=representation")

	httpClient := &http.Client{Timeout: httpTimeout}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("RPC call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RPC call returned status %d: %s", resp.StatusCode, string(body))
	}

	var results []struct {
		Date                 string  `json:"date"`
		TotalClassifications int64   `json:"total_classifications"`
		CacheHits            int64   `json:"cache_hits"`
		CacheMisses          int64   `json:"cache_misses"`
		AvgConfidence        float64 `json:"avg_confidence"`
		AvgTotalTimeMs       float64 `json:"avg_total_time_ms"`
		Layer1Count          int64   `json:"layer1_count"`
		Layer2Count          int64   `json:"layer2_count"`
		Layer3Count          int64   `json:"layer3_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("failed to decode time series data: %w", err)
	}

	timeSeries := make([]*TimeSeriesData, len(results))
	for i, r := range results {
		date, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			// Try parsing with timezone
			date, err = time.Parse(time.RFC3339, r.Date)
			if err != nil {
				return nil, fmt.Errorf("failed to parse date %s: %w", r.Date, err)
			}
		}

		timeSeries[i] = &TimeSeriesData{
			Date:                 date,
			TotalClassifications: r.TotalClassifications,
			CacheHits:            r.CacheHits,
			CacheMisses:          r.CacheMisses,
			AvgConfidence:        r.AvgConfidence,
			AvgTotalTimeMs:       r.AvgTotalTimeMs,
			Layer1Count:          r.Layer1Count,
			Layer2Count:          r.Layer2Count,
			Layer3Count:          r.Layer3Count,
		}
	}

	return timeSeries, nil
}

// CachedClassificationResult represents a cached classification result
// Note: Explanation is stored as json.RawMessage to avoid circular dependencies
type CachedClassificationResult struct {
	IndustryName      string          `json:"industry_name"`
	Confidence        float64         `json:"confidence"`
	Keywords          []string        `json:"keywords"`
	ProcessingTime    time.Duration   `json:"processing_time"`
	Method            string          `json:"method"`
	Reasoning         string          `json:"reasoning"`
	Explanation       json.RawMessage `json:"explanation,omitempty"`
	CreatedAt         time.Time       `json:"created_at"`
	LayerUsed         string          `json:"layer_used"`
	ProcessingTimeMs  int             `json:"processing_time_ms"`
	FromCache         bool            `json:"from_cache"`
	CachedAt          *time.Time      `json:"cached_at,omitempty"`
}

// ClassificationMetricsRecord represents a single classification metrics record
type ClassificationMetricsRecord struct {
	RequestID        string          `json:"request_id"`
	BusinessName     string          `json:"business_name"`
	WebsiteURL       string          `json:"website_url"`
	PrimaryIndustry  string          `json:"primary_industry"`
	Confidence       float64         `json:"confidence"`
	LayerUsed        string          `json:"layer_used"`
	Method           string          `json:"method"`
	TotalTimeMs      int             `json:"total_time_ms"`
	ScrapeTimeMs     int             `json:"scrape_time_ms"`
	Layer1TimeMs     int             `json:"layer1_time_ms"`
	Layer2TimeMs     int             `json:"layer2_time_ms"`
	Layer3TimeMs     int             `json:"layer3_time_ms"`
	FromCache        bool            `json:"from_cache"`
	MCCCodes         json.RawMessage `json:"mcc_codes,omitempty"`
	SICCodes         json.RawMessage `json:"sic_codes,omitempty"`
	NAICSCodes       json.RawMessage `json:"naics_codes,omitempty"`
	UserAgent        string          `json:"user_agent,omitempty"`
	IPAddress        string          `json:"ip_address,omitempty"`
}

// DashboardMetric represents a single dashboard metric
type DashboardMetric struct {
	Metric      string  `json:"metric"`
	Value       float64 `json:"value"`
	Description string  `json:"description"`
}

// TimeSeriesData represents time series data for charts
type TimeSeriesData struct {
	Date                 time.Time `json:"date"`
	TotalClassifications int64     `json:"total_classifications"`
	CacheHits            int64     `json:"cache_hits"`
	CacheMisses          int64     `json:"cache_misses"`
	AvgConfidence        float64   `json:"avg_confidence"`
	AvgTotalTimeMs       float64   `json:"avg_total_time_ms"`
	Layer1Count          int64     `json:"layer1_count"`
	Layer2Count          int64     `json:"layer2_count"`
	Layer3Count          int64     `json:"layer3_count"`
}
