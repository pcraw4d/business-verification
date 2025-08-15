#!/bin/bash

# Minimal Deployment Script for KYB Platform
# This script creates a basic working version for Railway deployment

set -e

echo "🚀 Creating minimal deployment version..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create a minimal main.go for deployment
print_status "Creating minimal main.go for deployment..."

cat > cmd/api/main-minimal.go << 'EOF'
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

	"github.com/joho/godotenv"
)

// MinimalServer represents a minimal API server for deployment
type MinimalServer struct {
	server *http.Server
}

// NewMinimalServer creates a new minimal server
func NewMinimalServer(port string) *MinimalServer {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"version":   "1.0.0-beta",
		})
	})

	// Status endpoint
	mux.HandleFunc("GET /v1/status", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "operational",
			"version":   "1.0.0-beta",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Metrics endpoint
	mux.HandleFunc("GET /v1/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uptime":    time.Since(time.Now()).String(),
			"requests":  0,
			"errors":    0,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Basic classification endpoint (placeholder)
	mux.HandleFunc("POST /v1/classify", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":              true,
			"business_id":          "demo-123",
			"primary_industry":     "Technology",
			"overall_confidence":   0.85,
			"classification_method": "basic",
			"processing_time":      "0.1s",
			"timestamp":            time.Now().UTC().Format(time.RFC3339),
			"message":              "Enhanced classification service coming soon!",
		})
	})

	// Batch classification endpoint (placeholder)
	mux.HandleFunc("POST /v1/classify/batch", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":         true,
			"classifications": []map[string]interface{}{},
			"processing_time": "0.1s",
			"timestamp":       time.Now().UTC().Format(time.RFC3339),
			"message":         "Enhanced batch classification service coming soon!",
		})
	})

	// Web interface
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>KYB Platform - Beta Testing</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; }
        .container { max-width: 800px; margin: 0 auto; }
        .card { background: rgba(255,255,255,0.1); padding: 30px; border-radius: 10px; margin: 20px 0; }
        .status { color: #4ade80; font-weight: bold; }
        .endpoint { background: rgba(0,0,0,0.2); padding: 10px; border-radius: 5px; margin: 10px 0; font-family: monospace; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🚀 KYB Platform - Beta Testing</h1>
        <div class="card">
            <h2>Status: <span class="status">✅ Ready for Beta Testing</span></h2>
            <p>The enhanced classification service is being prepared for deployment. This is a minimal version for testing infrastructure.</p>
        </div>
        
        <div class="card">
            <h3>Available Endpoints:</h3>
            <div class="endpoint">GET /health - Health check</div>
            <div class="endpoint">GET /v1/status - API status</div>
            <div class="endpoint">GET /v1/metrics - Metrics</div>
            <div class="endpoint">POST /v1/classify - Single classification (placeholder)</div>
            <div class="endpoint">POST /v1/classify/batch - Batch classification (placeholder)</div>
        </div>
        
        <div class="card">
            <h3>Enhanced Features Coming Soon:</h3>
            <ul>
                <li>✅ Website Analysis as Primary Method</li>
                <li>✅ Web Search Integration as Secondary Method</li>
                <li>✅ ML Model Integration</li>
                <li>✅ Geographic Region Support</li>
                <li>✅ Industry-Specific Improvements</li>
                <li>✅ Real-time Feedback Collection</li>
                <li>✅ Enhanced Monitoring and Observability</li>
            </ul>
        </div>
        
        <div class="card">
            <h3>Test the API:</h3>
            <p>Try the classification endpoint:</p>
            <div class="endpoint">
                curl -X POST http://localhost:8080/v1/classify \
                  -H "Content-Type: application/json" \
                  -d '{"business_name": "Test Company"}'
            </div>
        </div>
    </div>
</body>
</html>
		`))
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &MinimalServer{
		server: server,
	}
}

// Start starts the server
func (s *MinimalServer) Start() error {
	log.Printf("🚀 Starting KYB Platform minimal server on port %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *MinimalServer) Shutdown(ctx context.Context) error {
	log.Println("🛑 Shutting down server...")
	return s.server.Shutdown(ctx)
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create and start server
	server := NewMinimalServer(port)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Error during server shutdown: %v", err)
		}
	}()

	// Start server
	if err := server.Start(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}
EOF

print_success "Created minimal main.go"

# Create a minimal go.mod for deployment
print_status "Creating minimal go.mod..."

cat > go.mod.minimal << 'EOF'
module github.com/pcraw4d/business-verification

go 1.24

require (
	github.com/joho/godotenv v1.5.1
)

require (
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
EOF

print_success "Created minimal go.mod"

# Create a minimal Dockerfile for deployment
print_status "Creating minimal Dockerfile..."

cat > Dockerfile.minimal << 'EOF'
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod.minimal go.mod
COPY go.sum go.sum

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/api/main-minimal.go cmd/api/main.go

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kyb-platform ./cmd/api

# Create final image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary
COPY --from=builder /app/kyb-platform .

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership
RUN chown -R appuser:appgroup /root/

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./kyb-platform"]
EOF

print_success "Created minimal Dockerfile"

# Build the minimal version
print_status "Building minimal version..."

# Copy the minimal go.mod
cp go.mod.minimal go.mod

# Build the application
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kyb-platform ./cmd/api

if [ $? -eq 0 ]; then
    print_success "Minimal version built successfully"
else
    print_error "Failed to build minimal version"
    exit 1
fi

# Build Docker image
print_status "Building Docker image..."

docker build -f Dockerfile.minimal -t kyb-platform:minimal .

if [ $? -eq 0 ]; then
    print_success "Docker image built successfully"
else
    print_error "Failed to build Docker image"
    exit 1
fi

# Test the minimal version
print_status "Testing minimal version..."

# Start the server in background
./kyb-platform &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test health endpoint
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    print_success "Health check passed"
else
    print_error "Health check failed"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

# Test classification endpoint
if curl -f -X POST http://localhost:8080/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name":"test"}' > /dev/null 2>&1; then
    print_success "Classification endpoint test passed"
else
    print_error "Classification endpoint test failed"
    kill $SERVER_PID 2>/dev/null || true
    exit 1
fi

# Stop the server
kill $SERVER_PID 2>/dev/null || true

print_success "✅ Minimal version is ready for deployment!"

# Create deployment instructions
cat > DEPLOYMENT_READY.md << 'EOF'
# 🚀 KYB Platform - Minimal Deployment Ready!

## ✅ Status: Ready for Railway Deployment

The minimal version of the KYB Platform is now ready for deployment to Railway.

## 📋 What's Included

### ✅ Core Features
- Health check endpoint (`/health`)
- API status endpoint (`/v1/status`)
- Metrics endpoint (`/v1/metrics`)
- Basic classification endpoint (`/v1/classify`) - placeholder
- Batch classification endpoint (`/v1/classify/batch`) - placeholder
- Web interface with status information

### ✅ Infrastructure
- Minimal Go application
- Docker containerization
- Health checks
- Graceful shutdown
- Environment variable support

## 🚀 Deploy to Railway

### Option 1: Use Railway CLI
```bash
# Deploy using the minimal Dockerfile
railway up --dockerfile Dockerfile.minimal
```

### Option 2: Use Railway Dashboard
1. Go to your Railway project
2. Update the Dockerfile path to `Dockerfile.minimal`
3. Deploy the project

### Option 3: Use the deployment script
```bash
# The minimal version is ready for deployment
# Use Railway CLI or dashboard to deploy
```

## 🧪 Test the Deployment

Once deployed, test the endpoints:

```bash
# Health check
curl https://your-railway-url.railway.app/health

# API status
curl https://your-railway-url.railway.app/v1/status

# Classification (placeholder)
curl -X POST https://your-railway-url.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Company"}'
```

## 📊 Next Steps

1. **Deploy to Railway** using the minimal version
2. **Test the infrastructure** and basic endpoints
3. **Restore full features** once the deployment is working
4. **Enable enhanced classification** features

## 🔧 Restoring Full Features

Once the minimal deployment is working, you can restore the full features:

1. Restore the original files:
   ```bash
   mv internal/webanalysis.bak internal/webanalysis
   mv internal/observability/quality_monitor.go.bak internal/observability/quality_monitor.go
   mv internal/webanalysis/search_validator.go.bak internal/webanalysis/search_validator.go
   ```

2. Fix the import conflicts and build issues
3. Update the Dockerfile to use the full version
4. Redeploy with enhanced features

## 📞 Support

For issues during deployment, check:
- Railway logs for deployment errors
- Application logs for runtime errors
- Health check endpoint for system status

---

**Ready for deployment! 🚀**
EOF

print_success "Created deployment instructions"

echo ""
echo "🎉 Minimal deployment version is ready!"
echo ""
echo "📋 Next steps:"
echo "1. Deploy to Railway using Dockerfile.minimal"
echo "2. Test the basic endpoints"
echo "3. Restore full features once deployment is working"
echo ""
echo "📖 See DEPLOYMENT_READY.md for detailed instructions"
echo ""
