package webanalysis

import (
	"fmt"
	"testing"
	"time"
)

func TestEnhancedProxyManagerCreation(t *testing.T) {
	config := EnhancedProxyConfig{
		MaxProxies:             50,
		HealthCheckInterval:    time.Minute * 2,
		MaxFailures:            2,
		MinUptime:              0.9,
		MaxLatency:             time.Second * 5,
		RotationInterval:       time.Minute * 15,
		GeographicDistribution: true,
		LoadBalancing:          true,
		RateLimiting:           true,
		BotDetectionEvasion:    true,
		SSLVerification:        true,
		ConcurrentLimit:        5,
		RequestTimeout:         time.Second * 20,
		RetryAttempts:          2,
		BackoffMultiplier:      1.5,
	}

	epm := NewEnhancedProxyManager(config)

	if epm == nil {
		t.Fatal("EnhancedProxyManager should not be nil")
	}

	if epm.config.MaxProxies != 50 {
		t.Errorf("Expected MaxProxies to be 50, got %d", epm.config.MaxProxies)
	}

	if epm.healthChecker == nil {
		t.Error("HealthChecker should be initialized")
	}

	if epm.rateLimiter == nil {
		t.Error("RateLimiter should be initialized")
	}

	if epm.rotationEngine == nil {
		t.Error("RotationEngine should be initialized")
	}
}

func TestEnhancedProxyManagerAddProxy(t *testing.T) {
	epm := NewEnhancedProxyManager(EnhancedProxyConfig{})

	proxy := &EnhancedProxy{
		IP:              "192.168.1.100",
		Port:            8080,
		Protocol:        "http",
		Region:          "us-east",
		Country:         "US",
		City:            "New York",
		Provider:        "test-provider",
		Health:          true,
		SSLSupport:      true,
		AnonymityLevel:  "anonymous",
		ConcurrentLimit: 10,
		Capabilities:    map[string]bool{"ssl": true, "https": true},
		Headers:         map[string]string{"User-Agent": "test-agent"},
		UserAgents:      []string{"Mozilla/5.0 (Test)"},
	}

	err := epm.AddProxy(proxy)
	if err != nil {
		t.Fatalf("Failed to add proxy: %v", err)
	}

	if proxy.ID == "" {
		t.Error("Proxy ID should be auto-generated")
	}

	// Test adding duplicate proxy
	err = epm.AddProxy(proxy)
	if err == nil {
		t.Error("Should not allow adding duplicate proxy")
	}
}

func TestEnhancedProxyManagerGetProxy(t *testing.T) {
	epm := NewEnhancedProxyManager(EnhancedProxyConfig{
		MaxProxies:   10,
		RateLimiting: true,
	})

	// Add multiple proxies
	for i := 0; i < 3; i++ {
		proxy := &EnhancedProxy{
			IP:              fmt.Sprintf("192.168.1.%d", 100+i),
			Port:            8080 + i,
			Protocol:        "http",
			Region:          "us-east",
			Health:          true,
			SSLSupport:      true,
			AnonymityLevel:  "anonymous",
			ConcurrentLimit: 5,
			Latency:         time.Millisecond * 100,
			Uptime:          0.95,
		}
		epm.AddProxy(proxy)
	}

	request := &ProxyRequest{
		URL:         "https://example.com",
		Method:      "GET",
		SSLRequired: true,
		Timeout:     time.Second * 30,
	}

	proxy, err := epm.GetProxy(request)
	if err != nil {
		t.Fatalf("Failed to get proxy: %v", err)
	}

	if proxy == nil {
		t.Fatal("Should return a proxy")
	}

	if !proxy.Health {
		t.Error("Returned proxy should be healthy")
	}

	if !proxy.SSLSupport {
		t.Error("Returned proxy should support SSL")
	}
}

func TestEnhancedProxyManagerGeographicDistribution(t *testing.T) {
	epm := NewEnhancedProxyManager(EnhancedProxyConfig{
		GeographicDistribution: true,
		MaxProxies:             10,
	})

	// Add proxies from different regions
	regions := []string{"us-east", "us-west", "eu-west", "asia-pac"}
	for i, region := range regions {
		proxy := &EnhancedProxy{
			IP:              fmt.Sprintf("192.168.1.%d", 100+i),
			Port:            8080 + i,
			Protocol:        "http",
			Region:          region,
			Health:          true,
			SSLSupport:      true,
			AnonymityLevel:  "anonymous",
			ConcurrentLimit: 5,
		}
		epm.AddProxy(proxy)
	}

	// Test geographic preference
	request := &ProxyRequest{
		URL:                  "https://example.com",
		GeographicPreference: "us-east",
	}

	proxy, err := epm.GetProxy(request)
	if err != nil {
		t.Fatalf("Failed to get proxy: %v", err)
	}

	if proxy.Region != "us-east" {
		t.Errorf("Expected region us-east, got %s", proxy.Region)
	}
}

