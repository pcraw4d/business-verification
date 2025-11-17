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

### Frontend Service (`frontend/` and `cmd/frontend-service/`)
- **Purpose**: Web interface for business classification, risk assessment, and compliance
- **Technology**: Next.js with shadcn UI
- **Deployment**: Railway (frontend-UI service)
- **Status**: ✅ New UI is default, legacy UI removed (Phase 4 complete)

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
- [Frontend Service](frontend/README.md)
- [UI Migration Complete](docs/UI_MIGRATION_COMPLETE.md)
- [Legacy UI Deprecation](docs/LEGACY_UI_DEPRECATION.md)
- [Frontend UI Audit](docs/FRONTEND_UI_AUDIT_REPORT.md)
- [Development Guidelines](docs/development-guidelines.md)

## Frontend UI

The frontend uses Next.js with shadcn UI components. The legacy HTML/CSS/JS UI has been completely removed (Phase 4 complete).

- **UI**: Next.js with shadcn UI components
- **Location**: `frontend/` directory
- **Default**: New UI is the default (no environment variables needed)
- **Legacy UI**: Removed and archived to `archive/legacy-ui/`
- **Status**: ✅ All pages migrated