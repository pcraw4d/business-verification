package middleware

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// RequestLoggingConfig holds configuration for request logging
type RequestLoggingConfig struct {
	// Logging Level
	LogLevel zapcore.Level `json:"log_level" yaml:"log_level"`

	// Request Body Logging
	LogRequestBody     bool  `json:"log_request_body" yaml:"log_request_body"`
	MaxRequestBodySize int64 `json:"max_request_body_size" yaml:"max_request_body_size"`

	// Response Body Logging
	LogResponseBody     bool  `json:"log_response_body" yaml:"log_response_body"`
	MaxResponseBodySize int64 `json:"max_response_body_size" yaml:"max_response_body_size"`

	// Performance Logging
	LogPerformance       bool          `json:"log_performance" yaml:"log_performance"`
	SlowRequestThreshold time.Duration `json:"slow_request_threshold" yaml:"slow_request_threshold"`

	// Request ID
	GenerateRequestID bool   `json:"generate_request_id" yaml:"generate_request_id"`
	RequestIDHeader   string `json:"request_id_header" yaml:"request_id_header"`

	// Sensitive Data Masking
	MaskSensitiveHeaders []string `json:"mask_sensitive_headers" yaml:"mask_sensitive_headers"`
	MaskSensitiveFields  []string `json:"mask_sensitive_fields" yaml:"mask_sensitive_fields"`

	// Path Filtering
	IncludePaths []string `json:"include_paths" yaml:"include_paths"`
	ExcludePaths []string `json:"exclude_paths" yaml:"exclude_paths"`

	// Custom Fields
	CustomFields map[string]string `json:"custom_fields" yaml:"custom_fields"`

	// Error Logging
	LogErrors bool `json:"log_errors" yaml:"log_errors"`
	LogPanics bool `json:"log_panics" yaml:"log_panics"`
}

// RequestLoggingMiddleware provides comprehensive request logging
type RequestLoggingMiddleware struct {
	config *RequestLoggingConfig
	logger *zap.Logger
}

// responseWriter wraps http.ResponseWriter to capture response data
type responseWriter struct {
	http.ResponseWriter
	statusCode    int
	body          *bytes.Buffer
	contentLength int64
	written       bool
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	if !rw.written {
		rw.statusCode = statusCode
		rw.ResponseWriter.WriteHeader(statusCode)
		rw.written = true
	}
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}

	// Capture response body if enabled
	if rw.body != nil && int64(rw.body.Len()+len(data)) <= rw.contentLength {
		rw.body.Write(data)
	}

	return rw.ResponseWriter.Write(data)
}

// NewRequestLoggingMiddleware creates a new RequestLoggingMiddleware
func NewRequestLoggingMiddleware(config *RequestLoggingConfig, logger *zap.Logger) *RequestLoggingMiddleware {
	if config == nil {
		config = &RequestLoggingConfig{
			LogLevel:             zapcore.InfoLevel,
			LogRequestBody:       false,
			MaxRequestBodySize:   1024, // 1KB
			LogResponseBody:      false,
			MaxResponseBodySize:  1024, // 1KB
			LogPerformance:       true,
			SlowRequestThreshold: 1 * time.Second,
			GenerateRequestID:    true,
			RequestIDHeader:      "X-Request-ID",
			MaskSensitiveHeaders: []string{"Authorization", "X-API-Key", "Cookie"},
			MaskSensitiveFields:  []string{"password", "token", "secret", "key"},
			IncludePaths:         []string{},
			ExcludePaths:         []string{"/health", "/metrics", "/favicon.ico"},
			CustomFields:         map[string]string{},
			LogErrors:            true,
			LogPanics:            true,
		}
	}

	return &RequestLoggingMiddleware{
		config: config,
		logger: logger,
	}
}

// Middleware applies request logging to HTTP requests
func (m *RequestLoggingMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Check if path should be logged
		if !m.shouldLogPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		// Generate request ID
		requestID := m.getOrGenerateRequestID(r)

		// Create context with request ID
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		r = r.WithContext(ctx)

		// Set request ID header in response
		if m.config.GenerateRequestID {
			w.Header().Set(m.config.RequestIDHeader, requestID)
		}

		// Capture request body if enabled
		var requestBody []byte
		if m.config.LogRequestBody && r.Body != nil {
			requestBody = m.captureRequestBody(r)
		}

		// Create response writer wrapper
		var responseBody *bytes.Buffer
		if m.config.LogResponseBody {
			responseBody = bytes.NewBuffer(nil)
		}

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
			body:           responseBody,
			contentLength:  m.config.MaxResponseBodySize,
		}

		// Handle panics
		if m.config.LogPanics {
			defer func() {
				if err := recover(); err != nil {
					m.logPanic(r, requestID, err, start)
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
		}

		// Process request
		next.ServeHTTP(rw, r)

		// Log request
		m.logRequest(r, rw, requestID, requestBody, responseBody.Bytes(), start)
	})
}