func TestEnhancedProxyManagerHealthChecking(t *testing.T) {
	epm := NewEnhancedProxyManager(EnhancedProxyConfig{
		HealthCheckInterval: time.Second * 1,
		MaxFailures:         2,
	})

	proxy := &EnhancedProxy{
		IP:              "192.168.1.100",
		Port:            8080,
		Protocol:        "http",
		Health:          true,
		SSLSupport:      true,
		AnonymityLevel:  "anonymous",
		ConcurrentLimit: 5,
	}

	epm.AddProxy(proxy)

	// Mark proxy as failed
	epm.MarkProxyFailure(proxy.ID)
	epm.MarkProxyFailure(proxy.ID)

	// Proxy should be marked as unhealthy after max failures
	if proxy.Health {
		t.Error("Proxy should be marked as unhealthy after max failures")
	}

	// Mark proxy as successful
	epm.MarkProxySuccess(proxy.ID, time.Millisecond*100)

	if !proxy.Health {
		t.Error("Proxy should be marked as healthy after success")
	}

	if proxy.FailCount != 0 {
		t.Error("Fail count should be reset after success")
	}
}

func TestEnhancedProxyManagerRateLimiting(t *testing.T) {
	epm := NewEnhancedProxyManager(EnhancedProxyConfig{
		RateLimiting: true,
		MaxProxies:   10,
	})

	proxy := &EnhancedProxy{
		IP:              "192.168.1.100",
		Port:            8080,
		Protocol:        "http",
		Health:          true,
		SSLSupport:      true,
		AnonymityLevel:  "anonymous",
		ConcurrentLimit: 5,
	}

	epm.AddProxy(proxy)

	// Test rate limiting
	for i := 0; i < 105; i++ { // Default limit is 100
		allowed := epm.rateLimiter.AllowRequest(proxy.ID)
		if i >= 100 && allowed {
			t.Errorf("Request %d should be rate limited", i)
		}
	}
}

func TestEnhancedProxyManagerRotationStrategies(t *testing.T) {
	epm := NewEnhancedProxyManager(EnhancedProxyConfig{
		MaxProxies: 10,
	})

	// Add multiple proxies with different latencies
	for i := 0; i < 3; i++ {
		proxy := &EnhancedProxy{
			IP:              fmt.Sprintf("192.168.1.%d", 100+i),
			Port:            8080 + i,
			Protocol:        "http",
			Health:          true,
			SSLSupport:      true,
			AnonymityLevel:  "anonymous",
			ConcurrentLimit: 5,
			Latency:         time.Millisecond * time.Duration(100+i*50),
			CurrentLoad:     i,
		}
		epm.AddProxy(proxy)
	}

	request := &ProxyRequest{
		URL:    "https://example.com",
		Method: "GET",
	}

	// Test round-robin strategy
	epm.rotationEngine.current = "round_robin"
	proxy1, _ := epm.GetProxy(request)
	proxy2, _ := epm.GetProxy(request)

	if proxy1.ID == proxy2.ID {
		t.Error("Round-robin should select different proxies")
	}

	// Test load-balanced strategy
	epm.rotationEngine.current = "load_balanced"
	proxy3, _ := epm.GetProxy(request)

	// Should select proxy with lowest load (proxy with CurrentLoad = 0)
	if proxy3.CurrentLoad != 0 {
		t.Error("Load-balanced strategy should select proxy with lowest load")
	}

	// Test latency-based strategy
	epm.rotationEngine.current = "latency_based"
	proxy4, _ := epm.GetProxy(request)

	// Should select proxy with lowest latency (proxy with Latency = 100ms)
	if proxy4.Latency != time.Millisecond*100 {
		t.Error("Latency-based strategy should select proxy with lowest latency")
	}
}

