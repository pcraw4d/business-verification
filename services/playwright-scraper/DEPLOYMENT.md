# Playwright Scraper Service - Deployment Guide

## Overview

The Playwright Scraper Service is an optimized Node.js service for scraping JavaScript-heavy websites using Playwright. It features browser pooling, concurrency limiting, and request queuing to ensure efficient resource usage and high throughput.

## Key Features

- **Browser Pooling**: Reuses browser instances to eliminate launch overhead
- **Concurrency Limiting**: Limits concurrent requests to prevent resource exhaustion
- **Request Queuing**: Queues requests with timeout to ensure fair handling
- **Automatic Recovery**: Automatically recovers from dead browser instances
- **Comprehensive Monitoring**: Health checks with pool and queue metrics

## Railway Deployment Steps

1. **Create New Service in Railway:**
   - Go to Railway dashboard
   - Click "New Project" or add to existing project
   - Select "Deploy from GitHub repo"
   - Choose your repository

2. **Configure Service:**
   - Set root directory: `services/playwright-scraper`
   - Railway will auto-detect the Dockerfile
   - **Recommended**: Set memory to at least 2GB (previously 512MB)
   - Set CPU to at least 1.0 core

3. **Configure Environment Variables:**
   - `BROWSER_POOL_SIZE`: Number of browsers in pool (default: 3, recommended: 3-5)
   - `MAX_CONCURRENT_REQUESTS`: Max concurrent requests (default: 8, recommended: 8-12)
   - `QUEUE_TIMEOUT_MS`: Queue timeout in milliseconds (default: 25000)
   - `SCRAPE_TIMEOUT_MS`: Scrape timeout in milliseconds (default: 15000)
   - `BROWSER_RECOVERY_ENABLED`: Enable automatic browser recovery (default: true)
   - `PORT`: Server port (default: 3000)

4. **Get Service URL:**
   - After deployment, note the service URL (e.g., `https://playwright-scraper-production.up.railway.app`)
   - This will be used as `PLAYWRIGHT_SERVICE_URL`

5. **Update Classification Service Environment:**
   - Go to your classification service in Railway
   - Add environment variable: `PLAYWRIGHT_SERVICE_URL`
   - Set value to the Playwright service URL from step 4

6. **Verify Deployment:**
   - Test health endpoint: `curl https://your-playwright-service.railway.app/health`
   - Should return detailed health status including pool and queue metrics

## Configuration

### Environment Variables

| Variable | Default | Description | Recommended |
|----------|---------|-------------|-------------|
| `BROWSER_POOL_SIZE` | 3 | Number of browsers in pool | 3-5 |
| `MAX_CONCURRENT_REQUESTS` | 8 | Max concurrent requests | 8-12 |
| `QUEUE_TIMEOUT_MS` | 25000 | Queue timeout in milliseconds | 20000-30000 |
| `SCRAPE_TIMEOUT_MS` | 15000 | Scrape timeout in milliseconds | 10000-20000 |
| `BROWSER_RECOVERY_ENABLED` | true | Enable automatic browser recovery | true |
| `PORT` | 3000 | Server port | 3000 |

### Performance Tuning

**For High Load (10+ concurrent requests):**
```bash
BROWSER_POOL_SIZE=5
MAX_CONCURRENT_REQUESTS=12
QUEUE_TIMEOUT_MS=30000
```

**For Low Resource Environments:**
```bash
BROWSER_POOL_SIZE=2
MAX_CONCURRENT_REQUESTS=4
QUEUE_TIMEOUT_MS=20000
```

**For Maximum Throughput:**
```bash
BROWSER_POOL_SIZE=5
MAX_CONCURRENT_REQUESTS=15
QUEUE_TIMEOUT_MS=25000
SCRAPE_TIMEOUT_MS=10000
```

## Resource Requirements

### Memory

- **Minimum**: 512MB (not recommended)
- **Recommended**: 2GB
- **High Load**: 4GB

Each browser instance uses approximately 200-400MB of memory. With a pool of 3 browsers, expect 600MB-1.2GB base usage, plus overhead for requests.

### CPU

- **Minimum**: 0.5 cores
- **Recommended**: 1.0 core
- **High Load**: 2.0 cores

### Local Development (Docker Compose)

Resource limits are configured in `docker-compose.local.yml`:
- Memory limit: 2GB
- CPU limit: 1.0 core
- Process limit: 64
- File descriptor limit: 1024

## API Endpoints

### GET /health

Health check endpoint with detailed metrics.

**Response:**
```json
{
  "status": "ok",
  "service": "playwright-scraper",
  "browserPool": {
    "total": 3,
    "available": 2,
    "inUse": 1,
    "dead": 0,
    "utilization": 33.33
  },
  "queue": {
    "size": 2,
    "pending": 1,
    "maxConcurrent": 8
  },
  "metrics": {
    "totalRequests": 100,
    "completedRequests": 95,
    "failedRequests": 5,
    "timeoutRequests": 2,
    "successRate": 95.0
  },
  "memory": {
    "heapUsedMB": 512,
    "heapTotalMB": 1024,
    "rssMB": 1536
  }
}
```

**Status Values:**
- `ok`: Service is healthy
- `degraded`: Queue is >80% full
- `unhealthy`: All browsers are dead

### POST /scrape

Scrapes a website and returns the full HTML content.

**Request Body:**
```json
{
  "url": "https://example.com"
}
```

