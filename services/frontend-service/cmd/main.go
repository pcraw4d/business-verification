package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Get port from environment variable (Railway sets this)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create file server for static files
	staticDir := "./static"
	fs := http.FileServer(http.Dir(staticDir))

	// Main handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Get the requested path
		path := r.URL.Path

		// If root path, serve merchant portfolio
		if path == "/" {
			http.ServeFile(w, r, filepath.Join(staticDir, "merchant-portfolio.html"))
			return
		}

		// If path ends with /, try to serve index.html
		if strings.HasSuffix(path, "/") {
			indexPath := filepath.Join(staticDir, path, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				http.ServeFile(w, r, indexPath)
				return
			}
		}

		// Check if file exists
		filePath := filepath.Join(staticDir, path)
		if _, err := os.Stat(filePath); err == nil {
			// File exists, serve it
			fs.ServeHTTP(w, r)
			return
		}

		// Try to serve with .html extension
		htmlPath := filePath + ".html"
		if _, err := os.Stat(htmlPath); err == nil {
			http.ServeFile(w, r, htmlPath)
			return
		}

		// If not found, serve merchant portfolio as fallback
		http.ServeFile(w, r, filepath.Join(staticDir, "merchant-portfolio.html"))
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"frontend","version":"1.0.0"}`))
	})

	// API proxy endpoint (for development)
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Proxy to API Gateway
		apiGatewayURL := os.Getenv("API_GATEWAY_URL")
		if apiGatewayURL == "" {
			apiGatewayURL = "https://api-gateway-service-production-21fd.up.railway.app"
		}

		// For now, just return a message that API calls should go directly to the gateway
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Please use the API Gateway directly","gateway_url":"` + apiGatewayURL + `"}`))
	})

	log.Printf("🚀 Starting KYB Frontend Service on port %s", port)
	log.Printf("📁 Serving static files from: %s", staticDir)
	log.Printf("🌐 Frontend URL: http://localhost:%s", port)
	log.Printf("💚 Health check: http://localhost:%s/health", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
