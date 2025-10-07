package main

import (
	"log"
	"net/http"
	"os"
)

var version = "5.0.2-JAVASCRIPT-FIX-DEPLOY"

func main() {
	// Get port from environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Add health check endpoint first
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"service": "frontend-service", "status": "healthy", "timestamp": "` +
			`2025-10-05T20:00:00Z", "version": "` + version + `"}`))
	})

	// Serve static files from public directory
	fs := http.FileServer(http.Dir("./public/"))

	// Handle root path to serve merchant portfolio
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "./public/merchant-portfolio.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	log.Printf("🌐 Frontend server starting on port %s", port)
	log.Printf("📁 Serving files from ./public/ directory")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}
