# Development Workflow Guide

## 🎯 **Overview**

This guide explains how to effectively use GitHub and Railway CLI with the new monorepo structure to achieve true frontend/backend separation and follow development best practices.

## 🏗️ **New Architecture Benefits**

### **Before (Current Issues):**
- ❌ Frontend and backend mixed in same directory
- ❌ Changes to one service trigger deployment of both
- ❌ Multiple confusing Dockerfiles
- ❌ No clear service boundaries
- ❌ Difficult to maintain and scale

### **After (Proposed Solution):**
- ✅ Clear service separation in `services/` directory
- ✅ Independent deployments (frontend changes don't affect backend)
- ✅ Service-specific CI/CD pipelines
- ✅ Single responsibility per service
- ✅ Easy to maintain and scale

## 📁 **New Directory Structure**

```
kyb-platform/
├── services/
│   ├── api/                    # Backend API Service
│   │   ├── cmd/server/        # API server entry point
│   │   ├── internal/          # Private API code
│   │   ├── pkg/              # Public API packages
│   │   ├── Dockerfile        # API-specific Dockerfile
│   │   ├── railway.json      # API Railway config
│   │   └── README.md         # API documentation
│   │
│   └── frontend/              # Frontend Web Service
│       ├── public/           # Static web files
│       ├── cmd/              # Frontend server
│       ├── Dockerfile        # Frontend-specific Dockerfile
│       ├── railway.json      # Frontend Railway config
│       └── README.md         # Frontend documentation
│
├── shared/                    # Shared utilities
├── .github/workflows/         # Service-specific CI/CD
├── scripts/                   # Build and deployment scripts
└── Makefile                   # Development commands
```

## 🚀 **Implementation Steps**

### **Step 1: Restructure Codebase**
```bash
# Run the restructuring script
./scripts/restructure-codebase.sh
```

This script will:
- Create new directory structure
- Move backend code to `services/api/`
- Move frontend code to `services/frontend/`
- Create service-specific configurations
- Set up GitHub Actions workflows
- Create development tools (Makefile, docker-compose.yml)

### **Step 2: Update Railway Configurations**
```bash
# Update Railway service configurations
./scripts/update-railway-configs.sh
```

This script will:
- Update API service to use `services/api/` as root directory
- Update Frontend service to use `services/frontend/` as root directory
- Set correct Dockerfile paths and start commands

### **Step 3: Test Local Development**
```bash
# Test API service locally
make dev-api

# Test frontend service locally (in another terminal)
make dev-frontend

# Or use Docker Compose for both
docker-compose up
```

### **Step 4: Deploy and Test**
```bash
# Deploy API service
make deploy-api

# Deploy frontend service
make deploy-frontend
```

## 🔄 **Development Workflow**

### **Making API Changes**
1. **Edit files** in `services/api/`
2. **Test locally**: `make dev-api`
3. **Run tests**: `make test-api`
4. **Commit changes**: `git add services/api/ && git commit -m "API: Add new feature"`
5. **Push to GitHub**: `git push`
6. **Automatic deployment**: GitHub Actions will deploy only the API service

### **Making Frontend Changes**
1. **Edit files** in `services/frontend/`
2. **Test locally**: `make dev-frontend`
3. **Run tests**: `make test-frontend`
4. **Commit changes**: `git add services/frontend/ && git commit -m "Frontend: Update UI"`
5. **Push to GitHub**: `git push`
6. **Automatic deployment**: GitHub Actions will deploy only the frontend service

### **Making Shared Changes**
1. **Edit files** in `shared/`
2. **Test both services**: `make test-api && make test-frontend`
3. **Commit changes**: `git add shared/ && git commit -m "Shared: Update utilities"`
4. **Push to GitHub**: `git push`
5. **Automatic deployment**: Both services will redeploy

## 🛠️ **GitHub Actions Workflows**

### **API Service CI/CD** (`.github/workflows/api-ci.yml`)
```yaml
# Triggers on changes to:
# - services/api/**
# - shared/**

# Jobs:
# 1. Test - Run Go tests
# 2. Build - Compile the application
# 3. Deploy - Deploy to Railway (if main branch)
```

### **Frontend Service CI/CD** (`.github/workflows/frontend-ci.yml`)
```yaml
# Triggers on changes to:
# - services/frontend/**

# Jobs:
# 1. Test - Run frontend tests
# 2. Build - Compile the application
# 3. Deploy - Deploy to Railway (if main branch)
```

## 🎯 **Best Practices**

### **1. Service Separation**
- ✅ Keep API and frontend code completely separate
- ✅ Use shared utilities for common functionality
- ✅ Maintain clear service boundaries
- ❌ Don't mix frontend and backend code

### **2. Git Workflow**
- ✅ Use feature branches for new development
- ✅ Make atomic commits (one logical change per commit)
- ✅ Use descriptive commit messages with service prefix
- ✅ Test locally before pushing

### **3. Deployment Strategy**
- ✅ Deploy services independently
- ✅ Use feature flags for gradual rollouts
- ✅ Monitor deployments and rollback if needed
- ✅ Keep production and staging environments in sync

### **4. Development Environment**
- ✅ Use `make` commands for common tasks
- ✅ Use Docker Compose for local development
- ✅ Keep dependencies up to date
- ✅ Document any setup requirements

## 📋 **Available Commands**

### **Development Commands**
```bash
make help          # Show all available commands
make dev-api       # Start API service locally
make dev-frontend  # Start frontend service locally
make test-api      # Run API tests
make test-frontend # Run frontend tests
```

### **Deployment Commands**
```bash
make deploy-api      # Deploy API to Railway
make deploy-frontend # Deploy frontend to Railway
```

### **Utility Commands**
```bash
make clean          # Clean build artifacts
docker-compose up   # Start both services with Docker
```

## 🔧 **Railway CLI Usage**

### **Service Management**
```bash
# List all services
railway status

# Switch to API service
railway service shimmering-comfort

# Switch to Frontend service
railway service frontend-UI

# Deploy current service
railway up --detach

# View logs
railway logs

# Get service URL
railway domain
```

### **Environment Variables**
```bash
# Set environment variables
railway variables --set "KEY=value"

# View environment variables
railway variables

# Set multiple variables
railway variables --set "KEY1=value1" --set "KEY2=value2"
```

## 🚨 **Troubleshooting**

### **Common Issues**

#### **1. Service Not Deploying**
- Check Railway service configuration
- Verify Dockerfile path is correct
- Check build logs in Railway dashboard

#### **2. API Not Connecting to Frontend**
- Verify API_BASE_URL in frontend configuration
- Check CORS settings in API service
- Ensure both services are deployed

#### **3. Local Development Issues**
- Check if ports are available (8080 for API, 3000 for frontend)
- Verify environment variables are set
- Check Docker containers are running

### **Debugging Commands**
```bash
# Check service status
railway status

# View service logs
railway logs --follow

# Test API locally
curl http://localhost:8080/health

# Test frontend locally
curl http://localhost:3000/
```

## 📈 **Monitoring and Maintenance**

### **Health Checks**
- **API**: `https://shimmering-comfort-production.up.railway.app/health`
- **Frontend**: `https://frontend-ui-production-e727.up.railway.app/`

### **Performance Monitoring**
- Monitor Railway dashboard for resource usage
- Set up alerts for service failures
- Track deployment success rates

### **Regular Maintenance**
- Update dependencies monthly
- Review and clean up unused code
- Monitor service performance
- Update documentation as needed

## 🎉 **Benefits Achieved**

### **Development Benefits**
- ✅ **Faster Development**: Work on services independently
- ✅ **Clearer Codebase**: Easy to find and modify code
- ✅ **Better Testing**: Service-specific test suites
- ✅ **Easier Debugging**: Isolated service issues

### **Deployment Benefits**
- ✅ **Independent Deployments**: Deploy services separately
- ✅ **Faster Builds**: Only changed services rebuild
- ✅ **Reduced Risk**: Changes don't affect other services
- ✅ **Better Monitoring**: Service-specific metrics

### **Team Benefits**
- ✅ **Clear Ownership**: Teams can own specific services
- ✅ **Parallel Development**: Multiple developers can work simultaneously
- ✅ **Easier Onboarding**: Clear service boundaries
- ✅ **Better Collaboration**: Well-defined interfaces

This new structure provides a solid foundation for scalable, maintainable development while following industry best practices for microservices architecture.
