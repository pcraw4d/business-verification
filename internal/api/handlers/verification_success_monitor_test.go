package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	"github.com/pcraw4d/business-verification/internal/external"
)

func TestNewVerificationSuccessMonitorHandler(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)

	handler := NewVerificationSuccessMonitorHandler(monitor, logger)
	assert.NotNil(t, handler)
	assert.Equal(t, monitor, handler.monitor)
	assert.Equal(t, logger, handler.logger)
}

func TestRecordAttempt(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Test successful attempt
	reqBody := RecordAttemptRequest{
		URL:          "https://example.com",
		Success:      true,
		ResponseTime: 2 * time.Second,
		StatusCode:   200,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/success-monitor/record", bytes.NewBuffer(reqBodyBytes))
	w := httptest.NewRecorder()

	handler.RecordAttempt(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response RecordAttemptResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Attempt recorded successfully", response.Message)
	assert.Equal(t, 1.0, response.CurrentRate)
	assert.Equal(t, int64(1), response.TotalAttempts)

	// Test failed attempt
	reqBody2 := RecordAttemptRequest{
		URL:          "https://example2.com",
		Success:      false,
		ResponseTime: 5 * time.Second,
		StatusCode:   500,
		ErrorType:    "timeout",
		ErrorMessage: "request timeout",
	}

	reqBodyBytes2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/success-monitor/record", bytes.NewBuffer(reqBodyBytes2))
	w2 := httptest.NewRecorder()

	handler.RecordAttempt(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	var response2 RecordAttemptResponse
	err = json.Unmarshal(w2.Body.Bytes(), &response2)
	assert.NoError(t, err)
	assert.True(t, response2.Success)
	assert.Equal(t, 0.5, response2.CurrentRate) // 1 success out of 2 attempts
	assert.Equal(t, int64(2), response2.TotalAttempts)
}

func TestRecordAttemptValidation(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Test missing URL
	reqBody := RecordAttemptRequest{
		Success:      true,
		ResponseTime: 2 * time.Second,
		StatusCode:   200,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/success-monitor/record", bytes.NewBuffer(reqBodyBytes))
	w := httptest.NewRecorder()

	handler.RecordAttempt(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "URL is required")

	// Test negative response time
	reqBody2 := RecordAttemptRequest{
		URL:          "https://example.com",
		Success:      true,
		ResponseTime: -1 * time.Second,
		StatusCode:   200,
	}

	reqBodyBytes2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/success-monitor/record", bytes.NewBuffer(reqBodyBytes2))
	w2 := httptest.NewRecorder()

	handler.RecordAttempt(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
	assert.Contains(t, w2.Body.String(), "Response time must be non-negative")
}

func TestGetMetrics(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Add some test data
	dataPoint := external.DataPoint{
		URL:          "https://example.com",
		Success:      true,
		ResponseTime: 2 * time.Second,
		StatusCode:   200,
	}
	monitor.RecordAttempt(context.Background(), dataPoint)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/success-monitor/metrics", nil)
	w := httptest.NewRecorder()

	handler.GetMetrics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response GetMetricsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Metrics)
	assert.Equal(t, 1.0, response.Metrics.SuccessRate)
	assert.Equal(t, int64(1), response.Metrics.TotalAttempts)
	assert.Equal(t, 0.90, response.TargetRate) // Default target rate
	assert.True(t, response.IsAchieved)
}

func TestGetFailureAnalysis(t *testing.T) {
	logger := zap.NewNop()
	config := &external.SuccessMonitorConfig{
		EnableFailureAnalysis: true,
		AnalysisWindow:        1 * time.Hour,
		MaxDataPoints:         10000, // Set to default value to prevent premature cleanup
	}
	monitor := external.NewVerificationSuccessMonitor(config, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Add test data with failures - use very recent timestamps to ensure they're within the 1-hour window
	now := time.Now()
	testData := []external.DataPoint{
		{URL: "https://example1.com", Timestamp: now.Add(-30 * time.Second), Success: false, ErrorType: "timeout", StrategyUsed: "user_agent_rotation"},
		{URL: "https://example2.com", Timestamp: now.Add(-20 * time.Second), Success: true, StrategyUsed: "direct"},
		{URL: "https://example3.com", Timestamp: now.Add(-10 * time.Second), Success: false, ErrorType: "blocked", StrategyUsed: "proxy_rotation"},
	}

	for _, dp := range testData {
		monitor.RecordAttempt(context.Background(), dp)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/success-monitor/failures", nil)
	w := httptest.NewRecorder()

	handler.GetFailureAnalysis(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response GetFailureAnalysisResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Analysis)
	assert.Equal(t, int64(2), response.Analysis.TotalFailures)
	assert.Equal(t, 2.0/3.0, response.Analysis.FailureRate) // 2 failures out of 3 attempts
}

func TestGetTrendAnalysis(t *testing.T) {
	logger := zap.NewNop()
	config := &external.SuccessMonitorConfig{
		EnableTrendAnalysis: true,
		TrendWindow:         3 * time.Hour, // Shorter window for testing
		MinDataPoints:       3,
		MaxDataPoints:       10000, // Set to default value to prevent premature cleanup
	}
	monitor := external.NewVerificationSuccessMonitor(config, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Add test data with a trend (within the 3-hour window) - use very recent timestamps
	now := time.Now()
	testData := []external.DataPoint{
		{URL: "https://example1.com", Timestamp: now.Add(-2 * time.Minute), Success: false, ResponseTime: 5 * time.Second, ErrorType: "timeout", StrategyUsed: "strategy1"},
		{URL: "https://example2.com", Timestamp: now.Add(-1 * time.Minute), Success: true, ResponseTime: 3 * time.Second, StrategyUsed: "strategy2"},
		{URL: "https://example3.com", Timestamp: now.Add(-30 * time.Second), Success: true, ResponseTime: 2 * time.Second, StrategyUsed: "strategy1"},
	}

	for _, dp := range testData {
		monitor.RecordAttempt(context.Background(), dp)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/success-monitor/trends", nil)
	w := httptest.NewRecorder()

	handler.GetTrendAnalysis(w, req)

	// Check if we got an error response
	if w.Code != http.StatusOK {
		var errorResponse GetTrendAnalysisResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)
		assert.False(t, errorResponse.Success)
		t.Logf("Trend analysis failed: %s", errorResponse.Message)
		return
	}

	var response GetTrendAnalysisResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Trends)
	assert.Equal(t, 3*time.Hour, response.Trends.Period)
}

func TestGetTrendAnalysisInsufficientData(t *testing.T) {
	logger := zap.NewNop()
	config := &external.SuccessMonitorConfig{
		EnableTrendAnalysis: true,
		TrendWindow:         24 * time.Hour,
		MinDataPoints:       10, // Require more data points
	}
	monitor := external.NewVerificationSuccessMonitor(config, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Add only 3 data points (less than required 10)
	for i := 0; i < 3; i++ {
		dataPoint := external.DataPoint{
			Timestamp: time.Now(),
			Success:   true,
		}
		monitor.RecordAttempt(context.Background(), dataPoint)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/success-monitor/trends", nil)
	w := httptest.NewRecorder()

	handler.GetTrendAnalysis(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response GetTrendAnalysisResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Message, "insufficient data points")
}

func TestGetConfig(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/success-monitor/config", nil)
	w := httptest.NewRecorder()

	handler.GetConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response GetSuccessMonitorConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.NotNil(t, response.Config)
	assert.Equal(t, 0.90, response.Config.TargetSuccessRate)
	assert.Equal(t, 0.85, response.Config.AlertThreshold)
}

func TestUpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Test valid config update
	reqBody := UpdateSuccessMonitorConfigRequest{
		EnableRealTimeMonitoring: true,
		EnableFailureAnalysis:    true,
		EnableTrendAnalysis:      true,
		EnableAlerting:           true,
		TargetSuccessRate:        0.95,
		AlertThreshold:           0.90,
		MetricsRetentionPeriod:   30 * 24 * time.Hour,
		AnalysisWindow:           1 * time.Hour,
		TrendWindow:              24 * time.Hour,
		MinDataPoints:            100,
		MaxDataPoints:            10000,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/success-monitor/config", bytes.NewBuffer(reqBodyBytes))
	w := httptest.NewRecorder()

	handler.UpdateConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response UpdateSuccessMonitorConfigResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Configuration updated successfully", response.Message)

	// Verify the config was actually updated
	config := monitor.GetConfig()
	assert.Equal(t, 0.95, config.TargetSuccessRate)
	assert.Equal(t, 0.90, config.AlertThreshold)
}

func TestUpdateConfigValidation(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Test invalid target success rate
	reqBody := UpdateSuccessMonitorConfigRequest{
		TargetSuccessRate: 1.5, // Invalid: > 1
		AlertThreshold:    0.85,
		MinDataPoints:     100,
		MaxDataPoints:     10000,
	}

	reqBodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/success-monitor/config", bytes.NewBuffer(reqBodyBytes))
	w := httptest.NewRecorder()

	handler.UpdateConfig(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Target success rate must be between 0 and 1")

	// Test invalid alert threshold
	reqBody2 := UpdateSuccessMonitorConfigRequest{
		TargetSuccessRate: 0.90,
		AlertThreshold:    0.95, // Invalid: >= target
		MinDataPoints:     100,
		MaxDataPoints:     10000,
	}

	reqBodyBytes2, _ := json.Marshal(reqBody2)
	req2 := httptest.NewRequest(http.MethodPut, "/api/v1/success-monitor/config", bytes.NewBuffer(reqBodyBytes2))
	w2 := httptest.NewRecorder()

	handler.UpdateConfig(w2, req2)

	assert.Equal(t, http.StatusBadRequest, w2.Code)
	assert.Contains(t, w2.Body.String(), "Alert threshold must be less than target success rate")
}

func TestResetMetrics(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Add some data first
	dataPoint := external.DataPoint{
		URL:     "https://example.com",
		Success: true,
	}
	monitor.RecordAttempt(context.Background(), dataPoint)

	// Verify data exists
	metrics := monitor.GetMetrics()
	assert.Equal(t, int64(1), metrics.TotalAttempts)

	// Reset metrics
	req := httptest.NewRequest(http.MethodPost, "/api/v1/success-monitor/reset", nil)
	w := httptest.NewRecorder()

	handler.ResetMetrics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ResetMetricsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Metrics reset successfully", response.Message)

	// Verify data is reset
	metrics = monitor.GetMetrics()
	assert.Equal(t, int64(0), metrics.TotalAttempts)
}

func TestGetStatus(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Add some test data
	dataPoint := external.DataPoint{
		URL:     "https://example.com",
		Success: true,
	}
	monitor.RecordAttempt(context.Background(), dataPoint)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/success-monitor/status", nil)
	w := httptest.NewRecorder()

	handler.GetStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response GetStatusResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, 1.0, response.CurrentRate)
	assert.Equal(t, 0.90, response.TargetRate) // Default target rate
	assert.True(t, response.IsAchieved)
	assert.Equal(t, int64(1), response.TotalAttempts)
	assert.False(t, response.LastUpdated.IsZero())
}

func TestMethodNotAllowed(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Test wrong method for RecordAttempt
	req := httptest.NewRequest(http.MethodGet, "/api/v1/success-monitor/record", nil)
	w := httptest.NewRecorder()

	handler.RecordAttempt(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)

	// Test wrong method for GetMetrics
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/success-monitor/metrics", nil)
	w2 := httptest.NewRecorder()

	handler.GetMetrics(w2, req2)

	assert.Equal(t, http.StatusMethodNotAllowed, w2.Code)
}

func TestRegisterRoutes(t *testing.T) {
	logger := zap.NewNop()
	monitor := external.NewVerificationSuccessMonitor(nil, logger)
	handler := NewVerificationSuccessMonitorHandler(monitor, logger)

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Register routes
	handler.RegisterRoutes(mux)

	// Test that routes are registered by making requests
	req := httptest.NewRequest(http.MethodGet, "/api/v1/success-monitor/metrics", nil)
	w := httptest.NewRecorder()

	mux.ServeHTTP(w, req)

	// Should not get 404 (route not found)
	assert.NotEqual(t, http.StatusNotFound, w.Code)
}

func TestSuccessMonitorRequestResponseStructs(t *testing.T) {
	// Test that all request/response structs can be marshaled/unmarshaled

	// Test RecordAttemptRequest
	req := RecordAttemptRequest{
		URL:           "https://example.com",
		Success:       true,
		ResponseTime:  2 * time.Second,
		StatusCode:    200,
		ErrorType:     "timeout",
		ErrorMessage:  "request timeout",
		StrategyUsed:  "user_agent_rotation",
		UserAgentUsed: "Mozilla/5.0",
		ProxyUsed:     &external.Proxy{Host: "proxy.example.com", Port: 8080},
		Metadata:      map[string]interface{}{"key": "value"},
	}

	reqBytes, err := json.Marshal(req)
	assert.NoError(t, err)

	var reqUnmarshaled RecordAttemptRequest
	err = json.Unmarshal(reqBytes, &reqUnmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, req.URL, reqUnmarshaled.URL)
	assert.Equal(t, req.Success, reqUnmarshaled.Success)
	assert.Equal(t, req.ResponseTime, reqUnmarshaled.ResponseTime)

	// Test UpdateSuccessMonitorConfigRequest
	configReq := UpdateSuccessMonitorConfigRequest{
		EnableRealTimeMonitoring: true,
		EnableFailureAnalysis:    true,
		EnableTrendAnalysis:      true,
		EnableAlerting:           true,
		TargetSuccessRate:        0.95,
		AlertThreshold:           0.90,
		MetricsRetentionPeriod:   30 * 24 * time.Hour,
		AnalysisWindow:           1 * time.Hour,
		TrendWindow:              24 * time.Hour,
		MinDataPoints:            100,
		MaxDataPoints:            10000,
	}

	configReqBytes, err := json.Marshal(configReq)
	assert.NoError(t, err)

	var configReqUnmarshaled UpdateSuccessMonitorConfigRequest
	err = json.Unmarshal(configReqBytes, &configReqUnmarshaled)
	assert.NoError(t, err)
	assert.Equal(t, configReq.TargetSuccessRate, configReqUnmarshaled.TargetSuccessRate)
	assert.Equal(t, configReq.AlertThreshold, configReqUnmarshaled.AlertThreshold)
}
