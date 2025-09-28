package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type MonitoringService struct {
	serviceName string
	version     string
	port        string
}

func NewMonitoringService() *MonitoringService {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "kyb-monitoring"
	}

	version := "4.0.0-MONITORING"

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	return &MonitoringService{
		serviceName: serviceName,
		version:     version,
		port:        port,
	}
}

func (s *MonitoringService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   s.serviceName,
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   s.version,
	})
}

func (s *MonitoringService) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Simulate monitoring metrics
	metrics := map[string]interface{}{
		"service":   s.serviceName,
		"version":   s.version,
		"timestamp": time.Now().Format(time.RFC3339),
		"system_metrics": map[string]interface{}{
			"cpu_usage":    "45%",
			"memory_usage": "62%",
			"disk_usage":   "38%",
			"network_io":   "125MB/s",
		},
		"service_health": map[string]interface{}{
			"api_gateway":            "healthy",
			"classification_service": "healthy",
			"merchant_service":       "healthy",
			"redis_cache":            "healthy",
			"database":               "healthy",
		},
		"performance_metrics": map[string]interface{}{
			"avg_response_time":   "45ms",
			"requests_per_second": 1250,
			"error_rate":          "0.8%",
			"uptime":              "99.9%",
		},
		"alerts": []map[string]interface{}{
			{
				"level":     "info",
				"message":   "All services operating normally",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		},
	}

	json.NewEncoder(w).Encode(metrics)
}

func (s *MonitoringService) handleAlerts(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	alerts := map[string]interface{}{
		"service":   s.serviceName,
		"version":   s.version,
		"timestamp": time.Now().Format(time.RFC3339),
		"active_alerts": []map[string]interface{}{
			{
				"id":        "alert-001",
				"level":     "warning",
				"message":   "High memory usage detected",
				"service":   "kyb-classification-service",
				"timestamp": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
				"status":    "active",
			},
			{
				"id":        "alert-002",
				"level":     "info",
				"message":   "Cache hit rate below threshold",
				"service":   "kyb-redis",
				"timestamp": time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
				"status":    "resolved",
			},
		},
		"alert_summary": map[string]interface{}{
			"total_alerts":    2,
			"active_alerts":   1,
			"resolved_alerts": 1,
			"critical_alerts": 0,
		},
	}

	json.NewEncoder(w).Encode(alerts)
}

func (s *MonitoringService) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s Monitoring Dashboard v%s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #2c3e50; color: white; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .metric-card { background: white; padding: 20px; border-radius: 5px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        .metric-title { font-size: 18px; font-weight: bold; margin-bottom: 10px; color: #2c3e50; }
        .metric-value { font-size: 24px; font-weight: bold; color: #27ae60; }
        .status-healthy { color: #27ae60; }
        .status-warning { color: #f39c12; }
        .status-critical { color: #e74c3c; }
        .endpoint { background: #ecf0f1; padding: 15px; margin: 10px 0; border-radius: 3px; }
        .method { background: #3498db; color: white; padding: 3px 8px; border-radius: 3px; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s Monitoring Dashboard v%s</h1>
            <p>Real-time monitoring and alerting for KYB Platform services</p>
        </div>
        
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-title">System Health</div>
                <div class="metric-value status-healthy">99.9%% Uptime</div>
                <p>All services operational</p>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Performance</div>
                <div class="metric-value">45ms Avg Response</div>
                <p>1,250 requests/second</p>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Error Rate</div>
                <div class="metric-value status-healthy">0.8%%</div>
                <p>Below threshold</p>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Active Alerts</div>
                <div class="metric-value status-warning">1 Warning</div>
                <p>High memory usage detected</p>
            </div>
        </div>
        
        <div class="metric-card">
            <h3>Service Status</h3>
            <div class="endpoint">
                <span class="method">GET</span> API Gateway: <span class="status-healthy">Healthy</span>
            </div>
            <div class="endpoint">
                <span class="method">GET</span> Classification Service: <span class="status-healthy">Healthy</span>
            </div>
            <div class="endpoint">
                <span class="method">GET</span> Merchant Service: <span class="status-healthy">Healthy</span>
            </div>
            <div class="endpoint">
                <span class="method">GET</span> Redis Cache: <span class="status-healthy">Healthy</span>
            </div>
        </div>
        
        <div class="metric-card">
            <h3>API Endpoints</h3>
            <div class="endpoint">
                <span class="method">GET</span> /health - Service health check
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /metrics - System metrics
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /alerts - Active alerts
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /dashboard - This dashboard
            </div>
        </div>
    </div>
</body>
</html>`, s.serviceName, s.version, s.serviceName, s.version)

	fmt.Fprint(w, html)
}

func (s *MonitoringService) setupRoutes() {
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/metrics", s.handleMetrics)
	http.HandleFunc("/alerts", s.handleAlerts)
	http.HandleFunc("/dashboard", s.handleDashboard)
	http.HandleFunc("/", s.handleDashboard) // Default to dashboard
}

func main() {
	service := NewMonitoringService()
	service.setupRoutes()

	log.Printf("🚀 Starting %s v%s on :%s", service.serviceName, service.version, service.port)
	log.Printf("✅ %s v%s is ready and listening on :%s", service.serviceName, service.version, service.port)
	log.Printf("🔗 Health: http://localhost:%s/health", service.port)
	log.Printf("📊 Metrics: http://localhost:%s/metrics", service.port)
	log.Printf("🚨 Alerts: http://localhost:%s/alerts", service.port)
	log.Printf("📈 Dashboard: http://localhost:%s/dashboard", service.port)

	if err := http.ListenAndServe(":"+service.port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
