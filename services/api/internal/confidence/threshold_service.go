package confidence

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// IndustryThresholdService provides dynamic industry-specific confidence thresholds
type IndustryThresholdService struct {
	repository IndustryThresholdRepository
	cache      map[string]float64
	cacheMutex sync.RWMutex
	logger     Logger
	lastUpdate time.Time
	cacheTTL   time.Duration
}

// IndustryThresholdRepository defines the interface for accessing industry threshold data
type IndustryThresholdRepository interface {
	GetIndustryThreshold(ctx context.Context, industryName string) (float64, error)
	GetAllIndustryThresholds(ctx context.Context) (map[string]float64, error)
	GetIndustryByID(ctx context.Context, industryID int) (*Industry, error)
	GetIndustryByName(ctx context.Context, industryName string) (*Industry, error)
}

// Industry represents an industry with its threshold information
type Industry struct {
	ID                  int     `json:"id"`
	Name                string  `json:"name"`
	Description         string  `json:"description"`
	Category            string  `json:"category"`
	ConfidenceThreshold float64 `json:"confidence_threshold"`
	IsActive            bool    `json:"is_active"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}

// ThresholdCacheConfig holds configuration for threshold caching
type ThresholdCacheConfig struct {
	Enabled        bool
	TTL            time.Duration
	MaxSize        int
	WarmingEnabled bool
}

// NewIndustryThresholdService creates a new industry threshold service
func NewIndustryThresholdService(repository IndustryThresholdRepository, logger Logger) *IndustryThresholdService {
	if logger == nil {
		logger = &defaultLogger{}
	}

	return &IndustryThresholdService{
		repository: repository,
		cache:      make(map[string]float64),
		logger:     logger,
		cacheTTL:   5 * time.Minute, // Default 5-minute cache TTL
	}
}

// GetIndustryThreshold retrieves the confidence threshold for a specific industry
func (its *IndustryThresholdService) GetIndustryThreshold(ctx context.Context, industryName string) (float64, error) {
	its.logger.Printf("🎯 Getting industry threshold for: %s", industryName)

	// Check cache first
	its.cacheMutex.RLock()
	if threshold, exists := its.cache[industryName]; exists && !its.isCacheExpired() {
		its.cacheMutex.RUnlock()
		its.logger.Printf("✅ Retrieved threshold from cache: %s = %.3f", industryName, threshold)
		return threshold, nil
	}
	its.cacheMutex.RUnlock()

	// Cache miss or expired - fetch from database
	threshold, err := its.fetchThresholdFromDatabase(ctx, industryName)
	if err != nil {
		its.logger.Printf("❌ Failed to fetch threshold for %s: %v", industryName, err)
		return its.getDefaultThreshold(), nil // Return default on error
	}

	// Update cache
	its.cacheMutex.Lock()
	its.cache[industryName] = threshold
	its.lastUpdate = time.Now()
	its.cacheMutex.Unlock()

	its.logger.Printf("✅ Retrieved threshold from database: %s = %.3f", industryName, threshold)
	return threshold, nil
}

// GetAllIndustryThresholds retrieves all industry thresholds
func (its *IndustryThresholdService) GetAllIndustryThresholds(ctx context.Context) (map[string]float64, error) {
	its.logger.Printf("🎯 Getting all industry thresholds")

	// Check if cache is still valid
	its.cacheMutex.RLock()
	if len(its.cache) > 0 && !its.isCacheExpired() {
		// Return a copy of the cache
		thresholds := make(map[string]float64)
		for name, threshold := range its.cache {
			thresholds[name] = threshold
		}
		its.cacheMutex.RUnlock()
		its.logger.Printf("✅ Retrieved %d thresholds from cache", len(thresholds))
		return thresholds, nil
	}
	its.cacheMutex.RUnlock()

	// Cache miss or expired - fetch from database
	thresholds, err := its.repository.GetAllIndustryThresholds(ctx)
	if err != nil {
		its.logger.Printf("❌ Failed to fetch all thresholds: %v", err)
		return its.getDefaultThresholds(), nil // Return defaults on error
	}

	// Update cache
	its.cacheMutex.Lock()
	its.cache = make(map[string]float64)
	for name, threshold := range thresholds {
		its.cache[name] = threshold
	}
	its.lastUpdate = time.Now()
	its.cacheMutex.Unlock()

	its.logger.Printf("✅ Retrieved %d thresholds from database", len(thresholds))
	return thresholds, nil
}

// RefreshCache forces a refresh of the threshold cache
func (its *IndustryThresholdService) RefreshCache(ctx context.Context) error {
	its.logger.Printf("🔄 Refreshing threshold cache")

	thresholds, err := its.repository.GetAllIndustryThresholds(ctx)
	if err != nil {
		its.logger.Printf("❌ Failed to refresh cache: %v", err)
		return fmt.Errorf("failed to refresh threshold cache: %w", err)
	}

	its.cacheMutex.Lock()
	its.cache = make(map[string]float64)
	for name, threshold := range thresholds {
		its.cache[name] = threshold
	}
	its.lastUpdate = time.Now()
	its.cacheMutex.Unlock()

	its.logger.Printf("✅ Cache refreshed with %d thresholds", len(thresholds))
	return nil
}

// SetCacheTTL sets the cache time-to-live
func (its *IndustryThresholdService) SetCacheTTL(ttl time.Duration) {
	its.cacheMutex.Lock()
	its.cacheTTL = ttl
	its.cacheMutex.Unlock()
	its.logger.Printf("⏰ Cache TTL set to: %v", ttl)
}

// GetCacheStats returns cache statistics
func (its *IndustryThresholdService) GetCacheStats() map[string]interface{} {
	its.cacheMutex.RLock()
	defer its.cacheMutex.RUnlock()

	return map[string]interface{}{
		"cache_size":        len(its.cache),
		"last_update":       its.lastUpdate,
		"cache_ttl":         its.cacheTTL,
		"is_expired":        its.isCacheExpired(),
		"cached_industries": its.getCachedIndustryNames(),
	}
}

// fetchThresholdFromDatabase fetches a single threshold from the database
func (its *IndustryThresholdService) fetchThresholdFromDatabase(ctx context.Context, industryName string) (float64, error) {
	industry, err := its.repository.GetIndustryByName(ctx, industryName)
	if err != nil {
		return 0, fmt.Errorf("failed to get industry %s: %w", industryName, err)
	}

	if !industry.IsActive {
		its.logger.Printf("⚠️ Industry %s is inactive, using default threshold", industryName)
		return its.getDefaultThreshold(), nil
	}

	return industry.ConfidenceThreshold, nil
}

// isCacheExpired checks if the cache has expired
func (its *IndustryThresholdService) isCacheExpired() bool {
	return time.Since(its.lastUpdate) > its.cacheTTL
}

// getDefaultThreshold returns the default confidence threshold
func (its *IndustryThresholdService) getDefaultThreshold() float64 {
	return 0.50 // Default threshold for unknown industries
}

// getDefaultThresholds returns default thresholds for common industries
func (its *IndustryThresholdService) getDefaultThresholds() map[string]float64 {
	return map[string]float64{
		"Restaurants":      0.75,
		"Fast Food":        0.80,
		"Food & Beverage":  0.70,
		"Legal Services":   0.75,
		"Healthcare":       0.80,
		"Technology":       0.75,
		"Retail":           0.70,
		"Manufacturing":    0.70,
		"Construction":     0.70,
		"Transportation":   0.70,
		"Education":        0.75,
		"Entertainment":    0.65,
		"Agriculture":      0.70,
		"Energy":           0.70,
		"General Business": 0.50,
	}
}

// getCachedIndustryNames returns a list of cached industry names
func (its *IndustryThresholdService) getCachedIndustryNames() []string {
	names := make([]string, 0, len(its.cache))
	for name := range its.cache {
		names = append(names, name)
	}
	return names
}

// ClearCache clears the threshold cache
func (its *IndustryThresholdService) ClearCache() {
	its.cacheMutex.Lock()
	its.cache = make(map[string]float64)
	its.lastUpdate = time.Time{}
	its.cacheMutex.Unlock()
	its.logger.Printf("🗑️ Threshold cache cleared")
}

// ValidateThreshold validates that a threshold is within acceptable bounds
func (its *IndustryThresholdService) ValidateThreshold(threshold float64) error {
	if threshold < 0.1 || threshold > 1.0 {
		return fmt.Errorf("threshold %.3f is out of valid range [0.1, 1.0]", threshold)
	}
	return nil
}

// GetThresholdRecommendation provides recommendations for threshold values based on industry type
func (its *IndustryThresholdService) GetThresholdRecommendation(industryName string) float64 {
	// Industry-specific recommendations based on classification complexity
	recommendations := map[string]float64{
		// High-confidence industries (clear indicators)
		"Fast Food":      0.80,
		"Healthcare":     0.80,
		"Legal Services": 0.75,
		"Technology":     0.75,
		"Restaurants":    0.75,

		// Medium-confidence industries (moderate indicators)
		"Retail":         0.70,
		"Manufacturing":  0.70,
		"Construction":   0.70,
		"Transportation": 0.70,
		"Education":      0.75,
		"Agriculture":    0.70,
		"Energy":         0.70,

		// Lower-confidence industries (ambiguous indicators)
		"Entertainment":   0.65,
		"Food & Beverage": 0.70,

		// Default
		"General Business": 0.50,
	}

	if threshold, exists := recommendations[industryName]; exists {
		return threshold
	}

	return 0.60 // Default recommendation for unknown industries
}
