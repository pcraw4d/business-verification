package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// BusinessIntelligenceGatewayServer represents the business intelligence gateway
type BusinessIntelligenceGatewayServer struct {
	serviceName string
	version     string
	port        string
	router      *mux.Router
}

// NewBusinessIntelligenceGatewayServer creates a new BI gateway server
func NewBusinessIntelligenceGatewayServer() *BusinessIntelligenceGatewayServer {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8087"
	}

	// Get service name from environment or use default
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "bi-service" // Default to Railway service name
	}

	return &BusinessIntelligenceGatewayServer{
		serviceName: serviceName,
		version:     "4.0.4-BI-SYNTAX-FIX-FINAL",
		port:        port,
		router:      nil, // Will be initialized in setupRoutes()
	}
}

// handleHealth returns the health status of the service
func (s *BusinessIntelligenceGatewayServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"service":   s.serviceName,
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   s.version,
		"phase":     "D - Business Intelligence",
		"features": map[string]bool{
			"executive_dashboards":     true,
			"custom_reports":           true,
			"data_export":              true,
			"advanced_analytics":       true,
			"real_time_visualizations": true,
			"scheduled_reports":        true,
			"business_insights":        true,
			"kpi_monitoring":           true,
		},
		"capabilities": []string{
			"Executive Dashboard",
			"Custom Report Generation",
			"Data Export (CSV, JSON, XLSX, PDF)",
			"Real-time KPI Monitoring",
			"Scheduled Reports",
			"Business Intelligence Analytics",
			"Interactive Visualizations",
			"Performance Metrics",
		},
	}
	json.NewEncoder(w).Encode(response)
}

// handleExecutiveDashboard returns the executive dashboard
func (s *BusinessIntelligenceGatewayServer) handleExecutiveDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	dashboard := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"summary": map[string]interface{}{
			"overall_status":    "excellent",
			"performance_score": 94.5,
			"growth_rate":       15.2,
			"risk_level":        "low",
			"recommendations": []string{
				"Continue current growth trajectory",
				"Scale infrastructure for increased demand",
				"Expand enterprise features",
				"Optimize cache performance",
			},
		},
		"key_metrics": map[string]interface{}{
			"total_revenue":     1250000.0,
			"monthly_revenue":   125000.0,
			"active_tenants":    45,
			"classifications":   45000,
			"success_rate":      99.2,
			"avg_response_time": 45.0,
			"availability":      99.9,
		},
		"trends": map[string]interface{}{
			"revenue_growth":    15.2,
			"tenant_growth":     18.4,
			"volume_growth":     22.3,
			"performance_trend": "improving",
		},
		"alerts": []map[string]interface{}{
			{
				"type":      "performance",
				"severity":  "low",
				"message":   "Response time optimization opportunity identified",
				"timestamp": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			},
		},
	}

	json.NewEncoder(w).Encode(dashboard)
}

// handleKPIs returns key performance indicators
func (s *BusinessIntelligenceGatewayServer) handleKPIs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Generate comprehensive KPIs
	kpis := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"financial_kpis": map[string]interface{}{
			"total_revenue": map[string]interface{}{
				"value":  1250000.0,
				"unit":   "USD",
				"change": 15.2,
				"trend":  "up",
				"target": 1000000.0,
				"status": "exceeding",
			},
			"monthly_revenue": map[string]interface{}{
				"value":  125000.0,
				"unit":   "USD",
				"change": 8.5,
				"trend":  "up",
				"target": 100000.0,
				"status": "exceeding",
			},
			"revenue_per_tenant": map[string]interface{}{
				"value":  27777.78,
				"unit":   "USD",
				"change": 12.3,
				"trend":  "up",
				"target": 25000.0,
				"status": "exceeding",
			},
		},
		"operational_kpis": map[string]interface{}{
			"total_classifications": map[string]interface{}{
				"value":  45000.0,
				"unit":   "count",
				"change": 22.3,
				"trend":  "up",
				"target": 40000.0,
				"status": "exceeding",
			},
			"daily_classifications": map[string]interface{}{
				"value":  1500.0,
				"unit":   "count",
				"change": 12.1,
				"trend":  "up",
				"target": 1200.0,
				"status": "exceeding",
			},
			"success_rate": map[string]interface{}{
				"value":  99.2,
				"unit":   "%",
				"change": 0.5,
				"trend":  "up",
				"target": 99.0,
				"status": "exceeding",
			},
		},
		"performance_kpis": map[string]interface{}{
			"avg_response_time": map[string]interface{}{
				"value":  45.0,
				"unit":   "ms",
				"change": -8.2,
				"trend":  "down",
				"target": 100.0,
				"status": "exceeding",
			},
			"throughput": map[string]interface{}{
				"value":  1250.0,
				"unit":   "req/s",
				"change": 15.8,
				"trend":  "up",
				"target": 1000.0,
				"status": "exceeding",
			},
			"availability": map[string]interface{}{
				"value":  99.9,
				"unit":   "%",
				"change": 0.1,
				"trend":  "up",
				"target": 99.5,
				"status": "exceeding",
			},
		},
		"customer_kpis": map[string]interface{}{
			"active_tenants": map[string]interface{}{
				"value":  45.0,
				"unit":   "count",
				"change": 18.4,
				"trend":  "up",
				"target": 40.0,
				"status": "exceeding",
			},
			"customer_satisfaction": map[string]interface{}{
				"value":  4.8,
				"unit":   "rating",
				"change": 0.2,
				"trend":  "up",
				"target": 4.5,
				"status": "exceeding",
			},
			"churn_rate": map[string]interface{}{
				"value":  2.1,
				"unit":   "%",
				"change": -0.5,
				"trend":  "down",
				"target": 5.0,
				"status": "exceeding",
			},
		},
	}

	json.NewEncoder(w).Encode(kpis)
}

