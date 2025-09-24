# KYB Platform API Service

## Overview
This is the backend API service for the KYB Platform, providing business classification, risk assessment, and data enrichment capabilities.

## Features
- Enhanced business classification (MCC, NAICS, SIC codes)
- Website scraping and keyword extraction
- Risk assessment and compliance checking
- Supabase integration for data persistence

## Development

### Local Development
```bash
# From project root
make dev-api

# Or directly
cd services/api
go run cmd/server/main.go
```

### Testing
```bash
make test-api
```

### Deployment
```bash
make deploy-api
```

## API Endpoints
- `GET /health` - Health check
- `POST /v1/classify` - Business classification
- `GET /api/v1/merchants` - Merchant management

## Configuration
Environment variables:
- `SUPABASE_URL` - Supabase project URL
- `SUPABASE_ANON_KEY` - Supabase anonymous key
- `PORT` - Server port (default: 8080)
