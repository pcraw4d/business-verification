package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewRequestLoggingMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		config   *RequestLoggingConfig
		expected *RequestLoggingConfig
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			expected: &RequestLoggingConfig{
				LogLevel:             zapcore.InfoLevel,
				LogRequestBody:       false,
				MaxRequestBodySize:   1024,
				LogResponseBody:      false,
				MaxResponseBodySize:  1024,
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
			},
		},
		{
			name: "custom config",
			config: &RequestLoggingConfig{
				LogLevel:             zapcore.DebugLevel,
				LogRequestBody:       true,
				MaxRequestBodySize:   2048,
				LogResponseBody:      true,
				MaxResponseBodySize:  2048,
				LogPerformance:       false,
				SlowRequestThreshold: 2 * time.Second,
				GenerateRequestID:    false,
				RequestIDHeader:      "X-Custom-ID",
				MaskSensitiveHeaders: []string{"Authorization"},
				MaskSensitiveFields:  []string{"password"},
				IncludePaths:         []string{"/api"},
				ExcludePaths:         []string{"/health"},
				CustomFields:         map[string]string{"service": "test"},
				LogErrors:            false,
				LogPanics:            false,
			},
			expected: &RequestLoggingConfig{
				LogLevel:             zapcore.DebugLevel,
				LogRequestBody:       true,
				MaxRequestBodySize:   2048,
				LogResponseBody:      true,
				MaxResponseBodySize:  2048,
				LogPerformance:       false,
				SlowRequestThreshold: 2 * time.Second,
				GenerateRequestID:    false,
				RequestIDHeader:      "X-Custom-ID",
				MaskSensitiveHeaders: []string{"Authorization"},
				MaskSensitiveFields:  []string{"password"},
				IncludePaths:         []string{"/api"},
				ExcludePaths:         []string{"/health"},
				CustomFields:         map[string]string{"service": "test"},
				LogErrors:            false,
				LogPanics:            false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zap.NewNop()
			middleware := NewRequestLoggingMiddleware(tt.config, logger)

			assert.NotNil(t, middleware)
			assert.Equal(t, tt.expected.LogLevel, middleware.config.LogLevel)
			assert.Equal(t, tt.expected.LogRequestBody, middleware.config.LogRequestBody)
			assert.Equal(t, tt.expected.MaxRequestBodySize, middleware.config.MaxRequestBodySize)
			assert.Equal(t, tt.expected.LogResponseBody, middleware.config.LogResponseBody)
			assert.Equal(t, tt.expected.MaxResponseBodySize, middleware.config.MaxResponseBodySize)
			assert.Equal(t, tt.expected.LogPerformance, middleware.config.LogPerformance)
			assert.Equal(t, tt.expected.SlowRequestThreshold, middleware.config.SlowRequestThreshold)
			assert.Equal(t, tt.expected.GenerateRequestID, middleware.config.GenerateRequestID)
			assert.Equal(t, tt.expected.RequestIDHeader, middleware.config.RequestIDHeader)
			assert.Equal(t, tt.expected.MaskSensitiveHeaders, middleware.config.MaskSensitiveHeaders)
			assert.Equal(t, tt.expected.MaskSensitiveFields, middleware.config.MaskSensitiveFields)
			assert.Equal(t, tt.expected.IncludePaths, middleware.config.IncludePaths)
			assert.Equal(t, tt.expected.ExcludePaths, middleware.config.ExcludePaths)
			assert.Equal(t, tt.expected.CustomFields, middleware.config.CustomFields)
			assert.Equal(t, tt.expected.LogErrors, middleware.config.LogErrors)
			assert.Equal(t, tt.expected.LogPanics, middleware.config.LogPanics)
		})
	}
}

