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

	// Serve static files from web directory
	fs := http.FileServer(http.Dir("./web/"))
	http.Handle("/", fs)

	log.Printf("🌐 Frontend server starting on port %s", port)
	log.Printf("📁 Serving files from ./web/ directory")
	
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("❌ Server failed to start:", err)
	}
}