**Success Response (200):**
```json
{
  "html": "<html>...</html>",
  "success": true,
  "requestId": "req_1234567890_abc123",
  "metrics": {
    "scrapeDurationMs": 3500,
    "totalDurationMs": 4200,
    "queueWaitTimeMs": 700
  }
}
```

**Error Responses:**

- **400 Bad Request**: Invalid request (missing URL)
- **408 Request Timeout**: Scrape timeout exceeded
- **429 Too Many Requests**: Queue timeout exceeded
- **503 Service Unavailable**: Browser error or service overload

## Testing the Service

```bash
# Test health endpoint
curl https://your-playwright-service.railway.app/health

# Test scrape endpoint
curl -X POST https://your-playwright-service.railway.app/scrape \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

## Monitoring

### Key Metrics to Monitor

1. **Browser Pool Utilization**: Should be <80% under normal load
2. **Queue Size**: Should be <80% of max concurrent requests
3. **Success Rate**: Should be >95%
4. **Average Queue Wait Time**: Should be <5 seconds
5. **Memory Usage**: Should be <2GB for 3-browser pool

### Health Check Monitoring

Monitor the `/health` endpoint:
- Status should be `ok` under normal conditions
- `degraded` indicates high load (consider scaling)
- `unhealthy` indicates all browsers are dead (check logs)

## Troubleshooting

### Service Crashes

**Problem**: Service crashes or becomes unresponsive

**Solutions**:
- Check Railway logs for error messages
- Ensure memory is at least 2GB (not 512MB)
- Verify CPU allocation is at least 1.0 core
- Check for browser launch failures in logs
- Reduce `BROWSER_POOL_SIZE` if memory is limited

### Queue Timeout Errors

**Problem**: Requests return 429 (Too Many Requests)

**Solutions**:
- Increase `QUEUE_TIMEOUT_MS` (default: 25000ms)
- Increase `MAX_CONCURRENT_REQUESTS` (default: 8)
- Increase `BROWSER_POOL_SIZE` (default: 3)
- Check if upstream service is sending too many requests
- Monitor queue size in health endpoint

### Browser Pool Exhaustion

**Problem**: All browsers are dead or unavailable

**Solutions**:
- Check logs for browser crash reasons
- Verify `BROWSER_RECOVERY_ENABLED=true` (default)
- Increase `BROWSER_POOL_SIZE` for redundancy
- Check memory limits (browsers may be killed by OOM)
- Verify Docker/container resource limits

### High Memory Usage

**Problem**: Memory usage exceeds limits

**Solutions**:
- Reduce `BROWSER_POOL_SIZE` (each browser uses 200-400MB)
- Reduce `MAX_CONCURRENT_REQUESTS` (fewer concurrent = less memory)
- Increase container memory limit
- Check for memory leaks (monitor heap usage over time)

### Slow Response Times

**Problem**: Requests take too long

**Solutions**:
- Check queue wait time in response metrics
- If queue wait is high: increase `MAX_CONCURRENT_REQUESTS`
- If scrape time is high: check target website performance
- Reduce `SCRAPE_TIMEOUT_MS` if websites are consistently slow
- Consider increasing `BROWSER_POOL_SIZE` for better parallelism

### Connection Refused

**Problem**: Cannot connect to service

**Solutions**:
- Verify service URL is correct
- Check service is running in Railway dashboard
- Verify health endpoint is accessible
- Check network/firewall settings
- Verify `PORT` environment variable matches Railway configuration

## Performance Characteristics

### Throughput

- **Target**: Handle 10+ concurrent requests
- **Peak**: Can handle 15-20 concurrent requests with proper configuration
- **Bottleneck**: Browser pool size and memory limits

### Latency

- **Target**: <25s total (queue + scrape)
- **Queue Wait**: Typically <5s under normal load
- **Scrape Time**: Typically 2-10s depending on website
- **Browser Launch**: Eliminated (browsers are pooled)

### Resource Usage

- **Memory**: ~600MB-1.2GB base (3 browsers) + ~100MB per concurrent request
- **CPU**: ~0.5-1.0 cores under normal load, spikes to 1.5-2.0 cores under high load
- **Network**: Depends on website content size

## Best Practices

1. **Monitor Health Endpoint**: Regularly check `/health` for service status
2. **Tune Configuration**: Adjust environment variables based on actual load
3. **Set Appropriate Limits**: Don't set limits too high (causes resource exhaustion)
4. **Monitor Metrics**: Track success rate, queue size, and memory usage
5. **Scale Horizontally**: For very high load, deploy multiple instances behind a load balancer
6. **Use Resource Limits**: Always set memory and CPU limits in production

## Local Development

For local development, resource limits are configured in `docker-compose.local.yml`:

```yaml
deploy:
  resources:
    limits:
      memory: 2G
      cpus: '1.0'
    reservations:
      memory: 512M
      cpus: '0.5'
ulimits:
  nproc: 64
  nofile: 1024
```

These limits prevent the Playwright service from consuming all available resources on your local machine.

## Production Considerations

1. **Memory**: Allocate at least 2GB, preferably 4GB for high load
2. **CPU**: Allocate at least 1.0 core, 2.0 cores for high load
3. **Monitoring**: Set up alerts for health check failures
4. **Scaling**: Consider horizontal scaling for very high load
5. **Logging**: Enable structured logging for production debugging
