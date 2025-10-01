# KYB Platform Railway Deployment Testing Guide

## Overview

This guide provides comprehensive testing procedures for the deployed KYB Platform microservices on Railway. All 8 services have been successfully deployed and are ready for testing.

## Deployed Services

| Service | Port | Railway Service Name | Health Endpoint |
|---------|------|---------------------|-----------------|
| API Gateway | 8080 | `api-gateway-service` | `/health` |
| Merchant Service | 8082 | `merchant-service` | `/health` |
| Classification Service | 8081 | `classification-service` | `/health` |
| Pipeline Service | 8085 | `pipeline-service` | `/health` |
| Frontend Service | 8086 | `frontend-service` | `/health` |
| Service Discovery | 8086 | `service-discovery` | `/health` |
| Business Intelligence | 8087 | `bi-service` | `/health` |
| Monitoring Service | 8084 | `monitoring-service` | `/health` |

## Testing Procedures

### 1. Get Railway URLs

First, obtain the Railway URLs for each service:

1. Go to your Railway dashboard
2. Click on each service
3. Go to the "Deployments" tab
4. Copy the public URL (e.g., `https://api-gateway-service-production.up.railway.app`)

### 2. Health Check Testing

Test that all services are running and healthy:

```bash
# Test each service's health endpoint
curl https://your-api-gateway-url/health
curl https://your-merchant-service-url/health
curl https://your-classification-service-url/health
curl https://your-pipeline-service-url/health
curl https://your-frontend-service-url/health
curl https://your-service-discovery-url/health
curl https://your-bi-service-url/health
curl https://your-monitoring-service-url/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "service": "service-name",
  "version": "1.0.0",
  "timestamp": "2025-01-30T..."
}
```

### 3. Core API Functionality Testing

#### Classification Service
Test the business classification endpoint:

```bash
curl -X POST https://your-classification-service-url/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Corporation",
    "description": "A technology company specializing in software development"
  }'
```

**Expected Response:**
```json
{
  "business_name": "Acme Corporation",
  "classifications": {
    "mcc": [
      {"code": "7372", "description": "Prepackaged Software", "confidence": 0.95}
    ],
    "naics": [
      {"code": "541511", "description": "Custom Computer Programming Services", "confidence": 0.92}
    ],
    "sic": [
      {"code": "7372", "description": "Prepackaged Software", "confidence": 0.88}
    ]
  }
}
```

#### Merchant Service
Test merchant management endpoints:

```bash
# Get merchants list
curl https://your-merchant-service-url/api/v1/merchants

# Get merchant analytics
curl https://your-merchant-service-url/api/v1/merchants/analytics

# Get merchant statistics
curl https://your-merchant-service-url/api/v1/merchants/statistics
```

#### API Gateway
Test the API Gateway proxy functionality:

```bash
# Test classification through API Gateway
curl -X POST https://your-api-gateway-url/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Company",
    "description": "A test business"
  }'

# Test merchant endpoints through API Gateway
curl https://your-api-gateway-url/api/v1/merchants
```

### 4. Business Intelligence Testing

Test the BI service endpoints:

```bash
# Executive Dashboard
curl https://your-bi-service-url/dashboard/executive

# KPIs
curl https://your-bi-service-url/dashboard/kpis

# Charts
curl https://your-bi-service-url/dashboard/charts

# Reports
curl https://your-bi-service-url/reports

# Business Insights
curl https://your-bi-service-url/insights
```

### 5. Monitoring Service Testing

Test monitoring and metrics:

```bash
# Metrics endpoint
curl https://your-monitoring-service-url/metrics

# Health check
curl https://your-monitoring-service-url/health
```

### 6. Service Discovery Testing

Test service discovery functionality:

```bash
# Health check
curl https://your-service-discovery-url/health

# List registered services
curl https://your-service-discovery-url/services
```

### 7. Frontend Service Testing

Test the frontend service:

```bash
# Root endpoint
curl https://your-frontend-service-url/

# Health check
curl https://your-frontend-service-url/health
```

## Inter-Service Communication Testing

### API Gateway Proxy Testing

The API Gateway should proxy requests to the appropriate services:

1. **Classification Proxy**: `POST /api/v1/classify` → Classification Service
2. **Merchant Proxy**: `GET /api/v1/merchants` → Merchant Service
3. **Analytics Proxy**: `GET /api/v1/merchants/analytics` → Merchant Service

### Service-to-Service Communication

Test that services can communicate with each other:

```bash
# Test that API Gateway can reach Classification Service
curl -X POST https://your-api-gateway-url/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test", "description": "Test business"}'

# Test that API Gateway can reach Merchant Service
curl https://your-api-gateway-url/api/v1/merchants
```

## Supabase Integration Testing

### Database Connection

Verify that services can connect to Supabase:

1. Check service logs for Supabase connection messages
2. Test endpoints that require database access
3. Verify authentication is working

### Authentication Testing

Test JWT token generation and validation:

```bash
# Generate token (if auth endpoint exists)
curl -X POST https://your-api-gateway-url/auth/token \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test-user"}'

# Validate token
curl -X POST https://your-api-gateway-url/auth/validate \
  -H "Content-Type: application/json" \
  -d '{"token": "your-jwt-token"}'
```

## Automated Testing Script

Use the provided testing script for comprehensive testing:

```bash
# Run the interactive testing script
./test_services_simple.sh
```

This script will:
1. Prompt for Railway URLs
2. Test all health endpoints
3. Test core API functionality
4. Test inter-service communication
5. Provide a summary of results

## Troubleshooting

### Common Issues

1. **Service Not Responding**
   - Check Railway service logs
   - Verify environment variables are set
   - Ensure service is deployed and running

2. **Supabase Connection Errors**
   - Verify `SUPABASE_URL`, `SUPABASE_ANON_KEY`, and `SUPABASE_SERVICE_ROLE_KEY` are set
   - Check Supabase project is active
   - Verify network connectivity

3. **Inter-Service Communication Failures**
   - Check service URLs are correct
   - Verify API Gateway routing configuration
   - Check for CORS issues

4. **Authentication Issues**
   - Verify JWT secret is set
   - Check token expiration
   - Verify user permissions

### Log Analysis

Check Railway logs for each service:

1. Go to Railway dashboard
2. Click on the service
3. Go to "Deployments" tab
4. Click on the latest deployment
5. View logs for errors

### Environment Variables

Ensure all required environment variables are set:

```bash
# Supabase Configuration
SUPABASE_URL=your-supabase-url
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_ROLE_KEY=your-service-role-key
SUPABASE_JWT_SECRET=your-jwt-secret

# Service Configuration
PORT=8080  # (varies by service)
HOST=0.0.0.0
LOG_LEVEL=info
```

## Performance Testing

### Load Testing

Test service performance under load:

```bash
# Install Apache Bench (if not installed)
# macOS: brew install httpd
# Ubuntu: sudo apt-get install apache2-utils

# Test classification endpoint
ab -n 100 -c 10 -T "application/json" -p test-data.json https://your-classification-service-url/api/v1/classify

# Test health endpoint
ab -n 1000 -c 50 https://your-api-gateway-url/health
```

### Response Time Monitoring

Monitor response times for critical endpoints:

```bash
# Time a classification request
time curl -X POST https://your-classification-service-url/api/v1/classify \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test", "description": "Test"}'
```

## Security Testing

### CORS Testing

Test CORS configuration:

```bash
# Test CORS preflight
curl -X OPTIONS https://your-api-gateway-url/api/v1/classify \
  -H "Origin: https://example.com" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type"
```

### Rate Limiting Testing

Test rate limiting:

```bash
# Make multiple rapid requests
for i in {1..10}; do
  curl https://your-api-gateway-url/health
done
```

## Success Criteria

All tests should pass for a successful deployment:

- ✅ All health endpoints return 200 OK
- ✅ Classification API returns proper business codes
- ✅ Merchant API returns merchant data
- ✅ API Gateway proxies requests correctly
- ✅ Inter-service communication works
- ✅ Supabase integration is functional
- ✅ Authentication works (if implemented)
- ✅ CORS is configured correctly
- ✅ Rate limiting is working
- ✅ Services respond within acceptable time limits

## Next Steps

After successful testing:

1. **Monitor Performance**: Set up monitoring and alerting
2. **Load Testing**: Test under realistic load conditions
3. **Security Audit**: Perform security testing
4. **Documentation**: Update API documentation
5. **User Acceptance Testing**: Test with real business data
6. **Production Readiness**: Ensure all production requirements are met

## Support

If you encounter issues:

1. Check Railway service logs
2. Verify environment variables
3. Test individual services
4. Check Supabase connectivity
5. Review this testing guide
6. Contact support if needed

---

**Last Updated**: January 30, 2025  
**Version**: 1.0.0
