package cache

import "errors"

// Cache-specific errors
var (
	ErrCacheMiss       = errors.New("cache miss")
	ErrCacheFull       = errors.New("cache is full")
	ErrInvalidKey      = errors.New("invalid cache key")
	ErrSerialization   = errors.New("serialization error")
	ErrDeserialization = errors.New("deserialization error")
)
