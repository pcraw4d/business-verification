package routes

import (
	"net/http"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/api/middleware"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/routing"
	"go.opentelemetry.io/otel/trace"
)

// RouteConfig holds configuration for route registration
type RouteConfig struct {
	IntelligentRoutingHandler   *handlers.IntelligentRoutingHandler
	BusinessIntelligenceHandler *handlers.BusinessIntelligenceHandler
	AuthMiddleware              *middleware.AuthMiddleware
	RateLimiter                 *middleware.RateLimiter
	Logger                      *observability.Logger
	EnableEnhancedFeatures      bool
	EnableBackwardCompatibility bool
}

// RegisterRoutes registers all API routes with the given mux
func RegisterRoutes(mux *http.ServeMux, config *RouteConfig) {
	// Register intelligent routing endpoints
	registerIntelligentRoutingRoutes(mux, config)

	// Register enhanced business intelligence endpoints
	if config.EnableEnhancedFeatures {
		registerEnhancedBusinessIntelligenceRoutes(mux, config)
	}

	// Register backward compatibility endpoints
	if config.EnableBackwardCompatibility {
		registerBackwardCompatibilityRoutes(mux, config)
	}
}

// registerIntelligentRoutingRoutes registers the intelligent routing system endpoints
func registerIntelligentRoutingRoutes(mux *http.ServeMux, config *RouteConfig) {
	// Enhanced classification endpoints using intelligent routing
	mux.HandleFunc("POST /v2/classify", config.IntelligentRoutingHandler.ClassifyBusiness)
	mux.HandleFunc("POST /v2/classify/batch", config.IntelligentRoutingHandler.ClassifyBusinessBatch)

	// Health and metrics endpoints for intelligent routing system
	mux.HandleFunc("GET /v2/routing/health", config.IntelligentRoutingHandler.GetRoutingHealth)
	mux.HandleFunc("GET /v2/routing/metrics", config.IntelligentRoutingHandler.GetRoutingMetrics)

	config.Logger.Info("Intelligent routing routes registered", map[string]interface{}{
		"version": "v2",
		"endpoints": []string{
			"POST /v2/classify",
			"POST /v2/classify/batch",
			"GET /v2/routing/health",
			"GET /v2/routing/metrics",
		},
	})
}

