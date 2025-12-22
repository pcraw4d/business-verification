# Python ML Service Connectivity Audit - Track 6.1

## Executive Summary

Investigation of Python ML service connectivity reveals **circuit breaker is OPEN**, preventing ensemble voting and blocking ML-based classification. The service itself appears to be healthy, but the circuit breaker opened due to consecutive failures and has not recovered.

**Status**: ⚠️ **CRITICAL** - Circuit breaker OPEN blocking all ML requests

## Service Configuration

### Environment Variable

**Location**: `services/classification-service/cmd/main.go:189`

```go
pythonMLServiceURL := os.Getenv("PYTHON_ML_SERVICE_URL")
```

**Expected Value** (from `railway.json`):
- Production: `https://python-ml-service-production.up.railway.app`
- Staging: `https://python-ml-service-staging.up.railway.app`

**Status**: ✅ Configured (verified in code)

### Service Initialization

**Location**: `services/classification-service/cmd/main.go:193-200`

```go
pythonMLService = infrastructure.NewPythonMLService(pythonMLServiceURL, stdLogger)
if err := pythonMLService.InitializeWithRetry(initCtx, 3); err != nil {
    logger.Warn("⚠️ Failed to initialize Python ML Service after retries, continuing without enhanced classification",
        zap.Error(err))
    pythonMLService = nil
}
```

**Status**: ⚠️ May fail during initialization if service is unavailable

## Circuit Breaker Configuration

### Configuration Details

**Location**: `internal/machine_learning/infrastructure/python_ml_service.go:100-107`

```go
circuitBreakerConfig := resilience.DefaultCircuitBreakerConfig()
circuitBreakerConfig.FailureThreshold = 10  // Opens after 10 consecutive failures
circuitBreakerConfig.Timeout = 60 * time.Second // Stays open for 60s
circuitBreakerConfig.SuccessThreshold = 2 // Needs 2 successes to close
circuitBreakerConfig.ResetTimeout = 120 * time.Second // Reset timeout
```

**Current Settings**:
- **Failure Threshold**: 10 consecutive failures
- **Open Timeout**: 60 seconds
- **Success Threshold**: 2 successes to close
- **Reset Timeout**: 120 seconds

### Circuit Breaker States

1. **CLOSED**: Normal operation, requests pass through
2. **OPEN**: Circuit is open, requests are rejected immediately
3. **HALF_OPEN**: Testing recovery, allows limited requests through

### State Transitions

- **CLOSED → OPEN**: After 10 consecutive failures
- **OPEN → HALF_OPEN**: After 60 seconds timeout
- **HALF_OPEN → CLOSED**: After 2 successful requests
- **HALF_OPEN → OPEN**: After 1 failure

## Health Monitoring

### Automatic Health Checks

**Location**: `internal/machine_learning/infrastructure/python_ml_service.go:765-810`

**Configuration**:
- **Interval**: Every 30 seconds
- **Health Endpoint**: `/health`
- **Automatic Recovery**: Enabled

**Recovery Logic**:
```go
// If service is healthy and circuit is open, reset after 60s
if healthCheck.Status == "pass" && cbState == resilience.CircuitOpen {
    timeSinceOpen := time.Since(cbStats.StateChange)
    if timeSinceOpen >= 60*time.Second {
        pms.ResetCircuitBreaker()
    }
}
```

**Status**: ✅ Automatic recovery implemented

### Health Check Endpoint

**Location**: `internal/machine_learning/infrastructure/python_ml_service.go:958-990`

**Endpoint**: `/health` (with circuit breaker info)

**Returns**:
- Service health status
- Circuit breaker state
- Circuit breaker metrics (failures, successes, etc.)

## Test Results from Previous Analysis

### Service Health Status

**From**: `docs/ml_service_circuit_breaker_analysis.md`

**Service Health**: ✅ Healthy
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
- Response time: < 1 second

### Circuit Breaker Status

**State**: ❌ OPEN
- All ML classification requests being rejected
- Error: "Circuit breaker is OPEN - request rejected"
- Impact: System falling back to Go keyword-based classification

## Root Cause Analysis

### Why Circuit Breaker Opened

Based on previous analysis and current findings:

1. **Timeout Mismatch** ⚠️ **HIGH**
   - Classification service request timeout: 120s (updated from 10s)
   - ML service HTTP client timeout: 30s
   - But requests may still timeout due to:
     - Website scraping delays
     - Database query timeouts
     - Network latency

2. **Consecutive Failures** ⚠️ **HIGH**
   - Circuit breaker opens after 10 consecutive failures
   - Failures could be due to:
     - Service startup issues
     - Network connectivity problems
     - Timeout issues
     - Service overload

3. **Recovery Not Happening** ⚠️ **MEDIUM**
   - Circuit breaker should recover after 60s if service is healthy
   - But if requests continue to fail, it won't recover
   - Health monitoring should reset it, but may not be working

### Current State

**From Test Results** (`FINAL_VALIDATION_385_SAMPLE_ANALYSIS_20251222.md`):
- Circuit breaker is OPEN
- ML-based classification unavailable
- System using Go keyword-based classification only
- Classification accuracy: 10.7% (should be higher with ML)

