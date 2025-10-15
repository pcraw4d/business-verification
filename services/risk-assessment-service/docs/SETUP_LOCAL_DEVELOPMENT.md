# Local Development Setup Guide

## Overview

This guide will help you set up the Risk Assessment Service for local development. You'll learn how to run the service locally, configure the database, set up external API integrations, and run tests.

## Prerequisites

### System Requirements

- **Operating System**: macOS, Linux, or Windows 10/11
- **Memory**: 8GB RAM minimum, 16GB recommended
- **Storage**: 10GB free space
- **CPU**: 4 cores minimum, 8 cores recommended

### Required Software

- **Go**: Version 1.22 or later
- **Python**: Version 3.11 or later
- **Node.js**: Version 18 or later
- **Docker**: Version 20.10 or later
- **Docker Compose**: Version 2.0 or later
- **Git**: Version 2.30 or later
- **Make**: For running build scripts

### Development Tools (Recommended)

- **IDE**: VS Code, GoLand, or Vim/Emacs
- **API Client**: Postman, Insomnia, or curl
- **Database Client**: pgAdmin, DBeaver, or psql
- **Version Control**: Git with GitHub CLI

## Installation Steps

### 1. Clone the Repository

```bash
# Clone the repository
git clone https://github.com/kyb-platform/risk-assessment-service.git
cd risk-assessment-service

# Checkout the latest stable version
git checkout main
```

### 2. Install Go Dependencies

```bash
# Install Go dependencies
go mod download
go mod tidy

# Verify Go installation
go version
```

### 3. Install Python Dependencies

```bash
# Create virtual environment
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install Python dependencies
pip install -r requirements.txt
pip install -r requirements-dev.txt

# Verify Python installation
python --version
```

### 4. Install Node.js Dependencies

```bash
# Install Node.js dependencies
npm install

# Install development dependencies
npm install --dev

# Verify Node.js installation
node --version
npm --version
```

## Database Setup

### 1. PostgreSQL Installation

#### Option A: Using Docker (Recommended)

```bash
# Start PostgreSQL with Docker Compose
docker-compose up -d postgres

# Verify PostgreSQL is running
docker-compose ps postgres
```

#### Option B: Local Installation

**macOS (using Homebrew):**
```bash
brew install postgresql
brew services start postgresql
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

**Windows:**
Download and install from [PostgreSQL official website](https://www.postgresql.org/download/windows/)

### 2. Database Configuration

```bash
# Create database user
sudo -u postgres createuser --interactive
# Enter username: kyb_user
# Enter role: y

# Create database
sudo -u postgres createdb risk_assessment_db

# Set password for user
sudo -u postgres psql
```

```sql
-- In PostgreSQL shell
ALTER USER kyb_user PASSWORD 'your_secure_password';
GRANT ALL PRIVILEGES ON DATABASE risk_assessment_db TO kyb_user;
\q
```

### 3. Run Database Migrations

```bash
# Set database URL
export DATABASE_URL="postgres://kyb_user:your_secure_password@localhost:5432/risk_assessment_db"

# Run migrations
make migrate-up

# Verify migrations
make migrate-status
```

## Redis Setup

### 1. Redis Installation

#### Option A: Using Docker (Recommended)

```bash
# Start Redis with Docker Compose
docker-compose up -d redis

# Verify Redis is running
docker-compose ps redis
```

#### Option B: Local Installation

**macOS (using Homebrew):**
```bash
brew install redis
brew services start redis
```

**Ubuntu/Debian:**
```bash
sudo apt install redis-server
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

