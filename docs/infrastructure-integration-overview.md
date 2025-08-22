# Infrastructure Integration Overview

## 🏗️ **Complete Infrastructure Stack**

This document provides a comprehensive overview of how Docker, Railway, and Supabase work together to create a robust, scalable, and cost-effective infrastructure for the KYB Platform.

## 🐳 **Docker Containerization Strategy**

### **Multi-Stage Dockerfiles**

The project uses multiple Dockerfiles optimized for different deployment scenarios:

#### **1. Production Dockerfile (`Dockerfile`)**
```dockerfile
# Multi-stage build with security best practices
FROM golang:1.24-alpine AS builder
# Build stage with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o kyb-platform ./cmd/api

FROM alpine:3.19 AS production
# Non-root user for security
RUN addgroup -g 1001 -S kyb && adduser -u 1001 -S kyb -G kyb
USER kyb
# Health checks and monitoring
HEALTHCHECK --interval=30s --timeout=10s CMD curl -f http://localhost:8080/health
```

#### **2. Railway Deployment (`Dockerfile.beta`)**
```dockerfile
# Optimized for Railway deployment
FROM golang:1.22-alpine AS builder
# Railway-specific optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kyb-platform ./cmd/api

FROM alpine:latest
# Railway startup script integration
COPY --from=builder /app/kyb-platform .
COPY scripts/railway-startup.sh ./railway-startup.sh
CMD ["./railway-startup.sh"]
```

#### **3. Enhanced Features (`Dockerfile.enhanced`)**
```dockerfile
# For enhanced business intelligence features
FROM golang:1.24-alpine AS builder
# Enhanced build with additional dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -o kyb-platform ./cmd/api/main-enhanced.go
```

### **Docker Compose Development Environment**

```yaml
# docker-compose.yml - Local development
version: "3.8"
services:
  kyb-platform:
    build: .
    ports: ["8080:8080"]
    environment:
      - PROVIDER_DATABASE=postgres
      - DB_HOST=postgres
      - DB_PORT=5432
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=kyb_platform
      - POSTGRES_USER=kyb_user
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
```

## 🚀 **Railway Deployment Pipeline**

### **Railway Configuration**

```json
// railway.json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile.beta"
  },
  "deploy": {
    "startCommand": "./kyb-platform",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 300,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

### **Environment Variables Management**

```bash
# Railway automatically manages these environment variables
DATABASE_URL=postgresql://user:pass@host:port/db
CORS_ORIGIN=https://your-app.railway.app

# Supabase integration variables
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key

# Module-specific variables
MODULE_HEALTH_CHECK_INTERVAL=30s
MODULE_AUTO_RESTART=true
```

### **Railway Startup Script**

```bash
#!/bin/bash
# scripts/railway-startup.sh

# Wait for database to be ready
until pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USERNAME; do
  echo "Waiting for database..."
  sleep 2
done

# Run database migrations
./kyb-platform migrate

# Start the application
exec ./kyb-platform
```

## 🔗 **Supabase Integration**

### **Database Integration**

```go
// internal/database/supabase.go
type SupabaseClient struct {
    client *supa.Client
    url    string
    key    string
    logger *observability.Logger
}

func (s *SupabaseClient) Connect(ctx context.Context) error {
    // Connect to Supabase PostgreSQL
    _, err := s.client.DB.From("business_classifications").Select("count", false).Execute("")
    return err
}
```

### **Authentication Integration**

```go
// internal/auth/supabase_auth.go
type SupabaseAuthService struct {
    client *supa.Client
    config *config.SupabaseConfig
}

func (s *SupabaseAuthService) SignUp(ctx context.Context, email, password string) (*User, error) {
    // Use Supabase Auth
    result, err := s.client.Auth.SignUp(ctx, supa.UserCredentials{
        Email:    email,
        Password: password,
    })
    return &User{ID: result.User.ID, Email: result.User.Email}, err
}
```

### **Real-time Features**

```go
// internal/realtime/supabase_realtime.go
type SupabaseRealtimeModule struct {
    client *supa.Client
    config *config.SupabaseConfig
}

