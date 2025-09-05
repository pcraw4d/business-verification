package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/api/middleware"
	"github.com/pcraw4d/business-verification/internal/auth"
	"github.com/pcraw4d/business-verification/internal/classification"
	"go.uber.org/zap"

	// "github.com/pcraw4d/business-verification/internal/compliance" // TODO: Uncomment when compliance services are implemented
	"github.com/pcraw4d/business-verification/internal/config"
	"github.com/pcraw4d/business-verification/internal/database"
	"github.com/pcraw4d/business-verification/internal/observability"
	"github.com/pcraw4d/business-verification/internal/risk"
)

// Server represents the main API server
type Server struct {
	config             *config.Config
	logger             *observability.Logger
	metrics            *observability.Metrics
	classificationSvc  *classification.IndustryDetectionService
	riskService        *risk.RiskService
	riskHistoryService *risk.RiskHistoryService
	riskHandler        *handlers.RiskHandler
	dashboardHandler   *handlers.DashboardHandler
	authService        *auth.AuthService
	authHandler        *handlers.AuthHandler
	authMiddleware     *middleware.AuthMiddleware
	adminService       *auth.AdminService
	adminHandler       *handlers.AdminHandler
	complianceHandler  *handlers.ComplianceHandler
	soc2Handler        *handlers.SOC2Handler
	pciHandler         *handlers.PCIDSSHandler
	gdprHandler        *handlers.GDPRHandler
	auditHandler       *handlers.AuditHandler
	rateLimiter        *middleware.APIRateLimiter
	authRateLimiter    *middleware.AuthRateLimiter
	ipBlocker          *middleware.APIRateLimiter
	validator          *middleware.Validator
	docsHandlerSvc     *handlers.DocsHandler
	server             *http.Server
}

// NewServer creates a new server instance
func NewServer(
	config *config.Config,
	logger *observability.Logger,
	metrics *observability.Metrics,
	classificationSvc *classification.IndustryDetectionService,
	riskService *risk.RiskService,
	riskHistoryService *risk.RiskHistoryService,
	riskHandler *handlers.RiskHandler,
	dashboardHandler *handlers.DashboardHandler,
	authService *auth.AuthService,
	authHandler *handlers.AuthHandler,
	authMiddleware *middleware.AuthMiddleware,
	adminService *auth.AdminService,
	adminHandler *handlers.AdminHandler,
	complianceHandler *handlers.ComplianceHandler,
	soc2Handler *handlers.SOC2Handler,
	pciHandler *handlers.PCIDSSHandler,
	gdprHandler *handlers.GDPRHandler,
	auditHandler *handlers.AuditHandler,
	rateLimiter *middleware.APIRateLimiter,
	authRateLimiter *middleware.AuthRateLimiter,
	ipBlocker *middleware.APIRateLimiter,
	validator *middleware.Validator,
) *Server {
	// Read OpenAPI specification
	openAPISpec, err := os.ReadFile("docs/api/openapi.yaml")
	if err != nil {
		logger.Error("Failed to read OpenAPI specification", map[string]interface{}{
			"error": err.Error(),
		})
		// Use a minimal spec if file is not found
		openAPISpec = []byte(`openapi: 3.1.0
info:
  title: KYB Platform API
  version: 1.0.0`)
	}

	docsHandler := handlers.NewDocsHandler(openAPISpec)

	return &Server{
		config:             config,
		logger:             logger,
		metrics:            metrics,
		classificationSvc:  classificationSvc,
		riskService:        riskService,
		riskHistoryService: riskHistoryService,
		riskHandler:        riskHandler,
		dashboardHandler:   dashboardHandler,
		authService:        authService,
		authHandler:        authHandler,
		authMiddleware:     authMiddleware,
		adminService:       adminService,
		adminHandler:       adminHandler,
		complianceHandler:  complianceHandler,
		soc2Handler:        soc2Handler,
		pciHandler:         pciHandler,
		gdprHandler:        gdprHandler,
		auditHandler:       auditHandler,
		rateLimiter:        rateLimiter,
		authRateLimiter:    authRateLimiter,
		ipBlocker:          ipBlocker,
		validator:          validator,
		docsHandlerSvc:     docsHandler,
	}
}

