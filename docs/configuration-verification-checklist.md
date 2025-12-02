# Configuration Verification Checklist

## Step 1: Verify Environment Variables in Railway

### Required Variables for Website Scraping Optimizations

Go to **Railway Dashboard → Classification Service → Variables** and verify:

#### ✅ Fast-Path Scraping
```bash
ENABLE_FAST_PATH_SCRAPING=true
```

#### ✅ Parallel Processing
```bash
CLASSIFICATION_MAX_CONCURRENT_PAGES=3
```

#### ✅ Crawl Delays
```bash
CLASSIFICATION_CRAWL_DELAY_MS=500
```

#### ✅ Fast-Path Limits
```bash
CLASSIFICATION_FAST_PATH_MAX_PAGES=8
```

#### ✅ Timeout Configuration
```bash
CLASSIFICATION_WEBSITE_SCRAPING_TIMEOUT=5s
```

### Recommended Variables (Already Set)

#### ✅ Redis Caching
```bash
REDIS_ENABLED=true
REDIS_URL=${{Redis.REDIS_URL}}
ENABLE_WEBSITE_CONTENT_CACHE=true
```

#### ✅ General Caching
```bash
CACHE_ENABLED=true
CACHE_TTL=5m
```

---

## Step 2: Verify Configuration Script

Run the verification script locally (checks environment variables):

```bash
./scripts/verify-classification-config.sh
```

**Note**: This checks local environment. For Railway, verify in the dashboard.

---

## Step 3: Check Service Logs for Configuration

After deployment, check logs for:

```
✅ Fast-path scraping enabled
✅ Parallel processing enabled (max concurrent: 3)
✅ Website content cache initialized
```

---

## Quick Verification Commands

### Check if variables are set (Railway CLI)

```bash
railway variables --service classification-service
```

### Set missing variables (Railway CLI)

```bash
railway variables set ENABLE_FAST_PATH_SCRAPING=true --service classification-service
railway variables set CLASSIFICATION_MAX_CONCURRENT_PAGES=3 --service classification-service
railway variables set CLASSIFICATION_CRAWL_DELAY_MS=500 --service classification-service
railway variables set CLASSIFICATION_FAST_PATH_MAX_PAGES=8 --service classification-service
railway variables set CLASSIFICATION_WEBSITE_SCRAPING_TIMEOUT=5s --service classification-service
```

---

## Expected Configuration Values

| Variable | Expected Value | Purpose |
|----------|---------------|---------|
| `ENABLE_FAST_PATH_SCRAPING` | `true` | Enable fast-path mode |
| `CLASSIFICATION_MAX_CONCURRENT_PAGES` | `3` | Max parallel page requests |
| `CLASSIFICATION_CRAWL_DELAY_MS` | `500` | Delay between pages (ms) |
| `CLASSIFICATION_FAST_PATH_MAX_PAGES` | `8` | Max pages for fast-path |
| `CLASSIFICATION_WEBSITE_SCRAPING_TIMEOUT` | `5s` | Overall scraping timeout |

---

## Verification Checklist

- [ ] All required variables are set in Railway
- [ ] Variable values match expected values
- [ ] Service logs show configuration loaded
- [ ] No configuration errors in logs
- [ ] Fast-path mode is active
- [ ] Parallel processing is enabled
- [ ] Redis cache is configured

---

## Troubleshooting

### Variables Not Set

**Solution**: Add variables in Railway Dashboard → Classification Service → Variables

### Wrong Values

**Solution**: Update variable values to match expected values above

### Configuration Not Loading

**Solution**: 
1. Check service logs for errors
2. Verify variable names are exact (case-sensitive)
3. Redeploy service if needed

---

## Files

- **Verification Script**: `scripts/verify-classification-config.sh`
- **Checklist**: `docs/configuration-verification-checklist.md` (this document)

