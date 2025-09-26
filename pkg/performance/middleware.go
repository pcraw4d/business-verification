package performance

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"time"
)

// CompressionMiddleware provides gzip compression for responses
func CompressionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if client accepts gzip encoding
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		// Create gzip writer
		gz := gzip.NewWriter(w)
		defer gz.Close()

		// Set headers
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Vary", "Accept-Encoding")

		// Wrap response writer
		gzw := &gzipResponseWriter{ResponseWriter: w, Writer: gz}
		next.ServeHTTP(gzw, r)
	})
}

// gzipResponseWriter wraps http.ResponseWriter with gzip compression
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

func (gzw *gzipResponseWriter) Write(data []byte) (int, error) {
	return gzw.Writer.Write(data)
}

// CacheHeadersMiddleware adds caching headers to responses
func CacheHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add cache headers based on request type
		if strings.HasPrefix(r.URL.Path, "/api/") {
			// API responses - short cache
			w.Header().Set("Cache-Control", "public, max-age=300") // 5 minutes
		} else if strings.HasPrefix(r.URL.Path, "/static/") || strings.HasPrefix(r.URL.Path, "/js/") || strings.HasPrefix(r.URL.Path, "/css/") {
			// Static assets - long cache
			w.Header().Set("Cache-Control", "public, max-age=31536000") // 1 year
		} else {
			// Other responses - no cache
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		}

		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		next.ServeHTTP(w, r)
	})
}

// TimingMiddleware measures request processing time
func TimingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &timingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		// Log timing information
		duration := time.Since(start)
		w.Header().Set("X-Response-Time", duration.String())

		// Log slow requests (>1 second)
		if duration > time.Second {
			// This would typically go to a structured logger
			// For now, we'll just set a header
			w.Header().Set("X-Slow-Request", "true")
		}
	})
}

// timingResponseWriter wraps http.ResponseWriter to capture status code
type timingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (trw *timingResponseWriter) WriteHeader(code int) {
	trw.statusCode = code
	trw.ResponseWriter.WriteHeader(code)
}

// ConnectionPoolingMiddleware manages connection pooling
func ConnectionPoolingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set connection headers for keep-alive
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Keep-Alive", "timeout=5, max=1000")

		next.ServeHTTP(w, r)
	})
}
