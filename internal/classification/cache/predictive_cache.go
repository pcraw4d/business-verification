package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// ClassificationResultCache stores cached classification results
type ClassificationResultCache struct {
	mu       sync.RWMutex
	results  map[string]*CachedClassificationResult
	ttl      time.Duration
	logger   *log.Logger
}

// CachedClassificationResult represents a cached classification result
type CachedClassificationResult struct {
	PrimaryIndustry string
	Confidence      float64
	Keywords        []string
	Reasoning       string
	CachedAt        time.Time
	ExpiresAt       time.Time
}

// NewClassificationResultCache creates a new classification result cache
func NewClassificationResultCache(ttl time.Duration, logger *log.Logger) *ClassificationResultCache {
	if logger == nil {
		logger = log.Default()
	}
	
	cache := &ClassificationResultCache{
		results: make(map[string]*CachedClassificationResult),
		ttl:     ttl,
		logger:  logger,
	}
	
	// Start cleanup goroutine
	go cache.cleanup()
	
	return cache
}

// Get retrieves a cached classification result
func (c *ClassificationResultCache) Get(key string) (*CachedClassificationResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	result, exists := c.results[key]
	if !exists {
		return nil, false
	}
	
	// Check if expired
	if time.Now().After(result.ExpiresAt) {
		return nil, false
	}
	
	return result, true
}

// Set stores a classification result in the cache
func (c *ClassificationResultCache) Set(key string, result *CachedClassificationResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	now := time.Now()
	result.CachedAt = now
	result.ExpiresAt = now.Add(c.ttl)
	
	c.results[key] = result
}

// cleanup periodically removes expired entries
func (c *ClassificationResultCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		count := 0
		for key, result := range c.results {
			if now.After(result.ExpiresAt) {
				delete(c.results, key)
				count++
			}
		}
		c.mu.Unlock()
		
		if count > 0 {
			c.logger.Printf("🧹 Cleaned up %d expired cache entries", count)
		}
	}
}

// GetStats returns cache statistics
func (c *ClassificationResultCache) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	now := time.Now()
	active := 0
	expired := 0
	
	for _, result := range c.results {
		if now.After(result.ExpiresAt) {
			expired++
		} else {
			active++
		}
	}
	
	return map[string]interface{}{
		"total_entries": len(c.results),
		"active_entries": active,
		"expired_entries": expired,
		"ttl_seconds": c.ttl.Seconds(),
	}
}

// PredictiveCache provides predictive caching for classification results
// Pre-loads likely requests based on business name patterns
type PredictiveCache struct {
	cache          *ClassificationResultCache
	classifier     ClassificationPredictor
	patterns       map[string][]string // business_name_pattern -> likely_keywords
	mu             sync.RWMutex
	logger         *log.Logger
	preloadEnabled bool
}

// ClassificationPredictor interface for classification prediction
type ClassificationPredictor interface {
	Classify(ctx context.Context, businessName, description, websiteURL string) (*ClassificationPrediction, error)
}

// ClassificationPrediction represents a classification prediction
type ClassificationPrediction struct {
	PrimaryIndustry string
	Confidence      float64
	Keywords        []string
	Reasoning       string
}

// NewPredictiveCache creates a new predictive cache
func NewPredictiveCache(
	cache *ClassificationResultCache,
	classifier ClassificationPredictor,
	logger *log.Logger,
) *PredictiveCache {
	if logger == nil {
		logger = log.Default()
	}
	
	return &PredictiveCache{
		cache:          cache,
		classifier:     classifier,
		patterns:       make(map[string][]string),
		logger:         logger,
		preloadEnabled: true,
	}
}

// Get retrieves a cached result or returns false
func (pc *PredictiveCache) Get(businessName, description, websiteURL string) (*CachedClassificationResult, bool) {
	key := pc.generateCacheKey(businessName, description, websiteURL)
	return pc.cache.Get(key)
}

// Set stores a classification result in the cache
func (pc *PredictiveCache) Set(businessName, description, websiteURL string, result *CachedClassificationResult) {
	key := pc.generateCacheKey(businessName, description, websiteURL)
	pc.cache.Set(key, result)
}

// PreloadCache pre-caches likely requests based on business name variations
// Runs in background to avoid blocking
func (pc *PredictiveCache) PreloadCache(ctx context.Context, businessName, description, websiteURL string) {
	if !pc.preloadEnabled {
		return
	}
	
	// Generate name variations
	variations := pc.generateNameVariations(businessName)
	
	// Pre-cache in background
	go func() {
		for _, variation := range variations {
			// Skip if already cached
			key := pc.generateCacheKey(variation, description, websiteURL)
			if _, exists := pc.cache.Get(key); exists {
				continue
			}
			
			// Pre-classify and cache
			result, err := pc.classifyAndCache(ctx, variation, description, websiteURL)
			if err == nil && result != nil {
				pc.logger.Printf("✅ Pre-cached: %s", variation)
			}
		}
	}()
}

