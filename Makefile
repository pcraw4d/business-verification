# KYB Platform Makefile
# Provides easy commands for development, testing, and deployment

.PHONY: help install test lint build deploy clean

# Default target
help:
	@echo "KYB Platform - Available Commands:"
	@echo ""
	@echo "Development:"
	@echo "  install     - Install all dependencies"
	@echo "  dev         - Start development server"
	@echo "  dev-frontend - Start frontend development server"
	@echo ""
	@echo "Testing:"
	@echo "  test        - Run all tests"
	@echo "  test-backend - Run backend tests"
	@echo "  test-frontend - Run frontend tests"
	@echo "  test-integration - Run integration tests"
	@echo "  test-e2e    - Run end-to-end tests"
	@echo ""
	@echo "Quality:"
	@echo "  lint        - Run all linting"
	@echo "  lint-backend - Run backend linting"
	@echo "  lint-frontend - Run frontend linting"
	@echo "  security    - Run security scans"
	@echo ""
	@echo "Build:"
	@echo "  build       - Build all components"
	@echo "  build-backend - Build backend"
	@echo "  build-frontend - Build frontend"
	@echo ""
	@echo "Deployment:"
	@echo "  deploy      - Deploy to Railway"
	@echo "  deploy-test - Deploy to test environment"
	@echo ""
	@echo "Utilities:"
	@echo "  clean       - Clean build artifacts"
	@echo "  format      - Format all code"
	@echo "  pre-commit  - Run pre-commit checks"

# Installation
install:
	@echo "Installing Go dependencies..."
	go mod download
	go mod tidy
	@echo "Installing frontend dependencies..."
	cd web && npm install
	@echo "Installing pre-commit hooks..."
	pip install pre-commit
	pre-commit install

# Development
dev:
	@echo "Starting development server..."
	go run ./cmd/railway-server/main.go

dev-frontend:
	@echo "Starting frontend development server..."
	cd web && npm run dev

# Testing
test: test-backend test-frontend test-integration

test-backend:
	@echo "Running backend tests..."
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

test-frontend:
	@echo "Running frontend tests..."
	cd web && npm test

test-integration:
	@echo "Running integration tests..."
	cd web && node integration-test.js

test-e2e:
	@echo "Running end-to-end tests..."
	@echo "Starting application..."
	go run ./cmd/railway-server/main.go &
	@sleep 5
	@echo "Testing API endpoints..."
	curl -f http://localhost:8080/health || exit 1
	curl -f -X POST http://localhost:8080/v1/classify \
		-H "Content-Type: application/json" \
		-d '{"business_name": "Test Company", "description": "Test"}' || exit 1
	curl -f http://localhost:8080/api/v1/merchants || exit 1
	@echo "End-to-end tests passed!"

# Linting
lint: lint-backend lint-frontend

lint-backend:
	@echo "Running backend linting..."
	go vet ./...
	go fmt ./...

lint-frontend:
	@echo "Running frontend linting..."
	cd web && npm run lint

# Security
security:
	@echo "Running security scans..."
	go list -json -deps ./... | nancy sleuth
	cd web && npm audit

# Build
build: build-backend build-frontend

build-backend:
	@echo "Building backend..."
	go build -ldflags="-s -w" -o kyb-platform ./cmd/railway-server

build-frontend:
	@echo "Building frontend..."
	cd web && npm run build

# Deployment
deploy:
	@echo "Deploying to Railway..."
	railway up

deploy-test:
	@echo "Deploying to test environment..."
	railway up --environment test

# Utilities
clean:
	@echo "Cleaning build artifacts..."
	rm -f kyb-platform
	rm -f coverage.out coverage.html
	cd web && rm -rf dist node_modules/.cache
	rm -rf .git/hooks/pre-commit

format:
	@echo "Formatting code..."
	go fmt ./...
	cd web && npm run format

pre-commit:
	@echo "Running pre-commit checks..."
	pre-commit run --all-files

# CI/CD helpers
ci-test: install test lint security
	@echo "CI tests completed successfully!"

ci-build: build
	@echo "CI build completed successfully!"

ci-deploy: ci-test ci-build deploy
	@echo "CI deployment completed successfully!"

# Database helpers
db-migrate:
	@echo "Running database migrations..."
	@echo "Please run the migration scripts manually:"
	@echo "psql -h your_supabase_host -U postgres -d postgres -f supabase-full-integration-migration.sql"

db-test:
	@echo "Testing database connection..."
	curl -f https://shimmering-comfort-production.up.railway.app/health

# Monitoring
monitor:
	@echo "Monitoring application health..."
	watch -n 5 'curl -s https://shimmering-comfort-production.up.railway.app/health | jq .'

# Documentation
docs:
	@echo "Generating documentation..."
	godoc -http=:6060
	@echo "Documentation available at http://localhost:6060"

# Performance testing
perf-test:
	@echo "Running performance tests..."
	ab -n 100 -c 10 https://shimmering-comfort-production.up.railway.app/health
	ab -n 100 -c 10 https://shimmering-comfort-production.up.railway.app/api/v1/merchants