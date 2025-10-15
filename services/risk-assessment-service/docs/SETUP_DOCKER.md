# Docker Setup Guide

## Overview

This guide covers setting up the Risk Assessment Service using Docker and Docker Compose. Docker provides a consistent development environment across different operating systems and simplifies deployment.

## Prerequisites

### Required Software

- **Docker**: Version 20.10 or later
- **Docker Compose**: Version 2.0 or later
- **Git**: Version 2.30 or later
- **Make**: For running build scripts (optional)

### System Requirements

- **Memory**: 8GB RAM minimum, 16GB recommended
- **Storage**: 20GB free space
- **CPU**: 4 cores minimum, 8 cores recommended

## Installation

### 1. Install Docker

#### macOS

```bash
# Install Docker Desktop for Mac
# Download from: https://www.docker.com/products/docker-desktop

# Or install using Homebrew
brew install --cask docker

# Start Docker Desktop
open /Applications/Docker.app
```

#### Ubuntu/Debian

```bash
# Update package index
sudo apt update

# Install required packages
sudo apt install apt-transport-https ca-certificates curl gnupg lsb-release

# Add Docker's official GPG key
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

# Add Docker repository
echo "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null

# Install Docker
sudo apt update
sudo apt install docker-ce docker-ce-cli containerd.io docker-compose-plugin

# Add user to docker group
sudo usermod -aG docker $USER

# Log out and log back in for group changes to take effect
```

#### Windows

