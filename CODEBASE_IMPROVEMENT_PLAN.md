# Codebase Improvement Plan: True Frontend/Backend Separation

## 🎯 **Objective**
Achieve true separation between frontend and backend while maintaining efficient development workflows and following best practices.

## 📊 **Current State Analysis**

### **Issues Identified:**
- ❌ Mixed frontend/backend files in same repository
- ❌ 8+ different Dockerfiles causing confusion
- ❌ Multiple overlapping entry points in `cmd/`
- ❌ No clear deployment boundaries
- ❌ Changes to one service can trigger deployment of the other

### **Strengths to Preserve:**
- ✅ Working Railway deployment setup
- ✅ Enhanced classification features
- ✅ Supabase integration
- ✅ Go-based architecture

## 🏗️ **Proposed Structure: Monorepo with Service Separation**

```
kyb-platform/
├── services/
│   ├── api/                          # Backend API Service
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go           # Single, clean entry point
│   │   ├── internal/                 # Private API code
│   │   │   ├── handlers/            # HTTP handlers
│   │   │   ├── middleware/          # HTTP middleware
│   │   │   ├── classification/      # Business logic
│   │   │   ├── repository/          # Data access
│   │   │   └── config/              # Configuration
│   │   ├── pkg/                     # Public API packages
│   │   ├── Dockerfile               # Single, clean Dockerfile
│   │   ├── go.mod                   # Service-specific dependencies
│   │   ├── railway.json             # Railway config for API
│   │   └── README.md                # Service documentation
│   │
│   └── frontend/                     # Frontend Web Service
│       ├── public/                  # Static assets
│       │   ├── index.html
│       │   ├── dashboard.html
│       │   └── assets/
│       ├── src/                     # Source code (if using build tools)
│       ├── Dockerfile               # Frontend-specific Dockerfile
│       ├── package.json             # Frontend dependencies
│       ├── railway.json             # Railway config for frontend
│       └── README.md                # Frontend documentation
│
├── shared/                          # Shared utilities and types
│   ├── types/                      # Shared type definitions
│   ├── config/                     # Shared configuration
│   └── utils/                      # Shared utilities
│
├── docs/                           # Documentation
├── scripts/                        # Build and deployment scripts
├── .github/                        # GitHub Actions workflows
│   └── workflows/
│       ├── api-ci.yml              # API service CI/CD
│       ├── frontend-ci.yml         # Frontend service CI/CD
│       └── deploy.yml              # Deployment workflow
├── docker-compose.yml              # Local development
├── Makefile                        # Development commands
└── README.md                       # Main project documentation
```

## 🚀 **Implementation Plan**

### **Phase 1: Restructure Repository (Week 1)**

#### **Step 1.1: Create Service Directories**
```bash
mkdir -p services/api services/frontend shared/{types,config,utils}
```

#### **Step 1.2: Move Backend Code**
```bash
# Move API-related files
mv cmd/railway-server services/api/cmd/server/
mv internal/ services/api/
mv pkg/ services/api/
mv go.mod services/api/
mv go.sum services/api/
```

#### **Step 1.3: Move Frontend Code**
```bash
# Move frontend files
mv web/* services/frontend/public/
mv cmd/frontend-server services/frontend/
```

#### **Step 1.4: Clean Up Root Directory**
```bash
# Remove old files
rm -rf cmd/ internal/ pkg/ web/
rm Dockerfile* railway*.json
```

### **Phase 2: Service-Specific Configuration (Week 1)**

#### **Step 2.1: API Service Configuration**
```yaml
# services/api/railway.json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile"
  },
  "deploy": {
    "startCommand": "./server",
    "restartPolicyType": "ON_FAILURE"
  }
}
```

#### **Step 2.2: Frontend Service Configuration**
```yaml
# services/frontend/railway.json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile"
  },
  "deploy": {
    "startCommand": "./frontend-server",
    "restartPolicyType": "ON_FAILURE"
  }
}
```