// setupRoutes configures all API routes using Go 1.22's new ServeMux features
func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Serve static web files (must come before other routes)
	mux.HandleFunc("GET /", s.webHandler)
	mux.HandleFunc("GET /index.html", s.webHandler)

	// Health check endpoint
	mux.HandleFunc("GET /health", s.healthHandler)

	// API versioning with v1 prefix
	mux.HandleFunc("GET /v1/status", s.statusHandler)
	mux.HandleFunc("GET /v1/metrics", s.metricsHandler)

	// API documentation
	mux.HandleFunc("GET /docs", s.docsHandlerSvc.ServeDocs)
	mux.HandleFunc("GET /docs/", s.docsHandlerSvc.ServeDocs)
	mux.HandleFunc("GET /docs/openapi.yaml", s.docsHandlerSvc.ServeDocs)

	// Compliance endpoints (protected in future; currently public)
	mux.HandleFunc("POST /v1/compliance/check", s.complianceHandler.CheckComplianceHandler)
	mux.HandleFunc("POST /v1/compliance/report", s.complianceHandler.GenerateComplianceReportHandler)

	// Compliance status tracking endpoints
	mux.HandleFunc("GET /v1/compliance/status/{business_id}", s.complianceHandler.GetComplianceStatusHandler)
	mux.HandleFunc("GET /v1/compliance/status/{business_id}/history", s.complianceHandler.GetStatusHistoryHandler)
	mux.HandleFunc("GET /v1/compliance/status/{business_id}/alerts", s.complianceHandler.GetStatusAlertsHandler)
	mux.HandleFunc("POST /v1/compliance/status/{business_id}/alerts/{alert_id}/acknowledge", s.complianceHandler.AcknowledgeAlertHandler)
	mux.HandleFunc("POST /v1/compliance/status/{business_id}/alerts/{alert_id}/resolve", s.complianceHandler.ResolveAlertHandler)
	mux.HandleFunc("POST /v1/compliance/status/{business_id}/report", s.complianceHandler.GenerateStatusReportHandler)
	mux.HandleFunc("POST /v1/compliance/status/{business_id}/initialize", s.complianceHandler.InitializeBusinessStatusHandler)

	// Compliance alert system endpoints
	mux.HandleFunc("POST /v1/compliance/alerts/rules", s.complianceHandler.RegisterAlertRuleHandler)
	mux.HandleFunc("PUT /v1/compliance/alerts/rules/{rule_id}", s.complianceHandler.UpdateAlertRuleHandler)
	mux.HandleFunc("DELETE /v1/compliance/alerts/rules/{rule_id}", s.complianceHandler.DeleteAlertRuleHandler)
	mux.HandleFunc("GET /v1/compliance/alerts/rules/{rule_id}", s.complianceHandler.GetAlertRuleHandler)
	mux.HandleFunc("GET /v1/compliance/alerts/rules", s.complianceHandler.ListAlertRulesHandler)
	mux.HandleFunc("POST /v1/compliance/alerts/evaluate", s.complianceHandler.EvaluateAlertsHandler)
	mux.HandleFunc("GET /v1/compliance/alerts/analytics/{business_id}", s.complianceHandler.GetAlertAnalyticsHandler)
	mux.HandleFunc("POST /v1/compliance/alerts/escalations", s.complianceHandler.RegisterEscalationPolicyHandler)
	mux.HandleFunc("POST /v1/compliance/alerts/notifications", s.complianceHandler.RegisterNotificationChannelHandler)

	// Compliance export endpoints
	mux.HandleFunc("POST /v1/compliance/export", s.complianceHandler.ExportComplianceDataHandler)
	mux.HandleFunc("POST /v1/compliance/export/job", s.complianceHandler.CreateExportJobHandler)
	mux.HandleFunc("GET /v1/compliance/export/job/{job_id}", s.complianceHandler.GetExportJobHandler)
	mux.HandleFunc("GET /v1/compliance/export/jobs", s.complianceHandler.ListExportJobsHandler)
	mux.HandleFunc("GET /v1/compliance/export/download/{export_id}", s.complianceHandler.DownloadExportHandler)

	// Compliance data retention endpoints
	mux.HandleFunc("POST /v1/compliance/retention/policies", s.complianceHandler.RegisterRetentionPolicyHandler)
	mux.HandleFunc("PUT /v1/compliance/retention/policies/{policy_id}", s.complianceHandler.UpdateRetentionPolicyHandler)
	mux.HandleFunc("DELETE /v1/compliance/retention/policies/{policy_id}", s.complianceHandler.DeleteRetentionPolicyHandler)
	mux.HandleFunc("GET /v1/compliance/retention/policies/{policy_id}", s.complianceHandler.GetRetentionPolicyHandler)
	mux.HandleFunc("GET /v1/compliance/retention/policies", s.complianceHandler.ListRetentionPoliciesHandler)
	mux.HandleFunc("POST /v1/compliance/retention/jobs", s.complianceHandler.ExecuteRetentionJobHandler)
	mux.HandleFunc("GET /v1/compliance/retention/analytics", s.complianceHandler.GetRetentionAnalyticsHandler)

	// SOC 2 compliance endpoints
	mux.HandleFunc("POST /v1/soc2/initialize", s.soc2Handler.InitializeSOC2TrackingHandler)
	mux.HandleFunc("GET /v1/soc2/status/{business_id}", s.soc2Handler.GetSOC2StatusHandler)
	mux.HandleFunc("PUT /v1/soc2/requirements/{business_id}/{requirement_id}", s.soc2Handler.UpdateSOC2RequirementHandler)
	mux.HandleFunc("PUT /v1/soc2/criteria/{business_id}/{criteria_id}", s.soc2Handler.UpdateSOC2CriteriaHandler)
	mux.HandleFunc("POST /v1/soc2/assess/{business_id}", s.soc2Handler.AssessSOC2ComplianceHandler)
	mux.HandleFunc("GET /v1/soc2/report/{business_id}", s.soc2Handler.GetSOC2ReportHandler)
	mux.HandleFunc("GET /v1/soc2/criteria", s.soc2Handler.GetSOC2CriteriaHandler)
	mux.HandleFunc("GET /v1/soc2/requirements", s.soc2Handler.GetSOC2RequirementsHandler)

	// PCI DSS compliance endpoints
	mux.HandleFunc("POST /v1/pci-dss/initialize", s.pciHandler.InitializePCIDSSTrackingHandler)
	mux.HandleFunc("GET /v1/pci-dss/status/{business_id}", s.pciHandler.GetPCIDSSStatusHandler)
	mux.HandleFunc("PUT /v1/pci-dss/requirements/{business_id}/{requirement_id}", s.pciHandler.UpdatePCIDSSRequirementHandler)
	mux.HandleFunc("PUT /v1/pci-dss/categories/{business_id}/{category_id}", s.pciHandler.UpdatePCIDSSCategoryHandler)
	mux.HandleFunc("POST /v1/pci-dss/assess/{business_id}", s.pciHandler.AssessPCIDSSComplianceHandler)
	mux.HandleFunc("GET /v1/pci-dss/report/{business_id}", s.pciHandler.GetPCIDSSReportHandler)
	mux.HandleFunc("GET /v1/pci-dss/categories", s.pciHandler.GetPCIDSSCategoriesHandler)
	mux.HandleFunc("GET /v1/pci-dss/requirements", s.pciHandler.GetPCIDSSRequirementsHandler)

	// GDPR compliance endpoints
	mux.HandleFunc("POST /v1/gdpr/initialize", s.gdprHandler.InitializeGDPRTrackingHandler)
	mux.HandleFunc("GET /v1/gdpr/status/{business_id}", s.gdprHandler.GetGDPRStatusHandler)
	mux.HandleFunc("PUT /v1/gdpr/requirements/{business_id}/{requirement_id}", s.gdprHandler.UpdateGDPRRequirementHandler)
	mux.HandleFunc("PUT /v1/gdpr/principles/{business_id}/{principle_id}", s.gdprHandler.UpdateGDPRPrincipleHandler)
	mux.HandleFunc("PUT /v1/gdpr/rights/{business_id}/{right_id}", s.gdprHandler.UpdateGDPRDataSubjectRightHandler)
	mux.HandleFunc("POST /v1/gdpr/assess/{business_id}", s.gdprHandler.AssessGDPRComplianceHandler)
	mux.HandleFunc("GET /v1/gdpr/report/{business_id}", s.gdprHandler.GetGDPRReportHandler)
	mux.HandleFunc("GET /v1/gdpr/principles", s.gdprHandler.GetGDPRPrinciplesHandler)
	mux.HandleFunc("GET /v1/gdpr/rights", s.gdprHandler.GetGDPRDataSubjectRightsHandler)
	mux.HandleFunc("GET /v1/gdpr/requirements", s.gdprHandler.GetGDPRRequirementsHandler)

	// Compliance audit endpoints
	mux.HandleFunc("POST /v1/audit/events", s.auditHandler.RecordAuditEvent)
	mux.HandleFunc("GET /v1/audit/events", s.auditHandler.GetAuditEvents)
	mux.HandleFunc("GET /v1/audit/trail/{business_id}", s.auditHandler.GetAuditTrail)
	mux.HandleFunc("POST /v1/audit/reports", s.auditHandler.GenerateAuditReport)
	mux.HandleFunc("GET /v1/audit/metrics/{business_id}", s.auditHandler.GetAuditMetrics)
	mux.HandleFunc("PUT /v1/audit/metrics/{business_id}", s.auditHandler.UpdateAuditMetrics)

	// Authentication endpoints (public)
	mux.HandleFunc("POST /v1/auth/register", s.authHandler.RegisterHandler)
	mux.HandleFunc("POST /v1/auth/login", s.authHandler.LoginHandler)
	mux.HandleFunc("POST /v1/auth/refresh", s.authHandler.RefreshTokenHandler)
	mux.HandleFunc("GET /v1/auth/verify-email", s.authHandler.VerifyEmailHandler)
	mux.HandleFunc("POST /v1/auth/request-password-reset", s.authHandler.RequestPasswordResetHandler)
	mux.HandleFunc("POST /v1/auth/reset-password", s.authHandler.ResetPasswordHandler)

	// Protected authentication endpoints
	mux.Handle("POST /v1/auth/logout", s.authMiddleware.RequireAuth(http.HandlerFunc(s.authHandler.LogoutHandler)))
	mux.Handle("POST /v1/auth/change-password", s.authMiddleware.RequireAuth(http.HandlerFunc(s.authHandler.ChangePasswordHandler)))
	mux.Handle("GET /v1/auth/profile", s.authMiddleware.RequireAuth(http.HandlerFunc(s.authHandler.ProfileHandler)))

	// Classification endpoints (public for now, can be protected later)
	mux.HandleFunc("POST /v1/classify", s.classifyHandler)
	mux.HandleFunc("POST /v1/classify/batch", s.classifyBatchHandler)
	mux.HandleFunc("POST /v1/classify/confidence-report", s.classificationConfidenceHandler)
	mux.HandleFunc("GET /v1/classify/history/{business_id}", s.classificationHistoryHandler)
	mux.HandleFunc("GET /v1/datasources/health", s.dataSourcesHealthHandler)

	// Risk assessment endpoints (public for now, can be protected later)
	mux.HandleFunc("POST /v1/risk/assess", s.riskHandler.AssessRiskHandler)
	mux.HandleFunc("GET /v1/risk/categories", s.riskHandler.GetRiskCategoriesHandler)
	mux.HandleFunc("GET /v1/risk/factors", s.riskHandler.GetRiskFactorsHandler)
	mux.HandleFunc("GET /v1/risk/thresholds", s.riskHandler.GetRiskThresholdsHandler)

	// Risk history endpoints
	mux.HandleFunc("GET /v1/risk/history/{business_id}", s.riskHandler.GetRiskHistoryHandler)
	mux.HandleFunc("GET /v1/risk/trends/{business_id}", s.riskHandler.GetRiskTrendsHandler)
	mux.HandleFunc("GET /v1/risk/history/{business_id}/range", s.riskHandler.GetRiskHistoryByDateRangeHandler)

	// Risk alert endpoints
	mux.HandleFunc("GET /v1/risk/alerts/{business_id}", s.riskHandler.GetRiskAlertsHandler)
	mux.HandleFunc("GET /v1/risk/alert-rules", s.riskHandler.GetRiskAlertRulesHandler)
	mux.HandleFunc("POST /v1/risk/alerts/{alert_id}/acknowledge", s.riskHandler.AcknowledgeRiskAlertHandler)

	// Dashboard endpoints
	mux.HandleFunc("GET /v1/dashboard/overview", s.dashboardHandler.GetDashboardOverviewHandler)
	mux.HandleFunc("GET /v1/dashboard/business/{business_id}", s.dashboardHandler.GetDashboardBusinessHandler)
	mux.HandleFunc("GET /v1/dashboard/analytics", s.dashboardHandler.GetDashboardAnalyticsHandler)
	mux.HandleFunc("GET /v1/dashboard/alerts", s.dashboardHandler.GetDashboardAlertsHandler)
	mux.HandleFunc("GET /v1/dashboard/monitoring", s.dashboardHandler.GetDashboardMonitoringHandler)
	mux.HandleFunc("GET /v1/dashboard/thresholds", s.dashboardHandler.GetDashboardThresholdsHandler)

	// Compliance dashboard endpoints
	mux.HandleFunc("GET /v1/dashboard/compliance/overview", s.dashboardHandler.GetDashboardComplianceOverviewHandler)
	mux.HandleFunc("GET /v1/dashboard/compliance/business/{business_id}", s.dashboardHandler.GetDashboardComplianceBusinessHandler)
	mux.HandleFunc("GET /v1/dashboard/compliance/analytics", s.dashboardHandler.GetDashboardComplianceAnalyticsHandler)

	// Admin endpoints (protected)
	mux.Handle("POST /v1/admin/users", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.CreateUser)))
	mux.Handle("PUT /v1/admin/users/{id}", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.UpdateUser)))
	mux.Handle("DELETE /v1/admin/users/{id}", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.DeleteUser)))
	mux.Handle("POST /v1/admin/users/{id}/activate", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.ActivateUser)))
	mux.Handle("POST /v1/admin/users/{id}/deactivate", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.DeactivateUser)))
	mux.Handle("GET /v1/admin/users", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.ListUsers)))
	mux.Handle("GET /v1/admin/stats", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.GetSystemStats)))

	// Catch-all for undefined routes (excluding GET / which is handled by webHandler)
	mux.HandleFunc("POST /", s.notFoundHandler)
	mux.HandleFunc("PUT /", s.notFoundHandler)
	mux.HandleFunc("DELETE /", s.notFoundHandler)

	return mux
}

