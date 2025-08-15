package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/observability"
	supa "github.com/supabase-community/supabase-go"
)

// SupabaseCache represents a Supabase-based cache implementation
type SupabaseCache struct {
	client *supa.Client
	logger *observability.Logger
}

// NewSupabaseCache creates a new Supabase cache client
func NewSupabaseCache(cfg *config.SupabaseConfig, logger *observability.Logger) (*SupabaseCache, error) {
	client, err := supa.NewClient(cfg.URL, cfg.APIKey, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Supabase client: %w", err)
	}

	return &SupabaseCache{
		client: client,
		logger: logger,
	}, nil
}

// CacheEntry represents a cache entry in Supabase
type CacheEntry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Set stores a value in the cache
func (s *SupabaseCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	s.logger.Debug("Setting cache value in Supabase", "key", key, "ttl", ttl)

	// Serialize the value
	valueBytes, err := json.Marshal(value)
	if err != nil {
		s.logger.Error("Failed to marshal cache value", "error", err)
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	entry := CacheEntry{
		Key:       key,
		Value:     string(valueBytes),
		ExpiresAt: time.Now().Add(ttl),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Upsert the cache entry
	_, err = s.client.DB.From("cache_entries").
		Upsert(entry, false, "", "", "").
		Execute("")

	if err != nil {
		s.logger.Error("Failed to set cache value in Supabase", "error", err)
		return fmt.Errorf("failed to set cache value: %w", err)
	}

	s.logger.Info("Successfully set cache value in Supabase", "key", key)
	return nil
}

// Get retrieves a value from the cache
func (s *SupabaseCache) Get(ctx context.Context, key string) (interface{}, error) {
	s.logger.Debug("Getting cache value from Supabase", "key", key)

	result, err := s.client.DB.From("cache_entries").
		Select("*").
		Eq("key", key).
		Gt("expires_at", time.Now()).
		Single().
		Execute("")

	if err != nil {
		s.logger.Debug("Cache miss in Supabase", "key", key)
		return nil, fmt.Errorf("cache miss: %w", err)
	}

	var entry CacheEntry
	if err := result.Unmarshal(&entry); err != nil {
		s.logger.Error("Failed to unmarshal cache entry", "error", err)
		return nil, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		s.logger.Debug("Cache entry expired in Supabase", "key", key)
		// Delete expired entry
		s.Delete(ctx, key)
		return nil, fmt.Errorf("cache entry expired")
	}

	// Deserialize the value
	var value interface{}
	if err := json.Unmarshal([]byte(entry.Value), &value); err != nil {
		s.logger.Error("Failed to unmarshal cache value", "error", err)
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	s.logger.Debug("Cache hit in Supabase", "key", key)
	return value, nil
}

// Delete removes a value from the cache
func (s *SupabaseCache) Delete(ctx context.Context, key string) error {
	s.logger.Debug("Deleting cache value from Supabase", "key", key)

	_, err := s.client.DB.From("cache_entries").
		Delete("", "").
		Eq("key", key).
		Execute("")

	if err != nil {
		s.logger.Error("Failed to delete cache value from Supabase", "error", err)
		return fmt.Errorf("failed to delete cache value: %w", err)
	}

	s.logger.Info("Successfully deleted cache value from Supabase", "key", key)
	return nil
}

// Clear removes all values from the cache
func (s *SupabaseCache) Clear(ctx context.Context) error {
	s.logger.Debug("Clearing all cache values from Supabase")

	_, err := s.client.DB.From("cache_entries").
		Delete("", "").
		Execute("")

	if err != nil {
		s.logger.Error("Failed to clear cache values from Supabase", "error", err)
		return fmt.Errorf("failed to clear cache: %w", err)
	}

	s.logger.Info("Successfully cleared all cache values from Supabase")
	return nil
}

// GetWithTTL retrieves a value and its remaining TTL from the cache
func (s *SupabaseCache) GetWithTTL(ctx context.Context, key string) (interface{}, time.Duration, error) {
	s.logger.Debug("Getting cache value with TTL from Supabase", "key", key)

	result, err := s.client.DB.From("cache_entries").
		Select("*").
		Eq("key", key).
		Single().
		Execute("")

	if err != nil {
		s.logger.Debug("Cache miss in Supabase", "key", key)
		return nil, 0, fmt.Errorf("cache miss: %w", err)
	}

	var entry CacheEntry
	if err := result.Unmarshal(&entry); err != nil {
		s.logger.Error("Failed to unmarshal cache entry", "error", err)
		return nil, 0, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		s.logger.Debug("Cache entry expired in Supabase", "key", key)
		// Delete expired entry
		s.Delete(ctx, key)
		return nil, 0, fmt.Errorf("cache entry expired")
	}

	// Calculate remaining TTL
	remainingTTL := time.Until(entry.ExpiresAt)

	// Deserialize the value
	var value interface{}
	if err := json.Unmarshal([]byte(entry.Value), &value); err != nil {
		s.logger.Error("Failed to unmarshal cache value", "error", err)
		return nil, 0, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	s.logger.Debug("Cache hit in Supabase", "key", key, "remaining_ttl", remainingTTL)
	return value, remainingTTL, nil
}

// SetNX sets a value only if it doesn't already exist
func (s *SupabaseCache) SetNX(ctx context.Context, key string, value interface{}, ttl time.Duration) (bool, error) {
	s.logger.Debug("Setting cache value NX in Supabase", "key", key)

	// Check if key already exists
	existing, err := s.Get(ctx, key)
	if err == nil && existing != nil {
		s.logger.Debug("Cache key already exists in Supabase", "key", key)
		return false, nil
	}

	// Set the value
	err = s.Set(ctx, key, value, ttl)
	if err != nil {
		return false, err
	}

	s.logger.Info("Successfully set cache value NX in Supabase", "key", key)
	return true, nil
}

// Increment increments a numeric value in the cache
func (s *SupabaseCache) Increment(ctx context.Context, key string, value int64) (int64, error) {
	s.logger.Debug("Incrementing cache value in Supabase", "key", key, "value", value)

	// Get current value
	current, err := s.Get(ctx, key)
	if err != nil {
		// Key doesn't exist, start with 0
		current = int64(0)
	}

	// Convert to int64
	var currentInt int64
	switch v := current.(type) {
	case int64:
		currentInt = v
	case int:
		currentInt = int64(v)
	case float64:
		currentInt = int64(v)
	default:
		return 0, fmt.Errorf("cache value is not numeric")
	}

	// Calculate new value
	newValue := currentInt + value

	// Set the new value
	err = s.Set(ctx, key, newValue, 24*time.Hour) // Default TTL of 24 hours
	if err != nil {
		return 0, err
	}

	s.logger.Info("Successfully incremented cache value in Supabase", "key", key, "new_value", newValue)
	return newValue, nil
}

// GetKeys retrieves all cache keys matching a pattern
func (s *SupabaseCache) GetKeys(ctx context.Context, pattern string) ([]string, error) {
	s.logger.Debug("Getting cache keys from Supabase", "pattern", pattern)

	// Note: Supabase doesn't support pattern matching like Redis
	// This is a simplified implementation that returns all keys
	result, err := s.client.DB.From("cache_entries").
		Select("key").
		Gt("expires_at", time.Now()).
		Execute("")

	if err != nil {
		s.logger.Error("Failed to get cache keys from Supabase", "error", err)
		return nil, fmt.Errorf("failed to get cache keys: %w", err)
	}

	var entries []CacheEntry
	if err := result.Unmarshal(&entries); err != nil {
		s.logger.Error("Failed to unmarshal cache entries", "error", err)
		return nil, fmt.Errorf("failed to unmarshal cache entries: %w", err)
	}

	var keys []string
	for _, entry := range entries {
		keys = append(keys, entry.Key)
	}

	s.logger.Debug("Retrieved cache keys from Supabase", "count", len(keys))
	return keys, nil
}

// GetStats returns cache statistics
func (s *SupabaseCache) GetStats(ctx context.Context) (map[string]interface{}, error) {
	s.logger.Debug("Getting cache stats from Supabase")

	// Get total entries
	totalResult, err := s.client.DB.From("cache_entries").Select("count", false).Execute("")
	if err != nil {
		return nil, fmt.Errorf("failed to get total entries: %w", err)
	}

	var totalCount []map[string]interface{}
	if err := totalResult.Unmarshal(&totalCount); err != nil {
		return nil, fmt.Errorf("failed to unmarshal total count: %w", err)
	}

	// Get active entries (not expired)
	activeResult, err := s.client.DB.From("cache_entries").
		Select("count", false).
		Gt("expires_at", time.Now()).
		Execute("")
	if err != nil {
		return nil, fmt.Errorf("failed to get active entries: %w", err)
	}

	var activeCount []map[string]interface{}
	if err := activeResult.Unmarshal(&activeCount); err != nil {
		return nil, fmt.Errorf("failed to unmarshal active count: %w", err)
	}

	stats := map[string]interface{}{
		"provider":        "supabase",
		"total_entries":   totalCount[0]["count"],
		"active_entries":  activeCount[0]["count"],
		"expired_entries": totalCount[0]["count"].(float64) - activeCount[0]["count"].(float64),
	}

	s.logger.Info("Retrieved cache stats from Supabase", "stats", stats)
	return stats, nil
}

// CleanupExpired removes expired cache entries
func (s *SupabaseCache) CleanupExpired(ctx context.Context) error {
	s.logger.Debug("Cleaning up expired cache entries in Supabase")

	_, err := s.client.DB.From("cache_entries").
		Delete("", "").
		Lt("expires_at", time.Now()).
		Execute("")

	if err != nil {
		s.logger.Error("Failed to cleanup expired cache entries in Supabase", "error", err)
		return fmt.Errorf("failed to cleanup expired entries: %w", err)
	}

	s.logger.Info("Successfully cleaned up expired cache entries in Supabase")
	return nil
}
