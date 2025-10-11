.PHONY: help dev-api dev-frontend test-api test-frontend deploy-api deploy-frontend clean validate-ml test-ml-validation

help:
	@echo "Available commands:"
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