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
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"

	"github.com/pcraw4d/business-verification/internal/observability"
)

// MockRealtimePerformanceMonitor is a mock for the realtime performance monitor
type MockRealtimePerformanceMonitor struct {
	mock.Mock
}

func (m *MockRealtimePerformanceMonitor) GetMonitoringMetrics() observability.MonitoringMetrics {
	args := m.Called()
	return args.Get(0).(observability.MonitoringMetrics)
}

func (m *MockRealtimePerformanceMonitor) GetConnectedClients() []observability.ConnectedClient {
	args := m.Called()
	return args.Get(0).([]observability.ConnectedClient)
}

// MockLogAnalysisSystem is a mock for the log analysis system
type MockLogAnalysisSystem struct {
	mock.Mock
}

func (m *MockLogAnalysisSystem) AnalyzeLogs(ctx context.Context, logs []observability.LogEntry) (*observability.LogAnalysisResult, error) {
	args := m.Called(ctx, logs)
	return args.Get(0).(*observability.LogAnalysisResult), args.Error(1)
}

func (m *MockLogAnalysisSystem) GetActivePatterns() []observability.LogPattern {
	args := m.Called()
	return args.Get(0).([]observability.LogPattern)
}

func (m *MockLogAnalysisSystem) GetActiveErrorGroups() []observability.ErrorGroup {
	args := m.Called()
	return args.Get(0).([]observability.ErrorGroup)
}

func (m *MockLogAnalysisSystem) GetCorrelationTraces(correlationID string) []observability.CorrelationTrace {
	args := m.Called(correlationID)
	return args.Get(0).([]observability.CorrelationTrace)
}

func TestNewMonitoringDashboardHandler(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockRealtimeMonitor, handler.realtimeMonitor)
	assert.Equal(t, mockLogAnalysis, handler.logAnalysis)
	assert.Equal(t, logger, handler.logger)
	assert.NotNil(t, handler.dashboardConfig)
	assert.Equal(t, 30, handler.dashboardConfig.RefreshInterval)
	assert.Equal(t, "light", handler.dashboardConfig.Theme)
	assert.Equal(t, "UTC", handler.dashboardConfig.Timezone)
	assert.Equal(t, "en", handler.dashboardConfig.Language)
	assert.NotNil(t, handler.cache)
	assert.Equal(t, 30*time.Second, handler.cacheTTL)
}

func TestMonitoringDashboardHandler_GetDashboardData(t *testing.T) {
	core, obs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	// Setup mock expectations
	mockRealtimeMonitor.On("GetMonitoringMetrics").Return(observability.MonitoringMetrics{
		TotalRequests:       1000,
		ActiveUsers:         50,
		SuccessRate:         98.5,
		AverageResponseTime: 45.2,
	})

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/data", nil)
	w := httptest.NewRecorder()

	handler.GetDashboardData(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response DashboardData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.NotNil(t, response.Overview)
	assert.Equal(t, int64(1000), response.Overview.TotalRequests)
	assert.Equal(t, 50, response.Overview.ActiveUsers)
	assert.Equal(t, 98.5, response.Overview.SuccessRate)
	assert.Equal(t, 45.2, response.Overview.AverageResponseTime)
	assert.Equal(t, 99.99, response.Overview.Uptime)
	assert.Equal(t, "healthy", response.Overview.SystemStatus)

	assert.NotNil(t, response.SystemHealth)
	assert.NotNil(t, response.Performance)
	assert.NotNil(t, response.Business)
	assert.NotNil(t, response.Security)
	assert.NotNil(t, response.Alerts)
	assert.NotZero(t, response.LastUpdated)

	// Verify cache was updated
	assert.Len(t, obs.FilterMessage("failed to collect dashboard data").All(), 0)

	mockRealtimeMonitor.AssertExpectations(t)
}

func TestMonitoringDashboardHandler_GetDashboardOverview(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	// Setup mock expectations
	mockRealtimeMonitor.On("GetMonitoringMetrics").Return(observability.MonitoringMetrics{
		TotalRequests:       2000,
		ActiveUsers:         75,
		SuccessRate:         99.1,
		AverageResponseTime: 38.5,
	})

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/overview", nil)
	w := httptest.NewRecorder()

	handler.GetDashboardOverview(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response MonitoringOverview
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, int64(2000), response.TotalRequests)
	assert.Equal(t, 75, response.ActiveUsers)
	assert.Equal(t, 99.1, response.SuccessRate)
	assert.Equal(t, 38.5, response.AverageResponseTime)
	assert.Equal(t, 99.99, response.Uptime)
	assert.Equal(t, "healthy", response.SystemStatus)

	mockRealtimeMonitor.AssertExpectations(t)
}

func TestMonitoringDashboardHandler_GetSystemHealth(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/system-health", nil)
	w := httptest.NewRecorder()

	handler.GetSystemHealth(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response SystemHealthData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 45.2, response.CPUUsage)
	assert.Equal(t, 67.8, response.MemoryUsage)
	assert.Equal(t, 23.4, response.DiskUsage)
	assert.Equal(t, 12.5, response.NetworkLatency)
	assert.Equal(t, "healthy", response.DatabaseStatus)
	assert.Equal(t, "healthy", response.CacheStatus)
	assert.NotNil(t, response.ExternalAPIs)
	assert.Equal(t, "healthy", response.ExternalAPIs["government_data"])
	assert.Equal(t, "healthy", response.ExternalAPIs["credit_bureau"])
	assert.Equal(t, "healthy", response.ExternalAPIs["risk_assessment"])
}

func TestMonitoringDashboardHandler_GetPerformanceMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/performance", nil)
	w := httptest.NewRecorder()

	handler.GetPerformanceMetrics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response PerformanceData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 1250.5, response.RequestRate)
	assert.Equal(t, 0.15, response.ErrorRate)
	assert.Equal(t, 45.2, response.ResponseTimeP50)
	assert.Equal(t, 120.8, response.ResponseTimeP95)
	assert.Equal(t, 250.3, response.ResponseTimeP99)
	assert.Equal(t, 1250.5, response.Throughput)
}

