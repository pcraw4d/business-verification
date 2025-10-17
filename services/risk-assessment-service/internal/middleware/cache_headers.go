package middleware

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// CacheHeadersMiddleware adds appropriate cache headers to responses
type CacheHeadersMiddleware struct {
	logger *zap.Logger
	config *CacheHeadersConfig
}

// CacheHeadersConfig represents configuration for cache headers
type CacheHeadersConfig struct {
	DefaultTTL                 time.Duration `json:"default_ttl"`
	StaticAssetsTTL            time.Duration `json:"static_assets_ttl"`
	APITTL                     time.Duration `json:"api_ttl"`
	ModelPredictionsTTL        time.Duration `json:"model_predictions_ttl"`
	EnableETag                 bool          `json:"enable_etag"`
	EnableVary                 bool          `json:"enable_vary"`
	EnableStaleWhileRevalidate bool          `json:"enable_stale_while_revalidate"`
	StaleWhileRevalidateTTL    time.Duration `json:"stale_while_revalidate_ttl"`
}

// NewCacheHeadersMiddleware creates a new cache headers middleware
func NewCacheHeadersMiddleware(config *CacheHeadersConfig, logger *zap.Logger) *CacheHeadersMiddleware {
	if config == nil {
		config = &CacheHeadersConfig{
			DefaultTTL:                 5 * time.Minute,
			StaticAssetsTTL:            365 * 24 * time.Hour, // 1 year
			APITTL:                     5 * time.Minute,
			ModelPredictionsTTL:        1 * time.Hour,
			EnableETag:                 true,
			EnableVary:                 true,
			EnableStaleWhileRevalidate: true,
			StaleWhileRevalidateTTL:    2 * time.Hour,
		}
	}

	return &CacheHeadersMiddleware{
		logger: logger,
		config: config,
	}
}

// Middleware returns the cache headers middleware function
func (chm *CacheHeadersMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create response writer wrapper
		rw := &cacheHeadersResponseWriter{
			ResponseWriter: w,
			request:        r,
			config:         chm.config,
			logger:         chm.logger,
		}

		// Call next handler
		next.ServeHTTP(rw, r)

		// Add cache headers
		chm.addCacheHeaders(rw, r)
	})
}

// cacheHeadersResponseWriter wraps http.ResponseWriter to capture response data
type cacheHeadersResponseWriter struct {
	http.ResponseWriter
	request        *http.Request
	config         *CacheHeadersConfig
	logger         *zap.Logger
	statusCode     int
	contentLength  int
	body           []byte
	headersWritten bool
}

// WriteHeader captures the status code
func (rw *cacheHeadersResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
	rw.headersWritten = true
}

// Write captures the response body
func (rw *cacheHeadersResponseWriter) Write(b []byte) (int, error) {
	if !rw.headersWritten {
		rw.WriteHeader(http.StatusOK)
	}

	rw.body = append(rw.body, b...)
	rw.contentLength += len(b)

	return rw.ResponseWriter.Write(b)
}

// addCacheHeaders adds appropriate cache headers based on the request path
func (chm *CacheHeadersMiddleware) addCacheHeaders(rw *cacheHeadersResponseWriter, r *http.Request) {
	path := r.URL.Path
	method := r.Method

	// Only add cache headers for GET requests
	if method != http.MethodGet {
		return
	}

	// Determine cache TTL based on path
	ttl := chm.getCacheTTL(path)

	// Add Cache-Control header
	cacheControl := chm.buildCacheControlHeader(ttl, path)
	rw.Header().Set("Cache-Control", cacheControl)

	// Add ETag header if enabled
	if chm.config.EnableETag && len(rw.body) > 0 {
		etag := chm.generateETag(rw.body)
		rw.Header().Set("ETag", etag)

		// Check if client has the same ETag
		if match := r.Header.Get("If-None-Match"); match == etag {
			rw.WriteHeader(http.StatusNotModified)
			return
		}
	}

	// Add Vary header if enabled
	if chm.config.EnableVary {
		vary := chm.buildVaryHeader(path)
		if vary != "" {
			rw.Header().Set("Vary", vary)
		}
	}

	// Add Last-Modified header
	rw.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))

	// Add Expires header
	expires := time.Now().Add(ttl).UTC().Format(http.TimeFormat)
	rw.Header().Set("Expires", expires)

	chm.logger.Debug("Cache headers added",
		zap.String("path", path),
		zap.String("cache_control", cacheControl),
		zap.Duration("ttl", ttl))
}

// getCacheTTL returns the appropriate cache TTL for the given path
func (chm *CacheHeadersMiddleware) getCacheTTL(path string) time.Duration {
	// Static assets
	if chm.isStaticAsset(path) {
		return chm.config.StaticAssetsTTL
	}

	// Model predictions
	if strings.HasPrefix(path, "/api/v1/predictions/") {
		return chm.config.ModelPredictionsTTL
	}

	// API endpoints
	if strings.HasPrefix(path, "/api/") {
		return chm.config.APITTL
	}

	// Default TTL
	return chm.config.DefaultTTL
}

// isStaticAsset checks if the path is a static asset
func (chm *CacheHeadersMiddleware) isStaticAsset(path string) bool {
	staticExtensions := []string{
		".css", ".js", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico",
		".woff", ".woff2", ".ttf", ".eot", ".pdf", ".zip", ".tar", ".gz",
	}

	for _, ext := range staticExtensions {
		if strings.HasSuffix(strings.ToLower(path), ext) {
			return true
		}
	}

	return false
}

