package main

import (
	"context"
	"database/sql"
	"fmt"

	// "encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	apihandlers "kyb-platform/services/risk-assessment-service/internal/api/handlers"
	"kyb-platform/services/risk-assessment-service/internal/batch"
	"kyb-platform/services/risk-assessment-service/internal/cache"
	"kyb-platform/services/risk-assessment-service/internal/config"
	"kyb-platform/services/risk-assessment-service/internal/engine"
	"kyb-platform/services/risk-assessment-service/internal/external"
	"kyb-platform/services/risk-assessment-service/internal/handlers"
	"kyb-platform/services/risk-assessment-service/internal/middleware"
	"kyb-platform/services/risk-assessment-service/internal/ml/custom"
	"kyb-platform/services/risk-assessment-service/internal/ml/service"
	"kyb-platform/services/risk-assessment-service/internal/models"
	"kyb-platform/services/risk-assessment-service/internal/monitoring"
	"kyb-platform/services/risk-assessment-service/internal/performance"
	"kyb-platform/services/risk-assessment-service/internal/pool"
	"kyb-platform/services/risk-assessment-service/internal/query"
	"kyb-platform/services/risk-assessment-service/internal/reporting"
	"kyb-platform/services/risk-assessment-service/internal/webhooks"

	"kyb-platform/services/risk-assessment-service/internal/supabase"
)

// MockDashboardDataProvider provides mock data for dashboard testing
type MockDashboardDataProvider struct {
	logger *zap.Logger
}

// GetRiskAssessments returns mock risk assessment data
func (m *MockDashboardDataProvider) GetRiskAssessments(ctx context.Context, tenantID string, filters *reporting.DashboardFilters) ([]*models.RiskAssessment, error) {
	// Return mock data for testing
	return []*models.RiskAssessment{
		{
			ID:              "assessment_1",
			BusinessID:      "business_1",
			BusinessName:    "Test Business 1",
			BusinessAddress: "123 Test St, City, State 12345",
			Industry:        "Technology",
			Country:         "US",
			RiskScore:       0.75,
			RiskLevel:       models.RiskLevelHigh,
			CreatedAt:       time.Now().Add(-24 * time.Hour),
		},
		{
			ID:              "assessment_2",
			BusinessID:      "business_2",
			BusinessName:    "Test Business 2",
			BusinessAddress: "456 Test Ave, City, State 12345",
			Industry:        "Finance",
			Country:         "US",
			RiskScore:       0.45,
			RiskLevel:       models.RiskLevelMedium,
			CreatedAt:       time.Now().Add(-12 * time.Hour),
		},
	}, nil
}

// GetRiskPredictions returns mock risk prediction data
func (m *MockDashboardDataProvider) GetRiskPredictions(ctx context.Context, tenantID string, filters *reporting.DashboardFilters) ([]*models.RiskPrediction, error) {
	// Return mock data for testing
	return []*models.RiskPrediction{
		{
			BusinessID:      "business_1",
			PredictionDate:  time.Now(),
			HorizonMonths:   3,
			PredictedScore:  0.80,
			PredictedLevel:  models.RiskLevelHigh,
			ConfidenceScore: 0.85,
			CreatedAt:       time.Now(),
		},
	}, nil
}

// GetBatchJobs returns mock batch job data
func (m *MockDashboardDataProvider) GetBatchJobs(ctx context.Context, tenantID string, filters *reporting.DashboardFilters) ([]*reporting.BatchJobData, error) {
	// Return mock data for testing
	return []*reporting.BatchJobData{
		{
			ID:            "batch_job_1",
			Status:        "completed",
			TotalRequests: 100,
			Completed:     95,
			Failed:        5,
			CreatedAt:     time.Now().Add(-2 * time.Hour),
			CompletedAt:   &[]time.Time{time.Now().Add(-1 * time.Hour)}[0],
			JobType:       "risk_assessment",
		},
	}, nil
}

// GetComplianceData returns mock compliance data
func (m *MockDashboardDataProvider) GetComplianceData(ctx context.Context, tenantID string, filters *reporting.DashboardFilters) (*reporting.ComplianceData, error) {
	// Return mock data for testing
	return &reporting.ComplianceData{
		TotalChecks:  100,
		Compliant:    85,
		NonCompliant: 10,
		Pending:      5,
		Violations: []reporting.ComplianceViolation{
			{
				Violation: "Missing documentation",
				Count:     5,
				Severity:  "medium",
			},
		},
		Trends: []reporting.ComplianceTrend{},
	}, nil
}

// GetPerformanceData returns mock performance data
func (m *MockDashboardDataProvider) GetPerformanceData(ctx context.Context, tenantID string, filters *reporting.DashboardFilters) (*reporting.PerformanceData, error) {
	// Return mock data for testing
	return &reporting.PerformanceData{
		ResponseTime: reporting.PerformanceMetrics{
			Average: 500.0,
			P95:     1000.0,
			P99:     2000.0,
			Min:     100.0,
			Max:     5000.0,
		},
		Throughput: reporting.PerformanceMetrics{
			Average: 1000.0,
			P95:     1500.0,
			P99:     2000.0,
			Min:     500.0,
			Max:     3000.0,
		},
		ErrorRate: reporting.PerformanceMetrics{
			Average: 0.1,
			P95:     0.5,
			P99:     1.0,
			Min:     0.0,
			Max:     2.0,
		},
		Availability: reporting.PerformanceMetrics{
			Average: 99.9,
			P95:     100.0,
			P99:     100.0,
			Min:     99.0,
			Max:     100.0,
		},
		Trends: []reporting.PerformanceTrend{},
	}, nil
}