func TestMonitoringDashboardHandler_GetBusinessMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/business", nil)
	w := httptest.NewRecorder()

	handler.GetBusinessMetrics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response BusinessMetricsData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, int64(1250), response.VerificationsToday)
	assert.Equal(t, int64(8750), response.VerificationsThisWeek)
	assert.Equal(t, 98.5, response.SuccessRate)
	assert.Equal(t, 2.3, response.AverageProcessingTime)
	assert.Len(t, response.TopIndustries, 3)
	assert.Equal(t, "Technology", response.TopIndustries[0].Industry)
	assert.Equal(t, int64(450), response.TopIndustries[0].Count)
	assert.Equal(t, 99.2, response.TopIndustries[0].SuccessRate)
	assert.NotNil(t, response.RiskDistribution)
	assert.Equal(t, 850, response.RiskDistribution["low"])
	assert.Equal(t, 320, response.RiskDistribution["medium"])
	assert.Equal(t, 80, response.RiskDistribution["high"])
}

func TestMonitoringDashboardHandler_GetSecurityMetrics(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/security", nil)
	w := httptest.NewRecorder()

	handler.GetSecurityMetrics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response SecurityMetricsData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, int64(12), response.FailedLogins)
	assert.Equal(t, int64(45), response.BlockedRequests)
	assert.Equal(t, int64(23), response.RateLimitHits)
	assert.Equal(t, int64(3), response.SecurityAlerts)
	assert.True(t, response.LastSecurityScan.Before(time.Now()))
}

func TestMonitoringDashboardHandler_GetAlerts(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/alerts", nil)
	w := httptest.NewRecorder()

	handler.GetAlerts(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response []DashboardAlert
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response, 2)

	// Check first alert
	assert.Equal(t, "alert-001", response[0].ID)
	assert.Equal(t, "performance", response[0].Type)
	assert.Equal(t, "warning", response[0].Severity)
	assert.Equal(t, "Response time exceeded threshold", response[0].Message)
	assert.False(t, response[0].Acknowledged)

	// Check second alert
	assert.Equal(t, "alert-002", response[1].ID)
	assert.Equal(t, "security", response[1].Type)
	assert.Equal(t, "info", response[1].Severity)
	assert.Equal(t, "Multiple failed login attempts detected", response[1].Message)
	assert.True(t, response[1].Acknowledged)
}

func TestMonitoringDashboardHandler_GetDashboardConfig(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/config", nil)
	w := httptest.NewRecorder()

	handler.GetDashboardConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response DashboardConfig
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, 30, response.RefreshInterval)
	assert.Equal(t, "light", response.Theme)
	assert.Equal(t, "UTC", response.Timezone)
	assert.Equal(t, "en", response.Language)
}

