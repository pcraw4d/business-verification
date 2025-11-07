# Troubleshooting Fallback Metrics & Grafana Dashboard

## Issue: Metrics Not Being Exported

### Problem
When you run `curl http://localhost:8082/metrics | grep kyb_fallback`, you get:
```
curl: (7) Failed to connect to localhost port 8082 after 3 ms: Couldn't connect to server
```

### Solutions

#### 1. Check if Merchant Service is Running

```bash
# Check if service is running
ps aux | grep merchant-service

# Or check if port 8082 is in use
lsof -i :8082

# Or check with netstat
netstat -an | grep 8082
```

#### 2. Start the Merchant Service

If the service isn't running, start it:

```bash
# Navigate to merchant service directory
cd services/merchant-service

# Run the service
go run cmd/main.go

# Or if using Docker
docker-compose up merchant-service
```

#### 3. Verify Service is Listening on Correct Port

The merchant service defaults to port **8082**. Check the configuration:

```bash
# Check environment variable
echo $PORT

# Or check config file
cat services/merchant-service/.env
```

#### 4. Test Metrics Endpoint

Once the service is running:

```bash
# Test health endpoint first
curl http://localhost:8082/health

# Then test metrics endpoint
curl http://localhost:8082/metrics

# Filter for fallback metrics
curl http://localhost:8082/metrics | grep kyb_fallback
```

**Expected output** (if metrics are being recorded):
```
# HELP kyb_fallback_total Total number of fallback data usage events
# TYPE kyb_fallback_total counter
kyb_fallback_total{category="database_fallback",service="merchant-service",source="supabase"} 0

# HELP kyb_fallback_rate_percent Fallback usage rate as percentage of total requests
# TYPE kyb_fallback_rate_percent gauge
kyb_fallback_rate_percent{service="merchant-service"} 0
```

#### 5. Generate Some Fallback Events

If metrics show 0, you need to trigger some fallback scenarios:

```bash
# Simulate a database failure by stopping Supabase connection
# Or make requests that will trigger fallback

# Make a request that might trigger fallback
curl http://localhost:8082/api/v1/merchants/nonexistent-id
```

---

## Issue: No Panels Appearing in Grafana Dashboard

### Problem
After importing the dashboard, panels show "No data" or are empty.

### Solutions

#### 1. Verify Prometheus Data Source

1. Go to **Configuration** → **Data Sources**
2. Click on **Prometheus**
3. Click **Save & Test**
4. Should show: "Data source is working"

**If it fails:**
- Check Prometheus is running: `docker ps | grep prometheus`
- Verify Prometheus URL is correct:
  - Docker: `http://prometheus:9090`
  - Local: `http://localhost:9090`

#### 2. Check Prometheus is Scraping Merchant Service

1. Open Prometheus: `http://localhost:9090`
2. Go to **Status** → **Targets**
3. Find `merchant-service` target
4. Should show **State: UP**

**If it shows DOWN:**
- Check merchant service is running (see above)
- Verify port 8082 is accessible
- Check `monitoring/prometheus.yml` has merchant-service configured

#### 3. Verify Metrics Exist in Prometheus

1. Open Prometheus: `http://localhost:9090`
2. Go to **Graph** tab
3. Try these queries:

```promql
# Check if metrics exist
kyb_fallback_total

# Check fallback rate
kyb_fallback_rate_percent

# Check requests
kyb_requests_total
```

**If queries return "No data":**
- Metrics haven't been recorded yet
- Service might not be running
- Prometheus might not be scraping

#### 4. Check Dashboard Data Source Reference

The dashboard uses `${DS_PROMETHEUS}` variable. When importing:

1. Select your Prometheus data source from dropdown
2. This sets the `DS_PROMETHEUS` variable
3. All panels will use this data source

**If panels still show "No data":**
- Manually edit each panel
- Go to **Query** tab
- Verify data source is selected
- Check query syntax

#### 5. Verify Time Range

1. Check dashboard time range (top right)
2. Set to **Last 1 hour** or **Last 5 minutes**
3. If metrics were just created, use a shorter range

#### 6. Check Metric Names Match

Verify metric names in code match dashboard queries:

**In code** (`internal/metrics/fallback_metrics.go`):
- `kyb_fallback_total`
- `kyb_fallback_rate_percent`
- `kyb_fallback_duration_seconds`
- `kyb_requests_total`
- `kyb_fallback_by_category_total`
- `kyb_fallback_by_source_total`