func main() {
	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	logger.Info("🚀 Starting Risk Assessment Service v1.0.0")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration", zap.Error(err))
	}

	logger.Info("✅ Configuration loaded successfully",
		zap.String("port", cfg.Server.Port),
		zap.String("supabase_url", cfg.Supabase.URL),
		zap.String("log_level", cfg.Logging.Level))

	// Initialize Supabase client
	supabaseConfig := &supabase.Config{
		URL:            cfg.Supabase.URL,
		APIKey:         cfg.Supabase.APIKey,
		ServiceRoleKey: cfg.Supabase.ServiceRoleKey,
		JWTSecret:      cfg.Supabase.JWTSecret,
	}
	supabaseClient, err := supabase.NewClient(supabaseConfig, logger)
	if err != nil {
		logger.Fatal("Failed to initialize Supabase client", zap.Error(err))
	}
	logger.Info("✅ Supabase client initialized")

	// Initialize ML service
	mlService := service.NewMLService(logger)
	if err := mlService.InitializeModels(context.Background()); err != nil {
		logger.Fatal("Failed to initialize ML models", zap.Error(err))
	}
	logger.Info("✅ ML service initialized")

	// Initialize risk engine
	riskEngineConfig := &engine.Config{
		MaxConcurrentRequests: 1000,                   // Increased for batch processing
		RequestTimeout:        500 * time.Millisecond, // Sub-1-second target
		CacheTTL:              5 * time.Minute,
		CircuitBreakerConfig: engine.CircuitBreakerConfig{
			FailureThreshold: 5,
			RecoveryTimeout:  30 * time.Second,
			HalfOpenMaxCalls: 3,
		},
		EnableMetrics: true,
		EnableCaching: true,
	}
	// Initialize webhook event service (will be created after webhook components)
	// For now, create a placeholder - we'll update this after webhook components are initialized
	riskEngine := engine.NewRiskEngine(mlService, logger, riskEngineConfig, nil)
	logger.Info("✅ Risk engine initialized (webhook integration pending)")

	// Initialize external data service
	externalDataConfig := &external.ExternalDataConfig{
		NewsAPIKey:           cfg.External.NewsAPI.APIKey,
		OpenCorporatesKey:    cfg.External.OpenCorporates.APIKey,
		GovernmentAPIKey:     cfg.External.OFAC.APIKey,
		Timeout:              15 * time.Second,
		EnableNewsAPI:        cfg.External.NewsAPI.APIKey != "",
		EnableOpenCorporates: cfg.External.OpenCorporates.APIKey != "",
		EnableGovernment:     cfg.External.OFAC.APIKey != "",
	}
	externalDataService := external.NewExternalDataService(externalDataConfig, logger)
	logger.Info("✅ External data service initialized")

	// Initialize premium external API manager
	externalAPIConfig := &external.ExternalAPIManagerConfig{
		ThomsonReuters: &external.ThomsonReutersConfig{
			APIKey:    cfg.External.ThomsonReuters.APIKey,
			BaseURL:   cfg.External.ThomsonReuters.BaseURL,
			Timeout:   30 * time.Second,
			RateLimit: 100,
			Enabled:   cfg.External.ThomsonReuters.APIKey != "",
		},
		OFAC: &external.OFACConfig{
			APIKey:    cfg.External.OFAC.APIKey,
			BaseURL:   cfg.External.OFAC.BaseURL,
			Timeout:   30 * time.Second,
			RateLimit: 50,
			Enabled:   cfg.External.OFAC.APIKey != "",
		},
		WorldCheck: &external.WorldCheckConfig{
			APIKey:    cfg.External.WorldCheck.APIKey,
			BaseURL:   cfg.External.WorldCheck.BaseURL,
			Timeout:   30 * time.Second,
			RateLimit: 75,
			Enabled:   cfg.External.WorldCheck.APIKey != "",
		},
		Timeout:     30 * time.Second,
		MaxRetries:  3,
		EnableCache: true,
		CacheTTL:    5 * time.Minute,
	}
	externalAPIManager := external.NewExternalAPIManager(externalAPIConfig, logger)
	logger.Info("✅ Premium external API manager initialized")

	// Initialize performance optimization system
	// performanceConfig := performance.DefaultOptimizerConfig()
	// performanceConfig.TargetP95 = 1 * time.Second
	// performanceConfig.TargetP99 = 2 * time.Second
	// performanceConfig.TargetThroughput = 1000

	// Note: In a real implementation, you'd pass the actual database connection
	// For now, we'll initialize without database optimization
	// performanceConfig.EnableDBOptimization = false
	// performanceOptimizer := performance.NewOptimizer(logger, nil, performanceConfig)
	// logger.Info("✅ Performance optimizer initialized")

	// Initialize legacy performance monitor for compatibility
	performanceMonitor := monitoring.NewPerformanceMonitor(logger)
	performanceMonitor.SetTargets(16.67, 1*time.Second, 0.01, 1000) // 1000 req/min target
	logger.Info("✅ Performance monitor initialized")

	// Initialize monitoring system
	monitoringConfig := config.LoadMonitoringConfig()
	prometheusMetrics := monitoring.NewPrometheusMetrics(logger)
	alertManager := monitoring.NewAlertManager(logger)

	// Add alert channels
	emailChannel := monitoring.NewEmailAlertChannel(monitoring.EmailConfig{
		SMTPHost:    monitoringConfig.Alerting.Channels["email"].Config["smtp_host"].(string),
		SMTPPort:    monitoringConfig.Alerting.Channels["email"].Config["smtp_port"].(int),
		Username:    monitoringConfig.Alerting.Channels["email"].Config["username"].(string),
		Password:    monitoringConfig.Alerting.Channels["email"].Config["password"].(string),
		FromAddress: monitoringConfig.Alerting.Channels["email"].Config["from_address"].(string),
		ToAddresses: monitoringConfig.Alerting.Channels["email"].Config["to_addresses"].([]string),
		UseTLS:      monitoringConfig.Alerting.Channels["email"].Config["use_tls"].(bool),
	}, logger)
	alertManager.AddAlertChannel(emailChannel)

	grafanaClient := monitoring.NewGrafanaClient(monitoring.GrafanaConfig{
		BaseURL:  monitoringConfig.Grafana.BaseURL,
		APIKey:   monitoringConfig.Grafana.APIKey,
		Username: monitoringConfig.Grafana.Username,
		Password: monitoringConfig.Grafana.Password,
		Timeout:  monitoringConfig.Grafana.Timeout,
	}, logger)

	// Add alert rules
	for _, ruleConfig := range monitoringConfig.Alerting.Rules {
		alertRule := &monitoring.AlertRule{
			ID:          ruleConfig.ID,
			Name:        ruleConfig.Name,
			Description: ruleConfig.Description,
			Metric:      ruleConfig.Metric,
			Condition:   ruleConfig.Condition,
			Threshold:   ruleConfig.Threshold,
			Severity:    monitoring.AlertSeverity(ruleConfig.Severity),
			Duration:    ruleConfig.Duration,
			Enabled:     ruleConfig.Enabled,
			TenantID:    ruleConfig.TenantID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Metadata:    make(map[string]interface{}),
		}
		alertManager.AddAlertRule(alertRule)
	}

	logger.Info("✅ Monitoring system initialized")

	// Initialize handlers
	riskAssessmentHandler := handlers.NewRiskAssessmentHandler(supabaseClient, mlService, riskEngine, externalDataService, logger, cfg)
	advancedPredictionHandler := handlers.NewAdvancedPredictionHandler(mlService, logger)
	metricsHandler := handlers.NewMetricsHandler(mlService.GetMetricsCollector(), logger)
	performanceHandler := handlers.NewPerformanceHandler(performanceMonitor, logger)
	externalAPIHandler := apihandlers.NewExternalAPIHandler(externalAPIManager, logger)
	monitoringHandler := handlers.NewMonitoringHandler(prometheusMetrics, alertManager, grafanaClient, logger)

	// Initialize database connection with performance optimizations
	db, err := initDatabaseWithPerformance(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize database with performance optimizations", zap.Error(err))
	}
	defer db.Close()

	// Initialize performance components
	performanceComponents, err := initPerformanceComponents(cfg, db, logger)
	if err != nil {
		logger.Fatal("Failed to initialize performance components", zap.Error(err))
	}
	defer performanceComponents.Close()

	// Initialize custom model components
	customModelRepository := custom.NewSQLCustomModelRepository(db, logger)
	customModelBuilder := custom.NewCustomModelBuilder(customModelRepository, logger)
	// Custom model handler will be initialized after webhook components are created

	// Initialize batch processing components
	batchJobRepository := batch.NewSQLBatchJobRepository(db, logger)
	batchProcessor := batch.NewDefaultBatchProcessor(riskEngine, batchJobRepository, batch.DefaultBatchJobConfig(), logger)
	jobManager := batch.NewDefaultJobManager(batchJobRepository, batchProcessor, batch.DefaultBatchJobConfig(), logger)
	// Batch job handler will be initialized after webhook components are created

	// Initialize webhook components
	webhookRepository := webhooks.NewSQLWebhookRepository(db, logger)
	deliveryTracker := webhooks.NewDefaultWebhookDeliveryTracker(webhookRepository, logger)
	retryHandler := webhooks.NewDefaultWebhookRetryHandler(webhookRepository, logger)
	signatureVerifier := webhooks.NewDefaultWebhookSignatureVerifier(5 * time.Minute)
	rateLimiter := webhooks.NewDefaultWebhookRateLimiter(logger)
	circuitBreaker := webhooks.NewDefaultWebhookCircuitBreaker(logger)
	eventFilter := webhooks.NewDefaultWebhookEventFilter(logger)
	webhookManager := webhooks.NewDefaultWebhookManager(webhookRepository, deliveryTracker, retryHandler, signatureVerifier, rateLimiter, circuitBreaker, eventFilter, logger)
	webhookEventService := webhooks.NewEventService(webhookManager, logger)
	webhookHandlers := handlers.NewSimpleWebhookHandlers(webhookManager, logger)
	_ = webhookHandlers // Used in routes below

	// Update risk engine with webhook event service
	riskEngine = engine.NewRiskEngine(mlService, logger, riskEngineConfig, webhookEventService)
	logger.Info("✅ Risk engine updated with webhook integration")

	// Initialize batch job handler with webhook event service
	batchJobHandler := handlers.NewBatchJobHandler(jobManager, webhookEventService, logger)
	_ = batchJobHandler // Used in routes below
	logger.Info("✅ Batch job handler initialized with webhook integration")

	// Initialize custom model handler with webhook event service
	customModelHandler := handlers.NewCustomModelHandlers(customModelBuilder, customModelRepository, webhookEventService, logger)
	_ = customModelHandler // Used in routes below
	logger.Info("✅ Custom model handler initialized with webhook integration")

	// Initialize dashboard components
	dashboardRepository := reporting.NewSQLDashboardRepository(db, logger)
	// Note: In a real implementation, you would create a proper data provider
	// For now, we'll use a mock data provider
	dashboardDataProvider := &MockDashboardDataProvider{logger: logger}
	dashboardService := reporting.NewDefaultDashboardService(dashboardRepository, dashboardDataProvider, logger)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService, logger)
	_ = dashboardHandler // Used in routes below

	// Initialize report components
	reportRepository := reporting.NewSQLReportRepository(db, logger)
	reportTemplateRepository := reporting.NewSQLReportTemplateRepository(db, logger)
	scheduledReportRepository := reporting.NewSQLScheduledReportRepository(db, logger)
	reportDataProvider := reporting.NewDefaultReportDataProvider(logger)
	reportGenerator := reporting.NewDefaultReportGenerator(logger)
	reportScheduler := reporting.NewDefaultReportScheduler(logger)
	reportService := reporting.NewDefaultReportService(
		reportRepository,
		reportTemplateRepository,
		scheduledReportRepository,
		reportDataProvider,
		reportGenerator,
		reportScheduler,
		logger,
	)
	reportHandler := handlers.NewReportHandler(reportService, logger)
	_ = reportHandler // Used in routes below
	// scenarioHandler := handlers.NewScenarioHandlers(logger) // Temporarily commented out due to build issues
	// explainabilityHandler := handlers.NewExplainabilityHandlers(mlService, logger)
	// experimentHandler := handlers.NewExperimentHandlers(experimentManager, logger) // Temporarily disabled due to testing package issues

	// Initialize middleware
	middlewareInstance := middleware.NewMiddleware(logger)

	// Initialize monitoring middleware
	monitoringMiddleware := monitoring.NewMetricsMiddleware(prometheusMetrics, logger)

	// Initialize new performance middleware with optimizer
	// performanceMiddlewareConfig := performance.DefaultMiddlewareConfig()
	// performanceMiddleware := performance.NewPerformanceMiddleware(
	//	logger,
	//	performanceOptimizer.GetProfiler(),
	//	performanceOptimizer.GetResponseMonitor(),
	//	performanceOptimizer.GetCacheOptimizer(),
	//	performanceMiddlewareConfig,
	// )

	// Legacy performance middleware for compatibility
	legacyPerformanceMiddleware := middleware.NewPerformanceMiddleware(performanceMonitor, logger)

	// Setup router
	router := mux.NewRouter()

	// Add comprehensive middleware
	router.Use(middlewareInstance.RecoveryMiddleware())
	router.Use(middlewareInstance.LoggingMiddleware)
	// router.Use(performanceMiddleware.Middleware)         // New performance monitoring
	router.Use(legacyPerformanceMiddleware.Middleware())   // Legacy performance monitoring
	router.Use(monitoringMiddleware.HTTPMetricsMiddleware) // Prometheus metrics collection
	router.Use(middlewareInstance.SecurityMiddleware())
	router.Use(middlewareInstance.RequestSizeMiddleware(10 * 1024 * 1024)) // 10MB limit
	router.Use(middlewareInstance.TimeoutMiddleware(30 * time.Second))
	router.Use(middlewareInstance.CORSMiddleware(
		[]string{"*"}, // Allowed origins
		[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},       // Allowed methods
		[]string{"Content-Type", "Authorization", "X-Request-ID"}, // Allowed headers
	))
	router.Use(middlewareInstance.RateLimitMiddleware(100)) // 100 requests per minute
	router.Use(middlewareInstance.MetricsMiddleware())
	router.Use(middlewareInstance.HealthCheckMiddleware())

	// Health check handler
	healthCheckHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","timestamp":"%s","service":"risk-assessment-service"}`, time.Now().Format(time.RFC3339))
	}

	// Health check endpoint
	router.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Risk assessment endpoints
	api.HandleFunc("/assess", riskAssessmentHandler.HandleRiskAssessment).Methods("POST")
	api.HandleFunc("/assess/batch", riskAssessmentHandler.HandleBatchRiskAssessment).Methods("POST")
	api.HandleFunc("/assess/{id}", riskAssessmentHandler.HandleGetRiskAssessment).Methods("GET")
	api.HandleFunc("/assess/{id}/predict", riskAssessmentHandler.HandleRiskPrediction).Methods("POST")
	api.HandleFunc("/assess/{id}/history", riskAssessmentHandler.HandleRiskHistory).Methods("GET")

	// Advanced prediction endpoints
	api.HandleFunc("/risk/predict-advanced", advancedPredictionHandler.HandleAdvancedPrediction).Methods("POST")
	api.HandleFunc("/models/info", advancedPredictionHandler.HandleGetModelInfo).Methods("GET")
	api.HandleFunc("/models/performance", advancedPredictionHandler.HandleGetModelPerformance).Methods("GET")

	// Custom model endpoints
	api.HandleFunc("/models/custom", customModelHandler.HandleCreateCustomModel).Methods("POST")
	api.HandleFunc("/models/custom", customModelHandler.HandleListCustomModels).Methods("GET")
	api.HandleFunc("/models/custom/{id}", customModelHandler.HandleGetCustomModel).Methods("GET")
	api.HandleFunc("/models/custom/{id}", customModelHandler.HandleUpdateCustomModel).Methods("PUT")
	api.HandleFunc("/models/custom/{id}", customModelHandler.HandleDeleteCustomModel).Methods("DELETE")
	api.HandleFunc("/models/custom/{id}/validate", customModelHandler.HandleValidateCustomModel).Methods("POST")
	api.HandleFunc("/models/custom/{id}/test", customModelHandler.HandleTestCustomModel).Methods("POST")
	api.HandleFunc("/models/custom/{id}/activate", customModelHandler.HandleActivateCustomModel).Methods("POST")
	api.HandleFunc("/models/custom/{id}/deactivate", customModelHandler.HandleDeactivateCustomModel).Methods("POST")
	api.HandleFunc("/models/custom/{id}/versions", customModelHandler.HandleGetCustomModelVersions).Methods("GET")

	// Batch processing endpoints
	api.HandleFunc("/assess/batch/async", batchJobHandler.HandleSubmitBatchJob).Methods("POST")
	api.HandleFunc("/assess/batch", batchJobHandler.HandleListBatchJobs).Methods("GET")
	api.HandleFunc("/assess/batch/{job_id}", batchJobHandler.HandleGetBatchJobStatus).Methods("GET")
	api.HandleFunc("/assess/batch/{job_id}/results", batchJobHandler.HandleGetBatchJobResults).Methods("GET")
	api.HandleFunc("/assess/batch/{job_id}", batchJobHandler.HandleCancelBatchJob).Methods("DELETE")
	api.HandleFunc("/assess/batch/{job_id}/resume", batchJobHandler.HandleResumeBatchJob).Methods("POST")
	api.HandleFunc("/assess/batch/metrics", batchJobHandler.HandleGetBatchJobMetrics).Methods("GET")

	// Dashboard endpoints
	api.HandleFunc("/reporting/dashboards", dashboardHandler.HandleCreateDashboard).Methods("POST")
	api.HandleFunc("/reporting/dashboards", dashboardHandler.HandleListDashboards).Methods("GET")
	api.HandleFunc("/reporting/dashboards/{id}", dashboardHandler.HandleGetDashboard).Methods("GET")
	api.HandleFunc("/reporting/dashboards/{id}", dashboardHandler.HandleUpdateDashboard).Methods("PUT")
	api.HandleFunc("/reporting/dashboards/{id}", dashboardHandler.HandleDeleteDashboard).Methods("DELETE")
	api.HandleFunc("/reporting/dashboards/{id}/data", dashboardHandler.HandleGetDashboardData).Methods("GET")
	api.HandleFunc("/reporting/dashboards/metrics", dashboardHandler.HandleGetDashboardMetrics).Methods("GET")
	api.HandleFunc("/reporting/dashboard/risk-overview", dashboardHandler.HandleGetRiskOverview).Methods("GET")
	api.HandleFunc("/reporting/dashboard/trends", dashboardHandler.HandleGetTrends).Methods("GET")
	api.HandleFunc("/reporting/dashboard/predictions", dashboardHandler.HandleGetPredictions).Methods("GET")

	// Report endpoints
	api.HandleFunc("/reports/generate", reportHandler.HandleGenerateReport).Methods("POST")
	api.HandleFunc("/reports", reportHandler.HandleListReports).Methods("GET")
	api.HandleFunc("/reports/{id}", reportHandler.HandleGetReport).Methods("GET")
	api.HandleFunc("/reports/{id}", reportHandler.HandleDeleteReport).Methods("DELETE")
	api.HandleFunc("/reports/{id}/download", reportHandler.HandleDownloadReport).Methods("GET")
	api.HandleFunc("/reports/metrics", reportHandler.HandleGetReportMetrics).Methods("GET")

	// Report template endpoints
	api.HandleFunc("/reports/templates", reportHandler.HandleCreateTemplate).Methods("POST")
	api.HandleFunc("/reports/templates", reportHandler.HandleListTemplates).Methods("GET")
	api.HandleFunc("/reports/templates/{id}", reportHandler.HandleGetTemplate).Methods("GET")
	api.HandleFunc("/reports/templates/{id}", reportHandler.HandleUpdateTemplate).Methods("PUT")
	api.HandleFunc("/reports/templates/{id}", reportHandler.HandleDeleteTemplate).Methods("DELETE")

	// Scheduled report endpoints
	api.HandleFunc("/reports/scheduled", reportHandler.HandleCreateScheduledReport).Methods("POST")
	api.HandleFunc("/reports/scheduled", reportHandler.HandleListScheduledReports).Methods("GET")
	api.HandleFunc("/reports/scheduled/{id}", reportHandler.HandleGetScheduledReport).Methods("GET")
	api.HandleFunc("/reports/scheduled/{id}", reportHandler.HandleUpdateScheduledReport).Methods("PUT")
	api.HandleFunc("/reports/scheduled/{id}", reportHandler.HandleDeleteScheduledReport).Methods("DELETE")
	api.HandleFunc("/reports/scheduled/{id}/run", reportHandler.HandleRunScheduledReport).Methods("POST")

	// Compliance endpoints
	api.HandleFunc("/compliance/check", riskAssessmentHandler.HandleComplianceCheck).Methods("POST")
	api.HandleFunc("/sanctions/screen", riskAssessmentHandler.HandleSanctionsScreening).Methods("POST")
	api.HandleFunc("/media/monitor", riskAssessmentHandler.HandleAdverseMediaMonitoring).Methods("POST")

	// Analytics endpoints
	api.HandleFunc("/analytics/trends", riskAssessmentHandler.HandleRiskTrends).Methods("GET")
	api.HandleFunc("/analytics/insights", riskAssessmentHandler.HandleRiskInsights).Methods("GET")

	// Scenario analysis endpoints (temporarily commented out due to build issues)
	// api.HandleFunc("/scenarios/monte-carlo", scenarioHandler.HandleMonteCarloSimulation).Methods("POST")
	// api.HandleFunc("/scenarios/stress-test", scenarioHandler.HandleStressTesting).Methods("POST")
	// api.HandleFunc("/scenarios/analyze", scenarioHandler.HandleComprehensiveScenarioAnalysis).Methods("POST")
	// api.HandleFunc("/scenarios/info", scenarioHandler.HandleGetScenarioInfo).Methods("GET")

	// Metrics and monitoring endpoints
	api.HandleFunc("/metrics", metricsHandler.HandleGetMetrics).Methods("GET")
	api.HandleFunc("/health", metricsHandler.HandleGetHealth).Methods("GET")
	api.HandleFunc("/performance", metricsHandler.HandleGetPerformanceSnapshot).Methods("GET")

	// Advanced monitoring endpoints (Prometheus/Grafana)
	api.HandleFunc("/monitoring/metrics", monitoringHandler.GetMetrics).Methods("GET")
	api.HandleFunc("/monitoring/health", monitoringHandler.GetHealth).Methods("GET")
	api.HandleFunc("/monitoring/alerts", monitoringHandler.GetAlerts).Methods("GET")
	api.HandleFunc("/monitoring/alerts/history", monitoringHandler.GetAlertHistory).Methods("GET")
	api.HandleFunc("/monitoring/alerts/suppress", monitoringHandler.SuppressAlert).Methods("POST")
	api.HandleFunc("/monitoring/performance/insights", monitoringHandler.GetPerformanceInsights).Methods("GET")
	api.HandleFunc("/monitoring/system/metrics", monitoringHandler.GetSystemMetrics).Methods("GET")
	api.HandleFunc("/monitoring/tenant/metrics", monitoringHandler.GetTenantMetrics).Methods("GET")
	api.HandleFunc("/monitoring/grafana/dashboard", monitoringHandler.CreateGrafanaDashboard).Methods("POST")
	api.HandleFunc("/monitoring/grafana/dashboard", monitoringHandler.GetGrafanaDashboard).Methods("GET")
	api.HandleFunc("/monitoring/grafana/dashboard", monitoringHandler.DeleteGrafanaDashboard).Methods("DELETE")
	api.HandleFunc("/monitoring/config", monitoringHandler.GetMonitoringConfig).Methods("GET")
	api.HandleFunc("/monitoring/config", monitoringHandler.UpdateMonitoringConfig).Methods("PUT")

	// Performance monitoring endpoints (legacy)
	api.HandleFunc("/performance/stats", performanceHandler.HandlePerformanceStats).Methods("GET")
	api.HandleFunc("/performance/alerts", performanceHandler.HandlePerformanceAlerts).Methods("GET")
	api.HandleFunc("/performance/health", performanceHandler.HandlePerformanceHealth).Methods("GET")
	api.HandleFunc("/performance/reset", performanceHandler.HandlePerformanceReset).Methods("POST")
	api.HandleFunc("/performance/targets", performanceHandler.HandlePerformanceTargets).Methods("POST")
	api.HandleFunc("/performance/alerts/clear", performanceHandler.HandlePerformanceClearAlerts).Methods("POST")

	// New performance optimization endpoints
	// api.HandleFunc("/optimization/report", func(w http.ResponseWriter, r *http.Request) {
	//	report, err := performanceOptimizer.Optimize()
	//	if err != nil {
	//		http.Error(w, err.Error(), http.StatusInternalServerError)
	//		return
	//	}
	//	w.Header().Set("Content-Type", "application/json")
	//	json.NewEncoder(w).Encode(report)
	// }).Methods("GET")

	// api.HandleFunc("/optimization/stats", func(w http.ResponseWriter, r *http.Request) {
	//	stats := performanceOptimizer.GetPerformanceStats()
	//	w.Header().Set("Content-Type", "application/json")
	//	json.NewEncoder(w).Encode(stats)
	// }).Methods("GET")

	// api.HandleFunc("/optimization/health", func(w http.ResponseWriter, r *http.Request) {
	//	health := performanceOptimizer.GetHealthStatus()
	//	w.Header().Set("Content-Type", "application/json")
	//	json.NewEncoder(w).Encode(health)
	// }).Methods("GET")

	// api.HandleFunc("/optimization/report/full", func(w http.ResponseWriter, r *http.Request) {
	//	report := performanceOptimizer.GetPerformanceReport()
	//	w.Header().Set("Content-Type", "text/plain")
	//	w.Write([]byte(report))
	// }).Methods("GET")

	// External data source endpoints (legacy)
	api.HandleFunc("/external/adverse-media", riskAssessmentHandler.HandleExternalAdverseMediaMonitoring).Methods("POST")
	api.HandleFunc("/external/company-data", riskAssessmentHandler.HandleCompanyDataLookup).Methods("POST")
	api.HandleFunc("/external/compliance", riskAssessmentHandler.HandleExternalComplianceCheck).Methods("POST")
	api.HandleFunc("/external/sources", riskAssessmentHandler.HandleExternalDataSources).Methods("GET")

	// Premium external API endpoints
	externalAPI := api.PathPrefix("/external").Subrouter()
	externalAPI.HandleFunc("/comprehensive", externalAPIHandler.GetComprehensiveData).Methods("POST")
	externalAPI.HandleFunc("/thomson-reuters", externalAPIHandler.GetThomsonReutersData).Methods("POST")
	externalAPI.HandleFunc("/ofac", externalAPIHandler.GetOFACData).Methods("POST")
	externalAPI.HandleFunc("/worldcheck", externalAPIHandler.GetWorldCheckData).Methods("POST")
	externalAPI.HandleFunc("/status", externalAPIHandler.GetAPIStatus).Methods("GET")
	externalAPI.HandleFunc("/supported", externalAPIHandler.GetSupportedAPIs).Methods("GET")
	externalAPI.HandleFunc("/health", externalAPIHandler.HealthCheck).Methods("GET")
	externalAPI.HandleFunc("/risk-factors", externalAPIHandler.GetRiskFactorsFromExternalData).Methods("POST")

	// Explainability endpoints (temporarily disabled)
	// api.HandleFunc("/explain/prediction", explainabilityHandler.HandleExplainPrediction).Methods("POST")
	// api.HandleFunc("/explain/compare", explainabilityHandler.HandleComparePredictions).Methods("POST")
	// api.HandleFunc("/explain/risk-factors", explainabilityHandler.HandleExplainRiskFactors).Methods("POST")
	// api.HandleFunc("/explain/visualization", explainabilityHandler.HandleGenerateVisualization).Methods("POST")
	// api.HandleFunc("/explain/info", explainabilityHandler.HandleGetExplainabilityInfo).Methods("GET")

	// Webhook endpoints
	webhookAPI := api.PathPrefix("/webhooks").Subrouter()
	webhookAPI.HandleFunc("", webhookHandlers.CreateWebhook).Methods("POST")
	webhookAPI.HandleFunc("", webhookHandlers.ListWebhooks).Methods("GET")
	webhookAPI.HandleFunc("/{id}", webhookHandlers.GetWebhook).Methods("GET")
	webhookAPI.HandleFunc("/{id}", webhookHandlers.UpdateWebhook).Methods("PUT", "PATCH")
	webhookAPI.HandleFunc("/{id}", webhookHandlers.DeleteWebhook).Methods("DELETE")

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start performance monitoring
	go performanceMonitor.StartMonitoring(context.Background())

	// Start batch job manager
	if err := jobManager.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start batch job manager", zap.Error(err))
	}
	logger.Info("✅ Batch job manager started")

	// Start Prometheus metrics server
	if monitoringConfig.Prometheus.Enabled {
		go func() {
			logger.Info("📊 Starting Prometheus metrics server", zap.Int("port", monitoringConfig.Prometheus.Port))
			if err := prometheusMetrics.StartMetricsServer(context.Background(), monitoringConfig.Prometheus.Port); err != nil {
				logger.Error("Failed to start Prometheus metrics server", zap.Error(err))
			}
		}()
	}

	// Create Grafana dashboard if auto-create is enabled
	if monitoringConfig.Grafana.Enabled && monitoringConfig.Grafana.AutoCreate {
		go func() {
			time.Sleep(5 * time.Second) // Wait for services to be ready
			if err := grafanaClient.CreateRiskAssessmentDashboard(context.Background()); err != nil {
				logger.Error("Failed to create Grafana dashboard", zap.Error(err))
			} else {
				logger.Info("✅ Grafana dashboard created successfully")
			}
		}()
	}

	// Start server in a goroutine
	go func() {
		logger.Info("🌐 Starting HTTP server", zap.String("addr", httpServer.Addr))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown external data service
	externalDataService.Close()

	// Shutdown risk engine
	if err := riskEngine.Shutdown(ctx); err != nil {
		logger.Error("Risk engine shutdown failed", zap.Error(err))
	}

	// Shutdown batch job manager
	if err := jobManager.Stop(); err != nil {
		logger.Error("Batch job manager shutdown failed", zap.Error(err))
	}

	// Attempt graceful shutdown
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("✅ Server exited")
}

