package cache

import "sync"

// InMemoryCacheMetrics is a simple thread-safe metrics collector for cache ops
type InMemoryCacheMetrics struct {
	mu          sync.Mutex
	hits        int64
	misses      int64
	sets        int64
	deletes     int64
	evictions   int64
	expirations int64
}

func NewInMemoryCacheMetrics() *InMemoryCacheMetrics { return &InMemoryCacheMetrics{} }

func (m *InMemoryCacheMetrics) RecordHit(key string) {
	m.mu.Lock()
	m.hits++
	m.mu.Unlock()
}

func (m *InMemoryCacheMetrics) RecordMiss(key string) {
	m.mu.Lock()
	m.misses++
	m.mu.Unlock()
}

func (m *InMemoryCacheMetrics) RecordSet(key string, size int64) {
	m.mu.Lock()
	m.sets++
	m.mu.Unlock()
}

func (m *InMemoryCacheMetrics) RecordDelete(key string) {
	m.mu.Lock()
	m.deletes++
	m.mu.Unlock()
}

func (m *InMemoryCacheMetrics) RecordEviction(key string, reason string) {
	m.mu.Lock()
	m.evictions++
	m.mu.Unlock()
}

func (m *InMemoryCacheMetrics) RecordExpiration(key string) {
	m.mu.Lock()
	m.expirations++
	m.mu.Unlock()
}
