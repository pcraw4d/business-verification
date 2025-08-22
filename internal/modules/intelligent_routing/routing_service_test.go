package intelligent_routing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// Mock implementations for testing
type mockHealthChecker struct{}

func (m *mockHealthChecker) CheckHealth(ctx context.Context, moduleID string) (*ModuleAvailability, error) {
	return &ModuleAvailability{
		IsAvailable:     true,
		LastHealthCheck: time.Now(),
		HealthScore:     0.95,
		LoadPercentage:  0.3,
		QueueLength:     2,
	}, nil
}

func (m *mockHealthChecker) CheckAllModules(ctx context.Context) (map[string]*ModuleAvailability, error) {
	return map[string]*ModuleAvailability{
		"test_module": {
			IsAvailable:     true,
			LastHealthCheck: time.Now(),
			HealthScore:     0.95,
			LoadPercentage:  0.3,
			QueueLength:     2,
		},
	}, nil
}

func (m *mockHealthChecker) RegisterHealthCallback(ctx context.Context, moduleID string, callback func(*ModuleAvailability)) error {
	return nil
}

type mockLoadBalancer struct{}

func (m *mockLoadBalancer) DistributeLoad(ctx context.Context, modules []*ModuleCapability, request *VerificationRequest) (map[string]float64, error) {
	distribution := make(map[string]float64)
	for _, module := range modules {
		distribution[module.ModuleID] = 0.3
	}
	return distribution, nil
}

func (m *mockLoadBalancer) GetModuleLoad(ctx context.Context, moduleID string) (float64, error) {
	return 0.3, nil
}

func (m *mockLoadBalancer) UpdateModuleLoad(ctx context.Context, moduleID string, load float64) error {
	return nil
}

type mockMetricsCollector struct{}

func (m *mockMetricsCollector) RecordRoutingDecision(ctx context.Context, decision *RoutingDecision) error {
	return nil
}

func (m *mockMetricsCollector) RecordProcessingResult(ctx context.Context, result *ProcessingResult) error {
	return nil
}

func (m *mockMetricsCollector) GetMetrics(ctx context.Context) (*RoutingMetrics, error) {
	return &RoutingMetrics{
		TotalRequests:    100,
		SuccessfulRoutes: 95,
		FailedRoutes:     5,
		AverageLatency:   150.0,
		SuccessRate:      0.95,
		LoadDistribution: map[string]float64{
			"test_module": 0.3,
		},
		LastUpdated: time.Now(),
	}, nil
}

func (m *mockMetricsCollector) ResetMetrics(ctx context.Context) error {
	return nil
}

func TestNewRoutingService(t *testing.T) {
	logger := zap.NewNop()
	requestAnalyzer := NewRequestAnalyzer(nil, logger)
	moduleSelector := NewModuleSelector(nil, &mockHealthChecker{}, &mockLoadBalancer{}, logger)
	loadBalancer := &mockLoadBalancer{}
	healthChecker := &mockHealthChecker{}
	metricsCollector := &mockMetricsCollector{}

	service := NewRoutingService(
		nil,
		requestAnalyzer,
		moduleSelector,
		loadBalancer,
		healthChecker,
		metricsCollector,
		logger,
	)

	assert.NotNil(t, service)
}

func TestRoutingService_RouteRequest(t *testing.T) {
	logger := zap.NewNop()
	requestAnalyzer := NewRequestAnalyzer(nil, logger)
	moduleSelector := NewModuleSelector(nil, &mockHealthChecker{}, &mockLoadBalancer{}, logger)
	loadBalancer := &mockLoadBalancer{}
	healthChecker := &mockHealthChecker{}
	metricsCollector := &mockMetricsCollector{}

	service := NewRoutingService(
		nil,
		requestAnalyzer,
		moduleSelector,
		loadBalancer,
		healthChecker,
		metricsCollector,
		logger,
	)

	// Register a test module
	testModule := &ModuleCapability{
		ModuleID:     "test_module",
		ModuleName:   "Test Module",
		Capabilities: []string{"verification", "validation"},
		RequestTypes: []RequestType{RequestTypeBasic, RequestTypeEnhanced},
		Complexity:   ComplexityModerate,
		Performance: ModulePerformance{
			SuccessRate:    0.95,
			AverageLatency: 150.0,
			Throughput:     100.0,
			ErrorRate:      0.05,
			LastUpdated:    time.Now(),
		},
		Availability: ModuleAvailability{
			IsAvailable:     true,
			LastHealthCheck: time.Now(),
			HealthScore:     0.95,
			LoadPercentage:  0.3,
			QueueLength:     2,
		},
		Specialization: map[string]float64{
			"retail": 0.8,
			"basic":  0.9,
		},
	}

	err := service.RegisterModule(context.Background(), testModule)
	require.NoError(t, err)

	// Create a test request
	request := &VerificationRequest{
		ID:              "test_request_1",
		BusinessName:    "Test Shop",
		BusinessAddress: "123 Test St, Test City, TC 12345",
		RequestType:     RequestTypeBasic,
		Priority:        PriorityNormal,
		Complexity:      ComplexitySimple,
		UserID:          "user1",
		ClientID:        "client1",
		CreatedAt:       time.Now(),
	}

	// Route the request
	decision, err := service.RouteRequest(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, decision)
	assert.Equal(t, request.ID, decision.RequestID)
	assert.NotEmpty(t, decision.SelectedModules)
	assert.Greater(t, decision.Confidence, 0.0)
}

