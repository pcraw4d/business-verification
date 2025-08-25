package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewErrorHandlingMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		config   *ErrorHandlingConfig
		logger   *zap.Logger
		expected *ErrorHandlingMiddleware
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			logger: zap.NewNop(),
			expected: &ErrorHandlingMiddleware{
				config: &ErrorHandlingConfig{
					LogErrors:      true,
					LogLevel:       zapcore.ErrorLevel,
					IncludeStack:   false,
					IncludeDetails: true,
					IncludeContext: false,
					MaskInternal:   true,
					RecoverPanics:  true,
					CustomHandlers: make(map[ErrorType]ErrorHandler),
					TrackMetrics:   true,
					ErrorCodes: map[string]string{
						"validation_error":     "INVALID_INPUT",
						"authentication_error": "UNAUTHORIZED",
						"authorization_error":  "FORBIDDEN",
						"not_found_error":      "NOT_FOUND",
						"conflict_error":       "CONFLICT",
						"rate_limit_error":     "RATE_LIMITED",
						"internal_error":       "INTERNAL_ERROR",
						"external_error":       "EXTERNAL_ERROR",
						"timeout_error":        "TIMEOUT",
						"unavailable_error":    "SERVICE_UNAVAILABLE",
					},
					DefaultErrorType:     ErrorTypeInternal,
					DefaultErrorSeverity: ErrorSeverityMedium,
					DefaultStatusCode:    500,
				},
				metrics: &ErrorMetrics{
					ErrorsByType:     make(map[ErrorType]int64),
					ErrorsBySeverity: make(map[ErrorSeverity]int64),
				},
			},
		},
		{
			name: "custom config",
			config: &ErrorHandlingConfig{
				LogErrors:            false,
				LogLevel:             zapcore.DebugLevel,
				IncludeStack:         true,
				IncludeDetails:       false,
				IncludeContext:       true,
				MaskInternal:         false,
				RecoverPanics:        false,
				CustomHandlers:       make(map[ErrorType]ErrorHandler),
				TrackMetrics:         false,
				ErrorCodes:           map[string]string{"test": "TEST_ERROR"},
				DefaultErrorType:     ErrorTypeValidation,
				DefaultErrorSeverity: ErrorSeverityLow,
				DefaultStatusCode:    400,
			},
			logger: zap.NewNop(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := NewErrorHandlingMiddleware(tt.config, tt.logger)

			assert.NotNil(t, middleware)
			assert.Equal(t, tt.logger, middleware.logger)
			assert.NotNil(t, middleware.metrics)

			if tt.config == nil {
				// Check default config
				assert.Equal(t, tt.expected.config.LogErrors, middleware.config.LogErrors)
				assert.Equal(t, tt.expected.config.LogLevel, middleware.config.LogLevel)
				assert.Equal(t, tt.expected.config.IncludeDetails, middleware.config.IncludeDetails)
				assert.Equal(t, tt.expected.config.MaskInternal, middleware.config.MaskInternal)
				assert.Equal(t, tt.expected.config.RecoverPanics, middleware.config.RecoverPanics)
				assert.Equal(t, tt.expected.config.TrackMetrics, middleware.config.TrackMetrics)
				assert.Equal(t, tt.expected.config.DefaultErrorType, middleware.config.DefaultErrorType)
				assert.Equal(t, tt.expected.config.DefaultErrorSeverity, middleware.config.DefaultErrorSeverity)
				assert.Equal(t, tt.expected.config.DefaultStatusCode, middleware.config.DefaultStatusCode)

				// Check default error codes
				assert.Equal(t, "INVALID_INPUT", middleware.config.ErrorCodes["validation_error"])
				assert.Equal(t, "UNAUTHORIZED", middleware.config.ErrorCodes["authentication_error"])
				assert.Equal(t, "INTERNAL_ERROR", middleware.config.ErrorCodes["internal_error"])
			} else {
				// Check custom config
				assert.Equal(t, tt.config.LogErrors, middleware.config.LogErrors)
				assert.Equal(t, tt.config.LogLevel, middleware.config.LogLevel)
				assert.Equal(t, tt.config.IncludeStack, middleware.config.IncludeStack)
				assert.Equal(t, tt.config.IncludeDetails, middleware.config.IncludeDetails)
				assert.Equal(t, tt.config.IncludeContext, middleware.config.IncludeContext)
				assert.Equal(t, tt.config.MaskInternal, middleware.config.MaskInternal)
				assert.Equal(t, tt.config.RecoverPanics, middleware.config.RecoverPanics)
				assert.Equal(t, tt.config.TrackMetrics, middleware.config.TrackMetrics)
				assert.Equal(t, tt.config.DefaultErrorType, middleware.config.DefaultErrorType)
				assert.Equal(t, tt.config.DefaultErrorSeverity, middleware.config.DefaultErrorSeverity)
				assert.Equal(t, tt.config.DefaultStatusCode, middleware.config.DefaultStatusCode)
			}
		})
	}
}

