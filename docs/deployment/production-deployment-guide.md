# Production Deployment Guide - DistilBART Quantization

**Date**: November 26, 2025  
**Feature**: DistilBART Quantization Enabled by Default  
**Status**: ✅ Ready for Deployment

---

## Quick Start

### 1. Deploy Python ML Service with Docker

```bash
cd python_ml_service

# Build and start production service
docker-compose up -d --build python-ml-service

# Check service status
docker-compose ps

# View logs
docker-compose logs -f python-ml-service
```

### 2. Verify Quantization is Enabled

Look for this log message:
```
✅ DistilBART classifier initialized with quantization: True
```

---

## Deployment Steps

### Step 1: Build Docker Image

```bash
cd python_ml_service

# Build production image
docker build -t python-ml-service:latest --target production .

# Verify image was created
docker images | grep python-ml-service
```

### Step 2: Start Service

```bash
# Start with docker-compose (recommended)
docker-compose up -d python-ml-service

# Or start manually
docker run -d \
  --name python-ml-service \
  -p 8000:8000 \
  -e USE_QUANTIZATION=true \
  -e QUANTIZATION_DTYPE=qint8 \
  -e MODEL_SAVE_PATH=/app/models/distilbart \
  -e QUANTIZED_MODELS_PATH=/app/models/quantized \
  -v $(pwd)/models:/app/models \
  python-ml-service:latest
```

### Step 3: Monitor Deployment

```bash
# Watch logs in real-time
docker-compose logs -f python-ml-service

# Or with docker
docker logs -f python-ml-service
```

**Expected Log Output**:
```
🚀 DistilBART Business Classifier initializing on cpu
📥 Loading DistilBERT-MNLI for classification...
✅ DistilBERT classification model loaded
📥 Loading DistilBART for summarization...
✅ DistilBART summarization model loaded
✅ Summarization model quantized
✅ DistilBART classifier initialized with quantization: True
🚀 Starting Python ML Service...
```

### Step 4: Verify Service Health

```bash
# Check health endpoint
curl http://localhost:8000/health

# Check model info
curl http://localhost:8000/model-info | jq

# Expected response should show:
# "quantization_enabled": true
```

---

## Environment Variables

The following environment variables are set in `docker-compose.yml`:

```yaml
USE_QUANTIZATION=true          # Enable quantization (default: true)
QUANTIZATION_DTYPE=qint8      # Quantization dtype
MODEL_SAVE_PATH=/app/models/distilbart
QUANTIZED_MODELS_PATH=/app/models/quantized
```

**To disable quantization** (if needed):
```bash
# Set in docker-compose.yml or as environment variable
USE_QUANTIZATION=false
```

---

## Monitoring

### Check Quantization Status

```bash
# Via API
curl http://localhost:8000/model-info | jq '.quantization_enabled'

# Via logs
docker-compose logs python-ml-service | grep quantization
```

### Monitor Performance

```bash
# Test classification endpoint
time curl -X POST http://localhost:8000/classify \
  -H "Content-Type: application/json" \
  -d '{"content": "Technology company", "max_length": 512}'

# Expected: <200ms response time
```

### Check Resource Usage

```bash
# Monitor container resources
docker stats python-ml-service

# Expected:
# - Memory: 1-2GB (67% reduction from 2-3GB)
# - CPU: Varies based on load
```

---

## Troubleshooting

### Issue: Quantization Not Enabled

**Check**:
```bash
# Verify environment variable
docker-compose exec python-ml-service env | grep USE_QUANTIZATION

# Check logs
docker-compose logs python-ml-service | grep quantization
```

**Fix**: Ensure `USE_QUANTIZATION=true` is set in environment

### Issue: Models Not Loading

**Check**:
```bash
# Verify model directories exist
docker-compose exec python-ml-service ls -la /app/models/

# Check disk space
docker-compose exec python-ml-service df -h
```

**Fix**: Models download automatically on first use. Ensure internet connectivity.

### Issue: Slow Performance

**Check**:
```bash
# Verify quantization is actually enabled
curl http://localhost:8000/model-info | jq '.quantization_enabled'

# Check system resources
docker stats python-ml-service
```

**Fix**: If quantization is disabled, check logs for errors and ensure PyTorch >= 2.6.0

---

## Rollback Procedure

If quantization causes issues:

### Option 1: Disable via Environment Variable

```bash
# Update docker-compose.yml
USE_QUANTIZATION=false

# Restart service
docker-compose up -d python-ml-service
```

### Option 2: Revert to Previous Version

```bash
# Stop current service
docker-compose down

# Checkout previous commit
git checkout <previous-commit-hash>

# Rebuild and restart
docker-compose up -d --build python-ml-service
```

---

## Expected Performance Metrics

After deployment, monitor these metrics:

| Metric | Target | How to Check |
|--------|--------|--------------|
| **Inference Time** | <200ms | API response time |
| **Model Size** | ~137MB | `du -sh models/quantized` |
| **Memory Usage** | 1-2GB | `docker stats` |
| **Confidence Scores** | >85% | API responses |
| **Error Rate** | <1% | Log monitoring |

---

## Post-Deployment Checklist

- [ ] Service starts successfully
- [ ] Logs show `quantization: True`
- [ ] Health endpoint returns 200
- [ ] Model info shows `quantization_enabled: true`
- [ ] Classification endpoint responds <200ms
- [ ] Memory usage is 1-2GB
- [ ] No errors in logs
- [ ] Integration tests pass
- [ ] Monitor for 24 hours

---

## Production Monitoring

### Key Log Messages to Watch

**Success Indicators**:
```
✅ DistilBART classifier initialized with quantization: True
✅ Classification completed in X.XXXs (quantized: True)
```

**Warning Indicators**:
```
⚠️ Quantization failed, using original models
⚠️ Could not quantize summarization model
```

**Error Indicators**:
```
❌ Failed to load DistilBART models
❌ Quantization error
```

### Metrics to Track

1. **Inference Time**: Should average <200ms
2. **Memory Usage**: Should be 1-2GB (not 2-3GB)
3. **Model Size**: Should be ~137MB (not 550MB)
4. **Accuracy**: Confidence scores should be >85%
5. **Error Rate**: Should be <1%

---

## Next Steps

After successful deployment:

1. ✅ Monitor logs for 24 hours
2. ✅ Track performance metrics
3. ✅ Validate accuracy maintained
4. ✅ Document any issues
5. ✅ Update monitoring dashboards

---

**Deployment Date**: November 26, 2025  
**Deployed By**: [Your Name]  
**Status**: Ready for Production
