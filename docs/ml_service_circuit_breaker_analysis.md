# ML Service Circuit Breaker Analysis

**Date**: 2025-12-01  
**Status**: Investigation Complete

---

## Executive Summary

The accuracy tests ran successfully against Railway deployment, but revealed that the Python ML service circuit breaker is OPEN, preventing ML classification from being used. The ML service itself is healthy and responding correctly, indicating the circuit breaker opened during the test run due to timeouts or failures.

---

## Test Results Summary

### Accuracy Metrics
- **Overall Accuracy**: 2.52% (Target: 95%) ❌
- **Industry Accuracy**: 0.00% (Target: 95%) ❌
- **Code Accuracy**: 4.20% (Target: 90%) ❌
- **Test Cases**: 184 (All passed - no crashes)
- **Average Processing Time**: 7.95 seconds

### Circuit Breaker Status
- **State**: OPEN (circuit breaker is open)
- **Impact**: All ML classification requests are being rejected
- **Fallback**: System is using Go ML classifier → keyword-based classification

---

## Investigation Findings

### 1. Python ML Service Health ✅

**Status**: Healthy and Operational

```bash
curl https://python-ml-service-production-a6b8.up.railway.app/health
```

**Response**:
```json
{
  "status": "healthy",
  "service": "Python ML Service",
  "version": "2.0.0",
  "models_status": "loaded",
  "distilbart_classifier": "loaded"
}
```

**Direct Classification Test**: ✅ Working
- Successfully classified test requests
- Returns classifications with confidence scores
- Response time: < 1 second

### 2. Classification Service Health ⚠️

**Status**: Service is running but classification endpoint returns 502

**Health Endpoint**: ✅ Working
```json
{
  "status": "healthy",
  "ml_enabled": true,
  "features": {
    "ml_enabled": true,
    "ensemble_enabled": true
  }
}
```

**Classification Endpoint**: ❌ Returns 502
- Error: "Application failed to respond"
- Suggests timeout or crash during request processing

### 3. Circuit Breaker Analysis

**Why Circuit Breaker Opened**:

1. **Timeout Mismatch**:
   - HTTP Client timeout: 30 seconds
   - Classification service request timeout: 10 seconds
   - ML service may take longer than 10s for some requests
   - Circuit breaker opens after 10 consecutive failures

2. **Possible Causes**:
   - Network latency between services
   - ML service processing time > 10 seconds for complex requests
   - Website scraping taking too long (even with 5s timeout)
   - Database query timeouts

3. **Circuit Breaker Configuration**:
   - Failure threshold: 10 consecutive failures
   - Timeout: 60 seconds (stays open for 60s)
   - Success threshold: 2 successes to close
   - Reset timeout: 120 seconds

---

## Root Cause Analysis

### Primary Issue: Timeout Configuration Mismatch

The classification service has a **10-second request timeout**, but:
- ML service HTTP client timeout is **30 seconds**
- Website scraping timeout is **5 seconds** (but can accumulate)
- Database queries may take time
- Overall processing can exceed 10 seconds

When requests exceed 10 seconds, they timeout, causing the circuit breaker to count failures. After 10 consecutive failures, the circuit opens.

### Secondary Issue: Circuit Breaker Recovery

Once the circuit breaker opens:
- It stays open for 60 seconds
- Needs 2 successful requests to close
- But if requests are still timing out, it won't recover

---

## Solutions

### Solution 1: Increase Request Timeout (Immediate)

**File**: `services/classification-service/internal/config/config.go`

**Current**:
```go
RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 10*time.Second),
```

**Recommended**:
```go
RequestTimeout: getEnvAsDuration("REQUEST_TIMEOUT", 30*time.Second), // Match ML service timeout
```

**Or via Environment Variable**:
```bash
REQUEST_TIMEOUT=30s
```

### Solution 2: Add Circuit Breaker Status Endpoint (Monitoring)

**Status**: ✅ Implemented

Added circuit breaker status to `/health` endpoint. The health endpoint now includes:
- Circuit breaker state (CLOSED/OPEN/HALF_OPEN)
- Circuit breaker metrics (failure count, success count, etc.)
- ML service health status

**Usage**:
```bash
curl https://classification-service-production.up.railway.app/health | jq '.ml_service_status'
```

### Solution 3: Reset Circuit Breaker (Recovery)

**Option A: Redeploy Classification Service**
- Circuit breaker resets on service restart
- Quick fix but requires deployment

**Option B: Wait for Automatic Recovery**
- Circuit breaker will attempt recovery after 60 seconds
- Needs 2 successful requests to close
- May not recover if timeouts persist

