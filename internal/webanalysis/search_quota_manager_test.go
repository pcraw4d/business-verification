package webanalysis

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewSearchQuotaManager(t *testing.T) {
	manager := NewSearchQuotaManager()

	if manager == nil {
		t.Fatal("Expected manager to be created, got nil")
	}

	if len(manager.engines) == 0 {
		t.Error("Expected engines to be initialized")
	}

	if manager.globalQuota == nil {
		t.Error("Expected global quota to be initialized")
	}

	// Check that default engines are created
	expectedEngines := []string{"google", "bing", "duckduckgo"}
	for _, engineName := range expectedEngines {
		if _, exists := manager.engines[engineName]; !exists {
			t.Errorf("Expected engine %s to be initialized", engineName)
		}
	}

	config := manager.GetConfig()
	if !config.EnableQuotaManagement {
		t.Error("Expected EnableQuotaManagement true, got false")
	}

	if config.AlertThreshold != 0.8 {
		t.Errorf("Expected AlertThreshold 0.8, got %f", config.AlertThreshold)
	}
}

func TestSearchQuotaManager_RequestQuota_Success(t *testing.T) {
	manager := NewSearchQuotaManager()

	req := &QuotaRequest{
		EngineName: "google",
		RequestID:  "test-request-1",
		Priority:   1,
		Timeout:    time.Second * 30,
		Metadata:   map[string]string{"test": "value"},
	}

	ctx := context.Background()
	response, err := manager.RequestQuota(ctx, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if !response.IsAllowed {
		t.Error("Expected request to be allowed")
	}

	if response.EngineName != "google" {
		t.Errorf("Expected engine name 'google', got %s", response.EngineName)
	}

	if response.RequestID != "test-request-1" {
		t.Errorf("Expected request ID 'test-request-1', got %s", response.RequestID)
	}

	if response.QuotaRemaining <= 0 {
		t.Error("Expected quota remaining to be positive")
	}

	if response.QuotaUsed != 1 {
		t.Errorf("Expected quota used 1, got %d", response.QuotaUsed)
	}
}

func TestSearchQuotaManager_RequestQuota_EngineNotFound(t *testing.T) {
	manager := NewSearchQuotaManager()

	req := &QuotaRequest{
		EngineName: "nonexistent-engine",
		RequestID:  "test-request-2",
		Priority:   1,
		Timeout:    time.Second * 30,
	}

	ctx := context.Background()
	response, err := manager.RequestQuota(ctx, req)

	if err == nil {
		t.Error("Expected error for nonexistent engine")
	}

	if response != nil {
		t.Error("Expected nil response for error")
	}

	expectedError := "engine nonexistent-engine not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSearchQuotaManager_RequestQuota_EngineDisabled(t *testing.T) {
	manager := NewSearchQuotaManager()

	// Disable the google engine
	err := manager.DisableEngine("google")
	if err != nil {
		t.Fatalf("Failed to disable engine: %v", err)
	}

	req := &QuotaRequest{
		EngineName: "google",
		RequestID:  "test-request-3",
		Priority:   1,
		Timeout:    time.Second * 30,
	}

	ctx := context.Background()
	response, err := manager.RequestQuota(ctx, req)

	if err == nil {
		t.Error("Expected error for disabled engine")
	}

	if response != nil {
		t.Error("Expected nil response for error")
	}

	expectedError := "engine google is disabled"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSearchQuotaManager_RequestQuota_QuotaExceeded(t *testing.T) {
	manager := NewSearchQuotaManager()

	// Exhaust the quota for the duckduckgo engine (which has low limits)
	for i := 0; i < 2; i++ { // Exceed the minute quota of 1
		req := &QuotaRequest{
			EngineName: "duckduckgo",
			RequestID:  fmt.Sprintf("test-request-%d", i),
			Priority:   1,
			Timeout:    time.Second * 30,
		}

		ctx := context.Background()
		response, err := manager.RequestQuota(ctx, req)

		if i == 0 {
			// First request should succeed
			if err != nil {
				t.Fatalf("Expected no error for first request, got %v", err)
			}
			if !response.IsAllowed {
				t.Error("Expected first request to be allowed")
			}
		} else {
			// Second request should be denied due to quota exceeded
			if err != nil {
				t.Fatalf("Expected no error for second request, got %v", err)
			}
			if response.IsAllowed {
				t.Error("Expected second request to be denied")
			}
			if len(response.Errors) == 0 {
				t.Error("Expected error message for quota exceeded")
			}
		}
	}
}

func TestSearchQuotaManager_ReleaseQuota(t *testing.T) {
	manager := NewSearchQuotaManager()

	// Request quota first
	req := &QuotaRequest{
		EngineName: "google",
		RequestID:  "test-request-4",
		Priority:   1,
		Timeout:    time.Second * 30,
	}

	ctx := context.Background()
	response, err := manager.RequestQuota(ctx, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !response.IsAllowed {
		t.Error("Expected request to be allowed")
	}

	// Check that concurrent requests increased
	status := manager.GetQuotaStatus()
	engines := status["engines"].(map[string]interface{})
	googleStatus := engines["google"].(map[string]interface{})
	concurrentRequests := googleStatus["concurrent_requests"].(int)

	if concurrentRequests != 1 {
		t.Errorf("Expected concurrent requests 1, got %d", concurrentRequests)
	}

	// Release quota
	err = manager.ReleaseQuota("google", "test-request-4")
	if err != nil {
		t.Fatalf("Expected no error releasing quota, got %v", err)
	}

	// Check that concurrent requests decreased
	status = manager.GetQuotaStatus()
	engines = status["engines"].(map[string]interface{})
	googleStatus = engines["google"].(map[string]interface{})
	concurrentRequests = googleStatus["concurrent_requests"].(int)

	if concurrentRequests != 0 {
		t.Errorf("Expected concurrent requests 0, got %d", concurrentRequests)
	}
}

func TestSearchQuotaManager_GetQuotaStatus(t *testing.T) {
	manager := NewSearchQuotaManager()

	status := manager.GetQuotaStatus()

	if status == nil {
		t.Fatal("Expected status, got nil")
	}

	// Check global quota status
	global := status["global"].(map[string]interface{})
	if global["total_daily_quota_limit"] == nil {
		t.Error("Expected total_daily_quota_limit in global status")
	}

	// Check engines status
	engines := status["engines"].(map[string]interface{})
	if len(engines) != 3 {
		t.Errorf("Expected 3 engines, got %d", len(engines))
	}

	// Check google engine status
	googleStatus := engines["google"].(map[string]interface{})
	if googleStatus["engine_name"] != "google" {
		t.Errorf("Expected engine name 'google', got %s", googleStatus["engine_name"])
	}

	if googleStatus["daily_quota_limit"] != 10000 {
		t.Errorf("Expected daily quota limit 10000, got %v", googleStatus["daily_quota_limit"])
	}

	if googleStatus["priority"] != 1 {
		t.Errorf("Expected priority 1, got %v", googleStatus["priority"])
	}

	if !googleStatus["is_enabled"].(bool) {
		t.Error("Expected google engine to be enabled")
	}
}

func TestSearchQuotaManager_AddEngine(t *testing.T) {
	manager := NewSearchQuotaManager()

	engineInfo := &EngineQuotaInfo{
		EngineName:            "custom-engine",
		DailyQuotaLimit:       5000,
		HourlyQuotaLimit:      500,
		MinuteQuotaLimit:      5,
		MaxConcurrentRequests: 3,
		IsEnabled:             true,
		Priority:              4,
		FallbackEngines:       []string{"google", "bing"},
	}

	err := manager.AddEngine("custom-engine", engineInfo)
	if err != nil {
		t.Fatalf("Expected no error adding engine, got %v", err)
	}

	// Check that engine was added
	status := manager.GetQuotaStatus()
	engines := status["engines"].(map[string]interface{})

	if _, exists := engines["custom-engine"]; !exists {
		t.Error("Expected custom-engine to be added")
	}

	customStatus := engines["custom-engine"].(map[string]interface{})
	if customStatus["daily_quota_limit"] != 5000 {
		t.Errorf("Expected daily quota limit 5000, got %v", customStatus["daily_quota_limit"])
	}

	if customStatus["priority"] != 4 {
		t.Errorf("Expected priority 4, got %v", customStatus["priority"])
	}
}

func TestSearchQuotaManager_AddEngine_Duplicate(t *testing.T) {
	manager := NewSearchQuotaManager()

	engineInfo := &EngineQuotaInfo{
		EngineName:            "google", // Already exists
		DailyQuotaLimit:       5000,
		HourlyQuotaLimit:      500,
		MinuteQuotaLimit:      5,
		MaxConcurrentRequests: 3,
		IsEnabled:             true,
		Priority:              4,
		FallbackEngines:       []string{},
	}

	err := manager.AddEngine("google", engineInfo)
	if err == nil {
		t.Error("Expected error for duplicate engine")
	}

	expectedError := "engine google already exists"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSearchQuotaManager_UpdateEngine(t *testing.T) {
	manager := NewSearchQuotaManager()

	engineInfo := &EngineQuotaInfo{
		EngineName:            "google",
		DailyQuotaLimit:       15000, // Updated limit
		HourlyQuotaLimit:      1500,
		MinuteQuotaLimit:      15,
		MaxConcurrentRequests: 8,
		IsEnabled:             true,
		Priority:              1,
		FallbackEngines:       []string{"bing", "duckduckgo"},
	}

	err := manager.UpdateEngine("google", engineInfo)
	if err != nil {
		t.Fatalf("Expected no error updating engine, got %v", err)
	}

	// Check that engine was updated
	status := manager.GetQuotaStatus()
	engines := status["engines"].(map[string]interface{})
	googleStatus := engines["google"].(map[string]interface{})

	if googleStatus["daily_quota_limit"] != 15000 {
		t.Errorf("Expected updated daily quota limit 15000, got %v", googleStatus["daily_quota_limit"])
	}

	if googleStatus["max_concurrent_requests"] != 8 {
		t.Errorf("Expected updated max concurrent requests 8, got %v", googleStatus["max_concurrent_requests"])
	}
}

func TestSearchQuotaManager_UpdateEngine_NotFound(t *testing.T) {
	manager := NewSearchQuotaManager()

	engineInfo := &EngineQuotaInfo{
		EngineName:            "nonexistent-engine",
		DailyQuotaLimit:       5000,
		HourlyQuotaLimit:      500,
		MinuteQuotaLimit:      5,
		MaxConcurrentRequests: 3,
		IsEnabled:             true,
		Priority:              4,
		FallbackEngines:       []string{},
	}

	err := manager.UpdateEngine("nonexistent-engine", engineInfo)
	if err == nil {
		t.Error("Expected error for nonexistent engine")
	}

	expectedError := "engine nonexistent-engine not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSearchQuotaManager_RemoveEngine(t *testing.T) {
	manager := NewSearchQuotaManager()

	err := manager.RemoveEngine("duckduckgo")
	if err != nil {
		t.Fatalf("Expected no error removing engine, got %v", err)
	}

	// Check that engine was removed
	status := manager.GetQuotaStatus()
	engines := status["engines"].(map[string]interface{})

	if _, exists := engines["duckduckgo"]; exists {
		t.Error("Expected duckduckgo engine to be removed")
	}

	// Check that other engines still exist
	if _, exists := engines["google"]; !exists {
		t.Error("Expected google engine to still exist")
	}

	if _, exists := engines["bing"]; !exists {
		t.Error("Expected bing engine to still exist")
	}
}

func TestSearchQuotaManager_RemoveEngine_NotFound(t *testing.T) {
	manager := NewSearchQuotaManager()

	err := manager.RemoveEngine("nonexistent-engine")
	if err == nil {
		t.Error("Expected error for nonexistent engine")
	}

	expectedError := "engine nonexistent-engine not found"
	if err.Error() != expectedError {
		t.Errorf("Expected error '%s', got '%s'", expectedError, err.Error())
	}
}

func TestSearchQuotaManager_EnableDisableEngine(t *testing.T) {
	manager := NewSearchQuotaManager()

	// Disable engine
	err := manager.DisableEngine("google")
	if err != nil {
		t.Fatalf("Expected no error disabling engine, got %v", err)
	}

	// Check that engine is disabled
	status := manager.GetQuotaStatus()
	engines := status["engines"].(map[string]interface{})
	googleStatus := engines["google"].(map[string]interface{})

	if googleStatus["is_enabled"].(bool) {
		t.Error("Expected google engine to be disabled")
	}

	// Enable engine
	err = manager.EnableEngine("google")
	if err != nil {
		t.Fatalf("Expected no error enabling engine, got %v", err)
	}

	// Check that engine is enabled
	status = manager.GetQuotaStatus()
	engines = status["engines"].(map[string]interface{})
	googleStatus = engines["google"].(map[string]interface{})

	if !googleStatus["is_enabled"].(bool) {
		t.Error("Expected google engine to be enabled")
	}
}

func TestSearchQuotaManager_ResetQuotas(t *testing.T) {
	manager := NewSearchQuotaManager()

	// Use some quota first
	req := &QuotaRequest{
		EngineName: "google",
		RequestID:  "test-request-5",
		Priority:   1,
		Timeout:    time.Second * 30,
	}

	ctx := context.Background()
	response, err := manager.RequestQuota(ctx, req)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !response.IsAllowed {
		t.Error("Expected request to be allowed")
	}

	// Check that quota was used
	status := manager.GetQuotaStatus()
	engines := status["engines"].(map[string]interface{})
	googleStatus := engines["google"].(map[string]interface{})
	quotaUsed := googleStatus["daily_quota_used"].(int)

	if quotaUsed != 1 {
		t.Errorf("Expected quota used 1, got %d", quotaUsed)
	}

	// Reset quotas
	manager.ResetQuotas()

	// Check that quotas were reset
	status = manager.GetQuotaStatus()
	engines = status["engines"].(map[string]interface{})
	googleStatus = engines["google"].(map[string]interface{})
	quotaUsed = googleStatus["daily_quota_used"].(int)

	if quotaUsed != 0 {
		t.Errorf("Expected quota used 0 after reset, got %d", quotaUsed)
	}

	concurrentRequests := googleStatus["concurrent_requests"].(int)
	if concurrentRequests != 0 {
		t.Errorf("Expected concurrent requests 0 after reset, got %d", concurrentRequests)
	}
}

func TestSearchQuotaManager_GetAvailableEngines(t *testing.T) {
	manager := NewSearchQuotaManager()

	availableEngines := manager.GetAvailableEngines()

	if len(availableEngines) != 3 {
		t.Errorf("Expected 3 available engines, got %d", len(availableEngines))
	}

	// Check that all default engines are available
	expectedEngines := map[string]bool{"google": true, "bing": true, "duckduckgo": true}
	for _, engineName := range availableEngines {
		if !expectedEngines[engineName] {
			t.Errorf("Unexpected engine in available list: %s", engineName)
		}
	}

	// Disable an engine and check that it's not available
	err := manager.DisableEngine("google")
	if err != nil {
		t.Fatalf("Failed to disable engine: %v", err)
	}

	availableEngines = manager.GetAvailableEngines()
	if len(availableEngines) != 2 {
		t.Errorf("Expected 2 available engines after disabling google, got %d", len(availableEngines))
	}

	for _, engineName := range availableEngines {
		if engineName == "google" {
			t.Error("Expected google engine to not be available after disabling")
		}
	}
}

func TestSearchQuotaManager_GetFallbackEngine(t *testing.T) {
	manager := NewSearchQuotaManager()

	// Test fallback for google (should return bing)
	fallback := manager.GetFallbackEngine("google")
	if fallback != "bing" {
		t.Errorf("Expected fallback engine 'bing' for google, got %s", fallback)
	}

	// Test fallback for bing (should return duckduckgo)
	fallback = manager.GetFallbackEngine("bing")
	if fallback != "duckduckgo" {
		t.Errorf("Expected fallback engine 'duckduckgo' for bing, got %s", fallback)
	}

	// Test fallback for duckduckgo (should return empty string)
	fallback = manager.GetFallbackEngine("duckduckgo")
	if fallback != "" {
		t.Errorf("Expected no fallback engine for duckduckgo, got %s", fallback)
	}

	// Test fallback for nonexistent engine
	fallback = manager.GetFallbackEngine("nonexistent-engine")
	if fallback != "" {
		t.Errorf("Expected no fallback engine for nonexistent engine, got %s", fallback)
	}
}

func TestSearchQuotaManager_UpdateConfig(t *testing.T) {
	manager := NewSearchQuotaManager()

	newConfig := QuotaManagerConfig{
		EnableQuotaManagement: false,
		EnableRateLimiting:    false,
		EnableQuotaTracking:   false,
		EnableQuotaAlerts:     false,
		QuotaResetTime:        time.Now().Add(24 * time.Hour),
		QuotaResetInterval:    48 * time.Hour,
		MaxConcurrentRequests: 20,
		RequestTimeout:        time.Second * 60,
		AlertThreshold:        0.9,
		RetryDelay:            time.Second * 5,
		MaxRetries:            5,
	}

	manager.UpdateConfig(newConfig)

	updatedConfig := manager.GetConfig()

	if updatedConfig.EnableQuotaManagement {
		t.Error("Expected EnableQuotaManagement false, got true")
	}

	if updatedConfig.EnableRateLimiting {
		t.Error("Expected EnableRateLimiting false, got true")
	}

	if updatedConfig.EnableQuotaTracking {
		t.Error("Expected EnableQuotaTracking false, got true")
	}

	if updatedConfig.EnableQuotaAlerts {
		t.Error("Expected EnableQuotaAlerts false, got true")
	}

	if updatedConfig.MaxConcurrentRequests != 20 {
		t.Errorf("Expected MaxConcurrentRequests 20, got %d", updatedConfig.MaxConcurrentRequests)
	}

	if updatedConfig.RequestTimeout != 60*time.Second {
		t.Errorf("Expected RequestTimeout 60s, got %v", updatedConfig.RequestTimeout)
	}

	if updatedConfig.AlertThreshold != 0.9 {
		t.Errorf("Expected AlertThreshold 0.9, got %f", updatedConfig.AlertThreshold)
	}

	if updatedConfig.RetryDelay != 5*time.Second {
		t.Errorf("Expected RetryDelay 5s, got %v", updatedConfig.RetryDelay)
	}

	if updatedConfig.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries 5, got %d", updatedConfig.MaxRetries)
	}
}