// registerEnhancedBusinessIntelligenceRoutes registers enhanced business intelligence endpoints
func registerEnhancedBusinessIntelligenceRoutes(mux *http.ServeMux, config *RouteConfig) {
	// Market Analysis endpoints
	mux.HandleFunc("POST /v2/business-intelligence/market-analysis", config.BusinessIntelligenceHandler.CreateMarketAnalysis)
	mux.HandleFunc("GET /v2/business-intelligence/market-analysis", config.BusinessIntelligenceHandler.GetMarketAnalysis)
	mux.HandleFunc("GET /v2/business-intelligence/market-analyses", config.BusinessIntelligenceHandler.ListMarketAnalyses)
	mux.HandleFunc("POST /v2/business-intelligence/market-analysis/jobs", config.BusinessIntelligenceHandler.CreateMarketAnalysisJob)
	mux.HandleFunc("GET /v2/business-intelligence/market-analysis/jobs", config.BusinessIntelligenceHandler.GetMarketAnalysisJob)
	mux.HandleFunc("GET /v2/business-intelligence/market-analysis/jobs/list", config.BusinessIntelligenceHandler.ListMarketAnalysisJobs)

	// Competitive Analysis endpoints
	mux.HandleFunc("POST /v2/business-intelligence/competitive-analysis", config.BusinessIntelligenceHandler.CreateCompetitiveAnalysis)
	mux.HandleFunc("GET /v2/business-intelligence/competitive-analysis", config.BusinessIntelligenceHandler.GetCompetitiveAnalysis)
	mux.HandleFunc("GET /v2/business-intelligence/competitive-analyses", config.BusinessIntelligenceHandler.ListCompetitiveAnalyses)
	mux.HandleFunc("POST /v2/business-intelligence/competitive-analysis/jobs", config.BusinessIntelligenceHandler.CreateCompetitiveAnalysisJob)
	mux.HandleFunc("GET /v2/business-intelligence/competitive-analysis/jobs", config.BusinessIntelligenceHandler.GetCompetitiveAnalysisJob)
	mux.HandleFunc("GET /v2/business-intelligence/competitive-analysis/jobs/list", config.BusinessIntelligenceHandler.ListCompetitiveAnalysisJobs)

	// Growth Analytics endpoints
	mux.HandleFunc("POST /v2/business-intelligence/growth-analytics", config.BusinessIntelligenceHandler.CreateGrowthAnalytics)
	mux.HandleFunc("GET /v2/business-intelligence/growth-analytics", config.BusinessIntelligenceHandler.GetGrowthAnalytics)
	mux.HandleFunc("GET /v2/business-intelligence/growth-analytics/list", config.BusinessIntelligenceHandler.ListGrowthAnalytics)
	mux.HandleFunc("POST /v2/business-intelligence/growth-analytics/jobs", config.BusinessIntelligenceHandler.CreateGrowthAnalyticsJob)
	mux.HandleFunc("GET /v2/business-intelligence/growth-analytics/jobs", config.BusinessIntelligenceHandler.GetGrowthAnalyticsJob)
	mux.HandleFunc("GET /v2/business-intelligence/growth-analytics/jobs/list", config.BusinessIntelligenceHandler.ListGrowthAnalyticsJobs)

	// Business Intelligence Aggregation endpoints
	mux.HandleFunc("POST /v2/business-intelligence/aggregation", config.BusinessIntelligenceHandler.CreateBusinessIntelligenceAggregation)
	mux.HandleFunc("GET /v2/business-intelligence/aggregation", config.BusinessIntelligenceHandler.GetBusinessIntelligenceAggregation)
	mux.HandleFunc("GET /v2/business-intelligence/aggregations", config.BusinessIntelligenceHandler.ListBusinessIntelligenceAggregations)
	mux.HandleFunc("POST /v2/business-intelligence/aggregation/jobs", config.BusinessIntelligenceHandler.CreateBusinessIntelligenceAggregationJob)
	mux.HandleFunc("GET /v2/business-intelligence/aggregation/jobs", config.BusinessIntelligenceHandler.GetBusinessIntelligenceAggregationJob)
	mux.HandleFunc("GET /v2/business-intelligence/aggregation/jobs/list", config.BusinessIntelligenceHandler.ListBusinessIntelligenceAggregationJobs)

	// Enhanced business intelligence endpoints (legacy compatibility)
	mux.HandleFunc("POST /v2/business-intelligence/enhanced-classify", config.IntelligentRoutingHandler.ClassifyBusiness)
	mux.HandleFunc("POST /v2/business-intelligence/batch-enhanced", config.IntelligentRoutingHandler.ClassifyBusinessBatch)

	// Business intelligence analytics endpoints
	mux.HandleFunc("GET /v2/business-intelligence/analytics", func(w http.ResponseWriter, r *http.Request) {
		// Enhanced analytics endpoint
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"analytics": {
				"total_classifications": 0,
				"average_confidence": 0.0,
				"processing_time_avg": "0ms",
				"success_rate": 0.0
			},
			"message": "Enhanced business intelligence analytics endpoint"
		}`))
	})

	// Business intelligence insights endpoint
	mux.HandleFunc("GET /v2/business-intelligence/insights", func(w http.ResponseWriter, r *http.Request) {
		// Business insights endpoint
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"success": true,
			"insights": {
				"trends": [],
				"patterns": [],
				"recommendations": []
			},
			"message": "Enhanced business intelligence insights endpoint"
		}`))
	})

	config.Logger.Info("Enhanced business intelligence routes registered", map[string]interface{}{
		"version": "v2",
		"endpoints": []string{
			"POST /v2/business-intelligence/market-analysis",
			"GET /v2/business-intelligence/market-analysis",
			"GET /v2/business-intelligence/market-analyses",
			"POST /v2/business-intelligence/market-analysis/jobs",
			"GET /v2/business-intelligence/market-analysis/jobs",
			"GET /v2/business-intelligence/market-analysis/jobs/list",
			"POST /v2/business-intelligence/competitive-analysis",
			"GET /v2/business-intelligence/competitive-analysis",
			"GET /v2/business-intelligence/competitive-analyses",
			"POST /v2/business-intelligence/competitive-analysis/jobs",
			"GET /v2/business-intelligence/competitive-analysis/jobs",
			"GET /v2/business-intelligence/competitive-analysis/jobs/list",
			"POST /v2/business-intelligence/growth-analytics",
			"GET /v2/business-intelligence/growth-analytics",
			"GET /v2/business-intelligence/growth-analytics/list",
			"POST /v2/business-intelligence/growth-analytics/jobs",
			"GET /v2/business-intelligence/growth-analytics/jobs",
			"GET /v2/business-intelligence/growth-analytics/jobs/list",
			"POST /v2/business-intelligence/aggregation",
			"GET /v2/business-intelligence/aggregation",
			"GET /v2/business-intelligence/aggregations",
			"POST /v2/business-intelligence/aggregation/jobs",
			"GET /v2/business-intelligence/aggregation/jobs",
			"GET /v2/business-intelligence/aggregation/jobs/list",
			"POST /v2/business-intelligence/enhanced-classify",
			"POST /v2/business-intelligence/batch-enhanced",
			"GET /v2/business-intelligence/analytics",
			"GET /v2/business-intelligence/insights",
		},
	})
}

