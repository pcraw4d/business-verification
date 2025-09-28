package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

type PipelineService struct {
	serviceName string
	version     string
	port        string
}

func NewPipelineService() *PipelineService {
	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		serviceName = "kyb-pipeline-service"
	}

	version := "4.0.0-PIPELINE"

	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	return &PipelineService{
		serviceName: serviceName,
		version:     version,
		port:        port,
	}
}

func (s *PipelineService) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service":   s.serviceName,
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   s.version,
	})
}

func (s *PipelineService) handleProcess(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Simulate pipeline processing
	processingResult := map[string]interface{}{
		"service":         s.serviceName,
		"version":         s.version,
		"timestamp":       time.Now().Format(time.RFC3339),
		"pipeline_status": "processing",
		"stages": []map[string]interface{}{
			{
				"stage":     "data_validation",
				"status":    "completed",
				"duration":  "45ms",
				"timestamp": time.Now().Add(-100 * time.Millisecond).Format(time.RFC3339),
			},
			{
				"stage":     "business_classification",
				"status":    "completed",
				"duration":  "120ms",
				"timestamp": time.Now().Add(-50 * time.Millisecond).Format(time.RFC3339),
			},
			{
				"stage":     "risk_assessment",
				"status":    "in_progress",
				"duration":  "0ms",
				"timestamp": time.Now().Format(time.RFC3339),
			},
			{
				"stage":     "compliance_check",
				"status":    "pending",
				"duration":  "0ms",
				"timestamp": "",
			},
		},
		"metrics": map[string]interface{}{
			"total_processed":     1250,
			"success_rate":        94.4,
			"avg_processing_time": "165ms",
			"queue_size":          12,
		},
	}

	json.NewEncoder(w).Encode(processingResult)
}

func (s *PipelineService) handleQueue(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	queueStatus := map[string]interface{}{
		"service":   s.serviceName,
		"version":   s.version,
		"timestamp": time.Now().Format(time.RFC3339),
		"queue_status": map[string]interface{}{
			"pending_jobs":    12,
			"processing_jobs": 3,
			"completed_jobs":  1235,
			"failed_jobs":     8,
		},
		"workers": []map[string]interface{}{
			{
				"worker_id":       "worker-001",
				"status":          "active",
				"current_job":     "classification-789",
				"processed_today": 156,
			},
			{
				"worker_id":       "worker-002",
				"status":          "active",
				"current_job":     "risk-assessment-790",
				"processed_today": 142,
			},
			{
				"worker_id":       "worker-003",
				"status":          "idle",
				"current_job":     "",
				"processed_today": 98,
			},
		},
	}

	json.NewEncoder(w).Encode(queueStatus)
}

func (s *PipelineService) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	events := map[string]interface{}{
		"service":   s.serviceName,
		"version":   s.version,
		"timestamp": time.Now().Format(time.RFC3339),
		"recent_events": []map[string]interface{}{
			{
				"event_id":    "evt-001",
				"type":        "classification_completed",
				"business_id": "biz-789",
				"timestamp":   time.Now().Add(-2 * time.Minute).Format(time.RFC3339),
				"data": map[string]interface{}{
					"industry":        "Technology",
					"risk_score":      0.15,
					"processing_time": "165ms",
				},
			},
			{
				"event_id":    "evt-002",
				"type":        "risk_assessment_started",
				"business_id": "biz-790",
				"timestamp":   time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
				"data": map[string]interface{}{
					"assessment_type":    "comprehensive",
					"estimated_duration": "300ms",
				},
			},
			{
				"event_id":    "evt-003",
				"type":        "pipeline_error",
				"business_id": "biz-791",
				"timestamp":   time.Now().Add(-30 * time.Second).Format(time.RFC3339),
				"data": map[string]interface{}{
					"error_type":    "validation_failed",
					"error_message": "Invalid business data format",
					"retry_count":   2,
				},
			},
		},
		"event_summary": map[string]interface{}{
			"total_events_today": 1250,
			"successful_events":  1180,
			"failed_events":      70,
			"events_per_minute":  12.5,
		},
	}

	json.NewEncoder(w).Encode(events)
}

