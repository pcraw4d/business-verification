package services

import (
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestHealthCheckService_CheckAPIHealth(t *testing.T) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	result := service.CheckAPIHealth()

	if result.Name != "API Server" {
		t.Errorf("Expected name 'API Server', got '%s'", result.Name)
	}

	if result.Status == "" {
		t.Error("Status should not be empty")
	}

	if result.Message == "" {
		t.Error("Message should not be empty")
	}

	if result.Duration < 0 {
		t.Error("Duration should be non-negative")
	}

	if result.LastChecked.IsZero() {
		t.Error("LastChecked should not be zero")
	}
}

func TestHealthCheckService_CheckDatabaseHealth(t *testing.T) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	result := service.CheckDatabaseHealth()

	if result.Name != "Database" {
		t.Errorf("Expected name 'Database', got '%s'", result.Name)
	}

	if result.Status != HealthStatusCritical {
		t.Errorf("Expected status 'critical' for nil database, got '%s'", result.Status)
	}

	if result.Message == "" {
		t.Error("Message should not be empty")
	}
}

func TestHealthCheckService_CheckCacheHealth(t *testing.T) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	result := service.CheckCacheHealth()

	if result.Name != "Cache System" {
		t.Errorf("Expected name 'Cache System', got '%s'", result.Name)
	}

	if result.Status == "" {
		t.Error("Status should not be empty")
	}

	if result.Message == "" {
		t.Error("Message should not be empty")
	}

	if result.Duration < 0 {
		t.Error("Duration should be non-negative")
	}
}

func TestHealthCheckService_CheckExternalAPIsHealth(t *testing.T) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	result := service.CheckExternalAPIsHealth()

	if result.Name != "External APIs" {
		t.Errorf("Expected name 'External APIs', got '%s'", result.Name)
	}

	if result.Status == "" {
		t.Error("Status should not be empty")
	}

	if result.Message == "" {
		t.Error("Message should not be empty")
	}

	if result.Duration < 0 {
		t.Error("Duration should be non-negative")
	}
}

func TestHealthCheckService_CheckFileSystemHealth(t *testing.T) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	result := service.CheckFileSystemHealth()

	if result.Name != "File System" {
		t.Errorf("Expected name 'File System', got '%s'", result.Name)
	}

	if result.Status == "" {
		t.Error("Status should not be empty")
	}

	if result.Message == "" {
		t.Error("Message should not be empty")
	}

	if result.Duration < 0 {
		t.Error("Duration should be non-negative")
	}
}

func TestHealthCheckService_CheckMemoryHealth(t *testing.T) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	result := service.CheckMemoryHealth()

	if result.Name != "Memory" {
		t.Errorf("Expected name 'Memory', got '%s'", result.Name)
	}

	if result.Status == "" {
		t.Error("Status should not be empty")
	}

	if result.Message == "" {
		t.Error("Message should not be empty")
	}

	if result.Duration < 0 {
		t.Error("Duration should be non-negative")
	}

	if result.Details == nil {
		t.Error("Details should not be nil")
	}

	// Check that memory details are present
	if _, exists := result.Details["alloc_mb"]; !exists {
		t.Error("Details should contain 'alloc_mb'")
	}
	if _, exists := result.Details["sys_mb"]; !exists {
		t.Error("Details should contain 'sys_mb'")
	}
	if _, exists := result.Details["usage_percent"]; !exists {
		t.Error("Details should contain 'usage_percent'")
	}
}

func TestHealthCheckService_GetAllHealthChecks(t *testing.T) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	results := service.GetAllHealthChecks()

	expectedChecks := []string{
		"API Server",
		"Database",
		"Cache System",
		"External APIs",
		"File System",
		"Memory",
	}

	if len(results) != len(expectedChecks) {
		t.Errorf("Expected %d health checks, got %d", len(expectedChecks), len(results))
	}

	// Verify all expected checks are present
	checkNames := make(map[string]bool)
	for _, result := range results {
		checkNames[result.Name] = true
	}

	for _, expectedName := range expectedChecks {
		if !checkNames[expectedName] {
			t.Errorf("Expected health check '%s' not found", expectedName)
		}
	}
}

func TestHealthCheckService_GetOverallHealth(t *testing.T) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	overallHealth := service.GetOverallHealth()

	if overallHealth == "" {
		t.Error("Overall health should not be empty")
	}

	// Should be one of the valid health statuses
	validStatuses := []HealthStatus{
		HealthStatusHealthy,
		HealthStatusWarning,
		HealthStatusCritical,
	}

	isValid := false
	for _, status := range validStatuses {
		if overallHealth == status {
			isValid = true
			break
		}
	}

	if !isValid {
		t.Errorf("Overall health '%s' is not a valid status", overallHealth)
	}
}

func TestHealthCheckResult_Structure(t *testing.T) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	result := service.CheckAPIHealth()

	// Test that all required fields are populated
	if result.Name == "" {
		t.Error("Name should not be empty")
	}

	if result.Status == "" {
		t.Error("Status should not be empty")
	}

	if result.Message == "" {
		t.Error("Message should not be empty")
	}

	if result.LastChecked.IsZero() {
		t.Error("LastChecked should not be zero")
	}

	if result.Duration < 0 {
		t.Error("Duration should be non-negative")
	}
}

func TestHealthCheckService_Performance(t *testing.T) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	start := time.Now()
	service.GetAllHealthChecks()
	duration := time.Since(start)

	// Health checks should complete quickly (less than 1 second)
	if duration > time.Second {
		t.Errorf("Health checks took too long: %v", duration)
	}
}

func BenchmarkHealthCheckService_GetAllHealthChecks(b *testing.B) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.GetAllHealthChecks()
	}
}

func BenchmarkHealthCheckService_CheckMemoryHealth(b *testing.B) {
	logger := zap.NewNop()
	service := NewHealthCheckService(logger, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.CheckMemoryHealth()
	}
}
