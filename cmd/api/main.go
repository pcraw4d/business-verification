package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pcraw4d/business-verification/internal/api/handlers"
	"github.com/pcraw4d/business-verification/internal/api/middleware"
	"github.com/pcraw4d/business-verification/internal/auth"
	"github.com/pcraw4d/business-verification/internal/classification"
	"github.com/pcraw4d/business-verification/internal/compliance"
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
	classificationSvc  *classification.ClassificationService
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
	rateLimiter        *middleware.RateLimiter
	authRateLimiter    *middleware.AuthRateLimiter
	ipBlocker          *middleware.IPBlocker
	validator          *middleware.Validator
	server             *http.Server
}

// NewServer creates a new server instance
func NewServer(
	config *config.Config,
	logger *observability.Logger,
	metrics *observability.Metrics,
	classificationSvc *classification.ClassificationService,
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
	rateLimiter *middleware.RateLimiter,
	authRateLimiter *middleware.AuthRateLimiter,
	ipBlocker *middleware.IPBlocker,
	validator *middleware.Validator,
) *Server {
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
		rateLimiter:        rateLimiter,
		authRateLimiter:    authRateLimiter,
		ipBlocker:          ipBlocker,
		validator:          validator,
	}
}

// setupRoutes configures all API routes using Go 1.22's new ServeMux features
func (s *Server) setupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", s.healthHandler)

	// API versioning with v1 prefix
	mux.HandleFunc("GET /v1/status", s.statusHandler)
	mux.HandleFunc("GET /v1/metrics", s.metricsHandler)

	// API documentation
	mux.HandleFunc("GET /docs", s.docsHandler)
	mux.HandleFunc("GET /docs/", s.docsHandler)

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
	mux.HandleFunc("GET /v1/compliance/export/{export_id}/download", s.complianceHandler.DownloadExportHandler)

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

	// Admin endpoints (protected)
	mux.Handle("POST /v1/admin/users", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.CreateUser)))
	mux.Handle("PUT /v1/admin/users/{id}", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.UpdateUser)))
	mux.Handle("DELETE /v1/admin/users/{id}", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.DeleteUser)))
	mux.Handle("POST /v1/admin/users/{id}/activate", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.ActivateUser)))
	mux.Handle("POST /v1/admin/users/{id}/deactivate", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.DeactivateUser)))
	mux.Handle("GET /v1/admin/users", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.ListUsers)))
	mux.Handle("GET /v1/admin/stats", s.authMiddleware.RequireAuth(http.HandlerFunc(s.adminHandler.GetSystemStats)))

	// Catch-all for undefined routes
	mux.HandleFunc("GET /", s.notFoundHandler)
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
	handler = s.validator.Middleware(handler)
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
	s.logger.WithComponent("api").LogHealthCheck("api", "healthy", map[string]interface{}{
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

	s.logger.WithComponent("api").LogAPIRequest(r.Context(), "GET", r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

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

	s.logger.WithComponent("api").LogAPIRequest(r.Context(), "GET", r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

	// Serve Prometheus metrics
	s.metrics.ServeHTTP(w, r)
}

// docsHandler handles API documentation requests
func (s *Server) docsHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.WithComponent("api").LogAPIRequest(r.Context(), "GET", r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

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

// notFoundHandler handles undefined routes
func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), http.StatusNotFound, time.Since(start))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte(`{"error":"not_found","message":"The requested endpoint does not exist","path":"` + r.URL.Path + `"}`))
}

