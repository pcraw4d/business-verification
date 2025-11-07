# Grafana Fallback Metrics Dashboard Guide

## Overview

This guide explains how to access and use the Grafana dashboard for monitoring fallback data usage metrics in the KYB Platform.

---

## Prerequisites

1. **Docker and Docker Compose** installed
2. **Monitoring stack** running (Prometheus, Grafana)
3. **Services** configured to export metrics to Prometheus

---

## Quick Start

### 1. Start the Monitoring Stack

```bash
# Start Prometheus and Grafana
docker-compose -f docker-compose.monitoring.yml up -d

# Verify services are running
docker ps | grep -E "prometheus|grafana"
```

### 2. Access Grafana

1. **Open Grafana** in your browser:
   ```
   http://localhost:3000
   ```

2. **Login** with default credentials:
   - Username: `admin`
   - Password: `admin` (you'll be prompted to change it on first login)

### 3. Configure Prometheus Data Source

1. Go to **Configuration** → **Data Sources** (gear icon in left sidebar)
2. Click **Add data source**
3. Select **Prometheus**
4. Configure:
   - **URL**: `http://prometheus:9090` (if using Docker Compose) or `http://localhost:9090`
   - Click **Save & Test**
   - You should see "Data source is working"

### 4. Import the Fallback Metrics Dashboard

#### Option A: Import from JSON File

1. Go to **Dashboards** → **Import** (plus icon → Import)
2. Click **Upload JSON file**
3. Select `monitoring/grafana-fallback-dashboard.json`
4. Click **Load**
5. Select the Prometheus data source
6. Click **Import**

#### Option B: Import via Dashboard ID (if published)

1. Go to **Dashboards** → **Import**
2. Enter Dashboard ID: `[dashboard-id]` (if published to Grafana.com)
3. Click **Load**

### 5. Access the Dashboard

1. Go to **Dashboards** → **Browse**
2. Find **"KYB Platform - Fallback Metrics Dashboard"**
3. Click to open

---

## Dashboard Panels

The dashboard includes the following panels:

### 1. Fallback Usage Rate (%)
- **Metric**: `kyb_fallback_rate_percent`
- **Description**: Shows the percentage of requests using fallback data per service
- **Alert**: Triggers when rate > 10%
- **Location**: Top left

### 2. Total Fallback Events
- **Metric**: `rate(kyb_fallback_total[5m])`
- **Description**: Rate of fallback events per second
- **Location**: Top right

### 3. Fallback by Category
- **Metric**: `kyb_fallback_by_category_total`
- **Description**: Pie chart showing fallback usage by category:
  - `database_fallback`
  - `api_fallback`
  - `missing_record`
  - `incomplete_feature`
- **Location**: Middle left

### 4. Fallback by Source
- **Metric**: `kyb_fallback_by_source_total`
- **Description**: Pie chart showing fallback usage by source:
  - `supabase`
  - `risk-api`
  - `analytics-api`
- **Location**: Middle center

### 5. Fallback Duration
- **Metric**: `kyb_fallback_duration_seconds`
- **Description**: P50 and P95 latency for fallback operations
- **Location**: Middle right

### 6. Total Requests vs Fallbacks
- **Metrics**: 
  - `rate(kyb_requests_total[5m])`
  - `rate(kyb_fallback_total[5m])`
- **Description**: Comparison of total requests vs fallback usage
- **Location**: Bottom (full width)

### 7. Current Fallback Rate (Stat Panels)
- **Metrics**: `kyb_fallback_rate_percent` by service
- **Description**: Current fallback rate for each service with color coding:
  - Green: < 5%
  - Yellow: 5-10%
  - Red: > 10%
- **Location**: Bottom row

---

## Prometheus Metrics

The following metrics are exported by the fallback metrics system:

### Counters

- `kyb_fallback_total{service, category, source}` - Total fallback events
- `kyb_fallback_by_category_total{service, category}` - Fallbacks by category
- `kyb_fallback_by_source_total{service, source}` - Fallbacks by source
- `kyb_requests_total{service}` - Total requests (fallback + non-fallback)

### Gauges

- `kyb_fallback_rate_percent{service}` - Current fallback rate percentage

### Histograms

- `kyb_fallback_duration_seconds{service, category}` - Fallback operation duration

---

## Verifying Metrics are Available

### 1. Check Prometheus Targets

1. Open Prometheus: `http://localhost:9090`
2. Go to **Status** → **Targets**
3. Verify your service targets are **UP**

### 2. Query Metrics in Prometheus

1. Go to **Graph** tab in Prometheus
2. Try these queries:
   ```promql
   # Check if metrics are available
   kyb_fallback_total
   
   # Check fallback rate
   kyb_fallback_rate_percent
   
   # Check total requests
   kyb_requests_total
   ```

### 3. Check Service Metrics Endpoint

Verify your service is exposing metrics:

```bash
# For merchant service (adjust port as needed)
curl http://localhost:8082/metrics | grep kyb_fallback

# Should see metrics like:
# kyb_fallback_total{service="merchant-service",category="database_fallback",source="supabase"} 5
# kyb_fallback_rate_percent{service="merchant-service"} 2.5
```

---

## Troubleshooting

### Dashboard Shows "No Data"

1. **Check Prometheus Data Source**:
   - Verify Prometheus is running: `docker ps | grep prometheus`
   - Test connection in Grafana: Configuration → Data Sources → Prometheus → Test

2. **Check Metrics are Being Exported**:
   ```bash
   # Check if service is exposing /metrics endpoint
   curl http://localhost:8082/metrics
   ```

3. **Check Prometheus is Scraping**:
   - Go to Prometheus: `http://localhost:9090/targets`
   - Verify targets are UP

4. **Check Metric Names**:
   - Verify metric names match: `kyb_fallback_*`
   - Check labels match dashboard queries

### Metrics Not Appearing

1. **Verify Service is Recording Metrics**:
   - Check service logs for "Fallback usage recorded"
   - Verify `fallbackMetrics` is initialized in handlers

2. **Check Prometheus Configuration**:
   - Verify `monitoring/prometheus.yml` includes your service
   - Check scrape interval and timeout settings

3. **Restart Services**:
   ```bash
   # Restart Prometheus to pick up config changes
   docker-compose -f docker-compose.monitoring.yml restart prometheus
   ```

---

## Customizing the Dashboard

### Adding New Panels

1. Click **Add panel** (top right)
2. Select visualization type
3. Enter PromQL query
4. Configure panel settings
5. Click **Apply**

### Modifying Existing Panels

1. Click panel title → **Edit**
2. Modify query or settings
3. Click **Apply**

### Creating Alerts

1. Edit a panel
2. Go to **Alert** tab
3. Click **Create Alert**
4. Configure conditions:
   - **Condition**: `WHEN avg() OF query(A, 5m, now) IS ABOVE 10`
   - **For**: `5m`
5. Configure notifications
6. Click **Save**

---

## Production Deployment

### Railway/Cloud Deployment

If deploying to Railway or other cloud platforms:

1. **Set up Prometheus**:
   - Deploy Prometheus service
   - Configure scrape targets to point to your services

2. **Set up Grafana**:
   - Deploy Grafana service
   - Configure data source to point to Prometheus URL
   - Import dashboard JSON

3. **Service URLs**:
   - Update `monitoring/prometheus.yml` with production service URLs
   - Update Grafana data source URL

### Environment Variables

Set these in your deployment:

```bash
# Prometheus
PROMETHEUS_URL=http://prometheus:9090

# Grafana
GRAFANA_URL=http://grafana:3000
GRAFANA_ADMIN_PASSWORD=your-secure-password
```

---

## Useful PromQL Queries

### Fallback Rate by Service
```promql
kyb_fallback_rate_percent
```

### Fallback Events per Second
```promql
sum(rate(kyb_fallback_total[5m])) by (service)
```

### Fallback Rate Over Time
```promql
avg_over_time(kyb_fallback_rate_percent[5m])
```

### Top Fallback Categories
```promql
topk(5, sum(kyb_fallback_by_category_total) by (category))
```

### Fallback Duration P95
```promql
histogram_quantile(0.95, sum(rate(kyb_fallback_duration_seconds_bucket[5m])) by (service, le))
```

---

## Next Steps

1. **Set up Alerts**: Configure alerts for high fallback rates
2. **Create Additional Dashboards**: Create service-specific dashboards
3. **Export Dashboard**: Export and version control dashboard JSON
4. **Share Dashboard**: Share dashboard with team members

---

## Support

For issues or questions:
- Check service logs for metric recording
- Verify Prometheus targets are UP
- Test PromQL queries directly in Prometheus
- Review `docs/PLACEHOLDER-IMPLEMENTATION-SUMMARY.md` for implementation details

