# ML Service Integration Investigation

**Date**: 2025-11-30  
**Issue**: ML service not working with testing  
**Status**: 🔍 **Under Investigation**

---

## Problem Summary

According to the test results analysis, Phase 1 (keyword-enhanced ML input) is **implemented but NOT ACTIVE** because:

- No Phase 1 logs found in output
- `PYTHON_ML_SERVICE_URL` not set
- ML classification not being used

---

## Root Cause Analysis

### Issue 1: Environment Variable Not Set

**Location**: `cmd/comprehensive_accuracy_test/main.go:118`

```go
pythonMLServiceURL := os.Getenv("PYTHON_ML_SERVICE_URL")
if pythonMLServiceURL != "" {
    // Initialize service
} else {
    logger.Println("ℹ️  Python ML Service URL not configured (PYTHON_ML_SERVICE_URL), ML classification will not be available")
}
```

**Problem**: The test harness requires `PYTHON_ML_SERVICE_URL` to be set, but it's not being set when running tests.

**Impact**: ML service is never initialized, so Phase 1 never runs.

### Issue 2: Python ML Service Not Running

**Location**: `internal/machine_learning/infrastructure/python_ml_service.go:561-578`

```go
func (pms *PythonMLService) testConnection(ctx context.Context) error {
    httpReq, err := http.NewRequestWithContext(ctx, "GET", pms.endpoint+"/ping", nil)
    // ... makes HTTP request to /ping endpoint
}
```

**Problem**: Even if `PYTHON_ML_SERVICE_URL` is set, the service must be running and accessible.

**Impact**: If the service isn't running, initialization fails and ML is disabled.

### Issue 3: Service Initialization Failure Handling

**Location**: `cmd/comprehensive_accuracy_test/main.go:274-288`

```go
func initPythonMLService(endpoint string, logger *log.Logger) interface{} {
    service := infrastructure.NewPythonMLService(endpoint, logger)
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := service.Initialize(ctx); err != nil {
        logger.Printf("⚠️  Failed to initialize Python ML Service: %v", err)
        return nil  // Returns nil on failure
    }
    
    return service
}
```

**Problem**: If initialization fails, the service returns `nil`, which means ML is disabled but the test continues.

**Impact**: Tests run without ML, but no clear error is raised.

---

## Integration Flow

### Expected Flow (When Working)

```
1. Test starts
   ↓
2. Check PYTHON_ML_SERVICE_URL environment variable
   ↓
3. If set → Initialize PythonMLService
   ↓
4. Call service.Initialize()
   ↓
5. testConnection() → GET /ping
   ↓
6. If /ping succeeds → Service initialized ✅
   ↓
7. Create IndustryDetectionService with ML support
   ↓
8. Tests run with ML classification enabled
```

### Actual Flow (Current Issue)

```
1. Test starts
   ↓
2. Check PYTHON_ML_SERVICE_URL environment variable
   ↓
3. NOT SET → Skip ML initialization
   ↓
4. Create IndustryDetectionService WITHOUT ML support
   ↓
5. Tests run with keyword-based classification only
   ↓
6. Phase 1 logs never appear (ML never called)
```

---

## Verification Steps

### Step 1: Check Environment Variable

```bash
# Check if PYTHON_ML_SERVICE_URL is set
echo $PYTHON_ML_SERVICE_URL

# If not set, set it:
export PYTHON_ML_SERVICE_URL="http://localhost:8000"
```

### Step 2: Check Python ML Service Status

```bash
# Check if service is running
curl http://localhost:8000/ping

# Expected response:
# {"status": "ok", "message": "Python ML Service is running"}

# Check health
curl http://localhost:8000/health

# Expected response:
# {"status": "healthy", "timestamp": "...", "service": "Python ML Service", ...}
```

### Step 3: Check Test Logs

When running tests, look for:

**✅ Success:**
```
🐍 Initializing Python ML Service: http://localhost:8000
✅ Python ML Service initialized successfully
```

**❌ Failure:**
```
🐍 Initializing Python ML Service: http://localhost:8000
⚠️  Failed to initialize Python ML Service: failed to connect to Python ML service: ping request failed: ...
```

**ℹ️ Not Configured:**
```
ℹ️  Python ML Service URL not configured (PYTHON_ML_SERVICE_URL), ML classification will not be available
```

---

## Solutions

### Solution 1: Use Automated Script (Recommended)

The project includes a script that handles everything:

```bash
./scripts/run_ml_accuracy_tests.sh
```

This script:
1. Checks if Python ML service is running
2. Starts it if needed
3. Waits for it to be ready
4. Sets `PYTHON_ML_SERVICE_URL`
5. Runs tests

### Solution 2: Manual Setup

#### Step 1: Start Python ML Service

```bash
cd python_ml_service

# Activate virtual environment
source venv/bin/activate  # or: python3 -m venv venv && source venv/bin/activate

# Install dependencies (if needed)
pip install -r requirements.txt

# Start service
python app.py
# Or: uvicorn app:app --host 0.0.0.0 --port 8000
```

#### Step 2: Verify Service is Running

```bash
# In another terminal
curl http://localhost:8000/ping
curl http://localhost:8000/health
```

#### Step 3: Set Environment Variable and Run Tests

```bash
export PYTHON_ML_SERVICE_URL="http://localhost:8000"
./bin/comprehensive_accuracy_test -verbose -output accuracy_report_ml.json
```

### Solution 3: Use Railway URL (If Deployed)

If the Python ML service is deployed on Railway:

```bash
export PYTHON_ML_SERVICE_URL="https://python-ml-service-production-xxx.up.railway.app"
./bin/comprehensive_accuracy_test -verbose -output accuracy_report_ml.json
```

---

## Diagnostic Script

A diagnostic script has been created to help identify issues:

```bash
./scripts/diagnose_ml_service.sh
```

This script will:
1. Check if `PYTHON_ML_SERVICE_URL` is set
2. Check if Python ML service is running
3. Test `/ping` and `/health` endpoints
4. Verify service initialization
5. Provide recommendations

---

## Common Issues and Fixes

### Issue: "ping request failed: connection refused"

**Cause**: Python ML service is not running.

**Fix**: Start the Python ML service (see Solution 2, Step 1).

### Issue: "ping returned status 404"

**Cause**: Service is running but `/ping` endpoint doesn't exist (shouldn't happen, but possible if service is misconfigured).

**Fix**: Check `python_ml_service/app.py` for `/ping` endpoint (should be at line 924).

### Issue: "ping request failed: timeout"

**Cause**: Service is running but not responding (may be loading models).

**Fix**: Wait longer for service to start, or check service logs.

### Issue: Environment variable not persisting

**Cause**: Variable set in one terminal but test run in another.

**Fix**: Set variable in the same terminal where tests are run, or use the automated script.

---

## Next Steps

1. **Run Diagnostic Script**: `./scripts/diagnose_ml_service.sh`
2. **Start Python ML Service**: Use automated script or manual steps
3. **Re-run Tests**: With `PYTHON_ML_SERVICE_URL` set
4. **Verify Phase 1 Logs**: Look for `[Phase 1]` logs in test output
5. **Compare Results**: Compare ML-enabled vs keyword-only results

---

## References

- Test Results Analysis: `docs/integration_phases_test_results_analysis.md`
- ML Testing Guide: `docs/ml_accuracy_testing_guide.md`
- Python ML Service Setup: `python_ml_service/README.md`
- Automated Test Script: `scripts/run_ml_accuracy_tests.sh`

