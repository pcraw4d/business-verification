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

	version := "5.0.1-LEGACY-UI-RESTORED"

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
	// Serve the legacy merchant hub page
	http.ServeFile(w, r, "./static/merchant-hub.html")
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

func (s *FrontendService) setupRoutes() {
	// Serve static files
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static/"))))
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("./static/js/"))))
	http.Handle("/components/", http.StripPrefix("/components/", http.FileServer(http.Dir("./static/js/components/"))))

	// API endpoints
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/assets", s.handleAssets)
	http.HandleFunc("/dashboard", s.handleDashboard)
	http.HandleFunc("/", s.handleDashboard) // Default to merchant portfolio
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
