package routing

import (
	"context"
	"testing"
	"time"

	"kyb-platform/internal/architecture"
	"kyb-platform/internal/config"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/shared"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"go.opentelemetry.io/otel/trace"
)

// MockModuleManager is a mock implementation of ModuleManager
type MockModuleManager struct {
	mock.Mock
}

func (m *MockModuleManager) GetAvailableModules() map[string]architecture.Module {
	args := m.Called()
	return args.Get(0).(map[string]architecture.Module)
}

func (m *MockModuleManager) GetModuleByID(moduleID string) (architecture.Module, bool) {
	args := m.Called(moduleID)
	return args.Get(0).(architecture.Module), args.Bool(1)
}

func (m *MockModuleManager) GetModulesByType(moduleType string) []architecture.Module {
	args := m.Called(moduleType)
	return args.Get(0).([]architecture.Module)
}

func (m *MockModuleManager) GetModuleHealth(moduleID string) (architecture.ModuleStatus, error) {
	args := m.Called(moduleID)
	return args.Get(0).(architecture.ModuleStatus), args.Error(1)
}

// MockModule is a mock implementation of architecture.Module
type MockModule struct {
	mock.Mock
}

func (m *MockModule) ID() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockModule) Metadata() architecture.ModuleMetadata {
	args := m.Called()
	return args.Get(0).(architecture.ModuleMetadata)
}

func (m *MockModule) Config() architecture.ModuleConfig {
	args := m.Called()
	return args.Get(0).(architecture.ModuleConfig)
}

func (m *MockModule) Health() architecture.ModuleHealth {
	args := m.Called()
	return args.Get(0).(architecture.ModuleHealth)
}

func (m *MockModule) Start(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockModule) Stop(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockModule) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockModule) Process(ctx context.Context, req architecture.ModuleRequest) (architecture.ModuleResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(architecture.ModuleResponse), args.Error(1)
}

func (m *MockModule) CanHandle(req architecture.ModuleRequest) bool {
	args := m.Called(req)
	return args.Bool(0)
}

func (m *MockModule) HealthCheck(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockModule) OnEvent(event architecture.ModuleEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

// TestRequestAnalyzer tests the RequestAnalyzer functionality
func TestRequestAnalyzer(t *testing.T) {
	// Create test logger and tracer
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	// Create request analyzer with default config
	config := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
	}
	analyzer := NewRequestAnalyzer(logger, tracer, config)

	// Test case 1: Simple request
	t.Run("Simple Request Analysis", func(t *testing.T) {
		req := &shared.BusinessClassificationRequest{
			ID:           "test-1",
			BusinessName: "Test Business",
		}

		result, err := analyzer.AnalyzeRequest(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-1", result.RequestID)
		assert.Equal(t, RequestTypeSimple, result.RequestType)
		assert.Equal(t, ComplexityLevelLow, result.Complexity)
		assert.Equal(t, PriorityLevelMedium, result.Priority)
		assert.Len(t, result.Recommendations, 2) // web_search_analysis and keyword_classification
	})

	// Test case 2: Complex request with website
	t.Run("Complex Request with Website", func(t *testing.T) {
		req := &shared.BusinessClassificationRequest{
			ID:           "test-2",
			BusinessName: "Tech Solutions Inc",
			WebsiteURL:   "https://techsolutions.com",
			Keywords:     []string{"software", "technology", "consulting"},
			Description:  "Leading software development and technology consulting company",
		}

		result, err := analyzer.AnalyzeRequest(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, RequestTypeComplex, result.RequestType)
		assert.Equal(t, ComplexityLevelHigh, result.Complexity)
		assert.Equal(t, PriorityLevelHigh, result.Priority)
		assert.Len(t, result.Recommendations, 4) // website_analysis, ml_classification, web_search_analysis, keyword_classification
	})

	// Test case 3: Urgent request
	t.Run("Urgent Request", func(t *testing.T) {
		req := &shared.BusinessClassificationRequest{
			ID:           "test-3",
			BusinessName: "Urgent Business",
			Metadata: map[string]interface{}{
				"urgent": true,
			},
		}

		result, err := analyzer.AnalyzeRequest(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, RequestTypeUrgent, result.RequestType)
		assert.Equal(t, PriorityLevelUrgent, result.Priority)
	})

	// Test case 4: Batch request
	t.Run("Batch Request", func(t *testing.T) {
		req := &shared.BusinessClassificationRequest{
			ID:           "test-4",
			BusinessName: "Batch Business",
			Metadata: map[string]interface{}{
				"batch_size": 10,
			},
		}

		result, err := analyzer.AnalyzeRequest(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, RequestTypeBatch, result.RequestType)
	})
}

// TestModuleSelector tests the ModuleSelector functionality
func TestModuleSelector(t *testing.T) {
	// Create test logger and tracer
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})

	// Create mock module manager
	mockModuleManager := &MockModuleManager{}

	// Create test modules
	websiteModule := &MockModule{}
	websiteModule.On("ID").Return("website_analysis_1")
	websiteModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Website Analysis Module",
		Version:      "1.0.0",
		Description:  "Analyzes business websites",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityClassification, architecture.CapabilityWebAnalysis},
		Priority:     architecture.PriorityHigh,
	})
	websiteModule.On("IsRunning").Return(true)

	mlModule := &MockModule{}
	mlModule.On("ID").Return("ml_classification_1")
	mlModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "ML Classification Module",
		Version:      "1.0.0",
		Description:  "ML-based classification",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityClassification, architecture.CapabilityMLPrediction},
		Priority:     architecture.PriorityMedium,
	})
	mlModule.On("IsRunning").Return(true)

	// Setup mock module manager
	availableModules := map[string]architecture.Module{
		"website_analysis_1":  websiteModule,
		"ml_classification_1": mlModule,
	}
	mockModuleManager.On("GetAvailableModules").Return(availableModules)
	mockModuleManager.On("GetModuleByID", "website_analysis_1").Return(websiteModule, true)
	mockModuleManager.On("GetModuleByID", "ml_classification_1").Return(mlModule, true)
	mockModuleManager.On("GetModuleHealth", "website_analysis_1").Return(architecture.ModuleStatusHealthy, nil)
	mockModuleManager.On("GetModuleHealth", "ml_classification_1").Return(architecture.ModuleStatusHealthy, nil)

	// Create module selector
	config := ModuleSelectorConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       true,
		EnableFallbackRouting:     true,
	}
	selector := NewModuleSelector(logger, tracer, config, mockModuleManager, metrics)

	// Test case 1: Select module for website analysis
	t.Run("Select Module for Website Analysis", func(t *testing.T) {
		req := &shared.BusinessClassificationRequest{
			ID:           "test-1",
			BusinessName: "Test Business",
			WebsiteURL:   "https://testbusiness.com",
		}

		analysis := &RequestAnalysisResult{
			RequestType: RequestTypeStandard,
			Complexity:  ComplexityLevelMedium,
			Priority:    PriorityLevelMedium,
			Recommendations: []RoutingRecommendation{
				{
					ModuleType: "website_analysis",
					Confidence: 0.9,
				},
			},
		}

		result, err := selector.SelectModule(context.Background(), req, analysis)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.SelectedModule)
		assert.Equal(t, "website_analysis_1", result.SelectedModule.ModuleID)
		assert.Equal(t, "website_analysis", result.SelectedModule.ModuleType)
		assert.True(t, result.Confidence > 0)
	})

	// Test case 2: Select module for ML classification
	t.Run("Select Module for ML Classification", func(t *testing.T) {
		req := &shared.BusinessClassificationRequest{
			ID:           "test-2",
			BusinessName: "ML Business",
			Keywords:     []string{"technology", "software"},
			Description:  "Technology company",
		}

		analysis := &RequestAnalysisResult{
			RequestType: RequestTypeStandard,
			Complexity:  ComplexityLevelMedium,
			Priority:    PriorityLevelMedium,
			Recommendations: []RoutingRecommendation{
				{
					ModuleType: "ml_classification",
					Confidence: 0.8,
				},
			},
		}

		result, err := selector.SelectModule(context.Background(), req, analysis)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.SelectedModule)
		assert.Equal(t, "ml_classification_1", result.SelectedModule.ModuleID)
		assert.Equal(t, "ml_classification", result.SelectedModule.ModuleType)
	})

	// Test case 3: Performance tracking
	t.Run("Performance Tracking", func(t *testing.T) {
		// Update performance for a module
		selector.UpdateModulePerformance("website_analysis_1", true, 100*time.Millisecond)
		selector.UpdateModulePerformance("website_analysis_1", true, 150*time.Millisecond)
		selector.UpdateModulePerformance("website_analysis_1", false, 200*time.Millisecond)

		// Verify performance score is calculated
		score := selector.getPerformanceScore("website_analysis_1")
		assert.True(t, score > 0)
		assert.True(t, score <= 1.0)
	})
}

