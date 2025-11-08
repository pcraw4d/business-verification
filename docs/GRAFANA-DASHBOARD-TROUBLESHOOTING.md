# Grafana Dashboard Troubleshooting Guide

## Issue: Dashboard Shows No Data

If you've imported the Grafana dashboard but see no visualizations, follow these troubleshooting steps:

### Step 1: Verify Prometheus Data Source

1. Go to **Configuration** → **Data Sources**
2. Click on your Prometheus data source
3. Click **Save & Test**
4. Verify you see "Data source is working"

**Common Issues:**
- **URL incorrect**: Should be `http://prometheus:9090` (Docker) or `http://localhost:9090` (local)
- **Network access**: Prometheus must be accessible from Grafana

### Step 2: Check if Metrics Exist

1. Open Prometheus UI: `http://localhost:9090`
2. Go to **Graph** tab
3. Try these queries:
   ```
   kyb_fallback_rate_percent
   kyb_fallback_total
   kyb_requests_total
   ```
4. If queries return "No data", metrics haven't been generated yet

### Step 3: Verify Metrics Are Being Exported

Check if services are exporting metrics:

```bash
# Merchant Service
curl https://merchant-service-production.up.railway.app/metrics | grep kyb_

# API Gateway
curl https://api-gateway-service-production-21fd.up.railway.app/metrics | grep kyb_

# Risk Assessment Service
curl https://risk-assessment-service-production.up.railway.app/metrics | grep kyb_
```

**Expected Output:**
- If metrics exist, you'll see lines like:
  ```
  # HELP kyb_fallback_total Total number of fallback data usage events
  # TYPE kyb_fallback_total counter
  kyb_fallback_total{category="database_fallback",service="merchant-service",source="supabase"} 0
  ```

- If no output, metrics haven't been created yet (no fallback events occurred)

### Step 4: Generate Test Data

If no metrics exist, you need to trigger fallback events:

1. **Simulate Database Failure**:
   - Temporarily break Supabase connection
   - Make API requests to merchant service
   - Fallback metrics should be recorded

2. **Check Service Logs**:
   - Look for "Fallback usage recorded" messages
   - Verify metrics are being incremented

### Step 5: Fix Dashboard Queries for Empty Data

If metrics don't exist yet, update dashboard queries to show "No data" properly:

1. Edit the dashboard
2. For each panel, update the query to handle missing metrics:

**Original Query:**
```
kyb_fallback_rate_percent
```

**Updated Query (shows 0 when no data):**
```
kyb_fallback_rate_percent or vector(0)
```

**Or use default value:**
```
kyb_fallback_rate_percent or on() vector(0) * 0
```

### Step 6: Verify Time Range

1. Check the time range selector (top right)
2. Ensure it's set to a recent time range (e.g., "Last 1 hour")
3. If metrics are new, they might not appear in older time ranges

### Step 7: Check Panel Configuration

1. Click on a panel → **Edit**
2. Verify:
   - **Data source** is set to Prometheus
   - **Query** is correct
   - **Legend** format is set
   - **Unit** is appropriate (percent, count, etc.)

### Step 8: Verify Prometheus Scraping

1. Go to Prometheus → **Status** → **Targets**
2. Verify services show as **UP** (green)
3. Check **Last Scrape** time is recent
4. If DOWN, check service logs and network connectivity

## Common Issues and Solutions

### Issue: "No data" in all panels

**Cause**: Metrics haven't been generated yet (no fallback events)

**Solution**:
1. Wait for fallback events to occur naturally
2. Or simulate a failure to generate test data
3. Or update queries to show 0 when metrics don't exist

### Issue: Some panels show data, others don't

**Cause**: Specific metrics don't exist for those services

**Solution**:
1. Check which services are actually recording metrics
2. Update queries to filter by service: `kyb_fallback_rate_percent{service="merchant-service"}`
3. Remove panels for services that don't export metrics yet

### Issue: "Query failed" errors

**Cause**: Prometheus query syntax error or datasource issue

**Solution**:
1. Test query directly in Prometheus UI
2. Verify datasource is configured correctly
3. Check Grafana logs for detailed error messages

### Issue: Data appears but is all zeros

**Cause**: Metrics exist but no fallback events have occurred (this is good!)

**Solution**:
- This is expected behavior when services are healthy
- Metrics will populate when fallback events occur
- Consider adding a note in the dashboard explaining this

## Quick Test: Generate Fallback Metrics

To test the dashboard, you can temporarily cause a fallback:

1. **Stop Supabase connection** (or use invalid credentials)
2. **Make API requests**:
   ```bash
   curl https://merchant-service-production.up.railway.app/api/v1/merchants/test-id
   ```
3. **Check metrics**:
   ```bash
   curl https://merchant-service-production.up.railway.app/metrics | grep kyb_fallback
   ```
4. **Refresh Grafana dashboard** - data should appear

## Dashboard Query Reference

### Fallback Rate
```
kyb_fallback_rate_percent{service="merchant-service"}
```

### Total Fallback Events
```
sum(rate(kyb_fallback_total[5m])) by (service)
```

### Fallback by Category
```
sum(kyb_fallback_by_category_total) by (category)
```

### Fallback by Source
```
sum(kyb_fallback_by_source_total) by (source)
```

### Fallback Duration (P95)
```
histogram_quantile(0.95, sum(rate(kyb_fallback_duration_seconds_bucket[5m])) by (service, le))
```

### Total Requests
```
sum(rate(kyb_requests_total[5m])) by (service)
```

## Next Steps

1. **Wait for natural fallback events** - Metrics will appear automatically
2. **Monitor service health** - Healthy services won't generate fallback metrics
3. **Set up alerts** - Alert when fallback rate exceeds thresholds
4. **Review dashboard regularly** - Track trends over time

## Still Having Issues?

1. Check Grafana logs: `docker logs grafana`
2. Check Prometheus logs: `docker logs prometheus`
3. Verify services are running and accessible
4. Test queries directly in Prometheus UI
5. Check service logs for metric recording errors