// setupMiddleware configures the middleware stack
func (s *Server) setupMiddleware(handler http.Handler) http.Handler {
	// Apply middleware in order (last middleware is applied first)
	handler = s.securityHeadersMiddleware(handler)
	handler = s.corsMiddleware(handler)
	handler = s.validator.ValidationMiddleware(handler)
	handler = s.authRateLimiter.Middleware(handler) // Auth-specific rate limiting
	handler = s.rateLimiter.Middleware(handler)     // General rate limiting
	handler = s.ipBlocker.Middleware(handler)       // IP-based blocking
	handler = s.requestLoggingMiddleware(handler)
	handler = s.requestIDMiddleware(handler)
	handler = s.recoveryMiddleware(handler)

	return handler
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	s.logger.WithComponent("api").Info("health_check", map[string]interface{}{
		"endpoint":   "/health",
		"method":     r.Method,
		"user_agent": r.UserAgent(),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"}`))
}

// statusHandler handles API status requests
func (s *Server) statusHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.WithComponent("api").LogAPIRequest("GET", r.URL.Path, http.StatusOK, time.Since(start), map[string]interface{}{
		"user_agent": r.UserAgent(),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{
		"status":"operational",
		"version":"1.0.0",
		"timestamp":"` + time.Now().UTC().Format(time.RFC3339) + `"
	}`))
}

// metricsHandler handles metrics requests
func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.WithComponent("api").LogAPIRequest("GET", r.URL.Path, http.StatusOK, time.Since(start), map[string]interface{}{
		"user_agent": r.UserAgent(),
	})

	// Serve Prometheus metrics
	// s.metrics.ServeHTTP(w, r) // TODO: Implement metrics HTTP handler
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("Metrics endpoint not implemented"))
}

// docsHandler handles API documentation requests
func (s *Server) docsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.WithComponent("api").LogAPIRequest("GET", r.URL.Path, http.StatusOK, time.Since(start), map[string]interface{}{
		"user_agent": r.UserAgent(),
	})

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <title>KYB Tool API Documentation</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .endpoint { margin: 20px 0; padding: 10px; border-left: 4px solid #007cba; }
        .method { font-weight: bold; color: #007cba; }
    </style>
</head>
<body>
    <h1>KYB Tool API Documentation</h1>
    <p>Welcome to the KYB Tool API. This documentation will be enhanced with OpenAPI/Swagger specification.</p>
    
    <h2>Available Endpoints</h2>
    
    <div class="endpoint">
        <span class="method">GET</span> /health - Health check endpoint
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> /v1/status - API status information
    </div>
    
    <div class="endpoint">
        <span class="method">GET</span> /v1/metrics - Prometheus metrics endpoint
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> /v1/auth/register - User registration (coming soon)
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> /v1/auth/login - User login (coming soon)
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> /v1/classify - Business classification (coming soon)
    </div>
    
    <div class="endpoint">
        <span class="method">POST</span> /v1/classify/batch - Batch business classification (coming soon)
    </div>
</body>
</html>`))
}

// webHandler serves the main web interface
func (s *Server) webHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.WithComponent("api").LogAPIRequest(r.Method, r.URL.Path, http.StatusOK, time.Since(start), map[string]interface{}{
		"user_agent": r.UserAgent(),
	})

	// Read the web/index.html file
	content, err := os.ReadFile("web/index.html")
	if err != nil {
		s.logger.WithComponent("api").Error("Failed to read web/index.html", map[string]interface{}{
			"error": err.Error(),
		})
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// notFoundHandler handles undefined routes
func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.WithComponent("api").LogAPIRequest(r.Method, r.URL.Path, http.StatusNotFound, time.Since(start), map[string]interface{}{
		"user_agent": r.UserAgent(),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error":"not_found","message":"The requested endpoint does not exist","path":"` + r.URL.Path + `"}`))
}

// notImplementedHandler handles endpoints that are not yet implemented
func (s *Server) notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.WithComponent("api").LogAPIRequest(r.Method, r.URL.Path, http.StatusNotImplemented, time.Since(start), map[string]interface{}{
		"user_agent": r.UserAgent(),
	})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not_implemented","message":"This endpoint is not yet implemented","path":"` + r.URL.Path + `"}`))
}

// securityHeadersMiddleware adds security headers to responses
func (s *Server) securityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com; font-src 'self' https://cdn.jsdelivr.net https://cdnjs.cloudflare.com data:; script-src 'self' 'unsafe-inline'; img-src 'self' data: https:;")

		// Remove server information
		w.Header().Set("Server", "KYB-Tool")

		next.ServeHTTP(w, r)
	})
}