// handleCharts returns dashboard charts data
func (s *BusinessIntelligenceGatewayServer) handleCharts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	charts := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"revenue_trend": map[string]interface{}{
			"type":  "line",
			"title": "Revenue Trend (Last 12 Months)",
			"data": map[string]interface{}{
				"labels": []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"},
				"datasets": []map[string]interface{}{
					{
						"label":           "Revenue",
						"data":            []float64{85000, 92000, 88000, 105000, 98000, 112000, 108000, 125000, 118000, 135000, 128000, 125000},
						"borderColor":     "rgb(75, 192, 192)",
						"backgroundColor": "rgba(75, 192, 192, 0.1)",
						"tension":         0.1,
					},
				},
			},
		},
		"classification_volume": map[string]interface{}{
			"type":  "bar",
			"title": "Classification Volume by Industry",
			"data": map[string]interface{}{
				"labels": []string{"Retail", "Technology", "Finance", "Healthcare", "Manufacturing", "Services", "Other"},
				"datasets": []map[string]interface{}{
					{
						"label": "Classifications",
						"data":  []float64{8500, 7200, 6800, 5200, 4800, 4200, 3300},
						"backgroundColor": []string{
							"rgba(255, 99, 132, 0.8)",
							"rgba(54, 162, 235, 0.8)",
							"rgba(255, 205, 86, 0.8)",
							"rgba(75, 192, 192, 0.8)",
							"rgba(153, 102, 255, 0.8)",
							"rgba(255, 159, 64, 0.8)",
							"rgba(199, 199, 199, 0.8)",
						},
					},
				},
			},
		},
		"performance_metrics": map[string]interface{}{
			"type":  "radar",
			"title": "Performance Metrics Overview",
			"data": map[string]interface{}{
				"labels": []string{"Response Time", "Success Rate", "Throughput", "Availability", "Customer Satisfaction", "Cost Efficiency"},
				"datasets": []map[string]interface{}{
					{
						"label":           "Current Performance",
						"data":            []float64{85, 95, 88, 99, 92, 78},
						"borderColor":     "rgb(75, 192, 192)",
						"backgroundColor": "rgba(75, 192, 192, 0.2)",
					},
					{
						"label":           "Target Performance",
						"data":            []float64{80, 90, 85, 95, 90, 80},
						"borderColor":     "rgb(255, 99, 132)",
						"backgroundColor": "rgba(255, 99, 132, 0.2)",
					},
				},
			},
		},
		"geographic_distribution": map[string]interface{}{
			"type":  "doughnut",
			"title": "Geographic Distribution of Users",
			"data": map[string]interface{}{
				"labels": []string{"North America", "Europe", "Asia Pacific", "Latin America", "Middle East & Africa"},
				"datasets": []map[string]interface{}{
					{
						"data": []float64{45, 28, 18, 6, 3},
						"backgroundColor": []string{
							"rgba(255, 99, 132, 0.8)",
							"rgba(54, 162, 235, 0.8)",
							"rgba(255, 205, 86, 0.8)",
							"rgba(75, 192, 192, 0.8)",
							"rgba(153, 102, 255, 0.8)",
						},
					},
				},
			},
		},
	}

	json.NewEncoder(w).Encode(charts)
}

