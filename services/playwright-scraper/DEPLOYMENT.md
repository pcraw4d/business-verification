# Playwright Scraper Service - Deployment Guide

## Railway Deployment Steps

1. **Create New Service in Railway:**
   - Go to Railway dashboard
   - Click "New Project" or add to existing project
   - Select "Deploy from GitHub repo"
   - Choose your repository

2. **Configure Service:**
   - Set root directory: `services/playwright-scraper`
   - Railway will auto-detect the Dockerfile
   - Ensure the service has at least 512MB memory

3. **Get Service URL:**
   - After deployment, note the service URL (e.g., `https://playwright-scraper-production.up.railway.app`)
   - This will be used as `PLAYWRIGHT_SERVICE_URL`

4. **Update Classification Service Environment:**
   - Go to your classification service in Railway
   - Add environment variable: `PLAYWRIGHT_SERVICE_URL`
   - Set value to the Playwright service URL from step 3

5. **Verify Deployment:**
   - Test health endpoint: `curl https://your-playwright-service.railway.app/health`
   - Should return: `{"status":"ok","service":"playwright-scraper"}`

## Testing the Service

```bash
# Test health endpoint
curl https://your-playwright-service.railway.app/health

# Test scrape endpoint
curl -X POST https://your-playwright-service.railway.app/scrape \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

## Troubleshooting

- **Service crashes:** Check Railway logs, ensure memory is at least 512MB
- **Timeout errors:** Increase timeout in index.js if needed
- **Connection refused:** Verify service URL is correct and service is running

