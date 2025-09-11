# KYB Platform Makefile
# Provides convenient commands for building, testing, and running the application

.PHONY: help build test test-suite test-runner clean install deps lint format

# Default target
help:
	@echo "KYB Platform - Available Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build          Build the application"
	@echo "  build-test     Build the test runner"
	@echo ""
	@echo "Test Commands:"
	@echo "  test           Run all tests"
	@echo "  test-suite     Run automated accuracy test suite"
	@echo "  test-runner    Run test runner with default configuration"
	@echo "  test-verbose   Run tests with verbose output"
	@echo "  test-coverage  Run tests with coverage report"
	@echo ""
	@echo "Development Commands:"
	@echo "  install        Install dependencies"
	@echo "  deps           Download dependencies"
	@echo "  lint           Run linter"
	@echo "  format         Format code"
	@echo "  clean          Clean build artifacts"
	@echo ""
	@echo "Test Suite Commands:"
	@echo "  test-suite-json    Run test suite with JSON output"
	@echo "  test-suite-html    Run test suite with HTML output"
	@echo "  test-suite-xml     Run test suite with XML output"
	@echo "  test-suite-text    Run test suite with text output"
	@echo ""
	@echo "Manual Validation Commands:"
	@echo "  manual-validation  Run manual validation framework"
	@echo "  manual-validator   Build manual validator"
	@echo "  validation-help    Show manual validation help"
	@echo ""
	@echo "Configuration Commands:"
	@echo "  test-config     Show test configuration"
	@echo "  test-help       Show test runner help"

# Build commands
build:
	@echo "🔨 Building KYB Platform..."
	go build -o bin/kyb-platform ./cmd/server

build-test:
	@echo "🔨 Building test runner..."
	go build -o bin/test-runner ./cmd/test-runner

build-manual-validator:
	@echo "🔨 Building manual validator..."
	go build -o bin/manual-validator ./cmd/manual-validator

# Test commands
test:
	@echo "🧪 Running all tests..."
	go test ./... -v

test-suite:
	@echo "🧪 Running automated accuracy test suite..."
	go test -run TestAutomatedAccuracyTestSuite ./test -v

test-runner: build-test
	@echo "🧪 Running test runner..."
	./bin/test-runner

test-verbose:
	@echo "🧪 Running tests with verbose output..."
	go test ./... -v -count=1

test-coverage:
	@echo "🧪 Running tests with coverage..."
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

# Test suite with different output formats
test-suite-json: build-test
	@echo "🧪 Running test suite with JSON output..."
	./bin/test-runner -format json -output ./test-results

test-suite-html: build-test
	@echo "🧪 Running test suite with HTML output..."
	./bin/test-runner -format html -output ./test-results

test-suite-xml: build-test
	@echo "🧪 Running test suite with XML output..."
	./bin/test-runner -format xml -output ./test-results

test-suite-text: build-test
	@echo "🧪 Running test suite with text output..."
	./bin/test-runner -format text -output ./test-results

# Manual validation commands
manual-validation: build-manual-validator
	@echo "🔍 Running manual validation framework..."
	./bin/manual-validator

manual-validator: build-manual-validator
	@echo "🔍 Manual validator built successfully"

validation-help: build-manual-validator
	@echo "📋 Manual Validation Help:"
	./bin/manual-validator -help

# Code mapping validation commands
build-code-mapping-validator:
	@echo "🔨 Building code mapping validator..."
	go build -o bin/code-mapping-validator ./cmd/code-mapping-validator

code-mapping-validation: build-code-mapping-validator
	@echo "🔍 Running industry code mapping validation..."
	./bin/code-mapping-validator

code-mapping-validator: build-code-mapping-validator
	@echo "🔍 Code mapping validator built successfully"

code-mapping-help: build-code-mapping-validator
	@echo "📋 Code Mapping Validation Help:"
	./bin/code-mapping-validator -help

# Confidence calibration validation commands
build-confidence-calibration-validator:
	@echo "🔨 Building confidence calibration validator..."
	go build -o bin/confidence-calibration-validator ./cmd/confidence-calibration-validator

confidence-calibration-validation: build-confidence-calibration-validator
	@echo "🎯 Running confidence score calibration validation..."
	./bin/confidence-calibration-validator

confidence-calibration-validator: build-confidence-calibration-validator
	@echo "🎯 Confidence calibration validator built successfully"

confidence-calibration-help: build-confidence-calibration-validator
	@echo "📋 Confidence Calibration Validation Help:"
	./bin/confidence-calibration-validator -help

# Performance benchmarking commands
build-performance-benchmark-validator:
	@echo "🔨 Building performance benchmark validator..."
	go build -o bin/performance-benchmark-validator ./cmd/performance-benchmark-validator

performance-benchmarking: build-performance-benchmark-validator
	@echo "⚡ Running performance benchmarking..."
	./bin/performance-benchmark-validator

performance-benchmark-validator: build-performance-benchmark-validator
	@echo "⚡ Performance benchmark validator built successfully"