func TestEnhancedProxyManagerStats(t *testing.T) {
	epm := NewEnhancedProxyManager(EnhancedProxyConfig{
		MaxProxies: 10,
	})

	// Add proxies
	for i := 0; i < 3; i++ {
		proxy := &EnhancedProxy{
			IP:              fmt.Sprintf("192.168.1.%d", 100+i),
			Port:            8080 + i,
			Protocol:        "http",
			Health:          i < 2, // First 2 are healthy
			SSLSupport:      true,
			AnonymityLevel:  "anonymous",
			ConcurrentLimit: 5,
			Latency:         time.Millisecond * 100,
			CurrentLoad:     i,
		}
		epm.AddProxy(proxy)
	}

	stats := epm.GetStats()

	if stats["total_proxies"] != 3 {
		t.Errorf("Expected total_proxies to be 3, got %v", stats["total_proxies"])
	}

	if stats["healthy_proxies"] != 2 {
		t.Errorf("Expected healthy_proxies to be 2, got %v", stats["healthy_proxies"])
	}

	healthRatio := stats["health_ratio"].(float64)
	if healthRatio != 2.0/3.0 {
		t.Errorf("Expected health_ratio to be %f, got %f", 2.0/3.0, healthRatio)
	}

	if stats["active_load"] != 3 { // 0 + 1 + 2
		t.Errorf("Expected active_load to be 3, got %v", stats["active_load"])
	}
}

func TestEnhancedProxyManagerConcurrentAccess(t *testing.T) {
	epm := NewEnhancedProxyManager(EnhancedProxyConfig{
		MaxProxies: 100,
	})

	// Add a proxy
	proxy := &EnhancedProxy{
		IP:              "192.168.1.100",
		Port:            8080,
		Protocol:        "http",
		Health:          true,
		SSLSupport:      true,
		AnonymityLevel:  "anonymous",
		ConcurrentLimit: 10,
	}
	epm.AddProxy(proxy)

	// Test concurrent access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			request := &ProxyRequest{
				URL:    "https://example.com",
				Method: "GET",
			}
			_, err := epm.GetProxy(request)
			if err != nil {
				t.Errorf("Failed to get proxy: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check that proxy load is managed correctly
	if proxy.CurrentLoad > proxy.ConcurrentLimit {
		t.Errorf("Proxy load (%d) should not exceed concurrent limit (%d)", proxy.CurrentLoad, proxy.ConcurrentLimit)
	}
}

func TestEnhancedProxyManagerProxyRelease(t *testing.T) {
	epm := NewEnhancedProxyManager(EnhancedProxyConfig{
		MaxProxies: 10,
	})

	proxy := &EnhancedProxy{
		IP:              "192.168.1.100",
		Port:            8080,
		Protocol:        "http",
		Health:          true,
		SSLSupport:      true,
		AnonymityLevel:  "anonymous",
		ConcurrentLimit: 5,
	}
	epm.AddProxy(proxy)

	// Get proxy (increases load)
	request := &ProxyRequest{
		URL:    "https://example.com",
		Method: "GET",
	}
	_, err := epm.GetProxy(request)
	if err != nil {
		t.Fatalf("Failed to get proxy: %v", err)
	}

	initialLoad := proxy.CurrentLoad

	// Release proxy
	epm.ReleaseProxy(proxy.ID)

	if proxy.CurrentLoad != initialLoad-1 {
		t.Errorf("Expected load to decrease by 1, got %d", proxy.CurrentLoad)
	}
}

func TestEnhancedProxyManagerMaxProxiesLimit(t *testing.T) {
	epm := NewEnhancedProxyManager(EnhancedProxyConfig{
		MaxProxies: 2,
	})

	// Add proxies up to limit
	for i := 0; i < 2; i++ {
		proxy := &EnhancedProxy{
			IP:             fmt.Sprintf("192.168.1.%d", 100+i),
			Port:           8080 + i,
			Protocol:       "http",
			Health:         true,
			SSLSupport:     true,
			AnonymityLevel: "anonymous",
		}
		err := epm.AddProxy(proxy)
		if err != nil {
			t.Fatalf("Failed to add proxy %d: %v", i, err)
		}
	}

	// Try to add one more proxy
	proxy := &EnhancedProxy{
		IP:             "192.168.1.102",
		Port:           8082,
		Protocol:       "http",
		Health:         true,
		SSLSupport:     true,
		AnonymityLevel: "anonymous",
	}
	err := epm.AddProxy(proxy)
	if err == nil {
		t.Error("Should not allow adding proxy beyond max limit")
	}
}