// corsMiddleware handles Cross-Origin Resource Sharing
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusOK)
			return
		}

		// Set CORS headers for actual requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		next.ServeHTTP(w, r)
	})
}

// requestLoggingMiddleware logs all incoming requests
func (s *Server) requestLoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a custom response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// Metrics: in-flight
		// s.metrics.RecordHTTPRequestStart(r.Method, r.URL.Path) // TODO: Implement metrics

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Metrics: request duration and totals
		// s.metrics.RecordHTTPRequest(r.Method, r.URL.Path, rw.statusCode, duration) // TODO: Implement metrics
		// s.metrics.RecordHTTPRequestEnd(r.Method, r.URL.Path) // TODO: Implement metrics

		// Logging
		s.logger.WithComponent("api").LogAPIRequest(r.Method, r.URL.Path, rw.statusCode, duration, map[string]interface{}{
			"user_agent": r.UserAgent(),
		})
		// Simple local alerting for slow requests
		if duration > 300*time.Millisecond { // Default threshold
			s.logger.WithComponent("api").Warn("slow_request", map[string]interface{}{
				"method":      r.Method,
				"path":        r.URL.Path,
				"duration_ms": duration.Milliseconds(),
			})
		}
	})
}

// requestIDMiddleware adds request ID to context and headers
func (s *Server) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract request ID from header or generate new one
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = observability.GenerateRequestID()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to context
		ctx := context.WithValue(r.Context(), observability.RequestIDKey{}, requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware recovers from panics
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.WithComponent("api").WithError(fmt.Errorf("panic: %v", err)).Error("panic recovered", map[string]interface{}{
					"method": r.Method,
					"path":   r.URL.Path,
				})

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error":"internal_server_error","message":"An unexpected error occurred"}`))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// classifyHandler handles single business classification requests
func (s *Server) classifyHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement classification handler when ClassificationRequest type is available
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not_implemented","message":"Classification handler not implemented"}`))
}