func TestRequestLoggingMiddleware_Middleware(t *testing.T) {
	// Create observer to capture logs
	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := &RequestLoggingConfig{
		LogLevel:             zapcore.InfoLevel,
		LogRequestBody:       false,
		MaxRequestBodySize:   1024,
		LogResponseBody:      false,
		MaxResponseBodySize:  1024,
		LogPerformance:       true,
		SlowRequestThreshold: 1 * time.Second,
		GenerateRequestID:    true,
		RequestIDHeader:      "X-Request-ID",
		MaskSensitiveHeaders: []string{"Authorization", "X-API-Key"},
		MaskSensitiveFields:  []string{"password", "token"},
		IncludePaths:         []string{},
		ExcludePaths:         []string{"/health", "/metrics"},
		CustomFields:         map[string]string{},
		LogErrors:            true,
		LogPanics:            true,
	}

	middleware := NewRequestLoggingMiddleware(config, logger)

	tests := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		body           string
		expectedStatus int
		expectedLogs   int
		shouldLog      bool
	}{
		{
			name:           "normal request",
			method:         "GET",
			path:           "/api/test",
			headers:        map[string]string{"Content-Type": "application/json"},
			body:           "",
			expectedStatus: http.StatusOK,
			expectedLogs:   1,
			shouldLog:      true,
		},
		{
			name:           "excluded path",
			method:         "GET",
			path:           "/health",
			headers:        map[string]string{},
			body:           "",
			expectedStatus: http.StatusOK,
			expectedLogs:   0,
			shouldLog:      false,
		},
		{
			name:           "error request",
			method:         "POST",
			path:           "/api/error",
			headers:        map[string]string{"Content-Type": "application/json"},
			body:           `{"error": "test error"}`,
			expectedStatus: http.StatusBadRequest,
			expectedLogs:   1,
			shouldLog:      true,
		},
		{
			name:           "request with sensitive headers",
			method:         "GET",
			path:           "/api/secure",
			headers:        map[string]string{"Authorization": "Bearer secret-token", "X-API-Key": "secret-key"},
			body:           "",
			expectedStatus: http.StatusOK,
			expectedLogs:   1,
			shouldLog:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear previous logs
			obs.FilterMessage("HTTP request").Len()

			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.expectedStatus != http.StatusOK {
					http.Error(w, "Bad Request", tt.expectedStatus)
				} else {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("success"))
				}
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			// Check request ID header
			if tt.shouldLog && config.GenerateRequestID {
				requestID := recorder.Header().Get(config.RequestIDHeader)
				assert.NotEmpty(t, requestID)
				assert.Len(t, requestID, 32) // 16 bytes = 32 hex chars
			}

			// Check logs
			logs := obs.FilterMessage("HTTP request").All()
			assert.Len(t, logs, tt.expectedLogs)

			if tt.expectedLogs > 0 {
				log := logs[0]
				assert.Equal(t, tt.method, log.ContextMap()["method"])
				assert.Equal(t, tt.path, log.ContextMap()["path"])
				assert.Equal(t, tt.expectedStatus, log.ContextMap()["status_code"])
				assert.Contains(t, log.ContextMap(), "request_id")
				assert.Contains(t, log.ContextMap(), "duration")
				assert.Contains(t, log.ContextMap(), "remote_addr")
				assert.Contains(t, log.ContextMap(), "user_agent")
			}
		})
	}
}

func TestRequestLoggingMiddleware_RequestIDGeneration(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := &RequestLoggingConfig{
		GenerateRequestID: true,
		RequestIDHeader:   "X-Request-ID",
	}
	middleware := NewRequestLoggingMiddleware(config, logger)

	tests := []struct {
		name           string
		existingID     string
		shouldGenerate bool
	}{
		{
			name:           "generate new request ID",
			existingID:     "",
			shouldGenerate: true,
		},
		{
			name:           "use existing request ID",
			existingID:     "existing-request-id",
			shouldGenerate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs.FilterMessage("HTTP request").Len()

			req := httptest.NewRequest("GET", "/api/test", nil)
			if tt.existingID != "" {
				req.Header.Set(config.RequestIDHeader, tt.existingID)
			}

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			requestID := recorder.Header().Get(config.RequestIDHeader)
			assert.NotEmpty(t, requestID)

			if tt.shouldGenerate {
				assert.Len(t, requestID, 32) // 16 bytes = 32 hex chars
			} else {
				assert.Equal(t, tt.existingID, requestID)
			}
		})
	}
}