**Windows:**
Download and install from [Redis official website](https://redis.io/download)

### 2. Redis Configuration

```bash
# Test Redis connection
redis-cli ping
# Should return: PONG

# Set Redis URL
export REDIS_URL="redis://localhost:6379"
```

## Environment Configuration

### 1. Create Environment File

```bash
# Copy example environment file
cp .env.example .env

# Edit environment variables
nano .env
```

### 2. Environment Variables

```bash
# Database Configuration
DATABASE_URL=postgres://kyb_user:your_secure_password@localhost:5432/risk_assessment_db
DB_MAX_CONNECTIONS=25
DB_MAX_IDLE_CONNECTIONS=5
DB_CONNECTION_TIMEOUT=30s

# Redis Configuration
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Server Configuration
PORT=8080
HOST=localhost
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
MODEL_PATH=./models
MODEL_UPDATE_INTERVAL=24h
MODEL_CACHE_SIZE=1000

# Security Configuration
JWT_SECRET=your_jwt_secret_key_here
API_KEY_SECRET=your_api_key_secret_here
ENCRYPTION_KEY=your_encryption_key_here

# Monitoring Configuration
PROMETHEUS_ENABLED=true
PROMETHEUS_PORT=9090
JAEGER_ENDPOINT=http://localhost:14268/api/traces

# Development Configuration
DEBUG=true
HOT_RELOAD=true
AUTO_MIGRATE=true
```

## External API Setup

### 1. Thomson Reuters API

```bash
# Sign up for Thomson Reuters API
# Visit: https://developers.thomsonreuters.com/

# Get API key and configure
export THOMSON_REUTERS_API_KEY="your_api_key"
export THOMSON_REUTERS_BASE_URL="https://api.thomsonreuters.com"
```

### 2. OFAC API

```bash
# Sign up for OFAC API
# Visit: https://ofac.treasury.gov/

# Get API key and configure
export OFAC_API_KEY="your_api_key"
export OFAC_BASE_URL="https://api.ofac.treasury.gov"
```

### 3. World-Check API

```bash
# Sign up for World-Check API
# Visit: https://www.refinitiv.com/en/products/world-check-kyc-screening

# Get API key and configure
export WORLDCHECK_API_KEY="your_api_key"
export WORLDCHECK_BASE_URL="https://api.worldcheck.com"
```

## Running the Service

### 1. Start Dependencies

```bash
# Start all dependencies (PostgreSQL, Redis)
docker-compose up -d

# Verify all services are running
docker-compose ps
```

### 2. Run Database Migrations

```bash
# Run migrations
make migrate-up

# Seed development data
make seed-dev
```

### 3. Start the Service

#### Option A: Using Make (Recommended)

```bash
# Start the service with hot reload
make dev

# Or start without hot reload
make run
```

#### Option B: Direct Go Command

```bash
# Start the service
go run cmd/server/main.go

# Or build and run
go build -o bin/risk-assessment-service cmd/server/main.go
./bin/risk-assessment-service
```

#### Option C: Using Air (Hot Reload)

```bash
# Install Air for hot reload
go install github.com/cosmtrek/air@latest

# Start with hot reload
air
```

### 4. Verify Service is Running

```bash
# Check service health
curl http://localhost:8080/health

# Expected response:
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "version": "2.0.0",
  "uptime": 30
}
```

## Development Workflow

### 1. Code Structure

```
services/risk-assessment-service/
├── cmd/                    # Application entry points
│   ├── server/            # Main API server
│   ├── worker/            # Background worker
│   └── migrate/           # Database migration tool
├── internal/              # Private application code
│   ├── api/              # API layer
│   │   ├── handlers/     # HTTP handlers
│   │   ├── middleware/   # HTTP middleware
│   │   └── routes/       # Route definitions
│   ├── business/         # Business logic
│   │   ├── verification/ # Verification domain
│   │   ├── validation/   # Validation logic
│   │   └── notification/ # Notification logic
│   ├── repository/       # Data access layer
│   │   ├── postgres/     # PostgreSQL implementations
│   │   └── cache/        # Cache implementations
│   ├── external/         # External service clients
│   │   ├── govdata/      # Government data APIs
│   │   └── creditbureau/ # Credit bureau APIs
│   ├── config/           # Configuration management
│   ├── observability/    # Monitoring, logging, metrics
│   └── security/         # Security utilities
├── pkg/                  # Public packages
│   └── client/           # Go client SDK
├── api/                  # API definitions
│   └── openapi/          # OpenAPI specifications
├── scripts/              # Build and deployment scripts
├── docs/                 # Documentation
├── test/                 # Test utilities and fixtures
└── configs/              # Configuration files
```

### 2. Making Changes

```bash
# Create a new feature branch
git checkout -b feature/your-feature-name

# Make your changes
# ... edit files ...

# Run tests
make test

# Run linting
make lint

# Format code
make fmt

# Commit changes
git add .
git commit -m "feat: add your feature description"

# Push to remote
git push origin feature/your-feature-name
```

### 3. Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific test package
go test ./internal/business/verification/...

# Run integration tests
make test-integration

# Run load tests
make test-load
```

### 4. Code Quality

```bash
# Run linting
make lint

# Fix linting issues
make lint-fix

# Format code
make fmt

# Run security scan
make security-scan

# Run dependency check
make deps-check
```

## API Testing

### 1. Using curl

```bash
# Test health endpoint
curl http://localhost:8080/health

# Test risk assessment
curl -X POST http://localhost:8080/api/v1/assess \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_api_key" \
  -d '{
    "business_name": "Test Company",
    "business_address": "123 Test St, Test City, TC 12345",
    "industry": "Technology",
    "country": "US"
  }'