// TestIntelligentRouter tests the IntelligentRouter functionality
func TestIntelligentRouter(t *testing.T) {
	// Create test logger, tracer, and metrics
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})

	// Create mock module manager
	mockModuleManager := &MockModuleManager{}

	// Create test module
	testModule := &MockModule{}
	testModule.On("ID").Return("test_module_1")
	testModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Test Module",
		Version:      "1.0.0",
		Description:  "Test classification module",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityClassification},
		Priority:     architecture.PriorityMedium,
	})
	testModule.On("IsRunning").Return(true)
	testModule.On("Process").Return(architecture.ModuleResponse{
		ID:      "test-1",
		Success: true,
		Data: map[string]interface{}{
			"classification": map[string]interface{}{
				"industry": "Technology",
			},
		},
		Confidence: 0.85,
		Latency:    100 * time.Millisecond,
		Metadata:   map[string]interface{}{},
	}, nil)

	// Setup mock module manager
	availableModules := map[string]architecture.Module{
		"test_module_1": testModule,
	}
	mockModuleManager.On("GetAvailableModules").Return(availableModules)
	mockModuleManager.On("GetModuleByID", "test_module_1").Return(testModule, true)
	mockModuleManager.On("GetModuleHealth", "test_module_1").Return(architecture.ModuleStatusHealthy, nil)

	// Create request analyzer
	analyzerConfig := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
	}
	requestAnalyzer := NewRequestAnalyzer(logger, tracer, analyzerConfig)

	// Create module selector
	selectorConfig := ModuleSelectorConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       true,
		EnableFallbackRouting:     true,
	}
	moduleSelector := NewModuleSelector(logger, tracer, selectorConfig, mockModuleManager, metrics)

	// Create intelligent router
	routerConfig := IntelligentRouterConfig{
		EnableRequestAnalysis:    true,
		EnableModuleSelection:    true,
		EnableRetryLogic:         true,
		EnableFallbackProcessing: true,
		MaxConcurrentRequests:    10,
		RequestTimeout:           30 * time.Second,
		RetryAttempts:            3,
		RetryDelay:               1 * time.Second,
	}
	router := NewIntelligentRouter(logger, tracer, metrics, routerConfig, requestAnalyzer, moduleSelector, mockModuleManager)

	// Test case 1: Successful routing
	t.Run("Successful Routing", func(t *testing.T) {
		req := &shared.BusinessClassificationRequest{
			ID:           "test-1",
			BusinessName: "Test Business",
			WebsiteURL:   "https://testbusiness.com",
		}

		result, err := router.RouteRequest(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-1", result.ID)
		assert.Len(t, result.Classifications, 1)
		assert.Equal(t, "Technology", result.Classifications[0].IndustryName)
		assert.Equal(t, 0.85, result.OverallConfidence)
	})

	// Test case 2: Get active requests
	t.Run("Get Active Requests", func(t *testing.T) {
		activeRequests := router.GetActiveRequests()
		assert.NotNil(t, activeRequests)
		// Should be empty after request completion
		assert.Len(t, activeRequests, 0)
	})

	// Test case 3: Get router metrics
	t.Run("Get Router Metrics", func(t *testing.T) {
		metrics := router.GetRouterMetrics()
		assert.NotNil(t, metrics)
		assert.Equal(t, int64(1), metrics.TotalRequests)
		assert.Equal(t, int64(1), metrics.SuccessfulRequests)
		assert.Equal(t, int64(0), metrics.FailedRequests)
		assert.True(t, metrics.AverageProcessingTime > 0)
	})

	// Test case 4: Get request context
	t.Run("Get Request Context", func(t *testing.T) {
		context, exists := router.GetRequestContext("test-1")
		assert.False(t, exists) // Request should be completed and removed
		assert.Nil(t, context)
	})
}