**Option C: Add Manual Reset Endpoint** (Future Enhancement)
- Add `/health/reset-circuit-breaker` endpoint
- Requires authentication/authorization
- Allows manual recovery without redeployment

### Solution 4: Optimize Request Processing (Long-term)

1. **Reduce Website Scraping Time**:
   - Already optimized to 5s timeout
   - Caching implemented (24h TTL)
   - Consider reducing further or making it optional

2. **Optimize Database Queries**:
   - Already has caching infrastructure
   - Consider query optimization

3. **Parallel Processing**:
   - Already implemented for some operations
   - Can be enhanced further

---

## Immediate Actions

### 1. Check Current Circuit Breaker State

```bash
# Use the diagnostic script
./scripts/diagnose_ml_service_circuit_breaker.sh

# Or check health endpoint
curl https://classification-service-production.up.railway.app/health | jq '.ml_service_status'
```

### 2. Increase Request Timeout

**Via Railway Environment Variables**:
1. Go to Railway dashboard
2. Select `classification-service`
3. Go to Variables tab
4. Add/Update: `REQUEST_TIMEOUT=30s`
5. Service will redeploy automatically

### 3. Monitor Circuit Breaker Recovery

After increasing timeout:
- Monitor health endpoint for circuit breaker state
- Circuit breaker should attempt recovery after 60s
- Should close after 2 successful requests

### 4. Re-run Accuracy Tests

After circuit breaker recovers:
```bash
./scripts/run_tests_against_railway_production.sh
```

---

## Verification Steps

### Step 1: Verify Timeout Configuration

```bash
# Check current timeout (after update)
curl https://classification-service-production.up.railway.app/health | jq '.features'
```

### Step 2: Test Classification with ML

```bash
curl -X POST https://classification-service-production.up.railway.app/v1/classify \
  -H "Content-Type: application/json" \
  -d '{
    "business_name": "Acme Technology Corp",
    "description": "Software development and cloud services",
    "website_url": "https://www.acme.com"
  }' | jq '{industry: .industry_name, method: .classification_method, confidence: .confidence_score}'
```

**Expected**: `classification_method` should be `"ml_distilbart"` or `"ml"` if circuit breaker is closed.

### Step 3: Monitor Circuit Breaker State

```bash
# Continuous monitoring
watch -n 5 'curl -s https://classification-service-production.up.railway.app/health | jq ".ml_service_status.circuit_breaker_state"'
```

---

## Expected Outcomes

After implementing the timeout fix:

1. **Circuit Breaker**: Should close after 2 successful requests
2. **ML Service Usage**: Should increase from 0% to >80%
3. **Accuracy**: Should improve significantly (ML is more accurate than keyword-only)
4. **Processing Time**: May increase slightly but should stay < 10s for most requests

---

## Monitoring Recommendations

### 1. Set Up Alerts

- Alert when circuit breaker state = OPEN for > 2 minutes
- Alert when ML service utilization < 50% (when service is available)
- Alert when average processing time > 15 seconds

### 2. Regular Health Checks

- Run diagnostic script daily: `./scripts/diagnose_ml_service_circuit_breaker.sh`
- Monitor health endpoint: `/health`
- Track circuit breaker state changes

### 3. Performance Monitoring

- Track ML service response times
- Monitor circuit breaker failure rates
- Track classification method usage (ML vs fallback)

---

## Files Modified

1. ✅ `services/classification-service/internal/handlers/classification.go`
   - Added circuit breaker status to health endpoint
   - Includes state, metrics, and health check information

2. ✅ `scripts/diagnose_ml_service_circuit_breaker.sh`
   - Comprehensive diagnostic tool
   - Tests ML service directly
   - Checks classification service
   - Provides recommendations

---

## Next Steps

1. **Immediate**: Increase `REQUEST_TIMEOUT` to 30s in Railway
2. **Immediate**: Monitor circuit breaker recovery
3. **This Week**: Re-run accuracy tests after circuit breaker closes
4. **This Week**: Verify ML service is being used (>80% utilization)
5. **Future**: Add manual circuit breaker reset endpoint
6. **Future**: Optimize processing time further

---

## References

- Circuit Breaker Implementation: `internal/resilience/circuit_breaker.go`
- Python ML Service: `internal/machine_learning/infrastructure/python_ml_service.go`
- Classification Handler: `services/classification-service/internal/handlers/classification.go`
- Diagnostic Script: `scripts/diagnose_ml_service_circuit_breaker.sh`

