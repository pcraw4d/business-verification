# KYB Platform Frontend Service

## Overview
This is the frontend web service for the KYB Platform, providing the user interface for business classification and management.

## Features
- Business classification interface
- Real-time results display
- Enhanced UI with modern design
- API integration for backend services

## Development

### Local Development
```bash
# From project root
make dev-frontend

# Or directly
cd services/frontend
go run cmd/main.go
```

### Testing
```bash
make test-frontend
```

### Deployment
```bash
make deploy-frontend
```

## Configuration
The frontend is configured to call the API service at:
- Production: `https://creative-determination-production.up.railway.app/v1`
- Local: `http://localhost:8080/v1`

## File Structure
- `public/` - Static HTML, CSS, JS files
- `cmd/` - Go-based static file server
- `Dockerfile` - Container configuration