// TestComplexityCalculation tests complexity calculation logic
func TestComplexityCalculation(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	config := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
	}
	analyzer := NewRequestAnalyzer(logger, tracer, config)

	t.Run("Business Name Complexity", func(t *testing.T) {
		// Test short business name
		score := analyzer.calculateBusinessNameComplexity("ABC")
		assert.True(t, score < 0.1)

		// Test long business name
		score = analyzer.calculateBusinessNameComplexity("Very Long Business Name With Many Words And Special Characters!")
		assert.True(t, score > 0.1)
	})

	t.Run("Website Complexity", func(t *testing.T) {
		// Test simple URL
		score := analyzer.calculateWebsiteComplexity("https://example.com")
		assert.True(t, score < 0.1)

		// Test complex URL
		score = analyzer.calculateWebsiteComplexity("https://subdomain.example.com/path/to/page?param=value#fragment")
		assert.True(t, score > 0.1)
	})

	t.Run("Keywords Complexity", func(t *testing.T) {
		// Test simple keywords
		score := analyzer.calculateKeywordsComplexity([]string{"tech", "software"})
		assert.True(t, score < 0.1)

		// Test complex keywords
		score = analyzer.calculateKeywordsComplexity([]string{"very-long-keyword-with-special-characters!", "another-complex-keyword"})
		assert.True(t, score > 0.1)
	})

	t.Run("Description Complexity", func(t *testing.T) {
		// Test short description
		score := analyzer.calculateDescriptionComplexity("Simple description")
		assert.True(t, score < 0.1)

		// Test long description with technical terms
		longDesc := "This is a very long description that contains many technical terms like software development, engineering consulting, and technology services. It goes on and on with lots of details about the business operations and capabilities."
		score = analyzer.calculateDescriptionComplexity(longDesc)
		assert.True(t, score > 0.2)
	})
}

// TestPriorityCalculation tests priority calculation logic
func TestPriorityCalculation(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	config := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
	}
	analyzer := NewRequestAnalyzer(logger, tracer, config)

	t.Run("Basic Priority Calculation", func(t *testing.T) {
		req := &shared.BusinessClassificationRequest{
			ID:           "test-1",
			BusinessName: "Test Business",
		}

		score := analyzer.calculatePriorityScore(req)
		assert.True(t, score > 0)
		assert.True(t, score <= 1.0)
	})

	t.Run("High Priority Request", func(t *testing.T) {
		req := &shared.BusinessClassificationRequest{
			ID:           "test-2",
			BusinessName: "Test Business",
			WebsiteURL:   "https://test.com",
			Keywords:     []string{"tech", "software"},
			Description:  "Technology company",
			Metadata: map[string]interface{}{
				"urgent": true,
			},
		}

		score := analyzer.calculatePriorityScore(req)
		assert.True(t, score > 0.8) // Should be high priority
	})

	t.Run("Low Priority Request", func(t *testing.T) {
		req := &shared.BusinessClassificationRequest{
			ID: "test-3",
			// Minimal information
		}

		score := analyzer.calculatePriorityScore(req)
		assert.True(t, score < 0.5) // Should be low priority
	})
}

// TestLoadBalancingStrategies tests different load balancing strategies
func TestLoadBalancingStrategies(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "info",
		LogFormat: "json",
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})

	// Create mock module manager
	mockModuleManager := &MockModuleManager{}

	// Create test modules with different loads
	module1 := &MockModule{}
	module1.On("ID").Return("module_1")
	module1.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Module 1",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityClassification},
		Priority:     architecture.PriorityMedium,
	})
	module1.On("IsRunning").Return(true)

	module2 := &MockModule{}
	module2.On("ID").Return("module_2")
	module2.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Module 2",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityClassification},
		Priority:     architecture.PriorityMedium,
	})
	module2.On("IsRunning").Return(true)

	// Setup mock module manager
	availableModules := map[string]architecture.Module{
		"module_1": module1,
		"module_2": module2,
	}
	mockModuleManager.On("GetAvailableModules").Return(availableModules)
	mockModuleManager.On("GetModuleByID", "module_1").Return(module1, true)
	mockModuleManager.On("GetModuleByID", "module_2").Return(module2, true)
	mockModuleManager.On("GetModuleHealth", "module_1").Return(architecture.ModuleStatusHealthy, nil)
	mockModuleManager.On("GetModuleHealth", "module_2").Return(architecture.ModuleStatusHealthy, nil)

	t.Run("Adaptive Load Balancing", func(t *testing.T) {
		config := ModuleSelectorConfig{
			LoadBalancingStrategy: LoadBalancingStrategyAdaptive,
		}
		selector := NewModuleSelector(logger, tracer, config, mockModuleManager, metrics)

		// Update performance for modules
		selector.UpdateModulePerformance("module_1", true, 100*time.Millisecond)
		selector.UpdateModulePerformance("module_2", true, 200*time.Millisecond)

		req := &shared.BusinessClassificationRequest{
			ID:           "test-1",
			BusinessName: "Test Business",
		}

		analysis := &RequestAnalysisResult{
			RequestType: RequestTypeSimple,
			Complexity:  ComplexityLevelLow,
			Priority:    PriorityLevelMedium,
		}

		result, err := selector.SelectModule(context.Background(), req, analysis)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.SelectedModule)
		// Should select module_1 due to better performance
		assert.Equal(t, "module_1", result.SelectedModule.ModuleID)
	})

	t.Run("Best Performance Strategy", func(t *testing.T) {
		config := ModuleSelectorConfig{
			LoadBalancingStrategy: LoadBalancingStrategyBestPerformance,
		}
		selector := NewModuleSelector(logger, tracer, config, mockModuleManager, metrics)

		// Update performance for modules
		selector.UpdateModulePerformance("module_1", true, 150*time.Millisecond)
		selector.UpdateModulePerformance("module_2", true, 100*time.Millisecond)

		req := &shared.BusinessClassificationRequest{
			ID:           "test-2",
			BusinessName: "Test Business",
		}

		analysis := &RequestAnalysisResult{
			RequestType: RequestTypeSimple,
			Complexity:  ComplexityLevelLow,
			Priority:    PriorityLevelMedium,
		}

		result, err := selector.SelectModule(context.Background(), req, analysis)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.SelectedModule)
		// Should select module_2 due to better performance
		assert.Equal(t, "module_2", result.SelectedModule.ModuleID)
	})
}

