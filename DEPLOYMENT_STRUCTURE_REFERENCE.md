# Deployment Structure Reference

## 🚀 Monorepo Deployment Architecture

### Service Deployment Paths

**API Service**
- **Source Directory**: `services/api/`
- **Railway Service**: `shimmering-comfort` (API backend)
- **Deployment Trigger**: Changes to `services/api/**` or `shared/**`
- **Build Command**: `go build -o server ./cmd/server/main.go`
- **Dockerfile**: `services/api/Dockerfile`

**Frontend Service**
- **Source Directory**: `services/frontend/`
- **Railway Service**: `frontend-UI` (Web frontend)
- **Deployment Trigger**: Changes to `services/frontend/**`
- **Build Command**: `go build -o frontend-server ./cmd/main.go`
- **Dockerfile**: `services/frontend/Dockerfile`

### Development Commands

```bash
# API Service
make dev-api          # Start API locally
make test-api         # Test API service
make deploy-api       # Deploy API to Railway

# Frontend Service
make dev-frontend     # Start frontend locally
make test-frontend    # Test frontend service
make deploy-frontend  # Deploy frontend to Railway

# Both Services
docker-compose up     # Start both services locally
```

### CI/CD Pipeline

- **API Changes**: GitHub Actions workflow `.github/workflows/api-ci.yml`
- **Frontend Changes**: GitHub Actions workflow `.github/workflows/frontend-ci.yml`
- **Independent Deployments**: Each service deploys only when its files change

### Production URLs

- **API**: https://shimmering-comfort-production.up.railway.app
- **Frontend**: https://frontend-ui-production-e727.up.railway.app

---

**Remember**: Always work in the appropriate service directory (`services/api/` or `services/frontend/`) for changes to ensure proper deployment targeting.
