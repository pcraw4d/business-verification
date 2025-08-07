# KYB Tool - Development Makefile
# Common development commands and build targets

# Variables
BINARY_NAME=business-verification
BUILD_DIR=build
MAIN_PATH=cmd/api/main.go
DOCKER_IMAGE=business-verification
DOCKER_TAG=latest

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOWORK=$(GOCMD) work

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(shell git describe --tags --always --dirty) -X main.BuildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

.PHONY: help build clean test coverage lint format imports tidy deps run dev docker-build docker-run docker-push install-tools

# Default target
help: ## Show this help message
	@echo "KYB Tool - Development Commands"
	@echo "================================"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build targets
build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

build-linux: ## Build for Linux
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PATH)

build-darwin: ## Build for macOS
	@echo "Building $(BINARY_NAME) for macOS..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin $(MAIN_PATH)

build-windows: ## Build for Windows
	@echo "Building $(BINARY_NAME) for Windows..."
	@mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows.exe $(MAIN_PATH)

build-all: build-linux build-darwin build-windows ## Build for all platforms

# Clean targets
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f $(BINARY_NAME)

# Test targets
test: ## Run all tests
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	$(GOTEST) -race -v ./...

test-coverage: ## Run tests with coverage
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-short: ## Run short tests
	@echo "Running short tests..."
	$(GOTEST) -short ./...

# Code quality targets
lint: ## Run linter
	@echo "Running linter..."
	golangci-lint run

lint-fix: ## Run linter with auto-fix
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix

format: ## Format code
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

imports: ## Organize imports
	@echo "Organizing imports..."
	goimports -w .

# Dependency management
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	$(GOGET) -v -t -d ./...

tidy: ## Tidy go.mod and go.sum
	@echo "Tidying go.mod and go.sum..."
	$(GOMOD) tidy

# Development targets
run: ## Run the application
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run $(MAIN_PATH)

dev: ## Run with hot reload (requires air)
	@echo "Running with hot reload..."
	air

# Docker targets
docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run Docker container
	@echo "Running Docker container..."
	docker run -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-push: ## Push Docker image
	@echo "Pushing Docker image..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)

# Installation targets
install-tools: ## Install development tools
	@echo "Installing development tools..."
	@which golangci-lint > /dev/null || brew install golangci-lint
	@which goimports > /dev/null || $(GOCMD) install golang.org/x/tools/cmd/goimports@latest
	@which godoc > /dev/null || $(GOCMD) install golang.org/x/tools/cmd/godoc@latest
	@which air > /dev/null || $(GOCMD) install github.com/air-verse/air@latest
	@echo "Development tools installed successfully!"

# Documentation targets
docs: ## Generate documentation
	@echo "Generating documentation..."
	godoc -http=:6060 &
	@echo "Documentation available at http://localhost:6060"

# Security targets
security-scan: ## Run security scan
	@echo "Running security scan..."
	golangci-lint run --enable=gosec

# Performance targets
bench: ## Run benchmarks
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. ./...

bench-mem: ## Run benchmarks with memory profiling
	@echo "Running benchmarks with memory profiling..."
	$(GOTEST) -bench=. -benchmem ./...

# Database targets
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@echo "TODO: Implement database migration command"

db-seed: ## Seed database with test data
	@echo "Seeding database with test data..."
	@echo "TODO: Implement database seeding command"

# Monitoring targets
monitor: ## Start monitoring tools
	@echo "Starting monitoring tools..."
	@echo "TODO: Implement monitoring setup"

# Release targets
release: clean build test lint ## Prepare release build
	@echo "Preparing release build..."
	@echo "Release build completed successfully!"

# CI/CD targets
ci: test lint security-scan ## Run CI pipeline
	@echo "CI pipeline completed successfully!"

# Development setup
setup: install-tools deps tidy ## Complete development setup
	@echo "Development environment setup completed!"

# Utility targets
version: ## Show version information
	@echo "Version: $(shell git describe --tags --always --dirty)"
	@echo "Build Time: $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')"
	@echo "Go Version: $(shell go version)"

check: ## Check if all tools are available
	@echo "Checking development tools..."
	@which go > /dev/null || (echo "Go is not installed" && exit 1)
	@which golangci-lint > /dev/null || (echo "golangci-lint is not installed" && exit 1)
	@which goimports > /dev/null || (echo "goimports is not installed" && exit 1)
	@which air > /dev/null || (echo "air is not installed" && exit 1)
	@echo "All development tools are available!"

# Default target
.DEFAULT_GOAL := help