// TestModuleSelectorInputTypeBasedSelection tests input type-based module selection
func TestModuleSelectorInputTypeBasedSelection(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	config := ModuleSelectorConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       true,
		EnableFallbackRouting:     true,
		LoadBalancingStrategy:     LoadBalancingStrategyAdaptive,
		ConfidenceThreshold:       0.7,
	}

	// Create mock module manager
	moduleManager := &MockModuleManager{}

	// Create test modules with proper mock setup
	websiteModule := &MockModule{}
	websiteModule.On("ID").Return("website_analysis")
	websiteModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Website Analysis Module",
		Version:      "1.0.0",
		Description:  "Analyzes business websites",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityWebAnalysis, architecture.CapabilityDataExtraction},
		Priority:     architecture.PriorityHigh,
	})
	websiteModule.On("IsRunning").Return(true)

	webSearchModule := &MockModule{}
	webSearchModule.On("ID").Return("web_search_analysis")
	webSearchModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Web Search Analysis Module",
		Version:      "1.0.0",
		Description:  "Analyzes web search results",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityWebAnalysis, architecture.CapabilityDataExtraction},
		Priority:     architecture.PriorityMedium,
	})
	webSearchModule.On("IsRunning").Return(true)

	mlModule := &MockModule{}
	mlModule.On("ID").Return("ml_classification")
	mlModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "ML Classification Module",
		Version:      "1.0.0",
		Description:  "ML-based classification",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityMLPrediction, architecture.CapabilityDataExtraction},
		Priority:     architecture.PriorityHigh,
	})
	mlModule.On("IsRunning").Return(true)

	keywordModule := &MockModule{}
	keywordModule.On("ID").Return("keyword_classification")
	keywordModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Keyword Classification Module",
		Version:      "1.0.0",
		Description:  "Keyword-based classification",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityDataExtraction},
		Priority:     architecture.PriorityLow,
	})
	keywordModule.On("IsRunning").Return(true)

	// Setup mock module manager
	availableModules := map[string]architecture.Module{
		"website_analysis":       websiteModule,
		"web_search_analysis":    webSearchModule,
		"ml_classification":      mlModule,
		"keyword_classification": keywordModule,
	}
	moduleManager.On("GetAvailableModules").Return(availableModules)
	moduleManager.On("GetModuleByID", "website_analysis").Return(websiteModule, true)
	moduleManager.On("GetModuleByID", "web_search_analysis").Return(webSearchModule, true)
	moduleManager.On("GetModuleByID", "ml_classification").Return(mlModule, true)
	moduleManager.On("GetModuleByID", "keyword_classification").Return(keywordModule, true)
	moduleManager.On("GetModuleHealth", "website_analysis").Return(architecture.ModuleStatusHealthy, nil)
	moduleManager.On("GetModuleHealth", "web_search_analysis").Return(architecture.ModuleStatusHealthy, nil)
	moduleManager.On("GetModuleHealth", "ml_classification").Return(architecture.ModuleStatusHealthy, nil)
	moduleManager.On("GetModuleHealth", "keyword_classification").Return(architecture.ModuleStatusHealthy, nil)

	selector := NewModuleSelector(logger, tracer, config, moduleManager, metrics)

	tests := []struct {
		name           string
		request        *shared.BusinessClassificationRequest
		expectedModule string
		description    string
	}{
		{
			name: "Website URL request should select website_analysis",
			request: &shared.BusinessClassificationRequest{
				ID:           "test-1",
				BusinessName: "Test Company",
				WebsiteURL:   "https://example.com",
				Description:  "A technology company",
			},
			expectedModule: "website_analysis",
			description:    "Requests with website URLs should prefer website analysis",
		},
		{
			name: "Research request without website should select web_search_analysis",
			request: &shared.BusinessClassificationRequest{
				ID:           "test-2",
				BusinessName: "Research Company",
				Description:  "A company for research purposes",
				Keywords:     []string{"research", "analysis"},
			},
			expectedModule: "web_search_analysis",
			description:    "Research requests without websites should prefer web search analysis",
		},
		{
			name: "High-quality data request should select ml_classification",
			request: &shared.BusinessClassificationRequest{
				ID:               "test-3",
				BusinessName:     "Quality Company",
				Description:      "A comprehensive description of a high-quality company with detailed information about their business operations and industry focus",
				Keywords:         []string{"technology", "software", "development"},
				Industry:         "Technology",
				GeographicRegion: "US",
			},
			expectedModule: "ml_classification",
			description:    "Requests with high-quality, complete data should prefer ML classification",
		},
		{
			name: "Simple request should select keyword_classification",
			request: &shared.BusinessClassificationRequest{
				ID:           "test-4",
				BusinessName: "Simple Company",
			},
			expectedModule: "keyword_classification",
			description:    "Simple requests should prefer keyword classification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request analyzer to get analysis
			analyzerConfig := RequestAnalyzerConfig{
				EnableComplexityAnalysis: true,
				EnablePriorityAssessment: true,
				MaxRequestSize:           10000,
				DefaultTimeout:           30 * time.Second,
			}
			analyzer := NewRequestAnalyzer(logger, tracer, analyzerConfig)

			// Analyze the request
			analysis, err := analyzer.AnalyzeRequest(context.Background(), tt.request)
			assert.NoError(t, err)
			assert.NotNil(t, analysis)

			// Select module
			result, err := selector.SelectModule(context.Background(), tt.request, analysis)
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotNil(t, result.SelectedModule)

			// Verify the selected module
			assert.Equal(t, tt.expectedModule, result.SelectedModule.ModuleID, tt.description)

			// Verify selection confidence
			assert.Greater(t, result.Confidence, 0.5, "Selection confidence should be reasonable")

			// Verify selection reason
			assert.NotEmpty(t, result.SelectionReason, "Selection reason should be provided")
		})
	}
}

// TestModuleSelectorInputCharacteristics tests input characteristics analysis
func TestModuleSelectorInputCharacteristics(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	config := ModuleSelectorConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       true,
		EnableFallbackRouting:     true,
		LoadBalancingStrategy:     LoadBalancingStrategyAdaptive,
		ConfidenceThreshold:       0.7,
	}

	// Create mock module manager
	moduleManager := &MockModuleManager{}

	// Create test module
	websiteModule := &MockModule{}
	websiteModule.On("ID").Return("website_analysis")
	websiteModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Website Analysis Module",
		Version:      "1.0.0",
		Description:  "Analyzes business websites",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityWebAnalysis},
		Priority:     architecture.PriorityHigh,
	})
	websiteModule.On("IsRunning").Return(true)

	// Setup mock module manager
	availableModules := map[string]architecture.Module{
		"website_analysis": websiteModule,
	}
	moduleManager.On("GetAvailableModules").Return(availableModules)
	moduleManager.On("GetModuleByID", "website_analysis").Return(websiteModule, true)
	moduleManager.On("GetModuleHealth", "website_analysis").Return(architecture.ModuleStatusHealthy, nil)

	selector := NewModuleSelector(logger, tracer, config, moduleManager, metrics)

	tests := []struct {
		name                    string
		request                 *shared.BusinessClassificationRequest
		expectedHasWebsite      bool
		expectedHasBusinessName bool
		expectedDataQuality     float64
		expectedCompleteness    float64
	}{
		{
			name: "Complete request with website",
			request: &shared.BusinessClassificationRequest{
				ID:               "test-1",
				BusinessName:     "Complete Company",
				WebsiteURL:       "https://example.com",
				Description:      "A complete company description",
				Keywords:         []string{"technology", "software"},
				Industry:         "Technology",
				GeographicRegion: "US",
			},
			expectedHasWebsite:      true,
			expectedHasBusinessName: true,
			expectedDataQuality:     0.8, // High quality
			expectedCompleteness:    1.0, // Complete
		},
		{
			name: "Minimal request",
			request: &shared.BusinessClassificationRequest{
				ID:           "test-2",
				BusinessName: "Minimal Company",
			},
			expectedHasWebsite:      false,
			expectedHasBusinessName: true,
			expectedDataQuality:     0.9,  // Good quality (only business name)
			expectedCompleteness:    0.17, // Low completeness (1/6 fields)
		},
		{
			name: "Request with poor quality data",
			request: &shared.BusinessClassificationRequest{
				ID:           "test-3",
				BusinessName: "A",           // Too short
				WebsiteURL:   "invalid-url", // Invalid format
				Description:  "Too short",   // Too short
			},
			expectedHasWebsite:      true,
			expectedHasBusinessName: true,
			expectedDataQuality:     0.5, // Lower quality due to poor data
			expectedCompleteness:    0.5, // Medium completeness (3/6 fields)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request analyzer
			analyzerConfig := RequestAnalyzerConfig{
				EnableComplexityAnalysis: true,
				EnablePriorityAssessment: true,
			}
			analyzer := NewRequestAnalyzer(logger, tracer, analyzerConfig)

			// Analyze the request
			analysis, err := analyzer.AnalyzeRequest(context.Background(), tt.request)
			assert.NoError(t, err)
			assert.NotNil(t, analysis)

			// Extract input characteristics
			characteristics := selector.analyzeInputCharacteristics(analysis)

			// Verify characteristics
			assert.Equal(t, tt.expectedHasWebsite, characteristics.HasWebsiteURL)
			assert.Equal(t, tt.expectedHasBusinessName, characteristics.HasBusinessName)
			assert.InDelta(t, tt.expectedDataQuality, characteristics.DataQuality, 0.1)
			assert.InDelta(t, tt.expectedCompleteness, characteristics.DataCompleteness, 0.1)
		})
	}
}

