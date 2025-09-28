# KYB Business Intelligence Gateway

This service provides comprehensive business intelligence capabilities for the KYB platform, including executive dashboards, custom reports, data export, and advanced analytics.

## Features

### Executive Dashboard
- Real-time KPI monitoring
- Interactive visualizations
- Performance metrics overview
- Revenue and growth tracking

### Report Management
- Custom report generation
- Scheduled reports
- Multiple export formats (CSV, JSON, XLSX, PDF)
- Report templates

### Data Export
- Bulk data export
- Multiple formats support
- Secure download links
- Export history tracking

### Business Intelligence
- Advanced analytics
- Trend analysis
- Performance insights
- Actionable recommendations

## API Endpoints

### Health & Status
- `GET /health` - Service health check
- `GET /status` - Service status

### Executive Dashboard
- `GET /dashboard/executive` - Executive dashboard data
- `GET /dashboard/kpis` - Key performance indicators
- `GET /dashboard/charts` - Dashboard charts data

### Report Management
- `GET /reports` - List all reports
- `POST /reports` - Create new report
- `POST /reports/{id}/generate` - Generate report
- `GET /reports/templates` - Available report templates

### Data Export
- `POST /export` - Export data

### Business Intelligence
- `GET /insights` - Business insights and recommendations

## Configuration

The service uses environment variables for configuration:

- `PORT` - Service port (default: 8080)
- `SERVICE_NAME` - Service name (default: kyb-business-intelligence-gateway)
- `VERSION` - Service version (default: 4.0.0-BI)

## Development

### Local Development
```bash
cd cmd/business-intelligence-gateway
go run main.go
```

### Docker Build
```bash
docker build -t kyb-business-intelligence-gateway .
docker run -p 8080:8080 kyb-business-intelligence-gateway
```

### Railway Deployment
The service is configured for Railway deployment with:
- Dockerfile-based build
- Health check endpoint
- Automatic restarts
- Resource limits

## Architecture

The service integrates with the business intelligence package to provide:
- Executive dashboard generation
- Report management and generation
- Data export capabilities
- Business insights and analytics

## Monitoring

The service provides comprehensive monitoring through:
- Health check endpoints
- Performance metrics
- Business intelligence insights
- Report generation tracking