1. Download Docker Desktop for Windows from [Docker website](https://www.docker.com/products/docker-desktop)
2. Run the installer
3. Restart your computer
4. Start Docker Desktop

### 2. Verify Installation

```bash
# Check Docker version
docker --version

# Check Docker Compose version
docker compose version

# Test Docker installation
docker run hello-world
```

## Project Setup

### 1. Clone Repository

```bash
# Clone the repository
git clone https://github.com/kyb-platform/risk-assessment-service.git
cd risk-assessment-service

# Checkout the latest stable version
git checkout main
```

### 2. Environment Configuration

```bash
# Copy example environment file
cp .env.example .env

# Edit environment variables for Docker
nano .env
```

### 3. Docker Environment Variables

```bash
# Database Configuration (Docker)
DATABASE_URL=postgres://kyb_user:kyb_password@postgres:5432/risk_assessment_db
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_TIMEOUT=30s

# Redis Configuration (Docker)
REDIS_URL=redis://redis:6379
REDIS_PASSWORD=
REDIS_DB=0

# Server Configuration
PORT=8080
HOST=0.0.0.0
ENVIRONMENT=development
LOG_LEVEL=debug

# API Configuration
API_VERSION=v1
API_PREFIX=/api/v1
CORS_ORIGINS=http://localhost:3000,http://localhost:8080

# External API Keys (for development)
THOMSON_REUTERS_API_KEY=your_test_key
OFAC_API_KEY=your_test_key
WORLDCHECK_API_KEY=your_test_key

# ML Model Configuration
MODEL_PATH=/app/models
MODEL_UPDATE_INTERVAL=24h
MODEL_CACHE_SIZE=1000

# Security Configuration
JWT_SECRET=your_jwt_secret_key_here
API_KEY_SECRET=your_api_key_secret_here
ENCRYPTION_KEY=your_encryption_key_here

# Monitoring Configuration
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
JAEGER_ENDPOINT=http://jaeger:14268/api/traces

# Development Configuration
DEBUG=true
HOT_RELOAD=true
AUTO_MIGRATE=true
```

## Docker Compose Configuration

### 1. Main Docker Compose File

The project includes a comprehensive `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  # PostgreSQL Database
  postgres:
    image: postgres:15-alpine
    container_name: risk-assessment-postgres
    environment:
      POSTGRES_DB: risk_assessment_db
      POSTGRES_USER: kyb_user
      POSTGRES_PASSWORD: kyb_password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U kyb_user -d risk_assessment_db"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Redis Cache
  redis:
    image: redis:7-alpine
    container_name: risk-assessment-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5

  # Risk Assessment Service
  risk-assessment-service:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: risk-assessment-service
    ports:
      - "8080:8080"
      - "9090:9090"  # Prometheus metrics
    environment:
      - DATABASE_URL=postgres://kyb_user:kyb_password@postgres:5432/risk_assessment_db
      - REDIS_URL=redis://redis:6379
      - ENVIRONMENT=development
      - LOG_LEVEL=debug
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./models:/app/models
      - ./logs:/app/logs
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Background Worker
  worker:
    build:
      context: .
      dockerfile: Dockerfile.worker
    container_name: risk-assessment-worker
    environment:
      - DATABASE_URL=postgres://kyb_user:kyb_password@postgres:5432/risk_assessment_db
      - REDIS_URL=redis://redis:6379
      - ENVIRONMENT=development
      - LOG_LEVEL=debug
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    volumes:
      - ./models:/app/models
      - ./logs:/app/logs

  # Jaeger Tracing
  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: risk-assessment-jaeger
    ports:
      - "16686:16686"  # Jaeger UI
      - "14268:14268"  # Jaeger collector
    environment:
      - COLLECTOR_OTLP_ENABLED=true

  # Prometheus Monitoring
  prometheus:
    image: prom/prometheus:latest
    container_name: risk-assessment-prometheus
    ports:
      - "9091:9090"
    volumes:
      - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'

  # Grafana Dashboard
  grafana:
    image: grafana/grafana:latest
    container_name: risk-assessment-grafana
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
    volumes:
      - grafana_data:/var/lib/grafana
      - ./configs/grafana/dashboards:/etc/grafana/provisioning/dashboards
      - ./configs/grafana/datasources:/etc/grafana/provisioning/datasources

volumes:
  postgres_data:
  redis_data:
  prometheus_data:
  grafana_data:
```

### 2. Development Docker Compose

For development, use `docker-compose.dev.yml`:

```yaml
version: '3.8'

services:
  risk-assessment-service:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/app
      - /app/vendor
    environment:
      - HOT_RELOAD=true
      - DEBUG=true
    command: air

  worker:
    build:
      context: .
      dockerfile: Dockerfile.dev
    volumes:
      - .:/app
      - /app/vendor
    environment:
      - HOT_RELOAD=true
      - DEBUG=true
    command: air -c .air.worker.toml
```

## Dockerfile Configuration

### 1. Main Dockerfile

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates curl

# Create app user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Copy configuration files
COPY --from=builder /app/configs ./configs

# Create directories
RUN mkdir -p /app/models /app/logs
RUN chown -R appuser:appuser /app

# Switch to app user
USER appuser

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
```

### 2. Development Dockerfile

```dockerfile
FROM golang:1.22-alpine

# Install development dependencies
RUN apk add --no-cache git ca-certificates tzdata curl

# Install Air for hot reload
RUN go install github.com/cosmtrek/air@latest

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Expose ports
EXPOSE 8080 9090

# Run with hot reload
CMD ["air"]
```

### 3. Worker Dockerfile

```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the worker
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker cmd/worker/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates

# Create app user
RUN adduser -D -s /bin/sh appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/worker .

# Create directories
RUN mkdir -p /app/models /app/logs
RUN chown -R appuser:appuser /app

# Switch to app user
USER appuser

# Run the worker
CMD ["./worker"]
```

## Running the Service

### 1. Start All Services

```bash
# Start all services
docker-compose up -d

# View running services
docker-compose ps

# View logs
docker-compose logs -f
```

### 2. Start Specific Services

```bash
# Start only database and Redis
docker-compose up -d postgres redis

# Start the main service
docker-compose up -d risk-assessment-service

# Start worker
docker-compose up -d worker
```

### 3. Development Mode

```bash
# Start in development mode with hot reload
docker-compose -f docker-compose.yml -f docker-compose.dev.yml up -d

# View development logs
docker-compose -f docker-compose.yml -f docker-compose.dev.yml logs -f risk-assessment-service
```

## Database Operations

### 1. Run Migrations

```bash
# Run migrations
docker-compose exec risk-assessment-service ./main migrate up

# Or using make
make docker-migrate-up
```

### 2. Seed Data

```bash
# Seed development data
docker-compose exec risk-assessment-service ./main seed dev

# Or using make
make docker-seed-dev
```

### 3. Database Access

```bash
# Connect to PostgreSQL
docker-compose exec postgres psql -U kyb_user -d risk_assessment_db

# Or from host
psql -h localhost -p 5432 -U kyb_user -d risk_assessment_db
```

## Monitoring and Debugging

### 1. View Logs

```bash
# View all logs
docker-compose logs -f

# View specific service logs
docker-compose logs -f risk-assessment-service

# View last 100 lines
docker-compose logs --tail=100 risk-assessment-service
```

### 2. Access Services

- **API Service**: http://localhost:8080
- **Health Check**: http://localhost:8080/health
- **Metrics**: http://localhost:9090/metrics
- **Jaeger UI**: http://localhost:16686
- **Prometheus**: http://localhost:9091
- **Grafana**: http://localhost:3000 (admin/admin)

### 3. Debug Container

```bash
# Execute shell in container
docker-compose exec risk-assessment-service sh

# Check container resources
docker stats risk-assessment-service

# Inspect container
docker inspect risk-assessment-service
```

## Building and Deployment

### 1. Build Images

```bash
# Build all images
docker-compose build

# Build specific service
docker-compose build risk-assessment-service

# Build without cache
docker-compose build --no-cache
```

### 2. Push Images

```bash
# Tag images
docker tag risk-assessment-service:latest your-registry/risk-assessment-service:latest

# Push to registry
docker push your-registry/risk-assessment-service:latest
```

### 3. Production Deployment

```bash
# Use production compose file
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# Scale services
docker-compose up -d --scale risk-assessment-service=3
```

## Testing

### 1. Run Tests

```bash
# Run tests in container
docker-compose exec risk-assessment-service go test ./...

# Run tests with coverage
docker-compose exec risk-assessment-service go test -cover ./...

# Run integration tests
docker-compose exec risk-assessment-service go test -tags=integration ./...
```

### 2. Load Testing

```bash
# Run load tests
docker-compose exec risk-assessment-service go test -tags=load ./test/load/...

# Or using external tool
docker run --rm --network host williamyeh/wrk -t12 -c400 -d30s http://localhost:8080/health
```

## Common Issues and Solutions

### 1. Container Won't Start

**Problem**: Container exits immediately
```bash
# Check container logs
docker-compose logs risk-assessment-service

# Check container status
docker-compose ps

# Inspect container
docker inspect risk-assessment-service
```

**Solution**: Check environment variables and dependencies

### 2. Database Connection Issues

**Problem**: Cannot connect to database
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check database logs
docker-compose logs postgres

# Test connection
docker-compose exec postgres pg_isready -U kyb_user -d risk_assessment_db
```

**Solution**: Ensure PostgreSQL is healthy before starting the service

### 3. Port Conflicts

**Problem**: Port already in use
```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>

# Or change port in docker-compose.yml
```

### 4. Volume Issues

**Problem**: Data not persisting
```bash
# Check volume mounts
docker-compose exec risk-assessment-service ls -la /app

# Check volume data
docker volume ls
docker volume inspect risk-assessment-service_postgres_data
```

### 5. Memory Issues

**Problem**: Container running out of memory
```bash
# Check memory usage
docker stats

# Increase memory limits in docker-compose.yml
services:
  risk-assessment-service:
    deploy:
      resources:
        limits:
          memory: 2G
```

## Performance Optimization

### 1. Multi-stage Builds

```dockerfile
# Use multi-stage builds to reduce image size
FROM golang:1.22-alpine AS builder
# ... build steps ...

FROM alpine:latest
# ... copy only necessary files ...
```

### 2. Layer Caching

```dockerfile
# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code last
COPY . .
```

### 3. Resource Limits

```yaml
services:
  risk-assessment-service:
    deploy:
      resources:
        limits:
          memory: 1G
          cpus: '0.5'
        reservations:
          memory: 512M
          cpus: '0.25'
```

## Security Best Practices

### 1. Non-root User

```dockerfile
# Create and use non-root user
RUN adduser -D -s /bin/sh appuser
USER appuser
```

### 2. Minimal Base Image

```dockerfile
# Use minimal base image
FROM alpine:latest
```

### 3. Secrets Management

```yaml
services:
  risk-assessment-service:
    environment:
      - DATABASE_PASSWORD_FILE=/run/secrets/db_password
    secrets:
      - db_password

secrets:
  db_password:
    file: ./secrets/db_password.txt
```

## Next Steps

1. **Read the API Documentation**: [API_DOCUMENTATION.md](API_DOCUMENTATION.md)
2. **Explore the Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)
3. **Set up Monitoring**: Configure Grafana dashboards
4. **Deploy to Production**: Use production Docker Compose files
5. **Join the Community**: [Community Forum](https://community.kyb-platform.com)

## Support

- **Documentation**: [https://docs.kyb-platform.com](https://docs.kyb-platform.com)
- **GitHub Issues**: [https://github.com/kyb-platform/risk-assessment-service/issues](https://github.com/kyb-platform/risk-assessment-service/issues)
- **Community Forum**: [https://community.kyb-platform.com](https://community.kyb-platform.com)
- **Email Support**: [dev-support@kyb-platform.com](mailto:dev-support@kyb-platform.com)

---

**Last Updated**: January 15, 2024  
**Version**: 2.0.0  
**Next Review**: April 15, 2024
