package resilience

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	resilience "kyb-platform/services/risk-assessment-service/internal/resilience"

	"go.uber.org/zap/zaptest"
)

func TestBulkhead(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &resilience.BulkheadConfig{
		DefaultMaxConcurrency: 2,
		DefaultMaxQueueSize:   5,
		DefaultTimeout:        1 * time.Second,
		EnableMetrics:         true,
		EnableLogging:         true,
	}

	bulkhead := resilience.NewBulkhead(config, logger)

	// Test pool creation
	err := bulkhead.CreatePool("test_service", 3, 10, 2*time.Second)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	// Test successful request
	request := &resilience.BulkheadRequest{
		ID:        "test_request_1",
		Service:   "test_service",
		Operation: "test_operation",
		Data:      map[string]interface{}{"test": "data"},
		Timeout:   1 * time.Second,
		Priority:  1,
		CreatedAt: time.Now(),
	}

	processor := func(ctx context.Context, req *resilience.BulkheadRequest) (*resilience.BulkheadResponse, error) {
		time.Sleep(10 * time.Millisecond)
		return &resilience.BulkheadResponse{
			ID:        req.ID,
			Success:   true,
			Result:    map[string]interface{}{"result": "success"},
			CreatedAt: time.Now(),
		}, nil
	}

	ctx := context.Background()
	response, err := bulkhead.Execute(ctx, request, processor)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	if !response.Success {
		t.Fatalf("Expected successful response")
	}

	// Test statistics
	stats := bulkhead.GetStats()
	if stats.TotalRequests != 1 {
		t.Fatalf("Expected 1 total request, got %d", stats.TotalRequests)
	}

	if stats.SuccessfulRequests != 1 {
		t.Fatalf("Expected 1 successful request, got %d", stats.SuccessfulRequests)
	}
}

func TestBulkheadConcurrency(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &resilience.BulkheadConfig{
		DefaultMaxConcurrency: 2,
		DefaultMaxQueueSize:   5,
		DefaultTimeout:        1 * time.Second,
		EnableMetrics:         true,
		EnableLogging:         true,
	}

	bulkhead := resilience.NewBulkhead(config, logger)

	// Test concurrent requests
	numRequests := 5
	responses := make(chan *resilience.BulkheadResponse, numRequests)
	errors := make(chan error, numRequests)

	processor := func(ctx context.Context, req *resilience.BulkheadRequest) (*resilience.BulkheadResponse, error) {
		time.Sleep(50 * time.Millisecond)
		return &resilience.BulkheadResponse{
			ID:        req.ID,
			Success:   true,
			Result:    map[string]interface{}{"result": "success"},
			CreatedAt: time.Now(),
		}, nil
	}

	ctx := context.Background()

	// Start concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(i int) {
			request := &resilience.BulkheadRequest{
				ID:        fmt.Sprintf("test_request_%d", i),
				Service:   "test_service",
				Operation: "test_operation",
				Data:      map[string]interface{}{"test": "data"},
				Timeout:   1 * time.Second,
				Priority:  1,
				CreatedAt: time.Now(),
			}

			response, err := bulkhead.Execute(ctx, request, processor)
			if err != nil {
				errors <- err
			} else {
				responses <- response
			}
		}(i)
	}

	// Collect results
	successCount := 0
	errorCount := 0

	for i := 0; i < numRequests; i++ {
		select {
		case response := <-responses:
			if response.Success {
				successCount++
			}
		case err := <-errors:
			t.Logf("Request failed: %v", err)
			errorCount++
		case <-time.After(2 * time.Second):
			t.Fatalf("Timeout waiting for response")
		}
	}

	// With max concurrency of 2, we should have some successful requests
	if successCount == 0 {
		t.Fatalf("Expected at least one successful request")
	}

	t.Logf("Successful requests: %d, Failed requests: %d", successCount, errorCount)
}

