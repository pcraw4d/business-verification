# Risk Assessment Service

## Overview
This is the risk assessment service for the KYB Platform, providing comprehensive business risk assessment and predictive analytics capabilities.

## Features
- Real-time risk assessment with sub-1-second response times
- XGBoost and LSTM machine learning models for risk prediction
- 3-12 month risk forecasting capabilities
- SHAP explainability for risk factors
- Integration with external data sources (Thomson Reuters, OFAC, etc.)
- Comprehensive compliance screening
- Multi-tenant architecture for enterprise customers

## Development

### Local Development
```bash
# From project root
make dev-risk-assessment

# Or directly
cd services/risk-assessment-service
go run cmd/main.go
```

### Testing
```bash
make test-risk-assessment
```

### Deployment
```bash
make deploy-risk-assessment
```

## Configuration
The service is configured to integrate with:
- Supabase for data storage
- Redis for caching
- External APIs for risk data
- Prometheus for monitoring

## File Structure
- `cmd/` - Main application entry point
- `internal/` - Private application code
  - `config/` - Configuration management
  - `handlers/` - HTTP handlers
  - `models/` - Data models and structures
  - `ml/` - Machine learning models and training
  - `repository/` - Data access layer
  - `external/` - External API integrations
  - `validation/` - Input validation
- `pkg/client/` - Go client SDK
- `api/` - OpenAPI specifications
- `Dockerfile` - Container configuration