// shouldLogPath checks if the path should be logged based on include/exclude rules
func (m *RequestLoggingMiddleware) shouldLogPath(path string) bool {
	// Check exclude paths first
	for _, excludePath := range m.config.ExcludePaths {
		if strings.HasPrefix(path, excludePath) {
			return false
		}
	}

	// If include paths are specified, only log those paths
	if len(m.config.IncludePaths) > 0 {
		for _, includePath := range m.config.IncludePaths {
			if strings.HasPrefix(path, includePath) {
				return true
			}
		}
		return false
	}

	return true
}

// getOrGenerateRequestID gets existing request ID or generates a new one
func (m *RequestLoggingMiddleware) getOrGenerateRequestID(r *http.Request) string {
	if !m.config.GenerateRequestID {
		return ""
	}

	// Check for existing request ID
	if requestID := r.Header.Get(m.config.RequestIDHeader); requestID != "" {
		return requestID
	}

	// Generate new request ID
	return m.generateRequestID()
}

// generateRequestID generates a unique request ID
func (m *RequestLoggingMiddleware) generateRequestID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// captureRequestBody captures and returns the request body
func (m *RequestLoggingMiddleware) captureRequestBody(r *http.Request) []byte {
	if r.Body == nil {
		return nil
	}

	// Read body
	body, err := io.ReadAll(io.LimitReader(r.Body, m.config.MaxRequestBodySize))
	if err != nil {
		return nil
	}

	// Restore body for downstream handlers
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	return body
}

// logRequest logs the complete request information
func (m *RequestLoggingMiddleware) logRequest(r *http.Request, rw *responseWriter, requestID string, requestBody, responseBody []byte, start time.Time) {
	duration := time.Since(start)

	// Determine log level
	logLevel := m.config.LogLevel
	if m.config.LogErrors && rw.statusCode >= 400 {
		logLevel = zapcore.ErrorLevel
	}
	if m.config.LogPerformance && duration > m.config.SlowRequestThreshold {
		logLevel = zapcore.WarnLevel
	}

	// Build log fields
	fields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("query", r.URL.RawQuery),
		zap.String("remote_addr", m.getRemoteAddr(r)),
		zap.String("user_agent", r.UserAgent()),
		zap.Int("status_code", rw.statusCode),
		zap.Duration("duration", duration),
		zap.Int64("content_length", r.ContentLength),
	}

	// Add headers (masked if sensitive)
	headers := m.maskSensitiveHeaders(r.Header)
	fields = append(fields, zap.Any("headers", headers))

	// Add request body if enabled
	if m.config.LogRequestBody && len(requestBody) > 0 {
		maskedBody := m.maskSensitiveData(string(requestBody))
		fields = append(fields, zap.String("request_body", maskedBody))
	}

	// Add response body if enabled
	if m.config.LogResponseBody && len(responseBody) > 0 {
		maskedBody := m.maskSensitiveData(string(responseBody))
		fields = append(fields, zap.String("response_body", maskedBody))
	}

	// Add performance metrics
	if m.config.LogPerformance {
		fields = append(fields, zap.Float64("duration_ms", float64(duration.Microseconds())/1000))
	}

	// Add custom fields
	for key, value := range m.config.CustomFields {
		fields = append(fields, zap.String(key, value))
	}

	// Log based on level
	switch logLevel {
	case zapcore.DebugLevel:
		m.logger.Debug("HTTP request", fields...)
	case zapcore.InfoLevel:
		m.logger.Info("HTTP request", fields...)
	case zapcore.WarnLevel:
		m.logger.Warn("HTTP request", fields...)
	case zapcore.ErrorLevel:
		m.logger.Error("HTTP request", fields...)
	}
}

// logPanic logs panic information
func (m *RequestLoggingMiddleware) logPanic(r *http.Request, requestID string, err interface{}, start time.Time) {
	duration := time.Since(start)

	fields := []zap.Field{
		zap.String("request_id", requestID),
		zap.String("method", r.Method),
		zap.String("path", r.URL.Path),
		zap.String("remote_addr", m.getRemoteAddr(r)),
		zap.Any("panic", err),
		zap.Duration("duration", duration),
	}

	m.logger.Error("HTTP request panic", fields...)
}

