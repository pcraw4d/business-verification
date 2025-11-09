package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type FrontendService struct {
	serviceName string
	version     string
	port        string
}

func NewFrontendService() *FrontendService {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "frontend-service"
	}

	version := "5.0.6-BI-DEBUG-ENHANCED"

	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	return &FrontendService{
		serviceName: serviceName,
		version:     version,
		port:        port,
	}
}

func (s *FrontendService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   s.serviceName,
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   s.version,
	})
}

func (s *FrontendService) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Serve the main index page (legacy UI)
	http.ServeFile(w, r, "./static/index.html")
}

func (s *FrontendService) handleAssets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	assets := map[string]interface{}{
		"service":    s.serviceName,
		"version":    s.version,
		"timestamp":  time.Now().Format(time.RFC3339),
		"cdn_status": "active",
		"assets": map[string]interface{}{
			"total_size":        "1.2MB",
			"compressed_size":   "450KB",
			"compression_ratio": "62.5%%",
			"cache_hit_rate":    "99.9%%",
		},
		"static_files": []map[string]interface{}{
			{
				"file":       "app.js",
				"size":       "850KB",
				"compressed": "320KB",
				"cache_ttl":  "1 year",
			},
			{
				"file":       "app.css",
				"size":       "120KB",
				"compressed": "45KB",
				"cache_ttl":  "1 year",
			},
			{
				"file":       "vendor.js",
				"size":       "230KB",
				"compressed": "85KB",
				"cache_ttl":  "1 year",
			},
		},
		"performance": map[string]interface{}{
			"avg_load_time":    "245ms",
			"first_paint":      "180ms",
			"interactive":      "320ms",
			"lighthouse_score": 95,
		},
	}

	json.NewEncoder(w).Encode(assets)
}

// Legacy page handlers
func (s *FrontendService) handleMerchantHub(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/merchant-hub.html")
}

func (s *FrontendService) handleMerchantPortfolio(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/merchant-portfolio.html")
}

func (s *FrontendService) handleBusinessIntelligence(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/business-intelligence.html")
}

func (s *FrontendService) handleComplianceDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/compliance-dashboard.html")
}

func (s *FrontendService) handleRiskDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/risk-dashboard.html")
}

func (s *FrontendService) handleAddMerchant(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/add-merchant.html")
}

func (s *FrontendService) handleMerchantDetails(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/merchant-details.html")
}

func (s *FrontendService) handleMerchantComparison(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/merchant-comparison.html")
}

func (s *FrontendService) handleMerchantBulkOperations(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/merchant-bulk-operations.html")
}

func (s *FrontendService) handleMonitoringDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/monitoring-dashboard.html")
}

func (s *FrontendService) handleApiTest(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/api-test.html")
}

func (s *FrontendService) handleEnhancedRiskIndicators(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/enhanced-risk-indicators.html")
}

func (s *FrontendService) handleComplianceGapAnalysis(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/compliance-gap-analysis.html")
}

func (s *FrontendService) handleComplianceProgressTracking(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/compliance-progress-tracking.html")
}

func (s *FrontendService) handleMarketAnalysisDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/market-analysis-dashboard.html")
}

func (s *FrontendService) handleCompetitiveAnalysisDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/competitive-analysis-dashboard.html")
}

func (s *FrontendService) handleBusinessGrowthAnalytics(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/business-growth-analytics.html")
}

func (s *FrontendService) handleMerchantHubIntegration(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/merchant-hub-integration.html")
}

func (s *FrontendService) handleMerchantDetail(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/merchant-detail.html")
}

func (s *FrontendService) handleAdminDashboard(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/admin-dashboard.html")
}

func (s *FrontendService) handleRegister(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/register.html")
}

func (s *FrontendService) handleSessions(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/sessions.html")
}

func (s *FrontendService) handleAdminModels(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/admin-models.html")
}

func (s *FrontendService) handleAnalyticsInsights(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/analytics-insights.html")
}

func (s *FrontendService) handleAdminQueue(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/admin-queue.html")
}

func (s *FrontendService) handleDashboardHub(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/dashboard-hub.html")
}

func (s *FrontendService) handleRiskAssessmentPortfolio(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/risk-assessment-portfolio.html")
}

