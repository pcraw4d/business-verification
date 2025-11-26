# Monitoring Quantization in Production

## Quick Commands

### Check Quantization Status

```bash
# Via API
curl http://localhost:8000/model-info | jq '.quantization_enabled'

# Via logs
docker-compose logs python-ml-service | grep "quantization: True"
```

### Monitor Logs in Real-Time

```bash
# Follow logs
docker-compose logs -f python-ml-service

# Filter for quantization-related messages
docker-compose logs -f python-ml-service | grep -E "(quantization|DistilBART|✅|❌)"
```

---

## Key Log Messages

### ✅ Success Indicators

Look for these messages in the logs:

```
✅ DistilBART classifier initialized with quantization: True
✅ DistilBERT classification model loaded
✅ DistilBART summarization model loaded
✅ Summarization model quantized
✅ Classification completed in X.XXXs (quantized: True)
```

### ⚠️ Warning Indicators

```
⚠️ Quantization failed, using original models
⚠️ Could not quantize summarization model
```

### ❌ Error Indicators

```
❌ Failed to load DistilBART models
❌ Quantization error
```

---

## Verification Steps

### 1. Check Service is Running

```bash
docker-compose ps python-ml-service
```

**Expected**: Status should be "Up"

### 2. Verify Quantization Enabled

```bash
curl -s http://localhost:8000/model-info | jq '.quantization_enabled'
```

**Expected**: `true`

### 3. Check Model Info

```bash
curl -s http://localhost:8000/model-info | jq
```

**Expected Output**:
```json
{
  "quantization_enabled": true,
  "model": "distilbart-quantized",
  "model_size_quantized": "~137MB",
  "size_reduction": "75%"
}
```

### 4. Test Classification

```bash
time curl -X POST http://localhost:8000/classify \
  -H "Content-Type: application/json" \
  -d '{"content": "Technology company", "max_length": 512}'
```

**Expected**: Response time < 200ms

---

## Continuous Monitoring

### Watch Logs for Quantization Confirmation

```bash
# Watch for the specific confirmation message
docker-compose logs -f python-ml-service | grep "quantization: True"
```

### Monitor Performance Metrics

```bash
# Check resource usage
docker stats python-ml-service

# Expected memory: 1-2GB (not 2-3GB)
```

### Set Up Alerts

Monitor for:
- Quantization disabled unexpectedly
- Inference time > 300ms
- Memory usage > 2.5GB
- Error rate > 1%

---

## Troubleshooting

### If Quantization is Not Enabled

1. **Check Environment Variables**:
   ```bash
   docker-compose exec python-ml-service env | grep USE_QUANTIZATION
   ```

2. **Check Logs for Errors**:
   ```bash
   docker-compose logs python-ml-service | grep -i error
   ```

3. **Verify PyTorch Version**:
   ```bash
   docker-compose exec python-ml-service python3 -c "import torch; print(torch.__version__)"
   ```
   Should be >= 2.6.0

### If Performance is Not Improved

1. **Verify Quantization is Actually Enabled**:
   ```bash
   curl -s http://localhost:8000/model-info | jq '.quantization_enabled'
   ```

2. **Check Model Size**:
   ```bash
   docker-compose exec python-ml-service du -sh /app/models/quantized
   ```
   Should be ~137MB

3. **Monitor Inference Time**:
   ```bash
   # Run multiple requests and measure
   for i in {1..10}; do
     time curl -s -X POST http://localhost:8000/classify \
       -H "Content-Type: application/json" \
       -d '{"content": "Test", "max_length": 512}' > /dev/null
   done
   ```

---

## Expected Metrics

| Metric | Target | How to Check |
|--------|--------|--------------|
| **Quantization Enabled** | `true` | `curl .../model-info \| jq .quantization_enabled` |
| **Inference Time** | <200ms | API response time |
| **Memory Usage** | 1-2GB | `docker stats` |
| **Model Size** | ~137MB | `du -sh models/quantized` |
| **Confidence Scores** | >85% | API responses |

---

## Log Monitoring Script

Save this as `monitor-quantization.sh`:

```bash
#!/bin/bash
# Monitor quantization status

SERVICE_URL="${1:-http://localhost:8000}"

echo "🔍 Monitoring Quantization Status..."
echo ""

# Check if service is up
if ! curl -f "$SERVICE_URL/health" > /dev/null 2>&1; then
    echo "❌ Service is not responding"
    exit 1
fi

# Get model info
MODEL_INFO=$(curl -s "$SERVICE_URL/model-info")

# Check quantization
QUANT_ENABLED=$(echo "$MODEL_INFO" | jq -r '.quantization_enabled // false')

if [ "$QUANT_ENABLED" = "true" ]; then
    echo "✅ Quantization: ENABLED"
else
    echo "❌ Quantization: DISABLED"
fi

# Show model info
echo ""
echo "📊 Model Information:"
echo "$MODEL_INFO" | jq '{
    quantization_enabled,
    model,
    model_size_quantized,
    size_reduction
}'
```

Usage:
```bash
chmod +x monitor-quantization.sh
./monitor-quantization.sh
```