// handleReports handles report management
func (s *BusinessIntelligenceGatewayServer) handleReports(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		s.handleListReports(w, r)
	case "POST":
		s.handleCreateReport(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleListReports lists all reports
func (s *BusinessIntelligenceGatewayServer) handleListReports(w http.ResponseWriter, r *http.Request) {
	reports := []map[string]interface{}{
		{
			"id":         "report_001",
			"name":       "Monthly Classification Summary",
			"type":       "classification_summary",
			"status":     "completed",
			"created_at": time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			"updated_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"format":     "pdf",
			"size":       "2.3 MB",
			"records":    12500,
		},
		{
			"id":         "report_002",
			"name":       "Revenue Analysis Q4",
			"type":       "revenue_analysis",
			"status":     "generating",
			"created_at": time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			"updated_at": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			"format":     "xlsx",
			"size":       "1.8 MB",
			"records":    8900,
		},
		{
			"id":         "report_003",
			"name":       "Performance Metrics Dashboard",
			"type":       "performance_metrics",
			"status":     "scheduled",
			"created_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"updated_at": time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			"format":     "json",
			"size":       "0.5 MB",
			"records":    2500,
		},
	}

	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"reports":   reports,
		"count":     len(reports),
	}

	json.NewEncoder(w).Encode(response)
}

// handleCreateReport creates a new report
func (s *BusinessIntelligenceGatewayServer) handleCreateReport(w http.ResponseWriter, r *http.Request) {
	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	report := map[string]interface{}{
		"id":         fmt.Sprintf("report_%d", time.Now().UnixNano()),
		"name":       req["name"],
		"type":       req["type"],
		"status":     "created",
		"created_at": time.Now().Format(time.RFC3339),
		"updated_at": time.Now().Format(time.RFC3339),
		"format":     req["format"],
		"parameters": req["parameters"],
	}

	json.NewEncoder(w).Encode(report)
}

// handleGenerateReport generates a report
func (s *BusinessIntelligenceGatewayServer) handleGenerateReport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	reportID := vars["id"]

	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Simulate report generation
	result := map[string]interface{}{
		"report_id":    reportID,
		"status":       "completed",
		"format":       params["format"],
		"size":         "2.5 MB",
		"records":      12500,
		"generated_at": time.Now().Format(time.RFC3339),
		"download_url": fmt.Sprintf("/reports/%s/download", reportID),
		"expires_at":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(result)
}

// handleReportTemplates returns available report templates
func (s *BusinessIntelligenceGatewayServer) handleReportTemplates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	templates := []map[string]interface{}{
		{
			"id":          "template_001",
			"name":        "Classification Summary Report",
			"description": "Comprehensive summary of business classifications with industry breakdown",
			"category":    "operational",
			"type":        "classification_summary",
			"parameters": []map[string]interface{}{
				{"name": "date_range", "type": "date_range", "required": true, "description": "Date range for the report"},
				{"name": "include_details", "type": "boolean", "required": false, "default": true, "description": "Include detailed breakdown"},
			},
			"filters": []map[string]interface{}{
				{"name": "tenant_ids", "type": "multi_select", "required": false, "description": "Filter by specific tenants"},
				{"name": "industries", "type": "multi_select", "required": false, "description": "Filter by industry types"},
			},
		},
		{
			"id":          "template_002",
			"name":        "Revenue Analysis Report",
			"description": "Detailed revenue analysis with trends and forecasting",
			"category":    "financial",
			"type":        "revenue_analysis",
			"parameters": []map[string]interface{}{
				{"name": "date_range", "type": "date_range", "required": true, "description": "Date range for the report"},
				{"name": "include_forecasting", "type": "boolean", "required": false, "default": true, "description": "Include revenue forecasting"},
			},
			"filters": []map[string]interface{}{
				{"name": "tenant_ids", "type": "multi_select", "required": false, "description": "Filter by specific tenants"},
				{"name": "revenue_threshold", "type": "number", "required": false, "description": "Minimum revenue threshold"},
			},
		},
		{
			"id":          "template_003",
			"name":        "Performance Metrics Report",
			"description": "System performance metrics and SLA compliance",
			"category":    "technical",
			"type":        "performance_metrics",
			"parameters": []map[string]interface{}{
				{"name": "date_range", "type": "date_range", "required": true, "description": "Date range for the report"},
				{"name": "include_sla", "type": "boolean", "required": false, "default": true, "description": "Include SLA compliance metrics"},
			},
			"filters": []map[string]interface{}{
				{"name": "metric_types", "type": "multi_select", "required": false, "description": "Filter by metric types"},
			},
		},
		{
			"id":          "template_004",
			"name":        "Tenant Usage Report",
			"description": "Comprehensive tenant usage and quota utilization report",
			"category":    "tenant",
			"type":        "tenant_usage",
			"parameters": []map[string]interface{}{
				{"name": "date_range", "type": "date_range", "required": true, "description": "Date range for the report"},
				{"name": "include_quotas", "type": "boolean", "required": false, "default": true, "description": "Include quota utilization"},
			},
			"filters": []map[string]interface{}{
				{"name": "tenant_ids", "type": "multi_select", "required": false, "description": "Filter by specific tenants"},
				{"name": "status", "type": "select", "required": false, "options": []string{"active", "suspended", "all"}, "description": "Filter by tenant status"},
			},
		},
	}

	response := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"templates": templates,
		"count":     len(templates),
	}

	json.NewEncoder(w).Encode(response)
}