// TestModuleSelectorScoringAlgorithms tests the scoring algorithms for different module types
func TestModuleSelectorScoringAlgorithms(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	config := ModuleSelectorConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       true,
		EnableFallbackRouting:     true,
		LoadBalancingStrategy:     LoadBalancingStrategyAdaptive,
		ConfidenceThreshold:       0.7,
	}

	// Create mock module manager
	moduleManager := &MockModuleManager{}
	moduleManager.On("GetAvailableModules").Return(map[string]architecture.Module{})

	selector := NewModuleSelector(logger, tracer, config, moduleManager, metrics)

	tests := []struct {
		name            string
		moduleType      string
		characteristics *InputCharacteristics
		requestType     RequestType
		expectedScore   float64
		description     string
	}{
		{
			name:       "Website analysis with website URL",
			moduleType: "website_analysis",
			characteristics: &InputCharacteristics{
				HasWebsiteURL:    true,
				HasBusinessName:  true,
				HasDescription:   true,
				DataQuality:      0.8,
				DataCompleteness: 0.7,
			},
			requestType:   RequestTypeStandard,
			expectedScore: 0.8, // Should be high due to website URL
			description:   "Website analysis should score high when website URL is available",
		},
		{
			name:       "Website analysis without website URL",
			moduleType: "website_analysis",
			characteristics: &InputCharacteristics{
				HasWebsiteURL:    false,
				HasBusinessName:  true,
				HasDescription:   true,
				DataQuality:      0.8,
				DataCompleteness: 0.7,
			},
			requestType:   RequestTypeStandard,
			expectedScore: 0.4, // Should be lower without website URL
			description:   "Website analysis should score lower when website URL is not available",
		},
		{
			name:       "Web search analysis for research request",
			moduleType: "web_search_analysis",
			characteristics: &InputCharacteristics{
				HasWebsiteURL:    false,
				HasBusinessName:  true,
				HasDescription:   true,
				DataQuality:      0.7,
				DataCompleteness: 0.5,
			},
			requestType:   RequestTypeResearch,
			expectedScore: 0.8, // Should be high for research requests
			description:   "Web search analysis should score high for research requests",
		},
		{
			name:       "ML classification with high-quality data",
			moduleType: "ml_classification",
			characteristics: &InputCharacteristics{
				HasWebsiteURL:    false,
				HasBusinessName:  true,
				HasDescription:   true,
				DataQuality:      0.9,
				DataCompleteness: 0.8,
			},
			requestType:   RequestTypeComplex,
			expectedScore: 0.8, // Should be high for high-quality data
			description:   "ML classification should score high for high-quality data",
		},
		{
			name:       "Keyword classification for simple request",
			moduleType: "keyword_classification",
			characteristics: &InputCharacteristics{
				HasWebsiteURL:    false,
				HasBusinessName:  true,
				HasKeywords:      true,
				DataQuality:      0.6,
				DataCompleteness: 0.3,
			},
			requestType:   RequestTypeSimple,
			expectedScore: 0.7, // Should be good for simple requests
			description:   "Keyword classification should score well for simple requests",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			moduleInfo := ModuleInfo{
				ModuleID:   "test_module",
				ModuleType: tt.moduleType,
				Capabilities: []architecture.ModuleCapability{
					architecture.CapabilityDataExtraction,
				},
				Priority:       architecture.PriorityMedium,
				HealthStatus:   architecture.ModuleStatusHealthy,
				IsRunning:      true,
				CurrentLoad:    0,
				MaxConcurrency: 10,
			}

			score := selector.calculateInputTypeScore(moduleInfo, tt.characteristics, tt.requestType)

			assert.Greater(t, score, 0.0, "Score should be positive")
			assert.InDelta(t, tt.expectedScore, score, 0.2, tt.description)
		})
	}
}

// TestRequestAnalyzerDataQualityCalculation tests data quality calculation
func TestRequestAnalyzerDataQualityCalculation(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{Level: "debug"})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{Level: "debug"})
	tracer := observability.NewTracer(&config.ObservabilityConfig{Level: "debug"})

	config := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
	}

	analyzer := NewRequestAnalyzer(logger, tracer, config)

	tests := []struct {
		name            string
		request         *shared.BusinessClassificationRequest
		expectedQuality float64
		description     string
	}{
		{
			name: "High quality data",
			request: &shared.BusinessClassificationRequest{
				ID:               "test-1",
				BusinessName:     "High Quality Company",
				WebsiteURL:       "https://example.com",
				Description:      "A comprehensive description of a high-quality company with detailed information",
				Keywords:         []string{"technology", "software"},
				Industry:         "Technology",
				GeographicRegion: "United States",
			},
			expectedQuality: 0.8,
			description:     "Complete, well-formatted data should have high quality",
		},
		{
			name: "Poor quality data",
			request: &shared.BusinessClassificationRequest{
				ID:           "test-2",
				BusinessName: "A",           // Too short
				WebsiteURL:   "invalid-url", // Invalid format
				Description:  "Short",       // Too short
				Keywords:     []string{"a"}, // Too short
			},
			expectedQuality: 0.5,
			description:     "Poorly formatted data should have lower quality",
		},
		{
			name: "Empty request",
			request: &shared.BusinessClassificationRequest{
				ID: "test-3",
			},
			expectedQuality: 0.0,
			description:     "Empty request should have zero quality",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quality := analyzer.calculateDataQuality(tt.request)
			assert.InDelta(t, tt.expectedQuality, quality, 0.1, tt.description)
		})
	}
}

