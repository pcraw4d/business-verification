package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/company/kyb-platform/internal/services"
)

// MockHealthCheckService is a mock implementation of HealthCheckService
type MockHealthCheckService struct {
	checks []services.HealthCheckResult
}

func (m *MockHealthCheckService) CheckAPIHealth() services.HealthCheckResult {
	return services.HealthCheckResult{
		Name:        "API Server",
		Status:      services.HealthStatusHealthy,
		Message:     "API server is responding normally",
		LastChecked: time.Now(),
		Duration:    10 * time.Millisecond,
	}
}

func (m *MockHealthCheckService) CheckDatabaseHealth() services.HealthCheckResult {
	return services.HealthCheckResult{
		Name:        "Database",
		Status:      services.HealthStatusHealthy,
		Message:     "Database connection is healthy",
		LastChecked: time.Now(),
		Duration:    50 * time.Millisecond,
	}
}

func (m *MockHealthCheckService) CheckCacheHealth() services.HealthCheckResult {
	return services.HealthCheckResult{
		Name:        "Cache System",
		Status:      services.HealthStatusWarning,
		Message:     "High memory usage",
		LastChecked: time.Now(),
		Duration:    20 * time.Millisecond,
	}
}

func (m *MockHealthCheckService) CheckExternalAPIsHealth() services.HealthCheckResult {
	return services.HealthCheckResult{
		Name:        "External APIs",
		Status:      services.HealthStatusHealthy,
		Message:     "All external APIs are responding",
		LastChecked: time.Now(),
		Duration:    100 * time.Millisecond,
	}
}

func (m *MockHealthCheckService) CheckFileSystemHealth() services.HealthCheckResult {
	return services.HealthCheckResult{
		Name:        "File System",
		Status:      services.HealthStatusHealthy,
		Message:     "File system is accessible",
		LastChecked: time.Now(),
		Duration:    5 * time.Millisecond,
	}
}

func (m *MockHealthCheckService) CheckMemoryHealth() services.HealthCheckResult {
	return services.HealthCheckResult{
		Name:        "Memory",
		Status:      services.HealthStatusHealthy,
		Message:     "Memory usage is normal",
		LastChecked: time.Now(),
		Duration:    2 * time.Millisecond,
	}
}

func (m *MockHealthCheckService) GetAllHealthChecks() []services.HealthCheckResult {
	if m.checks != nil {
		return m.checks
	}

	return []services.HealthCheckResult{
		m.CheckAPIHealth(),
		m.CheckDatabaseHealth(),
		m.CheckCacheHealth(),
		m.CheckExternalAPIsHealth(),
		m.CheckFileSystemHealth(),
		m.CheckMemoryHealth(),
	}
}

func (m *MockHealthCheckService) GetOverallHealth() services.HealthStatus {
	checks := m.GetAllHealthChecks()

	hasCritical := false
	hasWarning := false

	for _, check := range checks {
		switch check.Status {
		case services.HealthStatusCritical:
			hasCritical = true
		case services.HealthStatusWarning:
			hasWarning = true
		}
	}

	if hasCritical {
		return services.HealthStatusCritical
	} else if hasWarning {
		return services.HealthStatusWarning
	}

	return services.HealthStatusHealthy
}

func TestHealthCheckHandler_GetHealth(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockHealthCheckService{}
	handler := NewHealthCheckHandler(logger, mockService)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.GetHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response HealthCheckResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response structure
	if response.Status == "" {
		t.Error("Status should not be empty")
	}
	if response.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}
	if response.Version == "" {
		t.Error("Version should not be empty")
	}
	if response.Environment == "" {
		t.Error("Environment should not be empty")
	}
	if len(response.Checks) == 0 {
		t.Error("Checks should not be empty")
	}
	if response.Summary.Total == 0 {
		t.Error("Summary total should not be zero")
	}
}