func (s *PipelineService) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>%s Pipeline Dashboard v%s</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: #8e44ad; color: white; padding: 20px; border-radius: 5px; margin-bottom: 20px; }
        .metrics-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .metric-card { background: white; padding: 20px; border-radius: 5px; box-shadow: 0 2px 5px rgba(0,0,0,0.1); }
        .metric-title { font-size: 18px; font-weight: bold; margin-bottom: 10px; color: #8e44ad; }
        .metric-value { font-size: 24px; font-weight: bold; color: #27ae60; }
        .status-active { color: #27ae60; }
        .status-pending { color: #f39c12; }
        .status-failed { color: #e74c3c; }
        .endpoint { background: #ecf0f1; padding: 15px; margin: 10px 0; border-radius: 3px; }
        .method { background: #8e44ad; color: white; padding: 3px 8px; border-radius: 3px; font-size: 12px; }
        .stage { padding: 10px; margin: 5px 0; border-radius: 3px; }
        .stage-completed { background: #d5f4e6; border-left: 4px solid #27ae60; }
        .stage-in-progress { background: #fef9e7; border-left: 4px solid #f39c12; }
        .stage-pending { background: #f8f9fa; border-left: 4px solid #95a5a6; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>%s Pipeline Dashboard v%s</h1>
            <p>Event processing pipeline for KYB Platform</p>
        </div>
        
        <div class="metrics-grid">
            <div class="metric-card">
                <div class="metric-title">Processing Status</div>
                <div class="metric-value status-active">94.4%% Success Rate</div>
                <p>1,250 total processed</p>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Queue Status</div>
                <div class="metric-value">12 Pending</div>
                <p>3 currently processing</p>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Performance</div>
                <div class="metric-value">165ms Avg</div>
                <p>Processing time</p>
            </div>
            
            <div class="metric-card">
                <div class="metric-title">Workers</div>
                <div class="metric-value status-active">3 Active</div>
                <p>2 processing, 1 idle</p>
            </div>
        </div>
        
        <div class="metric-card">
            <h3>Pipeline Stages</h3>
            <div class="stage stage-completed">
                <strong>Data Validation</strong> - Completed (45ms)
            </div>
            <div class="stage stage-completed">
                <strong>Business Classification</strong> - Completed (120ms)
            </div>
            <div class="stage stage-in-progress">
                <strong>Risk Assessment</strong> - In Progress
            </div>
            <div class="stage stage-pending">
                <strong>Compliance Check</strong> - Pending
            </div>
        </div>
        
        <div class="metric-card">
            <h3>API Endpoints</h3>
            <div class="endpoint">
                <span class="method">GET</span> /health - Service health check
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /process - Pipeline processing status
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /queue - Queue and worker status
            </div>
            <div class="endpoint">
                <span class="method">GET</span> /events - Recent pipeline events
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

func (s *PipelineService) setupRoutes() {
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/process", s.handleProcess)
	http.HandleFunc("/queue", s.handleQueue)
	http.HandleFunc("/events", s.handleEvents)
	http.HandleFunc("/dashboard", s.handleDashboard)
	http.HandleFunc("/", s.handleDashboard) // Default to dashboard
}

func main() {
	service := NewPipelineService()
	service.setupRoutes()

	log.Printf("🚀 Starting %s v%s on :%s", service.serviceName, service.version, service.port)
	log.Printf("✅ %s v%s is ready and listening on :%s", service.serviceName, service.version, service.port)
	log.Printf("🔗 Health: http://localhost:%s/health", service.port)
	log.Printf("⚙️ Process: http://localhost:%s/process", service.port)
	log.Printf("📋 Queue: http://localhost:%s/queue", service.port)
	log.Printf("📊 Events: http://localhost:%s/events", service.port)
	log.Printf("📈 Dashboard: http://localhost:%s/dashboard", service.port)

	if err := http.ListenAndServe(":"+service.port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