// buildCacheControlHeader builds the Cache-Control header value
func (chm *CacheHeadersMiddleware) buildCacheControlHeader(ttl time.Duration, path string) string {
	var directives []string

	// Add public directive for cacheable content
	directives = append(directives, "public")

	// Add max-age directive
	maxAge := int(ttl.Seconds())
	directives = append(directives, fmt.Sprintf("max-age=%d", maxAge))

	// Add stale-while-revalidate for model predictions
	if chm.config.EnableStaleWhileRevalidate && strings.HasPrefix(path, "/api/v1/predictions/") {
		staleTTL := int(chm.config.StaleWhileRevalidateTTL.Seconds())
		directives = append(directives, fmt.Sprintf("stale-while-revalidate=%d", staleTTL))
	}

	// Add no-cache for sensitive endpoints
	if chm.isSensitiveEndpoint(path) {
		directives = []string{"no-cache", "no-store", "must-revalidate"}
	}

	return strings.Join(directives, ", ")
}

// isSensitiveEndpoint checks if the endpoint should not be cached
func (chm *CacheHeadersMiddleware) isSensitiveEndpoint(path string) bool {
	sensitivePaths := []string{
		"/health",
		"/metrics",
		"/admin",
		"/auth",
		"/login",
		"/logout",
	}

	for _, sensitivePath := range sensitivePaths {
		if strings.HasPrefix(path, sensitivePath) {
			return true
		}
	}

	return false
}

// buildVaryHeader builds the Vary header value
func (chm *CacheHeadersMiddleware) buildVaryHeader(path string) string {
	var vary []string

	// Always vary on Accept-Encoding for compressed content
	vary = append(vary, "Accept-Encoding")

	// Vary on Authorization for authenticated endpoints
	if strings.HasPrefix(path, "/api/") {
		vary = append(vary, "Authorization")
	}

	// Vary on Accept for content negotiation
	if strings.HasPrefix(path, "/api/") {
		vary = append(vary, "Accept")
	}

	return strings.Join(vary, ", ")
}

// generateETag generates an ETag for the response body
func (chm *CacheHeadersMiddleware) generateETag(body []byte) string {
	hash := md5.Sum(body)
	return fmt.Sprintf("\"%x\"", hash)
}

// CacheHeadersHandler handles cache header operations
type CacheHeadersHandler struct {
	middleware *CacheHeadersMiddleware
	logger     *zap.Logger
}

// NewCacheHeadersHandler creates a new cache headers handler
func NewCacheHeadersHandler(config *CacheHeadersConfig, logger *zap.Logger) *CacheHeadersHandler {
	return &CacheHeadersHandler{
		middleware: NewCacheHeadersMiddleware(config, logger),
		logger:     logger,
	}
}

// GetCacheHeaders returns the current cache headers configuration
func (chh *CacheHeadersHandler) GetCacheHeaders() *CacheHeadersConfig {
	return chh.middleware.config
}

// UpdateCacheHeaders updates the cache headers configuration
func (chh *CacheHeadersHandler) UpdateCacheHeaders(config *CacheHeadersConfig) {
	chh.middleware.config = config
	chh.logger.Info("Cache headers configuration updated")
}

// GetCacheTTL returns the cache TTL for a given path
func (chh *CacheHeadersHandler) GetCacheTTL(path string) time.Duration {
	return chh.middleware.getCacheTTL(path)
}

// IsCacheable checks if a path should be cached
func (chh *CacheHeadersHandler) IsCacheable(path string, method string) bool {
	if method != http.MethodGet {
		return false
	}

	return !chh.middleware.isSensitiveEndpoint(path)
}

// GetCacheControlHeader returns the Cache-Control header for a given path
func (chh *CacheHeadersHandler) GetCacheControlHeader(path string) string {
	ttl := chh.middleware.getCacheTTL(path)
	return chh.middleware.buildCacheControlHeader(ttl, path)
}

// GetVaryHeader returns the Vary header for a given path
func (chh *CacheHeadersHandler) GetVaryHeader(path string) string {
	return chh.middleware.buildVaryHeader(path)
}

// GenerateETag generates an ETag for the given content
func (chh *CacheHeadersHandler) GenerateETag(content []byte) string {
	return chh.middleware.generateETag(content)
}

// ValidateETag validates an ETag against content
func (chh *CacheHeadersHandler) ValidateETag(etag string, content []byte) bool {
	expectedETag := chh.GenerateETag(content)
	return etag == expectedETag
}

// ParseCacheControl parses a Cache-Control header
func (chh *CacheHeadersHandler) ParseCacheControl(cacheControl string) map[string]string {
	directives := make(map[string]string)

	parts := strings.Split(cacheControl, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			directives[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		} else {
			directives[part] = ""
		}
	}

	return directives
}

// GetMaxAge returns the max-age value from a Cache-Control header
func (chh *CacheHeadersHandler) GetMaxAge(cacheControl string) (int, error) {
	directives := chh.ParseCacheControl(cacheControl)

	if maxAge, exists := directives["max-age"]; exists {
		return strconv.Atoi(maxAge)
	}

	return 0, fmt.Errorf("max-age not found in Cache-Control header")
}
