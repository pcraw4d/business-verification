#!/bin/bash

# Codebase Restructuring Script
# This script implements the monorepo structure with clear service separation

set -e  # Exit on any error

echo "🏗️  Starting Codebase Restructuring..."
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
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

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d "cmd" ]; then
    print_error "This script must be run from the project root directory"
    exit 1
fi

# Create backup
print_status "Creating backup of current structure..."
BACKUP_DIR="backup-$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"
cp -r . "$BACKUP_DIR/" 2>/dev/null || true
print_success "Backup created at: $BACKUP_DIR"

# Phase 1: Create new directory structure
print_status "Phase 1: Creating new directory structure..."

mkdir -p services/api/{cmd/server,internal,pkg}
mkdir -p services/frontend/{public,src,cmd}
mkdir -p shared/{types,config,utils}
mkdir -p .github/workflows

print_success "Directory structure created"

# Phase 2: Move backend code
print_status "Phase 2: Moving backend code..."

# Move main API server
if [ -f "cmd/railway-server/main.go" ]; then
    cp cmd/railway-server/main.go services/api/cmd/server/main.go
    print_success "Moved API server main.go"
fi

# Move internal packages
if [ -d "internal" ]; then
    cp -r internal/* services/api/internal/ 2>/dev/null || true
    print_success "Moved internal packages"
fi

# Move pkg if it exists
if [ -d "pkg" ]; then
    cp -r pkg/* services/api/pkg/ 2>/dev/null || true
    print_success "Moved pkg packages"
fi

# Move Go modules
if [ -f "go.mod" ]; then
    cp go.mod services/api/
    print_success "Moved go.mod"
fi

if [ -f "go.sum" ]; then
    cp go.sum services/api/
    print_success "Moved go.sum"
fi

# Phase 3: Move frontend code
print_status "Phase 3: Moving frontend code..."

# Move web files
if [ -d "web" ]; then
    cp -r web/* services/frontend/public/ 2>/dev/null || true
    print_success "Moved web files to frontend/public"
fi

# Move frontend server
if [ -f "cmd/frontend-server/main.go" ]; then
    cp cmd/frontend-server/main.go services/frontend/cmd/main.go
    print_success "Moved frontend server"
fi

# Phase 4: Create service-specific configurations
print_status "Phase 4: Creating service configurations..."

# API Railway config
cat > services/api/railway.json << 'EOF'
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile"
  },
  "deploy": {
    "startCommand": "./server",
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
EOF

# API Dockerfile
cat > services/api/Dockerfile << 'EOF'
# Multi-stage build for API service
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/server .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./server"]
EOF

# Frontend Railway config
cat > services/frontend/railway.json << 'EOF'
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile"
  },
  "deploy": {
    "startCommand": "./frontend-server",
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
EOF

# Frontend Dockerfile
cat > services/frontend/Dockerfile << 'EOF'
# Simple Go-based frontend server
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the frontend server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o frontend-server ./cmd/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/frontend-server .

# Copy public files
COPY --from=builder /app/public ./public

# Expose port
EXPOSE 8080

# Run the frontend server
CMD ["./frontend-server"]
EOF

# Frontend go.mod
cat > services/frontend/go.mod << 'EOF'
module kyb-platform-frontend

go 1.25

require (
    // Add frontend-specific dependencies here
)
EOF

print_success "Service configurations created"

# Phase 5: Create GitHub Actions workflows
print_status "Phase 5: Creating GitHub Actions workflows..."

# API CI/CD workflow
cat > .github/workflows/api-ci.yml << 'EOF'
name: API Service CI/CD

on:
  push:
    paths:
      - 'services/api/**'
      - 'shared/**'
  pull_request:
    paths:
      - 'services/api/**'
      - 'shared/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Test API
        run: |
          cd services/api
          go test ./...
      
      - name: Build API
        run: |
          cd services/api
          go build -o server ./cmd/server/main.go

  deploy:
    if: github.ref == 'refs/heads/main'
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Deploy API to Railway
        run: |
          cd services/api
          railway up --detach
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
EOF

# Frontend CI/CD workflow
cat > .github/workflows/frontend-ci.yml << 'EOF'
name: Frontend Service CI/CD

on:
  push:
    paths:
      - 'services/frontend/**'
  pull_request:
    paths:
      - 'services/frontend/**'

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.25'
      
      - name: Test Frontend
        run: |
          cd services/frontend
          go test ./...
      
      - name: Build Frontend
        run: |
          cd services/frontend
          go build -o frontend-server ./cmd/main.go

  deploy:
    if: github.ref == 'refs/heads/main'
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Deploy Frontend to Railway
        run: |
          cd services/frontend
          railway up --detach
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
EOF

print_success "GitHub Actions workflows created"

# Phase 6: Create development tools
print_status "Phase 6: Creating development tools..."

# Makefile
cat > Makefile << 'EOF'
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
EOF

# Docker Compose for local development
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  api:
    build:
      context: ./services/api
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - SUPABASE_URL=${SUPABASE_URL}
      - SUPABASE_ANON_KEY=${SUPABASE_ANON_KEY}
    volumes:
      - ./services/api:/app
    command: go run cmd/server/main.go

  frontend:
    build:
      context: ./services/frontend
      dockerfile: Dockerfile
    ports:
      - "3000:8080"
    volumes:
      - ./services/frontend:/app
    command: go run cmd/main.go
EOF

print_success "Development tools created"

# Phase 7: Create service READMEs
print_status "Phase 7: Creating service documentation..."

# API README
cat > services/api/README.md << 'EOF'
# KYB Platform API Service

## Overview
This is the backend API service for the KYB Platform, providing business classification, risk assessment, and data enrichment capabilities.

## Features
- Enhanced business classification (MCC, NAICS, SIC codes)
- Website scraping and keyword extraction
- Risk assessment and compliance checking
- Supabase integration for data persistence

## Development

### Local Development
```bash
# From project root
make dev-api

# Or directly
cd services/api
go run cmd/server/main.go
```

### Testing
```bash
make test-api
```

### Deployment
```bash
make deploy-api
```

## API Endpoints
- `GET /health` - Health check
- `POST /v1/classify` - Business classification
- `GET /api/v1/merchants` - Merchant management

## Configuration
Environment variables:
- `SUPABASE_URL` - Supabase project URL
- `SUPABASE_ANON_KEY` - Supabase anonymous key
- `PORT` - Server port (default: 8080)
EOF

# Frontend README
cat > services/frontend/README.md << 'EOF'
# KYB Platform Frontend Service

## Overview
This is the frontend web service for the KYB Platform, providing the user interface for business classification and management.

## Features
- Business classification interface
- Real-time results display
- Enhanced UI with modern design
- API integration for backend services

## Development

### Local Development
```bash
# From project root
make dev-frontend

# Or directly
cd services/frontend
go run cmd/main.go
```

### Testing
```bash
make test-frontend
```

### Deployment
```bash
make deploy-frontend
```

## Configuration
The frontend is configured to call the API service at:
- Production: `https://shimmering-comfort-production.up.railway.app/v1`
- Local: `http://localhost:8080/v1`

## File Structure
- `public/` - Static HTML, CSS, JS files
- `cmd/` - Go-based static file server
- `Dockerfile` - Container configuration
EOF

print_success "Service documentation created"

# Phase 8: Update main README
print_status "Phase 8: Updating main project README..."

cat > README.md << 'EOF'
# KYB Platform

A comprehensive Know Your Business (KYB) platform providing enhanced business classification, risk assessment, and compliance verification.

## Architecture

This project follows a microservices architecture with clear separation between frontend and backend services:

```
kyb-platform/
├── services/
│   ├── api/          # Backend API service
│   └── frontend/     # Frontend web service
├── shared/           # Shared utilities and types
├── docs/            # Documentation
└── scripts/         # Build and deployment scripts
```

## Services

### API Service (`services/api/`)
- **Purpose**: Backend API providing business classification and risk assessment
- **Technology**: Go with enhanced classification algorithms
- **Database**: Supabase integration
- **Deployment**: Railway (shimmering-comfort service)

### Frontend Service (`services/frontend/`)
- **Purpose**: Web interface for business classification
- **Technology**: HTML/CSS/JS with Go static file server
- **Deployment**: Railway (frontend-UI service)

## Quick Start

### Prerequisites
- Go 1.25+
- Railway CLI
- Docker (for local development)

### Local Development
```bash
# Start both services locally
make dev-api      # API on :8080
make dev-frontend # Frontend on :3000

# Or use Docker Compose
docker-compose up
```

### Testing
```bash
make test-api      # Test API service
make test-frontend # Test frontend service
```

### Deployment
```bash
make deploy-api      # Deploy API to Railway
make deploy-frontend # Deploy frontend to Railway
```

## Production URLs
- **API**: https://shimmering-comfort-production.up.railway.app
- **Frontend**: https://frontend-ui-production-e727.up.railway.app

## Development Workflow

### Making Changes
1. **API Changes**: Edit files in `services/api/` - only API service will redeploy
2. **Frontend Changes**: Edit files in `services/frontend/` - only frontend service will redeploy
3. **Shared Changes**: Edit files in `shared/` - both services may redeploy

### CI/CD
- **API Service**: Triggered by changes to `services/api/` or `shared/`
- **Frontend Service**: Triggered by changes to `services/frontend/`
- **Independent Deployments**: Each service deploys independently

## Contributing

1. Create feature branch from `main`
2. Make changes in appropriate service directory
3. Test locally with `make test-<service>`
4. Submit pull request
5. CI/CD will automatically test and deploy

## Documentation
- [API Service](services/api/README.md)
- [Frontend Service](services/frontend/README.md)
- [Development Guidelines](docs/development-guidelines.md)
EOF

print_success "Main README updated"

# Final summary
echo ""
echo "🎉 Codebase Restructuring Complete!"
echo "=================================="
print_success "New structure created with clear service separation"
print_success "Backup available at: $BACKUP_DIR"
echo ""
echo "📋 Next Steps:"
echo "1. Review the new structure in services/ directory"
echo "2. Test local development: make dev-api && make dev-frontend"
echo "3. Update Railway service configurations"
echo "4. Commit and push changes to trigger new deployments"
echo ""
echo "🔧 Available Commands:"
echo "  make help          - Show all available commands"
echo "  make dev-api       - Start API locally"
echo "  make dev-frontend  - Start frontend locally"
echo "  make test-api      - Test API service"
echo "  make test-frontend - Test frontend service"
echo ""
print_warning "Remember to update your Railway service configurations to use the new structure!"
