package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

func TestMonitoringHandler_GetMetrics(t *testing.T) {
	logger := zap.NewNop()
	handler := NewMonitoringHandler(logger)

	req := httptest.NewRequest("GET", "/api/v3/monitoring/metrics", nil)
	w := httptest.NewRecorder()

	handler.GetMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var metrics DashboardMetrics
	if err := json.NewDecoder(w.Body).Decode(&metrics); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify required fields are present
	if metrics.RequestRate < 0 {
		t.Error("RequestRate should be non-negative")
	}
	if metrics.ResponseTime < 0 {
		t.Error("ResponseTime should be non-negative")
	}
	if metrics.ErrorRate < 0 {
		t.Error("ErrorRate should be non-negative")
	}
	if metrics.ActiveUsers < 0 {
		t.Error("ActiveUsers should be non-negative")
	}
	if metrics.MemoryUsage < 0 {
		t.Error("MemoryUsage should be non-negative")
	}
	if metrics.CPUUsage < 0 {
		t.Error("CPUUsage should be non-negative")
	}
	if metrics.Timestamp <= 0 {
		t.Error("Timestamp should be positive")
	}
}

func TestMonitoringHandler_GetAlerts(t *testing.T) {
	logger := zap.NewNop()
	handler := NewMonitoringHandler(logger)

	req := httptest.NewRequest("GET", "/api/v3/monitoring/alerts", nil)
	w := httptest.NewRecorder()

	handler.GetAlerts(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var alerts []Alert
	if err := json.NewDecoder(w.Body).Decode(&alerts); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify alerts structure
	for _, alert := range alerts {
		if alert.ID == "" {
			t.Error("Alert ID should not be empty")
		}
		if alert.Title == "" {
			t.Error("Alert Title should not be empty")
		}
		if alert.Description == "" {
			t.Error("Alert Description should not be empty")
		}
		if alert.Severity == "" {
			t.Error("Alert Severity should not be empty")
		}
		if alert.Status == "" {
			t.Error("Alert Status should not be empty")
		}
		if alert.Timestamp <= 0 {
			t.Error("Alert Timestamp should be positive")
		}
	}
}

func TestMonitoringHandler_GetHealthChecks(t *testing.T) {
	logger := zap.NewNop()
	handler := NewMonitoringHandler(logger)

	req := httptest.NewRequest("GET", "/api/v3/monitoring/health", nil)
	w := httptest.NewRecorder()

	handler.GetHealthChecks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var healthChecks []HealthCheck
	if err := json.NewDecoder(w.Body).Decode(&healthChecks); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify health checks structure
	expectedChecks := []string{
		"API Server",
		"Database",
		"Redis Cache",
		"External APIs",
		"File System",
		"Memory",
	}

	if len(healthChecks) != len(expectedChecks) {
		t.Errorf("Expected %d health checks, got %d", len(expectedChecks), len(healthChecks))
	}

	for _, check := range healthChecks {
		if check.Name == "" {
			t.Error("Health check Name should not be empty")
		}
		if check.Status == "" {
			t.Error("Health check Status should not be empty")
		}
		if check.Message == "" {
			t.Error("Health check Message should not be empty")
		}
		if check.LastChecked <= 0 {
			t.Error("Health check LastChecked should be positive")
		}
		if check.Duration < 0 {
			t.Error("Health check Duration should be non-negative")
		}
	}
}

func TestMonitoringHandler_ContentType(t *testing.T) {
	logger := zap.NewNop()
	handler := NewMonitoringHandler(logger)

	tests := []struct {
		name     string
		handler  func(http.ResponseWriter, *http.Request)
		endpoint string
	}{
		{"GetMetrics", handler.GetMetrics, "/metrics"},
		{"GetAlerts", handler.GetAlerts, "/alerts"},
		{"GetHealthChecks", handler.GetHealthChecks, "/health"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.endpoint, nil)
			w := httptest.NewRecorder()

			tt.handler(w, req)

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Expected Content-Type 'application/json', got '%s'", contentType)
			}
		})
	}
}

func TestMonitoringHandler_ErrorHandling(t *testing.T) {
	logger := zap.NewNop()
	handler := NewMonitoringHandler(logger)

	// Test with invalid request method
	req := httptest.NewRequest("POST", "/api/v3/monitoring/metrics", nil)
	w := httptest.NewRecorder()

	handler.GetMetrics(w, req)

	// Should still return 200 OK for GET handler
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func BenchmarkMonitoringHandler_GetMetrics(b *testing.B) {
	logger := zap.NewNop()
	handler := NewMonitoringHandler(logger)

	req := httptest.NewRequest("GET", "/api/v3/monitoring/metrics", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.GetMetrics(w, req)
	}
}

func BenchmarkMonitoringHandler_GetAlerts(b *testing.B) {
	logger := zap.NewNop()
	handler := NewMonitoringHandler(logger)

	req := httptest.NewRequest("GET", "/api/v3/monitoring/alerts", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.GetAlerts(w, req)
	}
}

func BenchmarkMonitoringHandler_GetHealthChecks(b *testing.B) {
	logger := zap.NewNop()
	handler := NewMonitoringHandler(logger)

	req := httptest.NewRequest("GET", "/api/v3/monitoring/health", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		handler.GetHealthChecks(w, req)
	}
}