func (s *SupabaseRealtimeModule) SubscribeToChanges(channel string) error {
    // Subscribe to Supabase real-time changes
    return s.client.Realtime.Channel(channel).Subscribe()
}
```

## 🔄 **Provider-Agnostic Architecture**

### **Factory Pattern Implementation**

```go
// internal/factory.go
func NewDatabase(cfg *config.Config, logger *observability.Logger) (database.Database, error) {
    switch cfg.Provider.Database {
    case "supabase":
        return database.NewSupabaseDB(cfg.Supabase), nil
    case "aws":
        return database.NewAWSPostgresDB(cfg.AWS), nil
    case "gcp":
        return database.NewGCPPostgresDB(cfg.GCP), nil
    default:
        return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider.Database)
    }
}
```

### **Configuration Management**

```go
// internal/config/config.go
type Config struct {
    Provider ProviderConfig `json:"provider"`
    Supabase SupabaseConfig `json:"supabase"`
    AWS      AWSConfig      `json:"aws"`
    GCP      GCPConfig      `json:"gcp"`
}

type ProviderConfig struct {
    Database string `json:"database"` // "supabase", "aws", "gcp"
    Auth     string `json:"auth"`     // "supabase", "aws", "gcp"
    Cache    string `json:"cache"`    // "supabase", "aws", "gcp"
    Storage  string `json:"storage"`  // "supabase", "aws", "gcp"
}
```

## 🧩 **Module Architecture Integration**

### **Module Integration Manager**

```go
// internal/architecture/module_integration.go
type ModuleIntegrationManager struct {
    moduleManager    *ModuleManager
    lifecycleManager *LifecycleManager
    config           ModuleIntegrationConfig
    database         database.Database // Works with any provider
    logger           *observability.Logger
}

func (mim *ModuleIntegrationManager) InitializeModules(ctx context.Context) error {
    // Register core modules
    if err := mim.registerCoreModules(); err != nil {
        return err
    }
    
    // Initialize provider-specific modules
    if mim.config.DatabaseProvider == "supabase" {
        return mim.initializeSupabaseModules(ctx)
    }
    
    return nil
}
```

### **Provider-Aware Module Registration**

```go
func (mim *ModuleIntegrationManager) getProviderForModule(moduleID string) string {
    switch {
    case mim.isSupabaseModule(moduleID):
        return "supabase"
    case mim.isRailwayModule(moduleID):
        return "railway"
    case mim.isDatabaseModule(moduleID):
        return mim.config.DatabaseProvider // Uses existing provider config
    default:
        return "core"
    }
}
```

## 🔧 **CI/CD Pipeline Integration**

### **GitHub Actions Workflow**

```yaml
# .github/workflows/ci-cd.yml
name: CI/CD Pipeline

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: "1.24"
      - run: go test ./...

  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4
      - uses: docker/setup-buildx-action@v3
      - uses: docker/build-push-action@v5
        with:
          context: .
          platforms: linux/amd64,linux/arm64
          push: true
          tags: ${{ steps.meta.outputs.tags }}

  deploy:
    runs-on: ubuntu-latest
    needs: build
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      - run: railway up
```

## 📊 **Monitoring and Observability**

### **Health Checks**

```go
// Health check endpoint with module status
func (h *HealthHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
    response := HealthResponse{
        Status:    "healthy",
        Timestamp: time.Now(),
        Version:   "1.0.0",
        
        // Module status integration
        Modules: h.moduleIntegrationManager.GetModuleStatus(),
        
        // Infrastructure status
        Infrastructure: InfrastructureStatus{
            Database: h.database.Ping(r.Context()) == nil,
            Railway:  h.railwayHealthCheck(),
            Supabase: h.supabaseHealthCheck(),
        },
    }
    
    json.NewEncoder(w).Encode(response)
}
```

### **Metrics Collection**

```go
// OpenTelemetry integration
func (mim *ModuleIntegrationManager) InitializeModules(ctx context.Context) error {
    _, span := mim.tracer.Start(ctx, "InitializeModules")
    defer span.End()
    
    span.SetAttributes(
        attribute.String("provider.database", mim.config.DatabaseProvider),
        attribute.String("deployment.platform", "railway"),
    )
    
    // Module initialization with tracing
    return mim.initializeModulesWithTracing(ctx, span)
}
```

## 🚀 **Deployment Scenarios**

### **1. Local Development**

```bash
# Start local development environment
docker-compose up -d

