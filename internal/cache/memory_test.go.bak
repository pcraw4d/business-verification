package cache

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNewMemoryCache(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      30 * time.Minute,
		MaxSize:         100,
		CleanupInterval: 1 * time.Minute,
	}

	cache := NewMemoryCache(config)
	if cache == nil {
		t.Fatal("NewMemoryCache returned nil")
	}

	// Test default config
	cache = NewMemoryCache(nil)
	if cache == nil {
		t.Fatal("NewMemoryCache with nil config returned nil")
	}
}

func TestMemoryCache_GetSet(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	// Test basic get/set
	key := "test-key"
	value := []byte("test-value")

	err := cache.Set(ctx, key, value, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	retrieved, err := cache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != string(value) {
		t.Errorf("Expected %s, got %s", string(value), string(retrieved))
	}
}

func TestMemoryCache_GetNotFound(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	_, err := cache.Get(ctx, "nonexistent-key")
	if err == nil {
		t.Fatal("Expected error for nonexistent key")
	}

	if !IsNotFound(err) {
		t.Errorf("Expected CacheNotFoundError, got %T", err)
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	key := "test-key"
	value := []byte("test-value")

	// Set value
	err := cache.Set(ctx, key, value, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify it exists
	exists, err := cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("Key should exist after setting")
	}

	// Delete value
	err = cache.Delete(ctx, key)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it doesn't exist
	exists, err = cache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("Key should not exist after deleting")
	}
}

func TestMemoryCache_DeleteNotFound(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	err := cache.Delete(ctx, "nonexistent-key")
	if err == nil {
		t.Fatal("Expected error for deleting nonexistent key")
	}

	if !IsNotFound(err) {
		t.Errorf("Expected CacheNotFoundError, got %T", err)
	}
}

func TestMemoryCache_TTL(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	key := "test-key"
	value := []byte("test-value")
	ttl := 100 * time.Millisecond

	// Set value with TTL
	err := cache.Set(ctx, key, value, ttl)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Verify TTL
	retrievedTTL, err := cache.GetTTL(ctx, key)
	if err != nil {
		t.Fatalf("GetTTL failed: %v", err)
	}

	if retrievedTTL <= 0 || retrievedTTL > ttl {
		t.Errorf("Expected TTL between 0 and %v, got %v", ttl, retrievedTTL)
	}

	// Wait for expiration
	time.Sleep(ttl + 50*time.Millisecond)

	// Verify key is expired
	_, err = cache.Get(ctx, key)
	if err == nil {
		t.Fatal("Expected error for expired key")
	}

	if !IsNotFound(err) {
		t.Errorf("Expected CacheNotFoundError, got %T", err)
	}
}

func TestMemoryCache_SetTTL(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	key := "test-key"
	value := []byte("test-value")

	// Set value without TTL
	err := cache.Set(ctx, key, value, 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Set TTL
	newTTL := 1 * time.Second
	err = cache.SetTTL(ctx, key, newTTL)
	if err != nil {
		t.Fatalf("SetTTL failed: %v", err)
	}

	// Verify TTL
	retrievedTTL, err := cache.GetTTL(ctx, key)
	if err != nil {
		t.Fatalf("GetTTL failed: %v", err)
	}

	if retrievedTTL <= 0 || retrievedTTL > newTTL {
		t.Errorf("Expected TTL between 0 and %v, got %v", newTTL, retrievedTTL)
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	// Set multiple values
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		err := cache.Set(ctx, key, []byte("value"), 0)
		if err != nil {
			t.Fatalf("Set failed for %s: %v", key, err)
		}
	}

	// Clear cache
	err := cache.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify all keys are gone
	for _, key := range keys {
		exists, err := cache.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed for %s: %v", key, err)
		}
		if exists {
			t.Errorf("Key %s should not exist after clear", key)
		}
	}
}

func TestMemoryCache_GetStats(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	// Set some values
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		err := cache.Set(ctx, key, []byte("value"), 0)
		if err != nil {
			t.Fatalf("Set failed for %s: %v", key, err)
		}
	}

	// Get some values to generate hits
	for _, key := range keys {
		_, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed for %s: %v", key, err)
		}
	}

	// Get stats
	stats, err := cache.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.Size != 3 {
		t.Errorf("Expected size 3, got %d", stats.Size)
	}

	if stats.HitCount != 3 {
		t.Errorf("Expected hit count 3, got %d", stats.HitCount)
	}

	if stats.HitRate != 1.0 {
		t.Errorf("Expected hit rate 1.0, got %f", stats.HitRate)
	}
}

func TestMemoryCache_Eviction(t *testing.T) {
	config := &CacheConfig{
		Type:            MemoryCache,
		DefaultTTL:      1 * time.Hour,
		MaxSize:         2,
		CleanupInterval: 1 * time.Minute,
	}

	cache := NewMemoryCache(config)
	ctx := context.Background()

	// Set more values than max size
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		err := cache.Set(ctx, key, []byte("value"), 0)
		if err != nil {
			t.Fatalf("Set failed for %s: %v", key, err)
		}
	}

	// Verify only 2 keys exist (due to eviction)
	existingCount := 0
	for _, key := range keys {
		exists, err := cache.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed for %s: %v", key, err)
		}
		if exists {
			existingCount++
		}
	}

	if existingCount != 2 {
		t.Errorf("Expected 2 existing keys after eviction, got %d", existingCount)
	}

	// Check stats
	stats, err := cache.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats failed: %v", err)
	}

	if stats.Size != 2 {
		t.Errorf("Expected size 2, got %d", stats.Size)
	}

	if stats.EvictionCount != 1 {
		t.Errorf("Expected eviction count 1, got %d", stats.EvictionCount)
	}
}