func TestRequestLoggingMiddleware_BodyCapture(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := &RequestLoggingConfig{
		LogRequestBody:      true,
		MaxRequestBodySize:  1024,
		LogResponseBody:     true,
		MaxResponseBodySize: 1024,
		GenerateRequestID:   true,
		RequestIDHeader:     "X-Request-ID",
		MaskSensitiveFields: []string{"password", "token"},
	}
	middleware := NewRequestLoggingMiddleware(config, logger)

	tests := []struct {
		name           string
		requestBody    string
		responseBody   string
		expectedStatus int
		shouldLogBody  bool
	}{
		{
			name:           "request with body",
			requestBody:    `{"name": "test", "password": "secret123"}`,
			responseBody:   `{"status": "success"}`,
			expectedStatus: http.StatusOK,
			shouldLogBody:  true,
		},
		{
			name:           "request without body",
			requestBody:    "",
			responseBody:   "",
			expectedStatus: http.StatusOK,
			shouldLogBody:  false,
		},
		{
			name:           "large request body",
			requestBody:    strings.Repeat("a", 2048), // Larger than max size
			responseBody:   `{"status": "success"}`,
			expectedStatus: http.StatusOK,
			shouldLogBody:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs.FilterMessage("HTTP request").Len()

			req := httptest.NewRequest("POST", "/api/test", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.expectedStatus)
				if tt.responseBody != "" {
					w.Write([]byte(tt.responseBody))
				}
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			logs := obs.FilterMessage("HTTP request").All()
			assert.Len(t, logs, 1)

			log := logs[0]
			logFields := log.ContextMap()

			if tt.shouldLogBody && tt.requestBody != "" {
				assert.Contains(t, logFields, "request_body")
				requestBody := logFields["request_body"].(string)
				// Check that sensitive data is masked
				assert.Contains(t, requestBody, "[MASKED]")
				assert.NotContains(t, requestBody, "secret123")
			}

			if tt.shouldLogBody && tt.responseBody != "" {
				assert.Contains(t, logFields, "response_body")
			}
		})
	}
}

func TestRequestLoggingMiddleware_PathFiltering(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := &RequestLoggingConfig{
		IncludePaths:      []string{"/api", "/admin"},
		ExcludePaths:      []string{"/api/health", "/admin/metrics"},
		GenerateRequestID: true,
		RequestIDHeader:   "X-Request-ID",
	}
	middleware := NewRequestLoggingMiddleware(config, logger)

	tests := []struct {
		name         string
		path         string
		shouldLog    bool
		expectedLogs int
	}{
		{
			name:         "included path",
			path:         "/api/users",
			shouldLog:    true,
			expectedLogs: 1,
		},
		{
			name:         "excluded path",
			path:         "/api/health",
			shouldLog:    false,
			expectedLogs: 0,
		},
		{
			name:         "another included path",
			path:         "/admin/dashboard",
			shouldLog:    true,
			expectedLogs: 1,
		},
		{
			name:         "another excluded path",
			path:         "/admin/metrics",
			shouldLog:    false,
			expectedLogs: 0,
		},
		{
			name:         "not in include paths",
			path:         "/other/path",
			shouldLog:    false,
			expectedLogs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs.FilterMessage("HTTP request").Len()

			req := httptest.NewRequest("GET", tt.path, nil)
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			logs := obs.FilterMessage("HTTP request").All()
			assert.Len(t, logs, tt.expectedLogs)

			if tt.shouldLog {
				requestID := recorder.Header().Get(config.RequestIDHeader)
				assert.NotEmpty(t, requestID)
			} else {
				requestID := recorder.Header().Get(config.RequestIDHeader)
				assert.Empty(t, requestID)
			}
		})
	}
}

