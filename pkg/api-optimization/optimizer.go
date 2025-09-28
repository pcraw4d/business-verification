package apioptimization

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// APIOptimizer provides advanced API response optimization capabilities
type APIOptimizer struct {
	config *APIConfig
}

// APIConfig contains API optimization settings
type APIConfig struct {
	// Compression Settings
	EnableGzipCompression bool
	CompressionLevel      int
	MinCompressionSize    int

	// Caching Settings
	EnableResponseCache bool
	DefaultCacheTTL     time.Duration
	MaxCacheSize        int
	CacheKeyPrefix      string

	// Pagination Settings
	DefaultPageSize        int
	MaxPageSize            int
	EnableCursorPagination bool

	// Response Optimization
	EnableETags        bool
	EnableLastModified bool
	EnableCORS         bool
	CORSOrigins        []string

	// Performance Settings
	EnableRequestTiming   bool
	EnableResponseMetrics bool
	SlowRequestThreshold  time.Duration
}

// DefaultAPIConfig returns optimized API configuration
func DefaultAPIConfig() *APIConfig {
	return &APIConfig{
		// Compression Settings
		EnableGzipCompression: true,
		CompressionLevel:      6,    // Balanced compression
		MinCompressionSize:    1024, // 1KB minimum

		// Caching Settings
		EnableResponseCache: true,
		DefaultCacheTTL:     5 * time.Minute,
		MaxCacheSize:        1000,
		CacheKeyPrefix:      "api_cache:",

		// Pagination Settings
		DefaultPageSize:        20,
		MaxPageSize:            100,
		EnableCursorPagination: true,

		// Response Optimization
		EnableETags:        true,
		EnableLastModified: true,
		EnableCORS:         true,
		CORSOrigins:        []string{"*"},

		// Performance Settings
		EnableRequestTiming:   true,
		EnableResponseMetrics: true,
		SlowRequestThreshold:  100 * time.Millisecond,
	}
}

// NewAPIOptimizer creates a new API optimizer
func NewAPIOptimizer(config *APIConfig) *APIOptimizer {
	if config == nil {
		config = DefaultAPIConfig()
	}

	return &APIOptimizer{
		config: config,
	}
}

// OptimizeResponse applies all response optimizations
func (ao *APIOptimizer) OptimizeResponse(w http.ResponseWriter, r *http.Request, data interface{}) error {
	start := time.Now()

	// Set CORS headers
	if ao.config.EnableCORS {
		ao.setCORSHeaders(w, r)
	}

	// Set ETag and Last-Modified headers
	if ao.config.EnableETags {
		ao.setCacheHeaders(w, data)
	}

	// Check if client accepts gzip
	acceptsGzip := strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

	// Serialize response
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	// Apply compression if enabled and beneficial
	if ao.config.EnableGzipCompression && acceptsGzip && len(jsonData) >= ao.config.MinCompressionSize {
		compressed, err := ao.compressData(jsonData)
		if err == nil {
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(len(compressed)))
			w.WriteHeader(http.StatusOK)
			w.Write(compressed)
		} else {
			// Fallback to uncompressed
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
			w.WriteHeader(http.StatusOK)
			w.Write(jsonData)
		}
	} else {
		// Uncompressed response
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Length", strconv.Itoa(len(jsonData)))
		w.WriteHeader(http.StatusOK)
		w.Write(jsonData)
	}

	// Add performance headers
	if ao.config.EnableRequestTiming {
		duration := time.Since(start)
		w.Header().Set("X-Response-Time", duration.String())

		if duration > ao.config.SlowRequestThreshold {
			w.Header().Set("X-Slow-Request", "true")
		}
	}

	return nil
}