// classifyBatchHandler handles batch business classification requests
func (s *Server) classifyBatchHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement batch classification handler when BatchClassificationRequest type is available
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not_implemented","message":"Batch classification handler not implemented"}`))
}

// classificationHistoryHandler returns a paginated classification history for a business
func (s *Server) classificationHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement when GetClassificationHistory method is available
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not_implemented","message":"Classification history handler not implemented"}`))
}

// classificationConfidenceHandler accepts a classification response payload and summarizes confidence metrics
func (s *Server) classificationConfidenceHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement when ClassificationResponse type is available
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not_implemented","message":"Classification confidence handler not implemented"}`))
}

// dataSourcesHealthHandler returns the current health status of all configured data sources
func (s *Server) dataSourcesHealthHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement when DataSourcesHealth method is available
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error":"not_implemented","message":"Data sources health handler not implemented"}`))
}

// Start starts the HTTP server
func (s *Server) Start() error {
	// Setup routes
	mux := s.setupRoutes()

	// Setup middleware
	handler := s.setupMiddleware(mux)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port)
	s.server = &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  s.config.Server.ReadTimeout,
		WriteTimeout: s.config.Server.WriteTimeout,
		IdleTimeout:  s.config.Server.IdleTimeout,
	}

	s.logger.WithComponent("api").Info("server_startup", map[string]interface{}{
		"version":   "1.0.0",
		"env":       "dev",
		"timestamp": time.Now().Format(time.RFC3339),
	})

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.WithComponent("api").Info("server_shutdown", map[string]interface{}{
		"reason": "graceful_shutdown",
	})

	return s.server.Shutdown(ctx)
}

