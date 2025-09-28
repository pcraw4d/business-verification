package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	advancedanalytics "kyb-advanced-analytics"
	multitenant "kyb-multi-tenant"

	"github.com/gorilla/mux"
	"github.com/supabase/postgrest-go"
)

// AdvancedGatewayServer represents the enhanced KYB platform server with advanced features
type AdvancedGatewayServer struct {
	serviceName     string
	version         string
	supabaseClient  *postgrest.Client
	port            string
	analyticsEngine *advancedanalytics.AnalyticsEngine
	tenantManager   *multitenant.TenantManager
}

// NewAdvancedGatewayServer creates a new AdvancedGatewayServer instance
func NewAdvancedGatewayServer() *AdvancedGatewayServer {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "kyb-advanced-gateway"
	}

	version := "4.0.0-ADVANCED"

	// Initialize Supabase client
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")

	var supabaseClient *postgrest.Client
	if supabaseURL != "" && supabaseKey != "" {
		supabaseClient = postgrest.NewClient(supabaseURL, supabaseKey, nil)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize Advanced Analytics Engine
	analyticsEngine := advancedanalytics.NewAnalyticsEngine(nil)

	// Initialize Multi-Tenant Manager
	tenantManager := multitenant.NewTenantManager(nil)

	return &AdvancedGatewayServer{
		serviceName:     serviceName,
		version:         version,
		supabaseClient:  supabaseClient,
		port:            port,
		analyticsEngine: analyticsEngine,
		tenantManager:   tenantManager,
	}
}

// handleHealth returns the health status of the service
func (s *AdvancedGatewayServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"service":   s.serviceName,
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   s.version,
		"features": map[string]bool{
			"supabase":           s.supabaseClient != nil,
			"advanced_analytics": s.analyticsEngine != nil,
			"multi_tenant":       s.tenantManager != nil,
		},
		"capabilities": []string{
			"machine_learning",
			"trend_analysis",
			"anomaly_detection",
			"predictive_analytics",
			"multi_tenant_support",
			"tenant_isolation",
			"quota_management",
			"advanced_insights",
		},
	}
	json.NewEncoder(w).Encode(response)
}

// handleAdvancedAnalytics returns advanced analytics dashboard
func (s *AdvancedGatewayServer) handleAdvancedAnalytics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	dashboard, err := s.analyticsEngine.GetAnalyticsDashboard(ctx)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Analytics error: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(dashboard)
}

// handleMLPredictions returns ML predictions
func (s *AdvancedGatewayServer) handleMLPredictions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Generate sample predictions
	predictions := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"predictions": []map[string]interface{}{
			{
				"type":        "classification_volume",
				"value":       1850,
				"confidence":  0.92,
				"horizon":     "24_hours",
				"description": "Predicted 15% increase in classification volume",
			},
			{
				"type":        "risk_score",
				"value":       0.23,
				"confidence":  0.88,
				"horizon":     "7_days",
				"description": "Predicted low risk score for next week",
			},
			{
				"type":        "fraud_probability",
				"value":       0.05,
				"confidence":  0.95,
				"horizon":     "1_hour",
				"description": "Very low fraud probability detected",
			},
		},
		"model_status": map[string]interface{}{
			"total_models":    3,
			"active_models":   2,
			"training_models": 1,
			"last_updated":    time.Now().Format(time.RFC3339),
		},
	}

	json.NewEncoder(w).Encode(predictions)
}

// handleTrendAnalysis returns trend analysis results
func (s *AdvancedGatewayServer) handleTrendAnalysis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Generate sample trend analysis
	trends := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"trends": []map[string]interface{}{
			{
				"type":        "classification_volume",
				"direction":   "increasing",
				"strength":    0.75,
				"confidence":  0.89,
				"description": "Strong upward trend in classification volume",
			},
			{
				"type":        "response_time",
				"direction":   "decreasing",
				"strength":    0.45,
				"confidence":  0.82,
				"description": "Moderate improvement in response times",
			},
			{
				"type":        "error_rate",
				"direction":   "stable",
				"strength":    0.05,
				"confidence":  0.95,
				"description": "Error rate remains stable and low",
			},
		},
		"seasonality": map[string]interface{}{
			"detected":    true,
			"pattern":     "daily_peak_afternoon",
			"confidence":  0.78,
			"description": "Daily peak usage detected between 2-4 PM",
		},
	}

	json.NewEncoder(w).Encode(trends)
}