// notImplementedHandler handles endpoints that are not yet implemented
func (s *Server) notImplementedHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), http.StatusNotImplemented, time.Since(start))

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
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

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
		s.metrics.RecordHTTPRequestStart(r.Method, r.URL.Path)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		// Metrics: request duration and totals
		s.metrics.RecordHTTPRequest(r.Method, r.URL.Path, rw.statusCode, duration)
		s.metrics.RecordHTTPRequestEnd(r.Method, r.URL.Path)

		// Logging
		s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), rw.statusCode, duration)
		// Simple local alerting for slow requests
		if duration > s.config.Observability.SlowRequestThreshold {
			s.logger.WithComponent("api").Warn("slow_request", "method", r.Method, "path", r.URL.Path, "duration_ms", duration.Milliseconds())
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
		ctx := context.WithValue(r.Context(), observability.RequestIDKey, requestID)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// recoveryMiddleware recovers from panics
func (s *Server) recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				s.logger.WithComponent("api").WithError(fmt.Errorf("panic: %v", err)).Error("panic recovered", "method", r.Method, "path", r.URL.Path)

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
	start := time.Now()

	// Parse request body
	var req classification.ClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.WithComponent("api").WithError(err).Error("Failed to parse classification request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_request","message":"Invalid JSON in request body"}`))
		return
	}

	// Perform classification
	response, err := s.classificationSvc.ClassifyBusiness(r.Context(), &req)
	if err != nil {
		s.logger.WithComponent("api").WithError(err).Error("Classification failed")
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(strings.ToLower(err.Error()), "invalid request") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"invalid_request","message":"Invalid classification request"}`))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"classification_failed","message":"Failed to classify business"}`))
		}
		return
	}

	// Log successful classification
	s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// classifyBatchHandler handles batch business classification requests
func (s *Server) classifyBatchHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	// Parse request body
	var req classification.BatchClassificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.WithComponent("api").WithError(err).Error("Failed to parse batch classification request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_request","message":"Invalid JSON in request body"}`))
		return
	}

	// Basic validation: at least one business
	if len(req.Businesses) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_request","message":"At least one business is required"}`))
		return
	}

	// Perform batch classification
	response, err := s.classificationSvc.ClassifyBusinessesBatch(r.Context(), &req)
	if err != nil {
		s.logger.WithComponent("api").WithError(err).Error("Batch classification failed")
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(strings.ToLower(err.Error()), "batch size") {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"invalid_request","message":"Batch size exceeds limit"}`))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"batch_classification_failed","message":"Failed to classify businesses"}`))
		}
		return
	}

	// Log successful batch classification
	s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// classificationHistoryHandler returns a paginated classification history for a business
func (s *Server) classificationHistoryHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	businessID := r.PathValue("business_id")
	if businessID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_request","message":"business_id is required"}`))
		return
	}

	// Parse pagination
	limit := 50
	offset := 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}

	history, err := s.classificationSvc.GetClassificationHistory(r.Context(), businessID, limit, offset)
	if err != nil {
		s.logger.WithComponent("api").WithError(err).Error("Failed to fetch classification history")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal_error","message":"Failed to retrieve classification history"}`))
		return
	}

	s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"business_id": businessID,
		"count":       len(history),
		"results":     history,
		"limit":       limit,
		"offset":      offset,
	})
}