### **Phase 3: GitHub Actions Workflows (Week 2)**

#### **Step 3.1: Service-Specific CI/CD**
```yaml
# .github/workflows/api-ci.yml
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
  
  deploy:
    if: github.ref == 'refs/heads/main'
    needs: test
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Railway
        run: |
          cd services/api
          railway up --detach
```

#### **Step 3.2: Frontend CI/CD**
```yaml
# .github/workflows/frontend-ci.yml
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
      - name: Test Frontend
        run: |
          cd services/frontend
          # Add frontend tests here
  
  deploy:
    if: github.ref == 'refs/heads/main'
    needs: test
    runs-on: ubuntu-latest
    steps:
      - name: Deploy to Railway
        run: |
          cd services/frontend
          railway up --detach
```

### **Phase 4: Development Workflow (Week 2)**

#### **Step 4.1: Makefile for Development**
```makefile
# Makefile
.PHONY: help dev-api dev-frontend test-api test-frontend deploy-api deploy-frontend

help:
	@echo "Available commands:"
	@echo "  dev-api       - Start API service locally"
	@echo "  dev-frontend  - Start frontend service locally"
	@echo "  test-api      - Run API tests"
	@echo "  test-frontend - Run frontend tests"
	@echo "  deploy-api    - Deploy API to Railway"
	@echo "  deploy-frontend - Deploy frontend to Railway"

dev-api:
	cd services/api && go run cmd/server/main.go

dev-frontend:
	cd services/frontend && go run cmd/frontend-server/main.go

test-api:
	cd services/api && go test ./...

test-frontend:
	cd services/frontend && npm test

deploy-api:
	cd services/api && railway up --detach

deploy-frontend:
	cd services/frontend && railway up --detach
```

#### **Step 4.2: Docker Compose for Local Development**
```yaml
# docker-compose.yml
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
      - "3000:3000"
    volumes:
      - ./services/frontend:/app
    command: go run cmd/frontend-server/main.go
```

## 🎯 **Benefits of This Structure**

### **1. True Separation**
- ✅ Frontend changes don't trigger backend deployments
- ✅ Backend changes don't affect frontend
- ✅ Independent versioning and releases
- ✅ Separate dependency management

### **2. Improved Development Workflow**
- ✅ Service-specific CI/CD pipelines
- ✅ Faster builds (only changed services rebuild)
- ✅ Independent testing and deployment
- ✅ Clear ownership and responsibility

### **3. Better Maintainability**
- ✅ Single responsibility per service
- ✅ Clear boundaries and interfaces
- ✅ Easier onboarding for new developers
- ✅ Simplified debugging and monitoring

### **4. Scalability**
- ✅ Independent scaling of services
- ✅ Technology flexibility per service
- ✅ Microservices-ready architecture
- ✅ Easy to add new services

## 🛠️ **Migration Strategy**

### **Option 1: Gradual Migration (Recommended)**
1. Create new structure alongside existing
2. Migrate one service at a time
3. Test thoroughly before removing old structure
4. Update documentation and workflows

### **Option 2: Big Bang Migration**
1. Create complete new structure
2. Migrate all code at once
3. Update all configurations
4. Deploy and test everything

## 📋 **Next Steps**

1. **Review and Approve Plan** - Confirm this approach meets your needs
2. **Choose Migration Strategy** - Gradual vs Big Bang
3. **Create Implementation Timeline** - Set milestones and deadlines
4. **Begin Phase 1** - Start with repository restructuring

## 🔧 **Tools and Technologies**

- **Repository Management**: Git with feature branches
- **CI/CD**: GitHub Actions with service-specific workflows
- **Deployment**: Railway with separate service configurations
- **Local Development**: Docker Compose + Makefile
- **Monitoring**: Service-specific logging and metrics
- **Documentation**: Service-specific READMEs + main project docs

This structure will give you true separation, better development practices, and a scalable foundation for future growth.