func TestMemoryCache_GetKeys(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	// Set some values
	keys := []string{"user:1", "user:2", "product:1", "product:2"}
	for _, key := range keys {
		err := cache.Set(ctx, key, []byte("value"), 0)
		if err != nil {
			t.Fatalf("Set failed for %s: %v", key, err)
		}
	}

	// Get all keys
	retrievedKeys, err := cache.GetKeys(ctx, "")
	if err != nil {
		t.Fatalf("GetKeys failed: %v", err)
	}

	if len(retrievedKeys) != 4 {
		t.Errorf("Expected 4 keys, got %d", len(retrievedKeys))
	}

	// Check that all expected keys are present
	keyMap := make(map[string]bool)
	for _, key := range retrievedKeys {
		keyMap[key] = true
	}

	for _, expectedKey := range keys {
		if !keyMap[expectedKey] {
			t.Errorf("Expected key %s not found", expectedKey)
		}
	}
}

func TestMemoryCache_GetEntries(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	// Set some values
	entries := map[string][]byte{
		"key1": []byte("value1"),
		"key2": []byte("value2"),
		"key3": []byte("value3"),
	}

	for key, value := range entries {
		err := cache.Set(ctx, key, value, 0)
		if err != nil {
			t.Fatalf("Set failed for %s: %v", key, err)
		}
	}

	// Get entries
	keys := []string{"key1", "key2", "key3"}
	retrievedEntries, err := cache.GetEntries(ctx, keys)
	if err != nil {
		t.Fatalf("GetEntries failed: %v", err)
	}

	if len(retrievedEntries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(retrievedEntries))
	}

	// Verify entries
	for key, expectedValue := range entries {
		entry, exists := retrievedEntries[key]
		if !exists {
			t.Errorf("Expected entry for key %s not found", key)
			continue
		}

		if string(entry.Value) != string(expectedValue) {
			t.Errorf("Expected value %s for key %s, got %s",
				string(expectedValue), key, string(entry.Value))
		}
	}
}

func TestMemoryCache_SetEntries(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	// Create entries
	entries := map[string]*CacheEntry{
		"key1": {
			Key:   "key1",
			Value: []byte("value1"),
			TTL:   1 * time.Hour,
		},
		"key2": {
			Key:   "key2",
			Value: []byte("value2"),
			TTL:   1 * time.Hour,
		},
	}

	// Set entries
	err := cache.SetEntries(ctx, entries)
	if err != nil {
		t.Fatalf("SetEntries failed: %v", err)
	}

	// Verify entries
	for key, expectedEntry := range entries {
		value, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed for %s: %v", key, err)
		}

		if string(value) != string(expectedEntry.Value) {
			t.Errorf("Expected value %s for key %s, got %s",
				string(expectedEntry.Value), key, string(value))
		}
	}
}

func TestMemoryCache_DeleteEntries(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	// Set some values
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		err := cache.Set(ctx, key, []byte("value"), 0)
		if err != nil {
			t.Fatalf("Set failed for %s: %v", key, err)
		}
	}

	// Delete entries
	deleteKeys := []string{"key1", "key2"}
	err := cache.DeleteEntries(ctx, deleteKeys)
	if err != nil {
		t.Fatalf("DeleteEntries failed: %v", err)
	}

	// Verify deleted keys are gone
	for _, key := range deleteKeys {
		exists, err := cache.Exists(ctx, key)
		if err != nil {
			t.Fatalf("Exists failed for %s: %v", key, err)
		}
		if exists {
			t.Errorf("Key %s should not exist after deletion", key)
		}
	}

	// Verify remaining key exists
	exists, err := cache.Exists(ctx, "key3")
	if err != nil {
		t.Fatalf("Exists failed for key3: %v", err)
	}
	if !exists {
		t.Error("Key3 should exist after deletion")
	}
}

func TestMemoryCache_ConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	// Test concurrent reads and writes
	done := make(chan bool, 10)

	// Start goroutines
	for i := 0; i < 5; i++ {
		go func(id int) {
			key := fmt.Sprintf("key%d", id)
			value := []byte(fmt.Sprintf("value%d", id))

			// Set value
			err := cache.Set(ctx, key, value, 0)
			if err != nil {
				t.Errorf("Set failed in goroutine %d: %v", id, err)
			}

			// Get value
			retrieved, err := cache.Get(ctx, key)
			if err != nil {
				t.Errorf("Get failed in goroutine %d: %v", id, err)
			} else if string(retrieved) != string(value) {
				t.Errorf("Value mismatch in goroutine %d", id)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestMemoryCache_Close(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	// Set a value
	err := cache.Set(ctx, "key", []byte("value"), 0)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Close cache
	err = cache.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Verify cache is still functional (memory cache doesn't actually close)
	value, err := cache.Get(ctx, "key")
	if err != nil {
		t.Fatalf("Get failed after close: %v", err)
	}

	if string(value) != "value" {
		t.Errorf("Expected value 'value', got %s", string(value))
	}
}

func TestMemoryCache_String(t *testing.T) {
	cache := NewMemoryCache(nil)
	ctx := context.Background()

	// Set some values
	keys := []string{"key1", "key2", "key3"}
	for _, key := range keys {
		err := cache.Set(ctx, key, []byte("value"), 0)
		if err != nil {
			t.Fatalf("Set failed for %s: %v", key, err)
		}
	}

	// Get some values to generate hits
	for _, key := range keys {
		_, err := cache.Get(ctx, key)
		if err != nil {
			t.Fatalf("Get failed for %s: %v", key, err)
		}
	}

	// Test String method
	str := cache.String()
	if str == "" {
		t.Error("String method returned empty string")
	}

	// Verify it contains expected information
	if !strings.Contains(str, "MemoryCache") {
		t.Error("String should contain 'MemoryCache'")
	}
}