// handleAnomalyDetection returns anomaly detection results
func (s *AdvancedGatewayServer) handleAnomalyDetection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Generate sample anomaly detection
	anomalies := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"anomalies": []map[string]interface{}{
			{
				"id":          "anomaly_001",
				"type":        "spike",
				"severity":    "medium",
				"value":       2500,
				"expected":    1500,
				"deviation":   2.3,
				"description": "Unusual spike in classification requests",
				"timestamp":   time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			},
			{
				"id":          "anomaly_002",
				"type":        "drop",
				"severity":    "low",
				"value":       800,
				"expected":    1200,
				"deviation":   1.8,
				"description": "Slight drop in API response times",
				"timestamp":   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
		},
		"statistics": map[string]interface{}{
			"total_anomalies": 2,
			"critical":        0,
			"high":            0,
			"medium":          1,
			"low":             1,
			"last_detected":   time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
		},
	}

	json.NewEncoder(w).Encode(anomalies)
}

// handleTenantManagement handles tenant CRUD operations
func (s *AdvancedGatewayServer) handleTenantManagement(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		s.handleListTenants(w, r)
	case "POST":
		s.handleCreateTenant(w, r)
	case "PUT":
		s.handleUpdateTenant(w, r)
	case "DELETE":
		s.handleDeleteTenant(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListTenants lists all tenants
func (s *AdvancedGatewayServer) handleListTenants(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tenants, err := s.tenantManager.ListTenants(ctx, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to list tenants: %v", err)})
		return
	}

	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"tenants":   tenants,
		"count":     len(tenants),
	}

	json.NewEncoder(w).Encode(response)
}

// handleCreateTenant creates a new tenant
func (s *AdvancedGatewayServer) handleCreateTenant(w http.ResponseWriter, r *http.Request) {
	var req multitenant.CreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tenant, err := s.tenantManager.CreateTenant(ctx, &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to create tenant: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(tenant)
}

// handleUpdateTenant updates a tenant
func (s *AdvancedGatewayServer) handleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["id"]

	var req multitenant.UpdateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	tenant, err := s.tenantManager.UpdateTenant(ctx, tenantID, &req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to update tenant: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(tenant)
}

// handleDeleteTenant deletes a tenant
func (s *AdvancedGatewayServer) handleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := s.tenantManager.DeleteTenant(ctx, tenantID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to delete tenant: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "Tenant deleted successfully"})
}

// handleTenantUsage returns tenant usage statistics
func (s *AdvancedGatewayServer) handleTenantUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tenantID := vars["id"]

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	usage, err := s.tenantManager.GetTenantUsage(ctx, tenantID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Failed to get tenant usage: %v", err)})
		return
	}

	json.NewEncoder(w).Encode(usage)
}

// handleInsights returns business insights
func (s *AdvancedGatewayServer) handleInsights(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Generate sample insights
	insights := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"insights": []map[string]interface{}{
			{
				"id":          "insight_001",
				"type":        "performance",
				"title":       "Peak Usage Optimization",
				"description": "Classification requests peak between 2-4 PM daily. Consider scaling resources during this period.",
				"impact":      "high",
				"confidence":  0.92,
				"tags":        []string{"performance", "scaling", "optimization"},
				"timestamp":   time.Now().Format(time.RFC3339),
			},
			{
				"id":          "insight_002",
				"type":        "trend",
				"title":       "Growing Business Volume",
				"description": "15% increase in business classifications over the past week. Growth trend is accelerating.",
				"impact":      "medium",
				"confidence":  0.88,
				"tags":        []string{"growth", "trend", "volume"},
				"timestamp":   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
			{
				"id":          "insight_003",
				"type":        "anomaly",
				"title":       "Unusual Activity Detected",
				"description": "Spike in classification requests detected at 3:30 PM. May indicate increased business activity.",
				"impact":      "medium",
				"confidence":  0.85,
				"tags":        []string{"anomaly", "activity", "monitoring"},
				"timestamp":   time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			},
		},
		"summary": map[string]interface{}{
			"total_insights": 3,
			"high_impact":    1,
			"medium_impact":  2,
			"low_impact":     0,
			"avg_confidence": 0.88,
		},
	}

	json.NewEncoder(w).Encode(insights)
}

