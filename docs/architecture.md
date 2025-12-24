# KYB Platform Architecture Documentation

**Version**: 1.0.0  
**Last Updated**: December 23, 2025  
**Status**: Production

---

## Table of Contents

1. [System Architecture Overview](#system-architecture-overview)
2. [Component Architecture](#component-architecture)
3. [Service-to-Service Communication](#service-to-service-communication)
4. [Database Access Patterns](#database-access-patterns)
5. [Authentication Flow](#authentication-flow)
6. [Deployment Configuration](#deployment-configuration)
7. [Environment Variables and Configuration](#environment-variables-and-configuration)
8. [External API Integrations](#external-api-integrations)
9. [Data Flow Diagrams](#data-flow-diagrams)
10. [Infrastructure Components](#infrastructure-components)

---

## System Architecture Overview

The KYB Platform is a microservices-based architecture built with Go, following Clean Architecture principles and Domain-Driven Design (DDD). The system is deployed on Railway with Supabase as the primary database and authentication provider.

### High-Level Architecture Diagram

```mermaid
graph TB
    subgraph "Client Layer"
        WebApp[Web Application]
        MobileApp[Mobile App]
        API_Client[API Clients]
    end
    
    subgraph "API Gateway Layer"
        API_GW[API Gateway Service]
        LoadBalancer[Railway Load Balancer]
    end
    
    subgraph "Application Services Layer"
        Classify[Classification Service]
        Risk[Risk Assessment Service]
        Merchant[Merchant Service]
        Frontend[Frontend Service]
        BI[Business Intelligence Gateway]
    end
    
    subgraph "ML/AI Services Layer"
        Embedding[Embedding Service<br/>Python]
        LLM[LLM Service<br/>Python]
        ML[Python ML Service]
    end
    
    subgraph "Data Layer"
        Supabase[(Supabase PostgreSQL)]
        Redis[(Redis Cache)]
    end
    
    subgraph "External Services"
        NewsAPI[News API]
        OpenCorp[OpenCorporates]
        GovData[Government Data APIs]
        ThomsonReuters[Thomson Reuters]
        OFAC[OFAC API]
        WorldCheck[World-Check]
    end
    
    subgraph "Infrastructure"
        Railway[Railway Platform]
        Monitoring[Monitoring & Logging]
    end
    
    WebApp --> LoadBalancer
    MobileApp --> LoadBalancer
    API_Client --> LoadBalancer
    LoadBalancer --> API_GW
    
    API_GW --> Classify
    API_GW --> Risk
    API_GW --> Merchant
    API_GW --> Frontend
    API_GW --> BI
    
    Classify --> Embedding
    Classify --> LLM
    Classify --> ML
    Risk --> ML
    
    Classify --> Supabase
    Risk --> Supabase
    Merchant --> Supabase
    BI --> Supabase
    
    Classify --> Redis
    Risk --> Redis
    Merchant --> Redis
    
    Risk --> NewsAPI
    Risk --> OpenCorp
    Risk --> GovData
    Risk --> ThomsonReuters
    Risk --> OFAC
    Risk --> WorldCheck
    
    API_GW --> Railway
    Classify --> Railway
    Risk --> Railway
    Merchant --> Railway
```

### Architecture Principles

1. **Clean Architecture**: Clear separation of concerns with handlers, services, repositories, and domain models
2. **Microservices**: Independent, scalable services with well-defined boundaries
3. **Domain-Driven Design**: Business logic organized by domain boundaries
4. **Interface-Driven Development**: Services communicate via interfaces for flexibility and testability
5. **Stateless Services**: Services are stateless to enable horizontal scaling
6. **API-First Design**: RESTful APIs with versioning (v1, v2, v3)

---

## Component Architecture

### Core Services

#### 1. API Gateway Service
- **Location**: `services/api-gateway/`
- **Purpose**: Single entry point for all client requests, routing, authentication, and rate limiting
- **Technology**: Go with Gorilla Mux
- **Port**: 8080
- **Key Features**:
  - Request routing to backend services
  - Authentication and authorization middleware
  - CORS handling
  - Rate limiting
  - Request/response logging
  - Health check aggregation

#### 2. Classification Service
- **Location**: `services/classification-service/`
- **Purpose**: Business classification using MCC, NAICS, and SIC codes
- **Technology**: Go
- **Port**: 8080
- **Key Features**:
  - Multi-strategy classification (keyword-based, embedding-based, LLM-based)
  - Intelligent routing to best classification method
  - Confidence scoring
  - Batch processing support
  - Caching for performance

#### 3. Risk Assessment Service
- **Location**: `services/risk-assessment-service/`
- **Purpose**: Comprehensive business risk assessment and scoring
- **Technology**: Go with ML models (ONNX, XGBoost)
- **Port**: 8080
- **Key Features**:
  - ML-based risk scoring
  - External API integration (Thomson Reuters, OFAC, World-Check)
  - Adverse media monitoring
  - Compliance checking
  - Risk factor analysis

#### 4. Merchant Service
- **Location**: `services/merchant-service/`
- **Purpose**: Merchant data management and operations
- **Technology**: Go
- **Port**: 8080
- **Key Features**:
  - Merchant CRUD operations
  - Merchant analytics
  - Website analysis
  - Risk score tracking
  - Batch operations

#### 5. Frontend Service
- **Location**: `services/frontend-service/`
- **Purpose**: Serves static frontend assets and HTML pages
- **Technology**: Go HTTP server
- **Port**: 8080
- **Key Features**:
  - Static file serving
  - HTML page rendering
  - Dashboard UI
  - Merchant hub interface

#### 6. Business Intelligence Gateway
- **Location**: `cmd/business-intelligence-gateway/`
- **Purpose**: Business intelligence and analytics aggregation
- **Technology**: Go
- **Port**: 8080
- **Key Features**:
  - Market analysis
  - Competitive analysis
  - Growth analytics
  - Business intelligence aggregation

### ML/AI Services

#### 1. Embedding Service
- **Location**: `services/embedding-service/`
- **Purpose**: Text embedding generation for semantic similarity
- **Technology**: Python (FastAPI/Uvicorn)
- **Port**: 8000
- **Key Features**:
  - Sentence transformer embeddings
  - Semantic similarity calculation
  - Batch embedding generation

#### 2. LLM Service
- **Location**: `services/llm-service/`
- **Purpose**: Large language model inference for classification
- **Technology**: Python (FastAPI/Uvicorn)
- **Port**: 8000
- **Key Features**:
  - LLM-based classification
  - Natural language understanding
  - Context-aware classification

#### 3. Python ML Service
- **Location**: `python_ml_service/`
- **Purpose**: Machine learning model inference
- **Technology**: Python
- **Port**: 8000
- **Key Features**:
  - ML model serving
  - Enhanced classification
  - Model versioning

### Supporting Services

#### 1. Service Discovery
- **Location**: `cmd/service-discovery/`
- **Purpose**: Service registration and discovery
- **Technology**: Go
- **Features**:
  - Service registry
  - Health monitoring
  - Service metadata

#### 2. Redis Cache
- **Location**: `services/redis-cache/`
- **Purpose**: Distributed caching layer
- **Technology**: Redis
- **Features**:
  - Classification result caching
  - Session storage
  - Rate limiting state

---

## Service-to-Service Communication

### Communication Patterns

#### 1. HTTP/REST (Primary)
All service-to-service communication uses HTTP/REST APIs with JSON payloads.

**Pattern**: Synchronous request-response
- **Protocol**: HTTP/1.1 and HTTP/2
- **Format**: JSON
- **Timeout**: 30-120 seconds (configurable per service)
- **Retry**: Exponential backoff with circuit breaker

**Example Flow**:
```
API Gateway → Classification Service
  POST /api/v1/classify
  Headers: Authorization, Content-Type
  Body: { "business_name": "...", "description": "..." }
  Response: { "classifications": {...}, "confidence": 0.95 }
```

#### 2. Service Discovery
Services register themselves with the service discovery service for dynamic routing.

**Registration Flow**:
```go
Service → Service Discovery
  POST /services/register
  {
    "id": "classification-service-1",
    "name": "classification-service",
    "url": "http://classification-service:8080",
    "version": "1.0.0",
    "health_url": "/health"
  }
```

**Discovery Flow**:
```go
API Gateway → Service Discovery
  GET /services/discover?name=classification-service
  Response: [
    {
      "id": "classification-service-1",
      "url": "http://classification-service:8080",
      "status": "healthy"
    }
  ]
```

#### 3. Circuit Breaker Pattern
All external service calls use circuit breakers to prevent cascading failures.

**Implementation**:
- **Failure Threshold**: 5 consecutive failures
- **Timeout**: 30 seconds
- **Half-Open State**: After 60 seconds
- **State Management**: In-memory with Redis for distributed state

#### 4. Load Balancing
Railway provides built-in load balancing. Services can be scaled horizontally.

**Configuration**:
- **Min Instances**: 1
- **Max Instances**: 10 (configurable)
- **Health Check**: `/health` endpoint
- **Load Balancing Algorithm**: Round-robin

### Service Communication Diagram

```mermaid
sequenceDiagram
    participant Client
    participant API_GW as API Gateway
    participant Classify as Classification Service
    participant Embedding as Embedding Service
    participant Supabase
    
    Client->>API_GW: POST /api/v1/classify
    API_GW->>API_GW: Authenticate & Rate Limit
    API_GW->>Classify: POST /classify
    Classify->>Supabase: Query classification codes
    Supabase-->>Classify: Return codes
    Classify->>Embedding: POST /embed
    Embedding-->>Classify: Return embeddings
    Classify->>Classify: Classify business
    Classify-->>API_GW: Return classification
    API_GW-->>Client: Return response
```

### Service URLs (Railway)

Services communicate using Railway-generated URLs:

- **API Gateway**: `https://api-gateway-production-xxx.up.railway.app`
- **Classification Service**: `https://classification-service-production-xxx.up.railway.app`
- **Risk Assessment Service**: `https://risk-assessment-service-production-xxx.up.railway.app`
- **Merchant Service**: `https://merchant-service-production-xxx.up.railway.app`
- **Embedding Service**: `https://embedding-service-production-xxx.up.railway.app`
- **LLM Service**: `https://llm-service-production-xxx.up.railway.app`

**Internal Communication**: Services can also use internal Railway service names for faster communication within the same project.

---

## Database Access Patterns

### Supabase Integration

The platform uses Supabase as the primary database and authentication provider.

#### Connection Pattern

**HTTP/REST API Access** (Primary):
- Supabase uses PostgREST, which exposes PostgreSQL via HTTP/REST
- All database operations go through Supabase REST API
- Connection pooling is handled by Go's HTTP client

**Client Implementation**:
```go
// Location: internal/database/supabase_client.go
type SupabaseClient struct {
    client     *supa.Client          // Supabase Go client
    postgrest  *postgrest.Client     // PostgREST client for direct queries
    url        string
    apiKey     string
    serviceKey string
}
```

#### Database Access Methods

1. **Supabase Go Client** (High-level operations):
   - Authentication
   - Storage operations
   - Realtime subscriptions

2. **PostgREST Client** (Direct queries):
   - Custom SQL queries
   - Complex joins
   - Aggregations

3. **Connection Pooling**:
   - HTTP client connection pooling (Go's default transport)
   - Max idle connections: 64 per host
   - Idle connection timeout: 90 seconds
   - TLS handshake timeout: 10 seconds

#### Database Schema

**Core Tables**:
- `merchants` - Merchant/business information
- `classification_codes` - MCC, NAICS, SIC code metadata
- `business_classifications` - Classification results
- `business_risk_assessments` - Risk assessment results
- `users` - User accounts (managed by Supabase Auth)
- `api_keys` - API key management
- `sessions` - User sessions

**Access Pattern Example**:
```go
// Query classification codes
result, _, err := client.postgrest.
    From("classification_codes").
    Select("code,description,type", "", false).
    Eq("type", "MCC").
    Limit(10, "").
    Execute()
```

#### Row-Level Security (RLS)

Supabase RLS policies enforce data access control:
- Users can only access their own data
- Service role key bypasses RLS for admin operations
- API keys have restricted access based on tenant

#### Database Connection Configuration

**Environment Variables**:
```bash
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_API_KEY=anon_key
SUPABASE_SERVICE_ROLE_KEY=service_role_key
SUPABASE_JWT_SECRET=jwt_secret
```

**Connection Settings**:
- **Timeout**: 10 seconds for health checks, 30 seconds for queries
- **Retry**: 3 attempts with exponential backoff
- **SSL**: Required (enforced by Supabase)

---

## Authentication Flow

### Authentication Methods

The platform supports multiple authentication methods:

#### 1. JWT Token Authentication (Primary)

**Flow**:
```mermaid
sequenceDiagram
    participant Client
    participant API_GW as API Gateway
    participant Supabase
    participant Service
    
    Client->>Supabase: POST /auth/v1/token
    Supabase-->>Client: JWT Token
    Client->>API_GW: Request with Bearer Token
    API_GW->>Supabase: Validate Token
    Supabase-->>API_GW: Token Valid + User Info
    API_GW->>Service: Forward Request with User Context
    Service-->>API_GW: Response
    API_GW-->>Client: Response
```

**Token Structure**:
```json
{
  "user_id": "uuid",
    "email": "user@example.com",
  "role": "user",
    "iat": 1640995200,
  "exp": 1640996100,
  "iss": "supabase",
  "sub": "uuid"
}
```

**Token Types**:
- **Access Token**: Short-lived (15 minutes - 24 hours, configurable)
- **Refresh Token**: Long-lived (7 days, configurable)

#### 2. API Key Authentication

**Flow**:
```
Client → API Gateway
  Header: X-API-Key: <api_key>
  OR
  Header: Authorization: ApiKey <api_key>
```

**Validation**:
- API keys stored in Supabase `api_keys` table
- Keys have associated permissions and rate limits
- Keys can be scoped to specific tenants

#### 3. Public Endpoints

Some endpoints are publicly accessible without authentication:
- `/health` - Health checks
- `/v1/classify` - Public classification (rate-limited)
- `/` - Root endpoint

### Authentication Middleware

**Location**: `internal/api/middleware/auth.go`

**Middleware Chain**:
1. **CORS Middleware** - Handle preflight requests
2. **Security Headers** - Add security headers
3. **Logging Middleware** - Log requests
4. **Rate Limiting** - Enforce rate limits
5. **Authentication** - Validate tokens/API keys
6. **Authorization** - Check permissions

**Implementation**:
```go
// RequireAuth middleware
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Validate token with Supabase
        user, err := m.authService.ValidateToken(ctx, tokenString)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        // Add user context
        ctx := context.WithValue(r.Context(), "user_id", user.ID)
        ctx = context.WithValue(ctx, "user_email", user.Email)
        ctx = context.WithValue(ctx, "user_role", user.Role)
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### Authorization

**Role-Based Access Control (RBAC)**:
- **Admin**: Full access to all resources
- **User**: Access to own resources
- **Service**: Internal service-to-service communication

**Permission Checking**:
```go
// Check if user has permission
if user.Role != "admin" && user.ID != resourceOwnerID {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
}
```

---

## Deployment Configuration

### Railway Platform

The platform is deployed on Railway, a modern PaaS platform.

#### Railway Configuration

**File**: `railway.json`

**Services Configuration**:
```json
{
  "services": [
    {
      "name": "classification-service",
      "source": ".",
      "build": {
        "builder": "DOCKERFILE",
        "dockerfilePath": "services/classification-service/Dockerfile"
      },
      "deploy": {
        "startCommand": "./classification-service",
        "healthcheckPath": "/health",
        "healthcheckTimeout": 30,
        "restartPolicyType": "ON_FAILURE",
        "restartPolicyMaxRetries": 10
      }
    }
  ]
}
```

#### Railway Platform Settings

**Health Check Configuration**:
- **Path**: `/health`
- **Interval**: 30 seconds
- **Timeout**: 10 seconds
- **Start Period**: 60 seconds (for cold start)
- **Retries**: 3

**Resource Limits**:
- **Memory**: 512MB - 1GB per service
- **CPU**: 0.5 - 1 vCPU per service
- **Disk**: 1GB (default)

**Scaling Configuration**:
- **Min Instances**: 1 (keeps service warm)
- **Max Instances**: 10 (horizontal scaling)
- **Auto-scaling**: Enabled based on CPU/memory usage

**Always On Setting**:
- **Production**: Enabled (prevents cold starts)
- **Development**: Disabled (cost optimization)

#### Railway Environment Variables

Railway automatically provides:
- `PORT` - Service port (usually 8080)
- `RAILWAY_ENVIRONMENT` - Environment name (production, staging)
- `RAILWAY_PROJECT_ID` - Project identifier

**Service URLs**:
Railway generates unique URLs for each service:
- Format: `https://<service-name>-<environment>-<id>.up.railway.app`
- Internal communication: Use service names within Railway project

### Supabase Configuration

#### Supabase Project Setup

**Database**:
- **Provider**: Supabase (PostgreSQL)
- **Version**: PostgreSQL 15
- **Region**: Configurable (US, EU, etc.)
- **Backups**: Automatic daily backups

**Authentication**:
- **Provider**: Supabase Auth
- **Methods**: Email/Password, OAuth (configurable)
- **JWT Secret**: Managed by Supabase
- **Session Management**: Supabase handles sessions

**Storage**:
- **Provider**: Supabase Storage
- **Buckets**: Configurable for file storage
- **CDN**: Automatic CDN for static assets

#### Supabase Environment Variables

**Required Variables**:
```bash
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_API_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  # anon key
SUPABASE_SERVICE_ROLE_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...  # service role key
SUPABASE_JWT_SECRET=your-jwt-secret
```

**Database Connection**:
- Supabase provides connection string via `DATABASE_URL` (if using direct PostgreSQL connection)
- Primary access is via HTTP/REST API (PostgREST)

### Docker Configuration

**Dockerfile Structure**:
```dockerfile
# Multi-stage build
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/service .
CMD ["./service"]
```

**Build Process**:
1. Railway detects Dockerfile
2. Builds Docker image
3. Pushes to Railway registry
4. Deploys container
5. Runs health checks
6. Routes traffic

---

## Environment Variables and Configuration

### Configuration Management

The platform uses environment variables for configuration, loaded through `internal/config/config.go`.

### Core Environment Variables

#### Application Configuration
```bash
# Environment
ENV=production                    # development, staging, production
PORT=8080                         # Service port
HOST=0.0.0.0                      # Bind address

# Timeouts
READ_TIMEOUT=30s                  # HTTP read timeout
WRITE_TIMEOUT=30s                 # HTTP write timeout
IDLE_TIMEOUT=60s                  # HTTP idle timeout
```

#### Supabase Configuration
```bash
# Supabase Connection
SUPABASE_URL=https://xxx.supabase.co
SUPABASE_API_KEY=anon_key
SUPABASE_SERVICE_ROLE_KEY=service_role_key
SUPABASE_JWT_SECRET=jwt_secret

# Database (if using direct connection)
DATABASE_URL=postgresql://user:pass@host:5432/db
DB_DRIVER=postgres
DB_HOST=db.xxx.supabase.co
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=password
DB_DATABASE=postgres
DB_SSL_MODE=require
```

#### Authentication Configuration
```bash
# JWT Configuration
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h
REFRESH_EXPIRATION=168h

# Password Policy
MIN_PASSWORD_LENGTH=8
REQUIRE_UPPERCASE=true
REQUIRE_LOWERCASE=true
REQUIRE_NUMBERS=true
REQUIRE_SPECIAL=true
MAX_LOGIN_ATTEMPTS=5
LOCKOUT_DURATION=15m
```

#### CORS Configuration
```bash
CORS_ALLOWED_ORIGINS=*
CORS_ALLOWED_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_ALLOWED_HEADERS=*
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=86400
```

#### Rate Limiting
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER=100
RATE_LIMIT_WINDOW_SIZE=60
RATE_LIMIT_BURST_SIZE=200
```

#### Observability
```bash
LOG_LEVEL=info                    # debug, info, warn, error
LOG_FORMAT=json                   # json, text
METRICS_ENABLED=true
METRICS_PORT=9090
METRICS_PATH=/metrics
TRACING_ENABLED=true
TRACING_URL=http://localhost:14268/api/traces
HEALTH_CHECK_PATH=/health
```

#### Service URLs (Railway)
```bash
# Classification Service
CLASSIFICATION_SERVICE_URL=https://classification-service-production-xxx.up.railway.app

# Risk Assessment Service
RISK_ASSESSMENT_SERVICE_URL=https://risk-assessment-service-production-xxx.up.railway.app

# Merchant Service
MERCHANT_SERVICE_URL=https://merchant-service-production-xxx.up.railway.app

# ML Services
PYTHON_ML_SERVICE_URL=https://python-ml-service-production-xxx.up.railway.app
EMBEDDING_SERVICE_URL=https://embedding-service-production-xxx.up.railway.app
LLM_SERVICE_URL=https://llm-service-production-xxx.up.railway.app
HREQUESTS_SERVICE_URL=https://hrequests-scraper-production-xxx.up.railway.app
PLAYWRIGHT_SERVICE_URL=https://playwright-scraper-production-xxx.up.railway.app
```

#### Feature Flags
```bash
FEATURE_BUSINESS_CLASSIFICATION=true
FEATURE_RISK_ASSESSMENT=true
FEATURE_COMPLIANCE_FRAMEWORK=true
FEATURE_ADVANCED_ANALYTICS=false
FEATURE_REAL_TIME_MONITORING=false
```

#### Cache Configuration
```bash
CLASSIFICATION_CACHE_ENABLED=true
CLASSIFICATION_CACHE_TTL=10m
CLASSIFICATION_CACHE_MAX_ENTRIES=10000
CLASSIFICATION_CACHE_JANITOR_INTERVAL=1m
```

#### Security Configuration
```bash
IP_BLOCK_ENABLED=true
IP_BLOCK_THRESHOLD=20
IP_BLOCK_WINDOW=5m
IP_BLOCK_DURATION=30m
VALIDATION_ENABLED=true
VALIDATION_MAX_BODY_SIZE=10485760  # 10MB
```

### Configuration Loading

**Process**:
1. Environment variables are read at startup
2. Configuration struct is populated
3. Validation is performed
4. Default values are applied if missing
5. Configuration is logged (sensitive values masked)

**Example**:
```go
// Load configuration
cfg, err := config.Load()
if err != nil {
    log.Fatal("Failed to load configuration", zap.Error(err))
}

// Validate configuration
if err := cfg.Validate(); err != nil {
    log.Fatal("Invalid configuration", zap.Error(err))
}
```

---

## External API Integrations

### Risk Assessment External APIs

#### 1. News API
- **Purpose**: Adverse media monitoring and sentiment analysis
- **Endpoint**: `GET /news/search`, `GET /news/adverse`
- **Configuration**:
  ```bash
  NEWS_API_KEY=your-api-key
  NEWS_API_BASE_URL=https://newsapi.org/v2
  NEWS_API_TIMEOUT=30s
  NEWS_API_RATE_LIMIT=100  # requests per minute
  ```
- **Features**:
  - Real-time news monitoring
  - Sentiment analysis
  - Industry-specific filtering
  - Relevance scoring

#### 2. OpenCorporates API
- **Purpose**: Company registration and compliance data
- **Endpoint**: `GET /companies/search`, `GET /companies/{id}`
- **Configuration**:
  ```bash
  OPENCORPORATES_API_KEY=your-api-key
  OPENCORPORATES_BASE_URL=https://api.opencorporates.com/v0.4
  OPENCORPORATES_TIMEOUT=30s
  ```
- **Features**:
  - Company registration details
  - Officer information
  - Filing history
  - Compliance status

#### 3. Government Data APIs
- **Purpose**: Sanctions and compliance checking
- **Endpoint**: `POST /sanctions/check`
- **Configuration**:
  ```bash
  GOVERNMENT_API_KEY=your-api-key
  GOVERNMENT_API_BASE_URL=https://api.government-data.gov
  GOVERNMENT_API_TIMEOUT=30s
  ```
- **Features**:
  - Sanctions list checking
  - PEP (Politically Exposed Persons) screening
  - Compliance verification

#### 4. Thomson Reuters API (Premium)
- **Purpose**: Comprehensive business intelligence
- **Endpoint**: `POST /thomson-reuters/search`
- **Configuration**:
  ```bash
  THOMSON_REUTERS_API_KEY=your-api-key
  THOMSON_REUTERS_BASE_URL=https://api.thomsonreuters.com
  THOMSON_REUTERS_TIMEOUT=60s
  THOMSON_REUTERS_RATE_LIMIT=60  # requests per minute
  ```
- **Features**:
  - Business intelligence
  - Risk indicators
  - Compliance data
  - Financial information

#### 5. OFAC API (Premium)
- **Purpose**: Office of Foreign Assets Control sanctions checking
- **Endpoint**: `POST /ofac/check`
- **Configuration**:
  ```bash
  OFAC_API_KEY=your-api-key
  OFAC_BASE_URL=https://api.ofac.treasury.gov
  OFAC_TIMEOUT=30s
  ```
- **Features**:
  - Sanctions list checking
  - SDN (Specially Designated Nationals) screening
  - Real-time updates

#### 6. World-Check API (Premium)
- **Purpose**: Enhanced due diligence and risk screening
- **Endpoint**: `POST /worldcheck/search`
- **Configuration**:
  ```bash
  WORLDCHECK_API_KEY=your-api-key
  WORLDCHECK_BASE_URL=https://api.worldcheck.com
  WORLDCHECK_TIMEOUT=30s
  ```
- **Features**:
  - Enhanced due diligence
  - Risk screening
  - PEP screening
  - Adverse media monitoring

### External API Integration Pattern

**Implementation**:
```go
// Location: services/risk-assessment-service/internal/external/api_manager.go
type ExternalAPIManager struct {
    thomsonReuters *thomson_reuters.ThomsonReutersClient
    ofac           *ofac.OFACClient
    worldCheck     *worldcheck.WorldCheckClient
    newsAPI        *newsapi.NewsAPIClient
    logger         *zap.Logger
}

// Parallel API calls with timeout and circuit breaker
func (eam *ExternalAPIManager) GetComprehensiveData(ctx context.Context, business *models.RiskAssessmentRequest) (*PremiumExternalDataResult, error) {
    ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
    defer cancel()
    
    // Parallel API calls
    var wg sync.WaitGroup
    results := make(chan apiResult, 3)
    
    // Call all APIs in parallel
    wg.Add(1)
    go func() {
        defer wg.Done()
        data, err := eam.thomsonReuters.GetComprehensiveData(ctx, business)
        results <- apiResult{source: "thomson_reuters", data: data, err: err}
    }()
    
    // ... more API calls
    
    wg.Wait()
    close(results)
    
    // Aggregate results
    return aggregateResults(results)
}
```

**Resilience Features**:
- **Circuit Breaker**: Prevents cascading failures
- **Retry Logic**: Exponential backoff (3 retries)
- **Timeout**: 30-60 seconds per API
- **Parallel Execution**: All APIs called concurrently
- **Graceful Degradation**: System continues if some APIs fail

### Web Scraping Services

#### 1. HRequests Scraper
- **Purpose**: Fast, lightweight web scraping
- **Technology**: Python (hrequests library)
- **Endpoint**: `POST /scrape`
- **Configuration**:
  ```bash
  HREQUESTS_SERVICE_URL=http://hrequests-scraper:8080
  ```

#### 2. Playwright Scraper
- **Purpose**: JavaScript-heavy site scraping
- **Technology**: Node.js (Playwright)
- **Endpoint**: `POST /scrape`
- **Configuration**:
  ```bash
  PLAYWRIGHT_SERVICE_URL=http://playwright-scraper:3000
  ```

---

## Data Flow Diagrams

### Classification Request Flow

```mermaid
sequenceDiagram
    participant Client
    participant API_GW as API Gateway
    participant Classify as Classification Service
    participant Embedding as Embedding Service
    participant LLM as LLM Service
    participant Supabase
    participant Redis
    
    Client->>API_GW: POST /api/v1/classify
    API_GW->>API_GW: Authenticate & Rate Limit
    API_GW->>Classify: POST /classify
    
    Classify->>Redis: Check cache
    alt Cache hit
        Redis-->>Classify: Return cached result
        Classify-->>API_GW: Return classification
    else Cache miss
        Classify->>Supabase: Query classification codes
        Supabase-->>Classify: Return codes
        
        Classify->>Embedding: POST /embed
        Embedding-->>Classify: Return embeddings
        
        Classify->>LLM: POST /classify (if needed)
        LLM-->>Classify: Return LLM classification
        
        Classify->>Classify: Combine results & score
        Classify->>Redis: Cache result
        Classify->>Supabase: Save classification
        Classify-->>API_GW: Return classification
    end
    
    API_GW-->>Client: Return response
```

### Risk Assessment Flow

```mermaid
sequenceDiagram
    participant Client
    participant API_GW as API Gateway
    participant Risk as Risk Assessment Service
    participant External as External APIs
    participant ML as ML Models
    participant Supabase
    
    Client->>API_GW: POST /api/v1/risk
    API_GW->>Risk: POST /assess
    
    par Parallel External API Calls
        Risk->>External: News API
        Risk->>External: OpenCorporates
        Risk->>External: Government APIs
        Risk->>External: Thomson Reuters
        Risk->>External: OFAC
        Risk->>External: World-Check
    end
    
    External-->>Risk: Return data
    
    Risk->>ML: Assess risk with ML models
    ML-->>Risk: Return risk score
    
    Risk->>Risk: Aggregate & calculate final score
    Risk->>Supabase: Save assessment
    Risk-->>API_GW: Return assessment
    API_GW-->>Client: Return response
```

---

## Infrastructure Components

### Monitoring and Observability

#### 1. Logging
- **Format**: JSON structured logging
- **Levels**: Debug, Info, Warn, Error
- **Library**: zap (Uber's structured logger)
- **Output**: stdout (Railway captures logs)

#### 2. Metrics
- **Provider**: Prometheus
- **Endpoint**: `/metrics`
- **Metrics**:
  - Request count
  - Response time (p50, p95, p99)
  - Error rate
  - Active connections
  - Cache hit rate

#### 3. Tracing
- **Provider**: OpenTelemetry
- **Export**: Jaeger (configurable)
- **Spans**: All service boundaries traced
- **Context Propagation**: Trace IDs propagated across services

#### 4. Health Checks
- **Endpoint**: `/health`
- **Checks**:
  - Database connectivity
  - External API availability
  - Cache connectivity
  - Service dependencies

### Caching Strategy

#### Redis Cache
- **Purpose**: Classification results, session data, rate limiting
- **TTL**: 10 minutes (configurable)
- **Eviction**: LRU (Least Recently Used)
- **Connection**: Pooled connections (64 per host)

#### In-Memory Cache
- **Purpose**: Frequently accessed data (classification codes)
- **Implementation**: Go sync.Map or custom cache
- **TTL**: 5 minutes

### Security

#### Network Security
- **TLS**: All external communication uses HTTPS
- **CORS**: Configured per environment
- **Rate Limiting**: Per IP and per API key
- **IP Blocking**: Automatic blocking of abusive IPs

#### Data Security
- **Encryption**: Data encrypted at rest (Supabase)
- **Transit**: TLS 1.2+ for all communications
- **Secrets**: Environment variables (Railway secrets)
- **RLS**: Row-level security in Supabase

### Backup and Recovery

#### Database Backups
- **Provider**: Supabase automatic backups
- **Frequency**: Daily
- **Retention**: 7 days (configurable)
- **Recovery**: Point-in-time recovery available

#### Application Backups
- **Code**: Git repository (version control)
- **Configuration**: Railway environment variables
- **Deployments**: Railway deployment history

---

## Appendix

### Service Ports

| Service | Port | Protocol |
|---------|------|----------|
| API Gateway | 8080 | HTTP |
| Classification Service | 8080 | HTTP |
| Risk Assessment Service | 8080 | HTTP |
| Merchant Service | 8080 | HTTP |
| Frontend Service | 8080 | HTTP |
| Embedding Service | 8000 | HTTP |
| LLM Service | 8000 | HTTP |
| Python ML Service | 8000 | HTTP |
| Redis | 6379 | TCP |

### API Endpoints Summary

#### Public Endpoints
- `GET /health` - Health check
- `POST /v1/classify` - Business classification

#### Authenticated Endpoints
- `POST /api/v1/classify` - Enhanced classification
- `GET /api/v1/merchants` - List merchants
- `POST /api/v1/merchants` - Create merchant
- `GET /api/v1/merchants/{id}` - Get merchant
- `POST /api/v1/risk` - Risk assessment
- `GET /api/v3/dashboard/metrics` - Dashboard metrics

### Technology Stack

#### Backend
- **Language**: Go 1.22+
- **Framework**: Standard library net/http, Gorilla Mux
- **Database**: Supabase (PostgreSQL)
- **Cache**: Redis
- **Authentication**: Supabase Auth (JWT)

#### ML/AI Services
- **Language**: Python 3.12+
- **Framework**: FastAPI, Uvicorn
- **Libraries**: Transformers, Sentence Transformers, ONNX Runtime

#### Infrastructure
- **Platform**: Railway
- **Database**: Supabase
- **Monitoring**: Prometheus, OpenTelemetry
- **Logging**: Structured JSON logs

---

**Document Version**: 1.0.0  
**Last Updated**: December 23, 2025  
**Maintained By**: KYB Platform Team
