# Local Development Guide

This guide explains how to run the KYB Platform locally, matching the Railway production microservices architecture.

## Overview

The KYB Platform can be run in two ways:

1. **Microservices Mode (Recommended)**: Matches Railway production with separate services
2. **Unified Server Mode**: Single binary for quick development/testing

## Prerequisites

- Docker and Docker Compose installed
- Go 1.24+ (for unified server mode)
- Redis (for microservices mode without Docker)
- Access to Supabase project credentials

## Quick Start

### Option 1: Docker Compose (Recommended - Matches Production)

```bash
# 1. Copy environment template
cp .env.local.example .env.local

# 2. Edit .env.local with your Supabase credentials
# Set SUPABASE_URL, SUPABASE_ANON_KEY, DATABASE_URL, etc.

# 3. Start all services
make start-local

# 4. Check service status
make status-local

# 5. View logs
make logs-local

# 6. Check health
make health-local

# 7. Stop services
make stop-local
```

### Option 2: Unified Server (Quick Development)

```bash
# Start single unified server
make start-unified

# Or manually:
source railway.env && go run ./cmd/railway-server/main.go
```

## Service Architecture

### Microservices (Production-like)

| Service | Port | Description | Health Check |
|---------|------|-------------|--------------|
| API Gateway | 8080 | Routes requests to backend services | http://localhost:8080/health |
| Classification Service | 8081 | Business classification | http://localhost:8081/health |
| Risk Assessment Service | 8082 | Risk analysis and predictions | http://localhost:8082/health |
| Merchant Service | 8083 | Merchant management | http://localhost:8083/health |
| Frontend | 8086 | Web UI | http://localhost:8086/health |
| Redis Cache | 6379 | Caching layer | redis-cli ping |

### Service Communication

```
Frontend (8086)
    ↓
API Gateway (8080)
    ├──→ Classification Service (8081)
    ├──→ Merchant Service (8083)
    └──→ Risk Assessment Service (8082)
    ↓
Redis Cache (6379)
```

## Environment Variables

### Required Variables

Create `.env.local` from `.env.local.example`:

```bash
# Supabase Configuration
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your_anon_key
SUPABASE_SERVICE_ROLE_KEY=your_service_role_key
SUPABASE_JWT_SECRET=your_jwt_secret

# Database
DATABASE_URL=postgresql://postgres:password@host:5432/postgres
```

### Service URLs (Auto-configured)

In Docker Compose, service URLs are automatically set to Docker service names:
- `CLASSIFICATION_SERVICE_URL=http://classification-service:8081`
- `MERCHANT_SERVICE_URL=http://merchant-service:8080`
- `RISK_ASSESSMENT_SERVICE_URL=http://risk-assessment-service:8080`

These are configured in `docker-compose.local.yml` and don't need to be set manually.

## Makefile Commands

### Docker Compose Commands

```bash
make start-local          # Start all services
make stop-local           # Stop all services
make restart-local         # Restart all services
make status-local          # Show service status
make logs-local            # Show all logs
make logs-local-service SERVICE=api-gateway  # Show specific service logs
make build-local           # Build Docker images
make clean-local           # Remove containers, volumes, images
make health-local          # Check all service health endpoints
```

### Unified Server Commands

```bash
make start-unified         # Start unified server
```

## Manual Service Startup (Alternative)

If you prefer not to use Docker Compose:

```bash
# Start services directly with Go
./scripts/start-local-services.sh

# Stop services
./scripts/stop-local-services.sh
```

This requires:
- Go 1.24+ installed
- Redis running locally
- All dependencies installed

## Testing Endpoints

### API Gateway

```bash
# Health check
curl http://localhost:8080/health

# Merchant analytics
curl http://localhost:8080/api/v1/merchants/test-123/analytics \
  -H "Authorization: Bearer test-token"

# Risk assessment
curl -X POST http://localhost:8080/api/v1/risk/assess \
  -H "Authorization: Bearer test-token" \
  -H "Content-Type: application/json" \
  -d '{"merchantId":"test-123"}'
```

### Direct Service Access

```bash
# Classification Service
curl http://localhost:8081/health
curl -X POST http://localhost:8081/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Business","description":"A test business"}'

# Merchant Service
curl http://localhost:8083/health
curl http://localhost:8083/api/v1/merchants

# Risk Assessment Service
curl http://localhost:8082/health
```

## Troubleshooting

### Services Not Starting

1. **Check Docker is running**:
   ```bash
   docker ps
   ```

2. **Check environment variables**:
   ```bash
   cat .env.local
   ```

3. **View service logs**:
   ```bash
   make logs-local
   # Or specific service:
   make logs-local-service SERVICE=api-gateway
   ```

4. **Check service health**:
   ```bash
   make health-local
   ```

### Port Conflicts

If ports are already in use:

1. **Find process using port**:
   ```bash
   lsof -i :8080
   ```

2. **Kill process or change port in docker-compose.local.yml**

### Database Connection Issues

1. **Verify DATABASE_URL format**:
   ```
   postgresql://username:password@host:port/database
   ```

2. **Test connection**:
   ```bash
   psql "$DATABASE_URL" -c "SELECT 1;"
   ```

### Service Communication Issues

1. **Check Docker network**:
   ```bash
   docker network ls
   docker network inspect kyb-local-network
   ```

2. **Verify service names match docker-compose.local.yml**

## Development Workflow

### Recommended Workflow

1. **Start services with Docker Compose**:
   ```bash
   make start-local
   ```

2. **Make code changes** in your editor

3. **Rebuild affected service**:
   ```bash
   docker-compose -f docker-compose.local.yml build classification-service
   docker-compose -f docker-compose.local.yml up -d classification-service
   ```

4. **Test changes** via API Gateway or direct service endpoints

5. **View logs**:
   ```bash
   make logs-local-service SERVICE=classification-service
   ```

### Hot Reload (Development)

For faster iteration, use the unified server mode:

```bash
make start-unified
```

This runs a single binary that includes all functionality, making it easier to test changes quickly.

## Differences from Production

### Local (Docker Compose)
- Services communicate via Docker network (service names)
- All services on same machine
- Development logging enabled
- No SSL/TLS between services

### Production (Railway)
- Services communicate via Railway's internal network
- Services may be on different machines
- Production logging
- SSL/TLS enabled

## Next Steps

- See [RAILWAY_DEPLOYMENT.md](./RAILWAY_DEPLOYMENT.md) for production deployment
- See [API_DOCUMENTATION.md](./API_DOCUMENTATION.md) for API details
- See [TESTING.md](./TESTING.md) for testing guidelines