```

### 2. Using Postman

1. Import the OpenAPI specification from `api/openapi.yaml`
2. Set the base URL to `http://localhost:8080`
3. Configure authentication with your test API key
4. Test all endpoints

### 3. Using the Go Client

```go
package main

import (
    "context"
    "fmt"
    "log"
    "github.com/kyb-platform/go-sdk"
)

func main() {
    client := kyb.NewClient(&kyb.Config{
        BaseURL: "http://localhost:8080/api/v1",
        APIKey:  "test_api_key",
    })
    
    ctx := context.Background()
    assessment, err := client.AssessRisk(ctx, &kyb.RiskAssessmentRequest{
        BusinessName:    "Test Company",
        BusinessAddress: "123 Test St, Test City, TC 12345",
        Industry:        "Technology",
        Country:         "US",
    })
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Risk Score: %.2f\n", assessment.RiskScore)
}
```

## Monitoring and Debugging

### 1. Logs

```bash
# View service logs
make logs

# Follow logs in real-time
make logs-follow

# View specific log level
make logs-level=debug
```

### 2. Metrics

```bash
# View Prometheus metrics
curl http://localhost:9090/metrics

# View service metrics
curl http://localhost:8080/metrics
```

### 3. Tracing

```bash
# Start Jaeger for distributed tracing
docker-compose up -d jaeger

# View traces at http://localhost:16686
```

### 4. Database Debugging

```bash
# Connect to database
psql $DATABASE_URL

# View tables
\dt

# View risk assessments
SELECT * FROM risk_assessments LIMIT 10;

# View recent assessments
SELECT id, business_name, risk_score, created_at 
FROM risk_assessments 
ORDER BY created_at DESC 
LIMIT 10;
```

## Common Issues and Solutions

### 1. Database Connection Issues

**Problem**: Cannot connect to PostgreSQL
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# Check database URL
echo $DATABASE_URL

# Test connection
psql $DATABASE_URL -c "SELECT 1;"
```

**Solution**: Ensure PostgreSQL is running and DATABASE_URL is correct

### 2. Redis Connection Issues

**Problem**: Cannot connect to Redis
```bash
# Check if Redis is running
docker-compose ps redis

# Test Redis connection
redis-cli ping
```

**Solution**: Ensure Redis is running and REDIS_URL is correct

### 3. Port Already in Use

**Problem**: Port 8080 is already in use
```bash
# Find process using port 8080
lsof -i :8080

# Kill the process
kill -9 <PID>

# Or use a different port
export PORT=8081
```

### 4. Migration Issues

**Problem**: Database migrations fail
```bash
# Check migration status
make migrate-status

# Reset migrations (development only)
make migrate-reset

# Run migrations again
make migrate-up
```

### 5. External API Issues

**Problem**: External API calls fail
```bash
# Check API keys
echo $THOMSON_REUTERS_API_KEY
echo $OFAC_API_KEY
echo $WORLDCHECK_API_KEY

# Test API connectivity
curl -H "Authorization: Bearer $THOMSON_REUTERS_API_KEY" \
  "$THOMSON_REUTERS_BASE_URL/health"
```

## Development Tips

### 1. Hot Reload

```bash
# Install Air for hot reload
go install github.com/cosmtrek/air@latest

# Start with hot reload
air
```

### 2. Database Seeding

```bash
# Seed development data
make seed-dev

# Seed test data
make seed-test

# Clear all data
make seed-clear
```

### 3. API Key Generation

```bash
# Generate test API key
make generate-test-key

# Generate development API key
make generate-dev-key
```

### 4. Model Training

```bash
# Train models locally
make train-models

# Download pre-trained models
make download-models

# Validate models
make validate-models
```

## Next Steps

1. **Read the API Documentation**: [API_DOCUMENTATION.md](API_DOCUMENTATION.md)
2. **Explore the Architecture**: [ARCHITECTURE.md](ARCHITECTURE.md)
3. **Run the Tests**: `make test`
4. **Start Building**: Create your first feature
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
