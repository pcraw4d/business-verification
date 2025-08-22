package cache

import (
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"

	"go.uber.org/zap"
)

// CacheKeyManager handles cache key generation and management
type CacheKeyManager struct {
	// Configuration
	config CacheConfig
	
	// Thread safety
	mu sync.RWMutex
	
	// Key statistics
	keyStats     *KeyStats
	keyStatsLock sync.RWMutex
	
	// Logging
	logger *zap.Logger
}

// KeyStats holds key generation statistics
type KeyStats struct {
	KeysGenerated int64   `json:"keys_generated"`
	KeyCollisions int64   `json:"key_collisions"`
	AverageLength float64 `json:"average_length"`
	TotalLength   int64   `json:"total_length"`
}

// NewCacheKeyManager creates a new cache key manager
func NewCacheKeyManager(config CacheConfig, logger *zap.Logger) *CacheKeyManager {
	return &CacheKeyManager{
		config:   config,
		keyStats: &KeyStats{},
		logger:   logger,
	}
}

// GenerateKey generates a cache key from the given input
func (ckm *CacheKeyManager) GenerateKey(input string) string {
	ckm.mu.RLock()
	defer ckm.mu.RUnlock()

	// Add prefix if configured
	key := input
	if ckm.config.KeyPrefix != "" {
		key = ckm.config.KeyPrefix + ckm.config.KeySeparator + key
	}

	// Apply hash algorithm if configured
	if ckm.config.KeyHashAlgorithm != "" {
		key = ckm.hashKey(key)
	}

	// Update statistics
	ckm.updateKeyStats(key)

	return key
}

// GenerateKeyWithNamespace generates a cache key with a namespace
func (ckm *CacheKeyManager) GenerateKeyWithNamespace(namespace, key string) string {
	ckm.mu.RLock()
	defer ckm.mu.RUnlock()

	// Combine namespace and key
	fullKey := namespace + ckm.config.KeySeparator + key

	// Add prefix if configured
	if ckm.config.KeyPrefix != "" {
		fullKey = ckm.config.KeyPrefix + ckm.config.KeySeparator + fullKey
	}

	// Apply hash algorithm if configured
	if ckm.config.KeyHashAlgorithm != "" {
		fullKey = ckm.hashKey(fullKey)
	}

	// Update statistics
	ckm.updateKeyStats(fullKey)

	return fullKey
}

// GeneratePatternKey generates a pattern key for invalidation
func (ckm *CacheKeyManager) GeneratePatternKey(pattern string) string {
	ckm.mu.RLock()
	defer ckm.mu.RUnlock()

	// Add prefix if configured
	key := pattern
	if ckm.config.KeyPrefix != "" {
		key = ckm.config.KeyPrefix + ckm.config.KeySeparator + key
	}

	return key
}

// ValidateKey validates if a key is properly formatted
func (ckm *CacheKeyManager) ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if len(key) > 250 {
		return fmt.Errorf("key too long (max 250 characters)")
	}

	// Check for invalid characters
	invalidChars := []string{"\x00", "\n", "\r", "\t"}
	for _, char := range invalidChars {
		if strings.Contains(key, char) {
			return fmt.Errorf("key contains invalid character: %q", char)
		}
	}

	return nil
}

// ExtractNamespace extracts the namespace from a key
func (ckm *CacheKeyManager) ExtractNamespace(key string) (string, string, error) {
	ckm.mu.RLock()
	defer ckm.mu.RUnlock()

	// Remove prefix if present
	cleanKey := key
	if ckm.config.KeyPrefix != "" && strings.HasPrefix(key, ckm.config.KeyPrefix) {
		cleanKey = strings.TrimPrefix(key, ckm.config.KeyPrefix+ckm.config.KeySeparator)
	}

	// Split by separator
	parts := strings.SplitN(cleanKey, ckm.config.KeySeparator, 2)
	if len(parts) < 2 {
		return "", cleanKey, nil // No namespace
	}

	return parts[0], parts[1], nil
}

// GetKeyStats returns key generation statistics
func (ckm *CacheKeyManager) GetKeyStats() *KeyStats {
	ckm.keyStatsLock.RLock()
	defer ckm.keyStatsLock.RUnlock()

	stats := *ckm.keyStats
	return &stats
}

// ResetKeyStats resets key generation statistics
func (ckm *CacheKeyManager) ResetKeyStats() {
	ckm.keyStatsLock.Lock()
	defer ckm.keyStatsLock.Unlock()

	ckm.keyStats = &KeyStats{}
}

// Helper methods

func (ckm *CacheKeyManager) hashKey(key string) string {
	switch strings.ToUpper(ckm.config.KeyHashAlgorithm) {
	case "MD5":
		return ckm.hashMD5(key)
	case "SHA256":
		return ckm.hashSHA256(key)
	case "FNV":
		return ckm.hashFNV(key)
	default:
		// Default to MD5
		return ckm.hashMD5(key)
	}
}

func (ckm *CacheKeyManager) hashMD5(key string) string {
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%x", hash)
}

func (ckm *CacheKeyManager) hashSHA256(key string) string {
	hash := sha256.Sum256([]byte(key))
	return fmt.Sprintf("%x", hash)
}

func (ckm *CacheKeyManager) hashFNV(key string) string {
	h := fnv.New64a()
	h.Write([]byte(key))
	return fmt.Sprintf("%x", h.Sum64())
}

func (ckm *CacheKeyManager) updateKeyStats(key string) {
	ckm.keyStatsLock.Lock()
	defer ckm.keyStatsLock.Unlock()

	ckm.keyStats.KeysGenerated++
	ckm.keyStats.TotalLength += int64(len(key))
	ckm.keyStats.AverageLength = float64(ckm.keyStats.TotalLength) / float64(ckm.keyStats.KeysGenerated)
}
