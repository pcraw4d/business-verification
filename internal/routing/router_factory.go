package routing

import (
	"context"
	"time"

	"kyb-platform/internal/database"
	"kyb-platform/internal/modules/database_classification"
	"kyb-platform/internal/observability"
	"go.opentelemetry.io/otel/trace"
)

// RouterFactory creates and configures intelligent routers with modules
type RouterFactory struct {
	logger  *observability.Logger
	tracer  trace.Tracer
	metrics *observability.Metrics
}

// NewRouterFactory creates a new router factory
func NewRouterFactory(
	logger *observability.Logger,
	tracer trace.Tracer,
	metrics *observability.Metrics,
) *RouterFactory {
	return &RouterFactory{
		logger:  logger,
		tracer:  tracer,
		metrics: metrics,
	}
}

// CreateIntelligentRouterWithDatabaseClassification creates an intelligent router
// with the database classification module registered
func (rf *RouterFactory) CreateIntelligentRouterWithDatabaseClassification(
	supabaseClient *database.SupabaseClient,
	config IntelligentRouterConfig,
) (*IntelligentRouter, error) {
	// Create module manager
	moduleManager := NewDefaultModuleManager(rf.logger)

	// Create database classification module
	databaseModule, err := database_classification.NewDatabaseClassificationModule(
		supabaseClient,
		nil, // Use default logger
		database_classification.DefaultConfig(),
	)
	if err != nil {
		return nil, err
	}

	// Start the database module
	ctx := context.Background()
	if err := databaseModule.Start(ctx); err != nil {
		return nil, err
	}

	// Register the database classification module
	moduleManager.RegisterModule("database_classification", databaseModule)

	// Create request analyzer
	requestAnalyzerConfig := RequestAnalyzerConfig{
		EnableComplexityAnalysis: true,
		EnablePriorityAssessment: true,
		MaxRequestSize:           1024 * 1024, // 1MB
		DefaultTimeout:           5 * time.Second,
	}
	requestAnalyzer := NewRequestAnalyzer(rf.logger, rf.tracer, requestAnalyzerConfig)

	// Create module selector
	moduleSelectorConfig := ModuleSelectorConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       true,
		EnableFallbackRouting:     true,
		MaxRetries:                3,
		RetryDelay:                1 * time.Second,
		PerformanceWindow:         5 * time.Minute,
		LoadBalancingStrategy:     LoadBalancingStrategyRoundRobin,
		ConfidenceThreshold:       0.7,
	}
	moduleSelector := NewModuleSelector(rf.logger, rf.tracer, moduleSelectorConfig, moduleManager, rf.metrics)

	// Create intelligent router
	router := NewIntelligentRouter(
		rf.logger,
		rf.tracer,
		rf.metrics,
		config,
		requestAnalyzer,
		moduleSelector,
		moduleManager,
	)

	rf.logger.WithComponent("router_factory").Info("intelligent_router_created", map[string]interface{}{
		"modules_registered": moduleManager.GetModuleCount(),
		"module_ids":         moduleManager.GetModuleIDs(),
	})

	return router, nil
}

// CreateDefaultIntelligentRouter creates an intelligent router with default configuration
func (rf *RouterFactory) CreateDefaultIntelligentRouter(
	supabaseClient *database.SupabaseClient,
) (*IntelligentRouter, error) {
	config := IntelligentRouterConfig{
		EnableRequestAnalysis:    true,
		EnableModuleSelection:    true,
		EnableParallelProcessing: true,
		EnableRetryLogic:         true,
		EnableFallbackProcessing: true,
		MaxConcurrentRequests:    10,
		MaxParallelModules:       5,
		WorkerPoolSize:           5,
		RequestTimeout:           30 * time.Second,
		RetryAttempts:            3,
		RetryDelay:               1 * time.Second,
		FallbackTimeout:          10 * time.Second,
		EnableMetricsCollection:  true,
		ParallelProcessingMode:   ParallelProcessingModeHybrid,
	}

	return rf.CreateIntelligentRouterWithDatabaseClassification(supabaseClient, config)
}