func TestErrorHandlingMiddleware_Middleware(t *testing.T) {
	tests := []struct {
		name           string
		config         *ErrorHandlingConfig
		handler        http.HandlerFunc
		expectedStatus int
		expectedError  bool
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:   "successful request",
			config: GetDefaultErrorHandlingConfig(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"success": true}`))
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "400 error",
			config: GetDefaultErrorHandlingConfig(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Equal(t, ErrorTypeInternal, response.Error.Type)
				assert.Equal(t, "Invalid request", response.Error.Message)
				assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
				assert.Equal(t, "validation_error", rr.Header().Get("X-Error-Type"))
				assert.Equal(t, "INVALID_INPUT", rr.Header().Get("X-Error-Code"))
			},
		},
		{
			name:   "404 error",
			config: GetDefaultErrorHandlingConfig(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  true,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Equal(t, "Resource not found", response.Error.Message)
			},
		},
		{
			name:   "500 error",
			config: GetDefaultErrorHandlingConfig(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
			checkResponse: func(t *testing.T, rr *httptest.ResponseRecorder) {
				var response ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Equal(t, "Internal server error", response.Error.Message)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, obs := observer.New(zapcore.InfoLevel)
			logger := zap.New(core)

			middleware := NewErrorHandlingMiddleware(tt.config, logger)
			handler := middleware.Middleware(tt.handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, rr)
			}

			if tt.expectedError {
				assert.Greater(t, obs.Len(), 0)
			}
		})
	}
}

func TestErrorHandlingMiddleware_PanicRecovery(t *testing.T) {
	tests := []struct {
		name           string
		recoverPanics  bool
		handler        http.HandlerFunc
		expectedStatus int
		expectedError  bool
	}{
		{
			name:          "panic recovery enabled",
			recoverPanics: true,
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:          "panic recovery disabled",
			recoverPanics: false,
			handler: func(w http.ResponseWriter, r *http.Request) {
				panic("test panic")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			core, obs := observer.New(zapcore.InfoLevel)
			logger := zap.New(core)

			config := GetDefaultErrorHandlingConfig()
			config.RecoverPanics = tt.recoverPanics

			middleware := NewErrorHandlingMiddleware(config, logger)
			handler := middleware.Middleware(tt.handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rr := httptest.NewRecorder()

			// This should not panic due to recovery
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError {
				var response ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)

				assert.False(t, response.Success)
				assert.Equal(t, ErrorTypeInternal, response.Error.Type)
				assert.Equal(t, ErrorSeverityCritical, response.Error.Severity)
				assert.Equal(t, "Internal server error", response.Error.Message)
				assert.Contains(t, response.Error.Details, "Panic: test panic")

				// Check that panic was logged
				assert.Greater(t, obs.Len(), 0)
			}
		})
	}
}

func TestCustomErrorCreation(t *testing.T) {
	tests := []struct {
		name             string
		createFunc       func(string, string) *CustomError
		expectedType     ErrorType
		expectedSeverity ErrorSeverity
		expectedCode     string
		expectedStatus   int
	}{
		{
			name:             "validation error",
			createFunc:       CreateValidationError,
			expectedType:     ErrorTypeValidation,
			expectedSeverity: ErrorSeverityLow,
			expectedCode:     "INVALID_INPUT",
			expectedStatus:   http.StatusBadRequest,
		},
		{
			name:             "authentication error",
			createFunc:       CreateAuthenticationError,
			expectedType:     ErrorTypeAuthentication,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "UNAUTHORIZED",
			expectedStatus:   http.StatusUnauthorized,
		},
		{
			name:             "authorization error",
			createFunc:       CreateAuthorizationError,
			expectedType:     ErrorTypeAuthorization,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "FORBIDDEN",
			expectedStatus:   http.StatusForbidden,
		},
		{
			name:             "not found error",
			createFunc:       CreateNotFoundError,
			expectedType:     ErrorTypeNotFound,
			expectedSeverity: ErrorSeverityLow,
			expectedCode:     "NOT_FOUND",
			expectedStatus:   http.StatusNotFound,
		},
		{
			name:             "conflict error",
			createFunc:       CreateConflictError,
			expectedType:     ErrorTypeConflict,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "CONFLICT",
			expectedStatus:   http.StatusConflict,
		},
		{
			name:             "rate limit error",
			createFunc:       CreateRateLimitError,
			expectedType:     ErrorTypeRateLimit,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "RATE_LIMITED",
			expectedStatus:   http.StatusTooManyRequests,
		},
		{
			name:             "internal error",
			createFunc:       CreateInternalError,
			expectedType:     ErrorTypeInternal,
			expectedSeverity: ErrorSeverityHigh,
			expectedCode:     "INTERNAL_ERROR",
			expectedStatus:   http.StatusInternalServerError,
		},
		{
			name:             "external error",
			createFunc:       CreateExternalError,
			expectedType:     ErrorTypeExternal,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "EXTERNAL_ERROR",
			expectedStatus:   http.StatusBadGateway,
		},
		{
			name:             "timeout error",
			createFunc:       CreateTimeoutError,
			expectedType:     ErrorTypeTimeout,
			expectedSeverity: ErrorSeverityMedium,
			expectedCode:     "TIMEOUT",
			expectedStatus:   http.StatusRequestTimeout,
		},
		{
			name:             "unavailable error",
			createFunc:       CreateUnavailableError,
			expectedType:     ErrorTypeUnavailable,
			expectedSeverity: ErrorSeverityHigh,
			expectedCode:     "SERVICE_UNAVAILABLE",
			expectedStatus:   http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := "test message"
			details := "test details"

			err := tt.createFunc(message, details)

			assert.Equal(t, tt.expectedType, err.Type)
			assert.Equal(t, tt.expectedSeverity, err.Severity)
			assert.Equal(t, message, err.Message)
			assert.Equal(t, details, err.Details)
			assert.Equal(t, tt.expectedCode, err.Code)
			assert.Equal(t, tt.expectedStatus, err.StatusCode)
			assert.NotNil(t, err.Context)
		})
	}
}

func TestCustomError_Error(t *testing.T) {
	tests := []struct {
		name      string
		customErr *CustomError
		expected  string
	}{
		{
			name: "error with underlying error",
			customErr: &CustomError{
				Message: "Custom error",
				Err:     errors.New("underlying error"),
			},
			expected: "Custom error: underlying error",
		},
		{
			name: "error without underlying error",
			customErr: &CustomError{
				Message: "Custom error",
			},
			expected: "Custom error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.customErr.Error())
		})
	}
}

func TestErrorHandlingMiddleware_ErrorMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := GetDefaultErrorHandlingConfig()
	config.TrackMetrics = true

	middleware := NewErrorHandlingMiddleware(config, logger)

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	// Simulate error
	customErr := CreateValidationError("test error", "test details")
	apiError := middleware.createAPIError(customErr, req, "test-request-id", http.StatusBadRequest)

	// Track error
	middleware.trackError(apiError)

	// Check metrics
	metrics := middleware.GetErrorMetrics()
	assert.Equal(t, int64(1), metrics.TotalErrors)
	assert.Equal(t, int64(1), metrics.ErrorsByType[ErrorTypeValidation])
	assert.Equal(t, int64(1), metrics.ErrorsBySeverity[ErrorSeverityLow])
	assert.Equal(t, apiError.Timestamp, metrics.LastError)
}

func TestErrorHandlingMiddleware_RequestIDIntegration(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := GetDefaultErrorHandlingConfig()
	middleware := NewErrorHandlingMiddleware(config, logger)

	// Test with request ID in context
	req := httptest.NewRequest("GET", "/test", nil)
	ctx := context.WithValue(req.Context(), "request_id", "test-request-id")
	req = req.WithContext(ctx)

	handler := middleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Equal(t, "test-request-id", rr.Header().Get("X-Request-ID"))

	// Check that error was logged with request ID
	assert.Greater(t, obs.Len(), 0)
	logs := obs.All()
	assert.Contains(t, logs[0].ContextMap(), "request_id")
}

func TestErrorHandlingMiddleware_RemoteAddressDetection(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := GetDefaultErrorHandlingConfig()
	middleware := NewErrorHandlingMiddleware(config, logger)

	tests := []struct {
		name         string
		headers      map[string]string
		remoteAddr   string
		expectedAddr string
	}{
		{
			name: "X-Forwarded-For single IP",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1",
			},
			remoteAddr:   "127.0.0.1:8080",
			expectedAddr: "192.168.1.1",
		},
		{
			name: "X-Forwarded-For multiple IPs",
			headers: map[string]string{
				"X-Forwarded-For": "192.168.1.1, 10.0.0.1, 172.16.0.1",
			},
			remoteAddr:   "127.0.0.1:8080",
			expectedAddr: "192.168.1.1",
		},
		{
			name: "X-Real-IP",
			headers: map[string]string{
				"X-Real-IP": "192.168.1.2",
			},
			remoteAddr:   "127.0.0.1:8080",
			expectedAddr: "192.168.1.2",
		},
		{
			name:         "no proxy headers",
			headers:      map[string]string{},
			remoteAddr:   "192.168.1.3:8080",
			expectedAddr: "192.168.1.3:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr

			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			addr := middleware.getRemoteAddr(req)
			assert.Equal(t, tt.expectedAddr, addr)
		})
	}
}

func TestErrorHandlingMiddleware_SafeHeaders(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := GetDefaultErrorHandlingConfig()
	middleware := NewErrorHandlingMiddleware(config, logger)

	headers := http.Header{}
	headers.Set("Authorization", "Bearer token123")
	headers.Set("Cookie", "session=abc123")
	headers.Set("X-API-Key", "key123")
	headers.Set("X-CSRF-Token", "csrf123")
	headers.Set("Content-Type", "application/json")
	headers.Set("User-Agent", "test-agent")

	safeHeaders := middleware.getSafeHeaders(headers)

	// Check that sensitive headers are masked
	assert.Equal(t, "[MASKED]", safeHeaders["Authorization"])
	assert.Equal(t, "[MASKED]", safeHeaders["Cookie"])
	assert.Equal(t, "[MASKED]", safeHeaders["X-API-Key"])
	assert.Equal(t, "[MASKED]", safeHeaders["X-CSRF-Token"])

	// Check that non-sensitive headers are preserved
	assert.Equal(t, "application/json", safeHeaders["Content-Type"])
	assert.Equal(t, "test-agent", safeHeaders["User-Agent"])
}

func TestPredefinedConfigurations(t *testing.T) {
	tests := []struct {
		name     string
		config   *ErrorHandlingConfig
		expected map[string]interface{}
	}{
		{
			name:   "default configuration",
			config: GetDefaultErrorHandlingConfig(),
			expected: map[string]interface{}{
				"log_errors":      true,
				"log_level":       zapcore.ErrorLevel,
				"include_stack":   false,
				"include_details": true,
				"include_context": false,
				"mask_internal":   true,
				"recover_panics":  true,
				"track_metrics":   true,
			},
		},
		{
			name:   "verbose configuration",
			config: GetVerboseErrorHandlingConfig(),
			expected: map[string]interface{}{
				"log_errors":      true,
				"log_level":       zapcore.DebugLevel,
				"include_stack":   true,
				"include_details": true,
				"include_context": true,
				"mask_internal":   false,
				"recover_panics":  true,
				"track_metrics":   true,
			},
		},
		{
			name:   "production configuration",
			config: GetProductionErrorHandlingConfig(),
			expected: map[string]interface{}{
				"log_errors":      true,
				"log_level":       zapcore.ErrorLevel,
				"include_stack":   false,
				"include_details": false,
				"include_context": false,
				"mask_internal":   true,
				"recover_panics":  true,
				"track_metrics":   true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected["log_errors"], tt.config.LogErrors)
			assert.Equal(t, tt.expected["log_level"], tt.config.LogLevel)
			assert.Equal(t, tt.expected["include_stack"], tt.config.IncludeStack)
			assert.Equal(t, tt.expected["include_details"], tt.config.IncludeDetails)
			assert.Equal(t, tt.expected["include_context"], tt.config.IncludeContext)
			assert.Equal(t, tt.expected["mask_internal"], tt.config.MaskInternal)
			assert.Equal(t, tt.expected["recover_panics"], tt.config.RecoverPanics)
			assert.Equal(t, tt.expected["track_metrics"], tt.config.TrackMetrics)

			// Check that error codes are initialized
			assert.NotEmpty(t, tt.config.ErrorCodes)
			assert.Equal(t, "INVALID_INPUT", tt.config.ErrorCodes["validation_error"])
			assert.Equal(t, "UNAUTHORIZED", tt.config.ErrorCodes["authentication_error"])
			assert.Equal(t, "INTERNAL_ERROR", tt.config.ErrorCodes["internal_error"])
		})
	}
}

func TestErrorHandlingMiddleware_ErrorContext(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := GetDefaultErrorHandlingConfig()
	config.IncludeContext = true
	middleware := NewErrorHandlingMiddleware(config, logger)

	req := httptest.NewRequest("GET", "/test?param=value", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")

	customErr := CreateValidationError("test error", "test details")
	customErr.Context["custom_field"] = "custom_value"

	apiError := middleware.createAPIError(customErr, req, "test-request-id", http.StatusBadRequest)

	assert.NotNil(t, apiError.Context)
	assert.Equal(t, "/test?param=value", apiError.Context["url"])
	assert.Equal(t, "custom_value", apiError.Context["custom_field"])

	headers, ok := apiError.Context["headers"].(map[string]string)
	assert.True(t, ok)
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "test-agent", headers["User-Agent"])
}

func TestErrorHandlingMiddleware_MaskInternalErrors(t *testing.T) {
	core, _ := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := GetDefaultErrorHandlingConfig()
	config.MaskInternal = true
	middleware := NewErrorHandlingMiddleware(config, logger)

	req := httptest.NewRequest("GET", "/test", nil)

	// Test internal error masking
	internalErr := CreateInternalError("internal error", "sensitive details")
	apiError := middleware.createAPIError(internalErr, req, "test-request-id", http.StatusInternalServerError)

	assert.Equal(t, "An internal error occurred", apiError.Details)
	assert.Nil(t, apiError.Context)

	// Test non-internal error not masked
	validationErr := CreateValidationError("validation error", "validation details")
	apiError = middleware.createAPIError(validationErr, req, "test-request-id", http.StatusBadRequest)

	assert.Equal(t, "validation details", apiError.Details)
}

func TestErrorHandlingMiddleware_ErrorLogging(t *testing.T) {
	core, obs := observer.New(zapcore.InfoLevel)
	logger := zap.New(core)

	config := GetDefaultErrorHandlingConfig()
	config.LogErrors = true
	config.LogLevel = zapcore.ErrorLevel
	middleware := NewErrorHandlingMiddleware(config, logger)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "test-request-id")

	customErr := CreateValidationError("test error", "test details")
	apiError := middleware.createAPIError(customErr, req, "test-request-id", http.StatusBadRequest)

	middleware.logError(apiError, customErr)

	assert.Greater(t, obs.Len(), 0)
	logs := obs.All()

	// Check that error was logged with correct fields
	logEntry := logs[0]
	assert.Equal(t, "API Error", logEntry.Message)
	assert.Equal(t, zapcore.ErrorLevel, logEntry.Level)

	contextMap := logEntry.ContextMap()
	assert.Equal(t, "validation_error", contextMap["error_type"])
	assert.Equal(t, "low", contextMap["error_severity"])
	assert.Equal(t, "INVALID_INPUT", contextMap["error_code"])
	assert.Equal(t, "test error", contextMap["error_message"])
	assert.Equal(t, "test-request-id", contextMap["request_id"])
	assert.Equal(t, "/test", contextMap["path"])
	assert.Equal(t, "GET", contextMap["method"])
}
