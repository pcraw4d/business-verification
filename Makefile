.PHONY: help dev-api dev-frontend test-api test-frontend deploy-api deploy-frontend clean

help:
	@echo "Available commands:"
	@echo "  dev-api       - Start API service locally"
	@echo "  dev-frontend  - Start frontend service locally"
	@echo "  test-api      - Run API tests"
	@echo "  test-frontend - Run frontend tests"
	@echo "  deploy-api    - Deploy API to Railway"
	@echo "  deploy-frontend - Deploy frontend to Railway"
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

clean:
	find . -name "*.exe" -delete
	find . -name "*.test" -delete
	find . -name "*.out" -delete