// TestRequestAnalyzerDataCompletenessCalculation tests data completeness calculation
func TestRequestAnalyzerDataCompletenessCalculation(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{Level: "debug"})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{Level: "debug"})
	tracer := observability.NewTracer(&config.ObservabilityConfig{Level: "debug"})

	config := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
	}

	analyzer := NewRequestAnalyzer(logger, tracer, config)

	tests := []struct {
		name                 string
		request              *shared.BusinessClassificationRequest
		expectedCompleteness float64
		description          string
	}{
		{
			name: "Complete data",
			request: &shared.BusinessClassificationRequest{
				ID:               "test-1",
				BusinessName:     "Complete Company",
				WebsiteURL:       "https://example.com",
				Description:      "Description",
				Keywords:         []string{"keyword"},
				Industry:         "Technology",
				GeographicRegion: "US",
			},
			expectedCompleteness: 1.0,
			description:          "All fields filled should have 100% completeness",
		},
		{
			name: "Partial data",
			request: &shared.BusinessClassificationRequest{
				ID:           "test-2",
				BusinessName: "Partial Company",
				WebsiteURL:   "https://example.com",
				Description:  "Description",
			},
			expectedCompleteness: 0.5, // 3/6 fields
			description:          "Half the fields filled should have 50% completeness",
		},
		{
			name: "Minimal data",
			request: &shared.BusinessClassificationRequest{
				ID:           "test-3",
				BusinessName: "Minimal Company",
			},
			expectedCompleteness: 0.17, // 1/6 fields
			description:          "Only business name should have low completeness",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			completeness := analyzer.calculateDataCompleteness(tt.request)
			assert.InDelta(t, tt.expectedCompleteness, completeness, 0.01, tt.description)
		})
	}
}

// TestIntelligentRouterParallelProcessing tests parallel processing capabilities
func TestIntelligentRouterParallelProcessing(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	// Create mock module manager
	mockModuleManager := &MockModuleManager{}

	// Create test modules
	websiteModule := &MockModule{}
	websiteModule.On("ID").Return("website_analysis")
	websiteModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Website Analysis Module",
		Version:      "1.0.0",
		Description:  "Analyzes business websites",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityWebAnalysis},
		Priority:     architecture.PriorityHigh,
	})
	websiteModule.On("IsRunning").Return(true)

	mlModule := &MockModule{}
	mlModule.On("ID").Return("ml_classification")
	mlModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "ML Classification Module",
		Version:      "1.0.0",
		Description:  "ML-based classification",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityMLPrediction},
		Priority:     architecture.PriorityMedium,
	})
	mlModule.On("IsRunning").Return(true)

	keywordModule := &MockModule{}
	keywordModule.On("ID").Return("keyword_classification")
	keywordModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Keyword Classification Module",
		Version:      "1.0.0",
		Description:  "Keyword-based classification",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityDataExtraction},
		Priority:     architecture.PriorityLow,
	})
	keywordModule.On("IsRunning").Return(true)

	// Setup mock module manager
	availableModules := map[string]architecture.Module{
		"website_analysis":       websiteModule,
		"ml_classification":      mlModule,
		"keyword_classification": keywordModule,
	}
	mockModuleManager.On("GetAvailableModules").Return(availableModules)
	mockModuleManager.On("GetModuleByID", "website_analysis").Return(websiteModule, true)
	mockModuleManager.On("GetModuleByID", "ml_classification").Return(mlModule, true)
	mockModuleManager.On("GetModuleByID", "keyword_classification").Return(keywordModule, true)
	mockModuleManager.On("GetModuleHealth", "website_analysis").Return(architecture.ModuleStatusHealthy, nil)
	mockModuleManager.On("GetModuleHealth", "ml_classification").Return(architecture.ModuleStatusHealthy, nil)
	mockModuleManager.On("GetModuleHealth", "keyword_classification").Return(architecture.ModuleStatusHealthy, nil)

	// Create request analyzer
	analyzerConfig := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
	}
	requestAnalyzer := NewRequestAnalyzer(logger, tracer, analyzerConfig)

	// Create module selector
	selectorConfig := ModuleSelectorConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       true,
		EnableFallbackRouting:     true,
		LoadBalancingStrategy:     LoadBalancingStrategyAdaptive,
		ConfidenceThreshold:       0.7,
	}
	moduleSelector := NewModuleSelector(logger, tracer, selectorConfig, mockModuleManager, metrics)

	tests := []struct {
		name                 string
		config               IntelligentRouterConfig
		request              *shared.BusinessClassificationRequest
		expectedParallelMode ParallelProcessingMode
		description          string
	}{
		{
			name: "Concurrent parallel processing",
			config: IntelligentRouterConfig{
				EnableParallelProcessing: true,
				ParallelProcessingMode:   ParallelProcessingModeConcurrent,
				MaxConcurrentRequests:    10,
				MaxParallelModules:       3,
				WorkerPoolSize:           5,
				EnableFallbackProcessing: true,
			},
			request: &shared.BusinessClassificationRequest{
				ID:           "test-1",
				BusinessName: "Test Business",
				WebsiteURL:   "https://testbusiness.com",
			},
			expectedParallelMode: ParallelProcessingModeConcurrent,
			description:          "Should use concurrent parallel processing mode",
		},
		{
			name: "Hybrid parallel processing",
			config: IntelligentRouterConfig{
				EnableParallelProcessing: true,
				ParallelProcessingMode:   ParallelProcessingModeHybrid,
				MaxConcurrentRequests:    10,
				MaxParallelModules:       3,
				WorkerPoolSize:           5,
				EnableFallbackProcessing: true,
			},
			request: &shared.BusinessClassificationRequest{
				ID:           "test-2",
				BusinessName: "Test Business",
				WebsiteURL:   "https://testbusiness.com",
			},
			expectedParallelMode: ParallelProcessingModeHybrid,
			description:          "Should use hybrid parallel processing mode",
		},
		{
			name: "Sequential processing (parallel disabled)",
			config: IntelligentRouterConfig{
				EnableParallelProcessing: false,
				ParallelProcessingMode:   ParallelProcessingModeSequential,
				MaxConcurrentRequests:    10,
				MaxParallelModules:       3,
				WorkerPoolSize:           5,
				EnableFallbackProcessing: true,
			},
			request: &shared.BusinessClassificationRequest{
				ID:           "test-3",
				BusinessName: "Test Business",
				WebsiteURL:   "https://testbusiness.com",
			},
			expectedParallelMode: ParallelProcessingModeSequential,
			description:          "Should use sequential processing when parallel is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create intelligent router
			router := NewIntelligentRouter(
				logger,
				tracer,
				metrics,
				tt.config,
				requestAnalyzer,
				moduleSelector,
				mockModuleManager,
			)

			// Verify configuration
			assert.Equal(t, tt.expectedParallelMode, router.config.ParallelProcessingMode, tt.description)
			assert.Equal(t, tt.config.EnableParallelProcessing, router.config.EnableParallelProcessing)
			assert.Equal(t, tt.config.MaxConcurrentRequests, router.config.MaxConcurrentRequests)
			assert.Equal(t, tt.config.MaxParallelModules, router.config.MaxParallelModules)
			assert.Equal(t, tt.config.WorkerPoolSize, router.config.WorkerPoolSize)

			// Verify parallel processing components are initialized
			assert.NotNil(t, router.requestSemaphore, "Request semaphore should be initialized")
			assert.NotNil(t, router.workerPool, "Worker pool should be initialized")
		})
	}
}