func (s *FrontendService) setupRoutes() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./static/js/"))))
	// Components can be in both locations for compatibility
	http.Handle("/components/", http.StripPrefix("/components/", http.FileServer(http.Dir("./static/components/"))))
	http.Handle("/styles/", http.StripPrefix("/styles/", http.FileServer(http.Dir("./static/styles/"))))

	// API endpoints
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/assets", s.handleAssets)

	// Legacy page routes
	http.HandleFunc("/dashboard", s.handleDashboard)
	http.HandleFunc("/dashboard-hub", s.handleDashboardHub)
	http.HandleFunc("/merchant-hub", s.handleMerchantHub)
	http.HandleFunc("/merchant-portfolio", s.handleMerchantPortfolio)
	http.HandleFunc("/business-intelligence", s.handleBusinessIntelligence)
	http.HandleFunc("/compliance-dashboard", s.handleComplianceDashboard)
	http.HandleFunc("/risk-dashboard", s.handleRiskDashboard)
	http.HandleFunc("/add-merchant", s.handleAddMerchant)
	http.HandleFunc("/merchant-details", s.handleMerchantDetails)
	http.HandleFunc("/merchant-comparison", s.handleMerchantComparison)
	http.HandleFunc("/merchant-bulk-operations", s.handleMerchantBulkOperations)
	http.HandleFunc("/monitoring-dashboard", s.handleMonitoringDashboard)
	http.HandleFunc("/api-test", s.handleApiTest)
	http.HandleFunc("/risk-assessment-portfolio", s.handleRiskAssessmentPortfolio)

	// Additional routes for navigation
	http.HandleFunc("/enhanced-risk-indicators", s.handleEnhancedRiskIndicators)
	http.HandleFunc("/enhanced-risk-indicators.html", s.handleEnhancedRiskIndicators)
	http.HandleFunc("/compliance-gap-analysis", s.handleComplianceGapAnalysis)
	http.HandleFunc("/compliance-gap-analysis.html", s.handleComplianceGapAnalysis)
	http.HandleFunc("/compliance-progress-tracking", s.handleComplianceProgressTracking)
	http.HandleFunc("/compliance-progress-tracking.html", s.handleComplianceProgressTracking)
	http.HandleFunc("/market-analysis-dashboard", s.handleMarketAnalysisDashboard)
	http.HandleFunc("/market-analysis-dashboard.html", s.handleMarketAnalysisDashboard)
	http.HandleFunc("/competitive-analysis-dashboard", s.handleCompetitiveAnalysisDashboard)
	http.HandleFunc("/competitive-analysis-dashboard.html", s.handleCompetitiveAnalysisDashboard)
	http.HandleFunc("/business-growth-analytics", s.handleBusinessGrowthAnalytics)
	http.HandleFunc("/business-growth-analytics.html", s.handleBusinessGrowthAnalytics)
	http.HandleFunc("/merchant-hub-integration", s.handleMerchantHubIntegration)
	http.HandleFunc("/merchant-hub-integration.html", s.handleMerchantHubIntegration)
	http.HandleFunc("/merchant-detail", s.handleMerchantDetail)
	http.HandleFunc("/merchant-detail.html", s.handleMerchantDetail)
	http.HandleFunc("/dashboard-hub", s.handleDashboardHub)
	http.HandleFunc("/dashboard-hub.html", s.handleDashboardHub)
	http.HandleFunc("/risk-assessment-portfolio", s.handleRiskAssessmentPortfolio)
	http.HandleFunc("/risk-assessment-portfolio.html", s.handleRiskAssessmentPortfolio)
	http.HandleFunc("/admin", s.handleAdminDashboard)
	http.HandleFunc("/admin.html", s.handleAdminDashboard)
	http.HandleFunc("/register", s.handleRegister)
	http.HandleFunc("/register.html", s.handleRegister)
	http.HandleFunc("/sessions", s.handleSessions)
	http.HandleFunc("/sessions.html", s.handleSessions)
	http.HandleFunc("/admin/models", s.handleAdminModels)
	http.HandleFunc("/admin/models.html", s.handleAdminModels)
	http.HandleFunc("/analytics-insights", s.handleAnalyticsInsights)
	http.HandleFunc("/analytics-insights.html", s.handleAnalyticsInsights)
	http.HandleFunc("/admin/queue", s.handleAdminQueue)
	http.HandleFunc("/admin/queue.html", s.handleAdminQueue)

	// Backward compatibility routes with .html extensions
	http.HandleFunc("/add-merchant.html", s.handleAddMerchant)
	http.HandleFunc("/dashboard.html", s.handleDashboard)
	http.HandleFunc("/merchant-hub.html", s.handleMerchantHub)
	http.HandleFunc("/merchant-portfolio.html", s.handleMerchantPortfolio)
	http.HandleFunc("/business-intelligence.html", s.handleBusinessIntelligence)
	http.HandleFunc("/compliance-dashboard.html", s.handleComplianceDashboard)
	http.HandleFunc("/risk-dashboard.html", s.handleRiskDashboard)
	http.HandleFunc("/merchant-details.html", s.handleMerchantDetails)
	http.HandleFunc("/merchant-comparison.html", s.handleMerchantComparison)
	http.HandleFunc("/merchant-bulk-operations.html", s.handleMerchantBulkOperations)
	http.HandleFunc("/monitoring-dashboard.html", s.handleMonitoringDashboard)
	http.HandleFunc("/api-test.html", s.handleApiTest)
	http.HandleFunc("/index.html", s.handleDashboard)

	// Default route - serve main index page
	http.HandleFunc("/", s.handleDashboard)
}

func main() {
	service := NewFrontendService()
	service.setupRoutes()

	log.Printf("🚀 Starting %s v%s on :%s", service.serviceName, service.version, service.port)
	log.Printf("✅ %s v%s is ready and listening on :%s", service.serviceName, service.version, service.port)
	log.Printf("🔗 Health: http://localhost:%s/health", service.port)
	log.Printf("📁 Assets: http://localhost:%s/assets", service.port)
	log.Printf("🌐 Dashboard: http://localhost:%s/dashboard", service.port)

	if err := http.ListenAndServe(":"+service.port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