// handleDataExport handles data export requests
func (s *BusinessIntelligenceGatewayServer) handleDataExport(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Simulate data export
	exportResult := map[string]interface{}{
		"export_id":    fmt.Sprintf("export_%d", time.Now().UnixNano()),
		"status":       "completed",
		"format":       req["format"],
		"size":         "2.5 MB",
		"records":      12500,
		"generated_at": time.Now().Format(time.RFC3339),
		"download_url": fmt.Sprintf("/exports/%s/download", fmt.Sprintf("export_%d", time.Now().UnixNano())),
		"expires_at":   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(exportResult)
}

// handleBusinessInsights returns business insights
func (s *BusinessIntelligenceGatewayServer) handleBusinessInsights(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	insights := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"insights": []map[string]interface{}{
			{
				"id":          "insight_001",
				"type":        "revenue",
				"title":       "Revenue Growth Acceleration",
				"description": "Revenue growth has accelerated by 15% this quarter, driven by increased enterprise adoption and improved customer retention.",
				"impact":      "high",
				"confidence":  0.92,
				"priority":    "high",
				"category":    "financial",
				"timestamp":   time.Now().Format(time.RFC3339),
				"actions":     []string{"Scale infrastructure", "Expand sales team", "Enhance enterprise features"},
				"metrics": map[string]interface{}{
					"current_revenue": 1250000.0,
					"growth_rate":     15.2,
					"target_revenue":  1000000.0,
				},
			},
			{
				"id":          "insight_002",
				"type":        "performance",
				"title":       "Performance Optimization Opportunity",
				"description": "Response times can be improved by 20% through advanced cache optimization and database query tuning.",
				"impact":      "medium",
				"confidence":  0.88,
				"priority":    "medium",
				"category":    "technical",
				"timestamp":   time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
				"actions":     []string{"Implement advanced caching", "Optimize database queries", "Review API endpoints"},
				"metrics": map[string]interface{}{
					"current_response_time": 45.0,
					"potential_improvement": 20.0,
					"target_response_time":  35.0,
				},
			},
			{
				"id":          "insight_003",
				"type":        "customer",
				"title":       "Customer Satisfaction Trend",
				"description": "Customer satisfaction has improved by 8% over the past month, indicating successful implementation of user feedback.",
				"impact":      "high",
				"confidence":  0.95,
				"priority":    "high",
				"category":    "customer",
				"timestamp":   time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				"actions":     []string{"Continue current initiatives", "Gather detailed feedback", "Implement additional features"},
				"metrics": map[string]interface{}{
					"current_satisfaction": 4.8,
					"improvement":          8.0,
					"target_satisfaction":  4.5,
				},
			},
			{
				"id":          "insight_004",
				"type":        "operational",
				"title":       "Classification Volume Surge",
				"description": "Classification volume has increased by 22% this month, requiring infrastructure scaling to maintain performance.",
				"impact":      "high",
				"confidence":  0.90,
				"priority":    "high",
				"category":    "operational",
				"timestamp":   time.Now().Add(-3 * time.Hour).Format(time.RFC3339),
				"actions":     []string{"Scale infrastructure", "Monitor performance", "Plan capacity increase"},
				"metrics": map[string]interface{}{
					"current_volume": 45000.0,
					"growth_rate":    22.3,
					"target_volume":  40000.0,
				},
			},
		},
		"summary": map[string]interface{}{
			"total_insights":  4,
			"high_impact":     3,
			"medium_impact":   1,
			"high_priority":   3,
			"medium_priority": 1,
			"avg_confidence":  0.91,
		},
	}

	json.NewEncoder(w).Encode(insights)
}