// TestIntelligentRouterConcurrentProcessing tests concurrent module processing
func TestIntelligentRouterConcurrentProcessing(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	// Create mock module manager
	mockModuleManager := &MockModuleManager{}

	// Create test modules with mock processing
	websiteModule := &MockModule{}
	websiteModule.On("ID").Return("website_analysis")
	websiteModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Website Analysis Module",
		Version:      "1.0.0",
		Description:  "Analyzes business websites",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityWebAnalysis},
		Priority:     architecture.PriorityHigh,
	})
	websiteModule.On("IsRunning").Return(true)
	websiteModule.On("Process", mock.Anything, mock.Anything).Return(architecture.ModuleResponse{
		ID:         "test-1",
		Success:    true,
		Confidence: 0.9,
		Data: map[string]interface{}{
			"classification": map[string]interface{}{
				"classifications": []interface{}{
					map[string]interface{}{
						"industry_code":         "541511",
						"industry_name":         "Technology",
						"confidence_score":      0.9,
						"classification_method": "website_analysis",
						"description":           "Technology company",
					},
				},
			},
		},
		Metadata: map[string]interface{}{},
	}, nil)

	mlModule := &MockModule{}
	mlModule.On("ID").Return("ml_classification")
	mlModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "ML Classification Module",
		Version:      "1.0.0",
		Description:  "ML-based classification",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityMLPrediction},
		Priority:     architecture.PriorityMedium,
	})
	mlModule.On("IsRunning").Return(true)
	mlModule.On("Process", mock.Anything, mock.Anything).Return(architecture.ModuleResponse{
		ID:         "test-1",
		Success:    true,
		Confidence: 0.85,
		Data: map[string]interface{}{
			"classification": map[string]interface{}{
				"classifications": []interface{}{
					map[string]interface{}{
						"industry_code":         "541511",
						"industry_name":         "Technology",
						"confidence_score":      0.85,
						"classification_method": "ml_classification",
						"description":           "Technology company",
					},
				},
			},
		},
		Metadata: map[string]interface{}{},
	}, nil)

	// Setup mock module manager
	availableModules := map[string]architecture.Module{
		"website_analysis":  websiteModule,
		"ml_classification": mlModule,
	}
	mockModuleManager.On("GetAvailableModules").Return(availableModules)
	mockModuleManager.On("GetModuleByID", "website_analysis").Return(websiteModule, true)
	mockModuleManager.On("GetModuleByID", "ml_classification").Return(mlModule, true)
	mockModuleManager.On("GetModuleHealth", "website_analysis").Return(architecture.ModuleStatusHealthy, nil)
	mockModuleManager.On("GetModuleHealth", "ml_classification").Return(architecture.ModuleStatusHealthy, nil)

	// Create request analyzer
	analyzerConfig := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
	}
	requestAnalyzer := NewRequestAnalyzer(logger, tracer, analyzerConfig)

	// Create module selector
	selectorConfig := ModuleSelectorConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       true,
		EnableFallbackRouting:     true,
		LoadBalancingStrategy:     LoadBalancingStrategyAdaptive,
		ConfidenceThreshold:       0.7,
	}
	moduleSelector := NewModuleSelector(logger, tracer, selectorConfig, mockModuleManager, metrics)

	// Create intelligent router with concurrent processing
	routerConfig := IntelligentRouterConfig{
		EnableParallelProcessing: true,
		ParallelProcessingMode:   ParallelProcessingModeConcurrent,
		MaxConcurrentRequests:    10,
		MaxParallelModules:       3,
		WorkerPoolSize:           5,
		EnableFallbackProcessing: true,
		RequestTimeout:           30 * time.Second,
	}
	router := NewIntelligentRouter(
		logger,
		tracer,
		metrics,
		routerConfig,
		requestAnalyzer,
		moduleSelector,
		mockModuleManager,
	)

	t.Run("Concurrent module processing", func(t *testing.T) {
		request := &shared.BusinessClassificationRequest{
			ID:           "test-1",
			BusinessName: "Test Business",
			WebsiteURL:   "https://testbusiness.com",
		}

		// Route request
		response, err := router.RouteRequest(context.Background(), request)

		// Verify no error
		assert.NoError(t, err, "Concurrent processing should not return an error")
		assert.NotNil(t, response, "Response should not be nil")

		// Verify response structure
		assert.Equal(t, "test-1", response.ID)
		assert.Greater(t, response.OverallConfidence, 0.0, "Confidence should be positive")
		assert.NotEmpty(t, response.Classifications, "Should have classifications")

		// Verify parallel processing was used
		assert.Equal(t, ParallelProcessingModeConcurrent, router.config.ParallelProcessingMode)
	})
}

