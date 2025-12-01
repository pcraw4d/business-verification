# Streaming Response Testing Guide

## Overview

This guide explains how to test the streaming responses feature (OPTIMIZATION #17) for the classification service.

## Prerequisites

1. **Classification Service Running**:
   - Local: `http://localhost:8081` (default) or `http://localhost:8080`
   - Production: `https://classification-service-production.up.railway.app`

2. **Tools Required**:
   - `curl` (for command-line testing)
   - `jq` (for JSON parsing) - Install: `brew install jq` (macOS) or `apt-get install jq` (Linux)
   - `bash` (for running test scripts)

## Quick Test

### Manual Test with curl

```bash
curl -X POST "http://localhost:8081/v1/classify?stream=true" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Microsoft Corporation",
    "description": "Software development and cloud computing services",
    "website_url": "https://microsoft.com"
  }'
```

**Expected Output**: NDJSON format with progress updates:
```json
{"type":"progress","request_id":"req_...","status":"started","message":"Classification started","timestamp":"..."}
{"type":"progress","request_id":"req_...","status":"classifying","message":"Analyzing business and website","step":"classification"}
{"type":"progress","request_id":"req_...","status":"industry_detected","message":"Industry detected","step":"industry","primary_industry":"Technology","confidence":0.95}
...
{"type":"complete","request_id":"req_...","data":{...full response...},"processing_time_ms":2500}
```

## Automated Testing

### Option 1: Simple Test Script

```bash
# Set service URL (optional, defaults to localhost:8081)
export CLASSIFICATION_SERVICE_URL="http://localhost:8081"

# Run simple test
./scripts/test_streaming_simple.sh
```

### Option 2: Comprehensive Test Script

```bash
# Set service URL (optional, auto-detects)
export CLASSIFICATION_SERVICE_URL="http://localhost:8081"

# Run comprehensive test
./scripts/test_streaming_comprehensive.sh
```

**What it tests**:
- Health check
- Basic streaming request
- Progress updates parsing
- Streaming vs non-streaming comparison
- Multiple business types
- Performance metrics

## Understanding the Response Format

### NDJSON Format

Each line is a complete JSON object, separated by newlines (`\n`):

```
{"type":"progress","status":"started",...}
{"type":"progress","status":"classifying",...}
{"type":"progress","status":"industry_detected",...}
{"type":"complete","data":{...},...}
```

### Message Types

1. **Progress Messages** (`type: "progress"`):
   - `started`: Classification process initiated
   - `classifying`: Analyzing business and website
   - `industry_detected`: Industry identified
   - `generating_codes`: Generating classification codes
   - `codes_generated`: Codes generated
   - `assessing_risk`: Risk assessment in progress
   - `risk_assessed`: Risk assessment completed
   - `verification_complete`: Verification status generated

2. **Complete Message** (`type: "complete"`):
   - Contains full response in `data` field
   - Includes `processing_time_ms`

3. **Error Message** (`type: "error"`):
   - Contains error details
   - Includes `status_code`

## Testing Different Scenarios

### Test 1: Technology Company

```bash
curl -X POST "http://localhost:8081/v1/classify?stream=true" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Apple Inc",
    "description": "Consumer electronics and software",
    "website_url": "https://apple.com"
  }'
```

### Test 2: Financial Services

```bash
curl -X POST "http://localhost:8081/v1/classify?stream=true" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "JPMorgan Chase",
    "description": "Banking and financial services",
    "website_url": "https://jpmorganchase.com"
  }'
```

### Test 3: Healthcare

```bash
curl -X POST "http://localhost:8081/v1/classify?stream=true" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Mayo Clinic",
    "description": "Medical center and hospital services",
    "website_url": "https://mayoclinic.org"
  }'
```

## Performance Comparison

### Non-Streaming (Traditional)

```bash
time curl -X POST "http://localhost:8081/v1/classify" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Corp", "website_url": "https://test.com"}'
```

### Streaming

```bash
time curl -X POST "http://localhost:8081/v1/classify?stream=true" \
  -H "Content-Type: application/json" \
  -d '{"business_name": "Test Corp", "website_url": "https://test.com"}'
```

**Expected**: Streaming should show first progress update much faster than non-streaming returns full response.

## Troubleshooting

### Issue: "Streaming not supported"

**Cause**: Response writer doesn't implement `http.Flusher`
**Solution**: This shouldn't happen in production. Check server configuration.

### Issue: No progress updates

**Cause**: Service might be too fast, or caching
**Solution**: Try with a different business or clear cache

### Issue: Connection timeout

**Cause**: Service taking too long
**Solution**: Check service health, increase timeout

### Issue: Invalid JSON

**Cause**: Response might be corrupted
**Solution**: Check service logs, verify endpoint URL

## Integration Testing

### Test with API Gateway

If using API Gateway, test through gateway:

```bash
curl -X POST "https://api-gateway-service-production-21fd.up.railway.app/api/v1/classify?stream=true" \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Test Corp",
    "website_url": "https://test.com"
  }'
```

### Test with Frontend

Frontend should:
1. Parse NDJSON line by line
2. Display progress updates
3. Show final result when `type: "complete"` received

## Expected Performance

- **Time to first byte**: 200-500ms (vs 2-5s for non-streaming)
- **Perceived latency improvement**: 50-70%
- **Total processing time**: Same as non-streaming (but user sees progress earlier)

## Next Steps

1. ✅ Test streaming responses
2. ✅ Verify progress updates
3. ✅ Compare performance
4. ⏭️ Integrate with frontend
5. ⏭️ Monitor adoption metrics

