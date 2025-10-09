# Makefile for Risk Assessment Service

.PHONY: help test test-coverage test-benchmark test-race lint security build run clean

# Default target
help:
	@echo "Available targets:"
	@echo "  test          - Run all tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  test-benchmark- Run benchmark tests"
	@echo "  test-race     - Run race condition tests"
	@echo "  lint          - Run linting"
	@echo "  security      - Run security scan"
	@echo "  build         - Build the service"
	@echo "  run           - Run the service"
	@echo "  clean         - Clean build artifacts"

# Run all tests
test:
	@echo "🧪 Running all tests..."
	go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "📊 Running tests with coverage..."
	@mkdir -p coverage
	go test -v -race -coverprofile=coverage/coverage.out ./...
	go tool cover -html=coverage/coverage.out -o coverage/coverage.html
	@echo "📈 Coverage report generated at coverage/coverage.html"

# Run benchmark tests
test-benchmark:
	@echo "🏃 Running benchmark tests..."
	go test -bench=. -benchmem ./...

# Run race condition tests
test-race:
	@echo "🏁 Running race condition tests..."
	go test -race ./...

# Run linting
lint:
	@echo "🔍 Running linting..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "⚠️  golangci-lint not found, installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run; \
	fi

# Run security scan
security:
	@echo "🔒 Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "⚠️  gosec not found, installing..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

# Build the service
build:
	@echo "🔨 Building the service..."
	go build -o bin/risk-assessment-service ./cmd/main.go

# Run the service
run:
	@echo "🚀 Running the service..."
	go run ./cmd/main.go

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf coverage/
	go clean

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download

# Run the test script
test-script:
	@echo "📜 Running test script..."
	./scripts/run_tests.sh

# Run load tests
load-test:
	@echo "🚀 Running load tests..."
	./scripts/run_load_tests.sh

# Run load test with custom parameters
load-test-custom:
	@echo "🚀 Running custom load test..."
	@read -p "Service URL (default: http://localhost:8080): " url; \
	url=$${url:-http://localhost:8080}; \
	read -p "Duration (default: 5m): " duration; \
	duration=$${duration:-5m}; \
	read -p "Concurrent Users (default: 20): " users; \
	users=$${users:-20}; \
	read -p "Target RPS (default: 16.67): " rps; \
	rps=$${rps:-16.67}; \
	go run ./cmd/load_test.go -url=$$url -duration=$$duration -users=$$users -rps=$$rps -type=load -verbose

# Run stress test
stress-test:
	@echo "🔥 Running stress test..."
	go run ./cmd/load_test.go -type=stress -duration=10m -users=50 -verbose

# Run spike test
spike-test:
	@echo "⚡ Running spike test..."
	go run ./cmd/load_test.go -type=spike -duration=5m -users=30 -verbose

# Railway deployment commands
deploy:
	@echo "🚀 Deploying to Railway..."
	./scripts/deploy_railway.sh

deploy-verify:
	@echo "🔍 Verifying Railway deployment..."
	./scripts/deploy_railway.sh verify

deploy-status:
	@echo "📊 Checking Railway deployment status..."
	./scripts/deploy_railway.sh status

deploy-logs:
	@echo "📋 Viewing Railway logs..."
	./scripts/deploy_railway.sh logs

# Railway CLI commands
railway-login:
	@echo "🔐 Logging into Railway..."
	railway login

railway-status:
	@echo "📊 Railway project status..."
	railway status

railway-logs:
	@echo "📋 Railway logs..."
	railway logs --tail 50

railway-variables:
	@echo "🔐 Railway environment variables..."
	railway variables

railway-domain:
	@echo "🌐 Railway service domain..."
	railway domain

# Beta testing commands
beta-create-tester:
	@echo "👥 Creating beta tester..."
	./scripts/manage_beta_testing.sh create-tester

beta-send-invitation:
	@echo "📧 Sending beta testing invitation..."
	./scripts/manage_beta_testing.sh send-invitation

beta-view-testers:
	@echo "👥 Viewing beta testers..."
	./scripts/manage_beta_testing.sh view-testers

beta-view-feedback:
	@echo "💬 Viewing feedback..."
	./scripts/manage_beta_testing.sh view-feedback

beta-view-stats:
	@echo "📊 Viewing beta testing statistics..."
	./scripts/manage_beta_testing.sh view-feedback-stats

beta-view-program-stats:
	@echo "📊 Viewing beta program statistics..."
	./scripts/manage_beta_testing.sh view-program-stats

beta-performance-test:
	@echo "🚀 Running beta performance test..."
	./scripts/manage_beta_testing.sh performance-test

beta-check-health:
	@echo "🏥 Checking beta testing service health..."
	./scripts/manage_beta_testing.sh check-health

beta-generate-report:
	@echo "📋 Generating beta testing report..."
	./scripts/manage_beta_testing.sh generate-report

beta-start-services:
	@echo "🚀 Starting beta testing services..."
	./scripts/manage_beta_testing.sh start-services

beta-stop-services:
	@echo "🛑 Stopping beta testing services..."
	./scripts/manage_beta_testing.sh stop-services

# Run all quality checks
quality: lint security test-coverage
	@echo "✅ All quality checks completed!"

# CI/CD pipeline
ci: deps lint security test-coverage build
	@echo "✅ CI/CD pipeline completed successfully!"
