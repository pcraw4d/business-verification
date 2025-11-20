# Local Development Setup Guide

This guide explains how to configure the API Gateway for local development without impacting Railway deployment.

## Service URL Configuration

The API Gateway automatically detects the environment and uses appropriate service URLs:

- **Development (`ENVIRONMENT=development`)**: Uses `localhost` URLs with configurable ports
- **Production (default)**: Uses Railway URLs

## Environment Variables

### Required for Local Development

Set `ENVIRONMENT=development` to enable localhost URLs:

```bash
export ENVIRONMENT=development
```

### Optional Service Port Configuration

By default, services use these ports:
- Classification Service: `8081`
- Merchant Service: `8082`
- Frontend Service: `3000`
- BI Service: `8083`
- Risk Assessment Service: `8084`

You can override these with environment variables:

```bash
export CLASSIFICATION_SERVICE_PORT=8081
export MERCHANT_SERVICE_PORT=8082
export FRONTEND_SERVICE_PORT=3000
export BI_SERVICE_PORT=8083
export RISK_ASSESSMENT_SERVICE_PORT=8084
```

### Override Service URLs

If you need to override service URLs completely (e.g., for testing with remote services), set the service URL environment variables:

```bash
export CLASSIFICATION_SERVICE_URL=http://localhost:8081
export MERCHANT_SERVICE_URL=http://localhost:8082
export FRONTEND_URL=http://localhost:3000
export BI_SERVICE_URL=http://localhost:8083
export RISK_ASSESSMENT_SERVICE_URL=http://localhost:8084
```

## How It Works

The `getServiceURL()` function in `internal/config/config.go`:

1. **Checks for explicit URL**: If `CLASSIFICATION_SERVICE_URL` (etc.) is set, uses it
2. **Development mode**: If `ENVIRONMENT=development`, uses `http://localhost:{PORT}`
3. **Production mode**: Uses Railway URLs as defaults

## Railway Deployment

**No changes needed!** Railway deployments:
- Use `ENVIRONMENT=production` (or don't set it, defaults to production)
- Use Railway URLs automatically
- Can override with explicit service URL environment variables if needed

## Example Setup

### Local Development

```bash
# Set environment to development
export ENVIRONMENT=development

# Optional: Override ports if services run on different ports
export RISK_ASSESSMENT_SERVICE_PORT=8085

# Start API Gateway
cd services/api-gateway
go run cmd/main.go
```

### Railway Deployment

```bash
# No environment variables needed - uses production defaults
# Or explicitly set:
export ENVIRONMENT=production
# Service URLs will use Railway URLs automatically
```

## Testing

After setting up local development:

1. **Verify service URLs**:
   ```bash
   curl http://localhost:8080/health
   ```

2. **Check service connectivity**:
   ```bash
   curl http://localhost:8080/api/v1/risk/health
   curl http://localhost:8080/api/v1/merchant/health
   ```

3. **Test routes**:
   ```bash
   cd services/api-gateway
   ./scripts/test-routes.sh
   ```

## Troubleshooting

### Services Not Accessible

If services return 404:
1. Verify services are running locally
2. Check service ports match configuration
3. Verify `ENVIRONMENT=development` is set
4. Check service URL environment variables

### Railway Deployment Issues

If Railway deployment fails:
1. Ensure `ENVIRONMENT=production` (or not set)
2. Verify Railway service URLs are correct
3. Check service URL environment variables in Railway dashboard

## Migration Notes

This change is **backward compatible**:
- Existing Railway deployments continue to work (defaults to production)
- Local development requires `ENVIRONMENT=development` to be set
- Explicit service URL environment variables override defaults