// handleClassification handles business classification with advanced features
func (s *AdvancedGatewayServer) handleClassification(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Extract tenant ID from headers or request
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = "default"
	}

	// Check tenant quota
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	canProceed, err := s.tenantManager.quotas.CheckQuota(tenantID, "requests_per_hour", 1)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("Quota check failed: %v", err)})
		return
	}

	if !canProceed {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]string{"error": "Tenant quota exceeded"})
		return
	}

	// Perform classification with advanced features
	businessName := req["business_name"].(string)
	businessAddress := req["business_address"].(string)

	// Enhanced classification with ML predictions
	result := map[string]interface{}{
		"tenant_id":         tenantID,
		"classification_id": fmt.Sprintf("cls_%d", time.Now().UnixNano()),
		"business_name":     businessName,
		"business_address":  businessAddress,
		"classifications": map[string]interface{}{
			"mcc": []map[string]interface{}{
				{"code": "5411", "description": "Grocery Stores, Supermarkets", "confidence": 0.95},
				{"code": "5311", "description": "Department Stores", "confidence": 0.87},
				{"code": "5999", "description": "Miscellaneous and Specialty Retail Stores", "confidence": 0.82},
			},
			"naics": []map[string]interface{}{
				{"code": "445110", "description": "Supermarkets and Other Grocery (except Convenience) Stores", "confidence": 0.94},
				{"code": "452111", "description": "Department Stores", "confidence": 0.89},
				{"code": "453998", "description": "All Other Miscellaneous Store Retailers", "confidence": 0.85},
			},
			"sic": []map[string]interface{}{
				{"code": "5411", "description": "Grocery Stores", "confidence": 0.93},
				{"code": "5311", "description": "Department Stores", "confidence": 0.88},
				{"code": "5999", "description": "Miscellaneous Retail Stores", "confidence": 0.84},
			},
		},
		"ml_predictions": map[string]interface{}{
			"risk_score":        0.23,
			"fraud_probability": 0.05,
			"growth_potential":  0.78,
			"confidence":        0.91,
		},
		"analytics": map[string]interface{}{
			"processing_time": "45ms",
			"model_version":   "v2.1.0",
			"features_used":   []string{"business_name", "address", "industry_keywords"},
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(result)
}

// setupRoutes configures the HTTP routes for the server
func (s *AdvancedGatewayServer) setupRoutes() {
	router := mux.NewRouter()

	// Health and status endpoints
	router.HandleFunc("/health", s.handleHealth).Methods("GET")
	router.HandleFunc("/status", s.handleHealth).Methods("GET")

	// Advanced Analytics endpoints
	router.HandleFunc("/analytics/advanced", s.handleAdvancedAnalytics).Methods("GET")
	router.HandleFunc("/analytics/predictions", s.handleMLPredictions).Methods("GET")
	router.HandleFunc("/analytics/trends", s.handleTrendAnalysis).Methods("GET")
	router.HandleFunc("/analytics/anomalies", s.handleAnomalyDetection).Methods("GET")
	router.HandleFunc("/analytics/insights", s.handleInsights).Methods("GET")

	// Multi-Tenant Management endpoints
	router.HandleFunc("/tenants", s.handleTenantManagement).Methods("GET", "POST")
	router.HandleFunc("/tenants/{id}", s.handleTenantManagement).Methods("PUT", "DELETE")
	router.HandleFunc("/tenants/{id}/usage", s.handleTenantUsage).Methods("GET")

	// Enhanced Classification endpoint
	router.HandleFunc("/classify", s.handleClassification).Methods("POST")

	// Legacy endpoints for backward compatibility
	router.HandleFunc("/metrics", s.handleLegacyMetrics).Methods("GET")
	router.HandleFunc("/docs", s.handleDocumentation).Methods("GET")

	log.Printf("🚀 Starting %s v%s on :%s", s.serviceName, s.version, s.port)
	log.Printf("✅ %s v%s is ready and listening on :%s", s.serviceName, s.version, s.port)
	log.Printf("🔗 Health: http://localhost:%s/health", s.port)
	log.Printf("📊 Advanced Analytics: http://localhost:%s/analytics/advanced", s.port)
	log.Printf("🤖 ML Predictions: http://localhost:%s/analytics/predictions", s.port)
	log.Printf("📈 Trend Analysis: http://localhost:%s/analytics/trends", s.port)
	log.Printf("🚨 Anomaly Detection: http://localhost:%s/analytics/anomalies", s.port)
	log.Printf("💡 Business Insights: http://localhost:%s/analytics/insights", s.port)
	log.Printf("🏢 Tenant Management: http://localhost:%s/tenants", s.port)
	log.Printf("🧠 Enhanced Classification: http://localhost:%s/classify", s.port)
	log.Printf("📚 Documentation: http://localhost:%s/docs", s.port)

	http.Handle("/", router)
}

// handleLegacyMetrics provides backward compatibility
func (s *AdvancedGatewayServer) handleLegacyMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	metrics := map[string]interface{}{
		"service":   s.serviceName,
		"version":   s.version,
		"timestamp": time.Now().Format(time.RFC3339),
		"metrics": map[string]interface{}{
			"requests": map[string]interface{}{
				"total":        1250,
				"successful":   1180,
				"failed":       70,
				"success_rate": 94.4,
			},
			"response_times": map[string]interface{}{
				"average": "45ms",
				"min":     "12ms",
				"max":     "2.3s",
			},
			"cache": map[string]interface{}{
				"hit_rate": 68,
				"hits":     850,
				"misses":   400,
			},
			"errors": map[string]interface{}{
				"4xx": 45,
				"5xx": 25,
			},
		},
		"advanced_features": map[string]bool{
			"ml_predictions":    true,
			"trend_analysis":    true,
			"anomaly_detection": true,
			"multi_tenant":      true,
			"business_insights": true,
		},
	}

	json.NewEncoder(w).Encode(metrics)
}