// PaginateResponse creates a paginated response
func (ao *APIOptimizer) PaginateResponse(w http.ResponseWriter, r *http.Request, data []interface{}, total int) error {
	// Parse pagination parameters
	page, pageSize := ao.parsePaginationParams(r)

	// Validate page size
	if pageSize > ao.config.MaxPageSize {
		pageSize = ao.config.MaxPageSize
	}
	if pageSize <= 0 {
		pageSize = ao.config.DefaultPageSize
	}

	// Calculate pagination
	offset := (page - 1) * pageSize
	end := offset + pageSize

	if end > len(data) {
		end = len(data)
	}

	// Get page data
	pageData := data[offset:end]

	// Create paginated response
	response := map[string]interface{}{
		"data": pageData,
		"pagination": map[string]interface{}{
			"page":         page,
			"page_size":    pageSize,
			"total_items":  total,
			"total_pages":  (total + pageSize - 1) / pageSize,
			"has_next":     end < total,
			"has_previous": page > 1,
		},
		"meta": map[string]interface{}{
			"timestamp": time.Now().Format(time.RFC3339),
			"count":     len(pageData),
		},
	}

	return ao.OptimizeResponse(w, r, response)
}

// CompressData compresses data using gzip
func (ao *APIOptimizer) compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	writer, err := gzip.NewWriterLevel(&buf, ao.config.CompressionLevel)
	if err != nil {
		return nil, err
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// setCORSHeaders sets CORS headers
func (ao *APIOptimizer) setCORSHeaders(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")

	// Check if origin is allowed
	allowed := false
	for _, allowedOrigin := range ao.config.CORSOrigins {
		if allowedOrigin == "*" || allowedOrigin == origin {
			allowed = true
			break
		}
	}

	if allowed {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		w.Header().Set("Access-Control-Max-Age", "86400")
	}
}

// setCacheHeaders sets cache-related headers
func (ao *APIOptimizer) setCacheHeaders(w http.ResponseWriter, data interface{}) {
	// Generate ETag based on data hash
	etag := ao.generateETag(data)
	w.Header().Set("ETag", etag)

	// Set Last-Modified
	if ao.config.EnableLastModified {
		w.Header().Set("Last-Modified", time.Now().Format(http.TimeFormat))
	}

	// Set cache control
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(ao.config.DefaultCacheTTL.Seconds())))
}

// generateETag generates an ETag for the data
func (ao *APIOptimizer) generateETag(data interface{}) string {
	// Simple ETag generation based on data hash
	jsonData, _ := json.Marshal(data)
	return fmt.Sprintf(`"%x"`, len(jsonData))
}

// parsePaginationParams parses pagination parameters from request
func (ao *APIOptimizer) parsePaginationParams(r *http.Request) (int, int) {
	page := 1
	pageSize := ao.config.DefaultPageSize

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if sizeStr := r.URL.Query().Get("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil && s > 0 {
			pageSize = s
		}
	}

	return page, pageSize
}

// Middleware creates HTTP middleware for API optimization
func (ao *APIOptimizer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Handle preflight requests
		if r.Method == "OPTIONS" && ao.config.EnableCORS {
			ao.setCORSHeaders(w, r)
			w.WriteHeader(http.StatusOK)
			return
		}

		// Create response writer wrapper to capture response
		wrapper := &ResponseWriterWrapper{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// Call next handler
		next.ServeHTTP(wrapper, r)

		// Add performance headers
		if ao.config.EnableRequestTiming {
			duration := time.Since(start)
			w.Header().Set("X-Response-Time", duration.String())
			w.Header().Set("X-Request-ID", ao.generateRequestID())

			if duration > ao.config.SlowRequestThreshold {
				w.Header().Set("X-Slow-Request", "true")
			}
		}
	})
}

// ResponseWriterWrapper wraps http.ResponseWriter to capture status code
type ResponseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *ResponseWriterWrapper) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// generateRequestID generates a unique request ID
func (ao *APIOptimizer) generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// GetOptimizationStats returns API optimization statistics
func (ao *APIOptimizer) GetOptimizationStats() *OptimizationStats {
	return &OptimizationStats{
		Config:    ao.config,
		Timestamp: time.Now(),
		Features: map[string]bool{
			"gzip_compression":  ao.config.EnableGzipCompression,
			"response_cache":    ao.config.EnableResponseCache,
			"etags":             ao.config.EnableETags,
			"cors":              ao.config.EnableCORS,
			"request_timing":    ao.config.EnableRequestTiming,
			"cursor_pagination": ao.config.EnableCursorPagination,
		},
	}
}

// OptimizationStats contains API optimization statistics
type OptimizationStats struct {
	Config    *APIConfig      `json:"config"`
	Timestamp time.Time       `json:"timestamp"`
	Features  map[string]bool `json:"features"`
}