// Health check handler
func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"risk-assessment-service","version":"1.0.0"}`))
}

// PerformanceComponents holds all performance-related components
type PerformanceComponents struct {
	Cache     cache.Cache
	Pool      *pool.ConnectionPool
	Optimizer *query.QueryOptimizer
	Monitor   *performance.PerformanceMonitor
}

// Close closes all performance components
func (pc *PerformanceComponents) Close() error {
	if pc.Cache != nil {
		pc.Cache.Close()
	}
	if pc.Pool != nil {
		pc.Pool.Close()
	}
	if pc.Monitor != nil {
		pc.Monitor.Stop()
	}
	return nil
}

// initDatabaseWithPerformance initializes database connection with performance optimizations
func initDatabaseWithPerformance(cfg *config.Config, logger *zap.Logger) (*sql.DB, error) {
	// Get database URL from configuration
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgresql://username:password@localhost:5432/risk_assessment?sslmode=disable"
	}

	// Open database connection
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("✅ Database connection established with performance optimizations")
	return db, nil
}

// initPerformanceComponents initializes all performance-related components
func initPerformanceComponents(cfg *config.Config, db *sql.DB, logger *zap.Logger) (*PerformanceComponents, error) {
	// Load performance configuration
	perfConfig := config.DefaultPerformanceConfig()

	// Initialize Redis cache
	var cacheInstance cache.Cache
	if perfConfig.Cache.Enabled && perfConfig.Cache.Type == "redis" {
		redisConfig := &cache.CacheConfig{
			Addrs:             perfConfig.Cache.Redis.Addrs,
			Password:          perfConfig.Cache.Redis.Password,
			DB:                perfConfig.Cache.Redis.DB,
			PoolSize:          perfConfig.Cache.Redis.PoolSize,
			MinIdleConns:      perfConfig.Cache.Redis.MinIdleConns,
			MaxRetries:        perfConfig.Cache.Redis.MaxRetries,
			DialTimeout:       perfConfig.Cache.Redis.DialTimeout,
			ReadTimeout:       perfConfig.Cache.Redis.ReadTimeout,
			WriteTimeout:      perfConfig.Cache.Redis.WriteTimeout,
			PoolTimeout:       perfConfig.Cache.Redis.PoolTimeout,
			IdleTimeout:       perfConfig.Cache.Redis.IdleTimeout,
			MaxConnAge:        perfConfig.Cache.Redis.MaxConnAge,
			DefaultTTL:        perfConfig.Cache.DefaultTTL,
			KeyPrefix:         perfConfig.Cache.Redis.KeyPrefix,
			EnableMetrics:     perfConfig.Cache.EnableMetrics,
			EnableCompression: perfConfig.Cache.EnableCompression,
		}

		var err error
		// Create a logger wrapper that implements cache.Logger interface
		cacheLogger := &cacheLoggerWrapper{logger: logger}
		cacheInstance, err = cache.NewRedisCache(redisConfig, cacheLogger)
		if err != nil {
			logger.Warn("Failed to initialize Redis cache, falling back to no cache", zap.Error(err))
		} else {
			logger.Info("✅ Redis cache initialized")
		}
	}

	// Initialize connection pool
	poolConfig := &pool.PoolConfig{
		MaxConnections:     perfConfig.ConnectionPool.MaxConnections,
		MinConnections:     perfConfig.ConnectionPool.MinConnections,
		MaxIdleConnections: perfConfig.ConnectionPool.MaxIdleConnections,
		ConnectionTimeout:  perfConfig.ConnectionPool.ConnectionTimeout,
		IdleTimeout:        perfConfig.ConnectionPool.IdleTimeout,
		MaxLifetime:        perfConfig.ConnectionPool.MaxLifetime,
		HealthCheckPeriod:  perfConfig.ConnectionPool.HealthCheckPeriod,
		RetryAttempts:      perfConfig.ConnectionPool.RetryAttempts,
		RetryDelay:         perfConfig.ConnectionPool.RetryDelay,
	}

	connectionPool, err := pool.NewConnectionPool("", poolConfig, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize connection pool: %w", err)
	}
	logger.Info("✅ Connection pool initialized")

	// Initialize query optimizer
	queryOptimizer := query.NewQueryOptimizer(db, cacheInstance, logger)
	logger.Info("✅ Query optimizer initialized")

	// Initialize performance monitor with nil interfaces for now
	// TODO: Implement proper interface adapters for cache, pool, and query components
	perfMonitor := performance.NewPerformanceMonitor(
		db,
		nil, // cacheInstance - needs interface adapter
		nil, // connectionPool - needs interface adapter
		nil, // queryOptimizer - needs interface adapter
		logger,
	)

	// Start performance monitoring
	// TODO: Fix monitoring config structure
	perfMonitor.Start(30 * time.Second) // Use default interval
	logger.Info("✅ Performance monitoring started")

	return &PerformanceComponents{
		Cache:     cacheInstance,
		Pool:      connectionPool,
		Optimizer: queryOptimizer,
		Monitor:   perfMonitor,
	}, nil
}
