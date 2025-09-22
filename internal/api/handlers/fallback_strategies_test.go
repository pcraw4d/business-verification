package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"kyb-platform/internal/external"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewFallbackStrategiesHandler(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewFallbackStrategyManager(nil, logger)

	handler := NewFallbackStrategiesHandler(manager, logger)
	assert.NotNil(t, handler)
	assert.Equal(t, manager, handler.manager)
	assert.Equal(t, logger, handler.logger)
}

func TestFallbackStrategiesHandler_ExecuteFallbackStrategies(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewFallbackStrategyManager(nil, logger)
	handler := NewFallbackStrategiesHandler(manager, logger)

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "valid request",
			method: "POST",
			requestBody: ExecuteFallbackStrategiesRequest{
				URL:           "https://example.com",
				OriginalError: "test error",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid method",
			method:         "GET",
			requestBody:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  true,
		},
		{
			name:   "missing url",
			method: "POST",
			requestBody: ExecuteFallbackStrategiesRequest{
				URL:           "",
				OriginalError: "test error",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "missing original error",
			method: "POST",
			requestBody: ExecuteFallbackStrategiesRequest{
				URL:           "https://example.com",
				OriginalError: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(tt.method, "/api/v1/fallback/execute", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.ExecuteFallbackStrategies(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response ExecuteFallbackStrategiesResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Result)
			}
		})
	}
}

func TestFallbackStrategiesHandler_GetConfig(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewFallbackStrategyManager(nil, logger)
	handler := NewFallbackStrategiesHandler(manager, logger)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedError  bool
	}{
		{
			name:           "valid request",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid method",
			method:         "POST",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/api/v1/fallback/config", nil)
			w := httptest.NewRecorder()

			handler.GetConfig(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response GetConfigResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Config)
			}
		})
	}
}

