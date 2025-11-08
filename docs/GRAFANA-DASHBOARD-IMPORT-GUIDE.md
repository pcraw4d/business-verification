# Grafana Dashboard Import Guide

## Quick Import Steps

1. **Open Grafana**: Navigate to `http://localhost:3000`
2. **Go to Dashboards** → **Import** (plus icon → Import)
3. **Upload JSON**: Click "Upload JSON file" and select `monitoring/grafana-fallback-dashboard.json`
4. **Select Data Source**: When prompted, select your Prometheus data source
5. **Import**: Click "Import"

## Important: Data Source Configuration

The dashboard uses `${DS_PROMETHEUS}` variable which Grafana will replace during import. If you see "No data" after importing:

### Option 1: Set Data Source During Import (Recommended)
- When importing, Grafana will prompt you to select a data source
- Select your Prometheus data source
- Grafana will automatically replace `${DS_PROMETHEUS}` with the correct UID

### Option 2: Manually Update Data Source
1. Edit the dashboard
2. Click on any panel → **Edit**
3. In the query editor, click the data source dropdown
4. Select your Prometheus data source
5. Click **Save dashboard**
6. All panels will update automatically

### Option 3: Use Direct Data Source Reference
If the variable doesn't work, you can manually replace all datasource references:

1. Export the dashboard JSON
2. Find and replace: `"uid": "${DS_PROMETHEUS}"` with `"uid": "YOUR_PROMETHEUS_UID"`
3. To find your Prometheus UID:
   - Go to **Configuration** → **Data Sources**
   - Click on Prometheus
   - The UID is shown in the URL or in the data source settings

## Why You Might See "No Data"

### Reason 1: Metrics Don't Exist Yet
**Most Common**: If no fallback events have occurred, the metrics won't exist in Prometheus.

**Solution**: 
- This is normal! Healthy services don't generate fallback metrics
- Metrics will appear automatically when fallback events occur
- The dashboard queries are configured to show `0` when metrics don't exist

### Reason 2: Wrong Time Range
**Solution**: 
- Check the time range selector (top right)
- Set to "Last 1 hour" or "Last 6 hours"
- Metrics are only available after they're created

### Reason 3: Data Source Not Configured
**Solution**:
- Verify Prometheus data source is configured
- Test the connection: **Configuration** → **Data Sources** → **Save & Test**
- Should show "Data source is working"

### Reason 4: Services Not Scraping
**Solution**:
- Check Prometheus targets: `http://localhost:9090/targets`
- Verify services show as **UP** (green)
- If DOWN, check service logs and network connectivity

## Testing the Dashboard

### Step 1: Verify Metrics Exist
```bash
# Check if metrics are being exported
curl https://merchant-service-production.up.railway.app/metrics | grep kyb_fallback
```

### Step 2: Test Queries in Prometheus
1. Open Prometheus: `http://localhost:9090`
2. Go to **Graph** tab
3. Try these queries:
   ```
   kyb_fallback_rate_percent
   kyb_fallback_total
   kyb_requests_total
   ```
4. If queries return data, the dashboard should work

### Step 3: Generate Test Data (Optional)
To see data immediately, you can trigger a fallback:
1. Temporarily break Supabase connection
2. Make API requests to merchant service
3. Check metrics endpoint for new data
4. Refresh Grafana dashboard

## Dashboard Panels Explained

1. **Fallback Usage Rate (%)**: Percentage of requests using fallback data
2. **Total Fallback Events**: Rate of fallback events per second
3. **Fallback by Category**: Breakdown by category (database, API, etc.)
4. **Fallback by Source**: Breakdown by source (supabase, external API, etc.)
5. **Fallback Duration**: P50 and P95 duration of fallback operations
6. **Request Comparison**: Total requests vs. fallback requests
7. **Service-Specific Panels**: Individual service fallback rates

## Troubleshooting Checklist

- [ ] Prometheus data source is configured and working
- [ ] Prometheus is scraping services (check `/targets`)
- [ ] Services are exporting metrics (check `/metrics` endpoint)
- [ ] Time range is set correctly (Last 1 hour)
- [ ] Dashboard queries are using correct metric names
- [ ] Data source variable is set during import

## Still Not Working?

1. **Check Grafana Logs**:
   ```bash
   docker logs grafana
   ```

2. **Check Prometheus Logs**:
   ```bash
   docker logs prometheus
   ```

3. **Test Query Directly**:
   - Open Prometheus UI
   - Try the exact query from the dashboard
   - If it works in Prometheus but not Grafana, it's a datasource issue

4. **Verify Metric Names**:
   - Check what metrics are actually exported
   - Compare with dashboard queries
   - Update queries if metric names differ

## Expected Behavior

- **Healthy Services**: Dashboard shows `0` for all metrics (no fallback events)
- **After Fallback**: Metrics appear automatically and update in real-time
- **Empty Panels**: If no metrics exist, panels show `0` or "No data" (this is expected)