// registerBackwardCompatibilityRoutes registers backward compatibility endpoints
func registerBackwardCompatibilityRoutes(mux *http.ServeMux, config *RouteConfig) {
	// Legacy v1 endpoints that route through intelligent routing system
	mux.HandleFunc("POST /v1/classify", config.IntelligentRoutingHandler.ClassifyBusiness)
	mux.HandleFunc("POST /v1/classify/batch", config.IntelligentRoutingHandler.ClassifyBusinessBatch)

	// Legacy health endpoint
	mux.HandleFunc("GET /v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"status": "healthy",
			"version": "v1",
			"message": "Legacy health endpoint - consider upgrading to v2"
		}`))
	})

	config.Logger.Info("Backward compatibility routes registered", map[string]interface{}{
		"version": "v1",
		"endpoints": []string{
			"POST /v1/classify",
			"POST /v1/classify/batch",
			"GET /v1/health",
		},
		"note": "These endpoints route through intelligent routing system for enhanced functionality",
	})
}

// CreateIntelligentRoutingHandler creates and configures the intelligent routing handler
func CreateIntelligentRoutingHandler(
	router *routing.IntelligentRouter,
	logger *observability.Logger,
	metrics *observability.Metrics,
	tracer trace.Tracer,
) *handlers.IntelligentRoutingHandler {
	return handlers.NewIntelligentRoutingHandler(router, logger, metrics, tracer)
}

// CreateBusinessIntelligenceHandler creates and configures the business intelligence handler
func CreateBusinessIntelligenceHandler() *handlers.BusinessIntelligenceHandler {
	return handlers.NewBusinessIntelligenceHandler()
}

// RouteDocumentation provides documentation for all registered routes
func RouteDocumentation() map[string]interface{} {
	return map[string]interface{}{
		"version":     "2.0.0",
		"description": "Enhanced Business Intelligence System API Routes",
		"endpoints": map[string]interface{}{
			"v2": map[string]interface{}{
				"classification": map[string]interface{}{
					"POST /v2/classify": map[string]interface{}{
						"description": "Enhanced single business classification using intelligent routing",
						"features": []string{
							"Intelligent routing to best classification method",
							"Enhanced confidence scoring",
							"Multiple industry code mappings",
							"Real-time processing",
						},
					},
					"POST /v2/classify/batch": map[string]interface{}{
						"description": "Enhanced batch business classification using intelligent routing",
						"features": []string{
							"Batch processing up to 100 businesses",
							"Parallel processing for performance",
							"Partial success handling",
							"Comprehensive error reporting",
						},
					},
				},
				"routing": map[string]interface{}{
					"GET /v2/routing/health": map[string]interface{}{
						"description": "Health check for intelligent routing system",
					},
					"GET /v2/routing/metrics": map[string]interface{}{
						"description": "Performance metrics for intelligent routing system",
					},
				},
				"business_intelligence": map[string]interface{}{
					"POST /v2/business-intelligence/market-analysis": map[string]interface{}{
						"description": "Create and execute market analysis immediately",
						"features": []string{
							"Market size analysis",
							"Market trends identification",
							"Opportunity and threat assessment",
							"Industry benchmarking",
							"Strategic recommendations",
						},
					},
					"GET /v2/business-intelligence/market-analysis": map[string]interface{}{
						"description": "Retrieve specific market analysis by ID",
					},
					"GET /v2/business-intelligence/market-analyses": map[string]interface{}{
						"description": "List all market analyses",
					},
					"POST /v2/business-intelligence/market-analysis/jobs": map[string]interface{}{
						"description": "Create background market analysis job",
					},
					"GET /v2/business-intelligence/market-analysis/jobs": map[string]interface{}{
						"description": "Get market analysis job status",
					},
					"GET /v2/business-intelligence/market-analysis/jobs/list": map[string]interface{}{
						"description": "List all market analysis jobs",
					},
					"POST /v2/business-intelligence/competitive-analysis": map[string]interface{}{
						"description": "Create and execute competitive analysis immediately",
						"features": []string{
							"Competitor analysis",
							"Market position assessment",
							"Competitive gap identification",
							"Advantage and threat analysis",
							"Strategic recommendations",
						},
					},
					"GET /v2/business-intelligence/competitive-analysis": map[string]interface{}{
						"description": "Retrieve specific competitive analysis by ID",
					},
					"GET /v2/business-intelligence/competitive-analyses": map[string]interface{}{
						"description": "List all competitive analyses",
					},
					"POST /v2/business-intelligence/competitive-analysis/jobs": map[string]interface{}{
						"description": "Create background competitive analysis job",
					},
					"GET /v2/business-intelligence/competitive-analysis/jobs": map[string]interface{}{
						"description": "Get competitive analysis job status",
					},
					"GET /v2/business-intelligence/competitive-analysis/jobs/list": map[string]interface{}{
						"description": "List all competitive analysis jobs",
					},
					"POST /v2/business-intelligence/growth-analytics": map[string]interface{}{
						"description": "Create and execute growth analytics analysis immediately",
						"features": []string{
							"Growth trend analysis",
							"Growth projections",
							"Growth driver identification",
							"Growth barrier assessment",
							"Growth opportunity analysis",
							"Strategic recommendations",
						},
					},
					"GET /v2/business-intelligence/growth-analytics": map[string]interface{}{
						"description": "Retrieve specific growth analytics analysis by ID",
					},
					"GET /v2/business-intelligence/growth-analytics/list": map[string]interface{}{
						"description": "List all growth analytics analyses",
					},
					"POST /v2/business-intelligence/growth-analytics/jobs": map[string]interface{}{
						"description": "Create background growth analytics job",
					},
					"GET /v2/business-intelligence/growth-analytics/jobs": map[string]interface{}{
						"description": "Get growth analytics job status",
					},
					"GET /v2/business-intelligence/growth-analytics/jobs/list": map[string]interface{}{
						"description": "List all growth analytics jobs",
					},
					"POST /v2/business-intelligence/aggregation": map[string]interface{}{
						"description": "Create comprehensive business intelligence aggregation report",
						"features": []string{
							"Multi-analysis aggregation",
							"Cross-analysis insights",
							"Comprehensive recommendations",
							"Executive summary",
							"Strategic priorities",
						},
					},
					"GET /v2/business-intelligence/aggregation": map[string]interface{}{
						"description": "Retrieve specific business intelligence aggregation by ID",
					},
					"GET /v2/business-intelligence/aggregations": map[string]interface{}{
						"description": "List all business intelligence aggregations",
					},
					"POST /v2/business-intelligence/aggregation/jobs": map[string]interface{}{
						"description": "Create background business intelligence aggregation job",
					},
					"GET /v2/business-intelligence/aggregation/jobs": map[string]interface{}{
						"description": "Get business intelligence aggregation job status",
					},
					"GET /v2/business-intelligence/aggregation/jobs/list": map[string]interface{}{
						"description": "List all business intelligence aggregation jobs",
					},
					"POST /v2/business-intelligence/enhanced-classify": map[string]interface{}{
						"description": "Enhanced classification with business intelligence features",
					},
					"POST /v2/business-intelligence/batch-enhanced": map[string]interface{}{
						"description": "Enhanced batch classification with business intelligence features",
					},
					"GET /v2/business-intelligence/analytics": map[string]interface{}{
						"description": "Business intelligence analytics and insights",
					},
					"GET /v2/business-intelligence/insights": map[string]interface{}{
						"description": "Business intelligence trends and patterns",
					},
				},
			},
			"v1": map[string]interface{}{
				"classification": map[string]interface{}{
					"POST /v1/classify": map[string]interface{}{
						"description": "Legacy single business classification (routes through intelligent routing)",
						"deprecation": "Consider upgrading to v2 endpoints",
					},
					"POST /v1/classify/batch": map[string]interface{}{
						"description": "Legacy batch business classification (routes through intelligent routing)",
						"deprecation": "Consider upgrading to v2 endpoints",
					},
				},
				"health": map[string]interface{}{
					"GET /v1/health": map[string]interface{}{
						"description": "Legacy health check endpoint",
						"deprecation": "Consider upgrading to v2 endpoints",
					},
				},
			},
		},
		"migration_guide": map[string]interface{}{
			"from_v1_to_v2": map[string]interface{}{
				"endpoints": map[string]string{
					"POST /v1/classify":       "POST /v2/classify",
					"POST /v1/classify/batch": "POST /v2/classify/batch",
					"GET /v1/health":          "GET /v2/routing/health",
				},
				"benefits": []string{
					"Enhanced intelligent routing",
					"Improved performance",
					"Better error handling",
					"More comprehensive responses",
					"Advanced business intelligence features",
				},
			},
		},
	}
}