func TestFallbackStrategiesHandler_UpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewFallbackStrategyManager(nil, logger)
	handler := NewFallbackStrategiesHandler(manager, logger)

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "valid request",
			method: "PUT",
			requestBody: UpdateFallbackConfigRequest{
				Config: &external.FallbackConfig{
					EnableUserAgentRotation:   true,
					EnableHeaderCustomization: true,
					MaxFallbackAttempts:       3,
					FallbackDelay:             1 * time.Second,
				},
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid method",
			method:         "GET",
			requestBody:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  true,
		},
		{
			name:   "missing config",
			method: "PUT",
			requestBody: UpdateFallbackConfigRequest{
				Config: nil,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(tt.method, "/api/v1/fallback/config", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.UpdateConfig(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response UpdateFallbackConfigResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestFallbackStrategiesHandler_AddProxy(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewFallbackStrategyManager(nil, logger)
	handler := NewFallbackStrategiesHandler(manager, logger)

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "valid request",
			method: "POST",
			requestBody: AddProxyRequest{
				Proxy: &external.Proxy{
					Host:     "proxy.example.com",
					Port:     8080,
					Protocol: "http",
					Active:   true,
				},
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid method",
			method:         "GET",
			requestBody:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  true,
		},
		{
			name:   "missing proxy",
			method: "POST",
			requestBody: AddProxyRequest{
				Proxy: nil,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "missing host",
			method: "POST",
			requestBody: AddProxyRequest{
				Proxy: &external.Proxy{
					Host:     "",
					Port:     8080,
					Protocol: "http",
					Active:   true,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "invalid port",
			method: "POST",
			requestBody: AddProxyRequest{
				Proxy: &external.Proxy{
					Host:     "proxy.example.com",
					Port:     0,
					Protocol: "http",
					Active:   true,
				},
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(tt.method, "/api/v1/fallback/proxy", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.AddProxy(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response AddProxyResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestFallbackStrategiesHandler_RemoveProxy(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewFallbackStrategyManager(nil, logger)
	handler := NewFallbackStrategiesHandler(manager, logger)

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "valid request",
			method: "DELETE",
			requestBody: RemoveProxyRequest{
				Host: "proxy.example.com",
				Port: 8080,
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid method",
			method:         "GET",
			requestBody:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  true,
		},
		{
			name:   "missing host",
			method: "DELETE",
			requestBody: RemoveProxyRequest{
				Host: "",
				Port: 8080,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "invalid port",
			method: "DELETE",
			requestBody: RemoveProxyRequest{
				Host: "proxy.example.com",
				Port: 0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(tt.method, "/api/v1/fallback/proxy", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.RemoveProxy(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response RemoveProxyResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
			}
		})
	}
}

func TestFallbackStrategiesHandler_TestFallbackStrategy(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewFallbackStrategyManager(nil, logger)
	handler := NewFallbackStrategiesHandler(manager, logger)

	tests := []struct {
		name           string
		method         string
		requestBody    interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "valid user agent rotation test",
			method: "POST",
			requestBody: TestFallbackStrategyRequest{
				URL:      "https://example.com",
				Strategy: "user_agent_rotation",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "valid header customization test",
			method: "POST",
			requestBody: TestFallbackStrategyRequest{
				URL:      "https://example.com",
				Strategy: "header_customization",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "valid proxy rotation test",
			method: "POST",
			requestBody: TestFallbackStrategyRequest{
				URL:      "https://example.com",
				Strategy: "proxy_rotation",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "valid alternative sources test",
			method: "POST",
			requestBody: TestFallbackStrategyRequest{
				URL:      "https://example.com",
				Strategy: "alternative_sources",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:           "invalid method",
			method:         "GET",
			requestBody:    nil,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedError:  true,
		},
		{
			name:   "missing url",
			method: "POST",
			requestBody: TestFallbackStrategyRequest{
				URL:      "",
				Strategy: "user_agent_rotation",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "missing strategy",
			method: "POST",
			requestBody: TestFallbackStrategyRequest{
				URL:      "https://example.com",
				Strategy: "",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:   "invalid strategy",
			method: "POST",
			requestBody: TestFallbackStrategyRequest{
				URL:      "https://example.com",
				Strategy: "invalid_strategy",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error
			if tt.requestBody != nil {
				body, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(tt.method, "/api/v1/fallback/test", bytes.NewBuffer(body))
			w := httptest.NewRecorder()

			handler.TestFallbackStrategy(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if !tt.expectedError {
				var response TestFallbackStrategyResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.True(t, response.Success)
				assert.NotNil(t, response.Result)
			}
		})
	}
}

func TestFallbackStrategiesHandler_RegisterRoutes(t *testing.T) {
	logger := zap.NewNop()
	manager := external.NewFallbackStrategyManager(nil, logger)
	handler := NewFallbackStrategiesHandler(manager, logger)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register routes
	handler.RegisterRoutes(mux)

	// Test that routes are registered by checking if they respond to requests
	endpoints := []struct {
		path   string
		method string
	}{
		{"/api/v1/fallback/execute", "POST"},
		{"/api/v1/fallback/config", "GET"},
		{"/api/v1/fallback/config", "PUT"},
		{"/api/v1/fallback/proxy", "POST"},
		{"/api/v1/fallback/proxy", "DELETE"},
		{"/api/v1/fallback/test", "POST"},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.path+"_"+endpoint.method, func(t *testing.T) {
			req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			// Should not get 404 (route not found)
			assert.NotEqual(t, http.StatusNotFound, w.Code)
		})
	}
}

func TestRequestResponseStructs(t *testing.T) {
	// Test ExecuteFallbackStrategiesRequest
	req := ExecuteFallbackStrategiesRequest{
		URL:           "https://example.com",
		OriginalError: "test error",
	}
	assert.Equal(t, "https://example.com", req.URL)
	assert.Equal(t, "test error", req.OriginalError)

	// Test ExecuteFallbackStrategiesResponse
	resp := ExecuteFallbackStrategiesResponse{
		Success: true,
		Result:  &external.FallbackResult{},
		Error:   "",
	}
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Result)
	assert.Empty(t, resp.Error)

	// Test GetConfigResponse
	configResp := GetConfigResponse{
		Success: true,
		Config:  &external.FallbackConfig{},
		Error:   "",
	}
	assert.True(t, configResp.Success)
	assert.NotNil(t, configResp.Config)
	assert.Empty(t, configResp.Error)

	// Test UpdateFallbackConfigRequest
	updateReq := UpdateFallbackConfigRequest{
		Config: &external.FallbackConfig{
			EnableUserAgentRotation: true,
		},
	}
	assert.NotNil(t, updateReq.Config)
	assert.True(t, updateReq.Config.EnableUserAgentRotation)

	// Test UpdateFallbackConfigResponse
	updateResp := UpdateFallbackConfigResponse{
		Success: true,
		Error:   "",
	}
	assert.True(t, updateResp.Success)
	assert.Empty(t, updateResp.Error)

	// Test AddProxyRequest
	addProxyReq := AddProxyRequest{
		Proxy: &external.Proxy{
			Host:     "proxy.example.com",
			Port:     8080,
			Protocol: "http",
			Active:   true,
		},
	}
	assert.NotNil(t, addProxyReq.Proxy)
	assert.Equal(t, "proxy.example.com", addProxyReq.Proxy.Host)
	assert.Equal(t, 8080, addProxyReq.Proxy.Port)

	// Test AddProxyResponse
	addProxyResp := AddProxyResponse{
		Success: true,
		Error:   "",
	}
	assert.True(t, addProxyResp.Success)
	assert.Empty(t, addProxyResp.Error)

	// Test RemoveProxyRequest
	removeProxyReq := RemoveProxyRequest{
		Host: "proxy.example.com",
		Port: 8080,
	}
	assert.Equal(t, "proxy.example.com", removeProxyReq.Host)
	assert.Equal(t, 8080, removeProxyReq.Port)

	// Test RemoveProxyResponse
	removeProxyResp := RemoveProxyResponse{
		Success: true,
		Error:   "",
	}
	assert.True(t, removeProxyResp.Success)
	assert.Empty(t, removeProxyResp.Error)

	// Test TestFallbackStrategyRequest
	testReq := TestFallbackStrategyRequest{
		URL:      "https://example.com",
		Strategy: "user_agent_rotation",
	}
	assert.Equal(t, "https://example.com", testReq.URL)
	assert.Equal(t, "user_agent_rotation", testReq.Strategy)

	// Test TestFallbackStrategyResponse
	testResp := TestFallbackStrategyResponse{
		Success: true,
		Result:  &external.FallbackResult{},
		Error:   "",
	}
	assert.True(t, testResp.Success)
	assert.NotNil(t, testResp.Result)
	assert.Empty(t, testResp.Error)
}