func TestRoutingService_ProcessRequest(t *testing.T) {
	logger := zap.NewNop()
	requestAnalyzer := NewRequestAnalyzer(nil, logger)
	moduleSelector := NewModuleSelector(nil, &mockHealthChecker{}, &mockLoadBalancer{}, logger)
	loadBalancer := &mockLoadBalancer{}
	healthChecker := &mockHealthChecker{}
	metricsCollector := &mockMetricsCollector{}

	service := NewRoutingService(
		nil,
		requestAnalyzer,
		moduleSelector,
		loadBalancer,
		healthChecker,
		metricsCollector,
		logger,
	)

	// Register a test module
	testModule := &ModuleCapability{
		ModuleID:     "test_module",
		ModuleName:   "Test Module",
		Capabilities: []string{"verification", "validation"},
		RequestTypes: []RequestType{RequestTypeBasic, RequestTypeEnhanced},
		Complexity:   ComplexityModerate,
		Performance: ModulePerformance{
			SuccessRate:    0.95,
			AverageLatency: 150.0,
			Throughput:     100.0,
			ErrorRate:      0.05,
			LastUpdated:    time.Now(),
		},
		Availability: ModuleAvailability{
			IsAvailable:     true,
			LastHealthCheck: time.Now(),
			HealthScore:     0.95,
			LoadPercentage:  0.3,
			QueueLength:     2,
		},
		Specialization: map[string]float64{
			"retail": 0.8,
			"basic":  0.9,
		},
	}

	err := service.RegisterModule(context.Background(), testModule)
	require.NoError(t, err)

	// Create a test request
	request := &VerificationRequest{
		ID:              "test_request_2",
		BusinessName:    "Test Shop",
		BusinessAddress: "123 Test St, Test City, TC 12345",
		RequestType:     RequestTypeBasic,
		Priority:        PriorityNormal,
		Complexity:      ComplexitySimple,
		UserID:          "user1",
		ClientID:        "client1",
		CreatedAt:       time.Now(),
	}

	// Process the request
	results, err := service.ProcessRequest(context.Background(), request)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
	assert.Equal(t, request.ID, results[0].RequestID)
	assert.Equal(t, testModule.ModuleID, results[0].ModuleID)
	assert.Equal(t, StatusCompleted, results[0].Status)
}

func TestRequestAnalyzer_AnalyzeRequest(t *testing.T) {
	logger := zap.NewNop()
	analyzer := NewRequestAnalyzer(nil, logger)

	request := &VerificationRequest{
		ID:              "test_request_3",
		BusinessName:    "Acme Bank Corporation",
		BusinessAddress: "456 Financial St, New York, NY 10001",
		RequestType:     RequestTypeCompliance,
		Priority:        PriorityHigh,
		Complexity:      ComplexityComplex,
		UserID:          "user1",
		ClientID:        "client1",
		CreatedAt:       time.Now(),
	}

	analysis, err := analyzer.AnalyzeRequest(context.Background(), request)
	require.NoError(t, err)
	assert.NotNil(t, analysis)
	assert.Equal(t, request.ID, analysis.RequestID)
	assert.NotNil(t, analysis.Classification)
	assert.Equal(t, RequestTypeCompliance, analysis.Classification.RequestType)
	assert.Equal(t, "financial", analysis.Classification.Industry)
	assert.Greater(t, analysis.Confidence, 0.0)
}