func main() {
	// Load environment variables from .env file (optional)
	if err := godotenv.Load(); err != nil {
		// Only log as warning, don't fail - Railway uses environment variables
		log.Printf("Warning: Failed to load .env file: %v (this is normal in Railway)", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to initialize zap logger: %v", err)
	}
	logger := observability.NewLogger(zapLogger)

	// Initialize metrics
	metrics := observability.NewMetrics()

	// Load industry data for classification (optional)
	// TODO: Implement when LoadIndustryCodes function is available
	_ = "industryData" // Placeholder

	// Initialize database connection
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()
	db, err := database.NewDatabaseWithConnection(dbCtx, &cfg.Database)
	if err != nil {
		logger.WithComponent("api").WithError(err).Error("Failed to connect to database", map[string]interface{}{})
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize classification service
	// TODO: Implement when NewClassificationServiceWithData is available
	classificationSvc := &classification.IndustryDetectionService{} // Placeholder

	// Initialize authentication service
	authService := auth.NewAuthService(&cfg.Auth, zapLogger)

	// Initialize authentication handlers and middleware
	authHandler := handlers.NewAuthHandler(authService, zapLogger, cfg)
	authMiddleware := middleware.NewAuthMiddleware(authService, zapLogger)

	// Initialize admin service and handlers
	// TODO: Implement when admin service functions are available
	adminService := &auth.AdminService{}     // Placeholder
	adminHandler := &handlers.AdminHandler{} // Placeholder

	// Initialize compliance engines
	// TODO: Implement when compliance service functions are available
	_ = "mappingSystem"           // Placeholder
	_ = "trackingSystem"          // Placeholder
	_ = "ruleEngine"              // Placeholder
	_ = "checkEngine"             // Placeholder
	_ = "statusSystem"            // Placeholder
	_ = "gapAnalyzer"             // Placeholder
	_ = "scoringEngine"           // Placeholder
	_ = "recommendations"         // Placeholder
	_ = "complianceReportService" // Placeholder

	// Initialize compliance alert system
	_ = "complianceAlertSystem" // Placeholder

	// Initialize compliance export system
	_ = "complianceExportSystem" // Placeholder

	// Initialize SOC 2 tracking service
	_ = "soc2TrackingService" // Placeholder

	// Initialize PCI DSS tracking service
	_ = "pciTrackingService" // Placeholder

	// Initialize GDPR tracking service
	_ = "gdprTrackingService" // Placeholder

	// Initialize compliance handler
	_ = "complianceHandler" // Placeholder

	// Initialize SOC 2 handler
	_ = "soc2Handler" // Placeholder

	// Initialize PCI DSS handler
	_ = "pciHandler" // Placeholder

	// Initialize GDPR handler
	_ = "gdprHandler" // Placeholder

	// Initialize rate limiting middleware
	rateLimitConfig := &middleware.RateLimitConfig{
		Enabled:           cfg.RateLimit.Enabled,
		RequestsPerMinute: cfg.RateLimit.RequestsPer,
		BurstSize:         cfg.RateLimit.BurstSize,
		WindowSize:        time.Duration(cfg.RateLimit.WindowSize) * time.Second,
		Strategy:          "token_bucket",
		Distributed:       false,
		CleanupInterval:   5 * time.Minute,
		MaxKeys:           10000,
	}
	rateLimiter := middleware.NewAPIRateLimiter(rateLimitConfig, zapLogger)

	// Initialize auth-specific rate limiting middleware
	authRateLimitConfig := &middleware.AuthRateLimitConfig{
		Enabled:                  true, // Default to enabled
		LoginAttemptsPer:         5,    // Default values
		RegisterAttemptsPer:      3,
		PasswordResetAttemptsPer: 3,
		WindowSize:               60 * time.Second,
		LockoutDuration:          15 * time.Minute,
		MaxLockouts:              3,
		PermanentLockoutDuration: 24 * time.Hour,
		Distributed:              false,
	}
	authRateLimiter := middleware.NewAuthRateLimiter(authRateLimitConfig, zapLogger)

	// Initialize IP blocker middleware
	// TODO: Implement when NewIPBlocker is available
	ipBlocker := &middleware.APIRateLimiter{} // Placeholder

	// Initialize risk assessment components
	_ = risk.CreateDefaultRiskCategories() // Placeholder
	_ = risk.CreateDefaultThresholds()     // Placeholder
	_ = risk.CreateDefaultIndustryModels() // Placeholder

	// Initialize risk calculation components
	_ = risk.NewRiskFactorCalculator(nil)  // Placeholder
	_ = risk.NewWeightedScoringAlgorithm() // Placeholder
	_ = risk.NewRiskPredictionAlgorithm()  // Placeholder

	// Initialize risk history service
	// TODO: Implement when risk service functions are available
	_ = "riskHistoryService" // Placeholder

	// Initialize alert service
	_ = "alertService" // Placeholder

	// Initialize report service
	_ = "reportService" // Placeholder

	// Initialize export service
	_ = "exportService" // Placeholder

	// Initialize financial provider manager
	// TODO: Implement when financial provider functions are available
	_ = "financialProviderManager" // Placeholder

	// Initialize regulatory provider manager
	// TODO: Implement when risk service functions are available
	_ = "regulatoryProviderManager" // Placeholder

	// Initialize media provider manager
	_ = "mediaProviderManager" // Placeholder

	// Initialize market data provider manager
	_ = "marketDataProviderManager" // Placeholder

	// Initialize data validation manager
	_ = "dataValidationManager" // Placeholder

	// Initialize threshold monitoring manager
	_ = "thresholdMonitoringManager" // Placeholder

	// Initialize automated alert service
	// TODO: Implement when risk service functions are available
	_ = "automatedAlertService" // Placeholder

	// Initialize trend analysis service
	_ = "trendAnalysisService" // Placeholder

	// Initialize reporting system
	_ = "reportingSystem" // Placeholder

	// Initialize risk service
	riskService := risk.NewRiskService(logger)

	// Initialize risk handler
	// TODO: Implement when risk handler is available
	_ = "riskHandler" // Placeholder

	// Initialize dashboard handler
	dashboardHandler := handlers.NewDashboardHandler(logger, riskService)

	// Initialize validation middleware
	validationConfig := &middleware.ValidationConfig{
		MaxRequestSize:      10 * 1024 * 1024, // 10MB default
		AllowedOrigins:      []string{},
		EnableSanitization:  true,
		StrictMode:          false,
		LogValidationErrors: true,
	}
	validator := middleware.NewValidator(validationConfig, zapLogger)

	// Initialize audit system
	// TODO: Implement when compliance audit system is available
	_ = "auditSystem"  // Placeholder
	_ = "auditHandler" // Placeholder

	// Create server
	// TODO: Fix type assertions when all services are properly implemented
	server := NewServer(cfg, logger, metrics, classificationSvc, riskService, nil, nil, dashboardHandler, authService, authHandler, authMiddleware, adminService, adminHandler, nil, nil, nil, nil, nil, rateLimiter, authRateLimiter, ipBlocker, validator)

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("api").WithError(err).Error("server_start_failed", map[string]interface{}{})
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.WithComponent("api").WithError(err).Error("server_shutdown_failed", map[string]interface{}{})
		log.Fatalf("Server shutdown failed: %v", err)
	}
	// Close database connection
	if db != nil {
		if err := db.Close(); err != nil {
			logger.WithComponent("api").WithError(err).Error("database_close_failed", map[string]interface{}{})
		}
	}

	logger.WithComponent("api").Info("server_shutdown_complete", map[string]interface{}{})
}
