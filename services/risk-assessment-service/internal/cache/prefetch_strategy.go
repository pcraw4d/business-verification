package cache

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PrefetchStrategy implements intelligent cache warming and prefetching
type PrefetchStrategy struct {
	coordinator *CacheCoordinator
	logger      *zap.Logger
	stats       *PrefetchStats
	mu          sync.RWMutex
	enabled     bool
}

// PrefetchStats represents statistics for prefetch operations
type PrefetchStats struct {
	TotalPrefetches      int64         `json:"total_prefetches"`
	SuccessfulPrefetches int64         `json:"successful_prefetches"`
	FailedPrefetches     int64         `json:"failed_prefetches"`
	PrefetchTime         time.Duration `json:"prefetch_time"`
	CacheHits            int64         `json:"cache_hits"`
	CacheMisses          int64         `json:"cache_misses"`
	LastPrefetch         time.Time     `json:"last_prefetch"`
}

// PrefetchItem represents an item to be prefetched
type PrefetchItem struct {
	Key       string        `json:"key"`
	Value     interface{}   `json:"value"`
	TTL       time.Duration `json:"ttl"`
	Priority  int           `json:"priority"`
	Frequency int64         `json:"frequency"`
	LastUsed  time.Time     `json:"last_used"`
}

// PrefetchConfig represents configuration for prefetch strategy
type PrefetchConfig struct {
	Enabled           bool          `json:"enabled"`
	MaxPrefetchItems  int           `json:"max_prefetch_items"`
	PrefetchInterval  time.Duration `json:"prefetch_interval"`
	PopularThreshold  int64         `json:"popular_threshold"`
	PriorityThreshold int           `json:"priority_threshold"`
	TTLMultiplier     float64       `json:"ttl_multiplier"`
}

// NewPrefetchStrategy creates a new prefetch strategy
func NewPrefetchStrategy(coordinator *CacheCoordinator, config *PrefetchConfig, logger *zap.Logger) *PrefetchStrategy {
	if config == nil {
		config = &PrefetchConfig{
			Enabled:           true,
			MaxPrefetchItems:  1000,
			PrefetchInterval:  5 * time.Minute,
			PopularThreshold:  10,
			PriorityThreshold: 5,
			TTLMultiplier:     1.5,
		}
	}

	strategy := &PrefetchStrategy{
		coordinator: coordinator,
		logger:      logger,
		stats:       &PrefetchStats{},
		enabled:     config.Enabled,
	}

	// Start prefetch routine if enabled
	if config.Enabled && config.PrefetchInterval > 0 {
		go strategy.startPrefetchRoutine(config)
	}

	logger.Info("Prefetch strategy initialized",
		zap.Bool("enabled", config.Enabled),
		zap.Duration("interval", config.PrefetchInterval),
		zap.Int("max_items", config.MaxPrefetchItems))

	return strategy
}

// PrefetchPopularItems prefetches popular items based on access patterns
func (ps *PrefetchStrategy) PrefetchPopularItems(ctx context.Context, items []*PrefetchItem) error {
	if !ps.enabled {
		return fmt.Errorf("prefetch strategy is disabled")
	}

	start := time.Now()
	defer func() {
		ps.mu.Lock()
		ps.stats.PrefetchTime += time.Since(start)
		ps.stats.LastPrefetch = time.Now()
		ps.mu.Unlock()
	}()

	// Sort items by priority and frequency
	sortedItems := ps.sortItemsByPriority(items)

	successful := 0
	failed := 0

	for _, item := range sortedItems {
		if err := ps.prefetchItem(ctx, item); err != nil {
			ps.logger.Warn("Failed to prefetch item",
				zap.String("key", item.Key),
				zap.Error(err))
			failed++
		} else {
			successful++
		}
	}

	ps.mu.Lock()
	ps.stats.TotalPrefetches += int64(len(items))
	ps.stats.SuccessfulPrefetches += int64(successful)
	ps.stats.FailedPrefetches += int64(failed)
	ps.mu.Unlock()

	ps.logger.Info("Prefetch operation completed",
		zap.Int("total", len(items)),
		zap.Int("successful", successful),
		zap.Int("failed", failed),
		zap.Duration("duration", time.Since(start)))

	return nil
}