# Access application
curl http://localhost:8080/health

# View logs
docker-compose logs -f kyb-platform
```

### **2. Railway Beta Deployment**

```bash
# Deploy to Railway
railway up

# Set environment variables
railway variables set SUPABASE_URL=https://your-project.supabase.co
railway variables set SUPABASE_ANON_KEY=your-anon-key

# Monitor deployment
railway logs
```

### **3. Production Deployment**

```bash
# Build production image
docker build -t kyb-platform:latest .

# Deploy to production
kubectl apply -f k8s/production/

# Monitor deployment
kubectl get pods -n production
```

## 🔒 **Security Considerations**

### **Docker Security**

- **Non-root users**: All containers run as non-root users
- **Multi-stage builds**: Minimize attack surface
- **Health checks**: Automated health monitoring
- **Secrets management**: Environment variables for sensitive data

### **Railway Security**

- **Automatic HTTPS**: Railway provides SSL certificates
- **Environment isolation**: Separate environments for staging/production
- **Secrets management**: Railway secrets for sensitive data
- **Access control**: Railway team management

### **Supabase Security**

- **Row Level Security (RLS)**: Database-level security policies
- **JWT authentication**: Secure token-based authentication
- **API key management**: Separate keys for different access levels
- **Data encryption**: Automatic encryption at rest and in transit

## 📈 **Performance Optimization**

### **Docker Optimization**

```dockerfile
# Multi-stage build for smaller images
FROM golang:1.24-alpine AS builder
# Build with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o kyb-platform ./cmd/api

FROM alpine:latest
# Minimal runtime image
COPY --from=builder /app/kyb-platform .
```

### **Railway Optimization**

```json
// railway.json with performance settings
{
  "deploy": {
    "healthcheckPath": "/health",
    "healthcheckTimeout": 300,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 10
  }
}
```

### **Supabase Optimization**

```sql
-- Database optimizations
CREATE INDEX idx_business_classifications_user_id ON public.business_classifications(user_id);
CREATE INDEX idx_business_classifications_created_at ON public.business_classifications(created_at);

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
```

## 🎯 **Benefits of This Infrastructure**

### **1. Cost Effectiveness**
- **Railway**: Pay-per-use pricing model
- **Supabase**: Generous free tier with reasonable paid plans
- **Docker**: Efficient resource utilization

### **2. Scalability**
- **Railway**: Automatic scaling based on demand
- **Supabase**: Built-in scaling and performance optimization
- **Docker**: Horizontal scaling capabilities

### **3. Developer Experience**
- **Railway**: Simple deployment with Git integration
- **Supabase**: Excellent developer tools and dashboard
- **Docker**: Consistent development environment

### **4. Reliability**
- **Railway**: High availability with automatic failover
- **Supabase**: Managed database with automatic backups
- **Docker**: Consistent runtime environment

### **5. Security**
- **Railway**: Built-in security features
- **Supabase**: Enterprise-grade security
- **Docker**: Container isolation and security

## 🔄 **Migration and Evolution**

### **Provider Migration Strategy**

The infrastructure is designed to support easy migration between providers:

```bash
# Switch from Supabase to AWS
export PROVIDER_DATABASE=aws
export PROVIDER_AUTH=aws
export PROVIDER_CACHE=aws

# Deploy with new provider
railway up
```

### **Infrastructure Evolution**

The modular architecture allows for gradual infrastructure evolution:

1. **Start with Supabase**: Use Supabase for MVP and beta testing
2. **Scale with Railway**: Leverage Railway for deployment and scaling
3. **Optimize with Docker**: Use Docker for consistent environments
4. **Migrate as needed**: Easy migration to other providers when required

---

**Key Takeaway**: This infrastructure provides a robust, scalable, and cost-effective foundation that supports rapid development, easy deployment, and future growth while maintaining flexibility for provider changes and architectural evolution.