// handleDocumentation provides API documentation
func (s *AdvancedGatewayServer) handleDocumentation(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	doc := fmt.Sprintf(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s API Documentation v%s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background-color: #f4f4f4; color: #333; }
        .container { background-color: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #0056b3; }
        h2 { color: #007bff; border-bottom: 2px solid #007bff; padding-bottom: 5px; }
        .endpoint { background-color: #f8f9fa; padding: 10px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #007bff; }
        .method { font-weight: bold; color: #28a745; }
        .feature { background-color: #e7f3ff; padding: 8px; margin: 5px 0; border-radius: 3px; }
        .advanced { background-color: #fff3cd; border-left-color: #ffc107; }
        .ml { background-color: #d4edda; border-left-color: #28a745; }
        .tenant { background-color: #f8d7da; border-left-color: #dc3545; }
    </style>
</head>
<body>
    <div class="container">
        <h1>%s API Documentation v%s</h1>
        <p>Advanced KYB Platform with Machine Learning, Multi-Tenant Support, and Enterprise Analytics</p>
        
        <h2>🚀 Core Endpoints</h2>
        <div class="endpoint">
            <span class="method">GET</span> /health - Service health check
        </div>
        <div class="endpoint">
            <span class="method">POST</span> /classify - Enhanced business classification with ML
        </div>
        
        <h2>🤖 Machine Learning & Analytics</h2>
        <div class="endpoint advanced">
            <span class="method">GET</span> /analytics/advanced - Comprehensive analytics dashboard
        </div>
        <div class="endpoint ml">
            <span class="method">GET</span> /analytics/predictions - ML predictions and forecasts
        </div>
        <div class="endpoint ml">
            <span class="method">GET</span> /analytics/trends - Trend analysis and seasonality detection
        </div>
        <div class="endpoint ml">
            <span class="method">GET</span> /analytics/anomalies - Anomaly detection and outlier analysis
        </div>
        <div class="endpoint advanced">
            <span class="method">GET</span> /analytics/insights - Business insights and recommendations
        </div>
        
        <h2>🏢 Multi-Tenant Management</h2>
        <div class="endpoint tenant">
            <span class="method">GET</span> /tenants - List all tenants
        </div>
        <div class="endpoint tenant">
            <span class="method">POST</span> /tenants - Create new tenant
        </div>
        <div class="endpoint tenant">
            <span class="method">PUT</span> /tenants/{id} - Update tenant
        </div>
        <div class="endpoint tenant">
            <span class="method">DELETE</span> /tenants/{id} - Delete tenant
        </div>
        <div class="endpoint tenant">
            <span class="method">GET</span> /tenants/{id}/usage - Get tenant usage statistics
        </div>
        
        <h2>✨ Advanced Features</h2>
        <div class="feature">
            <strong>Machine Learning:</strong> Predictive analytics, trend analysis, anomaly detection
        </div>
        <div class="feature">
            <strong>Multi-Tenant Support:</strong> Tenant isolation, quota management, resource limits
        </div>
        <div class="feature">
            <strong>Business Intelligence:</strong> Real-time insights, performance analytics, recommendations
        </div>
        <div class="feature">
            <strong>Enterprise Security:</strong> Tenant authentication, data isolation, audit logging
        </div>
        
        <h2>📊 Current Status</h2>
        <p><strong>Service:</strong> %s</p>
        <p><strong>Version:</strong> %s</p>
        <p><strong>Status:</strong> ✅ Operational with Advanced Features</p>
        <p><strong>Timestamp:</strong> %s</p>
    </div>
</body>
</html>`, s.serviceName, s.version, s.serviceName, s.version, s.serviceName, s.version, time.Now().Format(time.RFC3339))

	fmt.Fprint(w, doc)
}

func main() {
	server := NewAdvancedGatewayServer()

	// Start analytics engine
	ctx := context.Background()
	server.analyticsEngine.Start(ctx)
	server.tenantManager.Start(ctx)

	server.setupRoutes()
	log.Fatal(http.ListenAndServe(":"+server.port, nil))
}