// getRemoteAddr gets the real remote address considering proxies
func (m *RequestLoggingMiddleware) getRemoteAddr(r *http.Request) string {
	// Check for forwarded headers
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Take the first IP in the list
		if commaIndex := strings.Index(forwardedFor, ","); commaIndex != -1 {
			return strings.TrimSpace(forwardedFor[:commaIndex])
		}
		return strings.TrimSpace(forwardedFor)
	}

	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}

// maskSensitiveHeaders masks sensitive header values
func (m *RequestLoggingMiddleware) maskSensitiveHeaders(headers http.Header) map[string]string {
	masked := make(map[string]string)

	for key, values := range headers {
		lowerKey := strings.ToLower(key)
		isSensitive := false

		for _, sensitiveHeader := range m.config.MaskSensitiveHeaders {
			if strings.ToLower(sensitiveHeader) == lowerKey {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			masked[key] = "[MASKED]"
		} else {
			masked[key] = strings.Join(values, ", ")
		}
	}

	return masked
}

// maskSensitiveData masks sensitive data in request/response bodies
func (m *RequestLoggingMiddleware) maskSensitiveData(data string) string {
	masked := data

	for _, field := range m.config.MaskSensitiveFields {
		// Simple pattern matching for JSON-like data
		// This is a simplified implementation
		// In production, you might want to use proper JSON parsing
		if strings.Contains(strings.ToLower(masked), strings.ToLower(field)) {
			masked = strings.ReplaceAll(masked, field, "[MASKED]")
		}
	}

	return masked
}

// GetDefaultRequestLoggingConfig returns a default request logging configuration
func GetDefaultRequestLoggingConfig() *RequestLoggingConfig {
	return &RequestLoggingConfig{
		LogLevel:             zapcore.InfoLevel,
		LogRequestBody:       false,
		MaxRequestBodySize:   1024, // 1KB
		LogResponseBody:      false,
		MaxResponseBodySize:  1024, // 1KB
		LogPerformance:       true,
		SlowRequestThreshold: 1 * time.Second,
		GenerateRequestID:    true,
		RequestIDHeader:      "X-Request-ID",
		MaskSensitiveHeaders: []string{"Authorization", "X-API-Key", "Cookie"},
		MaskSensitiveFields:  []string{"password", "token", "secret", "key"},
		IncludePaths:         []string{},
		ExcludePaths:         []string{"/health", "/metrics", "/favicon.ico"},
		CustomFields:         map[string]string{},
		LogErrors:            true,
		LogPanics:            true,
	}
}

// GetVerboseRequestLoggingConfig returns a verbose request logging configuration
func GetVerboseRequestLoggingConfig() *RequestLoggingConfig {
	return &RequestLoggingConfig{
		LogLevel:             zapcore.DebugLevel,
		LogRequestBody:       true,
		MaxRequestBodySize:   4096, // 4KB
		LogResponseBody:      true,
		MaxResponseBodySize:  4096, // 4KB
		LogPerformance:       true,
		SlowRequestThreshold: 500 * time.Millisecond,
		GenerateRequestID:    true,
		RequestIDHeader:      "X-Request-ID",
		MaskSensitiveHeaders: []string{"Authorization", "X-API-Key", "Cookie"},
		MaskSensitiveFields:  []string{"password", "token", "secret", "key"},
		IncludePaths:         []string{},
		ExcludePaths:         []string{"/health", "/metrics", "/favicon.ico"},
		CustomFields:         map[string]string{},
		LogErrors:            true,
		LogPanics:            true,
	}
}

// GetProductionRequestLoggingConfig returns a production request logging configuration
func GetProductionRequestLoggingConfig() *RequestLoggingConfig {
	return &RequestLoggingConfig{
		LogLevel:             zapcore.InfoLevel,
		LogRequestBody:       false,
		MaxRequestBodySize:   0,
		LogResponseBody:      false,
		MaxResponseBodySize:  0,
		LogPerformance:       true,
		SlowRequestThreshold: 2 * time.Second,
		GenerateRequestID:    true,
		RequestIDHeader:      "X-Request-ID",
		MaskSensitiveHeaders: []string{"Authorization", "X-API-Key", "Cookie", "X-CSRF-Token"},
		MaskSensitiveFields:  []string{"password", "token", "secret", "key", "api_key"},
		IncludePaths:         []string{},
		ExcludePaths:         []string{"/health", "/metrics", "/favicon.ico", "/robots.txt"},
		CustomFields:         map[string]string{},
		LogErrors:            true,
		LogPanics:            true,
	}
}
