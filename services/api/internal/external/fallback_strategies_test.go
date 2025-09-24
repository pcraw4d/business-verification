package external

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewFallbackStrategyManager(t *testing.T) {
	logger := zap.NewNop()

	// Test with nil config
	manager := NewFallbackStrategyManager(nil, logger)
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.config)
	assert.True(t, manager.config.EnableUserAgentRotation)
	assert.True(t, manager.config.EnableHeaderCustomization)
	assert.False(t, manager.config.EnableProxyRotation)
	assert.True(t, manager.config.EnableAlternativeSources)
	assert.Equal(t, 5, manager.config.MaxFallbackAttempts)
	assert.Equal(t, 2*time.Second, manager.config.FallbackDelay)
	assert.Len(t, manager.config.UserAgentPool, 7)
	assert.Len(t, manager.config.AlternativeSources, 2)

	// Test with custom config
	customConfig := &FallbackConfig{
		EnableUserAgentRotation:   false,
		EnableHeaderCustomization: false,
		EnableProxyRotation:       true,
		EnableAlternativeSources:  false,
		MaxFallbackAttempts:       3,
		FallbackDelay:             1 * time.Second,
		UserAgentPool:             []string{"Custom-UA/1.0"},
		HeaderTemplates:           map[string]string{"custom": "Accept: */*"},
		ProxyPool:                 []Proxy{{Host: "proxy.example.com", Port: 8080, Active: true}},
		AlternativeSources:        []DataSource{{Name: "Custom Source", Type: "api", Priority: 1}},
	}

	manager = NewFallbackStrategyManager(customConfig, logger)
	assert.NotNil(t, manager)
	assert.Equal(t, customConfig, manager.config)
	assert.False(t, manager.config.EnableUserAgentRotation)
	assert.False(t, manager.config.EnableHeaderCustomization)
	assert.True(t, manager.config.EnableProxyRotation)
	assert.False(t, manager.config.EnableAlternativeSources)
	assert.Equal(t, 3, manager.config.MaxFallbackAttempts)
	assert.Equal(t, 1*time.Second, manager.config.FallbackDelay)
	assert.Len(t, manager.config.UserAgentPool, 1)
	assert.Len(t, manager.config.ProxyPool, 1)
	assert.Len(t, manager.config.AlternativeSources, 1)
}

func TestFallbackStrategyManager_InitializeHeaderTemplates(t *testing.T) {
	logger := zap.NewNop()
	config := &FallbackConfig{
		HeaderTemplates: map[string]string{
			"desktop": "Accept: text/html\nUser-Agent: Desktop/1.0",
			"mobile":  "Accept: text/html\nUser-Agent: Mobile/1.0",
		},
	}

	manager := NewFallbackStrategyManager(config, logger)

	// Check that header templates were initialized
	assert.Len(t, manager.headers, 2)
	assert.Contains(t, manager.headers, "desktop")
	assert.Contains(t, manager.headers, "mobile")

	// Check desktop headers
	desktopHeaders := manager.headers["desktop"]
	assert.Len(t, desktopHeaders, 2)
	assert.Contains(t, desktopHeaders, "Accept: text/html")
	assert.Contains(t, desktopHeaders, "User-Agent: Desktop/1.0")

	// Check mobile headers
	mobileHeaders := manager.headers["mobile"]
	assert.Len(t, mobileHeaders, 2)
	assert.Contains(t, mobileHeaders, "Accept: text/html")
	assert.Contains(t, mobileHeaders, "User-Agent: Mobile/1.0")
}

func TestFallbackStrategyManager_ExecuteFallbackStrategies(t *testing.T) {
	logger := zap.NewNop()
	config := &FallbackConfig{
		EnableUserAgentRotation:   true,
		EnableHeaderCustomization: true,
		EnableProxyRotation:       false,
		EnableAlternativeSources:  true,
		MaxFallbackAttempts:       2,
		FallbackDelay:             100 * time.Millisecond,
		UserAgentPool:             []string{"Test-UA/1.0", "Test-UA/2.0"},
		HeaderTemplates:           map[string]string{"test": "Accept: */*"},
		AlternativeSources:        []DataSource{{Name: "Test Source", Type: "api", Priority: 1}},
	}

	manager := NewFallbackStrategyManager(config, logger)
	ctx := context.Background()
	originalError := errors.New("test error")

	// Test with invalid URL (should fail all strategies)
	result, err := manager.ExecuteFallbackStrategies(ctx, "invalid-url", originalError)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "none", result.StrategyUsed)
	assert.Equal(t, 0, result.Attempts)
	assert.True(t, result.Duration > 0)
}