performance-benchmark-help: build-performance-benchmark-validator
	@echo "📋 Performance Benchmarking Help:"
	./bin/performance-benchmark-validator -help

# Development commands
install:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

deps:
	@echo "📦 Downloading dependencies..."
	go mod download

lint:
	@echo "🔍 Running linter..."
	golangci-lint run

format:
	@echo "🎨 Formatting code..."
	go fmt ./...
	goimports -w .

clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -rf bin/
	rm -rf test-results/
	rm -f coverage.out coverage.html
	go clean

# Configuration commands
test-config:
	@echo "📋 Test Configuration:"
	@echo "  Suite Name: KYB Classification Accuracy Test Suite"
	@echo "  Output Directory: ./test-results"
	@echo "  Report Format: json"
	@echo "  Verbose: true"
	@echo "  Parallel Tests: true"
	@echo "  Max Concurrency: 4"
	@echo "  Timeout: 30m"
	@echo "  Retry Count: 2"
	@echo "  Min Accuracy Threshold: 0.7"
	@echo "  Min Performance Threshold: 0.8"
	@echo "  Include Performance: true"
	@echo "  Include Accuracy: true"
	@echo "  Include Reliability: true"
	@echo "  Include Comparison: true"

test-help: build-test
	@echo "📋 Test Runner Help:"
	./bin/test-runner -help

# CI/CD commands
ci-test:
	@echo "🔄 Running CI tests..."
	go test ./... -race -coverprofile=coverage.out
	go tool cover -func=coverage.out

ci-build:
	@echo "🔄 Building for CI..."
	go build -o bin/kyb-platform ./cmd/server
	go build -o bin/test-runner ./cmd/test-runner

# Docker commands (if needed)
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t kyb-platform .

docker-test:
	@echo "🐳 Running tests in Docker..."
	docker run --rm -v $(PWD):/app -w /app golang:1.22 go test ./...

# Performance testing
benchmark:
	@echo "⚡ Running benchmarks..."
	go test -bench=. ./...

benchmark-mem:
	@echo "⚡ Running memory benchmarks..."
	go test -bench=. -benchmem ./...

# Documentation
docs:
	@echo "📚 Generating documentation..."
	godoc -http=:6060

# Security scanning
security:
	@echo "🔒 Running security scan..."
	gosec ./...

# Dependencies check
deps-check:
	@echo "📦 Checking dependencies..."
	go list -u -m all

# Module commands
mod-init:
	@echo "📦 Initializing Go module..."
	go mod init github.com/pcraw4d/business-verification

mod-tidy:
	@echo "📦 Tidying Go module..."
	go mod tidy

mod-verify:
	@echo "📦 Verifying Go module..."
	go mod verify

# Test data generation
test-data:
	@echo "📊 Generating test data..."
	@echo "Test data generation not implemented yet"

# Database commands (if needed)
db-migrate:
	@echo "🗄️ Running database migrations..."
	@echo "Database migrations not implemented yet"

db-seed:
	@echo "🌱 Seeding database..."
	@echo "Database seeding not implemented yet"

# Monitoring commands
monitor:
	@echo "📊 Starting monitoring..."
	@echo "Monitoring not implemented yet"

# Health check
health:
	@echo "🏥 Health check..."
	@echo "Health check not implemented yet"

# Version information
version:
	@echo "📋 Version Information:"
	@echo "  Go Version: $(shell go version)"
	@echo "  Git Commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "  Build Time: $(shell date -u +%Y-%m-%dT%H:%M:%SZ)"

# All-in-one commands
all: clean install build test
	@echo "✅ All tasks completed successfully"

test-all: test test-suite test-coverage
	@echo "✅ All tests completed successfully"

# Development workflow
dev-setup: install deps
	@echo "🚀 Development environment setup complete"

dev-test: format lint test
	@echo "🧪 Development tests completed"

# Production build
prod-build:
	@echo "🏭 Building for production..."
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/kyb-platform ./cmd/server
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/test-runner ./cmd/test-runner

# Quick commands
quick-test:
	@echo "⚡ Quick test run..."
	go test ./test -run TestClassificationAccuracyComprehensive -v

quick-build:
	@echo "⚡ Quick build..."
	go build -o bin/test-runner ./cmd/test-runner

# Help for specific categories
help-test:
	@echo "Test Commands:"
	@echo "  test           - Run all tests"
	@echo "  test-suite     - Run automated accuracy test suite"
	@echo "  test-runner    - Run test runner with default configuration"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  quick-test     - Quick test run"

help-build:
	@echo "Build Commands:"
	@echo "  build          - Build the application"
	@echo "  build-test     - Build the test runner"
	@echo "  prod-build     - Build for production"
	@echo "  quick-build    - Quick build"

help-dev:
	@echo "Development Commands:"
	@echo "  install        - Install dependencies"
	@echo "  deps           - Download dependencies"
	@echo "  lint           - Run linter"
	@echo "  format         - Format code"
	@echo "  clean          - Clean build artifacts"
	@echo "  dev-setup      - Setup development environment"
	@echo "  dev-test       - Run development tests"