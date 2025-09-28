package main

import (
	"log"
	"net/http"
	"os"
)

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
		w.Write([]byte(`{"service": "kyb-frontend", "status": "healthy", "timestamp": "` +
			`2025-09-28T05:45:00Z", "version": "4.0.0-FRONTEND-404-FIX"}`))
	})

	// Serve static files from public directory
	fs := http.FileServer(http.Dir("./public/"))
	http.Handle("/", fs)

	log.Printf("🌐 Frontend server starting on port %s", port)
	log.Printf("📁 Serving files from ./public/ directory")

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}
