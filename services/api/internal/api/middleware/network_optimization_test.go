package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNetworkOptimizationManager(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultNetworkOptimizationConfig()
	manager := NewNetworkOptimizationManager(config, logger)

	t.Run("default configuration", func(t *testing.T) {
		if manager.config == nil {
			t.Error("expected config to be set")
		}
		if manager.clientPool == nil {
			t.Error("expected client pool to be initialized")
		}
		if manager.loadBalancer == nil {
			t.Error("expected load balancer to be initialized")
		}
		if manager.rateLimiter == nil {
			t.Error("expected rate limiter to be initialized")
		}
		if manager.circuitBreaker == nil {
			t.Error("expected circuit breaker to be initialized")
		}
	})

	t.Run("get HTTP client", func(t *testing.T) {
		client := manager.GetHTTPClient("example.com")
		if client == nil {
			t.Error("expected HTTP client to be returned")
		}

		// Test that same host returns same client
		client2 := manager.GetHTTPClient("example.com")
		if client != client2 {
			t.Error("expected same client for same host")
		}
	})

	t.Run("add and remove endpoint", func(t *testing.T) {
		err := manager.AddEndpoint("http://example1.com", 1)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		err = manager.AddEndpoint("http://example2.com", 2)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		err = manager.RemoveEndpoint("http://example1.com")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		err = manager.RemoveEndpoint("nonexistent.com")
		if err == nil {
			t.Error("expected error for nonexistent endpoint")
		}
	})

	t.Run("shutdown", func(t *testing.T) {
		err := manager.Shutdown()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestHTTPClientPool(t *testing.T) {
	config := DefaultNetworkOptimizationConfig()
	pool := NewHTTPClientPool(config)

	t.Run("get client for new host", func(t *testing.T) {
		client := pool.GetClient("example.com")
		if client == nil {
			t.Error("expected HTTP client to be returned")
		}

		// Verify client configuration
		if client.Timeout != config.RequestTimeout {
			t.Errorf("expected timeout %v, got %v", config.RequestTimeout, client.Timeout)
		}
	})

	t.Run("get client for existing host", func(t *testing.T) {
		client1 := pool.GetClient("test.com")
		client2 := pool.GetClient("test.com")
		if client1 != client2 {
			t.Error("expected same client for same host")
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				client := pool.GetClient("concurrent.com")
				if client == nil {
					t.Error("expected HTTP client to be returned")
				}
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestNetworkLoadBalancer(t *testing.T) {
	config := DefaultNetworkOptimizationConfig()
	lb := NewNetworkLoadBalancer(config)

	t.Run("add endpoints", func(t *testing.T) {
		err := lb.AddEndpoint("http://example1.com", 1)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		err = lb.AddEndpoint("http://example2.com", 2)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(lb.endpoints) != 2 {
			t.Errorf("expected 2 endpoints, got %d", len(lb.endpoints))
		}
	})

	t.Run("remove endpoint", func(t *testing.T) {
		err := lb.RemoveEndpoint("http://example1.com")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if len(lb.endpoints) != 1 {
			t.Errorf("expected 1 endpoint, got %d", len(lb.endpoints))
		}

		err = lb.RemoveEndpoint("nonexistent.com")
		if err == nil {
			t.Error("expected error for nonexistent endpoint")
		}
	})

	t.Run("select endpoint with no endpoints", func(t *testing.T) {
		lb2 := NewNetworkLoadBalancer(config)
		endpoint := lb2.SelectEndpoint()
		if endpoint != nil {
			t.Error("expected nil endpoint when no endpoints available")
		}
	})
}

func TestNetworkLoadBalancingStrategies(t *testing.T) {
	t.Run("round robin strategy", func(t *testing.T) {
		strategy := &RoundRobinStrategy{}
		endpoints := []*Endpoint{
			{URL: "http://example1.com"},
			{URL: "http://example2.com"},
			{URL: "http://example3.com"},
		}

		// Test multiple selections
		selected := make(map[string]int)
		for i := 0; i < 30; i++ {
			endpoint := strategy.SelectEndpoint(endpoints)
			if endpoint == nil {
				t.Error("expected endpoint to be selected")
			}
			selected[endpoint.URL]++
		}

		// Each endpoint should be selected approximately 10 times
		for url, count := range selected {
			if count < 8 || count > 12 {
				t.Errorf("endpoint %s selected %d times, expected around 10", url, count)
			}
		}
	})

	t.Run("weighted strategy", func(t *testing.T) {
		strategy := &WeightedStrategy{}
		endpoints := []*Endpoint{
			{URL: "http://example1.com", Weight: 1},
			{URL: "http://example2.com", Weight: 2},
			{URL: "http://example3.com", Weight: 3},
		}

		// Test multiple selections
		selected := make(map[string]int)
		for i := 0; i < 60; i++ {
			endpoint := strategy.SelectEndpoint(endpoints)
			if endpoint == nil {
				t.Error("expected endpoint to be selected")
			}
			selected[endpoint.URL]++
		}

		// Weight 3 endpoint should be selected more than weight 1
		if selected["http://example3.com"] <= selected["http://example1.com"] {
			t.Error("weighted strategy not working correctly")
		}
	})

	t.Run("least connections strategy", func(t *testing.T) {
		strategy := &LeastConnectionsStrategy{}
		endpoints := []*Endpoint{
			{URL: "http://example1.com", SuccessCount: 10, ErrorCount: 5},
			{URL: "http://example2.com", SuccessCount: 5, ErrorCount: 2},
			{URL: "http://example3.com", SuccessCount: 15, ErrorCount: 8},
		}

		endpoint := strategy.SelectEndpoint(endpoints)
		if endpoint == nil {
			t.Error("expected endpoint to be selected")
		}

		// Should select endpoint with least connections (example2.com: 7 total)
		if endpoint.URL != "http://example2.com" {
			t.Errorf("expected example2.com, got %s", endpoint.URL)
		}
	})
}

func TestRateLimiter(t *testing.T) {
	config := DefaultNetworkOptimizationConfig()
	config.RateLimitPerSecond = 10
	config.RateLimitBurst = 5

	rl := NewRateLimiter(config)

	t.Run("initial burst", func(t *testing.T) {
		// Should allow initial burst
		for i := 0; i < 5; i++ {
			if !rl.Allow() {
				t.Errorf("expected request %d to be allowed", i)
			}
		}

		// Should reject after burst
		if rl.Allow() {
			t.Error("expected request to be rejected after burst")
		}
	})

	t.Run("token refill", func(t *testing.T) {
		// Wait for token refill
		time.Sleep(200 * time.Millisecond)

		// Should allow some requests after refill
		allowed := 0
		for i := 0; i < 10; i++ {
			if rl.Allow() {
				allowed++
			}
		}

		if allowed == 0 {
			t.Error("expected some requests to be allowed after refill")
		}
	})
}

func TestCircuitBreaker(t *testing.T) {
	config := DefaultNetworkOptimizationConfig()
	config.FailureThreshold = 3
	config.RecoveryTimeout = 100 * time.Millisecond
	config.HalfOpenLimit = 2

	cb := NewCircuitBreaker(config)

	t.Run("initial state", func(t *testing.T) {
		if !cb.CanExecute() {
			t.Error("expected circuit breaker to be closed initially")
		}
	})

	t.Run("failure threshold", func(t *testing.T) {
		// Record failures
		for i := 0; i < 3; i++ {
			cb.RecordFailure()
		}

		// Should be open after failure threshold
		if cb.CanExecute() {
			t.Error("expected circuit breaker to be open after failure threshold")
		}
	})

	t.Run("recovery", func(t *testing.T) {
		// Wait for recovery timeout
		time.Sleep(150 * time.Millisecond)

		// Should be half-open
		if !cb.CanExecute() {
			t.Error("expected circuit breaker to be half-open after recovery timeout")
		}

		// Record success
		cb.RecordSuccess()
		cb.RecordSuccess()

		// Should be closed again
		if !cb.CanExecute() {
			t.Error("expected circuit breaker to be closed after successful recovery")
		}
	})
}

func TestNetworkOptimizationManager_DoRequest(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultNetworkOptimizationConfig()
	config.RateLimitPerSecond = 100
	config.RateLimitBurst = 10
	manager := NewNetworkOptimizationManager(config, logger)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	t.Run("successful request", func(t *testing.T) {
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		resp, err := manager.DoRequest(context.Background(), req)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if resp == nil {
			t.Error("expected response")
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("rate limiting", func(t *testing.T) {
		config2 := DefaultNetworkOptimizationConfig()
		config2.RateLimitPerSecond = 1
		config2.RateLimitBurst = 1
		manager2 := NewNetworkOptimizationManager(config2, logger)

		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		// First request should succeed
		resp, err := manager2.DoRequest(context.Background(), req)
		if err != nil {
			t.Errorf("expected first request to succeed, got %v", err)
		}
		if resp != nil {
			resp.Body.Close()
		}

		// Second request should be rate limited
		resp, err = manager2.DoRequest(context.Background(), req)
		if err == nil {
			t.Error("expected second request to be rate limited")
		}
		if resp != nil {
			resp.Body.Close()
		}
	})

	t.Run("circuit breaker", func(t *testing.T) {
		config3 := DefaultNetworkOptimizationConfig()
		config3.FailureThreshold = 1
		config3.RecoveryTimeout = 100 * time.Millisecond
		manager3 := NewNetworkOptimizationManager(config3, logger)

		// Create failing server
		failingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer failingServer.Close()

		req, err := http.NewRequest("GET", failingServer.URL, nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		// First request should fail (but not due to circuit breaker)
		resp, err := manager3.DoRequest(context.Background(), req)
		if err != nil {
			t.Errorf("expected first request to fail but not error, got %v", err)
		}
		if resp != nil {
			resp.Body.Close()
		}

		// Second request should be blocked by circuit breaker
		resp, err = manager3.DoRequest(context.Background(), req)
		if err == nil {
			t.Error("expected second request to be blocked by circuit breaker")
		} else if !strings.Contains(err.Error(), "circuit breaker is open") {
			t.Errorf("expected circuit breaker error, got: %v", err)
		}
		if resp != nil {
			resp.Body.Close()
		}
	})
}

func TestNetworkOptimizationManager_GetStats(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultNetworkOptimizationConfig()
	manager := NewNetworkOptimizationManager(config, logger)

	t.Run("initial stats", func(t *testing.T) {
		stats := manager.GetStats()
		if stats == nil {
			t.Error("expected stats to be returned")
		}
		if stats.TotalRequests != 0 {
			t.Errorf("expected 0 total requests, got %d", stats.TotalRequests)
		}
	})

	t.Run("stats after request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}

		manager.DoRequest(context.Background(), req)

		stats := manager.GetStats()
		if stats.TotalRequests != 1 {
			t.Errorf("expected 1 total request, got %d", stats.TotalRequests)
		}
		if stats.SuccessfulRequests != 1 {
			t.Errorf("expected 1 successful request, got %d", stats.SuccessfulRequests)
		}
	})
}

func TestNetworkOptimizationManager_OptimizeNetwork(t *testing.T) {
	logger := zap.NewNop()
	config := DefaultNetworkOptimizationConfig()
	config.MetricsEnabled = false // Disable monitoring for tests
	manager := NewNetworkOptimizationManager(config, logger)

	t.Run("optimization with no activity", func(t *testing.T) {
		err := manager.OptimizeNetwork()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("optimization with high error rate", func(t *testing.T) {
		// Simulate high error rate
		manager.stats.TotalRequests = 100
		manager.stats.FailedRequests = 20

		err := manager.OptimizeNetwork()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Rate limit should be reduced
		if manager.config.RateLimitPerSecond >= 100 {
			t.Error("expected rate limit to be reduced")
		}
	})
}

func BenchmarkNetworkOptimizationManager_DoRequest(b *testing.B) {
	logger := zap.NewNop()
	config := DefaultNetworkOptimizationConfig()
	manager := NewNetworkOptimizationManager(config, logger)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		b.Fatalf("failed to create request: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := manager.DoRequest(context.Background(), req)
		if err != nil {
			b.Errorf("request failed: %v", err)
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
}

func BenchmarkHTTPClientPool_GetClient(b *testing.B) {
	config := DefaultNetworkOptimizationConfig()
	pool := NewHTTPClientPool(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := pool.GetClient("example.com")
		if client == nil {
			b.Error("expected HTTP client")
		}
	}
}

func BenchmarkRateLimiter_Allow(b *testing.B) {
	config := DefaultNetworkOptimizationConfig()
	config.RateLimitPerSecond = 1000
	config.RateLimitBurst = 100
	rl := NewRateLimiter(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rl.Allow()
	}
}

func BenchmarkCircuitBreaker_CanExecute(b *testing.B) {
	config := DefaultNetworkOptimizationConfig()
	cb := NewCircuitBreaker(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.CanExecute()
	}
}