func TestHealthCheckHandler_GetHealth_CriticalStatus(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockHealthCheckService{
		checks: []services.HealthCheckResult{
			{
				Name:        "Critical Service",
				Status:      services.HealthStatusCritical,
				Message:     "Service is down",
				LastChecked: time.Now(),
				Duration:    100 * time.Millisecond,
			},
		},
	}
	handler := NewHealthCheckHandler(logger, mockService)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.GetHealth(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var response HealthCheckResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response.Status != "critical" {
		t.Errorf("Expected status 'critical', got '%s'", response.Status)
	}
	if response.Summary.Critical != 1 {
		t.Errorf("Expected 1 critical check, got %d", response.Summary.Critical)
	}
}

func TestHealthCheckHandler_GetHealthDetailed(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockHealthCheckService{}
	handler := NewHealthCheckHandler(logger, mockService)

	req := httptest.NewRequest("GET", "/health/detailed", nil)
	w := httptest.NewRecorder()

	handler.GetHealthDetailed(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify required fields
	requiredFields := []string{"timestamp", "version", "environment", "checks", "system", "dependencies"}
	for _, field := range requiredFields {
		if _, exists := response[field]; !exists {
			t.Errorf("Response should contain field '%s'", field)
		}
	}
}

func TestHealthCheckHandler_GetHealthLiveness(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockHealthCheckService{}
	handler := NewHealthCheckHandler(logger, mockService)

	req := httptest.NewRequest("GET", "/health/live", nil)
	w := httptest.NewRecorder()

	handler.GetHealthLiveness(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "alive" {
		t.Errorf("Expected status 'alive', got '%v'", response["status"])
	}
}

func TestHealthCheckHandler_GetHealthReadiness_Ready(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockHealthCheckService{}
	handler := NewHealthCheckHandler(logger, mockService)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()

	handler.GetHealthReadiness(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "ready" {
		t.Errorf("Expected status 'ready', got '%v'", response["status"])
	}
}

func TestHealthCheckHandler_GetHealthReadiness_NotReady(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockHealthCheckService{
		checks: []services.HealthCheckResult{
			{
				Name:        "Critical Service",
				Status:      services.HealthStatusCritical,
				Message:     "Service is down",
				LastChecked: time.Now(),
				Duration:    100 * time.Millisecond,
			},
		},
	}
	handler := NewHealthCheckHandler(logger, mockService)

	req := httptest.NewRequest("GET", "/health/ready", nil)
	w := httptest.NewRecorder()

	handler.GetHealthReadiness(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "not_ready" {
		t.Errorf("Expected status 'not_ready', got '%v'", response["status"])
	}
}

func TestHealthCheckHandler_GetHealthStartup(t *testing.T) {
	logger := zap.NewNop()
	mockService := &MockHealthCheckService{}
	handler := NewHealthCheckHandler(logger, mockService)

	req := httptest.NewRequest("GET", "/health/startup", nil)
	w := httptest.NewRecorder()

	handler.GetHealthStartup(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "started" {
		t.Errorf("Expected status 'started', got '%v'", response["status"])
	}
}

func TestHealthCheckResponse_Structure(t *testing.T) {
	response := HealthCheckResponse{
		Status:      "healthy",
		Timestamp:   time.Now(),
		Version:     "1.0.0",
		Environment: "test",
		Checks: []services.HealthCheckResult{
			{
				Name:        "Test Check",
				Status:      services.HealthStatusHealthy,
				Message:     "Test message",
				LastChecked: time.Now(),
				Duration:    10 * time.Millisecond,
			},
		},
		Summary: HealthCheckSummary{
			Total:    1,
			Healthy:  1,
			Warning:  0,
			Critical: 0,
		},
	}

	// Test JSON marshaling
	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var unmarshaled HealthCheckResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if unmarshaled.Status != response.Status {
		t.Errorf("Expected status %s, got %s", response.Status, unmarshaled.Status)
	}
	if unmarshaled.Version != response.Version {
		t.Errorf("Expected version %s, got %s", response.Version, unmarshaled.Version)
	}
	if unmarshaled.Summary.Total != response.Summary.Total {
		t.Errorf("Expected total %d, got %d", response.Summary.Total, unmarshaled.Summary.Total)
	}
}

func BenchmarkHealthCheckHandler_GetHealth(b *testing.B) {
	logger := zap.NewNop()
	mockService := &MockHealthCheckService{}
	handler := NewHealthCheckHandler(logger, mockService)

	req := httptest.NewRequest("GET", "/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.GetHealth(w, req)
	}
}

func BenchmarkHealthCheckHandler_GetHealthLiveness(b *testing.B) {
	logger := zap.NewNop()
	mockService := &MockHealthCheckService{}
	handler := NewHealthCheckHandler(logger, mockService)

	req := httptest.NewRequest("GET", "/health/live", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.GetHealthLiveness(w, req)
	}
}