func TestFallbackStrategy(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &resilience.FallbackConfig{
		DefaultTimeout: 1 * time.Second,
		EnableCaching:  true,
		CacheTTL:       5 * time.Minute,
		EnableMetrics:  true,
		EnableLogging:  true,
		MaxRetries:     3,
		RetryDelay:     100 * time.Millisecond,
	}

	fallback := resilience.NewFallbackStrategy(config, logger)

	// Register fallback handler
	handler := &resilience.FallbackHandler{
		Name:         "test_fallback",
		Service:      "test_service",
		FallbackType: "cached_response",
		Config:       map[string]interface{}{"cache_ttl": 300},
		Enabled:      true,
		Timeout:      1 * time.Second,
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
	}

	err := fallback.RegisterFallback(handler)
	if err != nil {
		t.Fatalf("Failed to register fallback: %v", err)
	}

	// Test successful primary processor
	request := &resilience.FallbackRequest{
		ID:        "test_request_1",
		Service:   "test_service",
		Operation: "test_operation",
		Data:      map[string]interface{}{"test": "data"},
		Timeout:   1 * time.Second,
		Priority:  1,
		CreatedAt: time.Now(),
	}

	primaryProcessor := func(ctx context.Context, req *resilience.FallbackRequest) (*resilience.FallbackResponse, error) {
		return &resilience.FallbackResponse{
			ID:        req.ID,
			Success:   true,
			Result:    map[string]interface{}{"result": "primary_success"},
			CreatedAt: time.Now(),
		}, nil
	}

	ctx := context.Background()
	response, err := fallback.ExecuteWithFallback(ctx, request, primaryProcessor)
	if err != nil {
		t.Fatalf("Failed to execute with fallback: %v", err)
	}

	if !response.Success {
		t.Fatalf("Expected successful response")
	}

	if response.FallbackUsed {
		t.Fatalf("Expected primary processor to succeed, not fallback")
	}

	// Test fallback execution
	failingProcessor := func(ctx context.Context, req *resilience.FallbackRequest) (*resilience.FallbackResponse, error) {
		return nil, errors.New("primary processor failed")
	}

	response, err = fallback.ExecuteWithFallback(ctx, request, failingProcessor)
	if err != nil {
		t.Fatalf("Failed to execute fallback: %v", err)
	}

	if !response.Success {
		t.Fatalf("Expected successful fallback response")
	}

	if !response.FallbackUsed {
		t.Fatalf("Expected fallback to be used")
	}

	if response.FallbackType != "cached_response" {
		t.Fatalf("Expected fallback type 'cached_response', got '%s'", response.FallbackType)
	}
}

func TestFallbackStrategyDifferentTypes(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &resilience.FallbackConfig{
		DefaultTimeout: 1 * time.Second,
		EnableCaching:  true,
		CacheTTL:       5 * time.Minute,
		EnableMetrics:  true,
		EnableLogging:  true,
		MaxRetries:     3,
		RetryDelay:     100 * time.Millisecond,
	}

	fallback := resilience.NewFallbackStrategy(config, logger)

	// Test different fallback types
	fallbackTypes := []string{
		"cached_response",
		"alternative_service",
		"degraded_mode",
		"default_response",
	}

	for _, fallbackType := range fallbackTypes {
		handler := &resilience.FallbackHandler{
			Name:         fmt.Sprintf("test_fallback_%s", fallbackType),
			Service:      fmt.Sprintf("test_service_%s", fallbackType),
			FallbackType: fallbackType,
			Config:       map[string]interface{}{"test": "config"},
			Enabled:      true,
			Timeout:      1 * time.Second,
			CacheEnabled: true,
			CacheTTL:     5 * time.Minute,
		}

		err := fallback.RegisterFallback(handler)
		if err != nil {
			t.Fatalf("Failed to register fallback for type %s: %v", fallbackType, err)
		}

		request := &resilience.FallbackRequest{
			ID:        fmt.Sprintf("test_request_%s", fallbackType),
			Service:   fmt.Sprintf("test_service_%s", fallbackType),
			Operation: "test_operation",
			Data:      map[string]interface{}{"test": "data"},
			Timeout:   1 * time.Second,
			Priority:  1,
			CreatedAt: time.Now(),
		}

		failingProcessor := func(ctx context.Context, req *resilience.FallbackRequest) (*resilience.FallbackResponse, error) {
			return nil, errors.New("primary processor failed")
		}

		ctx := context.Background()
		response, err := fallback.ExecuteWithFallback(ctx, request, failingProcessor)
		if err != nil {
			t.Fatalf("Failed to execute fallback for type %s: %v", fallbackType, err)
		}

		if !response.Success {
			t.Fatalf("Expected successful fallback response for type %s", fallbackType)
		}

		if !response.FallbackUsed {
			t.Fatalf("Expected fallback to be used for type %s", fallbackType)
		}

		if response.FallbackType != fallbackType {
			t.Fatalf("Expected fallback type '%s', got '%s'", fallbackType, response.FallbackType)
		}
	}
}

func TestBulkheadManager(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := resilience.NewBulkheadManager(logger)

	// Test getting bulkhead for different services
	config1 := &resilience.BulkheadConfig{
		DefaultMaxConcurrency: 5,
		DefaultMaxQueueSize:   10,
		DefaultTimeout:        2 * time.Second,
		EnableMetrics:         true,
		EnableLogging:         true,
	}

	config2 := &resilience.BulkheadConfig{
		DefaultMaxConcurrency: 3,
		DefaultMaxQueueSize:   5,
		DefaultTimeout:        1 * time.Second,
		EnableMetrics:         true,
		EnableLogging:         true,
	}

	bulkhead1 := manager.GetBulkhead("service1", config1)
	bulkhead2 := manager.GetBulkhead("service2", config2)

	if bulkhead1 == bulkhead2 {
		t.Fatalf("Expected different bulkhead instances")
	}

	// Test getting the same bulkhead again
	bulkhead1Again := manager.GetBulkhead("service1", config1)
	if bulkhead1 != bulkhead1Again {
		t.Fatalf("Expected same bulkhead instance")
	}

	// Test statistics
	allStats := manager.GetAllStats()
	if len(allStats) != 2 {
		t.Fatalf("Expected 2 bulkhead statistics, got %d", len(allStats))
	}

	// Test reset statistics
	manager.ResetAllStats()

	allStatsAfterReset := manager.GetAllStats()
	for _, stats := range allStatsAfterReset {
		if stats.TotalRequests != 0 {
			t.Fatalf("Expected 0 total requests after reset, got %d", stats.TotalRequests)
		}
	}
}