## Investigation Steps

### Step 1: Check Service Availability

**Test Script**: `scripts/test_python_ml_service.go`

**Tests**:
1. Ping endpoint (`/ping`)
2. Health check (`/health`)
3. Fast classification (`/classify-fast`)
4. Enhanced classification (`/classify-enhanced`)

**Expected Service URL**: `https://python-ml-service-production.up.railway.app`

### Step 2: Review Circuit Breaker Status

**Health Endpoint**: `/health` on classification service

**Check**:
```bash
curl https://classification-service-production.up.railway.app/health | jq '.ml_service_status.circuit_breaker_state'
```

**Expected States**:
- `CLOSED`: ✅ Normal operation
- `OPEN`: ❌ Blocking requests
- `HALF_OPEN`: 🔄 Testing recovery

### Step 3: Review Circuit Breaker Metrics

**Check Metrics**:
```bash
curl https://classification-service-production.up.railway.app/health | jq '.ml_service_status.circuit_breaker_metrics'
```

**Metrics to Review**:
- Failure count
- Success count
- State change time
- Last failure time

### Step 4: Test Service Manually

**Use Test Script**:
```bash
go run scripts/test_python_ml_service.go https://python-ml-service-production.up.railway.app
```

**Expected Results**:
- All tests should pass if service is healthy
- If tests fail, identify the specific issue

## Recommendations

### Immediate Actions (High Priority)

1. **Check Circuit Breaker State**:
   - Use health endpoint to check current state
   - Review circuit breaker metrics
   - Identify why it opened

2. **Test Service Connectivity**:
   - Run test script against production service
   - Verify service is actually healthy
   - Check response times

3. **Reset Circuit Breaker** (if service is healthy):
   - Use manual reset endpoint: `POST /admin/circuit-breaker/reset`
   - Or wait for automatic recovery (60s timeout)
   - Monitor recovery process

### Medium Priority Actions

4. **Review Timeout Configuration**:
   - Ensure timeouts are aligned
   - ML service timeout: 30s
   - Classification service timeout: 120s
   - Website scraping timeout: 15s

5. **Improve Circuit Breaker Recovery**:
   - Review automatic recovery logic
   - Ensure health monitoring is working
   - Consider reducing failure threshold if too sensitive

6. **Add Monitoring**:
   - Track circuit breaker state changes
   - Alert when circuit breaker opens
   - Monitor recovery success rate

## Code Locations

- **Service Initialization**: `services/classification-service/cmd/main.go:187-200`
- **Circuit Breaker Config**: `internal/machine_learning/infrastructure/python_ml_service.go:100-107`
- **Health Monitoring**: `internal/machine_learning/infrastructure/python_ml_service.go:765-810`
- **Circuit Breaker Methods**: `internal/machine_learning/infrastructure/python_ml_service.go:881-894`
- **Health Endpoint**: `services/classification-service/internal/handlers/classification.go:5075-5175`
- **Reset Endpoint**: `services/classification-service/internal/handlers/classification.go:4971-5031`

## Service Initialization

### Initialization Process

**Location**: `internal/machine_learning/infrastructure/python_ml_service.go:138-184`

**Steps**:
1. Initialize metrics and health status
2. Test connection (5s timeout)
3. Load available models (30s timeout)
4. Start health monitoring (30s interval)
5. Start metrics collection (60s interval)

**Initialization with Retry**:
- **Location**: `internal/machine_learning/infrastructure/python_ml_service.go:186-220`
- **Max Retries**: 3 (configurable)
- **Backoff**: Exponential (2s, 4s, 6s)
- **Circuit Breaker Reset**: Resets before initialization

**Connection Test**:
- **Location**: `internal/machine_learning/infrastructure/python_ml_service.go:163-170`
- **Timeout**: 5 seconds
- **Endpoint**: `/ping` or `/health`

## Manual Reset Endpoint

### Reset Circuit Breaker

**Location**: `services/classification-service/internal/handlers/classification.go:4971-5031`

**Endpoint**: `POST /admin/circuit-breaker/reset`

**Usage**:
```bash
curl -X POST https://classification-service-production.up.railway.app/admin/circuit-breaker/reset
```

**Response**:
```json
{
  "success": true,
  "message": "Circuit breaker reset successfully",
  "old_state": "open",
  "new_state": "closed",
  "old_metrics": {...},
  "new_metrics": {...}
}
```

**Status**: ✅ Available for manual recovery

## Next Steps

1. ✅ **Complete Track 6.1 Investigation** - This document
2. **Run Service Connectivity Tests** - Use test script
3. **Check Circuit Breaker State** - Via health endpoint
4. **Reset Circuit Breaker** - If service is healthy (manual or automatic)
5. **Monitor Recovery** - Track state changes
6. **Validate ML Service Usage** - Verify ensemble voting is working

## Expected Impact

After fixing circuit breaker:

1. **ML Service Usage**: 0% → >80%
2. **Classification Accuracy**: 10.7% → 50-70% (with ML boost)
3. **Confidence Scores**: 24.65% → 50-60% (with ML boost)
4. **Ensemble Voting**: Enabled, combining Go + ML results