// TestIntelligentRouterHybridProcessing tests hybrid parallel/sequential processing
func TestIntelligentRouterHybridProcessing(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	// Create mock module manager
	mockModuleManager := &MockModuleManager{}

	// Create test modules
	websiteModule := &MockModule{}
	websiteModule.On("ID").Return("website_analysis")
	websiteModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Website Analysis Module",
		Version:      "1.0.0",
		Description:  "Analyzes business websites",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityWebAnalysis},
		Priority:     architecture.PriorityHigh,
	})
	websiteModule.On("IsRunning").Return(true)
	websiteModule.On("Process", mock.Anything, mock.Anything).Return(architecture.ModuleResponse{
		ID:         "test-1",
		Success:    true,
		Confidence: 0.9,
		Data: map[string]interface{}{
			"classification": map[string]interface{}{
				"classifications": []interface{}{
					map[string]interface{}{
						"industry_code":         "541511",
						"industry_name":         "Technology",
						"confidence_score":      0.9,
						"classification_method": "website_analysis",
						"description":           "Technology company",
					},
				},
			},
		},
		Metadata: map[string]interface{}{},
	}, nil)

	// Setup mock module manager
	availableModules := map[string]architecture.Module{
		"website_analysis": websiteModule,
	}
	mockModuleManager.On("GetAvailableModules").Return(availableModules)
	mockModuleManager.On("GetModuleByID", "website_analysis").Return(websiteModule, true)
	mockModuleManager.On("GetModuleHealth", "website_analysis").Return(architecture.ModuleStatusHealthy, nil)

	// Create request analyzer
	analyzerConfig := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
	}
	requestAnalyzer := NewRequestAnalyzer(logger, tracer, analyzerConfig)

	// Create module selector
	selectorConfig := ModuleSelectorConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       true,
		EnableFallbackRouting:     true,
		LoadBalancingStrategy:     LoadBalancingStrategyAdaptive,
		ConfidenceThreshold:       0.7,
	}
	moduleSelector := NewModuleSelector(logger, tracer, selectorConfig, mockModuleManager, metrics)

	// Create intelligent router with hybrid processing
	routerConfig := IntelligentRouterConfig{
		EnableParallelProcessing: true,
		ParallelProcessingMode:   ParallelProcessingModeHybrid,
		MaxConcurrentRequests:    10,
		MaxParallelModules:       3,
		WorkerPoolSize:           5,
		EnableFallbackProcessing: true,
		RequestTimeout:           30 * time.Second,
	}
	router := NewIntelligentRouter(
		logger,
		tracer,
		metrics,
		routerConfig,
		requestAnalyzer,
		moduleSelector,
		mockModuleManager,
	)

	t.Run("Hybrid processing with primary module success", func(t *testing.T) {
		request := &shared.BusinessClassificationRequest{
			ID:           "test-1",
			BusinessName: "Test Business",
			WebsiteURL:   "https://testbusiness.com",
		}

		// Route request
		response, err := router.RouteRequest(context.Background(), request)

		// Verify no error
		assert.NoError(t, err, "Hybrid processing should not return an error")
		assert.NotNil(t, response, "Response should not be nil")

		// Verify response structure
		assert.Equal(t, "test-1", response.ID)
		assert.Greater(t, response.OverallConfidence, 0.0, "Confidence should be positive")
		assert.NotEmpty(t, response.Classifications, "Should have classifications")

		// Verify hybrid processing was used
		assert.Equal(t, ParallelProcessingModeHybrid, router.config.ParallelProcessingMode)
	})
}

// TestIntelligentRouterResourceManagement tests resource management for parallel processing
func TestIntelligentRouterResourceManagement(t *testing.T) {
	logger := observability.NewLogger(&config.ObservabilityConfig{
		LogLevel:  "debug",
		LogFormat: "json",
	})
	metrics, _ := observability.NewMetrics(&config.ObservabilityConfig{
		MetricsEnabled: true,
	})
	tracer := trace.NewNoopTracerProvider().Tracer("test")

	// Create mock module manager
	mockModuleManager := &MockModuleManager{}

	// Create test modules
	websiteModule := &MockModule{}
	websiteModule.On("ID").Return("website_analysis")
	websiteModule.On("Metadata").Return(architecture.ModuleMetadata{
		Name:         "Website Analysis Module",
		Version:      "1.0.0",
		Description:  "Analyzes business websites",
		Capabilities: []architecture.ModuleCapability{architecture.CapabilityWebAnalysis},
		Priority:     architecture.PriorityHigh,
	})
	websiteModule.On("IsRunning").Return(true)

	// Setup mock module manager
	availableModules := map[string]architecture.Module{
		"website_analysis": websiteModule,
	}
	mockModuleManager.On("GetAvailableModules").Return(availableModules)
	mockModuleManager.On("GetModuleByID", "website_analysis").Return(websiteModule, true)
	mockModuleManager.On("GetModuleHealth", "website_analysis").Return(architecture.ModuleStatusHealthy, nil)

	// Create request analyzer
	analyzerConfig := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
	}
	requestAnalyzer := NewRequestAnalyzer(logger, tracer, analyzerConfig)

	// Create module selector
	selectorConfig := ModuleSelectorConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       true,
		EnableFallbackRouting:     true,
		LoadBalancingStrategy:     LoadBalancingStrategyAdaptive,
		ConfidenceThreshold:       0.7,
	}
	moduleSelector := NewModuleSelector(logger, tracer, selectorConfig, mockModuleManager, metrics)

	tests := []struct {
		name                   string
		maxConcurrentRequests  int
		maxParallelModules     int
		workerPoolSize         int
		expectedSemaphoreSize  int
		expectedWorkerPoolSize int
		description            string
	}{
		{
			name:                   "Default resource limits",
			maxConcurrentRequests:  100,
			maxParallelModules:     5,
			workerPoolSize:         20,
			expectedSemaphoreSize:  100,
			expectedWorkerPoolSize: 20,
			description:            "Should set default resource limits",
		},
		{
			name:                   "Custom resource limits",
			maxConcurrentRequests:  50,
			maxParallelModules:     3,
			workerPoolSize:         10,
			expectedSemaphoreSize:  50,
			expectedWorkerPoolSize: 10,
			description:            "Should set custom resource limits",
		},
		{
			name:                   "Low resource limits",
			maxConcurrentRequests:  5,
			maxParallelModules:     2,
			workerPoolSize:         3,
			expectedSemaphoreSize:  5,
			expectedWorkerPoolSize: 3,
			description:            "Should set low resource limits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create intelligent router with specific resource limits
			routerConfig := IntelligentRouterConfig{
				EnableParallelProcessing: true,
				ParallelProcessingMode:   ParallelProcessingModeConcurrent,
				MaxConcurrentRequests:    tt.maxConcurrentRequests,
				MaxParallelModules:       tt.maxParallelModules,
				WorkerPoolSize:           tt.workerPoolSize,
				EnableFallbackProcessing: true,
				RequestTimeout:           30 * time.Second,
			}
			router := NewIntelligentRouter(
				logger,
				tracer,
				metrics,
				routerConfig,
				requestAnalyzer,
				moduleSelector,
				mockModuleManager,
			)

			// Verify resource limits are set correctly
			assert.Equal(t, tt.expectedSemaphoreSize, cap(router.requestSemaphore), tt.description)
			assert.Equal(t, tt.expectedWorkerPoolSize, cap(router.workerPool), tt.description)
			assert.Equal(t, tt.maxConcurrentRequests, router.config.MaxConcurrentRequests)
			assert.Equal(t, tt.maxParallelModules, router.config.MaxParallelModules)
			assert.Equal(t, tt.workerPoolSize, router.config.WorkerPoolSize)
		})
	}
}
