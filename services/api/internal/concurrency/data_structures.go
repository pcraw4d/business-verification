package concurrency

import (
	"sync"
	"time"
)

// ThreadSafeDataStructures provides thread-safe implementations of common data structures
type ThreadSafeDataStructures struct {
	// Thread-safe maps
	stringMap   *ThreadSafeMap[string, interface{}]
	intMap      *ThreadSafeMap[int, interface{}]
	resourceMap *ThreadSafeMap[string, *Resource]

	// Thread-safe slices
	stringSlice *ThreadSafeSlice[string]
	intSlice    *ThreadSafeSlice[int]

	// Thread-safe queues
	requestQueue  *ThreadSafeQueue[*ConcurrentRequest]
	responseQueue *ThreadSafeQueue[*ConcurrentResponse]

	// Thread-safe counters
	counter *ThreadSafeCounter
}

// NewThreadSafeDataStructures creates a new instance of thread-safe data structures
func NewThreadSafeDataStructures() *ThreadSafeDataStructures {
	return &ThreadSafeDataStructures{
		stringMap:     NewThreadSafeMap[string, interface{}](),
		intMap:        NewThreadSafeMap[int, interface{}](),
		resourceMap:   NewThreadSafeMap[string, *Resource](),
		stringSlice:   NewThreadSafeSlice[string](),
		intSlice:      NewThreadSafeSlice[int](),
		requestQueue:  NewThreadSafeQueue[*ConcurrentRequest](),
		responseQueue: NewThreadSafeQueue[*ConcurrentResponse](),
		counter:       NewThreadSafeCounter(),
	}
}

// ThreadSafeMap provides a thread-safe map implementation
type ThreadSafeMap[K comparable, V any] struct {
	data  map[K]V
	mutex sync.RWMutex
}

// NewThreadSafeMap creates a new thread-safe map
func NewThreadSafeMap[K comparable, V any]() *ThreadSafeMap[K, V] {
	return &ThreadSafeMap[K, V]{
		data: make(map[K]V),
	}
}

// Set sets a key-value pair in the map
func (m *ThreadSafeMap[K, V]) Set(key K, value V) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data[key] = value
}

// Get retrieves a value by key
func (m *ThreadSafeMap[K, V]) Get(key K) (V, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	value, exists := m.data[key]
	return value, exists
}

// Delete removes a key-value pair from the map
func (m *ThreadSafeMap[K, V]) Delete(key K) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	delete(m.data, key)
}

// Has checks if a key exists in the map
func (m *ThreadSafeMap[K, V]) Has(key K) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, exists := m.data[key]
	return exists
}

// Size returns the number of key-value pairs in the map
func (m *ThreadSafeMap[K, V]) Size() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.data)
}

// Keys returns all keys in the map
func (m *ThreadSafeMap[K, V]) Keys() []K {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	keys := make([]K, 0, len(m.data))
	for key := range m.data {
		keys = append(keys, key)
	}
	return keys
}

// Values returns all values in the map
func (m *ThreadSafeMap[K, V]) Values() []V {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	values := make([]V, 0, len(m.data))
	for _, value := range m.data {
		values = append(values, value)
	}
	return values
}

// Clear removes all key-value pairs from the map
func (m *ThreadSafeMap[K, V]) Clear() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.data = make(map[K]V)
}

// Copy returns a copy of the map
func (m *ThreadSafeMap[K, V]) Copy() map[K]V {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	result := make(map[K]V)
	for key, value := range m.data {
		result[key] = value
	}
	return result
}

// ThreadSafeSlice provides a thread-safe slice implementation
type ThreadSafeSlice[T any] struct {
	data  []T
	mutex sync.RWMutex
}

// NewThreadSafeSlice creates a new thread-safe slice
func NewThreadSafeSlice[T any]() *ThreadSafeSlice[T] {
	return &ThreadSafeSlice[T]{
		data: make([]T, 0),
	}
}

// Append adds an element to the end of the slice
func (s *ThreadSafeSlice[T]) Append(value T) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = append(s.data, value)
}

// Get retrieves an element by index
func (s *ThreadSafeSlice[T]) Get(index int) (T, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if index < 0 || index >= len(s.data) {
		var zero T
		return zero, false
	}
	return s.data[index], true
}

// Set sets an element at a specific index
func (s *ThreadSafeSlice[T]) Set(index int, value T) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if index < 0 || index >= len(s.data) {
		return false
	}
	s.data[index] = value
	return true
}

// Remove removes an element at a specific index
func (s *ThreadSafeSlice[T]) Remove(index int) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if index < 0 || index >= len(s.data) {
		return false
	}

	s.data = append(s.data[:index], s.data[index+1:]...)
	return true
}

// Length returns the length of the slice
func (s *ThreadSafeSlice[T]) Length() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.data)
}

