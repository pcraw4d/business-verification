# Prometheus Railway Configuration Guide

## Overview

This document explains the Prometheus configuration for scraping metrics from Railway-deployed services.

## Configuration Changes

### Updated Targets

The Prometheus configuration has been updated to scrape from Railway production URLs instead of localhost:

1. **API Gateway** (business-verification-v3-api)
   - URL: `https://api-gateway-service-production-21fd.up.railway.app`
   - Metrics Path: `/metrics`
   - Scheme: `https`

2. **Merchant Service**
   - URL: `https://merchant-service-production.up.railway.app`
   - Metrics Path: `/metrics`
   - Scheme: `https`

3. **Risk Assessment Service**
   - URL: `https://risk-assessment-service-production.up.railway.app`
   - Metrics Path: `/metrics`
   - Scheme: `https`

### Local-Only Services

The following exporters remain configured for localhost (disable if not running locally):

- **Node Exporter**: `localhost:9100` - System metrics
- **Postgres Exporter**: `localhost:9187` - Database metrics
- **Redis Exporter**: `localhost:9121` - Cache metrics

These will show as DOWN if not running locally, which is expected.

## Verification Steps

### 1. Verify Services Expose Metrics

Check that each service exposes a `/metrics` endpoint:

```bash
# Merchant Service
curl https://merchant-service-production.up.railway.app/metrics

# Risk Assessment Service
curl https://risk-assessment-service-production.up.railway.app/metrics

# API Gateway
curl https://api-gateway-service-production-21fd.up.railway.app/metrics
```

### 2. Check Prometheus Targets

1. Open Prometheus UI: `http://localhost:9090`
2. Navigate to **Status** → **Targets**
3. Verify that Railway services show as **UP** (green)
4. Local exporters (node, postgres, redis) may show as **DOWN** if not running locally - this is expected

### 3. Verify Metrics Collection

1. In Prometheus UI, go to **Graph**
2. Try querying a metric:
   ```
   kyb_fallback_total
   ```
3. If metrics appear, scraping is working correctly

## Troubleshooting

### Services Show as DOWN

**Possible Causes:**
1. Services not deployed or not running
2. `/metrics` endpoint not exposed
3. Network/firewall blocking access
4. SSL certificate issues

**Solutions:**
1. Verify services are deployed and running on Railway
2. Check service logs for errors
3. Test metrics endpoint directly with curl
4. Verify Railway service URLs are correct

### Connection Refused Errors

If you see `connection refused` errors:
- For Railway services: Check that services are deployed and accessible
- For local exporters: These are expected if exporters aren't running locally

### SSL/TLS Errors

If you see SSL errors:
- Verify Railway URLs use HTTPS
- Check that Prometheus can validate Railway SSL certificates
- Consider adding `insecure_skip_verify: true` if needed (not recommended for production)

## Metrics Available

### Merchant Service Metrics

- `kyb_fallback_total` - Total fallback usage events
- `kyb_fallback_rate_percent` - Fallback usage rate
- `kyb_fallback_duration_seconds` - Fallback operation duration
- `kyb_requests_total` - Total requests (fallback and non-fallback)
- `kyb_fallback_by_category_total` - Fallback usage by category
- `kyb_fallback_by_source_total` - Fallback usage by source

### Risk Assessment Service Metrics

- Custom metrics via `/metrics` endpoint
- Additional metrics via `/monitoring/metrics` endpoint

### API Gateway Metrics

- Standard HTTP metrics
- Request/response metrics

## Next Steps

1. **Import Grafana Dashboard**: Use the fallback metrics dashboard (`monitoring/grafana-fallback-dashboard.json`)
2. **Set Up Alerts**: Configure alerts for high fallback usage rates
3. **Monitor Trends**: Track fallback usage over time to identify issues

## Configuration File

The Prometheus configuration is located at:
- `monitoring/prometheus.yml`

To apply changes:
1. Edit the configuration file
2. Restart Prometheus:
   ```bash
   docker-compose -f docker-compose.monitoring.yml restart prometheus
   ```
3. Or reload configuration:
   ```bash
   curl -X POST http://localhost:9090/-/reload
   ```

## Last Updated

**Date**: November 7, 2025  
**Status**: ✅ Configuration updated for Railway deployments