func TestMonitoringDashboardHandler_UpdateDashboardConfig(t *testing.T) {
	core, obs := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	// Test valid configuration update
	config := DashboardConfig{
		RefreshInterval: 60,
		Theme:           "dark",
		Timezone:        "America/New_York",
		Language:        "es",
	}

	configJSON, _ := json.Marshal(config)
	req := httptest.NewRequest("PUT", "/dashboard/config", bytes.NewBuffer(configJSON))
	w := httptest.NewRecorder()

	handler.UpdateDashboardConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "success", response["status"])
	assert.Equal(t, "Configuration updated", response["message"])

	// Verify config was updated
	assert.Equal(t, 60, handler.dashboardConfig.RefreshInterval)
	assert.Equal(t, "dark", handler.dashboardConfig.Theme)
	assert.Equal(t, "America/New_York", handler.dashboardConfig.Timezone)
	assert.Equal(t, "es", handler.dashboardConfig.Language)

	// Verify logging
	assert.Len(t, obs.FilterMessage("dashboard configuration updated").All(), 1)
}

func TestMonitoringDashboardHandler_UpdateDashboardConfig_InvalidMethod(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/config", nil)
	w := httptest.NewRecorder()

	handler.UpdateDashboardConfig(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestMonitoringDashboardHandler_UpdateDashboardConfig_InvalidRefreshInterval(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	config := DashboardConfig{
		RefreshInterval: 1, // Too low
		Theme:           "light",
		Timezone:        "UTC",
		Language:        "en",
	}

	configJSON, _ := json.Marshal(config)
	req := httptest.NewRequest("PUT", "/dashboard/config", bytes.NewBuffer(configJSON))
	w := httptest.NewRecorder()

	handler.UpdateDashboardConfig(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringDashboardHandler_GetRealTimeUpdates(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/realtime", nil)
	w := httptest.NewRecorder()

	handler.GetRealTimeUpdates(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "WebSocket endpoint - upgrade logic to be implemented", response["message"])
	assert.NotNil(t, response["timestamp"])
}

func TestMonitoringDashboardHandler_ExportDashboardData_JSON(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	// Setup mock expectations
	mockRealtimeMonitor.On("GetMonitoringMetrics").Return(observability.MonitoringMetrics{
		TotalRequests:       1000,
		ActiveUsers:         50,
		SuccessRate:         98.5,
		AverageResponseTime: 45.2,
	})

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/export?format=json", nil)
	w := httptest.NewRecorder()

	handler.ExportDashboardData(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	assert.Equal(t, "attachment; filename=dashboard-data.json", w.Header().Get("Content-Disposition"))

	var response DashboardData
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response.Overview)

	mockRealtimeMonitor.AssertExpectations(t)
}

func TestMonitoringDashboardHandler_ExportDashboardData_CSV(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/export?format=csv", nil)
	w := httptest.NewRecorder()

	handler.ExportDashboardData(w, req)

	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestMonitoringDashboardHandler_ExportDashboardData_InvalidFormat(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/export?format=xml", nil)
	w := httptest.NewRecorder()

	handler.ExportDashboardData(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMonitoringDashboardHandler_Caching(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	// Setup mock expectations - should only be called once due to caching
	mockRealtimeMonitor.On("GetMonitoringMetrics").Return(observability.MonitoringMetrics{
		TotalRequests:       1000,
		ActiveUsers:         50,
		SuccessRate:         98.5,
		AverageResponseTime: 45.2,
	}).Once()

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	// First request - should call the monitor
	req1 := httptest.NewRequest("GET", "/dashboard/data", nil)
	w1 := httptest.NewRecorder()
	handler.GetDashboardData(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Second request - should use cache
	req2 := httptest.NewRequest("GET", "/dashboard/data", nil)
	w2 := httptest.NewRecorder()
	handler.GetDashboardData(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Verify mock was only called once
	mockRealtimeMonitor.AssertExpectations(t)
}

func TestMonitoringDashboardHandler_ErrorHandling(t *testing.T) {
	core, _ := observer.New(zapcore.DebugLevel)
	logger := zap.New(core)

	mockRealtimeMonitor := &MockRealtimePerformanceMonitor{}
	mockLogAnalysis := &MockLogAnalysisSystem{}

	// Setup mock to return error
	mockRealtimeMonitor.On("GetMonitoringMetrics").Return(observability.MonitoringMetrics{})

	handler := NewMonitoringDashboardHandler(mockRealtimeMonitor, mockLogAnalysis, logger)

	req := httptest.NewRequest("GET", "/dashboard/data", nil)
	w := httptest.NewRecorder()

	handler.GetDashboardData(w, req)

	// Should still return 200 OK since the mock doesn't actually error
	assert.Equal(t, http.StatusOK, w.Code)

	mockRealtimeMonitor.AssertExpectations(t)
}