func TestRequestLoggingMiddleware_SensitiveDataMasking(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := &RequestLoggingConfig{
		LogRequestBody:       true,
		MaxRequestBodySize:   1024,
		LogResponseBody:      true,
		MaxResponseBodySize:  1024,
		GenerateRequestID:    true,
		RequestIDHeader:      "X-Request-ID",
		MaskSensitiveHeaders: []string{"Authorization", "X-API-Key", "Cookie"},
		MaskSensitiveFields:  []string{"password", "token", "secret"},
	}
	middleware := NewRequestLoggingMiddleware(config, logger)

	tests := []struct {
		name           string
		headers        map[string]string
		requestBody    string
		responseBody   string
		expectedStatus int
	}{
		{
			name: "sensitive headers and body",
			headers: map[string]string{
				"Authorization": "Bearer secret-token",
				"X-API-Key":     "secret-api-key",
				"Cookie":        "session=secret-session",
				"Content-Type":  "application/json",
			},
			requestBody:    `{"username": "test", "password": "secret123", "token": "secret-token"}`,
			responseBody:   `{"status": "success", "secret": "hidden-data"}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs.FilterMessage("HTTP request").Len()

			req := httptest.NewRequest("POST", "/api/test", strings.NewReader(tt.requestBody))
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.expectedStatus)
				w.Write([]byte(tt.responseBody))
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			logs := obs.FilterMessage("HTTP request").All()
			assert.Len(t, logs, 1)

			log := logs[0]
			logFields := log.ContextMap()

			// Check headers are masked
			headers := logFields["headers"].(map[string]interface{})
			assert.Equal(t, "[MASKED]", headers["Authorization"])
			assert.Equal(t, "[MASKED]", headers["X-API-Key"])
			assert.Equal(t, "[MASKED]", headers["Cookie"])
			assert.Equal(t, "application/json", headers["Content-Type"])

			// Check request body is masked
			requestBody := logFields["request_body"].(string)
			assert.Contains(t, requestBody, "[MASKED]")
			assert.NotContains(t, requestBody, "secret123")
			assert.NotContains(t, requestBody, "secret-token")

			// Check response body is masked
			responseBody := logFields["response_body"].(string)
			assert.Contains(t, responseBody, "[MASKED]")
			assert.NotContains(t, responseBody, "hidden-data")
		})
	}
}

func TestRequestLoggingMiddleware_PerformanceLogging(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := &RequestLoggingConfig{
		LogPerformance:       true,
		SlowRequestThreshold: 100 * time.Millisecond,
		GenerateRequestID:    true,
		RequestIDHeader:      "X-Request-ID",
	}
	middleware := NewRequestLoggingMiddleware(config, logger)

	tests := []struct {
		name           string
		delay          time.Duration
		expectedLevel  zapcore.Level
		expectedStatus int
	}{
		{
			name:           "fast request",
			delay:          50 * time.Millisecond,
			expectedLevel:  zapcore.InfoLevel,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "slow request",
			delay:          200 * time.Millisecond,
			expectedLevel:  zapcore.WarnLevel,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "error request",
			delay:          50 * time.Millisecond,
			expectedLevel:  zapcore.ErrorLevel,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs.FilterMessage("HTTP request").Len()

			req := httptest.NewRequest("GET", "/api/test", nil)
			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(tt.delay)
				w.WriteHeader(tt.expectedStatus)
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			logs := obs.FilterMessage("HTTP request").All()
			assert.Len(t, logs, 1)

			log := logs[0]
			assert.Equal(t, tt.expectedLevel, log.Level)

			logFields := log.ContextMap()
			assert.Contains(t, logFields, "duration")
			assert.Contains(t, logFields, "duration_ms")
			assert.Equal(t, tt.expectedStatus, logFields["status_code"])
		})
	}
}

func TestRequestLoggingMiddleware_RemoteAddr(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := &RequestLoggingConfig{
		GenerateRequestID: true,
		RequestIDHeader:   "X-Request-ID",
	}
	middleware := NewRequestLoggingMiddleware(config, logger)

	tests := []struct {
		name         string
		headers      map[string]string
		remoteAddr   string
		expectedAddr string
	}{
		{
			name: "X-Forwarded-For header",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1",
			},
			remoteAddr:   "127.0.0.1:8080",
			expectedAddr: "192.168.1.1",
		},
		{
			name: "X-Real-IP header",
			headers: map[string]string{
				"X-Real-IP": "203.0.113.1",
			},
			remoteAddr:   "127.0.0.1:8080",
			expectedAddr: "203.0.113.1",
		},
		{
			name:         "no proxy headers",
			headers:      map[string]string{},
			remoteAddr:   "127.0.0.1:8080",
			expectedAddr: "127.0.0.1:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obs.FilterMessage("HTTP request").Len()

			req := httptest.NewRequest("GET", "/api/test", nil)
			req.RemoteAddr = tt.remoteAddr
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			recorder := httptest.NewRecorder()
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware.Middleware(handler).ServeHTTP(recorder, req)

			logs := obs.FilterMessage("HTTP request").All()
			assert.Len(t, logs, 1)

			log := logs[0]
			logFields := log.ContextMap()
			assert.Equal(t, tt.expectedAddr, logFields["remote_addr"])
		})
	}
}

func TestRequestLoggingMiddleware_PanicHandling(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := &RequestLoggingConfig{
		LogPanics:         true,
		GenerateRequestID: true,
		RequestIDHeader:   "X-Request-ID",
	}
	middleware := NewRequestLoggingMiddleware(config, logger)

	req := httptest.NewRequest("GET", "/api/panic", nil)
	recorder := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware.Middleware(handler).ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)

	// Check panic log
	panicLogs := obs.FilterMessage("HTTP request panic").All()
	assert.Len(t, panicLogs, 1)

	panicLog := panicLogs[0]
	logFields := panicLog.ContextMap()
	assert.Equal(t, "test panic", logFields["panic"])
	assert.Equal(t, "GET", logFields["method"])
	assert.Equal(t, "/api/panic", logFields["path"])
	assert.Contains(t, logFields, "request_id")
	assert.Contains(t, logFields, "duration")
}

func TestGetDefaultRequestLoggingConfig(t *testing.T) {
	config := GetDefaultRequestLoggingConfig()

	assert.NotNil(t, config)
	assert.Equal(t, zapcore.InfoLevel, config.LogLevel)
	assert.False(t, config.LogRequestBody)
	assert.Equal(t, int64(1024), config.MaxRequestBodySize)
	assert.False(t, config.LogResponseBody)
	assert.Equal(t, int64(1024), config.MaxResponseBodySize)
	assert.True(t, config.LogPerformance)
	assert.Equal(t, 1*time.Second, config.SlowRequestThreshold)
	assert.True(t, config.GenerateRequestID)
	assert.Equal(t, "X-Request-ID", config.RequestIDHeader)
	assert.True(t, config.LogErrors)
	assert.True(t, config.LogPanics)
}

func TestGetVerboseRequestLoggingConfig(t *testing.T) {
	config := GetVerboseRequestLoggingConfig()

	assert.NotNil(t, config)
	assert.Equal(t, zapcore.DebugLevel, config.LogLevel)
	assert.True(t, config.LogRequestBody)
	assert.Equal(t, int64(4096), config.MaxRequestBodySize)
	assert.True(t, config.LogResponseBody)
	assert.Equal(t, int64(4096), config.MaxResponseBodySize)
	assert.True(t, config.LogPerformance)
	assert.Equal(t, 500*time.Millisecond, config.SlowRequestThreshold)
	assert.True(t, config.GenerateRequestID)
	assert.Equal(t, "X-Request-ID", config.RequestIDHeader)
	assert.True(t, config.LogErrors)
	assert.True(t, config.LogPanics)
}

func TestGetProductionRequestLoggingConfig(t *testing.T) {
	config := GetProductionRequestLoggingConfig()

	assert.NotNil(t, config)
	assert.Equal(t, zapcore.InfoLevel, config.LogLevel)
	assert.False(t, config.LogRequestBody)
	assert.Equal(t, int64(0), config.MaxRequestBodySize)
	assert.False(t, config.LogResponseBody)
	assert.Equal(t, int64(0), config.MaxResponseBodySize)
	assert.True(t, config.LogPerformance)
	assert.Equal(t, 2*time.Second, config.SlowRequestThreshold)
	assert.True(t, config.GenerateRequestID)
	assert.Equal(t, "X-Request-ID", config.RequestIDHeader)
	assert.True(t, config.LogErrors)
	assert.True(t, config.LogPanics)
}