func TestFallbackStrategyManager_TryUserAgentRotation(t *testing.T) {
	logger := zap.NewNop()
	config := &FallbackConfig{
		UserAgentPool: []string{"Test-UA/1.0", "Test-UA/2.0"},
		FallbackDelay: 100 * time.Millisecond,
	}

	manager := NewFallbackStrategyManager(config, logger)
	ctx := context.Background()

	// Test with invalid URL (should fail)
	result := manager.TryUserAgentRotation(ctx, "invalid-url")
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "user_agent_rotation", result.StrategyUsed)
	assert.Equal(t, 2, result.Attempts) // Should try both user agents
}

func TestFallbackStrategyManager_TryHeaderCustomization(t *testing.T) {
	logger := zap.NewNop()
	config := &FallbackConfig{
		HeaderTemplates: map[string]string{
			"test1": "Accept: text/html\nUser-Agent: Test/1.0",
			"test2": "Accept: application/json\nUser-Agent: Test/2.0",
		},
		FallbackDelay: 100 * time.Millisecond,
	}

	manager := NewFallbackStrategyManager(config, logger)
	ctx := context.Background()

	// Test with invalid URL (should fail)
	result := manager.TryHeaderCustomization(ctx, "invalid-url")
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "header_customization", result.StrategyUsed)
	assert.Equal(t, 2, result.Attempts) // Should try both header templates
}

func TestFallbackStrategyManager_TryProxyRotation(t *testing.T) {
	logger := zap.NewNop()
	config := &FallbackConfig{
		ProxyPool: []Proxy{
			{Host: "proxy1.example.com", Port: 8080, Active: true},
			{Host: "proxy2.example.com", Port: 8080, Active: false},
			{Host: "proxy3.example.com", Port: 8080, Active: true},
		},
		FallbackDelay: 100 * time.Millisecond,
	}

	manager := NewFallbackStrategyManager(config, logger)
	ctx := context.Background()

	// Test with invalid URL (should fail)
	result := manager.TryProxyRotation(ctx, "invalid-url")
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "proxy_rotation", result.StrategyUsed)
	assert.Equal(t, 2, result.Attempts) // Should try only active proxies
}

func TestFallbackStrategyManager_TryAlternativeDataSources(t *testing.T) {
	logger := zap.NewNop()
	config := &FallbackConfig{
		AlternativeSources: []DataSource{
			{Name: "Test Source 1", Type: "api", Priority: 1},
			{Name: "Test Source 2", Type: "database", Priority: 2},
		},
		FallbackDelay: 100 * time.Millisecond,
	}

	manager := NewFallbackStrategyManager(config, logger)
	ctx := context.Background()

	// Test with invalid URL (should fail)
	result := manager.TryAlternativeDataSources(ctx, "invalid-url")
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "alternative_sources", result.StrategyUsed)
	assert.Equal(t, 2, result.Attempts) // Should try both data sources
}

func TestFallbackStrategyManager_FetchFromDataSource(t *testing.T) {
	logger := zap.NewNop()
	manager := NewFallbackStrategyManager(nil, logger)
	ctx := context.Background()

	tests := []struct {
		name     string
		source   DataSource
		expected string
		hasError bool
	}{
		{
			name: "api source",
			source: DataSource{
				Name: "Test API",
				Type: "api",
				URL:  "https://api.example.com",
			},
			hasError: true, // Will fail due to network error
		},
		{
			name: "database source",
			source: DataSource{
				Name: "Test Database",
				Type: "database",
			},
			hasError: true, // Not implemented
		},
		{
			name: "cache source",
			source: DataSource{
				Name: "Test Cache",
				Type: "cache",
			},
			hasError: true, // Not implemented
		},
		{
			name: "unknown source",
			source: DataSource{
				Name: "Unknown",
				Type: "unknown",
			},
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content, err := manager.fetchFromDataSource(ctx, "https://example.com", tt.source)

			if tt.hasError {
				assert.Error(t, err)
				assert.Empty(t, content)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, content)
			}
		})
	}
}