func TestFallbackManager(t *testing.T) {
	logger := zaptest.NewLogger(t)
	manager := resilience.NewFallbackManager(logger)

	// Test getting fallback strategy for different services
	config1 := &resilience.FallbackConfig{
		DefaultTimeout: 2 * time.Second,
		EnableCaching:  true,
		CacheTTL:       10 * time.Minute,
		EnableMetrics:  true,
		EnableLogging:  true,
		MaxRetries:     5,
		RetryDelay:     200 * time.Millisecond,
	}

	config2 := &resilience.FallbackConfig{
		DefaultTimeout: 1 * time.Second,
		EnableCaching:  false,
		CacheTTL:       5 * time.Minute,
		EnableMetrics:  true,
		EnableLogging:  true,
		MaxRetries:     3,
		RetryDelay:     100 * time.Millisecond,
	}

	strategy1 := manager.GetStrategy("strategy1", config1)
	strategy2 := manager.GetStrategy("strategy2", config2)

	if strategy1 == strategy2 {
		t.Fatalf("Expected different fallback strategy instances")
	}

	// Test getting the same strategy again
	strategy1Again := manager.GetStrategy("strategy1", config1)
	if strategy1 != strategy1Again {
		t.Fatalf("Expected same fallback strategy instance")
	}

	// Test statistics
	allStats := manager.GetAllStats()
	if len(allStats) != 2 {
		t.Fatalf("Expected 2 fallback strategy statistics, got %d", len(allStats))
	}

	// Test reset statistics
	manager.ResetAllStats()

	allStatsAfterReset := manager.GetAllStats()
	for _, stats := range allStatsAfterReset {
		if stats.TotalRequests != 0 {
			t.Fatalf("Expected 0 total requests after reset, got %d", stats.TotalRequests)
		}
	}
}

func TestBulkheadTimeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &resilience.BulkheadConfig{
		DefaultMaxConcurrency: 1,
		DefaultMaxQueueSize:   1,
		DefaultTimeout:        100 * time.Millisecond,
		EnableMetrics:         true,
		EnableLogging:         true,
	}

	bulkhead := resilience.NewBulkhead(config, logger)

	// Test timeout scenario
	request := &resilience.BulkheadRequest{
		ID:        "test_request_timeout",
		Service:   "test_service",
		Operation: "test_operation",
		Data:      map[string]interface{}{"test": "data"},
		Timeout:   50 * time.Millisecond,
		Priority:  1,
		CreatedAt: time.Now(),
	}

	slowProcessor := func(ctx context.Context, req *resilience.BulkheadRequest) (*resilience.BulkheadResponse, error) {
		time.Sleep(200 * time.Millisecond)
		return &resilience.BulkheadResponse{
			ID:        req.ID,
			Success:   true,
			Result:    map[string]interface{}{"result": "success"},
			CreatedAt: time.Now(),
		}, nil
	}

	ctx := context.Background()
	_, err := bulkhead.Execute(ctx, request, slowProcessor)
	if err == nil {
		t.Fatalf("Expected timeout error")
	}

	// Check if it's a timeout error
	if !strings.Contains(err.Error(), "timeout") && !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Fatalf("Expected timeout error, got: %v", err)
	}
}

func TestFallbackDisabled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &resilience.FallbackConfig{
		DefaultTimeout: 1 * time.Second,
		EnableCaching:  true,
		CacheTTL:       5 * time.Minute,
		EnableMetrics:  true,
		EnableLogging:  true,
		MaxRetries:     3,
		RetryDelay:     100 * time.Millisecond,
	}

	fallback := resilience.NewFallbackStrategy(config, logger)

	// Register disabled fallback handler
	handler := &resilience.FallbackHandler{
		Name:         "test_fallback_disabled",
		Service:      "test_service",
		FallbackType: "cached_response",
		Config:       map[string]interface{}{"test": "config"},
		Enabled:      false, // Disabled
		Timeout:      1 * time.Second,
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
	}

	err := fallback.RegisterFallback(handler)
	if err != nil {
		t.Fatalf("Failed to register fallback: %v", err)
	}

	request := &resilience.FallbackRequest{
		ID:        "test_request_disabled",
		Service:   "test_service",
		Operation: "test_operation",
		Data:      map[string]interface{}{"test": "data"},
		Timeout:   1 * time.Second,
		Priority:  1,
		CreatedAt: time.Now(),
	}

	failingProcessor := func(ctx context.Context, req *resilience.FallbackRequest) (*resilience.FallbackResponse, error) {
		return nil, errors.New("primary processor failed")
	}

	ctx := context.Background()
	_, err = fallback.ExecuteWithFallback(ctx, request, failingProcessor)
	if err == nil {
		t.Fatalf("Expected error when fallback is disabled")
	}

	if !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("Expected error about disabled fallback, got: %v", err)
	}
}
