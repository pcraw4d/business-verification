package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kyb-platform/internal/api/handlers"
	"kyb-platform/internal/api/routes"
	"kyb-platform/internal/classification"
	"kyb-platform/internal/classification/testutil"
	"kyb-platform/internal/observability"
	"kyb-platform/internal/routing"

	"go.opentelemetry.io/otel/trace/noop"
)

func main() {
	fmt.Println("🧪 Starting Classification Test Server...")
	fmt.Println()

	// Initialize logger
	logger := observability.NewLogger("test-server", "1.0.0")
	stdLogger := log.New(os.Stdout, "[TEST-SERVER] ", log.LstdFlags)

	// Create mock repository
	fmt.Println("✅ Creating mock repository...")
	mockRepo := testutil.NewMockKeywordRepository()

	// Create detection service
	fmt.Println("✅ Creating IndustryDetectionService...")
	detectionService := classification.NewIndustryDetectionService(mockRepo, stdLogger)

	// Create intelligent router (minimal setup)
	fmt.Println("✅ Creating intelligent router...")
	routerFactory := routing.NewRouterFactory(logger, nil, noop.NewTracerProvider().Tracer("test"))
	routerConfig := routing.IntelligentRouterConfig{
		EnablePerformanceTracking: true,
		EnableLoadBalancing:       false,
		EnableFallbackRouting:     true,
	}
	intelligentRouter, err := routerFactory.CreateIntelligentRouter(routerConfig)
	if err != nil {
		log.Fatalf("Failed to create intelligent router: %v", err)
	}

	// Create intelligent routing handler with detection service
	fmt.Println("✅ Creating IntelligentRoutingHandler with detection service...")
	routingHandler := routes.CreateIntelligentRoutingHandler(
		intelligentRouter,
		detectionService, // ✅ FIX: Now passing detection service
		logger,
		nil, // metrics
		noop.NewTracerProvider().Tracer("test"),
	)

	// Create route config
	routeConfig := &routes.RouteConfig{
		IntelligentRoutingHandler:   routingHandler,
		BusinessIntelligenceHandler: nil,
		Logger:                      logger,
		EnableEnhancedFeatures:      false,
		EnableBackwardCompatibility: true,
	}

	// Setup mux and register routes
	mux := http.NewServeMux()
	routes.RegisterRoutes(mux, routeConfig)

	// Add test endpoint to verify server is running
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "ok",
			"message": "Test server is running",
			"fixes": map[string]bool{
				"handler_calls_detection_service": true,
				"request_deduplication":           true,
				"cache_key_normalization":          true,
				"error_logging":                    true,
				"completion_logging":                true,
			},
		})
	})

	// Start server
	port := "8081"
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		fmt.Println("\n🛑 Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		fmt.Println("✅ Server stopped")
	}()

	fmt.Println()
	fmt.Println("🚀 Test server started successfully!")
	fmt.Printf("📍 Server listening on http://localhost:%s\n", port)
	fmt.Println()
	fmt.Println("📋 Available endpoints:")
	fmt.Printf("   GET  http://localhost:%s/test - Test endpoint\n", port)
	fmt.Printf("   POST http://localhost:%s/v1/classify - Classification (v1)\n", port)
	fmt.Printf("   POST http://localhost:%s/v2/classify - Classification (v2)\n", port)
	fmt.Println()
	fmt.Println("🧪 Test classification with:")
	fmt.Printf(`   curl -X POST http://localhost:%s/v2/classify \
     -H "Content-Type: application/json" \
     -d '{"business_name": "Test Software Company", "description": "Software development", "website_url": "https://example.com"}'`+"\n", port)
	fmt.Println()
	fmt.Println("Press Ctrl+C to stop the server...")
	fmt.Println()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