// PrefetchByPattern prefetches items matching a specific pattern
func (ps *PrefetchStrategy) PrefetchByPattern(ctx context.Context, pattern string, generator func(string) (interface{}, error)) error {
	if !ps.enabled {
		return fmt.Errorf("prefetch strategy is disabled")
	}

	// This is a simplified implementation
	// In a real implementation, you would generate keys based on the pattern
	// and use the generator function to create values

	ps.logger.Info("Prefetch by pattern",
		zap.String("pattern", pattern))

	return nil
}

// PrefetchMLModelResults prefetches ML model results for common business profiles
func (ps *PrefetchStrategy) PrefetchMLModelResults(ctx context.Context) error {
	if !ps.enabled {
		return fmt.Errorf("prefetch strategy is disabled")
	}

	// Common business profiles to prefetch
	commonProfiles := []map[string]interface{}{
		{
			"industry": "Technology",
			"country":  "US",
			"size":     "Small",
		},
		{
			"industry": "Finance",
			"country":  "US",
			"size":     "Medium",
		},
		{
			"industry": "Healthcare",
			"country":  "US",
			"size":     "Large",
		},
		{
			"industry": "Retail",
			"country":  "US",
			"size":     "Small",
		},
		{
			"industry": "Manufacturing",
			"country":  "US",
			"size":     "Large",
		},
	}

	var items []*PrefetchItem
	for i, profile := range commonProfiles {
		key := fmt.Sprintf("ml_model:%s:%s:%s", profile["industry"], profile["country"], profile["size"])

		// Generate mock ML model result
		value := map[string]interface{}{
			"risk_score":    0.3 + float64(i)*0.1,
			"confidence":    0.85 + float64(i)*0.02,
			"factors":       []string{"industry_risk", "country_risk", "size_risk"},
			"model_version": "v1.0",
			"generated_at":  time.Now(),
		}

		items = append(items, &PrefetchItem{
			Key:       key,
			Value:     value,
			TTL:       1 * time.Hour,
			Priority:  8 - i, // Higher priority for more common profiles
			Frequency: 100 - int64(i*10),
			LastUsed:  time.Now(),
		})
	}

	return ps.PrefetchPopularItems(ctx, items)
}

// PrefetchIndustryData prefetches industry-specific risk data
func (ps *PrefetchStrategy) PrefetchIndustryData(ctx context.Context) error {
	if !ps.enabled {
		return fmt.Errorf("prefetch strategy is disabled")
	}

	// Popular industries to prefetch
	industries := []string{
		"Technology", "Finance", "Healthcare", "Retail", "Manufacturing",
		"Real Estate", "Education", "Transportation", "Energy", "Media",
	}

	var items []*PrefetchItem
	for i, industry := range industries {
		key := fmt.Sprintf("industry_data:%s", industry)

		// Generate mock industry data
		value := map[string]interface{}{
			"industry":     industry,
			"risk_level":   "Medium",
			"growth_rate":  0.05 + float64(i)*0.01,
			"volatility":   0.3 + float64(i)*0.02,
			"regulations":  []string{"compliance_1", "compliance_2"},
			"last_updated": time.Now(),
		}

		items = append(items, &PrefetchItem{
			Key:       key,
			Value:     value,
			TTL:       6 * time.Hour,
			Priority:  7 - i/2, // Higher priority for top industries
			Frequency: 50 - int64(i*3),
			LastUsed:  time.Now(),
		})
	}

	return ps.PrefetchPopularItems(ctx, items)
}

