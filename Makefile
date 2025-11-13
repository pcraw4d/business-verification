.PHONY: help dev-api dev-frontend test-api test-frontend deploy-api deploy-frontend clean validate-ml test-ml-validation
.PHONY: start-local stop-local restart-local status-local logs-local logs-local-service build-local clean-local health-local
.PHONY: start-unified

# =============================================================================
# LOCAL DEVELOPMENT (Docker Compose - Mirrors Railway Production)
# =============================================================================

start-local: ## Start all microservices locally using Docker Compose
	@echo "🚀 Starting KYB Platform microservices (Docker Compose)..."
	@if [ ! -f .env.local ]; then \
		echo "⚠️  Warning: .env.local not found. Using railway.env if available."; \
		if [ -f railway.env ]; then \
			docker-compose -f docker-compose.local.yml --env-file railway.env up -d; \
		else \
			echo "❌ Error: No .env.local or railway.env found. Please create .env.local from .env.local.example"; \
			exit 1; \
		fi \
	else \
		docker-compose -f docker-compose.local.yml --env-file .env.local up -d; \
	fi
	@echo "✅ Services starting. Check status with: make status-local"
	@echo "📊 View logs with: make logs-local"
	@echo "🌐 API Gateway: http://localhost:8080"
	@echo "🌐 Frontend: http://localhost:8086"

stop-local: ## Stop all local microservices
	@echo "🛑 Stopping KYB Platform microservices..."
	@docker-compose -f docker-compose.local.yml down
	@echo "✅ All services stopped"

restart-local: stop-local start-local ## Restart all local microservices

status-local: ## Show status of local microservices
	@docker-compose -f docker-compose.local.yml ps

logs-local: ## Show logs from all local microservices
	@docker-compose -f docker-compose.local.yml logs -f

logs-local-service: ## Show logs from a specific service (usage: make logs-local-service SERVICE=api-gateway)
	@if [ -z "$(SERVICE)" ]; then \
		echo "❌ Error: SERVICE not specified. Usage: make logs-local-service SERVICE=api-gateway"; \
		exit 1; \
	fi
	@docker-compose -f docker-compose.local.yml logs -f $(SERVICE)

build-local: ## Build all local microservices Docker images
	@echo "🔨 Building KYB Platform microservices..."
	@docker-compose -f docker-compose.local.yml build
	@echo "✅ Build complete"

clean-local: ## Clean up local Docker containers, volumes, and images
	@echo "🧹 Cleaning up local Docker resources..."
	@docker-compose -f docker-compose.local.yml down -v --rmi local
	@echo "✅ Cleanup complete"

health-local: ## Check health of all local microservices
	@echo "🏥 Checking service health..."
	@echo ""
	@echo "Classification Service:"
	@curl -s http://localhost:8081/health | jq . || echo "❌ Not responding"
	@echo ""
	@echo "Merchant Service:"
	@curl -s http://localhost:8083/health | jq . || echo "❌ Not responding"
	@echo ""
	@echo "Risk Assessment Service:"
	@curl -s http://localhost:8082/health | jq . || echo "❌ Not responding"
	@echo ""
	@echo "API Gateway:"
	@curl -s http://localhost:8080/health | jq . || echo "❌ Not responding"
	@echo ""
	@echo "Frontend:"
	@curl -s http://localhost:8086/health | jq . || echo "❌ Not responding"

# =============================================================================
# UNIFIED SERVER (Quick Development - Single Binary)
# =============================================================================

start-unified: ## Start unified server (cmd/railway-server) for quick development
	@echo "🚀 Starting unified KYB Platform server..."
	@if [ -f railway.env ]; then \
		source railway.env && PORT=$${PORT:-8080} go run ./cmd/railway-server/main.go; \
	else \
		echo "⚠️  Warning: railway.env not found. Starting with defaults."; \
		PORT=8080 go run ./cmd/railway-server/main.go; \
	fi

# =============================================================================
# ORIGINAL COMMANDS
# =============================================================================

help:
	@echo "Available commands:"
	@echo ""
	@echo "Local Development (Microservices - Matches Production):"
	@echo "  make start-local          - Start all microservices with Docker Compose"
	@echo "  make stop-local            - Stop all microservices"
	@echo "  make restart-local        - Restart all microservices"
	@echo "  make status-local         - Show service status"
	@echo "  make logs-local           - Show all service logs"
	@echo "  make logs-local-service SERVICE=api-gateway  - Show specific service logs"
	@echo "  make build-local          - Build Docker images"
	@echo "  make clean-local          - Clean up Docker resources"
	@echo "  make health-local         - Check all service health endpoints"
	@echo ""
	@echo "Unified Server (Quick Development):"
	@echo "  make start-unified         - Start unified server (single binary)"
	@echo ""
	@echo "Original Commands:"
	@echo "  dev-api       - Start API service locally"
	@echo "  dev-frontend  - Start frontend service locally"
	@echo "  test-api      - Run API tests"
	@echo "  test-frontend - Run frontend tests"
	@echo "  deploy-api    - Deploy API to Railway"
	@echo "  deploy-frontend - Deploy frontend to Railway"
	@echo "  validate-ml   - Run ML model validation with cross-validation"
	@echo "  test-ml-validation - Run ML validation tests"
	@echo "  clean         - Clean build artifacts"

dev-api:
	cd services/api && go run cmd/server/main.go

dev-frontend:
	cd services/frontend && go run cmd/main.go

test-api:
	cd services/api && go test ./...

test-frontend:
	cd services/frontend && go test ./...

deploy-api:
	cd services/api && railway up --detach

deploy-frontend:
	cd services/frontend && railway up --detach

validate-ml:
	@echo "Running ML model validation with cross-validation..."
	go run cmd/validate_model.go -k 5 -samples 1000 -time-range 365d -confidence 0.95 -verbose

validate-ml-quick:
	@echo "Running quick ML model validation..."
	go run cmd/validate_model.go -k 3 -samples 100 -time-range 30d -confidence 0.95

validate-ml-comprehensive:
	@echo "Running comprehensive ML model validation..."
	go run cmd/validate_model.go -k 10 -samples 5000 -time-range 730d -confidence 0.99 -output validation_report.json -verbose

test-ml-validation:
	@echo "Running ML validation tests..."
	go test -v ./internal/ml/validation/...

clean:
	find . -name "*.exe" -delete
	find . -name "*.test" -delete
	find . -name "*.out" -delete
	find . -name "validation_report.json" -delete