// classificationConfidenceHandler accepts a classification response payload and summarizes confidence metrics
func (s *Server) classificationConfidenceHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	var resp classification.ClassificationResponse
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_request","message":"Invalid JSON in request body"}`))
		return
	}

	// If client only sends classifications array, primary may be nil; compute a quick summary
	total := len(resp.Classifications)
	if total == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"invalid_request","message":"classifications array required"}`))
		return
	}

	// Aggregate stats
	sum := 0.0
	maxScore := -1.0
	var top classification.IndustryClassification
	methodCounts := make(map[string]int)
	codeCounts := make(map[string]int)
	for _, cl := range resp.Classifications {
		sum += cl.ConfidenceScore
		methodCounts[cl.ClassificationMethod]++
		codeCounts[cl.IndustryCode]++
		if cl.ConfidenceScore > maxScore {
			maxScore = cl.ConfidenceScore
			top = cl
		}
	}
	avg := sum / float64(total)

	s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":                total,
		"average_confidence":   avg,
		"top_industry_code":    top.IndustryCode,
		"top_industry_name":    top.IndustryName,
		"top_confidence_score": top.ConfidenceScore,
		"method_counts":        methodCounts,
		"code_agreement":       codeCounts,
	})
}

// dataSourcesHealthHandler returns the current health status of all configured data sources
func (s *Server) dataSourcesHealthHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	// reach into classification svc enricher for health; if nil, return empty
	var statuses []map[string]interface{}
	if s.classificationSvc != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 1500*time.Millisecond)
		defer cancel()
		hs := s.classificationSvc.DataSourcesHealth(ctx)
		statuses = make([]map[string]interface{}, 0, len(hs))
		for _, h := range hs {
			statuses = append(statuses, map[string]interface{}{
				"source_name": h.SourceName,
				"healthy":     h.Healthy,
				"checked_at":  h.CheckedAt,
				"latency_ms":  h.Latency.Milliseconds(),
				"error":       h.Error,
			})
		}
	}
	s.logger.WithComponent("api").LogAPIRequest(r.Context(), r.Method, r.URL.Path, r.UserAgent(), http.StatusOK, time.Since(start))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data_sources": statuses,
		"count":        len(statuses),
		"timestamp":    time.Now().UTC().Format(time.RFC3339),
	})
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

	s.logger.WithComponent("api").LogStartup("1.0.0", "dev", time.Now().Format(time.RFC3339))

	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.WithComponent("api").LogShutdown("graceful_shutdown")

	return s.server.Shutdown(ctx)
}

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger := observability.NewLogger(&cfg.Observability)

	// Initialize metrics
	metrics, err := observability.NewMetrics(&cfg.Observability)
	if err != nil {
		log.Fatalf("Failed to initialize metrics: %v", err)
	}

	// Load industry data for classification
	industryData, err := classification.LoadIndustryCodes("Codes")
	if err != nil {
		log.Fatalf("Failed to load industry codes: %v", err)
	}

	// Initialize database connection
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()
	db, err := database.NewDatabaseWithConnection(dbCtx, &cfg.Database)
	if err != nil {
		logger.WithComponent("api").WithError(err).Error("Failed to connect to database")
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize classification service
	classificationSvc := classification.NewClassificationServiceWithData(
		&cfg.ExternalServices,
		db,
		logger,
		metrics,
		industryData,
	)

	// Initialize authentication service
	authService := auth.NewAuthService(&cfg.Auth, db, logger, metrics)

	// Initialize authentication handlers and middleware
	authHandler := handlers.NewAuthHandler(authService, logger, metrics, cfg)
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)

	// Initialize admin service and handlers
	rbacService := auth.NewRBACService(authService)
	roleService := auth.NewRoleService(db, logger, rbacService)
	apiKeyService := auth.NewAPIKeyService(db, logger)
	adminService := auth.NewAdminService(db, logger, authService, roleService, apiKeyService)
	adminHandler := handlers.NewAdminHandler(adminService, logger)

	// Initialize compliance engines
	mappingSystem := compliance.NewFrameworkMappingSystem(logger)
	trackingSystem := compliance.NewTrackingSystem(logger)
	ruleEngine := compliance.NewRuleEngine(logger)
	checkEngine := compliance.NewCheckEngine(logger, ruleEngine, trackingSystem, mappingSystem)
	statusSystem := compliance.NewComplianceStatusSystem(logger)
	gapAnalyzer := compliance.NewGapAnalyzer(logger, trackingSystem, mappingSystem)
	scoringEngine := compliance.NewScoringEngine(logger, compliance.ScoreWeights{})
	recommendations := compliance.NewRecommendationEngine(logger, scoringEngine, gapAnalyzer)
	complianceReportService := compliance.NewReportGenerationService(logger, checkEngine, trackingSystem, gapAnalyzer, recommendations)

	// Initialize compliance alert system
	complianceAlertSystem := compliance.NewAlertSystem(logger, statusSystem, checkEngine)

	// Initialize compliance export system
	complianceExportSystem := compliance.NewExportSystem(logger, statusSystem, complianceReportService, complianceAlertSystem)

	// Initialize compliance handler
	complianceHandler := handlers.NewComplianceHandler(logger, checkEngine, statusSystem, complianceReportService, complianceAlertSystem, complianceExportSystem)

	// Initialize rate limiting middleware
	rateLimitConfig := &middleware.RateLimitConfig{
		RequestsPerMinute: cfg.Server.RateLimit.RequestsPer,
		BurstSize:         cfg.Server.RateLimit.BurstSize,
		Enabled:           cfg.Server.RateLimit.Enabled,
	}
	rateLimiter := middleware.NewRateLimiter(rateLimitConfig, logger)

	// Initialize auth-specific rate limiting middleware
	authRateLimitConfig := &middleware.AuthRateLimitConfig{
		Enabled:                  cfg.Server.AuthRateLimit.Enabled,
		LoginAttemptsPer:         cfg.Server.AuthRateLimit.LoginAttemptsPer,
		RegisterAttemptsPer:      cfg.Server.AuthRateLimit.RegisterAttemptsPer,
		PasswordResetAttemptsPer: cfg.Server.AuthRateLimit.PasswordResetAttemptsPer,
		WindowSize:               cfg.Server.AuthRateLimit.WindowSize,
		LockoutDuration:          cfg.Server.AuthRateLimit.LockoutDuration,
	}
	authRateLimiter := middleware.NewAuthRateLimiter(authRateLimitConfig, logger)

	// Initialize IP blocker middleware
	ipBlocker := middleware.NewIPBlocker(
		cfg.Server.IPBlock.Enabled,
		cfg.Server.IPBlock.Threshold,
		cfg.Server.IPBlock.Window,
		cfg.Server.IPBlock.BlockDuration,
		cfg.Server.IPBlock.Whitelist,
		cfg.Server.IPBlock.Blacklist,
		logger,
	)

	// Initialize risk assessment components
	categoryRegistry := risk.CreateDefaultRiskCategories()
	thresholdManager := risk.CreateDefaultThresholds()
	industryModelRegistry := risk.CreateDefaultIndustryModels()

	// Initialize risk calculation components
	calculator := risk.NewRiskFactorCalculator(categoryRegistry)
	scoringAlgorithm := risk.NewWeightedScoringAlgorithm()
	predictionAlgorithm := risk.NewRiskPredictionAlgorithm()

	// Initialize risk history service
	riskHistoryService := risk.NewRiskHistoryService(logger, db)

	// Initialize alert service
	alertService := risk.NewAlertService(logger, thresholdManager)

	// Initialize report service
	reportService := risk.NewReportService(logger, riskHistoryService, alertService)

	// Initialize export service
	exportService := risk.NewExportService(logger, riskHistoryService, alertService, reportService)

	// Initialize financial provider manager
	financialProviderManager := risk.NewFinancialProviderManager(logger)

	// Register mock providers for testing
	mockProvider := risk.NewMockFinancialProvider("mock_provider")
	backupProvider := risk.NewMockFinancialProvider("backup_provider")
	financialProviderManager.RegisterProvider("mock_provider", mockProvider)
	financialProviderManager.RegisterProvider("backup_provider", backupProvider)

	// Initialize regulatory provider manager
	regulatoryProviderManager := risk.NewRegulatoryProviderManager(logger)

	// Initialize media provider manager
	mediaProviderManager := risk.NewMediaProviderManager(logger)

	// Initialize market data provider manager
	marketDataProviderManager := risk.NewMarketDataProviderManager(logger)

	// Initialize data validation manager
	dataValidationManager := risk.NewDataValidationManager(logger)

	// Initialize threshold monitoring manager
	thresholdMonitoringManager := risk.NewThresholdMonitoringManager(logger)

	// Initialize automated alert service
	automatedAlertService := risk.NewAutomatedAlertService(logger)

	// Initialize trend analysis service
	trendAnalysisService := risk.NewTrendAnalysisService(logger)

	// Initialize reporting system
	reportingSystem := risk.NewReportingSystem(logger, reportService, trendAnalysisService, riskHistoryService, alertService)

	// Initialize risk service
	riskService := risk.NewRiskService(
		logger,
		calculator,
		scoringAlgorithm,
		predictionAlgorithm,
		thresholdManager,
		categoryRegistry,
		industryModelRegistry,
		riskHistoryService,
		alertService,
		reportService,
		exportService,
		financialProviderManager,
		regulatoryProviderManager,
		mediaProviderManager,
		marketDataProviderManager,
		dataValidationManager,
		thresholdMonitoringManager,
		automatedAlertService,
		trendAnalysisService,
		reportingSystem,
	)

	// Initialize risk handler
	riskHandler := handlers.NewRiskHandler(logger, riskService, riskHistoryService)

	// Initialize dashboard handler
	dashboardHandler := handlers.NewDashboardHandler(logger, riskService)

	// Initialize validation middleware
	validationConfig := &middleware.ValidationConfig{
		MaxBodySize:   cfg.Server.Validation.MaxBodySize,
		RequiredPaths: cfg.Server.Validation.RequiredPaths,
		Enabled:       cfg.Server.Validation.Enabled,
	}
	validator := middleware.NewValidator(validationConfig, logger)

	// Create server
	server := NewServer(cfg, logger, metrics, classificationSvc, riskService, riskHistoryService, riskHandler, dashboardHandler, authService, authHandler, authMiddleware, adminService, adminHandler, complianceHandler, rateLimiter, authRateLimiter, ipBlocker, validator)

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			logger.WithComponent("api").WithError(err).LogShutdown("server_start_failed")
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
		logger.WithComponent("api").WithError(err).LogShutdown("server_shutdown_failed")
		log.Fatalf("Server shutdown failed: %v", err)
	}
	// Close database connection
	if db != nil {
		if err := db.Close(); err != nil {
			logger.WithComponent("api").WithError(err).LogShutdown("database_close_failed")
		}
	}

	logger.WithComponent("api").LogShutdown("server_shutdown_complete")
}