// PrefetchCountryData prefetches country-specific risk data
func (ps *PrefetchStrategy) PrefetchCountryData(ctx context.Context) error {
	if !ps.enabled {
		return fmt.Errorf("prefetch strategy is disabled")
	}

	// Popular countries to prefetch
	countries := []string{
		"US", "CA", "GB", "DE", "FR", "JP", "AU", "NL", "SE", "CH",
	}

	var items []*PrefetchItem
	for i, country := range countries {
		key := fmt.Sprintf("country_data:%s", country)

		// Generate mock country data
		value := map[string]interface{}{
			"country":      country,
			"risk_level":   "Low",
			"stability":    0.8 + float64(i)*0.01,
			"gdp_growth":   0.02 + float64(i)*0.005,
			"regulations":  []string{"kyc_requirements", "aml_requirements"},
			"last_updated": time.Now(),
		}

		items = append(items, &PrefetchItem{
			Key:       key,
			Value:     value,
			TTL:       12 * time.Hour,
			Priority:  6 - i/3, // Higher priority for major countries
			Frequency: 30 - int64(i*2),
			LastUsed:  time.Now(),
		})
	}

	return ps.PrefetchPopularItems(ctx, items)
}

// WarmupCache performs initial cache warming on startup
func (ps *PrefetchStrategy) WarmupCache(ctx context.Context) error {
	if !ps.enabled {
		return fmt.Errorf("prefetch strategy is disabled")
	}

	ps.logger.Info("Starting cache warmup")

	// Prefetch different types of data
	if err := ps.PrefetchMLModelResults(ctx); err != nil {
		ps.logger.Warn("Failed to prefetch ML model results", zap.Error(err))
	}

	if err := ps.PrefetchIndustryData(ctx); err != nil {
		ps.logger.Warn("Failed to prefetch industry data", zap.Error(err))
	}

	if err := ps.PrefetchCountryData(ctx); err != nil {
		ps.logger.Warn("Failed to prefetch country data", zap.Error(err))
	}

	ps.logger.Info("Cache warmup completed")
	return nil
}

// GetStats returns prefetch statistics
func (ps *PrefetchStrategy) GetStats() *PrefetchStats {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	stats := *ps.stats
	return &stats
}

// Enable enables the prefetch strategy
func (ps *PrefetchStrategy) Enable() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.enabled = true
}

// Disable disables the prefetch strategy
func (ps *PrefetchStrategy) Disable() {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.enabled = false
}

// Helper methods

func (ps *PrefetchStrategy) prefetchItem(ctx context.Context, item *PrefetchItem) error {
	// Check if item already exists in cache
	exists, err := ps.coordinator.Exists(ctx, item.Key)
	if err != nil {
		return fmt.Errorf("failed to check if item exists: %w", err)
	}

	if exists {
		ps.mu.Lock()
		ps.stats.CacheHits++
		ps.mu.Unlock()
		return nil // Item already cached
	}

	// Set item in cache
	if err := ps.coordinator.SetWithTTL(ctx, item.Key, item.Value, item.TTL); err != nil {
		ps.mu.Lock()
		ps.stats.CacheMisses++
		ps.mu.Unlock()
		return fmt.Errorf("failed to set item in cache: %w", err)
	}

	ps.mu.Lock()
	ps.stats.CacheMisses++
	ps.mu.Unlock()

	return nil
}

func (ps *PrefetchStrategy) sortItemsByPriority(items []*PrefetchItem) []*PrefetchItem {
	// Create a copy to avoid modifying the original slice
	sorted := make([]*PrefetchItem, len(items))
	copy(sorted, items)

	// Sort by priority (descending) and then by frequency (descending)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Priority != sorted[j].Priority {
			return sorted[i].Priority > sorted[j].Priority
		}
		return sorted[i].Frequency > sorted[j].Frequency
	})

	return sorted
}

func (ps *PrefetchStrategy) startPrefetchRoutine(config *PrefetchConfig) {
	ticker := time.NewTicker(config.PrefetchInterval)
	defer ticker.Stop()

	for range ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		// Perform periodic prefetch operations
		if err := ps.PrefetchMLModelResults(ctx); err != nil {
			ps.logger.Warn("Periodic ML model prefetch failed", zap.Error(err))
		}

		if err := ps.PrefetchIndustryData(ctx); err != nil {
			ps.logger.Warn("Periodic industry data prefetch failed", zap.Error(err))
		}

		if err := ps.PrefetchCountryData(ctx); err != nil {
			ps.logger.Warn("Periodic country data prefetch failed", zap.Error(err))
		}

		cancel()
	}
}