// Clear removes all elements from the slice
func (s *ThreadSafeSlice[T]) Clear() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = make([]T, 0)
}

// Copy returns a copy of the slice
func (s *ThreadSafeSlice[T]) Copy() []T {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := make([]T, len(s.data))
	copy(result, s.data)
	return result
}

// ThreadSafeQueue provides a thread-safe queue implementation
type ThreadSafeQueue[T any] struct {
	data  []T
	mutex sync.RWMutex
}

// NewThreadSafeQueue creates a new thread-safe queue
func NewThreadSafeQueue[T any]() *ThreadSafeQueue[T] {
	return &ThreadSafeQueue[T]{
		data: make([]T, 0),
	}
}

// Enqueue adds an element to the end of the queue
func (q *ThreadSafeQueue[T]) Enqueue(value T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.data = append(q.data, value)
}

// Dequeue removes and returns the first element from the queue
func (q *ThreadSafeQueue[T]) Dequeue() (T, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.data) == 0 {
		var zero T
		return zero, false
	}

	value := q.data[0]
	q.data = q.data[1:]
	return value, true
}

// Peek returns the first element without removing it
func (q *ThreadSafeQueue[T]) Peek() (T, bool) {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	if len(q.data) == 0 {
		var zero T
		return zero, false
	}

	return q.data[0], true
}

// Size returns the number of elements in the queue
func (q *ThreadSafeQueue[T]) Size() int {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.data)
}

// IsEmpty checks if the queue is empty
func (q *ThreadSafeQueue[T]) IsEmpty() bool {
	q.mutex.RLock()
	defer q.mutex.RUnlock()
	return len(q.data) == 0
}

// Clear removes all elements from the queue
func (q *ThreadSafeQueue[T]) Clear() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.data = make([]T, 0)
}

// Copy returns a copy of the queue
func (q *ThreadSafeQueue[T]) Copy() []T {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	result := make([]T, len(q.data))
	copy(result, q.data)
	return result
}

// ThreadSafeCounter provides a thread-safe counter implementation
type ThreadSafeCounter struct {
	value int64
	mutex sync.RWMutex
}

// NewThreadSafeCounter creates a new thread-safe counter
func NewThreadSafeCounter() *ThreadSafeCounter {
	return &ThreadSafeCounter{
		value: 0,
	}
}

// Increment increments the counter by 1
func (c *ThreadSafeCounter) Increment() int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value++
	return c.value
}

// Decrement decrements the counter by 1
func (c *ThreadSafeCounter) Decrement() int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value--
	return c.value
}

// Add adds a value to the counter
func (c *ThreadSafeCounter) Add(value int64) int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value += value
	return c.value
}

// Subtract subtracts a value from the counter
func (c *ThreadSafeCounter) Subtract(value int64) int64 {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value -= value
	return c.value
}

// Get returns the current value of the counter
func (c *ThreadSafeCounter) Get() int64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.value
}

// Set sets the counter to a specific value
func (c *ThreadSafeCounter) Set(value int64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value = value
}

// Reset resets the counter to 0
func (c *ThreadSafeCounter) Reset() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value = 0
}

// ThreadSafeCache provides a thread-safe cache implementation with TTL
type ThreadSafeCache[K comparable, V any] struct {
	data  map[K]*cacheEntry[V]
	mutex sync.RWMutex
}

type cacheEntry[V any] struct {
	value     V
	expiresAt time.Time
}

// NewThreadSafeCache creates a new thread-safe cache
func NewThreadSafeCache[K comparable, V any]() *ThreadSafeCache[K, V] {
	return &ThreadSafeCache[K, V]{
		data: make(map[K]*cacheEntry[V]),
	}
}

// Set sets a key-value pair with optional TTL
func (c *ThreadSafeCache[K, V]) Set(key K, value V, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiresAt := time.Now().Add(ttl)
	c.data[key] = &cacheEntry[V]{
		value:     value,
		expiresAt: expiresAt,
	}
}

// Get retrieves a value by key, returns false if expired or not found
func (c *ThreadSafeCache[K, V]) Get(key K) (V, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		var zero V
		return zero, false
	}

	if time.Now().After(entry.expiresAt) {
		var zero V
		return zero, false
	}

	return entry.value, true
}

// Delete removes a key-value pair from the cache
func (c *ThreadSafeCache[K, V]) Delete(key K) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

// Clear removes all expired entries
func (c *ThreadSafeCache[K, V]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, entry := range c.data {
		if now.After(entry.expiresAt) {
			delete(c.data, key)
		}
	}
}

// Size returns the number of non-expired entries
func (c *ThreadSafeCache[K, V]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	now := time.Now()
	count := 0
	for _, entry := range c.data {
		if now.Before(entry.expiresAt) {
			count++
		}
	}
	return count
}