// generateNameVariations generates likely name variations for predictive caching
func (pc *PredictiveCache) generateNameVariations(name string) []string {
	variations := []string{name}
	seen := make(map[string]bool)
	seen[name] = true
	
	// Remove common suffixes
	suffixes := []string{" Inc", " LLC", " Corp", " Ltd", " Co", " Inc.", " LLC.", " Corp.", " Ltd.", " Co."}
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			variation := strings.TrimSuffix(name, suffix)
			if !seen[variation] {
				variations = append(variations, variation)
				seen[variation] = true
			}
		}
	}
	
	// Add common prefixes
	prefixes := []string{"The ", "A "}
	for _, prefix := range prefixes {
		if !strings.HasPrefix(name, prefix) {
			variation := prefix + name
			if !seen[variation] {
				variations = append(variations, variation)
				seen[variation] = true
			}
		}
	}
	
	// Remove common prefixes
	if strings.HasPrefix(name, "The ") {
		variation := strings.TrimPrefix(name, "The ")
		if !seen[variation] {
			variations = append(variations, variation)
			seen[variation] = true
		}
	}
	if strings.HasPrefix(name, "A ") {
		variation := strings.TrimPrefix(name, "A ")
		if !seen[variation] {
			variations = append(variations, variation)
			seen[variation] = true
		}
	}
	
	// Generate lowercase variation
	lowerVariation := strings.ToLower(name)
	if !seen[lowerVariation] {
		variations = append(variations, lowerVariation)
		seen[lowerVariation] = true
	}
	
	return variations
}

// classifyAndCache classifies a business and caches the result
func (pc *PredictiveCache) classifyAndCache(
	ctx context.Context,
	businessName, description, websiteURL string,
) (*CachedClassificationResult, error) {
	// Create timeout context for pre-caching (shorter timeout)
	preloadCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	// Classify
	prediction, err := pc.classifier.Classify(preloadCtx, businessName, description, websiteURL)
	if err != nil {
		return nil, fmt.Errorf("classification failed: %w", err)
	}
	
	// Create cached result
	result := &CachedClassificationResult{
		PrimaryIndustry: prediction.PrimaryIndustry,
		Confidence:      prediction.Confidence,
		Keywords:        prediction.Keywords,
		Reasoning:       prediction.Reasoning,
	}
	
	// Cache the result
	pc.Set(businessName, description, websiteURL, result)
	
	return result, nil
}

// normalizeBusinessName normalizes a business name for cache key generation
func normalizeBusinessName(name string) string {
	// Trim whitespace
	name = strings.TrimSpace(name)
	
	// Remove common prefixes
	name = strings.TrimPrefix(name, "The ")
	name = strings.TrimPrefix(name, "A ")
	
	// Remove common suffixes (case-insensitive)
	suffixes := []string{" Inc", " LLC", " Corp", " Ltd", " Co", " Inc.", " LLC.", " Corp.", " Ltd.", " Co.",
		" inc", " llc", " corp", " ltd", " co", " inc.", " llc.", " corp.", " ltd.", " co."}
	for _, suffix := range suffixes {
		if strings.HasSuffix(name, suffix) {
			name = strings.TrimSuffix(name, suffix)
		}
	}
	
	// Lowercase and trim again
	return strings.ToLower(strings.TrimSpace(name))
}

// generateCacheKey generates a cache key from business information
// Fix: Normalizes business name to improve cache hit rate
func (pc *PredictiveCache) generateCacheKey(businessName, description, websiteURL string) string {
	// Normalize business name for better cache matching
	normalizedName := normalizeBusinessName(businessName)
	
	data := fmt.Sprintf("%s|%s|%s", 
		normalizedName,
		strings.ToLower(strings.TrimSpace(description)),
		strings.ToLower(strings.TrimSpace(websiteURL)))
	
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// SetPreloadEnabled enables or disables predictive preloading
func (pc *PredictiveCache) SetPreloadEnabled(enabled bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.preloadEnabled = enabled
}

// AddPattern adds a business name pattern with likely keywords
func (pc *PredictiveCache) AddPattern(pattern string, likelyKeywords []string) {
	pc.mu.Lock()
	defer pc.mu.Unlock()
	pc.patterns[pattern] = likelyKeywords
}

// GetPattern returns likely keywords for a pattern
func (pc *PredictiveCache) GetPattern(pattern string) ([]string, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()
	keywords, exists := pc.patterns[pattern]
	return keywords, exists
}