func TestModuleSelector_SelectModules(t *testing.T) {
	logger := zap.NewNop()
	selector := NewModuleSelector(nil, &mockHealthChecker{}, &mockLoadBalancer{}, logger)

	// Register test modules
	module1 := &ModuleCapability{
		ModuleID:     "module_1",
		ModuleName:   "Basic Module",
		Capabilities: []string{"verification", "validation"},
		RequestTypes: []RequestType{RequestTypeBasic},
		Complexity:   ComplexitySimple,
		Performance: ModulePerformance{
			SuccessRate:    0.9,
			AverageLatency: 100.0,
			Throughput:     50.0,
			ErrorRate:      0.1,
			LastUpdated:    time.Now(),
		},
		Availability: ModuleAvailability{
			IsAvailable:     true,
			LastHealthCheck: time.Now(),
			HealthScore:     0.9,
			LoadPercentage:  0.2,
			QueueLength:     1,
		},
	}

	module2 := &ModuleCapability{
		ModuleID:     "module_2",
		ModuleName:   "Enhanced Module",
		Capabilities: []string{"verification", "validation", "enhanced_analysis"},
		RequestTypes: []RequestType{RequestTypeBasic, RequestTypeEnhanced},
		Complexity:   ComplexityModerate,
		Performance: ModulePerformance{
			SuccessRate:    0.95,
			AverageLatency: 150.0,
			Throughput:     100.0,
			ErrorRate:      0.05,
			LastUpdated:    time.Now(),
		},
		Availability: ModuleAvailability{
			IsAvailable:     true,
			LastHealthCheck: time.Now(),
			HealthScore:     0.95,
			LoadPercentage:  0.3,
			QueueLength:     2,
		},
	}

	err := selector.RegisterModule(module1)
	require.NoError(t, err)
	err = selector.RegisterModule(module2)
	require.NoError(t, err)

	request := &VerificationRequest{
		ID:              "test_request_4",
		BusinessName:    "Test Shop",
		BusinessAddress: "123 Test St, Test City, TC 12345",
		RequestType:     RequestTypeBasic,
		Priority:        PriorityNormal,
		Complexity:      ComplexitySimple,
		UserID:          "user1",
		ClientID:        "client1",
		CreatedAt:       time.Now(),
	}

	analysis := &RequestAnalysis{
		RequestID: request.ID,
		Classification: &RequestClassification{
			RequestType: RequestTypeBasic,
			Industry:    "retail",
			Confidence:  0.8,
		},
		Complexity: ComplexitySimple,
		Priority:   PriorityNormal,
		Confidence: 0.8,
	}

	modules, err := selector.SelectModules(context.Background(), request, analysis)
	require.NoError(t, err)
	assert.NotEmpty(t, modules)
	assert.Len(t, modules, 2) // Should select both modules as they both support basic requests
}

func TestRoutingService_GetModuleCapabilities(t *testing.T) {
	logger := zap.NewNop()
	requestAnalyzer := NewRequestAnalyzer(nil, logger)
	moduleSelector := NewModuleSelector(nil, &mockHealthChecker{}, &mockLoadBalancer{}, logger)
	loadBalancer := &mockLoadBalancer{}
	healthChecker := &mockHealthChecker{}
	metricsCollector := &mockMetricsCollector{}

	service := NewRoutingService(
		nil,
		requestAnalyzer,
		moduleSelector,
		loadBalancer,
		healthChecker,
		metricsCollector,
		logger,
	)

	// Register test modules
	module1 := &ModuleCapability{
		ModuleID:     "module_1",
		ModuleName:   "Basic Module",
		Capabilities: []string{"verification", "validation"},
		RequestTypes: []RequestType{RequestTypeBasic},
		Complexity:   ComplexitySimple,
	}

	module2 := &ModuleCapability{
		ModuleID:     "module_2",
		ModuleName:   "Enhanced Module",
		Capabilities: []string{"verification", "validation", "enhanced_analysis"},
		RequestTypes: []RequestType{RequestTypeBasic, RequestTypeEnhanced},
		Complexity:   ComplexityModerate,
	}

	err := service.RegisterModule(context.Background(), module1)
	require.NoError(t, err)
	err = service.RegisterModule(context.Background(), module2)
	require.NoError(t, err)

	capabilities, err := service.GetModuleCapabilities(context.Background())
	require.NoError(t, err)
	assert.Len(t, capabilities, 2)
}

func TestRoutingService_CheckModuleHealth(t *testing.T) {
	logger := zap.NewNop()
	requestAnalyzer := NewRequestAnalyzer(nil, logger)
	moduleSelector := NewModuleSelector(nil, &mockHealthChecker{}, &mockLoadBalancer{}, logger)
	loadBalancer := &mockLoadBalancer{}
	healthChecker := &mockHealthChecker{}
	metricsCollector := &mockMetricsCollector{}

	service := NewRoutingService(
		nil,
		requestAnalyzer,
		moduleSelector,
		loadBalancer,
		healthChecker,
		metricsCollector,
		logger,
	)

	health, err := service.CheckModuleHealth(context.Background(), "test_module")
	require.NoError(t, err)
	assert.NotNil(t, health)
	assert.True(t, health.IsAvailable)
	assert.Greater(t, health.HealthScore, 0.0)
}

func TestRoutingService_GetRoutingMetrics(t *testing.T) {
	logger := zap.NewNop()
	requestAnalyzer := NewRequestAnalyzer(nil, logger)
	moduleSelector := NewModuleSelector(nil, &mockHealthChecker{}, &mockLoadBalancer{}, logger)
	loadBalancer := &mockLoadBalancer{}
	healthChecker := &mockHealthChecker{}
	metricsCollector := &mockMetricsCollector{}

	service := NewRoutingService(
		nil,
		requestAnalyzer,
		moduleSelector,
		loadBalancer,
		healthChecker,
		metricsCollector,
		logger,
	)

	metrics, err := service.GetRoutingMetrics(context.Background())
	require.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(100), metrics.TotalRequests)
	assert.Equal(t, int64(95), metrics.SuccessfulRoutes)
	assert.Equal(t, float64(0.95), metrics.SuccessRate)
}
