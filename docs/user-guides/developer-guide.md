# Enhanced Business Intelligence System - Developer Guide

## Table of Contents

1. [Development Environment Setup](#development-environment-setup)
2. [Project Structure](#project-structure)
3. [API Integration](#api-integration)
4. [SDK Usage](#sdk-usage)
5. [Testing](#testing)
6. [Contributing](#contributing)
7. [Debugging](#debugging)
8. [Performance Optimization](#performance-optimization)
9. [Security Best Practices](#security-best-practices)
10. [Deployment](#deployment)

## Development Environment Setup

### Prerequisites

Before setting up your development environment, ensure you have:

#### Required Software
- **Go**: Version 1.22 or later
- **Docker**: Version 20.10 or later
- **Docker Compose**: Version 2.0 or later
- **Git**: Version 2.25 or later
- **Make**: For build automation
- **PostgreSQL**: Version 13 or later (or Supabase)
- **Redis**: Version 6 or later

#### Optional Tools
- **VS Code**: Recommended IDE with Go extension
- **Postman**: API testing tool
- **pgAdmin**: PostgreSQL administration tool
- **Redis Commander**: Redis administration tool

### Local Development Setup

#### 1. Clone the Repository

```bash
# Clone the repository
git clone https://github.com/your-org/kyb-platform.git
cd kyb-platform

# Checkout the development branch
git checkout develop

# Install Git hooks
make install-hooks
```

#### 2. Environment Configuration

```bash
# Copy development environment file
cp configs/development.env.example configs/development.env

# Edit the configuration
nano configs/development.env
```

**Development Environment Configuration**:

```bash
# configs/development.env
ENVIRONMENT=development
LOG_LEVEL=debug
API_PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=kyb_platform_dev
DB_USER=kyb_dev
DB_PASSWORD=dev_password
DB_SSL_MODE=disable

# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# Security Configuration (for development only)
JWT_SECRET=dev-jwt-secret-key-change-in-production
API_KEY_SECRET=dev-api-key-secret-change-in-production
ENCRYPTION_KEY=dev-encryption-key-change-in-production

# External Services
EXTERNAL_API_TIMEOUT=30s
EXTERNAL_API_RETRIES=3

# Development Features
DEBUG_MODE=true
ENABLE_METRICS=true
ENABLE_TRACING=true
ENABLE_PROFILING=true

# Testing Configuration
TEST_DB_NAME=kyb_platform_test
TEST_REDIS_DB=1
```

#### 3. Database Setup

```bash
# Start PostgreSQL and Redis with Docker
docker-compose -f docker-compose.dev.yml up -d postgres redis

# Create development database
docker exec -it kyb-platform-postgres psql -U postgres -c "
CREATE DATABASE kyb_platform_dev;
CREATE USER kyb_dev WITH PASSWORD 'dev_password';
GRANT ALL PRIVILEGES ON DATABASE kyb_platform_dev TO kyb_dev;
"

# Run database migrations
make migrate-dev

# Seed development data
make seed-dev
```

#### 4. Application Setup

```bash
# Install Go dependencies
go mod download

# Build the application
make build-dev

# Start the development server
make run-dev
```

#### 5. Verify Setup

```bash
# Check if the application is running
curl http://localhost:8080/health

# Check API documentation
open http://localhost:8080/docs

# Check metrics endpoint
curl http://localhost:8080/metrics
```

### Development Tools

#### 1. Code Quality Tools

```bash
# Install development tools
make install-tools

# Run code formatting
make fmt

# Run linting
make lint

# Run security checks
make security-check

# Run all quality checks
make quality-check
```

#### 2. Testing Tools

```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Run performance tests
make test-performance

# Generate test coverage
make test-coverage

# Run all tests
make test-all
```

#### 3. Development Scripts

```bash
# Start development environment
make dev-start

# Stop development environment
make dev-stop

# Restart development environment
make dev-restart

# View logs
make dev-logs

# Clean development environment
make dev-clean
```

## Project Structure

### Directory Layout

```
kyb-platform/
├── cmd/                    # Application entry points
│   ├── api/               # Main API server
│   ├── worker/            # Background worker
│   └── migrate/           # Database migration tool
├── internal/              # Private application code
│   ├── api/              # API layer
│   │   ├── handlers/     # HTTP handlers
│   │   ├── middleware/   # HTTP middleware
│   │   └── routes/       # Route definitions
│   ├── business/         # Business logic
│   │   ├── classification/ # Classification domain
│   │   ├── risk/         # Risk assessment domain
│   │   └── discovery/    # Data discovery domain
│   ├── repository/       # Data access layer
│   ├── external/         # External service clients
│   ├── config/           # Configuration management
│   └── shared/           # Shared utilities
├── pkg/                  # Public packages
│   ├── client/           # Go client SDK
│   └── validators/       # Validation utilities
├── api/                  # API definitions
│   └── openapi/          # OpenAPI specifications
├── configs/              # Configuration files
├── scripts/              # Build and deployment scripts
├── test/                 # Test utilities and fixtures
├── docs/                 # Documentation
└── deployments/          # Deployment configurations
```

### Key Components

#### 1. API Layer (`internal/api/`)

```go
// Example handler structure
package handlers

import (
    "net/http"
    "github.com/your-org/kyb-platform/internal/business/classification"
)

type ClassificationHandler struct {
    service classification.Service
    logger  *zap.Logger
}

func (h *ClassificationHandler) ClassifyBusiness(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

#### 2. Business Logic (`internal/business/`)

```go
// Example service structure
package classification

import (
    "context"
    "github.com/your-org/kyb-platform/internal/repository"
)

type Service interface {
    ClassifyBusiness(ctx context.Context, input ClassificationInput) (*ClassificationResult, error)
}

type service struct {
    repo repository.ClassificationRepository
    cache cache.Cache
    logger *zap.Logger
}
```

#### 3. Repository Layer (`internal/repository/`)

```go
// Example repository structure
package repository

import (
    "context"
    "database/sql"
)

type ClassificationRepository interface {
    Save(ctx context.Context, classification *Classification) error
    GetByID(ctx context.Context, id string) (*Classification, error)
    List(ctx context.Context, filters Filters) ([]*Classification, error)
}

type postgresRepository struct {
    db *sql.DB
}
```

## API Integration

### Authentication

#### 1. API Key Authentication

```bash
# Generate API key
curl -X POST "http://localhost:8080/api/v3/admin/api-keys" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN" \
  -d '{
    "name": "Development API Key",
    "permissions": ["read", "write"],
    "expires_at": "2025-12-31T23:59:59Z"
  }'
```

#### 2. JWT Authentication

```bash
# Login to get JWT token
curl -X POST "http://localhost:8080/api/v3/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "developer@company.com",
    "password": "your-password"
  }'
```

### API Usage Examples

#### 1. Business Classification

```bash
# Classify a business
curl -X POST "http://localhost:8080/api/v3/classify" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_name": "Acme Corporation",
    "website_url": "https://www.acme.com",
    "description": "Technology solutions provider"
  }'
```

**Response**:
```json
{
  "business_id": "biz_12345",
  "classification": {
    "primary_industry": "Technology",
    "secondary_industries": ["Software", "Consulting"],
    "business_size": "Medium",
    "business_type": "B2B",
    "geographic_scope": ["United States", "Global"],
    "confidence_score": 0.92,
    "classification_methods": ["hybrid", "ml_based"]
  },
  "metadata": {
    "classification_date": "2024-12-19T10:30:00Z",
    "processing_time": "2.3s"
  }
}
```

#### 2. Risk Assessment

```bash
# Perform risk assessment
curl -X POST "http://localhost:8080/api/v3/risk/assess" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_id": "biz_12345",
    "assessment_type": "comprehensive",
    "risk_factors": ["industry", "geographic", "size", "compliance"]
  }'
```

**Response**:
```json
{
  "assessment_id": "risk_67890",
  "business_id": "biz_12345",
  "risk_assessment": {
    "overall_risk_score": 0.65,
    "risk_level": "medium",
    "risk_factors": {
      "industry_risk": 0.7,
      "geographic_risk": 0.3,
      "size_risk": 0.8,
      "compliance_risk": 0.6
    },
    "recommendations": [
      "Monitor compliance status regularly",
      "Conduct due diligence on business partners"
    ]
  },
  "metadata": {
    "assessment_date": "2024-12-19T10:30:00Z",
    "processing_time": "5.2s"
  }
}
```

#### 3. Data Discovery

```bash
# Discover data for a business
curl -X POST "http://localhost:8080/api/v3/discovery/start" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "business_id": "biz_12345",
    "discovery_types": ["business_info", "compliance_data", "risk_indicators"],
    "data_sources": ["business_registry", "news_articles", "financial_reports"]
  }'
```

**Response**:
```json
{
  "discovery_id": "disc_11111",
  "business_id": "biz_12345",
  "status": "in_progress",
  "discovery_config": {
    "types": ["business_info", "compliance_data", "risk_indicators"],
    "sources": ["business_registry", "news_articles", "financial_reports"]
  },
  "metadata": {
    "started_at": "2024-12-19T10:30:00Z",
    "estimated_completion": "2024-12-19T10:35:00Z"
  }
}
```

### Error Handling

#### 1. Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid business name provided",
    "details": {
      "field": "business_name",
      "constraint": "required",
      "value": null
    },
    "request_id": "req_12345",
    "timestamp": "2024-12-19T10:30:00Z"
  }
}
```

#### 2. Common Error Codes

- **VALIDATION_ERROR**: Input validation failed
- **AUTHENTICATION_ERROR**: Invalid or missing authentication
- **AUTHORIZATION_ERROR**: Insufficient permissions
- **NOT_FOUND**: Resource not found
- **RATE_LIMIT_EXCEEDED**: Rate limit exceeded
- **INTERNAL_ERROR**: Internal server error

#### 3. Rate Limiting

```bash
# Check rate limit headers
curl -I "http://localhost:8080/api/v3/classify" \
  -H "Authorization: Bearer YOUR_API_KEY"

# Response headers
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 95
X-RateLimit-Reset: 1640000000
```

## SDK Usage

### Go SDK

#### 1. Installation

```bash
go get github.com/your-org/kyb-platform/pkg/client
```

#### 2. Basic Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/your-org/kyb-platform/pkg/client"
)

func main() {
    // Create client
    c := client.NewClient("https://api.kyb-platform.com", "your-api-key")
    
    // Classify business
    result, err := c.ClassifyBusiness(context.Background(), client.ClassificationInput{
        BusinessName: "Acme Corporation",
        WebsiteURL:   "https://www.acme.com",
        Description:  "Technology solutions provider",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Classification: %+v\n", result)
}
```

#### 3. Advanced Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "github.com/your-org/kyb-platform/pkg/client"
)

func main() {
    // Create client with custom configuration
    c := client.NewClientWithConfig(client.Config{
        BaseURL:     "https://api.kyb-platform.com",
        APIKey:      "your-api-key",
        Timeout:     30 * time.Second,
        Retries:     3,
        UserAgent:   "MyApp/1.0",
    })
    
    // Perform risk assessment
    riskResult, err := c.AssessRisk(context.Background(), client.RiskAssessmentInput{
        BusinessID:     "biz_12345",
        AssessmentType: "comprehensive",
        RiskFactors:    []string{"industry", "geographic", "size"},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Risk Assessment: %+v\n", riskResult)
    
    // Start data discovery
    discoveryResult, err := c.StartDiscovery(context.Background(), client.DiscoveryInput{
        BusinessID:     "biz_12345",
        DiscoveryTypes: []string{"business_info", "compliance_data"},
        DataSources:    []string{"business_registry", "news_articles"},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Discovery: %+v\n", discoveryResult)
}
```

### Python SDK

#### 1. Installation

```bash
pip install kyb-platform-client
```

#### 2. Basic Usage

```python
from kyb_platform import Client

# Create client
client = Client(
    base_url="https://api.kyb-platform.com",
    api_key="your-api-key"
)

# Classify business
result = client.classify_business(
    business_name="Acme Corporation",
    website_url="https://www.acme.com",
    description="Technology solutions provider"
)

print(f"Classification: {result}")
```

#### 3. Advanced Usage

```python
from kyb_platform import Client
import asyncio

async def main():
    # Create async client
    client = Client(
        base_url="https://api.kyb-platform.com",
        api_key="your-api-key",
        timeout=30
    )
    
    # Perform multiple operations
    tasks = [
        client.classify_business(
            business_name="Acme Corporation",
            website_url="https://www.acme.com"
        ),
        client.assess_risk(
            business_id="biz_12345",
            assessment_type="comprehensive"
        ),
        client.start_discovery(
            business_id="biz_12345",
            discovery_types=["business_info", "compliance_data"]
        )
    ]
    
    results = await asyncio.gather(*tasks)
    
    for i, result in enumerate(results):
        print(f"Result {i+1}: {result}")

# Run async function
asyncio.run(main())
```

### JavaScript SDK

#### 1. Installation

```bash
npm install @kyb-platform/client
```

#### 2. Basic Usage

```javascript
import { Client } from '@kyb-platform/client';

// Create client
const client = new Client({
    baseURL: 'https://api.kyb-platform.com',
    apiKey: 'your-api-key'
});

// Classify business
const result = await client.classifyBusiness({
    businessName: 'Acme Corporation',
    websiteUrl: 'https://www.acme.com',
    description: 'Technology solutions provider'
});

console.log('Classification:', result);
```

#### 3. Advanced Usage

```javascript
import { Client } from '@kyb-platform/client';

async function main() {
    // Create client with custom configuration
    const client = new Client({
        baseURL: 'https://api.kyb-platform.com',
        apiKey: 'your-api-key',
        timeout: 30000,
        retries: 3,
        userAgent: 'MyApp/1.0'
    });
    
    try {
        // Perform multiple operations
        const [classification, riskAssessment, discovery] = await Promise.all([
            client.classifyBusiness({
                businessName: 'Acme Corporation',
                websiteUrl: 'https://www.acme.com'
            }),
            client.assessRisk({
                businessId: 'biz_12345',
                assessmentType: 'comprehensive'
            }),
            client.startDiscovery({
                businessId: 'biz_12345',
                discoveryTypes: ['business_info', 'compliance_data']
            })
        ]);
        
        console.log('Classification:', classification);
        console.log('Risk Assessment:', riskAssessment);
        console.log('Discovery:', discovery);
        
    } catch (error) {
        console.error('Error:', error.message);
    }
}

main();
```

## Testing

### Unit Testing

#### 1. Writing Unit Tests

```go
// internal/business/classification/service_test.go
package classification

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestService_ClassifyBusiness(t *testing.T) {
    tests := []struct {
        name           string
        input          ClassificationInput
        mockSetup      func(*mocks.Repository)
        expectedResult *ClassificationResult
        expectedError  string
    }{
        {
            name: "successful classification",
            input: ClassificationInput{
                BusinessName: "Acme Corporation",
                WebsiteURL:   "https://www.acme.com",
            },
            mockSetup: func(repo *mocks.Repository) {
                repo.On("Save", mock.Anything, mock.Anything).Return(nil)
            },
            expectedResult: &ClassificationResult{
                BusinessID: "biz_12345",
                Classification: Classification{
                    PrimaryIndustry: "Technology",
                    ConfidenceScore: 0.92,
                },
            },
        },
        {
            name: "validation error",
            input: ClassificationInput{
                BusinessName: "", // Invalid: empty name
            },
            expectedError: "business name is required",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockRepo := &mocks.Repository{}
            if tt.mockSetup != nil {
                tt.mockSetup(mockRepo)
            }
            
            // Create service
            service := NewService(mockRepo, nil, nil)
            
            // Execute test
            result, err := service.ClassifyBusiness(context.Background(), tt.input)
            
            // Assert results
            if tt.expectedError != "" {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.expectedError)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
                assert.Equal(t, tt.expectedResult.BusinessID, result.BusinessID)
            }
            
            // Verify mocks
            mockRepo.AssertExpectations(t)
        })
    }
}
```

#### 2. Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -run TestService_ClassifyBusiness ./internal/business/classification/

# Run tests with race detection
go test -race ./...
```

### Integration Testing

#### 1. Integration Test Setup

```go
// test/integration/classification_test.go
package integration

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/require"
    "github.com/your-org/kyb-platform/internal/api/handlers"
    "github.com/your-org/kyb-platform/internal/business/classification"
)

func TestClassificationIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Setup test environment
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create service and handler
    repo := postgres.NewClassificationRepository(db)
    service := classification.NewService(repo, nil, nil)
    handler := handlers.NewClassificationHandler(service, nil)
    
    // Test classification flow
    t.Run("complete classification flow", func(t *testing.T) {
        // Create test request
        req := createTestRequest(t, "POST", "/api/v3/classify", map[string]interface{}{
            "business_name": "Test Company",
            "website_url":   "https://www.testcompany.com",
        })
        
        // Execute request
        w := httptest.NewRecorder()
        handler.ClassifyBusiness(w, req)
        
        // Assert response
        require.Equal(t, http.StatusOK, w.Code)
        
        var response map[string]interface{}
        err := json.Unmarshal(w.Body.Bytes(), &response)
        require.NoError(t, err)
        
        assert.NotEmpty(t, response["business_id"])
        assert.NotEmpty(t, response["classification"])
    })
}
```

#### 2. Running Integration Tests

```bash
# Run integration tests
go test -tags=integration ./test/integration/

# Run integration tests with database
TEST_DB_URL="postgres://user:pass@localhost:5432/kyb_platform_test" go test -tags=integration ./test/integration/

# Run integration tests with coverage
go test -tags=integration -cover ./test/integration/
```

### Performance Testing

#### 1. Benchmark Tests

```go
// internal/business/classification/service_bench_test.go
package classification

import (
    "context"
    "testing"
)

func BenchmarkService_ClassifyBusiness(b *testing.B) {
    service := setupBenchmarkService(b)
    
    input := ClassificationInput{
        BusinessName: "Acme Corporation",
        WebsiteURL:   "https://www.acme.com",
        Description:  "Technology solutions provider",
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.ClassifyBusiness(context.Background(), input)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkService_ClassifyBusiness_Parallel(b *testing.B) {
    service := setupBenchmarkService(b)
    
    input := ClassificationInput{
        BusinessName: "Acme Corporation",
        WebsiteURL:   "https://www.acme.com",
        Description:  "Technology solutions provider",
    }
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _, err := service.ClassifyBusiness(context.Background(), input)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}
```

#### 2. Running Benchmarks

```bash
# Run all benchmarks
go test -bench=. ./...

# Run specific benchmark
go test -bench=BenchmarkService_ClassifyBusiness ./internal/business/classification/

# Run benchmarks with memory profiling
go test -bench=. -benchmem ./...

# Run benchmarks with CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./...
```

## Contributing

### Development Workflow

#### 1. Branch Strategy

```bash
# Create feature branch
git checkout -b feature/your-feature-name

# Create bugfix branch
git checkout -b bugfix/your-bug-description

# Create hotfix branch
git checkout -b hotfix/urgent-fix
```

#### 2. Code Standards

```go
// Follow Go coding standards
package main

import (
    "context"
    "fmt"
    "time"
    
    "github.com/your-org/kyb-platform/internal/business/classification"
)

// Service provides business classification functionality
type Service struct {
    repo   classification.Repository
    cache  cache.Cache
    logger *zap.Logger
}

// NewService creates a new classification service
func NewService(repo classification.Repository, cache cache.Cache, logger *zap.Logger) *Service {
    return &Service{
        repo:   repo,
        cache:  cache,
        logger: logger,
    }
}

// ClassifyBusiness performs business classification
func (s *Service) ClassifyBusiness(ctx context.Context, input ClassificationInput) (*ClassificationResult, error) {
    // Implementation
}
```

#### 3. Commit Guidelines

```bash
# Use conventional commit format
git commit -m "feat: add new classification algorithm"
git commit -m "fix: resolve memory leak in cache"
git commit -m "docs: update API documentation"
git commit -m "test: add integration tests for risk assessment"
git commit -m "refactor: improve error handling in handlers"
```

#### 4. Pull Request Process

1. **Create feature branch**
2. **Make changes**
3. **Write tests**
4. **Update documentation**
5. **Run quality checks**
6. **Create pull request**
7. **Address review comments**
8. **Merge when approved**

### Code Review Guidelines

#### 1. Review Checklist

- [ ] Code follows Go conventions
- [ ] Tests are included and passing
- [ ] Documentation is updated
- [ ] Error handling is appropriate
- [ ] Performance considerations addressed
- [ ] Security implications considered
- [ ] No breaking changes (or documented)

#### 2. Review Comments

```go
// Good: Specific, actionable feedback
// Consider using a more descriptive variable name here
// instead of 'r', use 'result' or 'classification'

// Good: Suggesting improvements
// This could be optimized by using a connection pool
// to reduce database connection overhead

// Good: Security considerations
// Make sure to validate user input before processing
// to prevent potential injection attacks
```

## Debugging

### Debugging Tools

#### 1. Application Debugging

```go
// Enable debug mode
DEBUG_MODE=true

// Use debug logging
logger.Debug("Processing classification request", 
    zap.String("business_name", input.BusinessName),
    zap.String("user_id", userID),
)

// Add debug breakpoints
if DEBUG_MODE {
    debug.PrintStack()
}
```

#### 2. Profiling

```go
// CPU profiling
import "runtime/pprof"

func enableCPUProfiling() {
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()
}

// Memory profiling
import "runtime/pprof"

func enableMemoryProfiling() {
    f, err := os.Create("memory.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()
    
    pprof.WriteHeapProfile(f)
}
```

#### 3. Tracing

```go
// Enable tracing
import "go.opentelemetry.io/otel"

func enableTracing() {
    tracer := otel.Tracer("kyb-platform")
    
    ctx, span := tracer.Start(context.Background(), "classify_business")
    defer span.End()
    
    // Add attributes to span
    span.SetAttributes(
        attribute.String("business_name", input.BusinessName),
        attribute.String("user_id", userID),
    )
}
```

### Common Debugging Scenarios

#### 1. Database Issues

```go
// Enable SQL query logging
db.LogMode(true)

// Check connection pool
stats := db.Stats()
log.Printf("DB Stats: %+v", stats)

// Monitor slow queries
db.SetLogger(gorm.Logger{
    LogLevel: gorm.LogLevel(logger.Info),
    Colorful: true,
})
```

#### 2. Cache Issues

```go
// Check cache hit rate
cacheStats := cache.Stats()
log.Printf("Cache Hit Rate: %.2f%%", cacheStats.HitRate()*100)

// Monitor cache size
log.Printf("Cache Size: %d items", cacheStats.Size())

// Debug cache operations
cache.Set("debug_key", "debug_value", time.Minute)
value, err := cache.Get("debug_key")
log.Printf("Cache Get Result: %v, %v", value, err)
```

#### 3. API Issues

```go
// Enable request/response logging
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Log request
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        
        // Call next handler
        next.ServeHTTP(w, r)
        
        // Log response time
        log.Printf("Response Time: %v", time.Since(start))
    })
}
```

## Performance Optimization

### Database Optimization

#### 1. Query Optimization

```go
// Use indexes effectively
CREATE INDEX idx_business_classifications_user_id ON business_classifications(user_id);
CREATE INDEX idx_business_classifications_created_at ON business_classifications(created_at);

// Use prepared statements
stmt, err := db.Prepare("SELECT * FROM business_classifications WHERE user_id = $1")
if err != nil {
    return err
}
defer stmt.Close()

// Use connection pooling
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

#### 2. Caching Strategy

```go
// Implement multi-level caching
type CacheService struct {
    l1Cache *sync.Map // In-memory cache
    l2Cache cache.Cache // Redis cache
}

func (c *CacheService) Get(key string) (interface{}, error) {
    // Check L1 cache first
    if value, ok := c.l1Cache.Load(key); ok {
        return value, nil
    }
    
    // Check L2 cache
    value, err := c.l2Cache.Get(key)
    if err == nil {
        // Store in L1 cache
        c.l1Cache.Store(key, value)
        return value, nil
    }
    
    return nil, err
}
```

### Application Optimization

#### 1. Goroutine Management

```go
// Use worker pools for heavy operations
type WorkerPool struct {
    workers int
    jobs    chan Job
    results chan Result
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    for job := range wp.jobs {
        result := processJob(job)
        wp.results <- result
    }
}
```

#### 2. Memory Optimization

```go
// Use object pools for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func processData(data []byte) {
    buffer := bufferPool.Get().([]byte)
    defer bufferPool.Put(buffer)
    
    // Use buffer for processing
    buffer = buffer[:0] // Reset buffer
    buffer = append(buffer, data...)
}
```

## Security Best Practices

### Input Validation

#### 1. Validation Functions

```go
// Validate business input
func validateBusinessInput(input BusinessInput) error {
    if input.Name == "" {
        return errors.New("business name is required")
    }
    
    if len(input.Name) > 255 {
        return errors.New("business name too long")
    }
    
    if input.WebsiteURL != "" {
        if !isValidURL(input.WebsiteURL) {
            return errors.New("invalid website URL")
        }
    }
    
    return nil
}

// Validate URL
func isValidURL(url string) bool {
    parsed, err := url.Parse(url)
    return err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https")
}
```

#### 2. SQL Injection Prevention

```go
// Use parameterized queries
func (r *Repository) GetByID(id string) (*Business, error) {
    query := "SELECT * FROM businesses WHERE id = $1"
    row := r.db.QueryRow(query, id)
    
    var business Business
    err := row.Scan(&business.ID, &business.Name, &business.CreatedAt)
    return &business, err
}

// Use ORM with parameterized queries
func (r *Repository) GetByID(id string) (*Business, error) {
    var business Business
    err := r.db.Where("id = ?", id).First(&business).Error
    return &business, err
}
```

### Authentication and Authorization

#### 1. JWT Token Validation

```go
// Validate JWT token
func validateJWT(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return []byte(jwtSecret), nil
    })
    
    if err != nil {
        return nil, err
    }
    
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    
    return nil, errors.New("invalid token")
}
```

#### 2. Role-Based Access Control

```go
// Check user permissions
func (h *Handler) requirePermission(permission string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            user := getUserFromContext(r.Context())
            
            if !user.HasPermission(permission) {
                http.Error(w, "insufficient permissions", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

## Deployment

### Development Deployment

#### 1. Local Development

```bash
# Start development environment
make dev-start

# Build and run locally
make build-dev
make run-dev

# Run with hot reload
make dev-watch
```

#### 2. Docker Development

```bash
# Build development image
docker build -f Dockerfile.dev -t kyb-platform:dev .

# Run development container
docker run -p 8080:8080 -v $(pwd):/app kyb-platform:dev

# Use docker-compose for development
docker-compose -f docker-compose.dev.yml up -d
```

### Production Deployment

#### 1. Build for Production

```bash
# Build production image
make build-prod

# Build multi-stage image
docker build -f Dockerfile -t kyb-platform:latest .

# Push to registry
docker tag kyb-platform:latest your-registry/kyb-platform:latest
docker push your-registry/kyb-platform:latest
```

#### 2. Deploy to Production

```bash
# Deploy with docker-compose
docker-compose -f docker-compose.prod.yml up -d

# Deploy to Kubernetes
kubectl apply -f deployments/kubernetes/

# Deploy to AWS ECS
aws ecs update-service --cluster kyb-platform --service kyb-platform-api --force-new-deployment
```

### Monitoring and Observability

#### 1. Health Checks

```go
// Implement health check endpoint
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
    health := HealthStatus{
        Status:    "healthy",
        Timestamp: time.Now(),
        Version:   "1.0.0",
        Uptime:    time.Since(startTime),
    }
    
    // Check dependencies
    if err := h.checkDatabase(); err != nil {
        health.Status = "unhealthy"
        health.Errors = append(health.Errors, "database: "+err.Error())
    }
    
    if err := h.checkCache(); err != nil {
        health.Status = "unhealthy"
        health.Errors = append(health.Errors, "cache: "+err.Error())
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

#### 2. Metrics Collection

```go
// Collect application metrics
func (h *Handler) collectMetrics() {
    // Request metrics
    prometheus.Register(requestCounter)
    prometheus.Register(requestDuration)
    
    // Business metrics
    prometheus.Register(classificationCounter)
    prometheus.Register(classificationDuration)
    
    // Error metrics
    prometheus.Register(errorCounter)
}
```

---

**Document Version**: 1.0.0  
**Last Updated**: December 19, 2024  
**Next Review**: March 19, 2025