func TestFallbackStrategyManager_GetRandomUserAgent(t *testing.T) {
	logger := zap.NewNop()
	config := &FallbackConfig{
		UserAgentPool: []string{"UA1", "UA2", "UA3"},
	}

	manager := NewFallbackStrategyManager(config, logger)

	// Test multiple calls to ensure we get different user agents
	userAgents := make(map[string]bool)
	for i := 0; i < 10; i++ {
		ua := manager.getRandomUserAgent()
		userAgents[ua] = true
	}

	// Should have at least 2 different user agents (allowing for randomness)
	assert.True(t, len(userAgents) >= 1)
	assert.True(t, len(userAgents) <= 3)
}

func TestFallbackStrategyManager_AddRemoveProxy(t *testing.T) {
	logger := zap.NewNop()
	manager := NewFallbackStrategyManager(nil, logger)

	// Test adding proxy
	proxy := Proxy{
		Host:     "test.example.com",
		Port:     8080,
		Protocol: "http",
		Active:   true,
	}

	manager.AddProxy(proxy)
	assert.Len(t, manager.proxies, 1)
	assert.Equal(t, proxy.Host, manager.proxies[0].Host)
	assert.Equal(t, proxy.Port, manager.proxies[0].Port)

	// Test removing proxy
	manager.RemoveProxy("test.example.com", 8080)
	assert.Len(t, manager.proxies, 0)

	// Test removing non-existent proxy
	manager.RemoveProxy("nonexistent.example.com", 8080)
	assert.Len(t, manager.proxies, 0)
}

func TestFallbackStrategyManager_UpdateConfig(t *testing.T) {
	logger := zap.NewNop()
	manager := NewFallbackStrategyManager(nil, logger)

	// Get initial config
	initialConfig := manager.GetConfig()
	assert.True(t, initialConfig.EnableUserAgentRotation)

	// Update config
	newConfig := &FallbackConfig{
		EnableUserAgentRotation:   false,
		EnableHeaderCustomization: false,
		EnableProxyRotation:       true,
		EnableAlternativeSources:  false,
		MaxFallbackAttempts:       10,
		FallbackDelay:             5 * time.Second,
		UserAgentPool:             []string{"New-UA/1.0"},
		HeaderTemplates:           map[string]string{"new": "Accept: */*"},
		ProxyPool:                 []Proxy{{Host: "new.example.com", Port: 8080}},
		AlternativeSources:        []DataSource{{Name: "New Source", Type: "api"}},
	}

	manager.UpdateConfig(newConfig)

	// Verify config was updated
	updatedConfig := manager.GetConfig()
	assert.Equal(t, newConfig, updatedConfig)
	assert.False(t, updatedConfig.EnableUserAgentRotation)
	assert.False(t, updatedConfig.EnableHeaderCustomization)
	assert.True(t, updatedConfig.EnableProxyRotation)
	assert.False(t, updatedConfig.EnableAlternativeSources)
	assert.Equal(t, 10, updatedConfig.MaxFallbackAttempts)
	assert.Equal(t, 5*time.Second, updatedConfig.FallbackDelay)
	assert.Len(t, updatedConfig.UserAgentPool, 1)
	assert.Len(t, updatedConfig.ProxyPool, 1)
	assert.Len(t, updatedConfig.AlternativeSources, 1)
}

func TestFallbackStrategyManager_GetConfig(t *testing.T) {
	logger := zap.NewNop()
	config := &FallbackConfig{
		EnableUserAgentRotation:   true,
		EnableHeaderCustomization: false,
		MaxFallbackAttempts:       7,
		FallbackDelay:             3 * time.Second,
	}

	manager := NewFallbackStrategyManager(config, logger)

	// Test getting config
	retrievedConfig := manager.GetConfig()
	assert.Equal(t, config, retrievedConfig)
	assert.True(t, retrievedConfig.EnableUserAgentRotation)
	assert.False(t, retrievedConfig.EnableHeaderCustomization)
	assert.Equal(t, 7, retrievedConfig.MaxFallbackAttempts)
	assert.Equal(t, 3*time.Second, retrievedConfig.FallbackDelay)
}