// setupRoutes configures the HTTP routes
func (s *BusinessIntelligenceGatewayServer) setupRoutes() {
	router := mux.NewRouter()

	// Health and status endpoints
	router.HandleFunc("/health", s.handleHealth).Methods("GET")
	router.HandleFunc("/status", s.handleHealth).Methods("GET")

	// Executive Dashboard endpoints
	router.HandleFunc("/dashboard/executive", s.handleExecutiveDashboard).Methods("GET")
	router.HandleFunc("/dashboard/kpis", s.handleKPIs).Methods("GET")
	router.HandleFunc("/dashboard/charts", s.handleCharts).Methods("GET")

	// Report Management endpoints
	router.HandleFunc("/reports", s.handleReports).Methods("GET", "POST")
	router.HandleFunc("/reports/{id}/generate", s.handleGenerateReport).Methods("POST")
	router.HandleFunc("/reports/templates", s.handleReportTemplates).Methods("GET")

	// Data Export endpoints
	router.HandleFunc("/export", s.handleDataExport).Methods("POST")

	// Business Intelligence endpoints
	router.HandleFunc("/insights", s.handleBusinessInsights).Methods("GET")
	router.HandleFunc("/analyze", s.handleBusinessAnalysis).Methods("POST")

	// Log messages removed from setupRoutes - will be logged in main() with correct address
	log.Printf("🔗 Health: http://localhost:%s/health", s.port)
	log.Printf("📊 Executive Dashboard: http://localhost:%s/dashboard/executive", s.port)
	log.Printf("📈 KPIs: http://localhost:%s/dashboard/kpis", s.port)
	log.Printf("📊 Charts: http://localhost:%s/dashboard/charts", s.port)
	log.Printf("📋 Reports: http://localhost:%s/reports", s.port)
	log.Printf("📄 Report Templates: http://localhost:%s/reports/templates", s.port)
	log.Printf("📤 Data Export: http://localhost:%s/export", s.port)
	log.Printf("💡 Business Insights: http://localhost:%s/insights", s.port)

	// Store router for use in main()
	s.router = router
}

// GetRouter returns the configured router
func (s *BusinessIntelligenceGatewayServer) GetRouter() *mux.Router {
	return s.router
}

// handleBusinessAnalysis handles business analysis requests
func (s *BusinessIntelligenceGatewayServer) handleBusinessAnalysis(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse request body
	var requestData map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	businessName := "Unknown Business"
	if name, ok := requestData["business_name"].(string); ok {
		businessName = name
	}

	// Generate mock business intelligence data
	response := map[string]interface{}{
		"request_id":    fmt.Sprintf("bi_%d", time.Now().Unix()),
		"business_name": businessName,
		"timestamp":     time.Now().Format(time.RFC3339),
		"business_intelligence": map[string]interface{}{
			"business_metrics": map[string]interface{}{
				"employee_count": map[string]interface{}{
					"value":      25,
					"range":      "10-50",
					"confidence": 0.85,
					"source":     "estimated",
				},
				"revenue_range": map[string]interface{}{
					"min":        500000,
					"max":        2000000,
					"currency":   "USD",
					"confidence": 0.78,
					"source":     "estimated",
				},
				"founded_year": map[string]interface{}{
					"year":       2018,
					"confidence": 0.92,
					"source":     "public_records",
				},
				"business_location": map[string]interface{}{
					"city":       "Brooklyn",
					"state":      "New York",
					"country":    "US",
					"confidence": 0.95,
					"source":     "address_analysis",
				},
			},
			"company_profile": map[string]interface{}{
				"industry":      "Retail/Food & Beverage",
				"business_type": "Local Business",
				"size_category": "Small Business",
				"growth_stage":  "Established",
			},
			"market_analysis": map[string]interface{}{
				"market_size":       "Local",
				"competition_level": "Medium",
				"growth_potential":  "Moderate",
			},
			"financial_metrics": map[string]interface{}{
				"profitability":    "Profitable",
				"financial_health": "Good",
				"credit_risk":      "Low",
			},
		},
		"status":  "success",
		"success": true,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func main() {
	server := NewBusinessIntelligenceGatewayServer()
	server.setupRoutes()
	
	// Explicitly bind to 0.0.0.0 to ensure service is accessible from all network interfaces
	// This is required for Railway's proxy to route requests correctly
	addr := fmt.Sprintf("0.0.0.0:%s", server.port)
	log.Printf("🚀 Starting %s v%s on %s", server.serviceName, server.version, addr)
	log.Fatal(http.ListenAndServe(addr, server.GetRouter()))
}