**In dashboard** (`monitoring/grafana-fallback-dashboard.json`):
- Same names should be used

---

## Step-by-Step Verification Checklist

### ✅ Service Running
- [ ] Merchant service is running on port 8082
- [ ] Health endpoint responds: `curl http://localhost:8082/health`
- [ ] Metrics endpoint responds: `curl http://localhost:8082/metrics`

### ✅ Prometheus Configuration
- [ ] Prometheus is running: `docker ps | grep prometheus`
- [ ] `monitoring/prometheus.yml` includes merchant-service job
- [ ] Prometheus target shows UP: `http://localhost:9090/targets`

### ✅ Metrics Being Scraped
- [ ] Prometheus can query metrics: `http://localhost:9090/graph`
- [ ] Query `kyb_fallback_total` returns results
- [ ] Metrics have been recorded (not all zeros)

### ✅ Grafana Configuration
- [ ] Prometheus data source configured and tested
- [ ] Dashboard imported successfully
- [ ] Dashboard time range is appropriate
- [ ] Data source variable `${DS_PROMETHEUS}` is set

### ✅ Dashboard Panels
- [ ] Panels show data (not "No data")
- [ ] Queries match metric names
- [ ] Time range includes when metrics were recorded

---

## Quick Fix Commands

```bash
# 1. Start monitoring stack
docker-compose -f docker-compose.monitoring.yml up -d

# 2. Check Prometheus targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | select(.labels.job=="merchant-service")'

# 3. Check if metrics endpoint works
curl http://localhost:8082/metrics | head -20

# 4. Query Prometheus directly
curl 'http://localhost:9090/api/v1/query?query=kyb_fallback_total'

# 5. Restart Prometheus to pick up config changes
docker-compose -f docker-compose.monitoring.yml restart prometheus
```

---

## Common Issues & Fixes

### Issue: "No data" in all panels

**Cause**: Metrics haven't been recorded yet or service isn't running

**Fix**:
1. Ensure merchant service is running
2. Make some API requests to generate metrics
3. Wait a few minutes for Prometheus to scrape
4. Check Prometheus directly: `http://localhost:9090/graph?g0.expr=kyb_fallback_total`

### Issue: "Query error" in panels

**Cause**: Invalid PromQL query or metric doesn't exist

**Fix**:
1. Test query in Prometheus first: `http://localhost:9090/graph`
2. Verify metric name is correct
3. Check for typos in dashboard JSON

### Issue: Service shows DOWN in Prometheus

**Cause**: Can't connect to service or wrong port

**Fix**:
1. Verify service is running: `curl http://localhost:8082/health`
2. Check firewall/network settings
3. Verify port in `prometheus.yml` matches service port
4. Check if service is listening on `0.0.0.0` not just `localhost`

### Issue: Metrics show 0 values

**Cause**: No fallback events have occurred yet

**Fix**:
1. This is normal if no fallbacks have happened
2. To test, simulate a failure (stop Supabase, make invalid requests)
3. Or wait for real fallback scenarios

---

## Testing Metrics Generation

To verify metrics are working, you can:

1. **Make a request that triggers fallback**:
   ```bash
   # Request non-existent merchant (might trigger fallback in dev)
   curl http://localhost:8082/api/v1/merchants/test-123
   ```

2. **Check metrics immediately**:
   ```bash
   curl http://localhost:8082/metrics | grep kyb_fallback
   ```

3. **Wait for Prometheus scrape** (default: 15s):
   ```bash
   # Query Prometheus
   curl 'http://localhost:9090/api/v1/query?query=kyb_fallback_total'
   ```

4. **Check Grafana dashboard** - should show data after scrape

---

## Still Having Issues?

1. **Check logs**:
   ```bash
   # Merchant service logs
   docker logs merchant-service
   
   # Prometheus logs
   docker logs kyb-prometheus
   
   # Grafana logs
   docker logs kyb-grafana
   ```

2. **Verify all services are running**:
   ```bash
   docker ps | grep -E "prometheus|grafana|merchant"
   ```

3. **Check network connectivity**:
   ```bash
   # From Prometheus container
   docker exec kyb-prometheus wget -O- http://host.docker.internal:8082/metrics
   ```

4. **Review configuration files**:
   - `monitoring/prometheus.yml` - scrape config
   - `services/merchant-service/cmd/main.go` - metrics endpoint
   - `internal/metrics/fallback_metrics.go` - metric definitions