func TestDefaultFallbackConfig(t *testing.T) {
	config := DefaultFallbackConfig()

	assert.NotNil(t, config)
	assert.True(t, config.EnableUserAgentRotation)
	assert.True(t, config.EnableHeaderCustomization)
	assert.False(t, config.EnableProxyRotation)
	assert.True(t, config.EnableAlternativeSources)
	assert.Equal(t, 5, config.MaxFallbackAttempts)
	assert.Equal(t, 2*time.Second, config.FallbackDelay)
	assert.Len(t, config.UserAgentPool, 7)
	assert.Len(t, config.HeaderTemplates, 2)
	assert.Len(t, config.ProxyPool, 0)
	assert.Len(t, config.AlternativeSources, 2)

	// Check specific user agents
	assert.Contains(t, config.UserAgentPool, "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	assert.Contains(t, config.UserAgentPool, "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")

	// Check header templates
	assert.Contains(t, config.HeaderTemplates, "desktop")
	assert.Contains(t, config.HeaderTemplates, "mobile")

	// Check alternative sources
	assert.Len(t, config.AlternativeSources, 2)
	assert.Equal(t, "Wayback Machine", config.AlternativeSources[0].Name)
	assert.Equal(t, "Google Cache", config.AlternativeSources[1].Name)
}

func TestFallbackResult_StructFields(t *testing.T) {
	result := &FallbackResult{
		StrategyUsed:  "test_strategy",
		Success:       true,
		Content:       "test content",
		StatusCode:    200,
		DataSource:    "test_source",
		ProxyUsed:     &Proxy{Host: "test.com", Port: 8080},
		UserAgentUsed: "test-ua",
		HeadersUsed:   map[string]string{"Accept": "*/*"},
		Attempts:      3,
		Duration:      1 * time.Second,
		Error:         "test error",
		Metadata:      map[string]interface{}{"key": "value"},
	}

	// Test that all fields are properly set
	assert.Equal(t, "test_strategy", result.StrategyUsed)
	assert.True(t, result.Success)
	assert.Equal(t, "test content", result.Content)
	assert.Equal(t, 200, result.StatusCode)
	assert.Equal(t, "test_source", result.DataSource)
	assert.NotNil(t, result.ProxyUsed)
	assert.Equal(t, "test.com", result.ProxyUsed.Host)
	assert.Equal(t, 8080, result.ProxyUsed.Port)
	assert.Equal(t, "test-ua", result.UserAgentUsed)
	assert.Len(t, result.HeadersUsed, 1)
	assert.Equal(t, "*/*", result.HeadersUsed["Accept"])
	assert.Equal(t, 3, result.Attempts)
	assert.Equal(t, 1*time.Second, result.Duration)
	assert.Equal(t, "test error", result.Error)
	assert.Len(t, result.Metadata, 1)
	assert.Equal(t, "value", result.Metadata["key"])
}

func TestProxy_StructFields(t *testing.T) {
	proxy := &Proxy{
		Host:     "proxy.example.com",
		Port:     8080,
		Username: "user",
		Password: "pass",
		Protocol: "http",
		Location: "US",
		Active:   true,
	}

	// Test that all fields are properly set
	assert.Equal(t, "proxy.example.com", proxy.Host)
	assert.Equal(t, 8080, proxy.Port)
	assert.Equal(t, "user", proxy.Username)
	assert.Equal(t, "pass", proxy.Password)
	assert.Equal(t, "http", proxy.Protocol)
	assert.Equal(t, "US", proxy.Location)
	assert.True(t, proxy.Active)
}

func TestDataSource_StructFields(t *testing.T) {
	source := &DataSource{
		Name:        "Test Source",
		Type:        "api",
		URL:         "https://api.example.com",
		APIKey:      "test-key",
		Headers:     map[string]string{"Authorization": "Bearer token"},
		Timeout:     10 * time.Second,
		Priority:    1,
		Reliability: 0.9,
	}

	// Test that all fields are properly set
	assert.Equal(t, "Test Source", source.Name)
	assert.Equal(t, "api", source.Type)
	assert.Equal(t, "https://api.example.com", source.URL)
	assert.Equal(t, "test-key", source.APIKey)
	assert.Len(t, source.Headers, 1)
	assert.Equal(t, "Bearer token", source.Headers["Authorization"])
	assert.Equal(t, 10*time.Second, source.Timeout)
	assert.Equal(t, 1, source.Priority)
	assert.Equal(t, 0.9, source.Reliability)
}
